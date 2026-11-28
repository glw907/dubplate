package auth

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/settings"
)

var errNonceReplayed = errors.New("nonce already consumed")

// nonceRecordVersion labels the record layout. A new layout takes a new
// label, so a record is read with the meanings it was written under, and
// one from another layout is unparseable rather than misread.
const nonceRecordVersion = "dubplate-nonce-1"

const nonceRecordFieldCount = 3

// errNonceRecordUnparseable is the sweep's internal signal; sweep never
// returns it.
var errNonceRecordUnparseable = errors.New("nonce record: unparseable")

// A NonceStore is the set of consumed magic-link nonces, one record per
// nonce under the auth dir's nonce subdirectory.
type NonceStore struct {
	authDir *Dir

	dir string

	// retention is the shortest life a new record is written with. Each
	// record carries its own life and the sweep prunes on that, so
	// lowering retention never shortens a record already on disk.
	retention settings.Reader[time.Duration]

	// linkExpiryCeiling is the registered ceiling on the magic-link
	// expiry, never the value in force ([NonceStore.mtimeFallback]).
	linkExpiryCeiling time.Duration
}

// A NonceSweep reports one sweep's counts. Malformed counts records
// retained because they could not be read or parsed, a crash or
// permissions signal worth logging.
type NonceSweep struct {
	Removed   int
	Malformed int
}

// OpenNonceStore returns the consumed-nonce store under authDir and
// sweeps it once at now. retention reads DUBPLATE_NONCE_RETENTION, and
// linkExpiryCeiling is the registered ceiling on the link expiry. The
// caller runs [NonceStore.RunSweep] thereafter.
func OpenNonceStore(ctx context.Context, authDir *Dir, retention settings.Reader[time.Duration], linkExpiryCeiling time.Duration, now time.Time) (*NonceStore, NonceSweep, error) {
	s := &NonceStore{
		authDir:              authDir,
		dir:               authDir.subdir(nonceSubdir),
		retention:         retention,
		linkExpiryCeiling: linkExpiryCeiling,
	}
	sweep, err := s.sweep(ctx, now)
	if err != nil {
		return nil, sweep, err
	}
	return s, sweep, nil
}

// SweepInterval is the period the serve verb drives RunSweep at. Replay
// protection never depends on how promptly a swept record disappears,
// so the interval is fixed rather than configurable.
const SweepInterval = time.Hour

// RunSweep sweeps at now() for every value received on tick until ctx
// is done; the caller owns the ticker. A sweep's error is logged, never
// returned, so one failure does not stop later sweeps. A failed setting
// read logs a fixed reason naming its key, never the read's text.
func (s *NonceStore) RunSweep(ctx context.Context, tick <-chan time.Time, now func() time.Time, logger *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			result, err := s.sweep(ctx, now())
			if err != nil {
				reason := err.Error()
				if errors.Is(err, errSettingUnreadable) {
					reason = settingFailureReason(err)
				}
				logger.Error("nonce sweep failed", "reason", reason)
				continue
			}
			logger.Info("nonce sweep completed", "removed", result.Removed, "malformed", result.Malformed)
		}
	}
}

// insert records nonce as consumed at now and returns errNonceReplayed
// when a record exists. expiresAt is the expiry of the token that bore
// the nonce; the caller inserts only after it verifies, so an
// unauthenticated value cannot fill the store. The record outlives that
// token whatever retention is configured. Retention is read first, so a
// failed read, wrapping errSettingUnreadable, consumes nothing.
func (s *NonceStore) insert(ctx context.Context, nonce [nonceLength]byte, now, expiresAt time.Time) error {
	retention, err := readSetting(ctx, s.retention)
	if err != nil {
		return err
	}
	path, err := s.authDir.noncePath(recordID(nonceSubdir, string(nonce[:])))
	if err != nil {
		return err
	}

	// O_EXCL, not atomicfile: its rename over the target would let a
	// replay replace the record that proves consumption and succeed.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errNonceReplayed
		}
		return fmt.Errorf("create consumed nonce record: %w", err)
	}
	consumedAt := now.UTC().Truncate(time.Second)
	sweepableAt := consumedAt.Add(retention)
	if expiresAt.After(sweepableAt) {
		sweepableAt = expiresAt
	}
	if err := writeAndSync(f, nonceRecord(consumedAt, sweepableAt)); err != nil {
		return err
	}

	// Without the directory sync, a crash can lose the record's name
	// while its link is still live, and single use with it.
	if err := atomicfile.SyncDir(s.dir); err != nil {
		return fmt.Errorf("sync nonce directory: %w", err)
	}
	return nil
}

