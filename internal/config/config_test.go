package config

import (
	"errors"
	"strings"
	"testing"
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
		"DUBPLATE_RUN_DIR":                  "/run/dubplate",
		"DUBPLATE_REVIEW_DIR":               "/srv/music/review",
		"DUBPLATE_QUARANTINE_DIR":           "/srv/music/quarantine",
		"DUBPLATE_BACKUP_STAGING":           "/srv/music/state/backup-staging",
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
		QuiescentMin:  10,
		FreeFloorGB:   5,
		ArchiveCapGB:  2,
		RetentionDays: 60,
		DiskPct:       90,
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
