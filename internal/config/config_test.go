package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

// baseEnviron returns a minimal environ with every owned-now key set
// to a valid value, so a test can override or add exactly the entries
// its case needs.
func baseEnviron() []string {
	return []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
	}
}

func TestLoad_ContractTable(t *testing.T) {
	tests := []struct {
		name       string
		environ    []string
		wantErr    bool
		errContain []string // every substring must appear in the joined error
		check      func(t *testing.T, cfg *Config)
	}{
		{
			name:    "all present",
			environ: baseEnviron(),
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.LibraryRoot != "/srv/music/library" {
					t.Errorf("LibraryRoot = %q, want /srv/music/library", cfg.LibraryRoot)
				}
				if cfg.StateDir != "/var/lib/dubplate" {
					t.Errorf("StateDir = %q, want /var/lib/dubplate", cfg.StateDir)
				}
				if cfg.SiteName != "Test Family" {
					t.Errorf("SiteName = %q, want Test Family", cfg.SiteName)
				}
				if cfg.AdminEmail != "admin@example.org" {
					t.Errorf("AdminEmail = %q, want admin@example.org", cfg.AdminEmail)
				}
			},
		},
		{
			name: "one missing",
			environ: []string{
				"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
				"DUBPLATE_STATE_DIR=/var/lib/dubplate",
				"DUBPLATE_SITE_NAME=Test Family",
			},
			wantErr:    true,
			errContain: []string{"DUBPLATE_ADMIN_EMAIL"},
		},
		{
			name: "several missing",
			environ: []string{
				"DUBPLATE_SITE_NAME=Test Family",
			},
			wantErr: true,
			errContain: []string{
				"DUBPLATE_LIBRARY_ROOT",
				"DUBPLATE_STATE_DIR",
				"DUBPLATE_ADMIN_EMAIL",
			},
		},
		{
			name: "bad email",
			environ: append(dropKey(baseEnviron(), "DUBPLATE_ADMIN_EMAIL"),
				"DUBPLATE_ADMIN_EMAIL=not-an-address"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_ADMIN_EMAIL"},
		},
		{
			name: "email normalized",
			environ: append(dropKey(baseEnviron(), "DUBPLATE_ADMIN_EMAIL"),
				"DUBPLATE_ADMIN_EMAIL=Geoff <GEOFF@X.ORG>"),
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.AdminEmail != "geoff@x.org" {
					t.Errorf("AdminEmail = %q, want geoff@x.org", cfg.AdminEmail)
				}
			},
		},
		{
			name:       "reserved key early",
			environ:    append(baseEnviron(), "DUBPLATE_UPLOADS_DIR=x"),
			wantErr:    true,
			errContain: []string{"intake", "rung 7"},
		},
		{
			name:       "deleted, dubplate_navidrome_url",
			environ:    append(baseEnviron(), "DUBPLATE_NAVIDROME_URL=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_NAVIDROME_URL: unknown configuration key"},
		},
		{
			name:       "deleted, dubplate_navidrome_admin_user",
			environ:    append(baseEnviron(), "DUBPLATE_NAVIDROME_ADMIN_USER=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_NAVIDROME_ADMIN_USER: unknown configuration key"},
		},
		{
			name:       "deleted, dubplate_navidrome_admin_password",
			environ:    append(baseEnviron(), "DUBPLATE_NAVIDROME_ADMIN_PASSWORD=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_NAVIDROME_ADMIN_PASSWORD: unknown configuration key"},
		},
		{
			name:       "deleted, dubplate_ping_navidrome_url",
			environ:    append(baseEnviron(), "DUBPLATE_PING_NAVIDROME_URL=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_PING_NAVIDROME_URL: unknown configuration key"},
		},
		{
			name:       "retired, dubplate_public_navidrome_url",
			environ:    append(baseEnviron(), "DUBPLATE_PUBLIC_NAVIDROME_URL=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_PUBLIC_NAVIDROME_URL: retired: dubplate integrates with no external streamer"},
		},
		{
			name:       "retired, dubplate_navidrome_db",
			environ:    append(baseEnviron(), "DUBPLATE_NAVIDROME_DB=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_NAVIDROME_DB: retired: dubplate integrates with no external streamer"},
		},
		{
			name:       "retired, dubplate_access_team_domain",
			environ:    append(baseEnviron(), "DUBPLATE_ACCESS_TEAM_DOMAIN=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_ACCESS_TEAM_DOMAIN: retired: dubplate delegates sign-in to no external service"},
		},
		{
			name:       "retired, dubplate_access_aud",
			environ:    append(baseEnviron(), "DUBPLATE_ACCESS_AUD=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_ACCESS_AUD: retired: dubplate delegates sign-in to no external service"},
		},
		{
			// Spec criterion 19: each of the six retired backup names
			// fails the load with its message verbatim, one row per name.
			name:    "superseded, r2 account id",
			environ: append(baseEnviron(), "DUBPLATE_R2_ACCOUNT_ID=x"),
			wantErr: true,
			errContain: []string{"DUBPLATE_R2_ACCOUNT_ID: superseded by DUBPLATE_BACKUP_REMOTE; " +
				"configure the remote in rclone's RCLONE_CONFIG_<NAME>_* environment form"},
		},
		{
			name:    "superseded, r2 bucket",
			environ: append(baseEnviron(), "DUBPLATE_R2_BUCKET=x"),
			wantErr: true,
			errContain: []string{"DUBPLATE_R2_BUCKET: superseded by DUBPLATE_BACKUP_REMOTE; " +
				"configure the remote in rclone's RCLONE_CONFIG_<NAME>_* environment form"},
		},
		{
			name:    "superseded, r2 remote",
			environ: append(baseEnviron(), "DUBPLATE_R2_REMOTE=x"),
			wantErr: true,
			errContain: []string{"DUBPLATE_R2_REMOTE: superseded by DUBPLATE_BACKUP_REMOTE; " +
				"configure the remote in rclone's RCLONE_CONFIG_<NAME>_* environment form"},
		},
		{
			name:    "superseded, r2 access key id",
			environ: append(baseEnviron(), "DUBPLATE_R2_ACCESS_KEY_ID=x"),
			wantErr: true,
			errContain: []string{"DUBPLATE_R2_ACCESS_KEY_ID: superseded by DUBPLATE_BACKUP_REMOTE; " +
				"configure the remote in rclone's RCLONE_CONFIG_<NAME>_* environment form"},
		},
		{
			name:    "superseded, r2 secret access key",
			environ: append(baseEnviron(), "DUBPLATE_R2_SECRET_ACCESS_KEY=x"),
			wantErr: true,
			errContain: []string{"DUBPLATE_R2_SECRET_ACCESS_KEY: superseded by DUBPLATE_BACKUP_REMOTE; " +
				"configure the remote in rclone's RCLONE_CONFIG_<NAME>_* environment form"},
		},
		{
			name:    "superseded, backup staging",
			environ: append(baseEnviron(), "DUBPLATE_BACKUP_STAGING=x"),
			wantErr: true,
			errContain: []string{"DUBPLATE_BACKUP_STAGING: superseded: backup staging now lives " +
				"in the store leaf under DUBPLATE_STATE_DIR"},
		},
		{
			name:       "superseded rename, library",
			environ:    append(baseEnviron(), "DUBPLATE_LIBRARY_DIR=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_LIBRARY_ROOT"},
		},
		{
			name:       "superseded rename, site name",
			environ:    append(baseEnviron(), "DUBPLATE_SERVE_SITE_NAME=x"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_SITE_NAME"},
		},
		{
			name:       "unknown key",
			environ:    append(baseEnviron(), "DUBPLATE_FROBNICATE=1"),
			wantErr:    true,
			errContain: []string{"DUBPLATE_FROBNICATE", "unknown"},
		},
		{
			name:    "non-namespace noise ignored",
			environ: append(baseEnviron(), "PATH=/usr/bin", "HOME=/home/geoff"),
			wantErr: false,
		},
		{
			name: "empty value treated as missing",
			environ: append(dropKey(baseEnviron(), "DUBPLATE_STATE_DIR"),
				"DUBPLATE_STATE_DIR="),
			wantErr:    true,
			errContain: []string{"DUBPLATE_STATE_DIR"},
		},
		{
			name: "duplicate key, last wins",
			environ: append(dropKey(baseEnviron(), "DUBPLATE_STATE_DIR"),
				"DUBPLATE_STATE_DIR=/first", "DUBPLATE_STATE_DIR=/second"),
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.StateDir != "/second" {
					t.Errorf("StateDir = %q, want /second (last wins)", cfg.StateDir)
				}
			},
		},
		{
			name: "equals sign in value",
			environ: append(dropKey(baseEnviron(), "DUBPLATE_SITE_NAME"),
				"DUBPLATE_SITE_NAME=A=B"),
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.SiteName != "A=B" {
					t.Errorf("SiteName = %q, want A=B", cfg.SiteName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.environ, NewRegistry())

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() error = nil, want error containing %v", tt.errContain)
				}
				for _, sub := range tt.errContain {
					if !strings.Contains(err.Error(), sub) {
						t.Errorf("Load() error = %q, want it to contain %q", err.Error(), sub)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

// dropKey returns a copy of environ with any entry named name removed,
// so a test case can replace an owned key's value without producing a
// duplicate.
func dropKey(environ []string, name string) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		n, _, _ := strings.Cut(kv, "=")
		if n == name {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// classificationRegistry returns a registry carrying one fixture
// declaration in every family the wall classifies a DUBPLATE_* name
// against, plus a pending owned row. Every name is this test's own; no
// rung's real key appears.
func classificationRegistry() *Registry {
	reg := NewRegistry()
	reg.AddWallRow(WallRow{Name: "DUBPLATE_CLASSIFY_OWNED"})
	reg.AddWallRow(WallRow{Name: "DUBPLATE_CLASSIFY_PENDING", Pending: "pending (arrives later)"})
	reg.AddReservedClass(ReservedClass{
		Name: "classification fixture",
		Rung: 99,
		Keys: []string{"DUBPLATE_CLASSIFY_RESERVED"},
	})
	reg.AddSuperseded("DUBPLATE_CLASSIFY_SUPERSEDED", "renamed to DUBPLATE_CLASSIFY_OWNED")
	reg.AddIntegrationClass(IntegrationClass{
		Name:     "classification integration",
		Rung:     99,
		Flag:     "DUBPLATE_CLASSIFY_ENABLED",
		Required: []string{"DUBPLATE_CLASSIFY_REQUIRED"},
	})
	reg.AddThreshold(Threshold{
		Key:     "DUBPLATE_CLASSIFY_THRESHOLD",
		Kind:    KindInt,
		Default: "1",
		Doc:     "an integer",
	})
	reg.AddIdentitySource(IdentitySource{
		Name:     "classification source",
		Selector: "DUBPLATE_CLASSIFY_SELECTOR",
		Required: []string{"DUBPLATE_CLASSIFY_SOURCE_KEY"},
	})
	reg.AddSetting(Setting{
		Key:     "DUBPLATE_CLASSIFY_SETTING",
		Kind:    KindInt,
		Default: "14",
		Doc:     "an integer",
		Apply:   ApplyNextOperation,
		Floor:   "3",
		Ceiling: "1000",
	})
	return reg
}

// classificationEnviron satisfies every row classificationRegistry
// requires: the base wall, its own owned-now row, and the selector
// with the declared source's required key. Load requires every
// owned-now row a registry carries and the selector of any declared
// source, so a test that wants Load to succeed against this registry
// has to satisfy both.
func classificationEnviron() []string {
	return append(baseEnviron(),
		"DUBPLATE_CLASSIFY_OWNED=/srv/classify",
		"DUBPLATE_CLASSIFY_SELECTOR=classification source",
		"DUBPLATE_CLASSIFY_SOURCE_KEY=classify-value",
	)
}

// TestClassificationFamilies proves the seven classification helpers
// rejectNamespace consults each answer for their own family's names
// and for no other's, and that a name declared in none of them
// classifies as unknown.
//
// This package tests the mechanics alone. It cannot test them against
// a real environment template, and the assertion that once did was
// moved to cmd/dubplate-server: under the rung 1b registry contract
// each rung declares its keys in its own file, so internal/config's
// base registry can no longer see a name rung 4 or rung 5 declares,
// and only the composed production registry can classify a template
// without reporting a declared key as unknown.
func TestClassificationFamilies(t *testing.T) {
	const (
		owned       = "owned"
		pending     = "pending"
		reserved    = "reserved"
		superseded  = "superseded"
		integration = "integration"
		threshold   = "threshold"
		setting     = "setting"
		source      = "source"
		unknown     = "unknown"
	)

	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "owned-now row", key: "DUBPLATE_CLASSIFY_OWNED", want: owned},
		{name: "pending owned row", key: "DUBPLATE_CLASSIFY_PENDING", want: pending},
		{name: "reserved class key", key: "DUBPLATE_CLASSIFY_RESERVED", want: reserved},
		{name: "superseded name", key: "DUBPLATE_CLASSIFY_SUPERSEDED", want: superseded},
		{name: "integration flag", key: "DUBPLATE_CLASSIFY_ENABLED", want: integration},
		{name: "integration required key", key: "DUBPLATE_CLASSIFY_REQUIRED", want: integration},
		{name: "registered threshold", key: "DUBPLATE_CLASSIFY_THRESHOLD", want: threshold},
		{name: "store-homed setting", key: "DUBPLATE_CLASSIFY_SETTING", want: setting},
		{name: "identity source selector", key: "DUBPLATE_CLASSIFY_SELECTOR", want: source},
		{name: "identity source required key", key: "DUBPLATE_CLASSIFY_SOURCE_KEY", want: source},
		{name: "declared nowhere", key: "DUBPLATE_CLASSIFY_ABSENT", want: unknown},
	}

	reg := classificationRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each helper is asserted on its own, so a row states one
			// true answer and seven false ones rather than echoing
			// rejectNamespace's own branch order back at itself.
			_, isReserved := reg.reservedClassOf(tt.key)
			_, isSuperseded := reg.supersessionOf(tt.key)
			_, isIntegration := reg.integrationClassOf(tt.key)
			got := map[string]bool{
				owned:       reg.ownedNow(tt.key),
				pending:     wallRowPending(reg, tt.key),
				reserved:    isReserved,
				superseded:  isSuperseded,
				integration: isIntegration,
				threshold:   thresholdOf(reg, tt.key),
				setting:     settingOf(reg, tt.key),
				source:      reg.identitySourceKey(tt.key),
			}
			for family, reported := range got {
				if reported != (family == tt.want) {
					t.Errorf("%s classified as %s = %t, want %t", tt.key, family, reported, family == tt.want)
				}
			}
		})
	}
}

