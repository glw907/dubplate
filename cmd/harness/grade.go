package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/harness/subsonic/grade"
	"github.com/spf13/cobra"
)

// errUnledgered marks a grade whose report holds a finding no ledger row
// covers.
var errUnledgered = errors.New("unledgered findings")

// newGradeCmd builds the grade subcommand: it grades one recorded session
// between two of its sources and writes the report.
func newGradeCmd() *cobra.Command {
	var sessionDir, pair, out, ledger string

	cmd := &cobra.Command{
		Use:           "grade",
		Short:         "Grade a recorded session between two of its sources",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sources := strings.Split(pair, ",")
			if sessionDir == "" || out == "" || len(sources) != 2 {
				return errors.New("grade needs --session, --out, and --pair naming two sources")
			}
			root, err := repoRoot()
			if err != nil {
				return err
			}
			data := filepath.Join(root, "internal", "harness", "subsonic", "data")
			if ledger == "" {
				ledger = filepath.Join(data, "ledger.json")
			}
			cfg := grade.Config{
				Ledger:   ledger,
				Sessions: filepath.Dir(filepath.Clean(sessionDir)),
				OpenAPI:  filepath.Join(data, "openapi"),
				Manifest: filepath.Join(data, "fixture", "manifest.json"),
			}
			g, err := grade.New(cfg)
			if err != nil {
				return err
			}
			return runGrade(cmd.Context(), g, sessionDir, [2]string{sources[0], sources[1]}, out, cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&sessionDir, "session", "", "session directory to grade; the ledger's sessions are its siblings")
	cmd.Flags().StringVar(&pair, "pair", "", "the two sources to compare, comma-separated")
	cmd.Flags().StringVar(&out, "out", "", "file to write the report to")
	cmd.Flags().StringVar(&ledger, "ledger", "", "divergence ledger (default: the committed one)")
	return cmd
}

// runGrade grades the session and writes its report to out. It fails on
// a refusal, a stale ledger row, or any unledgered finding, writing the
// report first whenever there is one.
func runGrade(ctx context.Context, g *grade.Grader, sessionDir string, pair [2]string, out string, status io.Writer) error {
	r, gradeErr := g.Grade(ctx, sessionDir, pair)
	if r == nil {
		return gradeErr
	}
	if err := atomicfile.WriteFile(out, r.Format(), 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	unledgered := r.Unledgered()
	if _, err := fmt.Fprintf(status, "grade: %d findings, %d unledgered; report at %s\n", len(r.Findings), unledgered, out); err != nil {
		return err
	}
	if unledgered > 0 {
		gradeErr = errors.Join(gradeErr, fmt.Errorf("%w: %d", errUnledgered, unledgered))
	}
	return gradeErr
}
