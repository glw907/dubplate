package auth

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/identity"
)

// Sentinels the two refusal checks return. Distinct from every other
// credential refusal in this package, so a caller can log a revocation
// as what it is rather than as an expiry or a signature failure.
var (
	// ErrSessionRevoked reports a session whose anchor predates its
	// address's revocation epoch.
	ErrSessionRevoked = errors.New("auth: session revoked")

	// ErrLinkRevoked reports a magic link whose mint time predates its
	// address's revocation epoch.
	ErrLinkRevoked = errors.New("auth: magic link revoked")
)

// ErrEpochBackward reports a bump whose time is earlier than the epoch
// already stored for its address. The stored epoch is left in place,
// so the revocation it records still holds; the sentinel lets a caller
// tell that outcome from a bump that failed to write.
var ErrEpochBackward = errors.New("auth: revocation epoch would move backward")

// epochRecordVersion labels the record's field layout. A later field
// lands under a new label rather than changing what an existing field
// means, so a record written before the change is read with the
// meanings it was written under.
const epochRecordVersion = "dubplate-epoch-1"

// epochRecordFieldCount is how many whitespace-separated fields a
// record carries: the version label and the epoch second.
const epochRecordFieldCount = 2

// An EpochStore is the set of revocation epochs, one record per
// normalized login address under the auth leaf's epoch subdirectory.
//
// The address is the record's deliberate key (R4): every other record
// this package writes is keyed on a value that outlives the person it
// names, but a revocation is stamped on the person, and the address is
// the only identity this rung has for one. Rung 6's re-key onto a
// surrogate id moves this key alongside the identity that owns it
// (R34); until then, the address is what Bump, Read, and both refusal
// checks digest through [RecordID].
//
// A bump overwrites its record, which is why it goes through
// [atomicfile.WriteStream] rather than the first-writer-wins shape
// [NonceStore] uses: overwriting is what a revocation does, and
// WriteStream is the one entry point that fsyncs the containing
// directory after the rename (atomicfile.go:82-89, R11).
//
// ruled, conductor 2026-09-12, A8 fold 1
type EpochStore struct {
	store *Store

	// mu serializes each bump's read of the stored epoch with its
	// write, and guards subscribers.
	mu          sync.Mutex
	subscribers []func(address string, epoch time.Time)
}

// OpenEpochStore returns the revocation-epoch store under store's leaf.
func OpenEpochStore(store *Store) *EpochStore {
	return &EpochStore{store: store}
}

// Bump revokes every credential for address anchored or minted at or
// before now: it writes now as address's new epoch and notifies every
// subscriber registered at the time of the call.
//
// The stored epoch only moves forward. A bump earlier than it writes
// nothing, notifies no one, and returns [ErrEpochBackward], so a
// backward clock step cannot reinstate a credential an earlier bump
// revoked. A bump at the stored epoch's own second has nothing to move
// and returns nil without a notification. The read of the stored epoch
// and the write of the new one happen under one lock, so of two
// concurrent bumps the later time is the one left stored.
//
// The write happens before the notification, so a subscriber that
// receives one is guaranteed to find it on a read; a subscriber that
// misses the notification entirely, because it was not yet registered
// or was not running, still finds the bump on its next read, which is
// the property a publish-only event cannot offer.
func (e *EpochStore) Bump(address string, now time.Time) error {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return fmt.Errorf("bump epoch: %w", err)
	}
	epoch := now.UTC().Truncate(time.Second)
	subscribers, err := e.advance(normalized, epoch)
	if err != nil {
		return fmt.Errorf("bump epoch: %w", err)
	}
	for _, fn := range subscribers {
		fn(normalized, epoch)
	}
	return nil
}

