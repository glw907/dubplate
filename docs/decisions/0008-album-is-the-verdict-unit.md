# 8. The album is the verdict unit

## Status

Accepted

## Context

A library that admits half an album leaves members with gaps nothing reports. The rule dates from
the start of the dubplate era (commit `006f2d4`, 2026-08-31).

## Decision

The album is the verdict unit for an upload under `internal/intake`: one failing track rejects the
whole batch, and a partial album is never admitted. Separate tracks are the other kind: each file
gets its own verdict, and the passing ones are accepted.

## Consequences

Review holds the rule against `internal/intake` and the gate under it.
