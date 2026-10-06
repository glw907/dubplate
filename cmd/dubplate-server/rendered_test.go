//go:build rendered

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/auth"
)

// renderedPrefix is the owned prefix the fixture server mounts under.
// It differs from the default so every rendered row also proves the
// bundle resolves its URLs from the runtime prefix.
const renderedPrefix = "/rendered"

// renderedScriptTimeout bounds one run of the browser script, at more
// than twice its measured wall time on a contended machine (378s;
// about 257s uncontended, the capture matrix's 150 fresh browser
// contexts dominating it) and under
// renderedMagicLinkExpiry, so a late capture never meets an expired
// link. The Makefile's RENDERED_TIMEOUT sits above it plus the
// missing-browser run, and below the expiry too.
const renderedScriptTimeout = 13 * time.Minute

// renderedInstallHint is the command the script names when
// Playwright's Chromium is not installed.
const renderedInstallHint = "npx --prefix web playwright install chromium"

// renderedMagicLinkExpiry is the fixture's own DUBPLATE_MAGIC_LINK_EXPIRY,
// raised from the 15m default (`magicLinkExpiryThreshold`, `auth.go`)
// so the capture matrix's own minted links, spent well after the run
// starts, stay valid for renderedScriptTimeout's own length: it stays
// under magicLinkExpiryCeiling (one hour, `auth.go`).
const renderedMagicLinkExpiry = 25 * time.Minute

// The fixture's four rate limits, set through their configuration keys.
// A full run's spend (renderedSpend) must stay below each, so no row
// meets a rate_limited answer it did not ask for.
const (
	renderedLinkRequestsPerAddress       = 50
	renderedLinkRequestsPerClient        = 100
	renderedConsumptionAttemptsPerClient = 100
	renderedLinkMailCeiling              = 100
)

// renderedBrandingPrimary and renderedOperatorPrimary are the
// --color-primary values G7's computed half (R-5ciii8) checks. Neither
// matches any shipped theme's own declared value (web/src/lib/themes/
// label.css, label-night.css, and system.css all declare #c1352b), so
// a passing row proves a real override rather than a coincidence.
const (
	renderedBrandingPrimary = "#336699"
	renderedOperatorPrimary = "#1a2b3c"
)

// A renderedRow is one row of the manifest: the ID the script must
// report exactly once, and the link requests and link consumptions the
// row spends against the fixture's rate limits.
type renderedRow struct {
	id           string
	linkRequests int
	consumptions int
}

// renderedManifest is every row the browser script must report. A row
// the script never reports, a row it reports that is not here, and a
// row reported without its fixture-state result each fail. It is
// renderedBaseManifest plus the capture matrix's own rows
// (renderedCaptureRows).
var renderedManifest = append(append([]renderedRow(nil), renderedBaseManifest...), renderedCaptureRows()...)

