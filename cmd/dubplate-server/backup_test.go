package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	"github.com/glw907/dubplate/internal/ops"
)

// testBackupBusy is every backup test store's busy timeout.
const testBackupBusy = 5 * time.Second

// The environment a re-executed test binary reads to run as the real
// backup verb, over os.Environ() rather than a constructed slice, so
// the child's environment is exactly what exec.Cmd.Env sets on it.
const backupVerbChildEnv = "BACKUP_TEST_VERB_CHILD"

// The environment the rclone stand-in reads. The test binary runs as
// the stand-in whenever it is invoked under the name "rclone" (a
// symlink to the test binary itself on a test's own PATH).
const (
	// backupStandinRecordEnv names the file the stand-in appends one
	// line to for its argv, and one line "SIGINT" per SIGINT it
	// receives while sleeping.
	backupStandinRecordEnv = "BACKUP_TEST_STANDIN_RECORD"

	// backupStandinPIDEnv names the file the stand-in writes its pid
	// to, the readiness signal a test waits on.
	backupStandinPIDEnv = "BACKUP_TEST_STANDIN_PID"

	// backupStandinExitEnv sets the exit code per subcommand, as
	// "sync=1,copyto=1"; an unnamed subcommand exits 0.
	backupStandinExitEnv = "BACKUP_TEST_STANDIN_EXIT"

	// backupStandinSleepEnv names the one subcommand that sleeps until
	// a SIGINT arrives or 30 seconds pass, whichever comes first.
	backupStandinSleepEnv = "BACKUP_TEST_STANDIN_SLEEP"

	// backupStandinHonorEnv is "1" to exit 130 on the sleeping
	// subcommand's first SIGINT; any other value ignores every SIGINT,
	// so only the transport's WaitDelay kill ends it (R-6bi12).
	backupStandinHonorEnv = "BACKUP_TEST_STANDIN_HONOR"

	// backupStandinStdoutEnv holds the literal bytes the stand-in
	// writes to stdout for the "backend" subcommand, Task 7's features
	// probe.
	backupStandinStdoutEnv = "BACKUP_TEST_STANDIN_STDOUT"

	// backupStandinStderrEnv holds the literal bytes the stand-in
	// writes to stderr for the "backend" subcommand.
	backupStandinStderrEnv = "BACKUP_TEST_STANDIN_STDERR"

	// backupStandinExecEnv names a real rclone that the stand-in runs
	// in its own place, once it has recorded its argv, for every
	// subcommand but the one backupStandinSleepEnv names, so a test can
	// block one step of a real transfer. An exec'd run writes no pid.
	backupStandinExecEnv = "BACKUP_TEST_STANDIN_EXEC"
)

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "rclone" {
		os.Exit(backupStandIn(os.Args[1:]))
	}
	if os.Getenv(backupVerbChildEnv) != "" {
		cmd := newBackupCmd(os.Environ(), newProductionRegistry(), os.Stdout, os.Stderr)
		os.Exit(run(cmd))
	}
	if raw := os.Getenv(restoreVerbChildEnv); raw != "" {
		os.Exit(runRestoreChild(raw))
	}
	os.Exit(m.Run())
}

// backupStandIn is the stand-in rclone's whole life: it records its
// argv and pid, then, on the one subcommand backupStandinSleepEnv
// names, sleeps until a SIGINT arrives or 30 seconds pass, recording
// every SIGINT it receives.
func backupStandIn(args []string) int {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	record := os.Getenv(backupStandinRecordEnv)
	appendBackupStandinLine(record, "argv "+strings.Join(args, " "))
	if real := os.Getenv(backupStandinExecEnv); real != "" && sub != os.Getenv(backupStandinSleepEnv) {
		err := syscall.Exec(real, append([]string{real}, args...), os.Environ())
		fmt.Fprintf(os.Stderr, "stand-in exec %s: %v\n", real, err)
		return 2
	}
	if pidFile := os.Getenv(backupStandinPIDEnv); pidFile != "" {
		_ = os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o600)
	}
	if sub == "backend" {
		_, _ = io.WriteString(os.Stdout, os.Getenv(backupStandinStdoutEnv))
		_, _ = io.WriteString(os.Stderr, os.Getenv(backupStandinStderrEnv))
	}
	code := backupStandinExitCode(os.Getenv(backupStandinExitEnv), sub)
	if os.Getenv(backupStandinSleepEnv) != sub {
		return code
	}

	sigs := make(chan os.Signal, 8)
	signal.Notify(sigs, os.Interrupt)
	honor := os.Getenv(backupStandinHonorEnv) == "1"
	wake := time.After(30 * time.Second)
	for {
		select {
		case <-sigs:
			appendBackupStandinLine(record, "SIGINT")
			if honor {
				return 130
			}
		case <-wake:
			return code
		}
	}
}

// backupStandinExitCode reads sub's exit code from spec, "name=code"
// pairs joined by commas.
func backupStandinExitCode(spec, sub string) int {
	for pair := range strings.SplitSeq(spec, ",") {
		name, code, ok := strings.Cut(pair, "=")
		if ok && name == sub {
			n, _ := strconv.Atoi(code)
			return n
		}
	}
	return 0
}

func appendBackupStandinLine(path, line string) {
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line + "\n")
	_ = f.Close()
}

// TestBackupCmd_LongNamesExitCodes proves newBackupCmd's help text names
// every exit status Ruling 3 assigns this verb, so a drift between the
// codes and the documented table would fail here rather than only
// showing up as a wrong number an operator reads past. Each real-trigger
// test below (Unavailable, Conflict, Succeeded, Failed,
// SignalHandling) proves the code itself; this proves the code is
// documented.
func TestBackupCmd_LongNamesExitCodes(t *testing.T) {
	cmd := newBackupCmd(nil, newProductionRegistry(), io.Discard, io.Discard)
	long := cmd.Long

	tests := []struct {
		name string
		want string
	}{
		{"succeeded", "\n  0    the backup succeeded"},
		{"failed", "\n  1    the backup failed"},
		{"unavailable", "\n  " + strconv.Itoa(exitBackupUnavailable) + "   backup integration is disabled"},
		{"conflict", "\n  " + strconv.Itoa(exitBackupConflict) + "   a backup is already running"},
		{"canceled", "\n  " + strconv.Itoa(exitBackupCanceled) + "  canceled by any means"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(long, tt.want) {
				t.Errorf("Long = %q, want it to contain %q", long, tt.want)
			}
		})
	}
	if !strings.Contains(long, "canceled by any means") {
		t.Errorf("Long = %q, want it to contain %q", long, "canceled by any means")
	}
}

