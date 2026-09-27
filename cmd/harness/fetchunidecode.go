package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
	"github.com/spf13/cobra"
)

// fetchUnidecodeTimeout bounds one wheel download; Unidecode 1.4.0's
// wheel is 235,837 bytes, so this is generous.
const fetchUnidecodeTimeout = 30 * time.Second

// newFetchUnidecodeCmd builds the fetch-unidecode subcommand: it fills
// the workstation's cached copy of Unidecode's pinned wheel, the file
// package unidecode's tests and dubplate-server's install verb both
// read rather than downloading on every run. It is a no-op, making no
// request, when the cached file already verifies.
func newFetchUnidecodeCmd(client *http.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "fetch-unidecode",
		Short:         "Fill the workstation's cached copy of Unidecode's pinned wheel",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFetchUnidecode(cmd.Context(), client, cmd.ErrOrStderr())
		},
	}
	return cmd
}

func runFetchUnidecode(ctx context.Context, client *http.Client, stderr io.Writer) error {
	path, err := unidecode.WheelCachePath()
	if err != nil {
		return err
	}

	if cached, err := os.ReadFile(path); err == nil && wheelChecksumMatches(cached) {
		_, err := fmt.Fprintf(stderr, "cached wheel at %s already verifies; no request made\n", path)
		return err
	}

	wheel, err := unidecode.Fetch(ctx, client)
	if err != nil {
		return fmt.Errorf("fetch unidecode wheel: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	if err := atomicfile.WriteFile(path, wheel, 0o644); err != nil {
		return fmt.Errorf("write cached wheel: %w", err)
	}
	_, err = fmt.Fprintf(stderr, "fetched and verified wheel to %s\n", path)
	return err
}

func wheelChecksumMatches(data []byte) bool {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]) == unidecode.WheelSHA256
}
