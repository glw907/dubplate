package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
)

// identitySourceSelectorRow declares DUBPLATE_IDENTITY_SOURCE as an
// owned wall row: the selector is required, alongside
// [magicLinkIdentitySource] itself.
var identitySourceSelectorRow = config.WallRow{
	Key:         "DUBPLATE_IDENTITY_SOURCE",
	Description: "the name of the identity source that signs members in",
	Sensitivity: config.Sensitive,
}

// The magic-link identity source's key names: [magicLinkIdentitySource]
// declares them, its Validate hook checks them, and serve reads them
// through [config.Config.IdentitySourceValue]. None of these constants
// carries a secret value; each names the environment variable a real
// deployment carries one under.
const (
	MagicLinkSecretKey         = "DUBPLATE_MAGIC_LINK_SECRET"          // secret-guard-allow
	MagicLinkSecretPreviousKey = "DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS" // secret-guard-allow
	SessionSecretKey           = "DUBPLATE_SESSION_SECRET"             // secret-guard-allow
	SessionSecretPreviousKey   = "DUBPLATE_SESSION_SECRET_PREVIOUS"    // secret-guard-allow
	PublicURLKey               = "DUBPLATE_PUBLIC_URL"
	SMTPHostKey                = "DUBPLATE_SMTP_HOST"
	SMTPPortKey                = "DUBPLATE_SMTP_PORT"
	SMTPUserKey                = "DUBPLATE_SMTP_USER"
	SMTPPasswordKey            = "DUBPLATE_SMTP_PASSWORD" // secret-guard-allow
	SMTPFromKey                = "DUBPLATE_SMTP_FROM"
)

// secretKeys names every key [validateMagicLinkSource] holds to
// [secretLengthFloor]: both HMAC secrets and their rotation-overlap
// predecessors.
var secretKeys = []string{MagicLinkSecretKey, SessionSecretKey, MagicLinkSecretPreviousKey, SessionSecretPreviousKey}

// magicLinkIdentitySource is the shipped default identity source: a login link
// mailed to a roster address, backed by an HMAC-signed session cookie.
// It declares eight required names, the five DUBPLATE_SMTP_* names
// among them, each required when the source is selected. The two
// optional names carry a rotation-overlap secret each, so a deployment
// can roll a secret without a window where every existing token or
// session fails at once: they are the source's Optional keys rather
// than registered thresholds, because a threshold's KindString default
// must be non-empty and doctor prints a threshold's effective value,
// which would print a secret on an operator's terminal.
//
// Validate is [validateMagicLinkSource], the source's value rules: the
// secret length floor, the public URL's shape, and the mail and serve
// ports.
var magicLinkIdentitySource = config.IdentitySource{
	Name:     "magic-link",
	Selector: identitySourceSelectorRow.Key,
	Required: []config.KeyDecl{
		{Key: MagicLinkSecretKey, Description: "the secret that signs sign-in links, at least 32 bytes", Sensitivity: config.Sensitive},
		{Key: SessionSecretKey, Description: "the secret that signs session cookies, at least 32 bytes", Sensitivity: config.Sensitive},
		{Key: PublicURLKey, Description: "the address members reach the server at, used in sign-in links", Sensitivity: config.Sensitive},
		{Key: SMTPHostKey, Description: "the mail server the server sends sign-in links through", Sensitivity: config.Sensitive},
		{Key: SMTPPortKey, Description: "the port of that mail server", Sensitivity: config.Sensitive},
		{Key: SMTPUserKey, Description: "the user name the server signs in to the mail server with", Sensitivity: config.Sensitive},
		{Key: SMTPPasswordKey, Description: "the password the server signs in to the mail server with", Sensitivity: config.Sensitive},
		{Key: SMTPFromKey, Description: "the sender address on the mail the server sends", Sensitivity: config.Sensitive},
	},
	Optional: []config.KeyDecl{
		{Key: MagicLinkSecretPreviousKey, Description: "the previous link-signing secret, kept so links minted before a rotation still verify", Sensitivity: config.Sensitive},
		{Key: SessionSecretPreviousKey, Description: "the previous session-signing secret, kept so sessions minted before a rotation still verify", Sensitivity: config.Sensitive},
	},
	Validate: validateMagicLinkSource,
}

// secretLengthFloor is the minimum accepted byte length of a magic-link
// or session secret, HMAC-SHA256's own recommended key length: RFC 2104
// section 3 (https://www.rfc-editor.org/rfc/rfc2104#section-3, read
// 2026-09-12) recommends a key at least as long as the hash's output,
// 32 bytes for SHA-256. The count is of raw bytes with no
// transformation, so an adopter pasting a base64 rendering of 32 random
// bytes passes with room to spare.
const secretLengthFloor = 32

