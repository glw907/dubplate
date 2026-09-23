package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// sessionBirth, sessionLifetimeFixture, and sessionCapFixture pin the
// whole of what the literals below depend on. The two durations are of
// the shape the deployed thresholds carry, a lifetime far shorter than
// the cap, which is what makes a refreshed session able to outlive its
// own expiry and not its anchor.
var (
	sessionBirth           = time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	sessionLifetimeFixture = 30 * 24 * time.Hour
	sessionCapFixture      = 180 * 24 * time.Hour
)

// fixtureSession is the session every literal below was computed from.
// Its address is fixed rather than a parameter: the address-shape cases
// the wire format has to survive are the link's golden vectors, and
// repeating them here would freeze the same property twice.
func fixtureSession() Session {
	return Session{
		Address:    "alice@example.test",
		IssuedAt:   sessionBirth,
		AnchoredAt: sessionBirth,
		ExpiresAt:  sessionBirth.Add(sessionLifetimeFixture),
	}
}

// TestSessionCookieName writes R-5b9's name out rather than reading it
// from the constant under test. The __Host- prefix is the half a
// browser enforces: it refuses to store a cookie so named unless the
// cookie is Secure, carries no Domain, and has Path=/, so the name and
// the attributes asserted below are one rule, not two.
func TestSessionCookieName(t *testing.T) {
	if SessionCookieName != "__Host-dubplate_session" {
		t.Errorf("SessionCookieName = %q, want %q", SessionCookieName, "__Host-dubplate_session")
	}
}

// TestSessionCookieAttributes asserts every attribute on a written
// cookie. Domain is the one asserted empty: a __Host- cookie carrying
// a Domain is refused outright by the browser, which is also why
// same-origin is a stated precondition of the SSO capability rather
// than something the server can relax later.
func TestSessionCookieAttributes(t *testing.T) {
	const wireSample = "ZHVicGxhdGUtc2Vzc2lvbi0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc3MjYwMDc2Nw.MTc3MjYwMDc2Nw.MTc3NTE5Mjc2Nw.gTOGtfHaf6_k8TXPiRrX2vk5pWlAUyg0K65h2uOLpMw"
	now := sessionBirth
	expires := sessionBirth.Add(sessionLifetimeFixture)

	c := SessionCookie(wireSample, now, expires)

	if c.Name != "__Host-dubplate_session" {
		t.Errorf("Name = %q, want %q", c.Name, "__Host-dubplate_session")
	}
	if c.Value != wireSample {
		t.Errorf("Value = %q, want the value it was built from", c.Value)
	}
	if !c.Secure {
		t.Error("Secure = false, want true")
	}
	if !c.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want %v", c.SameSite, http.SameSiteLaxMode)
	}
	if c.Path != "/" {
		t.Errorf("Path = %q, want %q", c.Path, "/")
	}
	if c.Domain != "" {
		t.Errorf("Domain = %q, want it empty: a __Host- cookie carrying a Domain is refused by the browser", c.Domain)
	}
	if !c.Expires.Equal(expires) {
		t.Errorf("Expires = %v, want %v", c.Expires, expires)
	}
	if want := int(sessionLifetimeFixture.Seconds()); c.MaxAge != want {
		t.Errorf("MaxAge = %d, want %d", c.MaxAge, want)
	}
}

// TestSessionCookieRenderedHeader asserts the exact Set-Cookie header
// SessionCookie produces, Max-Age included, against a written-out
// expectation rather than by reading the struct's fields back (review
// m1, D5 acceptance).
func TestSessionCookieRenderedHeader(t *testing.T) {
	const wireSample = "ZHVicGxhdGUtc2Vzc2lvbi0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc3MjYwMDc2Nw.MTc3MjYwMDc2Nw.MTc3NTE5Mjc2Nw.gTOGtfHaf6_k8TXPiRrX2vk5pWlAUyg0K65h2uOLpMw"
	now := sessionBirth
	expires := sessionBirth.Add(sessionLifetimeFixture)

	got := SessionCookie(wireSample, now, expires).String()
	want := "__Host-dubplate_session=" + wireSample +
		"; Path=/; Expires=" + expires.Format(http.TimeFormat) +
		"; Max-Age=2592000; HttpOnly; Secure; SameSite=Lax"
	if got != want {
		t.Errorf("Cookie.String() = %q, want %q", got, want)
	}
}

