package serve

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

// newTestShell builds a [Shell] over testPrefix (prefix_test.go), a
// [Strip] with sessionCookie as its session-cookie name and no
// trusted-proxy or front-exemption configuration, none of which this
// file's order and extension-point assertions exercise. The driver
// header is "X-Driver-Identity", matching expectedIdentitySpellings
// (strip_test.go), which already carries that name and its underscore
// twin as the written-out proof of the driver-contribution mechanism.
func newTestShell(t *testing.T, sessionCookie string) *Shell {
	t.Helper()
	prefix, err := NewPrefix(testPrefix)
	if err != nil {
		t.Fatalf("NewPrefix(%q): %v", testPrefix, err)
	}
	strip := NewStrip("X-Driver-Identity", sessionCookie, TrustedProxyConfig{}, FrontExemption{})
	return NewShell(prefix, strip, nil)
}

// capturingHandler is a [slog.Handler] that keeps every record's
// attributes in memory, so a test can assert on the shell's own
// structured log record against the production root handler rather
// than against a fake.
type capturingHandler struct {
	mu      sync.Mutex
	records []map[string]any
}

func (h *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, attrs)
	return nil
}

func (h *capturingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *capturingHandler) WithGroup(string) slog.Handler      { return h }

func (h *capturingHandler) snapshot() []map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]map[string]any(nil), h.records...)
}

// TestOrderStripBeforeOwnedSubtree is the first order assertion
// (spec:938-944), moved to serve.go's owned-prefix half: a request
// bearing every spelling in the strip set reaches a handler registered
// under the owned prefix with none of them.
func TestOrderStripBeforeOwnedSubtree(t *testing.T) {
	shell := newTestShell(t, "")
	var got recorder
	shell.Register("echo", recordingHandler(&got))

	r := httptest.NewRequest(http.MethodGet, testPrefix+"/echo", nil)
	for _, name := range expectedIdentitySpellings {
		r.Header.Set(name, "attacker")
	}

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if got.header == nil {
		t.Fatal("the registered subtree was never invoked")
	}
	for _, name := range expectedIdentitySpellings {
		if v := got.header.Get(name); v != "" {
			t.Errorf("owned-prefix handler received %s = %q, want removed", name, v)
		}
	}
}

// TestOrderStripBeforeFallthrough is the on-the-fallthrough half of the
// same assertion: the same spelling set, on a path outside the owned
// prefix, carries none of them either.
func TestOrderStripBeforeFallthrough(t *testing.T) {
	shell := newTestShell(t, "")
	var got recorder
	shell.SetFallthrough(recordingHandler(&got))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	for _, name := range expectedIdentitySpellings {
		r.Header.Set(name, "attacker")
	}

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if got.header == nil {
		t.Fatal("the fallthrough was never invoked")
	}
	for _, name := range expectedIdentitySpellings {
		if v := got.header.Get(name); v != "" {
			t.Errorf("fallthrough handler received %s = %q, want removed", name, v)
		}
	}
}

// TestOrderRawVsDecodedRefusedBeforeAnyHandler is the second order
// assertion: a request whose escaped and decoded paths disagree is
// refused, and no registered handler and no fallthrough handler is
// invoked. Normalize wraps the whole routing decision (C1 step 2), so
// both rows matter: a mismatch under the owned prefix, where the
// registered subtree would otherwise see it, and one outside it, where
// only the fallthrough would.
func TestOrderRawVsDecodedRefusedBeforeAnyHandler(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"under the owned prefix", "GET " + testPrefix + "/echo/%2e%2e HTTP/1.1\r\nHost: example.test\r\n\r\n"},
		{"outside the owned prefix, on the fallthrough", "GET /outside%2e%2e HTTP/1.1\r\nHost: example.test\r\n\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := newTestShell(t, "")
			var owned, fall recorder
			shell.Register("echo", recordingHandler(&owned))
			shell.SetFallthrough(recordingHandler(&fall))

			r := readRequest(t, tt.raw)

			rec := httptest.NewRecorder()
			shell.Handler().ServeHTTP(rec, r)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if owned.header != nil {
				t.Error("the owned-prefix handler was invoked on a refused request")
			}
			if fall.header != nil {
				t.Error("the fallthrough handler was invoked on a refused request")
			}
		})
	}
}

