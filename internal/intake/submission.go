package intake

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
)

// A state is a submission's place in its journey, as submissions.state
// stores it.
type state string

const (
	stateUploading       state = "uploading"
	stateChecking        state = "checking"
	stateCheckUnfinished state = "check-unfinished"
	stateMatching        state = "matching"
	stateAccepted        state = "accepted"
	stateRejected        state = "rejected"
	stateCancelled       state = "cancelled"
	stateExpired         state = "expired"
)

// states is every state this pass emits, in journey order.
var states = []state{
	stateUploading, stateChecking, stateCheckUnfinished, stateMatching,
	stateAccepted, stateRejected, stateCancelled, stateExpired,
}

// terminal reports whether st ends a submission's journey: no
// transition leaves it, and cleanup takes all its bytes.
func (st state) terminal() bool {
	return st == stateRejected || st == stateCancelled || st == stateExpired
}

// eventStalled is the one event kind that is not a state: a submission
// past the stall threshold keeps its state, so recovery needs no extra
// transition.
const eventStalled = "stalled"

// openSubmission is the SQL test for a submission that still holds its
// reservation: every state but the three terminal ones.
const openSubmission = `state NOT IN ('rejected', 'cancelled', 'expired')`

var (
	errNotFound   = errors.New("no such submission")
	errWrongState = errors.New("submission is not in a state this action leaves")
)

// writeWait bounds each intake write's wait for the store's write
// connection.
const writeWait = 10 * time.Second

// submissions writes intake's rows. Every write is one store Update
// stamped from now, and a write that moves a submission to checking
// pokes wake.
type submissions struct {
	store     *store.Handle
	log       *slog.Logger
	staging   string
	now       func() time.Time
	freeBytes func(dir string) (uint64, error)
	// sync flushes one file or directory to disk.
	sync func(path string) error
	// wake holds at most one poke. The store is the worker's queue, so a
	// poke only says to look, and one already waiting says the same.
	wake chan struct{}
}

func newSubmissions(h *store.Handle, log *slog.Logger, staging string, now func() time.Time,
	freeBytes func(dir string) (uint64, error)) *submissions {
	return &submissions{
		store: h, log: log, staging: staging, now: now, freeBytes: freeBytes,
		sync: syncPath, wake: make(chan struct{}, 1),
	}
}

// update runs fn in one store Update, handing it the instant every row
// the write stamps carries.
func (s *submissions) update(ctx context.Context, fn func(ctx context.Context, tx *store.Tx, at int64) error) error {
	ctx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	at := s.now().UnixMicro()
	return s.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		return fn(ctx, tx, at)
	})
}

// poke tells the worker to look at the queue without blocking.
func (s *submissions) poke() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// parseID parses a submission or file id as the wire carries it: exactly
// 32 lowercase hex characters, never the zero id. [hex.DecodeString]
// accepts uppercase, so s must also equal its own re-encoding.
func parseID(s string) ([]byte, bool) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 16 || hex.EncodeToString(b) != s || [16]byte(b) == [16]byte{} {
		return nil, false
	}
	return b, true
}

// newSubmission is a member's request to create a submission.
type newSubmission struct {
	manifest gate.Manifest
	// sourceNote and storeURL are optional, and empty stores NULL. The
	// seam has already refused a store URL that is not https.
	sourceNote, storeURL string
}

// created is a new submission's id and its files' ids, which the member
// uploads each file to.
type created struct {
	id    []byte
	files []createdFile
}

type createdFile struct {
	id   []byte
	path string
}

// manifestRefusal is a create refused by its manifest, carrying every
// reason found.
type manifestRefusal struct {
	refusal gate.Refusal
}

func (e *manifestRefusal) Error() string {
	return fmt.Sprintf("manifest refused: %d submission reasons, %d files", len(e.refusal.Submission), len(e.refusal.Files))
}

