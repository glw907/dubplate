package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/notify"
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

// The magic-link identity source's key names. [magicLinkIdentitySource]
// declares them, its Validate hook checks them, and [ConfigFrom] reads
// them; restore names the four secrets in its rotation advice. None of
// these constants carries a secret value. Each names the environment
// variable a real deployment carries one under.
const (
	MagicLinkSecretKey         = "DUBPLATE_MAGIC_LINK_SECRET"          // secret-guard-allow
	MagicLinkSecretPreviousKey = "DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS" // secret-guard-allow
	SessionSecretKey           = "DUBPLATE_SESSION_SECRET"             // secret-guard-allow
	SessionSecretPreviousKey   = "DUBPLATE_SESSION_SECRET_PREVIOUS"    // secret-guard-allow
	publicURLKey               = "DUBPLATE_PUBLIC_URL"
)

// secretKeys names every key [validateMagicLinkSource] holds to
// [secretLengthFloor]: both HMAC secrets and their rotation-overlap
// predecessors.
var secretKeys = []string{MagicLinkSecretKey, SessionSecretKey, MagicLinkSecretPreviousKey, SessionSecretPreviousKey}

// magicLinkIdentitySource is the shipped default identity source: a login link
// mailed to a roster address, backed by an HMAC-signed session cookie.
// It declares eight required names, the five SMTP names [notify.Keys]
// declares among them, each required when the source is selected. The two
// optional names carry a rotation-overlap secret each, so a deployment
// can roll a secret without a window where every existing token or
// session fails at once: they are the source's Optional keys rather
// than registered thresholds, because a threshold's KindString default
// must be non-empty and doctor prints a threshold's effective value,
// which would print a secret on an operator's terminal.
//
// Validate is [validateMagicLinkSource], the source's value rules: the
// secret length floor, the public URL's shape, the sender address, and
// the mail port.
var magicLinkIdentitySource = config.IdentitySource{
	Name:     "magic-link",
	Selector: identitySourceSelectorRow.Key,
	Required: append([]config.KeyDecl{
		{Key: MagicLinkSecretKey, Description: "the secret that signs sign-in links, at least 32 bytes", Sensitivity: config.Sensitive},
		{Key: SessionSecretKey, Description: "the secret that signs session cookies, at least 32 bytes", Sensitivity: config.Sensitive},
		{Key: publicURLKey, Description: "the address members reach the server at, used in sign-in links", Sensitivity: config.Sensitive},
	}, notify.Keys()...),
	Optional: []config.KeyDecl{
		{Key: MagicLinkSecretPreviousKey, Description: "the previous link-signing secret, kept so links minted before a rotation still verify", Sensitivity: config.Sensitive},
		{Key: SessionSecretPreviousKey, Description: "the previous session-signing secret, kept so sessions minted before a rotation still verify", Sensitivity: config.Sensitive},
	},
	Validate: validateMagicLinkSource,
}

// secretLengthFloor is the minimum byte length of a magic-link or
// session secret: RFC 2104 section 3 recommends an HMAC key at least as
// long as the hash's output, 32 bytes for SHA-256. It counts raw bytes,
// so a base64 rendering of 32 random bytes passes with room to spare.
const secretLengthFloor = 32

// MagicLinkExpiryCeiling bounds how long a mailed link can stay
// redeemable: generous headroom above the 15-minute default, short of a
// link surviving a day in a mail scanner's cache or a browser's history.
// It is [magicLinkExpirySetting]'s registered ceiling.
const MagicLinkExpiryCeiling = time.Hour

