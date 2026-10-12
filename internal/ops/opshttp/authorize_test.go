package opshttp

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// A seams is the fakes an authorizing seam is built over.
type seams struct {
	jobs  *fakeSeam
	admin *fakeAdmin
	s     *authorizingSeam
}

// newSeams builds an authorizing seam over fakes that succeed at every
// operation, admitting through auth.
func newSeams(t *testing.T, auth authorizer) seams {
	t.Helper()
	jobs := &fakeSeam{answers: successes()}
	admin := &fakeAdmin{answers: adminSuccesses()}
	s, err := newAuthorizingSeam(jobs, admin, auth, identity.ChannelCookie)
	if err != nil {
		t.Fatalf("newAuthorizingSeam: %v", err)
	}
	return seams{jobs: jobs, admin: admin, s: s}
}

// seamCall is one operation of the authorizing seam, called with a
// planted actor and a request that reaches the fake.
type seamCall struct {
	op   string
	call func(ctx context.Context, s *authorizingSeam, planted ops.Actor) error
}

// adminCalls is one call for each of the seven admin operations.
func adminCalls() []seamCall {
	return []seamCall{
		{ops.OpListMembers, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.ListMembers(ctx, a, ops.ListMembersRequest{})
			return err
		}},
		{ops.OpAddMember, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.AddMember(ctx, a, ops.AddMemberRequest{Address: "bo@example.com", DisplayName: "Bo", Username: "bo"})
			return err
		}},
		{ops.OpRemoveMember, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.RemoveMember(ctx, a, ops.RemoveMemberRequest{UserID: testHandle})
			return err
		}},
		{ops.OpDemoteMember, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.DemoteMember(ctx, a, ops.DemoteMemberRequest{UserID: testHandle})
			return err
		}},
		{ops.OpListSettings, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.ListSettings(ctx, a, ops.ListSettingsRequest{})
			return err
		}},
		{ops.OpSetSetting, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.SetSetting(ctx, a, ops.SetSettingRequest{Key: "K", Value: "v"})
			return err
		}},
		{ops.OpResetSetting, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.ResetSetting(ctx, a, ops.ResetSettingRequest{Key: "K"})
			return err
		}},
	}
}

// jobCalls is one call for each of the four job-seam operations.
func jobCalls() []seamCall {
	return []seamCall{
		{ops.OpStartBackup, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.StartBackup(ctx, a, ops.StartBackupRequest{})
			return err
		}},
		{ops.OpJobStatus, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.JobStatus(ctx, a, ops.JobStatusRequest{JobID: testHandle})
			return err
		}},
		{ops.OpCancelJob, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.CancelJob(ctx, a, ops.CancelJobRequest{JobID: testHandle})
			return err
		}},
		{ops.OpCatalog, func(ctx context.Context, s *authorizingSeam, a ops.Actor) error {
			_, err := s.Catalog(ctx, a, ops.CatalogRequest{})
			return err
		}},
	}
}

// TestAuthorizingSeamDiscardsItsCallersActor plants a command-line
// actor on every call, all eleven operations; the seam beneath must
// see only the actor the authorizer admitted, with its session.
func TestAuthorizingSeamDiscardsItsCallersActor(t *testing.T) {
	planted := ops.LocalActor("root")
	ctx := context.Background()
	for _, tt := range slices.Concat(jobCalls(), adminCalls()) {
		t.Run(tt.op, func(t *testing.T) {
			f := newSeams(t, admitAnn())
			if err := tt.call(ctx, f.s, planted); err != nil {
				t.Fatalf("%s: %v", tt.op, err)
			}
			calls := slices.Concat(f.jobs.calls, f.admin.calls)
			if len(calls) != 1 {
				t.Fatalf("seam calls = %d, want 1", len(calls))
			}
			got := calls[0]
			if got.op != tt.op || got.actor.String() != annWire {
				t.Errorf("seam got %s as %q, want %s as %q", got.op, got.actor, tt.op, annWire)
			}
			if got.actor.UserID() != ann.UserID || !got.actor.Anchor().Equal(annAnchor) {
				t.Errorf("seam's actor carries user %x anchored %v, want %x anchored %v",
					got.actor.UserID(), got.actor.Anchor(), ann.UserID, annAnchor)
			}
		})
	}
}

// TestAuthorizingSeamMemberReachesOnlyTheCatalog is the authorizer
// half of the spec's member criterion: a member caller reaches the
// catalog, and each admin operation, and each job operation, answers
// forbidden before any seam sees the call. A refused call that reached
// a fake fails here.
func TestAuthorizingSeamMemberReachesOnlyTheCatalog(t *testing.T) {
	ctx := context.Background()
	for _, tt := range slices.Concat(jobCalls(), adminCalls()) {
		t.Run(tt.op, func(t *testing.T) {
			f := newSeams(t, grant(ann, ops.PrivilegeMember))
			err := tt.call(ctx, f.s, ops.Actor{})
			if tt.op == ops.OpCatalog {
				if err != nil {
					t.Fatalf("member's catalog = %v, want admitted", err)
				}
				return
			}
			if !errors.Is(err, ops.ErrForbidden) {
				t.Errorf("member's %s = %v, want ops.ErrForbidden", tt.op, err)
			}
			if n := len(f.jobs.calls) + len(f.admin.calls); n != 0 {
				t.Errorf("a refused %s reached a seam %d times", tt.op, n)
			}
		})
	}
}

