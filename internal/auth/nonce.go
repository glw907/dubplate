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
)

// ErrNonceReplayed reports a nonce whose record already exists: the
// link carrying it has been consumed once and every later presentation
// of it is a replay.
var ErrNonceReplayed = errors.New("auth: nonce already consumed")

// nonceRecordVersion labels the record's field layout. A later field
// lands under a new label rather than changing what an existing field
// means, so a record written before the change is read with the
// meanings it was written under, and one written after a downgrade is
// unparseable rather than misread.
const nonceRecordVersion = "dubplate-nonce-1"

// nonceRecordFieldCount is how many whitespace-separated fields a
// record carries: the version label, the consumed-at second, and the
// seconds after it at which the record may be swept.
const nonceRecordFieldCount = 3

// errNonceRecordUnparseable is the sweep's internal signal that a
// record's bytes are not a record. No caller branches on it, so it is
// unexported and never returned: the sweep answers a record it cannot
// read from the file's mtime instead.
var errNonceRecordUnparseable = errors.New("nonce record: unparseable")

// A NonceStore is the set of consumed magic-link nonces, one record per
// nonce under the auth leaf's nonce subdirectory.
//
// The insert is first-writer-wins and neither atomicfile.WriteFile nor
// atomicfile.WriteStream can implement it: both write a temp file and
// rename it over the target, so a second insert of a consumed nonce
// would silently replace the record that proves it was consumed and
// then report success. The insert opens the record with
// O_CREATE|O_EXCL instead, which is the filesystem's own answer to
// first-writer-wins, and syncs both the record and the directory that
// now names it (R11).
//
// ruled, conductor 2026-09-08, R11
type NonceStore struct {
	store *Store

	// dir is the nonce subdirectory. It names a directory rather
	// than a record, so it does not go through the store's record
	// path function; every record path below still does.
	dir string

	// retention reads DUBPLATE_NONCE_RETENTION's value in force at
	// each insert and each sweep: the shortest life a new record is
	// written with. Each record carries its own life, and the sweep
	// prunes on the carried value rather than on this one, so
	// lowering the retention cannot shorten the life of a record
	// already on disk. The value a sweep reads also answers for a
	// record whose own carried value cannot be read.
	retention Setting[time.Duration]

	// linkExpiryCeiling is the registered ceiling on the magic-link
	// expiry, never the value in force: the longest a token minted
	// under any configuration can outlive its own mint. mtimeFallback
	// answers with max(retention, linkExpiryCeiling) rather than
	// retention alone, so a record this store can neither read nor
	// parse cannot be swept, and its nonce replayed, before every
	// token that could have produced it has already expired.
	linkExpiryCeiling time.Duration
}

// A NonceSweep reports one sweep's counts. Malformed counts the records
// the sweep could not parse or could not even read and therefore
// retained, which is a crash or a permissions signal worth logging
// rather than a routine outcome (review m8).
type NonceSweep struct {
	Removed   int
	Malformed int
}

// OpenNonceStore returns the consumed-nonce store under store's leaf,
// with retention reading DUBPLATE_NONCE_RETENTION's value in force and
// linkExpiryCeiling the registered ceiling on the magic-link expiry
// (m5), and sweeps it once at now before returning. The caller runs
// Sweep on an interval thereafter; this package owns no timer.
func OpenNonceStore(ctx context.Context, store *Store, retention Setting[time.Duration], linkExpiryCeiling time.Duration, now time.Time) (*NonceStore, NonceSweep, error) {
	s := &NonceStore{
		store:             store,
		dir:               store.subdir(nonceSubdir),
		retention:         retention,
		linkExpiryCeiling: linkExpiryCeiling,
	}
	sweep, err := s.Sweep(ctx, now)
	if err != nil {
		return nil, sweep, err
	}
	return s, sweep, nil
}

// SweepInterval is the period the serve verb drives Sweep on after
// startup. An hour keeps the store's filesystem cost, one directory
// listing plus a stat per malformed record, far below anything a
// group's own traffic would notice, without adding a second
// configurable interval to a store whose replay guarantee never
// depends on how promptly a swept record disappears.
const SweepInterval = time.Hour

// RunSweep calls sweep, a NonceStore's own Sweep method in production,
// once for every value received on tick, passing it ctx and, as the
// sweep's instant, now(), until ctx is done. The caller owns tick's
// ticker, which is what lets a test drive one sweep without waiting
// SweepInterval out.
//
// A sweep's own error is logged and never returned: after startup, a
// record's presence, not its readability, is what blocks a replay, so
// one failed sweep must not stop the ones after it.
func RunSweep(ctx context.Context, sweep func(ctx context.Context, now time.Time) (NonceSweep, error), now func() time.Time, tick <-chan time.Time, logger *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			result, err := sweep(ctx, now())
			if err != nil {
				logger.Error("nonce sweep failed", "reason", err.Error())
				continue
			}
			logger.Info("nonce sweep completed", "removed", result.Removed, "malformed", result.Malformed)
		}
	}
}