// renderedBaseManifest is every row the harness reported before the
// capture matrix.
var renderedBaseManifest = []renderedRow{
	{id: "first-paint/cookie-none/light"},
	{id: "first-paint/cookie-none/dark"},
	{id: "first-paint/cookie-label/light"},
	{id: "first-paint/cookie-label/dark"},
	{id: "first-paint/cookie-label-night/light"},
	{id: "first-paint/cookie-label-night/dark"},
	{id: "csp-clean/login"},
	{id: "csp-clean/login-link"},
	{id: "csp-violation-listener/img-src"},
	{id: "g7/computed/with-branding/label"},
	{id: "g7/computed/with-branding/label-night"},
	{id: "g7/computed/with-branding/system-light"},
	{id: "g7/computed/with-branding/system-dark"},
	{id: "g7/computed/without-branding/label"},
	{id: "g7/computed/without-branding/label-night"},
	{id: "g7/computed/without-branding/system-light"},
	{id: "g7/computed/without-branding/system-dark"},
	{id: "g7/computed/operator/label"},
	{id: "g10/structural/dubplate-label"},
	{id: "g10/structural/dubplate-label-night"},
	{id: "g10/light-control"},
	{id: "head/theme-color/system/light"},
	{id: "head/theme-color/system/dark"},
	{id: "head/theme-color/cookie-label"},
	{id: "head/theme-color/cookie-label-night"},
	{id: "head/icon/login"},
	{id: "head/icon/login-link"},
	{id: "head/icon/client-navigation"},
	{id: "head/no-favicon-ico"},
	{id: "component/focus/login/label"},
	{id: "component/focus/login/label-night"},
	{id: "component/focus/login-link/label"},
	{id: "component/focus/login-link/label-night"},
	{id: "component/focus/login/system-light"},
	{id: "component/focus/login/system-dark"},
	{id: "component/focus/login-link/system-light"},
	{id: "component/focus/login-link/system-dark"},
	{id: "component/pressed"},
	{id: "component/reduced-motion"},
	{id: "component/pressed/flat"},
	{id: "component/reduced-motion/kill-switch"},
	{id: "component/surface/shadow/button/label"},
	{id: "component/surface/shadow/button/label-night"},
	{id: "component/surface/shadow/card/label"},
	{id: "component/surface/shadow/card/label-night"},
	{id: "component/surface/shadow/alert/label"},
	{id: "component/surface/shadow/alert/label-night"},
	{id: "component/surface/keyline-base/label"},
	{id: "component/surface/keyline-base/label-night"},
	{id: "component/surface/keyline-fill/label"},
	{id: "component/surface/keyline-fill/label-night"},
	{id: "component/surface/keyline-outline/label"},
	{id: "component/surface/keyline-outline/label-night"},
	{id: "component/surface/border-width/label"},
	{id: "component/surface/border-width/label-night"},
	{id: "component/surface/ghost/label"},
	{id: "component/surface/ghost/label-night"},
	{id: "component/surface/keyline-controls/label"},
	{id: "component/surface/keyline-controls/label-night"},
	{id: "component/surface/flat/label"},
	{id: "component/surface/flat/label-night"},
	{id: "component/badge/label"},
	{id: "component/badge/label-night"},
	{id: "component/badge/system-light"},
	{id: "component/badge/system-dark"},
	{id: "component/input-states/label"},
	{id: "component/input-states/label-night"},
	{id: "component/link/label"},
	{id: "component/link/label-night"},
	{id: "component/utility-wins"},
	{id: "component/display-face"},
	{id: "component/page-title"},
	{id: "fonts/alfa-slab-one"},
	{id: "fonts/work-sans"},
	{id: "fonts/ibm-plex-mono"},
	{id: "auth/static/login"},
	{id: "auth/static/login-link"},
	{id: "auth/login/site-name-heading"},
	{id: "auth/login/sent"},
	{id: "auth/login/requested"},
	{id: "auth/login/unknown-with-admin", linkRequests: 1},
	{id: "auth/login/unknown-no-admin"},
	{id: "auth/login/rate-limited"},
	{id: "auth/login/invalid-address", linkRequests: 1},
	{id: "auth/login/too-large"},
	{id: "auth/login/malformed"},
	{id: "auth/login/failed", linkRequests: 1},
	{id: "auth/login/cross-origin"},
	{id: "auth/login/network-failure"},
	{id: "auth/login/single-outcome"},
	{id: "auth/login/inherited-outcome"},
	{id: "auth/login/non-json-5xx"},
	{id: "auth/login/repeat-outcome"},
	{id: "auth/login/busy-focus"},
	{id: "auth/login/different-address-focus"},
	{id: "auth/interstitial/token-order", consumptions: 1},
	{id: "auth/interstitial/token-order/fragment", consumptions: 1},
	{id: "auth/interstitial/signed-in", consumptions: 1},
	{id: "auth/interstitial/refused-reload"},
	{id: "auth/interstitial/refused-reload/fragment"},
	{id: "auth/interstitial/refused"},
	{id: "auth/interstitial/rate-limited"},
	{id: "auth/interstitial/too-large"},
	{id: "auth/interstitial/malformed"},
	{id: "auth/interstitial/failed"},
	{id: "members/add-member", consumptions: 1},
	{id: "members/non-admin", consumptions: 1},
	// Every row from here to the end of this list shares one signed-in
	// admin session, established once by `establishSession` and reused
	// by cookie across every context; the one consumption it spends is
	// declared on the first of these rows. Sign-out clears the cookie in
	// its own context only and revokes nothing, so the rows after it
	// still hold a live session.
	{id: "component/focus/members/label", consumptions: 1},
	{id: "component/focus/members/label-night"},
	{id: "component/focus/members/system-light"},
	{id: "component/focus/members/system-dark"},
	{id: "header/contrast/label"},
	{id: "header/contrast/label-night"},
	{id: "header/ruling-4"},
	{id: "layout/content-measure"},
	{id: "header/wordmark-clearance"},
	{id: "members/list-join"},
	{id: "members/unreachable"},
	{id: "members/malformed-roster/empty-object"},
	{id: "members/malformed-roster/null-members"},
	{id: "component/typography-ladder"},
	{id: "header/toggle/system-light"},
	{id: "header/toggle/system-dark"},
	{id: "header/toggle/cookie-label"},
	{id: "header/toggle/cookie-label-night"},
	{id: "header/structure"},
	{id: "header/sign-out-contrast/label"},
	{id: "header/sign-out-contrast/label-night"},
	{id: "members/inherited-outcome"},
	{id: "header/sign-out/in-flight"},
	{id: "header/sign-out/signed-out"},
	{id: "header/sign-out/network-failure"},
	{id: "header/sign-out/cross-origin"},
	{id: "header/sign-out/not-ok"},
	{id: "header/sign-out/not-found"},
	{id: "harness/settle-guard"},
}

