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
// wall: owned wall rows, reserved classes, integration classes,
// thresholds, identity sources, and store-homed settings with the
// correlated rules over them, plus the superseded-name map.
// [NewRegistry] returns a fresh value already carrying rung 1's own
// declarations; a rung whose package exists registers its own from
// there.
//
// Registration is an explicit call rather than an init function, so
// declaration order stays visible in the caller's own code and a test
// that imports this package alone still sees rung 1's own rows, the
// reason rung 1 gave for Threshold registration, carried forward to
// every family here.
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
// ruled, conductor 2026-09-07, R1
type Registry struct {
	ownedWall          []WallRow
	reservedClasses    []ReservedClass
	integrationClasses []IntegrationClass
	thresholds         []Threshold
	identitySources    []IdentitySource
	settings           []Setting
	correlatedRules    []CorrelatedRule
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
// ruled, conductor 2026-09-07, R4
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

// familyOwnedWall through familySettings name the seven declaration
// families a key can belong to, for the collision panic's message.
const (
	familyOwnedWall      = "owned wall"
	familyReserved       = "reserved classes"
	familyIntegration    = "integration classes"
	familyThresholds     = "thresholds"
	familyIdentitySource = "identity sources"
	familySuperseded     = "superseded names"
	familySettings       = "settings"
)

// panicDuplicate panics reporting that key, already declared in
// existing, cannot also be declared in family. Every family's
// collision check shares this one message shape, so a rung reading
// one panic recognizes them all.
func panicDuplicate(key, existing, family string) {
	panic(fmt.Sprintf("%s: already declared in %s, cannot also declare in %s", key, existing, family))
}

// requirePrefix panics if key does not carry the DUBPLATE_ prefix (O5
// class 3, Resolution R16: every family's keys share one flat
// environment namespace, so a declared key without the prefix would
// collide with an adopter's own variables). Every Add* entry point
// calls this for each of its keys before any claim, so a key that
// never reaches claim still gets checked.
func requirePrefix(family, key string) {
	if !strings.HasPrefix(key, "DUBPLATE_") {
		panic(fmt.Sprintf("%s: declared in %s without the DUBPLATE_ prefix", key, family))
	}
}

// claim panics if key was already claimed by a family other than
// family (O5 class 2), then records that family claims key. It
// deliberately does not flag a repeat claim from the same family: that
// is a within-family duplicate (O5 class 1), and each family's own
// entry point checks for it separately, so the two checks stay
// independent of each other. Callers run [requirePrefix] on key before
// calling claim.
func (r *Registry) claim(family, key string) {
	if existing, ok := r.claims[key]; ok && existing != family {
		panicDuplicate(key, existing, family)
	}
	r.claims[key] = family
}

// AddWallRow registers row as an owned wall row, a key [Load]
// requires. Its Name must carry the DUBPLATE_ prefix (O5 class 6) and
// is checked against every other declared key.
func (r *Registry) AddWallRow(row WallRow) {
	requirePrefix(familyOwnedWall, row.Name)
	if r.ownedNow(row.Name) {
		panicDuplicate(row.Name, familyOwnedWall, familyOwnedWall)
	}
	r.claim(familyOwnedWall, row.Name)
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
			panicDuplicate(key, familyReserved, familyReserved)
		}
	}
	for _, key := range class.Keys {
		requirePrefix(familyReserved, key)
		for _, existing := range r.reservedClasses {
			if slices.Contains(existing.Keys, key) {
				panicDuplicate(key, familyReserved, familyReserved)
			}
		}
		r.claim(familyReserved, key)
	}
	r.reservedClasses = append(r.reservedClasses, class)
}

// ReservedClasses returns the reserved classes in declaration order,
// never sorted: doctor prints them in table order, which is not rung
// order, and that order is the contract (Resolution R15).
//
// ruled, conductor 2026-09-07, R15
func (r *Registry) ReservedClasses() []ReservedClass {
	return slices.Clone(r.reservedClasses)
}

