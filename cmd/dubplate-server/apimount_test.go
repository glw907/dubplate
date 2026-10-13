package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/serve"
)

// The fixture roster's ids, written out: seedMembers gives the member
// at position i the id whose first two bytes are 0xf0 and i+1.
const (
	apiAdminID  = "f0010000000000000000000000000000"
	apiAdmin2ID = "f0020000000000000000000000000000"
	apiMemberID = "f0030000000000000000000000000000"
)

// apiMembers is two admins and a member, so one admin may leave or be
// demoted while another remains.
var apiMembers = []seedMember{
	{address: "admin@example.com", displayName: "Group Admin", username: "groupadmin", admin: true},
	{address: "admin2@example.com", displayName: "Second Admin", username: "secondadmin", admin: true},
	{address: "member@example.com", displayName: "Member", username: "member"},
}

// A syncBuffer is a log sink the server's goroutines and the test may
// touch together.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// records returns every JSON log record written so far.
func (b *syncBuffer) records(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(b.String()), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("log line %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

// An apiHarness is the real serve shell over a store holding a seeded
// roster, driven through its handler.
type apiHarness struct {
	t       *testing.T
	environ []string
	handler http.Handler
	h       *store.Handle
	prefix  serve.Prefix
	logs    *syncBuffer
}

// newAPIHarness builds the shell over members with environ rewritten by
// tweak, which may be nil.
func newAPIHarness(t *testing.T, members []seedMember, tweak func([]string) []string) *apiHarness {
	t.Helper()
	environ := serveAuthEnviron(t)
	if tweak != nil {
		environ = tweak(environ)
	}
	seedBackupStore(t, environValue(environ, "DUBPLATE_STATE_DIR"))
	seedMembers(t, environValue(environ, "DUBPLATE_STATE_DIR"), members...)
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	prefix, addrs, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig: %v", err)
	}
	logs := &syncBuffer{}
	srv, err := newServeShell(t.Context(), environ, productionRegistry(), prefix, addrs, cfg, slog.New(slog.NewJSONHandler(logs, nil)))
	if err != nil {
		t.Fatalf("newServeShell: %v", err)
	}
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("close the serve build: %v", err)
		}
	})
	return &apiHarness{t: t, environ: environ, handler: srv.shell.Handler(), h: srv.store, prefix: prefix, logs: logs}
}

// session mints a session for address issued at issued.
func (a *apiHarness) session(address string, issued time.Time) string {
	a.t.Helper()
	return mintServeSession(a.t, a.environ, address, issued)
}

// path is the URL path of an API operation.
func (a *apiHarness) path(op string) string { return a.prefix.String() + "/api/v1/" + op }

// call sends one same-origin JSON POST to op with session as its
// cookie, when non-empty.
func (a *apiHarness) call(op, session, body string) *httptest.ResponseRecorder {
	a.t.Helper()
	return a.callAt(a.path(op), session, "same-origin", body)
}

// callAt is call to an arbitrary path with an explicit Sec-Fetch-Site.
func (a *apiHarness) callAt(path, session, site, body string) *httptest.ResponseRecorder {
	a.t.Helper()
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if site != "" {
		r.Header.Set("Sec-Fetch-Site", site)
	}
	if session != "" {
		r.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: session})
	}
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, r)
	return rec
}

// errorCode reads the wire code of rec's JSON error body.
func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("answer %d %.80q is not a JSON error: %v", rec.Code, rec.Body, err)
	}
	return body.Error.Code
}

