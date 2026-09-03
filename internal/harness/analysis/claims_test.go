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

// TestCheckClaimsZeroCandidateMode proves the declared zero-candidate
// mode (min_candidates: 0, best_album_id absent) does what it claims
// rather than trivially passing on any candidate count: it accepts a
// record that carries no candidate at all, and it rejects one that
// carries a candidate the claim never expected, which is the mutation
// that proves the mode cannot mask a degraded normal case.
func TestCheckClaimsZeroCandidateMode(t *testing.T) {
	c, err := LoadClaims("testdata/claims/zero-candidates")
	if err != nil {
		t.Fatalf("LoadClaims: %v", err)
	}

	t.Run("a record with genuinely no candidate satisfies the claim", func(t *testing.T) {
		exp := loadRecord(t, "records/zero-candidates.json")
		if err := CheckClaims(exp, c); err != nil {
			t.Fatalf("CheckClaims: %v", err)
		}
	})

	t.Run("a record carrying a candidate fails the claim", func(t *testing.T) {
		exp := loadRecord(t, "claims/record.json")
		exp.Recommendation = "none" // isolate the candidate-count assertion from the unrelated recommendation mismatch
		err := CheckClaims(exp, c)
		if err == nil {
			t.Fatal("CheckClaims accepted a zero-candidate claim against a record that carries a candidate")
		}
		if !strings.Contains(err.Error(), "carries 1") {
			t.Errorf("error %q does not name the candidate count the claim failed to mask", err)
		}
	})
}

// TestLoadClaimsAllowsZeroCandidateModeWithNoBestAlbumID proves the
// best_album_id gate LoadClaims otherwise enforces is lifted only when
// min_candidates: 0 is declared, not simply whenever the field is
// absent.
func TestLoadClaimsAllowsZeroCandidateModeWithNoBestAlbumID(t *testing.T) {
	c, err := LoadClaims("testdata/claims/zero-candidates")
	if err != nil {
		t.Fatalf("LoadClaims rejected a zero-candidate claims file: %v", err)
	}
	if c.BestAlbumID != "" {
		t.Errorf("best_album_id = %q, want empty", c.BestAlbumID)
	}
}
