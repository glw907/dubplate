package notify

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/config"
)

// The five SMTP key names. None carries a secret value; each names the
// environment variable a real deployment carries one under.
const (
	hostKey     = "DUBPLATE_SMTP_HOST"
	portKey     = "DUBPLATE_SMTP_PORT"
	userKey     = "DUBPLATE_SMTP_USER"
	passwordKey = "DUBPLATE_SMTP_PASSWORD" // secret-guard-allow
	fromKey     = "DUBPLATE_SMTP_FROM"
)

// Keys declares the five SMTP keys for an identity source that mails
// its members, each one required while the source is selected. The
// magic-link source requires them, so they resolve through
// [config.Config.IdentitySourceValue].
func Keys() []config.KeyDecl {
	return []config.KeyDecl{
		{Key: hostKey, Description: "the mail server the server sends sign-in links through", Sensitivity: config.Sensitive},
		{Key: portKey, Description: "the port of that mail server", Sensitivity: config.Sensitive},
		{Key: userKey, Description: "the user name the server signs in to the mail server with", Sensitivity: config.Sensitive},
		{Key: passwordKey, Description: "the password the server signs in to the mail server with", Sensitivity: config.Sensitive},
		{Key: fromKey, Description: "the sender address on the mail the server sends", Sensitivity: config.Sensitive},
	}
}

// ValidateKeys checks the two SMTP values with a grammar of their own,
// the sender address and then the port, among an identity source's
// resolved values, and joins a [config.KeyError] for each that fails. A
// key absent from values is left to [config.Load]'s requirement check.
// It runs from the source's Validate hook, at load.
func ValidateKeys(values map[string]string) error {
	var errs []error
	if v, ok := values[fromKey]; ok {
		if err := validateFrom(v); err != nil {
			errs = append(errs, err)
		}
	}
	if v, ok := values[portKey]; ok {
		if err := validatePort(v); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// validatePort checks value parses as an integer from 1 through 65535.
// A source key carries no registered bound, so the bound lives here.
func validatePort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return &config.KeyError{Key: portKey, Reason: "not an integer"}
	}
	if port < 1 || port > 65535 {
		return &config.KeyError{Key: portKey, Reason: fmt.Sprintf("%d is not in 1 through 65535", port)}
	}
	return nil
}

// validateFrom checks that value parses as a single RFC 5322 address,
// display name allowed, with no carriage return or line feed to inject
// a mail header. Checked at load, a typo fails doctor rather than the
// first login's send. The returned error never repeats value.
func validateFrom(value string) error {
	if _, err := mail.ParseAddress(value); err != nil || strings.ContainsAny(value, "\r\n") {
		return &config.KeyError{Key: fromKey, Reason: "invalid email address"}
	}
	return nil
}

// ConfigFrom resolves the SMTP values of cfg's selected identity source
// into a [Config]. [ValidateKeys] has already bounded the port at load,
// so the parse here only turns the string into the int Config takes.
func ConfigFrom(cfg *config.Config) (Config, error) {
	port, err := strconv.Atoi(cfg.IdentitySourceValue(portKey))
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", portKey, err)
	}
	return Config{
		Host:     cfg.IdentitySourceValue(hostKey),
		Port:     port,
		User:     cfg.IdentitySourceValue(userKey),
		Password: cfg.IdentitySourceValue(passwordKey),
		From:     cfg.IdentitySourceValue(fromKey),
	}, nil
}
