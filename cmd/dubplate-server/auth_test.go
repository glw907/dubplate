package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
)

// TestRegisterAuthConfig_SelectorIsOwnedNow proves the selector lands as
// an owned wall row: a registry that registers nothing else still
// carries DUBPLATE_IDENTITY_SOURCE on its wall once auth.Register
// has run.
func TestRegisterAuthConfig_SelectorIsOwnedNow(t *testing.T) {
	reg := config.NewRegistry()
	auth.Register(reg)

	if !slices.ContainsFunc(reg.OwnedWall(), func(row config.WallRow) bool { return row.Key == "DUBPLATE_IDENTITY_SOURCE" }) {
		t.Fatal("OwnedWall() has no DUBPLATE_IDENTITY_SOURCE row")
	}
}

// TestRegisterAuthConfig_MagicLinkSource proves the magic-link source
// declares exactly the eight required and two optional names R-5b3
// rules, on the selector auth.Register also registers as a wall
// row.
func TestRegisterAuthConfig_MagicLinkSource(t *testing.T) {
	reg := config.NewRegistry()
	auth.Register(reg)

	sources := reg.IdentitySources()
	if len(sources) != 1 {
		t.Fatalf("IdentitySources() = %d entries, want 1", len(sources))
	}
	src := sources[0]

	if src.Name != "magic-link" {
		t.Errorf("Name = %q, want %q", src.Name, "magic-link")
	}
	if src.Selector != "DUBPLATE_IDENTITY_SOURCE" {
		t.Errorf("Selector = %q, want %q", src.Selector, "DUBPLATE_IDENTITY_SOURCE")
	}

	wantRequired := []string{
		"DUBPLATE_MAGIC_LINK_SECRET",
		"DUBPLATE_SESSION_SECRET",
		"DUBPLATE_PUBLIC_URL",
		"DUBPLATE_SMTP_HOST",
		"DUBPLATE_SMTP_PORT",
		"DUBPLATE_SMTP_USER",
		"DUBPLATE_SMTP_PASSWORD",
		"DUBPLATE_SMTP_FROM",
	}
	if !slices.Equal(config.KeyNames(src.Required), wantRequired) {
		t.Errorf("Required = %v, want %v", src.Required, wantRequired)
	}

	wantOptional := []string{
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS",
		"DUBPLATE_SESSION_SECRET_PREVIOUS",
	}
	if !slices.Equal(config.KeyNames(src.Optional), wantOptional) {
		t.Errorf("Optional = %v, want %v", src.Optional, wantOptional)
	}
}

// TestRegisterAuthConfig_RetiresAuthAndIdentityReservedClass proves
// the removal half of the rung 1b registry contract landed in this
// commit, on TestRegisterServeConfig_ClaimsThreeKeysAndRetiresReservedClass's
// own precedent (serve_test.go): the class auth.Register's
// declarations replace is gone from the production registry.
func TestRegisterAuthConfig_RetiresAuthAndIdentityReservedClass(t *testing.T) {
	reg := productionRegistry()

	if slices.ContainsFunc(reg.ReservedClasses(), func(class config.ReservedClass) bool {
		return class.Name == "auth and identity"
	}) {
		t.Error(`ReservedClasses() still carries the "auth and identity" class, want it retired`)
	}
}

// TestDoctorCmd_UnsetSelectorExitsNonzero proves doctor's exit gate on
// the production registry, rather than only on a scratch one: with
// DUBPLATE_IDENTITY_SOURCE unset, Execute returns a non-nil error and
// stdout carries the selector's own "missing" row.
func TestDoctorCmd_UnsetSelectorExitsNonzero(t *testing.T) {
	environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	cmd := rootFor(environ, productionRegistry(), &stdout, &stderr, "doctor")
	cmd.SetOut(&stdout)

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	wantPrefix := fmt.Sprintf("%-32s %s", "DUBPLATE_IDENTITY_SOURCE", "missing")
	found := false
	for line := range strings.SplitSeq(stdout.String(), "\n") {
		if strings.HasPrefix(line, wantPrefix) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("stdout = %q, want a line with prefix %q", stdout.String(), wantPrefix)
	}
}