// TestAPI_OtherPathsUnderAPIAnswerJSONNotFound proves every path under
// <prefix>/api that names no v1 operation answers 404 not_found as JSON,
// session or not (AC4).
func TestAPI_OtherPathsUnderAPIAnswerJSONNotFound(t *testing.T) {
	a := newAPIHarness(t, apiMembers, nil)
	session := a.session("admin@example.com", time.Now())
	base := a.prefix.String() + "/api"
	for _, path := range []string{base, base + "/", base + "/v1", base + "/v2/x", base + "/v2/catalog", base + "/x"} {
		for _, cookie := range []string{"", session} {
			rec := a.callAt(path, cookie, "same-origin", `{}`)
			if rec.Code != http.StatusNotFound {
				t.Errorf("%s (session %v) = %d, want 404", path, cookie != "", rec.Code)
				continue
			}
			if code := errorCode(t, rec); code != "not_found" {
				t.Errorf("%s (session %v) code = %q, want not_found", path, cookie != "", code)
			}
		}
	}
}

// TestAPI_NoValidSessionIs401 proves a request with no valid session,
// a removed member's session, or only a bearer header answers 401
// unauthenticated and reaches no operation, and that a bogus bearer
// header beside a valid cookie changes nothing (AC5).
func TestAPI_NoValidSessionIs401(t *testing.T) {
	a := newAPIHarness(t, append(append([]seedMember(nil), apiMembers...),
		seedMember{address: "gone@example.com", displayName: "Gone", username: "gone", tombstoned: true}), nil)
	now := time.Now()

	for name, cookie := range map[string]string{
		"no cookie":               "",
		"a removed member":        a.session("gone@example.com", now),
		"an address on no roster": a.session("stranger@example.com", now),
	} {
		rec := a.call("catalog", cookie, `{}`)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", name, rec.Code)
			continue
		}
		if code := errorCode(t, rec); code != "unauthenticated" {
			t.Errorf("%s: code = %q, want unauthenticated", name, code)
		}
	}

	bearer := "Bearer 0123456789abcdef0123456789abcdef"
	r := httptest.NewRequest(http.MethodPost, a.path("catalog"), strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	r.Header.Set("Authorization", bearer)
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, r)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("a bearer header with no cookie = %d, want 401", rec.Code)
	}

	r = httptest.NewRequest(http.MethodPost, a.path("catalog"), strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	r.Header.Set("Authorization", bearer)
	r.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: a.session("member@example.com", now)})
	rec = httptest.NewRecorder()
	a.handler.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Errorf("a valid cookie beside a bogus bearer header = %d, want 200 as the cookie's member", rec.Code)
	}
}

// TestAPI_MemberReachesOnlyTheCatalog proves a member's session meets
// the catalog and is refused every admin operation, and an admin's
// session meets them all (AC5).
func TestAPI_MemberReachesOnlyTheCatalog(t *testing.T) {
	a := newAPIHarness(t, apiMembers, nil)
	member := a.session("member@example.com", time.Now())
	admin := a.session("admin@example.com", time.Now())

	if rec := a.call("catalog", member, `{}`); rec.Code != http.StatusOK {
		t.Fatalf("a member's catalog = %d %s, want 200", rec.Code, rec.Body)
	}
	for _, op := range []string{"list_members", "list_settings"} {
		rec := a.call(op, member, `{}`)
		if rec.Code != http.StatusForbidden {
			t.Errorf("a member's %s = %d, want 403", op, rec.Code)
		}
		if rec := a.call(op, admin, `{}`); rec.Code != http.StatusOK {
			t.Errorf("an admin's %s = %d %s, want 200", op, rec.Code, rec.Body)
		}
	}
	rec := a.call("list_members", admin, `{}`)
	var list ops.ListMembersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list.Members) != 3 {
		t.Fatalf("list_members = %s (%v), want the three seeded members", rec.Body, err)
	}
}

// TestAPI_AdminEmailGrantsNothing proves DUBPLATE_ADMIN_EMAIL naming a
// live non-admin changes no privilege: that member's list_members
// answers forbidden (AC21).
func TestAPI_AdminEmailGrantsNothing(t *testing.T) {
	a := newAPIHarness(t, apiMembers, func(environ []string) []string {
		return replaceKey(environ, "DUBPLATE_ADMIN_EMAIL", "member@example.com")
	})
	rec := a.call("list_members", a.session("member@example.com", time.Now()), `{}`)
	if rec.Code != http.StatusForbidden {
		t.Errorf("list_members for the member named by DUBPLATE_ADMIN_EMAIL = %d, want 403", rec.Code)
	}
}

