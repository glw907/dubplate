package conform

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/harness/analysis"
)

// TestNullabilityPolicyCoversEveryField asserts fieldNullability's row
// count against reflect.TypeFor[matcher.AlbumInfo]().NumField() - 1 +
// reflect.TypeFor[matcher.TrackInfo]().NumField(), the "- 1" being
// AlbumInfo.Tracks (compared as a list, never through the policy). A
// fifteenth vocabulary field added to either struct without extending
// fieldNullability fails this test (fold, M-c).
func TestNullabilityPolicyCoversEveryField(t *testing.T) {
	want := reflect.TypeFor[matcher.AlbumInfo]().NumField() - 1 + reflect.TypeFor[matcher.TrackInfo]().NumField()
	if got := len(fieldNullability); got != want {
		t.Fatalf("fieldNullability carries %d rows, want %d (AlbumInfo.NumField()-1 + TrackInfo.NumField())", got, want)
	}

	var strict, never, zero int
	for _, row := range fieldNullability {
		switch row.Policy {
		case nullStrict:
			strict++
		case nullNeverExpected:
			never++
		case nullEqualsZero:
			zero++
		}
		if row.Reason == "" {
			t.Errorf("field %s carries no reason", row.Field)
		}
	}
	if strict != 3 || never != 9 || zero != 9 {
		t.Errorf("fieldNullability policy split = %d strict, %d never-expected, %d zero-equivalent; want 3, 9, 9", strict, never, zero)
	}
}

// wantCleanAlbum returns a fully populated LedgerAlbumInfo/matcher.AlbumInfo
// pair that agree on every field, so each mutation test below can
// perturb exactly one field from a known-good baseline.
func wantCleanAlbum() (analysis.LedgerAlbumInfo, *matcher.AlbumInfo) {
	want := analysis.LedgerAlbumInfo{
		Artist:        new("Fleetwood Mac"),
		Album:         new("Rumours"),
		VA:            new(false),
		Media:         new("CD"),
		Mediums:       new(1),
		Year:          new(1977),
		OriginalYear:  new(1977),
		Country:       new("US"),
		Label:         new("Warner Bros."),
		CatalogNum:    new("BSK 3010"),
		AlbumDisambig: new(""),
		AlbumID:       new("1a8c4ac3-6f47-4570-aedb-f1928a0498ea"),
		DataSource:    new("MusicBrainz"),
		Tracks: []analysis.LedgerTrackInfo{
			{
				Title:       new("Second Hand News"),
				Artist:      new("Fleetwood Mac"),
				Length:      new(173.666),
				Index:       new(1),
				MediumIndex: new(1),
				Medium:      new(1),
				TrackID:     new("b25f351b-206f-4297-95bd-78459b704144"),
				DataSource:  new("MusicBrainz"),
			},
		},
	}
	got := &matcher.AlbumInfo{
		Artist:        "Fleetwood Mac",
		Album:         "Rumours",
		VA:            false,
		Media:         "CD",
		Mediums:       1,
		Year:          1977,
		OriginalYear:  1977,
		Country:       "US",
		Label:         "Warner Bros.",
		CatalogNum:    "BSK 3010",
		AlbumDisambig: "",
		AlbumID:       "1a8c4ac3-6f47-4570-aedb-f1928a0498ea",
		DataSource:    "MusicBrainz",
		Tracks: []matcher.TrackInfo{
			{
				Title:       "Second Hand News",
				Artist:      "Fleetwood Mac",
				Length:      173.666,
				Index:       1,
				MediumIndex: 1,
				Medium:      1,
				TrackID:     "b25f351b-206f-4297-95bd-78459b704144",
				DataSource:  "MusicBrainz",
			},
		},
	}
	return want, got
}

// TestCheckTranslatedAlbumAgreesOnACleanPair proves the baseline both
// mutation tests below perturb is itself clean, so a later failure can
// only be attributed to the perturbation.
func TestCheckTranslatedAlbumAgreesOnACleanPair(t *testing.T) {
	want, got := wantCleanAlbum()
	if err := CheckTranslatedAlbum("clean-match", "1a8c4ac3", want, got); err != nil {
		t.Fatalf("expected no disagreement on a clean pair: %v", err)
	}
}

// TestCheckTranslatedAlbumStrictNullFires proves mutation 1: a
// hand-authored ledger with album.album null fires the strict half,
// naming the case and the field, regardless of the Go side's value.
func TestCheckTranslatedAlbumStrictNullFires(t *testing.T) {
	want, got := wantCleanAlbum()
	want.Album = nil

	err := CheckTranslatedAlbum("clean-match", "1a8c4ac3", want, got)
	if err == nil {
		t.Fatal("expected a null album.album to fire the strict policy")
	}
	if !strings.Contains(err.Error(), "clean-match") || !strings.Contains(err.Error(), "album.album") {
		t.Errorf("error does not name the case and field: %v", err)
	}
}

