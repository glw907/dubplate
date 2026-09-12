# Draft STATUS text for the conductor's fold (rung 8c close)

Written by Task 4 (`docs/superpowers/plans/2026-09-08-rung-8c-absorb-repoint.md`).
The implementer never edits `docs/STATUS.md`; the conductor folds this
draft in, keeping the file present tense and under 60 lines, moving
whatever it displaces to `docs/HISTORY.md`.

## Proposed STATUS.md changes

- **Current state**: add that rung 8 is now closed in full (8a, 8b,
  8c), `internal/engine/shadow` holds no FLAC container knowledge, and
  the shadow's read path routes through `tagio.Parse` and the newly
  exported `tagio.Tags`. Lane 1 (main tree) is idle, 8c's plan
  discharged; lane 2 (`.claude/worktrees/lane-2-5b`) continues 5b-i.
- **Lane registry row for lane 1**: replace "RUNNING 8c segment 1A ...
  Task 3 in flight" with "IDLE. Rung 8c closed at `<Task-4-commit-sha>`
  (Tasks 1-4, plus one Task-2 fix round and one conductor-owed R6 fix).
  Next action: rung 5b-ii's plan, or a fresh lane-1 assignment, per the
  conductor's next dispatch."
- **Open items (carried)**: bump "Sixty-three declared coverage gaps"
  to the new running total (this pass's own count, stated in the
  HISTORY entry, added to sixty-three). State plainly whether 8c
  declared any new gap beyond a restatement of an already-counted 8a
  or 8b gap (the HISTORY entry's "declared coverage gaps" section has
  the arithmetic).
- **R4's denominator**: record as **UNREAD**. The conductor's box read
  failed at 03:14 AKDT (ssh timed out at the connection level, twice;
  no quiet-box check was possible), so no `ObservedApplied` /
  `ObservedSkipped` / `Unreconciled` sum exists yet. State zero known
  exclusions out of the (unread) denominator, name the exclusion set
  (Task 3's seven-row, now eight-row, enumeration) as the bound, and
  say the read is owed to the next attended session before rung 15
  reads it as final.
- **Rung 9 handoff line**: rung 8's absorption is complete; rung 9 is
  the first caller after the shadow to read a file's tags and its
  STREAMINFO off one `tagio.Parse`, and it inherits the `io.ReaderAt`
  gap (R5) as its own decision, not one rung 8 deferred silently.

## ROADMAP.md text (per the LEDGER OVERRIDE, folded here rather than
## edited directly)

ROADMAP.md carries no per-rung ledger of its own; the greenfield
section's only rung-specific sentence is the original, long-stale
"Rung 1 ... landed 2026-09-02. Next: rung 2's plan (corpus and
harness)" (line 33), never updated since rung 1 closed. Rung 8's
closure does not, by itself, justify rewriting that whole sentence's
history — `docs/STATUS.md` and `docs/HISTORY.md` are where per-rung
state actually lives — but the acceptance criterion asks for rung 8's
entry to read as closed and the next rung's entry to read as the
queue's next item, so the smallest correction that satisfies it:

> Rung 1 (foundation) through rung 8 (tag I/O: the writer core and
> reader, PICTURE and the conformance instrument, and the shadow's
> absorption and repoint) are landed, alongside rung 4 (the shadow),
> rung 1b, rung 3, and rung 5's pass 5a. Next: rung 5's pass 5b
> (identity, auth, config contract, session, roster, notify).

**Decision the plan did not cover**: ROADMAP.md's rung-tracking
granularity does not match a "one entry per rung" structure, so this
draft proposes replacing the single stale sentence rather than
inventing a new per-rung entry format. The conductor may prefer to
leave ROADMAP's coarse-grained sentence alone entirely and let
STATUS/HISTORY carry rung-level state, in which case this section is
skippable.

## Budget line for STATUS's spend-to-date sentence

8c's own spend, from this run's journal
(`wf_c808ec5e-1ba`): Task 1 accepted first pass; Task 2 accepted after
one comment-only fix round plus one conductor-owed R6 fix; Task 3
accepted after one comment-only enumeration-table fix. Roll 8c's total
into the "Spend to date" line once the conductor has the per-agent
token figures (this task did not have API-metered token counts on
hand; the journal's agent transcripts are the source if an exact
figure is wanted).
