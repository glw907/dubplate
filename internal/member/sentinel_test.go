package member_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/store"
)

// TestRosterSentinels_MatchFromOutside proves each refusal the roster
// exports for internal/ops/admin to branch on is matched by errors.Is
// from outside the package. Fixture state: a store holding one live
// admin and one live member, so each refusal is the one its row names.
// No case reaches a revocation, so the roster holds no Revoker.
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
	roster := member.NewRoster(h, nil, time.Now, slog.New(slog.DiscardHandler))
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	operator := identity.LocalActor("operator")

	for _, m := range []member.Member{
		{Address: "admin@example.com", DisplayName: "Admin", Username: "admin", Admin: true},
		{Address: "member@example.com", DisplayName: "Member", Username: "member"},
	} {
		if _, err := roster.AddMember(ctx, m, operator); err != nil {
			t.Fatalf("seed %s: %v", m.Username, err)
		}
	}
	admin, err := roster.Lookup(ctx, "admin@example.com")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	plain, err := roster.Lookup(ctx, "member@example.com")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}

	add := func(address, displayName, username string) func() error {
		return func() error {
			_, err := roster.AddMember(ctx, member.Member{Address: address, DisplayName: displayName, Username: username}, operator)
			return err
		}
	}
	tests := []struct {
		name  string
		write func() error
		want  error
	}{
		{"an unknown id", func() error { return roster.RemoveMember(ctx, identity.UserID{0x0b, 0x5e}, operator) }, member.ErrMemberNotFound},
		{"the last admin", func() error { return roster.DemoteMember(ctx, admin.ID, operator) }, member.ErrNoAdmin},
		{"a non-admin actor", func() error {
			return roster.RemoveMember(ctx, admin.ID, identity.NetworkActor(identity.Identity{
				Address: plain.Address, UserID: plain.ID, Username: plain.Username,
				Channel: identity.ChannelCookie,
			}, time.Now(), "192.0.2.1"))
		}, member.ErrActorNotAdmin},
		{"a duplicate address", add("Member@example.com", "Again", "again"), member.ErrDuplicateAddress},
		{"a duplicate username", add("new@example.com", "New", "MEMBER"), member.ErrDuplicateUsername},
		{"an invalid address", add("not-an-address", "New", "new"), member.ErrInvalidMemberAddress},
		{"an invalid display name", add("new@example.com", "Ne\u202ew", "new"), member.ErrInvalidDisplayName},
		{"a blank display name", add("new@example.com", "  ", "new"), member.ErrBlankDisplayName},
		{"an invalid username", add("new@example.com", "New", ".x"), member.ErrInvalidUsername},
		{"a missing username", add("new@example.com", "New", ""), member.ErrMissingUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.write(); !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want it to match %v", err, tt.want)
			}
		})
	}
}
