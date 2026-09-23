package serve

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// originRequest builds a request to target on host example.com, the
// Host httptest.NewRequest sets, carrying each header in headers.
func originRequest(method, target string, headers map[string]string) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	for name, value := range headers {
		r.Header.Set(name, value)
	}
	return r
}

// serveOriginCheck runs r through check wrapped around a handler that
// answers 200, and returns the status the chain wrote.
func serveOriginCheck(check *originCheck, r *http.Request) int {
	rec := httptest.NewRecorder()
	check.wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, r)
	return rec.Code
}

func TestOriginCheck(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		headers map[string]string
		want    int
	}{
		{
			// The stdlib alone allows this request; the tightening is
			// the one respect in which the check differs from it.
			name:   "POST carrying neither header is denied",
			method: http.MethodPost,
			want:   http.StatusForbidden,
		},
		{
			name:    "POST carrying a same-origin Origin is allowed",
			method:  http.MethodPost,
			headers: map[string]string{"Origin": "https://example.com"},
			want:    http.StatusOK,
		},
		{
			// Sec-Fetch-Site is honored on its own, with no Origin to
			// corroborate it.
			name:    "POST carrying Sec-Fetch-Site same-origin and no Origin is allowed",
			method:  http.MethodPost,
			headers: map[string]string{"Sec-Fetch-Site": "same-origin"},
			want:    http.StatusOK,
		},
		{
			name:    "POST carrying a cross-origin Origin is denied",
			method:  http.MethodPost,
			headers: map[string]string{"Origin": "https://attacker.example"},
			want:    http.StatusForbidden,
		},
		{
			// The stdlib always allows a safe method, which is why the
			// type's doc comment forbids a state change on one.
			name:   "GET carrying neither header is allowed",
			method: http.MethodGet,
			want:   http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := originRequest(tt.method, "/anything", tt.headers)
			if got := serveOriginCheck(newOriginCheck(), r); got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestOriginCheckBearerBypass exercises a pattern registered through
// addBearerBypass against the same request sent to a path no pattern
// claims. The rule that makes the exemption safe, that a handler behind
// a bypassed pattern never reads the session cookie to authenticate, is
// a contract on that handler rather than a behavior of this check, and
// no test here can reach it: it is stated on originCheck for the
// reviewer of every handler registered behind a pattern.
func TestOriginCheckBearerBypass(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		headers map[string]string
		want    int
	}{
		{
			// The check's own set carries this row: the stdlib allows
			// a request carrying neither header before it consults its
			// own bypass patterns.
			name:   "bypassed path, POST carrying neither header, is allowed",
			target: "/rest/ping",
			want:   http.StatusOK,
		},
		{
			name:   "unregistered path, POST carrying neither header, is denied",
			target: "/other",
			want:   http.StatusForbidden,
		},
		{
			// The stdlib's own set carries this row: a cross-origin
			// request is refused inside Check, which the check's own
			// set is never asked about.
			name:    "bypassed path, POST carrying a cross-origin Origin, is allowed",
			target:  "/rest/ping",
			headers: map[string]string{"Origin": "https://attacker.example"},
			want:    http.StatusOK,
		},
		{
			name:    "unregistered path, POST carrying a cross-origin Origin, is denied",
			target:  "/other",
			headers: map[string]string{"Origin": "https://attacker.example"},
			want:    http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := newOriginCheck()
			check.addBearerBypass("/rest/")

			r := originRequest(http.MethodPost, tt.target, tt.headers)
			if got := serveOriginCheck(check, r); got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
