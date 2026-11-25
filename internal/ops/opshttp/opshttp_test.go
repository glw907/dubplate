package opshttp

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// TestServeHTTPRefusesBeforeAuthorizing proves the method, path, query,
// and content type checks answer before the privilege check. Fixture
// state: a member grant, so an admin row that reached admission would
// answer 403 rather than its own refusal.
func TestServeHTTPRefusesBeforeAuthorizing(t *testing.T) {
	idBody := `{"submission_id":"` + testHandle + `"}`
	tests := []struct {
		name        string
		method      string
		target      string
		contentType string
		body        string
		wantStatus  int
		wantCode    ops.Code
	}{
		{"GET on get_submission", http.MethodGet, "/get_submission", "application/json", idBody, http.StatusMethodNotAllowed, ops.CodeInvalidRequest},
		{"PUT on list_submissions", http.MethodPut, "/list_submissions", "application/json", `{}`, http.StatusMethodNotAllowed, ops.CodeInvalidRequest},
		{"unknown name", http.MethodPost, "/restore", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"removed operation", http.MethodPost, "/start_backup", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"removed job_status", http.MethodPost, "/job_status", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"root path", http.MethodPost, "/", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"trailing slash", http.MethodPost, "/list_submissions/", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"subpath", http.MethodPost, "/list_submissions/now", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"no leading slash", http.MethodPost, "list_submissions", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"GET on an unknown name", http.MethodGet, "/restore", "", "", http.StatusNotFound, ops.CodeNotFound},
		{"query string", http.MethodPost, "/list_submissions?actor=local:root", "application/json", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"empty query string", http.MethodPost, "/list_submissions?", "application/json", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"no content type", http.MethodPost, "/list_submissions", "", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"unparseable content type", http.MethodPost, "/list_submissions", "application/", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"form content type", http.MethodPost, "/list_submissions", "application/x-www-form-urlencoded", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"GET on list_members", http.MethodGet, "/list_members", "application/json", `{}`, http.StatusMethodNotAllowed, ops.CodeInvalidRequest},
		{"query string on set_setting", http.MethodPost, "/set_setting?key=K", "application/json", `{"key":"K","value":"v"}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"no content type on add_member", http.MethodPost, "/add_member", "", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"unknown operation beside the admin ones", http.MethodPost, "/list_member", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, grant(ann, ops.PrivilegeMember))
			rec := f.do(tt.method, tt.target, tt.contentType, tt.body)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body %q", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if code, _, _ := errorAnswer(t, rec); code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if tt.wantStatus == http.StatusMethodNotAllowed {
				if allow := rec.Header().Get("Allow"); allow != "POST" {
					t.Errorf("Allow = %q, want POST", allow)
				}
			}
			if n := operationCalls(f); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestServeHTTPRefusesABadBodyAfterAdmitting posts each malformed body
// to a member operation as a member who holds its privilege. The
// mount verifies the caller once, answers 400, and no seam sees the
// call. Fixture state: a member grant, so the refusal can come only
// from the body.
func TestServeHTTPRefusesABadBodyAfterAdmitting(t *testing.T) {
	tests := []struct {
		name string
		op   ops.Op
		body string
	}{
		{"empty body", ops.OpListSubmissions, ""},
		{"body one byte over the cap", ops.OpListSubmissions, paddedObject(4097)},
		{"null body", ops.OpListSubmissions, `null`},
		{"array body", ops.OpListSubmissions, `[]`},
		{"string body", ops.OpListSubmissions, `"start"`},
		{"number body", ops.OpListSubmissions, `1`},
		{"second object after the first", ops.OpListSubmissions, `{} {}`},
		{"garbage after the object", ops.OpListSubmissions, `{}x`},
		{"wrongly typed field", ops.OpGetSubmission, `{"submission_id":5}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, grant(ann, ops.PrivilegeMember))
			rec := f.post(tt.op, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body %q", rec.Code, rec.Body.String())
			}
			if code, _, _ := errorAnswer(t, rec); code != ops.CodeInvalidRequest {
				t.Errorf("code = %q, want %q", code, ops.CodeInvalidRequest)
			}
			if f.verified != 1 {
				t.Errorf("verifier calls = %d, want 1", f.verified)
			}
			if n := operationCalls(f); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestServeHTTPRefusesAMalformedID posts each malformed id an id field
// can meet, to a member operation and to an admin one, as an admin. The
// decoder answers 400 with the id sentence and no seam sees the call.
// Fixture state: an admin grant, so the refusal can come only from the
// id's text.
func TestServeHTTPRefusesAMalformedID(t *testing.T) {
	const want = "request body names a malformed id: want 32 lowercase hex characters"
	for _, field := range []struct {
		op  ops.Op
		key string
	}{
		{ops.OpGetSubmission, "submission_id"},
		{ops.OpRemoveMember, "user_id"},
		{ops.OpDemoteMember, "user_id"},
	} {
		for _, bad := range []string{
			"",
			"0123456789abcdef0123456789abcde",
			"0123456789ABCDEF0123456789ABCDEF",
			"0123456789abcdef0123456789abcdeg",
		} {
			t.Run(string(field.op)+" "+bad, func(t *testing.T) {
				f := newFixture(t, admitAnn())
				rec := f.post(field.op, `{"`+field.key+`":"`+bad+`"}`)
				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want 400; body %q", rec.Code, rec.Body.String())
				}
				if code, reason, _ := errorAnswer(t, rec); code != ops.CodeInvalidRequest || reason != want {
					t.Errorf("answer = code %q, reason %q; want code %q, reason %q", code, reason, ops.CodeInvalidRequest, want)
				}
				if n := operationCalls(f); n != 0 {
					t.Errorf("seam calls = %d, want 0", n)
				}
			})
		}
	}
}

// TestServeHTTPRefusesAnUnprivilegedCallerBeforeTheBody proves an admin
// operation checks the caller's privilege before it reads the body.
// Fixture state: a caller granted member, posting four bodies that
// decode to nothing valid, to an admin operation that carries fields
// and one that carries none. Each answers 403 with the verifier
// asked once, so a malformed body never tells a member whether the
// operation exists for them.
func TestServeHTTPRefusesAnUnprivilegedCallerBeforeTheBody(t *testing.T) {
	bodies := []struct{ name, body string }{
		{"an array opening", `[`},
		{"null", `null`},
		{"an empty array", `[]`},
		{"an empty body", ``},
	}
	for _, op := range []ops.Op{ops.OpAddMember, ops.OpSetSetting} {
		for _, b := range bodies {
			t.Run(string(op)+" "+b.name, func(t *testing.T) {
				f := newFixture(t, grant(ann, ops.PrivilegeMember))
				rec := f.post(op, b.body)
				if rec.Code != http.StatusForbidden {
					t.Errorf("status = %d, want 403; body %q", rec.Code, rec.Body.String())
				}
				if code, _, _ := errorAnswer(t, rec); code != ops.CodeForbidden {
					t.Errorf("code = %q, want %q", code, ops.CodeForbidden)
				}
				if want := "granted privilege does not reach the operation's"; !strings.Contains(f.log.String(), want) {
					t.Errorf("log = %q, want it to name %q", f.log.String(), want)
				}
				if f.verified != 1 {
					t.Errorf("verifier calls = %d, want 1", f.verified)
				}
				if n := operationCalls(f); n != 0 {
					t.Errorf("seam calls = %d, want 0", n)
				}
			})
		}
	}
}

// TestServeHTTPAdmitsBeforeDecodingAnAdminBody proves a caller who holds
// the privilege is asked once and is then refused for its body, which
// the seam never sees. Fixture state: an admin posting a body of the
// wrong shape to each operation the refusal table moved here.
func TestServeHTTPAdmitsBeforeDecodingAnAdminBody(t *testing.T) {
	tests := []struct {
		name string
		op   ops.Op
		body string
	}{
		{"array body on remove_member", ops.OpRemoveMember, `[]`},
		{"wrongly typed field on reset_setting", ops.OpResetSetting, `{"key":5}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, admitAnn())
			rec := f.post(tt.op, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body %q", rec.Code, rec.Body.String())
			}
			if code, _, _ := errorAnswer(t, rec); code != ops.CodeInvalidRequest {
				t.Errorf("code = %q, want %q", code, ops.CodeInvalidRequest)
			}
			if f.verified != 1 {
				t.Errorf("verifier calls = %d, want 1", f.verified)
			}
			if n := operationCalls(f); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestServeHTTPRoutesAdminOperations posts each admin operation to the
// mount and proves the admin seam, not the submission seam, receives
// the decoded request as the verified actor, with its session anchor,
// and that the answer is the seam's response.
func TestServeHTTPRoutesAdminOperations(t *testing.T) {
	tests := []struct {
		op       ops.Op
		body     string
		wantReq  any
		wantBody string
	}{
		{ops.OpListMembers, `{}`, ops.ListMembersRequest{},
			`{"members":[{"user_id":"` + testHandle + `","address":"ann@example.com","display_name":"","username":"ann","admin":true}]}`},
		{ops.OpAddMember, `{"address":"bo@example.com","display_name":"Bo","username":"bo"}`,
			ops.AddMemberRequest{Address: "bo@example.com", DisplayName: "Bo", Username: "bo"},
			`{"member":{"user_id":"` + testHandle + `","address":"bo@example.com","display_name":"","username":"bo","admin":false}}`},
		{ops.OpRemoveMember, `{"user_id":"` + testHandle + `"}`, ops.RemoveMemberRequest{UserID: testID}, `{}`},
		{ops.OpDemoteMember, `{"user_id":"` + testHandle + `"}`, ops.DemoteMemberRequest{UserID: testID}, `{}`},
		{ops.OpListSettings, `{}`, ops.ListSettingsRequest{},
			`{"settings":[{"key":"DUBPLATE_REAUTH_WINDOW","home":"store","description":"d","apply":"next_operation","sensitivity":"sensitive"}]}`},
		{ops.OpSetSetting, `{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h"}`,
			ops.SetSettingRequest{Key: "DUBPLATE_REAUTH_WINDOW", Value: "3h"},
			`{"key":"DUBPLATE_REAUTH_WINDOW","apply":"next_operation","source":"admin"}`},
		{ops.OpResetSetting, `{"key":"DUBPLATE_REAUTH_WINDOW"}`,
			ops.ResetSettingRequest{Key: "DUBPLATE_REAUTH_WINDOW"},
			`{"key":"DUBPLATE_REAUTH_WINDOW","apply":"next_operation","source":"env_seed"}`},
	}
	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			f := newFixture(t, admitAnn())
			rec := f.post(tt.op, tt.body)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
			if n := len(f.subs.calls); n != 0 {
				t.Errorf("submission seam calls = %d, want 0", n)
			}
			if len(f.admin.calls) != 1 {
				t.Fatalf("admin seam calls = %d, want 1", len(f.admin.calls))
			}
			call := f.admin.calls[0]
			if call.op != tt.op || call.req != tt.wantReq {
				t.Errorf("admin seam got %s %+v, want %s %+v", call.op, call.req, tt.op, tt.wantReq)
			}
			if call.actor.String() != annWire || !call.actor.Anchor().Equal(annAnchor) {
				t.Errorf("admin seam got actor %q anchored %v, want %q anchored %v", call.actor, call.actor.Anchor(), annWire, annAnchor)
			}
		})

	}
}

