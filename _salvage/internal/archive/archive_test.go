package archive_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/glw907/dubplate/internal/archive"
)

func TestExtract_NativeFormats(t *testing.T) {
	entries := []fixtureEntry{
		{name: "album", dir: true},
		{name: "album/track.txt", content: []byte("payload")},
		{name: "cover.txt", content: []byte("art")},
	}
	tests := []struct {
		name  string
		build func(t *testing.T) string
	}{
		{"zip", func(t *testing.T) string { return buildZip(t, entries) }},
		{"tar", func(t *testing.T) string { return buildTarArchive(t, ".tar", entries) }},
		{"tar.gz", func(t *testing.T) string { return buildTarArchive(t, ".tar.gz", entries) }},
		{"tgz", func(t *testing.T) string { return buildTarArchive(t, ".tgz", entries) }},
		{"tar.bz2", func(t *testing.T) string { return buildTarArchive(t, ".tar.bz2", entries) }},
		{"tbz2", func(t *testing.T) string { return buildTarArchive(t, ".tbz2", entries) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := tt.build(t)
			destDir := t.TempDir()

			if err := archive.Extract(context.Background(), archivePath, destDir, archive.Limits{MaxExtractedBytes: 1 << 20}); err != nil {
				t.Fatalf("Extract: %v", err)
			}

			got, err := os.ReadFile(filepath.Join(destDir, "album", "track.txt"))
			if err != nil {
				t.Fatalf("read nested member: %v", err)
			}
			if string(got) != "payload" {
				t.Errorf("nested member content = %q, want %q", got, "payload")
			}
			got, err = os.ReadFile(filepath.Join(destDir, "cover.txt"))
			if err != nil {
				t.Fatalf("read root member: %v", err)
			}
			if string(got) != "art" {
				t.Errorf("root member content = %q, want %q", got, "art")
			}
		})
	}
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "album.rar")
	if err := os.WriteFile(archivePath, []byte("not really an archive"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	destDir := t.TempDir()

	if err := archive.Extract(context.Background(), archivePath, destDir, archive.Limits{MaxExtractedBytes: 1 << 20}); err == nil {
		t.Fatal("Extract: want error for an unrecognized extension, got nil")
	}
}

func TestExtract_RejectsPathTraversal(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T) string
	}{
		{"zip", func(t *testing.T) string {
			return buildZip(t, []fixtureEntry{{name: "../evil.txt", content: []byte("pwned")}})
		}},
		{"tar", func(t *testing.T) string {
			return buildTarArchive(t, ".tar", []fixtureEntry{{name: "../evil.txt", content: []byte("pwned")}})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := tt.build(t)
			parent := t.TempDir()
			destDir := filepath.Join(parent, "dest")
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				t.Fatalf("mkdir destDir: %v", err)
			}

			err := archive.Extract(context.Background(), archivePath, destDir, archive.Limits{MaxExtractedBytes: 1 << 20})
			if !errors.Is(err, archive.ErrPathTraversal) {
				t.Fatalf("Extract err = %v, want wrapped ErrPathTraversal", err)
			}
			if _, err := os.Stat(filepath.Join(parent, "evil.txt")); !os.IsNotExist(err) {
				t.Errorf("traversal member escaped destDir: stat err = %v, want not-exist", err)
			}
		})
	}
}

func TestExtract_StopsAtMaxExtractedBytes(t *testing.T) {
	content := bytes.Repeat([]byte("x"), 100)
	zipPath := buildZip(t, []fixtureEntry{{name: "track.txt", content: content}})

	t.Run("exactly at the cap succeeds", func(t *testing.T) {
		destDir := t.TempDir()
		lim := archive.Limits{MaxExtractedBytes: int64(len(content))}
		if err := archive.Extract(context.Background(), zipPath, destDir, lim); err != nil {
			t.Fatalf("Extract: %v", err)
		}
		got, err := os.ReadFile(filepath.Join(destDir, "track.txt"))
		if err != nil {
			t.Fatalf("read extracted file: %v", err)
		}
		if !bytes.Equal(got, content) {
			t.Errorf("extracted content = %q, want %q", got, content)
		}
	})

	t.Run("one byte over the cap fails", func(t *testing.T) {
		destDir := t.TempDir()
		lim := archive.Limits{MaxExtractedBytes: int64(len(content)) - 1}
		err := archive.Extract(context.Background(), zipPath, destDir, lim)
		if !errors.Is(err, archive.ErrTooLarge) {
			t.Fatalf("Extract err = %v, want wrapped ErrTooLarge", err)
		}
	})
}

