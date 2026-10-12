package intake

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// testNow is the fixed instant every fixture write is stamped with, in
// microseconds.
const testNow = 1_700_000_000_000_000

// plentyFree is a free-space read no fixture's floor comes near.
const plentyFree = 1 << 50

// newTestSubmissions returns a store with one live admin, so the route
// collections can be created, and a submissions writer over it whose
// clock is fixed, whose free-space read answers plentyFree, and whose
// syncs succeed without touching the disk.
func newTestSubmissions(t *testing.T) (*store.Handle, *submissions) {
	t.Helper()
	h := openTestStore(t)
	insertUser(t, h, userSpec{admin: true, createdAt: 1})
	s := newSubmissions(h, slog.New(slog.DiscardHandler), t.TempDir(),
		func() time.Time { return time.UnixMicro(testNow) },
		func(string) (uint64, error) { return plentyFree, nil })
	s.sync = func(string) error { return nil }
	return h, s
}

// member inserts a plain member and returns its id as the actor's
// UserID carries it.
func member(t *testing.T, h *store.Handle) identity.UserID {
	t.Helper()
	return identity.UserID(insertUser(t, h, userSpec{createdAt: 10}))
}

// testLimits are limits no fixture meets unless it sets one lower.
func testLimits() limits {
	return limits{
		policy: gate.Policy{
			MaxAlbumBytes: 1 << 40,
			MaxFiles:      500,
			MaxPathDepth:  8,
			Allowlist:     []string{"flac"},
		},
		sessionsPerMember: 100,
	}
}

// album is a loose-files album manifest of the given file sizes, named
// 01.flac, 02.flac, and so on.
func album(sizes ...int64) newSubmission {
	m := gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles, Source: "bought"}
	for i, size := range sizes {
		m.Files = append(m.Files, gate.ManifestFile{Path: string(rune('0'+i/10)) + string(rune('1'+i%10)) + ".flac", Size: size})
	}
	return newSubmission{manifest: m}
}

// tracks is album's separate-tracks twin.
func tracks(sizes ...int64) newSubmission {
	n := album(sizes...)
	n.manifest.Unit = gate.UnitTrack
	return n
}

// zipOf is a one-archive submission of the given archive size.
func zipOf(unit gate.Unit, size int64) newSubmission {
	return newSubmission{manifest: gate.Manifest{
		Unit: unit, Container: gate.ContainerZip, Source: "cd",
		Files: []gate.ManifestFile{{Path: "music.zip", Size: size}},
	}}
}

// mustCreate creates n for uploader and fails the test on any error.
func mustCreate(t *testing.T, s *submissions, uploader identity.UserID, n newSubmission, lim limits) created {
	t.Helper()
	c, err := s.create(deadline(t), uploader, n, lim)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return c
}

// receiveAll runs receipt for every file of c.
func receiveAll(t *testing.T, s *submissions, c created) {
	t.Helper()
	for _, f := range c.files {
		if err := s.receive(deadline(t), f.id); err != nil {
			t.Fatalf("receive %s: %v", f.path, err)
		}
	}
}

// passing is a verdict for every loose file of c with no refusal.
func passing(c created, to state) verdict {
	v := verdict{to: to}
	for _, f := range c.files {
		v.files = append(v.files, fileVerdict{id: f.id})
	}
	return v
}

// inState creates a loose-files album for uploader and moves it to st
// through the transition functions themselves.
func inState(t *testing.T, s *submissions, uploader identity.UserID, st state, sizes ...int64) created {
	t.Helper()
	ctx := deadline(t)
	c := mustCreate(t, s, uploader, album(sizes...), testLimits())
	var err error
	switch st {
	case stateUploading:
		return c
	case stateCancelled:
		err = s.cancel(ctx, uploader, c.id)
	case stateExpired:
		err = s.expire(ctx, c.id, "idle past the resume window", stateUploading)
	default:
		receiveAll(t, s, c)
		switch st {
		case stateChecking:
		case stateMatching, stateCheckUnfinished:
			err = s.commitVerdict(ctx, c.id, passing(c, st))
		case stateAccepted:
			// An album moves to accepted only through accept, which reads
			// the target collection; this one holds no twin.
			err = s.commitVerdict(ctx, c.id, passing(c, stateAccepted))
		case stateRejected:
			v := passing(c, stateRejected)
			v.files[0].result.Verdict = reject.Corrupt
			err = s.commitVerdict(ctx, c.id, v)
		default:
			t.Fatalf("no fixture path to %q", st)
		}
	}
	if err != nil {
		t.Fatalf("move to %q: %v", st, err)
	}
	if got := stateOf(t, s.store, c.id); got != string(st) {
		t.Fatalf("fixture reached %q, want %q", got, st)
	}
	return c
}

