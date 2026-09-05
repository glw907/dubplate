package conform

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// releaseLookupPrefix is the WS/2 path a release lookup by id starts
// with, mirroring analysis.releaseLookupPrefix's own definition
// (unexported there, behind the territory wall).
const releaseLookupPrefix = "/ws/2/release/"

// browseMaxTracks is BROWSE_MAXTRACKS (beetsplug/musicbrainz.py:67),
// the track count beyond which album_for_id switches from a direct
// release fetch to the paged browse endpoint (:609-613). Issuing that
// alternate request is 3c's client work; this guard exists so a corpus
// case crossing the line is noticed the moment it lands, since nothing
// else in the pipeline would otherwise notice a lookup silently taking
// the untested path.
const browseMaxTracks = 500

// nullPolicy says what a JSON null in the ledger means for one field.
// It exists because beets' own consumers do not treat them alike: three
// add_string call sites on the whole distance path are unguarded by a
// truthiness test on the info side, so only those three can turn a None
// into arithmetic, and for every other field None, "" and 0 are
// indistinguishable at the consumer.
type nullPolicy int

const (
	// nullStrict: a null here is a gate failure. These are the three
	// unguarded add_string sites (autotag/distance.py:451, :448, :402),
	// the only shapes in which a real divergence could arrive. The
	// MusicBrainz translation cannot produce one today, and this is
	// the check that would catch a source that could.
	nullStrict nullPolicy = iota

	// nullNeverExpected: the derivation cannot produce a null here, so
	// one means the derivation changed rather than that the value is
	// absent.
	nullNeverExpected

	// nullEqualsZero: a null compares equal to the Go zero value,
	// because beets' own truthiness guard at the consumer makes None,
	// "" and 0 indistinguishable.
	nullEqualsZero
)

// fieldNullability is the twenty-one-field nullability policy (R1,
// R11): every scalar matcher.AlbumInfo and matcher.TrackInfo carries,
// one line of reason each, so a reader sees why a field tolerates null
// without reading beets. Three fields are nullStrict, nine
// nullNeverExpected, nine nullEqualsZero.
//
// The row count is asserted in translate_test.go against
// reflect.TypeOf(matcher.AlbumInfo{}).NumField() - 1 +
// reflect.TypeOf(matcher.TrackInfo{}).NumField(), the "- 1" being
// AlbumInfo.Tracks, which is compared as a list rather than through
// this policy. Adding a field to either struct without extending this
// table fails that assertion.
var fieldNullability = []struct {
	Field  string
	Policy nullPolicy
	Reason string
}{
	{"album.album", nullStrict, "distance.py:451, unguarded add_string"},
	{"album.artist", nullStrict, "distance.py:448, guarded only by \"not va\""},
	{"track.title", nullStrict, "distance.py:402, unguarded add_string"},
	{"album.va", nullNeverExpected, "a Python bool, never None"},
	{"album.mediums", nullNeverExpected, "len(...), may be 0, never None"},
	{"album.album_id", nullNeverExpected, "release[\"id\"], always present"},
	{"album.data_source", nullNeverExpected, "the literal \"MusicBrainz\""},
	{"track.index", nullNeverExpected, "written unconditionally by track_info"},
	{"track.medium_index", nullNeverExpected, "written unconditionally by track_info"},
	{"track.medium", nullNeverExpected, "written unconditionally by track_info"},
	{"track.track_id", nullNeverExpected, "written unconditionally by track_info"},
	{"track.data_source", nullNeverExpected, "written unconditionally by track_info"},
	{"album.media", nullEqualsZero, "truthiness-guarded, distance.py:455; a single-medium null format really does yield None"},
	{"album.year", nullEqualsZero, "truthiness-guarded, distance.py:472,480"},
	{"album.original_year", nullEqualsZero, "truthiness-guarded, distance.py:475,484"},
	{"album.country", nullEqualsZero, "truthiness-guarded, distance.py:498,501"},
	{"album.label", nullEqualsZero, "truthiness-guarded, distance.py:505"},
	{"album.catalognum", nullEqualsZero, "truthiness-guarded, distance.py:509"},
	{"album.albumdisambig", nullEqualsZero, "truthiness-guarded, distance.py:515"},
	{"track.length", nullEqualsZero, "truthiness-guarded, distance.py:397"},
	{"track.artist", nullEqualsZero, "truthiness-guarded, distance.py:404-410"},
}

