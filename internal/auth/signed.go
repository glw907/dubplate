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

// fieldSeparator joins the encoded fields. Every field is
// base64url-encoded before the join, so no field value can carry the
// separator: a raw join would split a dotted local part.
const fieldSeparator = "."

// fieldEncoding is unpadded base64url that refuses non-zero trailing
// bits in a field's final character. A lax decoder ignores those bits,
// so a field whose byte length is not a multiple of three has several
// encodings that decode alike, and every one of them would verify.
// Refusing them is the canonical encoding RFC 4648 section 3.5
// describes. A strict decoder still skips CR and LF, so splitSigned
// refuses those itself, and between the two each credential has exactly
// one string.
var fieldEncoding = base64.RawURLEncoding.Strict()

// encodeSigned renders fields and the MAC computed over them under
// purpose.
func encodeSigned(secret []byte, purpose string, fields [][]byte) string {
	parts := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		parts = append(parts, fieldEncoding.EncodeToString(field))
	}
	parts = append(parts, fieldEncoding.EncodeToString(sign(secret, purpose, fields...)))
	return strings.Join(parts, fieldSeparator)
}

// splitSigned decodes value into count MAC-covered fields and the MAC
// that follows them, authenticating none of them. malformed is the
// caller's sentinel, so a refusal names the credential that was
// malformed.
//
// The split stops after count+2 parts, one more than a well-formed
// value holds, so the field-count check stays exact while a value of
// nothing but separators costs a fixed handful of substrings rather
// than one per byte.
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

// macInput length-prefixes the domain-separation label and every field.
// Without the prefixes, two different field sets whose values
// concatenate to the same bytes share a MAC input, so one token's
// signature authenticates the other's fields: a character moved from
// the end of an address onto the front of an expiry would carry its
// signature with it.
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
