package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// binaryHead is the embedded migration set's highest version, the
// version every mode opens a store at: 0001_ledger, 0002_jobs, and
// 0003_catalog.
const binaryHead = 3

// openBusy is the busy timeout every open-mode test passes.
const openBusy = 2 * time.Second

// newLeafDir returns a fresh state directory with a complete leaf and
// no database file.
func newLeafDir(t *testing.T) (string, leaf) {
	t.Helper()
	stateDir := t.TempDir()
	l := newLeaf(stateDir)
	if err := l.create(stateDir); err != nil {
		t.Fatalf("create: %v", err)
	}
	return stateDir, l
}

// buildStore makes a complete leaf under stateDir holding a store
// migrated through version through of the embedded set, with every
// connection closed.
func buildStore(t *testing.T, stateDir string, through int) leaf {
	t.Helper()
	l := newLeaf(stateDir)
	if err := l.create(stateDir); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := preCreate(l.dbPath); err != nil {
		t.Fatalf("preCreate: %v", err)
	}
	h, err := openHandle(l.dbPath, openBusy)
	if err != nil {
		t.Fatalf("openHandle: %v", err)
	}
	src := migrationSourceThrough(t, through)
	if err := migrate(context.Background(), h, l, readMigrationsT(t, src), defaultOpts()); err != nil {
		_ = h.Close()
		t.Fatalf("migrate through %d: %v", through, err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return l
}

// futureLedgerRow sets a store to one version above binaryHead and
// records that version in the ledger, as a newer binary would.
const futureLedgerRow = `PRAGMA user_version = 4;
INSERT INTO migrations (version, name, sha256) VALUES (4, '0004_future.sql', '0000000000000000000000000000000000000000000000000000000000000000');`

// hashFile returns the SHA-256 of path's bytes. A caller hashes a
// database file only with every connection to it in the test process
// closed, since reading it through os while a connection holds its
// locks would drop them.
func hashFile(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return sha256.Sum256(b)
}

// treeListing names every entry under root with its mode and, for a
// regular file, its size, so an open that creates, removes, or grows
// anything under the leaf changes the listing.
func treeListing(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		size := int64(0)
		if info.Mode().IsRegular() {
			size = info.Size()
		}
		out = append(out, fmt.Sprintf("%s %s %d", rel, info.Mode(), size))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

// errState returns the state a refusal reports as, failing the test
// when err is not one.
func errState(t *testing.T, err error) State {
	t.Helper()
	var se *stateError
	if !errors.As(err, &se) {
		t.Fatalf("error %v is not a refusal carrying a state", err)
	}
	return se.state
}

// holdersHandler records the want attribute of every holders record a
// serve-mode lock wait logs, and closes signal on the first whose want
// is signalWant.
type holdersHandler struct {
	signalWant string
	signal     chan struct{}
	once       sync.Once

	mu    sync.Mutex
	wants []string
}

func newHoldersHandler(signalWant string) *holdersHandler {
	return &holdersHandler{signalWant: signalWant, signal: make(chan struct{})}
}

func (h *holdersHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *holdersHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Message != holdersLogMsg {
		return nil
	}
	var want string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "want" {
			want = a.Value.String()
		}
		return true
	})
	h.mu.Lock()
	h.wants = append(h.wants, want)
	h.mu.Unlock()
	if want == h.signalWant {
		h.once.Do(func() { close(h.signal) })
	}
	return nil
}

func (h *holdersHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *holdersHandler) WithGroup(string) slog.Handler      { return h }

func (h *holdersHandler) recorded() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.wants)
}

// serveOpts returns the serve options every test uses, logging to
// logger, with a wait long enough that no row meets it by accident.
func serveOpts(logger *slog.Logger) ServeOptions {
	return ServeOptions{BusyTimeout: openBusy, MigrationWait: 30 * time.Second, Keep: 2, Logger: logger}
}

// quietServe opens stateDir in serve mode, discarding every log
// record, and closes the handle at cleanup.
func quietServe(t *testing.T, stateDir string) (*Handle, error) {
	t.Helper()
	h, err := OpenServe(context.Background(), stateDir, serveOpts(slog.New(slog.DiscardHandler)))
	if err == nil {
		t.Cleanup(func() { _ = h.Close() })
	}
	return h, err
}

