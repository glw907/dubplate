package main

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
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
// declares exactly the nine required and two optional names R-5b3
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
		"DUBPLATE_ROSTER_FILE",
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
	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(baseEnviron(), productionRegistry(), &stdout, &stderr)
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

// TestRegisterAuthConfig_ThresholdDefaults proves all ten thresholds
// registerAuthConfig declares carry their literal keys, kinds, and
// defaults, asserted against the fixed literals rather than by reading
// the declaration back. The four durations trace to R-5b10; the four
// rates and the unknown-address answer trace to fork 3's ruled
// defaults (spec:216-229) and R25/fork 4 respectively; the auth
// request body cap traces to the stated reason on its own declaration
// (auth.go, authRequestBodyCapFloor/Ceiling's doc comment): 4KiB
// comfortably fits the login form's mailed-address field. Each row
// runs against both a fresh registry and productionRegistry(), on
// TestRegisterServeConfig_ClaimsThreeKeysAndRetiresReservedClass's own
// precedent (serve_test.go), so a threshold that reaches the binary's
// own registry only through a path registerAuthConfig alone does not
// exercise is still caught.
func TestRegisterAuthConfig_ThresholdDefaults(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		kind    config.ThresholdKind
		wantDef string
	}{
		{"magic link expiry", "DUBPLATE_MAGIC_LINK_EXPIRY", config.KindDuration, "15m"},
		{"session lifetime", "DUBPLATE_SESSION_LIFETIME", config.KindDuration, "2160h"},
		{"session absolute cap", "DUBPLATE_SESSION_ABSOLUTE_CAP", config.KindDuration, "4320h"},
		{"nonce retention", "DUBPLATE_NONCE_RETENTION", config.KindDuration, "24h"},
		{"link requests per address", "DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", config.KindInt, "5"},
		{"link requests per client", "DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", config.KindInt, "20"},
		{"consumption attempts per client", "DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", config.KindInt, "30"},
		{"link mail ceiling", "DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", config.KindInt, "60"},
		{"auth request body cap", "DUBPLATE_AUTH_REQUEST_BODY_CAP", config.KindInt, "4096"},
		{"unknown address answer", "DUBPLATE_UNKNOWN_ADDRESS_ANSWER", config.KindString, "kind"},
	}

	registries := map[string]*config.Registry{
		"fresh":      config.NewRegistry(),
		"production": productionRegistry(),
	}
	registerAuthConfig(registries["fresh"])

	for regName, reg := range registries {
		thresholds := reg.Thresholds()
		for _, tt := range tests {
			t.Run(regName+"/"+tt.name, func(t *testing.T) {
				i := slices.IndexFunc(thresholds, func(th config.Threshold) bool { return th.Key == tt.key })
				if i < 0 {
					t.Fatalf("Thresholds() has no %s", tt.key)
				}
				th := thresholds[i]
				if th.Kind != tt.kind {
					t.Errorf("Kind = %v, want %v", th.Kind, tt.kind)
				}
				if th.Default != tt.wantDef {
					t.Errorf("Default = %q, want %q", th.Default, tt.wantDef)
				}
			})
		}
	}
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
// finding 10: a public URL carrying a password in its userinfo, and a
// separately unparseable value that also carries a password, both
// refuse without the password (or any other part of the value) ever
// reaching the error text. url.Parse's own *url.Error formats the whole
// URL it was given, so the unparseable row is the one that fails if
// validatePublicURL ever wraps that error verbatim instead of reading
// its inner Err alone.
func TestValidateMagicLinkSource_PublicURLNeverEchoesTheValue(t *testing.T) {
	const password = "fixture-pass-not-real"

	tests := []struct {
		name string
		url  string
	}{
		{"userinfo carries the password", "https://operator:" + password + "@dubplate.example.test"},
		{"unparseable value also carries the password", "https://operator:" + password + "@dubplate.example.test:notaport"},
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

// TestValidateMagicLinkSource_NonceRetention proves R31's floor is
// inclusive: a retention equal to the link expiry passes and one below
// it fails, both read from the environment snapshot as thresholds
// rather than source keys.
func TestValidateMagicLinkSource_NonceRetention(t *testing.T) {
	tests := []struct {
		name      string
		expiry    string
		retention string
		wantErr   bool
	}{
		{"retention equal to expiry", "15m", "15m", false},
		{"retention one minute below expiry", "15m", "14m", true},
		{"retention well above expiry", "15m", "24h", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(baseEnviron(),
				"DUBPLATE_MAGIC_LINK_EXPIRY="+tt.expiry,
				"DUBPLATE_NONCE_RETENTION="+tt.retention,
			)
			err := validateMagicLinkSource(nil, environ)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				for _, key := range []string{"DUBPLATE_NONCE_RETENTION", "DUBPLATE_MAGIC_LINK_EXPIRY"} {
					if !strings.Contains(err.Error(), key) {
						t.Errorf("error = %q, want it to name %s", err.Error(), key)
					}
				}
			}
		})
	}
}