// wallRowPending reports whether key is an owned wall row carrying a
// Pending status, the one classification with no helper of its own:
// ownedNow answers false for a pending row and for an undeclared name
// alike, so the test above needs this to tell the two apart.
func wallRowPending(reg *Registry, key string) bool {
	for _, row := range reg.OwnedWall() {
		if row.Name == key {
			return row.Pending != ""
		}
	}
	return false
}

// TestLoad_RegisteredThresholdIsNotUnknown proves the fifth
// classification branch: a threshold a rung declares in its own file
// is a declared name, so setting it loads cleanly rather than erroring
// as an unknown key. Without the branch a binary would refuse the very
// tunable it registered. The negative half is the same registry with
// the threshold's name changed by one character.
func TestLoad_RegisteredThresholdIsNotUnknown(t *testing.T) {
	reg := classificationRegistry()

	if _, err := Load(append(classificationEnviron(), "DUBPLATE_CLASSIFY_THRESHOLD=7"), reg); err != nil {
		t.Errorf("Load() error = %v, want nil for a registered threshold", err)
	}

	_, err := Load(append(classificationEnviron(), "DUBPLATE_CLASSIFY_THRESHOLDX=7"), reg)
	if err == nil {
		t.Fatal("Load() error = nil for an undeclared key, want an unknown-key error")
	}
	if !strings.Contains(err.Error(), "unknown configuration key") {
		t.Errorf("Load() error = %v, want it to name the key as unknown", err)
	}
}