// renderedCaptureWidths are W4.1's five CSS-pixel widths the capture
// matrix renders every state at.
var renderedCaptureWidths = []int{320, 390, 768, 1440, 2560}

// renderedCaptureSchemes are the two emulated color-scheme preferences
// the capture matrix renders every state under.
var renderedCaptureSchemes = []string{"light", "dark"}

// renderedCaptureStates is the capture matrix's fifteen page states, as
// Go literals. linkRequests and consumptions are per cell (one of the state's ten width-by-scheme
// captures), not per state: a state that signs in mints and consumes
// one link per cell, since the harness opens a fresh browser context,
// and so a fresh session, for every cell.
var renderedCaptureStates = []struct {
	id           string
	linkRequests int
	consumptions int
}{
	{id: "login/empty"},
	{id: "login/sent"},
	{id: "login/unknown-with-admin", linkRequests: 1},
	{id: "login/failed", linkRequests: 1},
	{id: "login/rate-limited"},
	{id: "login/malformed"},
	{id: "interstitial/ready"},
	{id: "interstitial/refused"},
	{id: "interstitial/signed-in", consumptions: 1},
	{id: "members/populated", consumptions: 1},
	{id: "members/empty", consumptions: 1},
	{id: "members/added", consumptions: 1},
	{id: "members/duplicate-address", consumptions: 1},
	{id: "members/non-admin", consumptions: 1},
	{id: "error/not-found"},
}

// renderedCaptureRows expands renderedCaptureStates over every width and
// scheme into the capture matrix's own 150 manifest rows (15 states by
// 5 widths by 2 schemes), each ID "capture/<state>/<width>/<scheme>".
func renderedCaptureRows() []renderedRow {
	rows := make([]renderedRow, 0, len(renderedCaptureStates)*len(renderedCaptureWidths)*len(renderedCaptureSchemes))
	for _, s := range renderedCaptureStates {
		for _, w := range renderedCaptureWidths {
			for _, scheme := range renderedCaptureSchemes {
				rows = append(rows, renderedRow{
					id:           fmt.Sprintf("capture/%s/%d/%s", s.id, w, scheme),
					linkRequests: s.linkRequests,
					consumptions: s.consumptions,
				})
			}
		}
	}
	return rows
}

// reportedRow is one JSON line the script writes to stdout. The two
// results are pointers so a line that omits one is told apart from a
// line reporting false.
type reportedRow struct {
	ID       string `json:"id"`
	Fixture  *bool  `json:"fixture"`
	Property *bool  `json:"property"`
	Detail   string `json:"detail"`
}

// A loggedRequest is one request the fixture server received, as the
// script reads it back from the request log.
type loggedRequest struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Query   string      `json:"query"`
	Headers http.Header `json:"headers"`
}

// requestLog records every request reaching the fixture server's
// listener, including any outside the owned prefix, and serves the
// record as JSON. A browser fetches some resources (a favicon, for
// one) without Playwright's request events ever seeing them, so the
// server's own record is the only complete one. The shell itself
// gains no route for it: the log is served from a second listener.
type requestLog struct {
	mu      sync.Mutex
	records []loggedRequest
}

func (l *requestLog) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.mu.Lock()
		l.records = append(l.records, loggedRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   r.URL.RawQuery,
			Headers: r.Header.Clone(),
		})
		l.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (l *requestLog) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	l.mu.Lock()
	body, err := json.Marshal(l.records)
	l.mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

// snapshot returns a copy of every record l holds so far.
func (l *requestLog) snapshot() []loggedRequest {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.records)
}

