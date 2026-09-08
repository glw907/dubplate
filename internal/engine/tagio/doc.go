// Package tagio reads and writes FLAC tags. It knows nothing about
// matching: no consumer's type appears in its signatures, it imports no
// package under internal/harness, and it reads no clock. Every path and
// mode this package touches reaches it as a parameter from its caller.
//
// It sits beside internal/engine/tag, which computes the applied field
// and touches no file (internal/engine/tag/doc.go:8-11).
//
// Read and Write dispatch through the seam registry (seam.go), an
// unexported map keyed by Format that pairs each container's magic
// bytes with its read and write funcs; Sniff and SupportedFormats
// are the registry's exported surface. The registry carries exactly
// one row today, FLAC. The MP3 pass adds the registry's second row
// and, at that same moment, the interface this package deliberately
// does not declare yet: a struct of func values dispatches Read and
// Write with only one implementation, and a second registration
// costs exactly one more registry entry plus that interface, nothing
// more (ruling 14).
//
// Read stays tag-only: it returns a TagSet and nothing else. Pictures
// live on Container, populated by Parse, because a picture is not one
// of the 69 writable keys the tag projection carries (R12). A caller
// that wants both a file's tags and its pictures therefore parses the
// file twice, once through Read and once through Parse, rather than
// reaching for a second read entry point this package does not
// expose. The double parse is the accepted cost of keeping Read's
// contract narrow; rung 9, the pipeline that first needs both, is the
// caller that may fold the two calls into one if the cost is ever
// worth removing.
package tagio
