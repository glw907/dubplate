package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/spf13/cobra"
)

// rootFor returns the root command over environ and reg, writing to
// stdout and stderr, with args as its command line.
func rootFor(environ []string, reg *config.Registry, stdout, stderr io.Writer, args ...string) *cobra.Command {
	root := newRootCmd(environ, reg)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	return root
}

// TestRootCmd_UnknownSubcommand proves the fix for the silent-failure
// defect: with SilenceUsage and SilenceErrors set, cobra itself
// writes nothing to either stream on an unrecognized subcommand,
// leaving main's own stderr print (see main.go) as the single place
// the error reaches the operator.
func TestRootCmd_UnknownSubcommand(t *testing.T) {
	cmd := newRootCmd(nil, productionRegistry())
	cmd.SetArgs([]string{"bogus"})

	var stdout, stderr bytes.Buffer
	// Cobra writes a usage block through OutOrStderr, which falls back to
	// os.Stderr when only SetErr is wired, so SetOut must be wired too or
	// a leaked usage block escapes both buffers.
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty; cobra must stay silent so main.go's own print is the only one", stderr.String())
	}
}

// TestNewProductionRegistry_ConstructsIndependentRegistries proves each
// call builds its own registry rather than sharing one. A registry
// panics on a duplicate key, so a shared one would panic here on the
// second construction.
func TestNewProductionRegistry_ConstructsIndependentRegistries(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("first newProductionRegistry() panicked: %v", r)
			}
		}()
		newProductionRegistry()
	}()
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("second newProductionRegistry() panicked: %v", r)
			}
		}()
		newProductionRegistry()
	}()
}

// TestNewVersionCmd_PrintsTwoLabeledRows proves R3: the version verb
// prints exactly two rows, "version" then "matcher-commit", in
// doctor.go's own "%-32s %s" column shape, and that the assertion
// reads the package variables' values rather than their labels.
func TestNewVersionCmd_PrintsTwoLabeledRows(t *testing.T) {
	origVersion, origMatcherCommit := version, matcherCommit
	defer func() { version, matcherCommit = origVersion, origMatcherCommit }()
	version, matcherCommit = "v1.2.3", "abc123"

	cmd := newVersionCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// want is a literal, not a Sprintf built from the same "%-32s %s"
	// format printRow uses: a shared format string would let both sides
	// drift together if the column width or separator regressed.
	want := "version                          v1.2.3\n" +
		"matcher-commit                   abc123\n"
	got := stdout.String()
	if got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}

	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("stdout has %d lines, want 2: %q", len(lines), got)
	}
	if !strings.HasPrefix(lines[1], "matcher-commit") {
		t.Errorf("second line = %q, want it to start with %q", lines[1], "matcher-commit")
	}
}

// TestExecuteWithSignalContext_GivesTheCommandACancellableContext
// proves the production wiring hands the root command a context a
// signal can cancel. cobra's own Execute passes context.Background,
// whose Done channel is nil, so serve.Run's drain select would block on
// a channel that never closes and the graceful shutdown the serve verb
// carries would be unreachable in the binary. The assertion is on
// Done() rather than on an actual signal: delivering SIGTERM to the
// test process would take the whole run down should the wiring be the
// very thing that regressed.
func TestExecuteWithSignalContext_GivesTheCommandACancellableContext(t *testing.T) {
	var got context.Context
	cmd := &cobra.Command{
		Use:           "probe",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			got = cmd.Context()
			return nil
		},
	}
	cmd.SetArgs(nil)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := executeWithSignalContext(cmd); err != nil {
		t.Fatalf("executeWithSignalContext: %v", err)
	}
	if got == nil {
		t.Fatal("RunE was never reached")
	}
	if got.Done() == nil {
		t.Error("cmd.Context().Done() = nil, want a cancellable context: a nil Done channel makes serve's graceful drain unreachable")
	}
	if got == context.Background() {
		t.Error("cmd.Context() is context.Background(), want the signal-derived context")
	}
}

// TestNewRootCmd_RegistersRestore proves the root command carries the
// restore verb.
func TestNewRootCmd_RegistersRestore(t *testing.T) {
	found, _, err := newRootCmd(nil, productionRegistry()).Find([]string{"restore"})
	if err != nil {
		t.Fatalf("Find(restore): %v", err)
	}
	if found.Name() != "restore" {
		t.Errorf("Find(restore) = %q, want the restore verb", found.Name())
	}
}

// TestProductionRegistry_RefusesRetiredShadowKey proves a retired
// DUBPLATE_SHADOW_* key in an otherwise valid environment fails the
// load as an unknown key, so a stale server.env is refused rather
// than silently ignored.
func TestProductionRegistry_RefusesRetiredShadowKey(t *testing.T) {
	environ := authEnviron(t)
	if _, err := config.Load(environ, productionRegistry()); err != nil {
		t.Fatalf("fixture state: config.Load without the retired key: %v", err)
	}

	_, err := config.Load(append(environ, "DUBPLATE_SHADOW_ENABLED=true"), productionRegistry())
	if err == nil {
		t.Fatal("config.Load with DUBPLATE_SHADOW_ENABLED set error = nil, want unknown key")
	}
	if want := "DUBPLATE_SHADOW_ENABLED: unknown configuration key"; !strings.Contains(err.Error(), want) {
		t.Errorf("config.Load error = %q, want it to contain %q", err, want)
	}
}
