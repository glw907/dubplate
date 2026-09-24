package main

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
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

// nonceRetentionCeiling bounds how long a consumed link's nonce record
// stays on disk. Ten times the 24h default, on the same ten-times rule
// fork 3's four rate ceilings already follow (linkRequestsPerAddressCeiling):
// generous headroom above the default without admitting a retention
// that keeps every spent nonce indefinitely.
const nonceRetentionCeiling = 240 * time.Hour

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

// rateFloor is the positive floor every one of fork 3's four rate
// thresholds shares: zero is ambiguous between "unlimited" and "blocked
// outright", so an operator who wants either extreme states an explicit
// large or small number instead.
const rateFloor = 1

// linkRequestsPerAddressCeiling, linkRequestsPerClientCeiling,
// consumptionAttemptsPerClientCeiling, and linkMailCeilingCeiling bound
// how far an adopter can raise fork 3's four rates above their default.
// Each is ten times its threshold's own default, generous headroom for
// a group larger than the estate's own without admitting a rate that
// defeats R28's own reasoning: a per-address or per-client limiter set
// high enough stops bounding anything an enumerator does.
const (
	linkRequestsPerAddressCeiling       = 50
	linkRequestsPerClientCeiling        = 200
	consumptionAttemptsPerClientCeiling = 300
	linkMailCeilingCeiling              = 600
)

// authRequestBodyCapFloor and authRequestBodyCapCeiling bound
// DUBPLATE_AUTH_REQUEST_BODY_CAP. The floor, 512 bytes, comfortably
// fits an RFC 5321 address (254 bytes at most) plus its form-encoded
// field name; the ceiling, 64KiB, is generous headroom that still
// keeps security constraint 3's amplification bound (M5) meaningful,
// because a cap large enough to admit an arbitrary upload defeats the
// reason the cap exists. The default, 4096 bytes, is stated as a
// literal: the login form the auth handlers accept carries only a
// mailed-address field, and 4KiB comfortably fits that field plus any
// form-encoding overhead with room to spare.
const (
	authRequestBodyCapFloor   = 512
	authRequestBodyCapCeiling = 65536
)

// Fork 3's four remaining rate thresholds (R28), the auth request body
// cap security constraint 3 requires (spec:1590), and the unknown-
// address answer fork 4 rules (R25, R-5b22) are rung 5's own remaining
// six thresholds: spec:1587-1593 lists ten, and 5b-i landed the other
// four above.
var (
	linkRequestsPerAddressThreshold = config.Threshold{
		Key:     "DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR",
		Kind:    config.KindInt,
		Default: "5",
		Doc:     "a positive integer, for example 5",
	}
	linkRequestsPerClientThreshold = config.Threshold{
		Key:     "DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR",
		Kind:    config.KindInt,
		Default: "20",
		Doc:     "a positive integer, for example 20",
	}
	consumptionAttemptsPerClientThreshold = config.Threshold{
		Key:     "DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR",
		Kind:    config.KindInt,
		Default: "30",
		Doc:     "a positive integer, for example 30",
	}
	linkMailCeilingThreshold = config.Threshold{
		Key:     "DUBPLATE_LINK_MAIL_CEILING_PER_HOUR",
		Kind:    config.KindInt,
		Default: "60",
		Doc:     "a positive integer, for example 60",
	}
	authRequestBodyCapThreshold = config.Threshold{
		Key:     "DUBPLATE_AUTH_REQUEST_BODY_CAP",
		Kind:    config.KindInt,
		Default: "4096",
		Doc:     "a positive integer of bytes, for example 4096",
	}
	unknownAddressAnswerThreshold = config.Threshold{
		Key:     "DUBPLATE_UNKNOWN_ADDRESS_ANSWER",
		Kind:    config.KindString,
		Default: "kind",
		Doc:     "one of kind or generic",
	}
)

