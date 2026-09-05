// Package tag holds beets' tag_album (autotag/match.py:383-468) and
// nothing else: the matcher's whole public entry point, the one place
// that both fetches candidate releases and scores them. It is the only
// package that may see both internal/engine/matcher and
// internal/engine/musicbrainz, because musicbrainz already imports
// matcher and the flow interleaves I/O with scoring (Resolution R18).
//
// It sits beside rung 8's planned internal/engine/tagio, and the
// pairing is deliberate: tagio reads and writes FLAC tags and knows
// nothing about matching, while this package computes no applied field
// and touches no file.
//
// It never imports internal/harness and never reads the clock: every
// evaluation date this package uses is a parameter.
package tag