// TestClearSessionCookie proves the clearing cookie carries the same
// identity-bearing attributes a __Host- cookie needs to be recognized
// as the one to delete, with MaxAge negative and an empty value
// (review m9).
func TestClearSessionCookie(t *testing.T) {
	c := ClearSessionCookie()

	if c.Name != SessionCookieName {
		t.Errorf("Name = %q, want %q", c.Name, SessionCookieName)
	}
	if c.Value != "" {
		t.Errorf("Value = %q, want empty", c.Value)
	}
	if c.Path != "/" {
		t.Errorf("Path = %q, want %q", c.Path, "/")
	}
	if c.Domain != "" {
		t.Errorf("Domain = %q, want it empty", c.Domain)
	}
	if !c.Secure {
		t.Error("Secure = false, want true")
	}
	if !c.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want %v", c.SameSite, http.SameSiteLaxMode)
	}
	if c.MaxAge >= 0 {
		t.Errorf("MaxAge = %d, want negative", c.MaxAge)
	}
	if !c.Expires.Before(sessionBirth) {
		t.Errorf("Expires = %v, want a moment before every fixture instant", c.Expires)
	}
}

// TestSessionGoldenVector freezes the wire format against a literal
// derived from the format's written description (length-prefixed
// fields under both labels, each field base64url-encoded before the
// join) rather than recomputed with the code under test.
func TestSessionGoldenVector(t *testing.T) {
	const wantEncoded = "ZHVicGxhdGUtc2Vzc2lvbi0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc3MjYwMDc2Nw.MTc3MjYwMDc2Nw.MTc3NTE5Mjc2Nw.gTOGtfHaf6_k8TXPiRrX2vk5pWlAUyg0K65h2uOLpMw"

	if got := encodeSession([]byte(currentFixture), fixtureSession()); got != wantEncoded {
		t.Errorf("encodeSession() = %q, want %q", got, wantEncoded)
	}
}

func TestVerifySession(t *testing.T) {
	valid := encodeSession([]byte(currentFixture), fixtureSession())
	underPrevious := encodeSession([]byte(previousFixture), fixtureSession())
	inside := sessionBirth.Add(time.Hour)
	pastExpiry := sessionBirth.Add(sessionLifetimeFixture)

	tests := []struct {
		name  string
		value string
		now   time.Time
		want  error
	}{
		{
			name:  "a one-bit flip in the MAC",
			value: flipFieldBit(t, valid, sessionFieldCount),
			now:   inside,
			want:  ErrSessionMAC,
		},
		{
			name:  "a one-bit flip in the address",
			value: flipFieldBit(t, valid, 1),
			now:   inside,
			want:  ErrSessionMAC,
		},
		{
			name:  "a one-bit flip in the anchor",
			value: flipFieldBit(t, valid, 3),
			now:   inside,
			want:  ErrSessionMAC,
		},
		{
			// The anchor is inside the MAC, so a holder cannot
			// move it forward to buy a session past the absolute
			// cap the anchor is there to enforce.
			name:  "an anchor moved forward without the MAC",
			value: replaceField(valid, 3, []byte("1788325567")),
			now:   inside,
			want:  ErrSessionMAC,
		},
		{
			name:  "past its expiry and inside its cap",
			value: valid,
			now:   pastExpiry,
			want:  ErrSessionExpired,
		},
		{
			name:  "exactly at its expiry",
			value: valid,
			now:   sessionBirth.Add(sessionLifetimeFixture),
			want:  ErrSessionExpired,
		},
		{
			// The MAC is checked first, so a value that fails both
			// reports the MAC. Reversing the order would act on an
			// expiry read off an unauthenticated field.
			name:  "expired and MAC broken reports the MAC",
			value: flipFieldBit(t, valid, sessionFieldCount),
			now:   pastExpiry,
			want:  ErrSessionMAC,
		},
		{
			name:  "too few fields",
			value: strings.Join(strings.Split(valid, ".")[:sessionFieldCount], "."),
			now:   inside,
			want:  ErrSessionMalformed,
		},
		{
			name:  "a field that is not base64url",
			value: "!!!." + strings.SplitN(valid, ".", 2)[1],
			now:   inside,
			want:  ErrSessionMalformed,
		},
		{
			name:  "an unknown version label",
			value: replaceField(valid, 0, []byte("dubplate-session-99")),
			now:   inside,
			want:  ErrSessionMalformed,
		},
		{
			name:  "an issue time that is not a number",
			value: replaceField(valid, 2, []byte("recently")),
			now:   inside,
			want:  ErrSessionMalformed,
		},
		{
			name:  "an anchor that is not a number",
			value: replaceField(valid, 3, []byte("at birth")),
			now:   inside,
			want:  ErrSessionMalformed,
		},
		{
			name:  "an expiry that is not a number",
			value: replaceField(valid, 4, []byte("never")),
			now:   inside,
			want:  ErrSessionMalformed,
		},
		{
			name:  "an empty value",
			value: "",
			now:   inside,
			want:  ErrSessionMalformed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifySession([]byte(currentFixture), []byte(previousFixture), tt.value, tt.now, sessionCapFixture)
			if !errors.Is(err, tt.want) {
				t.Fatalf("VerifySession error = %v, want %v", err, tt.want)
			}
			if got != (Session{}) {
				t.Errorf("Session = %+v on a refused value, want the zero value", got)
			}
		})
	}

	// A session minted under either secret verifies, which is the
	// whole of what rotation overlap buys; MintSession takes no
	// previous secret at all, so no value is ever minted under one.
	for _, value := range []string{valid, underPrevious} {
		got, err := VerifySession([]byte(currentFixture), []byte(previousFixture), value, inside, sessionCapFixture)
		if err != nil {
			t.Fatalf("VerifySession error = %v, want nil", err)
		}
		if want := fixtureSession(); got != want {
			t.Errorf("Session = %+v, want %+v", got, want)
		}
	}
}