// TestOrderPrefixRequiresSegmentBoundary is the third order assertion:
// a path that string-prefixes the owned prefix without a segment
// boundary reaches the fallthrough, never the owned surface.
func TestOrderPrefixRequiresSegmentBoundary(t *testing.T) {
	shell := newTestShell(t, "")
	var owned, fall recorder
	shell.Register("foo", recordingHandler(&owned))
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, testPrefix+"foo", nil)
	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if owned.header != nil {
		t.Error("the owned-prefix handler was invoked for a path that only string-prefixes the owned prefix")
	}
	if fall.header == nil {
		t.Fatal("the fallthrough was never invoked")
	}
}

// TestRefusalLogsIdentityAlreadyGone is the fourth order assertion, the
// one the route table cannot show: a request carrying both an identity
// header and a ".." segment is refused, and the identity header was
// already gone when the refusal happened. The mechanism is a
// structured log record on the refusal path, asserted through a test
// logger against the production root handler, never by reading the
// chain's construction.
func TestRefusalLogsIdentityAlreadyGone(t *testing.T) {
	handler := &capturingHandler{}
	prefix, err := NewPrefix(testPrefix)
	if err != nil {
		t.Fatalf("NewPrefix(%q): %v", testPrefix, err)
	}
	strip := NewStrip("", "", TrustedProxyConfig{}, FrontExemption{})
	shell := NewShell(prefix, strip, slog.New(handler))

	var owned, fall recorder
	shell.Register("echo", recordingHandler(&owned))
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, testPrefix+"/../secret", nil)
	r.Header.Set("Remote-User", "attacker")

	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, r)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if owned.header != nil {
		t.Error("the owned-prefix handler was invoked on a refused request")
	}
	if fall.header != nil {
		t.Error("the fallthrough handler was invoked on a refused request")
	}

	records := handler.snapshot()
	if len(records) != 1 {
		t.Fatalf("got %d log record(s), want 1: %v", len(records), records)
	}
	present, ok := records[0]["identity_spelling_present"].(bool)
	if !ok {
		t.Fatalf("log record carries no identity_spelling_present bool: %v", records[0])
	}
	if present {
		t.Error("identity_spelling_present = true, want false: the strip should have already removed Remote-User before normalization refused the request")
	}
}

// TestRegisterDuplicateSubtreePanics: two registrations of the same
// subtree name panic at construction, naming the subtree.
func TestRegisterDuplicateSubtreePanics(t *testing.T) {
	shell := newTestShell(t, "")
	shell.Register("dup", http.NotFoundHandler())

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Register did not panic on a duplicate subtree name")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "dup") {
			t.Errorf("panic value = %v, want a message naming %q", r, "dup")
		}
	}()
	shell.Register("dup", http.NotFoundHandler())
}

// TestSetFallthroughTwicePanics: the fallthrough slot takes at most one
// handler; a second call panics.
func TestSetFallthroughTwicePanics(t *testing.T) {
	shell := newTestShell(t, "")
	shell.SetFallthrough(http.NotFoundHandler())

	defer func() {
		if recover() == nil {
			t.Fatal("SetFallthrough did not panic on a second call")
		}
	}()
	shell.SetFallthrough(http.NotFoundHandler())
}

// TestEmptyFallthroughAnswers404: with no fallthrough registered, a
// path outside the owned prefix answers 404 and invokes no handler.
func TestEmptyFallthroughAnswers404(t *testing.T) {
	shell := newTestShell(t, "")

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, r)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// cookieNames returns the names got's Cookie header carries, decoded
// through the stdlib cookie parser rather than a substring check.
func cookieNames(got *recorder) []string {
	if got.header == nil {
		return nil
	}
	req := &http.Request{Header: got.header}
	var names []string
	for _, c := range req.Cookies() {
		names = append(names, c.Name)
	}
	return names
}