// TestExtract_MaxExtractedBytesSharedAcrossMembers proves the budget covers
// the whole archive rather than resetting per member: two files individually
// under the cap still fail once their combined size exceeds it.
func TestExtract_MaxExtractedBytesSharedAcrossMembers(t *testing.T) {
	a := bytes.Repeat([]byte("a"), 60)
	b := bytes.Repeat([]byte("b"), 60)
	zipPath := buildZip(t, []fixtureEntry{{name: "a.txt", content: a}, {name: "b.txt", content: b}})
	destDir := t.TempDir()

	err := archive.Extract(context.Background(), zipPath, destDir, archive.Limits{MaxExtractedBytes: 100})
	if !errors.Is(err, archive.ErrTooLarge) {
		t.Fatalf("Extract err = %v, want wrapped ErrTooLarge", err)
	}
}

// TestExtract_FreeSpaceCheckedBeforeExtraction proves the free-space
// preflight runs first: archivePath never exists, so if the free-space
// check ran second the failure would instead be an open-archive error, not
// ErrInsufficientSpace, and destDir would still be empty either way.
// Asserting the specific sentinel is what pins the ordering.
func TestExtract_FreeSpaceCheckedBeforeExtraction(t *testing.T) {
	destDir := t.TempDir()
	avail, err := archive.FreeSpace(destDir)
	if err != nil {
		t.Fatalf("FreeSpace: %v", err)
	}
	lim := archive.Limits{MaxExtractedBytes: 1 << 30, MinFreeBytes: avail + 1<<30}
	archivePath := filepath.Join(t.TempDir(), "does-not-exist.zip")

	err = archive.Extract(context.Background(), archivePath, destDir, lim)
	if !errors.Is(err, archive.ErrInsufficientSpace) {
		t.Fatalf("Extract err = %v, want wrapped ErrInsufficientSpace", err)
	}
	entries, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatalf("ReadDir destDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("destDir has %d entries after a space-rejected Extract, want 0", len(entries))
	}
}

func TestExtract_ContextCanceled(t *testing.T) {
	zipPath := buildZip(t, []fixtureEntry{{name: "track.txt", content: []byte("x")}})
	destDir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := archive.Extract(ctx, zipPath, destDir, archive.Limits{MaxExtractedBytes: 1 << 20})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Extract err = %v, want context.Canceled", err)
	}
	entries, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatalf("ReadDir destDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("destDir has %d entries after a canceled Extract, want 0", len(entries))
	}
}

// TestExtract_DoesNotTouchExistingUnrelatedFiles guards the isolated-dest-dir
// contract from the other direction: extraction must never wipe or clobber
// whatever a caller already placed in destDir before calling Extract.
func TestExtract_DoesNotTouchExistingUnrelatedFiles(t *testing.T) {
	destDir := t.TempDir()
	sibling := filepath.Join(destDir, "sibling.txt")
	if err := os.WriteFile(sibling, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("write sibling: %v", err)
	}

	zipPath := buildZip(t, []fixtureEntry{{name: "track.txt", content: []byte("payload")}})
	if err := archive.Extract(context.Background(), zipPath, destDir, archive.Limits{MaxExtractedBytes: 1 << 20}); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	got, err := os.ReadFile(sibling)
	if err != nil {
		t.Fatalf("read sibling: %v", err)
	}
	if string(got) != "keep me" {
		t.Errorf("sibling file mutated by Extract: got %q", got)
	}
}

// TestExtract_RecreatesSymlinkMembers proves a symlink member survives
// extraction as a symlink rather than being followed or silently dropped:
// rejecting it outright is internal/gate's job, run over the extracted tree
// afterward.
func TestExtract_RecreatesSymlinkMembers(t *testing.T) {
	const target = "/nonexistent/target"
	tests := []struct {
		name  string
		build func(t *testing.T) string
	}{
		{"zip", func(t *testing.T) string {
			return buildZip(t, []fixtureEntry{{name: "link.flac", symlink: target}})
		}},
		{"tar", func(t *testing.T) string {
			return buildTarArchive(t, ".tar", []fixtureEntry{{name: "link.flac", symlink: target}})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := tt.build(t)
			destDir := t.TempDir()
			if err := archive.Extract(context.Background(), archivePath, destDir, archive.Limits{MaxExtractedBytes: 1 << 20}); err != nil {
				t.Fatalf("Extract: %v", err)
			}

			linkPath := filepath.Join(destDir, "link.flac")
			info, err := os.Lstat(linkPath)
			if err != nil {
				t.Fatalf("lstat: %v", err)
			}
			if info.Mode()&fs.ModeSymlink == 0 {
				t.Fatalf("extracted member mode = %v, want a symlink", info.Mode())
			}
			got, err := os.Readlink(linkPath)
			if err != nil {
				t.Fatalf("readlink: %v", err)
			}
			if got != target {
				t.Errorf("symlink target = %q, want %q", got, target)
			}
		})
	}
}

