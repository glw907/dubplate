package tagio

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// pictureTypeMax is the ID3v2 APIC enumeration's highest defined
// value, publisher_logo. An out-of-range index degrades to 0, other,
// on read (P4), which is what mediafile's Image.type_index does when
// handed an integer index outside its enumeration
// (mediafile/utils/image.py:40-47, pinned).
const pictureTypeMax = 20

// pictureTypeFront is the front-cover index inside the APIC
// enumeration (P4, P5).
const pictureTypeFront = 3

// picturePayloadMin is a PICTURE block's shortest legal payload: the
// eight u(32) fields RFC 9639 section 8.8 requires (picture type,
// media-type length, description length, width, height, colour
// depth, colour count, data length), 32 bytes, before a single byte
// of the three variable-length fields the block also carries. Those
// eight fields are not contiguous, so this is a minimum length rather
// than a header to read in one pass.
const picturePayloadMin = 32

// Picture is one decoded PICTURE block (RFC 9639 section 8.8, Table
// 12), carrying the eight values the table defines in the order it
// defines them (P1). The block's three declared lengths, for the
// media type, the description and the data, are framing rather than
// values and this type carries none of them.
//
// Geometry (Width, Height, Depth, Colors) is decoded on read, but a
// write always emits zero for all four (P6): mediafile's FLAC image
// write builds a fresh picture and sets only data, type, mime and
// desc, so a Go writer that filled geometry from a decoded image
// would write more than the reference did.
type Picture struct {
	// Type is the ID3v2 APIC enumeration index, 0 through 20, front
	// cover at 3. An out-of-range value on read degrades to 0 (P4).
	Type int
	// MediaType is the block's declared media type string, for
	// example "image/png".
	MediaType string
	// Description is the block's declared description string.
	Description string
	// Width is the picture's width in pixels, as decoded from the
	// block.
	Width uint32
	// Height is the picture's height in pixels, as decoded from the
	// block.
	Height uint32
	// Depth is the picture's colour depth in bits per pixel, as
	// decoded from the block.
	Depth uint32
	// Colors is the number of colours used for an indexed-colour
	// picture, as decoded from the block.
	Colors uint32
	// Data is the picture's raw bytes.
	Data []byte
}

// FrontCover selects c's front-cover picture, reproducing mediafile's
// CoverArtField.guess_cover_image (mediafile/fields.py:253-260,
// INV:466-470): the single candidate when exactly one picture exists,
// otherwise the first picture typed front, otherwise the first
// picture of any type. The single-candidate rule runs before the
// front-cover rule, so a lone back cover is the cover (P5). The
// second return value is false only when c carries no picture at
// all.
func (c Container) FrontCover() (Picture, bool) {
	if len(c.Pictures) == 0 {
		return Picture{}, false
	}
	if len(c.Pictures) == 1 {
		return c.Pictures[0], true
	}
	for _, p := range c.Pictures {
		if p.Type == pictureTypeFront {
			return p, true
		}
	}
	return c.Pictures[0], true
}

// decodePicture decodes one PICTURE block's payload (RFC 9639 section
// 8.8, Table 12, INV:770-776), bounds-checking each of its three
// declared lengths, the media type's, the description's and the
// data's, against the payload's own remaining extent before a slice
// is taken (P3). No length is used to size an allocation: every field
// this function returns is a re-slice of payload, so a crafted length
// is refused without the process ever allocating on its behalf. index
// names the block's position in the chain for the returned error.
func decodePicture(payload []byte, index int) (Picture, error) {
	if len(payload) < picturePayloadMin {
		return Picture{}, fmt.Errorf("PICTURE block %d carries %d bytes, want at least %d", index, len(payload), picturePayloadMin)
	}

	typeIndex := binary.BigEndian.Uint32(payload[0:4])
	if typeIndex > pictureTypeMax {
		typeIndex = 0
	}

	mediaType, rest, err := pictureField(payload[4:], "media type")
	if err != nil {
		return Picture{}, fmt.Errorf("PICTURE block %d: %w", index, err)
	}
	description, rest, err := pictureField(rest, "description")
	if err != nil {
		return Picture{}, fmt.Errorf("PICTURE block %d: %w", index, err)
	}

	const geometrySize = 20 // width, height, depth, colours, data length: five u(32) fields
	if len(rest) < geometrySize {
		return Picture{}, fmt.Errorf("PICTURE block %d carries no geometry, %d bytes remain", index, len(rest))
	}
	width := binary.BigEndian.Uint32(rest[0:4])
	height := binary.BigEndian.Uint32(rest[4:8])
	depth := binary.BigEndian.Uint32(rest[8:12])
	colors := binary.BigEndian.Uint32(rest[12:16])
	dataLen := binary.BigEndian.Uint32(rest[16:20])
	rest = rest[20:]

	if uint64(dataLen) > uint64(len(rest)) {
		return Picture{}, fmt.Errorf("PICTURE block %d declares %d data bytes, %d remain", index, dataLen, len(rest))
	}

	return Picture{
		Type:        int(typeIndex),
		MediaType:   string(mediaType),
		Description: string(description),
		Width:       width,
		Height:      height,
		Depth:       depth,
		Colors:      colors,
		Data:        rest[:dataLen],
	}, nil
}

// Recognized media types this package sniffs on write (P7, R2),
// matching pinned filetype 1.2.0's own MIME constants for the raster
// formats a cover realistically carries.
const (
	mimeJPEG = "image/jpeg"
	mimePNG  = "image/png"
	mimeGIF  = "image/gif"
	mimeWEBP = "image/webp"
	mimeTIFF = "image/tiff"
	mimeBMP  = "image/bmp"
)

