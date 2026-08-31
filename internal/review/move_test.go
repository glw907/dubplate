package review

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMove_sameFilesystemRename(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src.flac")
	dst := filepath.Join(root, "dst.flac")
	if err := os.WriteFile(src, []byte("track"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Move(src, dst); err != nil {
		t.Fatalf("Move() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "track" {
		t.Errorf("dst content = %q, want %q", got, "track")
	}
	if _, err := os.Stat(src); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("src still exists at %s after Move", src)
	}
}

func TestMove_missingSourceIsAnError(t *testing.T) {
	root := t.TempDir()
	err := Move(filepath.Join(root, "does-not-exist"), filepath.Join(root, "dst"))
	if err == nil {
		t.Fatal("Move() = nil, want an error for a missing source")
	}
}

// copyAndRemove exercises the EXDEV fallback's actual code path
// directly: forcing a genuine cross-device rename failure inside a
// sandboxed test would need a second real filesystem, which no fixture
// here can guarantee, so the fallback's own contract -- copy
// preserving mode bits, then remove the source -- is proven on the
// unexported function Move calls once it has already detected EXDEV.
func TestCopyAndRemove_preservesModeAndTree(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "Bad Album")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "track1.flac"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "track2.flac"), []byte("two"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("track1.flac", filepath.Join(src, "link.flac")); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(root, "dest", "Bad Album")
	if err := copyAndRemove(src, dst); err != nil {
		t.Fatalf("copyAndRemove() error = %v", err)
	}

	if _, err := os.Stat(src); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("source still exists at %s after copyAndRemove", src)
	}

	topInfo, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if perm := topInfo.Mode().Perm(); perm != 0o750 {
		t.Errorf("dst dir mode = %o, want %o", perm, 0o750)
	}

	fileInfo, err := os.Stat(filepath.Join(dst, "track1.flac"))
	if err != nil {
		t.Fatalf("stat copied file: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0o644 {
		t.Errorf("copied file mode = %o, want %o", perm, 0o644)
	}
	if got, err := os.ReadFile(filepath.Join(dst, "track1.flac")); err != nil || string(got) != "one" {
		t.Errorf("copied file content = %q, %v, want %q, nil", got, err, "one")
	}

	nestedInfo, err := os.Stat(filepath.Join(dst, "sub", "track2.flac"))
	if err != nil {
		t.Fatalf("stat nested copied file: %v", err)
	}
	if perm := nestedInfo.Mode().Perm(); perm != 0o600 {
		t.Errorf("nested copied file mode = %o, want %o", perm, 0o600)
	}

	linkTarget, err := os.Readlink(filepath.Join(dst, "link.flac"))
	if err != nil {
		t.Fatalf("readlink copied symlink: %v", err)
	}
	if linkTarget != "track1.flac" {
		t.Errorf("copied symlink target = %q, want %q", linkTarget, "track1.flac")
	}
}

// TestCopyAndRemove_preservesSetgidAndGroupWrite proves the part of the
// "preserving the 2000:2775 setgid tree semantics" criterion the
// 0o750/0o644/0o600 fixture above cannot: every one of those bits
// survives os.MkdirAll/os.OpenFile under the standard umask 022
// unchanged, so a regression that drops the setgid bit or masks
// group-write would still pass that test. This fixture's source top
// directory carries setgid plus 0o775, and its file 0o664 -- both
// altered by umask 022 if copyDir/copyFile rely on MkdirAll/OpenFile's
// mode argument alone.
func TestCopyAndRemove_preservesSetgidAndGroupWrite(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "Setgid Album")
	if err := os.MkdirAll(src, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(src, os.ModeSetgid|0o775); err != nil {
		t.Fatal(err)
	}
	trackPath := filepath.Join(src, "track.flac")
	if err := os.WriteFile(trackPath, []byte("x"), 0o664); err != nil {
		t.Fatal(err)
	}
	// WriteFile's create mode is itself subject to the process umask, so
	// without this the source file would never actually be 0o664 and
	// the assertion below would prove nothing.
	if err := os.Chmod(trackPath, 0o664); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(root, "dest", "Setgid Album")
	if err := copyAndRemove(src, dst); err != nil {
		t.Fatalf("copyAndRemove() error = %v", err)
	}

	dirInfo, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if dirInfo.Mode()&os.ModeSetgid == 0 {
		t.Errorf("dst dir mode = %v, want setgid bit set", dirInfo.Mode())
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o775 {
		t.Errorf("dst dir perm = %o, want %o", perm, 0o775)
	}

	fileInfo, err := os.Stat(filepath.Join(dst, "track.flac"))
	if err != nil {
		t.Fatalf("stat copied file: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0o664 {
		t.Errorf("copied file perm = %o, want %o", perm, 0o664)
	}
}

func TestCopyAndRemove_leavesSourceOnCopyFailure(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src.flac")
	if err := os.WriteFile(src, []byte("track"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A plain file blocking the destination's parent directory makes
	// the copy step fail deterministically (ENOTDIR) without depending
	// on the test runner's privilege level.
	blocked := filepath.Join(root, "blocked")
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(blocked, "dst.flac")

	err := copyAndRemove(src, dst)
	if err == nil {
		t.Fatal("copyAndRemove() = nil, want an error when the copy step fails")
	}
	if _, err := os.Stat(src); err != nil {
		t.Errorf("source removed despite a failed copy: %v", err)
	}
}
