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

// Device refuses with [errors.ErrUnsupported].
func Device(path string) (uint64, error) {
	return 0, fmt.Errorf("stat %s: %w", path, errors.ErrUnsupported)
}
