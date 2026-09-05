//go:build oracle

// defect_free.go is the Go half of ruling A / M6's defect_free block:
// it hands derive_defect_free.py the string-heuristic subjects' own
// committed release response and item set in one process, since
// (unlike a rendered-profile derivation) nothing here touches beets'
// config between subjects. It carries the oracle tag because it
// launches Python; the gate-resident half of the analysis is
// internal/harness/analysis's ledger checks.
package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// DefectFreeSubject is one string-heuristic case the defect-free
// derivation runs over. ResponsePath and Items are what the script
// needs to rebuild the candidate and the item set: an album distance
// takes the candidate's whole track list and a solved assignment
// (autotag/distance.py:429-432, autotag/match.py:53), neither of which
// the committed record carries. PatternTable names the substitution the
// figures are derived under, and is one of the two legal values
// DefectFree.PatternTable carries; an empty PatternTable asks the
// script to leave SD_PATTERNS untouched, the both-ways check's
// baseline half.
type DefectFreeSubject struct {
	CaseID       string
	ReleaseMBID  string
	ResponsePath string
	EvalDate     string
	PatternTable string
}

// DefectFreeRow is one subject's derived figures, exactly as
// derive_defect_free.py prints them. Mapping is the re-solved
// assign_items pairing under the substituted table, item index to
// track index, so a caller can verify the mapping itself did not
// change under the substitution rather than assuming it held.
type DefectFreeRow struct {
	CaseID         string         `json:"case_id"`
	Distance       float64        `json:"distance"`
	Recommendation string         `json:"recommendation"`
	Action         string         `json:"action"`
	PatternTable   string         `json:"pattern_table"`
	Mapping        map[string]int `json:"mapping"`
}

// defectFreeSubjectInput is one subject as the script reads it. Items
// come from the case's own committed input rather than from
// [DefectFreeSubject] directly, because the subject is this package's
// caller-facing identity for a case and the item set is the case's own
// data, resolved here through [cassette.LoadCase].
type defectFreeSubjectInput struct {
	CaseID       string                `json:"case_id"`
	ResponsePath string                `json:"response_path"`
	ReleaseMBID  string                `json:"release_mbid"`
	EvalDate     string                `json:"eval_date"`
	PatternTable string                `json:"pattern_table"`
	Items        []cassette.ItemFields `json:"items"`
}

// DeriveDefectFree runs derive_defect_free.py over subjects and returns
// the figures it prints, unparsed: one document carrying a
// [DefectFreeRow] per subject, from one beets process for the whole
// batch, because nothing here renders a profile or otherwise changes
// beets' config between subjects.
func DeriveDefectFree(ctx context.Context, opts Options, subjects []DefectFreeSubject) ([]byte, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}

	inputs := make([]defectFreeSubjectInput, 0, len(subjects))
	for _, s := range subjects {
		c, err := cassette.LoadCase(filepath.Join(root, "corpus", "cases", s.CaseID))
		if err != nil {
			return nil, fmt.Errorf("oracle: defect free %s: load case: %w", s.CaseID, err)
		}
		inputs = append(inputs, defectFreeSubjectInput{
			CaseID:       s.CaseID,
			ResponsePath: filepath.Join(root, "corpus", "cases", s.CaseID, s.ResponsePath),
			ReleaseMBID:  s.ReleaseMBID,
			EvalDate:     s.EvalDate,
			PatternTable: s.PatternTable,
			Items:        c.Input.Items,
		})
	}

	scratch, err := os.MkdirTemp("", "dubplate-defect-free-*")
	if err != nil {
		return nil, fmt.Errorf("oracle: defect free: scratch dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	input, err := json.Marshal(struct {
		Subjects []defectFreeSubjectInput `json:"subjects"`
	}{inputs})
	if err != nil {
		return nil, fmt.Errorf("oracle: defect free: encode input: %w", err)
	}
	inputPath := filepath.Join(scratch, "input.json")
	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		return nil, fmt.Errorf("oracle: defect free: write input: %w", err)
	}

	raw, err := runPython(ctx, opts, "derive_defect_free.py", inputPath)
	if err != nil {
		return nil, fmt.Errorf("oracle: defect free: %w", err)
	}
	return raw, nil
}
