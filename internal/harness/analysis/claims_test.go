package analysis

import (
	"strings"
	"testing"
)

// TestCheckClaimsAgainstARealRecord runs the checker over a verbatim
// copy of the committed clean-match record and a set of claims files
// mutated one field at a time. The good/ claims pass; every mutation
// has to fail and name what it got wrong, which is what proves the
// checker is not simply returning nil.
func TestCheckClaimsAgainstARealRecord(t *testing.T) {
	exp := loadRecord(t, "claims/record.json")

	tests := []struct {
		name    string
		dir     string
		wantHas string // empty means the claims are satisfied
	}{
		{"written off the record", "good", ""},
		{"best album id changed by one character", "mutated-album-id", "best_album_id"},
		{"a zero penalty claimed nonzero", "mutated-penalty", `penalty "year" claimed nonzero`},
		{"one assignment pair moved", "mutated-assignment", "assignment item 3"},
		{"an applied spot value moved", "mutated-applied-field", "applied_fields item 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := LoadClaims("testdata/claims/" + tt.dir)
			if err != nil {
				t.Fatalf("LoadClaims: %v", err)
			}
			err = CheckClaims(exp, c)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckClaims: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckClaims accepted claims whose %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

func TestLoadClaimsRejectsUnusableFiles(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantHas string
	}{
		{"a claim key no parser recognizes", "testdata/claims/unknown-key", "best_distance"},
		{"no best_album_id", "testdata/claims/missing-best-album-id", "best_album_id"},
		{"no claims file at all", "testdata/claims", "no such file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadClaims(tt.dir)
			if err == nil {
				t.Fatalf("LoadClaims accepted a case with %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

// TestCheckClaimsRefusesToVerifyAnEmptyCandidateList proves the checker
// reports what it cannot verify instead of indexing an empty slice: a
// zero-track candidate release is dropped by beets' own _add_candidate,
// so a committed record with no candidate at all is reachable.
func TestCheckClaimsRefusesToVerifyAnEmptyCandidateList(t *testing.T) {
	exp := loadRecord(t, "records/zero-candidates.json")
	c, err := LoadClaims("testdata/claims/good")
	if err != nil {
		t.Fatalf("LoadClaims: %v", err)
	}
	err = CheckClaims(exp, c)
	if err == nil {
		t.Fatal("CheckClaims accepted claims about a record with no candidates")
	}
	if !strings.Contains(err.Error(), "no candidate") {
		t.Errorf("error %q does not name the empty candidate list", err)
	}
}
