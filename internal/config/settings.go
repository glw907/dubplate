package config

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// ApplyMode names when a changed setting takes effect in a running
// server. The zero value names no mode, so a declaration that forgets
// the field panics at [Registry.AddSetting] rather than silently
// taking one.
type ApplyMode int

const (
	// ApplyImmediately takes effect on the next read of the setting.
	ApplyImmediately ApplyMode = iota + 1
	// ApplyNextOperation takes effect when the next operation that
	// reads the setting starts; one already running keeps the value it
	// started with.
	ApplyNextOperation
	// ApplyRestart takes effect when the server next starts.
	ApplyRestart
)

// String returns m's name as a report prints it, or "ApplyMode(N)"
// for a mode no one named, the zero value included.
func (m ApplyMode) String() string {
	switch m {
	case ApplyImmediately:
		return "immediately"
	case ApplyNextOperation:
		return "next operation"
	case ApplyRestart:
		return "restart"
	default:
		return fmt.Sprintf("ApplyMode(%d)", int(m))
	}
}

// Setting describes one store-homed key: a value the store holds, for
// which the environment supplies only a seed. It is a family of its
// own rather than a [Threshold] with a home field, so no function that
// reads a threshold from the environment accepts one; this package
// exports no reader of a setting's environment value at all.
//
// Floor and Ceiling are the inclusive bounds of a KindInt,
// KindDuration, or KindFloat setting, in string form, parsed under
// Kind. Values is a KindString setting's bound instead: the closed set
// of values it accepts, compared byte for byte. [Registry.AddSetting]
// builds the setting's validator from whichever bound its kind takes,
// so every refusal and every report prints a bound from the same data
// the check compares against.
type Setting struct {
	Key string
	// Kind names the parser the value is read through: KindInt,
	// KindDuration, KindFloat, or KindString.
	Kind ThresholdKind
	// Default is the value in force while the store holds no row for
	// Key, in string form.
	Default string
	// Doc is the accepted form a refusal reports on a parse failure.
	Doc     string
	Apply   ApplyMode
	Floor   string
	Ceiling string
	Values  []string
	// Description says what the setting controls, in a sentence a
	// member of the group can read. Doc stays the accepted form a
	// refusal prints.
	Description string
	// Mark is whether a live change to the value is dangerous, though
	// it is no secret: a write to a Sensitive setting requires fresh
	// re-authentication. A report prints its value and the mark.
	Mark Sensitivity

	// validate is the check AddSetting built from the setting's bound;
	// nil on a Setting that never passed through registration.
	validate func(value string) error
}

// Validate reports whether value parses under s's kind and falls
// within its bounds, with an error naming s.Key and the bound or the
// accepted form it failed. A float that is NaN or infinite is refused
// by name, since no comparison against a bound would catch it.
//
// Only a Setting returned by [Registry.Settings] carries a validator;
// any other refuses every value, naming s.Key as unregistered.
func (s Setting) Validate(value string) error {
	if s.validate == nil {
		return fmt.Errorf("%s: not a registered setting", s.Key)
	}
	return s.validate(value)
}

// CorrelatedRule is a check over two or more settings that no single
// setting's bounds can express, such as one duration having to cover
// another. Check receives each of Keys' values in force, in string
// form and keyed by key, and returns an error naming every key it
// compares, or nil when the values agree.
type CorrelatedRule struct {
	Name  string
	Keys  []string
	Check func(values map[string]string) error
}

// errNotFinite marks a float value that parsed but is NaN or infinite.
var errNotFinite = errors.New("not a finite number")

// parseFiniteFloat parses s as a float64, refusing NaN and both
// infinities, which strconv.ParseFloat accepts by name.
func parseFiniteFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, errNotFinite
	}
	return f, nil
}

// buildValidator returns s's validator, panicking if s's kind has no
// bound, its bounds do not parse or are out of order, or it carries
// the bound of another kind.
func buildValidator(s Setting) func(string) error {
	if s.Kind != KindString && len(s.Values) > 0 {
		panic(fmt.Sprintf("%s: a %v setting carries allowed values, which only a KindString setting takes", s.Key, s.Kind))
	}
	switch s.Kind {
	case KindInt:
		return boundedValidator(s, strconv.Atoi)
	case KindDuration:
		return boundedValidator(s, time.ParseDuration)
	case KindFloat:
		return boundedValidator(s, parseFiniteFloat)
	case KindString:
		return setValidator(s)
	default:
		panic(fmt.Sprintf("%s: setting kind %v has no bounds, want KindInt, KindDuration, KindFloat, or KindString", s.Key, s.Kind))
	}
}

// setValidator returns a check that a value is one of s.Values,
// panicking when the set is empty or s carries a floor or ceiling,
// which no string compares against.
func setValidator(s Setting) func(string) error {
	if len(s.Values) == 0 {
		panic(fmt.Sprintf("%s: KindString setting declared with no allowed values", s.Key))
	}
	if s.Floor != "" || s.Ceiling != "" {
		panic(fmt.Sprintf("%s: KindString setting declared with a floor or ceiling; its bound is Values", s.Key))
	}
	values := slices.Clone(s.Values)
	set := strings.Join(values, ", ")
	return func(value string) error {
		if !slices.Contains(values, value) {
			return fmt.Errorf("%s: value %q is not one of %s", s.Key, value, set)
		}
		return nil
	}
}

// boundedValidator parses s's bounds through parse once and returns a
// check of a value against them. The value is parsed through the same
// function, so a bound and a value compare as one type.
func boundedValidator[T cmp.Ordered](s Setting, parse func(string) (T, error)) func(string) error {
	floor, err := parse(s.Floor)
	if err != nil {
		panic(fmt.Sprintf("%s: invalid floor %q", s.Key, s.Floor))
	}
	ceiling, err := parse(s.Ceiling)
	if err != nil {
		panic(fmt.Sprintf("%s: invalid ceiling %q", s.Key, s.Ceiling))
	}
	if floor > ceiling {
		panic(fmt.Sprintf("%s: floor %s is above the ceiling %s", s.Key, s.Floor, s.Ceiling))
	}

	return func(value string) error {
		v, err := parse(value)
		switch {
		case errors.Is(err, errNotFinite):
			return fmt.Errorf("%s: value %q is not a finite number", s.Key, value)
		case err != nil:
			return fmt.Errorf("%s: invalid value %q, want %s", s.Key, value, s.Doc)
		case v < floor:
			return fmt.Errorf("%s: value %q is below the floor %s", s.Key, value, s.Floor)
		case v > ceiling:
			return fmt.Errorf("%s: value %q is above the ceiling %s", s.Key, value, s.Ceiling)
		}
		return nil
	}
}

// hasSetting reports whether name is a setting registered on r. A
// set setting is a declared name: its environment value is a seed the
// seed evaluation parses and bounds, so [Load] neither flags nor
// parses it.
func (r *Registry) hasSetting(name string) bool {
	return slices.ContainsFunc(r.settings, func(s Setting) bool { return s.Key == name })
}
