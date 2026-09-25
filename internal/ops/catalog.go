package ops

import "fmt"

// A Privilege is who may invoke a catalog operation. Its zero value is
// invalid, so a [CatalogEntry] built without setting one is caught by
// [BuildCatalog] rather than silently answering as one of the two.
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

// R-6bi5's four operation names, each a [CatalogEntry.Name].
const (
	OpStartBackup = "start_backup"
	OpJobStatus   = "job_status"
	OpCancelJob   = "cancel_job"
	OpCatalog     = "catalog"
)

// KindBackup is the job kind a backup job's record and lease carry:
// 1 to 64 lowercase ASCII letters, digits, and hyphens, the shape
// internal/engine/store's TakeKindLease requires of a kind lease's
// name (Task 1T). validJobKind checks the same shape here, since
// internal/ops imports no store type to check it directly.
const KindBackup = "backup"

// maxJobKindLen mirrors internal/engine/store's kind-lease name bound,
// restated here for the same reason handleLen restates NewID's length.
const maxJobKindLen = 64

// validJobKind reports whether s is a shape a kind lease's name
// admits: 1 to [maxJobKindLen] lowercase ASCII letters, digits, and
// hyphens.
func validJobKind(s string) bool {
	if s == "" || len(s) > maxJobKindLen {
		return false
	}
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			continue
		}
		return false
	}
	return true
}

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
// privilege is the zero value (spec criterion 1).
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
