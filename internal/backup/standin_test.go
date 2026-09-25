package backup_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/local"
)

// The environment the stand-in reads. The test binary runs as the
// stand-in whenever it is executed under the name rclone.
const (
	// standinRecordEnv names the file the stand-in appends one line to
	// for its argv, one per SIGINT it receives, and one for its exit.
	standinRecordEnv = "DUBPLATE_STANDIN_RECORD"

	// standinPIDEnv names the file the stand-in writes its pid to once
	// its SIGINT handler is installed.
	standinPIDEnv = "DUBPLATE_STANDIN_PID"

	// standinExitEnv sets the exit code per subcommand, as
	// "sync=7,copyto=1"; an unnamed subcommand exits 0.
	standinExitEnv = "DUBPLATE_STANDIN_EXIT"

	// standinSleepEnv names the one subcommand that sleeps for a
	// minute before it exits.
	standinSleepEnv = "DUBPLATE_STANDIN_SLEEP"

	// standinSIGINTEnv is "honor" to exit 130 on SIGINT; any other
	// value ignores it.
	standinSIGINTEnv = "DUBPLATE_STANDIN_SIGINT"

	// standinLockStageEnv, when set, has copyto make its source's
	// directory read-only before it exits, so the stage cannot be
	// removed.
	standinLockStageEnv = "DUBPLATE_STANDIN_LOCK_STAGE"
)

// The environment a re-executed child reads to run one backup through
// a runner over a state directory's store, then sleep until killed.
const (
	childDirEnv    = "DUBPLATE_BACKUP_CHILD_DIR"
	childRemoteEnv = "DUBPLATE_BACKUP_CHILD_REMOTE"
	childRootEnv   = "DUBPLATE_BACKUP_CHILD_ROOT"
)

// testBusy is every test store's busy timeout.
const testBusy = 5 * time.Second

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "rclone" {
		os.Exit(standIn(os.Args[1:]))
	}
	if dir := os.Getenv(childDirEnv); dir != "" {
		os.Exit(runChild(dir))
	}
	os.Exit(m.Run())
}

// standIn is the stand-in rclone's whole life.
func standIn(args []string) int {
	sigs := make(chan os.Signal, 16)
	signal.Notify(sigs, os.Interrupt)
	record := os.Getenv(standinRecordEnv)
	argv, _ := json.Marshal(args)
	appendLine(record, "argv "+string(argv))
	pidFile := os.Getenv(standinPIDEnv)
	if err := os.WriteFile(pidFile+".tmp", []byte(strconv.Itoa(os.Getpid())), 0o600); err == nil {
		_ = os.Rename(pidFile+".tmp", pidFile)
	}

	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	code := standinExit(os.Getenv(standinExitEnv), sub)
	exit := func(code int) int {
		if code != 0 {
			fmt.Fprintf(os.Stderr, "stand-in %s\nstand-in %s exits %d\n\n", sub, sub, code)
		}
		appendLine(record, "exit "+strconv.Itoa(code))
		return code
	}
	if sub == "copyto" && os.Getenv(standinLockStageEnv) != "" && len(args) > 1 {
		_ = os.Chmod(filepath.Dir(args[1]), 0o500)
	}
	if os.Getenv(standinSleepEnv) != sub {
		return exit(code)
	}
	wake := time.After(time.Minute)
	for {
		select {
		case <-sigs:
			appendLine(record, "SIGINT")
			if os.Getenv(standinSIGINTEnv) == "honor" {
				return exit(130)
			}
		case <-wake:
			return exit(code)
		}
	}
}

// standinExit reads sub's exit code from spec.
func standinExit(spec, sub string) int {
	for pair := range strings.SplitSeq(spec, ",") {
		name, code, ok := strings.Cut(pair, "=")
		if ok && name == sub {
			n, _ := strconv.Atoi(code)
			return n
		}
	}
	return 0
}

func appendLine(path, line string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line + "\n")
	_ = f.Close()
}

