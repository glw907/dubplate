// Package tagio reads and writes FLAC tags. It knows nothing about
// matching: no consumer's type appears in its signatures, it imports no
// package under internal/harness, and it reads no clock. Every path and
// mode this package touches reaches it as a parameter from its caller.
//
// It sits beside internal/engine/tag, which computes the applied field
// and touches no file (internal/engine/tag/doc.go:8-11).
package tagio
