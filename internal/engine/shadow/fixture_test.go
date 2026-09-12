package shadow

import (
	"os"
	"testing"

	"github.com/glw907/dubplate/internal/engine/tagio"
)

// committedTemplateFLACOneSecond is
// internal/harness/oracle/testdata/template.flac's 42 bytes with its
// STREAMINFO pack's total-samples field spliced from 0 to 44100 (one
// second at the template's own 44100 Hz sample rate): the pack's last
// two bytes change from 0x00, 0x00 to 0x00, 0xac, 0x44's low bytes,
// everything else byte-for-byte identical to the template mutagen's
// own StreamInfo.write produced (mono, 44100 Hz, 16-bit). It is a byte
// edit of the committed template rather than a container built fresh
// (Task 1, R3), and it is the seed every writer-built fixture in this
// package starts from, so a content join comparing two fixtures'
// Length never compares zero to zero vacuously.
var committedTemplateFLACOneSecond = []byte{
	'f', 'L', 'a', 'C',
	0x80, 0x00, 0x00, 0x22, // the last block, type 0, 34 bytes of payload
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x0a, 0xc4, 0x40, 0xf0, 0x00, 0x00, 0xac, 0x44,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
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
