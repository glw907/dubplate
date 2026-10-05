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

// probeOver sends the shared probe request, built for proxy, to srv
// and decodes the counts its route answers, failing the test on
// anything but a 200 carrying the JSON object.
func probeOver(t *testing.T, srv *httptest.Server, prefix Prefix, proxy TrustedProxyConfig) ProbeCounts {
	t.Helper()
	var counts ProbeCounts
	if err := json.Unmarshal(probeBody(t, srv, prefix, proxy), &counts); err != nil {
		t.Fatalf("decode probe answer: %v", err)
	}
	return counts
}

// probeBody sends the shared probe request, built for proxy, to srv and
// returns the body of the route's 200, failing the test on any other
// answer.
func probeBody(t *testing.T, srv *httptest.Server, prefix Prefix, proxy TrustedProxyConfig) []byte {
	t.Helper()
	req, err := NewProbeRequest(t.Context(), prefix, srv.Listener.Addr().String(), proxy)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read probe answer: %v", err)
	}
	return body
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

	got := probeOver(t, srv, prefix, TrustedProxyConfig{})
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

	got := probeOver(t, srv, prefix, TrustedProxyConfig{})
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
		{name: "GET from loopback configured as a trusted proxy answers", bind: []string{"127.0.0.1"}, proxy: loopbackProxy, method: http.MethodGet, remoteAddr: "127.0.0.1:40000", wantOK: true},
		{name: "POST from loopback configured as a trusted proxy", bind: []string{"127.0.0.1"}, proxy: loopbackProxy, method: http.MethodPost, remoteAddr: "127.0.0.1:40000"},
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
	if _, err := NewProbeRequest(t.Context(), prefix, "127.0.0.1:%zz", TrustedProxyConfig{}); err == nil {
		t.Error("NewProbeRequest with an unparseable host: error = nil, want one")
	}
}

// TestCountPresentMatchesNormalizedKeys proves the count reads a key by
// the same normalized match the strip removes it by, so a spelling the
// strip handles is a spelling the probe can see. Each key the header
// carries is counted once; an unrelated key is not.
func TestCountPresentMatchesNormalizedKeys(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		want   int
	}{
		{"the dash form", http.Header{"Remote-User": {"x"}}, 1},
		{"the underscore form", http.Header{"Remote_user": {"x"}}, 1},
		{"a mixed-separator form", http.Header{"X-Forwarded_user": {"x"}}, 1},
		{"a key the parser never canonicalized", http.Header{"x-forwarded-user": {"x"}}, 1},
		{"two spellings of one name", http.Header{"Remote-User": {"x"}, "Remote_user": {"x"}}, 2},
		{"an unrelated key", http.Header{"X-App-Debug-Trace": {"x"}}, 0},
		{"a key with no value", http.Header{"Remote-User": {}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countPresent(tt.header, probeIdentityNames()); got != tt.want {
				t.Errorf("countPresent(%v) = %d, want %d", tt.header, got, tt.want)
			}
		})
	}
}

// TestProbeConfiguredHeader proves the probe carries, and counts, the
// trusted-proxy header an adopter configured outside the fixed family:
// the one header [ClientIP] reads. Without the strip the route counts
// it beside the fixed family's ten, so the zero through the shell is
// the strip removing it, not a probe that never sent it. The peer,
// 127.0.0.1, is not the configured proxy, so the strip treats it as
// any client.
func TestProbeConfiguredHeader(t *testing.T) {
	prefix, err := NewPrefix("/dubplate")
	if err != nil {
		t.Fatal(err)
	}
	proxy := NewTrustedProxyConfig(addrs("192.0.2.10"), "X-House-Client-Address")

	req, err := NewProbeRequest(t.Context(), prefix, "127.0.0.1:1", proxy)
	if err != nil {
		t.Fatalf("NewProbeRequest: %v", err)
	}
	if got := req.Header.Get("X-House-Client-Address"); got != "probe" {
		t.Errorf("probe request X-House-Client-Address = %q, want %q", got, "probe")
	}

	t.Run("without the strip the route counts it", func(t *testing.T) {
		srv := httptest.NewServer(NewProbeRoute([]string{"127.0.0.1"}, proxy))
		defer srv.Close()
		got := probeOver(t, srv, prefix, proxy)
		if want := wantUnstrippedForwarded + 1; got.Forwarded != want {
			t.Errorf("forwarded count = %d, want %d", got.Forwarded, want)
		}
	})

	t.Run("through the shell the strip removes it", func(t *testing.T) {
		strip := NewStrip("", "dubplate_session", proxy, FrontExemption{})
		shell := NewShell(prefix, strip, slog.New(slog.NewTextHandler(io.Discard, nil)))
		shell.Register(ProbeRouteName, NewProbeRoute([]string{"127.0.0.1"}, proxy))
		srv := httptest.NewServer(shell.Handler())
		defer srv.Close()
		if got := probeOver(t, srv, prefix, proxy); got != (ProbeCounts{}) {
			t.Errorf("counts through the shell = %+v, want all zero", got)
		}
	})
}

// TestProbeRoute_TrustedPeerReceivesIdentityOnly probes from 127.0.0.1
// configured as the trusted proxy, the same-host front's topology. The
// strip keeps group 2 for that peer by design, so the route answers
// the identity count alone, never a forwarded or sentinel count a
// reader could take for a strip failure. Without the strip the
// identity count is every spelling, which is what makes the zero
// through the shell the strip's work.
func TestProbeRoute_TrustedPeerReceivesIdentityOnly(t *testing.T) {
	prefix, err := NewPrefix("/dubplate")
	if err != nil {
		t.Fatal(err)
	}
	proxy := NewTrustedProxyConfig(addrs("127.0.0.1"), "X-Forwarded-For")

	tests := []struct {
		name         string
		strip        bool
		wantIdentity int
	}{
		{"without the strip", false, wantUnstrippedIdentity},
		{"through the shell", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewProbeRoute([]string{"127.0.0.1"}, proxy)
			if tt.strip {
				shell := NewShell(prefix, NewStrip("", "dubplate_session", proxy, FrontExemption{}), slog.New(slog.NewTextHandler(io.Discard, nil)))
				shell.Register(ProbeRouteName, h)
				h = shell.Handler()
			}
			srv := httptest.NewServer(h)
			defer srv.Close()

			var answer map[string]int
			if err := json.Unmarshal(probeBody(t, srv, prefix, proxy), &answer); err != nil {
				t.Fatalf("decode probe answer: %v", err)
			}
			if want := map[string]int{"identity": tt.wantIdentity}; !maps.Equal(answer, want) {
				t.Errorf("answer = %v, want %v", answer, want)
			}
		})
	}
}
