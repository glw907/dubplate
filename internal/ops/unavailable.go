package ops

import "context"

// UnavailableJobs is a [Seam] for a server that runs no job runner. It
// answers each job operation with [ErrUnavailable] and the reason it
// was built with, and reports the same in its catalog, so the catalog
// operation still works and tells a caller why the others do not.
type UnavailableJobs struct {
	reason string
}

var _ Seam = UnavailableJobs{}

// NewUnavailableJobs builds an [UnavailableJobs] that gives reason as
// every refusal's cause. It panics on an empty reason, as
// [NewUnavailableError] does.
func NewUnavailableJobs(reason string) UnavailableJobs {
	if reason == "" {
		panic("ops: NewUnavailableJobs called with an empty reason")
	}
	return UnavailableJobs{reason: reason}
}

// StartBackup answers [ErrUnavailable].
func (u UnavailableJobs) StartBackup(context.Context, Actor, StartBackupRequest) (StartBackupResponse, error) {
	return StartBackupResponse{}, NewUnavailableError(u.reason)
}

// JobStatus answers [ErrUnavailable].
func (u UnavailableJobs) JobStatus(context.Context, Actor, JobStatusRequest) (JobStatusResponse, error) {
	return JobStatusResponse{}, NewUnavailableError(u.reason)
}

// CancelJob answers [ErrUnavailable].
func (u UnavailableJobs) CancelJob(context.Context, Actor, CancelJobRequest) (CancelJobResponse, error) {
	return CancelJobResponse{}, NewUnavailableError(u.reason)
}

// Catalog lists the four job-seam operations: the three job operations
// unavailable for the reason, and the catalog itself available.
func (u UnavailableJobs) Catalog(context.Context, Actor, CatalogRequest) (CatalogResponse, error) {
	entries := []CatalogEntry{
		{Name: OpStartBackup, Reason: u.reason},
		{Name: OpJobStatus, Reason: u.reason},
		{Name: OpCancelJob, Reason: u.reason},
		{Name: OpCatalog, Available: true},
	}
	for i, e := range entries {
		entries[i].Privilege, _ = PrivilegeOf(e.Name)
	}
	cat, err := BuildCatalog(entries...)
	if err != nil {
		return CatalogResponse{}, err
	}
	return CatalogResponse{Operations: cat.Entries()}, nil
}
