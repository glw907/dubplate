// Package journey composes the estate's one contributor-facing view of
// an upload's progress: [ledger]'s persisted history plus serve's own
// in-memory view of sessions still in flight. The dependency runs this
// package -> internal/ledger, one direction; serve builds the live half
// itself and passes it in via [Resolve], so this package never reaches
// back into internal/serve and no import cycle exists between them.
//
// This package also owns the two externally observable facts that move
// a session's ledger history forward on serve's behalf: confirming a
// scan-completed promotion to the library ([AdvanceInLibrary]) and
// flagging an overdue session as stalled ([AdvanceStalled]). Neither
// runs inside [Resolve] itself, which stays a pure read composing from
// the ledger and live state only; internal/serve/api.go calls both
// before every poll.
package journey

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/ledger"
)

// LiveSession is serve's in-memory view of one session its upload store
// still tracks: an upload in flight, or one that finished uploading
// before the importer ever appended its first ledger event. It is owned
// here rather than in internal/serve so the dependency runs
// serve -> journey -> ledger in one direction, with no cycle back into
// serve.
type LiveSession struct {
	SessionID string
	Email     string
	State     ledger.State
	Files     []intake.FileVerdict
}

// LiveSet is every [LiveSession] serve currently holds, keyed by session
// ID.
type LiveSet map[string]LiveSession

// Journey is one session's current state as [Resolve] composes it: the
// one shape both GET /api/journey/<id> and GET /api/history render.
type Journey struct {
	SessionID   string
	State       ledger.State
	Detail      string
	Album       string
	LibraryLink string
	Files       []intake.FileVerdict
}

// ErrNotFound marks a session absent from both the ledger and live: the
// case a caller answers with an HTTP 404, the same status a session
// that exists but belongs to someone else gets, so a prober can never
// tell "doesn't exist" from "isn't yours."
var ErrNotFound = errors.New("journey: session not found")

// Resolve composes sessionID's current [Journey] from l's ledger history
// and live, and nothing else: no run report, no filesystem probe (run
// reports are per-run monitoring artifacts journey never reads). A
// ledger event, when one exists, is authoritative: the ledger is only
// ever written for a state past what live tracks, so live only
// contributes Files and, while the ledger has no event yet for
// sessionID, State itself. Resolve does not check ownership -- scoping
// a read to a caller's verified email is internal/serve/api.go's job,
// since it is the one holding that email, not Resolve.
func Resolve(sessionID string, l *ledger.Ledger, live LiveSet) (Journey, error) {
	events, err := l.Session(sessionID)
	if err != nil {
		return Journey{}, fmt.Errorf("resolve session %s: %w", sessionID, err)
	}
	liveSess, isLive := live[sessionID]

	if len(events) == 0 {
		if !isLive {
			return Journey{}, ErrNotFound
		}
		return Journey{SessionID: sessionID, State: liveSess.State, Files: liveSess.Files}, nil
	}

	latest := events[0]
	j := Journey{SessionID: sessionID, State: latest.State, Detail: latest.Detail, Album: latest.Album}
	if isLive {
		j.Files = liveSess.Files
	}
	return j, nil
}

// Terminal reports whether s is one of the three states that settle a
// session for good: no further event, live or ledger, is ever expected
// past one of these.
func Terminal(s ledger.State) bool {
	return s == ledger.StateInLibrary || s == ledger.StateReview || s == ledger.StateRejected
}

// SendFunc matches [*notify.Mailer.Send]'s signature: the seam
// [AdvanceStalled] sends its one admin notification through, so a test
// can substitute a plain closure instead of standing up a real SMTP
// server. internal/serve/api.go passes a real Mailer's Send method
// value.
type SendFunc func(ctx context.Context, to, subject, body string) error