// MagicLinkExpiryCeiling bounds how long a mailed link can stay
// redeemable. The threat model's mail-scanner residual is bounded by
// "single use, a 15-minute default expiry"; an hour is generous
// headroom above the default without admitting a link that survives a
// whole day in a mail provider's link-scanner cache or a browser's
// history (review m5). It is [MagicLinkExpirySetting]'s registered
// ceiling, and the nonce store's mtime fallback keeps a record past it
// rather than past the value in force, since only the ceiling outlives
// every token a changed expiry could have minted.
const MagicLinkExpiryCeiling = time.Hour

// MagicLinkExpirySetting, SessionLifetimeSetting,
// SessionAbsoluteCapSetting, and NonceRetentionSetting are the four
// auth durations.
//
// Every duration takes a positive floor, since zero or a negative
// duration cannot mean "never" or "immediately" here (review m5). The
// link expiry's floor is a minute instead of a second: a link that
// expires inside a minute leaves no time for the mail to arrive, and
// the mailed copy would read "0 minutes". The absolute cap's ceiling,
// two years, is generous headroom above its 4320h (180-day) default
// without admitting a cap that is, in every practical sense,
// permanent, and the lifetime shares it, since a lifetime above the
// cap is refused anyway. The nonce retention's ceiling, 240h, is ten
// times its default, the ten-times rule the four rates follow.
//
// The link expiry and both session durations are sensitive: a live
// change to any of them changes how long every member stays signed in
// or how long a mailed link stays redeemable.
var (
	MagicLinkExpirySetting = config.Setting{
		Key:         "DUBPLATE_MAGIC_LINK_EXPIRY",
		Kind:        config.KindDuration,
		Default:     "15m",
		Doc:         "a Go duration string, for example 15m",
		Apply:       config.ApplyNextOperation,
		Floor:       "1m",
		Ceiling:     MagicLinkExpiryCeiling.String(),
		Description: "how long a mailed sign-in link stays redeemable",
		Sensitivity: config.Sensitive,
	}
	SessionLifetimeSetting = config.Setting{
		Key:         "DUBPLATE_SESSION_LIFETIME",
		Kind:        config.KindDuration,
		Default:     "2160h",
		Doc:         "a Go duration string, for example 2160h (90 days)",
		Apply:       config.ApplyNextOperation,
		Floor:       "1s",
		Ceiling:     "17520h",
		Description: "how long a session stays valid before the member signs in again, unless it is refreshed",
		Sensitivity: config.Sensitive,
	}
	SessionAbsoluteCapSetting = config.Setting{
		Key:         "DUBPLATE_SESSION_ABSOLUTE_CAP",
		Kind:        config.KindDuration,
		Default:     "4320h",
		Doc:         "a Go duration string, for example 4320h (180 days)",
		Apply:       config.ApplyImmediately,
		Floor:       "1s",
		Ceiling:     "17520h",
		Description: "the longest a session can last from its first sign-in, however often it is refreshed",
		Sensitivity: config.Sensitive,
	}
	NonceRetentionSetting = config.Setting{
		Key:         "DUBPLATE_NONCE_RETENTION",
		Kind:        config.KindDuration,
		Default:     "24h",
		Doc:         "a Go duration string, for example 24h",
		Apply:       config.ApplyNextOperation,
		Floor:       "1s",
		Ceiling:     "240h",
		Description: "how long the server remembers a used sign-in link so it cannot be used twice",
		Sensitivity: config.NotSensitive,
	}
)

