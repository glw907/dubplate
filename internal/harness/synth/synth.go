package synth

import (
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// maximalReleaseInclude mirrors cmd/harness/capture.go's own constant
// (beets' RELEASE_INCLUDES, mechanics study Q5): every ID-lookup case
// pinned beets issues, live-captured or synthetic, requests the same
// maximal inc= set, so a synthetic case never forces a re-capture the
// first time a future matcher improvement wants a field this run
// already carries (U6). Duplicated here for the same reason
// cmd/harness's own copy is duplicated from cassette: the two tools
// have no shared internal package to hold it without giving either one
// an import the plan does not ask for.
const maximalReleaseInclude = "artists+media+recordings+release-groups+labels+" +
	"artist-credits+aliases+recording-level-rels+work-rels+work-level-rels+" +
	"artist-rels+isrcs+url-rels+release-rels+genres+tags"

// Build derives a complete case from spec: the pinned item set (from
// spec.Media plus spec.Overrides, or spec.Items verbatim), the
// canonical ID-lookup request log entry, and the WS/2 release document
// that entry names. id becomes the case's own [cassette.Case.ID], the
// same convention `harness capture --out <dir>` already uses
// (filepath.Base of the output directory).
//
// The returned release document is the exact bytes a case's
// responses/001.json should hold; Build itself never touches disk,
// so a caller (the harness synth command, or a test) decides where
// and whether to write anything.
func Build(id string, spec *ReleaseSpec) (*cassette.Case, []byte, error) {
	if err := spec.validate(); err != nil {
		return nil, nil, err
	}

	items, err := buildItems(spec)
	if err != nil {
		return nil, nil, err
	}

	rel := buildRelease(spec)
	doc, err := marshalRelease(rel)
	if err != nil {
		return nil, nil, err
	}

	pathQuery := canonicalizeQuery("/ws/2/release/"+spec.MBAlbumID, map[string]string{
		"fmt": "json",
		"inc": maximalReleaseInclude,
	})

	c := &cassette.Case{
		ID:       id,
		Input:    cassette.Input{Items: items},
		Profile:  spec.Profile,
		EvalDate: spec.EvalDate,
		Requests: []cassette.Request{
			{Method: http.MethodGet, PathQuery: pathQuery, Response: "responses/001.json"},
		},
	}
	return c, doc, nil
}

// buildItems returns spec's pinned item set: spec.Items verbatim when
// set, or the release's own tracks (one item per track, release order)
// with spec.Overrides applied.
func buildItems(spec *ReleaseSpec) ([]cassette.ItemFields, error) {
	if len(spec.Items) > 0 {
		return spec.Items, nil
	}

	multiDisc := len(spec.Media) > 1
	var items []cassette.ItemFields
	for mi, m := range spec.Media {
		for _, tr := range m.Tracks {
			artist := spec.Artist
			if tr.ArtistCredit != "" {
				artist = tr.ArtistCredit
			}
			item := cassette.ItemFields{
				"album":       spec.Album,
				"albumartist": spec.Artist,
				"artist":      artist,
				"title":       tr.Title,
				"length":      tr.Length,
				"track":       tr.Position,
				"year":        releaseYear(spec.Date),
				"mb_albumid":  spec.MBAlbumID,
			}
			if spec.VariousArtists {
				item["comp"] = true
			}
			if multiDisc {
				item["disc"] = mi + 1
			}
			items = append(items, item)
		}
	}

	return applyOverrides(items, spec.Overrides, spec.MBAlbumID, spec.AllowAlbumIDOverride)
}

// releaseYear returns the calendar year from an "YYYY-MM-DD" (or
// "YYYY") release date, or 0 if date does not start with a parseable
// year: a malformed spec date is [ReleaseSpec.validate]'s job to
// reject elsewhere, not this helper's.
func releaseYear(date string) int {
	year, _, _ := strings.Cut(date, "-")
	n, err := strconv.Atoi(year)
	if err != nil {
		return 0
	}
	return n
}

// applyOverrides perturbs items per overrides: OpField merges Fields
// onto the item at Index, OpMissing drops it, and OpExtra appends a
// brand-new item carrying Fields plus the release's own albumID
// (never the override's own, so "an extra item carries the album's
// mb_albumid like every other item" holds without the author having
// to restate it). Every op returns an error naming the offending
// override rather than silently clamping an out-of-range Index.
func applyOverrides(items []cassette.ItemFields, overrides []Override, albumID string, allowAlbumIDOverride bool) ([]cassette.ItemFields, error) {
	kept := make([]bool, len(items))
	for i := range kept {
		kept[i] = true
	}
	merges := make(map[int]cassette.ItemFields, len(overrides))
	var extras []cassette.ItemFields

	for i, ov := range overrides {
		if !allowAlbumIDOverride {
			if _, ok := ov.Fields[mbAlbumIDField]; ok {
				return nil, fmt.Errorf("synth: overrides[%d]: %s is fenced off the override list outside family 10; set allow_album_id_override to lift it", i, mbAlbumIDField)
			}
		}

		switch ov.Op {
		case OpField:
			if ov.Index < 0 || ov.Index >= len(items) {
				return nil, fmt.Errorf("synth: overrides[%d]: index %d out of range for %d derived items", i, ov.Index, len(items))
			}
			merges[ov.Index] = ov.Fields
		case OpMissing:
			if ov.Index < 0 || ov.Index >= len(items) {
				return nil, fmt.Errorf("synth: overrides[%d]: index %d out of range for %d derived items", i, ov.Index, len(items))
			}
			kept[ov.Index] = false
		case OpExtra:
			extra := maps.Clone(ov.Fields)
			if extra == nil {
				extra = cassette.ItemFields{}
			}
			extra[mbAlbumIDField] = albumID
			extras = append(extras, extra)
		default:
			return nil, fmt.Errorf("synth: overrides[%d]: unknown op %q", i, ov.Op)
		}
	}

	result := make([]cassette.ItemFields, 0, len(items)+len(extras))
	for i, item := range items {
		if !kept[i] {
			continue
		}
		if fields, ok := merges[i]; ok {
			merged := maps.Clone(item)
			maps.Copy(merged, fields)
			item = merged
		}
		result = append(result, item)
	}
	result = append(result, extras...)
	return result, nil
}

// canonicalizeQuery joins path with params sorted byte-wise by their
// "key=value" pair, matching the cassette package's own
// canonicalization rule (corpus/README.md): each value is percent-
// encoded with url.QueryEscape before joining, the same encoding
// Python's requests library applies when beets' MusicBrainz client
// builds its own query string (a bare "+" joining inc= values becomes
// "%2B" on the wire, never a literal "+"). Duplicated from
// cmd/harness/capture.go's identical helper because cassette ships
// frozen and keeps its own canonicalization unexported (Task 1); the
// round-trip test proves this copy produces exactly what
// cassette.LoadCase accepts.
func canonicalizeQuery(path string, params map[string]string) string {
	pairs := make([]string, 0, len(params))
	for k, v := range params {
		pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	slices.Sort(pairs)
	return path + "?" + strings.Join(pairs, "&")
}
