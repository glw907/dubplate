# 1. Record architecture decisions

## Status

Accepted

## Context

dubplate's binding rules lived in prose in `CLAUDE.md` and `docs/standing-rules.md`, where a rule
and a later change to it sat side by side with no status to say which held. Agents build most of
the code, and an agent needs to know which rules bind it and which are history. The lean project
pattern (`lean-project-pattern/specs/2026-11-23-lean-project-pattern.md` in `glw907/initiatives`,
approved by Geoff on 2026-11-23) adopts Michael Nygard's decision records for this.

## Decision

Each binding decision is one record in `docs/decisions/`, numbered from 0001, with a title and the
sections `## Status`, `## Context`, `## Decision`, and `## Consequences`, and no other field.

The status is one of `Proposed`, `Accepted`, `Deprecated`, or `Superseded by NNNN`, where NNNN
names a record in the folder. An agent treats an accepted record as binding, a proposed one as
context, and a superseded or deprecated one as history.

A change of mind is a new record that supersedes the old one. The old record's text is never
edited; only its status line changes. A record that governs particular packages or paths names
them in its decision, and its context cites where the rule came from.

## Consequences

Before editing, an agent lists `docs/decisions/` and reads each accepted record whose title touches
the work. A rule lives in one record, and `CLAUDE.md` and the `dubplate-build` skill point at it by
number instead of restating it.

The gate's "decision record status" leg in `scripts/check.sh` fails a record with no status, a
status outside the four, or a `Superseded by` that names no record in the folder.