// TestCheckTranslatedAlbumLenientNullDoesNotFire proves mutation 2, a
// negative assertion: album.media null against a Go "" does not fire,
// because media is nullEqualsZero. This is the mutation that proves
// the policy is not uniformly strict.
func TestCheckTranslatedAlbumLenientNullDoesNotFire(t *testing.T) {
	want, got := wantCleanAlbum()
	want.Media = nil
	got.Media = ""

	if err := CheckTranslatedAlbum("clean-match", "1a8c4ac3", want, got); err != nil {
		t.Errorf("expected a null album.media against a Go \"\" not to fire: %v", err)
	}
}

// TestCheckTranslatedAlbumNeverExpectedNullFires proves mutation 3:
// album.va null fires the never-expected policy unconditionally.
func TestCheckTranslatedAlbumNeverExpectedNullFires(t *testing.T) {
	want, got := wantCleanAlbum()
	want.VA = nil

	err := CheckTranslatedAlbum("clean-match", "1a8c4ac3", want, got)
	if err == nil {
		t.Fatal("expected a null album.va to fire the never-expected policy")
	}
	if !strings.Contains(err.Error(), "album.va") {
		t.Errorf("error does not name the field: %v", err)
	}
}

// TestCheckTranslatedAlbumCatchesPerturbedLabel proves mutation 4: a
// one-character perturbation of album.label fails naming label and
// both values.
func TestCheckTranslatedAlbumCatchesPerturbedLabel(t *testing.T) {
	want, got := wantCleanAlbum()
	got.Label = "Warner Bros" // one character short of "Warner Bros."

	err := CheckTranslatedAlbum("clean-match", "1a8c4ac3", want, got)
	if err == nil {
		t.Fatal("expected a perturbed label to fire")
	}
	msg := err.Error()
	if !strings.Contains(msg, "album.label") || !strings.Contains(msg, "Warner Bros.") || !strings.Contains(msg, "Warner Bros") {
		t.Errorf("error does not name label and both values: %v", err)
	}
}

// TestCheckTranslatedAlbumCatchesShortTrackList proves mutation 5: a
// track list one element short of the ledger's own fires the length
// mismatch, naming both counts, and does not attempt a per-track
// comparison (which would otherwise misreport a shifted list).
func TestCheckTranslatedAlbumCatchesShortTrackList(t *testing.T) {
	want, got := wantCleanAlbum()
	want.Tracks = append(want.Tracks, analysis.LedgerTrackInfo{
		Title:       new("Dreams"),
		Artist:      new("Fleetwood Mac"),
		Length:      new(257.533),
		Index:       new(2),
		MediumIndex: new(2),
		Medium:      new(1),
		TrackID:     new("248cc9d1-97ea-493e-84d4-4c5ec718683b"),
		DataSource:  new("MusicBrainz"),
	})
	// got.Tracks stays at length 1; want now carries 2.

	err := CheckTranslatedAlbum("clean-match", "1a8c4ac3", want, got)
	if err == nil {
		t.Fatal("expected a short track list to fire")
	}
	msg := err.Error()
	if !strings.Contains(msg, "tracks") || !strings.Contains(msg, "1") || !strings.Contains(msg, "2") {
		t.Errorf("error does not name both track counts: %v", err)
	}
}

// TestMergeProfileOverlaysTunables proves the three profile names'
// overlay: stock and estate change nothing the translation reads, and
// locale sets Languages and IgnoredAliasTypes.
func TestMergeProfileOverlaysTunables(t *testing.T) {
	base := matcher.DefaultTunables()

	for _, name := range []string{"stock", "estate"} {
		got, err := mergeProfile(base, name)
		if err != nil {
			t.Fatalf("mergeProfile(%s): %v", name, err)
		}
		if !reflect.DeepEqual(got, base) {
			t.Errorf("mergeProfile(%s) changed the base tunables, want no change", name)
		}
	}

	locale, err := mergeProfile(base, "locale")
	if err != nil {
		t.Fatalf("mergeProfile(locale): %v", err)
	}
	if !reflect.DeepEqual(locale.Languages, []string{"ja"}) {
		t.Errorf("mergeProfile(locale).Languages = %v, want [ja]", locale.Languages)
	}
	if len(locale.IgnoredAliasTypes) != 0 {
		t.Errorf("mergeProfile(locale).IgnoredAliasTypes = %v, want empty", locale.IgnoredAliasTypes)
	}
}

// TestMergeProfileRejectsUnrecognizedName proves an unrecognized
// profile name is the failure, not a silent no-op overlay.
func TestMergeProfileRejectsUnrecognizedName(t *testing.T) {
	_, err := mergeProfile(matcher.DefaultTunables(), "bogus-profile")
	if err == nil {
		t.Fatal("expected an unrecognized profile name to fail")
	}
	if !strings.Contains(err.Error(), "bogus-profile") {
		t.Errorf("error does not name the profile: %v", err)
	}
}

// TestProfileKeyDriftGuardAgreesWithCommittedProfiles proves the real
// committed profiles carry only allowlisted keys.
func TestProfileKeyDriftGuardAgreesWithCommittedProfiles(t *testing.T) {
	root := repoRoot(t)
	if err := checkProfileKeyDrift(root, []string{"stock", "estate", "locale"}); err != nil {
		t.Error(err)
	}
}

