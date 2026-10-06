package auth

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/serve"
)

// adminRosterMembers holds an admin with a username, a second
// admin without one (a record written before the write path required
// it), and a non-admin, so a GET has every field shape to answer and a
// hand edit demoting the first admin still leaves the roster an admin.
const adminRosterMembers = `[
	{"address":"admin@example.com","display_name":"Group Admin","username":"groupadmin","admin":true},
	{"address":"founder@example.com","display_name":"Founder","admin":true},
	{"address":"member@example.com","display_name":"Member","username":"member","admin":false}
]`

// fixtureAddresses is every address a fixture roster or session names;
// no refusal's log record may carry any of them.
var fixtureAddresses = []string{"admin@example.com", "founder@example.com", "member@example.com", "gone@example.com"}

// gateNow is ten days after seamBirth, when every session minted at
// seamBirth still verifies and a refresh visibly moves its expiry: from
// seamBirth plus the 30-day lifetime to gateNow plus the same.
var gateNow = seamBirth.Add(10 * 24 * time.Hour)

// validPost is a form every field check accepts for a new member.
const validPost = "address=new%40example.com&display_name=New+Member&username=newmember"

const formType = "application/x-www-form-urlencoded"

type adminFixture struct {
	rosterPath string
	roster     *Roster
	epoch      *EpochStore
	seam       *MagicLinkSeam
	logs       *logCapture
	cfg        AdminConfig
	api        http.Handler
}

func newAdminFixture(t *testing.T) adminFixture {
	t.Helper()
	sf := newSeamFixture(t, adminRosterMembers)
	logs := newLogCapture()
	cfg := AdminConfig{
		Seam:     sf.seam,
		Roster:   sf.roster,
		ClientIP: func(r *http.Request) string { return serve.ClientIP(r, serve.TrustedProxyConfig{}) },
		Now:      func() time.Time { return gateNow },
		Logger:   slog.New(slog.NewJSONHandler(logs, nil)),
		BodyCap:  4096,
	}
	return adminFixture{
		rosterPath: sf.roster.path,
		roster:     sf.roster,
		epoch:      sf.epoch,
		seam:       sf.seam,
		logs:       logs,
		cfg:        cfg,
		api:        NewRosterAPI(cfg),
	}
}

// mintAt returns a session value for address signed under secret,
// minted at issued.
func mintAt(t *testing.T, secret, address string, issued time.Time) string {
	t.Helper()
	value, _, err := MintSession([]byte(secret), address, issued, seamLifetime, seamCap)
	if err != nil {
		t.Fatalf("MintSession(%s): %v", address, err)
	}
	return value
}

// rosterRequest builds a request to the roster API carrying session as
// its cookie when non-empty and a urlencoded body when body is
// non-empty.
func rosterRequest(method, target, session, body string) *http.Request {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", formType)
	}
	if session != "" {
		r.AddCookie(&http.Cookie{Name: SessionCookieName, Value: session})
	}
	return r
}

// notFoundFor is what http.NotFound answers r with, the terminator's
// own answer to an unregistered path under the prefix.
func notFoundFor(r *http.Request) response {
	w := httptest.NewRecorder()
	http.NotFound(w, r)
	return responseOf(w)
}

// withAdmin returns r carrying v the way the gate hands it on.
func withAdmin(r *http.Request, v VerifiedIdentity) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), adminKey{}, v))
}

// fixtureAdmin is the verified identity the seam yields for the
// fixture's first admin.
var fixtureAdmin = VerifiedIdentity{
	Identity: identity.Identity{
		Address:     "admin@example.com",
		Principal:   "groupadmin",
		DisplayName: "Group Admin",
		Source:      magicLinkSource,
		Channel:     identity.ChannelCookie,
	},
	Admin: true,
}

// stubSeam answers every request with v. It stands in for a verifier
// that emits an identity the magic-link seam never does.
type stubSeam struct{ v VerifiedIdentity }

func (s stubSeam) VerifyIdentity(*http.Request, time.Time, bool) (VerifiedIdentity, error) {
	return s.v, nil
}

func (s stubSeam) Read(string) (time.Time, bool, error) {
	return time.Time{}, false, errors.New("stub seam: Read is not part of the gate")
}

