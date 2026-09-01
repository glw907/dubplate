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

	mu          sync.Mutex
	reserved    int // upload slots admitted at pre-create
	doneUploads int // individual tus uploads that have finished
	verdicts    []intake.FileVerdict
	completed   bool
	outcome     sessionOutcome
	doneCh      chan struct{}
}

// expiresAt reports when sess becomes eligible for the partial-upload
// sweep, ttl after it was created.
func (sess *Session) expiresAt(ttl time.Duration) time.Time {
	return sess.createdAt.Add(ttl)
}

// reserve admits one more upload slot on sess, returning the slot index
// and true, unless every [Session.ExpectedFiles] slot is already
// reserved -- the N+1st upload's pre-create rejection point.
func (sess *Session) reserve() (int, bool) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.reserved >= sess.ExpectedFiles {
		return 0, false
	}
	slot := sess.reserved
	sess.reserved++
	return slot, true
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

// setOutcome records sess's final outcome and wakes every caller
// blocked in [Session.waitOutcome].
func (sess *Session) setOutcome(out sessionOutcome) {
	sess.mu.Lock()
	sess.outcome = out
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
// and whether it was promoted into the inbox.
type sessionOutcome struct {
	Verdict  intake.AlbumVerdict
	Promoted bool
	Dest     string
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
		sess.mu.Lock()
		done := sess.completed
		sess.mu.Unlock()
		if !done {
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
		sess.mu.Lock()
		done := sess.completed
		sess.mu.Unlock()
		if !done && sess.createdAt.Before(cutoff) {
			out = append(out, sess)
		}
	}
	return out
}
