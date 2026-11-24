package jobs

import (
	"errors"
	"testing"

	"github.com/glw907/dubplate/internal/ops"
)

// TestIDPinJobStatus pins what JobStatus answers for a malformed handle,
// the ParseHandle sentence, and for the all-zeros handle, which names no
// job and answers not found.
func TestIDPinJobStatus(t *testing.T) {
	h := openWrite(t, newStateDir(t))
	r := newRunnerFor(t, h, nil, nil)
	ctx := deadline(t)

	const zeros = "00000000000000000000000000000000"
	tests := []struct {
		name     string
		handle   string
		wantCode ops.Code
		wantMsg  string
	}{
		{"31 characters", "0123456789abcdef0123456789abcde", ops.CodeInvalidRequest, `job handle "0123456789abcdef0123456789abcde": want 32 lowercase hex characters`},
		{"uppercase", "0123456789ABCDEF0123456789ABCDEF", ops.CodeInvalidRequest, `job handle "0123456789ABCDEF0123456789ABCDEF": want 32 lowercase hex characters`},
		{"not hex", "0123456789abcdef0123456789abcdeg", ops.CodeInvalidRequest, `job handle "0123456789abcdef0123456789abcdeg": want 32 lowercase hex characters`},
		{"all zeros", zeros, ops.CodeNotFound, "job " + zeros},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.JobStatus(ctx, testActor, ops.JobStatusRequest{JobID: ops.Handle(tt.handle)})
			e, ok := errors.AsType[*ops.Error](err)
			if !ok {
				t.Fatalf("JobStatus(%q) = %v, want an *ops.Error", tt.handle, err)
			}
			if e.Code != tt.wantCode || e.Reason != tt.wantMsg {
				t.Errorf("JobStatus(%q) = code %q, reason %q; want code %q, reason %q", tt.handle, e.Code, e.Reason, tt.wantCode, tt.wantMsg)
			}
		})
	}
}
