package config

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"
)

// panicMessage runs fn and returns the recovered panic's message,
// failing the test if fn does not panic.
func panicMessage(t *testing.T, fn func()) string {
	t.Helper()
	var msg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				msg = fmt.Sprint(r)
			}
		}()
		fn()
	}()
	if msg == "" {
		t.Fatal("want panic, got none")
	}
	return msg
}

func TestNewRegistry_CarriesRung1Declarations(t *testing.T) {
	r := NewRegistry()

	if got := r.OwnedWall(); !slicesEqualWallRow(got, ownedWall) {
		t.Errorf("OwnedWall() = %+v, want %+v", got, ownedWall)
	}
	if got := r.ReservedClasses(); !reflect.DeepEqual(got, reservedClasses) {
		t.Errorf("ReservedClasses() = %+v, want %+v", got, reservedClasses)
	}
	if got := r.Superseded(); !maps.Equal(got, supersededKeys) {
		t.Errorf("Superseded() = %+v, want %+v", got, supersededKeys)
	}
	if got := r.IntegrationClasses(); len(got) != 0 {
		t.Errorf("IntegrationClasses() = %+v, want empty: rung 1 declares none", got)
	}
	if got := r.Thresholds(); len(got) != 0 {
		t.Errorf("Thresholds() = %+v, want empty: rung 1 registers none", got)
	}
}

func slicesEqualWallRow(a, b []WallRow) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRegistry_AccessorsReturnCopy(t *testing.T) {
	r := NewRegistry()
	r.AddIntegrationClass(IntegrationClass{Name: "test", Flag: "DUBPLATE_TEST_FLAG"})
	r.AddThreshold(Threshold{Key: "DUBPLATE_TEST_THRESH", Default: "1"})

	t.Run("OwnedWall", func(t *testing.T) {
		got := r.OwnedWall()
		got[0].Name = "mutated"
		if r.ownedWall[0].Name == "mutated" {
			t.Fatal("OwnedWall() leaked the backing array")
		}
	})
	t.Run("ReservedClasses", func(t *testing.T) {
		got := r.ReservedClasses()
		got[0].Name = "mutated"
		if r.reservedClasses[0].Name == "mutated" {
			t.Fatal("ReservedClasses() leaked the backing array")
		}
	})
	t.Run("IntegrationClasses", func(t *testing.T) {
		got := r.IntegrationClasses()
		got[0].Name = "mutated"
		if r.integrationClasses[0].Name == "mutated" {
			t.Fatal("IntegrationClasses() leaked the backing array")
		}
	})
	t.Run("Thresholds", func(t *testing.T) {
		got := r.Thresholds()
		got[0].Key = "mutated"
		if r.thresholds[0].Key == "mutated" {
			t.Fatal("Thresholds() leaked the backing array")
		}
	})
	t.Run("Superseded", func(t *testing.T) {
		got := r.Superseded()
		got["DUBPLATE_LIBRARY_DIR"] = "mutated"
		if r.superseded["DUBPLATE_LIBRARY_DIR"] == "mutated" {
			t.Fatal("Superseded() leaked the backing map")
		}
	})
}

// TestCollisionGate_WithinFamily fires O5 class 1 (a key declared twice
// within one family) once per family, so mutation 1 (drop the
// duplicate check from one family's entry point) can be counted
// per-family rather than as a single pass or fail.
func TestCollisionGate_WithinFamily(t *testing.T) {
	tests := []struct {
		name string
		fn   func(r *Registry)
	}{
		{
			name: "owned wall",
			fn: func(r *Registry) {
				r.AddWallRow(WallRow{Name: "DUBPLATE_TEST_DUP"})
				r.AddWallRow(WallRow{Name: "DUBPLATE_TEST_DUP"})
			},
		},
		{
			name: "reserved classes",
			fn: func(r *Registry) {
				r.AddReservedClass(ReservedClass{Name: "test a", Rung: 99, Keys: []string{"DUBPLATE_TEST_DUP"}})
				r.AddReservedClass(ReservedClass{Name: "test b", Rung: 99, Keys: []string{"DUBPLATE_TEST_DUP"}})
			},
		},
		{
			name: "reserved classes, single class repeats its own key",
			fn: func(r *Registry) {
				r.AddReservedClass(ReservedClass{Name: "test", Rung: 99, Keys: []string{"DUBPLATE_TEST_DUP", "DUBPLATE_TEST_DUP"}})
			},
		},
		{
			name: "integration classes",
			fn: func(r *Registry) {
				r.AddIntegrationClass(IntegrationClass{Name: "test a", Flag: "DUBPLATE_TEST_FLAG_A", Required: []string{"DUBPLATE_TEST_DUP"}})
				r.AddIntegrationClass(IntegrationClass{Name: "test b", Flag: "DUBPLATE_TEST_FLAG_B", Required: []string{"DUBPLATE_TEST_DUP"}})
			},
		},
		{
			name: "integration classes, single class repeats a key across Required and Optional",
			fn: func(r *Registry) {
				r.AddIntegrationClass(IntegrationClass{Name: "test", Flag: "DUBPLATE_TEST_FLAG", Required: []string{"DUBPLATE_TEST_DUP"}, Optional: []string{"DUBPLATE_TEST_DUP"}})
			},
		},
		{
			name: "thresholds",
			fn: func(r *Registry) {
				r.AddThreshold(Threshold{Key: "DUBPLATE_TEST_DUP", Default: "1"})
				r.AddThreshold(Threshold{Key: "DUBPLATE_TEST_DUP", Default: "2"})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := panicMessage(t, func() { tt.fn(NewRegistry()) })
			if !strings.Contains(msg, "DUBPLATE_TEST_DUP") {
				t.Errorf("panic message %q does not name the offending key", msg)
			}
		})
	}
}

