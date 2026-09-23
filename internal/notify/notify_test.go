package notify

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// serverTLSConfig and clientTLSConfig share one self-signed certificate,
// built once in TestMain: the server presents it, the client trusts it
// by CA rather than by disabling verification, so the TLS tests below
// exercise a real handshake end to end.
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

// generateSelfSignedCert builds an ECDSA certificate for 127.0.0.1, valid
// for one hour, so the tests below need no filesystem or network access
// to stand up a TLS SMTP server.
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

// delivery is one message as testServer saw it: the envelope, the raw
// RFC 5322 bytes, and the AUTH credentials presented for it.
type delivery struct {
	from string
	to   []string
	data []byte
	user string
	pass string
}

// testServer is a minimal SMTP server for this package's own tests, and
// never reaches a real network: every listener binds 127.0.0.1 on an
// ephemeral port. upgradeTLS, when set, both advertises STARTTLS in the
// EHLO response and performs the upgrade; left nil, the server never
// offers STARTTLS at all, which is what TestSend_StartTLSNotOffered
// needs.
type testServer struct {
	upgradeTLS *tls.Config

	// rejectRCPT, when set, answers every RCPT TO with a permanent 550
	// rather than accepting it, so TestSend_RCPTRejected can exercise
	// the send-failure path without reaching a real network.
	rejectRCPT bool

	mu         sync.Mutex
	deliveries []delivery
}

// startTestServer starts srv on 127.0.0.1:0 and returns its address.
// implicitTLS wraps the raw listener in TLS before any connection is
// accepted, the port-465 contract; otherwise the listener starts in
// plaintext and only STARTTLS (if srv.upgradeTLS is set) ever encrypts
// it.
func startTestServer(t *testing.T, srv *testServer, implicitTLS bool) string {
	t.Helper()

	rawLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := rawLn.Addr().String()

	ln := net.Listener(rawLn)
	if implicitTLS {
		ln = tls.NewListener(rawLn, serverTLSConfig)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go srv.handleConn(conn)
		}
	}()
	t.Cleanup(func() { _ = rawLn.Close() })

	return addr
}

func (s *testServer) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	writeLine := func(line string) {
		_, _ = fmt.Fprintf(w, "%s\r\n", line)
		_ = w.Flush()
	}
	writeLine("220 localhost ESMTP dubplate-test")

	var from string
	var to []string
	var user, pass string

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			_, _ = fmt.Fprintf(w, "250-localhost\r\n")
			if s.upgradeTLS != nil {
				_, _ = fmt.Fprintf(w, "250-STARTTLS\r\n")
			}
			_, _ = fmt.Fprintf(w, "250 AUTH PLAIN\r\n")
			_ = w.Flush()
		case strings.HasPrefix(upper, "STARTTLS"):
			writeLine("220 ready to start TLS")
			conn = tls.Server(conn, s.upgradeTLS)
			r = bufio.NewReader(conn)
			w = bufio.NewWriter(conn)
		case strings.HasPrefix(upper, "AUTH PLAIN"):
			if payload, ok := strings.CutPrefix(line, "AUTH PLAIN "); ok {
				user, pass = decodePlainAuth(payload)
			}
			writeLine("235 2.7.0 Authentication successful")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			from = extractAddr(line)
			writeLine("250 2.1.0 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			if s.rejectRCPT {
				writeLine("550 5.1.1 mailbox unavailable")
				continue
			}
			to = append(to, extractAddr(line))
			writeLine("250 2.1.5 OK")
		case upper == "DATA":
			writeLine("354 End data with <CR><LF>.<CR><LF>")
			data := readData(r)
			writeLine("250 2.0.0 OK: queued")
			s.mu.Lock()
			s.deliveries = append(s.deliveries, delivery{from: from, to: to, data: data, user: user, pass: pass})
			s.mu.Unlock()
		case upper == "NOOP":
			writeLine("250 2.0.0 OK")
		case upper == "RSET":
			from, to, user, pass = "", nil, "", ""
			writeLine("250 2.0.0 OK")
		case upper == "QUIT":
			writeLine("221 2.0.0 Bye")
			return
		default:
			writeLine("500 5.5.2 command not recognized")
		}
	}
}

// decodePlainAuth decodes an RFC 4616 PLAIN response
// ("authzid\x00authcid\x00passwd", base64) into its username and
// password fields.
func decodePlainAuth(b64 string) (user, pass string) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", ""
	}
	parts := strings.Split(string(raw), "\x00")
	if len(parts) != 3 {
		return "", ""
	}
	return parts[1], parts[2]
}

func extractAddr(line string) string {
	start := strings.Index(line, "<")
	end := strings.Index(line, ">")
	if start == -1 || end == -1 || end < start {
		return ""
	}
	return line[start+1 : end]
}

func readData(r *bufio.Reader) []byte {
	var buf bytes.Buffer
	for {
		line, err := r.ReadString('\n')
		if err != nil || line == ".\r\n" || line == ".\n" {
			return buf.Bytes()
		}
		buf.WriteString(line)
	}
}

