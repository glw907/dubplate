package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// waitDelay bounds how long a canceled rclone has to exit after its
// interrupt before it is killed.
const waitDelay = 10 * time.Second

// maxMessage bounds a failure message read from rclone's stderr.
const maxMessage = 512

// A failure is a run's failure as its job record carries it: a stable
// code and a message for a person.
type failure struct {
	code    string
	message string
}

func fail(code, format string, args ...any) *failure {
	return &failure{code: code, message: fmt.Sprintf(format, args...)}
}

func (f *failure) Error() string { return f.message }

// FailureCode returns the stable code a job record carries for f.
func (f *failure) FailureCode() string { return f.code }

// A transport runs rclone as a child that inherits the caller's locks.
type transport struct {
	bin string

	// inherit returns the descriptors the child inherits, from 3 on,
	// which run closes once the child exits: a backup's kind lease and
	// store lock, or a restore's exclusive store lock. A nil inherit
	// passes none.
	inherit func() ([]*os.File, error)
}

// run runs rclone with args, of which the first is the subcommand,
// writing its stdout to stdout, which may be nil to discard it. The
// child runs in its own process group, holds the descriptors
// t.inherit returns, and on ctx's cancel receives one SIGINT, then a
// kill once [waitDelay] has passed.
func (t transport) run(ctx context.Context, stdout io.Writer, args ...string) (err error) {
	var files []*os.File
	if t.inherit != nil {
		files, err = t.inherit()
		if err != nil {
			return fmt.Errorf("rclone %s: %w", args[0], err)
		}
	}
	defer func() {
		for _, f := range files {
			if cerr := f.Close(); cerr != nil {
				err = errors.Join(err, fmt.Errorf("close inherited lock: %w", cerr))
			}
		}
	}()

	var stderr lastLine
	cmd := exec.CommandContext(ctx, t.bin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	cmd.ExtraFiles = files
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = waitDelay
	return classify(ctx, args[0], cmd.Run(), stderr.String())
}

// A remoteEntry is one object rclone's lsjson lists. Size is -1 for a
// directory, and for an object whose backend does not report a size.
type remoteEntry struct {
	Name  string `json:"Name"`
	IsDir bool   `json:"IsDir"`
	Size  int64  `json:"Size"`
}

// snapshotStamp reports the stamp e names, when e is a file named
// <stamp>.db with a stamp [store.ParseStamp] accepts. A directory, or
// a name outside that shape, reports false.
func (e remoteEntry) snapshotStamp() (string, bool) {
	if e.IsDir {
		return "", false
	}
	stamp, ok := strings.CutSuffix(e.Name, ".db")
	if !ok {
		return "", false
	}
	if _, err := store.ParseStamp(stamp); err != nil {
		return "", false
	}
	return stamp, true
}

// list lists the entries directly under dest, a remote path, an absent
// directory read as empty: rclone v1.74.3's lsjson exits 3 with
// "directory not found" on one.
func (t transport) list(ctx context.Context, dest string, dirsOnly bool) ([]remoteEntry, error) {
	args := []string{"lsjson"}
	if dirsOnly {
		args = append(args, "--dirs-only")
	}
	args = append(args, dest)

	var stdout bytes.Buffer
	if err := t.run(ctx, &stdout, args...); err != nil {
		if f, ok := errors.AsType[*failure](err); ok && f.code == "rclone-exit-3" {
			return nil, nil
		}
		return nil, fmt.Errorf("list %s: %w", dest, err)
	}

	var entries []remoteEntry
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		return nil, fmt.Errorf("parse listing of %s: %w", dest, err)
	}
	return entries, nil
}

// classify decides how an rclone run that returned err ended. Any
// error once ctx is done is a cancel, whatever the exit status: Wait
// reports a child's own exit status after a cancel, and rclone exits
// 130 on the interrupt. A death by a signal nobody sent through the
// cancel is a failure naming the signal, and any other nonzero exit a
// failure naming the status. The failure's message is rclone's last
// line on stderr.
func classify(ctx context.Context, sub string, err error, line string) error {
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("rclone %s: %w", sub, ctx.Err())
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return fmt.Errorf("run rclone %s: %w", sub, err)
	}
	if line == "" {
		line = "rclone " + sub + ": " + ee.Error()
	}
	if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return &failure{code: "rclone-signal-" + signalName(ws.Signal()), message: line}
	}
	return &failure{code: "rclone-exit-" + strconv.Itoa(ee.ExitCode()), message: line}
}

// signalNames spells the signals a process is likely to die by as
// kill(1) does, without the SIG prefix.
var signalNames = map[syscall.Signal]string{
	syscall.SIGABRT: "ABRT",
	syscall.SIGALRM: "ALRM",
	syscall.SIGBUS:  "BUS",
	syscall.SIGFPE:  "FPE",
	syscall.SIGHUP:  "HUP",
	syscall.SIGILL:  "ILL",
	syscall.SIGINT:  "INT",
	syscall.SIGKILL: "KILL",
	syscall.SIGPIPE: "PIPE",
	syscall.SIGQUIT: "QUIT",
	syscall.SIGSEGV: "SEGV",
	syscall.SIGTERM: "TERM",
	syscall.SIGTRAP: "TRAP",
	syscall.SIGUSR1: "USR1",
	syscall.SIGUSR2: "USR2",
	syscall.SIGXCPU: "XCPU",
	syscall.SIGXFSZ: "XFSZ",
}

// signalName returns sig's name, or its number for a signal
// [signalNames] does not spell.
func signalName(sig syscall.Signal) string {
	if name, ok := signalNames[sig]; ok {
		return name
	}
	return strconv.Itoa(int(sig))
}

// A lastLine keeps the last non-blank line written to it, cut to
// [maxMessage] bytes, and nothing else, so a long run's stderr costs
// no memory.
type lastLine struct {
	cur, last []byte
}

func (l *lastLine) Write(p []byte) (int, error) {
	n := len(p)
	for {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			l.add(p)
			return n, nil
		}
		l.add(p[:i])
		if len(bytes.TrimSpace(l.cur)) > 0 {
			l.last = append(l.last[:0], l.cur...)
		}
		l.cur = l.cur[:0]
		p = p[i+1:]
	}
}

func (l *lastLine) add(p []byte) {
	room := maxMessage - len(l.cur)
	l.cur = append(l.cur, p[:min(room, len(p))]...)
}

// String returns the last non-blank line, an unterminated one
// included.
func (l *lastLine) String() string {
	line := l.last
	if len(bytes.TrimSpace(l.cur)) > 0 {
		line = l.cur
	}
	return strings.ToValidUTF8(strings.TrimSpace(string(line)), "")
}
