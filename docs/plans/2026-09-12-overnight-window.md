# Overnight window, 2026-09-12 to 2026-09-13

Status: **revision 2**, folded 2026-09-12 01:20 AKDT from two adversarial
reviews of revision 1 — a lane-mechanics review (M0 through M17,
`scratchpad/plan-review-mechanics.md`) and a budget-and-guards review (B1
through B5, M1 through M8, m1 through m5, G1,
`scratchpad/plan-review-budget.md`). Every finding is folded, declined, or
recorded as already fixed below; the conductor's ruling is stated at each
site.

Geoff's grants, given 2026-09-12 between 00:30 and 00:45 AKDT: the
sequence below as proposed; a 10M token cap for the window with the 80%
flag routed to the conductor; AC power (plugging in at 00:41); read-only
`deploy.sh --check` on the box and no writes; signed off for the evening
at 00:44, so no attended checkpoint exists until about 09:00.

**The grant reading, flagged as a reading (G1).** "Then start" was given
against the plan that named ruling the twelve conductor-owed folds by the
fold editor's recommendation, and this document reads that as approval of
that ruling. The reading is an inference from a terminal "then start"
following an open question Geoff never closed, not an explicit worded
grant, and it is flagged as such in the first line of STATUS's next-action
block so a wrong reading costs one morning edit. The precedent is the
2026-09-07 window's explicit grant ("rule recommended rulings myself,
listed for ratify-or-reverse"). The twelve are listed once, in the
ratify-or-reverse table below, rather than only marked at twelve sites
across 2858 and 1682 lines.

The window runs under the standing grants in CLAUDE.md's build process
section (workflow mode, batched questions, continue unblocked work, two
code lanes plus document lanes) and follows the 2026-09-07 window's shape,
recorded in memory as `unattended-window-2026-09-07`.

## What the reviews changed, at a glance

- Two blockers are **already fixed** and are recorded, not re-planned:
  M1 (the gate red on master) at `5309c74`, and M0 (the gate lock and
  `.gitleaks.toml`) at `3e15bd6`. Both landed before this revision.
- `pass-execute` runs every task in one uninterrupted call, so both
  passes are **segmented at their conductor boundaries**: five launches,
  not two (M2, M13).
- Every gate string is **absolute** (M6); the document lane is confined
  and never commits (M4); the lane-2 worktree is created **before** the
  Task 0 dispatches (M7).
- A **spend ledger** now measures the window, so the cap and the flag are
  enforceable for the first time (B1). The budget table is corrected and
  sums to the cap (B2); the 5b-ii tail is cut (M16).
- Six stop conditions are added (B-review M3) and the guard set is stated
  as armed with names and times (B3).

## Sequence

### 1. Pre-flight, in this session, before any lane launches

**1.0 — Already landed; verify, do not re-plan.**

