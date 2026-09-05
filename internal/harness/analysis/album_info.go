package analysis

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// AlbumInfoLedgerRowFloor is the row count corpus/analysis/album_info.json
// must not fall below: one row per distinct release lookup across the
// nineteen named AlbumInfoSubjects, measured off the tree at 2026-09-04
// by counting the distinct /ws/2/release/<mbid> requests in the
// nineteen cases' own logs (the derivation dedupes by mbid, and no
// committed subject repeats a lookup today, so the two counts agree).
// A shrink here means the derivation quietly dropped a lookup
// rather than an edit anyone would notice in review.
const AlbumInfoLedgerRowFloor = 50

// releaseLookupPrefix is the WS/2 path a release lookup by id starts
// with. The search endpoint is /ws/2/release? and never matches it, so
// the prefix alone separates the two kinds of request a case logs.
const releaseLookupPrefix = "/ws/2/release/"

// AlbumInfoSubject names one committed case the AlbumInfo ledger derives
// rows for. Every distinct release lookup in the case's own log becomes
// a row; a case is a subject because a named translation-layer branch runs
// through it, and Why carries that reason beside the id rather than in a
// plan nobody reads at execution time. CheckAlbumInfoLedger requires Why
// to be non-empty, so a subject added without a rationale fails the
// gate. FileTagSubject puts its rationale in the slice's own doc
// comment; this type diverges on purpose, with a reader, rather than
// carrying a field nothing reads.
type AlbumInfoSubject struct {
	CaseID string
	Why    string
}

// AlbumInfoPseudoSubjects names the committed cases whose release set
// contains a pseudo-release, measured off the tree at 2026-09-04:
// exactly four case directories carry a relation with type
// "transl-tracklisting" and direction "backward"
// (beetsplug/musicbrainz.py:785-798). It exists so CheckAlbumInfoLedger's
// Merged check compares against a named expectation instead of deriving
// one from the rows it is checking, which would pass on a ledger
// carrying no Merged block at all.
var AlbumInfoPseudoSubjects = []string{
	"non-latin-pseudo-release",
	"real-nonlatin-girls-generation",
	"real-nonlatin-zemfira",
	"real-nonlatin-zemfira-stock",
}

// AlbumInfoSubjects is the fixed nineteen-case subject set. Every
// distinct release lookup in each case's own log becomes a row, which
// is 50 rows at 2026-09-04.
var AlbumInfoSubjects = []AlbumInfoSubject{
	{"clean-match", "the plain baseline: single medium, single credit, no alias, one release event taken through the fallback"},
	{"media-baseline", "mediums counts unfiltered media while extraction filters them (musicbrainz.py:659 against :638-640)"},
	{"media-uneven-split", "index assignment 1-based across all valid media (musicbrainz.py:645-646)"},
	{"va-exact", "the various-artists override and the va_name read (musicbrainz.py:673-681)"},
	{"various-artists", "the VA override beside the corpus's one release whose two events carry different dates"},
	{"should-reject", "search fallback: six release lookups in one case, plus multiple release events"},
	{"non-latin-pseudo-release", "pseudo-release merge; six fields move"},
	{"real-nonlatin-zemfira", "pseudo-release merge; three fields move"},
	{"real-nonlatin-zemfira-stock", "the same merge under the stock profile"},
	{"real-nonlatin-girls-generation", "pseudo-release merge, an oracle-wrong alias probe, and three video recordings at once"},
	{"real-nonlatin-anna-vissi-essential", "oracle-wrong alias probe (adj-adv-alias-vissi-native-greek)"},
	{"real-nonlatin-alla-pugacheva-vol1", "oracle-correct alias probe: entity name and release credit coincide"},
	{"real-nonlatin-x-japan-art-of-life", "oracle-correct alias probe on a single-track release"},
	{"real-nonlatin-utada-first-love", "bug-locale-alias-ignored's A half, under the estate profile"},
	{"real-nonlatin-utada-first-love-locale", "the B half, and the corpus's only case running a non-empty import.languages"},
	{"real-nonlatin-hamasaki-ayumi", "138 video recordings: ignore_video_tracks empties a medium and _add_candidate drops the candidate"},
	{"real-genre-metallica-master-of-puppets", "one release event whose date equals the top-level date, so the fallback supplies country and date"},
	{"real-compilation-queen-greatest-hits-ii", "two release events, both equal to the top-level date: the same fallback under a second stratum"},
	{"real-noise-dylan-bootleg", "the mixed-format route to the \"Media\" sentinel: two Hybrid SACD layers on one release (musicbrainz.py:654-658)"},
}

