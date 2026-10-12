package staging

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Read stats path and reports its device and mount. A failure is
// carried in the result for [Check] to word, not returned.
func Read(path string) Dir {
	d := Dir{Path: path}

	var st unix.Stat_t
	if err := unix.Stat(path, &st); err != nil {
		d.Err = fmt.Errorf("stat %s: %w", path, err)
		return d
	}
	if st.Mode&unix.S_IFMT != unix.S_IFDIR {
		d.Err = fmt.Errorf("%s is not a directory", path)
		return d
	}
	if err := unix.Access(path, unix.W_OK); err != nil {
		d.Err = fmt.Errorf("write access to %s: %w", path, err)
		return d
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		d.Err = fmt.Errorf("resolve %s: %w", path, err)
		return d
	}
	d.Resolved = resolved
	d.Dev = uint64(st.Dev)

	var stx unix.Statx_t
	if err := unix.Statx(unix.AT_FDCWD, path, 0, unix.STATX_MNT_ID, &stx); err != nil {
		d.StatxErr = fmt.Errorf("statx %s: %w", path, err)
		return d
	}
	d.Mask = stx.Mask
	d.MountID = stx.Mnt_id
	return d
}

// FreeBytes returns the bytes available to an unprivileged user on the
// filesystem holding dir: f_bavail counts f_frsize units.
func FreeBytes(dir string) (uint64, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(dir, &st); err != nil {
		return 0, fmt.Errorf("statfs %s: %w", dir, err)
	}
	return st.Bavail * uint64(st.Frsize), nil
}
