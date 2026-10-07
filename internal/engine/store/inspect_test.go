package store

import (
	"context"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

// childActionEnv and childStateDirEnv tell a re-executed test binary
// which write TestInspectChildWrite makes, and on which store.
const (
	childActionEnv   = "DUBPLATE_STORE_TEST_CHILD"
	childStateDirEnv = "DUBPLATE_STORE_TEST_CHILD_STATE_DIR"
)

// TestInspectChildWrite is not a test on its own: it does nothing
// unless a parent test re-executes the binary naming it. The child
// makes the write that decides the store's state with automatic
// checkpoints off, and exits without closing, so the write lives only
// in the -wal the way a crashed or still-running process leaves it.
func TestInspectChildWrite(t *testing.T) {
	action := os.Getenv(childActionEnv)
	if action == "" {
		return
	}
	l := newLeaf(os.Getenv(childStateDirEnv))
	h, err := openHandle(l.dbPath, openBusy)
	if err != nil {
		t.Fatalf("child open: %v", err)
	}
	ctx := context.Background()
	if _, err := h.write.ExecContext(ctx, "PRAGMA wal_autocheckpoint = 0"); err != nil {
		t.Fatalf("child autocheckpoint off: %v", err)
	}
	switch action {
	case "older":
		if err := migrate(ctx, h, l, readMigrationsT(t, migrationSourceThrough(t, 2)), defaultOpts()); err != nil {
			t.Fatalf("child migrate: %v", err)
		}
	case "newer":
		if _, err := h.write.ExecContext(ctx, futureLedgerRow); err != nil {
			t.Fatalf("child newer: %v", err)
		}
	default:
		t.Fatalf("child action %q", action)
	}
	os.Exit(0)
}

// runChildWrite re-executes the test binary to make action's write on
// stateDir's store, leaving it in a hot -wal.
func runChildWrite(t *testing.T, stateDir, action string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestInspectChildWrite$", "-test.count=1")
	cmd.Env = append(os.Environ(), childActionEnv+"="+action, childStateDirEnv+"="+stateDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child %s: %v\n%s", action, err, out)
	}
	info, err := os.Stat(newLeaf(stateDir).dbPath + "-wal")
	if err != nil || info.Size() == 0 {
		t.Fatalf("child %s left no hot -wal: %v", action, err)
	}
}

// TestInspect_HotWAL is spec criterion 15's in-package half: inspect
// reports an absent, an older, and a newer store by name and leaves the
// database file's bytes and mode unchanged. The older and newer stores
// carry the write that decides their state only in a hot -wal, so the
// reported version proves inspect read through it, and a read that
// checkpoints on close would change the main file.
func TestInspect_HotWAL(t *testing.T) {
	tests := []struct {
		name        string
		through     int
		child       string
		wantState   State
		wantVersion int
		wantPending []string
	}{
		{name: "absent", wantState: StateAbsent},
		{name: "older", through: 1, child: "older", wantState: StateOlder, wantVersion: 2, wantPending: []string{"0003_catalog.sql", "0004_settings.sql", "0005_routed.sql"}},
		{name: "newer", through: binaryHead, child: "newer", wantState: StateNewer, wantVersion: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			if tt.through == 0 {
				r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
				if r.State != tt.wantState {
					t.Errorf("state = %v (%s), want %v", r.State, r.Reason, tt.wantState)
				}
				if left := listDir(t, stateDir); len(left) != 0 {
					t.Errorf("inspect created %q", left)
				}
				return
			}
			l := buildStore(t, stateDir, tt.through)
			runChildWrite(t, stateDir, tt.child)
			hash := hashFile(t, l.dbPath)
			before, err := os.Stat(l.dbPath)
			if err != nil {
				t.Fatal(err)
			}

			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.State != tt.wantState {
				t.Errorf("state = %v (%s), want %v", r.State, r.Reason, tt.wantState)
			}
			if r.StoreVersion != tt.wantVersion || r.BinaryVersion != binaryHead {
				t.Errorf("versions = store %d, binary %d; want %d, %d", r.StoreVersion, r.BinaryVersion, tt.wantVersion, binaryHead)
			}
			if !slices.Equal(r.Pending, tt.wantPending) {
				t.Errorf("pending = %q, want %q", r.Pending, tt.wantPending)
			}
			if hashFile(t, l.dbPath) != hash {
				t.Error("the database file's bytes changed")
			}
			after, err := os.Stat(l.dbPath)
			if err != nil {
				t.Fatal(err)
			}
			if after.Mode() != before.Mode() {
				t.Errorf("mode = %v, want %v unchanged", after.Mode(), before.Mode())
			}
		})
	}
}

