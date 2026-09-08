package tagio

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// goldenTagSet is the tag set golden.txt's INPUT section records,
// field by field, transcribed from the provenance record rather than
// computed from anything golden.flac's bytes or this package's code
// produce. It is the input a write of the golden's own recorded input
// applies to a copy of the golden (R4).
func goldenTagSet() TagSet {
	return TagSet{
		Album:               new("Al"),
		Albumartists:        []string{"X", "Y"},
		Artist:              new("A"),
		Bpm:                 new(120),
		Comp:                new(true),
		Disc:                new(1),
		Disctotal:           new(2),
		Title:               new("T"),
		Track:               new(3),
		Tracktotal:          new(12),
		Albumtype:           new("album"),
		Albumtypes:          []string{"album", "live"},
		Year:                new(0),
		OriginalYear:        new(1999),
		OriginalMonth:       new(3),
		OriginalDay:         new(4),
		Arrangers:           EmptyList(),
		Catalognum:          EmptyString(),
		Barcode:             EmptyString(),
		AcoustIDFingerprint: new("invented-fingerprint-value"),
		AcoustIDID:          new("11111111-2222-3333-4444-555555555555"),
		Comments:            new("Invented comment text."),
		Encoder:             new("Invented Encoder 1.0"),
		Genres:              []string{"Ambient", "Drone"},
		Grouping:            new("Invented Grouping"),
		InitialKey:          new("Am"),
		Lyrics:              new("la la la"),
		R128AlbumGain:       new(-2.25),
		R128TrackGain:       new(-1.5),
		Remixers:            []string{"Invented Remixer"},
		RgAlbumGain:         new(-6.5),
		RgAlbumPeak:         new(0.654321),
		RgTrackGain:         new(-7.25),
		RgTrackPeak:         new(0.123456),
		Subtitle:            new("Invented Subtitle"),
	}
}

// copyFixture copies src into a fresh file under t.TempDir(), so a
// test that rewrites a fixture never touches the committed original.
func copyFixture(t *testing.T, src string) string {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	return writeFixture(t, data)
}

// blockHeaderAndPayload reads path's raw bytes and slices out the
// header and payload of the block at index in that file's own block
// chain, computed by summing the declared lengths Parse already
// decoded rather than by re-deriving them.
func blockHeaderAndPayload(t *testing.T, path string, index int) (header, payload []byte) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse %s: %v", path, err)
	}
	offset := int64(len(flacSignature))
	for i, b := range c.Blocks {
		if i == index {
			return data[offset : offset+4], data[offset+4 : offset+4+int64(b.Length)]
		}
		offset += 4 + int64(b.Length)
	}
	t.Fatalf("%s: block index %d out of range, chain carries %d blocks", path, index, len(c.Blocks))
	return nil, nil
}

// frameHash digests path's frame region alone, at the offset and
// length Parse records, so a test can compare it before and after a
// write without comparing the rest of the file.
func frameHash(t *testing.T, path string) [32]byte {
	t.Helper()
	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse %s: %v", path, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return sha256.Sum256(data[c.FrameOffset : c.FrameOffset+c.FrameLength])
}

// TestWriteGoldenBytes is R4: a write of the golden's own recorded
// input, applied to a copy of the golden, emits a VORBIS_COMMENT
// header and payload byte-identical to the golden's own. The header
// comparison pins the three-byte big-endian length and the cleared
// last-block flag together, C11's asymmetry (risk 2b).
func TestWriteGoldenBytes(t *testing.T) {
	goldenPath := filepath.Join("testdata", "golden.flac")
	copyPath := copyFixture(t, goldenPath)

	if err := Write(copyPath, goldenTagSet(), 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}

	wantHeader, wantPayload := blockHeaderAndPayload(t, goldenPath, 1)
	gotHeader, gotPayload := blockHeaderAndPayload(t, copyPath, 1)
	if !bytes.Equal(wantHeader, gotHeader) {
		t.Errorf("VORBIS_COMMENT header = % x, want % x", gotHeader, wantHeader)
	}
	if !bytes.Equal(wantPayload, gotPayload) {
		t.Errorf("VORBIS_COMMENT payload differs from the golden's")
	}
}