// TestReportOutcome_Table proves the states table's terminal-state half
// directly: every [ops.JobState] reportOutcome switches on, including
// [ops.JobInterrupted], the state unreachable through a live job
// (R-6bi11's third rule: this verb's own runner holds the job's lease
// for its whole life, so recovery never touches its own record).
func TestReportOutcome_Table(t *testing.T) {
	tests := []struct {
		name       string
		job        ops.JobStatus
		wantErr    bool
		wantCode   int // 0 means "a plain error, run's own default of 1"
		wantStdout string
	}{
		{
			name:       "succeeded, with a final phase",
			job:        ops.JobStatus{ID: "abc123", State: ops.JobSucceeded, Phase: "pruned 3 old snapshots"},
			wantStdout: "backup abc123: succeeded: pruned 3 old snapshots\n",
		},
		{
			name:       "succeeded, no final phase",
			job:        ops.JobStatus{ID: "abc123", State: ops.JobSucceeded},
			wantStdout: "backup abc123: succeeded\n",
		},
		{
			name:     "canceled",
			job:      ops.JobStatus{ID: "abc123", State: ops.JobCanceled},
			wantErr:  true,
			wantCode: exitBackupCanceled,
		},
		{
			name:    "failed",
			job:     ops.JobStatus{ID: "abc123", State: ops.JobFailed, FailureCode: "rclone-exit-7", FailureMessage: "max-delete exceeded"},
			wantErr: true,
		},
		{
			name:    "interrupted",
			job:     ops.JobStatus{ID: "abc123", State: ops.JobInterrupted},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			err := reportOutcome(&stdout, tt.job)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("reportOutcome() error = %v, want nil", err)
				}
				if stdout.String() != tt.wantStdout {
					t.Errorf("stdout = %q, want %q", stdout.String(), tt.wantStdout)
				}
				return
			}
			if err == nil {
				t.Fatalf("reportOutcome() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), string(tt.job.ID)) {
				t.Errorf("error %q does not name the job's handle %s", err, tt.job.ID)
			}
			ece, wrapped := errors.AsType[*exitCodeErr](err)
			if tt.wantCode == 0 {
				if wrapped {
					t.Errorf("reportOutcome() wrapped exit code %d, want a plain error (run's own default of 1)", ece.code)
				}
				return
			}
			if !wrapped {
				t.Fatalf("reportOutcome() = %v, want it to wrap an exit code", err)
			}
			if ece.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", ece.code, tt.wantCode)
			}
		})
	}
}

// TestWrapExitCode_Table proves R-6bi10's exit branch on StartBackup's
// two coded refusals, and that every other error is left unwrapped for
// run's own default of 1 (mutation: mapping every error to exit 1
// would still pass a hand test that checks only the default case, so
// this table asserts both coded rows fire their own code too).
func TestWrapExitCode_Table(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int // 0 means "left unwrapped"
	}{
		{name: "unavailable", err: ops.NewUnavailableError("backup integration disabled"), wantCode: exitBackupUnavailable},
		{name: "conflict", err: ops.NewConflictError("backup job abc123 is running"), wantCode: exitBackupConflict},
		{name: "any other error", err: errors.New("boom"), wantCode: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapExitCode(tt.err)
			ece, wrapped := errors.AsType[*exitCodeErr](got)
			if tt.wantCode == 0 {
				if wrapped {
					t.Errorf("wrapExitCode(%v) wrapped exit code %d, want it left unwrapped", tt.err, ece.code)
				}
				if !errors.Is(got, tt.err) {
					t.Errorf("wrapExitCode(%v) = %v, want it to still wrap the original error", tt.err, got)
				}
				return
			}
			if !wrapped {
				t.Fatalf("wrapExitCode(%v) = %v, want it to wrap an exit code", tt.err, got)
			}
			if ece.code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", ece.code, tt.wantCode)
			}
		})
	}
}

