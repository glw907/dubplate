package settings

import (
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

// TestEvaluateFacts_TableAbsent: a facts value marked table absent
// resolves every key to its default and reports the store holds no
// settings yet, never comparing against the environment even when it
// sets a value.
func TestEvaluateFacts_TableAbsent(t *testing.T) {
	reg := newRegistry()
	facts := store.Facts{SettingsAbsent: true}

	ev := EvaluateFacts(reg, []string{keyKeep + "=999"}, facts)

	if ev.Err != nil {
		t.Fatalf("EvaluateFacts: %v", ev.Err)
	}
	if len(ev.Outcomes) != len(reg.Settings()) {
		t.Fatalf("outcomes = %d, want %d, one per registered setting", len(ev.Outcomes), len(reg.Settings()))
	}
	for _, o := range ev.Outcomes {
		if o.Report != "the store holds no settings yet" {
			t.Errorf("%s: report = %q, want the table-absent report", o.Key, o.Report)
		}
		if o.Change != Unchanged || o.Err != nil {
			t.Errorf("%s: change %v, err %v, want unchanged and no error", o.Key, o.Change, o.Err)
		}
	}
	// Fixture state: keyKeep's registered default is 14 ([newRegistry]),
	// distinct from the env value of 999 above, so a bug that read the
	// environment instead would read 999, not 14.
	for _, o := range ev.Outcomes {
		if o.Key == keyKeep && o.Value != "14" {
			t.Errorf("%s: value = %q, want the default 14, unread from the environment", o.Key, o.Value)
		}
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
