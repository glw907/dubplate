// Package checks runs dubplate's health checks and its doctor report.
// [DiskChecker] and [NavidromeChecker] port the retired shell layer's
// scripts/musicbox-disk-check and scripts/musicbox-navidrome-check: each
// pings its own slug success or failure through [config.Config.PingURLs].
// [RunDoctor] builds the report `dubplate doctor` prints: every required
// config field's resolution state, then, only once config resolves,
// read-only probes of SMTP, R2, Navidrome, and the Access certs endpoint;
// a writability probe of the upload store and the contributions ledger;
// and syntax-only checks of PublicNavidromeURL and every configured ping
// URL. See uploader.go for the upload-path probes.
package checks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ping"
)

const (
	diskSlug      = "disk"
	navidromeSlug = "navidrome"
)

// probeTimeout bounds every doctor probe's dial and round trip. Doctor is
// a human-facing command, not a systemd timer, so this is generous
// without risking a stuck run.
const probeTimeout = 10 * time.Second

// DiskChecker runs the disk-usage check: how full the filesystem holding
// [config.Paths.LibraryDir] is, against [config.Thresholds.DiskPct].
type DiskChecker struct {
	Cfg  *config.Config
	Ping *ping.Client

	// UsagePct reports how full (0-100) the filesystem holding dir is.
	// Production leaves it nil, which runs [statfsUsagePct]; tests
	// override it to avoid depending on the real filesystem's fill level.
	UsagePct func(dir string) (int, error)
}

// Run pings the "disk" slug success when LibraryDir's filesystem is at or
// under DiskPct, failure otherwise (or when the usage read itself
// fails), and returns a non-nil error in every case but success.
func (c *DiskChecker) Run(ctx context.Context) error {
	usagePct := c.UsagePct
	if usagePct == nil {
		usagePct = statfsUsagePct
	}

	pingURL := c.Cfg.PingURLs[diskSlug]
	dir := c.Cfg.Paths.LibraryDir

	pct, err := usagePct(dir)
	if err != nil {
		pingResult(ctx, c.Ping, pingURL, false)
		return fmt.Errorf("disk usage %s: %w", dir, err)
	}
	if pct > c.Cfg.Thresholds.DiskPct {
		pingResult(ctx, c.Ping, pingURL, false)
		return fmt.Errorf("%s is %d%% full, over the %d%% threshold", dir, pct, c.Cfg.Thresholds.DiskPct)
	}
	pingResult(ctx, c.Ping, pingURL, true)
	return nil
}

// statfsUsagePct reports the percentage of the filesystem holding dir
// that is in use, matching df -P's "Capacity" column: used blocks over
// used-plus-available, not used-over-total, so blocks reserved for root
// count as used rather than as headroom this process could never
// actually write into.
func statfsUsagePct(dir string) (int, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(dir, &stat); err != nil {
		return 0, fmt.Errorf("statfs %s: %w", dir, err)
	}
	used := stat.Blocks - stat.Bfree
	total := used + stat.Bavail
	if total == 0 {
		return 0, nil
	}
	return int(math.Round(100 * float64(used) / float64(total))), nil
}

// NavidromeChecker runs the Navidrome reachability check: a bare GET
// against the Subsonic `/rest/ping` endpoint. It is deliberately
// unauthenticated -- Subsonic always answers HTTP 200 even on auth
// failure, so a non-2xx or a request error is the only signal that
// Navidrome's listener is unreachable, and this check needs no
// credentials to tell that.
type NavidromeChecker struct {
	Cfg  *config.Config
	Ping *ping.Client

	// Client overrides the HTTP client Run probes with. Production
	// leaves it nil, which runs a client with a [probeTimeout] timeout;
	// tests override it to reach an httptest server.
	Client *http.Client
}

// Run pings the "navidrome" slug success when /rest/ping answers under
// 400, failure otherwise, and returns a non-nil error in every case but
// success.
func (c *NavidromeChecker) Run(ctx context.Context) error {
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: probeTimeout}
	}

	pingURL := c.Cfg.PingURLs[navidromeSlug]
	if err := navidromePing(ctx, client, c.Cfg.Navidrome.URL); err != nil {
		pingResult(ctx, c.Ping, pingURL, false)
		return fmt.Errorf("navidrome check: %w", err)
	}
	pingResult(ctx, c.Ping, pingURL, true)
	return nil
}