// TestLoad_StoreHomedSettingIsDeclaredNeverParsed proves Load's
// classification of a store-homed key: whatever its env value, Load
// accepts it as declared and does not parse it, because the value is
// only a seed and the seed evaluation is where it is parsed and
// bounded. The last row is the negative half: the same registry with
// the setting's name changed by one character.
func TestLoad_StoreHomedSettingIsDeclaredNeverParsed(t *testing.T) {
	tests := []struct {
		name    string
		entry   string // appended to classificationEnviron; empty adds nothing
		wantErr string // empty: Load succeeds
	}{
		{name: "unset", entry: ""},
		{name: "set but empty", entry: "DUBPLATE_CLASSIFY_SETTING="},
		{name: "set, parsing, in bounds", entry: "DUBPLATE_CLASSIFY_SETTING=14"},
		{name: "set, parsing, below the floor", entry: "DUBPLATE_CLASSIFY_SETTING=1"},
		{name: "set, not parsing", entry: "DUBPLATE_CLASSIFY_SETTING=fourteen"},
		{name: "an undeclared near-miss", entry: "DUBPLATE_CLASSIFY_SETTINGX=14", wantErr: "DUBPLATE_CLASSIFY_SETTINGX: unknown configuration key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := classificationEnviron()
			if tt.entry != "" {
				environ = append(environ, tt.entry)
			}
			_, err := Load(environ, classificationRegistry())
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Load() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// fixtureIntegrationClass is a fixture integration class this test
// declares for itself. It drives Load's generic required-when-enabled
// resolution (Resolution R5).
var fixtureIntegrationClass = IntegrationClass{
	Name: "fixture",
	Flag: "DUBPLATE_FIXTURE_ENABLED",
	Required: []string{
		"DUBPLATE_FIXTURE_REQUIRED_A",
		"DUBPLATE_FIXTURE_REQUIRED_B",
	},
}

// fixtureNoRequiredClass is an integration declaring no Required key
// at all. Its enabled state must be reachable independently of whether
// Load resolved any key.
var fixtureNoRequiredClass = IntegrationClass{
	Name: "fixture-no-required",
	Flag: "DUBPLATE_FIXTURE_NO_REQUIRED_ENABLED",
}

// fixtureRegistry returns a registry seeded with rung 1's own
// declarations plus the two fixture integration classes above.
func fixtureRegistry() *Registry {
	reg := NewRegistry()
	reg.AddIntegrationClass(fixtureIntegrationClass)
	reg.AddIntegrationClass(fixtureNoRequiredClass)
	return reg
}

// TestLoadIntegrationRequiredWhenEnabled proves Load's generic
// required-when-enabled resolution, reachable by class name: a
// required key set with the flag unset is a load error naming the
// flag, every required key is reported at once when the flag is set
// and resolves, and an optional key is accepted and never required.
func TestLoadIntegrationRequiredWhenEnabled(t *testing.T) {
	requiredKeys := []string{
		"DUBPLATE_FIXTURE_REQUIRED_A=a-value",
		"DUBPLATE_FIXTURE_REQUIRED_B=b-value",
	}

	tests := []struct {
		name        string
		extra       []string
		wantErr     []string // substrings every one of which must appear
		wantEnabled bool
	}{
		{
			name:  "flag unset, no fixture keys: loads, not enabled",
			extra: nil,
		},
		{
			name:    "a required key set with the flag unset names the flag",
			extra:   []string{"DUBPLATE_FIXTURE_REQUIRED_A=a-value"},
			wantErr: []string{"DUBPLATE_FIXTURE_REQUIRED_A", "DUBPLATE_FIXTURE_ENABLED"},
		},
		{
			name:    "flag set with no required keys reports every missing row at once",
			extra:   []string{"DUBPLATE_FIXTURE_ENABLED=true"},
			wantErr: []string{"DUBPLATE_FIXTURE_REQUIRED_A", "DUBPLATE_FIXTURE_REQUIRED_B"},
		},
		{
			name:        "flag set with every required key resolves, reachable by class name",
			extra:       append([]string{"DUBPLATE_FIXTURE_ENABLED=true"}, requiredKeys...),
			wantEnabled: true,
		},
		{
			name:    "an unparseable flag produces exactly one error",
			extra:   []string{"DUBPLATE_FIXTURE_ENABLED=maybe"},
			wantErr: []string{"DUBPLATE_FIXTURE_ENABLED"},
		},
		{
			name:  "an explicit false loads clean, not enabled",
			extra: []string{"DUBPLATE_FIXTURE_ENABLED=false"},
		},
		{
			name:    "an explicit false with a required key still names the flag",
			extra:   []string{"DUBPLATE_FIXTURE_ENABLED=false", "DUBPLATE_FIXTURE_REQUIRED_A=a-value"},
			wantErr: []string{"DUBPLATE_FIXTURE_REQUIRED_A", "DUBPLATE_FIXTURE_ENABLED"},
		},
	}

	reg := fixtureRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(append(slices.Clone(baseEnviron()), tt.extra...), reg)
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if got := cfg.IntegrationEnabled(fixtureIntegrationClass.Name); got != tt.wantEnabled {
					t.Fatalf("IntegrationEnabled = %v, want %v", got, tt.wantEnabled)
				}
				if tt.wantEnabled && cfg.IntegrationValue(fixtureIntegrationClass.Name, "DUBPLATE_FIXTURE_REQUIRED_A") != "a-value" {
					t.Errorf("IntegrationValue = %q, want %q", cfg.IntegrationValue(fixtureIntegrationClass.Name, "DUBPLATE_FIXTURE_REQUIRED_A"), "a-value")
				}
				return
			}
			if err == nil {
				t.Fatal("Load returned nil; want an error")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error does not name %s: %v", want, err)
				}
			}
			if tt.name == "an unparseable flag produces exactly one error" {
				if n := strings.Count(err.Error(), "DUBPLATE_FIXTURE_ENABLED"); n != 1 {
					t.Errorf("error names DUBPLATE_FIXTURE_ENABLED %d times, want exactly 1: %v", n, err)
				}
			}
		})
	}
}

