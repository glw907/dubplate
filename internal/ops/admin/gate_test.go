package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/ops"
)

// A seamCall is one write through the seam, with the fixture state it
// needs first and a check that it committed.
type seamCall struct {
	name      string
	prep      func(t *testing.T, f *fixture)
	call      func(ctx context.Context, f *fixture, a ops.Actor) error
	committed func(t *testing.T, f *fixture)
}

// gatedCalls are the five calls the re-authentication gate holds: a
// sensitive set and reset, and every member write (D10).
var gatedCalls = []seamCall{
	{
		name: "set_setting on a sensitive key",
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.SetSetting(ctx, a, ops.SetSettingRequest{Key: keyLifetime, Value: "3h"})
			return err
		},
		committed: func(t *testing.T, f *fixture) {
			if v, src := storedValue(t, f.h, keyLifetime); v != "3h" || src != "admin" {
				t.Errorf("stored %s = %q (%s), want 3h (admin)", keyLifetime, v, src)
			}
		},
	},
	{
		name: "reset_setting on a sensitive key",
		prep: func(t *testing.T, f *fixture) { plantAdminRow(t, f, keyLifetime, "3h") },
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.ResetSetting(ctx, a, ops.ResetSettingRequest{Key: keyLifetime})
			return err
		},
		committed: func(t *testing.T, f *fixture) {
			if v, _ := storedValue(t, f.h, keyLifetime); v != "" {
				t.Errorf("stored %s = %q, want no row", keyLifetime, v)
			}
		},
	},
	{
		name: "add_member",
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.AddMember(ctx, a, ops.AddMemberRequest{Address: "erin@example.com", DisplayName: "Erin", Username: "erin"})
			return err
		},
		committed: func(t *testing.T, f *fixture) {
			var n int
			if err := f.h.QueryRow(context.Background(), `SELECT count(*) FROM users WHERE address = 'erin@example.com'`).Scan(&n); err != nil || n != 1 {
				t.Errorf("erin rows = %d (%v), want 1", n, err)
			}
		},
	},
	{
		name: "remove_member",
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.RemoveMember(ctx, a, ops.RemoveMemberRequest{UserID: carol.hex})
			return err
		},
		committed: func(t *testing.T, f *fixture) {
			if _, tombstoned := userState(t, f.h, carol); !tombstoned {
				t.Error("carol is still live")
			}
		},
	},
	{
		name: "demote_member",
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.DemoteMember(ctx, a, ops.DemoteMemberRequest{UserID: bob.hex})
			return err
		},
		committed: func(t *testing.T, f *fixture) {
			if admin, _ := userState(t, f.h, bob); admin {
				t.Error("bob is still an admin")
			}
		},
	},
}

// ungatedCalls are a non-sensitive set and reset, which a stale session
// still makes.
var ungatedCalls = []seamCall{
	{
		name: "set_setting on a non-sensitive key",
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.SetSetting(ctx, a, ops.SetSettingRequest{Key: keyKeep, Value: "20"})
			return err
		},
		committed: func(t *testing.T, f *fixture) {
			if v, src := storedValue(t, f.h, keyKeep); v != "20" || src != "admin" {
				t.Errorf("stored %s = %q (%s), want 20 (admin)", keyKeep, v, src)
			}
		},
	},
	{
		name: "reset_setting on a non-sensitive key",
		prep: func(t *testing.T, f *fixture) { plantAdminRow(t, f, keyKeep, "20") },
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.ResetSetting(ctx, a, ops.ResetSettingRequest{Key: keyKeep})
			return err
		},
		committed: func(t *testing.T, f *fixture) {
			if v, _ := storedValue(t, f.h, keyKeep); v != "" {
				t.Errorf("stored %s = %q, want no row", keyKeep, v)
			}
		},
	},
}

// TestGate_AnchorInsideTheWindowCommits proves each gated call from a
// session anchored inside the window commits, and each from one
// anchored outside it answers reauthentication_required and writes
// nothing (AC12).
func TestGate_AnchorInsideTheWindowCommits(t *testing.T) {
	for _, c := range gatedCalls {
		t.Run(c.name+", fresh", func(t *testing.T) {
			f := newFixture(t)
			if c.prep != nil {
				c.prep(t, f)
			}
			if err := c.call(writeCtx(t), f, actor(alice, fresh)); err != nil {
				t.Fatalf("call = %v, want nil", err)
			}
			c.committed(t, f)
		})
		t.Run(c.name+", stale", func(t *testing.T) {
			f := newFixture(t)
			if c.prep != nil {
				c.prep(t, f)
			}
			before := snapshot(t, f.h)
			err := c.call(writeCtx(t), f, actor(alice, stale))
			wantCode(t, err, ops.CodeReauthenticationRequired, "")
			if !errors.Is(err, ops.ErrReauthenticationRequired) {
				t.Errorf("err = %v, want it to wrap ErrReauthenticationRequired", err)
			}
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a refused call wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
		t.Run(c.name+", anchored in the future", func(t *testing.T) {
			f := newFixture(t)
			if c.prep != nil {
				c.prep(t, f)
			}
			before := snapshot(t, f.h)
			err := c.call(writeCtx(t), f, actor(alice, testNow.Add(time.Hour)))
			wantCode(t, err, ops.CodeReauthenticationRequired, "")
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a refused call wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}

// TestGate_StaleSessionMakesANonSensitiveWrite proves the gate holds
// only what it names: a stale session's non-sensitive set and reset
// commit.
func TestGate_StaleSessionMakesANonSensitiveWrite(t *testing.T) {
	for _, c := range ungatedCalls {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			if c.prep != nil {
				c.prep(t, f)
			}
			if err := c.call(writeCtx(t), f, actor(alice, stale)); err != nil {
				t.Fatalf("call = %v, want nil", err)
			}
			c.committed(t, f)
		})
	}
}

// TestGate_UnreadableWindowRefuses proves a window that cannot be read
// refuses every gated call with internal, even from a fresh session, so
// no default stands in for it.
func TestGate_UnreadableWindowRefuses(t *testing.T) {
	broken := func(context.Context) (time.Duration, error) { return 0, errors.New("the window row does not parse") }
	for _, c := range gatedCalls {
		t.Run(c.name, func(t *testing.T) {
			f := newFixtureWith(t, fixtureOptions{window: broken})
			if c.prep != nil {
				c.prep(t, f)
			}
			before := snapshot(t, f.h)
			wantCode(t, c.call(writeCtx(t), f, actor(alice, fresh)), ops.CodeInternal, "")
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a refused call wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}
