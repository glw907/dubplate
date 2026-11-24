package opshttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
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
		{ops.OpRemoveMember, ops.RemoveMemberRequest{UserID: testID}},
		{ops.OpDemoteMember, ops.DemoteMemberRequest{UserID: testID}},
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
		{ops.OpGetSubmission, ops.GetSubmissionRequest{SubmissionID: testID}},
		{ops.OpRetrySubmission, ops.RetrySubmissionRequest{SubmissionID: testID}},
		{ops.OpCancelSubmission, ops.CancelSubmissionRequest{SubmissionID: testID}},
	}
}

// operationCalls is the number of calls every seam under f has received.
func operationCalls(f *fixture) int {
	return len(f.admin.calls) + len(f.subs.calls)
}

// TestAdmissionDiscardsItsCallersActor posts every one of the
// twelve operations; the seam beneath must see only the actor the
// authorizer admitted, with its session. No request field names an
// actor, so the handler has none to pass on but the admitted one.
func TestAdmissionDiscardsItsCallersActor(t *testing.T) {
	for _, tt := range slices.Concat(adminCalls(), submissionCalls()) {
		t.Run(tt.op, func(t *testing.T) {
			f := newFixture(t, admitAnn())
			if rec := f.post(tt.op, tt.body(t)); rec.Code != http.StatusOK {
				t.Fatalf("%s answered %d, want 200; body %q", tt.op, rec.Code, rec.Body.String())
			}
			calls := slices.Concat(f.admin.calls, f.subs.calls)
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

// TestAdmissionMemberReachesTheSubmissions is the authorizer half
// of the spec's member criterion: a member caller reaches the five
// submission operations, and each admin operation answers forbidden
// before any seam sees the call. A refused call that reached a fake
// fails here.
func TestAdmissionMemberReachesTheSubmissions(t *testing.T) {
	for _, tt := range slices.Concat(adminCalls(), submissionCalls()) {
		t.Run(tt.op, func(t *testing.T) {
			f := newFixture(t, grant(ann, ops.PrivilegeMember))
			rec := f.post(tt.op, tt.body(t))
			if slices.ContainsFunc(submissionCalls(), func(c seamCall) bool { return c.op == tt.op }) {
				if rec.Code != http.StatusOK {
					t.Fatalf("member's %s answered %d, want admitted; body %q", tt.op, rec.Code, rec.Body.String())
				}
				return
			}
			if rec.Code != http.StatusForbidden {
				t.Errorf("member's %s answered %d, want 403; body %q", tt.op, rec.Code, rec.Body.String())
			}
			if n := operationCalls(f); n != 0 {
				t.Errorf("a refused %s reached a seam %d times", tt.op, n)
			}
		})
	}
}

// TestAdmissionDenialIsForbidden checks the in-process shape of a
// denial: the error matches ops.ErrForbidden and still carries the
// authorizer's own error, on a submission operation and on an admin one.
func TestAdmissionDenialIsForbidden(t *testing.T) {
	errSession := errors.New("session expired")
	for _, tt := range []seamCall{submissionCalls()[4], adminCalls()[1]} {
		t.Run(tt.op, func(t *testing.T) {
			deny := &fakeAuthorizer{decide: func(string) (identity.Actor, ops.Privilege, error) {
				return identity.Actor{}, 0, errSession
			}}
			f := newFixture(t, deny)
			_, err := f.h.admission.admit(context.Background(), tt.op)
			if !errors.Is(err, ops.ErrForbidden) || !errors.Is(err, errSession) {
				t.Errorf("%s = %v, want ops.ErrForbidden wrapping the authorizer's error", tt.op, err)
			}
			if code := ops.CodeOf(err); code != ops.CodeForbidden {
				t.Errorf("CodeOf = %q, want forbidden", code)
			}
			if n := operationCalls(f); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestAdmissionChecksTheGrant admits a call only when the
// privilege the authorizer granted reaches the one the operation
// requires. Each row's authorizer admits without error, so a denial
// can come only from the seam's own check.
func TestAdmissionChecksTheGrant(t *testing.T) {
	listMembers, getSubmission := adminCalls()[0], submissionCalls()[2]
	tests := []struct {
		name    string
		op      seamCall
		granted ops.Privilege
		admit   bool
	}{
		{"admin on an admin operation", listMembers, ops.PrivilegeAdmin, true},
		{"member on an admin operation", listMembers, ops.PrivilegeMember, false},
		{"zero grant on an admin operation", listMembers, 0, false},
		{"unknown grant on an admin operation", listMembers, ops.Privilege(99), false},
		{"admin on a submission operation", getSubmission, ops.PrivilegeAdmin, true},
		{"member on a submission operation", getSubmission, ops.PrivilegeMember, true},
		{"zero grant on a submission operation", getSubmission, 0, false},
		{"unknown grant on a submission operation", getSubmission, ops.Privilege(99), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, grant(ann, tt.granted))
			rec := f.post(tt.op.op, tt.op.body(t))
			n := operationCalls(f)
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

// TestAdmissionRefusesAnOperationOffTheTable proves an operation
// the privilege table does not hold answers not found before the
// authorizer is asked, so a removed operation cannot be reached by name.
func TestAdmissionRefusesAnOperationOffTheTable(t *testing.T) {
	for _, op := range []string{"start_backup", "job_status", "cancel_job", "catalog", "delete_everything"} {
		t.Run(op, func(t *testing.T) {
			f := newFixture(t, admitAnn())
			_, err := f.h.admission.admit(context.Background(), op)
			if !errors.Is(err, ops.ErrNotFound) {
				t.Errorf("admit(%q) = %v, want ops.ErrNotFound", op, err)
			}
			if len(f.auth.calls) != 0 {
				t.Errorf("the authorizer was asked %d times, want 0", len(f.auth.calls))
			}
		})
	}
}