// TestProfileKeyDriftGuardFiresOnAnUnknownKey proves checkProfileKeys
// fires, naming the key, against a hand-authored profile text carrying
// a match: key the allowlist does not carry.
func TestProfileKeyDriftGuardFiresOnAnUnknownKey(t *testing.T) {
	fixture := "plugins: musicbrainz\nmatch:\n  strong_rec_thresh: 0.1\n"
	err := checkProfileKeys("fixture", fixture)
	if err == nil {
		t.Fatal("expected checkProfileKeys to fire on an unrecognized top-level key")
	}
	if !strings.Contains(err.Error(), "match") {
		t.Errorf("error does not name the unrecognized key: %v", err)
	}
}

// TestProfileKeyDriftGuardFiresOnAnUnknownImportKey proves the guard
// also scans inside import:, against a hand-authored import.bogus_key.
func TestProfileKeyDriftGuardFiresOnAnUnknownImportKey(t *testing.T) {
	fixture := "plugins: musicbrainz\nimport:\n  bogus_key: yes\n"
	err := checkProfileKeys("fixture", fixture)
	if err == nil {
		t.Fatal("expected checkProfileKeys to fire on an unrecognized import key")
	}
	if !strings.Contains(err.Error(), "import.bogus_key") {
		t.Errorf("error does not name the unrecognized key: %v", err)
	}
}

// TestSkippedLedgerRowFailsTheCount proves mutation 6: a fixture ledger
// naming a real committed case but a release mbid that case never
// logged cannot be resolved, so gradeLedger both reports an error and
// leaves compared short of the ledger's own row count. A row counter
// hardcoded to len(l.Rows) instead of counting actual comparisons would
// pass this fixture silently; this test is what catches that.
func TestSkippedLedgerRowFailsTheCount(t *testing.T) {
	root := repoRoot(t)
	ledger, err := analysis.LoadAlbumInfoLedger(filepath.Join("testdata", "translate", "skipped-row-ledger.json"))
	if err != nil {
		t.Fatalf("load fixture ledger: %v", err)
	}

	compared, gradeErr := gradeLedger(root, ledger)
	if gradeErr == nil {
		t.Fatal("expected gradeLedger to report an error for the unresolvable row")
	}
	if compared == len(ledger.Rows) {
		t.Errorf("compared (%d) equals the ledger's row count (%d); the skipped row was not caught", compared, len(ledger.Rows))
	}
}

// TestTranslationLedgerGradesCleanly is the pass's own instrument: it
// reads corpus/analysis/album_info.json, rebuilds every row's AlbumInfo
// through the Go translation, and asserts the runner compared every
// row (a skipped row is a failure, tier 0's own mitigation) and
// disagreed with none.
func TestTranslationLedgerGradesCleanly(t *testing.T) {
	root := repoRoot(t)
	ledger, err := analysis.LoadAlbumInfoLedger(filepath.Join(root, "corpus", "analysis", "album_info.json"))
	if err != nil {
		t.Fatalf("load corpus/analysis/album_info.json: %v", err)
	}

	start := time.Now()
	compared, gradeErr := gradeLedger(root, ledger)
	elapsed := time.Since(start)
	t.Logf("graded %d/%d ledger rows in %s (%s/row)", compared, len(ledger.Rows), elapsed, elapsed/time.Duration(max(len(ledger.Rows), 1)))

	if compared != len(ledger.Rows) {
		t.Errorf("compared %d rows, want %d (the ledger's own row count); a row was skipped", compared, len(ledger.Rows))
	}
	if gradeErr != nil {
		t.Error(gradeErr)
	}
}

// TestNoCommittedReleaseReachesTheBrowseThreshold is the browse-fallback
// guard (R9, T16): no committed release-lookup response may sum
// len(medium.tracks) to browseMaxTracks or more.
func TestNoCommittedReleaseReachesTheBrowseThreshold(t *testing.T) {
	root := repoRoot(t)
	if err := checkNoReleaseReachesBrowseThreshold(root, browseMaxTracks); err != nil {
		t.Error(err)
	}
}

// TestBrowseThresholdGuardFiresOnLoweredThreshold proves mutation 7:
// BROWSE_MAXTRACKS lowered to 400 in a test-local copy fires naming
// both real-boxset-rachmaninov-complete and its -stock twin, each at
// 470, proving the guard can see the real corpus rather than passing
// vacuously.
func TestBrowseThresholdGuardFiresOnLoweredThreshold(t *testing.T) {
	root := repoRoot(t)
	err := checkNoReleaseReachesBrowseThreshold(root, 400)
	if err == nil {
		t.Fatal("expected the browse threshold guard to fire at a lowered threshold")
	}
	msg := err.Error()
	if !strings.Contains(msg, "real-boxset-rachmaninov-complete") {
		t.Errorf("error does not name real-boxset-rachmaninov-complete: %v", err)
	}
	if !strings.Contains(msg, "real-boxset-rachmaninov-complete-stock") {
		t.Errorf("error does not name real-boxset-rachmaninov-complete-stock: %v", err)
	}
	if !strings.Contains(msg, "470") {
		t.Errorf("error does not name the track count 470: %v", err)
	}
}