- **M1 is fixed.** The two `generic-api-key` hits in
  `docs/superpowers/research/2026-09-08-rung-5b-plan-review.md` were
  quoted example literals; they were rewritten at `5309c74` ("Rewrite two
  quoted example literals in the 5b plan review so gitleaks no longer
  fires on master"), by the rename remedy the house rule prescribes
  rather than a `.gitleaksignore` entry. Master's gate is green again.
- **M0 is fixed.** The uncommitted `scripts/check.sh` flock and the
  untracked `.gitleaks.toml` the mechanics review found warm in the tree
  were the conductor's own implementer dispatch, in flight during that
  review. It committed at `3e15bd6`: a `flock` on
  `$(git rev-parse --git-dir)/check.sh.lock`, which serializes each tree
  against itself and not against the others, and a tracked
  `.gitleaks.toml` allowlisting `.claude/worktrees/`. **`.gitleaks.toml`
  is now tracked, which is the precondition `git worktree add` needs: a
  worktree checks out tracked files only.** Lane 2 launches only after
  that commit, which it now is.
- **Still not built, and recorded as deferred (B5, M5).** `check.sh` has
  no fast/full scope flag and `make tidy-check` still mutates `go.mod`.
  The flock closes the `go.mod` race inside a tree, which was the whole
  exposure; the scope flag is **not built this window** and every task
  pays the full three legs. Carried to the next attended session.
- Confirm before step 1.1: `git status --porcelain` shows no warm change
  under `scripts/`, `internal/`, or `cmd/`; `bash scripts/check.sh` is
  green on master; `git worktree list` reports the main tree alone.

**1.1 — Rule the twelve folds and commit both plans as revision 2.1.**
The ratify-or-reverse table below is the ruling of record. Mark each fold
at its site with the ruling and the date, and commit both plan files.
Step 1.1 must be **fully committed** before step 1.4's Task 0 dispatches,
because those tasks read these files.

**1.2 — Refresh both plans' `docs/STATUS.md` pointers (M8).** The
2026-09-11/12 design sitting rewrote STATUS at `0f0748f`, after both
plans were resolved against `6a9570d`. In the same revision-2.1 commit,
re-resolve every `docs/STATUS.md:` citation in both plans against the
live file: the registry table is now `## Lane registry` at `:59-64`, not
the Next action prose at `:22-33`; the sixty-three coverage-gap total is
`:80-82`, not `:36-38`. Correct 5b Task 0 item 1's "Expected" block,
which still says lane state lives in the Next action section. Without
this refresh both Task 0s return a stale verdict, which under the stop
conditions halts both lanes.

**1.3 — Write the contested-surface rows into the ownership table.** The
rung 5 spec's rows and 8c's `internal/engine/shadow` amendment go into
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`'s
ownership table, as STATUS's resume prompt requires. Include the M17
amendment verbatim (see the fold table, id 13).

**1.4 — Create the lane-2 worktree, then dispatch both Task 0s (M7).**
Revision 1 ran the Task 0s before the worktree existed, which 5b's Task 0
item 1 cannot answer and which makes 8c's Task 0 item 1 return an answer
Task 1's gate baseline then contradicts. The order is fixed:

1. `git worktree add .claude/worktrees/lane-2-5b -b lane-2-5b master`
2. Write both lane registry entries into `docs/STATUS.md` — lane name,
   tree path, state, and the workflow run id once known — **before the
   first dispatch**, since the registry is the only valid live-executor
   check from here on and `pgrep -f Projects/dubplate` matches every lane
   once a worktree exists (m5).
3. Run the lane-launch checklist from the parallel-build approach,
   section 4, for each lane: reverse edges (cleared for both, per their
   plans), contested surfaces (step 1.3), runaway guard, lane registry
   entry, venv (needed by neither lane), sub-ceiling, AC power (see the
   power gate below), throughput baseline.
4. Dispatch each plan's Task 0 in fresh context at `opus`, one per plan,
   both at once. Each dispatch carries a scope expectation and a hard
   turn-level token target of **0.3M**; a staleness check costing more
   than that has failed at its job and stops its lane (B3).

**Throughput baseline, corrected (m4).** The measured chain figure is
**0.19M per task** (0.093M-0.097M per agent, two agents per task, across
four passes and 74 agents), not the 0.25M revision 1 cited. Close cost is
a separate line at 0.9M-1.4M per pass. Compare a lane's mid-flight burn
against 0.19M per task.

### 2. Two code lanes at once, segmented at their conductor boundaries

`pass-execute` loops over `args.tasks` with no pause point and no
conductor hook: `runTask` is implement → review → fix-loop → return, and
`main` iterates to completion and returns one tally. It cannot honor
either plan's mid-pass conductor steps, nor the window's own checkpoint
cadence, nor a decision on a second `fix` verdict before every later task
has already run (M2, M13). Both passes therefore run as **multiple
launches**, segmented at the boundaries their own plans name.

**Lane 1 — pass 8c, main tree, sub-ceiling 2.1M, flag 1.68M. Two
launches.**

| Segment | Tasks | Conductor work at its boundary |
|---|---|---|
| 1A | 0 through 3 | batched `code-simplifier` round over the segment's accepted Go, committed on its own; the read-only box reading for R4's denominator; STATUS checkpoint; HISTORY draft harvest |
| 1B | 4 | close; full gate on master; STATUS and HISTORY folded |

8c's execution model already places both conductor steps between Task 3
and Task 4, and declares Task 4 to depend on the simplifier round, so the
segmentation is the plan's own boundary rather than a new one.

**Lane 2 — pass 5b-i, worktree `.claude/worktrees/lane-2-5b` on branch
`lane-2-5b`, sub-ceiling 5.0M, flag 4.0M. Three launches.**

| Segment | Tasks | Conductor work at its boundary |
|---|---|---|
| 2A | 0 through 3 | batched `code-simplifier` round; STATUS checkpoint; re-derive the ceiling from actuals, as the plan requires after Task 3 |
| 2B | 4 through 7 | batched `code-simplifier` round; STATUS checkpoint |
| 2C | 8 through 10 | close; security cycle under its gate; merge ritual |

**Recorded deviation from `5b:2730-2734`.** That plan calls for
`code-simplifier` "after that task's review accepts and before the
close" — eleven interleaved dispatches. **Ruled: replaced by one batched
round per segment, three rounds rather than eleven.** The chain's
implementer commits each task's files, so CLAUDE.md's run-the-simplifier-
before-committing rule is already deviated from on every task the chain
runs; 8c states that as its deviation 1 and 5b does not. This is the same
deviation, stated. Eleven extra dispatches is not what the 5.0M ceiling
priced.

**Absolute gate strings (M6).** `pass-execute`'s `repo` argument is
interpolated into a prompt string and nothing else: there is no `cwd`, no
`chdir`, no shell invocation in the workflow. A subagent's shell working
directory resets between Bash calls, so an implementer running
`cairn-run-gate 'bash scripts/check.sh'` from its default cwd gates **the
main tree** and reports it green. Both gate strings are absolute, and
`check.sh` derives `REPO_ROOT` from `BASH_SOURCE` and `cd`s there, so an
absolute path pins the tree regardless of the caller's cwd:

- Lane 1: `bash /var/home/glw907/Projects/dubplate/scripts/check.sh`
- Lane 2: `bash /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-2-5b/scripts/check.sh`

Additionally, **every task's `notes` field names its tree's absolute path
as the working directory**, instructs that every Bash call is either
absolute or re-`cd`s, and requires the implementer's report to name the
tree its commit landed in (`git -C <path> log -1`).

**Lane 1 commits straight to master (M14).** 8c runs in the main tree and
names no branch; the implementer is instructed never to switch or create
one. Its five task commits land on master directly. This is a standing
departure from CLAUDE.md's "if on the default branch, branch first",
established at 8b, and stated here because this is the document that
should state it. **Lane 1's merge ritual therefore reduces to the full
gate on master after each segment** — no rebase, no `--ff-only`, no
merged range for a `diff-reviewer`. The merge token exists to serialize
lane 2's `--ff-only` against lane 1's task commits.

**Lane 2's merge ritual is the full one** (parallel-build approach,
section 3): merge token, rebase onto master, full gate in-tree,
`--ff-only`, full gate on master, one `diff-reviewer` over the merged
range. Master moves under lane 2 throughout, so the rebase will not be a
fast-forward and the extra review fires as priced. The rebase should be
textually clean: the territories are disjoint by package, and the one
shared directory, `cmd/dubplate-server`, is excluded by 8c.

**Ledger resolution rules, restated (M11, B-M5).** Lane branches never
commit `docs/STATUS.md`, `docs/HISTORY.md`, or `ROADMAP.md`. On a rebase
conflict, **STATUS takes master's version and HISTORY keeps both**. The
conductor's own checkpoint commits are path-limited
(`git commit -o docs/STATUS.md`) so they never sweep an implementer's
staged index. One committing agent per tree at a time.

**The rung 5 security cycle (M16, B-M4).** The full fresh-context
`web-auth-security-reviewer` read runs at 5b-i's close **only if the
spend ledger reads at least 1.2M below the flag at that point** — that
is, at most 6.8M. Otherwise it does not start at all and becomes the
first item of the next attended session, recorded in STATUS. The cycle is
**atomic**: read, fold, record. A started read whose fold cannot complete
blocks 5b-i's merge rather than merging with findings outstanding.

### 3. Document lane, alongside, confined (M4)

The web surface standard's fold runs in the main tree beside lane 1,
under three constraints that make the shared tree safe:

- **It writes only under `docs/superpowers/specs/` and
  `docs/superpowers/research/`, and it never commits.** The conductor
  commits its files at a segment boundary, path-limited. Both pass plans
  state "no second executor enters this tree"; a non-committing lane
  confined to two doc directories does not contend for the index, and the
  gate lock at `3e15bd6` serializes any gate run it makes against lane
  1's.
- **Its inputs are harvested to disk before anything else (M15).** The
  seven adversarial reviews of the standard are in-session agent returns;
  a compaction or harness restart loses them. Its first action is writing
  each review's output to a file under `docs/superpowers/research/`
  named for the standard, so the fold survives a restart. The harvest
  inputs are already on disk at
  `docs/superpowers/research/2026-09-12-cairn-harvest-reference.md` and
  in the session scratchpad.
- **It is interruptible at a stage boundary.** Stages: harvest → fold →
  prose pass → second fold → hand to the conductor for commit. The
  standard is declared commit-worthy **after the first fold**, so if the
  ledger reads 8M the prose pass is held to the next session and the fold
  lands committed rather than half-applied in the working tree.

The standard's routing into rungs 5c and 7 changes no plan this window
executes.

### 4. Tail: cut (M16)

The 5b-ii prep tail is **cut from this window**. It launches nothing,
the next attended session can prep it cheaply against a quiet tree, and
its three named inputs — the security read's findings record, the
re-derived ceiling from 5b-i's actual, and Task 10's committed handoff —
are not all guaranteed to exist inside this window. STATUS records 5b-ii
prep as the next session's first item after the security cycle.

## The twelve folds: ratify-or-reverse

Ruled by the conductor 2026-09-12, **with the recommendation as read at
each site**. Where the site states no recommendation or states one
ambiguously, the ruling is **"as applied by the fold editor at that
site"**, and the table says so in the Ruling column. Membership is the
**union of both plans' lists**, which resolves 8c's two conflicting
memberships (`8c:20-22` omits R2; `8c:1430` includes it) in favor of
including R2 — the one item 8c makes a hard precondition of Task 2
(M10). Correct `8c:20-22` and STATUS's copy of it in the revision-2.1
commit so the two lists agree.

**Never use the bare word "fold" to mean the item being ruled** (M9).
`8c:548` says each is applied *as the recommendation*, so the Applied
text is the ruling; `8c:611` uses "fold" for the *alternative*. In this
table, "as applied" always means the Applied text.

| # | Id and site | Recommendation as the plan states it | Ruling | Reversal cost |
|---|---|---|---|---|
| 1 | 5b epoch anchor check (A8, Task 9) — `5b:847`, `:2465` | none at site; `docs/STATUS.md:38-41` states "rule all nine as applied" | As applied: the epoch compares against the session's **anchor**, and a token whose mint time predates the epoch is also refused. The alternative is a silent revocation bypass. | **Code-bearing** (Task 9) |
| 2 | 5b hook's environment snapshot (A2, Task 2) — `5b:727`, `:1511` | as above | As applied: the source's validation hook receives `os.Environ()` alongside the resolved values. | **Code-bearing** (Task 2) |
| 3 | 5b token's wire encoding (A4, Task 5) — `5b:772`, `:1937` | as above | As applied: base64url each field before joining. Raw joining breaks every address whose local part carries a dot, with every ordinary test green. A correctness bug, not a taste call. | **Code-bearing** (Task 5) |
| 4 | 5b nonce record's carried value (A7, Task 8) — `5b:826`, `:2341` | as above | As applied: the record stores consumed-at plus the resolved retention, not the token's own expiry. | **Code-bearing** (Task 8) |
| 5 | 5b malformed-record retention (Task 8) — `5b:833`, `:2352` | as above | As applied: retain an unparseable nonce record until its mtime passes retention. Deleting on sight would make that token consumable again for the rest of its expiry. | **Code-bearing** (Task 8) |
| 6 | 5b R33 / R-V / R24 routing (primary input 1, Inheritance) — `5b:157`, `:330` | as above; no alternative stated | As applied: the token carries no destination field at 5b-i and the rule travels to 5b-ii and 5c as a constraint. | Doc-only |
| 7 | 5b's 5.5M overage against R23's reservation (ceiling section) — `5b:522`, `:2752-2757` | as above | As applied: record the overage rather than funding it from another pass. Arithmetic verified by the review: 4.25 + 5.0 + 4.5 + 4.5 + 3.0 = 21.25M against a live 15.75M. | Doc-only |
| 8 | 5b R38's loopback widening (Task 4) — `5b:1829` | as above | As applied: the HTTPS carve-out covers hostname `localhost` **or** a parsed loopback address. The narrow reading refuses a loopback-IP adopter for a reason that does not apply, against the universality bar. | **Code-bearing** (Task 4) |
| 9 | 5b lane registry's home in STATUS (Global Constraints, Task 0 item 1) — `5b:550-555`, mark at `:554` | "either, conductor's choice" — **ambiguous at the site** | **Moot; do not rule.** `docs/STATUS.md:59` already carries a `## Lane registry` heading and table, landed at `0f0748f` after this plan was written. Strike the mark as resolved by the tree and fix Task 0 item 1's stale "Expected" premise (step 1.2). | Doc-only |
| 10 | 8c R1, the projection export and the 8b handoff half — `8c:558` (heading `:556`); deviations 4 at `:1583` and 5 at `:1595` | `8c:548`, "each is applied as the recommendation" | **As applied**, stated to break `8c:611`'s inverted use of "fold": the shadow consumes `Parse` and `tagio.Tags`, taking the 8b handoff's **second half**; `tagio` gains one exported symbol; one parse per file; 8b's coverage gap 4 retires; **Task 2 stays `sonnet`**. Deviation 5 is ruled with this and needs no separate mark (`8c:1601-1602`). | **Code-bearing, architectural** — the most expensive reversal in the table |
| 11 | 8c R2, the seven reader tests' destinations — `8c:627` | `8c:548` | As applied: two of the five spec-named tests stay in `shadow` because their subject is `matcher.ItemFields` and ruling 3 forbids `tagio` naming that type; the seven-row destination table stands. The alternative costs a projection with eighteen rows and no test. **8c's execution model makes this a precondition of Task 2's dispatch**, which the list at `8c:20-22` omitted. | **Code-bearing** (Task 2) |
| 12 | 8c R4 item 2, the composed denominator — `8c:694` | `8c:548`; the site is open-ended by design ("a better composition found while reading the report replaces this one") | As applied by the fold editor at that site: `ObservedApplied` + `ObservedSkipped` + the `Unreconciled` counts, read by the conductor through `dubplate-server shadow report`. The site's own escape hatch stands: a better composition found while reading the report replaces it. | Doc-only (Task 4's close prose) |
| 13 | **Amendment, not a fold** — M17, 5b Task 3 and `keys.go` | — | Task 3 genuinely edits `internal/config/keys.go`: two deletions and three supersession-message rewrites (`5b:1643-1647`). The rung 1b rule and CLAUDE.md both say "never by editing `internal/config/keys.go`", and they mean **no lane declares keys there**. Rung 5's twelve names are declared in `cmd/dubplate-server/auth.go` (R-5b2), which is the rung-5 file the registry rule requires; Task 3's `keys.go` edits are **retirements** of a reserved class and a pending wall row plus message rewrites, which is the rung 1b contract's removal half. **The task is not deferred.** This sentence goes verbatim into the ownership table (step 1.3) and into Task 3's `notes` field so the implementer reads it in its dispatch prompt rather than stopping to report. | Doc-only (a permission, already reflected in the plan) |