func TestFreeSpace(t *testing.T) {
	t.Run("reports positive availability for a real directory", func(t *testing.T) {
		avail, err := archive.FreeSpace(t.TempDir())
		if err != nil {
			t.Fatalf("FreeSpace: %v", err)
		}
		if avail <= 0 {
			t.Errorf("FreeSpace = %d, want > 0", avail)
		}
	})

	t.Run("errors on a nonexistent path", func(t *testing.T) {
		if _, err := archive.FreeSpace(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
			t.Fatal("FreeSpace: want error for a nonexistent path, got nil")
		}
	})
}

// --- fixture archive construction ------------------------------------

// fixtureEntry describes one member of a test archive: a directory (dir),
// a symlink (symlink holds its target), or a regular file (content).
type fixtureEntry struct {
	name    string
	dir     bool
	symlink string
	content []byte
}

func buildZip(t *testing.T, entries []fixtureEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	for _, e := range entries {
		switch {
		case e.dir:
			if _, err := zw.Create(e.name + "/"); err != nil {
				t.Fatalf("zip dir entry %s: %v", e.name, err)
			}
		case e.symlink != "":
			fh := &zip.FileHeader{Name: e.name}
			fh.SetMode(os.ModeSymlink | 0o777)
			w, err := zw.CreateHeader(fh)
			if err != nil {
				t.Fatalf("zip symlink entry %s: %v", e.name, err)
			}
			if _, err := w.Write([]byte(e.symlink)); err != nil {
				t.Fatalf("write zip symlink target: %v", err)
			}
		default:
			w, err := zw.Create(e.name)
			if err != nil {
				t.Fatalf("zip file entry %s: %v", e.name, err)
			}
			if _, err := w.Write(e.content); err != nil {
				t.Fatalf("write zip content: %v", err)
			}
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
	return path
}

func buildTarBytes(t *testing.T, entries []fixtureEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		switch {
		case e.dir:
			if err := tw.WriteHeader(&tar.Header{Name: e.name + "/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
				t.Fatalf("tar dir header %s: %v", e.name, err)
			}
		case e.symlink != "":
			if err := tw.WriteHeader(&tar.Header{Name: e.name, Typeflag: tar.TypeSymlink, Linkname: e.symlink, Mode: 0o777}); err != nil {
				t.Fatalf("tar symlink header %s: %v", e.name, err)
			}
		default:
			if err := tw.WriteHeader(&tar.Header{Name: e.name, Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(e.content))}); err != nil {
				t.Fatalf("tar file header %s: %v", e.name, err)
			}
			if _, err := tw.Write(e.content); err != nil {
				t.Fatalf("write tar content: %v", err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	return buf.Bytes()
}

// buildTarArchive wraps buildTarBytes' raw tar stream to match ext.
// ".tar.bz2"/".tbz2" shell out to the bzip2 CLI since compress/bzip2 is
// read-only in the standard library; ffmpeg, flac, and metaflac are
// likewise assumed present on this workstation for the gate package's
// fixtures, so requiring bzip2 here matches that existing precedent rather
// than adding a new one.
func buildTarArchive(t *testing.T, ext string, entries []fixtureEntry) string {
	t.Helper()
	raw := buildTarBytes(t, entries)
	path := filepath.Join(t.TempDir(), "fixture"+ext)

	switch ext {
	case ".tar":
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			t.Fatalf("write tar: %v", err)
		}
	case ".tar.gz", ".tgz":
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("create %s: %v", path, err)
		}
		gz := gzip.NewWriter(f)
		if _, err := gz.Write(raw); err != nil {
			t.Fatalf("gzip write: %v", err)
		}
		if err := gz.Close(); err != nil {
			t.Fatalf("gzip close: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("close %s: %v", path, err)
		}
	case ".tar.bz2", ".tbz2":
		cmd := exec.Command("bzip2", "-c")
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("bzip2 compress fixture: %v", err)
		}
		if err := os.WriteFile(path, out, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	default:
		t.Fatalf("buildTarArchive: unsupported ext %q", ext)
	}
	return path
}
