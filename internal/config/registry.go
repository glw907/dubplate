package config

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Registry holds every declaration a rung contributes to the config
// wall: owned wall rows, reserved classes, integration classes, and
// thresholds, plus the superseded-name map. [NewRegistry] returns a
// fresh value already carrying rung 1's own declarations; a rung whose
// package exists registers its own from there.
//
// Registration is an explicit call rather than an init function, so
// declaration order stays visible in the caller's own code and a test
// that imports this package alone still sees rung 1's own rows
// (thresholds.go:36-39's reason, carried forward to every family here).
//
// A duplicate or malformed declaration panics rather than returning an
// error, on the precedent of the three stdlib registries that face the
// same choice: flag.(*FlagSet).Var, database/sql.Register, and
// expvar.Publish all panic on a collision, because the caller is a
// programmer who mis-declared, never an adopter who mis-configured. A
// gate-resident test constructs the production registry exactly as the
// binary does, so a collision fails the gate on the lane that
// introduced it rather than surfacing only when the binary runs.
//
// provisional, conductor ruling owed, R1
type Registry struct {
	ownedWall          []WallRow
	reservedClasses    []ReservedClass
	integrationClasses []IntegrationClass
	thresholds         []Threshold
	superseded         map[string]string

	// claims tracks every key claimed so far to the name of the
	// family that claimed it, so a key declared in a second family
	// panics naming both. Each family's entry point checks its own
	// slice for a within-family duplicate before this map ever sees
	// the key, so the two checks are independent and either can be
	// removed without disturbing the other.
	claims map[string]string
}

// NewRegistry returns a Registry seeded with rung 1's own
// declarations: the owned wall rows, the reserved classes, and the
// superseded-name map, all declared inside internal/config because
// that is rung 1's own package. Seeding here, rather than leaving the
// registry to start empty, closes the empty-wall hazard: a registry a
// test or a wiring mistake constructs without registering rung 1's
// rows would let Load require nothing and accept everything.
//
// A registry built from NewRegistry is independent of every other:
// registering into one never changes another, and the same
// declaration registered into two separately-constructed registries
// panics in neither.
//
// provisional, conductor ruling owed, R4
func NewRegistry() *Registry {
	r := &Registry{
		claims:     make(map[string]string),
		superseded: make(map[string]string),
	}
	for _, row := range ownedWall {
		r.AddWallRow(row)
	}
	for _, class := range reservedClasses {
		r.AddReservedClass(class)
	}
	for name, reason := range supersededKeys {
		r.AddSuperseded(name, reason)
	}
	return r
}

// familyOwnedWall through familySuperseded name the five declaration
// families a key can belong to, for the collision panic's message.
const (
	familyOwnedWall   = "owned wall"
	familyReserved    = "reserved classes"
	familyIntegration = "integration classes"
	familyThresholds  = "thresholds"
	familySuperseded  = "superseded names"
)

// claim panics if key was already claimed by a family other than
// family (O5 class 2), then records that family claims key. It
// deliberately does not flag a repeat claim from the same family: that
// is a within-family duplicate (O5 class 1), and each family's own
// entry point checks for it separately, so the two checks stay
// independent of each other.
func (r *Registry) claim(family, key string) {
	if existing, ok := r.claims[key]; ok && existing != family {
		panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", key, existing, family))
	}
	r.claims[key] = family
}

// claimCrossFamily panics if key does not carry the DUBPLATE_ prefix
// (O5 class 3, Resolution R16: rejectNamespace walks names by that
// prefix, so a declared key without it would never be classified and
// the declaration would silently do nothing), then claims key as
// claim does. It is claim's own prefix-enforcing variant, used by the
// three families rejectNamespace classifies by that walk: owned wall,
// reserved classes, and integration classes (and the superseded map,
// whose names share the same namespace). AddThreshold uses claim
// instead: a threshold's Key is read directly by wall.get and never
// classified by rejectNamespace's DUBPLATE_ walk, so R16's own
// justification for the prefix requirement does not reach it.
func (r *Registry) claimCrossFamily(family, key string) {
	if !strings.HasPrefix(key, "DUBPLATE_") {
		panic(fmt.Sprintf("%s: declared in %s without the DUBPLATE_ prefix", key, family))
	}
	r.claim(family, key)
}