**Sequencing the code-bearing folds (B-M8).** The instruction is to land
code-bearing folds as late as the plan's task order allows. Both plans'
task orders are fixed by stated dependency chains (5b: 2→3→4→5→6, 7→8,
{5,6,7}→9, all→10; 8c: 1→2→3→4), so **no reordering is available**, and
the rule is discharged as recorded rather than as a change. The lever
that does exist is the segment boundary: ids 1, 4, 5 land in lane 2's
last segment (2C), id 8 in 2B, ids 2 and 3 in 2A and 2B. Id 10 is the
single most expensive reversal in the window and lands in 8c's Task 2, in
segment 1A, which cannot be moved without holding the whole lane. It is
ruled as applied and flagged here so a morning reversal is priced with
eyes open.

## Budget

**Spend is measured (B1).** Revision 1 named a cap and a flag with no
mechanism that produces a number while the window runs, which is how
1b, 8b, and 5a each crossed a flag that was only reconstructed
afterwards. The ledger armed for this window is `scratchpad/spend.py`,
summing `usage` fields (input, cache-creation, cache-read, output)
across this session's subagent transcripts under
`~/.claude/projects/-var-home-glw907-Projects-dubplate/`, **emitting
every 30 minutes with the flag tripping at 8M**. Its unit is calibrated
to the harness's own per-agent counts, which is the unit the HISTORY
figures are denominated in, so the 8M flag is on the right scale. Its
**baseline is the window spend at 01:00: 1.89M**, covering the seven
standards reviews, the two adversarial plan reviews, this fold, and the
gate implementer. The conductor reads the ledger file at every checkpoint
and records the number in the checkpoint STATUS block; a per-lane figure
is scoped by restricting the scan to each workflow's
`<session>/subagents/workflows/<runId>/` directory.

