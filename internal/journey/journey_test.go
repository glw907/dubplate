package journey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/reject"
)

func openTestLedger(t *testing.T) *ledger.Ledger {
	t.Helper()
	l, err := ledger.Open(filepath.Join(t.TempDir(), "contributions.jsonl"))
	if err != nil {
		t.Fatalf("ledger.Open: %v", err)
	}
	return l
}

func TestResolveNoLedgerNoLiveReturnsErrNotFound(t *testing.T) {
	l := openTestLedger(t)

	_, err := Resolve("nonexistent-session", l, LiveSet{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Resolve = %v, want ErrNotFound", err)
	}
}

func TestResolveLiveOnlyFallsBackToLiveState(t *testing.T) {
	l := openTestLedger(t)
	live := LiveSet{
		"sess-1": {
			SessionID: "sess-1",
			Email:     "alice@example.com",
			State:     ledger.StateUploading,
			Files:     []intake.FileVerdict{{Name: "01.flac", OK: true}},
		},
	}

	j, err := Resolve("sess-1", l, live)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if j.State != ledger.StateUploading {
		t.Errorf("State = %q, want %q", j.State, ledger.StateUploading)
	}
	if len(j.Files) != 1 || j.Files[0].Name != "01.flac" {
		t.Errorf("Files = %+v, want the live session's verdicts", j.Files)
	}
	if j.LibraryLink != "" {
		t.Errorf("LibraryLink = %q, want empty for a non-in-library state", j.LibraryLink)
	}
}

func TestResolveLedgerEventIsAuthoritativeOverLive(t *testing.T) {
	l := openTestLedger(t)
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	if err := l.Append(ledger.Event{
		SessionID: "sess-1", Email: "alice@example.com", Album: "Loveless",
		State: ledger.StateTagging, At: now,
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// A stale live entry (as if serve had not yet noticed completion)
	// must not override the ledger's own, later-stage event.
	live := LiveSet{"sess-1": {SessionID: "sess-1", Email: "alice@example.com", State: ledger.StateUploading}}

	j, err := Resolve("sess-1", l, live)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if j.State != ledger.StateTagging {
		t.Errorf("State = %q, want %q (ledger wins over live)", j.State, ledger.StateTagging)
	}
	if j.Album != "Loveless" {
		t.Errorf("Album = %q, want %q", j.Album, "Loveless")
	}
}

func TestResolveDetailFromLatestLedgerEvent(t *testing.T) {
	l := openTestLedger(t)
	now := time.Now()
	if err := l.Append(ledger.Event{
		SessionID: "sess-1", Email: "alice@example.com", Album: "Loveless",
		State: ledger.StateReview, At: now, Detail: "needs a human look",
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	j, err := Resolve("sess-1", l, LiveSet{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if j.Detail != "needs a human look" {
		t.Errorf("Detail = %q, want %q", j.Detail, "needs a human look")
	}
}

func TestTerminal(t *testing.T) {
	tests := []struct {
		state ledger.State
		want  bool
	}{
		{ledger.StateInLibrary, true},
		{ledger.StateReview, true},
		{ledger.StateRejected, true},
		{ledger.StateUploading, false},
		{ledger.StateUploaded, false},
		{ledger.StateChecking, false},
		{ledger.StateTagging, false},
		{ledger.StateQueued, false},
		{ledger.StateStalled, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := Terminal(tt.state); got != tt.want {
				t.Errorf("Terminal(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

// --- AdvanceStalled ---------------------------------------------------

// fakeSend records every call AdvanceStalled makes through it, and
// returns sendErr when set -- the per-call error seam assertion
// discipline requires of any fake standing in for a fallible send.
type fakeSend struct {
	mu      sync.Mutex
	calls   int
	lastTo  string
	sendErr error
}

func (f *fakeSend) Send(_ context.Context, to, _, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.lastTo = to
	return f.sendErr
}

func (f *fakeSend) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestAdvanceStalledAppendsAndNotifiesOnce(t *testing.T) {
	l := openTestLedger(t)
	uploadedAt := time.Date(2026, 8, 31, 9, 0, 0, 0, time.UTC)
	if err := l.Append(ledger.Event{
		SessionID: "sess-1", Email: "alice@example.com", Album: "Loveless",
		State: ledger.StateUploaded, At: uploadedAt,
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	fs := &fakeSend{}
	now := uploadedAt.Add(45 * time.Minute) // past a 30-minute stall floor
	events, err := l.Session("sess-1")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}

	if err := AdvanceStalled(context.Background(), l, fs.Send, "admin@example.com", 30, events, now); err != nil {
		t.Fatalf("AdvanceStalled: %v", err)
	}
	if fs.callCount() != 1 {
		t.Fatalf("send called %d times, want 1", fs.callCount())
	}
	if fs.lastTo != "admin@example.com" {
		t.Errorf("sent to %q, want %q", fs.lastTo, "admin@example.com")
	}

	events, err = l.Session("sess-1")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	if len(events) != 2 || events[0].State != ledger.StateStalled {
		t.Fatalf("events = %+v, want a StateStalled event appended", events)
	}

	// A second call -- standing in for a restart wiping any in-memory
	// de-dupe -- must not send again: the appended event is itself the
	// de-dupe record.
	if err := AdvanceStalled(context.Background(), l, fs.Send, "admin@example.com", 30, events, now.Add(time.Hour)); err != nil {
		t.Fatalf("AdvanceStalled (second call): %v", err)
	}
	if fs.callCount() != 1 {
		t.Fatalf("send called %d times after a second call, want still 1 (restart cannot re-send)", fs.callCount())
	}
}

func TestAdvanceStalledNoOpBeforeThreshold(t *testing.T) {
	l := openTestLedger(t)
	at := time.Date(2026, 8, 31, 9, 0, 0, 0, time.UTC)
	if err := l.Append(ledger.Event{SessionID: "sess-1", Email: "alice@example.com", State: ledger.StateUploaded, At: at}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	fs := &fakeSend{}
	events, _ := l.Session("sess-1")

	if err := AdvanceStalled(context.Background(), l, fs.Send, "admin@example.com", 30, events, at.Add(5*time.Minute)); err != nil {
		t.Fatalf("AdvanceStalled: %v", err)
	}
	if fs.callCount() != 0 {
		t.Fatalf("send called %d times before the stall threshold, want 0", fs.callCount())
	}
	events, _ = l.Session("sess-1")
	if len(events) != 1 {
		t.Fatalf("events = %+v, want no stalled event appended yet", events)
	}
}

func TestAdvanceStalledNoOpOnceTerminal(t *testing.T) {
	l := openTestLedger(t)
	at := time.Date(2026, 8, 31, 9, 0, 0, 0, time.UTC)
	if err := l.Append(ledger.Event{SessionID: "sess-1", Email: "alice@example.com", State: ledger.StateInLibrary, At: at}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	fs := &fakeSend{}
	events, _ := l.Session("sess-1")

	if err := AdvanceStalled(context.Background(), l, fs.Send, "admin@example.com", 30, events, at.Add(48*time.Hour)); err != nil {
		t.Fatalf("AdvanceStalled: %v", err)
	}
	if fs.callCount() != 0 {
		t.Fatalf("send called %d times for an already-settled session, want 0", fs.callCount())
	}
}

// --- AdvanceInLibrary ---------------------------------------------------

// stubNavidrome serves a fixed Subsonic getScanStatus response, standing
// in for a deployed Navidrome (the plan's "stubbed-Navidrome test").
func stubNavidrome(t *testing.T, scanning bool, lastScan *time.Time) *NavidromeScanClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/rest/getScanStatus") {
			http.NotFound(w, r)
			return
		}
		var lastScanJSON string
		if lastScan != nil {
			b, _ := json.Marshal(lastScan)
			lastScanJSON = string(b)
		} else {
			lastScanJSON = "null"
		}
		_, _ = fmt.Fprintf(w, `{"subsonic-response":{"status":"ok","scanStatus":{"scanning":%v,"lastScan":%s}}}`, scanning, lastScanJSON)
	}))
	t.Cleanup(srv.Close)
	return &NavidromeScanClient{URL: srv.URL, User: "admin", Password: "secret"}
}

func TestAdvanceInLibraryPromotesOnceScanConfirmsCompletion(t *testing.T) {
	l := openTestLedger(t)
	taggedAt := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	if err := l.Append(ledger.Event{
		SessionID: "sess-1", Email: "alice@example.com", Album: "Loveless",
		State: ledger.StateTagging, At: taggedAt, LibraryPaths: []string{"/lib/loveless"},
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	lastScan := taggedAt.Add(5 * time.Minute)
	nav := stubNavidrome(t, false, &lastScan)
	events, _ := l.Session("sess-1")

	if err := AdvanceInLibrary(context.Background(), l, nav, events, taggedAt.Add(10*time.Minute)); err != nil {
		t.Fatalf("AdvanceInLibrary: %v", err)
	}

	events, err := l.Session("sess-1")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	if len(events) != 2 || events[0].State != ledger.StateInLibrary {
		t.Fatalf("events = %+v, want an in-library event appended", events)
	}
}

func TestAdvanceInLibraryNoOpWhileScanning(t *testing.T) {
	l := openTestLedger(t)
	taggedAt := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	if err := l.Append(ledger.Event{SessionID: "sess-1", Email: "alice@example.com", State: ledger.StateTagging, At: taggedAt}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	nav := stubNavidrome(t, true, nil) // a scan is running right now
	events, _ := l.Session("sess-1")

	if err := AdvanceInLibrary(context.Background(), l, nav, events, taggedAt.Add(time.Minute)); err != nil {
		t.Fatalf("AdvanceInLibrary: %v", err)
	}
	events, _ = l.Session("sess-1")
	if len(events) != 1 {
		t.Fatalf("events = %+v, want no in-library event while a scan is in progress", events)
	}
}

func TestAdvanceInLibraryNoOpWhenLastScanPredatesTagging(t *testing.T) {
	l := openTestLedger(t)
	taggedAt := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	if err := l.Append(ledger.Event{SessionID: "sess-1", Email: "alice@example.com", State: ledger.StateTagging, At: taggedAt}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	staleScan := taggedAt.Add(-time.Hour) // the last completed scan predates the tag
	nav := stubNavidrome(t, false, &staleScan)
	events, _ := l.Session("sess-1")

	if err := AdvanceInLibrary(context.Background(), l, nav, events, taggedAt.Add(time.Minute)); err != nil {
		t.Fatalf("AdvanceInLibrary: %v", err)
	}
	events, _ = l.Session("sess-1")
	if len(events) != 1 {
		t.Fatalf("events = %+v, want no in-library event promoted from a stale scan", events)
	}
}

func TestAdvanceInLibraryNoOpUnlessLatestIsTagging(t *testing.T) {
	l := openTestLedger(t)
	at := time.Now()
	if err := l.Append(ledger.Event{SessionID: "sess-1", Email: "alice@example.com", State: ledger.StateUploaded, At: at}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	nav := stubNavidrome(t, false, nil)
	events, _ := l.Session("sess-1")

	if err := AdvanceInLibrary(context.Background(), l, nav, events, at.Add(time.Minute)); err != nil {
		t.Fatalf("AdvanceInLibrary: %v", err)
	}
	events, _ = l.Session("sess-1")
	if len(events) != 1 {
		t.Fatalf("events = %+v, want no in-library event when the latest state isn't tagging", events)
	}
}

// --- QueuedDetail / BuildLibraryLink ------------------------------------

func TestQueuedDetail(t *testing.T) {
	tests := []struct {
		name             string
		ownEarlier, held bool
		want             string
	}{
		{"own earlier upload wins", true, true, "waiting on your other upload"},
		{"own earlier upload alone", true, false, "waiting on your other upload"},
		{"library busy alone", false, true, "waiting for the library to finish its current import"},
		{"neither", false, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QueuedDetail(tt.ownEarlier, tt.held); got != tt.want {
				t.Errorf("QueuedDetail(%v, %v) = %q, want %q", tt.ownEarlier, tt.held, got, tt.want)
			}
		})
	}
}

func TestBuildLibraryLink(t *testing.T) {
	link := BuildLibraryLink("https://music.example.com", "Loveless")
	if !strings.HasPrefix(link, "https://music.example.com/app/#/album?filter=") {
		t.Fatalf("link = %q, want the /app/#/album route with a filter param", link)
	}
	if !strings.Contains(link, "name") {
		t.Errorf("link = %q, want the album-name filter key", link)
	}

	if got := BuildLibraryLink("", "Loveless"); got != "" {
		t.Errorf("BuildLibraryLink with empty publicURL = %q, want empty", got)
	}
	if got := BuildLibraryLink("https://music.example.com", ""); got != "" {
		t.Errorf("BuildLibraryLink with empty album = %q, want empty", got)
	}
}

// TestReasonRoundTripsThroughJourneyFiles proves a rejected verdict's
// reject.Reason survives being carried on a Journey.Files entry all the
// way onto the wire: it marshals the resolved Journey and asserts the
// emitted Files[0].reason object, not just a Go-side method call, so a
// broken MarshalJSON on the threaded-through field would fail this test.
func TestReasonRoundTripsThroughJourneyFiles(t *testing.T) {
	live := LiveSet{"sess-1": {
		SessionID: "sess-1", Email: "alice@example.com", State: ledger.StateRejected,
		Files: []intake.FileVerdict{{Name: "01.flac", Reason: reject.ReasonBelowCDQuality}},
	}}
	l := openTestLedger(t)

	j, err := Resolve("sess-1", l, live)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	body, err := json.Marshal(j)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded struct {
		Files []struct {
			Reason struct {
				Code     string `json:"code"`
				Friendly string `json:"friendly"`
				NextStep string `json:"next_step"`
			} `json:"reason"`
		} `json:"files"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v (body: %s)", err, body)
	}
	if len(decoded.Files) != 1 {
		t.Fatalf("files = %+v, want exactly one entry", decoded.Files)
	}
	got := decoded.Files[0].Reason
	wantFriendly := "This FLAC is below CD quality."
	wantNextStep := "Get a CD-quality or better FLAC and upload that instead."
	if got.Code != "below-cd-quality" || got.Friendly != wantFriendly || got.NextStep != wantNextStep {
		t.Errorf("Files[0].reason = %+v, want {code: below-cd-quality, friendly: %q, next_step: %q}",
			got, wantFriendly, wantNextStep)
	}
}
