package config

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// testSetting returns a well-formed integer setting under key, so a
// test can change exactly the field its case is about.
func testSetting(key string) Setting {
	return Setting{
		Key:         key,
		Kind:        KindInt,
		Default:     "14",
		Doc:         "an integer",
		Apply:       ApplyNextOperation,
		Floor:       "3",
		Ceiling:     "1000",
		Description: testDescription,
		Mark:        NotSensitive,
	}
}

func TestAddSetting_ListedInDeclarationOrder(t *testing.T) {
	r := NewRegistry()
	r.AddSetting(testSetting("DUBPLATE_TEST_SETTING_B"))
	r.AddSetting(testSetting("DUBPLATE_TEST_SETTING_A"))
	r.AddSetting(testSetting("DUBPLATE_TEST_SETTING_C"))

	var got []string
	for _, s := range r.Settings() {
		got = append(got, s.Key)
	}
	want := []string{"DUBPLATE_TEST_SETTING_B", "DUBPLATE_TEST_SETTING_A", "DUBPLATE_TEST_SETTING_C"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Settings() keys = %v, want %v", got, want)
	}
}

func TestSettings_ReturnsCopy(t *testing.T) {
	r := NewRegistry()
	r.AddSetting(testSetting("DUBPLATE_TEST_SETTING_COPY"))

	got := r.Settings()
	got[0].Key = "DUBPLATE_MUTATED"
	if r.Settings()[0].Key != "DUBPLATE_TEST_SETTING_COPY" {
		t.Error("mutating Settings()'s result changed the registry")
	}
}