// cookieValue returns the value got's Cookie header carries for name,
// decoded through the stdlib cookie parser, or "" if name is absent.
func cookieValue(got *recorder, name string) string {
	if got.header == nil {
		return ""
	}
	req := &http.Request{Header: got.header}
	c, err := req.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

const testSessionCookie = "dubplate_session"

// TestSessionCookieRemovedAtFallthroughPresentAtOwnedPrefix: the
// session cookie is absent at the fallthrough handler and present at a
// handler registered under the owned prefix, since the removal is on
// the way out and not a second strip group.
func TestSessionCookieRemovedAtFallthroughPresentAtOwnedPrefix(t *testing.T) {
	shell := newTestShell(t, testSessionCookie)
	var owned, fall recorder
	shell.Register("echo", recordingHandler(&owned))
	shell.SetFallthrough(recordingHandler(&fall))
	root := shell.Handler()

	ownedReq := httptest.NewRequest(http.MethodGet, testPrefix+"/echo", nil)
	ownedReq.AddCookie(&http.Cookie{Name: testSessionCookie, Value: "abc"})
	root.ServeHTTP(httptest.NewRecorder(), ownedReq)

	if !slices.Contains(cookieNames(&owned), testSessionCookie) {
		t.Errorf("owned-prefix handler cookies = %v, want %q present", cookieNames(&owned), testSessionCookie)
	}

	fallReq := httptest.NewRequest(http.MethodGet, "/outside", nil)
	fallReq.AddCookie(&http.Cookie{Name: testSessionCookie, Value: "abc"})
	root.ServeHTTP(httptest.NewRecorder(), fallReq)

	if slices.Contains(cookieNames(&fall), testSessionCookie) {
		t.Errorf("fallthrough handler cookies = %v, want %q removed", cookieNames(&fall), testSessionCookie)
	}
}

// TestSessionCookieOtherCookiesSurviveAtFallthrough: a request carrying
// three cookies, including the named one, arrives at the fallthrough
// with the other two unchanged in both name and value. Asserting names
// alone would pass an implementation that rebuilds each surviving
// cookie's text (for example by round-tripping through
// [http.Request.Cookies], which strips RFC 6265 double quotes on
// read), which is why this row checks the exact surviving value too.
func TestSessionCookieOtherCookiesSurviveAtFallthrough(t *testing.T) {
	shell := newTestShell(t, testSessionCookie)
	var fall recorder
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	r.AddCookie(&http.Cookie{Name: "a", Value: "1"})
	r.AddCookie(&http.Cookie{Name: testSessionCookie, Value: "abc"})
	r.AddCookie(&http.Cookie{Name: "b", Value: "2"})

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	names := cookieNames(&fall)
	if slices.Contains(names, testSessionCookie) {
		t.Errorf("fallthrough handler cookies = %v, want %q removed", names, testSessionCookie)
	}
	if got, want := cookieValue(&fall, "a"), "1"; got != want {
		t.Errorf("fallthrough handler cookie %q = %q, want %q", "a", got, want)
	}
	if got, want := cookieValue(&fall, "b"), "2"; got != want {
		t.Errorf("fallthrough handler cookie %q = %q, want %q", "b", got, want)
	}
}

// TestSessionCookieRawHeaderSurvivesVerbatim: a request whose raw
// Cookie header carries a double-quoted value arrives at the
// fallthrough with that exact header text minus the named cookie, not
// a value net/http's cookie parser has normalized. This is the row
// that would catch an implementation rebuilding the header from
// [http.Request.Cookies], which the row above's name-and-value check
// cannot distinguish from a correct one when the surviving value has
// no quotes to lose.
func TestSessionCookieRawHeaderSurvivesVerbatim(t *testing.T) {
	shell := newTestShell(t, testSessionCookie)
	var fall recorder
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	r.Header.Set("Cookie", `a="quoted"; `+testSessionCookie+`=abc; b=2`)

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if fall.header == nil {
		t.Fatal("the fallthrough handler was never invoked")
	}
	want := `a="quoted"; b=2`
	if got := fall.header.Get("Cookie"); got != want {
		t.Errorf("fallthrough handler Cookie header = %q, want %q", got, want)
	}
}

// TestSessionCookieDuplicateNamedCookieBothRemoved: a request carrying
// the named cookie twice loses both instances at the fallthrough,
// never leaving one under arrival order.
func TestSessionCookieDuplicateNamedCookieBothRemoved(t *testing.T) {
	shell := newTestShell(t, testSessionCookie)
	var fall recorder
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	r.AddCookie(&http.Cookie{Name: testSessionCookie, Value: "first"})
	r.AddCookie(&http.Cookie{Name: testSessionCookie, Value: "second"})

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if names := cookieNames(&fall); slices.Contains(names, testSessionCookie) {
		t.Errorf("fallthrough handler cookies = %v, want %q removed entirely", names, testSessionCookie)
	}
}

// TestTwoSubtreesEachReceiveOnlyOwnPaths admits the static surface's
// two-handler shape (C5, DL R4.5) without building it: two distinct
// subtrees registered at construction, each receiving only its own
// paths.
func TestTwoSubtreesEachReceiveOnlyOwnPaths(t *testing.T) {
	shell := newTestShell(t, "")
	var alpha, beta recorder
	shell.Register("alpha", recordingHandler(&alpha))
	shell.Register("beta", recordingHandler(&beta))
	root := shell.Handler()

	root.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, testPrefix+"/alpha/x", nil))
	if alpha.header == nil {
		t.Error("alpha was not invoked for its own path")
	}
	if beta.header != nil {
		t.Error("beta was invoked for a path under alpha")
	}

	alpha.header = nil
	root.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, testPrefix+"/beta/y", nil))
	if beta.header == nil {
		t.Error("beta was not invoked for its own path")
	}
	if alpha.header != nil {
		t.Error("alpha was invoked for a path under beta")
	}
}

