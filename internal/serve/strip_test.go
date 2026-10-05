package serve

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Documentation-reserved fixture addresses (RFC 5737): the TEST-NET-1
// and TEST-NET-2 blocks. Never an estate address.

// recorder captures the header and trailer the strip's wrapped handler
// hands it, so a test can assert on what the handler actually
// received rather than on the strip's own construction.
type recorder struct {
	header  http.Header
	trailer http.Header
}

func recordingHandler(dst *recorder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dst.header = r.Header.Clone()
		io.Copy(io.Discard, r.Body) //nolint:errcheck // test helper, draining is the point
		dst.trailer = r.Trailer.Clone()
		w.WriteHeader(http.StatusOK)
	})
}

func TestUnderscoreAndDashAreDistinctCanonicalKeys(t *testing.T) {
	// spec:779-787: the plan asserts textproto.CanonicalMIMEHeaderKey's
	// underscore behavior directly rather than from memory, since it
	// is what makes group 1's underscore spellings a real attack
	// surface rather than a no-op.
	h := http.Header{}
	h.Set("Remote-User", "dash-value")
	h.Set("Remote_User", "underscore-value")

	if got := len(h); got != 2 {
		t.Fatalf("Remote-User and Remote_User collapsed into %d key(s), want 2: %v", got, h)
	}

	h.Del("Remote-User")

	if h.Get("Remote-User") != "" {
		t.Errorf("Remote-User still present after Del(\"Remote-User\")")
	}
	if got := h.Get("Remote_User"); got != "underscore-value" {
		t.Errorf("Remote_User = %q after deleting Remote-User, want %q (deleting one spelling must leave the other standing)", got, "underscore-value")
	}
}

// expectedIdentitySpellings is written out independently of
// identityExtras, so a strip that lost a name still fails this test.
var expectedIdentitySpellings = []string{
	"Remote-User", "Remote_User",
	"Remote-Name", "Remote_Name",
	"Remote-Email", "Remote_Email",
	"Remote-Groups", "Remote_Groups",
	"X-Forwarded-User", "X_Forwarded_User",
	"X-Forwarded-Email", "X_Forwarded_Email",
	"X-Forwarded-Preferred-Username", "X_Forwarded_Preferred_Username",
	"X-Forwarded-Groups", "X_Forwarded_Groups",
	"X-Forwarded-Access-Token", "X_Forwarded_Access_Token",
	"X-Forwarded-Uri", "X_Forwarded_Uri",
	"X-Forwarded-Method", "X_Forwarded_Method",
	"Cf-Access-Jwt-Assertion", "Cf_Access_Jwt_Assertion",
	"Cf-Access-Authenticated-User-Email", "Cf_Access_Authenticated_User_Email",
	"X-WEBAUTH-USER", "X_WEBAUTH_USER",
	"X-authentik-username", "X_authentik_username",
	"X-authentik-email", "X_authentik_email",
	"X-authentik-groups", "X_authentik_groups",
	"X-authentik-uid", "X_authentik_uid",
	"X-authentik-jwt", "X_authentik_jwt",
	"Tailscale-User-Login", "Tailscale_User_Login",
	"Tailscale-User-Name", "Tailscale_User_Name",
	"X-Original-URL", "X_Original_URL",
	// The driver's configured header name and its underscore twin,
	// proving the driver-contribution mechanism rung 12 will use.
	"X-Driver-Identity", "X_Driver_Identity",
}

func TestStripWrapRemovesEveryIdentitySpelling(t *testing.T) {
	strip := NewStrip("X-Driver-Identity", "", TrustedProxyConfig{}, FrontExemption{})

	r := httptest.NewRequest("GET", "/", nil)
	for _, name := range expectedIdentitySpellings {
		r.Header.Set(name, "attacker")
	}

	var got recorder
	strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

	for _, name := range expectedIdentitySpellings {
		if v := got.header.Get(name); v != "" {
			t.Errorf("handler received %s = %q, want removed", name, v)
		}
	}
}