// TestAddSetting_Panics fires every malformed declaration the family
// refuses at registration. Each want names the key and the field or
// the other family, so a panic for the wrong reason fails the row.
func TestAddSetting_Panics(t *testing.T) {
	const key = "DUBPLATE_TEST_SETTING"
	with := func(edit func(*Setting)) Setting {
		s := testSetting(key)
		edit(&s)
		return s
	}

	tests := []struct {
		name  string
		setup func(r *Registry)
		s     Setting
		want  []string
	}{
		{
			name: "a key already a threshold",
			setup: func(r *Registry) {
				r.AddThreshold(Threshold{Key: key, Description: testDescription, Mark: NotSensitive, Default: "1"})
			},
			s:    testSetting(key),
			want: []string{key, familyThresholds, familySettings},
		},
		{
			name:  "a key already an owned wall row",
			setup: func(r *Registry) { r.AddWallRow(WallRow{Name: key, Description: testDescription, Mark: NotSensitive}) },
			s:     testSetting(key),
			want:  []string{key, familyOwnedWall, familySettings},
		},
		{
			name:  "a key already a reserved class key",
			setup: func(r *Registry) { r.AddReservedClass(ReservedClass{Name: "test", Rung: 99, Keys: []string{key}}) },
			s:     testSetting(key),
			want:  []string{key, familyReserved, familySettings},
		},
		{
			name: "a key already an integration class key",
			setup: func(r *Registry) {
				r.AddIntegrationClass(IntegrationClass{Name: "test", Flag: testKey("DUBPLATE_TEST_FLAG", NotSensitive), Required: []KeyDecl{testKey(key, NotSensitive)}})
			},
			s:    testSetting(key),
			want: []string{key, familyIntegration, familySettings},
		},
		{
			name:  "a key already a superseded name",
			setup: func(r *Registry) { r.AddSuperseded(key, "retired") },
			s:     testSetting(key),
			want:  []string{key, familySuperseded, familySettings},
		},
		{
			name:  "a key already a setting",
			setup: func(r *Registry) { r.AddSetting(testSetting(key)) },
			s:     testSetting(key),
			want:  []string{key, familySettings},
		},
		{
			name: "a key without the prefix",
			s:    testSetting("BAD_NO_PREFIX"),
			want: []string{"BAD_NO_PREFIX", "without the DUBPLATE_ prefix"},
		},
		{
			name: "a default that does not parse under its kind",
			s:    with(func(s *Setting) { s.Default = "0.5" }),
			want: []string{key, `invalid default "0.5"`},
		},
		{
			name: "a floor that does not parse",
			s:    with(func(s *Setting) { s.Floor = "three" }),
			want: []string{key, `invalid floor "three"`},
		},
		{
			name: "a ceiling that does not parse",
			s:    with(func(s *Setting) { s.Ceiling = "" }),
			want: []string{key, `invalid ceiling ""`},
		},
		{
			name: "a floor above the ceiling",
			s:    with(func(s *Setting) { s.Floor, s.Ceiling, s.Default = "10", "5", "7" }),
			want: []string{key, "floor 10 is above the ceiling 5"},
		},
		{
			name: "a default below its own floor",
			s:    with(func(s *Setting) { s.Default = "2" }),
			want: []string{key, "invalid default", "below the floor 3"},
		},
		{
			name: "a default above its own ceiling",
			s:    with(func(s *Setting) { s.Default = "1001" }),
			want: []string{key, "invalid default", "above the ceiling 1000"},
		},
		{
			name: "a NaN float default inside finite bounds",
			s: with(func(s *Setting) {
				s.Kind, s.Floor, s.Ceiling, s.Default = KindFloat, "0.25", "1", "NaN"
			}),
			want: []string{key, "invalid default", `"NaN" is not a finite number`},
		},
		{
			name: "a float bound that is not finite",
			s: with(func(s *Setting) {
				s.Kind, s.Floor, s.Ceiling, s.Default = KindFloat, "0.25", "Inf", "0.5"
			}),
			want: []string{key, `invalid ceiling "Inf"`},
		},
		{
			name: "a zero apply mode",
			s:    with(func(s *Setting) { s.Apply = 0 }),
			want: []string{key, "no apply mode"},
		},
		{
			name: "an apply mode no one named",
			s:    with(func(s *Setting) { s.Apply = ApplyMode(99) }),
			want: []string{key, "ApplyMode(99)"},
		},
		{
			name: "a string kind with no allowed values",
			s: with(func(s *Setting) {
				s.Kind, s.Floor, s.Ceiling, s.Default = KindString, "", "", "kind"
			}),
			want: []string{key, "no allowed values"},
		},
		{
			name: "a string kind whose default is outside its values",
			s: with(func(s *Setting) {
				s.Kind, s.Floor, s.Ceiling, s.Default, s.Values = KindString, "", "", "open", []string{"kind", "generic"}
			}),
			want: []string{key, `invalid default "open"`, "one of kind, generic"},
		},
		{
			name: "a string kind carrying a floor",
			s: with(func(s *Setting) {
				s.Kind, s.Ceiling, s.Default, s.Values = KindString, "", "kind", []string{"kind", "generic"}
			}),
			want: []string{key, "floor or ceiling"},
		},
		{
			name: "a bounded kind carrying allowed values",
			s:    with(func(s *Setting) { s.Values = []string{"14"} }),
			want: []string{key, "allowed values"},
		},
		{
			name: "a kind with no parser",
			s:    with(func(s *Setting) { s.Kind = ThresholdKind(99) }),
			want: []string{key, "ThresholdKind(99)", "has no bounds"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			if tt.setup != nil {
				tt.setup(r)
			}
			msg := panicMessage(t, func() { r.AddSetting(tt.s) })
			for _, want := range tt.want {
				if !strings.Contains(msg, want) {
					t.Errorf("panic message %q, want it to contain %q", msg, want)
				}
			}
		})
	}
}

// TestAddThreshold_RefusesASettingKey is the collision in the order the
// table above cannot reach: a setting registered first, then a
// threshold on the same key.
func TestAddThreshold_RefusesASettingKey(t *testing.T) {
	r := NewRegistry()
	r.AddSetting(testSetting("DUBPLATE_TEST_SETTING"))

	msg := panicMessage(t, func() {
		r.AddThreshold(Threshold{Key: "DUBPLATE_TEST_SETTING", Description: testDescription, Mark: NotSensitive, Default: "1"})
	})
	for _, want := range []string{"DUBPLATE_TEST_SETTING", familySettings, familyThresholds} {
		if !strings.Contains(msg, want) {
			t.Errorf("panic message %q, want it to contain %q", msg, want)
		}
	}
}

