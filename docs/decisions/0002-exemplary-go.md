# 2. Exemplary Go

## Status

Accepted

## Context

Geoff set the bar on 2026-09-02, when the greenfield build began (commit `cb3a5df`): the code is
part of the product, and a correct but awkward implementation misses the bar the way an unclean UI
does.

## Decision

dubplate's Go meets the bar of idiomatic, architecturally beautiful Go: standard-library style and
clean architecture, never mere correctness. Every Go change under `cmd/`, `internal/`, and `tools/`
is held to it.

## Consequences

A reviewer may return `fix` on code that works but is awkward, and every dispatch that reviews Go
names this bar as an acceptance dimension.

golangci-lint (`.golangci.yml`) holds the floor with errcheck, govet, ineffassign, staticcheck,
unused, modernize, misspell, and unparam. Four checks added with this record hold more of it: dupl
(threshold 150) for copied blocks, gocognit (min-complexity 30) for functions too complex to read,
forbidigo for the clock and `exec.Command`, and the `deadcode -test` leg of `scripts/check.sh` for
code nothing reaches. The 12 functions over the complexity bar when gocognit was adopted are named
exclusions in `.golangci.yml`, and that list only shrinks.
