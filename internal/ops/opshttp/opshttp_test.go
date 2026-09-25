package opshttp

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

func TestServeHTTPRefusesBeforeAuthorizing(t *testing.T) {
	jobBody := `{"job_id":"` + testHandle + `"}`
	tests := []struct {
		name        string
		method      string
		target      string
		contentType string
		body        string
		wantStatus  int
		wantCode    ops.Code
	}{
		{"GET on job_status", http.MethodGet, "/job_status", "application/json", jobBody, http.StatusMethodNotAllowed, ops.CodeInvalidRequest},
		{"PUT on start_backup", http.MethodPut, "/start_backup", "application/json", `{}`, http.StatusMethodNotAllowed, ops.CodeInvalidRequest},
		{"unknown name", http.MethodPost, "/restore", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"root path", http.MethodPost, "/", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"trailing slash", http.MethodPost, "/start_backup/", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"subpath", http.MethodPost, "/start_backup/now", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"no leading slash", http.MethodPost, "start_backup", "application/json", `{}`, http.StatusNotFound, ops.CodeNotFound},
		{"GET on an unknown name", http.MethodGet, "/restore", "", "", http.StatusNotFound, ops.CodeNotFound},
		{"query string", http.MethodPost, "/start_backup?actor=local:root", "application/json", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"empty query string", http.MethodPost, "/start_backup?", "application/json", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"no content type", http.MethodPost, "/start_backup", "", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"unparseable content type", http.MethodPost, "/start_backup", "application/", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"form content type", http.MethodPost, "/start_backup", "application/x-www-form-urlencoded", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"empty body", http.MethodPost, "/start_backup", "application/json", "", http.StatusBadRequest, ops.CodeInvalidRequest},
		{"body one byte over the cap", http.MethodPost, "/start_backup", "application/json", paddedObject(4097), http.StatusBadRequest, ops.CodeInvalidRequest},
		{"null body", http.MethodPost, "/start_backup", "application/json", `null`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"array body", http.MethodPost, "/start_backup", "application/json", `[]`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"string body", http.MethodPost, "/start_backup", "application/json", `"start"`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"number body", http.MethodPost, "/start_backup", "application/json", `1`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"second object after the first", http.MethodPost, "/start_backup", "application/json", `{} {}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"garbage after the object", http.MethodPost, "/start_backup", "application/json", `{}x`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"wrongly typed field", http.MethodPost, "/job_status", "application/json", `{"job_id":5}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"malformed handle on job_status", http.MethodPost, "/job_status", "application/json", `{"job_id":"ABC"}`, http.StatusBadRequest, ops.CodeInvalidRequest},
		{"missing handle on cancel_job", http.MethodPost, "/cancel_job", "application/json", `{}`, http.StatusBadRequest, ops.CodeInvalidRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, admitAnn(), successes())
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
			if n := len(f.auth.calls); n != 0 {
				t.Errorf("authorizer calls = %d, want 0", n)
			}
			if n := len(f.seam.calls); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

func TestServeHTTPReadsABodyAtTheCap(t *testing.T) {
	body := paddedObject(4096)
	if len(body) != 4096 {
		t.Fatalf("fixture body is %d bytes, want 4096", len(body))
	}
	f := newFixture(t, admitAnn(), successes())
	rec := f.post(ops.OpStartBackup, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
	}
	if n := len(f.seam.calls); n != 1 {
		t.Errorf("seam calls = %d, want 1", n)
	}
}

func TestServeHTTPIgnoresUnknownKeys(t *testing.T) {
	tests := []struct {
		name string
		op   string
		body string
	}{
		{"actor", ops.OpStartBackup, `{"actor":"local:root"}`},
		{"ACTOR", ops.OpStartBackup, `{"ACTOR":"local:root"}`},
		{"privilege", ops.OpCatalog, `{"privilege":"admin"}`},
		{"actor beside a handle", ops.OpCancelJob, `{"job_id":"` + testHandle + `","actor":"local:root"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, admitAnn(), successes())
			rec := f.post(tt.op, tt.body)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
			}
			if len(f.seam.calls) != 1 {
				t.Fatalf("seam calls = %d, want 1", len(f.seam.calls))
			}
			call := f.seam.calls[0]
			if call.op != tt.op || call.actor.String() != annWire {
				t.Errorf("seam got %s as %q, want %s as %q", call.op, call.actor, tt.op, annWire)
			}
		})
	}
}

func TestServeHTTPPassesTheDecodedRequest(t *testing.T) {
	f := newFixture(t, admitAnn(), successes())
	rec := f.post(ops.OpJobStatus, `{"job_id":"`+testHandle+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body.String())
	}
	want := ops.JobStatusRequest{JobID: testHandle}
	if len(f.seam.calls) != 1 || f.seam.calls[0].req != want {
		t.Fatalf("seam calls = %+v, want one carrying %+v", f.seam.calls, want)
	}
	var got map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	if got["job"]["id"] != testHandle || got["job"]["state"] != "running" {
		t.Errorf("body = %q, want the running job %s", rec.Body.String(), testHandle)
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
// and a principal where a denial could echo them, in the authorizer's
// error or in the actor it answered; neither may reach the log.
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
		decide     func(op string) (ops.Actor, ops.Privilege, error)
		wantReason string
	}{
		{
			name: "authorizer answers forbidden",
			decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.Actor{}, 0, ops.NewForbiddenError(address + " (" + principal + ") is not an admin")
			},
			wantReason: "authorizer denied the caller",
		},
		{
			name: "authorizer error",
			decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.Actor{}, 0, errors.New("read session /var/lib/dubplate/sessions/" + address + ": permission denied")
			},
			wantReason: "authorizer denied the caller",
		},
		{
			name: "member grant on an admin operation",
			decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.NetworkActor(planted), ops.PrivilegeMember, nil
			},
			wantReason: "granted privilege does not reach the operation's",
		},
		{
			name: "zero grant",
			decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.NetworkActor(planted), 0, nil
			},
			wantReason: "granted privilege does not reach the operation's",
		},
		{
			name: "command-line actor on a cookie mount",
			decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.LocalActor(principal), ops.PrivilegeAdmin, nil
			},
			wantReason: "admitted actor's channel differs from the mount's",
		},
		{
			name: "zero actor admitted",
			decide: func(string) (ops.Actor, ops.Privilege, error) {
				return ops.Actor{}, ops.PrivilegeAdmin, nil
			},
			wantReason: "admitted actor's channel differs from the mount's",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, &fakeAuthorizer{decide: tt.decide}, successes())
			rec := f.post(ops.OpStartBackup, `{}`)
			if rec.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403; body %q", rec.Code, rec.Body.String())
			}
			code, reason, _ := errorAnswer(t, rec)
			if code != ops.CodeForbidden || reason != "start_backup requires the admin privilege" {
				t.Errorf("answer = %q %q, want forbidden %q", code, reason, "start_backup requires the admin privilege")
			}
			log := f.log.String()
			for _, want := range []string{"op=start_backup", "code=forbidden", `reason="` + tt.wantReason + `"`} {
				if !strings.Contains(log, want) {
					t.Errorf("log %q lacks %q", log, want)
				}
			}
			for _, leak := range []string{address, principal, "/var/lib"} {
				if strings.Contains(log, leak) || strings.Contains(rec.Body.String(), leak) {
					t.Errorf("log %q or body %q names %q", log, rec.Body.String(), leak)
				}
			}
			if n := len(f.auth.calls); n != 1 {
				t.Errorf("authorizer calls = %d, want 1", n)
			}
			if n := len(f.seam.calls); n != 0 {
				t.Errorf("seam calls = %d, want 0", n)
			}
		})
	}
}

