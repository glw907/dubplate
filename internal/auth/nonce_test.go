package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// The fixtures below pin everything the nonce rows depend on. The nonce
// values are visibly synthetic byte runs rather than random ones, so no
// literal here reads as a credential.
var (
	nonceClock     = time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	consumedSample = [NonceLength]byte{1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 121, 98, 219}
	secondSample   = [NonceLength]byte{2, 2, 4, 6, 10, 16, 26, 42, 68, 110, 178, 32, 210, 242, 196, 182}

	// sampleExpiry stands for the expiry of the token that bore a
	// sample nonce. It is shorter than fixtureRetention, so every row
	// that inserts a sample reads the retention as the record's life.
	sampleExpiry = nonceClock.Add(fixtureExpiry)
)

// fixtureRetention is longer than fixtureExpiry, which is the relation
// config load enforces (R31) and the relation every row here assumes
// unless it says otherwise.
const fixtureRetention = 24 * time.Hour

// fixtureLinkExpiryCeiling stands for the configured ceiling on the
// magic-link expiry, passed into OpenNonceStore by its caller. It is
// shorter than fixtureRetention, so every row that opens a store with
// it reads the retention, not the ceiling, as the mtime fallback's
// life, unless the row says otherwise.
const fixtureLinkExpiryCeiling = time.Hour

func newNonceStore(t *testing.T, stateDir string, retention time.Duration, now time.Time) *NonceStore {
	t.Helper()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	nonces, _, err := OpenNonceStore(t.Context(), store, fixed(retention), fixtureLinkExpiryCeiling, now)
	if err != nil {
		t.Fatalf("OpenNonceStore: %v", err)
	}
	return nonces
}

func nonceDirEntries(t *testing.T, stateDir string) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(StoreDir(stateDir), nonceSubdir))
	if err != nil {
		t.Fatalf("read nonce directory: %v", err)
	}
	return entries
}

func recordPathFor(t *testing.T, stateDir string, nonce [NonceLength]byte) string {
	t.Helper()
	return filepath.Join(StoreDir(stateDir), nonceSubdir, RecordID(nonceSubdir, string(nonce[:])))
}

// TestInsertRefusesReplay is the create-exclusive insert's behavioural
// assertion: the second insert of a consumed nonce reports the replay
// sentinel and leaves the first record's bytes untouched. An insert
// that renamed a fresh file over the record would pass the sentinel
// half of nothing and fail the byte-identity half.
func TestInsertRefusesReplay(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)

	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("first Insert: %v", err)
	}
	path := recordPathFor(t, stateDir, consumedSample)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}

	err = nonces.Insert(t.Context(), consumedSample, nonceClock.Add(time.Minute), sampleExpiry)
	if !errors.Is(err, ErrNonceReplayed) {
		t.Fatalf("second Insert error = %v, want ErrNonceReplayed", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("re-read record: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("record changed across a refused insert: %q then %q", before, after)
	}
}

// TestInsertRefusesReplayAcrossRestart is spec:2168's first exit-gate
// assertion: the refusal is a property of the record on disk, not of a
// process's memory, so a restarted server refuses the replay a running
// one refused.
func TestInsertRefusesReplayAcrossRestart(t *testing.T) {
	stateDir := t.TempDir()

	first := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := first.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert before restart: %v", err)
	}

	restarted := newNonceStore(t, stateDir, fixtureRetention, nonceClock.Add(time.Minute))
	err := restarted.Insert(t.Context(), consumedSample, nonceClock.Add(time.Minute), sampleExpiry)
	if !errors.Is(err, ErrNonceReplayed) {
		t.Fatalf("Insert after restart error = %v, want ErrNonceReplayed", err)
	}
}