// TestValidateMagicLinkSource_NonceRetentionNamesBothKeysAsDistinctErrors
// proves R-5c10's keyed-type mechanism directly: a retention-vs-expiry
// refusal joins two distinct *sourceKeyError values, one per key, so
// doctor's errors.As walk (cmd/dubplate-server/doctor.go,
// identitySourceStatus) can recover both rather than only the one whose
// message happens to come first.
func TestValidateMagicLinkSource_NonceRetentionNamesBothKeysAsDistinctErrors(t *testing.T) {
	environ := append(baseEnviron(),
		"DUBPLATE_MAGIC_LINK_EXPIRY=15m",
		"DUBPLATE_NONCE_RETENTION=14m",
	)
	err := validateMagicLinkSource(nil, environ)
	if err == nil {
		t.Fatal("validateMagicLinkSource() error = nil, want a refusal: retention is below expiry")
	}

	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("validateMagicLinkSource() error does not implement Unwrap() []error: %v", err)
	}

	var keys []string
	for _, leaf := range joined.Unwrap() {
		if keyed, ok := errors.AsType[*sourceKeyError](leaf); ok {
			keys = append(keys, keyed.key)
		}
	}
	for _, want := range []string{"DUBPLATE_NONCE_RETENTION", "DUBPLATE_MAGIC_LINK_EXPIRY"} {
		if !slices.Contains(keys, want) {
			t.Errorf("keyed errors named %v, want %s among them", keys, want)
		}
	}
}

