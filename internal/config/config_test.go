package config

import (
	"bufio"
	"os"
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

// templateEnvPaths names every environment template this package's
// tests classify against, one entry per adopter-facing template
// scripts/deploy.sh renders (ENV_TEMPLATE at deploy.sh:27,
// SERVER_ENV_TEMPLATE at deploy.sh:28, both rendered by render_template
// at deploy.sh:205, called at deploy.sh:240 and deploy.sh:241). A name
// missing from either fails deploy the same way, which is why both are
// read rather than one.
//
// dubplate.env.template is the frozen pre-pivot template, kept so a
// name this package still classifies as superseded stays proven
// reachable from a real file rather than only from supersededKeys'
// own literal table. dubplate-server.env.template is the greenfield
// template (Task 0, landed c8b6d71): the completeness direction below
// reads it by name, and it is also the template carrying the six
// DUBPLATE_SHADOW_* rows that make the classification walk's fourth
// branch mandatory rather than optional.
var templateEnvPaths = []string{
	"../../env/dubplate.env.template",
	"../../env/dubplate-server.env.template",
}

// classificationFixtureIntegrationClass mirrors, by literal key name
// only, rung 4's shadowIntegrationClass (cmd/dubplate-server/shadow.go)
// so this test's classification walk can resolve
// env/dubplate-server.env.template's six DUBPLATE_SHADOW_* rows through
// integrationClassOf's fourth branch (Resolution R10). internal/config
// cannot import cmd/dubplate-server, which imports internal/config
// already, so this test classifies against a registry it builds for
// itself rather than the production registry cmd/dubplate-server
// assembles: NewRegistry() (rung 1's own declarations) plus this one
// fixture class. It declares no name that is not already assigned in
// the tracked template; it only re-states the shape a real integration
// class registers so the generic classification mechanism has
// something to resolve against.
//
// provisional, conductor ruling owed, R10
var classificationFixtureIntegrationClass = IntegrationClass{
	Name: "shadow classification fixture",
	Flag: "DUBPLATE_SHADOW_ENABLED",
	Required: []string{
		"DUBPLATE_SHADOW_OBSERVE_DIR",
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR",
		"DUBPLATE_SHADOW_REVIEW_DIR",
	},
	Optional: []string{
		"DUBPLATE_SHADOW_FLOOR",
		"DUBPLATE_SHADOW_BUDGET",
	},
}

// TestLoad_TemplateKeysClassified proves the template disposition walk
// is complete, in both directions Resolution R10 states.
//
// The classification direction (mechanical, asserted here): every
// DUBPLATE_* name any template under env/ lists must classify as owned,
// reserved, superseded, or as an integration class's Flag, Required key
// or Optional key, never unknown, or a real adopter's template-derived
// env file would trip a load error this package never intended. The
// fourth branch exists because env/dubplate-server.env.template carries
// six DUBPLATE_SHADOW_* rows that classify only through
// integrationClassOf; without it those six report unknown and this
// test reds for every lane. It classifies against NewRegistry() plus
// classificationFixtureIntegrationClass, stated above, since this
// package cannot reach rung 4's own registration.
//
// The completeness direction (mechanical only against a greenfield
// template, asserted here as TestLoad_OwnedNowRowsInGreenfieldTemplate):
// every owned-now row must appear in the template the greenfield server
// ships. It is not asserted against dubplate.env.template, whose own
// header says it describes the frozen pre-pivot binary's wall.
//
// provisional, conductor ruling owed, R10
//
// The rule that cannot be mechanical: a rung that declares a new
// owned-now row adds it to env/dubplate-server.env.template in the same
// pass, with the minting that implies. scripts/deploy.sh renders both
// templates (ENV_TEMPLATE at :27, SERVER_ENV_TEMPLATE at :28,
// render_template at :205, called at :240 and :241) and refuses to
// deploy when render_missing is non-empty (:243-247), so a name missing
// from either template fails deploy the same way; a rung that adds an
// owned-now row without adding it here discovers the gap at deploy time
// rather than at test time. Neither this test nor any other in this
// package may add that name itself: doing so would mint a value outside
// the age store, an owner touchpoint this pass does not take.
func TestLoad_TemplateKeysClassified(t *testing.T) {
	reg := NewRegistry()
	reg.AddIntegrationClass(classificationFixtureIntegrationClass)

	for _, path := range templateEnvPaths {
		names := templateKeys(t, path)
		if len(names) == 0 {
			t.Fatalf("%s: templateKeys found no DUBPLATE_* names; template path or format changed", path)
		}
		t.Logf("%s: classified %d DUBPLATE_* names", path, len(names))

		for _, name := range names {
			if ownedNow(reg, name) {
				continue
			}
			if _, ok := reservedClassOf(reg, name); ok {
				continue
			}
			if _, ok := supersessionOf(reg, name); ok {
				continue
			}
			if _, ok := integrationClassOf(reg, name); ok {
				continue
			}
			t.Errorf("%s: %s classifies as unknown; add it to ownedWall, a reserved class, supersededKeys, or an integration class", path, name)
		}
	}
}

// TestLoad_OwnedNowRowsInGreenfieldTemplate asserts R10's completeness
// direction: every owned-now row (ownedNow reports true, meaning
// [Load] requires it now rather than carrying a Pending status)
// appears in env/dubplate-server.env.template, the greenfield template
// Task 0 confirmed landed at c8b6d71 carrying all four
// (DUBPLATE_LIBRARY_ROOT:36, DUBPLATE_STATE_DIR:53,
// DUBPLATE_SITE_NAME:61, DUBPLATE_ADMIN_EMAIL:66). This test never
// writes the template; a missing row is reported, not repaired.
func TestLoad_OwnedNowRowsInGreenfieldTemplate(t *testing.T) {
	const greenfieldTemplate = "../../env/dubplate-server.env.template"
	names := templateKeys(t, greenfieldTemplate)
	if len(names) == 0 {
		t.Fatalf("%s: templateKeys found no DUBPLATE_* names; template path or format changed", greenfieldTemplate)
	}

	reg := NewRegistry()
	for _, row := range reg.OwnedWall() {
		if !ownedNow(reg, row.Name) {
			continue
		}
		if !slices.Contains(names, row.Name) {
			t.Errorf("%s: owned-now row %s does not appear; a rung adding an owned-now row must add it to the greenfield template in the same pass", greenfieldTemplate, row.Name)
		}
	}
}

// templateKeys parses the template at path for every DUBPLATE_* name
// assigned on its own line, in file order.
func templateKeys(t *testing.T, path string) []string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("close template: %v", err)
		}
	}()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "DUBPLATE_") {
			continue
		}
		name, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		names = append(names, name)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan template: %v", err)
	}
	return names
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