func TestStripWrapClearsAnnouncedTrailerOnceBodyIsRead(t *testing.T) {
	// R-P: Go populates a server request's Trailer map only once the
	// handler reads the body through to io.EOF, so the strip must
	// clear it on that same pass, not only on r.Header up front.
	var beforeHeader, afterHeader, afterTrailer string
	var beforeTrailerPresent bool
	srv := httptest.NewServer(NewStrip("", "", TrustedProxyConfig{}, FrontExemption{}).Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			beforeHeader = r.Header.Get("Remote-User")
			_, beforeTrailerPresent = r.Trailer["Remote-User"]
			io.Copy(io.Discard, r.Body) //nolint:errcheck // test drains to trigger trailer population
			afterHeader = r.Header.Get("Remote-User")
			afterTrailer = r.Trailer.Get("Remote-User")
			w.WriteHeader(http.StatusOK)
		}),
	))
	defer srv.Close()

	pr, pw := io.Pipe()
	req, err := http.NewRequest(http.MethodPost, srv.URL, pr)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Trailer", "Remote-User")
	req.Trailer = http.Header{"Remote-User": nil}

	go func() {
		_, _ = pw.Write([]byte("body"))
		req.Trailer.Set("Remote-User", "attacker")
		_ = pw.Close()
	}()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_ = resp.Body.Close()

	if beforeHeader != "" {
		t.Errorf("before body read: r.Header[Remote-User] = %q, want empty", beforeHeader)
	}
	if beforeTrailerPresent {
		t.Error("before body read: r.Trailer still has the Remote-User key announced by the request, want the strip to have deleted it up front")
	}
	if afterHeader != "" {
		t.Errorf("after body read: r.Header[Remote-User] = %q, want empty", afterHeader)
	}
	if afterTrailer != "" {
		t.Errorf("after body read: r.Trailer[Remote-User] = %q, want removed, the smuggled value should never reach the handler", afterTrailer)
	}
}

func TestStripWrapConnectionListedRemoval(t *testing.T) {
	tests := []struct {
		name       string
		connection []string // one entry per repeated Connection header
		set        map[string]string
		wantAbsent []string
		wantValue  map[string]string
	}{
		{
			// The induced-downgrade attack by name (spec:796-806):
			// a client-controlled Connection value that names a real
			// identity header. Group 1 also covers Remote-User
			// unconditionally, so this row is the literal real-world
			// case rather than the isolating one below.
			name:       "induced downgrade by name: Connection lists Remote-User",
			connection: []string{"Remote-User"},
			set:        map[string]string{"Remote-User": "attacker"},
			wantAbsent: []string{"Remote-User"},
		},
		{
			// A header outside both group 1 and group 2, so only
			// group 3's mechanism can be responsible for its removal;
			// this isolates the Connection-listed rule from the
			// identity strip's own coverage of Remote-User above.
			name:       "generic header named by Connection, outside every other group",
			connection: []string{"X-App-Debug-Trace"},
			set:        map[string]string{"X-App-Debug-Trace": "trace-id"},
			wantAbsent: []string{"X-App-Debug-Trace"},
		},
		{
			name:       "multi-valued Connection header",
			connection: []string{"X-App-Debug-Trace", "X-App-Session-Hint"},
			set: map[string]string{
				"X-App-Debug-Trace":  "trace-id",
				"X-App-Session-Hint": "hint",
			},
			wantAbsent: []string{"X-App-Debug-Trace", "X-App-Session-Hint"},
		},
		{
			// R-O carve-out 1: a websocket handshake names Upgrade in
			// Connection, and both Upgrade and Connection survive
			// alongside the identity header still being removed.
			name:       "Connection: Upgrade, Remote-User keeps the handshake",
			connection: []string{"Upgrade, Remote-User"},
			set: map[string]string{
				"Remote-User": "attacker",
				"Upgrade":     "websocket",
			},
			wantAbsent: []string{"Remote-User"},
			wantValue: map[string]string{
				"Upgrade":    "websocket",
				"Connection": "Upgrade, Remote-User",
			},
		},
		{
			// R-S carve-out 2: Origin is never removed by the listed
			// rule, whatever Connection says.
			name:       "Connection: Origin leaves Origin standing",
			connection: []string{"Origin"},
			set:        map[string]string{"Origin": "https://example.invalid"},
			wantValue:  map[string]string{"Origin": "https://example.invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strip := NewStrip("", "", TrustedProxyConfig{}, FrontExemption{})

			r := httptest.NewRequest("GET", "/", nil)
			for _, c := range tt.connection {
				r.Header.Add("Connection", c)
			}
			for name, value := range tt.set {
				r.Header.Set(name, value)
			}

			var got recorder
			strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

			for _, name := range tt.wantAbsent {
				if v := got.header.Get(name); v != "" {
					t.Errorf("%s = %q, want removed", name, v)
				}
			}
			for name, want := range tt.wantValue {
				if got := got.header.Get(name); got != want {
					t.Errorf("%s = %q, want %q", name, got, want)
				}
			}
		})
	}
}

