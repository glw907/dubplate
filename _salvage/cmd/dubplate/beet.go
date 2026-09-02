package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/lock"
)

func newBeetCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "beet -- [args...]",
		Short:              "Run beet directly, blocking until any in-progress import releases the lock",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBeet(cmd.Context(), args)
		},
	}
}

// runBeet acquires the import lock, blocking until any in-progress
// dubplate import releases it, then execs the real beet with argv, tty,
// and environment intact. It returns only on failure: on success,
// syscall.Exec replaces this process's image outright, so there is no
// "after" to return to.
func runBeet(ctx context.Context, args []string) error {
	args = trimDashDash(args)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	l, err := lock.AcquireBlocking(ctx, lock.Path(cfg.Paths.StateDir))
	if err != nil {
		return fmt.Errorf("acquire import lock: %w", err)
	}

	beetPath, err := exec.LookPath("beet")
	if err != nil {
		_ = l.Release()
		return fmt.Errorf("find beet on PATH: %w", err)
	}

	// l's lock file descriptor has FD_CLOEXEC cleared (internal/lock), so
	// the lock rides straight through this exec into beet -- no window
	// where an overlapping dubplate import could acquire it in between.
	err = syscall.Exec(beetPath, append([]string{"beet"}, args...), os.Environ())
	_ = l.Release()
	return fmt.Errorf("exec beet: %w", err)
}

// trimDashDash drops a single leading "--" separator, the documented
// `dubplate beet -- <args>` shape. DisableFlagParsing leaves beet's own
// flags unparsed and passes "--" through literally rather than having
// cobra consume it, unlike ordinary flag parsing.
func trimDashDash(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}
	return args
}
