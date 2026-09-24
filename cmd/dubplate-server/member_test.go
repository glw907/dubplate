package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// memberEnviron returns authEnviron(t) with DUBPLATE_ROSTER_FILE
// pointed at a fresh path under a directory the test owns: fixture
// state for the create rows, which need a roster path whose
// containing directory exists but whose file does not.
func memberEnviron(t *testing.T) []string {
	t.Helper()
	dir := t.TempDir()
	return replaceKey(authEnviron(t), "DUBPLATE_ROSTER_FILE", filepath.Join(dir, "roster.json"))
}

// writeFixtureRoster writes content to path, the way an operator's
// hand edit or a prior run would have left it: a literal JSON string,
// never round-tripped through the roster package's own encoder, so a
// test asserting against this fixture is not proving the encoder
// agrees with itself.
func writeFixtureRoster(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture roster: %v", err)
	}
}

func TestMemberAddCmd_CreatesRosterWithFirstAdmin(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "Alice", "--streamer-username", "alice", "a@example.test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil; stderr = %q", err, stderr.String())
	}

	info, err := os.Stat(rosterPath)
	if err != nil {
		t.Fatalf("Stat(roster) after create: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("roster file mode = %v, want 0600", perm)
	}

	if audit := stderr.String(); !strings.Contains(audit, `address=a@example.test`) ||
		!strings.Contains(audit, `admin=true`) ||
		!strings.Contains(audit, `channel=command-line`) ||
		!strings.Contains(audit, `client_ip=local`) {
		t.Errorf("audit record on stderr = %q, want actor, channel=command-line, address=a@example.test, admin=true, client_ip=local", audit)
	}
}

func TestMemberAddCmd_AddsToExistingRosterWithAdmin(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")
	writeFixtureRoster(t, rosterPath, `{"schema_version":1,"members":[{"address":"a@example.test","display_name":"A","streamer_username":"a","admin":true}]}`)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "Bob", "--streamer-username", "bob", "b@example.test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil; stderr = %q", err, stderr.String())
	}

	data, err := os.ReadFile(rosterPath)
	if err != nil {
		t.Fatalf("read roster: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `"address":"a@example.test"`) {
		t.Errorf("roster after add = %q, want a@example.test still present", got)
	}
	if !strings.Contains(got, `"address":"b@example.test"`) {
		t.Errorf("roster after add = %q, want b@example.test added", got)
	}
}

func TestMemberAddCmd_NoAdminRefusedOnCreate(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--name", "Carol", "--streamer-username", "carol", "c@example.test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the no-admin refusal")
	}
	if !strings.Contains(err.Error(), "must name at least one admin") {
		t.Errorf("error = %q, want it to name the no-admin refusal", err.Error())
	}
	if strings.Contains(err.Error(), "c@example.test") {
		t.Errorf("error = %q, want it never to carry the refused address", err.Error())
	}
	if _, statErr := os.Stat(rosterPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("Stat(roster) after refusal error = %v, want ErrNotExist; nothing should be created", statErr)
	}
}

func TestMemberAddCmd_NoAdminRefusedOnHandEditedRoster(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")
	before := `{"schema_version":1,"members":[{"address":"nonadmin@example.test","display_name":"N","streamer_username":"n"}]}`
	writeFixtureRoster(t, rosterPath, before)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--name", "Dave", "--streamer-username", "dave", "d@example.test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the no-admin refusal")
	}
	if !strings.Contains(err.Error(), "must name at least one admin") {
		t.Errorf("error = %q, want it to name the no-admin refusal", err.Error())
	}
	after, readErr := os.ReadFile(rosterPath)
	if readErr != nil {
		t.Fatalf("read roster after refusal: %v", readErr)
	}
	if string(after) != before {
		t.Errorf("roster file changed after refusal: got %q, want unchanged %q", after, before)
	}
}

func TestMemberAddCmd_MissingUsernameRefusedOnCreate(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "A", "--streamer-username", "", "a@example.test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the missing-username refusal")
	}
	if !strings.Contains(err.Error(), "streamer_username is required") {
		t.Errorf("error = %q, want it to name the missing-username refusal", err.Error())
	}
	if _, statErr := os.Stat(rosterPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("Stat(roster) after refusal error = %v, want ErrNotExist; nothing should be created", statErr)
	}
}

func TestMemberAddCmd_MissingRequiredFlagsRefusedBeforeOpen(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing --name", args: []string{"add", "--streamer-username", "a", "a@example.test"}},
		{name: "missing --streamer-username", args: []string{"add", "--name", "A", "a@example.test"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := memberEnviron(t)
			rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")

			var stdout, stderr bytes.Buffer
			cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err == nil {
				t.Fatal("Execute() error = nil, want a required-flag error")
			}
			if _, statErr := os.Stat(rosterPath); !errors.Is(statErr, os.ErrNotExist) {
				t.Errorf("Stat(roster) error = %v, want ErrNotExist; a missing required flag must refuse before any open", statErr)
			}
		})
	}
}

