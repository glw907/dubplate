// Package store manages the group's SQLite database: the leaf it lives
// under, the lock that serializes access to it, and, as later work
// lands, the connections, migrations, and write path built on that
// foundation.
//
// # Migrations
//
// A migration is a SQL file named "<version>_<label>.sql", numbered
// from 1 with no gap, applied once in its own transaction. The first
// sets PRAGMA application_id to the store marker, 0x64756270, and
// creates the ledger table, migrations, whose columns version, name,
// and sha256 record each applied file's number, name, and lowercase
// hex SHA-256. A merged file never changes, because an existing
// store's ledger would then refuse it. A file
// carries no statement that ends or nests a transaction (COMMIT, END,
// ROLLBACK, SAVEPOINT, RELEASE) and no VACUUM or ATTACH.
//
// A file that rebuilds a table with foreign keys off, the procedure
// sqlite.org/lang_altertable.html gives for changes ALTER TABLE cannot
// make, begins with this exact first line:
//
//	-- dubplate:foreign-keys-off
//
// It then runs on a connection of its own with enforcement off, and
// commits only when PRAGMA foreign_key_check reports no row.
//
// The package takes every value it needs as a parameter. It reads no
// environment variable and calls no clock function; a caller passes a
// state directory, a context for any bounded wait, and any threshold
// the package needs.
package store