// AddWallRow registers row as an owned wall row. A row is either a
// DUBPLATE_-prefixed key, checked against every other declared key, or
// a status row whose Pending text doctor prints in a key's place; a
// row that is neither panics (O5 class 6).
//
// The status-row carve-out exists for the identity source, whose Name
// is the prose "identity source" rather than an env key: it has not
// been assigned one yet, so the prefix rule must not reach it, and
// treating it as a key would make rejectNamespace claim "identity
// source" is owned.
//
// provisional, conductor ruling owed, R16
func (r *Registry) AddWallRow(row WallRow) {
	switch {
	case strings.HasPrefix(row.Name, "DUBPLATE_"):
		if slices.ContainsFunc(r.ownedWall, func(existing WallRow) bool { return existing.Name == row.Name }) {
			panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", row.Name, familyOwnedWall, familyOwnedWall))
		}
		r.claimCrossFamily(familyOwnedWall, row.Name)
	case row.Pending != "":
		// Status row: not a key, not claimed.
	default:
		panic(fmt.Sprintf("%s: wall row is neither a DUBPLATE_-prefixed key nor a status row with Pending text", row.Name))
	}
	r.ownedWall = append(r.ownedWall, row)
}

// OwnedWall returns the owned wall's rows in declaration order.
func (r *Registry) OwnedWall() []WallRow {
	return slices.Clone(r.ownedWall)
}

// AddReservedClass registers class's keys as reserved. Each key is
// checked against every already-reserved key and then against every
// other family, so a reserved key that is also, say, an integration
// key panics naming both classes.
func (r *Registry) AddReservedClass(class ReservedClass) {
	for i, key := range class.Keys {
		if slices.Contains(class.Keys[:i], key) {
			panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", key, familyReserved, familyReserved))
		}
	}
	for _, key := range class.Keys {
		for _, existing := range r.reservedClasses {
			if slices.Contains(existing.Keys, key) {
				panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", key, familyReserved, familyReserved))
			}
		}
		r.claimCrossFamily(familyReserved, key)
	}
	r.reservedClasses = append(r.reservedClasses, class)
}

// ReservedClasses returns the reserved classes in declaration order,
// never sorted: doctor prints them in table order, which is not rung
// order, and that order is the contract (Resolution R15).
//
// provisional, conductor ruling owed, R15
func (r *Registry) ReservedClasses() []ReservedClass {
	return slices.Clone(r.reservedClasses)
}

// AddIntegrationClass registers class as an adopter-optional
// integration, distinct from a reserved class because it is enabled by
// a flag rather than blocked until a rung lands (Resolution R2). class
// must carry a non-empty Flag that appears in neither its own Required
// nor its own Optional set; an integration whose enable flag is itself
// one of the keys it gates could never be told apart from a required
// row in an error message.
//
// Only Flag and Required claim across families here; Optional does
// not. An Optional row is accepted and never required precisely
// because it is also a registered [Threshold] (IntegrationClass's own
// doc), so the same key legitimately carries two declarations, one per
// family, and AddThreshold's own claim is what the gate checks for it
// (Resolution R5, re-expressing rung 4's ruling 1: rung 4's shadow
// class registers DUBPLATE_SHADOW_FLOOR and DUBPLATE_SHADOW_BUDGET
// this way).
//
// provisional, conductor ruling owed, R2
func (r *Registry) AddIntegrationClass(class IntegrationClass) {
	if class.Flag == "" {
		panic(fmt.Sprintf("%s: integration class declared with an empty Flag", class.Name))
	}
	if slices.Contains(class.Required, class.Flag) || slices.Contains(class.Optional, class.Flag) {
		panic(fmt.Sprintf("%s: integration class's Flag %s also appears in its own key set", class.Name, class.Flag))
	}

	keys := make([]string, 0, 1+len(class.Required)+len(class.Optional))
	keys = append(keys, class.Flag)
	keys = append(keys, class.Required...)
	keys = append(keys, class.Optional...)

	for i, key := range keys {
		if slices.Contains(keys[:i], key) {
			panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", key, familyIntegration, familyIntegration))
		}
	}
	for _, key := range keys {
		for _, existing := range r.integrationClasses {
			if key == existing.Flag || slices.Contains(existing.Required, key) || slices.Contains(existing.Optional, key) {
				panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", key, familyIntegration, familyIntegration))
			}
		}
	}

	claimed := append([]string{class.Flag}, class.Required...)
	for _, key := range claimed {
		r.claimCrossFamily(familyIntegration, key)
	}
	r.integrationClasses = append(r.integrationClasses, class)
}

