package auth

import (
	"context"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/store"
)

func TestEarliestAdmin(t *testing.T) {
	first := identity.UserID{0x20}
	second := identity.UserID{0x10}
	tiedLow := identity.UserID{0x01}
	tiedHigh := identity.UserID{0x02}

	tests := []struct {
		name   string
		rows   []userRow
		want   identity.UserID
		wantOK bool
	}{
		{
			name: "the first-created of two admins, whatever the ids",
			rows: []userRow{
				{id: second, address: "b@example.com", displayName: "B", username: "b", admin: true, createdAt: 200},
				{id: first, address: "a@example.com", displayName: "A", username: "a", admin: true, createdAt: 100},
			},
			want:   first,
			wantOK: true,
		},
		{
			name: "a tombstoned earliest admin is skipped",
			rows: []userRow{
				{id: first, address: "a@example.com", displayName: "A", username: "a", admin: true, tombstoned: true, createdAt: 100},
				{id: second, address: "b@example.com", displayName: "B", username: "b", admin: true, createdAt: 200},
			},
			want:   second,
			wantOK: true,
		},
		{
			name: "a member who is not an admin is skipped",
			rows: []userRow{
				{id: first, address: "a@example.com", displayName: "A", username: "a", createdAt: 100},
				{id: second, address: "b@example.com", displayName: "B", username: "b", admin: true, createdAt: 200},
			},
			want:   second,
			wantOK: true,
		},
		{
			name: "two admins created at one instant answer the lower id",
			rows: []userRow{
				{id: tiedHigh, address: "h@example.com", displayName: "H", username: "h", admin: true, createdAt: 100},
				{id: tiedLow, address: "l@example.com", displayName: "L", username: "l", admin: true, createdAt: 100},
			},
			want:   tiedLow,
			wantOK: true,
		},
		{
			name: "no live admin answers false",
			rows: []userRow{
				{id: first, address: "a@example.com", displayName: "A", username: "a", admin: true, tombstoned: true, createdAt: 100},
				{id: second, address: "b@example.com", displayName: "B", username: "b", createdAt: 200},
			},
		},
		{name: "an empty roster answers false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUsersStore(t)
			insertUsers(t, h, tt.rows...)
			var got identity.UserID
			var ok bool
			err := h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
				var err error
				got, ok, err = EarliestAdmin(ctx, tx)
				return err
			})
			if err != nil {
				t.Fatalf("EarliestAdmin: %v", err)
			}
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("EarliestAdmin = (%x, %v), want (%x, %v)", got[:], ok, tt.want[:], tt.wantOK)
			}
		})
	}
}

func TestLiveMembers(t *testing.T) {
	tests := []struct {
		name string
		rows []userRow
		want int
	}{
		{name: "an empty roster counts none"},
		{
			name: "a tombstoned member is not counted",
			rows: []userRow{
				{id: identity.UserID{0x01}, address: "a@example.com", displayName: "A", username: "a", createdAt: 100},
				{id: identity.UserID{0x02}, address: "b@example.com", displayName: "B", username: "b", tombstoned: true, createdAt: 100},
				{id: identity.UserID{0x03}, address: "c@example.com", displayName: "C", username: "c", admin: true, createdAt: 100},
			},
			want: 2,
		},
		{
			name: "only tombstoned members count none",
			rows: []userRow{
				{id: identity.UserID{0x01}, address: "a@example.com", displayName: "A", username: "a", tombstoned: true, createdAt: 100},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newUsersStore(t)
			insertUsers(t, h, tt.rows...)
			got, err := LiveMembers(t.Context(), h)
			if err != nil {
				t.Fatalf("LiveMembers: %v", err)
			}
			if got != tt.want {
				t.Errorf("LiveMembers = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestReadErrorsAreBare holds the contract the two callers rely on: a read
// that fails returns the driver's error unprefixed, so reconcile and
// RouteCollection print the one prefix their operators already read.
func TestReadErrorsAreBare(t *testing.T) {
	ctx := writeCtx(t)

	closed := newUsersStore(t)
	if err := closed.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := LiveMembers(ctx, closed); err == nil || strings.HasPrefix(err.Error(), "auth:") {
		t.Errorf("LiveMembers on a closed store = %v, want a bare driver error", err)
	}

	// The drop fails the read inside a live transaction; the failed Update
	// rolls it back.
	h := newUsersStore(t)
	err := h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if _, err := tx.Exec(ctx, `DROP TABLE users`); err != nil {
			t.Fatalf("drop users: %v", err)
		}
		_, _, err := EarliestAdmin(ctx, tx)
		if err == nil || strings.HasPrefix(err.Error(), "auth:") {
			t.Errorf("EarliestAdmin with no users table = %v, want a bare driver error", err)
		}
		return err
	})
	if err == nil {
		t.Error("Update returned nil, want the read's error")
	}
}
