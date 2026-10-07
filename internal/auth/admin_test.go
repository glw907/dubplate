package auth

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/serve"
)

// adminRosterMembers holds two admins and a non-admin, so a hand edit
// demoting the first admin still leaves the roster an admin.
var adminRosterMembers = []userRow{
	{id: adminID, address: "admin@example.com", displayName: "Group Admin", username: "groupadmin", admin: true},
	{address: "founder@example.com", displayName: "Founder", username: "founder", admin: true},
	{address: "member@example.com", displayName: "Member", username: "member"},
}

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
	users  *store.Handle
	roster *Roster
	epoch  *EpochStore
	seam   *MagicLinkSeam
	logs   *logCapture
	cfg    AdminConfig
	api    http.Handler
}

func newAdminFixture(t *testing.T) adminFixture {
	t.Helper()
	sf := newSeamFixture(t, slices.Clone(adminRosterMembers)...)
	logs := newLogCapture()
	cfg := AdminConfig{
		Seam:     sf.seam,
		Roster:   sf.roster,
		ClientIP: func(r *http.Request) string { return serve.ClientIP(r, serve.TrustedProxyConfig{}) },
		Now:      func() time.Time { return gateNow },
		Logger:   slog.New(slog.NewJSONHandler(logs, nil)),
		BodyCap:  fixed[int64](4096),
	}
	return adminFixture{
		users:  sf.users,
		roster: sf.roster,
		epoch:  sf.epoch,
		seam:   sf.seam,
		logs:   logs,
		cfg:    cfg,
		api:    NewRosterAPI(cfg),
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
		r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session})
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
		UserID:      identity.UserID{0xad, 0x01},
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
		name        string
		seam        Seam
		storeClosed bool // the store is closed, so no table is left to compare
		request     func(t *testing.T, fx adminFixture) *http.Request
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
			name:        "roster store fails on the lookup",
			storeClosed: true,
			request: func(t *testing.T, fx adminFixture) *http.Request {
				if err := fx.users.Close(); err != nil {
					t.Fatalf("close store: %v", err)
				}
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
			before := dumpUsers(t, fx.users)
			r := tt.request(t, fx)
			var cookieValue string
			if c, err := r.Cookie(sessionCookieName); err == nil {
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

			if !tt.storeClosed {
				assertUnchanged(t, fx.users, before)
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
// in the order members were added with exactly the four fields per member, HEAD answers
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
		map[string]any{"address": "founder@example.com", "display_name": "Founder", "username": "founder", "admin": true},
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

// TestRosterAPI_ListStoreFailure proves a roster whose store fails
// answers failed and no member list. It runs on the inner handler:
// through the gate the seam's own lookup reads the same broken store and
// refuses first.
func TestRosterAPI_ListStoreFailure(t *testing.T) {
	fx := newAdminFixture(t)
	if err := fx.users.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
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
			before := dumpUsers(t, fx.users)
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
				assertUnchanged(t, fx.users, before)
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
			if _, err := fx.roster.Lookup(t.Context(), "new@example.com"); err != nil {
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

// assertUnchanged fails the test when h's users table no longer reads
// as before, a dumpUsers rendering taken earlier.
func assertUnchanged(t *testing.T, h *store.Handle, before string) {
	t.Helper()
	if after := dumpUsers(t, h); after != before {
		t.Errorf("the users table changed:\nbefore %s\nafter  %s", before, after)
	}
}

// TestRosterAPI_PostOverCap proves the body cap is installed before the
// form parses: an oversized body is refused having been read no further
// than one byte past the cap.
func TestRosterAPI_PostOverCap(t *testing.T) {
	fx := newAdminFixture(t)
	before := dumpUsers(t, fx.users)
	const bodyCap = 512
	fx.cfg.BodyCap = fixed[int64](bodyCap)
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
	assertUnchanged(t, fx.users, before)
}

// TestRosterAPI_PostFailed proves the refusals a verified admin can
// still meet at the store answer 500 failed and log no audit record.
// They run on the inner handler, since each changes the roster after
// the identity was verified. A refusal the write decides changes no row;
// a closed store has no table left to compare.
func TestRosterAPI_PostFailed(t *testing.T) {
	tests := []struct {
		name        string
		storeClosed bool
		change      func(t *testing.T, h *store.Handle)
	}{
		{
			name: "actor demoted after the identity verified",
			change: func(t *testing.T, h *store.Handle) {
				err := h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
					_, err := tx.Exec(ctx, `UPDATE users SET admin = 0 WHERE address = 'admin@example.com'`)
					return err
				})
				if err != nil {
					t.Fatalf("demote the actor: %v", err)
				}
			},
		},
		{
			name:        "the store closed",
			storeClosed: true,
			change: func(t *testing.T, h *store.Handle) {
				if err := h.Close(); err != nil {
					t.Fatalf("close store: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newAdminFixture(t)
			tt.change(t, fx.users)
			var before string
			if !tt.storeClosed {
				before = dumpUsers(t, fx.users)
			}
			inner := &rosterAPI{cfg: fx.cfg, writeWait: time.Second}

			w := serveOne(inner, withAdmin(rosterRequest(http.MethodPost, "/dubplate/roster", "", validPost), fixtureAdmin))
			if w.Code != http.StatusInternalServerError || outcomeOf(t, w) != outcomeFailed {
				t.Errorf("POST = %d %s, want 500 failed", w.Code, w.Body)
			}
			if !tt.storeClosed {
				assertUnchanged(t, fx.users, before)
			}
			if audits := auditRecords(t, fx.logs); len(audits) != 0 {
				t.Errorf("a refused POST logged an audit record: %v", audits)
			}
		})
	}
}

// TestRosterAPI_PostDuplicateUsernameOverUsersTable proves the page's
// half of the username rule on the users table: through the gate, a new
// address asking for an existing username in another case answers
// duplicate_username, never a constraint failure's 500, and writes
// nothing.
func TestRosterAPI_PostDuplicateUsernameOverUsersTable(t *testing.T) {
	h := newUsersStore(t)
	insertUsers(t, h, liveAdmin)
	roster := NewRoster(h)
	logs := newLogCapture()
	cfg := AdminConfig{
		Seam:     NewMagicLinkSeam(roster, newEpochStore(t, t.TempDir()), []byte(currentFixture), []byte(previousFixture), fixed(seamLifetime), fixed(seamCap)),
		Roster:   roster,
		ClientIP: func(r *http.Request) string { return serve.ClientIP(r, serve.TrustedProxyConfig{}) },
		Now:      func() time.Time { return gateNow },
		Logger:   slog.New(slog.NewJSONHandler(logs, nil)),
		BodyCap:  fixed[int64](4096),
	}
	before := dumpUsers(t, h)

	body := "address=new%40example.com&display_name=New&username=GROUPADMIN"
	w := serveOne(NewRosterAPI(cfg), rosterRequest(http.MethodPost, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), body))
	if w.Code != http.StatusConflict || outcomeOf(t, w) != outcomeDuplicateUsername {
		t.Fatalf("POST = %d %s, want 409 %s", w.Code, w.Body, outcomeDuplicateUsername)
	}
	if after := dumpUsers(t, h); after != before {
		t.Errorf("a refused POST changed the users table:\nbefore %s\nafter  %s", before, after)
	}
}

// TestRequireAdmin_RefusalLogLevel proves a refusal a visitor can
// cause logs at Info under the cause's own name, while a lookup the
// store failed logs at Error under a fixed reason that carries none of
// the error's text. Every row still answers http.NotFound's bytes.
func TestRequireAdmin_RefusalLogLevel(t *testing.T) {
	tests := []struct {
		name       string
		request    func(t *testing.T, fx adminFixture) *http.Request
		wantLevel  string
		wantReason string
	}{
		{
			name: "forged cookie",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, foreignFixture, "admin@example.com", seamBirth), "")
			},
			wantLevel:  "INFO",
			wantReason: ErrSessionMAC.Error(),
		},
		{
			name: "member removed from the roster",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "gone@example.com", seamBirth), "")
			},
			wantLevel:  "INFO",
			wantReason: ErrMemberNotFound.Error(),
		},
		{
			name: "store closed under the lookup",
			request: func(t *testing.T, fx adminFixture) *http.Request {
				if err := fx.users.Close(); err != nil {
					t.Fatalf("close store: %v", err)
				}
				return rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), "")
			},
			wantLevel:  "ERROR",
			wantReason: "roster lookup failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newAdminFixture(t)
			r := tt.request(t, fx)
			w := serveOne(RequireAdmin(fx.cfg, &recordingHandler{}), r)
			if want := notFoundFor(r.Clone(t.Context())); w.Code != want.status {
				t.Errorf("status = %d, want %d", w.Code, want.status)
			}
			var gate []map[string]any
			for _, rec := range fx.logs.records(t) {
				if rec["msg"] == "admin gate refused" {
					gate = append(gate, rec)
				}
			}
			if len(gate) != 1 {
				t.Fatalf("gate refusal records = %d, want 1:\n%s", len(gate), fx.logs.text())
			}
			if gate[0]["level"] != tt.wantLevel || gate[0]["reason"] != tt.wantReason {
				t.Errorf("record level %v reason %v, want %s %q", gate[0]["level"], gate[0]["reason"], tt.wantLevel, tt.wantReason)
			}
		})
	}
}