// TestServeHTTPAdmitsByGrant runs each operation under an admin grant
// and a member grant: admin reaches all four, member only the catalog.
func TestServeHTTPAdmitsByGrant(t *testing.T) {
	jobBody := `{"job_id":"` + testHandle + `"}`
	tests := []struct {
		op           string
		body         string
		memberStatus int
	}{
		{ops.OpStartBackup, `{}`, http.StatusForbidden},
		{ops.OpJobStatus, jobBody, http.StatusForbidden},
		{ops.OpCancelJob, jobBody, http.StatusForbidden},
		{ops.OpCatalog, `{}`, http.StatusOK},
	}
	for _, tt := range tests {
		for _, g := range []struct {
			granted ops.Privilege
			status  int
		}{
			{ops.PrivilegeAdmin, http.StatusOK},
			{ops.PrivilegeMember, tt.memberStatus},
		} {
			t.Run(tt.op+" as "+g.granted.String(), func(t *testing.T) {
				f := newFixture(t, grant(ann, g.granted), successes())
				rec := f.post(tt.op, tt.body)
				if rec.Code != g.status {
					t.Fatalf("status = %d, want %d; body %q", rec.Code, g.status, rec.Body.String())
				}
				if len(f.auth.calls) != 1 || f.auth.calls[0] != tt.op {
					t.Errorf("authorizer calls = %q, want [%s]", f.auth.calls, tt.op)
				}
				wantSeamCalls := 0
				if g.status == http.StatusOK {
					wantSeamCalls = 1
				}
				if len(f.seam.calls) != wantSeamCalls {
					t.Fatalf("seam calls = %d, want %d", len(f.seam.calls), wantSeamCalls)
				}
				if wantSeamCalls == 1 && f.seam.calls[0].actor.String() != annWire {
					t.Errorf("seam got %q, want %q", f.seam.calls[0].actor, annWire)
				}
			})
		}
	}
}

