package tagio

import (
	"encoding/binary"
	"errors"
	"path/filepath"
	"testing"
)

// picturePayload assembles a PICTURE block payload (RFC 9639 section
// 8.8, Table 12) from its seven variable inputs, big-endian
// throughout, with every declared length taken from the caller rather
// than derived from the byte slice it prefixes, so a refusal fixture
// can declare a length its own bytes do not back. Geometry (width,
// height, depth, colours) is fixed at zero: no test in this package
// needs a nonzero value, since P6 emits zero on every write and this
// task decodes geometry only to prove it round-trips (P1).
func picturePayload(typeIndex, mediaTypeLen uint32, mediaType string, descLen uint32, description string, dataLen uint32, data []byte) []byte {
	var buf []byte
	buf = binary.BigEndian.AppendUint32(buf, typeIndex)
	buf = binary.BigEndian.AppendUint32(buf, mediaTypeLen)
	buf = append(buf, mediaType...)
	buf = binary.BigEndian.AppendUint32(buf, descLen)
	buf = append(buf, description...)
	buf = binary.BigEndian.AppendUint32(buf, 0) // width
	buf = binary.BigEndian.AppendUint32(buf, 0) // height
	buf = binary.BigEndian.AppendUint32(buf, 0) // depth
	buf = binary.BigEndian.AppendUint32(buf, 0) // colours
	buf = binary.BigEndian.AppendUint32(buf, dataLen)
	buf = append(buf, data...)
	return buf
}

// wellFormedPicturePayload builds a PICTURE payload whose three
// declared lengths match the bytes that follow them, so a test can
// vary one field, typeIndex, without every row restating the rest.
func wellFormedPicturePayload(typeIndex uint32) []byte {
	data := []byte{0xAA, 0xBB}
	return picturePayload(typeIndex, uint32(len("image/png")), "image/png", uint32(len("cover")), "cover", uint32(len(data)), data)
}

// TestDecodePictureTypeIndex covers P4: the type index degrades to 0
// above 20, front cover (3) and the boundary at 20 do not degrade.
func TestDecodePictureTypeIndex(t *testing.T) {
	tests := []struct {
		name  string
		index uint32
		want  int
	}{
		{"other, 0", 0, 0},
		{"front cover, 3", 3, 3},
		{"boundary, 20 does not degrade", 20, 20},
		{"just above the boundary, 21 degrades", 21, 0},
		{"far above the boundary, 255 degrades", 255, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := decodePicture(wellFormedPicturePayload(tt.index), 0)
			if err != nil {
				t.Fatalf("decodePicture: %v", err)
			}
			if p.Type != tt.want {
				t.Errorf("Type = %d, want %d", p.Type, tt.want)
			}
		})
	}
}

// TestParsePictureRefusals covers P3: each of the block's three
// declared lengths overdeclaring against its own payload extent, and
// a block shorter than the 32-byte minimum. Every row goes through
// Parse, the container-level entry point, and asserts both the
// sentinel and a zero-value container, since a malformed picture
// block refuses the whole file the way every other C7 shape does.
func TestParsePictureRefusals(t *testing.T) {
	data := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "media type length overdeclares",
			payload: picturePayload(3, 100, "image/png", uint32(len("cover")), "cover", uint32(len(data)), data),
		},
		{
			name:    "description length overdeclares",
			payload: picturePayload(3, uint32(len("image/png")), "image/png", 100, "cover", uint32(len(data)), data),
		},
		{
			name:    "data length overdeclares near the block header's own 3-byte maximum",
			payload: picturePayload(3, uint32(len("image/png")), "image/png", uint32(len("cover")), "cover", 0xFFFFFF, data),
		},
		{
			name:    "block shorter than the 32-byte minimum",
			payload: wellFormedPicturePayload(3)[:20],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := append(wellFormedBlocks()[:2], testBlock{Type: blockPicture, Payload: tt.payload})
			blocks = append(blocks, wellFormedBlocks()[3])
			path := writeFixture(t, buildContainer(blocks, nil))

			c, err := Parse(path)
			if err == nil {
				t.Fatalf("Parse: got nil error, want errors.Is(err, ErrMalformedContainer)")
			}
			if !errors.Is(err, ErrMalformedContainer) {
				t.Errorf("Parse: err = %v, want errors.Is(err, ErrMalformedContainer)", err)
			}
			if len(c.Blocks) != 0 || len(c.Pictures) != 0 || len(c.VorbisComment.Comments) != 0 || c.FrameOffset != 0 || c.FrameLength != 0 {
				t.Errorf("Parse: returned a partly filled container on refusal: %+v", c)
			}
		})
	}
}