// recordingHandler notes whether the gate admitted a request.
type recordingHandler struct{ reached bool }

func (h *recordingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.reached = true
	w.WriteHeader(http.StatusTeapot)
}

// TestRequireAdmin_RefusalsMatchNotFound proves every refusal answers
// http.NotFound's own bytes, carries no cookie, and never reaches the
// wrapped handler. Comparing refusals only with each other would pass a
// gate answering its own JSON 404.
//
// The non-admin rows are same-origin and old enough that the seam
// refreshes them, and they carry an admin form value and an X-Admin
// header, so a gate that set the cookie before the admin check, or
// read the flag from the request, fails them.
func TestRequireAdmin_RefusalsMatchNotFound(t *testing.T) {
	commandLineAdmin := fixtureAdmin
	commandLineAdmin.Identity.Channel = identity.ChannelCommandLine
	commandLineAdmin.RefreshedValue = "refreshed-by-stub"

	tests := []struct {
		name    string
		seam    Seam
		request func(t *testing.T, fx adminFixture) *http.Request
	}{
		{
			name: "no cookie",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				return rosterRequest(http.MethodGet, "/dubplate/roster", "", "")
			},
		},
		{
			name: "forged cookie",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, foreignFixture, "admin@example.com", seamBirth), "")
			},
		},
		{
			name: "expired session",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", gateNow.Add(-seamLifetime-time.Hour)), "")
			},
		},
		{
			name: "session older than its epoch",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				if err := fx.epoch.Bump("admin@example.com", seamBirth.Add(time.Hour)); err != nil {
					t.Fatalf("Bump: %v", err)
				}
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), "")
			},
		},
		{
			name: "member removed from the roster",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "gone@example.com", seamBirth), "")
			},
		},
		{
			name: "roster store fails on the lookup",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				replaceRosterFile(t, fx.rosterPath, "not json")
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), "")
			},
		},
		{
			name: "verified non-admin GET with an admin claim",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				r := rosterRequest(http.MethodGet, "/dubplate/roster?admin=on", nonAdminRefreshing(t, fx), "")
				r.Header.Set("Sec-Fetch-Site", "same-origin")
				r.Header.Set("X-Admin", "1")
				return r
			},
		},
		{
			name: "verified non-admin POST with an admin claim",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				r := rosterRequest(http.MethodPost, "/dubplate/roster", nonAdminRefreshing(t, fx), validPost+"&admin=on")
				r.Header.Set("Sec-Fetch-Site", "same-origin")
				r.Header.Set("X-Admin", "1")
				return r
			},
		},
		{
			name: "verified non-admin PUT",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				r := rosterRequest(http.MethodPut, "/dubplate/roster", nonAdminRefreshing(t, fx), "")
				r.Header.Set("Sec-Fetch-Site", "same-origin")
				return r
			},
		},
		{
			name: "command-line identity holding the admin flag",
			seam: stubSeam{v: commandLineAdmin},
			request: func(t *testing.T, fx adminFixture) *http.Request {
				r := rosterRequest(http.MethodGet, "/dubplate/roster", "", "")
				r.Header.Set("Sec-Fetch-Site", "same-origin")
				return r
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newAdminFixture(t)
			cfg := fx.cfg
			if tt.seam != nil {
				cfg.Seam = tt.seam
			}
			r := tt.request(t, fx)
			before, err := os.ReadFile(fx.rosterPath)
			if err != nil {
				t.Fatalf("read roster: %v", err)
			}
			var cookieValue string
			if c, err := r.Cookie(SessionCookieName); err == nil {
				cookieValue = c.Value
			}
			want := notFoundFor(r.Clone(t.Context()))

			handlers := map[string]http.Handler{"roster API": NewRosterAPI(cfg)}
			next := &recordingHandler{}
			handlers["gate"] = RequireAdmin(cfg, next)
			for name, h := range handlers {
				got := responseOf(serveOne(h, r.Clone(t.Context())))
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s answered %+v, want http.NotFound's %+v", name, got, want)
				}
				if c := got.header.Values("Set-Cookie"); len(c) > 0 {
					t.Errorf("%s refusal set a cookie: %q", name, c)
				}
			}
			if next.reached {
				t.Error("the gate admitted a refused request to the wrapped handler")
			}

			after, err := os.ReadFile(fx.rosterPath)
			if err != nil {
				t.Fatalf("read roster after: %v", err)
			}
			if !bytes.Equal(before, after) {
				t.Errorf("a refused request changed the roster:\nbefore %s\nafter  %s", before, after)
			}
			assertRefusalLogs(t, fx.logs, 2)
			if cookieValue != "" && strings.Contains(fx.logs.text(), cookieValue) {
				t.Errorf("a refusal log carries the cookie value:\n%s", fx.logs.text())
			}
		})
	}
}

