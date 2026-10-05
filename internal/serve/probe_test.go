package serve

import (
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"slices"
	"testing"
)

// The counts a route that strips nothing must report for the request
// [NewProbeRequest] builds. Each is measured from strip.go's own lists
// with a command rather than taken from them in code:
//
//   - identityExtras holds 22 names, every one carrying a dash, so each
//     has an underscore twin; CanonicalMIMEHeaderKey treats "_" as an
//     ordinary token byte, so Remote_User arrives as Remote_user, a key
//     distinct from Remote-User. 22 names in two spellings are 44
//     distinct canonical keys.
//   - forwardedAddressHeaders holds 10 names.
//   - the sentinel is one header.
const (
	wantUnstrippedIdentity  = 44
	wantUnstrippedForwarded = 10
	wantUnstrippedSentinel  = 1
)

// probeOver sends the shared probe request to srv and decodes the
// counts its route answers, failing the test on anything but a 200
// carrying the JSON object.
func probeOver(t *testing.T, srv *httptest.Server, prefix Prefix) ProbeCounts {
	t.Helper()
	req, err := NewProbeRequest(t.Context(), prefix, srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("NewProbeRequest: %v", err)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("probe status = %d (%q), want 200", resp.StatusCode, body)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	var counts ProbeCounts
	if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
		t.Fatalf("decode probe answer: %v", err)
	}
	return counts
}

// TestProbeRoute_ThroughShellCountsNothing serves the route on a real
// shell over its real strip and probes it from 127.0.0.1 across a
// socket: every count is zero, because the strip removed every header
// the route counts before the mux handed it the request.
func TestProbeRoute_ThroughShellCountsNothing(t *testing.T) {
	prefix, err := NewPrefix("/dubplate")
	if err != nil {
		t.Fatal(err)
	}
	strip := NewStrip("", "dubplate_session", TrustedProxyConfig{}, FrontExemption{})
	shell := NewShell(prefix, strip, slog.New(slog.NewTextHandler(io.Discard, nil)))
	shell.Register(ProbeRouteName, NewProbeRoute([]string{"127.0.0.1"}, TrustedProxyConfig{}))
	srv := httptest.NewServer(shell.Handler())
	defer srv.Close()

	got := probeOver(t, srv, prefix)
	if got != (ProbeCounts{}) {
		t.Errorf("counts through the shell = %+v, want all zero", got)
	}
}

// TestProbeRoute_WithoutStripCountsEverySpelling serves the route
// alone, with no strip, and probes it over the same parser a live dial
// crosses. Every spelling the request carries arrives under its own
// canonical key, and Go's server removes no header the request's
// Connection header lists, so a zero through the shell is the strip's
// work and nothing else's.
func TestProbeRoute_WithoutStripCountsEverySpelling(t *testing.T) {
	prefix, err := NewPrefix("/dubplate")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(NewProbeRoute([]string{"127.0.0.1"}, TrustedProxyConfig{}))
	defer srv.Close()

	got := probeOver(t, srv, prefix)
	want := ProbeCounts{
		Identity:  wantUnstrippedIdentity,
		Forwarded: wantUnstrippedForwarded,
		Sentinel:  wantUnstrippedSentinel,
	}
	if got != want {
		t.Errorf("counts without the strip = %+v, want %+v", got, want)
	}
}

