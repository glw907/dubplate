package tagio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// flacSignature opens every FLAC stream, ahead of the metadata
// blocks.
const flacSignature = "fLaC"

// Block types this package recognizes by number (RFC 9639 section
// 8.1). Every other value from 7 through 126 is a reserved type this
// package preserves without decoding (C10); 127 is never a legal
// block type, reserved by the format itself so a reader can tell a
// metadata block header apart from the start of an audio frame,
// whose sync code always begins with the byte 0xFF.
const (
	blockStreamInfo    = 0
	blockVorbisComment = 4
	blockSeekTable     = 3
	blockInvalidType   = 127
)

// streamInfoSize is STREAMINFO's fixed payload width (RFC 9639
// section 8.2). A block of any other size is refused rather than
// read defensively.
const streamInfoSize = 34

// ErrMalformedContainer marks a container ruling 9 refuses outright:
// a second VORBIS_COMMENT or SEEKTABLE block, a STREAMINFO that is
// not first or not 34 bytes, a block following a header whose
// last-block flag is set, a declared length running past the end of
// the file, or a comment field name carrying a code point outside
// U+0020 through U+007E. Parse returns the zero Container alongside
// this error: a malformed file is never partly read.
var ErrMalformedContainer = errors.New("tagio: malformed FLAC container")

// ErrUnsupportedContainer marks a stream whose first four bytes are
// not the FLAC signature. Rung 7's allowlist gate branches on it
// (C14).
var ErrUnsupportedContainer = errors.New("tagio: unsupported container")

// Block is one metadata block, carried by type, declared length and
// payload, so a block this package does not decode still survives a
// write unchanged (C10).
type Block struct {
	// Type is the block's seven-bit type. Values 0 through 6 are the
	// types RFC 9639 defines; 7 through 126 are reserved types this
	// package preserves without understanding.
	Type byte
	// Last reports whether this block's header carried the
	// last-metadata-block flag.
	Last bool
	// Length is the block's declared payload length in bytes.
	Length int
	// Payload is the block's raw bytes, exactly Length long.
	Payload []byte
}

// StreamInfo is STREAMINFO's decoded payload (RFC 9639 section 8.2),
// laid out at the payload offsets
// plans/2026-09-05-rung-4a-shadow-runner.md:2404-2416 corrects: the
// four size fields, the eight-byte pack at [10:18], then the
// sixteen-byte MD5 at [18:34].
type StreamInfo struct {
	MinBlockSize  uint16
	MaxBlockSize  uint16
	MinFrameSize  uint32 // 24-bit
	MaxFrameSize  uint32 // 24-bit
	SampleRate    uint32
	Channels      uint8
	BitsPerSample uint8
	TotalSamples  uint64
	MD5           [16]byte
}

// Comment is one VORBIS_COMMENT entry, its field name and value cut
// apart at the first '=' the way RFC 9639 section 8.6 defines a
// comment.
type Comment struct {
	Name  string
	Value string
}

// VorbisComment is the decoded VORBIS_COMMENT block: the vendor
// string and every comment as a sequence, not a map, because a key
// may repeat and a map would keep only one occurrence
// (spec:688-690).
type VorbisComment struct {
	Vendor   string
	Comments []Comment
}

// Container is one FLAC file's parsed metadata: its block chain in
// order, STREAMINFO and VORBIS_COMMENT decoded, and the frame
// region's extent as two numbers rather than as bytes (C15).
type Container struct {
	Blocks        []Block
	StreamInfo    StreamInfo
	VorbisComment VorbisComment
	// FrameOffset is the byte offset, from the start of the file,
	// where the audio frames begin.
	FrameOffset int64
	// FrameLength is the number of bytes the frame region holds.
	FrameLength int64
}