// TestFailedMACNeverReachesNonceStore is spec:2168's second exit-gate
// assertion and risk 3's shape: the consumption order is the MAC first
// and the insert only after it, so a forged token cannot fill the
// store. The order is proved by counting the directory rather than by
// reading VerifyLink's statements.
func TestFailedMACNeverReachesNonceStore(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)

	current := []byte(currentFixture)
	value, _, err := MintLink(current, "member@example.com", nonceClock, fixtureExpiry)
	if err != nil {
		t.Fatalf("MintLink: %v", err)
	}
	forged := flipFieldBit(t, value, linkFieldCount)

	// The consumption path this task can reach, in 5b-ii's own
	// order: verify, then insert what verification returned.
	tok, err := VerifyLink(current, nil, forged, nonceClock)
	if !errors.Is(err, ErrLinkMAC) {
		t.Fatalf("VerifyLink error = %v, want ErrLinkMAC", err)
	}
	// This branch never runs: the Fatalf above has already ended the
	// test on the only error VerifyLink can return here. It stands as
	// the compiled statement of the order, so a change to either
	// signature breaks this test rather than the prose.
	if err == nil {
		if err := nonces.Insert(t.Context(), tok.Nonce, nonceClock, tok.ExpiresAt); err != nil {
			t.Fatalf("Insert: %v", err)
		}
	}

	if entries := nonceDirEntries(t, stateDir); len(entries) != 0 {
		t.Errorf("nonce directory holds %d records after a failed MAC, want 0", len(entries))
	}
}

// TestSweepPrunesOnCarriedExpiry covers the three positions around a
// record's carried expiry. The boundary belongs to the record: at
// exactly its expiry it is retained.
func TestSweepPrunesOnCarriedExpiry(t *testing.T) {
	tests := []struct {
		name        string
		sweepAt     time.Time
		wantRemoved int
		wantPresent bool
	}{
		{"before expiry", nonceClock.Add(fixtureRetention - time.Second), 0, true},
		{"at expiry", nonceClock.Add(fixtureRetention), 0, true},
		{"past expiry", nonceClock.Add(fixtureRetention + time.Second), 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
			if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
				t.Fatalf("Insert: %v", err)
			}

			sweep, err := nonces.Sweep(t.Context(), tt.sweepAt)
			if err != nil {
				t.Fatalf("Sweep: %v", err)
			}
			if sweep.Removed != tt.wantRemoved {
				t.Errorf("Removed = %d, want %d", sweep.Removed, tt.wantRemoved)
			}
			if sweep.Malformed != 0 {
				t.Errorf("Malformed = %d, want 0", sweep.Malformed)
			}

			_, statErr := os.Stat(recordPathFor(t, stateDir, consumedSample))
			if got := statErr == nil; got != tt.wantPresent {
				t.Errorf("record present = %v, want %v (stat: %v)", got, tt.wantPresent, statErr)
			}
		})
	}
}

// TestRecordCarriesTheLongerRetention is TestRecordOutlivesItsToken's
// other side. The token expires fifteen minutes after its mint; the
// record is still there a day later, because a record lives for the
// longer of the retention and the token's expiry.
func TestRecordCarriesTheLongerRetention(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	sweepAt := nonceClock.Add(fixtureExpiry).Add(time.Hour)
	if _, err := nonces.Sweep(t.Context(), sweepAt); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if _, err := os.Stat(recordPathFor(t, stateDir, consumedSample)); err != nil {
		t.Fatalf("record removed %v after mint, before its retention elapsed: %v", sweepAt.Sub(nonceClock), err)
	}

	// A sweep past the retention does remove it, so the survival
	// above is the carried expiry holding rather than a sweep that
	// never removes anything.
	if _, err := nonces.Sweep(t.Context(), nonceClock.Add(fixtureRetention+time.Second)); err != nil {
		t.Fatalf("Sweep past retention: %v", err)
	}
	if _, err := os.Stat(recordPathFor(t, stateDir, consumedSample)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stat after retention = %v, want not-exist", err)
	}
}

