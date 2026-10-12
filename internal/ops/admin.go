package ops

import "context"

// A Member is one roster entry as the admin operations carry it. UserID
// is the member's durable id as 32 lowercase hex characters, the name a
// remove or demote request addresses the member by.
type Member struct {
	UserID      string `json:"user_id"`
	Address     string `json:"address"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Admin       bool   `json:"admin"`
}

// A ListMembersRequest lists the live roster. It carries no fields.
type ListMembersRequest struct{}

// A ListMembersResponse carries the live members [AdminSeam.ListMembers]
// read.
type ListMembersResponse struct {
	Members []Member `json:"members"`
}

// An AddMemberRequest names the member to add. It carries no admin
// field: the admin grant stays the command line's alone, and a body
// naming one decodes to this request without it.
type AddMemberRequest struct {
	Address     string `json:"address"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
}

// An AddMemberResponse carries the member [AdminSeam.AddMember] added.
type AddMemberResponse struct {
	Member Member `json:"member"`
}

// A RemoveMemberRequest names the member [AdminSeam.RemoveMember]
// removes by UserID.
type RemoveMemberRequest struct {
	UserID string `json:"user_id"`
}

// A RemoveMemberResponse acknowledges a removal. It carries no fields.
type RemoveMemberResponse struct{}

// A DemoteMemberRequest names the member [AdminSeam.DemoteMember]
// demotes by UserID.
type DemoteMemberRequest struct {
	UserID string `json:"user_id"`
}

// A DemoteMemberResponse acknowledges a demotion. It carries no fields.
type DemoteMemberResponse struct{}

// A ListSettingsRequest lists every registered setting. It carries no
// fields.
type ListSettingsRequest struct{}

// A SettingEntry describes one registered key. Value, Source, and
// SeedValue are set for a store-homed key only, so a nil one means the
// key has no stored state to report, while a pointer to the empty
// string is a stored empty value. A wall or integration key's value
// never appears.
type SettingEntry struct {
	Name        string  `json:"name"`
	Home        string  `json:"home"`
	Description string  `json:"description"`
	Apply       string  `json:"apply"`
	Mark        string  `json:"mark"`
	Value       *string `json:"value,omitempty"`
	Source      *string `json:"source,omitempty"`
	SeedValue   *string `json:"seed_value,omitempty"`
}

// A ListSettingsResponse carries every registered key
// [AdminSeam.ListSettings] read.
type ListSettingsResponse struct {
	Settings []SettingEntry `json:"settings"`
}

// A SetSettingRequest names the key to write and its new value.
type SetSettingRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// A SetSettingResponse names the key written, the apply mode that says
// when the value takes effect, and the source in force afterwards.
type SetSettingResponse struct {
	Key    string `json:"key"`
	Apply  string `json:"apply"`
	Source string `json:"source"`
}

// A ResetSettingRequest names the key whose admin value is removed.
type ResetSettingRequest struct {
	Key string `json:"key"`
}

// A ResetSettingResponse names the key reset, the apply mode that says
// when the reset takes effect, and the source in force afterwards.
type ResetSettingResponse struct {
	Key    string `json:"key"`
	Apply  string `json:"apply"`
	Source string `json:"source"`
}

// The seven admin operation names, each a [CatalogEntry.Name].
const (
	OpListMembers  = "list_members"
	OpAddMember    = "add_member"
	OpRemoveMember = "remove_member"
	OpDemoteMember = "demote_member"
	OpListSettings = "list_settings"
	OpSetSetting   = "set_setting"
	OpResetSetting = "reset_setting"
)

// adminOperations lists the admin seam's operations in the order its
// methods declare them.
var adminOperations = []string{
	OpListMembers, OpAddMember, OpRemoveMember, OpDemoteMember,
	OpListSettings, OpSetSetting, OpResetSetting,
}

// AdminCatalogEntries returns one available entry for each admin
// operation, its privilege read from the one table [PrivilegeOf]
// reads. An operation missing from the table fails the call, so the
// catalog cannot list an operation no authorizer enforces.
func AdminCatalogEntries() ([]CatalogEntry, error) {
	return catalogEntries(adminOperations)
}

// An AdminSeam is what a caller reaches the seven member and settings
// operations through, beside [Seam]'s job operations. It is a second
// interface because [Seam]'s one implementation takes an instance
// lease and recovers jobs, which a server that runs no jobs must not
// do. Every method takes the acting [Actor] as its own parameter and
// returns an error wrapping one of this package's sentinels when it
// cannot answer. Each call is admin only; the authorizing layer
// enforces it before any method runs.
type AdminSeam interface {
	// ListMembers lists the live roster.
	ListMembers(ctx context.Context, actor Actor, req ListMembersRequest) (ListMembersResponse, error)

	// AddMember adds a member without the admin flag, or reactivates a
	// removed one.
	AddMember(ctx context.Context, actor Actor, req AddMemberRequest) (AddMemberResponse, error)

	// RemoveMember tombstones a member and revokes their sessions.
	RemoveMember(ctx context.Context, actor Actor, req RemoveMemberRequest) (RemoveMemberResponse, error)

	// DemoteMember clears a member's admin flag.
	DemoteMember(ctx context.Context, actor Actor, req DemoteMemberRequest) (DemoteMemberResponse, error)

	// ListSettings reports every registered key.
	ListSettings(ctx context.Context, actor Actor, req ListSettingsRequest) (ListSettingsResponse, error)

	// SetSetting writes a store-homed key's admin value.
	SetSetting(ctx context.Context, actor Actor, req SetSettingRequest) (SetSettingResponse, error)

	// ResetSetting removes a store-homed key's admin value.
	ResetSetting(ctx context.Context, actor Actor, req ResetSettingRequest) (ResetSettingResponse, error)
}
