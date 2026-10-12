package intake

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/intake/gate"
)

// setModTime stamps name under staging with at.
func (f *intakeFixture) setModTime(t *testing.T, name string, at time.Time) {
	t.Helper()
	if err := os.Chtimes(f.path(name), at, at); err != nil {
		t.Fatal(err)
	}
}

// TestSweepExpiresPastTheirWindows is AC 28's windows: an upload idle
// past the resume window and a held check past the hold expire, and
// their bytes go in the same sweep; an upload whose data file was
// written inside the window, though its updated_at is older, and a
// check held inside the hold are left alone.
func TestSweepExpiresPastTheirWindows(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	idle := mustCreate(t, f.s, m, album(10), testLimits())
	f.stageUpload(t, idle.files[0].id, []byte("abc"))
	f.setModTime(t, tusOf(idle.files[0].id), time.UnixMicro(testNow))
	busy := mustCreate(t, f.s, m, album(10), testLimits())
	f.stageUpload(t, busy.files[0].id, []byte("abc"))
	heldLong := inState(t, f.s, m, stateCheckUnfinished, 10)

	f.clock.advance(testWindows.ResumeWindow + 2*time.Hour)
	f.setModTime(t, tusOf(busy.files[0].id), f.clock.Now().Add(-time.Hour))
	heldShort := inState(t, f.s, m, stateCheckUnfinished, 10)

	f.in.sweep(deadline(t))

	for _, tc := range []struct {
		name string
		c    created
		want string
	}{
		{"an idle upload", idle, "expired"},
		{"an upload written inside the window", busy, "uploading"},
		{"a check held past the hold", heldLong, "expired"},
		{"a check held inside the hold", heldShort, "check-unfinished"},
	} {
		if got := stateOf(t, f.h, tc.c.id); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, got, tc.want)
		}
	}
	if f.exists(t, tusOf(idle.files[0].id)) {
		t.Error("the expired upload's bytes outlived the sweep that expired it")
	}
	if got := f.read(t, tusOf(busy.files[0].id)); string(got) != "abc" {
		t.Errorf("the live upload holds %q, want abc", got)
	}
	assertStatesMatchEvents(t, f.h)
}

// TestSweepStallsOnceAndMailsAfterTheCommit is AC 28's stall: two sweeps
// over a check past STALL_AFTER record one stalled event and send one
// mail to each live admin, each sent with no write callback holding the
// store. A check inside the threshold is left alone.
func TestSweepStallsOnceAndMailsAfterTheCommit(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	tracks := twoTrackAlbum(t)
	long := f.checkingWith(t, m, albumOf(tracks...), tracks...)
	f.clock.advance(testWindows.StallAfter + time.Minute)
	recent := f.checkingWith(t, m, albumOf(tracks...), tracks...)

	f.in.sweep(deadline(t))
	f.in.sweep(deadline(t))

	if got := eventsOf(t, f.h, long.id); !slices.Equal(got, []string{"uploading", "checking", "stalled"}) {
		t.Errorf("events = %q, want one stalled event", got)
	}
	if got := eventsOf(t, f.h, recent.id); slices.Contains(got, "stalled") {
		t.Errorf("a check inside the threshold stalled: %q", got)
	}
	mails := f.mail.mails()
	var to []string
	for _, mail := range mails {
		to = append(to, mail.to)
		if mail.waited {
			t.Errorf("the mail to %s went while a write callback held the store", mail.to)
		}
		if !bytes.Contains([]byte(mail.body), []byte(wireID(long.id))) {
			t.Errorf("the mail to %s names no stalled submission: %q", mail.to, mail.body)
		}
	}
	if !slices.Equal(to, []string{"ann@example.invalid", "bo@example.invalid"}) {
		t.Errorf("mailed %q, want each live admin once", to)
	}
	if got := stateOf(t, f.h, long.id); got != "checking" {
		t.Errorf("state = %q, want checking: a stall is an event, not a state", got)
	}
	assertStatesMatchEvents(t, f.h)
}

// TestSweepLeavesAnUploadingPartialFile: a sweep over an uploading
// submission's half-written, unlocked tus upload leaves both its files
// byte-identical.
func TestSweepLeavesAnUploadingPartialFile(t *testing.T) {
	f := newIntakeFixture(t)
	c := mustCreate(t, f.s, member(t, f.h), album(10), testLimits())
	f.stageUpload(t, c.files[0].id, []byte("half"))
	info := f.read(t, tusOf(c.files[0].id)+".info")

	f.in.sweep(deadline(t))

	if got := f.read(t, tusOf(c.files[0].id)); string(got) != "half" {
		t.Errorf("data file = %q, want half", got)
	}
	if got := f.read(t, tusOf(c.files[0].id)+".info"); !bytes.Equal(got, info) {
		t.Errorf(".info = %q, want %q", got, info)
	}
}

