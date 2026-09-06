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
			cfg, err := Load(tt.environ)

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

// TestLoad_TemplateKeysClassified proves the template disposition walk
// is complete: every DUBPLATE_* name env/dubplate.env.template lists
// must classify as owned, reserved, or superseded, never unknown, or a
// real adopter's template-derived env file would trip a load error
// this package never intended.
func TestLoad_TemplateKeysClassified(t *testing.T) {
	names := templateKeys(t)
	if len(names) == 0 {
		t.Fatal("templateKeys found no DUBPLATE_* names; template path or format changed")
	}

	for _, name := range names {
		if ownedNow(name) {
			continue
		}
		if _, ok := reservedClassOf(name); ok {
			continue
		}
		if _, ok := supersessionOf(name); ok {
			continue
		}
		t.Errorf("%s: classifies as unknown; add it to ownedWall, a reserved class, or supersededKeys", name)
	}
}

// templateKeys parses env/dubplate.env.template for every DUBPLATE_*
// name assigned on its own line, in file order.
func templateKeys(t *testing.T) []string {
	t.Helper()

	f, err := os.Open("../../env/dubplate.env.template")
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

// TestLoadShadowIntegration proves ruling 1's required-when-enabled
// semantics: a shadow key set with the flag unset is a load error
// naming the flag, every required path is reported at once when the
// flag is set, a threshold row is accepted and never required, and an
// unparseable flag is an error rather than a silent off.
func TestLoadShadowIntegration(t *testing.T) {
	base := []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/library",
		"DUBPLATE_STATE_DIR=/var/lib/x",
		"DUBPLATE_SITE_NAME=Example",
		"DUBPLATE_ADMIN_EMAIL=admin@example.test",
	}
	shadowPaths := []string{
		"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox",
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR=/var/lib/incumbent",
		"DUBPLATE_SHADOW_REVIEW_DIR=/srv/review",
	}

	tests := []struct {
		name       string
		extra      []string
		wantErr    []string // substrings every one of which must appear
		wantShadow bool
	}{
		{
			name:  "flag unset, no shadow keys: loads with a nil Shadow",
			extra: nil,
		},
		{
			name:    "a shadow key set with the flag unset names the flag",
			extra:   []string{"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ENABLED"},
		},
		{
			name:    "flag set with no paths reports every missing row at once",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=true"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ORACLE_STATE_DIR", "DUBPLATE_SHADOW_REVIEW_DIR"},
		},
		{
			name:       "flag set with every path resolves Shadow",
			extra:      append([]string{"DUBPLATE_SHADOW_ENABLED=true"}, shadowPaths...),
			wantShadow: true,
		},
		{
			name:       "a threshold row is accepted and never required",
			extra:      append([]string{"DUBPLATE_SHADOW_ENABLED=true", "DUBPLATE_SHADOW_FLOOR=48h"}, shadowPaths...),
			wantShadow: true,
		},
		{
			name:    "an unparseable flag is an error rather than off",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=maybe"},
			wantErr: []string{"DUBPLATE_SHADOW_ENABLED"},
		},
		// The two rows fold M-d adds. An adopter who turns the
		// integration off explicitly must be able to start the
		// server, and a shadow key set beside an off flag must
		// still name the flag.
		{
			name:  "an explicit false loads clean with a nil Shadow",
			extra: []string{"DUBPLATE_SHADOW_ENABLED=false"},
		},
		{
			name:    "an explicit false with a shadow key still names the flag",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=false", "DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ENABLED"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(append(slices.Clone(base), tt.extra...))
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if (cfg.Shadow != nil) != tt.wantShadow {
					t.Fatalf("Shadow non-nil = %v, want %v", cfg.Shadow != nil, tt.wantShadow)
				}
				if tt.wantShadow && cfg.Shadow.ObserveDir != "/srv/inbox" {
					t.Errorf("ObserveDir = %q, want %q", cfg.Shadow.ObserveDir, "/srv/inbox")
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
			if tt.name == "an unparseable flag is an error rather than off" {
				if n := strings.Count(err.Error(), "DUBPLATE_SHADOW_ENABLED"); n != 1 {
					t.Errorf("error names DUBPLATE_SHADOW_ENABLED %d times, want exactly 1: %v", n, err)
				}
			}
		})
	}
}
