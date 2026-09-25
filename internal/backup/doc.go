// Package backup runs one backup of a library and its store to an
// rclone remote.
//
// A run syncs the library tree to library/ under the remote's path,
// moving every file the sync would replace or delete into
// archive/<stamp>/, and then sends a verified copy of the store to
// snapshots/<stamp>.db. Each destination joins the remote's path with
// path.Join, so a remote named name: gives name:library. Every run
// takes one stamp, at its start, in the store's stamp layout.
//
// A sync treats its source as the new truth, so a library mount that
// never came up, or a tree emptied by accident, would archive every
// file on the remote. Guards against the store's catalog run before
// rclone ever starts, and the sync runs with a bound on how many files
// it may delete.
//
// rclone runs in its own process group, so a terminal's interrupt
// reaches the caller alone, and inherits the backup kind lease and the
// store lock, so a transfer that outlives its caller still keeps a
// second backup out until it exits.
package backup