// AdvanceStalled appends [ledger.StateStalled] for the session events
// describes -- newest first, as [ledger.Ledger.Session] returns them --
// once it has sat stallMinutes past its latest event with no terminal
// state reached, and sends exactly one admin notification through send
// alongside it. The appended event is the de-dupe record, not merely a
// display state (N4): a later call, even after a restart drops any
// in-memory memory of having already notified, finds the
// [ledger.StateStalled] event already in events and sends nothing a
// second time. events must be non-empty.
func AdvanceStalled(ctx context.Context, l *ledger.Ledger, send SendFunc, adminEmail string, stallMinutes int, events []ledger.Event, now time.Time) error {
	latest := events[0]
	if Terminal(latest.State) || hasState(events, ledger.StateStalled) {
		return nil
	}
	if now.Sub(latest.At) < time.Duration(stallMinutes)*time.Minute {
		return nil
	}

	if err := l.Append(ledger.Event{
		SessionID:  latest.SessionID,
		Email:      latest.Email,
		Album:      latest.Album,
		Collection: latest.Collection,
		State:      ledger.StateStalled,
		At:         now,
	}); err != nil {
		return fmt.Errorf("append stalled event for session %s: %w", latest.SessionID, err)
	}

	subject := fmt.Sprintf("dubplate: %s's upload has stalled", latest.Email)
	body := fmt.Sprintf("Session %s (album %q, contributor %s) has had no progress since %s.",
		latest.SessionID, latest.Album, latest.Email, latest.At.Format(time.RFC3339))
	if err := send(ctx, adminEmail, subject, body); err != nil {
		return fmt.Errorf("notify admin of stalled session %s: %w", latest.SessionID, err)
	}
	return nil
}

// AdvanceInLibrary appends [ledger.StateInLibrary] for the session
// events describes once nav confirms Navidrome's own scan finished
// after the importer's [ledger.StateTagging] event -- never on
// StateTagging alone, since triggering a scan and the scan completing
// are asynchronous (M4: in-library waits for scan COMPLETION). events
// must be non-empty; AdvanceInLibrary is a no-op unless the latest event
// is StateTagging.
func AdvanceInLibrary(ctx context.Context, l *ledger.Ledger, nav *NavidromeScanClient, events []ledger.Event, now time.Time) error {
	latest := events[0]
	if latest.State != ledger.StateTagging {
		return nil
	}

	status, err := nav.ScanStatus(ctx)
	if err != nil {
		return fmt.Errorf("check navidrome scan status for session %s: %w", latest.SessionID, err)
	}
	if status.Scanning || status.LastScan == nil || !status.LastScan.After(latest.At) {
		// Either a scan is still running, or the last completed scan
		// finished before this session was tagged -- it has not yet
		// confirmed this session's files landed.
		return nil
	}

	if err := l.Append(ledger.Event{
		SessionID:  latest.SessionID,
		Email:      latest.Email,
		Album:      latest.Album,
		Collection: latest.Collection,
		State:      ledger.StateInLibrary,
		At:         now,
	}); err != nil {
		return fmt.Errorf("append in-library event for session %s: %w", latest.SessionID, err)
	}
	return nil
}

func hasState(events []ledger.Event, s ledger.State) bool {
	for _, e := range events {
		if e.State == s {
			return true
		}
	}
	return false
}

// QueuedDetail returns the copy variant a [ledger.StateQueued] Journey's
// Detail carries: an own-upload wait takes priority over the shared
// import lock, since it is the more specific, more actionable fact for
// the contributor to see (M6).
func QueuedDetail(ownEarlierUpload, importLockHeld bool) string {
	switch {
	case ownEarlierUpload:
		return "waiting on your other upload"
	case importLockHeld:
		return "waiting for the library to finish its current import"
	default:
		return ""
	}
}

// navidromeAppPath and navidromeAlbumFilterKey are Navidrome's own web
// UI route and react-admin filter field, both verified against the
// pinned deluan/navidrome image tag 0.63.2 (compose.yaml): the SPA
// mounts at /app (consts.URLPathUI) with hash-based client routing, and
// the album list's SearchInput targets the "name" source
// (ui/src/album/AlbumList.jsx), so a react-admin filter query param of
// {"name":"<title>"} pre-fills that search.
const (
	navidromeAppPath        = "/app/#/album"
	navidromeAlbumFilterKey = "name"
)