func receive(t *testing.T, srv *testServer) delivery {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		srv.mu.Lock()
		n := len(srv.deliveries)
		srv.mu.Unlock()
		if n > 0 {
			srv.mu.Lock()
			d := srv.deliveries[0]
			srv.mu.Unlock()
			return d
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no message delivered within 5s")
	return delivery{}
}

// assertParity checks the received message's From, Date, Message-Id,
// the RFC 2047 encoded-word Subject, and the RFC 2045 section 6.7
// quoted-printable body, each decoded back to the exact value Send was
// given.
func assertParity(t *testing.T, d delivery, cfg Config, to, wantSubject, wantBody string) {
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

	// RFC 2047 section 4.2: a non-ASCII subject must travel as a
	// Q-encoded word, not raw UTF-8. This literal was derived by hand,
	// not by calling the encoder under test: each octet of "Sigur Rós"
	// outside the token charset is replaced by "=" plus its hex value,
	// and each space becomes "_".
	const wantEncodedSubject = "=?UTF-8?q?Sigur_R=C3=B3s:_upload_rejected?="
	if got := msg.Header.Get("Subject"); got != wantEncodedSubject {
		t.Errorf("Subject header = %q, want %q (RFC 2047 section 4.2 Q-encoded word)", got, wantEncodedSubject)
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

	// RFC 2045 section 6.7: quoted-printable transfer encoding. This
	// literal was derived by hand: each non-ASCII octet of "Björk" is
	// replaced by "=" plus its hex value, and the trailing CRLFs are
	// preserved literally since quoted-printable passes CRLF through.
	const wantEncodedBody = "Bj=C3=B6rk's album was rejected.\r\nReason: door & window.\r\n"
	if got := msg.Header.Get("Content-Transfer-Encoding"); got != "quoted-printable" {
		t.Errorf("Content-Transfer-Encoding = %q, want quoted-printable", got)
	}
	rawBody, err := io.ReadAll(msg.Body)
	if err != nil {
		t.Fatalf("read raw body: %v", err)
	}
	if string(rawBody) != wantEncodedBody {
		t.Errorf("raw body = %q, want %q (RFC 2045 section 6.7 quoted-printable)", rawBody, wantEncodedBody)
	}
	bodyBytes, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(rawBody)))
	if err != nil {
		t.Fatalf("decode quoted-printable body: %v", err)
	}
	if string(bodyBytes) != wantBody {
		t.Errorf("body = %q, want %q", bodyBytes, wantBody)
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

func TestSend_ImplicitTLS465(t *testing.T) {
	srv := &testServer{}
	addr := startTestServer(t, srv, true)
	host, port := splitHostPort(t, addr)

	cfg := Config{Host: host, Port: port, User: "smtp-user@example.com", Password: "s3cret-app-password", From: "dubplate@example.com"}
	mailer := New(cfg)
	mailer.tlsConfig = clientTLSConfig
	mailer.forceImplicitTLS = true // stand-in for real port 465; see field doc.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const subject = "Sigur Rós: upload rejected"
	const body = "Björk's album was rejected.\r\nReason: door & window.\r\n"
	if err := mailer.Send(ctx, "contributor@example.com", subject, body); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	assertParity(t, receive(t, srv), cfg, "contributor@example.com", subject, body)
}

func TestSend_StartTLS587(t *testing.T) {
	srv := &testServer{upgradeTLS: serverTLSConfig}
	addr := startTestServer(t, srv, false)
	host, port := splitHostPort(t, addr)

	cfg := Config{Host: host, Port: port, User: "smtp-user@example.com", Password: "s3cret-app-password", From: "dubplate@example.com"}
	mailer := New(cfg)
	mailer.tlsConfig = clientTLSConfig
	// forceImplicitTLS left false: cfg.Port is not 465, so Send takes the
	// same STARTTLS branch the deployed box takes on 587.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const subject = "Sigur Rós: upload rejected"
	const body = "Björk's album was rejected.\r\nReason: door & window.\r\n"
	if err := mailer.Send(ctx, "contributor@example.com", subject, body); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	assertParity(t, receive(t, srv), cfg, "contributor@example.com", subject, body)
}

// TestSend_StartTLSNotOffered proves Send fails closed rather than
// falling back to plaintext when the server offers no STARTTLS at all.
func TestSend_StartTLSNotOffered(t *testing.T) {
	srv := &testServer{} // upgradeTLS nil: EHLO never advertises STARTTLS.
	addr := startTestServer(t, srv, false)
	host, port := splitHostPort(t, addr)

	cfg := Config{Host: host, Port: port, User: "u", Password: "p", From: "dubplate@example.com"}
	mailer := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mailer.Send(ctx, "contributor@example.com", "subject", "body")
	if err == nil {
		t.Fatal("Send() error = nil, want non-nil when the server offers no STARTTLS")
	}
	if !strings.Contains(err.Error(), "does not support STARTTLS") {
		t.Errorf("Send() error = %q, want it to mention the server not supporting STARTTLS", err)
	}
	if !errors.Is(err, ErrSend) {
		t.Errorf("Send() error = %v, want errors.Is(err, ErrSend)", err)
	}
}

func TestSend_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The listener is created and closed before Send runs, so a dial
	// attempt reaching the network would fail loudly (connection
	// refused) rather than silently succeeding.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	host, port := splitHostPort(t, addr)

	mailer := New(Config{Host: host, Port: port, User: "u", Password: "p", From: "dubplate@example.com"})
	if err := mailer.Send(ctx, "contributor@example.com", "subject", "body"); !errors.Is(err, context.Canceled) {
		t.Errorf("Send() error = %v, want context.Canceled", err)
	}
}

func TestSend_HeaderInjectionGuard(t *testing.T) {
	// The listener is created and closed before each Send call, so a
	// guard that fails to fire reaches the network and returns a
	// connection-refused error instead of ErrHeaderInjection.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	host, port := splitHostPort(t, addr)

	tests := []struct {
		name    string
		to      string
		subject string
	}{
		{name: "recipient carries CR", to: "victim@example.com\rBcc: attacker@example.com", subject: "subject"},
		{name: "recipient carries LF", to: "victim@example.com\nBcc: attacker@example.com", subject: "subject"},
		{name: "subject carries CR", to: "contributor@example.com", subject: "subject\rBcc: attacker@example.com"},
		{name: "subject carries LF", to: "contributor@example.com", subject: "subject\nBcc: attacker@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := New(Config{Host: host, Port: port, User: "u", Password: "p", From: "dubplate@example.com"})

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := mailer.Send(ctx, tt.to, tt.subject, "body")
			if !errors.Is(err, ErrHeaderInjection) {
				t.Fatalf("Send() error = %v, want ErrHeaderInjection", err)
			}
		})
	}
}

