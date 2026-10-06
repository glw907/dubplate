package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// restoreVerbChildEnv carries, as a JSON array, the arguments a
// re-executed test binary passes to the real restore verb, run through
// main's own run so its exit status is the one an operator sees.
const restoreVerbChildEnv = "RESTORE_TEST_VERB_ARGS"

// runRestoreChild runs the restore verb with the arguments raw holds
// and returns its exit status.
func runRestoreChild(raw string) int {
	var args []string
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		fmt.Fprintf(os.Stderr, "restore child arguments %q: %v\n", raw, err)
		return 2
	}
	cmd := newRestoreCmd(os.Environ(), newProductionRegistry(), os.Stdout, os.Stderr)
	cmd.SetArgs(args)
	return run(cmd)
}

// The two tracks every backed-up fixture holds, by path relative to
// the library root, with the bytes of each file.
var restoreFixtureFiles = map[string]string{
	"main/a.flac": "the first track's bytes",
	"main/b.flac": "the second track's bytes",
}

// The fixture's rows: the owner the backup holds, the display name a
// later write changes, and the member a later write adds.
const (
	restoreOwner        = "owner@example.org"
	restoreOwnerName    = "Owner"
	restoreChangedName  = "Changed after the backup"
	restoreAddedAddress = "added@example.org"
)

// TestRestoreCmd_LongNamesExitCodes proves the verb's help lists each
// of the four exit statuses the verb assigns, one per line.
func TestRestoreCmd_LongNamesExitCodes(t *testing.T) {
	long := newRestoreCmd(nil, productionRegistry(), io.Discard, io.Discard).Long
	for _, want := range []string{
		"\n  0    the restore finished",
		"\n  1    a flag, load, or store refusal",
		"\n  75   another process holds the store's lock",
		"\n  130  interrupted by a signal (SIGINT, SIGTERM, SIGHUP)",
	} {
		if !strings.Contains(long, want) {
			t.Errorf("Long does not contain %q:\n%s", want, long)
		}
	}
}

