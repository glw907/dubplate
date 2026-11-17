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
	f, err := openFile(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	if _, err := readSignature(f, path); err != nil {
		return 0, err
	}
	return FormatFLAC, nil
}

func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("tagio: open %s: %w", path, err)
	}
	return f, nil
}

// readSignature returns f's leading four bytes, zero past a short read.
// When they are not FLAC's signature it also returns
// [ErrUnsupportedContainer], or the read's own error when the read
// failed short of end of file.
func readSignature(f *os.File, path string) ([4]byte, error) {
	var sig [4]byte
	n, err := f.ReadAt(sig[:], 0)
	if n < len(sig) && err != nil && !errors.Is(err, io.EOF) {
		return sig, fmt.Errorf("tagio: read %s: %w", path, err)
	}
	if string(sig[:]) != flacSignature {
		return sig, fmt.Errorf("tagio: %s: %w", path, ErrUnsupportedContainer)
	}
	return sig, nil
}