// sweep removes every record whose carried expiry has passed at now. A
// record at exactly its expiry is retained.
//
// A record the sweep cannot read or parse is retained and counted, not
// failed: its presence is what blocks a replay, and a crash between
// create and write leaves a zero-length file whose token may still
// verify. It goes once its [NonceStore.mtimeFallback] bound passes.
//
// sweep fails closed only on a failed retention read, removing nothing,
// and on an unreadable store directory.
func (s *NonceStore) sweep(ctx context.Context, now time.Time) (NonceSweep, error) {
	retention, err := readSetting(ctx, s.retention)
	if err != nil {
		return NonceSweep{}, fmt.Errorf("sweep consumed nonces: %w", err)
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return NonceSweep{}, fmt.Errorf("sweep consumed nonces: %w", err)
	}

	var sweep NonceSweep
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		expiresAt, err := s.recordExpiry(path, retention)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			// Another sweep removed it since the listing.
			continue
		case errors.Is(err, errNonceRecordUnparseable):
			if !now.After(expiresAt) {
				sweep.Malformed++
				continue
			}
		case err != nil:
			// The fallback's stat failed, most often on permissions.
			// Retain and count it rather than return an error: one
			// bad record must not stop OpenNonceStore starting.
			sweep.Malformed++
			continue
		case !now.After(expiresAt):
			continue
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
			continue
		}
		sweep.Removed++
	}
	if len(errs) > 0 {
		return sweep, fmt.Errorf("sweep consumed nonces: %w", errors.Join(errs...))
	}
	return sweep, nil
}

// recordExpiry returns when the record at path may be removed. For a
// record it cannot read or parse, it returns the mtime fallback with
// errNonceRecordUnparseable; an absent file returns fs.ErrNotExist.
func (s *NonceStore) recordExpiry(path string, retention time.Duration) (time.Time, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return time.Time{}, err
		}
		return s.mtimeFallback(path, retention)
	}
	expiresAt, err := parseNonceRecord(data)
	if err == nil {
		return expiresAt, nil
	}
	return s.mtimeFallback(path, retention)
}

// mtimeFallback dates a record whose bytes cannot be read or parsed at
// its mtime plus max(retention, s.linkExpiryCeiling). A readable record
// lives at least as long as its token and no token outlives the
// ceiling, so the bound is never shorter than the record it stands for.
// With retention alone, lowering it and the link expiry together could
// sweep an unreadable record while its link can still replay.
func (s *NonceStore) mtimeFallback(path string, retention time.Duration) (time.Time, error) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		return time.Time{}, statErr
	}
	return info.ModTime().Add(max(retention, s.linkExpiryCeiling)), errNonceRecordUnparseable
}

// nonceRecord renders the version, the consumed-at second, and the life
// in seconds, rounded up so rounding can only lengthen it.
func nonceRecord(consumedAt, sweepableAt time.Time) []byte {
	life := sweepableAt.Sub(consumedAt)
	return fmt.Appendf(nil, "%s %d %d\n",
		nonceRecordVersion,
		consumedAt.Unix(),
		int64((life+time.Second-1)/time.Second),
	)
}

// parseNonceRecord returns when the sweep may remove the record.
func parseNonceRecord(data []byte) (time.Time, error) {
	fields := strings.Fields(string(data))
	if len(fields) != nonceRecordFieldCount || fields[0] != nonceRecordVersion {
		return time.Time{}, errNonceRecordUnparseable
	}
	consumed, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return time.Time{}, errNonceRecordUnparseable
	}
	retention, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return time.Time{}, errNonceRecordUnparseable
	}
	return time.Unix(consumed, 0).UTC().Add(time.Duration(retention) * time.Second), nil
}

// writeAndSync closes f on every path and reports the first failure.
func writeAndSync(f *os.File, data []byte) error {
	_, err := f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("write consumed nonce record: %w", err)
	}
	return nil
}