func TestServeHTTPSeamErrors(t *testing.T) {
	const path = "/var/lib/dubplate/db/dubplate.db"
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   ops.Code
		wantReason string // "" means the body carries no reason key
		wantLog    string // "" means the log carries the error's own text
	}{
		{"unavailable", ops.NewUnavailableError("backup integration is off"), 503, ops.CodeUnavailable, "backup integration is off", ""},
		{"not found", ops.NewNotFoundError("no job " + testHandle), 404, ops.CodeNotFound, "no job " + testHandle, ""},
		{"conflict", ops.NewConflictError("backup job already running"), 409, ops.CodeConflict, "backup job already running", ""},
		{"invalid request", ops.NewInvalidRequestError("bad field"), 400, ops.CodeInvalidRequest, "bad field", ""},
		{"canceled", ops.NewCanceledError("caller went away"), 503, ops.CodeCanceled, "caller went away", ""},
		{"forbidden", ops.NewForbiddenError("not yours"), 403, ops.CodeForbidden, "not yours", `reason=denied`},
		{"coded internal", ops.NewInternalError("store closed at " + path), 500, ops.CodeInternal, "", ""},
		{"bare context canceled", fmt.Errorf("wait for the write connection: %w", context.Canceled), 503, ops.CodeCanceled, "", ""},
		{"coded error wrapped with a path", fmt.Errorf("lease %s: %w", path, ops.NewConflictError("backup job already running")), 409, ops.CodeConflict, "backup job already running", ""},
		{"uncoded error", fmt.Errorf("open %s: permission denied", path), 500, ops.CodeInternal, "", ""},
		{"zero ops error", &ops.Error{}, 500, ops.CodeInternal, "", `msg="ops error code has no HTTP status" op=start_backup code=""`},
		{"unregistered code", &ops.Error{Code: "teapot", Reason: "brewing at " + path}, 500, ops.CodeInternal, "", "code=teapot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			answers := successes()
			answers[ops.OpStartBackup] = fakeAnswer{err: tt.err}
			f := newFixture(t, admitAnn(), answers)
			rec := f.post(ops.OpStartBackup, `{}`)
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
			if !strings.Contains(f.log.String(), "op=start_backup") {
				t.Errorf("log %q lacks the operation", f.log.String())
			}
		})
	}
}

func TestStatuses(t *testing.T) {
	want := map[ops.Code]int{
		"invalid_request": 400,
		"forbidden":       403,
		"not_found":       404,
		"conflict":        409,
		"unavailable":     503,
		"canceled":        503,
		"internal":        500,
	}
	if !maps.Equal(statuses, want) {
		t.Errorf("statuses = %v, want %v", statuses, want)
	}
}

func TestCheckStatusesNamesAMissingCode(t *testing.T) {
	if err := checkStatuses(statuses); err != nil {
		t.Fatalf("checkStatuses(statuses) = %v, want nil", err)
	}
	partial := maps.Clone(statuses)
	delete(partial, ops.CodeCanceled)
	err := checkStatuses(partial)
	if err == nil || !strings.Contains(err.Error(), `"canceled"`) {
		t.Errorf("checkStatuses without canceled = %v, want an error naming \"canceled\"", err)
	}
}

func TestNewHandlerRefuses(t *testing.T) {
	log := slog.New(slog.DiscardHandler)
	tests := []struct {
		name    string
		seam    ops.Seam
		auth    authorizer
		channel identity.Channel
		log     *slog.Logger
		want    string
	}{
		{"nil seam", nil, admitAnn(), identity.ChannelCookie, log, "seam is nil"},
		{"nil authorizer", &fakeSeam{}, nil, identity.ChannelCookie, log, "authorizer is nil"},
		{"nil logger", &fakeSeam{}, admitAnn(), identity.ChannelCookie, nil, "logger is nil"},
		{"zero channel", &fakeSeam{}, admitAnn(), identity.Channel(0), log, "channel 0 is not"},
		{"unknown channel", &fakeSeam{}, admitAnn(), identity.Channel(99), log, "channel 99 is not"},
		{"command-line channel", &fakeSeam{}, admitAnn(), identity.ChannelCommandLine, log, "command line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := newHandler(tt.seam, tt.auth, tt.channel, tt.log)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("newHandler = %v, %v; want an error containing %q", h, err, tt.want)
			}
		})
	}
}
