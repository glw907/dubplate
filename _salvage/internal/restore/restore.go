// Package restore pulls the library and database snapshots back from R2
// into a target tree: backup's read-only inverse, sharing its rclone and
// sqlite3 subprocess seams. It is the DR drill's executable half and the
// adopter's migration path for bringing an existing collection into a
// fresh estate.
package restore

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// ErrTargetNotEmpty marks a restore refused because the target directory
// already holds entries and Options.Force was not set. Pulling into an
// occupied tree by mistake is exactly the accident --force exists to
// require deliberately.
var ErrTargetNotEmpty = errors.New("restore target directory is not empty")

// ErrIntegrityCheck marks a restored database that failed sqlite3's
// "PRAGMA integrity_check": a snapshot that opens but does not verify is
// not a usable restore.
var ErrIntegrityCheck = errors.New("restored database failed integrity check")

// trees are the R2 bucket's top-level directories a restore pulls,
// matching [backup.Runner]'s own tree names.
var trees = []string{"library", "state"}

// databaseFiles are the snapshot file names backup writes under the
// state tree, each verified once restored.
var databaseFiles = []string{"library.db", "navidrome.db"}

// Options configures one restore run.
type Options struct {
	// To is the target directory to restore into.
	To string
	// Force allows restoring into a target directory that already holds
	// entries.
	Force bool
}

// TreeReport is one bucket tree's restore outcome.
type TreeReport struct {
	Name          string `json:"name"`
	RestoredFiles int    `json:"restored_files"`
	// BucketOnly lists a path lsjson found in the bucket tree that the
	// local tree does not have after copy -- a sanity check independent
	// of rclone's own exit code, best-effort: a listing failure leaves
	// this empty rather than failing the run.
	BucketOnly []string `json:"bucket_only,omitempty"`
}

// ArchiveEntry is one dated tree found under the bucket's
// `_archive/<tree>/<date>` layout while listing, without pulling it.
type ArchiveEntry struct {
	Tree string `json:"tree"`
	Date string `json:"date"`
}

// DatabaseCheck is one restored database's sqlite3 integrity result.
type DatabaseCheck struct {
	Path string `json:"path"`
	// Present is false when the database was never backed up, so
	// nothing arrived at Path to check.
	Present bool `json:"present"`
	OK      bool `json:"ok"`
	// Detail carries the integrity_check output verbatim when the
	// database is present: "ok" on success, sqlite3's error lines
	// otherwise.
	Detail string `json:"detail,omitempty"`
}