// TestInspect_EveryOtherState produces each remaining report state once
// from a real fixture.
func TestInspect_EveryOtherState(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, stateDir string)
		euid       int
		wantState  State
		wantReason string
	}{
		{
			name: "restore in progress",
			setup: func(t *testing.T, stateDir string) {
				l := buildStore(t, stateDir, binaryHead)
				if err := os.WriteFile(l.restoreIntentPath, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantState:  StateRestoreInProgress,
			wantReason: "restore --resume",
		},
		{
			name: "busy",
			setup: func(t *testing.T, stateDir string) {
				takeRawExclusive(t, buildStore(t, stateDir, binaryHead).lockPath)
			},
			wantState:  StateBusy,
			wantReason: "held by",
		},
		{
			name: "not a store",
			setup: func(t *testing.T, stateDir string) {
				l := newLeaf(stateDir)
				if err := l.create(stateDir); err != nil {
					t.Fatal(err)
				}
				precreateFile(t, l.dbPath)
			},
			wantState:  StateNotAStore,
			wantReason: "start serve to finish creating it",
		},
		{
			name:       "refused, owner",
			setup:      func(t *testing.T, stateDir string) { buildStore(t, stateDir, binaryHead) },
			euid:       os.Geteuid() + 1,
			wantState:  StateRefused,
			wantReason: "runuser -u",
		},
		{
			name:      "current",
			setup:     func(t *testing.T, stateDir string) { buildStore(t, stateDir, binaryHead) },
			wantState: StateCurrent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			tt.setup(t, stateDir)
			if tt.euid != 0 {
				withGeteuid(t, tt.euid)
			}

			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.State != tt.wantState {
				t.Errorf("state = %v (%s), want %v", r.State, r.Reason, tt.wantState)
			}
			if !strings.Contains(r.Reason, tt.wantReason) {
				t.Errorf("reason %q does not name %q", r.Reason, tt.wantReason)
			}
			if tt.wantState == StateCurrent && r.StoreVersion != binaryHead {
				t.Errorf("store version = %d, want %d", r.StoreVersion, binaryHead)
			}
		})
	}
}

// TestInspect_AbsentLeafSkipsOwnerCheck: a state directory with no
// leaf is absent whatever uid asks, since there is nothing yet to own.
// The euid differs from the state directory's owner, so an owner check
// on the absent branch would refuse.
func TestInspect_AbsentLeafSkipsOwnerCheck(t *testing.T) {
	stateDir := t.TempDir()
	withGeteuid(t, os.Geteuid()+1)
	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateAbsent {
		t.Errorf("state = %v (%s), want absent", r.State, r.Reason)
	}
}

// TestInspect_FUSEWarns: a store on FUSE opens and reports its state,
// carrying a warning that names FUSE.
func TestInspect_FUSEWarns(t *testing.T) {
	stateDir := t.TempDir()
	buildStore(t, stateDir, binaryHead)
	orig := statfsFunc
	t.Cleanup(func() { statfsFunc = orig })
	statfsFunc = func(_ string, st *syscall.Statfs_t) error {
		st.Type = fuseSuperMagic
		return nil
	}

	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateCurrent {
		t.Errorf("state = %v (%s), want current", r.State, r.Reason)
	}
	if !strings.Contains(r.Warning, "FUSE") {
		t.Errorf("warning %q does not name FUSE", r.Warning)
	}
}

// TestStateString pins each state's printed name, which status lines
// carry.
func TestStateString(t *testing.T) {
	want := map[State]string{
		StateAbsent:            "absent",
		StateRestoreInProgress: "restore in progress",
		StateBusy:              "busy",
		StateNotAStore:         "not a store",
		StateRefused:           "refused",
		StateOlder:             "older",
		StateNewer:             "newer",
		StateCurrent:           "current",
		0:                      "unknown",
	}
	for s, name := range want {
		if got := s.String(); got != name {
			t.Errorf("State(%d).String() = %q, want %q", int(s), got, name)
		}
	}
}

// insertFinished inserts one terminal job record through h.
func insertFinished(t *testing.T, h *Handle, kind, state string, finished int64, progress any) {
	t.Helper()
	err := h.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx,
			"INSERT INTO jobs (id, kind, at, state, owner_instance, actor, progress, updated_at, finished_at) VALUES (?, ?, 1, ?, ?, 'local:test', ?, ?, ?)",
			NewID(), kind, state, testInstance, progress, finished, finished)
		return err
	})
	if err != nil {
		t.Fatalf("insert %s %s job: %v", kind, state, err)
	}
}