// The fixture roster's own addresses and display names, and the site
// name the login heading and title rows read back. The site name and
// the admin's display name each carry `"`, `<`, or `&` (universal at
// `example.test` scope), so a page that rendered either with
// `{@html}` would leave a stray element behind rather than the
// literal text a rendered row asserts.
const (
	renderedSiteName         = `Fixture "Site" <Name> & Co`
	renderedAdminAddress     = "admin@example.test"
	renderedAdminDisplayName = "Fixture <Admin>"
	renderedMemberAddress    = "member@example.test"

	// renderedNewMemberAddress is the address the add-member row writes:
	// on no roster at the run's start, so `duplicate_address` on the
	// second submit proves the first write landed.
	renderedNewMemberAddress = "new-member@example.test"
)

// renderedEnviron is the fixture environment: a temporary state
// directory, a roster of one admin and one member at example.test, the
// fixed test secrets, publicURL as the server's own origin, an SMTP port
// nothing listens on, the kind unknown-address answer, the non-default
// prefix, and the fixture's four rate limits.
func renderedEnviron(t *testing.T, publicURL string) []string {
	t.Helper()
	rosterPath := filepath.Join(t.TempDir(), "roster.json")
	roster := `{"schema_version":1,"members":[` +
		`{"address":"` + renderedAdminAddress + `","display_name":"` + renderedAdminDisplayName + `","admin":true},` +
		`{"address":"` + renderedMemberAddress + `","display_name":"Fixture Member"}]}`
	if err := os.WriteFile(rosterPath, []byte(roster), 0o600); err != nil {
		t.Fatalf("write fixture roster: %v", err)
	}
	environ := serveAuthEnviron(t)
	for name, value := range map[string]string{
		"DUBPLATE_ROSTER_FILE":                                   rosterPath,
		"DUBPLATE_PUBLIC_URL":                                    publicURL,
		"DUBPLATE_SMTP_HOST":                                     "127.0.0.1",
		"DUBPLATE_SMTP_PORT":                                     strconv.Itoa(closedPort(t)),
		"DUBPLATE_UNKNOWN_ADDRESS_ANSWER":                        "kind",
		"DUBPLATE_SERVE_PREFIX":                                  renderedPrefix,
		"DUBPLATE_SITE_NAME":                                     renderedSiteName,
		"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR":            strconv.Itoa(renderedLinkRequestsPerAddress),
		"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR":             strconv.Itoa(renderedLinkRequestsPerClient),
		"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR": strconv.Itoa(renderedConsumptionAttemptsPerClient),
		"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR":                    strconv.Itoa(renderedLinkMailCeiling),
		"DUBPLATE_MAGIC_LINK_EXPIRY":                             renderedMagicLinkExpiry.String(),
	} {
		environ = replaceKey(environ, name, value)
	}
	return environ
}

// closedPort returns a loopback TCP port nothing listens on: one the
// kernel allocated and this test released.
func closedPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate a closed port: %v", err)
	}
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("a tcp listener's address is a %T", ln.Addr())
	}
	if err := ln.Close(); err != nil {
		t.Fatalf("release the closed port: %v", err)
	}
	return addr.Port
}

// startRenderedServer builds the real serve shell through
// [newServeShell] on the fixture environment and serves it on a
// loopback port, every request recorded in the returned log. It
// returns the environment too, so a caller reads the configuration
// the shell was actually built from.
func startRenderedServer(t *testing.T) (*httptest.Server, *requestLog, []string) {
	t.Helper()
	srv := httptest.NewUnstartedServer(nil)
	environ := renderedEnviron(t, "http://"+srv.Listener.Addr().String())
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err != nil {
		t.Fatalf("newServeShell: %v", err)
	}
	log := &requestLog{}
	srv.Config.Handler = log.wrap(shell.Handler())
	srv.Start()
	t.Cleanup(srv.Close)
	return srv, log, environ
}

