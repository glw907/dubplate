// Package config is dubplate-server's portability wall: the only place
// the process environment is read. A misconfigured adopter install fails
// loudly at [Load], never silently borrowing a value from someone else's
// deployment.
//
// Every DUBPLATE_* name falls into exactly one of four classes, and
// [Load] enforces the classification rather than merely reading values:
//
//  1. Owned, required now: DUBPLATE_LIBRARY_ROOT, DUBPLATE_STATE_DIR,
//     DUBPLATE_SITE_NAME, and DUBPLATE_ADMIN_EMAIL. A fifth wall row,
//     the active identity source, is owned from rung 1 but has no
//     env key of its own yet: magic-link is the shipped default
//     source, and its discriminator plus required-when-selected
//     block (the two secrets, DUBPLATE_PUBLIC_URL, and an SMTP
//     transport) arrive at rung 5, when login is server purview.
//     Until then this row reports pending, never satisfied and never
//     an error.
//  2. Reserved for a rung: a name the wall recognizes as belonging to
//     a capability that has not landed yet. Setting one early is a
//     load error naming its class and rung, never "unknown". The
//     classes are auth and identity (rung 5), backup targets
//     (rung 6), intake (rung 7), pipeline (rung 9), monitoring
//     (rung 9), decision and review queues (rung 13), and the
//     streamer (rung 12). See [ReservedClasses] for the exact
//     membership.
//  3. Superseded: a pre-pivot name with no meaning here. Load errors
//     naming the replacement, or, for a capability the greenfield
//     server drops outright (the estate's beets and contributor-file
//     config, the Cloudflare Tunnel credential compose.yaml reads
//     directly), naming that it retired with no successor.
//  4. Unknown: any other DUBPLATE_* name. This is the compat
//     discipline evolution lesson U5 draws from beets 2.4.0's silent
//     plugin-list regression: an unrecognized key refuses the start
//     rather than degrading it.
//
// [Thresholds] is a separate, open-ended registry for the behavior
// tunables rungs introduce as they land; rung 1 registers none.
package config

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
)

// ErrMissing marks a required wall row Load found unset. Load wraps it
// once per missing row and joins every wrap with [errors.Join], so a
// caller such as the doctor subcommand can report the complete list
// rather than stopping at the first miss.
var ErrMissing = errors.New("required setting unset")

// Config is the resolved configuration of one dubplate-server process.
// Rung 1 carries the owned wall only; the identity-source selector and
// the integrations behind it join their fields at their own rungs.
type Config struct {
	LibraryRoot string
	StateDir    string
	SiteName    string
	// AdminEmail is DUBPLATE_ADMIN_EMAIL parsed as a single RFC 5322
	// address and lowercased: adopters compare it, never display it,
	// so the wall normalizes case at load time rather than leaving
	// every caller to do it.
	AdminEmail string
	// Shadow is rung 4's integration, resolved from the
	// required-when-enabled rows of the shadow class. It is nil when
	// DUBPLATE_SHADOW_ENABLED is unset, which is every adopter who has
	// not opted in.
	Shadow *ShadowIntegration
}

// ShadowIntegration is the shadow integration's resolved paths, the
// three required-when-enabled rows of the shadow class. The floor and
// the request cap are registered thresholds rather than fields here,
// because both carry defaults and Config carries no defaulted value.
type ShadowIntegration struct {
	ObserveDir     string
	OracleStateDir string
	ReviewDir      string
}

// wall indexes one environment snapshot and accumulates load errors. It
// is the single choke point this package reads the environment through:
// get owns the duplicate-key (last wins), empty-equals-unset, and
// =-in-value decisions, and every other accessor in this package calls
// get rather than reading environ or env directly.
type wall struct {
	env  map[string]string
	errs []error
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
	return &wall{env: env}
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
	if !ok {
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
// wall's environment and records an error for any that is not an owned
// row: a reserved name errors naming its class and rung, a superseded
// name errors naming its replacement, and anything else errors as
// unknown. Owned rows are validated by their own require calls, not
// here, so a valid owned value is never flagged.
func (w *wall) rejectNamespace() {
	for name := range w.env {
		if !strings.HasPrefix(name, "DUBPLATE_") || ownedNow(name) {
			continue
		}
		if _, ok := w.get(name); !ok {
			continue // set but empty: treated as unset, nothing to flag
		}

		if class, ok := integrationClassOf(name); ok {
			// The flag names itself, so it is never "set without"
			// itself. Load already reports an unparseable flag value;
			// reporting it again here would emit two errors for one
			// typo, one of them self-contradictory.
			if name == class.Flag {
				continue
			}
			if enabled, err := w.integrationEnabled(class); err != nil || !enabled {
				w.errs = append(w.errs, fmt.Errorf("%s: set without %s, the %s integration's enable flag", name, class.Flag, class.Name))
			}
			continue
		}
		if class, ok := reservedClassOf(name); ok {
			w.errs = append(w.errs, fmt.Errorf("%s: reserved for %s, arrives at rung %d", name, class.Name, class.Rung))
			continue
		}
		if msg, ok := supersessionOf(name); ok {
			w.errs = append(w.errs, fmt.Errorf("%s: %s", name, msg))
			continue
		}
		w.errs = append(w.errs, fmt.Errorf("%s: unknown configuration key", name))
	}
}

// Load resolves [Config] from environ (os.Environ() form). It reads
// only DUBPLATE_-prefixed keys and returns an error naming every
// problem at once, rather than stopping at the first one: every unset
// owned row, every reserved or superseded key set before its rung, and
// every unrecognized DUBPLATE_* key.
func Load(environ []string) (*Config, error) {
	w := newWall(environ)

	cfg := &Config{
		LibraryRoot: w.require("DUBPLATE_LIBRARY_ROOT"),
		StateDir:    w.require("DUBPLATE_STATE_DIR"),
		SiteName:    w.require("DUBPLATE_SITE_NAME"),
		AdminEmail:  w.requireEmail("DUBPLATE_ADMIN_EMAIL"),
	}

	for _, class := range integrationClasses {
		enabled, err := w.integrationEnabled(class)
		if err != nil {
			w.errs = append(w.errs, err)
			continue
		}
		if !enabled {
			continue
		}
		resolved := make(map[string]string, len(class.Required))
		for _, key := range class.Required {
			resolved[key] = w.require(key)
		}
		if class.Name == "shadow" {
			cfg.Shadow = &ShadowIntegration{
				ObserveDir:     resolved["DUBPLATE_SHADOW_OBSERVE_DIR"],
				OracleStateDir: resolved["DUBPLATE_SHADOW_ORACLE_STATE_DIR"],
				ReviewDir:      resolved["DUBPLATE_SHADOW_REVIEW_DIR"],
			}
		}
	}

	w.rejectNamespace()

	if len(w.errs) > 0 {
		return nil, errors.Join(w.errs...)
	}
	return cfg, nil
}
