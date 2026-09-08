package tagio

import (
	"encoding/binary"
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
