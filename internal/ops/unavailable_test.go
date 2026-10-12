package ops

import (
	"context"
	"errors"
	"slices"
	"testing"
)

const noRunner = "this server runs no job runner"

const someHandle = "0123456789abcdef0123456789abcdef"

func TestUnavailableJobs_RefusesEveryJobOperation(t *testing.T) {
	jobs := NewUnavailableJobs(noRunner)
	ctx := context.Background()
	tests := []struct {
		name string
		call func() error
	}{
		{OpStartBackup, func() error {
			_, err := jobs.StartBackup(ctx, Actor{}, StartBackupRequest{})
			return err
		}},
		{OpJobStatus, func() error {
			_, err := jobs.JobStatus(ctx, Actor{}, JobStatusRequest{JobID: someHandle})
			return err
		}},
		{OpCancelJob, func() error {
			_, err := jobs.CancelJob(ctx, Actor{}, CancelJobRequest{JobID: someHandle})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if !errors.Is(err, ErrUnavailable) {
				t.Fatalf("err = %v, want ErrUnavailable", err)
			}
			if e, ok := errors.AsType[*Error](err); !ok || e.Code != CodeUnavailable || e.Reason != noRunner {
				t.Errorf("err = %+v, want code unavailable with reason %q", e, noRunner)
			}
		})
	}
}

// TestUnavailableJobs_Catalog proves the three job operations are
// reported unavailable with the reason and the catalog operation is
// not, each with its privilege from the one table.
func TestUnavailableJobs_Catalog(t *testing.T) {
	resp, err := NewUnavailableJobs(noRunner).Catalog(context.Background(), Actor{}, CatalogRequest{})
	if err != nil {
		t.Fatalf("Catalog() = %v", err)
	}
	want := []CatalogEntry{
		{Name: "start_backup", Privilege: PrivilegeAdmin, Available: false, Reason: noRunner},
		{Name: "job_status", Privilege: PrivilegeAdmin, Available: false, Reason: noRunner},
		{Name: "cancel_job", Privilege: PrivilegeAdmin, Available: false, Reason: noRunner},
		{Name: "catalog", Privilege: PrivilegeMember, Available: true},
	}
	if !slices.Equal(resp.Operations, want) {
		t.Errorf("Catalog() = %+v, want %+v", resp.Operations, want)
	}
}

func TestNewUnavailableJobs_EmptyReasonPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewUnavailableJobs(\"\") did not panic")
		}
	}()
	NewUnavailableJobs("")
}
