package serve

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/journey"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/notify"
)

const testEmailBob = "bob@example.com"

// testJourneyAPI bundles a [JourneyAPI] with the ledger, lock path, and
// stubbed-Navidrome server it was built over, plus a running server that
// stands in for cmd/dubplate/serve.go's future mount of Handler.
type testJourneyAPI struct {
	te        *testEngine
	japi      *JourneyAPI
	l         *ledger.Ledger
	lockPath  string
	navServer *httptest.Server
	navScan   func(scanning bool, lastScan *time.Time)
	server    *httptest.Server
	client    *http.Client
	now       time.Time
}

// newTestJourneyAPI builds a JourneyAPI over a fresh testEngine (Alice
// only, matching newTestEngine's own roster), a real temp-file ledger, a
// stubbed Navidrome getScanStatus server whose response newTestJourneyAPI's
// caller can change between requests, and a not-yet-created lock path so
// [lock.Held] starts out false.
func newTestJourneyAPI(t *testing.T) *testJourneyAPI {
	t.Helper()

	te := newTestEngine(t, nil)

	l, err := ledger.Open(filepath.Join(t.TempDir(), "contributions.jsonl"))
	if err != nil {
		t.Fatalf("ledger.Open: %v", err)
	}

	var scanning bool
	var lastScan *time.Time
	navServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var lastScanJSON string
		if lastScan != nil {
			b, _ := json.Marshal(lastScan)
			lastScanJSON = string(b)
		} else {
			lastScanJSON = "null"
		}
		_, _ = w.Write([]byte(`{"subsonic-response":{"status":"ok","scanStatus":{"scanning":` +
			boolJSON(scanning) + `,"lastScan":` + lastScanJSON + `}}}`))
	}))
	t.Cleanup(navServer.Close)

	now := te.now
	deps := JourneyDeps{
		Ledger:             l,
		Navidrome:          &journey.NavidromeScanClient{URL: navServer.URL, User: "admin", Password: "secret"},
		Mailer:             notify.New(config.SMTPConfig{Host: "127.0.0.1", Port: 25, From: "dubplate@example.com"}),
		LockPath:           filepath.Join(t.TempDir(), "import.lock"),
		AdminEmail:         "admin@example.com",
		StallMinutes:       30,
		PublicNavidromeURL: "https://music.example.com",
		Now:                func() time.Time { return now },
	}
	japi := NewJourneyAPI(te.engine, deps)

	srv := httptest.NewServer(withTestAuth(japi.Handler()))
	t.Cleanup(srv.Close)

	return &testJourneyAPI{
		te:        te,
		japi:      japi,
		l:         l,
		lockPath:  deps.LockPath,
		navServer: navServer,
		navScan:   func(s bool, ls *time.Time) { scanning, lastScan = s, ls },
		server:    srv,
		client:    srv.Client(),
		now:       now,
	}
}

func boolJSON(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (tj *testJourneyAPI) getJourney(t *testing.T, email, id string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, tj.server.URL+"/api/journey/"+id, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if email != "" {
		req.Header.Set(testEmailHeader, email)
	}
	resp, err := tj.client.Do(req)
	if err != nil {
		t.Fatalf("GET journey: %v", err)
	}
	return resp
}

func (tj *testJourneyAPI) getHistory(t *testing.T, email string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, tj.server.URL+"/api/history", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if email != "" {
		req.Header.Set(testEmailHeader, email)
	}
	resp, err := tj.client.Do(req)
	if err != nil {
		t.Fatalf("GET history: %v", err)
	}
	return resp
}

func decodeJourney(t *testing.T, resp *http.Response) journey.Journey {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	var j journey.Journey
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("decode journey: %v (body so far: %q)", err, body)
	}
	return j
}

// --- authorization --------------------------------------------------

