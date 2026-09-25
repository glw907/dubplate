package backup

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/glw907/dubplate/internal/engine/store"
)

// guardFraction is the share of the catalog's tracks a library tree
// must hold as regular files before a sync may run.
const guardFraction = 0.5

// resolveRoot resolves every symlink in root and returns the directory
// it names. filepath.WalkDir does not follow a symlinked root while
// rclone does, so the count guard and the sync must both read the
// resolved path.
func resolveRoot(root string) (string, error) {
	if root == "" {
		return "", fail("library-missing", "no library root is configured")
	}
	resolved, err := filepath.EvalSymlinks(root)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
		return "", fail("library-missing", "library root %s does not exist", root)
	}
	if err != nil {
		return "", fail("library-unreadable", "resolve library root %s: %v", root, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fail("library-unreadable", "stat library root %s: %v", resolved, err)
	}
	if !info.IsDir() {
		return "", fail("library-missing", "library root %s is not a directory", resolved)
	}
	return resolved, nil
}

// checkLibrary runs the guards on the resolved root against h's
// catalog, and returns the --max-delete bound the sync takes.
func checkLibrary(ctx context.Context, h *store.Handle, root string) (int, error) {
	var tracks int
	if err := h.QueryRow(ctx, "SELECT count(*) FROM tracks").Scan(&tracks); err != nil {
		return 0, fmt.Errorf("count catalog tracks: %w", err)
	}
	files, err := countFiles(root)
	if err != nil {
		return 0, fail("library-unreadable", "read library root %s: %v", root, err)
	}
	return guard(root, files, tracks)
}

// countFiles counts the regular files under root. Directories,
// symlinks, and every other kind of entry count for nothing.
func countFiles(root string) (int, error) {
	n := 0
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			n++
		}
		return nil
	})
	return n, err
}

// guard decides whether a tree holding files regular files may be
// synced over a catalog of tracks tracks, and returns the sync's
// --max-delete bound. With tracks in the catalog, the bound is the
// track count. With none, the bound is 0: a move into the archive
// counts as a delete, so no file on the remote can be archived.
func guard(root string, files, tracks int) (int, error) {
	if tracks == 0 {
		if files == 0 {
			return 0, fail("library-empty", "library root %s holds no files", root)
		}
		return 0, nil
	}
	if float64(files) < guardFraction*float64(tracks) {
		return 0, fail("library-short", "library root %s holds %d files, fewer than half the catalog's %d tracks", root, files, tracks)
	}
	return tracks, nil
}
