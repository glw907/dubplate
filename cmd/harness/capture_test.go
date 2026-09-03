package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/stub"
)

// fakeFetch resolves a pathQuery to a canned response or an error,
// recording every call it saw. Every capture_test.go test installs one
// through this seam instead of touching the network: no live network
// in go test, ever.
type fakeFetch struct {
	responses map[string][]byte
	err       error
	calls     []struct{ pathQuery, userAgent string }
}

func (f *fakeFetch) fetch(_ context.Context, pathQuery, userAgent string) ([]byte, error) {
	f.calls = append(f.calls, struct{ pathQuery, userAgent string }{pathQuery, userAgent})
	if f.err != nil {
		return nil, f.err
	}
	body, ok := f.responses[pathQuery]
	if !ok {
		return nil, errors.New("fakeFetch: no response stubbed for " + pathQuery)
	}
	return body, nil
}

// withFetch installs f as doFetch for the duration of the test.
func withFetch(t *testing.T, f *fakeFetch) {
	t.Helper()
	orig := doFetch
	doFetch = f.fetch
	t.Cleanup(func() { doFetch = orig })
}

// withSleep replaces sleep with a call counter and returns it, so a
// test can prove the pacing contract without waiting on real time.
func withSleep(t *testing.T) *[]time.Duration {
	t.Helper()
	var calls []time.Duration
	orig := sleep
	sleep = func(d time.Duration) { calls = append(calls, d) }
	t.Cleanup(func() { sleep = orig })
	return &calls
}

// --- request planning ---

func TestReleaseRequestIsCanonical(t *testing.T) {
	got := releaseRequest("00000000-0000-0000-0000-000000000000")
	want := "/ws/2/release/00000000-0000-0000-0000-000000000000?fmt=json&inc=" + url.QueryEscape(maximalReleaseInclude)
	if got != want {
		t.Errorf("releaseRequest = %q, want %q", got, want)
	}
}

func TestPseudoReleaseFollowUpDetectsBackwardRelation(t *testing.T) {
	body := []byte(`{
		"status": "Pseudo-Release",
		"relations": [
			{"type": "artist-rels", "direction": "forward", "release": {"id": "wrong-relation"}},
			{"type": "transl-tracklisting", "direction": "backward", "release": {"id": "actual-release-id"}}
		]
	}`)
	mbid, ok := pseudoReleaseFollowUp(body)
	if !ok {
		t.Fatal("pseudoReleaseFollowUp: ok = false, want true")
	}
	if mbid != "actual-release-id" {
		t.Errorf("pseudoReleaseFollowUp mbid = %q, want %q", mbid, "actual-release-id")
	}
}

func TestPseudoReleaseFollowUpIgnoresForwardDirection(t *testing.T) {
	body := []byte(`{
		"status": "Pseudo-Release",
		"relations": [
			{"type": "transl-tracklisting", "direction": "forward", "release": {"id": "should-not-follow"}}
		]
	}`)
	if _, ok := pseudoReleaseFollowUp(body); ok {
		t.Error("pseudoReleaseFollowUp followed a forward relation, want only backward")
	}
}

func TestPseudoReleaseFollowUpIgnoresNonPseudoStatus(t *testing.T) {
	body := []byte(`{
		"status": "Official",
		"relations": [
			{"type": "transl-tracklisting", "direction": "backward", "release": {"id": "irrelevant"}}
		]
	}`)
	if _, ok := pseudoReleaseFollowUp(body); ok {
		t.Error("pseudoReleaseFollowUp followed a relation on a non-Pseudo-Release status")
	}
}

func TestPseudoReleaseFollowUpRejectsMalformedJSON(t *testing.T) {
	if _, ok := pseudoReleaseFollowUp([]byte("not json")); ok {
		t.Error("pseudoReleaseFollowUp: ok = true on malformed body, want false")
	}
}

// --- runCapture: canonical-form writing and wire discipline ---

