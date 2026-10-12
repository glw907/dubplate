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

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// testHandle is a well-formed job handle: 32 lowercase hex characters.
const testHandle = "0123456789abcdef0123456789abcdef"

// ann is the verified member the admitting authorizers answer with.
var ann = identity.Identity{
	Address:     "ann@example.com",
	UserID:      identity.UserID{0x01},
	Principal:   "ann",
	DisplayName: "Ann",
	Source:      "magic-link",
	Channel:     identity.ChannelCookie,
}

// annWire is ann's network actor's wire form, written out rather than
// read back from ops.NetworkActor: member: and the hex of ann's UserID.
const annWire = "member:01000000000000000000000000000000"

// annAnchor is the instant every test session was anchored.
var annAnchor = time.Date(2026, 10, 11, 9, 30, 0, 0, time.UTC)

// A fakeCall is one call a fake seam received.
type fakeCall struct {
	op    string
	actor ops.Actor
	req   any
}

// A fakeAnswer is what a fake seam answers one operation with.
type fakeAnswer struct {
	resp any
	err  error
}

// A fakeSeam records every call and the actor it received. It answers
// only an operation it was configured for; any other call fails, so an
// unexpected dispatch can never read as a success.
type fakeSeam struct {
	answers map[string]fakeAnswer
	calls   []fakeCall
}

// A fakeAdmin is the admin seam's counterpart to fakeSeam, with the
// same recording and the same refusal of an unconfigured operation.
type fakeAdmin struct {
	answers map[string]fakeAnswer
	calls   []fakeCall
}

