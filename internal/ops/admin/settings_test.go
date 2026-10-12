package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ops"
)

// keyLimit is a third test setting, never stored, so its listing reads
// the default.
const keyLimit = "DUBPLATE_TEST_LIMIT"

var limitSetting = config.Setting{
	Key: keyLimit, Kind: config.KindInt, Default: "5", Doc: "an integer",
	Apply: config.ApplyImmediately, Floor: "1", Ceiling: "50",
	Description: "how many test things at once", Mark: config.NotSensitive,
}

// fullRegistry declares a key of every family: rung 1's wall rows, an
// identity source with its selector, an integration class, a
// threshold, and three settings.
func fullRegistry() *config.Registry {
	reg := config.NewRegistry()
	reg.AddWallRow(config.WallRow{Name: "DUBPLATE_TEST_SOURCE", Description: "the test source", Mark: config.Sensitive})
	reg.AddIdentitySource(config.IdentitySource{
		Name: "test", Selector: "DUBPLATE_TEST_SOURCE",
		Required: []config.KeyDecl{{Name: "DUBPLATE_TEST_SECRET", Description: "the test secret", Mark: config.Sensitive}},
		Optional: []config.KeyDecl{{Name: "DUBPLATE_TEST_SECRET_PREVIOUS", Description: "the previous test secret", Mark: config.Sensitive}},
	})
	reg.AddIntegrationClass(config.IntegrationClass{
		Name: "testsync", Rung: 6,
		Flag:     config.KeyDecl{Name: "DUBPLATE_TEST_SYNC_ENABLED", Description: "turns on the test sync", Mark: config.Sensitive},
		Required: []config.KeyDecl{{Name: "DUBPLATE_TEST_SYNC_REMOTE", Description: "where the test sync goes", Mark: config.Sensitive}},
	})
	reg.AddThreshold(config.Threshold{
		Key: "DUBPLATE_TEST_WAIT", Kind: config.KindDuration, Default: "5s", Doc: "a Go duration",
		Description: "how long a test waits", Mark: config.NotSensitive,
	})
	reg.AddSetting(lifetimeSetting)
	reg.AddSetting(keepSetting)
	reg.AddSetting(limitSetting)
	return reg
}

// TestListSettings_EveryKeyNoEnvironmentValue proves list_settings
// answers every registered key with its home, description, apply mode,
// and mark; a store-homed key also with its stored value, source, and
// seed; and that no value only the environment holds reaches the
// answer (AC9's content half).
func TestListSettings_EveryKeyNoEnvironmentValue(t *testing.T) {
	reg := fullRegistry()
	var environ []string
	for i, name := range []string{
		"DUBPLATE_LIBRARY_ROOT", "DUBPLATE_STATE_DIR", "DUBPLATE_SITE_NAME", "DUBPLATE_ADMIN_EMAIL",
		"DUBPLATE_TEST_SOURCE", "DUBPLATE_TEST_SECRET", "DUBPLATE_TEST_SECRET_PREVIOUS",
		"DUBPLATE_TEST_SYNC_ENABLED", "DUBPLATE_TEST_SYNC_REMOTE", "DUBPLATE_TEST_WAIT",
	} {
		environ = append(environ, fmt.Sprintf("%s=SENTINEL-%d", name, i))
	}
	f := newFixtureWith(t, fixtureOptions{reg: reg, env: environ})
	execSQL(t, f.h, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES (?, '2h', 'env-seed', '2h', 'system', 1)`, keyLifetime)
	execSQL(t, f.h, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES (?, '30', 'env-seed', '30', 'system', 1)`, keyKeep)
	plantAdminRow(t, f, keyKeep, "20")

	resp, err := f.seam.ListSettings(writeCtx(t), actor(alice, stale), ops.ListSettingsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "SENTINEL") {
		t.Errorf("list_settings carries an environment value: %s", raw)
	}

	str := func(s string) *string { return &s }
	env := func(name, description, mark string) ops.SettingEntry {
		return ops.SettingEntry{Name: name, Home: "environment", Description: description, Apply: "restart", Mark: mark}
	}
	want := []ops.SettingEntry{
		{Name: keyLifetime, Home: "store", Description: "how long a test session lasts", Apply: "next_operation", Mark: "sensitive",
			Value: str("2h"), Source: str("env-seed"), SeedValue: str("2h")},
		{Name: keyKeep, Home: "store", Description: "how many test snapshots to keep", Apply: "restart", Mark: "not_sensitive",
			Value: str("20"), Source: str("admin"), SeedValue: str("30")},
		{Name: keyLimit, Home: "store", Description: "how many test things at once", Apply: "immediately", Mark: "not_sensitive",
			Value: str("5"), Source: str("default")},
	}
	for _, row := range reg.OwnedWall() {
		want = append(want, env(row.Name, row.Description, markName(row.Mark)))
	}
	want = append(want,
		env("DUBPLATE_TEST_SECRET", "the test secret", "sensitive"),
		env("DUBPLATE_TEST_SECRET_PREVIOUS", "the previous test secret", "sensitive"),
		env("DUBPLATE_TEST_SYNC_ENABLED", "turns on the test sync", "sensitive"),
		env("DUBPLATE_TEST_SYNC_REMOTE", "where the test sync goes", "sensitive"),
		env("DUBPLATE_TEST_WAIT", "how long a test waits", "not_sensitive"),
	)
	if !reflect.DeepEqual(resp.Settings, want) {
		got, _ := json.MarshalIndent(resp.Settings, "", " ")
		exp, _ := json.MarshalIndent(want, "", " ")
		t.Errorf("settings =\n%s\nwant\n%s", got, exp)
	}
}