// TestBackupCmd_WallErrors pins spec criterion 19: setting each of a
// retired name, the flag without its required remote, and the flag
// with an unparseable value fails the load, each exiting 1 and naming
// the row's own key or message in stderr.
func TestBackupCmd_WallErrors(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func([]string) []string
		wantSubstr string
	}{
		{
			name:       "retired name set",
			mutate:     func(e []string) []string { return append(e, "DUBPLATE_R2_BUCKET=x") },
			wantSubstr: "DUBPLATE_R2_BUCKET: superseded by DUBPLATE_BACKUP_REMOTE",
		},
		{
			name:       "flag set, remote unset",
			mutate:     func(e []string) []string { return append(e, "DUBPLATE_BACKUP_ENABLED=true") },
			wantSubstr: "DUBPLATE_BACKUP_REMOTE",
		},
		{
			name:       "flag unparseable",
			mutate:     func(e []string) []string { return append(e, "DUBPLATE_BACKUP_ENABLED=maybe") },
			wantSubstr: "DUBPLATE_BACKUP_ENABLED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := tt.mutate(serveAuthEnviron(t))
			var stdout bytes.Buffer
			code, err := backupRunFor(context.Background(), environ, &stdout)
			if code != 1 {
				t.Errorf("exit = %d, want 1; error: %v", code, err)
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantSubstr)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

// TestBackupCmd_StoreAbsent pins spec criterion 18: against a state
// directory with no leaf, the verb exits 1 naming "start serve first",
// and the directory's own listing is unchanged (store.OpenWrite never
// creates).
func TestBackupCmd_StoreAbsent(t *testing.T) {
	environ := serveAuthEnviron(t)
	dir := environValue(environ, "DUBPLATE_STATE_DIR")
	before := readDirNames(t, dir)

	var stdout bytes.Buffer
	code, err := backupRunFor(context.Background(), environ, &stdout)
	if code != 1 {
		t.Errorf("exit = %d, want 1; error: %v", code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "start serve first") {
		t.Errorf("error = %v, want it to name \"start serve first\"", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if after := readDirNames(t, dir); !slices.Equal(before, after) {
		t.Errorf("state directory changed: before %v, after %v", before, after)
	}
}

// TestBackupCmd_Unavailable pins the "flag off" row: a current store
// with backup integration off answers unavailable, naming the same
// reason the catalog gives, and exits 69.
func TestBackupCmd_Unavailable(t *testing.T) {
	environ := serveAuthEnviron(t)
	seedBackupStore(t, environValue(environ, "DUBPLATE_STATE_DIR"))

	var stdout bytes.Buffer
	code, err := backupRunFor(context.Background(), environ, &stdout)
	if code != exitBackupUnavailable {
		t.Errorf("exit = %d, want %d; error: %v", code, exitBackupUnavailable, err)
	}
	if err == nil || !strings.Contains(err.Error(), "backup integration disabled") {
		t.Errorf("error = %v, want it to name \"backup integration disabled\"", err)
	}
}

// TestBackupCmd_Succeeded runs a real job over a current store against
// the stand-in's default (every subcommand exits 0 at once): the verb
// prints the job's handle when it starts and "succeeded" when it ends,
// and exits 0.
func TestBackupCmd_Succeeded(t *testing.T) {
	useBackupStandin(t)
	environ := backupJobEnviron(t)

	var stdout bytes.Buffer
	code, err := backupRunFor(context.Background(), environ, &stdout)
	if code != 0 {
		t.Errorf("exit = %d, want 0; error: %v", code, err)
	}
	if !strings.Contains(stdout.String(), "started") {
		t.Errorf("stdout = %q, want it to print the job's handle when it starts", stdout.String())
	}
	if !strings.Contains(stdout.String(), "succeeded") {
		t.Errorf("stdout = %q, want it to print the job's terminal state when it ends", stdout.String())
	}
}

// TestBackupCmd_Failed runs a real job whose sync subcommand exits 7:
// the verb exits 1 naming the job's handle, its failure code, and its
// message.
func TestBackupCmd_Failed(t *testing.T) {
	useBackupStandin(t)
	t.Setenv(backupStandinExitEnv, "sync=7")
	environ := backupJobEnviron(t)

	var stdout bytes.Buffer
	code, err := backupRunFor(context.Background(), environ, &stdout)
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if err == nil || !strings.Contains(err.Error(), "rclone-exit-7") {
		t.Errorf("error = %v, want it to name the failure code rclone-exit-7", err)
	}
}

// TestBackupCmd_Conflict starts a real job whose sync subcommand sleeps
// until canceled, then starts a second while the first still runs: the
// second answers conflict naming the first's handle and exits 75.
func TestBackupCmd_Conflict(t *testing.T) {
	pidFile := useBackupStandin(t)
	t.Setenv(backupStandinSleepEnv, "sync")
	t.Setenv(backupStandinHonorEnv, "1")
	environ := backupJobEnviron(t)

	ctx, cancel := context.WithCancel(context.Background())
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		var stdout bytes.Buffer
		_, _ = backupRunFor(ctx, environ, &stdout)
	}()
	t.Cleanup(func() {
		cancel()
		<-firstDone
	})

	awaitBackupStandinPID(t, pidFile)

	var stdout bytes.Buffer
	code, err := backupRunFor(context.Background(), environ, &stdout)
	if code != exitBackupConflict {
		t.Errorf("exit = %d, want %d; error: %v", code, exitBackupConflict, err)
	}
	if err == nil || !strings.Contains(err.Error(), "running") {
		t.Errorf("error = %v, want it to name the running job", err)
	}
}

// plantCatalogTracks plants a member, a collection, an album, and n
// tracks into stateDir's store, the fixture the guard's end-to-end row
// needs (internal/backup/job_test.go's own plantTracks, its cmd-level
// twin).
func plantCatalogTracks(t *testing.T, stateDir string, n int) {
	t.Helper()
	updateStore(t, stateDir, func(ctx context.Context, tx *store.Tx) error {
		user, coll, album := store.NewID(), store.NewID(), store.NewID()
		if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, created_at, updated_at)
VALUES (?, 'owner@example.org', 'Owner', 'owner', 1, 1)`, user); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer)
VALUES (?, 'main', CAST('main' AS BLOB), ?, 'group', 'server')`, coll, user); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO albums (id, collection_id, created_at, updated_at) VALUES (?, ?, 1, 1)`, album, coll); err != nil {
			return err
		}
		for i := range n {
			if _, err := tx.Exec(ctx, `INSERT INTO tracks (id, album_id, collection_id, rel_path, created_at, updated_at)
VALUES (?, ?, ?, CAST(? AS BLOB), 1, 1)`, store.NewID(), album, coll, fmt.Sprintf("t%d.flac", i)); err != nil {
				return err
			}
		}
		return nil
	})
}

// plantFinishedJob inserts a finished backup job record aged age past
// now and returns its id, the fixture the retention end-to-end row
// needs.
func plantFinishedJob(t *testing.T, stateDir string, age time.Duration) []byte {
	t.Helper()
	id := store.NewID()
	at := time.Now().Add(-age).UnixMicro()
	updateStore(t, stateDir, func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO jobs (id, kind, at, state, owner_instance, actor, updated_at, finished_at)
VALUES (?, 'backup', ?, 'failed', 'planted', 'local:planted', ?, ?)`, id, at, at, at)
		return err
	})
	return id
}

// jobRecordExists reports whether id's job record still exists in
// stateDir's store.
func jobRecordExists(t *testing.T, stateDir string, id []byte) bool {
	t.Helper()
	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	var n int
	if err := h.QueryRow(context.Background(), "SELECT count(*) FROM jobs WHERE id = ?", id).Scan(&n); err != nil {
		t.Fatalf("count job record: %v", err)
	}
	return n > 0
}

// TestBackupCmd_GuardAdminRow is spec criterion 13's admin half,
// end to end: a library tree holding 6 of a 10-track catalog's files
// passes the registry's own default guard fraction (0.5) and aborts
// under an admin row of 0.9, in each case recording no rclone
// invocation when it aborts.
func TestBackupCmd_GuardAdminRow(t *testing.T) {
	tests := []struct {
		name      string
		admin     string
		wantAbort bool
	}{
		{name: "the default 0.5 passes"},
		{name: "an admin row of 0.9 aborts", admin: "0.9", wantAbort: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useBackupStandin(t)
			environ := serveAuthEnviron(t)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			seedBackupStore(t, stateDir)
			plantCatalogTracks(t, stateDir, 10)

			root := t.TempDir()
			for i := range 6 {
				p := filepath.Join(root, fmt.Sprintf("t%d.flac", i))
				if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			environ = replaceKey(environ, "DUBPLATE_LIBRARY_ROOT", root)
			environ = append(environ, "DUBPLATE_BACKUP_ENABLED=true", "DUBPLATE_BACKUP_REMOTE=teststand:"+t.TempDir())

			if tt.admin != "" {
				h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
				if err != nil {
					t.Fatalf("OpenWrite: %v", err)
				}
				seedSettingRow(t, h, "DUBPLATE_BACKUP_GUARD_FRACTION", tt.admin, "admin")
				if err := h.Close(); err != nil {
					t.Fatalf("close: %v", err)
				}
			}

			var stdout bytes.Buffer
			code, err := backupRunFor(context.Background(), environ, &stdout)
			if tt.wantAbort {
				if code != 1 {
					t.Errorf("exit = %d, want 1; error: %v", code, err)
				}
				if err == nil || !strings.Contains(err.Error(), "library-short") {
					t.Errorf("error = %v, want it to name library-short", err)
				}
				if lines := backupStandinLines(t, os.Getenv(backupStandinRecordEnv)); len(lines) != 0 {
					t.Errorf("stand-in recorded %d invocation(s), want 0: %v", len(lines), lines)
				}
			} else if code != 0 {
				t.Errorf("exit = %d, want 0; error: %v", code, err)
			}
		})
	}
}

// TestBackupCmd_RetentionReaderPrune is criterion 13's retention half,
// end to end: an env-seed row of 48h for DUBPLATE_JOB_RETENTION prunes
// a finished job record aged 72h that the registry's own 2160h default
// would have kept.
func TestBackupCmd_RetentionReaderPrune(t *testing.T) {
	useBackupStandin(t)
	environ := backupJobEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")

	old := plantFinishedJob(t, stateDir, 72*time.Hour)

	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	seedSettingRow(t, h, "DUBPLATE_JOB_RETENTION", "48h", "env-seed")
	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	var stdout bytes.Buffer
	code, err := backupRunFor(context.Background(), environ, &stdout)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; error: %v", code, err)
	}
	if jobRecordExists(t, stateDir, old) {
		t.Errorf("job record %x kept, want it pruned under the 48h retention read at start", old)
	}
}