// registerAuthConfig adds rung 5's twelve names and ten thresholds to
// reg: the selector as an owned wall row, magic-link as a declared
// identity source, R-5b10's first four thresholds, and the six
// remaining thresholds spec:1587-1593 lists. It is a plain function on
// registerServeConfig's own precedent (serve.go), called once from the
// binary's shared registry constructor (newProductionRegistry,
// main.go).
func registerAuthConfig(reg *config.Registry) {
	reg.AddWallRow(identitySourceSelectorRow)
	reg.AddIdentitySource(magicLinkSource)
	reg.AddThreshold(magicLinkExpiryThreshold)
	reg.AddThreshold(sessionLifetimeThreshold)
	reg.AddThreshold(sessionAbsoluteCapThreshold)
	reg.AddThreshold(nonceRetentionThreshold)
	reg.AddThreshold(linkRequestsPerAddressThreshold)
	reg.AddThreshold(linkRequestsPerClientThreshold)
	reg.AddThreshold(consumptionAttemptsPerClientThreshold)
	reg.AddThreshold(linkMailCeilingThreshold)
	reg.AddThreshold(authRequestBodyCapThreshold)
	reg.AddThreshold(unknownAddressAnswerThreshold)
}

// sourceKeyError names one identity source key and the reason it
// refuses, letting doctor recover the key by reading this type's own
// field through errors.As rather than by parsing the rendered message:
// a reason that happens to carry a colon of its own cannot be mistaken
// for a key. Error renders "KEY: reason", so every refusal built from
// this type reads exactly as the plain-string refusals it replaces.
type sourceKeyError struct {
	key    string
	reason string
}

func (e *sourceKeyError) Error() string {
	return e.key + ": " + e.reason
}

// validateMagicLinkSource is [magicLinkSource]'s Validate hook: it
// checks the secret length floor on both rotation pairs, the public
// URL's shape, the ten thresholds' bounds (review m5, R28, security
// constraint 3), the session-lifetime-vs-absolute-cap and
// nonce-retention-vs-expiry relations, the unknown-address answer's
// accepted set (R25), DUBPLATE_SMTP_FROM (spec:1289), and the SMTP port
// bound, joining every failure it finds so [config.Load] reports a
// short secret and a bad URL together. Every returned error names the
// offending key and never the value or the length of a secret, so a
// refusal narrows no guess about what the value was.
func validateMagicLinkSource(values map[string]string, environ []string) error {
	var errs []error

	for _, key := range []string{
		"DUBPLATE_MAGIC_LINK_SECRET",
		"DUBPLATE_SESSION_SECRET",
		"DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS",
		"DUBPLATE_SESSION_SECRET_PREVIOUS",
	} {
		if v, ok := values[key]; ok && len(v) < secretLengthFloor {
			errs = append(errs, &sourceKeyError{key: key, reason: fmt.Sprintf("too short, want at least %d bytes", secretLengthFloor)})
		}
	}

	if v, ok := values["DUBPLATE_PUBLIC_URL"]; ok {
		if err := validatePublicURL(v); err != nil {
			errs = append(errs, err)
		}
	}

	if v, ok := values["DUBPLATE_SMTP_FROM"]; ok {
		if err := validateSMTPFrom(v); err != nil {
			errs = append(errs, err)
		}
	}

	if v, ok := values[smtpPortKey]; ok {
		if err := validateSMTPPort(v); err != nil {
			errs = append(errs, err)
		}
	}

	var expiry, retention, lifetime, absoluteCap time.Duration
	var expiryErr, retentionErr, lifetimeErr, absoluteCapErr error
	errs, expiry, expiryErr = checkDurationBounds(errs, environ, magicLinkExpiryThreshold, magicLinkExpiryCeiling)
	errs, lifetime, lifetimeErr = checkDurationBounds(errs, environ, sessionLifetimeThreshold, 0)
	errs, absoluteCap, absoluteCapErr = checkDurationBounds(errs, environ, sessionAbsoluteCapThreshold, sessionAbsoluteCapCeiling)
	errs, retention, retentionErr = checkDurationBounds(errs, environ, nonceRetentionThreshold, nonceRetentionCeiling)

	// A relational refusal names both keys it compares: the pair below
	// appends one sourceKeyError per key rather than one error naming
	// only the first, so doctor's per-key walk reports both sides of
	// the comparison.
	if expiryErr == nil && retentionErr == nil && retention < expiry {
		errs = append(errs,
			&sourceKeyError{key: nonceRetentionThreshold.Key, reason: fmt.Sprintf("%s is below %s's %s", retention, magicLinkExpiryThreshold.Key, expiry)},
			&sourceKeyError{key: magicLinkExpiryThreshold.Key, reason: fmt.Sprintf("%s is above %s's %s", expiry, nonceRetentionThreshold.Key, retention)},
		)
	}
	if lifetimeErr == nil && absoluteCapErr == nil && lifetime > absoluteCap {
		errs = append(errs,
			&sourceKeyError{key: sessionLifetimeThreshold.Key, reason: fmt.Sprintf("%s is above %s's %s", lifetime, sessionAbsoluteCapThreshold.Key, absoluteCap)},
			&sourceKeyError{key: sessionAbsoluteCapThreshold.Key, reason: fmt.Sprintf("%s is below %s's %s", absoluteCap, sessionLifetimeThreshold.Key, lifetime)},
		)
	}

	errs = checkIntBounds(errs, environ, linkRequestsPerAddressThreshold, rateFloor, linkRequestsPerAddressCeiling)
	errs = checkIntBounds(errs, environ, linkRequestsPerClientThreshold, rateFloor, linkRequestsPerClientCeiling)
	errs = checkIntBounds(errs, environ, consumptionAttemptsPerClientThreshold, rateFloor, consumptionAttemptsPerClientCeiling)
	errs = checkIntBounds(errs, environ, linkMailCeilingThreshold, rateFloor, linkMailCeilingCeiling)
	errs = checkIntBounds(errs, environ, authRequestBodyCapThreshold, authRequestBodyCapFloor, authRequestBodyCapCeiling)

	if answer, err := config.ParseString(environ, unknownAddressAnswerThreshold); err != nil {
		errs = append(errs, err)
	} else if answer != "kind" && answer != "generic" {
		errs = append(errs, fmt.Errorf("%s: %q is not one of kind, generic", unknownAddressAnswerThreshold.Key, answer))
	}

	return errors.Join(errs...)
}

