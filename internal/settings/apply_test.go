package settings

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// fixtureGeneration is the generation every seed-table fixture starts
// at, above 0 so a reset to 0 would show.
const fixtureGeneration = 5

// TestApply_SeedTable runs the apply over every stored-row and
// environment state the seed rule's closed table names, one row each.
// A refused row must leave the fixture's rows, the audit table, and
// the generation exactly as they were, and no row may add or remove an
// admin row.
func TestApply_SeedTable(t *testing.T) {
	micros := testNow.UnixMicro()
	audit := func(kind, key, oldVal, newVal string) auditRow {
		return auditRow{at: micros, kind: kind, actor: "system", channel: "env-seed", key: key, oldVal: oldVal, newVal: newVal}
	}
	envSeed := func(key, value string) fixtureRow { return fixtureRow{key, value, "env-seed", value} }
	admin := func(key, value, seed string) fixtureRow { return fixtureRow{key, value, "admin", seed} }

	tests := []struct {
		name    string
		fixture []fixtureRow
		env     []string
		// wantErr is nil for an apply that succeeds, else every part
		// its error must name.
		wantErr []string
		// wantRows is the settings table afterward; a refused row
		// leaves the fixture, which the test checks without it.
		wantRows  []storedRow
		wantAudit []auditRow
		wantBump  bool
		wantLog   []string
		// wantKeep, when non-zero, is what an accessor reads for the
		// keep setting afterward.
		wantKeep int
	}{
		{
			name:     "no row, env absent: no row, the default reads",
			wantKeep: 14,
		},
		{
			name:      "no row, env present: seeds an env-seed row",
			env:       []string{keyKeep + "=20"},
			wantRows:  []storedRow{{keyKeep, "20", "env-seed", "20"}},
			wantAudit: []auditRow{audit("seed", keyKeep, "<NULL>", "20")},
			wantBump:  true,
			wantKeep:  20,
		},
		{
			name:     "env-seed row, env equal by typed value: nothing changes",
			fixture:  []fixtureRow{envSeed(keyExpiry, "1h")},
			env:      []string{keyExpiry + "=60m"},
			wantRows: []storedRow{{keyExpiry, "1h", "env-seed", "1h"}},
		},
		{
			name:      "env-seed row, env different: re-seeded and logged",
			fixture:   []fixtureRow{envSeed(keyKeep, "20")},
			env:       []string{keyKeep + "=30"},
			wantRows:  []storedRow{{keyKeep, "30", "env-seed", "30"}},
			wantAudit: []auditRow{audit("reseed", keyKeep, "20", "30")},
			wantBump:  true,
			wantLog:   []string{"setting reseed", keyKeep, `\"20\"`, `\"30\"`},
			wantKeep:  30,
		},
		{
			name:      "env-seed row, env absent: removed to the default and logged",
			fixture:   []fixtureRow{envSeed(keyKeep, "20")},
			wantAudit: []auditRow{audit("remove", keyKeep, "20", "<NULL>")},
			wantBump:  true,
			wantLog:   []string{"setting remove", keyKeep},
			wantKeep:  14,
		},
		{
			name:      "env-seed row, env empty: removed as if absent",
			fixture:   []fixtureRow{envSeed(keyKeep, "20")},
			env:       []string{keyKeep + "="},
			wantAudit: []auditRow{audit("remove", keyKeep, "20", "<NULL>")},
			wantBump:  true,
			wantKeep:  14,
		},
		{
			name:     "admin row, env equal to its value: nothing changes",
			fixture:  []fixtureRow{admin(keyKeep, "30", "20")},
			env:      []string{keyKeep + "=30"},
			wantRows: []storedRow{{keyKeep, "30", "admin", "20"}},
			wantKeep: 30,
		},
		{
			name:    "admin row, env equal to its seed value: refused",
			fixture: []fixtureRow{admin(keyKeep, "30", "20")},
			env:     []string{keyKeep + "=20"},
			wantErr: []string{keyKeep, `seeds "20"`, `administrator set "30"`, "unset " + keyKeep},
		},
		{
			name:     "admin row, env absent: the admin value stands",
			fixture:  []fixtureRow{admin(keyKeep, "30", "20")},
			wantRows: []storedRow{{keyKeep, "30", "admin", "20"}},
			wantKeep: 30,
		},
		{
			name:      "env-seed row out of bounds, env in bounds: healed by re-seed",
			fixture:   []fixtureRow{envSeed(keyKeep, "2")},
			env:       []string{keyKeep + "=20"},
			wantRows:  []storedRow{{keyKeep, "20", "env-seed", "20"}},
			wantAudit: []auditRow{audit("reseed", keyKeep, "2", "20")},
			wantBump:  true,
			wantKeep:  20,
		},
		{
			name:      "env-seed row out of bounds, env absent: healed by removal",
			fixture:   []fixtureRow{envSeed(keyKeep, "2")},
			wantAudit: []auditRow{audit("remove", keyKeep, "2", "<NULL>")},
			wantBump:  true,
			wantKeep:  14,
		},
		{
			name:    "env-seed row out of bounds, env equal to it: refused naming the bound",
			fixture: []fixtureRow{envSeed(keyKeep, "2")},
			env:     []string{keyKeep + "=2"},
			wantErr: []string{keyKeep, "below the floor 3"},
		},
		{
			name:    "admin row out of bounds, env absent: refused naming the bound",
			fixture: []fixtureRow{admin(keyKeep, "2", "20")},
			wantErr: []string{keyKeep, "below the floor 3"},
		},
		{
			name:    "admin row out of bounds, env different: refused naming the bound",
			fixture: []fixtureRow{admin(keyKeep, "2", "20")},
			env:     []string{keyKeep + "=5"},
			wantErr: []string{keyKeep, "below the floor 3"},
		},
		{
			name:     "a row for an unregistered key: left alone",
			fixture:  []fixtureRow{envSeed(keyOrphan, "7")},
			env:      []string{keyOrphan + "=8"},
			wantRows: []storedRow{{keyOrphan, "7", "env-seed", "7"}},
		},
		{
			name:    "env value outside the bounds, no row: refused naming the bound",
			env:     []string{keyKeep + "=2"},
			wantErr: []string{keyKeep, "below the floor 3"},
		},
		{
			name:    "env value that does not parse: refused naming the reason",
			env:     []string{keyKeep + "=many"},
			wantErr: []string{keyKeep, `invalid value "many"`, "an integer"},
		},
		{
			name:    "stored value that does not parse, env absent: refused naming the reason",
			fixture: []fixtureRow{envSeed(keyKeep, "many")},
			wantErr: []string{keyKeep, `invalid value "many"`, "an integer"},
		},
		{
			name:    "two keys refusing for different reasons: both named",
			env:     []string{keyKeep + "=many", keyRetention + "=1h"},
			wantErr: []string{keyKeep, `invalid value "many"`, keyRetention, "below the floor 24h"},
		},
		{
			name:      "a later env entry wins over an earlier one",
			env:       []string{keyKeep + "=20", keyKeep + "=40"},
			wantRows:  []storedRow{{keyKeep, "40", "env-seed", "40"}},
			wantAudit: []auditRow{audit("seed", keyKeep, "<NULL>", "40")},
			wantBump:  true,
			wantKeep:  40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h, tt.fixture...)
			fixtureRows := readSettings(t, h)
			adminBefore := countAdmin(t, h)
			reg := newRegistry()
			log, buf := bufferLog()

			err := Apply(ctxFor(t), h, reg, tt.env, testNow, log)

			wantRows, wantGen := tt.wantRows, int64(fixtureGeneration)
			if tt.wantErr != nil {
				wantErrText(t, err, tt.wantErr...)
				wantRows = fixtureRows
			} else if err != nil {
				t.Fatalf("apply: %v", err)
			}
			if tt.wantBump {
				wantGen++
			}
			if got := readSettings(t, h); !slices.Equal(got, wantRows) {
				t.Errorf("settings = %v, want %v", got, wantRows)
			}
			if got := readAudit(t, h); !slices.Equal(got, tt.wantAudit) {
				t.Errorf("audit = %v, want %v", got, tt.wantAudit)
			}
			if got := readGeneration(t, h); got != wantGen {
				t.Errorf("generation = %d, want %d", got, wantGen)
			}
			if got := countAdmin(t, h); got != adminBefore {
				t.Errorf("admin rows = %d, want the fixture's %d", got, adminBefore)
			}
			for _, part := range tt.wantLog {
				if !strings.Contains(buf.String(), part) {
					t.Errorf("log = %q, want it to hold %q", buf.String(), part)
				}
			}
			if tt.wantErr != nil && buf.Len() != 0 {
				t.Errorf("refused apply logged %q, want nothing", buf.String())
			}
			if tt.wantKeep != 0 {
				got, err := NewAccessor(h, reg).Int(context.Background(), keepSetting("14"))
				if err != nil {
					t.Fatalf("read keep: %v", err)
				}
				if got != tt.wantKeep {
					t.Errorf("keep reads %d, want %d", got, tt.wantKeep)
				}
			}
		})
	}
}

