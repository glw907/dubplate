package checks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"

	"github.com/glw907/dubplate/internal/config"
)

func TestProbeAccessSuccess(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newAccessCertsTestServer(t, "kid-1", &priv.PublicKey)
	teamDomain := strings.TrimPrefix(srv.URL, "https://")

	cfg := config.ServeConfig{AccessTeamDomain: teamDomain, AccessAUD: strings.Repeat("a1", 32)}
	if err := probeAccess(context.Background(), srv.Client(), cfg); err != nil {
		t.Fatalf("probeAccess() = %v, want nil", err)
	}
}

func TestProbeAccessBadAUDNeverDialsCertsEndpoint(t *testing.T) {
	hits := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	teamDomain := strings.TrimPrefix(srv.URL, "https://")

	tests := []struct {
		name string
		aud  string
	}{
		{"too short", "a1"},
		{"uppercase hex", strings.Repeat("A1", 32)},
		{"non-hex", strings.Repeat("zz", 32)},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.ServeConfig{AccessTeamDomain: teamDomain, AccessAUD: tt.aud}
			if err := probeAccess(context.Background(), srv.Client(), cfg); err == nil {
				t.Errorf("probeAccess(aud=%q) = nil, want error", tt.aud)
			}
		})
	}
	if hits != 0 {
		t.Errorf("certs endpoint hit %d times, want 0: a malformed AUD must reject before any network call", hits)
	}
}

func TestProbeAccessUnreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	cfg := config.ServeConfig{AccessTeamDomain: addr, AccessAUD: strings.Repeat("a1", 32)}
	if err := probeAccess(context.Background(), &http.Client{}, cfg); err == nil {
		t.Fatal("probeAccess() = nil, want an error when the certs endpoint refuses connections")
	}
}

func TestProbeAccessNoKeysPublished(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(srv.Close)
	teamDomain := strings.TrimPrefix(srv.URL, "https://")

	cfg := config.ServeConfig{AccessTeamDomain: teamDomain, AccessAUD: strings.Repeat("a1", 32)}
	if err := probeAccess(context.Background(), srv.Client(), cfg); err == nil {
		t.Fatal("probeAccess() = nil, want an error when the JWKS has no keys")
	}
}

func TestProbeUploadsWritableSuccess(t *testing.T) {
	if err := probeUploadsWritable(t.TempDir()); err != nil {
		t.Errorf("probeUploadsWritable() = %v, want nil", err)
	}
}

func TestProbeUploadsWritableSuccessLeavesNoProbeFile(t *testing.T) {
	dir := t.TempDir()
	if err := probeUploadsWritable(dir); err != nil {
		t.Fatalf("probeUploadsWritable() = %v, want nil", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has %d entries after probe, want 0: the probe file must not be left behind", len(entries))
	}
}

func TestProbeUploadsWritableMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	if err := probeUploadsWritable(dir); err == nil {
		t.Fatal("probeUploadsWritable() = nil, want an error for a missing directory")
	}
}

func TestProbeLedgerAppendableSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contributions.jsonl")
	if err := probeLedgerAppendable(path); err != nil {
		t.Errorf("probeLedgerAppendable() = %v, want nil", err)
	}
}

func TestProbeLedgerAppendablePreservesExistingContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contributions.jsonl")
	want := []byte(`{"state":"queued"}` + "\n")
	if err := os.WriteFile(path, want, 0o644); err != nil {
		t.Fatalf("seed ledger: %v", err)
	}

	if err := probeLedgerAppendable(path); err != nil {
		t.Fatalf("probeLedgerAppendable() = %v, want nil", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("ledger content = %q, want unchanged %q", got, want)
	}
}

func TestProbeLedgerAppendableMissingFileLeavesNoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contributions.jsonl")
	if err := probeLedgerAppendable(path); err != nil {
		t.Fatalf("probeLedgerAppendable() = %v, want nil for a missing file in a writable dir", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("ledger file exists after probing a missing path: stat err = %v, want IsNotExist", err)
	}
}

func TestProbeLedgerAppendableMissingParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist", "contributions.jsonl")
	if err := probeLedgerAppendable(path); err == nil {
		t.Fatal("probeLedgerAppendable() = nil, want an error for a missing parent directory")
	}
}

func TestProbePublicNavidromeURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://navidrome.example/", false},
		{"valid http", "http://localhost:4533", false},
		{"missing scheme", "navidrome.example", true},
		{"unsupported scheme", "ftp://navidrome.example", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probePublicNavidromeURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("probePublicNavidromeURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestProbeSameFilesystemSuccess(t *testing.T) {
	dir := t.TempDir()
	if err := probeSameFilesystem(dir, dir); err != nil {
		t.Errorf("probeSameFilesystem() = %v, want nil for the same directory", err)
	}
}

func TestProbeSameFilesystemStatFailure(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist")
	if err := probeSameFilesystem(dir, missing); err == nil {
		t.Fatal("probeSameFilesystem() = nil, want an error when one path does not exist")
	}
}

func TestSameDevice(t *testing.T) {
	tests := []struct {
		name string
		a, b uint64
		want bool
	}{
		{"equal", 42, 42, true},
		{"different", 42, 43, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statA := unix.Stat_t{Dev: tt.a}
			statB := unix.Stat_t{Dev: tt.b}
			if got := sameDevice(statA, statB); got != tt.want {
				t.Errorf("sameDevice(%d, %d) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"lowercase hex", "0123456789abcdef", true},
		{"uppercase rejected", "0123456789ABCDEF", false},
		{"non-hex rejected", "hello-world", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHex(tt.s); got != tt.want {
				t.Errorf("isHex(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
