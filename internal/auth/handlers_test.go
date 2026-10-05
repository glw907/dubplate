package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/glw907/dubplate/internal/serve"
)

// sessionFixture signs the sessions a consumed link mints, and
// foreignFixture signs a link under a secret the handlers do not hold.
// Both are visibly synthetic, like token_test.go's fixtures.
const (
	sessionFixture = "dubplate-fixture-session-0000000"
	foreignFixture = "dubplate-fixture-foreign-0000000"
)

// handlerNow is the instant every handler request is evaluated at, and
// handlerOrigin the configured public URL mailed links are built from.
var (
	handlerNow    = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	handlerOrigin = "https://music.example.com"
)

// handlerPrefix is the owned prefix every fixture's mailed links build
// under, on servePrefixThreshold's own default.
const handlerPrefix = "/dubplate"

// handlerRoster carries one admin and four members, so a flood row has
// distinct known addresses to walk and the kind answer has an admin to
// name.
const handlerRoster = `[
	{"address":"admin@example.com","display_name":"Group Admin","admin":true},
	{"address":"member@example.com","display_name":"Member","admin":false},
	{"address":"member1@example.com","display_name":"Member One","admin":false},
	{"address":"member2@example.com","display_name":"Member Two","admin":false},
	{"address":"member3@example.com","display_name":"Member Three","admin":false}
]`

// generousLimits admits every request a row sends unless the row sets
// its own.
var generousLimits = Limits{
	LinkRequestsPerAddress:       100,
	LinkRequestsPerClient:        100,
	ConsumptionAttemptsPerClient: 100,
	LinkMailCeiling:              100,
}

// peerIP is the client address every request dials from unless a row
// sets its own, and peerIPPort the RemoteAddr carrying it.
const (
	peerIP     = "192.0.2.10"
	peerIPPort = peerIP + ":40000"
)

// sentMail is one call to fakeMail.Send.
type sentMail struct {
	to, subject, body string
}

// fakeMail is the handlers' mail seam. err fails every send; release,
// when set, holds every send until it is closed; maxCalls, when
// positive, fails the test on any call past it.
type fakeMail struct {
	t        *testing.T
	err      error
	release  chan struct{}
	maxCalls int

	mu    sync.Mutex
	sent  []sentMail
	calls chan sentMail
}

func newFakeMail(t *testing.T) *fakeMail {
	return &fakeMail{t: t, calls: make(chan sentMail, 64)}
}

func (f *fakeMail) Send(ctx context.Context, to, subject, body string) error {
	m := sentMail{to: to, subject: subject, body: body}
	f.mu.Lock()
	f.sent = append(f.sent, m)
	n := len(f.sent)
	f.mu.Unlock()
	if f.maxCalls > 0 && n > f.maxCalls {
		f.t.Errorf("mail seam called %d times, want at most %d", n, f.maxCalls)
	}
	f.calls <- m
	if f.release != nil {
		<-f.release
	}
	return f.err
}

func (f *fakeMail) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

// waitCall waits for the next send, failing the test when none arrives:
// a generic-answer send runs after the response, so a row reads it
// here rather than straight after ServeHTTP returns.
func (f *fakeMail) waitCall(t *testing.T) sentMail {
	t.Helper()
	select {
	case m := <-f.calls:
		return m
	case <-time.After(5 * time.Second):
		t.Fatal("no mail was sent")
		return sentMail{}
	}
}

// logCapture collects the handlers' JSON log records and wakes a
// waiter on each one, so a row can read a record written after the
// response without sleeping.
type logCapture struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	wakeup chan struct{}
}

func newLogCapture() *logCapture {
	return &logCapture{wakeup: make(chan struct{}, 1)}
}

func (c *logCapture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf.Write(p)
	select {
	case c.wakeup <- struct{}{}:
	default:
	}
	return len(p), nil
}

func (c *logCapture) text() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

