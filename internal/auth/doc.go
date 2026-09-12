// Package auth mints and verifies the credentials a dubplate login
// produces: the magic-link token a member receives by mail, and the
// session cookie a consumed link mints.
//
// Every value the package signs is a MAC over length-prefixed fields
// under two labels, a format version and a domain-separation label, so
// a value minted for one purpose cannot be presented for another and no
// field can be shifted across a boundary. Nothing here reads the
// environment or imports internal/config: secrets, lifetimes, and the
// public origin arrive as plain arguments resolved in
// cmd/dubplate-server, which keeps the construction testable with
// fixtures and keeps configuration in one place.
//
// No error string and no log record this package produces carries a
// token, a nonce, or a secret. A refusal names the check that failed.
package auth
