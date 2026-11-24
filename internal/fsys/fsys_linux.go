package fsys

import (
	"fmt"

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

// Device returns st_dev of path, following a symlink. Two paths on one
// file system share a device, so a directory whose device differs from
// its parent's is a mount point.
func Device(path string) (uint64, error) {
	var st unix.Stat_t
	if err := unix.Stat(path, &st); err != nil {
		return 0, fmt.Errorf("stat %s: %w", path, err)
	}
	return st.Dev, nil
}