func TestRunCaptureWritesVerbatimResponseAndCanonicalLog(t *testing.T) {
	dir := t.TempDir()
	mbid := "00000000-0000-0000-0000-000000000000"
	body := []byte(`{"id":"00000000-0000-0000-0000-000000000000","status":"Official"}`)

	f := &fakeFetch{responses: map[string][]byte{releaseRequest(mbid): body}}
	withFetch(t, f)

	if err := runCapture(context.Background(), "1.2.3", "geoff@example.org", mbid, filepath.Join(dir, "case-a")); err != nil {
		t.Fatalf("runCapture: %v", err)
	}

	if _, err := cassette.LoadCase(filepath.Join(dir, "case-a")); err == nil {
		t.Fatal("LoadCase succeeded on a seeded (Profile/EvalDate-less) case; want a validation error until a human fills those in")
	}
	reqs := seededRequests(t, filepath.Join(dir, "case-a"))
	if !hasRequest(reqs, releaseRequest(mbid)) {
		t.Errorf("case.json requests %+v do not carry the canonical release request %q", reqs, releaseRequest(mbid))
	}

	got, err := os.ReadFile(filepath.Join(dir, "case-a", "responses", "001.json"))
	if err != nil {
		t.Fatalf("read response file: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("response body = %q, want the fetched bytes verbatim %q", got, body)
	}

	if len(f.calls) != 1 {
		t.Fatalf("doFetch calls = %d, want 1 (no pseudo-release relation)", len(f.calls))
	}
	if f.calls[0].userAgent != "dubplate-harness/1.2.3 ( geoff@example.org )" {
		t.Errorf("User-Agent = %q, want the identified form", f.calls[0].userAgent)
	}
}

func TestRunCaptureFollowsPseudoReleaseAndCapturesBothResponses(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "case-pseudo")
	mbid := "pseudo-release-id"
	actualID := "actual-release-id"

	pseudoBody := []byte(`{"id":"pseudo-release-id","status":"Pseudo-Release","relations":[` +
		`{"type":"transl-tracklisting","direction":"backward","release":{"id":"actual-release-id"}}]}`)
	actualBody := []byte(`{"id":"actual-release-id","status":"Official"}`)

	f := &fakeFetch{responses: map[string][]byte{
		releaseRequest(mbid):     pseudoBody,
		releaseRequest(actualID): actualBody,
	}}
	withFetch(t, f)
	calls := withSleep(t)

	if err := runCapture(context.Background(), "1.0.0", "geoff@example.org", mbid, dir); err != nil {
		t.Fatalf("runCapture: %v", err)
	}

	reqs := seededRequests(t, dir)
	if !hasRequest(reqs, releaseRequest(mbid)) || !hasRequest(reqs, releaseRequest(actualID)) {
		t.Errorf("case.json requests %+v are missing one of the two pseudo-release requests", reqs)
	}

	first, err := os.ReadFile(filepath.Join(dir, "responses", "001.json"))
	if err != nil || !bytes.Equal(first, pseudoBody) {
		t.Errorf("responses/001.json = %q, %v; want the pseudo-release body verbatim", first, err)
	}
	second, err := os.ReadFile(filepath.Join(dir, "responses", "002.json"))
	if err != nil || !bytes.Equal(second, actualBody) {
		t.Errorf("responses/002.json = %q, %v; want the follow-up body verbatim", second, err)
	}

	if len(f.calls) != 2 {
		t.Fatalf("doFetch calls = %d, want 2 (release plus pseudo-release follow-up)", len(f.calls))
	}
	// Two requests: the first pays no wait, the second waits
	// requestInterval, proving the one-request-per-second discipline.
	if len(*calls) != 1 || (*calls)[0] != requestInterval {
		t.Errorf("sleep calls = %v, want exactly one call of %v", *calls, requestInterval)
	}
}

func TestRunCaptureSurfacesFetchError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "case-err")
	f := &fakeFetch{err: errors.New("connection refused")}
	withFetch(t, f)

	err := runCapture(context.Background(), "1.0.0", "geoff@example.org", "some-id", dir)
	if err == nil {
		t.Fatal("runCapture with a failing fetch: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error %q does not name the underlying fetch failure", err)
	}
}

// --- runCaptureMisses: the --case verb ---