// nonAdminRefreshing returns a session for the fixture's non-admin,
// first proving the seam refreshes it on a same-origin request and
// reads it as no admin, so a row using it cannot pass because nothing
// was refreshed.
func nonAdminRefreshing(t *testing.T, fx adminFixture) string {
	t.Helper()
	value := mintAt(t, currentFixture, "member@example.com", seamBirth)
	v, err := fx.seam.VerifyIdentity(requestWithSession(t, value), gateNow, true)
	if err != nil {
		t.Fatalf("VerifyIdentity(non-admin): %v", err)
	}
	if v.RefreshedValue == "" || v.Admin {
		t.Fatalf("non-admin session: refreshed %q, admin %v; want a refresh and no admin flag", v.RefreshedValue, v.Admin)
	}
	return value
}

// assertRefusalLogs fails when fewer than minRecords records were
// logged, or when any record carries a fixture address, so "carries
// none" is read over real lines.
func assertRefusalLogs(t *testing.T, logs *logCapture, minRecords int) {
	t.Helper()
	if n := len(logs.records(t)); n < minRecords {
		t.Errorf("log holds %d records, want at least %d:\n%s", n, minRecords, logs.text())
	}
	for _, addr := range fixtureAddresses {
		if strings.Contains(logs.text(), addr) {
			t.Errorf("a refusal log carries the address %s:\n%s", addr, logs.text())
		}
	}
}

// TestRosterAPI_InnerRefusesAlone proves the handler the gate wraps
// answers http.NotFound's bytes when its context carries no valid
// admin, so a mount that skipped the gate discloses nothing.
func TestRosterAPI_InnerRefusesAlone(t *testing.T) {
	notAdmin := fixtureAdmin
	notAdmin.Admin = false
	commandLine := fixtureAdmin
	commandLine.Identity.Channel = identity.ChannelCommandLine

	tests := []struct {
		name string
		ctx  *VerifiedIdentity
	}{
		{name: "no identity"},
		{name: "valid identity without the admin flag", ctx: &notAdmin},
		{name: "command-line identity with the admin flag", ctx: &commandLine},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newAdminFixture(t)
			inner := &rosterAPI{cfg: fx.cfg, writeWait: time.Second}
			r := rosterRequest(http.MethodGet, "/dubplate/roster", "", "")
			if tt.ctx != nil {
				r = withAdmin(r, *tt.ctx)
			}
			want := notFoundFor(r)
			if got := responseOf(serveOne(inner, r)); !reflect.DeepEqual(got, want) {
				t.Errorf("inner handler answered %+v, want http.NotFound's %+v", got, want)
			}
		})
	}
}