func TestHandleJourney_Unauthenticated(t *testing.T) {
	tj := newTestJourneyAPI(t)
	resp := tj.getJourney(t, "", "anything")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestHandleJourney_InvalidSessionIDFormat404(t *testing.T) {
	tj := newTestJourneyAPI(t)
	resp := tj.getJourney(t, testEmailAlice, "not-a-valid-id")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleJourney_UnknownSession404(t *testing.T) {
	tj := newTestJourneyAPI(t)
	resp := tj.getJourney(t, testEmailAlice, "aaaaaaaaaaaaaaaaaaaaaaaaaa")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

// TestHandleJourney_ForeignSessionDenied404 proves a session that exists
// but belongs to another email answers 404, identically to an unknown
// session ID -- no existence oracle, matching upload.go's own B3
// enforcement style.
func TestHandleJourney_ForeignSessionDenied404(t *testing.T) {
	tj := newTestJourneyAPI(t)
	if err := tj.l.Append(ledger.Event{
		SessionID: "alicegenericsessionxxxxxxx", Email: testEmailAlice, Album: "Loveless",
		State: ledger.StateReview, At: tj.now,
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	resp := tj.getJourney(t, testEmailBob, "alicegenericsessionxxxxxxx")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for a session belonging to another email", resp.StatusCode)
	}
}

func TestHandleJourney_OwnLedgerSessionResolves(t *testing.T) {
	tj := newTestJourneyAPI(t)
	if err := tj.l.Append(ledger.Event{
		SessionID: "alicegenericsessionxxxxxxx", Email: testEmailAlice, Album: "Loveless",
		State: ledger.StateReview, At: tj.now, Detail: "needs a human look",
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	resp := tj.getJourney(t, testEmailAlice, "alicegenericsessionxxxxxxx")
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}
	j := decodeJourney(t, resp)
	if j.State != ledger.StateReview || j.Detail != "needs a human look" {
		t.Errorf("journey = %+v, want state review with the recorded detail", j)
	}
}

// TestHandleJourney_WireKeysAreSnakeCase proves the handler actually
// emits journey.Journey's snake_case JSON tags on the wire, rather than
// only being provable by decoding back into the same tagged struct: it
// decodes the raw response into map[string]any and asserts the key set
// directly.
func TestHandleJourney_WireKeysAreSnakeCase(t *testing.T) {
	tj := newTestJourneyAPI(t)
	if err := tj.l.Append(ledger.Event{
		SessionID: "alicegenericsessionxxxxxxx", Email: testEmailAlice, Album: "Loveless",
		State: ledger.StateReview, At: tj.now, Detail: "needs a human look",
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	resp := tj.getJourney(t, testEmailAlice, "alicegenericsessionxxxxxxx")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode raw journey: %v", err)
	}
	if raw["session_id"] != "alicegenericsessionxxxxxxx" {
		t.Errorf(`raw["session_id"] = %v, want "alicegenericsessionxxxxxxx"`, raw["session_id"])
	}
	if raw["state"] != string(ledger.StateReview) {
		t.Errorf(`raw["state"] = %v, want %q`, raw["state"], ledger.StateReview)
	}
	if raw["detail"] != "needs a human look" {
		t.Errorf(`raw["detail"] = %v, want "needs a human look"`, raw["detail"])
	}
	if raw["album"] != "Loveless" {
		t.Errorf(`raw["album"] = %v, want "Loveless"`, raw["album"])
	}
	if _, present := raw["SessionID"]; present {
		t.Error(`raw["SessionID"] present, want only the snake_case key`)
	}
	if _, present := raw["library_link"]; present {
		t.Error(`raw["library_link"] present for a review-state journey, want it omitted`)
	}
	if _, present := raw["files"]; present {
		t.Error(`raw["files"] present for a journey with no live files, want it omitted`)
	}
}

// --- live state -------------------------------------------------------

// TestHandleJourney_LiveUploadingState proves an in-flight, not-yet-
// completed session (no ledger event exists for it at all yet) reads
// its state from Engine's live store.
func TestHandleJourney_LiveUploadingState(t *testing.T) {
	tj := newTestJourneyAPI(t)
	te := tj.te

	// Register a two-file session and finish only one upload, so the
	// session stays live: not completed, still tracked in the store.
	sess, err := te.engine.sessions.create(testEmailAlice, "Two Files", "", 2, te.cfg.Collections, tj.now)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	data := cdFlacFixture()
	resp := te.uploadWhole(t, data, map[string]string{
		"filename": "01.flac", "session_id": sess.ID,
	})
	_ = resp.Body.Close()

	journeyResp := tj.getJourney(t, testEmailAlice, sess.ID)
	j := decodeJourney(t, journeyResp)
	if j.State != ledger.StateUploading {
		t.Errorf("State = %q, want %q for a session with one of two files uploaded", j.State, ledger.StateUploading)
	}
	if len(j.Files) != 1 {
		t.Errorf("Files = %+v, want the one verdict landed so far", j.Files)
	}
}

// --- in-library link ----------------------------------------------------

// TestHandleJourney_InLibraryPromotesFromStubbedNavidromeScan proves the
// polling handler itself drives AdvanceInLibrary: a tagging event whose
// scan the stubbed Navidrome reports complete promotes to in-library and
// carries a LibraryLink, in the same request that observed it.
func TestHandleJourney_InLibraryPromotesFromStubbedNavidromeScan(t *testing.T) {
	tj := newTestJourneyAPI(t)
	taggedAt := tj.now.Add(-time.Hour)
	if err := tj.l.Append(ledger.Event{
		SessionID: "alicegenericsessionxxxxxxx", Email: testEmailAlice, Album: "Loveless",
		State: ledger.StateTagging, At: taggedAt, LibraryPaths: []string{"/lib/loveless/01.flac"},
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	lastScan := taggedAt.Add(time.Minute)
	tj.navScan(false, &lastScan)

	resp := tj.getJourney(t, testEmailAlice, "alicegenericsessionxxxxxxx")
	j := decodeJourney(t, resp)
	if j.State != ledger.StateInLibrary {
		t.Fatalf("State = %q, want %q", j.State, ledger.StateInLibrary)
	}
	if j.LibraryLink == "" {
		t.Error("LibraryLink is empty for an in-library journey")
	}

	events, err := tj.l.Session("alicegenericsessionxxxxxxx")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	if len(events) != 2 || events[0].State != ledger.StateInLibrary {
		t.Fatalf("events = %+v, want an in-library event appended by the poll itself", events)
	}
}

// --- queued: two copy variants (M6) --------------------------------

// TestHandleJourney_QueuedLibraryBusy proves a completed, promoted
// session with the shared import lock held by another run reads as
// queued with the library-busy copy.
func TestHandleJourney_QueuedLibraryBusy(t *testing.T) {
	tj := newTestJourneyAPI(t)
	te := tj.te

	data := buildZip(t, map[string][]byte{"01.flac": cdFlacFixture()})
	resp := te.uploadWhole(t, data, map[string]string{"filename": "album.zip", "album": "Queued Album", "collection": ""})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: status %d, body %q", resp.StatusCode, body)
	}

	// Find the session ID the store assigned it.
	sessions := te.engine.sessionsForEmail(testEmailAlice)
	if len(sessions) != 1 {
		t.Fatalf("sessionsForEmail = %d sessions, want 1", len(sessions))
	}
	id := sessions[0].ID

	held, err := lock.Acquire(tj.lockPath)
	if err != nil {
		t.Fatalf("Acquire import lock: %v", err)
	}
	t.Cleanup(func() { _ = held.Release() })

	journeyResp := tj.getJourney(t, testEmailAlice, id)
	j := decodeJourney(t, journeyResp)
	if j.State != ledger.StateQueued {
		t.Fatalf("State = %q, want %q while the import lock is held", j.State, ledger.StateQueued)
	}
	if j.Detail != journey.QueuedDetail(false, true) {
		t.Errorf("Detail = %q, want the library-busy copy", j.Detail)
	}
}

// TestHandleJourney_QueuedOwnEarlierUpload proves a second, later
// session for the same contributor reads as queued behind their own
// still-unsettled earlier session, even with the import lock free.
func TestHandleJourney_QueuedOwnEarlierUpload(t *testing.T) {
	tj := newTestJourneyAPI(t)
	te := tj.te

	// Mint both sessions directly, with explicit, distinct createdAt
	// values, then drive each through a real single-file upload keyed to
	// its session_id so both reach completed-and-promoted the same way
	// production does (te.uploadWhole's implicit-ZIP path always stamps
	// the engine's one fixed test clock, which can't produce two
	// distinct timestamps on its own).
	earlierSess, err := te.engine.sessions.create(testEmailAlice, "Earlier", "", 1, te.cfg.Collections, tj.now.Add(-time.Minute))
	if err != nil {
		t.Fatalf("create earlier session: %v", err)
	}
	earlierResp := te.uploadWhole(t, cdFlacFixture(), map[string]string{"filename": "01.flac", "session_id": earlierSess.ID})
	_, _ = io.ReadAll(earlierResp.Body)
	_ = earlierResp.Body.Close()

	laterSess, err := te.engine.sessions.create(testEmailAlice, "Later", "", 1, te.cfg.Collections, tj.now.Add(time.Minute))
	if err != nil {
		t.Fatalf("create later session: %v", err)
	}
	laterResp := te.uploadWhole(t, cdFlacFixture(), map[string]string{"filename": "01.flac", "session_id": laterSess.ID})
	_, _ = io.ReadAll(laterResp.Body)
	_ = laterResp.Body.Close()

	resp := tj.getJourney(t, testEmailAlice, laterSess.ID)
	j := decodeJourney(t, resp)
	if j.State != ledger.StateQueued {
		t.Fatalf("State = %q, want %q behind an unsettled earlier session", j.State, ledger.StateQueued)
	}
	if j.Detail != journey.QueuedDetail(true, false) {
		t.Errorf("Detail = %q, want the own-earlier-upload copy", j.Detail)
	}
}

// --- history ------------------------------------------------------------

func TestHandleHistory_ScopedToEmail(t *testing.T) {
	tj := newTestJourneyAPI(t)
	if err := tj.l.Append(ledger.Event{SessionID: "alicehistorysessionxxxxxxx", Email: testEmailAlice, Album: "A", State: ledger.StateReview, At: tj.now}); err != nil {
		t.Fatalf("Append alice: %v", err)
	}
	if err := tj.l.Append(ledger.Event{SessionID: "bobhistorysessionidxxxxxxx", Email: testEmailBob, Album: "B", State: ledger.StateReview, At: tj.now}); err != nil {
		t.Fatalf("Append bob: %v", err)
	}

	resp := tj.getHistory(t, testEmailAlice)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var out []journey.Journey
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if len(out) != 1 || out[0].SessionID != "alicehistorysessionxxxxxxx" {
		t.Fatalf("history = %+v, want only alice's session", out)
	}
}

// TestHandleHistory_TerminalLinkNeverFails proves an in-library session
// discovered only through /api/history (never individually polled) still
// gets advanced and carries a working LibraryLink -- the M4 guarantee
// that a history link can never 404 because the state was never advanced.
func TestHandleHistory_TerminalLinkNeverFails(t *testing.T) {
	tj := newTestJourneyAPI(t)
	taggedAt := tj.now.Add(-time.Hour)
	if err := tj.l.Append(ledger.Event{
		SessionID: "alicegenericsessionxxxxxxx", Email: testEmailAlice, Album: "Loveless",
		State: ledger.StateTagging, At: taggedAt,
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	lastScan := taggedAt.Add(time.Minute)
	tj.navScan(false, &lastScan)

	resp := tj.getHistory(t, testEmailAlice)
	defer func() { _ = resp.Body.Close() }()
	var out []journey.Journey
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if len(out) != 1 || out[0].State != ledger.StateInLibrary || out[0].LibraryLink == "" {
		t.Fatalf("history = %+v, want the session advanced to in-library with a link", out)
	}
}

// --- stalled: de-dupe survives a restart (N4) ------------------------

func TestHandleJourney_StalledAppendedOnceAcrossRepeatedPolls(t *testing.T) {
	tj := newTestJourneyAPI(t)
	uploadedAt := tj.now.Add(-time.Hour) // well past the 30-minute stall floor
	if err := tj.l.Append(ledger.Event{
		SessionID: "alicegenericsessionxxxxxxx", Email: testEmailAlice, Album: "Loveless",
		State: ledger.StateUploaded, At: uploadedAt,
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	for range 3 {
		resp := tj.getJourney(t, testEmailAlice, "alicegenericsessionxxxxxxx")
		_ = resp.Body.Close()
	}

	events, err := tj.l.Session("alicegenericsessionxxxxxxx")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	stalledCount := 0
	for _, e := range events {
		if e.State == ledger.StateStalled {
			stalledCount++
		}
	}
	if stalledCount != 1 {
		t.Fatalf("stalled events = %d after 3 polls (simulating restarts between them), want exactly 1", stalledCount)
	}
}