// TestRosterAPI_PostWriteLockWait proves a POST that cannot get the
// store's write connection before its wait ends answers 503 busy and
// writes nothing, while the connection's holder is untouched.
func TestRosterAPI_PostWriteLockWait(t *testing.T) {
	fx := newAdminFixture(t)
	before := dumpUsers(t, fx.users)

	held, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- fx.users.Update(writeCtx(t), func(context.Context, *store.Tx) error {
			close(held)
			<-release
			return nil
		})
	}()
	<-held
	inner := &rosterAPI{cfg: fx.cfg, writeWait: 50 * time.Millisecond}
	w := serveOne(inner, withAdmin(rosterRequest(http.MethodPost, "/dubplate/roster", "", validPost), fixtureAdmin))
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("holder's Update: %v", err)
	}

	if w.Code != http.StatusServiceUnavailable || outcomeOf(t, w) != outcomeBusy {
		t.Errorf("POST = %d %s, want 503 busy", w.Code, w.Body)
	}
	assertUnchanged(t, fx.users, before)
}

// TestAdditionOutcome_StoreBusy proves a write that the store gave up on
// after its busy timeout, as a cross-process lock holder causes, answers
// 503 busy rather than the generic failure.
func TestAdditionOutcome_StoreBusy(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"bare", store.ErrBusy},
		{"wrapped", fmt.Errorf("add member: %w", store.ErrBusy)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, outcome := additionOutcome(tt.err)
			if status != http.StatusServiceUnavailable || outcome != outcomeBusy {
				t.Errorf("additionOutcome = %d %q, want 503 %q", status, outcome, outcomeBusy)
			}
		})
	}
}

