package admin

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/ops"
)

// liveAdmins is who a notice goes to in the fixture: alice and bob.
// carol is a member and dave a removed admin, and neither hears.
var liveAdmins = []string{"alice@example.com", "bob@example.com"}

// TestNotice_SensitiveChangeMailsEveryLiveAdmin proves a committed
// sensitive set and reset each mail every live admin once, and the
// message names the key, the actor's principal, and the old and new
// values (AC13).
func TestNotice_SensitiveChangeMailsEveryLiveAdmin(t *testing.T) {
	tests := []struct {
		name               string
		prep               func(t *testing.T, f *fixture)
		call               func(ctx context.Context, f *fixture) error
		oldValue, newValue string
	}{
		{
			name: "set_setting",
			call: func(ctx context.Context, f *fixture) error {
				_, err := f.seam.SetSetting(ctx, actor(alice, fresh), ops.SetSettingRequest{Key: keyLifetime, Value: "3h"})
				return err
			},
			oldValue: "1h", newValue: "3h",
		},
		{
			name: "reset_setting",
			prep: func(t *testing.T, f *fixture) { plantAdminRow(t, f, keyLifetime, "7h") },
			call: func(ctx context.Context, f *fixture) error {
				_, err := f.seam.ResetSetting(ctx, actor(alice, fresh), ops.ResetSettingRequest{Key: keyLifetime})
				return err
			},
			oldValue: "7h", newValue: "1h",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			if tt.prep != nil {
				tt.prep(t, f)
			}
			if err := tt.call(writeCtx(t), f); err != nil {
				t.Fatal(err)
			}
			if got := f.mail.recipients(); !slices.Equal(got, liveAdmins) {
				t.Fatalf("notices went to %v, want %v", got, liveAdmins)
			}
			for _, m := range f.mail.sent {
				text := m.subject + "\n" + m.body
				for _, want := range []string{keyLifetime, "alice", `"` + tt.oldValue + `"`, `"` + tt.newValue + `"`} {
					if !strings.Contains(text, want) {
						t.Errorf("notice to %s = %q, want it to name %s", m.to, text, want)
					}
				}
			}
		})
	}
}

// TestNotice_NonSensitiveOrUnwrittenChangeMailsNoOne proves a
// non-sensitive set and reset, and a sensitive reset that finds nothing
// to reset, send nothing.
func TestNotice_NonSensitiveOrUnwrittenChangeMailsNoOne(t *testing.T) {
	calls := append(slices.Clone(ungatedCalls), seamCall{
		name: "reset_setting on a sensitive key with no row",
		call: func(ctx context.Context, f *fixture, a ops.Actor) error {
			_, err := f.seam.ResetSetting(ctx, a, ops.ResetSettingRequest{Key: keyLifetime})
			return err
		},
	})
	for _, c := range calls {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			if c.prep != nil {
				c.prep(t, f)
			}
			if err := c.call(writeCtx(t), f, actor(alice, fresh)); err != nil {
				t.Fatal(err)
			}
			if got := f.mail.recipients(); len(got) != 0 {
				t.Errorf("notices went to %v, want none", got)
			}
		})
	}
}

// TestNotice_FailedSendLeavesTheWrite proves a notice that fails is
// logged and the sensitive write stays committed.
func TestNotice_FailedSendLeavesTheWrite(t *testing.T) {
	f := newFixture(t)
	f.mail.err = errors.New("smtp: connection refused")
	if _, err := f.seam.SetSetting(writeCtx(t), actor(alice, fresh), ops.SetSettingRequest{Key: keyLifetime, Value: "3h"}); err != nil {
		t.Fatalf("set = %v, want nil: a failed notice does not fail the write", err)
	}
	if v, src := storedValue(t, f.h, keyLifetime); v != "3h" || src != "admin" {
		t.Errorf("stored %s = %q (%s), want 3h (admin)", keyLifetime, v, src)
	}
	logs := f.logs.String()
	if !strings.Contains(logs, "notice") || !strings.Contains(logs, "connection refused") || !strings.Contains(logs, keyLifetime) {
		t.Errorf("log = %q, want the failed notice logged with the key and its cause", logs)
	}
}

// TestNotice_SendHoldsNoWriter proves the notice runs after the write
// commits: while a send blocks, another write on the same handle gets
// the write connection inside a short deadline, which it could not if
// the send held the write. It also proves the send's context outlives
// the operation's and carries its own deadline.
func TestNotice_SendHoldsNoWriter(t *testing.T) {
	f := newFixture(t)
	f.mail.gate = make(chan struct{})
	f.mail.entered = make(chan struct{}, len(liveAdmins))

	opCtx, cancelOp := context.WithTimeout(context.Background(), 30*time.Second)
	done := make(chan error, 1)
	go func() {
		_, err := f.seam.SetSetting(opCtx, actor(alice, fresh), ops.SetSettingRequest{Key: keyLifetime, Value: "3h"})
		done <- err
	}()
	select {
	case <-f.mail.entered:
	case err := <-done:
		cancelOp()
		t.Fatalf("set returned %v before any send", err)
	case <-time.After(10 * time.Second):
		cancelOp()
		t.Fatal("no send began")
	}

	writeErr := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := f.seam.SetSetting(ctx, actor(bob, fresh), ops.SetSettingRequest{Key: keyKeep, Value: "20"})
		return err
	}()
	cancelOp()
	close(f.mail.gate)
	if err := <-done; err != nil {
		t.Errorf("sensitive set = %v, want nil", err)
	}
	if writeErr != nil {
		t.Errorf("a write while the notice was sending = %v, want nil", writeErr)
	}
	if f.mail.releasedErr != nil || !f.mail.releasedDeadline {
		t.Errorf("send context err = %v, deadline = %v; want no error from the canceled operation and a deadline of its own",
			f.mail.releasedErr, f.mail.releasedDeadline)
	}
}