func stateOf(t *testing.T, h *store.Handle, id []byte) string {
	t.Helper()
	var st string
	if err := h.QueryRow(deadline(t), `SELECT state FROM submissions WHERE id = ?`, id).Scan(&st); err != nil {
		t.Fatalf("read state: %v", err)
	}
	return st
}

// eventsOf is a submission's event kinds in seq order.
func eventsOf(t *testing.T, h *store.Handle, id []byte) []string {
	t.Helper()
	rows, err := h.Query(deadline(t), `SELECT kind FROM submission_events WHERE submission_id = ? ORDER BY seq`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var kinds []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, k)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return kinds
}

func TestStatesMatchTheSharedVocabulary(t *testing.T) {
	var got []string
	for _, st := range states {
		got = append(got, string(st))
	}
	if !slices.Equal(got, journeyStates) {
		t.Errorf("states = %q, want %q", got, journeyStates)
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []byte
	}{
		{name: "32 lowercase hex", in: "000102030405060708090a0b0c0d0eff",
			want: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 255}},
		{name: "uppercase hex", in: "000102030405060708090A0B0C0D0EFF"},
		{name: "31 characters", in: "000102030405060708090a0b0c0d0ef"},
		{name: "33 characters", in: "000102030405060708090a0b0c0d0eff0"},
		{name: "not hex", in: "000102030405060708090a0b0c0d0egg"},
		{name: "the zero id", in: "00000000000000000000000000000000"},
		{name: "a slash inside", in: "0001020304050607/8090a0b0c0d0eff"},
		{name: "empty", in: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseID(tt.in)
			if ok != (tt.want != nil) || !bytes.Equal(got, tt.want) {
				t.Errorf("parseID(%q) = %x, %v; want %x, %v", tt.in, got, ok, tt.want, tt.want != nil)
			}
		})
	}
}

// TestRetryAndCancelFromEveryState is AC 26's table for the states this
// pass emits: retry leaves check-unfinished alone, cancel leaves
// uploading and check-unfinished, and every other cell is a conflict
// that changes nothing.
func TestRetryAndCancelFromEveryState(t *testing.T) {
	type action struct {
		name string
		run  func(s *submissions, ctx context.Context, uploader identity.UserID, id []byte) error
		// moves maps each state the action leaves to where it lands.
		moves map[state]state
	}
	actions := []action{
		{name: "retry", run: (*submissions).retry, moves: map[state]state{stateCheckUnfinished: stateChecking}},
		{name: "cancel", run: (*submissions).cancel, moves: map[state]state{
			stateUploading: stateCancelled, stateCheckUnfinished: stateCancelled,
		}},
	}
	for _, a := range actions {
		for _, from := range states {
			t.Run(a.name+" from "+string(from), func(t *testing.T) {
				h, s := newTestSubmissions(t)
				uploader := member(t, h)
				c := inState(t, s, uploader, from, 10, 20)
				before := eventsOf(t, h, c.id)

				err := a.run(s, deadline(t), uploader, c.id)

				to, moves := a.moves[from]
				if !moves {
					if !errors.Is(err, errWrongState) {
						t.Fatalf("error = %v, want errWrongState", err)
					}
					if got := stateOf(t, h, c.id); got != string(from) {
						t.Errorf("state = %q after a refused %s, want %q", got, a.name, from)
					}
					if got := eventsOf(t, h, c.id); !slices.Equal(got, before) {
						t.Errorf("events = %q after a refused %s, want %q", got, a.name, before)
					}
					assertStatesMatchEvents(t, h)
					return
				}
				if err != nil {
					t.Fatalf("%s: %v", a.name, err)
				}
				if got := stateOf(t, h, c.id); got != string(to) {
					t.Errorf("state = %q, want %q", got, to)
				}
				if got := eventsOf(t, h, c.id); !slices.Equal(got, append(before, string(to))) {
					t.Errorf("events = %q, want %q then %q", got, before, to)
				}
				assertStatesMatchEvents(t, h)
			})
		}
	}
}

func TestRetryPokesTheWorker(t *testing.T) {
	h, s := newTestSubmissions(t)
	uploader := member(t, h)
	c := inState(t, s, uploader, stateCheckUnfinished, 10)
	drain(s)

	if err := s.retry(deadline(t), uploader, c.id); err != nil {
		t.Fatal(err)
	}
	if len(s.wake) != 1 {
		t.Errorf("wake holds %d pokes after a retry, want 1", len(s.wake))
	}
	assertStatesMatchEvents(t, h)
}