// TestParseTwoPictureBlocks proves a second PICTURE block is legal
// (P2), not a refusal, and that both decode in on-disk order.
func TestParseTwoPictureBlocks(t *testing.T) {
	first := wellFormedPicturePayload(3)
	second := wellFormedPicturePayload(4)
	blocks := append(wellFormedBlocks()[:2],
		testBlock{Type: blockPicture, Payload: first},
		testBlock{Type: blockPicture, Payload: second},
		wellFormedBlocks()[3],
	)
	path := writeFixture(t, buildContainer(blocks, nil))

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.Pictures) != 2 {
		t.Fatalf("got %d pictures, want 2", len(c.Pictures))
	}
	if c.Pictures[0].Type != 3 || c.Pictures[1].Type != 4 {
		t.Errorf("Pictures = %+v, want types [3 4] in order", c.Pictures)
	}
}

// TestFrontCover covers P5's selection table: the single-candidate
// rule runs before the front-cover rule, so a lone back cover is the
// cover, and among several pictures the first typed front wins over a
// preceding back cover.
func TestFrontCover(t *testing.T) {
	front := Picture{Type: pictureTypeFront, Description: "front"}
	back := Picture{Type: 4, Description: "back"}
	other1 := Picture{Type: 0, Description: "other1"}
	other2 := Picture{Type: 5, Description: "other2"}
	other3 := Picture{Type: 6, Description: "other3"}
	frontAgain := Picture{Type: pictureTypeFront, Description: "front-again"}

	tests := []struct {
		name     string
		pictures []Picture
		want     Picture
		wantOK   bool
	}{
		{"no pictures", nil, Picture{}, false},
		{"one back cover, single-candidate rule selects it", []Picture{back}, back, true},
		{"back cover then front cover, front wins", []Picture{back, front}, front, true},
		{"two front covers, first wins", []Picture{front, frontAgain}, front, true},
		{"three non-front pictures, first wins", []Picture{other1, other2, other3}, other1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Container{Pictures: tt.pictures}
			got, ok := c.FrontCover()
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got.Description != tt.want.Description {
				t.Errorf("FrontCover = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// goldenPictureData is golden.txt:203-210's recorded picture data:
// PNG's eight-byte signature followed by the invented payload bytes
// gen_golden_flac.py wrote, transcribed from the provenance record
// rather than computed from this package's own decode.
var goldenPictureData = append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, "invented-picture-payload"...)

// TestPictureGoldenDecode parses testdata/golden.flac and checks its
// PICTURE block decodes field by field to the values golden.txt:203-210
// records, transcribed from that provenance record rather than
// produced by calling this package's own decode. It also checks the
// decoded field widths sum to the block's declared length of 78,
// which catches a field read at the wrong offset even when every
// individual value still looks plausible.
func TestPictureGoldenDecode(t *testing.T) {
	c, err := Parse(filepath.Join("testdata", "golden.flac"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.Pictures) != 1 {
		t.Fatalf("got %d pictures, want 1", len(c.Pictures))
	}
	p := c.Pictures[0]

	if p.Type != 3 {
		t.Errorf("Type = %d, want 3", p.Type)
	}
	if p.MediaType != "image/png" {
		t.Errorf("MediaType = %q, want %q", p.MediaType, "image/png")
	}
	if p.Description != "cover" {
		t.Errorf("Description = %q, want %q", p.Description, "cover")
	}
	if p.Width != 0 || p.Height != 0 || p.Depth != 0 || p.Colors != 0 {
		t.Errorf("geometry = {%d %d %d %d}, want all zero", p.Width, p.Height, p.Depth, p.Colors)
	}
	if string(p.Data) != string(goldenPictureData) {
		t.Errorf("Data = %x, want %x", p.Data, goldenPictureData)
	}

	// The block's own arithmetic: eight u(32) fields (32 bytes) plus
	// the three variable-length fields sums to the declared length of
	// 78 (golden.txt PICTURE block, c.Blocks[2].Length).
	sum := picturePayloadMin + len(p.MediaType) + len(p.Description) + len(p.Data)
	if sum != c.Blocks[2].Length {
		t.Errorf("field widths sum to %d, want %d (the block's declared length)", sum, c.Blocks[2].Length)
	}
}
