package admin

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

func TestListMembers_LiveRosterByHexID(t *testing.T) {
	f := newFixture(t)
	resp, err := f.seam.ListMembers(writeCtx(t), actor(alice, stale), ops.ListMembersRequest{})
	if err != nil {
		t.Fatal(err)
	}
	want := []ops.Member{
		{UserID: alice.hex, Address: "alice@example.com", DisplayName: "Alice", Username: "alice", Admin: true},
		{UserID: bob.hex, Address: "bob@example.com", DisplayName: "Bob", Username: "bob", Admin: true},
		{UserID: carol.hex, Address: "carol@example.com", DisplayName: "Carol", Username: "carol"},
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
		detail string
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
	if len(resp.Member.UserID) != 32 || strings.ToLower(resp.Member.UserID) != resp.Member.UserID || resp.Member.UserID == strings.Repeat("0", 32) {
		t.Errorf("added user_id = %q, want 32 lowercase hex characters", resp.Member.UserID)
	}

	readd, err := f.seam.AddMember(writeCtx(t), actor(alice, fresh), ops.AddMemberRequest{Address: dave.address, DisplayName: "Dave", Username: "dave"})
	if err != nil {
		t.Fatal(err)
	}
	if readd.Member.UserID != dave.hex || readd.Member.Admin {
		t.Errorf("re-added dave = %+v, want user_id %s without the flag", readd.Member, dave.hex)
	}
	if admin, tombstoned := userState(t, f.h, dave); admin || tombstoned {
		t.Errorf("dave admin=%v tombstoned=%v, want a live member", admin, tombstoned)
	}
}

// TestRemoveAndDemote_Outcomes covers the roster refusals the seam maps
// for a removal and a demotion.
func TestRemoveAndDemote_Outcomes(t *testing.T) {
	remove := func(ctx context.Context, f *fixture, a ops.Actor, id string) error {
		_, err := f.seam.RemoveMember(ctx, a, ops.RemoveMemberRequest{UserID: id})
		return err
	}
	demote := func(ctx context.Context, f *fixture, a ops.Actor, id string) error {
		_, err := f.seam.DemoteMember(ctx, a, ops.DemoteMemberRequest{UserID: id})
		return err
	}
	type write = func(ctx context.Context, f *fixture, a ops.Actor, id string) error
	tests := []struct {
		name   string
		prep   func(t *testing.T, f *fixture)
		write  write
		target string
		code   ops.Code
		detail string
	}{
		{name: "remove an unknown id", write: remove, target: "ee000000000000000000000000000000", code: ops.CodeNotFound},
		{name: "demote an unknown id", write: demote, target: "ee000000000000000000000000000000", code: ops.CodeNotFound},
		{name: "remove a removed member", write: remove, target: dave.hex, code: ops.CodeNotFound},
		{name: "demote a removed member", write: demote, target: dave.hex, code: ops.CodeNotFound},
		{
			name: "remove the last admin", write: remove, target: alice.hex, code: ops.CodeConflict, detail: ops.DetailLastAdmin,
			prep: func(t *testing.T, f *fixture) { execSQL(t, f.h, `UPDATE users SET admin = 0 WHERE id = ?`, bob.id[:]) },
		},
		{
			name: "demote the last admin", write: demote, target: alice.hex, code: ops.CodeConflict, detail: ops.DetailLastAdmin,
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
	if _, err := f.seam.DemoteMember(writeCtx(t), actor(alice, fresh), ops.DemoteMemberRequest{UserID: carol.hex}); err != nil {
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
	if _, err := f.seam.RemoveMember(writeCtx(t), actor(alice, fresh), ops.RemoveMemberRequest{UserID: carol.hex}); err != nil {
		t.Fatal(err)
	}
	epoch, ok, err := f.epoch.Read(carol.address)
	if err != nil || !ok || !epoch.Equal(testNow) {
		t.Errorf("carol's epoch = %v, %v, %v; want %v", epoch, ok, err, testNow)
	}
}

// TestUserID_Malformed proves P11's parse: each malformed id answers
// invalid_request on a removal and a demotion and writes nothing.
func TestUserID_Malformed(t *testing.T) {
	ids := map[string]string{
		"empty":     "",
		"uppercase": strings.ToUpper(carol.hex),
		"30 chars":  carol.hex[:30],
		"34 chars":  carol.hex + "00",
		"not hex":   strings.Repeat("zz", 16),
		"all zeros": strings.Repeat("0", 32),
	}
	for name, id := range ids {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			before := snapshot(t, f.h)
			_, err := f.seam.RemoveMember(writeCtx(t), actor(alice, fresh), ops.RemoveMemberRequest{UserID: id})
			wantCode(t, err, ops.CodeInvalidRequest, "")
			_, err = f.seam.DemoteMember(writeCtx(t), actor(alice, fresh), ops.DemoteMemberRequest{UserID: id})
			wantCode(t, err, ops.CodeInvalidRequest, "")
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a malformed id wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
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
				_, err := f.seam.DemoteMember(ctx, actor(alice, fresh), ops.DemoteMemberRequest{UserID: bob.hex})
				return err
			})
		})
	}
}
