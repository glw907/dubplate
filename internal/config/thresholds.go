package config

import (
	"fmt"
	"strconv"
)

// Threshold describes one behavior tunable: its environment key, its
// default value in string form, and, in Doc, the accepted form a
// caller reports on a parse error. Rungs register thresholds as they
// introduce them; doctor reports each with its effective value.
type Threshold struct {
	Key     string
	Default string
	Doc     string
}

// registeredThresholds holds every [Threshold] added by [Register], in
// registration order. Rung 1 registers none: the registry is the
// deliverable, not a deferred set of values.
var registeredThresholds []Threshold

// Register adds t to the threshold registry. A rung's own package
// calls Register once per tunable it introduces, from an explicit
// setup call rather than an init function, so registration order stays
// visible in the caller's own code.
func Register(t Threshold) {
	registeredThresholds = append(registeredThresholds, t)
}

// Thresholds returns the registered tunables in registration order.
func Thresholds() []Threshold {
	return clone(registeredThresholds)
}

// ParseInt resolves t as an integer against environ: t.Default's own
// parse when t.Key is unset or empty, or environ's value when set. It
// records an error naming t.Key and t.Doc when a set value is not a
// valid integer, and an error naming t.Key when Default itself does
// not parse, which is a registration bug rather than an environment
// one.
func ParseInt(environ []string, t Threshold) (int, error) {
	w := newWall(environ)

	v, ok := w.get(t.Key)
	if !ok {
		def, err := strconv.Atoi(t.Default)
		if err != nil {
			return 0, fmt.Errorf("%s: invalid default %q", t.Key, t.Default)
		}
		return def, nil
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid value %q, want %s", t.Key, v, t.Doc)
	}
	return n, nil
}
