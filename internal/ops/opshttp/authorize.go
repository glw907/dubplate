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
// [admission], never the authorizer, decides whether that privilege
// reaches the one op requires. Any error denies; it must carry no
// identity text, no address or principal, because it stays in the
// chain of the denial the admission returns. It receives no request, so a verified
// identity reaches it only through ctx.
type authorizer interface {
	authorize(ctx context.Context, op string) (identity.Actor, ops.Privilege, error)
}

// An admission decides who may call an operation. [admission.admit]
// answers the actor a call proceeds as only when the authorizer answers
// a caller whose privilege reaches the operation's, so no caller can
// choose the actor a submission or an audit row records. An admitted actor
// whose channel differs from the mount's is refused, so an authorizer
// wired to the wrong mount fails closed.
type admission struct {
	auth    authorizer
	channel identity.Channel
}

// newAdmission builds the admission gate for a mount whose
// verified identities arrive on channel. It fails on a nil auth, and on
// a channel that is not valid or that names the command line, which no
// network mount can carry.
func newAdmission(auth authorizer, channel identity.Channel) (*admission, error) {
	switch {
	case auth == nil:
		return nil, errors.New("authorizer is nil")
	case !channel.Valid():
		return nil, fmt.Errorf("mount channel %d is not a recognized channel", int(channel))
	case channel == identity.ChannelCommandLine:
		return nil, errors.New("mount channel is the command line, which no network mount carries")
	}
	return &admission{auth: auth, channel: channel}, nil
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
// an operation itself does. It never returns err's own text, which may
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
func (s *admission) admit(ctx context.Context, op string) (identity.Actor, error) {
	p, ok := ops.PrivilegeOf(op)
	if !ok {
		return identity.Actor{}, ops.NewError(ops.CodeNotFound, fmt.Sprintf("no operation %q", op))
	}
	denied := ops.NewError(ops.CodeForbidden, fmt.Sprintf("%s requires the %s privilege", op, p))
	actor, granted, err := s.auth.authorize(ctx, op)
	if err != nil {
		return identity.Actor{}, fmt.Errorf("%w: %w: %w", denied, errAuthorizerDenied, err)
	}
	if actor.Channel() != s.channel {
		return identity.Actor{}, fmt.Errorf("%w: %w", denied, errChannelMismatch)
	}
	if !reaches(granted, p) {
		return identity.Actor{}, fmt.Errorf("%w: %w", denied, errPrivilegeShort)
	}
	return actor, nil
}