// IntegrationClasses returns the integration classes in declaration
// order.
func (r *Registry) IntegrationClasses() []IntegrationClass {
	return slices.Clone(r.integrationClasses)
}

// AddThreshold registers t as a behavior tunable. t.Default must parse
// under t.Kind, and t.Kind must name a parser; both are refused here,
// at registration, rather than left for Load to discover as an
// environment-shaped parse failure (Resolution R11). This task builds
// the check's seat only: rung 1 registers no thresholds, and the kind
// that motivates the check's own test lands with Task 5.
//
// provisional, conductor ruling owed, R11
func (r *Registry) AddThreshold(t Threshold) {
	if slices.ContainsFunc(r.thresholds, func(existing Threshold) bool { return existing.Key == t.Key }) {
		panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", t.Key, familyThresholds, familyThresholds))
	}
	validateThresholdDefault(t)
	r.claim(familyThresholds, t.Key)
	r.thresholds = append(r.thresholds, t)
}

// validateThresholdDefault panics if t's Default cannot be parsed
// under its declared Kind, or if Kind names no parser: both are
// registration bugs, never an adopter's environment problem. This is
// the check that catches a threshold whose Kind was never set: KindInt
// is iota's zero (Resolution R11), so a float threshold with no
// explicit Kind carries a named kind and would pass an enum-membership
// test alone; only parsing Default under the declared Kind catches it.
//
// KindString's own invalid Default is the empty string, checked here
// rather than by a parse attempt, since every string parses; the
// reason is [ParseString]'s own doc.
//
// provisional, conductor ruling owed, R11
func validateThresholdDefault(t Threshold) {
	switch t.Kind {
	case KindInt:
		if _, err := strconv.Atoi(t.Default); err != nil {
			panic(fmt.Sprintf("%s: invalid default %q", t.Key, t.Default))
		}
	case KindDuration:
		if _, err := time.ParseDuration(t.Default); err != nil {
			panic(fmt.Sprintf("%s: invalid default %q", t.Key, t.Default))
		}
	case KindFloat:
		if _, err := strconv.ParseFloat(t.Default, 64); err != nil {
			panic(fmt.Sprintf("%s: invalid default %q", t.Key, t.Default))
		}
	case KindString:
		if t.Default == "" {
			panic(fmt.Sprintf("%s: invalid default %q", t.Key, t.Default))
		}
	default:
		panic(fmt.Sprintf("%s: unnamed threshold kind %v", t.Key, t.Kind))
	}
}

// Thresholds returns the registered thresholds in declaration order.
func (r *Registry) Thresholds() []Threshold {
	return slices.Clone(r.thresholds)
}

// AddSuperseded registers name as a pre-pivot key with no meaning in
// the greenfield wall, reason being the message rejectNamespace reports
// for it. The superseded family stays a map rather than a slice
// because no caller enumerates it: supersessionOf is a single lookup,
// and order is meaningless for one (Resolution R8).
//
// provisional, conductor ruling owed, R8
func (r *Registry) AddSuperseded(name, reason string) {
	if _, ok := r.superseded[name]; ok {
		panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", name, familySuperseded, familySuperseded))
	}
	r.claimCrossFamily(familySuperseded, name)
	r.superseded[name] = reason
}

// Superseded returns the superseded-name map, keyed by the pre-pivot
// name and valued by its disposition message.
func (r *Registry) Superseded() map[string]string {
	return maps.Clone(r.superseded)
}
