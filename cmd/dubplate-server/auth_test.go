package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
)

// TestRegisterAuthConfig_SelectorIsOwnedNow proves the selector lands as
// an owned wall row with no Pending, retiring keys.go's own pending
// "identity source" status row (R5): a registry that registers nothing
// else still requires DUBPLATE_IDENTITY_SOURCE once registerAuthConfig
// has run.
func TestRegisterAuthConfig_SelectorIsOwnedNow(t *testing.T) {
	reg := config.NewRegistry()
	registerAuthConfig(reg)

	rows := reg.OwnedWall()
	i := slices.IndexFunc(rows, func(row config.WallRow) bool {
		return row.Name == "DUBPLATE_IDENTITY_SOURCE"
	})
	if i < 0 {
		t.Fatal("OwnedWall() has no DUBPLATE_IDENTITY_SOURCE row")
	}
	if rows[i].Pending != "" {
		t.Errorf("DUBPLATE_IDENTITY_SOURCE.Pending = %q, want empty", rows[i].Pending)
	}
}

// TestRegisterAuthConfig_MagicLinkSource proves the magic-link source
// declares exactly the eight required and two optional names R-5b3
// rules, on the selector registerAuthConfig also registers as a wall
// row.
func TestRegisterAuthConfig_MagicLinkSource(t *testing.T) {
	reg := config.NewRegistry()
	registerAuthConfig(reg)

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
	if !slices.Equal(src.Required, wantRequired) {
		t.Errorf("Required = %v, want %v", src.Required, wantRequired)
	}

	wantOptional := []string{
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS",
		"DUBPLATE_SESSION_SECRET_PREVIOUS",
	}
	if !slices.Equal(src.Optional, wantOptional) {
		t.Errorf("Optional = %v, want %v", src.Optional, wantOptional)
	}
}

// TestRegisterAuthConfig_RetiresAuthAndIdentityReservedClass proves
// the removal half of the rung 1b registry contract landed in this
// commit, on TestRegisterServeConfig_ClaimsThreeKeysAndRetiresReservedClass's
// own precedent (serve_test.go): the class registerAuthConfig's
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
	cmd := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr)
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

// fixtureSecretAtFloor and fixtureSecretBelowFloor are built rather
// than hand-counted, so the byte-boundary the test asserts against is
// exactly secretLengthFloor and one below it: strings.Repeat's low
// entropy keeps gitleaks quiet (Global Constraints), and no literal
// here is shaped like a real key.
var (
	fixtureSecretAtFloor    = filler(secretLengthFloor)
	fixtureSecretBelowFloor = filler(secretLengthFloor - 1)

	// fixtureSecretMultiByte is exactly secretLengthFloor bytes but
	// eight runes: proof that the floor counts bytes, never runes.
	fixtureSecretMultiByte = "\U0001F512\U0001F511\U0001F3AF\U0001F3B5\U0001F3B6\U0001F3BC\U0001F3B9\U0001F3A7"
)

// filler returns n bytes of repeated low-entropy filler.
func filler(n int) string {
	return strings.Repeat("ab", n/2+1)[:n]
}

// TestValidateMagicLinkSource_SecretLengthFloor proves the byte-length
// floor on every one of the four secret keys, including both
// rotation-overlap keys and a multi-byte fixture proving the count is
// of bytes rather than runes (R-5b8).
func TestValidateMagicLinkSource_SecretLengthFloor(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{"magic link secret at floor", "DUBPLATE_MAGIC_LINK_SECRET", fixtureSecretAtFloor, false},
		{"magic link secret below floor", "DUBPLATE_MAGIC_LINK_SECRET", fixtureSecretBelowFloor, true},
		{"session secret at floor", "DUBPLATE_SESSION_SECRET", fixtureSecretAtFloor, false},
		{"session secret below floor", "DUBPLATE_SESSION_SECRET", fixtureSecretBelowFloor, true},
		{"magic link secret previous at floor", "DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS", fixtureSecretAtFloor, false},
		{"magic link secret previous below floor", "DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS", fixtureSecretBelowFloor, true},
		{"session secret previous at floor", "DUBPLATE_SESSION_SECRET_PREVIOUS", fixtureSecretAtFloor, false},
		{"session secret previous below floor", "DUBPLATE_SESSION_SECRET_PREVIOUS", fixtureSecretBelowFloor, true},
		{"multi-byte secret at the byte floor", "DUBPLATE_MAGIC_LINK_SECRET", fixtureSecretMultiByte, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{tt.key: tt.value}
			err := validateMagicLinkSource(values, authEnviron(t))
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.key) {
				t.Errorf("error = %q, want it to name %s", err.Error(), tt.key)
			}
		})
	}
}