// TestRefreshHoldsTheAnchor refreshes across the whole lifetime and
// asserts the anchor is byte-identical every time, against an anchor
// written out here rather than read from the value under test. The
// anchor's invariance is what makes the absolute cap real: were refresh
// to move it, a session refreshed often enough would never reach the
// cap at all, and nothing else in the suite would notice.
func TestRefreshHoldsTheAnchor(t *testing.T) {
	wantAnchor := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)

	value, session, err := MintSession([]byte(currentFixture), "alice@example.test", sessionBirth, sessionLifetimeFixture)
	if err != nil {
		t.Fatalf("MintSession error = %v, want nil", err)
	}
	if !session.AnchoredAt.Equal(wantAnchor) {
		t.Fatalf("minted AnchoredAt = %v, want %v", session.AnchoredAt, wantAnchor)
	}

	// Twelve refreshes a fortnight apart carry the session 168 days
	// past its birth, five times its own thirty-day lifetime and just
	// inside the cap, so the session survives at all only because
	// refresh moved the expiry.
	const step = 14 * 24 * time.Hour
	for i := 1; i <= 12; i++ {
		at := sessionBirth.Add(time.Duration(i) * step)

		session, err = VerifySession([]byte(currentFixture), nil, value, at, sessionCapFixture)
		if err != nil {
			t.Fatalf("refresh %d: VerifySession error = %v, want nil", i, err)
		}
		value, session = RefreshSession([]byte(currentFixture), session, at, sessionLifetimeFixture, sessionCapFixture)

		if !session.AnchoredAt.Equal(wantAnchor) {
			t.Fatalf("refresh %d: AnchoredAt = %v, want %v", i, session.AnchoredAt, wantAnchor)
		}
		// The issue time and the expiry both move, and they move
		// separately from the anchor, so a reader (or Task 9's epoch
		// check) cannot confuse the field refresh advances with the
		// field it holds.
		if !session.IssuedAt.Equal(at) {
			t.Errorf("refresh %d: IssuedAt = %v, want %v", i, session.IssuedAt, at)
		}
		// The last two rounds' unclamped expiry would land past the
		// cap, so the clamp this task adds (review m2) is what this
		// want value must reflect too.
		want := at.Add(sessionLifetimeFixture)
		if capAt := wantAnchor.Add(sessionCapFixture); want.After(capAt) {
			want = capAt
		}
		if !session.ExpiresAt.Equal(want) {
			t.Errorf("refresh %d: ExpiresAt = %v, want %v", i, session.ExpiresAt, want)
		}
		if !session.IssuedAt.After(session.AnchoredAt) {
			t.Errorf("refresh %d: IssuedAt = %v, want it past the anchor %v", i, session.IssuedAt, session.AnchoredAt)
		}
	}
}

