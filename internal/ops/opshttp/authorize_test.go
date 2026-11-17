package opshttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/ops"
)

// A seamCall is one operation with a request that reaches the fake
// beneath it.
type seamCall struct {
	op  string
	req any
}

// body is the call's request as the wire carries it.
func (c seamCall) body(t *testing.T) string {
	t.Helper()
	data, err := json.Marshal(c.req)
	if err != nil {
		t.Fatalf("marshal the %s request: %v", c.op, err)
	}
	return string(data)
}

// adminCalls is one call for each of the seven admin operations.
func adminCalls() []seamCall {
	return []seamCall{
		{ops.OpListMembers, ops.ListMembersRequest{}},
		{ops.OpAddMember, ops.AddMemberRequest{Address: "bo@example.com", DisplayName: "Bo", Username: "bo"}},
		{ops.OpRemoveMember, ops.RemoveMemberRequest{UserID: testHandle}},
		{ops.OpDemoteMember, ops.DemoteMemberRequest{UserID: testHandle}},
		{ops.OpListSettings, ops.ListSettingsRequest{}},
		{ops.OpSetSetting, ops.SetSettingRequest{Key: "K", Value: "v"}},
		{ops.OpResetSetting, ops.ResetSettingRequest{Key: "K"}},
	}
}

// submissionCalls is one call for each of the five submission
// operations.
func submissionCalls() []seamCall {
	return []seamCall{
		{ops.OpCreateSubmission, ops.CreateSubmissionRequest{VerdictUnit: "album", Container: "files", Source: "vinyl"}},
		{ops.OpListSubmissions, ops.ListSubmissionsRequest{}},
		{ops.OpGetSubmission, ops.GetSubmissionRequest{SubmissionID: testHandle}},
		{ops.OpRetrySubmission, ops.RetrySubmissionRequest{SubmissionID: testHandle}},
		{ops.OpCancelSubmission, ops.CancelSubmissionRequest{SubmissionID: testHandle}},
	}
}

// jobCalls is one call for each of the four job-seam operations.
func jobCalls() []seamCall {
	return []seamCall{
		{ops.OpStartBackup, ops.StartBackupRequest{}},
		{ops.OpJobStatus, ops.JobStatusRequest{JobID: testHandle}},
		{ops.OpCancelJob, ops.CancelJobRequest{JobID: testHandle}},
		{ops.OpCatalog, ops.CatalogRequest{}},
	}
}

// seamCalls is the number of calls every seam under f has received.
func seamCalls(f *fixture) int {
	return len(f.seam.calls) + len(f.admin.calls) + len(f.subs.calls)
}