// TestRestoreCmd_FlagRefusals runs each refused set of flags against a
// nil environment: the verb must refuse on the flags alone, naming
// them, since a load would fail on the empty wall with a different
// error, and no store or rclone is reachable from there.
func TestRestoreCmd_FlagRefusals(t *testing.T) {
	tests := []struct {
		name string
		f    restoreFlags
		want string
	}{
		{name: "no source", f: restoreFlags{}, want: "restore: name one source: --snapshot, --from-file, or --resume"},
		{name: "a snapshot and a file", f: restoreFlags{snapshot: "latest", fromFile: "x.db"}, want: "restore: --snapshot and --from-file never combine: name one source"},
		{name: "a file and a resume", f: restoreFlags{fromFile: "x.db", resume: true}, want: "restore: --from-file and --resume never combine: name one source"},
		{name: "all three sources", f: restoreFlags{snapshot: "latest", fromFile: "x.db", resume: true}, want: "restore: --snapshot and --from-file and --resume never combine: name one source"},
		{name: "older with a file", f: restoreFlags{fromFile: "x.db", older: true}, want: "restore: --older never combines with --from-file"},
		{name: "older with a resume", f: restoreFlags{resume: true, older: true}, want: "restore: --older never combines with --resume"},
		{name: "older with the library", f: restoreFlags{snapshot: "latest", older: true, library: true}, want: "restore: --older never combines with --library"},
		{name: "the library with a file", f: restoreFlags{fromFile: "x.db", library: true}, want: "restore: --library never combines with --from-file"},
		{name: "the library with a resume", f: restoreFlags{resume: true, library: true}, want: "restore: --library never combines with --resume"},
		{name: "discard with a resume", f: restoreFlags{resume: true, discardCurrent: true}, want: "restore: --discard-current never combines with --resume"},
		{name: "into-nonempty without the library", f: restoreFlags{snapshot: "latest", intoNonempty: true}, want: "restore: --into-nonempty applies only with --library"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := restoreRunFor(context.Background(), nil, tt.f, io.Discard, io.Discard)
			if code != 1 {
				t.Errorf("exit = %d, want 1", code)
			}
			if err == nil || err.Error() != tt.want {
				t.Errorf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

// TestRestoreFlags_CheckAccepts is each set of flags the verb runs:
// check must pass them all, so the refusals above are the rules and
// not a check that refuses everything.
func TestRestoreFlags_CheckAccepts(t *testing.T) {
	tests := []struct {
		name string
		f    restoreFlags
	}{
		{name: "the latest snapshot", f: restoreFlags{snapshot: "latest"}},
		{name: "an older snapshot", f: restoreFlags{snapshot: "20260925T120000.000000Z", older: true}},
		{name: "a snapshot and the library", f: restoreFlags{snapshot: "latest", library: true}},
		{name: "the library into a non-empty root", f: restoreFlags{snapshot: "latest", library: true, intoNonempty: true}},
		{name: "a snapshot discarding the current store", f: restoreFlags{snapshot: "latest", discardCurrent: true}},
		{name: "a file", f: restoreFlags{fromFile: "x.db"}},
		{name: "a file discarding the current store", f: restoreFlags{fromFile: "x.db", discardCurrent: true}},
		{name: "a resume", f: restoreFlags{resume: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.check(); err != nil {
				t.Errorf("check() = %v, want nil", err)
			}
		})
	}
}

// TestRestoreCmd_ConfigLoadFails is the load's refusal: a wall
// missing its required keys exits 1 with the load's error, before any
// leaf is created.
func TestRestoreCmd_ConfigLoadFails(t *testing.T) {
	stateDir := t.TempDir()
	environ := []string{"DUBPLATE_STATE_DIR=" + stateDir}
	code, err := restoreRunFor(context.Background(), environ, restoreFlags{fromFile: "x.db"}, io.Discard, io.Discard)
	if code != 1 {
		t.Errorf("exit = %d, want 1; error: %v", code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "DUBPLATE_LIBRARY_ROOT") {
		t.Errorf("error = %v, want the load's error naming the unset DUBPLATE_LIBRARY_ROOT", err)
	}
	if names := readDirNames(t, stateDir); len(names) != 0 {
		t.Errorf("state directory holds %v, want nothing created", names)
	}
}

// TestRestoreCmd_SnapshotWithIntegrationDisabled is a --snapshot
// restore with the backup integration off: it exits 1 naming the
// integration's flag, and creates no leaf.
func TestRestoreCmd_SnapshotWithIntegrationDisabled(t *testing.T) {
	environ := serveAuthEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	code, err := restoreRunFor(context.Background(), environ, restoreFlags{snapshot: "latest"}, io.Discard, io.Discard)
	if code != 1 {
		t.Errorf("exit = %d, want 1; error: %v", code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "backup integration is disabled (DUBPLATE_BACKUP_ENABLED)") {
		t.Errorf("error = %v, want it to name DUBPLATE_BACKUP_ENABLED", err)
	}
	if names := readDirNames(t, stateDir); len(names) != 0 {
		t.Errorf("state directory holds %v, want nothing created", names)
	}
}

// TestRestoreCmd_RecordPresent is a new --snapshot restore beside a
// restore-intent record, here an empty one: it exits 1 naming
// restore --resume and runs no rclone.
func TestRestoreCmd_RecordPresent(t *testing.T) {
	useBackupStandin(t)
	record := os.Getenv(backupStandinRecordEnv)
	environ := backupJobEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	if err := os.WriteFile(filepath.Join(stateDir, "store", "restore-intent"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	code, err := restoreRunFor(context.Background(), environ, restoreFlags{snapshot: "latest"}, io.Discard, io.Discard)
	if code != 1 {
		t.Errorf("exit = %d, want 1; error: %v", code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "restore --resume") {
		t.Errorf("error = %v, want it to name restore --resume", err)
	}
	if lines := backupStandinLines(t, record); len(lines) != 0 {
		t.Errorf("rclone ran %v, want no run", lines)
	}
}

// TestRestoreVerb_LockHeld runs the verb as a separate process while
// this one holds the store's lock through a write-mode handle: the
// verb exits 75 naming the holder, and the store is untouched.
func TestRestoreVerb_LockHeld(t *testing.T) {
	environ := serveAuthEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	seedBackupStore(t, stateDir)
	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })

	src := filepath.Join(t.TempDir(), "absent.db")
	code, _, stderr := runRestoreVerb(t, environ, "--from-file", src)
	if code != exitRestoreLocked {
		t.Errorf("exit = %d, want %d; stderr:\n%s", code, exitRestoreLocked, stderr)
	}
	if !strings.Contains(stderr, "store.lock: held by ") {
		t.Errorf("stderr = %q, want it to name the lock's holder", stderr)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "store", "restore-intent")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("restore-intent record: %v, want none", err)
	}
}

// TestRestoreVerb_SignalDuringListing interrupts the verb while the
// snapshots/ listing blocks in the stand-in, which honors SIGINT. Each
// signal, to the verb alone or as a terminal sends it to the verb's
// group, exits 130, leaves no record and an empty staging/, and reaches
// the stand-in as exactly one SIGINT, the transport's own: the stand-in
// runs in its own process group.
func TestRestoreVerb_SignalDuringListing(t *testing.T) {
	tests := []struct {
		name    string
		signal  syscall.Signal
		toGroup bool
	}{
		{name: "SIGTERM to the verb alone", signal: syscall.SIGTERM},
		{name: "SIGHUP to the verb alone", signal: syscall.SIGHUP},
		{name: "SIGINT to the verb's process group", signal: syscall.SIGINT, toGroup: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			standinDir := standinOnly(t)
			files := t.TempDir()
			record := filepath.Join(files, "record")
			pidFile := filepath.Join(files, "pid")
			environ := append(verbChildEnviron(t, standinDir),
				backupStandinRecordEnv+"="+record,
				backupStandinPIDEnv+"="+pidFile,
				backupStandinSleepEnv+"=lsjson",
				backupStandinHonorEnv+"=1",
			)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")

			v := startRestoreVerb(t, environ, "--snapshot", "latest")
			standinPID := awaitBackupStandinPID(t, pidFile)
			v.signal(t, tt.signal, tt.toGroup)
			if code := v.wait(t); code != exitRestoreInterrupted {
				t.Errorf("exit = %d, want %d; stderr:\n%s", code, exitRestoreInterrupted, v.stderr.String())
			}
			waitBackupStandinGone(t, standinPID, 5*time.Second)

			lines := backupStandinLines(t, record)
			if n := countLines(lines, "SIGINT"); n != 1 {
				t.Errorf("the stand-in recorded %d SIGINT lines, want 1: %v", n, lines)
			}
			if !strings.HasPrefix(firstOr(lines), "argv lsjson") {
				t.Errorf("the stand-in's first run = %q, want the snapshots/ listing", firstOr(lines))
			}
			if _, err := os.Stat(filepath.Join(stateDir, "store", "restore-intent")); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("restore-intent record: %v, want none", err)
			}
			if names := readDirNames(t, filepath.Join(stateDir, "store", "staging")); len(names) != 0 {
				t.Errorf("staging/ holds %v, want nothing", names)
			}
		})
	}
}

// TestRestoreVerb_SignalDuringLibraryCopyThenResume interrupts a
// library restore while its copy blocks in the stand-in, every earlier
// rclone run passed through to the real rclone. The verb exits 130
// with the record owing the library and the copy's rclone ended. A
// resume with the real rclone first on PATH then exits 0 and leaves
// the tree the backup holds, the record gone, and the report written.
func TestRestoreVerb_SignalDuringLibraryCopyThenResume(t *testing.T) {
	fix := backedUpLeaf(t)

	standinDir := standinOnly(t)
	files := t.TempDir()
	record := filepath.Join(files, "record")
	pidFile := filepath.Join(files, "pid")
	environ := append(fix.childEnviron(standinDir),
		backupStandinRecordEnv+"="+record,
		backupStandinPIDEnv+"="+pidFile,
		backupStandinSleepEnv+"=copy",
		backupStandinHonorEnv+"=1",
		backupStandinExecEnv+"="+fix.rclone,
	)

	v := startRestoreVerb(t, environ, "--snapshot", "latest", "--library")
	standinPID := awaitBackupStandinPID(t, pidFile)
	v.signal(t, syscall.SIGINT, false)
	if code := v.wait(t); code != exitRestoreInterrupted {
		t.Fatalf("exit = %d, want %d; stderr:\n%s", code, exitRestoreInterrupted, v.stderr.String())
	}
	waitBackupStandinGone(t, standinPID, 5*time.Second)
	if n := countLines(backupStandinLines(t, record), "SIGINT"); n != 1 {
		t.Errorf("the copy's stand-in recorded %d SIGINT lines, want 1", n)
	}
	if reason := recordReason(t, fix.stateDir); !strings.Contains(reason, "(state library)") {
		t.Errorf("the record after the interrupt: %q, want it to owe the library", reason)
	}

	code, stdout, stderr := runRestoreVerb(t, fix.childEnviron(filepath.Dir(fix.rclone)), "--resume")
	if code != 0 {
		t.Fatalf("resume exit = %d, want 0; stderr:\n%s", code, stderr)
	}
	assertLibrary(t, fix.root)
	assertRestored(t, fix.stateDir)
	if !strings.HasSuffix(stdout, restoreRotation+"\n") {
		t.Errorf("stdout does not end with the rotation instruction:\n%s", stdout)
	}
}

// TestRestoreVerb_EndToEnd restores a store and its library from a
// backup the backup verb wrote to a real rclone local remote. After the
// backup, one row is added, one row changed, and the library root
// emptied: the restore exits 0, the added row is gone, the changed row
// holds its backed-up value, every backed-up file is back with its
// bytes, and the output ends with the rotation instruction naming both
// secrets.
func TestRestoreVerb_EndToEnd(t *testing.T) {
	fix := backedUpLeaf(t)

	code, stdout, stderr := runRestoreVerb(t, fix.childEnviron(filepath.Dir(fix.rclone)), "--snapshot", "latest", "--library")
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr:\n%s", code, stderr)
	}
	added, owner := restoreRows(t, fix.stateDir)
	if added != 0 {
		t.Errorf("the member added after the backup: %d rows, want 0", added)
	}
	if owner != restoreOwnerName {
		t.Errorf("the owner's display name = %q, want the backed-up %q", owner, restoreOwnerName)
	}
	assertLibrary(t, fix.root)
	assertRestored(t, fix.stateDir)

	for _, want := range []string{
		"tracks without a file: 0\nfiles without a track: 0\n",
		"restored; the reconciliation report is " + filepath.Join(fix.stateDir, "store", "restore-report-"),
		"is not a mount point",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout)
		}
	}
	if !strings.HasSuffix(stdout, restoreRotation+"\n") {
		t.Errorf("stdout does not end with the rotation instruction:\n%s", stdout)
	}
	for _, key := range []string{"DUBPLATE_SESSION_SECRET", "DUBPLATE_MAGIC_LINK_SECRET", "DUBPLATE_SESSION_SECRET_PREVIOUS", "DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS"} {
		if !strings.Contains(restoreRotation, key) {
			t.Errorf("the rotation instruction does not name %s", key)
		}
	}
}

// TestRestoreCmd_ReconcilePending restores a file one version below
// this binary: the restore leaves the record for serve and exits 0,
// saying serve must start, then prints the rotation instruction. A
// resume of that record refuses, exiting 1 and naming serve.
func TestRestoreCmd_ReconcilePending(t *testing.T) {
	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_LIBRARY_ROOT", t.TempDir())
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	src := storeFile(t, 1)

	var stdout bytes.Buffer
	code, err := restoreRunFor(context.Background(), environ, restoreFlags{fromFile: src}, &stdout, io.Discard)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; error: %v", code, err)
	}
	if !strings.Contains(stdout.String(), "start serve to finish the restore\n") {
		t.Errorf("stdout does not say serve must start:\n%s", stdout.String())
	}
	if !strings.HasSuffix(stdout.String(), restoreRotation+"\n") {
		t.Errorf("stdout does not end with the rotation instruction:\n%s", stdout.String())
	}
	if reason := recordReason(t, stateDir); !strings.Contains(reason, "(state reconcile-pending)") {
		t.Errorf("the record: %q, want it reconcile-pending", reason)
	}

	code, err = restoreRunFor(context.Background(), environ, restoreFlags{resume: true}, io.Discard, io.Discard)
	if code != 1 {
		t.Errorf("resume exit = %d, want 1; error: %v", code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "start serve to finish the restore") {
		t.Errorf("resume error = %v, want it to name serve", err)
	}
}

// TestRestoreCmd_OutputLostAfterCompletion restores a current-version
// file with a stdout whose every write fails. The restore still
// completes and exits 0, since a lost line of output changes nothing
// on disk, and stderr says the output was lost and repeats the
// rotation instruction, which the report file does not carry.
func TestRestoreCmd_OutputLostAfterCompletion(t *testing.T) {
	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_LIBRARY_ROOT", t.TempDir())
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	src := storeFile(t, 0)

	var stderr bytes.Buffer
	code, err := restoreRunFor(context.Background(), environ, restoreFlags{fromFile: src}, failingWriter{}, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; error: %v", code, err)
	}
	assertRestored(t, stateDir)
	if !strings.Contains(stderr.String(), "restore: finished, but its output to stdout was lost: stdout closed") {
		t.Errorf("stderr = %q, want it to say the output was lost", stderr.String())
	}
	if !strings.Contains(stderr.String(), restoreRotation+"\n") {
		t.Errorf("stderr = %q, want it to carry the rotation instruction", stderr.String())
	}
}

// failingWriter fails every write, as a closed stdout does.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("stdout closed") }

// restoreRunFor runs the restore verb's logic in this process and
// returns its exit status as run would, with the error.
func restoreRunFor(ctx context.Context, environ []string, f restoreFlags, stdout, stderr io.Writer) (int, error) {
	err := runRestore(ctx, environ, productionRegistry(), f, stdout, stderr)
	if err == nil {
		return 0, nil
	}
	if ece, ok := errors.AsType[*exitCodeErr](err); ok {
		return ece.code, err
	}
	return 1, err
}

// A restoreVerb is the restore verb running as a re-executed test
// binary, the leader of its own process group.
type restoreVerb struct {
	cmd            *exec.Cmd
	stdout, stderr bytes.Buffer
	exited         chan error
}

// startRestoreVerb starts the verb with args under environ alone, and
// kills its process group at cleanup.
func startRestoreVerb(t *testing.T, environ []string, args ...string) *restoreVerb {
	t.Helper()
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	v := &restoreVerb{exited: make(chan error, 1)}
	v.cmd = exec.Command(os.Args[0], "-test.run=^$")
	v.cmd.Env = append(slices.Clone(environ), restoreVerbChildEnv+"="+string(raw))
	v.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	v.cmd.Stdout = &v.stdout
	v.cmd.Stderr = &v.stderr
	if err := v.cmd.Start(); err != nil {
		t.Fatalf("start the restore verb: %v", err)
	}
	done := make(chan struct{})
	go func() {
		v.exited <- v.cmd.Wait()
		close(done)
	}()
	t.Cleanup(func() {
		_ = syscall.Kill(-v.cmd.Process.Pid, syscall.SIGKILL)
		<-done
	})
	return v
}

// signal sends sig to the verb, or to its whole process group.
func (v *restoreVerb) signal(t *testing.T, sig syscall.Signal, toGroup bool) {
	t.Helper()
	target := v.cmd.Process.Pid
	if toGroup {
		target = -target
	}
	if err := syscall.Kill(target, sig); err != nil {
		t.Fatalf("signal the restore verb: %v", err)
	}
}

// wait waits for the verb to exit and returns its exit status, -1 for
// a death by a signal.
func (v *restoreVerb) wait(t *testing.T) int {
	t.Helper()
	select {
	case err := <-v.exited:
		if err == nil {
			return 0
		}
		ee, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			t.Fatalf("the restore verb: %v", err)
		}
		return ee.ExitCode()
	case <-time.After(60 * time.Second):
		t.Fatalf("the restore verb did not exit in 60s; stderr:\n%s", v.stderr.String())
		return 0
	}
}