// TestRosterAPI_List proves a verified admin's GET answers the roster
// in file order with exactly the four fields per member, HEAD answers
// the same header without a body, and any other method answers 405.
func TestRosterAPI_List(t *testing.T) {
	fx := newAdminFixture(t)
	session := mintAt(t, currentFixture, "admin@example.com", seamBirth)

	w := serveOne(fx.api, rosterRequest(http.MethodGet, "/dubplate/roster", session, ""))
	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body %s", w.Code, w.Body)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("GET Cache-Control = %q, want no-store", got)
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode GET body %q: %v", w.Body, err)
	}
	want := map[string]any{"members": []any{
		map[string]any{"address": "admin@example.com", "display_name": "Group Admin", "username": "groupadmin", "admin": true},
		map[string]any{"address": "founder@example.com", "display_name": "Founder", "username": "", "admin": true},
		map[string]any{"address": "member@example.com", "display_name": "Member", "username": "member", "admin": false},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GET body = %v, want %v", got, want)
	}

	head := serveOne(fx.api, rosterRequest(http.MethodHead, "/dubplate/roster", session, ""))
	if head.Code != w.Code || !reflect.DeepEqual(head.Header(), w.Header()) {
		t.Errorf("HEAD = %d %v, want GET's %d %v", head.Code, head.Header(), w.Code, w.Header())
	}
	if head.Body.Len() != 0 {
		t.Errorf("HEAD body = %q, want none", head.Body)
	}

	put := serveOne(fx.api, rosterRequest(http.MethodPut, "/dubplate/roster", session, ""))
	if put.Code != http.StatusMethodNotAllowed {
		t.Errorf("admin PUT status = %d, want 405", put.Code)
	}
	if got := put.Header().Get("Allow"); got != "GET, HEAD, POST" {
		t.Errorf("admin PUT Allow = %q, want %q", got, "GET, HEAD, POST")
	}
}

// TestRosterAPI_ListStoreFailure proves a roster that fails to load
// answers failed and no member list. It runs on the inner handler:
// through the gate the seam's own lookup reads the same broken file and
// refuses first.
func TestRosterAPI_ListStoreFailure(t *testing.T) {
	fx := newAdminFixture(t)
	replaceRosterFile(t, fx.rosterPath, "not json")
	inner := &rosterAPI{cfg: fx.cfg, writeWait: time.Second}

	w := serveOne(inner, withAdmin(rosterRequest(http.MethodGet, "/dubplate/roster", "", ""), fixtureAdmin))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body %q: %v", w.Body, err)
	}
	if got["outcome"] != outcomeFailed {
		t.Errorf("outcome = %v, want failed", got["outcome"])
	}
	if _, ok := got["members"]; ok {
		t.Errorf("a failed listing carries members: %v", got)
	}
}

// TestRequireAdmin_RefreshOnlySameOrigin proves the gate refreshes an
// admin's session only for Sec-Fetch-Site: same-origin, moving the
// expiry and holding the anchor. The session is ten days old, so a
// refresh moves its expiry by ten days.
func TestRequireAdmin_RefreshOnlySameOrigin(t *testing.T) {
	tests := []struct {
		site        string
		wantRefresh bool
	}{
		{site: "same-origin", wantRefresh: true},
		{site: "same-site"},
		{site: "cross-site"},
		{site: "none"},
		{site: ""},
	}
	for _, tt := range tests {
		t.Run("Sec-Fetch-Site "+tt.site, func(t *testing.T) {
			fx := newAdminFixture(t)
			r := rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), "")
			if tt.site != "" {
				r.Header.Set("Sec-Fetch-Site", tt.site)
			}
			w := serveOne(fx.api, r)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", w.Code)
			}
			cookies := w.Header().Values("Set-Cookie")
			if !tt.wantRefresh {
				if len(cookies) > 0 {
					t.Errorf("a %q request refreshed the session: %q", tt.site, cookies)
				}
				return
			}
			if len(cookies) != 1 {
				t.Fatalf("Set-Cookie = %q, want one refreshed session", cookies)
			}
			c, err := http.ParseSetCookie(cookies[0])
			if err != nil {
				t.Fatalf("ParseSetCookie: %v", err)
			}
			sess, err := VerifySession([]byte(currentFixture), nil, c.Value, gateNow, seamCap)
			if err != nil {
				t.Fatalf("VerifySession(refreshed): %v", err)
			}
			if want := seamBirth.Add(40 * 24 * time.Hour); !sess.ExpiresAt.Equal(want) {
				t.Errorf("refreshed expiry = %v, want %v", sess.ExpiresAt, want)
			}
			if !sess.AnchoredAt.Equal(seamBirth) {
				t.Errorf("refreshed anchor = %v, want %v", sess.AnchoredAt, seamBirth)
			}
		})
	}
}

