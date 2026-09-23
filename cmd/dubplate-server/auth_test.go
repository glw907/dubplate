package main

import (
	"bytes"
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

// TestRegisterAuthConfig_ThresholdDefaults proves the four thresholds
// registerAuthConfig declares carry R-5b10's literal keys, kinds, and
// defaults, asserted against the fixed literals rather than by reading
// the declaration back (R-5b10).
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
	}

	reg := config.NewRegistry()
	registerAuthConfig(reg)
	thresholds := reg.Thresholds()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			err := validateMagicLinkSource(values, authEnviron())
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{"DUBPLATE_PUBLIC_URL": tt.url}
			err := validateMagicLinkSource(values, authEnviron())
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "DUBPLATE_PUBLIC_URL") {
				t.Errorf("error = %q, want it to name DUBPLATE_PUBLIC_URL", err.Error())
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
		{"session absolute cap at the one-second floor", bound("DUBPLATE_SESSION_ABSOLUTE_CAP", "1s"), false},
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

	err := validateMagicLinkSource(values, authEnviron())
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
	environ := dropKey(authEnviron(), "DUBPLATE_LIBRARY_ROOT")
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

// TestResolveAuthConfig proves the resolver reads the four thresholds'
// defaults when unset and an adopter's override when set, on
// resolveServeConfig's own precedent.
func TestResolveAuthConfig(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    [4]time.Duration
	}{
		{
			name:    "defaults",
			environ: baseEnviron(),
			want:    [4]time.Duration{15 * time.Minute, 2160 * time.Hour, 4320 * time.Hour, 24 * time.Hour},
		},
		{
			name: "overrides",
			environ: append(baseEnviron(),
				"DUBPLATE_MAGIC_LINK_EXPIRY=10m",
				"DUBPLATE_SESSION_LIFETIME=1h",
				"DUBPLATE_SESSION_ABSOLUTE_CAP=2h",
				"DUBPLATE_NONCE_RETENTION=1h",
			),
			want: [4]time.Duration{10 * time.Minute, time.Hour, 2 * time.Hour, time.Hour},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiry, lifetime, cap_, retention, err := resolveAuthConfig(tt.environ)
			if err != nil {
				t.Fatalf("resolveAuthConfig() error = %v", err)
			}
			got := [4]time.Duration{expiry, lifetime, cap_, retention}
			if got != tt.want {
				t.Errorf("resolveAuthConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
