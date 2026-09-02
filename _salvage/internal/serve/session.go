package serve

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/glw907/dubplate/internal/intake"
)

// sessionIDAlphabet is the plan-time decision's session-ID charset: the
// standard base32 alphabet, lowercased, so an ID never needs
// case-folding at any ingress that compares it.
const sessionIDAlphabet = "abcdefghijklmnopqrstuvwxyz234567"

// sessionIDPattern is the format [newSessionID] produces and every
// ingress that accepts a session ID (API paths, tus metadata) validates
// against via [validSessionID]: 128 bits of crypto/rand, base32-lowercase
// encoded with no padding, always 26 characters.
var sessionIDPattern = regexp.MustCompile(`^[a-z2-7]{26}$`)

var sessionIDEncoding = base32.NewEncoding(sessionIDAlphabet).WithPadding(base32.NoPadding)

// newSessionID returns a fresh session ID: 128 bits from crypto/rand,
// base32-lowercase encoded.
func newSessionID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return sessionIDEncoding.EncodeToString(raw[:]), nil
}

// validSessionID reports whether id matches the [newSessionID] format.
func validSessionID(id string) bool {
	return sessionIDPattern.MatchString(id)
}

// maxAlbumSlugLen bounds a sanitized album slug well above any real
// album title while keeping the resulting directory name well inside
// filesystem name-length limits.
const maxAlbumSlugLen = 100

// maxExpectedFiles bounds a session's declared expected_files: with no
// cap, a session could reserve an unbounded number of upload slots
// (each pinning a Session.verdicts entry) regardless of the aggregate
// byte cap [Session.reserve] separately enforces. A few hundred is well
// above any real multi-disc box set; intake's own maxZipEntries (2000,
// for a single ZIP's central directory) is a sibling constant, not a
// bound this path needs to match.
const maxExpectedFiles = 300

// ErrInvalidAlbum marks an album name that [slugifyAlbum] could not
// reduce to a non-empty slug.
var ErrInvalidAlbum = errors.New("serve: invalid album name")

// ErrUnknownCollection marks a collection name absent from
// [config.Config.Collections].
var ErrUnknownCollection = errors.New("serve: unknown collection")

// slugifyAlbum normalizes raw to a bounded, filesystem-safe slug before
// any path use. raw is NFC-normalized first, so two differently-encoded
// spellings of the same visual name (a precomposed accent versus a
// combining-mark pair) collapse to the same slug rather than aliasing
// two different directories; every rune outside [a-z0-9] then collapses
// to a single '-' separator. Because '.' and '/' are never in the
// allowed set, no output can contain a ".." traversal segment, a
// leading dot, or an absolute-path prefix -- the empty result for an
// all-punctuation input is the only rejection this function needs.
func slugifyAlbum(raw string) (string, error) {
	normalized := norm.NFC.String(raw)

	var b strings.Builder
	dash := false
	for _, r := range normalized {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
			dash = false
		case b.Len() > 0 && !dash:
			b.WriteByte('-')
			dash = true
		}
	}

	slug := strings.TrimRight(b.String(), "-")
	if len(slug) > maxAlbumSlugLen {
		slug = strings.TrimRight(slug[:maxAlbumSlugLen], "-")
	}
	if slug == "" {
		return "", fmt.Errorf("%w: %q", ErrInvalidAlbum, raw)
	}
	return slug, nil
}

// validCollection reports whether name is a member of collections, the
// membership pre-create must enforce before any path use.
func validCollection(name string, collections []string) bool {
	return slices.Contains(collections, name)
}

// Session is one contributor's upload in progress, whether created
// explicitly via POST /api/session (a multi-file album) or implicitly at
// tus pre-create (a lone ZIP). The exported fields are the plan's public
// shape; the unexported fields are the completion bookkeeping serve
// needs to enforce the album-as-verdict-unit rule exactly once per
// session.
type Session struct {
	ID            string
	Email         string
	Album         string
	Collection    string
	ExpectedFiles int

	createdAt time.Time

	mu            sync.Mutex
	reserved      int   // upload slots admitted at pre-create
	admittedBytes int64 // declared sizes of every admitted slot, summed
	doneUploads   int   // individual tus uploads that have finished
	verdicts      []intake.FileVerdict
	completed     bool
	completedAt   time.Time
	outcome       sessionOutcome
	doneCh        chan struct{}
}

