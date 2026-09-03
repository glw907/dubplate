//go:build oracle

// ties.go is the Go half of the assignment-tie detector: it sets up the
// same replay stub and rendered profile a derivation runs under, then
// hands derive_ties.py the one candidate the tie question is asked of.
// It carries the oracle tag because it launches Python; the gate-resident
// half of the analysis is internal/harness/analysis's ledger checks.
package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/stub"
)

// tieInput is the envelope derive_ties.py reads: the case's items and
// pinned date, plus the identity of the candidate whose assignment is
// measured.
type tieInput struct {
	CaseID   string                `json:"case_id"`
	EvalDate string                `json:"eval_date"`
	AlbumID  string                `json:"album_id"`
	Source   string                `json:"source"`
	Items    []cassette.ItemFields `json:"items"`
}

// DeriveTies measures whether the optimal track assignment for c's best
// candidate is unique, returning the ledger row for it. The case must
// already carry a derived record: Candidates[0] is the selector, so a
// case that has never been frozen has no candidate to ask about. A
// record carrying no candidate at all (a zero-track release, dropped
// by beets' own _add_candidate) has no tie question to ask either;
// DeriveTies returns a [analysis.TieSkipped] row for it instead of an
// error, and never starts the replay stub or the Python derivation.
func DeriveTies(ctx context.Context, c *cassette.Case, opts Options) (analysis.TieEntry, error) {
	if c.Expected == nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: case carries no derived record", c.ID)
	}
	if len(c.Expected.Candidates) == 0 {
		return analysis.TieEntry{
			CaseID: c.ID,
			Status: analysis.TieSkipped,
			Reason: "record carries no candidate; a zero-track release is dropped by beets' own _add_candidate before the tie question can be asked",
		}, nil
	}
	best := c.Expected.Candidates[0]

	dir := opts.Dir
	if dir == "" {
		var err error
		dir, err = caseDir(c)
		if err != nil {
			return analysis.TieEntry{}, err
		}
	}

	srv, err := stub.New(c, dir)
	if err != nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: start stub: %w", c.ID, err)
	}
	defer func() { _ = srv.Close() }()

	scratch, err := os.MkdirTemp("", "dubplate-ties-*")
	if err != nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: scratch dir: %w", c.ID, err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	profile, err := renderProfile(c.Profile, scratch, filepath.Join(scratch, "library.db"), srv.Addr())
	if err != nil {
		return analysis.TieEntry{}, err
	}
	profilePath := filepath.Join(scratch, "config.yaml")
	if err := os.WriteFile(profilePath, profile, 0o644); err != nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: write rendered profile: %w", c.ID, err)
	}

	input, err := json.Marshal(tieInput{
		CaseID:   c.ID,
		EvalDate: c.EvalDate,
		AlbumID:  best.AlbumID,
		Source:   best.Source,
		Items:    c.Input.Items,
	})
	if err != nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: encode input: %w", c.ID, err)
	}
	inputPath := filepath.Join(scratch, "input.json")
	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: write input: %w", c.ID, err)
	}

	raw, err := runPython(ctx, opts, "derive_ties.py", profilePath, inputPath)
	if err != nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: %w", c.ID, err)
	}

	var entry analysis.TieEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: decode measurement: %w", c.ID, err)
	}
	if entry.CaseID != c.ID {
		return analysis.TieEntry{}, fmt.Errorf("oracle: ties %s: measurement reports case %q", c.ID, entry.CaseID)
	}
	return entry, nil
}

// runPython launches one venv interpreter on the named script in this
// package and returns its stdout verbatim. It is the tie and component
// drivers' counterpart to runShim, which is pinned to derive_oracle.py
// and to the fake-shim override the untagged tests need.
func runPython(ctx context.Context, opts Options, script string, args ...string) ([]byte, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	python := filepath.Join(opts.Venv, "bin", "python3")
	path := filepath.Join(root, "internal", "harness", "oracle", script)

	cmd := exec.CommandContext(ctx, python, append([]string{path}, args...)...)
	cmd.Env = append(os.Environ(), "PYTHONHASHSEED=0")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run %s: %w: %s", script, err, stderr.String())
	}
	return stdout.Bytes(), nil
}
