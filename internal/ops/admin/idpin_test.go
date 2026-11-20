package admin

import (
	"errors"
	"testing"

	"github.com/glw907/dubplate/internal/ops"
)

// TestIDPinUserID pins the member id a removal and a demotion refuse:
// a short, an uppercase, a non-hex, an empty, and the all-zeros id each
// answer one invalid-request sentence.
func TestIDPinUserID(t *testing.T) {
	f := newFixture(t)
	const want = "user_id must be 32 lowercase hex characters naming a member"
	calls := map[string]func(id string) error{
		"remove": func(id string) error {
			_, err := f.seam.RemoveMember(writeCtx(t), actor(alice, fresh), ops.RemoveMemberRequest{UserID: id})
			return err
		},
		"demote": func(id string) error {
			_, err := f.seam.DemoteMember(writeCtx(t), actor(alice, fresh), ops.DemoteMemberRequest{UserID: id})
			return err
		},
	}
	ids := []struct{ name, id string }{
		{"31 characters", "0123456789abcdef0123456789abcde"},
		{"uppercase", "0123456789ABCDEF0123456789ABCDEF"},
		{"not hex", "0123456789abcdef0123456789abcdeg"},
		{"empty", ""},
		{"all zeros", "00000000000000000000000000000000"},
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
