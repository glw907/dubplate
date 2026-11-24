package settings_test

import (
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ops/jobs"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// TestLocalSnapshotKeep is the keep decision's table: one subtest per
// state, over a store.Report built by hand ([TestStoreRowStatus]'s own
// precedent) and the real store registry. The first row's fixture
// leaves the keep key unset in the environment, since that is the
// state in which the outcome's post-seed value (the default, 3) and
// the stored row's value (1) differ; a row that set the key would pass
// by reading either.
func TestLocalSnapshotKeep(t *testing.T) {
	reg := config.NewRegistry()
	jobs.Register(reg)
	store.Register(reg)
	key := store.LocalSnapshotKeepSetting.Key
	retention := "DUBPLATE_JOB_RETENTION"

	facts := func(rows ...store.SettingsRow) *store.Facts { return &store.Facts{Settings: rows} }
	row := func(key, value, source string) store.SettingsRow {
		return store.SettingsRow{Key: key, Value: value, Source: source, SeedValue: value}
	}
	current := func(f *store.Facts) store.Report { return store.Report{State: store.StateCurrent, Facts: f} }
	older := func(f *store.Facts) store.Report { return store.Report{State: store.StateOlder, Facts: f} }

	tests := []struct {
		name    string
		report  store.Report
		environ []string
		want    int
		wantErr []string
	}{
		{
			name:   "env-seed row of 1, key unset in the environment",
			report: older(facts(row(key, "1", "env-seed"))),
			want:   1,
		},
		{
			name:    "in-bounds row, the environment names another in-bounds value",
			report:  older(facts(row(key, "5", "env-seed"))),
			environ: []string{key + "=7"},
			want:    5,
		},
		{
			name:    "admin row outside its bounds",
			report:  older(facts(row(key, "500", "admin"))),
			wantErr: []string{key, "above the ceiling 100"},
		},
		{
			name:    "env-seed row outside its bounds, the environment re-seeds it in bounds",
			report:  older(facts(row(key, "500", "env-seed"))),
			environ: []string{key + "=4"},
			want:    4,
		},
		{
			name:   "env-seed row outside its bounds, the environment no longer sets it",
			report: older(facts(row(key, "500", "env-seed"))),
			want:   3,
		},
		{
			name:   "no keep row",
			report: older(facts(row(retention, "48h", "env-seed"))),
			want:   3,
		},
		{
			name:   "no facts: absent",
			report: store.Report{State: store.StateAbsent},
			want:   3,
		},
		{
			name:   "no facts: busy",
			report: store.Report{State: store.StateBusy},
			want:   3,
		},
		{
			name:   "no facts: restore in progress",
			report: store.Report{State: store.StateRestoreInProgress},
			want:   3,
		},
		{
			name:   "no facts: refused",
			report: store.Report{State: store.StateRefused},
			want:   3,
		},
		{
			name:   "no settings table",
			report: older(&store.Facts{SettingsAbsent: true}),
			want:   3,
		},
		{
			name:    "older store, another setting's stored row refuses",
			report:  older(facts(row(retention, "1h", "admin"))),
			wantErr: []string{retention, "below the floor 24h"},
		},
		{
			name:   "current store, another setting's stored row refuses: the seed decides",
			report: current(facts(row(retention, "1h", "admin"))),
			want:   3,
		},
		{
			name:    "older store, the environment's keep value does not parse",
			report:  older(facts(row(retention, "48h", "env-seed"))),
			environ: []string{key + "=many"},
			wantErr: []string{key, "invalid value"},
		},
		{
			name:    "current store, the environment's keep value does not parse: the seed decides",
			report:  current(facts(row(retention, "48h", "env-seed"))),
			environ: []string{key + "=many"},
			want:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := settings.LocalSnapshotKeep(tt.report, tt.environ, reg)
			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatalf("settings.LocalSnapshotKeep() = %d, nil, want an error naming %q", got, tt.wantErr)
				}
				for _, want := range tt.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error = %q, want it to contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("settings.LocalSnapshotKeep() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("settings.LocalSnapshotKeep() = %d, want %d", got, tt.want)
			}
		})
	}
}
