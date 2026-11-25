package ops

// A Privilege is who may invoke an operation. Its zero value is
// invalid, so a caller that forgets to set one is refused rather than
// silently admitted as one of the two.
type Privilege int

const (
	// PrivilegeMember admits any signed-in member.
	PrivilegeMember Privilege = iota + 1

	// PrivilegeAdmin admits an admin only.
	PrivilegeAdmin
)

// String returns "member" or "admin", or "invalid" for the zero value
// or any other unrecognized value.
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

// An Op is an operation's name, the wire name a client posts to.
type Op string

// privileges maps each operation to the privilege it requires.
// [PrivilegeOf] is the only reader, so the authorizing layer has one
// answer about what an operation needs.
var privileges = map[Op]Privilege{
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
// false and a zero [Privilege], never a guessed one, so a caller
// treats it as no operation at all.
func PrivilegeOf(op Op) (Privilege, bool) {
	p, ok := privileges[op]
	return p, ok
}