func (c *logCapture) records(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for line := range strings.Lines(c.text()) {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("log line %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

// waitFor returns the first record whose msg is msg, waiting for one to
// be written when none has been yet.
func (c *logCapture) waitFor(t *testing.T, msg string) map[string]any {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		for _, rec := range c.records(t) {
			if rec["msg"] == msg {
				return rec
			}
		}
		select {
		case <-c.wakeup:
		case <-deadline:
			t.Fatalf("no log record %q in:\n%s", msg, c.text())
		}
	}
}

// fixtureOptions are what a row varies; the zero value is the kind
// answer under generousLimits with ClientIP in direct mode.
type fixtureOptions struct {
	generic    bool
	limits     Limits
	proxy      serve.TrustedProxyConfig
	bodyCap    int64
	mailErr    error
	maxCalls   int
	linkExpiry time.Duration
}

type handlerFixture struct {
	handlers   Handlers
	stateDir   string
	rosterPath string
	epoch      *EpochStore
	limiter    *Limiter
	mail       *fakeMail
	logs       *logCapture
}

func newHandlerFixture(t *testing.T, opts fixtureOptions) handlerFixture {
	t.Helper()
	stateDir := t.TempDir()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	nonces, _, err := OpenNonceStore(store, fixtureRetention, fixtureLinkExpiryCeiling, handlerNow)
	if err != nil {
		t.Fatalf("OpenNonceStore: %v", err)
	}
	epoch := OpenEpochStore(store)
	rosterPath := writeRosterFile(t, `{"schema_version":1,"members":`+handlerRoster+`}`)
	roster, err := OpenRoster(rosterPath)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}

	limits := opts.limits
	if limits == (Limits{}) {
		limits = generousLimits
	}
	bodyCap := opts.bodyCap
	if bodyCap == 0 {
		bodyCap = 4096
	}
	mail := newFakeMail(t)
	mail.err = opts.mailErr
	mail.maxCalls = opts.maxCalls
	logs := newLogCapture()
	limiter := NewLimiter(limits)

	linkExpiry := opts.linkExpiry
	if linkExpiry == 0 {
		linkExpiry = fixtureExpiry
	}

	h := NewHandlers(HandlersConfig{
		Roster:             roster,
		Nonces:             nonces,
		Epoch:              epoch,
		Limiter:            limiter,
		Send:               mail.Send,
		ClientIP:           func(r *http.Request) string { return serve.ClientIP(r, opts.proxy) },
		Now:                func() time.Time { return handlerNow },
		Logger:             slog.New(slog.NewJSONHandler(logs, nil)),
		LinkSecret:         []byte(currentFixture),
		PreviousLinkSecret: []byte(previousFixture),
		SessionSecret:      []byte(sessionFixture),
		LinkExpiry:         linkExpiry,
		SessionLifetime:    seamLifetime,
		SessionAbsoluteCap: seamCap,
		PublicOrigin:       handlerOrigin,
		Prefix:             handlerPrefix,
		BodyCap:            bodyCap,
		GenericAnswer:      opts.generic,
	})
	return handlerFixture{handlers: h, stateDir: stateDir, rosterPath: rosterPath, epoch: epoch, limiter: limiter, mail: mail, logs: logs}
}

// post builds a urlencoded POST to target from peer carrying body.
func post(target, peer, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.RemoteAddr = peer
	return r
}

func serveOne(h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func requestLinkBody(address string) string {
	return url.Values{addressField: {address}}.Encode()
}

func consumeBody(token string) string {
	return url.Values{tokenField: {token}}.Encode()
}

// mintFixtureLink mints a link for address the way a mailed one is,
// under secret, at mintedAt.
func mintFixtureLink(t *testing.T, secret, address string, mintedAt time.Time) string {
	t.Helper()
	token, _, err := MintLink([]byte(secret), address, mintedAt, fixtureExpiry)
	if err != nil {
		t.Fatalf("MintLink: %v", err)
	}
	return token
}

// outcomeOf decodes a response's outcome field.
func outcomeOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var a answer
	if err := json.Unmarshal(w.Body.Bytes(), &a); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return a.Outcome
}

// response is everything a client receives from a recorded response.
type response struct {
	status int
	header http.Header
	body   string
}

func responseOf(w *httptest.ResponseRecorder) response {
	return response{status: w.Code, header: w.Header().Clone(), body: w.Body.String()}
}

// TestConsumeLink_SignsIn proves the success path: a mailed link's
// token, posted back, sets a session cookie for its address and
// consumes exactly one nonce.
func TestConsumeLink_SignsIn(t *testing.T) {
	fx := newHandlerFixture(t, fixtureOptions{})
	w := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("Member@Example.com")))
	if got := outcomeOf(t, w); got != "sent" {
		t.Fatalf("request outcome = %q, want sent", got)
	}
	token := tokenFromMail(t, fx.mail.waitCall(t))

	w = serveOne(fx.handlers.ConsumeLink, post("/login/link", peerIPPort, consumeBody(token)))
	if w.Code != http.StatusOK || outcomeOf(t, w) != "signed_in" {
		t.Fatalf("consume = %d %s, want 200 signed_in", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "__Host-dubplate_session" {
		t.Fatalf("cookies = %v, want one __Host-dubplate_session", cookies)
	}
	sess, err := VerifySession([]byte(sessionFixture), nil, cookies[0].Value, handlerNow, seamCap)
	if err != nil {
		t.Fatalf("VerifySession: %v", err)
	}
	if sess.Address != "member@example.com" {
		t.Errorf("session address = %q, want member@example.com", sess.Address)
	}
	if got := len(nonceDirEntries(t, fx.stateDir)); got != 1 {
		t.Errorf("nonce records = %d, want 1", got)
	}
}

// tokenFromMail extracts the token from a mailed link, checking the link
// is built on the configured origin rather than on anything the request
// carried.
func tokenFromMail(t *testing.T, m sentMail) string {
	t.Helper()
	for field := range strings.FieldsSeq(m.body) {
		if !strings.HasPrefix(field, "https://") {
			continue
		}
		u, err := url.Parse(field)
		if err != nil {
			t.Fatalf("parse mailed link: %v", err)
		}
		if u.Scheme+"://"+u.Host != handlerOrigin {
			t.Fatalf("mailed link origin = %s://%s, want %s", u.Scheme, u.Host, handlerOrigin)
		}
		return u.Query().Get("t")
	}
	t.Fatalf("no link in mail body %q", m.body)
	return ""
}

// TestConsumeLink_Refusals proves the consumption order by what each
// refusal leaves behind rather than by reading the function: a refusal
// before the insert leaves the nonce directory empty, and no refusal
// sets a cookie. Across rows, every client response is byte-identical
// while the server log names a distinct cause for each.
//
// The unreadable-epoch row is the non-sentinel refusal: the epoch
// directory is unreadable, so the epoch check fails with an I/O error
// rather than a named refusal, and it must deny all the same.
func TestConsumeLink_Refusals(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, fx handlerFixture) string
		wantNonces int
	}{
		{
			name: "MAC fails",
			setup: func(t *testing.T, fx handlerFixture) string {
				return mintFixtureLink(t, foreignFixture, "member@example.com", handlerNow)
			},
		},
		{
			name: "link expired",
			setup: func(t *testing.T, fx handlerFixture) string {
				return mintFixtureLink(t, currentFixture, "member@example.com", handlerNow.Add(-2*fixtureExpiry))
			},
		},
		{
			name: "address not on the roster",
			setup: func(t *testing.T, fx handlerFixture) string {
				return mintFixtureLink(t, currentFixture, "removed@example.com", handlerNow)
			},
		},
		{
			name: "link minted before the epoch",
			setup: func(t *testing.T, fx handlerFixture) string {
				token := mintFixtureLink(t, currentFixture, "member@example.com", handlerNow.Add(-time.Minute))
				if err := fx.epoch.Bump("member@example.com", handlerNow); err != nil {
					t.Fatalf("Bump: %v", err)
				}
				return token
			},
		},
		{
			name: "epoch unreadable",
			setup: func(t *testing.T, fx handlerFixture) string {
				if os.Geteuid() == 0 {
					t.Skip("running as root: permission bits do not block reads")
				}
				dir := filepath.Join(StoreDir(fx.stateDir), epochSubdir)
				if err := os.Chmod(dir, 0o000); err != nil {
					t.Fatalf("Chmod: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
				return mintFixtureLink(t, currentFixture, "member@example.com", handlerNow)
			},
		},
		{
			name: "replayed link",
			setup: func(t *testing.T, fx handlerFixture) string {
				token := mintFixtureLink(t, currentFixture, "member@example.com", handlerNow)
				w := serveOne(fx.handlers.ConsumeLink, post("/login/link", peerIPPort, consumeBody(token)))
				if outcomeOf(t, w) != "signed_in" {
					t.Fatalf("first presentation = %s, want signed_in", w.Body.String())
				}
				return token
			},
			wantNonces: 1,
		},
	}

	responses := map[string]response{}
	reasons := map[string]string{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newHandlerFixture(t, fixtureOptions{})
			token := tt.setup(t, fx)
			logStart := len(fx.logs.records(t))

			w := serveOne(fx.handlers.ConsumeLink, post("/login/link", peerIPPort, consumeBody(token)))
			if got := w.Header().Values("Set-Cookie"); len(got) != 0 {
				t.Errorf("Set-Cookie = %q, want none", got)
			}
			if got := len(nonceDirEntries(t, fx.stateDir)); got != tt.wantNonces {
				t.Errorf("nonce records = %d, want %d", got, tt.wantNonces)
			}
			rec := fx.logs.records(t)[logStart:]
			if len(rec) != 1 || rec[0]["msg"] != "auth link consumption failed" {
				t.Fatalf("log records = %v, want one consumption failure", rec)
			}
			reason, _ := rec[0]["reason"].(string)
			responses[tt.name] = responseOf(w)
			reasons[tt.name] = reason
		})
	}

	var first response
	var firstName string
	seen := map[string]string{}
	for _, tt := range tests {
		got, ok := responses[tt.name]
		if !ok {
			continue
		}
		if firstName == "" {
			first, firstName = got, tt.name
		} else if !reflect.DeepEqual(got, first) {
			t.Errorf("response for %q = %+v, differs from %q's %+v", tt.name, got, firstName, first)
		}
		if other, dup := seen[reasons[tt.name]]; dup {
			t.Errorf("log reason %q names both %q and %q", reasons[tt.name], other, tt.name)
		}
		seen[reasons[tt.name]] = tt.name
	}
}

