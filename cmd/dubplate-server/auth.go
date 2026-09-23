package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// identitySourceSelectorRow declares DUBPLATE_IDENTITY_SOURCE as an
// owned wall row with no Pending, retiring the pending row keys.go
// carried from rung 1 (R5): the selector is required now, alongside
// [magicLinkSource] itself, rather than pending a rung that has
// arrived.
var identitySourceSelectorRow = config.WallRow{Name: "DUBPLATE_IDENTITY_SOURCE"}

// magicLinkSource is the shipped default identity source: a login link
// mailed to a roster address, backed by an HMAC-signed session cookie.
// Its nine required names and two optional names retire
// internal/config/keys.go's "auth and identity" reserved class in the
// same commit that declares them (R-5b3): DUBPLATE_ROSTER_FILE and the
// five DUBPLATE_SMTP_* names land here, declared and
// required-when-selected, though their consumers (the roster store,
// the mail seam) arrive at 5b-ii. The two optional names carry a
// rotation-overlap secret each, so a deployment can roll a secret
// without a window where every existing token or session fails at
// once (R-5b6): they are the source's Optional keys rather than
// registered thresholds, because a threshold's KindString default
// must be non-empty and doctor prints a threshold's effective value,
// which would print a secret on an operator's terminal.
//
// Validate is [validateMagicLinkSource]: Task 4 assigns the source's
// value rules (the secret length floor, the public URL's shape, the
// retention-vs-expiry relation) to this field.
//
// ruled, conductor 2026-09-12, R-5b2, R-5b3
var magicLinkSource = config.IdentitySource{
	Name:     "magic-link",
	Selector: identitySourceSelectorRow.Name,
	Required: []string{
		"DUBPLATE_ROSTER_FILE",
		"DUBPLATE_MAGIC_LINK_SECRET",
		"DUBPLATE_SESSION_SECRET",
		"DUBPLATE_PUBLIC_URL",
		"DUBPLATE_SMTP_HOST",
		"DUBPLATE_SMTP_PORT",
		"DUBPLATE_SMTP_USER",
		"DUBPLATE_SMTP_PASSWORD",
		"DUBPLATE_SMTP_FROM",
	},
	Optional: []string{
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS",
		"DUBPLATE_SESSION_SECRET_PREVIOUS",
	},
	Validate: validateMagicLinkSource,
}

// secretLengthFloor is the minimum accepted byte length of a magic-link
// or session secret, HMAC-SHA256's own recommended key length: RFC 2104
// section 3 (https://www.rfc-editor.org/rfc/rfc2104#section-3, read
// 2026-09-12) recommends a key at least as long as the hash's output,
// 32 bytes for SHA-256 (R-5b8). The count is of raw bytes with no
// transformation, so an adopter pasting a base64 rendering of 32 random
// bytes passes with room to spare.
const secretLengthFloor = 32

// durationFloor is the positive floor every one of rung 5's four
// duration thresholds shares: zero or a negative duration cannot mean
// "never" or "immediately" here, and config.ParseDuration accepts
// both, so this floor is the only refusal of them (review m5).
const durationFloor = time.Second

// magicLinkExpiryCeiling bounds how long a mailed link can stay
// redeemable. The threat model's mail-scanner residual is bounded by
// "single use, a 15-minute default expiry"; an hour is generous
// headroom above the default without admitting a link that survives a
// whole day in a mail provider's link-scanner cache or a browser's
// history (review m5).
const magicLinkExpiryCeiling = time.Hour

// sessionAbsoluteCapCeiling bounds how long a session can survive
// however many times it is refreshed. Two years is generous headroom
// above the 4320h (180-day) default without admitting a cap that is, in
// every practical sense, permanent (review m5).
const sessionAbsoluteCapCeiling = 2 * 365 * 24 * time.Hour

