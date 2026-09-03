# Pass 2b case tasks: staleness corrections (conductor record)

Product of the fresh-context staleness review of plan Tasks 4-14
against the machinery as shipped (commits cec3c12 through c1b8c05),
plus the conductor's three rulings, 2026-09-02. Every case-task
dispatch (Tasks 4-13) folds the corrections below; Task 14 folds its
own section. The plan file itself is unchanged; where it and this
document disagree on these specific points, this document wins.

## Conductor rulings

1. Family 14's zero-candidate case is unblocked by the machinery
   pre-task 2b-3b (tie-ledger skip row, DeriveTies skip return,
   claims `min_candidates: 0` mode), not deferred to 2c. Counts stay
   61 cases / 68 committed.
2. `ReleaseSpec` gains optional `original_date` (pre-task 2b-3b);
   Task 9 uses it rather than hand-editing the release-group date.
3. The response hand-edit escape hatch is blessed generally: a shape
   `ReleaseSpec` cannot express (absent recording artist credit,
   non-Official status, extra label-info, candidate-side absences) is
   reached by editing the generated `responses/001.json` BEFORE the
   first freeze, and the case's README inventory row names the edit.
   Tasks 6, 8, and 10 use this route where their sections need it.

## Mid-sweep conductor rulings (2026-09-02, overnight)

4. Family 2's "combined" shape (nonzero extra_items AND extra_tracks
   from one assign_items call) is unexpressible - the lapjv leftover
   set is one-sided, verified against pinned source by the reviewer.
   The two combined-with-scrambled-order cases stand; the exclusivity
   argument lives in the README. Task accepted.
5. Family 3's misflag case (nonva-va-artists-misflag): the pinned
   source proves match.py:453's va_likely has NO record-level
   observable under the empty-search convention; its witness is the
   frozen request log's arid: query, which make derive mechanically
   enforces (a wrong query would miss the stub and go red). Accepted
   with the request-log witness in place of a claims.json field; the
   family-10-only non-empty-search convention stands.

## Corrections folded into every case-task dispatch (4-13)

- **Promotion ordering** (plan's stated order is impossible as
  shipped): promote the frozen case into `corpus/cases/<id>/` FIRST,
  then `go test -tags oracle ./internal/harness/oracle -run Ledger
  -update-ledgers`, then write `claims.json` and any proximity row,
  then run both gates, then commit. The invariant is "never commit
  red", not "never move before the row exists".
- **Ledger membership**: `ties.json` gets a generated row per case;
  `proximity.json` is an EXCEPTION ledger - a row only for a case the
  flag returns `flagged` or `skipped` for, with `boundaries` (flagged
  only) and a written reason. Listing a clear case fails the gate.
- **Files list**: every case task's commit includes each new case's
  `claims.json` (the plan's per-task Files lists omit it).
- **Search-fallback entries** are hand-added to `case.json`'s
  `requests` array (copy `path_query` verbatim from `misses.json`);
  README section "Authoring a synthetic case" sanctions this.
- **`extra` overrides** must not name `mb_albumid` in `fields` (the
  fence check runs before the OpExtra branch that sets it).

## Task-specific notes

- Task 4 creates the README structural-inventory section (does not
  exist yet); Task 14 later rewrites it whole.
- Task 6's "VA with absent per-track credits both sides": empty the
  recording `artist-credit` in the generated response before first
  freeze (ruling 3).
- Task 8: derived items carry `disc` only on multi-medium specs and
  never `disctotal`/`media`; both baseline and mismatch need explicit
  overrides; `corpus/cases/multi-disc/claims.json` is the
  applied-fields claims template.
- Task 9 uses `ReleaseSpec.original_date` (pre-task) for both
  year-penalty branches.
- Task 10: family 10 goes through the spec's independent-items mode
  (overrides skipped entirely, no `allow_album_id_override` needed);
  candidate-absent country needs a response hand-edit; each search
  hit costs its own release lookup in the request log.
- Task 11: "candidate length absent" is `length: 0` in the spec
  track, with the item's own length restored by an override.
- Task 12: `TrackInfo.track_id` is the RECORDING id - a "matched"
  case copies `media[].tracks[].recording.id` from the generated
  document; an "absent" track index is `track: 0`.
- Task 13 runs after pre-task 2b-3b; the zero-candidate case uses
  the ties skip row and the `min_candidates: 0` claims mode.
- Task 14: README work is the inventory table alone; read the
  committed-case total off the tree; report the determinism run's
  wall time broken down (corpus re-derive vs ties vs component) as
  the 2c budget line.
- Cost model: `make derive` and `-update-ledgers` are O(all committed
  cases) per run, so later case tasks cost more wall time than early
  ones; budget accordingly, not per new case.
