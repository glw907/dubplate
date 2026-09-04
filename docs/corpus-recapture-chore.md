# Corpus re-capture chore

Standing operational doc, alongside `docs/bringup-runbook.md`. Self-contained:
whoever runs this chore, in a year or in an afternoon, may be neither Geoff nor
an agent that has read the rung 2 plans. It describes a chore, not a script
that exists yet; nothing in this file runs anything, and reading it issues no
live network request on its own.

## Framing: additive recall, not a rewrite

The corpus's frozen `responses/*.json` files are not a museum piece to
preserve untouched forever. The highest-leverage open deficiencies in beets
today widen candidate retrieval rather than change scoring: seeding
candidates from embedded MBIDs, release-group lookup, better search queries
(finding U8, `docs/superpowers/research/2026-09-02-beets-evolution-lessons.md:413-430`).
An improvement of that shape adds candidates and leaves the beets-equivalent
subset's scores untouched, which is exactly why the fixture format admits
recall-extension cases from day one (`Case.Extra`, `corpus/README.md`'s own
layout section): a case can carry fixtures for lookups beets never issued,
graded on the shared candidate subset. This is why a post-launch retrieval
improvement is graded by the same harness the corpus already runs, rather
than escaping it into some separate, one-off proof. Re-capture is how new
recall surface, and any response-shape drift MusicBrainz itself introduces,
gets folded into that same harness instead of accumulating outside it.

## When this chore runs

Name the triggers; "quarterly, or sooner if things change" is exactly the
soft guidance this repo's gate philosophy rejects. Any one of the following
starts the chore, whichever comes first:

- **A beets point release.** Even a point release can move a fixture-visible
  detail (a plugin default, a config key's accepted values); a point release
  is lower risk than the version-bump differential job's own major-version
  concern (`docs/corpus-version-bump-differential.md`), but it is still a
  trigger, not a skip.
- **A `mediafile` bump.** Pass 2d pinned `mediafile` as the corpus's sixth
  reconciled component (`corpus/manifest.json`'s `mediafile` field,
  `internal/harness/oracle/reconcile.go`'s `pinned` tuple); a `mediafile`
  release can move which tags a real write actually lands, independent of
  any beets change.
- **A MusicBrainz schema change noticed as response-shape drift.** If a
  captured response body's shape stops matching what a fresh capture of the
  same release would return today (a new field, a renamed relation, a
  changed `inc=` default), that drift is itself a trigger, discovered
  incidentally rather than scheduled.
- **The calendar quarter.** Quarterly is the floor, not the ceiling: if none
  of the three triggers above has fired in a quarter, the chore still runs.

## Procedure: `Reconcile`'s own contract

The chore's procedure is not a separate design; it is `Reconcile`
(`internal/harness/oracle/reconcile.go`) run in anger, against a venv that
may have drifted, with the reconciliation-time question restated as a
re-capture-time one. `Reconcile`'s own doc comment states why this matters
in the first place:

> unidecode's transliteration tables, jellyfish's edit distance, and lap's
> assignment solver all sit directly in the distance path, so a silent bump
> there would change verdicts with nothing in the corpus to show for it.
> (`reconcile.go:15-20`)

Pass 2d extended the pinned set from five components to six: `beets`,
`unidecode`, `jellyfish`, `lap`, `numpy`, and now `mediafile`
(`corpus/manifest.json`; `internal/harness/oracle/reconcile.go`'s `pinned`
tuple and Python `versionProbe`, both). A re-capture event re-pins all six,
not just whichever one triggered it: run `Reconcile` against the venv first,
record whatever disagrees, and decide whether the venv's newer versions are
adopted (new pins written into `corpus/manifest.json`) or held back (the
venv re-pinned to the old versions) before doing anything else. Only once
the pins are settled does a re-capture proceed: re-run `harness capture`
against every case whose fixture needs refreshing (a subset, chosen by
whichever trigger fired, never blanket unless the trigger was itself
blanket), then `make derive` to confirm the new pins and responses still
derive to zero misses.

**Any disagreement `Reconcile` reports is the corpus-revalidation event the
matcher mechanics study already names as the standing plan**
(`docs/superpowers/research/2026-09-02-beets-matcher-mechanics.md:602-606`:
"Library-version pins are corpus metadata... a beets bump is a
corpus-revalidation event (already the standing plan)"). A version
disagreement is not a chore to defer; it is the chore already having started,
whether or not the calendar trigger has fired yet.

## Cost: see the differential doc

The differential doc (`docs/corpus-version-bump-differential.md`) prices a
full re-derivation of the committed corpus, corpus leg and ties leg broken
out and summed, against the client's 0.25-second inter-request floor. A
full re-capture pays that same shape scaled to whichever subset of cases is
being refreshed; a full-corpus re-capture pays the differential doc's own
total, unchanged. Read the number there rather than recomputing it here:
one number, one home, so the two docs cannot drift apart the way two
independently maintained copies of the same arithmetic always eventually do.

## No network in this doc

This document describes a chore that drives live capture against
MusicBrainz; it is not the chore. Nothing above was run to write this file,
no `/ws/2` request was issued, and no `harness capture` invocation happened
here.
