// Package store manages the group's SQLite database: the leaf it lives
// under, the lock that serializes access to it, its migrations, the
// write path every change goes through, and the leases that tell a
// live job's owner from a dead one.
//
// [OpenServe] creates or migrates a store and returns a handle at the
// binary's version; [OpenWrite] opens only a store already at it.
// Every change runs inside [Handle.Update], and [Inspect] reads a
// store's state without changing it.
//
// # Leases
//
// The leaf's leases/ directory holds two kinds of lease, each an
// exclusive flock(2) on a file there. A process that runs jobs holds
// an instance lease for its life, and [Handle.TryLease] tells whether
// the owner a job record names is still alive. A kind lease admits one
// job of a kind at a time across every process.
//
// # Backups
//
// [Handle.BackupCopy] stages a verified copy of the store for a caller
// to send elsewhere, and [Handle.WriteBackupRef] then records the send
// in snapshots/ as backup-<stamp>.ref. Every stamp is written by
// [FormatStamp] and read by [ParseStamp], the layout's one home. A
// reference, like a pre-migration copy, keeps serve from creating an
// empty store where a database file has gone missing.
//
// # Restores
//
// [OpenRestore] takes the lock for a restore and returns a
// [RestoreSession]. Before the restore renames any file,
// [RestoreSession.Record] writes the leaf's restore-intent record, and
// from then on [OpenWrite] refuses the leaf until the restore removes
// the record. The record is in one of four states, each owing one
// step: swap ([StepSwap]), library ([StepLibrary]), reconcile
// ([StepReconcile]), and reconcile-pending ([StepServe]).
//
// A resume takes up at the step its record owes. It judges the store
// in db/ against this binary rather than the version the record
// carries: a store below the binary's version is left
// reconcile-pending, the one state [OpenServe] accepts, and serve
// finishes it by migrating the store and writing the report
// [ServeOptions.Reconcile] returns.
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
