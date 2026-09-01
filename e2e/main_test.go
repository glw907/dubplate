// Package e2e end-to-end tests the dubplate binary: it builds the real
// compiled binary plus stub beet, rclone, and sqlite3 programs once in
// TestMain, then execs the binary against a fixture inbox for each test,
// asserting on filed outcomes, mail delivered to a local test SMTP
// server, and ping requests captured by a local test HTTP server.
package e2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mhale/smtpd"
)

// binPath is the compiled dubplate binary, built once in TestMain.
// stubBinDir holds the compiled beet, rclone, and sqlite3 stand-ins,
// placed first on every test's PATH.
var (
	binPath       string
	stubBinDir    string
	repoRoot      string
	smtpCACert    *tls.Certificate
	smtpCACertPEM string // path to a PEM file holding the test CA, for SSL_CERT_FILE
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Fprintln(os.Stderr, "e2e setup:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func setup() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	repoRoot = filepath.Dir(cwd) // e2e/ sits directly under the repo root

	work, err := os.MkdirTemp("", "dubplate-e2e-*")
	if err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}

	binPath = filepath.Join(work, "dubplate")
	if err := buildGo(repoRoot, "./cmd/dubplate", binPath); err != nil {
		return fmt.Errorf("build dubplate: %w", err)
	}

	stubBinDir = filepath.Join(work, "stub-bin")
	if err := os.MkdirAll(stubBinDir, 0o755); err != nil {
		return fmt.Errorf("create stub bin dir: %w", err)
	}
	for _, name := range []string{"beet", "rclone", "sqlite3"} {
		src := "./testdata/stubs/" + name
		dst := filepath.Join(stubBinDir, name)
		if err := buildGo(cwd, src, dst); err != nil {
			return fmt.Errorf("build stub %s: %w", name, err)
		}
	}

	cert, certPEM, err := generateSelfSignedCert()
	if err != nil {
		return fmt.Errorf("generate test TLS cert: %w", err)
	}
	smtpCACert = &cert
	smtpCACertPEM = filepath.Join(work, "smtp-test-ca.pem")
	if err := os.WriteFile(smtpCACertPEM, certPEM, 0o644); err != nil {
		return fmt.Errorf("write test CA: %w", err)
	}

	flacBytes, err := generateGoodFlac(work)
	if err != nil {
		return fmt.Errorf("generate fixture FLAC: %w", err)
	}
	goodFlacBytes = flacBytes

	return nil
}

// generateGoodFlac synthesizes one real, valid 44.1kHz/16-bit FLAC via
// ffmpeg and flac, exactly like the retired shell layer's
// tests/music-import.bats' make_flac helper: the quality gate runs a real
// `flac -t` and `metaflac` probe over every staged track, so a fixture
// album that must pass the gate needs a file that actually does.
func generateGoodFlac(workDir string) ([]byte, error) {
	wav := filepath.Join(workDir, "gen.wav")
	flacPath := filepath.Join(workDir, "gen.flac")

	if out, err := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-ar", "44100", "-sample_fmt", "s16", wav).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w: %s", err, out)
	}
	if out, err := exec.Command("flac", "--best", "-f", "-o", flacPath, wav).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("flac: %w: %s", err, out)
	}
	return os.ReadFile(flacPath)
}

// goodFlacBytes holds one valid 44.1kHz/16-bit FLAC's bytes, encoded
// once in TestMain and reused by every test that needs an album the
// quality gate accepts.
var goodFlacBytes []byte

// writeGoodFlac writes goodFlacBytes to path, creating parent
// directories as needed.
func writeGoodFlac(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, goodFlacBytes, 0o644); err != nil {
		t.Fatalf("write flac %s: %v", path, err)
	}
}

// --- SMTP, ping, and Navidrome test servers -----------------------------

type smtpDelivery struct {
	from string
	to   []string
	data []byte
}

// startSMTPServer starts a local SMTP test server offering STARTTLS with
// the shared self-signed cert, matching the box's own 587/STARTTLS
// route. The compiled dubplate binary trusts this cert via SSL_CERT_FILE
// (set by [testEnv.buildEnv]), not by any in-process pool this package
// could otherwise reach into.
func startSMTPServer(t *testing.T) (addr string, deliveries chan smtpDelivery) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	deliveries = make(chan smtpDelivery, 32)
	srv := &smtpd.Server{
		Hostname:     "localhost",
		Appname:      "dubplate-e2e",
		TLSConfig:    &tls.Config{Certificates: []tls.Certificate{*smtpCACert}},
		AuthRequired: true,
		AuthHandler: func(_ net.Addr, _ string, _, _, _ []byte) (bool, error) {
			return true, nil
		},
		Handler: func(_ net.Addr, from string, to []string, data []byte) error {
			deliveries <- smtpDelivery{from: from, to: to, data: data}
			return nil
		},
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = ln.Close() })
	return ln.Addr().String(), deliveries
}

type pingHit struct {
	method string
	path   string
}