| Item | Reservation | Basis |
|---|---|---|
| Window spend to 01:00 (actual, counted) | 1.89M | ledger baseline: seven standards reviews, two plan reviews, this fold, the gate implementer |
| Pre-flight: twelve folds ruled and marked, STATUS pointer refresh, ownership rows, revision-2.1 commits, worktree, guard arming | 0.5M | conductor inline edits across two large plan files |
| Two Task 0 dispatches (`opus`, fresh context) | 0.3M | 0.10M-0.14M measured, x2, larger plan files; hard 0.3M cap per dispatch |
| Lane 1 — 8c: two chain segments (4 tasks, 8 agents) plus close | 2.1M | 0.84M chain at 0.19M/task + 1.2M close |
| Lane 2 — 5b-i: three chain segments (10 tasks, 20-22 agents) plus close | 3.5M | 2.1M chain + 1.4M close |
| Document lane: fold, prose pass, second fold, hand-off | 0.5M | its seven reviews are already inside the 1.89M above |
| Conductor turns, five segment checkpoints, STATUS writes, guard re-arms, merge orchestration, HISTORY drafts | 0.4M | two lanes, a checkpoint per segment, 0.15M per HISTORY draft |
| **Subtotal committed** | **9.19M** | |
| Reserve (variance allowance; 1b overran its ceiling by 48%) | 0.81M | |
| **Cap** | **10.0M** | |
| Rung 5 security cycle (read, fold, record) — **conditional** | 0.6M | drawn against the reserve only if the ledger reads at most 6.8M at 5b-i's close (M16); 5a measured 0.40M on a smaller surface |
| 5b-ii prep tail | **cut** | M16; the next attended session's item |