// validateSMTPPort checks value parses as an integer from 1 through
// 65535. DUBPLATE_SMTP_PORT is a required source key rather than a
// registered [config.Threshold], because it feeds
// [strconv.Atoi] rather than one of the registry's parseable kinds, so
// its bound lives here instead of in [checkIntBounds]. Landing the
// check in this Validate hook, rather than leaving it to newMailer's
// own parse, is what makes doctor and every verb see a bad port before
// the serve verb ever dials out.
func validateSMTPPort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s: not an integer", smtpPortKey)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s: %d is not in 1 through 65535", smtpPortKey, port)
	}
	return nil
}

// validateSMTPFrom checks value the way [config.Load] checks
// DUBPLATE_ADMIN_EMAIL (config.go's requireEmail): it parses as a
// single RFC 5322 address, and it carries no carriage return and no
// line feed, so a value that would ride into a mail header cannot
// inject one (spec:1289). The check runs at load rather than at send,
// because R9 turns every failed send into a failed login, and a typo
// caught only at send time would otherwise pass doctor and start
// clean. The returned error never repeats value.
func validateSMTPFrom(value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return errors.New("DUBPLATE_SMTP_FROM: invalid email address")
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return errors.New("DUBPLATE_SMTP_FROM: invalid email address")
	}
	return nil
}