// TestRecordOutlivesItsToken proves a record lives at least as long as
// the token that bore its nonce. The store's retention is one second, far below the link's expiry, as
// it would be after an operator shortened the retention while a link
// minted under a longer one was still outstanding. A sweep past the
// retention must leave the record in place, because the token bearing
// its nonce still verifies, and the replay the store exists to refuse
// must still be refused.
func TestRecordOutlivesItsToken(t *testing.T) {
	const shortRetention = time.Second
	if shortRetention >= fixtureExpiry {
		t.Fatalf("fixture retention %v is not below the link expiry %v", shortRetention, fixtureExpiry)
	}

	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, shortRetention, nonceClock)

	current := []byte(currentFixture)
	value, tok, err := MintLink(current, "member@example.com", nonceClock, fixtureExpiry)
	if err != nil {
		t.Fatalf("MintLink: %v", err)
	}
	if err := nonces.Insert(t.Context(), tok.Nonce, nonceClock, tok.ExpiresAt); err != nil {
		t.Fatalf("first Insert: %v", err)
	}
	if err := nonces.Insert(t.Context(), tok.Nonce, nonceClock, tok.ExpiresAt); !errors.Is(err, ErrNonceReplayed) {
		t.Fatalf("second Insert = %v, want ErrNonceReplayed", err)
	}

	sweepAt := nonceClock.Add(2 * shortRetention)
	sweep, err := nonces.Sweep(t.Context(), sweepAt)
	if err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if sweep.Removed != 0 {
		t.Errorf("Removed = %d at %v past the mint, want 0 while the token is live", sweep.Removed, sweepAt.Sub(nonceClock))
	}
	if _, err := VerifyLink(current, nil, value, sweepAt); err != nil {
		t.Fatalf("VerifyLink at %v: %v, want a live token", sweepAt, err)
	}
	if err := nonces.Insert(t.Context(), tok.Nonce, sweepAt, tok.ExpiresAt); !errors.Is(err, ErrNonceReplayed) {
		t.Fatalf("replay Insert after the sweep = %v, want ErrNonceReplayed", err)
	}

	// Past the token's expiry the record is collectable, so the
	// survival above is the token's expiry holding rather than a
	// sweep that never removes anything.
	sweep, err = nonces.Sweep(t.Context(), tok.ExpiresAt.Add(time.Second))
	if err != nil {
		t.Fatalf("Sweep past the token's expiry: %v", err)
	}
	if sweep.Removed != 1 {
		t.Errorf("Removed past the token's expiry = %d, want 1", sweep.Removed)
	}
}

// TestSweepRetainsMalformedRecord proves a record the sweep cannot
// parse is retained and counted while the sweep finishes its listing:
// an unreadable record neither disappears nor stops the sweep.
func TestSweepRetainsMalformedRecord(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	garbage := recordPathFor(t, stateDir, secondSample)
	if err := os.WriteFile(garbage, []byte("not a record at all\n"), 0o600); err != nil {
		t.Fatalf("write malformed record: %v", err)
	}

	sweep, err := nonces.Sweep(t.Context(), nonceClock.Add(fixtureRetention+time.Second))
	if err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if sweep.Removed != 1 {
		t.Errorf("Removed = %d, want 1 (the parseable, expired record)", sweep.Removed)
	}
	if sweep.Malformed != 1 {
		t.Errorf("Malformed = %d, want 1", sweep.Malformed)
	}
	if _, err := os.Stat(garbage); err != nil {
		t.Errorf("malformed record was not retained: %v", err)
	}
}