// Insert records nonce as consumed at now and reports ErrNonceReplayed
// when a record for it already exists. expiresAt is the expiry of the
// token that bore the nonce. The caller inserts only after that token
// has verified, so an unauthenticated value can never fill the store.
//
// The record may be swept at the later of now plus the store's
// retention and expiresAt, so it outlives the token that bore its
// nonce whatever retention is configured: a retention shortened while
// a longer-lived link is outstanding, or a clock that steps forward,
// can lengthen a record's life and never shorten it below the token's.
// The retention is read before the record is created, so a failed read,
// reported wrapping errSettingUnreadable, consumes nothing.
func (s *NonceStore) Insert(ctx context.Context, nonce [NonceLength]byte, now, expiresAt time.Time) error {
	retention, err := s.retention.read(ctx)
	if err != nil {
		return err
	}
	path, err := s.store.NoncePath(RecordID(nonceSubdir, string(nonce[:])))
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return ErrNonceReplayed
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

	// The record's name reaches the directory only when the
	// directory itself is synced. Without this, a crash after a
	// consumption can lose the record while the link that produced
	// it is still within its expiry, which is single use lost to a
	// power cut.
	return syncDir(s.dir)
}

// Sweep removes every record whose carried expiry has passed at now and
// returns what it did. A record at exactly its expiry is retained, so
// the boundary belongs to the record.
//
// A record the sweep cannot parse or cannot read is retained and
// counted, not deleted and not reported as a failure: a record's
// presence, not its readability, is what blocks a replay, so one bad
// mode or one bad inode must not take the login surface down (review
// m8). The insert creates the record before writing it, so a crash in
// between leaves a zero-length file whose token may still verify, and
// deleting it on sight would make that token consumable a second time
// for the rest of its expiry. Such a record is removed once its mtime
// has passed max(the store's retention, its link-expiry ceiling), which
// is the one filesystem timestamp this store reads and is read only
// because the record's own expiry is unavailable.
//
// The two failures Sweep still fails closed on are a failed retention
// read, which removes nothing rather than sweep under a default, and
// the directory listing itself: an unreadable store directory is a
// deployment defect this store cannot route around, unlike a single
// unreadable record.
//
// ruled, conductor 2026-09-12, A7 fold 5
func (s *NonceStore) Sweep(ctx context.Context, now time.Time) (NonceSweep, error) {
	retention, err := s.retention.read(ctx)
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
			// Another sweep removed it between the listing
			// and the read.
			continue
		case errors.Is(err, errNonceRecordUnparseable):
			if !now.After(expiresAt) {
				sweep.Malformed++
				continue
			}
		case err != nil:
			// The mtime fallback's own stat failed for a reason
			// other than absence (a removal surfaces as
			// fs.ErrNotExist and the case above already consumes
			// it), most often a permission error on the directory
			// entry. Retained and counted rather than joined into
			// a returned error: a single bad record must not stop
			// OpenNonceStore from starting the login surface at
			// all.
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

// recordExpiry returns when the record at path may be removed, with
// retention the value in force for this sweep. For a record it cannot
// read or cannot parse it returns the mtime fallback alongside
// errNonceRecordUnparseable, so the caller has both the
// reason and the answer: an absent file is reported as its own error
// instead, since the sweep's caller already answers for that case (it
// was removed between the listing and the read).
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

// mtimeFallback answers for a record recordExpiry could not resolve
// from its own bytes, whether because it could not be read (an
// unreadable mode, most often) or because it could not be parsed: both
// take the same mtime-plus-life answer, so an unreadable record ages
// out of the store exactly as an unparseable one already does.
//
// The life added is max(retention, s.linkExpiryCeiling), not
// retention alone (m5): a readable record's own life is at least
// its token's expiry (Insert), and a token's expiry is at most the
// configured link-expiry ceiling, so this bound keeps the fallback at
// least as long-lived as any record the sweep could have read
// instead. Without it, lowering DUBPLATE_NONCE_RETENTION and the link
// expiry together, while a link minted under a longer retention is
// still outstanding, could sweep its unreadable record and let the
// link replay before the token it names has expired.
func (s *NonceStore) mtimeFallback(path string, retention time.Duration) (time.Time, error) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		return time.Time{}, statErr
	}
	return info.ModTime().Add(max(retention, s.linkExpiryCeiling)), errNonceRecordUnparseable
}

// nonceRecord renders a record at second granularity, which is the
// resolution a retention measured in hours needs and the resolution
// every other timestamp this package writes already uses. The record
// carries how long after consumedAt it may be swept, rounded up to a
// whole second so the rounding can only lengthen its life.
func nonceRecord(consumedAt, sweepableAt time.Time) []byte {
	life := sweepableAt.Sub(consumedAt)
	return fmt.Appendf(nil, "%s %d %d\n",
		nonceRecordVersion,
		consumedAt.Unix(),
		int64((life+time.Second-1)/time.Second),
	)
}

// parseNonceRecord returns the record's consumed-at plus its carried
// life, which is the moment the sweep may remove it.
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

// writeAndSync writes data to f, syncs it, and closes it, reporting the
// first failure. It closes f on every path, so a caller that returns
// early leaks nothing.
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

// syncDir fsyncs dir so a rename or a creation within it survives a
// crash.
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open nonce directory: %w", err)
	}
	err = d.Sync()
	if closeErr := d.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("sync nonce directory: %w", err)
	}
	return nil
}
