package auth_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
)

// TestRosterSentinels_MatchFromOutside proves each refusal the roster
// exports for internal/ops/admin to branch on is matched by errors.Is
// from outside the package. Fixture state: a store holding one live
// admin and one live member, so each refusal is the one its row names.
func TestRosterSentinels_MatchFromOutside(t *testing.T) {
	h, err := store.OpenServe(t.Context(), t.TempDir(), store.ServeOptions{
		BusyTimeout:   5 * time.Second,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	leaf, err := auth.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	roster := auth.NewRoster(h, auth.OpenEpochStore(leaf), time.Now)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	logger := slog.New(slog.DiscardHandler)
	operator := auth.OperatorActor("operator")

	for _, m := range []auth.Member{
		{Address: "admin@example.com", DisplayName: "Admin", Username: "admin", Admin: true},
		{Address: "member@example.com", DisplayName: "Member", Username: "member"},
	} {
		if err := roster.AddMember(ctx, m, operator, logger); err != nil {
			t.Fatalf("seed %s: %v", m.Username, err)
		}
	}
	admin, err := roster.Lookup(ctx, "admin@example.com")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	member, err := roster.Lookup(ctx, "member@example.com")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}

	add := func(address, displayName, username string) func() error {
		return func() error {
			return roster.AddMember(ctx, auth.Member{Address: address, DisplayName: displayName, Username: username}, operator, logger)
		}
	}
	tests := []struct {
		name  string
		write func() error
		want  error
	}{
		{"an unknown id", func() error { return roster.RemoveMember(ctx, identity.UserID{0x0b, 0x5e}, operator, logger) }, auth.ErrMemberNotFound},
		{"the last admin", func() error { return roster.DemoteMember(ctx, admin.ID, operator, logger) }, auth.ErrNoAdmin},
		{"a non-admin actor", func() error {
			return roster.RemoveMember(ctx, admin.ID, auth.MemberActor(member.ID, "192.0.2.1"), logger)
		}, auth.ErrActorNotAdmin},
		{"a duplicate address", add("Member@example.com", "Again", "again"), auth.ErrDuplicateAddress},
		{"a duplicate username", add("new@example.com", "New", "MEMBER"), auth.ErrDuplicateUsername},
		{"an invalid address", add("not-an-address", "New", "new"), auth.ErrInvalidMemberAddress},
		{"an invalid display name", add("new@example.com", "Ne\u202ew", "new"), auth.ErrInvalidDisplayName},
		{"a blank display name", add("new@example.com", "  ", "new"), auth.ErrBlankDisplayName},
		{"an invalid username", add("new@example.com", "New", ".x"), auth.ErrInvalidUsername},
		{"a missing username", add("new@example.com", "New", ""), auth.ErrMissingUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.write(); !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want it to match %v", err, tt.want)
			}
		})
	}
}
