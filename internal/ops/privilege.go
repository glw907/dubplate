package ops

import "fmt"

// A Privilege is who may invoke an operation. Its zero value is
// invalid, so a caller that forgets to set one is refused rather than
// silently admitted as one of the two. It marshals as the text
// "member" or "admin", never as an integer, so the wire vocabulary
// survives a reordering of these constants.
type Privilege int

const (
	// PrivilegeMember admits any signed-in member.
	PrivilegeMember Privilege = iota + 1

	// PrivilegeAdmin admits an admin only.
	PrivilegeAdmin
)

func (p Privilege) valid() bool {
	return p == PrivilegeMember || p == PrivilegeAdmin
}

// String returns p's wire text, or "invalid" for the zero value or
// any other unrecognized value.
func (p Privilege) String() string {
	switch p {
	case PrivilegeMember:
		return "member"
	case PrivilegeAdmin:
		return "admin"
	default:
		return "invalid"
	}
}

// MarshalText implements [encoding.TextMarshaler], rendering p as
// "member" or "admin". It fails for the zero value or any other
// unrecognized value, so an invalid [Privilege] never reaches the
// wire silently.
func (p Privilege) MarshalText() ([]byte, error) {
	if !p.valid() {
		return nil, fmt.Errorf("ops: privilege %d has no wire text", int(p))
	}
	return []byte(p.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]. It accepts
// only the literal text "member" or "admin"; any other text,
// including empty, is refused.
func (p *Privilege) UnmarshalText(text []byte) error {
	switch string(text) {
	case "member":
		*p = PrivilegeMember
	case "admin":
		*p = PrivilegeAdmin
	default:
		return fmt.Errorf("ops: privilege %q is not member or admin", text)
	}
	return nil
}

// KindBackup is the job kind a backup job's record and lease carry.
const KindBackup = "backup"

// privileges maps each of the twelve operation names to the privilege
// it requires. [PrivilegeOf] is the only reader, so the authorizing
// layer and the API document can never disagree about what an
// operation needs.
var privileges = map[string]Privilege{
	OpListMembers:  PrivilegeAdmin,
	OpAddMember:    PrivilegeAdmin,
	OpRemoveMember: PrivilegeAdmin,
	OpDemoteMember: PrivilegeAdmin,
	OpListSettings: PrivilegeAdmin,
	OpSetSetting:   PrivilegeAdmin,
	OpResetSetting: PrivilegeAdmin,

	OpCreateSubmission: PrivilegeMember,
	OpListSubmissions:  PrivilegeMember,
	OpGetSubmission:    PrivilegeMember,
	OpRetrySubmission:  PrivilegeMember,
	OpCancelSubmission: PrivilegeMember,
}

// PrivilegeOf returns the privilege op requires and reports whether op
// is a recognized operation name. An unrecognized name answers ok
// false and a zero [Privilege], never a guessed one: its caller
// answers not_found before any authorizer runs.
func PrivilegeOf(op string) (Privilege, bool) {
	p, ok := privileges[op]
	return p, ok
}