// AddIntegrationClass registers class as an adopter-optional
// integration, distinct from a reserved class because it is enabled by
// a flag rather than blocked until a rung lands (Resolution R2). class
// must carry a non-empty Flag that is not one of its own Required
// keys; an integration whose enable flag is itself one of the keys it
// gates could never be told apart from a required row in an error
// message.
//
// Flag and Required claim across families, so a key that collides
// with an owned-wall, reserved, or superseded name is refused rather
// than registering silently.
//
// ruled, conductor 2026-09-07, R2
func (r *Registry) AddIntegrationClass(class IntegrationClass) {
	if class.Flag == "" {
		panic(fmt.Sprintf("%s: integration class declared with an empty Flag", class.Name))
	}
	if slices.Contains(class.Required, class.Flag) {
		panic(fmt.Sprintf("%s: integration class's Flag %s also appears in its own key set", class.Name, class.Flag))
	}

	keys := append([]string{class.Flag}, class.Required...)

	for _, key := range keys {
		requirePrefix(familyIntegration, key)
	}
	for i, key := range keys {
		if slices.Contains(keys[:i], key) {
			panicDuplicate(key, familyIntegration, familyIntegration)
		}
	}
	for _, key := range keys {
		for _, existing := range r.integrationClasses {
			if key == existing.Flag || slices.Contains(existing.Required, key) {
				panicDuplicate(key, familyIntegration, familyIntegration)
			}
		}
	}

	for _, key := range keys {
		r.claim(familyIntegration, key)
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
// ruled, conductor 2026-09-07, R11
func (r *Registry) AddThreshold(t Threshold) {
	requirePrefix(familyThresholds, t.Key)
	if slices.ContainsFunc(r.thresholds, func(existing Threshold) bool { return existing.Key == t.Key }) {
		panicDuplicate(t.Key, familyThresholds, familyThresholds)
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
// ruled, conductor 2026-09-07, R11
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
		if _, err := parseFiniteFloat(t.Default); err != nil {
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

// IdentitySource is one way a deployment establishes who a request
// comes from. A deployment runs exactly one, named by the value of
// Selector, which is why the family selects over declarations rather
// than gating each on a boolean flag the way [IntegrationClass] does
// (Resolution R-5b5). The rung that builds a source declares it from
// its own file, so this package names no source and no key of one.
//
// Required and Optional differ in what [Load] does with them, not in
// what the wall accepts: Load requires every Required key once the
// source is selected and reads every Optional key that is set, while
// an unselected source's keys are accepted and never read, so an
// adopter can keep a second source's configuration in place and switch
// between them by one value.
type IdentitySource struct {
	// Name is the value Selector carries to select this source, and
	// the name doctor prints for its row.
	Name string
	// Selector is the environment key whose value names the active
	// source. Every declared source names the same one. It is not
	// claimed as this family's key: the rung declaring the first
	// source registers it as an owned wall row, which is what makes
	// Load require it.
	Selector string
	Required []string
	Optional []string
	// Validate reports what this source makes of its own
	// configuration: values carries its resolved Required and Optional
	// values by key, absent for any that resolved unset, and environ
	// is the same os.Environ() snapshot Load read. The snapshot is
	// what lets a hook validate a registered [Threshold], which
	// resolves at its point of use and never through Load.
	//
	// Load runs the hook whenever the source is selected, including
	// when a required key is unset, so a value rule and a missing row
	// are reported together rather than one pass at a time; a hook
	// therefore checks a value only when values carries it. Its error
	// joins Load's others. A nil Validate is a source with no value
	// rules of its own.
	//
	// ruled, conductor 2026-09-12, B2
	Validate func(values map[string]string, environ []string) error
}

// Selected reports whether environ selects s, by comparing s.Name
// against the value of s.Selector. It is doctor's own derivation, on
// [IntegrationClass.Enabled]'s precedent.
func (s IdentitySource) Selected(environ []string) bool {
	v, _ := newWall(environ).get(s.Selector)
	return v == s.Name
}

// AddIdentitySource registers src as one of the identity sources a
// deployment selects between. Its Required and Optional keys claim
// across families, so a source key colliding with an owned-wall,
// reserved, integration, or superseded name is refused here rather
// than misclassifying at load. Two sources may share a key, since two
// ways of establishing identity can legitimately need one value, and
// [Registry.claim] permits a repeat claim from the same family.
//
// A second source declared on a different Selector panics: a
// deployment selects one source through one key, and two selectors
// would leave Load with no single value to read.
func (r *Registry) AddIdentitySource(src IdentitySource) {
	if src.Name == "" {
		panic("identity source declared with an empty Name")
	}
	requirePrefix(familyIdentitySource, src.Selector)
	for _, existing := range r.identitySources {
		if existing.Name == src.Name {
			panic(fmt.Sprintf("%s: identity source already declared", src.Name))
		}
		if existing.Selector != src.Selector {
			panic(fmt.Sprintf("%s: identity source declared on selector %s, but %s selects on %s", src.Name, src.Selector, existing.Name, existing.Selector))
		}
	}

	keys := append(slices.Clone(src.Required), src.Optional...)
	if slices.Contains(keys, src.Selector) {
		panic(fmt.Sprintf("%s: identity source's Selector %s also appears in its own key set", src.Name, src.Selector))
	}
	for _, key := range keys {
		requirePrefix(familyIdentitySource, key)
	}
	for i, key := range keys {
		if slices.Contains(keys[:i], key) {
			panicDuplicate(key, familyIdentitySource, familyIdentitySource)
		}
	}
	for _, key := range keys {
		r.claim(familyIdentitySource, key)
	}
	r.identitySources = append(r.identitySources, src)
}

// IdentitySources returns the declared identity sources in
// declaration order.
func (r *Registry) IdentitySources() []IdentitySource {
	return slices.Clone(r.identitySources)
}

// identitySourceKey reports whether name is declared by an identity
// source in r: the selector every source shares, or any source's
// Required or Optional key. The selector answers here as well as on
// the owned wall, because a registry can declare a source before any
// rung registers the selector as a wall row, and the wall must not
// call the selector unknown in the meantime.
func (r *Registry) identitySourceKey(name string) bool {
	for _, src := range r.identitySources {
		if name == src.Selector || slices.Contains(src.Required, name) || slices.Contains(src.Optional, name) {
			return true
		}
	}
	return false
}

// AddSetting registers s as a store-homed key. Its key claims across
// families like every other, so a setting that collides with a
// threshold, a wall row, a class key, an identity source key, or a
// superseded name panics naming both families.
//
// Registration builds s's validator from its bounds and panics on a
// declaration that could never be read: a kind with no bound, a bound
// that does not parse or a floor above the ceiling, a KindString
// setting with no allowed values or a bound of another kind, a default
// that does not parse or falls outside the bounds, or an apply mode
// left at its zero value or set to one no one named.
func (r *Registry) AddSetting(s Setting) {
	requirePrefix(familySettings, s.Key)
	if r.hasSetting(s.Key) {
		panicDuplicate(s.Key, familySettings, familySettings)
	}
	switch {
	case s.Apply == 0:
		panic(fmt.Sprintf("%s: setting declared with no apply mode", s.Key))
	case s.Apply > ApplyRestart:
		panic(fmt.Sprintf("%s: unnamed apply mode %v", s.Key, s.Apply))
	}
	s.validate = buildValidator(s)
	if err := s.validate(s.Default); err != nil {
		panic(fmt.Sprintf("%s: invalid default %q: %v", s.Key, s.Default, err))
	}
	r.claim(familySettings, s.Key)
	r.settings = append(r.settings, s)
}

// Settings returns the registered settings in declaration order, each
// carrying the validator [Registry.AddSetting] built for it.
func (r *Registry) Settings() []Setting {
	return slices.Clone(r.settings)
}

// AddCorrelatedRule registers rule as a check over two or more
// registered settings. A rule naming a key that is not a registered
// setting panics naming it, a wall threshold included: a threshold
// resolves from the environment at its point of use, never through
// the seed evaluation that runs these rules.
func (r *Registry) AddCorrelatedRule(rule CorrelatedRule) {
	if rule.Name == "" {
		panic("correlated rule declared with an empty Name")
	}
	if slices.ContainsFunc(r.correlatedRules, func(existing CorrelatedRule) bool { return existing.Name == rule.Name }) {
		panic(fmt.Sprintf("correlated rule %q already declared", rule.Name))
	}
	if len(rule.Keys) < 2 {
		panic(fmt.Sprintf("correlated rule %q names %d keys, want two or more", rule.Name, len(rule.Keys)))
	}
	if rule.Check == nil {
		panic(fmt.Sprintf("correlated rule %q declared with no Check", rule.Name))
	}
	for i, key := range rule.Keys {
		if slices.Contains(rule.Keys[:i], key) {
			panic(fmt.Sprintf("correlated rule %q names %s more than once", rule.Name, key))
		}
		if r.hasSetting(key) {
			continue
		}
		if family, ok := r.claims[key]; ok {
			panic(fmt.Sprintf("%s: correlated rule %q names a key that is not a registered setting, declared in %s", key, rule.Name, family))
		}
		panic(fmt.Sprintf("%s: correlated rule %q names a key that is not a registered setting", key, rule.Name))
	}
	r.correlatedRules = append(r.correlatedRules, rule)
}

// CorrelatedRules returns the registered correlated rules in
// declaration order.
func (r *Registry) CorrelatedRules() []CorrelatedRule {
	return slices.Clone(r.correlatedRules)
}

// AddSuperseded registers name as a pre-pivot key with no meaning in
// the greenfield wall, reason being the message rejectNamespace reports
// for it. The superseded family stays a map rather than a slice
// because no caller enumerates it: supersessionOf is a single lookup,
// and order is meaningless for one (Resolution R8).
//
// ruled, conductor 2026-09-07, R8
func (r *Registry) AddSuperseded(name, reason string) {
	requirePrefix(familySuperseded, name)
	if _, ok := r.superseded[name]; ok {
		panicDuplicate(name, familySuperseded, familySuperseded)
	}
	r.claim(familySuperseded, name)
	r.superseded[name] = reason
}

// Superseded returns the superseded-name map, keyed by the pre-pivot
// name and valued by its disposition message.
func (r *Registry) Superseded() map[string]string {
	return maps.Clone(r.superseded)
}
