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
	return config.KeyDecl{Name: name, Description: testDescription, Mark: config.NotSensitive}
}

// storeHomedSensitive is the exact set of store-homed keys marked
// sensitive in the production registry: a live change to one changes
// how long a member stays signed in or how long a mailed link stays
// redeemable. A key joins it only by a ruling.
var storeHomedSensitive = []string{
	"DUBPLATE_MAGIC_LINK_EXPIRY",
	"DUBPLATE_SESSION_ABSOLUTE_CAP",
	"DUBPLATE_SESSION_LIFETIME",
}

func TestProductionRegistry_StoreHomedSensitiveSet(t *testing.T) {
	settings := newProductionRegistry().Settings()
	if len(settings) < 10 {
		t.Fatalf("production registry holds %d settings, want the populated registry", len(settings))
	}
	var got []string
	for _, s := range settings {
		switch s.Mark {
		case config.Sensitive:
			got = append(got, s.Key)
		case config.NotSensitive:
		default:
			t.Errorf("%s: mark = %v, want a named mark", s.Key, s.Mark)
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

// envOnlyMarks pins the mark of every key only the environment sets,
// which changes only at a restart: Sensitive marks a path, an address,
// a credential, or a trust boundary, a change that needs care.
var envOnlyMarks = map[string]config.Sensitivity{
	"DUBPLATE_LIBRARY_ROOT":               config.Sensitive,
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
}

func TestProductionRegistry_EnvironmentOnlyKeysCarryTheirMarks(t *testing.T) {
	reg := newProductionRegistry()
	got := make(map[string]config.Sensitivity)
	described := make(map[string]string)
	add := func(name, description string, mark config.Sensitivity) {
		got[name] = mark
		described[name] = description
	}
	for _, row := range reg.OwnedWall() {
		add(row.Name, row.Description, row.Mark)
	}
	for _, th := range reg.Thresholds() {
		add(th.Key, th.Description, th.Mark)
	}
	for _, src := range reg.IdentitySources() {
		for _, k := range slices.Concat(src.Required, src.Optional) {
			add(k.Name, k.Description, k.Mark)
		}
	}
	for _, class := range reg.IntegrationClasses() {
		for _, k := range slices.Concat([]config.KeyDecl{class.Flag}, class.Required) {
			add(k.Name, k.Description, k.Mark)
		}
	}
	if !maps.Equal(got, envOnlyMarks) {
		for name, mark := range got {
			if want, ok := envOnlyMarks[name]; !ok || want != mark {
				t.Errorf("%s: mark = %v, pinned %v (pinned: %v)", name, mark, want, ok)
			}
		}
		for name := range envOnlyMarks {
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
		if row.Name != "DUBPLATE_ADMIN_EMAIL" {
			continue
		}
		if !strings.Contains(row.Description, "seed") {
			t.Errorf("description = %q, want it to name the seed role", row.Description)
		}
		return
	}
	t.Fatal("DUBPLATE_ADMIN_EMAIL is not an owned wall row")
}