func TestMemberAddCmd_InvalidRosterReportsLoadError(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")
	writeFixtureRoster(t, rosterPath, `not valid json`)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "A", "--streamer-username", "a", "a@example.test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the roster load error")
	}
	if !strings.Contains(err.Error(), rosterPath) {
		t.Errorf("error = %q, want it to name the roster path %q", err.Error(), rosterPath)
	}
}

func TestMemberAddCmd_RosterDirectoryAbsentRefused(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")
	missingDir := filepath.Join(filepath.Dir(rosterPath), "does-not-exist")
	rosterPath = filepath.Join(missingDir, "roster.json")
	environ = replaceKey(environ, "DUBPLATE_ROSTER_FILE", rosterPath)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "A", "--streamer-username", "a", "a@example.test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the roster directory refusal")
	}
	if !strings.Contains(err.Error(), missingDir) {
		t.Errorf("error = %q, want it to name the missing directory %q", err.Error(), missingDir)
	}
	if _, statErr := os.Stat(rosterPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("Stat(roster) error = %v, want ErrNotExist; nothing should be created", statErr)
	}
}

// TestMemberAddCmd_ForeignDirectoryOwnerRefused proves R-5c14's uid
// guard fires and creates nothing, using /tmp itself as dir: /tmp is
// root-owned with the sticky bit and world-writable, so this test's
// own process can create a file directly under it despite not owning
// it, which is what lets the row prove the guard, rather than a bare
// permission error, is what refuses the write. Pointing the row at
// "/" instead would fail with EACCES whether or not the guard exists,
// proving nothing.
func TestMemberAddCmd_ForeignDirectoryOwnerRefused(t *testing.T) {
	info, err := os.Stat("/tmp")
	if err != nil {
		t.Skipf("Stat(/tmp): %v", err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Skip("cannot read /tmp's owning uid on this platform")
	}
	euid := os.Geteuid()
	if euid == 0 {
		t.Skip("running as root: the uid guard never refuses root")
	}
	if int(stat.Uid) == euid {
		t.Skip("/tmp is owned by this test's own uid on this system")
	}

	rosterPath := filepath.Join("/tmp", fmt.Sprintf("dubplate-member-test-%d-%d.json", os.Getpid(), time.Now().UnixNano()))
	t.Cleanup(func() { _ = os.Remove(rosterPath) })

	environ := replaceKey(memberEnviron(t), "DUBPLATE_ROSTER_FILE", rosterPath)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "A", "--streamer-username", "a", "a@example.test"})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the foreign-owner refusal")
	}
	if !errors.Is(err, errForeignRosterOwner) {
		t.Errorf("error = %v, want it to wrap errForeignRosterOwner", err)
	}
	if !strings.Contains(err.Error(), fmt.Sprint(euid)) || !strings.Contains(err.Error(), "/tmp") {
		t.Errorf("error = %q, want it to name both the effective uid %d and /tmp", err.Error(), euid)
	}
	if _, statErr := os.Stat(rosterPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("Stat(roster) error = %v, want ErrNotExist; nothing should be created", statErr)
	}
}

// TestCheckRosterDirOwner_OwnedDirectoryPasses is the positive row:
// checkRosterDirOwner accepts a directory this test process owns.
func TestCheckRosterDirOwner_OwnedDirectoryPasses(t *testing.T) {
	if err := checkRosterDirOwner(t.TempDir()); err != nil {
		t.Errorf("checkRosterDirOwner(t.TempDir()) error = %v, want nil", err)
	}
}

func TestMemberAddCmd_NoPasswdEntryFallsBackToUID(t *testing.T) {
	orig := osUserLookupID
	defer func() { osUserLookupID = orig }()
	osUserLookupID = func(string) (*user.User, error) {
		return nil, errors.New("no such user")
	}

	environ := memberEnviron(t)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "A", "--streamer-username", "a", "a@example.test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil; stderr = %q", err, stderr.String())
	}

	wantActor := fmt.Sprintf("actor=uid:%d", os.Geteuid())
	if audit := stderr.String(); !strings.Contains(audit, wantActor) {
		t.Errorf("audit record = %q, want it to contain %q", audit, wantActor)
	}
}

func TestMemberAddCmd_LockHeldPastDeadlineReportsBusy(t *testing.T) {
	orig := memberVerbWriteWait
	memberVerbWriteWait = 200 * time.Millisecond
	defer func() { memberVerbWriteWait = orig }()

	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")
	// The directory a roster write locks must exist before the flock
	// call below; create the empty roster so both the fixture roster
	// and its directory are in place before the lock is taken.
	writeFixtureRoster(t, rosterPath, `{"schema_version":1,"members":[{"address":"a@example.test","display_name":"A","streamer_username":"a","admin":true}]}`)

	dir, err := os.Open(filepath.Dir(rosterPath))
	if err != nil {
		t.Fatalf("open roster directory: %v", err)
	}
	defer func() { _ = dir.Close() }()
	if err := syscall.Flock(int(dir.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock roster directory: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"add", "--admin", "--name", "B", "--streamer-username", "b", "b@example.test"})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the busy refusal")
	}
	if !strings.Contains(err.Error(), "roster busy") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "roster busy")
	}
}

