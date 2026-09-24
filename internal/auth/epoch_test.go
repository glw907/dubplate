package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

// epochClock is the fixture instant every row below anchors its
// timestamps to, unless a row states an offset from it.
var epochClock = time.Date(2026, 4, 5, 6, 7, 8, 0, time.UTC)

const epochFixtureAddress = "member@example.com"

func newEpochStore(t *testing.T, stateDir string) *EpochStore {
	t.Helper()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return OpenEpochStore(store)
}

// TestBumpReadDurable proves the bump is durable rather than an
// in-memory cache: the read runs through a fresh EpochStore opened over
// the same root as the one that bumped.
func TestBumpReadDurable(t *testing.T) {
	stateDir := t.TempDir()
	first := newEpochStore(t, stateDir)

	if err := first.Bump(epochFixtureAddress, epochClock); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	second := newEpochStore(t, stateDir)
	got, ok, err := second.Read(epochFixtureAddress)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !ok {
		t.Fatalf("Read: ok = false, want true")
	}
	if !got.Equal(epochClock.Truncate(time.Second)) {
		t.Errorf("Read = %v, want %v", got, epochClock.Truncate(time.Second))
	}
}

// TestReadNoRecord proves an address with no bump reports no
// revocation rather than an error, which is the ordinary case for
// every member who has never been offboarded.
func TestReadNoRecord(t *testing.T) {
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)

	epoch, ok, err := store.Read(epochFixtureAddress)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if ok {
		t.Fatalf("Read: ok = true, want false")
	}
	if !epoch.IsZero() {
		t.Errorf("Read epoch = %v, want zero", epoch)
	}
}

