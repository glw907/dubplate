package shadow

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// buildFLAC assembles a minimal FLAC container in memory: the fLaC
// signature, a STREAMINFO block carrying rate and samples, and a
// VORBIS_COMMENT block carrying tags in the order given. It is the
// fixture builder the whole package shares, and it is a builder rather
// than a vendored file so a mutation is a byte edit rather than a
// re-derivation (rung 4 spec, the FLAC reader's own cost section).
//
// The stream is always mono 16-bit, which is the shape the committed
// oracle template carries, so the only fixture axes are the tags, the
// sample count and the rate.
func buildFLAC(t *testing.T, tags []string, samples uint64, rate uint32) []byte {
	t.Helper()

	const (
		channels      = 1
		bitsPerSample = 16
	)
	pack := uint64(rate)<<44 | uint64(channels-1)<<41 | uint64(bitsPerSample-1)<<36 | samples&0xFFFFFFFFF

	streaminfo := make([]byte, 34)
	binary.BigEndian.PutUint64(streaminfo[10:18], pack)

	comment := binary.LittleEndian.AppendUint32(nil, 0) // an empty vendor string
	comment = binary.LittleEndian.AppendUint32(comment, uint32(len(tags)))
	for _, tag := range tags {
		comment = binary.LittleEndian.AppendUint32(comment, uint32(len(tag)))
		comment = append(comment, tag...)
	}

	out := []byte("fLaC")
	out = append(out, blockHeader(t, 0, false, len(streaminfo))...)
	out = append(out, streaminfo...)
	out = append(out, blockHeader(t, 4, true, len(comment))...)
	return append(out, comment...)
}

// blockHeader builds one metadata block header: the last-block flag in
// the high bit of the type byte, then a three-byte big-endian payload
// length.
func blockHeader(t *testing.T, blockType byte, last bool, length int) []byte {
	t.Helper()

	first := blockType
	if last {
		first |= 0x80
	}
	return []byte{first, byte(length >> 16), byte(length >> 8), byte(length)}
}

// writeTemp writes data to a file in the test's own temp directory and
// returns its path.
func writeTemp(t *testing.T, data []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "track.flac")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestReadFLACMapsEveryVorbisKey(t *testing.T) {
	tags := []string{
		"TITLE=Sub Rosa", "ARTIST=Cabinet", "ALBUM=Windrose",
		"ALBUMARTIST=Cabinet Ensemble",
		"TRACKNUMBER=3/12", "DISCNUMBER=2", "TOTALDISCS=2",
		"COMPILATION=1", "PUBLISHER=Nightwood", "CATALOGNUMBER=NW-004",
		"BARCODE=5901234123457", "RELEASECOUNTRY=GB", "MEDIA=CD",
		"MUSICBRAINZ_ALBUMCOMMENT=deluxe",
		"MUSICBRAINZ_TRACKID=11111111-1111-1111-1111-111111111111",
		"MUSICBRAINZ_ALBUMID=22222222-2222-2222-2222-222222222222",
		"DATE=1994-05-01",
	}
	path := writeTemp(t, buildFLAC(t, tags, 441000, 44100))

	got, err := readFLAC(path)
	if err != nil {
		t.Fatalf("readFLAC: %v", err)
	}
	want := matcher.ItemFields{
		Title: "Sub Rosa", Artist: "Cabinet", Album: "Windrose",
		AlbumArtist: "Cabinet Ensemble",
		Track:       3, Disc: 2, DiscTotal: 2, Comp: true,
		Label: "Nightwood", CatalogNum: "NW-004", Barcode: "5901234123457",
		Country: "GB", Media: "CD", AlbumDisambig: "deluxe",
		MBTrackID: "11111111-1111-1111-1111-111111111111",
		MBAlbumID: "22222222-2222-2222-2222-222222222222",
		Year:      1994, Length: 10,
		// DataSource stays "": beets keeps it as a flexible attribute
		// rather than in the file, so no Vorbis key exists.
	}
	if got != want {
		t.Errorf("readFLAC =\n%+v\nwant\n%+v", got, want)
	}
}

