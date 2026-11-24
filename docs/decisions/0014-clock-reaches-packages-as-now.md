# 14. The clock reaches packages as Now

## Status

Accepted

## Context

A package that reads the clock itself can't be tested at a chosen time, and its tests sleep or
race. The rule landed with work `structure-keys` (commit `8e92787`, 2026-11-20).

## Decision

The time reaches a package as a `Now` function in its configuration. No package under `internal`
reads the clock, apart from the test harness under `internal/harness`. A timeout is a deadline on
the request's context.

## Consequences

forbidigo in `.golangci.yml` fails a call to `time.Now`, `time.Since`, or `time.Until` outside a
test file, `cmd/`, `tools/`, and `internal/harness/`.
