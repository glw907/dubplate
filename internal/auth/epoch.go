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

// The revocation refusals are distinct from every other credential
// refusal, so a caller logs a revocation as what it is.
var (
	errSessionRevoked = errors.New("session revoked")

	errLinkRevoked = errors.New("magic link revoked")
)

// errEpochBackward reports a bump earlier than the stored epoch, which
// is left in place, so a caller can tell it from a failed write.
var errEpochBackward = errors.New("revocation epoch would move backward")

// epochRecordVersion labels the record layout. A new layout takes a new
// label, so a record is read with the meanings it was written under.
const epochRecordVersion = "dubplate-epoch-1"

const epochRecordFieldCount = 2

// An EpochStore is the set of revocation epochs, one record per
// normalized login address under the auth dir's epoch subdirectory. A
// revocation is stamped on the person, so it is keyed on the address a
// credential carries.
type EpochStore struct {
	authDir *Dir

	// mu serializes each bump's read of the stored epoch with its write.
	mu sync.Mutex
}

// NewEpochStore returns the revocation-epoch store under authDir.
func NewEpochStore(authDir *Dir) *EpochStore {
	return &EpochStore{authDir: authDir}
}

// bump revokes every credential for address anchored or minted at or
// before now. The stored epoch only moves forward: an earlier bump
// writes nothing and returns [errEpochBackward], so a backward clock
// step cannot reinstate a revoked credential.
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

// Revoke bumps address's epoch to at for a roster write, which calls it
// inside its transaction and rolls back when it fails. A bump refused as
// backward is logged and Revoke returns nil: the later stored epoch
// already refuses every credential this bump would. A session anchored
// after that epoch but before the clock stepped back survives, since no
// bump at the stepped-back time can reach it.
func (e *EpochStore) Revoke(ctx context.Context, address string, at time.Time, logger *slog.Logger) error {
	err := e.bump(address, at)
	if errors.Is(err, errEpochBackward) {
		logger.Log(ctx, slog.LevelWarn, "revocation epoch is later than the clock; the write commits on the stored epoch", "address", address)
		return nil
	}
	return err
}

// advance reads and writes under one lock, so of two concurrent bumps
// the later time is the one stored. The write is a rename through
// [atomicfile.WriteStream], not the first-writer-wins create
// [NonceStore] uses.
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
	path, err := e.authDir.epochPath(recordID(epochSubdir, normalized))
	if err != nil {
		return err
	}
	return atomicfile.WriteStream(path, filePerm, func(w io.Writer) error {
		_, err := w.Write(epochRecord(epoch))
		return err
	})
}

// Read returns address's current revocation epoch. ok is false, with a
// nil error, when address has never been revoked.
func (e *EpochStore) Read(address string) (epoch time.Time, ok bool, err error) {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("read epoch: %w", err)
	}
	return e.readNormalized(normalized)
}

func (e *EpochStore) readNormalized(normalized string) (time.Time, bool, error) {
	path, err := e.authDir.epochPath(recordID(epochSubdir, normalized))
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

// checkSession refuses s when its anchor is at or before its address's
// revocation epoch. It reads [Session.AnchoredAt], not IssuedAt: one
// refresh after a bump would carry a later issue time and be permitted
// from then on.
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

// checkLink refuses tok when its mint time is at or before its address's
// revocation epoch. Without it, an outstanding link survives an
// offboarding and mints a session whose fresh anchor passes
// [EpochStore.checkSession]; the clicker chooses when to redeem, so the
// refusal keys on the mint time.
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

func epochRecord(epoch time.Time) []byte {
	return fmt.Appendf(nil, "%s %d\n", epochRecordVersion, epoch.Unix())
}

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
