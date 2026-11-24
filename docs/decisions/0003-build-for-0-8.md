# 3. Build for 0.8

## Status

Accepted

## Context

Geoff ruled on 2026-09-07, re-keyed the version on 2026-09-25, and re-planned it on 2026-10-05.
He refined it on 2026-09-12: a forward hook is built by the task that first consumes it. The
streamer is the built-in OpenSubsonic streamer, slices 1 to 8 of
`docs/specs/2026-09-07-streamer-2-0-design.md` (the spec still says 2.0). A store schema built
without 0.8 in view is rework, and code exported ahead of its caller reads as justified in the diff
that added it and as dead weight in the package.

## Decision

Every spec weighs its surface against the streamer spec's forward-hooks list and states, per task,
whether the task lands a hook or is neutral. A newly found hook joins that list. The list's numbers
are stable: a retired hook stays in place, marked retired. A hook is named in the spec of the task
that first consumes it and built there. No 0.8 feature is built early.

The same spec carries a prior-art list: the OpenSubsonic specification, Navidrome and gonic (for
behavior only, under the GPL), taglib via wazero, modernc sqlite, goose, mewkiz/flac, Picard's tag
tables, and chromaprint. A spec names the entries it consulted and adopts or declines each with a
reason, so no task rebuilds what a license-compatible library already does well. An adoption is
judged on license, code quality against record 0002, maintenance (release cadence, open-issue age,
bus factor), and adoption (importers, who depends on it), with the evidence in the spec.

Nothing is exported ahead of a caller, and a caller is production code. Every new export names its
production caller in the task that adds it.

## Consequences

The `deadcode -test` leg of `scripts/check.sh` holds the floor: it fails when no code or test
reaches a function. A test counts as a caller there, so review holds the rest of the rule. The
`deadcode` warning leg, which counts only a `main` as a caller, reports today's exports that only
tests reach.
