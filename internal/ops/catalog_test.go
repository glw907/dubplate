package ops

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

// TestBuildCatalog_Privileges pins spec criterion 1: every catalog
// entry's privilege is member or admin, and the zero value fails
// construction naming the operation.
func TestBuildCatalog_Privileges(t *testing.T) {
	tests := []struct {
		name      string
		privilege Privilege
		wantErr   bool
	}{
		{"member", PrivilegeMember, false},
		{"admin", PrivilegeAdmin, false},
		{"zero value", Privilege(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildCatalog(CatalogEntry{Name: OpCatalog, Privilege: tt.privilege, Available: true})
			if tt.wantErr {
				if err == nil {
					t.Fatal("BuildCatalog() = nil, want an error naming the operation")
				}
				if !strings.Contains(err.Error(), OpCatalog) {
					t.Errorf("error %q does not name the operation %q", err, OpCatalog)
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildCatalog() = %v, want no error", err)
			}
		})
	}
}

func TestBuildCatalog_Entries(t *testing.T) {
	want := []CatalogEntry{
		{Name: OpStartBackup, Privilege: PrivilegeAdmin, Available: true},
		{Name: OpJobStatus, Privilege: PrivilegeAdmin, Available: true},
		{Name: OpCancelJob, Privilege: PrivilegeAdmin, Available: true},
		{Name: OpCatalog, Privilege: PrivilegeMember, Available: true},
	}
	cat, err := BuildCatalog(want...)
	if err != nil {
		t.Fatalf("BuildCatalog() = %v, want no error", err)
	}
	got := cat.Entries()
	if len(got) != len(want) {
		t.Fatalf("Entries() has %d entries, want %d", len(got), len(want))
	}
	for i, e := range got {
		if e != want[i] {
			t.Errorf("Entries()[%d] = %+v, want %+v", i, e, want[i])
		}
	}
}

// TestCatalogEntry_MarshalsPrivilegeAsText proves a [CatalogEntry]'s
// privilege renders as the wire literal "member" or "admin", never as
// the underlying integer.
func TestCatalogEntry_MarshalsPrivilegeAsText(t *testing.T) {
	got, err := json.Marshal(CatalogEntry{Name: OpStartBackup, Privilege: PrivilegeAdmin, Available: true})
	if err != nil {
		t.Fatalf("Marshal() = %v, want no error", err)
	}
	if want := `"privilege":"admin"`; !strings.Contains(string(got), want) {
		t.Errorf("Marshal() = %s, want it to contain %s", got, want)
	}
}

// TestPrivilegeOf pins R-6bii1's one privilege table against 6b-i's
// shipped catalog values (local.go's Catalog method, before this
// pass): start_backup, job_status, and cancel_job are admin, catalog
// is member, and any other name is not found.
func TestPrivilegeOf(t *testing.T) {
	tests := []struct {
		op     string
		want   Privilege
		wantOK bool
	}{
		{OpStartBackup, PrivilegeAdmin, true},
		{OpJobStatus, PrivilegeAdmin, true},
		{OpCancelJob, PrivilegeAdmin, true},
		{OpCatalog, PrivilegeMember, true},
		{OpListMembers, PrivilegeAdmin, true},
		{OpAddMember, PrivilegeAdmin, true},
		{OpRemoveMember, PrivilegeAdmin, true},
		{OpDemoteMember, PrivilegeAdmin, true},
		{OpListSettings, PrivilegeAdmin, true},
		{OpSetSetting, PrivilegeAdmin, true},
		{OpResetSetting, PrivilegeAdmin, true},
		{"delete_everything", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			got, ok := PrivilegeOf(tt.op)
			if ok != tt.wantOK {
				t.Fatalf("PrivilegeOf(%q) ok = %v, want %v", tt.op, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("PrivilegeOf(%q) = %v, want %v", tt.op, got, tt.want)
			}
		})
	}
}

// TestAdminOperationNames pins the seven admin operations' wire names
// as literals: a client posts to each one.
func TestAdminOperationNames(t *testing.T) {
	want := map[string]string{
		OpListMembers:  "list_members",
		OpAddMember:    "add_member",
		OpRemoveMember: "remove_member",
		OpDemoteMember: "demote_member",
		OpListSettings: "list_settings",
		OpSetSetting:   "set_setting",
		OpResetSetting: "reset_setting",
	}
	for got, literal := range want {
		if got != literal {
			t.Errorf("operation constant = %q, want %q", got, literal)
		}
	}
}

// TestAdminCatalogEntries proves the admin seam's catalog share is one
// available admin entry for each of the seven operations, in the
// order the admin seam declares them.
func TestAdminCatalogEntries(t *testing.T) {
	want := []CatalogEntry{
		{Name: "list_members", Privilege: PrivilegeAdmin, Available: true},
		{Name: "add_member", Privilege: PrivilegeAdmin, Available: true},
		{Name: "remove_member", Privilege: PrivilegeAdmin, Available: true},
		{Name: "demote_member", Privilege: PrivilegeAdmin, Available: true},
		{Name: "list_settings", Privilege: PrivilegeAdmin, Available: true},
		{Name: "set_setting", Privilege: PrivilegeAdmin, Available: true},
		{Name: "reset_setting", Privilege: PrivilegeAdmin, Available: true},
	}
	got, err := AdminCatalogEntries()
	if err != nil {
		t.Fatalf("AdminCatalogEntries() = %v, want no error", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("AdminCatalogEntries() = %+v, want %+v", got, want)
	}
}

// TestPrivilege_UnmarshalText proves UnmarshalText accepts only the
// two wire literals, refusing any other text including empty.
func TestPrivilege_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    Privilege
		wantErr bool
	}{
		{"member", "member", PrivilegeMember, false},
		{"admin", "admin", PrivilegeAdmin, false},
		{"empty", "", 0, true},
		{"unknown", "owner", 0, true},
		{"wrong case", "Admin", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Privilege
			err := p.UnmarshalText([]byte(tt.text))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalText(%q) = nil, want an error", tt.text)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalText(%q) = %v, want no error", tt.text, err)
			}
			if p != tt.want {
				t.Errorf("UnmarshalText(%q) set %v, want %v", tt.text, p, tt.want)
			}
		})
	}
}