// TestSettingValidate proves the validator AddSetting builds from each
// kind's bounds data: a value at the floor and at the ceiling is
// accepted, one step past either is refused naming that bound, and an
// unparseable or non-finite value is refused naming the value. The
// bounds are the three backup settings' protective ranges (keep 3 to
// 1000, archive retention 168h to 87600h, guard fraction 0.25 to 1).
func TestSettingValidate(t *testing.T) {
	keep := Setting{Key: "DUBPLATE_TEST_KEEP", Description: testDescription, Mark: NotSensitive, Kind: KindInt, Default: "14", Doc: "an integer",
		Apply: ApplyNextOperation, Floor: "3", Ceiling: "1000"}
	retention := Setting{Key: "DUBPLATE_TEST_RETENTION", Description: testDescription, Mark: NotSensitive, Kind: KindDuration, Default: "2160h", Doc: "a duration",
		Apply: ApplyNextOperation, Floor: "168h", Ceiling: "87600h"}
	fraction := Setting{Key: "DUBPLATE_TEST_FRACTION", Description: testDescription, Mark: NotSensitive, Kind: KindFloat, Default: "0.5", Doc: "a number",
		Apply: ApplyNextOperation, Floor: "0.25", Ceiling: "1"}
	answer := Setting{Key: "DUBPLATE_TEST_ANSWER", Description: testDescription, Mark: NotSensitive, Kind: KindString, Default: "kind", Doc: "one of kind or generic",
		Apply: ApplyNextOperation, Values: []string{"kind", "generic"}}

	r := NewRegistry()
	r.AddSetting(keep)
	r.AddSetting(retention)
	r.AddSetting(fraction)
	r.AddSetting(answer)
	registered := make(map[string]Setting)
	for _, s := range r.Settings() {
		registered[s.Key] = s
	}

	tests := []struct {
		name  string
		key   string
		value string
		want  string // empty: accepted
	}{
		{name: "int at the floor", key: keep.Key, value: "3"},
		{name: "int at the ceiling", key: keep.Key, value: "1000"},
		{name: "int one below the floor", key: keep.Key, value: "2", want: `DUBPLATE_TEST_KEEP: value "2" is below the floor 3`},
		{name: "int one above the ceiling", key: keep.Key, value: "1001", want: `DUBPLATE_TEST_KEEP: value "1001" is above the ceiling 1000`},
		{name: "int negative", key: keep.Key, value: "-1", want: "below the floor 3"},
		{name: "int unparseable", key: keep.Key, value: "many", want: `DUBPLATE_TEST_KEEP: invalid value "many", want an integer`},

		{name: "duration at the floor", key: retention.Key, value: "168h"},
		{name: "duration at the floor in other units", key: retention.Key, value: "10080m"},
		{name: "duration at the ceiling", key: retention.Key, value: "87600h"},
		{name: "duration just below the floor", key: retention.Key, value: "167h59m59s", want: "below the floor 168h"},
		{name: "duration just above the ceiling", key: retention.Key, value: "87600h0m1s", want: "above the ceiling 87600h"},
		{name: "duration with a day unit", key: retention.Key, value: "90d", want: `invalid value "90d", want a duration`},

		{name: "float at the floor", key: fraction.Key, value: "0.25"},
		{name: "float at the ceiling", key: fraction.Key, value: "1"},
		{name: "float just below the floor", key: fraction.Key, value: "0.2499", want: "below the floor 0.25"},
		{name: "float just above the ceiling", key: fraction.Key, value: "1.0001", want: "above the ceiling 1"},
		{name: "float NaN", key: fraction.Key, value: "NaN", want: `DUBPLATE_TEST_FRACTION: value "NaN" is not a finite number`},
		{name: "float Inf", key: fraction.Key, value: "Inf", want: `value "Inf" is not a finite number`},
		{name: "float -Inf", key: fraction.Key, value: "-Inf", want: `value "-Inf" is not a finite number`},
		{name: "float unparseable", key: fraction.Key, value: "half", want: `invalid value "half", want a number`},

		{name: "string, the default value", key: answer.Key, value: "kind"},
		{name: "string, the other allowed value", key: answer.Key, value: "generic"},
		{name: "string outside the set", key: answer.Key, value: "open", want: `DUBPLATE_TEST_ANSWER: value "open" is not one of kind, generic`},
		{name: "string differing only in case", key: answer.Key, value: "Generic", want: "not one of kind, generic"},
		{name: "string empty", key: answer.Key, value: "", want: "not one of kind, generic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registered[tt.key].Validate(tt.value)
			if tt.want == "" {
				if err != nil {
					t.Errorf("Validate(%q) = %v, want nil", tt.value, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate(%q) = nil, want an error containing %q", tt.value, tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Validate(%q) = %v, want it to contain %q", tt.value, err, tt.want)
			}
		})
	}
}

