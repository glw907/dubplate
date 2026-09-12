package shadow

import (
	"bytes"
	"os"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/engine/tagio"
)

// committedTemplateFLAC is internal/harness/oracle/testdata/template.flac
// byte for byte: the 42 bytes mutagen's own StreamInfo.write produced
// (mono, 44100 Hz, 16-bit, zero total samples), a STREAMINFO block and
// nothing else. It is spelled here rather than read back from that
// path so this package's fixtures carry no relative dependency on
// another package's testdata directory.
var committedTemplateFLAC = []byte{
	'f', 'L', 'a', 'C',
	0x80, 0x00, 0x00, 0x22, // the last block, type 0, 34 bytes of payload
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x0a, 0xc4, 0x40, 0xf0, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// committedTemplateFLACOneSecond is committedTemplateFLAC with its
// STREAMINFO total-samples field spliced from 0 to 44100, one second at
// the template's own 44100 Hz sample rate. It is a byte edit of the
// committed template rather than a container built fresh (Task 1, R3),
// and it is the seed every writer-built fixture in this package starts
// from, so a content join comparing two fixtures' Length never compares
// zero to zero vacuously.
var committedTemplateFLACOneSecond = oneSecondSeed()

func oneSecondSeed() []byte {
	seed := slices.Clone(committedTemplateFLAC)
	// The 36-bit total-samples field ends where the 16-byte MD5
	// signature that closes the STREAMINFO payload begins, and 44100
	// fits in the field's low two bytes.
	md5 := len(seed) - 16
	seed[md5-2], seed[md5-1] = 0xac, 0x44
	return seed
}

// writeFixtureTrack writes ts to path through the repo's own writer
// rather than a second, test-only FLAC encoder (ruling 4): it seeds
// path with committedTemplateFLACOneSecond, calls tagio.Write with a
// nil picture set (the seed carries no PICTURE block, so R1's
// leave-alone and strip states produce identical bytes) at the
// package's one fixture mode, then pins the modification time to
// fixedModTime only after the write, because tagio.Write's atomic
// rename replaces the file and a mtime pinned before the call would
// not survive it.
func writeFixtureTrack(t *testing.T, path string, ts tagio.TagSet) {
	t.Helper()

	const fixtureMode = 0o644
	if err := os.WriteFile(path, committedTemplateFLACOneSecond, fixtureMode); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	if err := tagio.Write(path, ts, nil, fixtureMode); err != nil {
		t.Fatalf("tagio.Write: %v", err)
	}
	if err := os.Chtimes(path, fixedModTime, fixedModTime); err != nil {
		t.Fatalf("chtimes fixture: %v", err)
	}
}

// spliceDuplicateVorbisComment reads path (a track writeFixtureTrack
// already wrote) and returns its bytes with a second, byte-for-byte
// copy of the VORBIS_COMMENT block inserted immediately after the
// first: ruling 9's "more than one VORBIS_COMMENT block" refusal
// (spec:310-317), a shape tagio.Write never produces on its own and
// this package's only way to put one in front of tagio.Parse (Task 3,
// R3). It emits only containers the writer refuses to produce.
//
// It borrows no block-type constant, no header encoder and no
// binary-encoding standard-library import (A1): the VORBIS_COMMENT
// block is the one whose payload opens with the vendor string
// tagio.Parse already decoded onto Container.VorbisComment.Vendor
// (the payload's first field is that string behind a length prefix,
// write.go:217-231), and its byte range is four (the block header)
// plus the running sum of each preceding block's own four-plus-Length
// span, read off Container.Blocks. The duplicate is that range copied
// whole, header included, never a block this package builds itself.
func spliceDuplicateVorbisComment(t *testing.T, path string) []byte {
	t.Helper()

	c, err := tagio.Parse(path)
	if err != nil {
		t.Fatalf("parse fixture to splice: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture to splice: %v", err)
	}

	vendor := []byte(c.VorbisComment.Vendor)
	offset := len("fLaC")
	start, length := -1, 0
	for _, b := range c.Blocks {
		if len(b.Payload) >= 4+len(vendor) && bytes.Equal(b.Payload[4:4+len(vendor)], vendor) {
			start = offset
			length = 4 + b.Length
			break
		}
		offset += 4 + b.Length
	}
	if start < 0 {
		t.Fatalf("splice: no VORBIS_COMMENT block found in %s", path)
	}

	block := raw[start : start+length]
	out := make([]byte, 0, len(raw)+length)
	out = append(out, raw[:start+length]...)
	out = append(out, block...)
	out = append(out, raw[start+length:]...)
	return out
}

// spliceInvalidUTF8TitleByte reads path (a track writeFixtureTrack
// already wrote with Title "before_after") and returns its bytes with
// that ASCII underscore replaced by the single byte 0x80, an
// ill-formed UTF-8 continuation byte with no leading byte before it
// (R6): tagio.Write only ever accepts values already valid UTF-8
// (ruling 10, rules_test.go's TestRulesInvalidUTF8ReadBack), so this
// package's only way to put an invalid sequence in front of
// tagio.Read is a length-preserving byte splice of a writer-built
// fixture, the same shape spliceDuplicateVorbisComment uses above. No
// length field moves: one byte is replaced by one byte.
func spliceInvalidUTF8TitleByte(t *testing.T, path string) []byte {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture to splice: %v", err)
	}

	marker := []byte("before_after")
	idx := bytes.Index(raw, marker)
	if idx < 0 {
		t.Fatalf("splice: title marker not found in %s", path)
	}
	underscore := idx + len("before")

	out := make([]byte, len(raw))
	copy(out, raw)
	out[underscore] = 0x80
	return out
}