// settingsTableDump is what dumpSettingsTables reads back, in a form
// two runs compare with ==.
type settingsTableDump struct {
	settings, audit, generation string
}

// dumpSettingsTables reads every settings, settings_audit, and
// settings_generation row out of stateDir's store, formatted so two
// dumps compare with ==. It opens and closes its own handle: the
// caller's own handle, if any, must already be closed.
func dumpSettingsTables(t *testing.T, stateDir string) settingsTableDump {
	t.Helper()
	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() {
		if err := h.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	}()

	return settingsTableDump{
		settings:   dumpRows(t, h, `SELECT key, value, source, seed_value, actor, updated_at FROM settings ORDER BY key`),
		audit:      dumpRows(t, h, `SELECT id, at, kind, actor, channel, key, old_value, new_value FROM settings_audit ORDER BY at, key`),
		generation: dumpRows(t, h, `SELECT singleton, generation FROM settings_generation`),
	}
}

// dumpRows renders every row query returns as one line per row, column
// values tab-separated, a NULL rendered "<NULL>".
func dumpRows(t *testing.T, h *store.Handle, query string) string {
	t.Helper()
	rs, err := h.Query(context.Background(), query)
	if err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	defer func() { _ = rs.Close() }()

	cols, err := rs.Columns()
	if err != nil {
		t.Fatalf("columns for %q: %v", query, err)
	}
	var out strings.Builder
	for rs.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rs.Scan(ptrs...); err != nil {
			t.Fatalf("scan %q: %v", query, err)
		}
		for i, v := range values {
			if i > 0 {
				out.WriteByte('\t')
			}
			if v == nil {
				out.WriteString("<NULL>")
				continue
			}
			fmt.Fprintf(&out, "%v", v)
		}
		out.WriteByte('\n')
	}
	if err := rs.Err(); err != nil {
		t.Fatalf("rows.Err %q: %v", query, err)
	}
	return out.String()
}

