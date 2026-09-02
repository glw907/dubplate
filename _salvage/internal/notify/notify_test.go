package notify

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/mhale/smtpd"
)

// serverTLSConfig and clientTLSConfig share one self-signed certificate,
// built once in TestMain: the server presents it, the client trusts it
// by CA rather than by disabling verification, so the parity tests below
// exercise a real TLS handshake end to end.
var (
	serverTLSConfig *tls.Config
	clientTLSConfig *tls.Config
)

func TestMain(m *testing.M) {
	cert, leaf, err := generateSelfSignedCert()
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate test TLS cert:", err)
		os.Exit(1)
	}
	serverTLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	clientTLSConfig = &tls.Config{RootCAs: pool, ServerName: "127.0.0.1"}
	os.Exit(m.Run())
}

// generateSelfSignedCert builds an ECDSA certificate for 127.0.0.1,
// valid for one hour, so the parity tests need no filesystem or network
// access to stand up a TLS SMTP server.
func generateSelfSignedCert() (tls.Certificate, *x509.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("generate key: %w", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("create certificate: %w", err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("parse certificate: %w", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("marshal key: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("load key pair: %w", err)
	}
	return cert, leaf, nil
}

// delivery is one message as the test SMTP server saw it: the envelope,
// the raw RFC 5322 bytes, and the AUTH credentials presented for it.
type delivery struct {
	from string
	to   []string
	data []byte
	user string
	pass string
}

// startServer starts a local SMTP test server and returns its address
// plus a channel that receives each delivered message. implicitTLS
// wraps the listener in TLS before smtpd ever sees a connection (the
// port-465 contract); otherwise the server only offers STARTTLS.
func startServer(t *testing.T, implicitTLS bool) (addr string, deliveries chan delivery) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr = ln.Addr().String()

	serveLn := net.Listener(ln)
	if implicitTLS {
		serveLn = tls.NewListener(ln, serverTLSConfig)
	}

	deliveries = make(chan delivery, 1)
	var authUser, authPass []byte
	srv := &smtpd.Server{
		Hostname:     "localhost",
		Appname:      "dubplate-test",
		TLSConfig:    serverTLSConfig,
		AuthRequired: true,
		AuthHandler: func(_ net.Addr, _ string, username, password, _ []byte) (bool, error) {
			authUser, authPass = username, password
			return true, nil
		},
		Handler: func(_ net.Addr, from string, to []string, data []byte) error {
			deliveries <- delivery{from: from, to: to, data: data, user: string(authUser), pass: string(authPass)}
			return nil
		},
	}

	go func() { _ = srv.Serve(serveLn) }()
	t.Cleanup(func() { _ = serveLn.Close() })

	return addr, deliveries
}

// receive waits for one delivery, failing the test if none arrives
// within a few seconds (a hung dial or a protocol mismatch, not a slow
// disk).
func receive(t *testing.T, deliveries chan delivery) delivery {
	t.Helper()
	select {
	case d := <-deliveries:
		return d
	case <-time.After(5 * time.Second):
		t.Fatal("no message delivered within 5s")
		return delivery{}
	}
}

// assertParity checks the received message against the msmtp-parity
// criteria: synthesized From/Date/Message-ID, an RFC 2047 subject, and a
// quoted-printable body -- decoded back to the exact bytes Send was
// given, not merely "present".
func assertParity(t *testing.T, d delivery, cfg config.SMTPConfig, to, wantSubject, wantBody string) {
	t.Helper()

	if d.user != cfg.User {
		t.Errorf("AUTH user = %q, want %q", d.user, cfg.User)
	}
	if d.pass != cfg.Password {
		t.Errorf("AUTH password = %q, want %q", d.pass, cfg.Password)
	}
	if len(d.to) != 1 || d.to[0] != to {
		t.Errorf("RCPT TO = %v, want [%s]", d.to, to)
	}

	msg, err := mail.ReadMessage(bytes.NewReader(d.data))
	if err != nil {
		t.Fatalf("parse received message: %v", err)
	}

	fromAddr, err := mail.ParseAddress(msg.Header.Get("From"))
	if err != nil {
		t.Fatalf("parse From header %q: %v", msg.Header.Get("From"), err)
	}
	if fromAddr.Address != cfg.From {
		t.Errorf("From = %q, want %q", fromAddr.Address, cfg.From)
	}

	if _, err := msg.Header.Date(); err != nil {
		t.Errorf("Date header missing or unparseable: %v", err)
	}

	if msg.Header.Get("Message-Id") == "" {
		t.Error("Message-Id header missing")
	}

	subject, err := (&mime.WordDecoder{}).DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		t.Fatalf("decode Subject header %q: %v", msg.Header.Get("Subject"), err)
	}
	if subject != wantSubject {
		t.Errorf("Subject = %q, want %q", subject, wantSubject)
	}
	if msg.Header.Get("Subject") == wantSubject {
		t.Error("Subject header sent as raw UTF-8, want RFC 2047 encoded-word form")
	}

	if got := msg.Header.Get("Content-Transfer-Encoding"); got != "quoted-printable" {
		t.Errorf("Content-Transfer-Encoding = %q, want quoted-printable", got)
	}
	bodyBytes, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
	if err != nil {
		t.Fatalf("decode quoted-printable body: %v", err)
	}
	if string(bodyBytes) != wantBody {
		t.Errorf("body = %q, want %q", bodyBytes, wantBody)
	}
}

func TestSend_ImplicitTLS465(t *testing.T) {
	addr, deliveries := startServer(t, true)
	host, port := splitHostPort(t, addr)

	cfg := config.SMTPConfig{
		Host:     host,
		Port:     port,
		User:     "geoff-login@907.life",
		Password: "s3cret-app-password",
		From:     "dubplate@example.com",
	}
	mailer := New(cfg)
	mailer.tlsConfig = clientTLSConfig
	mailer.forceImplicitTLS = true // stand-in for real port 465; see field doc

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const subject = "Sigur Rós: upload rejected"
	const body = "Björk's album was rejected.\r\nReason: door & window.\r\n"
	if err := mailer.Send(ctx, "contributor@example.com", subject, body); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	assertParity(t, receive(t, deliveries), cfg, "contributor@example.com", subject, body)
}

func TestSend_StartTLS587(t *testing.T) {
	addr, deliveries := startServer(t, false)
	host, port := splitHostPort(t, addr)

	cfg := config.SMTPConfig{
		Host:     host,
		Port:     port,
		User:     "geoff-login@907.life",
		Password: "s3cret-app-password",
		From:     "dubplate@example.com",
	}
	mailer := New(cfg)
	mailer.tlsConfig = clientTLSConfig
	// forceImplicitTLS left false: cfg.Port is not 465, so Send takes
	// the same STARTTLS branch the deployed box takes on 587.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const subject = "Sigur Rós: upload rejected"
	const body = "Björk's album was rejected.\r\nReason: door & window.\r\n"
	if err := mailer.Send(ctx, "contributor@example.com", subject, body); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	assertParity(t, receive(t, deliveries), cfg, "contributor@example.com", subject, body)
}

// TestSend_StartTLSRequiredNotOffered proves the STARTTLS branch fails
// closed rather than falling back to plaintext when the server offers no
// TLS at all -- the msmtp contract never sends credentials in the clear.
func TestSend_StartTLSRequiredNotOffered(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	srv := &smtpd.Server{
		Hostname: "localhost",
		Appname:  "dubplate-test",
		// No TLSConfig: the server never advertises STARTTLS.
		Handler: func(net.Addr, string, []string, []byte) error {
			t.Error("Handler called; message should never have been accepted over plaintext")
			return nil
		},
	}
	go func() { _ = srv.Serve(ln) }()

	host, port := splitHostPort(t, ln.Addr().String())
	cfg := config.SMTPConfig{Host: host, Port: port, User: "u", Password: "p", From: "dubplate@example.com"}
	mailer := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = mailer.Send(ctx, "contributor@example.com", "subject", "body")
	if err == nil {
		t.Fatal("Send() error = nil, want non-nil when the server offers no STARTTLS")
	}
	// Assert on the failure reason, not merely its presence: a plaintext-
	// fallback regression also returns a non-nil error (AUTH fails because
	// this server has no AuthHandler), which would pass a bare err != nil
	// check. Pinning the message to the STARTTLS refusal keeps the test
	// red under that regression too.
	if !strings.Contains(err.Error(), "does not support STARTTLS") {
		t.Errorf("Send() error = %q, want it to mention the server not supporting STARTTLS", err)
	}
}

func splitHostPort(t *testing.T, addr string) (host string, port int) {
	t.Helper()
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host:port %q: %v", addr, err)
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		t.Fatalf("parse port %q: %v", p, err)
	}
	return h, n
}
