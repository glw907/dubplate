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

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/spf13/cobra"
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
		{name: "remove", args: []string{"remove", "a@example.test"}},
		{name: "demote", args: []string{"demote", "a@example.test"}},
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
	orig := operatorVerbWriteWait
	operatorVerbWriteWait = 200 * time.Millisecond
	defer func() { operatorVerbWriteWait = orig }()

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

// rosterFixture returns an environment whose store holds two admins,
// alice and carol, and one member, bob, added through the verb. Bob
// carries one row in each of member_preferences, member_overrides, and
// member_quota, so a removal that touched the wrong table, or none,
// shows.
func rosterFixture(t *testing.T) (environ []string, stateDir string) {
	t.Helper()
	environ, stateDir = memberEnviron(t)
	for _, args := range [][]string{
		{"add", "--admin", "--name", "Alice", "--username", "alice", "a@example.test"},
		{"add", "--admin", "--name", "Carol", "--username", "carol", "c@example.test"},
		{"add", "--name", "Bob", "--username", "bob", "b@example.test"},
	} {
		if _, stderr, err := runMember(t, environ, args...); err != nil {
			t.Fatalf("%v error = %v; stderr = %q", args, err, stderr)
		}
	}
	path := storeDBPath(stateDir)
	for _, q := range []string{
		`INSERT INTO member_preferences (user_id, collapse_duplicates, updated_at) SELECT id, 1, 1 FROM users WHERE address = 'b@example.test'`,
		`INSERT INTO member_overrides (user_id, quota_bytes, updated_at) SELECT id, 1000, 1 FROM users WHERE address = 'b@example.test'`,
		`INSERT INTO member_quota (user_id, charged_bytes, updated_at) SELECT id, 500, 1 FROM users WHERE address = 'b@example.test'`,
	} {
		execStoreSQL(t, path, q)
	}
	return environ, stateDir
}