// TestSettingValidate_Unregistered proves a Setting that never passed
// through AddSetting carries no validator and refuses every value,
// in-bounds ones included, rather than accepting them unchecked.
func TestSettingValidate_Unregistered(t *testing.T) {
	err := testSetting("DUBPLATE_TEST_UNREGISTERED").Validate("14")
	if err == nil {
		t.Fatal("Validate on an unregistered Setting = nil, want an error")
	}
	if !strings.Contains(err.Error(), "DUBPLATE_TEST_UNREGISTERED: not a registered setting") {
		t.Errorf("Validate error = %v, want it to name the key as unregistered", err)
	}
}

func TestApplyMode_String(t *testing.T) {
	tests := []struct {
		mode ApplyMode
		want string
	}{
		{ApplyImmediately, "immediately"},
		{ApplyNextOperation, "next operation"},
		{ApplyRestart, "restart"},
		{0, "ApplyMode(0)"},
		{ApplyMode(7), "ApplyMode(7)"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("ApplyMode(%d).String() = %q, want %q", int(tt.mode), got, tt.want)
		}
	}
}

// TestThreshold_StillResolvesFromTheEnvironment proves a threshold
// keeps reading from the environment through ParseDuration once a
// registry also carries store-homed settings: registering the new
// family changes nothing about the wall's own tunables.
func TestThreshold_StillResolvesFromTheEnvironment(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_TEST_WALL_WINDOW", Description: testDescription, Mark: NotSensitive, Kind: KindDuration, Default: "15m", Doc: "a duration"}
	r := NewRegistry()
	r.AddThreshold(th)
	r.AddSetting(testSetting("DUBPLATE_TEST_SETTING"))

	got, err := ParseDuration([]string{"DUBPLATE_TEST_WALL_WINDOW=45m"}, th)
	if err != nil {
		t.Fatalf("ParseDuration() error = %v", err)
	}
	if got != 45*time.Minute {
		t.Errorf("ParseDuration() = %v, want 45m from the environment", got)
	}
}

// ruleRegistry returns a registry carrying two settings a correlated
// rule can compare, plus a threshold a rule must refuse.
func ruleRegistry() *Registry {
	r := NewRegistry()
	r.AddSetting(Setting{Key: "DUBPLATE_TEST_EXPIRY", Description: testDescription, Mark: NotSensitive, Kind: KindDuration, Default: "15m", Doc: "a duration",
		Apply: ApplyNextOperation, Floor: "1m", Ceiling: "24h"})
	r.AddSetting(Setting{Key: "DUBPLATE_TEST_RETAIN", Description: testDescription, Mark: NotSensitive, Kind: KindDuration, Default: "1h", Doc: "a duration",
		Apply: ApplyNextOperation, Floor: "1m", Ceiling: "720h"})
	r.AddThreshold(Threshold{Key: "DUBPLATE_TEST_WALL", Description: testDescription, Mark: NotSensitive, Kind: KindInt, Default: "1"})
	return r
}

