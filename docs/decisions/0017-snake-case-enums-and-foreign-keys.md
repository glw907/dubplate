# 17. Snake_case enums, and foreign keys named for their table

## Status

Accepted

## Context

The server spelled its enum values two ways. Reject codes, submission states, settings sources,
backup failure codes, and the command-line identity channel were kebab-case (`bad-extension`,
`check-unfinished`, `env-seed`), while JSON field names, apply modes, and sensitivities were
snake_case (`next_step`, `next_operation`, `not_sensitive`). A TypeScript client had to quote the
kebab keys, and a reader could not guess a value's spelling from its neighbours. Foreign keys to
`users` were named three ways as well: `user_id` in most tables, `owner` in `collections`, and
`uploader` in `submissions`. The architecture read of 2026-11-23
(`docs/research/2026-11-23-architecture-read.md`) found both, and work `wire-contract` settles them.

## Decision

Every enum value the server sends on the wire or writes to the store is lowercase snake_case, such
as `bad_extension`, `check_unfinished`, `env_seed`, and `rclone_exit_7`. A value built from an
outside name lowercases it, so a signal code is `rclone_signal_kill`.

A foreign key column is named for the table it references, in the singular with `_id`: a column
that references `users` is `user_id`.

These are not enum values and keep their spelling: route names and URL paths, file names, token and
record versions, setting keys, and command-line flags.

Migration 0007 rewrites every stored value and renames the two columns. A restore-intent record
written before it may still read `reconcile-pending`, and the decoder accepts that one retired
spelling.

## Consequences

`TestEveryReasonCarriesItsFields` in `internal/reject/reject_test.go` fails on a reject code that is
not snake_case, and the settings table's `CHECK` refuses a source outside `env_seed` and `admin`.
Review holds the rest of the rule, including the foreign key names.
