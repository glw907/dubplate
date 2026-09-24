package store

import "crypto/rand"

// idLen is the byte length of every entity id this package mints.
const idLen = 16

// id mints idLen random bytes for a new entity's primary key.
// crypto/rand.Read "fills b with cryptographically secure random
// bytes. It never returns an error, and always fills b entirely" (go
// doc crypto/rand.Read, go1.27.1), so id reports no error either.
func id() []byte {
	b := make([]byte, idLen)
	_, _ = rand.Read(b)
	return b
}
