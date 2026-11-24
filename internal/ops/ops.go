package ops

import (
	"fmt"
	"time"

	"github.com/glw907/dubplate/internal/id"
)

// handleLen is a job [Handle]'s wire length: 32 lowercase hex
// characters, the hex encoding of 16 random bytes. internal/ops
// imports no store type, so the length is restated here rather than
// referenced.
const handleLen = 32

// A Handle identifies a job: an opaque value rendered as 32 lowercase
// hex characters on the wire, minted by the runner that starts the
// job and never accepted on a start request.
type Handle string

// ParseHandle parses s as a [Handle]. An empty, wrong-length,
// uppercase, or non-hex s is refused with an [*Error] wrapping
// [ErrInvalidRequest] and naming s.
func ParseHandle(s string) (Handle, error) {
	if _, err := id.Parse(s); err != nil {
		return "", NewInvalidRequestError(fmt.Sprintf("job handle %q: want %d lowercase hex characters", s, handleLen))
	}
	return Handle(s), nil
}

// A JobState is one of a job's six lifecycle states, unchanged once it
// reaches a terminal one.
type JobState string

const (
	JobQueued      JobState = "queued"
	JobRunning     JobState = "running"
	JobSucceeded   JobState = "succeeded"
	JobFailed      JobState = "failed"
	JobCanceled    JobState = "canceled"
	JobInterrupted JobState = "interrupted"
)

// Terminal reports whether s never changes again.
func (s JobState) Terminal() bool {
	switch s {
	case JobSucceeded, JobFailed, JobCanceled, JobInterrupted:
		return true
	default:
		return false
	}
}

// A JobStatus is a snapshot of one job's record.
type JobStatus struct {
	ID             Handle     `json:"id"`
	Kind           string     `json:"kind"`
	State          JobState   `json:"state"`
	Phase          string     `json:"phase,omitempty"`
	FailureCode    string     `json:"failure_code,omitempty"`
	FailureMessage string     `json:"failure_message,omitempty"`
	QueuedAt       time.Time  `json:"queued_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

// A StartBackupRequest starts a backup job. It carries no fields:
// backup is the only kind this package declares, and the acting
// identity is the actor parameter of the call, never a request field.
type StartBackupRequest struct{}

// A StartBackupResponse names the job a start_backup call started.
type StartBackupResponse struct {
	JobID Handle `json:"job_id"`
}

// A JobStatusRequest names the job a job_status call reads.
type JobStatusRequest struct {
	JobID Handle `json:"job_id"`
}

// A JobStatusResponse carries the job a job_status call read.
type JobStatusResponse struct {
	Job JobStatus `json:"job"`
}

// A CancelJobRequest names the job a cancel_job call cancels.
type CancelJobRequest struct {
	JobID Handle `json:"job_id"`
}

// A CancelJobResponse carries the job's state after a cancel_job call
// asked for cancellation.
type CancelJobResponse struct {
	Job JobStatus `json:"job"`
}