// navidromePing GETs base + "/rest/ping" using client. It is shared by
// [NavidromeChecker.Run] and doctor's Navidrome probe, which is the same
// request for a different purpose (report a result vs. ping a slug).
func navidromePing(ctx context.Context, client *http.Client, base string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(base, "/")+"/rest/ping", nil)
	if err != nil {
		return fmt.Errorf("build navidrome ping request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("navidrome ping %s: %w", base, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("navidrome ping %s: status %s", base, resp.Status)
	}
	return nil
}

// pingResult reports ok as a success or failure ping to url. A ping
// delivery failure is logged and swallowed, matching every other
// pingSuccess helper in the codebase: monitoring must never take down
// what it monitors.
func pingResult(ctx context.Context, pc *ping.Client, pingURL string, ok bool) {
	var err error
	if ok {
		err = pc.Success(ctx, pingURL)
	} else {
		err = pc.Fail(ctx, pingURL)
	}
	if err != nil {
		slog.Warn("checks: ping", "err", err)
	}
}

// FieldStatus is one required config field's env var name and whether it
// resolved.
type FieldStatus struct {
	Name    string
	Missing bool
}

// requiredFields lists the env var names [config.Load] requires, in the
// same order config.go's require() calls declare them. It exists so
// doctor can report on every field without reading the environment
// itself -- config stays the only place that does -- deriving each
// field's status entirely from [config.Load]'s returned error instead.
var requiredFields = []string{
	"DUBPLATE_ADMIN_EMAIL",
	"DUBPLATE_SMTP_HOST",
	"DUBPLATE_SMTP_PORT",
	"DUBPLATE_SMTP_USER",
	"DUBPLATE_SMTP_PASSWORD",
	"DUBPLATE_SMTP_FROM",
	"DUBPLATE_R2_ACCOUNT_ID",
	"DUBPLATE_R2_BUCKET",
	"DUBPLATE_R2_REMOTE",
	"DUBPLATE_R2_ACCESS_KEY_ID",
	"DUBPLATE_R2_SECRET_ACCESS_KEY",
	"DUBPLATE_NAVIDROME_URL",
	"DUBPLATE_NAVIDROME_ADMIN_USER",
	"DUBPLATE_NAVIDROME_ADMIN_PASSWORD",
	"DUBPLATE_CONTRIBUTORS_FILE",
	"DUBPLATE_REJECT_EMAIL_TEMPLATE",
	"DUBPLATE_BEETS_CONFIG",
	"DUBPLATE_BEETS_DB",
	"DUBPLATE_NAVIDROME_DB",
	"DUBPLATE_LIBRARY_DIR",
	"DUBPLATE_INBOX_DIR",
	"DUBPLATE_STAGING_DIR",
	"DUBPLATE_STATE_DIR",
	"DUBPLATE_RUN_DIR",
	"DUBPLATE_REVIEW_DIR",
	"DUBPLATE_QUARANTINE_DIR",
	"DUBPLATE_BACKUP_STAGING",
	"DUBPLATE_UPLOADS_DIR",
	"DUBPLATE_SERVE_BIND_ADDRS",
	"DUBPLATE_SERVE_PORT",
	"DUBPLATE_ACCESS_TEAM_DOMAIN",
	"DUBPLATE_ACCESS_AUD",
	"DUBPLATE_PUBLIC_NAVIDROME_URL",
}

// FieldStatuses reports every field [config.Load] requires and whether
// cfgErr -- Load's returned error, possibly nil -- marks it missing.
func FieldStatuses(cfgErr error) []FieldStatus {
	missing := missingFields(cfgErr)
	statuses := make([]FieldStatus, len(requiredFields))
	for i, name := range requiredFields {
		statuses[i] = FieldStatus{Name: name, Missing: missing[name]}
	}
	return statuses
}