// TestValidateMagicLinkSource_PublicURL proves the URL check's three
// conditions: it parses, it carries a non-empty host, and its scheme
// is https unless its host is loopback, covering the whole loopback
// set rather than the literal name localhost alone (review m6, m7,
// ruled conductor 2026-09-12).
func TestValidateMagicLinkSource_PublicURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https non-loopback", "https://example.test", false},
		{"http non-loopback", "http://example.test", true},
		{"http localhost", "http://localhost:8080", false},
		{"http loopback IPv4", "http://127.0.0.1:8080", false},
		{"http private IPv4, not loopback", "http://192.168.1.10:8080", true},
		{"bare hostname, empty host", "example.test", true},
		{"https scheme, empty host", "https:///path", true},
		{"unparseable", "://bad", true},
		{"userinfo", "https://evil.example@dubplate.example", true},
		{"query string", "https://dubplate.example?x=1", true},
		{"fragment", "https://dubplate.example#frag", true},
		{"clean https origin", "https://dubplate.example", false},
		{"empty path", "https://dubplate.example", false},
		{"root path", "https://dubplate.example/", false},
		{"a path", "https://dubplate.example/dubplate", true},
		{"a path with a trailing separator", "https://dubplate.example/x/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{"DUBPLATE_PUBLIC_URL": tt.url}
			err := validateMagicLinkSource(values, authEnviron(t))
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "DUBPLATE_PUBLIC_URL") {
				t.Errorf("error = %q, want it to name DUBPLATE_PUBLIC_URL", err.Error())
			}
		})
	}
}

// TestValidateMagicLinkSource_PublicURLNeverEchoesTheValue holds review
// finding 10 and its m2 residual: a public URL carrying a password in
// its userinfo, an unparseable value that also carries a password, and
// a password containing "/", "?" or "#" (each of which makes
// url.Parse cut the authority early and misread "user:pass" as
// "host:port", quoting the password up to that character in its own
// *url.Error) all refuse without the password, or any other part of
// the value, ever reaching the error text.
func TestValidateMagicLinkSource_PublicURLNeverEchoesTheValue(t *testing.T) {
	const password = "fixture-pass-not-real"

	tests := []struct {
		name string
		url  string
	}{
		{"userinfo carries the password", "https://operator:" + password + "@dubplate.example.test"},
		{"unparseable value also carries the password", "https://operator:" + password + "@dubplate.example.test:notaport"},
		{"password contains a slash", "https://operator:" + password + "/notreal@dubplate.example.test"},
		{"password contains a question mark", "https://operator:" + password + "?notreal@dubplate.example.test"},
		{"password contains a hash", "https://operator:" + password + "#notreal@dubplate.example.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{"DUBPLATE_PUBLIC_URL": tt.url}
			err := validateMagicLinkSource(values, authEnviron(t))
			if err == nil {
				t.Fatal("validateMagicLinkSource() error = nil, want a refusal")
			}
			if strings.Contains(err.Error(), password) {
				t.Errorf("error = %q, want it to never contain the password", err.Error())
			}
			if strings.Contains(err.Error(), tt.url) {
				t.Errorf("error = %q, want it to never contain the whole value", err.Error())
			}
		})
	}
}

// TestValidateMagicLinkSource_SMTPPort proves R-5c11's port bound: 1
// and 65535 are the accepted extremes, 0 and 65536 each refuse naming
// the key, and a non-integer value refuses too (named mutation 6: an
// Atoi-only parse would let 0 load, since Atoi accepts it).
func TestValidateMagicLinkSource_SMTPPort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"port at the low bound", "1", false},
		{"port at the high bound", "65535", false},
		{"port zero", "0", true},
		{"port one past the high bound", "65536", true},
		{"port negative", "-1", true},
		{"port not an integer", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{"DUBPLATE_SMTP_PORT": tt.port}
			err := validateMagicLinkSource(values, authEnviron(t))
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "DUBPLATE_SMTP_PORT") {
				t.Errorf("error = %q, want it to name DUBPLATE_SMTP_PORT", err.Error())
			}
		})
	}
}

