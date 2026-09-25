package opshttp

import (
	"context"
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// An authorizer names the caller ctx carries for a call to op: the
// actor the call proceeds as and the privilege that caller holds. The
// seam, never the authorizer, decides whether that privilege reaches
// the one op requires. Any error denies; it must carry no identity
// text, no address or principal, because it stays in the chain of the
// denial the seam returns. It receives no request, so a verified
// identity reaches it only through ctx.
type authorizer interface {
	authorize(ctx context.Context, op string) (ops.Actor, ops.Privilege, error)
}

// An authorizingSeam is an [ops.Seam] that admits a call to next only
// when its authorizer answers a caller whose privilege reaches the
// operation's, and then only as the actor the authorizer answered. It discards the actor its own caller passes, so no caller
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

// The causes of a denial. Each is fixed text naming no caller, so a
// log may name the cause a denial wraps.
var (
	errAuthorizerDenied = errors.New("authorizer denied the caller")
	errChannelMismatch  = errors.New("admitted actor's channel differs from the mount's")
	errPrivilegeShort   = errors.New("granted privilege does not reach the operation's")
)

// denials lists the causes [denialReason] names, in the order it
// checks them.
var denials = []error{errAuthorizerDenied, errChannelMismatch, errPrivilegeShort}

// denialReason names a forbidden err by the first cause it wraps from
// denials, or "denied" when it wraps none, as a forbidden answer from
// beneath the seam does. It never returns err's own text, which may
// quote an authorizer's error.
func denialReason(err error) string {
	for _, cause := range denials {
		if errors.Is(err, cause) {
			return cause.Error()
		}
	}
	return "denied"
}

// reaches reports whether granted satisfies required: admin satisfies
// either privilege, member satisfies member only, and any other
// granted value, the zero value included, satisfies nothing.
func reaches(granted, required ops.Privilege) bool {
	switch granted {
	case ops.PrivilegeAdmin:
		return required == ops.PrivilegeAdmin || required == ops.PrivilegeMember
	case ops.PrivilegeMember:
		return required == ops.PrivilegeMember
	}
	return false
}

// admit returns the actor a call to op proceeds as. A denial wraps the
// forbidden answer first, so [ops.CodeOf] and a reason read through
// [errors.AsType] see it ahead of anything the authorizer returned,
// then its cause from denials.
func (s *authorizingSeam) admit(ctx context.Context, op string) (ops.Actor, error) {
	p, ok := ops.PrivilegeOf(op)
	if !ok {
		return ops.Actor{}, ops.NewNotFoundError(fmt.Sprintf("no operation %q", op))
	}
	denied := ops.NewForbiddenError(fmt.Sprintf("%s requires the %s privilege", op, p))
	actor, granted, err := s.auth.authorize(ctx, op)
	if err != nil {
		return ops.Actor{}, fmt.Errorf("%w: %w: %w", denied, errAuthorizerDenied, err)
	}
	if actor.Channel() != s.channel {
		return ops.Actor{}, fmt.Errorf("%w: %w", denied, errChannelMismatch)
	}
	if !reaches(granted, p) {
		return ops.Actor{}, fmt.Errorf("%w: %w", denied, errPrivilegeShort)
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
