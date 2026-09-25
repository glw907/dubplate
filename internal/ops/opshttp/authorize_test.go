package opshttp

import (
	"context"
	"errors"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// TestAuthorizingSeamDiscardsItsCallersActor plants a command-line
// actor on every call; the seam beneath must see only the actor the
// authorizer admitted.
func TestAuthorizingSeamDiscardsItsCallersActor(t *testing.T) {
	planted := ops.LocalActor("root")
	ctx := context.Background()
	tests := []struct {
		op   string
		call func(s *authorizingSeam) error
	}{
		{ops.OpStartBackup, func(s *authorizingSeam) error {
			_, err := s.StartBackup(ctx, planted, ops.StartBackupRequest{})
			return err
		}},
		{ops.OpJobStatus, func(s *authorizingSeam) error {
			_, err := s.JobStatus(ctx, planted, ops.JobStatusRequest{JobID: testHandle})
			return err
		}},
		{ops.OpCancelJob, func(s *authorizingSeam) error {
			_, err := s.CancelJob(ctx, planted, ops.CancelJobRequest{JobID: testHandle})
			return err
		}},
		{ops.OpCatalog, func(s *authorizingSeam) error {
			_, err := s.Catalog(ctx, planted, ops.CatalogRequest{})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			seam := &fakeSeam{answers: successes()}
			s, err := newAuthorizingSeam(seam, admitAnn(), identity.ChannelCookie)
			if err != nil {
				t.Fatalf("newAuthorizingSeam: %v", err)
			}
			if err := tt.call(s); err != nil {
				t.Fatalf("%s: %v", tt.op, err)
			}
			if len(seam.calls) != 1 {
				t.Fatalf("seam calls = %d, want 1", len(seam.calls))
			}
			if got := seam.calls[0]; got.op != tt.op || got.actor.String() != annWire {
				t.Errorf("seam got %s as %q, want %s as %q", got.op, got.actor, tt.op, annWire)
			}
		})
	}
}

// TestAuthorizingSeamDenialIsForbidden checks the in-process shape of a
// denial: the error matches ops.ErrForbidden and still carries the
// authorizer's own error for the log.
func TestAuthorizingSeamDenialIsForbidden(t *testing.T) {
	errSession := errors.New("session expired")
	seam := &fakeSeam{answers: successes()}
	deny := &fakeAuthorizer{decide: func(string, ops.Privilege) (ops.Actor, error) {
		return ops.Actor{}, errSession
	}}
	s, err := newAuthorizingSeam(seam, deny, identity.ChannelCookie)
	if err != nil {
		t.Fatalf("newAuthorizingSeam: %v", err)
	}
	_, err = s.CancelJob(context.Background(), ops.LocalActor("root"), ops.CancelJobRequest{JobID: testHandle})
	if !errors.Is(err, ops.ErrForbidden) || !errors.Is(err, errSession) {
		t.Errorf("CancelJob = %v, want ops.ErrForbidden wrapping the authorizer's error", err)
	}
	if code := ops.CodeOf(err); code != ops.CodeForbidden {
		t.Errorf("CodeOf = %q, want forbidden", code)
	}
	if len(seam.calls) != 0 {
		t.Errorf("seam calls = %d, want 0", len(seam.calls))
	}
}