// magicLinkExpirySetting, sessionLifetimeSetting,
// SessionAbsoluteCapSetting, and nonceRetentionSetting are the four
// auth durations.
//
// Every duration takes a positive floor, since zero or a negative
// duration cannot mean "never" or "immediately" here. The
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
	magicLinkExpirySetting = config.Setting{
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
	sessionLifetimeSetting = config.Setting{
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
	nonceRetentionSetting = config.Setting{
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

// The four rate limits each take a floor of one, since zero is
// ambiguous between "unlimited" and "blocked outright", and a ceiling
// of ten times the default: headroom for a larger group without a
// limit so high it stops bounding anything an enumerator does. The
// limiter is built once, so each applies at the next start.
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

// authRequestBodyCapSetting bounds an auth request's body. Its floor,
// 512 bytes, fits an RFC 5321 address (254 bytes at most) plus its
// form-encoded field name; its ceiling, 64KiB, still keeps the
// amplification bound meaningful, because a cap large enough to admit
// an arbitrary upload defeats the reason the cap exists.
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
		Keys:  []string{magicLinkExpirySetting.Key, nonceRetentionSetting.Key},
		Check: durationAtLeast(nonceRetentionSetting.Key, magicLinkExpirySetting.Key),
	}
	sessionLifetimeWithinCap = config.CorrelatedRule{
		Name:  "session lifetime within the absolute cap",
		Keys:  []string{sessionLifetimeSetting.Key, SessionAbsoluteCapSetting.Key},
		Check: durationAtLeast(SessionAbsoluteCapSetting.Key, sessionLifetimeSetting.Key),
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
// settings they name. It is a plain function on serve.Register's
// precedent, called once from the binary's shared registry constructor.
func Register(reg *config.Registry) {
	reg.AddWallRow(identitySourceSelectorRow)
	reg.AddIdentitySource(magicLinkIdentitySource)
	reg.AddSuperseded(rosterFileKey, rosterFileSupersession)
	for _, s := range []config.Setting{
		magicLinkExpirySetting,
		sessionLifetimeSetting,
		SessionAbsoluteCapSetting,
		nonceRetentionSetting,
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
// URL's shape, and the two SMTP values [notify.ValidateKeys] checks,
// joining every failure it finds so [config.Load] reports a short
// secret and a bad URL together. Every
// returned error names the offending key and never the value or the
// length of a secret, so a refusal narrows no guess about what the
// value was. The auth settings' bounds are the seed rule's, not this
// hook's.
func validateMagicLinkSource(values map[string]string, _ []string) error {
	var errs []error

	for _, key := range secretKeys {
		if v, ok := values[key]; ok && len(v) < secretLengthFloor {
			errs = append(errs, &config.KeyError{Key: key, Reason: fmt.Sprintf("too short, want at least %d bytes", secretLengthFloor)})
		}
	}

	if v, ok := values[publicURLKey]; ok {
		if err := validatePublicURL(v); err != nil {
			errs = append(errs, err)
		}
	}

	if err := notify.ValidateKeys(values); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// validatePublicURL checks value against the conditions a configured
// public origin must meet: it parses; it carries no userinfo, query
// string, fragment, or path other than empty or "/", any of which would
// ride into every link [LinkURL] builds; it has a host, since a
// path-only link would land wherever the mail client chose; and its
// scheme is https unless its host is loopback, the whole loopback set
// the W3C Secure Contexts definition treats as potentially trustworthy.
//
// No returned error quotes value. It is checked for "@" before it is
// parsed, because url.Parse misreads a password containing "/", "?", or
// "#" as host:port and quotes it in its error.
func validatePublicURL(value string) error {
	if strings.Contains(value, "@") {
		return &config.KeyError{Key: publicURLKey, Reason: "carries userinfo, which would become an attacker-chosen prefix of every mailed link"}
	}
	u, err := url.Parse(value)
	if err != nil {
		return &config.KeyError{Key: publicURLKey, Reason: "invalid URL"}
	}
	if u.User != nil {
		return &config.KeyError{Key: publicURLKey, Reason: "carries userinfo, which would become an attacker-chosen prefix of every mailed link"}
	}
	if u.RawQuery != "" {
		return &config.KeyError{Key: publicURLKey, Reason: "carries a query string"}
	}
	if u.Fragment != "" {
		return &config.KeyError{Key: publicURLKey, Reason: "carries a fragment"}
	}
	if u.Path != "" && u.Path != "/" {
		return &config.KeyError{Key: publicURLKey, Reason: "must carry no path; the serve prefix sets where dubplate lives"}
	}
	host := u.Hostname()
	if host == "" {
		return &config.KeyError{Key: publicURLKey, Reason: "has no host"}
	}
	if u.Scheme != "https" && !IsLoopbackHost(host) {
		return &config.KeyError{Key: publicURLKey, Reason: "must use https unless its host is loopback"}
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

// bodyCap returns a reader of the auth request body cap in force
// through acc at each call.
func bodyCap(acc *settings.Accessor) settings.Reader[int64] {
	return func(ctx context.Context) (int64, error) {
		n, err := acc.Int(ctx, authRequestBodyCapSetting)
		return int64(n), err
	}
}

// genericAnswer returns a reader reporting at each call whether the
// unknown-address answer in force through acc is generic.
func genericAnswer(acc *settings.Accessor) settings.Reader[bool] {
	return func(ctx context.Context) (bool, error) {
		v, err := acc.String(ctx, unknownAddressAnswerSetting)
		return v == "generic", err
	}
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

// errSourceNotSelected is the refusal [ConfigFrom] returns for a config
// whose selected identity source is not the magic-link source.
var errSourceNotSelected = errors.New("the magic-link identity source is not selected")

// Config is everything the magic-link source resolves from the wall and
// the settings store, read once at serve's start. The secrets and the
// public origin are fixed for the process. Each reader reads its
// setting at every call, so a stored change applies to the next request
// that reads it.
type Config struct {
	LinkSecret            []byte
	PreviousLinkSecret    []byte
	SessionSecret         []byte
	PreviousSessionSecret []byte

	// PublicOrigin is the address members reach the server at, the
	// origin every sign-in link is built from.
	PublicOrigin string

	LinkExpiry         settings.Reader[time.Duration]
	SessionLifetime    settings.Reader[time.Duration]
	SessionAbsoluteCap settings.Reader[time.Duration]
	NonceRetention     settings.Reader[time.Duration]
	BodyCap            settings.Reader[int64]

	// GenericAnswer reports whether an unknown address gets the same
	// answer a known one does.
	GenericAnswer settings.Reader[bool]
}

// ConfigFrom resolves the magic-link source's values from cfg and its
// settings through acc. It refuses a cfg that selects another source,
// since that config resolves every secret to empty and an empty HMAC
// key would sign every token.
func ConfigFrom(cfg *config.Config, acc *settings.Accessor) (Config, error) {
	if cfg.IdentitySource != magicLinkIdentitySource.Name {
		return Config{}, errSourceNotSelected
	}
	return Config{
		LinkSecret:            []byte(cfg.IdentitySourceValue(MagicLinkSecretKey)),
		PreviousLinkSecret:    []byte(cfg.IdentitySourceValue(MagicLinkSecretPreviousKey)),
		SessionSecret:         []byte(cfg.IdentitySourceValue(SessionSecretKey)),
		PreviousSessionSecret: []byte(cfg.IdentitySourceValue(SessionSecretPreviousKey)),
		PublicOrigin:          cfg.IdentitySourceValue(publicURLKey),
		LinkExpiry:            acc.DurationReader(magicLinkExpirySetting),
		SessionLifetime:       acc.DurationReader(sessionLifetimeSetting),
		SessionAbsoluteCap:    acc.DurationReader(SessionAbsoluteCapSetting),
		NonceRetention:        acc.DurationReader(nonceRetentionSetting),
		BodyCap:               bodyCap(acc),
		GenericAnswer:         genericAnswer(acc),
	}, nil
}

// errSettingUnreadable marks a failed read of a setting's value in
// force. A handler answers it as a server error, never as a refusal of
// what the request carried, and never falls back to a default.
var errSettingUnreadable = errors.New("auth setting unreadable")

// settingFailureReason is the log reason for err, a failed setting
// read: a fixed phrase naming the key err's [settings.ReadError]
// carries, or "unnamed" when it carries none. A handler logs this and
// never err's text, which may carry whatever the read met.
func settingFailureReason(err error) string {
	key := "unnamed"
	if re, ok := errors.AsType[*settings.ReadError](err); ok {
		key = re.Key
	}
	return "setting unreadable: " + key
}

// readSetting calls s, wrapping a failure in errSettingUnreadable.
func readSetting[T any](ctx context.Context, s settings.Reader[T]) (T, error) {
	v, err := s(ctx)
	if err != nil {
		return v, fmt.Errorf("%w: %w", errSettingUnreadable, err)
	}
	return v, nil
}