func TestReadFLACStylePrecedence(t *testing.T) {
	// Every row builds a FLAC at 44100 Hz with zero samples, so Length
	// stays 0 and the fixture stays a container pinned mutagen would
	// open: a STREAMINFO sample rate of 0 is refused outright
	// (mutagen/flac.py:222-223), so a rate-0 fixture would be a file
	// the oracle itself could not read.
	tests := []struct {
		name string
		tags []string
		want matcher.ItemFields
	}{
		{
			name: "album artist takes the first declared style",
			tags: []string{"ALBUM ARTIST=First", "ALBUM_ARTIST=Second", "ALBUMARTIST=Third"},
			want: matcher.ItemFields{AlbumArtist: "First"},
		},
		{
			name: "an empty first style falls through to the next",
			tags: []string{"ALBUM ARTIST=", "ALBUM_ARTIST=Second", "ALBUMARTIST=Third"},
			want: matcher.ItemFields{AlbumArtist: "Second"},
		},
		{
			name: "track prefers TRACK over TRACKNUMBER",
			tags: []string{"TRACKNUMBER=9", "TRACK=4"},
			want: matcher.ItemFields{Track: 4},
		},
		{
			name: "a zero first style does not fall through",
			tags: []string{"TRACK=0", "TRACKNUMBER=5"},
			want: matcher.ItemFields{},
		},
		{
			name: "disc prefers DISC over DISCNUMBER",
			tags: []string{"DISCNUMBER=9", "DISC=2"},
			want: matcher.ItemFields{Disc: 2},
		},
		{
			name: "disctotal walks its three styles in order",
			tags: []string{"TOTALDISCS=9", "DISCC=3"},
			want: matcher.ItemFields{DiscTotal: 3},
		},
		{
			name: "label falls through to PUBLISHER",
			tags: []string{"PUBLISHER=Nightwood"},
			want: matcher.ItemFields{Label: "Nightwood"},
		},
		{
			name: "catalognum falls through to a read-only alias",
			tags: []string{"CATALOGID=RO-1"},
			want: matcher.ItemFields{CatalogNum: "RO-1"},
		},
		{
			name: "barcode falls through the four read-only aliases in order",
			tags: []string{"EAN=E-1", "UPN=U-1"},
			want: matcher.ItemFields{Barcode: "E-1"},
		},
		{
			name: "year falls back to YEAR when DATE has no year component",
			tags: []string{"DATE=", "YEAR=1979"},
			want: matcher.ItemFields{Year: 1979},
		},
		{
			name: "a DATE with no parseable year yields zero",
			tags: []string{"DATE=circa-05-01", "YEAR=1979"},
			want: matcher.ItemFields{},
		},
		{
			name: "keys are case-insensitive",
			tags: []string{"title=lowercase key"},
			want: matcher.ItemFields{Title: "lowercase key"},
		},
		{
			name: "the first occurrence of a repeated key wins",
			tags: []string{"TITLE=first", "TITLE=second"},
			want: matcher.ItemFields{Title: "first"},
		},
		{
			name: "an unparseable int takes the zero value",
			tags: []string{"TRACKNUMBER=A-side"},
			want: matcher.ItemFields{},
		},
		{
			name: "an int takes the leading digits of a slashed pair",
			tags: []string{"TRACKNUMBER=3/12"},
			want: matcher.ItemFields{Track: 3},
		},
		{
			name: "a non-numeric compilation is false",
			tags: []string{"COMPILATION=yes"},
			want: matcher.ItemFields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFLAC(writeTemp(t, buildFLAC(t, tt.tags, 0, 44100)))
			if err != nil {
				t.Fatalf("readFLAC: %v", err)
			}
			if got != tt.want {
				t.Errorf("readFLAC =\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}
}

func TestReadFLACLengthFromStreaminfo(t *testing.T) {
	tests := []struct {
		name    string
		samples uint64
		rate    uint32
		want    float64
	}{
		{name: "ten seconds at 44100 Hz", samples: 441000, rate: 44100, want: 10},
		{name: "a fractional length", samples: 22050, rate: 44100, want: 0.5},
		{name: "a zero rate is zero, never a division by zero", samples: 441000, rate: 0, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFLAC(writeTemp(t, buildFLAC(t, nil, tt.samples, tt.rate)))
			if err != nil {
				t.Fatalf("readFLAC: %v", err)
			}
			if got.Length != tt.want {
				t.Errorf("Length = %v, want %v", got.Length, tt.want)
			}
		})
	}
}

// committedTemplateFLAC is the 42 bytes of
// internal/harness/oracle/testdata/template.flac, embedded here as a Go
// byte literal rather than read from that path: it is harness territory
// and stays unread by the engine (rung 4 spec, the FLAC reader's
// fixture section). Its provenance is that directory's
// gen_template_flac.py, which builds it through mutagen's own
// StreamInfo.write as mono, 44100 Hz, 16-bit, zero samples.
var committedTemplateFLAC = []byte{
	'f', 'L', 'a', 'C',
	0x80, 0x00, 0x00, 0x22, // the last block, type 0, 34 bytes of payload
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x0a, 0xc4, 0x40, 0xf0, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// TestReadFLACParsesTheCommittedTemplateBytes and
// TestBuildFLACEmitsTheStreaminfoLayoutByteForByte are the two
// assertions that make the container format checkable at all. Every
// other test in this file reads bytes buildFLAC produced, so the
// builder and the parser could agree on a wrong layout forever; these
// two pin each of them to something outside this package.
func TestReadFLACParsesTheCommittedTemplateBytes(t *testing.T) {
	got, err := parseFLAC(committedTemplateFLAC)
	if err != nil {
		t.Fatalf("parseFLAC: %v", err)
	}
	want := streamInfo{SampleRate: 44100, Channels: 1, BitsPerSample: 16, TotalSamples: 0}
	if got.info != want {
		t.Errorf("stream info = %+v, want %+v", got.info, want)
	}
	if len(got.tags) != 0 {
		t.Errorf("tags = %v, want none", got.tags)
	}
}

func TestBuildFLACEmitsTheStreaminfoLayoutByteForByte(t *testing.T) {
	// A hand-written expectation derived from the format, not from the
	// builder: ten zero bytes of block and frame sizes, then the pack
	// (44100 Hz over twenty bits, mono, sixteen bits per sample, and
	// 441000 total samples over thirty-six bits), then a zero MD5.
	want := []byte{
		0x00, 0x00, // minimum block size
		0x00, 0x00, // maximum block size
		0x00, 0x00, 0x00, // minimum frame size
		0x00, 0x00, 0x00, // maximum frame size
		0x0a, 0xc4, 0x40, 0xf0, 0x00, 0x06, 0xba, 0xa8, // the pack
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // the MD5 signature
	}

	data := buildFLAC(t, nil, 441000, 44100)
	const payloadStart = len("fLaC") + 4
	got := data[payloadStart : payloadStart+34]
	if string(got) != string(want) {
		t.Errorf("STREAMINFO payload =\n% x\nwant\n% x", got, want)
	}
}

func TestReadFLACRefusesAMalformedContainer(t *testing.T) {
	// A VORBIS_COMMENT block whose single comment declares more bytes
	// than the block carries.
	truncatedComment := binary.LittleEndian.AppendUint32(nil, 0)
	truncatedComment = binary.LittleEndian.AppendUint32(truncatedComment, 1)
	truncatedComment = binary.LittleEndian.AppendUint32(truncatedComment, 100)
	truncatedComment = append(truncatedComment, "TITLE=x"...)

	tests := []struct {
		name string
		data []byte
	}{
		{name: "not a FLAC", data: []byte("OggS....")},
		{name: "truncated block header", data: []byte("fLaC\x00\x00")},
		{name: "block length past the end", data: []byte("fLaC\x00\x00\x00\xff")},
		{
			name: "truncated vorbis comment payload",
			data: append([]byte("fLaC"), append(blockHeader(t, 4, true, len(truncatedComment)), truncatedComment...)...),
		},
		{
			name: "streaminfo shorter than 34 bytes",
			data: append([]byte("fLaC"), append(blockHeader(t, 0, true, 12), make([]byte, 12)...)...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFLAC(writeTemp(t, tt.data))
			if err == nil {
				t.Fatalf("readFLAC = %+v, want an error", got)
			}
			if got != (matcher.ItemFields{}) {
				t.Errorf("readFLAC = %+v, want the zero fields on an error", got)
			}
		})
	}
}

func TestItemBagRoundTripsThroughProjectItemFields(t *testing.T) {
	// Every one of the nineteen fields carries a distinguishable
	// non-zero value, so a name this table spells wrong shows up as a
	// dropped field rather than as a coincidence.
	want := matcher.ItemFields{
		Artist:        "Cabinet",
		Album:         "Windrose",
		AlbumArtist:   "Cabinet Ensemble",
		Year:          1994,
		DiscTotal:     2,
		MBAlbumID:     "22222222-2222-2222-2222-222222222222",
		Label:         "Nightwood",
		Barcode:       "5901234123457",
		CatalogNum:    "NW-004",
		Country:       "GB",
		Media:         "CD",
		AlbumDisambig: "deluxe",
		DataSource:    "MusicBrainz",
		Length:        612.5,
		Title:         "Sub Rosa",
		Track:         3,
		MBTrackID:     "11111111-1111-1111-1111-111111111111",
		Disc:          2,
		Comp:          true,
	}
	if got := matcher.ProjectItemFields(itemBag(want)); got != want {
		t.Errorf("round trip =\n%+v\nwant\n%+v", got, want)
	}
}
