package main

import (
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// A routeClass says who may reach a route: anyone, any signed-in
// member, or an admin.
type routeClass int

const (
	routePublic routeClass = iota + 1
	routeSession
	routeAdmin
)

// routeRow is one route the serve build registers: its class, and for
// a session or admin route the request that must be refused when nobody
// is signed in.
type routeRow struct {
	class routeClass

	// method and path are the request an unauthenticated visitor
	// sends, path relative to the owned prefix.
	method, path string

	// optional marks a route the build registers only when its
	// configuration asks for it.
	optional bool
}

// routeInventory names every route the serve build registers and its
// class (PS A2). A route the table does not name fails the test, so a
// new route cannot ship without a ruling on who may reach it.
var routeInventory = map[string]routeRow{
	requestLinkRouteName: {class: routePublic},
	consumeLinkRouteName: {class: routePublic},
	signOutRouteName:     {class: routePublic},
	serve.ProbeRouteName: {class: routePublic},
	loginRouteName:       {class: routePublic},
	assetsRouteName:      {class: routePublic},
	fontsRouteName:       {class: routePublic},
	faviconSVGRouteName:  {class: routePublic},
	faviconPNGRouteName:  {class: routePublic},
	markRouteName:        {class: routePublic},
	brandingRouteName:    {class: routePublic},
	operatorCSSRouteName: {class: routePublic, optional: true},

	apiRouteName:     {class: routeSession, method: http.MethodPost, path: "/api/v1/catalog"},
	tusRouteName:     {class: routeSession, method: http.MethodPost, path: "/" + tusRouteName + "/"},
	membersRouteName: {class: routeAdmin, method: http.MethodGet, path: "/" + membersRouteName},

	serverSettingsRouteName: {class: routeAdmin, method: http.MethodGet, path: "/" + serverSettingsRouteName},
}

// A recordingRegistrar forwards every registration to its shell and
// notes the name.
type recordingRegistrar struct {
	*serve.Shell
	names *[]string
}

func (r recordingRegistrar) Register(name string, h http.Handler) {
	*r.names = append(*r.names, name)
	r.Shell.Register(name, h)
}

// TestRouteInventory proves the table names exactly the routes the real
// serve build registers, and that an unauthenticated request is refused
// on every session and admin row: a session route answers 401, an admin
// route answers the gate's 404, and a signed-in non-admin meets that
// same 404 on an admin route.
func TestRouteInventory(t *testing.T) {
	environ := rosterAPIEnviron(t)
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	prefix, addrs, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig: %v", err)
	}
	var registered []string
	srv, err := newServeShellVia(t.Context(), environ, productionRegistry(), prefix, addrs, cfg, discardTestLogger(),
		func(s *serve.Shell) registrar { return recordingRegistrar{Shell: s, names: &registered} })
	if err != nil {
		t.Fatalf("newServeShellVia: %v", err)
	}
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("close the serve build: %v", err)
		}
	})

	for _, mismatch := range inventoryMismatches(routeInventory, registered) {
		t.Error(mismatch)
	}

	// A fixture inventory missing the tus row must fail the comparison
	// above, so the test cannot pass with the row gone.
	missing := maps.Clone(routeInventory)
	delete(missing, tusRouteName)
	if got := inventoryMismatches(missing, registered); !slices.ContainsFunc(got, func(m string) bool { return strings.Contains(m, `"tus"`) }) {
		t.Errorf("an inventory missing the tus row reports %q, want the tus route named", got)
	}

	handler := srv.shell.Handler()
	member := mintServeSession(t, environ, rosterMemberAddress, time.Now())
	for name, row := range routeInventory {
		if row.class == routePublic || !slices.Contains(registered, name) {
			continue
		}
		send := func(session string) *httptest.ResponseRecorder {
			r := httptest.NewRequest(row.method, prefix.String()+row.path, strings.NewReader(`{}`))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Sec-Fetch-Site", "same-origin")
			if session != "" {
				r.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: session})
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r)
			return rec
		}
		switch row.class {
		case routeSession:
			if rec := send(""); rec.Code != http.StatusUnauthorized {
				t.Errorf("%s: an unauthenticated request = %d, want 401", name, rec.Code)
			}
		case routeAdmin:
			for who, session := range map[string]string{"an unauthenticated request": "", "a signed-in non-admin": member} {
				rec := send(session)
				if rec.Code != http.StatusNotFound || rec.Body.String() != wantNotFoundBody {
					t.Errorf("%s: %s = %d %.60q, want the gate's 404", name, who, rec.Code, rec.Body)
				}
			}
		}
	}
}

// inventoryMismatches names each route registered that inventory lacks,
// and each route inventory requires that is not registered.
func inventoryMismatches(inventory map[string]routeRow, registered []string) []string {
	var out []string
	for _, name := range registered {
		if _, ok := inventory[name]; !ok {
			out = append(out, fmt.Sprintf("the build registers %q, which the inventory does not name", name))
		}
	}
	for name, row := range inventory {
		if !row.optional && !slices.Contains(registered, name) {
			out = append(out, fmt.Sprintf("the inventory names %q, which the build does not register", name))
		}
	}
	return out
}

// TestServerSettingsRoute_ServesAdminsOnly proves the settings document
// answers a verified admin with the page, never cached, and answers a
// signed-in non-admin, a visitor with no session, and a forged one the
// gate's 404. The inventory pins the class on the same route.
func TestServerSettingsRoute_ServesAdminsOnly(t *testing.T) {
	environ := rosterAPIEnviron(t)
	shell, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell: %v", err)
	}
	handler := shell.Handler()
	get := func(session string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, sessionRequest(http.MethodGet, "/dubplate/server-settings", session, "", ""))
		return rec
	}

	admin := get(mintServeSession(t, environ, rosterAdminAddress, time.Now()))
	if admin.Code != http.StatusOK || !strings.Contains(admin.Body.String(), "<html") {
		t.Fatalf("admin GET = %d %.80q, want the page", admin.Code, admin.Body)
	}
	if got := admin.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	for who, session := range map[string]string{
		"a non-admin":  mintServeSession(t, environ, rosterMemberAddress, time.Now()),
		"no session":   "",
		"a forged one": "not-a-session",
	} {
		rec := get(session)
		if rec.Code != http.StatusNotFound || rec.Body.String() != wantNotFoundBody {
			t.Errorf("%s: GET = %d %.60q, want the gate's 404", who, rec.Code, rec.Body)
		}
	}
}
