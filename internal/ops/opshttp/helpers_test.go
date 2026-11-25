package opshttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/reject"
)

// testHandle is a well-formed record id: 32 lowercase hex characters.
// testID holds the same 16 bytes, written out by hand.
const testHandle = "0123456789abcdef0123456789abcdef"

var testID = id.ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}

// ann is the verified member the admitting verifiers answer with.
var ann = identity.Identity{
	Address:     "ann@example.com",
	UserID:      identity.UserID{0x01},
	Principal:   "ann",
	DisplayName: "Ann",
	Source:      "magic-link",
	Channel:     identity.ChannelCookie,
}

// annWire is ann's network actor's wire form, written out rather than
// read back from identity.NetworkActor: member: and the hex of ann's UserID.
const annWire = "member:01000000000000000000000000000000"

// annAnchor is the instant every test session was anchored.
var annAnchor = time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

// A fakeCall is one call a fake seam received.
type fakeCall struct {
	op    ops.Op
	actor identity.Actor
	req   any
}

// A fakeAnswer is what a fake seam answers one operation with.
type fakeAnswer struct {
	resp any
	err  error
}

// A fakeAdmin records every call and the actor it received. It answers
// only an operation it was configured for; any other call fails, so an
// unexpected dispatch can never read as a success.
type fakeAdmin struct {
	answers map[ops.Op]fakeAnswer
	calls   []fakeCall
}

// A fakeSubmissions is the submission seam's counterpart to fakeAdmin.
type fakeSubmissions struct {
	answers map[ops.Op]fakeAnswer
	calls   []fakeCall
}

func answer[T any](answers map[ops.Op]fakeAnswer, calls *[]fakeCall, op ops.Op, actor identity.Actor, req any) (T, error) {
	*calls = append(*calls, fakeCall{op: op, actor: actor, req: req})
	var zero T
	a, ok := answers[op]
	if !ok {
		return zero, fmt.Errorf("fake seam: %s not configured", op)
	}
	if a.err != nil {
		return zero, a.err
	}
	resp, ok := a.resp.(T)
	if !ok {
		return zero, fmt.Errorf("fake seam: %s configured with %T", op, a.resp)
	}
	return resp, nil
}

func (f *fakeAdmin) ListMembers(_ context.Context, actor identity.Actor, req ops.ListMembersRequest) (ops.ListMembersResponse, error) {
	return answer[ops.ListMembersResponse](f.answers, &f.calls, ops.OpListMembers, actor, req)
}

func (f *fakeAdmin) AddMember(_ context.Context, actor identity.Actor, req ops.AddMemberRequest) (ops.AddMemberResponse, error) {
	return answer[ops.AddMemberResponse](f.answers, &f.calls, ops.OpAddMember, actor, req)
}

func (f *fakeAdmin) RemoveMember(_ context.Context, actor identity.Actor, req ops.RemoveMemberRequest) (ops.RemoveMemberResponse, error) {
	return answer[ops.RemoveMemberResponse](f.answers, &f.calls, ops.OpRemoveMember, actor, req)
}

func (f *fakeAdmin) DemoteMember(_ context.Context, actor identity.Actor, req ops.DemoteMemberRequest) (ops.DemoteMemberResponse, error) {
	return answer[ops.DemoteMemberResponse](f.answers, &f.calls, ops.OpDemoteMember, actor, req)
}

func (f *fakeAdmin) ListSettings(_ context.Context, actor identity.Actor, req ops.ListSettingsRequest) (ops.ListSettingsResponse, error) {
	return answer[ops.ListSettingsResponse](f.answers, &f.calls, ops.OpListSettings, actor, req)
}

func (f *fakeAdmin) SetSetting(_ context.Context, actor identity.Actor, req ops.SetSettingRequest) (ops.SetSettingResponse, error) {
	return answer[ops.SetSettingResponse](f.answers, &f.calls, ops.OpSetSetting, actor, req)
}

func (f *fakeAdmin) ResetSetting(_ context.Context, actor identity.Actor, req ops.ResetSettingRequest) (ops.ResetSettingResponse, error) {
	return answer[ops.ResetSettingResponse](f.answers, &f.calls, ops.OpResetSetting, actor, req)
}

