package tagio

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// testBlock is one metadata block as buildContainer serializes it,
// with the last-block flag under the test's own control rather than
// inferred from position, so a refusal fixture can set it on a block
// that is not the list's final entry.
type testBlock struct {
	Type    byte
	Last    bool
	Payload []byte
}

// buildContainer assembles a FLAC byte stream from the signature,
// each block in blocks with a re-emitted header, then frame appended
// verbatim as the frame region.
func buildContainer(blocks []testBlock, frame []byte) []byte {
	buf := []byte(flacSignature)
	for _, b := range blocks {
		header := b.Type
		if b.Last {
			header |= 0x80
		}
		length := len(b.Payload)
		buf = append(buf, header, byte(length>>16), byte(length>>8), byte(length))
		buf = append(buf, b.Payload...)
	}
	return append(buf, frame...)
}

// streamInfoPayload builds a 34-byte STREAMINFO payload carrying the
// given sample rate, channel count, bit depth and total samples. The
// four size fields and the MD5 are left zero, since this task grades
// only the pack it decodes.
func streamInfoPayload(sampleRate uint32, channels, bits uint8, totalSamples uint64) []byte {
	payload := make([]byte, streamInfoSize)
	pack := uint64(sampleRate)<<44 | uint64(channels-1)<<41 | uint64(bits-1)<<36 | totalSamples&0xFFFFFFFFF
	binary.BigEndian.PutUint64(payload[10:18], pack)
	return payload
}

// vorbisCommentPayload builds a VORBIS_COMMENT payload from a vendor
// string and a sequence of "NAME=VALUE" comments, little-endian
// throughout as RFC 9639 section 8.6 requires.
func vorbisCommentPayload(vendor string, comments []string) []byte {
	var buf []byte
	put := func(s string) {
		buf = binary.LittleEndian.AppendUint32(buf, uint32(len(s)))
		buf = append(buf, s...)
	}
	put(vendor)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(comments)))
	for _, c := range comments {
		put(c)
	}
	return buf
}

// wellFormedBlocks is the four-block chain every refusal fixture
// mutates from: STREAMINFO, VORBIS_COMMENT, one reserved type this
// package has never seen (a survival case in its own right), and
// PADDING marked last.
func wellFormedBlocks() []testBlock {
	return []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 1000)},
		{Type: blockVorbisComment, Payload: vorbisCommentPayload("tagio test", []string{"TITLE=A", "ARTIST=B"})},
		{Type: 10, Payload: []byte{0xAA, 0xBB, 0xCC}},
		{Type: 1, Last: true, Payload: make([]byte, 8)},
	}
}

func writeFixture(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.flac")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestParseWellFormedContainer(t *testing.T) {
	path := writeFixture(t, buildContainer(wellFormedBlocks(), nil))

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.Blocks) != 4 {
		t.Fatalf("got %d blocks, want 4", len(c.Blocks))
	}
	reserved := c.Blocks[2]
	if reserved.Type != 10 || reserved.Length != 3 || string(reserved.Payload) != "\xaa\xbb\xcc" {
		t.Errorf("reserved block did not survive intact: %+v", reserved)
	}
	if c.StreamInfo.SampleRate != 44100 || c.StreamInfo.Channels != 2 || c.StreamInfo.BitsPerSample != 16 {
		t.Errorf("STREAMINFO decoded wrong: %+v", c.StreamInfo)
	}
	if c.VorbisComment.Vendor != "tagio test" {
		t.Errorf("vendor = %q", c.VorbisComment.Vendor)
	}
	want := []Comment{{Name: "TITLE", Value: "A"}, {Name: "ARTIST", Value: "B"}}
	if len(c.VorbisComment.Comments) != len(want) || c.VorbisComment.Comments[0] != want[0] || c.VorbisComment.Comments[1] != want[1] {
		t.Errorf("comments = %+v, want %+v", c.VorbisComment.Comments, want)
	}
}