// BuildLibraryLink returns the deep link a [ledger.StateInLibrary]
// Journey's LibraryLink carries: publicURL's Navidrome app, scoped to a
// search for album's title via the album-list filter route named above.
// It returns "" if either input is empty, so a caller never links to a
// broken URL; the album-list search is the only link resolvable from
// ledger data (the album title) without new config -- Navidrome assigns
// no stable, ledger-visible album ID this codebase could deep-link to
// instead.
func BuildLibraryLink(publicURL, album string) string {
	if publicURL == "" || album == "" {
		return ""
	}
	filter, err := json.Marshal(map[string]string{navidromeAlbumFilterKey: album})
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(publicURL, "/") + navidromeAppPath + "?filter=" + url.QueryEscape(string(filter))
}

// scanStatusTimeout bounds one getScanStatus request against a deployed
// Navidrome.
const scanStatusTimeout = 10 * time.Second

// subsonicVersion and subsonicClient mirror internal/beets.TriggerScan's
// own copies of the same two constants: each package that makes this
// one Subsonic call owns its own literal rather than importing a shared
// constant, the choice internal/serve/upload.go's cdMinSampleRate doc
// comment already names for this codebase.
const (
	subsonicVersion = "1.16.1"
	subsonicClient  = "dubplate-journey"
)

// NavidromeScanClient calls Navidrome's Subsonic getScanStatus endpoint
// (GET /rest/getScanStatus, verified against the pinned deluan/navidrome
// image tag 0.63.2 in compose.yaml): [AdvanceInLibrary]'s only way to
// learn whether a library scan has actually finished since the importer
// tagged a session, rather than trusting the tagging event alone.
type NavidromeScanClient struct {
	URL      string
	User     string
	Password string

	// HTTPClient overrides the client's transport; the zero value dials
	// with a [scanStatusTimeout]-bounded client. Tests point this at an
	// httptest.Server standing in for Navidrome.
	HTTPClient *http.Client
}

// ScanStatus is Subsonic getScanStatus's scanStatus element, decoded to
// the two fields [AdvanceInLibrary] needs: whether a scan is running
// right now, and when the last one finished.
type ScanStatus struct {
	Scanning bool
	LastScan *time.Time
}

type scanStatusEnvelope struct {
	SubsonicResponse struct {
		Status     string `json:"status"`
		ScanStatus struct {
			Scanning bool       `json:"scanning"`
			LastScan *time.Time `json:"lastScan"`
		} `json:"scanStatus"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	} `json:"subsonic-response"`
}

// ScanStatus fetches Navidrome's current scan status. Authentication is
// Subsonic's query-param scheme -- t is md5(password+salt), s a fresh
// random salt per call, u/v/c the username, API version, and client name
// -- mirroring internal/beets.TriggerScan's own copy of the same
// construction.
func (c *NavidromeScanClient) ScanStatus(ctx context.Context) (ScanStatus, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return ScanStatus{}, fmt.Errorf("generate scan-status salt: %w", err)
	}
	saltHex := hex.EncodeToString(salt)
	sum := md5.Sum([]byte(c.Password + saltHex))
	token := hex.EncodeToString(sum[:])

	u, err := url.Parse(strings.TrimSuffix(c.URL, "/") + "/rest/getScanStatus")
	if err != nil {
		return ScanStatus{}, fmt.Errorf("parse navidrome url %s: %w", c.URL, err)
	}
	q := u.Query()
	q.Set("u", c.User)
	q.Set("t", token)
	q.Set("s", saltHex)
	q.Set("v", subsonicVersion)
	q.Set("c", subsonicClient)
	q.Set("f", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ScanStatus{}, fmt.Errorf("build scan-status request: %w", err)
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: scanStatusTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return ScanStatus{}, fmt.Errorf("navidrome scan-status request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return ScanStatus{}, fmt.Errorf("navidrome scan-status request: status %s", resp.Status)
	}

	var env scanStatusEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return ScanStatus{}, fmt.Errorf("parse navidrome scan-status response: %w", err)
	}
	if env.SubsonicResponse.Status != "ok" {
		return ScanStatus{}, fmt.Errorf("navidrome scan-status: %s", env.SubsonicResponse.Error.Message)
	}

	return ScanStatus{
		Scanning: env.SubsonicResponse.ScanStatus.Scanning,
		LastScan: env.SubsonicResponse.ScanStatus.LastScan,
	}, nil
}
