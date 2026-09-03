package mbrequest

import "testing"

func TestCanonicalizeQuerySortsPairsAndEscapesPlus(t *testing.T) {
	got := CanonicalizeQuery("/ws/2/release/x", map[string]string{
		"fmt": "json",
		"inc": "artists+media",
	})
	want := "/ws/2/release/x?fmt=json&inc=artists%2Bmedia"
	if got != want {
		t.Errorf("CanonicalizeQuery = %q, want %q", got, want)
	}
}

func TestCanonicalizeQueryOrderIsDeterministic(t *testing.T) {
	params := map[string]string{"b": "2", "a": "1", "c": "3"}
	first := CanonicalizeQuery("/x", params)
	for range 10 {
		if got := CanonicalizeQuery("/x", params); got != first {
			t.Errorf("CanonicalizeQuery is not deterministic across map iteration: got %q, want %q", got, first)
		}
	}
}
