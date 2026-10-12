package auth

import (
	"errors"
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
}

func newAdminFixture(t *testing.T) adminFixture {
	t.Helper()
	sf := newSeamFixture(t, slices.Clone(adminRosterMembers)...)
	logs := newLogCapture()
	cfg := AdminConfig{
		Seam:     sf.seam,
		ClientIP: func(r *http.Request) string { return serve.ClientIP(r, serve.TrustedProxyConfig{}) },
		Now:      func() time.Time { return gateNow },
		Logger:   slog.New(slog.NewJSONHandler(logs, nil)),
	}
	return adminFixture{
		users:  sf.users,
		roster: sf.roster,
		epoch:  sf.epoch,
		seam:   sf.seam,
		logs:   logs,
		cfg:    cfg,
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

			next := &recordingHandler{}
			got := responseOf(serveOne(RequireAdmin(cfg, next), r.Clone(t.Context())))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("the gate answered %+v, want http.NotFound's %+v", got, want)
			}
			if c := got.header.Values("Set-Cookie"); len(c) > 0 {
				t.Errorf("the gate's refusal set a cookie: %q", c)
			}
			if next.reached {
				t.Error("the gate admitted a refused request to the wrapped handler")
			}

			if !tt.storeClosed {
				assertUnchanged(t, fx.users, before)
			}
			assertRefusalLogs(t, fx.logs, 1)
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
			next := &recordingHandler{}
			w := serveOne(RequireAdmin(fx.cfg, next), r)
			if !next.reached || w.Code != http.StatusTeapot {
				t.Fatalf("status = %d, reached %v, want the admitted handler's 418", w.Code, next.reached)
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

// assertUnchanged fails the test when h's users table no longer reads
// as before, a dumpUsers rendering taken earlier.
func assertUnchanged(t *testing.T, h *store.Handle, before string) {
	t.Helper()
	if after := dumpUsers(t, h); after != before {
		t.Errorf("the users table changed:\nbefore %s\nafter  %s", before, after)
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
				ClientIP: func(r *http.Request) string { return serve.ClientIP(r, serve.TrustedProxyConfig{}) },
				Now:      func() time.Time { return gateNow },
				Logger:   slog.New(slog.NewJSONHandler(logs, nil)),
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
			wantSettingRecord(t, logs.waitFor(t, "admin gate refused"))
		})
	}
}
