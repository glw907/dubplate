package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/serve"
)

// checkRequest builds a request carrying session as its cookie and site
// as its Sec-Fetch-Site header when non-empty.
func checkRequest(session, site string) *http.Request {
	r := rosterRequest(http.MethodPost, "/dubplate/api/v1/catalog", session, "")
	if site != "" {
		r.Header.Set("Sec-Fetch-Site", site)
	}
	return r
}

// TestCheckRequest_AdmitsAnyVerifiedMember proves the check admits a
// verified non-admin as well as an admin, reports the flag the roster
// holds, and never reads it from the request (a claim in the query, a
// header, or the body changes nothing).
func TestCheckRequest_AdmitsAnyVerifiedMember(t *testing.T) {
	for _, tt := range []struct {
		address string
		admin   bool
	}{
		{"admin@example.com", true},
		{"member@example.com", false},
	} {
		t.Run(tt.address, func(t *testing.T) {
			fx := newAdminFixture(t)
			r := rosterRequest(http.MethodPost, "/dubplate/api/v1/catalog?admin=on", mintAt(t, currentFixture, tt.address, seamBirth), "admin=on")
			r.Header.Set("X-Admin", "1")
			w := httptest.NewRecorder()
			v, err := CheckRequest(fx.cfg, w, r)
			if err != nil {
				t.Fatalf("CheckRequest: %v", err)
			}
			if v.Admin != tt.admin || v.Identity.Address != tt.address || !v.Identity.Valid() {
				t.Errorf("verified = admin %v for %q (valid %v), want admin %v for %q", v.Admin, v.Identity.Address, v.Identity.Valid(), tt.admin, tt.address)
			}
		})
	}
}

