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

	"github.com/glw907/dubplate/internal/engine/store"
)

// memberEnviron returns an environment whose state directory is a fresh
// t.TempDir() holding a current store, and that directory.
func memberEnviron(t *testing.T) (environ []string, stateDir string) {
	t.Helper()
	environ = serveAuthEnviron(t)
	stateDir = environValue(environ, "DUBPLATE_STATE_DIR")
	seedBackupStore(t, stateDir)
	return environ, stateDir
}

// runMember runs the member verb with args against environ and returns
// what it wrote to stdout and stderr beside its error.
func runMember(t *testing.T, environ []string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := newMemberCmd(environ, productionRegistry(), &out, &errOut)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

// userAddresses returns the address of every users row the store under
// stateDir holds, tombstoned rows included, in address order.
func userAddresses(t *testing.T, stateDir string) []string {
	t.Helper()
	h, err := store.OpenWrite(t.Context(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	rows, err := h.Query(t.Context(), `SELECT address FROM users ORDER BY address`)
	if err != nil {
		t.Fatalf("read users: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			t.Fatalf("scan address: %v", err)
		}
		out = append(out, address)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read users: %v", err)
	}
	return out
}

const memberTableHeader = "ADDRESS\tDISPLAY NAME\tUSERNAME\tADMIN\n"

// TestMemberCmd_AddThenListRoundTrip adds an admin and then a second
// member through the verb against a store serve's own open created, and
// lists them. Fixture state: two members added in a known order, so
// List's created_at order and the count both show; the second add is
// not an admin, so the admin column shows both values. The want text is
// a literal, independent of printMemberList's own format string.
func TestMemberCmd_AddThenListRoundTrip(t *testing.T) {
	environ, _ := memberEnviron(t)

	_, stderr, err := runMember(t, environ, "add", "--admin", "--name", "Alice", "--username", "alice", "a@example.test")
	if err != nil {
		t.Fatalf("first add error = %v, want nil; stderr = %q", err, stderr)
	}
	for _, want := range []string{"address=a@example.test", "admin=true", "channel=command-line", "client_ip=local"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("audit record on stderr = %q, want it to contain %q", stderr, want)
		}
	}
	if _, stderr, err := runMember(t, environ, "add", "--name", "Bob", "--username", "bob", "b@example.test"); err != nil {
		t.Fatalf("second add error = %v, want nil; stderr = %q", err, stderr)
	}

	stdout, _, err := runMember(t, environ, "list")
	if err != nil {
		t.Fatalf("list error = %v, want nil", err)
	}
	want := memberTableHeader +
		"a@example.test\tAlice\talice\tyes\n" +
		"b@example.test\tBob\tbob\t\n"
	if stdout != want {
		t.Errorf("list stdout = %q, want %q", stdout, want)
	}
}

// TestMemberListCmd_OmitsTombstonedMembers proves list prints live rows
// only. Fixture state: three members added through the verb and one of
// them tombstoned by a raw statement, since no verb removes a member.
func TestMemberListCmd_OmitsTombstonedMembers(t *testing.T) {
	environ, stateDir := memberEnviron(t)
	for _, args := range [][]string{
		{"add", "--admin", "--name", "Alice", "--username", "alice", "a@example.test"},
		{"add", "--name", "Bob", "--username", "bob", "b@example.test"},
		{"add", "--name", "Cy", "--username", "cy", "c@example.test"},
	} {
		if _, stderr, err := runMember(t, environ, args...); err != nil {
			t.Fatalf("%v error = %v; stderr = %q", args, err, stderr)
		}
	}
	execStoreSQL(t, storeDBPath(stateDir), `UPDATE users SET tombstoned_at = 1 WHERE address = 'b@example.test'`)

	stdout, _, err := runMember(t, environ, "list")
	if err != nil {
		t.Fatalf("list error = %v, want nil", err)
	}
	want := memberTableHeader +
		"a@example.test\tAlice\talice\tyes\n" +
		"c@example.test\tCy\tcy\t\n"
	if stdout != want {
		t.Errorf("list stdout = %q, want %q", stdout, want)
	}
}

// TestMemberCmd_StoreStates proves each state the store can be in
// refuses both verbs with the store's own words, and that a refusal
// changes nothing. The owner refusal is not rebuilt here: the store's
// geteuid seam is unexported, and TestOpenModes_OwnerAndFilesystemRefused
// proves it through OpenWrite, which both verbs call.
func TestMemberCmd_StoreStates(t *testing.T) {
	head := store.Inspect(t.Context(), store.InspectOptions{StateDir: t.TempDir(), BusyTimeout: testBackupBusy}).BinaryVersion
	reconcilePending := `{"format": 1, "state": "reconcile-pending", "stamp": "20260925T120000.000000Z",
"source": {"remote": "remote:dubplate", "snapshot": "20260924T030000.000000Z"},
"incoming": "restore-20260925T120000.000000Z.db", "pre_restore": "pre-restore-20260925T120000.000000Z.db",
"discard_current": false, "library": null, "version": 2}`
	tests := []struct {
		name    string
		setup   func(t *testing.T, stateDir string)
		wantErr string
		// wantNoLeaf marks the state with no store: nothing may be
		// created for the verb to have refused.
		wantNoLeaf bool
	}{
		{
			name:       "absent",
			setup:      func(*testing.T, string) {},
			wantErr:    "start serve first",
			wantNoLeaf: true,
		},
		{
			name:    "older than the binary",
			setup:   func(t *testing.T, stateDir string) { olderStore(t, stateDir, head-1) },
			wantErr: "start serve to migrate",
		},
		{
			name: "newer than the binary",
			setup: func(t *testing.T, stateDir string) {
				execStoreSQL(t, olderStore(t, stateDir, head), fmt.Sprintf("PRAGMA user_version = %d", head+1))
			},
			wantErr: fmt.Sprintf("version %d is newer than this binary's version %d", head+1, head),
		},
		{
			name: "a restore-intent record",
			setup: func(t *testing.T, stateDir string) {
				seedBackupStore(t, stateDir)
				if err := os.WriteFile(filepath.Join(stateDir, "store", "restore-intent"), []byte(reconcilePending), 0o600); err != nil {
					t.Fatalf("write the restore-intent record: %v", err)
				}
			},
			wantErr: "restore-intent",
		},
	}
	verbs := []struct {
		name string
		args []string
	}{
		{name: "add", args: []string{"add", "--admin", "--name", "Alice", "--username", "alice", "a@example.test"}},
		{name: "list", args: []string{"list"}},
	}
	for _, tt := range tests {
		for _, verb := range verbs {
			t.Run(tt.name+"/"+verb.name, func(t *testing.T) {
				environ := serveAuthEnviron(t)
				stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
				tt.setup(t, stateDir)

				stdout, _, err := runMember(t, environ, verb.args...)
				if err == nil {
					t.Fatalf("%s error = nil, want a refusal naming %q", verb.name, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("%s error = %q, want it to contain %q", verb.name, err, tt.wantErr)
				}
				if strings.Contains(err.Error(), "a@example.test") {
					t.Errorf("%s error = %q, want it never to carry the address", verb.name, err)
				}
				if stdout != "" {
					t.Errorf("%s stdout = %q, want nothing beside a refusal", verb.name, stdout)
				}
				if _, statErr := os.Stat(filepath.Join(stateDir, "store")); tt.wantNoLeaf != errors.Is(statErr, os.ErrNotExist) {
					t.Errorf("Stat(store leaf) error = %v, want absent = %v: a refused verb must create nothing", statErr, tt.wantNoLeaf)
				}
			})
		}
	}
}

// TestMemberAddCmd_WriteRefusals proves the write path's own refusals
// reach the operator and leave the users table unchanged. Fixture
// state: a current store holding one admin, so a refusal that wrote
// regardless would add a second row.
func TestMemberAddCmd_WriteRefusals(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "a missing username",
			args:    []string{"add", "--admin", "--name", "B", "--username", "", "b@example.test"},
			wantErr: "username is required",
		},
		{
			name:    "an empty name",
			args:    []string{"add", "--admin", "--name", "", "--username", "b", "b@example.test"},
			wantErr: "display name is empty or blank",
		},
		{
			name:    "a blank name",
			args:    []string{"add", "--admin", "--name", "   ", "--username", "b", "b@example.test"},
			wantErr: "display name is empty or blank",
		},
		{
			name:    "a duplicate address",
			args:    []string{"add", "--name", "Again", "--username", "again", "A@Example.test"},
			wantErr: "duplicate address",
		},
		{
			name:    "a duplicate username in another case",
			args:    []string{"add", "--name", "Other", "--username", "ALICE", "b@example.test"},
			wantErr: "duplicate username",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ, stateDir := memberEnviron(t)
			if _, stderr, err := runMember(t, environ, "add", "--admin", "--name", "Alice", "--username", "alice", "a@example.test"); err != nil {
				t.Fatalf("fixture admin add error = %v; stderr = %q", err, stderr)
			}

			_, _, err := runMember(t, environ, tt.args...)
			if err == nil {
				t.Fatalf("Execute() error = nil, want a refusal naming %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
			if strings.Contains(err.Error(), "b@example.test") {
				t.Errorf("error = %q, want it never to carry the refused address", err)
			}
			if got := userAddresses(t, stateDir); len(got) != 1 || got[0] != "a@example.test" {
				t.Errorf("users after a refusal = %v, want only the fixture admin", got)
			}
		})
	}
}

// TestMemberAddCmd_NoAdminRefusedIntoEmptyStore proves a non-admin add
// into a store holding no member refuses naming member add --admin and
// writes no row.
func TestMemberAddCmd_NoAdminRefusedIntoEmptyStore(t *testing.T) {
	environ, stateDir := memberEnviron(t)

	_, _, err := runMember(t, environ, "add", "--name", "Carol", "--username", "carol", "c@example.test")
	if err == nil {
		t.Fatal("Execute() error = nil, want the no-admin refusal")
	}
	for _, want := range []string{"must name at least one admin", "member add --admin"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
	if got := userAddresses(t, stateDir); len(got) != 0 {
		t.Errorf("users after the refusal = %v, want none", got)
	}
}

func TestMemberAddCmd_MissingRequiredFlagsRefusedBeforeOpen(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing --name", args: []string{"add", "--username", "a", "a@example.test"}, want: `required flag(s) "name" not set`},
		{name: "missing --username", args: []string{"add", "--name", "A", "a@example.test"}, want: `required flag(s) "username" not set`},
		{name: "the retired --streamer-username", args: []string{"add", "--name", "A", "--streamer-username", "a", "a@example.test"}, want: "unknown flag: --streamer-username"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := serveAuthEnviron(t)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")

			_, _, err := runMember(t, environ, tt.args...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Execute() error = %v, want it to contain %q", err, tt.want)
			}
			if _, statErr := os.Stat(filepath.Join(stateDir, "store")); !errors.Is(statErr, os.ErrNotExist) {
				t.Errorf("Stat(store leaf) error = %v, want ErrNotExist: a missing required flag must refuse before any open", statErr)
			}
		})
	}
}

func TestMemberAddCmd_NoPasswdEntryFallsBackToUID(t *testing.T) {
	orig := osUserLookupID
	defer func() { osUserLookupID = orig }()
	osUserLookupID = func(string) (*user.User, error) {
		return nil, errors.New("no such user")
	}
	environ, _ := memberEnviron(t)

	_, stderr, err := runMember(t, environ, "add", "--admin", "--name", "A", "--username", "a", "a@example.test")
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil; stderr = %q", err, stderr)
	}
	wantActor := fmt.Sprintf("actor=uid:%d", os.Geteuid())
	if !strings.Contains(stderr, wantActor) {
		t.Errorf("audit record = %q, want it to contain %q", stderr, wantActor)
	}
}

