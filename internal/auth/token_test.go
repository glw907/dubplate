package auth

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// The fixtures below are visibly synthetic and deliberately low
// entropy: every one clears the 32-byte length floor while reading as
// an example rather than as a credential, which is what keeps the
// gate's secret scan quiet without an allowlist entry.
const (
	currentFixture  = "dubplate-fixture-current-0000000"
	previousFixture = "dubplate-fixture-previous-000000"

	// foreignLabelFixture is the domain-separation label the session
	// cookie signs under. It is written out here rather than read
	// from session.go, which lands next task, so the cross-purpose
	// row exists before its neighbour does; the row asserts only
	// that a label other than linkPurpose fails, which holds whatever
	// string the session ends up choosing, so long as it is not
	// linkPurpose.
	foreignLabelFixture = "dubplate/session"
)

// fixtureMint, fixtureExpiry, and fixtureNonce pin everything the
// golden vectors depend on, so the checked-in literals below are a
// frozen statement of the wire format rather than whatever the code
// happens to produce today.
var (
	fixtureMint   = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	fixtureExpiry = 15 * time.Minute
	fixtureNonce  = [NonceLength]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
)

func fixtureLink(address string) LinkToken {
	return LinkToken{
		Address:   address,
		MintedAt:  fixtureMint,
		ExpiresAt: fixtureMint.Add(fixtureExpiry),
		Nonce:     fixtureNonce,
	}
}

// TestLinkGoldenVectors freezes the wire format against literals
// derived from the format's written description rather than recomputed
// with the code under test. The second and third rows are the ones a
// raw join would break: a dotted local part is the majority of real
// addresses, and "+" and "=" are both ordinary address characters and
// characters a query would have to escape.
func TestLinkGoldenVectors(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		wantEncoded string
	}{
		{
			name:        "plain address",
			address:     "alice@example.test",
			wantEncoded: "ZHVicGxhdGUtbGluay0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc2NzMyMzk0NQ.MTc2NzMyMzA0NQ.AAECAwQFBgcICQoLDA0ODw.AlrfWVh3uFMLq21CzjH0PEnBk_Ai8FZeiotgxKl0MRg",
		},
		{
			name:        "local part carrying the field separator",
			address:     "first.last@example.test",
			wantEncoded: "ZHVicGxhdGUtbGluay0x.Zmlyc3QubGFzdEBleGFtcGxlLnRlc3Q.MTc2NzMyMzk0NQ.MTc2NzMyMzA0NQ.AAECAwQFBgcICQoLDA0ODw.rOmgpcI4NPwkZAOupqWAV7zoKk6mpQRnQ7QKDj-Bizk",
		},
		{
			name:        "local part carrying plus and equals",
			address:     "a+b=c@example.test",
			wantEncoded: "ZHVicGxhdGUtbGluay0x.YStiPWNAZXhhbXBsZS50ZXN0.MTc2NzMyMzk0NQ.MTc2NzMyMzA0NQ.AAECAwQFBgcICQoLDA0ODw.LyEZD7p-4N5aSAkl3cxLv6LDQvhcdHrbgskGp5jeZkI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeLink([]byte(currentFixture), fixtureLink(tt.address))
			if got != tt.wantEncoded {
				t.Errorf("encodeLink(%q) = %q, want %q", tt.address, got, tt.wantEncoded)
			}
			parsed, err := VerifyLink([]byte(currentFixture), nil, tt.wantEncoded, fixtureMint)
			if err != nil {
				t.Fatalf("VerifyLink(golden) error = %v, want nil", err)
			}
			if parsed.Address != tt.address {
				t.Errorf("address = %q, want %q", parsed.Address, tt.address)
			}
			if !parsed.MintedAt.Equal(fixtureMint) {
				t.Errorf("MintedAt = %v, want %v", parsed.MintedAt, fixtureMint)
			}
			if !parsed.ExpiresAt.Equal(fixtureMint.Add(fixtureExpiry)) {
				t.Errorf("ExpiresAt = %v, want %v", parsed.ExpiresAt, fixtureMint.Add(fixtureExpiry))
			}
			if parsed.Nonce != fixtureNonce {
				t.Errorf("Nonce = %x, want %x", parsed.Nonce, fixtureNonce)
			}
			// base64url spends no character a query would escape,
			// so an address carrying "+" or "=" reaches the link
			// as itself rather than as a percent sequence a mail
			// client might mangle.
			if escaped := url.QueryEscape(got); escaped != got {
				t.Errorf("the encoded value needs query escaping: %q escapes to %q", got, escaped)
			}
		})
	}
}

