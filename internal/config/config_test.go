package config

import (
	"slices"
	"strings"
	"testing"
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
			environ:    append(baseEnviron(), "DUBPLATE_R2_BUCKET=x"),
			wantErr:    true,
			errContain: []string{"backup targets", "rung 6"},
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
		Optional: []string{"DUBPLATE_CLASSIFY_OPTIONAL"},
	})
	reg.AddThreshold(Threshold{
		Key:     "DUBPLATE_CLASSIFY_THRESHOLD",
		Kind:    KindInt,
		Default: "1",
		Doc:     "an integer",
	})
	return reg
}

// TestClassificationFamilies proves the five classification helpers
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
		{name: "integration optional key", key: "DUBPLATE_CLASSIFY_OPTIONAL", want: integration},
		{name: "registered threshold", key: "DUBPLATE_CLASSIFY_THRESHOLD", want: threshold},
		{name: "declared nowhere", key: "DUBPLATE_CLASSIFY_ABSENT", want: unknown},
	}

	reg := classificationRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each helper is asserted on its own, so a row states one
			// true answer and five false ones rather than echoing
			// rejectNamespace's own branch order back at itself.
			_, isReserved := reservedClassOf(reg, tt.key)
			_, isSuperseded := supersessionOf(reg, tt.key)
			_, isIntegration := integrationClassOf(reg, tt.key)
			got := map[string]bool{
				owned:       ownedNow(reg, tt.key),
				pending:     wallRowPending(reg, tt.key),
				reserved:    isReserved,
				superseded:  isSuperseded,
				integration: isIntegration,
				threshold:   thresholdOf(reg, tt.key),
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

	if _, err := Load(append(baseEnviron(), "DUBPLATE_CLASSIFY_THRESHOLD=7"), reg); err != nil {
		t.Errorf("Load() error = %v, want nil for a registered threshold", err)
	}

	_, err := Load(append(baseEnviron(), "DUBPLATE_CLASSIFY_THRESHOLDX=7"), reg)
	if err == nil {
		t.Fatal("Load() error = nil for an undeclared key, want an unknown-key error")
	}
	if !strings.Contains(err.Error(), "unknown configuration key") {
		t.Errorf("Load() error = %v, want it to name the key as unknown", err)
	}
}

// fixtureIntegrationClass is a fixture integration class this test
// declares for itself, with no rung-4 name. It drives Load's generic
// required-when-enabled resolution (Resolution R5): the eight-row
// table TestLoadShadowIntegration used to prove moved verbatim to
// cmd/dubplate-server/shadow_test.go, where the shadow class now
// lives, and this test proves the same job generically instead.
var fixtureIntegrationClass = IntegrationClass{
	Name: "fixture",
	Flag: "DUBPLATE_FIXTURE_ENABLED",
	Required: []string{
		"DUBPLATE_FIXTURE_REQUIRED_A",
		"DUBPLATE_FIXTURE_REQUIRED_B",
	},
	Optional: []string{
		"DUBPLATE_FIXTURE_OPTIONAL",
	},
}

// fixtureNoRequiredClass is hook 6's motivating shape: an integration
// declaring no Required key at all, the way a driver declining every
// capability would. Its enabled state must be reachable independently
// of whether Load resolved any key.
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
			name:        "an optional key is accepted and never required",
			extra:       append([]string{"DUBPLATE_FIXTURE_ENABLED=true", "DUBPLATE_FIXTURE_OPTIONAL=x"}, requiredKeys...),
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

// TestLoadIntegrationEnabledIndependentOfResolvedKeys proves hook 6's
// motivating shape: a class with an empty Required set still reports
// enabled when its flag is set, and reports not enabled when its flag
// is unset. A driver declining every capability declares no required
// key and must still be told apart from a disabled one.
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
			name:        "flag set on an empty-Required class still reports enabled (hook 6)",
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