// TestLoadIntegrationEnabledIndependentOfResolvedKeys proves a class
// with an empty Required set still reports enabled when its flag is
// set, and reports not enabled when its flag is unset. A class with no
// required key must still be told apart from a disabled one.
func TestLoadIntegrationEnabledIndependentOfResolvedKeys(t *testing.T) {
	tests := []struct {
		name        string
		extra       []string
		wantEnabled bool
	}{
		{
			name: "flag unset: not enabled",
		},
		{
			name:        "flag set on an empty-Required class still reports enabled",
			extra:       []string{"DUBPLATE_FIXTURE_NO_REQUIRED_ENABLED=true"},
			wantEnabled: true,
		},
	}

	reg := fixtureRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(append(slices.Clone(baseEnviron()), tt.extra...), reg)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := cfg.IntegrationEnabled(fixtureNoRequiredClass.Name); got != tt.wantEnabled {
				t.Errorf("IntegrationEnabled = %v, want %v", got, tt.wantEnabled)
			}
		})
	}
}

// The fixture identity source below is this package's own: every name
// is a DUBPLATE_TEST_SCRATCH_* scratch name, on
// TestDoctorCmd_LineCountFollowsRegistry's precedent, so the config
// core's own tests carry no rung's key.
const (
	scratchSelector   = "DUBPLATE_TEST_SCRATCH_SOURCE"
	scratchRequiredA  = "DUBPLATE_TEST_SCRATCH_A_REQUIRED"
	scratchOptionalA  = "DUBPLATE_TEST_SCRATCH_A_OPTIONAL"
	scratchRequiredB  = "DUBPLATE_TEST_SCRATCH_B_REQUIRED"
	scratchSourceA    = "scratch-a"
	scratchSourceB    = "scratch-b"
	scratchRefuseWord = "refuse"
)