// TestVerifyLinkAcceptsPreviousSecret proves rotation: the literal was
// signed with the previous secret alone, and nothing in this package
// can produce it, because MintLink takes one secret.
func TestVerifyLinkAcceptsPreviousSecret(t *testing.T) {
	const rotatedVector = "ZHVicGxhdGUtbGluay0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc2NzMyMzk0NQ.MTc2NzMyMzA0NQ.AAECAwQFBgcICQoLDA0ODw.49wAiyiCUlzUEgQ_aTQIexh29plOX3m493a6-1zJIBM"

	parsed, err := VerifyLink([]byte(currentFixture), []byte(previousFixture), rotatedVector, fixtureMint)
	if err != nil {
		t.Fatalf("VerifyLink with previous secret error = %v, want nil", err)
	}
	if parsed.Address != "alice@example.test" {
		t.Errorf("address = %q, want %q", parsed.Address, "alice@example.test")
	}
	if _, err := VerifyLink([]byte(currentFixture), nil, rotatedVector, fixtureMint); !errors.Is(err, ErrLinkMAC) {
		t.Errorf("VerifyLink without previous secret error = %v, want ErrLinkMAC", err)
	}
}

func TestVerifyLinkRefusals(t *testing.T) {
	valid := encodeLink([]byte(currentFixture), fixtureLink("alice@example.test"))
	expired := fixtureMint.Add(fixtureExpiry).Add(time.Second)

	tests := []struct {
		name  string
		value string
		now   time.Time
		want  error
	}{
		{
			name:  "one bit flipped in the MAC",
			value: flipFieldBit(t, valid, linkFieldCount),
			now:   fixtureMint,
			want:  ErrLinkMAC,
		},
		{
			name:  "one bit flipped in the address",
			value: flipFieldBit(t, valid, 1),
			now:   fixtureMint,
			want:  ErrLinkMAC,
		},
		{
			// A value signed under the session cookie's own
			// domain-separation label, which is what a caller
			// would present if a session value were handed to the
			// link verifier under a shared secret.
			name:  "signed under the session domain-separation label",
			value: encodeUnderLabel([]byte(currentFixture), foreignLabelFixture, fixtureLink("alice@example.test")),
			now:   fixtureMint,
			want:  ErrLinkMAC,
		},
		{
			name:  "past its expiry",
			value: valid,
			now:   expired,
			want:  ErrLinkExpired,
		},
		{
			// The MAC is checked first, so a value that fails
			// both reports the MAC. Reversing the order would
			// report an expiry read off an unauthenticated field.
			name:  "expired and MAC broken reports the MAC",
			value: flipFieldBit(t, valid, linkFieldCount),
			now:   expired,
			want:  ErrLinkMAC,
		},
		{
			name:  "exactly at its expiry",
			value: valid,
			now:   fixtureMint.Add(fixtureExpiry),
			want:  ErrLinkExpired,
		},
		{
			name:  "too few fields",
			value: strings.Join(strings.Split(valid, ".")[:linkFieldCount], "."),
			now:   fixtureMint,
			want:  ErrLinkMalformed,
		},
		{
			name:  "a field that is not base64url",
			value: "!!!." + strings.SplitN(valid, ".", 2)[1],
			now:   fixtureMint,
			want:  ErrLinkMalformed,
		},
		{
			name:  "an unknown version label",
			value: replaceField(valid, 0, []byte("dubplate-link-99")),
			now:   fixtureMint,
			want:  ErrLinkMalformed,
		},
		{
			name:  "a mint time that is not a number",
			value: replaceField(valid, 3, []byte("soon")),
			now:   fixtureMint,
			want:  ErrLinkMalformed,
		},
		{
			name:  "a nonce of the wrong length",
			value: replaceField(valid, 4, fixtureNonce[:NonceLength-1]),
			now:   fixtureMint,
			want:  ErrLinkMalformed,
		},
		{
			name:  "an empty value",
			value: "",
			now:   fixtureMint,
			want:  ErrLinkMalformed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyLink([]byte(currentFixture), []byte(previousFixture), tt.value, tt.now)
			if !errors.Is(err, tt.want) {
				t.Fatalf("VerifyLink error = %v, want %v", err, tt.want)
			}
			// A refused value yields no nonce, so a caller cannot
			// insert one for a value that failed either check.
			if got.Nonce != ([NonceLength]byte{}) {
				t.Errorf("Nonce = %x on a refused value, want the zero value", got.Nonce)
			}
			if got != (LinkToken{}) {
				t.Errorf("LinkToken = %+v on a refused value, want the zero value", got)
			}
		})
	}
}

