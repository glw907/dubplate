package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/store"
)

// settingsDoctorRowsTestSetting and settingsDoctorRowsSensitive are the
// two settings TestDoctorRows registers: an ordinary integer (floor
// 3, ceiling 100) and a sensitive one, whose value is withheld.
var (
	settingsDoctorRowsTestSetting = config.Setting{
		Key: "DUBPLATE_TEST_SETTING", Kind: config.KindInt, Default: "10",
		Doc: "an integer", Apply: config.ApplyNextOperation, Floor: "3", Ceiling: "100",
		Description: testDescription, Sensitivity: config.NotSensitive,
	}
	settingsDoctorRowsSensitive = config.Setting{
		Key: "DUBPLATE_TEST_SECRET", Kind: config.KindInt, Default: "0",
		Doc: "an integer", Apply: config.ApplyImmediately, Floor: "0", Ceiling: "100",
		Description: testDescription, Sensitivity: config.Sensitive,
	}
)

// settingsDoctorRowsRegistry declares the two settings above.
func settingsDoctorRowsRegistry() *config.Registry {
	reg := config.NewRegistry()
	reg.AddSetting(settingsDoctorRowsTestSetting)
	reg.AddSetting(settingsDoctorRowsSensitive)
	return reg
}

// TestDoctorRows is the settings rows' golden: one literal expected
// row per state of the doctor table, from a store.Report and settings
// facts built by hand, so an older or table-absent store needs no
// runner to build it.
func TestDoctorRows(t *testing.T) {
	key, secretKey := settingsDoctorRowsTestSetting.Key, settingsDoctorRowsSensitive.Key
	current := store.Report{State: store.StateCurrent, Facts: &store.Facts{}}

	tests := []struct {
		name    string
		environ []string
		report  store.Report
		facts   Facts
		want    []doctor.Row
	}{
		{
			name:   "no row, env absent: the default is in force",
			report: current,
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "no row, env present: seeds at serve's next start",
			environ: []string{key + "=20"},
			report:  current,
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); env "20" seeds at serve's next start`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "env_seed row, env equal: the seeded value is in force",
			environ: []string{key + "=20"},
			report:  current,
			facts:   Facts{Rows: []Row{{Key: key, Value: "20", Source: "env_seed", SeedValue: "20"}}},
			want: []doctor.Row{
				{Name: key, Status: `"20" (source env_seed, next operation)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "env_seed row, env different: re-seeds at serve's next start",
			environ: []string{key + "=30"},
			report:  current,
			facts:   Facts{Rows: []Row{{Key: key, Value: "20", Source: "env_seed", SeedValue: "20"}}},
			want: []doctor.Row{
				{Name: key, Status: `"20" (source env_seed, next operation); env "30" re-seeds at serve's next start`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:   "env_seed row, env absent: returns to the default at serve's next start",
			report: current,
			facts:  Facts{Rows: []Row{{Key: key, Value: "20", Source: "env_seed", SeedValue: "20"}}},
			want: []doctor.Row{
				{Name: key, Status: `"20" (source env_seed, next operation); returns to the default at serve's next start`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:   "admin row, env absent: the administrator's value is in force",
			report: current,
			facts:  Facts{Rows: []Row{{Key: key, Value: "40", Source: "admin", SeedValue: "20"}}},
			want: []doctor.Row{
				{Name: key, Status: `"40" (source admin, next operation)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "admin row, env equal: the administrator's value is in force",
			environ: []string{key + "=40"},
			report:  current,
			facts:   Facts{Rows: []Row{{Key: key, Value: "40", Source: "admin", SeedValue: "20"}}},
			want: []doctor.Row{
				{Name: key, Status: `"40" (source admin, next operation)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "admin row, env different: refusing",
			environ: []string{key + "=50"},
			report:  current,
			facts:   Facts{Rows: []Row{{Key: key, Value: "40", Source: "admin", SeedValue: "20"}}},
			want: []doctor.Row{
				{Name: key, Status: `refused: DUBPLATE_TEST_SETTING: the environment seeds "50" but an administrator set "40"; unset DUBPLATE_TEST_SETTING, or set it to "40", or clear the administrator's value by running dubplate-server settings reset DUBPLATE_TEST_SETTING`, Failed: true},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:   "out-of-bounds admin value stays in force: refusing",
			report: current,
			facts:  Facts{Rows: []Row{{Key: key, Value: "1", Source: "admin", SeedValue: "1"}}},
			want: []doctor.Row{
				{Name: key, Status: `refused: stored admin value refused: DUBPLATE_TEST_SETTING: value "1" is below the floor 3; clear the administrator's value with: dubplate-server settings reset DUBPLATE_TEST_SETTING`, Failed: true},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "out-of-bounds env_seed row healed by re-seed",
			environ: []string{key + "=15"},
			report:  current,
			facts:   Facts{Rows: []Row{{Key: key, Value: "1", Source: "env_seed", SeedValue: "1"}}},
			want: []doctor.Row{
				{Name: key, Status: `"1" (source env_seed, next operation); env "15" re-seeds at serve's next start`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:   "out-of-bounds env_seed row healed by removal",
			report: current,
			facts:  Facts{Rows: []Row{{Key: key, Value: "1", Source: "env_seed", SeedValue: "1"}}},
			want: []doctor.Row{
				{Name: key, Status: `"1" (source env_seed, next operation); returns to the default at serve's next start`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:   "a stored value that does not parse: refusing naming the reason",
			report: current,
			facts:  Facts{Rows: []Row{{Key: key, Value: "not-a-number", Source: "env_seed", SeedValue: "not-a-number"}}},
			want: []doctor.Row{
				{Name: key, Status: `refused: stored env_seed value refused: DUBPLATE_TEST_SETTING: invalid value "not-a-number", want an integer`, Failed: true},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "an env value that does not parse: refusing naming the reason",
			environ: []string{key + "=abc"},
			report:  current,
			want: []doctor.Row{
				{Name: key, Status: `refused: environment seed refused: DUBPLATE_TEST_SETTING: invalid value "abc", want an integer`, Failed: true},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
			},
		},
		{
			name:   "an orphan row for an unregistered key",
			report: current,
			facts:  Facts{Rows: []Row{{Key: "DUBPLATE_TEST_RETIRED", Value: "1", Source: "admin", SeedValue: "1"}}},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive)"},
				{Name: "DUBPLATE_TEST_RETIRED", Status: "no longer registered; ignored"},
			},
		},
		{
			name:   "absent store: holds no settings yet",
			report: store.Report{State: store.StateAbsent},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store holds no settings yet`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:    "absent store, env present: seeds at serve's next start",
			environ: []string{key + "=20"},
			report:  store.Report{State: store.StateAbsent},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store holds no settings yet; env "20" seeds at serve's next start`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:   "older store below the version that adds the settings table",
			report: store.Report{State: store.StateOlder, StoreVersion: 3, Facts: &store.Facts{}},
			facts:  Facts{Absent: true},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store holds no settings yet`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:   "older store below the version that adds the jobs table: nil facts",
			report: store.Report{State: store.StateOlder, StoreVersion: 1},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store holds no settings yet`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:   "busy: not checked",
			report: store.Report{State: store.StateBusy},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked (busy)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (busy)"},
			},
		},
		{
			name:    "busy, env present: not checked against the store",
			environ: []string{key + "=20"},
			report:  store.Report{State: store.StateBusy},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked (busy); env "20" not checked against the store`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (busy)"},
			},
		},
		{
			name:   "restore in progress: not checked",
			report: store.Report{State: store.StateRestoreInProgress},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked (restore in progress)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (restore in progress)"},
			},
		},
		{
			name:   "not a store: not checked",
			report: store.Report{State: store.StateNotAStore},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked (not a store)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (not a store)"},
			},
		},
		{
			name:   "refused: not checked",
			report: store.Report{State: store.StateRefused},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked (refused)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (refused)"},
			},
		},
		{
			name:   "newer: not checked",
			report: store.Report{State: store.StateNewer},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked (newer)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (newer)"},
			},
		},
		{
			name:   "an unrecognized state: not checked",
			report: store.Report{State: store.State(99), Facts: &store.Facts{}},
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked (unknown)`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (unknown)"},
			},
		},
		{
			name: "the store row never ran Inspect: not checked",
			want: []doctor.Row{
				{Name: key, Status: `"10" (source default, next operation); store not checked`},
				{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := DoctorRows(tt.environ, settingsDoctorRowsRegistry(), tt.report, tt.facts)
			if !slices.Equal(rows, tt.want) {
				t.Errorf("rows = %+v, want %+v", rows, tt.want)
			}
		})
	}
}

// TestDoctorRows_SensitiveWithheld: no row of a [config.Sensitive]
// setting prints its stored or environment value, in force, changing,
// unchecked, or refused. Fixture state: the stored value 42 and the
// environment value 53 appear nowhere else in a row, so either leaking
// shows as its digits.
func TestDoctorRows_SensitiveWithheld(t *testing.T) {
	secretKey := settingsDoctorRowsSensitive.Key
	current := store.Report{State: store.StateCurrent, Facts: &store.Facts{}}
	tests := []struct {
		name    string
		reg     func() *config.Registry
		environ []string
		report  store.Report
		facts   Facts
		want    doctor.Row
	}{
		{
			name:   "a stored value in force",
			report: current,
			facts:  Facts{Rows: []Row{{Key: secretKey, Value: "42", Source: "admin", SeedValue: "42"}}},
			want:   doctor.Row{Name: secretKey, Status: "(withheld) (source admin, immediately, sensitive)"},
		},
		{
			name:    "an environment value that re-seeds",
			environ: []string{secretKey + "=53"},
			report:  current,
			facts:   Facts{Rows: []Row{{Key: secretKey, Value: "42", Source: "env_seed", SeedValue: "42"}}},
			want:    doctor.Row{Name: secretKey, Status: "(withheld) (source env_seed, immediately, sensitive); env (withheld) re-seeds at serve's next start"},
		},
		{
			name:    "an environment value that seeds an absent store",
			environ: []string{secretKey + "=53"},
			report:  store.Report{State: store.StateAbsent},
			want:    doctor.Row{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store holds no settings yet; env (withheld) seeds at serve's next start"},
		},
		{
			name:    "an environment value not checked against a busy store",
			environ: []string{secretKey + "=53"},
			report:  store.Report{State: store.StateBusy},
			want:    doctor.Row{Name: secretKey, Status: "(withheld) (source default, immediately, sensitive); store not checked (busy); env (withheld) not checked against the store"},
		},
		{
			name:    "an environment value an administrator's value refuses",
			environ: []string{secretKey + "=53"},
			report:  current,
			facts:   Facts{Rows: []Row{{Key: secretKey, Value: "42", Source: "admin", SeedValue: "42"}}},
			want: doctor.Row{Name: secretKey, Failed: true, Status: "refused: DUBPLATE_TEST_SECRET: the environment seeds (withheld) but an administrator set (withheld); " +
				"unset DUBPLATE_TEST_SECRET, or set it to (withheld), or clear the administrator's value by running dubplate-server settings reset DUBPLATE_TEST_SECRET"},
		},
		{
			name:    "an environment value that does not parse",
			environ: []string{secretKey + "=53.5"},
			report:  current,
			want:    doctor.Row{Name: secretKey, Failed: true, Status: "refused: environment seed refused: DUBPLATE_TEST_SECRET: invalid value (withheld), want an integer"},
		},
		{
			name:    "an environment value above the ceiling",
			environ: []string{secretKey + "=530"},
			report:  current,
			want:    doctor.Row{Name: secretKey, Failed: true, Status: "refused: environment seed refused: DUBPLATE_TEST_SECRET: value (withheld) is above the ceiling 100"},
		},
		{
			name:   "a stored value a correlated rule refuses",
			reg:    secretBelowSettingRegistry,
			report: current,
			facts:  Facts{Rows: []Row{{Key: secretKey, Value: "42", Source: "admin", SeedValue: "42"}}},
			want: doctor.Row{Name: secretKey, Failed: true, Status: `refused: correlated rule "secret below the setting": ` +
				`DUBPLATE_TEST_SECRET (withheld) is not below DUBPLATE_TEST_SETTING "10"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := settingsDoctorRowsRegistry
			if tt.reg != nil {
				reg = tt.reg
			}
			rows := DoctorRows(tt.environ, reg(), tt.report, tt.facts)
			i := slices.IndexFunc(rows, func(r doctor.Row) bool { return r.Name == secretKey })
			if i < 0 {
				t.Fatalf("rows carry no entry for %s: %+v", secretKey, rows)
			}
			if rows[i] != tt.want {
				t.Errorf("row = %+v, want %+v", rows[i], tt.want)
			}
			// Every row, the other key's included, withholds the
			// sensitive values: a correlated refusal prints on both.
			for _, row := range rows {
				if strings.Contains(row.Status, "42") || strings.Contains(row.Status, "53") {
					t.Errorf("%s row %q prints a sensitive value", row.Name, row.Status)
				}
			}
		})
	}
}

// secretBelowSettingRegistry declares the two settings of
// [settingsDoctorRowsRegistry] and a correlated rule that the secret
// stays below the ordinary setting, quoting both values as every
// production rule does.
func secretBelowSettingRegistry() *config.Registry {
	reg := settingsDoctorRowsRegistry()
	secret, other := settingsDoctorRowsSensitive.Key, settingsDoctorRowsTestSetting.Key
	reg.AddCorrelatedRule(config.CorrelatedRule{
		Name: "secret below the setting",
		Keys: []string{secret, other},
		Check: func(v map[string]string) error {
			s, err := strconv.Atoi(v[secret])
			if err != nil {
				return err
			}
			o, err := strconv.Atoi(v[other])
			if err != nil {
				return err
			}
			if s >= o {
				return fmt.Errorf("%s %q is not below %s %q", secret, v[secret], other, v[other])
			}
			return nil
		},
	})
	return reg
}

// TestStoreRowsAgreeOnVouch pins that the store row and the settings
// rows read one verdict per store state: the store row fails exactly
// the states whose settings rows read "not checked", an unrecognized
// state among them. Every report carries facts, so a state the
// settings rows wrongly vouched for would evaluate them instead.
func TestStoreRowsAgreeOnVouch(t *testing.T) {
	states := []store.State{
		store.StateAbsent, store.StateRestoreInProgress, store.StateBusy, store.StateNotAStore,
		store.StateRefused, store.StateOlder, store.StateNewer, store.StateCurrent, store.State(99),
	}
	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			report := store.Report{State: state, Facts: &store.Facts{}}
			storeFailed := report.Row().Failed
			rows := DoctorRows(nil, settingsDoctorRowsRegistry(), report, Facts{})
			if len(rows) == 0 {
				t.Fatal("DoctorRows rendered no rows")
			}
			for _, row := range rows {
				notChecked := strings.Contains(row.Status, "store not checked")
				if notChecked != storeFailed {
					t.Errorf("state %d: store row failed = %v, but %s row = %q", state, storeFailed, row.Name, row.Status)
				}
			}
		})
	}
}

// TestDoctorRows_LocalKeep: a stored env_seed value of "2" for
// DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP renders "2" under store.Register's
// registry, matching the environment, so the row carries no change
// note.
func TestDoctorRows_LocalKeep(t *testing.T) {
	reg := config.NewRegistry()
	store.Register(reg)
	key := store.LocalSnapshotKeepSetting.Key

	report := store.Report{State: store.StateCurrent, Facts: &store.Facts{}}
	facts := Facts{Rows: []Row{{Key: key, Value: "2", Source: "env_seed", SeedValue: "2"}}}
	rows := DoctorRows([]string{key + "=2"}, reg, report, facts)
	want := doctor.Row{Name: key, Status: `"2" (source env_seed, next operation)`}
	if i := slices.IndexFunc(rows, func(r doctor.Row) bool { return r.Name == key }); i < 0 || rows[i] != want {
		t.Errorf("rows = %+v, want one %+v", rows, want)
	}
}

// TestDoctorRows_CorrelatedRule: a stored administrator's retention of
// 30m stays in force beside an expiry the environment raises to 45m, so
// the post-seed pair breaks the retention rule, and each key's row
// refuses naming both, not only the joined evaluation error.
func TestDoctorRows_CorrelatedRule(t *testing.T) {
	report := store.Report{State: store.StateCurrent, Facts: &store.Facts{}}
	facts := Facts{Rows: []Row{
		{Key: keyExpiry, Value: "15m", Source: "env_seed", SeedValue: "15m"},
		{Key: keyNonce, Value: "30m", Source: "admin", SeedValue: "24h"},
	}}

	rows := DoctorRows([]string{keyExpiry + "=45m"}, correlatedRegistry(), report, facts)

	for _, key := range []string{keyExpiry, keyNonce} {
		i := slices.IndexFunc(rows, func(r doctor.Row) bool { return r.Name == key })
		if i < 0 {
			t.Fatalf("rows carry no entry for %s: %+v", key, rows)
		}
		row := rows[i]
		if !row.Failed || !strings.Contains(row.Status, "refused:") || !strings.Contains(row.Status, keyExpiry) || !strings.Contains(row.Status, keyNonce) {
			t.Errorf("%s: row = %+v, want a failed refusal naming both %s and %s", key, row, keyExpiry, keyNonce)
		}
	}
}