// TestAuthorizingSeamDenialIsForbidden checks the in-process shape of a
// denial: the error matches ops.ErrForbidden and still carries the
// authorizer's own error, on a job operation and on an admin one.
func TestAuthorizingSeamDenialIsForbidden(t *testing.T) {
	ctx := context.Background()
	errSession := errors.New("session expired")
	for _, tt := range []seamCall{jobCalls()[2], adminCalls()[1]} {
		t.Run(tt.op, func(t *testing.T) {
			deny := &fakeAuthorizer{decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.Actor{}, 0, errSession
			}}
			f := newSeams(t, deny)
			err := tt.call(ctx, f.s, ops.LocalActor("root"))
			if !errors.Is(err, ops.ErrForbidden) || !errors.Is(err, errSession) {
				t.Errorf("%s = %v, want ops.ErrForbidden wrapping the authorizer's error", tt.op, err)
			}
			if code := ops.CodeOf(err); code != ops.CodeForbidden {
				t.Errorf("CodeOf = %q, want forbidden", code)
			}
			if n := len(f.jobs.calls) + len(f.admin.calls); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestAuthorizingSeamChecksTheGrant admits a call only when the
// privilege the authorizer granted reaches the one the operation
// requires. Each row's authorizer admits without error, so a denial
// can come only from the seam's own check.
func TestAuthorizingSeamChecksTheGrant(t *testing.T) {
	ctx := context.Background()
	startBackup, catalog, listMembers := jobCalls()[0], jobCalls()[3], adminCalls()[0]
	tests := []struct {
		name    string
		op      seamCall
		granted ops.Privilege
		admit   bool
	}{
		{"admin on an admin operation", startBackup, ops.PrivilegeAdmin, true},
		{"member on an admin operation", startBackup, ops.PrivilegeMember, false},
		{"zero grant on an admin operation", startBackup, 0, false},
		{"unknown grant on an admin operation", startBackup, ops.Privilege(99), false},
		{"admin on a member operation", catalog, ops.PrivilegeAdmin, true},
		{"member on a member operation", catalog, ops.PrivilegeMember, true},
		{"zero grant on a member operation", catalog, 0, false},
		{"unknown grant on a member operation", catalog, ops.Privilege(99), false},
		{"admin on an admin seam operation", listMembers, ops.PrivilegeAdmin, true},
		{"member on an admin seam operation", listMembers, ops.PrivilegeMember, false},
		{"zero grant on an admin seam operation", listMembers, 0, false},
		{"unknown grant on an admin seam operation", listMembers, ops.Privilege(99), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSeams(t, grant(ann, tt.granted))
			err := tt.op.call(ctx, f.s, ops.Actor{})
			n := len(f.jobs.calls) + len(f.admin.calls)
			if tt.admit {
				if err != nil || n != 1 {
					t.Errorf("err = %v, seam calls = %d; want admitted once", err, n)
				}
				return
			}
			if !errors.Is(err, ops.ErrForbidden) {
				t.Errorf("err = %v, want ops.ErrForbidden", err)
			}
			if n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestAuthorizingSeamCatalogJoinsBothSeams proves the catalog answer is
// the job seam's entries followed by one admin entry per admin
// operation, whoever asks.
func TestAuthorizingSeamCatalogJoinsBothSeams(t *testing.T) {
	jobEntries := []ops.CatalogEntry{
		{Name: "start_backup", Privilege: ops.PrivilegeAdmin, Available: false, Reason: "no runner"},
		{Name: "catalog", Privilege: ops.PrivilegeMember, Available: true},
	}
	want := slices.Concat(jobEntries, []ops.CatalogEntry{
		{Name: "list_members", Privilege: ops.PrivilegeAdmin, Available: true},
		{Name: "add_member", Privilege: ops.PrivilegeAdmin, Available: true},
		{Name: "remove_member", Privilege: ops.PrivilegeAdmin, Available: true},
		{Name: "demote_member", Privilege: ops.PrivilegeAdmin, Available: true},
		{Name: "list_settings", Privilege: ops.PrivilegeAdmin, Available: true},
		{Name: "set_setting", Privilege: ops.PrivilegeAdmin, Available: true},
		{Name: "reset_setting", Privilege: ops.PrivilegeAdmin, Available: true},
	})
	for _, granted := range []ops.Privilege{ops.PrivilegeMember, ops.PrivilegeAdmin} {
		t.Run(granted.String(), func(t *testing.T) {
			f := newSeams(t, grant(ann, granted))
			f.jobs.answers[ops.OpCatalog] = fakeAnswer{resp: ops.CatalogResponse{Operations: jobEntries}}
			resp, err := f.s.Catalog(context.Background(), ops.Actor{}, ops.CatalogRequest{})
			if err != nil {
				t.Fatalf("Catalog = %v", err)
			}
			if !slices.Equal(resp.Operations, want) {
				t.Errorf("Catalog = %+v, want %+v", resp.Operations, want)
			}
			if len(f.admin.calls) != 0 {
				t.Errorf("the catalog called the admin seam %d times, want 0", len(f.admin.calls))
			}
		})
	}
}

// TestAuthorizingSeamCatalogKeepsAJobSeamFailure proves a failing job
// seam answers its own error, not a catalog of the admin entries alone.
func TestAuthorizingSeamCatalogKeepsAJobSeamFailure(t *testing.T) {
	f := newSeams(t, admitAnn())
	f.jobs.answers[ops.OpCatalog] = fakeAnswer{err: ops.NewUnavailableError("no runner")}
	_, err := f.s.Catalog(context.Background(), ops.Actor{}, ops.CatalogRequest{})
	if !errors.Is(err, ops.ErrUnavailable) {
		t.Errorf("Catalog = %v, want ops.ErrUnavailable", err)
	}
}