// TestRegisterAuthConfig_MagicLinkKeyWithoutSelectorNamesSelector
// proves the declared-name discipline against the production
// registry: setting a magic-link source key while the selector is
// unset is a load error naming DUBPLATE_IDENTITY_SOURCE, never an
// "unknown configuration key" error, because rung 5's real names must
// be declared as a source's Required/Optional rows rather than as
// plain wall or threshold entries.
func TestRegisterAuthConfig_MagicLinkKeyWithoutSelectorNamesSelector(t *testing.T) {
	environ := append(baseEnviron(),
		"DUBPLATE_MAGIC_LINK_SECRET=fixture magic link secret, not a real key")

	_, err := config.Load(environ, productionRegistry())
	if err == nil {
		t.Fatal("Load() error = nil, want error naming DUBPLATE_IDENTITY_SOURCE")
	}
	if !strings.Contains(err.Error(), "DUBPLATE_IDENTITY_SOURCE") {
		t.Errorf("Load() error = %q, want it to name DUBPLATE_IDENTITY_SOURCE", err.Error())
	}
	if strings.Contains(err.Error(), "unknown configuration key") {
		t.Errorf("Load() error = %q, want no unknown-key error", err.Error())
	}
}

// fixtureSecretBelowFloor is 31 bytes, one under the 32-byte secret
// floor. It is built from low-entropy filler so gitleaks stays quiet
// and no literal here is shaped like a real key.
var fixtureSecretBelowFloor = strings.Repeat("ab", 16)[:31]

// TestConfigLoad_MagicLinkValidationJoinsOtherErrors proves the hook
// runs inside config.Load and joins its refusal with Load's own: an
// unset DUBPLATE_LIBRARY_ROOT and a too-short magic-link secret both
// name themselves in one returned error.
func TestConfigLoad_MagicLinkValidationJoinsOtherErrors(t *testing.T) {
	environ := dropKey(authEnviron(t), "DUBPLATE_LIBRARY_ROOT")
	environ = append(dropKey(environ, "DUBPLATE_MAGIC_LINK_SECRET"),
		"DUBPLATE_MAGIC_LINK_SECRET="+fixtureSecretBelowFloor)

	_, err := config.Load(environ, productionRegistry())
	if err == nil {
		t.Fatal("Load() error = nil, want a joined refusal")
	}
	for _, want := range []string{"DUBPLATE_LIBRARY_ROOT", "DUBPLATE_MAGIC_LINK_SECRET"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Load() error = %q, want it to name %s", err.Error(), want)
		}
	}
}

// rosterSupersessionMessage is the wording every wall-loading verb
// reports for a retired DUBPLATE_ROSTER_FILE; the test holds it as a
// literal so the registration cannot be compared against itself.
const rosterSupersessionMessage = "DUBPLATE_ROSTER_FILE: superseded: members live in the store; " +
	"add the first admin with dubplate-server member add ADDRESS --admin"

