package settings

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/store"
)

// TestEvaluateFacts_ConvertsStoreRows: EvaluateFacts against a facts
// value holding a seeded row resolves that value, the same as
// [Evaluate] against the same row through a handle.
func TestEvaluateFacts_ConvertsStoreRows(t *testing.T) {
	reg := newRegistry()
	facts := store.Facts{Settings: []store.SettingsRow{
		{Key: keyKeep, Value: "20", Source: "env-seed", SeedValue: "20"},
	}}

	ev := EvaluateFacts(reg, []string{keyKeep + "=20"}, facts)

	if ev.Err != nil {
		t.Fatalf("EvaluateFacts: %v", ev.Err)
	}
	var got Outcome
	found := false
	for _, o := range ev.Outcomes {
		if o.Key == keyKeep {
			got, found = o, true
		}
	}
	if !found {
		t.Fatalf("outcomes = %+v, want one for %s", ev.Outcomes, keyKeep)
	}
	if got.Value != "20" || got.Stored == nil || got.Stored.Source != "env-seed" {
		t.Errorf("outcome = %+v, want value 20 from an env-seed row", got)
	}
}

// TestEvaluateFacts_TableAbsent: a facts value marked table absent is
// evaluated against an empty row set, so the environment's value is
// compared and bounded exactly as serve's first seed into that table
// would compare it. Fixture state: keyKeep's registered default is 14
// ([newRegistry]), distinct from every environment value below, so a
// row that skipped the environment would read 14.
func TestEvaluateFacts_TableAbsent(t *testing.T) {
	tests := []struct {
		name       string
		env        []string
		wantValue  string
		wantChange Change
		wantErr    []string
	}{
		{
			name:      "no environment value: the default",
			wantValue: "14",
		},
		{
			name:       "an in-bounds environment value: seeds",
			env:        []string{keyKeep + "=20"},
			wantValue:  "20",
			wantChange: Seed,
		},
		{
			name:    "an environment value above the ceiling: refused naming the key and the bound",
			env:     []string{keyKeep + "=5000"},
			wantErr: []string{keyKeep, "above the ceiling 1000"},
		},
		{
			name:    "an environment value that does not parse: refused naming the accepted form",
			env:     []string{keyKeep + "=many"},
			wantErr: []string{keyKeep, `invalid value "many"`, "an integer"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newRegistry()
			ev := EvaluateFacts(reg, tt.env, store.Facts{SettingsAbsent: true})

			if len(ev.Outcomes) != len(reg.Settings()) {
				t.Fatalf("outcomes = %d, want %d, one per registered setting", len(ev.Outcomes), len(reg.Settings()))
			}
			var o Outcome
			for _, oc := range ev.Outcomes {
				if oc.Key == keyKeep {
					o = oc
				}
			}
			if o.Stored != nil {
				t.Errorf("Stored = %+v, want nil: the table holds no row", o.Stored)
			}
			if tt.wantErr != nil {
				wantErrText(t, ev.Err, tt.wantErr...)
				wantErrText(t, o.Err, tt.wantErr...)
				return
			}
			if ev.Err != nil {
				t.Fatalf("Err = %v, want nil", ev.Err)
			}
			if o.Value != tt.wantValue || o.Change != tt.wantChange {
				t.Errorf("outcome = value %q change %v, want value %q change %v", o.Value, o.Change, tt.wantValue, tt.wantChange)
			}
		})
	}
}

// TestEvaluate_CorrelatedRuleRefusesBothRows proves the per-key half of
// Task 5's doctor table: a correlated rule's refusal marks both keys'
// own outcomes refused, naming both, not only the joined evaluation
// error.
func TestEvaluate_CorrelatedRuleRefusesBothRows(t *testing.T) {
	reg := correlatedRegistry()
	rows := []Row{
		{Key: keyExpiry, Value: "72h", Source: "env-seed", SeedValue: "72h"},
		{Key: keyNonce, Value: "48h", Source: "env-seed", SeedValue: "48h"},
	}

	ev := evaluate(reg, []string{keyExpiry + "=72h", keyNonce + "=48h"}, rows)

	if ev.Err == nil {
		t.Fatal("Err = nil, want the correlated rule's refusal")
	}
	for _, key := range []string{keyExpiry, keyNonce} {
		var o Outcome
		found := false
		for _, oc := range ev.Outcomes {
			if oc.Key == key {
				o, found = oc, true
			}
		}
		if !found {
			t.Fatalf("no outcome for %s", key)
		}
		if o.Err == nil {
			t.Errorf("%s: Err = nil, want the correlated rule's refusal on its own row", key)
		}
		wantErrText(t, o.Err, keyExpiry, keyNonce, "shorter than")
	}
}

// TestEvaluate_UnparseableRowWrapsParseError: an admin row of an int
// setting whose value does not parse is refused with an error that
// wraps the parse failure, so a caller can branch on it, and whose
// text still names the key and the accepted form.
func TestEvaluate_UnparseableRowWrapsParseError(t *testing.T) {
	reg := newRegistry()
	facts := store.Facts{Settings: []store.SettingsRow{
		{Key: keyKeep, Value: "many", Source: "admin", SeedValue: "20"},
	}}

	ev := EvaluateFacts(reg, nil, facts)

	if !errors.Is(ev.Err, strconv.ErrSyntax) {
		t.Errorf("Err = %v, want it to wrap strconv.ErrSyntax", ev.Err)
	}
	wantErrText(t, ev.Err, keyKeep, `invalid value "many"`, "an integer")
}

// TestEvaluate_AdminRefusalsNameTheResetVerb: each refusal the seed
// rule raises over an admin row names the verb that clears it, and the
// parse refusal over an env-seed row, which a reset leaves in place,
// does not. Fixture: one populated row per case, on a registry whose
// keep setting the row's value refuses.
func TestEvaluate_AdminRefusalsNameTheResetVerb(t *testing.T) {
	verb := "settings reset " + keyKeep
	tests := []struct {
		name     string
		row      Row
		environ  []string
		wantVerb bool
		wantText string
	}{
		{
			name:     "an admin value that does not parse",
			row:      Row{Key: keyKeep, Value: "many", Source: "admin", SeedValue: "20"},
			wantVerb: true,
			wantText: `invalid value "many"`,
		},
		{
			name:     "an admin value the environment contradicts",
			row:      Row{Key: keyKeep, Value: "30", Source: "admin", SeedValue: "20"},
			environ:  []string{keyKeep + "=40"},
			wantVerb: true,
			wantText: "unset " + keyKeep,
		},
		{
			name:     "an admin value outside its bounds",
			row:      Row{Key: keyKeep, Value: "2", Source: "admin", SeedValue: "20"},
			wantVerb: true,
			wantText: "below the floor 3",
		},
		{
			name:     "an env-seed value that does not parse",
			row:      Row{Key: keyKeep, Value: "many", Source: "env-seed", SeedValue: "many"},
			wantText: `invalid value "many"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := evaluate(newRegistry(), tt.environ, []Row{tt.row})

			wantErrText(t, ev.Err, keyKeep, tt.wantText)
			if got := strings.Contains(ev.Err.Error(), verb); got != tt.wantVerb {
				t.Errorf("err = %q, names %q = %v, want %v", ev.Err, verb, got, tt.wantVerb)
			}
		})
	}
}
