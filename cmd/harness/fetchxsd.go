package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/harness/subsonic/xmlrule"
	"github.com/spf13/cobra"
)

// fetchXSDTimeout bounds one schema download; the pinned schema is
// 32,775 bytes.
const fetchXSDTimeout = 30 * time.Second

// newFetchXSDCmd builds the fetch-subsonic-xsd subcommand: it fills the
// workstation's cached copy of the pinned Subsonic schema, the file the
// XML tier reads rather than downloading on every run. It is a no-op,
// making no request, when the cached file already verifies.
func newFetchXSDCmd(client *http.Client) *cobra.Command {
	return &cobra.Command{
		Use:           "fetch-subsonic-xsd",
		Short:         "Fill the workstation's cached copy of the pinned Subsonic XSD",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFetchXSD(cmd.Context(), client, cmd.ErrOrStderr())
		},
	}
}

func runFetchXSD(ctx context.Context, client *http.Client, stderr io.Writer) error {
	path, err := xmlrule.CachePath()
	if err != nil {
		return err
	}
	if _, err := xmlrule.Cached(); err == nil {
		_, err := fmt.Fprintf(stderr, "cached schema at %s already verifies; no request made\n", path)
		return err
	}

	data, err := xmlrule.Fetch(ctx, client)
	if err != nil {
		return fmt.Errorf("fetch subsonic xsd: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	if err := atomicfile.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write cached schema: %w", err)
	}
	_, err = fmt.Fprintf(stderr, "fetched and verified schema to %s\n", path)
	return err
}
