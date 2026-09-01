// api.go implements the polling read side of a contributor's upload:
// GET /api/journey/{id} and GET /api/history. It is a separate handler
// from Engine's own tus/session mux (upload.go) because it depends on
// internal/journey and internal/ledger, concerns Engine's write-side
// upload-completion logic has no need of; cmd/dubplate/serve.go mounts
// both under one top-level mux.
package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/journey"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/notify"
)

// JourneyDeps holds [JourneyAPI]'s dependencies beyond Engine's own live
// session store: built once by cmd/dubplate/serve.go and passed to
// [NewJourneyAPI].
type JourneyDeps struct {
	Ledger             *ledger.Ledger
	Navidrome          *journey.NavidromeScanClient
	Mailer             *notify.Mailer
	LockPath           string
	AdminEmail         string
	StallMinutes       int
	PublicNavidromeURL string

	// Now stands in for time.Now; the zero value uses the real clock.
	// Tests override it the same way [Options.Now] overrides Engine's.
	Now func() time.Time
}

// JourneyAPI serves GET /api/journey/{id} and GET /api/history: the
// polling contract [journey.Resolve] answers by composing Engine's live
// session store with deps.Ledger. The zero JourneyAPI is not usable;
// construct with [NewJourneyAPI].
type JourneyAPI struct {
	engine *Engine
	deps   JourneyDeps
}

// NewJourneyAPI builds a JourneyAPI over e's live session store and
// deps.
func NewJourneyAPI(e *Engine, deps JourneyDeps) *JourneyAPI {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	return &JourneyAPI{engine: e, deps: deps}
}

// Handler returns JourneyAPI's mux: GET /api/journey/{id} and GET
// /api/history.
func (j *JourneyAPI) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/journey/{id}", j.handleJourney)
	mux.HandleFunc("GET /api/history", j.handleHistory)
	return mux
}

// handleJourney serves GET /api/journey/{id}: the 5s polling contract a
// contributor's browser reads while their upload works its way to the
// library. A session that exists but belongs to someone else answers
// 404, identically to an unknown ID -- no existence oracle.
func (j *JourneyAPI) handleJourney(w http.ResponseWriter, r *http.Request) {
	email, ok := EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	if !validSessionID(id) {
		http.NotFound(w, r)
		return
	}

	now := j.deps.Now()
	if err := j.advance(r.Context(), id, now); err != nil {
		slog.Error("journey: advance session", "session_id", id, "err", err)
	}

	lb := j.buildLive(email)
	authorized, err := j.authorize(id, email, lb)
	if err != nil {
		http.Error(w, "read journey", http.StatusInternalServerError)
		return
	}
	if !authorized {
		http.NotFound(w, r)
		return
	}

	jrn, err := journey.Resolve(id, j.deps.Ledger, lb.set)
	if err != nil {
		if errors.Is(err, journey.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "resolve journey", http.StatusInternalServerError)
		return
	}
	j.finish(&jrn, lb)
	writeJSON(w, http.StatusOK, jrn)
}

// handleHistory serves GET /api/history: every session the requester's
// verified email has ever touched, terminal sessions checkable from
// outside the flow alongside any still in flight.
func (j *JourneyAPI) handleHistory(w http.ResponseWriter, r *http.Request) {
	email, ok := EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	ledgerEvents, err := j.deps.Ledger.History(email)
	if err != nil {
		http.Error(w, "read history", http.StatusInternalServerError)
		return
	}

	now := j.deps.Now()
	lb := j.buildLive(email)

	ids := make([]string, 0, len(ledgerEvents)+len(lb.set))
	seen := make(map[string]bool, cap(ids))
	for _, e := range ledgerEvents {
		if seen[e.SessionID] {
			continue
		}
		seen[e.SessionID] = true
		ids = append(ids, e.SessionID)
	}
	for id := range lb.set {
		if seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}

	out := make([]journey.Journey, 0, len(ids))
	for _, id := range ids {
		if err := j.advance(r.Context(), id, now); err != nil {
			slog.Error("journey: advance session", "session_id", id, "err", err)
		}
		jrn, err := journey.Resolve(id, j.deps.Ledger, lb.set)
		if err != nil {
			slog.Error("journey: resolve for history", "session_id", id, "err", err)
			continue
		}
		j.finish(&jrn, lb)
		out = append(out, jrn)
	}

	writeJSON(w, http.StatusOK, out)
}