// checkIntBounds is [checkDurationBounds]'s integer counterpart, minus
// the parsed value: none of the six thresholds it checks feeds a
// further relational check the way the nonce-retention-vs-expiry pair
// does, so this returns only the extended errs.
func checkIntBounds(errs []error, environ []string, th config.Threshold, floor, ceiling int) []error {
	n, err := config.ParseInt(environ, th)
	if err != nil {
		return append(errs, err)
	}
	if n < floor {
		errs = append(errs, fmt.Errorf("%s: %d is below the floor of %d", th.Key, n, floor))
	}
	if n > ceiling {
		errs = append(errs, fmt.Errorf("%s: %d is above the ceiling of %d", th.Key, n, ceiling))
	}
	return errs
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
// userinfo, no query string, no fragment, and no path other than empty
// or "/", because any of the four would ride into every link [LinkURL]
// builds and make an attacker-chosen string part of its visible prefix
// or suffix (review m4): the owned serve prefix is what sets where
// dubplate lives, so the public URL carries none of its own; it
// carries a non-empty host, because net/url
// accepts a bare hostname with a nil error and an empty Host, and a URL
// with no host builds a path-only link that defeats security review
// finding 7; and its scheme is https unless its host is loopback, the
// carve-out the W3C Secure Contexts definition of a potentially
// trustworthy origin extends to the whole loopback set
// (https://w3c.github.io/webappsec-secure-contexts/#is-origin-trustworthy,
// read 2026-09-12), not the name localhost alone (review m6, ruled,
// conductor 2026-09-12). The comparison reads the parsed hostname
// rather than the host, since url.Parse("http://localhost:8080").Host
// is "localhost:8080" and only Hostname() is "localhost".
//
// No returned error quotes value: every message names the key and the
// defect alone. url.Parse's own error is not
// wrapped verbatim either, since its Error method formats the whole
// URL, password included
// (net/url/url.go, fmt.Sprintf("%s %q: %s", e.Op, e.URL, e.Err), read
// 2026-09-23); this reports the parse failure's own inner reason
// through [errors.AsType] instead.
func validatePublicURL(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		if uerr, ok := errors.AsType[*url.Error](err); ok {
			return &sourceKeyError{key: publicURLKey, reason: fmt.Sprintf("invalid URL: %v", uerr.Err)}
		}
		return &sourceKeyError{key: publicURLKey, reason: "invalid URL"}
	}
	if u.User != nil {
		return &sourceKeyError{key: publicURLKey, reason: "carries userinfo, which would become an attacker-chosen prefix of every mailed link"}
	}
	if u.RawQuery != "" {
		return &sourceKeyError{key: publicURLKey, reason: "carries a query string"}
	}
	if u.Fragment != "" {
		return &sourceKeyError{key: publicURLKey, reason: "carries a fragment"}
	}
	if u.Path != "" && u.Path != "/" {
		return &sourceKeyError{key: publicURLKey, reason: "must carry no path; the serve prefix sets where dubplate lives"}
	}
	host := u.Hostname()
	if host == "" {
		return &sourceKeyError{key: publicURLKey, reason: "has no host"}
	}
	if u.Scheme != "https" && !isLoopbackHost(host) {
		return &sourceKeyError{key: publicURLKey, reason: "must use https unless its host is loopback"}
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

// authConfig is resolveAuthConfig's return type: rung 5's ten
// thresholds, resolved and typed, the auth core takes as plain
// construction arguments. It replaces a positional return once ten
// values of mixed type, durations, ints, and a string, made a
// positional return unreadable; a struct field's name says which
// threshold it came from at every call site, where a positional return
// said so only in the signature.
type authConfig struct {
	magicLinkExpiry    time.Duration
	sessionLifetime    time.Duration
	sessionAbsoluteCap time.Duration
	nonceRetention     time.Duration

	linkRequestsPerAddress       int
	linkRequestsPerClient        int
	consumptionAttemptsPerClient int
	linkMailCeiling              int

	authRequestBodyCap int

	unknownAddressAnswer string
}

// resolveAuthConfig reads rung 5's ten thresholds against environ and
// returns an [authConfig], on [resolveServeConfig]'s own precedent
// (serve.go:81-106): internal/auth never reads the environment itself.
func resolveAuthConfig(environ []string) (authConfig, error) {
	var cfg authConfig
	var err error

	cfg.magicLinkExpiry, err = config.ParseDuration(environ, magicLinkExpiryThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.sessionLifetime, err = config.ParseDuration(environ, sessionLifetimeThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.sessionAbsoluteCap, err = config.ParseDuration(environ, sessionAbsoluteCapThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.nonceRetention, err = config.ParseDuration(environ, nonceRetentionThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.linkRequestsPerAddress, err = config.ParseInt(environ, linkRequestsPerAddressThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.linkRequestsPerClient, err = config.ParseInt(environ, linkRequestsPerClientThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.consumptionAttemptsPerClient, err = config.ParseInt(environ, consumptionAttemptsPerClientThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.linkMailCeiling, err = config.ParseInt(environ, linkMailCeilingThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.authRequestBodyCap, err = config.ParseInt(environ, authRequestBodyCapThreshold)
	if err != nil {
		return authConfig{}, err
	}
	cfg.unknownAddressAnswer, err = config.ParseString(environ, unknownAddressAnswerThreshold)
	if err != nil {
		return authConfig{}, err
	}
	return cfg, nil
}
