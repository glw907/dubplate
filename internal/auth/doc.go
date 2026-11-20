// Package auth mints and verifies the credentials a dubplate login
// produces: the magic-link token a member receives by mail, and the
// session cookie a consumed link mints.
//
// Every value the package signs is a MAC over length-prefixed fields
// under two labels, a format version and a domain-separation label, so
// a value minted for one purpose cannot be presented for another and no
// field can be shifted across a boundary. Only config.go reads the
// environment or imports internal/config: it declares the package's
// settings and turns them into plain values, so secrets, lifetimes, and
// the public origin reach the rest of the package as arguments, which
// keeps the construction testable with fixtures.
//
// No error string and no log record this package produces carries a
// token, a nonce, or a secret. A refusal names the check that failed.
package auth
