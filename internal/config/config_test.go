package config

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// requiredEnv is a complete set of every required environment variable,
// set to a distinct placeholder value per key so a test can unset one at
// a time and still leave every other required field satisfied.
func requiredEnv() map[string]string {
	return map[string]string{
		"DUBPLATE_ADMIN_EMAIL":              "admin@example.com",
		"DUBPLATE_SMTP_HOST":                "smtp.example.com",
		"DUBPLATE_SMTP_PORT":                "465",
		"DUBPLATE_SMTP_USER":                "smtp-user",
		"DUBPLATE_SMTP_PASSWORD":            "smtp-pass",
		"DUBPLATE_SMTP_FROM":                "dubplate@example.com",
		"DUBPLATE_R2_ACCOUNT_ID":            "account-id",
		"DUBPLATE_R2_BUCKET":                "bucket",
		"DUBPLATE_R2_REMOTE":                "r2",
		"DUBPLATE_R2_ACCESS_KEY_ID":         "access-key",
		"DUBPLATE_R2_SECRET_ACCESS_KEY":     "secret-key",
		"DUBPLATE_NAVIDROME_URL":            "http://localhost:4533",
		"DUBPLATE_NAVIDROME_ADMIN_USER":     "nd-admin",
		"DUBPLATE_NAVIDROME_ADMIN_PASSWORD": "nd-pass",
		"DUBPLATE_CONTRIBUTORS_FILE":        "/etc/dubplate/contributors.yaml",
		"DUBPLATE_REJECT_EMAIL_TEMPLATE":    "/opt/dubplate/config/reject-email.txt",
		"DUBPLATE_BEETS_CONFIG":             "/etc/dubplate/beets-config.yaml",
		"DUBPLATE_BEETS_DB":                 "/srv/music/state/library.db",
		"DUBPLATE_NAVIDROME_DB":             "/srv/music/state/navidrome/navidrome.db",
		"DUBPLATE_LIBRARY_DIR":              "/srv/music/library",
		"DUBPLATE_INBOX_DIR":                "/srv/music/inbox",
		"DUBPLATE_STAGING_DIR":              "/srv/music/staging",
		"DUBPLATE_STATE_DIR":                "/srv/music/state",
		"DUBPLATE_TRIGGER_DIR":              "/run/dubplate",
		"DUBPLATE_REVIEW_DIR":               "/srv/music/review",
		"DUBPLATE_QUARANTINE_DIR":           "/srv/music/quarantine",
		"DUBPLATE_BACKUP_STAGING":           "/srv/music/state/backup-staging",
		"DUBPLATE_UPLOADS_DIR":              "/srv/music/uploads",
		"DUBPLATE_SERVE_BIND_ADDRS":         "127.0.0.1,10.89.0.1",
		"DUBPLATE_SERVE_PORT":               "8080",
		"DUBPLATE_ACCESS_TEAM_DOMAIN":       "family.cloudflareaccess.com",
		"DUBPLATE_ACCESS_AUD":               "aud-value",
		"DUBPLATE_PUBLIC_NAVIDROME_URL":     "https://music.example.com",
		"DUBPLATE_SERVE_SITE_NAME":          "dubplate",
	}
}

