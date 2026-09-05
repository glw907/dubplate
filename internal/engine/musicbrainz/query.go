package musicbrainz

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// releaseIncludes is beets' RELEASE_INCLUDES
// (beetsplug/_utils/musicbrainz.py:50-67), in its own order, joined with
// a literal "+" at request time (:634-635).
//
// This is deliberately a SECOND definition of a string
// internal/harness/mbrequest already carries (mbrequest.go:22-24), and it
// must stay one. mbrequest is what the capture tool and the synth tool
// use to PRODUCE a committed request; the committed request log is the
// expectation and this list is the system under test. One shared
// definition would make internal/harness/conform's request grading a
// self-derived expectation, which is the anti-pattern the conformance run
// exists to catch. The two copies are independent on purpose and the
// conformance run is what proves they agree.
var releaseIncludes = []string{
	"artists", "media", "recordings", "release-groups", "labels",
	"artist-credits", "aliases", "recording-level-rels", "work-rels",
	"work-level-rels", "artist-rels", "isrcs", "url-rels",
	"release-rels", "genres", "tags",
}

// fieldsToMBKeys is FIELDS_TO_MB_KEYS (beetsplug/musicbrainz.py:47-56):
// the eight item fields musicbrainz.extra_tags may name, each mapped to
// the MusicBrainz search field it becomes. A configured tag outside this
// map contributes no criterion (:697-712).
var fieldsToMBKeys = map[string]string{
	"barcode":    "barcode",
	"catalognum": "catno",
	"country":    "country",
	"label":      "label",
	"media":      "format",
	"year":       "date",
	"tracks":     "tracks",
	"alias":      "alias",
}

// luceneSpecial is LUCENE_SPECIAL_CHAR_PAT's character class
// (beetsplug/_utils/musicbrainz.py:48): the nineteen characters
// format_search_term backslash-escapes. The apostrophe, comma, period,
// semicolon, angle brackets, equals sign, hash, percent, at sign and
// dollar sign are absent from it and therefore reach the wire unescaped.
const luceneSpecial = `-+&|!(){}[]^"~*?:\/`

// AlbumSearch is one album search request, built and not yet issued.
type AlbumSearch struct {
	// Query is the rendered Lucene query, escaped and wrapped. It is
	// NOT percent-encoded; the client encodes it once, at request time.
	Query string
	Limit int
}

// AlbumSearchFor builds the one album search request tag_album issues for
// items, mirroring get_album_criteria (beetsplug/musicbrainz.py:714-734),
// get_search_query_with_filters (:736-753), format_search_term
// (beetsplug/_utils/musicbrainz.py:651-663) and search's own join
// (:679-681), in that order.
//
// It derives the search terms the way tag_album does rather than taking
// them: beets lets a caller pass explicit search_artist and search_name
// (autotag/match.py:443-446), and no rung 3 caller has any, so modelling
// the override now would be a dead parameter. 3d's flow or a later
// operator surface is where an explicit term would arrive.
//
// Each value is lowered and stripped twice, once in
// get_search_query_with_filters' own comprehension (:750-752) and once in
// format_search_term (:655), and both passes are reproduced here rather
// than collapsed as idempotent: a later beets change to one and not the
// other is exactly the divergence this port must be able to see.
func AlbumSearchFor(items []matcher.ItemFields, tun matcher.Tunables) AlbumSearch {
	likelies, consensus := matcher.MostCommonTags(items)
	criteria := albumCriteria(likelies, len(items), vaLikely(likelies, consensus, items), tun)

	var terms []string
	for _, c := range criteria {
		value := trimSearchTerm(lowerSearchTerm(c.Value))
		if value == "" {
			continue
		}
		if term := formatSearchTerm(c.Field, value); term != "" {
			terms = append(terms, term)
		}
	}
	return AlbumSearch{Query: strings.Join(terms, " "), Limit: tun.SearchLimit}
}

// vaLikely reports whether tag_album would treat items as a
// various-artists release: no artist consensus, OR the likely artist
// naming a various-artists placeholder, OR any item flagged comp
// (autotag/match.py:451-455). consensus is MostCommonTags' second return
// value; the likely artist is likelies.Artist, which MostCommonTags may
// already have overwritten with the album artist while leaving
// consensus["artist"] computed from the raw per-item values
// (util/__init__.py:842-844).
//
// It is unexported because AlbumSearchFor is this package's only caller
// and no consumer outside it exists at rung 3; 3d's flow exports it the
// day match_by_id's no-consensus branch needs it.
func vaLikely(likelies matcher.Likelies, consensus map[string]bool, items []matcher.ItemFields) bool {
	if !consensus["artist"] || matcher.IsVariousArtistsName(likelies.Artist) {
		return true
	}
	return slices.ContainsFunc(items, func(i matcher.ItemFields) bool { return i.Comp })
}