// TestProbeRoute_PeerAndMethod calls the route directly with
// RemoteAddr set. Every refusal must be byte-equal to the terminator's
// answer for an unknown path, so the route's existence is invisible to
// a peer it refuses. The answering rows are the fixture state that
// makes each refusal row meaningful: the same handler does answer a
// permitted GET.
func TestProbeRoute_PeerAndMethod(t *testing.T) {
	proxied := NewTrustedProxyConfig([]netip.Addr{netip.MustParseAddr("192.0.2.10")}, "X-Forwarded-For")
	loopbackProxy := NewTrustedProxyConfig([]netip.Addr{netip.MustParseAddr("127.0.0.1")}, "X-Forwarded-For")

	tests := []struct {
		name       string
		bind       []string
		proxy      TrustedProxyConfig
		method     string
		remoteAddr string
		header     map[string]string
		wantOK     bool
	}{
		{name: "GET from loopback answers", bind: []string{"127.0.0.1"}, method: http.MethodGet, remoteAddr: "127.0.0.1:40000", wantOK: true},
		{name: "GET from another loopback address answers", bind: []string{"127.0.0.1"}, method: http.MethodGet, remoteAddr: "127.0.0.2:40000", wantOK: true},
		{name: "GET from mapped loopback answers", bind: []string{"127.0.0.1"}, method: http.MethodGet, remoteAddr: "[::ffff:127.0.0.1]:40000", wantOK: true},
		{name: "GET from a configured bind address answers", bind: []string{"198.51.100.4"}, method: http.MethodGet, remoteAddr: "198.51.100.4:40000", wantOK: true},
		{name: "GET from a mapped bind entry's peer answers", bind: []string{"::ffff:198.51.100.4"}, method: http.MethodGet, remoteAddr: "198.51.100.4:40000", wantOK: true},
		{name: "GET from a non-local peer", bind: []string{"127.0.0.1"}, method: http.MethodGet, remoteAddr: "198.51.100.7:40000"},
		{name: "POST from a non-local peer", bind: []string{"127.0.0.1"}, method: http.MethodPost, remoteAddr: "198.51.100.7:40000"},
		{name: "POST from loopback", bind: []string{"127.0.0.1"}, method: http.MethodPost, remoteAddr: "127.0.0.1:40000"},
		{name: "HEAD from loopback", bind: []string{"127.0.0.1"}, method: http.MethodHead, remoteAddr: "127.0.0.1:40000"},
		{name: "GET from loopback configured as a trusted proxy", bind: []string{"127.0.0.1"}, proxy: loopbackProxy, method: http.MethodGet, remoteAddr: "127.0.0.1:40000"},
		{
			name:       "GET from a non-local trusted proxy naming loopback",
			bind:       []string{"127.0.0.1"},
			proxy:      proxied,
			method:     http.MethodGet,
			remoteAddr: "192.0.2.10:40000",
			header:     map[string]string{"X-Forwarded-For": "127.0.0.1"},
		},
		{name: "a RemoteAddr that does not parse", bind: []string{"127.0.0.1"}, method: http.MethodGet, remoteAddr: "not-an-address"},
	}

	unknown := httptest.NewRecorder()
	Terminator().ServeHTTP(unknown, httptest.NewRequest(http.MethodGet, "/dubplate/no-such-route", nil))
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("fixture state: terminator answered %d, want 404", unknown.Code)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/dubplate/"+ProbeRouteName, nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.header {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			NewProbeRoute(tt.bind, tt.proxy).ServeHTTP(rec, req)

			if tt.wantOK {
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200", rec.Code)
				}
				if got := rec.Header().Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", got)
				}
				return
			}
			if rec.Code != unknown.Code {
				t.Errorf("status = %d, want the terminator's %d", rec.Code, unknown.Code)
			}
			if got, want := rec.Body.String(), unknown.Body.String(); got != want {
				t.Errorf("body = %q, want the terminator's %q", got, want)
			}
			if got, want := rec.Header(), unknown.Header(); !maps.EqualFunc(got, want, slices.Equal) {
				t.Errorf("header = %v, want the terminator's %v", got, want)
			}
			if allow := rec.Header().Values("Allow"); len(allow) != 0 {
				t.Errorf("Allow = %q, want none: a 405-style hint names the route", allow)
			}
		})
	}
}

// TestNewProbeRequest_RefusesBadInput pins the builder's one failure:
// a host that leaves the URL unparseable.
func TestNewProbeRequest_RefusesBadInput(t *testing.T) {
	prefix, err := NewPrefix("/dubplate")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewProbeRequest(t.Context(), prefix, "127.0.0.1:%zz"); err == nil {
		t.Error("NewProbeRequest with an unparseable host: error = nil, want one")
	}
}