// TestValidateMagicLinkSource_SMTPFrom proves DUBPLATE_SMTP_FROM's
// load-time check (spec:1289): a value net/mail.ParseAddress refuses,
// a value carrying CR, and a value carrying LF are each refused, a
// well-formed value passes, and no refusal message carries the value.
func TestValidateMagicLinkSource_SMTPFrom(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		wantErr bool
	}{
		{"well-formed address", "dubplate@example.test", false},
		{"unparseable", "not-an-address", true},
		{"two addresses", "a@example.test, b@example.test", true},
		{"carries CR", "dubplate@example.test\rBcc: x@example.test", true},
		{"carries LF", "dubplate@example.test\nBcc: x@example.test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{"DUBPLATE_SMTP_FROM": tt.from}
			err := validateMagicLinkSource(values, authEnviron(t))
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if !strings.Contains(err.Error(), "DUBPLATE_SMTP_FROM") {
					t.Errorf("error = %q, want it to name DUBPLATE_SMTP_FROM", err.Error())
				}
				if strings.Contains(err.Error(), tt.from) {
					t.Errorf("error text %q contains the refused value %q", err.Error(), tt.from)
				}
			}
		})
	}
}

// TestValidateMagicLinkSource_NoSecretInErrors proves that a joined
// refusal covering both secret keys never carries either fixture
// secret's value: the hook names keys and floors, never content or an
// exact length that would narrow a guess.
func TestValidateMagicLinkSource_NoSecretInErrors(t *testing.T) {
	values := map[string]string{
		"DUBPLATE_MAGIC_LINK_SECRET": fixtureSecretBelowFloor,
		"DUBPLATE_SESSION_SECRET":    fixtureSecretBelowFloor,
		"DUBPLATE_PUBLIC_URL":        "http://example.test",
	}

	err := validateMagicLinkSource(values, authEnviron(t))
	if err == nil {
		t.Fatal("validateMagicLinkSource() error = nil, want a joined refusal")
	}
	text := err.Error()
	for _, secret := range []string{fixtureSecretBelowFloor, fixtureSecretAtFloor} {
		if strings.Contains(text, secret) {
			t.Errorf("error text %q contains fixture secret %q", text, secret)
		}
	}
}

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

