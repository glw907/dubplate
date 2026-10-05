// Package config is dubplate-server's portability wall: the only place
// the process environment is read. A misconfigured adopter install fails
// loudly at [Load], never silently borrowing a value from someone else's
// deployment.
//
// Every DUBPLATE_* name falls into exactly one of six classes, and
// [Load] enforces the classification rather than merely reading values:
//
//  1. Owned, required now: DUBPLATE_LIBRARY_ROOT, DUBPLATE_STATE_DIR,
//     DUBPLATE_SITE_NAME, DUBPLATE_ADMIN_EMAIL, and, from rung 5,
//     DUBPLATE_IDENTITY_SOURCE, the selector naming which declared
//     identity source is active (cmd/dubplate-server/auth.go
//     registers it as an owned wall row alongside the source itself).
//  2. Reserved for a rung: a name the wall recognizes as belonging to
//     a capability that has not landed yet. Setting one early is a
//     load error naming its class and rung, never "unknown". The
//     classes are intake (rung 7), pipeline (rung 9), monitoring
//     (rung 9), decision and review queues (rung 13), and the
//     streamer (rung 12). See [ReservedClasses] for the exact
//     membership.
//  3. Adopter-optional integration: a name accepted only when its
//     class's own Flag resolves true. The class's Required rows are
//     required at Load once the flag is set, and its Optional rows
//     are accepted as registered thresholds rather than wall fields.
//     Setting a Required or Optional row without its flag enabled is
//     a load error naming the flag. See [Registry.IntegrationClasses]
//     for the exact membership; the owning package registers its own
//     class (the shadow and backup classes, for example).
//  4. Superseded: a pre-pivot name with no meaning here. Load errors
//     naming the replacement, or, for a capability the greenfield
//     server drops outright (the estate's beets config, the
//     Cloudflare Tunnel credential of the retired incumbent stack),
//     naming that it retired with no successor.
//  5. Declared by an identity source: the selector naming the active
//     source, or a key one of the registry's identity sources
//     requires or accepts. The selected source's required keys are
//     required at [Load] and its optional keys are read when set,
//     while an unselected source's keys are accepted and never read,
//     so an adopter can keep a second source's configuration in place
//     and switch between them by one value. See
//     [Registry.IdentitySources] for the exact membership; the owning
//     package registers its own source.
//  6. Unknown: any other DUBPLATE_* name. This is the compat
//     discipline evolution lesson U5 draws from beets 2.4.0's silent
//     plugin-list regression: an unrecognized key refuses the start
//     rather than degrading it.
//
// Two kinds of declared name sit outside those six, and rejectNamespace
// accepts both without parsing their values. A behavior tunable a rung
// registers as a [Threshold] resolves from the environment at its point
// of use. A store-homed key a rung registers as a [Setting] lives in
// the store, and its environment value is only a seed that the seed
// evaluation parses and bounds; this package never reads it.
//
// Every declaration above (the owned wall, the reserved classes, the
// adopter-optional integrations, the identity sources, the superseded
// names, the thresholds, and the settings with the correlated rules
// over them) lives in one [Registry] a caller assembles and passes to
// [Load]. [NewRegistry] returns a registry already seeded with rung 1's
// own rows; rung 1 registers no threshold, identity source, or setting.
package config

