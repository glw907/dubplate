package opshttp

import (
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// The causes of a denial. Each is fixed text naming no caller, so a
// log may name the cause a denial wraps.
var (
	errChannelMismatch = errors.New("admitted actor's channel differs from the mount's")
	errPrivilegeShort  = errors.New("granted privilege does not reach the operation's")
)

// denials lists the causes [denialReason] names, in the order it
// checks them.
var denials = []error{errChannelMismatch, errPrivilegeShort}

// denialReason names a forbidden err by the first cause it wraps from
// denials, or "denied" when it wraps none, as a forbidden answer from
// an operation itself does. It never returns err's own text, which may
// quote an operation's reason.
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

// admit refuses a call to op, which requires required, unless actor
// arrived on the cookie channel and granted reaches required, so a
// verifier wired to the wrong mount fails closed. A denial wraps the
// forbidden answer first, so [ops.CodeOf] and a reason read through
// [errors.AsType] see it, then its cause from denials.
func admit(op ops.Op, required ops.Privilege, actor identity.Actor, granted ops.Privilege) error {
	denied := ops.NewError(ops.CodeForbidden, fmt.Sprintf("%s requires the %s privilege", op, required))
	if actor.Channel() != identity.ChannelCookie {
		return fmt.Errorf("%w: %w", denied, errChannelMismatch)
	}
	if !reaches(granted, required) {
		return fmt.Errorf("%w: %w", denied, errPrivilegeShort)
	}
	return nil
}
