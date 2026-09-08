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

// String returns k's name, or "ThresholdKind(N)" for a kind no rung
// has named, so a registration-bug panic naming an unrecognized Kind
// reports it rather than a bare integer.
func (k ThresholdKind) String() string {
	switch k {
	case KindInt:
		return "KindInt"
	case KindDuration:
		return "KindDuration"
	case KindFloat:
		return "KindFloat"
	case KindString:
		return "KindString"
	default:
		return fmt.Sprintf("ThresholdKind(%d)", int(k))
	}
}

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

// parseThreshold resolves t against environ through parse: t.Default's
// own parse when t.Key is unset or empty, or environ's value when set.
// It records an error naming t.Key and t.Doc when a set value fails to
// parse, and an error naming t.Key alone when Default itself fails to
// parse, which is a registration bug rather than an environment one.
// ParseInt, ParseDuration, and ParseFloat each call it with their own
// parse function, so the unset/set-value/bad-default shape lives once
// for every numeric kind.
func parseThreshold[T any](environ []string, t Threshold, parse func(string) (T, error)) (T, error) {
	w := newWall(environ)

	v, ok := w.get(t.Key)
	if !ok {
		def, err := parse(t.Default)
		if err != nil {
			var zero T
			return zero, fmt.Errorf("%s: invalid default %q", t.Key, t.Default)
		}
		return def, nil
	}

	n, err := parse(v)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("%s: invalid value %q, want %s", t.Key, v, t.Doc)
	}
	return n, nil
}

// ParseInt resolves t as an integer against environ; see
// [parseThreshold] for the unset/set-value/bad-default shape it
// shares with ParseDuration and ParseFloat.
func ParseInt(environ []string, t Threshold) (int, error) {
	return parseThreshold(environ, t, strconv.Atoi)
}

// ParseDuration resolves t as a duration against environ; see
// [parseThreshold] for the unset/set-value/bad-default shape it
// shares with ParseInt and ParseFloat.
func ParseDuration(environ []string, t Threshold) (time.Duration, error) {
	return parseThreshold(environ, t, time.ParseDuration)
}

// ParseFloat resolves t as a float64 against environ; see
// [parseThreshold] for the unset/set-value/bad-default shape it
// shares with ParseInt and ParseDuration.
func ParseFloat(environ []string, t Threshold) (float64, error) {
	return parseThreshold(environ, t, func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
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