// TestSweepRetainsUnreadableRecord proves the review m8 fix: a record
// this process cannot even read is retained and counted alongside a
// malformed one, and neither the sweep nor OpenNonceStore reports an
// error for it. A record's presence, not its readability, is what
// blocks a replay, so one bad mode must not take the login surface
// down. It also proves the unreadable record still refuses a replay of
// its own nonce while retained, and that a later sweep, run past the
// file's own mtime plus retention, removes it the same way an
// unparseable record already ages out (the mutation this row guards:
// skipping the mtime fallback on a read error would keep the record
// forever). The chmod this test relies on has no effect for root, which
// bypasses file mode entirely, so the test skips there and says so.
func TestSweepRetainsUnreadableRecord(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: chmod 0o000 does not block a root read")
	}

	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	unreadable := recordPathFor(t, stateDir, secondSample)
	if err := os.WriteFile(unreadable, []byte("irrelevant"), 0o600); err != nil {
		t.Fatalf("write unreadable record: %v", err)
	}
	if err := os.Chmod(unreadable, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o600) })

	sweep, err := nonces.Sweep(t.Context(), nonceClock.Add(fixtureRetention+time.Second))
	if err != nil {
		t.Fatalf("Sweep: %v, want nil: an unreadable record must not fail the sweep", err)
	}
	if sweep.Removed != 1 {
		t.Errorf("Removed = %d, want 1 (the parseable, expired record)", sweep.Removed)
	}
	if sweep.Malformed != 1 {
		t.Errorf("Malformed = %d, want 1 (the unreadable record)", sweep.Malformed)
	}
	if _, err := os.Stat(unreadable); err != nil {
		t.Errorf("unreadable record was not retained: %v", err)
	}

	// The record's own nonce still refuses a replay while it is
	// retained: O_CREATE|O_EXCL detects the existing file without
	// reading it, so the insert never needed the mode that blocks
	// os.ReadFile.
	if err := nonces.Insert(t.Context(), secondSample, nonceClock, sampleExpiry); !errors.Is(err, ErrNonceReplayed) {
		t.Errorf("Insert() error = %v, want ErrNonceReplayed while the record is retained", err)
	}

	// OpenNonceStore runs a sweep at open, so the same store carrying
	// the same unreadable record must still start.
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, err := OpenNonceStore(t.Context(), store, fixed(fixtureRetention), fixtureLinkExpiryCeiling, nonceClock.Add(fixtureRetention+time.Second)); err != nil {
		t.Errorf("OpenNonceStore: %v, want nil: a single unreadable record must not stop startup", err)
	}

	// A later sweep run past the file's own (real) mtime plus
	// retention removes it: the fallback this file takes for being
	// unreadable is the same mtime-plus-retention answer an
	// unparseable record already takes.
	info, err := os.Stat(unreadable)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	past := info.ModTime().Add(fixtureRetention + time.Second)
	sweep, err = nonces.Sweep(t.Context(), past)
	if err != nil {
		t.Fatalf("Sweep: %v, want nil", err)
	}
	if sweep.Removed != 1 {
		t.Errorf("Removed = %d, want 1: the unreadable record past its mtime plus retention", sweep.Removed)
	}
	if _, err := os.Stat(unreadable); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Stat(unreadable) error = %v, want ErrNotExist: the record should have been swept", err)
	}
}

// TestMtimeFallbackRespectsLinkExpiryCeiling proves m5: when the
// configured retention is lower than the link-expiry ceiling, an
// unreadable record survives past mtime+retention and is swept only
// once it has aged past mtime+ceiling. Without the ceiling in the
// fallback, an operator who lowers the retention below the ceiling
// while a longer-lived link is still outstanding could have an
// unreadable record for that link swept, and its nonce replayed,
// before every token that could have produced it has expired.
func TestMtimeFallbackRespectsLinkExpiryCeiling(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: chmod 0o000 does not block a root read")
	}

	const shortRetention = 30 * time.Minute // below fixtureLinkExpiryCeiling (1h)

	stateDir := t.TempDir()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	nonces, _, err := OpenNonceStore(t.Context(), store, fixed(shortRetention), fixtureLinkExpiryCeiling, nonceClock)
	if err != nil {
		t.Fatalf("OpenNonceStore: %v", err)
	}

	unreadable := recordPathFor(t, stateDir, consumedSample)
	if err := os.WriteFile(unreadable, []byte("irrelevant"), 0o600); err != nil {
		t.Fatalf("write unreadable record: %v", err)
	}
	if err := os.Chmod(unreadable, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o600) })

	info, err := os.Stat(unreadable)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	pastRetention := info.ModTime().Add(shortRetention + time.Second)
	sweep, err := nonces.Sweep(t.Context(), pastRetention)
	if err != nil {
		t.Fatalf("Sweep: %v, want nil", err)
	}
	if sweep.Removed != 0 || sweep.Malformed != 1 {
		t.Fatalf("sweep past retention alone = %+v, want {Removed:0 Malformed:1}: the ceiling must hold the record", sweep)
	}
	if _, err := os.Stat(unreadable); err != nil {
		t.Errorf("record swept before the ceiling: %v", err)
	}

	pastCeiling := info.ModTime().Add(fixtureLinkExpiryCeiling + time.Second)
	sweep, err = nonces.Sweep(t.Context(), pastCeiling)
	if err != nil {
		t.Fatalf("Sweep: %v, want nil", err)
	}
	if sweep.Removed != 1 {
		t.Errorf("sweep past the ceiling = %+v, want Removed:1", sweep)
	}
	if _, err := os.Stat(unreadable); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Stat(unreadable) error = %v, want ErrNotExist: the record should have been swept", err)
	}
}

