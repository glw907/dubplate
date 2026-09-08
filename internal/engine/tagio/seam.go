package tagio

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
)

// Format identifies one container format tagio's seam registry
// recognizes. FLAC is the registry's only entry: the MP3 pass adds
// go-mp4's or bogem/id3v2's implementation as the registry's second
// row and, at that same moment, the interface this package
// deliberately does not declare yet (ruling 14). Today a single
// implementation has no second caller to satisfy, and a struct of
// func values already gives Read and Write their dispatch without
// one; a second registration costs exactly one more registry entry
// plus that interface, nothing more.
type Format int

const (
	// FormatFLAC is native FLAC, RFC 9639's own container.
	FormatFLAC Format = iota
)

// formatEntry is one seam registry row: the container's leading
// magic bytes and the func values Read and Write dispatch through.
// It is a struct of func values, never an interface (ruling 14): a
// second registration adds a second value of this same struct type,
// never a new interface satisfaction to prove.
type formatEntry struct {
	magic [4]byte
	read  func(path string) (TagSet, error)
	write func(path string, ts TagSet, pictures []Picture, mode os.FileMode) error
}

// registry is the seam's one unexported map: every format Read,
// Write and Sniff recognize, keyed by its exported Format value.
//
// Parse (container.go) deliberately never consults this registry: it
// is FLAC's own entry point, returns a FLAC Container, and a
// registry whose later rows return other formats' containers has
// nothing to hand it for those inputs (ruled, conductor 2026-09-08,
// m6). The bypass is asserted, not merely stated: emptying this map
// for a test stops Read and Write from reaching the golden while
// Parse keeps parsing it, which is the proof the bypass is a
// decision rather than an oversight.
var registry = map[Format]formatEntry{
	FormatFLAC: {
		magic: flacMagic,
		read:  readFLAC,
		write: writeFLAC,
	},
}

// flacMagic is flacSignature's four bytes as the fixed-size array the
// registry compares against, computed once from the constant rather
// than restated as a byte literal.
var flacMagic = [4]byte([]byte(flacSignature))

// SupportedFormats returns a copy of the seam registry's key set,
// ordered by each Format's own value, so rung 7 can bound its
// allowlist against exactly what tagio can handle in a deterministic
// order rather than a map's. A Format's value is assigned by
// declaration order (the iota block above), so this ordering is also
// registration order. Each call clones the underlying set, so
// mutating the returned slice never touches the registry itself
// (Mapping's own precedent, tagset.go).
func SupportedFormats() []Format {
	out := slices.Collect(maps.Keys(registry))
	slices.Sort(out)
	return out
}

// Sniff reads path's leading four bytes and reports which registered
// Format they identify. A file shorter than four bytes, or one whose
// leading bytes match no registered row, answers
// [ErrUnsupportedContainer]; the comparison is the full four-byte
// magic, so a stream that agrees with a registered row on its first
// three bytes but not its fourth still sniffs as unsupported.
func Sniff(path string) (Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("tagio: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var magic [4]byte
	n, err := f.ReadAt(magic[:], 0)
	if n < len(magic) {
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, fmt.Errorf("tagio: read %s: %w", path, err)
		}
		return 0, fmt.Errorf("tagio: %s: %w", path, ErrUnsupportedContainer)
	}

	for format, entry := range registry {
		if entry.magic == magic {
			return format, nil
		}
	}
	return 0, fmt.Errorf("tagio: %s: %w", path, ErrUnsupportedContainer)
}

// resolveEntry sniffs path and returns its seam registry row. It
// answers [ErrUnsupportedContainer] both when Sniff does not
// recognize path and when the sniffed format's registry row exists
// but is empty, the shape a test registers to prove Read and Write
// refuse a func-less row rather than panic on it.
func resolveEntry(path string) (formatEntry, error) {
	format, err := Sniff(path)
	if err != nil {
		return formatEntry{}, err
	}
	entry, ok := registry[format]
	if !ok {
		return formatEntry{}, fmt.Errorf("tagio: %s: %w", path, ErrUnsupportedContainer)
	}
	return entry, nil
}
