package auth

import (
	"errors"
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

func newNonceStore(t *testing.T, stateDir string, retention time.Duration, now time.Time) *NonceStore {
	t.Helper()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	nonces, _, err := OpenNonceStore(store, retention, now)
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

	if err := nonces.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("first Insert: %v", err)
	}
	path := recordPathFor(t, stateDir, consumedSample)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read record: %v", err)
	}

	err = nonces.Insert(consumedSample, nonceClock.Add(time.Minute), sampleExpiry)
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
	if err := first.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert before restart: %v", err)
	}

	restarted := newNonceStore(t, stateDir, fixtureRetention, nonceClock.Add(time.Minute))
	err := restarted.Insert(consumedSample, nonceClock.Add(time.Minute), sampleExpiry)
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
		if err := nonces.Insert(tok.Nonce, nonceClock, tok.ExpiresAt); err != nil {
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
			if err := nonces.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
				t.Fatalf("Insert: %v", err)
			}

			sweep, err := nonces.Sweep(tt.sweepAt)
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
	if err := nonces.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	sweepAt := nonceClock.Add(fixtureExpiry).Add(time.Hour)
	if _, err := nonces.Sweep(sweepAt); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if _, err := os.Stat(recordPathFor(t, stateDir, consumedSample)); err != nil {
		t.Fatalf("record removed %v after mint, before its retention elapsed: %v", sweepAt.Sub(nonceClock), err)
	}

	// A sweep past the retention does remove it, so the survival
	// above is the carried expiry holding rather than a sweep that
	// never removes anything.
	if _, err := nonces.Sweep(nonceClock.Add(fixtureRetention + time.Second)); err != nil {
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
	if err := nonces.Insert(tok.Nonce, nonceClock, tok.ExpiresAt); err != nil {
		t.Fatalf("first Insert: %v", err)
	}
	if err := nonces.Insert(tok.Nonce, nonceClock, tok.ExpiresAt); !errors.Is(err, ErrNonceReplayed) {
		t.Fatalf("second Insert = %v, want ErrNonceReplayed", err)
	}

	sweepAt := nonceClock.Add(2 * shortRetention)
	sweep, err := nonces.Sweep(sweepAt)
	if err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if sweep.Removed != 0 {
		t.Errorf("Removed = %d at %v past the mint, want 0 while the token is live", sweep.Removed, sweepAt.Sub(nonceClock))
	}
	if _, err := VerifyLink(current, nil, value, sweepAt); err != nil {
		t.Fatalf("VerifyLink at %v: %v, want a live token", sweepAt, err)
	}
	if err := nonces.Insert(tok.Nonce, sweepAt, tok.ExpiresAt); !errors.Is(err, ErrNonceReplayed) {
		t.Fatalf("replay Insert after the sweep = %v, want ErrNonceReplayed", err)
	}

	// Past the token's expiry the record is collectable, so the
	// survival above is the token's expiry holding rather than a
	// sweep that never removes anything.
	sweep, err = nonces.Sweep(tok.ExpiresAt.Add(time.Second))
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
	if err := nonces.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	garbage := recordPathFor(t, stateDir, secondSample)
	if err := os.WriteFile(garbage, []byte("not a record at all\n"), 0o600); err != nil {
		t.Fatalf("write malformed record: %v", err)
	}

	sweep, err := nonces.Sweep(nonceClock.Add(fixtureRetention + time.Second))
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
// down.
func TestSweepRetainsUnreadableRecord(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := nonces.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
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

	sweep, err := nonces.Sweep(nonceClock.Add(fixtureRetention + time.Second))
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

	// OpenNonceStore runs a sweep at open, so the same store carrying
	// the same unreadable record must still start.
	if err := os.Chmod(unreadable, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, err := OpenNonceStore(store, fixtureRetention, nonceClock.Add(fixtureRetention+time.Second)); err != nil {
		t.Errorf("OpenNonceStore: %v, want nil: a single unreadable record must not stop startup", err)
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

	sweep, err := nonces.Sweep(nonceClock)
	if err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if sweep.Removed != 0 || sweep.Malformed != 1 {
		t.Errorf("sweep = %+v, want {Removed:0 Malformed:1}", sweep)
	}
	if err := nonces.Insert(consumedSample, nonceClock, sampleExpiry); !errors.Is(err, ErrNonceReplayed) {
		t.Errorf("Insert after the sweep = %v, want ErrNonceReplayed", err)
	}

	// Once the record is older than the retention, no token bearing
	// its nonce can still verify, so the file is collectable.
	stale := nonceClock.Add(-fixtureRetention - time.Hour)
	if err := os.Chtimes(truncated, stale, stale); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
	sweep, err = nonces.Sweep(nonceClock)
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
// insert, with no interval driver in this package.
func TestOpenSweeps(t *testing.T) {
	stateDir := t.TempDir()
	first := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := first.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, sweep, err := OpenNonceStore(store, fixtureRetention, nonceClock.Add(fixtureRetention+time.Second))
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
				results[i] = nonces.Insert(nonce, nonceClock, sampleExpiry)
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

// TestNonceErrorsCarryNoNonceValue holds the package's own rule on the
// one store that handles a nonce: a refusal names the check, never the
// value that failed it.
func TestNonceErrorsCarryNoNonceValue(t *testing.T) {
	stateDir := t.TempDir()
	nonces := newNonceStore(t, stateDir, fixtureRetention, nonceClock)
	if err := nonces.Insert(consumedSample, nonceClock, sampleExpiry); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	err := nonces.Insert(consumedSample, nonceClock, sampleExpiry)
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