// startFixtureServer builds the real serve shell through
// [newServeShell] on renderedEnviron with overrides applied over it,
// and serves it on its own loopback port. It is the harness's own
// mechanism, reused for G7's branding and operator fixtures
// (R-5ciii8), which need a differently configured server than
// [startRenderedServer]'s.
func startFixtureServer(t *testing.T, overrides map[string]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(nil)
	environ := renderedEnviron(t, "http://"+srv.Listener.Addr().String())
	for name, value := range overrides {
		environ = replaceKey(environ, name, value)
	}
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err != nil {
		t.Fatalf("newServeShell: %v", err)
	}
	srv.Config.Handler = shell.Handler()
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

// startBrandingServer serves the fixture environment with
// DUBPLATE_BRANDING declaring renderedBrandingPrimary: G7's
// with-branding fixture.
func startBrandingServer(t *testing.T) *httptest.Server {
	t.Helper()
	return startFixtureServer(t, map[string]string{
		"DUBPLATE_BRANDING": "--color-primary: " + renderedBrandingPrimary,
	})
}

// startOperatorServer serves the fixture environment with the same
// branding as [startBrandingServer] plus an operator CSS file
// declaring renderedOperatorPrimary for the same name: G7's operator
// fixture, which proves tier 3 wins over tier 2 (R4.1).
func startOperatorServer(t *testing.T) *httptest.Server {
	t.Helper()
	path := writeOperatorCSS(t, fmt.Appendf(nil, ":root {\n\t--color-primary: %s;\n}\n", renderedOperatorPrimary))
	return startFixtureServer(t, map[string]string{
		"DUBPLATE_BRANDING":          "--color-primary: " + renderedBrandingPrimary,
		"DUBPLATE_OPERATOR_CSS_FILE": path,
	})
}

// checkRateBudget fails t unless the manifest's total link requests
// and consumptions stay below the limits the shell resolved from
// environ. Every row's requests are counted against the per-address
// limit, the case where they all name one address.
func checkRateBudget(t *testing.T, environ []string) {
	t.Helper()
	cfg, err := resolveAuthConfig(environ)
	if err != nil {
		t.Fatalf("resolveAuthConfig: %v", err)
	}
	var requests, consumptions int
	for _, row := range renderedManifest {
		requests += row.linkRequests
		consumptions += row.consumptions
	}
	t.Logf("rate budget: %d link requests (limits: %d per address, %d per client, %d mail), %d consumptions (limit %d per client)",
		requests, cfg.linkRequestsPerAddress, cfg.linkRequestsPerClient, cfg.linkMailCeiling,
		consumptions, cfg.consumptionAttemptsPerClient)
	for _, limit := range []struct {
		name         string
		spend, limit int
	}{
		{"link requests per address", requests, cfg.linkRequestsPerAddress},
		{"link requests per client", requests, cfg.linkRequestsPerClient},
		{"link mail", requests, cfg.linkMailCeiling},
		{"consumption attempts per client", consumptions, cfg.consumptionAttemptsPerClient},
	} {
		if limit.spend >= limit.limit {
			t.Fatalf("the manifest spends %d %s, at or over the fixture's limit of %d", limit.spend, limit.name, limit.limit)
		}
	}
}

// mintRenderedLink mints one magic link for address against environ's
// own link secret and expiry (`auth.MintLink`, `auth.LinkURL`), so the
// rendered leg's ordering and `signed_in` rows spend real, single-use
// links rather than a value neither `auth.VerifyLink` nor the roster
// would accept.
func mintRenderedLink(t *testing.T, environ []string, address string) string {
	t.Helper()
	link, err := mintLinkFor(environ, address)
	if err != nil {
		t.Fatalf("mintLinkFor: %v", err)
	}
	return link
}

// mintLinkFor mints one magic link for address against environ's own
// link secret and expiry (`auth.MintLink`, `auth.LinkURL`): the shared
// mechanism behind both [mintRenderedLink] and [mintServer].
func mintLinkFor(environ []string, address string) (string, error) {
	cfg, err := resolveAuthConfig(environ)
	if err != nil {
		return "", fmt.Errorf("resolveAuthConfig: %w", err)
	}
	secret := []byte(environValue(environ, "DUBPLATE_MAGIC_LINK_SECRET"))
	token, _, err := auth.MintLink(secret, address, time.Now(), cfg.magicLinkExpiry)
	if err != nil {
		return "", fmt.Errorf("auth.MintLink: %w", err)
	}
	link, err := auth.LinkURL(environValue(environ, "DUBPLATE_PUBLIC_URL"), renderedPrefix, token)
	if err != nil {
		return "", fmt.Errorf("auth.LinkURL: %w", err)
	}
	return link, nil
}

// mintRequest is the capture matrix's own request shape to [mintServer]:
// which fixture address to mint a fresh, single-use link for.
type mintRequest struct {
	Kind string `json:"kind"`
}

// mintResponse is [mintServer]'s answer: one freshly minted link.
type mintResponse struct {
	URL string `json:"url"`
}

// requirePOST answers 405 and reports false unless r is a POST, the
// shared guard [mintServer] and [rosterResetServer] open their
// ServeHTTP with.
func requirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// mintServer answers a freshly minted magic link for the fixture admin
// or member address, one per POST, through [mintLinkFor]: the capture
// matrix's on-demand source for the sixty links its six signed-in
// states spend (one per cell, R-5ciii4's "one minted link per
// capture"), rather than sixty environment variables. It is its own
// loopback listener, mounting no path on the fixture server's own
// handler, so it adds no bypass to any production path.
type mintServer struct {
	environ []string
}

func (m *mintServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}
	var req mintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var address string
	switch req.Kind {
	case "admin":
		address = renderedAdminAddress
	case "member":
		address = renderedMemberAddress
	default:
		http.Error(w, "unknown address kind "+req.Kind, http.StatusBadRequest)
		return
	}
	link, err := mintLinkFor(m.environ, address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(mintResponse{URL: link})
}

// rosterResetServer writes the fixture roster back to a baseline on
// every POST, so each capture cell starts from the same member list and
// every cell of one members state renders the same count. The baseline
// is the file as the first POST finds it: the capture matrix runs after
// every other row, so it keeps the members those rows added, and the
// members/roster-file check still finds the add-member row's write.
// Each POST answers {"members": n}, the baseline's member count. The
// write goes through atomicfile, so the file is replaced under a new
// inode and the server's cached roster is re-read on its next request.
// Like [mintServer], it is its own loopback listener.
type rosterResetServer struct {
	path string

	mu       sync.Mutex
	baseline []byte
	members  int
}

func (s *rosterResetServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.baseline == nil {
		data, err := os.ReadFile(s.path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var roster struct {
			Members []json.RawMessage `json:"members"`
		}
		if err := json.Unmarshal(data, &roster); err != nil {
			http.Error(w, "parse the baseline roster: "+err.Error(), http.StatusInternalServerError)
			return
		}
		s.baseline, s.members = data, len(roster.Members)
	} else if err := atomicfile.WriteFile(s.path, s.baseline, 0o600); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Members int `json:"members"`
	}{s.members})
}