// Report is one restore run's full record.
type Report struct {
	To        string          `json:"to"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   time.Time       `json:"ended_at"`
	Trees     []TreeReport    `json:"trees"`
	Archives  []ArchiveEntry  `json:"archives"`
	Databases []DatabaseCheck `json:"databases"`
}

// Runner runs one restore, pulling `library/` and `state/` from R2 into
// a target tree via `rclone copy` -- never sync, so the bucket can never
// lose an object to this command -- listing the dated `_archive/` trees
// without pulling them, and verifying every restored database via
// sqlite3.
type Runner struct {
	Cfg *config.Config

	// SqlitePath and RclonePath are the integrity-check and copy
	// subprocess binaries, backup's own seams; test seams, production
	// resolves both on PATH.
	SqlitePath string
	RclonePath string

	// Now is the clock seam Run reads for the report's timestamps;
	// production sets it to time.Now, tests pin it.
	Now func() time.Time
}

// Run executes one restore into opts.To, returning the run [Report]
// alongside any error so a caller can print what happened even on a
// hard failure. A nil Report means Run never reached the copy phase (the
// target refusal or a directory-creation failure).
func (r *Runner) Run(ctx context.Context, opts Options) (*Report, error) {
	if err := checkTarget(opts.To, opts.Force); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(opts.To, 0o755); err != nil {
		return nil, fmt.Errorf("create target directory %s: %w", opts.To, err)
	}

	report := &Report{To: opts.To, StartedAt: r.Now()}
	defer func() { report.EndedAt = r.Now() }()

	for _, tree := range trees {
		tr, err := r.restoreTree(ctx, tree, opts.To)
		if err != nil {
			return report, err
		}
		report.Trees = append(report.Trees, tr)
	}

	report.Archives = r.listArchives(ctx)

	for _, name := range databaseFiles {
		check, err := r.verifyDatabase(ctx, filepath.Join(opts.To, "state", name))
		report.Databases = append(report.Databases, check)
		if err != nil {
			return report, err
		}
	}

	return report, nil
}

// checkTarget refuses a target directory that already holds entries
// unless force is set.
func checkTarget(to string, force bool) error {
	if force {
		return nil
	}
	entries, err := os.ReadDir(to)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read target directory %s: %w", to, err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s: %w", to, ErrTargetNotEmpty)
	}
	return nil
}

// restoreTree pulls one bucket tree into <to>/<tree> via `rclone copy`,
// then counts what actually arrived locally and lists any bucket path
// that is not there -- a sanity check independent of rclone's own exit
// code.
func (r *Runner) restoreTree(ctx context.Context, tree, to string) (TreeReport, error) {
	dest := filepath.Join(to, tree)
	src := r.remotePath(tree)

	cmd := exec.CommandContext(ctx, r.RclonePath, "copy", src, dest)
	cmd.Env = append(os.Environ(), rcloneEnv(r.Cfg.R2)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return TreeReport{}, fmt.Errorf("rclone copy %s to %s: %w: %s", src, dest, err, strings.TrimSpace(string(out)))
	}

	restored, err := countFiles(dest)
	if err != nil {
		return TreeReport{}, err
	}

	return TreeReport{Name: tree, RestoredFiles: restored, BucketOnly: r.bucketOnly(ctx, src, dest)}, nil
}

// bucketOnly lists every file lsjson reports under src that dest does
// not have. It is best-effort report detail, not part of the restore's
// core guarantee: a listing failure is logged and reported as no known
// differences rather than failing the run.
func (r *Runner) bucketOnly(ctx context.Context, src, dest string) []string {
	entries, err := r.listBucket(ctx, src, "--files-only")
	if err != nil {
		slog.Warn("restore: list bucket tree for comparison", "src", src, "err", err)
		return nil
	}

	var missing []string
	for _, e := range entries {
		if _, err := os.Stat(filepath.Join(dest, e.Path)); errors.Is(err, os.ErrNotExist) {
			missing = append(missing, e.Path)
		}
	}
	slices.Sort(missing)
	return missing
}

// listArchives lists the bucket's dated `_archive/<tree>/<date>` layout
// without pulling any of it. Best-effort, like [Runner.bucketOnly]: a
// listing failure is logged and reported as no known archives.
func (r *Runner) listArchives(ctx context.Context) []ArchiveEntry {
	entries, err := r.listBucket(ctx, r.remotePath("_archive"), "--dirs-only")
	if err != nil {
		slog.Warn("restore: list archive trees", "err", err)
		return nil
	}

	var archives []ArchiveEntry
	for _, e := range entries {
		tree, date, ok := strings.Cut(e.Path, "/")
		if !ok || strings.Contains(date, "/") {
			continue // depth-1 tree directories and anything nested deeper than tree/date
		}
		archives = append(archives, ArchiveEntry{Tree: tree, Date: date})
	}
	slices.SortFunc(archives, func(a, b ArchiveEntry) int {
		return cmp.Or(cmp.Compare(a.Tree, b.Tree), cmp.Compare(a.Date, b.Date))
	})
	return archives
}

// lsjsonEntry is the subset of `rclone lsjson`'s per-entry fields
// [Runner.listBucket] reads.
type lsjsonEntry struct {
	Path  string `json:"Path"`
	IsDir bool   `json:"IsDir"`
}

// listBucket runs `rclone lsjson --recursive` against src, the one
// read-only listing call [Runner.bucketOnly] and [Runner.listArchives]
// both need.
func (r *Runner) listBucket(ctx context.Context, src string, extra ...string) ([]lsjsonEntry, error) {
	args := append([]string{"lsjson", "--recursive"}, extra...)
	args = append(args, src)

	cmd := exec.CommandContext(ctx, r.RclonePath, args...)
	cmd.Env = append(os.Environ(), rcloneEnv(r.Cfg.R2)...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("rclone lsjson %s: %w", src, err)
	}

	var entries []lsjsonEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("parse rclone lsjson output for %s: %w", src, err)
	}
	return entries, nil
}

// verifyDatabase runs sqlite3's own consistency check against a restored
// database. A snapshot that opens but does not verify is a hard failure,
// unlike backup's treatment of a missing source database: this is the DR
// drill's correctness guarantee, not an optional probe.
func (r *Runner) verifyDatabase(ctx context.Context, path string) (DatabaseCheck, error) {
	check := DatabaseCheck{Path: path}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return check, nil // never backed up; nothing to verify
		}
		return check, fmt.Errorf("stat restored database %s: %w", path, err)
	}
	check.Present = true

	cmd := exec.CommandContext(ctx, r.SqlitePath, path, "PRAGMA integrity_check;")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return check, fmt.Errorf("sqlite3 integrity check %s: %w: %s", path, err, strings.TrimSpace(string(out)))
	}

	check.Detail = strings.TrimSpace(string(out))
	if check.Detail != "ok" {
		return check, fmt.Errorf("%s: %w: %s", path, ErrIntegrityCheck, check.Detail)
	}
	check.OK = true
	return check, nil
}

// remotePath builds the rclone remote path for one bucket-relative
// path, matching [backup.Runner]'s own "<remote>:<bucket>/<tree>" shape.
func (r *Runner) remotePath(sub string) string {
	return r.Cfg.R2.Remote + ":" + r.Cfg.R2.Bucket + "/" + sub
}

// countFiles returns the number of regular files under dir, walking the
// tree rclone copy just populated -- the restored-file count the report
// carries, independent of what rclone itself claims to have transferred.
// A dir that copy never created (an empty bucket tree) counts as zero,
// not an error.
func countFiles(dir string) (int, error) {
	n := 0
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			n++
		}
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("count restored files in %s: %w", dir, err)
	}
	return n, nil
}

// rcloneEnv builds the RCLONE_CONFIG_<REMOTE>_* variables that define r2
// as an on-the-fly rclone remote. Mirrors [backup.Runner]'s own copy of
// this helper (kept separate rather than shared: each is a small,
// single-purpose adapter over its own package's subprocess call, and
// restore's is read-only where backup's is not).
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