// TestRosterFileKey_Superseded runs the ten verbs that load the wall
// with the retired key set to a real, valid roster file, and asserts
// each fails the load with the supersession message: a check that
// refused only a missing file would let the fixture through. The
// leaf walk below fails when a verb is added without being classified
// as a loader or a non-loader.
func TestRosterFileKey_Superseded(t *testing.T) {
	rosterPath := filepath.Join(t.TempDir(), "roster.json")
	if err := os.WriteFile(rosterPath, []byte(`{"schema_version":1,"members":[]}`), 0o600); err != nil {
		t.Fatalf("write roster fixture: %v", err)
	}

	verbs := []struct {
		name string
		args []string
	}{
		{"check", []string{"check"}},
		{"doctor", []string{"doctor"}},
		{"serve", []string{"serve"}},
		{"member add", []string{"member", "add", "a@example.test", "--name", "A", "--username", "a"}},
		{"member remove", []string{"member", "remove", "a@example.test"}},
		{"member demote", []string{"member", "demote", "a@example.test"}},
		{"member list", []string{"member", "list"}},
		{"settings reset", []string{"settings", "reset", "DUBPLATE_JOB_RETENTION"}},
		{"backup", []string{"backup"}},
		{"restore", []string{"restore", "--snapshot", "latest"}},
	}

	for _, value := range []struct{ name, set string }{
		{"a real file", rosterPath},
		{"an absent path", filepath.Join(t.TempDir(), "absent.json")},
		{"a malformed value", "not a path\x01"},
	} {
		t.Run(value.name, func(t *testing.T) {
			environ := append(doctorEnviron(t), "DUBPLATE_ROSTER_FILE="+value.set)
			for _, v := range verbs {
				t.Run(v.name, func(t *testing.T) {
					var stdout, stderr bytes.Buffer
					err := rootFor(environ, productionRegistry(), &stdout, &stderr, v.args...).Execute()

					if v.name == "doctor" {
						// doctor prints the load error no row claims.
						if err == nil {
							t.Fatal("doctor error = nil, want the incomplete-configuration failure")
						}
						line := fmt.Sprintf("%-32s %s", "DUBPLATE_ROSTER_FILE", "invalid: superseded: members live in the store; "+
							"add the first admin with dubplate-server member add ADDRESS --admin")
						if !slices.Contains(strings.Split(stdout.String(), "\n"), line) {
							t.Errorf("doctor output lacks the unclaimed-error line %q:\n%s", line, stdout.String())
						}
						return
					}
					if err == nil || !strings.Contains(err.Error(), rosterSupersessionMessage) {
						t.Errorf("%s error = %v, want it to carry %q", v.name, err, rosterSupersessionMessage)
					}
				})
			}
		})
	}

	t.Run("an empty value loads", func(t *testing.T) {
		environ := append(doctorEnviron(t), "DUBPLATE_ROSTER_FILE=")
		if _, err := config.Load(environ, productionRegistry()); err != nil {
			t.Errorf("config.Load error = %v, want the empty value treated as unset", err)
		}
	})

	t.Run("the command tree is classified", func(t *testing.T) {
		var leaves []string
		var walk func(c *cobra.Command)
		walk = func(c *cobra.Command) {
			if len(c.Commands()) == 0 {
				leaves = append(leaves, strings.TrimPrefix(c.CommandPath(), "dubplate-server "))
				return
			}
			for _, sub := range c.Commands() {
				walk(sub)
			}
		}
		walk(newRootCmd(nil, productionRegistry()))
		slices.Sort(leaves)

		want := []string{"backup", "check", "doctor", "member add", "member demote", "member list", "member remove",
			"restore", "serve", "settings reset", "unidecode install", "version"}
		if !slices.Equal(leaves, want) {
			t.Errorf("leaf verbs = %q, want %q: classify a new verb as a wall loader (add it to verbs) or a non-loader", leaves, want)
		}
	})
}