// TestConsumeLink_LimiterPrecedesMAC floods the consumption endpoint
// from one client with tokens whose MAC fails. The first N reach the
// verifier; the N+1th is refused by the limiter, which only a limiter
// placed ahead of the MAC can do, since every one of these tokens fails
// there.
func TestConsumeLink_LimiterPrecedesMAC(t *testing.T) {
	const attempts = 3
	limits := generousLimits
	limits.ConsumptionAttemptsPerClient = attempts
	fx := newHandlerFixture(t, fixtureOptions{limits: limits})
	forged := mintFixtureLink(t, foreignFixture, "member@example.com", handlerNow)

	for i := range attempts + 1 {
		w := serveOne(fx.handlers.ConsumeLink, post("/login/link", peerIPPort, consumeBody(forged)))
		rec := fx.logs.records(t)
		last := rec[len(rec)-1]
		wantMsg, wantStatus := "auth link consumption failed", http.StatusBadRequest
		if i == attempts {
			wantMsg, wantStatus = "auth rate limit tripped", http.StatusTooManyRequests
		}
		if last["msg"] != wantMsg || w.Code != wantStatus {
			t.Errorf("attempt %d: log %q status %d, want %q %d", i+1, last["msg"], w.Code, wantMsg, wantStatus)
		}
	}
}