// countingReader counts the bytes read through it.
type countingReader struct {
	r io.Reader
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

// TestRosterAPI_Post proves every POST outcome through the gate as a
// verified admin. Every refused row leaves the roster's bytes as they
// were and logs no audit record; the added rows log exactly one, naming
// the gate's actor on the cookie channel.
func TestRosterAPI_Post(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		contentType string // "-" sends no Content-Type
		body        string
		wantStatus  int
		wantOutcome string
	}{
		{name: "added", body: validPost, wantStatus: 201, wantOutcome: "added"},
		{name: "added under a charset parameter", contentType: formType + "; charset=utf-8", body: validPost, wantStatus: 201, wantOutcome: "added"},
		{name: "added under Fetch's own serialization", contentType: formType + ";charset=UTF-8", body: validPost, wantStatus: 201, wantOutcome: "added"},

		{name: "duplicate address", body: "address=MEMBER%40example.com&display_name=Dup&username=dup", wantStatus: 409, wantOutcome: "duplicate_address"},
		{name: "duplicate username", body: "address=new%40example.com&display_name=New&username=GroupAdmin", wantStatus: 409, wantOutcome: "duplicate_username"},
		{name: "duplicate principal", body: "address=new%40example.com&display_name=New&username=founder%40example.com", wantStatus: 409, wantOutcome: "duplicate_principal"},
		{name: "invalid address net/mail quotes", body: "address=new%40example.com+member%40example.com&display_name=New&username=newmember", wantStatus: 400, wantOutcome: "invalid_address"},
		{name: "address carrying an invisible rune", body: "address=%22new%E2%80%8B%22%40example.com&display_name=New&username=newmember", wantStatus: 400, wantOutcome: "invalid_address"},
		{name: "invalid display name", body: "address=new%40example.com&display_name=New%E2%80%AE&username=newmember", wantStatus: 400, wantOutcome: "invalid_display_name"},
		{name: "blank display name", body: "address=new%40example.com&display_name=+++&username=newmember", wantStatus: 400, wantOutcome: "blank_display_name"},
		{name: "absent display name", body: "address=new%40example.com&username=newmember", wantStatus: 400, wantOutcome: "blank_display_name"},
		{name: "invalid username", body: "address=new%40example.com&display_name=New&username=-new", wantStatus: 400, wantOutcome: "invalid_username"},
		{name: "missing username", body: "address=new%40example.com&display_name=New", wantStatus: 400, wantOutcome: "missing_username"},

		{name: "admin=on in the body", body: validPost + "&admin=on", wantStatus: 400, wantOutcome: "admin_not_grantable"},
		{name: "admin=false in the body", body: validPost + "&admin=false", wantStatus: 400, wantOutcome: "admin_not_grantable"},
		{name: "admin= in the body", body: validPost + "&admin=", wantStatus: 400, wantOutcome: "admin_not_grantable"},
		{name: "admin= in the query", target: "/dubplate/roster?admin=", body: validPost, wantStatus: 400, wantOutcome: "admin_not_grantable"},
		{name: "admin=on in the query", target: "/dubplate/roster?admin=on", body: validPost, wantStatus: 400, wantOutcome: "admin_not_grantable"},

		{name: "Admin=on", body: validPost + "&Admin=on", wantStatus: 400, wantOutcome: "unknown_field"},
		{name: "is_admin=1", body: validPost + "&is_admin=1", wantStatus: 400, wantOutcome: "unknown_field"},
		{name: "streamer_username=x", body: validPost + "&streamer_username=x", wantStatus: 400, wantOutcome: "unknown_field"},
		{name: "actor=x", body: validPost + "&actor=x", wantStatus: 400, wantOutcome: "unknown_field"},
		{name: "foo=1 in the query", target: "/dubplate/roster?foo=1", body: validPost, wantStatus: 400, wantOutcome: "unknown_field"},

		{name: "repeated address", body: validPost + "&address=other%40example.com", wantStatus: 400, wantOutcome: "malformed"},
		{name: "a field in the query and the body", target: "/dubplate/roster?display_name=Other", body: validPost, wantStatus: 400, wantOutcome: "malformed"},
		{name: "multipart media type", contentType: "multipart/form-data; boundary=x", body: validPost, wantStatus: 400, wantOutcome: "malformed"},
		{name: "no Content-Type", contentType: "-", body: validPost, wantStatus: 400, wantOutcome: "malformed"},
		{name: "unparseable Content-Type", contentType: formType + "; ;", body: validPost, wantStatus: 400, wantOutcome: "malformed"},
		{name: "unparseable body", body: validPost + "&%zz", wantStatus: 400, wantOutcome: "malformed"},
		{name: "fields in the query, empty body", target: "/dubplate/roster?address=new%40example.com&display_name=x&username=y", body: "", wantStatus: 400, wantOutcome: "malformed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newAdminFixture(t)
			before, err := os.ReadFile(fx.rosterPath)
			if err != nil {
				t.Fatalf("read roster: %v", err)
			}
			target := cmp.Or(tt.target, "/dubplate/roster")
			r := rosterRequest(http.MethodPost, target, mintAt(t, currentFixture, "Admin@Example.com", seamBirth), tt.body)
			switch tt.contentType {
			case "":
				r.Header.Set("Content-Type", formType)
			case "-":
				r.Header.Del("Content-Type")
			default:
				r.Header.Set("Content-Type", tt.contentType)
			}

			w := serveOne(fx.api, r)
			if w.Code != tt.wantStatus || outcomeOf(t, w) != tt.wantOutcome {
				t.Fatalf("POST = %d %s, want %d %s", w.Code, w.Body, tt.wantStatus, tt.wantOutcome)
			}
			audits := auditRecords(t, fx.logs)
			if tt.wantOutcome != outcomeAdded {
				assertUnchanged(t, fx.rosterPath, before)
				if len(audits) != 0 {
					t.Errorf("a refused POST logged an audit record: %v", audits)
				}
				assertRefusalLogs(t, fx.logs, 1)
				return
			}

			if len(audits) != 1 {
				t.Fatalf("audit records = %v, want exactly one", audits)
			}
			for key, want := range map[string]any{"actor": "admin@example.com", "channel": "cookie", "address": "new@example.com", "admin": false} {
				if got := audits[0][key]; got != want {
					t.Errorf("audit record[%q] = %v, want %v", key, got, want)
				}
			}
			if _, err := fx.roster.Lookup("new@example.com"); err != nil {
				t.Errorf("Lookup(new member) after added: %v", err)
			}
		})
	}
}

