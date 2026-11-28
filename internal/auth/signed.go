package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

// Both credentials share one wire shape: a run of base64url-encoded
// fields joined by fieldSeparator with the MAC last. They differ only in
// their field list and their domain-separation label.

// Fields are base64url-encoded before the join, so no field can carry
// the separator, as a dotted local part otherwise would.
const fieldSeparator = "."

// fieldEncoding refuses non-zero trailing bits, the canonical encoding
// of RFC 4648 section 3.5: a lax decoder ignores them, giving a field
// several encodings that all verify. A strict decoder still skips CR
// and LF, so splitSigned refuses those, and each credential has exactly
// one string.
var fieldEncoding = base64.RawURLEncoding.Strict()

func encodeSigned(secret []byte, purpose string, fields [][]byte) string {
	parts := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		parts = append(parts, fieldEncoding.EncodeToString(field))
	}
	parts = append(parts, fieldEncoding.EncodeToString(sign(secret, purpose, fields...)))
	return strings.Join(parts, fieldSeparator)
}

// splitSigned decodes value into count fields and the MAC after them,
// authenticating none; malformed is the caller's sentinel. Splitting at
// count+2 parts keeps the count check exact while a value of only
// separators costs a fixed handful of substrings.
func splitSigned(value string, count int, malformed error) ([][]byte, []byte, error) {
	if strings.ContainsAny(value, "\r\n") {
		return nil, nil, fmt.Errorf("%w: field encoding", malformed)
	}
	parts := strings.SplitN(value, fieldSeparator, count+2)
	if len(parts) != count+1 {
		return nil, nil, fmt.Errorf("%w: field count", malformed)
	}
	decoded := make([][]byte, 0, len(parts))
	for _, part := range parts {
		raw, err := fieldEncoding.DecodeString(part)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: field encoding", malformed)
		}
		decoded = append(decoded, raw)
	}
	return decoded[:count], decoded[count], nil
}

// macMatches reports, in constant time, whether mac is the MAC of
// fields under secret and purpose. An empty secret matches nothing, so
// an unconfigured rotation overlap is absent rather than permissive.
func macMatches(secret []byte, purpose string, fields [][]byte, mac []byte) bool {
	if len(secret) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(sign(secret, purpose, fields...), mac) == 1
}

// sign returns the HMAC-SHA256 of purpose and fields under secret.
func sign(secret []byte, purpose string, fields ...[]byte) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(macInput(purpose, fields))
	return mac.Sum(nil)
}

// macInput length-prefixes the label and every field. Without the
// prefixes, a character moved from the end of an address onto the front
// of an expiry would keep the signature valid.
func macInput(purpose string, fields [][]byte) []byte {
	buf := appendField(nil, []byte(purpose))
	for _, field := range fields {
		buf = appendField(buf, field)
	}
	return buf
}

func appendField(dst, field []byte) []byte {
	dst = binary.BigEndian.AppendUint64(dst, uint64(len(field)))
	return append(dst, field...)
}