// TestSend_InvalidRecipientOmitsAddress proves an unparseable recipient
// fails with ErrInvalidRecipient and that the address itself never
// appears in the error text, unlike go-mail's own parse error which
// quotes it.
func TestSend_InvalidRecipientOmitsAddress(t *testing.T) {
	mailer := New(Config{Host: "127.0.0.1", Port: 1, User: "u", Password: "p", From: "dubplate@example.com"})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const badRecipient = "not-an-address"
	err := mailer.Send(ctx, badRecipient, "subject", "body")
	if !errors.Is(err, ErrInvalidRecipient) {
		t.Fatalf("Send() error = %v, want errors.Is(err, ErrInvalidRecipient)", err)
	}
	if strings.Contains(err.Error(), badRecipient) {
		t.Errorf("Send() error = %q, must not contain the recipient literal %q", err, badRecipient)
	}
}

// TestSend_RCPTRejected proves a mid-transaction SMTP rejection wraps
// ErrSend and that the rejected recipient never appears in the error
// text, unlike go-mail's own *SendError.Error, which appends "affected
// recipient(s): <address>".
func TestSend_RCPTRejected(t *testing.T) {
	srv := &testServer{upgradeTLS: serverTLSConfig, rejectRCPT: true}
	addr := startTestServer(t, srv, false)
	host, port := splitHostPort(t, addr)

	mailer := New(Config{Host: host, Port: port, User: "u", Password: "p", From: "dubplate@example.com"})
	mailer.tlsConfig = clientTLSConfig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const recipient = "contributor@example.com"
	err := mailer.Send(ctx, recipient, "subject", "body")
	if !errors.Is(err, ErrSend) {
		t.Fatalf("Send() error = %v, want errors.Is(err, ErrSend)", err)
	}
	if strings.Contains(err.Error(), recipient) {
		t.Errorf("Send() error = %q, must not contain the recipient literal %q", err, recipient)
	}
}

// TestSend_SuccessRecordOmitsRecipient captures the default slog output
// on a successful send and scans it for the recipient's own literal,
// which must not appear: Global Constraints reserve an address-carrying
// record to the roster audit log.
func TestSend_SuccessRecordOmitsRecipient(t *testing.T) {
	var logBuf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	srv := &testServer{upgradeTLS: serverTLSConfig}
	addr := startTestServer(t, srv, false)
	host, port := splitHostPort(t, addr)

	const recipient = "contributor@example.com"
	mailer := New(Config{Host: host, Port: port, User: "u", Password: "p", From: "dubplate@example.com"})
	mailer.tlsConfig = clientTLSConfig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mailer.Send(ctx, recipient, "subject", "body"); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	if strings.Contains(logBuf.String(), recipient) {
		t.Errorf("log output %q names the recipient, want no recipient logged", logBuf.String())
	}
	if !strings.Contains(logBuf.String(), "notify sent") {
		t.Errorf("log output %q missing the success record", logBuf.String())
	}
}