// magicLinkExpiryThreshold, sessionLifetimeThreshold,
// sessionAbsoluteCapThreshold, and nonceRetentionThreshold are rung
// 5b-i's four thresholds (R-5b10). The first three carry product fork
// 3's ruled durations; the fourth is not a fork value; see the ruling
// for the reasoning behind its default.
var (
	magicLinkExpiryThreshold = config.Threshold{
		Key:     "DUBPLATE_MAGIC_LINK_EXPIRY",
		Kind:    config.KindDuration,
		Default: "15m",
		Doc:     "a Go duration string, for example 15m",
	}
	sessionLifetimeThreshold = config.Threshold{
		Key:     "DUBPLATE_SESSION_LIFETIME",
		Kind:    config.KindDuration,
		Default: "2160h",
		Doc:     "a Go duration string, for example 2160h (90 days)",
	}
	sessionAbsoluteCapThreshold = config.Threshold{
		Key:     "DUBPLATE_SESSION_ABSOLUTE_CAP",
		Kind:    config.KindDuration,
		Default: "4320h",
		Doc:     "a Go duration string, for example 4320h (180 days)",
	}
	nonceRetentionThreshold = config.Threshold{
		Key:     "DUBPLATE_NONCE_RETENTION",
		Kind:    config.KindDuration,
		Default: "24h",
		Doc:     "a Go duration string, for example 24h",
	}
)

// registerAuthConfig adds rung 5's twelve names and four thresholds to
// reg: the selector as an owned wall row, magic-link as a declared
// identity source, and the four thresholds R-5b10 rules. It is a plain
// function on registerServeConfig's own precedent (serve.go), called
// once from the binary's shared registry constructor
// (newProductionRegistry, main.go).
func registerAuthConfig(reg *config.Registry) {
	reg.AddWallRow(identitySourceSelectorRow)
	reg.AddIdentitySource(magicLinkSource)
	reg.AddThreshold(magicLinkExpiryThreshold)
	reg.AddThreshold(sessionLifetimeThreshold)
	reg.AddThreshold(sessionAbsoluteCapThreshold)
	reg.AddThreshold(nonceRetentionThreshold)
}

// validateMagicLinkSource is [magicLinkSource]'s Validate hook: it
// checks the secret length floor on both rotation pairs, the public
// URL's shape, the four duration thresholds' bounds (review m5), and
// the nonce-retention-vs-expiry relation (R31), joining every failure
// it finds so [config.Load] reports a short secret and a bad URL
// together. Every returned error names the offending key and never the
// value or the length of a secret, so a refusal narrows no guess about
// what the value was.
func validateMagicLinkSource(values map[string]string, environ []string) error {
	var errs []error

	for _, key := range []string{
		"DUBPLATE_MAGIC_LINK_SECRET",
		"DUBPLATE_SESSION_SECRET",
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS",
		"DUBPLATE_SESSION_SECRET_PREVIOUS",
	} {
		if v, ok := values[key]; ok && len(v) < secretLengthFloor {
			errs = append(errs, fmt.Errorf("%s: too short, want at least %d bytes", key, secretLengthFloor))
		}
	}

	if v, ok := values["DUBPLATE_PUBLIC_URL"]; ok {
		if err := validatePublicURL(v); err != nil {
			errs = append(errs, err)
		}
	}

	var expiry, retention time.Duration
	var expiryErr, retentionErr error
	errs, expiry, expiryErr = checkDurationBounds(errs, environ, magicLinkExpiryThreshold, magicLinkExpiryCeiling)
	errs, _, _ = checkDurationBounds(errs, environ, sessionLifetimeThreshold, 0)
	errs, _, _ = checkDurationBounds(errs, environ, sessionAbsoluteCapThreshold, sessionAbsoluteCapCeiling)
	errs, retention, retentionErr = checkDurationBounds(errs, environ, nonceRetentionThreshold, 0)

	if expiryErr == nil && retentionErr == nil && retention < expiry {
		errs = append(errs, fmt.Errorf("%s: %s is below %s's %s", nonceRetentionThreshold.Key, retention, magicLinkExpiryThreshold.Key, expiry))
	}

	return errors.Join(errs...)
}