// startPingServer starts a local HTTP server standing in for a pings
// provider, recording every request's method and path.
func startPingServer(t *testing.T) (url string, hits chan pingHit) {
	t.Helper()
	hits = make(chan pingHit, 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits <- pingHit{method: r.Method, path: r.URL.Path}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server.URL, hits
}

// startNavidromeServer starts a local HTTP server standing in for
// Navidrome's Subsonic API, recording every scan-trigger request.
func startNavidromeServer(t *testing.T) (url string, hits chan *http.Request) {
	t.Helper()
	hits = make(chan *http.Request, 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits <- r
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"subsonic-response":{"status":"ok"}}`))
	}))
	t.Cleanup(server.Close)
	return server.URL, hits
}

func smtpHostPort(addr string) (host, port string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic("split SMTP test server address: " + err.Error())
	}
	return host, port
}

// --- testEnv: one fixture estate per test -------------------------------

// defaultContributorsYAML matches every e2e fixture's contributors.yaml
// unless a test overrides testEnv.contributorsFile's content directly:
// two known contributors, alice and bob, mirroring the bats suite's own
// fixture.
const defaultContributorsYAML = `contributors:
  - name: alice
    email: alice@example.com
  - name: bob
    email: bob@example.com
`

// testEnv is one test's fixture estate: its own inbox, staging, review,
// quarantine, and state directories, contributors roster, and stand-in
// SMTP/ping/Navidrome servers, wired into the env vars [testEnv.run]
// passes to the compiled dubplate binary.
type testEnv struct {
	t *testing.T

	root, inboxDir, stagingDir, stateDir, reviewDir, quarantineDir string
	beetsDB, contributorsFile                                      string

	smtpDeliveries chan smtpDelivery
	pingHits       chan pingHit
	navidromeHits  chan *http.Request

	// env holds every DUBPLATE_* (and STUB_*) variable this env's run
	// passes to the binary. A test mutates it directly before calling
	// run to override a threshold or point STUB_BEET_SKIP at specific
	// paths.
	env map[string]string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	root := t.TempDir()

	e := &testEnv{
		t:                t,
		root:             root,
		inboxDir:         filepath.Join(root, "inbox"),
		stagingDir:       filepath.Join(root, "staging"),
		stateDir:         filepath.Join(root, "state"),
		reviewDir:        filepath.Join(root, "review"),
		quarantineDir:    filepath.Join(root, "quarantine-failed"),
		beetsDB:          filepath.Join(root, "state", "library.db"),
		contributorsFile: filepath.Join(root, "contributors.yaml"),
	}

	for _, dir := range []string{e.inboxDir, e.stagingDir, e.stateDir, e.reviewDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(e.contributorsFile, []byte(defaultContributorsYAML), 0o644); err != nil {
		t.Fatalf("write contributors.yaml: %v", err)
	}

	smtpAddr, smtpDeliveries := startSMTPServer(t)
	e.smtpDeliveries = smtpDeliveries
	smtpHost, smtpPort := smtpHostPort(smtpAddr)

	pingURL, pingHits := startPingServer(t)
	e.pingHits = pingHits

	navidromeURL, navidromeHits := startNavidromeServer(t)
	e.navidromeHits = navidromeHits

	e.env = map[string]string{
		"DUBPLATE_ADMIN_EMAIL":              "admin@example.com",
		"DUBPLATE_SMTP_HOST":                smtpHost,
		"DUBPLATE_SMTP_PORT":                smtpPort,
		"DUBPLATE_SMTP_USER":                "dubplate",
		"DUBPLATE_SMTP_PASSWORD":            "s3cret",
		"DUBPLATE_SMTP_FROM":                "dubplate@example.com",
		"DUBPLATE_R2_ACCOUNT_ID":            "acct",
		"DUBPLATE_R2_BUCKET":                "bucket",
		"DUBPLATE_R2_REMOTE":                "r2",
		"DUBPLATE_R2_ACCESS_KEY_ID":         "key",
		"DUBPLATE_R2_SECRET_ACCESS_KEY":     "secret",
		"DUBPLATE_NAVIDROME_URL":            navidromeURL,
		"DUBPLATE_NAVIDROME_ADMIN_USER":     "admin",
		"DUBPLATE_NAVIDROME_ADMIN_PASSWORD": "adminpass",
		"DUBPLATE_CONTRIBUTORS_FILE":        e.contributorsFile,
		"DUBPLATE_REJECT_EMAIL_TEMPLATE":    filepath.Join(repoRoot, "config", "reject-email.txt"),
		"DUBPLATE_BEETS_CONFIG":             filepath.Join(root, "beets-config.yaml"),
		"DUBPLATE_BEETS_DB":                 e.beetsDB,
		"DUBPLATE_NAVIDROME_DB":             filepath.Join(root, "navidrome.db"),
		"DUBPLATE_LIBRARY_DIR":              filepath.Join(root, "library"),
		"DUBPLATE_INBOX_DIR":                e.inboxDir,
		"DUBPLATE_STAGING_DIR":              e.stagingDir,
		"DUBPLATE_STATE_DIR":                e.stateDir,
		"DUBPLATE_TRIGGER_DIR":              filepath.Join(root, "run"),
		"DUBPLATE_REVIEW_DIR":               e.reviewDir,
		"DUBPLATE_QUARANTINE_DIR":           e.quarantineDir,
		"DUBPLATE_BACKUP_STAGING":           filepath.Join(root, "backup-staging"),
		"DUBPLATE_UPLOADS_DIR":              filepath.Join(root, "uploads"),
		"DUBPLATE_SERVE_BIND_ADDRS":         "127.0.0.1",
		"DUBPLATE_SERVE_PORT":               "8080",
		"DUBPLATE_ACCESS_TEAM_DOMAIN":       "test.cloudflareaccess.com",
		"DUBPLATE_ACCESS_AUD":               "test-aud",
		"DUBPLATE_PUBLIC_NAVIDROME_URL":     navidromeURL,
		"DUBPLATE_SERVE_SITE_NAME":          "dubplate",
		"DUBPLATE_FREE_FLOOR_GB":            "0",
		"DUBPLATE_QUIESCENT_MINUTES":        "10",
		"DUBPLATE_PING_IMPORT_URL":          pingURL,
		"DUBPLATE_PING_BACKUP_URL":          pingURL,
		"STUB_BEETS_DB":                     e.beetsDB,
	}

	return e
}

// run execs the compiled dubplate binary with args against e's fixture
// estate and returns its combined stdout/stderr and exit code.
func (e *testEnv) run(args ...string) (output string, exitCode int) {
	e.t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = e.buildEnv()

	out, err := cmd.CombinedOutput()
	output = string(out)
	if err == nil {
		return output, 0
	}
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return output, exitErr.ExitCode()
	}
	e.t.Fatalf("run dubplate %v: %v (output: %s)", args, err, out)
	return output, -1
}

// startAsync starts the compiled dubplate binary with args against e's
// fixture estate without waiting for it to finish, for a test that needs
// to observe another dubplate invocation while this one is still
// running (a `dubplate beet` holding the import lock).
func (e *testEnv) startAsync(args ...string) *exec.Cmd {
	e.t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = e.buildEnv()
	if err := cmd.Start(); err != nil {
		e.t.Fatalf("start dubplate %v: %v", args, err)
	}
	return cmd
}

func (e *testEnv) buildEnv() []string {
	env := []string{
		"PATH=" + stubBinDir + ":" + os.Getenv("PATH"),
		"SSL_CERT_FILE=" + smtpCACertPEM,
	}
	for k, v := range e.env {
		env = append(env, k+"="+v)
	}
	return env
}

// drainMail collects every message the test SMTP server has received,
// waiting up to 300ms after the last one for a further delivery. The
// binary's own Send calls are synchronous within the subprocess, which
// [testEnv.run] has already waited for by the time a test calls this, so
// the wait is only a safety margin, not the primary synchronization.
func (e *testEnv) drainMail() []smtpDelivery {
	var got []smtpDelivery
	for {
		select {
		case d := <-e.smtpDeliveries:
			got = append(got, d)
		case <-time.After(300 * time.Millisecond):
			return got
		}
	}
}

// writeUploadedFlac stages a single-track album ready for the inbox:
// inboxDir/<user>/<album>/track1.flac, aged well past the quiescence
// window.
func (e *testEnv) writeUploadedFlac(user, album string) {
	dir := filepath.Join(e.inboxDir, user, album)
	writeGoodFlac(e.t, filepath.Join(dir, "track1.flac"))
	e.ageInbox(user)
}

// ageInbox sets every file under the given contributor's inbox directory
// to a modification time well past the quiescence window, so this run
// picks it up -- matching the bats suite's touch_old helper.
func (e *testEnv) ageInbox(user string) {
	e.t.Helper()
	old := time.Now().Add(-30 * time.Minute)
	dir := filepath.Join(e.inboxDir, user)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return os.Chtimes(path, old, old)
	})
	if err != nil {
		e.t.Fatalf("age inbox for %s: %v", user, err)
	}
}

// buildGo builds the package at pkgPath (resolved relative to dir) into
// outPath.
func buildGo(dir, pkgPath, outPath string) error {
	cmd := exec.Command("go", "build", "-o", outPath, pkgPath)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

// generateSelfSignedCert builds an ECDSA certificate for 127.0.0.1,
// valid for one hour. The dubplate binary trusts it via SSL_CERT_FILE
// (crypto/x509's Unix system-pool loader reads that env var), so the
// compiled binary's own SMTP client -- which this package has no access
// to configure directly -- verifies the local test SMTP server the same
// way it would verify a real one.
func generateSelfSignedCert() (tls.Certificate, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("generate key: %w", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("create certificate: %w", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("marshal key: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("load key pair: %w", err)
	}
	return cert, certPEM, nil
}