// TestMemberAddCmd_LockHeldPastDeadlineRefuses holds the store's lock
// exclusively, as a restore would, and proves the verb gives up once its
// own wait ends rather than waiting on the holder.
func TestMemberAddCmd_LockHeldPastDeadlineRefuses(t *testing.T) {
	orig := memberVerbWriteWait
	memberVerbWriteWait = 200 * time.Millisecond
	defer func() { memberVerbWriteWait = orig }()

	environ, stateDir := memberEnviron(t)
	lock, err := os.Open(filepath.Join(stateDir, "store", "store.lock"))
	if err != nil {
		t.Fatalf("open store.lock: %v", err)
	}
	defer func() { _ = lock.Close() }()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock store.lock: %v", err)
	}

	start := time.Now()
	_, _, err = runMember(t, environ, "add", "--admin", "--name", "B", "--username", "b", "b@example.test")
	elapsed := time.Since(start)
	if !errors.Is(err, store.ErrLockBusy) {
		t.Fatalf("Execute() error = %v, want it to wrap store.ErrLockBusy", err)
	}
	if elapsed < 200*time.Millisecond || elapsed > 10*time.Second {
		t.Errorf("the verb gave up after %v, want just past the 200ms wait", elapsed)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatalf("release store.lock: %v", err)
	}
	if got := userAddresses(t, stateDir); len(got) != 0 {
		t.Errorf("users after the refusal = %v, want none", got)
	}
}

