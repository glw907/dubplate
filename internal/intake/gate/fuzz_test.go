package gate

import (
	"path/filepath"
	"testing"
)

// FuzzValidPath feeds arbitrary bytes to ValidPath. It must never panic,
// and a path it accepts must be local by the standard library's own
// reading, so no accepted member name can climb out of its directory.
func FuzzValidPath(f *testing.F) {
	for _, s := range []string{"a/b.flac", "../x", "/abs", "", "a\x00b", "CON", "a//b"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, p []byte) {
		if ValidPath(p) && !filepath.IsLocal(string(p)) {
			t.Fatalf("ValidPath accepted %q, which filepath.IsLocal refuses", p)
		}
	})
}

// FuzzWalkZIP stages arbitrary bytes as an archive and walks it as an
// album and as separate tracks. A walk may refuse; it must never panic,
// and a walk that succeeds keeps only members with an accepted, local
// path, whose declared sizes sum to Total within the policy's cap.
func FuzzWalkZIP(f *testing.F) {
	f.Add(newZIPFixture(f).file("01 Track.flac", "fLaC").bytes())
	f.Add(newZIPFixture(f).file("../01 Track.flac", "fLaC").bytes())
	f.Add(newZIPFixture(f).bytes())
	p := zipPolicy()
	f.Fuzz(func(t *testing.T, data []byte) {
		for _, tracks := range []bool{false, true} {
			_, w, err := walkStaged(t, data, p, tracks)
			if err != nil {
				continue
			}
			var sum int64
			for _, m := range w.Members {
				if !ValidPath(m.Path) || !filepath.IsLocal(string(m.Path)) {
					t.Errorf("separateTracks=%v: kept member %q, a path ValidPath or filepath.IsLocal refuses", tracks, m.Path)
				}
				sum += m.Size
			}
			if w.Total != sum {
				t.Errorf("separateTracks=%v: Total = %d, want the members' sum %d", tracks, w.Total, sum)
			}
			if w.Total > p.MaxAlbumBytes {
				t.Errorf("separateTracks=%v: Total = %d, over the policy's cap %d", tracks, w.Total, p.MaxAlbumBytes)
			}
		}
	})
}
