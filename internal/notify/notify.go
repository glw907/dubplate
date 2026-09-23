// Package notify sends dubplate's contributor-facing mail over SMTP. It
// is a general send seam, born general rather than as a magic-link
// mailer: rung 7's reject copy and rung 9's notifications are its other
// callers, and it carries a subject and a body as data so it never
// duplicates internal/reject's wording.
package notify

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wneessen/go-mail"
)

// implicitTLSPort is the port at which Send dials with TLS from the
// first byte instead of upgrading via STARTTLS. Hetzner blocks outbound
// 465 by default, so the box's own route is STARTTLS on 587; this branch
// exists for any provider that still reserves 465 for implicit TLS.
const implicitTLSPort = 465

// ErrHeaderInjection reports that a recipient, subject, or From address
// carried a CR or LF, either of which could inject an additional header
// into the composed message.
var ErrHeaderInjection = errors.New("notify: header value contains CR or LF")

// Config names the five SMTP values a Mailer needs. The caller resolves
// them; Send reads no environment variable of its own.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Mailer sends mail over one SMTP transport.
type Mailer struct {
	cfg Config

	// tlsConfig overrides the client's trust store. Test-only: New never
	// sets it, so every real Mailer verifies the server certificate
	// against the system trust store. This package's own tests set it to
	// trust a self-signed local test server.
	tlsConfig *tls.Config

	// forceImplicitTLS routes Send down the implicit-TLS branch
	// regardless of cfg.Port. Test-only: New never sets it. Binding the
	// real port 465 needs root, which `go test` does not have, so the
	// implicit-TLS test exercises this field against an ephemeral port
	// instead.
	forceImplicitTLS bool
}

// New returns a Mailer that delivers through the SMTP server cfg
// describes.
func New(cfg Config) *Mailer {
	return &Mailer{cfg: cfg}
}

// Send delivers one email to to with subject and body. It refuses to
// dial when ctx is already done, and refuses to send when to, subject,
// or the configured From address carries a CR or LF, since any of them
// could inject an additional header into the message it is about to
// compose. Subject and body may hold non-ASCII text (contributor
// names, album titles): go-mail encodes the subject per RFC 2047 and the
// body as quoted-printable per RFC 2045 section 6.7. Send reports
// failure to its caller and makes no retry or best-effort policy of its
// own; that decision belongs to each caller.
func (m *Mailer) Send(ctx context.Context, to, subject, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := guardHeaderValue("recipient", to); err != nil {
		return err
	}
	if err := guardHeaderValue("subject", subject); err != nil {
		return err
	}
	if err := guardHeaderValue("from", m.cfg.From); err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From(m.cfg.From); err != nil {
		return fmt.Errorf("set from %s: %w", m.cfg.From, err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("set to %s: %w", to, err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)
	msg.SetDate()
	msg.SetMessageID()

	opts := []mail.Option{
		mail.WithPort(m.cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.cfg.User),
		mail.WithPassword(m.cfg.Password),
	}
	if m.cfg.Port == implicitTLSPort || m.forceImplicitTLS {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}
	if m.tlsConfig != nil {
		opts = append(opts, mail.WithTLSConfig(m.tlsConfig))
	}

	client, err := mail.NewClient(m.cfg.Host, opts...)
	if err != nil {
		return fmt.Errorf("smtp client %s:%d: %w", m.cfg.Host, m.cfg.Port, err)
	}

	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send via %s:%d: %w", m.cfg.Host, m.cfg.Port, err)
	}

	// The success record names no recipient: Global Constraints reserve
	// an address-carrying record to the roster audit log, and logging one
	// here would be a second, uncontrolled copy of the same fact.
	slog.Info("notify sent", "host", m.cfg.Host, "port", m.cfg.Port)
	return nil
}

func guardHeaderValue(field, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s: %w", field, ErrHeaderInjection)
	}
	return nil
}
