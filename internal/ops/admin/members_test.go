package admin

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/store"
)

func TestListMembers_LiveRosterByHexID(t *testing.T) {
	f := newFixture(t)
	resp, err := f.seam.ListMembers(writeCtx(t), actor(alice, stale), ops.ListMembersRequest{})
	if err != nil {
		t.Fatal(err)
	}
	want := []ops.Member{
		{UserID: alice.wire(), Address: "alice@example.com", DisplayName: "Alice", Username: "alice", Admin: true},
		{UserID: bob.wire(), Address: "bob@example.com", DisplayName: "Bob", Username: "bob", Admin: true},
		{UserID: carol.wire(), Address: "carol@example.com", DisplayName: "Carol", Username: "carol"},
	}
	if !reflect.DeepEqual(resp.Members, want) {
		t.Errorf("members = %+v, want %+v", resp.Members, want)
	}
}

// TestAddMember_RefusalsCarryTheirDetail proves each of the roster's
// seven add refusals answers its own detail, under conflict for a
// duplicate and invalid_request otherwise, and writes nothing.
func TestAddMember_RefusalsCarryTheirDetail(t *testing.T) {
	tests := []struct {
		name   string
		req    ops.AddMemberRequest
		code   ops.Code
		detail ops.Detail
	}{
		{"a live member's address", ops.AddMemberRequest{Address: "Carol@Example.com", DisplayName: "C", Username: "carol2"}, ops.CodeConflict, ops.DetailDuplicateAddress},
		{"a taken username", ops.AddMemberRequest{Address: "erin@example.com", DisplayName: "Erin", Username: "CAROL"}, ops.CodeConflict, ops.DetailDuplicateUsername},
		{"an address that does not parse", ops.AddMemberRequest{Address: "not an address", DisplayName: "Erin", Username: "erin"}, ops.CodeInvalidRequest, ops.DetailInvalidAddress},
		{"a display name too long", ops.AddMemberRequest{Address: "erin@example.com", DisplayName: strings.Repeat("e", 201), Username: "erin"}, ops.CodeInvalidRequest, ops.DetailInvalidDisplayName},
		{"a blank display name", ops.AddMemberRequest{Address: "erin@example.com", DisplayName: "   ", Username: "erin"}, ops.CodeInvalidRequest, ops.DetailBlankDisplayName},
		{"a username of the wrong shape", ops.AddMemberRequest{Address: "erin@example.com", DisplayName: "Erin", Username: "-erin"}, ops.CodeInvalidRequest, ops.DetailInvalidUsername},
		{"no username", ops.AddMemberRequest{Address: "erin@example.com", DisplayName: "Erin"}, ops.CodeInvalidRequest, ops.DetailMissingUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			before := snapshot(t, f.h)
			_, err := f.seam.AddMember(writeCtx(t), actor(alice, fresh), tt.req)
			wantCode(t, err, tt.code, tt.detail)
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a refused add wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}

// TestAddMember_NeverGrantsAdmin proves the add request has no admin
// field to carry a grant, the added member answers without the flag,
// and a removed admin's address returns as a member (AC17's seam half).
func TestAddMember_NeverGrantsAdmin(t *testing.T) {
	typ := reflect.TypeFor[ops.AddMemberRequest]()
	for field := range typ.Fields() {
		if strings.EqualFold(field.Name, "admin") || strings.HasPrefix(field.Tag.Get("json"), "admin") {
			t.Errorf("AddMemberRequest carries %s, an admin field", field.Name)
		}
	}

	f := newFixture(t)
	resp, err := f.seam.AddMember(writeCtx(t), actor(alice, fresh), ops.AddMemberRequest{Address: "Erin@Example.com", DisplayName: "Erin", Username: "erin"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Member.Admin || resp.Member.Address != "erin@example.com" || resp.Member.Username != "erin" || resp.Member.DisplayName != "Erin" {
		t.Errorf("added = %+v, want erin@example.com as a member", resp.Member)
	}
	if resp.Member.UserID.IsZero() {
		t.Errorf("added user_id = %s, want a minted id, never the zero id", resp.Member.UserID)
	}

	readd, err := f.seam.AddMember(writeCtx(t), actor(alice, fresh), ops.AddMemberRequest{Address: dave.address, DisplayName: "Dave", Username: "dave"})
	if err != nil {
		t.Fatal(err)
	}
	if readd.Member.UserID != dave.wire() || readd.Member.Admin {
		t.Errorf("re-added dave = %+v, want user_id %s without the flag", readd.Member, dave.wire())
	}
	if admin, tombstoned := userState(t, f.h, dave); admin || tombstoned {
		t.Errorf("dave admin=%v tombstoned=%v, want a live member", admin, tombstoned)
	}
}

// cancelOnRecord is a slog.Handler that calls cancel when it receives a
// record whose message is msg, and drops every record.
type cancelOnRecord struct {
	msg    string
	cancel context.CancelFunc
}

func (h cancelOnRecord) Enabled(context.Context, slog.Level) bool { return true }
func (h cancelOnRecord) Handle(_ context.Context, r slog.Record) error {
	if r.Message == h.msg {
		h.cancel()
	}
	return nil
}
func (h cancelOnRecord) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h cancelOnRecord) WithGroup(string) slog.Handler      { return h }

// TestAddMember_AnswersTheStoredMemberOnceCommitted proves an add whose
// write committed answers success with the stored member, though the
// caller's context ends the moment the roster logs the commit, before
// any second read could run.
func TestAddMember_AnswersTheStoredMemberOnceCommitted(t *testing.T) {
	tests := []struct {
		name string
		req  ops.AddMemberRequest
		want ops.Member
	}{
		{
			name: "a new address",
			req:  ops.AddMemberRequest{Address: "Erin@Example.com", DisplayName: "Erin", Username: "erin"},
			want: ops.Member{Address: "erin@example.com", DisplayName: "Erin", Username: "erin"},
		},
		{
			name: "a removed member's address",
			req:  ops.AddMemberRequest{Address: dave.address, DisplayName: "Dave", Username: "dave"},
			want: ops.Member{UserID: dave.wire(), Address: "dave@example.com", DisplayName: "Dave", Username: "dave"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			t.Cleanup(cancel)
			f.seam.cfg.Logger = slog.New(cancelOnRecord{msg: "roster member added", cancel: cancel})

			resp, err := f.seam.AddMember(ctx, actor(alice, fresh), tt.req)
			if err != nil {
				t.Fatalf("AddMember() error = %v, want nil: the write committed", err)
			}
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatalf("ctx.Err() = %v, want context.Canceled: the commit log did not fire, so the test proves nothing", ctx.Err())
			}

			list, err := f.seam.ListMembers(writeCtx(t), actor(alice, fresh), ops.ListMembersRequest{})
			if err != nil {
				t.Fatal(err)
			}
			var stored ops.Member
			for _, m := range list.Members {
				if m.Address == tt.want.Address {
					stored = m
				}
			}
			if stored.UserID.IsZero() {
				t.Fatalf("the roster holds no member at %s after the add", tt.want.Address)
			}
			if resp.Member != stored {
				t.Errorf("answered member = %+v, want the stored row %+v", resp.Member, stored)
			}
			if !tt.want.UserID.IsZero() && resp.Member.UserID != tt.want.UserID {
				t.Errorf("answered user_id = %s, want the existing id %s", resp.Member.UserID, tt.want.UserID)
			}
			if got := resp.Member; got.Address != tt.want.Address || got.DisplayName != tt.want.DisplayName || got.Username != tt.want.Username || got.Admin {
				t.Errorf("answered member = %+v, want %s as a member named %s", got, tt.want.Address, tt.want.DisplayName)
			}
			if resp.Member.UserID.IsZero() {
				t.Errorf("answered user_id = %s, want a stored id, never the candidate's zero id", resp.Member.UserID)
			}
		})
	}
}

// TestRemoveAndDemote_Outcomes covers the roster refusals the seam maps
// for a removal and a demotion.
func TestRemoveAndDemote_Outcomes(t *testing.T) {
	remove := func(ctx context.Context, f *fixture, a identity.Actor, target id.ID) error {
		_, err := f.seam.RemoveMember(ctx, a, ops.RemoveMemberRequest{UserID: target})
		return err
	}
	demote := func(ctx context.Context, f *fixture, a identity.Actor, target id.ID) error {
		_, err := f.seam.DemoteMember(ctx, a, ops.DemoteMemberRequest{UserID: target})
		return err
	}
	type write = func(ctx context.Context, f *fixture, a identity.Actor, target id.ID) error
	tests := []struct {
		name   string
		prep   func(t *testing.T, f *fixture)
		write  write
		target id.ID
		code   ops.Code
		detail ops.Detail
	}{
		{name: "remove an unknown id", write: remove, target: id.ID{0xee}, code: ops.CodeNotFound},
		{name: "demote an unknown id", write: demote, target: id.ID{0xee}, code: ops.CodeNotFound},
		{name: "remove a removed member", write: remove, target: dave.wire(), code: ops.CodeNotFound},
		{name: "demote a removed member", write: demote, target: dave.wire(), code: ops.CodeNotFound},
		{
			name: "remove the last admin", write: remove, target: alice.wire(), code: ops.CodeConflict, detail: ops.DetailLastAdmin,
			prep: func(t *testing.T, f *fixture) { execSQL(t, f.h, `UPDATE users SET admin = 0 WHERE id = ?`, bob.id[:]) },
		},
		{
			name: "demote the last admin", write: demote, target: alice.wire(), code: ops.CodeConflict, detail: ops.DetailLastAdmin,
			prep: func(t *testing.T, f *fixture) { execSQL(t, f.h, `UPDATE users SET admin = 0 WHERE id = ?`, bob.id[:]) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			if tt.prep != nil {
				tt.prep(t, f)
			}
			before := snapshot(t, f.h)
			wantCode(t, tt.write(writeCtx(t), f, actor(alice, fresh), tt.target), tt.code, tt.detail)
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a refused write wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}

// TestDemoteMember_NonAdminWritesNothing proves demoting a member who
// holds no flag succeeds and writes nothing.
func TestDemoteMember_NonAdminWritesNothing(t *testing.T) {
	f := newFixture(t)
	before := snapshot(t, f.h)
	if _, err := f.seam.DemoteMember(writeCtx(t), actor(alice, fresh), ops.DemoteMemberRequest{UserID: carol.wire()}); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(t, f.h); after != before {
		t.Errorf("demoting a member wrote:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestRemoveMember_BumpsTheEpoch proves the seam's removal reaches the
// roster's revoking path: the address's epoch reads the removal time.
func TestRemoveMember_BumpsTheEpoch(t *testing.T) {
	f := newFixture(t)
	if _, err := f.seam.RemoveMember(writeCtx(t), actor(alice, fresh), ops.RemoveMemberRequest{UserID: carol.wire()}); err != nil {
		t.Fatal(err)
	}
	epoch, ok, err := f.epoch.Read(carol.address)
	if err != nil || !ok || !epoch.Equal(testNow) {
		t.Errorf("carol's epoch = %v, %v, %v; want %v", epoch, ok, err, testNow)
	}
}

// TestUserID_Zero proves P11's parse for the one id the seam refuses by
// value: the zero id answers invalid_request on a removal and a
// demotion and writes nothing. A malformed id never reaches the seam,
// since the JSON decoder refuses it upstream.
func TestUserID_Zero(t *testing.T) {
	f := newFixture(t)
	before := snapshot(t, f.h)
	_, err := f.seam.RemoveMember(writeCtx(t), actor(alice, fresh), ops.RemoveMemberRequest{UserID: id.ID{}})
	wantCode(t, err, ops.CodeInvalidRequest, "")
	_, err = f.seam.DemoteMember(writeCtx(t), actor(alice, fresh), ops.DemoteMemberRequest{UserID: id.ID{}})
	wantCode(t, err, ops.CodeInvalidRequest, "")
	if after := snapshot(t, f.h); after != before {
		t.Errorf("a zero id wrote:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestMemberWrites_ActorNoLongerAdmin proves an actor whose row a
// direct store write has demoted or removed is refused forbidden on
// every member write, which the roster's in-write re-read decides.
func TestMemberWrites_ActorNoLongerAdmin(t *testing.T) {
	changes := map[string]string{
		"demoted": `UPDATE users SET admin = 0 WHERE id = ?`,
		"removed": `UPDATE users SET tombstoned_at = 1 WHERE id = ?`,
	}
	for name, stmt := range changes {
		for _, c := range gatedCalls {
			t.Run(name+", "+c.name, func(t *testing.T) {
				f := newFixture(t)
				if c.prep != nil {
					c.prep(t, f)
				}
				execSQL(t, f.h, stmt, alice.id[:])
				before := snapshot(t, f.h)
				wantCode(t, c.call(writeCtx(t), f, actor(alice, fresh)), ops.CodeForbidden, "")
				if after := snapshot(t, f.h); after != before {
					t.Errorf("a refused write wrote:\nbefore:\n%s\nafter:\n%s", before, after)
				}
			})
		}
	}
}

// busyArms are the two ways a write finds the store busy: the seam's
// own handle's write connection held past the call's deadline, and
// another handle's write lock held past the busy timeout.
var busyArms = []struct {
	name string
	run  func(t *testing.T, call func(ctx context.Context, f *fixture) error)
}{
	{
		name: "deadline on the held write connection",
		run: func(t *testing.T, call func(ctx context.Context, f *fixture) error) {
			f := newFixture(t)
			before := snapshot(t, f.h)
			release := holdUpdate(t, f.h)
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			err := call(ctx, f)
			release()
			wantCode(t, err, ops.CodeUnavailable, ops.DetailBusy)
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a busy write wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		},
	},
	{
		name: "busy timeout on another handle's lock",
		run: func(t *testing.T, call func(ctx context.Context, f *fixture) error) {
			f := newFixtureWith(t, fixtureOptions{busy: 100 * time.Millisecond})
			before := snapshot(t, f.h)
			release := holdUpdate(t, f.secondHandle(t))
			err := call(writeCtx(t), f)
			release()
			wantCode(t, err, ops.CodeUnavailable, ops.DetailBusy)
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a busy write wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		},
	},
}

// holdUpdate starts an Update on h whose callback blocks, and returns
// once it holds the write connection. The returned release ends it.
func holdUpdate(t *testing.T, h *store.Handle) (release func()) {
	t.Helper()
	locked := make(chan struct{})
	gate := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- h.Update(writeCtx(t), func(context.Context, *store.Tx) error {
			close(locked)
			<-gate
			return nil
		})
	}()
	select {
	case <-locked:
	case err := <-done:
		t.Fatalf("the held write ended before holding: %v", err)
	}
	return func() {
		close(gate)
		if err := <-done; err != nil {
			t.Errorf("the held write: %v", err)
		}
	}
}

// TestMemberWrite_Busy proves a member write that finds the store busy
// answers unavailable with detail busy and writes nothing.
func TestMemberWrite_Busy(t *testing.T) {
	for _, arm := range busyArms {
		t.Run(arm.name, func(t *testing.T) {
			arm.run(t, func(ctx context.Context, f *fixture) error {
				_, err := f.seam.DemoteMember(ctx, actor(alice, fresh), ops.DemoteMemberRequest{UserID: bob.wire()})
				return err
			})
		})
	}
}