// markName is the wire form of a mark, written out by hand.
func markName(m config.Sensitivity) string {
	if m == config.Sensitive {
		return "sensitive"
	}
	return "not_sensitive"
}

// TestSetSetting_RecordsTheNetworkActor proves a cookie-mount write's
// audit row records the actor as member:<hex UserID> with channel
// cookie, and a rename between two writes leaves one actor (AC6).
func TestSetSetting_RecordsTheNetworkActor(t *testing.T) {
	f := newFixture(t)
	resp, err := f.seam.SetSetting(writeCtx(t), actor(alice, stale), ops.SetSettingRequest{Key: keyKeep, Value: "20"})
	if err != nil {
		t.Fatal(err)
	}
	if want := (ops.SetSettingResponse{Key: keyKeep, Apply: "restart", Source: "admin"}); resp != want {
		t.Errorf("response = %+v, want %+v", resp, want)
	}
	execSQL(t, f.h, `UPDATE users SET username = 'alice2' WHERE id = ?`, alice.id[:])
	renamed := alice
	renamed.username = "alice2"
	if _, err := f.seam.SetSetting(writeCtx(t), actor(renamed, stale), ops.SetSettingRequest{Key: keyKeep, Value: "30"}); err != nil {
		t.Fatal(err)
	}

	got := audits(t, f.h)
	if len(got) != 2 {
		t.Fatalf("audit rows = %d, want 2", len(got))
	}
	for i, wantNew := range []string{"20", "30"} {
		a := got[i]
		if a.kind != "set" || a.actor != "member:a1000000000000000000000000000000" || a.channel != "cookie" || a.key != keyKeep ||
			a.newValue == nil || *a.newValue != wantNew {
			t.Errorf("audit row %d = %+v, want set by member:a1000000000000000000000000000000 over cookie to %s", i, a, wantNew)
		}
	}
}

// TestResetSetting_AnswersTheApplyMode proves a reset's answer names
// the key, its apply mode, and the source in force afterwards.
func TestResetSetting_AnswersTheApplyMode(t *testing.T) {
	f := newFixture(t)
	plantAdminRow(t, f, keyLifetime, "3h")
	resp, err := f.seam.ResetSetting(writeCtx(t), actor(alice, fresh), ops.ResetSettingRequest{Key: keyLifetime})
	if err != nil {
		t.Fatal(err)
	}
	if want := (ops.ResetSettingResponse{Key: keyLifetime, Apply: "next_operation", Source: "default"}); resp != want {
		t.Errorf("response = %+v, want %+v", resp, want)
	}
}

