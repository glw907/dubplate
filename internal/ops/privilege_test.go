package ops

import (
	"encoding/json"
	"testing"
)

// TestPrivilege_MarshalsAsText proves a [Privilege] renders as the
// wire literal "member" or "admin", never as the underlying integer,
// and refuses the zero value.
func TestPrivilege_MarshalsAsText(t *testing.T) {
	tests := []struct {
		name    string
		p       Privilege
		want    string
		wantErr bool
	}{
		{"member", PrivilegeMember, `"member"`, false},
		{"admin", PrivilegeAdmin, `"admin"`, false},
		{"zero value", Privilege(0), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.p)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Marshal(%d) = %s, want an error", int(tt.p), got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Marshal() = %v, want no error", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestPrivilegeOf pins the one privilege table: the seven admin
// operations require admin, the five submission operations require
// member, and any other name is not found, the four operations that
// left the web API among them.
func TestPrivilegeOf(t *testing.T) {
	tests := []struct {
		op     string
		want   Privilege
		wantOK bool
	}{
		{OpListMembers, PrivilegeAdmin, true},
		{OpAddMember, PrivilegeAdmin, true},
		{OpRemoveMember, PrivilegeAdmin, true},
		{OpDemoteMember, PrivilegeAdmin, true},
		{OpListSettings, PrivilegeAdmin, true},
		{OpSetSetting, PrivilegeAdmin, true},
		{OpResetSetting, PrivilegeAdmin, true},
		{OpCreateSubmission, PrivilegeMember, true},
		{OpListSubmissions, PrivilegeMember, true},
		{OpGetSubmission, PrivilegeMember, true},
		{OpRetrySubmission, PrivilegeMember, true},
		{OpCancelSubmission, PrivilegeMember, true},
		{"start_backup", 0, false},
		{"job_status", 0, false},
		{"cancel_job", 0, false},
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

// TestSubmissionOperationNames pins the five submission operations'
// wire names as literals: a client posts to each one.
func TestSubmissionOperationNames(t *testing.T) {
	want := map[string]string{
		OpCreateSubmission: "create_submission",
		OpListSubmissions:  "list_submissions",
		OpGetSubmission:    "get_submission",
		OpRetrySubmission:  "retry_submission",
		OpCancelSubmission: "cancel_submission",
	}
	for got, literal := range want {
		if got != literal {
			t.Errorf("operation constant = %q, want %q", got, literal)
		}
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
