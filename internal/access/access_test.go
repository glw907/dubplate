package access

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testTeamDomain = "family.cloudflareaccess.com"
	testIss        = "https://" + testTeamDomain
	testAUD        = "test-application-audience"
	testEmail      = "geoff@example.com"
)

// jwksServer serves keys at [certsPath] and reports how many requests it
// received, so tests can assert the rate limit rather than only the
// verification outcome.
type jwksServer struct {
	*httptest.Server
	requests atomic.Int32
}

// newJWKSServer starts an httptest server publishing pub under kid. No
// test in this package dials a real network JWKS endpoint.
func newJWKSServer(t *testing.T, kid string, pub *rsa.PublicKey) *jwksServer {
	t.Helper()

	js := &jwksServer{}
	mux := http.NewServeMux()
	mux.HandleFunc(certsPath, func(w http.ResponseWriter, r *http.Request) {
		js.requests.Add(1)
		doc := jwksDocument{Keys: []jwk{encodeJWK(kid, pub)}}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			t.Fatalf("encode jwks response: %v", err)
		}
	})
	js.Server = httptest.NewServer(mux)
	t.Cleanup(js.Close)
	return js
}

// encodeJWK is the test-side inverse of [rsaPublicKeyFromJWK]: it builds
// the JSON Web Key a real Cloudflare Access certs endpoint would publish
// for pub.
func encodeJWK(kid string, pub *rsa.PublicKey) jwk {
	e := big.NewInt(int64(pub.E)).Bytes()
	return jwk{
		Kid: kid,
		Kty: "RSA",
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(e),
	}
}

// newTestVerifier returns a Verifier wired at server rather than the real
// Cloudflare Access certs endpoint.
func newTestVerifier(server *httptest.Server) *Verifier {
	v := NewVerifier(testTeamDomain, testAUD)
	v.certsURL = server.URL + certsPath
	return v
}

// signToken signs claims with method and key, setting the kid header
// when kid is non-empty.
func signToken(t *testing.T, method jwt.SigningMethod, key any, kid string, claims accessClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(method, claims)
	if kid != "" {
		token.Header["kid"] = kid
	}
	ss, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return ss
}

// validClaims returns claims that pass every check except whatever the
// caller overrides afterward.
func validClaims() accessClaims {
	now := time.Now()
	return accessClaims{
		Email:     testEmail,
		Issuer:    testIss,
		Audience:  jwt.ClaimStrings{testAUD},
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		NotBefore: jwt.NewNumericDate(now.Add(-time.Minute)),
		IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
	}
}

func TestVerifierAcceptsValidToken(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	tok := signToken(t, jwt.SigningMethodRS256, priv, "kid-1", validClaims())

	email, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if email != testEmail {
		t.Errorf("email = %q, want %q", email, testEmail)
	}
}

func TestVerifierRejectsWrongIssuer(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	claims := validClaims()
	claims.Issuer = "https://someone-elses-team.cloudflareaccess.com"
	tok := signToken(t, jwt.SigningMethodRS256, priv, "kid-1", claims)

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("Verify: want error for wrong issuer, got nil")
	}
}

func TestVerifierRejectsWrongAudience(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	claims := validClaims()
	claims.Audience = jwt.ClaimStrings{"some-other-application"}
	tok := signToken(t, jwt.SigningMethodRS256, priv, "kid-1", claims)

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("Verify: want error for wrong audience, got nil")
	}
}

func TestVerifierRejectsExpiredToken(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	tok := signToken(t, jwt.SigningMethodRS256, priv, "kid-1", claims)

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("Verify: want error for expired token, got nil")
	}
}

func TestVerifierRejectsNotYetValidToken(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	claims := validClaims()
	claims.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Hour))
	tok := signToken(t, jwt.SigningMethodRS256, priv, "kid-1", claims)

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("Verify: want error for not-yet-valid token, got nil")
	}
}

// TestVerifierRejectsAlgNone proves the alg allowlist runs before any key
// lookup: a "none"-signed token, which by definition needs no key,  must
// still be rejected rather than accepted for having no signature to
// break.
func TestVerifierRejectsAlgNone(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	tok := signToken(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, "kid-1", validClaims())

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("Verify: want error for alg=none, got nil")
	}
}

// TestVerifierRejectsHS256 proves the RS256-only allowlist blocks the
// classic algorithm-confusion attack: an HMAC-signed token using (say)
// the RSA public key's bytes as the HMAC secret must never verify.
func TestVerifierRejectsHS256(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	tok := signToken(t, jwt.SigningMethodHS256, []byte("attacker-controlled-secret"), "kid-1", validClaims())

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("Verify: want error for HS256, got nil")
	}
}

// TestVerifierUnknownKidRateLimitsRefetch proves a burst of tokens
// carrying a kid the JWKS never publishes triggers exactly one
// certs-endpoint refetch, not one per request.
func TestVerifierUnknownKidRateLimitsRefetch(t *testing.T) {
	published, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	unpublished, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-published", &published.PublicKey)
	v := newTestVerifier(js.Server)

	tok := signToken(t, jwt.SigningMethodRS256, unpublished, "kid-unpublished", validClaims())

	const burst = 5
	for i := range burst {
		if _, err := v.Verify(context.Background(), tok); err == nil {
			t.Fatalf("Verify call %d: want error for unpublished kid, got nil", i)
		}
	}

	if got := js.requests.Load(); got != 1 {
		t.Errorf("jwks requests = %d, want exactly 1 for a %d-request burst", got, burst)
	}
}

// TestVerifierRejectsMissingEmail proves a syntactically valid,
// correctly signed token with no email claim is still rejected: the
// verified email is what the rest of dubplate keys identity on, so a
// token that carries none must never resolve to a caller.
func TestVerifierRejectsMissingEmail(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	js := newJWKSServer(t, "kid-1", &priv.PublicKey)
	v := newTestVerifier(js.Server)

	claims := validClaims()
	claims.Email = ""
	tok := signToken(t, jwt.SigningMethodRS256, priv, "kid-1", claims)

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("Verify: want error for missing email claim, got nil")
	}
}