// TestSettingWrites_Refusals covers the write path's refusals the seam
// maps, each writing nothing.
func TestSettingWrites_Refusals(t *testing.T) {
	set := func(key, value string) func(ctx context.Context, f *fixture) error {
		return func(ctx context.Context, f *fixture) error {
			_, err := f.seam.SetSetting(ctx, actor(alice, fresh), ops.SetSettingRequest{Key: key, Value: value})
			return err
		}
	}
	tests := []struct {
		name   string
		env    []string
		prep   func(t *testing.T, f *fixture)
		call   func(ctx context.Context, f *fixture) error
		code   ops.Code
		detail string
	}{
		{name: "an unregistered key", call: set("DUBPLATE_NOT_A_KEY", "1"), code: ops.CodeInvalidRequest},
		{name: "a wall key", call: set("DUBPLATE_SITE_NAME", "x"), code: ops.CodeInvalidRequest},
		{name: "a value that does not parse", call: set(keyKeep, "many"), code: ops.CodeInvalidRequest},
		{name: "a value out of bounds", call: set(keyKeep, "2"), code: ops.CodeInvalidRequest},
		{name: "a value the environment contradicts", env: []string{keyKeep + "=50"}, call: set(keyKeep, "20"),
			code: ops.CodeConflict, detail: ops.DetailEnvironmentSets},
		{
			name: "a reset of an admin row the environment sets", env: []string{keyKeep + "=20"},
			prep: func(t *testing.T, f *fixture) { plantAdminRow(t, f, keyKeep, "20") },
			call: func(ctx context.Context, f *fixture) error {
				_, err := f.seam.ResetSetting(ctx, actor(alice, fresh), ops.ResetSettingRequest{Key: keyKeep})
				return err
			},
			code: ops.CodeConflict, detail: ops.DetailEnvironmentSets,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixtureWith(t, fixtureOptions{env: tt.env})
			if tt.prep != nil {
				tt.prep(t, f)
			}
			before := snapshot(t, f.h)
			wantCode(t, tt.call(writeCtx(t), f), tt.code, tt.detail)
			if after := snapshot(t, f.h); after != before {
				t.Errorf("a refused write wrote:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}

// TestSetSetting_EqualToTheEnvironment proves a value the serving
// environment already sets writes nothing and answers the environment
// as the source in force.
func TestSetSetting_EqualToTheEnvironment(t *testing.T) {
	f := newFixtureWith(t, fixtureOptions{env: []string{keyLifetime + "=3h"}})
	before := snapshot(t, f.h)
	resp, err := f.seam.SetSetting(writeCtx(t), actor(alice, fresh), ops.SetSettingRequest{Key: keyLifetime, Value: "3h"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Source != "environment" {
		t.Errorf("source = %q, want environment", resp.Source)
	}
	if after := snapshot(t, f.h); after != before {
		t.Errorf("a no-op write wrote:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if n := len(f.mail.recipients()); n != 0 {
		t.Errorf("a no-op sensitive write sent %d notices, want 0", n)
	}
}

// TestSettingWrites_ActorNoLongerAdmin proves the seam passes the
// write path its in-write admin check: an admin-privileged actor whose
// row a direct store write demoted or removed is refused forbidden on
// a set and a reset, and nothing is written.
func TestSettingWrites_ActorNoLongerAdmin(t *testing.T) {
	changes := map[string]string{
		"demoted": `UPDATE users SET admin = 0 WHERE id = ?`,
		"removed": `UPDATE users SET tombstoned_at = 1 WHERE id = ?`,
	}
	for name, stmt := range changes {
		for _, c := range ungatedCalls {
			t.Run(name+", "+c.name, func(t *testing.T) {
				f := newFixture(t)
				if c.prep != nil {
					c.prep(t, f)
				}
				execSQL(t, f.h, stmt, alice.id[:])
				before := snapshot(t, f.h)
				wantCode(t, c.call(writeCtx(t), f, actor(alice, fresh)), ops.CodeForbidden, "")
				if after := snapshot(t, f.h); after != before {
					t.Errorf("a refused write wrote:\nbefore:\n%s\nafter:\n%s", before, after)
				}
			})
		}
	}
}

// TestSetSetting_InFlightDemotion proves the admin check runs inside
// the write: a demotion committed while the write waited for the lock
// refuses it. A check read before the write would see alice still an
// admin.
func TestSetSetting_InFlightDemotion(t *testing.T) {
	f := newFixture(t)
	err := whileHeld(t, f.secondHandle(t), `UPDATE users SET admin = 0 WHERE id = ?`, []any{alice.id[:]}, func() error {
		_, err := f.seam.SetSetting(writeCtx(t), actor(alice, fresh), ops.SetSettingRequest{Key: keyKeep, Value: "20"})
		return err
	})
	wantCode(t, err, ops.CodeForbidden, "")
	if v, _ := storedValue(t, f.h, keyKeep); v != "" {
		t.Errorf("stored %s = %q, want no row", keyKeep, v)
	}
}

// TestSettingWrite_Busy proves a settings write that finds the store
// busy answers unavailable with detail busy and writes nothing.
func TestSettingWrite_Busy(t *testing.T) {
	for _, arm := range busyArms {
		t.Run(arm.name, func(t *testing.T) {
			arm.run(t, func(ctx context.Context, f *fixture) error {
				_, err := f.seam.SetSetting(ctx, actor(alice, fresh), ops.SetSettingRequest{Key: keyKeep, Value: "20"})
				return err
			})
		})
	}
}
