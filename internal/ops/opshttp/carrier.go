package opshttp

import (
	"context"
	"errors"

	"github.com/glw907/dubplate/internal/ops"
)

// admittedKey is the context key an admitted caller rides under. It is
// an unexported type of this package, so no other package can read or
// forge the value.
type admittedKey struct{}

// An admitted is the caller a mount's verifier admitted: the actor
// the call proceeds as and the privilege that actor holds.
type admitted struct {
	actor     ops.Actor
	privilege ops.Privilege
}

// withAdmitted returns ctx carrying the caller the mount admitted. The
// mount calls it after its verifier answers, and only then.
func withAdmitted(ctx context.Context, actor ops.Actor, privilege ops.Privilege) context.Context {
	return context.WithValue(ctx, admittedKey{}, admitted{actor: actor, privilege: privilege})
}

// errNoCaller denies a request whose context carries no admitted
// caller.
var errNoCaller = errors.New("request carries no admitted caller")

// A contextAuthorizer is the one reader of [withAdmitted]'s value: it
// answers the actor and privilege the mount put on the context, and
// denies a context that carries none.
type contextAuthorizer struct{}

func (contextAuthorizer) authorize(ctx context.Context, _ string) (ops.Actor, ops.Privilege, error) {
	c, ok := ctx.Value(admittedKey{}).(admitted)
	if !ok {
		return ops.Actor{}, 0, errNoCaller
	}
	return c.actor, c.privilege, nil
}
