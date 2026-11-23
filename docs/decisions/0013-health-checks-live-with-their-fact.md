# 13. Health checks live with their fact

## Status

Accepted

## Context

The `doctor` command once carried every health check itself, so it imported each package's
internals. Work `structure-doctor` (PR #25, merged 2026-11-23) moved the eleven checks into the
packages that know each fact, and the rule landed at commit `12bab17`.

## Decision

A health check lives in the package that knows the fact, in that package's `doctor.go`. The
`doctor` command in `cmd/dubplate-server` gains one line that places the row.

## Consequences

Record 0010 lets `doctor.go` import configuration beside `config.go`, and its import test holds
that. Review holds the rest.
