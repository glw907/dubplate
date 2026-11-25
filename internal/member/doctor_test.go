package member

import (
	"testing"

	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/store"
)

// read is a report whose inspection read facts, at version 3.
func read(state store.State) store.Report {
	return store.Report{State: state, StoreVersion: 3, Facts: &store.Facts{}}
}

func TestDoctorRows(t *testing.T) {
	tests := []struct {
		name   string
		report store.Report
		facts  Facts
		want   doctor.Row
	}{
		{name: "the store was not inspected", want: doctor.Row{Name: "members", Status: "not read"}},
		{
			name:   "no store yet",
			report: store.Report{State: store.StateAbsent},
			want:   doctor.Row{Name: "members", Status: "no store yet; start serve, then member add --admin"},
		},
		{
			name:   "a state the store row refuses",
			report: store.Report{State: store.StateBusy},
			want:   doctor.Row{Name: "members", Status: "not read"},
		},
		{
			name:   "a store older than the users table",
			report: store.Report{State: store.StateOlder, StoreVersion: 2, Facts: &store.Facts{}},
			want:   doctor.Row{Name: "members", Status: "members not recorded before store version 3"},
		},
		{
			name:   "a current store that carries no facts",
			report: store.Report{State: store.StateCurrent, StoreVersion: 3},
			facts:  Facts{members: 2, admins: 1},
			want:   doctor.Row{Name: "members", Status: "not read"},
		},
		{
			name:   "no members",
			report: read(store.StateCurrent),
			want:   doctor.Row{Name: "members", Status: "no members; run dubplate-server member add --admin", Failed: true},
		},
		{
			name:   "members and no admin",
			report: read(store.StateCurrent),
			facts:  Facts{members: 1},
			want:   doctor.Row{Name: "members", Status: "members: 1, admins: 0; run dubplate-server member add --admin", Failed: true},
		},
		{
			name:   "members and an admin",
			report: read(store.StateCurrent),
			facts:  Facts{members: 2, admins: 1},
			want:   doctor.Row{Name: "members", Status: "members: 2, admins: 1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DoctorRows(tt.report, tt.facts)
			if len(got) != 1 || got[0] != tt.want {
				t.Errorf("DoctorRows() = %+v, want [%+v]", got, tt.want)
			}
		})
	}
}

func TestRefineAdminEmail(t *testing.T) {
	const (
		notMember = "warning: this address is not a live member; run dubplate-server member add <address> --admin, or set DUBPLATE_ADMIN_EMAIL to a live admin"
		notAdmin  = "warning: this address is a live member without the admin flag; set DUBPLATE_ADMIN_EMAIL to a live admin"
	)
	older := store.Report{State: store.StateOlder, StoreVersion: 2, Facts: &store.Facts{}}
	tests := []struct {
		name   string
		wall   string
		report store.Report
		facts  Facts
		want   string
	}{
		{name: "a wall failure stands", wall: "missing", report: read(store.StateCurrent), facts: Facts{1, 1, standingNone}, want: "missing"},
		{name: "the store was not inspected", wall: "ok", facts: Facts{1, 1, standingNone}, want: "ok"},
		{name: "a state the store row refuses", wall: "ok", report: read(store.StateBusy), facts: Facts{1, 1, standingNone}, want: "ok"},
		{name: "no facts", wall: "ok", report: store.Report{State: store.StateCurrent, StoreVersion: 3}, facts: Facts{1, 1, standingNone}, want: "ok"},
		{name: "a store older than the users table", wall: "ok", report: older, facts: Facts{1, 1, standingNone}, want: "ok"},
		{name: "no live member, which the members row fails", wall: "ok", report: read(store.StateCurrent), facts: Facts{0, 0, standingNone}, want: "ok"},
		{name: "an address no member holds", wall: "ok", report: read(store.StateCurrent), facts: Facts{1, 1, standingNone}, want: notMember},
		{name: "a removed member", wall: "ok", report: read(store.StateCurrent), facts: Facts{1, 1, standingRemoved}, want: notMember},
		{name: "a live member without the admin flag", wall: "ok", report: read(store.StateCurrent), facts: Facts{2, 1, standingMember}, want: notAdmin},
		{name: "a live admin", wall: "ok", report: read(store.StateCurrent), facts: Facts{2, 1, standingAdmin}, want: "ok"},
		{name: "a standing the inspection did not read", wall: "ok", report: read(store.StateCurrent), facts: Facts{2, 1, standingNotRead}, want: "ok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := doctor.Row{Name: "DUBPLATE_ADMIN_EMAIL", Status: tt.wall}
			want := doctor.Row{Name: "DUBPLATE_ADMIN_EMAIL", Status: tt.want}
			if got := RefineAdminEmail(row, tt.report, tt.facts); got != want {
				t.Errorf("RefineAdminEmail() = %+v, want %+v", got, want)
			}
		})
	}
}

// TestRefineAdminEmail_OtherRow: a wall row other than the admin
// email's passes through unchanged, in the state that would warn the
// admin email row.
func TestRefineAdminEmail_OtherRow(t *testing.T) {
	row := doctor.Row{Name: "DUBPLATE_SITE_NAME", Status: "ok"}
	got := RefineAdminEmail(row, read(store.StateCurrent), Facts{1, 1, standingNone})
	if got != row {
		t.Errorf("RefineAdminEmail() = %+v, want %+v", got, row)
	}
}
