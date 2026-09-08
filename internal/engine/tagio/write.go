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

// paddingSize is the fixed PADDING reservation every write emits
// (rulings 5 and 6): 8 KiB, regardless of what the container it
// rewrites carried, and regardless of how many PADDING blocks (if
// any) that container had.
const paddingSize = 8 * 1024

// vendorPrefix marks every vendor string this package creates fresh,
// so a byte inspecting a file this package has touched can tell tagio
// wrote it (C9, R9).
const vendorPrefix = "dubplate "

// developmentVendorMarker is freshVendor's fallback when the build
// toolchain embeds no version: a `go test` binary and an unmodulated
// `go build` both leave [debug.BuildInfo.Main.Version] empty or
// "(devel)".
const developmentVendorMarker = "development"

// Write sniffs path, then rewrites its tags and pictures through the
// seam registry's row for the format Sniff reports. A format Sniff
// does not recognize, including a FLAC file once the registry has
// been emptied of its FLAC row, hands back [ErrUnsupportedContainer]
// unchanged and never opens path for writing.
func Write(path string, ts TagSet, pictures []Picture, mode os.FileMode) error {
	format, err := Sniff(path)
	if err != nil {
		return err
	}
	entry, ok := registry[format]
	if !ok || entry.write == nil {
		return fmt.Errorf("tagio: %s: %w", path, ErrUnsupportedContainer)
	}
	return entry.write(path, ts, pictures, mode)
}

// writeFLAC is FLAC's seam registry row for Write: every key ts owns
// is replaced, matching a comment's field name case insensitively as
// RFC 9639 section 8.6 requires, and every other comment survives in
// place and in order (C8). pictures carries R1's three caller states:
// nil leaves path's existing pictures untouched, a non-nil empty
// slice strips every picture, and a non-empty slice replaces the
// whole set in the order passed (P8). Every picture's media type is
// sniffed from its own data and overrides whatever the caller
// declared (P7), and every picture's geometry is written as zero
// regardless of what the caller carried (P6). writeFLAC refuses a
// malformed container, or a picture whose data matches no recognized
// image format, before path is ever opened for writing (C7, C14), so
// a refusal never half-writes a file, and every block this package
// does not own survives under a re-emitted header, with the
// last-block flag owned by the trailing PADDING block this write
// always appends (C10, C11). mode is the file's mode after the write,
// passed straight through to the rename rather than inherited from a
// temp file's default. The rewrite goes through atomicfile.WriteStream
// (C13), so a crash or a concurrent reader never observes a partial
// file.
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
// final payload. Every writeBlock carries the last-block flag cleared;
// only the trailing PADDING block writeBlocks appends carries it.
type writeBlock struct {
	Type    byte
	Payload []byte
}

// planBlocks decides the block chain Write emits, without the trailing
// PADDING block. Every block in c.Blocks survives in place, except
// every PADDING block, which is dropped outright (rulings 5 and 6);
// the VORBIS_COMMENT block, which is replaced with ts merged onto
// whatever comments c already carries (C8), keeping its loaded vendor;
// and PICTURE, whose fate depends on replacePictures (P8, R1). A
// container carrying no VORBIS_COMMENT block gets one created fresh,
// with a fresh vendor, immediately after STREAMINFO (C9a): hasComment
// is resolved once, ahead of the loop, so that placement happens
// exactly once regardless of where the loop is when it learns
// STREAMINFO has gone by.
//
// replacePictures is pictures != nil, resolved by the caller so this
// function only ever sees a bool and the already-encoded newPictures
// (R1's nil-vs-non-nil-empty distinction is Write's to make, not
// planBlocks's). When it is false, every incoming PICTURE block
// survives untouched: the leave-alone state. When it is true, every
// incoming PICTURE block is dropped and newPictures takes its place:
// the first incoming PICTURE block's own position if one existed,
// replacing it in place, or after VORBIS_COMMENT and before the
// trailing PADDING block if none did (P8), which planBlocks gets for
// free by appending newPictures once the loop over c.Blocks ends
// without ever having placed it.
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

