package store

import "crypto/rand"

// idLen is the byte length of every id [NewID] mints.
const idLen = 16

// NewID returns 16 random bytes for a new entity's primary key, a job's
// id among them. It reports no error because [rand.Read] never returns
// one and always fills its buffer entirely.
func NewID() []byte {
	b := make([]byte, idLen)
	_, _ = rand.Read(b)
	return b
}