// TestServeHTTPAddMemberIgnoresAnAdminField proves the add operation
// cannot be asked to grant admin: a body naming the flag, or an actor,
// reaches the seam as the request without either.
func TestServeHTTPAddMemberIgnoresAnAdminField(t *testing.T) {
	f := newFixture(t, admitAnn())
	body := `{"address":"bo@example.com","display_name":"Bo","username":"bo","admin":true,"actor":"local:root"}`
	if rec := f.post(ops.OpAddMember, body); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
	}
	want := ops.AddMemberRequest{Address: "bo@example.com", DisplayName: "Bo", Username: "bo"}
	if len(f.admin.calls) != 1 || f.admin.calls[0].req != want || f.admin.calls[0].actor.String() != annWire {
		t.Errorf("admin seam calls = %+v, want one %+v as %q", f.admin.calls, want, annWire)
	}
}

// TestServeHTTPErrorDetail proves an error's detail reaches the body
// beside its code and reason, through a wrapping error, and that a
// coded error with no detail writes no detail key.
func TestServeHTTPErrorDetail(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"conflict with detail", ops.NewError(ops.CodeConflict, "no admin would remain").WithDetail(ops.DetailLastAdmin), 409,
			`{"error":{"code":"conflict","reason":"no admin would remain","detail":"last_admin"}}`},
		{"detail through a wrapper", fmt.Errorf("remove: %w", ops.NewError(ops.CodeConflict, "taken").WithDetail(ops.DetailDuplicateAddress)), 409,
			`{"error":{"code":"conflict","reason":"taken","detail":"duplicate_address"}}`},
		{"invalid request with detail", ops.NewError(ops.CodeInvalidRequest, "bad name").WithDetail(ops.DetailInvalidUsername), 400,
			`{"error":{"code":"invalid_request","reason":"bad name","detail":"invalid_username"}}`},
		{"unavailable with detail", ops.NewError(ops.CodeUnavailable, "store busy").WithDetail(ops.DetailBusy), 503,
			`{"error":{"code":"unavailable","reason":"store busy","detail":"busy"}}`},
		{"no detail", ops.NewError(ops.CodeConflict, "elsewhere"), 409,
			`{"error":{"code":"conflict","reason":"elsewhere"}}`},
		{"internal never carries one", ops.NewError(ops.CodeInternal, "secret path").WithDetail(ops.DetailBusy), 500,
			`{"error":{"code":"internal"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, admitAnn())
			f.admin.answers[ops.OpRemoveMember] = fakeAnswer{err: tt.err}
			rec := f.post(ops.OpRemoveMember, `{"user_id":"`+testHandle+`"}`)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
			if _, present := errorDetailOf(t, rec); present != strings.Contains(tt.wantBody, `"detail"`) {
				t.Errorf("detail key present = %v for body %s", present, rec.Body.String())
			}
		})
	}
}

func TestServeHTTPReadsABodyAtTheCap(t *testing.T) {
	body := paddedObject(4096)
	if len(body) != 4096 {
		t.Fatalf("fixture body is %d bytes, want 4096", len(body))
	}
	f := newFixture(t, admitAnn())
	rec := f.post(ops.OpListMembers, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
	}
	if n := len(f.admin.calls); n != 1 {
		t.Errorf("seam calls = %d, want 1", n)
	}
}

func TestServeHTTPIgnoresUnknownKeys(t *testing.T) {
	tests := []struct {
		name string
		op   ops.Op
		body string
	}{
		{"actor", ops.OpListMembers, `{"actor":"local:root"}`},
		{"ACTOR", ops.OpListMembers, `{"ACTOR":"local:root"}`},
		{"privilege", ops.OpListSubmissions, `{"privilege":"admin"}`},
		{"actor beside a handle", ops.OpRemoveMember, `{"user_id":"` + testHandle + `","actor":"local:root"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, admitAnn())
			rec := f.post(tt.op, tt.body)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
			}
			calls := slices.Concat(f.admin.calls, f.subs.calls)
			if len(calls) != 1 {
				t.Fatalf("seam calls = %d, want 1", len(calls))
			}
			call := calls[0]
			if call.op != tt.op || call.actor.String() != annWire {
				t.Errorf("seam got %s as %q, want %s as %q", call.op, call.actor, tt.op, annWire)
			}
		})
	}
}

