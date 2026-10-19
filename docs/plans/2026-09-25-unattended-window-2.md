# Unattended window 2, 2026-09-25

A run of about five hours, prepped in the attended session of 2026-09-25 and executed by a fresh
session after Geoff clears. It executes two approved passes in parallel. Lane 1 runs the matcher
word-boundary pass and lane 2 runs 6b-ii-a, the HTTP projection. Once the matcher pass merges, lane
1 starts 5c-iii and runs whole segments until the window ends. Nothing deploys.

## Grants (Geoff, 2026-09-25)

- **Scope:** the matcher pass and 6b-ii-a, each through its close and merge. Then 5c-iii in lane
  1, one whole segment at a time (A, then B, then C), stopping at a segment boundary on a green
  commit. 6b-ii-b does not start. 5c-iv does not start.
- **Spend cap:** 11M subagent tokens for the window. At the 80% flag (8.8M), start no new unit of
  work: finish the one in flight, write STATUS, and stop. Do not ask; no one is there to ask.
- **Box contact:** read-only ssh is allowed if a Task 0 staleness read needs a box fact. No
  write, no deploy, no restart. Nothing deploys until pass 5d.
- **Standing:** the conductor rules any fold, escalation, or second `fix` by the plan and the
  standing rules, and records each ruling in STATUS as a reading Geoff can ratify or reverse. Per
  memory `no-ceremony-rulings`, a question with a clearly correct architectural answer is taken
  and recorded. Only a genuine owner fork (product, taste, a risk only the owner can accept) goes
  into STATUS's morning question, and the lane it blocks stops at its last green commit.

## Starting state

Read `docs/STATUS.md` first. All three plans are approved, with their rulings recorded in each
plan (Geoff, 2026-09-25). No code lane is open, and master is the only worktree.

- Matcher: `docs/superpowers/plans/2026-09-25-matcher-word-boundary.md`. Per-task Agent dispatch,
  sequential, in the main tree (corpus and harness work never moves to a worktree). Ceiling 2.75M,
  flag 2.2M.
- 6b-ii-a: `docs/superpowers/plans/2026-09-25-rung-6b-ii-restore-http.md`, split per Ruling 1:
  Task 0, Tasks 1 to 3, then the 6b-ii-a close. pass-execute in a new worktree
  `.claude/worktrees/lane-2-6b-ii-a` on branch `lane-2-6b-ii-a`. Ceiling 3.90M, flag 3.12M.
- 5c-iii: `docs/superpowers/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`, split per
  Ruling 1. pass-execute, one invocation per segment (A: Tasks 1, 2; B: 3a, 3b; C: 4, 5, then its
  simplifier, reviewers, and fold). Task 0 runs first in fresh context at `opus`. Ceiling 10.8M,
  flag 8.64M; this window funds part of it.

## Sequence

1. **Launch both lanes.** Run the lane-launch checklist for each: reverse edges cleared, contested
   surfaces written (both plans name theirs; the one shared file is
   `cmd/dubplate-server/main.go`, if 6b-ii-a touches it), runaway guard armed, lane registry
   entry written in STATUS, sub-ceiling assigned, AC power checked, throughput baseline noted.
   Lane 2's worktree needs `npm ci --prefix <worktree>/web` before its first gate. Each lane
   starts from its plan's Task 0, and Task 0's blockers are ruled before Task 1.
2. **Matcher pass (lane 1).** Run its tasks by per-task Agent dispatch, as its plan's "Execution
   model" states. Then its close: the conductor adds the new generator to CLAUDE.md's two
   tracked-Python lists (the plan's conductor action 5), the simplifier round, one
   `go-architecture-reader` per touched Go package, and the ledger fold. The lane works on master
   directly, so no merge ritual applies.
3. **6b-ii-a (lane 2).** pass-execute with segment A1 (Tasks 0 to 3), then the 6b-ii-a close:
   simplifier, `web-auth-security-reviewer` on the projection, Task 8's drafts (suffixed
   `-6b-ii-a`), then the merge ritual. That ritual is a rebase onto master, the full gate in the
   worktree, `--ff-only`, the full gate on master, and one `diff-reviewer` if the rebase was not a
   fast-forward. Then the architecture reads, the ledger fold, and removing the worktree.
4. **5c-iii (lane 1, after the matcher close commits).** Task 0 first, then segments A, B, C in
   order, one pass-execute invocation each. Before each segment, check the spend flag. From Task 3a
   on, every full gate launches Chromium (the heavy lane). If lane 2 is still gating at the same
   time, expect lock waits: `cairn-run-gate` prints a NOTE, and the NOTE is not a failure. The
   5c-iii close (Task 6) runs only if segment C finishes with budget to spare under the flag.
5. **Morning handoff.** STATUS carries the ledger (what merged, spend against 11M), the
   readings taken, and any owner fork in one combined question. Update the memory handoff entry.

## Guards and recovery

- **Runaway guard:** arm `claude-wf-guard <transcript-dir> implementer <run-id>` in the background
  for every workflow run. For per-task Agent dispatches, arm it on the session's subagent
  transcript directory. An idle alarm after a run finishes is expected.
- **Sleep:** the activity lease and `cairn-run-gate` hold the inhibitor. Check that
  `/sys/class/power_supply/AC*/online` reads 1 at the start, and note it in STATUS if not (the
  last window started on battery).
- **Network drops:** when an agent dies with `EAI_AGAIN`, `EHOSTUNREACH`, or `ECONNREFUSED`, check
  the tree before anything else. A clean tree means relaunch with `resumeFromRunId` in the same
  session, with a "start fresh" note on the dead task. Warm edits from a standalone agent mean
  resuming that agent with SendMessage. Warm staged work from a workflow task means relaunching
  that task with a note to review and finish it.
- **Gate strings:** every dispatch names the exact absolute gate string for its tree and the
  exact string to report, with no `cd`. Lane 1:
  `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`. Lane 2:
  `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-2-6b-ii-a/scripts/check.sh'`.
- **Committing:** one committing agent per tree at a time. The conductor commits docs
  path-limited, never `git add -A`.
- **Spend ledger:** write STATUS at every segment boundary with the running total against the 11M
  cap.
