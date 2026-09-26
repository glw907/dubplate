package backup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// snapshotKeep is how many remote snapshots a prune keeps in total,
// the run's own included.
const snapshotKeep = 14

// othersKeep is how many of the remaining remote snapshots a prune
// keeps besides the run's own, which is always kept by name.
const othersKeep = snapshotKeep - 1

// archiveRetention is how long an archive directory survives past its
// stamp.
const archiveRetention = 90 * 24 * time.Hour

// A prune is one run's pruning of the remote's snapshots and archive,
// and of the local snapshots directory, run once the send has
// succeeded.
type prune struct {
	t        transport
	remote   Remote
	stamp    string
	localDir string
	start    time.Time
}

// run prunes p's remote snapshots, its archive, and its local copies.
// It never fails the run: every step logs its own error to logger and
// folds it into the returned text, which becomes the succeeded
// record's final progress. An empty result means nothing needs
// reporting.
func (p prune) run(ctx context.Context, logger *slog.Logger) string {
	entries, err := p.t.list(ctx, p.remote.join("snapshots"), false)
	if err != nil {
		logger.Warn("prune: list remote snapshots", "err", err)
		return "prune skipped: list remote snapshots: " + err.Error()
	}
	if reason, ok := p.checkSafety(entries); !ok {
		logger.Warn("prune skipped", "reason", reason)
		return "prune skipped: " + reason
	}

	var failed []string
	if err := p.pruneSnapshots(ctx, entries); err != nil {
		logger.Warn("prune: remote snapshots", "err", err)
		failed = append(failed, err.Error())
	}
	if err := p.pruneArchive(ctx); err != nil {
		logger.Warn("prune: archive", "err", err)
		failed = append(failed, err.Error())
	}
	if err := p.pruneLocal(); err != nil {
		logger.Warn("prune: local copies", "err", err)
		failed = append(failed, err.Error())
	}
	if len(failed) == 0 {
		return ""
	}
	return "prune failed: " + strings.Join(failed, "; ")
}

// checkSafety reports whether entries, snapshots/'s listing, holds the
// run's own snapshot with its stamp strictly greater than every
// other's. A sane clock is the only case that holds: a stepped-back
// clock, a reset RTC, or a restored box writing to the same remote
// would otherwise make keep-newest delete the snapshot this run just
// sent. reason names why the check failed; it is empty when ok is
// true.
func (p prune) checkSafety(entries []remoteEntry) (reason string, ok bool) {
	found := false
	for _, e := range entries {
		stamp, isStamp := e.snapshotStamp()
		if !isStamp {
			continue
		}
		if stamp == p.stamp {
			found = true
			continue
		}
		if stamp >= p.stamp {
			return fmt.Sprintf("remote snapshot %s is not before this run's own %s", stamp, p.stamp), false
		}
	}
	if !found {
		return fmt.Sprintf("this run's own snapshot %s is missing from the remote listing", p.stamp), false
	}
	return "", true
}

// pruneSnapshots deletes every remote snapshot beyond the run's own
// and the [othersKeep] newest of the rest, [snapshotKeep] in total. A
// name that does not parse as <stamp>.db is left alone.
func (p prune) pruneSnapshots(ctx context.Context, entries []remoteEntry) error {
	var others []string
	for _, e := range entries {
		stamp, isStamp := e.snapshotStamp()
		if !isStamp || stamp == p.stamp {
			continue
		}
		others = append(others, stamp)
	}
	slices.Sort(others)
	slices.Reverse(others)
	if len(others) <= othersKeep {
		return nil
	}

	var errs []error
	for _, stamp := range others[othersKeep:] {
		dest := p.remote.join("snapshots/" + stamp + ".db")
		if err := p.t.run(ctx, nil, "deletefile", dest); err != nil {
			errs = append(errs, fmt.Errorf("delete %s: %w", dest, err))
		}
	}
	return errors.Join(errs...)
}

// pruneArchive purges every archive/<stamp>/ directory more than
// [archiveRetention] before the run's stamp. A name that does not
// parse as a stamp is left alone.
func (p prune) pruneArchive(ctx context.Context) error {
	entries, err := p.t.list(ctx, p.remote.join("archive"), true)
	if err != nil {
		return fmt.Errorf("list archive: %w", err)
	}
	own, err := store.ParseStamp(p.stamp)
	if err != nil {
		return fmt.Errorf("parse run stamp %s: %w", p.stamp, err)
	}

	var errs []error
	for _, e := range entries {
		if !e.IsDir {
			continue
		}
		stamp, err := store.ParseStamp(e.Name)
		if err != nil {
			continue
		}
		if own.Sub(stamp) <= archiveRetention {
			continue
		}
		dest := p.remote.join("archive/" + e.Name)
		if err := p.t.run(ctx, nil, "purge", dest); err != nil {
			errs = append(errs, fmt.Errorf("purge %s: %w", dest, err))
		}
	}
	return errors.Join(errs...)
}

// pruneLocal removes every local pre-migration copy in p.localDir
// modified before p.start. A copy written after the run began was not
// captured by it and stays. A pre-restore copy always stays: it holds
// the rows its restore did not bring back, so no backup captured them,
// and like a discarded store it is removed only by hand. The reference
// and every other name are untouched.
func (p prune) pruneLocal() error {
	entries, err := os.ReadDir(p.localDir)
	if err != nil {
		return fmt.Errorf("read %s: %w", p.localDir, err)
	}

	var errs []error
	for _, e := range entries {
		if e.IsDir() || !store.IsPreMigrateCopy(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			errs = append(errs, fmt.Errorf("stat %s: %w", e.Name(), err))
			continue
		}
		if !info.ModTime().Before(p.start) {
			continue
		}
		path := filepath.Join(p.localDir, e.Name())
		if err := os.Remove(path); err != nil {
			errs = append(errs, fmt.Errorf("remove %s: %w", path, err))
		}
	}
	return errors.Join(errs...)
}