func TestServeHTTPPassesTheDecodedRequest(t *testing.T) {
	f := newFixture(t, admitAnn())
	rec := f.post(ops.OpGetSubmission, `{"submission_id":"`+testHandle+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
	}
	want := ops.GetSubmissionRequest{SubmissionID: testID}
	if len(f.subs.calls) != 1 || f.subs.calls[0].req != want {
		t.Fatalf("seam calls = %+v, want one carrying %+v", f.subs.calls, want)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	if got["submission_id"] != testHandle || got["state"] != "rejected" {
		t.Errorf("body = %q, want the rejected submission %s", rec.Body.String(), testHandle)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

// TestServeHTTPDenials refuses every denial the same way on the wire
// and logs it by a cause from a closed set. Each row plants an address
// and a principal in the actor the verifier answered, where a denial
// could echo them; neither may reach the log.
func TestServeHTTPDenials(t *testing.T) {
	const (
		address   = "planted@example.invalid"
		principal = "plantedprincipal"
	)
	planted := ann
	planted.Address = address
	planted.Principal = principal
	tests := []struct {
		name       string
		decide     Verifier
		wantReason string
	}{
		{
			name:       "member grant on an admin operation",
			decide:     grant(planted, ops.PrivilegeMember),
			wantReason: "granted privilege does not reach the operation's",
		},
		{
			name:       "zero grant",
			decide:     grant(planted, 0),
			wantReason: "granted privilege does not reach the operation's",
		},
		{
			name: "command-line actor on a cookie mount",
			decide: func(http.ResponseWriter, *http.Request) (identity.Actor, ops.Privilege, error) {
				return identity.LocalActor(principal), ops.PrivilegeAdmin, nil
			},
			wantReason: "admitted actor's channel differs from the mount's",
		},
		{
			name: "zero actor admitted",
			decide: func(http.ResponseWriter, *http.Request) (identity.Actor, ops.Privilege, error) {
				return identity.Actor{}, ops.PrivilegeAdmin, nil
			},
			wantReason: "admitted actor's channel differs from the mount's",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, tt.decide)
			rec := f.post(ops.OpListMembers, `{}`)
			if rec.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403; body %q", rec.Code, rec.Body.String())
			}
			code, reason, _ := errorAnswer(t, rec)
			if code != ops.CodeForbidden || reason != "list_members requires the admin privilege" {
				t.Errorf("answer = %q %q, want forbidden %q", code, reason, "list_members requires the admin privilege")
			}
			log := f.log.String()
			for _, want := range []string{"op=list_members", "code=forbidden", `reason="` + tt.wantReason + `"`} {
				if !strings.Contains(log, want) {
					t.Errorf("log %q lacks %q", log, want)
				}
			}
			for _, leak := range []string{address, principal, "/var/lib"} {
				if strings.Contains(log, leak) || strings.Contains(rec.Body.String(), leak) {
					t.Errorf("log %q or body %q names %q", log, rec.Body.String(), leak)
				}
			}
			if f.verified != 1 {
				t.Errorf("verifier calls = %d, want 1", f.verified)
			}
			if n := operationCalls(f); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestServeHTTPAdmitsByGrant runs each operation under an admin grant
// and a member grant: admin reaches every operation, member the
// submission operations only. A refused call reaches no seam.
func TestServeHTTPAdmitsByGrant(t *testing.T) {
	userBody := `{"user_id":"` + testHandle + `"}`
	submissionBody := `{"submission_id":"` + testHandle + `"}`
	tests := []struct {
		op           ops.Op
		body         string
		memberStatus int
	}{
		{ops.OpListMembers, `{}`, http.StatusForbidden},
		{ops.OpAddMember, `{"address":"bo@example.com","display_name":"Bo","username":"bo"}`, http.StatusForbidden},
		{ops.OpRemoveMember, userBody, http.StatusForbidden},
		{ops.OpDemoteMember, userBody, http.StatusForbidden},
		{ops.OpListSettings, `{}`, http.StatusForbidden},
		{ops.OpSetSetting, `{"key":"K","value":"v"}`, http.StatusForbidden},
		{ops.OpResetSetting, `{"key":"K"}`, http.StatusForbidden},
		{ops.OpCreateSubmission, `{"verdict_unit":"album","container":"files","files":[],"source":"vinyl"}`, http.StatusOK},
		{ops.OpListSubmissions, `{}`, http.StatusOK},
		{ops.OpGetSubmission, submissionBody, http.StatusOK},
		{ops.OpRetrySubmission, submissionBody, http.StatusOK},
		{ops.OpCancelSubmission, submissionBody, http.StatusOK},
	}
	for _, tt := range tests {
		for _, g := range []struct {
			granted ops.Privilege
			status  int
		}{
			{ops.PrivilegeAdmin, http.StatusOK},
			{ops.PrivilegeMember, tt.memberStatus},
		} {
			t.Run(string(tt.op)+" as "+g.granted.String(), func(t *testing.T) {
				f := newFixture(t, grant(ann, g.granted))
				rec := f.post(tt.op, tt.body)
				if rec.Code != g.status {
					t.Fatalf("status = %d, want %d; body %q", rec.Code, g.status, rec.Body.String())
				}
				if f.verified != 1 {
					t.Errorf("verifier calls = %d, want 1", f.verified)
				}
				wantSeamCalls := 0
				if g.status == http.StatusOK {
					wantSeamCalls = 1
				}
				calls := slices.Concat(f.admin.calls, f.subs.calls)
				if len(calls) != wantSeamCalls {
					t.Fatalf("seam calls = %d, want %d", len(calls), wantSeamCalls)
				}
				if wantSeamCalls == 1 && calls[0].actor.String() != annWire {
					t.Errorf("seam got %q, want %q", calls[0].actor, annWire)
				}
			})
		}
	}
}

func TestServeHTTPOperationErrors(t *testing.T) {
	const path = "/var/lib/dubplate/db/dubplate.db"
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   ops.Code
		wantReason string // "" means the body carries no reason key
		wantLog    string // "" means the log carries the error's own text
	}{
		{"unavailable", ops.NewError(ops.CodeUnavailable, "backup integration is off"), 503, ops.CodeUnavailable, "backup integration is off", ""},
		{"not found", ops.NewError(ops.CodeNotFound, "no job "+testHandle), 404, ops.CodeNotFound, "no job " + testHandle, ""},
		{"conflict", ops.NewError(ops.CodeConflict, "backup job already running"), 409, ops.CodeConflict, "backup job already running", ""},
		{"invalid request", ops.NewError(ops.CodeInvalidRequest, "bad field"), 400, ops.CodeInvalidRequest, "bad field", ""},
		{"canceled", ops.NewError(ops.CodeCanceled, "caller went away"), 503, ops.CodeCanceled, "caller went away", ""},
		{"forbidden", ops.NewError(ops.CodeForbidden, "not yours"), 403, ops.CodeForbidden, "not yours", `reason=denied`},
		{"unauthenticated", ops.NewError(ops.CodeUnauthenticated, "sign in again"), 401, ops.CodeUnauthenticated, "sign in again", ""},
		{"reauthentication required", ops.NewError(ops.CodeReauthenticationRequired, "sign in from this browser"), 403, ops.CodeReauthenticationRequired, "sign in from this browser", ""},
		{"coded internal", ops.NewError(ops.CodeInternal, "store closed at "+path), 500, ops.CodeInternal, "", ""},
		{"bare context canceled", fmt.Errorf("wait for the write connection: %w", context.Canceled), 503, ops.CodeCanceled, "", ""},
		{"coded error wrapped with a path", fmt.Errorf("lease %s: %w", path, ops.NewError(ops.CodeConflict, "backup job already running")), 409, ops.CodeConflict, "backup job already running", ""},
		{"uncoded error", fmt.Errorf("open %s: permission denied", path), 500, ops.CodeInternal, "", ""},
		{"zero ops error", &ops.Error{}, 500, ops.CodeInternal, "", `msg="ops error code has no HTTP status" op=list_members code=""`},
		{"unregistered code", &ops.Error{Code: "teapot", Reason: "brewing at " + path}, 500, ops.CodeInternal, "", "code=teapot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, admitAnn())
			f.admin.answers[ops.OpListMembers] = fakeAnswer{err: tt.err}
			rec := f.post(ops.OpListMembers, `{}`)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			code, reason, hasReason := errorAnswer(t, rec)
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if tt.wantReason == "" && hasReason {
				t.Errorf("body %q carries a reason, want none", rec.Body.String())
			}
			if reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", reason, tt.wantReason)
			}
			if strings.Contains(rec.Body.String(), path) {
				t.Errorf("body %q names %s", rec.Body.String(), path)
			}
			wantLog := cmp.Or(tt.wantLog, tt.err.Error())
			if !strings.Contains(f.log.String(), wantLog) {
				t.Errorf("log %q lacks %q", f.log.String(), wantLog)
			}
			if !strings.Contains(f.log.String(), "op=list_members") {
				t.Errorf("log %q lacks the operation", f.log.String())
			}
		})
	}
}

func TestStatuses(t *testing.T) {
	want := map[ops.Code]int{
		"invalid_request":           400,
		"forbidden":                 403,
		"not_found":                 404,
		"conflict":                  409,
		"unavailable":               503,
		"canceled":                  503,
		"internal":                  500,
		"unauthenticated":           401,
		"reauthentication_required": 403,
	}
	if !maps.Equal(statuses, want) {
		t.Errorf("statuses = %v, want %v", statuses, want)
	}
}

// TestStatusesCoverEveryCode reads the Code constants from package ops's
// source, so a code added there with no status here fails.
func TestStatusesCoverEveryCode(t *testing.T) {
	codes := goConstants(t, "../errors.go", "Code")
	if !slices.Contains(codes, "canceled") {
		t.Fatalf("read %v from errors.go, want the Code constants including canceled", codes)
	}
	for _, c := range codes {
		if _, ok := statuses[ops.Code(c)]; !ok {
			t.Errorf("wire code %q has no HTTP status", c)
		}
	}
}
