package store

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

// swapChildEnv tells a re-executed test binary which connection
// TestSwapChild holds on the store.
const swapChildEnv = "DUBPLATE_STORE_SWAP_CHILD"

// The job ids these tests plant, as hex(id) reads them: the current
// store's, the incoming store's, and the three a killed child commits.
const (
	currentJob  = "0123456789ABCDEF0123456789ABCDEF"
	incomingJob = "FEDCBA9876543210FEDCBA9876543210"
)

var committedJobs = []string{
	"C0000000000000000000000000000001",
	"C0000000000000000000000000000002",
	"C0000000000000000000000000000003",
}

// testPreRestore is the pre-restore copy's name for a restore stamped
// testStamp.
const testPreRestore = PreRestorePrefix + testStamp + ".db"

// insertJobSQL is the statement every planted job row goes through.
const insertJobSQL = `INSERT INTO jobs (id, kind, at, state, owner_instance, actor, updated_at)
VALUES (X'%s', 'backup', 1, 'succeeded', 'inst', 'local:test', 1)`

// TestSwapChild is not a test on its own: it does nothing unless a
// parent test re-executes the binary naming it. As "commit" it commits
// three job rows through a write-mode handle, prints "ready", and
// sleeps until the parent kills it, so the rows live only in the -wal.
// As "reader" it writes a row, holds a read transaction open, and
// prints "ready"; as "idle" it writes a row and keeps its connection
// open with no transaction. Both wait for their stdin to close.
func TestSwapChild(t *testing.T) {
	action := os.Getenv(swapChildEnv)
	if action == "" {
		return
	}
	l := newLeaf(os.Getenv(childStateDirEnv))
	ctx := context.Background()
	switch action {
	case "commit":
		h, err := OpenWrite(ctx, filepath.Dir(l.root), openBusy)
		if err != nil {
			t.Fatalf("child open write: %v", err)
		}
		wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		err = h.Update(wctx, func(ctx context.Context, tx *Tx) error {
			for _, id := range committedJobs {
				if _, err := tx.Exec(ctx, fmt.Sprintf(insertJobSQL, id)); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("child commit: %v", err)
		}
		fmt.Println("ready")
		time.Sleep(time.Hour)
	case "reader", "idle":
		db, err := sql.Open(sqliteDriverName, writeDSN(l.dbPath, openBusy))
		if err != nil {
			t.Fatalf("child open: %v", err)
		}
		if _, err := db.Exec(fmt.Sprintf(insertJobSQL, committedJobs[0])); err != nil {
			t.Fatalf("child write: %v", err)
		}
		if action == "reader" {
			rdb, err := sql.Open(sqliteDriverName, readDSN(l.dbPath, openBusy))
			if err != nil {
				t.Fatalf("child open read: %v", err)
			}
			tx, err := rdb.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			if err != nil {
				t.Fatalf("child begin: %v", err)
			}
			var n int
			if err := tx.QueryRow("SELECT count(*) FROM jobs").Scan(&n); err != nil {
				t.Fatalf("child read: %v", err)
			}
		}
		fmt.Println("ready")
		_, _ = io.Copy(io.Discard, os.Stdin)
	default:
		t.Fatalf("child action %q", action)
	}
	os.Exit(0)
}

// startSwapChild re-executes the test binary as TestSwapChild in role
// action on stateDir and waits for it to print "ready". A "commit"
// child is then killed with SIGKILL and reaped; any other child is
// released at cleanup by closing its stdin.
func startSwapChild(t *testing.T, stateDir, action string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestSwapChild$", "-test.count=1")
	cmd.Env = append(os.Environ(), swapChildEnv+"="+action, childStateDirEnv+"="+stateDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if strings.TrimSpace(line) != "ready" {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("child %s not ready: %q %v\n%s", action, line, err, stderr.Bytes())
	}
	if action == "commit" {
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("kill child: %v", err)
		}
		_ = cmd.Wait()
		return
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		if err := cmd.Wait(); err != nil {
			t.Errorf("child %s: %v\n%s", action, err, stderr.Bytes())
		}
	})
}

// jobIDs reads every job id in the store at path, in order, through a
// read-only connection of its own.
func jobIDs(t *testing.T, path string) []string {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, inspectDSN(path, openBusy))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	return queryStrings(t, db, "SELECT hex(id) FROM jobs ORDER BY id")
}

// incomingBytes returns a closed store at the binary's version holding
// the one job incomingJob, after stmts run on it.
func incomingBytes(t *testing.T, stmts string) []byte {
	t.Helper()
	return storeBytes(t, binaryHead, fmt.Sprintf(insertJobSQL, incomingJob)+";"+stmts)
}

