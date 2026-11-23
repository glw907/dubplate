# 12. One id type

## Status

Accepted

## Context

Ids parsed by hand in each package drift in what they accept. Work `structure-base` gave ids one
parser in `internal/id`, and the rule landed in `CLAUDE.md` at commit `8e92787` (2026-11-20).

## Decision

An id is the type in `internal/id`, or a named type built on it. Only `id.Parse` reads one from
text.

## Consequences

Review holds the rule; no test checks it.