// TestAPI_RecordsTheSessionsActor proves a set_setting body naming an
// actor changes nothing: the audit row records the session's own
// member:<hex id> and the cookie channel, and one log record per
// admitted operation carries the actor, the principal under its own
// key, the channel, and client_ip (AC5, AC6).
func TestAPI_RecordsTheSessionsActor(t *testing.T) {
	a := newAPIHarness(t, apiMembers, nil)
	admin := a.session("admin@example.com", time.Now())
	rec := a.call("set_setting", admin, `{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h","actor":"member:`+apiMemberID+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set_setting = %d %s, want 200", rec.Code, rec.Body)
	}
	var actor, channel string
	err := a.h.QueryRow(t.Context(), `SELECT actor, channel FROM settings_audit WHERE key = 'DUBPLATE_REAUTH_WINDOW'`).Scan(&actor, &channel)
	if err != nil {
		t.Fatalf("read the audit row: %v", err)
	}
	if want := "member:" + apiAdminID; actor != want || channel != "cookie" {
		t.Errorf("audit row = %q over %q, want %q over cookie", actor, channel, want)
	}

	var admitted []map[string]any
	for _, r := range a.logs.records(t) {
		if r["msg"] == "api request admitted" {
			admitted = append(admitted, r)
		}
	}
	if len(admitted) != 1 {
		t.Fatalf("admitted-operation records = %d, want 1: %s", len(admitted), a.logs)
	}
	got := admitted[0]
	if got["op"] != "set_setting" || got["actor"] != "member:"+apiAdminID || got["principal"] != "groupadmin" || got["channel"] != "cookie" {
		t.Errorf("record = %v, want op, actor, principal, and channel", got)
	}
	if ip, _ := got["client_ip"].(string); ip == "" {
		t.Errorf("record = %v, want a client_ip", got)
	}
}

// TestAPI_RefreshKeepsTheAnchor proves a same-origin request sets the
// refreshed cookie and one without Sec-Fetch-Site: same-origin is
// admitted with none, and that a refreshed session whose issue time is
// inside the re-authentication window while its anchor is outside it is
// refused a sensitive write: the gate reads the anchor (AC12).
func TestAPI_RefreshKeepsTheAnchor(t *testing.T) {
	a := newAPIHarness(t, apiMembers, nil)
	stale := a.session("admin@example.com", time.Now().Add(-3*time.Hour))

	cross := a.callAt(a.path("catalog"), stale, "none", `{}`)
	if cross.Code != http.StatusOK || len(cross.Header().Values("Set-Cookie")) != 0 {
		t.Fatalf("a call with Sec-Fetch-Site none = %d with cookies %q, want 200 and none", cross.Code, cross.Header().Values("Set-Cookie"))
	}
	same := a.call("catalog", stale, `{}`)
	if same.Code != http.StatusOK {
		t.Fatalf("a same-origin call = %d, want 200", same.Code)
	}
	cookies := same.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("a same-origin call set %d cookies, want the refreshed session", len(cookies))
	}
	refreshed := cookies[0].Value
	if refreshed == stale {
		t.Fatal("the refreshed session equals the stale one, so nothing refreshed")
	}

	rec := a.call("set_setting", refreshed, `{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("a sensitive write on a refreshed stale session = %d %s, want 403", rec.Code, rec.Body)
	}
	if code := errorCode(t, rec); code != "reauthentication_required" {
		t.Errorf("code = %q, want reauthentication_required", code)
	}
	if rec := a.call("set_setting", a.session("admin@example.com", time.Now()), `{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h"}`); rec.Code != http.StatusOK {
		t.Errorf("the same write on a fresh session = %d %s, want 200", rec.Code, rec.Body)
	}
}

