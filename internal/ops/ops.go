package ops

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// localActorPrefix frames a local [Actor]'s wire form.
const localActorPrefix = "local"

// An Actor identifies who invoked an operation, verified before the
// call reaches [Seam]: never a field a request type carries, so no
// request body can widen its own privilege (R-6bi4). 6b-i builds the
// local kind only, from an operating-system username; 6b-ii adds the
// kind a verified network identity asserts.
type Actor struct {
	wire string
}

// LocalActor builds the local kind of [Actor] from username, the
// operating-system user a CLI verb runs as (the shape
// internal/engine/store's ownerName already produces, which never
// returns an empty string). It renders as "local:<username>" on the
// wire and in every log line.
func LocalActor(username string) Actor {
	if username == "" {
		panic("ops: LocalActor called with an empty username")
	}
	return Actor{wire: localActorPrefix + ":" + username}
}

// String returns a's wire form.
func (a Actor) String() string { return a.wire }

// handleLen is a job [Handle]'s wire length: 32 lowercase hex
// characters, the hex encoding of 16 random bytes. internal/ops
// imports no store type (Global constraints), so the length is
// restated here rather than referenced.
const handleLen = 32

// A Handle identifies a job: an opaque value rendered as 32 lowercase
// hex characters on the wire, minted by the runner that starts the
// job and never accepted on a start request.
type Handle string

// ParseHandle parses s as a [Handle]. An empty, wrong-length,
// uppercase, or non-hex s is refused with an [*Error] wrapping
// [ErrInvalidRequest] and naming s.
func ParseHandle(s string) (Handle, error) {
	if len(s) != handleLen || s != strings.ToLower(s) {
		return "", NewInvalidRequestError(fmt.Sprintf("job handle %q: want %d lowercase hex characters", s, handleLen))
	}
	if _, err := hex.DecodeString(s); err != nil {
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
// backup is 6b-i's only kind, and the acting identity is
// [Seam.StartBackup]'s actor parameter, never a request field (spec
// criterion 2).
type StartBackupRequest struct{}

// A StartBackupResponse names the job [Seam.StartBackup] started.
type StartBackupResponse struct {
	JobID Handle `json:"job_id"`
}

// A JobStatusRequest names the job [Seam.JobStatus] reads.
type JobStatusRequest struct {
	JobID Handle `json:"job_id"`
}

// A JobStatusResponse carries the job [Seam.JobStatus] read.
type JobStatusResponse struct {
	Job JobStatus `json:"job"`
}

// A CancelJobRequest names the job [Seam.CancelJob] cancels.
type CancelJobRequest struct {
	JobID Handle `json:"job_id"`
}

// A CancelJobResponse carries the job's state after [Seam.CancelJob]
// asked for cancellation.
type CancelJobResponse struct {
	Job JobStatus `json:"job"`
}

// A CatalogRequest lists every operation [Seam.Catalog] exposes. It
// carries no fields.
type CatalogRequest struct{}

// A CatalogResponse carries the catalog [Seam.Catalog] answered with.
type CatalogResponse struct {
	Operations []CatalogEntry `json:"operations"`
}

// A Seam is what a caller reaches R-6bi5's four operations through:
// the vocabulary an HTTP projection and a CLI verb alike bind to a
// runner. Every method takes the acting identity as its own parameter
// (R-6bi4), never as a request field, and returns an error wrapping
// one of this package's sentinels when it cannot answer.
type Seam interface {
	// StartBackup starts a backup job and returns its handle. It
	// returns an error wrapping [ErrUnavailable] when backup
	// integration is off, or [ErrConflict] naming the job already
	// running.
	StartBackup(ctx context.Context, actor Actor, req StartBackupRequest) (StartBackupResponse, error)

	// JobStatus reads a job's current record. It returns an error
	// wrapping [ErrNotFound] when no record with that handle exists.
	JobStatus(ctx context.Context, actor Actor, req JobStatusRequest) (JobStatusResponse, error)

	// CancelJob requests a job's cancellation. Canceling a queued job
	// ends it canceled at once; canceling a running job asks its
	// owning runner to stop; canceling a finished job reports its
	// state unchanged.
	CancelJob(ctx context.Context, actor Actor, req CancelJobRequest) (CancelJobResponse, error)

	// Catalog lists every operation this seam exposes, with its
	// availability and privilege.
	Catalog(ctx context.Context, actor Actor, req CatalogRequest) (CatalogResponse, error)
}