// runRestoreVerb runs the verb to its exit and returns its status and
// both streams.
func runRestoreVerb(t *testing.T, environ []string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	v := startRestoreVerb(t, environ, args...)
	code = v.wait(t)
	return code, v.stdout.String(), v.stderr.String()
}

// standinOnly returns a directory holding only the stand-in rclone.
func standinOnly(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("test binary: %v", err)
	}
	dir := t.TempDir()
	if err := os.Symlink(exe, filepath.Join(dir, "rclone")); err != nil {
		t.Fatalf("link the stand-in: %v", err)
	}
	return dir
}

// A restoreFixture is a leaf whose store and library a real backup
// holds on a local rclone remote, changed since.
type restoreFixture struct {
	environ  []string
	stateDir string
	root     string

	// rclone is the real rclone's path.
	rclone string

	// rcloneEnv configures the remote for a child process.
	rcloneEnv []string
}

// childEnviron returns the fixture's environment for a child process,
// with PATH holding only pathDir.
func (f restoreFixture) childEnviron(pathDir string) []string {
	env := append(slices.Clone(f.environ), f.rcloneEnv...)
	return append(env, "PATH="+pathDir)
}

// backedUpLeaf builds the end-to-end fixture: a current store holding
// an owner and two tracks whose files the library root holds, a backup
// of both through the backup verb to a real rclone local remote, and
// then one member added, the owner's display name changed, and the
// library root emptied.
func backedUpLeaf(t *testing.T) restoreFixture {
	t.Helper()
	bin, err := exec.LookPath("rclone")
	if err != nil {
		t.Fatalf("rclone is a prerequisite of this test: %v", err)
	}
	conf := filepath.Join(t.TempDir(), "rclone.conf")
	if err := os.WriteFile(conf, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RCLONE_CONFIG", conf)
	t.Setenv("RCLONE_CONFIG_RESTORETEST_TYPE", "local")

	root := t.TempDir()
	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_LIBRARY_ROOT", root)
	environ = append(environ,
		"DUBPLATE_BACKUP_ENABLED=true",
		"DUBPLATE_BACKUP_REMOTE=restoretest:"+t.TempDir(),
	)
	fix := restoreFixture{
		environ:   environ,
		stateDir:  environValue(environ, "DUBPLATE_STATE_DIR"),
		root:      root,
		rclone:    bin,
		rcloneEnv: []string{"RCLONE_CONFIG=" + conf, "RCLONE_CONFIG_RESTORETEST_TYPE=local"},
	}

	seedBackupStore(t, fix.stateDir)
	updateStore(t, fix.stateDir, func(ctx context.Context, tx *store.Tx) error {
		user, coll, album := store.NewID(), store.NewID(), store.NewID()
		if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, created_at, updated_at)
VALUES (?, ?, ?, 'owner', 1, 1)`, user, restoreOwner, restoreOwnerName); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer)
VALUES (?, 'main', ?, ?, 'group', 'server')`, coll, []byte("main"), user); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO albums (id, collection_id, created_at, updated_at) VALUES (?, ?, 1, 1)`, album, coll); err != nil {
			return err
		}
		for _, rel := range []string{"a.flac", "b.flac"} {
			if _, err := tx.Exec(ctx, `INSERT INTO tracks (id, album_id, collection_id, rel_path, created_at, updated_at)
VALUES (?, ?, ?, ?, 1, 1)`, store.NewID(), album, coll, []byte(rel)); err != nil {
				return err
			}
		}
		return nil
	})
	for rel, body := range restoreFixtureFiles {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if code, err := backupRunFor(context.Background(), environ, io.Discard); code != 0 {
		t.Fatalf("the fixture's backup exit = %d: %v", code, err)
	}

	updateStore(t, fix.stateDir, func(ctx context.Context, tx *store.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, created_at, updated_at)
VALUES (?, ?, 'Added', 'added', 2, 2)`, store.NewID(), restoreAddedAddress); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE users SET display_name = ? WHERE address = ?`, restoreChangedName, restoreOwner)
		return err
	})
	if added, owner := restoreRows(t, fix.stateDir); added != 1 || owner != restoreChangedName {
		t.Fatalf("the fixture's later writes: %d added rows and owner %q, want 1 and %q", added, owner, restoreChangedName)
	}
	if err := os.RemoveAll(filepath.Join(root, "main")); err != nil {
		t.Fatal(err)
	}
	return fix
}

// updateStore runs fn in one write transaction on stateDir's store.
func updateStore(t *testing.T, stateDir string, fn func(context.Context, *store.Tx) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	h, err := store.OpenWrite(ctx, stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	if err := h.Update(ctx, fn); err != nil {
		t.Fatalf("update the store: %v", err)
	}
}

// restoreRows reads the fixture's two probes from stateDir's store:
// how many rows hold the member added after the backup, and the
// owner's display name.
func restoreRows(t *testing.T, stateDir string) (added int, owner string) {
	t.Helper()
	ctx := context.Background()
	h, err := store.OpenWrite(ctx, stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	if err := h.QueryRow(ctx, "SELECT count(*) FROM users WHERE address = ?", restoreAddedAddress).Scan(&added); err != nil {
		t.Fatalf("count the added member: %v", err)
	}
	if err := h.QueryRow(ctx, "SELECT display_name FROM users WHERE address = ?", restoreOwner).Scan(&owner); err != nil {
		t.Fatalf("read the owner: %v", err)
	}
	return added, owner
}

// assertLibrary checks that root holds exactly the fixture's files,
// each with its bytes.
func assertLibrary(t *testing.T, root string) {
	t.Helper()
	for rel, body := range restoreFixtureFiles {
		got, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Errorf("library file %s: %v", rel, err)
			continue
		}
		if string(got) != body {
			t.Errorf("library file %s = %q, want %q", rel, got, body)
		}
	}
	n := 0
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			n++
		}
		return err
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if n != len(restoreFixtureFiles) {
		t.Errorf("library root holds %d files, want %d", n, len(restoreFixtureFiles))
	}
}

// assertRestored checks the parts of a finished restore the verb's
// tests can see from outside the store: no record, an empty staging/,
// and one report file at 0600.
func assertRestored(t *testing.T, stateDir string) {
	t.Helper()
	leafDir := filepath.Join(stateDir, "store")
	if _, err := os.Stat(filepath.Join(leafDir, "restore-intent")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("restore-intent record: %v, want none", err)
	}
	if names := readDirNames(t, filepath.Join(leafDir, "staging")); len(names) != 0 {
		t.Errorf("staging/ holds %v, want nothing", names)
	}
	reports, err := filepath.Glob(filepath.Join(leafDir, "restore-report-*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("report files %v, want one", reports)
	}
	info, err := os.Stat(reports[0])
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("report file mode %#o, want 0600", info.Mode().Perm())
	}
}

// recordReason returns what inspect says of stateDir's store, which
// names a restore-intent record's state.
func recordReason(t *testing.T, stateDir string) string {
	t.Helper()
	rep := store.Inspect(context.Background(), store.InspectOptions{StateDir: stateDir, BusyTimeout: testBackupBusy})
	if rep.State != store.StateRestoreInProgress {
		t.Fatalf("inspect state %v (%s), want %v", rep.State, rep.Reason, store.StateRestoreInProgress)
	}
	return rep.Reason
}

// storeFile returns a store file, cleanly closed with no -wal or -shm
// beside it, whose version is below versions under this binary's, its
// ledger cut to match: 0 is a current store, and 1 a store an older
// binary backed up. The write-mode close leaves a -wal, so a raw
// connection opens the file and is the last to close it, which
// checkpoints and removes the -wal.
func storeFile(t *testing.T, below int) string {
	t.Helper()
	dir := t.TempDir()
	seedBackupStore(t, dir)
	path := filepath.Join(dir, "store", "db", "dubplate.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	var v int
	if err := db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		t.Fatal(err)
	}
	if below > 0 {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM migrations WHERE version > %d; PRAGMA user_version = %d", v-below, v-below)); err != nil {
			t.Fatalf("lower %s to version %d: %v", path, v-below, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, side := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(path + side); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("%s%s after the close: %v, want none", path, side, err)
		}
	}
	return path
}

// countLines returns how many of lines read exactly line.
func countLines(lines []string, line string) int {
	n := 0
	for _, l := range lines {
		if l == line {
			n++
		}
	}
	return n
}

// firstOr returns lines' first line, or "" when there is none.
func firstOr(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}