// Fork 3's four rate limits (R28) each take a floor of one, since zero
// is ambiguous between "unlimited" and "blocked outright", and a
// ceiling of ten times the default: generous headroom for a larger
// group without admitting a rate that defeats R28's own reasoning, as
// a per-address or per-client limiter set high enough stops bounding
// anything an enumerator does. The limiter is built once, so each
// applies at the next start.
var (
	linkRequestsPerAddressSetting = config.Setting{
		Key:         "DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR",
		Kind:        config.KindInt,
		Default:     "5",
		Doc:         "a positive integer, for example 5",
		Apply:       config.ApplyRestart,
		Floor:       "1",
		Ceiling:     "50",
		Description: "how many sign-in links one address can request in an hour",
		Sensitivity: config.NotSensitive,
	}
	linkRequestsPerClientSetting = config.Setting{
		Key:         "DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR",
		Kind:        config.KindInt,
		Default:     "20",
		Doc:         "a positive integer, for example 20",
		Apply:       config.ApplyRestart,
		Floor:       "1",
		Ceiling:     "200",
		Description: "how many sign-in links one client can request in an hour",
		Sensitivity: config.NotSensitive,
	}
	consumptionAttemptsPerClientSetting = config.Setting{
		Key:         "DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR",
		Kind:        config.KindInt,
		Default:     "30",
		Doc:         "a positive integer, for example 30",
		Apply:       config.ApplyRestart,
		Floor:       "1",
		Ceiling:     "300",
		Description: "how many sign-in link redemptions one client can attempt in an hour",
		Sensitivity: config.NotSensitive,
	}
	linkMailCeilingSetting = config.Setting{
		Key:         "DUBPLATE_LINK_MAIL_CEILING_PER_HOUR",
		Kind:        config.KindInt,
		Default:     "60",
		Doc:         "a positive integer, for example 60",
		Apply:       config.ApplyRestart,
		Floor:       "1",
		Ceiling:     "600",
		Description: "how many sign-in mails the server sends in an hour, across every address",
		Sensitivity: config.NotSensitive,
	}
)

// RequestBodyCapCeiling is the auth request body cap's ceiling.
// Sign-out bounds its body with it directly, so the one auth endpoint
// no limiter charges reads no setting.
const RequestBodyCapCeiling = 65536

// authRequestBodyCapSetting is the auth request body cap security
// constraint 3 requires. Its floor, 512 bytes, comfortably fits an RFC
// 5321 address (254 bytes at most) plus its form-encoded field name;
// its ceiling, 64KiB, is generous headroom that still keeps the
// amplification bound (M5) meaningful, because a cap large enough to
// admit an arbitrary upload defeats the reason the cap exists. Its
// 4096-byte default fits the login form's one mailed-address field
// with room to spare.
var authRequestBodyCapSetting = config.Setting{
	Key:         "DUBPLATE_AUTH_REQUEST_BODY_CAP",
	Kind:        config.KindInt,
	Default:     "4096",
	Doc:         "a positive integer of bytes, for example 4096",
	Apply:       config.ApplyNextOperation,
	Floor:       "512",
	Ceiling:     strconv.Itoa(RequestBodyCapCeiling),
	Description: "the largest sign-in request body the server reads, in bytes",
	Sensitivity: config.NotSensitive,
}

// unknownAddressAnswerSetting is the answer to a login request for an
// address no member holds: "kind" says so, and
// "generic" answers exactly as for a member.
var unknownAddressAnswerSetting = config.Setting{
	Key:         "DUBPLATE_UNKNOWN_ADDRESS_ANSWER",
	Kind:        config.KindString,
	Default:     "kind",
	Doc:         "one of kind or generic",
	Apply:       config.ApplyNextOperation,
	Values:      []string{"kind", "generic"},
	Description: "whether a sign-in request for an address no member holds is told so (kind) or answered as for a member (generic)",
	Sensitivity: config.Sensitive,
}

// nonceRetentionCoversExpiry and sessionLifetimeWithinCap are the
// auth settings' two relations no single bound expresses. A consumed
// link's nonce must outlive every token that could still carry it, or
// the link replays once its record is swept; and a session lifetime
// above the absolute cap mints a cookie the server itself refuses
// before it expires. Each refusal names both keys.
var (
	nonceRetentionCoversExpiry = config.CorrelatedRule{
		Name:  "nonce retention covers the link expiry",
		Keys:  []string{MagicLinkExpirySetting.Key, NonceRetentionSetting.Key},
		Check: durationAtLeast(NonceRetentionSetting.Key, MagicLinkExpirySetting.Key),
	}
	sessionLifetimeWithinCap = config.CorrelatedRule{
		Name:  "session lifetime within the absolute cap",
		Keys:  []string{SessionLifetimeSetting.Key, SessionAbsoluteCapSetting.Key},
		Check: durationAtLeast(SessionAbsoluteCapSetting.Key, SessionLifetimeSetting.Key),
	}
)

// durationAtLeast returns a correlated check that the duration in
// force for longer is at least the one for shorter, refusing with an
// error that names both keys and both values. Each value has already
// passed its own setting's validator, so a parse failure here is a
// registration defect, reported rather than passed.
func durationAtLeast(longer, shorter string) func(map[string]string) error {
	return func(values map[string]string) error {
		l, err := time.ParseDuration(values[longer])
		if err != nil {
			return fmt.Errorf("%s: %w", longer, err)
		}
		s, err := time.ParseDuration(values[shorter])
		if err != nil {
			return fmt.Errorf("%s: %w", shorter, err)
		}
		if l < s {
			return fmt.Errorf("%s %q is below %s %q", longer, values[longer], shorter, values[shorter])
		}
		return nil
	}
}

