package checks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/glw907/dubplate/internal/config"
)

// accessCertsPath is appended to the team domain to build the JWKS
// endpoint Cloudflare Access publishes its signing keys at. It duplicates
// internal/access's own copy of the same path rather than importing it:
// each package owns its own probe of the surface it cares about, the
// pattern already established between internal/backup, internal/restore,
// and this package's own rcloneEnv.
const accessCertsPath = "/cdn-cgi/access/certs"

// accessAUDLen is the length of a Cloudflare Access application's AUD
// tag: a hex-encoded SHA-256 digest.
const accessAUDLen = 64

// probeAccess checks that cfg's Access team domain publishes a JWKS
// (proving `dubplate serve` can reach the endpoint [access.Verifier]
// depends on) and that AccessAUD is shaped like a real Access application
// tag. It never validates a token -- doctor has none to check -- so a
// misconfigured AUD naming a different, equally real application on the
// same team is outside what this probe can catch.
func probeAccess(ctx context.Context, client *http.Client, cfg config.ServeConfig) error {
	if len(cfg.AccessAUD) != accessAUDLen || !isHex(cfg.AccessAUD) {
		return fmt.Errorf("access AUD %q is not a %d-character hex tag", cfg.AccessAUD, accessAUDLen)
	}

	certsURL := "https://" + cfg.AccessTeamDomain + accessCertsPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certsURL, nil)
	if err != nil {
		return fmt.Errorf("build access certs request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch access certs %s: %w", certsURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch access certs %s: status %s", certsURL, resp.Status)
	}

	var doc struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&doc); err != nil {
		return fmt.Errorf("parse access certs %s: %w", certsURL, err)
	}
	if len(doc.Keys) == 0 {
		return fmt.Errorf("access certs %s: no keys published", certsURL)
	}
	return nil
}

// isHex reports whether every byte of s is a lowercase hex digit.
func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// probeUploadsWritable creates and removes a probe file inside dir,
// proving the tus upload store is actually writable by the process --
// not merely present -- rather than the SMTP/R2/Navidrome probes' style
// of a read-only connection, since writability has no read-only proxy.
func probeUploadsWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".dubplate-doctor-probe-*")
	if err != nil {
		return fmt.Errorf("create probe file in %s: %w", dir, err)
	}
	path := f.Name()
	closeErr := f.Close()
	removeErr := os.Remove(path)
	if closeErr != nil {
		return fmt.Errorf("close probe file %s: %w", path, closeErr)
	}
	if removeErr != nil {
		return fmt.Errorf("remove probe file %s: %w", path, removeErr)
	}
	return nil
}

// probeLedgerAppendable proves the ledger append path is reachable and
// writable without adding a record doctor did not actually produce, and
// never creates the ledger: doctor runs as root ahead of the first
// `serve` start (see docs/bringup-runbook.md), while `serve` runs as
// User=music, so a root-created file would leave a ledger the music user
// can never append to. When path already exists it is opened O_APPEND and
// closed unwritten; when it does not, a temp file is created and removed
// in the same directory instead (probeUploadsWritable's technique),
// proving the parent directory is writable without touching path itself.
func probeLedgerAppendable(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		if err := f.Close(); err != nil {
			return fmt.Errorf("close ledger %s: %w", path, err)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("open ledger %s for append: %w", path, err)
	}
	return probeUploadsWritable(filepath.Dir(path))
}

// probePublicNavidromeURL checks that raw is an absolute http(s) URL. It
// never dials the address: [ServeConfig.PublicNavidromeURL] is what the
// journey UI links contributors to, on a network doctor's own host may
// not share, so syntax is all a probe run from the box can promise.
func probePublicNavidromeURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse %q: %w", raw, err)
	}
	if parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%q is not an absolute http(s) URL", raw)
	}
	return nil
}

// probeSameFilesystem checks that a and b resolve to the same device,
// the precondition promotion relies on: moving a finished upload from
// UploadsDir into InboxDir is a single [os.Rename], which is atomic only
// within one filesystem and silently falls back to a non-atomic
// copy-and-delete across a mount boundary -- exactly the partial-album
// risk the pass's global constraints forbid.
func probeSameFilesystem(a, b string) error {
	var statA, statB unix.Stat_t
	if err := unix.Stat(a, &statA); err != nil {
		return fmt.Errorf("stat %s: %w", a, err)
	}
	if err := unix.Stat(b, &statB); err != nil {
		return fmt.Errorf("stat %s: %w", b, err)
	}
	if !sameDevice(statA, statB) {
		return fmt.Errorf("%s and %s are on different filesystems (dev %d vs %d)", a, b, statA.Dev, statB.Dev)
	}
	return nil
}

// sameDevice reports whether a and b were stat'd from the same device.
// Split out from [probeSameFilesystem] so the comparison itself is
// testable without two real mount points.
func sameDevice(a, b unix.Stat_t) bool {
	return a.Dev == b.Dev
}
