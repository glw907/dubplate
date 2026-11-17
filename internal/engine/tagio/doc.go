// Package tagio reads and writes FLAC tags. It knows nothing about
// matching: no consumer's type appears in its signatures, it imports no
// package under internal/harness, and it reads no clock. Every path and
// mode this package touches reaches it as a parameter from its caller.
//
// It sits beside internal/engine/tag, which computes the applied field
// and touches no file (internal/engine/tag/doc.go:8-11).
//
// Read and Write sniff the file first and refuse one that is not FLAC;
// Sniff and SupportedFormats name the formats tagio handles, and FLAC
// is the only one. Parse is FLAC's own entry point and returns a FLAC
// Container.
//
// Read stays tag-only: it returns a TagSet and nothing else. Pictures
// live on Container, populated by Parse, because a picture is not one
// of the 69 writable keys the tag projection carries (R12). A caller
// that wants both a file's tags and its pictures therefore parses the
// file twice, once through Read and once through Parse, rather than
// reaching for a second read entry point this package does not
// expose. The double parse is the accepted cost of keeping Read's
// contract narrow.
//
// Tags is Read's tag projection exported over an already-parsed
// Container, for a caller that wants a file's tags alongside a
// Container field Read does not return, such as StreamInfo. The import
// reader is the intended consumer, parsing a file once for its length
// and its tags rather than twice through Read and Parse both; no
// caller exists yet. Tags performs no I/O.
package tagio