// countRows returns the number query selects from the store under
// stateDir.
func countRows(t *testing.T, stateDir, query string) int {
	t.Helper()
	h, err := store.OpenWrite(t.Context(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	var n int
	if err := h.QueryRow(t.Context(), query).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

// hasEpoch reports whether the auth leaf under stateDir holds a
// revocation epoch for address.
func hasEpoch(t *testing.T, stateDir, address string) bool {
	t.Helper()
	leaf, err := auth.Open(stateDir)
	if err != nil {
		t.Fatalf("auth.Open: %v", err)
	}
	_, ok, err := auth.OpenEpochStore(leaf).Read(address)
	if err != nil {
		t.Fatalf("read the epoch of %s: %v", address, err)
	}
	return ok
}

// TestMemberRemoveCmd_TombstonesAndRevokes is AC15's command-line half:
// the removal tombstones the row, deletes the member's preferences,
// keeps their override and quota rows, bumps their address's epoch, and
// leaves its record on the terminal. Fixture: bob holds one row in each
// of the three tables, and alice holds none, so an epoch bump for the
// wrong address would show.
func TestMemberRemoveCmd_TombstonesAndRevokes(t *testing.T) {
	environ, stateDir := rosterFixture(t)

	stdout, stderr, err := runMember(t, environ, "remove", "B@Example.test")
	if err != nil {
		t.Fatalf("remove error = %v; stderr = %q", err, stderr)
	}

	for _, want := range []string{"roster member removed", "address=b@example.test", "channel=command-line", "client_ip=local"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("audit record on stderr = %q, want it to contain %q", stderr, want)
		}
	}
	if got := countRows(t, stateDir, `SELECT count(*) FROM users WHERE address = 'b@example.test' AND tombstoned_at IS NOT NULL`); got != 1 {
		t.Errorf("tombstoned rows for bob = %d, want 1", got)
	}
	for table, want := range map[string]int{"member_preferences": 0, "member_overrides": 1, "member_quota": 1} {
		if got := countRows(t, stateDir, "SELECT count(*) FROM "+table); got != want {
			t.Errorf("%s rows = %d, want %d", table, got, want)
		}
	}
	if !hasEpoch(t, stateDir, "b@example.test") {
		t.Error("no revocation epoch for the removed address, want one")
	}
	if hasEpoch(t, stateDir, "a@example.test") {
		t.Error("a revocation epoch for alice, who was not removed")
	}
	list, _, err := runMember(t, environ, "list")
	if err != nil || strings.Contains(list, "b@example.test") || !strings.Contains(list, "a@example.test") {
		t.Errorf("list after the removal = %q, %v, want alice and no bob", list, err)
	}
	if stdout != "" {
		t.Errorf("remove stdout = %q, want none", stdout)
	}
}

// TestMemberRemoveCmd_NotFound proves an unknown address and an address
// already removed answer not found and write nothing. Fixture: bob
// removed once before the second attempt.
func TestMemberRemoveCmd_NotFound(t *testing.T) {
	environ, stateDir := rosterFixture(t)
	if _, _, err := runMember(t, environ, "remove", "b@example.test"); err != nil {
		t.Fatalf("first remove: %v", err)
	}

	for _, address := range []string{"b@example.test", "nobody@example.test", "not an address"} {
		_, stderr, err := runMember(t, environ, "remove", address)
		if !errors.Is(err, auth.ErrMemberNotFound) {
			t.Errorf("remove %q error = %v, want ErrMemberNotFound", address, err)
		}
		if stderr != "" {
			t.Errorf("remove %q stderr = %q, want no audit record", address, stderr)
		}
	}
	if got := countRows(t, stateDir, `SELECT count(*) FROM users WHERE tombstoned_at IS NOT NULL`); got != 1 {
		t.Errorf("tombstoned rows = %d, want only bob", got)
	}
}

// TestMemberDemoteCmd clears the flag of a live admin and writes
// nothing for a member who holds none. Fixture: two admins and a
// member.
func TestMemberDemoteCmd(t *testing.T) {
	environ, stateDir := rosterFixture(t)

	_, stderr, err := runMember(t, environ, "demote", "c@example.test")
	if err != nil {
		t.Fatalf("demote error = %v; stderr = %q", err, stderr)
	}
	for _, want := range []string{"roster member demoted", "address=c@example.test", "channel=command-line"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("audit record on stderr = %q, want it to contain %q", stderr, want)
		}
	}
	if got := countRows(t, stateDir, `SELECT count(*) FROM users WHERE admin = 1`); got != 1 {
		t.Errorf("admins after the demotion = %d, want only alice", got)
	}
	if hasEpoch(t, stateDir, "c@example.test") {
		t.Error("a revocation epoch for the demoted admin, want none: a demotion ends no session")
	}

	_, stderr, err = runMember(t, environ, "demote", "b@example.test")
	if err != nil || stderr != "" {
		t.Errorf("demote of a member with no flag = %v, stderr %q, want nil and no record", err, stderr)
	}
	if _, _, err := runMember(t, environ, "demote", "nobody@example.test"); !errors.Is(err, auth.ErrMemberNotFound) {
		t.Errorf("demote of an unknown address error = %v, want ErrMemberNotFound", err)
	}
}

// TestMemberCmd_LastAdminRefused is AC16's command-line half: a removal
// or demotion that would leave no live admin is refused and changes
// nothing. Fixture: alice is the only admin, bob a member.
func TestMemberCmd_LastAdminRefused(t *testing.T) {
	for _, verb := range []string{"remove", "demote"} {
		t.Run(verb, func(t *testing.T) {
			environ, stateDir := memberEnviron(t)
			for _, args := range [][]string{
				{"add", "--admin", "--name", "Alice", "--username", "alice", "a@example.test"},
				{"add", "--name", "Bob", "--username", "bob", "b@example.test"},
			} {
				if _, stderr, err := runMember(t, environ, args...); err != nil {
					t.Fatalf("%v error = %v; stderr = %q", args, err, stderr)
				}
			}

			_, stderr, err := runMember(t, environ, verb, "a@example.test")

			if !errors.Is(err, auth.ErrNoAdmin) {
				t.Fatalf("%s error = %v, want ErrNoAdmin", verb, err)
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want no audit record for a refusal", stderr)
			}
			if got := countRows(t, stateDir, `SELECT count(*) FROM users WHERE admin = 1 AND tombstoned_at IS NULL`); got != 1 {
				t.Errorf("live admins after the refusal = %d, want 1", got)
			}
			if hasEpoch(t, stateDir, "a@example.test") {
				t.Error("a revocation epoch for the refused removal's target, want none")
			}
		})
	}
}

// TestMemberCmd_WideAuthLeafRefusesRemoveAndDemote proves both verbs
// open the auth leaf and refuse one wider than 0700, naming the verb
// and the leaf, before writing. Fixture: a leaf created at 0755 under a
// store holding two admins and a member.
func TestMemberCmd_WideAuthLeafRefusesRemoveAndDemote(t *testing.T) {
	for _, tt := range []struct{ verb, address string }{
		{"remove", "b@example.test"},
		{"demote", "c@example.test"},
	} {
		t.Run(tt.verb, func(t *testing.T) {
			environ, stateDir := rosterFixture(t)
			leaf := filepath.Join(stateDir, "auth")
			if err := os.Chmod(leaf, 0o755); err != nil {
				t.Fatalf("widen the auth leaf the fixture's adds created: %v", err)
			}

			_, _, err := runMember(t, environ, tt.verb, tt.address)

			if err == nil {
				t.Fatalf("%s over a wide auth leaf error = nil, want a refusal", tt.verb)
			}
			for _, want := range []string{"member " + tt.verb, leaf} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err, want)
				}
			}
			if got := countRows(t, stateDir, `SELECT count(*) FROM users WHERE tombstoned_at IS NOT NULL`); got != 0 {
				t.Errorf("tombstoned rows after the refused %s = %d, want none", tt.verb, got)
			}
			if got := countRows(t, stateDir, `SELECT count(*) FROM users WHERE admin = 1`); got != 2 {
				t.Errorf("admins after the refused %s = %d, want both", tt.verb, got)
			}
		})
	}
}

