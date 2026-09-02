package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestValidState(t *testing.T) {
	tests := []struct {
		state State
		want  bool
	}{
		{StateUploading, true},
		{StateUploaded, true},
		{StateChecking, true},
		{StateTagging, true},
		{StateInLibrary, true},
		{StateReview, true},
		{StateRejected, true},
		{StateQueued, true},
		{StateStalled, true},
		{State(""), false},
		{State("Uploading"), false},
		{State("in_library"), false},
		{State("done"), false},
		{State("stall"), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := ValidState(tt.state); got != tt.want {
				t.Errorf("ValidState(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestOpenRejectsEmptyPath(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("Open(\"\") returned a nil error, want an error")
	}
}

func TestAppendRejectsInvalidState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contributions.jsonl")
	l, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	err = l.Append(Event{SessionID: "s1", Email: "geoff@example.com", State: State("bogus")})
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("Append error = %v, want errors.Is ErrInvalidState", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("Append with an invalid state created %s; it must reject before writing", path)
	}
}

func TestAppendWritesSingleJSONLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contributions.jsonl")
	l, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	at := time.Date(2026, 8, 31, 9, 30, 0, 0, time.UTC)
	want := Event{
		SessionID:  "abcdefghijklmnopqrstuvwxyz",
		Email:      "geoff@example.com",
		Album:      "Test Album",
		Collection: "family",
		State:      StateUploaded,
		At:         at,
		Detail:     "1 of 1 files received",
	}
	if err := l.Append(want); err != nil {
		t.Fatalf("Append: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.HasSuffix(string(raw), "\n") {
		t.Errorf("record is not newline-terminated: %q", raw)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1: %q", len(lines), raw)
	}

	var got Event
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.SessionID != want.SessionID || got.Email != want.Email || got.Album != want.Album ||
		got.Collection != want.Collection || got.State != want.State || got.Detail != want.Detail {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if !got.At.Equal(want.At) {
		t.Errorf("At = %v, want %v", got.At, want.At)
	}
	if len(got.LibraryPaths) != 0 {
		t.Errorf("LibraryPaths = %v, want empty", got.LibraryPaths)
	}
}

func TestHistoryMissingFileIsEmpty(t *testing.T) {
	dir := t.TempDir()
	l, err := Open(filepath.Join(dir, "contributions.jsonl"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	events, err := l.History("geoff@example.com")
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("got %d events against a missing ledger file, want 0", len(events))
	}

	sessionEvents, err := l.Session("some-session")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	if len(sessionEvents) != 0 {
		t.Errorf("got %d events against a missing ledger file, want 0", len(sessionEvents))
	}
}

func TestHistoryOrdersNewestFirstAndFiltersByEmail(t *testing.T) {
	dir := t.TempDir()
	l, err := Open(filepath.Join(dir, "contributions.jsonl"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	base := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	events := []Event{
		{SessionID: "s1", Email: "geoff@example.com", State: StateUploading, At: base},
		{SessionID: "s2", Email: "other@example.com", State: StateUploading, At: base.Add(time.Minute)},
		{SessionID: "s1", Email: "geoff@example.com", State: StateUploaded, At: base.Add(2 * time.Minute)},
		{SessionID: "s1", Email: "geoff@example.com", State: StateChecking, At: base.Add(3 * time.Minute)},
	}
	for _, e := range events {
		if err := l.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got, err := l.History("geoff@example.com")
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	wantStates := []State{StateChecking, StateUploaded, StateUploading}
	if len(got) != len(wantStates) {
		t.Fatalf("got %d events, want %d", len(got), len(wantStates))
	}
	for i, want := range wantStates {
		if got[i].State != want {
			t.Errorf("event %d state = %q, want %q", i, got[i].State, want)
		}
		if got[i].Email != "geoff@example.com" {
			t.Errorf("event %d email = %q; another contributor's event leaked into History", i, got[i].Email)
		}
	}
}

func TestSessionFiltersBySessionID(t *testing.T) {
	dir := t.TempDir()
	l, err := Open(filepath.Join(dir, "contributions.jsonl"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	base := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	for _, e := range []Event{
		{SessionID: "s1", Email: "a@example.com", State: StateUploading, At: base},
		{SessionID: "s2", Email: "b@example.com", State: StateUploading, At: base.Add(time.Minute)},
		{SessionID: "s1", Email: "a@example.com", State: StateTagging, At: base.Add(2 * time.Minute)},
	} {
		if err := l.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got, err := l.Session("s1")
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d events for session s1, want 2", len(got))
	}
	if got[0].State != StateTagging || got[1].State != StateUploading {
		t.Errorf("Session did not return s1's events newest first: %+v", got)
	}
	for _, e := range got {
		if e.SessionID != "s1" {
			t.Errorf("Session(%q) returned an event from session %q", "s1", e.SessionID)
		}
	}
}

// fakeSyncWriter is the syncWriter double used to assert appendLine's
// call shape directly: exactly one Write with the whole record,
// followed by Sync only when the write fully succeeded.
type fakeSyncWriter struct {
	writes    [][]byte
	writeErr  error
	shortBy   int
	syncCalls int
	syncErr   error
}

func (w *fakeSyncWriter) Write(p []byte) (int, error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	w.writes = append(w.writes, cp)
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return len(p) - w.shortBy, nil
}

func (w *fakeSyncWriter) Sync() error {
	w.syncCalls++
	return w.syncErr
}

func TestAppendLineSingleWriteCall(t *testing.T) {
	line := []byte(`{"state":"uploading"}` + "\n")
	tests := []struct {
		name      string
		writeErr  error
		shortBy   int
		wantErr   bool
		wantSyncs int
	}{
		{name: "writes once and syncs once", wantSyncs: 1},
		{name: "propagates write error without syncing", writeErr: errors.New("disk full"), wantErr: true, wantSyncs: 0},
		{name: "short write errors instead of retrying", shortBy: 2, wantErr: true, wantSyncs: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &fakeSyncWriter{writeErr: tt.writeErr, shortBy: tt.shortBy}
			err := appendLine(w, line)
			if (err != nil) != tt.wantErr {
				t.Fatalf("appendLine error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(w.writes) != 1 {
				t.Fatalf("Write called %d times, want exactly 1 (no short-write retry loop)", len(w.writes))
			}
			if tt.writeErr == nil && !bytes.Equal(w.writes[0], line) {
				t.Errorf("Write received %q, want the whole record %q in one call", w.writes[0], line)
			}
			if w.syncCalls != tt.wantSyncs {
				t.Errorf("Sync called %d times, want %d", w.syncCalls, tt.wantSyncs)
			}
		})
	}
}

// TestAppendConcurrentMultiWriter simulates serve and the importer
// appending to the same ledger path from separate, uncoordinated
// Ledger instances: the plan's multi-writer guarantee is that every
// record lands whole, in some order, never torn or dropped by another
// writer's concurrent record.
func TestAppendConcurrentMultiWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contributions.jsonl")

	const writers = 8
	const perWriter = 25

	var wg sync.WaitGroup
	for w := range writers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			l, err := Open(path)
			if err != nil {
				t.Errorf("Open: %v", err)
				return
			}
			for i := range perWriter {
				e := Event{
					SessionID: fmt.Sprintf("writer%d", w),
					Email:     "geoff@example.com",
					Album:     "Concurrent Album",
					State:     StateUploading,
					At:        time.Now(),
					Detail:    fmt.Sprintf("record %d from writer %d", i, w),
				}
				if err := l.Append(e); err != nil {
					t.Errorf("Append: %v", err)
				}
			}
		}(w)
	}
	wg.Wait()

	l, err := Open(path)
	if err != nil {
		t.Fatalf("Open for read: %v", err)
	}
	events, err := l.readAll()
	if err != nil {
		t.Fatalf("readAll: %v", err)
	}
	want := writers * perWriter
	if len(events) != want {
		t.Fatalf("got %d events, want %d (a lost or corrupted record means the interleave guarantee broke)", len(events), want)
	}
	seen := make(map[string]bool, want)
	for _, e := range events {
		key := e.SessionID + "|" + e.Detail
		if seen[key] {
			t.Errorf("duplicate record %s (a torn write merged two records)", key)
		}
		seen[key] = true
	}
}

// realisticMultiDiscPaths builds a beets-resolved LibraryPaths list
// shaped like a real multi-disc box set: long album and track titles
// under per-disc subdirectories, the case the plan requires the
// degradation test to exercise instead of a synthetic short fixture.
func realisticMultiDiscPaths() []string {
	root := "/srv/musicbox/library/Various Artists/A Very Long Box Set Title (Deluxe Anniversary Edition) [2024]"
	var paths []string
	for disc := 1; disc <= 3; disc++ {
		for track := 1; track <= 18; track++ {
			paths = append(paths, fmt.Sprintf(
				"%s/CD%d/%02d - Some Reasonably Long Track Title That Real Releases Actually Have.flac",
				root, disc, track))
		}
	}
	return paths
}

func TestEncodeRecordDegradesLibraryPathsOnMultiDiscRelease(t *testing.T) {
	paths := realisticMultiDiscPaths()
	e := Event{
		SessionID:    "abcdefghijklmnopqrstuvwxyz",
		Email:        "geoff@example.com",
		Album:        "A Very Long Box Set Title (Deluxe Anniversary Edition)",
		Collection:   "family",
		State:        StateTagging,
		At:           time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC),
		Detail:       "imported",
		LibraryPaths: paths,
	}
	rawLine, err := marshalRecord(e)
	if err != nil {
		t.Fatalf("marshalRecord: %v", err)
	}
	if len(rawLine) <= maxRecordBytes {
		t.Fatalf("test fixture is only %d bytes; strengthen it so degradation actually triggers", len(rawLine))
	}

	line, err := encodeRecord(e)
	if err != nil {
		t.Fatalf("encodeRecord: %v", err)
	}
	if len(line) > maxRecordBytes {
		t.Fatalf("degraded record is %d bytes, want <= %d", len(line), maxRecordBytes)
	}

	var got Event
	if err := json.Unmarshal(bytes.TrimRight(line, "\n"), &got); err != nil {
		t.Fatalf("unmarshal degraded record: %v", err)
	}
	if got.Detail != e.Detail {
		t.Errorf("Detail degraded to %q; the LibraryPaths collapse alone should have been enough", got.Detail)
	}
	if len(got.LibraryPaths) != 2 {
		t.Fatalf("degraded LibraryPaths has %d elements, want 2 (album root, track count)", len(got.LibraryPaths))
	}
	wantRoot := "/srv/musicbox/library/Various Artists/A Very Long Box Set Title (Deluxe Anniversary Edition) [2024]"
	if got.LibraryPaths[0] != wantRoot {
		t.Errorf("album root = %q, want %q", got.LibraryPaths[0], wantRoot)
	}
	if got.LibraryPaths[1] != strconv.Itoa(len(paths)) {
		t.Errorf("track count = %q, want %q", got.LibraryPaths[1], strconv.Itoa(len(paths)))
	}
}

func TestEncodeRecordTruncatesDetailWhenCollapseIsNotEnough(t *testing.T) {
	hugeDetail := strings.Repeat("x", 6000)
	e := Event{
		SessionID:    "abcdefghijklmnopqrstuvwxyz",
		Email:        "geoff@example.com",
		Album:        "Album",
		State:        StateReview,
		At:           time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC),
		Detail:       hugeDetail,
		LibraryPaths: realisticMultiDiscPaths(),
	}

	line, err := encodeRecord(e)
	if err != nil {
		t.Fatalf("encodeRecord: %v", err)
	}
	if len(line) > maxRecordBytes {
		t.Fatalf("record is still %d bytes after degradation, want <= %d", len(line), maxRecordBytes)
	}

	var got Event
	if err := json.Unmarshal(bytes.TrimRight(line, "\n"), &got); err != nil {
		t.Fatalf("unmarshal degraded record: %v", err)
	}
	if len(got.Detail) >= len(hugeDetail) {
		t.Errorf("Detail was not truncated: got %d chars, started with %d", len(got.Detail), len(hugeDetail))
	}
	if len(got.LibraryPaths) != 2 {
		t.Errorf("LibraryPaths not collapsed: %d elements", len(got.LibraryPaths))
	}
}

func TestEncodeRecordNeverRejectsAPathologicallyOversizedEvent(t *testing.T) {
	e := Event{
		SessionID: "abcdefghijklmnopqrstuvwxyz",
		Email:     strings.Repeat("a", 10000) + "@example.com",
		Album:     "Album",
		State:     StateRejected,
		At:        time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC),
	}

	line, err := encodeRecord(e)
	if err != nil {
		t.Fatalf("encodeRecord returned an error instead of a best-effort record: %v", err)
	}
	if len(line) <= maxRecordBytes {
		t.Fatalf("test fixture didn't actually exceed the ceiling even after full degradation; strengthen it")
	}

	var got Event
	if err := json.Unmarshal(bytes.TrimRight(line, "\n"), &got); err != nil {
		t.Fatalf("degraded-but-still-oversized record is not valid JSON: %v", err)
	}
	if got.Email != e.Email {
		t.Errorf("terminal event content was altered instead of written best-effort: got %q", got.Email)
	}
}

// recordingHandler is a minimal slog.Handler that counts warnings, used
// to prove readAll actually logs when it skips a torn line rather than
// silently dropping it.
type recordingHandler struct {
	warnings int
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level == slog.LevelWarn {
		h.warnings++
	}
	return nil
}

func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }

func TestReadAllSkipsTornFinalLineWithLoggedWarning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contributions.jsonl")
	l, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	base := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	if err := l.Append(Event{SessionID: "s1", Email: "geoff@example.com", State: StateUploading, At: base}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := l.Append(Event{SessionID: "s1", Email: "geoff@example.com", State: StateUploaded, At: base.Add(time.Minute)}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// Simulate a crash mid-write: cut the file off partway through the
	// final record, leaving a torn, unparseable last line with no
	// trailing newline.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(raw) <= 10 {
		t.Fatalf("fixture record is only %d bytes; too short to truncate meaningfully", len(raw))
	}
	torn := raw[:len(raw)-10]
	if err := os.WriteFile(path, torn, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	handler := &recordingHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(prev)

	events, err := l.History("geoff@example.com")
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events after a torn final line, want 1 (the earlier, intact record)", len(events))
	}
	if events[0].State != StateUploading {
		t.Errorf("surviving event state = %q, want %q", events[0].State, StateUploading)
	}
	if handler.warnings == 0 {
		t.Errorf("readAll did not log a warning for the torn final line")
	}
}
