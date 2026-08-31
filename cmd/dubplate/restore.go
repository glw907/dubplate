package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/restore"
)

func newRestoreCmd() *cobra.Command {
	var opts restore.Options

	cmd := &cobra.Command{
		Use:   "restore --to <dir>",
		Short: "Pull the library and database snapshots back from R2 into a target tree",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			return runRestore(ctx, cmd.OutOrStdout(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.To, "to", "", "target directory to restore into (required)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "restore into a target directory that already holds files")
	if err := cmd.MarkFlagRequired("to"); err != nil {
		panic(err)
	}

	return cmd
}

// runRestore loads config and wires it into a [restore.Runner]. It is
// the only place the restore command's dependencies are assembled, so
// cmd/ stays wiring, not logic. The report prints on any outcome that
// reaches it -- including a hard failure -- so a caller can see every
// tree and database restore.Run reached before it stopped.
func runRestore(ctx context.Context, out io.Writer, opts restore.Options) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	r := &restore.Runner{
		Cfg:        cfg,
		SqlitePath: "sqlite3",
		RclonePath: "rclone",
		Now:        time.Now,
	}

	report, runErr := r.Run(ctx, opts)
	if report == nil {
		return runErr
	}
	if err := printRestoreReport(out, report); err != nil && runErr == nil {
		return fmt.Errorf("print restore report: %w", err)
	}
	return runErr
}

// printRestoreReport writes report in doctor's plain-text style: every
// tree's restored count and any bucket-only gap, the archive trees found
// but left unpulled, and each restored database's verification result.
// It builds the full text in memory first, so only the single final
// write to w can fail.
func printRestoreReport(w io.Writer, report *restore.Report) error {
	var b strings.Builder

	fmt.Fprintf(&b, "restore to %s\n", report.To)

	for _, tr := range report.Trees {
		fmt.Fprintf(&b, "  %s: %d file(s) restored\n", tr.Name, tr.RestoredFiles)
		for _, p := range tr.BucketOnly {
			fmt.Fprintf(&b, "    bucket-only, not restored: %s\n", p)
		}
	}

	if len(report.Archives) == 0 {
		b.WriteString("  _archive: none found\n")
	} else {
		fmt.Fprintf(&b, "  _archive: %d dated tree(s) in the bucket, not pulled\n", len(report.Archives))
		for _, a := range report.Archives {
			fmt.Fprintf(&b, "    %s/%s\n", a.Tree, a.Date)
		}
	}

	for _, db := range report.Databases {
		switch {
		case !db.Present:
			fmt.Fprintf(&b, "  %s: not present, skipped\n", db.Path)
		case db.OK:
			fmt.Fprintf(&b, "  %s: integrity check ok\n", db.Path)
		default:
			fmt.Fprintf(&b, "  %s: integrity check FAILED: %s\n", db.Path, db.Detail)
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}
