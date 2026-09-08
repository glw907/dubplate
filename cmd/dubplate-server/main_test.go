package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestRootCmd_UnknownSubcommand proves the fix for the silent-failure
// defect: with SilenceUsage and SilenceErrors set, cobra itself
// writes nothing to either stream on an unrecognized subcommand,
// leaving main's own stderr print (see main.go) as the single place
// the error reaches the operator.
func TestRootCmd_UnknownSubcommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"bogus"})

	var stdout, stderr bytes.Buffer
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

// TestNewRootCmd_ConstructsIndependentRegistries proves R12's
// double-registration hazard is gone under R1's value registry:
// newRootCmd runs more than once in a test binary (this file's own
// TestRootCmd_UnknownSubcommand, plus doctor_test.go's direct
// registerShadowThresholds calls), and each call now builds its own
// fresh *config.Registry rather than sharing one. A registry panics on
// a duplicate key (internal/config/registry.go), so a newRootCmd that
// leaked a shared registry across calls would panic here on the
// second construction; this replaces the retired
// TestNewRootCmd_RegistersShadowThresholdsOnce, whose own proof read
// config.Thresholds(), a package-level function Task 3 retires along
// with registeredThresholds. internal/config/registry_test.go's own
// TestNewRegistry_Independent proves the same property at the
// registry's own level.
func TestNewRootCmd_ConstructsIndependentRegistries(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("first newRootCmd() panicked: %v", r)
			}
		}()
		newRootCmd()
	}()
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("second newRootCmd() panicked: %v", r)
			}
		}()
		newRootCmd()
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