// setEnv sets every key in env via t.Setenv, which is unset automatically
// when the test (or subtest) ends.
func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoad_allRequiredSet(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.AdminEmail != "admin@example.com" {
		t.Errorf("AdminEmail = %q, want admin@example.com", cfg.AdminEmail)
	}
	if cfg.SMTP.Port != 465 {
		t.Errorf("SMTP.Port = %d, want 465", cfg.SMTP.Port)
	}
	if cfg.R2.Bucket != "bucket" {
		t.Errorf("R2.Bucket = %q, want bucket", cfg.R2.Bucket)
	}
	if cfg.Navidrome.AdminUser != "nd-admin" {
		t.Errorf("Navidrome.AdminUser = %q, want nd-admin", cfg.Navidrome.AdminUser)
	}
	if cfg.Paths.LibraryDir != "/srv/music/library" {
		t.Errorf("Paths.LibraryDir = %q, want /srv/music/library", cfg.Paths.LibraryDir)
	}
	if cfg.Paths.UploadsDir != "/srv/music/uploads" {
		t.Errorf("Paths.UploadsDir = %q, want /srv/music/uploads", cfg.Paths.UploadsDir)
	}
	wantAddrs := []string{"127.0.0.1", "10.89.0.1"}
	if len(cfg.Serve.BindAddrs) != len(wantAddrs) {
		t.Fatalf("Serve.BindAddrs = %v, want %v", cfg.Serve.BindAddrs, wantAddrs)
	}
	for i, addr := range wantAddrs {
		if cfg.Serve.BindAddrs[i] != addr {
			t.Errorf("Serve.BindAddrs[%d] = %q, want %q", i, cfg.Serve.BindAddrs[i], addr)
		}
	}
	if cfg.Serve.Port != 8080 {
		t.Errorf("Serve.Port = %d, want 8080", cfg.Serve.Port)
	}
	if cfg.Serve.AccessTeamDomain != "family.cloudflareaccess.com" {
		t.Errorf("Serve.AccessTeamDomain = %q, want family.cloudflareaccess.com", cfg.Serve.AccessTeamDomain)
	}
	if cfg.Serve.AccessAUD != "aud-value" {
		t.Errorf("Serve.AccessAUD = %q, want aud-value", cfg.Serve.AccessAUD)
	}
	if cfg.Serve.PublicNavidromeURL != "https://music.example.com" {
		t.Errorf("Serve.PublicNavidromeURL = %q, want https://music.example.com", cfg.Serve.PublicNavidromeURL)
	}
	if cfg.Serve.SiteName != "dubplate" {
		t.Errorf("Serve.SiteName = %q, want dubplate", cfg.Serve.SiteName)
	}
}

// --- Serve.BindAddrs list parsing --------------------------------------

func TestLoad_bindAddrsBlankAfterTrimIsMissing(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_SERVE_BIND_ADDRS"] = " , , "
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil for an all-blank bind-addrs list")
	}
	if !errors.Is(err, ErrMissing) || !strings.Contains(err.Error(), "DUBPLATE_SERVE_BIND_ADDRS") {
		t.Errorf("Load() error = %v, want ErrMissing naming DUBPLATE_SERVE_BIND_ADDRS", err)
	}
}

func TestLoad_bindAddrsMissingIsErrMissing(t *testing.T) {
	env := requiredEnv()
	delete(env, "DUBPLATE_SERVE_BIND_ADDRS")
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "DUBPLATE_SERVE_BIND_ADDRS") {
		t.Errorf("Load() error %q missing DUBPLATE_SERVE_BIND_ADDRS", err.Error())
	}
}

// TestLoad_triggerDirMissingIsErrMissing pins Paths.TriggerDir, the
// purpose-named replacement for the retired RunDir field: unset
// DUBPLATE_TRIGGER_DIR joins the ErrMissing list exactly like every
// other required path.
func TestLoad_triggerDirMissingIsErrMissing(t *testing.T) {
	env := requiredEnv()
	delete(env, "DUBPLATE_TRIGGER_DIR")
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
	if !errors.Is(err, ErrMissing) || !strings.Contains(err.Error(), "DUBPLATE_TRIGGER_DIR") {
		t.Errorf("Load() error = %v, want ErrMissing naming DUBPLATE_TRIGGER_DIR", err)
	}
}

// TestLoad_newRequiredFieldsMissingAreNamed pins this pass's remaining new
// required fields (Serve.Port/AccessTeamDomain/AccessAUD/
// PublicNavidromeURL/SiteName, Paths.UploadsDir): each unset field's env
// var name appears in Load's joined error, exactly like every
// pre-existing required field.
func TestLoad_newRequiredFieldsMissingAreNamed(t *testing.T) {
	env := requiredEnv()
	newFields := []string{
		"DUBPLATE_SERVE_PORT",
		"DUBPLATE_ACCESS_TEAM_DOMAIN",
		"DUBPLATE_ACCESS_AUD",
		"DUBPLATE_PUBLIC_NAVIDROME_URL",
		"DUBPLATE_UPLOADS_DIR",
		"DUBPLATE_SERVE_SITE_NAME",
	}
	for _, name := range newFields {
		delete(env, name)
	}
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
	if !errors.Is(err, ErrMissing) {
		t.Errorf("errors.Is(err, ErrMissing) = false, want true")
	}
	msg := err.Error()
	for _, name := range newFields {
		if !strings.Contains(msg, name) {
			t.Errorf("Load() error %q missing %q", msg, name)
		}
	}
}