// scratchWindow is a threshold the fixture hook reads back out of the
// environment snapshot. It stands in for R31's real pair: a threshold
// resolves at its point of use and Load never reads one, so a hook
// handed the resolved source values alone could not see it.
var scratchWindow = Threshold{
	Key:     "DUBPLATE_TEST_SCRATCH_WINDOW",
	Kind:    KindDuration,
	Default: "1h",
	Doc:     "a duration",
}

// scratchValidate is the fixture source's validation hook: it refuses
// one literal value of its required key, and it refuses a window
// shorter than an hour, which it can only learn from the environment
// snapshot.
func scratchValidate(values map[string]string, environ []string) error {
	var errs []error
	if values[scratchRequiredA] == scratchRefuseWord {
		errs = append(errs, fmt.Errorf("%s: value refused by the source", scratchRequiredA))
	}
	window, err := ParseDuration(environ, scratchWindow)
	if err != nil {
		errs = append(errs, err)
	} else if window < time.Hour {
		errs = append(errs, fmt.Errorf("%s: below the one-hour floor", scratchWindow.Key))
	}
	return errors.Join(errs...)
}

// scratchSourceRegistry returns a registry declaring two fixture
// identity sources on one selector, plus the threshold the first
// source's hook reads.
func scratchSourceRegistry() *Registry {
	reg := NewRegistry()
	reg.AddThreshold(scratchWindow)
	reg.AddIdentitySource(IdentitySource{
		Name:     scratchSourceA,
		Selector: scratchSelector,
		Required: []string{scratchRequiredA},
		Optional: []string{scratchOptionalA},
		Validate: scratchValidate,
	})
	reg.AddIdentitySource(IdentitySource{
		Name:     scratchSourceB,
		Selector: scratchSelector,
		Required: []string{scratchRequiredB},
	})
	return reg
}