// TestAuthorizingSeamDiscardsItsCallersActor posts every one of the
// sixteen operations; the seam beneath must see only the actor the
// authorizer admitted, with its session. No request field names an
// actor, so the handler has none to pass on but the admitted one.
func TestAuthorizingSeamDiscardsItsCallersActor(t *testing.T) {
	for _, tt := range slices.Concat(jobCalls(), adminCalls(), submissionCalls()) {
		t.Run(tt.op, func(t *testing.T) {
			f := newFixture(t, admitAnn(), successes())
			if rec := f.post(tt.op, tt.body(t)); rec.Code != http.StatusOK {
				t.Fatalf("%s answered %d, want 200; body %q", tt.op, rec.Code, rec.Body.String())
			}
			calls := slices.Concat(f.seam.calls, f.admin.calls, f.subs.calls)
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

// TestAuthorizingSeamMemberReachesTheCatalogAndSubmissions is the
// authorizer half of the spec's member criterion: a member caller
// reaches the catalog and the five submission operations, and each
// admin operation and each job operation answers forbidden before any
// seam sees the call. A refused call that reached a fake fails here.
func TestAuthorizingSeamMemberReachesTheCatalogAndSubmissions(t *testing.T) {
	for _, tt := range slices.Concat(jobCalls(), adminCalls(), submissionCalls()) {
		t.Run(tt.op, func(t *testing.T) {
			f := newFixture(t, grant(ann, ops.PrivilegeMember), successes())
			rec := f.post(tt.op, tt.body(t))
			if tt.op == ops.OpCatalog || slices.ContainsFunc(submissionCalls(), func(c seamCall) bool { return c.op == tt.op }) {
				if rec.Code != http.StatusOK {
					t.Fatalf("member's %s answered %d, want admitted; body %q", tt.op, rec.Code, rec.Body.String())
				}
				return
			}
			if rec.Code != http.StatusForbidden {
				t.Errorf("member's %s answered %d, want 403; body %q", tt.op, rec.Code, rec.Body.String())
			}
			if n := seamCalls(f); n != 0 {
				t.Errorf("a refused %s reached a seam %d times", tt.op, n)
			}
		})
	}
}

// TestAuthorizingSeamDenialIsForbidden checks the in-process shape of a
// denial: the error matches ops.ErrForbidden and still carries the
// authorizer's own error, on a job operation and on an admin one.
func TestAuthorizingSeamDenialIsForbidden(t *testing.T) {
	errSession := errors.New("session expired")
	for _, tt := range []seamCall{jobCalls()[2], adminCalls()[1]} {
		t.Run(tt.op, func(t *testing.T) {
			deny := &fakeAuthorizer{decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.Actor{}, 0, errSession
			}}
			f := newFixture(t, deny, successes())
			_, err := f.h.seam.admit(context.Background(), tt.op)
			if !errors.Is(err, ops.ErrForbidden) || !errors.Is(err, errSession) {
				t.Errorf("%s = %v, want ops.ErrForbidden wrapping the authorizer's error", tt.op, err)
			}
			if code := ops.CodeOf(err); code != ops.CodeForbidden {
				t.Errorf("CodeOf = %q, want forbidden", code)
			}
			if n := seamCalls(f); n != 0 {
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
	startBackup, catalog, listMembers, getSubmission := jobCalls()[0], jobCalls()[3], adminCalls()[0], submissionCalls()[2]
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
		{"admin on a submission operation", getSubmission, ops.PrivilegeAdmin, true},
		{"member on a submission operation", getSubmission, ops.PrivilegeMember, true},
		{"zero grant on a submission operation", getSubmission, 0, false},
		{"unknown grant on a submission operation", getSubmission, ops.Privilege(99), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, grant(ann, tt.granted), successes())
			rec := f.post(tt.op.op, tt.op.body(t))
			n := seamCalls(f)
			if tt.admit {
				if rec.Code != http.StatusOK || n != 1 {
					t.Errorf("status = %d, seam calls = %d; want admitted once", rec.Code, n)
				}
				return
			}
			if rec.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403", rec.Code)
			}
			if n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestAuthorizingSeamCatalogJoinsEverySeam proves the catalog answer is
// the job seam's entries followed by one entry per admin operation and
// one per submission operation, whoever asks.
func TestAuthorizingSeamCatalogJoinsEverySeam(t *testing.T) {
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
		{Name: "create_submission", Privilege: ops.PrivilegeMember, Available: true},
		{Name: "list_submissions", Privilege: ops.PrivilegeMember, Available: true},
		{Name: "get_submission", Privilege: ops.PrivilegeMember, Available: true},
		{Name: "retry_submission", Privilege: ops.PrivilegeMember, Available: true},
		{Name: "cancel_submission", Privilege: ops.PrivilegeMember, Available: true},
	})
	for _, granted := range []ops.Privilege{ops.PrivilegeMember, ops.PrivilegeAdmin} {
		t.Run(granted.String(), func(t *testing.T) {
			f := newFixture(t, grant(ann, granted), successes())
			f.seam.answers[ops.OpCatalog] = fakeAnswer{resp: ops.CatalogResponse{Operations: jobEntries}}
			resp, err := f.h.seam.Catalog(context.Background(), ops.Actor{}, ops.CatalogRequest{})
			if err != nil {
				t.Fatalf("Catalog = %v", err)
			}
			if !slices.Equal(resp.Operations, want) {
				t.Errorf("Catalog = %+v, want %+v", resp.Operations, want)
			}
			if n := len(f.admin.calls) + len(f.subs.calls); n != 0 {
				t.Errorf("the catalog called the admin or submission seam %d times, want 0", n)
			}
		})
	}
}

// TestAuthorizingSeamCatalogKeepsAJobSeamFailure proves a failing job
// seam answers its own error, not a catalog of the admin entries alone.
func TestAuthorizingSeamCatalogKeepsAJobSeamFailure(t *testing.T) {
	f := newFixture(t, admitAnn(), successes())
	f.seam.answers[ops.OpCatalog] = fakeAnswer{err: ops.NewUnavailableError("no runner")}
	_, err := f.h.seam.Catalog(context.Background(), ops.Actor{}, ops.CatalogRequest{})
	if !errors.Is(err, ops.ErrUnavailable) {
		t.Errorf("Catalog = %v, want ops.ErrUnavailable", err)
	}
}
