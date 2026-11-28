// Package notify sends plain-text mail over SMTP.
package notify

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	netmail "net/mail"
	"strings"

	"github.com/wneessen/go-mail"
)

// implicitTLSPort is the port at which Send dials with TLS from the
// first byte instead of upgrading via STARTTLS.
var implicitTLSPort = 465

// tlsConfig, when non-nil, replaces the client's TLS configuration. Nil
// keeps go-mail's default, which verifies the server against the system
// trust store.
var tlsConfig *tls.Config

// errHeaderInjection reports that a To address or a subject carried a
// CR or LF, either of which could inject an additional header into the
// composed message.
var errHeaderInjection = errors.New("header value contains CR or LF")

// errInvalidFrom reports that the configured From address is not one
// RFC 5322 mailbox. The error text never repeats the address: net/mail's
// own parse error quotes it verbatim.
var errInvalidFrom = errors.New("invalid from address")

// errInvalidTo reports that the To address does not parse as an RFC 5322
// mailbox. The error text never repeats the address, for the same reason
// as errInvalidFrom.
var errInvalidTo = errors.New("invalid to address")

// errSend reports that connecting to or talking with the SMTP server
// failed. The error text carries the host, port, and go-mail's failure
// reason and code, but never the recipient list: go-mail's own
// *SendError.Error appends "affected recipient(s): <address>". A config
// go-mail refuses to build a client from fails before any connection
// and is not errSend.
var errSend = errors.New("send failed")

// smtpConfig holds the five SMTP values a Mailer needs. Send reads no
// environment variable of its own.
type smtpConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Mailer sends mail over one SMTP transport.
type Mailer struct {
	cfg  smtpConfig
	from *netmail.Address
}

func newMailer(cfg smtpConfig) (*Mailer, error) {
	from, err := parseFrom(cfg.From)
	if err != nil {
		return nil, err
	}
	return &Mailer{cfg: cfg, from: from}, nil
}

// parseFrom parses value as one RFC 5322 mailbox. It also refuses a CR
// or LF, which net/mail accepts inside a comment.
func parseFrom(value string) (*netmail.Address, error) {
	addr, err := netmail.ParseAddress(value)
	if err != nil || strings.ContainsAny(value, "\r\n") {
		return nil, errInvalidFrom
	}
	return addr, nil
}

// Send delivers one email to to with subject and body. It refuses to
// dial when ctx is already done, and refuses to send when to or subject
// carries a CR or LF, since either could inject an additional header
// into the message it is about to compose. Subject and body may hold
// non-ASCII text (contributor names, album titles): go-mail encodes the
// subject per RFC 2047 and the body as quoted-printable per RFC 2045
// section 6.7. Send reports failure to its caller and makes no retry or
// best-effort policy of its own; that decision belongs to each caller.
func (m *Mailer) Send(ctx context.Context, to, subject, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := guardHeaderValue("to", to); err != nil {
		return err
	}
	if err := guardHeaderValue("subject", subject); err != nil {
		return err
	}

	msg := mail.NewMsg()
	msg.FromMailAddress(m.from)
	if err := msg.To(to); err != nil {
		return errInvalidTo
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)
	msg.SetDate()
	msg.SetMessageID()

	opts := []mail.Option{
		mail.WithPort(m.cfg.Port),
		// AutoDiscover picks the strongest mechanism the server offers,
		// so a LOGIN-only server still authenticates.
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(m.cfg.User),
		mail.WithPassword(m.cfg.Password),
	}
	if m.cfg.Port == implicitTLSPort {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}
	if tlsConfig != nil {
		opts = append(opts, mail.WithTLSConfig(tlsConfig))
	}

	client, err := mail.NewClient(m.cfg.Host, opts...)
	if err != nil {
		return fmt.Errorf("smtp client %s:%d: %w", m.cfg.Host, m.cfg.Port, err)
	}

	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		if sendErr, ok := errors.AsType[*mail.SendError](err); ok {
			return fmt.Errorf("%w: %s:%d: %s (code %d)", errSend, m.cfg.Host, m.cfg.Port, sendErr.Reason, sendErr.ErrorCode())
		}
		return fmt.Errorf("%w: %s:%d: %w", errSend, m.cfg.Host, m.cfg.Port, err)
	}
	return nil
}

func guardHeaderValue(field, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s: %w", field, errHeaderInjection)
	}
	return nil
}