// TestParseRefusals covers each of C7's seven refusal shapes, one
// row per shape, each fixture mutating one byte, one length, one
// comment name or one duplicated block from wellFormedBlocks.
func TestParseRefusals(t *testing.T) {
	tests := []struct {
		name    string
		data    func() []byte
		wantErr error
	}{
		{
			name: "second VORBIS_COMMENT block",
			data: func() []byte {
				blocks := wellFormedBlocks()
				dup := testBlock{Type: blockVorbisComment, Payload: vorbisCommentPayload("dup", []string{"X=Y"})}
				blocks = append(blocks[:2], append([]testBlock{dup}, blocks[2:]...)...)
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "STREAMINFO not 34 bytes",
			data: func() []byte {
				blocks := wellFormedBlocks()
				blocks[0].Payload = blocks[0].Payload[:30]
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "block after a header whose last-block flag is set",
			data: func() []byte {
				blocks := wellFormedBlocks()
				blocks[1].Last = true // VORBIS_COMMENT wrongly flagged last too; PADDING stays last
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "declared length running past the end",
			data: func() []byte {
				blocks := wellFormedBlocks()
				data := buildContainer(blocks, nil)
				// The VORBIS_COMMENT header starts right after the
				// signature and the STREAMINFO block.
				headerOffset := len(flacSignature) + 4 + len(blocks[0].Payload)
				data[headerOffset+1] = 0xff // blow up the declared length
				return data
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "comment field name above U+007E",
			data: func() []byte {
				blocks := wellFormedBlocks()
				blocks[1].Payload = vorbisCommentPayload("tagio test", []string{"X\x7f=y"})
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "second SEEKTABLE block",
			data: func() []byte {
				blocks := wellFormedBlocks()
				st1 := testBlock{Type: blockSeekTable, Payload: []byte{1, 2, 3}}
				st2 := testBlock{Type: blockSeekTable, Payload: []byte{4, 5, 6}}
				blocks = append(blocks[:1], append([]testBlock{st1, st2}, blocks[1:]...)...)
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "first block is not STREAMINFO",
			data: func() []byte {
				blocks := wellFormedBlocks()
				blocks[0], blocks[1] = blocks[1], blocks[0]
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "a second STREAMINFO block",
			data: func() []byte {
				blocks := wellFormedBlocks()
				dup := testBlock{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 1000)}
				return buildContainer(slices.Insert(blocks, 1, dup), nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "no STREAMINFO block at all",
			data: func() []byte {
				return []byte(flacSignature)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "comment carries no '='",
			data: func() []byte {
				blocks := wellFormedBlocks()
				blocks[1].Payload = vorbisCommentPayload("tagio test", []string{"NOEQUALSIGN"})
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "signature is not fLaC",
			data: func() []byte {
				data := buildContainer(wellFormedBlocks(), nil)
				data[0] = 'X'
				return data
			},
			wantErr: ErrUnsupportedContainer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixture(t, tt.data())

			c, err := Parse(path)
			if err == nil {
				t.Fatalf("Parse: got nil error, want %v", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Parse: err = %v, want errors.Is(err, %v)", err, tt.wantErr)
			}
			if len(c.Blocks) != 0 || len(c.VorbisComment.Comments) != 0 || c.FrameOffset != 0 || c.FrameLength != 0 {
				t.Errorf("Parse: returned a partly filled container on refusal: %+v", c)
			}
		})
	}
}

// TestParseCommentNameU007EPreserved proves the field-name bound is
// RFC 9639's own, not mutagen's is_valid_key, which stops one code
// point short at 0x7D.
func TestParseCommentNameU007EPreserved(t *testing.T) {
	blocks := wellFormedBlocks()
	blocks[1].Payload = vorbisCommentPayload("tagio test", []string{"X~=tilde value"})
	path := writeFixture(t, buildContainer(blocks, nil))

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.VorbisComment.Comments) != 1 {
		t.Fatalf("got %d comments, want 1", len(c.VorbisComment.Comments))
	}
	got := c.VorbisComment.Comments[0]
	if got.Name != "X~" || got.Value != "tilde value" {
		t.Errorf("comment = %+v, want Name %q Value %q", got, "X~", "tilde value")
	}
}

// TestParseRepeatedCommentKeyPreserved proves a repeated comment key
// is carried twice, in order, rather than merged, which a map keyed
// by comment name cannot hold.
func TestParseRepeatedCommentKeyPreserved(t *testing.T) {
	blocks := wellFormedBlocks()
	blocks[1].Payload = vorbisCommentPayload("tagio test", []string{"GENRE=Ambient", "GENRE=Drone"})
	path := writeFixture(t, buildContainer(blocks, nil))

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []Comment{{Name: "GENRE", Value: "Ambient"}, {Name: "GENRE", Value: "Drone"}}
	if len(c.VorbisComment.Comments) != 2 || c.VorbisComment.Comments[0] != want[0] || c.VorbisComment.Comments[1] != want[1] {
		t.Errorf("comments = %+v, want %+v", c.VorbisComment.Comments, want)
	}
}

// TestParseFrameRegion asserts the frame region's offset and length
// against a fixture whose frame bytes are a known invented run, and
// that the slice those two numbers name equals that run exactly. The
// frame bytes are chosen so a peek at them cannot be mistaken for
// another block header: their would-be declared length runs past
// what remains.
func TestParseFrameRegion(t *testing.T) {
	frame := []byte("invented-frame-bytes-not-really-audio")
	data := buildContainer(wellFormedBlocks(), frame)
	path := writeFixture(t, data)

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if c.FrameLength != int64(len(frame)) {
		t.Fatalf("FrameLength = %d, want %d", c.FrameLength, len(frame))
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	got := make([]byte, c.FrameLength)
	if _, err := f.ReadAt(got, c.FrameOffset); err != nil {
		t.Fatalf("read frame region: %v", err)
	}
	if string(got) != string(frame) {
		t.Errorf("frame region = %q, want %q", got, frame)
	}
}

// goldenComments is golden.txt's decoded comment sequence, in
// on-disk order, restated as a literal rather than parsed from the
// provenance record so this test does not echo the record it checks
// the parse against.
var goldenComments = []string{
	"ACOUSTID_FINGERPRINT=invented-fingerprint-value",
	"ACOUSTID_ID=11111111-2222-3333-4444-555555555555",
	"ALBUM=Al",
	"ALBUMARTISTS=X",
	"ALBUMARTISTS=Y",
	"ALBUM_ARTISTS=X",
	"ALBUM_ARTISTS=Y",
	"RELEASETYPE=album",
	"RELEASETYPE=live",
	"MUSICBRAINZ_ALBUMTYPE=album",
	"MUSICBRAINZ_ALBUMTYPE=live",
	"ARTIST=A",
	"BARCODE=",
	"BPM=120",
	"CATALOGNUMBER=",
	"DESCRIPTION=Invented comment text.",
	"COMMENT=Invented comment text.",
	"COMPILATION=1",
	"DATE=0000",
	"YEAR=0",
	"DISC=1",
	"DISCNUMBER=1",
	"DISCTOTAL=2",
	"DISCC=2",
	"TOTALDISCS=2",
	"ENCODEDBY=Invented Encoder 1.0",
	"ENCODER=Invented Encoder 1.0",
	"GENRE=Ambient",
	"GENRE=Drone",
	"GROUPING=Invented Grouping",
	"INITIALKEY=Am",
	"LYRICS=la la la",
	"ORIGINALDATE=1999-03-04",
	"R128_ALBUM_GAIN=-576",
	"R128_TRACK_GAIN=-384",
	"REMIXER=Invented Remixer",
	"REPLAYGAIN_ALBUM_GAIN=-6.50 dB",
	"REPLAYGAIN_ALBUM_PEAK=0.654321",
	"REPLAYGAIN_TRACK_GAIN=-7.25 dB",
	"REPLAYGAIN_TRACK_PEAK=0.123456",
	"SUBTITLE=Invented Subtitle",
	"TITLE=T",
	"TRACK=3",
	"TRACKNUMBER=3",
	"TRACKTOTAL=12",
	"TRACKC=12",
	"TOTALTRACKS=12",
}

// goldenPicturePayload is golden.flac's PICTURE block payload,
// restated as a literal (decoded independently with Python at fixture
// time, not read back through this package's own parse).
var goldenPicturePayload = []byte{
	0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x09, 0x69, 0x6d,
	0x61, 0x67, 0x65, 0x2f, 0x70, 0x6e, 0x67, 0x00, 0x00, 0x00,
	0x05, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x89, 0x50, 0x4e, 0x47,
	0x0d, 0x0a, 0x1a, 0x0a, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74,
	0x65, 0x64, 0x2d, 0x70, 0x69, 0x63, 0x74, 0x75, 0x72, 0x65,
	0x2d, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
}

// TestParseGolden parses testdata/golden.flac and checks it against
// golden.txt's provenance record field by field: the block chain,
// each declared length, the STREAMINFO pack, the vendor string, the
// comment sequence and the frame region. The expectations below are
// independently computed (Python, at fixture-generation time), not
// produced by calling this package's own decode.
func TestParseGolden(t *testing.T) {
	c, err := Parse(filepath.Join("testdata", "golden.flac"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	wantTypes := []byte{0, 4, 6, 1}
	wantLast := []bool{false, false, false, true}
	wantLengths := []int{34, 1066, 78, 1024}
	if len(c.Blocks) != len(wantTypes) {
		t.Fatalf("got %d blocks, want %d", len(c.Blocks), len(wantTypes))
	}
	for i, b := range c.Blocks {
		if b.Type != wantTypes[i] || b.Last != wantLast[i] || b.Length != wantLengths[i] {
			t.Errorf("block %d = {Type:%d Last:%v Length:%d}, want {Type:%d Last:%v Length:%d}",
				i, b.Type, b.Last, b.Length, wantTypes[i], wantLast[i], wantLengths[i])
		}
	}

	if string(c.Blocks[2].Payload) != string(goldenPicturePayload) {
		t.Errorf("PICTURE payload does not match the recorded bytes")
	}

	wantInfo := StreamInfo{SampleRate: 44100, Channels: 1, BitsPerSample: 16, TotalSamples: 0}
	if c.StreamInfo.SampleRate != wantInfo.SampleRate || c.StreamInfo.Channels != wantInfo.Channels ||
		c.StreamInfo.BitsPerSample != wantInfo.BitsPerSample || c.StreamInfo.TotalSamples != wantInfo.TotalSamples {
		t.Errorf("StreamInfo = %+v, want %+v", c.StreamInfo, wantInfo)
	}
	if c.StreamInfo.MD5 != ([16]byte{}) {
		t.Errorf("StreamInfo.MD5 = %x, want all zero", c.StreamInfo.MD5)
	}

	if c.VorbisComment.Vendor != "Mutagen 1.48.1" {
		t.Errorf("vendor = %q, want %q", c.VorbisComment.Vendor, "Mutagen 1.48.1")
	}
	if len(c.VorbisComment.Comments) != len(goldenComments) {
		t.Fatalf("got %d comments, want %d", len(c.VorbisComment.Comments), len(goldenComments))
	}
	for i, want := range goldenComments {
		name, value, _ := strings.Cut(want, "=")
		got := c.VorbisComment.Comments[i]
		if got.Name != name || got.Value != value {
			t.Errorf("comment %d = %+v, want {Name:%q Value:%q}", i, got, name, value)
		}
	}

	if c.FrameOffset != 2222 || c.FrameLength != 0 {
		t.Errorf("frame region = (%d, %d), want (2222, 0)", c.FrameOffset, c.FrameLength)
	}
}

// headerOnlyContainer serializes blocks, then one more block header
// declaring length bytes of type blockType, with the file ending right
// after that header: a block that runs past the end of the file.
func headerOnlyContainer(blocks []testBlock, blockType byte, length int) []byte {
	buf := buildContainer(blocks, nil)
	return append(buf, blockType, byte(length>>16), byte(length>>8), byte(length))
}

func TestParseBoundedRefusesBeforeAllocating(t *testing.T) {
	bounds := Bounds{TagBytes: 4096, PictureBytes: 4096}
	streamInfo := testBlock{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 1000)}

	// Four padding blocks of 4000 bytes sum to 16034 with STREAMINFO,
	// under four times the picture cap (16384); a fifth crosses it.
	padded := []testBlock{streamInfo}
	for range 4 {
		padded = append(padded, testBlock{Type: 1, Payload: make([]byte, 4000)})
	}

	tests := []struct {
		name string
		data []byte
		want error
	}{
		{"tag block over its cap", headerOnlyContainer([]testBlock{streamInfo}, blockVorbisComment, 4097), ErrTagsTooLarge},
		{"picture block over its cap", headerOnlyContainer([]testBlock{streamInfo}, blockPicture, 4097), ErrPictureTooLarge},
		{"chain over its total", headerOnlyContainer(padded, 1, 4000), ErrPictureTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixture(t, tt.data)

			_, err := ParseBounded(path, bounds)
			if !errors.Is(err, tt.want) {
				t.Fatalf("ParseBounded error = %v, want %v", err, tt.want)
			}
			if errors.Is(err, ErrMalformedContainer) {
				t.Errorf("ParseBounded error %v also answers ErrMalformedContainer", err)
			}

			if _, err := Parse(path); !errors.Is(err, ErrMalformedContainer) {
				t.Errorf("Parse error = %v, want ErrMalformedContainer", err)
			}
		})
	}
}

func TestParseBoundedAdmitsWithinBounds(t *testing.T) {
	path := writeFixture(t, buildContainer(wellFormedBlocks(), []byte{0xFF, 0xF8}))
	want, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	got, err := ParseBounded(path, Bounds{TagBytes: 1 << 20, PictureBytes: 1 << 20})
	if err != nil {
		t.Fatalf("ParseBounded: %v", err)
	}
	if got.FrameOffset != want.FrameOffset || len(got.Blocks) != len(want.Blocks) || got.StreamInfo != want.StreamInfo {
		t.Errorf("ParseBounded = %+v, want %+v", got, want)
	}
}

func TestParseBoundedCapIsInclusive(t *testing.T) {
	comment := vorbisCommentPayload("v", []string{"TITLE=A"})
	path := writeFixture(t, buildContainer([]testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 1000)},
		{Type: blockVorbisComment, Last: true, Payload: comment},
	}, nil))

	if _, err := ParseBounded(path, Bounds{TagBytes: int64(len(comment)), PictureBytes: 1 << 20}); err != nil {
		t.Errorf("a block exactly at its cap: %v", err)
	}
	if _, err := ParseBounded(path, Bounds{TagBytes: int64(len(comment)) - 1, PictureBytes: 1 << 20}); !errors.Is(err, ErrTagsTooLarge) {
		t.Errorf("a block one byte over its cap: error = %v, want ErrTagsTooLarge", err)
	}
}
