package tagio

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Format identifies a container format tagio reads and writes. FLAC is
// the only one; a second format brings its own constant, and Read and
// Write then branch on what Sniff reports.
type Format int

const (
	// FormatFLAC is native FLAC, RFC 9639's own container.
	FormatFLAC Format = iota
)

// SupportedFormats returns every Format tagio handles, in ascending
// order, so the upload gate can bound its allowlist against exactly
// what this package can read. Each call returns a new slice.
func SupportedFormats() []Format {
	return []Format{FormatFLAC}
}

// Sniff reads path's leading four bytes and reports which Format they
// identify. A file shorter than four bytes, or one whose leading bytes
// are not a supported format's signature, answers
// [ErrUnsupportedContainer]; the comparison is the full four bytes, so
// a stream that agrees with FLAC's signature on its first three bytes
// but not its fourth still sniffs as unsupported.
func Sniff(path string) (Format, error) {
	f, _, err := openFLAC(path)
	if f == nil {
		return 0, err
	}
	_ = f.Close()
	if err != nil {
		return 0, err
	}
	return FormatFLAC, nil
}

// openFLAC opens path and checks that it starts with FLAC's signature.
// A failed open returns a nil file and the open error. Otherwise the
// file comes back open, for the caller to close, with the four bytes
// the check read (zero past a short read) and a non-nil error when they
// are not the signature: the read's own error when it failed short of
// end of file, and [ErrUnsupportedContainer] otherwise. Sniff reports
// that error as it stands; parse reports every mismatch as an
// unsupported signature instead.
func openFLAC(path string) (*os.File, [4]byte, error) {
	var sig [4]byte
	f, err := os.Open(path)
	if err != nil {
		return nil, sig, fmt.Errorf("tagio: open %s: %w", path, err)
	}
	n, err := f.ReadAt(sig[:], 0)
	if n < len(sig) && err != nil && !errors.Is(err, io.EOF) {
		return f, sig, fmt.Errorf("tagio: read %s: %w", path, err)
	}
	if string(sig[:]) != flacSignature {
		return f, sig, fmt.Errorf("tagio: %s: %w", path, ErrUnsupportedContainer)
	}
	return f, sig, nil
}