// TestRosterFileKey_Superseded runs the seven verbs that load the wall
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
		cmd  func(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command
	}{
		{"check", []string{"check"}, newCheckCmd},
		{"doctor", []string{"doctor"}, newDoctorCmd},
		{"serve", []string{"serve"}, newServeCmd},
		{"member add", []string{"member", "add", "a@example.test", "--name", "A", "--username", "a"}, newMemberCmd},
		{"member list", []string{"member", "list"}, newMemberCmd},
		{"backup", []string{"backup"}, newBackupCmd},
		{"restore", []string{"restore", "--snapshot", "latest"}, newRestoreCmd},
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
					root := &cobra.Command{Use: "dubplate-server", SilenceUsage: true, SilenceErrors: true}
					root.AddCommand(v.cmd(environ, productionRegistry(), &stdout, &stderr))
					root.SetArgs(v.args)
					err := root.Execute()

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
		walk(newRootCmd())
		slices.Sort(leaves)

		want := []string{"backup", "check", "doctor", "member add", "member list", "restore", "serve",
			"unidecode install", "version"}
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
		{"DUBPLATE_UNKNOWN_ADDRESS_ANSWER", config.KindString, "kind", "", "", []string{"kind", "generic"}, config.ApplyNextOperation, false},
	}

	fresh := config.NewRegistry()
	registerAuthConfig(fresh)
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
					!slices.Equal(s.Values, tt.values) || s.Apply != tt.apply || s.Sensitive != tt.sensitive {
					t.Errorf("setting = kind %v default %q floor %q ceiling %q values %q apply %v sensitive %v, want %v %q %q %q %q %v %v",
						s.Kind, s.Default, s.Floor, s.Ceiling, s.Values, s.Apply, s.Sensitive,
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
	row := func(key, value, source string) store.SettingsRow {
		return store.SettingsRow{Key: key, Value: value, Source: source, SeedValue: value}
	}
	tests := []struct {
		name     string
		rows     []store.SettingsRow
		env      []string
		wantKeys []string // nil: passes
	}{
		{
			name: "retention and expiry raised together in one seed set",
			rows: []store.SettingsRow{row(expiry, "15m", "env-seed"), row(retention, "30m", "env-seed")},
			env:  []string{expiry + "=45m", retention + "=1h"},
		},
		{
			name:     "only the expiry raised past the stored retention",
			rows:     []store.SettingsRow{row(expiry, "15m", "env-seed"), row(retention, "30m", "admin")},
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
			ev := settings.EvaluateFacts(productionRegistry(), tt.env, store.Facts{Settings: tt.rows})
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
// seeds one env-seed row for the key alone, audits it, and the reader
// serve wires for the key reads that value. The handler readers are
// read from [authHandlersConfig]'s own return value, the limits through
// [readAuthLimits], and the retention through the reader serve hands
// the nonce store. Each want is a literal, so a reader wired to the
// wrong key fails its row.
func TestAuthSettings_SeedFromEnv(t *testing.T) {
	type readers struct {
		handlers auth.HandlersConfig
		acc      *settings.Accessor
	}
	limit := func(pick func(auth.Limits) int) func(context.Context, readers) (any, error) {
		return func(ctx context.Context, r readers) (any, error) {
			limits, err := readAuthLimits(ctx, r.acc)
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
		{"DUBPLATE_NONCE_RETENTION", "2h", func(ctx context.Context, r readers) (any, error) {
			return durationSetting(r.acc, nonceRetentionSetting)(ctx)
		}, 2 * time.Hour},
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
			if want := map[string]settingRow{tt.key: {tt.value, "env-seed"}}; fmt.Sprint(rows) != fmt.Sprint(want) {
				t.Errorf("settings = %v, want %v", rows, want)
			}
			if want := "seed system env-seed " + tt.key; !slices.Equal(audit, []string{want}) {
				t.Errorf("audit = %q, want [%q]", audit, want)
			}
			cfg, err := config.Load(environ, productionRegistry())
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			err = withStoreAccessor(t.Context(), environ, func(acc *settings.Accessor) error {
				r := readers{handlers: authHandlersConfig(serve.Prefix{}, cfg, acc, nil, nil, nil, nil, nil, nil, nil), acc: acc}
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
// survives the next start: the stored env-seed row equals the
// environment, so the second start leaves it, writes no audit row, and
// does not refuse. The first start is what stores the row, so the
// second meets exactly the state every restart after it meets.
func TestAuthSettings_StringRowSecondStart(t *testing.T) {
	const key = "DUBPLATE_UNKNOWN_ADDRESS_ANSWER"
	environ := append(serveAuthEnviron(t), key+"=generic")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")

	ctx, cancel := context.WithCancel(t.Context())
	if _, err := buildServeShellContext(t, ctx, environ, discardTestLogger()); err != nil {
		t.Fatalf("first start: %v", err)
	}
	cancel()
	rows, audit, generation := readSeedState(t, stateDir)
	if want := map[string]settingRow{key: {"generic", "env-seed"}}; fmt.Sprint(rows) != fmt.Sprint(want) {
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

// TestAuthSettingReaders_FailureNamesTheKey proves each reader serve
// hands the auth package reports a failed read as an
// [*auth.SettingError] carrying the setting's key, the shape the
// handlers log by key alone. The accessor is built over an empty
// registry, so every read fails before it reaches a store.
func TestAuthSettingReaders_FailureNamesTheKey(t *testing.T) {
	acc := settings.NewAccessor(nil, config.NewRegistry())
	tests := []struct {
		name string
		read func(context.Context) error
		key  string
	}{
		{"duration", func(ctx context.Context) error {
			_, err := durationSetting(acc, sessionAbsoluteCapSetting)(ctx)
			return err
		}, "DUBPLATE_SESSION_ABSOLUTE_CAP"},
		{"body cap", func(ctx context.Context) error {
			_, err := authBodyCap(acc)(ctx)
			return err
		}, "DUBPLATE_AUTH_REQUEST_BODY_CAP"},
		{"generic answer", func(ctx context.Context) error {
			_, err := genericAnswer(acc)(ctx)
			return err
		}, "DUBPLATE_UNKNOWN_ADDRESS_ANSWER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.read(t.Context())
			se, ok := errors.AsType[*auth.SettingError](err)
			if !ok {
				t.Fatalf("read error = %v, want an *auth.SettingError", err)
			}
			if se.Key != tt.key {
				t.Errorf("SettingError key = %q, want %q", se.Key, tt.key)
			}
		})
	}
}