// TestOversizeBodyRefusedBeforeParse posts a body one byte over the cap
// to each endpoint. Each body is a well-formed form whose parse would
// succeed and whose handling would act: send mail, mint a session, or
// clear the cookie. The control half serves the same body under a cap
// that admits it and sees that action, so the refusal is the cap's.
func TestOversizeBodyRefusedBeforeParse(t *testing.T) {
	const bodyCap = 512
	pad := func(body string) string {
		return body + "&pad=" + strings.Repeat("x", bodyCap+1-len(body)-len("&pad="))
	}
	tests := []struct {
		name  string
		serve func(t *testing.T, fx handlerFixture, body func(string) string) *httptest.ResponseRecorder
		acted func(t *testing.T, fx handlerFixture, w *httptest.ResponseRecorder) bool
	}{
		{
			name: "request link",
			serve: func(t *testing.T, fx handlerFixture, body func(string) string) *httptest.ResponseRecorder {
				return serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, body(requestLinkBody("member@example.com"))))
			},
			acted: func(t *testing.T, fx handlerFixture, w *httptest.ResponseRecorder) bool {
				return fx.mail.count() > 0
			},
		},
		{
			name: "consume link",
			serve: func(t *testing.T, fx handlerFixture, body func(string) string) *httptest.ResponseRecorder {
				token := mintFixtureLink(t, currentFixture, "member@example.com", handlerNow)
				return serveOne(fx.handlers.ConsumeLink, post("/login/link", peerIPPort, body(consumeBody(token))))
			},
			acted: func(t *testing.T, fx handlerFixture, w *httptest.ResponseRecorder) bool {
				return len(w.Header().Values("Set-Cookie")) > 0 || len(nonceDirEntries(t, fx.stateDir)) > 0
			},
		},
		{
			name: "sign out",
			serve: func(t *testing.T, fx handlerFixture, body func(string) string) *httptest.ResponseRecorder {
				return serveOne(fx.handlers.SignOut, post("/logout", peerIPPort, body("signout=1")))
			},
			acted: func(t *testing.T, fx handlerFixture, w *httptest.ResponseRecorder) bool {
				return len(w.Header().Values("Set-Cookie")) > 0
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newHandlerFixture(t, fixtureOptions{bodyCap: bodyCap})
			w := tt.serve(t, fx, func(b string) string {
				padded := pad(b)
				if len(padded) != bodyCap+1 {
					t.Fatalf("padded body is %d bytes, want %d", len(padded), bodyCap+1)
				}
				return padded
			})
			if w.Code != http.StatusRequestEntityTooLarge {
				t.Errorf("status = %d, want 413", w.Code)
			}
			if tt.acted(t, fx, w) {
				t.Errorf("an oversized body was acted on")
			}

			control := newHandlerFixture(t, fixtureOptions{bodyCap: bodyCap + 1})
			w = tt.serve(t, control, pad)
			if w.Code != http.StatusOK || !tt.acted(t, control, w) {
				t.Errorf("control: status %d, acted %v; want the same body acted on under a cap that admits it", w.Code, tt.acted(t, control, w))
			}
		})
	}
}

// TestRequestLink_GenericAnswersAlike compares the generic answer to a
// known address against its answer to an unknown one, whole response
// against whole response, once with the known address's send
// succeeding and once with it failing. A failed send is logged, never
// rendered.
func TestRequestLink_GenericAnswersAlike(t *testing.T) {
	tests := []struct {
		name    string
		mailErr error
		wantLog string
	}{
		{name: "send succeeds", wantLog: ""},
		{name: "send fails", mailErr: errors.New("smtp: connection refused"), wantLog: "auth link send failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newHandlerFixture(t, fixtureOptions{generic: true, mailErr: tt.mailErr})
			known := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
			if m := fx.mail.waitCall(t); m.to != "member@example.com" {
				t.Errorf("mail to %q, want member@example.com", m.to)
			}
			unknown := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("stranger@example.com")))

			if got, want := responseOf(known), responseOf(unknown); !reflect.DeepEqual(got, want) {
				t.Errorf("known address answered %+v, unknown answered %+v", got, want)
			}
			if tt.wantLog != "" {
				fx.logs.waitFor(t, tt.wantLog)
			}
			if got := fx.mail.count(); got != 1 {
				t.Errorf("mail sent %d times, want 1 (the unknown address sends none)", got)
			}
		})
	}
}

// TestRequestLink_GenericDoesNotWaitOnSend holds the known address's
// send open and requires the response anyway. A handler that answered
// only after the send would carry the transport's whole delay, which
// the unknown address never does; here that delay never ends until the
// response has been seen, so a waiting handler fails the deadline.
func TestRequestLink_GenericDoesNotWaitOnSend(t *testing.T) {
	fx := newHandlerFixture(t, fixtureOptions{generic: true})
	fx.mail.release = make(chan struct{})
	defer close(fx.mail.release)

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		done <- serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
	}()
	select {
	case w := <-done:
		if outcomeOf(t, w) != "requested" {
			t.Errorf("outcome = %s, want requested", w.Body.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the generic answer waited on a send that had not completed")
	}
	fx.mail.waitCall(t)
}

// TestRequestLink_KindAnswers covers the kind answer: a known address is
// mailed, an unknown one is told so and handed the roster's admin
// contact, and a known address whose send fails is told its request
// failed rather than that mail is on its way.
func TestRequestLink_KindAnswers(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		mailErr    error
		wantStatus int
		wantBody   string
		wantMail   int
	}{
		{
			name:       "known address is mailed",
			address:    "member@example.com",
			wantStatus: http.StatusOK,
			wantBody:   `{"outcome":"sent"}`,
			wantMail:   1,
		},
		{
			name:       "unknown address gets the admin contact",
			address:    "stranger@example.com",
			wantStatus: http.StatusOK,
			wantBody:   `{"outcome":"unknown","admin_contact":{"display_name":"Group Admin","address":"admin@example.com"}}`,
		},
		{
			name:       "known address whose send fails is a failed login",
			address:    "member@example.com",
			mailErr:    errors.New("smtp: connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"outcome":"failed"}`,
			wantMail:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newHandlerFixture(t, fixtureOptions{mailErr: tt.mailErr})
			w := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody(tt.address)))
			if w.Code != tt.wantStatus || strings.TrimSpace(w.Body.String()) != tt.wantBody {
				t.Errorf("got %d %s, want %d %s", w.Code, w.Body.String(), tt.wantStatus, tt.wantBody)
			}
			if got := fx.mail.count(); got != tt.wantMail {
				t.Errorf("mail sent %d times, want %d", got, tt.wantMail)
			}
		})
	}
}