// TestWriteCommentSet asserts the emitted comment set key by key
// against golden.txt's recorded sequence (INV:409-427 is the shape
// that record covers, not the expectation itself), including the
// absence of a key for an absent value and the absence of ALBUM
// ARTISTS, a read-only style, anywhere in the output.
func TestWriteCommentSet(t *testing.T) {
	copyPath := copyFixture(t, filepath.Join("testdata", "golden.flac"))
	if err := Write(copyPath, goldenTagSet(), 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}

	c, err := Parse(copyPath)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := commentsFrom(goldenComments...)
	if !reflect.DeepEqual(c.VorbisComment.Comments, want) {
		t.Errorf("comment set =\n%+v\nwant\n%+v", c.VorbisComment.Comments, want)
	}
	for _, c := range c.VorbisComment.Comments {
		if c.Name == "ALBUM ARTISTS" {
			t.Errorf("read-only style ALBUM ARTISTS present in a write's output")
		}
	}
}

// TestWriteRoundTripEveryBlockType covers a container carrying one
// block of every type, 0 through 6, plus one reserved type: every
// preserved payload survives byte for byte, in the original relative
// order, under a re-emitted header. buildContainer is this task's one
// stated exception to "no helper builds a container the writer also
// builds", because no well-formed round-trip fixture can carry a
// reserved block type any other way.
func TestWriteRoundTripEveryBlockType(t *testing.T) {
	streamInfo := streamInfoPayload(48000, 2, 16, 1000)
	application := []byte("APPLpayload")
	seekTable := []byte{1, 2, 3, 4, 5}
	comment := vorbisCommentPayload("existing vendor", []string{"TITLE=old"})
	cueSheet := []byte("cuesheetpayload")
	picture := []byte("picturepayload")
	reserved := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfo},
		{Type: blockPadding, Payload: make([]byte, 16)},
		{Type: 2, Payload: application},
		{Type: blockSeekTable, Payload: seekTable},
		{Type: blockVorbisComment, Payload: comment},
		{Type: 5, Payload: cueSheet},
		{Type: 6, Payload: picture},
		{Type: 50, Last: true, Payload: reserved},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	if err := Write(path, TagSet{Title: new("new")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	wantOrder := []byte{blockStreamInfo, 2, blockSeekTable, blockVorbisComment, 5, 6, 50, blockPadding}
	if len(c.Blocks) != len(wantOrder) {
		t.Fatalf("got %d blocks, want %d: %+v", len(c.Blocks), len(wantOrder), c.Blocks)
	}
	for i, b := range c.Blocks {
		if b.Type != wantOrder[i] {
			t.Errorf("block %d: type = %d, want %d", i, b.Type, wantOrder[i])
		}
	}
	if !bytes.Equal(c.Blocks[0].Payload, streamInfo) {
		t.Errorf("STREAMINFO payload changed")
	}
	if !bytes.Equal(c.Blocks[1].Payload, application) {
		t.Errorf("APPLICATION payload changed")
	}
	if !bytes.Equal(c.Blocks[2].Payload, seekTable) {
		t.Errorf("SEEKTABLE payload changed")
	}
	if !bytes.Equal(c.Blocks[4].Payload, cueSheet) {
		t.Errorf("CUESHEET payload changed")
	}
	if !bytes.Equal(c.Blocks[5].Payload, picture) {
		t.Errorf("PICTURE payload changed")
	}
	if !bytes.Equal(c.Blocks[6].Payload, reserved) {
		t.Errorf("reserved-type payload changed")
	}
}

// TestWriteForeignKeyPreserved covers a fixture whose comment block
// carries a foreign key, one of the 69 never own: COPYRIGHT and
// ORGANIZATION, the measured case (spec:360-364). A write asserts that
// key intact, with its value and its position, afterward.
func TestWriteForeignKeyPreserved(t *testing.T) {
	comment := vorbisCommentPayload("vendor", []string{"COPYRIGHT=(c) someone", "TITLE=old", "ORGANIZATION=a label"})
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 1, 16, 0)},
		{Type: blockVorbisComment, Payload: comment},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	if err := Write(path, TagSet{Title: new("new")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// TITLE is one of the 69 keys, so its delete removes the loaded
	// TITLE=old wherever it sat and its write lands at the running
	// end, after the two foreign keys neither delete nor write ever
	// touches; COPYRIGHT and ORGANIZATION keep both their value and
	// their relative order.
	want := []Comment{{Name: "COPYRIGHT", Value: "(c) someone"}, {Name: "ORGANIZATION", Value: "a label"}, {Name: "TITLE", Value: "new"}}
	if !reflect.DeepEqual(c.VorbisComment.Comments, want) {
		t.Errorf("comments = %+v, want %+v", c.VorbisComment.Comments, want)
	}
}

// TestWriteReplacesLowercaseTitleCaseInsensitively is C8's
// case-insensitive rule: a fixture whose comment block carries
// title=old in lowercase is replaced, rather than duplicated, by a
// write of the title field, leaving exactly one title comment.
func TestWriteReplacesLowercaseTitleCaseInsensitively(t *testing.T) {
	comment := vorbisCommentPayload("vendor", []string{"title=old"})
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(96000, 2, 24, 0)},
		{Type: blockVorbisComment, Payload: comment},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	if err := Write(path, TagSet{Title: new("new")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []Comment{{Name: "TITLE", Value: "new"}}
	if !reflect.DeepEqual(c.VorbisComment.Comments, want) {
		t.Errorf("comments = %+v, want %+v (exactly one title comment)", c.VorbisComment.Comments, want)
	}
}

// TestWriteCollidingPairLeavesPluralComments is INV:306-308's own
// measurement: writing both members of each colliding pair in one
// call leaves the plural's comments. All three colliding pairs are
// covered here at the surviving-comments level that Write owns;
// tagset_test.go covers WriteOrder only.
func TestWriteCollidingPairLeavesPluralComments(t *testing.T) {
	tests := []struct {
		name string
		ts   TagSet
		want []Comment
	}{
		{
			name: "Albumtype/Albumtypes",
			ts:   TagSet{Albumtype: new("album"), Albumtypes: []string{"album", "live"}},
			want: []Comment{
				{Name: "RELEASETYPE", Value: "album"},
				{Name: "RELEASETYPE", Value: "live"},
				{Name: "MUSICBRAINZ_ALBUMTYPE", Value: "album"},
				{Name: "MUSICBRAINZ_ALBUMTYPE", Value: "live"},
			},
		},
		{
			name: "MBArtistid/MBArtistids",
			ts:   TagSet{MBArtistid: new("single-id"), MBArtistids: []string{"id-1", "id-2"}},
			want: []Comment{
				{Name: "MUSICBRAINZ_ARTISTID", Value: "id-1"},
				{Name: "MUSICBRAINZ_ARTISTID", Value: "id-2"},
			},
		},
		{
			name: "MBAlbumartistid/MBAlbumartistids",
			ts:   TagSet{MBAlbumartistid: new("single-id"), MBAlbumartistids: []string{"id-1", "id-2"}},
			want: []Comment{
				{Name: "MUSICBRAINZ_ALBUMARTISTID", Value: "id-1"},
				{Name: "MUSICBRAINZ_ALBUMARTISTID", Value: "id-2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := []testBlock{
				{Type: blockStreamInfo, Payload: streamInfoPayload(22050, 1, 8, 0)},
				{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
			}
			path := writeFixture(t, buildContainer(blocks, nil))

			if err := Write(path, tt.ts, 0o644); err != nil {
				t.Fatalf("Write: %v", err)
			}

			c, err := Parse(path)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !reflect.DeepEqual(c.VorbisComment.Comments, tt.want) {
				t.Errorf("comments = %+v, want %+v", c.VorbisComment.Comments, tt.want)
			}
		})
	}
}

// TestWriteBlockChainPaddingPolicy covers rulings 5 and 6: the block
// chain after a write reads exactly one PADDING block, 8 KiB, last,
// and writing twice over the same file yields the same padding as
// writing once. Every incoming PADDING block is dropped.
func TestWriteBlockChainPaddingPolicy(t *testing.T) {
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 24, 0)},
		{Type: blockPadding, Payload: make([]byte, 4096)},
		{Type: blockPadding, Last: true, Payload: make([]byte, 128)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	if err := Write(path, TagSet{Title: new("A")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}
	sizeOnce, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if err := Write(path, TagSet{Title: new("A")}, 0o644); err != nil {
		t.Fatalf("second Write: %v", err)
	}
	sizeTwice, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if sizeOnce.Size() != sizeTwice.Size() {
		t.Errorf("file size after two writes = %d, after one write = %d", sizeTwice.Size(), sizeOnce.Size())
	}

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	last := c.Blocks[len(c.Blocks)-1]
	paddingCount := 0
	for _, b := range c.Blocks {
		if b.Type == blockPadding {
			paddingCount++
		}
	}
	if paddingCount != 1 {
		t.Errorf("PADDING block count = %d, want 1", paddingCount)
	}
	if last.Type != blockPadding || !last.Last || last.Length != paddingSize {
		t.Errorf("trailing block = %+v, want PADDING, last=true, length=%d", last, paddingSize)
	}
}

// TestWriteFrameRegionUnchanged covers C12 on a fixture carrying real
// frame bytes rather than the 42-byte template alone: the frame
// region's hash is equal before and after a write.
func TestWriteFrameRegionUnchanged(t *testing.T) {
	frame := bytes.Repeat([]byte{0xFF, 0xAB, 0x12, 0x34, 0x56}, 500)
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(48000, 6, 16, 0)},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, frame))

	before := frameHash(t, path)
	if err := Write(path, TagSet{Title: new("A")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}
	after := frameHash(t, path)
	if before != after {
		t.Errorf("frame region hash changed: before %x, after %x", before, after)
	}

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(data[c.FrameOffset:c.FrameOffset+c.FrameLength], frame) {
		t.Errorf("frame region bytes differ from the source frame")
	}
}

// TestWriteModeIsCallerMode asserts the renamed file carries the mode
// the caller passed, with a mode that is not 0600 (os.CreateTemp's
// default) and not the fixture's own default (0644).
func TestWriteModeIsCallerMode(t *testing.T) {
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 0)},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	const mode = 0o640
	if err := Write(path, TagSet{Title: new("A")}, mode); err != nil {
		t.Fatalf("Write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != mode {
		t.Errorf("mode = %o, want %o", info.Mode().Perm(), mode)
	}
}

