package auth

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/store"
)

// revokeFixture is a roster over a fresh store whose revoker is a real
// epoch store, with the verifier that reads the same epochs.
type revokeFixture struct {
	roster *member.Roster
	h      *store.Handle
	leaf   *Dir
	epoch  *EpochStore
	clock  *revokeClock
}

// revokeClock is a clock a test moves by hand while the roster reads it.
type revokeClock struct {
	mu sync.Mutex
	at time.Time
}

func (c *revokeClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at
}

func (c *revokeClock) set(at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = at
}

// operatorActor is the command-line actor every revocation fixture
// writes as.
var operatorActor = identity.LocalActor("operator")

// newRevokeFixture returns a revokeFixture holding a live admin and
// alice, a live member, with its clock at rosterNow.
func newRevokeFixture(t *testing.T) revokeFixture {
	t.Helper()
	fx := revokeFixture{h: newUsersStore(t), clock: &revokeClock{at: rosterNow}}
	insertUsers(t, fx.h,
		userRow{id: adminID, address: "admin@example.com", displayName: "Admin", username: "groupadmin", admin: true, createdAt: 100},
		userRow{id: aliceID, address: "alice@example.com", displayName: "Alice", username: "alice", createdAt: 110},
	)
	leaf, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	fx.leaf = leaf
	fx.epoch = OpenEpochStore(leaf)
	fx.roster = member.NewRoster(fx.h, fx.epoch, fx.clock.now)
	return fx
}

// verifyAliceAt verifies a session minted for alice at minted, checked at
// minted plus a minute, through a verifier over fx's roster and epochs.
func (fx revokeFixture) verifyAliceAt(t *testing.T, minted time.Time) (VerifiedIdentity, error) {
	t.Helper()
	value, _, err := MintSession([]byte(currentFixture), "alice@example.com", minted, verifierLifetime, verifierCap)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	seam := NewMagicLinkVerifier(fx.roster, fx.epoch, []byte(currentFixture), []byte(previousFixture), fixed(verifierLifetime), fixed(verifierCap))
	return seam.VerifyIdentity(requestWithSession(t, value), minted.Add(time.Minute), false)
}

// tombstoned reports whether address's users row is tombstoned.
func (fx revokeFixture) tombstoned(t *testing.T, address string) bool {
	t.Helper()
	var gone bool
	if err := fx.h.QueryRow(t.Context(), `SELECT tombstoned_at IS NOT NULL FROM users WHERE address = ?`, address).Scan(&gone); err != nil {
		t.Fatalf("read %s: %v", address, err)
	}
	return gone
}

// TestRosterRemoval_RevokesSessions proves a roster removal ends the
// member's sessions through the epoch store: a session minted before the
// removal is refused after it, and stays refused after the address is
// re-added under its old id, even with the removal's epoch record
// deleted first, because the re-add revokes too. A sign-in one second
// past the re-add verifies. Fixture state: a live member whose session
// verifies before the removal.
func TestRosterRemoval_RevokesSessions(t *testing.T) {
	fx := newRevokeFixture(t)
	minted := rosterNow.Add(-time.Hour)
	if _, err := fx.verifyAliceAt(t, minted); err != nil {
		t.Fatalf("the pre-removal session does not verify before the removal: %v", err)
	}

	if err := fx.roster.RemoveMember(writeCtx(t), aliceID, operatorActor, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	if _, err := fx.verifyAliceAt(t, minted); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("pre-removal session after removal: %v, want ErrSessionRevoked", err)
	}

	path, err := fx.leaf.EpochPath(RecordID(epochSubdir, "alice@example.com"))
	if err != nil {
		t.Fatalf("EpochPath: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("delete the removal's epoch record: %v", err)
	}

	readdedAt := rosterNow.Add(time.Hour)
	fx.clock.set(readdedAt)
	readd := member.Member{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"}
	if _, err := fx.roster.AddMember(writeCtx(t), readd, operatorActor, slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("re-add: %v", err)
	}
	if _, err := fx.verifyAliceAt(t, minted); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("pre-removal session after re-add: %v, want ErrSessionRevoked", err)
	}
	if _, err := fx.verifyAliceAt(t, readdedAt.Add(-time.Minute)); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("session minted between the removal and the re-add: %v, want ErrSessionRevoked", err)
	}

	got, err := fx.verifyAliceAt(t, readdedAt.Add(time.Second))
	if err != nil {
		t.Fatalf("sign-in one second past the re-add's revocation: %v, want it to verify", err)
	}
	if got.Identity.UserID != aliceID {
		t.Errorf("signed-in id = %x, want %x", got.Identity.UserID, aliceID)
	}
}