// seedValidCase writes a minimal but LoadCase-valid case (Profile and
// EvalDate set) at dir, with one existing logged request and response,
// so runCaptureMisses has something to append to.
func seedValidCase(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), []byte(`{"existing":true}`), 0o644); err != nil {
		t.Fatalf("write existing response: %v", err)
	}
	c := &cassette.Case{
		ID:       filepath.Base(dir),
		Profile:  "estate",
		EvalDate: "2024-01-15",
		Requests: []cassette.Request{
			{Method: http.MethodGet, PathQuery: "/ws/2/release/existing?fmt=json", Response: "responses/001.json"},
		},
	}
	if err := c.Write(dir); err != nil {
		t.Fatalf("write case: %v", err)
	}
}

// seededRequests decodes dir/case.json's request log directly,
// bypassing cassette.LoadCase's validation: a case runCapture just
// seeded has no Profile or EvalDate yet (a human fills those in), so
// LoadCase rejects it even though case.json itself is well-formed.
func seededRequests(t *testing.T, dir string) []cassette.Request {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatalf("read case.json: %v", err)
	}
	var c cassette.Case
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("parse case.json: %v", err)
	}
	return c.Requests
}

// hasRequest reports whether reqs carries an entry matching pathQuery.
func hasRequest(reqs []cassette.Request, pathQuery string) bool {
	for _, r := range reqs {
		if r.PathQuery == pathQuery {
			return true
		}
	}
	return false
}

func writeMissFileFixture(t *testing.T, dir string, misses []stub.Miss) {
	t.Helper()
	data, err := json.Marshal(misses)
	if err != nil {
		t.Fatalf("marshal misses fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, missFileName), data, 0o644); err != nil {
		t.Fatalf("write misses fixture: %v", err)
	}
}

func TestRunCaptureMissesAppendsRequestsAndConsumesMissFile(t *testing.T) {
	dir := t.TempDir()
	seedValidCase(t, dir)
	misses := []stub.Miss{
		{Method: http.MethodGet, PathQuery: "/ws/2/release/first-miss?fmt=json"},
		{Method: http.MethodGet, PathQuery: "/ws/2/release/second-miss?fmt=json"},
	}
	writeMissFileFixture(t, dir, misses)

	f := &fakeFetch{responses: map[string][]byte{
		"/ws/2/release/first-miss?fmt=json":  []byte(`{"id":"first-miss"}`),
		"/ws/2/release/second-miss?fmt=json": []byte(`{"id":"second-miss"}`),
	}}
	withFetch(t, f)
	withSleep(t) // pacing is proven separately; this test never waits on real time

	n, err := runCaptureMisses(context.Background(), "1.0.0", "geoff@example.org", dir)
	if err != nil {
		t.Fatalf("runCaptureMisses: %v", err)
	}
	if n != 2 {
		t.Errorf("runCaptureMisses count = %d, want 2", n)
	}

	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("reload case: %v", err)
	}
	if len(c.Requests) != 3 {
		t.Fatalf("Requests = %d, want 3 (1 existing plus 2 captured)", len(c.Requests))
	}
	for _, miss := range misses {
		found := false
		for _, req := range c.Requests {
			if req.PathQuery == miss.PathQuery && req.Method == miss.Method {
				found = true
			}
		}
		if !found {
			t.Errorf("Requests does not contain miss %+v", miss)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, missFileName)); !os.IsNotExist(err) {
		t.Errorf("misses.json still present after a successful capture: %v", err)
	}
}

func TestRunCaptureMissesNoOpWithoutMissFile(t *testing.T) {
	dir := t.TempDir()
	seedValidCase(t, dir)

	f := &fakeFetch{}
	withFetch(t, f)

	n, err := runCaptureMisses(context.Background(), "1.0.0", "geoff@example.org", dir)
	if err != nil {
		t.Fatalf("runCaptureMisses: %v", err)
	}
	if n != 0 {
		t.Errorf("runCaptureMisses count = %d, want 0", n)
	}
	if len(f.calls) != 0 {
		t.Errorf("doFetch was called %d times with no miss file present, want 0", len(f.calls))
	}
}