// TestMACLengthPrefixesResistFieldShifting presents two field sets
// whose values concatenate to the same bytes: a trailing "1" moved from
// the front of the expiry onto the end of the address. Without a length
// prefix per field the two MAC inputs are identical, so one token's MAC
// would authenticate the other's fields, which moves an expiry fifteen
// years into the future.
func TestMACLengthPrefixesResistFieldShifting(t *testing.T) {
	left := fixtureLink("alice@example.test")
	right := fixtureLink("alice@example.test1")
	right.ExpiresAt = time.Unix(767323945, 0).UTC()

	leftJoined := left.Address + strconv.FormatInt(left.ExpiresAt.Unix(), 10)
	rightJoined := right.Address + strconv.FormatInt(right.ExpiresAt.Unix(), 10)
	if leftJoined != rightJoined {
		t.Fatalf("fixtures do not collide when concatenated: %q vs %q", leftJoined, rightJoined)
	}

	leftMAC := macOf(t, encodeLink([]byte(currentFixture), left))
	rightMAC := macOf(t, encodeLink([]byte(currentFixture), right))
	if leftMAC == rightMAC {
		t.Fatalf("colliding field sets share a MAC (%q); the length prefix is gone", leftMAC)
	}

	// The same property from the verifier's side: the right-hand
	// fields carrying the left-hand MAC are refused.
	spliced := replaceField(encodeLink([]byte(currentFixture), right), linkFieldCount, decodeField(t, leftMAC))
	if _, err := VerifyLink([]byte(currentFixture), nil, spliced, fixtureMint); !errors.Is(err, ErrLinkMAC) {
		t.Errorf("spliced value error = %v, want ErrLinkMAC", err)
	}
}

func TestMintLink(t *testing.T) {
	now := time.Date(2026, 3, 4, 5, 6, 7, 890_000_000, time.UTC)

	value, minted, err := MintLink([]byte(currentFixture), "Alice.Smith@Example.TEST", now, fixtureExpiry)
	if err != nil {
		t.Fatalf("MintLink error = %v, want nil", err)
	}
	// The minter normalizes rather than trusting its caller, so the
	// address a MAC covers is the address every later lookup keys on.
	if minted.Address != "alice.smith@example.test" {
		t.Errorf("address = %q, want %q", minted.Address, "alice.smith@example.test")
	}
	if want := now.Truncate(time.Second); !minted.MintedAt.Equal(want) {
		t.Errorf("MintedAt = %v, want %v", minted.MintedAt, want)
	}
	if want := now.Truncate(time.Second).Add(fixtureExpiry); !minted.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", minted.ExpiresAt, want)
	}
	if minted.Nonce == ([NonceLength]byte{}) {
		t.Error("Nonce is the zero value, want random bytes")
	}

	parsed, err := VerifyLink([]byte(currentFixture), nil, value, now)
	if err != nil {
		t.Fatalf("VerifyLink of a minted value error = %v, want nil", err)
	}
	if parsed != minted {
		t.Errorf("verified token = %+v, want %+v", parsed, minted)
	}

	_, second, err := MintLink([]byte(currentFixture), "alice.smith@example.test", now, fixtureExpiry)
	if err != nil {
		t.Fatalf("MintLink error = %v, want nil", err)
	}
	if second.Nonce == minted.Nonce {
		t.Error("two mints share a nonce, want a fresh 128-bit value per link")
	}

	if _, _, err := MintLink([]byte(currentFixture), "not-an-address", now, fixtureExpiry); !errors.Is(err, identity.ErrInvalidAddress) {
		t.Errorf("MintLink of an unparseable address error = %v, want identity.ErrInvalidAddress", err)
	}
}

