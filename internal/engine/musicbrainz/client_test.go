package musicbrainz

import (
	"context"
	"errors"
	"io"
	"net/http"
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

// stubSleep swaps the package's backoff wait for fn until t ends. No
// test in this package runs in parallel, which is what makes a
// package-level seam safe.
func stubSleep(t *testing.T, fn func(context.Context, time.Duration) error) {
	t.Helper()
	orig := sleep
	sleep = fn
	t.Cleanup(func() { sleep = orig })
}

// newTestClient builds a Client over a scripted transport whose
// backoff never waits.
func newTestClient(t *testing.T, rt http.RoundTripper) *Client {
	t.Helper()
	c, err := NewClient(ClientConfig{
		BaseURL:   "http://stub.invalid",
		UserAgent: "dubplate-test/0 ( test )",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.http.Transport = rt
	stubSleep(t, noSleep)
	return c
}

// Mutation 1: NewClient accepts an empty UserAgent.
func TestNewClientRequiresUserAgent(t *testing.T) {
	if _, err := NewClient(ClientConfig{}); err == nil {
		t.Fatal("NewClient with empty UserAgent = nil error, want one")
	}
}

func TestReleaseDecodesSuccessfulResponseAtTheCommittedURLShape(t *testing.T) {
	var gotURL string
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		gotURL = req.URL.String()
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(releaseFixture))}, nil
	})
	c := newTestClient(t, rt)

	rel, err := c.release(context.Background(), releaseFixtureID)
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if rel.ID != releaseFixtureID {
		t.Errorf("rel.ID = %q, want %q", rel.ID, releaseFixtureID)
	}

	// The include set is spelled out rather than read from
	// releaseIncludes, so dropping an include fails here.
	const wantURL = "http://stub.invalid/ws/2/release/" + releaseFixtureID +
		"?fmt=json&inc=artists%2Bmedia%2Brecordings%2Brelease-groups%2Blabels%2B" +
		"artist-credits%2Baliases%2Brecording-level-rels%2Bwork-rels%2Bwork-level-rels%2B" +
		"artist-rels%2Bisrcs%2Burl-rels%2Brelease-rels%2Bgenres%2Btags"
	if gotURL != wantURL {
		t.Errorf("request URL = %q, want %q", gotURL, wantURL)
	}
}

func TestReleaseDecodeErrorWrapsMalformedRelease(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{not valid json`))}, nil
	})
	c := newTestClient(t, rt)

	_, err := c.release(context.Background(), releaseFixtureID)
	if !errors.Is(err, ErrMalformedRelease) {
		t.Fatalf("release() = %v, want an error wrapping ErrMalformedRelease", err)
	}
}

// TestSearchReleasesReportsAnUndecodableBody proves a search body that
// does not decode wraps ErrMalformedRelease, the sentinel Release and
// SearchReleases share.
func TestSearchReleasesReportsAnUndecodableBody(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{not valid json`))}, nil
	})
	c := newTestClient(t, rt)

	ids, err := c.SearchReleases(context.Background(), AlbumSearch{Query: "release:(x)", Limit: 5})
	if err == nil {
		t.Fatalf("SearchReleases() = %v, nil, want an error", ids)
	}
	if !errors.Is(err, ErrMalformedRelease) {
		t.Errorf("SearchReleases() = %v, want an error wrapping ErrMalformedRelease", err)
	}
}

// Mutation 4: exhausted retries return ErrNotFound instead of
// ErrUnavailable.
func TestReleaseRetryExhaustionReturnsUnavailable(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	c := newTestClient(t, rt)

	_, err := c.release(context.Background(), releaseFixtureID)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("release() = %v, want an error wrapping ErrUnavailable", err)
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
	stubSleep(t, func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil })

	if _, err := c.release(context.Background(), releaseFixtureID); err != nil {
		t.Fatalf("release: %v", err)
	}
	if want := []time.Duration{3 * time.Second}; !slices.Equal(slept, want) {
		t.Errorf("slept = %v, want %v (the Retry-After header, not the computed backoff)", slept, want)
	}
}

