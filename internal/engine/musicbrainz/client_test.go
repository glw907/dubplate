package musicbrainz

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// roundTripperFunc adapts a function to http.RoundTripper, so a test
// scripts a transport without a fake server. No test here reaches the
// network.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// readClientTestdataFile reads testdata/client/name, panicking on
// failure: it backs package-level fixture variables initialized
// before any test's *testing.T exists, the same reason
// releaseFixture cannot use the package's own readTestdata helper.
func readClientTestdataFile(name string) []byte {
	data, err := os.ReadFile(filepath.Join("testdata", "client", name))
	if err != nil {
		panic("reading testdata/client/" + name + ": " + err.Error())
	}
	return data
}

// releaseFixture is testdata/client/release_minimal.json, byte for
// byte, and releaseFixtureID is that body's own release id.
var releaseFixture = string(readClientTestdataFile("release_minimal.json"))

const releaseFixtureID = "11111111-1111-1111-1111-111111111111"

// noSleep is the sleep seam for tests that do not exercise the backoff
// schedule at all: it never spends wall time and never records
// anything.
func noSleep(context.Context, time.Duration) error { return nil }

// newTestClient builds a Client against an unmetered budget and a
// scripted transport, for tests that are not exercising the limiter's
// own per-attempt token.
func newTestClient(t *testing.T, rt http.RoundTripper) *Client {
	t.Helper()
	c, err := NewClient(ClientConfig{
		BaseURL:   "http://stub.invalid",
		UserAgent: "dubplate-test/0 ( test )",
		Budget:    Unmetered(),
		HTTP:      &http.Client{Transport: rt},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.sleep = noSleep
	return c
}

// Mutation 1: NewClient accepts an empty UserAgent.
func TestNewClientRequiresUserAgent(t *testing.T) {
	if _, err := NewClient(ClientConfig{Budget: Unmetered()}); err == nil {
		t.Fatal("NewClient with empty UserAgent = nil error, want one")
	}
}

// Mutation 2: NewClient accepts a nil Budget.
func TestNewClientRequiresBudget(t *testing.T) {
	if _, err := NewClient(ClientConfig{UserAgent: "dubplate-test/0 ( test )"}); err == nil {
		t.Fatal("NewClient with nil Budget = nil error, want one")
	}
}

func TestReleaseDecodesSuccessfulResponseAtTheCommittedURLShape(t *testing.T) {
	var gotURL string
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		gotURL = req.URL.String()
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(releaseFixture))}, nil
	})
	c := newTestClient(t, rt)

	rel, err := c.Release(context.Background(), releaseFixtureID)
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	if rel.ID != releaseFixtureID {
		t.Errorf("rel.ID = %q, want %q", rel.ID, releaseFixtureID)
	}

	wantURL := "http://stub.invalid/ws/2/release/" + releaseFixtureID +
		"?fmt=json&inc=" + url.QueryEscape(strings.Join(releaseIncludes, "+"))
	if gotURL != wantURL {
		t.Errorf("request URL = %q, want %q", gotURL, wantURL)
	}
}