// advance writes epoch as normalized's epoch when it is later than the
// stored one, and returns the subscribers to notify of the move. It
// returns no subscribers when the stored epoch already equals epoch,
// and [ErrEpochBackward] when the stored epoch is later.
func (e *EpochStore) advance(normalized string, epoch time.Time) ([]func(address string, epoch time.Time), error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	stored, ok, err := e.readNormalized(normalized)
	if err != nil {
		return nil, err
	}
	if ok && epoch.Before(stored) {
		return nil, ErrEpochBackward
	}
	if ok && epoch.Equal(stored) {
		return nil, nil
	}
	path, err := e.store.EpochPath(RecordID(normalized))
	if err != nil {
		return nil, err
	}
	if err := atomicfile.WriteStream(path, filePerm, func(w io.Writer) error {
		_, err := w.Write(epochRecord(epoch))
		return err
	}); err != nil {
		return nil, err
	}
	return slices.Clone(e.subscribers), nil
}

// Read returns address's current revocation epoch. ok is false when
// address carries no record, which is the ordinary case for a member
// who has never been offboarded, and is reported as such rather than as
// an error.
func (e *EpochStore) Read(address string) (epoch time.Time, ok bool, err error) {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("read epoch: %w", err)
	}
	return e.readNormalized(normalized)
}

func (e *EpochStore) readNormalized(normalized string) (time.Time, bool, error) {
	path, err := e.store.EpochPath(RecordID(normalized))
	if err != nil {
		return time.Time{}, false, fmt.Errorf("read epoch: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, fmt.Errorf("read epoch: %w", err)
	}
	epoch, err := parseEpochRecord(data)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("read epoch: %w", err)
	}
	return epoch, true, nil
}

// Subscribe registers fn to be called, with the normalized address and
// the new epoch, on every future [EpochStore.Bump]. It is the
// convenience half of the seam: the authority is [EpochStore.Read],
// which a subscriber that missed a notification, or never registered
// at all, can still consult.
func (e *EpochStore) Subscribe(fn func(address string, epoch time.Time)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subscribers = append(e.subscribers, fn)
}

// CheckSession refuses s when its anchor predates or equals its
// address's revocation epoch.
//
// The comparison reads [Session.AnchoredAt] rather than
// [Session.IssuedAt]: a refresh advances the issue time (Task 6) but
// never the anchor, so checking the issue time would let a session
// refreshed once after a bump carry a post-bump issue time and be
// permitted from then on, permanently. The anchor is what refresh never
// moves, so no ordering rule between this check and a refresh is
// load-bearing.
//
// ruled, conductor 2026-09-12, A8 fold 1
func (e *EpochStore) CheckSession(s Session) error {
	normalized, err := identity.NormalizeAddress(s.Address)
	if err != nil {
		return fmt.Errorf("check session epoch: %w", err)
	}
	epoch, ok, err := e.readNormalized(normalized)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !s.AnchoredAt.After(epoch) {
		return ErrSessionRevoked
	}
	return nil
}

// CheckLink refuses tok when its mint time predates or equals its
// address's revocation epoch.
//
// Without this check, an outstanding magic link survives an offboarding
// for the whole of its expiry and mints a session [EpochStore.CheckSession]
// then honors, because the session's anchor is set at that later mint.
// The attacker chooses when to click, so the refusal must be on the
// link's own mint time, not on when it happens to be redeemed.
//
// ruled, conductor 2026-09-12, A8 fold 1
func (e *EpochStore) CheckLink(tok LinkToken) error {
	normalized, err := identity.NormalizeAddress(tok.Address)
	if err != nil {
		return fmt.Errorf("check link epoch: %w", err)
	}
	epoch, ok, err := e.readNormalized(normalized)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !tok.MintedAt.After(epoch) {
		return ErrLinkRevoked
	}
	return nil
}

// epochRecord renders a record at second granularity, the resolution
// every other timestamp this package writes already uses.
func epochRecord(epoch time.Time) []byte {
	return fmt.Appendf(nil, "%s %d\n", epochRecordVersion, epoch.Unix())
}

// parseEpochRecord returns the epoch a record carries.
func parseEpochRecord(data []byte) (time.Time, error) {
	fields := strings.Fields(string(data))
	if len(fields) != epochRecordFieldCount || fields[0] != epochRecordVersion {
		return time.Time{}, fmt.Errorf("epoch record: unparseable")
	}
	seconds, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("epoch record: unparseable: %w", err)
	}
	return time.Unix(seconds, 0).UTC(), nil
}