// auditRecords returns every roster audit record the log holds.
func auditRecords(t *testing.T, logs *logCapture) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, rec := range logs.records(t) {
		if msg, _ := rec["msg"].(string); strings.HasPrefix(msg, "roster member added") {
			out = append(out, rec)
		}
	}
	return out
}

func assertUnchanged(t *testing.T, path string, before []byte) {
	t.Helper()
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read roster after: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("the roster changed:\nbefore %s\nafter  %s", before, after)
	}
}

// TestRosterAPI_PostOverCap proves the body cap is installed before the
// form parses: an oversized body is refused having been read no further
// than one byte past the cap.
func TestRosterAPI_PostOverCap(t *testing.T) {
	fx := newAdminFixture(t)
	before, err := os.ReadFile(fx.rosterPath)
	if err != nil {
		t.Fatalf("read roster: %v", err)
	}
	const bodyCap = 512
	fx.cfg.BodyCap = bodyCap
	api := NewRosterAPI(fx.cfg)

	body := &countingReader{r: strings.NewReader(validPost + "&pad=" + strings.Repeat("x", 64<<10))}
	r := rosterRequest(http.MethodPost, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), "x")
	r.Body = io.NopCloser(body)
	r.ContentLength = -1

	w := serveOne(api, r)
	if w.Code != http.StatusRequestEntityTooLarge || outcomeOf(t, w) != outcomeTooLarge {
		t.Errorf("POST = %d %s, want 413 too_large", w.Code, w.Body)
	}
	if body.n > bodyCap+1 {
		t.Errorf("read %d bytes of an oversized body, want at most %d", body.n, bodyCap+1)
	}
	assertUnchanged(t, fx.rosterPath, before)
}

