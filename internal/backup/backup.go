// Package backup runs one dubplate backup: a consistent sqlite3 snapshot
// of the beets and Navidrome databases, then an rclone sync of the
// library and the snapshot to R2. It replaces scripts/music-backup.
package backup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/ping"
)

// ErrLibraryMissing marks a backup run aborted because
// [config.Paths.LibraryDir] does not exist. A missing library is never a
// legitimate "nothing to back up" state -- it means a mount never came up
// or the tree was wiped -- and rclone must never see it: sync treats an
// empty source as the new truth and would delete every object already in
// R2 to match.
var ErrLibraryMissing = errors.New("library directory does not exist")

// ErrLibraryEmpty marks a backup run aborted because LibraryDir exists
// but holds nothing, the same rclone-sync hazard [ErrLibraryMissing]
// guards against.
var ErrLibraryEmpty = errors.New("library directory is empty")

// pingSlug is the check name a run's success ping resolves through
// [config.Config.PingURLs], matching the shell layer's `ping_hc backup`.
const pingSlug = "backup"

// lockName is the backup lock's file name within [config.Paths.StateDir].
// It is deliberately separate from [lock.Path]'s import lock: an
// in-progress import and a backup run are unrelated operations that must
// stay free to overlap, only two backup runs must not.
const lockName = "backup.lock"

// Runner runs one backup, snapshotting the beets and Navidrome databases
// and syncing the library and snapshot to R2.
type Runner struct {
	Cfg  *config.Config
	Ping *ping.Client

	// SqlitePath and RclonePath are the snapshot and sync subprocess
	// binaries; test seams, production sets both resolved on PATH.
	SqlitePath string
	RclonePath string

	// Now is the clock seam Run reads for the R2 archive-dir date stamp.
	// Production sets it to time.Now; tests pin it to a fixed instant.
	Now func() time.Time
}

// Run executes one backup. A held lock is healthy -- another backup run
// (rclone syncing a large tree can run past a single day) is still
// going -- and is reported success, touching nothing else. Every other
// abort returns a wrapped error and relies on the caller's OnFailure=
// chain (`dubplate ping-fail backup`) to page the admin; Run never sends
// a failure ping itself.
func (r *Runner) Run(ctx context.Context) error {
	l, err := lock.Acquire(lockPath(r.Cfg.Paths.StateDir))
	if errors.Is(err, lock.ErrHeld) {
		r.pingSuccess(ctx)
		return nil
	}
	if err != nil {
		return fmt.Errorf("acquire backup lock: %w", err)
	}
	defer func() {
		if err := l.Release(); err != nil {
			slog.Error("backup: release backup lock", "err", err)
		}
	}()

	if err := checkLibraryDir(r.Cfg.Paths.LibraryDir); err != nil {
		return err
	}

	if err := os.RemoveAll(r.Cfg.Paths.BackupStaging); err != nil {
		return fmt.Errorf("clear backup staging %s: %w", r.Cfg.Paths.BackupStaging, err)
	}
	if err := os.MkdirAll(r.Cfg.Paths.BackupStaging, 0o755); err != nil {
		return fmt.Errorf("create backup staging %s: %w", r.Cfg.Paths.BackupStaging, err)
	}

	if err := r.snapshotDB(ctx, r.Cfg.BeetsDB, filepath.Join(r.Cfg.Paths.BackupStaging, "library.db")); err != nil {
		return err
	}
	if err := r.snapshotDB(ctx, r.Cfg.NavidromeDB, filepath.Join(r.Cfg.Paths.BackupStaging, "navidrome.db")); err != nil {
		return err
	}

	dateStr := r.Now().Format("2006-01-02")
	if err := r.rcloneSync(ctx, r.Cfg.Paths.LibraryDir, "library", dateStr); err != nil {
		return err
	}
	if err := r.rcloneSync(ctx, r.Cfg.Paths.BackupStaging, "state", dateStr); err != nil {
		return err
	}

	r.pingSuccess(ctx)
	return nil
}

// lockPath returns the backup lock's path within stateDir: the
// persistent state directory, not a systemd RuntimeDirectory, since the
// lock itself -- not its content -- must outlive a unit restart.
func lockPath(stateDir string) string {
	return filepath.Join(stateDir, lockName)
}