func TestStripWrapGroup2ForwardedAddress(t *testing.T) {
	// The whole enumerated family (R-R): X-Forwarded-For, X-Real-IP,
	// Forwarded, and the X-Forwarded-Host/Proto/Prefix trio, written
	// out rather than read from forwardedAddressHeaders. X-Forwarded-For
	// carries the address of the trusted proxy the tests below
	// configure (203.0.113.10), the value an attacker would spoof to
	// impersonate the trusted peer; the group 2 trust decision must
	// come from PeerAddress(r), Task 1's function, never from this
	// header, so this value alone proves nothing about who dialed the
	// socket.
	family := map[string]string{
		"X-Forwarded-For":    "203.0.113.10",
		"X-Real-IP":          "198.51.100.9",
		"Forwarded":          "for=198.51.100.9",
		"X-Forwarded-Host":   "attacker.invalid",
		"X-Forwarded-Proto":  "https",
		"X-Forwarded-Prefix": "/evil",
	}

	tests := []struct {
		name       string
		proxy      TrustedProxyConfig
		remoteAddr string
		wantAbsent bool
	}{
		{
			name:       "mode off: family removed regardless of peer",
			proxy:      TrustedProxyConfig{},
			remoteAddr: "203.0.113.10:1111",
			wantAbsent: true,
		},
		{
			name:       "mode on, matching peer: family survives",
			proxy:      NewTrustedProxyConfig(addrs("203.0.113.10"), "X-Forwarded-For"),
			remoteAddr: "203.0.113.10:1111",
			wantAbsent: false,
		},
		{
			name:       "mode on, non-matching peer: family removed",
			proxy:      NewTrustedProxyConfig(addrs("203.0.113.10"), "X-Forwarded-For"),
			remoteAddr: "192.0.2.5:1111",
			wantAbsent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strip := NewStrip("", "", tt.proxy, FrontExemption{})

			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			for name, value := range family {
				r.Header.Set(name, value)
			}

			var got recorder
			strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

			for name := range family {
				v := got.header.Get(name)
				if tt.wantAbsent && v != "" {
					t.Errorf("%s = %q, want removed", name, v)
				}
				if !tt.wantAbsent && v == "" {
					t.Errorf("%s missing, want it to survive a trusted, matching peer", name)
				}
			}
		})
	}
}