// TestRequestLink_MailCopy proves the mail body a known address receives:
// the link, the expiry stated in whole minutes computed from LinkExpiry
// (fixtureExpiry is 15 minutes), the only-once and ignore-it sentences,
// and nothing else configured. Stripping the link out of the body and
// searching what remains for the roster's admin address, the public
// origin's bare host, and the sign-in address itself is what proves the
// last clause: none of them appears anywhere but inside the link this
// test already extracted. The expiry assertion checks the full "expires
// in 15 minutes." phrase, not the bare digits "15": the link's own token
// can itself contain the substring "15", so a bare-digit check could
// pass even if the digits came from the URL rather than the expiry.
func TestRequestLink_MailCopy(t *testing.T) {
	fx := newHandlerFixture(t, fixtureOptions{})
	serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
	m := fx.mail.waitCall(t)

	if m.subject == "" {
		t.Error("mail subject is empty")
	}
	if !strings.Contains(m.body, "works only once") {
		t.Errorf("mail body = %q, want it to say the link works only once", m.body)
	}
	if !strings.Contains(m.body, "ignore") {
		t.Errorf("mail body = %q, want it to tell a reader who did not ask to ignore it", m.body)
	}
	const wantExpiryPhrase = "expires in 15 minutes."
	if !strings.Contains(m.body, wantExpiryPhrase) {
		t.Errorf("mail body = %q, want it to contain %q", m.body, wantExpiryPhrase)
	}

	token := tokenFromMail(t, m)
	link, err := LinkURL(handlerOrigin, handlerPrefix, token)
	if err != nil {
		t.Fatalf("LinkURL: %v", err)
	}
	stripped := strings.Replace(m.body, link, "", 1)
	for _, forbidden := range []string{"member@example.com", "admin@example.com", "music.example.com"} {
		if strings.Contains(stripped, forbidden) {
			t.Errorf("mail body with the link removed still contains %q: %q", forbidden, stripped)
		}
	}
}

// TestRequestLink_MailCopyExpiryRounding proves sendLink rounds any
// partial minute of LinkExpiry up rather than truncating, and switches
// between "1 minute" singular and "N minutes" plural on the rounded
// count alone: a 90-second expiry must never read "1 minute", which
// truncation would produce. Each row asserts the full "expires in ..."
// phrase, not the bare digit, on the same reasoning
// TestRequestLink_MailCopy documents: the link's own token can contain
// a digit that would let a bare-digit check pass for the wrong reason.
func TestRequestLink_MailCopyExpiryRounding(t *testing.T) {
	tests := []struct {
		name       string
		linkExpiry time.Duration
		wantPhrase string
	}{
		{"exactly one minute reads singular", time.Minute, "expires in 1 minute."},
		{"ninety seconds rounds up to two minutes", 90 * time.Second, "expires in 2 minutes."},
		{"fifteen minutes reads plural", 15 * time.Minute, "expires in 15 minutes."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newHandlerFixture(t, fixtureOptions{linkExpiry: tt.linkExpiry})
			serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
			m := fx.mail.waitCall(t)

			if !strings.Contains(m.body, tt.wantPhrase) {
				t.Errorf("mail body = %q, want it to contain %q", m.body, tt.wantPhrase)
			}
		})
	}
}

// TestRequestLink_BrokenRoster breaks the roster file after the server
// has opened it, the way a bad hand edit would, and asks for a link.
// Under the kind answer a broken roster is a failed request, never an
// unknown address: answering unknown would hand a member the admin
// contact meant for a stranger, and would name that admin from a roster
// that no longer loads. Under the generic answer the response is the
// same bytes a working roster answers with, and the failure reaches
// the log alone.
func TestRequestLink_BrokenRoster(t *testing.T) {
	tests := []struct {
		name    string
		generic bool
		address string
	}{
		{name: "kind answer, known address", address: "member@example.com"},
		{name: "kind answer, unknown address", address: "stranger@example.com"},
		{name: "generic answer, known address", generic: true, address: "member@example.com"},
		{name: "generic answer, unknown address", generic: true, address: "stranger@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newHandlerFixture(t, fixtureOptions{generic: tt.generic})
			if err := os.WriteFile(fx.rosterPath, []byte(`{"schema_version":1,"members":[`), 0o600); err != nil {
				t.Fatalf("break roster: %v", err)
			}

			w := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody(tt.address)))
			fx.logs.waitFor(t, "auth roster unavailable")

			if tt.generic {
				// The control is a working roster answering the same
				// request, so the comparison is to what the page sees
				// normally, not to a literal this test could get wrong.
				control := newHandlerFixture(t, fixtureOptions{generic: true})
				want := serveOne(control.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody(tt.address)))
				if got := responseOf(w); !reflect.DeepEqual(got, responseOf(want)) {
					t.Errorf("broken roster answered %+v, a working one %+v", got, responseOf(want))
				}
			} else {
				if w.Code != http.StatusInternalServerError || strings.TrimSpace(w.Body.String()) != `{"outcome":"failed"}` {
					t.Errorf("got %d %s, want 500 {\"outcome\":\"failed\"}", w.Code, w.Body.String())
				}
				if strings.Contains(w.Body.String(), "admin@example.com") {
					t.Errorf("body %s names the admin contact", w.Body.String())
				}
			}
			if got := fx.mail.count(); got != 0 {
				t.Errorf("mail sent %d times against a broken roster, want 0", got)
			}
			for _, rec := range fx.logs.records(t) {
				if rec["msg"] == "auth link requested" && rec["known"] == false {
					t.Errorf("a broken roster was logged as an unknown address: %v", rec)
				}
			}
		})
	}
}