// TestRetryAfterHTTPDateFallsThroughToComputedBackoff fails when an
// HTTP-date Retry-After is parsed as seconds, or parsed at all, rather
// than falling through to the computed backoff.
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
	stubSleep(t, func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil })

	if _, err := c.release(context.Background(), releaseFixtureID); err != nil {
		t.Fatalf("release: %v", err)
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

	if _, err := c.release(context.Background(), "not-a-release-id"); err == nil {
		t.Fatal("release(junk id) = nil error, want one")
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

// TestReleaseRetriesOnTransientStatus covers the retry schedule and the
// User-Agent header on every attempt, without spending wall time.
func TestReleaseRetriesOnTransientStatus(t *testing.T) {
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
	c := newTestClient(t, rt)
	var slept []time.Duration
	stubSleep(t, func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil })

	if _, err := c.release(context.Background(), releaseFixtureID); err != nil {
		t.Fatalf("release: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
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

func TestReleaseStopsWhenTheContextEndsDuringABackoffWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var requests int
	rt := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	c := newTestClient(t, rt)
	stubSleep(t, func(ctx context.Context, _ time.Duration) error {
		cancel()
		return ctx.Err()
	})

	_, err := c.release(ctx, releaseFixtureID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("release() = %v, want an error wrapping context.Canceled", err)
	}
	if requests != 1 {
		t.Errorf("requests = %d, want 1: a cancelled wait must not be followed by another request", requests)
	}
}

func TestReleaseStopsWhenTheContextEndsDuringARequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var requests int
	rt := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		requests++
		cancel()
		return nil, ctx.Err()
	})
	c := newTestClient(t, rt)

	_, err := c.release(ctx, releaseFixtureID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("release() = %v, want an error wrapping context.Canceled", err)
	}
	if errors.Is(err, ErrUnavailable) {
		t.Errorf("release() = %v, want the context's error and never ErrUnavailable", err)
	}
	if requests != 1 {
		t.Errorf("requests = %d, want 1: a cancelled context must not be retried", requests)
	}
}

// zeroBody streams remaining zero bytes and counts how many the client
// consumed, so a body far over the cap costs no allocation of its own.
type zeroBody struct {
	remaining int
	consumed  int
}

func (z *zeroBody) Read(p []byte) (int, error) {
	if z.remaining == 0 {
		return 0, io.EOF
	}
	n := min(len(p), z.remaining)
	clear(p[:n])
	z.remaining -= n
	z.consumed += n
	return n, nil
}

// TestReleaseBoundsTheResponseBody pins the 32 MiB cap with literal sizes
// and proves the client stops reading one byte past it.
func TestReleaseBoundsTheResponseBody(t *testing.T) {
	const capBytes = 32 << 20
	tests := []struct {
		name    string
		size    int
		wantErr error
	}{
		{"exactly the cap", capBytes, ErrMalformedRelease},
		{"one byte over the cap", capBytes + 1, errBodyTooLarge},
		{"far over the cap", 2 * capBytes, errBodyTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests int
			body := &zeroBody{remaining: tt.size}
			rt := roundTripperFunc(func(*http.Request) (*http.Response, error) {
				requests++
				return &http.Response{StatusCode: 200, Body: io.NopCloser(body)}, nil
			})
			c := newTestClient(t, rt)

			_, err := c.release(context.Background(), releaseFixtureID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("release() = %v, want an error wrapping %v", err, tt.wantErr)
			}
			if body.consumed > capBytes+1 {
				t.Errorf("consumed %d bytes, want at most %d", body.consumed, capBytes+1)
			}
			if requests != 1 {
				t.Errorf("requests = %d, want 1", requests)
			}
		})
	}
}