// TestBackupCmd_SettingsEvaluationNeverApplies is spec criterion 8: a
// store holding an env-seed row of 14 for DUBPLATE_BACKUP_SNAPSHOT_KEEP,
// with the environment at 20, leaves every settings, settings_audit,
// and generation row byte-identical, logs the disagreement naming the
// key, and runs the job with 14. It runs against a real local rclone
// remote, never the stand-in, so the prune it plants can observe the
// keep the job actually ran with: 18 snapshots already on the remote,
// more than 14 and fewer than 20, so a keep of 14 prunes 5 of them
// (13 survivors plus the run's own, 14 total) while a keep of 20 (the
// environment's value, had an apply reached it) would have kept every
// one (19 total). Fixture state: a seeded row, so an apply would
// write.
func TestBackupCmd_SettingsEvaluationNeverApplies(t *testing.T) {
	if _, err := exec.LookPath("rclone"); err != nil {
		t.Fatalf("rclone is a prerequisite of this test: %v", err)
	}
	environ := backupJobEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")

	remoteDir := t.TempDir()
	t.Setenv("RCLONE_CONFIG_SETTINGSTEST_TYPE", "local")
	environ = replaceKey(environ, "DUBPLATE_BACKUP_REMOTE", "settingstest:"+remoteDir)
	environ = append(environ, "RCLONE_CONFIG_SETTINGSTEST_TYPE=local")

	snapshotsDir := filepath.Join(remoteDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	const seeded = 18
	for i := 1; i <= seeded; i++ {
		stamp := store.FormatStamp(now.Add(-time.Duration(i) * time.Hour))
		if err := os.WriteFile(filepath.Join(snapshotsDir, stamp+".db"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	seedSettingRow(t, h, "DUBPLATE_BACKUP_SNAPSHOT_KEEP", "14", "env-seed")
	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	before := dumpSettingsTables(t, stateDir)
	environ = append(environ, "DUBPLATE_BACKUP_SNAPSHOT_KEEP=20")

	var stdout, stderr bytes.Buffer
	if err := runBackup(context.Background(), environ, productionRegistry(), &stdout, &stderr); err != nil {
		t.Fatalf("runBackup: %v", err)
	}
	if !strings.Contains(stderr.String(), "DUBPLATE_BACKUP_SNAPSHOT_KEEP") {
		t.Errorf("stderr = %q, want a warning naming the disagreeing key", stderr.String())
	}

	after := dumpSettingsTables(t, stateDir)
	if before != after {
		t.Errorf("settings tables changed:\nbefore: %+v\nafter:  %+v", before, after)
	}

	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		t.Fatalf("read %s: %v", snapshotsDir, err)
	}
	if len(entries) != 14 {
		t.Errorf("snapshots/ holds %d files, want 14 (13 seeded survivors plus the run's own): the job ran with a keep other than the stored 14", len(entries))
	}
}

// TestBackupCmd_SettingsEvaluationRefusalNeverFails is spec criterion
// 8's other half: backup exits by its job's outcome alone, never by
// the seed rule's own verdict. An admin row for
// DUBPLATE_BACKUP_SNAPSHOT_KEEP disagreeing with the environment
// refuses under the seed rule (Ruling 2's doctor and check exit
// nonzero on this same state), yet backup still returns nil, only
// logging the refusal, and writes no settings row.
func TestBackupCmd_SettingsEvaluationRefusalNeverFails(t *testing.T) {
	useBackupStandin(t)
	environ := backupJobEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")

	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	seedSettingRow(t, h, "DUBPLATE_BACKUP_SNAPSHOT_KEEP", "14", "admin")
	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	before := dumpSettingsTables(t, stateDir)
	environ = append(environ, "DUBPLATE_BACKUP_SNAPSHOT_KEEP=20")

	var stdout, stderr bytes.Buffer
	if err := runBackup(context.Background(), environ, productionRegistry(), &stdout, &stderr); err != nil {
		t.Fatalf("runBackup: %v, want nil: backup exits by its job's outcome alone, never the seed rule's refusal", err)
	}
	if !strings.Contains(stderr.String(), "DUBPLATE_BACKUP_SNAPSHOT_KEEP") {
		t.Errorf("stderr = %q, want a warning naming the refusing key", stderr.String())
	}

	after := dumpSettingsTables(t, stateDir)
	if before != after {
		t.Errorf("settings tables changed:\nbefore: %+v\nafter:  %+v", before, after)
	}
}

// TestBackupCmd_AuthSettingOutOfBoundsNeverFails proves a timer-run
// backup never fails over an auth setting: an environment rate below
// its floor passes the wall load, since its bound is the seed rule's,
// and backup only logs the refusal, writes no settings row, and exits
// by its job's outcome.
func TestBackupCmd_AuthSettingOutOfBoundsNeverFails(t *testing.T) {
	const key = "DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR"
	useBackupStandin(t)
	environ := backupJobEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	before := dumpSettingsTables(t, stateDir)
	environ = append(environ, key+"=0")

	var stdout, stderr bytes.Buffer
	if err := runBackup(context.Background(), environ, productionRegistry(), &stdout, &stderr); err != nil {
		t.Fatalf("runBackup: %v, want nil: backup exits by its job's outcome alone", err)
	}
	if log := stderr.String(); !strings.Contains(log, "setting refused") || !strings.Contains(log, key) {
		t.Errorf("stderr = %q, want a refusal warning naming %s", log, key)
	}
	if after := dumpSettingsTables(t, stateDir); before != after {
		t.Errorf("settings tables changed:\nbefore: %+v\nafter:  %+v", before, after)
	}
}

// backupRunFor runs the backup verb's own logic directly, exactly as
// run(newBackupCmd(...)) does but returning the error rather than
// printing it to the real os.Stderr, the one write run() and cobra's
// SilenceErrors leave for it (main.go): a test that wants the error's
// own text, rather than only run's exit status, calls this instead of
// going through cobra.
func backupRunFor(ctx context.Context, environ []string, stdout io.Writer) (code int, err error) {
	err = runBackup(ctx, environ, productionRegistry(), stdout, io.Discard)
	if err == nil {
		return 0, nil
	}
	if ece, ok := errors.AsType[*exitCodeErr](err); ok {
		return ece.code, err
	}
	return 1, err
}

// TestBackupVerb_SignalHandling pins spec criterion 20: the verb runs
// as a re-executed process, its own process group's leader
// (Setpgid), against the SIGINT-ignoring stand-in. SIGINT, SIGTERM,
// and SIGHUP sent to the verb's pid alone each end the job canceled
// and the verb exits 130, the stand-in gone once WaitDelay's kill has
// run; SIGINT sent to the verb's own process group, as a terminal
// does, ends the same way but the stand-in's record file holds exactly
// one SIGINT line, the runner's own Cancel, never the terminal's group
// signal (R-6bi12's Setpgid on the transport keeps the two groups
// apart).
func TestBackupVerb_SignalHandling(t *testing.T) {
	tests := []struct {
		name    string
		signal  syscall.Signal
		toGroup bool
	}{
		{name: "SIGINT to the verb alone", signal: syscall.SIGINT},
		{name: "SIGTERM to the verb alone", signal: syscall.SIGTERM},
		{name: "SIGHUP to the verb alone", signal: syscall.SIGHUP},
		{name: "SIGINT to the verb's process group", signal: syscall.SIGINT, toGroup: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exe, err := os.Executable()
			if err != nil {
				t.Fatalf("test binary: %v", err)
			}
			standinDir := t.TempDir()
			if err := os.Symlink(exe, filepath.Join(standinDir, "rclone")); err != nil {
				t.Fatalf("link stand-in: %v", err)
			}
			files := t.TempDir()
			record := filepath.Join(files, "record")
			pidFile := filepath.Join(files, "pid")

			environ := verbChildEnviron(t, standinDir)
			environ = append(environ,
				backupStandinRecordEnv+"="+record,
				backupStandinPIDEnv+"="+pidFile,
				backupStandinSleepEnv+"=sync",
				backupVerbChildEnv+"=1",
			)

			cmd := exec.Command(exe, "-test.run=^$")
			cmd.Env = environ
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Start(); err != nil {
				t.Fatalf("start verb child: %v", err)
			}
			// exited carries cmd.Wait()'s one and only result: the
			// goroutine below sends to it exactly once, so whichever of
			// the select below and Cleanup receives first must not also
			// be the other, or the second receive blocks forever on a
			// channel nothing will ever send to again. received guards
			// that: Cleanup only waits on exited when the select below
			// timed out without ever reading it.
			exited := make(chan error, 1)
			go func() { exited <- cmd.Wait() }()
			received := false
			t.Cleanup(func() {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				if !received {
					<-exited
				}
			})

			standinPID := awaitBackupStandinPID(t, pidFile)

			target := cmd.Process.Pid
			if tt.toGroup {
				target = -cmd.Process.Pid
			}
			if err := syscall.Kill(target, tt.signal); err != nil {
				t.Fatalf("signal the verb child: %v", err)
			}

			select {
			case err := <-exited:
				received = true
				exitErr, ok := errors.AsType[*exec.ExitError](err)
				if !ok {
					t.Fatalf("verb child exited %v, want an *exec.ExitError; stderr:\n%s", err, stderr.String())
				}
				if code := exitErr.ExitCode(); code != exitBackupCanceled {
					t.Errorf("verb child exit code = %d, want %d; stderr:\n%s", code, exitBackupCanceled, stderr.String())
				}
			case <-time.After(25 * time.Second):
				t.Fatalf("verb child did not exit; stderr:\n%s", stderr.String())
			}

			waitBackupStandinGone(t, standinPID, 5*time.Second)

			// The cancel operation's own write succeeds (it runs on
			// opsCtx, never the done signal context), so the verb never
			// falls back to canceling the job in-process and never logs
			// a cancel-write error (mutation: canceling on the done
			// context instead would make the write fail at once, masked
			// in the outcome by the in-process fallback, so this is the
			// row's only proof the write itself succeeded).
			if strings.Contains(stderr.String(), "cancel job") {
				t.Errorf("stderr names a cancel-write error, want none: the cancel operation's own write should succeed:\n%s", stderr.String())
			}

			sigintCount := 0
			for _, l := range backupStandinLines(t, record) {
				if l == "SIGINT" {
					sigintCount++
				}
			}
			if tt.toGroup {
				if sigintCount != 1 {
					t.Errorf("stand-in recorded %d SIGINT line(s), want exactly 1: the runner's own Cancel, never the terminal's group signal", sigintCount)
				}
			} else if sigintCount < 1 {
				t.Errorf("stand-in recorded %d SIGINT line(s), want at least 1: the runner's own Cancel", sigintCount)
			}
		})
	}
}

// seedBackupStore creates a current store under dir, with no handle
// left open, so a later store.OpenWrite finds a leaf to open.
func seedBackupStore(t *testing.T, dir string) {
	t.Helper()
	h, err := store.OpenServe(context.Background(), dir, store.ServeOptions{
		BusyTimeout:   testBackupBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close serve handle: %v", err)
	}
}

// backupJobEnviron returns serveAuthEnviron with a current store, the
// backup integration on, and a library root holding one file (so the
// guard's zero-track, non-empty-root row passes and the job reaches
// rclone).
func backupJobEnviron(t *testing.T) []string {
	t.Helper()
	environ := serveAuthEnviron(t)
	seedBackupStore(t, environValue(environ, "DUBPLATE_STATE_DIR"))

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "track.flac"), []byte("not a real flac file"), 0o600); err != nil {
		t.Fatalf("write library fixture: %v", err)
	}

	environ = replaceKey(environ, "DUBPLATE_LIBRARY_ROOT", root)
	return append(environ,
		"DUBPLATE_BACKUP_ENABLED=true",
		"DUBPLATE_BACKUP_REMOTE=teststand:"+t.TempDir(),
	)
}

// verbChildEnviron returns a full environ for a re-executed verb
// child: a fresh current store, a library root holding one file, the
// backup integration on against a remote a fake name satisfies (the
// stand-in never reads it), and standinDir first on PATH so
// exec.LookPath("rclone") in the child resolves the stand-in.
func verbChildEnviron(t *testing.T, standinDir string) []string {
	t.Helper()
	stateDir := t.TempDir()
	seedBackupStore(t, stateDir)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "track.flac"), []byte("not a real flac file"), 0o600); err != nil {
		t.Fatalf("write library fixture: %v", err)
	}

	return []string{
		"PATH=" + standinDir,
		"DUBPLATE_LIBRARY_ROOT=" + root,
		"DUBPLATE_STATE_DIR=" + stateDir,
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
		"DUBPLATE_IDENTITY_SOURCE=magic-link",
		"DUBPLATE_MAGIC_LINK_SECRET=fixture magic link secret, not a real key",
		"DUBPLATE_SESSION_SECRET=fixture session secret, not a real key",
		"DUBPLATE_PUBLIC_URL=https://dubplate.example.test",
		"DUBPLATE_SMTP_HOST=smtp.example.test",
		"DUBPLATE_SMTP_PORT=587",
		"DUBPLATE_SMTP_USER=dubplate",
		"DUBPLATE_SMTP_PASSWORD=fixture smtp password, not a real key",
		"DUBPLATE_SMTP_FROM=dubplate@example.test",
		"DUBPLATE_BACKUP_ENABLED=true",
		"DUBPLATE_BACKUP_REMOTE=teststand:" + t.TempDir(),
	}
}

