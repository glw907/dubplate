//go:build !linux

package fsys

import (
	"errors"
	"fmt"
)

// FreeBytes refuses with [errors.ErrUnsupported].
func FreeBytes(dir string) (uint64, error) {
	return 0, fmt.Errorf("statfs %s: %w", dir, errors.ErrUnsupported)
}

// MountPoint refuses with [errors.ErrUnsupported].
func MountPoint(dir string) (bool, error) {
	return false, fmt.Errorf("stat %s: %w", dir, errors.ErrUnsupported)
}