// finish fills in the two Journey fields [journey.Resolve] cannot on its
// own: the queued copy variant, which needs the shared import lock and
// the requester's other live sessions, and the in-library deep link,
// which needs config's PublicNavidromeURL -- neither is ledger or live
// state, so Resolve is never handed them.
func (j *JourneyAPI) finish(jrn *journey.Journey, lb liveBuild) {
	switch jrn.State {
	case ledger.StateQueued:
		jrn.Detail = lb.details[jrn.SessionID]
	case ledger.StateInLibrary:
		jrn.LibraryLink = journey.BuildLibraryLink(j.deps.PublicNavidromeURL, jrn.Album)
	}
}

// advance brings sessionID's ledger history up to date with the two
// externally observable facts [journey.Resolve] itself never checks: a
// scan-confirmed promotion to in-library, and an overdue stall. Both are
// margin, not a contributor-facing failure -- an error here is logged
// by the caller, never turned into a failed poll.
func (j *JourneyAPI) advance(ctx context.Context, sessionID string, now time.Time) error {
	events, err := j.deps.Ledger.Session(sessionID)
	if err != nil {
		return fmt.Errorf("read session %s history: %w", sessionID, err)
	}
	if len(events) == 0 {
		// Still live-only: the importer has not appended anything yet,
		// so there is nothing in the ledger to advance.
		return nil
	}

	if err := journey.AdvanceInLibrary(ctx, j.deps.Ledger, j.deps.Navidrome, events, now); err != nil {
		return err
	}

	// Re-read: a fresh in-library event above must be reflected here, or
	// AdvanceStalled would judge staleness against the pre-promotion
	// events slice and append a spurious stalled event right behind a
	// session that just settled.
	events, err = j.deps.Ledger.Session(sessionID)
	if err != nil {
		return fmt.Errorf("read session %s history: %w", sessionID, err)
	}
	if err := journey.AdvanceStalled(ctx, j.deps.Ledger, j.deps.Mailer.Send, j.deps.AdminEmail, j.deps.StallMinutes, events, now); err != nil {
		return err
	}
	return nil
}

// authorize reports whether email owns sessionID, checked against both
// sources [journey.Resolve] itself reads: a live session in lb, or a
// ledger event. Neither alone is authoritative during the brief window
// after upload completes but before the importer's first ledger event,
// so both are checked; a session in neither answers false, the same
// answer a session owned by someone else gets, so a prober can never
// tell "doesn't exist" from "isn't yours" -- B3's no-existence-oracle
// rule, extended to this read path.
func (j *JourneyAPI) authorize(sessionID, email string, lb liveBuild) (bool, error) {
	if _, ok := lb.set[sessionID]; ok {
		return true, nil
	}
	events, err := j.deps.Ledger.Session(sessionID)
	if err != nil {
		return false, fmt.Errorf("read session %s history: %w", sessionID, err)
	}
	if len(events) == 0 {
		return false, nil
	}
	return events[0].Email == email, nil
}

// liveBuild is [JourneyAPI]'s live half for one request: every session
// email currently has in Engine's store, plus the queued Detail text for
// any of them Resolve will report as [ledger.StateQueued] -- computed
// once per request in [JourneyAPI.buildLive] rather than a second time
// in [JourneyAPI.finish], since both need the same lock probe and
// sibling-session scan.
type liveBuild struct {
	set     journey.LiveSet
	details map[string]string
}

