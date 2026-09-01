package checks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhale/smtpd"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ping"
)

// newPingTestServer starts a local HTTP server that increments *hits and
// records the last method it saw, standing in for a pings provider.
func newPingTestServer(t *testing.T, hits *int, gotMethod *string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*hits++
		*gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func TestDiskCheckerRunUnderThresholdPingsSuccess(t *testing.T) {
	hits, method := 0, ""
	cfg := &config.Config{
		Paths:      config.Paths{LibraryDir: "/srv/music"},
		Thresholds: config.Thresholds{DiskPct: 90},
		PingURLs:   map[string]string{diskSlug: newPingTestServer(t, &hits, &method)},
	}
	c := &DiskChecker{
		Cfg:      cfg,
		Ping:     ping.NewClient(),
		UsagePct: func(string) (int, error) { return 50, nil },
	}

	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
	if hits != 1 || method != http.MethodGet {
		t.Errorf("ping hits=%d method=%q, want 1 GET", hits, method)
	}
}

func TestDiskCheckerRunAtThresholdPingsSuccess(t *testing.T) {
	// The shell layer's `-gt` test: usage exactly at the threshold is
	// still healthy, only usage strictly over it fails.
	hits, method := 0, ""
	cfg := &config.Config{
		Paths:      config.Paths{LibraryDir: "/srv/music"},
		Thresholds: config.Thresholds{DiskPct: 90},
		PingURLs:   map[string]string{diskSlug: newPingTestServer(t, &hits, &method)},
	}
	c := &DiskChecker{
		Cfg:      cfg,
		Ping:     ping.NewClient(),
		UsagePct: func(string) (int, error) { return 90, nil },
	}

	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("Run() = %v, want nil at exactly the threshold", err)
	}
	if method != http.MethodGet {
		t.Errorf("ping method = %q, want GET", method)
	}
}

func TestDiskCheckerRunOverThresholdPingsFailAndErrors(t *testing.T) {
	hits, method := 0, ""
	cfg := &config.Config{
		Paths:      config.Paths{LibraryDir: "/srv/music"},
		Thresholds: config.Thresholds{DiskPct: 90},
		PingURLs:   map[string]string{diskSlug: newPingTestServer(t, &hits, &method)},
	}
	c := &DiskChecker{
		Cfg:      cfg,
		Ping:     ping.NewClient(),
		UsagePct: func(string) (int, error) { return 91, nil },
	}

	err := c.Run(context.Background())
	if err == nil {
		t.Fatal("Run() = nil, want an error when usage exceeds the threshold")
	}
	if hits != 1 || method != http.MethodPost {
		t.Errorf("ping hits=%d method=%q, want 1 POST (fail)", hits, method)
	}
}

func TestDiskCheckerRunUsagePctErrorPingsFailAndErrors(t *testing.T) {
	hits, method := 0, ""
	cfg := &config.Config{
		Paths:      config.Paths{LibraryDir: "/srv/music"},
		Thresholds: config.Thresholds{DiskPct: 90},
		PingURLs:   map[string]string{diskSlug: newPingTestServer(t, &hits, &method)},
	}
	wantErr := errors.New("statfs: no such device")
	c := &DiskChecker{
		Cfg:      cfg,
		Ping:     ping.NewClient(),
		UsagePct: func(string) (int, error) { return 0, wantErr },
	}

	err := c.Run(context.Background())
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("Run() = %v, want it to wrap %v", err, wantErr)
	}
	if method != http.MethodPost {
		t.Errorf("ping method = %q, want POST (fail)", method)
	}
}

