package intake

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// journeyStates is the closed set of states this pass emits, written out
// here as the expected value: the production vocabulary lives once, in
// submission.go, and a test there checks it against this literal.
// "stalled" is an event kind and never a state, so it is absent on
// purpose.
var journeyStates = []string{
	"uploading", "checking", "check-unfinished", "matching",
	"accepted", "rejected", "cancelled", "expired",
}

// stateEventMismatches reads every submission and reports each whose
// state is not the kind of its latest state event by seq, with checked
// the number of submissions it read. An event whose kind is outside
// journeyStates (stalled) is not a state event and never counts.
func stateEventMismatches(ctx context.Context, h *store.Handle) (mismatches []string, checked int, err error) {
	rows, err := h.Query(ctx, `SELECT s.id, s.state, e.kind
		FROM submissions s LEFT JOIN submission_events e ON e.submission_id = s.id
		ORDER BY s.id, e.seq`)
	if err != nil {
		return nil, 0, fmt.Errorf("read submissions and events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	state := map[string]string{}
	latest := map[string]string{}
	for rows.Next() {
		var id []byte
		var st string
		var kind *string
		if err := rows.Scan(&id, &st, &kind); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		key := fmt.Sprintf("%x", id)
		state[key] = st
		if kind != nil && slices.Contains(journeyStates, *kind) {
			latest[key] = *kind
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read events: %w", err)
	}
	for key, st := range state {
		if latest[key] != st {
			mismatches = append(mismatches, fmt.Sprintf("submission %s: state %q, latest state event %q", key, st, latest[key]))
		}
	}
	slices.Sort(mismatches)
	return mismatches, len(state), nil
}

// assertStatesMatchEvents fails t unless every submission's state equals
// the kind of its latest state event by seq, and fails on a store with no
// submission, which would pass vacuously. Every test that moves a
// submission between states ends with it.
func assertStatesMatchEvents(t *testing.T, h *store.Handle) {
	t.Helper()
	mismatches, checked, err := stateEventMismatches(deadline(t), h)
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no submission to check: the assertion would pass on an empty store")
	}
	for _, m := range mismatches {
		t.Error(m)
	}
}

// seedSubmission inserts a submission in state with events appended in
// the order given at seq 1, 2, and so on, all at one instant so only seq
// orders them. insertOrder permutes the insert order of the events (by
// zero-based index) so a read keyed to row order fails.
func seedSubmission(t *testing.T, h *store.Handle, state string, kinds []string, insertOrder []int) []byte {
	t.Helper()
	owner, coll := seedOwnerAndCollection(t, h)
	id := store.NewID()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO submissions
			(id, uploader, collection_id, verdict_unit, container, state, source, declared_bytes, reserved_bytes, created_at, updated_at)
			VALUES (?, ?, ?, 'album', 'files', ?, 'bandcamp', 0, 0, 1, 1)`, id, owner, coll, state); err != nil {
			return err
		}
		for _, i := range insertOrder {
			if _, err := tx.Exec(ctx, `INSERT INTO submission_events (id, submission_id, seq, at, kind)
				VALUES (?, ?, ?, 1700000000000000, ?)`, store.NewID(), id, i+1, kinds[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed submission in %q: %v", state, err)
	}
	return id
}

// seedOwnerAndCollection inserts one user and one collection for a
// submission to reference, and returns their ids.
func seedOwnerAndCollection(t *testing.T, h *store.Handle) (owner, collection []byte) {
	t.Helper()
	owner = insertUser(t, h, userSpec{admin: true, createdAt: 1})
	collection = store.NewID()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer)
			VALUES (?, 'Seed', ?, ?, 'group', 'server')`, collection, []byte(fmt.Sprintf("seed-%x", collection)), owner)
		return err
	})
	if err != nil {
		t.Fatalf("seed collection: %v", err)
	}
	return owner, collection
}

func inOrder(n int) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	return order
}

// TestStateEventAssertion proves the assertion can fail: each row seeds
// one fixture state, and wantMismatch names whether the helper must
// report it.
func TestStateEventAssertion(t *testing.T) {
	tests := []struct {
		name         string
		state        string
		kinds        []string
		insertOrder  []int
		wantMismatch string
	}{
		{
			name:  "state equals the latest state event",
			state: "matching", kinds: []string{"uploading", "checking", "matching"}, insertOrder: inOrder(3),
		},
		{
			name:  "a stalled event after the state event is not a state",
			state: "matching", kinds: []string{"uploading", "checking", "matching", "stalled"}, insertOrder: inOrder(4),
		},
		{
			name:  "a stalled event after an earlier state in checking",
			state: "checking", kinds: []string{"uploading", "checking", "stalled"}, insertOrder: inOrder(3),
		},
		{
			name:  "seq orders the events, not the row order",
			state: "checking", kinds: []string{"uploading", "checking"}, insertOrder: []int{1, 0},
		},
		{
			name:  "state written out of step with its latest event",
			state: "accepted", kinds: []string{"uploading", "checking", "matching"}, insertOrder: inOrder(3),
			wantMismatch: `state "accepted", latest state event "matching"`,
		},
		{
			name:  "a state the latest event never reached under a stalled tail",
			state: "checking", kinds: []string{"uploading", "checking", "matching", "stalled"}, insertOrder: inOrder(4),
			wantMismatch: `state "checking", latest state event "matching"`,
		},
		{
			name:  "no state event at all",
			state: "uploading", kinds: []string{"stalled"}, insertOrder: inOrder(1),
			wantMismatch: `state "uploading", latest state event ""`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := openTestStore(t)
			seedSubmission(t, h, tt.state, tt.kinds, tt.insertOrder)

			mismatches, checked, err := stateEventMismatches(deadline(t), h)
			if err != nil {
				t.Fatal(err)
			}
			if checked != 1 {
				t.Fatalf("checked %d submissions, want 1", checked)
			}
			if tt.wantMismatch == "" {
				if len(mismatches) != 0 {
					t.Errorf("mismatches = %q, want none", mismatches)
				}
				assertStatesMatchEvents(t, h)
				return
			}
			if len(mismatches) != 1 || !strings.Contains(mismatches[0], tt.wantMismatch) {
				t.Errorf("mismatches = %q, want one containing %q", mismatches, tt.wantMismatch)
			}
		})
	}
}

// deadline returns a context that ends with the test and carries the
// deadline store.Handle.Update requires.
func deadline(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}
