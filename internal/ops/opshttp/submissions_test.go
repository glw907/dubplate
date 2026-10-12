package opshttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/reject"
)

// TestServeHTTPRoutesSubmissionOperations posts each of the five
// submission operations and proves the submission seam receives the
// decoded request as the admitted actor, and that the answer is the
// seam's response.
func TestServeHTTPRoutesSubmissionOperations(t *testing.T) {
	idBody := `{"submission_id":"` + testHandle + `"}`
	tests := []struct {
		op      string
		body    string
		wantReq any
		want    func(t *testing.T) any
	}{
		{ops.OpCreateSubmission,
			`{"verdict_unit":"track","container":"zip","files":[{"path":"a/1.flac","size":10}],"source":"vinyl","source_note":"n","store_url":"https://x.example/a"}`,
			ops.CreateSubmissionRequest{VerdictUnit: "track", Container: "zip", Files: []ops.ManifestFile{{Path: "a/1.flac", Size: 10}},
				Source: "vinyl", SourceNote: "n", StoreURL: "https://x.example/a"},
			func(*testing.T) any { return submissionSuccesses()[ops.OpCreateSubmission].resp }},
		{ops.OpListSubmissions, `{}`, ops.ListSubmissionsRequest{},
			func(*testing.T) any { return submissionSuccesses()[ops.OpListSubmissions].resp }},
		{ops.OpGetSubmission, idBody, ops.GetSubmissionRequest{SubmissionID: testHandle},
			func(*testing.T) any { return testSubmission }},
		{ops.OpRetrySubmission, idBody, ops.RetrySubmissionRequest{SubmissionID: testHandle},
			func(*testing.T) any { return testSubmission }},
		{ops.OpCancelSubmission, idBody, ops.CancelSubmissionRequest{SubmissionID: testHandle},
			func(*testing.T) any { return testSubmission }},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			f := newFixture(t, grant(ann, ops.PrivilegeMember), successes())
			rec := f.post(tt.op, tt.body)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body)
			}
			want, err := json.Marshal(tt.want(t))
			if err != nil {
				t.Fatal(err)
			}
			if got := rec.Body.String(); got != string(want) {
				t.Errorf("body = %s, want %s", got, want)
			}
			if n := len(f.seam.calls) + len(f.admin.calls); n != 0 {
				t.Errorf("job and admin seam calls = %d, want 0", n)
			}
			if len(f.subs.calls) != 1 {
				t.Fatalf("submission seam calls = %d, want 1", len(f.subs.calls))
			}
			call := f.subs.calls[0]
			if call.op != tt.op || !reflect.DeepEqual(call.req, tt.wantReq) {
				t.Errorf("submission seam got %s %+v, want %s %+v", call.op, call.req, tt.op, tt.wantReq)
			}
			if call.actor.String() != annWire {
				t.Errorf("submission seam got actor %q, want %q", call.actor, annWire)
			}
		})
	}
}

// TestServeHTTPCreateSubmissionIgnoresForeignFields proves a create
// body naming an uploader, a collection, a state, a reservation, a role,
// and an actor reaches the seam as the request without any of them.
func TestServeHTTPCreateSubmissionIgnoresForeignFields(t *testing.T) {
	f := newFixture(t, grant(ann, ops.PrivilegeMember), successes())
	body := `{"verdict_unit":"album","container":"files","files":[{"path":"a.flac","size":1,"role":"audio"}],"source":"cd",
		"uploader":"` + testHandle + `","collection":"` + testHandle + `","collection_id":"` + testHandle + `",
		"state":"accepted","reserved_bytes":0,"role":"admin","actor":"local:root"}`
	if rec := f.post(ops.OpCreateSubmission, body); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %q", rec.Code, rec.Body)
	}
	want := ops.CreateSubmissionRequest{VerdictUnit: "album", Container: "files", Files: []ops.ManifestFile{{Path: "a.flac", Size: 1}}, Source: "cd"}
	if len(f.subs.calls) != 1 || !reflect.DeepEqual(f.subs.calls[0].req, want) || f.subs.calls[0].actor.String() != annWire {
		t.Errorf("submission seam calls = %+v, want one %+v as %q", f.subs.calls, want, annWire)
	}
}