// TestLoad_RequiresEveryRegisteredOwnedNowRow proves the requirement
// mechanism (R-5b4): a wall row a rung registers from its own file is
// required by the registration itself, and a row carrying Pending is
// neither required nor read. Rung 1's four rows are unaffected because
// they are the rows NewRegistry seeds; the rows here are this test's
// own.
func TestLoad_RequiresEveryRegisteredOwnedNowRow(t *testing.T) {
	const (
		scratchOwned   = "DUBPLATE_TEST_SCRATCH_OWNED"
		scratchPending = "DUBPLATE_TEST_SCRATCH_PENDING"
	)

	registry := func() *Registry {
		reg := NewRegistry()
		reg.AddWallRow(WallRow{Name: scratchOwned})
		reg.AddWallRow(WallRow{Name: scratchPending, Pending: "pending (arrives later)"})
		return reg
	}

	tests := []struct {
		name     string
		extra    []string
		wantErr  []string
		wantOnce string
	}{
		{
			name:     "a registered owned-now row unset is required, named once",
			wantErr:  []string{scratchOwned, "required setting unset"},
			wantOnce: scratchOwned,
		},
		{
			name:  "a registered owned-now row set loads clean, with the pending row unset",
			extra: []string{scratchOwned + "=/srv/scratch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(append(slices.Clone(baseEnviron()), tt.extra...), registry())

			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Load() error = nil, want one naming %v", tt.wantErr)
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Load() error = %q, want it to contain %q", err.Error(), want)
				}
			}
			if n := strings.Count(err.Error(), tt.wantOnce); n != 1 {
				t.Errorf("Load() error names %s %d times, want exactly 1: %v", tt.wantOnce, n, err)
			}
		})
	}
}