func noopCheck(map[string]string) error { return nil }

func TestAddCorrelatedRule_Registered(t *testing.T) {
	r := ruleRegistry()
	errCheck := errors.New("retain below expiry")
	r.AddCorrelatedRule(CorrelatedRule{
		Name:  "retain covers expiry",
		Keys:  []string{"DUBPLATE_TEST_RETAIN", "DUBPLATE_TEST_EXPIRY"},
		Check: func(map[string]string) error { return errCheck },
	})

	rules := r.CorrelatedRules()
	if len(rules) != 1 || rules[0].Name != "retain covers expiry" {
		t.Fatalf("CorrelatedRules() = %+v, want the one registered rule", rules)
	}
	if err := rules[0].Check(nil); !errors.Is(err, errCheck) {
		t.Errorf("registered rule's Check() = %v, want the declared function's error", err)
	}
}

func TestAddCorrelatedRule_Panics(t *testing.T) {
	tests := []struct {
		name string
		rule CorrelatedRule
		want []string
	}{
		{
			name: "a key no family registered",
			rule: CorrelatedRule{Name: "r", Keys: []string{"DUBPLATE_TEST_RETAIN", "DUBPLATE_TEST_ABSENT"}, Check: noopCheck},
			want: []string{"DUBPLATE_TEST_ABSENT", "not a registered setting"},
		},
		{
			name: "a threshold key",
			rule: CorrelatedRule{Name: "r", Keys: []string{"DUBPLATE_TEST_WALL", "DUBPLATE_TEST_RETAIN"}, Check: noopCheck},
			want: []string{"DUBPLATE_TEST_WALL", "not a registered setting", familyThresholds},
		},
		{
			name: "one key",
			rule: CorrelatedRule{Name: "r", Keys: []string{"DUBPLATE_TEST_RETAIN"}, Check: noopCheck},
			want: []string{`correlated rule "r"`, "two or more"},
		},
		{
			name: "a key repeated",
			rule: CorrelatedRule{Name: "r", Keys: []string{"DUBPLATE_TEST_RETAIN", "DUBPLATE_TEST_RETAIN"}, Check: noopCheck},
			want: []string{`correlated rule "r"`, "DUBPLATE_TEST_RETAIN", "more than once"},
		},
		{
			name: "no check",
			rule: CorrelatedRule{Name: "r", Keys: []string{"DUBPLATE_TEST_RETAIN", "DUBPLATE_TEST_EXPIRY"}},
			want: []string{`correlated rule "r"`, "no Check"},
		},
		{
			name: "no name",
			rule: CorrelatedRule{Keys: []string{"DUBPLATE_TEST_RETAIN", "DUBPLATE_TEST_EXPIRY"}, Check: noopCheck},
			want: []string{"empty Name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ruleRegistry()
			msg := panicMessage(t, func() { r.AddCorrelatedRule(tt.rule) })
			for _, want := range tt.want {
				if !strings.Contains(msg, want) {
					t.Errorf("panic message %q, want it to contain %q", msg, want)
				}
			}
		})
	}
}

func TestAddCorrelatedRule_DuplicateName(t *testing.T) {
	r := ruleRegistry()
	rule := CorrelatedRule{Name: "retain covers expiry", Keys: []string{"DUBPLATE_TEST_RETAIN", "DUBPLATE_TEST_EXPIRY"}, Check: noopCheck}
	r.AddCorrelatedRule(rule)

	msg := panicMessage(t, func() { r.AddCorrelatedRule(rule) })
	if !strings.Contains(msg, `correlated rule "retain covers expiry" already declared`) {
		t.Errorf("panic message %q, want it to name the duplicate rule", msg)
	}
}