// TestRefreshedSessionRefusedAtTheAbsoluteCap is the row that proves
// the cap is real: a refresh close enough to the cap has its expiry
// clamped to the cap itself (review m2), and VerifySession refuses once
// the anchor plus the cap has passed. A cap checked against the issue
// time would admit this session forever.
func TestRefreshedSessionRefusedAtTheAbsoluteCap(t *testing.T) {
	_, session, err := MintSession([]byte(currentFixture), "alice@example.test", sessionBirth, sessionLifetimeFixture)
	if err != nil {
		t.Fatalf("MintSession error = %v, want nil", err)
	}
	refreshedAt := sessionBirth.Add(175 * 24 * time.Hour)
	value, session := RefreshSession([]byte(currentFixture), session, refreshedAt, sessionLifetimeFixture, sessionCapFixture)

	atCap := sessionBirth.Add(sessionCapFixture)
	if !session.ExpiresAt.Equal(atCap) {
		t.Fatalf("refreshed ExpiresAt = %v, want it clamped to the cap %v", session.ExpiresAt, atCap)
	}

	for _, now := range []time.Time{atCap, atCap.Add(24 * time.Hour)} {
		got, err := VerifySession([]byte(currentFixture), nil, value, now, sessionCapFixture)
		if !errors.Is(err, ErrSessionCapReached) {
			t.Errorf("VerifySession at %v error = %v, want ErrSessionCapReached", now, err)
		}
		if got != (Session{}) {
			t.Errorf("Session = %+v on a refused value, want the zero value", got)
		}
	}

	// One second before the cap the same value still verifies, so the
	// row above reports the cap rather than a session broken some
	// other way.
	if _, err := VerifySession([]byte(currentFixture), nil, value, atCap.Add(-time.Second), sessionCapFixture); err != nil {
		t.Errorf("VerifySession one second before the cap error = %v, want nil", err)
	}
}

// TestRefreshClampsAtTheExactBoundary is the D5 acceptance literal: a
// session anchored at T with a thirty-hour cap and a twenty-four-hour
// lifetime, refreshed at T+23h, lands its expiry at exactly T+30h
// rather than the T+47h an unclamped refresh would carry (review m2).
func TestRefreshClampsAtTheExactBoundary(t *testing.T) {
	const (
		lifetime = 24 * time.Hour
		cap      = 30 * time.Hour
	)
	anchor := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	_, session, err := MintSession([]byte(currentFixture), "alice@example.test", anchor, lifetime)
	if err != nil {
		t.Fatalf("MintSession error = %v, want nil", err)
	}

	refreshedAt := anchor.Add(23 * time.Hour)
	_, refreshed := RefreshSession([]byte(currentFixture), session, refreshedAt, lifetime, cap)

	want := anchor.Add(cap)
	if !refreshed.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v (anchor %v + cap %v)", refreshed.ExpiresAt, want, anchor, cap)
	}
}

// TestSessionAndLinkCannotCross holds the domain-separation label in
// both directions. The two label rows are the ones that fail if the
// session signs under linkPurpose: each presents a well-formed value of
// the verifier's own shape whose MAC covers the other verifier's label,
// so only the label separates them.
func TestSessionAndLinkCannotCross(t *testing.T) {
	linkValue := encodeLink([]byte(currentFixture), fixtureLink("alice@example.test"))
	sessionValue := encodeSession([]byte(currentFixture), fixtureSession())

	// A magic-link token presented as a session cookie, and a session
	// cookie presented as a magic link: each carries the other's
	// format-version label, so each is refused before its MAC is
	// reached.
	if _, err := VerifySession([]byte(currentFixture), []byte(previousFixture), linkValue, fixtureMint, sessionCapFixture); !errors.Is(err, ErrSessionMalformed) {
		t.Errorf("VerifySession of a link token error = %v, want ErrSessionMalformed", err)
	}
	if _, err := VerifyLink([]byte(currentFixture), []byte(previousFixture), sessionValue, sessionBirth.Add(time.Hour)); !errors.Is(err, ErrLinkMalformed) {
		t.Errorf("VerifyLink of a session value error = %v, want ErrLinkMalformed", err)
	}

	// Session fields signed under the magic link's own
	// domain-separation label, which is what a session value would
	// verify as if the two purposes were one string.
	underLink := encodeSessionUnderLabel([]byte(currentFixture), linkPurpose, fixtureSession())
	if _, err := VerifySession([]byte(currentFixture), []byte(previousFixture), underLink, sessionBirth.Add(time.Hour), sessionCapFixture); !errors.Is(err, ErrSessionMAC) {
		t.Errorf("VerifySession of a value signed under linkPurpose error = %v, want ErrSessionMAC", err)
	}

	// The same crossing from the link's side, signed under the label
	// session.go actually chose rather than a literal, so a session
	// that adopted linkPurpose would be caught here too.
	underSession := encodeUnderLabel([]byte(currentFixture), sessionPurpose, fixtureLink("alice@example.test"))
	if _, err := VerifyLink([]byte(currentFixture), []byte(previousFixture), underSession, fixtureMint); !errors.Is(err, ErrLinkMAC) {
		t.Errorf("VerifyLink of a value signed under sessionPurpose error = %v, want ErrLinkMAC", err)
	}
}

