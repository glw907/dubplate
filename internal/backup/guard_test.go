package backup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree creates each file under dir, with its parent directories.
func writeTree(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, f := range files {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", f, err)
		}
		if err := os.WriteFile(p, []byte(f), 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}
}

// codeOf returns the failure code err carries, or "" for none.
func codeOf(err error) string {
	if f, ok := errors.AsType[*failure](err); ok {
		return f.code
	}
	return ""
}

// TestResolveRoot pins the library root's rows: a symlink to a
// directory resolves to its target, and an absent root, a dangling
// symlink, and a regular file all read missing.
func TestResolveRoot(t *testing.T) {
	base := t.TempDir()
	lib := filepath.Join(base, "lib")
	writeTree(t, lib, "a/1.flac")
	link := filepath.Join(base, "link")
	if err := os.Symlink(lib, link); err != nil {
		t.Fatal(err)
	}
	dangling := filepath.Join(base, "dangling")
	if err := os.Symlink(filepath.Join(base, "gone"), dangling); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(lib, "a", "1.flac")

	tests := []struct {
		name     string
		root     string
		want     string
		wantCode string
	}{
		{name: "directory", root: lib, want: lib},
		{name: "symlink to a directory", root: link, want: lib},
		{name: "absent", root: filepath.Join(base, "absent"), wantCode: "library-missing"},
		{name: "dangling symlink", root: dangling, wantCode: "library-missing"},
		{name: "regular file", root: file, wantCode: "library-missing"},
		{name: "under a regular file", root: filepath.Join(file, "x"), wantCode: "library-missing"},
		{name: "unset", root: "", wantCode: "library-missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveRoot(tt.root)
			if code := codeOf(err); code != tt.wantCode {
				t.Fatalf("resolveRoot(%q) = %v, want code %q", tt.root, err, tt.wantCode)
			}
			if got != tt.want {
				t.Errorf("resolveRoot(%q) = %q, want %q", tt.root, got, tt.want)
			}
		})
	}
}

// TestCountFiles pins that only regular files count: a directory, a
// symlink to a file, and a symlinked subdirectory's contents do not.
func TestCountFiles(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, "a/1.flac", "a/b/2.flac", "3.flac")
	if err := os.MkdirAll(filepath.Join(root, "empty/deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "3.flac"), filepath.Join(root, "link.flac")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "a"), filepath.Join(root, "linkdir")); err != nil {
		t.Fatal(err)
	}
	n, err := countFiles(root)
	if err != nil {
		t.Fatalf("countFiles: %v", err)
	}
	if n != 3 {
		t.Errorf("countFiles = %d, want 3", n)
	}
}

// TestCountFiles_Unreadable pins the unreadable row: a subdirectory
// that cannot be read fails the count rather than reading short.
func TestCountFiles_Unreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads a mode-0 directory")
	}
	root := t.TempDir()
	writeTree(t, root, "a/1.flac")
	locked := filepath.Join(root, "a")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	if _, err := countFiles(root); !errors.Is(err, os.ErrPermission) {
		t.Errorf("countFiles = %v, want a permission error", err)
	}
}

// TestGuard pins the count guard's rows. The expectations are the
// spec's: half the catalog's tracks as regular files, --max-delete the
// track count with tracks, and 0 with none.
func TestGuard(t *testing.T) {
	tests := []struct {
		name          string
		files, tracks int
		wantMax       int
		wantCode      string
		wantMsg       []string
	}{
		{name: "zero tracks, empty root", files: 0, tracks: 0, wantCode: "library-empty"},
		{name: "zero tracks, populated root", files: 3, tracks: 0, wantMax: 0},
		{name: "below half", files: 1, tracks: 4, wantCode: "library-short", wantMsg: []string{"1 files", "4 tracks"}},
		{name: "no files, tracks", files: 0, tracks: 1, wantCode: "library-short"},
		{name: "exactly half", files: 2, tracks: 4, wantMax: 4},
		{name: "odd count, just under half", files: 2, tracks: 5, wantCode: "library-short"},
		{name: "odd count, just over half", files: 3, tracks: 5, wantMax: 5},
		{name: "above", files: 10, tracks: 4, wantMax: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := guard("/lib", tt.files, tt.tracks)
			if code := codeOf(err); code != tt.wantCode {
				t.Fatalf("guard(%d, %d) = %v, want code %q", tt.files, tt.tracks, err, tt.wantCode)
			}
			for _, m := range tt.wantMsg {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("guard message %q does not name %q", err, m)
				}
			}
			if err == nil && got != tt.wantMax {
				t.Errorf("guard(%d, %d) max-delete = %d, want %d", tt.files, tt.tracks, got, tt.wantMax)
			}
		})
	}
}
