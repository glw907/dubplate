package analysis

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// FileTagLedgerRowFloor is the row count corpus/analysis/file_tags.json
// must not fall below. The ledger is generated over a fixed set of
// twelve named subjects, so a shrink here means the derivation quietly
// dropped a subject rather than an edit anyone would notice in review.
const FileTagLedgerRowFloor = 12

// FileTagSubject names one case, and the item within it, the file-tags
// ledger derives a row for. All twelve named subjects here use item
// index 0, the case's first item; the type carries the index anyway so
// a future subject naming a later item is not a schema change.
type FileTagSubject struct {
	CaseID    string
	ItemIndex int
}

// FileTagSubjects is the fixed twelve-subject set
// corpus/analysis/file_tags.json is derived over: five smoke/structural
// cases plus seven real-catalog cases, per rung 2 pass 2d's Task 2. The
// twelfth entry, `real-decoy-mozart-requiem`, replaces the dossier's
// original `real-decoy-beethoven-symphony-5`: that id never resolved to
// a committed case (pass 2c's Task 4 found the decoy never surfaced in
// beets' own 5-result search window and the axis's reserve was already
// spent, corpus/README.md's "The two-release sub-axis finding"
// section), so it can never be a file-tags subject. The replacement is
// a conductor ruling recorded in corpus/README.md's file-tags section,
// not an implementer's own substitution.
var FileTagSubjects = []FileTagSubject{
	{CaseID: "clean-match"},
	{CaseID: "media-uneven-split"},
	{CaseID: "va-exact"},
	{CaseID: "non-latin-pseudo-release"},
	{CaseID: "year-reissue-ratio-penalty"},
	{CaseID: "real-classical-karajan-beethoven-89"},
	{CaseID: "real-reissue-nevermind-2011"},
	{CaseID: "real-boxset-james-brown-star-time"},
	{CaseID: "real-nonlatin-utada-first-love"},
	{CaseID: "real-noise-dylan-bootleg"},
	{CaseID: "real-genre-public-enemy-nation-millions"},
	{CaseID: "real-decoy-mozart-requiem"},
}

// FileTagLedger is corpus/analysis/file_tags.json: the file-level
// applied-tag ledger internal/harness/oracle's derive_file_tags.py
// emits by writing each subject's recorded fields into a fresh copy of
// the vendored template FLAC and reading the result back independently.
//
// Its consumer is rung 8, pure-Go FLAC tag writing: WritableFields is
// the conformance target, and Template holds the read-only remainder
// every read-back reports from the container itself, never from a
// subject. Nothing at rung 2 can compare against a Go writer, so the
// checks in this file are schema, coverage, and key-membership only.
type FileTagLedger struct {
	ShimVersion    string             `json:"shim_version"`
	BeetsVersion   string             `json:"beets_version"`
	WritableFields []string           `json:"writable_fields"`
	ReadonlyFields []string           `json:"readonly_fields"`
	Template       map[string]any     `json:"template"`
	Rows           []FileTagLedgerRow `json:"rows"`
}

// FileTagLedgerRow is one subject's write/read-back result. Values
// holds only the writable set; the read-only remainder every read-back
// reports lives once, in the ledger's own shared Template block.
type FileTagLedgerRow struct {
	CaseID    string         `json:"case_id"`
	ItemIndex int            `json:"item_index"`
	Values    map[string]any `json:"values"`
}

// LoadFileTagLedger reads corpus/analysis/file_tags.json from path.
func LoadFileTagLedger(path string) (*FileTagLedger, error) {
	var l FileTagLedger
	if err := readJSON(path, &l); err != nil {
		return nil, fmt.Errorf("file tag ledger: %w", err)
	}
	return &l, nil
}

// CheckFileTagLedger validates the ledger's schema, that every row
// resolves into a committed corpus case's own Expected.AppliedFields,
// that every row's value keys sit inside the ledger's own writable-set
// header (never a second, hand-maintained copy of that list) and never
// carry a read-only key, that the row set is exactly the twelve named
// FileTagSubjects, and the row-count floor. cases is every committed
// corpus case, keyed by id, so a row naming an id outside the twelve
// subjects is still recognized as a real (but wrong) corpus case rather
// than reported as unresolvable.
func CheckFileTagLedger(cases map[string]*cassette.Case, l *FileTagLedger) error {
	var errs []error

	if l.ShimVersion == "" {
		errs = append(errs, errors.New("file tag ledger: shim_version: missing"))
	}
	if l.BeetsVersion == "" {
		errs = append(errs, errors.New("file tag ledger: beets_version: missing"))
	}
	if len(l.WritableFields) == 0 {
		errs = append(errs, errors.New("file tag ledger: writable_fields: missing"))
	}
	if len(l.ReadonlyFields) == 0 {
		errs = append(errs, errors.New("file tag ledger: readonly_fields: missing"))
	}
	if len(l.Rows) < FileTagLedgerRowFloor {
		errs = append(errs, fmt.Errorf("file tag ledger: %d rows, want at least %d", len(l.Rows), FileTagLedgerRowFloor))
	}

	writable := make(map[string]bool, len(l.WritableFields))
	for _, f := range l.WritableFields {
		writable[f] = true
	}
	readonly := make(map[string]bool, len(l.ReadonlyFields))
	for _, f := range l.ReadonlyFields {
		readonly[f] = true
		if _, ok := l.Template[f]; !ok {
			errs = append(errs, fmt.Errorf("file tag ledger: template block is missing read-only field %q", f))
		}
	}

	wantSubjects := make(map[string]bool, len(FileTagSubjects))
	for _, s := range FileTagSubjects {
		wantSubjects[s.CaseID] = true
	}
	gotSubjects := make(map[string]bool, len(l.Rows))

	for i, row := range l.Rows {
		c, ok := cases[row.CaseID]
		if !ok {
			errs = append(errs, fmt.Errorf("file tag ledger: rows[%d]: case %q is not a committed corpus case", i, row.CaseID))
			continue
		}
		gotSubjects[row.CaseID] = true
		if !wantSubjects[row.CaseID] {
			errs = append(errs, fmt.Errorf("file tag ledger: %s: not one of the twelve named file-tags subjects", row.CaseID))
		}
		if c.Expected == nil {
			errs = append(errs, fmt.Errorf("file tag ledger: %s: case carries no expected record to resolve against", row.CaseID))
			continue
		}
		if _, ok := c.Expected.AppliedFields[strconv.Itoa(row.ItemIndex)]; !ok {
			errs = append(errs, fmt.Errorf("file tag ledger: %s: item index %d does not resolve into the case's own applied_fields", row.CaseID, row.ItemIndex))
		}
		for _, key := range slices.Sorted(maps.Keys(row.Values)) {
			if readonly[key] {
				errs = append(errs, fmt.Errorf("file tag ledger: %s: row carries read-only field %q; it belongs only in the shared template block", row.CaseID, key))
				continue
			}
			if !writable[key] {
				errs = append(errs, fmt.Errorf("file tag ledger: %s: row carries key %q outside the ledger's own writable_fields header", row.CaseID, key))
			}
		}
	}

	for _, s := range FileTagSubjects {
		if !gotSubjects[s.CaseID] {
			errs = append(errs, fmt.Errorf("file tag ledger: missing row for subject %q", s.CaseID))
		}
	}

	return errors.Join(errs...)
}
