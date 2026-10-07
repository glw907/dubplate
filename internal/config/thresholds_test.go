package config

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestRegister_AppearsInThresholds proves AddThreshold accumulates in
// registration order onto a registry a test constructs itself: rung 1
// registers no package-level threshold, so a test that needs one builds
// its own registry rather than saving and restoring shared state.
func TestRegister_AppearsInThresholds(t *testing.T) {
	reg := NewRegistry()
	if got := reg.Thresholds(); len(got) != 0 {
		t.Fatalf("Thresholds() before AddThreshold = %v, want empty", got)
	}

	first := Threshold{Key: "DUBPLATE_TEST_FIRST", Default: "1", Doc: "a positive integer"}
	second := Threshold{Key: "DUBPLATE_TEST_SECOND", Default: "2", Doc: "a positive integer"}
	reg.AddThreshold(first)
	reg.AddThreshold(second)

	got := reg.Thresholds()
	if len(got) != 2 {
		t.Fatalf("Thresholds() len = %d, want 2", len(got))
	}
	if got[0] != first || got[1] != second {
		t.Errorf("Thresholds() = %+v, want [%+v %+v] in registration order", got, first, second)
	}
}

// TestThresholds_ReturnsCopy proves the copy property [Registry.OwnedWall]
// already carries survives the accessor's move onto the registry:
// mutating a returned slice must not corrupt what a second call returns.
func TestThresholds_ReturnsCopy(t *testing.T) {
	reg := NewRegistry()
	reg.AddThreshold(Threshold{Key: "DUBPLATE_TEST_ONE", Default: "1", Doc: "a positive integer"})

	got := reg.Thresholds()
	got[0].Key = "mutated"

	if reg.thresholds[0].Key == "mutated" {
		t.Fatal("Thresholds() leaked the backing array; a caller can corrupt the registry")
	}
}

