package opshttp

import (
	"context"
	"testing"

	"github.com/glw907/dubplate/internal/ops"
)

// TestContextAuthorizerReadsTheCarriedCaller proves the authorizer
// answers exactly the actor and privilege the mount put on the
// context, whatever the operation.
func TestContextAuthorizerReadsTheCarriedCaller(t *testing.T) {
	for _, p := range []ops.Privilege{ops.PrivilegeMember, ops.PrivilegeAdmin} {
		t.Run(p.String(), func(t *testing.T) {
			want := ops.NetworkActor(ann, annAnchor)
			ctx := withAdmitted(context.Background(), want, p)
			got, granted, err := contextAuthorizer{}.authorize(ctx, ops.OpListMembers)
			if err != nil {
				t.Fatalf("authorize = %v", err)
			}
			if got.String() != annWire || !got.Anchor().Equal(annAnchor) || got.UserID() != ann.UserID {
				t.Errorf("actor = %q anchored %v, want %q anchored %v", got, got.Anchor(), annWire, annAnchor)
			}
			if granted != p {
				t.Errorf("granted = %v, want %v", granted, p)
			}
		})
	}
}

// TestContextAuthorizerDeniesWithoutACaller proves a request the mount
// never admitted, which carries no caller, is denied rather than
// admitted as a zero actor.
func TestContextAuthorizerDeniesWithoutACaller(t *testing.T) {
	actor, granted, err := contextAuthorizer{}.authorize(context.Background(), ops.OpCatalog)
	if err == nil {
		t.Fatalf("authorize = %q, %v, nil; want an error", actor, granted)
	}
}
