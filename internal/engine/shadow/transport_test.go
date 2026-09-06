package shadow

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// testUserAgent is the identity every test in this file builds a
// *musicbrainz.Client with, on internal/engine/tag/tag_test.go's own
// precedent.
const testUserAgent = "dubplate-test/0 ( test )"

// countingTransport counts every RoundTrip call it forwards to base, so
// a test can prove a 501 refusal costs exactly one attempt.
type countingTransport struct {
	base  http.RoundTripper
	calls int
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.calls++
	return t.base.RoundTrip(req)
}

func TestRecordingTransportRecordsInFlowOrder(t *testing.T) {
	bodies := map[string][]byte{
		"/search":  []byte(`{"n":"search"}`),
		"/lookup1": []byte(`{"n":"lookup1"}`),
		"/lookup2": []byte(`{"n":"lookup2"}`),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(bodies[r.URL.Path]) //nolint:errcheck // best-effort write to a test stub
	}))
	t.Cleanup(srv.Close)

	rec := &recordingTransport{base: http.DefaultTransport}
	client := &http.Client{Transport: rec}

	paths := []string{"/search?a=1&b=2", "/lookup1", "/lookup2"}
	var gotBodies [][]byte
	for _, p := range paths {
		resp, err := client.Get(srv.URL + p)
		if err != nil {
			t.Fatalf("GET %s: %v", p, err)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body for %s: %v", p, err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("close body for %s: %v", p, err)
		}
		gotBodies = append(gotBodies, b)
	}

	if len(rec.requests) != 3 {
		t.Fatalf("recorded %d requests, want 3", len(rec.requests))
	}
	for i, want := range paths {
		if rec.requests[i].Method != http.MethodGet {
			t.Errorf("request %d Method = %q, want GET", i, rec.requests[i].Method)
		}
		if rec.requests[i].PathQuery != want {
			t.Errorf("request %d PathQuery = %q, want %q (already sorted by key)", i, rec.requests[i].PathQuery, want)
		}
	}

	wantBodies := []string{"/search", "/lookup1", "/lookup2"}
	if len(rec.bodies) != 3 {
		t.Fatalf("recorded %d bodies, want 3", len(rec.bodies))
	}
	for i, key := range wantBodies {
		if string(rec.bodies[i]) != string(bodies[key]) {
			t.Errorf("stored body %d = %s, want %s", i, rec.bodies[i], bodies[key])
		}
		if string(gotBodies[i]) != string(bodies[key]) {
			t.Errorf("client-visible body %d = %s, want %s (the recorder must rewind what it reads)", i, gotBodies[i], bodies[key])
		}
	}
}

func TestReplayTransportRefusesAMissWithoutA404(t *testing.T) {
	const knownID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	const missID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	srv := releaseServer(t, map[string][]byte{
		knownID: releaseBody(knownID, "Windrose", "Cabinet", []string{"Track One"}),
	}, nil)

	rec := &recordingTransport{base: http.DefaultTransport}
	recordClient, err := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   srv.URL,
		UserAgent: testUserAgent,
		Budget:    musicbrainz.Unmetered(),
		HTTP:      &http.Client{Transport: rec},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := recordClient.Release(context.Background(), knownID); err != nil {
		t.Fatalf("recording the known release: %v", err)
	}
	if len(rec.requests) != 1 {
		t.Fatalf("recorded %d requests, want 1", len(rec.requests))
	}

	replay := &replayTransport{bodies: map[string][]byte{rec.requests[0].PathQuery: rec.bodies[0]}}
	counting := &countingTransport{base: replay}
	replayClient, err := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   "http://replay.invalid",
		UserAgent: testUserAgent,
		Budget:    musicbrainz.Unmetered(),
		HTTP:      &http.Client{Transport: counting},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = replayClient.Release(context.Background(), missID)
	if err == nil {
		t.Fatal("Release: want an error on a replay miss")
	}
	if errors.Is(err, musicbrainz.ErrNotFound) {
		t.Error("Release error wraps ErrNotFound: a miss must never look like a 404 (contract S18)")
	}
	if errors.Is(err, musicbrainz.ErrUnavailable) {
		t.Error("Release error wraps ErrUnavailable: 501 must be non-retryable, not exhausted retries")
	}
	if len(replay.misses) != 1 {
		t.Fatalf("misses = %v, want exactly one", replay.misses)
	}
	if replay.misses[0] == rec.requests[0].PathQuery {
		t.Errorf("misses[0] = %q, want the MISSED request's URI, not the known one that was served", replay.misses[0])
	}
	if counting.calls != 1 {
		t.Errorf("attempted %d requests, want exactly 1: a 501 must not be retried", counting.calls)
	}
}