// checkNullable compares want, a ledger scalar that may be JSON null,
// against got, the Go translation's own zero-or-real value, per
// policy. field and the case/release pair name the disagreement so a
// failure identifies the row without a second lookup.
func checkNullable[T comparable](caseID, mbid, field string, want *T, got T, policy nullPolicy) error {
	if want == nil {
		switch policy {
		case nullStrict:
			return fmt.Errorf("case %s: release %s: %s: ledger carries null, which nullStrict forbids", caseID, mbid, field)
		case nullNeverExpected:
			return fmt.Errorf("case %s: release %s: %s: ledger carries null, which the derivation should never produce", caseID, mbid, field)
		case nullEqualsZero:
			var zero T
			if got != zero {
				return fmt.Errorf("case %s: release %s: %s: ledger carries null (compares equal to the Go zero value), got %v", caseID, mbid, field, got)
			}
		}
		return nil
	}
	if *want != got {
		return fmt.Errorf("case %s: release %s: %s: got %v, want %v", caseID, mbid, field, got, *want)
	}
	return nil
}

// CheckTranslatedAlbum compares one ledger row's recorded AlbumInfo
// against the AlbumInfo the Go translation built, field by field, and
// reports every disagreement rather than the first. caseID and mbid
// name the row in every message so a failure identifies the release
// without a second lookup.
func CheckTranslatedAlbum(caseID, mbid string, want analysis.LedgerAlbumInfo, got *matcher.AlbumInfo) error {
	var errs []error
	add := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	add(checkNullable(caseID, mbid, "album.artist", want.Artist, got.Artist, nullStrict))
	add(checkNullable(caseID, mbid, "album.album", want.Album, got.Album, nullStrict))
	add(checkNullable(caseID, mbid, "album.va", want.VA, got.VA, nullNeverExpected))
	add(checkNullable(caseID, mbid, "album.media", want.Media, got.Media, nullEqualsZero))
	add(checkNullable(caseID, mbid, "album.mediums", want.Mediums, got.Mediums, nullNeverExpected))
	add(checkNullable(caseID, mbid, "album.year", want.Year, got.Year, nullEqualsZero))
	add(checkNullable(caseID, mbid, "album.original_year", want.OriginalYear, got.OriginalYear, nullEqualsZero))
	add(checkNullable(caseID, mbid, "album.country", want.Country, got.Country, nullEqualsZero))
	add(checkNullable(caseID, mbid, "album.label", want.Label, got.Label, nullEqualsZero))
	add(checkNullable(caseID, mbid, "album.catalognum", want.CatalogNum, got.CatalogNum, nullEqualsZero))
	add(checkNullable(caseID, mbid, "album.albumdisambig", want.AlbumDisambig, got.AlbumDisambig, nullEqualsZero))
	add(checkNullable(caseID, mbid, "album.album_id", want.AlbumID, got.AlbumID, nullNeverExpected))
	add(checkNullable(caseID, mbid, "album.data_source", want.DataSource, got.DataSource, nullNeverExpected))

	if len(want.Tracks) != len(got.Tracks) {
		errs = append(errs, fmt.Errorf("case %s: release %s: tracks: got %d tracks, want %d", caseID, mbid, len(got.Tracks), len(want.Tracks)))
	} else {
		for i := range want.Tracks {
			if err := checkTranslatedTrack(caseID, mbid, i, want.Tracks[i], got.Tracks[i]); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// checkTranslatedTrack compares one track element by element,
// prefixing every disagreement with its index in the track list.
func checkTranslatedTrack(caseID, mbid string, i int, want analysis.LedgerTrackInfo, got matcher.TrackInfo) error {
	var errs []error
	add := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	prefix := fmt.Sprintf("tracks[%d].", i)

	add(checkNullable(caseID, mbid, prefix+"title", want.Title, got.Title, nullStrict))
	add(checkNullable(caseID, mbid, prefix+"artist", want.Artist, got.Artist, nullEqualsZero))
	add(checkNullable(caseID, mbid, prefix+"length", want.Length, got.Length, nullEqualsZero))
	add(checkNullable(caseID, mbid, prefix+"index", want.Index, got.Index, nullNeverExpected))
	add(checkNullable(caseID, mbid, prefix+"medium_index", want.MediumIndex, got.MediumIndex, nullNeverExpected))
	add(checkNullable(caseID, mbid, prefix+"medium", want.Medium, got.Medium, nullNeverExpected))
	add(checkNullable(caseID, mbid, prefix+"track_id", want.TrackID, got.TrackID, nullNeverExpected))
	add(checkNullable(caseID, mbid, prefix+"data_source", want.DataSource, got.DataSource, nullNeverExpected))

	return errors.Join(errs...)
}

// mergeProfile applies profile's overlay atop base, in Tunables'
// defaults-then-profile order (fold, B4, the same rule the spec's M1
// fold states for tier 1). It is a profile-name-to-overlay table, never
// a YAML parse: go.mod carries no YAML dependency and this plan's
// Global Constraints forbid adding one. stock and estate overlay
// nothing the translation reads; locale sets Languages and
// IgnoredAliasTypes. An unrecognized profile name is the failure.
func mergeProfile(base matcher.Tunables, profile string) (matcher.Tunables, error) {
	switch profile {
	case "stock", "estate":
		return base, nil
	case "locale":
		base.Languages = []string{"ja"}
		base.IgnoredAliasTypes = []string{}
		return base, nil
	default:
		return matcher.Tunables{}, fmt.Errorf("profile %q is not a recognized profile name (stock, estate, locale)", profile)
	}
}

// profileKeyAllowlist names every top-level and import: key the three
// committed profiles (corpus/profiles/*.yaml) use today, each with
// either its Tunables field or the reason it has none (fold, B4's
// second adaptation): dropping the unknown-key rule with the YAML
// parse would let a future match: key land in a profile without ever
// reaching the translation, which is the silent-stop the rule existed
// to catch.
var profileKeyAllowlist = map[string]string{
	"plugins":                    "no Tunables home: the metadata-source list, not a value the translation reads",
	"directory":                  "no Tunables home: a box path substituted at derivation time, not corpus data",
	"library":                    "no Tunables home: a box path substituted at derivation time, not corpus data",
	"original_date":              "no Tunables home by design (M2): read only in AlbumInfo.item_data, never on the distance path",
	"per_disc_numbering":         "no Tunables home by design (M2): read only in TrackInfo.raw_data, never on the distance path",
	"import.quiet_fallback":      "matcher.Tunables.QuietFallback",
	"import.none_rec_action":     "matcher.Tunables.NoneRecAction",
	"import.timid":               "matcher.Tunables.Timid",
	"import.languages":           "matcher.Tunables.Languages",
	"import.ignored_alias_types": "matcher.Tunables.IgnoredAliasTypes",
}

// checkProfileKeys scans contents (one profile's raw YAML text) line by
// line, in countMetadataSources' shape (internal/harness/oracle/oracle.go:254-268),
// and fails naming any top-level or import: key profileKeyAllowlist
// does not carry. It never parses YAML; it only recognizes a key line
// by its leading indentation and a colon, which is all the allowlist
// needs to catch a future key landing in a profile unnoticed.
func checkProfileKeys(name, contents string) error {
	var errs []error
	inImport := false
	for line := range strings.SplitSeq(contents, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		key, _, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		if indent == 0 {
			inImport = trimmed == "import:"
			if inImport {
				continue
			}
			if _, allowed := profileKeyAllowlist[key]; !allowed {
				errs = append(errs, fmt.Errorf("profile %s: top-level key %q is not in the allowlist", name, key))
			}
			continue
		}
		if inImport {
			full := "import." + key
			if _, allowed := profileKeyAllowlist[full]; !allowed {
				errs = append(errs, fmt.Errorf("profile %s: import key %q is not in the allowlist", name, full))
			}
		}
	}
	return errors.Join(errs...)
}

// checkProfileKeyDrift reads every profile in profiles from
// corpus/profiles under root and runs checkProfileKeys against its
// committed text.
func checkProfileKeyDrift(root string, profiles []string) error {
	var errs []error
	for _, name := range profiles {
		path := filepath.Join(root, "corpus", "profiles", name+".yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("profile %s: %w", name, err))
			continue
		}
		if err := checkProfileKeys(name, string(data)); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// findReleaseRequest returns the request in c whose path names a
// release lookup for mbid: an exact match on the lookup path, or that
// path followed by "?", since the query string carries fmt and inc
// parameters the lookup itself does not vary by.
func findReleaseRequest(c *cassette.Case, mbid string) (*cassette.Request, error) {
	prefix := releaseLookupPrefix + mbid
	for i := range c.Requests {
		req := &c.Requests[i]
		if req.PathQuery == prefix || strings.HasPrefix(req.PathQuery, prefix+"?") {
			return req, nil
		}
	}
	return nil, fmt.Errorf("case %s: no release-lookup request for %s", c.ID, mbid)
}

// loadRelease decodes the committed response body a release-lookup
// request for mbid names within c, read from caseDir. No stub, no
// client, no network: the stub is 3c's business, and reading a
// committed response file from the harness is exactly the permitted
// direction.
func loadRelease(caseDir string, c *cassette.Case, mbid string) (*musicbrainz.Release, error) {
	req, err := findReleaseRequest(c, mbid)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(caseDir, req.Response))
	if err != nil {
		return nil, fmt.Errorf("case %s: read response %s: %w", c.ID, req.Response, err)
	}
	var rel musicbrainz.Release
	if err := json.Unmarshal(data, &rel); err != nil {
		return nil, fmt.Errorf("case %s: decode response %s: %w", c.ID, req.Response, err)
	}
	return &rel, nil
}

// gradeLedger rebuilds every row in l through the Go translation and
// compares it against l's own recorded AlbumInfo, resolving each row's
// case and release response under root. It returns how many rows it
// actually compared, not merely visited, so a caller that asserts
// compared == len(l.Rows) catches a row a broken resolver silently
// skipped rather than reported. A row carrying a Merged block also
// grades through musicbrainz.MergePseudoRelease against the actual
// release it targets, the only place that merge is graded against
// beets rather than a hand-authored fixture.
func gradeLedger(root string, l *analysis.AlbumInfoLedger) (compared int, err error) {
	var errs []error
	for _, row := range l.Rows {
		dir := filepath.Join(root, "corpus", "cases", row.CaseID)
		c, loadErr := cassette.LoadCase(dir)
		if loadErr != nil {
			errs = append(errs, fmt.Errorf("row %s/%s: load case: %w", row.CaseID, row.ReleaseMBID, loadErr))
			continue
		}

		rel, relErr := loadRelease(dir, c, row.ReleaseMBID)
		if relErr != nil {
			errs = append(errs, fmt.Errorf("row %s/%s: %w", row.CaseID, row.ReleaseMBID, relErr))
			continue
		}

		tun, tunErr := mergeProfile(matcher.DefaultTunables(), c.Profile)
		if tunErr != nil {
			errs = append(errs, fmt.Errorf("row %s/%s: %w", row.CaseID, row.ReleaseMBID, tunErr))
			continue
		}

		got, transErr := musicbrainz.TranslateAlbum(rel, tun)
		if transErr != nil {
			errs = append(errs, fmt.Errorf("row %s/%s: translate: %w", row.CaseID, row.ReleaseMBID, transErr))
			continue
		}

		if err := CheckTranslatedAlbum(row.CaseID, row.ReleaseMBID, row.Album, got); err != nil {
			errs = append(errs, err)
		}
		compared++

		if row.Merged == nil {
			continue
		}

		actualID, ok := musicbrainz.PseudoReleaseTarget(rel)
		if !ok {
			errs = append(errs, fmt.Errorf("row %s/%s: carries a Merged block but PseudoReleaseTarget reports it is not a pseudo-release", row.CaseID, row.ReleaseMBID))
			continue
		}
		actualRel, actualErr := loadRelease(dir, c, actualID)
		if actualErr != nil {
			errs = append(errs, fmt.Errorf("row %s/%s: merged target %s: %w", row.CaseID, row.ReleaseMBID, actualID, actualErr))
			continue
		}
		actualInfo, actualTransErr := musicbrainz.TranslateAlbum(actualRel, tun)
		if actualTransErr != nil {
			errs = append(errs, fmt.Errorf("row %s/%s: merged target %s: translate: %w", row.CaseID, row.ReleaseMBID, actualID, actualTransErr))
			continue
		}
		merged := musicbrainz.MergePseudoRelease(got, actualInfo)
		if err := CheckTranslatedAlbum(row.CaseID, row.ReleaseMBID+" (merged)", *row.Merged, merged); err != nil {
			errs = append(errs, err)
		}
	}
	return compared, errors.Join(errs...)
}

// checkNoReleaseReachesBrowseThreshold reads every committed corpus
// case's release-lookup responses under root, sums len(medium.tracks)
// across a release's own media, and fails naming the case, the
// release, and the count when any release reaches threshold. This is
// the one check in the pass that reads outside the fifty ledger rows,
// because a corpus case crossing BROWSE_MAXTRACKS is exactly the event
// nobody would otherwise notice.
func checkNoReleaseReachesBrowseThreshold(root string, threshold int) error {
	casesRoot := filepath.Join(root, "corpus", "cases")
	entries, err := os.ReadDir(casesRoot)
	if err != nil {
		return fmt.Errorf("read corpus/cases: %w", err)
	}

	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(casesRoot, e.Name())
		c, err := cassette.LoadCase(dir)
		if err != nil {
			errs = append(errs, fmt.Errorf("case %s: %w", e.Name(), err))
			continue
		}
		for _, req := range c.Requests {
			if !strings.HasPrefix(req.PathQuery, releaseLookupPrefix) {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, req.Response))
			if err != nil {
				errs = append(errs, fmt.Errorf("case %s: read response %s: %w", e.Name(), req.Response, err))
				continue
			}
			var rel struct {
				ID    string `json:"id"`
				Media []struct {
					Tracks []json.RawMessage `json:"tracks"`
				} `json:"media"`
			}
			if err := json.Unmarshal(data, &rel); err != nil {
				errs = append(errs, fmt.Errorf("case %s: decode response %s: %w", e.Name(), req.Response, err))
				continue
			}
			count := 0
			for _, m := range rel.Media {
				count += len(m.Tracks)
			}
			if count >= threshold {
				errs = append(errs, fmt.Errorf("case %s: release %s: %d tracks reaches BROWSE_MAXTRACKS threshold %d", e.Name(), rel.ID, count, threshold))
			}
		}
	}
	return errors.Join(errs...)
}