// missingFields unpacks cfgErr's joined [config.ErrMissing] entries into
// the set of env var names they name. cfgErr is always the direct result
// of [errors.Join] (config.Load wraps nothing further around it), so a
// plain type assertion for the unexported joinError's Unwrap() []error
// shape is enough; an error present for another reason (an invalid port,
// say) is simply absent from the returned set.
func missingFields(cfgErr error) map[string]bool {
	missing := make(map[string]bool)
	joined, ok := cfgErr.(interface{ Unwrap() []error })
	if !ok {
		return missing
	}
	for _, e := range joined.Unwrap() {
		if !errors.Is(e, config.ErrMissing) {
			continue
		}
		if name, _, ok := strings.Cut(e.Error(), ":"); ok {
			missing[name] = true
		}
	}
	return missing
}

// DoctorReport is `dubplate doctor`'s full output.
type DoctorReport struct {
	Fields []FieldStatus

	// Probed reports whether config resolved cleanly enough to run the
	// probes below. When false, SMTPErr, R2Err, and NavidromeErr are all
	// nil not because the probes passed but because they never ran, and
	// ConfigErr carries config.Load's error -- the reason, which is not
	// always a missing field: an invalid port or threshold also leaves
	// Probed false but marks no [FieldStatus] as missing.
	Probed    bool
	ConfigErr error

	SMTPErr      error
	R2Err        error
	NavidromeErr error

	// AccessErr reports whether the configured Access team domain
	// publishes a JWKS and AccessAUD is shaped like a real application
	// tag; see [probeAccess].
	AccessErr error
	// UploadsWritableErr reports whether the tus upload store
	// (Paths.UploadsDir) actually accepts a write, not merely exists.
	UploadsWritableErr error
	// LedgerAppendErr reports whether the contributions ledger path is
	// reachable in append mode.
	LedgerAppendErr error
	// PublicNavidromeURLErr reports whether Serve.PublicNavidromeURL is
	// a syntactically valid absolute http(s) URL.
	PublicNavidromeURLErr error
	// SameFilesystemErr reports whether UploadsDir and InboxDir share a
	// device, the precondition atomic promotion depends on.
	SameFilesystemErr error

	// PingURLErrs maps each configured ping slug to its syntactic
	// validation error. A slug absent from the map validated cleanly.
	PingURLErrs map[string]error
}

// Failed reports whether doctor should exit nonzero: a missing field, a
// probe that never ran, a probe failure, or a malformed ping URL.
func (r DoctorReport) Failed() bool {
	if !r.Probed {
		return true
	}
	if r.SMTPErr != nil || r.R2Err != nil || r.NavidromeErr != nil {
		return true
	}
	if r.AccessErr != nil || r.UploadsWritableErr != nil || r.LedgerAppendErr != nil {
		return true
	}
	if r.PublicNavidromeURLErr != nil || r.SameFilesystemErr != nil {
		return true
	}
	return len(r.PingURLErrs) > 0
}

// RunDoctor builds a [DoctorReport]. cfg and cfgErr are [config.Load]'s
// return values: when cfgErr is non-nil, cfg is nil and the probes are
// skipped entirely, since there is nothing to probe with. rclonePath is
// the rclone binary the R2 probe runs; production passes "rclone".
// httpClient is the client the Navidrome and Access probes share;
// production passes nil, which runs a client with a [probeTimeout]
// timeout, and tests override it to reach an httptest server (including
// a TLS one, for the Access probe's real https certs endpoint).
func RunDoctor(ctx context.Context, cfg *config.Config, cfgErr error, rclonePath string, httpClient *http.Client) DoctorReport {
	report := DoctorReport{Fields: FieldStatuses(cfgErr)}
	if cfg == nil {
		report.ConfigErr = cfgErr
		return report
	}

	report.Probed = true

	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: probeTimeout}
	}

	smtpCtx, cancel := withProbeTimeout(ctx)
	report.SMTPErr = probeSMTP(smtpCtx, cfg.SMTP)
	cancel()

	r2Ctx, cancel := withProbeTimeout(ctx)
	report.R2Err = probeR2(r2Ctx, rclonePath, cfg.R2)
	cancel()

	navCtx, cancel := withProbeTimeout(ctx)
	report.NavidromeErr = navidromePing(navCtx, client, cfg.Navidrome.URL)
	cancel()

	accessCtx, cancel := withProbeTimeout(ctx)
	report.AccessErr = probeAccess(accessCtx, client, cfg.Serve)
	cancel()

	report.UploadsWritableErr = probeUploadsWritable(cfg.Paths.UploadsDir)
	report.LedgerAppendErr = probeLedgerAppendable(cfg.LedgerPath)
	report.PublicNavidromeURLErr = probePublicNavidromeURL(cfg.Serve.PublicNavidromeURL)
	report.SameFilesystemErr = probeSameFilesystem(cfg.Paths.UploadsDir, cfg.Paths.InboxDir)

	report.PingURLErrs = validatePingURLs(cfg.PingURLs)
	return report
}