// useBackupStandin puts the stand-in alone on PATH for the current
// process (so a job this test starts through runBackup or newBackupCmd
// directly, never a subprocess, resolves it), and returns the record
// and pid files it writes to. A subprocess this test starts inherits
// these same environment variables when its own Env is nil, since
// t.Setenv changes the process's real environment (os/exec's own
// default).
func useBackupStandin(t *testing.T) (pid string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("test binary: %v", err)
	}
	bin := t.TempDir()
	if err := os.Symlink(exe, filepath.Join(bin, "rclone")); err != nil {
		t.Fatalf("link stand-in: %v", err)
	}
	t.Setenv("PATH", bin)

	files := t.TempDir()
	record := filepath.Join(files, "record")
	pid = filepath.Join(files, "pid")
	t.Setenv(backupStandinRecordEnv, record)
	t.Setenv(backupStandinPIDEnv, pid)
	return pid
}

// awaitBackupStandinPID waits for the stand-in to write its pid to
// path and returns it: the readiness signal that its rclone
// invocation, and therefore the verb's own await loop, has started.
func awaitBackupStandinPID(t *testing.T, path string) int {
	t.Helper()
	end := time.Now().Add(20 * time.Second)
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, perr := strconv.Atoi(strings.TrimSpace(string(data)))
			if perr != nil {
				t.Fatalf("stand-in pid file %q: %v", data, perr)
			}
			return pid
		}
		if time.Now().After(end) {
			t.Fatalf("the stand-in never wrote its pid")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// waitBackupStandinGone waits up to limit for pid to stop answering
// signal 0, the proof its process is gone.
func waitBackupStandinGone(t *testing.T, pid int, limit time.Duration) {
	t.Helper()
	end := time.Now().Add(limit)
	for {
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		if time.Now().After(end) {
			t.Fatalf("stand-in pid %d still alive after %v", pid, limit)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// backupStandinLines returns every line path holds, or nil when it was
// never written.
func backupStandinLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read stand-in record: %v", err)
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// readDirNames returns dir's entry names in directory order.
func readDirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}

// TestBackupRemoteRowStatus is the golden test for doctor's "backup
// remote" row's name, colon, and credential checks (R-6bi14): every
// case here is pure string and environment logic, run against a
// constructed environ with no rclone on PATH at all, since none of
// these states reaches the features probe.
func TestBackupRemoteRowStatus(t *testing.T) {
	on := func(kv ...string) []string {
		return append([]string{"DUBPLATE_BACKUP_ENABLED=true"}, kv...)
	}

	tests := []struct {
		name       string
		environ    []string
		wantStatus string
		wantFailed bool
	}{
		{
			name:       "flag off: unavailable, not a failure",
			environ:    nil,
			wantStatus: "unavailable: backup integration disabled",
			wantFailed: false,
		},
		{
			name:       "no colon",
			environ:    on("DUBPLATE_BACKUP_REMOTE=relative/path"),
			wantStatus: "invalid remote: want <name>:<path>",
			wantFailed: true,
		},
		{
			name:       "empty name",
			environ:    on("DUBPLATE_BACKUP_REMOTE=:dir"),
			wantStatus: "invalid remote: want <name>:<path>",
			wantFailed: true,
		},
		{
			name:       "a character outside rclone's remote-name alphabet",
			environ:    on("DUBPLATE_BACKUP_REMOTE=my-remote:dir"),
			wantStatus: `invalid remote: name "my-remote" holds "-", and rclone remote names "can only contain letters, digits, or the _ (underscore) character"`,
			wantFailed: true,
		},
		{
			name:       "empty path is accepted, and the check proceeds to _TYPE",
			environ:    on("DUBPLATE_BACKUP_REMOTE=name:"),
			wantStatus: "invalid remote: RCLONE_CONFIG_NAME_TYPE is not set",
			wantFailed: true,
		},
		{
			name:       "_TYPE unset",
			environ:    on("DUBPLATE_BACKUP_REMOTE=name:dir"),
			wantStatus: "invalid remote: RCLONE_CONFIG_NAME_TYPE is not set",
			wantFailed: true,
		},
		{
			name:       "crypt _REMOTE unset",
			environ:    on("DUBPLATE_BACKUP_REMOTE=name:dir", "RCLONE_CONFIG_NAME_TYPE=crypt"),
			wantStatus: "invalid remote: RCLONE_CONFIG_NAME_REMOTE is not set",
			wantFailed: true,
		},
		{
			name: "crypt wrapped remote's _TYPE unset",
			environ: on("DUBPLATE_BACKUP_REMOTE=name:dir",
				"RCLONE_CONFIG_NAME_TYPE=crypt",
				"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub"),
			wantStatus: "invalid remote: wrapped remote wrapped: RCLONE_CONFIG_WRAPPED_TYPE is not set",
			wantFailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PATH", t.TempDir()) // no rclone reachable: these states never look for it
			status, failed := backupRemoteRowStatus(context.Background(), tt.environ)
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// TestBackupRemoteRowStatus_RcloneNotFound proves the "rclone not
// found" row: a valid, fully-credentialed remote whose PATH holds no
// rclone binary at all.
func TestBackupRemoteRowStatus_RcloneNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	environ := []string{
		"DUBPLATE_BACKUP_ENABLED=true",
		"DUBPLATE_BACKUP_REMOTE=name:dir",
		"RCLONE_CONFIG_NAME_TYPE=local",
	}
	status, failed := backupRemoteRowStatus(context.Background(), environ)
	if status != "rclone not found" {
		t.Errorf("status = %q, want %q", status, "rclone not found")
	}
	if !failed {
		t.Error("failed = false, want true")
	}
}

// TestBackupRemoteRowStatus_NoSubprocessWhenFlagOff proves the flag-off
// state runs no subprocess: the stand-in alone on PATH records no
// invocation.
func TestBackupRemoteRowStatus_NoSubprocessWhenFlagOff(t *testing.T) {
	useBackupStandin(t)
	status, failed := backupRemoteRowStatus(context.Background(), nil)
	if status != "unavailable: backup integration disabled" {
		t.Errorf("status = %q", status)
	}
	if failed {
		t.Error("failed = true, want false")
	}
	if lines := backupStandinLines(t, os.Getenv(backupStandinRecordEnv)); len(lines) != 0 {
		t.Errorf("stand-in recorded %d invocation(s), want 0: %v", len(lines), lines)
	}
}

// TestBackupRemoteRowStatus_Features is the golden test for the
// features probe's own states (R-6bi9, R-6bi12's precedence, Ruling
// 2), each run through the rclone stand-in rather than a real backend.
func TestBackupRemoteRowStatus_Features(t *testing.T) {
	t.Run("crypt _REMOTE a plain path skips the wrapped check; the probe decides", func(t *testing.T) {
		useBackupStandin(t)
		t.Setenv(backupStandinStdoutEnv, `{"Features":{"Copy":false,"Move":true}}`)
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=/plain/path",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if status != "ok" {
			t.Errorf("status = %q, want %q (no wrapped remote's _TYPE is set, so the wrapped check must have been skipped)", status, "ok")
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})

	t.Run("crypt _REMOTE a connection string skips the wrapped check", func(t *testing.T) {
		useBackupStandin(t)
		t.Setenv(backupStandinStdoutEnv, `{"Features":{"Copy":false,"Move":true}}`)
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=:local,type=vfat:/mnt",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if status != "ok" {
			t.Errorf("status = %q, want %q", status, "ok")
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})

	t.Run("Copy and Move both false", func(t *testing.T) {
		useBackupStandin(t)
		t.Setenv(backupStandinStdoutEnv, `{"Features":{"Copy":false,"Move":false}}`)
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=http",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if !strings.Contains(status, "lacks server-side move and copy") {
			t.Errorf("status = %q, want it to name the backend and rclone's --backup-dir requirement", status)
		}
		if !strings.Contains(status, "name") {
			t.Errorf("status = %q, want it to name the backend", status)
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})

	t.Run("a nonzero exit before the deadline reports the last stderr line", func(t *testing.T) {
		useBackupStandin(t)
		t.Setenv(backupStandinExitEnv, "backend=1")
		t.Setenv(backupStandinStderrEnv, "2026/09/24 12:00:00 NOTICE: some rclone context\n401 bad_auth_token\n")
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub",
			"RCLONE_CONFIG_WRAPPED_TYPE=local",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if status != "401 bad_auth_token" {
			t.Errorf("status = %q, want the last non-blank stderr line %q", status, "401 bad_auth_token")
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})

	t.Run("a deadline hit reads not checked, whatever the exit", func(t *testing.T) {
		useBackupStandin(t)
		t.Setenv(backupStandinSleepEnv, "backend")
		t.Setenv(backupStandinHonorEnv, "1")
		orig := featuresProbeTimeout
		featuresProbeTimeout = 30 * time.Millisecond
		t.Cleanup(func() { featuresProbeTimeout = orig })
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub",
			"RCLONE_CONFIG_WRAPPED_TYPE=local",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if status != "not checked: rclone backend features timed out" {
			t.Errorf("status = %q, want %q", status, "not checked: rclone backend features timed out")
		}
		if failed {
			t.Error("failed = true, want false: a deadline hit never fails doctor")
		}
	})

	t.Run("a non-crypt type carries ruling 14's warning, never a failure", func(t *testing.T) {
		useBackupStandin(t)
		t.Setenv(backupStandinStdoutEnv, `{"Features":{"Copy":true,"Move":false}}`)
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=s3",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if !strings.Contains(status, backupCryptWarning) {
			t.Errorf("status = %q, want it to carry ruling 14's warning %q", status, backupCryptWarning)
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})

	t.Run("a crypt type carries no warning", func(t *testing.T) {
		useBackupStandin(t)
		t.Setenv(backupStandinStdoutEnv, `{"Features":{"Copy":false,"Move":true}}`)
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub",
			"RCLONE_CONFIG_WRAPPED_TYPE=local",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if status != "ok" {
			t.Errorf("status = %q, want %q", status, "ok")
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})
}

// TestBackupRemoteRowStatus_EndToEnd runs backupRemoteRowStatus against
// the real rclone binary rather than the stand-in, spec criterion 21's
// three end-to-end rows: a valid local remote, an http remote at a
// closed loopback port (both features false), and a crypt remote
// wrapping an undefined remote.
func TestBackupRemoteRowStatus_EndToEnd(t *testing.T) {
	if _, err := exec.LookPath("rclone"); err != nil {
		t.Fatalf("rclone is a prerequisite of this test: %v", err)
	}

	t.Run("a valid local remote", func(t *testing.T) {
		t.Setenv("RCLONE_CONFIG_E2ELOCAL_TYPE", "local")
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=e2elocal:" + t.TempDir(),
			"RCLONE_CONFIG_E2ELOCAL_TYPE=local",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		// local is not crypt, so ruling 14's warning is appended.
		if !strings.HasPrefix(status, "ok; ") {
			t.Errorf("status = %q, want it to start with %q", status, "ok; ")
		}
		if failed {
			t.Errorf("failed = true, want false: %s", status)
		}
	})

	t.Run("an http remote at a closed loopback port reports both features false", func(t *testing.T) {
		t.Setenv("RCLONE_CONFIG_E2EHTTP_TYPE", "http")
		t.Setenv("RCLONE_CONFIG_E2EHTTP_URL", "http://127.0.0.1:9/")
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=e2ehttp:dir",
			"RCLONE_CONFIG_E2EHTTP_TYPE=http",
			"RCLONE_CONFIG_E2EHTTP_URL=http://127.0.0.1:9/",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if !strings.Contains(status, "lacks server-side move and copy") {
			t.Errorf("status = %q, want it to report both features false", status)
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})

	t.Run("a crypt remote whose wrapped remote is undefined", func(t *testing.T) {
		t.Setenv("RCLONE_CONFIG_E2ECRYPT_TYPE", "crypt")
		t.Setenv("RCLONE_CONFIG_E2ECRYPT_REMOTE", "e2eundefined:sub")
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=e2ecrypt:dir",
			"RCLONE_CONFIG_E2ECRYPT_TYPE=crypt",
			"RCLONE_CONFIG_E2ECRYPT_REMOTE=e2eundefined:sub",
		}
		status, failed := backupRemoteRowStatus(context.Background(), environ)
		if status != "invalid remote: wrapped remote e2eundefined: RCLONE_CONFIG_E2EUNDEFINED_TYPE is not set" {
			t.Errorf("status = %q", status)
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})
}