// quietWrite opens stateDir in write mode and closes the handle at
// cleanup.
func quietWrite(t *testing.T, stateDir string) (*Handle, error) {
	t.Helper()
	h, err := OpenWrite(deadline(t, 5*time.Second), stateDir, openBusy)
	if err == nil {
		t.Cleanup(func() { _ = h.Close() })
	}
	return h, err
}

// TestOpenServe_BackupFieldControlsPruneOnMigrate is Task 1S's proof
// that settle passes ServeOptions.Backup through to the migration
// runner unchanged: with it off, the runner prunes pre-migration
// copies to keep after the migration it runs; with it on, it prunes
// nothing. Fixture: a store one version back from head, with two
// pre-migration copies already planted in snapshots/, keep 1.
func TestOpenServe_BackupFieldControlsPruneOnMigrate(t *testing.T) {
	tests := []struct {
		name   string
		backup bool
		want   int
	}{
		{name: "off prunes to keep", backup: false, want: 1},
		{name: "on keeps every copy", backup: true, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead-1)
			for _, name := range []string{"pre-migrate-v1.db", "pre-migrate-v99.db"} {
				if err := os.WriteFile(filepath.Join(l.snapshotsDir, name), []byte("planted"), 0o600); err != nil {
					t.Fatalf("plant %s: %v", name, err)
				}
			}

			opts := serveOpts(slog.New(slog.DiscardHandler))
			opts.Keep = 1
			opts.Backup = tt.backup
			h, err := OpenServe(context.Background(), stateDir, opts)
			if err != nil {
				t.Fatalf("OpenServe: %v", err)
			}
			t.Cleanup(func() { _ = h.Close() })

			if got := len(listDir(t, l.snapshotsDir)); got != tt.want {
				t.Errorf("snapshots = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestOpenWrite_RefusesAllButCurrent is spec criterion 13's write-mode
// half: every store write mode must not open is refused, by state and
// by the fix its text names.
func TestOpenWrite_RefusesAllButCurrent(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, stateDir string)
		wantState State
		wantText  []string
	}{
		{
			name:      "absent",
			setup:     func(*testing.T, string) {},
			wantState: StateAbsent,
			wantText:  []string{"start serve first"},
		},
		{
			name: "zero-byte file",
			setup: func(t *testing.T, stateDir string) {
				l := newLeaf(stateDir)
				if err := l.create(stateDir); err != nil {
					t.Fatal(err)
				}
				precreateFile(t, l.dbPath)
			},
			wantState: StateNotAStore,
			wantText:  []string{"start serve to finish creating it"},
		},
		{
			name: "text file",
			setup: func(t *testing.T, stateDir string) {
				l := newLeaf(stateDir)
				if err := l.create(stateDir); err != nil {
					t.Fatal(err)
				}
				text := strings.Repeat("this is not a database\n", 200)
				if err := os.WriteFile(l.dbPath, []byte(text), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantState: StateNotAStore,
			wantText:  []string{"is not a database"},
		},
		{
			name:      "older",
			setup:     func(t *testing.T, stateDir string) { buildStore(t, stateDir, 2) },
			wantState: StateOlder,
			wantText:  []string{"at version 2", "binary at version 3", "start serve to migrate"},
		},
		{
			name: "newer",
			setup: func(t *testing.T, stateDir string) {
				l := buildStore(t, stateDir, binaryHead)
				rawExec(t, l.dbPath, futureLedgerRow)
			},
			wantState: StateNewer,
			wantText:  []string{"at version 4", "binary's version 3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			tt.setup(t, stateDir)

			_, err := quietWrite(t, stateDir)
			if err == nil {
				t.Fatal("OpenWrite: got a handle, want a refusal")
			}
			if got := errState(t, err); got != tt.wantState {
				t.Errorf("refusal state = %v, want %v (%v)", got, tt.wantState, err)
			}
			assertErrNames(t, err, tt.wantText...)
		})
	}
}

// TestOpenWrite_AbsentCreatesNothing pins that write mode against a
// state directory with no leaf leaves the leaf absent.
func TestOpenWrite_AbsentCreatesNothing(t *testing.T) {
	stateDir := t.TempDir()
	if _, err := quietWrite(t, stateDir); err == nil {
		t.Fatal("OpenWrite: got a handle for an absent store")
	}
	if _, err := os.Stat(newLeaf(stateDir).root); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stat leaf after OpenWrite = %v, want not-exist", err)
	}
}

// TestOpenWrite_Current opens a current store, holding the lock shared
// and carrying the binary's version, and leaves a planted staging file
// alone: a live backup may be staging under its own shared lock.
func TestOpenWrite_Current(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	planted := filepath.Join(l.stagingDir, "backup-in-flight.db")
	if err := os.WriteFile(planted, []byte("staged"), 0o600); err != nil {
		t.Fatal(err)
	}

	h, err := quietWrite(t, stateDir)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	if h.version != binaryHead {
		t.Errorf("handle version = %d, want %d", h.version, binaryHead)
	}
	if h.lock == nil {
		t.Fatal("handle holds no lock")
	}
	if _, err := acquireExclusive(deadline(t, 100*time.Millisecond), l.lockPath); !errors.Is(err, ErrLockBusy) {
		t.Errorf("exclusive lock while the write handle is open = %v, want busy", err)
	}
	if got, err := os.ReadFile(planted); err != nil || string(got) != "staged" {
		t.Errorf("planted staging file after OpenWrite = %q, %v; want it untouched", got, err)
	}
}

// TestOpenWrite_LockHeldExclusively returns busy once ctx ends.
func TestOpenWrite_LockHeldExclusively(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	takeRawExclusive(t, l.lockPath)

	_, err := OpenWrite(deadline(t, 200*time.Millisecond), stateDir, openBusy)
	if !errors.Is(err, ErrLockBusy) {
		t.Errorf("OpenWrite under an exclusive holder = %v, want ErrLockBusy", err)
	}
}

// TestOpenModes_ForeignDatabaseUntouched: a foreign SQLite file at the
// store's path is not a store in every mode, and its bytes, hashed
// with every handle closed, are unchanged by the refusal. Both rows
// fail if classification opens a writing DSN, which sets WAL on open.
func TestOpenModes_ForeignDatabaseUntouched(t *testing.T) {
	tests := []struct {
		name  string
		stmts string
		want  string
	}{
		{
			name:  "own application_id",
			stmts: "PRAGMA application_id = 42; CREATE TABLE notes (body TEXT);",
			want:  "application_id 0x2a",
		},
		{
			name:  "rollback journal, user_version set, no application_id",
			stmts: "CREATE TABLE notes (body TEXT); PRAGMA user_version = 7;",
			want:  "never marked",
		},
	}
	modes := []struct {
		name string
		open func(t *testing.T, stateDir string) error
	}{
		{"inspect", func(t *testing.T, stateDir string) error {
			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			return refusal(r.State, "%s", r.Reason)
		}},
		{"write", func(t *testing.T, stateDir string) error { _, err := quietWrite(t, stateDir); return err }},
		{"serve", func(t *testing.T, stateDir string) error { _, err := quietServe(t, stateDir); return err }},
	}
	for _, tt := range tests {
		for _, m := range modes {
			t.Run(tt.name+"/"+m.name, func(t *testing.T) {
				stateDir, l := newLeafDir(t)
				precreateFile(t, l.dbPath)
				rawExec(t, l.dbPath, tt.stmts)
				before := hashFile(t, l.dbPath)

				err := m.open(t, stateDir)
				if err == nil {
					t.Fatal("got a handle on a foreign database")
				}
				if got := errState(t, err); got != StateNotAStore {
					t.Errorf("state = %v, want not a store (%v)", got, err)
				}
				assertErrNames(t, err, tt.want)
				if after := hashFile(t, l.dbPath); after != before {
					t.Error("the foreign database's bytes changed")
				}
			})
		}
	}
}

// TestOpenModes_LeafWithoutDatabase: a present leaf holding no database
// file is absent to inspect and "start serve first" to write mode, and
// neither changes anything in the leaf, the lock file included.
func TestOpenModes_LeafWithoutDatabase(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, l leaf, stateDir string)
	}{
		{
			name: "store.lock and db/",
			setup: func(t *testing.T, l leaf, stateDir string) {
				for _, dir := range []string{l.root, l.dbDir} {
					if err := os.Mkdir(dir, leafDirPerm); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(l.lockPath, nil, lockFilePerm); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "db/ and no store.lock",
			setup: func(t *testing.T, l leaf, stateDir string) {
				for _, dir := range []string{l.root, l.dbDir} {
					if err := os.Mkdir(dir, leafDirPerm); err != nil {
						t.Fatal(err)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := newLeaf(stateDir)
			tt.setup(t, l, stateDir)
			before := treeListing(t, l.root)

			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.State != StateAbsent {
				t.Errorf("inspect state = %v (%s), want absent", r.State, r.Reason)
			}
			_, err := quietWrite(t, stateDir)
			assertErrNames(t, err, "start serve first")

			if after := treeListing(t, l.root); !slices.Equal(after, before) {
				t.Errorf("leaf changed:\nbefore %q\nafter  %q", before, after)
			}
		})
	}
}

// headerOnlyDatabase makes the file a crash leaves between the
// pre-create's first open and migration 1's commit: pre-created, opened
// through the write DSN, a transaction begun and rolled back, and
// closed. The open writes SQLite's header page and nothing else.
func headerOnlyDatabase(t *testing.T, path string) {
	t.Helper()
	precreateFile(t, path)
	db, err := sql.Open(sqliteDriverName, writeDSN(path, openBusy))
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("roll back: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 4096 {
		t.Fatalf("header-only fixture is %d bytes, want 4096", info.Size())
	}
	if _, err := os.Stat(path + "-wal"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("header-only fixture has a -wal: %v", err)
	}
}

// TestOpenModes_InterruptedCreate: a zero-byte database with no -wal
// and a header-only one are each an unfinished create, which inspect
// and write mode report and serve completes to current.
func TestOpenModes_InterruptedCreate(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, path string)
	}{
		{name: "zero bytes", setup: precreateFile},
		{name: "header only", setup: headerOnlyDatabase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir, l := newLeafDir(t)
			tt.setup(t, l.dbPath)

			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.State != StateNotAStore || !strings.Contains(r.Reason, "start serve to finish creating it") {
				t.Errorf("inspect = %v (%s), want not a store naming the fix", r.State, r.Reason)
			}
			_, err := quietWrite(t, stateDir)
			assertErrNames(t, err, "start serve to finish creating it")

			h, err := quietServe(t, stateDir)
			if err != nil {
				t.Fatalf("OpenServe: %v", err)
			}
			if h.version != binaryHead {
				t.Errorf("handle version = %d, want %d", h.version, binaryHead)
			}
			if got := queryInt(t, l.dbPath, "PRAGMA user_version"); got != binaryHead {
				t.Errorf("user_version after serve = %d, want %d", got, binaryHead)
			}
		})
	}
}

// TestOpenModes_ZeroBytesBesideWAL: a zero-byte database with a -wal
// beside it is not a store in every mode, and neither file changes.
func TestOpenModes_ZeroBytesBesideWAL(t *testing.T) {
	stateDir, l := newLeafDir(t)
	precreateFile(t, l.dbPath)
	wal := l.dbPath + "-wal"
	if err := os.WriteFile(wal, []byte("frames a crash left"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := treeListing(t, l.dbDir)
	walHash := hashFile(t, wal)

	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateNotAStore {
		t.Errorf("inspect = %v (%s), want not a store", r.State, r.Reason)
	}
	_, werr := quietWrite(t, stateDir)
	_, serr := quietServe(t, stateDir)
	for mode, err := range map[string]error{"write": werr, "serve": serr} {
		if err == nil {
			t.Fatalf("%s: got a handle", mode)
		}
		if got := errState(t, err); got != StateNotAStore {
			t.Errorf("%s state = %v (%v), want not a store", mode, got, err)
		}
		assertErrNames(t, err, l.dbPath, wal)
	}
	if after := treeListing(t, l.dbDir); !slices.Equal(after, before) {
		t.Errorf("db/ changed:\nbefore %q\nafter  %q", before, after)
	}
	if hashFile(t, wal) != walHash {
		t.Error("the -wal's bytes changed")
	}
}

// TestOpenModes_UnreadableDatabase: a database file the owner cannot
// read is refused naming its path, never absent. SQLite answers the
// same code for it as for an absent file, so only os.Stat may decide
// absence.
func TestOpenModes_UnreadableDatabase(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Fatal("running as root, where no file mode denies a read: this row needs an unprivileged run")
	}
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	if err := os.Chmod(l.dbPath, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(l.dbPath, 0o600) })

	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateRefused || !strings.Contains(r.Reason, l.dbPath) {
		t.Errorf("inspect = %v (%s), want refused naming %s", r.State, r.Reason, l.dbPath)
	}
	_, werr := quietWrite(t, stateDir)
	_, serr := quietServe(t, stateDir)
	for mode, err := range map[string]error{"write": werr, "serve": serr} {
		if err == nil {
			t.Fatalf("%s: got a handle", mode)
		}
		if got := errState(t, err); got != StateRefused {
			t.Errorf("%s state = %v (%v), want refused", mode, got, err)
		}
	}
}

// TestOpenModes_OwnerAndFilesystemRefused: a leaf another uid owns, a
// multi-host filesystem, and a filesystem probe that fails are refused
// in every mode. The failed probe is reported as the I/O error it is,
// never as a multi-host refusal.
func TestOpenModes_OwnerAndFilesystemRefused(t *testing.T) {
	tests := []struct {
		name    string
		euid    int
		statfs  func(string, *syscall.Statfs_t) error
		want    string
		notWant string
	}{
		{
			name: "leaf owned by another uid",
			euid: os.Geteuid() + 1,
			want: "runuser -u",
		},
		{
			name:   "NFS",
			euid:   os.Geteuid(),
			statfs: func(_ string, st *syscall.Statfs_t) error { st.Type = nfsSuperMagic; return nil },
			want:   "spans hosts",
		},
		{
			name:    "statfs I/O error",
			euid:    os.Geteuid(),
			statfs:  func(string, *syscall.Statfs_t) error { return syscall.EIO },
			want:    "input/output error",
			notWant: "spans hosts",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			buildStore(t, stateDir, binaryHead)
			if tt.statfs != nil {
				orig := statfsFunc
				t.Cleanup(func() { statfsFunc = orig })
				statfsFunc = tt.statfs
			}

			withGeteuid(t, tt.euid)
			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			_, werr := OpenWrite(deadline(t, 5*time.Second), stateDir, openBusy)
			_, serr := OpenServe(context.Background(), stateDir, serveOpts(slog.New(slog.DiscardHandler)))
			reasons := map[string]string{"inspect": r.Reason}
			if r.State != StateRefused {
				t.Errorf("inspect state = %v (%s), want refused", r.State, r.Reason)
			}
			for mode, err := range map[string]error{"write": werr, "serve": serr} {
				if err == nil {
					t.Fatalf("%s: got a handle", mode)
				}
				if got := errState(t, err); got != StateRefused {
					t.Errorf("%s state = %v (%v), want refused", mode, got, err)
				}
				reasons[mode] = err.Error()
			}
			for mode, reason := range reasons {
				if !strings.Contains(reason, tt.want) {
					t.Errorf("%s reason %q does not name %q", mode, reason, tt.want)
				}
				if tt.notWant != "" && strings.Contains(reason, tt.notWant) {
					t.Errorf("%s reason %q names %q", mode, reason, tt.notWant)
				}
			}
		})
	}
}

// TestOpenModes_RestoreIntent: an intent record makes inspect report a
// restore in progress and write and serve refuse naming the fix, and
// serve leaves a staging file beside it alone, since the intent may
// name it.
func TestOpenModes_RestoreIntent(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	if err := os.WriteFile(l.restoreIntentPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(l.stagingDir, "pre-migrate-v1.db")
	if err := os.WriteFile(staged, []byte("restore's own"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
	if r.State != StateRestoreInProgress {
		t.Errorf("inspect state = %v (%s), want restore in progress", r.State, r.Reason)
	}
	_, werr := quietWrite(t, stateDir)
	assertErrNames(t, werr, "restore --resume")
	_, serr := quietServe(t, stateDir)
	assertErrNames(t, serr, "restore --resume")

	if got, err := os.ReadFile(staged); err != nil || string(got) != "restore's own" {
		t.Errorf("staging file beside the intent = %q, %v; want it untouched", got, err)
	}
}

// TestOpenServe_ClearsStagingLeftover: a staging file a crash left
// mid-snapshot, named as the next snapshot's stage, is removed before
// the migration, which then snapshots and completes.
func TestOpenServe_ClearsStagingLeftover(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, 1)
	leftover := filepath.Join(l.stagingDir, "pre-migrate-v1.db")
	if err := os.WriteFile(leftover, []byte("half a copy"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := quietServe(t, stateDir); err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if staged := listDir(t, l.stagingDir); len(staged) != 0 {
		t.Errorf("staging after serve = %q, want empty", staged)
	}
	if got := queryInt(t, l.dbPath, "PRAGMA user_version"); got != binaryHead {
		t.Errorf("user_version = %d, want %d", got, binaryHead)
	}
	if _, err := os.Stat(filepath.Join(l.snapshotsDir, "pre-migrate-v1.db")); err != nil {
		t.Errorf("the pre-migration copy is missing: %v", err)
	}
}

// TestOpenModes_LostDatabaseBesideCopy: a database that is absent or
// zero bytes beside a copy in snapshots/ may be a lost store rather
// than one never made, so serve refuses to create or complete it,
// naming the copy and the restore that brings it back, and inspect
// reports it refused.
func TestOpenModes_LostDatabaseBesideCopy(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, l leaf)
	}{
		{name: "no database file", setup: func(*testing.T, leaf) {}},
		{name: "zero-byte database", setup: func(t *testing.T, l leaf) { precreateFile(t, l.dbPath) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir, l := newLeafDir(t)
			tt.setup(t, l)
			cp := filepath.Join(l.snapshotsDir, "pre-migrate-v1.db")
			if err := os.WriteFile(cp, []byte("a copy"), 0o600); err != nil {
				t.Fatal(err)
			}
			before := treeListing(t, l.root)

			_, err := quietServe(t, stateDir)
			assertErrNames(t, err, cp, "restore --from-file")
			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.State != StateRefused || !strings.Contains(r.Reason, cp) {
				t.Errorf("inspect = %v (%s), want refused naming %s", r.State, r.Reason, cp)
			}
			if after := treeListing(t, l.root); !slices.Equal(after, before) {
				t.Errorf("leaf changed:\nbefore %q\nafter  %q", before, after)
			}
		})
	}
}

// TestOpenServe_LostDatabaseBesideBackupRef: once a backup has deleted
// every local copy, its reference is the leaf's one memory that a store
// existed, so serve refuses to create a store where the database file
// has gone, naming the reference and the snapshot to restore, and
// creates nothing.
func TestOpenServe_LostDatabaseBesideBackupRef(t *testing.T) {
	stateDir := currentStore(t)
	h := writeHandle(t, stateDir)
	if err := h.WriteBackupRef("20260924T120000.000000Z", "remote:path"); err != nil {
		t.Fatalf("WriteBackupRef: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	l := newLeaf(stateDir)
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(l.dbPath + suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
			t.Fatal(err)
		}
	}
	if got, want := listDir(t, l.snapshotsDir), []string{"backup-20260924T120000.000000Z.ref"}; !slices.Equal(got, want) {
		t.Fatalf("snapshots = %q, want only the reference %q", got, want)
	}
	before := treeListing(t, l.root)

	_, err := quietServe(t, stateDir)
	assertErrNames(t, err, filepath.Join(l.snapshotsDir, "backup-20260924T120000.000000Z.ref"), "restore --snapshot 20260924T120000.000000Z")
	if after := treeListing(t, l.root); !slices.Equal(after, before) {
		t.Errorf("leaf changed:\nbefore %q\nafter  %q", before, after)
	}
}

// TestOpenServe_CreatesAt0600 is spec criterion 16's third part: under
// a 0022 umask serve's create leaves a complete leaf and a current
// database at 0600, and a migration's snapshot is 0600 as well.
func TestOpenServe_CreatesAt0600(t *testing.T) {
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	t.Run("create", func(t *testing.T) {
		stateDir := t.TempDir()
		h, err := quietServe(t, stateDir)
		if err != nil {
			t.Fatalf("OpenServe: %v", err)
		}
		l := newLeaf(stateDir)
		assertLeafShape(t, l)
		if h.version != binaryHead {
			t.Errorf("handle version = %d, want %d", h.version, binaryHead)
		}
		for _, p := range []string{l.dbPath, l.dbPath + "-wal", l.dbPath + "-shm"} {
			info, err := os.Stat(p)
			if errors.Is(err, fs.ErrNotExist) && p != l.dbPath {
				continue
			}
			if err != nil {
				t.Fatalf("stat %s: %v", p, err)
			}
			if mode := info.Mode().Perm(); mode != 0o600 {
				t.Errorf("%s mode = %o, want 0600", p, mode)
			}
		}
	})
	t.Run("snapshot", func(t *testing.T) {
		stateDir := t.TempDir()
		l := buildStore(t, stateDir, 2)
		if _, err := quietServe(t, stateDir); err != nil {
			t.Fatalf("OpenServe: %v", err)
		}
		info, err := os.Stat(filepath.Join(l.snapshotsDir, "pre-migrate-v2.db"))
		if err != nil {
			t.Fatalf("stat the pre-migration copy: %v", err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Errorf("pre-migration copy mode = %o, want 0600", mode)
		}
	})
}

// TestOpenServe_LockHeldPastMigrationWait: serve waits up to the
// migration wait for each lock it needs, logs the holders while it
// waits, and then returns busy.
func TestOpenServe_LockHeldPastMigrationWait(t *testing.T) {
	tests := []struct {
		name string
		// hold takes the lock the row's other process holds.
		hold    int
		through int
		want    string
	}{
		{name: "exclusive held, serve needs shared", hold: syscall.LOCK_EX, through: binaryHead, want: "shared"},
		{name: "shared held, serve needs exclusive", hold: syscall.LOCK_SH, through: 2, want: "exclusive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, tt.through)
			f, err := os.OpenFile(l.lockPath, os.O_RDWR, 0)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = f.Close() })
			if err := syscall.Flock(int(f.Fd()), tt.hold|syscall.LOCK_NB); err != nil {
				t.Fatal(err)
			}

			logs := newHoldersHandler("")
			opts := serveOpts(slog.New(logs))
			opts.MigrationWait = 300 * time.Millisecond
			_, err = OpenServe(context.Background(), stateDir, opts)
			if !errors.Is(err, ErrLockBusy) {
				t.Fatalf("OpenServe under a holder = %v, want ErrLockBusy", err)
			}
			if got := logs.recorded(); len(got) == 0 || got[0] != tt.want {
				t.Errorf("holders records = %q, want the first wanting %s", got, tt.want)
			}
			if got := queryInt(t, l.dbPath, "PRAGMA user_version"); got != tt.through {
				t.Errorf("user_version = %d, want %d unchanged", got, tt.through)
			}
		})
	}
}

// serveResult is one OpenServe run's outcome.
type serveResult struct {
	h   *Handle
	err error
}

// startServe runs OpenServe in a goroutine, logging to logs, and
// returns its outcome's channel. The caller receives from it before the
// test ends.
func startServe(stateDir string, logs *holdersHandler) <-chan serveResult {
	done := make(chan serveResult, 1)
	go func() {
		h, err := OpenServe(context.Background(), stateDir, serveOpts(slog.New(logs)))
		done <- serveResult{h, err}
	}()
	return done
}

// holdShared takes a shared flock on path through a descriptor of its
// own, which flock(2) treats as another process.
func holdShared(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	return f
}

// awaitSignal waits for serve's first exclusive holders record, failing
// the test after a bound no healthy run approaches.
func awaitSignal(t *testing.T, logs *holdersHandler) bool {
	t.Helper()
	select {
	case <-logs.signal:
		return true
	case <-time.After(20 * time.Second):
		t.Error("serve logged no exclusive-lock holders record")
		return false
	}
}

// TestOpenServe_WaitsForExclusive is spec criterion 14's first row:
// serve reads an older store under the shared lock, releases it, and
// waits for the exclusive lock another holder blocks, changing nothing,
// then migrates once the holder releases. Every SQLite connection the
// test opens is closed before serve needs the file.
func TestOpenServe_WaitsForExclusive(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, 2)
	a := holdShared(t, l.lockPath)

	logs := newHoldersHandler("exclusive")
	done := startServe(stateDir, logs)
	if awaitSignal(t, logs) {
		if got := queryInt(t, l.dbPath, "PRAGMA user_version"); got != 2 {
			t.Errorf("user_version while serve waits = %d, want 2", got)
		}
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}

	res := <-done
	if res.err != nil {
		t.Fatalf("OpenServe: %v", res.err)
	}
	if err := res.h.Close(); err != nil {
		t.Fatal(err)
	}
	if got := queryInt(t, l.dbPath, "PRAGMA user_version"); got != binaryHead {
		t.Errorf("user_version after serve = %d, want %d", got, binaryHead)
	}
}

// TestOpenServe_ConversionLost is spec criterion 14's second row: while
// serve waits between its shared and exclusive holds it holds no lock
// and no connection, so another process can take the lock exclusively
// and replace the database. Serve then reads the replacement afresh
// and opens it unchanged rather than acting on what it read before.
//
// The replacement is a current store built elsewhere and closed. The
// older row also proves serve holds no descriptor on the file it had
// read, by os.SameFile over /proc/self/fd, since an fd link's target
// name follows a rename. The absent row starts from no database, where
// acting on the stale read would pre-create over the replacement.
func TestOpenServe_ConversionLost(t *testing.T) {
	tests := []struct {
		name    string
		through int // 0 leaves the leaf with no database file.
	}{
		{name: "older", through: 2},
		{name: "absent", through: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := newLeaf(stateDir)
			if tt.through > 0 {
				buildStore(t, stateDir, tt.through)
			} else if err := l.create(stateDir); err != nil {
				t.Fatal(err)
			}
			placed := buildStore(t, t.TempDir(), binaryHead).dbPath
			want := readState(t, placed)

			a := holdShared(t, l.lockPath)
			logs := newHoldersHandler("exclusive")
			done := startServe(stateDir, logs)
			if awaitSignal(t, logs) {
				replaceWhileHolding(t, a, l, placed, tt.through > 0)
			}
			if err := a.Close(); err != nil {
				t.Fatal(err)
			}

			res := <-done
			if res.err != nil {
				t.Fatalf("OpenServe: %v", res.err)
			}
			if err := res.h.Close(); err != nil {
				t.Fatal(err)
			}
			assertUnchanged(t, l.dbPath, want)
		})
	}
}