func TestStripWrapExemption(t *testing.T) {
	exemption := NewFrontExemption("X-Front-Identity", []string{"203.0.113.10"})
	strip := NewStrip("X-Front-Identity", "", TrustedProxyConfig{}, exemption)

	t.Run("matching peer: the named header survives", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.10:1111"
		r.Header.Set("X-Front-Identity", "front-asserted-user")

		var got recorder
		strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

		if v := got.header.Get("X-Front-Identity"); v != "front-asserted-user" {
			t.Errorf("X-Front-Identity = %q, want %q to survive the exemption", v, "front-asserted-user")
		}
	})

	t.Run("matching peer: the underscore twin does not survive", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.10:1111"
		r.Header.Set("X_Front_Identity", "smuggled")

		var got recorder
		strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

		if v := got.header.Get("X_Front_Identity"); v != "" {
			t.Errorf("X_Front_Identity = %q, want removed: only the exact configured spelling is exempt", v)
		}
	})

	t.Run("non-matching peer: the named header is stripped like any other", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "192.0.2.5:1111"
		r.Header.Set("X-Front-Identity", "front-asserted-user")

		var got recorder
		strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

		if v := got.header.Get("X-Front-Identity"); v != "" {
			t.Errorf("X-Front-Identity = %q, want removed: the peer does not match the front's configuration", v)
		}
	})

	t.Run("duplicate exempted name collapses to its last value", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.10:1111"
		r.Header.Add("X-Front-Identity", "first")
		r.Header.Add("X-Front-Identity", "second")

		var got recorder
		strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

		values := got.header.Values("X-Front-Identity")
		if len(values) != 1 || values[0] != "second" {
			t.Errorf("X-Front-Identity values = %v, want exactly [\"second\"]", values)
		}
	})
}

func TestStripWrapEmptyExemptionExemptsNothing(t *testing.T) {
	// A shell constructed with no front exempts nothing (spec:849-862).
	strip := NewStrip("X-Driver-Identity", "", TrustedProxyConfig{}, FrontExemption{})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.10:1111"
	r.Header.Set("X-Driver-Identity", "attacker")

	var got recorder
	strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

	if v := got.header.Get("X-Driver-Identity"); v != "" {
		t.Errorf("X-Driver-Identity = %q, want removed: no front is configured, so nothing is exempt", v)
	}
}

// TestStripWrapClearsUnannouncedTrailerOnceBodyIsRead is the
// announcement-free half of R-P: a chunked request may send a trailer
// section it never named in a Trailer header, and net/http merges that
// section into r.Trailer at end of stream all the same. Arming the
// read-time scrub on the announcement alone would leave this request's
// identity header standing, which is the whole of group 1 bypassed by
// omitting one header.
//
// The request is written onto the socket rather than built with
// http.NewRequest, because net/http's own client sends only the
// trailers a request announces; the server side is still net/http's
// own chunked parsing.
func TestStripWrapClearsUnannouncedTrailerOnceBodyIsRead(t *testing.T) {
	var afterHeader, afterTrailer string
	var trailerKeyPresent bool
	srv := httptest.NewServer(NewStrip("", "", TrustedProxyConfig{}, FrontExemption{}).Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.Copy(io.Discard, r.Body) //nolint:errcheck // test drains to trigger trailer population
			afterHeader = r.Header.Get("Remote-User")
			afterTrailer = r.Trailer.Get("Remote-User")
			_, trailerKeyPresent = r.Trailer["Remote-User"]
			w.WriteHeader(http.StatusOK)
		}),
	))
	defer srv.Close()

	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// No Trailer header names Remote-User: the trailer section arrives
	// unannounced, after the terminating zero-length chunk.
	raw := "POST / HTTP/1.1\r\n" +
		"Host: example.test\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\n" +
		"4\r\nbody\r\n" +
		"0\r\n" +
		"Remote-User: attacker\r\n" +
		"\r\n"
	if _, err := conn.Write([]byte(raw)); err != nil {
		t.Fatalf("write request: %v", err)
	}

	// Reading the response is what guarantees the handler has run.
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("ReadResponse: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if afterHeader != "" {
		t.Errorf("r.Header[Remote-User] = %q, want empty", afterHeader)
	}
	if afterTrailer != "" {
		t.Errorf("r.Trailer[Remote-User] = %q, want removed: an unannounced trailer must not carry an identity header past the strip", afterTrailer)
	}
	if trailerKeyPresent {
		t.Error("r.Trailer still carries the Remote-User key, want it deleted once the trailer section arrived")
	}
}

