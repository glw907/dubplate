package musicbrainz

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// extraTags is FIELDS_TO_MB_KEYS (beetsplug/musicbrainz.py:47-56) joined
// to get_album_criteria's per-tag value rule (:721-732): each item field
// musicbrainz.extra_tags may name, the MusicBrainz search field it
// becomes, and the value it carries. A configured tag outside this map
// contributes no criterion (:697-712).
var extraTags = map[string]struct {
	field string
	value func(likelies matcher.Likelies, itemCount int) string
}{
	"barcode": {"barcode", func(l matcher.Likelies, _ int) string { return l.Barcode }},
	"catalognum": {"catno", func(l matcher.Likelies, _ int) string {
		return strings.ReplaceAll(l.CatalogNum, " ", "")
	}},
	"country": {"country", func(l matcher.Likelies, _ int) string { return l.Country }},
	"label":   {"label", func(l matcher.Likelies, _ int) string { return l.Label }},
	"media":   {"format", func(l matcher.Likelies, _ int) string { return l.Media }},
	// An unset year renders as the string "0", which survives the
	// empty-value filter and ships as date:(0), narrowing the search to
	// nothing. beets does the same through str(most_common) over an int
	// field, and the port reproduces it rather than fixing it.
	"year":   {"date", func(l matcher.Likelies, _ int) string { return strconv.Itoa(l.Year) }},
	"tracks": {"tracks", func(_ matcher.Likelies, itemCount int) string { return strconv.Itoa(itemCount) }},
	"alias":  {"alias", func(l matcher.Likelies, _ int) string { return l.Album }},
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
// (:679-681), in that order. It derives the search terms from items, as
// tag_album does when no caller passes search_artist or search_name
// (autotag/match.py:443-446).
//
// beets lowers and strips each value twice, in
// get_search_query_with_filters' comprehension (:750-752) and again in
// format_search_term (:655). The two passes are idempotent, so
// formatSearchTerm runs the one.
func AlbumSearchFor(items []matcher.ItemFields, tun matcher.Tunables) AlbumSearch {
	likelies, consensus := matcher.MostCommonTags(items)
	criteria := albumCriteria(likelies, len(items), vaLikely(likelies, consensus, items), tun)

	var terms []string
	for _, c := range criteria {
		if term := formatSearchTerm(c.Field, c.Value); term != "" {
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
// The walk iterates tun.ExtraTags and looks the tag up in extraTags,
// never the reverse, because beets walks its own configured sequence
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
		extra, ok := extraTags[tag]
		if !ok || seen[tag] {
			continue
		}
		seen[tag] = true
		criteria = append(criteria, criterion{Field: extra.field, Value: extra.value(likelies, itemCount)})
	}
	return criteria
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
	return field + ":(" + escapeLucene(term) + ")"
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
// CPython disagree on a small set of code points
// (internal/engine/matcher/testdata/case-mapping-sweep.tsv), and the
// query path has nothing after it to rescue them, so they reach the
// wire. U+0130 is the stable member, the full-versus-simple case-mapping
// difference: Go lowers it to a single "i" and Python to "i" plus
// U+0307. The rest is a Unicode version skew between the two toolchains
// and moves when either does.
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
