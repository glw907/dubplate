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

func TestWriteFileMissingDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent-subdir", "target.txt")

	if err := WriteFile(path, []byte("data"), 0o644); err == nil {
		t.Fatal("WriteFile into a nonexistent directory: got nil error, want an error")
	}
}