// TestEpochStoreRevoke holds Revoke's failure policy: a revocation
// bumps the epoch to the time it names; one refused with
// ErrEpochBackward, because the stored epoch is later, logs a warning,
// returns nil, and leaves the later epoch in place; and any other bump
// failure comes back. Fixture states: no stored epoch, an epoch five
// hours past the revocation, and an epoch subdirectory no bump can
// write.
func TestEpochStoreRevoke(t *testing.T) {
	later := rosterNow.Add(5 * time.Hour)
	tests := []struct {
		name      string
		prepare   func(t *testing.T, leaf *Dir, epoch *EpochStore)
		wantErr   bool
		wantEpoch time.Time
		wantWarn  bool
	}{
		{name: "no stored epoch", prepare: func(*testing.T, *Dir, *EpochStore) {}, wantEpoch: rosterNow},
		{
			name: "a later stored epoch",
			prepare: func(t *testing.T, _ *Dir, epoch *EpochStore) {
				if err := epoch.Bump("alice@example.com", later); err != nil {
					t.Fatalf("Bump the later epoch: %v", err)
				}
			},
			wantEpoch: later,
			wantWarn:  true,
		},
		{
			name: "an unwritable epoch subdirectory",
			prepare: func(t *testing.T, leaf *Dir, _ *EpochStore) {
				dir := leaf.subdir(epochSubdir)
				if err := os.Chmod(dir, 0o500); err != nil {
					t.Fatalf("chmod the epoch subdirectory: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, dirPerm) })
				if probe, err := os.CreateTemp(dir, "probe"); err == nil {
					_ = probe.Close()
					t.Fatal("the epoch subdirectory still takes a file; the fixture cannot make a bump fail")
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leaf, err := Open(t.TempDir())
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			epoch := OpenEpochStore(leaf)
			tt.prepare(t, leaf, epoch)

			var logBuf bytes.Buffer
			err = epoch.Revoke(t.Context(), "alice@example.com", rosterNow, slog.New(slog.NewJSONHandler(&logBuf, nil)))
			if tt.wantErr {
				if err == nil || errors.Is(err, ErrEpochBackward) {
					t.Fatalf("Revoke = %v, want the bump's error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Revoke: %v", err)
			}
			if got, ok, err := epoch.Read("alice@example.com"); err != nil || !ok || !got.Equal(tt.wantEpoch) {
				t.Errorf("epoch = %v (present %v, err %v), want %v", got, ok, err, tt.wantEpoch)
			}
			warned := strings.Contains(logBuf.String(), `"level":"WARN"`) && strings.Contains(logBuf.String(), "later than the clock")
			if warned != tt.wantWarn {
				t.Errorf("log = %q, want a warning naming the later epoch %v", logBuf.String(), tt.wantWarn)
			}
		})
	}
}

// TestRosterWrites_EpochBackwardCommits proves a removal and a re-add
// commit when the stored epoch is later than the roster's clock, and
// leave the later epoch in place. Fixture state: alice's epoch bumped
// five hours past the clock.
func TestRosterWrites_EpochBackwardCommits(t *testing.T) {
	later := rosterNow.Add(5 * time.Hour)
	fx := newRevokeFixture(t)
	if err := fx.epoch.Bump("alice@example.com", later); err != nil {
		t.Fatalf("Bump the later epoch: %v", err)
	}
	logger := slog.New(slog.DiscardHandler)

	if err := fx.roster.RemoveMember(writeCtx(t), aliceID, operatorActor, logger); err != nil {
		t.Fatalf("removal against a later epoch = %v, want nil", err)
	}
	if !fx.tombstoned(t, "alice@example.com") {
		t.Error("the removal against a later epoch did not commit")
	}
	readd := member.Member{Address: "alice@example.com", DisplayName: "Alice", Username: "alice"}
	if _, err := fx.roster.AddMember(writeCtx(t), readd, operatorActor, logger); err != nil {
		t.Fatalf("re-add against a later epoch = %v, want nil", err)
	}
	if fx.tombstoned(t, "alice@example.com") {
		t.Error("the re-add against a later epoch did not commit")
	}
	if got, ok, err := fx.epoch.Read("alice@example.com"); err != nil || !ok || !got.Equal(later) {
		t.Errorf("epoch = %v (present %v, err %v), want the later one %v left in place", got, ok, err, later)
	}
}