The 80% flag is 8M. At the flag: finish the task in flight, write STATUS,
hold the document lane's prose pass, and take no new segment. Spend
before this window opened is about 61.6M on the initiative and is not
counted against the cap.

## Guards

**Armed at 01:05 AKDT, the full set, with names (B3).** Revision 1 ran
the two `opus` Task 0 dispatches — the most runaway-shaped dispatches in
the window — before any runaway guard existed. The set is:

1. **Two sleep inhibitors** named `dubplate-overnight`, `suspend` only,
   never `suspend:idle`, held to about 09:02 AKDT from the clock
   (`sleep 29843`, armed ~00:45). The cairn session's `cairn-overnight`
   inhibitor belongs to another session and is never touched.
2. **The battery watchdog**, triggering at 11% while `Discharging`,
   silent on AC, globbing `BAT*` and taking `head -1`.
3. **The runaway guard**, armed over this session's subagent transcript
   directory **before the first Task 0 dispatch**, at the 2MB
   research-reader threshold with a 25-minute idle threshold. Per-lane
   guards are added at each lane launch, 2MB where a research reader
   runs and 900KB otherwise. **The document lane gets one too** (m1) — it
   is the window's highest-agent-count activity, and seven concurrent
   Opus agents produced 1.99M in one workflow four days ago.
