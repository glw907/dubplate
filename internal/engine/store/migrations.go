package store

import (
	"embed"
	"io/fs"
)

// migrationFiles embeds every migration under migrations/, in the
// binary. readMigrations reads them through migrationSource, which
// roots the tree at the files themselves rather than the
// "migrations" directory they live in on disk.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrationSource returns the embedded migration set as an fs.FS whose
// root holds the files migrate expects, "<version>_<label>.sql" each.
func migrationSource() fs.FS {
	sub, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		// fs.Sub only fails on a malformed directory argument, and
		// "migrations" is a literal matching the embed directive
		// above: this cannot fail at runtime.
		panic("store: migrationSource: " + err.Error())
	}
	return sub
}