// manifestBody is a create_submission body declaring n files.
func manifestBody(n int) string {
	var b strings.Builder
	b.WriteString(`{"verdict_unit":"album","container":"files","source":"cd","files":[`)
	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"path":"disc 1/%04d - a track with a fairly long name to fill the manifest out.flac","size":123456789}`, i)
	}
	b.WriteString(`]}`)
	return b.String()
}

// TestServeHTTPSubmissionBodyCaps proves create_submission takes a
// 500-file manifest, which no 4 KiB body holds, and refuses a body one
// byte over 8 MiB, and that every other submission operation keeps the
// 4 KiB cap.
func TestServeHTTPSubmissionBodyCaps(t *testing.T) {
	manifest := manifestBody(500)
	if len(manifest) <= maxBody || len(manifest) > 8<<20 {
		t.Fatalf("fixture manifest is %d bytes, want over %d and under 8 MiB", len(manifest), maxBody)
	}
	tests := []struct {
		name string
		op   string
		body string
		want int
	}{
		{"500 files on create_submission", ops.OpCreateSubmission, manifest, http.StatusOK},
		{"a body exactly 8 MiB on create_submission", ops.OpCreateSubmission, paddedObject(8 << 20), http.StatusOK},
		{"a body one byte over 8 MiB on create_submission", ops.OpCreateSubmission, paddedObject(8<<20 + 1), http.StatusBadRequest},
		{"a body one byte over 4 KiB on get_submission", ops.OpGetSubmission, paddedObject(maxBody + 1), http.StatusBadRequest},
		{"a body one byte over 4 KiB on retry_submission", ops.OpRetrySubmission, paddedObject(maxBody + 1), http.StatusBadRequest},
		{"a body one byte over 4 KiB on cancel_submission", ops.OpCancelSubmission, paddedObject(maxBody + 1), http.StatusBadRequest},
		{"a body one byte over 4 KiB on list_submissions", ops.OpListSubmissions, paddedObject(maxBody + 1), http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, grant(ann, ops.PrivilegeMember), successes())
			rec := f.post(tt.op, tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body %.200q", rec.Code, tt.want, rec.Body)
			}
			if tt.want == http.StatusBadRequest {
				if code, _, _ := errorAnswer(t, rec); code != ops.CodeInvalidRequest {
					t.Errorf("code = %q, want invalid_request", code)
				}
			}
		})
	}
}

// TestServeHTTPRejectsEnvelope proves a refused create carries
// error.rejects beside the code and detail, each reason as its four wire
// strings, and that an error with no rejects writes no rejects key.
func TestServeHTTPRejectsEnvelope(t *testing.T) {
	refusal := ops.NewInvalidRequestError("the manifest was refused").WithRejects(ops.Rejects{
		Submission: []reject.Reason{reject.Of(reject.NoAudio)},
		Files:      []ops.FileReject{{Path: "a/../b.flac", Reason: reject.Of(reject.UnsafePath)}},
	})
	f := newFixture(t, grant(ann, ops.PrivilegeMember), successes())
	f.subs.answers[ops.OpCreateSubmission] = fakeAnswer{err: refusal}
	rec := f.post(ops.OpCreateSubmission, `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body %q", rec.Code, rec.Body)
	}
	var got struct {
		Error struct {
			Code    string `json:"code"`
			Rejects struct {
				Submission []map[string]string `json:"submission"`
				Files      []struct {
					Path   string            `json:"path"`
					Reason map[string]string `json:"reason"`
				} `json:"files"`
			} `json:"rejects"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body, err)
	}
	noAudio := reject.Of(reject.NoAudio)
	wantSubmission := map[string]string{"code": "no-audio", "label": noAudio.Label, "friendly": noAudio.Friendly, "next_step": noAudio.NextStep}
	if got.Error.Code != "invalid_request" || len(got.Error.Rejects.Submission) != 1 || !reflect.DeepEqual(got.Error.Rejects.Submission[0], wantSubmission) {
		t.Errorf("rejects.submission = %+v, want [%+v]", got.Error.Rejects.Submission, wantSubmission)
	}
	if len(got.Error.Rejects.Files) != 1 || got.Error.Rejects.Files[0].Path != "a/../b.flac" || got.Error.Rejects.Files[0].Reason["code"] != "unsafe-path" {
		t.Errorf("rejects.files = %+v, want one unsafe-path for a/../b.flac", got.Error.Rejects.Files)
	}

	f.subs.answers[ops.OpGetSubmission] = fakeAnswer{err: ops.NewNotFoundError("no such submission")}
	plain := f.post(ops.OpGetSubmission, `{"submission_id":"`+testHandle+`"}`)
	if strings.Contains(plain.Body.String(), "rejects") {
		t.Errorf("an error with no rejects wrote %s", plain.Body)
	}
}