// errSessionFull marks a [Session.reserve] refusal because every
// declared expected_files slot is already reserved.
var errSessionFull = errors.New("serve: session already has its expected number of files")

// errAlbumTooLarge marks a [Session.reserve] refusal because admitting
// size bytes would push the session's running total past maxBytes: the
// aggregate cap the tus protocol has no native hook for, enforced here
// instead since [Store.create]'s expected_files check alone bounds slot
// count, not bytes.
var errAlbumTooLarge = errors.New("serve: album exceeds the aggregate byte cap")

// expiresAt reports when sess becomes eligible for the partial-upload
// sweep, ttl after it was created.
func (sess *Session) expiresAt(ttl time.Duration) time.Time {
	return sess.createdAt.Add(ttl)
}

// reserve admits one more upload slot on sess, returning the slot index,
// unless every [Session.ExpectedFiles] slot is already reserved
// ([errSessionFull], the N+1st upload's pre-create rejection point) or
// admitting size bytes would push the session's running total past
// maxBytes ([errAlbumTooLarge], the aggregate album-total cap the
// multi-file path otherwise never enforces).
func (sess *Session) reserve(size, maxBytes int64) (int, error) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.reserved >= sess.ExpectedFiles {
		return 0, errSessionFull
	}
	if sess.admittedBytes+size > maxBytes {
		return 0, errAlbumTooLarge
	}
	sess.admittedBytes += size
	slot := sess.reserved
	sess.reserved++
	return slot, nil
}

// uploadDone records that one individual tus upload finished, appending
// its verdicts (a ZIP contributes many at once; a single FLAC
// contributes one) to sess's accumulated list. It reports whether every
// reserved slot has now reported in -- completion by count, the first of
// the two ways a session finishes (the other is an explicit /finish
// call via [Engine.completeSession]).
func (sess *Session) uploadDone(verdicts []intake.FileVerdict) bool {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	sess.verdicts = append(sess.verdicts, verdicts...)
	sess.doneUploads++
	return sess.doneUploads >= sess.ExpectedFiles
}

// doneCount reports how many individual uploads have finished so far.
func (sess *Session) doneCount() int {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	return sess.doneUploads
}

// completion reports whether sess has completed and, if it has, when
// [Session.setOutcome] recorded its outcome. Both fields are read under
// sess's own lock in one pass, so [Store]'s scans filter on them without
// each hand-rolling the same lock, read, unlock.
func (sess *Session) completion() (done bool, at time.Time) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	return sess.completed, sess.completedAt
}

// verdictsSnapshot returns a copy of sess's accumulated verdicts, safe
// for [intake.Decide] to range over without racing a concurrent
// [Session.uploadDone].
func (sess *Session) verdictsSnapshot() []intake.FileVerdict {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	out := make([]intake.FileVerdict, len(sess.verdicts))
	copy(out, sess.verdicts)
	return out
}

// completeOnce marks sess completed and reports true on exactly the one
// caller that gets to run [intake.Decide] and promote or discard sess:
// the count-reaching-ExpectedFiles path and an explicit /finish call can
// race for the same session, and only one may act.
func (sess *Session) completeOnce() bool {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.completed {
		return false
	}
	sess.completed = true
	sess.doneCh = make(chan struct{})
	return true
}

// setOutcome records sess's final outcome as of now and wakes every
// caller blocked in [Session.waitOutcome]. now is the sweep's own clock
// seam ([Engine.now]), not [time.Now] directly, so [Store.settled] stays
// controllable from a test the same way [Store.expired] already is.
func (sess *Session) setOutcome(out sessionOutcome, now time.Time) {
	sess.mu.Lock()
	sess.outcome = out
	sess.completedAt = now
	ch := sess.doneCh
	sess.mu.Unlock()
	close(ch)
}