// TestCollisionGate_CrossFamily fires O5 class 2: a key declared in
// one family and also in another. It must fail independently of
// TestCollisionGate_WithinFamily, proving the two checks are separate
// (mutation 2).
func TestCollisionGate_CrossFamily(t *testing.T) {
	r := NewRegistry()
	r.AddReservedClass(ReservedClass{Name: "test", Rung: 99, Keys: []string{"DUBPLATE_TEST_CROSS"}})

	msg := panicMessage(t, func() {
		r.AddIntegrationClass(IntegrationClass{Name: "test", Flag: "DUBPLATE_TEST_FLAG", Required: []string{"DUBPLATE_TEST_CROSS"}})
	})
	if !strings.Contains(msg, "DUBPLATE_TEST_CROSS") {
		t.Errorf("panic message %q does not name the offending key", msg)
	}
	if !strings.Contains(msg, familyReserved) || !strings.Contains(msg, familyIntegration) {
		t.Errorf("panic message %q does not name both families", msg)
	}
}

// TestCollisionGate_MissingPrefix fires O5 class 3: a declared key
// without the DUBPLATE_ prefix (Resolution R16).
func TestCollisionGate_MissingPrefix(t *testing.T) {
	r := NewRegistry()
	msg := panicMessage(t, func() {
		r.AddReservedClass(ReservedClass{Name: "test", Rung: 99, Keys: []string{"BAD_NO_PREFIX"}})
	})
	if !strings.Contains(msg, "BAD_NO_PREFIX") {
		t.Errorf("panic message %q does not name the offending key", msg)
	}
}

// TestCollisionGate_IdentitySourceCarveOut proves R16's carve-out in
// the other direction, O5 class 6: the identity-source status row must
// not be claimed as a key. If the prefix rule wrongly reached it, this
// test reports "identity source" registered as a key.
func TestCollisionGate_IdentitySourceCarveOut(t *testing.T) {
	r := NewRegistry()
	if family, ok := r.claims["identity source"]; ok {
		t.Fatalf("identity source registered as a key (family %q); the prefix rule must not reach the status row", family)
	}
}

// TestCollisionGate_WallRowNeitherKeyNorStatus fires O5 class 6 in its
// refusing direction: a wall row that is neither a prefixed key nor a
// status row carrying Pending text.
func TestCollisionGate_WallRowNeitherKeyNorStatus(t *testing.T) {
	r := NewRegistry()
	msg := panicMessage(t, func() {
		r.AddWallRow(WallRow{Name: "not a prefixed key"})
	})
	if !strings.Contains(msg, "not a prefixed key") {
		t.Errorf("panic message %q does not name the offending row", msg)
	}
}

// TestCollisionGate_IntegrationFlag fires O5 class 5: an integration
// class with an empty Flag, or whose Flag also appears in its own
// Required or Optional set.
func TestCollisionGate_IntegrationFlag(t *testing.T) {
	tests := []struct {
		name  string
		class IntegrationClass
		want  string
	}{
		{
			name:  "empty flag",
			class: IntegrationClass{Name: "test"},
			want:  "test",
		},
		{
			name:  "flag in required",
			class: IntegrationClass{Name: "test", Flag: "DUBPLATE_TEST_FLAG", Required: []string{"DUBPLATE_TEST_FLAG"}},
			want:  "DUBPLATE_TEST_FLAG",
		},
		{
			name:  "flag in optional",
			class: IntegrationClass{Name: "test", Flag: "DUBPLATE_TEST_FLAG", Optional: []string{"DUBPLATE_TEST_FLAG"}},
			want:  "DUBPLATE_TEST_FLAG",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := panicMessage(t, func() { NewRegistry().AddIntegrationClass(tt.class) })
			if !strings.Contains(msg, tt.want) {
				t.Errorf("panic message %q does not name %q", msg, tt.want)
			}
		})
	}
}

// TestNewRegistry_Independent proves the property Task 4's retirement
// of the sync.OnceFunc rests on: a registry built from NewRegistry is
// independent of every other, and registering the same declaration
// into two separately-constructed registries panics in neither.
func TestNewRegistry_Independent(t *testing.T) {
	r1 := NewRegistry()
	r2 := NewRegistry()

	before := len(r2.OwnedWall())
	r1.AddWallRow(WallRow{Name: "DUBPLATE_TEST_ONLY_R1"})
	if got := len(r2.OwnedWall()); got != before {
		t.Fatalf("registering into r1 changed r2's OwnedWall() length: %d, want %d", got, before)
	}

	r3 := NewRegistry()
	r4 := NewRegistry()
	shared := WallRow{Name: "DUBPLATE_TEST_SHARED"}
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				t.Fatalf("registering the shared declaration into r3 panicked: %v", rec)
			}
		}()
		r3.AddWallRow(shared)
	}()
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				t.Fatalf("registering the shared declaration into r4 panicked: %v", rec)
			}
		}()
		r4.AddWallRow(shared)
	}()
}
