// Package review files staged items that need a human's attention --
// something the FLAC quality gate rejected, or beets could not
// confidently tag -- into a review or quarantine tree, and prunes both
// trees on a retention schedule. It replaces the shell layer's
// unique_dest_path, safe_mv_to_review, and sweep_old_dirs.
package review

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrCouldNotMove indicates [Filer.MoveToReview] could not place path
// in either f.ReviewDir or f.QuarantineDir. path is unchanged: still at
// its original location, and needs an operator's manual attention --
// the shell layer's untested "COULD NOT MOVE BY ANY MEANS" branch.
var ErrCouldNotMove = errors.New("could not move by any means")

// Filer files items into a review or quarantine tree and prunes them by
// age. ReviewDir and QuarantineDir are whatever directory the caller
// wants a given call to act on: [Filer.MoveToReview] treats the field
// as the direct destination (and, on failure, the fallback), so a
// caller wanting the shell layer's review/<date>/<user>/ nesting builds
// that path into ReviewDir before calling; [Filer.Sweep] treats the
// field as holding top-level dated subdirectories to prune by age. Now
// is the clock seam both methods read from -- MoveToReview's collision
// suffix, Sweep's age cutoff -- production sets it to time.Now, tests
// pin it to a fixed instant.
type Filer struct {
	ReviewDir     string
	QuarantineDir string
	Now           func() time.Time
}

// MoveToReview moves path into f.ReviewDir under a name
// [Filer.uniqueDestPath] guarantees is free, and returns the resulting
// destination. If that move fails for any reason -- permissions,
// cross-device, a full disk -- MoveToReview retries once into
// f.QuarantineDir as a second chance, so one unmovable quality-gate
// failure never leaves staging holding a file beet import would
// otherwise ingest. The caller can tell the two outcomes apart by
// checking whether the returned path lies under f.QuarantineDir.
// MoveToReview returns a non-nil, [ErrCouldNotMove]-wrapping error only
// when both moves fail.
func (f *Filer) MoveToReview(path string) (string, error) {
	dest, err := f.moveInto(path, f.ReviewDir)
	if err == nil {
		return dest, nil
	}

	dest, qerr := f.moveInto(path, f.QuarantineDir)
	if qerr == nil {
		return dest, nil
	}

	return "", fmt.Errorf("%s: review (%v), quarantine (%v): %w", path, err, qerr, ErrCouldNotMove)
}

// moveInto moves path into destDir under a name uniqueDestPath
// guarantees is free, creating destDir first if it does not exist.
func (f *Filer) moveInto(path, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create directory %s: %w", destDir, err)
	}

	dest := f.uniqueDestPath(destDir, filepath.Base(path))
	if err := Move(path, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// uniqueDestPath returns dir joined with name if nothing is there yet,
// or a disambiguated path otherwise: first an HHMMSS suffix from
// f.Now, then -- if even that collides -- a numeric ".2", ".3", ...
// suffix. It never returns a path that already exists, so a same-day
// repeat quarantine of an identically named item lands beside the
// first instead of merging into or clobbering it. Matches
// unique_dest_path at HEAD.
func (f *Filer) uniqueDestPath(dir, name string) string {
	plain := filepath.Join(dir, name)
	if !exists(plain) {
		return plain
	}

	stem := filepath.Join(dir, name+"-"+f.Now().Format("150405"))
	if !exists(stem) {
		return stem
	}

	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s.%d", stem, n)
		if !exists(candidate) {
			return candidate
		}
	}
}

// exists reports whether path can be stat'd, matching the shell
// layer's `[ -e path ]` (which, like os.Stat, follows symlinks).
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Sweep removes every top-level subdirectory of f.ReviewDir and
// f.QuarantineDir whose modification time is older than retention,
// relative to f.Now, so neither tree grows without bound if nobody
// drains it by hand. Matches sweep_old_dirs at HEAD, applied to both
// directories -- fix-wave behavior, since the shell layer originally
// swept only REVIEW_DIR. A directory that does not exist yet is not an
// error: QuarantineDir in particular is never created ahead of its
// first fallback move.
func (f *Filer) Sweep(retention time.Duration) error {
	cutoff := f.Now().Add(-retention)
	if err := sweepDir(f.ReviewDir, cutoff); err != nil {
		return err
	}
	return sweepDir(f.QuarantineDir, cutoff)
}

func sweepDir(dir string, cutoff time.Time) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read directory %s: %w", dir, err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", filepath.Join(dir, e.Name()), err)
		}
		if info.ModTime().After(cutoff) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return fmt.Errorf("remove %s: %w", filepath.Join(dir, e.Name()), err)
		}
	}
	return nil
}
