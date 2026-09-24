// Package store manages the group's SQLite database: the leaf it lives
// under, the lock that serializes access to it, and, as later work
// lands, the connections, migrations, and write path built on that
// foundation.
//
// The package takes every value it needs as a parameter. It reads no
// environment variable and calls no clock function; a caller passes a
// state directory, a context for any bounded wait, and any threshold
// the package needs.
package store