// TestAPI_ServerFailuresAreNot401 proves a failed setting read and a
// failed roster lookup answer 500 internal, never a signed-out 401
// (P3).
func TestAPI_ServerFailuresAreNot401(t *testing.T) {
	t.Run("the setting read fails", func(t *testing.T) {
		a := newAPIHarness(t, apiMembers, nil)
		session := a.session("admin@example.com", time.Now())
		writeAdminRows(t, a.h, map[string]string{"DUBPLATE_SESSION_ABSOLUTE_CAP": "not a duration"})
		rec := a.call("catalog", session, `{}`)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d %s, want 500", rec.Code, rec.Body)
		}
		if code := errorCode(t, rec); code != "internal" {
			t.Errorf("code = %q, want internal", code)
		}
		if !strings.Contains(a.logs.String(), `"level":"ERROR"`) {
			t.Errorf("no Error record in the log: %s", a.logs)
		}
	})
	t.Run("the roster lookup fails", func(t *testing.T) {
		a := newAPIHarness(t, apiMembers, nil)
		session := a.session("admin@example.com", time.Now())
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		if err := a.h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
			_, err := tx.Exec(ctx, `ALTER TABLE users RENAME TO users_gone`)
			return err
		}); err != nil {
			t.Fatalf("break the roster: %v", err)
		}
		rec := a.call("catalog", session, `{}`)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d %s, want 500", rec.Code, rec.Body)
		}
		if code := errorCode(t, rec); code != "internal" {
			t.Errorf("code = %q, want internal", code)
		}
	})
}