// stageFrom opens a new restore on stateDir from a file holding b, and
// stages and verifies it, failing the test on any refusal.
func stageFrom(t *testing.T, stateDir string, b []byte) *RestoreSession {
	t.Helper()
	s := openSessionFrom(t, stateDir, sourceFile(t, b))
	if err := s.StageFile(); err != nil {
		t.Fatalf("stage: %v", err)
	}
	if _, err := s.VerifyStage(deadline(t, 10*time.Second)); err != nil {
		t.Fatalf("verify: %v", err)
	}
	return s
}

// recordAndSwap runs the steps after the quiesce: the record, the swap's
// two halves, and returns the paths SwapOut moved files to.
func recordAndSwap(t *testing.T, s *RestoreSession, in RestoreIntent) []string {
	t.Helper()
	if err := s.Record(in); err != nil {
		t.Fatalf("record: %v", err)
	}
	moved, err := s.SwapOut(deadline(t, 10*time.Second))
	if err != nil {
		t.Fatalf("swap out: %v", err)
	}
	if err := s.SwapIn(); err != nil {
		t.Fatalf("swap in: %v", err)
	}
	return moved
}

// readRecord returns l's record, failing the test when it has none.
func readRecord(t *testing.T, l leaf) intentRecord {
	t.Helper()
	r, err := l.readIntent()
	if err != nil || r == nil {
		t.Fatalf("read record: %v, %v", r, err)
	}
	return *r
}

// header returns the database header's write and read versions: 1 and
// 1 for a rollback-journal database, 2 and 2 for WAL.
func header(t *testing.T, path string) [2]byte {
	t.Helper()
	b := readFile(t, path)
	return [2]byte{b[18], b[19]}
}

