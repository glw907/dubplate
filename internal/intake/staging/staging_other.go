//go:build !linux

package staging

// Read refuses with [ErrLinuxOnly], carried in the result as on Linux.
func Read(path string) Dir {
	return Dir{Path: path, Err: ErrLinuxOnly}
}

// FreeBytes refuses with [ErrLinuxOnly].
func FreeBytes(dir string) (uint64, error) {
	return 0, ErrLinuxOnly
}