// TestAPI_RemovingYourselfSignsYouOut proves an admin who removes
// themself while another admin remains commits, and their next request
// answers 401.
func TestAPI_RemovingYourselfSignsYouOut(t *testing.T) {
	a := newAPIHarness(t, apiMembers, nil)
	admin := a.session("admin@example.com", time.Now())
	rec := a.call("remove_member", admin, `{"user_id":"`+apiAdminID+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("remove_member = %d %s, want 200", rec.Code, rec.Body)
	}
	if rec := a.call("catalog", admin, `{}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("the removed admin's next request = %d, want 401", rec.Code)
	}
}

// TestAPI_AddMemberNeverGrantsAdmin proves a body naming admin adds the
// member without the flag (AC17).
func TestAPI_AddMemberNeverGrantsAdmin(t *testing.T) {
	a := newAPIHarness(t, apiMembers, nil)
	admin := a.session("admin@example.com", time.Now())
	rec := a.call("add_member", admin, `{"address":"new@example.com","display_name":"New Member","username":"newmember","admin":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("add_member = %d %s, want 200", rec.Code, rec.Body)
	}
	var resp ops.AddMemberResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Member.Admin {
		t.Errorf("the added member = %+v, want no admin flag", resp.Member)
	}
	var admins int
	if err := a.h.QueryRow(t.Context(), `SELECT admin FROM users WHERE address = 'new@example.com'`).Scan(&admins); err != nil || admins != 0 {
		t.Errorf("stored admin flag = %d (%v), want 0", admins, err)
	}
}

// envSentinels gives every key only the environment holds a distinct
// value no /api/v1 answer may carry. A path, an address, a secret, a
// host, and every port each gets its own literal, and each is valid for
// its key's validator so the server starts on it. The two flags that
// switch an integration on are left true and the three closed
// enumerations are named in envClosed instead, since their values are
// words no answer could avoid.
var envSentinels = map[string]string{
	"DUBPLATE_LIBRARY_ROOT":               "/srv/sentinel-library-root",
	"DUBPLATE_STAGING_DIR":                "/srv/sentinel-staging-root",
	"DUBPLATE_INTAKE_FORMAT_ALLOWLIST":    "flac",
	"DUBPLATE_SITE_NAME":                  "Sentinel Site Name",
	"DUBPLATE_ADMIN_EMAIL":                "sentinel-admin@example.test",
	"DUBPLATE_MAGIC_LINK_SECRET":          "sentinel-magic-link-secret-0123456789abcdef",
	"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS": "sentinel-magic-link-previous-0123456789abcdef",
	"DUBPLATE_SESSION_SECRET":             "sentinel-session-secret-0123456789abcdefgh",
	"DUBPLATE_SESSION_SECRET_PREVIOUS":    "sentinel-session-previous-0123456789abcdefgh",
	"DUBPLATE_PUBLIC_URL":                 "https://sentinel-public.example.test",
	"DUBPLATE_SMTP_HOST":                  "sentinel-smtp.example.test",
	"DUBPLATE_SMTP_PORT":                  "2525",
	"DUBPLATE_SMTP_USER":                  "sentinel-smtp-user",
	"DUBPLATE_SMTP_PASSWORD":              "sentinel-smtp-password-0123456789abcdefgh",
	"DUBPLATE_SMTP_FROM":                  "sentinel-from@example.test",
	"DUBPLATE_SERVE_BIND_ADDRS":           "127.77.77.77",
	"DUBPLATE_SERVE_PORT":                 "7391",
	"DUBPLATE_SERVE_PREFIX":               "/sentinel-prefix",
	"DUBPLATE_TRUSTED_PROXY_ADDRS":        "203.0.113.77",
	"DUBPLATE_TRUSTED_PROXY_HEADER":       "X-Sentinel-Client-Ip",
	"DUBPLATE_BACKUP_REMOTE":              "teststand:/srv/sentinel-backup-remote",
	"DUBPLATE_STORE_BUSY_TIMEOUT":         "7777ms",
	"DUBPLATE_STORE_MIGRATION_WAIT":       "8888ms",
	"DUBPLATE_OPERATOR_CSS_FILE":          "none",
	"DUBPLATE_BRANDING":                   "none",
	"DUBPLATE_THEME_MATCH":                "dubplate",
	"DUBPLATE_THEME_PALETTE":              "system",
	"DUBPLATE_IDENTITY_SOURCE":            "magic-link",
	"DUBPLATE_TRUSTED_PROXY_ENABLED":      "true",
	"DUBPLATE_BACKUP_ENABLED":             "true",
}

// envClosed names the keys in envSentinels whose value is a word from a
// closed set, a flag, or "none", so the scan skips it.
var envClosed = []string{
	"DUBPLATE_OPERATOR_CSS_FILE", "DUBPLATE_BRANDING", "DUBPLATE_THEME_MATCH", "DUBPLATE_THEME_PALETTE",
	"DUBPLATE_IDENTITY_SOURCE", "DUBPLATE_TRUSTED_PROXY_ENABLED", "DUBPLATE_BACKUP_ENABLED",
	"DUBPLATE_INTAKE_FORMAT_ALLOWLIST",
}

// TestAPI_NoNonStoreValueReachesTheWire proves no value only the
// environment holds appears in the raw body of any /api/v1 answer,
// refusals included (AC9). Every wall, identity-source, integration,
// and threshold key carries a distinct literal, and the test fails when
// list_settings names an environment key this table does not.
func TestAPI_NoNonStoreValueReachesTheWire(t *testing.T) {
	a := newAPIHarness(t, apiMembers, func(environ []string) []string {
		stateDir := filepath.Join(environValue(environ, "DUBPLATE_STATE_DIR"), "sentinel-state-dir")
		if err := os.MkdirAll(stateDir, 0o700); err != nil {
			t.Fatalf("create the sentinel state directory: %v", err)
		}
		environ = replaceKey(environ, "DUBPLATE_STATE_DIR", stateDir)
		for name, value := range envSentinels {
			environ = replaceKey(environ, name, value)
		}
		// serve refuses a staging directory or library root that does not
		// exist, so these two sentinels are real sibling directories.
		music := t.TempDir()
		for _, key := range []string{"DUBPLATE_LIBRARY_ROOT", "DUBPLATE_STAGING_DIR"} {
			dir := filepath.Join(music, filepath.Base(envSentinels[key]))
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			environ = replaceKey(environ, key, dir)
		}
		return environ
	})
	admin := a.session("admin@example.com", time.Now())
	member := a.session("member@example.com", time.Now())

	listing := a.call("list_settings", admin, `{}`)
	var list ops.ListSettingsResponse
	if err := json.Unmarshal(listing.Body.Bytes(), &list); err != nil || len(list.Settings) == 0 {
		t.Fatalf("list_settings = %d %.200s (%v), want the registered keys", listing.Code, listing.Body, err)
	}
	environmentKeys := 0
	for _, entry := range list.Settings {
		if entry.Home != "environment" {
			continue
		}
		environmentKeys++
		if _, ok := envSentinels[entry.Name]; !ok && entry.Name != "DUBPLATE_STATE_DIR" {
			t.Errorf("list_settings names the environment key %s, which the sentinel table does not set", entry.Name)
		}
	}
	if environmentKeys < 20 {
		t.Fatalf("list_settings named %d environment keys, want the registry's whole set", environmentKeys)
	}

	bodies := []string{listing.Body.String()}
	collect := func(rec *httptest.ResponseRecorder) { bodies = append(bodies, rec.Body.String()) }
	for _, cookie := range []string{"", member, admin} {
		for _, op := range []string{"catalog", "list_members", "list_settings", "no_such_operation"} {
			collect(a.call(op, cookie, `{}`))
		}
		collect(a.callAt(a.prefix.String()+"/api/v2/x", cookie, "same-origin", `{}`))
	}
	for _, body := range []string{
		`{"key":"DUBPLATE_LIBRARY_ROOT","value":"/x"}`,
		`{"key":"DUBPLATE_SMTP_PASSWORD","value":"x"}`,
		`{"key":"DUBPLATE_NO_SUCH_KEY","value":"x"}`,
		`{"key":"DUBPLATE_REAUTH_WINDOW","value":"garbage"}`,
		`{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h"}`,
	} {
		collect(a.call("set_setting", admin, body))
	}
	collect(a.call("reset_setting", admin, `{"key":"DUBPLATE_STATE_DIR"}`))
	collect(a.call("reset_setting", admin, `{"key":"DUBPLATE_REAUTH_WINDOW"}`))
	collect(a.call("add_member", admin, `{"address":"not an address","display_name":"x","username":"x"}`))
	collect(a.call("remove_member", admin, `{"user_id":"`+strings.Repeat("0", 32)+`"}`))
	// The sentinel staging directory does not exist, so this create's
	// free-space read fails with an error that names it.
	collect(a.call("create_submission", member, `{"verdict_unit":"album","container":"files","source":"cd","files":[{"path":"01.flac","size":1}]}`))
	collect(a.call("get_submission", member, `{"submission_id":"`+strings.Repeat("ab", 16)+`"}`))

	secrets := maps.Clone(envSentinels)
	for _, key := range []string{"DUBPLATE_STATE_DIR", "DUBPLATE_LIBRARY_ROOT", "DUBPLATE_STAGING_DIR"} {
		secrets[key] = environValue(a.environ, key)
	}
	for _, name := range envClosed {
		delete(secrets, name)
	}
	if len(secrets) < 20 {
		t.Fatalf("the scan holds %d sentinels, want one for every key that holds a value", len(secrets))
	}
	for _, body := range bodies {
		if strings.Contains(strings.ToLower(body), "sentinel") {
			t.Errorf("a partial or reshaped sentinel reached an /api/v1 answer: %.200s", body)
		}
	}
	for name, value := range secrets {
		for _, body := range bodies {
			if strings.Contains(body, value) {
				t.Errorf("the value of %s, %q, reached an /api/v1 answer: %.200s", name, value, body)
			}
		}
	}
}