// TestSessionHasOneEncoding is TestLinkHasOneEncoding for the cookie,
// which shares the link's decoder: one session, one string.
func TestSessionHasOneEncoding(t *testing.T) {
	value, _, err := MintSession([]byte(currentFixture), "alice@example.test", sessionBirth, sessionLifetimeFixture)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	for _, variant := range encodingVariants(value) {
		if _, err := VerifySession([]byte(currentFixture), nil, variant, sessionBirth, sessionCapFixture); err == nil {
			t.Errorf("VerifySession(%q) verified a non-canonical encoding, want a refusal", variant)
		}
	}
}

func TestMintSession(t *testing.T) {
	now := time.Date(2026, 3, 4, 5, 6, 7, 890_000_000, time.UTC)

	value, minted, err := MintSession([]byte(currentFixture), "Alice.Smith@Example.TEST", now, sessionLifetimeFixture)
	if err != nil {
		t.Fatalf("MintSession error = %v, want nil", err)
	}
	// The session carries the normalizer's output, so the value the
	// MAC covers and the value the revocation epoch digests cannot
	// diverge.
	if minted.Address != "alice.smith@example.test" {
		t.Errorf("Address = %q, want %q", minted.Address, "alice.smith@example.test")
	}
	if want := now.Truncate(time.Second); !minted.IssuedAt.Equal(want) {
		t.Errorf("IssuedAt = %v, want %v", minted.IssuedAt, want)
	}
	if !minted.AnchoredAt.Equal(minted.IssuedAt) {
		t.Errorf("AnchoredAt = %v, want the issue time %v at the first mint", minted.AnchoredAt, minted.IssuedAt)
	}
	if want := now.Truncate(time.Second).Add(sessionLifetimeFixture); !minted.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", minted.ExpiresAt, want)
	}

	parsed, err := VerifySession([]byte(currentFixture), nil, value, now, sessionCapFixture)
	if err != nil {
		t.Fatalf("VerifySession of a minted value error = %v, want nil", err)
	}
	if parsed != minted {
		t.Errorf("verified session = %+v, want %+v", parsed, minted)
	}

	if _, _, err := MintSession([]byte(currentFixture), "not-an-address", now, sessionLifetimeFixture); !errors.Is(err, identity.ErrInvalidAddress) {
		t.Errorf("MintSession of an unparseable address error = %v, want identity.ErrInvalidAddress", err)
	}
}

// TestSessionErrorsCarryNoFixtureValues searches every error the
// session half can produce for the values a refusal must never echo. A
// refusal names the check that failed; the cookie value and the secret
// stay out of logs and out of an operator's terminal.
func TestSessionErrorsCarryNoFixtureValues(t *testing.T) {
	valid := encodeSession([]byte(currentFixture), fixtureSession())
	broken := flipFieldBit(t, valid, sessionFieldCount)

	var texts []string
	record := func(err error) {
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		texts = append(texts, err.Error())
	}

	_, _, err := MintSession([]byte(currentFixture), "not-an-address", sessionBirth, sessionLifetimeFixture)
	record(err)
	_, err = VerifySession([]byte(currentFixture), []byte(previousFixture), broken, sessionBirth, sessionCapFixture)
	record(err)
	_, err = VerifySession([]byte(currentFixture), []byte(previousFixture), valid, sessionBirth.Add(sessionLifetimeFixture), sessionCapFixture)
	record(err)
	_, err = VerifySession([]byte(currentFixture), []byte(previousFixture), valid, sessionBirth.Add(sessionCapFixture), sessionCapFixture)
	record(err)
	_, err = VerifySession([]byte(currentFixture), []byte(previousFixture), "!!!", sessionBirth, sessionCapFixture)
	record(err)

	joined := strings.Join(texts, "\n")
	forbidden := []string{currentFixture, previousFixture, valid, broken, macOf(t, valid)}
	for _, value := range forbidden {
		if strings.Contains(joined, value) {
			t.Errorf("an error string carries a fixture value it must never echo (%d bytes)", len(value))
		}
	}
}

// encodeSessionUnderLabel builds a session-shaped value whose MAC
// covers label instead of sessionPurpose, which is the adversary the
// domain-separation label exists to refuse.
func encodeSessionUnderLabel(secret []byte, label string, s Session) string {
	fields := sessionFields(s)
	parts := make([]string, 0, len(fields)+1)
	for _, f := range fields {
		parts = append(parts, base64.RawURLEncoding.EncodeToString(f))
	}
	parts = append(parts, base64.RawURLEncoding.EncodeToString(sign(secret, label, fields...)))
	return strings.Join(parts, ".")
}