// TestRequestLink_ClientFloodSendsNoFurtherMail walks distinct known
// addresses from one client. Every address is new, so only the
// per-client limit can trip, and once it has the mail seam must not be
// called again.
func TestRequestLink_ClientFloodSendsNoFurtherMail(t *testing.T) {
	const perClient = 3
	limits := generousLimits
	limits.LinkRequestsPerClient = perClient
	fx := newHandlerFixture(t, fixtureOptions{limits: limits, maxCalls: perClient})

	addresses := []string{"member@example.com", "member1@example.com", "member2@example.com", "member3@example.com", "admin@example.com"}
	for i, a := range addresses {
		w := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody(a)))
		want := "sent"
		if i >= perClient {
			want = "rate_limited"
		}
		if got := outcomeOf(t, w); got != want {
			t.Errorf("request %d: outcome %q, want %q", i+1, got, want)
		}
	}
	if got := fx.mail.count(); got != perClient {
		t.Errorf("mail sent %d times, want %d", got, perClient)
	}
}

// TestRequestLink_ClientAddressIsTheShells proves the per-client limit
// keys on the address the shell's trusted-proxy policy derives, and
// that the auth event logs that same address. A peer outside the
// allowlist forging a fresh client header on every request is still
// one client; a trusted proxy reporting distinct clients is many.
func TestRequestLink_ClientAddressIsTheShells(t *testing.T) {
	const perClient = 3
	limits := generousLimits
	limits.LinkRequestsPerClient = perClient
	proxy := serve.NewTrustedProxyConfig([]netip.Addr{netip.MustParseAddr("203.0.113.1")}, "X-Forwarded-For")

	tests := []struct {
		name        string
		peer        string
		wantTripAt  int // the 0-based request that trips, or -1
		wantLogged  func(i int) string
		headerValue func(i int) string
	}{
		{
			name:        "untrusted peer forging the header is limited under its peer address",
			peer:        peerIPPort,
			wantTripAt:  perClient,
			headerValue: func(i int) string { return fmt.Sprintf("198.51.100.%d", i+1) },
			wantLogged:  func(int) string { return peerIP },
		},
		{
			name:        "trusted proxy's reported clients are limited apart",
			peer:        "203.0.113.1:40000",
			wantTripAt:  -1,
			headerValue: func(i int) string { return fmt.Sprintf("198.51.100.%d", i+1) },
			wantLogged:  func(i int) string { return fmt.Sprintf("198.51.100.%d", i+1) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newHandlerFixture(t, fixtureOptions{limits: limits, proxy: proxy})
			for i := range perClient + 1 {
				r := post("/login", tt.peer, requestLinkBody(fmt.Sprintf("stranger%d@example.com", i)))
				r.Header.Set("X-Forwarded-For", tt.headerValue(i))
				w := serveOne(fx.handlers.RequestLink, r)

				wantLimited := i == tt.wantTripAt
				if limited := w.Code == http.StatusTooManyRequests; limited != wantLimited {
					t.Errorf("request %d: limited = %v, want %v", i+1, limited, wantLimited)
				}
				rec := fx.logs.records(t)
				if got := rec[len(rec)-1]["client_ip"]; got != tt.wantLogged(i) {
					t.Errorf("request %d: logged client_ip %v, want %s", i+1, got, tt.wantLogged(i))
				}
			}
		})
	}
}

// TestLogsCarryNoCredential runs a link through request and consumption,
// posting the consumption with the token in the query string as a
// mailed link carries it, then presents it again. No log record may
// carry the token or the session cookie's value, and the consumption
// records name the path.
func TestLogsCarryNoCredential(t *testing.T) {
	fx := newHandlerFixture(t, fixtureOptions{})
	serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
	token := tokenFromMail(t, fx.mail.waitCall(t))

	target := "/login/link?" + url.Values{"t": {token}}.Encode()
	w := serveOne(fx.handlers.ConsumeLink, post(target, peerIPPort, consumeBody(token)))
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %v, want the session", cookies)
	}
	serveOne(fx.handlers.ConsumeLink, post(target, peerIPPort, consumeBody(token)))

	logs := fx.logs.text()
	if strings.Contains(logs, token) {
		t.Errorf("a log record carries the token:\n%s", logs)
	}
	if strings.Contains(logs, cookies[0].Value) {
		t.Errorf("a log record carries the session cookie value:\n%s", logs)
	}
	for _, msg := range []string{"auth link consumed", "auth link consumption failed"} {
		if got := fx.logs.waitFor(t, msg)["path"]; got != "/login/link" {
			t.Errorf("%q path = %v, want /login/link", msg, got)
		}
	}
}

