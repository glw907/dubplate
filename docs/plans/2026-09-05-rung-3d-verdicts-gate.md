# Rung 3, Pass 3d (Verdicts, the Assignment Solve, Admission, Recommendation, and the Whole-Corpus Conformance Gate) Implementation Plan

**Revision 2.1, 2026-09-05 (conductor rulings on the fold, all three taken as recommended):** B2, `ModeDefault` re-scores each case's committed candidate set under the corrected table and issues no search, with the request-set rule `ModeCompat`-only (option a); m16, the entry point keeps the package name `internal/engine/tag`, both doc comments name the `tagio` pairing, and the build spec's tree listing joins Task 9's drafted corrections; m15, the `docs/` ledgers row is written. All four contested-surface amendments (`internal/engine/tag`, `internal/harness/oracle`, `corpus/analysis/*.json`, `docs/` ledgers) landed in the parallel-approach spec before launch, so Task 0 item 4 should now find nine of nine present. Every "provisional, conductor ruling owed" mark below reads as ruled.

**Revision 2, 2026-09-05:** folds the two adversarial plan reviews,
merged in `research/2026-09-05-rung-3d-plan-review.md`: blockers B1 to
B4, majors M-a through M-i, minors m1 to m20. Three reviewer citation
claims are rejected with the evidence in that document's header, and
three items are conductor-owed rather than folded silently; each of
those three is applied provisionally here and marked inline.

The load-bearing resolutions. **Every `lap.lapjv` vector this plan
quoted was the second return value** (B1): `x` is row-to-column and `y`
is column-to-row, they are inverse permutations that differ for every
n >= 3, and `solveLAP`'s own contract is `y`, which is what beets
consumes at `match.py:66`. Header figure 4, Task 0 item 11 and R3's
evidence paragraph now quote `y` and name `x` once as the inverse.
**`ModeDefault` can no longer re-run `tag.Album`** (B2): all ten
`defect_free` probe cases log one lookup and zero searches and score
`medium` under the corrected table, so re-running the flow loses the
strong-match early return, issues a search the case never logged, and
fails on the runner's own zero-miss rule before any `defect_free`
comparison happens. The request-set rule becomes `ModeCompat`-only and
`ModeDefault` re-scores each case's own committed candidate set.
**Task 1's various-artists fixture asserted the swap where the identity
is the unique optimum** (B3), so an implementer building to it would
have been driven into the exact defect it exists to catch.
**`AssignmentFloor` reds the gate on ruling F's escalation path** (B4),
which is the one path the pass's process design exists to keep green;
the assertion now subtracts `TiesAwaitingRuling` and `TiesExcluded` the
way the graded-field table and Task 8 already state it.

Five majors move real behavior. `policyFor`'s fixture table asserted
`GradeAlbumDistance: false` on every row through struct equality and
omitted the spec's `ties` unique row, so a resolver built to satisfy it
would have shipped album-distance grading off corpus-wide (M-a).
`mergeProfile` never applied the three `import.*` keys the committed
profiles set, so R12's honesty premise and `gradeTunables`' doc comment
were false as landed, with `none_rec_action` the one key whose value
actually differs (M-b). Task 6's five tie rows use a field the landed
table does not have, and the existing `dispositioned` fixture is an
unnamed seventh site the new rule breaks (M-c). One shared `ParseAction`
would have blessed `quiet_fallback: apply`, which confuse refuses
(M-d). Task 5's mutation 1 could not fire its own named test and its
corpus count is 39 rather than 48 (M-e). Task 4's mutations 1 and 6
claimed corpus detectors that do not exist (M-f, M-g), the 503 row spent
31 seconds of real wall time per gate run (M-h), and `casePolicy`'s
`CompatOn bool` could not express two independent per-class switches
(M-i).

Three folds take an adaptation rather than the literal fix, each named
where it lands. **B2** takes the re-score design rather than review 1's
first option of relaxing the request-set rule and pointing the client at
a differently configured stub, because re-scoring the committed
candidate set reproduces the ten `defect_free` figures exactly and costs
no second flow run; the medium rung of the ladder is decided by the
minimum distance alone (`match.py:288-290`), so a candidate set the
search would have widened cannot move a `medium` verdict off `medium`.
**m16** keeps `internal/engine/tag` and records the `tagio` pairing at
R18 rather than renaming the package, because `tag_album` is the beets
name the package holds and rung 8's `tagio` is FLAC I/O; the build
spec's tree listing joins Task 9's drafted corrections instead.
**m2** reconciles R19's case count to Task 9's five-case list and marks
the 17-tied figure as carried from the inventory and re-measured at Task
7, rather than asserting a number neither review could reproduce without
implementing the solver.

**Conductor rulings owed, all three applied provisionally and marked
inline:** B2's `ModeDefault` design (re-score against the committed
candidate set, versus re-running the flow with a relaxed request-set
rule and a search-404-to-empty mapping, versus dropping `VerdictMoves`);
m16's package name against the build spec's `tagio`; and m15's missing
`docs/` ledgers row in the contested-surface table, which is a fourth
conductor amendment or a standing-rule exemption.

**The merged review's 50-item "Verified sound" list is a second
do-not-re-litigate record** for this pass, alongside the rung 3 spec
review's twenty and the 3c plan review's forty-one. Both lenses
independently reproduced the `max_rec` census, the `lap` zero-pad
extension model, the fourth pattern table's byte identity, B3's
fires-on-nothing premise, and every one of contracts V1 through V20.

**Revision 1, 2026-09-05.** Drafted against the rung 3 spec at **revision
2** (`docs/specs/2026-09-03-rung-3-go-matcher.md`), ratified
on its six rulings and review-folded on all nineteen findings of
`research/2026-09-04-rung-3-spec-review.md`. That review's "Verified
sound" list is the do-not-re-litigate record: its twenty items are cited
here rather than re-derived, and an implementer who believes one of them
is wrong reports a finding and implements the record anyway. The same
standing applies to `docs/HISTORY.md`'s 3a, 3b and 3c entries, whose ten,
seventeen and sixteen rediscover-guards are cited below rather than
re-measured, and to `research/2026-09-04-rung-3c-plan-review.md`'s
41-item "Verified sound" list.

This plan's primary input is
`docs/research/2026-09-05-rung-3d-planning-inventory.md`, a
research report that read pinned beets 2.13.1 in `tools/venv/.venv`, the
harness and corpus as landed at `c7a85fd`, and the engine as 3a, 3b and
3c left it. Every mechanic below is cited to the tree, and the
inventory's twenty deduplicated open questions are **resolved in this
plan**, each with its evidence and its rejected alternative, in
"Resolutions this plan takes rather than discovers".

**Five inventory figures are settled here, off the tree at HEAD
`c7a85fd`, before any task inherits them.**

1. **The `max_rec` downgrade fires on exactly six cases, all
   strong to medium.** The inventory flagged this as an unreconciled
   disagreement between its own sections (a) and (c), 6 against 21.
   Measured directly by replaying `_recommendation`'s ladder and its
   downgrade over all 215 committed `expected` records: pre-downgrade
   branch census strong 173, medium 33, single-candidate low 3, gap low
   1, no-conclusion 4, no-candidates 1; the downgrade fires on
   `assign-extra-item`, `complete-combined-short-scrambled`,
   `complete-max-rec-boundary`, `complete-short-one`,
   `real-noise-pink-floyd-de1994` and
   `rec-max-rec-missing-tracks-downgrade`, and the model reproduces all
   215 recorded recommendations with zero mismatches. **Section (a)
   wins; section (c)'s "21" is the losing sentence**, and it is (b) §4's
   "21 cases exercise candidate ordering and the gap threshold" read
   into the wrong paragraph.
2. **No committed candidate carries a `max_rec` key at 0.0.** Zero of
   294 candidates carry `missing_tracks` or `unmatched_tracks` at
   exactly 0.0, so the added-key-versus-nonzero-view distinction is
   invisible to the corpus for the two keys that drive a downgrade.
   This is 3d's declared gap 5 below, and it confirms (b) §5 item 4.
3. **The fourth pattern table is safe to build.** Across the six
   `adv-ft-*` probe cases, **zero** compared strings carry a bracket or
   parenthesis; across the four `adv-remix-*` cases, **zero** carry a
   featuring trigger. The field set is stated rather than a bare
   denominator, because the denominator moves with it and three
   independent measurements at this plan's drafting and at both review
   lenses returned 234/156, 216/144 and 258/172 while all three returned
   zero matches (revision 2, fold m18): the strings compared are each
   case's `input.items` `title`, `album` and `artist`, plus every
   `title`, `name`, `artist`, `albumartist`, `sort-name` and
   `disambiguation` string in its response bodies. Measured
   mechanically, counts only, no body read into context. So a combined
   guarded-and-bracket-free table reproduces both sets of committed
   `defect_free` figures, and Resolution R2's fourth table is not a
   gamble.
4. **`lap` 0.5.13's pivot order and the rectangular extension model
   reproduce.** `lap.lapjv` returns three values and the port needs the
   **third**: `x[i]` is the column assigned to row `i`, `y[j]` is the
   row assigned to column `j`, and beets unpacks
   `_, _, assigned_item_idxs` (`match.py:66`), so `y` is what
   `solveLAP` reproduces. The two are inverse permutations and agree
   only at n = 2. On an all-zeros `n x n` matrix the pinned solver
   returns **column-to-row** `[1,0]`, `[1,2,0]`, `[1,2,3,0]`,
   `[1,3,2,4,0]`, `[1,4,3,2,5,0]`, `[1,10,9,8,7,6,5,4,3,2,11,0]` and
   `[1,12,11,10,9,8,7,6,5,4,3,2,13,0]` for n = 2,3,4,5,6,12,14. (The
   row-to-column `x` for the same shapes is `[1,0]`, `[2,0,1]`,
   `[3,0,1,2]`, `[4,0,2,1,3]`, `[5,0,3,2,1,4]`,
   `[11,0,9,8,7,6,5,4,3,2,1,10]`, `[13,0,11,10,9,8,7,6,5,4,3,2,1,12]`,
   named here once so a reader probing the wrong return value
   recognizes it rather than reporting a divergence; revision 2, fold
   B1.) The `n = max(rows, cols)` zero-pad model agreed with
   `extend_cost=True` on 262 of 262 random rectangular trials at this
   session's own seed, and both review lenses reproduced it
   independently at 1316 of 1316 and 256 of 256. The inventory's §10
   figures are confirmed rather than taken on trust.
5. **The corpus census the floors rest on**, re-read at `c7a85fd`: 215
   cases (178 estate, 36 stock, 1 locale); 294 candidates across 214
   candidate-bearing cases, 1 case with none; candidate counts 193 at
   one, 4 at two, 9 at five, 8 at six; 7 cases with a non-empty
   `extra_items` (max 33) and 14 with a non-empty `extra_tracks` (max
   4), none with both; 23 non-identity assignments; 2 cases carrying a
   candidate-distance tie; verdicts 167 strong/apply, 39 medium/skip, 4
   low/skip, 5 none/skip. `ties.json` 215 rows (213 with no
   disposition, 2 pinned, 0 excluded, 1 skipped), `proximity.json` 2 (1
   flagged, 1 skipped), `adjudications.json` 21 (18 `oracle-wrong`, 3
   `oracle-correct`, 0 pending, 10 carrying a `defect_free` block),
   `oracle_bugs.json` 6 (4 `oracle-wrong`, 2 `oracle-correct`, 0
   pending).

