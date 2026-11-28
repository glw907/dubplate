package ops

import (
	"context"
	"time"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/reject"
)

// The submission operations.
const (
	OpCreateSubmission Op = "create_submission"
	OpListSubmissions  Op = "list_submissions"
	OpGetSubmission    Op = "get_submission"
	OpRetrySubmission  Op = "retry_submission"
	OpCancelSubmission Op = "cancel_submission"
)

// A ManifestFile is one file a member declares when creating a
// submission.
type ManifestFile struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// A CreateSubmissionRequest declares a submission before any byte is
// sent. It carries no uploader, collection, state, reservation, or
// role: the server derives each, so a body naming one decodes to this
// request without it.
type CreateSubmissionRequest struct {
	// VerdictUnit is "album" or "track".
	VerdictUnit string `json:"verdict_unit"`
	// Container is "files" or "zip".
	Container string         `json:"container"`
	Files     []ManifestFile `json:"files"`
	// Source is where the music came from, one of the answers the gate
	// accepts.
	Source     string `json:"source"`
	SourceNote string `json:"source_note,omitempty"`
	StoreURL   string `json:"store_url,omitempty"`
}

// A CreatedFile is the id a declared file uploads under.
type CreatedFile struct {
	Path   string `json:"path"`
	FileID id.ID  `json:"file_id"`
}

// A CreateSubmissionResponse names the submission created and the id
// each declared file uploads under.
type CreateSubmissionResponse struct {
	SubmissionID id.ID         `json:"submission_id"`
	Files        []CreatedFile `json:"files"`
}

// A ListSubmissionsRequest lists the caller's submissions. It carries
// no fields.
type ListSubmissionsRequest struct{}

// A SubmissionSummary is one row of the caller's list.
type SubmissionSummary struct {
	SubmissionID id.ID     `json:"submission_id"`
	State        string    `json:"state"`
	VerdictUnit  string    `json:"verdict_unit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// A ListSubmissionsResponse carries the caller's submissions, newest
// first.
type ListSubmissionsResponse struct {
	Submissions []SubmissionSummary `json:"submissions"`
}

// A GetSubmissionRequest names the submission
// [Submissions.GetSubmission] reads.
type GetSubmissionRequest struct {
	SubmissionID id.ID `json:"submission_id"`
}

// A SubmissionEvent is one entry of a submission's journey. Reason is
// nil for an event that carries no reject code.
type SubmissionEvent struct {
	At     time.Time      `json:"at"`
	Kind   string         `json:"kind"`
	Reason *reject.Reason `json:"reason,omitempty"`
}

// A SubmissionFile is one file's progress and the reason it was
// refused. Reason is nil for a file that passed or has no answer yet.
type SubmissionFile struct {
	FileID   id.ID          `json:"file_id"`
	Path     string         `json:"path"`
	Size     int64          `json:"size"`
	Received bool           `json:"received"`
	Reason   *reject.Reason `json:"reason"`
}

// A SubmissionWarning is a reason that never blocks. Subject is the
// member name an extras warning names, and nil for a completeness one.
type SubmissionWarning struct {
	Reason  reject.Reason `json:"reason"`
	Subject *string       `json:"subject"`
}

// A GetSubmissionResponse carries one of the caller's submissions with
// its journey, its files, and its warnings. A path or name that is not
// valid UTF-8 is rendered with the replacement character.
type GetSubmissionResponse struct {
	SubmissionID id.ID               `json:"submission_id"`
	State        string              `json:"state"`
	VerdictUnit  string              `json:"verdict_unit"`
	Container    string              `json:"container"`
	Source       string              `json:"source"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Events       []SubmissionEvent   `json:"events"`
	Files        []SubmissionFile    `json:"files"`
	Warnings     []SubmissionWarning `json:"warnings"`
}

// A RetrySubmissionRequest names the submission
// [Submissions.RetrySubmission] sends back to its check.
type RetrySubmissionRequest GetSubmissionRequest

// A RetrySubmissionResponse carries the submission after the retry.
type RetrySubmissionResponse GetSubmissionResponse

// A CancelSubmissionRequest names the submission
// [Submissions.CancelSubmission] ends.
type CancelSubmissionRequest GetSubmissionRequest

// A CancelSubmissionResponse carries the submission after the cancel.
type CancelSubmissionResponse GetSubmissionResponse

// A Submissions is what a caller reaches the submission operations
// through, beside [Admin]. Each is a member's own: the caller is the
// actor's UserID, and an id that is another member's answers as an
// unknown one, [CodeNotFound].
// An action the submission's state does not allow answers [CodeConflict].
type Submissions interface {
	// CreateSubmission checks the manifest and the member's limits and
	// reserves the space. A refusal carries [Error.Rejects].
	CreateSubmission(ctx context.Context, actor identity.Actor, req CreateSubmissionRequest) (CreateSubmissionResponse, error)

	// ListSubmissions lists the actor's submissions, newest first.
	ListSubmissions(ctx context.Context, actor identity.Actor, req ListSubmissionsRequest) (ListSubmissionsResponse, error)

	// GetSubmission reads one of the actor's submissions.
	GetSubmission(ctx context.Context, actor identity.Actor, req GetSubmissionRequest) (GetSubmissionResponse, error)

	// RetrySubmission moves a submission whose check could not run back
	// to checking.
	RetrySubmission(ctx context.Context, actor identity.Actor, req RetrySubmissionRequest) (RetrySubmissionResponse, error)

	// CancelSubmission ends a submission that is uploading or whose
	// check could not run.
	CancelSubmission(ctx context.Context, actor identity.Actor, req CancelSubmissionRequest) (CancelSubmissionResponse, error)
}
