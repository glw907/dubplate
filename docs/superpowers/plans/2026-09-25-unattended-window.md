# Unattended window, 2026-09-25

A 4-hour-plus unattended run, prepped in the attended session of 2026-09-24/25 and executed by
a fresh session after Geoff clears. The run finishes the two open passes and prepares three
plans for Geoff's morning approval. No new pass executes, and nothing deploys.

## Grants (Geoff, 2026-09-25)

- **Scope:** close 5c-ii and 6b-i in full, then draft and fully review three plans: 5c-iii, 6b-ii,
  and the matcher word-boundary pass. No new pass executes. The plans end at "awaiting owner
  approval", with their rulings consolidated.
- **Spend cap:** 12M of subagent tokens for the window. At the 80% flag (9.6M), start no new
  unit of work: finish the one in flight, write STATUS, and stop. Do not ask; there is no one
  to ask.
- **Box contact:** read-only ssh is allowed if a plan's staleness read needs a box fact. No
  write, no deploy, no restart. Nothing deploys until pass 5d.
- **Standing:** the conductor rules any fold, escalation, or second `fix` by the plan and the
  standing rules, and records each ruling in STATUS as a reading Geoff can ratify or reverse.
  A genuine owner fork is never ruled: it goes into the plan's rulings section or STATUS's
  morning question.

## Starting state

Read `docs/STATUS.md` and the memory entry `session-2026-09-25-handoff` first.

- **5c-ii (lane 1, main tree, master):** Tasks 1-7, the simplifier (`2424b9a`), and the three
  pass-end reviews (`71de914`) are done. The review fold ran as workflow `wf_02fe1440-91c`
  (Fold A: the gate scripts; Fold B: security M1, the lows, the CSP pair, titles, the spec 1.1
  ink values). If STATUS does not record the fold's outcome, check `git log` on master for the
  Fold A and Fold B commits, and `git status` for warm edits, before anything else.
- **6b-i (lane 2, worktree `.claude/worktrees/lane-2-6b-i`, branch `lane-2-6b-i`):** every
  task, the simplifier (`25e5a3e`), and Task 8's drafts (`33ceff5`) are done. It is ready to
  merge.

## Sequence

Segments are ordered by the one contended resource: the main tree. Two executors never run
in one tree.

1. **Finish the 5c-ii fold** if STATUS shows it incomplete: relaunch only the unfinished fold
   task through `pass-execute` (a run cannot resume across sessions), with a note to review
   and finish any warm edits rather than start over.
2. **6b-i's merge ritual**, with the main tree quiet: in the worktree run
   `npm ci --prefix web` (master now carries 5c-ii's `web/`), then `git rebase master`
   (expect a textual conflict in `cmd/dubplate-server/main.go`, where both passes add a
   registration line: keep both in landing order), the full gate in the worktree,
   `git merge --ff-only lane-2-6b-i` on master, the full gate on master, and one
   `diff-reviewer` over the merged range. Then one `go-architecture-reader` each for
   `internal/ops`, `internal/ops/local`, and `internal/backup`, reading the worktree at the
   merged commit. Fold the drafts in `docs/superpowers/research/handoffs/draft-*-6b-i.md` into
   HISTORY and STATUS (STATUS 60 lines or fewer), land the drafted CLAUDE.md rclone Gate
   line, and remove the worktree and branch after the reads.
3. **5c-ii's close**, in the main tree: Task 8's close drafts (the plan's Task 8 section,
   including the owed errata and the 5c-iii handoff), then one `go-architecture-reader` each
   for `internal/auth`, `internal/serve`, and `internal/identity` (read-only), then the
   ledger fold. Carry the security review's M2 (the roster API shares an origin with the
   Navidrome fallthrough) to rung 12 as a question for Geoff.
4. **Three plans, as document lanes, in parallel once steps 2 and 3 have committed:**
   - **5c-iii:** fonts (the four woff2 files), the three real pages, the rendered gates
     (G7's computed half, G10), the render-and-read gate, and the identity spec 1.2 component
     layer: the new inks, the one focus rule, the refinement rules, the no-art sleeve if a
     5c-iii surface renders it. It carries the a11y reviewer's WCAG 1.4.11 question for
     night-primary fills, the Svelte review's S11 (reproducible builds) and S12 (absolute
     hrefs), and the security review's I2 (the login/link page consumes the token on a click
     only and drops `?t=` with `history.replaceState`). Every `web/` task reads the identity
     spec, per CLAUDE.md.
   - **6b-ii:** restore and the HTTP projection, from the rung 6 spec's pass table and 6b-i's
     handoff in its HISTORY entry.
   - **The matcher word-boundary pass:** from
     `docs/superpowers/research/2026-09-24-matcher-unicode-word-boundary.md` (Geoff ruled yes on
     2026-09-24): a documented Go translation of beets' featuring row, the translation-map test,
     the parity golden and its ruff-gated generator, the corpus case, and three unit rows.
   Each plan runs the full `spec-plan-review` cycle: three lenses, one fold, one verification
   read, a narrow second fold on a blocker or major, and the prose review. Each ends with its
   rulings consolidated and marked "awaiting owner approval".
5. **Morning handoff:** STATUS carries one combined question: the three plans' rulings, the
   rung 12 origin question, and any readings to ratify.

## Guards and recovery

- **Runaway guard:** arm `claude-wf-guard <transcript-dir> implementer <run-id>` in the
  background for every workflow run. An idle alarm after a run finishes is expected.
- **Sleep:** the activity lease and `cairn-run-gate` hold the inhibitor. Confirm AC power
  (`/sys/class/power_supply/AC*/online` reads 1) at the start and note it in STATUS if not.
- **Network drops** happened four times on 2026-09-24. When an agent dies with
  `EAI_AGAIN`, `EHOSTUNREACH`, or `ECONNREFUSED`, check the tree before anything else. A clean
  tree means relaunch with `resumeFromRunId` in the same session, plus a "start fresh" note on
  the dead task. Warm edits from a standalone agent mean resume that agent with SendMessage.
  Warm staged work from a workflow task means relaunch that task with a note to review and
  finish it.
- **Gate strings:** every dispatch names the exact absolute gate string and the exact string
  to report, with no `cd`, so a gate-string mismatch never escalates.
- **Spend ledger:** write STATUS at every segment boundary with the running total against
  the 12M cap.