// TestSignOut_ClearsCookie compares sign-out's Set-Cookie against the
// header written out by hand. RFC 6265 section 5.2.2 treats a Max-Age of
// zero or less as "expire now", and net/http renders a negative MaxAge
// as Max-Age=0, which is the form on the wire.
func TestSignOut_ClearsCookie(t *testing.T) {
	fx := newHandlerFixture(t, fixtureOptions{})
	w := serveOne(fx.handlers.SignOut, post("/logout", peerIPPort, ""))
	want := []string{"__Host-dubplate_session=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Lax"}
	if got := w.Header().Values("Set-Cookie"); !reflect.DeepEqual(got, want) {
		t.Errorf("Set-Cookie = %q, want %q", got, want)
	}
	if w.Code != http.StatusOK || outcomeOf(t, w) != "signed_out" {
		t.Errorf("response = %d %s, want 200 signed_out", w.Code, w.Body.String())
	}
}

// floodPeer is the RemoteAddr of the i'th flooding client, each in a
// /64 of its own, so no per-client limit can refuse a flood request
// and the per-client order cannot explain a row's result.
func floodPeer(i int) string {
	return fmt.Sprintf("[2001:db8:%x:%x::1]:40000", i>>16, i&0xffff)
}

// TestRequestLink_UnknownFloodLeavesSpentAddress spends a member's
// per-address budget, then floods as many made-up addresses as the
// limiter holds keys, each from a /64 of its own and all at the same
// instant. The member is still refused: no refill can explain a pass,
// and the flood never charged the per-address map, so it evicted
// nothing from it.
func TestRequestLink_UnknownFloodLeavesSpentAddress(t *testing.T) {
	const perAddress = 2
	limits := generousLimits
	limits.LinkRequestsPerAddress = perAddress
	fx := newHandlerFixture(t, fixtureOptions{limits: limits, maxCalls: perAddress})

	for i := range perAddress {
		w := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
		if got := outcomeOf(t, w); got != "sent" {
			t.Fatalf("victim request %d: outcome %q, want sent", i+1, got)
		}
	}
	before := fx.limiter.perAddress.size()
	for i := range maxLimiterKeys {
		serveOne(fx.handlers.RequestLink, post("/login", floodPeer(i), requestLinkBody(fmt.Sprintf("stranger%d@example.com", i))))
	}
	if after := fx.limiter.perAddress.size(); after != before {
		t.Errorf("per-address map held %d keys before the flood and %d after, want no change", before, after)
	}

	w := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
	if w.Code != http.StatusTooManyRequests || outcomeOf(t, w) != "rate_limited" {
		t.Errorf("victim after the flood: got %d %s, want 429 rate_limited", w.Code, w.Body.String())
	}
	if got := fx.mail.count(); got != perAddress {
		t.Errorf("mail sent %d times, want %d", got, perAddress)
	}
}

// TestRequestLink_ChargingOrder walks each state the request-link
// handler's limits meet. The client is charged first, for every
// request, so a refused client never reaches the address map and an
// address that fails to normalize still spends the client's token; the
// address is charged only once the roster knows it. Each row runs in a
// synctest bubble, so the generic answer's send, which outlives the
// response, has finished or never started when the mail is counted.
func TestRequestLink_ChargingOrder(t *testing.T) {
	type linkRequest struct{ address, peer string }
	oneClient := generousLimits
	oneClient.LinkRequestsPerClient = 1
	oneAddress := generousLimits
	oneAddress.LinkRequestsPerAddress = 1

	tests := []struct {
		name            string
		generic         bool
		limits          Limits
		before          []linkRequest
		final           linkRequest
		wantStatus      int
		wantOutcome     string
		wantAddressKeys int
		wantMail        int
	}{
		{
			name:            "a refused client never reaches the address map",
			limits:          oneClient,
			before:          []linkRequest{{"member1@example.com", peerIPPort}},
			final:           linkRequest{"member@example.com", peerIPPort},
			wantStatus:      http.StatusTooManyRequests,
			wantOutcome:     "rate_limited",
			wantAddressKeys: 1,
			wantMail:        1,
		},
		{
			name:            "an address that fails to normalize spends the client's token",
			limits:          oneClient,
			before:          []linkRequest{{"not an address", peerIPPort}},
			final:           linkRequest{"member@example.com", peerIPPort},
			wantStatus:      http.StatusTooManyRequests,
			wantOutcome:     "rate_limited",
			wantAddressKeys: 0,
			wantMail:        0,
		},
		{
			name:   "spellings of one address share its bucket",
			limits: oneAddress,
			before: []linkRequest{{"<MEMBER@Example.COM>", floodPeer(1)}},
			final:  linkRequest{"member@example.com", peerIPPort},
			// The first spelling was mailed; the second is refused.
			wantStatus:      http.StatusTooManyRequests,
			wantOutcome:     "rate_limited",
			wantAddressKeys: 1,
			wantMail:        1,
		},
		{
			name:        "kind answer, unknown address leaves the address map alone",
			final:       linkRequest{"stranger@example.com", peerIPPort},
			wantStatus:  http.StatusOK,
			wantOutcome: "unknown",
		},
		{
			name:        "generic answer, unknown address leaves the address map alone",
			generic:     true,
			final:       linkRequest{"stranger@example.com", peerIPPort},
			wantStatus:  http.StatusOK,
			wantOutcome: "requested",
		},
		{
			name:            "kind answer, known address over its limit is refused",
			limits:          oneAddress,
			before:          []linkRequest{{"member@example.com", floodPeer(1)}},
			final:           linkRequest{"member@example.com", peerIPPort},
			wantStatus:      http.StatusTooManyRequests,
			wantOutcome:     "rate_limited",
			wantAddressKeys: 1,
			wantMail:        1,
		},
		{
			name:            "generic answer, known address over its limit is told requested and not mailed",
			generic:         true,
			limits:          oneAddress,
			before:          []linkRequest{{"member@example.com", floodPeer(1)}},
			final:           linkRequest{"member@example.com", peerIPPort},
			wantStatus:      http.StatusOK,
			wantOutcome:     "requested",
			wantAddressKeys: 1,
			wantMail:        1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				fx := newHandlerFixture(t, fixtureOptions{generic: tt.generic, limits: tt.limits})
				for _, req := range tt.before {
					serveOne(fx.handlers.RequestLink, post("/login", req.peer, requestLinkBody(req.address)))
				}
				w := serveOne(fx.handlers.RequestLink, post("/login", tt.final.peer, requestLinkBody(tt.final.address)))
				synctest.Wait()

				if w.Code != tt.wantStatus || outcomeOf(t, w) != tt.wantOutcome {
					t.Errorf("got %d %s, want %d %s", w.Code, w.Body.String(), tt.wantStatus, tt.wantOutcome)
				}
				if got := fx.limiter.perAddress.size(); got != tt.wantAddressKeys {
					t.Errorf("per-address map holds %d keys, want %d", got, tt.wantAddressKeys)
				}
				if got := fx.mail.count(); got != tt.wantMail {
					t.Errorf("mail sent %d times, want %d", got, tt.wantMail)
				}
			})
		})
	}
}

