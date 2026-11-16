package config

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ThresholdKind names the parser a threshold's value is read through.
// KindInt is the zero value, so a Threshold declared before this field
// existed keeps its meaning.
//
// ruled, conductor 2026-09-07, R11
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

// errUnnamedKind marks a parse under a kind no rung has named.
var errUnnamedKind = errors.New("kind has no parser")

// Parse reads s under k, returning an int, a time.Duration, a finite
// float64, or s itself for KindString, so two values parsed under one
// kind compare with ==. A float refuses NaN and both infinities, which
// strconv.ParseFloat accepts by name.
func (k ThresholdKind) Parse(s string) (any, error) {
	switch k {
	case KindInt:
		return strconv.Atoi(s)
	case KindDuration:
		return time.ParseDuration(s)
	case KindFloat:
		return parseFiniteFloat(s)
	case KindString:
		return s, nil
	default:
		return nil, fmt.Errorf("%v: %w", k, errUnnamedKind)
	}
}

// parseAs is [ThresholdKind.Parse] typed as T, the type k parses to.
func parseAs[T any](k ThresholdKind, s string) (T, error) {
	var zero T
	v, err := k.Parse(s)
	if err != nil {
		return zero, err
	}
	typed, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("%v parses to %T, not %T", k, v, zero)
	}
	return typed, nil
}

// Threshold describes one behavior tunable: its environment key, its
// default value in string form, and, in Doc, the accepted form a
// caller reports on a parse error. Rungs register thresholds as they
// introduce them; doctor reports each with its effective value.
// Description says what the key controls and Mark whether a change to
// it needs care; every registered key carries both.
//
// Floor and Ceiling are the inclusive bounds of a KindInt,
// KindDuration, or KindFloat threshold, in string form, parsed under
// Kind; an empty one leaves that side open.
type Threshold struct {
	Key         string
	Kind        ThresholdKind
	Default     string
	Doc         string
	Description string
	Mark        Sensitivity
	Floor       string
	Ceiling     string
}

// checkBounds returns the phrase, completing "is ...", for the bound
// of t that v lies outside, or nil when v is within both. A bound that
// does not parse under t.Kind panics: it is a registration bug, as
// [Registry.AddSetting] treats one.
func checkBounds[T cmp.Ordered](t Threshold, v T) error {
	if t.Floor != "" {
		floor, err := parseAs[T](t.Kind, t.Floor)
		if err != nil {
			panic(fmt.Sprintf("%s: invalid floor %q", t.Key, t.Floor))
		}
		if v < floor {
			return fmt.Errorf("below the floor %s", t.Floor)
		}
	}
	if t.Ceiling != "" {
		ceiling, err := parseAs[T](t.Kind, t.Ceiling)
		if err != nil {
			panic(fmt.Sprintf("%s: invalid ceiling %q", t.Key, t.Ceiling))
		}
		if v > ceiling {
			return fmt.Errorf("above the ceiling %s", t.Ceiling)
		}
	}
	return nil
}

// parseThreshold resolves t against environ under kind: t.Default's
// own parse when t.Key is unset or empty, or environ's value when set.
// It returns a [KeyError] naming t.Doc when a set value fails to
// parse, and one naming the default alone when Default itself fails to
// parse, which is a registration bug rather than an environment one.
// ParseInt, ParseDuration, and ParseFloat each call it with their own
// kind, so the unset/set-value/bad-default shape lives once for every
// numeric kind. A value or default outside t's bounds returns a
// [KeyError] naming the bound.
func parseThreshold[T cmp.Ordered](environ []string, t Threshold, kind ThresholdKind) (T, error) {
	w := newWall(environ)

	v, ok := w.get(t.Key)
	if !ok {
		def, err := parseAs[T](kind, t.Default)
		if err != nil {
			return def, &KeyError{Key: t.Key, Reason: fmt.Sprintf("invalid default %q", t.Default)}
		}
		if err := checkBounds(t, def); err != nil {
			return def, &KeyError{Key: t.Key, Reason: fmt.Sprintf("default %q is %v", t.Default, err)}
		}
		return def, nil
	}

	n, err := parseAs[T](kind, v)
	if err != nil {
		return n, &KeyError{Key: t.Key, Reason: fmt.Sprintf("invalid value %q, want %s", v, t.Doc)}
	}
	if err := checkBounds(t, n); err != nil {
		return n, &KeyError{Key: t.Key, Reason: fmt.Sprintf("value %q is %v", v, err)}
	}
	return n, nil
}

// ParseInt resolves t as an integer against environ; see
// [parseThreshold] for the unset/set-value/bad-default shape it
// shares with ParseDuration and ParseFloat.
func ParseInt(environ []string, t Threshold) (int, error) {
	return parseThreshold[int](environ, t, KindInt)
}

// ParseDuration resolves t as a duration against environ; see
// [parseThreshold] for the unset/set-value/bad-default shape it
// shares with ParseInt and ParseFloat.
func ParseDuration(environ []string, t Threshold) (time.Duration, error) {
	return parseThreshold[time.Duration](environ, t, KindDuration)
}

// ParseFloat resolves t as a finite float64 against environ, refusing
// NaN and both infinities; see [parseThreshold] for the
// unset/set-value/bad-default shape it shares with ParseInt and
// ParseDuration.
func ParseFloat(environ []string, t Threshold) (float64, error) {
	return parseThreshold[float64](environ, t, KindFloat)
}

// hasThreshold reports whether name is a threshold registered on r.
func (r *Registry) hasThreshold(name string) bool {
	for _, t := range r.thresholds {
		if t.Key == name {
			return true
		}
	}
	return false
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
			return "", &KeyError{Key: t.Key, Reason: fmt.Sprintf("invalid default %q", t.Default)}
		}
		return t.Default, nil
	}

	return v, nil
}