// withProbeTimeout bounds a probe to [probeTimeout]. The SMTP dial and
// the Navidrome HTTP client already carry their own per-attempt timeout,
// but rclone (an external binary probeR2 execs) has no such bound of its
// own: with bad credentials against the real R2 endpoint, its S3 client
// retries for minutes before giving up, which would make one bad
// credential hang the whole doctor run. Wrapping every probe call the
// same way keeps that bound in one place rather than trusting each probe
// to impose it itself.
func withProbeTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, probeTimeout)
}

// probeSMTP connects to cfg's SMTP server, sends EHLO, then QUIT. It
// never authenticates or sends a message -- doctor's probes are
// read-only against every backend they touch.
func probeSMTP(ctx context.Context, cfg config.SMTPConfig) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	dialer := &net.Dialer{Timeout: probeTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp %s: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp handshake %s: %w", addr, err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Hello("dubplate-doctor"); err != nil {
		return fmt.Errorf("smtp EHLO %s: %w", addr, err)
	}
	return client.Quit()
}

// probeR2 runs a read-only rclone listing one level deep into cfg's
// bucket, confirming the R2 credentials and remote resolve without
// writing or deleting anything.
func probeR2(ctx context.Context, rclonePath string, cfg config.R2Config) error {
	target := cfg.Remote + ":" + cfg.Bucket
	cmd := exec.CommandContext(ctx, rclonePath, "lsjson", "--max-depth", "1", target)
	cmd.Env = append(os.Environ(), rcloneEnv(cfg)...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rclone lsjson %s: %w: %s", target, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// rcloneEnv builds the RCLONE_CONFIG_<REMOTE>_* variables that define r2
// as an on-the-fly rclone remote, keyed by r2.Remote uppercased. It
// duplicates internal/backup's and internal/restore's own copy rather
// than sharing one: each package owns its own subprocess seam, the
// pattern already established between those two.
func rcloneEnv(r2 config.R2Config) []string {
	prefix := "RCLONE_CONFIG_" + strings.ToUpper(strings.ReplaceAll(r2.Remote, "-", "_")) + "_"
	return []string{
		prefix + "TYPE=s3",
		prefix + "NO_CHECK_BUCKET=true",
		prefix + "PROVIDER=Cloudflare",
		prefix + "ENDPOINT=https://" + r2.AccountID + ".r2.cloudflarestorage.com",
		prefix + "ACCESS_KEY_ID=" + r2.AccessKeyID,
		prefix + "SECRET_ACCESS_KEY=" + r2.SecretAccessKey,
	}
}

// validatePingURLs checks that every configured ping URL is
// syntactically well-formed. It never dials one: a probing GET here
// would record a false success against a check that never actually ran,
// since success and failure pings are how the checks above report their
// own real results, not doctor's.
func validatePingURLs(urls map[string]string) map[string]error {
	problems := make(map[string]error)
	for slug, raw := range urls {
		parsed, err := url.Parse(raw)
		if err != nil {
			problems[slug] = fmt.Errorf("parse %q: %w", raw, err)
			continue
		}
		if parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			problems[slug] = fmt.Errorf("%q is not an absolute http(s) URL", raw)
		}
	}
	return problems
}
