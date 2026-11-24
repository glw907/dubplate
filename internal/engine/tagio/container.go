package tagio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// flacSignature opens every FLAC stream, ahead of the metadata
// blocks.
const flacSignature = "fLaC"

// Block types this package recognizes by number (RFC 9639 section
// 8.1). Every other value from 7 through 126 is a reserved type this
// package preserves without decoding; 127 is never a legal
// block type, reserved by the format itself so a reader can tell a
// metadata block header apart from the start of an audio frame,
// whose sync code always begins with the byte 0xFF.
const (
	blockStreamInfo    = 0
	blockVorbisComment = 4
	blockSeekTable     = 3
	blockPicture       = 6
	blockInvalidType   = 127
)

// streamInfoSize is STREAMINFO's fixed payload width (RFC 9639
// section 8.2). A block of any other size is refused rather than
// read defensively.
const streamInfoSize = 34

// ErrMalformedContainer marks a container this package refuses outright:
// a second VORBIS_COMMENT or SEEKTABLE block, a STREAMINFO that is
// not first or not 34 bytes, a block following a header whose
// last-block flag is set, a declared length running past the end of
// the file, or a comment field name carrying a code point outside
// U+0020 through U+007E. Parse returns the zero Container alongside
// this error: a malformed file is never partly read.
var ErrMalformedContainer = errors.New("tagio: malformed FLAC container")

// ErrUnsupportedContainer marks a stream whose first four bytes are
// not the FLAC signature. The intake gate's allowlist branches on it.
var ErrUnsupportedContainer = errors.New("tagio: unsupported container")

// Block is one metadata block, carried by type, declared length and
// payload, so a block this package does not decode still survives a
// write unchanged.
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
// laid out at its payload offsets: the four size fields, the
// eight-byte pack at [10:18], then the sixteen-byte MD5 at [18:34].
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