// flushingHandler calls http.NewResponseController(w).Flush() and
// records the error it returns, so a test can prove the writer a
// handler receives is not opaque to [http.ResponseController]: logRefusal
// wraps the whole chain, and rung 12's reverse proxy occupies the
// fallthrough this handler stands in for, so a writer that answers
// "feature not supported" here would defeat the streaming rung 12
// needs.
func flushingHandler(dst *error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*dst = http.NewResponseController(w).Flush()
		w.WriteHeader(http.StatusOK)
	})
}

// TestResponseControllerFlushReachesRealWriter: logRefusal's status
// capture stays transparent to http.ResponseController at both a
// registered subtree and the fallthrough, proving the root handler
// never hands rung 12's proxy an opaque writer.
func TestResponseControllerFlushReachesRealWriter(t *testing.T) {
	shell := newTestShell(t, "")
	var ownedErr, fallErr error
	shell.Register("echo", flushingHandler(&ownedErr))
	shell.SetFallthrough(flushingHandler(&fallErr))
	root := shell.Handler()

	root.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, testPrefix+"/echo", nil))
	if ownedErr != nil {
		t.Errorf("Flush() at the owned prefix = %v, want nil", ownedErr)
	}

	root.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/outside", nil))
	if fallErr != nil {
		t.Errorf("Flush() at the fallthrough = %v, want nil", fallErr)
	}
}

// cookieLines returns every Cookie header line got carries, in order,
// as the raw text the handler received.
func cookieLines(got *recorder) []string {
	if got.header == nil {
		return nil
	}
	return got.header["Cookie"]
}