// TestMemberAddCmd_ConcurrentBootstrapBothLand runs two --admin adds
// for distinct addresses at once against an absent roster, under
// -race: one wins CreateRoster, the other gets auth.ErrRosterExists
// and reopens, so both members land and the file holds both.
func TestMemberAddCmd_ConcurrentBootstrapBothLand(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")
	reg := productionRegistry()

	run := func(address, name, username string, stderr *bytes.Buffer) error {
		cmd := newMemberCmd(environ, reg, &bytes.Buffer{}, stderr)
		cmd.SetArgs([]string{"add", "--admin", "--name", name, "--streamer-username", username, address})
		return cmd.Execute()
	}

	var wg sync.WaitGroup
	errs := make([]error, 2)
	var stderrs [2]bytes.Buffer
	wg.Add(2)
	go func() {
		defer wg.Done()
		errs[0] = run("a@example.test", "A", "a", &stderrs[0])
	}()
	go func() {
		defer wg.Done()
		errs[1] = run("b@example.test", "B", "b", &stderrs[1])
	}()
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("add #%d: %v; stderr = %q", i, err, stderrs[i].String())
		}
	}

	data, err := os.ReadFile(rosterPath)
	if err != nil {
		t.Fatalf("read roster: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `"address":"a@example.test"`) || !strings.Contains(got, `"address":"b@example.test"`) {
		t.Errorf("roster = %q, want both a@example.test and b@example.test present", got)
	}

	for i, sb := range stderrs {
		if !strings.Contains(sb.String(), "roster member added") {
			t.Errorf("add #%d stderr = %q, want one audit record", i, sb.String())
		}
	}
}

func TestMemberListCmd_AbsentRosterReportsErrRosterMissing(t *testing.T) {
	environ := memberEnviron(t)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want the missing-roster error")
	}
	if !strings.Contains(err.Error(), "roster file missing") {
		t.Errorf("error = %q, want it to name the roster-missing refusal", err.Error())
	}
}

func TestMemberListCmd_PrintsExactRosterTable(t *testing.T) {
	environ := memberEnviron(t)
	rosterPath, _ := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE")
	writeFixtureRoster(t, rosterPath, `{"schema_version":1,"members":[`+
		`{"address":"a@example.test","display_name":"Alice","streamer_username":"alice","admin":true},`+
		`{"address":"b@example.test","display_name":"Bob","streamer_username":"bob"}`+
		`]}`)

	var stdout, stderr bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetArgs([]string{"list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// want is a literal, independent of printMemberList's own format
	// string, so a regression in that format (a dropped column, a
	// changed separator) is caught rather than echoed on both sides.
	want := "ADDRESS\tDISPLAY NAME\tSTREAMER USERNAME\tADMIN\n" +
		"a@example.test\tAlice\talice\tyes\n" +
		"b@example.test\tBob\tbob\t\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

// TestMemberCmd_RequiresNoRosterSourceRefused proves the outcome that
// the roster-backed identity source's own required set governs: a
// selected source that declares no roster key refuses both
// subcommands, naming the selector's value, rather than resolving an
// empty path and opening whatever that resolves to.
func TestMemberCmd_RequiresNoRosterSourceRefused(t *testing.T) {
	reg := config.NewRegistry()
	reg.AddWallRow(config.WallRow{Name: "DUBPLATE_IDENTITY_SOURCE"})
	reg.AddIdentitySource(config.IdentitySource{
		Name:     "no-roster",
		Selector: "DUBPLATE_IDENTITY_SOURCE",
	})
	environ := append(baseEnviron(), "DUBPLATE_IDENTITY_SOURCE=no-roster")

	for _, args := range [][]string{
		{"add", "--admin", "--name", "A", "--streamer-username", "a", "a@example.test"},
		{"list"},
	} {
		var stdout, stderr bytes.Buffer
		cmd := newMemberCmd(environ, reg, &stdout, &stderr)
		cmd.SetArgs(args)

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("Execute(%v) error = nil, want the no-roster refusal", args)
		}
		if !strings.Contains(err.Error(), "no-roster") {
			t.Errorf("Execute(%v) error = %q, want it to name the selected source %q", args, err.Error(), "no-roster")
		}
	}
}

func TestNewRootCmd_RegistersMemberCommand(t *testing.T) {
	cmd := newRootCmd()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "member" {
			return
		}
	}
	t.Error(`newRootCmd().Commands() does not include "member"`)
}