// TestRestore_CrashLeftWAL is spec criterion 6: a store whose last
// three commits live only in the -wal a killed writer left is quiesced
// into one file, so the swap moves that one file, leaves no -wal or
// -shm in db/, and the pre-restore copy holds the three rows. The
// fixture first proves the rows are in the -wal alone: a byte copy of
// the bare database file lacks them.
func TestRestore_CrashLeftWAL(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	startSwapChild(t, stateDir, "commit")

	info, err := os.Stat(l.dbPath + "-wal")
	if err != nil || info.Size() == 0 {
		t.Fatalf("the killed child left no -wal holding frames: %v", err)
	}
	bare := filepath.Join(t.TempDir(), "bare.db")
	if err := os.WriteFile(bare, readFile(t, l.dbPath), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := jobIDs(t, bare); len(got) != 0 {
		t.Fatalf("the bare database file holds jobs %q; the fixture needs them in the -wal alone", got)
	}

	s := stageFrom(t, stateDir, incomingBytes(t, ""))
	if err := s.Quiesce(deadline(t, 10*time.Second), false); err != nil {
		t.Fatalf("quiesce: %v", err)
	}
	if got := listDir(t, l.dbDir); !slices.Equal(got, []string{dbFileName}) {
		t.Errorf("db/ after the quiesce = %q, want the database alone", got)
	}
	if got := header(t, l.dbPath); got != [2]byte{1, 1} {
		t.Errorf("header versions after the quiesce = %v, want rollback-journal 1/1", got)
	}

	moved := recordAndSwap(t, s, RestoreIntent{File: "source.db"})
	copyPath := filepath.Join(l.snapshotsDir, testPreRestore)
	if !slices.Equal(moved, []string{copyPath}) {
		t.Errorf("moved = %q, want the one database file to %s", moved, copyPath)
	}
	if got := listDir(t, l.dbDir); !slices.Equal(got, []string{dbFileName}) {
		t.Errorf("db/ after the swap = %q, want the database alone", got)
	}
	if got := listDir(t, l.snapshotsDir); !slices.Equal(got, []string{testPreRestore}) {
		t.Errorf("snapshots/ = %q, want the pre-restore copy alone", got)
	}
	if got := listDir(t, l.stagingDir); len(got) != 0 {
		t.Errorf("staging/ = %q, want empty", got)
	}
	if got := jobIDs(t, copyPath); !slices.Equal(got, committedJobs) {
		t.Errorf("pre-restore copy jobs = %q, want the three committed %q", got, committedJobs)
	}
	if got := jobIDs(t, l.dbPath); !slices.Equal(got, []string{incomingJob}) {
		t.Errorf("restored store jobs = %q, want %q", got, incomingJob)
	}
	r := readRecord(t, l)
	if r.State != intentSwap || r.PreRestore != testPreRestore || r.DiscardCurrent || r.Incoming != testIncoming || r.Version != binaryHead || r.Source.File != "source.db" {
		t.Errorf("record = %+v, want state swap owing %s for %s at version %d", r, testPreRestore, testIncoming, binaryHead)
	}
}

// TestRestoreSession_QuiesceRefusals: each current store the quiesce
// cannot ready as one file is refused with its report, and no record
// is written. Another process holding a read transaction is named by
// the checkpoint's busy column; one holding an idle connection by the
// switch's SQLITE_BUSY; a sidecar the switch never touches by the net
// after the close. A file that is no store, or a sidecar beside no
// database, is refused naming --discard-current (criterion 9's
// current-store half).
func TestRestoreSession_QuiesceRefusals(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(t *testing.T, stateDir string, l leaf)
		want    func(l leaf) []string
	}{
		{
			name:    "a second process holds a read transaction",
			arrange: func(t *testing.T, stateDir string, _ leaf) { startSwapChild(t, stateDir, "reader") },
			want: func(l leaf) []string {
				return []string{l.dbPath, "a reader kept the checkpoint from finishing", inUseHint}
			},
		},
		{
			name:    "a second process holds an idle connection",
			arrange: func(t *testing.T, stateDir string, _ leaf) { startSwapChild(t, stateDir, "idle") },
			want:    func(l leaf) []string { return []string{l.dbPath, "leave WAL mode", "database is locked", inUseHint} },
		},
		{
			name: "already in DELETE mode, with a planted -shm",
			arrange: func(t *testing.T, _ string, l leaf) {
				rawExec(t, l.dbPath, "PRAGMA journal_mode=DELETE")
				if got := header(t, l.dbPath); got != [2]byte{1, 1} {
					t.Fatalf("fixture header = %v, want 1/1", got)
				}
				if err := os.WriteFile(l.dbPath+"-shm", []byte("planted"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: func(l leaf) []string { return []string{"found " + l.dbPath + "-shm after the close", inUseHint} },
		},
		{
			name: "cannot open",
			arrange: func(t *testing.T, _ string, l leaf) {
				if err := os.WriteFile(l.dbPath, bytes.Repeat([]byte("not sqlite "), 1000), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: func(l leaf) []string {
				return []string{l.dbPath + " cannot be opened", "not a database", "restore --discard-current"}
			},
		},
		{
			name: "a SQLite file that is no store",
			arrange: func(t *testing.T, _ string, l leaf) {
				if err := os.WriteFile(l.dbPath, sqliteBytes(t, "PRAGMA application_id = 4660; CREATE TABLE t (a INTEGER);"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: func(l leaf) []string {
				return []string{l.dbPath + " is not a dubplate store (application_id 0x1234)", "restore --discard-current"}
			},
		},
		{
			name: "a -wal beside no database",
			arrange: func(t *testing.T, _ string, l leaf) {
				if err := os.Remove(l.dbPath); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(l.dbPath+"-wal", []byte("stale"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: func(l leaf) []string {
				return []string{"found " + l.dbPath + "-wal beside no database", "restore --discard-current"}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead)
			plantJob(t, l.dbPath)
			tt.arrange(t, stateDir, l)
			s := stageFrom(t, stateDir, incomingBytes(t, ""))

			err := s.Quiesce(deadline(t, 20*time.Second), false)
			assertErrNames(t, err, tt.want(l)...)
			if err := s.Record(RestoreIntent{File: "source.db"}); err == nil {
				t.Error("record after a refused quiesce succeeded")
			}
			assertNoIntent(t, l)
		})
	}
}

// TestRestoreSession_NoCurrentStore: a leaf with nothing in db/ is not
// quiesced, its record owes no pre-restore copy and no discard, and the
// swap renames the incoming store in, moving nothing aside.
func TestRestoreSession_NoCurrentStore(t *testing.T) {
	stateDir, l := newLeafDir(t)
	s := stageFrom(t, stateDir, incomingBytes(t, ""))
	if err := s.Quiesce(deadline(t, 10*time.Second), false); err != nil {
		t.Fatalf("quiesce: %v", err)
	}
	moved := recordAndSwap(t, s, RestoreIntent{Remote: "remote:dubplate", Snapshot: "20260924T030000.000000Z", LibraryRoot: "/srv/music"})
	if len(moved) != 0 {
		t.Errorf("moved = %q, want nothing", moved)
	}
	r := readRecord(t, l)
	if r.PreRestore != "" || r.DiscardCurrent {
		t.Errorf("record owes pre_restore %q, discard %v; want neither", r.PreRestore, r.DiscardCurrent)
	}
	if r.Library == nil || r.Library.Root != "/srv/music" || r.Source.Snapshot != "20260924T030000.000000Z" {
		t.Errorf("record library %+v, source %+v; want root /srv/music and the snapshot", r.Library, r.Source)
	}
	if got := listDir(t, l.snapshotsDir); len(got) != 0 {
		t.Errorf("snapshots/ = %q, want empty", got)
	}
	if got := listDir(t, l.stagingDir); len(got) != 0 {
		t.Errorf("staging/ = %q, want empty", got)
	}
	if got := jobIDs(t, l.dbPath); !slices.Equal(got, []string{incomingJob}) {
		t.Errorf("restored store jobs = %q, want %q", got, incomingJob)
	}
}

// discardFixture plants in l's db/ the current files named by suffix,
// "" for an unopenable database, "-wal", and "-shm", each with its own
// bytes, and returns the bytes by suffix.
func discardFixture(t *testing.T, l leaf, suffixes ...string) map[string][]byte {
	t.Helper()
	contents := map[string][]byte{"": []byte("unopenable database"), "-wal": []byte("its wal"), "-shm": []byte("its shm")}
	planted := make(map[string][]byte, len(suffixes))
	for _, suffix := range suffixes {
		planted[suffix] = contents[suffix]
		if err := os.WriteFile(l.dbPath+suffix, planted[suffix], 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return planted
}

// allCurrent is every current file a discard moves.
var allCurrent = []string{"", "-wal", "-shm"}

// TestRestoreSession_Discard: with --discard-current, the quiesce opens
// and moves nothing, so a crash before the record leaves every current
// file in db/; after the record, the swap's first half moves the
// database, its -wal, and its -shm to their discarded names, naming
// each, and the second renames the incoming store in.
func TestRestoreSession_Discard(t *testing.T) {
	stateDir, l := newLeafDir(t)
	planted := discardFixture(t, l, allCurrent...)
	s := stageFrom(t, stateDir, incomingBytes(t, ""))

	if err := s.Quiesce(deadline(t, 10*time.Second), true); err != nil {
		t.Fatalf("quiesce: %v", err)
	}
	if got, want := listDir(t, l.dbDir), []string{dbFileName, dbFileName + "-shm", dbFileName + "-wal"}; !slices.Equal(got, want) {
		t.Fatalf("db/ after the quiesce, before the record = %q, want every current file %q", got, want)
	}
	if got := listDir(t, l.snapshotsDir); len(got) != 0 {
		t.Fatalf("snapshots/ before the record = %q, want empty", got)
	}

	moved := recordAndSwap(t, s, RestoreIntent{File: "source.db"})
	discarded := l.discardedPath(testStamp)
	if want := []string{discarded, discarded + "-wal", discarded + "-shm"}; !slices.Equal(moved, want) {
		t.Errorf("moved = %q, want %q", moved, want)
	}
	r := readRecord(t, l)
	if !r.DiscardCurrent || r.PreRestore != "" {
		t.Errorf("record owes discard %v, pre_restore %q; want the discard alone", r.DiscardCurrent, r.PreRestore)
	}
	for suffix, b := range planted {
		if got := readFile(t, discarded+suffix); !bytes.Equal(got, b) {
			t.Errorf("%s holds %q, want %q", discarded+suffix, got, b)
		}
	}
	if got := listDir(t, l.dbDir); !slices.Equal(got, []string{dbFileName}) {
		t.Errorf("db/ = %q, want the restored database alone", got)
	}
	if got := listDir(t, l.stagingDir); len(got) != 0 {
		t.Errorf("staging/ = %q, want empty", got)
	}
	if got := jobIDs(t, l.dbPath); !slices.Equal(got, []string{incomingJob}) {
		t.Errorf("restored store jobs = %q, want %q", got, incomingJob)
	}
}

// TestRestoreSession_DiscardSidecarsAlone: with --discard-current and
// a db/ holding a -wal and a -shm beside no database, the record still
// owes a discard, the swap's first half moves the two sidecars to their
// discarded names and no database, and the second renames the incoming
// store in.
func TestRestoreSession_DiscardSidecarsAlone(t *testing.T) {
	stateDir, l := newLeafDir(t)
	planted := discardFixture(t, l, "-wal", "-shm")
	s := stageFrom(t, stateDir, incomingBytes(t, ""))

	if err := s.Quiesce(deadline(t, 10*time.Second), true); err != nil {
		t.Fatalf("quiesce: %v", err)
	}
	moved := recordAndSwap(t, s, RestoreIntent{File: "source.db"})
	discarded := l.discardedPath(testStamp)
	if want := []string{discarded + "-wal", discarded + "-shm"}; !slices.Equal(moved, want) {
		t.Errorf("moved = %q, want %q", moved, want)
	}
	r := readRecord(t, l)
	if !r.DiscardCurrent || r.PreRestore != "" {
		t.Errorf("record owes discard %v, pre_restore %q; want the discard alone", r.DiscardCurrent, r.PreRestore)
	}
	for suffix, b := range planted {
		if got := readFile(t, discarded+suffix); !bytes.Equal(got, b) {
			t.Errorf("%s holds %q, want %q", discarded+suffix, got, b)
		}
	}
	base := filepath.Base(discarded)
	if got, want := listDir(t, l.snapshotsDir), []string{base + "-shm", base + "-wal"}; !slices.Equal(got, want) {
		t.Errorf("snapshots/ = %q, want the discarded sidecars alone %q", got, want)
	}
	if got := listDir(t, l.dbDir); !slices.Equal(got, []string{dbFileName}) {
		t.Errorf("db/ = %q, want the restored database alone", got)
	}
	if got := listDir(t, l.stagingDir); len(got) != 0 {
		t.Errorf("staging/ = %q, want empty", got)
	}
	if got := jobIDs(t, l.dbPath); !slices.Equal(got, []string{incomingJob}) {
		t.Errorf("restored store jobs = %q, want %q", got, incomingJob)
	}
}

// TestRestore_RollBackNewerCurrent is spec criterion 10: a current
// store one version above the binary, restored from a pre-migration
// copy at the binary's version, is quiesced and swapped, and the
// result opens in write mode under this binary.
func TestRestore_RollBackNewerCurrent(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	rawExec(t, l.dbPath, futureLedgerRow)
	plantJob(t, l.dbPath)
	if v := queryInt(t, l.dbPath, "PRAGMA user_version"); v != binaryHead+1 {
		t.Fatalf("fixture current store at version %d, want %d", v, binaryHead+1)
	}
	src := filepath.Join(l.snapshotsDir, preMigrateName(binaryHead))
	if err := os.WriteFile(src, incomingBytes(t, ""), 0o600); err != nil {
		t.Fatal(err)
	}

	s := openSessionFrom(t, stateDir, src)
	if err := s.StageFile(); err != nil {
		t.Fatalf("stage: %v", err)
	}
	if _, err := s.VerifyStage(deadline(t, 10*time.Second)); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := s.Quiesce(deadline(t, 10*time.Second), false); err != nil {
		t.Fatalf("quiesce a newer store: %v", err)
	}
	recordAndSwap(t, s, RestoreIntent{File: src})
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if v := queryInt(t, filepath.Join(l.snapshotsDir, testPreRestore), "PRAGMA user_version"); v != binaryHead+1 {
		t.Errorf("pre-restore copy at version %d, want the rolled-back %d", v, binaryHead+1)
	}
	if err := l.removeIntent(); err != nil {
		t.Fatal(err)
	}

	h, err := OpenWrite(deadline(t, 10*time.Second), stateDir, openBusy)
	if err != nil {
		t.Fatalf("write-mode open of the restored store: %v", err)
	}
	var id string
	err = h.QueryRow(context.Background(), "SELECT hex(id) FROM jobs").Scan(&id)
	if cerr := h.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil || id != incomingJob {
		t.Errorf("restored job = %q, %v; want %s", id, err, incomingJob)
	}
}

// TestRestoreSession_SwappedInFile: the file the swap renames into db/
// is the leaf owner's at 0600, and a rollback-journal file there reads
// as WAL once a write-mode open has opened it.
func TestRestoreSession_SwappedInFile(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	b := incomingBytes(t, "PRAGMA journal_mode=DELETE")
	if b[18] != 1 || b[19] != 1 {
		t.Fatalf("fixture header versions = %d/%d, want rollback-journal 1/1", b[18], b[19])
	}
	s := stageFrom(t, stateDir, b)
	if err := s.Quiesce(deadline(t, 10*time.Second), false); err != nil {
		t.Fatalf("quiesce: %v", err)
	}
	recordAndSwap(t, s, RestoreIntent{File: "source.db"})
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(l.dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("restored store mode = %04o, want 0600", mode)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("no unix stat info for %s", l.dbPath)
	}
	if int(st.Uid) != os.Geteuid() {
		t.Errorf("restored store owner = %d, want the leaf owner %d", st.Uid, os.Geteuid())
	}
	if got := header(t, l.dbPath); got != [2]byte{1, 1} {
		t.Fatalf("restored store header = %v before any open, want 1/1", got)
	}
	if err := l.removeIntent(); err != nil {
		t.Fatal(err)
	}
	h, err := OpenWrite(deadline(t, 10*time.Second), stateDir, openBusy)
	if err != nil {
		t.Fatalf("write-mode open: %v", err)
	}
	var mode string
	err = h.QueryRow(context.Background(), "PRAGMA journal_mode").Scan(&mode)
	if cerr := h.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil || mode != "wal" {
		t.Errorf("journal_mode after a write-mode open = %q, %v; want wal", mode, err)
	}
	if got := header(t, l.dbPath); got != [2]byte{2, 2} {
		t.Errorf("header after a write-mode open = %v, want WAL's 2/2", got)
	}
}

// swapKind is which of the three records a combination fixture plants.
type swapKind int

const (
	owesCopy swapKind = iota + 1
	owesNothing
	owesDiscard
	// owesWALDiscard is a discard record for a db/ that held a -wal
	// beside no database.
	owesWALDiscard
)

// mv renames from to to, failing the test on an error.
func mv(t *testing.T, from, to string) {
	t.Helper()
	if err := os.Rename(from, to); err != nil {
		t.Fatal(err)
	}
}

// TestRestoreSession_ResumeCombinations is the swap's combination table,
// one hand-built fixture per row: a record in state swap and the files
// a crash inside the swap can leave. Each row the table finishes ends
// with the incoming store in db/, staging/ empty, and the current store
// set aside as the record owes; each row it refuses changes nothing in
// the leaf. The fixture's base is the moment of the record: a current
// store holding currentJob in db/ (or the discard's three unopenable
// files, or nothing), and the incoming store holding incomingJob in
// staging/.
func TestRestoreSession_ResumeCombinations(t *testing.T) {
	tests := []struct {
		name    string
		kind    swapKind
		arrange func(t *testing.T, l leaf)
		// moved names the files SwapOut moves, as suffixes of the
		// discarded path, or "copy" for the pre-restore copy.
		moved   []string
		wantErr func(l leaf) []string
	}{
		{name: "copy owed, nothing moved", kind: owesCopy, moved: []string{"copy"}},
		{
			name: "copy owed, current written after the record",
			kind: owesCopy,
			arrange: func(t *testing.T, l leaf) {
				if err := os.WriteFile(l.dbPath+"-wal", []byte("written"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: func(l leaf) []string {
				return []string{"found " + l.dbPath + "-wal beside the current store", "a process wrote to it after the record", manualRecovery}
			},
		},
		{
			name:    "copy owed, current set aside",
			kind:    owesCopy,
			arrange: func(t *testing.T, l leaf) { mv(t, l.dbPath, filepath.Join(l.snapshotsDir, testPreRestore)) },
		},
		{
			name: "copy owed, swap finished",
			kind: owesCopy,
			arrange: func(t *testing.T, l leaf) {
				mv(t, l.dbPath, filepath.Join(l.snapshotsDir, testPreRestore))
				mv(t, filepath.Join(l.stagingDir, testIncoming), l.dbPath)
			},
		},
		{
			name:    "copy owed, current and copy both absent",
			kind:    owesCopy,
			arrange: func(t *testing.T, l leaf) { mv(t, l.dbPath, filepath.Join(t.TempDir(), "elsewhere.db")) },
			wantErr: func(l leaf) []string {
				return []string{
					"names files in no state a swap leaves",
					filepath.Join(l.stagingDir, testIncoming) + " present",
					l.dbPath + " absent",
					filepath.Join(l.snapshotsDir, testPreRestore) + " absent",
					manualRecovery,
				}
			},
		},
		{name: "no copy, nothing moved", kind: owesNothing},
		{
			name:    "no copy, swap finished",
			kind:    owesNothing,
			arrange: func(t *testing.T, l leaf) { mv(t, filepath.Join(l.stagingDir, testIncoming), l.dbPath) },
		},
		{
			name: "no copy, a current store appeared",
			kind: owesNothing,
			arrange: func(t *testing.T, l leaf) {
				if err := os.WriteFile(l.dbPath, storeBytes(t, binaryHead, ""), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: func(l leaf) []string {
				return []string{"names files in no state a swap leaves", l.dbPath + " present", manualRecovery}
			},
		},
		{
			name: "no copy, a stale -wal in db/",
			kind: owesNothing,
			arrange: func(t *testing.T, l leaf) {
				if err := os.WriteFile(l.dbPath+"-wal", []byte("stale"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: func(l leaf) []string {
				return []string{"the incoming store would land beside " + l.dbPath + "-wal", manualRecovery}
			},
		},
		{name: "discard owed, nothing moved", kind: owesDiscard, moved: []string{"", "-wal", "-shm"}},
		{
			name:    "discard owed, database moved",
			kind:    owesDiscard,
			arrange: func(t *testing.T, l leaf) { mv(t, l.dbPath, l.discardedPath(testStamp)) },
			moved:   []string{"-wal", "-shm"},
		},
		{
			name: "discard owed, every file moved",
			kind: owesDiscard,
			arrange: func(t *testing.T, l leaf) {
				for _, suffix := range []string{"", "-wal", "-shm"} {
					mv(t, l.dbPath+suffix, l.discardedPath(testStamp)+suffix)
				}
			},
		},
		{
			name: "discard owed, swap finished",
			kind: owesDiscard,
			arrange: func(t *testing.T, l leaf) {
				for _, suffix := range []string{"", "-wal", "-shm"} {
					mv(t, l.dbPath+suffix, l.discardedPath(testStamp)+suffix)
				}
				mv(t, filepath.Join(l.stagingDir, testIncoming), l.dbPath)
			},
		},
		{name: "discard of a -wal alone owed, nothing moved", kind: owesWALDiscard, moved: []string{"-wal"}},
		{
			name:    "discard of a -wal alone owed, -wal moved",
			kind:    owesWALDiscard,
			arrange: func(t *testing.T, l leaf) { mv(t, l.dbPath+"-wal", l.discardedPath(testStamp)+"-wal") },
		},
		{
			name: "discard of a -wal alone owed, swap finished",
			kind: owesWALDiscard,
			arrange: func(t *testing.T, l leaf) {
				mv(t, l.dbPath+"-wal", l.discardedPath(testStamp)+"-wal")
				mv(t, filepath.Join(l.stagingDir, testIncoming), l.dbPath)
			},
		},
		{
			name: "discard of a -wal alone owed, incoming lost",
			kind: owesWALDiscard,
			arrange: func(t *testing.T, l leaf) {
				mv(t, l.dbPath+"-wal", l.discardedPath(testStamp)+"-wal")
				mv(t, filepath.Join(l.stagingDir, testIncoming), filepath.Join(t.TempDir(), "elsewhere.db"))
			},
			wantErr: func(l leaf) []string {
				return []string{"names files in no state a swap leaves", l.dbPath + " absent", l.discardedPath(testStamp) + "-wal present", manualRecovery}
			},
		},
		{
			name: "discard owed, incoming lost before any move",
			kind: owesDiscard,
			arrange: func(t *testing.T, l leaf) {
				mv(t, filepath.Join(l.stagingDir, testIncoming), filepath.Join(t.TempDir(), "elsewhere.db"))
			},
			wantErr: func(l leaf) []string {
				return []string{"names files in no state a swap leaves", l.dbPath + " present", l.discardedPath(testStamp) + " absent", manualRecovery}
			},
		},
		{
			name: "discard owed, database in both places",
			kind: owesDiscard,
			arrange: func(t *testing.T, l leaf) {
				if err := os.WriteFile(l.discardedPath(testStamp), []byte("another"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: func(l leaf) []string {
				return []string{"names files in no state a swap leaves", l.dbPath + " present", l.discardedPath(testStamp) + " present", manualRecovery}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir, l := newLeafDir(t)
			r := testIntent(intentSwap)
			r.Version = binaryHead
			var planted map[string][]byte
			switch tt.kind {
			case owesCopy:
				if err := os.WriteFile(l.dbPath, storeBytes(t, binaryHead, fmt.Sprintf(insertJobSQL, currentJob)+"; PRAGMA journal_mode=DELETE"), 0o600); err != nil {
					t.Fatal(err)
				}
			case owesNothing:
				r.PreRestore = ""
			case owesDiscard:
				r.PreRestore, r.DiscardCurrent = "", true
				planted = discardFixture(t, l, allCurrent...)
			case owesWALDiscard:
				r.PreRestore, r.DiscardCurrent = "", true
				planted = discardFixture(t, l, "-wal")
			}
			if err := os.WriteFile(filepath.Join(l.stagingDir, testIncoming), incomingBytes(t, ""), 0o600); err != nil {
				t.Fatal(err)
			}
			plantIntent(t, l, r)
			if tt.arrange != nil {
				tt.arrange(t, l)
			}
			before := treeListing(t, l.root)

			s := openSession(t, stateDir, true)
			moved, err := s.SwapOut(deadline(t, 10*time.Second))
			if err == nil {
				err = s.SwapIn()
			}
			if tt.wantErr != nil {
				assertErrNames(t, err, tt.wantErr(l)...)
				if after := treeListing(t, l.root); !slices.Equal(after, before) {
					t.Errorf("a refused resume changed the leaf:\nbefore %q\nafter  %q", before, after)
				}
				return
			}
			if err != nil {
				t.Fatalf("resume the swap: %v", err)
			}

			var want []string
			for _, m := range tt.moved {
				if m == "copy" {
					want = append(want, filepath.Join(l.snapshotsDir, testPreRestore))
					continue
				}
				want = append(want, l.discardedPath(testStamp)+m)
			}
			if !slices.Equal(moved, want) {
				t.Errorf("moved = %q, want %q", moved, want)
			}
			if got := listDir(t, l.stagingDir); len(got) != 0 {
				t.Errorf("staging/ = %q, want empty", got)
			}
			switch tt.kind {
			case owesCopy:
				copyPath := filepath.Join(l.snapshotsDir, testPreRestore)
				if got := listDir(t, l.snapshotsDir); !slices.Equal(got, []string{testPreRestore}) {
					t.Errorf("snapshots/ = %q, want the pre-restore copy alone", got)
				}
				if got := jobIDs(t, copyPath); !slices.Equal(got, []string{currentJob}) {
					t.Errorf("pre-restore copy jobs = %q, want %q", got, currentJob)
				}
			case owesNothing:
				if got := listDir(t, l.snapshotsDir); len(got) != 0 {
					t.Errorf("snapshots/ = %q, want empty", got)
				}
			case owesDiscard, owesWALDiscard:
				var names []string
				for suffix := range planted {
					names = append(names, filepath.Base(l.discardedPath(testStamp))+suffix)
				}
				slices.Sort(names)
				if got := listDir(t, l.snapshotsDir); !slices.Equal(got, names) {
					t.Errorf("snapshots/ = %q, want the discarded files alone %q", got, names)
				}
				for suffix, b := range planted {
					if got := readFile(t, l.discardedPath(testStamp)+suffix); !bytes.Equal(got, b) {
						t.Errorf("discarded%s holds %q, want %q", suffix, got, b)
					}
					if got, err := os.ReadFile(l.dbPath + suffix); err == nil && bytes.Equal(got, b) {
						t.Errorf("the planted %s is still in db/", l.dbPath+suffix)
					}
				}
			}
			if got := jobIDs(t, l.dbPath); !slices.Equal(got, []string{incomingJob}) {
				t.Errorf("restored store jobs = %q, want %q", got, incomingJob)
			}
		})
	}
}

// TestRestoreSession_ResumeNewerIncoming: a swap resumed under a binary
// older than the incoming store is refused naming both versions before
// any rename, whether the incoming store still waits in staging/ or the
// swap already renamed it into db/. A staged store the record names is
// kept, byte for byte.
func TestRestoreSession_ResumeNewerIncoming(t *testing.T) {
	newer := storeBytes(t, binaryHead, futureLedgerRow)
	tests := []struct {
		name    string
		arrange func(t *testing.T, l leaf)
		want    func(l leaf) []string
	}{
		{
			name: "staged",
			arrange: func(t *testing.T, l leaf) {
				if err := os.WriteFile(l.dbPath, storeBytes(t, binaryHead, "PRAGMA journal_mode=DELETE"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(l.stagingDir, testIncoming), newer, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: func(l leaf) []string {
				return []string{filepath.Join(l.stagingDir, testIncoming), "copy at version 4 is newer than this binary's version 3"}
			},
		},
		{
			name: "renamed into db/",
			arrange: func(t *testing.T, l leaf) {
				if err := os.WriteFile(filepath.Join(l.snapshotsDir, testPreRestore), storeBytes(t, binaryHead, ""), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(l.dbPath, newer, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			want: func(l leaf) []string {
				return []string{l.dbPath + " at version 4 is newer than this binary's version 3"}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir, l := newLeafDir(t)
			tt.arrange(t, l)
			plantIntent(t, l, testIntent(intentSwap))
			stage := filepath.Join(l.stagingDir, testIncoming)
			db, staged := hashFile(t, l.dbPath), fileExistsHash(t, stage)

			s := openSession(t, stateDir, true)
			_, err := s.SwapOut(deadline(t, 10*time.Second))
			assertErrNames(t, err, tt.want(l)...)
			if got := hashFile(t, l.dbPath); got != db {
				t.Error("db/'s database changed under a refused resume")
			}
			if got := fileExistsHash(t, stage); got != staged {
				t.Error("the staged incoming store changed, or went, under a refused resume")
			}
		})
	}
}

// fileExistsHash returns path's SHA-256, or the zero hash when path is
// absent.
func fileExistsHash(t *testing.T, path string) [32]byte {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		return [32]byte{}
	}
	return hashFile(t, path)
}

// TestRestoreSession_StepOrder: each step refuses to run out of order,
// so no record exists before the stage is verified and the current
// store settled, and no rename runs without a record.
func TestRestoreSession_StepOrder(t *testing.T) {
	ctx := context.Background()
	t.Run("record before the verification", func(t *testing.T) {
		s := openSessionFrom(t, currentStore(t), sourceFile(t, incomingBytes(t, "")))
		if err := s.StageFile(); err != nil {
			t.Fatal(err)
		}
		if err := s.Quiesce(ctx, false); err != nil {
			t.Fatal(err)
		}
		assertErrNames(t, s.Record(RestoreIntent{}), "the stage is not verified")
		assertNoIntent(t, s.leaf)
	})
	t.Run("record before the quiesce", func(t *testing.T) {
		s := stageFrom(t, currentStore(t), incomingBytes(t, ""))
		assertErrNames(t, s.Record(RestoreIntent{}), "the current store is not quiesced")
		assertNoIntent(t, s.leaf)
	})
	t.Run("swap before the record", func(t *testing.T) {
		s := stageFrom(t, currentStore(t), incomingBytes(t, ""))
		if err := s.Quiesce(ctx, false); err != nil {
			t.Fatal(err)
		}
		_, err := s.SwapOut(ctx)
		assertErrNames(t, err, "no restore-intent record")
		if _, err := os.Stat(s.leaf.dbPath); err != nil {
			t.Errorf("current store after a refused swap: %v", err)
		}
	})
	t.Run("swap in before swap out", func(t *testing.T) {
		stateDir := currentStore(t)
		plantIntent(t, newLeaf(stateDir), testIntent(intentSwap))
		assertErrNames(t, openSession(t, stateDir, true).SwapIn(), "not set aside")
	})
	t.Run("quiesce on a resume", func(t *testing.T) {
		stateDir := currentStore(t)
		plantIntent(t, newLeaf(stateDir), testIntent(intentSwap))
		assertErrNames(t, openSession(t, stateDir, true).Quiesce(ctx, false), "restore-intent record", "settled before it")
	})
	t.Run("swap on a record past the swap", func(t *testing.T) {
		stateDir := currentStore(t)
		plantIntent(t, newLeaf(stateDir), testIntent(intentReconcile))
		_, err := openSession(t, stateDir, true).SwapOut(ctx)
		assertErrNames(t, err, "in state reconcile, past the swap")
	})
}