// TestSessionCookieRemovedFromEveryCookieLine: a client may send Cookie
// more than once, and net/http keeps each line separately. The named
// cookie goes from every line, each surviving line keeps its own text
// byte for byte (the quoted value included), and no line is folded
// into another. An implementation reading only the first line would
// leave the second line's session cookie standing and collapse the
// rest of the request's cookies into nothing.
func TestSessionCookieRemovedFromEveryCookieLine(t *testing.T) {
	shell := newTestShell(t, testSessionCookie)
	var fall recorder
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	r.Header.Add("Cookie", `a="quoted"; `+testSessionCookie+`=first`)
	r.Header.Add("Cookie", testSessionCookie+`=second; b=2`)

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if fall.header == nil {
		t.Fatal("the fallthrough handler was never invoked")
	}
	want := []string{`a="quoted"`, `b=2`}
	if got := cookieLines(&fall); !slices.Equal(got, want) {
		t.Errorf("fallthrough handler Cookie lines = %q, want %q", got, want)
	}
}

// TestSessionCookieRemovedFirstInLineLeavesNoLeadingSpace: dropping the
// first segment of a line must not leave the separator's own space at
// the head of the value, which would hand the fallthrough a Cookie
// header no client would have sent.
func TestSessionCookieRemovedFirstInLineLeavesNoLeadingSpace(t *testing.T) {
	shell := newTestShell(t, testSessionCookie)
	var fall recorder
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	r.Header.Set("Cookie", testSessionCookie+`=abc; a=1; b=2`)

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if fall.header == nil {
		t.Fatal("the fallthrough handler was never invoked")
	}
	want := `a=1; b=2`
	if got := fall.header.Get("Cookie"); got != want {
		t.Errorf("fallthrough handler Cookie header = %q, want %q", got, want)
	}
}

// TestSessionCookieOnlyCookieOnALineDropsTheLine: a line carrying
// nothing but the named cookie leaves no empty Cookie header behind.
func TestSessionCookieOnlyCookieOnALineDropsTheLine(t *testing.T) {
	shell := newTestShell(t, testSessionCookie)
	var fall recorder
	shell.SetFallthrough(recordingHandler(&fall))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	r.Header.Add("Cookie", testSessionCookie+`=abc`)
	r.Header.Add("Cookie", `a=1`)

	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if fall.header == nil {
		t.Fatal("the fallthrough handler was never invoked")
	}
	want := []string{`a=1`}
	if got := cookieLines(&fall); !slices.Equal(got, want) {
		t.Errorf("fallthrough handler Cookie lines = %q, want %q", got, want)
	}
}

// TestUnclaimedOwnedPathReachesTheTerminator: a path under the owned
// prefix that no registered subtree claims answers 404 through
// [Terminator], reaching neither a registered subtree nor the
// fallthrough. Rung 14 depends on this: an unknown owned path must not
// reach the proxy the fallthrough holds.
func TestUnclaimedOwnedPathReachesTheTerminator(t *testing.T) {
	shell := newTestShell(t, "")
	var owned, fall recorder
	shell.Register("echo", recordingHandler(&owned))
	shell.SetFallthrough(recordingHandler(&fall))

	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testPrefix+"/unclaimed", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if owned.header != nil {
		t.Error("the registered subtree was invoked for a path it does not claim")
	}
	if fall.header != nil {
		t.Error("the fallthrough was invoked for a path under the owned prefix; the terminator must stop it")
	}
}

// TestRegisterAfterHandlerPanics: registration belongs to
// construction. A registration after the root chain is composed lands
// on a mux an already-serving binary is reading, so it panics naming
// the subtree rather than racing.
func TestRegisterAfterHandlerPanics(t *testing.T) {
	shell := newTestShell(t, "")
	shell.Handler()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Register did not panic after Handler")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "late") {
			t.Errorf("panic value = %v, want a message naming %q", r, "late")
		}
	}()
	shell.Register("late", http.NotFoundHandler())
}

