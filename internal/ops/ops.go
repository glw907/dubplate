package ops

import (
	"time"

	"github.com/glw907/dubplate/internal/id"
)

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

// A JobStatus is a snapshot of one job's record. It carries no wire
// tags: no operation serves a job, so the type has no wire shape to
// hold.
type JobStatus struct {
	JobID          id.ID
	Kind           string
	State          JobState
	Phase          string
	FailureCode    string
	FailureMessage string
	QueuedAt       time.Time
	StartedAt      *time.Time
	UpdatedAt      time.Time
	FinishedAt     *time.Time
}