4. **The spend ledger** (above), emitting every 30 minutes.

Every guard touches a heartbeat file on each poll. Any harness restart
**and any harness cull under memory pressure** re-arms the whole set as
one checklist; the harness kills background Bash guards when free memory
is low, so restarts are not the only way the set goes stale.

**The power gate, corrected (B4).** Revision 1 gated the code lanes on
the power supply reading `Charging` or `Full`. That reads the wrong
sensor twice: a `*/status` glob matches the two `ucsi-source-psy-*` USB-C
entries, and this ThinkPad carries a charge threshold, so `Not charging`
on AC is the normal steady state at any threshold below 100 and the gate
would poll forever on a perfectly healthy machine. **The gate reads mains
presence:**

```bash
[ "$(cat /sys/class/power_supply/AC*/online 2>/dev/null || cat /sys/class/power_supply/ADP*/online)" = "1" ]
```

AC was confirmed present at 00:42. `BAT*/status = Discharging` is the
only fail state. The poll is bounded: if mains is still absent after 20
minutes, the document lane launches alone, STATUS records that the code
lanes were withheld for power, and the conductor does not poll to 09:00
in silence.

**Checkpoint health check (m2).** At every checkpoint, four commands:
read the spend ledger; `systemd-inhibit --list` (both `dubplate-overnight`
holds present, `cairn-overnight` untouched); guard heartbeat mtimes;
`journalctl -b --since <last checkpoint> | grep -i suspend`. Three of the
six added stop conditions are only detectable here.

