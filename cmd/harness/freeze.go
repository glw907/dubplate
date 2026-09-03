package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/oracle"
	"github.com/spf13/cobra"
)

// oracleVenvEnvVar overrides the pinned uv venv freeze derives
// against, matching the variable the oracle-tagged Go tests already
// honor (tools/venv/README.md).
const oracleVenvEnvVar = "DUBPLATE_ORACLE_VENV"

// newFreezeCmd builds the freeze subcommand: it derives the case at
// --case and writes the result back to that directory's case.json,
// the repeatable path a case moves from corpus/staging/<id>/ (or a
// committed corpus/cases/<id>/) to a frozen record.
func newFreezeCmd() *cobra.Command {
	var caseDir string

	cmd := &cobra.Command{
		Use:           "freeze",
		Short:         "Derive a case and write its expected record back to case.json",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if caseDir == "" {
				return errors.New("--case is required")
			}
			venv, err := resolveVenv()
			if err != nil {
				return err
			}
			if err := runFreeze(cmd.Context(), caseDir, oracle.Options{Venv: venv}); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "froze %s\n", caseDir)
			return err
		},
	}

	cmd.Flags().StringVar(&caseDir, "case", "", "case directory to derive and freeze")
	return cmd
}

// runFreeze loads the case at dir, derives it against opts (with Dir
// set to dir, authoritative over the corpus/cases/<ID> default), and
// writes the resulting record back through cassette.Case.Write. It
// refuses to write when the derivation leaves dir/misses.json behind:
// an unanswered request means the case is only half-derived, and
// freezing it would commit a record the case's own request log cannot
// reproduce.
func runFreeze(ctx context.Context, dir string, opts oracle.Options) error {
	c, err := cassette.LoadCase(dir)
	if err != nil {
		return fmt.Errorf("load case: %w", err)
	}

	opts.Dir = dir
	exp, deriveErr := oracle.Derive(ctx, c, opts)

	if _, statErr := os.Stat(filepath.Join(dir, missFileName)); statErr == nil {
		return fmt.Errorf("freeze %s: derivation left %s behind; the case is only half-derived", dir, missFileName)
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("freeze %s: check miss file: %w", dir, statErr)
	}
	if deriveErr != nil {
		return fmt.Errorf("freeze %s: %w", dir, deriveErr)
	}

	c.Expected = exp
	if err := c.Write(dir); err != nil {
		return fmt.Errorf("freeze %s: write record: %w", dir, err)
	}
	return nil
}

// resolveVenv returns the pinned oracle venv path: DUBPLATE_ORACLE_VENV
// when set, or tools/venv/.venv under the repository root otherwise
// (tools/venv/README.md), matching internal/harness/oracle's own
// test-only resolution.
func resolveVenv() (string, error) {
	if v := os.Getenv(oracleVenvEnvVar); v != "" {
		return v, nil
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "tools", "venv", ".venv"), nil
}

// repoRoot walks up from the working directory to the nearest
// ancestor holding go.mod, matching internal/harness/oracle's own
// helper: harness runs from any working directory, so the venv and
// case paths it resolves relative to the repo root cannot assume cwd
// is the module root.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("locate repo root: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("locate repo root: no go.mod above %s", dir)
		}
		dir = parent
	}
}