// countLogged reports how many records in records are a POST to a
// path ending in suffix (the owned prefix's own subtree name).
func countLogged(records []loggedRequest, suffix string) int {
	n := 0
	for _, r := range records {
		if r.Method == http.MethodPost && strings.HasSuffix(r.Path, suffix) {
			n++
		}
	}
	return n
}

// checkSpendReconciliation fails t, naming both counts, unless the
// manifest's declared link-request and consumption totals equal what
// the fixture server's own request log recorded, and unless both
// totals are nonzero: the state table's real rows always spend both,
// so a manifest whose counts were dropped cannot pass by also logging
// nothing.
func checkSpendReconciliation(t *testing.T, log *requestLog) {
	t.Helper()
	var declaredRequests, declaredConsumptions int
	for _, row := range renderedManifest {
		declaredRequests += row.linkRequests
		declaredConsumptions += row.consumptions
	}
	records := log.snapshot()
	loggedRequests := countLogged(records, "/request-link")
	loggedConsumptions := countLogged(records, "/consume-link")
	if declaredRequests == 0 {
		t.Fatalf("the manifest declares 0 link requests; the live rows always spend at least one")
	}
	if declaredConsumptions == 0 {
		t.Fatalf("the manifest declares 0 consumptions; the live rows always spend at least one")
	}
	if declaredRequests != loggedRequests {
		t.Fatalf("the manifest declares %d request-link POSTs, the log shows %d", declaredRequests, loggedRequests)
	}
	if declaredConsumptions != loggedConsumptions {
		t.Fatalf("the manifest declares %d consume-link POSTs, the log shows %d", declaredConsumptions, loggedConsumptions)
	}
	t.Logf("spend reconciliation: %d request-link POSTs, %d consume-link POSTs, both declared and logged", loggedRequests, loggedConsumptions)
}

