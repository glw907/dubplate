package backup

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// waitErr runs script under sh and returns what Wait reported.
func waitErr(t *testing.T, script string) error {
	t.Helper()
	err := exec.Command("sh", "-c", script).Run()
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			t.Fatalf("sh -c %q: %v", script, err)
		}
	}
	return err
}

// TestClassify pins the transfer's outcome precedence: once the
// context is done, any error is a cancel whatever the exit status,
// exit 130 included; an exit 0 is a success whatever the context; a
// signal death nobody sent through the cancel names the signal; any
// other nonzero exit names the status.
func TestClassify(t *testing.T) {
	done, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name         string
		ctx          context.Context
		script       string
		wantCanceled bool
		wantCode     string
	}{
		{name: "exit 0", ctx: context.Background(), script: "exit 0"},
		{name: "exit 0 after cancel", ctx: done, script: "exit 0"},
		{name: "exit 130 after cancel", ctx: done, script: "exit 130", wantCanceled: true},
		{name: "exit 7 after cancel", ctx: done, script: "exit 7", wantCanceled: true},
		{name: "killed after cancel", ctx: done, script: "kill -KILL $$", wantCanceled: true},
		{name: "exit 7", ctx: context.Background(), script: "exit 7", wantCode: "rclone-exit-7"},
		{name: "exit 130", ctx: context.Background(), script: "exit 130", wantCode: "rclone-exit-130"},
		{name: "killed", ctx: context.Background(), script: "kill -KILL $$", wantCode: "rclone-signal-KILL"},
		{name: "terminated", ctx: context.Background(), script: "kill -TERM $$", wantCode: "rclone-signal-TERM"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classify(tt.ctx, "sync", waitErr(t, tt.script), "last words")
			if got := errors.Is(err, context.Canceled); got != tt.wantCanceled {
				t.Errorf("classify() = %v, canceled %v, want %v", err, got, tt.wantCanceled)
			}
			if code := codeOf(err); code != tt.wantCode {
				t.Errorf("classify() = %v, code %q, want %q", err, code, tt.wantCode)
			}
			if tt.wantCode != "" && err.Error() != "last words" {
				t.Errorf("failure message = %q, want rclone's last stderr line", err)
			}
		})
	}
}

// TestClassify_NoStderr pins the message of a failure rclone wrote no
// stderr for: the subcommand and the exit status.
func TestClassify_NoStderr(t *testing.T) {
	err := classify(context.Background(), "copyto", waitErr(t, "exit 3"), "")
	if got, want := err.Error(), "rclone copyto: exit status 3"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

// TestLastLine pins the stderr tail: the last non-blank line across
// writes that split lines anywhere, an unterminated last line
// included, cut to 512 bytes.
func TestLastLine(t *testing.T) {
	long := strings.Repeat("x", 600)
	tests := []struct {
		name   string
		writes []string
		want   string
	}{
		{name: "none", want: ""},
		{name: "one line", writes: []string{"a\n"}, want: "a"},
		{name: "last of several", writes: []string{"a\nb\n", "c\n"}, want: "c"},
		{name: "split across writes", writes: []string{"first\nsec", "ond\n"}, want: "second"},
		{name: "unterminated", writes: []string{"a\nb"}, want: "b"},
		{name: "trailing blank lines", writes: []string{"a\n\n  \n"}, want: "a"},
		{name: "long line cut", writes: []string{long + "\n"}, want: long[:512]},
		{name: "long line across writes", writes: []string{long[:300], long[300:], "\n"}, want: long[:512]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l lastLine
			for _, w := range tt.writes {
				if n, err := l.Write([]byte(w)); n != len(w) || err != nil {
					t.Fatalf("Write(%q) = %d, %v", w, n, err)
				}
			}
			if got := l.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
