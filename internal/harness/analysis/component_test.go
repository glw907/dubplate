package analysis

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// fullTable builds a table that satisfies every check, so each subtest
// below can break exactly one thing and attribute the failure to it.
func fullTable() *ComponentTable {
	t := &ComponentTable{Beets: "2.13.1", ShimVersion: "1"}
	for i := range ComponentRowFloor {
		method := componentMethods[i%len(componentMethods)]
		t.Rows = append(t.Rows, ComponentRow{
			ID:     fmt.Sprintf("%s/%d", method, i),
			Method: method,
			Args:   json.RawMessage(`{"key":"album"}`),
			Output: json.RawMessage(`{"penalties":[0.5]}`),
		})
	}
	return t
}

func TestCheckComponentTable(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*ComponentTable)
		wantHas string // empty means the table is accepted
	}{
		{name: "a complete table", mutate: func(*ComponentTable) {}},
		{
			name:    "a shrunken table",
			mutate:  func(tbl *ComponentTable) { tbl.Rows = tbl.Rows[:4] },
			wantHas: "want at least",
		},
		{
			name: "a method that stopped being exercised",
			mutate: func(tbl *ComponentTable) {
				for i := range tbl.Rows {
					if tbl.Rows[i].Method == "add" {
						tbl.Rows[i].Method = "add_string"
					}
				}
			},
			wantHas: "no row exercises add",
		},
		{
			name:    "a method nothing recognizes",
			mutate:  func(tbl *ComponentTable) { tbl.Rows[0].Method = "add_everything" },
			wantHas: `unrecognized method "add_everything"`,
		},
		{
			name:    "two rows sharing an id",
			mutate:  func(tbl *ComponentTable) { tbl.Rows[1].ID = tbl.Rows[0].ID },
			wantHas: "is not unique",
		},
		{
			name:    "a row with no observed output",
			mutate:  func(tbl *ComponentTable) { tbl.Rows[2].Output = nil },
			wantHas: "output: missing",
		},
		{
			name:    "no beets pin",
			mutate:  func(tbl *ComponentTable) { tbl.Beets = "" },
			wantHas: "beets: missing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := fullTable()
			tt.mutate(tbl)
			err := CheckComponentTable(tbl)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckComponentTable: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckComponentTable accepted a table with %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}