// AlbumInfoLedger is corpus/analysis/album_info.json: the translated
// AlbumInfo every candidate the matcher scores is built from, dumped
// through pinned beets so the Go translation layer is graded against
// beets' own answer field by field rather than only through whole-case
// agreement (ruling E).
type AlbumInfoLedger struct {
	ShimVersion  string               `json:"shim_version"`
	BeetsVersion string               `json:"beets_version"`
	Rows         []AlbumInfoLedgerRow `json:"rows"`
}

// AlbumInfoLedgerRow is one release lookup's translation. Merged is
// non-nil only on a pseudo-release whose transl-tracklisting merge
// fired, and then Album carries the pre-merge value, so the merge is
// gradeable on its own rather than only through its result.
type AlbumInfoLedgerRow struct {
	CaseID      string           `json:"case_id"`
	ReleaseMBID string           `json:"release_mbid"`
	Album       LedgerAlbumInfo  `json:"album"`
	Merged      *LedgerAlbumInfo `json:"merged,omitempty"`
}

// LedgerAlbumInfo mirrors matcher.AlbumInfo field for field. Every
// scalar is a pointer so a beets None survives the round trip as a JSON
// null and is distinguishable from the Go zero value: this is the JSON
// boundary case go-conventions carves out, not defensive nil checking,
// and internal/harness/conform's nullability policy is what reads the
// distinction.
type LedgerAlbumInfo struct {
	Artist        *string           `json:"artist"`
	Album         *string           `json:"album"`
	VA            *bool             `json:"va"`
	Media         *string           `json:"media"`
	Mediums       *int              `json:"mediums"`
	Year          *int              `json:"year"`
	OriginalYear  *int              `json:"original_year"`
	Country       *string           `json:"country"`
	Label         *string           `json:"label"`
	CatalogNum    *string           `json:"catalognum"`
	AlbumDisambig *string           `json:"albumdisambig"`
	AlbumID       *string           `json:"album_id"`
	DataSource    *string           `json:"data_source"`
	Tracks        []LedgerTrackInfo `json:"tracks"`
}

// LedgerTrackInfo mirrors matcher.TrackInfo field for field, on
// LedgerAlbumInfo's own pointer convention.
type LedgerTrackInfo struct {
	Title       *string  `json:"title"`
	Artist      *string  `json:"artist"`
	Length      *float64 `json:"length"`
	Index       *int     `json:"index"`
	MediumIndex *int     `json:"medium_index"`
	Medium      *int     `json:"medium"`
	TrackID     *string  `json:"track_id"`
	DataSource  *string  `json:"data_source"`
}

// LoadAlbumInfoLedger reads corpus/analysis/album_info.json from path.
func LoadAlbumInfoLedger(path string) (*AlbumInfoLedger, error) {
	var l AlbumInfoLedger
	if err := readJSON(path, &l); err != nil {
		return nil, fmt.Errorf("album info ledger: %w", err)
	}
	return &l, nil
}