func TestNavidromeCheckerRunReachablePingsSuccess(t *testing.T) {
	navidrome := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/ping" {
			t.Errorf("navidrome saw path %q, want /rest/ping", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer navidrome.Close()

	hits, method := 0, ""
	cfg := &config.Config{
		Navidrome: config.NavidromeConfig{URL: navidrome.URL},
		PingURLs:  map[string]string{navidromeSlug: newPingTestServer(t, &hits, &method)},
	}
	c := &NavidromeChecker{Cfg: cfg, Ping: ping.NewClient()}

	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
	if hits != 1 || method != http.MethodGet {
		t.Errorf("ping hits=%d method=%q, want 1 GET", hits, method)
	}
}

func TestNavidromeCheckerRunUnreachablePingsFailAndErrors(t *testing.T) {
	navidrome := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer navidrome.Close()

	hits, method := 0, ""
	cfg := &config.Config{
		Navidrome: config.NavidromeConfig{URL: navidrome.URL},
		PingURLs:  map[string]string{navidromeSlug: newPingTestServer(t, &hits, &method)},
	}
	c := &NavidromeChecker{Cfg: cfg, Ping: ping.NewClient()}

	if err := c.Run(context.Background()); err == nil {
		t.Fatal("Run() = nil, want an error for a 503 response")
	}
	if method != http.MethodPost {
		t.Errorf("ping method = %q, want POST (fail)", method)
	}
}

func TestNavidromeCheckerRunConnectionRefusedPingsFail(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	hits, method := 0, ""
	cfg := &config.Config{
		Navidrome: config.NavidromeConfig{URL: "http://" + addr},
		PingURLs:  map[string]string{navidromeSlug: newPingTestServer(t, &hits, &method)},
	}
	c := &NavidromeChecker{Cfg: cfg, Ping: ping.NewClient()}

	if err := c.Run(context.Background()); err == nil {
		t.Fatal("Run() = nil, want an error when the port refuses connections")
	}
	if method != http.MethodPost {
		t.Errorf("ping method = %q, want POST (fail)", method)
	}
}

func TestFieldStatuses(t *testing.T) {
	cfgErr := errors.Join(
		fmt.Errorf("DUBPLATE_ADMIN_EMAIL: %w", config.ErrMissing),
		fmt.Errorf("DUBPLATE_R2_BUCKET: %w", config.ErrMissing),
	)

	statuses := FieldStatuses(cfgErr)
	missing := map[string]bool{}
	for _, s := range statuses {
		missing[s.Name] = s.Missing
	}

	if len(statuses) != len(requiredFields) {
		t.Fatalf("len(statuses) = %d, want %d (one per required field)", len(statuses), len(requiredFields))
	}
	if !missing["DUBPLATE_ADMIN_EMAIL"] {
		t.Error("DUBPLATE_ADMIN_EMAIL not reported missing")
	}
	if !missing["DUBPLATE_R2_BUCKET"] {
		t.Error("DUBPLATE_R2_BUCKET not reported missing")
	}
	if missing["DUBPLATE_SMTP_HOST"] {
		t.Error("DUBPLATE_SMTP_HOST reported missing, want resolved")
	}
}

func TestFieldStatusesNilErrorAllResolved(t *testing.T) {
	statuses := FieldStatuses(nil)
	for _, s := range statuses {
		if s.Missing {
			t.Errorf("field %s reported missing when cfgErr is nil", s.Name)
		}
	}
}

func TestValidatePingURLs(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://pings.example/hc/disk", false},
		{"valid http", "http://localhost:8080/hc/disk", false},
		{"missing scheme", "pings.example/hc/disk", true},
		{"unsupported scheme", "ftp://pings.example/hc/disk", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := validatePingURLs(map[string]string{"disk": tt.url})
			_, got := problems["disk"]
			if got != tt.wantErr {
				t.Errorf("validatePingURLs(%q) error present = %v, want %v", tt.url, got, tt.wantErr)
			}
		})
	}
}

func TestProbeSMTPSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &smtpd.Server{Hostname: "localhost", Appname: "dubplate-test"}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = ln.Close() })

	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split addr: %v", err)
	}
	portNum := 0
	if _, err := fmt.Sscanf(port, "%d", &portNum); err != nil {
		t.Fatalf("parse port: %v", err)
	}

	err = probeSMTP(context.Background(), config.SMTPConfig{Host: host, Port: portNum})
	if err != nil {
		t.Fatalf("probeSMTP() = %v, want nil", err)
	}
}

func TestProbeSMTPConnectionRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split addr: %v", err)
	}
	portNum := 0
	if _, err := fmt.Sscanf(port, "%d", &portNum); err != nil {
		t.Fatalf("parse port: %v", err)
	}

	if err := probeSMTP(context.Background(), config.SMTPConfig{Host: host, Port: portNum}); err == nil {
		t.Fatal("probeSMTP() = nil, want an error when nothing is listening")
	}
}

// writeStubRclone writes an executable stub rclone script and returns
// its path, the R2 probe's subprocess seam.
func writeStubRclone(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rclone")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("write stub rclone: %v", err)
	}
	return path
}

func TestProbeR2Success(t *testing.T) {
	argsLog := filepath.Join(t.TempDir(), "args.log")
	rclonePath := writeStubRclone(t, `echo "$@" >> "`+argsLog+`"`)

	cfg := config.R2Config{AccountID: "acct", Bucket: "bucket", Remote: "r2", AccessKeyID: "key", SecretAccessKey: "secret"}
	if err := probeR2(context.Background(), rclonePath, cfg); err != nil {
		t.Fatalf("probeR2() = %v, want nil", err)
	}

	data, err := os.ReadFile(argsLog)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}
	got := string(data)
	if want := "lsjson --max-depth 1 r2:bucket"; got != want+"\n" {
		t.Errorf("rclone args = %q, want %q", got, want)
	}
}

func TestProbeR2Failure(t *testing.T) {
	rclonePath := writeStubRclone(t, `echo "rclone: auth error" >&2; exit 1`)
	cfg := config.R2Config{Remote: "r2", Bucket: "bucket"}

	if err := probeR2(context.Background(), rclonePath, cfg); err == nil {
		t.Fatal("probeR2() = nil, want the rclone failure propagated")
	}
}

func TestRunDoctorConfigInvalidSkipsProbes(t *testing.T) {
	cfgErr := errors.Join(fmt.Errorf("DUBPLATE_ADMIN_EMAIL: %w", config.ErrMissing))

	report := RunDoctor(context.Background(), nil, cfgErr, "rclone", nil)

	if report.Probed {
		t.Error("Probed = true, want false when config did not resolve")
	}
	if report.SMTPErr != nil || report.R2Err != nil || report.NavidromeErr != nil {
		t.Error("a probe field is set even though probes were never run")
	}
	if report.AccessErr != nil || report.UploadsWritableErr != nil || report.LedgerAppendErr != nil {
		t.Error("a new probe field is set even though probes were never run")
	}
	if !report.Failed() {
		t.Error("Failed() = false, want true when config did not resolve")
	}
}

func TestRunDoctorConfigInvalidNotJustMissingSurfacesErr(t *testing.T) {
	// A parse error (invalid port, invalid threshold) is not an
	// ErrMissing entry, so no FieldStatus marks it missing -- ConfigErr
	// is the only place the report carries the reason for a reader.
	cfgErr := errors.Join(fmt.Errorf("DUBPLATE_SMTP_PORT: invalid port %q", "abc"))

	report := RunDoctor(context.Background(), nil, cfgErr, "rclone", nil)

	if report.Probed {
		t.Error("Probed = true, want false when config did not resolve")
	}
	if report.ConfigErr == nil || !strings.Contains(report.ConfigErr.Error(), `invalid port "abc"`) {
		t.Errorf("ConfigErr = %v, want it to surface the port parse error", report.ConfigErr)
	}
	for _, s := range report.Fields {
		if s.Missing {
			t.Errorf("field %s reported missing, want none: the failure is a parse error, not a missing field", s.Name)
		}
	}
}

