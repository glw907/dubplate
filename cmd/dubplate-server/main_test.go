package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/glw907/dubplate/internal/config"
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

// TestNewRootCmd_RegistersShadowThresholdsOnce proves R12:
// newRootCmd runs more than once in a test binary (main_test.go's own
// TestRootCmd_UnknownSubcommand above, plus doctor_test.go's direct
// registerShadowThresholds calls), and config.Register appends without
// dedup, so a bare call from newRootCmd would grow the registry every
// construction.
func TestNewRootCmd_RegistersShadowThresholdsOnce(t *testing.T) {
	newRootCmd()
	first := len(config.Thresholds())

	newRootCmd()
	second := len(config.Thresholds())

	if second != first {
		t.Errorf("Thresholds() len after a second newRootCmd() = %d, want %d (unchanged)", second, first)
	}
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

	want := fmt.Sprintf("%-32s %s\n%-32s %s\n", "version", "v1.2.3", "matcher-commit", "abc123")
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}
