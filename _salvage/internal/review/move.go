package review

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
)

// musicUID and musicGID are the "music" service user Move's
// copy-and-remove fallback best-effort chowns a copied item to,
// matching the ownership os.Rename would have preserved for free
// within one filesystem (cloud-init/user-data.yaml creates the music
// user at this uid/gid, owning the whole /srv/music tree).
const (
	musicUID = 2000
	musicGID = 2000
)

// Move moves src to dst. It tries the atomic os.Rename first; when
// that fails with EXDEV (src and dst are on different filesystems --
// os.Rename cannot cross that boundary, unlike the shell layer's mv),
// Move falls back to a recursive copy followed by removing src,
// preserving every entry's mode bits.
func Move(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	var linkErr *os.LinkError
	if !errors.As(err, &linkErr) || !errors.Is(linkErr.Err, syscall.EXDEV) {
		return fmt.Errorf("move %s to %s: %w", src, dst, err)
	}
	return copyAndRemove(src, dst)
}

// copyAndRemove recursively copies src to dst, then removes src, so a
// cross-device Move still ends with src gone, exactly as os.Rename
// would have left it. dst is removed on a failed copy so a partial
// tree never lingers where a caller might mistake it for a complete
// one.
func copyAndRemove(src, dst string) error {
	if err := copyTree(src, dst); err != nil {
		_ = os.RemoveAll(dst)
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}
	if err := os.RemoveAll(src); err != nil {
		return fmt.Errorf("remove %s after copying to %s: %w", src, dst, err)
	}
	return nil
}

// copyTree copies src to dst, recursing into directories and
// recreating symlinks rather than following them -- a symlink can
// reach dst by this path only because [internal/gate] already
// rejected the album containing it, and quarantine's job is to
// preserve exactly what was rejected, not resolve it.
func copyTree(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	switch {
	case info.Mode()&fs.ModeSymlink != 0:
		return copySymlink(src, dst)
	case info.IsDir():
		return copyDir(src, dst, info)
	default:
		return copyFile(src, dst, info)
	}
}

func copyDir(src, dst string, info os.FileInfo) error {
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return fmt.Errorf("create directory %s: %w", dst, err)
	}
	// MkdirAll's mode is masked by both Perm() (which drops
	// setgid/setuid/sticky) and the process umask, so the setgid bit
	// that keeps a review/quarantine tree group-owned is lost unless
	// restored explicitly. os.Chmod is not umask-masked.
	if err := os.Chmod(dst, info.Mode()&(fs.ModePerm|fs.ModeSetgid|fs.ModeSetuid|fs.ModeSticky)); err != nil {
		return fmt.Errorf("set mode on directory %s: %w", dst, err)
	}
	chown(dst)

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", src, err)
	}
	for _, e := range entries {
		if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, info os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("write %s: %w", dst, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dst, err)
	}
	// OpenFile's create mode is reduced by the process umask, so a
	// group-writable source file can land dst as non-group-writable
	// unless the mode is restored explicitly here.
	if err := os.Chmod(dst, info.Mode().Perm()); err != nil {
		return fmt.Errorf("set mode on %s: %w", dst, err)
	}
	chown(dst)
	return nil
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("read symlink %s: %w", src, err)
	}
	if err := os.Symlink(target, dst); err != nil {
		return fmt.Errorf("create symlink %s: %w", dst, err)
	}
	return nil
}

// chown best-effort sets path's owner to musicUID:musicGID. A failure
// here never fails the move: it only means the copied item keeps the
// current process's owner instead of music's, which a later
// administrative chown corrects.
func chown(path string) {
	if err := os.Chown(path, musicUID, musicGID); err != nil {
		slog.Debug("review: chown after copy", "path", path, "err", err)
	}
}