// checkDurationBounds parses th against environ and appends a bound
// refusal to errs when the value is below durationFloor or, when
// ceiling is positive, above it. It returns the parsed value alongside
// errs and the parse error, so a caller that needs the value for a
// further relational check (the retention-vs-expiry comparison) does
// not parse it a second time.
func checkDurationBounds(errs []error, environ []string, th config.Threshold, ceiling time.Duration) ([]error, time.Duration, error) {
	d, err := config.ParseDuration(environ, th)
	if err != nil {
		return append(errs, err), 0, err
	}
	if d < durationFloor {
		errs = append(errs, fmt.Errorf("%s: %s is below the floor of %s", th.Key, d, durationFloor))
	}
	if ceiling > 0 && d > ceiling {
		errs = append(errs, fmt.Errorf("%s: %s is above the ceiling of %s", th.Key, d, ceiling))
	}
	return errs, d, nil
}

// validatePublicURL checks value against the conditions a configured
// public origin must meet (review m7): it parses; it carries no
// userinfo, no query string, and no fragment, because any of the three
// would ride into every link [LinkURL] builds and make an
// attacker-chosen string part of its visible prefix or suffix (review
// m4); it carries a non-empty host, because net/url accepts a bare
// hostname with a nil error and an empty Host, and a URL with no host
// builds a path-only link that defeats security review finding 7; and
// its scheme is https unless its host is loopback, the carve-out the
// W3C Secure Contexts definition of a potentially trustworthy origin
// extends to the whole loopback set
// (https://w3c.github.io/webappsec-secure-contexts/#is-origin-trustworthy,
// read 2026-09-12), not the name localhost alone (review m6, ruled,
// conductor 2026-09-12). The comparison reads the parsed hostname
// rather than the host, since url.Parse("http://localhost:8080").Host
// is "localhost:8080" and only Hostname() is "localhost".
func validatePublicURL(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("DUBPLATE_PUBLIC_URL: invalid URL %q: %w", value, err)
	}
	if u.User != nil {
		return fmt.Errorf("DUBPLATE_PUBLIC_URL: %q carries userinfo, which would become an attacker-chosen prefix of every mailed link", value)
	}
	if u.RawQuery != "" {
		return fmt.Errorf("DUBPLATE_PUBLIC_URL: %q carries a query string", value)
	}
	if u.Fragment != "" {
		return fmt.Errorf("DUBPLATE_PUBLIC_URL: %q carries a fragment", value)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("DUBPLATE_PUBLIC_URL: %q has no host", value)
	}
	if u.Scheme != "https" && !isLoopbackHost(host) {
		return fmt.Errorf("DUBPLATE_PUBLIC_URL: %q must use https unless its host is loopback", value)
	}
	return nil
}

// isLoopbackHost reports whether host is the name localhost or an IP
// address in the loopback range, the whole loopback set R38's HTTPS
// carve-out covers (fold 8, ruled, conductor 2026-09-12).
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// resolveAuthConfig reads the four thresholds against environ and
// returns the values the auth core takes as plain construction
// arguments, on [resolveServeConfig]'s own precedent
// (serve.go:81-106): internal/auth never reads the environment itself.
//
// The results are named where resolveServeConfig's are not, because
// four results of one type say nothing about which duration is which.
// No return here is naked; the names are documentation godoc carries.
func resolveAuthConfig(environ []string) (magicLinkExpiry, sessionLifetime, sessionAbsoluteCap, nonceRetention time.Duration, err error) {
	magicLinkExpiry, err = config.ParseDuration(environ, magicLinkExpiryThreshold)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	sessionLifetime, err = config.ParseDuration(environ, sessionLifetimeThreshold)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	sessionAbsoluteCap, err = config.ParseDuration(environ, sessionAbsoluteCapThreshold)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	nonceRetention, err = config.ParseDuration(environ, nonceRetentionThreshold)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return magicLinkExpiry, sessionLifetime, sessionAbsoluteCap, nonceRetention, nil
}