// TestCheckRequest_RefreshesOnlySameOrigin proves the session is
// refreshed and its cookie set on w only for a same-origin request,
// that the refreshed session keeps its anchor while its issue time
// moves, and that the identity the check returns carries that anchor.
func TestCheckRequest_RefreshesOnlySameOrigin(t *testing.T) {
	for _, tt := range []struct {
		site        string
		wantRefresh bool
	}{
		{"same-origin", true},
		{"same-site", false},
		{"cross-site", false},
		{"none", false},
		{"", false},
	} {
		t.Run("Sec-Fetch-Site "+tt.site, func(t *testing.T) {
			fx := newAdminFixture(t)
			w := httptest.NewRecorder()
			v, err := CheckRequest(fx.cfg, w, checkRequest(mintAt(t, currentFixture, "admin@example.com", seamBirth), tt.site))
			if err != nil {
				t.Fatalf("CheckRequest: %v", err)
			}
			cookies := w.Header().Values("Set-Cookie")
			if !tt.wantRefresh {
				if len(cookies) != 0 {
					t.Errorf("a %q request set %q, want no cookie", tt.site, cookies)
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
			if !sess.AnchoredAt.Equal(seamBirth) || !v.Session.AnchoredAt.Equal(seamBirth) {
				t.Errorf("anchor after a refresh = cookie %v, identity %v, want %v for both", sess.AnchoredAt, v.Session.AnchoredAt, seamBirth)
			}
			if !sess.IssuedAt.Equal(gateNow) {
				t.Errorf("refreshed issue time = %v, want %v, so the refresh visibly moved it", sess.IssuedAt, gateNow)
			}
		})
	}
}

// TestCheckRequest_RefusalsAreNotSignedIn proves every verification
// refusal wraps ErrNotSignedIn, sets no cookie, and logs at Info a
// fixed reason carrying no address and no cookie value.
func TestCheckRequest_RefusalsAreNotSignedIn(t *testing.T) {
	tests := []struct {
		name    string
		seam    Seam
		request func(t *testing.T, fx adminFixture) *http.Request
	}{
		{"no cookie", nil, func(t *testing.T, fx adminFixture) *http.Request { return checkRequest("", "same-origin") }},
		{"forged cookie", nil, func(t *testing.T, fx adminFixture) *http.Request {
			return checkRequest(mintAt(t, foreignFixture, "admin@example.com", seamBirth), "same-origin")
		}},
		{"expired session", nil, func(t *testing.T, fx adminFixture) *http.Request {
			return checkRequest(mintAt(t, currentFixture, "admin@example.com", gateNow.Add(-seamLifetime-time.Hour)), "same-origin")
		}},
		{"session older than its epoch", nil, func(t *testing.T, fx adminFixture) *http.Request {
			if err := fx.epoch.Bump("admin@example.com", seamBirth.Add(time.Hour)); err != nil {
				t.Fatalf("Bump: %v", err)
			}
			return checkRequest(mintAt(t, currentFixture, "admin@example.com", seamBirth), "same-origin")
		}},
		{"member removed from the roster", nil, func(t *testing.T, fx adminFixture) *http.Request {
			return checkRequest(mintAt(t, currentFixture, "gone@example.com", seamBirth), "same-origin")
		}},
		{"verified identity that is not a login", stubSeam{v: VerifiedIdentity{Identity: identity.Identity{Channel: identity.ChannelCommandLine}, Admin: true, RefreshedValue: "refreshed-by-stub"}}, func(t *testing.T, fx adminFixture) *http.Request {
			return checkRequest("", "same-origin")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fx := newAdminFixture(t)
			cfg := fx.cfg
			if tt.seam != nil {
				cfg.Seam = tt.seam
			}
			r := tt.request(t, fx)
			w := httptest.NewRecorder()
			_, err := CheckRequest(cfg, w, r)
			if !errors.Is(err, ErrNotSignedIn) {
				t.Fatalf("CheckRequest = %v, want an error wrapping ErrNotSignedIn", err)
			}
			if c := w.Header().Values("Set-Cookie"); len(c) != 0 {
				t.Errorf("a refusal set a cookie: %q", c)
			}
			assertRefusalLogs(t, fx.logs, 1)
			for _, rec := range fx.logs.records(t) {
				if rec["level"] != "INFO" {
					t.Errorf("refusal logged at %v, want INFO: %v", rec["level"], rec)
				}
			}
			if c, cerr := r.Cookie(sessionCookieName); cerr == nil && strings.Contains(fx.logs.text(), c.Value) {
				t.Errorf("a refusal log carries the cookie value:\n%s", fx.logs.text())
			}
		})
	}
}

// TestCheckRequest_ServerFailuresAreNotRefusals proves a failed roster
// lookup or setting read is the server failing: the error is not a
// not-signed-in refusal, the log holds an Error record, and nothing is
// set. A refusal answered here would sign a member out on a store
// outage.
func TestCheckRequest_ServerFailuresAreNotRefusals(t *testing.T) {
	t.Run("the roster lookup fails", func(t *testing.T) {
		fx := newAdminFixture(t)
		if err := fx.users.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
		w := httptest.NewRecorder()
		_, err := CheckRequest(fx.cfg, w, checkRequest(mintAt(t, currentFixture, "admin@example.com", seamBirth), "same-origin"))
		if err == nil || errors.Is(err, ErrNotSignedIn) {
			t.Fatalf("CheckRequest = %v, want a server failure that is no ErrNotSignedIn", err)
		}
		if c := w.Header().Values("Set-Cookie"); len(c) != 0 {
			t.Errorf("a failed lookup set a cookie: %q", c)
		}
		assertRefusalLogs(t, fx.logs, 1)
		if rec := fx.logs.records(t)[0]; rec["level"] != "ERROR" {
			t.Errorf("failure logged at %v, want ERROR: %v", rec["level"], rec)
		}
	})
	for _, tt := range []struct {
		name    string
		failCap bool
		site    string
	}{
		{"the cap reader fails", true, "cross-site"},
		{"the lifetime reader fails on a same-origin refresh", false, "same-origin"},
	} {
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
			w := httptest.NewRecorder()
			_, err := CheckRequest(cfg, w, checkRequest(mintAt(t, currentFixture, "admin@example.com", seamBirth), tt.site))
			if err == nil || errors.Is(err, ErrNotSignedIn) {
				t.Fatalf("CheckRequest = %v, want a server failure that is no ErrNotSignedIn", err)
			}
			if c := w.Header().Values("Set-Cookie"); len(c) != 0 {
				t.Errorf("a failed read set a cookie: %q", c)
			}
			wantSettingRecord(t, logs.waitFor(t, "session check failed"))
		})
	}
}
