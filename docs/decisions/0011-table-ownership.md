# 11. Table ownership

## Status

Accepted

## Context

Two packages writing one table each hold half of its invariants, and neither can keep them. The
rule landed with work `structure-keys` (commit `8e92787`, 2026-11-20), after work `structure-base`
named the owner of every table.

## Decision

A table gets a migration file in `internal/store` and a line in the store's ownership list. Only
the owning package writes it. Another package calls the owner's function, and passes its
transaction when both changes must commit together.

## Consequences

`TestTableOwnersNameEveryTable` and `TestTablesAreWrittenOnlyByTheirOwner` in
`internal/store/owners_test.go` fail when a table has no owner or a package writes a table it does
not own.
