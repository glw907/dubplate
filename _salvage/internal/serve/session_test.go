package serve

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/reject"
)

func TestNewSessionIDFormat(t *testing.T) {
	seen := make(map[string]bool)
	for range 50 {
		id, err := newSessionID()
		if err != nil {
			t.Fatalf("newSessionID: %v", err)
		}
		if !sessionIDPattern.MatchString(id) {
			t.Fatalf("newSessionID() = %q, does not match %s", id, sessionIDPattern)
		}
		if !validSessionID(id) {
			t.Fatalf("validSessionID(%q) = false, want true", id)
		}
		if seen[id] {
			t.Fatalf("newSessionID produced a duplicate: %q", id)
		}
		seen[id] = true
	}
}

func TestValidSessionID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"empty", "", false},
		{"too short", "abc2345", false},
		{"uppercase", strings.Repeat("A", 26), false},
		{"digit outside 2-7", strings.Repeat("a", 25) + "1", false},
		{"path traversal", "../../../etc/passwd0000000", false},
		{"valid shape", strings.Repeat("a", 26), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validSessionID(tt.id); got != tt.want {
				t.Errorf("validSessionID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestSlugifyAlbum(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"simple", "Abbey Road", "abbey-road", false},
		{"path traversal", "../../etc/passwd", "etc-passwd", false},
		{"absolute path", "/etc/passwd", "etc-passwd", false},
		{"embedded NUL", "abbey\x00road", "abbey-road", false},
		{"only punctuation", "!!!///...", "", true},
		{"empty", "", "", true},
		{"leading dot equivalent", "...hidden", "hidden", false},
		{"over length cap", strings.Repeat("a", maxAlbumSlugLen+50), strings.Repeat("a", maxAlbumSlugLen), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := slugifyAlbum(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("slugifyAlbum(%q) = %q, nil, want error", tt.raw, got)
				}
				if !errors.Is(err, ErrInvalidAlbum) {
					t.Errorf("slugifyAlbum(%q) error = %v, want ErrInvalidAlbum", tt.raw, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("slugifyAlbum(%q) unexpected error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Errorf("slugifyAlbum(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

// TestSlugifyAlbumUnicodeNormalization proves two different byte
// encodings of the same visual album name -- NFC precomposed versus NFD
// combining-mark decomposed -- collapse to the identical slug, so a
// contributor can never bypass a filter or alias a second directory by
// choosing the other normalization form.
func TestSlugifyAlbumUnicodeNormalization(t *testing.T) {
	nfc := "Café Session"  // é as one precomposed rune
	nfd := "Café Session" // e + combining acute accent

	gotNFC, err := slugifyAlbum(nfc)
	if err != nil {
		t.Fatalf("slugifyAlbum(NFC) error: %v", err)
	}
	gotNFD, err := slugifyAlbum(nfd)
	if err != nil {
		t.Fatalf("slugifyAlbum(NFD) error: %v", err)
	}
	if gotNFC != gotNFD {
		t.Fatalf("slugifyAlbum(NFC) = %q, slugifyAlbum(NFD) = %q, want equal", gotNFC, gotNFD)
	}
}

func TestValidCollection(t *testing.T) {
	collections := []string{"", "family", "archive"}
	tests := []struct {
		name string
		want bool
	}{
		{"family", true},
		{"", true},
		{"unconfigured", false},
		{"../../etc", false},
	}
	for _, tt := range tests {
		if got := validCollection(tt.name, collections); got != tt.want {
			t.Errorf("validCollection(%q, %v) = %v, want %v", tt.name, collections, got, tt.want)
		}
	}
}

func TestStoreCreateRejectsUnknownCollection(t *testing.T) {
	s := newStore()
	_, err := s.create("a@example.com", "Abbey Road", "bootlegs", 1, []string{"", "family"}, time.Now())
	if !errors.Is(err, ErrUnknownCollection) {
		t.Fatalf("create() error = %v, want ErrUnknownCollection", err)
	}
}

func TestStoreCreateRejectsInvalidAlbum(t *testing.T) {
	s := newStore()
	_, err := s.create("a@example.com", "///", "", 1, []string{""}, time.Now())
	if !errors.Is(err, ErrInvalidAlbum) {
		t.Fatalf("create() error = %v, want ErrInvalidAlbum", err)
	}
}

func TestStoreCreateRejectsZeroExpectedFiles(t *testing.T) {
	s := newStore()
	if _, err := s.create("a@example.com", "Abbey Road", "", 0, []string{""}, time.Now()); err == nil {
		t.Fatal("create() with expected_files=0 succeeded, want error")
	}
}

// TestStoreCreateRejectsTooManyExpectedFiles proves create bounds
// expected_files: with no cap, a session could reserve an unbounded
// number of upload slots regardless of the aggregate byte cap
// [Session.reserve] separately enforces.
func TestStoreCreateRejectsTooManyExpectedFiles(t *testing.T) {
	s := newStore()
	if _, err := s.create("a@example.com", "Huge Box Set", "", maxExpectedFiles+1, []string{""}, time.Now()); err == nil {
		t.Fatalf("create() with expected_files=%d succeeded, want error", maxExpectedFiles+1)
	}
}

func TestStoreCountActive(t *testing.T) {
	s := newStore()
	now := time.Now()
	sess1, err := s.create("a@example.com", "Album One", "", 1, []string{""}, now)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.create("a@example.com", "Album Two", "", 1, []string{""}, now); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.create("b@example.com", "Other Person", "", 1, []string{""}, now); err != nil {
		t.Fatalf("create: %v", err)
	}

	if got := s.countActive("a@example.com"); got != 2 {
		t.Fatalf("countActive(a) = %d, want 2", got)
	}
	if got := s.countActive("b@example.com"); got != 1 {
		t.Fatalf("countActive(b) = %d, want 1", got)
	}
	if got := s.countActive("nobody@example.com"); got != 0 {
		t.Fatalf("countActive(nobody) = %d, want 0", got)
	}

	// Completing sess1 removes it from the active count.
	if !sess1.completeOnce() {
		t.Fatal("completeOnce() = false on a fresh session")
	}
	sess1.setOutcome(sessionOutcome{Promoted: true}, time.Now())
	if got := s.countActive("a@example.com"); got != 1 {
		t.Fatalf("countActive(a) after completion = %d, want 1", got)
	}
}

func TestSessionReserveBoundsAtExpectedFiles(t *testing.T) {
	s := newStore()
	sess, err := s.create("a@example.com", "Album", "", 2, []string{""}, time.Now())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	const maxBytes = 1 << 30
	slot0, err := sess.reserve(1024, maxBytes)
	if err != nil || slot0 != 0 {
		t.Fatalf("reserve() #1 = (%d, %v), want (0, nil)", slot0, err)
	}
	slot1, err := sess.reserve(1024, maxBytes)
	if err != nil || slot1 != 1 {
		t.Fatalf("reserve() #2 = (%d, %v), want (1, nil)", slot1, err)
	}
	if _, err := sess.reserve(1024, maxBytes); !errors.Is(err, errSessionFull) {
		t.Fatalf("reserve() #3 (the N+1st) error = %v, want errSessionFull", err)
	}
}

// TestSessionReserveBoundsAggregateBytes proves reserve refuses a slot
// once the session's running total would exceed maxBytes, independent of
// [Session.ExpectedFiles]: the aggregate album-total cap F2.5 found the
// multi-file path never enforced.
func TestSessionReserveBoundsAggregateBytes(t *testing.T) {
	s := newStore()
	sess, err := s.create("a@example.com", "Album", "", 5, []string{""}, time.Now())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	const maxBytes = 1000
	if _, err := sess.reserve(600, maxBytes); err != nil {
		t.Fatalf("reserve() #1 (600 of 1000) = %v, want nil", err)
	}
	if _, err := sess.reserve(500, maxBytes); !errors.Is(err, errAlbumTooLarge) {
		t.Fatalf("reserve() #2 (600+500 over 1000) error = %v, want errAlbumTooLarge", err)
	}
	// A slot that still fits within what's left is admitted.
	if _, err := sess.reserve(300, maxBytes); err != nil {
		t.Fatalf("reserve() #3 (600+300 of 1000) = %v, want nil", err)
	}
}

func TestSessionUploadDoneCompletesByCount(t *testing.T) {
	s := newStore()
	sess, err := s.create("a@example.com", "Album", "", 2, []string{""}, time.Now())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if done := sess.uploadDone([]intake.FileVerdict{{Name: "one.flac", OK: true}}); done {
		t.Fatal("uploadDone() reported done after 1 of 2 uploads")
	}
	if done := sess.uploadDone([]intake.FileVerdict{{Name: "two.flac", OK: true}}); !done {
		t.Fatal("uploadDone() reported not-done after 2 of 2 uploads")
	}

	verdicts := sess.verdictsSnapshot()
	if len(verdicts) != 2 {
		t.Fatalf("verdictsSnapshot() has %d entries, want 2", len(verdicts))
	}
}

// TestSessionCompleteOnceIsExclusive proves that when the count-reached
// path and an explicit /finish race for the same session, only one
// caller runs the completion logic; the other reads back the same
// cached outcome instead of running it a second time.
func TestSessionCompleteOnceIsExclusive(t *testing.T) {
	sess := &Session{ID: "x", ExpectedFiles: 1}

	const racers = 20
	var wg sync.WaitGroup
	wins := make(chan int, racers)
	wg.Add(racers)
	for i := range racers {
		go func(i int) {
			defer wg.Done()
			if sess.completeOnce() {
				wins <- i
			}
		}(i)
	}
	wg.Wait()
	close(wins)

	count := 0
	for range wins {
		count++
	}
	if count != 1 {
		t.Fatalf("completeOnce() returned true %d times across %d racers, want exactly 1", count, racers)
	}
}

func TestSessionWaitOutcomeBlocksUntilSet(t *testing.T) {
	sess := &Session{ID: "x", ExpectedFiles: 1}
	if !sess.completeOnce() {
		t.Fatal("completeOnce() = false on a fresh session")
	}

	done := make(chan sessionOutcome, 1)
	go func() {
		out, err := sess.waitOutcome(context.Background())
		if err != nil {
			t.Errorf("waitOutcome: %v", err)
			return
		}
		done <- out
	}()

	want := sessionOutcome{Promoted: true, Dest: "/inbox/x"}
	sess.setOutcome(want, time.Now())

	select {
	case got := <-done:
		if got.Dest != want.Dest || got.Promoted != want.Promoted {
			t.Fatalf("waitOutcome() = %+v, want %+v", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("waitOutcome never returned")
	}
}

func TestSessionWaitOutcomeContextCancel(t *testing.T) {
	sess := &Session{ID: "x", ExpectedFiles: 1}
	if !sess.completeOnce() {
		t.Fatal("completeOnce() = false on a fresh session")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := sess.waitOutcome(ctx); err == nil {
		t.Fatal("waitOutcome() with a cancelled context returned nil error")
	}
}

func TestStoreExpired(t *testing.T) {
	s := newStore()
	old := time.Now().Add(-2 * time.Hour)
	fresh := time.Now()

	stale, err := s.create("a@example.com", "Stale Album", "", 1, []string{""}, old)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.create("a@example.com", "Fresh Album", "", 1, []string{""}, fresh); err != nil {
		t.Fatalf("create: %v", err)
	}

	cutoff := time.Now().Add(-1 * time.Hour)
	expired := s.expired(cutoff)
	if len(expired) != 1 || expired[0].ID != stale.ID {
		t.Fatalf("expired(%v) = %v, want only %s", cutoff, ids(expired), stale.ID)
	}

	// A completed session is never reported expired, even if it is old.
	stale.completeOnce()
	stale.setOutcome(sessionOutcome{Promoted: true}, time.Now())
	if got := s.expired(cutoff); len(got) != 0 {
		t.Fatalf("expired(%v) after completion = %v, want none", cutoff, ids(got))
	}
}

// TestStoreSettled proves settled reports a completed session only once
// its own outcome-set time (not its creation time) has crossed cutoff,
// and never reports one still in flight.
func TestStoreSettled(t *testing.T) {
	s := newStore()
	now := time.Now()

	if _, err := s.create("a@example.com", "Still Uploading", "", 2, []string{""}, now); err != nil {
		t.Fatalf("create: %v", err)
	}

	longSettled, err := s.create("a@example.com", "Long Settled", "", 1, []string{""}, now)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	longSettled.completeOnce()
	longSettled.setOutcome(sessionOutcome{Promoted: true}, now.Add(-time.Hour))

	justSettled, err := s.create("a@example.com", "Just Settled", "", 1, []string{""}, now)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	justSettled.completeOnce()
	justSettled.setOutcome(sessionOutcome{Promoted: true}, now)

	cutoff := now.Add(-time.Minute)
	got := s.settled(cutoff)
	if len(got) != 1 || got[0].ID != longSettled.ID {
		t.Fatalf("settled(%v) = %v, want only %s", cutoff, ids(got), longSettled.ID)
	}
}

func ids(sessions []*Session) []string {
	out := make([]string, len(sessions))
	for i, s := range sessions {
		out[i] = s.ID
	}
	return out
}

// TestReasonErrorCarriesWireShape proves admission-control rejections
// wrap reject.Reason's own {"code","friendly","next_step"} encoding
// rather than a bare status with no explanation.
func TestReasonErrorCarriesWireShape(t *testing.T) {
	err := reasonError(reject.ReasonNoSpace, 507)
	if err.Error() == "" {
		t.Fatal("reasonError().Error() is empty")
	}
	if !strings.Contains(err.Error(), "dubplate-reject") {
		t.Fatalf("reasonError().Error() = %q, want it to name the error code", err.Error())
	}
}