// TestEvaluate_ReportsOrphans proves the evaluation lists a row no
// registered setting claims, while the apply above leaves it alone.
func TestEvaluate_ReportsOrphans(t *testing.T) {
	h := newStore(t)
	writeFixture(t, h, fixtureRow{keyOrphan, "7", "env-seed", "7"}, fixtureRow{keyKeep, "20", "env-seed", "20"})

	ev, err := Evaluate(context.Background(), h, newRegistry(), []string{keyKeep + "=20"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if ev.Err != nil {
		t.Fatalf("Evaluate refused: %v", ev.Err)
	}
	if len(ev.Orphans) != 1 || ev.Orphans[0].Key != keyOrphan || ev.Orphans[0].Value != "7" {
		t.Errorf("Orphans = %+v, want the one %s row with value 7", ev.Orphans, keyOrphan)
	}
	var keys []string
	for _, o := range ev.Outcomes {
		keys = append(keys, o.Key)
	}
	if want := []string{keyKeep, keyRetention, keyExpiry}; !slices.Equal(keys, want) {
		t.Errorf("outcome keys = %v, want %v in declaration order", keys, want)
	}
}

// TestEvaluate_WritesNothing proves the exported evaluation, meeting
// an environment that would change every row, leaves the store as it
// found it and reports each pending change.
func TestEvaluate_WritesNothing(t *testing.T) {
	h := newStore(t)
	writeFixture(t, h, fixtureRow{keyKeep, "20", "env-seed", "20"})
	before := readSettings(t, h)

	ev, err := Evaluate(context.Background(), h, newRegistry(), []string{keyRetention + "=48h"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if ev.Err != nil {
		t.Fatalf("Evaluate refused: %v", ev.Err)
	}
	got := map[string]Change{}
	for _, o := range ev.Outcomes {
		got[o.Key] = o.Change
	}
	if got[keyKeep] != Remove || got[keyRetention] != Seed {
		t.Errorf("changes = %v, want %s remove and %s seed", got, keyKeep, keyRetention)
	}
	if after := readSettings(t, h); !slices.Equal(after, before) {
		t.Errorf("settings = %v after Evaluate, want %v", after, before)
	}
	if n := len(readAudit(t, h)); n != 0 {
		t.Errorf("audit rows = %d after Evaluate, want 0", n)
	}
	if g := readGeneration(t, h); g != fixtureGeneration {
		t.Errorf("generation = %d after Evaluate, want %d", g, fixtureGeneration)
	}
}

// TestEvaluate_ShowsSensitiveValues proves a sensitive setting's
// report prints both values: the mark gates a write, never a read.
func TestEvaluate_ShowsSensitiveValues(t *testing.T) {
	reg := config.NewRegistry()
	s := keepSetting("14")
	s.Mark = config.Sensitive
	reg.AddSetting(s)

	ev := evaluate(reg, []string{keyKeep + "=731"}, []Row{{Key: keyKeep, Value: "529", Source: "env-seed", SeedValue: "529"}})
	if len(ev.Outcomes) != 1 || ev.Outcomes[0].Change != Reseed {
		t.Fatalf("outcomes = %+v, want one reseed", ev.Outcomes)
	}
	if r := ev.Outcomes[0].Report; !strings.Contains(r, `"731"`) || !strings.Contains(r, `"529"`) {
		t.Errorf("report = %q, want both values printed", r)
	}
}

// answerSetting is a closed-set string setting shaped like the
// unknown-address answer.
var answerSetting = config.Setting{
	Key: "DUBPLATE_TEST_ANSWER", Kind: config.KindString, Default: "kind", Doc: "one of kind or generic",
	Apply: config.ApplyNextOperation, Values: []string{"kind", "generic"},
	Description: testDescription, Mark: config.NotSensitive,
}

// TestApply_StringSetting runs the seed rule over a closed-set string
// setting. The equal-environment row is the one a second serve start
// meets after a first start seeded the value: the stored row must
// parse and compare, or that second start refuses.
func TestApply_StringSetting(t *testing.T) {
	const key = "DUBPLATE_TEST_ANSWER"
	tests := []struct {
		name     string
		fixture  []fixtureRow
		env      []string
		wantErr  []string
		wantRows []storedRow
	}{
		{
			name:     "no row, env in the set: seeded",
			env:      []string{key + "=generic"},
			wantRows: []storedRow{{key, "generic", "env-seed", "generic"}},
		},
		{
			name:     "env-seed row, env equal: unchanged, no refusal",
			fixture:  []fixtureRow{{key, "generic", "env-seed", "generic"}},
			env:      []string{key + "=generic"},
			wantRows: []storedRow{{key, "generic", "env-seed", "generic"}},
		},
		{
			name:     "env-seed row, env the other value: re-seeded",
			fixture:  []fixtureRow{{key, "generic", "env-seed", "generic"}},
			env:      []string{key + "=kind"},
			wantRows: []storedRow{{key, "kind", "env-seed", "kind"}},
		},
		{
			name:    "env outside the set: refused naming the set",
			env:     []string{key + "=open"},
			wantErr: []string{key, `"open"`, "not one of kind, generic"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h, tt.fixture...)
			reg := config.NewRegistry()
			reg.AddSetting(answerSetting)
			before := readSettings(t, h)

			err := Apply(ctxFor(t), h, reg, tt.env, testNow, discardLog)

			if tt.wantErr != nil {
				wantErrText(t, err, tt.wantErr...)
				if got := readSettings(t, h); !slices.Equal(got, before) {
					t.Errorf("settings = %v after a refusal, want %v", got, before)
				}
				return
			}
			if err != nil {
				t.Fatalf("apply: %v", err)
			}
			if got := readSettings(t, h); !slices.Equal(got, tt.wantRows) {
				t.Errorf("settings = %v, want %v", got, tt.wantRows)
			}
		})
	}
}

// Two duration settings shaped as the nonce-retention rule: the
// retention must cover the link expiry. Expiry is registered first,
// so a check run after each key in declaration order would see the new
// expiry beside the old retention.
var (
	expirySetting = config.Setting{
		Key: keyExpiry, Kind: config.KindDuration, Default: "15m", Doc: "a Go duration",
		Apply: config.ApplyNextOperation, Floor: "1m", Ceiling: "87600h",
		Description: testDescription, Mark: config.NotSensitive,
	}
	nonceSetting = config.Setting{
		Key: keyNonce, Kind: config.KindDuration, Default: "48h", Doc: "a Go duration",
		Apply: config.ApplyNextOperation, Floor: "1m", Ceiling: "87600h",
		Description: testDescription, Mark: config.NotSensitive,
	}
)

// correlatedRegistry declares expiry, then retention, and the rule
// that retention is at least the expiry.
func correlatedRegistry() *config.Registry {
	reg := config.NewRegistry()
	reg.AddSetting(expirySetting)
	reg.AddSetting(nonceSetting)
	reg.AddCorrelatedRule(config.CorrelatedRule{
		Name: "nonce retention covers link expiry",
		Keys: []string{keyExpiry, keyNonce},
		Check: func(v map[string]string) error {
			expiry, err := time.ParseDuration(v[keyExpiry])
			if err != nil {
				return err
			}
			retention, err := time.ParseDuration(v[keyNonce])
			if err != nil {
				return err
			}
			if retention < expiry {
				return fmt.Errorf("%s %s is shorter than %s %s", keyNonce, v[keyNonce], keyExpiry, v[keyExpiry])
			}
			return nil
		},
	})
	return reg
}

// TestApply_CorrelatedRule checks the rule against the values in force
// after the whole seed. Both rows are seeded and both env values are
// set, so the failing case's only fault is the post-seed pair.
func TestApply_CorrelatedRule(t *testing.T) {
	tests := []struct {
		name     string
		env      []string
		wantErr  []string
		wantRows []storedRow
	}{
		{
			name: "both rise together: applies",
			env:  []string{keyExpiry + "=72h", keyNonce + "=96h"},
			wantRows: []storedRow{
				{keyExpiry, "72h", "env-seed", "72h"},
				{keyNonce, "96h", "env-seed", "96h"},
			},
		},
		{
			name:    "expiry rises past the retention: refused naming both",
			env:     []string{keyExpiry + "=72h", keyNonce + "=48h"},
			wantErr: []string{keyNonce, keyExpiry, "shorter than"},
			wantRows: []storedRow{
				{keyExpiry, "15m", "env-seed", "15m"},
				{keyNonce, "48h", "env-seed", "48h"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h,
				fixtureRow{keyNonce, "48h", "env-seed", "48h"},
				fixtureRow{keyExpiry, "15m", "env-seed", "15m"})

			err := Apply(ctxFor(t), h, correlatedRegistry(), tt.env, testNow, discardLog)

			if tt.wantErr != nil {
				wantErrText(t, err, tt.wantErr...)
			} else if err != nil {
				t.Fatalf("apply: %v", err)
			}
			if got := readSettings(t, h); !slices.Equal(got, tt.wantRows) {
				t.Errorf("settings = %v, want %v", got, tt.wantRows)
			}
			if tt.wantErr != nil {
				if n := len(readAudit(t, h)); n != 0 {
					t.Errorf("audit rows = %d, want 0", n)
				}
			}
		})
	}
}

// TestApply_RefusedWritesNoAudit runs an apply in which one key would
// seed while another refuses: the refusal must discard the seed's row
// and its audit row, and leave the generation alone.
func TestApply_RefusedWritesNoAudit(t *testing.T) {
	h := newStore(t)
	writeFixture(t, h, fixtureRow{keyKeep, "30", "admin", "20"})

	err := Apply(ctxFor(t), h, newRegistry(), []string{keyKeep + "=20", keyRetention + "=48h"}, testNow, discardLog)

	wantErrText(t, err, keyKeep, `administrator set "30"`)
	if n := len(readAudit(t, h)); n != 0 {
		t.Errorf("audit rows = %d after a refused apply, want 0", n)
	}
	if got := readSettings(t, h); len(got) != 1 || got[0].key != keyKeep {
		t.Errorf("settings = %v, want only the admin row", got)
	}
	if g := readGeneration(t, h); g != fixtureGeneration {
		t.Errorf("generation = %d, want %d", g, fixtureGeneration)
	}
}

// TestApply_OneAuditRowPerSeed proves a seed writes exactly one audit
// row, by the system actor on the env-seed channel.
func TestApply_OneAuditRowPerSeed(t *testing.T) {
	h := newStore(t)
	if err := Apply(ctxFor(t), h, newRegistry(), []string{keyKeep + "=20"}, testNow, discardLog); err != nil {
		t.Fatalf("apply: %v", err)
	}
	got := readAudit(t, h)
	if len(got) != 1 {
		t.Fatalf("audit rows = %d, want 1", len(got))
	}
	if got[0].actor != "system" || got[0].channel != "env-seed" {
		t.Errorf("audit actor, channel = %q, %q, want system, env-seed", got[0].actor, got[0].channel)
	}
}

// TestApply_DefaultStoresNoRow proves a key at its default has no row,
// so a later binary's changed default reaches the store: an accessor
// under default 14 reads 14, and after reopening under default 20 over
// the same file it reads 20, with the table still empty.
func TestApply_DefaultStoresNoRow(t *testing.T) {
	dir := newStateDir(t)

	for _, def := range []int{14, 20} {
		h, err := store.OpenWrite(context.Background(), dir, testBusy)
		if err != nil {
			t.Fatalf("OpenWrite: %v", err)
		}
		reg := config.NewRegistry()
		s := keepSetting(fmt.Sprint(def))
		reg.AddSetting(s)
		if err := Apply(ctxFor(t), h, reg, nil, testNow, discardLog); err != nil {
			t.Fatalf("apply under default %d: %v", def, err)
		}
		got, err := NewAccessor(h, reg).Int(context.Background(), s)
		if err != nil {
			t.Fatalf("read under default %d: %v", def, err)
		}
		if got != def {
			t.Errorf("read under default %d = %d", def, got)
		}
		if rows := readSettings(t, h); len(rows) != 0 {
			t.Errorf("settings under default %d = %v, want no row", def, rows)
		}
		if err := h.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	}
}
