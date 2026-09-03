package analysis

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
)

// ComponentRowFloor is the row count corpus/analysis/components.json
// must not fall below. The table is generated, so a shrink is a
// derivation that quietly stopped emitting rows rather than an edit
// anyone would notice in review.
const ComponentRowFloor = 72

// componentMethods is the coverage set every generated table has to
// reach: each Distance adder, the two composite entry points the adders
// are reached through in real matching, and the one derived view
// (generic_penalty_keys) whose zero-penalty exclusion beets tests on its
// own. A table naming a method outside this list is rejected too: the
// list is what a rung 3 reader can rely on being able to consume.
var componentMethods = []string{
	"add",
	"add_data_source",
	"add_equality",
	"add_expr",
	"add_number",
	"add_priority",
	"add_ratio",
	"add_string",
	"generic_penalty_keys",
	"string_dist",
	"track_distance",
}

// ComponentTable is corpus/analysis/components.json: the component-level
// derivation table internal/harness/oracle's derive_component.py emits
// by calling pinned beets' distance primitives on literal arguments.
//
// Its consumer is rung 3, the Go matcher, which will compare its own
// primitives against these observed outputs. Nothing at rung 2 can do
// that: there is no Go matcher yet, so the checks in this file are
// schema, coverage, and non-emptiness only. They are not a conformance
// check and must not be read as one.
type ComponentTable struct {
	Beets       string         `json:"beets"`
	ShimVersion string         `json:"shim_version"`
	Rows        []ComponentRow `json:"rows"`
}

// ComponentRow is one derived call. Args and Output stay raw: their
// shape is per method (a call sequence and a penalty list for an adder,
// two strings and a float for string_dist), and rung 3 is what gives
// them their Go types.
type ComponentRow struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Args   json.RawMessage `json:"args"`
	Output json.RawMessage `json:"output"`
}

// LoadComponentTable reads corpus/analysis/components.json from path.
func LoadComponentTable(path string) (*ComponentTable, error) {
	var t ComponentTable
	if err := readJSON(path, &t); err != nil {
		return nil, fmt.Errorf("component table: %w", err)
	}
	return &t, nil
}

// CheckComponentTable validates the table's schema, the uniqueness of
// its row ids, its row-count floor, and its per-method coverage.
func CheckComponentTable(t *ComponentTable) error {
	var errs []error
	if t.Beets == "" {
		errs = append(errs, errors.New("component table: beets: missing"))
	}
	if t.ShimVersion == "" {
		errs = append(errs, errors.New("component table: shim_version: missing"))
	}
	if len(t.Rows) < ComponentRowFloor {
		errs = append(errs, fmt.Errorf("component table: %d rows, want at least %d", len(t.Rows), ComponentRowFloor))
	}

	seen := make(map[string]bool, len(t.Rows))
	covered := make(map[string]bool, len(componentMethods))
	for i, r := range t.Rows {
		switch {
		case r.ID == "":
			errs = append(errs, fmt.Errorf("component table: rows[%d]: id: missing", i))
		case seen[r.ID]:
			errs = append(errs, fmt.Errorf("component table: rows[%d]: id %q is not unique", i, r.ID))
		}
		seen[r.ID] = true
		if r.Method == "" {
			errs = append(errs, fmt.Errorf("component table: rows[%d]: method: missing", i))
		}
		covered[r.Method] = true
		if len(r.Args) == 0 {
			errs = append(errs, fmt.Errorf("component table: %s: args: missing", r.ID))
		}
		if len(r.Output) == 0 {
			errs = append(errs, fmt.Errorf("component table: %s: output: missing", r.ID))
		}
	}

	for _, method := range componentMethods {
		if !covered[method] {
			errs = append(errs, fmt.Errorf("component table: no row exercises %s", method))
		}
	}
	for _, method := range slices.Sorted(maps.Keys(covered)) {
		if !slices.Contains(componentMethods, method) {
			errs = append(errs, fmt.Errorf("component table: unrecognized method %q", method))
		}
	}

	return errors.Join(errs...)
}
