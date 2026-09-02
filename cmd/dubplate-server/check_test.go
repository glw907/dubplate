package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// baseEnviron returns a minimal environ with every owned-now key set
// to a valid value, mirroring internal/config's own test fixture, so a
// case can override or add exactly the entries it needs.
func baseEnviron() []string {
	return []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
	}
}

// dropKey returns a copy of environ with any entry named name removed.
func dropKey(environ []string, name string) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		n, _, _ := strings.Cut(kv, "=")
		if n == name {
			continue
		}
		out = append(out, kv)
	}
	return out
}

func TestCheckCmd(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		wantOK  bool
		errHas  []string
	}{
		{
			name:    "full valid wall",
			environ: baseEnviron(),
			wantOK:  true,
		},
		{
			name:    "missing STATE_DIR",
			environ: dropKey(baseEnviron(), "DUBPLATE_STATE_DIR"),
			wantOK:  false,
			errHas:  []string{"DUBPLATE_STATE_DIR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := newCheckCmd(tt.environ, &stdout, &stderr)
			err := cmd.Execute()

			if tt.wantOK {
				if err != nil {
					t.Fatalf("Execute() error = %v, want nil", err)
				}
				if got := stdout.String(); got != "configuration: ok\n" {
					t.Errorf("stdout = %q, want %q", got, "configuration: ok\n")
				}
				if stderr.Len() != 0 {
					t.Errorf("stderr = %q, want empty", stderr.String())
				}
				return
			}

			if err == nil {
				t.Fatal("Execute() error = nil, want error")
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty on failure", stdout.String())
			}
			// check itself must not print the load error: main is the
			// sole stderr writer, so this command's own stderr stays
			// empty and the error surfaces only through the return value.
			if stderr.Len() != 0 {
				t.Errorf("stderr = %q, want empty; check must not print its own error", stderr.String())
			}
			for _, sub := range tt.errHas {
				if !strings.Contains(err.Error(), sub) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), sub)
				}
			}
			if strings.Contains(err.Error(), "Usage:") {
				t.Errorf("error = %q, want no usage block", err.Error())
			}
		})
	}
}

// TestCheckCmd_EmptyEnvironment proves the acceptance criterion pinned
// in the rung 1 plan word for word: an empty environment yields the
// load error alone, with no usage block, printed by main exactly
// once (proved separately by TestRootCmd_UnknownSubcommand in
// main_test.go and the build-and-run probe); check's own stdout and
// stderr stay empty.
func TestCheckCmd_EmptyEnvironment(t *testing.T) {
	_, loadErr := config.Load(nil)
	if loadErr == nil {
		t.Fatal("config.Load(nil) error = nil, want error; fixture assumption broken")
	}

	var stdout, stderr bytes.Buffer
	cmd := newCheckCmd(nil, &stdout, &stderr)
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if err.Error() != loadErr.Error() {
		t.Errorf("Execute() error = %q, want %q", err.Error(), loadErr.Error())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty; check must not print its own error", stderr.String())
	}
}