// replaceWhileHolding converts a's shared lock to exclusive, as a
// restore would, and then moves the database and its -wal and -shm
// aside, when there is one, and moves placed into its path. With
// checkFDs it asserts no descriptor in the process is on the moved
// file.
func replaceWhileHolding(t *testing.T, a *os.File, l leaf, placed string, checkFDs bool) {
	t.Helper()
	if err := syscall.Flock(int(a.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Errorf("convert the test's lock to exclusive while serve waits: %v", err)
		return
	}
	aside := l.dbPath + ".aside"
	if checkFDs {
		for _, suffix := range []string{"", "-wal", "-shm"} {
			if err := os.Rename(l.dbPath+suffix, aside+suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
				t.Fatal(err)
			}
		}
	}
	if err := os.Rename(placed, l.dbPath); err != nil {
		t.Fatal(err)
	}
	if !checkFDs {
		return
	}

	moved, err := os.Stat(aside)
	if err != nil {
		t.Fatal(err)
	}
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatalf("list /proc/self/fd: %v", err)
	}
	for _, fd := range fds {
		info, err := os.Stat(filepath.Join("/proc/self/fd", fd.Name()))
		if err != nil {
			continue
		}
		if os.SameFile(info, moved) {
			t.Errorf("descriptor %s is open on the moved database", fd.Name())
		}
	}
}

// TestOpenServe_CurrentKeepsSharedLock: serve on a current store opens
// it without the exclusive lock, which another shared holder would
// block, and holds the lock shared for the handle's life.
func TestOpenServe_CurrentKeepsSharedLock(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	a := holdShared(t, l.lockPath)
	t.Cleanup(func() { _ = a.Close() })

	h, err := quietServe(t, stateDir)
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if h.version != binaryHead {
		t.Errorf("handle version = %d, want %d", h.version, binaryHead)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := acquireExclusive(deadline(t, 100*time.Millisecond), l.lockPath); !errors.Is(err, ErrLockBusy) {
		t.Errorf("exclusive lock while the serve handle is open = %v, want busy", err)
	}
}