// TestOperatorVerbs_RequiredFlagsSaySo is BACKLOG #12: every flag a
// verb marks required says "(required)" in its usage. member add's
// --name and --username are the required flags; remove, demote, and
// settings reset take their address or key as an argument and carry
// no flag at all, which the test asserts so a flag added later must
// join the table.
func TestOperatorVerbs_RequiredFlagsSaySo(t *testing.T) {
	root := newRootCmd()
	find := func(path ...string) *cobra.Command {
		t.Helper()
		cmd, _, err := root.Find(path)
		if err != nil || cmd.Name() != path[len(path)-1] {
			t.Fatalf("find %v = %v, %v", path, cmd, err)
		}
		return cmd
	}

	add := find("member", "add")
	for _, name := range []string{"name", "username"} {
		f := add.Flag(name)
		if f == nil {
			t.Fatalf("member add has no --%s", name)
		}
		if _, required := f.Annotations[cobra.BashCompOneRequiredFlag]; !required {
			t.Errorf("member add --%s is not marked required", name)
		}
		if !strings.Contains(f.Usage, "(required)") {
			t.Errorf("member add --%s usage = %q, want it to say (required)", name, f.Usage)
		}
	}
	if f := add.Flag("admin"); f == nil || strings.Contains(f.Usage, "(required)") {
		t.Errorf("member add --admin = %v, want an optional flag", f)
	}
	for _, path := range [][]string{{"member", "remove"}, {"member", "demote"}, {"settings", "reset"}} {
		if cmd := find(path...); cmd.LocalFlags().HasFlags() {
			t.Errorf("%v carries flags %q, want none: add any required one to this table", path, cmd.LocalFlags().FlagUsages())
		}
	}
}
