package opshttp

import (
	"context"
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// An authorizer decides whether the caller ctx carries may invoke op,
// which requires p. A nil error admits the caller as the returned
// actor; any error denies. It receives no request, so a verified
// identity reaches it only through ctx.
type authorizer interface {
	authorize(ctx context.Context, op string, p ops.Privilege) (ops.Actor, error)
}

// An authorizingSeam is an [ops.Seam] that admits a call to next only
// when its authorizer does, and then only as the actor the authorizer
// answered. It discards the actor its own caller passes, so no caller
// can choose the actor a job records. An admitted actor whose channel
// differs from the mount's is refused, so an authorizer wired to the
// wrong mount fails closed.
type authorizingSeam struct {
	next    ops.Seam
	auth    authorizer
	channel identity.Channel
}

// newAuthorizingSeam wraps next. It fails on a nil next or auth, and
// on a channel that is not valid or that names the command line, which
// no network mount can carry.
func newAuthorizingSeam(next ops.Seam, auth authorizer, channel identity.Channel) (*authorizingSeam, error) {
	switch {
	case next == nil:
		return nil, errors.New("opshttp: seam is nil")
	case auth == nil:
		return nil, errors.New("opshttp: authorizer is nil")
	case !channel.Valid():
		return nil, fmt.Errorf("opshttp: mount channel %d is not a recognized channel", int(channel))
	case channel == identity.ChannelCommandLine:
		return nil, errors.New("opshttp: mount channel is the command line, which no network mount carries")
	}
	return &authorizingSeam{next: next, auth: auth, channel: channel}, nil
}

// admit returns the actor a call to op proceeds as. A denial wraps the
// forbidden answer first, so [ops.CodeOf] and a reason read through
// [errors.AsType] see it ahead of anything the authorizer returned.
func (s *authorizingSeam) admit(ctx context.Context, op string) (ops.Actor, error) {
	p, ok := ops.PrivilegeOf(op)
	if !ok {
		return ops.Actor{}, ops.NewNotFoundError(fmt.Sprintf("no operation %q", op))
	}
	denied := ops.NewForbiddenError(fmt.Sprintf("%s requires the %s privilege", op, p))
	actor, err := s.auth.authorize(ctx, op, p)
	if err != nil {
		return ops.Actor{}, fmt.Errorf("%w: %w", denied, err)
	}
	if actor.Channel() != s.channel {
		return ops.Actor{}, fmt.Errorf("%w: authorizer answered actor %q on channel %s for a %s mount",
			denied, actor, actor.Channel(), s.channel)
	}
	return actor, nil
}

// StartBackup implements [ops.Seam].
func (s *authorizingSeam) StartBackup(ctx context.Context, _ ops.Actor, req ops.StartBackupRequest) (ops.StartBackupResponse, error) {
	actor, err := s.admit(ctx, ops.OpStartBackup)
	if err != nil {
		return ops.StartBackupResponse{}, err
	}
	return s.next.StartBackup(ctx, actor, req)
}

// JobStatus implements [ops.Seam].
func (s *authorizingSeam) JobStatus(ctx context.Context, _ ops.Actor, req ops.JobStatusRequest) (ops.JobStatusResponse, error) {
	actor, err := s.admit(ctx, ops.OpJobStatus)
	if err != nil {
		return ops.JobStatusResponse{}, err
	}
	return s.next.JobStatus(ctx, actor, req)
}

// CancelJob implements [ops.Seam].
func (s *authorizingSeam) CancelJob(ctx context.Context, _ ops.Actor, req ops.CancelJobRequest) (ops.CancelJobResponse, error) {
	actor, err := s.admit(ctx, ops.OpCancelJob)
	if err != nil {
		return ops.CancelJobResponse{}, err
	}
	return s.next.CancelJob(ctx, actor, req)
}

// Catalog implements [ops.Seam].
func (s *authorizingSeam) Catalog(ctx context.Context, _ ops.Actor, req ops.CatalogRequest) (ops.CatalogResponse, error) {
	actor, err := s.admit(ctx, ops.OpCatalog)
	if err != nil {
		return ops.CatalogResponse{}, err
	}
	return s.next.Catalog(ctx, actor, req)
}