// TestInspect_Facts: a current store reports its newest succeeded job
// per kind, its -wal size, the free space on its filesystem, and the
// count and size of its copies, a backup reference not among them.
// Fixture: a write handle held open, so the -wal holds the inserts.
func TestInspect_Facts(t *testing.T) {
	stateDir := currentStore(t)
	l := newLeaf(stateDir)
	h := writeHandle(t, stateDir)
	// 2026-09-24T12:00:00Z and a day later, in microseconds.
	const older, newest = 1790251200000000, 1790337600000000
	insertFinished(t, h, "backup", "succeeded", older, "an earlier run")
	insertFinished(t, h, "backup", "succeeded", newest, "prune skipped: a remote stamp is not older than this run's")
	insertFinished(t, h, "backup", "failed", newest+1, "a later failure")
	insertFinished(t, h, "scan", "succeeded", older, nil)

	if err := os.WriteFile(filepath.Join(l.snapshotsDir, "pre-migrate-v1.db"), []byte("five!"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.snapshotsDir, "pre-restore-20260101T000000.000000Z.db"), []byte("seven!!"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.snapshotsDir, "notes.txt"), []byte("not a copy"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := h.WriteBackupRef("20260925T120000.000000Z", "remote:path"); err != nil {
		t.Fatalf("WriteBackupRef: %v", err)
	}

	wal, err := os.Stat(l.dbPath + "-wal")
	if err != nil || wal.Size() == 0 {
		t.Fatalf("fixture needs a non-empty -wal: %v", err)
	}
	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateCurrent || r.Facts == nil {
		t.Fatalf("inspect = %v (%s), facts %v, want current with facts", r.State, r.Reason, r.Facts)
	}
	f := r.Facts
	want := map[string]JobSuccess{
		"backup": {Finished: time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC), Progress: "prune skipped: a remote stamp is not older than this run's"},
		"scan":   {Finished: time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)},
	}
	if !maps.EqualFunc(f.LastSuccess, want, func(a, b JobSuccess) bool {
		return a.Finished.Equal(b.Finished) && a.Finished.Location() == time.UTC && a.Progress == b.Progress
	}) {
		t.Errorf("last success = %v, want %v", f.LastSuccess, want)
	}
	if f.WALSize != wal.Size() {
		t.Errorf("-wal size = %d, want %d", f.WALSize, wal.Size())
	}
	if !f.FreeKnown || f.FreeBytes == 0 {
		t.Errorf("free space = %d (known %v), want a positive known count", f.FreeBytes, f.FreeKnown)
	}
	if f.Copies != 2 || f.CopiesSize != 12 {
		t.Errorf("copies = %d totaling %d bytes, want 2 totaling 12", f.Copies, f.CopiesSize)
	}
}

// TestInspect_FactsEdges: a failed statfs reads unknown without
// failing the inspection, no copy reads 0 and 0, and a store with no
// succeeded job reports none.
func TestInspect_FactsEdges(t *testing.T) {
	stateDir := currentStore(t)
	l := newLeaf(stateDir)
	orig := statfsFunc
	t.Cleanup(func() { statfsFunc = orig })
	statfsFunc = func(path string, st *syscall.Statfs_t) error {
		if path == l.root {
			return syscall.EIO
		}
		return orig(path, st)
	}

	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateCurrent || r.Facts == nil {
		t.Fatalf("inspect = %v (%s), facts %v, want current with facts", r.State, r.Reason, r.Facts)
	}
	if f := *r.Facts; f.FreeKnown || f.FreeBytes != 0 || f.Copies != 0 || f.CopiesSize != 0 || len(f.LastSuccess) != 0 {
		t.Errorf("facts = %+v, want free space unknown, no copies, and no success", f)
	}
}

// TestReadFacts_AbsentWAL: a store with no -wal reads a size of 0.
// Inspect's own read-only connection creates an empty -wal before the
// facts are read, so the absent state is met here directly.
func TestReadFacts_AbsentWAL(t *testing.T) {
	stateDir := currentStore(t)
	l := newLeaf(stateDir)
	// The last close checkpoints the store and leaves a zero-byte -wal,
	// which no connection now holds.
	if err := os.Remove(l.dbPath + "-wal"); err != nil {
		t.Fatalf("remove the empty -wal: %v", err)
	}

	f, err := readFacts(context.Background(), l, openBusy, binaryHead)
	if err != nil {
		t.Fatalf("readFacts: %v", err)
	}
	if f.WALSize != 0 {
		t.Errorf("-wal size = %d, want 0", f.WALSize)
	}
}

// TestInspect_NewerStoreHasNoFacts: this binary cannot trust a newer
// store's schema, so it reads none of its facts.
func TestInspect_NewerStoreHasNoFacts(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	rawExec(t, l.dbPath, futureLedgerRow)

	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateNewer {
		t.Fatalf("inspect = %v (%s), want newer", r.State, r.Reason)
	}
	if r.Facts != nil {
		t.Errorf("facts = %+v, want none for a newer store", r.Facts)
	}
}