**Pens-down (m3).** The inhibitors expire at about 09:02. **No new
dispatch after 08:15.** Checkpoint, STATUS, HISTORY drafts, and guard
release complete by 08:45. A chain returning at 08:55 with STATUS
unwritten can meet a suspend mid-handoff, which is the one failure that
loses the window's record rather than its work.

## Stop conditions

Any of these stops the affected lane, WIP-commits partial work on the
lane's branch, writes STATUS with the exact resume prompt, releases the
inhibitors when nothing else runs, and reports.

From revision 1:

- a product fork in either plan that the plan does not settle;
- a Task 0 verdict of stale that a re-fold cannot clear without a ruling;
- a second `fix` verdict on one task (the conductor's decision, taken by
  the recorded rule: accept with a plan revision note when the finding is
  comment-only or a naming choice, else re-dispatch once with `opus`,
  else stop);
- an `escalate` that names a contested surface the ownership table does
  not settle;
- the battery floor;
- the cap, or the flag with a task that cannot finish inside it;
- a merge whose rebase is not clean and whose merge agent cannot resolve
  it from both task specs.

Added by the budget review (M3), each with its incident of origin:

- **A `diff-reviewer` that never returns.** The 3a incident left a task
  committed yet unreviewed when an interrupt hit its reviewer. The
  checkpoint STATUS block names any task that is committed but whose
  review has not returned. On a stalled reviewer: `TaskStop`, then
  dispatch that one review standalone rather than restarting the chain.
  An intervention at 04:00 is recoverable; the same one discovered at
  09:00 is not, because `resumeFromRunId` is same-session only.
- **A Task 0 that runs past its scope.** Hard turn-level target 0.3M;
  past it, stop the lane.
- **An implementer looping on a red gate.** The second-`fix` condition is
  a *reviewer* output; an implementer that cannot get the gate green
  never reaches a reviewer and loops inside one dispatch, inflating its
  own transcript in a shape the runaway guard reads as a token runaway.
  **Cap in-dispatch gate attempts at three**, then report red with the
  failing leg's tail and stop the lane. Confirm with a tail sample before
  killing anything.
- **A harness cull of the background guards.** Heartbeat mtimes stale, or
  `systemd-inhibit --list` short a hold, re-arms the full set; a cull
  that recurs stops the window.
- **A suspend recorded in `journalctl`.** A suspend means the inhibitors
  failed and every in-flight API stream may be orphaned. That is a stop,
  not a diagnosis.
- **The weekly Fable pool.** Restored from the 2026-09-05 grant, which
  revision 1 dropped with no stated reason: stop at the cap **or earlier
  if the weekly Fable pool nears its 50% cap**. A cairn overnight
  workflow may be running concurrently and draws the same pool; check the
  pool state at each checkpoint, and never silently spend Fable credits.

