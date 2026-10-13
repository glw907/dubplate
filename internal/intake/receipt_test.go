package intake

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
)

func receivedAt(t *testing.T, s *submissions, fileID []byte) *int64 {
	t.Helper()
	var at *int64
	if err := s.store.QueryRow(deadline(t), `SELECT received_at FROM submission_files WHERE id = ?`, fileID).Scan(&at); err != nil {
		t.Fatal(err)
	}
	return at
}

// TestReceiveMovesOnTheLastFile is AC 11's two-file submission: the
// first receipt marks its file, the last moves the submission to
// checking and pokes the worker, and a replay changes nothing.
func TestReceiveMovesOnTheLastFile(t *testing.T) {
	h, s := newTestSubmissions(t)
	c := mustCreate(t, s, member(t, h), album(10, 20), testLimits())
	drain(s)

	if err := s.receive(deadline(t), c.files[0].id); err != nil {
		t.Fatal(err)
	}
	if at := receivedAt(t, s, c.files[0].id); at == nil || *at != testNow {
		t.Errorf("received_at = %v, want %d", at, testNow)
	}
	if got := stateOf(t, h, c.id); got != "uploading" || len(s.wake) != 0 {
		t.Errorf("after one of two files: state %q, %d pokes; want uploading, none", got, len(s.wake))
	}

	if err := s.receive(deadline(t), c.files[1].id); err != nil {
		t.Fatal(err)
	}
	if got := stateOf(t, h, c.id); got != "checking" {
		t.Errorf("state = %q after the last file, want checking", got)
	}
	if len(s.wake) != 1 {
		t.Errorf("wake holds %d pokes, want 1", len(s.wake))
	}

	// A replay of either receipt is a no-op, and a poke into a full wake
	// channel does not block.
	for _, f := range c.files {
		if err := s.receive(deadline(t), f.id); err != nil {
			t.Errorf("replay of %s: %v", f.path, err)
		}
	}
	s.poke()
	if got := eventsOf(t, h, c.id); !slices.Equal(got, []string{"uploading", "checking"}) {
		t.Errorf("events = %q, want uploading then one checking", got)
	}
	assertStatesMatchEvents(t, h)
}

// TestReceiveSyncsBeforeItCommits is AC 11's fault between the last byte
// and the receipt commit: a sync that fails commits nothing, and the
// replay the startup pass or a sweep runs moves the submission.
func TestReceiveSyncsBeforeItCommits(t *testing.T) {
	h, s := newTestSubmissions(t)
	s.sync = syncPath
	c := mustCreate(t, s, member(t, h), album(10, 20), testLimits())
	tus := filepath.Join(s.staging, "tus")
	if err := os.Mkdir(tus, 0o750); err != nil {
		t.Fatal(err)
	}
	for _, f := range c.files {
		name := filepath.Join(tus, hex.EncodeToString(f.id))
		if err := os.WriteFile(name, []byte("audio"), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	// The second file's .info is not yet on disk, so its sync fails.
	if err := os.WriteFile(filepath.Join(tus, hex.EncodeToString(c.files[0].id)+".info"), []byte("{}"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := s.receive(deadline(t), c.files[0].id); err != nil {
		t.Fatal(err)
	}

	if err := s.receive(deadline(t), c.files[1].id); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("receive with a failing sync: error = %v, want the sync's", err)
	}
	if at := receivedAt(t, s, c.files[1].id); at != nil || stateOf(t, h, c.id) != "uploading" {
		t.Fatalf("a failed sync committed: received_at %v, state %q", at, stateOf(t, h, c.id))
	}

	if err := os.WriteFile(filepath.Join(tus, hex.EncodeToString(c.files[1].id)+".info"), []byte("{}"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := s.receive(deadline(t), c.files[1].id); err != nil {
		t.Fatalf("replay: %v", err)
	}
	if got := stateOf(t, h, c.id); got != "checking" {
		t.Errorf("state = %q after the replay, want checking", got)
	}
	assertStatesMatchEvents(t, h)
}

// TestTwoFilesFinishingAtOnce is AC 11's race: each receipt's sync, its
// last act before its Update, blocks until both receipts reach it, and
// the submission moves to checking exactly once.
func TestTwoFilesFinishingAtOnce(t *testing.T) {
	h, s := newTestSubmissions(t)
	c := mustCreate(t, s, member(t, h), album(10, 20), testLimits())
	tus := filepath.Join(s.staging, "tus")

	var reached sync.WaitGroup
	reached.Add(2)
	s.sync = func(path string) error {
		if path == tus {
			reached.Done()
			reached.Wait()
		}
		return nil
	}
	var errs [2]error
	var done sync.WaitGroup
	for i, f := range c.files {
		done.Go(func() { errs[i] = s.receive(deadline(t), f.id) })
	}
	done.Wait()

	for _, err := range errs {
		if err != nil {
			t.Errorf("receive: %v", err)
		}
	}
	if got := eventsOf(t, h, c.id); !slices.Equal(got, []string{"uploading", "checking"}) {
		t.Errorf("events = %q, want exactly one move to checking", got)
	}
	assertStatesMatchEvents(t, h)
}

// TestReceiveAfterCancel is AC 11's last clause at the store: a file
// that finishes after its submission was cancelled changes nothing, and
// receipt reports the closed submission.
func TestReceiveAfterCancel(t *testing.T) {
	h, s := newTestSubmissions(t)
	m := member(t, h)
	c := mustCreate(t, s, m, album(10, 20), testLimits())
	if err := s.receive(deadline(t), c.files[0].id); err != nil {
		t.Fatal(err)
	}
	if err := s.cancel(deadline(t), m, c.id); err != nil {
		t.Fatal(err)
	}

	if err := s.receive(deadline(t), c.files[1].id); !errors.Is(err, errWrongState) {
		t.Fatalf("error = %v, want errWrongState", err)
	}
	if at := receivedAt(t, s, c.files[1].id); at != nil {
		t.Errorf("received_at = %d on a cancelled submission, want NULL", *at)
	}
	if got := eventsOf(t, h, c.id); !slices.Equal(got, []string{"uploading", "cancelled"}) {
		t.Errorf("events = %q, want uploading then cancelled", got)
	}
	if err := s.receive(deadline(t), []byte("0123456789abcdef")); !errors.Is(err, errNotFound) {
		t.Errorf("unknown file: error = %v, want errNotFound", err)
	}
	assertStatesMatchEvents(t, h)
}
