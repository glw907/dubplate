package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/stack"
)

func TestOracleUpNamesMissingPodman(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	cmd := newOracleUpCmd()
	cmd.SetArgs(nil)
	err := cmd.ExecuteContext(t.Context())
	if !errors.Is(err, stack.ErrNoPodman) {
		t.Fatalf("error = %v, want ErrNoPodman", err)
	}
	if !strings.Contains(err.Error(), "podman") {
		t.Errorf("error %q does not name podman", err)
	}
}

func TestOracleUpRegistered(t *testing.T) {
	cmd, _, err := newRootCmd().Find([]string{"oracle-up"})
	if err != nil || cmd.Name() != "oracle-up" {
		t.Fatalf("Find(oracle-up) = %v, %v", cmd, err)
	}
}