// TestStripWrapExemptedHeaderSurvivesConnectionListing: the one header
// a configured front is exempt from group 1 on survives group 3 too.
// Without the carve-out an attacker could name that header in its own
// Connection header and delete the front's identity assertion, turning
// a request a trusted peer authenticated into one the shell reads as
// anonymous.
func TestStripWrapExemptedHeaderSurvivesConnectionListing(t *testing.T) {
	exemption := NewFrontExemption("X-Front-Identity", []string{"203.0.113.10"})
	strip := NewStrip("X-Front-Identity", "", TrustedProxyConfig{}, exemption)

	t.Run("matching peer: Connection cannot delete the exempted header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.10:1111"
		r.Header.Set("X-Front-Identity", "front-asserted-user")
		r.Header.Set("Connection", "X-Front-Identity")

		var got recorder
		strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

		if v := got.header.Get("X-Front-Identity"); v != "front-asserted-user" {
			t.Errorf("X-Front-Identity = %q, want %q: the exemption must outrank a client-supplied Connection listing", v, "front-asserted-user")
		}
	})

	t.Run("non-matching peer: nothing is exempt from either group", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "192.0.2.5:1111"
		r.Header.Set("X-Front-Identity", "attacker")
		r.Header.Set("Connection", "X-Front-Identity")

		var got recorder
		strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

		if v := got.header.Get("X-Front-Identity"); v != "" {
			t.Errorf("X-Front-Identity = %q, want removed: the peer does not match the front's configuration", v)
		}
	})
}

// clientIPFamily is the client-address half of group 2, written out
// independently of forwardedAddressHeaders so a strip that lost one
// still fails. Every name is one a real front sets, which is what makes
// a client's own spelling of it a forgery a handler would otherwise
// read as the front's word.
var clientIPFamily = []string{
	"Cf-Connecting-IP",
	"True-Client-IP",
	"Fastly-Client-IP",
	"X-Azure-ClientIP",
}

func TestStripWrapRemovesTheClientIPFamily(t *testing.T) {
	strip := NewStrip("", "", TrustedProxyConfig{}, FrontExemption{})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.5:1111"
	for _, name := range clientIPFamily {
		r.Header.Set(name, "198.51.100.9")
	}

	var got recorder
	strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

	for _, name := range clientIPFamily {
		if v := got.header.Get(name); v != "" {
			t.Errorf("handler received %s = %q, want removed", name, v)
		}
	}
}

// TestStripWrapRemovesTheConfiguredProxyHeader is the group 2 gap the
// fixed family alone leaves open. An adopter who configures a
// trusted-proxy header outside the X-Forwarded family is naming the one
// header [ClientIP] reads; if group 2 removes only the fixed family,
// that header arrives from any peer untouched, and a client dialing the
// shell directly sets its own client address. The row set covers a name
// inside the family too, so the derivation cannot pass by appending
// blindly.
func TestStripWrapRemovesTheConfiguredProxyHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"a Cloudflare-shaped name", "Cf-Connecting-IP"},
		{"an Akamai-shaped name", "True-Client-IP"},
		{"an adopter's own name", "X-House-Client-Address"},
		{"a name already in the fixed family", "X-Forwarded-For"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := NewTrustedProxyConfig(addrs("203.0.113.10"), tt.header)
			strip := NewStrip("", "", proxy, FrontExemption{})

			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = "192.0.2.5:1111" // not a configured proxy
			r.Header.Set(tt.header, "198.51.100.9")

			var got recorder
			strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

			if v := got.header.Get(tt.header); v != "" {
				t.Errorf("handler received the configured header %s = %q from an untrusted peer, want removed", tt.header, v)
			}
			if ip := ClientIP(r, proxy); ip != "192.0.2.5" {
				t.Errorf("ClientIP = %q, want the peer address 192.0.2.5", ip)
			}
		})
	}
}

