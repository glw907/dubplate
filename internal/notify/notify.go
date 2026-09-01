// Package notify sends dubplate's contributor-facing emails -- review
// rejections today, any future run notification later -- over SMTP. It
// replaces the shell layer's msmtp invocation with a native TLS client
// that dials the server directly instead of shelling out to sendmail(1),
// and honors the same msmtprc TLS contract: implicit TLS on port 465
// ("tls on", "tls_starttls off"), STARTTLS on any other port.
package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/glw907/dubplate/internal/config"
	"github.com/wneessen/go-mail"
)

// implicitTLSPort is msmtp's "tls on, tls_starttls off" port: the
// connection is TLS from the first byte, no STARTTLS upgrade. Any other
// port gets mandatory STARTTLS -- the box's actual route as of
// 2026-08-31, since Hetzner blocks outbound 465 by default and the
// deployed msmtprc uses 587 with tls_starttls on.
const implicitTLSPort = 465

// Mailer sends notification emails over SMTP.
type Mailer struct {
	cfg config.SMTPConfig

	// tlsConfig overrides the client's TLS configuration when set. New
	// never sets it; this package's own tests do, to trust a self-signed
	// local test server's certificate without weakening the default
	// (system-trust) behavior every real Mailer uses.
	tlsConfig *tls.Config

	// forceImplicitTLS routes Send down the implicit-TLS branch
	// regardless of cfg.Port. New never sets it. It exists only so this
	// package's own parity test can exercise that branch against an
	// ephemeral test-server port: binding the real port 465 needs root,
	// which a normal `go test` run does not have.
	forceImplicitTLS bool
}

// New returns a Mailer that delivers through the SMTP server described
// by cfg.
func New(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

// Send delivers one email to to with subject and body, synthesizing the
// From, Date, and Message-ID headers. Subject and body may contain
// non-ASCII text (contributor names, album titles): the subject is
// RFC 2047 encoded and the body is quoted-printable, both go-mail
// defaults. It logs one slog record on success; the caller decides
// whether a failure aborts anything (notify is best-effort margin, not a
// per-run failure, per the spec's error doctrine).
func (m *Mailer) Send(ctx context.Context, to, subject, body string) error {
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
		return fmt.Errorf("send to %s via %s:%d: %w", to, m.cfg.Host, m.cfg.Port, err)
	}

	slog.Info("notify sent", "to", to, "subject", subject, "host", m.cfg.Host, "port", m.cfg.Port)
	return nil
}
