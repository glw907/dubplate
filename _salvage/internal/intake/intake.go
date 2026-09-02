// Package intake runs the uploader's in-request stream checks: a ZIP
// central-directory walk that never extracts a member, and a FLAC
// STREAMINFO parse that never decodes an audio frame. Both feed one
// FileVerdict per file into Decide, which enforces the album-as-verdict-
// unit rule the rest of the pipeline depends on: one failing file rejects
// the whole album. internal/gate's full decode (flac -t) stays the
// import's deep check; intake exists to give a contributor a
// same-request answer before that ever runs.
package intake

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/glw907/dubplate/internal/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// FileVerdict is one archive member's or one standalone upload's outcome
// from [CheckZIP] or [CheckFLAC]. OK is false unless Reason is
// [reject.ReasonNone].
type FileVerdict struct {
	Name   string
	OK     bool
	Reason reject.Reason
}

// AlbumVerdict is one contributor upload's album-wide outcome: the
// verdict unit the pipeline enforces end to end, never a partial admit.
type AlbumVerdict struct {
	OK    bool
	Files []FileVerdict
}

// Decide implements the album-as-verdict-unit rule: OK is true only when
// every file in files passed, so one failing file rejects the whole
// album rather than admitting the rest.
func Decide(files []FileVerdict) AlbumVerdict {
	ok := true
	for _, f := range files {
		if !f.OK {
			ok = false
			break
		}
	}
	return AlbumVerdict{OK: ok, Files: files}
}

// Caps on a ZIP's central directory, enforced from the header fields
// alone so CheckZIP never decompresses a member to reject it. They bound
// pathological inputs (a zip-bomb-shaped entry count or declared size),
// not ordinary albums: a multi-disc lossless box set sits comfortably
// under maxZipTotalUncompressed.
const (
	maxZipEntries           = 2000
	maxZipEntryNameBytes    = 4096
	maxZipEntryUncompressed = 2 << 30 // 2 GiB per member
	maxZipTotalUncompressed = 8 << 30 // 8 GiB across the whole archive
)

// ErrTooManyEntries marks a ZIP whose central directory lists more
// members than [CheckZIP] will walk.
var ErrTooManyEntries = errors.New("intake: zip has too many entries")

// ErrArchiveTooLarge marks a ZIP whose central directory declares more
// combined uncompressed content than [CheckZIP] allows, read from the
// headers alone -- the same zip-bomb defense [internal/archive.Limits]
// applies at extraction time, checked here before a single byte is
// decompressed.
var ErrArchiveTooLarge = errors.New("intake: zip declared size exceeds cap")

// CheckZIP walks r's central directory -- the member list, names, and
// declared sizes -- without decompressing a single member, and returns
// one FileVerdict per regular-file entry; a directory entry is skipped
// and never appears in the result. A symlink member, an absolute path,
// or a path whose ".." components would resolve outside the extraction
// root fails as [reject.ReasonUnsafePath]; a member over its
// declared-size cap fails as [reject.ReasonTooLarge]. An entry count or
// aggregate declared size past CheckZIP's own caps returns a wrapped
// error rather than a FileVerdict, since neither names one file to
// attach a friendly reason to. p is accepted for signature symmetry with
// [CheckFLAC] but unused here: CheckZIP checks the container, not audio
// content.
func CheckZIP(r io.ReaderAt, size int64, p gate.Policy) ([]FileVerdict, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("read zip central directory: %w", err)
	}
	if len(zr.File) > maxZipEntries {
		return nil, fmt.Errorf("%d entries: %w", len(zr.File), ErrTooManyEntries)
	}

	var total uint64
	verdicts := make([]FileVerdict, 0, len(zr.File))
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		total += f.UncompressedSize64

		switch {
		case !safeZipName(f.Name) || f.FileInfo().Mode()&fs.ModeSymlink != 0:
			verdicts = append(verdicts, FileVerdict{Name: f.Name, Reason: reject.ReasonUnsafePath})
		case len(f.Name) > maxZipEntryNameBytes || f.UncompressedSize64 > maxZipEntryUncompressed:
			verdicts = append(verdicts, FileVerdict{Name: f.Name, Reason: reject.ReasonTooLarge})
		default:
			verdicts = append(verdicts, FileVerdict{Name: f.Name, OK: true})
		}
	}
	if total > maxZipTotalUncompressed {
		return nil, fmt.Errorf("%d declared bytes: %w", total, ErrArchiveTooLarge)
	}

	return verdicts, nil
}

