package main

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// testDescription is the description every test declaration carries.
const testDescription = "a test key"

// testKey declares name as a described, not sensitive key.
func testKey(name string) config.KeyDecl {
	return config.KeyDecl{Key: name, Description: testDescription, Sensitivity: config.NotSensitive}
}

// storeHomedSensitive is the exact set of store-homed keys marked
// sensitive in the production registry: a live change to one changes
// how long a member stays signed in, how long a mailed link stays
// redeemable, how fresh a sign-in a sensitive write needs, whether an
// unknown address is told so, how far a protective backup floor
// reaches, or how little free space staging may fall to. A key joins it
// only by a ruling.
var storeHomedSensitive = []string{
	"DUBPLATE_BACKUP_ARCHIVE_RETENTION",
	"DUBPLATE_BACKUP_GUARD_FRACTION",
	"DUBPLATE_BACKUP_SNAPSHOT_KEEP",
	"DUBPLATE_INTAKE_FREE_SPACE_FLOOR",
	"DUBPLATE_MAGIC_LINK_EXPIRY",
	"DUBPLATE_REAUTH_WINDOW",
	"DUBPLATE_SESSION_ABSOLUTE_CAP",
	"DUBPLATE_SESSION_LIFETIME",
	"DUBPLATE_UNKNOWN_ADDRESS_ANSWER",
}

func TestProductionRegistry_StoreHomedSensitiveSet(t *testing.T) {
	settings := newProductionRegistry().Settings()
	if len(settings) < 10 {
		t.Fatalf("production registry holds %d settings, want the populated registry", len(settings))
	}
	var got []string
	for _, s := range settings {
		switch s.Sensitivity {
		case config.Sensitive:
			got = append(got, s.Key)
		case config.NotSensitive:
		default:
			t.Errorf("%s: sensitivity = %v, want a named sensitivity", s.Key, s.Sensitivity)
		}
		if strings.TrimSpace(s.Description) == "" {
			t.Errorf("%s: no description", s.Key)
		}
	}
	slices.Sort(got)
	if !slices.Equal(got, storeHomedSensitive) {
		t.Errorf("sensitive store-homed keys = %v, want %v", got, storeHomedSensitive)
	}
}

// envOnlySensitivity pins the sensitivity of every key only the
// environment sets, which changes only at a restart: Sensitive marks a path, an address,
// a credential, or a trust boundary, a change that needs care.
var envOnlySensitivity = map[string]config.Sensitivity{
	"DUBPLATE_LIBRARY_ROOT":               config.Sensitive,
	"DUBPLATE_STAGING_DIR":                config.Sensitive,
	"DUBPLATE_STATE_DIR":                  config.Sensitive,
	"DUBPLATE_SITE_NAME":                  config.NotSensitive,
	"DUBPLATE_ADMIN_EMAIL":                config.Sensitive,
	"DUBPLATE_IDENTITY_SOURCE":            config.Sensitive,
	"DUBPLATE_MAGIC_LINK_SECRET":          config.Sensitive,
	"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS": config.Sensitive,
	"DUBPLATE_SESSION_SECRET":             config.Sensitive,
	"DUBPLATE_SESSION_SECRET_PREVIOUS":    config.Sensitive,
	"DUBPLATE_PUBLIC_URL":                 config.Sensitive,
	"DUBPLATE_SMTP_HOST":                  config.Sensitive,
	"DUBPLATE_SMTP_PORT":                  config.Sensitive,
	"DUBPLATE_SMTP_USER":                  config.Sensitive,
	"DUBPLATE_SMTP_PASSWORD":              config.Sensitive,
	"DUBPLATE_SMTP_FROM":                  config.Sensitive,
	"DUBPLATE_SERVE_BIND_ADDRS":           config.Sensitive,
	"DUBPLATE_SERVE_PORT":                 config.Sensitive,
	"DUBPLATE_SERVE_PREFIX":               config.Sensitive,
	"DUBPLATE_STORE_BUSY_TIMEOUT":         config.NotSensitive,
	"DUBPLATE_STORE_MIGRATION_WAIT":       config.NotSensitive,
	"DUBPLATE_THEME_MATCH":                config.NotSensitive,
	"DUBPLATE_THEME_PALETTE":              config.NotSensitive,
	"DUBPLATE_BRANDING":                   config.NotSensitive,
	"DUBPLATE_OPERATOR_CSS_FILE":          config.Sensitive,
	"DUBPLATE_TRUSTED_PROXY_ENABLED":      config.Sensitive,
	"DUBPLATE_TRUSTED_PROXY_ADDRS":        config.Sensitive,
	"DUBPLATE_TRUSTED_PROXY_HEADER":       config.Sensitive,
	"DUBPLATE_BACKUP_ENABLED":             config.Sensitive,
	"DUBPLATE_BACKUP_REMOTE":              config.Sensitive,
	"DUBPLATE_INTAKE_FORMAT_ALLOWLIST":    config.NotSensitive,
}

func TestProductionRegistry_EnvironmentOnlyKeysCarryTheirMarks(t *testing.T) {
	reg := newProductionRegistry()
	got := make(map[string]config.Sensitivity)
	described := make(map[string]string)
	add := func(name, description string, sensitivity config.Sensitivity) {
		got[name] = sensitivity
		described[name] = description
	}
	for _, row := range reg.OwnedWall() {
		add(row.Key, row.Description, row.Sensitivity)
	}
	for _, th := range reg.Thresholds() {
		add(th.Key, th.Description, th.Sensitivity)
	}
	for _, src := range reg.IdentitySources() {
		for _, k := range slices.Concat(src.Required, src.Optional) {
			add(k.Key, k.Description, k.Sensitivity)
		}
	}
	for _, class := range reg.IntegrationClasses() {
		for _, k := range slices.Concat([]config.KeyDecl{class.Flag}, class.Required) {
			add(k.Key, k.Description, k.Sensitivity)
		}
	}
	if !maps.Equal(got, envOnlySensitivity) {
		for name, sensitivity := range got {
			if want, ok := envOnlySensitivity[name]; !ok || want != sensitivity {
				t.Errorf("%s: sensitivity = %v, pinned %v (pinned: %v)", name, sensitivity, want, ok)
			}
		}
		for name := range envOnlySensitivity {
			if _, ok := got[name]; !ok {
				t.Errorf("%s: pinned but not registered", name)
			}
		}
	}
	for name, description := range described {
		if strings.TrimSpace(description) == "" {
			t.Errorf("%s: no description", name)
		}
	}
}

func TestProductionRegistry_AdminEmailDescriptionNamesItsSeedRole(t *testing.T) {
	for _, row := range newProductionRegistry().OwnedWall() {
		if row.Key != "DUBPLATE_ADMIN_EMAIL" {
			continue
		}
		if !strings.Contains(row.Description, "seed") {
			t.Errorf("description = %q, want it to name the seed role", row.Description)
		}
		return
	}
	t.Fatal("DUBPLATE_ADMIN_EMAIL is not an owned wall row")
}
