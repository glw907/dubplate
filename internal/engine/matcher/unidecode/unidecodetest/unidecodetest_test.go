package unidecodetest

import (
	"fmt"
	"strings"
	"testing"
)

// recordingTB stands in for a *testing.T: Fatalf records its message
// and panics, as a real Fatalf stops the goroutine.
type recordingTB struct {
	testing.TB
	msg string
}

type stopped struct{}

func (r *recordingTB) Helper() {}

func (r *recordingTB) Fatalf(format string, args ...any) {
	r.msg = fmt.Sprintf(format, args...)
	panic(stopped{})
}

func TestLoadCachedTable(t *testing.T) {
	table := LoadCachedTable(t)
	got, ok := table.Lookup('é')
	if !ok || got != "e" {
		t.Fatalf("Lookup(é) = %q, %v, want \"e\", true", got, ok)
	}
}

func TestLoadCachedTableMissingCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	rec := &recordingTB{TB: t}
	func() {
		defer func() {
			if r := recover(); r != (stopped{}) {
				t.Fatalf("recover() = %v, want the fake's stop", r)
			}
		}()
		LoadCachedTable(rec)
	}()
	for _, want := range []string{"unidecode:", "make unidecode-wheel"} {
		if !strings.Contains(rec.msg, want) {
			t.Errorf("fatal message = %q, want it to contain %q", rec.msg, want)
		}
	}
}
