package shadow

import (
	"os"
	"testing"

	"github.com/glw907/dubplate/internal/engine/tagio"
)

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

// committedTemplateFLACOneSecond is committedTemplateFLAC with its
// STREAMINFO pack's total-samples field spliced from 0 to 44100 (one
// second at the template's own 44100 Hz sample rate): the pack's last
// two bytes change from 0x00, 0x00 to 0x00, 0xac, 0x44's low bytes,
// everything else byte-for-byte identical. It is a byte edit of the
// committed literal rather than a container built fresh (Task 1, R3),
// and it is the seed every writer-built fixture in this package starts
// from, so a content join comparing two fixtures' Length never compares
// zero to zero vacuously.
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
// leave-alone and strip states produce identical bytes) and mode, then
// pins the modification time to fixedModTime only after the write,
// because tagio.Write's atomic rename replaces the file and a mtime
// pinned before the call would not survive it.
func writeFixtureTrack(t *testing.T, path string, ts tagio.TagSet, mode os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, committedTemplateFLACOneSecond, 0o644); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	if err := tagio.Write(path, ts, nil, mode); err != nil {
		t.Fatalf("tagio.Write: %v", err)
	}
	if err := os.Chtimes(path, fixedModTime, fixedModTime); err != nil {
		t.Fatalf("chtimes fixture: %v", err)
	}
}