func TestRunCaptureMissesLeavesCaseUnchangedOnFetchError(t *testing.T) {
	dir := t.TempDir()
	seedValidCase(t, dir)
	writeMissFileFixture(t, dir, []stub.Miss{{Method: http.MethodGet, PathQuery: "/ws/2/release/broken?fmt=json"}})

	f := &fakeFetch{err: errors.New("timeout")}
	withFetch(t, f)

	if _, err := runCaptureMisses(context.Background(), "1.0.0", "geoff@example.org", dir); err == nil {
		t.Fatal("runCaptureMisses with a failing fetch: got nil error, want one")
	}

	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("reload case: %v", err)
	}
	if len(c.Requests) != 1 {
		t.Errorf("Requests = %d, want 1 (unchanged after a failed capture)", len(c.Requests))
	}
	if _, err := os.Stat(filepath.Join(dir, missFileName)); err != nil {
		t.Errorf("misses.json missing after a failed capture, want it left in place: %v", err)
	}
}

// --- contact refusal ---

func TestRequireContactMissing(t *testing.T) {
	_, err := requireContact(nil)
	if err == nil {
		t.Fatal("requireContact(nil): got nil error, want one")
	}
	if !strings.Contains(err.Error(), contactEnvVar) {
		t.Errorf("error %q does not name %q", err, contactEnvVar)
	}
}

func TestRequireContactEmptyValueTreatedAsUnset(t *testing.T) {
	_, err := requireContact([]string{contactEnvVar + "="})
	if err == nil {
		t.Fatal("requireContact with an empty value: got nil error, want one")
	}
}

func TestRequireContactPresent(t *testing.T) {
	got, err := requireContact([]string{"PATH=/usr/bin", contactEnvVar + "=geoff@example.org"})
	if err != nil {
		t.Fatalf("requireContact: %v", err)
	}
	if got != "geoff@example.org" {
		t.Errorf("requireContact = %q, want %q", got, "geoff@example.org")
	}
}

func TestCaptureCmdRefusesWithoutContact(t *testing.T) {
	f := &fakeFetch{}
	withFetch(t, f)

	dir := t.TempDir()
	cmd := newCaptureCmd(nil, "1.0.0")
	cmd.SetArgs([]string{"--mbid", "some-id", "--out", filepath.Join(dir, "case")})
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("capture without DUBPLATE_HARNESS_CONTACT: got nil error, want one")
	}
	if !strings.Contains(err.Error(), contactEnvVar) {
		t.Errorf("error %q does not name %q", err, contactEnvVar)
	}
	if len(f.calls) != 0 {
		t.Errorf("doFetch was called %d times before the contact check, want 0", len(f.calls))
	}
}

func TestCaptureCmdFlagValidation(t *testing.T) {
	environ := []string{contactEnvVar + "=geoff@example.org"}
	tests := []struct {
		name   string
		args   []string
		errHas string
	}{
		{"neither flag", nil, "one of --mbid or --case is required"},
		{"both flags", []string{"--mbid", "x", "--case", "y"}, "mutually exclusive"},
		{"mbid without out", []string{"--mbid", "x"}, "--out is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeFetch{}
			withFetch(t, f)

			cmd := newCaptureCmd(environ, "1.0.0")
			cmd.SetArgs(tt.args)
			var stderr bytes.Buffer
			cmd.SetErr(&stderr)

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("Execute(%v): got nil error, want one containing %q", tt.args, tt.errHas)
			}
			if !strings.Contains(err.Error(), tt.errHas) {
				t.Errorf("error = %q, want it to contain %q", err, tt.errHas)
			}
			if len(f.calls) != 0 {
				t.Errorf("doFetch was called %d times on a flag-validation failure, want 0", len(f.calls))
			}
		})
	}
}

// --- writer/loader round trip against the stub's canonical lookup ---