// TestValidateMagicLinkSource_DurationBounds proves the review m5 fix:
// each of the four thresholds refuses at and one past its floor, and
// the two that carry a ceiling (the link expiry and the session
// absolute cap) refuse at and one past it too. Every literal below is
// written out rather than derived from durationFloor,
// magicLinkExpiryCeiling, or sessionAbsoluteCapCeiling, so the test
// cannot pass by echoing the code under test.
func TestValidateMagicLinkSource_DurationBounds(t *testing.T) {
	// bound overrides one threshold key in a base environ that
	// otherwise satisfies every other threshold and relation, so a
	// row's own refusal is the only one that can fire.
	bound := func(key, value string) []string {
		environ := append(baseEnviron(),
			"DUBPLATE_MAGIC_LINK_EXPIRY=15m",
			"DUBPLATE_SESSION_LIFETIME=2160h",
			"DUBPLATE_SESSION_ABSOLUTE_CAP=4320h",
			"DUBPLATE_NONCE_RETENTION=24h",
		)
		return append(dropKey(environ, key), key+"="+value)
	}

	tests := []struct {
		name    string
		environ []string
		wantErr bool
	}{
		// A positive floor of one second on all four (review m5): zero
		// or a negative duration cannot mean "never" or "immediately"
		// here.
		{"link expiry at the one-second floor", bound("DUBPLATE_MAGIC_LINK_EXPIRY", "1s"), false},
		{"link expiry one below the floor (zero)", bound("DUBPLATE_MAGIC_LINK_EXPIRY", "0s"), true},
		{"session lifetime at the one-second floor", bound("DUBPLATE_SESSION_LIFETIME", "1s"), false},
		{"session lifetime one below the floor (zero)", bound("DUBPLATE_SESSION_LIFETIME", "0s"), true},
		// The cap row's own floor is asserted with the session lifetime
		// also lowered to one second, so R-5c11's separate
		// lifetime-vs-cap relation cannot fire alongside it.
		{"session absolute cap at the one-second floor", append(bound("DUBPLATE_SESSION_ABSOLUTE_CAP", "1s"), "DUBPLATE_SESSION_LIFETIME=1s"), false},
		{"session absolute cap one below the floor (zero)", bound("DUBPLATE_SESSION_ABSOLUTE_CAP", "0s"), true},
		// The retention row's own floor is asserted with the link
		// expiry also lowered to one second, so R31's separate
		// retention-vs-expiry relation cannot fire alongside it.
		{"nonce retention at the one-second floor", append(bound("DUBPLATE_NONCE_RETENTION", "1s"), "DUBPLATE_MAGIC_LINK_EXPIRY=1s"), false},
		{"nonce retention one below the floor (zero)", bound("DUBPLATE_NONCE_RETENTION", "0s"), true},

		// A ceiling of one hour on the link expiry (review m5): the
		// threat model bounds the mail-scanner residual at "a
		// 15-minute default expiry", and an hour is the generous
		// headroom the fix admits above it.
		{"link expiry at the one-hour ceiling", bound("DUBPLATE_MAGIC_LINK_EXPIRY", "1h"), false},
		{"link expiry one second past the ceiling", bound("DUBPLATE_MAGIC_LINK_EXPIRY", "1h0m1s"), true},

		// A ceiling of two years on the session absolute cap (review
		// m5): generous headroom above the 4320h default without
		// admitting a cap that is, in practice, permanent.
		{"session absolute cap at the two-year ceiling", bound("DUBPLATE_SESSION_ABSOLUTE_CAP", "17520h"), false},
		{"session absolute cap one hour past the ceiling", bound("DUBPLATE_SESSION_ABSOLUTE_CAP", "17521h"), true},

		// A ceiling of 240h on nonce retention (R-5c11), ten times its
		// 24h default on the same ten-times rule the rate thresholds'
		// ceilings already follow.
		{"nonce retention at the 240h ceiling", bound("DUBPLATE_NONCE_RETENTION", "240h"), false},
		{"nonce retention one second past the ceiling", bound("DUBPLATE_NONCE_RETENTION", "240h0m1s"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMagicLinkSource(nil, tt.environ)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateMagicLinkSource_RateAndCapBounds proves the fix rung
// 5b-ii Task 3 adds: fork 3's four remaining rates and the auth
// request body cap each refuse at and one past their floor, and each
// refuse at and one past their ceiling. Every literal is written out
// rather than derived from rateFloor or a *Ceiling constant, so the
// test cannot pass by echoing the code under test.
func TestValidateMagicLinkSource_RateAndCapBounds(t *testing.T) {
	// intBound overrides one threshold key in a base environ that
	// otherwise satisfies every other threshold and relation.
	intBound := func(key string, value int) []string {
		environ := append(baseEnviron(),
			"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR=5",
			"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR=20",
			"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR=30",
			"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR=60",
			"DUBPLATE_AUTH_REQUEST_BODY_CAP=4096",
		)
		return append(dropKey(environ, key), fmt.Sprintf("%s=%d", key, value))
	}

	tests := []struct {
		name    string
		environ []string
		wantErr bool
	}{
		{"link requests per address at the floor", intBound("DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", 1), false},
		{"link requests per address one below the floor", intBound("DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", 0), true},
		{"link requests per address at the ceiling", intBound("DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", 50), false},
		{"link requests per address one past the ceiling", intBound("DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", 51), true},

		{"link requests per client at the floor", intBound("DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", 1), false},
		{"link requests per client one below the floor", intBound("DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", 0), true},
		{"link requests per client at the ceiling", intBound("DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", 200), false},
		{"link requests per client one past the ceiling", intBound("DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", 201), true},

		{"consumption attempts per client at the floor", intBound("DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", 1), false},
		{"consumption attempts per client one below the floor", intBound("DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", 0), true},
		{"consumption attempts per client at the ceiling", intBound("DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", 300), false},
		{"consumption attempts per client one past the ceiling", intBound("DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", 301), true},

		{"link mail ceiling at the floor", intBound("DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", 1), false},
		{"link mail ceiling one below the floor", intBound("DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", 0), true},
		{"link mail ceiling at its own ceiling", intBound("DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", 600), false},
		{"link mail ceiling one past its own ceiling", intBound("DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", 601), true},

		{"auth request body cap at the floor", intBound("DUBPLATE_AUTH_REQUEST_BODY_CAP", 512), false},
		{"auth request body cap one below the floor", intBound("DUBPLATE_AUTH_REQUEST_BODY_CAP", 511), true},
		{"auth request body cap at the ceiling", intBound("DUBPLATE_AUTH_REQUEST_BODY_CAP", 65536), false},
		{"auth request body cap one past the ceiling", intBound("DUBPLATE_AUTH_REQUEST_BODY_CAP", 65537), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMagicLinkSource(nil, tt.environ)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateMagicLinkSource_SessionLifetimeAboveCap proves R-5c11's
// load-time half: a lifetime at or below the cap loads, and one above it
// refuses naming both keys. The doctor-level repro (lifetime 87600h,
// cap 1h) is proved separately, through the doctor harness, in
// doctor_test.go.
func TestValidateMagicLinkSource_SessionLifetimeAboveCap(t *testing.T) {
	tests := []struct {
		name     string
		lifetime string
		cap      string
		wantErr  bool
	}{
		{"lifetime equal to the cap", "4320h", "4320h", false},
		{"lifetime below the cap", "2160h", "4320h", false},
		{"lifetime one hour above the cap", "4321h", "4320h", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(baseEnviron(),
				"DUBPLATE_MAGIC_LINK_EXPIRY=15m",
				"DUBPLATE_SESSION_LIFETIME="+tt.lifetime,
				"DUBPLATE_SESSION_ABSOLUTE_CAP="+tt.cap,
				"DUBPLATE_NONCE_RETENTION=24h",
			)
			err := validateMagicLinkSource(nil, environ)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				for _, key := range []string{"DUBPLATE_SESSION_LIFETIME", "DUBPLATE_SESSION_ABSOLUTE_CAP"} {
					if !strings.Contains(err.Error(), key) {
						t.Errorf("error = %q, want it to name %s", err.Error(), key)
					}
				}
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

// TestValidateMagicLinkSource_UnknownAddressAnswer proves R-5b22's
// accepted set: kind and generic pass, and every other value refuses
// with an error naming both accepted values, written out rather than
// read back from unknownAddressAnswerThreshold.
func TestValidateMagicLinkSource_UnknownAddressAnswer(t *testing.T) {
	tests := []struct {
		name    string
		answer  string
		wantErr bool
	}{
		{"kind", "kind", false},
		{"generic", "generic", false},
		{"unrecognized", "silent", true},
		{"empty falls to the default", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := dropKey(baseEnviron(), "DUBPLATE_UNKNOWN_ADDRESS_ANSWER")
			if tt.answer != "" {
				environ = append(environ, "DUBPLATE_UNKNOWN_ADDRESS_ANSWER="+tt.answer)
			}
			err := validateMagicLinkSource(nil, environ)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				for _, want := range []string{"kind", "generic"} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error = %q, want it to name %s", err.Error(), want)
					}
				}
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

// TestResolveAuthConfig proves the resolver reads all ten thresholds'
// defaults when unset and an adopter's override when set, on
// resolveServeConfig's own precedent. want is built field by field
// rather than by re-deriving [authConfig] from the registry, so a
// resolver that reads the wrong threshold into the wrong field is
// caught by a mismatch against a value this test owns (named mutation
// 4: a swapped field reads a rate as a cap).
func TestResolveAuthConfig(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    authConfig
	}{
		{
			name:    "defaults",
			environ: baseEnviron(),
			want: authConfig{
				magicLinkExpiry:              15 * time.Minute,
				sessionLifetime:              2160 * time.Hour,
				sessionAbsoluteCap:           4320 * time.Hour,
				nonceRetention:               24 * time.Hour,
				linkRequestsPerAddress:       5,
				linkRequestsPerClient:        20,
				consumptionAttemptsPerClient: 30,
				linkMailCeiling:              60,
				authRequestBodyCap:           4096,
				unknownAddressAnswer:         "kind",
			},
		},
		{
			name: "overrides",
			environ: append(baseEnviron(),
				"DUBPLATE_MAGIC_LINK_EXPIRY=10m",
				"DUBPLATE_SESSION_LIFETIME=1h",
				"DUBPLATE_SESSION_ABSOLUTE_CAP=2h",
				"DUBPLATE_NONCE_RETENTION=1h",
				"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR=7",
				"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR=22",
				"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR=33",
				"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR=66",
				"DUBPLATE_AUTH_REQUEST_BODY_CAP=8192",
				"DUBPLATE_UNKNOWN_ADDRESS_ANSWER=generic",
			),
			want: authConfig{
				magicLinkExpiry:              10 * time.Minute,
				sessionLifetime:              time.Hour,
				sessionAbsoluteCap:           2 * time.Hour,
				nonceRetention:               time.Hour,
				linkRequestsPerAddress:       7,
				linkRequestsPerClient:        22,
				consumptionAttemptsPerClient: 33,
				linkMailCeiling:              66,
				authRequestBodyCap:           8192,
				unknownAddressAnswer:         "generic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAuthConfig(tt.environ)
			if err != nil {
				t.Fatalf("resolveAuthConfig() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("resolveAuthConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
