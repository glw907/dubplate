package atomicfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name    string
		initial []byte // pre-existing content at path, nil if none
		data    []byte
		perm    os.FileMode
	}{
		{
			name: "creates new file",
			data: []byte("hello"),
			perm: 0o644,
		},
		{
			name:    "overwrites existing file",
			initial: []byte("old content that is longer than new"),
			data:    []byte("new"),
			perm:    0o644,
		},
		{
			name: "sets requested permissions",
			data: []byte("secret"),
			perm: 0o600,
		},
		{
			name: "writes empty content",
			data: []byte{},
			perm: 0o644,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "target.txt")
			if tt.initial != nil {
				if err := os.WriteFile(path, tt.initial, 0o644); err != nil {
					t.Fatalf("seed initial content: %v", err)
				}
			}

			if err := WriteFile(path, tt.data, tt.perm); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read back %s: %v", path, err)
			}
			if string(got) != string(tt.data) {
				t.Errorf("content = %q, want %q", got, tt.data)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat %s: %v", path, err)
			}
			if info.Mode().Perm() != tt.perm {
				t.Errorf("perm = %o, want %o", info.Mode().Perm(), tt.perm)
			}
		})
	}
}

func TestWriteFileLeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target.txt")

	if err := WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	if len(entries) != 1 {
		t.Errorf("dir has %d entries after write, want 1 (no leftover temp file): %v", len(entries), entries)
	}
}

// TestWriteFileCreatesTempFileInTargetDir pins that WriteFile creates its
// temp file in path's own directory, never in $TMPDIR. It points TMPDIR at
// an unwritable directory and asserts the write still succeeds and leaves
// no temp file behind: if WriteFile ever created the temp file under
// os.TempDir() instead, this would fail with a permission error.
func TestWriteFileCreatesTempFileInTargetDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions, so an unwritable TMPDIR cannot be made to fail a write this way")
	}

	badTMPDIR := t.TempDir()
	dir := t.TempDir()

	if err := os.Chmod(badTMPDIR, 0o500); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(badTMPDIR, 0o700) })
	t.Setenv("TMPDIR", badTMPDIR)

	path := filepath.Join(dir, "target.txt")
	if err := WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("content = %q, want %q", got, "data")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	if len(entries) != 1 {
		t.Errorf("dir has %d entries after write, want 1 (no leftover temp file): %v", len(entries), entries)
	}
}

func TestWriteFileMissingDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent-subdir", "target.txt")

	if err := WriteFile(path, []byte("data"), 0o644); err == nil {
		t.Fatal("WriteFile into a nonexistent directory: got nil error, want an error")
	}
}

func TestWriteFileLeavesTheOldContentWhenTheWriteFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions, so the write cannot be made to fail this way")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := WriteFile(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Removing the target directory's write bit fails the write before
	// the rename and leaves path untouched. That is the crash-safety
	// property the shadow store rests on: a reader sees the old content
	// or the new content, never neither.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	if err := WriteFile(path, []byte("second"), 0o644); err == nil {
		t.Fatal("WriteFile into an unwritable directory returned nil; want an error")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "first" {
		t.Errorf("original content = %q, want %q", got, "first")
	}
}