// TestLoad_IdentitySource proves the per-source declaration family
// end to end against fixture sources: the selector is unrequired with
// none declared and required with one, a value naming no declared
// source is refused by name, the selected source's required keys are
// required and its optional keys are read, an unselected source's keys
// are declared rather than unknown, and the validation hook's errors
// join the rest.
func TestLoad_IdentitySource(t *testing.T) {
	selected := []string{scratchSelector + "=" + scratchSourceA, scratchRequiredA + "=a-value"}

	tests := []struct {
		name       string
		registry   func() *Registry
		drop       string // one base row to unset, for a row proving the hook's errors join
		extra      []string
		wantErr    []string
		wantNotErr []string
		check      func(t *testing.T, cfg *Config)
	}{
		{
			name:     "no source declared: the selector is not required",
			registry: NewRegistry,
			check: func(t *testing.T, cfg *Config) {
				if cfg.IdentitySource != "" {
					t.Errorf("IdentitySource = %q, want empty with no source declared", cfg.IdentitySource)
				}
			},
		},
		{
			name:     "one source declared: the selector is required",
			registry: scratchSourceRegistry,
			wantErr:  []string{scratchSelector, "required setting unset"},
		},
		{
			name:     "a selector naming no declared source lists the declared names",
			registry: scratchSourceRegistry,
			extra:    []string{scratchSelector + "=nope"},
			wantErr:  []string{`unknown identity source "nope"`, "scratch-a, scratch-b"},
		},
		{
			name:     "the selected source's required key is required",
			registry: scratchSourceRegistry,
			extra:    []string{scratchSelector + "=" + scratchSourceA},
			wantErr:  []string{scratchRequiredA, "required setting unset"},
		},
		{
			name:     "an optional key unset loads clean",
			registry: scratchSourceRegistry,
			extra:    selected,
			check: func(t *testing.T, cfg *Config) {
				if cfg.IdentitySource != scratchSourceA {
					t.Errorf("IdentitySource = %q, want %q", cfg.IdentitySource, scratchSourceA)
				}
				if got := cfg.IdentitySourceValue(scratchRequiredA); got != "a-value" {
					t.Errorf("IdentitySourceValue(%s) = %q, want %q", scratchRequiredA, got, "a-value")
				}
				if got := cfg.IdentitySourceValue(scratchOptionalA); got != "" {
					t.Errorf("IdentitySourceValue(%s) = %q, want empty when unset", scratchOptionalA, got)
				}
			},
		},
		{
			name:     "an optional key set is readable from Config",
			registry: scratchSourceRegistry,
			extra:    append(slices.Clone(selected), scratchOptionalA+"=opt-value"),
			check: func(t *testing.T, cfg *Config) {
				if got := cfg.IdentitySourceValue(scratchOptionalA); got != "opt-value" {
					t.Errorf("IdentitySourceValue(%s) = %q, want %q", scratchOptionalA, got, "opt-value")
				}
			},
		},
		{
			name:     "an unselected source's key is declared, never required and never unknown",
			registry: scratchSourceRegistry,
			extra:    append(slices.Clone(selected), scratchRequiredB+"=b-value"),
			check: func(t *testing.T, cfg *Config) {
				if got := cfg.IdentitySourceValue(scratchRequiredB); got != "" {
					t.Errorf("IdentitySourceValue(%s) = %q, want empty: the source is not selected", scratchRequiredB, got)
				}
			},
		},
		{
			name:       "a source key set with the selector unset is declared, not unknown",
			registry:   scratchSourceRegistry,
			extra:      []string{scratchRequiredA + "=a-value"},
			wantErr:    []string{scratchSelector},
			wantNotErr: []string{"unknown configuration key"},
		},
		{
			name:     "the hook refuses a resolved value, joined with every other failure",
			registry: scratchSourceRegistry,
			drop:     "DUBPLATE_SITE_NAME",
			extra:    []string{scratchSelector + "=" + scratchSourceA, scratchRequiredA + "=" + scratchRefuseWord},
			wantErr:  []string{"value refused by the source", "DUBPLATE_SITE_NAME"},
		},
		{
			name:     "the hook refuses a registered threshold's value it reads from the snapshot",
			registry: scratchSourceRegistry,
			drop:     "DUBPLATE_SITE_NAME",
			extra: []string{scratchSelector + "=" + scratchSourceA, scratchRequiredA + "=a-value",
				scratchWindow.Key + "=1m"},
			wantErr: []string{scratchWindow.Key + ": below the one-hour floor", "DUBPLATE_SITE_NAME"},
		},
		{
			name:     "an unselected source's hook never runs",
			registry: scratchSourceRegistry,
			extra: []string{scratchSelector + "=" + scratchSourceB, scratchRequiredB + "=b-value",
				scratchWindow.Key + "=1m"},
			check: func(t *testing.T, cfg *Config) {
				if cfg.IdentitySource != scratchSourceB {
					t.Errorf("IdentitySource = %q, want %q", cfg.IdentitySource, scratchSourceB)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := baseEnviron()
			if tt.drop != "" {
				base = dropKey(base, tt.drop)
			}
			cfg, err := Load(append(base, tt.extra...), tt.registry())

			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load() error = %v, want nil", err)
				}
				if tt.check != nil {
					tt.check(t, cfg)
				}
				return
			}
			if err == nil {
				t.Fatalf("Load() error = nil, want one naming %v", tt.wantErr)
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Load() error = %q, want it to contain %q", err.Error(), want)
				}
			}
			for _, unwanted := range tt.wantNotErr {
				if strings.Contains(err.Error(), unwanted) {
					t.Errorf("Load() error = %q, want it not to contain %q", err.Error(), unwanted)
				}
			}
		})
	}
}