// TestCheckSession is spec:2168's third exit-gate assertion: a session
// anchored before the bump is refused, one anchored after is accepted,
// and the boundary at exactly the epoch belongs to the refusal.
func TestCheckSession(t *testing.T) {
	tests := []struct {
		name    string
		anchor  time.Time
		wantErr error
	}{
		{"anchored before the bump", epochClock.Add(-time.Minute), ErrSessionRevoked},
		{"anchored exactly at the bump", epochClock, ErrSessionRevoked},
		{"anchored after the bump", epochClock.Add(time.Minute), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			store := newEpochStore(t, stateDir)
			if err := store.Bump(epochFixtureAddress, epochClock); err != nil {
				t.Fatalf("Bump: %v", err)
			}

			s := Session{Address: epochFixtureAddress, AnchoredAt: tt.anchor}
			err := store.CheckSession(s)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("CheckSession: %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckSession error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestCheckSessionRefreshedAfterBumpStillRefused is threat-model claim
// 4: a refresh performed between the bump and the check must not clear
// the refusal, which is what an issue-time check would let happen.
func TestCheckSessionRefreshedAfterBumpStillRefused(t *testing.T) {
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)

	_, session, err := MintSession([]byte("session-secret"), epochFixtureAddress, epochClock.Add(-time.Hour), time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}

	if err := store.Bump(epochFixtureAddress, epochClock); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	_, refreshed := RefreshSession([]byte("session-secret"), session, epochClock.Add(time.Minute), time.Hour, 24*time.Hour)

	if err := store.CheckSession(refreshed); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("CheckSession after refresh = %v, want ErrSessionRevoked", err)
	}
}

// TestCheckLink is the link-side half of spec:2168's threat model: a
// link minted before a bump is refused, asserted through its mint time
// rather than its expiry, because the attacker chooses when to click
// and not when the link expires.
func TestCheckLink(t *testing.T) {
	tests := []struct {
		name     string
		mintedAt time.Time
		wantErr  error
	}{
		{"minted before the bump", epochClock.Add(-time.Minute), ErrLinkRevoked},
		{"minted after the bump", epochClock.Add(time.Minute), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			store := newEpochStore(t, stateDir)
			if err := store.Bump(epochFixtureAddress, epochClock); err != nil {
				t.Fatalf("Bump: %v", err)
			}

			tok := LinkToken{
				Address:   epochFixtureAddress,
				MintedAt:  tt.mintedAt,
				ExpiresAt: tt.mintedAt.Add(time.Hour),
			}
			err := store.CheckLink(tok)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("CheckLink: %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckLink error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestCheckSessionNormalizesAddress proves the epoch check normalizes
// its own key rather than trusting the credential's wire value: a
// session is still refused whichever side of the bump/check pair
// carries the mixed-case address. A check that keyed on the raw
// address, either bumping normalized and checking mixed-case or the
// reverse, would silently fail to revoke it (spec:2370-2376, risk 4).
func TestCheckSessionNormalizesAddress(t *testing.T) {
	tests := []struct {
		name        string
		bumpAddress string
		sessAddress string
	}{
		{"bump mixed-case, check normalized", "Member@Example.com", "member@example.com"},
		{"bump normalized, check mixed-case", "member@example.com", "Member@Example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			store := newEpochStore(t, stateDir)
			if err := store.Bump(tt.bumpAddress, epochClock); err != nil {
				t.Fatalf("Bump: %v", err)
			}

			s := Session{Address: tt.sessAddress, AnchoredAt: epochClock.Add(-time.Minute)}
			if err := store.CheckSession(s); !errors.Is(err, ErrSessionRevoked) {
				t.Fatalf("CheckSession = %v, want ErrSessionRevoked", err)
			}
		})
	}
}

// TestBackwardBumpKeepsTheRevocation proves the epoch never moves
// backward. A bump carrying an earlier time than the stored epoch,
// which is what a backward clock step between two bumps produces, must
// not reinstate the session the first bump revoked, and reports the
// refusal rather than a success.
func TestBackwardBumpKeepsTheRevocation(t *testing.T) {
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)

	_, session, err := MintSession([]byte("session-secret"), epochFixtureAddress, epochClock, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	if err := store.Bump(epochFixtureAddress, epochClock.Add(time.Minute)); err != nil {
		t.Fatalf("Bump forward: %v", err)
	}
	if err := store.CheckSession(session); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("CheckSession after the forward bump = %v, want ErrSessionRevoked", err)
	}

	notified := 0
	store.Subscribe(func(string, time.Time) { notified++ })
	if err := store.Bump(epochFixtureAddress, epochClock.Add(-time.Hour)); !errors.Is(err, ErrEpochBackward) {
		t.Errorf("Bump backward = %v, want ErrEpochBackward", err)
	}
	if notified != 0 {
		t.Errorf("subscriber notified %d times by a bump that moved nothing, want 0", notified)
	}

	if err := store.CheckSession(session); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("CheckSession after the backward bump = %v, want ErrSessionRevoked", err)
	}
	got, _, err := store.Read(epochFixtureAddress)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if want := epochClock.Add(time.Minute); !got.Equal(want) {
		t.Errorf("Read after the backward bump = %v, want %v", got, want)
	}
}

// TestConcurrentBumpsKeepTheLater is what separates a read-modify-write
// held under one lock from a read taken outside it: the sequential row
// above passes either way, because the window between the read and the
// write is reachable only from a second writer. Each round races an
// earlier and a later bump for a fresh address, and across the rounds
// an unlocked read lets the earlier one land last somewhere.
func TestConcurrentBumpsKeepTheLater(t *testing.T) {
	const rounds = 64
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)
	earlier, later := epochClock, epochClock.Add(time.Hour)

	for round := range rounds {
		address := "member" + strconv.Itoa(round) + "@example.com"
		start := make(chan struct{})
		var wg sync.WaitGroup
		for _, at := range []time.Time{earlier, later} {
			wg.Go(func() {
				<-start
				// The earlier bump returns ErrEpochBackward
				// when it lands second; the stored value
				// below is the assertion.
				_ = store.Bump(address, at)
			})
		}
		close(start)
		wg.Wait()

		got, ok, err := store.Read(address)
		if err != nil || !ok {
			t.Fatalf("round %d: Read = %v, %v, %v", round, got, ok, err)
		}
		if !got.Equal(later) {
			t.Fatalf("round %d: stored epoch = %v, want the later bump %v", round, got, later)
		}
	}
}

// TestSubscribeNotifiesOnBump proves the notification half of the seam
// reaches a registered subscriber.
func TestSubscribeNotifiesOnBump(t *testing.T) {
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)

	var gotAddress string
	var gotEpoch time.Time
	calls := 0
	store.Subscribe(func(address string, epoch time.Time) {
		calls++
		gotAddress = address
		gotEpoch = epoch
	})

	if err := store.Bump(epochFixtureAddress, epochClock); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	if calls != 1 {
		t.Fatalf("subscriber calls = %d, want 1", calls)
	}
	if gotAddress != epochFixtureAddress {
		t.Errorf("subscriber address = %q, want %q", gotAddress, epochFixtureAddress)
	}
	if !gotEpoch.Equal(epochClock.Truncate(time.Second)) {
		t.Errorf("subscriber epoch = %v, want %v", gotEpoch, epochClock.Truncate(time.Second))
	}
}

// TestBumpRecoversFromPanickingSubscriber proves the review m14 fix: a
// subscriber that panics does not take Bump down, and the record it
// wrote before running any subscriber is still on disk.
func TestBumpRecoversFromPanickingSubscriber(t *testing.T) {
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)

	calledAfter := false
	store.Subscribe(func(string, time.Time) { panic("subscriber exploded") })
	store.Subscribe(func(string, time.Time) { calledAfter = true })

	if err := store.Bump(epochFixtureAddress, epochClock); err != nil {
		t.Fatalf("Bump: %v, want nil: a panicking subscriber must not fail the bump", err)
	}
	if !calledAfter {
		t.Error("the subscriber registered after the panicking one was never called")
	}

	got, ok, err := store.Read(epochFixtureAddress)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !ok {
		t.Fatal("Read: ok = false, want true: the record must still be on disk")
	}
	if !got.Equal(epochClock.Truncate(time.Second)) {
		t.Errorf("Read = %v, want %v", got, epochClock.Truncate(time.Second))
	}
}

// TestReadAuthoritativeWithoutSubscriber proves the read is the
// authority rather than the notification: a bump with no subscriber
// registered still lands on a fresh store's read. A publish-only event
// could not offer this.
func TestReadAuthoritativeWithoutSubscriber(t *testing.T) {
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)

	if err := store.Bump(epochFixtureAddress, epochClock); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	fresh := newEpochStore(t, stateDir)
	got, ok, err := fresh.Read(epochFixtureAddress)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !ok {
		t.Fatalf("Read: ok = false, want true")
	}
	if !got.Equal(epochClock.Truncate(time.Second)) {
		t.Errorf("Read = %v, want %v", got, epochClock.Truncate(time.Second))
	}
}

// TestBumpNoAddressInFilename re-asserts, for this record, that the raw
// address never appears in a created filename (Global Constraints).
func TestBumpNoAddressInFilename(t *testing.T) {
	stateDir := t.TempDir()
	store := newEpochStore(t, stateDir)

	if err := store.Bump(epochFixtureAddress, epochClock); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(StoreDir(stateDir), epochSubdir))
	if err != nil {
		t.Fatalf("read epoch directory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("epoch directory entries = %d, want 1", len(entries))
	}
	if entries[0].Name() == epochFixtureAddress {
		t.Errorf("filename %s carries the raw address", entries[0].Name())
	}
}
