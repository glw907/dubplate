package fsys

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestFreeBytes(t *testing.T) {
	got, err := FreeBytes(t.TempDir())
	if err != nil {
		t.Fatalf("FreeBytes: %v", err)
	}
	if got == 0 {
		t.Error("FreeBytes = 0 on a temporary directory")
	}
	if _, err := FreeBytes(filepath.Join(t.TempDir(), "absent")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("FreeBytes of an absent path = %v, want it to wrap fs.ErrNotExist", err)
	}
}

// TestMountPoint holds MountPoint to the two answers: /proc, a separate
// file system, is a mount point, and a directory made inside a
// temporary directory is not. An absent path refuses.
func TestMountPoint(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "child")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		path string
		want bool
	}{
		{"/proc", true},
		{dir, false},
	} {
		got, err := MountPoint(tt.path)
		if err != nil {
			t.Fatalf("MountPoint(%s): %v", tt.path, err)
		}
		if got != tt.want {
			t.Errorf("MountPoint(%s) = %v, want %v", tt.path, got, tt.want)
		}
	}

	if _, err := MountPoint(filepath.Join(dir, "absent")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("MountPoint of an absent path = %v, want it to wrap fs.ErrNotExist", err)
	}
}