**Two spec premises have moved since the spec was written, and the plan
reads the tree rather than the page.** The B3 paragraph (spec `:300`)
names three probe cases with no adjudication row; today every probe case
of every `oracle-wrong` hypothesis resolves to a row by
`subject.case_id`, verified case by case, so **B3's new rule fires on
nothing in the committed corpus and its firing fixture must be
hand-authored**. Ruling A's "five hypotheses pending" (spec `:77`) is now
zero pending. The spec anticipated both ("The count is read off the tree
at execution time"), so these are plan-header restatements rather than
spec edits, and Task 0 re-confirms them at dispatch.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section: pass-execute chain per task, sequential. The conductor
> consumes reports and never reads diffs inline. Subagents never
> dispatch other agents; `code-simplifier` is the conductor's dispatch
> after a task's review accepts. The conductor also owns the ledgers
> (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`), the lane
> registry, the contested-surface rows, and the batch commit of anything
> a task drafts into them. Rung 3 is not an adversarial-cycle rung
> (those are 5, 12 and 16), so review scales to the in-chain
> `diff-reviewer` per task plus the pass-end gate, **with one named
> exception this pass takes: a fresh-context `claude-opus-5` review of
> the conformance runner alone, after Task 7 lands** (spec `:362`,
> Resolution R11). **This pass runs in the MAIN TREE**, and nothing here
> may be executed from a worktree: the spec's tree column for 3d is
> "Main tree: writes `internal/harness/conform` and
> `internal/harness/analysis`" (spec `:331`), and the standing policy is
> that corpus and harness work never moves to a worktree
> (`docs/specs/2026-09-03-parallel-build-approach.md`,
> lane-launch checklist item 5).

## Primary inputs

Read these before Task 1, in this order:

1. **`docs/research/2026-09-05-rung-3d-planning-inventory.md`,
   in full.** It is this plan's argument. Section (a) is beets' verdict
   mechanics and the assignment solver with line citations against
   `tools/venv/.venv`; (b) is the harness ledgers and tiers as landed;
   (c) is the engine as landed and the proposed Go surface; (d)
   reconciles all three and lists the twenty open questions. Do not
   re-derive what it settles; cite it and verify the line if you carry
   one forward. Its five corrected or confirmed figures are named in
   this plan's header, and its `max_rec` count of 21 is superseded by 6.
2. **`docs/specs/2026-09-03-rung-3-go-matcher.md`, revision
   2**, in full. The pass-split table's 3d row (`:331`) is this plan's
   scope statement. Rulings A (`:71-87`), B (`:89-95`), E (`:115-124`)
   and F (`:125-133`), the M4 mechanism (`:131-133`), the B3 probe-case
   rule (`:300`), the M8 corpus-unreachable surfaces (`:234`), "The
   conformance gate" in full (`:253-321`) with its disposition table
   (`:285-299`), the precedence rule (`:302`), the pending-divergence
   path (`:304-314`), the assertion discipline (`:316-321`), and the
   review note at `:362` are the sections this pass implements.
3. **`docs/HISTORY.md`'s "Rung 3, pass 3c" entry, in full**, then "pass
   3b", then "pass 3a". 3c's guards 3 (the R1 seam handing the request
   set to 3d), 4 (the stub's blind side), 5 (`Case.Extra` has no
   producer), 9 (the three regime divergences), 10 (`MergePseudoRelease`
   re-owned to 3d and rung 9) and 13c (`vaLikely` unexported) are
   load-bearing here, as are 3b's guards 8, 10 and 11 (the merge's seven
   fields, the three `SD_PATTERNS` tables, the `\b` regime divergence)
   and 3a's guards 2 (the deferred four-knob registration), 3 (the
   `SD_PATTERNS`-as-a-value seam) and 7 (the two `inclArtist` call
   shapes).
4. **`docs/plans/2026-09-04-rung-3c-client-limiter.md`**,
   for its twenty client and query contracts and its Global Constraints.
   This plan does not restate them; it adds the verdict contracts below.
5. The tree, read-only: `internal/engine/matcher/{info,likelies,penalty,
   score,track,track_artist,album_tracks,album_extras,album_media,
   album_country,tunables,strdist}.go`;
   `internal/engine/musicbrainz/{client,query,limiter,translate,pseudo}.go`;
   `internal/harness/analysis/{ties,oraclebug,adjudication,proximity}.go`
   and their tests; `internal/harness/cassette/cassette.go`;
   `internal/harness/stub/stub.go`;
   `internal/harness/conform/{doc,component,defaults,translate,request}.go`
   and `arch_test.go`; `internal/harness/corpus_test.go`;
   `internal/harness/oracle/ties.go` and `ties_derive_test.go` (both
   `//go:build oracle`).
6. `CLAUDE.md`'s Build process and Gate sections, `docs/STATUS.md`'s
   lane registry, and `corpus/README.md`'s dispositions and
   recorded-fixed-point sections.
7. **The pinned venv itself**,
   `tools/venv/.venv/lib/python3.13/site-packages/`, abbreviated `$SP`
   below, for `beets/autotag/match.py`, `beets/autotag/distance.py`,
   `beets/autotag/hooks.py`, `beets/ui/commands/import_/session.py`,
   `beets/importer/actions.py`, `beets/metadata_plugins.py`,
   `beetsplug/musicbrainz.py`, and the `lap` package. Every contract
   below cites it. Short offline probes with `tools/venv/.venv/bin/python`
   are permitted and expected. **`make derive` does not run in this pass
   at all** (Resolution R17).

## Goal

Ship the verdict half of the Go matcher and the instrument that grades
the whole of it.

Land `internal/engine/matcher/assign.go` (the cost matrix, the
Jonker-Volgenant solve, `lapjv`'s tie behavior on the two pinned cases,
the extras and their sort orders), `candidate.go` (the five admission
gates in beets' order, insertion order under the `(data_source,
album_id)` identity, and the stable distance sort), `verdict.go`
(`Recommendation`, the three thresholds with their strict-versus-
inclusive asymmetry, the `max_rec` downgrade read against the nonzero
view, and `_summary_judgment`'s verdict-to-action mapping),
`compat.go` (the two ledgered string-heuristic switches and the fourth
pattern table both switching off requires), and
`internal/engine/tag/tag.go` (the public entry point: `tag_album`'s
interleaved fetch-score-return flow, the ID-match early return, and the
search fallback).

Land the harness half: `TieEntry.Adjudicator` and its excluded-only
sentinel check (M4), `CheckOracleBugs`' probe-case rule and the
signature change it needs (B3), `stub.Server.Unused()` closing 3c's
blind-side carry, and `internal/harness/conform`'s whole-corpus
conformance runner with its tier 1 and tier 2 gates, its two grading
modes, its per-disposition policy, and its pending-divergence path.

Nothing here reads a tag, writes a file, or reaches the network. The
scoring core is 3a's and is shipped; the translation layer is 3b's and
is shipped; the client, the query builder and the limiter are 3c's and
are shipped. Rung 3 closes here, and the rung 3 to rung 4 handoff is
this pass's close under ruling B option (b).

## Owner touchpoints: two, both recorded rather than blocking

3d expects no attended sitting. Two things can arise that only Geoff can
finish, and **both are recorded and carried rather than stopping the
pass**.

1. **Ruling F's escalation** (spec `:125-133`, Resolution R3). Two
   failed tie-mirroring attempts stop **Task 2 and only Task 2**. The
   implementer writes the finding to
   `docs/research/2026-09-05-rung-3d-tie-mirroring-finding.md`,
   the two pinned rows are marked awaiting Geoff, the pass continues
   through every other task, and **the conductor never rules the
   disposition**. `disposition` is owner-only, enforced by the sentinel
   this very pass extends to `ties.json`.
2. **An unadjudicated tier-2 divergence.** Today there are none and
   `adjudications.json` carries zero pending rows. If the Go matcher
   diverges on a verdict whose resolution is an owner question, the
   runner reports it, Task 9 records it in the pass report and writes it
   to `docs/research/2026-09-05-rung-3d-pending-divergences.md`,
   and the launch gate refuses it. The runner never writes a
   disposition and never edits `corpus/analysis/*` (Resolution R15).

**The pass can close with a pending divergence outstanding.** The spec
says so in the adjudication ledger's own words, quoted at spec `:306`:
pending "is legal at every gate this package runs (`bash
scripts/check.sh`); the launch gate, a different gate at a different
time, refuses it", and spec `:312` requires only that the count be
reported and any growth named with the case and the suspicion. Rung 4
inherits both under ruling B option (b).

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition below resolves to the conductor. Only a conductor
decision to escalate reaches Geoff. Under the 2026-09-05 grant recorded
at `docs/STATUS.md:23-28`, the 80%-of-ceiling flag routes to the
conductor through 2026-09-07 09:00; CLAUDE.md's rule that it routes to
Geoff stands otherwise.

Three conductor edits are owed, all contested-surface table rows rather
than rulings; see "Contested surfaces" below. Task 0 reports them, the
conductor writes them, and none is a launch blocker.

## Spec

`docs/specs/2026-09-03-rung-3-go-matcher.md` (revision 2),
and behind it `specs/2026-09-02-greenfield-server-build.md` (revision
2.1, ratified), ladder entry 3. This pass completes deliverable 3, quoted
in full: "**Album-level matching**, with track assignment pinned to the
pinned-beets solver's tie behavior on corpus cases, stable candidate
ordering, and the admission filters" (spec `:61`), and the verdict-to-
action layer the spec places here rather than at rung 13 (`:236`).
Deliverables 1, 2, 4 and 5 are 3a, 3b, 3a-plus-3c and 3c, all shipped.

**The rung 15 forward criterion travels in this pass's report**, per the
spec's m7 fold (`:145-156`), so rung 15's plan reads it from the ledger
rather than rediscovering it:

> Rung 15 re-runs `internal/harness/conform`'s three tiers over the
> complete corpus, at the matcher commit as landed, and its acceptance is
> the launch bar's matcher clause read against that run: zero false
> confident matches over every stratum including stock-config, and zero
> unadjudicated beets-right/Go-wrong cases. Rung 3's own acceptance is
> provisional against the corpus at its then-current size. Any tier-2
> divergence carrying a pending row at rung 15 is the launch gate's
> business, not rung 15's to wave through.

## Token ceiling: 6M

As a **starting** figure, per the spec's Band and ceiling section
(`:360`), and **re-derived from Task 2's actual report**, meaning after
the tie-mirroring task resolves ruling F's unknown, before Task 3
dispatches.

The basis, read off HISTORY rather than asserted. 3a closed at ≈2.84M
across nine tasks with its prep uncounted. 3b closed at **≈4.82M against
6M with zero attended interactions**, of which 2,106K was the plan-prep
workflow and 2.7M was execution across nine tasks, so ≈270K per executed
task. 3c closed at **≈4.90M against 6M**: prep 2.55M, Task 0 0.11M, the
Tasks 1 to 7 chain 1.74M (≈250K per task), Task 5b 0.13M, the simplify
pass 0.15M, the final fresh-context review 0.07M, and the close ≈0.15M.

3d is nine tasks plus a pre-task: two engine tasks at `opus` (Tasks 1 and
2), three engine tasks at `sonnet` (Tasks 3, 4, 5), one harness ledger
task (Task 6), one runner task at `opus` (Task 7), one grading task (Task
8), and a close (Task 9). At 3c's measured rate that is ≈2.25M, plus
roughly one task's worth again for each of the three `opus` upshifts, so
≈3.0M for execution. Ruling F's two-attempt budget adds up to ≈0.4M if
both attempts run. The fresh-context runner review adds ≈0.15M, the
`code-simplifier` pass and its review ≈0.25M, and this close is larger
than 3c's because it carries the rung 3 to rung 4 handoff, the
declared-gap reconciliation and two coverage tables, so ≈0.30M. 3d's own
plan-prep workflow is bounded by 3c's 2.55M.

**Expected total ≈6.8M against a 6M starting ceiling.** The arithmetic
already exceeds the start, and the plan says so rather than shading the
estimate: the ceiling is the stop rule's yardstick, not a forecast, and
the re-derivation point exists precisely because Task 2 is the pass's
one genuinely unbounded task. On 2c's precedent (closed at ≈8.2M against
a re-derived 8M ceiling with the overrun approved, spec `:360`), the
likely re-derived figure is **8M**, and the conductor sets it on Task 2's
number rather than on this paragraph.

**Flag at 80% of the live ceiling (4.8M against 6M):** finish the task in
flight, write STATUS, and one combined question goes to the conductor
before continuing, per the standing grant recorded at
`docs/STATUS.md:23-28`. **That same STATUS bullet records a window spend
cap of 15M**, which is the other live constraint on this window and is
named here rather than half-quoted (fold m19). An expected total near
6.8M plus a prep bound of 2.55M leaves the cap comfortable, and a
re-derivation past 8M is the point at which the conductor checks it
rather than assuming it.

**Checkpoint: STATUS after every task.** That is 2b's, 2c's, 2d's, 3a's,
3b's and 3c's cadence; every task reports its own spend and its own gate
evidence, and the conductor sizes the next task against it.

**Phase 1's combined ceiling: 12M, and it stays a reservation rather than
a live sum.** It is 3d's 6M plus a 6M reservation for the second code
slot (lane A, rung 5, or lane T, rung 8, one at a time).
`docs/STATUS.md`'s lane registry records lane-2 as BLOCKED until Phase 0b
and 0c land, so no second sub-ceiling exists to sum today, and 12M is not
a figure any spend is currently measured against. The conductor replaces
the reservation with that lane's own assigned sub-ceiling at its launch.
This restates 3a's, 3b's and 3c's figure unchanged rather than
re-deriving it.

## Global Constraints

- **Main tree.** The spec's tree column for 3d is "Main tree: writes
  `internal/harness/conform` and `internal/harness/analysis`" (spec
  `:331`), and the standing policy is that corpus and harness work never
  moves to a worktree (parallel-approach spec, lane-launch checklist item
  5). The lane registry entry is written before the first dispatch, and
  **the live-executor check is that registry, never `pgrep -f
  Projects/dubplate`**, which matches every lane once a worktree exists.
- **Package territory, stated as a wall.** This pass writes
  **`internal/engine/matcher`** (`assign.go`, `candidate.go`,
  `verdict.go`, `compat.go`, the fourth `SD_PATTERNS` table and
  `Tunables.Validate` in `tunables.go`), **`internal/engine/tag`** (born
  this pass, one file plus its test, Resolution R18),
  **`internal/harness/analysis`** (`TieEntry.Adjudicator`, the
  `CheckOracleBugs` signature and rule, the exported subject index, and
  the untagged tie-disposition merge helper),
  **`internal/harness/stub`** (`Unused()` alone),
  **`internal/harness/conform`** (tier 1 and tier 2),
  **`internal/harness/oracle/ties_derive_test.go`** (two lines, both
  `//go:build oracle`), **`internal/harness/corpus_test.go`** (the
  `CheckOracleBugs` call site), and **`corpus/analysis/ties.json`** (one
  new field on two rows, Resolution R10), plus conductor-owned `docs/`
  ledgers. It writes **nothing** under `corpus/cases/`,
  `corpus/profiles/`, `corpus/manifest.json`, `internal/harness/{cassette,
  synth,mbrequest}`, `internal/engine/musicbrainz`, `internal/config`,
  `cmd/`, `tools/`, `web/`, or `_salvage/`. A task that believes it needs
  one of those stops and reports.
- **The spec's tree column names two harness packages; this pass writes
  five paths under `internal/harness`, two engine packages, and one
  corpus file.** The extension is deliberate and each part has a reason
  stated at its task: `stub.Unused()` is 3c's own named carry to 3d
  (HISTORY 3c guard 4), `corpus_test.go` is the sole production caller of
  the checker whose signature changes, `ties_derive_test.go` is the
  oracle-tagged half M4 must not silently break, `ties.json`'s two rows
  are M4's own asymmetry landing (spec `:133`), and
  `internal/engine/tag` is Resolution R18. This mirrors 3b's m14 and 3c's
  own tree-column corrections; Task 9 drafts the spec fix for the
  conductor.
- **`corpus/cases/*` is frozen input, and read-only is a real
  permission.** Loading every committed `case.json` and replaying every
  `responses/*.json` through `stub.Server` is exactly what the
  conformance runner is for. **No agent reads a `responses/*.json` body
  into context.** Every figure about response content in this plan was
  produced mechanically.
- **`internal/config` is not edited, and neither is `keys.go`.** Ruling
  C's registration half is hard-inbound on rung 1b (Phase 0b, unlanded).
  The four-knob registration stays the named follow-on task for the first
  rung 3 pass that runs after 1b closes (3a HISTORY guard 2, 3c guard 2),
  and `Threshold.Default` is a string with `ParseInt` the only parser, so
  the three float rec thresholds need a float variant the registry does
  not yet have. The `.gitleaksignore` rider applies even though this pass
  does not touch the file: its single allowed finding is pinned by line
  number at `internal/config/keys_test.go:generic-api-key:66`, and an
  insertion above line 66 reds the gate for every lane at once.
- **The engine never imports the harness, and never reads the clock.**
  `TestEngineNeverImportsHarness`
  (`internal/harness/conform/arch_test.go:141-146`) and
  `TestEngineNeverReadsTheClock` (`:152-157`) both root at
  `internal/engine` and walk `_test.go` files too, so
  **`internal/engine/tag` is inside both walks from the moment it
  exists**. `evalDate` is a parameter everywhere, never a clock read. An
  implementer that finds itself proposing an exemption stops and reports.
- **Assertion discipline, at the spec's own bar.** "Every checker in the
  runner ships a fixture that makes it fire, and the task that lands it
  reports the mutation that proved it can fire" (spec `:318`). The rule
  reaches this revision's additions by name: "the `TieEntry.Adjudicator`
  sentinel check, the `CheckOracleBugs` probe-case rule ... each ship a
  fixture proving the failure, and each is hand-authored in its package's
  own `testdata` rather than borrowed from the corpus" (spec `:320`). No
  silent-success fakes, no vacuous assertions, no self-derived
  expectations. **Hand-authored fixtures live in the owning package's own
  `testdata`, never a corpus case**, on the proximity precedent.
- **Risk 2 is this pass's own risk.** "The conformance runner is an
  instrument that can pass silently. A runner that grades fewer fields
  than it claims produces a green gate that means nothing, and every
  later claim about the matcher rests on it" (spec `:413`). Its named
  mitigations are all obligations of Tasks 7 and 8: the assertion
  discipline, a mutation proof per checker, and **an explicit
  graded-field table in the pass plan that a reviewer can check field by
  field against `cassette.Expected`**, which is the table in Task 7 below
  and the object of the fresh-context review.
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects and nothing else, verified at HEAD. No linear-assignment
  package, no `gonum`, no YAML parser. A task that wants one stops and
  reports (Resolution R4 rejects one by name).
- **Universality.** No estate hostname, path, email, or catalog value
  anywhere outside `internal/config`. Hand-authored fixtures are invented
  strings, never estate data.
- **No live network. Anywhere.** The gate never reaches the network and
  neither does anything else here. Every conformance run points at
  `stub.Server` on an ephemeral loopback address, and every client is
  built with `musicbrainz.Unmetered()`.
- **`make derive` does not run in this pass** (Resolution R17). The
  standing invariant is unaffected and is restated so it does not fall
  off: after this pass `make derive` must still re-derive
  `corpus/analysis/components.json`, `ties.json`, `file_tags.json`,
  `album_info.json` and every `corpus/cases/*/case.json`
  byte-identically. M4's one risk to that invariant is closed by an
  untagged, gate-resident merge helper and its own test (Task 6).
- **Determinism.** No map-iteration order reaches an output. The
  candidate accumulator keeps an explicit insertion-order slice beside
  its map; the downgrade key set is built and then walked in sorted
  order even though `min` over the enum is commutative, so a failure
  message is reproducible; every sort is `slices.SortStableFunc` with
  `cmp.Compare` or `cmp.Or`.
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck,
  ruff over every tracked `*.py` (this pass adds none), vale over Go
  comment prose, gitleaks over the tracked worktree, `docker compose
  config`, `systemd-analyze verify`. **Never commit red. Specific files
  only.** Commit footer `Co-Authored-By: Claude <noreply@anthropic.com>`.
  **Tier-0 replay green over all 72 rows, the translation tier green over
  all 50 rows, and the request tier green at 215/303/48 are part of every
  task's gate expectation**, because all three run inside `go test ./...`.
- **`go-conventions` before any Go.** The bar is the standing
  exemplary-stdlib-idiom bar (`idiomatic-beautiful-go`). Concrete types
  over interfaces, no functional options, `slices.SortStableFunc` with
  `cmp.Compare`, `for range n` where the index is unread, `errors.Join`
  for multi-error checkers, `any` not `interface{}`, no `init()`, no
  `context.Context` in a struct, errors lowercase and context-first with
  `%w` only where a caller branches. Two places need a written reason in
  the doc comment because they look like anti-patterns and are not: the
  `Candidates` accumulator's mutable state (Resolution R5) and
  `internal/engine/tag` existing as a one-file package (Resolution R18).
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each before Task
  1 dispatches.

## Contested surfaces, and the three rows the conductor owes

Lane-launch checklist item 2 requires a written owner for every surface a
lane touches (parallel-approach spec, "Lane-launch checklist"). 3d
touches nine. **Five already carry a row in the parallel spec's
contested-surface table, one rests on a standing rule with no row at
all, and three need an amendment** (revision 2, fold m15: revision 1
counted six written, and the `docs/` ledgers row it counted does not
exist).

| Surface this pass writes | Row today | Status |
| --- | --- | --- |
| `internal/engine/matcher` | "rung 3 (the E2 lane): `matcher` born by 3a, `musicbrainz` born by 3b; no other lane writes either" | written |
| `internal/harness/conform` | "born by rung 3 pass 3a; rung 3's lane thereafter" | written |
| `internal/harness/corpus_test.go` | "rung 3's lane while rung 3 runs; lane E1 otherwise" | written |
| `internal/harness/{cassette,stub,analysis}` | "lane E1 as shipped; rung 3 pass 3c makes the ruled `Case.Extra` schema-and-stub change, and rung 3 pass 3b adds the `AlbumInfo` ledger and the `defect_free` block to `analysis`" | written for the surface; **3d's own additions are unnamed** |
| `corpus/README.md` | conductor-owned; a pass drafts a correction and never writes the file | written, and **this pass writes nothing there** |
| `docs/` ledgers | **no row exists.** `:175`'s `corpus/README.md` row refers to "the same standing rule the `docs/` ledgers row states", and that row was never written | conductor-owned by CLAUDE.md's standing rule; **provisional, conductor ruling owed** on whether a fourth amendment writes it (fold m15) |
| **`internal/engine/tag`** | **none.** The engine row names `{matcher,musicbrainz}` by name, and this package is born here | **conductor-owed**, one amendment |
| **`internal/harness/oracle`** | "lane E1 through 2d's close (landed); rung 3 pass 3b adds the derived-AlbumInfo entry point" | **conductor-owed**: 3d edits `ties_derive_test.go`'s field mask and merge, which the row does not name |
| **`corpus/analysis/*.json`** | "lane E1 authors cases and derives ledgers; rung 3 adds schema only through re-derivation" | **conductor-owed**: M4's `adjudicator` field lands on two `ties.json` rows by hand-edit, because `derive_ties.py` emits no adjudicator and re-derivation cannot produce it |

**The three amendments this plan asks the conductor to make**, before the
task that first touches each commits:

| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/engine/tag` | rung 3 (the E2 lane), born by pass 3d as the one package that may see both `matcher` and `musicbrainz`; no other lane writes it | rung 4's shadow, rung 9's pipeline, rung 15's re-run |
| `internal/harness/oracle` (amendment) | lane E1 through 2d's close; rung 3 pass 3b adds the derived-`AlbumInfo` entry point, and rung 3 pass 3d carries `TieEntry.Adjudicator` through `writeTieLedger` and the diff test's field mask | `make derive` |
| `corpus/analysis/*.json` (amendment) | lane E1 authors cases and derives ledgers; `disposition` stays owner-only; rung 3 adds schema through re-derivation, **and, for `ties.json`'s `adjudicator` alone, by hand-edit, because the derivation emits no such field and the value records who set the disposition rather than setting one** | rungs 3, 8, 15 |

**A fourth amendment may be owed**, and it is the `docs/` ledgers row
the table refers to and never carries. Writing it is a one-line
conductor edit; leaving `docs/` ownership on CLAUDE.md's standing rule
is equally defensible, and 3d writes nothing in `docs/` that the
conductor does not commit. **Provisional, conductor ruling owed;
recommended: write the row, because a table that cites a row it does not
contain is the shape a later lane trips over.**

**None is a launch blocker.** The checklist's rule blocks the *second*
lane, and `docs/STATUS.md`'s lane registry records lane-2 as BLOCKED
until Phase 0b and 0c land. Task 0 reports all three; the conductor
writes them.

## The verdict-path contracts, stated once

Every task below cites this section instead of restating it. Sources: the
pinned venv at `tools/venv/.venv/lib/python3.13/site-packages`,
abbreviated `$SP`, read at drafting time and verified line by line
against the files themselves. Verify any line citation you carry into
code.

**V1. Every cost cell is a normalized track distance with
`incl_artist` false.** `costs = [[float(track_distance(i, t)) for t in
tracks] for i in items]` (`$SP/beets/autotag/match.py:64`).
`track_distance`'s third parameter defaults to `False`
(`$SP/beets/autotag/distance.py:383-385`), and `float()` is
`Distance.__float__` (`distance.py:218-219`), which returns
`self.distance` (`:153-160`), which is `raw_distance / max_distance` or
`0.0` when the denominator is zero. **Each cell is normalized against
its own denominator**, so cells are not on a common scale and a Go
shortcut that sums raw penalties builds a different matrix and therefore
possibly a different tied optimum. The Go call is
`matcher.DistanceForTrack(tun, item, &tracks[t], false).Distance()`
(`internal/engine/matcher/track.go:45`, `penalty.go:96`).

**V2. The solve returns a track-indexed array.**
`_, _, assigned_item_idxs = lap.lapjv(np.array(costs), extend_cost=True)`
(`match.py:66`). The third return value is indexed by **track** and holds
an item index or `-1` (`match.py:68-71`).

**V3. The pair list is in track-index order.** `mapping = {items[iidx]:
t for iidx, t in zip(assigned_item_idxs, tracks) if iidx != -1}`
(`match.py:72-76`) zips over the track-indexed result, so the dict's
insertion order is ascending track index. `dict(item_info_pairs)`
preserves it and `distance()` walks it to build both `dist.tracks` and
the `tracks` penalty list (`distance.py:527-530`). **Verified against the
corpus rather than inferred**: for `should-ask` and `should-reject` the
committed `Expected.Candidates[0].Tracks` array equals the per-track
distances taken in track order and differs from the same values taken in
item order. `matcher.addAlbumTracks`
(`internal/engine/matcher/album_tracks.go:25-31`) iterates the caller's
`pairs` slice verbatim, so **3d's assignment must hand
`DistanceForAlbum` its pairs sorted by track index**, or both the graded
`Tracks` vector and the left-to-right float sum of the `tracks` penalty
list are wrong.

**V4. The leftover set is one-sided.** `assign_items` never returns both
a non-empty `extra_items` and a non-empty `extra_tracks`: `lapjv` matches
exactly `min(rows, cols)` pairs. Confirmed on 4000 random rectangular
probes, on all 214 candidate-bearing corpus cases, and by
`corpus/README.md:132`, which says a two-sided case is not an expressible
shape. Measured off the tree at `c7a85fd`: 7 cases carry a non-empty
`extra_items`, 14 a non-empty `extra_tracks`, and **none carries both**.

**V5. The rectangular extension is `n = max(rows, cols)` padded with
ZERO.** Beets always passes `extend_cost=True` (`match.py:66`), and
`lap` 0.5.13 is a compiled extension with no source in the venv
(`tools/venv/.venv/lib/python3.13/site-packages/lap/_lapjv.cpython-313-x86_64-linux-gnu.so`;
`corpus/manifest.json` pins `lap 0.5.13` and `numpy 2.5.2`). Six
candidate extension models were tested against it on 2634 random
rectangular matrices: `max(rows, cols)` with zero fill agreed 2634/2634,
`max(rows, cols)` with `max+1` fill 92.0%, and the classic
`(rows+cols, max+1)` recipe that `derive_ties.py:58-66`'s own docstring
assumes **disagreed on 30% of rectangular ties**. All models agree on the
optimal cost; they differ only on which tied optimum is chosen, which is
exactly what tier 1 grades. Re-confirmed at this plan's drafting on 262
of 262 independent trials. On square inputs `extend_cost` is
behaviorally transparent (identical assignments on 5000+ tie-rich square
trials), and both pinned corpus cases are square.

**V6. Extras and their sort keys.** `extra_items = list(set(items) -
mapping.keys())` sorted by `(i.disc, i.track, i.title)`; `extra_tracks =
list(set(tracks) - set(mapping.values()))` sorted by `(t.index, t.title)`
(`match.py:77-80`). Both pre-sort orders are Python `set` orders over
identity-hashed objects, so they derive from addresses and
`PYTHONHASHSEED=0` does not control them; the committed records are
reproducible only because the sort keys are total.
`extra_tracks`' key is **provably** total, because `album_info`
renumbers every track `1..n` across the whole release
(`$SP/beetsplug/musicbrainz.py:645-646`) so `t.index` is unique.
`extra_items`' key is total only as a fact of today's corpus: no
committed case's `extra_items` holds two items sharing `(disc, track,
title)`. Go sorts with `slices.SortStableFunc`, never `sort.Slice` and
never `slices.SortFunc`, so the port stays faithful if a tie ever
appears, and 3d ships a guard asserting the absence (Task 1).

**V7. `_add_candidate` applies five gates in a fixed order**
(`match.py:331-380`), and a later gate never runs on a candidate an
earlier one dropped:

| Step | Rule | Source |
| --- | --- | --- |
| 1 | zero-track drop: `if not info.tracks: return` | `match.py:345-347` |
| 2 | duplicate drop: `if info.album_id and info.identifier in results: return` | `match.py:350-352` |
| 3 | `match.required`: for each tag, `if getattr(info, req_tag) is None: return` | `match.py:355-359` |
| 4 | solve the assignment, then score `distance(items, info, item_info_pairs)` | `match.py:362-367` |
| 5 | `match.ignored`: `penalties = [key for key, _ in dist]`, drop if any ignored tag is in it | `match.py:370-375` |
| 6 | insert at `results[info.identifier]` | `match.py:378-380` |

Order is load-bearing: a candidate rejected by `ignored` has already paid
for a solve and a candidate rejected by `required` has not, and a port
that reorders 1 and 2 changes the observable outcome on a zero-track
duplicate.

**V8. `ignored` reads the NONZERO view.** `for key, _ in dist` is
`Distance.__iter__` (`distance.py:240-241`), which is `items()`
(`:190-205`), which drops zero-weighted penalties. This is the spec's
m5 hazard's **second site**, alongside `max_rec`.
`cassette.Candidate.Penalties` is the added-key view
(`internal/harness/oracle/derive_oracle.py:176`), so a port that wires
`ignored` to `AddedKeys` or `PerKey`
(`internal/engine/matcher/penalty.go:124, 133`) instead of `NonzeroKeys`
(`:166`) is wrong, and tier 1 cannot catch it because `match.ignored` is
corpus-unreachable.

**V9. `required` tests `is None`, which Go cannot express here.**
`getattr(info, req_tag) is None` distinguishes an absent or `None` field
from an empty one, and `getattr` on an unmodelled name raises
`AttributeError`, which propagates out of `tag_album` and crashes the
import; beets has no guard. `matcher.AlbumInfo`
(`internal/engine/matcher/info.go:10-67`) carries fourteen fields,
thirteen of them plain-typed scalars with no nil (the fourteenth is
`Tracks []TrackInfo`), which is 3a's R1 ruling as 3b resolved it (an
absent source field becomes the plain-string zero value, 3b guard 2). See
Resolution R6 for what 3d does instead.

**V10. Candidate identity is `(data_source, id)`.**
`Info.identifier` returns `(self.data_source, self.id)`
(`$SP/beets/autotag/hooks.py:166-168`), and `AlbumInfo.id` is
`album_id`. `cassette.Candidate` carries exactly `Source` and `AlbumID`
as its identity keys (`internal/harness/cassette/cassette.go:103-109`).

**V11. Insertion order, then a stable sort by distance.** `candidates`
is an insertion-ordered dict (`match.py:414`); ID-derived candidates go
in first (`match.py:424-428`), then search results in the search
response's order (3c contract C14). `_sort_candidates` is
`sorted(..., key=lambda match: match.distance)` (`match.py:326-328`),
Python's stable Timsort, comparing exact doubles through
`Distance.__lt__`/`__eq__`. Corpus reach: **exactly two cases carry an
exact candidate-distance tie**, `should-ask` (three candidates at
`0.23868312757201643`) and `should-reject` (five at
`0.3891203703703704`), and in both the tie group's recorded order equals
the order the case logged its release lookups. So the stable sort is
graded, thinly, and `slices.SortFunc` in place of
`slices.SortStableFunc` is a legitimate mutation that fires on those two
and passes the other 213.

**V12. `_recommendation`'s ladder** (`match.py:279-302`), with every
config read and its shipped default:

```
no results                                  -> none                       match.py:279-281
d  <  strong_rec_thresh (0.04)              -> strong    STRICT <          match.py:284-287
d  <= medium_rec_thresh (0.25)              -> medium    INCLUSIVE <=      match.py:288-290
len(results) == 1                           -> low                        match.py:291-293
results[1].distance - d >= rec_gap_thresh (0.25) -> low                   match.py:294-299
otherwise                                   -> none, RETURNED IMMEDIATELY match.py:300-302
```

The `<` / `<=` asymmetry is the proximity flag's whole reason (spec
hazard 3), and `rec-strong-medium-boundary` is the corpus's single
flagged proximity row. `results[1].distance - min_dist` resolves through
`Distance.__rsub__` to a plain `float - float` (`distance.py:221-225`),
so bit-identity of both distances is a precondition of the `>=`
comparison agreeing.

**V13. The no-conclusion branch returns before the downgrade**
(`match.py:300-302`, comment "Return immediately. Can't be downgraded any
further."). Four committed cases take it
(`rec-none-small-gap`, `rich-album-id-recovers-clean`,
`rich-album-id-recovers-country`, `should-reject`). A port that fell
through to the downgrade loop would still return `none`, so the defect is
invisible on the corpus and needs its own assertion.

**V14. The downgrade key set is the NONZERO view, unioned with every
track sub-distance's keys.** `keys = set(min_dist.keys())`
(`match.py:306`) and, for an `AlbumMatch`, `for track_dist in
min_dist.tracks.values(): keys.update(list(track_dist.keys()))`
(`match.py:307-309`). `Distance.keys()` is built from `items()`
(`distance.py:246-247`), so **a key added at 0.0 never downgrades**.
Measured off the tree: **zero of 294 committed candidates carry
`missing_tracks` or `unmatched_tracks` at exactly 0.0**, even though 272
carry some key at 0.0, so the corpus is blind to this distinction for
the two keys that actually drive a downgrade. The track union is inert
under the shipped `max_rec`, because neither default key is ever a
track-level key; a model computing the downgrade from album-level
nonzero keys alone reproduces all 215 recorded recommendations.

**V15. `min(rec, max_rec)` over an ordered enum.** `Recommendation` is
an `IntEnum` with `none=0 < low=1 < medium=2 < strong=3`
(`match.py:39-47`), and `as_choice` raises on any name outside those four
(`match.py:313-320`). Iteration over the Python `set` is unordered and
`min` is commutative, so no ordering hazard exists; Go still walks a
sorted key slice so a failure message is reproducible.
`Tunables.MaxRec` is `map[string]string`
(`internal/engine/matcher/tunables.go:61-67`), so 3d owns the name
validation beets gets from confuse.

**V16. `_summary_judgment` is the verdict-to-action mapping**
(`$SP/beets/ui/commands/import_/session.py:328-360`), executed rather
than restated by the shim
(`internal/harness/oracle/derive_oracle.py:193-206`):

```
if import.quiet:                                        session.py:337
    if rec == strong: return APPLY                      session.py:338-339
    action = import.quiet_fallback in {skip, asis}       session.py:340-342
elif import.timid: return None                          session.py:343-344
elif rec == none: action = import.none_rec_action in {skip, asis, ask->None}   session.py:345-352
else: return None                                       session.py:353-354
```

`Action` values are `APPLY`, `SKIP`, `ASIS`
(`$SP/beets/importer/actions.py:8-19`); the shim lowercases them and maps
`None` to `"ask"` (`derive_oracle.py:206`). **The quiet check comes
first, so `Timid` and `NoneRecAction` are unreachable whenever `Quiet` is
set.**

**V17. The trap this contract exists to prevent.**
`derive_oracle.py:204` sets `config["import"]["quiet"] = True` on every
derivation, because dubplate imports with `beet import -q`, and quiet is
a CLI flag rather than a profile key. `Tunables.Quiet` defaults to
`false` (the field is `tunables.go:186`, the default literal `:296`), which is beets' shipped value and which
ruling D forbids changing. So a conformance run grading `action` against
the record with unmodified merged tunables computes `ask` for all 215
cases. **The conformance run overlays `Quiet: true` on the merged
tunables and cites the shim's own line at the overlay site**
(Resolution R12). The corpus's dispositions encode exactly two actions,
`apply` 167 and `skip` 48.

**V18. `tag_album`'s flow, transcribed** (`match.py:383-468`):

```
likelies, consensus = get_most_common_tags(items)              match.py:407-409
candidates = {}                          insertion-ordered      match.py:414
if search_ids:                           explicit-ID branch, NO early return   match.py:417-420
    for info in albums_for_ids(search_ids): _add_candidate(...)
else:
    for info in match_by_id(likelies["mb_albumid"], consensus["mb_albumid"]):  match.py:424-428
        _add_candidate(...)
    rec = _recommendation(list(candidates.values()))   NOT sorted   match.py:430
    if candidates and not config["import"]["timid"] and rec == strong:         match.py:432-441
        return cur_artist, cur_album, Proposal(list(candidates.values()), rec)  NOT sorted
    if not (search_artist and search_name): search_artist, search_name = cur_artist, cur_album
    va_likely = ...                                                match.py:451-455
    for c in metadata_plugins.candidates(items, search_artist, search_name, va_likely):
        _add_candidate(...)                                        match.py:459-463
candidates_sorted = _sort_candidates(candidates.values())          match.py:466
return cur_artist, cur_album, Proposal(candidates_sorted, _recommendation(candidates_sorted))  match.py:467-468
```

`_recommendation` is called on an **unsorted** list at `:430` and the
early return ships that unsorted list at `:441`. Harmless in practice
because `match_by_id` passes a one-element id list (`match.py:264`) so at
most one candidate exists, and the Go port says so at the site rather
than sorting early and calling it equivalent.

**V19. `match_by_id`'s two early branches** (`match.py:252-266`): it
returns `()` when the album id is falsy, and again when consensus on it
is `False`, logging which. "Consensus" is unanimity of the plurality
value. `matcher.MostCommonTags` already returns both halves
(`internal/engine/matcher/likelies.go:47`), and `va_likely` is 3c's
contract C10, computed inside `musicbrainz.AlbumSearchFor`
(`query.go:79, :108`).

**V20. The tracks, missing_tracks and unmatched_tracks summation
order** (`distance.py:527-538`): `dist.tracks` is filled pair by pair in
the caller's pair order, each pair adding one `"tracks"` penalty; then
one `1.0` under `"missing_tracks"` per excess track, then one `1.0`
under `"unmatched_tracks"` per excess item, each in loop order. Float
addition is not associative, so V3's track-order requirement is a
correctness requirement rather than a cosmetic one.

## Resolutions this plan takes rather than discovers

The planning inventory returned twenty deduplicated open questions
(`docs/research/2026-09-05-rung-3d-planning-inventory.md`,
section (d), "Combined open questions"). Each is resolved here, with its
evidence and its rejected alternative. A task that believes a resolution
is wrong reports a finding and implements the record anyway.

### R1: only two of the four `oracle-wrong` classes get a compat switch (inventory Q1)

**Resolved: `matcher.Compat` carries exactly two fields, `FtPattern` and
`RemixParenthetical`, both defaulting off because both hypotheses are
ruled `oracle-wrong`. The other two `oracle-wrong` hypotheses ship no
switch and no fix, each with its written reason, and their probe cases'
corrected verdicts are recorded as ungraded expectations rather than
graded and rather than silently reproduced.**

The evidence, read off the tree:

| Hypothesis | Disposition | Probes | The "off" behavior | Implementable and gradable at rung 3? |
| --- | --- | --- | --- | --- |
| `bug-ft-pattern` | oracle-wrong | 6 | `SDPatternsWordBoundary` | Yes. All six rows carry a `defect_free` block |
| `bug-remix-parenthetical` | oracle-wrong | 4 | `SDPatternsNoBracketDiscount` | Yes. All four rows carry a block |
| `bug-nonlatin-alias-selection` | oracle-wrong | 4 (2 rows `oracle-wrong`, 2 `oracle-correct`) | read `CreditPart.Name` in place of `part.Artist.Name` at `internal/engine/musicbrainz/credit.go:53` | **No.** The corrected verdict is an applied artist string, rung 3 computes no applied field, and the change would be visible only through 3b's `album_info.json` ledger, which holds the defective values and has no defect-free variant |
| `bug-assignment-rejection-gap` | oracle-wrong | 6 | a per-pair rejection mechanism beets does not have | **No.** Ruling A says so itself: "the fix has no beets counterfactual" (spec `:75`). The corrected verdicts are prose naming an identity assignment and a would-be distance reachable only by inventing rejection |

Shipping either unimplementable fix would be an improvement riding the
port, which the build spec forbids outright, and shipping it ungraded
would be exactly risk 2's shape. So neither is shipped.

**What the six `adv-assign-*` and two alias probe cases do at tier 2.**
In `ModeCompat` they grade against their own committed records like every
other case. In `ModeDefault` their verdicts are **not graded**, and each
is listed in `CaseGradeSummary.Ungraded` with the reason. The pass report
names all eight, and Task 9 records them as improvements-register items
that the launch gate inherits.

**The close drafts one owner ask, batched into the adjudication stream
rather than raised mid-pass.** `bug-assignment-rejection-gap`'s
disposition would be exactly right as `no-correct-answer`: the spec's own
grading table says that disposition means "the class has no correct
answer to converge on. The switch defaults on, no probe case's verdict is
graded, and each probe case's distance is still graded at the tier its
own `proximity` row sets" (spec `:296`), which is the behavior this
resolution implements by hand. Re-dispositioning the hypothesis would let
the runner reach that behavior through the ledger's existing vocabulary
rather than through a special case. `disposition` is owner-only, so the
pass drafts the ask and never writes it.

Rejected: shipping a `Compat.AssignmentRejection` switch whose "off"
branch is unimplemented (a switch that cannot be flipped is dead surface,
and one that could be flipped would ship invented behavior); asking Geoff
mid-pass (the pass has a resolution that runs unattended, and the process
batches adjudication asks); grading the six against their prose corrected
verdicts (there is no machine-readable expectation, which is why
`CheckDefectFree`'s `StringHeuristicHypotheses` predicate already
excludes the class, `internal/harness/analysis/adjudication.go:86`).

### R2: both string switches off needs a fourth table, and it is safe (inventory Q4)

**Resolved: 3d builds a fourth table,
`matcher.SDPatternsGuardedNoBracketDiscount`, and `Compat.Apply` selects
one of four tables from the two booleans.**

`Tunables.StringDistPatterns` is one slot
(`internal/engine/matcher/tunables.go:217-224`) and neither committed
table fixes both defects: `SDPatternsWordBoundary` (`:365`) guards the
featuring pattern and keeps the two `0.3`-weight bracket rows, while
`SDPatternsNoBracketDiscount` (`:388`) drops the bracket rows and keeps
the unguarded featuring pattern. Both hypotheses are `oracle-wrong`, so
ruling A's scheduling rule defaults both switches off at once, and no
committed table expresses that.

**The fourth table is safe, measured rather than assumed.** Across the
six `adv-ft-*` probe cases, **zero** compared strings carry a bracket or
parenthesis; across the four `adv-remix-*` cases, **zero** carry a
featuring trigger. The compared set is each case's `input.items`
`title`, `album` and `artist` plus every `title`, `name`, `artist`,
`albumartist`, `sort-name` and `disambiguation` string in its response
bodies, named because the denominator moves with the field set and the
zero does not (fold m18). Counts only; no body was read into context.
So the combined table reproduces `SDPatternsWordBoundary`'s six
committed `defect_free` distances and `SDPatternsNoBracketDiscount`'s
four, exactly.

**The confirmation is the check rather than a re-derivation.**
`ModeDefault`'s tier-2 run grades all ten probe cases against their
committed `defect_free.distance`, `recommendation` and `action`. If the
combined table were wrong on any of the ten, that run fails naming the
case and both figures. No `make derive` is needed and none is run
(Resolution R17).

**`ModeDefault` re-scores rather than re-runs, and that is what makes
the check possible at all** (revision 2, fold B2; **provisional,
conductor ruling owed**). All ten probe cases log exactly one release
lookup and zero searches, they score `strong` under compat and `medium`
under the corrected table, and the ID-match early return needs
`rec == strong` (`match.py:436`, contract V18). So a `ModeDefault` run
that re-ran `tag.Album` would lose the early return on every one of the
ten, issue a search the case never logged, take a stub 404 and a
recorded `Miss`, and fail Task 7's own zero-miss rule before any
`defect_free` figure was compared. `ModeDefault` therefore replays each
case's own committed candidate set through `matcher` under the
corrected table (translate each logged release lookup, assign, score,
sort, `Recommend`, `ActionFor`) and issues no search of its own. That
reproduces the ten committed figures exactly, because the ladder's
medium rung is decided by the minimum distance alone
(`match.py:288-290`), so a candidate set a search would have widened
cannot move a `medium` verdict off `medium`.

The alternatives, for the conductor's ruling: (b) re-run the flow with
the request-set rule relaxed to `ModeCompat` and a
search-404-to-empty-result mapping in `tag.Album`, which is beets' own
shape (`$SP/beets/metadata_plugins.py:421-427`'s bare except returns
`()`) and what `SearchReleases`' own doc comment says 3d does, at the
cost of a second full flow run and a `Misses()` list nothing asserts;
(c) drop `ModeDefault` to the ten cases alone and give up
`VerdictMoves`. **Recommended: (a), the re-score above**, because it is
the cheapest, it grades exactly what tier 2 needs, and it leaves the
committed request log as an expectation of the compat run alone.

**One caveat the re-score design carries, stated rather than absorbed.**
A real default-mode import could fetch a different candidate set,
because a case that loses its early return would search. `VerdictMoves`
therefore measures the fix's effect on each case's already-fetched
candidates, which is the blast radius R9 wants, and not a full
counterfactual import. Task 9 states that limit where it records the
list.

The `\b` regime divergence rides with the new table unchanged: Python
`re`'s `\b` is Unicode-aware and Go RE2's is ASCII-only, benign today
because all six `adv-ft-*` pairs are ASCII (3b guard 11). The new table
inherits the same guard comment, and Task 9 records the three regime
divergences together as one family, as 3c did (3c guard 9).

**The new table is deliberately NOT a legal `pattern_table` value, and
the plan says so before someone helpfully adds it** (revision 2, fold
m9). `analysis.LegalPatternTables`
(`internal/harness/analysis/adjudication.go:96`) pins exactly
`SDPatternsWordBoundary` and `SDPatternsNoBracketDiscount`, enforced at
`:254-255`, and every committed `defect_free.pattern_table` carries one
of those two (six and four). That field records **which single-defect
edit produced the recorded figure**, which is a derivation fact about
pinned beets. `SDPatternsGuardedNoBracketDiscount` is a Go composition
of both edits and derived nothing, so adding its name to
`LegalPatternTables` would let a hand-edit claim a derivation that never
ran. `internal/harness/conform`'s existing cross-check
(`defaults_test.go:178`) keeps comparing the two ledgered names against
the two matcher tables they stand for, and the fourth table is compared
only against its own byte-identity test in `compat_test.go`.

Rejected: representing the compat surface as per-class edits composed
into a table at run time (it would make the shipped table a computed
value, so the byte-identity test that guards `BeetsSDPatterns` against a
silent "improvement", 3a guard 3, would have nothing to compare against);
running two separate default-mode passes with one switch each (it grades
a configuration nobody ships, since both dispositions are already
`oracle-wrong`).

### R3: ruling F, the two-attempt trigger, and what tier 1 does while the ruling is owed (inventory Q20)

**Resolved: the exact solver and the tie mirroring are two tasks, and
only the second carries ruling F's trigger.**

The split is the inventory's own recommendation and it is right for a
budget reason as well as a correctness one. A correct exact solver with
any tie rule reproduces **212 of the 214 measurable assignment rows**,
because `ties.json` records 213 rows with no disposition, 2 pinned and 1
skipped. Only two rows turn on `lapjv`'s choice among equal optima, and
they are the whole of ruling F's risk.

**Task 1 ships the exact solver.** Correctness is provable
independently of any tie rule: the returned assignment's cost equals a
brute-force optimum for small `n`, and exactly `min(rows, cols)` pairs
are matched.

**Task 2 attempts the mirroring, and carries ruling F verbatim.** Two
failed attempts stop **Task 2 and only Task 2**. The implementer then:

1. writes the finding to
   `docs/research/2026-09-05-rung-3d-tie-mirroring-finding.md`,
   carrying **both cost matrices at full float64 precision** (12x12 for
   `should-ask`, 14x14 for `should-reject`), **both solvers' optima and
   their costs** (`lapjv`'s and the Go solver's, with `ties.json`'s own
   `optimal_cost` figures `1.7962962962962963` and `5.282407407407407` as
   the cross-check), the committed `Expected.Assignment` for each case,
   and the demonstration from `ties.json`'s own two reasons that both
   optima produce the same album distance and the same recommendation;
2. states, as an offered option and never an assumption, that
   re-dispositioning both rows from `pinned` to `excluded` would cost
   two `assignment` fields and nothing else, since both cases carry
   empty `extra_items` and `extra_tracks`, and would move the
   ties-excluded count from 0 to 2, which is precisely the growth that
   "may not grow without an owner ruling" (spec `:292`) refers to;
3. marks the two rows as awaiting Geoff in the pass report and in
   STATUS. **It writes no `disposition` and edits no ledger field but
   `adjudicator`, which Task 6 owns.**

**The conductor never rules the disposition** and never converts the
finding into a decision. The pass continues through Tasks 3 to 9.

**What tier 1 does with the two pinned cases while the ruling is owed.**
They grade as **pending, never as passing**. `CaseGradeSummary` carries
`TiesAwaitingRuling []string`, tier 1 skips the assignment comparison for
those two case ids alone, still grades their distances, candidates,
per-track vectors and verdicts, and the gate test **fails if that slice
is non-empty and no `docs/research/2026-09-05-rung-3d-tie-mirroring-finding.md`
exists**, so the escape cannot be taken silently. When the finding
exists the gate is green, the count is reported at every run, and the
launch gate inherits the open ruling.

**This is a deliberate, named extension of the spec's pending
principle**, and the plan says so rather than letting it read as
compliance. Spec `:311` says the pending escape is "tier 2 only. Tier 0,
tier 1, and any `proximity`-flagged case admit no pending state at all,
because a mechanical disagreement is a Go defect until an owner says
otherwise." Ruling F is the one case the spec itself creates where a
mechanical disagreement is **not** a Go defect: the spec pre-authorized
the escalation and put the disposition in Geoff's hands, so the same
sentence that forbids the escape also creates the state the escape is
for. The alternative, leaving the module gate red until Geoff rules,
reds master for every lane over an owner question, which is exactly what
the M7 fold exists to prevent (spec `:306`). The conductor takes this
extension, it is recorded here and in HISTORY, and the launch gate
refuses the open ruling.

**Encouraging evidence for the attempt, and why two is a fair budget
rather than a formality.** `lapjv`'s pivot order is an algorithmic
artifact rather than a documented rule: on an all-zeros `n x n` matrix
it returns column-to-row `[1,0]`, `[1,2,0]`, `[1,2,3,0]`,
`[1,3,2,4,0]`, `[1,4,3,2,5,0]` for n = 2 to 6, and the shift-by-one
rule that fits `n <= 4` breaks at `n = 5`. (Column-to-row is the third
return value, which is what `solveLAP` reproduces and what beets
consumes; the argument reads the same on the second return value's
inverse permutation, and the numerals differ. Revision 2, fold B1.) A
Hungarian solver with a lowest-column-index
tie rule returns the identity on both pinned cases and fails both, since
`should-ask` swaps items 1 and 6 across 12 tracks and `should-reject`
runs a three-cycle 5 to 7 to 9 to 5 across 14. Against that, a synthetic
12x12 matrix with a uniform tied block at rows and columns 1 and 6
returns exactly `[0,6,2,3,4,5,1,7,8,9,10,11]`, the committed `should-ask`
mapping. That is one data point rather than a proof, and it is what makes
a faithful transliteration of the three C++ routines (column reduction
with reduction transfer, augmenting row reduction, Dijkstra
augmentation, named by `nm -D` as `_ccrrt_dense`, `_carr_dense`,
`_ca_dense`) worth two attempts.

### R4: the solver is hand-written, and no dependency is added (inventory Q8, second half)

**Resolved: `solveLAP` is unexported, hand-written, and lives in
`internal/engine/matcher/assign.go`. No module dependency is added.**

`go.mod` carries cobra and its two indirects and nothing else, verified
at HEAD, and a Global Constraint forbids adding one. More to the point, a
third-party assignment package would solve the wrong problem: the
requirement is not an optimal assignment but **`lap` 0.5.13's specific
optimum among equals**, which no general package promises and which a
version bump inside such a package would silently change.

**The differential evidence is frozen as testdata rather than shipped as
an oracle-tagged test.** Task 2 generates
`internal/engine/matcher/testdata/lapjv-optima.json` once, offline, with
`tools/venv/.venv/bin/python` driving the pinned `lap`, and commits it
with a header naming the pinned version and the exact one-line command
that regenerates it (the `case-mapping-sweep.tsv` precedent,
`internal/engine/matcher/testdata/case-mapping-sweep.tsv:7-10`). A
gate-resident Go test replays every row. That buys layer 3's value, the
establishment of the extension model and the pivot order against pinned
`lap`, in a form that **gates**, which an oracle-tagged test cannot do.

Rejected: an oracle-tagged differential test with a Python driver in
`derive_ties.py`'s shape (it adds a tracked `*.py` file to the ruff leg
and a `//go:build oracle` test the gate can never run, for a re-run
property that `corpus/manifest.json`'s `lap 0.5.13` pin plus `Reconcile`
already signal; Task 9 files it as a note for the lane that owns
`internal/harness/oracle`, to be built the day the `lap` pin moves).

### R5: `Candidates` is an accumulator, and it earns its state (inventory Q1's Go half)

**Resolved: `matcher.Candidates` is a small mutable accumulator with an
explicit insertion-order slice beside its identity map, and its doc
comment carries the reason go-conventions would otherwise reject it.**

The ID stage and the search stage add into the same result set, so V7's
duplicate rule spans both, and `tag_album` reads the set's length between
the two stages (`match.py:432`). A pure function returning a new slice
per candidate would rebuild the identity index on every add and would
lose the insertion order the stable sort depends on. `Add` performs no
I/O, so the type stays in `matcher` and the package stays importable by
`musicbrainz`.

`Add` returns a `bool` reporting admission rather than an error, because
in beets a drop is a `log.debug` and a `return`, never an exception.

### R6: `match.required`'s `is None` has no Go analogue, and 3d rules it (inventory Q3)

**Resolved on all three halves, with the divergence recorded on the
filter's own doc comment.**

- **Which names the filter accepts.** Exactly the thirteen scalar field
  names `matcher.AlbumInfo` models, matched through a name table in
  `candidate.go`. `tracks` is excluded on purpose: gate 1 already drops
  a zero-track candidate before gate 3 runs, so naming it in `required`
  could never change an outcome. Beets accepts any attribute name and raises
  `AttributeError` on an unknown one, which crashes the import mid-run.
- **What "None" means per field kind.** The field's Go zero value: `""`
  for a string, `0` for an int, `false` for a bool. **For the one bool
  that is a behavioral divergence rather than a modelling choice, the
  filter's doc comment says so** (revision 2, fold m4): beets declares
  `va: bool = False` (`$SP/beets/autotag/hooks.py:351`, assigned at
  `:381`) and never `None`, so `required: [va]` filters nothing in beets
  and would drop every non-VA release in Go. The three int fields
  (`Year`, `OriginalYear`, `Mediums`) carry a weaker form of the same,
  since beets can genuinely hold `None` there. 3b's guard 2 settled
  that no rung 3 source produces an empty-but-present field, because
  `internal/engine/musicbrainz` is the only translation source and every
  field it writes comes from a Go decode that yields `""` rather than a
  null. The doc comment names the condition that reopens it: a second
  metadata source, whose arrival the translation tier's three
  `nullStrict` fields would catch.
- **What an unknown name does.** `Tunables.Validate()` returns an error
  wrapping `ErrInvalidTunable` at config-load time. This is a
  **deliberate divergence from beets**, recorded with its reason: a
  refusal at start beats a crash mid-import, and beets' own behavior here
  is an unguarded `getattr` rather than a designed one.

Rejected: modelling `AlbumInfo`'s fields as pointers to express
`None` (it reverses 3a's ruled R1 resolution, which 3b closed, for a
corpus-unreachable filter); accepting arbitrary names through reflection
(it reproduces beets' crash and buys nothing, since the twenty-five
beets `AlbumInfo` fields Go does not model carry no values Go could
test).

### R7: `Compat` is its own type, and it never reaches the translation layer (inventory Q5)

**Resolved: `Compat` is a separate type threaded as its own parameter,
not a field on `Tunables`, and `musicbrainz.TranslateAlbum`'s signature
does not change.**

The two are different kinds of thing. A tunable is an operator-facing
setting with a beets default and a citation; a compat switch is a
selection the corpus's own adjudication ledger implies, and
`DefaultCompat`'s literals cite hypothesis ids rather than
`config_default.yaml` lines. Putting the switches on `Tunables` would
muddy a type whose doc comment promises the opposite.

The inventory's objection was that `AliasCreditName`'s fix site sits in
`internal/engine/musicbrainz/credit.go`, so `Compat` would have to reach
the translation layer. **R1 removes that switch**, so the objection
dissolves: with only the two string-heuristic switches, `Compat.Apply`
returns a `Tunables` with one field changed and nothing crosses a package
boundary.

### R8: two named runs, not one (inventory Q6)

**Resolved: the conformance runner has two modes and the plan names both
rather than leaving the second implied.**

- **`ModeCompat`** reproduces every ledgered defect. It drives
  `tag.Album` over all 215 cases against each case's own stub, grades
  candidates, distances, per-track vectors, assignments, extras,
  recommendation and action against every committed record including
  the adversarial stratum, at the tier each case's own `proximity` row
  sets, and enforces the request-set rule. This is the bit-identity
  run, and it is the one the phrase "tier 1" refers to. Its compat
  value is `matcher.FullCompat()`, both switches on (fold m6).
- **`ModeDefault`** takes the ledger's own dispositions. Both string
  switches are off, so `Compat.Apply` installs
  `SDPatternsGuardedNoBracketDiscount`. It **re-scores each case's own
  committed candidate set** rather than re-running the flow (revision 2,
  fold B2; **provisional, conductor ruling owed**, with the reason and
  the alternatives at R2), so it issues no search, and the request-set
  rule does not apply to it. It **grades verdicts only for the ten probe
  cases carrying a `defect_free` block**, against `defect_free.distance`
  (at 1e-12 relative), `defect_free.recommendation` and
  `defect_free.action`. No other case's verdict is graded, because no
  other case has a corrected expectation.

With both string classes ruled `oracle-wrong`, a default-mode run scores
every case against a non-beets pattern table, so the two-run shape is
mandatory rather than optional: one run cannot be both bit-identical
against the committed records and correct against the corrected
verdicts.

### R9: what `ModeDefault` does when a non-probe case's verdict moves (inventory Q7)

**Resolved: it is reported as a count and a case list, it fails nothing,
and Task 9 hands it to Geoff as one adjudication batch.**

A verdict that moves under a fix the owner already ruled correct is a
deliberate consequence of that ruling, not a divergence needing a pending
row per case. Writing twenty or forty pending rows unattended would
misrepresent them and would flood a queue the process expects to stay
small.

The blast radius is unmeasured and is not obviously small: by substring
presence, 133 committed cases carry a bracket or parenthesis somewhere in
a compared string and 44 carry a featuring trigger the word-boundary
guard would refuse. Both are upper bounds, because `StringDist` only
shifts when the deletion lowers the distance
(`internal/engine/matcher/strdist.go`'s `caseDelta == 0` continue,
matching `$SP/beets/autotag/distance.py:108-109`), so a parenthetical
present on both sides is inert. **The real count is a named deliverable
of Task 8**, obtained by running both modes and diffing.

`CaseGradeSummary.VerdictMoves` carries the case ids. It is measured by
re-scoring each case's committed candidate set under the corrected table
and comparing that verdict against the case's own `ModeCompat` verdict
(revision 2, fold B2), so it reports the fix's effect on the candidates
the case actually fetched and never a full counterfactual import; a case
that lost its early return would have searched, and no such search is
issued. Task 9 records that limit beside the list. Task 9 writes them
to `docs/research/2026-09-05-rung-3d-default-mode-verdict-moves.md`
with each case's compat and default verdicts, and the conductor batches
that document into the adjudication stream, which is a touchpoint the
process already expects (CLAUDE.md, "Expected owner touchpoints").

### R10: `TieEntry.Adjudicator`'s mechanism, end to end (inventory Q10)

**Resolved: the field lands, the sentinel applies to `excluded` only, the
oracle-tagged carry-forward moves into an untagged helper with its own
gate-resident test, and the two committed pinned rows take `agent` by
hand-edit.**

Seven sites, all of them named so none is discovered at a red `make
derive` or a red `go test` (revision 2, fold M-c added the seventh, an
existing passing fixture the new rule breaks):

| Site | Change |
| --- | --- |
| `internal/harness/analysis/ties.go:56-66` | add `Adjudicator string` with `json:"adjudicator,omitempty"`, and the asymmetry's reason in its own doc comment (spec `:133`) |
| `ties.go:114-142` (`checkTieEntry`) | require `Adjudicator` whenever `Disposition` is set; require a non-`AgentAdjudicator` value **when `Disposition == TieExcluded` only** |
| `ties.go:148-159` (`checkSkippedTieEntry`) | forbid `Adjudicator` on a skipped row, beside the existing `AlbumID` and `Disposition` prohibitions |
| `ties_test.go:12-14` | the existing `dispositioned` fixture is `TiePinned` with a `Reason` and NO `Adjudicator`, and its row at `:47-51` expects acceptance, so the new rule turns a passing row failing. It gains `Adjudicator: AgentAdjudicator`. The neighbouring `noReason` fixture (`:17-18`) needs nothing: `checkTieEntry` joins its errors and `wantHas` is a substring test |
| `ties_test.go:38-140` | five new table rows in the table's own `{name, caseIDs, ledger, wantHas}` shape (there is no `entry` field): an `excluded` row carrying `agent` (must fail), an `excluded` row carrying a human name (must pass), a `pinned` row carrying `agent` (must pass, which is the asymmetry's own proof), a dispositioned row with no adjudicator (must fail), and a skipped row carrying one (must fail) |
| `internal/harness/oracle/ties_derive_test.go:91` and `:133-141` | the diff test's field mask gains `want.Adjudicator = got.Adjudicator`, and `writeTieLedger`'s carry-forward moves to the new untagged helper |
| `corpus/analysis/ties.json` | hand-add `"adjudicator": "agent"` to the two pinned rows |

**Why the merge helper is untagged and lives in `analysis`.** Every file
in `internal/harness/oracle` is `//go:build oracle`, so an untagged
helper there would be that package's only gate-visible file and would
change its build shape. `TieEntry` lives in `analysis`, which is
untagged, so `analysis.MergeTieDispositions` is both the right home and
the only one where a gate-resident test can prove the carry-forward. Its
absence is the concrete failure this closes: `derive_ties.py` emits six
fields and never `disposition` or `reason`
(`internal/harness/oracle/derive_ties.py:118-126`), so without the carry
`-update-ledgers` erases the field on the next `make derive`, and the
gate cannot see it happen because both sites are oracle-tagged.

**Why `agent` on the pinned rows is correct rather than convenient.**
Zero excluded rows exist, so the sentinel half has **no corpus instance
and cannot fire against the committed tree**; its proof is the
hand-authored table row above. Both pinned rows were agent-authored,
which is the fact the spec's adaptation rests on (spec `:133`), and the
tree confirms it: `grep -c adjudicator corpus/analysis/ties.json` returns
0 today, so neither row carries a name of any kind. Writing `agent` on
both is the only value that is true and keeps the gate green, which is
exactly the outcome the asymmetry was chosen to produce.

### R11: 3d takes a fresh-context review of the runner alone (inventory Q11, spec `:362`)

**Resolved: yes, and scoped to the runner alone. One `claude-opus-5`
dispatch, after Task 7 lands, before Task 8.**

The spec raises this as the one candidate exception to rung 3's review
scale (`:362`), and 3c declined it for its own request tier with a
recorded reason. 3d takes it, for three reasons and none of them is
generic caution.

**The instrument argument is the spec's own** (risk 2, `:413`): every
later claim about the matcher rests on this runner, rung 15's re-run
criterion inherits it verbatim, and rung 4's shadow inherits its grading
vocabulary. A runner that grades fewer fields than it claims produces a
green gate that means nothing.

**The evidence argument is that this failure mode has already occurred
twice in rung 3.** 3a's `aggregate/sorted-penalties-tied` fixture could
not discriminate a `(distance, key)` sort from no sort at all and its
specified mutation did not fire; 3b's Task 5 mutation was provably
unobservable and was struck. Both were found by mutation testing rather
than by reading code, which is what a fresh reviewer with a graded-field
table repeats.

**The cost argument is that the scope is bounded and checkable.** The
review reads `GradeCases` against `cassette.Expected`'s eleven fields
(`internal/harness/cassette/cassette.go:76-88`) field by field using Task
7's graded-field table, confirms every row of the spec's disposition
table (`:285-299`) has a checker and a fixture, and re-runs Task 7's
named mutations. That is one dispatch against a bounded surface, not the
full adversarial cycle rungs 5, 12 and 16 carry.

**It is not extended to the engine half.** `assign.go` and `verdict.go`
are graded by 214 committed assignments, 294 candidate vectors and six
reachable recommendation branches, which is stronger evidence than a
reviewer produces. The in-chain `diff-reviewer`, the `code-simplifier`
dispatch and the pass-end gate stay as they are.

### R12: the `Quiet: true` overlay is explicit at the call site (inventory Q12, Q13)

**Resolved: `conform` gains `gradeTunables(profile string) (matcher.Tunables, error)`,
which is `mergeProfile(matcher.DefaultTunables(), profile)` followed by
one line setting `Quiet = true` with `derive_oracle.py:204`'s citation
beside it. The overlay does not go inside `mergeProfile`.**

`mergeProfile` (`internal/harness/conform/translate.go:193-204`) models
the three committed profile **files**, and `checkProfileKeyDrift`'s
allowlist (`:213-224`) is honest only while that stays true. `quiet` is
in no profile file; it is a CLI flag the shim forces. Hiding it inside
the profile table would make the drift guard describe something that is
not a profile.

**That premise is false as landed, and Task 7 makes it true before it
leans on it** (revision 2, fold M-b). `corpus/profiles/estate.yaml:17-19`
and `locale.yaml:17-19` both set `quiet_fallback: skip`,
`none_rec_action: skip` and `timid: no`, and `mergeProfile` returns
`base` unchanged for `stock` and `estate` and sets only `Languages` and
`IgnoredAliasTypes` for `locale`, while `profileKeyAllowlist` maps all
three keys to `Tunables` fields. One of the three genuinely differs from
beets' shipped value: `$SP/beets/config_default.yaml:34` is
`none_rec_action: ask`, which `DefaultTunables()` carries
(`tunables.go:298`), against `skip` in both profiles. So
`gradeTunables("estate").NoneRecAction` reads `"ask"` where the profile
says `"skip"`, invisible today only because the `Quiet: true` overlay
makes the field unreachable (contract V16). **Task 7 therefore extends
`mergeProfile`'s `estate` and `locale` arms with those three
assignments, each with a test asserting the value against the profile
file.** The change is behavior-preserving under the overlay and it is
what lets `gradeTunables`' doc comment be true rather than aspirational.

`Tunables.Quiet` keeps beets' shipped default of `false`, which ruling D
requires. `QuietFallback` at its `skip` default is the only fallback the
corpus reaches, and `asis`, `Timid` and `NoneRecAction` are declared gaps
closed by a table-driven test in `verdict.go`'s own file.

### R13: tier 1's M1 cross-check moves onto the merged result (inventory (b) §8 item 8)

**Resolved: `CheckCaseWeights` and `CheckCaseThresholds` keep their
current callers, and tier 1's own configuration path uses
`gradeTunables`, so the echo is compared against the merged result rather
than against `matcher.DefaultTunables()` directly.**

The M1 fold specifies "`Tunables`' beets defaults **merged with** the
case's profile, in that order" and states the cross-check against the
merge (spec `:277-279`). As landed, `defaults_test.go:43, 55-61` compares
the echo against `DefaultTunables()` directly. The two are equal today
because no committed profile sets a `match:` or `musicbrainz:` key, so
moving the comparison onto `mergeProfile` is behavior-preserving and
closes the drift before a fourth profile makes it real. Task 7 makes the
move and says so.

### R14: B3 lives in `CheckOracleBugs`, and the signature takes the ledger (inventory Q15)

**Resolved: the rule lands in `CheckOracleBugs` where the spec puts it
(`:300`), and the signature becomes
`CheckOracleBugs(caseIDs []string, adj *AdjudicationLedger, l *OracleBugLedger) error`.**

The rule needs a subject index and the existing `adjudication_refs` check
needs the id set, so passing the loaded ledger collapses both into one
parameter and lets the checker build both indexes itself. The only
subject index that exists today is `CheckDefectFree`'s local, unexported
`rows` map (`internal/harness/analysis/adjudication.go:224-229`), built
inside one function and thrown away; 3d exports it as
`(*AdjudicationLedger).BySubjectCase()` and both checkers use it.

Placing the rule in `CheckOracleBugs` rather than widening
`CheckDefectFree`'s probe walk keeps the error prefix truthful: a
hypothesis naming a probe case with no adjudication row is an oracle-bug
ledger defect, and `defect_free:` is the wrong prefix for it.

**Exactly two call sites change.** `internal/harness/corpus_test.go:406`
(which stops hand-building `adjIDs` at `:397-400`) and
`internal/harness/analysis/oraclebug_test.go:97`, whose sixteen-row table
swaps its `adjudicationIDs map[string]bool` field for an `adj
*AdjudicationLedger` field.

**The rule fires on nothing in the committed corpus and its fixture must
be hand-authored.** Verified case by case: all twenty probe cases of the
four `oracle-wrong` hypotheses resolve to an adjudication row by
`subject.case_id`. The two probe cases with no row,
`rec-strong-medium-boundary` and `real-nonlatin-utada-first-love`, both
sit under `oracle-correct` hypotheses, so the rule does not reach them.
Its whole value is prospective, against the next hypothesis Geoff rules
`oracle-wrong`.

### R15: the runner reports a pending divergence; it never writes one (inventory Q16)

**Resolved: `GradeCases` never writes into `corpus/`. A tier-2 verdict
disagreement with no adjudication row fails the gate, and the failure
message carries the exact JSON row a human pastes into
`adjudications.json`.**

The reason is not squeamishness. A gate that edits the corpus is not a
gate: the corpus discipline says a captured record is evidence and is
never edited, and `go test` writing a disposition-bearing ledger row
during `check.sh` would make the expectation a function of the run.
`CheckAdjudications` additionally requires `Suspected`, `ReleaseTruth`,
`Reason`, `Adjudicator` and `Date` all non-empty on a pending row
(`internal/harness/analysis/adjudication.go:159-173`), four of which are
prose a runner cannot write honestly.

So the pending path has two halves and the plan states both:

- **A case whose adjudication row already reads
  `pending-adjudication`.** Its verdict is not graded, `check.sh` stays
  green, its id is reported in `CaseGradeSummary.Pending`, and the launch
  gate refuses it. This is the spec's escape, working exactly as `:310`
  describes. Today `adjudications.json` carries zero pending rows, so
  this path has no live corpus instance and its checker's proof is a
  hand-authored fixture in `conform/testdata`.
- **A divergence with no row at all.** The gate **fails**, naming the
  case, the field, both values, and printing the ready-to-commit row.
  The conductor then rules which obligation it is, exactly as spec `:314`
  assigns: lane E1's red to clear by landing the row or reverting the
  case, or lane E2's Go defect to fix. That is a judgment call rather
  than a diff read and it costs no attended moment.

**Two dispositions are hard reds rather than pending rows**, because both
encode an owner decision an agent must not absorb: a `ties` excluded
count that grew above the recorded floor, and a `DefaultCompat` field
that no longer matches its hypothesis's disposition in `oracle_bugs.json`.
The second is the M1-shaped cross-check worth having here, and it fires
the day Geoff moves a disposition and nobody moves the engine.

Rejected: a writer for `adjudications.json` in `internal/harness/analysis`
(the module has a loader and no writer anywhere, deliberately, and the
first writer should be lane E1's tool rather than a test); reporting
without failing (an unadjudicated mechanical divergence that fails
nothing is risk 2 with extra steps).

### R16: the stub gains `Unused()`, and the union rule is a hard failure (inventory Q17)

**Resolved: `stub.Server.Unused() []cassette.Request` ships, and tier 1
requires `Unused()` and `Misses()` both empty on every case. A
logged-but-never-issued request is a hard failure.**

3c deliberately kept the accessor out because 3c's grader enumerates the
request set itself and asserts a per-case issued count against the log,
so under-issuance already failed there (3c HISTORY guard 4, R6). 3d's
flow decides the request set from the matcher's own behavior, so that
assertion has no counterpart, and without `Unused()` a matcher that
issues fewer requests than the log records leaves zero misses and passes.

**A failure rather than a report**, because U8 makes MusicBrainz request
construction a conformance surface and the committed log is the
expectation. A run that skipped a lookup took a different flow than
pinned beets took, which is a matcher defect.

**`Extra` entries never count as unused.** `Unused()` reports unhit
entries of the request map alone. The spec's carried fact is why: a
strong ID match legally orphans search captures into `Extra` (spec
`:265`), so the request set is a recorded fixed point rather than a
stable set, and an unconsumed `Extra` entry is the channel working as
designed. Today no committed case carries an `Extra` entry, so the
distinction is dormant and is stated so a later pass does not collapse
it.

Its firing fixture is hand-authored in `internal/harness/stub/testdata`:
a two-request case where the test issues one request and asserts
`Unused()` names the other.

### R17: `make derive` does not run, and nothing in this pass needs it (inventory Q9's mechanism half)

**Resolved: no task runs `make derive`, and every fact that would
otherwise want a re-derivation is obtained another way.**

Three would-be triggers and their answers. R2's fourth pattern table is
confirmed by `ModeDefault`'s own tier-2 grading against the ten committed
`defect_free` figures. R10's `TieEntry.Adjudicator` carry-forward is
proven by `analysis.MergeTieDispositions`' own gate-resident test rather
than by a derivation run. R4's `lapjv` differential is frozen as
committed testdata generated by one offline probe.

The standing invariant is restated in Global Constraints so it does not
fall off, and Task 9 records that no committed byte moved.

### R18: the public entry point lives in `internal/engine/tag` (inventory Q19)

**Resolved: a third small package, `internal/engine/tag`, one file plus
its test, holding beets' `tag_album` and nothing else.**

The flow interleaves I/O with scoring: fetch by ID, score, and return
without searching when the result is strong (V18). It cannot be split
into fetch-everything-then-score, because the request set is exactly what
tier 1's zero-miss and unused rules grade. So the entry point needs a
package that sees both `matcher` and `musicbrainz`, and **`matcher`
cannot import `musicbrainz` because `musicbrainz` already imports
`matcher`** (`query.go`, `client.go`, `translate.go`).

`internal/engine/tag` keeps `matcher` pure, sits inside `internal/engine`
so `TestEngineNeverReadsTheClock` and `TestEngineNeverImportsHarness`
keep covering it, and reads correctly at rung 9's call site. It deviates
from the rung 3 spec's tree listing (`:163-185`), which names only
`matcher`, `musicbrainz` and `conform`; the deviation is stated here on
the precedent of 3b's m14 and 3c's own conductor-ruled tree-column
corrections, and Task 9 drafts the spec fix.

**The name sits beside rung 8's planned `internal/engine/tagio`**
(`specs/2026-09-02-greenfield-server-build.md:312`, "FLAC tag read and
write"), and the pairing is deliberate rather than a collision nobody
noticed (revision 2, fold m16; **provisional, conductor ruling owed**).
`tag` holds beets' `tag_album` and computes no applied field and touches
no file; `tagio` reads and writes FLAC tags and knows nothing about
matching. The two doc comments each say what the other is, so a reader
landing in one is told immediately which it is. The alternative the
conductor may prefer is a rename to `internal/engine/album`, which costs
the beets-faithful name at the one place the port most wants it.
**Recommended: keep `tag` and record the pairing**, and Task 9's drafted
corrections gain the build spec's tree listing, which revision 1 did not
name.

Rejected: `TagAlbum` inside `internal/engine/musicbrainz` (it costs that
package's stated charter in `doc.go`, which is the WS seam rather than
the matcher's flow); a `matcher.Source` interface the flow calls through
(go-conventions rejects an interface with one implementation, and the
package cycle is the real constraint rather than a testing one, since
`stub.Server` already makes the concrete client testable offline).

### R19: the 17 unledgered tied candidate assignments are a ledger gap, filed rather than fixed (inventory Q2)

**Resolved: 3d does not extend the tie detector. Tier 1 grades every
candidate's `Tracks` vector as the spec requires, a divergence on any of
the 17 is a tier-1 red naming the case and the candidate, and the pass
files the ledger gap for lane E1.**

`ties.json` asks the tie question of `Candidates[0]` only
(`internal/harness/oracle/ties.go:54`), so 80 of the 294 committed
candidate assignments were never measured. Sweeping every candidate with
the ledger's own forbid-and-re-solve test found **17 tied candidate
assignments carrying no tie disposition**, across the **five** cases
Task 9 carry 3 names (`real-noise-dylan-bootleg`,
`real-noise-pink-floyd-de1994`, `real-nonlatin-girls-generation`,
`real-nonlatin-utada-first-love`,
`real-nonlatin-utada-first-love-locale`) plus the non-best candidates of
`should-ask` and `should-reject`, and five of the seventeen produce a
different graded `Tracks` vector on an alternative optimum, two of those
also changing the candidate's own `tracks` penalty sum
(`real-nonlatin-utada-first-love` candidate 4,
`real-nonlatin-utada-first-love-locale` candidate 4). Revision 1 said
"six cases beyond the two pinned ones" here and named five in Task 9;
the case list above is now the single statement and the count follows it
(revision 2, fold m2).

**These figures are carried from the planning inventory, not measured at
this plan's drafting**, because reproducing them means re-solving every
candidate's cost matrix through a solver that does not exist yet.
Neither review lens could reproduce them for the same reason. **Task 7
re-measures them once its sweep runs** and Task 9 records the measured
values rather than these.

Extending the detector to every candidate is a derivation change plus an
owner disposition batch, which is corpus-lane work and outside 3d's wall
and its no-derive constraint. What 3d owes is honesty about the exposure:
if the Go solver mirrors `lapjv`, all 294 vectors pass and the 17 are
invisible; if it diverges on one, tier 1 reds with no ledger row
explaining why, and **the failure message says so**, naming the
candidate index and the fact that `ties.json` measured only
`Candidates[0]`. Task 9 files the gap for lane E1 with the six case ids.

### R20: `profile_hashes` stays open, and 3d adds nothing (inventory Q18)

**Resolved: carried, unchanged.** `corpus/manifest.json`'s
`profile_hashes` carries `locale` alone, so an edit to `stock.yaml` or
`estate.yaml` has no hash to fail against and `checkProfileKeyDrift`
catches a new key while missing a changed value. The manifest is lane
E1's pin file and rung 3's contested-surface row permits moving
`schema_version` alone. The item has been open since 2b and 3d neither
closes it nor makes it worse. Task 9 records it as still open.

## The inbound carries, closed or carried, one line each

Every item `docs/STATUS.md`'s "To 3d" list and 3c's, 3b's and 3a's
HISTORY carries address to this pass, with its disposition and the task
that takes it.

| Carry | Source | Disposition in 3d |
| --- | --- | --- |
| `Preferred.Media` and `Preferred.Countries` config-load validation | 3a to 3b/3d, re-owned to 3d alone by 3b | **Closed** (Task 3): `Tunables.Validate` compiles both lists and returns a wrapped `ErrInvalidTunable`, and the two component panics stay with a doc-comment sentence naming `Validate` as what makes them unreachable |
| The compat switch's pattern-table selection | 3b guard 10, 3c's "To 3d" | **Closed** (Task 4, Resolution R2): both hypotheses are `oracle-wrong`, so both switches default off, which no committed table expresses; `SDPatternsGuardedNoBracketDiscount` is the fourth table and `ModeDefault`'s ten `defect_free` rows are its proof |
| The `\b` regime divergence | 3b guard 11 | **Carried**, unchanged, now travelling with the query path's lowering and `_extract_id`'s `\w` as one three-member regime family (Task 9 carry 8) |
| Tie mirroring needing both `inclArtist` call shapes | 3a guard 7 | **Closed** (Tasks 1 and 2): `AssignItems` passes `false` unconditionally and `addAlbumTracks` already passes `info.VA`; the VA fixture proves a mutation to `true` fires, and no corpus case does |
| 3c's R1 seam: 3d owns the ID-match early return, `_recommendation`, `_add_candidate`, the explicit `search_ids` branch and `match_by_id`'s no-id/no-consensus branch | 3c guard 3 | **Closed for the first three** (Tasks 3, 4, 5), **declared as gaps 1 to 3** for the last two, which no committed case reaches; the stub, the zero-miss rule and the client are inherited unchanged |
| R6's stub blind side: nothing reports which logged `Request` entries a run consumed | 3c guard 4 | **Closed** (Task 7, Resolution R16): `stub.Unused()` ships and tier 1 requires `Unused()` and `Misses()` both empty per case, as a hard failure |
| R7's `Case.Extra` producer | 3c guard 5 | **Carried**, unchanged: a producer is `cmd/harness` work in lane E1's territory, zero of 215 cases carry an entry, and `Unused()` excludes `Extra` on purpose |
| R11's interim `AdminEmail` User-Agent contact | 3c guard 7c | **Carried**: it belongs beside the four-knob registration, post rung 1b. `tag.Album` takes a built client and composes no User-Agent |
| `vaLikely` exported the day `match_by_id`'s no-consensus branch needs it | 3c guard 13c | **Closed as "stays unexported"** (Task 5): that branch is corpus-unreachable and the search path computes `vaLikely` inside `AlbumSearchFor`, so no caller outside `musicbrainz` ever needs it. The reason is recorded so a later pass does not export it for symmetry |
| `MergePseudoRelease`'s aliasing hazard, re-owned to 3d and rung 9 | 3b guard 8, 3c guard 10 | **Carried to rung 9, with the trigger named**: `AlbumForID` closes it by construction and `tag.Album` inherits that call site unchanged; `slices.Clone` is owed the day a caller retains both values, which is rung 9's pipeline |
| The four-knob wall registration | 3a guard 2, 3c guard 2 | **Carried unless rung 1b landed at Task 0**: `internal/config/keys.go` is the file no lane edits until then, and `Threshold.Default` is a string with `ParseInt` the only parser |
| The retry path, ungraded by the corpus forever | 3c guard 8 | **Carried**: the stub answers 200 or 404 only, `cassette.Request` carries no status field, and the response-status extension stays declined for a fourth pass |
| `profile_hashes` covering one profile of three | open since 2b | **Carried** (Resolution R20): `corpus/manifest.json` is lane E1's pin file and 3d writes nothing there |
| Rung 3's ten declared corpus-coverage gaps | 3a three, 3b six, 3c one | **Read as a running total, never re-counted**: 3d adds eleven of its own (Task 9's reconciled list, the eleventh added by revision 2's fold M-f), taking rung 3 to twenty-one |
| The rung 15 re-run criterion | every rung 3 pass close | **Carried forward verbatim** (the Spec section above, and Task 9 carry 1) |
| Five 3b `code-simplifier` leftovers | 3b close | **Not this pass's**: none touches a surface 3d writes, and none blocks. Task 9 restates that they stay open |

## Test helpers named in this plan

Every test above calls helpers that live in the same package's `_test.go`
file. They are named here once so no code block reads as a reference to
something no task defines.

- `internal/engine/matcher`: `evalDate` (a package-level
  `time.Time` literal the tests share), `assignmentCost`,
  `bruteForceOptimum`, `loadAssignFixture`, `loadLapjvOptima`,
  `ignoredFixture`, `candidatesAtDistances`,
  `candidatesWithEqualDistances`, `candidateWithZeroValuedKey`,
  `candidateWithNonzeroTrackKey`.
- `internal/engine/tag`: `evalDate`, `taggingServer` (an
  `httptest.Server` over a `testdata` directory, returning the recorded
  path list), `clientFor`, `statusServer` (which sends `Retry-After: 0`
  beside the status, fold M-h), `earlyReturnItems`, `mediumIDItems`
  (fold M-e), `searchFallbackItems`, `itemsWithoutAlbumID`,
  `itemsWithSplitAlbumID`.
- `internal/harness/analysis`: `loadProbeRuleFixture`.
- `internal/harness/stub`: `get` (exists, `stub_test.go:41`, taking an
  ADDRESS rather than a server). There is no `loadFixture` helper in
  this package and none is added: the new test calls
  `cassette.LoadCase` inline, in `stub_test.go:154-158`'s own shape
  (fold m12).
- `internal/harness/conform`: `repoRoot` (exists,
  `component_test.go:25`), `repoRootPath` (its `*testing.T`-free
  sibling, returning `(string, error)`, needed because the shared
  `sync.OnceValues` sweep runs outside any one test's `t`;
  `repoRoot` is rewritten as a one-line wrapper over it so the walk has
  one implementation, fold m8), `loadEveryCase`, `loadPolicyFixture`,
  `gradeFixtureCase`, `fullCompat` and `noCompat` (the two
  `matcher.Compat` literals the policy table names, fold M-i).

`gradeUserAgent` (`internal/harness/conform/request.go:21`) is reused by
tier 1 rather than re-declared; it is already the package's own
invented, non-site-specific literal.

---

## Task 0: The staleness pre-task

Model: `sonnet`, **fresh context**, dispatched before any implementer.
Reads this plan and the rung 3 spec against the tree as shipped. **Its
output is blockers with evidence, not fixes**; it writes no code and
edits no file. The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's citations were verified at `c7a85fd`. A citation
that has moved is a finding for this task, not a surprise for an
implementer.

The checklist is the spec's own (`:395-406`), per pass. The first three
are blockers: a "no" stops the launch rather than producing a note.

1. **Have Phase 0a, 0b and 0c landed?** *(Expected: 0a landed; **0b and
   0c unlanded, which is the expected state and not a blocker for 3d**,
   since 3d registers nothing on the wall, writes no `internal/config`,
   and needs no worktree. Report the state; do not stop on it.)*
2. **Has rung 1b's registry shipped?** *(Expected: no.
   `internal/config/keys.go:19-25` still carries `ownedWall` as a slice
   literal. Confirm the four-knob deferral is still recorded in 3a's
   HISTORY guard 2 and 3c's guard 2, and that `internal/config` is
   untouched by this plan.)*
3. **Is the lane registry consistent with where this pass intends to
   run, and is the main tree actually free?** `docs/STATUS.md`'s lane
   registry is the live-executor check, never `pgrep`. *(Expected:
   lane-1 reads "3c closed ... Next: the 3d plan-prep workflow, then 3d
   execution". Flipping lane-1 to a 3d-execution entry carrying a pid is
   a conductor edit owed before Task 1 dispatches. Report any live pid in
   this tree, or a lane-1 entry naming another workflow, as a blocker.
   lane-2 BLOCKED; doc lane IDLE.)*
4. **Have the contested-surface rows been written for every surface this
   pass touches?** *(Expected: **five of nine present**, with
   `internal/engine/tag` missing entirely, the `internal/harness/oracle`
   and `corpus/analysis/*.json` rows not naming 3d's edits, and **no
   `docs/` ledgers row anywhere in `:156-175`** even though `:175`
   refers to one (fold m15). Report all four; the conductor writes
   them; none is a launch blocker for the reason this plan's "Contested
   surfaces" section states.)*
5. Does every corpus artifact this plan names exist, at the path and
   shape named? Specifically: all 215 case directories load through
   `cassette.LoadCase`; `corpus/analysis/{ties,proximity,adjudications,
   oracle_bugs}.json` all load through their `analysis` loaders;
   `corpus/manifest.json` carries `schema_version` `"4"` and `lap`
   `"0.5.13"`.
6. Does every harness and engine API this plan calls have the signature
   named? `cassette.LoadCase`, `cassette.Case`, `cassette.Expected`,
   `cassette.Candidate`, `cassette.Thresholds`, `cassette.Request`;
   `stub.New`, `(*stub.Server).Misses`, `Consumed`, `Addr`, `Close`;
   `analysis.TieEntry`, `TieLedger`, `LoadTieLedger`, `CheckTieLedger`,
  `ProximityLedger`, `LoadProximityLedger`,
   `AdjudicationEntry`, `AdjudicationLedger`, `LoadAdjudications`,
   `CheckAdjudications`, `CheckDefectFree`, `OracleBugLedger`,
   `LoadOracleBugs`, `CheckOracleBugs`, `AgentAdjudicator`, `Pending`,
   `OracleWrong`, `OracleCorrect`, `NoCorrectAnswer`, `TieExcluded`,
   `TiePinned`, `TieSkipped`, `ProximityFlagged`, `ProximitySkipped`;
   `matcher.Tunables`, `DefaultTunables`, `AlbumInfo`, `TrackInfo`,
   `ItemFields`, `ProjectItemFields`, `AlbumPair`, `DistanceForAlbum`,
   `DistanceForTrack`, `(*Distance).Distance`, `AddedKeys`, `PerKey`,
   `NonzeroKeys`, `NonzeroItems`, `Tracks`, `MostCommonTags`,
   `IsVariousArtistsName`, `BeetsSDPatterns`, `SDPatternsWordBoundary`,
   `SDPatternsNoBracketDiscount`; `musicbrainz.NewClient`,
   `ClientConfig`, `Unmetered`, `Release`, `SearchReleases`,
   `AlbumForID`, `AlbumSearchFor`, `ErrNotFound`, `ErrUnavailable`;
   `conform`'s unexported `mergeProfile` and `releaseLookupPrefix`, and
   its exported `CheckCaseWeights`, `CheckCaseThresholds`,
   `GradeRequests`, `RequestCaseFloor`.
7. **Have any `oracle_bugs.json` or `adjudications.json` dispositions
   moved since this plan was drafted, and does any movement change 3d's
   scope?** *(Drafting-time snapshot: adjudications 21 rows, 18
   `oracle-wrong`, 3 `oracle-correct`, 0 pending, 10 carrying a
   `defect_free` block; oracle_bugs 6 hypotheses, 4 `oracle-wrong`
   (`bug-ft-pattern` 6 probes, `bug-remix-parenthetical` 4,
   `bug-nonlatin-alias-selection` 4, `bug-assignment-rejection-gap` 6), 2
   `oracle-correct`, 0 pending.* **A movement here changes R1's table and
   is a blocker, not a note**, because the compat surface and tier 2's
   expectation source both read it. *Also confirm the B3 fact: every
   probe case of every `oracle-wrong` hypothesis resolves to an
   adjudication row by `subject.case_id`, and the two unresolved probe
   cases, `rec-strong-medium-boundary` and
   `real-nonlatin-utada-first-love`, both sit under `oracle-correct`
   hypotheses.)*
8. **Have any `ties.json` dispositions or `proximity.json` rows moved?**
   *(Expected: `ties.json` 215 rows, 213 with no disposition, 2 pinned
   (`should-ask`, `should-reject`), **0 excluded**, 1 skipped
   (`admit-zero-track-dropped`), and `grep -c adjudicator` returns 0.
   `proximity.json` 2 rows, 1 flagged (`rec-strong-medium-boundary`, on
   `strong_rec_thresh`), 1 skipped. A grown excluded count is a blocker,
   because R15 makes it a hard red and the floor is 0.)*
9. **Do the corpus's committed figures match this plan's own
   arithmetic?** Read and report, mechanically, reading no response body
   into context: `ls corpus/cases | wc -l` *(215)*; profile split
   *(178/36/1)*; candidate-bearing cases *(214)* and total candidates
   *(294)*; the candidate-count histogram *(193 at 1, 4 at 2, 9 at 5, 8
   at 6, 1 at 0)*; cases with a non-empty `extra_items` *(7, max 33)* and
   `extra_tracks` *(14, max 4)* and both *(0)*; non-identity assignments
   *(23)*; cases carrying a candidate-distance tie *(2)*; the verdict
   distribution *(167 strong/apply, 39 medium/skip, 4 low/skip, 5
   none/skip)*; candidates carrying `missing_tracks` or
   `unmatched_tracks` at exactly 0.0 *(**0** of 294)*; and the
   `max_rec` downgrade firing count *(**6**: `assign-extra-item`,
   `complete-combined-short-scrambled`, `complete-max-rec-boundary`,
   `complete-short-one`, `real-noise-pink-floyd-de1994`,
   `rec-max-rec-missing-tracks-downgrade`)*. A divergence on any of these
   is a finding, because Tasks 4, 7 and 8's floors and mutation counts
   rest on them.
10. Has `internal/config` changed shape? Have
    `internal/engine/matcher/{assign,candidate,verdict,compat}.go` or
    `internal/engine/tag/` appeared? *(Expected: no on all six.)*
11. **Does the pinned solver still reproduce the plan's two probes?**
    Run them offline with `tools/venv/.venv/bin/python`, reading the
    **third** return value the way beets does
    (`_, _, y = lap.lapjv(...)`, `match.py:66`), never the second:
    `lap.lapjv` on an all-zeros `n x n` matrix returns column-to-row
    `[1,0]`, `[1,2,0]`, `[1,2,3,0]`, `[1,3,2,4,0]`, `[1,4,3,2,5,0]` for
    n = 2 to 6. Reading `x` instead yields `[1,0]`, `[2,0,1]`,
    `[3,0,1,2]`, `[4,0,2,1,3]`, `[5,0,3,2,1,4]`, the inverse
    permutation, which is a misread rather than a divergence (revision
    2, fold B1). And the `n = max(rows, cols)` zero-pad model agrees
    with `extend_cost=True` on at least 200 random rectangular trials.
    `corpus/manifest.json` still pins `lap 0.5.13` and `numpy 2.5.2`. A
    divergence is a blocker, because Task 1's extension model and Task
    2's whole objective rest on it.
12. **Is `go test ./...`'s wall time still what Task 7 sizes against?**
    Run it once at `-count=1` and report the total and the
    `internal/harness` and `internal/harness/conform` figures. *(3c's
    close measured 15.081s total, `internal/harness` 11.255s,
    `internal/harness/conform` 11.230s.)*

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
beyond item 12's timing measurement; the task writes nothing.

---

## Task 1: `assign.go`, the exact solve, the extras, and their sort orders

Model: **`opus`** (see Execution model). Depends on Task 0.

**Files:**
- Create: `internal/engine/matcher/assign.go` + `assign_test.go`
- Create: `internal/engine/matcher/testdata/assign/` (the hand-authored
  matrices and expectations named below)

**Interfaces:**

- Consumes: `matcher.ItemFields` (`info.go:124-155`), `matcher.TrackInfo`
  (`info.go:73-105`), `matcher.Tunables` (`tunables.go:42`),
  `matcher.DistanceForTrack` (`track.go:45`),
  `(*matcher.Distance).Distance` (`penalty.go:96`).
- Produces:

```go
// internal/engine/matcher/assign.go

// A Pair is one item index matched to one track index.
//
// Indexes rather than values, because beets keys its mapping on object
// identity (id(item), derive_oracle.py:184-189) while Go ItemFields is a
// comparable struct value, so two identical items would collapse into
// one map key. The index form is also what cassette.Expected.Assignment
// grades (cassette.go:84).
type Pair struct {
	Item  int
	Track int
}

// An Assignment is one solved item-to-track matching, in beets' own
// output shape (autotag/match.py:53-81).
//
// Pairs is in ASCENDING TRACK INDEX order, not item order. beets builds
// its mapping by zipping the solver's track-indexed result with the
// track list (match.py:72-76), and distance() then walks that dict to
// build both dist.tracks and the "tracks" penalty list
// (autotag/distance.py:527-530). Float addition is not associative, so
// the order fixes both the graded Tracks vector and the summed
// distance, and a pair list in item order is wrong in two ways at once.
type Assignment struct {
	Pairs       []Pair
	ExtraItems  []int // item indexes, sorted by (disc, track, title)
	ExtraTracks []int // track indexes, sorted by (index, title)
}

// AssignItems solves beets' track assignment (autotag/match.py:53-81):
// the minimum-cost matching between items and a candidate's tracks,
// where cell (i, t) is the normalized track distance between them.
//
// The cost is DistanceForTrack with inclArtist FALSE for every cell,
// unconditionally, including on a various-artists release. beets'
// assignment objective uses track_distance's own default
// (autotag/distance.py:383-385) while the final per-track distances a
// user sees pass album_info.va (distance.py:529). The two objectives
// disagree exactly on VA releases, and no committed corpus case
// discriminates them, so the hand-authored fixture in assign_test.go,
// TestAssignItemsUsesInclArtistFalseOnAVariousArtistsRelease, is what
// proves this line right: the index component makes the identity the
// unique optimum with the artist excluded, and including the artist
// flips the unique optimum to the swap.
//
// Exactly min(len(items), len(tracks)) pairs are returned, so at most
// one of ExtraItems and ExtraTracks is non-empty.
//
// An empty items or tracks slice returns an Assignment with no pairs
// and the whole non-empty side as extras, never a panic. beets never
// reaches that shape: _add_candidate's first gate drops a zero-track
// candidate before assign_items runs (match.py:345-347), and an album
// with no items is not an import. The behavior is defined anyway so a
// later caller outside that gate finds an answer rather than an index
// out of range (fold m11).
func AssignItems(tun Tunables, items []ItemFields, tracks []TrackInfo) Assignment

// AlbumPairsFor projects an assignment onto the pairing DistanceForAlbum
// consumes, preserving track order (contract V3).
func AlbumPairsFor(a Assignment, items []ItemFields, info *AlbumInfo) []AlbumPair
```

The solver itself is unexported and takes no `Tunables`:

```go
// solveLAP returns, for each column, the row assigned to it, or -1.
// It reproduces lap.lapjv(costs, extend_cost=True) (match.py:66):
// a dense Jonker-Volgenant solve over the matrix padded to
// n = max(rows, cols) with ZERO, which is the extension model measured
// against the pinned solver (contract V5) rather than the
// (rows+cols, max+1) recipe derive_ties.py's docstring assumes, which
// disagrees with this lap build on 30% of rectangular ties.
func solveLAP(costs [][]float64) []int
```

**Behavior contract:**

- **The cost matrix** is `costs[i][t] =
  DistanceForTrack(tun, items[i], &tracks[t], false).Distance()`
  (contract V1). Each cell is normalized against its own denominator; no
  raw sum, no shared denominator.
- **The result is track-indexed** (contract V2). `solveLAP` returns a
  slice of length `len(tracks)`; entry `t` holds an item index or `-1`.
- **`Pairs` is built by walking the result in ascending track index**
  (contract V3), so `Pairs[k].Track` is strictly increasing.
- **`ExtraItems`** is every item index not appearing in `Pairs`, sorted
  by `(Disc, Track, Title)`; **`ExtraTracks`** is every track index not
  appearing in `Pairs`, sorted by `(Index, Title)` (contract V6). Both
  use `slices.SortStableFunc` with `cmp.Or(cmp.Compare(...), ...)`, never
  `slices.SortFunc` and never `sort.Slice`, and the doc comment says why:
  beets' pre-sort order is a Python `set` order over identity-hashed
  objects, so the committed records are reproducible only while the sort
  keys are total, and `extra_items`' totality is a fact of today's corpus
  rather than a guarantee.
- **The exact-optimum property is what this task proves**, not the tie
  choice. Task 2 owns the tie choice.
- **Ruling F does not apply to this task.** A failure here is an ordinary
  defect.

**Step 1: Write the failing exact-optimum tests**

```go
func TestSolveLAPFindsAnOptimum(t *testing.T) {
	tests := []struct {
		name     string
		costs    [][]float64
		wantCost float64
		wantPairs int // min(rows, cols)
	}{
		{"square, unique optimum", [][]float64{{0, 1}, {1, 0}}, 0, 2},
		{"square, swap is optimal", [][]float64{{1, 0}, {0, 1}}, 0, 2},
		{"more rows than columns", [][]float64{{0, 0.9}, {0.9, 0}, {0.5, 0.5}}, 0, 2},
		{"more columns than rows", [][]float64{{10, 0, 5}, {0, 10, 5}}, 0, 2},
		{"3x3 with a forced detour", [][]float64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}}, 5, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := solveLAP(tt.costs)
			rows := len(tt.costs)
			cols := len(tt.costs[0])
			if len(got) != cols {
				t.Fatalf("solveLAP returned %d entries, want one per column (%d)", len(got), cols)
			}
			matched := 0
			seen := make(map[int]bool)
			var total float64
			for col, row := range got {
				if row == -1 {
					continue
				}
				if row < 0 || row >= rows {
					t.Fatalf("column %d assigned row %d, out of range [0,%d)", col, row, rows)
				}
				if seen[row] {
					t.Fatalf("row %d assigned to two columns", row)
				}
				seen[row] = true
				matched++
				total += tt.costs[row][col]
			}
			if matched != tt.wantPairs {
				t.Errorf("matched %d pairs, want min(rows, cols) = %d", matched, tt.wantPairs)
			}
			if total != tt.wantCost {
				t.Errorf("assignment cost = %v, want the optimum %v", total, tt.wantCost)
			}
		})
	}
}

func TestSolveLAPMatchesBruteForceOnSmallMatrices(t *testing.T) {
	// Deterministic pseudo-random matrices with small integer costs, so
	// ties are common and an optimum is cheap to enumerate. The
	// assertion is on COST, never on which optimum is chosen: the tie
	// choice is Task 2's, and asserting it here would make this test
	// fail for the right solver and the wrong reason.
	rng := rand.New(rand.NewPCG(1, 2))
	for n := 2; n <= 6; n++ {
		for range 200 {
			costs := make([][]float64, n)
			for i := range costs {
				costs[i] = make([]float64, n)
				for j := range costs[i] {
					costs[i][j] = float64(rng.IntN(4))
				}
			}
			got := assignmentCost(costs, solveLAP(costs))
			want := bruteForceOptimum(costs)
			if got != want {
				t.Fatalf("n=%d: solveLAP cost %v, brute force optimum %v, matrix %v", n, got, want, costs)
			}
		}
	}
}
```

`assignmentCost` and `bruteForceOptimum` are test helpers in
`assign_test.go`; `bruteForceOptimum` enumerates permutations with
`slices.Permutations`-free recursion and is used nowhere else.

**Step 2: Run to verify it fails**

Run: `go test ./internal/engine/matcher/ -run 'TestSolveLAP' -count=1 -v`
Expected: FAIL, "undefined: solveLAP".

**Step 3: Write the solver and the assignment**

`solveLAP` pads to `n = max(rows, cols)` with zero, runs column reduction
with reduction transfer, augmenting row reduction, then Dijkstra-style
augmentation for any unassigned row, and maps the padded result back,
returning `-1` for any column whose assigned row is a pad row.
`AssignItems` builds the matrix per V1, calls `solveLAP`, builds `Pairs`
in track order per V3, and sorts the extras per V6.

**Step 4: Write the extras and track-order tests**

```go
func TestAssignItemsOrdersPairsByTrackIndex(t *testing.T) {
	// Three items whose optimum is the scramble {item 1 -> track 0,
	// item 2 -> track 1, item 0 -> track 3}: the shape
	// complete-combined-over-scrambled carries. Pairs must come back
	// [(1,0), (2,1), (0,3)], ascending by Track.
	items := []ItemFields{
		{Title: "delta", Track: 1, Length: 100},
		{Title: "alpha", Track: 2, Length: 100},
		{Title: "bravo", Track: 3, Length: 100},
	}
	tracks := []TrackInfo{
		{Title: "alpha", Index: 1, Length: 100},
		{Title: "bravo", Index: 2, Length: 100},
		{Title: "charlie", Index: 3, Length: 100},
		{Title: "delta", Index: 4, Length: 100},
	}
	got := AssignItems(DefaultTunables(), items, tracks)
	want := []Pair{{Item: 1, Track: 0}, {Item: 2, Track: 1}, {Item: 0, Track: 3}}
	if !slices.Equal(got.Pairs, want) {
		t.Errorf("Pairs = %v, want %v (ascending track index, match.py:72-76)", got.Pairs, want)
	}
	if len(got.ExtraItems) != 0 {
		t.Errorf("ExtraItems = %v, want none: three items, four tracks", got.ExtraItems)
	}
	if !slices.Equal(got.ExtraTracks, []int{2}) {
		t.Errorf("ExtraTracks = %v, want [2]", got.ExtraTracks)
	}
}

func TestAssignItemsSortsExtraItemsByDiscTrackTitle(t *testing.T) {
	// Four items, one track. Three items fall out, and their recorded
	// order is (disc, track, title), not input order and not solver
	// order.
	items := []ItemFields{
		{Title: "zulu", Disc: 2, Track: 1, Length: 100},
		{Title: "alpha", Disc: 1, Track: 9, Length: 100},
		{Title: "bravo", Disc: 1, Track: 2, Length: 100},
		{Title: "match", Disc: 1, Track: 1, Length: 100},
	}
	tracks := []TrackInfo{{Title: "match", Index: 1, Length: 100}}
	got := AssignItems(DefaultTunables(), items, tracks)
	if !slices.Equal(got.ExtraItems, []int{2, 1, 0}) {
		t.Errorf("ExtraItems = %v, want [2 1 0]: (1,2,bravo) < (1,9,alpha) < (2,1,zulu)", got.ExtraItems)
	}
}

func TestAssignItemsUsesInclArtistFalseOnAVariousArtistsRelease(t *testing.T) {
	// Two items and two tracks whose titles and lengths are symmetric,
	// so the only components that can move the optimum are the track
	// index and the track artist. Under inclArtist FALSE the index
	// component (distance.py:411-413, weight 1.0) makes the IDENTITY
	// the unique optimum, cost 0.0 against the swap's 0.3333. Under
	// inclArtist TRUE the artist component (weight 2.0,
	// tunables.go:272-273) makes the SWAP the unique optimum, 0.25
	// against the identity's 0.4444. So the fixture discriminates the
	// two objectives outright and depends on lapjv's tie rule nowhere.
	// Measured against pinned beets on these four objects. The mutation
	// that proves this test fires: pass true in AssignItems'
	// DistanceForTrack call.
	tun := DefaultTunables()
	items := []ItemFields{
		{Title: "same", Artist: "Ana Reed", Track: 1, Length: 200},
		{Title: "same", Artist: "Cole Vance", Track: 2, Length: 200},
	}
	tracks := []TrackInfo{
		{Title: "same", Artist: "Cole Vance", Index: 1, Length: 200},
		{Title: "same", Artist: "Ana Reed", Index: 2, Length: 200},
	}
	withArtist := AssignItems(tun, items, tracks)
	for _, p := range withArtist.Pairs {
		if cost := DistanceForTrack(tun, items[p.Item], &tracks[p.Track], true).Distance(); cost == 0 {
			t.Fatalf("pair %v scores 0 with the artist included; the fixture no longer discriminates the two objectives", p)
		}
	}
	// The assignment objective ignores the artist entirely, so the
	// index component alone decides and the identity is the unique
	// optimum at cost 0.0 (revision 2, fold B3: revision 1 asserted the
	// swap here, which is the answer an inclArtist-true matrix gives).
	want := []Pair{{Item: 0, Track: 0}, {Item: 1, Track: 1}}
	if !slices.Equal(withArtist.Pairs, want) {
		t.Errorf("Pairs = %v, want %v: the cost matrix must be built with inclArtist false (distance.py:383-385)", withArtist.Pairs, want)
	}
}
```

**Step 5: Record the corpus totality obligation in the doc comment**

`AssignItems`' doc comment gains one sentence: beets' `extra_items`
pre-sort order is a Python `set` order over identity-hashed objects,
outside `PYTHONHASHSEED`'s reach, so every committed record is
reproducible only while `(disc, track, title)` is total over that case's
extras, and it is total today on every committed case.

**The guard that enforces it is Task 7's**, not this task's, because it
reads the corpus and `internal/engine` never does
(`internal/harness/conform/arch_test.go:141-146`). This step exists so
Task 1's implementer knows the obligation is placed rather than dropped;
Task 1 writes the sentence and no test.

**Step 6: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/matcher/assign.go internal/engine/matcher/assign_test.go internal/engine/matcher/testdata/assign && git commit`

**Acceptance criteria:**

- `AssignItems` returns exactly `min(len(items), len(tracks))` pairs on
  every input, and at most one of `ExtraItems`, `ExtraTracks` is
  non-empty (contract V4).
- `Pairs` is strictly increasing in `Track` on every input (contract V3).
- `solveLAP`'s cost equals a brute-force optimum on 1000 small random
  matrices across `n` 2 to 6 (contract V5's optimality half).
- The cost matrix is built with `inclArtist` false unconditionally, and
  the VA fixture above proves a mutation to `true` fires.
- Both extras sort with `slices.SortStableFunc`.
- Tier 0 replay, the translation tier and the request tier all stay
  green; `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. `solveLAP` returns the identity permutation for square inputs →
   `TestSolveLAPFindsAnOptimum`'s "swap is optimal" row and
   `TestSolveLAPMatchesBruteForceOnSmallMatrices` both fail.
2. Build `Pairs` in item order instead of track order →
   `TestAssignItemsOrdersPairsByTrackIndex` fails, reporting
   `[(0,3) (1,0) (2,1)]`.
3. Pass `info.VA` (that is, `true`) into the cost matrix →
   `TestAssignItemsUsesInclArtistFalseOnAVariousArtistsRelease` fails.
   **No committed corpus case fires this mutation**, which is declared
   gap 10.
4. Sort `ExtraItems` by title alone →
   `TestAssignItemsSortsExtraItemsByDiscTrackTitle` fails, reporting
   `[1 2 0]`.
5. Pad the rectangular matrix with `max+1` rather than zero → the
   "more rows than columns" and "more columns than rows" rows still pass
   on cost, and **nothing in this task fails**. This mutation is
   deliberately reported as *not* caught here; Task 2's frozen `lapjv`
   table is what catches it, which is why that table is a deliverable
   rather than a convenience.

**Deliverable count: 4** (`solveLAP` with its zero-pad extension;
`AssignItems` with its inclArtist-false matrix and track-ordered pairs;
the two extras with their sort orders; `AlbumPairsFor`). Gate green;
commit. **Report actual token spend and the five mutations' outcomes,
including mutation 5's non-firing.**

---

## Task 2: `lapjv` tie mirroring, under ruling F

Model: **`opus`**. Depends on Task 1.

**This task carries ruling F's escalation trigger, verbatim** (spec
`:125-133`, Resolution R3):

> Two failed mirroring attempts stop the task and produce a written
> finding with the cost matrices and both solvers' optima, and Geoff
> rules the disposition.

An attempt is one design of the pivot order plus its run against the two
pinned cases and the frozen table. On the **second** failure the
implementer stops, writes
`docs/research/2026-09-05-rung-3d-tie-mirroring-finding.md`
with the contents R3 enumerates, marks `should-ask` and `should-reject`
as awaiting Geoff in its return, and **does not attempt a third design**.
The pass continues with Task 3. **The conductor never rules the
disposition.**

**Files:**
- Modify: `internal/engine/matcher/assign.go` (the solver's internals
  only; `AssignItems`' signature and behavior contract do not change)
- Modify: `internal/engine/matcher/assign_test.go`
- Create: `internal/engine/matcher/testdata/lapjv-optima.json`
- Create: `internal/engine/matcher/testdata/assign/should-ask-costs.json`
- Create: `internal/engine/matcher/testdata/assign/should-reject-costs.json`
- Create, only on the second failure:
  `docs/research/2026-09-05-rung-3d-tie-mirroring-finding.md`

**Interfaces:**

- Consumes: Task 1's `solveLAP`, `AssignItems`, `Assignment`, `Pair`.
- Produces: no new exported surface. The behavior change is that
  `solveLAP` now reproduces `lap` 0.5.13's choice among equal optima.

**Behavior contract:**

- **The two pinned matrices are committed as testdata**, extracted once
  offline. Each file is `{"case_id": ..., "rows": N, "cols": N, "costs":
  [[...]], "want": [...], "optimal_cost": ...}`, the last field being
  what `TestSolveLAPMirrorsLapjvOnThePinnedTiedCases` reads as
  `fx.OptimalCost` (fold m14), with every cost written as a `%.17g` decimal,
  which round-trips a float64 exactly, and `want` the committed
  `Expected.Assignment` re-expressed as the solver's track-indexed form.
  `should-ask` is 12x12 with optimum `1.7962962962962963`;
  `should-reject` is 14x14 with optimum `5.282407407407407`. Both figures
  cross-check against `corpus/analysis/ties.json`'s own `optimal_cost`
  for those rows.
- **`should-ask`'s tie is a single free transposition.** Rows 1 and 6
  agree bit for bit on the two contested columns: cells (1,t1) and
  (6,t1) both read `0.90740740740740744`, and (1,t6) and (6,t6) both
  read `0.88888888888888895`. `lapjv` chose the **swap**, `{0:0, 1:6,
  2:2, 3:3, 4:4, 5:5, 6:1, 7:7, 8:8, 9:9, 10:10, 11:11}`, and the one
  equally optimal alternative is the identity. **A Go solver that returns
  the obvious identity fails this row**, which is the sharpest available
  statement of the mirroring requirement.
- **`should-reject`'s tie is harder.** No two item rows are bit-identical;
  the ties come from equal sums across permutations. `lapjv` chose the
  identity everywhere except a three-cycle on items 5, 7 and 9, and
  forbidding each chosen pair in turn yields five equal-cost
  alternatives, all permutations within items `{1,3,5,7,9,11}` over
  tracks `{1,3,5,7,9,11}`.
- **`testdata/lapjv-optima.json` freezes the differential evidence**
  (Resolution R4). It carries at least 400 rows: square all-equal
  matrices for `n` 2 to 15, tie-rich small-integer matrices for `n` 2 to
  10, and rectangular shapes covering both orientations, each with
  `lap` 0.5.13's own answer. The file's header object records
  `"lap": "0.5.13"`, `"numpy": "2.5.2"` and the exact one-line command
  that regenerates it, on the `case-mapping-sweep.tsv:7-10` precedent.
  It is generated once, offline, with `tools/venv/.venv/bin/python`;
  **no tracked Python file is added**, so the gate's ruff leg is
  untouched.

**Step 1: Write the failing pinned-case test**

```go
func TestSolveLAPMirrorsLapjvOnThePinnedTiedCases(t *testing.T) {
	// The two cases corpus/analysis/ties.json pins. Both optima cost the
	// same, so an optimality assertion cannot discriminate them; only
	// the exact choice can, and reproducing lap.lapjv's choice is what
	// ladder entry 2's "pinned" disposition requires.
	for _, name := range []string{"should-ask", "should-reject"} {
		t.Run(name, func(t *testing.T) {
			fx := loadAssignFixture(t, name+"-costs.json")
			got := solveLAP(fx.Costs)
			if !slices.Equal(got, fx.Want) {
				t.Errorf("solveLAP chose %v, lap 0.5.13 chose %v; the costs are equal, so this is a pivot-order divergence rather than a wrong optimum", got, fx.Want)
			}
			if c, want := assignmentCost(fx.Costs, got), fx.OptimalCost; math.Abs(c-want) > 1e-12 {
				t.Errorf("assignment cost %v, ties.json records the optimum as %v", c, want)
			}
		})
	}
}

func TestSolveLAPMirrorsTheFrozenLapjvTable(t *testing.T) {
	table := loadLapjvOptima(t) // testdata/lapjv-optima.json
	if len(table.Rows) < 400 {
		t.Fatalf("frozen table carries %d rows, want at least 400: a shrunken table is a silently weakened check", len(table.Rows))
	}
	square, rect := 0, 0
	for _, row := range table.Rows {
		got := solveLAP(row.Costs)
		if !slices.Equal(got, row.Want) {
			t.Errorf("%s: solveLAP %v, lap %v", row.Name, got, row.Want)
		}
		if len(row.Costs) == len(row.Costs[0]) {
			square++
		} else {
			rect++
		}
	}
	if square == 0 || rect == 0 {
		t.Errorf("table covers %d square and %d rectangular shapes; both are required, because the extension model only shows on rectangular input", square, rect)
	}
}
```

**Step 2: Run to verify it fails**

Run: `go test ./internal/engine/matcher/ -run 'MirrorsLapjv|MirrorsTheFrozen' -count=1 -v`
Expected: FAIL. Task 1's solver returns an optimum whose tie choice is
its own, so at least `should-ask` fails, reporting the identity against
the recorded swap.

**Step 3: Attempt one, the pivot-order transliteration**

Transliterate the three dense routines `nm -D` names in the pinned
extension: column reduction with reduction transfer (`_ccrrt_dense`),
augmenting row reduction (`_carr_dense`), and Dijkstra augmentation
(`_ca_dense` with `find_path_dense` and `_scan_dense`). Every comparison
is a plain float64 comparison; Go's spec forbids silent fusing without an
explicit conversion, which helps, and the C++ side is not under our
control, which is the residual hazard the frozen table exists to expose.

**Step 4: Run both tests**

Run: `go test ./internal/engine/matcher/ -run 'MirrorsLapjv|MirrorsTheFrozen' -count=1 -v`
If both pass, go to Step 6.

**Step 5: Attempt two, then stop**

Attempt two revisits the pivot order against the frozen table's failing
rows, which name the shape and `n` that diverge. **If attempt two also
fails, the task stops.** The implementer then:

- writes `docs/research/2026-09-05-rung-3d-tie-mirroring-finding.md`
  carrying, per ruling F and Resolution R3: both cost matrices at
  `%.17g`; both solvers' optima and their costs for both cases; the
  committed `Expected.Assignment` for both; `ties.json`'s own
  `optimal_cost` and `difference` figures; the frozen table's per-shape
  agreement rate; and the quoted sentences from `ties.json`'s two pinned
  reasons stating that both optima produce the same album distance and
  the same recommendation;
- offers, without assuming, the disposition R3 names: re-dispositioning
  both rows from `pinned` to `excluded`, which costs two `assignment`
  fields and nothing else (both cases carry empty `extra_items` and
  `extra_tracks`) and moves the ties-excluded count from 0 to 2;
- reverts `solveLAP` to Task 1's exact solver, keeps
  `testdata/lapjv-optima.json` and both cost fixtures committed, and
  marks `TestSolveLAPMirrorsLapjvOnThePinnedTiedCases` and
  `TestSolveLAPMirrorsTheFrozenLapjvTable` with `t.Skip` carrying the
  finding's path in the skip message, which is the one place a
  conditional skip is legitimate under go-conventions because the
  condition is an owner ruling rather than an environment;
- reports the two pinned case ids to the conductor for
  `CaseGradeSummary.TiesAwaitingRuling`, which Task 7 reads.

**Step 6: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/matcher/assign.go internal/engine/matcher/assign_test.go internal/engine/matcher/testdata && git commit`

**Acceptance criteria, on success:**

- `solveLAP` reproduces `lap` 0.5.13's exact assignment on both pinned
  matrices and on every row of the frozen table.
- The frozen table carries at least 400 rows spanning square and
  rectangular shapes, with its provenance header.
- `AssignItems`' contract from Task 1 is unchanged, and Task 1's five
  tests still pass.

**Acceptance criteria, on escalation:**

- The finding exists at the named path with every item R3 enumerates.
- `solveLAP` is Task 1's exact solver, the gate is green, and the two
  mirroring tests skip with the finding's path in the message.
- The two pinned case ids are reported for `TiesAwaitingRuling`.
- **No `disposition` is written anywhere**, and `corpus/analysis/ties.json`
  is untouched by this task.

**Named mutations this task reports (success path only):**

1. Replace the pivot order with lowest-column-index-wins → `should-ask`
   fails (the solver returns the identity where the record has the swap)
   and `should-reject` fails.
2. Pad rectangular matrices with `max+1` rather than zero → the frozen
   table's rectangular rows fail while every square row passes, which is
   the discrimination Task 1's mutation 5 could not make.
3. Skip reduction transfer → the frozen table's tie-rich square rows
   fail at `n >= 5`, where the shift-by-one rule breaks.

**Deliverable count: 3** (the mirroring solver; the two pinned cost
fixtures with their track-indexed expectations; the frozen `lapjv` table
with its provenance header and its two-shape coverage assertion). Gate
green; commit. **Report actual token spend, how many attempts ran, and,
on escalation, the finding's path.**

---

## Task 3: `candidate.go`, admission, insertion order, the stable sort, and `Tunables.Validate`

Model: `sonnet`. Depends on Task 1 (it needs `Assignment` and
`AlbumPairsFor`). Independent of Task 2.

**Files:**
- Create: `internal/engine/matcher/candidate.go` + `candidate_test.go`
- Modify: `internal/engine/matcher/tunables.go` (add `ErrInvalidTunable`
  and `Tunables.Validate`)
- Modify: `internal/engine/matcher/tunables_test.go`
- Modify: `internal/engine/matcher/album_media.go:36-46` and
  `album_country.go:31-41` (doc comments on the two panics only; the
  panics stay)
- Create: `internal/engine/matcher/testdata/candidate/` (the `required`
  and `ignored` fixtures)

**Interfaces:**

- Consumes: `AssignItems`, `AlbumPairsFor`, `Assignment` (Task 1);
  `DistanceForAlbum` (`score.go:60`); `(*Distance).NonzeroKeys`
  (`penalty.go:166`); `Tunables` (`tunables.go:42`), `Preferred`
  (`:227-235`), `CompilePatternOption` (`penalty.go:368`).
- Produces:

```go
// internal/engine/matcher/candidate.go

// A Candidate is one admitted release with its solved pairing and its
// accumulated distance: beets' AlbumMatch (autotag/match.py:142-201).
type Candidate struct {
	Info       *AlbumInfo
	Dist       *Distance
	Assignment Assignment
}

// Candidates accumulates admitted candidates in insertion order under
// beets' own (data_source, album_id) identity, reproducing tag_album's
// results dict (autotag/match.py:414, hooks.py:166-168).
//
// It is a mutable accumulator rather than a pure function, and it earns
// that: the ID stage and the search stage add into the SAME set, so the
// duplicate rule spans both, tag_album reads the set's length between
// them (match.py:432), and the insertion order the stable sort falls
// back on is exactly what a rebuilt slice would lose. Add performs no
// I/O, so this type stays in matcher and matcher stays importable by
// internal/engine/musicbrainz.
type Candidates struct {
	order []identity
	byID  map[identity]Candidate
}

// identity is beets' Info.identifier: the cross-provider key
// (data_source, id) (hooks.py:166-168), matching cassette.Candidate's
// own Source and AlbumID pair (cassette.go:103-109).
type identity struct {
	Source  string
	AlbumID string
}

// NewCandidates returns an empty accumulator.
func NewCandidates() *Candidates

// Add scores info against items and admits it, reporting whether it was
// admitted. A drop is not an error: beets logs at debug and returns
// (match.py:345-375). It returns an error only for a malformed
// Tunables value Validate would also have caught.
//
// An admitted candidate whose identity is already present OVERWRITES
// the stored value and keeps the earlier insertion position, matching
// Python's own dict assignment at match.py:378-380. It is reachable
// only through an empty album id, because gate 2 drops a duplicate with a
// non-empty one (match.py:350), and it is why order is appended to only
// when the identity is new: appending on every Add would double-count
// the position and desynchronize order from byID. No committed corpus
// case reaches it (fold m7).
func (c *Candidates) Add(tun Tunables, items []ItemFields, info *AlbumInfo, evalDate time.Time) (bool, error)

// Len reports how many candidates have been admitted.
func (c *Candidates) Len() int

// All returns the admitted candidates in insertion order.
func (c *Candidates) All() []Candidate

// Sorted returns the admitted candidates by ascending distance, stably,
// so an exact distance tie keeps insertion order (match.py:326-328,
// Python's stable Timsort).
func (c *Candidates) Sorted() []Candidate
```

```go
// internal/engine/matcher/tunables.go

// ErrInvalidTunable reports a Tunables value that names something the
// matcher cannot resolve: a recommendation level, an action, a required
// field, or a preferred-media or preferred-country pattern.
//
// It exists because beets gets this validation from confuse and from
// re.compile, both of which raise mid-import: an unknown max_rec level
// raises inside _recommendation (match.py:313-320) and an unknown
// required tag raises AttributeError inside _add_candidate
// (match.py:357). A refusal at config load beats a crash mid-import,
// and that difference is a deliberate divergence recorded here.
var ErrInvalidTunable = errors.New("invalid tunable")

// Validate resolves every name-typed and pattern-typed field once, so a
// malformed configuration fails where an operator can see it rather
// than inside a scoring run. It reports every problem rather than the
// first.
//
// It is a config-load check and never a precondition the rest of this
// package assumes: Recommend and ActionFor re-resolve their own names
// and return their own errors, because a Tunables value can reach them
// without passing through here.
func (t Tunables) Validate() error
```

**Behavior contract:**

- **The five gates run in beets' order** (contract V7), and the doc
  comment on `Add` states that a port reordering gate 1 and gate 2
  changes the observable outcome on a zero-track duplicate while
  reordering 3 and 4 changes only cost.
- **Gate 2's duplicate rule is keyed on `identity`** and only applies
  when `info.AlbumID != ""` (contract V10, `match.py:350`).
- **Gate 3, `match.required`**, tests the Go zero value of the named
  field, resolves names through a thirteen-entry table over
  `AlbumInfo`'s own fields, and returns an error wrapping
  `ErrInvalidTunable` on a name the table does not carry (Resolution R6).
  The doc comment records the `is None` divergence, cites 3b's guard 2
  as the ground that no rung 3 source produces an empty-but-present
  field, and names the second-source condition that reopens it.
- **Gate 5, `match.ignored`, reads `Dist.NonzeroKeys()`** (contract V8),
  never `AddedKeys` and never `PerKey`. The doc comment names this as the
  m5 hazard's second site and says outright that tier 1 cannot catch a
  wrong choice here because `match.ignored` is corpus-unreachable.
- **`Sorted` uses `slices.SortStableFunc` with `cmp.Compare` on the
  distance**, and its doc comment names the two corpus cases that grade
  it (contract V11).
- **`Validate` compiles both `Preferred` pattern lists the way their own
  components compile them** (revision 2, fold m3), which closes 3a's
  carry addressed to 3b or 3d
  and re-owned to 3d alone by 3b ("the `Preferred.Media` and
  `Preferred.Countries` compile panics want a wrapped error at
  config-load time"). **The two panics at `album_media.go:41` and
  `album_country.go:36` stay**, because the component call sites have no
  error path, and each gains one doc-comment sentence recording that
  `Tunables.Validate` is the check that makes them unreachable in a
  configured system. That is the honest minimum: moving the compile into
  `Tunables` would mean caching compiled state on a value type that is
  copied everywhere. **The two lists compile differently and `Validate`
  must match each**: `compileMediaPatterns` compiles
  `` `(\d+x)?(`+pattern+`)` `` (`album_media.go:39`) while
  `compileCountryPatterns` compiles the pattern bare
  (`album_country.go:34`). A bare-valid pattern is always valid wrapped,
  so a bare check would still imply the panics are unreachable, and a
  pattern such as `a)(b` compiles wrapped and fails bare, so a bare
  check would reject a media configuration the component accepts.
  `Validate` calls the two existing helpers rather than
  `CompilePatternOption` directly, so the check and the panic are the
  same check.
- **`Validate` also parses** every `MaxRec` value through
  `ParseRecommendation` and `QuietFallback` and `NoneRecAction` through
  `ParseAction`. Those two parsers are Task 4's, so **Task 3 lands
  `Validate`'s `Required` and `Preferred` halves and Task 4 extends it**;
  Task 3's own test asserts the two halves it lands and Task 4's asserts
  the other two. Both halves are named here so neither is discovered.

**Step 1: Write the failing admission-order tests**

```go
func TestAddDropsAZeroTrackCandidateBeforeTheDuplicateCheck(t *testing.T) {
	// Gate 1 before gate 2 (match.py:345-352). A zero-track candidate
	// sharing an identity with an admitted one must be dropped by the
	// zero-track rule, so the admitted one survives untouched. A port
	// that checked duplicates first would drop it for the wrong reason
	// and report the same bool, which is why the assertion is on the
	// SURVIVING candidate rather than on the return value alone.
	c := NewCandidates()
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	good := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz",
		Tracks: []TrackInfo{{Title: "one", Index: 1, Length: 100}}}
	empty := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz"}
	if ok, err := c.Add(DefaultTunables(), items, good, evalDate); err != nil || !ok {
		t.Fatalf("Add(good) = %t, %v; want true, nil", ok, err)
	}
	if ok, err := c.Add(DefaultTunables(), items, empty, evalDate); err != nil || ok {
		t.Fatalf("Add(zero-track duplicate) = %t, %v; want false, nil", ok, err)
	}
	if c.Len() != 1 {
		t.Fatalf("Len = %d, want 1", c.Len())
	}
	if got := c.All()[0].Info; got != good {
		t.Error("the admitted candidate was replaced; the zero-track gate must run before the duplicate gate")
	}
}

func TestAddDropsADuplicateIdentity(t *testing.T) {
	c := NewCandidates()
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	tracks := []TrackInfo{{Title: "one", Index: 1, Length: 100}}
	first := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz", Tracks: tracks}
	same := &AlbumInfo{Album: "b", AlbumID: "id-1", DataSource: "MusicBrainz", Tracks: tracks}
	other := &AlbumInfo{Album: "c", AlbumID: "id-2", DataSource: "MusicBrainz", Tracks: tracks}
	noID := &AlbumInfo{Album: "d", AlbumID: "", DataSource: "MusicBrainz", Tracks: tracks}
	// A second empty-album-id candidate is admitted too (gate 2 guards
	// on info.album_id), and it OVERWRITES the first at the same
	// identity while keeping its position, exactly as Python's dict
	// assignment does at match.py:378-380. An accumulator that appends
	// to its order slice on every Add reads Len 4 here (fold m7).
	noID2 := &AlbumInfo{Album: "e", AlbumID: "", DataSource: "MusicBrainz", Tracks: tracks}
	for _, tt := range []struct {
		info *AlbumInfo
		want bool
	}{{first, true}, {same, false}, {other, true}, {noID, true}, {noID2, true}} {
		if ok, err := c.Add(DefaultTunables(), items, tt.info, evalDate); err != nil || ok != tt.want {
			t.Errorf("Add(%s) = %t, %v; want %t, nil", tt.info.Album, ok, err, tt.want)
		}
	}
	if c.Len() != 3 {
		t.Errorf("Len = %d, want 3: both empty-album-id candidates are admitted (match.py:350 guards on info.album_id) and the second overwrites the first at the same identity", c.Len())
	}
	if got := c.All()[2].Info; got != noID2 {
		t.Errorf("the third slot holds %q, want the second empty-album-id candidate: a repeated identity overwrites in place", got.Album)
	}
}

func TestAddAppliesRequiredAgainstTheZeroValue(t *testing.T) {
	tun := DefaultTunables()
	tun.Required = []string{"label"}
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	tracks := []TrackInfo{{Title: "one", Index: 1, Length: 100}}
	c := NewCandidates()
	missing := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz", Tracks: tracks}
	present := &AlbumInfo{Album: "b", AlbumID: "id-2", DataSource: "MusicBrainz", Label: "Harbour", Tracks: tracks}
	if ok, _ := c.Add(tun, items, missing, evalDate); ok {
		t.Error("a candidate with no label was admitted under required = [label]")
	}
	if ok, err := c.Add(tun, items, present, evalDate); err != nil || !ok {
		t.Errorf("Add(present) = %t, %v; want true, nil", ok, err)
	}
}

func TestAddRejectsAnUnknownRequiredFieldName(t *testing.T) {
	tun := DefaultTunables()
	tun.Required = []string{"nosuchfield"}
	c := NewCandidates()
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	info := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz",
		Tracks: []TrackInfo{{Title: "one", Index: 1, Length: 100}}}
	_, err := c.Add(tun, items, info, evalDate)
	if !errors.Is(err, ErrInvalidTunable) {
		t.Errorf("Add with required = [nosuchfield] returned %v, want an error wrapping ErrInvalidTunable; beets raises AttributeError here and crashes the import (match.py:357)", err)
	}
}

func TestAddAppliesIgnoredAgainstTheNonzeroView(t *testing.T) {
	// The m5 hazard's second site. The fixture is built so the candidate
	// carries "media" in the ADDED-key view at exactly 0.0 and carries
	// "year" nonzero. ignored = ["media"] must NOT drop it, because
	// beets iterates Distance.items(), which omits zero-weighted
	// penalties (distance.py:190-205, 240-241). The mutation that proves
	// this test fires: read AddedKeys() instead of NonzeroKeys().
	tun := DefaultTunables()
	tun.Ignored = []string{"media"}
	items, info := ignoredFixture(t) // testdata/candidate/ignored-zero-media.json
	c := NewCandidates()
	ok, err := c.Add(tun, items, info, evalDate)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !ok {
		t.Fatal("candidate dropped on a media penalty of exactly 0.0; ignored reads the nonzero view")
	}
	if !slices.Contains(c.All()[0].Dist.AddedKeys(), "media") {
		t.Fatal("the fixture no longer adds media at 0.0, so this test can no longer discriminate the two views")
	}
	if slices.Contains(c.All()[0].Dist.NonzeroKeys(), "media") {
		t.Fatal("the fixture's media penalty is nonzero, so this test can no longer discriminate the two views")
	}
}

func TestSortedIsStableOnAnExactDistanceTie(t *testing.T) {
	// Two candidates at identical distances, admitted in a known order.
	// slices.SortFunc in place of slices.SortStableFunc is the mutation,
	// and it fires here and on exactly two committed corpus cases,
	// should-ask and should-reject.
	c := candidatesWithEqualDistances(t, "first", "second", "third")
	got := make([]string, 0, 3)
	for _, cand := range c.Sorted() {
		got = append(got, cand.Info.AlbumID)
	}
	if !slices.Equal(got, []string{"first", "second", "third"}) {
		t.Errorf("Sorted = %v, want insertion order preserved across an exact tie (match.py:326-328)", got)
	}
}
```

**Step 2: Run to verify they fail**

Run: `go test ./internal/engine/matcher/ -run 'TestAdd|TestSorted' -count=1 -v`
Expected: FAIL, "undefined: NewCandidates".

**Step 3: Write `candidate.go` and `Tunables.Validate`'s first two halves**

**Step 4: Write the `Validate` tests**

```go
func TestValidateRejectsAnUnknownRequiredFieldAndABadPattern(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*Tunables)
		want string
	}{
		{"unknown required field", func(t *Tunables) { t.Required = []string{"nosuchfield"} }, "nosuchfield"},
		{"unparseable preferred media pattern", func(t *Tunables) { t.Preferred.Media = []string{"("} }, "preferred.media"},
		{"unparseable preferred country pattern", func(t *Tunables) { t.Preferred.Countries = []string{"[a-"} }, "preferred.countries"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tt.mut(&tun)
			err := tun.Validate()
			if !errors.Is(err, ErrInvalidTunable) {
				t.Fatalf("Validate = %v, want an error wrapping ErrInvalidTunable", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Validate = %v, want the message to name %q", err, tt.want)
			}
		})
	}
	if err := DefaultTunables().Validate(); err != nil {
		t.Errorf("DefaultTunables().Validate() = %v, want nil", err)
	}
}

func TestValidateReportsEveryProblemRatherThanTheFirst(t *testing.T) {
	tun := DefaultTunables()
	tun.Required = []string{"nosuchfield"}
	tun.Preferred.Media = []string{"("}
	err := tun.Validate()
	for _, want := range []string{"nosuchfield", "preferred.media"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Validate = %v, want the joined message to name %q", err, want)
		}
	}
}
```

**Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/matcher/candidate.go internal/engine/matcher/candidate_test.go internal/engine/matcher/tunables.go internal/engine/matcher/tunables_test.go internal/engine/matcher/album_media.go internal/engine/matcher/album_country.go internal/engine/matcher/testdata/candidate && git commit`

**Acceptance criteria:**

- The five gates run in beets' order and each has a test that fails when
  it is skipped.
- `Add` returns `(false, nil)` on every drop and never an error for a
  drop.
- `match.ignored` reads `NonzeroKeys()`, proven by the zero-media
  fixture.
- `Sorted` is stable, proven by the equal-distance fixture.
- `Tunables.Validate` reports every problem, wraps `ErrInvalidTunable`,
  and returns nil for `DefaultTunables()`.
- Both `Preferred` compile sites carry the doc-comment sentence naming
  `Validate`.
- Tier 0, the translation tier and the request tier stay green;
  `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Move gate 2 (duplicate) above gate 1 (zero tracks) →
   `TestAddDropsAZeroTrackCandidateBeforeTheDuplicateCheck` fails.
2. Drop the `info.album_id != ""` guard on gate 2 →
   `TestAddDropsADuplicateIdentity`'s `noID` row fails, `Len` reads 2.
3. Read `AddedKeys()` in place of `NonzeroKeys()` in gate 5 →
   `TestAddAppliesIgnoredAgainstTheNonzeroView` fails. **No committed
   corpus case fires this mutation**, which is why the fixture is
   hand-authored (declared gap: `match.ignored`, inherited from 3a).
4. `slices.SortFunc` in place of `slices.SortStableFunc` in `Sorted` →
   `TestSortedIsStableOnAnExactDistanceTie` fails, and Task 7's tier-1
   sweep fails on `should-ask` and `should-reject`.
5. Accept an unknown `Required` name silently (skip the table lookup) →
   `TestAddRejectsAnUnknownRequiredFieldName` fails.

**Deliverable count: 4** (the `Candidates` accumulator with the five
gates in order; the `required` filter with its name table and its
recorded `is None` divergence; the `ignored` filter on the nonzero view;
`Tunables.Validate`'s `Required` and `Preferred` halves with the two
panic doc comments). Gate green; commit. **Report actual token spend and
the five mutations' outcomes.**

---

## Task 4: `verdict.go` and `compat.go`

Model: `sonnet`. Depends on Task 3 (it extends `Tunables.Validate` and
reads `Candidate`).

**Files:**
- Create: `internal/engine/matcher/verdict.go` + `verdict_test.go`
- Create: `internal/engine/matcher/compat.go` + `compat_test.go`
- Modify: `internal/engine/matcher/tunables.go` (the fourth
  `SD_PATTERNS` table; `Validate`'s `MaxRec`, `QuietFallback` and
  `NoneRecAction` halves)
- Modify: `internal/engine/matcher/tunables_test.go`

**Interfaces:**

- Consumes: `Candidate` (Task 3), `(*Distance).NonzeroKeys` and
  `Distance.Tracks` (`penalty.go:166`, `:41`), `Tunables`,
  `ErrInvalidTunable` (Task 3), `BeetsSDPatterns`,
  `SDPatternsWordBoundary`, `SDPatternsNoBracketDiscount`
  (`tunables.go:348, 365, 388`).
- Produces:

```go
// internal/engine/matcher/verdict.go

// A Recommendation is beets' Recommendation (autotag/match.py:39-47).
// The values are ordered so that Go's min reproduces the max_rec
// downgrade's own min(rec, max_rec) (match.py:321).
type Recommendation int

const (
	RecNone Recommendation = iota
	RecLow
	RecMedium
	RecStrong
)

// String returns the name beets uses, which is also the name
// cassette.Expected.Recommendation carries.
func (r Recommendation) String() string

// ParseRecommendation resolves one of the four names beets' as_choice
// accepts (match.py:313-320), returning an error wrapping
// ErrInvalidTunable for anything else.
func ParseRecommendation(s string) (Recommendation, error)

// An Action is the decision beets takes unprompted
// (beets/importer/actions.py:8-19), lowercased as the shim records it
// and with "ask" standing for beets' None return
// (derive_oracle.py:206).
type Action string

const (
	ActionApply Action = "apply"
	ActionSkip  Action = "skip"
	ActionAsIs  Action = "asis"
	ActionAsk   Action = "ask"
)

// ParseAction resolves an action name to an Action, returning an error
// wrapping ErrInvalidTunable for anything else. It is the shared lexer
// and never the membership check: beets accepts only skip and asis for
// import.quiet_fallback (session.py:341) and skip, asis and ask for
// import.none_rec_action (session.py:348-350), so Validate and
// ActionFor test membership per key against those two sets. A shared
// four-name set would bless quiet_fallback: apply, which confuse
// raises on (fold M-d).
func ParseAction(s string) (Action, error)

// LegalQuietFallbacks and LegalNoneRecActions are those two sets, each
// citing the as_choice call it mirrors. They are exported so
// Tunables.Validate and ActionFor test the same membership and a
// reader can see that they differ on purpose.
var (
	LegalQuietFallbacks = []Action{ActionSkip, ActionAsIs}
	LegalNoneRecActions = []Action{ActionSkip, ActionAsIs, ActionAsk}
)

// Recommend is _recommendation (autotag/match.py:269-323). cands must be
// Candidates.Sorted's output; an empty slice returns RecNone.
//
// The three thresholds are asymmetric on purpose: strong compares with a
// strict < and medium with an inclusive <= (match.py:285, 288). That
// asymmetry is the whole reason corpus/analysis/proximity.json exists,
// so a case sitting exactly on a threshold admits no epsilon.
func Recommend(tun Tunables, cands []Candidate) (Recommendation, error)

// ActionFor is _summary_judgment
// (beets/ui/commands/import_/session.py:328-360).
//
// The quiet branch is tested FIRST, so Timid and NoneRecAction are
// unreachable whenever Quiet is set. Every committed corpus record was
// derived with quiet forced true (derive_oracle.py:204), so the corpus
// grades two of this function's five outcomes and the rest are graded
// by the table below.
func ActionFor(tun Tunables, rec Recommendation) (Action, error)
```

```go
// internal/engine/matcher/compat.go

// Compat selects, per ledgered defect class, whether the matcher
// reproduces pinned beets' defect or the adjudicated corrected
// behavior. A true field means "reproduce beets".
//
// Two classes and no more. corpus/analysis/oracle_bugs.json carries four
// oracle-wrong hypotheses, and only these two have a fix a port can
// express: bug-nonlatin-alias-selection's corrected verdict is an
// applied artist string rung 3 computes none of, and
// bug-assignment-rejection-gap needs a per-pair rejection mechanism
// beets does not have, which ruling A itself calls a fix with no beets
// counterfactual. Both are recorded as improvements-register items
// rather than shipped as switches nobody can flip.
type Compat struct {
	// FtPattern reproduces bug-ft-pattern: SD_PATTERNS' featuring
	// pattern matching "ft" inside an ordinary word such as "Draft".
	FtPattern bool
	// RemixParenthetical reproduces bug-remix-parenthetical: the two
	// 0.3-weight bracket discounts.
	RemixParenthetical bool
}

// FullCompat returns the selection that reproduces every ledgered
// defect, both switches on. It is a named engine surface rather than a
// bare struct literal in the harness because it IS a selection with a
// meaning: it is what internal/harness/conform's ModeCompat runs, the
// bit-identity claim against every committed record rests on it, and
// DefaultCompat is not its complement in any sense a reader should have
// to infer (fold m6).
func FullCompat() Compat

// DefaultCompat returns the selection corpus/analysis/oracle_bugs.json's
// own dispositions imply, as literals citing the hypothesis id per
// field. The engine never reads corpus/;
// internal/harness/conform cross-checks these two values against the
// ledger, and that check is what fires the day a disposition moves and
// nobody moves this function.
func DefaultCompat() Compat

// Apply returns tun with c's string-heuristic selection installed in
// StringDistPatterns. All four combinations name a real table, which is
// why SDPatternsGuardedNoBracketDiscount exists.
func (c Compat) Apply(tun Tunables) Tunables
```

```go
// internal/engine/matcher/tunables.go

// SDPatternsGuardedNoBracketDiscount is BeetsSDPatterns with BOTH
// adjudicated string-heuristic defects removed: the featuring pattern
// carries SDPatternsWordBoundary's \b guard and the two 0.3-weight
// bracket rows are absent, as in SDPatternsNoBracketDiscount.
//
// It exists because Tunables.StringDistPatterns is one slot and both
// hypotheses are ruled oracle-wrong, so ruling A defaults both switches
// off at once and neither committed table expresses that. It reproduces
// each committed defect_free figure because the two defects do not
// overlap on the ten probe cases: measured across all ten, no adv-ft-*
// compared string carries a bracket or parenthesis and no adv-remix-*
// compared string carries a featuring trigger, over the item titles,
// albums and artists and every title, name, artist, albumartist,
// sort-name and disambiguation string in their response bodies. The
// proof that survives is internal/harness/conform's ModeDefault run,
// which grades all ten against their committed figures.
//
// This name is deliberately NOT a legal analysis.LegalPatternTables
// value (adjudication.go:96): that field records which single-defect
// beets edit produced a recorded figure, and this table derived
// nothing.
//
// The \b regime divergence rides with it unchanged: Python re's \b is
// Unicode-aware and Go RE2's is ASCII-only, benign today because all
// six adv-ft-* pairs are ASCII (3b guard 11).
var SDPatternsGuardedNoBracketDiscount = []SDPattern{
	{Pattern: `^the `, Weight: 0.1},
	{Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
	{Pattern: `\b[\[\(]?(featuring|feat|ft)[\. :].+`, Weight: 0.1},
	{Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
}
```

**Behavior contract:**

- **`Recommend` implements contract V12's ladder exactly**, including the
  strict `<` and the inclusive `<=`, and **returns immediately on the
  no-conclusion branch** without running the downgrade (contract V13).
- **The downgrade key set is `cands[0].Dist.NonzeroKeys()` unioned with
  every entry of `cands[0].Dist.Tracks`'s own `Dist.NonzeroKeys()`** (contract
  V14). The union is walked in sorted order so a failure message is
  reproducible, even though `min` over the enum is commutative.
- **`ActionFor` implements contract V16's branch order**, quiet first.
- **`Validate` gains its `MaxRec`, `QuietFallback` and `NoneRecAction`
  halves**, and the two action keys carry **different legal sets**
  (revision 2, fold M-d). beets resolves `quiet_fallback` through
  `as_choice({"skip": SKIP, "asis": ASIS})`
  (`$SP/beets/ui/commands/import_/session.py:340-342`), two values, and
  `none_rec_action` through
  `as_choice({"skip": SKIP, "asis": ASIS, "ask": None})` (`:346-352`),
  three. A single four-value `ParseAction` would bless
  `quiet_fallback: apply` and `quiet_fallback: ask`, both of which
  confuse raises on, and `ActionFor` would then return a value beets can
  never produce on a configuration `Validate` approved. `ParseAction`
  stays as the shared lexer from name to `Action`; the membership check
  is per key, in `Validate` and again in `ActionFor`, each citing its
  own `as_choice` line.
- **`DefaultCompat` returns `{FtPattern: false, RemixParenthetical:
  false}` today**, because both hypotheses read `oracle-wrong`, and each
  field's literal carries its hypothesis id in a comment.

**Step 1: Write the failing recommendation tests**

```go
func TestRecommendWalksTheLadder(t *testing.T) {
	tun := DefaultTunables() // 0.04 / 0.25 / 0.25
	tests := []struct {
		name      string
		distances []float64
		want      Recommendation
	}{
		{"no candidates", nil, RecNone},
		{"below strong", []float64{0.039}, RecStrong},
		{"exactly strong is NOT strong", []float64{0.04}, RecMedium},
		{"exactly medium IS medium", []float64{0.25}, RecMedium},
		{"single candidate above medium", []float64{0.30}, RecLow},
		{"gap at or above rec_gap_thresh", []float64{0.30, 0.55}, RecLow},
		{"gap below rec_gap_thresh", []float64{0.30, 0.54}, RecNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Recommend(tun, candidatesAtDistances(t, tt.distances...))
			if err != nil {
				t.Fatalf("Recommend: %v", err)
			}
			if got != tt.want {
				t.Errorf("Recommend = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecommendReturnsBeforeTheDowngradeOnNoConclusion(t *testing.T) {
	// match.py:300-302 returns immediately, "Can't be downgraded any
	// further". A port that fell through still returns none, so the
	// only thing that can detect the difference is a max_rec that would
	// RAISE the value, which min() never does. The assertion is
	// therefore on the downgrade never being consulted: max_rec names a
	// key the candidate carries, and a strong level for it, and the
	// result stays none.
	tun := DefaultTunables()
	tun.MaxRec = map[string]string{"nosuchpenalty": "strong"}
	cands := candidatesAtDistances(t, 0.30, 0.54)
	got, err := Recommend(tun, cands)
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if got != RecNone {
		t.Errorf("Recommend = %v, want none: the no-conclusion branch returns before the downgrade", got)
	}
	// And the guard that keeps this test honest: an unresolvable max_rec
	// level on that same key must NOT be reported, because the branch
	// returned before reading max_rec at all.
	tun.MaxRec = map[string]string{"nosuchpenalty": "not-a-level"}
	if _, err := Recommend(tun, cands); err != nil {
		t.Errorf("Recommend = %v, want nil: the no-conclusion branch never reads max_rec", err)
	}
}

func TestRecommendDowngradesAgainstTheNonzeroViewOnly(t *testing.T) {
	// The m5 hazard's first site, and the corpus is blind to it: zero of
	// 294 committed candidates carry missing_tracks or unmatched_tracks
	// at exactly 0.0. This fixture builds one. The mutation that proves
	// it fires: read AddedKeys() instead of NonzeroKeys().
	tun := DefaultTunables() // max_rec: missing_tracks -> medium
	cand := candidateWithZeroValuedKey(t, "missing_tracks", 0.02)
	got, err := Recommend(tun, []Candidate{cand})
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if got != RecStrong {
		t.Errorf("Recommend = %v, want strong: a key added at 0.0 never downgrades (distance.py:246-247)", got)
	}
	if !slices.Contains(cand.Dist.AddedKeys(), "missing_tracks") {
		t.Fatal("the fixture no longer adds missing_tracks at 0.0, so this test can no longer discriminate the two views")
	}
}

func TestRecommendDowngradesAgainstATrackLevelKey(t *testing.T) {
	// match.py:307-309 unions every track sub-distance's keys into the
	// downgrade set. Under the shipped max_rec that loop is inert,
	// because neither missing_tracks nor unmatched_tracks is ever a
	// track-level key, and a model omitting it reproduces all 214
	// recorded recommendations. So this needs a hand-authored max_rec
	// naming a track key. The mutation: delete the union loop.
	tun := DefaultTunables()
	tun.MaxRec = map[string]string{"track_length": "low"}
	cand := candidateWithNonzeroTrackKey(t, "track_length", 0.01)
	got, err := Recommend(tun, []Candidate{cand})
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if got != RecLow {
		t.Errorf("Recommend = %v, want low: max_rec's key set unions every track sub-distance's nonzero keys", got)
	}
}
```

**Step 2: Run to verify they fail**

Run: `go test ./internal/engine/matcher/ -run 'TestRecommend' -count=1 -v`
Expected: FAIL, "undefined: Recommend".

**Step 3: Write `verdict.go`**

**Step 4: Write the verdict-to-action table**

```go
func TestActionForCoversEveryBranch(t *testing.T) {
	tests := []struct {
		name          string
		quiet         bool
		quietFallback string
		timid         bool
		noneRecAction string
		rec           Recommendation
		want          Action
	}{
		{"quiet and strong applies", true, "skip", false, "ask", RecStrong, ActionApply},
		{"quiet and medium takes the fallback", true, "skip", false, "ask", RecMedium, ActionSkip},
		{"quiet fallback asis", true, "asis", false, "ask", RecNone, ActionAsIs},
		{"quiet outranks timid", true, "skip", true, "ask", RecMedium, ActionSkip},
		{"timid always asks", false, "skip", true, "skip", RecStrong, ActionAsk},
		{"none_rec_action skip", false, "skip", false, "skip", RecNone, ActionSkip},
		{"none_rec_action asis", false, "skip", false, "asis", RecNone, ActionAsIs},
		{"none_rec_action ask", false, "skip", false, "ask", RecNone, ActionAsk},
		{"not quiet, not timid, not none, asks", false, "skip", false, "skip", RecMedium, ActionAsk},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.Quiet, tun.QuietFallback, tun.Timid, tun.NoneRecAction = tt.quiet, tt.quietFallback, tt.timid, tt.noneRecAction
			got, err := ActionFor(tun, tt.rec)
			if err != nil {
				t.Fatalf("ActionFor: %v", err)
			}
			if got != tt.want {
				t.Errorf("ActionFor = %q, want %q", got, tt.want)
			}
		})
	}
}
```

Six of these nine rows are corpus-unreachable and are declared gaps 7 and
8: every committed record was derived with `quiet` forced true and
`quiet_fallback` at `skip`, so the corpus reaches only "quiet and strong
applies" and "quiet and medium takes the fallback".

**Step 5: Write the compat tests**

```go
func TestApplySelectsATableForEveryCombination(t *testing.T) {
	tests := []struct {
		name string
		c    Compat
		want []SDPattern
	}{
		{"both on reproduces beets", Compat{true, true}, BeetsSDPatterns},
		{"ft fixed only", Compat{false, true}, SDPatternsWordBoundary},
		{"remix fixed only", Compat{true, false}, SDPatternsNoBracketDiscount},
		{"both fixed", Compat{false, false}, SDPatternsGuardedNoBracketDiscount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Apply(DefaultTunables()).StringDistPatterns
			if !slices.Equal(got, tt.want) {
				t.Errorf("Apply().StringDistPatterns = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSDPatternsGuardedNoBracketDiscountIsBothFixesAndNothingElse(t *testing.T) {
	// Byte-identity against its two parents, so a silent edit to any
	// pattern literal here fails rather than quietly changing ten
	// committed defect_free figures. The guarded featuring pattern comes
	// from SDPatternsWordBoundary; every other row comes from
	// SDPatternsNoBracketDiscount, in that table's own order.
	want := slices.Clone(SDPatternsNoBracketDiscount)
	for i := range want {
		if strings.Contains(want[i].Pattern, "featuring") {
			want[i] = SDPatternsWordBoundary[2]
		}
	}
	if !slices.Equal(SDPatternsGuardedNoBracketDiscount, want) {
		t.Errorf("SDPatternsGuardedNoBracketDiscount = %v, want %v", SDPatternsGuardedNoBracketDiscount, want)
	}
}

func TestDefaultCompatDefaultsBothSwitchesOff(t *testing.T) {
	// Ruling A's scheduling rule: a hypothesis at oracle-wrong defaults
	// its switch OFF. Both string-heuristic hypotheses read oracle-wrong
	// in corpus/analysis/oracle_bugs.json as of 2026-09-05.
	// internal/harness/conform cross-checks this against the ledger; the
	// engine never reads corpus/.
	if got := (DefaultCompat()); got != (Compat{}) {
		t.Errorf("DefaultCompat = %+v, want both switches off", got)
	}
}
```

**Step 6: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/matcher/verdict.go internal/engine/matcher/verdict_test.go internal/engine/matcher/compat.go internal/engine/matcher/compat_test.go internal/engine/matcher/tunables.go internal/engine/matcher/tunables_test.go internal/engine/matcher/testdata && git commit`

**Acceptance criteria:**

- All six of `_recommendation`'s exits are reached by the ladder table
  and each row's expectation matches contract V12.
- The no-conclusion branch returns before the downgrade, proven by both
  halves of its test.
- The downgrade reads the nonzero view and unions the track keys, each
  proven by its own hand-authored fixture.
- All nine `ActionFor` branches are covered, with the six
  corpus-unreachable ones named as declared gaps in the task report.
- `Apply` names a real table for all four combinations, and
  `SDPatternsGuardedNoBracketDiscount` is byte-identical to its two
  parents' composition.
- `Validate` now covers `MaxRec`, `QuietFallback` and `NoneRecAction`,
  with the two action keys tested against their own legal sets: a
  `quiet_fallback` of `apply` or `ask` is refused and a
  `none_rec_action` of `apply` is refused, each with its `as_choice`
  citation (fold M-d).
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Change the strong comparison from `<` to `<=` →
   `TestRecommendWalksTheLadder`'s "exactly strong is NOT strong" row
   fails, and **nothing in the corpus fails** (revision 2, fold M-f).
   `rec-strong-medium-boundary` records `distance =
   0.040000199999999986`, which its own `proximity` reason calls "2e-7
   above strong_rec_thresh", so both `<` and `<=` resolve it to medium;
   a sweep over all 215 cases finds zero candidates whose distance
   equals 0.04 or 0.25 exactly. The flagged row grades bit-identity, not
   the comparison direction, which is what Task 7's own mutation 4
   already says. The hand-authored ladder row is the sole detector, and
   Task 9 records it as declared gap 11.
2. Change the medium comparison from `<=` to `<` → the "exactly medium
   IS medium" row fails.
3. Fall through the no-conclusion branch into the downgrade →
   `TestRecommendReturnsBeforeTheDowngradeOnNoConclusion`'s second half
   fails, because `max_rec` is then read and its bad level reported.
4. Read `AddedKeys()` in place of `NonzeroKeys()` in the downgrade →
   `TestRecommendDowngradesAgainstTheNonzeroViewOnly` fails. **No
   committed corpus case fires this mutation**: zero of 294 candidates
   carry a `max_rec` key at 0.0 (declared gap 5).
5. Delete the track-key union loop →
   `TestRecommendDowngradesAgainstATrackLevelKey` fails, and **no
   committed corpus case fires it** (declared gap 6).
6. Move the `timid` check above the `quiet` check in `ActionFor` → the
   "quiet outranks timid" row fails, and **nothing in the corpus fails**
   (revision 2, fold M-g). The timid arm returns only when
   `config["import"]["timid"]` is truthy (`session.py:343-344`), and
   `timid` is false on every case: `estate.yaml:19` and `locale.yaml:19`
   set `timid: no`, `stock.yaml` sets nothing and beets ships `timid:
   no` (`$SP/beets/config_default.yaml:27`), and `DefaultTunables()`
   carries `Timid: false` (`tunables.go:293`). The unit row is the sole
   detector, consistent with declared gap 7.
7. Point `Apply`'s both-fixed arm at `SDPatternsWordBoundary` →
   `TestApplySelectsATableForEveryCombination` fails, and Task 8's
   `ModeDefault` run fails on the four `adv-remix-*` cases.

**Deliverable count: 4** (`Recommendation` and `Recommend` with the
ladder, the early return and the two-view downgrade; `Action` and
`ActionFor` with all nine branches; `Compat`, `DefaultCompat` and
`Apply`; `SDPatternsGuardedNoBracketDiscount` plus `Validate`'s
remaining three halves). Gate green; commit. **Report actual token spend
and the seven mutations' outcomes.**

---

## Task 5: `internal/engine/tag`, the matcher's public entry point

Model: `sonnet`. Depends on Tasks 3 and 4. Independent of Task 2.

**Files:**
- Create: `internal/engine/tag/doc.go`, `tag.go`, `tag_test.go`
- Create: `internal/engine/tag/testdata/` (the hand-authored cases for
  the four corpus-unreachable branches, plus `medium-id/`, the
  medium-scoring ID candidate that makes mutation 1 detectable, fold
  M-e)

**Interfaces:**

- Consumes: `matcher.ItemFields`, `MostCommonTags` (`likelies.go:47`),
  `Candidates`/`Candidate` (Task 3), `Recommend`, `ActionFor`,
  `Recommendation`, `Action`, `Compat`, `Tunables`, `Validate` (Tasks 3
  and 4); `musicbrainz.Client`, `AlbumForID` (`client.go:296`),
  `SearchReleases` (`:255`), `AlbumSearchFor` (`query.go:79`),
  `ErrNotFound` (`client.go:72`), `ErrUnavailable` (`:84`).
- Produces:

```go
// internal/engine/tag/tag.go

// ErrNoCandidates reports that the flow admitted nothing: every lookup
// 404'd, the search returned no ids, or every returned release was
// dropped by an admission gate. It is a sentinel because a caller
// branches on it: rung 9's pipeline reports "no match" differently from
// a transport failure.
var ErrNoCandidates = errors.New("no candidates")

// A Proposal is the matcher's answer for one album: beets' Proposal
// (autotag/match.py:247-249) with tag_album's two leading return values
// folded in, plus the action beets' own quiet-mode mapping would take.
type Proposal struct {
	CurArtist      string
	CurAlbum       string
	Candidates     []matcher.Candidate // ascending distance, stable
	Recommendation matcher.Recommendation
	Action         matcher.Action
}

// Options carries tag_album's optional inputs.
//
// SearchIDs is import.search_ids (beets/config_default.yaml:50), which
// selects the explicit-ID branch and takes no early return
// (match.py:417-420). beets' search_artist and search_name overrides are
// a recorded non-port (3c's close): no committed profile sets them and
// the automatic album-match path is rung 3's whole scope, so they never
// join this struct.
type Options struct {
	SearchIDs []string
}

// Album is tag_album (autotag/match.py:383-468): the matcher's whole
// public entry point.
//
// It issues one release lookup per item-consensus album id, scores what
// comes back, and returns WITHOUT searching when that match is strong
// and tun.Timid is false. Otherwise it searches, fans out over the
// returned ids in the search response's order, and returns every
// admitted candidate sorted by distance. The interleaving is
// load-bearing rather than incidental: the request set is exactly what
// internal/harness/conform's tier 1 grades, so a "fetch everything, then
// score" rearrangement would issue requests beets never issued.
//
// Every request Album issues spends one token from src's own budget, so
// a caller caps its MusicBrainz spend by the budget it built the client
// with (musicbrainz.NewClient requires one, client.go:128-130). A
// conformance run passes musicbrainz.Unmetered(). Rung 4's shadow passes
// a deferred background budget; rung 9's pipeline passes a batch one.
//
// evalDate is a parameter and never a clock read.
//
// It returns ErrNoCandidates when nothing was admitted. A lookup that
// 404s drops that candidate silently, reproducing album_for_id's return
// None (beetsplug/musicbrainz.py:777-781). Every other client error is
// returned wrapped, including musicbrainz.ErrUnavailable, which rung 4's
// shadow branches on to defer rather than record a divergence.
func Album(ctx context.Context, src *musicbrainz.Client, items []matcher.ItemFields,
	tun matcher.Tunables, compat matcher.Compat, evalDate time.Time, opts Options) (*Proposal, error)
```

**Behavior contract, step by step against contract V18:**

- **`tun = compat.Apply(tun)` first, then `tun.Validate()`.** A malformed
  tunable fails before any request is issued, so a misconfiguration
  never spends budget.
- `likelies, consensus := matcher.MostCommonTags(items)`;
  `CurArtist = likelies.Artist`, `CurAlbum = likelies.Album`.
- **If `opts.SearchIDs` is non-empty**, look each up through
  `src.AlbumForID` and add each; **take no early return**
  (`match.py:417-420`). Skip the search entirely.
- **Otherwise**: if `likelies.MBAlbumID != ""` and
  `consensus["mb_albumid"]`, look that one id up and add it. Beets'
  `match_by_id` returns `()` on either condition failing
  (`match.py:252-266`), and the Go port logs neither, because
  `internal/engine` has no logger; the doc comment records the two
  branches by name so a reader is not left wondering.
- **The early return**: `rec, err := matcher.Recommend(tun,
  cands.All())` on the **unsorted** insertion-order slice, per contract
  V18, with a comment saying beets does the same and that it is harmless
  only because at most one candidate exists here. If `cands.Len() > 0
  && !tun.Timid && rec == matcher.RecStrong`, return a `Proposal`
  carrying `cands.All()` unsorted, that `rec`, and `ActionFor(tun, rec)`.
- **The search**: `s := musicbrainz.AlbumSearchFor(items, tun)`, then
  `src.SearchReleases(ctx, s)`, then one `src.AlbumForID` per returned
  id **in response order** (3c contract C14), adding each. `vaLikely` is
  computed inside `AlbumSearchFor` (`query.go:108`) and stays unexported
  there, which is the resolution of 3c's carry 13c: the no-consensus
  branch is corpus-unreachable and no caller outside `musicbrainz` ever
  needs it, so it is **never exported** rather than exported for
  symmetry.
- **The close**: `sorted := cands.Sorted()`, `rec, err :=
  matcher.Recommend(tun, sorted)`, `action, err := matcher.ActionFor(tun,
  rec)`, return the `Proposal`.
- **Error policy.** `errors.Is(err, musicbrainz.ErrNotFound)` on a lookup
  drops that candidate and continues. Everything else is returned
  wrapped with `%w`. An empty `Candidates` at the close returns
  `ErrNoCandidates`, and the doc comment justifies the sentinel by
  naming rung 9 as the branching caller.

**Step 1: Write the failing flow tests**

Every test drives a `stub.Server`? No: `internal/engine` never imports
`internal/harness` (`arch_test.go:141-146`). The tests build a
`*musicbrainz.Client` against an `httptest.Server` the test owns,
serving hand-authored bodies from `internal/engine/tag/testdata`.

```go
func TestAlbumTakesTheIDMatchEarlyReturnAndIssuesNoSearch(t *testing.T) {
	// The 167-case shape: a unanimous mb_albumid whose lookup scores
	// strong. The assertion that matters is the REQUEST SET, because
	// that is what conform's tier 1 grades: exactly one lookup, no
	// search. The mutation that proves it fires: relax the rec ==
	// strong guard, which turns this into two requests.
	srv, paths := taggingServer(t, "early-return")
	defer srv.Close()
	p, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if p.Recommendation != matcher.RecStrong {
		t.Fatalf("Recommendation = %v, want strong", p.Recommendation)
	}
	if got := *paths; len(got) != 1 || !strings.HasPrefix(got[0], "/ws/2/release/") {
		t.Errorf("issued %v, want exactly one release lookup and no search", got)
	}
}

func TestAlbumSearchesWhenTheIDMatchIsExactlyMedium(t *testing.T) {
	// The detector for "relax rec == strong to rec >= medium", which
	// the early-return test above cannot see: its fixture scores strong
	// and a strong match stays on the same branch under either guard
	// (revision 2, fold M-e). This fixture's single ID candidate scores
	// medium, so the flow must fall through: one lookup, then one
	// search. Relaxing the guard turns that into one lookup and no
	// search.
	srv, paths := taggingServer(t, "medium-id")
	defer srv.Close()
	p, err := Album(context.Background(), clientFor(t, srv), mediumIDItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if p.Recommendation == matcher.RecStrong {
		t.Fatal("the fixture's ID candidate scores strong, so it can no longer discriminate the guard")
	}
	var searches int
	for _, path := range *paths {
		if strings.HasPrefix(path, "/ws/2/release?") {
			searches++
		}
	}
	if searches != 1 {
		t.Errorf("issued %v, want one lookup and one search: the early return needs rec == strong (match.py:436)", *paths)
	}
}

func TestAlbumFallsThroughToSearchWhenTheIDMatchIsNotStrong(t *testing.T) {
	srv, paths := taggingServer(t, "search-fallback")
	defer srv.Close()
	p, err := Album(context.Background(), clientFor(t, srv), searchFallbackItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if len(p.Candidates) < 2 {
		t.Fatalf("Candidates = %d, want the ID candidate plus at least one search candidate", len(p.Candidates))
	}
	var searches int
	for _, path := range *paths {
		if strings.HasPrefix(path, "/ws/2/release?") {
			searches++
		}
	}
	if searches != 1 {
		t.Errorf("issued %d searches, want exactly 1", searches)
	}
}

func TestAlbumTimidNeverTakesTheEarlyReturn(t *testing.T) {
	// import.timid is false in all three committed profiles, so this
	// branch is corpus-unreachable (declared gap 4). The hand fixture is
	// the early-return case with Timid set: the same strong ID match
	// must now also search.
	tun := matcher.DefaultTunables()
	tun.Timid = true
	srv, paths := taggingServer(t, "early-return")
	defer srv.Close()
	if _, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), tun, matcher.DefaultCompat(), evalDate, Options{}); err != nil {
		t.Fatalf("Album: %v", err)
	}
	if len(*paths) < 2 {
		t.Errorf("issued %v, want a search as well: timid forbids the early return (match.py:432)", *paths)
	}
}

func TestAlbumExplicitSearchIDsSkipTheSearchAndTakeNoEarlyReturn(t *testing.T) {
	// import.search_ids is set by no committed case (declared gap 1).
	// Two ids in, two lookups out, no search, and no early return even
	// when the first scores strong.
	srv, paths := taggingServer(t, "explicit-ids")
	defer srv.Close()
	p, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate,
		Options{SearchIDs: []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"}})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if len(p.Candidates) != 2 {
		t.Errorf("Candidates = %d, want 2", len(p.Candidates))
	}
	for _, path := range *paths {
		if strings.HasPrefix(path, "/ws/2/release?") {
			t.Errorf("issued a search under search_ids: match.py:417-420 skips it entirely")
		}
	}
}

func TestAlbumMatchByIDEarlyBranches(t *testing.T) {
	// Both branches are corpus-unreachable: all 215 committed cases
	// carry a full mb_albumid consensus (declared gaps 2 and 3). Each
	// must skip the lookup and go straight to the search.
	tests := []struct {
		name  string
		items []matcher.ItemFields
	}{
		{"no album id at all", itemsWithoutAlbumID()},
		{"no consensus on the album id", itemsWithSplitAlbumID()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, paths := taggingServer(t, "search-fallback")
			defer srv.Close()
			if _, err := Album(context.Background(), clientFor(t, srv), tt.items, matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{}); err != nil {
				t.Fatalf("Album: %v", err)
			}
			if strings.HasPrefix((*paths)[0], "/ws/2/release/") {
				t.Errorf("first request was a lookup (%s); match_by_id returns () here (match.py:252-266)", (*paths)[0])
			}
		})
	}
}

func TestAlbumDropsA404LookupAndPropagatesEverythingElse(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr error
	}{
		{"404 drops the candidate", http.StatusNotFound, ErrNoCandidates},
		{"503 propagates", http.StatusServiceUnavailable, musicbrainz.ErrUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// statusServer sends Retry-After: 0 alongside the status.
			// Without it the 503 row spends 31 seconds of real wall
			// time on every gate run: maxAttempts is 7
			// (client.go:42), 503 is retryable (:50-56), and the six
			// backoff waits are 0+1+2+4+8+16 (:418-423). The sleep
			// hook is unexported and ClientConfig exposes no way to
			// inject one (:93-99, :118), so the header is the only
			// lever a caller outside musicbrainz has, and retryWait
			// parses delta-seconds ahead of the backoff (:403-411).
			// Revision 2, fold M-h.
			start := time.Now()
			srv := statusServer(t, tt.status)
			defer srv.Close()
			_, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{SearchIDs: []string{"11111111-1111-1111-1111-111111111111"}})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Album = %v, want an error wrapping %v", err, tt.wantErr)
			}
			if elapsed := time.Since(start); elapsed > time.Second {
				t.Errorf("took %v; statusServer must send Retry-After: 0 so the retry schedule collapses (fold M-h)", elapsed)
			}
		})
	}
}
```

**Step 2: Run to verify they fail**

Run: `go test ./internal/engine/tag/ -count=1 -v`
Expected: FAIL, the package does not exist.

**Step 3: Write `doc.go` and `tag.go`**

`doc.go` states the package's one reason to exist, in the shape
`internal/engine/musicbrainz/doc.go` uses: this is the only package that
may see both `matcher` and `musicbrainz`, because `musicbrainz` already
imports `matcher` and the flow interleaves I/O with scoring, and it holds
`tag_album` and nothing else.

**Step 4: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/tag && git commit`

**Acceptance criteria:**

- The early return issues exactly one request on the strong-ID shape,
  and the medium-ID fixture issues one lookup plus one search, which is
  what makes mutation 1 detectable (fold M-e).
- The search fallback issues one search and one lookup per returned id,
  in response order.
- `statusServer` sends `Retry-After: 0` and the 503 row completes in
  under a second, asserted rather than assumed (fold M-h).
- All four corpus-unreachable branches have a hand-authored test and are
  named as declared gaps 1 to 4 in the task report.
- `ErrNoCandidates`, `musicbrainz.ErrNotFound` and
  `musicbrainz.ErrUnavailable` behave as the contract states, each proven.
- `compat.Apply` then `tun.Validate()` run before the first request.
- `TestEngineNeverImportsHarness` and `TestEngineNeverReadsTheClock`
  both still pass with the new package inside their walks.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Relax the early return's `rec == strong` guard to `rec >= medium` →
   `TestAlbumTakesTheIDMatchEarlyReturnAndIssuesNoSearch` **cannot see
   it**, because its fixture scores strong and a strong match stays on
   the same branch under either guard (revision 2, fold M-e; this is the
   third instance in rung 3 of the failure class R11 cites). The
   detector is the new medium-ID fixture named in Step 1,
   `TestAlbumSearchesWhenTheIDMatchIsExactlyMedium`, which asserts one
   lookup plus one search. Task 7's tier-1 sweep then fails on **39** of
   the 48 search-fallback cases as unused log entries: measured by
   joining each search-logging case's first release lookup to the
   candidate carrying that album id, 39 record an ID-candidate distance
   at or below 0.25 (six of them below 0.04, reaching medium only
   through the `max_rec` downgrade), 7 record one above 0.25, and 2
   carry no such candidate at all (`admit-zero-track-dropped`,
   `real-nonlatin-hamasaki-ayumi`). The remaining 9 still search.
2. Drop the early return entirely → the same test fails, and Task 7's
   sweep fails on the **167** early-return cases as stub misses.
3. Sort the candidates before the early return → nothing in this task
   fails and nothing in the corpus fails, because `match_by_id` yields at
   most one candidate. Reported as a deliberately unfired mutation and
   recorded at the call site, which is the honest treatment of contract
   V18's "harmless in practice".
4. Take the early return under `search_ids` →
   `TestAlbumExplicitSearchIDsSkipTheSearchAndTakeNoEarlyReturn` fails
   (one candidate instead of two).
5. Return the client's 404 as an error rather than dropping the
   candidate → `TestAlbumDropsA404LookupAndPropagatesEverythingElse`'s
   first row fails, reporting `ErrNotFound` where `ErrNoCandidates` is
   wanted.

**Deliverable count: 4** (`Album`'s ID-match, early-return and
search-fallback flow; the explicit `search_ids` branch and
`match_by_id`'s two early branches; the `Proposal` type with its
`Recommendation` and `Action`; the error policy with its three
sentinels). Gate green; commit. **Report actual token spend, the five
mutations' outcomes including mutation 3's non-firing, and the four
declared gaps.**

---

## Task 6: M4 and B3 in `internal/harness/analysis`

Model: `sonnet`. Depends on Task 0 only. Independent of Tasks 1 to 5.

**Files:**
- Modify: `internal/harness/analysis/ties.go` (the field, both checkers,
  and the new merge helper)
- Modify: `internal/harness/analysis/ties_test.go`
- Modify: `internal/harness/analysis/adjudication.go` (the exported
  subject index; `CheckDefectFree` uses it in place of its local map)
- Modify: `internal/harness/analysis/adjudication_test.go`
- Modify: `internal/harness/analysis/oraclebug.go` (the signature and the
  probe-case rule)
- Modify: `internal/harness/analysis/oraclebug_test.go` (the sixteen-row
  table's field, plus the new rows)
- Modify: `internal/harness/corpus_test.go` (the one production call
  site, `:397-406`)
- Modify: `internal/harness/oracle/ties_derive_test.go` (`:91`, and
  `writeTieLedger` at `:133-141`), both `//go:build oracle`
- Modify: `corpus/analysis/ties.json` (one field on two rows)
- Create: `internal/harness/analysis/testdata/oracle_bugs_probe_rule/`
  (`oracle_bugs.json` plus `adjudications.json`)

**Interfaces:**

- Consumes: `AgentAdjudicator` (`adjudication.go:30`), `Pending`,
  `OracleWrong` (`:17-22`), `TieExcluded`, `TiePinned`, `TieSkipped`
  (`ties.go:19-29`).
- Produces:

```go
// internal/harness/analysis/ties.go

type TieEntry struct {
	CaseID            string  `json:"case_id"`
	Status            string  `json:"status,omitempty"`
	AlbumID           string  `json:"album_id,omitempty"`
	OptimalCost       float64 `json:"optimal_cost"`
	BestForbiddenCost float64 `json:"best_forbidden_cost"`
	Difference        float64 `json:"difference"`
	Unique            bool    `json:"unique"`
	Disposition       string  `json:"disposition,omitempty"`
	Reason            string  `json:"reason,omitempty"`

	// Adjudicator names who set Disposition, required whenever
	// Disposition is set. The AgentAdjudicator sentinel is refused for
	// TieExcluded and ACCEPTED for TiePinned, and the asymmetry is
	// deliberate rather than an oversight: pinned keeps a case graded
	// while excluded stops grading it, so only the second can hide a
	// defect behind a disposition an agent chose. Requiring a human name
	// on pinned too would force an owner sitting immediately, because
	// the two committed pinned rows were authored by an agent (rung 3
	// spec, ruling F's M4 fold).
	Adjudicator string `json:"adjudicator,omitempty"`
}

// MergeTieDispositions carries the hand-written fields a fresh
// derivation cannot produce (Disposition, Reason and Adjudicator) from
// prior onto derived, matching by case id, and returns derived.
//
// It is untagged and lives here rather than beside the derivation
// because every file in internal/harness/oracle is //go:build oracle, so
// a helper there could never be exercised by scripts/check.sh.
// derive_ties.py emits six measured fields and never a disposition
// (internal/harness/oracle/derive_ties.py:118-126), so without this
// carry-forward `make derive -update-ledgers` silently erases every
// hand-written field, and the gate could not see it happen.
func MergeTieDispositions(derived []TieEntry, prior *TieLedger) []TieEntry
```

```go
// internal/harness/analysis/adjudication.go

// BySubjectCase indexes the ledger's case-subject rows by case id. A row
// whose subject is the album-plus-release-mbid shape rung 4's shadow
// writes is omitted, because it has no case to join on.
//
// It is exported because two checkers need it: CheckDefectFree's probe
// walk, which built this map privately, and CheckOracleBugs' probe-case
// rule (rung 3 spec, the B3 fold).
func (l *AdjudicationLedger) BySubjectCase() map[string]AdjudicationEntry
```

```go
// internal/harness/analysis/oraclebug.go

// CheckOracleBugs proves every row's probe case ids resolve to a
// committed corpus case, every adjudication_refs entry resolves to a row
// in adj, a row reproduced "yes" names at least one adjudication row,
// the same owner-only disposition rule CheckAdjudications enforces, and
// that EVERY probe case of a hypothesis at OracleWrong resolves to an
// adjudication row by subject.case_id.
//
// That last rule is the B3 fold. Oracle-bug rows are per hypothesis and
// adjudication rows are per case, so a hypothesis ruled oracle-wrong
// whose probe case carries no row would have no corrected verdict to
// produce and would be silently ungraded at the conformance gate. It
// takes the whole adjudication ledger rather than an id set because it
// needs both indexes, the row ids for adjudication_refs and the subject
// index for the probe rule, and building them here beats threading two
// maps through every caller.
func CheckOracleBugs(caseIDs []string, adj *AdjudicationLedger, l *OracleBugLedger) error
```

**Behavior contract:**

- **`checkTieEntry` gains two rules** (`ties.go:114-142`): `Adjudicator`
  is required whenever `Disposition` is set, and it must not equal
  `AgentAdjudicator` **when `Disposition == TieExcluded` only**.
- **`checkSkippedTieEntry` gains one** (`ties.go:148-159`):
  `Adjudicator` is forbidden on a skipped row, beside the existing
  `AlbumID` and `Disposition` prohibitions.
- **`corpus/analysis/ties.json`'s two pinned rows gain
  `"adjudicator": "agent"`.** Zero excluded rows exist, so this is the
  only value that is true and keeps the gate green, which is the outcome
  the asymmetry was chosen to produce. **No `disposition` is written or
  changed.**
- **`ties_derive_test.go` gains two lines**: `want.Adjudicator =
  got.Adjudicator` in the diff test's field mask at `:91`, and a call to
  `analysis.MergeTieDispositions` replacing `writeTieLedger`'s inline
  disposition-and-reason loop at `:133-141`. Both files are
  `//go:build oracle`, so the gate cannot run them; the untagged helper's
  own test is what proves the carry.
- **`CheckDefectFree` drops its local `rows` map** (`:224-229`) and calls
  `adj.BySubjectCase()`, which is behavior-preserving and removes the
  duplication B3 would otherwise create.
- **`CheckOracleBugs`' new rule** walks every hypothesis at
  `OracleWrong` and requires each `ProbeCaseIDs` entry to resolve in
  `adj.BySubjectCase()`. A probe with no row is a gate failure naming the
  hypothesis and the case id, never a silently ungraded case.

**Step 1: Write the failing tie-ledger table rows**

The table's landed shape is
`struct { name string; caseIDs []string; ledger TieLedger; wantHas string }`
and the loop calls `CheckTieLedger(tt.caseIDs, &tt.ledger)`
(`ties_test.go:38-41`), so each new row supplies a one-element
`caseIDs` and a one-element `TieLedger` (revision 2, fold M-c; revision
1's rows used an `entry` field the table does not have).

```go
// The shared base, beside the file's existing `tied` fixture.
excludedByAgent := tied
excludedByAgent.Disposition, excludedByAgent.Reason, excludedByAgent.Adjudicator = TieExcluded, "why", AgentAdjudicator

excludedByHuman := excludedByAgent
excludedByHuman.Adjudicator = "Geoff"

pinnedByAgent := tied
pinnedByAgent.Disposition, pinnedByAgent.Reason, pinnedByAgent.Adjudicator = TiePinned, "why", AgentAdjudicator

pinnedNoAdjudicator := tied
pinnedNoAdjudicator.Disposition, pinnedNoAdjudicator.Reason = TiePinned, "why"

skippedWithAdjudicator := skipped
skippedWithAdjudicator.Adjudicator = "Geoff"

// Added to ties_test.go's existing fourteen-row table.
{
	name:    "an excluded row adjudicated by the agent sentinel",
	caseIDs: []string{"clean"},
	ledger:  TieLedger{Entries: []TieEntry{excludedByAgent}},
	wantHas: "only a human name may exclude",
},
{
	name:    "an excluded row adjudicated by a human",
	caseIDs: []string{"clean"},
	ledger:  TieLedger{Entries: []TieEntry{excludedByHuman}},
},
{
	// The asymmetry's own proof: pinned keeps the case graded, so an
	// agent may set it, and the two committed pinned rows do.
	name:    "a pinned row adjudicated by the agent sentinel",
	caseIDs: []string{"clean"},
	ledger:  TieLedger{Entries: []TieEntry{pinnedByAgent}},
},
{
	name:    "a dispositioned row with no adjudicator",
	caseIDs: []string{"clean"},
	ledger:  TieLedger{Entries: []TieEntry{pinnedNoAdjudicator}},
	wantHas: "adjudicator: missing",
},
{
	name:    "a skipped row carrying an adjudicator",
	caseIDs: []string{"no-candidate"},
	ledger:  TieLedger{Entries: []TieEntry{skippedWithAdjudicator}},
	wantHas: "skipped, but carries adjudicator",
},
```

**And the existing `dispositioned` fixture at `ties_test.go:12-14` gains
`Adjudicator: AgentAdjudicator`.** It is `TiePinned` with a `Reason` and
no adjudicator, and its row at `:47-51` expects acceptance, so without
that one line the new requirement turns a passing row failing. This is
the seventh site R10's table names.

**Step 2: Write the failing merge-helper test**

```go
func TestMergeTieDispositionsCarriesEveryHandWrittenField(t *testing.T) {
	// derive_ties.py emits six measured fields and never disposition,
	// reason or adjudicator, so a fresh derivation's rows carry all
	// three empty. This is the whole reason make derive does not erase
	// the ledger, and both sites that would otherwise prove it are
	// //go:build oracle.
	derived := []TieEntry{
		{CaseID: "kept", AlbumID: "a", OptimalCost: 1, BestForbiddenCost: 1, Unique: false},
		{CaseID: "fresh", AlbumID: "b", OptimalCost: 1, BestForbiddenCost: 2, Difference: 1, Unique: true},
	}
	prior := &TieLedger{Entries: []TieEntry{
		{CaseID: "kept", Disposition: TiePinned, Reason: "both optima agree", Adjudicator: AgentAdjudicator},
		{CaseID: "gone", Disposition: TieExcluded, Reason: "retired", Adjudicator: "Geoff"},
	}}
	got := MergeTieDispositions(derived, prior)
	if len(got) != 2 {
		t.Fatalf("MergeTieDispositions returned %d rows, want 2: it carries fields, never rows", len(got))
	}
	if got[0].Disposition != TiePinned || got[0].Reason == "" || got[0].Adjudicator != AgentAdjudicator {
		t.Errorf("row %q = %+v, want disposition, reason and adjudicator all carried", got[0].CaseID, got[0])
	}
	if got[1].Disposition != "" || got[1].Adjudicator != "" {
		t.Errorf("row %q = %+v, want nothing carried: no prior row names it", got[1].CaseID, got[1])
	}
	if got[0].OptimalCost != 1 || got[0].BestForbiddenCost != 1 {
		t.Errorf("row %q measured fields were overwritten: %+v", got[0].CaseID, got[0])
	}
}
```

**Step 3: Write the failing B3 test**

```go
func TestCheckOracleBugsRequiresEveryOracleWrongProbeToResolve(t *testing.T) {
	// The rule fires on NOTHING in the committed corpus: every probe
	// case of every oracle-wrong hypothesis resolves today, and the two
	// probe cases with no adjudication row sit under oracle-correct
	// hypotheses. So its proof is this hand-authored pair, per the
	// spec's assertion discipline.
	tests := []struct {
		name    string
		fixture string // testdata/oracle_bugs_probe_rule/<name>
		wantHas string
	}{
		{"every oracle-wrong probe resolves", "resolved", ""},
		{"an oracle-wrong probe with no adjudication row", "unresolved", `probe case "probe-b": no adjudication row`},
		{"an oracle-correct probe with no row is fine", "correct-unresolved", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bugs, adj, caseIDs := loadProbeRuleFixture(t, tt.fixture)
			err := CheckOracleBugs(caseIDs, adj, bugs)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckOracleBugs: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantHas) {
				t.Fatalf("CheckOracleBugs = %v, want a message containing %q", err, tt.wantHas)
			}
		})
	}
}
```

**Step 4: Run to verify all three fail**

Run: `go test ./internal/harness/analysis/ -count=1 -v`
Expected: FAIL. The tie rows fail with "unknown field Adjudicator", the
merge test with "undefined: MergeTieDispositions", and the B3 test with
a signature mismatch on `CheckOracleBugs`.

**Step 5: Land the changes, including the two call sites**

`internal/harness/corpus_test.go:397-406` stops building `adjIDs` and
passes `adjLedger` directly. `oraclebug_test.go`'s sixteen-row table
swaps `adjudicationIDs map[string]bool` for `adj *AdjudicationLedger`,
with the two rows that carried `map[string]bool{"adj-1": true}` becoming
`&AdjudicationLedger{Entries: []AdjudicationEntry{{ID: "adj-1"}}}` and
the fourteen `nil` rows becoming `&AdjudicationLedger{}`.

**Step 6: Hand-add the two `adjudicator` fields**

`corpus/analysis/ties.json`: the `should-ask` and `should-reject` rows
gain `"adjudicator": "agent"`. Nothing else in the file changes, and the
diff is exactly two added lines.

**Step 7: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/harness/analysis internal/harness/corpus_test.go internal/harness/oracle/ties_derive_test.go corpus/analysis/ties.json && git commit`

**Acceptance criteria:**

- `CheckTieLedger` passes over the committed `ties.json` with the two
  `adjudicator` fields present and fails without them.
- All five new tie table rows behave as stated, including the pinned
  row accepting `agent`, and the existing `dispositioned` fixture
  carries an `Adjudicator` so no previously passing row flips (fold
  M-c).
- `MergeTieDispositions` carries all three hand-written fields and
  overwrites no measured field.
- `CheckOracleBugs` takes `*AdjudicationLedger`, both call sites compile,
  and the probe-case rule passes over the committed ledgers.
- `CheckDefectFree` uses `BySubjectCase` and its nineteen existing table
  rows plus its two `testdata/defect_free` fixtures still pass.
- The two `//go:build oracle` edits are present and the plan records
  that the gate cannot execute them, which is why the untagged helper
  exists.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Apply the sentinel to `pinned` as well as `excluded` → the committed
   `ties.json` reds, because both pinned rows carry `agent`. **This is
   the asymmetry's own proof and the reason the spec chose it** (spec
   `:133`).
2. Drop the `Adjudicator` requirement when `Disposition` is set → the
   "a dispositioned row with no adjudicator" row fails.
3. Have `MergeTieDispositions` carry `Disposition` and `Reason` but not
   `Adjudicator` → `TestMergeTieDispositionsCarriesEveryHandWrittenField`
   fails, which is exactly the silent erasure at the next `make derive`.
4. Restrict B3's walk to `StringHeuristicHypotheses` →
   the "an oracle-wrong probe with no adjudication row" fixture passes,
   because its hypothesis is not a string-heuristic one, and the test
   fails on the expected message. This mutation is why the rule lives in
   `CheckOracleBugs` rather than as a widening of `CheckDefectFree`.
5. Apply B3's rule to `oracle-correct` hypotheses too → the
   "an oracle-correct probe with no row is fine" row fails, and the
   committed corpus reds on `rec-strong-medium-boundary` and
   `real-nonlatin-utada-first-love`.

**Deliverable count: 4** (`TieEntry.Adjudicator` with its excluded-only
sentinel and its skipped-row prohibition; `MergeTieDispositions` with its
gate-resident test and the two oracle-tagged call sites;
`CheckOracleBugs`' probe-case rule with its signature change and both
call sites; `AdjudicationLedger.BySubjectCase` replacing
`CheckDefectFree`'s private map). Gate green; commit. **Report actual
token spend, the five mutations' outcomes, and the exact two-line diff
to `corpus/analysis/ties.json`.**

---

## Task 7: Tier 1, the whole-corpus mechanical runner, and the stub's `Unused()`

Model: **`opus`**. Depends on Tasks 1, 3, 4, 5 and 6. Task 2's outcome
feeds `TiesAwaitingRuling` and does not block dispatch.

**Files:**
- Create: `internal/harness/conform/case.go` + `case_test.go`
- Modify: `internal/harness/conform/doc.go` (the charter gains tier 1)
- Modify: `internal/harness/conform/defaults.go` (add `gradeTunables`)
- Modify: `internal/harness/conform/defaults_test.go` (R13's move onto
  the merged result)
- Modify: `internal/harness/conform/translate.go` (`mergeProfile`'s
  `estate` and `locale` arms gain the three `import.*` values those
  files set, fold M-b) and `translate_test.go` (one assertion per value
  against the profile file)
- Modify: `internal/harness/stub/stub.go` (`Unused()` alone) and
  `stub_test.go`
- Create: `internal/harness/stub/testdata/unused-fixture/`
- Create: `internal/harness/conform/testdata/case/` (the fixtures named
  below)

**Interfaces:**

- Consumes: `cassette.LoadCase`, `Case`, `Expected`, `Candidate`,
  `Thresholds`, `Request`; `stub.New`, `Addr`, `Misses`, `Consumed`,
  `Unused`, `Close`; `analysis.LoadTieLedger`, `LoadProximityLedger`,
  `LoadAdjudications`, `LoadOracleBugs` and their disposition constants;
  `matcher.DefaultTunables`, `ProjectItemFields`, `DefaultCompat`,
  `Recommendation`, `Action`; `tag.Album`, `Proposal`, `Options`;
  `musicbrainz.NewClient`, `ClientConfig`, `Unmetered`; `conform`'s own
  `mergeProfile` (`translate.go:193`).
- Produces:

```go
// internal/harness/stub/stub.go

// Unused returns the logged requests the run never issued, the
// complement of Misses.
//
// Misses alone cannot catch a client that issues FEWER requests than the
// case's log records: an unissued request leaves no trace at all, so a
// truncated flow passes a zero-miss check silently (3c's own carry to
// 3d). Extra entries are excluded on purpose: a strong ID match legally
// orphans search captures into Extra (rung 3 spec, :265), so an
// unconsumed Extra entry is the recall channel working as designed
// rather than a defect.
func (s *Server) Unused() []cassette.Request
```

```go
// internal/harness/conform/defaults.go

// gradeTunables is the configuration every conformance sweep runs
// under: Tunables' beets defaults merged with the case's profile, in
// that order (rung 3 spec, the M1 fold), with import.quiet forced true.
//
// The quiet overlay lives HERE and outside mergeProfile, because
// mergeProfile models the three committed profile FILES and
// checkProfileKeyDrift's allowlist is honest only while that stays
// true. That claim is load-bearing, so this pass makes it true:
// mergeProfile's estate and locale arms now set QuietFallback,
// NoneRecAction and Timid from those files, none_rec_action being the
// one whose value differs from beets' shipped default (fold M-b).
// quiet is in no profile: derive_oracle.py sets
// config["import"]["quiet"] = True before calling beets' own mapping
// (internal/harness/oracle/derive_oracle.py:204), because dubplate
// imports with `beet import -q` and quiet is a CLI flag. Without the
// overlay every committed action would grade as "ask".
func gradeTunables(profile string) (matcher.Tunables, error)
```

```go
// internal/harness/conform/case.go

// Floors for the whole-corpus sweep, measured off the tree at
// 2026-09-05. They are floors rather than exact counts because lane E1
// adds cases: a shrink means a case, a candidate or a comparison went
// missing, which is the failure worth catching, while growth is the
// corpus doing its job.
//
// AssignmentFloor is the 214 candidate-bearing cases, and it is never
// compared bare: ruling F's escalation moves the two pinned cases into
// TiesAwaitingRuling and an owner ruling can move a case into ties
// excluded, both of which legally shrink the population. The sweep
// compares against AssignmentFloor - len(TiesAwaitingRuling) -
// TiesExcluded, which is what the graded-field table has always said
// and what Task 8's verdict assertion already does (revision 2, fold
// B4). Comparing bare would red the gate on the one path ruling F
// pre-authorizes.
//
// CaseFloor is RequestCaseFloor (request.go:28), which is the same
// count of the same directories, so it is defined from it rather than
// re-typed (fold m5).
const (
	CaseFloor        = RequestCaseFloor
	CandidateFloor   = 294
	DistanceFloor    = 294
	AssignmentFloor  = 214
	VerdictFloor     = 215
	TrackVectorFloor = 294
)

// TiesExcludedCeiling is the number of ties.json rows dispositioned
// excluded today. It is a CEILING rather than a floor: excluding a case
// stops grading it, disposition is owner-only, and "the excluded count
// is reported at every run and may not grow without an owner ruling"
// (rung 3 spec, :292). Growth is a hard failure, never a pending row.
const TiesExcludedCeiling = 0

// A NearMiss is a comparison that matched within 1e-12 relative without
// being bit-identical. It is a reported finding rather than a failure,
// because accumulating them is how the port's remaining arithmetic
// drift stays visible (rung 3 spec, :141).
type NearMiss struct {
	CaseID string
	Field  string
	Want   float64
	Got    float64
}

// CaseGradeSummary reports what one sweep actually compared, never what
// it visited.
type CaseGradeSummary struct {
	Mode               GradeMode
	Cases              int
	Candidates         int
	Distances          int
	TrackVectors       int
	Assignments        int
	Verdicts           int
	TiesExcluded       int
	TiesAwaitingRuling []string // ruling F is owed on these; graded as pending, never as passing
	Pending            []string // case ids whose adjudication row leaves the verdict ungraded
	Ungraded           []string // probe cases of an oracle-wrong class with no expressible rung-3 fix
	NearMisses         []NearMiss
	VerdictMoves       []string // ModeDefault only; Task 8 fills it
}

// GradeMode selects which of ruling A's two runs a sweep is. Task 7
// implements ModeCompat; Task 8 implements ModeDefault and extends this
// comment with its reason.
type GradeMode int

const (
	// ModeCompat reproduces every ledgered defect: bit-identity against
	// every committed record, including the adversarial stratum. This is
	// the run tier 1 means.
	ModeCompat GradeMode = iota
	// ModeDefault takes the ledger's own dispositions (Task 8).
	ModeDefault
)

// GradeCases replays every committed case through the Go matcher's own
// public entry point and grades the result against the case's record,
// honoring each case's dispositions across all four analysis ledgers.
//
// It grades the request SET as well as the verdict: each case runs
// against a stub built from its own log, and a run leaving a non-empty
// Misses OR a non-empty Unused fails naming both. 3c graded request
// construction and handed the set here (3c's R1 seam), and the set is a
// function of the ID-match early return, _recommendation and
// _add_candidate, which this pass owns.
func GradeCases(root string, mode GradeMode) (CaseGradeSummary, error)

// distanceCompare names how one case's distances are compared, which is
// what corpus/analysis/proximity.json's own row decides.
type distanceCompare int

const (
	// compareEpsilon compares at 1e-12 relative and reports a match that
	// is not bit-identical as a NearMiss.
	compareEpsilon distanceCompare = iota
	// compareBitIdentical admits no tolerance at all, because a
	// flagged case sits on a comparison boundary where an epsilon can
	// change the verdict (rung 3 spec, :288).
	compareBitIdentical
)

// verdictSource names where a case's expected verdict comes from.
type verdictSource int

const (
	// verdictFromRecord is the case's own Expected.Recommendation and
	// Expected.Action.
	verdictFromRecord verdictSource = iota
	// verdictFromAdjudication is the corrected verdict on the case's own
	// adjudications.json row, resolved by subject.case_id, used once its
	// hypothesis's compat switch is off (rung 3 spec, :295).
	verdictFromAdjudication
)

// A casePolicy is how one case is graded, joined from all four analysis
// ledgers. It is the Go form of the rung 3 spec's disposition table
// (:285-299), and every field exists because one row of that table asks
// for it.
// Compat is the whole per-class selection rather than one bool,
// because "a switch is per defect class and cannot be half on" (spec
// :302) is a statement about EACH class: a case probing a pending
// bug-ft-pattern while bug-remix-parenthetical stays oracle-wrong needs
// {true, false}, which a single bool would render as {true, true} and
// bake into rung 15's reference behavior. It is dormant today only
// because oracle_bugs.json carries zero pending rows (revision 2, fold
// M-i).
type casePolicy struct {
	DistanceCompare    distanceCompare
	GradeAssignment    bool
	GradeAlbumDistance bool
	GradeVerdict       bool
	ExpectNoCandidate  bool
	Compat             matcher.Compat
	Excluded           bool
	Pending            bool
	VerdictFrom        verdictSource
}

// policyFor joins the four ledgers by case id and returns how the case
// is graded. It is the single place the spec's disposition table lives,
// so a row of that table has exactly one implementation and one fixture.
func policyFor(caseID string, ties *analysis.TieLedger, prox *analysis.ProximityLedger,
	adj *analysis.AdjudicationLedger, bugs *analysis.OracleBugLedger) casePolicy
```

**The graded-field table. This is what the fresh-context review reads
(Resolution R11), and what risk 2 names as the mitigation** (spec `:413`,
"an explicit graded-field table in the pass plan that a reviewer can
check field by field against `cassette.Expected`"). All eleven fields of
`cassette.Expected` (`internal/harness/cassette/cassette.go:76-88`):

| `Expected` field | Graded how | Population | Tier |
| --- | --- | --- | --- |
| `Recommendation` | `Proposal.Recommendation.String()` compared exactly | 215 | 2 (Task 8) |
| `Action` | `Proposal.Action` compared exactly, under the `Quiet: true` overlay | 215 | 2 (Task 8) |
| `EchoedDate` | must equal the case's own `EvalDate`, and that value is what is passed to `tag.Album` as `evalDate` | 215 | 1 |
| `ConfigHash` | **not graded.** It hashes the shim's effective beets config, which has no Go analogue; the Go run's configuration is graded instead through `Weights` and `Thresholds` | 0 | recorded non-grade |
| `Weights` | `CheckCaseWeights` against `gradeTunables(profile).DistanceWeights`, twenty keys per case | 215 | 1 |
| `Thresholds` | `CheckCaseThresholds` against `gradeTunables(profile)`, three floats plus `MaxRec` | 215 | 1 |
| `Candidates` | matched in order by `(Source, AlbumID)`, then per candidate: `Distance`, `Penalties` (the added-key view, key set and values), `Tracks` (elementwise) | 294 candidates | 1 |
| `Assignment` | `map[itemIndex]trackIndex` built from `Candidates.Sorted()[0].Assignment.Pairs` | 214, less `TiesAwaitingRuling` and less `ties` excluded | 1 |
| `ExtraItems` | elementwise against `Sorted()[0].Assignment.ExtraItems` | 215 (7 non-empty) | 1 |
| `ExtraTracks` | elementwise against `Sorted()[0].Assignment.ExtraTracks` | 215 (14 non-empty) | 1 |
| `AppliedFields` | **not graded.** Rung 3 computes no applied field; `corpus/analysis/file_tags.json` is rung 8's ledger and `internal/harness/analysis`' `AlbumInfoLedger` is 3b's | 0 | recorded non-grade, spec `:248` |

Two fields are deliberately not graded and both are recorded as
non-grades rather than passing checks, per the spec's M8 discipline.

**The per-case grading policy, from the spec's disposition table**
(`:285-299`). One resolver, `policyFor`, joins four ledgers by case id
and returns the policy; each row below is its own switch arm with its own
fixture.

| Signal | What the runner does |
| --- | --- |
| `proximity` clear | distances compared at 1e-12 relative; a match that is not bit-identical is appended to `NearMisses` |
| `proximity` flagged | distance must be **bit-identical**; no epsilon, no pending escape. One row today, `rec-strong-medium-boundary` |
| `proximity` skipped | the record carries no candidate; the Go run must also produce none. One row today, `admit-zero-track-dropped` |
| `ties` unique (no disposition) | assignment compared exactly |
| `ties` pinned | assignment compared exactly. Two rows today, and if ruling F is owed they move to `TiesAwaitingRuling` instead |
| `ties` excluded | assignment and album distance not graded; the verdict still is; `TiesExcluded` incremented and checked against `TiesExcludedCeiling`. **Zero rows today**, so its fixture is hand-authored |
| `ties` skipped | no assignment to grade |
| `oracle_bugs` `pending-adjudication` | compat switch on **for that class alone**, the other class keeping its own ledger's answer; the case grades like any other. **Zero rows today**, hand-authored fixture, and two of them: both classes pending, and one class pending (fold M-i) |
| `oracle_bugs` `oracle-wrong` | compat switch off; the corrected verdict comes from the case's own `adjudications.json` row, resolved by `subject.case_id` (Task 8) |
| `oracle_bugs` `no-correct-answer` | switch on, no probe case's verdict graded, distances still graded at the tier the case's own `proximity` row sets. **Zero rows today**, hand-authored fixture |
| `oracle_bugs` `oracle-correct` | grades like any other case |
| `adjudications` `no-correct-answer` | verdict not graded, distance still is. **Zero rows today**, hand-authored fixture |
| `adjudications` `pending-adjudication` | verdict not graded, case id appended to `Pending`, gate stays green, launch gate refuses. **Zero rows today**, hand-authored fixture |

**The precedence rule, implemented as the spec states it** (`:302`):
**the hypothesis row decides behavior; the case row supplies the
expectation.** `oracle_bugs.disposition` alone sets whether a compat
switch is on, because a switch is per defect class and cannot be half
on. `adjudications.disposition` with `corrected_verdict` alone supplies
what a case must produce once that switch is off. So a case row ruled
`oracle-wrong` under a hypothesis still at `pending-adjudication`
changes no behavior; and a probe case whose own row is `oracle-correct`
under a hypothesis ruled `oracle-wrong` grades against its own row.
`policyFor` has a test row for each of those two shapes.

**Behavior contract:**

- **Per case, in this order.** Load the case; start `stub.New` and close
  it before the next case, so peak memory is one case's bodies rather
  than the corpus's (3c's own discipline, and the largest single body is
  8,894 KB against 77.6 MB total); build a client with `BaseURL: "http://"
  + srv.Addr()`, `UserAgent: gradeUserAgent`, `Budget:
  musicbrainz.Unmetered()`; resolve `tun, err := gradeTunables(c.Profile)`;
  install `mode`'s compat selection, which is `matcher.FullCompat()` for
  `ModeCompat` and `matcher.DefaultCompat()` for `ModeDefault` (fold
  m6), overridden per case by `policyFor`'s own `Compat` where a
  hypothesis is pending (fold M-i); parse `c.EvalDate`; call
  `tag.Album(ctx, client, matcher.ProjectItemFields`-projected items,
  `tun, compat, evalDate, tag.Options{})`.
- **The request-set rule, and it is `ModeCompat`-only**: after the call,
  `srv.Misses()` **and** `srv.Unused()` must both be empty. A non-empty
  `Misses` names every unmatched `PathQuery`; a non-empty `Unused` names
  every logged request the flow never issued. Both fail the case
  (Resolution R16). `srv.Consumed()` is reported, not asserted, because
  zero committed cases carry an `Extra` entry. **The rule does not reach
  `ModeDefault`**, and the reason is stated rather than left implicit
  (revision 2, fold B2): a corrected verdict legally changes the flow,
  so the committed request log is an expectation of the compat run
  alone, and `ModeDefault` re-scores each case's committed candidate set
  instead of issuing requests at all.
- **`tag.ErrNoCandidates` is not automatically a failure.** One committed
  case, `admit-zero-track-dropped`, produces no candidate by design and
  its `proximity` row reads `skipped`. The runner treats
  `ErrNoCandidates` as "zero candidates" and grades that against the
  record's own candidate count.
- **Candidate matching is by `(Source, AlbumID)` in order**, so a
  reordering fails with a readable message rather than a cascade of
  value mismatches.
- **Every comparison that fails appends to a joined error** rather than
  returning the first, so one systemic defect reports its full count in
  one run and the mutation figures below are readable off the failure.
- **`TiesAwaitingRuling`** is populated from a small literal list the
  task carries only if Task 2 escalated, and the gate test fails when the
  slice is non-empty and Task 2's finding document does not exist
  (Resolution R3).
- **R13's move.** `defaults_test.go:43, 55-61` compares the echo against
  `matcher.DefaultTunables()` directly; it moves onto `gradeTunables`, so
  the M1 cross-check reads against the merged result the spec specifies
  (`:279`). The two are equal today because no committed profile sets a
  `match:` or `musicbrainz:` key, so the move is behavior-preserving and
  closes the drift before a fourth profile makes it real.

**Step 1: Write the failing `Unused` test**

```go
// internal/harness/stub/stub_test.go
func TestUnusedReportsALoggedRequestTheRunNeverIssued(t *testing.T) {
	// cassette.LoadCase inline, in stub_test.go's own landed shape
	// (:154-158); there is no loadFixture helper in this package, and
	// get takes an address rather than a server (:41). Fold m12.
	dir := filepath.Join("testdata", "unused-fixture") // two requests, one extra
	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase(%s): %v", dir, err)
	}
	srv, err := New(c, dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer srv.Close()
	get(t, srv.Addr(), c.Requests[0].PathQuery)
	if got := srv.Misses(); len(got) != 0 {
		t.Fatalf("Misses = %v, want none", got)
	}
	unused := srv.Unused()
	if len(unused) != 1 || unused[0].PathQuery != c.Requests[1].PathQuery {
		t.Errorf("Unused = %v, want exactly the second logged request %q", unused, c.Requests[1].PathQuery)
	}
	// An unconsumed Extra entry is the recall channel working as
	// designed, not an unused request (rung 3 spec, :265).
	for _, r := range unused {
		if r.Response == c.Extra[0].Response {
			t.Error("Unused reported an Extra entry; a strong ID match legally orphans search captures into Extra")
		}
	}
}
```

**Step 2: Run to verify it fails**

Run: `go test ./internal/harness/stub/ -run TestUnused -count=1 -v`
Expected: FAIL, "srv.Unused undefined".

**Step 3: Land `Unused()`**

`Server` keeps only `responses map[key][]byte` and `extra
map[key]extraEntry` today (`internal/harness/stub/stub.go:63-72`) and
never retains `c.Requests`, so `Unused()` needs a field as well as an
accessor (fold m13). `New` gains `requests map[key]cassette.Request`
built alongside `responses` from the same loop, and `Server` gains
`hit map[key]bool`, set in `ServeHTTP`'s request-map arm under the
existing mutex. `Unused()` returns the `requests` entries whose key is
unhit, as clones, in `c.Requests` order so the message is stable.

**Step 4: Write the failing policy-resolver tests**

```go
func TestPolicyForHonorsEveryDispositionRow(t *testing.T) {
	// Sixteen rows: the twelve rows of the rung 3 spec's disposition
	// table (:285-299), the adjudications-pending path (:304-314), the
	// half-on shape that proves the per-class Compat field, and the two
	// precedence shapes (:302), against hand-authored ledgers in
	// testdata/case/policy/. Eight of the sixteen have ZERO
	// committed corpus instances (ties excluded; oracle_bugs pending,
	// no-correct-answer, half-on; adjudications no-correct-answer,
	// pending; and both precedence shapes), which is exactly why the
	// fixtures are hand-authored rather than borrowed.
	//
	// EVERY field of every want literal is written out, including the
	// ones that happen to equal the zero value. The comparison is
	// struct equality, so an omitted field is a silent assertion that
	// it is false, and revision 1's table asserted GradeAlbumDistance
	// false on all fourteen rows, which would have shipped a resolver
	// that turned album-distance grading off corpus-wide against a
	// DistanceFloor of 294 (revision 2, fold M-a). Compat is
	// matcher.Compat rather than a single bool, because a switch is per
	// defect class and cannot be half on (spec :302, fold M-i);
	// fullCompat is Compat{true, true} and noCompat is Compat{}.
	tests := []struct {
		name string
		want casePolicy
	}{
		{"proximity clear", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"proximity flagged", casePolicy{DistanceCompare: compareBitIdentical, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"proximity skipped", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: false, GradeAlbumDistance: false, GradeVerdict: true, ExpectNoCandidate: true, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties unique", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties pinned", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties excluded", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: false, GradeAlbumDistance: false, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: true, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties skipped", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: false, GradeAlbumDistance: false, GradeVerdict: true, ExpectNoCandidate: true, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"oracle bug pending", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: fullCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"oracle bug pending on one class only", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: matcher.Compat{FtPattern: true}, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"oracle bug oracle-wrong", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromAdjudication}},
		{"oracle bug no-correct-answer", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: false, ExpectNoCandidate: false, Compat: fullCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"oracle bug oracle-correct", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"adjudication no-correct-answer", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: false, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"adjudication pending", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: false, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: true, VerdictFrom: verdictFromRecord}},
		{"case row oracle-wrong under a pending hypothesis", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: fullCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"probe row oracle-correct under an oracle-wrong hypothesis", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
	}
	if len(tests) != 16 {
		t.Fatalf("the table carries %d rows; it must carry one per spec disposition row (12), the adjudications-pending path, the per-class pending shape, and both precedence shapes", len(tests))
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each row's fixture directory is testdata/case/policy/<slug>
			// and carries all four ledgers plus the case id the row is
			// about, so a policy is resolved from exactly the inputs the
			// real sweep resolves it from.
			ties, prox, adj, bugs, caseID := loadPolicyFixture(t, tt.name)
			got := policyFor(caseID, ties, prox, adj, bugs)
			if got != tt.want {
				t.Errorf("policyFor(%s) = %+v, want %+v", caseID, got, tt.want)
			}
		})
	}
}
```

**Step 5: Write the gate-resident sweep entry point**

```go
// compatSummary runs the ModeCompat sweep once for the whole package.
// Three whole-corpus sweeps live here after Task 8 lands and two of
// them are the same run, which at the sweep's own 6-to-20s estimate is
// enough on its own to push go test ./... past the 45s trigger this
// task sets (fold m8).
var compatSummary = sync.OnceValues(func() (CaseGradeSummary, error) {
	root, err := repoRootPath()
	if err != nil {
		return CaseGradeSummary{}, err
	}
	return GradeCases(root, ModeCompat)
})

func TestWholeCorpusGradesCleanlyInCompatMode(t *testing.T) {
	root := repoRoot(t)
	summary, err := compatSummary()
	if err != nil {
		t.Fatalf("GradeCases: %v", err)
	}
	// The assignment floor is the only one that moves, and it moves for
	// exactly two owner-owned reasons: ruling F's escalation parks the
	// two pinned cases in TiesAwaitingRuling, and an owner ruling can
	// exclude a case. Subtracting both keeps the gate green on ruling
	// F's pre-authorized path and still fires on a shrink from any
	// other cause (revision 2, fold B4).
	assignmentFloor := AssignmentFloor - len(summary.TiesAwaitingRuling) - summary.TiesExcluded
	for _, f := range []struct {
		name  string
		got   int
		floor int
	}{
		{"cases", summary.Cases, CaseFloor},
		{"candidates", summary.Candidates, CandidateFloor},
		{"distances", summary.Distances, DistanceFloor},
		{"track vectors", summary.TrackVectors, TrackVectorFloor},
		{"assignments", summary.Assignments, assignmentFloor},
	} {
		if f.got < f.floor {
			t.Errorf("compared %d %s, floor is %d (assignments: %d candidate-bearing cases less %d awaiting ruling F and %d excluded): a comparison went missing",
				f.got, f.name, f.floor, AssignmentFloor, len(summary.TiesAwaitingRuling), summary.TiesExcluded)
		}
	}
	if summary.TiesExcluded > TiesExcludedCeiling {
		t.Errorf("ties.json carries %d excluded rows, ceiling is %d; excluding a case stops grading it and disposition is owner-only (spec :292)", summary.TiesExcluded, TiesExcludedCeiling)
	}
	// The assignment population is 214 measurable rows less whatever
	// ruling F still owes. A non-empty TiesAwaitingRuling is legal only
	// while the escalation finding exists.
	if len(summary.TiesAwaitingRuling) > 0 {
		if _, err := os.Stat(filepath.Join(root, "docs", "superpowers", "research", "2026-09-05-rung-3d-tie-mirroring-finding.md")); err != nil {
			t.Errorf("%d tie rows await a ruling (%v) but the ruling F finding is missing: %v", len(summary.TiesAwaitingRuling), summary.TiesAwaitingRuling, err)
		}
		t.Logf("awaiting Geoff's ruling F disposition, graded as pending rather than passing: %v", summary.TiesAwaitingRuling)
	}
	for _, nm := range summary.NearMisses {
		t.Logf("near miss: %s %s want %.17g got %.17g", nm.CaseID, nm.Field, nm.Want, nm.Got)
	}
}

func TestNoCommittedCaseHasTwoExtraItemsSharingASortKey(t *testing.T) {
	// beets' extra_items pre-sort order is a Python set order over
	// identity-hashed objects, outside PYTHONHASHSEED's reach, so every
	// committed record is reproducible only while (disc, track, title)
	// is total over that case's extras. It is today, on all 215. The day
	// lane E1 lands a case where it is not, beets' own record stops
	// being reproducible run to run, and this is where that is learned
	// rather than at a flaky gate.
	//
	// extra_tracks needs no such guard: album_info renumbers every track
	// 1..n across the whole release
	// (beetsplug/musicbrainz.py:645-646), so t.index is unique and
	// (index, title) can never tie.
	root := repoRoot(t)
	checked := 0
	for _, c := range loadEveryCase(t, root) {
		if c.Expected == nil || len(c.Expected.ExtraItems) < 2 {
			continue
		}
		checked++
		seen := make(map[[3]string]bool, len(c.Expected.ExtraItems))
		for _, idx := range c.Expected.ExtraItems {
			it := matcher.ProjectItemFields(c.Input.Items[idx])
			k := [3]string{strconv.Itoa(it.Disc), strconv.Itoa(it.Track), it.Title}
			if seen[k] {
				t.Errorf("%s: two extra items share the sort key %v; beets' own record is no longer reproducible run to run", c.ID, k)
			}
			seen[k] = true
		}
	}
	if checked == 0 {
		t.Error("no case carried two or more extra items, so this guard asserted nothing; three cases do today")
	}
}
```

**Step 6: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/harness/conform internal/harness/stub && git commit`

**Wall-time expectation, stated so the report can be graded against it.**
3c's request tier runs 3.34s over 215 cases at `-count=1` with one stub
per case, decoding 303 lookup bodies into the full `musicbrainz.Release`.
Tier 1 adds a `TranslateAlbum` per candidate (294 across the corpus), an
assignment solve per candidate, and a full `DistanceForAlbum` per
candidate on top of the same stub lifecycle. **Expect 6 to 20 seconds at
`-count=1`**, against `go test ./...` at 15.081s and `bash
scripts/check.sh` at 1m25.9s with gitleaks dominant (3c's close
figures). If the measured figure exceeds roughly 45 seconds for the
single-mode sweep, the task reports it as a finding and the conductor
decides whether the whole-corpus run becomes its own leg of
`scripts/check.sh` at full scope, which the spec already names as the
escape (`:267`) and which is explicitly **not** an optional target.

**Acceptance criteria:**

- Every one of `cassette.Expected`'s eleven fields appears in the
  graded-field table above with either a grading rule and a population
  or a recorded reason for not being graded.
- Every one of the sixteen policy rows has a switch arm and a fixture:
  the spec's twelve disposition rows, the adjudications-pending path,
  the half-on Compat shape, and both precedence shapes. **Every field of
  every `want` literal is written out**, including the ones equal to
  the zero value, because the comparison is struct equality (fold M-a).
  The eight rows with zero committed instances have hand-authored
  fixtures in `conform/testdata/case/policy/`.
- `casePolicy.Compat` is `matcher.Compat` and the half-on fixture proves
  a single bool would have been wrong (fold M-i).
- `Misses()` and `Unused()` are both empty on all 215 cases **in
  `ModeCompat`**, which is the only mode the rule reaches (fold B2).
- The five floors and the excluded ceiling all hold, with the assignment
  floor subtracting `TiesAwaitingRuling` and `TiesExcluded` so ruling
  F's escalation path stays green (fold B4).
- `gradeTunables` overlays `Quiet: true` with the shim citation,
  `mergeProfile`'s `estate` and `locale` arms carry the three `import.*`
  values those files set with a test per value (fold M-b), and
  `defaults_test.go` compares the echo against `gradeTunables`.
- `Sorted()[0]` is the candidate whose assignment and extras are graded,
  matching beets grading its own best candidate.
- The `extra_items` totality guard runs over every committed case,
  asserts it checked at least one, and passes today (three cases carry
  two or more extra items).
- Tier 0, the translation tier and the request tier stay green; `bash
  scripts/check.sh` green.

**Named mutations this task reports, each with its measured count:**

1. Grade `Candidate.Penalties` against `NonzeroKeys()` rather than the
   added-key view → fails on the **272** committed candidates carrying
   at least one key at 0.0.
2. Feed `DistanceForAlbum` its pairs in item order → fails on the **23**
   cases with a non-identity assignment, on both the `Tracks` vector and
   the album distance.
3. Drop the `Unused()` half of the request-set rule and truncate the
   flow (return after the ID lookup) → the sweep passes the miss rule and
   fails only the verdict comparison, which is the silent shape
   `Unused()` exists to convert into a loud one; with `Unused()` in place
   it fails on the **48** search-fallback cases.
7. Set `NoneRecAction` from beets' default rather than the profile in
   `mergeProfile` → nothing fails, because the `Quiet: true` overlay
   makes the field unreachable (contract V16). Reported as a
   deliberately unfired mutation, which is why `translate_test.go`
   asserts each merged value against the profile file directly rather
   than through a sweep (fold M-b).
4. Compare distances at 1e-12 on the `proximity` flagged row → the
   flagged case still passes, and the mutation is reported as **not
   caught by the corpus**, because `rec-strong-medium-boundary` is
   bit-identical today. The check that does catch it is the policy
   resolver's own fixture row, which is why row-level fixtures exist
   beside the corpus sweep.
5. Skip a case silently on a load error rather than joining its error →
   `summary.Cases` drops below `CaseFloor` and the floor fires.
6. Use `All()` rather than `Sorted()` when picking the graded candidate
   → fails on the **21** multi-candidate cases wherever insertion order
   and distance order differ.

**Deliverable count: 5** (`stub.Unused()` with its fixture; the
`policyFor` resolver covering all sixteen policy rows with every field
written out; `GradeCases`' tier-1 sweep with the graded-field
table's nine graded fields and two recorded non-grades; the
`extra_items` totality guard; the floors, the excluded ceiling,
`TiesAwaitingRuling` and the seven named mutations with their measured
counts, plus the measured wall time at `-count=1`, plus `mergeProfile`'s
three-key completion and its per-value test). **Five is one past
this plan's own roughly-four bar, and it is named rather than absorbed:
`stub.Unused()` and the totality guard are each a single accessor or a
single loop, and both are inseparable from the sweep that is their only
consumer.** Gate green; commit. **Report actual token spend, the sweep's wall time, and
the per-field compared counts.**

**After this task lands and before Task 8 dispatches, the conductor
dispatches one fresh-context `claude-opus-5` review of this task's diff
alone** (Resolution R11), reading `GradeCases` against
`cassette.Expected`'s eleven fields using the table above, confirming
every policy row has a checker and a fixture, and re-running the seven
mutations. Its verdict routes like any `diff-reviewer` verdict: accept,
one re-dispatch on fix, and a second fix is the conductor's decision.

---

## Task 8: Tier 2, the two runs, and the pending-divergence path

Model: `sonnet`. Depends on Task 7 and its fresh-context review.

**Files:**
- Modify: `internal/harness/conform/case.go` (verdict grading, the two
  modes, the compat cross-check)
- Modify: `internal/harness/conform/case_test.go`
- Create: `internal/harness/conform/testdata/case/verdict/` (the
  hand-authored ledger fixtures for the four zero-instance verdict rows)

**Interfaces:**

- Consumes: everything Task 7 produced, plus
  `analysis.AdjudicationLedger.BySubjectCase` (Task 6),
  `analysis.DefectFree`, `analysis.StringHeuristicHypotheses`,
  `matcher.DefaultCompat`, `matcher.Compat`.
- Produces:

```go
// internal/harness/conform/case.go

// GradeMode's doc comment, extended from Task 7's stub, now that both
// modes exist:
//
// Two runs rather than one, and the reason is arithmetic. Both
// string-heuristic hypotheses are ruled oracle-wrong, so ruling A
// defaults both compat switches OFF, and a default-mode run scores every
// case against a non-beets pattern table and diverges from every
// committed distance. One run cannot be both bit-identical against the
// committed records and correct against the adjudicated verdicts.
//
// ModeDefault takes the ledger's own dispositions. Each probe case of an
// oracle-wrong hypothesis with a machine-readable defect_free block is
// graded against that block; NO other case's verdict is graded, because
// no other case has a corrected expectation.
//
// ModeDefault does not drive tag.Album. It re-scores each case's own
// committed candidate set under the corrected table, and the reason is
// that a corrected verdict changes the FLOW: all ten defect_free probe
// cases log one lookup and zero searches and score strong under compat
// and medium under the corrected table, so re-running the flow would
// lose the strong-match early return (match.py:436), issue a search the
// case never logged, and fail the request-set rule before any
// defect_free figure was compared. The request-set rule is therefore
// ModeCompat's alone. Re-scoring reproduces the ten committed figures
// exactly, because the ladder's medium rung is decided by the minimum
// distance alone (match.py:288-290) and a candidate set a search would
// have widened cannot move a medium verdict off medium. Revision 2,
// fold B2; provisional, conductor ruling owed.

// DefectFreeFloor is the number of adjudication rows carrying a
// defect_free block today, which is ModeDefault's whole graded
// population.
const DefectFreeFloor = 10

// ProbeCaseFloor is the number of probe cases across the four
// oracle-wrong hypotheses.
const ProbeCaseFloor = 20

// ProbeCaseOracleCorrectRows is the number of those probe cases whose
// OWN adjudication row reads oracle-correct, so they carry no corrected
// verdict to grade and belong to neither the graded nor the ungraded
// population: real-nonlatin-alla-pugacheva-vol1 and
// real-nonlatin-x-japan-art-of-life, both probes of
// bug-nonlatin-alias-selection. Naming it keeps the accounting
// assertion readable and makes it obvious what goes stale the day Geoff
// rules another row; the assertion recomputes the figure from the
// ledgers and compares against this constant, so a movement fails
// loudly rather than being absorbed (fold m17).
const ProbeCaseOracleCorrectRows = 2

// CheckDefaultCompatMatchesTheLedger proves matcher.DefaultCompat's two
// literals still agree with corpus/analysis/oracle_bugs.json's own
// dispositions. The engine never reads corpus/, so this is the only
// place the two can be compared, and it is a HARD failure rather than a
// pending row: a disposition is an owner decision, and an agent
// silently absorbing one is exactly what the sentinel rules exist to
// prevent.
func CheckDefaultCompatMatchesTheLedger(root string) error
```

**Behavior contract:**

- **`ModeCompat` grades `Recommendation` and `Action`** for every case
  whose policy says `GradeVerdict`, against the record's own values, with
  the `Quiet: true` overlay in force (contract V17). Population 215 less
  the zero-instance exceptions.
- **`ModeDefault` installs `matcher.DefaultCompat()`** (today both
  switches off, so `SDPatternsGuardedNoBracketDiscount`) and grades **only
  the ten probe cases carrying a `defect_free` block**, against
  `defect_free.distance` at 1e-12 relative, `defect_free.recommendation`
  and `defect_free.action`. It grades no other case's verdict.
- **`ModeDefault` re-scores, it does not re-run** (revision 2, fold B2;
  **provisional, conductor ruling owed**, with the alternatives at R2).
  Per case: start the case's stub, replay only the release lookups the
  record's own candidates name, `TranslateAlbum` each, `AssignItems`,
  `DistanceForAlbum`, `Candidates.Sorted`, `Recommend`, `ActionFor`. No
  search is issued and **the request-set rule does not apply**, because a
  corrected verdict legally changes the flow and the committed request
  log is `ModeCompat`'s expectation. `ModeCompat` installs
  `matcher.FullCompat()` for the same symmetry (fold m6).
- **Every other case's `ModeDefault` verdict is compared against its own
  `ModeCompat` verdict**, and a difference appends the case id to
  `VerdictMoves`. This is the measurement Resolution R9 makes a named
  deliverable: the upper bounds are 133 cases carrying a bracket or
  parenthesis in a compared string and 44 carrying a featuring trigger,
  and the real count is unknown until this runs, because `StringDist`
  only shifts when the deletion lowers the distance. It measures the
  fix's effect on the candidates each case actually fetched and never a
  full counterfactual import, since a case that lost its early return
  would have searched and no such search is issued (fold B2); Task 9
  records that limit beside the list. `VerdictMoves`
  **fails nothing**; Task 9 writes the list to
  `docs/research/2026-09-05-rung-3d-default-mode-verdict-moves.md`
  and the conductor batches it into the adjudication stream.
- **The eight ungraded probe cases.** The six `adv-assign-*` rows and the
  two alias rows whose own disposition is `oracle-wrong` have no
  expressible rung-3 corrected expectation (Resolution R1). In
  `ModeCompat` they grade normally against their committed records; in
  `ModeDefault` their verdicts are not graded and each id is appended to
  `Ungraded` with its reason. The sweep asserts that
  `len(Ungraded) + DefectFreeFloor` accounts for every probe case of an
  `oracle-wrong` hypothesis whose own row is also `oracle-wrong`, so a
  case cannot fall out of both populations unnoticed.
- **The pending path** (Resolution R15). A case whose
  `adjudications.json` row reads `pending-adjudication` has its verdict
  ungraded, its id appended to `Pending`, and the gate stays green. A
  verdict disagreement on a case with **no** row **fails**, and the
  failure message carries the case id, the field, both values, and the
  ready-to-commit JSON row a human pastes into `adjudications.json`,
  carrying `subject.case_id`, `disposition: pending-adjudication`,
  `adjudicator: agent`, and empty `suspected`, `release_truth`, `reason`
  and `date` fields marked `TODO` so `CheckAdjudications`' own five
  non-empty rules force a human to fill them. **The runner never writes
  into `corpus/`.**
- **`CheckDefaultCompatMatchesTheLedger`** loads `oracle_bugs.json` and
  asserts `DefaultCompat().FtPattern == (bug-ft-pattern's disposition is
  not oracle-wrong)` and the same for `bug-remix-parenthetical`. It runs
  as its own gate-resident test.

**Step 1: Write the failing verdict tests**

```go
func TestWholeCorpusGradesVerdictsInCompatMode(t *testing.T) {
	// compatSummary is a package-level sync.OnceValues over
	// GradeCases(repoRoot, ModeCompat), so Task 7's sweep test and this
	// one share ONE run rather than paying for the corpus twice in the
	// same package (fold m8). Task 7's test switches to it in the same
	// change. ModeDefault stays its own run and is cheap, because it
	// re-scores committed candidates and issues no requests (fold B2).
	summary, err := compatSummary()
	if err != nil {
		t.Fatalf("GradeCases: %v", err)
	}
	if summary.Verdicts < VerdictFloor-len(summary.Pending)-len(summary.Ungraded) {
		t.Errorf("graded %d verdicts, want at least %d (floor %d less %d pending and %d ungraded)",
			summary.Verdicts, VerdictFloor-len(summary.Pending)-len(summary.Ungraded), VerdictFloor, len(summary.Pending), len(summary.Ungraded))
	}
	for _, id := range summary.Pending {
		t.Logf("pending divergence, verdict ungraded, launch gate refuses: %s", id)
	}
}

func TestWholeCorpusGradesTheCorrectedVerdictsInDefaultMode(t *testing.T) {
	summary, err := GradeCases(repoRoot(t), ModeDefault)
	if err != nil {
		t.Fatalf("GradeCases: %v", err)
	}
	if summary.Verdicts != DefectFreeFloor {
		t.Errorf("default mode graded %d verdicts, want exactly the %d defect_free rows: no other case has a corrected expectation", summary.Verdicts, DefectFreeFloor)
	}
	want := ProbeCaseFloor - ProbeCaseOracleCorrectRows
	if got := len(summary.Ungraded) + summary.Verdicts; got < want {
		t.Errorf("graded plus ungraded probe cases = %d, want at least %d (%d probe cases less the %d whose own row reads oracle-correct); a probe case fell out of both populations",
			got, want, ProbeCaseFloor, ProbeCaseOracleCorrectRows)
	}
	t.Logf("default mode moved %d non-probe verdicts: %v", len(summary.VerdictMoves), summary.VerdictMoves)
}

func TestCheckDefaultCompatMatchesTheLedger(t *testing.T) {
	if err := CheckDefaultCompatMatchesTheLedger(repoRoot(t)); err != nil {
		t.Errorf("matcher.DefaultCompat no longer agrees with corpus/analysis/oracle_bugs.json: %v", err)
	}
}

func TestPendingDivergenceLeavesTheGateGreenAndTheVerdictUngraded(t *testing.T) {
	// Zero committed rows read pending-adjudication, so this is a
	// hand-authored four-ledger fixture in testdata/case/verdict/pending.
	// The case's Go verdict deliberately differs from its record; the
	// pending row makes that legal at this gate and refused at the
	// launch gate (rung 3 spec, :310).
	summary, err := gradeFixtureCase(t, "pending")
	if err != nil {
		t.Fatalf("grade: %v", err)
	}
	if !slices.Contains(summary.Pending, "fixture-pending") {
		t.Errorf("Pending = %v, want the case id", summary.Pending)
	}
	if summary.Verdicts != 0 {
		t.Errorf("graded %d verdicts, want 0: a pending row leaves the verdict ungraded", summary.Verdicts)
	}
}

func TestAnUnadjudicatedDivergenceFailsAndPrintsTheRowAHumanMustWrite(t *testing.T) {
	_, err := gradeFixtureCase(t, "unadjudicated")
	if err == nil {
		t.Fatal("an unadjudicated verdict divergence passed; a mechanical disagreement is a Go defect until an owner says otherwise (spec :311)")
	}
	for _, want := range []string{"fixture-unadjudicated", "pending-adjudication", `"adjudicator": "agent"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %v does not carry %q; the message is what a human pastes into adjudications.json", err, want)
		}
	}
}
```

**Step 2: Run to verify they fail**

Run: `go test ./internal/harness/conform/ -run 'Verdict|DefaultCompat|Pending|Unadjudicated' -count=1 -v`
Expected: FAIL, `ModeDefault` and `CheckDefaultCompatMatchesTheLedger`
undefined.

**Step 3: Land tier 2 and both modes**

**Step 4: Run both sweeps, measure, and report**

Run: `go test ./internal/harness/conform/ -run 'WholeCorpus' -count=1 -v`
Record: each mode's wall time, `Verdicts`, `Pending`, `Ungraded`,
`VerdictMoves` and `NearMisses` counts, and the combined figure against
`go test ./...`.

**Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/harness/conform && git commit`

**Acceptance criteria:**

- `ModeCompat` grades 215 verdicts less `Pending` and less `Ungraded`,
  and all 167 `apply` and 48 `skip` actions agree.
- `ModeDefault` grades exactly the ten `defect_free` rows and no others,
  and all ten agree on distance, recommendation and action, which is the
  proof that `SDPatternsGuardedNoBracketDiscount` is right (Resolution
  R2).
- `VerdictMoves` is measured and reported as a count and a case list, and
  fails nothing.
- `Ungraded` names the eight probe cases with no expressible rung-3
  expectation, each with its reason.
- All four zero-instance verdict dispositions have hand-authored
  fixtures that fire.
- An unadjudicated divergence fails and prints the row.
- `CheckDefaultCompatMatchesTheLedger` passes and its mutation fires.
- **The combined two-mode wall time is measured and reported.** If it
  pushes `go test ./...` past roughly 45 seconds, the task reports it as
  a finding and the conductor rules on the `scripts/check.sh` leg (spec
  `:267`).
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Grade every case's verdict in `ModeDefault` →
   `TestWholeCorpusGradesTheCorrectedVerdictsInDefaultMode` fails, and
   the run reds on however many cases `VerdictMoves` names.
2. Grade the ten probe cases against their committed records rather than
   their `defect_free` blocks in `ModeDefault` → all ten fail, reporting
   each case's committed distance against its `defect_free.distance`.
3. Treat an unadjudicated divergence as pending →
   `TestAnUnadjudicatedDivergenceFailsAndPrintsTheRowAHumanMustWrite`
   fails, which is the exact silent-absorption the spec's `:311` forbids.
4. Flip `DefaultCompat().FtPattern` to `true` →
   `TestCheckDefaultCompatMatchesTheLedger` fails, and `ModeDefault`
   fails on the six `adv-ft-*` cases.
5. Drop the `Ungraded` accounting assertion and silently skip the eight
   probe cases → the graded-plus-ungraded assertion fires, reporting a
   count below `ProbeCaseFloor - ProbeCaseOracleCorrectRows` and naming both (fold m17).

**Deliverable count: 4** (tier 2's verdict grading in `ModeCompat`;
`ModeDefault` with the ten `defect_free` expectations and the eight
recorded `Ungraded` cases; the pending path with its green-gate arm and
its failing arm printing the row; `CheckDefaultCompatMatchesTheLedger`
plus the measured `VerdictMoves` list and both modes' wall times). Gate
green; commit. **Report actual token spend, both modes' wall times, and
every summary count.**

---

## Task 9: The pass close, and the rung 3 to rung 4 handoff

Model: `sonnet`. Depends on Tasks 0 through 8.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit)
- Create: `docs/research/2026-09-05-rung-3d-default-mode-verdict-moves.md`
- Create, only if any exists:
  `docs/research/2026-09-05-rung-3d-pending-divergences.md`
- Draft, handed to the conductor in the task return: one rung 3 spec
  tree-column correction; three contested-surface rows; two
  `corpus/README.md` corrections; two filed notes; one owner ask

**Behavior contract:**

- **The HISTORY entry is drafted in the house shape**: what landed, what
  the gate caught, and what a later pass would be wrong to rediscover.
- **The verdict coverage table, as a table a reviewer can check.** Per
  contract V1 through V20: its Go site, what grades it, and, for a rule
  no corpus case can reach, which hand-authored fixture grades it and
  what mutation proved that fixture fires.
- **The graded-field table from Task 7 travels into HISTORY**, all
  eleven `cassette.Expected` fields with their grading rule and
  population, including the two recorded non-grades (`ConfigHash` and
  `AppliedFields`). Rung 15's re-run reads it rather than deriving it.
- **The declared corpus-coverage gaps are recorded as gaps, not as
  passing checks** (spec, M8), and **the count is reconciled once
  rather than averaged.** The inventory's three sections disagreed (its
  (a) §13 lists eleven, (b) §5 four, (c) §12 seven); this plan settles
  the canonical list at **ten new**, on one criterion: a surface **3d
  itself builds** that no committed corpus case can reach, graded by a
  hand-authored fixture instead.

  1. `import.search_ids` and the explicit-ID branch (0 of 215).
  2. `match_by_id`'s no-id branch (0 of 215).
  3. `match_by_id`'s no-consensus branch (0 of 215).
  4. `import.timid`'s gate on the ID-match early return (0 of 215).
  5. The `max_rec` downgrade read against a key added at 0.0: **0 of 294
     committed candidates carry `missing_tracks` or `unmatched_tracks`
     at exactly 0.0**, though 272 carry some key at 0.0.
  6. `_recommendation`'s per-track key union in the downgrade set: the
     loop runs and never changes an outcome under the shipped `max_rec`.
  7. `import.quiet = false`, and with it the whole `timid` /
     `none_rec_action` / "ask" half of `_summary_judgment` (the shim
     forces quiet, `derive_oracle.py:204`).
  8. `quiet_fallback: asis`, and the `ASIS` action entirely.
  9. Every non-default value of the four recommendation knobs: all 215
     cases echo `[0.04, 0.25, 0.25]` and the default `max_rec`, so tier
     1's M1 cross-check proves the transcription and never the behavior.
  10. The assignment objective's `inclArtist=false` on a VA release: no
      tied or assignment-sensitive VA case exists, so passing `info.VA`
      into the cost matrix would pass every committed case.
  11. The recommendation ladder's strict-versus-inclusive asymmetry: **no
      committed candidate sits exactly on a threshold.** A sweep over all
      215 cases finds zero candidates at 0.04 or 0.25, and the single
      `proximity`-flagged row, `rec-strong-medium-boundary`, records
      `0.040000199999999986`, 2e-7 above the threshold, so it grades
      bit-identity rather than the comparison direction. The
      hand-authored ladder rows are the sole detector (fold M-f).

  **Together with 3a's three, 3b's six and 3c's one, rung 3 now carries
  twenty-one declared gaps.** The entry states the running total so rung
  15 does not re-count.

  **Three items are explicitly NOT counted, each with its reason.**
  `match.required` and `match.ignored` are 3a's gaps 1 and 2 already;
  3d builds their consumers and gives each a fixture, which moves them
  from deferred to closed-by-unit-test rather than making them new.
  "`strong` reached after a text search" is a corpus fact (no-search if
  and only if strong, exactly) rather than a code path 3d builds.
  `profile_hashes` covering one profile of three is a corpus-lane item
  open since 2b, restated below as carried.

- **Two surfaces are recorded as non-ports rather than gaps**, each with
  its reason: `bug-assignment-rejection-gap`'s fix, which needs a
  per-pair rejection mechanism beets does not have; and
  `bug-nonlatin-alias-selection`'s fix, whose corrected verdict is an
  applied artist string rung 3 computes none of. Both join the
  improvements register.
- **The evidence, not a bare "green".** Tier 1: cases, candidates,
  distances, track vectors, assignments and extras compared, the
  zero-miss and zero-unused results, the near-miss list, and the measured
  wall time at `-count=1`. Tier 2: both modes' verdict counts, the ten
  `defect_free` agreements, `VerdictMoves`, `Pending`, `Ungraded`, and
  `TiesExcluded` against its ceiling of 0. Tier 0 still 72/72
  bit-identical, the translation tier still 50/50, the request tier still
  215/303/48 at zero misses. `go test ./...` and `bash scripts/check.sh`
  with broken-down wall time. Ruling F's outcome: mirrored, or escalated
  with the finding's path and the two case ids awaiting a ruling.
- **The rung 3 to rung 4 handoff, stated as what rung 4's spec needs from
  3d** (ruling B option (b): rung 4 deploys at functional completeness
  with every divergence stamped with the matcher commit):
  1. **The entry point.**
     `tag.Album(ctx, src *musicbrainz.Client, items []matcher.ItemFields,
     tun matcher.Tunables, compat matcher.Compat, evalDate time.Time,
     opts tag.Options) (*tag.Proposal, error)`, in
     `internal/engine/tag`. `ctx` first, `evalDate` a parameter and never
     a clock read, `opts.SearchIDs` the only optional input.
     `Proposal` carries `CurArtist`, `CurAlbum`, `Candidates` (ascending
     distance, stable), `Recommendation` and `Action`.
  2. **The divergence stamp.** Every shadow divergence row rung 4 writes
     carries the matcher commit, so a divergence from a superseded
     matcher revision is re-evaluated mechanically rather than consuming
     owner adjudication. It joins `analysis.AdjudicationEntry` beside the
     existing fields, and the shadow's subject shape
     (`{album, release_mbid}`, `adjudication.go:35`) is already modeled
     and carries **zero committed rows**, so rung 4 is its first
     producer.
  3. **The shadow's budget class.** `musicbrainz.NewClient` requires a
     `Budget` (`client.go:128-130`) and every request `tag.Album` issues
     spends one token, so rung 4 caps its own MusicBrainz spend by the
     budget it builds the client with. The shadow's class is
     `ClassBackground` (`limiter.go:22-26`), deferred behind rung 7's
     user-present probe, and the wire pace is `DefaultInterval =
     time.Second` (3c contract C20). Rung 4 branches on
     `musicbrainz.ErrUnavailable` to defer rather than record a
     divergence, and on `tag.ErrNoCandidates` to record "no match".
  4. **The pending-divergence file format.** A row a human commits into
     `corpus/analysis/adjudications.json` carries `id`,
     `subject.case_id` (or, for the shadow, `subject.album` plus
     `subject.release_mbid`), `suspected`, `citation`, `release_truth`,
     `disposition: pending-adjudication`, `reason`, `adjudicator: agent`
     and `date`, all five prose fields non-empty
     (`adjudication.go:159-173`). The runner prints exactly that row and
     writes nothing; rung 4's shadow inherits the same rule, and **the
     launch gate refuses any pending row** (spec `:310`).
- **Forward carries, each stated so rung 4 and rung 15 read them rather
  than derive them:**
  1. **The rung 15 re-run criterion**, quoted verbatim from the Spec
     section above. Every rung 3 pass close carries it.
  2. **The four-knob wall registration** stays the named follow-on task
     for the first pass that runs after rung 1b closes:
     `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`,
     `SearchLimit`, `internal/config/keys.go` the file no lane edits
     until then, and `Threshold.Default` a string with `ParseInt` the
     only parser, so the three float thresholds need a float variant the
     registry does not yet have. **`AdminEmail`'s interim User-Agent
     contact sits beside it** (3c guard 7c): a dedicated contact row is
     the post-1b follow-on, because `Config.AdminEmail`'s own doc comment
     says adopters compare it and never display it.
  3. **`ties.json` measures `Candidates[0]` only**
     (`internal/harness/oracle/ties.go:54`), so 80 of the 294 committed
     candidate assignments were never asked the tie question, and **17
     of them are tied with no disposition** (a figure carried from the
     planning inventory and **re-measured by Task 7's own sweep**, which
     is the first run able to solve every candidate; the entry records
     the measured value, fold m2), across
     `real-noise-dylan-bootleg`, `real-noise-pink-floyd-de1994`,
     `real-nonlatin-girls-generation`, `real-nonlatin-utada-first-love`,
     `real-nonlatin-utada-first-love-locale`, and the non-best
     candidates of `should-ask` and `should-reject`. Five change tier
     1's graded `Tracks` vector on an alternative optimum and two also
     change the `tracks` penalty sum. **Filed for lane E1** as a
     derivation ask; tier 1 grades all 294 vectors meanwhile, and a
     divergence on one of the 17 reds with a message naming the
     candidate index and the `Candidates[0]` limitation.
  4. **The `lapjv` extension model is behavioral, not source.**
     `n = max(rows, cols)` with zero fill, established against
     `lap` 0.5.13 on 2634 random rectangular matrices and re-confirmed
     on 262 at this plan's drafting; the classic `(rows+cols, max+1)`
     recipe `derive_ties.py:58-66`'s docstring assumes disagrees on 30%
     of rectangular ties. Frozen as
     `internal/engine/matcher/testdata/lapjv-optima.json` with its
     provenance header. **The day `corpus/manifest.json`'s `lap` pin
     moves, that table must be regenerated**, and the oracle-tagged
     differential test this pass declined becomes worth building.
  5. **`MergePseudoRelease`'s aliasing hazard is now rung 9's alone.**
     `merged.Tracks` aliases the pseudo's backing array
     (`internal/engine/musicbrainz/pseudo.go`), `AlbumForID` closes it by
     construction because only the merged value escapes, and `tag.Album`
     inherits that same call site unchanged. The `slices.Clone` is owed
     the day a caller retains both values, which is rung 9's pipeline.
  6. **`vaLikely` stays unexported, permanently.**
     `match_by_id`'s no-consensus branch is corpus-unreachable and the
     search path computes `vaLikely` inside `AlbumSearchFor`
     (`query.go:108`), so no caller outside `musicbrainz` ever needs it.
     3c's carry 13c is closed as "stays unexported", with the reason
     recorded so a later pass does not export it for symmetry.
  7. **The retry path is ungraded by the corpus and always will be**:
     the stub answers 200 or 404 only, `cassette.Request` carries no
     status field, and the response-status extension stays declined, now
     for the fourth pass running.
  8. **Three regime divergences travel together**, unchanged from 3c
     (guard 9): 3b's `\b` in `SDPatternsWordBoundary` and now in
     `SDPatternsGuardedNoBracketDiscount`, the query path's
     full-versus-simple case lowering, and `_extract_id`'s `\w`. All
     three are Unicode-aware in Python and ASCII-only in Go RE2, and all
     three are benign on the committed corpus.
  9. **`Case.Extra` still has no producer** (3c guard 5), zero of 215
     cases carry an entry, and `Unused()` deliberately excludes `Extra`
     because a strong ID match legally orphans search captures into it.
  10. **`profile_hashes` covers `locale` alone** (open since 2b) and the
      response-status extension stays declined (open since 2c). Neither
      moved this pass.
  11. **The measured wall times**: tier 1, tier 2 in both modes, tier 0,
      the translation tier, the request tier, `go test ./...`, and
      `scripts/check.sh`, which the version-bump differential doc will
      want, the same courtesy 2b passed to 2c, 2c to 2d, 2d to 3a, 3a to
      3b, 3b to 3c, and 3c to 3d.
- **The drafted corrections, each one line for the conductor's batch:**
  - The rung 3 spec `:331`, the 3d tree column: this pass also writes
    `internal/engine/{matcher,tag}`, `internal/harness/{stub,corpus_test.go}`,
    `internal/harness/oracle/ties_derive_test.go` and
    `corpus/analysis/ties.json`, each justified in this plan's Global
    Constraints. An uncorrected spec is the next pass's stale input,
    which is why 3b and 3c each drafted their own.
  - The rung 3 spec `:163-185`, the tree listing: `internal/engine/tag`
    joins `matcher`, `musicbrainz` and `conform`, with Resolution R18's
    package-cycle reason.
  - **The build spec's own tree listing**,
    `specs/2026-09-02-greenfield-server-build.md:312`, which lists
    `internal/engine/{matcher, musicbrainz, tagio}`: `tag` joins it,
    with one clause distinguishing it from rung 8's `tagio` so the two
    are not read as one surface (revision 2, fold m16). This is the
    second stale document revision 1's correction list missed.
  - The rung 3 spec `:300`, B3's paragraph: the three probe cases it
    names as carrying no adjudication row are now two, both under
    `oracle-correct` hypotheses, so the rule fires on nothing today.
    The spec anticipated the movement, so this is an **optional
    clarification** rather than a correction.
  - The three contested-surface rows this plan's "Contested surfaces"
    section drafts: `internal/engine/tag`, and the
    `internal/harness/oracle` and `corpus/analysis/*.json` amendments.
  - `corpus/README.md:2091-2094`, which says "No case in the smoke
    corpus is flagged today, so the ledger is empty";
    `proximity.json` carries two rows, one flagged
    (`rec-strong-medium-boundary`) and one skipped
    (`admit-zero-track-dropped`).
  - `corpus/README.md`'s tie-ledger section: `ties.json` rows now carry
    an `adjudicator` field, required whenever a disposition is set,
    refused as `agent` on `excluded` and accepted on `pinned`, with the
    asymmetry's reason.
- **One owner ask drafted, batched rather than raised** (Resolution R1):
  re-dispositioning `bug-assignment-rejection-gap` from `oracle-wrong` to
  `no-correct-answer`, whose semantics the spec's own grading table
  already describes as exactly the behavior this pass implements by hand
  (`:296`). `disposition` is owner-only and this pass writes none.
- **Two notes filed, neither this pass's work:** the oracle-tagged
  `lapjv` differential test, for the lane that owns
  `internal/harness/oracle`, to be built the day the `lap` pin moves
  (carry 4); and extending `derive_ties.py` to ask the tie question of
  every candidate rather than `Candidates[0]`, for lane E1 (carry 3).
- **STATUS stays present tense and under 60 lines**; anything historical
  moves to HISTORY. The next action points at rung 4, and the lane
  registry entry is updated.
- **Both budgets are scored**: token spend against the re-derived
  ceiling, and attended interactions as two counts, planning misses and
  execution sittings, against a target of zero.

**Deliverable count: 4** (the HISTORY draft in the house shape with its
eleven forward carries; the verdict coverage table and the graded-field
table, with the reconciled ten new declared gaps, the two non-ports and
the running total of twenty; the evidence record with both tiers' counts,
both modes' wall times and ruling F's outcome; the STATUS rewrite plus
the rung 3 to rung 4 handoff section, the six drafted corrections and
rows, the two filed notes and the one owner ask). Gate green; commit.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain per task,
**strictly sequential**. Nothing here runs in parallel, and if a second
code slot opens at all it belongs to lane A or lane T in its own
worktree, never to a second executor in this tree. `docs/STATUS.md`'s
lane registry records lane-2 as BLOCKED until Phase 0b and 0c land, so
the expected state through this pass is one code lane. Every task runs
the repo's Sonnet implementer (`cairn-implementer`, which commits its own
task's specific files once `bash scripts/check.sh` is green) then
`diff-reviewer` then the gate, with one re-dispatch on `fix` and a second
`fix` as the conductor's decision. The conductor dispatches
`code-simplifier` over each accepted task's changed Go after review and
before the pass close, and owns every ledger edit, the lane registry, the
three contested-surface rows, the six drafted corrections and the batch
commit. **Runaway guard armed at launch**, per the standing requirement.
**Subagents never dispatch agents.**

**Models.** `sonnet` by default. **Three named upshifts to `opus`:**

- **Task 1**, the exact assignment solve. It is a from-scratch
  Jonker-Volgenant implementation whose every downstream consumer depends
  on it, its cost matrix has three independent ways to be plausibly wrong
  (raw sums instead of normalized cells, `inclArtist` true, pairs in item
  order), and one of those three is invisible to every committed corpus
  case. The zero-pad extension model is a measured behavioral claim
  rather than a documented one.
- **Task 2**, the tie mirroring. It is the pass's one genuinely unbounded
  task, it carries ruling F's escalation, and mirroring a specific C++
  build's pivot order is named in the spec as "the hardest single piece
  of the port" (`:411`).
- **Task 7**, the tier-1 runner. It is the instrument every later claim
  about the matcher rests on, its failure mode is a silent pass, and the
  same failure mode has already occurred twice in rung 3 (3a's
  `sorted-penalties-tied` fixture and 3b's struck Task 5 mutation). It
  also carries the eleven-field graded-field table the fresh-context
  review checks.

**Task 8 is the runner-up**: the conductor upshifts it if Task 7's
fresh-context review returns `fix`, because the two grading modes and the
pending path sit on the same instrument.

**Review.** Rung 3 is not one of the three full adversarial-cycle rungs
(5, 12 and 16), so review scales to the in-chain `diff-reviewer` per task
plus the pass-end gate, **with one exception this pass takes and 3c
declined**: a fresh-context `claude-opus-5` review of Task 7's diff
alone, dispatched after Task 7 lands and before Task 8, per Resolution
R11 and spec `:362`. It is scoped to the runner, not the engine. Its
verdict routes like any other: accept, one re-dispatch on `fix`, a second
`fix` is the conductor's.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- **Task 1 → Task 2.** Task 2 changes `solveLAP`'s internals and nothing
  else.
- **Task 1 → Task 3.** `Candidates.Add` calls `AssignItems` and
  `AlbumPairsFor`.
- **Task 3 → Task 4.** `Recommend` reads `Candidate`, and Task 4 extends
  the `Validate` Task 3 lands.
- **Tasks 3, 4 → Task 5.** `tag.Album` composes the accumulator, the
  recommendation and the action.
- **Task 6 depends on Task 0 alone** and is independent of Tasks 1 to 5.
- **Tasks 1, 3, 4, 5, 6 → Task 7.** The sweep needs the whole engine
  flow, the stub accessor, and the ledger surfaces.
- **Task 7 → the fresh-context review → Task 8.**
- **Tasks 0 to 8 → Task 9.**
- Task 6 is independent of Tasks 1 through 5 and Task 2 is independent of
  Tasks 3 through 6. **They are still run sequentially**, because one
  tree runs one executor and the module gate runs over `./...`
  regardless.

**Where this pass splits, if it splits.** Nine tasks plus a pre-task is
past the roughly-eight bar, so the split point is named rather than the
scope packed. **The split is at the Task 5 / Task 6 boundary**: Tasks 0
to 5 are the engine half (`internal/engine/{matcher,tag}`) and write no
`internal/harness` at all; Tasks 6 to 9 are the harness half plus the
gate.

**The recommendation is to keep one pass.** Three reasons, in order of
weight. **Rung 3's whole point is the conformance gate**, and an engine
half shipped alone would be exactly the "instrument nobody points at
anything" shape the spec's risk 2 names, with five new engine files and
no run grading any of them beyond their own unit tests. **The rung 3 to
rung 4 handoff is at 3d's close** (spec `:333`), and rung 4 needs the
entry point and the divergence vocabulary together, so a split delays
rung 4 by a full close and a full staleness pre-task for no throughput
gain while lane-2 stays BLOCKED and nothing can take the freed slot.
**And the two halves are cheap in opposite places**: the engine half
carries two of the three `opus` upshifts and all of ruling F's risk,
while the harness half carries the fresh-context review, so splitting
would put a second close and a second gate cycle between the two
expensive things rather than between a cheap thing and an expensive one.

**Take the split only if Task 2's actual spend re-derives the ceiling past
8M.** The conductor takes the call at the Task 2 checkpoint, on the
number rather than on a feeling. If ruling F escalates, that is a
*cheaper* Task 2 rather than a dearer one, and it argues against the
split.

**Budget:** 6M starting ceiling, re-derived after Task 2; phase 1's
combined ceiling 12M as a reservation; flag at 80% of the live ceiling
(4.8M against 6M), routed to the conductor under the 2026-09-05 grant
through 2026-09-07 09:00; STATUS after every task. Spend is scored at
close against the re-derived ceiling and against **zero attended
interactions**, which is this pass's target. Ruling F's escalation and a
pending divergence are both recorded rather than blocking, so neither
counts as an execution sitting.

---