// --- Upload defaults -----------------------------------------------------

func TestLoad_uploadDefaults(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	want := UploadConfig{
		ChunkMB:            8,
		MaxAlbumBytes:      2 * 1024 * 1024 * 1024,
		PartialTTL:         24 * time.Hour,
		MaxSessionsPerUser: 3,
	}
	if cfg.Upload != want {
		t.Errorf("Upload = %+v, want %+v", cfg.Upload, want)
	}
}

func TestLoad_uploadOverride(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_UPLOAD_CHUNK_MB"] = "12"
	env["DUBPLATE_UPLOAD_MAX_ALBUM_BYTES"] = "1073741824"
	env["DUBPLATE_UPLOAD_PARTIAL_TTL"] = "48h"
	env["DUBPLATE_UPLOAD_MAX_SESSIONS_PER_USER"] = "1"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	want := UploadConfig{
		ChunkMB:            12,
		MaxAlbumBytes:      1073741824,
		PartialTTL:         48 * time.Hour,
		MaxSessionsPerUser: 1,
	}
	if cfg.Upload != want {
		t.Errorf("Upload = %+v, want %+v", cfg.Upload, want)
	}
}

func TestLoad_uploadPartialTTLInvalidDuration(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_UPLOAD_PARTIAL_TTL"] = "not-a-duration"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil for invalid duration")
	}
	if !strings.Contains(err.Error(), "DUBPLATE_UPLOAD_PARTIAL_TTL") {
		t.Errorf("Load() error %q missing DUBPLATE_UPLOAD_PARTIAL_TTL", err.Error())
	}
}

// --- LedgerPath ------------------------------------------------------------

func TestLoad_ledgerPathDefaultsFromStateDir(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	want := "/srv/music/state/contributions.jsonl"
	if cfg.LedgerPath != want {
		t.Errorf("LedgerPath = %q, want %q", cfg.LedgerPath, want)
	}
}

func TestLoad_ledgerPathOverride(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_LEDGER_PATH"] = "/custom/ledger.jsonl"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.LedgerPath != "/custom/ledger.jsonl" {
		t.Errorf("LedgerPath = %q, want /custom/ledger.jsonl", cfg.LedgerPath)
	}
}

// --- StallMinutes / ReportRetention ----------------------------------------

func TestStallMinutesDefaultExceedsTimerPlusQuiescence(t *testing.T) {
	floor := importTimerIntervalMin + defaultQuiescentMin
	if defaultStallMinutes <= floor {
		t.Fatalf("defaultStallMinutes = %d, want > importTimerIntervalMin(%d)+defaultQuiescentMin(%d) = %d",
			defaultStallMinutes, importTimerIntervalMin, defaultQuiescentMin, floor)
	}
}

func TestLoad_stallAndReportRetentionDefaults(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Thresholds.StallMinutes != defaultStallMinutes {
		t.Errorf("Thresholds.StallMinutes = %d, want %d", cfg.Thresholds.StallMinutes, defaultStallMinutes)
	}
	if cfg.Thresholds.ReportRetention != defaultReportRetention {
		t.Errorf("Thresholds.ReportRetention = %d, want %d", cfg.Thresholds.ReportRetention, defaultReportRetention)
	}
}

func TestLoad_stallMinutesOverride(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_STALL_MINUTES"] = "45"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Thresholds.StallMinutes != 45 {
		t.Errorf("Thresholds.StallMinutes = %d, want 45", cfg.Thresholds.StallMinutes)
	}
}

func TestLoad_missingFieldsJoinedNotFirstWins(t *testing.T) {
	env := requiredEnv()
	delete(env, "DUBPLATE_ADMIN_EMAIL")
	delete(env, "DUBPLATE_R2_BUCKET")
	delete(env, "DUBPLATE_NAVIDROME_URL")
	setEnv(t, env)

	cfg, err := Load()
	if cfg != nil {
		t.Fatalf("Load() cfg = %+v, want nil on error", cfg)
	}
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
	if !errors.Is(err, ErrMissing) {
		t.Errorf("errors.Is(err, ErrMissing) = false, want true")
	}

	// errors.Join, not first-wins: every missing field's name is present
	// in the joined message, not just the first one encountered.
	msg := err.Error()
	for _, want := range []string{"DUBPLATE_ADMIN_EMAIL", "DUBPLATE_R2_BUCKET", "DUBPLATE_NAVIDROME_URL"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Load() error %q missing %q", msg, want)
		}
	}
}

