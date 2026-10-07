package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/dubplate/internal/harness/synth"
	"github.com/spf13/cobra"
)

// newSynthCmd builds the synth subcommand: it reads a compact release
// spec and emits a ready-to-derive case directory, the input side of
// the capture-derive loop for every synthetic structural-stratum case
// (corpus/README.md, "The synth tool"). --out names both the case's
// own id (its base name) and the directory synth writes into, the
// same convention `harness capture --out` already uses.
func newSynthCmd() *cobra.Command {
	var specPath, out string

	cmd := &cobra.Command{
		Use:           "synth",
		Short:         "Generate a synthetic case's case.json and release response from a release spec",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if specPath == "" {
				return errors.New("--spec is required")
			}
			if out == "" {
				return errors.New("--out is required")
			}
			if err := runSynth(specPath, out); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.ErrOrStderr(), "synthesized case at %s\n", out)
			return err
		},
	}

	cmd.Flags().StringVar(&specPath, "spec", "", "release spec file (JSON) to build the case from")
	cmd.Flags().StringVar(&out, "out", "", "case directory to create")
	return cmd
}

// runSynth reads specPath, builds the case it describes, and writes
// case.json plus responses/001.json under outDir. It refuses to
// overwrite an existing case.json: `harness synth` seeds a case once,
// and every edit after that runs through the freeze verb instead,
// sharing `harness capture --out`'s guard.
func runSynth(specPath, outDir string) error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}

	var spec synth.ReleaseSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}

	if err := refuseExistingCase(outDir); err != nil {
		return fmt.Errorf("synth: %w", err)
	}

	c, doc, err := synth.Build(filepath.Base(outDir), &spec)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(outDir, "responses"), 0o755); err != nil {
		return fmt.Errorf("create case dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "responses", "001.json"), doc, 0o644); err != nil {
		return fmt.Errorf("write release response: %w", err)
	}
	if err := c.Write(outDir); err != nil {
		return fmt.Errorf("write case: %w", err)
	}
	return nil
}