// TestWriteVendorRewritePreservesLoaded covers R9's rewrite half: a
// container that already carries a VORBIS_COMMENT block keeps its
// loaded vendor exactly.
func TestWriteVendorRewritePreservesLoaded(t *testing.T) {
	comment := vorbisCommentPayload("Some Existing Tagger 1.0", []string{"TITLE=old"})
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 1, 16, 0)},
		{Type: blockVorbisComment, Payload: comment},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	if err := Write(path, TagSet{Title: new("new")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}
	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if c.VorbisComment.Vendor != "Some Existing Tagger 1.0" {
		t.Errorf("vendor = %q, want the preserved loaded vendor", c.VorbisComment.Vendor)
	}
}

// TestWriteVendorFreshOnCreatePrefixed covers R9's create half: a
// write into a container carrying no VORBIS_COMMENT block emits a
// fresh vendor carrying the dubplate prefix. The assertion is the
// prefix rather than the exact string, because a test binary and a
// release build report different versions (R9).
func TestWriteVendorFreshOnCreatePrefixed(t *testing.T) {
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 24, 0)},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	if err := Write(path, TagSet{Title: new("A")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}
	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !bytes.HasPrefix([]byte(c.VorbisComment.Vendor), []byte(vendorPrefix)) {
		t.Errorf("vendor = %q, want a %q prefix", c.VorbisComment.Vendor, vendorPrefix)
	}
}

