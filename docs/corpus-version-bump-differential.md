# Version-bump differential job

Standing operational doc, alongside `docs/server-runbook.md`. Self-contained:
whoever runs this job, in a year or in an afternoon, may be neither Geoff nor
an agent that has read the rung 2 plans. It describes a job, not a script that
exists yet; nothing in this file runs anything, and running it issues no live
network request on its own (see "No network in this doc" below).

## What the job does and what it protects against

The differential job drives pinned beets as a subprocess against an estate
config and a local stub (`internal/harness/stub`), the same invocation shape
`oracle.Derive` already uses on the frozen corpus. Its purpose is to catch a
class of regression the frozen-corpus gate cannot see on its own: a new beets
version that breaks the **invocation contract**—the config keys it accepts,
which plugins it enables, how its CLI behaves—independently of any change to
a matching verdict. A verdict change is expected and adjudicated (the corpus
exists to catch it); a broken invocation contract is not a verdict change at
all, it is the harness failing to talk to beets correctly, and the frozen-verdict
gate alone cannot tell the two apart.

beets 3.0.0 is the concrete instance this job exists for. It is beets' one
active compatibility-shedding plan: the legacy plugin layer is TODO-marked for
removal, and `docs/dev/plugins/autotagger.rst` promises removal in v3.0.0
(finding U7, `docs/research/2026-09-02-beets-evolution-lessons.md:396-412`).
When that release lands, or any beets major version after it, this is the job
that finds out whether dubplate's harness still knows how to invoke beets
before anything downstream (a corpus re-derivation, an adjudication pass)
depends on it.

## Falsifiable success condition

The job's premise is that everything the harness knows about invoking beets
lives behind one adapter package. State the premise so it can be checked
against, not just believed: **when the differential is pointed at a beets
3.x, only `internal/harness/oracle` should need edits**—its Go driver files
and its Python shims (`derive_oracle.py`, `derive_ties.py`,
`derive_component.py`, and any sibling shim the package gains later)—and
nothing outside that package. If a beets bump forces a change anywhere else
(`internal/harness/cassette`, `internal/harness/analysis`, `corpus/`
fixtures themselves), the premise this job is built on has failed, and that
failure is itself the finding to report: the isolation the harness rung
promised did not hold, not merely that beets changed.

## Why the blast radius is scoped to scoring, not response shape

A beets bump is expected to move scoring and verdicts, and is not expected to
move the shape of a captured MusicBrainz response. That asymmetry is
deliberate, not incidental (finding U6,
`beets-evolution-lessons.md:387-395`): the capture tool records the maximal
practical `inc=` response body for a release, not the minimal set of fields
today's matcher happens to read. A matcher change that starts reading a field
the corpus never captured would be a real gap; a matcher change that reads a
field the corpus already captured, just previously unused, costs nothing. The
committed `responses/*.json` files are frozen precisely so the differential
job's job is comparing beets' *decisions* across a version bump, never
re-fetching or reshaping the data beets decides over.

## Wall-time budget

The formula is the client's hard 0.25-second inter-request floor, in the
shape ruling M8 states it: **1.04s + 0.25s x (requests - 1)** per case, never
a flat per-case figure—beets pays a fixed per-request floor on every
MusicBrainz call the case's request log replays, not a fixed cost per case.

A full run pays this floor twice, because `DeriveTies` starts its own stub
and its own Python process (`internal/harness/oracle/ties.go:65` and `:101`)
entirely independently of `Derive`'s own stub
(`internal/harness/oracle/oracle.go:163`), replaying every case's requests a
second time from a cold state. The two legs are priced separately below and
then summed, matching the derive-report shape every task in this rung already
reports against ("corpus derive, ties derive, component derive" broken out,
never a bare "green").

Figures below are read off the committed tree at execution time (2026-09-04),
not estimated:

| leg | cases | total requests | formula | wall time |
| --- | --- | --- | --- | --- |
| corpus (`oracle.Derive`) | 215 | 351 | 215 x 1.04s + 0.25s x (351 - 215) = 223.6s + 34.0s | 257.6s (~4.3 min) |
| ties (`DeriveTies`) | 215 | 351 | same formula, same request replay, its own stub | 257.6s (~4.3 min) |
| **corpus + ties total** | | | | **515.2s (~8.6 min)** |
| file-tags (`derive_file_tags.py`) | 12 subjects | 0 (no network) | no inter-request floor applies (no MusicBrainz call at all) | 0.55s-0.59s measured (two independent `GOFLAGS=-count=1` runs, pass 2d's Task 4, 2026-09-04) |

The file-tags leg issues no MusicBrainz request at all (twelve subjects, one
Python process, reading already-committed `Expected.AppliedFields` off the
corpus rather than deriving anything from a stub), so no inter-request floor
applies to it and its cost does not belong in the table above by formula. Its
measured wall time above (Task 4's own two clean-tree derive runs) confirms
the prediction: negligible next to the corpus and ties legs, never the
bottleneck a re-derivation needs to budget for.
Re-run the corpus and requests counts (`ls corpus/cases | wc -l`; sum each
case's `requests` array length) before trusting this table again—the
committed corpus grows across rungs, and a stale count here silently
under-prices the next differential run.

**The corpus and ties rows above are the M8 formula's floor, not a
measurement, and pass 2d's own close measured the real thing.** Two
independent uncached (`GOFLAGS=-count=1`) full `make derive` runs against
the committed 215-case corpus, at pass 2d's Task 4 close (2026-09-04),
landed at 980s and 974s total, both well above this table's
257.6s-per-leg formula floor: corpus derive measured 462.94s and 541.20s
against the floor's 257.6s, ties derive 503.90s and 420.06s against the
same 257.6s. The formula prices the client's own inter-request wait, which
is a floor on wall time, not a prediction of it; real per-request Python,
beets, and MusicBrainz-response-parsing overhead sits on top of that floor
and roughly doubled it here. A version-bump differential run should budget
against the measured figure, not the formula table above, which exists to
show the floor is dominated by network-request pacing rather than by case
count alone.

## No network in this doc

This document describes a job that drives a subprocess against MusicBrainz
data; it is not the job. Nothing above was run to write this file, no
`/ws/2` request was issued, and no `harness capture` invocation happened here.
The corpus this file prices is 2c's already-committed fixture set, read off
the tree, never re-fetched.