// sniffMediaType reports data's media type by its whole magic-byte
// test, reproducing pinned filetype 1.2.0's own image type checks
// (filetype/types/image.py, read at the pinned version) rather than a
// leading-prefix approximation. Two of the six tests carry an
// exclusion a leading-prefix sniff misses. WEBP additionally requires
// "VP" at offset 12, since a RIFF container that is not a WebP
// bitstream would otherwise match on "RIFF" and "WEBP" alone
// (:144-155). TIFF additionally excludes "CR" at offset 8, the byte
// pair that routes a Canon raw file to CR2 rather than TIFF in the
// reference (:193-199); a CR2-prefixed blob therefore matches no row
// here, which is R2's stated deviation, recorded in deviations.go: the
// reference reports image/x-canon-cr2 where this package refuses. A
// blob matching no row is refused with a wrapped error rather than
// written with a guessed or empty media type (R2); no caller branches
// on this case yet, so it carries no new sentinel (C14).
func sniffMediaType(data []byte) (string, error) {
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return mimeJPEG, nil
	case len(data) >= 4 && data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G':
		return mimePNG, nil
	case len(data) >= 3 && data[0] == 'G' && data[1] == 'I' && data[2] == 'F':
		return mimeGIF, nil
	case len(data) >= 14 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" && data[12] == 'V' && data[13] == 'P':
		return mimeWEBP, nil
	case len(data) >= 10 && isTIFFMagic(data) && (data[8] != 'C' || data[9] != 'R'):
		return mimeTIFF, nil
	case len(data) >= 2 && data[0] == 0x42 && data[1] == 0x4D:
		return mimeBMP, nil
	default:
		return "", errors.New("tagio: picture data matches no recognized image format")
	}
}

// isTIFFMagic reports whether data opens with either byte order TIFF
// permits: little-endian "II\x2A\x00" or big-endian "MM\x00\x2A".
func isTIFFMagic(data []byte) bool {
	return (data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A && data[3] == 0x00) ||
		(data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00 && data[3] == 0x2A)
}

// encodePicture lays out p as a PICTURE block payload (RFC 9639
// section 8.8, Table 12), big-endian throughout, the same asymmetry
// with VORBIS_COMMENT's little-endian framing decodePicture already
// carries. The media type is sniffed from p.Data and overrides
// whatever p.MediaType declares (P7); geometry is always written zero
// regardless of what p carries (P6), since mediafile's FLAC image
// write builds a fresh picture and sets only data, type, mime and desc
// (mediafile/storage/flac.py:30-37).
func encodePicture(p Picture) (writeBlock, error) {
	mediaType, err := sniffMediaType(p.Data)
	if err != nil {
		return writeBlock{}, fmt.Errorf("tagio: encode picture: %w", err)
	}
	var buf []byte
	buf = binary.BigEndian.AppendUint32(buf, uint32(p.Type))
	buf = appendPictureField(buf, mediaType)
	buf = appendPictureField(buf, p.Description)
	buf = binary.BigEndian.AppendUint32(buf, 0) // width, always zero (P6)
	buf = binary.BigEndian.AppendUint32(buf, 0) // height, always zero (P6)
	buf = binary.BigEndian.AppendUint32(buf, 0) // depth, always zero (P6)
	buf = binary.BigEndian.AppendUint32(buf, 0) // colours, always zero (P6)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(p.Data)))
	buf = append(buf, p.Data...)
	return writeBlock{Type: blockPicture, Payload: buf}, nil
}

// appendPictureField appends s to buf as a big-endian uint32 length
// prefix followed by s's bytes, PICTURE's own framing (unlike
// VORBIS_COMMENT's little-endian appendVorbisField).
func appendPictureField(buf []byte, s string) []byte {
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(s)))
	return append(buf, s...)
}

// encodePictures encodes every picture in pictures, in order, index by
// index, wrapping any sniff refusal with the picture's position. A nil
// pictures returns nil, nil: Write never calls this for R1's
// leave-alone state, since there is nothing to encode.
func encodePictures(pictures []Picture) ([]writeBlock, error) {
	if pictures == nil {
		return nil, nil
	}
	out := make([]writeBlock, 0, len(pictures))
	for i, p := range pictures {
		b, err := encodePicture(p)
		if err != nil {
			return nil, fmt.Errorf("picture %d: %w", i, err)
		}
		out = append(out, b)
	}
	return out, nil
}

// pictureField splits one length-prefixed field off the front of b,
// bounds-checking the declared length against what remains before any
// slice is taken (P3). Unlike VORBIS_COMMENT's vorbisField, the
// length is big-endian: RFC 9639 section 8.8 is the single asymmetry
// with VORBIS_COMMENT's little-endian framing (INV:770-776). label
// names the field in the returned error.
func pictureField(b []byte, label string) (field, rest []byte, err error) {
	if len(b) < 4 {
		return nil, nil, fmt.Errorf("%s length wants 4 bytes, %d remain", label, len(b))
	}
	size := binary.BigEndian.Uint32(b[:4])
	if uint64(4)+uint64(size) > uint64(len(b)) {
		return nil, nil, fmt.Errorf("%s declares %d bytes, %d remain", label, size, len(b)-4)
	}
	return b[4 : 4+size], b[4+size:], nil
}