// create checks a member's manifest, then the create limits, and writes
// the submission with its reservation in one Update, so two creates can
// never each pass a limit that together they exceed. The manifest's
// reasons come back as a [*manifestRefusal], the first failing limit as
// a [*limitRefusal], and a malformed manifest as the gate's error.
//
// An album takes the main route and separate tracks the unlisted one.
// Free space is read before the Update, the last thing create does
// outside it, so no syscall runs under the write lock.
func (s *submissions) create(ctx context.Context, uploader identity.UserID, n newSubmission, lim limits) (created, error) {
	route := RouteDefault
	if n.manifest.Unit == gate.UnitTrack {
		route = RouteUnlisted
	}
	free := s.readFree()
	var c created
	err := s.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		coll, err := RouteCollection(ctx, tx, s.log, route)
		if err != nil {
			return err
		}
		policy := lim.policy
		if policy.Allowlist, err = effectiveAllowlist(ctx, tx, coll, policy.Allowlist); err != nil {
			return err
		}
		refusal, err := gate.CheckManifest(n.manifest, policy)
		if err != nil {
			return err
		}
		if len(refusal.Submission) > 0 || len(refusal.Files) > 0 {
			return &manifestRefusal{refusal: refusal}
		}

		// A clean manifest's sizes are non-negative and sum within the
		// album cap, so the total cannot overflow.
		var total int64
		for _, f := range n.manifest.Files {
			total += f.Size
		}
		if err := checkCreate(ctx, tx, uploader, total, lim, free); err != nil {
			return err
		}

		c = created{id: store.NewID()}
		_, err = tx.Exec(ctx, `INSERT INTO submissions (id, uploader, collection_id, verdict_unit, container, state,
			source, source_note, store_url, declared_bytes, reserved_bytes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.id, uploader[:], coll, string(n.manifest.Unit), string(n.manifest.Container), stateUploading,
			n.manifest.Source, nullable(n.sourceNote), nullable(n.storeURL), total, total, at, at)
		if err != nil {
			return fmt.Errorf("write submission: %w", err)
		}
		for _, f := range n.manifest.Files {
			file := createdFile{id: store.NewID(), path: f.Path}
			_, err := tx.Exec(ctx, `INSERT INTO submission_files (id, submission_id, rel_path, size, role)
				VALUES (?, ?, ?, ?, ?)`, file.id, c.id, []byte(f.Path), f.Size, roleOf(n.manifest.Container, f.Path))
			if err != nil {
				return fmt.Errorf("write submission file: %w", err)
			}
			c.files = append(c.files, file)
		}
		return appendEvent(ctx, tx, c.id, at, string(stateUploading), "")
	})
	if err != nil {
		return created{}, err
	}
	return c, nil
}

// roleOf is a declared file's role, which the server derives from the
// extension and never takes from a request.
func roleOf(c gate.Container, path string) string {
	switch {
	case c == gate.ContainerZip:
		return "archive"
	case gate.IsAudio(path):
		return "audio"
	}
	return "extra"
}

// retry moves the uploader's check-unfinished submission back to
// checking and pokes the worker.
func (s *submissions) retry(ctx context.Context, uploader identity.UserID, id []byte) error {
	err := s.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		if err := owned(ctx, tx, uploader, id); err != nil {
			return err
		}
		return move(ctx, tx, id, at, stateChecking, "", stateCheckUnfinished)
	})
	if err != nil {
		return err
	}
	s.poke()
	return nil
}

// cancel ends the uploader's submission while it uploads or waits on a
// check that could not run.
func (s *submissions) cancel(ctx context.Context, uploader identity.UserID, id []byte) error {
	return s.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		if err := owned(ctx, tx, uploader, id); err != nil {
			return err
		}
		return move(ctx, tx, id, at, stateCancelled, "", stateUploading, stateCheckUnfinished)
	})
}

// expire ends a submission in one of from, recording reason: the
// sweep's windows, or the startup pass finding its staged bytes gone.
func (s *submissions) expire(ctx context.Context, id []byte, reason string, from ...state) error {
	return s.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		return move(ctx, tx, id, at, stateExpired, reason, from...)
	})
}

// owned returns errNotFound unless uploader uploaded submission id, so
// another member's id reads as an unknown one.
func owned(ctx context.Context, tx *store.Tx, uploader identity.UserID, id []byte) error {
	var n int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM submissions WHERE id = ? AND uploader = ?`, id, uploader[:]).Scan(&n); err != nil {
		return fmt.Errorf("read submission owner: %w", err)
	}
	if n == 0 {
		return errNotFound
	}
	return nil
}

// move compare-and-sets submission id from one of from to to and appends
// to's event at the next seq. It returns errNotFound for no such
// submission and errWrongState for one in no state of from, and the
// caller returns either, so its Update discards everything else it
// wrote.
func move(ctx context.Context, tx *store.Tx, id []byte, at int64, to state, reason string, from ...state) error {
	var current state
	err := tx.QueryRow(ctx, `SELECT state FROM submissions WHERE id = ?`, id).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return errNotFound
	}
	if err != nil {
		return fmt.Errorf("read submission state: %w", err)
	}
	if !slices.Contains(from, current) {
		return fmt.Errorf("%w: %s to %s", errWrongState, current, to)
	}
	if _, err := tx.Exec(ctx, `UPDATE submissions SET state = ?, updated_at = ? WHERE id = ? AND state = ?`,
		to, at, id, current); err != nil {
		return fmt.Errorf("move submission to %s: %w", to, err)
	}
	return appendEvent(ctx, tx, id, at, string(to), reason)
}

// appendEvent writes an event for submission id at the seq after its
// latest. kind must be a state or the stalled event: the column carries
// no CHECK, so a later rung adds a state without a table rebuild, and
// this is where the vocabulary is closed.
func appendEvent(ctx context.Context, tx *store.Tx, id []byte, at int64, kind, reason string) error {
	if !slices.Contains(states, state(kind)) && kind != eventStalled {
		return fmt.Errorf("event kind %q is outside the journey's vocabulary", kind)
	}
	_, err := tx.Exec(ctx, `INSERT INTO submission_events (id, submission_id, seq, at, kind, reason)
		SELECT ?, ?, coalesce(max(seq), 0) + 1, ?, ?, ? FROM submission_events WHERE submission_id = ?`,
		store.NewID(), id, at, kind, nullable(reason), id)
	if err != nil {
		return fmt.Errorf("append %s event: %w", kind, err)
	}
	return nil
}

// nullable stores an empty string as NULL.
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
