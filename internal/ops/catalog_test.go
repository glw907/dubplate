package ops

import (
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

// TestKindBackup_SatisfiesLeaseKindRule proves the catalog's job kind
// meets the shape internal/engine/store's TakeKindLease requires
// (Task 1T), which internal/ops cannot check directly since it
// imports no store type.
func TestKindBackup_SatisfiesLeaseKindRule(t *testing.T) {
	if !validJobKind(KindBackup) {
		t.Errorf("validJobKind(%q) = false, want true", KindBackup)
	}
}

func TestValidJobKind(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want bool
	}{
		{"lowercase letters", "backup", true},
		{"digits and hyphens", "backup-2", true},
		{"empty", "", false},
		{"uppercase", "Backup", false},
		{"slash", "back/up", false},
		{"space", "back up", false},
		{"too long", strings.Repeat("a", maxJobKindLen+1), false},
		{"at the bound", strings.Repeat("a", maxJobKindLen), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validJobKind(tt.kind); got != tt.want {
				t.Errorf("validJobKind(%q) = %v, want %v", tt.kind, got, tt.want)
			}
		})
	}
}