// TestRegisterAuthConfig_Settings proves the ten auth settings carry
// their literal keys, kinds, defaults, bounds, apply modes, and
// sensitive marks, asserted against fixed literals rather than read
// back from the declarations, on a fresh registry and on the binary's
// own. None of the ten may remain a threshold: a key that was both
// would be read from the environment at its point of use, past the
// seed rule.
func TestRegisterAuthConfig_Settings(t *testing.T) {
	tests := []struct {
		key            string
		kind           config.ThresholdKind
		def            string
		floor, ceiling string
		values         []string
		apply          config.ApplyMode
		sensitive      bool
	}{
		{"DUBPLATE_MAGIC_LINK_EXPIRY", config.KindDuration, "15m", "1m", "1h0m0s", nil, config.ApplyNextOperation, true},
		{"DUBPLATE_SESSION_LIFETIME", config.KindDuration, "2160h", "1s", "17520h", nil, config.ApplyNextOperation, true},
		{"DUBPLATE_SESSION_ABSOLUTE_CAP", config.KindDuration, "4320h", "1s", "17520h", nil, config.ApplyImmediately, true},
		{"DUBPLATE_NONCE_RETENTION", config.KindDuration, "24h", "1s", "240h", nil, config.ApplyNextOperation, false},
		{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", config.KindInt, "5", "1", "50", nil, config.ApplyRestart, false},
		{"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", config.KindInt, "20", "1", "200", nil, config.ApplyRestart, false},
		{"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", config.KindInt, "30", "1", "300", nil, config.ApplyRestart, false},
		{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", config.KindInt, "60", "1", "600", nil, config.ApplyRestart, false},
		{"DUBPLATE_AUTH_REQUEST_BODY_CAP", config.KindInt, "4096", "512", "65536", nil, config.ApplyNextOperation, false},
		{"DUBPLATE_UNKNOWN_ADDRESS_ANSWER", config.KindString, "kind", "", "", []string{"kind", "generic"}, config.ApplyNextOperation, true},
	}

	fresh := config.NewRegistry()
	auth.Register(fresh)
	for regName, reg := range map[string]*config.Registry{"fresh": fresh, "production": productionRegistry()} {
		settingsList := reg.Settings()
		for _, tt := range tests {
			t.Run(regName+"/"+tt.key, func(t *testing.T) {
				i := slices.IndexFunc(settingsList, func(s config.Setting) bool { return s.Key == tt.key })
				if i < 0 {
					t.Fatalf("Settings() has no %s", tt.key)
				}
				s := settingsList[i]
				if s.Kind != tt.kind || s.Default != tt.def || s.Floor != tt.floor || s.Ceiling != tt.ceiling ||
					!slices.Equal(s.Values, tt.values) || s.Apply != tt.apply || (s.Sensitivity == config.Sensitive) != tt.sensitive {
					t.Errorf("setting = kind %v default %q floor %q ceiling %q values %q apply %v sensitive %v, want %v %q %q %q %q %v %v",
						s.Kind, s.Default, s.Floor, s.Ceiling, s.Values, s.Apply, s.Sensitivity == config.Sensitive,
						tt.kind, tt.def, tt.floor, tt.ceiling, tt.values, tt.apply, tt.sensitive)
				}
				if slices.ContainsFunc(reg.Thresholds(), func(th config.Threshold) bool { return th.Key == tt.key }) {
					t.Errorf("%s is still a threshold", tt.key)
				}
			})
		}
	}
}

// TestAuthSettings_Bounds walks every auth setting's bound through the
// validator the production registry built: a value at each bound is
// accepted and one step past it refused, naming the key and the bound.
// Every literal is written out rather than derived from a declaration,
// so the test cannot pass by echoing the code under test.
func TestAuthSettings_Bounds(t *testing.T) {
	tests := []struct {
		key, value string
		want       string // empty: accepted
	}{
		{"DUBPLATE_MAGIC_LINK_EXPIRY", "1m", ""},
		{"DUBPLATE_MAGIC_LINK_EXPIRY", "59s", "below the floor 1m"},
		{"DUBPLATE_MAGIC_LINK_EXPIRY", "1h", ""},
		{"DUBPLATE_MAGIC_LINK_EXPIRY", "1h0m1s", "above the ceiling 1h0m0s"},
		{"DUBPLATE_SESSION_LIFETIME", "1s", ""},
		{"DUBPLATE_SESSION_LIFETIME", "0s", "below the floor 1s"},
		{"DUBPLATE_SESSION_LIFETIME", "17520h", ""},
		{"DUBPLATE_SESSION_LIFETIME", "17521h", "above the ceiling 17520h"},
		{"DUBPLATE_SESSION_ABSOLUTE_CAP", "1s", ""},
		{"DUBPLATE_SESSION_ABSOLUTE_CAP", "0s", "below the floor 1s"},
		{"DUBPLATE_SESSION_ABSOLUTE_CAP", "17520h", ""},
		{"DUBPLATE_SESSION_ABSOLUTE_CAP", "17521h", "above the ceiling 17520h"},
		{"DUBPLATE_NONCE_RETENTION", "1s", ""},
		{"DUBPLATE_NONCE_RETENTION", "0s", "below the floor 1s"},
		{"DUBPLATE_NONCE_RETENTION", "240h", ""},
		{"DUBPLATE_NONCE_RETENTION", "240h0m1s", "above the ceiling 240h"},
		{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", "1", ""},
		{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", "0", "below the floor 1"},
		{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", "50", ""},
		{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", "51", "above the ceiling 50"},
		{"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", "1", ""},
		{"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", "0", "below the floor 1"},
		{"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", "200", ""},
		{"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", "201", "above the ceiling 200"},
		{"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", "1", ""},
		{"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", "0", "below the floor 1"},
		{"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", "300", ""},
		{"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", "301", "above the ceiling 300"},
		{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", "1", ""},
		{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", "0", "below the floor 1"},
		{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", "600", ""},
		{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", "601", "above the ceiling 600"},
		{"DUBPLATE_AUTH_REQUEST_BODY_CAP", "512", ""},
		{"DUBPLATE_AUTH_REQUEST_BODY_CAP", "511", "below the floor 512"},
		{"DUBPLATE_AUTH_REQUEST_BODY_CAP", "65536", ""},
		{"DUBPLATE_AUTH_REQUEST_BODY_CAP", "65537", "above the ceiling 65536"},
		{"DUBPLATE_UNKNOWN_ADDRESS_ANSWER", "kind", ""},
		{"DUBPLATE_UNKNOWN_ADDRESS_ANSWER", "generic", ""},
		{"DUBPLATE_UNKNOWN_ADDRESS_ANSWER", "open", `"open" is not one of kind, generic`},
	}

	registered := productionRegistry().Settings()
	for _, tt := range tests {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			i := slices.IndexFunc(registered, func(s config.Setting) bool { return s.Key == tt.key })
			if i < 0 {
				t.Fatalf("no setting %s", tt.key)
			}
			err := registered[i].Validate(tt.value)
			if tt.want == "" {
				if err != nil {
					t.Errorf("Validate(%q) = %v, want nil", tt.value, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) || !strings.Contains(err.Error(), tt.key) {
				t.Errorf("Validate(%q) = %v, want a refusal naming %s and %q", tt.value, err, tt.key, tt.want)
			}
		})
	}
}

// TestAuthCorrelatedRules is spec criterion 5 over the real pair: the
// seed rule runs the two relations over the values in force after the
// whole seed, never over the stored rows. Each refusal names both
// keys, on each key's own outcome. Fixture state: a stored retention
// row of 30m, below the 45m expiry the rows raise it to, so a rule
// compared against the stored rows or one key at a time would decide
// differently from one compared after the seed.
func TestAuthCorrelatedRules(t *testing.T) {
	const (
		expiry    = "DUBPLATE_MAGIC_LINK_EXPIRY"
		retention = "DUBPLATE_NONCE_RETENTION"
		lifetime  = "DUBPLATE_SESSION_LIFETIME"
		absCap    = "DUBPLATE_SESSION_ABSOLUTE_CAP"
	)
	row := func(key, value, source string) settings.Row {
		return settings.Row{Key: key, Value: value, Source: source, SeedValue: value}
	}
	tests := []struct {
		name     string
		rows     []settings.Row
		env      []string
		wantKeys []string // nil: passes
	}{
		{
			name: "retention and expiry raised together in one seed set",
			rows: []settings.Row{row(expiry, "15m", "env_seed"), row(retention, "30m", "env_seed")},
			env:  []string{expiry + "=45m", retention + "=1h"},
		},
		{
			name:     "only the expiry raised past the stored retention",
			rows:     []settings.Row{row(expiry, "15m", "env_seed"), row(retention, "30m", "admin")},
			env:      []string{expiry + "=45m"},
			wantKeys: []string{expiry, retention},
		},
		{
			name:     "lifetime above the cap",
			env:      []string{lifetime + "=4321h", absCap + "=4320h"},
			wantKeys: []string{lifetime, absCap},
		},
		{
			name: "lifetime equal to the cap",
			env:  []string{lifetime + "=4320h", absCap + "=4320h"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := settings.EvaluateFacts(productionRegistry(), tt.env, settings.Facts{Rows: tt.rows})
			if tt.wantKeys == nil {
				if ev.Err != nil {
					t.Fatalf("Err = %v, want the pair to pass", ev.Err)
				}
				return
			}
			if ev.Err == nil {
				t.Fatal("Err = nil, want the correlated rule's refusal")
			}
			for _, o := range ev.Outcomes {
				if !slices.Contains(tt.wantKeys, o.Key) {
					continue
				}
				if o.Err == nil {
					t.Errorf("%s: Err = nil, want the rule's refusal on its own outcome", o.Key)
					continue
				}
				for _, key := range tt.wantKeys {
					if !strings.Contains(o.Err.Error(), key) {
						t.Errorf("%s: Err = %q, want it to name %s", o.Key, o.Err, key)
					}
				}
			}
		})
	}
}

// TestPrepareServe_RefusesAuthCorrelatedRule proves the seed apply,
// the authority, refuses the start on a broken pair before any
// listener could open, and writes no row.
func TestPrepareServe_RefusesAuthCorrelatedRule(t *testing.T) {
	environ := append(serveAuthEnviron(t),
		"DUBPLATE_SESSION_LIFETIME=4321h", "DUBPLATE_SESSION_ABSOLUTE_CAP=4320h")
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err == nil || shell != nil {
		t.Fatalf("prepare = %v, %v, want a refusal", shell, err)
	}
	for _, key := range []string{"DUBPLATE_SESSION_LIFETIME", "DUBPLATE_SESSION_ABSOLUTE_CAP"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error = %q, want it to name %s", err, key)
		}
	}
	if rows, _, _ := readSeedState(t, environValue(environ, "DUBPLATE_STATE_DIR")); len(rows) != 0 {
		t.Errorf("settings = %v after a refused start, want none", rows)
	}
}

// TestAuthSettings_SeedFromEnv is spec criterion 10's seed half, the
// seed table's second row once per auth key: with no stored row and the
// environment naming a value other than the default, serve's start
// seeds one env_seed row for the key alone, audits it, and the reader
// serve wires for the key reads that value. The handler readers are
// read from [authHandlersConfig]'s own return value, the limits through
// [auth.ReadLimits], and the retention from [auth.ConfigFrom]'s
// retention reader, the one serve opens the nonce store with. Each want
// is a literal, so a reader wired to the wrong key fails its row.
func TestAuthSettings_SeedFromEnv(t *testing.T) {
	type readers struct {
		handlers  auth.HandlersConfig
		acc       *settings.Accessor
		retention settings.Reader[time.Duration]
	}
	limit := func(pick func(auth.Limits) int) func(context.Context, readers) (any, error) {
		return func(ctx context.Context, r readers) (any, error) {
			limits, err := auth.ReadLimits(ctx, r.acc)
			return pick(limits), err
		}
	}
	tests := []struct {
		key, value string
		got        func(context.Context, readers) (any, error)
		want       any
	}{
		{"DUBPLATE_MAGIC_LINK_EXPIRY", "20m", func(ctx context.Context, r readers) (any, error) { return r.handlers.LinkExpiry(ctx) }, 20 * time.Minute},
		{"DUBPLATE_SESSION_LIFETIME", "1h", func(ctx context.Context, r readers) (any, error) { return r.handlers.SessionLifetime(ctx) }, time.Hour},
		{"DUBPLATE_SESSION_ABSOLUTE_CAP", "4000h", func(ctx context.Context, r readers) (any, error) { return r.handlers.SessionAbsoluteCap(ctx) }, 4000 * time.Hour},
		{"DUBPLATE_NONCE_RETENTION", "2h", func(ctx context.Context, r readers) (any, error) { return r.retention(ctx) }, 2 * time.Hour},
		{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", "7", limit(func(l auth.Limits) int { return l.LinkRequestsPerAddress }), 7},
		{"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", "22", limit(func(l auth.Limits) int { return l.LinkRequestsPerClient }), 22},
		{"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", "33", limit(func(l auth.Limits) int { return l.ConsumptionAttemptsPerClient }), 33},
		{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", "66", limit(func(l auth.Limits) int { return l.LinkMailCeiling }), 66},
		{"DUBPLATE_AUTH_REQUEST_BODY_CAP", "8192", func(ctx context.Context, r readers) (any, error) { return r.handlers.BodyCap(ctx) }, int64(8192)},
		{"DUBPLATE_UNKNOWN_ADDRESS_ANSWER", "generic", func(ctx context.Context, r readers) (any, error) { return r.handlers.GenericAnswer(ctx) }, true},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			environ := append(serveAuthEnviron(t), tt.key+"="+tt.value)
			if _, err := buildServeShellWithLogger(t, environ, discardTestLogger()); err != nil {
				t.Fatalf("prepare: %v", err)
			}

			rows, audit, _ := readSeedState(t, environValue(environ, "DUBPLATE_STATE_DIR"))
			if want := map[string]settingRow{tt.key: {tt.value, "env_seed"}}; fmt.Sprint(rows) != fmt.Sprint(want) {
				t.Errorf("settings = %v, want %v", rows, want)
			}
			if want := "seed system env_seed " + tt.key; !slices.Equal(audit, []string{want}) {
				t.Errorf("audit = %q, want [%q]", audit, want)
			}
			cfg, err := config.Load(environ, productionRegistry())
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			err = withStoreAccessor(t.Context(), environ, func(acc *settings.Accessor) error {
				ac := authConfigFor(t, cfg, acc)
				r := readers{handlers: authHandlersConfig(serve.Prefix{}, ac, nil, nil, nil, nil, nil, nil, nil), acc: acc, retention: ac.NonceRetention}
				got, err := tt.got(t.Context(), r)
				if err != nil {
					return err
				}
				if got != tt.want {
					t.Errorf("reader for %s = %v, want %v", tt.key, got, tt.want)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("read %s: %v", tt.key, err)
			}
		})
	}
}

// TestAuthSettings_StringRowSecondStart proves a seeded string row
// survives the next start: the stored env_seed row equals the
// environment, so the second start leaves it, writes no audit row, and
// does not refuse. The first start is what stores the row, so the
// second meets exactly the state every restart after it meets.
func TestAuthSettings_StringRowSecondStart(t *testing.T) {
	const key = "DUBPLATE_UNKNOWN_ADDRESS_ANSWER"
	environ := append(serveAuthEnviron(t), key+"=generic")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")

	// The first start closes before the second, which would otherwise
	// meet the intake lease the first holds.
	first, _, _, err := startServeShell(t, t.Context(), environ)
	if err != nil {
		t.Fatalf("first start: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close the first start: %v", err)
	}
	rows, audit, generation := readSeedState(t, stateDir)
	if want := map[string]settingRow{key: {"generic", "env_seed"}}; fmt.Sprint(rows) != fmt.Sprint(want) {
		t.Fatalf("settings after the first start = %v, want %v", rows, want)
	}

	if _, err := buildServeShellWithLogger(t, environ, discardTestLogger()); err != nil {
		t.Fatalf("second start: %v, want the equal string row left as it is", err)
	}
	rows2, audit2, generation2 := readSeedState(t, stateDir)
	if fmt.Sprint(rows2) != fmt.Sprint(rows) || len(audit2) != len(audit) || generation2 != generation {
		t.Errorf("after the second start: settings %v, %d audit rows, generation %d; want %v, %d, %d",
			rows2, len(audit2), generation2, rows, len(audit), generation)
	}
}

// authConfigFor is [auth.ConfigFrom] for a test whose cfg selects the
// magic-link source, failing the test on a refusal.
func authConfigFor(t *testing.T, cfg *config.Config, acc *settings.Accessor) auth.Config {
	t.Helper()
	ac, err := auth.ConfigFrom(cfg, acc)
	if err != nil {
		t.Fatalf("auth.ConfigFrom: %v", err)
	}
	return ac
}
