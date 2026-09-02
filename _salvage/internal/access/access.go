// Package access verifies Cloudflare Access identity assertions. Every
// request into `dubplate serve` must carry a valid Cf-Access-Jwt-Assertion
// header signed by the configured Access team, which is the only proof
// that the host's reverse-proxy layer (cloudflared, terminated behind
// Access) actually sat in front of the request rather than a process on
// the box forging the header directly.
package access

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwksRefreshCooldown bounds how often an unknown kid may trigger a JWKS
// refetch. Access rotates its signing keys on its own six-week cycle with
// a seven-day overlap, so a legitimate new key is never urgent; the
// cooldown exists to stop a burst of tokens carrying unknown kids -- a
// stale cache mid-rotation, or an attacker probing kids -- from forcing
// one certs-endpoint request per request.
const jwksRefreshCooldown = 5 * time.Minute

// certsPath is appended to the team domain to build the JWKS endpoint
// Cloudflare Access publishes its signing keys at.
const certsPath = "/cdn-cgi/access/certs"

// jwksTimeout bounds one certs-endpoint fetch.
const jwksTimeout = 10 * time.Second

// ErrUnknownKey marks a token whose kid is absent from the verifier's
// JWKS, even after a refetch.
var ErrUnknownKey = errors.New("access: unknown signing key")

// accessClaims is the subset of a Cloudflare Access identity assertion
// [Verifier.Verify] needs: the registered claims (iss/aud/exp/nbf) plus
// the verified user's email, which Access always includes.
type accessClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// jwk is one entry in a JSON Web Key Set, restricted to the RSA fields
// Cloudflare Access publishes.
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// jwksDocument is the certs endpoint's top-level JSON shape.
type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

// Verifier validates Cloudflare Access identity assertions for one team
// domain and audience. The zero value is not usable; construct with
// [NewVerifier].
type Verifier struct {
	certsURL string
	iss      string
	aud      string
	client   *http.Client

	mu          sync.Mutex
	keys        map[string]*rsa.PublicKey
	lastRefresh time.Time
}

// NewVerifier returns a Verifier for teamDomain (a bare Cloudflare Access
// team domain, such as "family.cloudflareaccess.com" -- no scheme) and
// the pinned application audience aud. It performs no network I/O; the
// JWKS is fetched lazily, on the first call to [Verifier.Verify] that
// needs a key.
func NewVerifier(teamDomain, aud string) *Verifier {
	return &Verifier{
		certsURL: "https://" + teamDomain + certsPath,
		iss:      "https://" + teamDomain,
		aud:      aud,
		client:   &http.Client{Timeout: jwksTimeout},
	}
}

// Verify parses and validates token as a Cloudflare Access identity
// assertion: RS256 signature only (alg=none and every other algorithm are
// rejected before the key lookup runs), against the JWKS key named by the
// token's kid; issuer equal to the team domain; audience equal to the
// pinned AUD; exp and nbf both present and within range. It returns the
// token's verified email claim.
func (v *Verifier) Verify(ctx context.Context, token string) (string, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.iss),
		jwt.WithAudience(v.aud),
		jwt.WithExpirationRequired(),
		jwt.WithNotBeforeRequired(),
	)

	var claims accessClaims
	_, err := parser.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("access: token header has no kid")
		}
		return v.key(ctx, kid)
	})
	if err != nil {
		return "", fmt.Errorf("verify access token: %w", err)
	}
	if claims.Email == "" {
		return "", errors.New("access: token has no email claim")
	}
	return claims.Email, nil
}

// key returns the RSA public key for kid, refetching the JWKS once if kid
// is not yet cached. A refetch is skipped, returning [ErrUnknownKey]
// directly, when the last refetch was within [jwksRefreshCooldown] -- the
// rate limit that keeps a burst of unknown kids to at most one
// certs-endpoint request per window.
func (v *Verifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	if !v.lastRefresh.IsZero() && time.Since(v.lastRefresh) < jwksRefreshCooldown {
		return nil, ErrUnknownKey
	}
	if err := v.refreshLocked(ctx); err != nil {
		return nil, err
	}
	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	return nil, ErrUnknownKey
}

// refreshLocked fetches and parses the JWKS, replacing v.keys wholesale.
// It records the attempt in v.lastRefresh before the request runs, so a
// slow or failing certs endpoint still starts the next cooldown window
// rather than being retried on every subsequent unknown kid. Callers hold
// v.mu.
func (v *Verifier) refreshLocked(ctx context.Context) error {
	v.lastRefresh = time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.certsURL, nil)
	if err != nil {
		return fmt.Errorf("build jwks request: %w", err)
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch jwks %s: %w", v.certsURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch jwks %s: status %s", v.certsURL, resp.Status)
	}

	// 1MiB is generous for a key set that in practice holds a handful of
	// RSA keys; the cap exists only to bound a misbehaving or malicious
	// endpoint, not because real responses approach it.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read jwks %s: %w", v.certsURL, err)
	}

	var doc jwksDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("parse jwks %s: %w", v.certsURL, err)
	}

	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		pub, err := rsaPublicKeyFromJWK(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	v.keys = keys
	return nil
}

// rsaPublicKeyFromJWK decodes a JWK's base64url-encoded modulus and
// exponent into an [rsa.PublicKey].
func rsaPublicKeyFromJWK(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode jwk %s modulus: %w", k.Kid, err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode jwk %s exponent: %w", k.Kid, err)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}
