package tagio

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// FuzzParseBounded feeds arbitrary bytes to the bounded container walk,
// to Read, and to Sniff. Each may refuse and none may panic. A refusal
// from ParseBounded carries the zero Container, and a container it
// accepts places its frame region inside the file.
func FuzzParseBounded(f *testing.F) {
	golden, err := os.ReadFile(filepath.Join("testdata", "golden.flac"))
	if err != nil {
		f.Fatal(err)
	}
	f.Add(golden)
	f.Add(golden[:64])
	f.Add([]byte("fLaC"))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		p := filepath.Join(t.TempDir(), "in.flac")
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
		c, err := ParseBounded(p, Bounds{TagBytes: 1 << 16, PictureBytes: 1 << 18})
		if err != nil && !reflect.ValueOf(c).IsZero() {
			t.Fatalf("refusal returned a non-zero Container: %v", err)
		}
		if err == nil && (c.FrameOffset < 0 || c.FrameLength < 0 || c.FrameOffset+c.FrameLength > int64(len(data))) {
			t.Fatalf("frame region %d+%d lies outside the file's %d bytes", c.FrameOffset, c.FrameLength, len(data))
		}
		_, _ = Read(p)
		_, _ = Sniff(p)
	})
}