// TestSweepRetainsCrashTruncatedRecord is the crash case the insert
// leaves behind: the record was created and the process died before its
// bytes were written. Deleting a zero-length record on sight would make
// its token consumable a second time for the rest of its expiry, so the
// sweep retains it until its mtime has passed the retention.
func TestSweepRetainsCrashTruncatedRecord(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)

	truncated := recordPathFor(t, stateDir, consumedSample)
	if err := os.WriteFile(truncated, nil, 0o600); err != nil {
		t.Fatalf("write zero-length record: %v", err)
	}

	sweep, err := nonces.Sweep(t.Context(), nonceClock)
	if err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if sweep.Removed != 0 || sweep.Malformed != 1 {
		t.Errorf("sweep = %+v, want {Removed:0 Malformed:1}", sweep)
	}
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); !errors.Is(err, ErrNonceReplayed) {
		t.Errorf("Insert after the sweep = %v, want ErrNonceReplayed", err)
	}

	// Once the record is older than the retention, no token bearing
	// its nonce can still verify, so the file is collectable.
	stale := nonceClock.Add(-fixtureRetention - time.Hour)
	if err := os.Chtimes(truncated, stale, stale); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
	sweep, err = nonces.Sweep(t.Context(), nonceClock)
	if err != nil {
		t.Fatalf("Sweep after ageing: %v", err)
	}
	if sweep.Removed != 1 || sweep.Malformed != 0 {
		t.Errorf("sweep after ageing = %+v, want {Removed:1 Malformed:0}", sweep)
	}
	if _, err := os.Stat(truncated); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stat after ageing = %v, want not-exist", err)
	}
}

// TestOpenSweeps proves the sweep at open is real: a store opened over
// a root holding an expired record prunes it before the caller's first
// insert, with [RunSweep]'s own interval driver never called.
func TestOpenSweeps(t *testing.T) {
	stateDir := t.TempDir()
	first := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := first.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, sweep, err := OpenNonceStore(t.Context(), store, fixed(fixtureRetention), fixtureLinkExpiryCeiling, nonceClock.Add(fixtureRetention+time.Second))
	if err != nil {
		t.Fatalf("OpenNonceStore: %v", err)
	}
	if sweep.Removed != 1 {
		t.Errorf("Removed at open = %d, want 1", sweep.Removed)
	}
	if entries := nonceDirEntries(t, stateDir); len(entries) != 0 {
		t.Errorf("nonce directory holds %d records after the sweep at open, want 0", len(entries))
	}
}

// TestConcurrentInsertAdmitsOne is what separates an exclusive create
// from a check followed by a create: the check-then-create passes every
// sequential row above, because the window between its check and its
// create is only reachable from a second writer.
//
// The rounds are what make the row reliable rather than lucky. One race
// between two writers is lost more often than it is won, so a single
// round leaves a check-then-create insert passing most runs; across
// rounds it double-admits somewhere, while the exclusive create admits
// exactly one writer in every round.
func TestConcurrentInsertAdmitsOne(t *testing.T) {
	const (
		rounds  = 64
		writers = 8
	)
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)

	for round := range rounds {
		// One nonce per round, so each round races for a record
		// no earlier round created.
		var nonce [NonceLength]byte
		nonce[0] = byte(round)

		var wg sync.WaitGroup
		results := make([]error, writers)
		start := make(chan struct{})
		for i := range writers {
			wg.Go(func() {
				<-start
				results[i] = nonces.Insert(t.Context(), nonce, nonceClock, sampleExpiry)
			})
		}
		close(start)
		wg.Wait()

		admitted := 0
		for i, err := range results {
			switch {
			case err == nil:
				admitted++
			case errors.Is(err, ErrNonceReplayed):
			default:
				t.Fatalf("round %d writer %d: %v, want nil or ErrNonceReplayed", round, i, err)
			}
		}
		if admitted != 1 {
			t.Fatalf("round %d: %d of %d concurrent inserts succeeded, want exactly 1", round, admitted, writers)
		}
	}
}