// newAccessCertsTestServer starts an httptest TLS server publishing pub
// under kid at [accessCertsPath], standing in for the real
// https-only Cloudflare Access certs endpoint.
func newAccessCertsTestServer(t *testing.T, kid string, pub *rsa.PublicKey) *httptest.Server {
	t.Helper()
	e := big.NewInt(int64(pub.E)).Bytes()
	doc := struct {
		Keys []map[string]string `json:"keys"`
	}{Keys: []map[string]string{{
		"kid": kid,
		"kty": "RSA",
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(e),
	}}}

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != accessCertsPath {
			t.Errorf("access certs saw path %q, want %q", r.URL.Path, accessCertsPath)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			t.Fatalf("encode jwks response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRunDoctorAllGreen(t *testing.T) {
	navidrome := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer navidrome.Close()

	smtpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &smtpd.Server{Hostname: "localhost", Appname: "dubplate-test"}
	go func() { _ = srv.Serve(smtpLn) }()
	t.Cleanup(func() { _ = smtpLn.Close() })
	smtpHost, smtpPort, err := net.SplitHostPort(smtpLn.Addr().String())
	if err != nil {
		t.Fatalf("split addr: %v", err)
	}
	portNum := 0
	if _, err := fmt.Sscanf(smtpPort, "%d", &portNum); err != nil {
		t.Fatalf("parse port: %v", err)
	}

	rclonePath := writeStubRclone(t, okScript)

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	accessSrv := newAccessCertsTestServer(t, "kid-1", &priv.PublicKey)
	accessTeamDomain := strings.TrimPrefix(accessSrv.URL, "https://")

	root := t.TempDir()
	uploadsDir := filepath.Join(root, "uploads")
	inboxDir := filepath.Join(root, "inbox")
	for _, dir := range []string{uploadsDir, inboxDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	ledgerPath := filepath.Join(root, "contributions.jsonl")

	cfg := &config.Config{
		SMTP:      config.SMTPConfig{Host: smtpHost, Port: portNum},
		R2:        config.R2Config{Remote: "r2", Bucket: "bucket"},
		Navidrome: config.NavidromeConfig{URL: navidrome.URL},
		Serve: config.ServeConfig{
			AccessTeamDomain:   accessTeamDomain,
			AccessAUD:          strings.Repeat("a1", 32),
			PublicNavidromeURL: "https://navidrome.example.com",
		},
		Paths:      config.Paths{UploadsDir: uploadsDir, InboxDir: inboxDir},
		LedgerPath: ledgerPath,
		PingURLs:   map[string]string{"disk": "https://pings.example/hc/disk"},
	}

	report := RunDoctor(context.Background(), cfg, nil, rclonePath, accessSrv.Client())

	if !report.Probed {
		t.Fatal("Probed = false, want true")
	}
	if report.SMTPErr != nil {
		t.Errorf("SMTPErr = %v, want nil", report.SMTPErr)
	}
	if report.R2Err != nil {
		t.Errorf("R2Err = %v, want nil", report.R2Err)
	}
	if report.NavidromeErr != nil {
		t.Errorf("NavidromeErr = %v, want nil", report.NavidromeErr)
	}
	if report.AccessErr != nil {
		t.Errorf("AccessErr = %v, want nil", report.AccessErr)
	}
	if report.UploadsWritableErr != nil {
		t.Errorf("UploadsWritableErr = %v, want nil", report.UploadsWritableErr)
	}
	if report.LedgerAppendErr != nil {
		t.Errorf("LedgerAppendErr = %v, want nil", report.LedgerAppendErr)
	}
	if report.PublicNavidromeURLErr != nil {
		t.Errorf("PublicNavidromeURLErr = %v, want nil", report.PublicNavidromeURLErr)
	}
	if report.SameFilesystemErr != nil {
		t.Errorf("SameFilesystemErr = %v, want nil", report.SameFilesystemErr)
	}
	if len(report.PingURLErrs) != 0 {
		t.Errorf("PingURLErrs = %v, want none", report.PingURLErrs)
	}
	if report.Failed() {
		t.Error("Failed() = true, want false when every probe and field is clean")
	}
}

func TestDoctorReportFailedOnMalformedPingURL(t *testing.T) {
	report := DoctorReport{
		Fields:      FieldStatuses(nil),
		Probed:      true,
		PingURLErrs: map[string]error{"disk": errors.New("not an absolute http(s) URL")},
	}
	if !report.Failed() {
		t.Error("Failed() = false, want true when a ping URL is malformed")
	}
}

const okScript = ": # ok"
