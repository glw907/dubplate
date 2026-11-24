// Package id holds the 16-byte identifier every store row and every
// wire id is built on. It imports no other internal package, so any
// package may name an id without a layering cost.
package id

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"errors"
)

var errSyntax = errors.New("want 32 lowercase hex characters")

// An ID is a random 16-byte key. It prints, and is parsed from, 32
// lowercase hex characters.
type ID [16]byte

// New returns a random id. It reports no error because [rand.Read]
// never returns one and always fills its buffer entirely.
func New() ID {
	var id ID
	_, _ = rand.Read(id[:])
	return id
}

// Parse returns the id s spells: exactly 32 lowercase hex characters.
// It says nothing about the all-zero id, which a caller that refuses it
// checks with [ID.IsZero].
func Parse(s string) (ID, error) {
	if len(s) != hex.EncodedLen(len(ID{})) {
		return ID{}, errSyntax
	}
	b, err := hex.DecodeString(s)
	if err != nil || hex.EncodeToString(b) != s {
		return ID{}, errSyntax
	}
	return ID(b), nil
}

// String returns the 32 lowercase hex characters [Parse] accepts.
func (id ID) String() string { return hex.EncodeToString(id[:]) }

// IsZero reports whether id is all zero bytes.
func (id ID) IsZero() bool { return id == ID{} }

// Value implements [driver.Valuer]: the store holds an id as a 16-byte
// blob.
func (id ID) Value() (driver.Value, error) { return id[:], nil }
