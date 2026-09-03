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
// the mistagged and partial-rip axes. --facts switches to a read-only
// mode: it prints the case's structural facts (release date,
// release-group first-release-date, disc/track counts, status,
// documented aliases) to stdout and writes nothing, the sanctioned
// confirmation channel for Task 3's fact-verified axes
// (corpus/README.md, "The seed-items tool"; the plan's Global
// Constraints).
func newSeedItemsCmd() *cobra.Command {
	var caseDir, overridesPath string
	var facts bool

	cmd := &cobra.Command{
		Use:           "seed-items",
		Short:         "Project a staged case's Input.Items from its own captured response",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if caseDir == "" {
				return errors.New("--case is required")
			}
			if facts {
				if overridesPath != "" {
					return errors.New("--facts is read-only and does not accept --overrides")
				}
				out, err := runSeedItemsFacts(caseDir)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), out)
				return err
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
	cmd.Flags().BoolVar(&facts, "facts", false, "print the case's structural facts to stdout and write nothing")
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

// runSeedItemsFacts loads the case at dir and returns its structural
// facts, deterministically formatted, read from the same primary
// response runSeedItems projects Input.Items from. It never calls
// cassette.Case.Write: --facts is read-only by design, so it may run
// against a case whose Input.Items are already projected without
// perturbing the case directory at all.
func runSeedItemsFacts(dir string) (string, error) {
	c, err := cassette.LoadCase(dir)
	if err != nil {
		return "", fmt.Errorf("load case: %w", err)
	}
	if len(c.Requests) == 0 {
		return "", fmt.Errorf("seed-items: %s: case carries no requests to read facts from", dir)
	}

	doc, err := os.ReadFile(filepath.Join(dir, c.Requests[0].Response))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	facts, err := synth.ExtractFacts(doc)
	if err != nil {
		return "", fmt.Errorf("extract facts: %w", err)
	}
	return synth.FormatFacts(facts), nil
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
