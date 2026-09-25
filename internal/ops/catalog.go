package ops

import "fmt"

// A Privilege is who may invoke a catalog operation. Its zero value is
// invalid, so a [CatalogEntry] built without setting one is caught by
// [BuildCatalog] rather than silently answering as one of the two. It
// marshals as the text "member" or "admin", never as an integer, so
// the wire vocabulary survives a reordering of these constants.
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

// The four operation names, each a [CatalogEntry.Name].
const (
	OpStartBackup = "start_backup"
	OpJobStatus   = "job_status"
	OpCancelJob   = "cancel_job"
	OpCatalog     = "catalog"
)

// KindBackup is the job kind a backup job's record and lease carry.
const KindBackup = "backup"

// A CatalogEntry reports one operation's privilege and whether a
// caller can invoke it right now.
type CatalogEntry struct {
	Name      string    `json:"name"`
	Privilege Privilege `json:"privilege"`
	Available bool      `json:"available"`
	Reason    string    `json:"reason,omitempty"`
}

// A Catalog is the entry set [BuildCatalog] validated.
type Catalog struct {
	entries []CatalogEntry
}

// BuildCatalog validates entries and returns the [Catalog]
// [Seam.Catalog] answers with. It fails, naming the first entry whose
// privilege is the zero value.
func BuildCatalog(entries ...CatalogEntry) (Catalog, error) {
	for _, e := range entries {
		if !e.Privilege.valid() {
			return Catalog{}, fmt.Errorf("ops: catalog entry %q: privilege is not set", e.Name)
		}
	}
	return Catalog{entries: append([]CatalogEntry(nil), entries...)}, nil
}

// Entries returns c's validated entries.
func (c Catalog) Entries() []CatalogEntry { return c.entries }
