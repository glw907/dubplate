package store

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// restoreNow is the clock every new restore in these tests reads; it
// formats as testStamp, so a new restore stages testIncoming.
var restoreNow = time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

// restoreOpts returns the options every restore test opens with.
func restoreOpts(resume bool) RestoreOptions {
	return RestoreOptions{BusyTimeout: openBusy, Now: restoreNow, Resume: resume, Logger: slog.New(slog.DiscardHandler)}
}

// openSession opens a restore session on stateDir, failing the test on
// a refusal, and closes it at cleanup.
func openSession(t *testing.T, stateDir string, resume bool) *RestoreSession {
	t.Helper()
	s, err := OpenRestore(stateDir, restoreOpts(resume))
	if err != nil {
		t.Fatalf("open restore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// openSessionFrom opens a new restore session on stateDir that
// restores from the local file src, failing the test on a refusal, and
// closes it at cleanup.
func openSessionFrom(t *testing.T, stateDir, src string) *RestoreSession {
	t.Helper()
	opts := restoreOpts(false)
	opts.FromFile = src
	s, err := OpenRestore(stateDir, opts)
	if err != nil {
		t.Fatalf("open restore from %s: %v", src, err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// storeFile returns the database file of a store migrated through
// version through, in a state directory of its own, with every
// connection closed: a cleanly closed WAL-header file with no -wal or
// -shm beside it.
func storeFile(t *testing.T, through int) string {
	t.Helper()
	return buildStore(t, t.TempDir(), through).dbPath
}

// sourceFile writes b to a fresh file at mode 0644 outside any leaf and
// returns its path.
func sourceFile(t *testing.T, b []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "source.db")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// fileMode returns path's permission bits.
func fileMode(t *testing.T, path string) fs.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.Mode().Perm()
}

func TestOneDevice(t *testing.T) {
	tests := []struct {
		name    string
		dirs    []dirDevice
		wantErr string
	}{
		{name: "one device", dirs: []dirDevice{{"/s/db", 60}, {"/s/staging", 60}, {"/s/snapshots", 60}}},
		{
			name:    "db elsewhere",
			dirs:    []dirDevice{{"/shm/db", 27}, {"/s/staging", 60}, {"/s/snapshots", 60}},
			wantErr: "restore renames between /shm/db (device 27), /s/staging (device 60), /s/snapshots (device 60), which must share one device",
		},
		{
			name:    "last elsewhere",
			dirs:    []dirDevice{{"/s/db", 60}, {"/s/staging", 60}, {"/m/snapshots", 61}},
			wantErr: "restore renames between /s/db (device 60), /s/staging (device 60), /m/snapshots (device 61), which must share one device",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := oneDevice(tt.dirs)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("oneDevice = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("oneDevice = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

// TestOpenRestore_Leaf: an absent leaf is created when the caller owns
// the state directory and refused, creating nothing, when it does not;
// a present leaf another uid owns is refused naming the owner and the
// runuser fix.
func TestOpenRestore_Leaf(t *testing.T) {
	cur, err := user.Current()
	if err != nil {
		t.Fatalf("current user: %v", err)
	}
	fix := "runuser -u " + cur.Username + " --"

	t.Run("absent, caller owns the state directory", func(t *testing.T) {
		stateDir := t.TempDir()
		openSession(t, stateDir, false)
		assertLeafShape(t, newLeaf(stateDir))
	})
	t.Run("absent, another uid owns the state directory", func(t *testing.T) {
		stateDir := t.TempDir()
		withGeteuid(t, os.Geteuid()+1)
		_, err := OpenRestore(stateDir, restoreOpts(false))
		assertErrNames(t, err, cur.Username, fix)
		if _, err := os.Stat(newLeaf(stateDir).root); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("leaf after a refused create: %v, want none", err)
		}
	})
	t.Run("present, another uid owns the leaf", func(t *testing.T) {
		stateDir, l := newLeafDir(t)
		before := treeListing(t, l.root)
		withGeteuid(t, os.Geteuid()+1)
		_, err := OpenRestore(stateDir, restoreOpts(false))
		assertErrNames(t, err, "leaf "+l.root, cur.Username, fix)
		if after := treeListing(t, l.root); !slices.Equal(after, before) {
			t.Errorf("leaf changed:\nbefore %q\nafter  %q", before, after)
		}
	})
}

// TestOpenRestore_LostDatabaseOpens: a leaf whose database is gone
// beside a pre-migration copy and a backup reference, which every
// other writing open refuses, opens for a restore.
func TestOpenRestore_LostDatabaseOpens(t *testing.T) {
	stateDir, l := newLeafDir(t)
	for _, name := range []string{"pre-migrate-v2.db", "backup-20260924T030000.000000Z.ref"} {
		if err := os.WriteFile(filepath.Join(l.snapshotsDir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := OpenWrite(deadline(t, 5*time.Second), stateDir, openBusy); err == nil {
		t.Fatal("OpenWrite beside a lost database: got a handle, want the refusal the fixture needs")
	}
	openSession(t, stateDir, false)
}

// TestOpenRestore_Record: a new restore refuses a leaf holding a
// record, readable or not, naming restore --resume and changing
// nothing; a resume opens beside a readable record, keeping the staged
// file it names, and refuses a missing or unreadable one, naming the
// hand recovery for the latter. Fixture: a current store and the
// staged incoming file the record names.
func TestOpenRestore_Record(t *testing.T) {
	tests := []struct {
		name    string
		plant   func(t *testing.T, l leaf)
		resume  bool
		wantErr []string
	}{
		{
			name:    "record, new restore",
			plant:   func(t *testing.T, l leaf) { plantIntent(t, l, testIntent(intentSwap)) },
			wantErr: []string{"(state swap): finish the restore with restore --resume"},
		},
		{
			name:    "unreadable record, new restore",
			plant:   func(t *testing.T, l leaf) { plantIntentFile(t, l, "malformed.json") },
			wantErr: []string{"is unreadable: parse: unexpected EOF: finish the restore with restore --resume"},
		},
		{
			name:   "record, resume",
			plant:  func(t *testing.T, l leaf) { plantIntent(t, l, testIntent(intentSwap)) },
			resume: true,
		},
		{
			name:    "no record, resume",
			plant:   func(*testing.T, leaf) {},
			resume:  true,
			wantErr: []string{"no restore-intent record at", "there is no restore to resume"},
		},
		{
			name:    "unreadable record, resume",
			plant:   func(t *testing.T, l leaf) { plantIntentFile(t, l, "unknown-state.json") },
			resume:  true,
			wantErr: []string{`is unreadable: unknown state "copying"`, "restore --resume cannot finish it", "runbook's restore-intent procedure"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead)
			staged := filepath.Join(l.stagingDir, testIncoming)
			if err := os.WriteFile(staged, []byte("incoming"), 0o600); err != nil {
				t.Fatal(err)
			}
			tt.plant(t, l)
			before := treeListing(t, l.root)

			s, err := OpenRestore(stateDir, restoreOpts(tt.resume))
			if tt.wantErr != nil {
				assertErrNames(t, err, append(tt.wantErr, l.restoreIntentPath)...)
				if after := treeListing(t, l.root); !slices.Equal(after, before) {
					t.Errorf("leaf changed:\nbefore %q\nafter  %q", before, after)
				}
				return
			}
			if err != nil {
				t.Fatalf("open restore: %v", err)
			}
			t.Cleanup(func() { _ = s.Close() })
			if s.StagePath() != staged {
				t.Errorf("stage = %s, want the record's %s", s.StagePath(), staged)
			}
			if got := listDir(t, l.stagingDir); !slices.Equal(got, []string{testIncoming}) {
				t.Errorf("staging on resume = %q, want %q kept", got, testIncoming)
			}
		})
	}
}

// TestOpenRestore_ClearsStaging: a new restore with no record removes
// every entry in staging/, rclone's partial file and a directory
// included.
func TestOpenRestore_ClearsStaging(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	for _, name := range []string{testIncoming + ".6b1f0e2a.partial", "backup-" + testInstance + ".db", "stray.txt"} {
		if err := os.WriteFile(filepath.Join(l.stagingDir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(l.stagingDir, "dir", "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	s := openSession(t, stateDir, false)
	if got := listDir(t, l.stagingDir); len(got) != 0 {
		t.Errorf("staging = %q, want empty", got)
	}
	if want := filepath.Join(l.stagingDir, testIncoming); s.StagePath() != want {
		t.Errorf("stage = %s, want %s", s.StagePath(), want)
	}
}

// TestOpenRestore_Devices: a db/ that is a symlink to a directory on
// another device is refused before anything is staged, naming each
// directory and its device. The fixture's two filesystems are /dev/shm
// and the test's temp directory, measured on this workstation as
// separate tmpfs mounts; the test first proves they differ.
func TestOpenRestore_Devices(t *testing.T) {
	stateDir, l := newLeafDir(t)
	elsewhere, err := os.MkdirTemp("/dev/shm", "dubplate-restore-db-*")
	if err != nil {
		t.Fatalf("make a directory under /dev/shm: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(elsewhere) })
	if err := os.Remove(l.dbDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, l.dbDir); err != nil {
		t.Fatal(err)
	}
	devOf := func(p string) uint64 {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			t.Fatalf("stat %s: no unix stat info", p)
		}
		return st.Dev
	}
	shmDev, stagingDev := devOf(elsewhere), devOf(l.stagingDir)
	if shmDev == stagingDev {
		t.Fatalf("%s and %s share device %d; the fixture needs two", elsewhere, l.stagingDir, shmDev)
	}

	_, err = OpenRestore(stateDir, restoreOpts(false))
	assertErrNames(t, err,
		fmt.Sprintf("%s (device %d)", elsewhere, shmDev),
		fmt.Sprintf("%s (device %d)", l.stagingDir, stagingDev),
		fmt.Sprintf("%s (device %d)", l.snapshotsDir, stagingDev),
	)
	if got := listDir(t, l.stagingDir); len(got) != 0 {
		t.Errorf("staging = %q, want nothing staged", got)
	}
}

// restoreChildEnv tells a re-executed test binary which holder of the
// store lock TestRestoreChildHold stands in for.
const restoreChildEnv = "DUBPLATE_STORE_RESTORE_CHILD"

// TestRestoreChildHold is not a test on its own: it does nothing unless
// a parent test re-executes the binary naming it. As "write", the child
// holds a write-mode handle with a read connection open; as "inspect",
// it holds the store lock shared with an inspect-DSN connection open,
// the reader doctor and check use. Either prints "ready" and holds
// until its stdin closes. As "orphan", it opens a write-mode handle,
// starts a grandchild that inherits a duplicate of its store-lock
// descriptor and its own stdin, prints "ready", and sleeps until the
// parent kills it.
func TestRestoreChildHold(t *testing.T) {
	action := os.Getenv(restoreChildEnv)
	if action == "" {
		return
	}
	stateDir := os.Getenv(childStateDirEnv)
	l := newLeaf(stateDir)
	ctx := context.Background()
	var n int
	switch action {
	case "write", "orphan":
		h, err := OpenWrite(ctx, stateDir, openBusy)
		if err != nil {
			t.Fatalf("child open write: %v", err)
		}
		if err := h.QueryRow(ctx, "SELECT count(*) FROM jobs").Scan(&n); err != nil {
			t.Fatalf("child read: %v", err)
		}
		if action == "orphan" {
			startOrphan(t, h)
		}
	case "inspect":
		if _, err := acquireSharedNonblocking(l.lockPath); err != nil {
			t.Fatalf("child lock: %v", err)
		}
		db, err := sql.Open(sqliteDriverName, inspectDSN(l.dbPath, openBusy))
		if err != nil {
			t.Fatalf("child open inspect: %v", err)
		}
		if err := db.QueryRow("SELECT count(*) FROM jobs").Scan(&n); err != nil {
			t.Fatalf("child read: %v", err)
		}
	default:
		t.Fatalf("child action %q", action)
	}
	fmt.Println("ready")
	if action == "orphan" {
		time.Sleep(time.Hour)
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}

// startOrphan starts the grandchild the "orphan" child leaves holding
// its store lock: TestLeaseChildHold, which holds its inherited
// descriptors until its stdin, this child's own, closes.
func startOrphan(t *testing.T, h *Handle) {
	dup, err := dupCloexec(h.lock.file)
	if err != nil {
		t.Fatalf("child duplicate: %v", err)
	}
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestLeaseChildHold$", "-test.count=1")
	cmd.Env = append(os.Environ(), leaseChildEnv+"=1")
	cmd.ExtraFiles = []*os.File{dup}
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		t.Fatalf("child start grandchild: %v", err)
	}
}

// startLockHolder re-executes the test binary as TestRestoreChildHold
// in role action on stateDir and waits for it to print "ready". The
// returned func releases the holder: it closes the stdin the holder
// waits on and, for a role that leaves a grandchild, waits until the
// store lock is free.
func startLockHolder(t *testing.T, stateDir, action string) (release func()) {
	t.Helper()
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestRestoreChildHold$", "-test.count=1")
	cmd.Env = append(os.Environ(), restoreChildEnv+"="+action, childStateDirEnv+"="+stateDir)
	cmd.Stdin = stdinR
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start holder: %v", err)
	}
	_ = stdinR.Close()
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if strings.TrimSpace(line) != "ready" {
		_ = stdinW.Close()
		_ = cmd.Wait()
		t.Fatalf("holder %s not ready: %q %v\n%s", action, line, err, stderr.Bytes())
	}

	if action == "orphan" {
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("kill the orphan's parent: %v", err)
		}
		_ = cmd.Wait()
	}
	var once sync.Once
	release = func() {
		once.Do(func() {
			_ = stdinW.Close()
			if action != "orphan" {
				if err := cmd.Wait(); err != nil {
					t.Errorf("holder %s: %v\n%s", action, err, stderr.Bytes())
				}
				return
			}
			limit := time.After(10 * time.Second)
			for storeLockBusy(t, stateDir) {
				select {
				case <-limit:
					t.Error("the orphaned grandchild still holds the store lock 10s after its stdin closed")
					return
				case <-time.After(lockPollInterval):
				}
			}
		})
	}
	t.Cleanup(release)
	return release
}

// TestOpenRestore_LockHeld is spec criterion 5: while a second process
// holds the store lock (a write-mode handle, an inspect-mode open, or a
// transfer child that inherited it from a parent since killed), the
// session open refuses at once, naming the lock as held, and the
// database file's bytes and mode and every entry of the leaf are
// unchanged. Fixture: a current store holding one job row.
func TestOpenRestore_LockHeld(t *testing.T) {
	for _, action := range []string{"write", "inspect", "orphan"} {
		t.Run(action, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead)
			plantJob(t, l.dbPath)
			release := startLockHolder(t, stateDir, action)
			if action == "orphan" && !storeLockBusy(t, stateDir) {
				t.Fatal("the store lock is free after the orphan's parent died; the fixture needs the grandchild holding it")
			}
			hash, mode := hashFile(t, l.dbPath), fileMode(t, l.dbPath)
			before := treeListing(t, l.root)

			type result struct {
				s   *RestoreSession
				err error
			}
			done := make(chan result, 1)
			go func() {
				s, err := OpenRestore(stateDir, restoreOpts(false))
				done <- result{s, err}
			}()
			var res result
			select {
			case res = <-done:
			case <-time.After(5 * time.Second):
				t.Error("OpenRestore still waiting on the held lock after 5s, want an immediate refusal")
				release()
				res = <-done
			}
			if res.err == nil {
				_ = res.s.Close()
				t.Fatal("OpenRestore with the lock held: got a session, want a refusal")
			}
			if !errors.Is(res.err, ErrLockBusy) || !strings.Contains(res.err.Error(), "lock "+l.lockPath+": held by ") {
				t.Errorf("OpenRestore = %v, want the lock named as held", res.err)
			}
			if hashFile(t, l.dbPath) != hash || fileMode(t, l.dbPath) != mode {
				t.Error("the database file's bytes or mode changed")
			}
			if after := treeListing(t, l.root); !slices.Equal(after, before) {
				t.Errorf("leaf changed:\nbefore %q\nafter  %q", before, after)
			}
		})
	}
}

// TestRestoreSession_StageFile is spec criterion 13's mode half on the
// local-file path: under umask 0, a 0644 source stages at 0600, byte
// for byte, and the source is untouched.
func TestRestoreSession_StageFile(t *testing.T) {
	old := syscall.Umask(0)
	t.Cleanup(func() { syscall.Umask(old) })

	src := sourceFile(t, readFile(t, storeFile(t, binaryHead)))
	srcHash := hashFile(t, src)
	s := openSessionFrom(t, currentStore(t), src)
	if err := s.StageFile(); err != nil {
		t.Fatalf("stage: %v", err)
	}
	if mode := fileMode(t, s.StagePath()); mode != 0o600 {
		t.Errorf("stage mode = %04o, want 0600", mode)
	}
	if hashFile(t, s.StagePath()) != srcHash {
		t.Error("the stage's bytes differ from the source's")
	}
	if hashFile(t, src) != srcHash || fileMode(t, src) != 0o644 {
		t.Error("the source changed")
	}
}

// TestRestoreSession_StageFileRefusals: each source a local restore
// must not read is refused naming it, and nothing is staged.
func TestRestoreSession_StageFileRefusals(t *testing.T) {
	good := readFile(t, storeFile(t, binaryHead))
	tests := []struct {
		name    string
		source  func(t *testing.T, l leaf) string
		wantErr string
	}{
		{
			name:    "absent",
			source:  func(t *testing.T, _ leaf) string { return filepath.Join(t.TempDir(), "gone.db") },
			wantErr: "no such file or directory",
		},
		{
			name:    "a directory",
			source:  func(t *testing.T, _ leaf) string { return t.TempDir() },
			wantErr: "not a regular file",
		},
		{
			name: "unreadable",
			source: func(t *testing.T, _ leaf) string {
				p := sourceFile(t, good)
				if err := os.Chmod(p, 0); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: "permission denied",
		},
		{
			name: "a -wal beside it",
			source: func(t *testing.T, _ leaf) string {
				p := sourceFile(t, good)
				if err := os.WriteFile(p+"-wal", nil, 0o600); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: "source.db-wal exists beside it",
		},
		{
			name: "a -journal beside it",
			source: func(t *testing.T, _ leaf) string {
				p := sourceFile(t, good)
				if err := os.WriteFile(p+"-journal", []byte("j"), 0o600); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: "source.db-journal exists beside it",
		},
		{
			name:    "the live database",
			source:  func(_ *testing.T, l leaf) string { return l.dbPath },
			wantErr: "it is the live database",
		},
		{
			name: "a hard link to the live database",
			source: func(t *testing.T, l leaf) string {
				p := filepath.Join(filepath.Dir(l.root), "link.db")
				if err := os.Link(l.dbPath, p); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: "it is the live database",
		},
		{
			name: "no room in staging",
			source: func(t *testing.T, _ leaf) string {
				stubFree(t, uint64(len(good))+journalSizeLimit-1)
				return sourceFile(t, good)
			},
			wantErr: "bytes free, need",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := currentStore(t)
			src := tt.source(t, newLeaf(stateDir))
			s := openSessionFrom(t, stateDir, src)
			before := listDir(t, s.leaf.stagingDir)
			err := s.StageFile()
			assertErrNames(t, err, src, tt.wantErr)
			if got := listDir(t, s.leaf.stagingDir); !slices.Equal(got, before) {
				t.Errorf("staging = %q, want %q: nothing staged", got, before)
			}
		})
	}
}

// TestOpenRestore_FromFileInStaging: a new restore from a file that
// is already in staging/ when the restore opens, or that resolves into
// it, is refused naming the path, and staging/ keeps every entry, the
// source included. The control row opens from a file outside staging/
// beside the same planted entry and finds staging/ cleared, so the
// refusal rows fail if the check moves after the clear.
func TestOpenRestore_FromFileInStaging(t *testing.T) {
	good := readFile(t, storeFile(t, binaryHead))
	// plant writes a store at name in staging/ and returns its path.
	plant := func(t *testing.T, l leaf, name string) string {
		t.Helper()
		p := filepath.Join(l.stagingDir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, good, 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	// link makes a symlink at p to target and returns p.
	link := func(t *testing.T, target, p string) string {
		t.Helper()
		if err := os.Symlink(target, p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	tests := []struct {
		name string
		// source plants the row's staging/ entries and returns the
		// file to restore from.
		source  func(t *testing.T, l leaf) string
		refused bool
	}{
		{
			name:    "a file in staging",
			source:  func(t *testing.T, l leaf) string { return plant(t, l, "other.db") },
			refused: true,
		},
		{
			name:    "a file under a directory in staging",
			source:  func(t *testing.T, l leaf) string { return plant(t, l, filepath.Join("dir", "other.db")) },
			refused: true,
		},
		{
			name: "a symlink outside staging to a file in it",
			source: func(t *testing.T, l leaf) string {
				return link(t, plant(t, l, "other.db"), filepath.Join(t.TempDir(), "via.db"))
			},
			refused: true,
		},
		{
			name: "a symlink in staging to a file outside it",
			source: func(t *testing.T, l leaf) string {
				return link(t, sourceFile(t, good), filepath.Join(l.stagingDir, "via.db"))
			},
			refused: true,
		},
		{
			name: "a file outside staging",
			source: func(t *testing.T, l leaf) string {
				plant(t, l, "other.db")
				return sourceFile(t, good)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := currentStore(t)
			l := newLeaf(stateDir)
			src := tt.source(t, l)
			before := listDir(t, l.stagingDir)
			if len(before) == 0 {
				t.Fatal("fixture planted nothing in staging")
			}
			opts := restoreOpts(false)
			opts.FromFile = src
			s, err := OpenRestore(stateDir, opts)
			if !tt.refused {
				if err != nil {
					t.Fatalf("open restore from %s: %v", src, err)
				}
				t.Cleanup(func() { _ = s.Close() })
				if got := listDir(t, l.stagingDir); len(got) != 0 {
					t.Errorf("staging = %q, want empty", got)
				}
				return
			}
			if err == nil {
				_ = s.Close()
			}
			assertErrNames(t, err, "restore from "+src, "which a restore clears")
			if got := listDir(t, l.stagingDir); !slices.Equal(got, before) {
				t.Errorf("staging = %q, want %q: nothing cleared", got, before)
			}
			if !bytes.Equal(readFile(t, src), good) {
				t.Error("the source changed")
			}
		})
	}
}

// TestRestoreSession_StageFileNoSource: a session opened without a
// file to restore from stages nothing.
func TestRestoreSession_StageFileNoSource(t *testing.T) {
	s := openSession(t, currentStore(t), false)
	assertErrNames(t, s.StageFile(), "names no file to restore from")
	if got := listDir(t, s.leaf.stagingDir); len(got) != 0 {
		t.Errorf("staging = %q, want empty", got)
	}
}

// TestRestoreSession_CheckSpace: the stage needs its size plus the WAL
// bound, and a size the caller does not know is refused.
func TestRestoreSession_CheckSpace(t *testing.T) {
	const size = 4096
	tests := []struct {
		name    string
		free    uint64
		size    int64
		wantErr string
	}{
		{name: "exactly enough", free: size + 67108864, size: size},
		{name: "one byte short", free: size + 67108863, size: size, wantErr: "67112959 bytes free, need 67112960"},
		{name: "unknown size", free: 1 << 40, size: -1, wantErr: "size -1 is unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openSession(t, currentStore(t), false)
			stubFree(t, tt.free)
			err := s.CheckSpace(tt.size)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckSpace = %v, want nil", err)
				}
				return
			}
			assertErrNames(t, err, tt.wantErr)
		})
	}
}

// truncated returns the first half of a store's pages.
func truncated(t *testing.T) []byte {
	t.Helper()
	b := readFile(t, storeFile(t, binaryHead))
	return b[:len(b)/2/4096*4096]
}

// storeBytes returns the bytes of a store through version through after
// stmts run on it, closed cleanly.
func storeBytes(t *testing.T, through int, stmts string) []byte {
	t.Helper()
	p := storeFile(t, through)
	if stmts != "" {
		rawExec(t, p, stmts)
	}
	return readFile(t, p)
}

// sqliteBytes returns the bytes of a SQLite file that is no dubplate
// store, made by stmts on an empty file.
func sqliteBytes(t *testing.T, stmts string) []byte {
	t.Helper()
	p := filepath.Join(t.TempDir(), "plain.db")
	precreateFile(t, p)
	rawExec(t, p, stmts)
	return readFile(t, p)
}

// TestRestoreSession_VerifyStage: a stage at or below the binary's
// version verifies, returning its version, with the -wal and -shm the
// verification's read left removed; each wrong-archive stage the store
// can recognize is refused with its report and removed, so no record
// could follow it.
func TestRestoreSession_VerifyStage(t *testing.T) {
	tests := []struct {
		name        string
		stage       func(t *testing.T) []byte
		wantVersion int
		wantErr     string
	}{
		{name: "current", stage: func(t *testing.T) []byte { return storeBytes(t, binaryHead, "") }, wantVersion: binaryHead},
		{name: "one version older", stage: func(t *testing.T) []byte { return storeBytes(t, 2, "") }, wantVersion: 2},
		{name: "truncated", stage: truncated, wantErr: "database disk image is malformed"},
		{name: "not SQLite", stage: func(*testing.T) []byte { return bytes.Repeat([]byte("not sqlite "), 1000) }, wantErr: "not a database"},
		{name: "zero bytes", stage: func(*testing.T) []byte { return nil }, wantErr: "application_id 0x0, want the store marker 0x64756270"},
		{
			name: "unmarked",
			stage: func(t *testing.T) []byte {
				return sqliteBytes(t, "PRAGMA application_id = 4660; CREATE TABLE t (a INTEGER);")
			},
			wantErr: "application_id 0x1234, want the store marker 0x64756270",
		},
		{
			name:    "newer",
			stage:   func(t *testing.T) []byte { return storeBytes(t, binaryHead, futureLedgerRow) },
			wantErr: "copy at version 4 is newer than this binary's version 3",
		},
		{
			name: "ledger mismatch",
			stage: func(t *testing.T) []byte {
				return storeBytes(t, 2, "UPDATE migrations SET sha256 = '"+strings.Repeat("0", 64)+"' WHERE version = 2")
			},
			wantErr: "ledger version 2 records 0002_jobs.sql (sha256 " + strings.Repeat("0", 64) + ")",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := currentStore(t)
			s := openSessionFrom(t, stateDir, sourceFile(t, tt.stage(t)))
			if err := s.StageFile(); err != nil {
				t.Fatalf("stage: %v", err)
			}

			version, err := s.VerifyStage(deadline(t, 10*time.Second))
			if tt.wantErr != "" {
				assertErrNames(t, err, s.StagePath(), tt.wantErr)
				if got := listDir(t, s.leaf.stagingDir); len(got) != 0 {
					t.Errorf("staging after a refusal = %q, want empty", got)
				}
				assertNoIntent(t, s.leaf)
				return
			}
			if err != nil {
				t.Fatalf("verify: %v", err)
			}
			if version != tt.wantVersion {
				t.Errorf("version = %d, want %d", version, tt.wantVersion)
			}
			if got := listDir(t, s.leaf.stagingDir); !slices.Equal(got, []string{testIncoming}) {
				t.Errorf("staging after verification = %q, want only %q", got, testIncoming)
			}
		})
	}
}

// TestRestoreSession_VerifyStageLeavesNoSidecars: the verification's
// read of a byte copy of a cleanly closed WAL-header store creates a
// -wal and a -shm beside the stage, and the verification removes both.
// The first half proves the read creates them, so the second cannot
// pass for want of anything to remove.
func TestRestoreSession_VerifyStageLeavesNoSidecars(t *testing.T) {
	b := storeBytes(t, binaryHead, "")
	if b[18] != 2 || b[19] != 2 {
		t.Fatalf("fixture header read/write versions = %d/%d, want WAL's 2/2", b[18], b[19])
	}
	probe := sourceFile(t, b)
	if err := verifyCopy(context.Background(), probe, openBusy, binaryHead); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(probe + suffix); err != nil {
			t.Fatalf("an inspect read of the fixture left no %s: %v", suffix, err)
		}
	}

	s := openSessionFrom(t, currentStore(t), sourceFile(t, b))
	if err := s.StageFile(); err != nil {
		t.Fatalf("stage: %v", err)
	}
	if _, err := s.VerifyStage(deadline(t, 10*time.Second)); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got := listDir(t, s.leaf.stagingDir); !slices.Equal(got, []string{testIncoming}) {
		t.Errorf("staging = %q, want only %q", got, testIncoming)
	}
}

// TestRestoreSession_VerifyStageChmods: a stage a fetch renamed in at
// 0644 under umask 022 is 0600 once verified.
func TestRestoreSession_VerifyStageChmods(t *testing.T) {
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	s := openSession(t, currentStore(t), false)
	if err := os.WriteFile(s.StagePath(), storeBytes(t, binaryHead, ""), 0o644); err != nil {
		t.Fatal(err)
	}
	if mode := fileMode(t, s.StagePath()); mode != 0o644 {
		t.Fatalf("fixture stage mode = %04o, want 0644", mode)
	}
	if _, err := s.VerifyStage(deadline(t, 10*time.Second)); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if mode := fileMode(t, s.StagePath()); mode != 0o600 {
		t.Errorf("stage mode = %04o, want 0600", mode)
	}
}

// TestRestoreSession_VerifyStageHotWAL: a stage beside a -wal holding
// frames is refused naming the -wal, since the stage alone lacks what
// it holds. Fixture: a child migrates a version-1 store to 2 and exits
// without closing, leaving the migration only in its -wal; the stage
// and its -wal are byte copies of the two, taken with no connection
// open.
func TestRestoreSession_VerifyStageHotWAL(t *testing.T) {
	hotDir := t.TempDir()
	hot := buildStore(t, hotDir, 1)
	runChildWrite(t, hotDir, "older")

	s := openSession(t, currentStore(t), false)
	if err := os.WriteFile(s.StagePath(), readFile(t, hot.dbPath), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.StagePath()+"-wal", readFile(t, hot.dbPath+"-wal"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := s.VerifyStage(deadline(t, 10*time.Second))
	assertErrNames(t, err, s.StagePath()+"-wal", "the stage alone lacks")
	if got := listDir(t, s.leaf.stagingDir); len(got) != 0 {
		t.Errorf("staging after a refusal = %q, want empty", got)
	}
}

// TestRestoreSession_VerifyStageKeepsRecordedStage: on a resume, a
// stage the record names that fails verification is refused and kept.
func TestRestoreSession_VerifyStageKeepsRecordedStage(t *testing.T) {
	stateDir := currentStore(t)
	l := newLeaf(stateDir)
	plantIntent(t, l, testIntent(intentSwap))
	s := openSession(t, stateDir, true)
	if err := os.WriteFile(s.StagePath(), storeBytes(t, binaryHead, futureLedgerRow), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := s.VerifyStage(deadline(t, 10*time.Second))
	assertErrNames(t, err, "copy at version 4 is newer than this binary's version 3")
	if _, err := os.Stat(s.StagePath()); err != nil {
		t.Errorf("the recorded stage after a refusal: %v, want it kept", err)
	}
}

// TestCopyInto_ExactVersion: backup's copy keeps exact equality, so a
// store one version below the version the copy wants is refused.
func TestCopyInto_ExactVersion(t *testing.T) {
	src := storeFile(t, binaryHead-1)
	db, err := openMaintenanceConn(src, openBusy)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	staged := filepath.Join(t.TempDir(), "copy.db")
	precreateFile(t, staged)

	err = copyInto(context.Background(), db, staged, openBusy, binaryHead)
	assertErrNames(t, err, staged, "copy at version 2, want 3")
}

// TestRestoreSession_OpenRead: while the session holds the lock
// exclusively, a read handle on a current store opens and reads at
// once; a store below the binary's version is refused.
func TestRestoreSession_OpenRead(t *testing.T) {
	t.Run("current", func(t *testing.T) {
		stateDir := t.TempDir()
		l := buildStore(t, stateDir, binaryHead)
		id := plantJob(t, l.dbPath)
		s := openSession(t, stateDir, false)

		var got string
		err := assertWithin(t, 5*time.Second, func() error {
			h, err := s.OpenRead(deadline(t, 10*time.Second))
			if err != nil {
				return err
			}
			return errors.Join(h.QueryRow(context.Background(), "SELECT hex(id) FROM jobs").Scan(&got), h.Close())
		})
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if got != id {
			t.Errorf("job id = %s, want %s", got, id)
		}
		if !storeLockBusy(t, stateDir) {
			t.Error("the store lock is free with the session open")
		}
	})
	t.Run("older", func(t *testing.T) {
		stateDir := t.TempDir()
		buildStore(t, stateDir, 2)
		s := openSession(t, stateDir, false)
		_, err := s.OpenRead(deadline(t, 10*time.Second))
		assertErrNames(t, err, "at version 2, binary at version 3: only serve opens it")
	})
}

// TestRestoreSession_WriteBackupRef: two writes leave one reference,
// naming the second stamp and its remote.
func TestRestoreSession_WriteBackupRef(t *testing.T) {
	s := openSession(t, currentStore(t), false)
	if err := s.WriteBackupRef("20260924T030000.000000Z", "first:a"); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteBackupRef("20260925T030000.000000Z", "second:b"); err != nil {
		t.Fatal(err)
	}
	want := "backup-20260925T030000.000000Z.ref"
	if got := listDir(t, s.leaf.snapshotsDir); !slices.Equal(got, []string{want}) {
		t.Fatalf("snapshots = %q, want %q", got, want)
	}
	if got := string(readFile(t, filepath.Join(s.leaf.snapshotsDir, want))); got != "stamp 20260925T030000.000000Z\nremote second:b\n" {
		t.Errorf("reference reads %q", got)
	}
}

// TestRestoreSession_DupFiles: the duplicate is close-on-exec, and a
// child that inherits it holds the store lock after the session closes,
// until the child exits.
func TestRestoreSession_DupFiles(t *testing.T) {
	stateDir := currentStore(t)
	s, err := OpenRestore(stateDir, restoreOpts(false))
	if err != nil {
		t.Fatal(err)
	}
	dups, err := s.DupFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(dups) != 1 || !fdCloexec(t, dups[0]) {
		t.Fatalf("duplicates = %d, cloexec %v; want one, close-on-exec", len(dups), len(dups) == 1 && fdCloexec(t, dups[0]))
	}

	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestLeaseChildHold$", "-test.count=1")
	cmd.Env = append(os.Environ(), leaseChildEnv+"=1")
	cmd.ExtraFiles = dups
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdin.Close(); _ = cmd.Wait() })
	if err := errors.Join(dups[0].Close(), s.Close()); err != nil {
		t.Fatal(err)
	}
	if !storeLockBusy(t, stateDir) {
		t.Error("the store lock is free with the child alive")
	}
	if err := stdin.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("child: %v", err)
	}
	if storeLockBusy(t, stateDir) {
		t.Error("the store lock is held after the child exited")
	}
}

// TestOpenRestore_NilLogger: the options are refused before any file
// is touched.
func TestOpenRestore_NilLogger(t *testing.T) {
	stateDir := t.TempDir()
	opts := restoreOpts(false)
	opts.Logger = nil
	_, err := OpenRestore(stateDir, opts)
	assertErrNames(t, err, "nil logger")
	if got := listDir(t, stateDir); len(got) != 0 {
		t.Errorf("state directory = %q, want untouched", got)
	}
}

// swappedLeaf returns a state directory whose swap has finished, as a
// restore with no current store leaves it: db/ holds a store through
// version through, after stmts, staging/ is empty, and the leaf's
// record, in state, names no pre-restore copy and carries library.
func swappedLeaf(t *testing.T, through int, stmts string, state intentState, library *libraryOwed) (string, leaf) {
	t.Helper()
	stateDir, l := newLeafDir(t)
	if err := os.WriteFile(l.dbPath, storeBytes(t, through, stmts), 0o600); err != nil {
		t.Fatal(err)
	}
	r := testIntent(state)
	r.PreRestore = ""
	r.Library = library
	plantIntent(t, l, r)
	return stateDir, l
}

// TestRestoreSession_Advance: each record state advances to the step
// the store in db/ earns under this binary, whatever version the record
// carries, and a refusal leaves the record as it was. The unfinished
// swap is the no-copy one with its incoming store still in staging/. Every fixture's
// record says version 2, so a verdict read from it would owe serve.
func TestRestoreSession_Advance(t *testing.T) {
	lib := &libraryOwed{Root: "/srv/library"}
	tests := []struct {
		name      string
		through   int
		stmts     string
		state     intentState
		library   *libraryOwed
		stage     bool
		want      RestoreStep
		wantState intentState
		wantErr   []string
	}{
		{name: "swap done, no library, current version", through: binaryHead, state: intentSwap, want: StepReconcile, wantState: intentReconcile},
		{name: "swap done, library owed", through: binaryHead, state: intentSwap, library: lib, want: StepLibrary, wantState: intentLibrary},
		{name: "swap done, older store", through: 2, state: intentSwap, want: StepServe, wantState: intentReconcilePending},
		{name: "library copied, current version", through: binaryHead, state: intentLibrary, library: lib, want: StepReconcile, wantState: intentReconcile},
		{name: "reconcile under a newer binary", through: 2, state: intentReconcile, want: StepServe, wantState: intentReconcilePending},
		{name: "reconcile, current version", through: binaryHead, state: intentReconcile, want: StepReconcile, wantState: intentReconcile},
		{name: "reconcile-pending stays", through: 2, state: intentReconcilePending, want: StepServe, wantState: intentReconcilePending},
		{
			name: "reconcile under an older binary", through: binaryHead, stmts: futureLedgerRow, state: intentReconcile, wantState: intentReconcile,
			wantErr: []string{"at version 4 is newer than this binary's version 3"},
		},
		{
			name: "swap unfinished", through: binaryHead, state: intentSwap, stage: true, wantState: intentSwap,
			wantErr: []string{"the swap has not finished"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir, l := swappedLeaf(t, tt.through, tt.stmts, tt.state, tt.library)
			if tt.stage {
				mv(t, l.dbPath, filepath.Join(l.stagingDir, testIncoming))
			}
			s := openSession(t, stateDir, true)
			got, err := s.Advance(deadline(t, 10*time.Second))
			if tt.wantErr != nil {
				assertErrNames(t, err, tt.wantErr...)
			} else if err != nil {
				t.Fatalf("Advance(): %v", err)
			}
			if got != tt.want {
				t.Errorf("Advance() = %d, want %d", got, tt.want)
			}
			if r := readRecord(t, l); r.State != tt.wantState || r.Version != 2 {
				t.Errorf("record state %s version %d, want %s version 2", r.State, r.Version, tt.wantState)
			}
			if s.Owed() != stepOf[tt.wantState] {
				t.Errorf("Owed() = %d, want %d", s.Owed(), stepOf[tt.wantState])
			}
		})
	}
}

// TestRestoreSession_AdvanceNoRecord: a new restore has nothing to
// advance.
func TestRestoreSession_AdvanceNoRecord(t *testing.T) {
	s := openSession(t, currentStore(t), false)
	_, err := s.Advance(context.Background())
	assertErrNames(t, err, "no restore-intent record")
	if s.Owed() != 0 {
		t.Errorf("Owed() = %d before any record, want 0", s.Owed())
	}
}

// TestRestoreSession_Intent: a resume reads back the source and library
// half the record carries, and a new restore reads the zero value.
func TestRestoreSession_Intent(t *testing.T) {
	stateDir, _ := swappedLeaf(t, binaryHead, "", intentLibrary, &libraryOwed{Root: "/srv/library", IntoNonempty: true})
	s := openSession(t, stateDir, true)
	want := RestoreIntent{Remote: "remote:dubplate", Snapshot: "20260924T030000.000000Z", LibraryRoot: "/srv/library", IntoNonempty: true}
	if got := s.Intent(); got != want {
		t.Errorf("Intent() = %+v, want %+v", got, want)
	}
	if got := s.Stamp(); got != testStamp {
		t.Errorf("Stamp() = %q, want %q", got, testStamp)
	}
	if got := openSession(t, currentStore(t), false).Intent(); got != (RestoreIntent{}) {
		t.Errorf("Intent() on a new restore = %+v, want the zero value", got)
	}
}

// TestRestoreSession_ReportAndComplete: the report is written only when
// the record owes it, replaces an earlier attempt's at 0600, and the
// record goes only once this session has written it.
func TestRestoreSession_ReportAndComplete(t *testing.T) {
	t.Run("reconcile", func(t *testing.T) {
		stateDir, l := swappedLeaf(t, binaryHead, "", intentReconcile, nil)
		p := filepath.Join(l.root, "restore-report-"+testStamp+".txt")
		if err := os.WriteFile(p, []byte("an interrupted attempt's report\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		s := openSession(t, stateDir, true)
		assertErrNames(t, s.Complete(), "report is not written")
		if _, err := os.Stat(l.restoreIntentPath); err != nil {
			t.Fatalf("record after a refused complete: %v", err)
		}
		got, err := s.WriteReport([]byte("the report\n"))
		if err != nil {
			t.Fatalf("WriteReport(): %v", err)
		}
		if got != p {
			t.Errorf("WriteReport() = %q, want %q", got, p)
		}
		assertReport(t, l, "the report\n")
		if err := s.Complete(); err != nil {
			t.Fatalf("Complete(): %v", err)
		}
		assertNoIntent(t, l)
	})
	for _, state := range []intentState{intentSwap, intentLibrary, intentReconcilePending} {
		t.Run(string(state), func(t *testing.T) {
			stateDir, l := swappedLeaf(t, binaryHead, "", state, &libraryOwed{Root: "/srv/library"})
			s := openSession(t, stateDir, true)
			_, err := s.WriteReport([]byte("the report\n"))
			assertErrNames(t, err, "does not owe one")
			if _, err := os.Stat(l.reportPath(testStamp)); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("report beside a record in state %s: %v, want none", state, err)
			}
			assertErrNames(t, s.Complete(), "report is not written")
			if r := readRecord(t, l); r.State != state {
				t.Errorf("record state %s, want %s", r.State, state)
			}
		})
	}
}

// TestRestoreSession_BackupRemote: the remote the newest backup
// reference names, "" with none, and a refusal for a reference naming
// no remote.
func TestRestoreSession_BackupRemote(t *testing.T) {
	const older, newer = "20260923T030000.000000Z", "20260924T030000.000000Z"
	tests := []struct {
		name    string
		refs    map[string]string
		want    string
		wantErr string
	}{
		{name: "none", want: ""},
		{name: "one", refs: map[string]string{newer: "stamp " + newer + "\nremote b2:dubplate\n"}, want: "b2:dubplate"},
		{
			name: "two, a prune interrupted",
			refs: map[string]string{
				older: "stamp " + older + "\nremote old:path\n",
				newer: "stamp " + newer + "\nremote new:path\n",
			},
			want: "new:path",
		},
		{name: "no remote line", refs: map[string]string{newer: "stamp " + newer + "\n"}, wantErr: "names no remote"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir, l := newLeafDir(t)
			for stamp, body := range tt.refs {
				if err := os.WriteFile(filepath.Join(l.snapshotsDir, "backup-"+stamp+".ref"), []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := openSession(t, stateDir, false).BackupRemote()
			if tt.wantErr != "" {
				assertErrNames(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("BackupRemote(): %v", err)
			}
			if got != tt.want {
				t.Errorf("BackupRemote() = %q, want %q", got, tt.want)
			}
		})
	}
}
