package serve

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// boundBytes is the path bound rule 5 refuses past, written out here
// rather than read from maxPathBytes. A test that asks the
// implementation for its own bound cannot fail when the bound moves.
const boundBytes = 2048

// pathRecorder is the terminal handler's record of what reached it. A
// refused request must leave called false, which is C3's
// no-partial-routing property stated as an assertion rather than
// inferred from a status code.
type pathRecorder struct {
	called bool
	path   string
}

func recordingPathHandler(dst *pathRecorder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dst.called = true
		dst.path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
}

// readRequest parses raw the way a server does, so an authority-form
// CONNECT target and an absolute-form target arrive in the exact shape
// net/http hands a handler. httptest.NewRequest cannot express either
// one: it rewrites what it is given.
func readRequest(t *testing.T, raw string) *http.Request {
	t.Helper()
	r, err := http.ReadRequest(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		t.Fatalf("read request %q: %v", raw, err)
	}
	return r
}

func originForm(target string) string {
	return "GET " + target + " HTTP/1.1\r\nHost: example.test\r\n\r\n"
}

func TestNormalize(t *testing.T) {
	atBound := "/" + strings.Repeat("a", boundBytes-1)
	pastBound := "/" + strings.Repeat("a", boundBytes)

	tests := []struct {
		name string
		// raw is the request as it arrives on the wire.
		raw string
		// want is the status the client sees, a literal per row.
		want int
		// admitted rows reach the terminal handler; wantPath is the
		// path it must receive, byte for byte.
		admitted bool
		wantPath string
	}{
		{name: "plain path", raw: originForm("/dubplate/x"), want: 200, admitted: true, wantPath: "/dubplate/x"},
		{name: "encoded separator in a segment", raw: originForm("/dubplate/a%2Fb"), want: 400},
		{name: "encoded backslash", raw: originForm("/dubplate/a%5Cb"), want: 400},
		{name: "double-encoded separator", raw: originForm("/dubplate/a%252Fb"), want: 400},
		{name: "benign encoded byte", raw: originForm("/%E2%98%83"), want: 400},
		{name: "dot dot segment", raw: originForm("/dubplate/../x"), want: 400},
		{name: "encoded dot dot segment", raw: originForm("/dubplate/%2E%2E/x"), want: 400},
		{name: "single dot segment", raw: originForm("/dubplate/./x"), want: 400},
		{name: "doubled separator", raw: originForm("/dubplate//x"), want: 400},
		{name: "root", raw: originForm("/"), want: 200, admitted: true, wantPath: "/"},
		{name: "prefix with a trailing separator", raw: originForm("/dubplate/"), want: 200, admitted: true, wantPath: "/dubplate/"},
		{name: "doubled separator at the front", raw: originForm("//dubplate"), want: 400},
		{name: "backslash", raw: originForm("/dubplate\\x"), want: 400},
		{name: "path at the bound", raw: originForm(atBound), want: 200, admitted: true, wantPath: atBound},
		{name: "path one byte past the bound", raw: originForm(pastBound), want: 414},
		{name: "authority-form CONNECT target", raw: "CONNECT example.test:443 HTTP/1.1\r\nHost: example.test:443\r\n\r\n", want: 400},
		{name: "absolute-form target", raw: "GET http://example.test/dubplate/x HTTP/1.1\r\nHost: example.test\r\n\r\n", want: 200, admitted: true, wantPath: "/dubplate/x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got pathRecorder
			rec := httptest.NewRecorder()

			Normalize(recordingPathHandler(&got)).ServeHTTP(rec, readRequest(t, tt.raw))

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
			if got.called != tt.admitted {
				t.Fatalf("terminal handler called = %v, want %v (a refused request must be answered by the wrapper alone)", got.called, tt.admitted)
			}
			if tt.admitted && got.path != tt.wantPath {
				t.Errorf("terminal handler received path %q, want %q (normalization is a gate, never a rewriter)", got.path, tt.wantPath)
			}
		})
	}
}

// TestUnencodedPathCarriesEmptyRawPath pins the net/url fact R-D's
// mechanism rests on and the Normalize doc comment states: RawPath is
// populated only when the escaped form differs from the default
// encoding of Path, so a path needing no escaping carries an empty
// RawPath and a comparison against it would admit every encoded path
// whose decoded form happens to re-encode identically.
func TestUnencodedPathCarriesEmptyRawPath(t *testing.T) {
	tests := []struct {
		target      string
		wantRawPath string
		wantEscaped string
		wantPath    string
	}{
		{target: "/dubplate/x", wantRawPath: "", wantEscaped: "/dubplate/x", wantPath: "/dubplate/x"},
		{target: "/dubplate/a%5Cb", wantRawPath: "", wantEscaped: "/dubplate/a%5Cb", wantPath: "/dubplate/a\\b"},
		{target: "/dubplate/a%2Fb", wantRawPath: "/dubplate/a%2Fb", wantEscaped: "/dubplate/a%2Fb", wantPath: "/dubplate/a/b"},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			u := readRequest(t, originForm(tt.target)).URL

			if u.RawPath != tt.wantRawPath {
				t.Errorf("RawPath = %q, want %q", u.RawPath, tt.wantRawPath)
			}
			if u.EscapedPath() != tt.wantEscaped {
				t.Errorf("EscapedPath() = %q, want %q", u.EscapedPath(), tt.wantEscaped)
			}
			if u.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", u.Path, tt.wantPath)
			}
		})
	}
}

// TestServeMuxCleansBeforeARouteRuns is risk 1 stated as behavior: a
// normalization check registered as a route never sees a dot segment,
// because http.ServeMux redirects the request before any route runs.
// This is the measurement behind Normalize wrapping the mux, and it
// fails if a later change registers a check as a route instead.
func TestServeMuxCleansBeforeARouteRuns(t *testing.T) {
	tests := []struct {
		target       string
		wantStatus   int
		wantLocation string
	}{
		{target: "/dubplate/../x", wantStatus: 307, wantLocation: "/x"},
		{target: "/dubplate/./x", wantStatus: 307, wantLocation: "/dubplate/x"},
		{target: "/dubplate//x", wantStatus: 307, wantLocation: "/dubplate/x"},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			var got pathRecorder
			mux := http.NewServeMux()
			mux.Handle("/dubplate/", Normalize(recordingPathHandler(&got)))

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, readRequest(t, originForm(tt.target)))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if location := rec.Header().Get("Location"); location != tt.wantLocation {
				t.Errorf("Location = %q, want %q", location, tt.wantLocation)
			}
			if got.called {
				t.Errorf("terminal handler ran on %q, which the mux redirect should have prevented", tt.target)
			}
		})
	}
}