func TestLoad_invalidSMTPPort(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_SMTP_PORT"] = "not-a-number"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil for invalid port")
	}
	if !strings.Contains(err.Error(), "DUBPLATE_SMTP_PORT") {
		t.Errorf("Load() error %q missing DUBPLATE_SMTP_PORT", err.Error())
	}
}

func TestLoad_thresholdDefaults(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	want := Thresholds{
		QuiescentMin:    10,
		FreeFloorGB:     5,
		ArchiveCapGB:    2,
		RetentionDays:   60,
		DiskPct:         90,
		StallMinutes:    defaultStallMinutes,
		ReportRetention: defaultReportRetention,
	}
	if cfg.Thresholds != want {
		t.Errorf("Thresholds = %+v, want %+v", cfg.Thresholds, want)
	}
}

func TestLoad_thresholdOverride(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_QUIESCENT_MINUTES"] = "20"
	env["DUBPLATE_DISK_PCT"] = "80"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Thresholds.QuiescentMin != 20 {
		t.Errorf("Thresholds.QuiescentMin = %d, want 20", cfg.Thresholds.QuiescentMin)
	}
	if cfg.Thresholds.DiskPct != 80 {
		t.Errorf("Thresholds.DiskPct = %d, want 80", cfg.Thresholds.DiskPct)
	}
	// Untouched thresholds keep their default.
	if cfg.Thresholds.FreeFloorGB != 5 {
		t.Errorf("Thresholds.FreeFloorGB = %d, want 5", cfg.Thresholds.FreeFloorGB)
	}
}

func TestLoad_pingURLsPrefixScan(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_PING_IMPORT_URL"] = "https://pings.example/import"
	env["DUBPLATE_PING_DISK_URL"] = "https://pings.example/disk"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	want := map[string]string{
		"import": "https://pings.example/import",
		"disk":   "https://pings.example/disk",
	}
	if len(cfg.PingURLs) != len(want) {
		t.Fatalf("PingURLs = %v, want %v", cfg.PingURLs, want)
	}
	for slug, url := range want {
		if cfg.PingURLs[slug] != url {
			t.Errorf("PingURLs[%q] = %q, want %q", slug, cfg.PingURLs[slug], url)
		}
	}
}

// TestLoad_pingURLsPrefixScanEdgeCases pins the prefix-scan boundary: a
// name where the prefix and suffix overlap (DUBPLATE_PING_URL, too short
// to carry a slug between them), a name with the prefix but not the
// suffix, and a name with neither, are all ignored rather than
// mis-sliced or mistaken for a ping URL.
func TestLoad_pingURLsPrefixScanEdgeCases(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_PING_URL"] = "https://pings.example/no-slug"
	env["DUBPLATE_PING_FOO"] = "https://pings.example/no-suffix"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(cfg.PingURLs) != 0 {
		t.Errorf("PingURLs = %v, want empty", cfg.PingURLs)
	}
	// DUBPLATE_ADMIN_EMAIL, set by requiredEnv, carries neither the ping
	// prefix nor the _URL suffix and must never surface as a ping entry.
	if _, ok := cfg.PingURLs["admin_email"]; ok {
		t.Error(`PingURLs["admin_email"] present, want absent`)
	}
}

func TestLoad_pingURLsEmptyWhenUnset(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(cfg.PingURLs) != 0 {
		t.Errorf("PingURLs = %v, want empty", cfg.PingURLs)
	}
}

func TestLoad_collectionsDefaultUnnamed(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	want := []string{""}
	if len(cfg.Collections) != 1 || cfg.Collections[0] != want[0] {
		t.Errorf("Collections = %v, want %v", cfg.Collections, want)
	}
}

func TestLoad_collectionsFromEnv(t *testing.T) {
	env := requiredEnv()
	env["DUBPLATE_COLLECTIONS"] = "family, live-recordings"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	want := []string{"family", "live-recordings"}
	if len(cfg.Collections) != len(want) {
		t.Fatalf("Collections = %v, want %v", cfg.Collections, want)
	}
	for i, c := range want {
		if cfg.Collections[i] != c {
			t.Errorf("Collections[%d] = %q, want %q", i, cfg.Collections[i], c)
		}
	}
}
