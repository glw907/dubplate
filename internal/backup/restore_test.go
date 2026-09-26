package backup

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// restoreChildEnv carries a re-executed test binary's restore, as
// JSON, to TestRestoreChild.
const restoreChildEnv = "DUBPLATE_RESTORE_CHILD"

// restoreNow is every new restore's clock in these tests, and
// restoreStamp its stamp in the store's layout.
var restoreNow = time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

const restoreStamp = "20260925T120000.000000Z"

// The users each fixture store holds, one per store, so a row read
// back names the store it came from.
const (
	currentUser  = "current@example.org"
	incomingUser = "incoming@example.org"
)

// testInstance names a fixture's backup copy in its own staging/.
const testInstance = "0123456789abcdef0123456789abcdef"

// A childRestore is one restore a re-executed test binary runs, and
// the step at which it kills itself.
type childRestore struct {
	StateDir    string
	Snapshot    string
	FromFile    string
	Remote      string
	Library     bool
	LibraryRoot string
	Discard     bool
	CrashAt     string
}

// TestRestoreChild is not a test on its own: it does nothing unless a
// parent test re-executes the binary naming it. It runs the restore
// restoreChildEnv describes, with the step hook sending the process
// SIGKILL once the step CrashAt names is done, and a SIGINT canceling
// the restore. It exits 1, printing the error, when the restore fails.
func TestRestoreChild(t *testing.T) {
	raw := os.Getenv(restoreChildEnv)
	if raw == "" {
		return
	}
	var c childRestore
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatalf("child config: %v", err)
	}
	stepHook = func(step string) {
		if step == c.CrashAt {
			_ = syscall.Kill(os.Getpid(), syscall.SIGKILL)
			select {}
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := Restore(ctx, RestoreConfig{
		StateDir:       c.StateDir,
		BusyTimeout:    testBusy,
		Now:            restoreNow,
		Snapshot:       c.Snapshot,
		FromFile:       c.FromFile,
		Remote:         c.Remote,
		Library:        c.Library,
		LibraryRoot:    c.LibraryRoot,
		DiscardCurrent: c.Discard,
		Out:            os.Stdout,
		Logger:         slog.New(slog.DiscardHandler),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

// testBusy is every test store's busy timeout.
const testBusy = 5 * time.Second

// startChild re-executes the test binary as TestRestoreChild running c,
// in its own process group, and returns the started command and its
// stderr.
func startChild(t *testing.T, c childRestore) (*exec.Cmd, *bytes.Buffer) {
	t.Helper()
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestRestoreChild$", "-test.count=1")
	cmd.Env = append(os.Environ(), restoreChildEnv+"="+string(raw))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	return cmd, &stderr
}

// crashChild runs c in a re-executed child and fails the test unless
// the child died by the SIGKILL its step hook sent at c.CrashAt.
func crashChild(t *testing.T, c childRestore) {
	t.Helper()
	cmd, stderr := startChild(t, c)
	err := cmd.Wait()
	ee, ok := errors.AsType[*exec.ExitError](err)
	if !ok {
		t.Fatalf("child crashing at %s: %v, want death by SIGKILL\n%s", c.CrashAt, err, stderr)
	}
	ws, ok := ee.Sys().(syscall.WaitStatus)
	if !ok || !ws.Signaled() || ws.Signal() != syscall.SIGKILL {
		t.Fatalf("child crashing at %s ended %v, want death by SIGKILL\n%s", c.CrashAt, err, stderr)
	}
}

// leafPaths are the store leaf's paths under a state directory, as the
// store lays them out.
type leafPaths struct {
	db, staging, snapshots, intent, report string
}

func leafOf(stateDir string) leafPaths {
	root := filepath.Join(stateDir, "store")
	return leafPaths{
		db:        filepath.Join(root, "db", "dubplate.db"),
		staging:   filepath.Join(root, "staging"),
		snapshots: filepath.Join(root, "snapshots"),
		intent:    filepath.Join(root, "restore-intent"),
		report:    filepath.Join(root, "restore-report-"+restoreStamp+".txt"),
	}
}

// ctxFor returns a context bounded for one test call.
func ctxFor(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// storeWithUser creates a current store under a fresh state directory
// holding one user, address, and returns the directory with no handle
// open.
func storeWithUser(t *testing.T, address string) string {
	t.Helper()
	dir := t.TempDir()
	h, err := store.OpenServe(ctxFor(t), dir, store.ServeOptions{
		BusyTimeout:   testBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	err = h.Update(ctxFor(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, streamer_username, created_at, updated_at)
VALUES (?, ?, 'User', ?, 1, 1)`, store.NewID(), address, strings.Split(address, "@")[0])
		return err
	})
	if err != nil {
		t.Fatalf("plant user %s: %v", address, err)
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	return dir
}

// snapshotFile returns a verified backup copy of a store holding the
// one user incomingUser: a rollback-journal database, as every
// snapshot is.
func snapshotFile(t *testing.T) string {
	t.Helper()
	dir := storeWithUser(t, incomingUser)
	h, err := store.OpenWrite(ctxFor(t), dir, testBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	p, err := h.BackupCopy(ctxFor(t), slog.New(slog.DiscardHandler), testInstance)
	if err != nil {
		t.Fatalf("BackupCopy: %v", err)
	}
	if got := headerBytes(t, p); got != [2]byte{1, 1} {
		t.Fatalf("snapshot fixture header %v, want rollback-journal 1/1", got)
	}
	return p
}

// walFile returns the database file of a cleanly closed store holding
// the one user incomingUser: a WAL-header file with no -wal beside it,
// which a read creates a -wal and -shm beside.
func walFile(t *testing.T) string {
	t.Helper()
	p := leafOf(storeWithUser(t, incomingUser)).db
	if got := headerBytes(t, p); got != [2]byte{2, 2} {
		t.Fatalf("WAL fixture header %v, want 2/2", got)
	}
	if _, err := os.Stat(p + "-wal"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("WAL fixture has a -wal: %v", err)
	}
	return p
}

// headerBytes returns a database header's write and read versions: 1
// and 1 for a rollback journal, 2 and 2 for WAL.
func headerBytes(t *testing.T, path string) [2]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return [2]byte{b[18], b[19]}
}

// rawDB opens path with no journal-mode or version handling of its
// own, closed at cleanup.
func rawDB(t *testing.T, path string, readOnly bool) *sql.DB {
	t.Helper()
	dsn := (&url.URL{Scheme: "file", Path: path}).String()
	if readOnly {
		dsn += "?mode=ro"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// users returns every user address in the store file at path.
func users(t *testing.T, path string) []string {
	t.Helper()
	db := rawDB(t, path, true)
	rows, err := db.Query("SELECT address FROM users ORDER BY address")
	if err != nil {
		t.Fatalf("read users in %s: %v", path, err)
	}
	defer func() { _ = rows.Close() }()
	var got []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			t.Fatal(err)
		}
		got = append(got, a)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return got
}

// setVersion rewrites the store file at path to claim version v, with
// the ledger to match: the version below its own stands in for a
// binary one version newer than the one that wrote it, and the version
// above for a binary one version older.
func setVersion(t *testing.T, path string, v int) {
	t.Helper()
	db := rawDB(t, path, false)
	stmts := fmt.Sprintf("DELETE FROM migrations WHERE version > %d; PRAGMA user_version = %d", v, v)
	if v > currentVersion(t, path) {
		stmts = fmt.Sprintf("INSERT INTO migrations (version, name, sha256) VALUES (%d, '%04d_future.sql', '%s'); PRAGMA user_version = %d",
			v, v, strings.Repeat("0", 64), v)
	}
	if _, err := db.Exec(stmts); err != nil {
		t.Fatalf("set %s to version %d: %v", path, v, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// currentVersion returns the store file at path's user_version.
func currentVersion(t *testing.T, path string) int {
	t.Helper()
	var v int
	if err := rawDB(t, path, true).QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

// recordState returns the state of stateDir's restore-intent record,
// or "" when there is none.
func recordState(t *testing.T, stateDir string) string {
	t.Helper()
	b, err := os.ReadFile(leafOf(stateDir).intent)
	if errors.Is(err, fs.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("record %s: %v", b, err)
	}
	return r.State
}

// entries returns the names in dir, sorted.
func entries(t *testing.T, dir string) []string {
	t.Helper()
	des, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, de := range des {
		names = append(names, de.Name())
	}
	return names
}

// assertWriteOpen opens stateDir in write mode and fails the test
// unless the open refuses naming refusal, or, with refusal "", opens.
func assertWriteOpen(t *testing.T, stateDir, refusal string) {
	t.Helper()
	h, err := store.OpenWrite(ctxFor(t), stateDir, testBusy)
	if refusal == "" {
		if err != nil {
			t.Fatalf("write-mode open: %v, want it to proceed", err)
		}
		if err := h.Close(); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err == nil {
		_ = h.Close()
		t.Fatalf("write-mode open succeeded, want a refusal naming %q", refusal)
	}
	if !strings.Contains(err.Error(), refusal) {
		t.Errorf("write-mode open: %v, want it to name %q", err, refusal)
	}
}

// assertReadsWAL fails the test unless the store in stateDir, left in
// rollback-journal mode by a quiesce, opens in write mode and reads in
// WAL mode after the open, as the write DSN sets.
func assertReadsWAL(t *testing.T, stateDir string) {
	t.Helper()
	db := leafOf(stateDir).db
	if got := headerBytes(t, db); got != [2]byte{1, 1} {
		t.Errorf("current store header after the quiesce %v, want 1/1", got)
	}
	h, err := store.OpenWrite(ctxFor(t), stateDir, testBusy)
	if err != nil {
		t.Fatalf("write-mode open: %v, want it to proceed", err)
	}
	var mode string
	err = h.QueryRow(ctxFor(t), "PRAGMA journal_mode").Scan(&mode)
	if cerr := h.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil || mode != "wal" {
		t.Errorf("journal_mode after a write-mode open = %q, %v; want wal", mode, err)
	}
	if got := headerBytes(t, db); got != [2]byte{2, 2} {
		t.Errorf("current store header after a write-mode open %v, want 2/2", got)
	}
}

// resume runs restore --resume on stateDir in this process, returning
// its error and what it printed.
func resume(t *testing.T, stateDir string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := Restore(ctxFor(t), RestoreConfig{
		StateDir:    stateDir,
		BusyTimeout: testBusy,
		Now:         restoreNow.Add(time.Hour),
		Resume:      true,
		Out:         &out,
		Logger:      slog.New(slog.DiscardHandler),
	})
	return out.String(), err
}

// A completion is what a finished restore must have left, per the
// completion tuple.
type completion struct {
	// incoming is the user db/'s store must hold.
	incoming string

	// current is the user the pre-restore copy must hold, "" when no
	// copy is owed.
	current string

	// discarded lists the suffixes ("" for the database) the fixture
	// planted in db/ that must now sit under the discarded name.
	discarded []string

	// refs lists the stamps of the backup references snapshots/ must
	// hold.
	refs []string

	// pending is a restore left for serve: the record reconcile-pending
	// and no report.
	pending bool
}

// assertComplete fails the test unless stateDir holds every part of
// want.
func assertComplete(t *testing.T, stateDir string, want completion) {
	t.Helper()
	l := leafOf(stateDir)
	if got := entries(t, l.staging); len(got) != 0 {
		t.Errorf("staging/ holds %q, want it empty", got)
	}
	switch state := recordState(t, stateDir); {
	case want.pending && state != "reconcile-pending":
		t.Errorf("record state %q, want reconcile-pending", state)
	case !want.pending && state != "":
		t.Errorf("record state %q, want the record removed", state)
	}
	info, err := os.Stat(l.report)
	switch {
	case want.pending && !errors.Is(err, fs.ErrNotExist):
		t.Errorf("report beside a reconcile-pending record: %v, want none", err)
	case !want.pending && err != nil:
		t.Errorf("report: %v", err)
	case !want.pending && info.Mode().Perm() != 0o600:
		t.Errorf("report mode %v, want 0600", info.Mode().Perm())
	case !want.pending:
		if b, err := os.ReadFile(l.report); err != nil || !strings.HasPrefix(string(b), "restore "+restoreStamp+"\n") {
			t.Errorf("report holds %q (%v), want this restore's", b, err)
		}
	}

	pre := filepath.Join(l.snapshots, "pre-restore-"+restoreStamp+".db")
	if want.current == "" {
		if _, err := os.Stat(pre); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("pre-restore copy %s: %v, want none", pre, err)
		}
	} else {
		for _, suffix := range []string{"-wal", "-shm"} {
			if _, err := os.Stat(pre + suffix); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("%s beside the pre-restore copy: %v, want none", suffix, err)
			}
		}
		if got := users(t, pre); !slices.Equal(got, []string{want.current}) {
			t.Errorf("pre-restore copy holds %q, want %q", got, want.current)
		}
	}
	discarded := filepath.Join(l.snapshots, "discarded-"+restoreStamp+".db")
	for _, suffix := range want.discarded {
		if _, err := os.Stat(discarded + suffix); err != nil {
			t.Errorf("discarded file: %v", err)
		}
		if suffix != "" {
			if _, err := os.Stat(l.db + suffix); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("planted %s still in db/: %v", l.db+suffix, err)
			}
		}
	}

	var refs []string
	for _, name := range entries(t, l.snapshots) {
		if stamp, ok := strings.CutPrefix(name, "backup-"); ok {
			refs = append(refs, strings.TrimSuffix(stamp, ".ref"))
		}
	}
	if !slices.Equal(refs, want.refs) {
		t.Errorf("backup references %q, want %q", refs, want.refs)
	}
	if got := users(t, l.db); !slices.Equal(got, []string{want.incoming}) {
		t.Errorf("db/ holds %q, want %q", got, want.incoming)
	}
}

// snapshotRemote configures a real rclone local remote holding the
// snapshot file src as snapshots/<stampNew>.db, and returns the remote
// as a restore names it and the remote's directory.
func snapshotRemote(t *testing.T, src string) (string, string) {
	t.Helper()
	_, _, dir := rcloneRemote(t)
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	seedRemoteSnapshots(t, dir, string(b), []string{stampNew})
	return "fetchtest:" + dir, dir
}

// TestRestore_Crashes is spec criterion 7, held from the record on:
// each row kills a re-executed restore by SIGKILL after one step, then
// checks what a write-mode open does beside what the crash left, and
// what restore --resume does. Before the record (step 4) a crash leaves
// the current store usable; from the record on, every write-mode open
// refuses and a resume reaches the completion tuple. Each resume runs
// with a stand-in rclone alone on PATH, so a resume that fetched again
// would record a run.
func TestRestore_Crashes(t *testing.T) {
	type fixture struct {
		stateDir string
		child    childRestore
		want     completion
	}
	fromFile := func(t *testing.T, current bool, src string) fixture {
		f := fixture{stateDir: t.TempDir(), want: completion{incoming: incomingUser}}
		if current {
			f.stateDir = storeWithUser(t, currentUser)
			f.want.current = currentUser
		}
		f.child = childRestore{StateDir: f.stateDir, FromFile: src}
		return f
	}
	head := currentVersion(t, snapshotFile(t))
	fromSnapshot := func(t *testing.T) fixture {
		f := fromFile(t, true, "")
		f.child.FromFile = ""
		f.child.Remote, _ = snapshotRemote(t, snapshotFile(t))
		f.child.Snapshot = stampNew
		f.want.refs = []string{stampNew}
		return f
	}

	tests := []struct {
		name    string
		arrange func(t *testing.T) fixture
		crashAt string
		// afterCrash changes the leaf between the crash and the checks.
		afterCrash func(t *testing.T, f fixture)
		wantRecord string
		// wantOpen is what a write-mode open refuses naming, or "" when
		// it proceeds.
		wantOpen string
		// wantResume is what the resume refuses naming, or "" when it
		// finishes.
		wantResume []string
		wantOut    string
	}{
		{
			name:       "after step 1",
			arrange:    func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt:    stepOpened,
			wantResume: []string{"no restore-intent record", "there is no restore to resume"},
		},
		{
			name:       "after step 2",
			arrange:    func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt:    stepStaged,
			wantResume: []string{"no restore-intent record"},
		},
		{
			name:       "after step 3",
			arrange:    func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt:    stepQuiesced,
			wantResume: []string{"no restore-intent record"},
		},
		{
			name:       "after step 4",
			arrange:    func(t *testing.T) fixture { return fromFile(t, true, walFile(t)) },
			crashAt:    stepRecorded,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name:       "between step 5's two calls",
			arrange:    func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt:    stepSetAside,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name: "between step 5's two calls, discarding",
			arrange: func(t *testing.T) fixture {
				f := fromFile(t, false, snapshotFile(t))
				f.stateDir = storeWithUser(t, currentUser)
				f.child.StateDir, f.child.Discard = f.stateDir, true
				l := leafOf(f.stateDir)
				for _, suffix := range []string{"", "-wal", "-shm"} {
					if err := os.WriteFile(l.db+suffix, []byte("not a database "+suffix), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				f.want.discarded = []string{"", "-wal", "-shm"}
				return f
			},
			crashAt:    stepSetAside,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name:       "after step 5, before the reference",
			arrange:    fromSnapshot,
			crashAt:    stepSwapped,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name: "after the reference",
			arrange: func(t *testing.T) fixture {
				f := fromSnapshot(t)
				older := filepath.Join(leafOf(f.stateDir).snapshots, "backup-"+stampOld+".ref")
				if err := os.WriteFile(older, []byte("stamp "+stampOld+"\nremote "+f.child.Remote+"\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				return f
			},
			crashAt:    stepReferenced,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name:       "after step 4, no current store",
			arrange:    func(t *testing.T) fixture { return fromFile(t, false, snapshotFile(t)) },
			crashAt:    stepRecorded,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name:       "after step 5's first call, no current store",
			arrange:    func(t *testing.T) fixture { return fromFile(t, false, snapshotFile(t)) },
			crashAt:    stepSetAside,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name:       "after step 5, no current store",
			arrange:    func(t *testing.T) fixture { return fromFile(t, false, snapshotFile(t)) },
			crashAt:    stepSwapped,
			wantRecord: "swap",
			wantOpen:   "restore --resume",
		},
		{
			name:    "after step 4, then a binary below the staged store's version",
			arrange: func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt: stepRecorded,
			afterCrash: func(t *testing.T, f fixture) {
				stage := filepath.Join(leafOf(f.stateDir).staging, "restore-"+restoreStamp+".db")
				setVersion(t, stage, currentVersion(t, stage)+1)
			},
			wantRecord: "swap",
			wantOpen:   "restore --resume",
			wantResume: []string{fmt.Sprintf("copy at version %d is newer than this binary's version %d", head+1, head)},
		},
		{
			name:       "after step 5 without a library, current version",
			arrange:    func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt:    stepReconcileOwed,
			wantRecord: "reconcile",
			wantOpen:   "restore --resume",
		},
		{
			name: "after step 5 without a library, older store",
			arrange: func(t *testing.T) fixture {
				src := snapshotFile(t)
				setVersion(t, src, currentVersion(t, src)-1)
				return fromFile(t, true, src)
			},
			crashAt:    stepReconcileOwed,
			wantRecord: "reconcile-pending",
			wantOpen:   "start serve",
			wantResume: []string{"start serve to finish the restore"},
		},
		{
			name:    "reconcile, then a binary one version newer",
			arrange: func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt: stepReconcileOwed,
			afterCrash: func(t *testing.T, f fixture) {
				db := leafOf(f.stateDir).db
				setVersion(t, db, currentVersion(t, db)-1)
			},
			wantRecord: "reconcile",
			wantOpen:   "restore --resume",
			wantOut:    "start serve to finish the restore",
		},
		{
			name:    "after step 7's report, before its removal",
			arrange: func(t *testing.T) fixture { return fromFile(t, true, snapshotFile(t)) },
			crashAt: stepReported,
			afterCrash: func(t *testing.T, f fixture) {
				if err := os.WriteFile(leafOf(f.stateDir).report, []byte("the crashed attempt's report\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantRecord: "reconcile",
			wantOpen:   "restore --resume",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.arrange(t)
			f.child.CrashAt = tt.crashAt
			crashChild(t, f.child)
			if tt.afterCrash != nil {
				tt.afterCrash(t, f)
			}
			l := leafOf(f.stateDir)
			if got := recordState(t, f.stateDir); got != tt.wantRecord {
				t.Fatalf("record after the crash %q, want %q", got, tt.wantRecord)
			}
			if tt.crashAt == stepQuiesced {
				assertReadsWAL(t, f.stateDir)
			} else {
				assertWriteOpen(t, f.stateDir, tt.wantOpen)
			}
			if tt.crashAt == stepStaged {
				if got := entries(t, l.staging); !slices.Equal(got, []string{"restore-" + restoreStamp + ".db"}) {
					t.Errorf("staging/ after a crash after the stage %q, want the stage left", got)
				}
			}
			if tt.crashAt == stepRecorded {
				// Spec criterion 8: the refused open leaves the incoming
				// store where the record names it.
				if got := entries(t, l.staging); !slices.Equal(got, []string{"restore-" + restoreStamp + ".db"}) {
					t.Errorf("staging/ after a refused write-mode open %q, want the incoming store alone", got)
				}
			}

			bin, s := useFetchStandin(t, "", "")
			t.Setenv("PATH", filepath.Dir(bin))
			out, err := resume(t, f.stateDir)
			if tt.wantResume != nil {
				if err == nil {
					t.Fatalf("resume succeeded, want a refusal naming %q", tt.wantResume)
				}
				for _, sub := range tt.wantResume {
					if !strings.Contains(err.Error(), sub) {
						t.Errorf("resume: %v, want it to name %q", err, sub)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("resume: %v", err)
			}
			if lines := s.lines(t); len(lines) != 0 {
				t.Errorf("the resume ran rclone: %q", lines)
			}
			if tt.wantOut != "" {
				if !strings.Contains(out, tt.wantOut) {
					t.Errorf("resume printed %q, want it to name %q", out, tt.wantOut)
				}
				f.want.pending = true
			}
			assertComplete(t, f.stateDir, f.want)
		})
	}
}

// libraryRemote configures a real rclone local remote holding the
// snapshot file src as snapshots/<stampNew>.db and two library files,
// and returns the remote as a restore names it and the library files
// by relative path.
func libraryRemote(t *testing.T, src string) (string, []string) {
	t.Helper()
	remote, dir := snapshotRemote(t, src)
	files := []string{"artist/album/one.flac", "artist/album/two.flac"}
	writeTree(t, filepath.Join(dir, "library"), files...)
	return remote, files
}

// assertTree fails the test unless root holds exactly files, each
// holding its own name as writeTree writes it.
func assertTree(t *testing.T, root string, files []string) {
	t.Helper()
	var got []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if string(b) != rel {
			t.Errorf("%s holds %q, want %q", rel, b, rel)
		}
		got = append(got, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if !slices.Equal(got, files) {
		t.Errorf("library root holds %q, want %q", got, files)
	}
}

// TestRestore_LibraryCrash is the library row of spec criterion 7: a
// restore killed once the record owes the library copy leaves the
// record in state library, which every write-mode open refuses, and a
// resume runs the copy through a real rclone and finishes.
func TestRestore_LibraryCrash(t *testing.T) {
	remote, files := libraryRemote(t, snapshotFile(t))
	stateDir := storeWithUser(t, currentUser)
	root := t.TempDir()
	crashChild(t, childRestore{StateDir: stateDir, Snapshot: stampNew, Remote: remote, Library: true, LibraryRoot: root, CrashAt: stepLibraryOwed})
	if got := recordState(t, stateDir); got != "library" {
		t.Fatalf("record after the crash %q, want library", got)
	}
	assertTree(t, root, nil)
	assertWriteOpen(t, stateDir, "restore --resume")

	if _, err := resume(t, stateDir); err != nil {
		t.Fatalf("resume: %v", err)
	}
	assertTree(t, root, files)
	assertComplete(t, stateDir, completion{incoming: incomingUser, current: currentUser, refs: []string{stampNew}})
}

// TestRestore_Snapshot is a whole restore from a real rclone remote
// under umask 022, the store and the library: spec criterion 13 on the
// snapshot path (copyto renames the stage in at 0644 under that umask,
// measured, so the restored store reads 0600 only because the restore
// sets it), the reference rewritten to the restored stamp, and the
// warning a remote other than the reference's draws.
func TestRestore_Snapshot(t *testing.T) {
	withUmask(t, 0o022)
	remote, files := libraryRemote(t, snapshotFile(t))
	stateDir := storeWithUser(t, currentUser)
	older := filepath.Join(leafOf(stateDir).snapshots, "backup-"+stampOld+".ref")
	if err := os.WriteFile(older, []byte("stamp "+stampOld+"\nremote elsewhere:backups\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	var out bytes.Buffer
	err := Restore(ctxFor(t), RestoreConfig{
		StateDir:    stateDir,
		BusyTimeout: testBusy,
		Now:         restoreNow,
		Snapshot:    "latest",
		Remote:      remote,
		Library:     true,
		LibraryRoot: root,
		Out:         &out,
		Logger:      slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("Restore(): %v", err)
	}
	info, err := os.Stat(leafOf(stateDir).db)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("restored store mode %v, want 0600", got)
	}
	assertTree(t, root, files)
	assertComplete(t, stateDir, completion{incoming: incomingUser, current: currentUser, refs: []string{stampNew}})
	for _, want := range []string{
		"warning: restoring from " + remote + ", but the backup reference names elsewhere:backups",
		"the reconciliation report is " + leafOf(stateDir).report,
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("Restore printed %q, want it to name %q", out.String(), want)
		}
	}
}

// TestRestore_Older is a restore of a store one version below this
// binary: it swaps, leaves the record reconcile-pending for serve, and
// says so; the backup reference is left alone on the file path.
func TestRestore_Older(t *testing.T) {
	src := snapshotFile(t)
	setVersion(t, src, currentVersion(t, src)-1)
	stateDir := storeWithUser(t, currentUser)
	var out bytes.Buffer
	err := Restore(ctxFor(t), RestoreConfig{
		StateDir: stateDir, BusyTimeout: testBusy, Now: restoreNow, FromFile: src,
		Out: &out, Logger: slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("Restore(): %v", err)
	}
	if !strings.Contains(out.String(), "start serve to finish the restore") {
		t.Errorf("Restore printed %q, want it to name serve", out.String())
	}
	assertComplete(t, stateDir, completion{incoming: incomingUser, current: currentUser, pending: true})
	assertWriteOpen(t, stateDir, "start serve")
}

// TestRestore_WrongArchive: a truncated store is refused by its
// verification before any record, naming SQLite's answer, and its
// stage is removed, leaving the current store as it was.
func TestRestore_WrongArchive(t *testing.T) {
	b, err := os.ReadFile(snapshotFile(t))
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "truncated.db")
	if err := os.WriteFile(src, b[:len(b)/2/4096*4096], 0o600); err != nil {
		t.Fatal(err)
	}
	stateDir := storeWithUser(t, currentUser)
	err = Restore(ctxFor(t), RestoreConfig{
		StateDir: stateDir, BusyTimeout: testBusy, Now: restoreNow, FromFile: src,
		Out: &bytes.Buffer{}, Logger: slog.New(slog.DiscardHandler),
	})
	if err == nil || !strings.Contains(err.Error(), "database disk image is malformed") {
		t.Fatalf("Restore() = %v, want the integrity refusal", err)
	}
	if got := recordState(t, stateDir); got != "" {
		t.Errorf("record %q after a refused stage, want none", got)
	}
	if got := entries(t, leafOf(stateDir).staging); len(got) != 0 {
		t.Errorf("staging/ holds %q after a refused stage, want it empty", got)
	}
	if got := users(t, leafOf(stateDir).db); !slices.Equal(got, []string{currentUser}) {
		t.Errorf("current store holds %q, want %q", got, currentUser)
	}
}

// TestRestore_ConfigRefusals: each source and flag combination a
// restore cannot run, and a remote that does not parse, is refused
// before the store leaf is touched.
func TestRestore_ConfigRefusals(t *testing.T) {
	tests := []struct {
		name string
		cfg  RestoreConfig
		want string
	}{
		{name: "no source", cfg: RestoreConfig{}, want: "exactly one of a snapshot, a file, or a resume"},
		{name: "two sources", cfg: RestoreConfig{Snapshot: "latest", FromFile: "x.db"}, want: "exactly one"},
		{name: "older from a file", cfg: RestoreConfig{FromFile: "x.db", Older: true}, want: "--older applies only to a snapshot"},
		{name: "older with the library", cfg: RestoreConfig{Snapshot: stampOld, Older: true, Library: true}, want: "--older never combines with --library"},
		{name: "library from a file", cfg: RestoreConfig{FromFile: "x.db", Library: true}, want: "--library needs a snapshot"},
		{name: "into-nonempty alone", cfg: RestoreConfig{Snapshot: "latest", IntoNonempty: true}, want: "--into-nonempty applies only to --library"},
		{name: "remote with no colon", cfg: RestoreConfig{Snapshot: "latest", Remote: "backups"}, want: `backup remote "backups": no colon`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			tt.cfg.StateDir = stateDir
			tt.cfg.Out = &bytes.Buffer{}
			tt.cfg.Logger = slog.New(slog.DiscardHandler)
			err := Restore(context.Background(), tt.cfg)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Restore() = %v, want a refusal naming %q", err, tt.want)
			}
			if got := entries(t, stateDir); len(got) != 0 {
				t.Errorf("state directory holds %q after a refusal, want nothing", got)
			}
		})
	}
}

// awaitPartial waits for rclone's temporary download in staging to
// appear, and returns its name.
func awaitPartial(t *testing.T, staging string) string {
	t.Helper()
	for end := time.Now().Add(20 * time.Second); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
		des, _ := os.ReadDir(staging)
		for _, de := range des {
			if strings.HasSuffix(de.Name(), ".partial") {
				return de.Name()
			}
		}
	}
	t.Fatal("no .partial appeared in staging/ in 20s")
	return ""
}

// rclonePID returns the pid of the rclone process whose arguments name
// arg, read from /proc.
func rclonePID(t *testing.T, arg string) int {
	t.Helper()
	dirs, err := filepath.Glob("/proc/[0-9]*")
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range dirs {
		b, err := os.ReadFile(filepath.Join(d, "cmdline"))
		if err != nil {
			continue
		}
		argv := strings.Split(strings.TrimSuffix(string(b), "\x00"), "\x00")
		if filepath.Base(argv[0]) == "rclone" && slices.Contains(argv, arg) {
			pid, err := strconv.Atoi(filepath.Base(d))
			if err != nil {
				t.Fatal(err)
			}
			return pid
		}
	}
	t.Fatalf("no rclone process names %s", arg)
	return 0
}

// TestRestore_FetchInterrupted is the two mid-fetch rows, a real rclone
// held mid-transfer by a bandwidth limit: a SIGINT to the restore
// cancels the fetch, rclone removes its .partial (measured), and the
// stage goes with it; a SIGKILL to rclone itself leaves its .partial,
// which the next new restore clears. Neither leaves a record, so the
// current store opens in write mode and there is nothing to resume.
func TestRestore_FetchInterrupted(t *testing.T) {
	big := filepath.Join(t.TempDir(), "big.db")
	if err := os.WriteFile(big, bytes.Repeat([]byte("snapshot"), 256<<10), 0o600); err != nil {
		t.Fatal(err)
	}
	remote, _ := snapshotRemote(t, big)
	t.Setenv("RCLONE_BWLIMIT", "128k")
	stage := "restore-" + restoreStamp + ".db"

	t.Run("SIGINT to the restore", func(t *testing.T) {
		stateDir := storeWithUser(t, currentUser)
		l := leafOf(stateDir)
		cmd, stderr := startChild(t, childRestore{StateDir: stateDir, Snapshot: stampNew, Remote: remote})
		awaitPartial(t, l.staging)
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Fatal(err)
		}
		if err := cmd.Wait(); cmd.ProcessState.ExitCode() != 1 || !strings.Contains(stderr.String(), "context canceled") {
			t.Fatalf("child after SIGINT: %v, want exit 1 on the cancel\n%s", err, stderr)
		}
		if got := entries(t, l.staging); len(got) != 0 {
			t.Errorf("staging/ after the interrupted fetch holds %q, want it empty", got)
		}
		if got := recordState(t, stateDir); got != "" {
			t.Errorf("record %q after the interrupted fetch, want none", got)
		}
		assertWriteOpen(t, stateDir, "")
		if _, err := resume(t, stateDir); err == nil || !strings.Contains(err.Error(), "no restore-intent record") {
			t.Errorf("resume: %v, want the no-record refusal", err)
		}
	})

	t.Run("SIGKILL to rclone", func(t *testing.T) {
		stateDir := storeWithUser(t, currentUser)
		l := leafOf(stateDir)
		cmd, stderr := startChild(t, childRestore{StateDir: stateDir, Snapshot: stampNew, Remote: remote})
		partial := awaitPartial(t, l.staging)
		if err := syscall.Kill(rclonePID(t, filepath.Join(l.staging, stage)), syscall.SIGKILL); err != nil {
			t.Fatal(err)
		}
		if err := cmd.Wait(); cmd.ProcessState.ExitCode() != 1 || !strings.Contains(stderr.String(), "rclone-signal-KILL") {
			t.Fatalf("child after rclone's SIGKILL: %v, want exit 1 naming the signal\n%s", err, stderr)
		}
		if got := entries(t, l.staging); !slices.Equal(got, []string{partial}) {
			t.Errorf("staging/ after the killed fetch holds %q, want the leftover %s alone", got, partial)
		}
		if got := recordState(t, stateDir); got != "" {
			t.Errorf("record %q after the killed fetch, want none", got)
		}
		assertWriteOpen(t, stateDir, "")
		sess, err := store.OpenRestore(stateDir, store.RestoreOptions{BusyTimeout: testBusy, Now: restoreNow, Logger: slog.New(slog.DiscardHandler)})
		if err != nil {
			t.Fatalf("a new restore: %v", err)
		}
		if err := sess.Close(); err != nil {
			t.Fatal(err)
		}
		if got := entries(t, l.staging); len(got) != 0 {
			t.Errorf("staging/ after a new restore opened holds %q, want it cleared", got)
		}
	})
}
