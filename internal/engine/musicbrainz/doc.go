// Package musicbrainz translates an already-decoded MusicBrainz WS/2
// release into the matcher's vocabulary (internal/engine/matcher),
// builds the Lucene album search query that asks for one (query.go),
// issues the WS/2 requests themselves (client.go), and paces every
// request this process sends against the shared wire budget
// (limiter.go). Nothing here reads the clock, and the client is the
// only file that reaches the network at all.
//
// Two things this package deliberately does not do. First, the browse
// fallback beets calls _ensure_complete_recordings
// (beetsplug/musicbrainz.py:602-629) is not ported: it fires when a
// release's medium.tracks sum exceeds BROWSE_MAXTRACKS (500) and issues
// paginated browse_recordings requests, which is client work this
// package has no channel for. The largest committed release
// (real-boxset-rachmaninov-complete-stock, 470 tracks) stays under the
// line, and the translation tier guards the assumption by failing loudly
// the day a corpus case crosses it, rather than silently drifting from
// an oracle that fetched pages a cassette never carried. Second, no
// Unicode normalization runs anywhere in this package, matching pinned
// beets: no beets translation call site invokes unicodedata.normalize,
// so a Go port that "helpfully" normalized would diverge from what it
// ports.
package musicbrainz