// TestActionsOnAnotherMembersSubmission holds that one member's id is
// unknown to another: neither action finds it, and nothing moves.
func TestActionsOnAnotherMembersSubmission(t *testing.T) {
	h, s := newTestSubmissions(t)
	owner, other := member(t, h), member(t, h)
	c := inState(t, s, owner, stateCheckUnfinished, 10)

	for name, run := range map[string]func(context.Context, identity.UserID, []byte) error{
		"retry": s.retry, "cancel": s.cancel,
	} {
		if err := run(deadline(t), other, c.id); !errors.Is(err, errNotFound) {
			t.Errorf("%s by another member: error = %v, want errNotFound", name, err)
		}
		if err := run(deadline(t), owner, []byte("0123456789abcdef")); !errors.Is(err, errNotFound) {
			t.Errorf("%s of an unknown id: error = %v, want errNotFound", name, err)
		}
	}
	if got := stateOf(t, h, c.id); got != string(stateCheckUnfinished) {
		t.Errorf("state = %q, want it unchanged", got)
	}
	assertStatesMatchEvents(t, h)
}

// TestFailedCompareAndSetDiscardsTheResult commits a verdict on a
// submission the member cancelled while the check ran: the move fails,
// so neither the file verdicts nor the warning it carried land.
func TestFailedCompareAndSetDiscardsTheResult(t *testing.T) {
	h, s := newTestSubmissions(t)
	uploader := member(t, h)
	c := inState(t, s, uploader, stateCheckUnfinished, 10, 20)
	if err := s.cancel(deadline(t), uploader, c.id); err != nil {
		t.Fatal(err)
	}

	v := passing(c, stateRejected)
	v.files[0].result.Verdict = reject.Corrupt
	v.warnings = []gate.FileReason{{Reason: reject.Of(reject.NoCoverArt)}}
	if err := s.commitVerdict(deadline(t), c.id, v); !errors.Is(err, errWrongState) {
		t.Fatalf("error = %v, want errWrongState", err)
	}

	var verdicts, warnings int
	if err := h.QueryRow(deadline(t), `SELECT count(verdict) FROM submission_files WHERE submission_id = ?`, c.id).Scan(&verdicts); err != nil {
		t.Fatal(err)
	}
	if err := h.QueryRow(deadline(t), `SELECT count(*) FROM submission_warnings WHERE submission_id = ?`, c.id).Scan(&warnings); err != nil {
		t.Fatal(err)
	}
	if verdicts != 0 || warnings != 0 {
		t.Errorf("%d file verdicts and %d warnings landed from a failed move, want none", verdicts, warnings)
	}
	if got := stateOf(t, h, c.id); got != string(stateCancelled) {
		t.Errorf("state = %q, want cancelled", got)
	}
	assertStatesMatchEvents(t, h)
}

func TestAppendEventRefusesAKindOutsideTheVocabulary(t *testing.T) {
	h, s := newTestSubmissions(t)
	c := mustCreate(t, s, member(t, h), album(10), testLimits())

	for _, kind := range []string{"checked", "", "Uploading"} {
		err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
			return appendEvent(ctx, tx, c.id, testNow, kind, "")
		})
		if err == nil {
			t.Errorf("appendEvent(%q) committed, want a refusal", kind)
		}
	}
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		return appendEvent(ctx, tx, c.id, testNow, eventStalled, "")
	})
	if err != nil {
		t.Errorf("appendEvent(stalled): %v", err)
	}
	if got, want := eventsOf(t, h, c.id), []string{"uploading", "stalled"}; !slices.Equal(got, want) {
		t.Errorf("events = %q, want %q", got, want)
	}
	assertStatesMatchEvents(t, h)
}

// TestExpire ends a submission from each state its caller names, with
// the reason on the event, and refuses one in any other state.
func TestExpire(t *testing.T) {
	h, s := newTestSubmissions(t)
	m := member(t, h)
	held := inState(t, s, m, stateCheckUnfinished, 10)
	waiting := inState(t, s, m, stateMatching, 10)

	if err := s.expire(deadline(t), held.id, "held past the hold", stateCheckUnfinished); err != nil {
		t.Fatal(err)
	}
	if got, reason := stateOf(t, h, held.id), lastReason(t, h, held.id); got != "expired" || reason != "held past the hold" {
		t.Errorf("held = %q (%q), want expired with its reason", got, reason)
	}
	if err := s.expire(deadline(t), waiting.id, "idle", stateUploading, stateCheckUnfinished); !errors.Is(err, errWrongState) {
		t.Errorf("expire from matching: error = %v, want errWrongState", err)
	}
	assertStatesMatchEvents(t, h)
}

// drain empties the wake channel so a test can see its own poke.
func drain(s *submissions) {
	for len(s.wake) > 0 {
		<-s.wake
	}
}