// TestRequestLink_GenericOverLimitAnswersAsUnknown compares the generic
// answer to a known address over its per-address limit against its
// answer to an unknown address, whole response against whole response,
// and requires that the over-limit address is not mailed. A 429 here
// would tell a stranger the address is on the roster.
func TestRequestLink_GenericOverLimitAnswersAsUnknown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		limits := generousLimits
		limits.LinkRequestsPerAddress = 1
		fx := newHandlerFixture(t, fixtureOptions{generic: true, limits: limits})

		serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
		synctest.Wait()
		if got := fx.mail.count(); got != 1 {
			t.Fatalf("first request mailed %d times, want 1", got)
		}

		overLimit := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
		unknown := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("stranger@example.com")))
		synctest.Wait()

		if got, want := responseOf(overLimit), responseOf(unknown); !reflect.DeepEqual(got, want) {
			t.Errorf("over-limit known address answered %+v, unknown answered %+v", got, want)
		}
		if got := fx.mail.count(); got != 1 {
			t.Errorf("mail sent %d times, want 1 (the over-limit address is not mailed)", got)
		}
	})
}

// TestRequestLink_ClientChargedBeforeBodyParses holds n1: the
// per-client limit is charged before the body is even read, so a body
// that fails to parse still costs the client a request, matching
// consumeLink's own AllowConsumption-before-readForm order. The body
// carries an invalid percent-escape ("%zz"), which r.ParseForm
// refuses, so readForm answers malformed.
func TestRequestLink_ClientChargedBeforeBodyParses(t *testing.T) {
	limits := generousLimits
	limits.LinkRequestsPerClient = 1
	fx := newHandlerFixture(t, fixtureOptions{limits: limits})

	// "%zz" is not a valid percent-escape, which url.ParseQuery (and so
	// r.ParseForm) refuses.
	bad := post("/login", peerIPPort, "address=%zz")

	first := serveOne(fx.handlers.RequestLink, bad)
	if got := outcomeOf(t, first); got != outcomeMalformed {
		t.Fatalf("first request outcome = %q, want %q", got, outcomeMalformed)
	}

	second := serveOne(fx.handlers.RequestLink, post("/login", peerIPPort, requestLinkBody("member@example.com")))
	if got := outcomeOf(t, second); got != outcomeRateLimited {
		t.Errorf("second request outcome = %q, want %q: the malformed body should already have charged the client's one token", got, outcomeRateLimited)
	}
}

// TestReadForm proves the shared body path enforces the cap its caller
// passes, not a value of its own: a body of exactly the cap parses, one
// byte more is refused as too large, and a body that does not decode is
// refused as malformed. Each refusal answers the request itself.
func TestReadForm(t *testing.T) {
	const bodyCap = 64
	atCap := "a=" + strings.Repeat("x", bodyCap-len("a="))
	tests := []struct {
		name        string
		body        string
		wantOK      bool
		wantStatus  int
		wantOutcome string
	}{
		{name: "exactly the cap", body: atCap, wantOK: true},
		{name: "one byte over the cap", body: atCap + "x", wantStatus: http.StatusRequestEntityTooLarge, wantOutcome: outcomeTooLarge},
		{name: "an escape that does not decode", body: "a=%zz", wantStatus: http.StatusBadRequest, wantOutcome: outcomeMalformed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := post("/", peerIPPort, tt.body)
			ok := readForm(w, r, bodyCap, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if ok != tt.wantOK {
				t.Fatalf("readForm() = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK {
				if got := r.PostForm.Get("a"); got != atCap[len("a="):] {
					t.Errorf("PostForm a = %q, want the body's value", got)
				}
				return
			}
			if w.Code != tt.wantStatus || outcomeOf(t, w) != tt.wantOutcome {
				t.Errorf("readForm answered %d %s, want %d %s", w.Code, w.Body, tt.wantStatus, tt.wantOutcome)
			}
		})
	}
}