// buildLive constructs email's live half: every session Engine's store
// still holds for them, each mapped to its [journey.LiveSession] view.
func (j *JourneyAPI) buildLive(email string) liveBuild {
	sessions := j.engine.sessionsForEmail(email)
	lb := liveBuild{set: make(journey.LiveSet, len(sessions)), details: make(map[string]string)}
	for _, sess := range sessions {
		ls, detail := j.liveSession(sess)
		lb.set[sess.ID] = ls
		if detail != "" {
			lb.details[sess.ID] = detail
		}
	}
	return lb
}

// liveSession converts sess's current bookkeeping into the
// [journey.LiveSession] view Resolve reads, plus a queued Detail string
// when sess should read as [ledger.StateQueued]: an unfinished earlier
// session for the same email, from [JourneyAPI.hasEarlierUnsettled],
// takes priority over the shared import lock ([lock.Held], the queued
// probe N10 authorized), since it is the more specific, more actionable
// fact for the contributor to see.
func (j *JourneyAPI) liveSession(sess *Session) (journey.LiveSession, string) {
	completed, doneUploads, outcome, verdicts := sess.snapshot()

	ls := journey.LiveSession{SessionID: sess.ID, Email: sess.Email, Files: verdicts}
	switch {
	case !completed:
		ls.State = ledger.StateUploading
		if doneUploads >= sess.ExpectedFiles {
			ls.State = ledger.StateChecking
		}
		return ls, ""
	case !outcome.Verdict.OK:
		// Failed serve's own intake check and was never promoted: the
		// synchronous upload response already told the contributor why,
		// but a poll that lands after the tab reloads sees the same
		// per-file reasons here.
		ls.State = ledger.StateRejected
		return ls, ""
	}

	ownEarlier := j.hasEarlierUnsettled(sess)
	held, err := lock.Held(j.deps.LockPath)
	if err != nil {
		slog.Warn("journey: probe import lock", "err", err)
	}
	if !ownEarlier && !held {
		ls.State = ledger.StateUploaded
		return ls, ""
	}
	ls.State = ledger.StateQueued
	return ls, journey.QueuedDetail(ownEarlier, held)
}

// hasEarlierUnsettled reports whether sess's owner has another session
// in Engine's live store, created before sess and not yet settled in
// the ledger (no [journey.Terminal] event) -- the "waiting on your other
// upload" queued variant.
func (j *JourneyAPI) hasEarlierUnsettled(sess *Session) bool {
	for _, other := range j.engine.sessionsForEmail(sess.Email) {
		if other.ID == sess.ID || !other.createdAt.Before(sess.createdAt) {
			continue
		}
		events, err := j.deps.Ledger.Session(other.ID)
		if err != nil {
			continue
		}
		if len(events) > 0 && journey.Terminal(events[0].State) {
			continue
		}
		return true
	}
	return false
}

// snapshot returns sess's completion bookkeeping in one locked read:
// whether it has completed, how many uploads have finished, its final
// outcome (zero until completed), and a copy of its accumulated
// verdicts -- what [JourneyAPI.liveSession] needs to build a
// [journey.LiveSession] without racing [Session.uploadDone] or
// [Session.setOutcome].
func (sess *Session) snapshot() (completed bool, doneUploads int, outcome sessionOutcome, verdicts []intake.FileVerdict) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	out := make([]intake.FileVerdict, len(sess.verdicts))
	copy(out, sess.verdicts)
	return sess.completed, sess.doneUploads, sess.outcome, out
}

// sessionsForEmail returns every session in e's live store belonging to
// owner. Order is unspecified.
func (e *Engine) sessionsForEmail(owner string) []*Session {
	e.sessions.mu.Lock()
	defer e.sessions.mu.Unlock()
	var out []*Session
	for _, sess := range e.sessions.sessions {
		if sess.Email == owner {
			out = append(out, sess)
		}
	}
	return out
}