func TestParseInt(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_TEST_QUIESCENT_MIN", Default: "10", Doc: "a positive integer number of minutes"}

	tests := []struct {
		name    string
		environ []string
		want    int
		wantErr bool
		errSub  string
	}{
		{
			name:    "unset uses default",
			environ: nil,
			want:    10,
		},
		{
			name:    "empty value uses default",
			environ: []string{"DUBPLATE_TEST_QUIESCENT_MIN="},
			want:    10,
		},
		{
			name:    "set value parses",
			environ: []string{"DUBPLATE_TEST_QUIESCENT_MIN=30"},
			want:    30,
		},
		{
			name:    "malformed value errors naming key and doc",
			environ: []string{"DUBPLATE_TEST_QUIESCENT_MIN=soon"},
			wantErr: true,
			errSub:  "DUBPLATE_TEST_QUIESCENT_MIN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.environ, th)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseInt() error = nil, want error containing %q", tt.errSub)
				}
				if got := err.Error(); !strings.Contains(got, tt.errSub) || !strings.Contains(got, th.Doc) {
					t.Errorf("ParseInt() error = %q, want it to contain %q and %q", got, tt.errSub, th.Doc)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseInt() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("ParseInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseInt_InvalidDefaultIsRegistrationError(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_TEST_BROKEN_DEFAULT", Default: "not-a-number", Doc: "a positive integer"}

	_, err := ParseInt(nil, th)
	if err == nil {
		t.Fatal("ParseInt() error = nil, want error for an unparsable Default")
	}
	if !strings.Contains(err.Error(), th.Key) {
		t.Errorf("ParseInt() error = %q, want it to contain %q", err.Error(), th.Key)
	}
}

func TestParseDuration(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_TEST_WINDOW", Kind: KindDuration, Default: "720h", Doc: "a Go duration"}
	tests := []struct {
		name    string
		environ []string
		want    time.Duration
		wantErr bool
	}{
		{name: "unset takes the default", want: 720 * time.Hour},
		{name: "empty takes the default", environ: []string{"DUBPLATE_TEST_WINDOW="}, want: 720 * time.Hour},
		{name: "set wins", environ: []string{"DUBPLATE_TEST_WINDOW=48h"}, want: 48 * time.Hour},
		{name: "a bad value errors", environ: []string{"DUBPLATE_TEST_WINDOW=soon"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.environ, th)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseDuration() error = nil, want error")
				}
				if !strings.Contains(err.Error(), th.Key) || !strings.Contains(err.Error(), th.Doc) {
					t.Errorf("ParseDuration() error = %q, want it to contain %q and %q", err.Error(), th.Key, th.Doc)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDuration() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDurationRejectsABadDefaultAsARegistrationBug(t *testing.T) {
	_, err := ParseDuration(nil, Threshold{Key: "K", Kind: KindDuration, Default: "not a duration"})
	if err == nil || !strings.Contains(err.Error(), "K") {
		t.Errorf("ParseDuration on a bad default = %v; want an error naming the key", err)
	}
}

// TestKindInt_IsZeroValue pins Resolution R11: KindInt is iota's zero,
// so a Threshold literal that omits Kind keeps meaning KindInt across
// every rung that ever wrote one before KindDuration, KindFloat and
// KindString existed. Named mutation 4 reorders the const block so
// KindFloat sits at position zero instead; this test is the direct
// proof that the reorder changes every such literal's meaning.
func TestKindInt_IsZeroValue(t *testing.T) {
	if KindInt != 0 {
		t.Fatalf("KindInt = %d, want 0", KindInt)
	}
	var zero ThresholdKind
	if zero != KindInt {
		t.Fatalf("the zero ThresholdKind = %v, want KindInt", zero)
	}
}

// TestThresholdKinds is the one test-row table the four parsers share:
// one row per kind, and per row, the same four cases every parser must
// answer identically. wantErrSub is empty when the case wants no error.
//
// The set+invalid-value case is n/a for KindString, because no string
// value can fail ParseString; that row's third cell is skipped rather
// than forced into a shape it cannot take, and the row's fourth cell
// (its one invalid Default, the empty string) is the case that still
// proves it distinct from a merely-absent third cell.
func TestThresholdKinds(t *testing.T) {
	tests := []struct {
		name           string
		kind           ThresholdKind
		doc            string
		validDefault   string
		wantDefault    string // string form validDefault parses to
		validValue     string
		wantValid      string // string form validValue parses to
		invalidValue   string // "", meaning n/a, for KindString
		invalidDefault string
		parse          func(environ []string, th Threshold) (string, error)
	}{
		{
			name:           "KindInt",
			kind:           KindInt,
			doc:            "an integer",
			validDefault:   "10",
			wantDefault:    "10",
			validValue:     "30",
			wantValid:      "30",
			invalidValue:   "soon",
			invalidDefault: "not-a-number",
			parse: func(environ []string, th Threshold) (string, error) {
				v, err := ParseInt(environ, th)
				return strconv.Itoa(v), err
			},
		},
		{
			name:           "KindDuration",
			kind:           KindDuration,
			doc:            "a Go duration",
			validDefault:   "10m",
			wantDefault:    "10m0s",
			validValue:     "30m",
			wantValid:      "30m0s",
			invalidValue:   "soon",
			invalidDefault: "not a duration",
			parse: func(environ []string, th Threshold) (string, error) {
				d, err := ParseDuration(environ, th)
				return d.String(), err
			},
		},
		{
			name:           "KindFloat",
			kind:           KindFloat,
			doc:            "a float",
			validDefault:   "0.1",
			wantDefault:    "0.1",
			validValue:     "0.04",
			wantValid:      "0.04",
			invalidValue:   "soon",
			invalidDefault: "not-a-float",
			parse: func(environ []string, th Threshold) (string, error) {
				f, err := ParseFloat(environ, th)
				return strconv.FormatFloat(f, 'g', -1, 64), err
			},
		},
		{
			name:         "KindString",
			kind:         KindString,
			doc:          "a string",
			validDefault: "default-value",
			wantDefault:  "default-value",
			validValue:   "set-value",
			wantValid:    "set-value",
			// invalidValue is n/a: no string fails ParseString.
			invalidDefault: "",
			parse: func(environ []string, th Threshold) (string, error) {
				return ParseString(environ, th)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "DUBPLATE_TEST_" + tt.name

			// unset, valid default.
			th := Threshold{Key: key, Kind: tt.kind, Default: tt.validDefault, Doc: tt.doc}
			got, err := tt.parse(nil, th)
			if err != nil {
				t.Fatalf("unset+valid-default: parse() error = %v, want nil", err)
			}
			if got != tt.wantDefault {
				t.Errorf("unset+valid-default: parse() = %q, want %q", got, tt.wantDefault)
			}

			// set, valid value.
			got, err = tt.parse([]string{key + "=" + tt.validValue}, th)
			if err != nil {
				t.Fatalf("set+valid-value: parse() error = %v, want nil", err)
			}
			if got != tt.wantValid {
				t.Errorf("set+valid-value: parse() = %q, want %q", got, tt.wantValid)
			}

			// set, invalid value: n/a for KindString.
			if tt.invalidValue != "" {
				_, err := tt.parse([]string{key + "=" + tt.invalidValue}, th)
				if err == nil {
					t.Fatal("set+invalid-value: parse() error = nil, want error naming the key and Doc")
				}
				if !strings.Contains(err.Error(), key) || !strings.Contains(err.Error(), tt.doc) {
					t.Errorf("set+invalid-value: parse() error = %q, want it to contain %q and %q", err.Error(), key, tt.doc)
				}
			}

			// unset, invalid default: a registration bug, naming the
			// key alone. Distinguishable from the set+invalid-value
			// case above because it must not name Doc: an environment
			// error explains the accepted form, a registration error
			// does not, since no environment value is at fault.
			badDefault := Threshold{Key: key, Kind: tt.kind, Default: tt.invalidDefault, Doc: tt.doc}
			_, err = tt.parse(nil, badDefault)
			if err == nil {
				t.Fatal("unset+invalid-default: parse() error = nil, want error naming the key alone")
			}
			if !strings.Contains(err.Error(), key) {
				t.Errorf("unset+invalid-default: parse() error = %q, want it to contain %q", err.Error(), key)
			}
			if tt.invalidValue != "" && strings.Contains(err.Error(), tt.doc) {
				t.Errorf("unset+invalid-default: parse() error = %q, must not contain Doc %q, or it is indistinguishable from a set+invalid-value error", err.Error(), tt.doc)
			}
		})
	}
}

// TestParseFloat_RefusesNonFinite proves a KindFloat threshold refuses
// the three non-finite values strconv.ParseFloat accepts by name, from
// the environment as an invalid value and as a Default as a
// registration bug, the way a KindFloat [Setting] refuses them: no
// bound comparison catches NaN, and an infinity defeats any ratio it
// scales.
func TestParseFloat_RefusesNonFinite(t *testing.T) {
	const key = "DUBPLATE_TEST_FINITE"
	th := Threshold{Key: key, Kind: KindFloat, Default: "0.5", Doc: "a finite number"}
	for _, value := range []string{"NaN", "Inf", "-Inf", "+infinity"} {
		t.Run(value, func(t *testing.T) {
			got, err := ParseFloat([]string{key + "=" + value}, th)
			if err == nil {
				t.Fatalf("ParseFloat(%s=%s) = %v, want an error", key, value, got)
			}
			if !strings.Contains(err.Error(), th.Doc) {
				t.Errorf("ParseFloat(%s=%s) error = %q, want it to name %q", key, value, err, th.Doc)
			}

			bad := Threshold{Key: key, Kind: KindFloat, Default: value, Doc: th.Doc}
			if got, err := ParseFloat(nil, bad); err == nil {
				t.Errorf("ParseFloat with Default %q = %v, want an error", value, got)
			}
			msg := panicMessage(t, func() { NewRegistry().AddThreshold(bad) })
			if want := key + ": invalid default"; !strings.Contains(msg, want) {
				t.Errorf("AddThreshold with Default %q panic = %q, want it to contain %q", value, msg, want)
			}
		})
	}
}