// TestRequireAdmin_SettingReadFailure proves a failed setting read in
// the seam reaches the gate's server-error path: the gate answers
// http.NotFound's bytes, as for every refusal, sets no cookie, never
// calls next, and logs at Error with the reader's text, which names the
// key.
func TestRequireAdmin_SettingReadFailure(t *testing.T) {
	tests := []struct {
		name    string
		failCap bool
		sfs     string
	}{
		{"the cap reader fails", true, "cross-site"},
		{"the lifetime reader fails on a same-origin refresh", false, "same-origin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifetime, absoluteCap := newSettingFake(seamLifetime), newSettingFake(seamCap)
			sf := newSeamFixtureReading(t, lifetime.setting(), absoluteCap.setting(), slices.Clone(adminRosterMembers)...)
			logs := newLogCapture()
			cfg := AdminConfig{
				Seam:     sf.seam,
				Roster:   sf.roster,
				ClientIP: func(r *http.Request) string { return serve.ClientIP(r, serve.TrustedProxyConfig{}) },
				Now:      func() time.Time { return gateNow },
				Logger:   slog.New(slog.NewJSONHandler(logs, nil)),
				BodyCap:  fixed[int64](4096),
			}
			if tt.failCap {
				absoluteCap.set(0, errFixtureRead)
			} else {
				lifetime.set(0, errFixtureRead)
			}
			next := &recordingHandler{}
			r := rosterRequest(http.MethodGet, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), "")
			r.Header.Set("Sec-Fetch-Site", tt.sfs)
			w := serveOne(RequireAdmin(cfg, next), r)

			if got, want := responseOf(w), notFoundFor(r.Clone(t.Context())); got.status != want.status || got.body != want.body {
				t.Errorf("response = %d %q, want %d %q", got.status, got.body, want.status, want.body)
			}
			if len(w.Result().Cookies()) != 0 {
				t.Errorf("a failed read set a cookie")
			}
			if next.reached {
				t.Errorf("the gate called next after a failed read")
			}
			rec := logs.waitFor(t, "admin gate refused")
			reason, _ := rec["reason"].(string)
			if rec["level"] != "ERROR" || !strings.Contains(reason, "DUBPLATE_FIXTURE_KEY") {
				t.Errorf("record level %v reason %q, want ERROR naming the key", rec["level"], reason)
			}
		})
	}
}

// TestRosterAPI_PostBodyCapReadFailure proves a POST reads the body cap
// when it starts and answers a failed read with a server error that
// writes nothing, rather than parse under a default cap.
func TestRosterAPI_PostBodyCapReadFailure(t *testing.T) {
	fx := newAdminFixture(t)
	before := dumpUsers(t, fx.users)
	bodyCap := newSettingFake[int64](4096)
	fx.cfg.BodyCap = bodyCap.setting()
	api := NewRosterAPI(fx.cfg)
	bodyCap.set(0, errFixtureRead)

	w := serveOne(api, rosterRequest(http.MethodPost, "/dubplate/roster", mintAt(t, currentFixture, "admin@example.com", seamBirth), validPost))
	wantSettingFailure(t, w, fx.logs)
	assertUnchanged(t, fx.users, before)
}