// TestMemberAddCmd_ConcurrentAdminAddsBothLand runs two --admin adds for
// distinct addresses at once against one store, under -race: both land
// and each logs its own audit record.
func TestMemberAddCmd_ConcurrentAdminAddsBothLand(t *testing.T) {
	environ, stateDir := memberEnviron(t)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	var stderrs [2]string
	for i, member := range []struct{ address, name, username string }{
		{"a@example.test", "A", "a"},
		{"b@example.test", "B", "b"},
	} {
		wg.Go(func() {
			_, stderrs[i], errs[i] = runMember(t, environ, "add", "--admin", "--name", member.name, "--username", member.username, member.address)
		})
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("add #%d: %v; stderr = %q", i, err, stderrs[i])
		}
		if !strings.Contains(stderrs[i], "roster member added") {
			t.Errorf("add #%d stderr = %q, want one audit record", i, stderrs[i])
		}
	}
	got := userAddresses(t, stateDir)
	if len(got) != 2 || got[0] != "a@example.test" || got[1] != "b@example.test" {
		t.Errorf("users = %v, want both a@example.test and b@example.test", got)
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

// TestMemberAddCmd_WideAuthLeafRefused proves member add opens the auth
// leaf for the re-add bump and refuses a leaf wider than 0700, naming the
// verb and the leaf, before writing any row. Fixture state: an auth leaf
// created at 0755 under a current store holding no member.
func TestMemberAddCmd_WideAuthLeafRefused(t *testing.T) {
	environ, stateDir := memberEnviron(t)
	leaf := filepath.Join(stateDir, "auth")
	if err := os.Mkdir(leaf, 0o700); err != nil {
		t.Fatalf("create the auth leaf: %v", err)
	}
	if err := os.Chmod(leaf, 0o755); err != nil {
		t.Fatalf("widen the auth leaf: %v", err)
	}

	_, _, err := runMember(t, environ, "add", "--admin", "--name", "Alice", "--username", "alice", "a@example.test")
	if err == nil {
		t.Fatal("add over a wide auth leaf error = nil, want a refusal")
	}
	for _, want := range []string{"member add", leaf} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
	if got := userAddresses(t, stateDir); len(got) != 0 {
		t.Errorf("users after the refusal = %v, want none", got)
	}
}

// TestMemberListCmd_OpensNoAuthLeaf proves member list reads the roster
// without opening the auth leaf: a store with no leaf still has none
// after a list.
func TestMemberListCmd_OpensNoAuthLeaf(t *testing.T) {
	environ, stateDir := memberEnviron(t)
	leaf := filepath.Join(stateDir, "auth")
	if _, err := os.Stat(leaf); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat(auth leaf) before list = %v, want it absent", err)
	}
	if _, _, err := runMember(t, environ, "list"); err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, err := os.Stat(leaf); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Stat(auth leaf) after list = %v, want it still absent", err)
	}
}
