package main

import (
	"slices"
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

	var found *config.WallRow
	for _, row := range reg.OwnedWall() {
		if row.Name == "DUBPLATE_IDENTITY_SOURCE" {
			row := row
			found = &row
			break
		}
	}
	if found == nil {
		t.Fatal("OwnedWall() has no DUBPLATE_IDENTITY_SOURCE row")
	}
	if found.Pending != "" {
		t.Errorf("DUBPLATE_IDENTITY_SOURCE.Pending = %q, want empty", found.Pending)
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

	for _, class := range reg.ReservedClasses() {
		if class.Name == "auth and identity" {
			t.Error(`ReservedClasses() still carries the "auth and identity" class, want it retired`)
		}
	}
}