// checkServedVersion fails t unless the server's _app/version.json is
// byte-identical to the one on disk under webDir's build, so a binary
// embedding a stale bundle cannot pass the leg.
func checkServedVersion(t *testing.T, base, webDir string) {
	t.Helper()
	built, err := os.ReadFile(filepath.Join(webDir, "build", "_app", "version.json"))
	if err != nil {
		t.Fatalf("read the built version.json: %v", err)
	}
	resp, err := http.Get(base + "_app/version.json")
	if err != nil {
		t.Fatalf("the server does not answer: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	served, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read the served version.json: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("the served version.json answered %d", resp.StatusCode)
	}
	if !bytes.Equal(served, built) {
		t.Fatalf("the served version.json %q differs from the build's %q; the binary embeds a stale bundle", served, built)
	}
}

// scriptEnvNames are the only parent variables the browser script
// inherits: what Node and Playwright need to find themselves, their
// browser cache, and a temporary directory. The test's own
// environment can hold workstation secrets, and nothing in the
// script needs them.
var scriptEnvNames = []string{
	"PATH", "HOME", "TMPDIR", "LANG", "LC_ALL", "TZ",
	"XDG_CACHE_HOME", "XDG_CONFIG_HOME", "XDG_RUNTIME_DIR",
	"PLAYWRIGHT_BROWSERS_PATH",
}

// scriptEnviron is the parent environment filtered to scriptEnvNames.
func scriptEnviron() []string {
	var env []string
	for _, name := range scriptEnvNames {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return env
}

// runRenderedScript runs the browser script under webDir with env,
// returning its stdout and stderr and its exit error.
func runRenderedScript(ctx context.Context, webDir string, env []string) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, renderedScriptTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "node", "scripts/rendered/run.ts")
	cmd.Dir = webDir
	cmd.Env = env
	cmd.WaitDelay = 10 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// parseRows decodes the script's stdout, one row per line. A line that
// is not a row object, a row with no ID, and a second report of one ID
// are each a problem.
func parseRows(stdout []byte) (map[string]reportedRow, []string) {
	rows := make(map[string]reportedRow)
	var problems []string
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.DisallowUnknownFields()
		var row reportedRow
		if err := dec.Decode(&row); err != nil {
			problems = append(problems, "the script wrote a line that is not a row: "+line)
			continue
		}
		if row.ID == "" {
			problems = append(problems, "the script reported a row with no ID: "+line)
			continue
		}
		if _, dup := rows[row.ID]; dup {
			problems = append(problems, "the script reported row "+row.ID+" twice")
			continue
		}
		rows[row.ID] = row
	}
	if err := sc.Err(); err != nil {
		problems = append(problems, "read the script's stdout: "+err.Error())
	}
	return rows, problems
}