**The gitleaks remedy, pre-authorized in every implementer prompt (M2 of
the budget review).** 5b-i builds token, session, and nonce records, so a
`generic-api-key` hit on an authored fixture is likely. One sentence goes
into every implementer dispatch: *a gitleaks hit on a fixture you
authored is resolved by renaming the field or lowering the fixture
value's entropy, never by a `.gitleaksignore` entry and never by editing
the gitleaks leg; report the rename as a decision the plan did not
cover.* A gitleaks hit the rename remedy does not clear is a stop.

## Box contact

Read-only `deploy.sh --check` is allowed for evidence. **Widened for
R4's denominator (M12):** `ssh <box> dubplate-server shadow report`,
read-only, run by the conductor under the quiet-box check, never the bare
binary as root — the command STATUS already blesses. It runs at lane 1's
1A/1B boundary, after Task 3 and before Task 4's draft, so the close
carries the denominator or the reason it does not. No writes, no deploy,
no unit changes. Neither plan deploys.

## Close-task ledger discipline (M11)

Both plans' close tasks modify `docs/STATUS.md` and `docs/HISTORY.md`
without committing them. In a worktree that leaves the tree dirty at
exactly the moment the merge ritual rebases it, and `git rebase` and
`git worktree remove` both refuse a dirty tree. The instruction to both
implementers, in the task `notes`:

- **Write the ledger text as drafts, never edit the ledgers.** 5b-i Task
  10 writes `docs/superpowers/research/handoffs/draft-STATUS-5b-i.md` and
  `docs/superpowers/research/handoffs/draft-HISTORY-5b-i.md`. 8c Task 4
  writes `draft-STATUS-8c.md` and `draft-HISTORY-8c.md` in the same
  directory, in the main tree, plus its `ROADMAP.md` text in the same
  draft.
- The conductor folds the drafts into the real ledgers at the merge, then
  commits them path-limited.
- The drafts are committed artifacts, so they survive a restart and a
  fresh session can read them.

**A HISTORY entry draft is required per closed segment**, not only per
closed pass (M7 of the budget review). The material — what the gate
caught, what a later pass would be wrong to rediscover — lives in
implementer reports and reviewer verdicts that are gone once the chain
returns, and no later session can reconstruct it. One dispatched agent
writes the draft from the segment's reports before the conductor moves
on; 0.15M per pass is in the budget.

**STATUS is over its cap.** `docs/STATUS.md` stands at 103 lines against
the doctrine's 60. Pruning means moving: each lane's close moves its
history into `docs/HISTORY.md` as part of the close, and the window's
final STATUS is under 60 lines.

## Resume prompt

Written to STATUS at every segment boundary, at every split, and at any
stop. The first line of the next-action block carries the grant reading
flagged as a reading (G1).

**Per lane, every time:**

- lane, tree path, branch, plan path, and whether the tree holds
  uncommitted work;
- **the workflow `runId` and the absolute path to its transcript
  directory** (`<session>/subagents/workflows/<runId>/`), recorded
  always, not only for the same-session case;
- committed tasks, task in flight, last `diff-reviewer` verdict, last
  journal result;
- **any task committed whose review has not returned**;
- the spend ledger reading at this checkpoint, against the sub-ceiling;
- the next action.

**Window-level, once:**

- the session id (`7968f95f-d580-4a67-909a-eac9091b4953`) and the
  absolute scratchpad path;
- which guards are armed, with pids and the inhibitor expiry (~09:02), so
  the next session re-arms rather than orphaning or double-arming;
- the twelve fold rulings as ruled — a link to the ratify-or-reverse
  table above, not twelve site marks to hunt;
- the segment ledger: which of the five `pass-execute` launches have
  returned, and with what tally.

**The `runId` is kept for the same-session case and for the
cross-session case, for different reasons.** Same-session,
`resumeFromRunId` works. **Cross-session it does not**: a fresh session
relaunches fresh with the uncommitted tasks, and the runId is how it
finds the old run's `journal.jsonl` and `*.meta.json`. Journal first,
then relaunch. A task can be committed yet unreviewed, so the journal is
read before anything is assumed landed.