// waitOutcome blocks until sess's outcome is set by the caller that won
// [Session.completeOnce], or ctx is done.
func (sess *Session) waitOutcome(ctx context.Context) (sessionOutcome, error) {
	sess.mu.Lock()
	ch := sess.doneCh
	sess.mu.Unlock()

	select {
	case <-ch:
		sess.mu.Lock()
		defer sess.mu.Unlock()
		return sess.outcome, nil
	case <-ctx.Done():
		return sessionOutcome{}, ctx.Err()
	}
}

// sessionOutcome is a session's completion result: the album verdict,
// whether it was promoted into the inbox, and -- for a main-lane session
// that passed the format verdict -- the custody-policy tag-check probe's
// result. PushedBack and ProbeUnavailable are mutually exclusive with
// Promoted and with each other; MatchedArtist/MatchedAlbum/MatchedYear
// are set only alongside Promoted for a main-lane session (a private-lane
// promotion skips the probe and leaves them empty).
type sessionOutcome struct {
	Verdict          intake.AlbumVerdict
	Promoted         bool
	Dest             string
	PushedBack       bool
	ProbeUnavailable bool
	Detail           string
	MatchedArtist    string
	MatchedAlbum     string
	MatchedYear      string
}

// Store holds every in-flight [Session], keyed by ID. The zero Store is
// not usable; construct with newStore.
type Store struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

func newStore() *Store {
	return &Store{sessions: make(map[string]*Session)}
}

// create validates and registers a new session for email, returning
// [ErrUnknownCollection] or the [slugifyAlbum] error unmodified when
// either input fails validation.
func (s *Store) create(email, rawAlbum, collection string, expectedFiles int, collections []string, now time.Time) (*Session, error) {
	if expectedFiles < 1 {
		return nil, fmt.Errorf("serve: expected_files must be at least 1, got %d", expectedFiles)
	}
	if expectedFiles > maxExpectedFiles {
		return nil, fmt.Errorf("serve: expected_files must be at most %d, got %d", maxExpectedFiles, expectedFiles)
	}
	if !validCollection(collection, collections) {
		return nil, fmt.Errorf("%w: %q", ErrUnknownCollection, collection)
	}
	album, err := slugifyAlbum(rawAlbum)
	if err != nil {
		return nil, err
	}
	id, err := newSessionID()
	if err != nil {
		return nil, err
	}

	sess := &Session{
		ID:            id,
		Email:         email,
		Album:         album,
		Collection:    collection,
		ExpectedFiles: expectedFiles,
		createdAt:     now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = sess
	return sess, nil
}

// get returns the session named by id, and whether one was found.
func (s *Store) get(id string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	return sess, ok
}

// delete removes id from the store.
func (s *Store) delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

// forEmail returns every session belonging to email, completed or not.
// Order is unspecified.
func (s *Store) forEmail(email string) []*Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*Session
	for _, sess := range s.sessions {
		if sess.Email == email {
			out = append(out, sess)
		}
	}
	return out
}

// countActive reports how many of email's sessions have not yet
// completed -- the count [config.UploadConfig.MaxSessionsPerUser] bounds.
func (s *Store) countActive(email string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, sess := range s.sessions {
		if sess.Email != email {
			continue
		}
		if done, _ := sess.completion(); !done {
			n++
		}
	}
	return n
}

// expired returns every session created before cutoff that has never
// completed -- the sweep's target list.
func (s *Store) expired(cutoff time.Time) []*Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*Session
	for _, sess := range s.sessions {
		if done, _ := sess.completion(); !done && sess.createdAt.Before(cutoff) {
			out = append(out, sess)
		}
	}
	return out
}

// settled returns every completed session whose outcome was set before
// cutoff: entries the sweep can now forget from the map, since the
// contributor's journey poll has long since had the chance to read the
// final state back. Unlike [Store.expired], a settled session's own
// on-disk content (if any survives it -- a pushed-back, probe-
// unavailable, or failed-promote directory) is not this method's or its
// caller's map-only deletion to remove; [Engine.sweepOrphans] reclaims
// that directly from UploadsDir once it is unclaimed here and old enough.
func (s *Store) settled(cutoff time.Time) []*Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*Session
	for _, sess := range s.sessions {
		if done, completedAt := sess.completion(); done && completedAt.Before(cutoff) {
			out = append(out, sess)
		}
	}
	return out
}