// insertSetting inserts one settings row and bumps the generation by
// raw SQL, the shape a fixture below the settings runner's own writer
// plants.
const insertSettingSQL = `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
	VALUES ('DUBPLATE_TEST_KEY', 'seven', 'env-seed', 'seven', 'system', 1790251200000000);
UPDATE settings_generation SET generation = 3 WHERE singleton = 1;`

// TestInspect_SettingsFacts is R-6c10's own table, in-package: a
// current store and an older store whose version holds the settings
// table (0004, with 0005 pending) both read the stored row and the
// generation; an older store below that version reads the table
// absent marker instead, never an empty result mistaken for one.
func TestInspect_SettingsFacts(t *testing.T) {
	tests := []struct {
		name    string
		through int
	}{
		{name: "current", through: binaryHead},
		{name: "older, settings table present, 0005 pending", through: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, tt.through)
			rawExec(t, l.dbPath, insertSettingSQL)

			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.Facts == nil {
				t.Fatalf("inspect = %v (%s), want facts", r.State, r.Reason)
			}
			f := r.Facts
			if f.SettingsAbsent {
				t.Fatalf("SettingsAbsent = true, want false for a store holding the settings table")
			}
			// Fixture state: the store holds one settings row, so an
			// empty result here would pass a bug that never reads the
			// table at all.
			want := []SettingsRow{{Key: "DUBPLATE_TEST_KEY", Value: "seven", Source: "env-seed", SeedValue: "seven", Actor: "system", UpdatedAt: time.UnixMicro(1790251200000000).UTC()}}
			if !slices.Equal(f.Settings, want) {
				t.Errorf("settings = %+v, want %+v", f.Settings, want)
			}
			if f.SettingsGeneration != 3 {
				t.Errorf("settings generation = %d, want 3", f.SettingsGeneration)
			}
		})
	}
}

// TestInspect_SettingsAbsent: a store whose version predates the
// settings table (through 0003) reports it absent, distinct from an
// empty table, and holds no rows or generation.
func TestInspect_SettingsAbsent(t *testing.T) {
	stateDir := t.TempDir()
	buildStore(t, stateDir, 3)

	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.Facts == nil {
		t.Fatalf("inspect = %v (%s), want facts", r.State, r.Reason)
	}
	f := r.Facts
	if !f.SettingsAbsent {
		t.Errorf("SettingsAbsent = false, want true for a store below the settings table's version")
	}
	if len(f.Settings) != 0 || f.SettingsGeneration != 0 {
		t.Errorf("settings = %+v at generation %d, want none for a store predating the table", f.Settings, f.SettingsGeneration)
	}
}

const insertMembersSQL = `INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at) VALUES
	(x'01', 'boss@example.test', 'Boss', 'boss', 1, NULL, 1, 1),
	(x'02', 'plain@example.test', 'Plain', 'plain', 0, NULL, 2, 2),
	(x'03', 'gone-admin@example.test', 'Gone', 'goneadmin', 1, 3, 3, 3),
	(x'04', 'gone-plain@example.test', 'Gone', 'goneplain', 0, 4, 4, 4);`

// TestInspect_MemberFacts: a store whose version holds the users table
// counts its live rows and the live admins among them, tombstoned rows
// excluded from both. Fixture state: the table holds four rows, two of
// them tombstoned and one of those an admin, so a count over every row
// reads 4 and 2. An older store below the table's version reads no
// counts and is not refused for the missing table.
func TestInspect_MemberFacts(t *testing.T) {
	tests := []struct {
		name        string
		through     int
		insert      bool
		wantMembers int
		wantAdmins  int
	}{
		{name: "current", through: binaryHead, insert: true, wantMembers: 2, wantAdmins: 1},
		{name: "older at the users table's own version", through: UsersVersion, insert: true, wantMembers: 2, wantAdmins: 1},
		{name: "current, an empty table", through: binaryHead},
		{name: "older, below the users table", through: UsersVersion - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, tt.through)
			if tt.insert {
				rawExec(t, l.dbPath, insertMembersSQL)
			}

			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.Facts == nil {
				t.Fatalf("inspect = %v (%s), want facts", r.State, r.Reason)
			}
			if r.Facts.Members != tt.wantMembers || r.Facts.Admins != tt.wantAdmins {
				t.Errorf("members, admins = %d, %d, want %d, %d", r.Facts.Members, r.Facts.Admins, tt.wantMembers, tt.wantAdmins)
			}
		})
	}
}
