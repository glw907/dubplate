package jobs

import (
	"errors"
	"testing"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/ops"
)

// TestIDPinJobStatus pins what JobStatus answers for the all-zeros id,
// which names no job and answers not found. A malformed id never
// reaches the runner: the JSON decoder refuses it upstream.
func TestIDPinJobStatus(t *testing.T) {
	h := openWrite(t, newStateDir(t))
	r := newRunnerFor(t, h, nil, nil)

	const zeros = "00000000000000000000000000000000"
	_, err := r.JobStatus(deadline(t), testActor, id.ID{})
	e, ok := errors.AsType[*ops.Error](err)
	if !ok {
		t.Fatalf("JobStatus(%q) = %v, want an *ops.Error", zeros, err)
	}
	if want := "job " + zeros; e.Code != ops.CodeNotFound || e.Reason != want {
		t.Errorf("JobStatus(%q) = code %q, reason %q; want code %q, reason %q", zeros, e.Code, e.Reason, ops.CodeNotFound, want)
	}
}