// TestRefusalRecordOnlyForTheShellsOwnRefusals: a registered subtree
// answering 400 to its own bad input is not an ingress refusal and
// emits no record, while the shell's own normalization refusal on the
// same subtree does. Keying the record on the response status alone
// would report an ingress refusal every time a handler rejected a
// malformed request of its own.
func TestRefusalRecordOnlyForTheShellsOwnRefusals(t *testing.T) {
	newShellWithLog := func(t *testing.T) (*Shell, *capturingHandler) {
		t.Helper()
		prefix, err := NewPrefix(testPrefix)
		if err != nil {
			t.Fatalf("NewPrefix(%q): %v", testPrefix, err)
		}
		log := &capturingHandler{}
		strip := NewStrip("", "", TrustedProxyConfig{}, FrontExemption{})
		return NewShell(prefix, strip, slog.New(log)), log
	}

	t.Run("a subtree's own 400 emits no record", func(t *testing.T) {
		shell, log := newShellWithLog(t)
		shell.Register("echo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad form", http.StatusBadRequest)
		}))

		rec := httptest.NewRecorder()
		shell.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, testPrefix+"/echo", nil))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if records := log.snapshot(); len(records) != 0 {
			t.Errorf("got %d log record(s) for a handler's own 400, want none: %v", len(records), records)
		}
	})

	t.Run("the shell's own refusal on the same subtree emits one", func(t *testing.T) {
		shell, log := newShellWithLog(t)
		shell.Register("echo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad form", http.StatusBadRequest)
		}))

		rec := httptest.NewRecorder()
		shell.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testPrefix+"/echo/../secret", nil))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if records := log.snapshot(); len(records) != 1 {
			t.Fatalf("got %d log record(s) for the shell's own refusal, want 1: %v", len(records), records)
		}
	})
}

// readerFromRecorder is an http.ResponseWriter that implements
// [io.ReaderFrom], standing in for net/http's own response writer,
// which implements it to reach sendfile(2).
type readerFromRecorder struct {
	*httptest.ResponseRecorder
	used bool
}

func (w *readerFromRecorder) ReadFrom(src io.Reader) (int64, error) {
	w.used = true
	return io.Copy(w.Body, src)
}

// plainReader hides the io.WriterTo a strings.Reader carries. io.Copy
// consults the source for io.WriterTo before the destination for
// io.ReaderFrom, so a bare strings.Reader would take its own fast path
// and the test would prove nothing about the destination.
type plainReader struct{ io.Reader }

// TestReadFromReachesRealWriter: the status capture forwards
// [io.ReaderFrom] to the writer beneath it, so a handler copying a
// stream through io.Copy keeps the underlying writer's fast path.
// Without the forward, io.Copy would find no ReaderFrom on the capture
// and silently fall back to a userspace copy for every stream rung
// 12's proxy serves.
func TestReadFromReachesRealWriter(t *testing.T) {
	shell := newTestShell(t, "")
	shell.SetFallthrough(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(w, plainReader{strings.NewReader("payload")}); err != nil {
			t.Errorf("io.Copy: %v", err)
		}
	}))

	w := &readerFromRecorder{ResponseRecorder: httptest.NewRecorder()}
	shell.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/outside", nil))

	if !w.used {
		t.Error("io.Copy did not reach the underlying writer's ReadFrom, want the capture to forward it")
	}
	if got := w.Body.String(); got != "payload" {
		t.Errorf("body = %q, want %q", got, "payload")
	}
}

// TestIdentitySpellingPresentReadsKeyPresence: the refusal record says
// whether an identity header arrived, not whether it arrived carrying
// something. A header sent with an empty value is one the client sent,
// and a value test would report it as absent.
func TestIdentitySpellingPresentReadsKeyPresence(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		want   bool
	}{
		{
			name:   "an empty-valued identity header is present",
			header: http.Header{"Remote-User": []string{""}},
			want:   true,
		},
		{
			name:   "a valued identity header is present",
			header: http.Header{"Remote-User": []string{"attacker"}},
			want:   true,
		},
		{
			name:   "an unrelated header is not",
			header: http.Header{"X-App-Debug-Trace": []string{"trace-id"}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header = tt.header
			if got := identitySpellingPresent(r, ""); got != tt.want {
				t.Errorf("identitySpellingPresent() = %t, want %t", got, tt.want)
			}
		})
	}
}
