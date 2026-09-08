// Package atomicfile writes files atomically, so a crash or a concurrent
// reader never observes a partial write.
package atomicfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteFile writes data to path atomically: it creates a temp file in
// path's directory, writes and syncs data to it, then renames it over
// path. The rename is atomic on the same filesystem, so readers see
// either the old content or the new content, never a partial file.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmpPath, err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("chmod %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tmpPath, path, err)
	}
	return nil
}

// WriteStream writes path atomically the same way WriteFile does, except
// the content comes from write, which streams into an io.Writer rather
// than arriving as an in-memory []byte. Content never buffers in full, so
// a multi-gigabyte rewrite costs one file's worth of disk and a fixed
// amount of memory. perm is a parameter rather than inherited from
// os.CreateTemp's 0600 default, because a caller rewriting a library file
// in place needs the rename to land the file's real mode, not a
// temp-file mode that would leave it unreadable. After the rename,
// WriteStream opens path's containing directory and fsyncs it, so the
// directory entry survives a power cut along with the file's new
// content.
func WriteStream(path string, perm os.FileMode, write func(io.Writer) error) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()

	if err := write(tmp); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("sync %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close %s: %w", tmpPath, err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("chmod %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %s to %s: %w", tmpPath, path, err)
	}

	dirFile, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open %s: %w", dir, err)
	}
	defer func() { _ = dirFile.Close() }()
	if err := dirFile.Sync(); err != nil {
		return fmt.Errorf("fsync %s: %w", dir, err)
	}
	return nil
}
