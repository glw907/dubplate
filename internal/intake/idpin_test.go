package intake

import (
	"net/http"
	"testing"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/ops"
)

const (
	pinZeros        = "00000000000000000000000000000000"
	pinLowerHex     = "0123456789abcdef0123456789abcdef"
	pinUpperHex     = "0123456789ABCDEF0123456789ABCDEF"
	pinShortHex     = "0123456789abcdef0123456789abcde"
	pinNotHex       = "0123456789abcdef0123456789abcdeg"
	pinNoSubmission = "submission_id is missing or names no submission"
)

// TestIDPinSubmissionID pins the wire id as the seam refuses it. The
// JSON decoder upstream refuses a malformed id, so the seam sees only
// a typed id: the all-zeros id answers one invalid-request sentence,
// and a well-formed id that names nothing answers not found.
func TestIDPinSubmissionID(t *testing.T) {
	h, seam := newTestService(t)
	m := member(t, h)
	ctx := deadline(t)
	unknown, err := id.Parse(pinLowerHex)
	if err != nil {
		t.Fatalf("id.Parse(%q): %v", pinLowerHex, err)
	}
	tests := []struct {
		name     string
		id       id.ID
		wantCode ops.Code
		wantMsg  string
	}{
		{"all zeros", id.ID{}, ops.CodeInvalidRequest, pinNoSubmission},
		{"well formed, unknown", unknown, ops.CodeNotFound, "no such submission"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := seam.GetSubmission(ctx, actorOf(m), ops.GetSubmissionRequest{SubmissionID: tt.id})
			e := opsError(t, err)
			if e.Code != tt.wantCode || e.Reason != tt.wantMsg {
				t.Errorf("GetSubmission(%q) = code %q, reason %q; want code %q, reason %q", tt.id, e.Code, e.Reason, tt.wantCode, tt.wantMsg)
			}
		})
	}
}

// TestIDPinTusPath pins the wrapper's answers to an upload id in the
// path: a malformed id and the all-zeros id each answer the wrapper's
// own 404, written before tusd sees the request.
func TestIDPinTusPath(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	for _, tt := range []struct{ name, id string }{
		{"31 characters", pinShortHex},
		{"uppercase", pinUpperHex},
		{"not hex", pinNotHex},
		{"all zeros", pinZeros},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := patchRequest(m, nil, 0, []byte("zz"))
			req.path = tusPrefix + "/tus/" + tt.id
			reply := f.do(t, req)
			assertRefusal(t, "PATCH "+tt.id, reply, http.StatusNotFound, "ERR_UPLOAD_NOT_FOUND")
			if !refusedBeforeTusd(reply) {
				t.Errorf("PATCH %s: tusd answered, want the wrapper's own 404", tt.id)
			}
		})
	}
}

// TestIDPinTusCreate pins a creation whose metadata names a malformed
// or all-zeros submission or file id: the 400 invalid_request a creation
// naming no file answers.
func TestIDPinTusCreate(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	for _, tt := range []struct{ name, submission, file string }{
		{"short submission", pinShortHex, pinLowerHex},
		{"uppercase file", pinLowerHex, pinUpperHex},
		{"zero submission", pinZeros, pinLowerHex},
		{"zero file", pinLowerHex, pinZeros},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reply := f.do(t, tusRequest{
				method: http.MethodPost, path: tusPrefix + "/tus/", member: &m,
				header: map[string]string{
					"Tus-Resumable":   "1.0.0",
					"Upload-Length":   "4",
					"Upload-Metadata": metadata("submission", tt.submission, "file", tt.file),
				},
			})
			assertRefusal(t, "create", reply, http.StatusBadRequest, "invalid_request")
		})
	}
}
