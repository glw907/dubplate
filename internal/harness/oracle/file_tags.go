//go:build oracle

// file_tags.go drives derive_file_tags.py, this package's third beets
// entry point: no stub, no rendered profile, and no MusicBrainz
// response, matching component.go's minimalism. It never runs from
// [Derive], and it never gives an item a path that reaches the
// corpus: every write happens inside a scratch copy of the vendored
// template FLAC (testdata/template.flac).
package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileTagSubject is one subject the file-tags ledger derives a row
// for: a committed case id, the item's index in that case's own
// Expected.AppliedFields, and the field values themselves, read
// straight off the corpus rather than re-specified here.
type FileTagSubject struct {
	CaseID        string         `json:"case_id"`
	ItemIndex     int            `json:"item_index"`
	AppliedFields map[string]any `json:"applied_fields"`
}

// DeriveFileTags runs derive_file_tags.py against subjects and
// returns the ledger document it prints, unparsed: one process for
// every subject, one JSON document out. Which cases are subjects is
// the caller's decision; this driver only owns the mechanism of
// deriving a row for whichever subjects it is given.
func DeriveFileTags(ctx context.Context, opts Options, subjects []FileTagSubject) ([]byte, error) {
	scratch, err := os.MkdirTemp("", "dubplate-file-tags-*")
	if err != nil {
		return nil, fmt.Errorf("oracle: derive file tags: scratch dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	data, err := json.Marshal(subjects)
	if err != nil {
		return nil, fmt.Errorf("oracle: derive file tags: encode subjects: %w", err)
	}
	inputPath := filepath.Join(scratch, "input.json")
	if err := os.WriteFile(inputPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("oracle: derive file tags: write subjects: %w", err)
	}

	raw, err := runPython(ctx, opts, "derive_file_tags.py", inputPath)
	if err != nil {
		return nil, fmt.Errorf("oracle: derive file tags: %w", err)
	}
	return raw, nil
}