// TestSweepDuringExtraction: a sweep run while a checking archive stands
// paused mid-extraction leaves its x/ alone, and the check ends at the
// verdict an uninterrupted run reaches.
func TestSweepDuringExtraction(t *testing.T) {
	f := newIntakeFixture(t)
	archive := twoTrackZIP(t)
	c := f.checkingWith(t, member(t, f.h), zipped(gate.UnitAlbum, archive), archive)
	xDir := "submissions/" + wireID(c.id) + "/x"

	ctx := newPauseWhen(func() bool { return f.entries(t, xDir) == 1 })
	done := make(chan struct{})
	go func() { defer close(done); f.in.check(ctx, c.id) }()
	<-ctx.paused
	f.in.sweep(deadline(t))
	close(ctx.release)
	<-done

	if got := stateOf(t, f.h, c.id); got != "matching" {
		t.Fatalf("state = %q (%s), want matching", got, eventKinds(t, f.h, c.id))
	}
	for path, id := range membersOf(t, f.h, c.id) {
		if !f.exists(t, xOf(c.id, id)) {
			t.Errorf("member %s is gone", path)
		}
	}
	assertStatesMatchEvents(t, f.h)
}

// TestCleanupLeavesAFileWhoseLockIsHeld: a cancelled submission's upload
// whose tus lock another holder keeps, asked to release and never
// releasing, survives the first sweep; the next sweep after the release
// removes it.
func TestCleanupLeavesAFileWhoseLockIsHeld(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	c := mustCreate(t, f.s, m, album(10), testLimits())
	f.stageUpload(t, c.files[0].id, []byte("partial"))
	if err := f.s.cancel(deadline(t), m, c.id); err != nil {
		t.Fatal(err)
	}
	lock, err := f.in.locker.NewLock(wireID(c.files[0].id))
	if err != nil {
		t.Fatal(err)
	}
	var asked atomic.Bool
	if err := lock.Lock(deadline(t), func() { asked.Store(true) }); err != nil {
		t.Fatal(err)
	}

	f.in.sweep(deadline(t))
	if !asked.Load() {
		t.Error("cleanup never asked the holder to release")
	}
	if !f.exists(t, tusOf(c.files[0].id)) {
		t.Fatal("cleanup removed an upload whose lock it could not take")
	}

	if err := lock.Unlock(); err != nil {
		t.Fatal(err)
	}
	f.in.sweep(deadline(t))
	if f.exists(t, tusOf(c.files[0].id)) || f.exists(t, tusOf(c.files[0].id)+".info") {
		t.Error("the released upload survived the next sweep")
	}
}

// TestSweepRemovesAnOrphanPastTheWindow: a tus upload no row names goes
// once it is older than the resume window, and not before.
func TestSweepRemovesAnOrphanPastTheWindow(t *testing.T) {
	f := newIntakeFixture(t)
	old, young := store.NewID(), store.NewID()
	f.stageUpload(t, old, []byte("old"))
	f.stageUpload(t, young, []byte("young"))
	f.clock.advance(testWindows.ResumeWindow + time.Hour)
	for _, name := range []string{tusOf(old), tusOf(old) + ".info"} {
		f.setModTime(t, name, time.UnixMicro(testNow))
	}
	for _, name := range []string{tusOf(young), tusOf(young) + ".info"} {
		f.setModTime(t, name, f.clock.Now().Add(-time.Hour))
	}

	f.in.sweep(deadline(t))

	if f.exists(t, tusOf(old)) || f.exists(t, tusOf(old)+".info") {
		t.Error("an orphan older than the window survived")
	}
	if !f.exists(t, tusOf(young)) || !f.exists(t, tusOf(young)+".info") {
		t.Error("an orphan inside the window was removed")
	}
}

// TestPatchInFlightAfterCancelIsGoneAfterOneSweep is AC 11's last
// clause: a PATCH still sending when its submission is cancelled is
// stopped through its lock by the next sweep, which removes its bytes,
// and the submission stays cancelled.
func TestPatchInFlightAfterCancelIsGoneAfterOneSweep(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	c := mustCreate(t, f.s, m, album(1000), testLimits())
	mustStatus(t, "create", f.do(t, createRequest(m, c, 0, 1000, nil)), http.StatusCreated)

	body, feed := io.Pipe()
	defer func() { _ = feed.Close() }()
	r, err := http.NewRequestWithContext(deadline(t), http.MethodPatch, f.url+uploadPath(c.files[0].id), body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set(memberHeader, wireID(m[:]))
	r.Header.Set("Tus-Resumable", "1.0.0")
	r.Header.Set("Upload-Offset", "0")
	r.Header.Set("Content-Type", "application/offset+octet-stream")
	r.ContentLength = 1000
	answered := make(chan int, 1)
	go func() {
		resp, err := f.client.Do(r)
		if err != nil {
			answered <- 0
			return
		}
		_ = resp.Body.Close()
		answered <- resp.StatusCode
	}()
	if _, err := feed.Write(bytes.Repeat([]byte("x"), 100)); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return len(f.tusFile(t, c.files[0].id)) == 100 })

	if err := f.s.cancel(deadline(t), m, c.id); err != nil {
		t.Fatal(err)
	}
	f.in.sweep(deadline(t))

	if f.tusFile(t, c.files[0].id) != nil || f.infoExists(t, c.files[0].id) {
		t.Error("a cancelled upload's bytes survived the sweep")
	}
	_ = feed.Close()
	<-answered
	if got := stateOf(t, f.h, c.id); got != "cancelled" {
		t.Errorf("state = %q, want cancelled", got)
	}
}

// waitFor polls cond until it holds.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition never held")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
