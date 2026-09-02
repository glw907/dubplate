package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/ping"
)

func newPingFailCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping-fail <slug>",
		Short: "Report a failed run to slug's ping URL, from systemd's OnFailure= chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext(cmd)
			defer cancel()
			return runPingFail(ctx, args[0])
		},
	}
}

// runPingFail resolves DUBPLATE_PING_<SLUG>_URL directly from the
// process environment, deliberately bypassing config.Load: this is the
// OnFailure= chain's entry point, so an unrelated required-field miss
// elsewhere in config must never silence a failure ping for a slug whose
// own URL is perfectly fine.
func runPingFail(ctx context.Context, slug string) error {
	name := pingURLEnvName(slug)
	pingURL, ok := os.LookupEnv(name)
	if !ok || pingURL == "" {
		// Monitoring must never take down what it monitors: an adopter
		// running with no pings provider configured is a valid setup, not
		// an error.
		slog.Info("ping-fail: no ping URL configured; no-op", "slug", slug, "var", name)
		return nil
	}
	parsed, err := url.Parse(pingURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("ping fail %s: malformed ping URL %q", slug, pingURL)
	}

	// A failed delivery (provider outage, DNS, a non-2xx after retries
	// exhaust) is logged and swallowed, never surfaced as an error: this
	// command runs from OnFailure=, and a monitoring hiccup must not put
	// the unit that reports the original failure into a failed state
	// itself.
	if err := ping.NewClient().Fail(ctx, pingURL); err != nil {
		slog.Warn("ping-fail: delivery failed", "slug", slug, "error", err)
	}
	return nil
}

// pingURLEnvName builds the env var name a slug resolves through:
// uppercased, with "-" turned into "_" to make a legal name (an
// OnFailure= instance name may carry a dash; an env var name cannot).
func pingURLEnvName(slug string) string {
	return "DUBPLATE_PING_" + strings.ToUpper(strings.ReplaceAll(slug, "-", "_")) + "_URL"
}