// rosterFileKey names the pre-store roster file's location key, which
// the wall now rejects with rosterFileSupersession.
const rosterFileKey = "DUBPLATE_ROSTER_FILE"

// rosterFileSupersession is the message the wall reports for
// rosterFileKey: members live in the store, and the first admin comes
// from the member verb.
const rosterFileSupersession = "superseded: members live in the store; " +
	"add the first admin with dubplate-server member add ADDRESS --admin"

// Register adds the auth configuration to reg: the
// selector as an owned wall row, magic-link as a declared identity
// source, the retired roster file key as a superseded name, the ten
// auth settings, and their two correlated rules, registered after the
// settings they name. It is a plain function on [serve.Register]'s
// precedent, called once from the binary's shared registry constructor.
func Register(reg *config.Registry) {
	reg.AddWallRow(identitySourceSelectorRow)
	reg.AddIdentitySource(magicLinkIdentitySource)
	reg.AddSuperseded(rosterFileKey, rosterFileSupersession)
	for _, s := range []config.Setting{
		MagicLinkExpirySetting,
		SessionLifetimeSetting,
		SessionAbsoluteCapSetting,
		NonceRetentionSetting,
		linkRequestsPerAddressSetting,
		linkRequestsPerClientSetting,
		consumptionAttemptsPerClientSetting,
		linkMailCeilingSetting,
		authRequestBodyCapSetting,
		unknownAddressAnswerSetting,
	} {
		reg.AddSetting(s)
	}
	reg.AddCorrelatedRule(nonceRetentionCoversExpiry)
	reg.AddCorrelatedRule(sessionLifetimeWithinCap)
}

// validateMagicLinkSource is [magicLinkIdentitySource]'s Validate hook: it
// checks the secret length floor on both rotation pairs, the public
// URL's shape, DUBPLATE_SMTP_FROM (spec:1289), the SMTP port bound,
// and the serve port bound, joining every failure it finds so
// [config.Load] reports a short secret and a bad URL together. Every
// returned error names the offending key and never the value or the
// length of a secret, so a refusal narrows no guess about what the
// value was. The auth settings' bounds are the seed rule's, not this
// hook's.
func validateMagicLinkSource(values map[string]string, environ []string) error {
	var errs []error

	for _, key := range secretKeys {
		if v, ok := values[key]; ok && len(v) < secretLengthFloor {
			errs = append(errs, &config.KeyError{Key: key, Reason: fmt.Sprintf("too short, want at least %d bytes", secretLengthFloor)})
		}
	}

	if v, ok := values[PublicURLKey]; ok {
		if err := validatePublicURL(v); err != nil {
			errs = append(errs, err)
		}
	}

	if v, ok := values[SMTPFromKey]; ok {
		if err := validateSMTPFrom(v); err != nil {
			errs = append(errs, err)
		}
	}

	if v, ok := values[SMTPPortKey]; ok {
		if err := validateSMTPPort(v); err != nil {
			errs = append(errs, err)
		}
	}

	if err := serve.ValidatePort(environ); err != nil {
		errs = append(errs, err)
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
		return &config.KeyError{Key: SMTPPortKey, Reason: "not an integer"}
	}
	if port < 1 || port > 65535 {
		return &config.KeyError{Key: SMTPPortKey, Reason: fmt.Sprintf("%d is not in 1 through 65535", port)}
	}
	return nil
}