// TestWrittenCaseRoundTripsThroughLoadCaseAndStub is the review's
// minor 7: a response set produced by this tool's own writer
// (runCapture plus writeResponse for a manually appended request) must
// load through cassette.LoadCase and match the stub's canonical
// lookup byte-for-byte, including a percent-encoded Lucene-escaped
// search query and a percent-encoded path segment. The latter proves
// the accepted review finding that the stub must key lookups on the
// request's raw form (r.URL.EscapedPath()), not Go's already-decoded
// r.URL.Path: keying on the decoded form would turn "a%2Bb" into
// "a+b" and this request would come back a 404 instead of a hit.
func TestWrittenCaseRoundTripsThroughLoadCaseAndStub(t *testing.T) {
	dir := t.TempDir()
	mbid := "00000000-0000-0000-0000-000000000000"
	actualID := "actual-release-id"

	pseudoBody := []byte(`{"id":"00000000-0000-0000-0000-000000000000","status":"Pseudo-Release","relations":[` +
		`{"type":"transl-tracklisting","direction":"backward","release":{"id":"actual-release-id"}}]}`)
	actualBody := []byte(`{"id":"actual-release-id","status":"Official"}`)

	f := &fakeFetch{responses: map[string][]byte{
		releaseRequest(mbid):     pseudoBody,
		releaseRequest(actualID): actualBody,
	}}
	withFetch(t, f)
	withSleep(t) // pacing is proven separately; this test never waits on real time

	if err := runCapture(context.Background(), "1.0.0", "geoff@example.org", mbid, dir); err != nil {
		t.Fatalf("runCapture: %v", err)
	}

	c := &cassette.Case{
		ID:       filepath.Base(dir),
		Profile:  "estate",
		EvalDate: "2024-01-15",
	}
	raw, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatalf("read seeded case.json: %v", err)
	}
	if err := json.Unmarshal(raw, c); err != nil {
		t.Fatalf("parse seeded case.json: %v", err)
	}
	c.Profile = "estate"
	c.EvalDate = "2024-01-15"

	// A Lucene-escaped search query (the sample from cassette's own
	// canonicalization test: "Sigur Rós" with an escaped colon and
	// literal-plus space), plus a request whose path segment itself
	// needs percent-encoding.
	luceneQuery := `/ws/2/release?fmt=json&query=artist%3A%22Sigur%5C+R%C3%B3s%22`
	luceneBody := []byte(`{"releases":[]}`)
	respName, err := writeResponse(dir, 3, luceneBody)
	if err != nil {
		t.Fatalf("writeResponse (lucene): %v", err)
	}
	c.Requests = append(c.Requests, cassette.Request{Method: http.MethodGet, PathQuery: luceneQuery, Response: respName})

	encodedPathQuery := "/ws/2/release/a%2Bb?fmt=json"
	encodedPathBody := []byte(`{"id":"a+b"}`)
	respName, err = writeResponse(dir, 4, encodedPathBody)
	if err != nil {
		t.Fatalf("writeResponse (encoded path): %v", err)
	}
	c.Requests = append(c.Requests, cassette.Request{Method: http.MethodGet, PathQuery: encodedPathQuery, Response: respName})

	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}

	loaded, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase: %v", err)
	}
	if len(loaded.Requests) != 4 {
		t.Fatalf("loaded Requests = %d, want 4", len(loaded.Requests))
	}

	srv, err := stub.New(loaded, dir)
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })

	for _, req := range loaded.Requests {
		want, err := os.ReadFile(filepath.Join(dir, req.Response))
		if err != nil {
			t.Fatalf("read expected body for %s: %v", req.PathQuery, err)
		}

		resp, err := http.Get("http://" + srv.Addr() + req.PathQuery) //nolint:gosec,noctx // fixed loopback test address
		if err != nil {
			t.Fatalf("GET %s: %v", req.PathQuery, err)
		}
		got, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("read response body for %s: %v", req.PathQuery, err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200 (a stored request must hit, never a canonicalization miss)", req.PathQuery, resp.StatusCode)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("GET %s: body = %q, want %q", req.PathQuery, got, want)
		}
	}

	if misses := srv.Misses(); len(misses) != 0 {
		t.Errorf("stub recorded misses %v for a case where every request was logged", misses)
	}
}
