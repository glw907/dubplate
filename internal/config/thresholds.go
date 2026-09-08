package config

import (
	"fmt"
	"strconv"
	"time"
)

// ThresholdKind names the parser a threshold's value is read through.
// KindInt is the zero value, so a Threshold declared before this field
// existed keeps its meaning.
//
// provisional, conductor ruling owed, R11
type ThresholdKind int

const (
	KindInt ThresholdKind = iota
	KindDuration
	KindFloat
	KindString
)

// Threshold describes one behavior tunable: its environment key, its
// default value in string form, and, in Doc, the accepted form a
// caller reports on a parse error. Rungs register thresholds as they
// introduce them; doctor reports each with its effective value.
type Threshold struct {
	Key     string
	Kind    ThresholdKind
	Default string
	Doc     string
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

// ParseDuration resolves t as a duration against environ, in ParseInt's
// own shape: t.Default's parse when t.Key is unset, environ's value
// when set, an error naming t.Key and t.Doc on a bad value, and an
// error naming t.Key alone on a bad default, which is a registration
// bug rather than an environment one.
func ParseDuration(environ []string, t Threshold) (time.Duration, error) {
	w := newWall(environ)

	v, ok := w.get(t.Key)
	if !ok {
		def, err := time.ParseDuration(t.Default)
		if err != nil {
			return 0, fmt.Errorf("%s: invalid default %q", t.Key, t.Default)
		}
		return def, nil
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid value %q, want %s", t.Key, v, t.Doc)
	}
	return d, nil
}

// ParseFloat resolves t as a float64 against environ, in ParseInt's own
// shape: t.Default's parse when t.Key is unset, environ's value when
// set, an error naming t.Key and t.Doc on a bad value, and an error
// naming t.Key alone on a bad default, which is a registration bug
// rather than an environment one.
func ParseFloat(environ []string, t Threshold) (float64, error) {
	w := newWall(environ)

	v, ok := w.get(t.Key)
	if !ok {
		def, err := strconv.ParseFloat(t.Default, 64)
		if err != nil {
			return 0, fmt.Errorf("%s: invalid default %q", t.Key, t.Default)
		}
		return def, nil
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid value %q, want %s", t.Key, v, t.Doc)
	}
	return f, nil
}

// ParseString resolves t as a string against environ: t.Default when
// t.Key is unset, or environ's raw value when set. A set value can
// never fail to parse, since any string is a valid string; the one
// invalid Default is the empty string, because wall.get's
// empty-equals-unset convention would otherwise make an empty-string
// default indistinguishable from no default at all. That case is a
// registration bug, in ParseInt's own naming-the-key-alone shape.
func ParseString(environ []string, t Threshold) (string, error) {
	w := newWall(environ)

	v, ok := w.get(t.Key)
	if !ok {
		if t.Default == "" {
			return "", fmt.Errorf("%s: invalid default %q", t.Key, t.Default)
		}
		return t.Default, nil
	}

	return v, nil
}
