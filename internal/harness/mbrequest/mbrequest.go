// Package mbrequest builds the canonical MusicBrainz WS/2 request query
// that a release lookup always issues, live-captured or synthetic: the
// maximal inc= set beets itself requests (mechanics study Q5), and the
// query-canonicalization rule cassette's [cassette.Request.PathQuery]
// requires (corpus/README.md). It is the one home for logic that
// cmd/harness's capture command and internal/harness/synth both need but
// neither can import from the other, and that cassette cannot expose
// without reshaping its frozen surface.
package mbrequest

import (
	"net/url"
	"slices"
	"strings"
)

// MaximalReleaseInclude is the release lookup's inc= parameter,
// mirroring beets' own RELEASE_INCLUDES (mechanics study Q5) so a
// captured or synthetic case never forces a re-capture the first time a
// future matcher improvement wants a field this run already fetched
// (U6).
const MaximalReleaseInclude = "artists+media+recordings+release-groups+labels+" +
	"artist-credits+aliases+recording-level-rels+work-rels+work-level-rels+" +
	"artist-rels+isrcs+url-rels+release-rels+genres+tags"

// CanonicalizeQuery joins path with params sorted byte-wise by their
// "key=value" pair, matching cassette's own canonicalization rule
// (corpus/README.md): each value is percent-encoded with
// url.QueryEscape before joining, the same encoding Python's requests
// library applies when beets' MusicBrainz client builds its own query
// string (a bare "+" joining inc= values becomes "%2B" on the wire,
// never a literal "+"; a request that stored the literal form would
// never match what pinned beets actually requests during derivation).
func CanonicalizeQuery(path string, params map[string]string) string {
	pairs := make([]string, 0, len(params))
	for k, v := range params {
		pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	slices.Sort(pairs)
	return path + "?" + strings.Join(pairs, "&")
}