// TestRendered is the rendered leg: the real serve shell on a loopback
// port, and Playwright's own Chromium driven against it by
// web/scripts/rendered/run.ts. It runs only under the rendered build
// tag, from the Makefile's web-rendered target, which also fails
// unless this test's own pass event appears in the go test -json
// stream.
func TestRendered(t *testing.T) {
	if renderedScriptTimeout >= renderedMagicLinkExpiry {
		t.Fatalf("renderedScriptTimeout %v is not below renderedMagicLinkExpiry %v; a late row would meet an expired link", renderedScriptTimeout, renderedMagicLinkExpiry)
	}
	webDir, err := filepath.Abs(filepath.Join("..", "..", "web"))
	if err != nil {
		t.Fatalf("resolve web/: %v", err)
	}
	srv, log, environ := startRenderedServer(t)
	logSrv := httptest.NewServer(log)
	t.Cleanup(logSrv.Close)

	base := srv.URL + renderedPrefix + "/"
	checkServedVersion(t, base, webDir)
	checkRateBudget(t, environ)

	brandingSrv := startBrandingServer(t)
	operatorSrv := startOperatorServer(t)

	// The token-ordering row, its fragment variant, and the
	// state-table's own `signed_in` row each spend one real, single-use
	// link, so each gets its own: reusing one would leave a later row
	// consuming an already-spent link.
	loginLinkOrdering := mintRenderedLink(t, environ, renderedAdminAddress)
	loginLinkOrderingFragment := mintRenderedLink(t, environ, renderedAdminAddress)
	loginLinkSignedIn := mintRenderedLink(t, environ, renderedAdminAddress)

	// The add-member row signs the fixture admin in once; the non-admin
	// row signs the fixture member in once. Each spends its own minted
	// link, single-use like every other consumption row's.
	membersAddLink := mintRenderedLink(t, environ, renderedAdminAddress)
	membersNonAdminLink := mintRenderedLink(t, environ, renderedMemberAddress)

	// The header's own rows sign in once and reuse the session cookie
	// across every context they open, so this is the only link they
	// spend.
	membersHeaderLink := mintRenderedLink(t, environ, renderedAdminAddress)

	// The capture matrix's own "ready" state is never consumed (its
	// screenshot shows the interstitial before the click), so its ten
	// cells share one minted link.
	captureInterstitialReadyLink := mintRenderedLink(t, environ, renderedAdminAddress)

	mintSrv := httptest.NewServer(&mintServer{environ: environ})
	t.Cleanup(mintSrv.Close)
	resetSrv := httptest.NewServer(&rosterResetServer{path: environValue(environ, "DUBPLATE_ROSTER_FILE")})
	t.Cleanup(resetSrv.Close)

	env := append(scriptEnviron(),
		"DUBPLATE_RENDERED_BASE_URL="+base,
		"DUBPLATE_RENDERED_REQUEST_LOG="+logSrv.URL,
		"DUBPLATE_RENDERED_BRANDING_BASE_URL="+brandingSrv.URL+renderedPrefix+"/",
		"DUBPLATE_RENDERED_OPERATOR_BASE_URL="+operatorSrv.URL+renderedPrefix+"/",
		"DUBPLATE_RENDERED_BRANDING_PRIMARY="+renderedBrandingPrimary,
		"DUBPLATE_RENDERED_OPERATOR_PRIMARY="+renderedOperatorPrimary,
		"DUBPLATE_RENDERED_SITE_NAME="+renderedSiteName,
		"DUBPLATE_RENDERED_ADMIN_DISPLAY_NAME="+renderedAdminDisplayName,
		"DUBPLATE_RENDERED_LOGIN_LINK_ORDERING="+loginLinkOrdering,
		"DUBPLATE_RENDERED_LOGIN_LINK_ORDERING_FRAGMENT="+loginLinkOrderingFragment,
		"DUBPLATE_RENDERED_LOGIN_LINK_SIGNED_IN="+loginLinkSignedIn,
		"DUBPLATE_RENDERED_MEMBERS_ADD_LINK="+membersAddLink,
		"DUBPLATE_RENDERED_MEMBERS_NON_ADMIN_LINK="+membersNonAdminLink,
		"DUBPLATE_RENDERED_MEMBERS_HEADER_LINK="+membersHeaderLink,
		"DUBPLATE_RENDERED_CAPTURE_MINT_URL="+mintSrv.URL,
		"DUBPLATE_RENDERED_CAPTURE_INTERSTITIAL_READY_LINK="+captureInterstitialReadyLink,
		"DUBPLATE_RENDERED_CAPTURE_ROSTER_RESET_URL="+resetSrv.URL,
	)
	stdout, stderr, err := runRenderedScript(t.Context(), webDir, env)
	if err != nil {
		t.Errorf("the browser script exited: %v\n%s", err, stderr)
	}
	checkSpendReconciliation(t, log)

	rows, problems := parseRows(stdout)
	for _, p := range problems {
		t.Error(p)
	}
	manifest := make(map[string]bool, len(renderedManifest))
	for _, row := range renderedManifest {
		manifest[row.id] = true
	}
	for id := range rows {
		if !manifest[id] {
			t.Errorf("the script reported row %q, which the manifest does not hold", id)
		}
	}
	t.Logf("manifest: %d rows", len(renderedManifest))

	for _, want := range renderedManifest {
		t.Run(want.id, func(t *testing.T) {
			row, ok := rows[want.id]
			switch {
			case !ok:
				t.Fatalf("the script never reported row %q", want.id)
			case row.Fixture == nil:
				t.Fatalf("row %q was reported without its fixture-state result: %s", want.id, row.Detail)
			case !*row.Fixture:
				t.Fatalf("row %q could not stand up its fixture state: %s", want.id, row.Detail)
			case row.Property == nil || !*row.Property:
				t.Fatalf("row %q failed: %s", want.id, row.Detail)
			}
			t.Log(row.Detail)
		})
	}

	// The add-member row's write reaches the fixture roster file on
	// disk, proving the POST the script drove was the real roster API,
	// not a fulfilled route.
	t.Run("members/roster-file", func(t *testing.T) {
		rosterPath := environValue(environ, "DUBPLATE_ROSTER_FILE")
		data, err := os.ReadFile(rosterPath)
		if err != nil {
			t.Fatalf("read the fixture roster file: %v", err)
		}
		if !bytes.Contains(data, []byte(renderedNewMemberAddress)) {
			t.Fatalf("the fixture roster file does not carry the added address %q: %s", renderedNewMemberAddress, data)
		}
	})

	// A browser missing from Playwright's cache must fail the leg with
	// the install hint, never skip it: an empty browsers path takes the
	// same registry lookup a cleared cache or a Playwright bump takes.
	t.Run("refusal/missing-browser", func(t *testing.T) {
		missing := slices.Concat(env, []string{"PLAYWRIGHT_BROWSERS_PATH=" + t.TempDir()})
		_, stderr, err := runRenderedScript(t.Context(), webDir, missing)
		if exitErr, ok := errors.AsType[*exec.ExitError](err); !ok || exitErr.ExitCode() == 0 {
			t.Fatalf("with no browser installed the script exited %v, want a nonzero exit", err)
		}
		if !strings.Contains(string(stderr), renderedInstallHint) {
			t.Fatalf("with no browser installed the script's stderr does not name %q:\n%s", renderedInstallHint, stderr)
		}
	})
}
