// Package backup backs up a library and its store to an rclone remote,
// and restores them from it.
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
//
// A restore stages and verifies the incoming store, then writes the
// store's restore-intent record before it renames any file. The record
// names the step still owed, and a resume picks up at that step from
// the source and library root the record names, never fetching the
// store again; a resumed library step still runs rclone copy against
// the remote's library/ once more. A fresh library restore refuses a
// root that holds files unless run with --into-nonempty, and one run
// that way, or a resume, moves each local file its copy would replace
// into a sibling displaced directory and never deletes one. The store
// a restore replaces is kept as a pre-restore copy in snapshots/; with
// --discard-current, a current store that cannot even be opened moves
// to a discarded copy there instead. Neither copy is ever pruned: each
// stays until the operator removes it by hand.
package backup