func TestReleaseDecodeErrorWrapsErrMalformedRelease(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{not valid json`))}, nil
	})
	c := newTestClient(t, rt)

	_, err := c.Release(context.Background(), releaseFixtureID)
	if !errors.Is(err, ErrMalformedRelease) {
		t.Fatalf("Release() = %v, want an error wrapping ErrMalformedRelease", err)
	}
}

// Mutation 4: exhausted retries return ErrNotFound instead of
// ErrUnavailable.
func TestReleaseRetryExhaustionReturnsErrUnavailable(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	c := newTestClient(t, rt)

	_, err := c.Release(context.Background(), releaseFixtureID)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Release() = %v, want an error wrapping ErrUnavailable", err)
	}
}

// Mutation 5: backoff computes 2^n instead of 2^(n-1).
func TestBackoffSchedule(t *testing.T) {
	want := []time.Duration{0, time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second}
	for i, w := range want {
		attempt := i + 1
		if got := backoff(attempt); got != w {
			t.Errorf("backoff(%d) = %v, want %v", attempt, got, w)
		}
	}
}

// Mutation 6: a Retry-After header carrying a delta-seconds value is
// ignored.
func TestRetryAfterDeltaSecondsHonored(t *testing.T) {
	var attempts int
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return &http.Response{
				StatusCode: 503,
				Header:     http.Header{"Retry-After": {"3"}},
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(releaseFixture))}, nil
	})
	c := newTestClient(t, rt)
	var slept []time.Duration
	c.sleep = func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil }

	if _, err := c.Release(context.Background(), releaseFixtureID); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if want := []time.Duration{3 * time.Second}; !slices.Equal(slept, want) {
		t.Errorf("slept = %v, want %v (the Retry-After header, not the computed backoff)", slept, want)
	}
}

// Mutation 6b: an HTTP-date Retry-After is parsed as seconds, or
// parsed at all, rather than falling through to the computed backoff
// (fold M-c).
func TestRetryAfterHTTPDateFallsThroughToComputedBackoff(t *testing.T) {
	var attempts int
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return &http.Response{
				StatusCode: 503,
				Header:     http.Header{"Retry-After": {"Wed, 21 Oct 2099 07:28:00 GMT"}},
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(releaseFixture))}, nil
	})
	c := newTestClient(t, rt)
	var slept []time.Duration
	c.sleep = func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil }

	if _, err := c.Release(context.Background(), releaseFixtureID); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if want := []time.Duration{0}; !slices.Equal(slept, want) {
		t.Errorf("slept = %v, want %v (the computed backoff; an HTTP-date Retry-After must never be parsed)", slept, want)
	}
}

// Mutation 7: extractReleaseID removed (or bypassed) issues a request
// for a junk id instead of costing none.
func TestExtractReleaseIDRejectsJunkAndIssuesNoRequest(t *testing.T) {
	var requests int
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(releaseFixture))}, nil
	})
	c := newTestClient(t, rt)

	if _, err := c.Release(context.Background(), "not-a-release-id"); err == nil {
		t.Fatal("Release(junk id) = nil error, want one")
	}
	if requests != 0 {
		t.Errorf("requests = %d, want 0: a junk id must cost no request, exactly as album_for_id costs none", requests)
	}
}

func TestExtractReleaseID(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		wantID string
		wantOK bool
	}{
		{"bare id", releaseFixtureID, releaseFixtureID, true},
		{"embedded in a URL", "https://musicbrainz.org/release/" + releaseFixtureID, releaseFixtureID, true},
		{"junk", "not-a-release-id", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := extractReleaseID(tt.in)
			if id != tt.wantID || ok != tt.wantOK {
				t.Errorf("extractReleaseID(%q) = (%q, %v), want (%q, %v)", tt.in, id, ok, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestSearchReleasesDecodesTheObjectShapeAndPreservesOrder(t *testing.T) {
	body := string(readClientTestdataFile("search_two_hits.json"))
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	c := newTestClient(t, rt)

	ids, err := c.SearchReleases(context.Background(), AlbumSearch{Query: "release:(x)", Limit: 5})
	if err != nil {
		t.Fatalf("SearchReleases: %v", err)
	}
	want := []string{"22222222-2222-2222-2222-222222222222", "33333333-3333-3333-3333-333333333333"}
	if !slices.Equal(ids, want) {
		t.Errorf("ids = %v, want %v", ids, want)
	}
}

// Mutation 9: AlbumForID skips the pseudo-release's second fetch.
func TestAlbumForIDMergesThePseudoAndActualRelease(t *testing.T) {
	pseudoBody := string(readClientTestdataFile("pseudo_release.json"))
	actualBody := string(readClientTestdataFile("actual_release.json"))
	const pseudoID = "44444444-4444-4444-4444-444444444444"
	const actualID = "55555555-5555-5555-5555-555555555555"

	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(req.URL.Path, pseudoID):
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(pseudoBody))}, nil
		case strings.Contains(req.URL.Path, actualID):
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(actualBody))}, nil
		default:
			return nil, errors.New("unexpected request path " + req.URL.Path)
		}
	})
	c := newTestClient(t, rt)

	info, err := c.AlbumForID(context.Background(), pseudoID, matcher.DefaultTunables())
	if err != nil {
		t.Fatalf("AlbumForID: %v", err)
	}
	if info.Country != "GB" {
		t.Errorf("Country = %q, want %q (the actual release's country, merged over the pseudo's)", info.Country, "GB")
	}
	if info.Album != "Pseudo Title" {
		t.Errorf("Album = %q, want %q (the pseudo release's own title, unmerged)", info.Album, "Pseudo Title")
	}
}

// TestReleaseRetriesOnTransientStatusAndMetersEveryAttempt is the
// retry seam, covering the schedule, the per-attempt token, and the
// header on every attempt, without spending a second of wall time.
func TestReleaseRetriesOnTransientStatusAndMetersEveryAttempt(t *testing.T) {
	var attempts int
	var agents []string
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		agents = append(agents, req.Header.Get("User-Agent"))
		if attempts < 3 {
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))}, nil
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(releaseFixture))}, nil
	})

	// One tick per observed attempt, gated on the dispatcher's own
	// cumulative registration count rather than fed ahead of demand:
	// the spare is capped at one, so a tick sent before the next
	// attempt has registered becomes an unclaimed spare instead of that
	// attempt's own token, and the client's single in-flight request
	// means exactly one new registration happens between any two ticks
	// (fold B3). awaitQueued(i) waits for the i-th cumulative
	// registration, escalating the same way
	// TestDispatcherPrefersUserPresentAndStillFeedsBackground's 1-then-9
	// call does, rather than repeating one count that only the first
	// tick would actually need to wait for.
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer func() {
		if err := l.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()
	budget := l.Budget(ClassBatch, 0)
	go func() {
		for i := 1; i <= 3; i++ {
			l.awaitQueued(i)
			ticks <- time.Time{}
		}
	}()

	c, err := NewClient(ClientConfig{
		BaseURL:   "http://stub.invalid",
		UserAgent: "dubplate-test/0 ( test )",
		Budget:    budget,
		HTTP:      &http.Client{Transport: rt},
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	var slept []time.Duration
	c.sleep = func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil }

	if _, err := c.Release(context.Background(), releaseFixtureID); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if got := budget.Spent(); got != 3 {
		t.Errorf("budget.Spent() = %d, want 3: every retry is a real request against the wire budget", got)
	}
	if want := []time.Duration{0, time.Second}; !slices.Equal(slept, want) {
		t.Errorf("backoff schedule = %v, want %v", slept, want)
	}
	for i, ua := range agents {
		if ua != "dubplate-test/0 ( test )" {
			t.Errorf("attempt %d User-Agent = %q, want the configured one", i, ua)
		}
	}
}