// CheckAlbumInfoLedger validates the ledger's schema and its coverage
// against the committed corpus, reporting every failure rather than the
// first. It checks six things: the two header fields are present; every
// row resolves to a committed case; the row set is exactly
// AlbumInfoSubjects in both directions, each carrying a rationale; every
// row's release mbid is a release lookup in that case's own request log
// and every such lookup has exactly one row; the row-count floor holds;
// and a merged block sits on exactly the cases named in
// AlbumInfoPseudoSubjects, one row each. cases is every committed corpus
// case, keyed by id, so a row naming an id outside the subject set is
// still recognized as a real (but wrong) corpus case rather than
// reported as unresolvable.
func CheckAlbumInfoLedger(cases map[string]*cassette.Case, l *AlbumInfoLedger) error {
	var errs []error

	if l.ShimVersion == "" {
		errs = append(errs, errors.New("album info ledger: shim_version: missing"))
	}
	if l.BeetsVersion == "" {
		errs = append(errs, errors.New("album info ledger: beets_version: missing"))
	}
	if len(l.Rows) < AlbumInfoLedgerRowFloor {
		errs = append(errs, fmt.Errorf("album info ledger: %d rows, want at least %d", len(l.Rows), AlbumInfoLedgerRowFloor))
	}

	wantSubjects := make(map[string]bool, len(AlbumInfoSubjects))
	for _, s := range AlbumInfoSubjects {
		if s.Why == "" {
			errs = append(errs, fmt.Errorf("album info ledger: subject %q carries no rationale", s.CaseID))
		}
		wantSubjects[s.CaseID] = true
	}

	// rowsByCase counts each case's rows per release mbid, so check 4
	// can report both a row naming a lookup the case never made and a
	// lookup no row covers, and check 6 can insist a pseudo case's
	// merged block sits on exactly one of its rows.
	rowsByCase := make(map[string]map[string]int, len(l.Rows))
	mergedByCase := make(map[string]int, len(l.Rows))

	for i, row := range l.Rows {
		c, ok := cases[row.CaseID]
		if !ok {
			errs = append(errs, fmt.Errorf("album info ledger: rows[%d]: case %q is not a committed corpus case", i, row.CaseID))
			continue
		}
		if !wantSubjects[row.CaseID] {
			errs = append(errs, fmt.Errorf("album info ledger: %s: not one of the nineteen named AlbumInfo subjects", row.CaseID))
			continue
		}
		if rowsByCase[row.CaseID] == nil {
			rowsByCase[row.CaseID] = make(map[string]int)
		}
		rowsByCase[row.CaseID][row.ReleaseMBID]++
		if row.Merged != nil {
			mergedByCase[row.CaseID]++
		}
		if !slices.Contains(releaseLookups(c), row.ReleaseMBID) {
			errs = append(errs, fmt.Errorf("album info ledger: %s: release %q is not a release lookup in the case's own request log", row.CaseID, row.ReleaseMBID))
		}
	}

	for _, s := range AlbumInfoSubjects {
		c, ok := cases[s.CaseID]
		if !ok {
			errs = append(errs, fmt.Errorf("album info ledger: subject %q is not a committed corpus case", s.CaseID))
			continue
		}
		rows := rowsByCase[s.CaseID]
		if len(rows) == 0 {
			errs = append(errs, fmt.Errorf("album info ledger: missing rows for subject %q", s.CaseID))
			continue
		}
		for _, mbid := range releaseLookups(c) {
			switch rows[mbid] {
			case 1:
			case 0:
				errs = append(errs, fmt.Errorf("album info ledger: %s: release lookup %q has no row; the derivation skipped it", s.CaseID, mbid))
			default:
				errs = append(errs, fmt.Errorf("album info ledger: %s: release lookup %q has %d rows, want exactly 1", s.CaseID, mbid, rows[mbid]))
			}
		}
	}

	for _, s := range AlbumInfoSubjects {
		merged := mergedByCase[s.CaseID]
		if slices.Contains(AlbumInfoPseudoSubjects, s.CaseID) {
			if merged != 1 {
				errs = append(errs, fmt.Errorf("album info ledger: %s: %d rows carry a merged block, want exactly 1: the case is named in AlbumInfoPseudoSubjects", s.CaseID, merged))
			}
			continue
		}
		if merged != 0 {
			errs = append(errs, fmt.Errorf("album info ledger: %s: %d rows carry a merged block, want none: the case is not named in AlbumInfoPseudoSubjects", s.CaseID, merged))
		}
	}

	return errors.Join(errs...)
}

// releaseLookups returns the release mbids c's own request log fetches
// by id, in log order. A search request's path is /ws/2/release? and
// never carries the trailing slash, so the prefix alone is the filter.
func releaseLookups(c *cassette.Case) []string {
	var mbids []string
	for _, req := range c.Requests {
		rest, ok := strings.CutPrefix(req.PathQuery, releaseLookupPrefix)
		if !ok {
			continue
		}
		mbid, _, _ := strings.Cut(rest, "?")
		if mbid != "" {
			mbids = append(mbids, mbid)
		}
	}
	return mbids
}
