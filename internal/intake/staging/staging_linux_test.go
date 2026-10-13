package staging

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestMntIDMaskIsStatxMntID(t *testing.T) {
	if mntIDMask != unix.STATX_MNT_ID {
		t.Fatalf("mntIDMask = %#x, want STATX_MNT_ID %#x", mntIDMask, unix.STATX_MNT_ID)
	}
}

func TestReadSiblingDirectories(t *testing.T) {
	root := t.TempDir()
	stagingPath := filepath.Join(root, "staging")
	libraryPath := filepath.Join(root, "library")
	for _, p := range []string{stagingPath, libraryPath} {
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	staging, library := Read(stagingPath), Read(libraryPath)

	if staging.Err != nil || library.Err != nil || staging.StatxErr != nil || library.StatxErr != nil {
		t.Fatalf("Read errors: %v, %v, %v, %v", staging.Err, library.Err, staging.StatxErr, library.StatxErr)
	}
	if staging.Mask&unix.STATX_MNT_ID == 0 || staging.MountID == 0 || staging.Dev == 0 {
		t.Errorf("Read left the mount unset: %+v", staging)
	}
	if err := Check(staging, library); err != nil {
		t.Errorf("Check over sibling directories = %v, want nil", err)
	}
}

func TestReadResolvesSymlinksForNesting(t *testing.T) {
	root := t.TempDir()
	stagingPath := filepath.Join(root, "staging")
	if err := os.MkdirAll(filepath.Join(stagingPath, "inner"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "library")
	if err := os.Symlink(filepath.Join(stagingPath, "inner"), link); err != nil {
		t.Fatal(err)
	}

	if err := Check(Read(stagingPath), Read(link)); err == nil {
		t.Fatal("Check = nil for a library root that is a symlink into staging")
	}
}

func TestReadFailures(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Read(filepath.Join(root, "absent")).Err; !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("absent path: Err = %v, want it to wrap fs.ErrNotExist", err)
	}
	if err := Read(file).Err; err == nil {
		t.Error("a file: Err = nil, want a refusal")
	}
}

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