// runChild opens a runner over dir's store, starts a backup, and
// sleeps until it is killed.
func runChild(dir string) int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	h, err := store.OpenWrite(context.Background(), dir, testBusy)
	if err != nil {
		logger.Error("child open", "err", err)
		return 2
	}
	r, err := local.New(context.Background(), local.Config{
		Store: h,
		Backup: local.BackupConfig{
			Enabled:     true,
			Remote:      os.Getenv(childRemoteEnv),
			LibraryRoot: os.Getenv(childRootEnv),
		},
		Logger: logger,
	})
	if err != nil {
		logger.Error("child runner", "err", err)
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := r.StartBackup(ctx, ops.LocalActor("child"), ops.StartBackupRequest{}); err != nil {
		logger.Error("child start", "err", err)
		return 2
	}
	time.Sleep(time.Hour)
	return 0
}

// A standin is the stand-in rclone a test put on PATH, and the files
// it writes.
type standin struct {
	record string
	pid    string
}

// standinConfig sets the stand-in's behavior; see the env constants.
type standinConfig struct {
	exit      string
	sleep     string
	sigint    string
	lockStage bool
}

// useStandin puts the stand-in alone on PATH, configured by cfg, and
// returns the files it writes.
func useStandin(t *testing.T, cfg standinConfig) standin {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("test binary: %v", err)
	}
	bin := t.TempDir()
	if err := os.Symlink(exe, filepath.Join(bin, "rclone")); err != nil {
		t.Fatalf("link the stand-in: %v", err)
	}
	t.Setenv("PATH", bin)
	return configureStandin(t, cfg)
}

// configureStandin sets the stand-in's behavior for its next
// invocations and returns fresh files for them.
func configureStandin(t *testing.T, cfg standinConfig) standin {
	t.Helper()
	files := t.TempDir()
	s := standin{record: filepath.Join(files, "record"), pid: filepath.Join(files, "pid")}
	t.Setenv(standinRecordEnv, s.record)
	t.Setenv(standinPIDEnv, s.pid)
	t.Setenv(standinExitEnv, cfg.exit)
	t.Setenv(standinSleepEnv, cfg.sleep)
	t.Setenv(standinSIGINTEnv, cfg.sigint)
	lock := ""
	if cfg.lockStage {
		lock = "1"
	}
	t.Setenv(standinLockStageEnv, lock)
	return s
}

// invocations returns each argv the stand-in recorded, in order.
func (s standin) invocations(t *testing.T) [][]string {
	t.Helper()
	var argvs [][]string
	for _, line := range s.lines(t) {
		raw, ok := strings.CutPrefix(line, "argv ")
		if !ok {
			continue
		}
		var argv []string
		if err := json.Unmarshal([]byte(raw), &argv); err != nil {
			t.Fatalf("stand-in argv %q: %v", raw, err)
		}
		argvs = append(argvs, argv)
	}
	return argvs
}

// count returns how many recorded lines read exactly line.
func (s standin) count(t *testing.T, line string) int {
	t.Helper()
	n := 0
	for _, l := range s.lines(t) {
		if l == line {
			n++
		}
	}
	return n
}

func (s standin) lines(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(s.record)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read stand-in record: %v", err)
	}
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

// awaitPID waits for the stand-in to write its pid and returns it.
func (s standin) awaitPID(t *testing.T) int {
	t.Helper()
	end := time.Now().Add(20 * time.Second)
	for {
		data, err := os.ReadFile(s.pid)
		if err == nil {
			pid, err := strconv.Atoi(string(data))
			if err != nil {
				t.Fatalf("stand-in pid %q: %v", data, err)
			}
			return pid
		}
		if time.Now().After(end) {
			t.Fatalf("the stand-in wrote no pid in 20s")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// killAtCleanup kills pid at cleanup, whether or not it still runs.
func killAtCleanup(t *testing.T, pid int) {
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
}

// syncBuffer is a bytes.Buffer safe for a job goroutine's logger.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
