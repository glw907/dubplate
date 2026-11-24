package admin

import (
	"errors"
	"testing"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/ops"
)

// TestIDPinUserID pins the member id a removal and a demotion refuse.
// The JSON decoder upstream refuses a malformed id, so the seam sees
// only a typed id: the all-zeros id answers one invalid-request
// sentence.
func TestIDPinUserID(t *testing.T) {
	f := newFixture(t)
	const want = "user_id must be 32 lowercase hex characters naming a member"
	calls := map[string]func(target id.ID) error{
		"remove": func(target id.ID) error {
			_, err := f.seam.RemoveMember(writeCtx(t), actor(alice, fresh), ops.RemoveMemberRequest{UserID: target})
			return err
		},
		"demote": func(target id.ID) error {
			_, err := f.seam.DemoteMember(writeCtx(t), actor(alice, fresh), ops.DemoteMemberRequest{UserID: target})
			return err
		},
	}
	ids := []struct {
		name string
		id   id.ID
	}{
		{"all zeros", id.ID{}},
	}
	for name, call := range calls {
		for _, tt := range ids {
			t.Run(name+"/"+tt.name, func(t *testing.T) {
				err := call(tt.id)
				e, ok := errors.AsType[*ops.Error](err)
				if !ok {
					t.Fatalf("%s(%q) = %v, want an *ops.Error", name, tt.id, err)
				}
				if e.Code != ops.CodeInvalidRequest || e.Reason != want {
					t.Errorf("%s(%q) = code %q, reason %q; want code %q, reason %q", name, tt.id, e.Code, e.Reason, ops.CodeInvalidRequest, want)
				}
			})
		}
	}
}
