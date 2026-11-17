package tagio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// blockPadding is PADDING's block type (RFC 9639 section 8.1).
const blockPadding = 1

// paddingSize is the one PADDING block every write emits, whatever the
// rewritten container carried.
const paddingSize = 8 * 1024

// vendorPrefix marks every vendor string this package creates.
const vendorPrefix = "dubplate "

const developmentVendorMarker = "development"

// Write rewrites path's tags and pictures. Every key ts owns is
// replaced, a comment's field name matching case insensitively as RFC
// 9639 section 8.6 requires, and every other comment survives in place
// and in order.
//
// pictures carries three states: nil leaves path's pictures untouched,
// a non-nil empty slice strips every picture, and a non-empty slice
// replaces the whole set in the order passed. Each picture's media
// type is sniffed from its data and overrides the declared one, and
// its geometry is written as zero.
//
// A file Sniff does not recognize hands back Sniff's
// [ErrUnsupportedContainer] unchanged. That refusal, a malformed
// container, and a picture whose data matches no recognized image
// format all come back before path is opened for writing, so a refusal
// never half-writes a file. Every block this package does not own
// survives, and one fresh PADDING block always ends the chain. The
// rewrite is atomic, so a crash or a concurrent reader never sees a
// partial file, and mode is the file's mode afterwards.
func Write(path string, ts TagSet, pictures []Picture, mode os.FileMode) error {
	if _, err := Sniff(path); err != nil {
		return err
	}
	return writeFLAC(path, ts, pictures, mode)
}

// writeFLAC is Write for a FLAC file: it parses the container, plans the
// block chain, and streams the rewrite through atomicfile.
func writeFLAC(path string, ts TagSet, pictures []Picture, mode os.FileMode) error {
	c, err := Parse(path)
	if err != nil {
		return err
	}

	newPictures, err := encodePictures(pictures)
	if err != nil {
		return err
	}

	blocks := planBlocks(c, ts, pictures != nil, newPictures)
	return atomicfile.WriteStream(path, mode, func(w io.Writer) error {
		if err := writeBlocks(w, blocks); err != nil {
			return err
		}
		return copyFrameRegion(w, path, c.FrameOffset, c.FrameLength)
	})
}

// writeBlock is one block Write is about to emit: its type and its
// final payload.
type writeBlock struct {
	Type    byte
	Payload []byte
}

// planBlocks decides the block chain Write emits, without the trailing
// PADDING block. Every block survives in place except PADDING, which
// is dropped; VORBIS_COMMENT, which carries ts merged onto c's
// comments under c's vendor, or is created fresh right after
// STREAMINFO; and PICTURE, which survives when replacePictures is
// false and otherwise gives way to newPictures, placed where the first
// incoming PICTURE sat or at the end.
func planBlocks(c Container, ts TagSet, replacePictures bool, newPictures []writeBlock) []writeBlock {
	hasComment, vendor := false, freshVendor()
	for _, b := range c.Blocks {
		if b.Type == blockVorbisComment {
			hasComment, vendor = true, c.VorbisComment.Vendor
			break
		}
	}
	comment := writeBlock{
		Type:    blockVorbisComment,
		Payload: encodeVorbisComment(vendor, mergeComments(c.VorbisComment.Comments, ts)),
	}

	out := make([]writeBlock, 0, len(c.Blocks)+1+len(newPictures))
	picturesPlaced := false
	for _, b := range c.Blocks {
		switch b.Type {
		case blockPadding:
			continue
		case blockPicture:
			if !replacePictures {
				out = append(out, writeBlock{Type: b.Type, Payload: b.Payload})
				continue
			}
			if !picturesPlaced {
				out = append(out, newPictures...)
				picturesPlaced = true
			}
		case blockVorbisComment:
			out = append(out, comment)
		case blockStreamInfo:
			out = append(out, writeBlock{Type: b.Type, Payload: b.Payload})
			if !hasComment {
				out = append(out, comment)
			}
		default:
			out = append(out, writeBlock{Type: b.Type, Payload: b.Payload})
		}
	}
	if replacePictures && !picturesPlaced {
		out = append(out, newPictures...)
	}
	return out
}