// TestLoad_SiteNameGrammar covers DUBPLATE_SITE_NAME's grammar: the
// value enters every served document and the page title, so it must be
// valid UTF-8 with no control character, no bidirectional override or
// isolate, and at most siteNameMaxBytes bytes. A refusal names the key
// and never echoes the value.
func TestLoad_SiteNameGrammar(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr string
	}{
		{name: "plain", value: "The Hendersons"},
		{name: "non-ASCII letters", value: "Café Ñandú 音楽"},
		{name: "exactly the byte limit", value: strings.Repeat("a", siteNameMaxBytes)},
		{name: "one byte over", value: strings.Repeat("a", siteNameMaxBytes+1), wantErr: "over 100 bytes"},
		{name: "invalid UTF-8", value: "Group \xff", wantErr: "not valid UTF-8"},
		{name: "a C0 control", value: "Group\x07", wantErr: "a control character"},
		{name: "a tab", value: "Group\tName", wantErr: "a control character"},
		{name: "DEL", value: "Group\x7f", wantErr: "a control character"},
		{name: "a C1 control", value: "Group\u0085", wantErr: "a control character"},
		{name: "a right-to-left override", value: "Group\u202e", wantErr: "a bidirectional override"},
		{name: "a first-strong isolate", value: "Group\u2068", wantErr: "a bidirectional override"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(dropKey(baseEnviron(), "DUBPLATE_SITE_NAME"), "DUBPLATE_SITE_NAME="+tt.value)
			cfg, err := Load(environ, NewRegistry())
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if cfg.SiteName != tt.value {
					t.Errorf("SiteName = %q, want %q", cfg.SiteName, tt.value)
				}
				return
			}
			if err == nil {
				t.Fatalf("Load: got nil error, want one containing %q", tt.wantErr)
			}
			msg := err.Error()
			if !strings.Contains(msg, "DUBPLATE_SITE_NAME") || !strings.Contains(msg, tt.wantErr) {
				t.Errorf("Load error = %q, want it to name DUBPLATE_SITE_NAME and contain %q", msg, tt.wantErr)
			}
			if strings.Contains(msg, tt.value) {
				t.Errorf("Load error = %q echoes the refused value", msg)
			}
		})
	}
}

// TestLoad_UnparseableFlagIsNotSetWithout proves a class key set beside
// a flag that is present but unparseable is not reported as set without
// the flag: the flag is set, and its own invalid-value error is the one
// report the operator needs.
func TestLoad_UnparseableFlagIsNotSetWithout(t *testing.T) {
	environ := append(slices.Clone(baseEnviron()),
		"DUBPLATE_FIXTURE_ENABLED=maybe",
		"DUBPLATE_FIXTURE_REQUIRED_A=a-value",
	)
	_, err := Load(environ, fixtureRegistry())
	if err == nil {
		t.Fatal("Load returned nil; want the flag's invalid-value error")
	}
	const want = `DUBPLATE_FIXTURE_ENABLED: invalid value "maybe", want a boolean`
	if err.Error() != want {
		t.Errorf("Load error = %q, want exactly %q", err, want)
	}
}

// TestLoad_RefusalOrderIsSorted proves the namespace errors in Load's
// joined refusal are listed in sorted key order on every run, whatever
// order the environment and Go's map iteration present them in. Twenty
// runs over five keys would all agree by chance with probability far
// below one in a million under randomized map order.
func TestLoad_RefusalOrderIsSorted(t *testing.T) {
	environ := append(slices.Clone(baseEnviron()),
		"DUBPLATE_ZULU=1",
		"DUBPLATE_FIXTURE_REQUIRED_B=b",
		"DUBPLATE_ALPHA=1",
		"DUBPLATE_MIKE=1",
		"DUBPLATE_BRAVO=1",
	)
	const want = "DUBPLATE_ALPHA: unknown configuration key\n" +
		"DUBPLATE_BRAVO: unknown configuration key\n" +
		"DUBPLATE_FIXTURE_REQUIRED_B: set without DUBPLATE_FIXTURE_ENABLED, the fixture integration's enable flag\n" +
		"DUBPLATE_MIKE: unknown configuration key\n" +
		"DUBPLATE_ZULU: unknown configuration key"
	reg := fixtureRegistry()
	for run := range 20 {
		_, err := Load(environ, reg)
		if err == nil {
			t.Fatalf("run %d: Load returned nil; want the joined refusal", run)
		}
		if err.Error() != want {
			t.Fatalf("run %d: Load error =\n%s\nwant\n%s", run, err, want)
		}
	}
}