// Container is one FLAC file's parsed metadata: its block chain in
// order, STREAMINFO and VORBIS_COMMENT decoded, every PICTURE block
// decoded in on-disk order, and the frame region's extent as two
// numbers rather than as bytes.
type Container struct {
	Blocks        []Block
	StreamInfo    StreamInfo
	VorbisComment VorbisComment
	// Pictures holds every PICTURE block this container carried,
	// decoded in the order they appeared. RFC 9639 section 8.8 allows
	// more than one, unlike SEEKTABLE and VORBIS_COMMENT, so a second
	// PICTURE block is not a refusal.
	Pictures []Picture
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
// payload slice is read.
//
// A signature other than fLaC returns [ErrUnsupportedContainer]. Any
// of the refusal shapes [ErrMalformedContainer] lists returns it, and
// the returned Container is always the zero value in that case.
func Parse(path string) (Container, error) {
	return parse(path, nil)
}

// Bounds caps the metadata a walk will allocate, in bytes. The chain's
// total is not a field of its own: it is four times PictureBytes.
type Bounds struct {
	// TagBytes is the largest VORBIS_COMMENT block admitted.
	TagBytes int64
	// PictureBytes is the largest PICTURE block admitted, and a quarter
	// of the largest chain of blocks of any type.
	PictureBytes int64
}

// ErrTagsTooLarge marks a VORBIS_COMMENT block whose declared length
// exceeds [Bounds].TagBytes, or a chain of more than 1024 blocks.
var ErrTagsTooLarge = errors.New("tagio: tag block too large")

// ErrPictureTooLarge marks a PICTURE block whose declared length
// exceeds [Bounds].PictureBytes, or a chain whose declared lengths
// total more than four times that.
var ErrPictureTooLarge = errors.New("tagio: picture block or metadata chain too large")

// chainFactor is how many picture caps a whole metadata chain may
// total, each block's four header bytes included.
const chainFactor = 4

// maxBlocks is the most blocks a bounded walk admits. RFC 9639 sets no
// number and real files carry a handful, so the bound exists for the walk's
// own cost: every block, however small, adds a Block and a header read, and
// a member's bytes could otherwise declare millions of empty ones.
const maxBlocks = 1024

// admit refuses a block from its declared length alone, so the
// refusal precedes the allocation of its payload. total includes this
// block.
func (b Bounds) admit(path string, blockType byte, length, total int64, blocks int) error {
	switch {
	case blocks > maxBlocks:
		return fmt.Errorf("tagio: %s: %w: the chain holds more than %d blocks", path, ErrTagsTooLarge, maxBlocks)
	case blockType == blockVorbisComment && length > b.TagBytes:
		return fmt.Errorf("tagio: %s: %w: a tag block declares %d bytes, the cap is %d", path, ErrTagsTooLarge, length, b.TagBytes)
	case blockType == blockPicture && length > b.PictureBytes:
		return fmt.Errorf("tagio: %s: %w: a picture block declares %d bytes, the cap is %d", path, ErrPictureTooLarge, length, b.PictureBytes)
	case total > chainFactor*b.PictureBytes:
		return fmt.Errorf("tagio: %s: %w: the chain declares %d bytes, the cap is %d", path, ErrPictureTooLarge, total, chainFactor*b.PictureBytes)
	}
	return nil
}

// ParseBounded is [Parse] with each block's declared length checked
// against bounds before its payload is allocated: a Vorbis comment
// block over TagBytes, or a chain of more than 1024 blocks, answers
// [ErrTagsTooLarge], and a picture block over PictureBytes, or a chain
// totalling more than four times PictureBytes, answers
// [ErrPictureTooLarge]. A block that is both too large and runs past
// the end of the file answers the bound's error, not
// [ErrMalformedContainer].
func ParseBounded(path string, bounds Bounds) (Container, error) {
	return parse(path, &bounds)
}

// parse opens path and walks its container; a nil bounds admits any
// block size. A file that is not FLAC is refused with the four bytes
// it carries, whatever kept them from being the signature.
func parse(path string, bounds *Bounds) (Container, error) {
	f, err := openFile(path)
	if err != nil {
		return Container{}, err
	}
	defer func() { _ = f.Close() }()

	sig, sigErr := readSignature(f, path)
	stat, err := f.Stat()
	if err != nil {
		return Container{}, fmt.Errorf("tagio: stat %s: %w", path, err)
	}
	if sigErr != nil {
		return Container{}, fmt.Errorf("tagio: %s: %w: signature %q", path, ErrUnsupportedContainer, sig[:])
	}
	return parseBlocks(f, path, stat.Size(), bounds)
}

// readContainer is [Parse] for Read and Write, opening path once. A
// file that is not FLAC is refused with [Sniff]'s error, not parse's.
func readContainer(path string) (Container, error) {
	f, err := openFile(path)
	if err != nil {
		return Container{}, err
	}
	defer func() { _ = f.Close() }()

	if _, err := readSignature(f, path); err != nil {
		return Container{}, err
	}
	stat, err := f.Stat()
	if err != nil {
		return Container{}, fmt.Errorf("tagio: stat %s: %w", path, err)
	}
	return parseBlocks(f, path, stat.Size(), nil)
}

// parseBlocks reads the block chain that follows the signature. The
// chain ends where the next four bytes are not a valid block header: a
// clean end of file, the illegal type 127, or a declared length running
// past the end of the file. A short header read failing with anything
// but io.EOF is an I/O error, not the end of the chain. The chain must
// end on exactly one block carrying the last-block flag, so a missing
// flag and a block after the flag are refused alike.
func parseBlocks(f *os.File, path string, size int64, bounds *Bounds) (Container, error) {
	var (
		c     chain
		total int64
	)
	offset := int64(len(flacSignature))
	var header [4]byte
	for {
		n, err := f.ReadAt(header[:], offset)
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
		if blockType == blockInvalidType {
			break
		}
		total += 4 + length
		if bounds != nil {
			if err := bounds.admit(path, blockType, length, total, len(c.blocks)+1); err != nil {
				return Container{}, err
			}
		}
		if bodyOffset+length > size {
			break
		}

		if len(c.blocks) == 0 && blockType != blockStreamInfo {
			return Container{}, malformed(path, "STREAMINFO is not the first block")
		}

		payload := make([]byte, length)
		if _, err := f.ReadAt(payload, bodyOffset); err != nil {
			return Container{}, fmt.Errorf("tagio: %s: read block type %d at offset %d: %w", path, blockType, bodyOffset, err)
		}
		if err := c.decode(path, blockType, payload); err != nil {
			return Container{}, err
		}

		c.blocks = append(c.blocks, Block{Type: blockType, Last: last, Length: int(length), Payload: payload})
		offset = bodyOffset + length
	}

	if !c.sawStreamInfo {
		return Container{}, malformed(path, "no STREAMINFO block")
	}
	if err := checkLastFlag(c.blocks); err != nil {
		return Container{}, malformed(path, err.Error())
	}

	return Container{
		Blocks:        c.blocks,
		StreamInfo:    c.streamInfo,
		VorbisComment: c.vorbisComment,
		Pictures:      c.pictures,
		FrameOffset:   offset,
		FrameLength:   size - offset,
	}, nil
}

// chain holds the blocks parseBlocks has read so far and what it has
// decoded from them.
type chain struct {
	blocks           []Block
	streamInfo       StreamInfo
	vorbisComment    VorbisComment
	pictures         []Picture
	sawStreamInfo    bool
	sawVorbisComment bool
	sawSeekTable     bool
}

// decode decodes payload as the next block of blockType and keeps what
// it carries. It refuses a STREAMINFO after the first block, a second
// VORBIS_COMMENT or SEEKTABLE, and a payload its type's decoder
// rejects. A block of any other type carries nothing parse keeps.
func (c *chain) decode(path string, blockType byte, payload []byte) error {
	switch blockType {
	case blockStreamInfo:
		if len(c.blocks) != 0 {
			return malformed(path, "STREAMINFO is not the first block")
		}
		si, err := decodeStreamInfo(payload)
		if err != nil {
			return malformed(path, err.Error())
		}
		c.streamInfo = si
		c.sawStreamInfo = true
	case blockVorbisComment:
		if c.sawVorbisComment {
			return malformed(path, "more than one VORBIS_COMMENT block")
		}
		vc, err := decodeVorbisComment(payload)
		if err != nil {
			return malformed(path, err.Error())
		}
		c.vorbisComment = vc
		c.sawVorbisComment = true
	case blockSeekTable:
		if c.sawSeekTable {
			return malformed(path, "more than one SEEKTABLE block")
		}
		c.sawSeekTable = true
	case blockPicture:
		p, err := decodePicture(payload, len(c.blocks))
		if err != nil {
			return malformed(path, err.Error())
		}
		c.pictures = append(c.pictures, p)
	}
	return nil
}

// checkLastFlag enforces RFC 9639 section 8.1's one chain invariant:
// exactly one block carries the last-metadata-block flag, and it is
// the final block.
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

// malformed wraps ErrMalformedContainer with path and a reason.
func malformed(path, reason string) error {
	return fmt.Errorf("tagio: %s: %w: %s", path, ErrMalformedContainer, reason)
}