func answer[T any](answers map[string]fakeAnswer, calls *[]fakeCall, op string, actor ops.Actor, req any) (T, error) {
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

func (f *fakeSeam) StartBackup(_ context.Context, actor ops.Actor, req ops.StartBackupRequest) (ops.StartBackupResponse, error) {
	return answer[ops.StartBackupResponse](f.answers, &f.calls, ops.OpStartBackup, actor, req)
}

func (f *fakeSeam) JobStatus(_ context.Context, actor ops.Actor, req ops.JobStatusRequest) (ops.JobStatusResponse, error) {
	return answer[ops.JobStatusResponse](f.answers, &f.calls, ops.OpJobStatus, actor, req)
}

func (f *fakeSeam) CancelJob(_ context.Context, actor ops.Actor, req ops.CancelJobRequest) (ops.CancelJobResponse, error) {
	return answer[ops.CancelJobResponse](f.answers, &f.calls, ops.OpCancelJob, actor, req)
}

func (f *fakeSeam) Catalog(_ context.Context, actor ops.Actor, req ops.CatalogRequest) (ops.CatalogResponse, error) {
	return answer[ops.CatalogResponse](f.answers, &f.calls, ops.OpCatalog, actor, req)
}

func (f *fakeAdmin) ListMembers(_ context.Context, actor ops.Actor, req ops.ListMembersRequest) (ops.ListMembersResponse, error) {
	return answer[ops.ListMembersResponse](f.answers, &f.calls, ops.OpListMembers, actor, req)
}

func (f *fakeAdmin) AddMember(_ context.Context, actor ops.Actor, req ops.AddMemberRequest) (ops.AddMemberResponse, error) {
	return answer[ops.AddMemberResponse](f.answers, &f.calls, ops.OpAddMember, actor, req)
}

func (f *fakeAdmin) RemoveMember(_ context.Context, actor ops.Actor, req ops.RemoveMemberRequest) (ops.RemoveMemberResponse, error) {
	return answer[ops.RemoveMemberResponse](f.answers, &f.calls, ops.OpRemoveMember, actor, req)
}

func (f *fakeAdmin) DemoteMember(_ context.Context, actor ops.Actor, req ops.DemoteMemberRequest) (ops.DemoteMemberResponse, error) {
	return answer[ops.DemoteMemberResponse](f.answers, &f.calls, ops.OpDemoteMember, actor, req)
}

func (f *fakeAdmin) ListSettings(_ context.Context, actor ops.Actor, req ops.ListSettingsRequest) (ops.ListSettingsResponse, error) {
	return answer[ops.ListSettingsResponse](f.answers, &f.calls, ops.OpListSettings, actor, req)
}

func (f *fakeAdmin) SetSetting(_ context.Context, actor ops.Actor, req ops.SetSettingRequest) (ops.SetSettingResponse, error) {
	return answer[ops.SetSettingResponse](f.answers, &f.calls, ops.OpSetSetting, actor, req)
}

func (f *fakeAdmin) ResetSetting(_ context.Context, actor ops.Actor, req ops.ResetSettingRequest) (ops.ResetSettingResponse, error) {
	return answer[ops.ResetSettingResponse](f.answers, &f.calls, ops.OpResetSetting, actor, req)
}

// adminSuccesses configures every admin operation to succeed.
func adminSuccesses() map[string]fakeAnswer {
	return map[string]fakeAnswer{
		ops.OpListMembers:  {resp: ops.ListMembersResponse{Members: []ops.Member{{UserID: testHandle, Address: "ann@example.com", Username: "ann", Admin: true}}}},
		ops.OpAddMember:    {resp: ops.AddMemberResponse{Member: ops.Member{UserID: testHandle, Address: "bo@example.com", Username: "bo"}}},
		ops.OpRemoveMember: {resp: ops.RemoveMemberResponse{}},
		ops.OpDemoteMember: {resp: ops.DemoteMemberResponse{}},
		ops.OpListSettings: {resp: ops.ListSettingsResponse{Settings: []ops.SettingEntry{{Name: "DUBPLATE_REAUTH_WINDOW", Home: "store", Description: "d", Apply: "next-operation", Mark: "sensitive"}}}},
		ops.OpSetSetting:   {resp: ops.SetSettingResponse{Key: "DUBPLATE_REAUTH_WINDOW", Apply: "next-operation", Source: "admin"}},
		ops.OpResetSetting: {resp: ops.ResetSettingResponse{Key: "DUBPLATE_REAUTH_WINDOW", Apply: "next-operation", Source: "env-seed"}},
	}
}

// successes configures every operation to succeed.
func successes() map[string]fakeAnswer {
	return map[string]fakeAnswer{
		ops.OpStartBackup: {resp: ops.StartBackupResponse{JobID: testHandle}},
		ops.OpJobStatus:   {resp: ops.JobStatusResponse{Job: ops.JobStatus{ID: testHandle, State: ops.JobRunning}}},
		ops.OpCancelJob:   {resp: ops.CancelJobResponse{Job: ops.JobStatus{ID: testHandle, State: ops.JobCanceled}}},
		ops.OpCatalog:     {resp: ops.CatalogResponse{}},
	}
}

// A fakeAuthorizer answers with decide and records the operation of
// each call.
type fakeAuthorizer struct {
	decide func(op string) (ops.Actor, ops.Privilege, error)
	calls  []string
}

func (a *fakeAuthorizer) authorize(_ context.Context, op string) (ops.Actor, ops.Privilege, error) {
	a.calls = append(a.calls, op)
	return a.decide(op)
}

// grant admits every call as the actor id's network actor, granted p.
func grant(id identity.Identity, p ops.Privilege) *fakeAuthorizer {
	return &fakeAuthorizer{decide: func(string) (ops.Actor, ops.Privilege, error) {
		return ops.NetworkActor(id, annAnchor), p, nil
	}}
}

// admitAnn admits every call as ann, granted admin.
func admitAnn() *fakeAuthorizer {
	return grant(ann, ops.PrivilegeAdmin)
}

// A fixture is a handler over fake seams and a fake authorizer, with
// its log captured. The admin seam answers every operation with a
// success until a test sets admin.answers.
type fixture struct {
	h     *handler
	seam  *fakeSeam
	admin *fakeAdmin
	auth  *fakeAuthorizer
	log   *bytes.Buffer
}

func newFixture(t *testing.T, auth *fakeAuthorizer, answers map[string]fakeAnswer) *fixture {
	t.Helper()
	seam := &fakeSeam{answers: answers}
	admin := &fakeAdmin{answers: adminSuccesses()}
	var buf bytes.Buffer
	h, err := newHandler(seam, admin, auth, identity.ChannelCookie, slog.New(slog.NewTextHandler(&buf, nil)))
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}
	return &fixture{h: h, seam: seam, admin: admin, auth: auth, log: &buf}
}

// do sends one request; an empty contentType sends no Content-Type
// header at all. A target with no leading slash, which
// [httptest.NewRequest] refuses to parse, is set as the request's path
// after the request is built.
func (f *fixture) do(method, target, contentType, body string) *httptest.ResponseRecorder {
	relative, ok := strings.CutPrefix(target, "/")
	req := httptest.NewRequest(method, "/"+relative, strings.NewReader(body))
	if !ok {
		req.URL.Path = target
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	f.h.ServeHTTP(rec, req)
	return rec
}

func (f *fixture) post(op, body string) *httptest.ResponseRecorder {
	return f.do(http.MethodPost, "/"+op, "application/json", body)
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
