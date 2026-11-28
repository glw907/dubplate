package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
)

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
			err := validateMagicLinkSource(values, nil)
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
			err := validateMagicLinkSource(values, nil)
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
			err := validateMagicLinkSource(values, nil)
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
			err := validateMagicLinkSource(values, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateMagicLinkSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "DUBPLATE_SMTP_PORT") {
				t.Errorf("error = %q, want it to name DUBPLATE_SMTP_PORT", err.Error())
			}
		})
	}
}

// TestValidateMagicLinkSource_LeavesServePort holds the serve port to
// the binary's own check: with both ports wrong, the source refuses the
// mail port alone, so a bad serve port is reported once, by serve's
// bound, and never a second time from here.
func TestValidateMagicLinkSource_LeavesServePort(t *testing.T) {
	values := map[string]string{"DUBPLATE_SMTP_PORT": "abc"}
	environ := []string{"DUBPLATE_SERVE_PORT=80"}

	err := validateMagicLinkSource(values, environ)
	if err == nil {
		t.Fatal("validateMagicLinkSource() error = nil, want the mail port's refusal")
	}
	if got, want := err.Error(), "DUBPLATE_SMTP_PORT: not an integer"; got != want {
		t.Errorf("error = %q, want %q", got, want)
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
			err := validateMagicLinkSource(values, nil)
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

	err := validateMagicLinkSource(values, nil)
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

// magicLinkEnviron is a wall the magic-link source loads from: each
// secret a distinct literal at the length floor, so a field wired to
// the wrong key reads another key's value and fails its comparison.
func magicLinkEnviron() []string {
	return []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
		"DUBPLATE_IDENTITY_SOURCE=magic-link",
		"DUBPLATE_MAGIC_LINK_SECRET=" + strings.Repeat("l", secretLengthFloor),
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS=" + strings.Repeat("m", secretLengthFloor),
		"DUBPLATE_SESSION_SECRET=" + strings.Repeat("s", secretLengthFloor),
		"DUBPLATE_SESSION_SECRET_PREVIOUS=" + strings.Repeat("t", secretLengthFloor),
		"DUBPLATE_PUBLIC_URL=https://music.example.org",
		"DUBPLATE_SMTP_HOST=smtp.example.org",
		"DUBPLATE_SMTP_PORT=587",
		"DUBPLATE_SMTP_USER=mailer",
		"DUBPLATE_SMTP_PASSWORD=fixture-password",
		"DUBPLATE_SMTP_FROM=Music <music@example.org>",
	}
}

// TestConfigFrom_ResolvesTheSource proves [ConfigFrom] resolves each
// magic-link value from its own key. Every want is a literal from
// [magicLinkEnviron].
func TestConfigFrom_ResolvesTheSource(t *testing.T) {
	reg := config.NewRegistry()
	Register(reg)
	cfg, err := config.Load(magicLinkEnviron(), reg)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	got, err := ConfigFrom(cfg, settings.NewAccessor(nil, reg))
	if err != nil {
		t.Fatalf("ConfigFrom: %v", err)
	}
	for _, tt := range []struct {
		name      string
		got, want string
	}{
		{"LinkSecret", string(got.LinkSecret), strings.Repeat("l", secretLengthFloor)},
		{"PreviousLinkSecret", string(got.PreviousLinkSecret), strings.Repeat("m", secretLengthFloor)},
		{"SessionSecret", string(got.SessionSecret), strings.Repeat("s", secretLengthFloor)},
		{"PreviousSessionSecret", string(got.PreviousSessionSecret), strings.Repeat("t", secretLengthFloor)},
		{"PublicOrigin", got.PublicOrigin, "https://music.example.org"},
	} {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

// TestConfigFrom_RefusesAnotherSource proves [ConfigFrom] refuses a
// config whose selected source is not the magic-link source, so serve
// never signs with the empty secrets such a config resolves to.
func TestConfigFrom_RefusesAnotherSource(t *testing.T) {
	_, err := ConfigFrom(&config.Config{}, settings.NewAccessor(nil, config.NewRegistry()))
	if !errors.Is(err, errSourceNotSelected) {
		t.Fatalf("ConfigFrom() error = %v, want errSourceNotSelected", err)
	}
}

// TestConfigFrom_ReadersNameTheirKeys proves each reader [ConfigFrom]
// wires reads its own setting and reports a failed read as a
// [*settings.ReadError] carrying that setting's key, the shape the
// handlers log by key alone. The accessor is built over an empty
// registry, so every read fails before it reaches a store, and the key
// in the failure names the setting the reader was wired to.
func TestConfigFrom_ReadersNameTheirKeys(t *testing.T) {
	reg := config.NewRegistry()
	Register(reg)
	cfg, err := config.Load(magicLinkEnviron(), reg)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	ac, err := ConfigFrom(cfg, settings.NewAccessor(nil, config.NewRegistry()))
	if err != nil {
		t.Fatalf("ConfigFrom: %v", err)
	}
	tests := []struct {
		name string
		read func(context.Context) error
		key  string
	}{
		{"LinkExpiry", func(ctx context.Context) error { _, err := ac.LinkExpiry(ctx); return err }, "DUBPLATE_MAGIC_LINK_EXPIRY"},
		{"SessionLifetime", func(ctx context.Context) error { _, err := ac.SessionLifetime(ctx); return err }, "DUBPLATE_SESSION_LIFETIME"},
		{"SessionAbsoluteCap", func(ctx context.Context) error { _, err := ac.SessionAbsoluteCap(ctx); return err }, "DUBPLATE_SESSION_ABSOLUTE_CAP"},
		{"NonceRetention", func(ctx context.Context) error { _, err := ac.NonceRetention(ctx); return err }, "DUBPLATE_NONCE_RETENTION"},
		{"BodyCap", func(ctx context.Context) error { _, err := ac.BodyCap(ctx); return err }, "DUBPLATE_AUTH_REQUEST_BODY_CAP"},
		{"GenericAnswer", func(ctx context.Context) error { _, err := ac.GenericAnswer(ctx); return err }, "DUBPLATE_UNKNOWN_ADDRESS_ANSWER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.read(t.Context())
			se, ok := errors.AsType[*settings.ReadError](err)
			if !ok {
				t.Fatalf("read error = %v, want a *settings.ReadError", err)
			}
			if se.Key != tt.key {
				t.Errorf("ReadError key = %q, want %q", se.Key, tt.key)
			}
		})
	}
}

// TestSettingFailureReason proves a failed setting read's log reason
// is a fixed phrase naming the key a [settings.ReadError] carries, however
// deeply wrapped, and never any error's own text.
func TestSettingFailureReason(t *testing.T) {
	keyed := &settings.ReadError{Key: "DUBPLATE_FIXTURE_KEY", Err: errors.New("near a@b.example")}
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"a setting error", keyed, "setting unreadable: DUBPLATE_FIXTURE_KEY"},
		{"a wrapped setting error", fmt.Errorf("verify identity: %w: %w", errSettingUnreadable, keyed), "setting unreadable: DUBPLATE_FIXTURE_KEY"},
		{"an error naming no key", fmt.Errorf("%w: near a@b.example", errSettingUnreadable), "setting unreadable: unnamed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := settingFailureReason(tt.err); got != tt.want {
				t.Errorf("settingFailureReason() = %q, want %q", got, tt.want)
			}
		})
	}
}