// TestWriteNoCommentBlockInsertsAfterStreamInfo covers C9a: a write
// into a container with no VORBIS_COMMENT block places the new block
// immediately after STREAMINFO, on a buildContainer fixture carrying
// STREAMINFO only (R3 keeps every Go test inside this package; the
// golden already has a block).
func TestWriteNoCommentBlockInsertsAfterStreamInfo(t *testing.T) {
	blocks := []testBlock{
		{Type: blockStreamInfo, Last: true, Payload: streamInfoPayload(44100, 2, 16, 0)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	if err := Write(path, TagSet{Title: new("A")}, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}
	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.Blocks) < 2 || c.Blocks[0].Type != blockStreamInfo || c.Blocks[1].Type != blockVorbisComment {
		t.Fatalf("block chain = %+v, want STREAMINFO then VORBIS_COMMENT", c.Blocks)
	}
}

// TestWriteRefusalNeverCreatesTempFile proves the container is
// validated before the target is ever opened for writing, on a
// directory Write itself cannot write into: if Write tried
// atomicfile.WriteStream's os.CreateTemp before Parse ran, this fixture
// would fail with a permission error instead of ErrMalformedContainer.
// A refusal that reports the right sentinel here proves the ordering a
// bytes-unchanged assertion alone cannot, because atomicfile.WriteStream
// cleans up its own temp file on any error regardless of when that
// error happens.
func TestWriteRefusalNeverCreatesTempFile(t *testing.T) {
	blocks := wellFormedBlocks()
	blocks[0].Payload = blocks[0].Payload[:30] // STREAMINFO not 34 bytes
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.flac")
	if err := os.WriteFile(path, buildContainer(blocks, nil), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := Write(path, TagSet{Title: new("A")}, 0o644)
	if !errors.Is(err, ErrMalformedContainer) {
		t.Errorf("err = %v, want errors.Is(err, ErrMalformedContainer): a permission error here means the target directory was opened for writing before the container was validated", err)
	}
}

// TestWriteRefusals covers one refusal test per malformed shape,
// reusing TestParseRefusals's own fixture generators (container_test.go),
// each asserting the target file is unchanged on disk: a refusal never
// half-writes a file.
func TestWriteRefusals(t *testing.T) {
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
				blocks[1].Last = true
				return buildContainer(blocks, nil)
			},
			wantErr: ErrMalformedContainer,
		},
		{
			name: "declared length running past the end",
			data: func() []byte {
				blocks := wellFormedBlocks()
				data := buildContainer(blocks, nil)
				headerOffset := len(flacSignature) + 4 + len(blocks[0].Payload)
				data[headerOffset+1] = 0xff
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
			data := tt.data()
			path := writeFixture(t, data)

			err := Write(path, TagSet{Title: new("A")}, 0o644)
			if err == nil {
				t.Fatalf("Write: got nil error, want %v", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Write: err = %v, want errors.Is(err, %v)", err, tt.wantErr)
			}
			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read %s after refusal: %v", path, readErr)
			}
			if !bytes.Equal(got, data) {
				t.Errorf("target file changed on disk after a refused write")
			}
		})
	}
}
