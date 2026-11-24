package fsys

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
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

// TestDevice holds Device to the device os.Stat reports, and holds the
// /proc mount, a separate file system, apart from the root's.
func TestDevice(t *testing.T) {
	dir := t.TempDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("os.Stat(%s).Sys() = %T, want *syscall.Stat_t", dir, info.Sys())
	}
	want := st.Dev
	got, err := Device(dir)
	if err != nil {
		t.Fatalf("Device: %v", err)
	}
	if got != want {
		t.Errorf("Device(%s) = %d, want %d", dir, got, want)
	}

	proc, err := Device("/proc")
	if err != nil {
		t.Fatalf("Device(/proc): %v", err)
	}
	root, err := Device("/")
	if err != nil {
		t.Fatalf("Device(/): %v", err)
	}
	if proc == root {
		t.Errorf("Device(/proc) = Device(/) = %d, want the proc mount apart", proc)
	}

	if _, err := Device(filepath.Join(dir, "absent")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Device of an absent path = %v, want it to wrap fs.ErrNotExist", err)
	}
}
