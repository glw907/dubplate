# Rung 6c plan: prose-and-facts review

Graded commit `cebff05` (revision 2.1, plus its approval block). Folded into revision 2.2 of
`docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md`. Line citations below are
`cebff05`'s own; on the plan's current HEAD, lines 13-75 sit one line earlier and lines from 76 on sit
15 lines later.

## Findings

**F1 (blocker).** The plan's 5d premise is stale. The 5d plan's revision 2 (`9177429`) keeps
`internal/config` byte-identical across 5d-i and 5d-ii (its `:269-278`), runs in lane 1 after 5c-iv
closes, and edits `doctor.go`, `doctor_test.go`, `check.go`, `check_test.go`, and `main.go`, which
6c's Task 5 also edits. Conductor rule (decided): 5d-i's executor launches only after 6c merges. Fold
the 6c text at `cebff05:154-159` accordingly, align the header's conditions (`:20-21`) and Task 0
blocker 4 (`:754-756`) and the contested row and conductor action 7 (`:1325`) to that one rule, drop
the stale "drafted (`af7d941`, revision 1, pre-review)" line, and remove `:1402-1403` from "Not
verified."

**F2.** `cebff05:141-142`'s first lane-condition bullet is false: `git log -- internal/config` now
shows `75fde16` (the 5c-iii security fold, L2's site-name grammar) above `6e810c4`. Rewrite so it
agrees with bullet 3 and Ruling 1's reason.

**F3.** Header `:16-17` misreads the spec's band row: the spec's 7 for 6c counts build tasks beside
its own pre-task and close (spec `:1127`, `:1140`); say "seven (Task 0, five build tasks, and the
close)." In Ruling 1's "no" branch (`:89`), make clear the spec's own 4.5M is that row's flag, and the
plan's "no" figure is `(0.2 + 7 × 0.46 + 0.6) × 1.1 = 4.42M` rounded, a different number.

**F4.** `:86` "the flag at 4.32M, the projection itself": the projection is 4.33M, so an on-plan pass
crosses the flag at its last architecture read; say so.

**F5.** Task 1 `:819`: split the "second row" state into two rows: a second `settings_generation` row
with a key other than the pinned value (extended code 275) and one repeating the pinned value
(extended code 1555, `SQLITE_CONSTRAINT_PRIMARYKEY`, driver `lib/sqlite.go:3228`; confirmed in the
module cache).

**F6.** R-6c6 `:464-470`: state the settings `actor` column is TEXT with no foreign key, since
invariant 5 refuses a table pairing a foreign key with key and value columns.

**T1.** `:399` setup-colon payoff; rewrite "The compiler guarantees only that no `Parse*` function
(`internal/config/thresholds.go:84` onward), and no other function, accepts a `Setting`."

**T2.** `:46-54` one 100-word sentence carrying five deliverables; split into one sentence per
deliverable.

**S1.** The plan says `internal/backup` pointers are by symbol, yet R-6c13 and Task 4 cite line
numbers from `30a66ac`. Suffix each with "at `30a66ac`," and fix the code-tick text that is not a
verbatim quote.

**S2.** `:426` "fifteen wall thresholds": count `AddThreshold` registrations at master (the reviewer
counted twenty: auth 10, serve 3, shadow 2, web 4, store busy timeout 1) and state the true number.

**S3.** `:632` and D5 `:687` attribute "and 6c re-homes it after" to R-6bi8; it is from the 6b-i
plan's `:178-180`. Fix the citation.

**S4.** `:172-173` label the two subtracted figures as priced, not measured, and note the 1.95M
actual includes reads shared with the matcher pass.

**S5.** `:88-89` drop the translit-order hedge: STATUS has 6c right after 6b-ii-b merges, then
translit.

**S6.** Cut two decorative contrast frames (`:729` "not fixes"; `:349-350` "not a new column").

**S7.** Cut reflexive "exactly" at `:245`, `:399`, `:448`, `:559-560`, `:672`, `:798`, `:1271`; keep
exit gate 13's.

**Rulings as a standalone read.** Moot, ruled: the Rulings section is not re-opened by this fold, per
the conductor's instruction; the reviewer's notes on that section are not acted on.

