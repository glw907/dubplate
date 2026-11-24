package fsys

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// FreeBytes returns the bytes available to an unprivileged user on the
// file system holding dir: f_bavail counts f_frsize units.
func FreeBytes(dir string) (uint64, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(dir, &st); err != nil {
		return 0, fmt.Errorf("statfs %s: %w", dir, err)
	}
	return st.Bavail * uint64(st.Frsize), nil
}

// MountPoint reports whether dir is a mount point: whether its st_dev,
// following a symlink, differs from its parent directory's. Two paths on
// one file system share a device.
func MountPoint(dir string) (bool, error) {
	dev, err := device(dir)
	if err != nil {
		return false, err
	}
	parent, err := device(filepath.Dir(dir))
	if err != nil {
		return false, err
	}
	return dev != parent, nil
}

func device(path string) (uint64, error) {
	var st unix.Stat_t
	if err := unix.Stat(path, &st); err != nil {
		return 0, fmt.Errorf("stat %s: %w", path, err)
	}
	return st.Dev, nil
}