// mergeComments applies Serialize(ts) to existing one field at a time,
// in Serialize's order: the field's keys are deleted case insensitively
// wherever they sit, then its fresh comments are appended. A key no
// field owns, such as a foreign tagger's, is never deleted. Because a
// colliding pair's plural serializes after its singular, the plural's
// delete removes the singular's comments and its values take their
// place.
func mergeComments(existing []Comment, ts TagSet) []Comment {
	out := append([]Comment(nil), existing...)
	for _, kw := range Serialize(ts) {
		out = deleteKeys(out, kw.Deletes)
		out = append(out, kw.Writes...)
	}
	return out
}

// deleteKeys drops every comment in cs whose name matches one of keys
// case insensitively (RFC 9639 section 8.6), keeping every other
// comment in place and in order.
func deleteKeys(cs []Comment, keys []string) []Comment {
	if len(keys) == 0 {
		return cs
	}
	out := make([]Comment, 0, len(cs))
	for _, c := range cs {
		if !matchesAnyKey(c.Name, keys) {
			out = append(out, c)
		}
	}
	return out
}

func matchesAnyKey(name string, keys []string) bool {
	for _, k := range keys {
		if strings.EqualFold(name, k) {
			return true
		}
	}
	return false
}

// freshVendor names the vendor of a VORBIS_COMMENT block this package
// creates: vendorPrefix and the main module's version, or
// developmentVendorMarker when the build embeds none, as a test binary
// and an unmodulated go build do.
func freshVendor() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return vendorPrefix + developmentVendorMarker
	}
	return vendorPrefix + info.Main.Version
}

// writeBlocks emits the FLAC signature, every block in blocks with the
// last-block flag cleared, then the trailing PADDING block, the only
// block that carries the flag.
func writeBlocks(w io.Writer, blocks []writeBlock) error {
	if _, err := w.Write([]byte(flacSignature)); err != nil {
		return fmt.Errorf("tagio: write signature: %w", err)
	}
	for _, b := range blocks {
		if err := emitBlock(w, b.Type, false, b.Payload); err != nil {
			return err
		}
	}
	return emitBlock(w, blockPadding, true, make([]byte, paddingSize))
}

// maxBlockPayload is the smallest payload a block header's three-byte
// length field cannot declare (RFC 9639 section 8.1).
const maxBlockPayload = 1 << 24

// emitBlock writes one block header, the last-block flag in its top bit
// and the type in its low seven, then the payload's three-byte
// big-endian length and the payload itself. A payload too long for the
// length field is refused before anything is written.
func emitBlock(w io.Writer, blockType byte, last bool, payload []byte) error {
	length := len(payload)
	if length >= maxBlockPayload {
		return fmt.Errorf("tagio: block payload %d bytes: %w", length, ErrMalformedContainer)
	}
	header := blockType
	if last {
		header |= 0x80
	}
	if _, err := w.Write([]byte{header, byte(length >> 16), byte(length >> 8), byte(length)}); err != nil {
		return fmt.Errorf("tagio: write block header: %w", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("tagio: write block payload: %w", err)
	}
	return nil
}

// appendLengthPrefixed appends s to buf behind its length as a uint32
// in order, the framing VORBIS_COMMENT and PICTURE share apart from
// byte order.
func appendLengthPrefixed(order binary.AppendByteOrder, buf []byte, s string) []byte {
	buf = order.AppendUint32(buf, uint32(len(s)))
	return append(buf, s...)
}

// copyFrameRegion copies path's frame region into w through a section
// reader, so the audio is never held in memory whatever its size. It
// uses CopyN because a section reader over a file that has shrunk
// since the parse ends early without an error, and a short copy would
// otherwise be renamed into place as a truncated FLAC.
func copyFrameRegion(w io.Writer, path string, offset, length int64) error {
	src, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("tagio: open %s: %w", path, err)
	}
	defer func() { _ = src.Close() }()
	if _, err := io.CopyN(w, io.NewSectionReader(src, offset, length), length); err != nil {
		return fmt.Errorf("tagio: copy frame region of %s: %w", path, err)
	}
	return nil
}