// TestRunSweepRunsOnTickWithoutWaitingSweepInterval proves the driver
// calls sweep on receipt of a tick rather than on a real-time schedule:
// the test drives tick directly and never waits SweepInterval (one
// hour, the production ticker's period) out.
func TestRunSweepRunsOnTickWithoutWaitingSweepInterval(t *testing.T) {
	if SweepInterval != time.Hour {
		t.Fatalf("SweepInterval = %v, want 1h", SweepInterval)
	}

	calls := make(chan struct{}, 1)
	sweep := func(context.Context, time.Time) (NonceSweep, error) {
		calls <- struct{}{}
		return NonceSweep{Removed: 1}, nil
	}
	tick := make(chan time.Time, 1)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go RunSweep(ctx, sweep, func() time.Time { return nonceClock }, tick, discardLogger())

	tick <- nonceClock
	select {
	case <-calls:
	case <-time.After(5 * time.Second):
		t.Fatal("RunSweep did not call sweep after a tick")
	}
}

// TestRunSweepStopsWithContext is the leak assertion: canceling ctx
// stops the driver's goroutine, proved by a channel the goroutine
// itself closes on return rather than by inspecting RunSweep's source.
func TestRunSweepStopsWithContext(t *testing.T) {
	sweep := func(context.Context, time.Time) (NonceSweep, error) { return NonceSweep{}, nil }
	tick := make(chan time.Time)
	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan struct{})
	go func() {
		defer close(done)
		RunSweep(ctx, sweep, func() time.Time { return nonceClock }, tick, discardLogger())
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("RunSweep left its goroutine running after ctx was canceled")
	}
}

// TestRunSweepLogsAFailureAndContinues proves a failed sweep never
// stops the driver: a second tick after a failing sweep still reaches
// sweep, since a sweep's own error is never fatal after startup.
func TestRunSweepLogsAFailureAndContinues(t *testing.T) {
	calls := make(chan struct{}, 2)
	first := true
	sweep := func(context.Context, time.Time) (NonceSweep, error) {
		calls <- struct{}{}
		if first {
			first = false
			return NonceSweep{}, errors.New("boom")
		}
		return NonceSweep{}, nil
	}
	tick := make(chan time.Time, 2)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go RunSweep(ctx, sweep, func() time.Time { return nonceClock }, tick, discardLogger())

	tick <- nonceClock
	tick <- nonceClock
	for range 2 {
		select {
		case <-calls:
		case <-time.After(5 * time.Second):
			t.Fatal("RunSweep stopped calling sweep after the first call failed")
		}
	}
}

// discardLogger returns a logger RunSweep's tests can pass without
// asserting on its output.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestNonceErrorsCarryNoNonceValue holds the package's own rule on the
// one store that handles a nonce: a refusal names the check, never the
// value that failed it.
func TestNonceErrorsCarryNoNonceValue(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry)
	if err == nil {
		t.Fatal("second Insert returned nil")
	}

	for _, value := range []string{
		string(consumedSample[:]),
		RecordID(nonceSubdir, string(consumedSample[:])),
	} {
		if strings.Contains(err.Error(), value) {
			t.Errorf("replay error echoes a value it must never carry (%d bytes)", len(value))
		}
	}
}