// TestRosterAPI_PostFailed proves the refusals a verified admin can
// still meet at the store answer 500 failed, change no byte, and log no
// audit record. They run on the inner handler, since each changes the
// roster after the identity was verified.
func TestRosterAPI_PostFailed(t *testing.T) {
	demoted := strings.Replace(adminRosterMembers, `"username":"groupadmin","admin":true`, `"username":"groupadmin","admin":false`, 1)
	tests := []struct {
		name   string
		change func(t *testing.T, path string)
	}{
		{
			name: "actor demoted by a hand edit",
			change: func(t *testing.T, path string) {
				replaceRosterFile(t, path, `{"schema_version":1,"members":`+demoted+`}`)
				// The edited roster still loads, so the refusal is the
				// actor check under the lock, not the load's no-admin
				// refusal.
				if _, err := OpenRoster(path); err != nil {
					t.Fatalf("OpenRoster(edited): %v", err)
				}
			},
		},
		{
			name: "roster writable by others",
			change: func(t *testing.T, path string) {
				if err := os.Chmod(path, 0o666); err != nil {
					t.Fatalf("chmod: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newAdminFixture(t)
			tt.change(t, fx.rosterPath)
			before, err := os.ReadFile(fx.rosterPath)
			if err != nil {
				t.Fatalf("read roster: %v", err)
			}
			inner := &rosterAPI{cfg: fx.cfg, writeWait: time.Second}

			w := serveOne(inner, withAdmin(rosterRequest(http.MethodPost, "/dubplate/roster", "", validPost), fixtureAdmin))
			if w.Code != http.StatusInternalServerError || outcomeOf(t, w) != outcomeFailed {
				t.Errorf("POST = %d %s, want 500 failed", w.Code, w.Body)
			}
			assertUnchanged(t, fx.rosterPath, before)
			if audits := auditRecords(t, fx.logs); len(audits) != 0 {
				t.Errorf("a refused POST logged an audit record: %v", audits)
			}
		})
	}
}

// TestRosterAPI_PostWriteUnconfirmed proves a write that fails after
// its rename answers write_unconfirmed, with the member on disk.
func TestRosterAPI_PostWriteUnconfirmed(t *testing.T) {
	fx := newAdminFixture(t)
	t.Cleanup(func() { writeRosterStream = atomicfile.WriteStream })
	writeRosterStream = func(path string, perm os.FileMode, write func(io.Writer) error) error {
		if err := atomicfile.WriteStream(path, perm, write); err != nil {
			return err
		}
		return errors.New("injected failure after the rename")
	}

	w := serveOne(fx.api, rosterRequest(http.MethodPost, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), validPost))
	if w.Code != http.StatusInternalServerError || outcomeOf(t, w) != outcomeWriteUnconfirmed {
		t.Errorf("POST = %d %s, want 500 write_unconfirmed", w.Code, w.Body)
	}
	reopened, err := OpenRoster(fx.rosterPath)
	if err != nil {
		t.Fatalf("OpenRoster: %v", err)
	}
	if _, err := reopened.Lookup("new@example.com"); err != nil {
		t.Errorf("Lookup(new member) after write_unconfirmed: %v; want the member on disk", err)
	}
}

// TestRosterAPI_PostBusy holds the roster directory's lock as another
// process would, and proves the write gives up with 503 busy once its
// own wait ends, well before the request would otherwise.
func TestRosterAPI_PostBusy(t *testing.T) {
	fx := newAdminFixture(t)
	before, err := os.ReadFile(fx.rosterPath)
	if err != nil {
		t.Fatalf("read roster: %v", err)
	}
	dir, err := os.Open(filepath.Dir(fx.rosterPath))
	if err != nil {
		t.Fatalf("open roster directory: %v", err)
	}
	defer func() { _ = dir.Close() }()
	if err := syscall.Flock(int(dir.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock roster directory: %v", err)
	}

	const wait = 100 * time.Millisecond
	inner := &rosterAPI{cfg: fx.cfg, writeWait: wait}
	start := time.Now()
	w := serveOne(inner, withAdmin(rosterRequest(http.MethodPost, "/dubplate/roster", "", validPost), fixtureAdmin))
	elapsed := time.Since(start)

	if w.Code != http.StatusServiceUnavailable || outcomeOf(t, w) != outcomeBusy {
		t.Errorf("POST = %d %s, want 503 busy", w.Code, w.Body)
	}
	if elapsed < wait || elapsed > 20*wait {
		t.Errorf("busy answered after %v, want just past the %v wait", elapsed, wait)
	}
	assertUnchanged(t, fx.rosterPath, before)
}
