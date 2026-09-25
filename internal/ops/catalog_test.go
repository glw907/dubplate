package ops

import (
	"encoding/json"
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