// Parse reads path's metadata region and decodes it into a
// Container. It never reads the frame region: the region is measured
// as an offset and a length, and every declared block length is
// bounds-checked against the file's remaining extent before any
// payload slice is read (C7a).
//
// A signature other than fLaC returns [ErrUnsupportedContainer]. Any
// of C7's refusal shapes returns [ErrMalformedContainer], and the
// returned Container is always the zero value in that case.
func Parse(path string) (Container, error) {
	f, err := os.Open(path)
	if err != nil {
		return Container{}, fmt.Errorf("tagio: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return Container{}, fmt.Errorf("tagio: stat %s: %w", path, err)
	}

	sig := make([]byte, len(flacSignature))
	if _, err := f.ReadAt(sig, 0); err != nil || string(sig) != flacSignature {
		return Container{}, fmt.Errorf("tagio: %s: %w: signature %q", path, ErrUnsupportedContainer, sig)
	}

	return parseBlocks(f, path, stat.Size())
}

// parseBlocks reads the block chain starting right after the
// signature. It stops the first time the next four bytes cannot be
// read as a structurally valid block header: a clean end of file, a
// declared type of 127 (never legal, per the format's own reserved
// value), or a declared length running past the end of the file. A
// short header read that carries an error other than io.EOF is a
// genuine I/O failure, not a clean end of chain, and is returned
// wrapped rather than treated as the frame region's start. That
// stopping point is the frame region's start, provided the chain
// ends with exactly one block carrying the last-block flag; otherwise
// the file is malformed, whether it ran out of valid blocks before
// ever setting the flag or kept offering valid blocks after it
// (refusal 3 and refusal 4 of C7 share this one mechanism).
func parseBlocks(f *os.File, path string, size int64) (Container, error) {
	var (
		blocks           []Block
		streamInfo       StreamInfo
		vorbisComment    VorbisComment
		sawStreamInfo    bool
		sawVorbisComment bool
		sawSeekTable     bool
	)

	offset := int64(len(flacSignature))
	for {
		header := make([]byte, 4)
		n, err := f.ReadAt(header, offset)
		if n < 4 {
			if err != nil && !errors.Is(err, io.EOF) {
				return Container{}, fmt.Errorf("tagio: %s: read block header at offset %d: %w", path, offset, err)
			}
			break
		}
		last := header[0]&0x80 != 0
		blockType := header[0] & 0x7f
		length := int64(header[1])<<16 | int64(header[2])<<8 | int64(header[3])
		bodyOffset := offset + 4
		if blockType == blockInvalidType || bodyOffset+length > size {
			break
		}

		if len(blocks) == 0 && blockType != blockStreamInfo {
			return Container{}, malformed(path, "STREAMINFO is not the first block")
		}

		payload := make([]byte, length)
		if _, err := f.ReadAt(payload, bodyOffset); err != nil {
			return Container{}, fmt.Errorf("tagio: %s: read block type %d at offset %d: %w", path, blockType, bodyOffset, err)
		}

		switch blockType {
		case blockStreamInfo:
			if len(blocks) != 0 {
				return Container{}, malformed(path, "STREAMINFO is not the first block")
			}
			si, err := decodeStreamInfo(payload)
			if err != nil {
				return Container{}, malformed(path, err.Error())
			}
			streamInfo = si
			sawStreamInfo = true
		case blockVorbisComment:
			if sawVorbisComment {
				return Container{}, malformed(path, "more than one VORBIS_COMMENT block")
			}
			vc, err := decodeVorbisComment(payload)
			if err != nil {
				return Container{}, malformed(path, err.Error())
			}
			vorbisComment = vc
			sawVorbisComment = true
		case blockSeekTable:
			if sawSeekTable {
				return Container{}, malformed(path, "more than one SEEKTABLE block")
			}
			sawSeekTable = true
		}

		blocks = append(blocks, Block{Type: blockType, Last: last, Length: int(length), Payload: payload})
		offset = bodyOffset + length
	}

	if !sawStreamInfo {
		return Container{}, malformed(path, "no STREAMINFO block")
	}
	if err := checkLastFlag(blocks); err != nil {
		return Container{}, malformed(path, err.Error())
	}

	return Container{
		Blocks:        blocks,
		StreamInfo:    streamInfo,
		VorbisComment: vorbisComment,
		FrameOffset:   offset,
		FrameLength:   size - offset,
	}, nil
}

// checkLastFlag enforces the one invariant RFC 9639 section 8.1
// states: exactly one block carries the last-metadata-block flag,
// and it is the chain's final block. A block carrying the flag
// anywhere else, or no block carrying it at all, is refusal 3 of C7.
func checkLastFlag(blocks []Block) error {
	count := 0
	for i, b := range blocks {
		if !b.Last {
			continue
		}
		count++
		if i != len(blocks)-1 {
			return errors.New("a block follows a header whose last-block flag is set")
		}
	}
	if count != 1 {
		return errors.New("no block carries the last-block flag")
	}
	return nil
}

// decodeStreamInfo decodes STREAMINFO's 34-byte payload. The eight-byte
// pack at [10:18] holds the sample rate (20 bits), the channel count
// less one (3 bits), the sample width less one (5 bits) and the
// total sample count (36 bits), big-endian across the eight bytes;
// the payload's last sixteen bytes are the audio MD5, not part of
// the pack.
func decodeStreamInfo(payload []byte) (StreamInfo, error) {
	if len(payload) != streamInfoSize {
		return StreamInfo{}, fmt.Errorf("STREAMINFO carries %d bytes, want %d", len(payload), streamInfoSize)
	}
	pack := binary.BigEndian.Uint64(payload[10:18])
	var md5 [16]byte
	copy(md5[:], payload[18:34])
	return StreamInfo{
		MinBlockSize:  binary.BigEndian.Uint16(payload[0:2]),
		MaxBlockSize:  binary.BigEndian.Uint16(payload[2:4]),
		MinFrameSize:  uint32(payload[4])<<16 | uint32(payload[5])<<8 | uint32(payload[6]),
		MaxFrameSize:  uint32(payload[7])<<16 | uint32(payload[8])<<8 | uint32(payload[9]),
		SampleRate:    uint32(pack >> 44),
		Channels:      uint8(pack>>41&0x7) + 1,
		BitsPerSample: uint8(pack>>36&0x1f) + 1,
		TotalSamples:  pack & 0xFFFFFFFFF,
		MD5:           md5,
	}, nil
}

// decodeVorbisComment decodes a VORBIS_COMMENT payload into its
// vendor string and its comment sequence. The block is little-endian
// throughout: a vendor length and string, a comment count, then each
// comment's length and its NAME=VALUE bytes.
func decodeVorbisComment(payload []byte) (VorbisComment, error) {
	vendor, rest, err := vorbisField(payload)
	if err != nil {
		return VorbisComment{}, fmt.Errorf("vendor string: %w", err)
	}
	if len(rest) < 4 {
		return VorbisComment{}, fmt.Errorf("VORBIS_COMMENT carries no comment count, %d bytes remain", len(rest))
	}
	count := binary.LittleEndian.Uint32(rest[:4])
	rest = rest[4:]

	var comments []Comment
	for i := range count {
		field, remainder, err := vorbisField(rest)
		if err != nil {
			return VorbisComment{}, fmt.Errorf("comment %d: %w", i, err)
		}
		rest = remainder
		name, value, found := strings.Cut(string(field), "=")
		if !found {
			return VorbisComment{}, fmt.Errorf("comment %d: %q carries no '='", i, field)
		}
		if err := validateFieldName(name); err != nil {
			return VorbisComment{}, fmt.Errorf("comment %d: %w", i, err)
		}
		comments = append(comments, Comment{Name: name, Value: value})
	}
	return VorbisComment{Vendor: string(vendor), Comments: comments}, nil
}

// validateFieldName refuses a comment field name carrying a code
// point outside U+0020 through U+007E (RFC 9639 section 8.6). U+007E
// itself is inside the set: mutagen's is_valid_key stops one code
// point short, at 0x7D (mutagen/_vorbis.py:25-41), but tagio's bound
// is the RFC's rather than mutagen's.
func validateFieldName(name string) error {
	for _, r := range name {
		if r < 0x20 || r > 0x7e {
			return fmt.Errorf("comment field name %q carries a code point outside U+0020-U+007E", name)
		}
	}
	return nil
}

// vorbisField splits one length-prefixed field off the front of b,
// bounds-checking the declared length against what remains before
// any slice is taken.
func vorbisField(b []byte) (field, rest []byte, err error) {
	if len(b) < 4 {
		return nil, nil, fmt.Errorf("length field wants 4 bytes, %d remain", len(b))
	}
	size := binary.LittleEndian.Uint32(b[:4])
	if uint64(4)+uint64(size) > uint64(len(b)) {
		return nil, nil, fmt.Errorf("field declares %d bytes, %d remain", size, len(b)-4)
	}
	return b[4 : 4+size], b[4+size:], nil
}

// malformed wraps ErrMalformedContainer with path and a reason.
func malformed(path, reason string) error {
	return fmt.Errorf("tagio: %s: %w: %s", path, ErrMalformedContainer, reason)
}