func (f *fakeSubmissions) CreateSubmission(_ context.Context, actor identity.Actor, req ops.CreateSubmissionRequest) (ops.CreateSubmissionResponse, error) {
	return answer[ops.CreateSubmissionResponse](f.answers, &f.calls, ops.OpCreateSubmission, actor, req)
}

func (f *fakeSubmissions) ListSubmissions(_ context.Context, actor identity.Actor, req ops.ListSubmissionsRequest) (ops.ListSubmissionsResponse, error) {
	return answer[ops.ListSubmissionsResponse](f.answers, &f.calls, ops.OpListSubmissions, actor, req)
}

func (f *fakeSubmissions) GetSubmission(_ context.Context, actor identity.Actor, req ops.GetSubmissionRequest) (ops.GetSubmissionResponse, error) {
	return answer[ops.GetSubmissionResponse](f.answers, &f.calls, ops.OpGetSubmission, actor, req)
}

func (f *fakeSubmissions) RetrySubmission(_ context.Context, actor identity.Actor, req ops.RetrySubmissionRequest) (ops.RetrySubmissionResponse, error) {
	return answer[ops.RetrySubmissionResponse](f.answers, &f.calls, ops.OpRetrySubmission, actor, req)
}

func (f *fakeSubmissions) CancelSubmission(_ context.Context, actor identity.Actor, req ops.CancelSubmissionRequest) (ops.CancelSubmissionResponse, error) {
	return answer[ops.CancelSubmissionResponse](f.answers, &f.calls, ops.OpCancelSubmission, actor, req)
}

// testSubmission is a submission status every fixture's get, retry, and
// cancel answers with, carrying one of each reason-bearing field.
var testSubmission = ops.GetSubmissionResponse{
	SubmissionID: testID,
	State:        "rejected",
	VerdictUnit:  "album",
	Container:    "files",
	Source:       "vinyl",
	CreatedAt:    annAnchor,
	UpdatedAt:    annAnchor,
	Events: []ops.SubmissionEvent{
		{At: annAnchor, Kind: "uploading"},
		{At: annAnchor, Kind: "rejected", Reason: new(reject.Of(reject.Corrupt))},
	},
	Files: []ops.SubmissionFile{
		{FileID: testID, Path: "a/1.flac", Size: 10, Received: true, Reason: new(reject.Of(reject.Corrupt))},
		{FileID: testID, Path: "a/2.flac", Size: 10, Received: false},
	},
	Warnings: []ops.SubmissionWarning{
		{Reason: reject.Of(reject.ExtrasDropped), Subject: new("scan.jpg")},
		{Reason: reject.Of(reject.NoCoverArt)},
	},
}

// submissionSuccesses configures every submission operation to succeed.
func submissionSuccesses() map[ops.Op]fakeAnswer {
	return map[ops.Op]fakeAnswer{
		ops.OpCreateSubmission: {resp: ops.CreateSubmissionResponse{SubmissionID: testID, Files: []ops.CreatedFile{{Path: "a/1.flac", FileID: testID}}}},
		ops.OpListSubmissions: {resp: ops.ListSubmissionsResponse{Submissions: []ops.SubmissionSummary{
			{SubmissionID: testID, State: "uploading", VerdictUnit: "album", CreatedAt: annAnchor, UpdatedAt: annAnchor}}}},
		ops.OpGetSubmission:    {resp: testSubmission},
		ops.OpRetrySubmission:  {resp: ops.RetrySubmissionResponse(testSubmission)},
		ops.OpCancelSubmission: {resp: ops.CancelSubmissionResponse(testSubmission)},
	}
}

// adminSuccesses configures every admin operation to succeed.
func adminSuccesses() map[ops.Op]fakeAnswer {
	return map[ops.Op]fakeAnswer{
		ops.OpListMembers:  {resp: ops.ListMembersResponse{Members: []ops.Member{{UserID: testID, Address: "ann@example.com", Username: "ann", Admin: true}}}},
		ops.OpAddMember:    {resp: ops.AddMemberResponse{Member: ops.Member{UserID: testID, Address: "bo@example.com", Username: "bo"}}},
		ops.OpRemoveMember: {resp: ops.RemoveMemberResponse{}},
		ops.OpDemoteMember: {resp: ops.DemoteMemberResponse{}},
		ops.OpListSettings: {resp: ops.ListSettingsResponse{Settings: []ops.SettingEntry{{Key: "DUBPLATE_REAUTH_WINDOW", Home: "store", Description: "d", Apply: "next_operation", Sensitivity: "sensitive"}}}},
		ops.OpSetSetting:   {resp: ops.SetSettingResponse{Key: "DUBPLATE_REAUTH_WINDOW", Apply: "next_operation", Source: "admin"}},
		ops.OpResetSetting: {resp: ops.ResetSettingResponse{Key: "DUBPLATE_REAUTH_WINDOW", Apply: "next_operation", Source: "env_seed"}},
	}
}

