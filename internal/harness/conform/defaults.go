package conform

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// CheckCaseWeights asserts that got, a committed case's echoed
// match.distance_weights, equals want key by key, naming the key and
// both values on any disagreement. This is the M1 defaults
// cross-check's own half that can fire before tier 1 exists: no
// committed profile sets a match: key, so every one of the 216
// committed cases echoes the same twenty shipped defaults, and a
// transcription error in Tunables' own defaults shows up here
// immediately rather than waiting on 3d's profile-merge runner.
func CheckCaseWeights(caseID string, want, got map[string]float64) error {
	var errs []error
	for _, key := range sortedKeys(want) {
		g, ok := got[key]
		if !ok {
			errs = append(errs, fmt.Errorf("case %s: weights: %s: missing (want %v)", caseID, key, want[key]))
			continue
		}
		if g != want[key] {
			errs = append(errs, fmt.Errorf("case %s: weights: %s: got %v, want %v", caseID, key, g, want[key]))
		}
	}
	for _, key := range sortedKeys(got) {
		if _, ok := want[key]; !ok {
			errs = append(errs, fmt.Errorf("case %s: weights: %s: unexpected (not one of Tunables' defaults)", caseID, key))
		}
	}
	return errors.Join(errs...)
}

// CheckCaseThresholds asserts that got, a committed case's echoed
// match thresholds and max_rec, equals want's shipped defaults.
func CheckCaseThresholds(caseID string, want matcher.Tunables, got cassette.Thresholds) error {
	var errs []error
	if got.StrongRecThresh != want.StrongRecThresh {
		errs = append(errs, fmt.Errorf("case %s: strong_rec_thresh: got %v, want %v", caseID, got.StrongRecThresh, want.StrongRecThresh))
	}
	if got.MediumRecThresh != want.MediumRecThresh {
		errs = append(errs, fmt.Errorf("case %s: medium_rec_thresh: got %v, want %v", caseID, got.MediumRecThresh, want.MediumRecThresh))
	}
	if got.RecGapThresh != want.RecGapThresh {
		errs = append(errs, fmt.Errorf("case %s: rec_gap_thresh: got %v, want %v", caseID, got.RecGapThresh, want.RecGapThresh))
	}
	if !maps.Equal(got.MaxRec, want.MaxRec) {
		errs = append(errs, fmt.Errorf("case %s: max_rec: got %v, want %v", caseID, got.MaxRec, want.MaxRec))
	}
	return errors.Join(errs...)
}

// gradeTunables is the configuration every conformance sweep runs
// under: Tunables' beets defaults merged with the case's profile, in
// that order (rung 3 spec, the M1 fold), with import.quiet forced true.
//
// The quiet overlay lives here and outside mergeProfile, because
// mergeProfile models the three committed profile FILES and
// checkProfileKeyDrift's allowlist is honest only while that stays
// true. quiet is in no profile: derive_oracle.py sets
// config["import"]["quiet"] = True before calling beets' own mapping
// (internal/harness/oracle/derive_oracle.py:204), because dubplate
// imports with `beet import -q` and quiet is a CLI flag. Without the
// overlay every committed action would grade as "ask".
func gradeTunables(profile string) (matcher.Tunables, error) {
	tun, err := mergeProfile(matcher.DefaultTunables(), profile)
	if err != nil {
		return matcher.Tunables{}, err
	}
	tun.Quiet = true
	return tun, nil
}

// CheckUnidecodeVersion asserts that manifest's pinned unidecode
// version equals the engine's own embedded transliteration table
// version (fold, M-c). Task 3 exports the stamp from package matcher;
// this is the one place the corpus's own pin is read back against it,
// because reading the corpus from the harness is the direction the
// architectural invariant permits (see arch_test.go), and reading it
// from the engine would trip that same check.
func CheckUnidecodeVersion(manifest cassette.Manifest) error {
	if manifest.Unidecode != matcher.UnidecodeVersion {
		return fmt.Errorf("unidecode version: manifest pins %q, matcher embeds %q", manifest.Unidecode, matcher.UnidecodeVersion)
	}
	return nil
}

func sortedKeys(m map[string]float64) []string {
	return slices.Sorted(maps.Keys(m))
}