// validateSMTPFrom checks that value parses as a single RFC 5322
// address, display name allowed, and carries no carriage return and no
// line feed, so a value that would ride into a mail header cannot
// inject one (spec:1289). The check runs at load rather than at send,
// because R9 turns every failed send into a failed login, and a typo
// caught only at send time would otherwise pass doctor and start
// clean. The returned error never repeats value.
func validateSMTPFrom(value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return &config.KeyError{Key: SMTPFromKey, Reason: "invalid email address"}
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return &config.KeyError{Key: SMTPFromKey, Reason: "invalid email address"}
	}
	return nil
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
// No returned error quotes value or any part of it: every message names
// the key and the defect alone. url.Parse's own error is never wrapped,
// even its inner reason, because a raw value carrying userinfo with a
// "/", "?" or "#" inside the password makes url.Parse cut the authority
// there and misread "user:pass" as "host:port", so its error quotes the
// password up to that character
// (net/url/url.go, fmt.Sprintf("%s %q: %s", e.Op, e.URL, e.Err), read
// 2026-09-23). value is checked for "@" before it is parsed at all, so
// a value net/url would misparse this way is still caught as userinfo
// rather than reaching url.Parse's error path.
func validatePublicURL(value string) error {
	if strings.Contains(value, "@") {
		return &config.KeyError{Key: PublicURLKey, Reason: "carries userinfo, which would become an attacker-chosen prefix of every mailed link"}
	}
	u, err := url.Parse(value)
	if err != nil {
		return &config.KeyError{Key: PublicURLKey, Reason: "invalid URL"}
	}
	if u.User != nil {
		return &config.KeyError{Key: PublicURLKey, Reason: "carries userinfo, which would become an attacker-chosen prefix of every mailed link"}
	}
	if u.RawQuery != "" {
		return &config.KeyError{Key: PublicURLKey, Reason: "carries a query string"}
	}
	if u.Fragment != "" {
		return &config.KeyError{Key: PublicURLKey, Reason: "carries a fragment"}
	}
	if u.Path != "" && u.Path != "/" {
		return &config.KeyError{Key: PublicURLKey, Reason: "must carry no path; the serve prefix sets where dubplate lives"}
	}
	host := u.Hostname()
	if host == "" {
		return &config.KeyError{Key: PublicURLKey, Reason: "has no host"}
	}
	if u.Scheme != "https" && !IsLoopbackHost(host) {
		return &config.KeyError{Key: PublicURLKey, Reason: "must use https unless its host is loopback"}
	}
	return nil
}

// IsLoopbackHost reports whether host is a loopback name or an IP
// address in the loopback range, the whole loopback set the HTTPS
// carve-out covers. A loopback name is localhost or any name under
// .localhost, matched case-insensitively and with one trailing dot
// removed, since RFC 6761 section 6.3 reserves that whole domain for
// loopback and a browser resolves every name in it there.
func IsLoopbackHost(host string) bool {
	name := strings.TrimSuffix(strings.ToLower(host), ".")
	if name == "localhost" || strings.HasSuffix(name, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// DurationSetting returns an [Setting] reading s's value in force
// through acc at each call.
func DurationSetting(acc *settings.Accessor, s config.Setting) Setting[time.Duration] {
	return func(ctx context.Context) (time.Duration, error) {
		d, err := acc.Duration(ctx, s)
		return d, settingReadError(s, err)
	}
}

// BodyCap returns an [Setting] reading the auth request body
// cap in force through acc at each call.
func BodyCap(acc *settings.Accessor) Setting[int64] {
	return func(ctx context.Context) (int64, error) {
		n, err := acc.Int(ctx, authRequestBodyCapSetting)
		return int64(n), settingReadError(authRequestBodyCapSetting, err)
	}
}

// GenericAnswer returns an [Setting] reporting at each call
// whether the unknown-address answer in force through acc is generic.
func GenericAnswer(acc *settings.Accessor) Setting[bool] {
	return func(ctx context.Context) (bool, error) {
		v, err := acc.String(ctx, unknownAddressAnswerSetting)
		return v == "generic", settingReadError(unknownAddressAnswerSetting, err)
	}
}

// settingReadError returns err, a failed read of s, as the
// [SettingError] the auth handlers log by key, or nil when err is
// nil.
func settingReadError(s config.Setting, err error) error {
	if err == nil {
		return nil
	}
	return &SettingError{Key: s.Key, Err: err}
}

// ReadLimits reads the four rate limits in force through acc. They
// apply at the next start, so serve reads them once, after its seed
// apply, and builds its limiter from them; a failed read fails the
// start naming the key, never falling back to a default.
func ReadLimits(ctx context.Context, acc *settings.Accessor) (Limits, error) {
	var limits Limits
	for _, n := range []struct {
		dst *int
		s   config.Setting
	}{
		{&limits.LinkRequestsPerAddress, linkRequestsPerAddressSetting},
		{&limits.LinkRequestsPerClient, linkRequestsPerClientSetting},
		{&limits.ConsumptionAttemptsPerClient, consumptionAttemptsPerClientSetting},
		{&limits.LinkMailCeiling, linkMailCeilingSetting},
	} {
		v, err := acc.Int(ctx, n.s)
		if err != nil {
			return Limits{}, err
		}
		*n.dst = v
	}
	return limits, nil
}
