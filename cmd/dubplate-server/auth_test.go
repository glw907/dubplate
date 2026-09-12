package main

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// TestRegisterAuthConfig_SelectorIsOwnedNow proves the selector lands as
// an owned wall row with no Pending, retiring keys.go's own pending
// "identity source" status row (R5): a registry that registers nothing
// else still requires DUBPLATE_IDENTITY_SOURCE once registerAuthConfig
// has run.
func TestRegisterAuthConfig_SelectorIsOwnedNow(t *testing.T) {
	reg := config.NewRegistry()
	registerAuthConfig(reg)

	rows := reg.OwnedWall()
	i := slices.IndexFunc(rows, func(row config.WallRow) bool {
		return row.Name == "DUBPLATE_IDENTITY_SOURCE"
	})
	if i < 0 {
		t.Fatal("OwnedWall() has no DUBPLATE_IDENTITY_SOURCE row")
	}
	if rows[i].Pending != "" {
		t.Errorf("DUBPLATE_IDENTITY_SOURCE.Pending = %q, want empty", rows[i].Pending)
	}
}

// TestRegisterAuthConfig_MagicLinkSource proves the magic-link source
// declares exactly the nine required and two optional names R-5b3
// rules, on the selector registerAuthConfig also registers as a wall
// row.
func TestRegisterAuthConfig_MagicLinkSource(t *testing.T) {
	reg := config.NewRegistry()
	registerAuthConfig(reg)

	sources := reg.IdentitySources()
	if len(sources) != 1 {
		t.Fatalf("IdentitySources() = %d entries, want 1", len(sources))
	}
	src := sources[0]

	if src.Name != "magic-link" {
		t.Errorf("Name = %q, want %q", src.Name, "magic-link")
	}
	if src.Selector != "DUBPLATE_IDENTITY_SOURCE" {
		t.Errorf("Selector = %q, want %q", src.Selector, "DUBPLATE_IDENTITY_SOURCE")
	}

	wantRequired := []string{
		"DUBPLATE_ROSTER_FILE",
		"DUBPLATE_MAGIC_LINK_SECRET",
		"DUBPLATE_SESSION_SECRET",
		"DUBPLATE_PUBLIC_URL",
		"DUBPLATE_SMTP_HOST",
		"DUBPLATE_SMTP_PORT",
		"DUBPLATE_SMTP_USER",
		"DUBPLATE_SMTP_PASSWORD",
		"DUBPLATE_SMTP_FROM",
	}
	if !slices.Equal(src.Required, wantRequired) {
		t.Errorf("Required = %v, want %v", src.Required, wantRequired)
	}

	wantOptional := []string{
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS",
		"DUBPLATE_SESSION_SECRET_PREVIOUS",
	}
	if !slices.Equal(src.Optional, wantOptional) {
		t.Errorf("Optional = %v, want %v", src.Optional, wantOptional)
	}
}

// TestRegisterAuthConfig_RetiresAuthAndIdentityReservedClass proves
// the removal half of the rung 1b registry contract landed in this
// commit, on TestRegisterServeConfig_ClaimsThreeKeysAndRetiresReservedClass's
// own precedent (serve_test.go): the class registerAuthConfig's
// declarations replace is gone from the production registry.
func TestRegisterAuthConfig_RetiresAuthAndIdentityReservedClass(t *testing.T) {
	reg := productionRegistry()

	if slices.ContainsFunc(reg.ReservedClasses(), func(class config.ReservedClass) bool {
		return class.Name == "auth and identity"
	}) {
		t.Error(`ReservedClasses() still carries the "auth and identity" class, want it retired`)
	}
}

// TestDoctorCmd_UnsetSelectorExitsNonzero proves doctor's exit gate on
// the production registry, rather than only on a scratch one: with
// DUBPLATE_IDENTITY_SOURCE unset, Execute returns a non-nil error and
// stdout carries the selector's own "missing" row.
func TestDoctorCmd_UnsetSelectorExitsNonzero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(baseEnviron(), productionRegistry(), &stdout, &stderr)
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	wantPrefix := fmt.Sprintf("%-32s %s", "DUBPLATE_IDENTITY_SOURCE", "missing")
	found := false
	for line := range strings.SplitSeq(stdout.String(), "\n") {
		if strings.HasPrefix(line, wantPrefix) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("stdout = %q, want a line with prefix %q", stdout.String(), wantPrefix)
	}
}

// TestRegisterAuthConfig_MagicLinkKeyWithoutSelectorNamesSelector
// proves the declared-name discipline against the production
// registry: setting a magic-link source key while the selector is
// unset is a load error naming DUBPLATE_IDENTITY_SOURCE, never an
// "unknown configuration key" error, because rung 5's real names must
// be declared as a source's Required/Optional rows rather than as
// plain wall or threshold entries.
func TestRegisterAuthConfig_MagicLinkKeyWithoutSelectorNamesSelector(t *testing.T) {
	environ := append(baseEnviron(),
		"DUBPLATE_MAGIC_LINK_SECRET=fixture magic link secret, not a real key")

	_, err := config.Load(environ, productionRegistry())
	if err == nil {
		t.Fatal("Load() error = nil, want error naming DUBPLATE_IDENTITY_SOURCE")
	}
	if !strings.Contains(err.Error(), "DUBPLATE_IDENTITY_SOURCE") {
		t.Errorf("Load() error = %q, want it to name DUBPLATE_IDENTITY_SOURCE", err.Error())
	}
	if strings.Contains(err.Error(), "unknown configuration key") {
		t.Errorf("Load() error = %q, want no unknown-key error", err.Error())
	}
}