// TestLinkURL holds security review finding 7: the link is built from
// the configured public origin and from nothing a request carries, so
// the URL changes when, and only when, the configuration does.
func TestLinkURL(t *testing.T) {
	const vector = "ZHVicGxhdGUtbGluay0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc2NzMyMzk0NQ.MTc2NzMyMzA0NQ.AAECAwQFBgcICQoLDA0ODw.AlrfWVh3uFMLq21CzjH0PEnBk_Ai8FZeiotgxKl0MRg"

	tests := []struct {
		name    string
		origin  string
		want    string
		wantErr bool
	}{
		{
			name:   "origin at the root",
			origin: "https://example.test/",
			want:   "https://example.test/login/link?t=ZHVicGxhdGUtbGluay0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc2NzMyMzk0NQ.MTc2NzMyMzA0NQ.AAECAwQFBgcICQoLDA0ODw.AlrfWVh3uFMLq21CzjH0PEnBk_Ai8FZeiotgxKl0MRg",
		},
		{
			name:   "a different origin, under a path prefix",
			origin: "https://listen.example.test/dubplate",
			want:   "https://listen.example.test/dubplate/login/link?t=ZHVicGxhdGUtbGluay0x.YWxpY2VAZXhhbXBsZS50ZXN0.MTc2NzMyMzk0NQ.MTc2NzMyMzA0NQ.AAECAwQFBgcICQoLDA0ODw.AlrfWVh3uFMLq21CzjH0PEnBk_Ai8FZeiotgxKl0MRg",
		},
		{
			name:    "an origin net/url refuses",
			origin:  "://bad",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LinkURL(tt.origin, vector)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("LinkURL(%q) error = nil, want an error", tt.origin)
				}
				return
			}
			if err != nil {
				t.Fatalf("LinkURL(%q) error = %v, want nil", tt.origin, err)
			}
			if got != tt.want {
				t.Errorf("LinkURL(%q) = %q, want %q", tt.origin, got, tt.want)
			}
		})
	}
}

// TestErrorsCarryNoFixtureValues searches every error this package can
// produce for the values a refusal must never echo. An error names the
// check that failed; the value that failed it stays out of logs and out
// of an operator's terminal.
func TestErrorsCarryNoFixtureValues(t *testing.T) {
	valid := encodeLink([]byte(currentFixture), fixtureLink("alice@example.test"))
	broken := flipFieldBit(t, valid, linkFieldCount)

	var texts []string
	record := func(err error) {
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		texts = append(texts, err.Error())
	}

	_, _, err := MintLink([]byte(currentFixture), "not-an-address", fixtureMint, fixtureExpiry)
	record(err)
	_, err = VerifyLink([]byte(currentFixture), []byte(previousFixture), broken, fixtureMint)
	record(err)
	_, err = VerifyLink([]byte(currentFixture), []byte(previousFixture), valid, fixtureMint.Add(time.Hour))
	record(err)
	_, err = VerifyLink([]byte(currentFixture), []byte(previousFixture), "!!!", fixtureMint)
	record(err)
	_, err = LinkURL("://bad", valid)
	record(err)

	joined := strings.Join(texts, "\n")
	forbidden := []string{
		currentFixture,
		previousFixture,
		valid,
		broken,
		base64.RawURLEncoding.EncodeToString(fixtureNonce[:]),
	}
	for _, value := range forbidden {
		if strings.Contains(joined, value) {
			t.Errorf("an error string carries a fixture value it must never echo (%d bytes)", len(value))
		}
	}
}

// flipFieldBit returns value with the low bit of field index i's first
// decoded byte inverted, re-encoded in place.
func flipFieldBit(t *testing.T, value string, i int) string {
	t.Helper()
	fields := strings.Split(value, ".")
	if i >= len(fields) {
		t.Fatalf("field index %d out of range for %d fields", i, len(fields))
	}
	raw, err := base64.RawURLEncoding.DecodeString(fields[i])
	if err != nil {
		t.Fatalf("decode field %d: %v", i, err)
	}
	raw[0] ^= 0x01
	fields[i] = base64.RawURLEncoding.EncodeToString(raw)
	return strings.Join(fields, ".")
}

func replaceField(value string, i int, raw []byte) string {
	fields := strings.Split(value, ".")
	fields[i] = base64.RawURLEncoding.EncodeToString(raw)
	return strings.Join(fields, ".")
}

func macOf(t *testing.T, value string) string {
	t.Helper()
	fields := strings.Split(value, ".")
	return fields[len(fields)-1]
}

func decodeField(t *testing.T, field string) []byte {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(field)
	if err != nil {
		t.Fatalf("decode field: %v", err)
	}
	return raw
}

// encodeUnderLabel builds a link-shaped value whose MAC covers label
// instead of linkPurpose, which is the adversary the domain-separation
// label exists to refuse.
func encodeUnderLabel(secret []byte, label string, tok LinkToken) string {
	fields := linkFields(tok)
	parts := make([]string, 0, len(fields)+1)
	for _, f := range fields {
		parts = append(parts, base64.RawURLEncoding.EncodeToString(f))
	}
	parts = append(parts, base64.RawURLEncoding.EncodeToString(sign(secret, label, fields...)))
	return strings.Join(parts, ".")
}
