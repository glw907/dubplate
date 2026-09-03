package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/synth"
	"github.com/spf13/cobra"
)

// newSeedItemsCmd builds the seed-items subcommand: it reads a staged
// case's primary captured response and writes a correct baseline
// Input.Items into case.json, the sanctioned way to author a
// real-catalog case's items (corpus/README.md, "The seed-items tool";
// B1 in the rung 2c plan review). --overrides applies the same
// field/missing/extra vocabulary `harness synth` already carries, for
// the mistagged and partial-rip axes.
func newSeedItemsCmd() *cobra.Command {
	var caseDir, overridesPath string

	cmd := &cobra.Command{
		Use:           "seed-items",
		Short:         "Project a staged case's Input.Items from its own captured response",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if caseDir == "" {
				return errors.New("--case is required")
			}
			n, err := runSeedItems(caseDir, overridesPath)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "projected %d item(s) into %s\n", n, filepath.Join(caseDir, "case.json"))
			return err
		},
	}

	cmd.Flags().StringVar(&caseDir, "case", "", "staged or committed case directory to project items into")
	cmd.Flags().StringVar(&overridesPath, "overrides", "", "JSON file of synth.Override entries to perturb the projected baseline")
	return cmd
}

// runSeedItems loads the case at dir (case.json must already carry a
// valid id, profile, eval date, and request log; freeze.go's own
// LoadCase precondition), projects Input.Items from the response its
// first request names, applies any overrides, and writes the result
// back through cassette.Case.Write. It never sets Profile or EvalDate:
// those are a human's job once a case is seeded, exactly as
// `harness capture`'s own doc comment already states.
func runSeedItems(dir, overridesPath string) (int, error) {
	c, err := cassette.LoadCase(dir)
	if err != nil {
		return 0, fmt.Errorf("load case: %w", err)
	}
	if len(c.Requests) == 0 {
		return 0, fmt.Errorf("seed-items: %s: case carries no requests to project items from", dir)
	}

	doc, err := os.ReadFile(filepath.Join(dir, c.Requests[0].Response))
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}

	overrides, err := loadOverrides(overridesPath)
	if err != nil {
		return 0, err
	}

	items, err := synth.ProjectItems(doc, overrides, false)
	if err != nil {
		return 0, fmt.Errorf("project items: %w", err)
	}

	c.Input.Items = items
	if err := c.Write(dir); err != nil {
		return 0, fmt.Errorf("write case: %w", err)
	}
	return len(items), nil
}

// loadOverrides reads path as a JSON array of synth.Override entries,
// or returns nil unperturbed when path is empty: most real-catalog
// cases (a clean rip) need no overrides at all.
func loadOverrides(path string) ([]synth.Override, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read overrides: %w", err)
	}
	var overrides []synth.Override
	if err := json.Unmarshal(data, &overrides); err != nil {
		return nil, fmt.Errorf("parse overrides: %w", err)
	}
	return overrides, nil
}