// mountPrefix is the owned prefix every test mount serves under.
const mountPrefix = "/dub"

// grant is a verifier admitting every request as who's network actor,
// granted p.
func grant(who identity.Identity, p ops.Privilege) Verifier {
	return func(http.ResponseWriter, *http.Request) (identity.Actor, ops.Privilege, error) {
		return identity.NetworkActor(who, annAnchor, "192.0.2.1"), p, nil
	}
}

// admitAnn admits every request as ann, granted admin.
func admitAnn() Verifier {
	return grant(ann, ops.PrivilegeAdmin)
}

// A fixture is a mount over fake seams and a verifier that counts its
// calls, with its log captured. Each seam answers every operation with
// a success until a test sets its answers.
type fixture struct {
	m        *mount
	admin    *fakeAdmin
	subs     *fakeSubmissions
	verified int
	log      *bytes.Buffer
}

// newFixture builds a mount whose verifier answers decide.
func newFixture(t *testing.T, decide Verifier) *fixture {
	t.Helper()
	f := &fixture{
		admin: &fakeAdmin{answers: adminSuccesses()},
		subs:  &fakeSubmissions{answers: submissionSuccesses()},
		log:   &bytes.Buffer{},
	}
	verify := func(w http.ResponseWriter, r *http.Request) (identity.Actor, ops.Privilege, error) {
		f.verified++
		return decide(w, r)
	}
	h, err := NewMount(f.admin, f.subs, mountPrefix, verify, slog.New(slog.NewTextHandler(f.log, nil)))
	if err != nil {
		t.Fatalf("NewMount: %v", err)
	}
	m, ok := h.(*mount)
	if !ok {
		t.Fatalf("NewMount returned %T, want *mount", h)
	}
	f.m = m
	return f
}

// operationCalls is the number of calls both seams under f have
// received.
func operationCalls(f *fixture) int {
	return len(f.admin.calls) + len(f.subs.calls)
}

// do sends one request to target, which follows the mount's v1 base
// as written; an empty contentType sends no Content-Type header at all.
func (f *fixture) do(method, target, contentType, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, mountPrefix+"/api/v1"+target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	f.m.ServeHTTP(rec, req)
	return rec
}

func (f *fixture) post(op ops.Op, body string) *httptest.ResponseRecorder {
	return f.do(http.MethodPost, "/"+string(op), "application/json", body)
}

// send sends one JSON request to path, the whole path the mount sees.
func (f *fixture) send(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/x", strings.NewReader(body))
	req.URL.Path = path
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.m.ServeHTTP(rec, req)
	return rec
}

// errorAnswer decodes rec's JSON error body, failing t when the body is
// not exactly that shape or the answer may be cached.
func errorAnswer(t *testing.T, rec *httptest.ResponseRecorder) (code ops.Code, reason string, hasReason bool) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
	var body map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", rec.Body.String(), err)
	}
	detail, ok := body["error"]
	if !ok || len(body) != 1 {
		t.Fatalf("error body %q: want one key, error", rec.Body.String())
	}
	reason, hasReason = detail["reason"]
	return ops.Code(detail["code"]), reason, hasReason
}

// errorDetailOf reads the optional error.detail of rec's JSON error
// body, and whether the key is present.
func errorDetailOf(t *testing.T, rec *httptest.ResponseRecorder) (detail string, present bool) {
	t.Helper()
	var body map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", rec.Body.String(), err)
	}
	detail, present = body["error"]["detail"]
	return detail, present
}

// paddedObject returns a JSON object of exactly n bytes: one unknown
// key whose string value fills the rest.
func paddedObject(n int) string {
	const frame = `{"pad":""}`
	return `{"pad":"` + strings.Repeat("x", n-len(frame)) + `"}`
}