// checkLibraryDir guards the two states [Runner.Run] must never hand to
// rclone sync: a missing directory (a mount that never came up) and an
// empty one (nothing staged yet, or a wipe). Reading only dir's top-level
// entries is enough -- any entry nested deeper still has an ancestor
// entry at the top level, so an empty top level and an empty tree are the
// same fact.
func checkLibraryDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s: %w", dir, ErrLibraryMissing)
	}
	if err != nil {
		return fmt.Errorf("read library directory %s: %w", dir, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("%s: %w", dir, ErrLibraryEmpty)
	}
	return nil
}

// snapshotDB takes a consistent sqlite3 snapshot of dbPath into destPath
// via the `.backup` dot-command, rather than letting rclone sync read a
// live database file directly: a mid-write copy would be a torn backup
// that opens fine and restores garbage. A missing database (beets or
// Navidrome not yet run on a fresh box) is logged and skipped, never
// aborted.
func (r *Runner) snapshotDB(ctx context.Context, dbPath, destPath string) error {
	if _, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("backup: database not found; skipping", "db", dbPath)
			return nil
		}
		return fmt.Errorf("stat database %s: %w", dbPath, err)
	}

	// The single quotes are sqlite3's own dot-command tokenizer, not
	// shell quoting: exec passes this as one argv element regardless, but
	// sqlite3 itself still splits ".backup <path>" on whitespace, so a
	// destPath containing a space needs them to parse as one token.
	cmd := exec.CommandContext(ctx, r.SqlitePath, dbPath, fmt.Sprintf(".backup '%s'", destPath))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sqlite3 backup %s: %w: %s", dbPath, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// rcloneSync syncs src into the R2 bucket's tree, moving anything sync
// would otherwise delete under `_archive/<tree>/<dateStr>` instead of
// discarding it -- the shell layer's pinned layout, untouched. One sync
// per tree, each pinned to its own --backup-dir: a single root-level sync
// would have each tree's run delete the other tree's objects, and rclone
// refuses an overlapping backup-dir outright.
func (r *Runner) rcloneSync(ctx context.Context, src, tree, dateStr string) error {
	remote := r.Cfg.R2.Remote
	dst := remote + ":" + r.Cfg.R2.Bucket + "/" + tree
	backupDir := remote + ":" + r.Cfg.R2.Bucket + "/_archive/" + tree + "/" + dateStr

	cmd := exec.CommandContext(ctx, r.RclonePath, "sync", src, dst,
		"--backup-dir", backupDir, "--exclude", "/_archive/**")
	cmd.Env = append(os.Environ(), rcloneEnv(r.Cfg.R2)...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rclone sync %s to %s: %w: %s", src, dst, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// rcloneEnv builds the RCLONE_CONFIG_<REMOTE>_* variables that define r2
// as an on-the-fly rclone remote, keyed by r2.Remote uppercased.
// NO_CHECK_BUCKET matters: the token is bucket-scoped (no
// ListBuckets/CreateBucket), and --backup-dir's first server-side archive
// move otherwise probes the bucket first, 403s, and rclone falls back to
// CreateBucket, which the token cannot do -- Cloudflare's R2 rclone guide
// prescribes this flag for object-scoped tokens.
func rcloneEnv(r2 config.R2Config) []string {
	prefix := "RCLONE_CONFIG_" + strings.ToUpper(strings.ReplaceAll(r2.Remote, "-", "_")) + "_"
	return []string{
		prefix + "TYPE=s3",
		prefix + "NO_CHECK_BUCKET=true",
		prefix + "PROVIDER=Cloudflare",
		prefix + "ENDPOINT=https://" + r2.AccountID + ".r2.cloudflarestorage.com",
		prefix + "ACCESS_KEY_ID=" + r2.AccessKeyID,
		prefix + "SECRET_ACCESS_KEY=" + r2.SecretAccessKey,
	}
}

// pingSuccess reports this run's success. A ping failure is best-effort
// margin: logged, never returned, per the doctrine that monitoring must
// never take down the pipeline it monitors.
func (r *Runner) pingSuccess(ctx context.Context) {
	if err := r.Ping.Success(ctx, r.Cfg.PingURLs[pingSlug]); err != nil {
		slog.Warn("backup: success ping", "err", err)
	}
}
