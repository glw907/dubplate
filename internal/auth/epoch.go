package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
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
	// errSessionRevoked reports a session whose anchor predates its
	// address's revocation epoch.
	errSessionRevoked = errors.New("session revoked")

	// errLinkRevoked reports a magic link whose mint time predates its
	// address's revocation epoch.
	errLinkRevoked = errors.New("magic link revoked")
)

// errEpochBackward reports a bump whose time is earlier than the epoch
// already stored for its address. The stored epoch is left in place,
// so the revocation it records still holds; the sentinel lets a caller
// tell that outcome from a bump that failed to write.
var errEpochBackward = errors.New("revocation epoch would move backward")

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
// A revocation is stamped on the person, so the record is
// keyed on the address, the identity a credential carries.
//
// A bump overwrites its record through [atomicfile.WriteStream], which
// fsyncs the containing directory after the rename, rather than the
// first-writer-wins create [NonceStore] uses.
type EpochStore struct {
	leaf *Dir

	// mu serializes each bump's read of the stored epoch with its write.
	mu sync.Mutex
}

// NewEpochStore returns the revocation-epoch store under leaf.
func NewEpochStore(leaf *Dir) *EpochStore {
	return &EpochStore{leaf: leaf}
}

// bump revokes every credential for address anchored or minted at or
// before now: it writes now as address's new epoch.
//
// The stored epoch only moves forward. A bump earlier than it writes
// nothing and returns [errEpochBackward], so a backward clock step
// cannot reinstate a credential an earlier bump revoked. A bump at the
// stored epoch's own second has nothing to move and returns nil. The
// read of the stored epoch and the write of the new one happen under
// one lock, so of two concurrent bumps the later time is the one left
// stored.
func (e *EpochStore) bump(address string, now time.Time) error {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return fmt.Errorf("bump epoch: %w", err)
	}
	if err := e.advance(normalized, now.UTC().Truncate(time.Second)); err != nil {
		return fmt.Errorf("bump epoch: %w", err)
	}
	return nil
}

// Revoke bumps address's epoch to at for a roster write. The write calls
// it inside its transaction, and a failed bump rolls the write back. A
// bump refused with [errEpochBackward] is logged to logger, and Revoke
// returns nil. The stored epoch is later than at and already refuses
// every credential this bump would. What committing accepts is a
// session anchored between that stored epoch and the clock's step back:
// no bump at the stepped-back time can reach it.
func (e *EpochStore) Revoke(ctx context.Context, address string, at time.Time, logger *slog.Logger) error {
	err := e.bump(address, at)
	if errors.Is(err, errEpochBackward) {
		logger.Log(ctx, slog.LevelWarn, "revocation epoch is later than the clock; the write commits on the stored epoch", "address", address)
		return nil
	}
	return err
}

// advance writes epoch as normalized's epoch when it is later than the
// stored one. It writes nothing when the stored epoch already equals
// epoch, and returns [errEpochBackward] when the stored epoch is later.
func (e *EpochStore) advance(normalized string, epoch time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	stored, ok, err := e.readNormalized(normalized)
	if err != nil {
		return err
	}
	if ok && epoch.Before(stored) {
		return errEpochBackward
	}
	if ok && epoch.Equal(stored) {
		return nil
	}
	path, err := e.leaf.epochPath(recordID(epochSubdir, normalized))
	if err != nil {
		return err
	}
	return atomicfile.WriteStream(path, filePerm, func(w io.Writer) error {
		_, err := w.Write(epochRecord(epoch))
		return err
	})
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
	path, err := e.leaf.epochPath(recordID(epochSubdir, normalized))
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

// checkSession refuses s when its anchor predates or equals its
// address's revocation epoch.
//
// The comparison reads [Session.AnchoredAt] rather than
// [Session.IssuedAt]: a refresh advances the issue time but
// never the anchor, so checking the issue time would let a session
// refreshed once after a bump carry a post-bump issue time and be
// permitted from then on, permanently. The anchor is what refresh never
// moves, so no ordering rule between this check and a refresh is
// load-bearing.
func (e *EpochStore) checkSession(s Session) error {
	revoked, err := e.revokedAt(s.Address, s.AnchoredAt)
	if err != nil {
		return fmt.Errorf("check session epoch: %w", err)
	}
	if revoked {
		return errSessionRevoked
	}
	return nil
}

// checkLink refuses tok when its mint time predates or equals its
// address's revocation epoch.
//
// Without this check, an outstanding magic link survives an offboarding
// for the whole of its expiry and mints a session [EpochStore.checkSession]
// then honors, because the session's anchor is set at that later mint.
// The attacker chooses when to click, so the refusal must be on the
// link's own mint time, not on when it happens to be redeemed.
func (e *EpochStore) checkLink(tok LinkToken) error {
	revoked, err := e.revokedAt(tok.Address, tok.MintedAt)
	if err != nil {
		return fmt.Errorf("check link epoch: %w", err)
	}
	if revoked {
		return errLinkRevoked
	}
	return nil
}

// revokedAt reports whether a credential for address stamped at t falls
// at or before address's revocation epoch.
func (e *EpochStore) revokedAt(address string, t time.Time) (bool, error) {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return false, err
	}
	epoch, ok, err := e.readNormalized(normalized)
	if err != nil {
		return false, err
	}
	return ok && !t.After(epoch), nil
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
