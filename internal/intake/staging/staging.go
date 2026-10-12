// Package staging holds the two operating-system-specific reads that
// intake makes of its staging directory: which device and mount it sits
// on, and how much space is free there. The reads are built for Linux;
// elsewhere both refuse, so a server started on another platform stops
// at its startup check.
package staging

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
)

// ErrLinuxOnly is the refusal the reads give on a platform other than
// Linux.
var ErrLinuxOnly = errors.New("staging checks are built for Linux only")

// mntIDMask is the statx(2) request and result bit for the mount id,
// STATX_MNT_ID; staging_linux_test.go holds it equal to the x/sys
// constant. It is declared here so the check builds on every platform.
const mntIDMask = 0x1000

// Dir is what [Read] learned of one directory: its configured and
// symlink-resolved paths, its device, and its mount.
type Dir struct {
	// Path is the directory as configured.
	Path string
	// Resolved is Path after symlink resolution.
	Resolved string
	// Dev is st_dev.
	Dev uint64
	// Mask is statx's stx_mask: the fields the kernel returned.
	Mask uint32
	// MountID is the mount id statx returned, meaningful only when Mask
	// carries STATX_MNT_ID.
	MountID uint64
	// Err is the first failure that made the directory unusable: it
	// wraps [io/fs.ErrNotExist] when the directory is absent, and
	// otherwise names the failing operation. The fields above it are
	// unset when Err is set.
	Err error
	// StatxErr is the failure of the statx call itself.
	StatxErr error
}

// Check returns the reason the staging directory and the library root
// cannot share one rename, or nil when they can. It tests, in order,
// each directory's own failure, nesting of the resolved paths, the
// device, the statx call and its mask, and the mount id. A check that
// could not run is a refusal, never a pass.
func Check(staging, library Dir) error {
	if err := dirErr("staging directory", staging); err != nil {
		return err
	}
	if err := dirErr("library root", library); err != nil {
		return err
	}
	if nested(staging.Resolved, library.Resolved) {
		return fmt.Errorf("staging directory %s and library root %s nest (resolved to %s and %s)",
			staging.Path, library.Path, staging.Resolved, library.Resolved)
	}
	if staging.Dev != library.Dev {
		return fmt.Errorf("staging directory %s (device %d) and library root %s (device %d) sit on different devices",
			staging.Path, staging.Dev, library.Path, library.Dev)
	}
	for _, d := range []struct {
		role string
		dir  Dir
	}{{"staging directory", staging}, {"library root", library}} {
		if d.dir.StatxErr != nil {
			return fmt.Errorf("%s: %w", d.role, d.dir.StatxErr)
		}
		if d.dir.Mask&mntIDMask == 0 {
			return fmt.Errorf("%s: statx %s did not return STATX_MNT_ID (mask %#x)", d.role, d.dir.Path, d.dir.Mask)
		}
	}
	if staging.MountID != library.MountID {
		return fmt.Errorf("staging directory %s (mount %d) and library root %s (mount %d) must sit in one mount",
			staging.Path, staging.MountID, library.Path, library.MountID)
	}
	return nil
}

// dirErr words d's own failure for the directory named by role.
func dirErr(role string, d Dir) error {
	if d.Err == nil {
		return nil
	}
	if errors.Is(d.Err, fs.ErrNotExist) {
		return fmt.Errorf("%s %s does not exist", role, d.Path)
	}
	return fmt.Errorf("%s: %w", role, d.Err)
}

// nested reports whether either resolved path equals or lies inside
// the other.
func nested(a, b string) bool {
	within := func(base, target string) bool {
		rel, err := filepath.Rel(base, target)
		return err == nil && filepath.IsLocal(rel)
	}
	return within(a, b) || within(b, a)
}