// mergeComments applies Serialize(ts)'s per-field instructions to
// existing, one field at a time in Serialize's own order: a field's
// owned keys are deleted case insensitively wherever they currently
// sit, then the field's fresh comments are appended at the running
// end. A key no field in ts owns, such as a foreign tagger's comment,
// is never touched by any delete and so survives untouched; a field
// this write leaves absent deletes its keys and appends nothing.
// Running every field's delete-then-append in Serialize's fixed order
// is also what resolves the three colliding pairs (C3): the plural
// field always serializes after the singular it shares a key with, so
// its delete removes the singular's comments and its append leaves
// the plural's values in their place.
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

// matchesAnyKey reports whether name matches one of keys, case
// insensitively.
func matchesAnyKey(name string, keys []string) bool {
	for _, k := range keys {
		if strings.EqualFold(name, k) {
			return true
		}
	}
	return false
}

// encodeVorbisComment lays out a VORBIS_COMMENT payload: the vendor
// string, the comment count, then each comment's own length-prefixed
// "NAME=VALUE" bytes, little-endian throughout (RFC 9639 section 8.6).
func encodeVorbisComment(vendor string, comments []Comment) []byte {
	buf := appendVorbisField(nil, vendor)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(comments)))
	for _, c := range comments {
		buf = appendVorbisField(buf, c.Name+"="+c.Value)
	}
	return buf
}

// appendVorbisField appends s to buf as a little-endian uint32 length
// prefix followed by s's bytes.
func appendVorbisField(buf []byte, s string) []byte {
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(s)))
	return append(buf, s...)
}

// freshVendor names a vendor string for a container that carries no
// VORBIS_COMMENT block yet: vendorPrefix followed by the main module's
// build version, resolved once from the build information the
// toolchain already embeds. A test binary and an unmodulated `go
// build` both report no usable version, so both fall back to
// developmentVendorMarker (R9).
func freshVendor() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return vendorPrefix + developmentVendorMarker
	}
	return vendorPrefix + info.Main.Version
}

// writeBlocks emits the FLAC signature, every block in blocks with the
// last-block flag cleared, then the trailing PADDING block that
// carries it (C11): PADDING is always the last block a write emits,
// and it is the only block that owns the flag.
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

// maxBlockPayload is the largest payload a three-byte length field can
// declare (RFC 9639 section 8.1): 1<<24 bytes and above overflow it.
const maxBlockPayload = 1 << 24

// checkBlockPayloadSize reports [ErrMalformedContainer] when length
// cannot fit emitBlock's three-byte length field, taking the length
// alone so a test can grade the boundary without allocating a
// maxBlockPayload-sized slice.
func checkBlockPayloadSize(length int) error {
	if length >= maxBlockPayload {
		return fmt.Errorf("tagio: block payload %d bytes: %w", length, ErrMalformedContainer)
	}
	return nil
}

// emitBlock writes one block header, RFC 9639 section 8.1's
// last-block flag in the header's top bit and the type in its low
// seven bits, then the payload's length as a three-byte big-endian
// integer, followed by payload itself. A payload at or beyond
// maxBlockPayload cannot fit the three-byte length field, so emitBlock
// refuses it before writing anything.
func emitBlock(w io.Writer, blockType byte, last bool, payload []byte) error {
	length := len(payload)
	if err := checkBlockPayloadSize(length); err != nil {
		return err
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

// copyFrameRegion copies path's frame region, at the offset and length
// Parse recorded, straight into w: a section reader over the source
// file feeding io.Copy, never a []byte holding the whole region, so
// the audio is untouched (C12) and the region's cost stays fixed
// regardless of the file's size (C13).
func copyFrameRegion(w io.Writer, path string, offset, length int64) error {
	src, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("tagio: open %s: %w", path, err)
	}
	defer func() { _ = src.Close() }()
	if _, err := io.Copy(w, io.NewSectionReader(src, offset, length)); err != nil {
		return fmt.Errorf("tagio: copy frame region of %s: %w", path, err)
	}
	return nil
}