// safeZipName reports whether name is safe to eventually extract: no
// absolute path, no backslash (ZIP entries are always "/"-separated, so
// a backslash is either a literal filename character no legitimate
// encoder produces or a Windows-style path smuggled past a naive
// "/"-only check), and no ".." component that would resolve outside the
// extraction root.
func safeZipName(name string) bool {
	if name == "" || strings.Contains(name, `\`) || path.IsAbs(name) {
		return false
	}
	clean := path.Clean(name)
	return clean != ".." && !strings.HasPrefix(clean, "../")
}

// flacMagic is the four-byte marker every FLAC stream starts with.
var flacMagic = []byte("fLaC")

// streamInfoBlockType is the metadata block type STREAMINFO carries; the
// FLAC format spec requires it be the first metadata block in a stream.
const streamInfoBlockType = 0

// streamInfoLength is STREAMINFO's fixed payload length in bytes, per
// the FLAC format spec.
const streamInfoLength = 34

// CheckFLAC parses a FLAC stream's STREAMINFO block from r -- the magic
// marker, sample rate, and bits-per-sample -- without decoding a single
// audio frame, and returns one FileVerdict for name. A missing magic
// marker, a first metadata block that isn't STREAMINFO, or a stream that
// ends before STREAMINFO's bytes arrive all fail as
// [reject.ReasonCorrupt]; a sample rate or bit depth below p's floor
// fails as [reject.ReasonBelowCDQuality] -- the same
// [gate.Policy.MinSampleRate] and [gate.Policy.MinBitsPerSample] floor
// [gate.Check]'s deep pass enforces, threaded through the same Policy
// value rather than a second copy of the constants. This is the fast,
// in-request check; [gate.Check]'s `flac -t` full decode stays the
// import's deep check, catching frame-level corruption STREAMINFO alone
// cannot see.
func CheckFLAC(r io.Reader, name string, p gate.Policy) (FileVerdict, error) {
	header := make([]byte, 8) // 4-byte magic + 4-byte metadata block header
	if err := readExact(r, header); err != nil {
		if isShortRead(err) {
			return FileVerdict{Name: name, Reason: reject.ReasonCorrupt}, nil
		}
		return FileVerdict{}, err
	}
	if !bytes.Equal(header[:4], flacMagic) {
		return FileVerdict{Name: name, Reason: reject.ReasonCorrupt}, nil
	}

	blockType := header[4] & 0x7f
	blockLen := int(header[5])<<16 | int(header[6])<<8 | int(header[7])
	if blockType != streamInfoBlockType || blockLen != streamInfoLength {
		return FileVerdict{Name: name, Reason: reject.ReasonCorrupt}, nil
	}

	info := make([]byte, streamInfoLength)
	if err := readExact(r, info); err != nil {
		if isShortRead(err) {
			return FileVerdict{Name: name, Reason: reject.ReasonCorrupt}, nil
		}
		return FileVerdict{}, err
	}

	// Bytes 10-17 of STREAMINFO pack sample rate (20 bits), channel count
	// minus one (3 bits), bits-per-sample minus one (5 bits), and total
	// sample count (36 bits) into one 64-bit big-endian field; channels
	// and total samples go unused here since CheckFLAC only enforces the
	// CD-quality floor.
	packed := binary.BigEndian.Uint64(info[10:18])
	sampleRate := int(packed >> 44)
	bitsPerSample := int((packed>>36)&0x1f) + 1

	if sampleRate == 0 {
		return FileVerdict{Name: name, Reason: reject.ReasonCorrupt}, nil
	}
	if sampleRate < p.MinSampleRate || bitsPerSample < p.MinBitsPerSample {
		return FileVerdict{Name: name, Reason: reject.ReasonBelowCDQuality}, nil
	}

	return FileVerdict{Name: name, OK: true}, nil
}

// readExact fills buf entirely from r, wrapping [io.ReadFull]'s error
// with context CheckFLAC's callers can log.
func readExact(r io.Reader, buf []byte) error {
	if _, err := io.ReadFull(r, buf); err != nil {
		return fmt.Errorf("read flac stream: %w", err)
	}
	return nil
}

// isShortRead reports whether err marks a stream that ended before the
// bytes CheckFLAC needed arrived -- a truncated upload, not a transport
// failure.
func isShortRead(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}
