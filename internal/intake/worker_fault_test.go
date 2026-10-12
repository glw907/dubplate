package intake

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

// TestCheckStoreFaultLeavesTheSubmissionChecking is P19's other half: a
// fault of the store's, here the limits read failing, is never the member's
// verdict. The submission stays checking with its bytes, reservation, and
// events untouched, and the next check, with the fault cleared, reaches the
// verdict an uninterrupted run reaches.
func TestCheckStoreFaultLeavesTheSubmissionChecking(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	tracks := twoTrackAlbum(t)
	c := f.checkingWith(t, m, albumOf(tracks...), tracks...)
	events, reserved := eventKinds(t, f.h, c.id), reservedBytes(t, f.h, c.id)

	good := f.in.cfg.Limits
	f.in.cfg.Limits = func(context.Context) (Limits, error) {
		return Limits{}, errors.New("settings read: database is locked")
	}
	f.in.check(deadline(t), c.id)

	if got := stateOf(t, f.h, c.id); got != "checking" {
		t.Fatalf("state after a store fault = %q, want checking", got)
	}
	if got := eventKinds(t, f.h, c.id); got != events {
		t.Errorf("events = %s, want %s: a fault wrote an event", got, events)
	}
	if got := reservedBytes(t, f.h, c.id); got != reserved {
		t.Errorf("reservation = %d, want %d", got, reserved)
	}
	for i, file := range c.files {
		if !bytes.Equal(f.read(t, tusOf(file.id)), tracks[i]) {
			t.Errorf("%s: staged bytes changed or gone", file.path)
		}
	}

	f.in.cfg.Limits = good
	f.in.check(deadline(t), c.id)
	if got := stateOf(t, f.h, c.id); got != "matching" {
		t.Fatalf("state after the fault cleared = %q (%s), want matching", got, eventKinds(t, f.h, c.id))
	}
	assertStatesMatchEvents(t, f.h)
}

// TestCheckContainsAPanic: a panic in a decoder on a member's bytes ends
// that check as could-not-run, with the bytes and reservation kept, and does
// not take the server down, where the startup pass would re-queue the same
// submission into a crash loop.
func TestCheckContainsAPanic(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	tracks := twoTrackAlbum(t)
	c := f.checkingWith(t, m, albumOf(tracks...), tracks...)
	reserved := reservedBytes(t, f.h, c.id)

	f.in.cfg.Limits = func(context.Context) (Limits, error) { panic("decoder: index out of range") }
	f.in.check(deadline(t), c.id)

	if got := stateOf(t, f.h, c.id); got != "check-unfinished" {
		t.Fatalf("state after a panic = %q (%s), want check-unfinished", got, eventKinds(t, f.h, c.id))
	}
	if got := reservedBytes(t, f.h, c.id); got != reserved {
		t.Errorf("reservation = %d, want %d", got, reserved)
	}
	for i, file := range c.files {
		if !bytes.Equal(f.read(t, tusOf(file.id)), tracks[i]) {
			t.Errorf("%s: staged bytes changed or gone", file.path)
		}
	}
	assertStatesMatchEvents(t, f.h)
}