// TestStripWrapTrustedPeerKeepsForwardedListedInConnection closes the
// interaction between groups 2 and 3. Group 2 preserves the forwarded
// address headers for a trusted front, and group 3 then deleted
// whatever the request's own Connection header listed, so a client
// behind the front sending "Connection: X-Forwarded-For" erased the
// header the front had set and collapsed ClientIP to the front's own
// address. The assertion is on ClientIP rather than the header alone:
// the header surviving is the mechanism, and the client's address
// reaching the handler is the property.
func TestStripWrapTrustedPeerKeepsForwardedListedInConnection(t *testing.T) {
	const (
		proxyAddr  = "203.0.113.10"
		clientAddr = "198.51.100.9"
	)

	tests := []struct {
		name       string
		remoteAddr string
		connection string
		wantIP     string
	}{
		{
			name:       "trusted peer, the address header listed",
			remoteAddr: proxyAddr + ":1111",
			connection: "X-Forwarded-For",
			wantIP:     clientAddr,
		},
		{
			name:       "trusted peer, listed among other tokens",
			remoteAddr: proxyAddr + ":1111",
			connection: "keep-alive, X-Forwarded-For, X-Forwarded-Proto",
			wantIP:     clientAddr,
		},
		{
			name:       "untrusted peer, the address header listed",
			remoteAddr: "192.0.2.5:1111",
			connection: "X-Forwarded-For",
			wantIP:     "192.0.2.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := NewTrustedProxyConfig(addrs(proxyAddr), "X-Forwarded-For")
			strip := NewStrip("", "", proxy, FrontExemption{})

			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			r.Header.Set("X-Forwarded-For", clientAddr)
			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set("Connection", tt.connection)

			var got recorder
			strip.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got.header = r.Header.Clone()
				if ip := ClientIP(r, proxy); ip != tt.wantIP {
					t.Errorf("ClientIP = %q, want %q", ip, tt.wantIP)
				}
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(httptest.NewRecorder(), r)

			if tt.wantIP == clientAddr && got.header.Get("X-Forwarded-For") != clientAddr {
				t.Errorf("X-Forwarded-For = %q, want %q: a Connection listing must not delete what group 2 preserved for a trusted front", got.header.Get("X-Forwarded-For"), clientAddr)
			}
		})
	}
}

// TestStripWrapConnectionStillRemovesAnOrdinaryListedHeader is the
// negative control for the keep above: the trusted-peer carve-out
// covers group 2's names and nothing else, so an ordinary header a
// request lists in Connection is still removed from a trusted peer.
func TestStripWrapConnectionStillRemovesAnOrdinaryListedHeader(t *testing.T) {
	proxy := NewTrustedProxyConfig(addrs("203.0.113.10"), "X-Forwarded-For")
	strip := NewStrip("", "", proxy, FrontExemption{})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.10:1111"
	r.Header.Set("X-House-Header", "value")
	r.Header.Set("Connection", "X-House-Header")

	var got recorder
	strip.Wrap(recordingHandler(&got)).ServeHTTP(httptest.NewRecorder(), r)

	if v := got.header.Get("X-House-Header"); v != "" {
		t.Errorf("X-House-Header = %q, want removed: only group 2's own names are kept for a trusted peer", v)
	}
}

// TestStripWrapScrubsTrailerOnCloseWithoutRead closes the trailer
// scrub's other arming point. net/http's own body Close drains the
// remaining chunked stream inside its reader rather than through the
// wrapper's Read, so a handler that closes the body without reading it
// reaches end of stream, mergeSetHeader populates r.Trailer with the
// smuggled value, and the Read-armed scrub never fired. The handler
// here does exactly what such a handler does: Close, then read the
// trailer.
func TestStripWrapScrubsTrailerOnCloseWithoutRead(t *testing.T) {
	var afterTrailer string
	srv := httptest.NewServer(NewStrip("", "", TrustedProxyConfig{}, FrontExemption{}).Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = r.Body.Close()
			afterTrailer = r.Trailer.Get("Remote-User")
			w.WriteHeader(http.StatusOK)
		}),
	))
	defer srv.Close()

	pr, pw := io.Pipe()
	req, err := http.NewRequest(http.MethodPost, srv.URL, pr)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Trailer", "Remote-User")
	req.Trailer = http.Header{"Remote-User": nil}

	go func() {
		_, _ = pw.Write([]byte("body"))
		req.Trailer.Set("Remote-User", "attacker")
		_ = pw.Close()
	}()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_ = resp.Body.Close()

	if afterTrailer != "" {
		t.Errorf("after Close without Read: r.Trailer[Remote-User] = %q, want removed", afterTrailer)
	}
}
