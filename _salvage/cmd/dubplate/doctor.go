package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/checks"
	"github.com/glw907/dubplate/internal/config"
)

// rclonePath is the rclone binary doctor's R2 probe runs; deploy.sh puts
// rclone on PATH under this name, same as internal/backup and
// internal/restore.
const rclonePath = "rclone"

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Print every config field's resolution state and probe SMTP, R2, Navidrome, and Access read-only",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext(cmd)
			defer cancel()
			return runDoctor(ctx, cmd.OutOrStdout())
		},
	}
}

var errDoctorFailed = errors.New("doctor: one or more checks failed")

// runDoctor loads config (regardless of whether it resolves -- an
// incomplete config is exactly what doctor exists to report), builds the
// report, and prints it. It returns [errDoctorFailed] on any miss so
// main() exits nonzero.
func runDoctor(ctx context.Context, w io.Writer) error {
	cfg, cfgErr := config.Load()
	report := checks.RunDoctor(ctx, cfg, cfgErr, rclonePath, nil)
	if err := printDoctorReport(w, report); err != nil {
		return fmt.Errorf("print doctor report: %w", err)
	}
	if report.Failed() {
		return errDoctorFailed
	}
	return nil
}

// printDoctorReport writes report in restore's report style: the full
// text is built in memory first, so only the single final write to w can
// fail.
func printDoctorReport(w io.Writer, r checks.DoctorReport) error {
	var b strings.Builder

	b.WriteString("Config fields:\n")
	for _, f := range r.Fields {
		status := "resolved"
		if f.Missing {
			status = "MISSING"
		}
		fmt.Fprintf(&b, "  %-35s %s\n", f.Name, status)
	}

	if !r.Probed {
		fmt.Fprintf(&b, "\nProbes skipped: config did not resolve:\n%v\n", r.ConfigErr)
		_, err := io.WriteString(w, b.String())
		return err
	}

	b.WriteString("\nProbes (read-only; nothing is written or sent):\n")
	writeProbe(&b, "SMTP (connect, EHLO, QUIT -- no send)", r.SMTPErr)
	writeProbe(&b, "R2 (rclone lsjson --max-depth 1)", r.R2Err)
	writeProbe(&b, "Navidrome (GET /rest/ping)", r.NavidromeErr)
	writeProbe(&b, "Access (certs endpoint + AUD shape)", r.AccessErr)
	writeProbe(&b, "UploadsDir (create + remove a probe file)", r.UploadsWritableErr)
	writeProbe(&b, "Ledger (open for append)", r.LedgerAppendErr)
	writeProbe(&b, "PublicNavidromeURL (syntax only)", r.PublicNavidromeURLErr)
	writeProbe(&b, "UploadsDir/InboxDir (same filesystem)", r.SameFilesystemErr)

	b.WriteString("\nPing URLs (syntax only -- never fetched: a GET would record a\n")
	b.WriteString("false success ping for a check that never actually ran):\n")
	if len(r.PingURLErrs) == 0 {
		b.WriteString("  all configured ping URLs are well-formed\n")
	} else {
		for _, slug := range slices.Sorted(maps.Keys(r.PingURLErrs)) {
			fmt.Fprintf(&b, "  %-15s MALFORMED: %v\n", slug, r.PingURLErrs[slug])
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func writeProbe(b *strings.Builder, label string, err error) {
	if err != nil {
		fmt.Fprintf(b, "  %-38s FAILED: %v\n", label, err)
		return
	}
	fmt.Fprintf(b, "  %-38s ok\n", label)
}