// TestNonceStore_ReadsTheRetentionAtEachUse proves the store reads the
// retention in force at each insert and each sweep: a record inserted
// after the retention rises carries the higher value, and a sweep after
// the retention drops prunes nothing a record still carries.
func TestNonceStore_ReadsTheRetentionAtEachUse(t *testing.T) {
	stateDir := t.TempDir()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	retention := newSettingFake(fixtureRetention)
	nonces, _, err := OpenNonceStore(t.Context(), store, retention.setting(), fixtureLinkExpiryCeiling, nonceClock)
	if err != nil {
		t.Fatalf("OpenNonceStore: %v", err)
	}
	retention.set(2*fixtureRetention, nil)
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	retention.set(time.Hour, nil)
	if _, err := nonces.Sweep(t.Context(), nonceClock.Add(fixtureRetention+time.Second)); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if _, err := os.Stat(recordPathFor(t, stateDir, consumedSample)); err != nil {
		t.Errorf("the record inserted under the raised retention was swept at the original one: %v", err)
	}
}

// TestNonceStore_RetentionReadFailure proves a failed retention read
// fails the insert before any record exists and fails the sweep before
// it removes anything, each with an error wrapping errSettingUnreadable
// and the reader's text, never a default retention.
func TestNonceStore_RetentionReadFailure(t *testing.T) {
	stateDir := t.TempDir()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	retention := newSettingFake(fixtureRetention)
	nonces, _, err := OpenNonceStore(t.Context(), store, retention.setting(), fixtureLinkExpiryCeiling, nonceClock)
	if err != nil {
		t.Fatalf("OpenNonceStore: %v", err)
	}
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	retention.set(0, errFixtureRead)

	err = nonces.Insert(t.Context(), secondSample, nonceClock, sampleExpiry)
	if !errors.Is(err, errSettingUnreadable) || !errors.Is(err, errFixtureRead) {
		t.Errorf("Insert error = %v, want errSettingUnreadable wrapping the reader's error", err)
	}
	if _, statErr := os.Stat(recordPathFor(t, stateDir, secondSample)); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("a failed read left a record for the second nonce (stat: %v)", statErr)
	}

	sweep, err := nonces.Sweep(t.Context(), nonceClock.Add(10*fixtureRetention))
	if !errors.Is(err, errSettingUnreadable) || !strings.Contains(err.Error(), "DUBPLATE_FIXTURE_KEY") {
		t.Errorf("Sweep error = %v, want errSettingUnreadable naming the key", err)
	}
	if sweep != (NonceSweep{}) {
		t.Errorf("Sweep = %+v, want nothing done", sweep)
	}
	if got := len(nonceDirEntries(t, stateDir)); got != 1 {
		t.Errorf("records after the failed sweep = %d, want the 1 inserted", got)
	}
}

// TestRunSweep_RetentionReadFailureLogsTheKey proves the sweep driver
// logs a failed retention read naming the key and deletes nothing,
// driving a real store's Sweep.
func TestRunSweep_RetentionReadFailureLogsTheKey(t *testing.T) {
	stateDir := t.TempDir()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	retention := newSettingFake(fixtureRetention)
	nonces, _, err := OpenNonceStore(t.Context(), store, retention.setting(), fixtureLinkExpiryCeiling, nonceClock)
	if err != nil {
		t.Fatalf("OpenNonceStore: %v", err)
	}
	if err := nonces.Insert(t.Context(), consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	retention.set(0, errFixtureRead)

	logs := newLogCapture()
	tick := make(chan time.Time, 1)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go RunSweep(ctx, nonces.Sweep, func() time.Time { return nonceClock.Add(10 * fixtureRetention) }, tick, slog.New(slog.NewJSONHandler(logs, nil)))
	tick <- nonceClock

	rec := logs.waitFor(t, "nonce sweep failed")
	if reason, _ := rec["reason"].(string); !strings.Contains(reason, "DUBPLATE_FIXTURE_KEY") {
		t.Errorf("log reason = %q, want it to name the key", reason)
	}
	if got := len(nonceDirEntries(t, stateDir)); got != 1 {
		t.Errorf("records after the failed sweep = %d, want the 1 inserted", got)
	}
}
