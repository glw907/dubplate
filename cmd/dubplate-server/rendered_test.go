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
)

// renderedPrefix is the owned prefix the fixture server mounts under.
// It differs from the default so every rendered row also proves the
// bundle resolves its URLs from the runtime prefix.
const renderedPrefix = "/rendered"

// renderedScriptTimeout bounds one run of the browser script, at
// about three times its measured wall time (about 50s once the
// component rows landed), since a heavy gate shares the machine.
const renderedScriptTimeout = 150 * time.Second

// renderedInstallHint is the command the script names when
// Playwright's Chromium is not installed.
const renderedInstallHint = "npx --prefix web playwright install chromium"

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
// row reported without its fixture-state result each fail.
var renderedManifest = []renderedRow{
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

// renderedEnviron is the fixture environment: a temporary state
// directory, a roster of one admin and one member at example.test, the
// fixed test secrets, publicURL as the server's own origin, an SMTP port
// nothing listens on, the kind unknown-address answer, the non-default
// prefix, and the fixture's four rate limits.
func renderedEnviron(t *testing.T, publicURL string) []string {
	t.Helper()
	rosterPath := filepath.Join(t.TempDir(), "roster.json")
	roster := `{"schema_version":1,"members":[` +
		`{"address":"admin@example.test","display_name":"Fixture Admin","admin":true},` +
		`{"address":"member@example.test","display_name":"Fixture Member"}]}`
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
		"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR":            strconv.Itoa(renderedLinkRequestsPerAddress),
		"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR":             strconv.Itoa(renderedLinkRequestsPerClient),
		"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR": strconv.Itoa(renderedConsumptionAttemptsPerClient),
		"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR":                    strconv.Itoa(renderedLinkMailCeiling),
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
	shell, _, err := buildServeShellWithLogger(t, environ, discardTestLogger())
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
	shell, _, err := buildServeShellWithLogger(t, environ, discardTestLogger())
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

	env := append(scriptEnviron(),
		"DUBPLATE_RENDERED_BASE_URL="+base,
		"DUBPLATE_RENDERED_REQUEST_LOG="+logSrv.URL,
		"DUBPLATE_RENDERED_BRANDING_BASE_URL="+brandingSrv.URL+renderedPrefix+"/",
		"DUBPLATE_RENDERED_OPERATOR_BASE_URL="+operatorSrv.URL+renderedPrefix+"/",
		"DUBPLATE_RENDERED_BRANDING_PRIMARY="+renderedBrandingPrimary,
		"DUBPLATE_RENDERED_OPERATOR_PRIMARY="+renderedOperatorPrimary,
	)
	stdout, stderr, err := runRenderedScript(t.Context(), webDir, env)
	if err != nil {
		t.Errorf("the browser script exited: %v\n%s", err, stderr)
	}

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