// criterion is one entry of get_album_criteria's dict
// (beetsplug/musicbrainz.py:714-734), kept as an ordered slice because
// the dict's insertion order is the query's term order and a Go map has
// none.
type criterion struct {
	Field string
	Value string
}

// albumCriteria is get_album_criteria (beetsplug/musicbrainz.py:714-734).
// The release name leads, the various-artists id or the artist follows,
// and one entry per configured extra tag closes, in the operator's own
// musicbrainz.extra_tags order.
//
// The walk iterates tun.ExtraTags and looks the MusicBrainz key up, never
// the reverse, because beets walks its own configured sequence
// (:697-712) and a Go map walk has no order. A tag repeated in ExtraTags
// yields one criterion, as beets' tag-keyed dict does.
//
// The six plurality-valued tags read likelies rather than recomputing a
// plurality: beets computes them with util.plurality over the same items
// (:727, util/__init__.py:799-807) that MostCommonTags already
// reproduces, so a second pass here would be a second implementation of
// one rule.
func albumCriteria(likelies matcher.Likelies, itemCount int, isVA bool, tun matcher.Tunables) []criterion {
	criteria := []criterion{{Field: "release", Value: likelies.Album}}
	if isVA {
		criteria = append(criteria, criterion{Field: "arid", Value: variousArtistsID})
	} else {
		criteria = append(criteria, criterion{Field: "artist", Value: likelies.Artist})
	}

	seen := make(map[string]bool, len(tun.ExtraTags))
	for _, tag := range tun.ExtraTags {
		mbField, ok := fieldsToMBKeys[tag]
		if !ok || seen[tag] {
			continue
		}
		seen[tag] = true
		criteria = append(criteria, criterion{Field: mbField, Value: extraTagValue(tag, likelies, itemCount)})
	}
	return criteria
}

// extraTagValue is get_album_criteria's per-tag value rule
// (beetsplug/musicbrainz.py:721-732). Callers gate tag on fieldsToMBKeys,
// which is why an unknown tag returns the empty string rather than
// panicking: an empty value drops its criterion downstream anyway.
func extraTagValue(tag string, likelies matcher.Likelies, itemCount int) string {
	switch tag {
	case "tracks":
		return strconv.Itoa(itemCount)
	case "alias":
		return likelies.Album
	case "barcode":
		return likelies.Barcode
	case "catalognum":
		return strings.ReplaceAll(likelies.CatalogNum, " ", "")
	case "country":
		return likelies.Country
	case "label":
		return likelies.Label
	case "media":
		return likelies.Media
	case "year":
		// An unset year renders as the string "0", which survives the
		// empty-value filter and ships as date:(0), narrowing the search
		// to nothing. beets does the same through str(most_common) over
		// an int field, and the port reproduces it rather than fixing it.
		return strconv.Itoa(likelies.Year)
	default:
		return ""
	}
}

// formatSearchTerm is format_search_term
// (beetsplug/_utils/musicbrainz.py:651-663): lower, strip, drop when
// empty, escape, then wrap. The wrapper is applied after the escape, so
// the wrapping colon and parentheses are never escaped while a colon or a
// parenthesis inside term is.
func formatSearchTerm(field, term string) string {
	term = trimSearchTerm(lowerSearchTerm(term))
	if term == "" {
		return ""
	}
	term = escapeLucene(term)
	if field == "" {
		return term
	}
	return field + ":(" + term + ")"
}

// escapeLucene prefixes each of LUCENE_SPECIAL_CHAR_PAT's nineteen
// characters with one backslash (beetsplug/_utils/musicbrainz.py:48,
// :659).
func escapeLucene(term string) string {
	var b strings.Builder
	b.Grow(len(term))
	for _, r := range term {
		if strings.ContainsRune(luceneSpecial, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// lowerSearchTerm is Python's str.lower over a search term. Go and
// CPython disagree on 56 of the 1,112,064 valid code points
// (internal/engine/matcher/testdata/case-mapping-sweep.tsv), and the
// query path has no post-transliteration [a-z0-9] strip to rescue any of
// them, so all 56 reach the wire.
//
// The 56 are two phenomena, and only one of them is stable. U+0130 is the
// full-versus-simple case-mapping difference, which no toolchain release
// will move: Go lowers it to a single "i" and Python to "i" plus U+0307.
// The other 55 are a Unicode version skew, Go 1.27's 17.0.0 case tables
// against CPython 3.13's 15.1.0, and they move the day either toolchain
// moves. The asserted fixture row is therefore U+0130 alone.
//
// Adopting golang.org/x/text/cases would close the divergence and is
// rejected: it adds a module dependency, and the divergence changes which
// candidates MusicBrainz returns rather than any distance the conformance
// gate grades.
func lowerSearchTerm(s string) string {
	return strings.ToLower(s)
}

// trimSearchTerm is Python's str.strip over a search term. Python treats
// U+001C through U+001F as whitespace and Go's unicode.IsSpace does not,
// so those four code points are named here; every other member of the two
// sets agrees.
func trimSearchTerm(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f)
	})
}