import (
	"errors"
	"fmt"
	"maps"
	"net/mail"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrMissing marks a required wall row Load found unset. Load wraps it
// once per missing row and joins every wrap with [errors.Join], so a
// caller such as the doctor subcommand can report the complete list
// rather than stopping at the first miss.
var ErrMissing = errors.New("required setting unset")

// Config is the resolved configuration of one dubplate-server process.
// Rung 1 carries the owned wall only; the active identity source joins
// its fields as IdentitySource once a rung declares one, and that
// source's own resolved values are reachable by key through
// [Config.IdentitySourceValue] rather than a named field per source.
// An enabled integration class's
// resolved Required values are reachable by class name through
// [Config.IntegrationEnabled] and [Config.IntegrationValue] rather than
// a named field per class, so a second integration class needs no
// second field (Resolution R5).
type Config struct {
	LibraryRoot string
	StateDir    string
	SiteName    string
	// AdminEmail is DUBPLATE_ADMIN_EMAIL parsed as a single RFC 5322
	// address and lowercased: adopters compare it, never display it,
	// so the wall normalizes case at load time rather than leaving
	// every caller to do it.
	AdminEmail string

	// IdentitySource is the Name of the identity source the selector
	// selected. It is empty when the registry declares no source, and
	// when the selector names none of the declared ones, which Load
	// reports as an error rather than resolving.
	IdentitySource string

	// integrations holds every integration class's resolved state,
	// keyed by class Name. It is populated for every class Load reads
	// off reg, whether or not its flag resolved true.
	integrations map[string]integrationResolution

	// sourceValues holds the active identity source's resolved
	// Required and Optional values by key, with no entry for a key
	// that resolved unset.
	sourceValues map[string]string
}

// integrationResolution is one integration class's resolved state:
// Enabled records whether its Flag resolved true, independently of
// whether it resolved any key, so a class with an empty Required set
// (hook 6's driver-declines-capabilities case) still reports enabled
// when its flag is set. Required carries the class's own Required key
// values, populated only when Enabled is true.
type integrationResolution struct {
	Enabled  bool
	Required map[string]string
}

// IntegrationEnabled reports whether the integration class named
// className had its Flag resolve true at Load. It answers false for a
// class name Load never saw registered.
func (c *Config) IntegrationEnabled(className string) bool {
	return c.integrations[className].Enabled
}

// IntegrationValue returns the resolved value of key, one of
// className's own Required rows, as Load resolved it. It returns the
// empty string when className is not enabled or key is not one of its
// Required rows.
func (c *Config) IntegrationValue(className, key string) string {
	return c.integrations[className].Required[key]
}

// IdentitySourceValue returns the resolved value of key, one of the
// active identity source's Required or Optional rows, as Load resolved
// it. It returns the empty string for an unset key, for a key the
// active source does not declare, and for a Config whose registry
// declared no source. Keying by name rather than adding a field per
// source is [Config.IntegrationValue]'s own precedent: a second source
// needs no second field.
func (c *Config) IdentitySourceValue(key string) string {
	return c.sourceValues[key]
}

// wall indexes one environment snapshot and accumulates load errors. It
// is the single choke point this package reads the environment through:
// get owns the duplicate-key (last wins), empty-equals-unset, and
// =-in-value decisions, and every other accessor in this package calls
// get rather than reading environ or env directly.
type wall struct {
	env  map[string]string
	errs []error
	// missing records the names require has already flagged, so a name
	// required twice is reported once. The identity selector is the
	// case: it is an owned wall row and the source family's selector
	// both, and an adopter who left it unset should read one line.
	missing map[string]bool
}

// newWall indexes environ (os.Environ() form) into a wall, later
// entries winning over earlier ones for the same name.
func newWall(environ []string) *wall {
	env := make(map[string]string, len(environ))
	for _, kv := range environ {
		name, value, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		env[name] = value
	}
	return &wall{env: env, missing: make(map[string]bool)}
}

// get returns name's value, treating an empty value as unset.
func (w *wall) get(name string) (string, bool) {
	v, ok := w.env[name]
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

// require returns name's value, recording [ErrMissing] against name
// when it is unset.
func (w *wall) require(name string) string {
	v, ok := w.get(name)
	if !ok && !w.missing[name] {
		w.missing[name] = true
		w.errs = append(w.errs, fmt.Errorf("%s: %w", name, ErrMissing))
	}
	return v
}

// requireEmail is [wall.require] for DUBPLATE_ADMIN_EMAIL: it parses
// the value as a single RFC 5322 address and returns the address
// portion lowercased, recording a parse error against name when the
// value is set but not a valid address.
func (w *wall) requireEmail(name string) string {
	v := w.require(name)
	if v == "" {
		return ""
	}
	addr, err := mail.ParseAddress(v)
	if err != nil {
		w.errs = append(w.errs, fmt.Errorf("%s: invalid email address %q", name, v))
		return ""
	}
	return strings.ToLower(addr.Address)
}

// siteNameMaxBytes bounds DUBPLATE_SITE_NAME: room for any group's
// name, and short enough for a page title and a mail subject.
const siteNameMaxBytes = 100

// requireSiteName is [wall.require] for DUBPLATE_SITE_NAME, which
// every served document and page title carries: it records an error
// against name when the value is over siteNameMaxBytes, is not valid
// UTF-8, or holds a C0, DEL, or C1 control character or a
// bidirectional embedding, override, or isolate, any of which could
// hide or reorder the text around it. The error never echoes the
// value.
func (w *wall) requireSiteName(name string) string {
	v := w.require(name)
	var problem string
	switch {
	case v == "":
		return ""
	case len(v) > siteNameMaxBytes:
		problem = fmt.Sprintf("is over %d bytes", siteNameMaxBytes)
	case !utf8.ValidString(v):
		problem = "is not valid UTF-8"
	case strings.ContainsFunc(v, unicode.IsControl):
		problem = "holds a control character"
	case strings.ContainsFunc(v, isBidiControl):
		problem = "holds a bidirectional override"
	default:
		return v
	}
	w.errs = append(w.errs, fmt.Errorf("%s: the site name %s", name, problem))
	return ""
}

// isBidiControl reports whether r is a bidirectional embedding or
// override (U+202A to U+202E) or isolate (U+2066 to U+2069).
func isBidiControl(r rune) bool {
	return (r >= '\u202a' && r <= '\u202e') || (r >= '\u2066' && r <= '\u2069')
}

// integrationEnabled reports whether class's flag resolves true in w's
// environment. An unset or empty flag is off. A value strconv.ParseBool
// refuses is reported as an error rather than silently read as off,
// because a typo that disables an unattended observer is the failure
// this class exists to prevent.
func (w *wall) integrationEnabled(class IntegrationClass) (bool, error) {
	v, ok := w.get(class.Flag)
	if !ok {
		return false, nil
	}
	enabled, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s: invalid value %q, want a boolean", class.Flag, v)
	}
	return enabled, nil
}

// rejectNamespace walks every DUBPLATE_*-prefixed name present in the
// wall's environment and records an error for any that is not a row reg
// owns: a reserved name errors naming its class and rung, a superseded
// name errors naming its replacement, and anything else errors as
// unknown. Owned rows are validated by their own require calls, not
// here, so a valid owned value is never flagged. Names are walked in
// sorted order so the joined refusal reads the same on every run.
//
// ruled, conductor 2026-09-07, R1
func (w *wall) rejectNamespace(reg *Registry) {
	for _, name := range slices.Sorted(maps.Keys(w.env)) {
		if !strings.HasPrefix(name, "DUBPLATE_") || ownedNow(reg, name) {
			continue
		}
		if _, ok := w.get(name); !ok {
			continue // set but empty: treated as unset, nothing to flag
		}

		if class, ok := integrationClassOf(reg, name); ok {
			// The flag names itself, so it is never "set without"
			// itself. Load already reports an unparseable flag value;
			// reporting it again here would emit two errors for one
			// typo, one of them self-contradictory.
			if name == class.Flag {
				continue
			}
			// An unparseable flag is set, and Load reports it on its
			// own; only a flag that resolves false leaves the key set
			// without it.
			if enabled, err := w.integrationEnabled(class); err == nil && !enabled {
				w.errs = append(w.errs, fmt.Errorf("%s: set without %s, the %s integration's enable flag", name, class.Flag, class.Name))
			}
			continue
		}
		if class, ok := reservedClassOf(reg, name); ok {
			w.errs = append(w.errs, fmt.Errorf("%s: reserved for %s, arrives at rung %d", name, class.Name, class.Rung))
			continue
		}
		if msg, ok := supersessionOf(reg, name); ok {
			w.errs = append(w.errs, fmt.Errorf("%s: %s", name, msg))
			continue
		}
		if identitySourceKey(reg, name) {
			// A declared source's key is declared whether or not its
			// source is the selected one, so an adopter who keeps a
			// second source's configuration in place is not told it is
			// unknown.
			continue
		}
		if thresholdOf(reg, name) {
			// A registered threshold is a declared name whose own
			// Parse* call resolves and validates it at the point of
			// use. Flagging it here would make a binary refuse the
			// very tunable it registered.
			continue
		}
		if settingOf(reg, name) {
			// A store-homed setting's environment value is only a
			// seed: the seed evaluation parses and bounds it, and
			// reports what it refuses, so Load must not parse it.
			continue
		}
		w.errs = append(w.errs, fmt.Errorf("%s: unknown configuration key", name))
	}
}

// resolveIdentitySource requires the selector the declared sources
// share, resolves the selected source's keys, and runs that source's
// validation hook, recording every failure on w. It returns the
// selected source's name and its resolved values, both zero when the
// selector is unset or names no declared source.
func (w *wall) resolveIdentitySource(sources []IdentitySource, environ []string) (string, map[string]string) {
	selector := sources[0].Selector
	name := w.require(selector)
	if name == "" {
		return "", nil
	}

	i := slices.IndexFunc(sources, func(s IdentitySource) bool { return s.Name == name })
	if i < 0 {
		declared := make([]string, len(sources))
		for j, s := range sources {
			declared[j] = s.Name
		}
		w.errs = append(w.errs, fmt.Errorf("%s: unknown identity source %q, want one of %s", selector, name, strings.Join(declared, ", ")))
		return "", nil
	}

	src := sources[i]
	values := make(map[string]string, len(src.Required)+len(src.Optional))
	for _, key := range src.Required {
		if v := w.require(key); v != "" {
			values[key] = v
		}
	}
	for _, key := range src.Optional {
		if v, ok := w.get(key); ok {
			values[key] = v
		}
	}
	if src.Validate != nil {
		if err := src.Validate(values, environ); err != nil {
			w.errs = append(w.errs, err)
		}
	}
	return src.Name, values
}

// Load resolves [Config] from environ (os.Environ() form), enforced
// against reg. It reads only DUBPLATE_-prefixed keys and returns an
// error naming every problem at once, rather than stopping at the
// first one: every unset owned row, every unsatisfied row of the
// selected identity source, every reserved or superseded key set
// before its rung, and every unrecognized DUBPLATE_* key.
//
// What Load requires comes from reg, never from a name written here: a
// rung that registers an owned-now wall row or an identity source from
// its own file gets the requirement from the registration
// (Resolution R-5b4).
//
// ruled, conductor 2026-09-07, R1
func Load(environ []string, reg *Registry) (*Config, error) {
	w := newWall(environ)

	cfg := &Config{
		LibraryRoot:  w.require("DUBPLATE_LIBRARY_ROOT"),
		StateDir:     w.require("DUBPLATE_STATE_DIR"),
		SiteName:     w.requireSiteName("DUBPLATE_SITE_NAME"),
		AdminEmail:   w.requireEmail("DUBPLATE_ADMIN_EMAIL"),
		integrations: make(map[string]integrationResolution, len(reg.integrationClasses)),
	}

	// The four rows above are rung 1's own and resolve into named
	// fields; every owned-now row any rung registered is required
	// here, whatever package declared it. require reports a name once
	// however many times it is called for, so the four do not report
	// twice.
	for _, row := range reg.ownedWall {
		if row.Pending == "" {
			w.require(row.Name)
		}
	}

	// Every enabled class's Required keys resolve here, generically, by
	// class name: no class name appears in this loop, so a second
	// integration class needs no second branch (Resolution R5). Enabled
	// is recorded independently of whether Required resolved any key,
	// which is what lets a class with an empty Required set (hook 6's
	// driver-declines-capabilities case) still report enabled.
	for _, class := range reg.integrationClasses {
		enabled, err := w.integrationEnabled(class)
		if err != nil {
			w.errs = append(w.errs, err)
			continue
		}
		resolution := integrationResolution{Enabled: enabled}
		if enabled {
			resolution.Required = make(map[string]string, len(class.Required))
			for _, key := range class.Required {
				resolution.Required[key] = w.require(key)
			}
		}
		cfg.integrations[class.Name] = resolution
	}

	// With no source declared nothing here runs, which is what leaves
	// the selector unrequired for a registry that declares none.
	if len(reg.identitySources) > 0 {
		cfg.IdentitySource, cfg.sourceValues = w.resolveIdentitySource(reg.identitySources, environ)
	}

	w.rejectNamespace(reg)

	if len(w.errs) > 0 {
		return nil, errors.Join(w.errs...)
	}
	return cfg, nil
}