## Dispositions

- F1: folded. Rewrote the "Lane condition, answered" section's 5d bullet, the header's condition text,
  Task 0 blocker 4, the contested-surfaces row for `doctor.go`/`check.go`, conductor action 7, and the
  header status line; removed the stale "drafted (`af7d941`, revision 1, pre-review)" text and the
  `:1402-1403` "Not verified" bullet. Verified against the tree: 5d plan is revision 2 at `9177429`,
  split into 5d-i/5d-ii, its own text rules 5d-i waits on 6c's merge (its `:10` and its
  contested-surfaces row `:1483`).
- F2: folded. `git log -- internal/config` confirms `75fde16` sits above `6e810c4`; rewrote the bullet
  to agree with bullet 3 and Ruling 1's reason.
- F3: folded. Header now reads "seven (Task 0, five build tasks, and the close)"; the Ruling 1 "no"
  branch now distinguishes the spec's own band-table flag (4.5M for its 5.6M ceiling, which includes
  planning) from the plan's own execution-only arithmetic (4.42M rounded to 4.5M).
- F4: folded. Noted the projection is 4.33M against a 4.32M flag, so an on-plan pass crosses the flag
  at its last architecture read.
- F5: folded. Split the `settings_generation` "second row" state into the two rows, with the
  `SQLITE_CONSTRAINT_PRIMARYKEY` = 1555 constant confirmed at `lib/sqlite.go:3228` in the module cache.
- F6: folded. Added the actor column's TEXT-with-no-foreign-key note and its invariant-5 reason.
- T1: folded, using the reviewer's exact rewrite.
- T2: folded. Split the "What this pass is" sentence into one sentence per deliverable (five deliverables plus the carried R13 risk).
- S1: folded. Verified `guard.go:86`, `guard.go:87`, `prune.go:123`, `prune.go:154`, and `prune.go:18-20`
  against `30a66ac` (confirmed by diffing the branch's blobs against master's line numbers, a 3-line
  offset on `prune.go`); appended "at `30a66ac`" to each, and rewrote the two paraphrased expressions
  (`0*float64(tracks)`, `own.Sub(stamp) <= 0`) to show the real source line with the zero value
  substituted in, rather than presenting the substitution as a verbatim quote.
- S2: folded. Counted `AddThreshold` call sites at master: auth.go 10, serve.go 3, shadow.go 2,
  web.go 4, store.go 1 = 20. Replaced "fifteen" with "twenty" and the breakdown.
- S3: folded. Confirmed the 6b-i plan's `:177-180` carries the "so 6d declares it and 6c re-homes it
  after" text under the R-6bi8 heading; cited the plan and line range at both sites instead of naming
  R-6bi8 alone.
- S4: folded. Confirmed against `docs/HISTORY.md`'s 6b-ii-a "Both budgets" entry (1.95M measured,
  including reads shared with the matcher pass) and the 6b-ii plan's priced Task 0 (0.11M) and priced
  close (0.99M) figures; relabeled accordingly.
- S5: folded, folded together with F3/F4's edit to the same paragraph. STATUS confirms 6c runs right
  after 6b-ii-b merges, then translit; dropped the "if the conductor keeps that order" hedge.
- S6: folded. Cut "not fixes" from Task 0's model line and "not a new column" from R-6c6's "Read in
  the tree" bullet.
- S7: folded. Cut "exactly" at the four sites that map to the plan's current lines 271, 475, 588, and
  830 (one of the seven cited sites, `:399`, was already rewritten whole by T1, and one,
  `:672`/`prune.go`'s "exactly what the registry exists to prevent," was cut as part of S1's rewrite of
  the same sentence). Left exit gate 13's own "exactly the correlated-rule entry point" untouched, and
  left the other unlisted "exactly" occurrences (`:422`, the Ruling 3 validator table, criterion 9,
  and "Dependencies, stated exactly") alone, since the finding did not name them.
- Rulings as a standalone read: recorded as moot, ruled. The Rulings section was not reopened; the
  approval block is unchanged.
