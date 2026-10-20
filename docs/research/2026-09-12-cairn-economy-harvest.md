# cairn-cms pass economy: what it cut, where it is encoded, what it measured

Read-only harvest, 2026-09-12, for dubplate. Sources: cairn-cms `docs/HISTORY.md` (Chassis-A
2026-09-08 through Chassis-B2 2026-09-09, the newest entry in the file), `docs/STATUS.md`,
the chassis and polish plan post-mortems, `docs/internal/record/2026-09-08-polish-inputs/`,
workstation `~/.claude/CLAUDE.md`, `~/.claude/docs/model-economy.md`,
`~/.claude/workflows/pass-execute-chains.js`, `~/.claude/workflows/cairn-overnight-to-release.js`,
`~/.local/bin/cairn-run-gate`, `~/.claude/agents/cairn-implementer.md`,
`~/.claude/agents/diff-reviewer.md`.

The economy work clusters in three days: chassis-A (2026-09-05 to 09-08) exposed the cost shape,
chassis-B2 (09-08 to 09-09) measured the overrun, and the polish plans (09-08 onward) encoded the
fixes. The governing numbers: chassis-A about **4.4M tokens against a 7.5M ceiling**; chassis-B1
**839,281 subagent tokens for Tasks 5 to 7 alone**; chassis-B2 **near 12.7M against a 6M ceiling,
roughly double**, "concentrated in the pass-end ritual". At public API rates cairn now prices a task
cycle at **"about $8 Sonnet implementer plus $3.50 Opus review, about $12 with a fix round"**, a full
polish pass at **"$130 to $150 at list"** (`docs/STATUS.md`, the 2026-09-09 measured-cost paragraph).

---

## 1. Run `npm run package` once at the head of the gate string (wrapper change, not list change)

**What.** The per-task gate string used to chain twelve npm check scripts that each re-run
`npm run package` as a prerequisite, so one gate run built `dist/` thirteen times. The rewrite calls
each dependent check at its `node scripts/checks/*.mjs` entry point after one packaging at the head.
Every check still runs, in the same order, against the same artifact.

**Encoded at.** `cairn-cms/docs/plans/2026-09-08-polish-11a-pass.md:1558` (the gate
string), with the rationale at `:1574-1595`. Proposal and measurement at
`cairn-cms/docs/internal/record/2026-09-08-polish-inputs/plan-11a-review-executor.md:243-262`.

**Measured.** "At a conservative 40 to 60 seconds per build that is **9 to 13 minutes per gate run**
spent rebuilding an unchanged `dist`." Whole-string estimate "**25 to 40 minutes per gate run**, and
the pass runs it about nineteen times ..., so roughly **8 to 12 hours of gate wall clock, of which
2.5 to 4 hours is the repeated packaging alone**." After the rewrite: "**Expected wall clock per
gate: 16 to 28 minutes** ... Removing twelve builds removes 8 to 12 minutes."

**Quality safeguard kept.** Coverage-identical: "No check is dropped for cost; the spec forecloses
that." The pass-end and pre-merge run uses the original npm-script form "so the wrappers themselves
are proven at least once before merge" (`plan-11a:1596-1599`, `:1608-1610`).

**Precondition.** None of the twelve underlying scripts rebuilds (verified against `package.json`
and each script's body). Also a single-writer rule on `dist/`: the plan refuses to re-time the gate
on the main checkout because "a `npm run package` in the main checkout writes the `dist/` that a
live chassis-B2 worktree's showcase resolves through".

**Born from.** No incident; it is the executor-lens plan review finding that the spec's
"re-derive the gate list from `.github/workflows/`" rule had been read as "keep the wrappers too".

---

## 2. The showcase e2e leaves the per-task gate for paint-neutral tasks

**What.** A browser e2e / visual suite joins the per-task gate only for tasks that move rendered
paint. Paint-neutral tasks run the check suite and unit tests; the pass-end gate and CI on every
push run the full suite.

**Encoded at.** Workstation `~/.claude/CLAUDE.md`, "Gate economy on a pass", first bullet.
`cairn-cms/docs/specs/2026-09-08-polish-passes-design.md:597-600` (the 2026-09-09
amendment). `cairn-cms/docs/plans/2026-09-08-polish-11a-pass.md:1569-1572`: "the
per-task gate for a paint-neutral task is the fence's string with the trailing
`&& CI=1 npm --prefix examples/showcase run test:e2e` removed ... Every other task (1-6, 8-13) is
paint-neutral and drops the e2e clause." Mechanism: the per-task `gate` override field, see §5.

**Measured.** "Evidence: chassis-B2's **eight per-task e2e runs caught nothing the check suite, the
diff reviewer, and CI did not**" (spec `:598-600`). CLAUDE.md adds the clock share: "at **a third of
each task's hour**". Against the 11a gate that is roughly 8 to 12 minutes of the 25 to 40 (Playwright
over 33 spec files behind a preview build, `plan-11a-review-executor.md:250-253`).

**Quality safeguard kept.** Paint tasks keep the full fence; the pass-end ritual runs it regardless;
CI runs it on every push and "Nothing merges on a red CI"; a pass-end fresh-context `visual-verifier`
still grades the whole surface set.

**Precondition.** A real CI e2e job on push (`e2e.yml`), a plan that classifies each task
paint / paint-neutral, and a runner that accepts a per-task gate override.

**Born from.** chassis-B2's own per-task e2e runs producing no unique finding, at a third of each
task's hour.

---

## 3. Comment-only fix rounds run a reduced gate

**What.** The reviewer marks each blocking finding `commentOnly` when its fix changes only comment
or doc text and no code behavior. A fix round whose findings are all comment-only runs the comment
linters, the doc link gate, and the touched files' unit tests instead of the full string.

**Encoded at.** `~/.claude/workflows/pass-execute-chains.js:45` (`commentOnly` in REVIEW_SCHEMA),
`:90-92` (the implementer instruction: "run `npm run check:comments && npm run check:symbols &&
npm run check:docs` plus the unit test files that cover the touched files, through cairn-run-gate"),
`:103` (the reviewer is told to mark it honestly and to expect the reduced gate when reviewing such
a round). `~/.claude/agents/diff-reviewer.md:60` (the `[comment-only]` tag in the report format) and
`:67-70`. Workstation `~/.claude/CLAUDE.md`, "Gate economy on a pass", second bullet.

**Measured.** No isolated number; the saving is the difference between the reduced gate (a few
minutes) and the full 16 to 28 minute string, multiplied by the fix-round count cairn actually pays:
chassis-B2 recorded "**four of the five tasks from 4 onward needed a `diff-reviewer` fix round**"
(`2026-09-07-chassis-b2-pass.md` Budgets), each of which previously re-ran the whole gate for a
comment edit.

**Quality safeguard kept.** The classification is the reviewer's, not the implementer's, and the
reviewer re-verifies it on the fix round; any finding touching behavior forfeits the reduction.

**Precondition.** A structured reviewer verdict schema with a per-finding boolean, and repo-level
comment/doc linters worth running alone (`check:comments`, `check:symbols`, `check:docs`).

**Born from.** chassis-B2's fix-round concentration and the resulting 12.7M-vs-6M overrun.

---

## 4. `cairn-run-gate`: one blocking runner, no log polling

**What.** A dotfiles binary that starts the gate detached, keyed by `(cwd, gate string)`, blocks up
to 540s, and on timeout prints "gate still running" with exit 75 so the caller re-issues the same
command and reattaches. It prints the exit status and the last 60 lines only.

**Encoded at.** `~/.local/bin/cairn-run-gate:1-14` (contract), `:25-31` (detached start),
`:44-49` (the reattach instruction: "Re-run this exact cairn-run-gate command to keep waiting; do
not tail, wc, ps, or sleep on the log"), `:57-59` (exit + 60-line tail).
Callers: `pass-execute-chains.js:80`, `pass-execute.js:110`,
`~/.claude/agents/cairn-implementer.md:97-104` ("never run the gate or a test suite with
run_in_background, and never tail, wc, cat, ps, or sleep on a log (a transcript of such polls is a
task failure the conductor halts)"). Workstation `CLAUDE.md`: "Gates run through
`cairn-run-gate '<string>'` (dotfiles bin; generic despite the name)".

**Measured.** No minute figure. The token effect is stated as a defect it removes:
chassis-A's "What a later pass would be wrong to rediscover" says "**An implementer that polls its
own gate with no-op commands inflates its transcript into the runaway guard**"
(`docs/HISTORY.md`, Chassis-A entry; also `2026-09-04-chassis-a-pass.md` post-mortem). CLAUDE.md's
rationale: "it blocks to completion and prints the tail, so an implementer never polls a log and
**its transcript stays small for the reviewer**" — the reviewer reads the implementer's report, so
transcript bloat is paid twice.

**Quality safeguard kept.** The real exit status is reported, not a summary line; the implementer
must paste it, and the reviewer re-runs the gate when the report omits it.

**Precondition.** A shell tool with a ~10 minute cap (hence `CAIRN_GATE_WAIT` default 540) and a
gate string that is deterministic per working directory.

**Born from.** chassis-A implementers polling with no-op commands and tripping the runaway guard.

---

## 5. Per-task gate-string override and per-task model upshift in the chain runner

**What.** Each task record may carry `gate` (overriding the chain's gate string for that task alone)
and `model` (upshifting that task's implementer). This is the mechanism §2 rides on.

**Encoded at.** `~/.claude/workflows/pass-execute-chains.js:4-6` (header comment: "A task may carry
`model` ... and `gate` to override the chain's gate string for that task alone (paint tasks keep the
e2e)"), `:79` and `:106` (`t.gate || a.gate` in both the implement and review prompts), `:133` and
`:162` (`...(t.model ? { model: t.model } : {})`). Same override exists in `pass-execute.js:109,127`.

**Measured.** Carries §2's numbers; no independent measurement.

**Quality safeguard kept.** The reviewer prompt is handed the same overridden gate string, so the
reviewer grades against the gate the task actually ran, not a different one.

**Precondition.** The plan must state the per-task gate; an unstated override silently weakens a gate.

**Born from.** The e2e ruling needing a per-task, not per-pass, expression.

---

## 6. Scope the engine test suite to the blast radius

**What.** A task whose Files touch nothing under the engine's source runs the consumer's own unit
suite, not the engine's.

**Encoded at.** Workstation `~/.claude/CLAUDE.md`, "Gate economy on a pass", third bullet. In cairn
the same logic appears as the deliberate **left-to-CI list**:
`2026-09-08-polish-11a-pass.md:1600-1607` leaves `design.yml` (`check:public-tokens`, `test:reskin`,
the styleguide e2e), `norms.yml`, `scaffold.yml`, `create-site.yml` (the packed-tarball scaffold
proofs), `tsgo.yml` and `publish.yml` to CI, "None of them is skipped as unimportant; each needs a
clean checkout, a browser matrix, or a full pack that a per-task local loop cannot afford".

**Measured.** Not separately quantified; folded into the 16-28 minute per-gate figure.

**Quality safeguard kept.** "the PR's own CI run is the gate that clears them. Nothing merges on a
red CI." Two named exceptions run per task anyway: "The CLI suite and `check:template` run per task,
because a file move can carry pinned text" (chassis-A's lesson, spec `:590-591`).

**Precondition.** A CI that runs the omitted suites on every push, and a merge policy that blocks
on red CI.

**Born from.** chassis-A: the mechanical Prettier reformat reflowed a `githubApp` literal that
`create-cairn-site`'s finalize pins byte-for-byte, and neither the plan's gate nor the dispatched
gate had run the CLI suite — the reason the CLI suite is now a per-task exception.

---

## 7. Pre-flight checklist in the implementer agent definition

**What.** Five checks the implementer runs before reporting, each one a fix round cairn already paid
for: no comment claims what its assertion does not prove; the labeled report block verbatim; no
process citations (pass, plan, ruling id, task number) in shipped comments; counts found / changed /
deferred with the deferred list named; re-emit any generated tree before the gate and commit it in
the same commit.

**Encoded at.** `~/.claude/agents/cairn-implementer.md:86-95` ("Pre-flight checklist (Geoff,
2026-09-09) ... **each one cost a full fix round on chassis-B2**"). Workstation `CLAUDE.md`,
"Gate economy on a pass", fourth bullet.

**Measured.** "each one cost a full fix round on chassis-B2" — against the recorded rate of four
fix rounds across Tasks 4 to 8, and a fix round priced at the delta between a $8 and a $12 task cycle
(`STATUS.md`), plus one more full gate run.

**Quality safeguard kept.** It adds checks rather than removing them; the diff-reviewer still grades
every one of these independently.

**Precondition.** A repo whose fix rounds are dominated by report-shape and comment-hygiene defects,
which cairn established by counting them.

**Born from.** chassis-B2 Tasks 4 through 8, and chassis-A Tasks 11 and 12 (history-narration
comments; a friction-log rewrite that dropped `main`'s newer entries).

---

## 8. Fixed report schemas; the conductor reads records, never diffs or logs

**What.** The implementer returns a JSON object (filesTouched, gate, gateOutput, commits,
unspecifiedDecisions, couldNotDo, summary); the reviewer returns (verdict, summary, blocking[],
nonBlocking[], gate, unspecified). The conductor consumes only these. Workflow intermediate results
stay in script variables, never in the caller's context.

**Encoded at.** `pass-execute-chains.js:18-30` (IMPL_SCHEMA), `:32-65` (REVIEW_SCHEMA), `:107-108`
(the implementer report is handed to the reviewer as JSON), `:186-187` and `:232` (the conductor
sees only tallies). `diff-reviewer.md:52-65` (the exact report block, "nothing before or after it").
`~/.claude/docs/model-economy.md:41-56` and `:150-158`.

**Measured.** The rationale is metered rather than timed: "The 50% cap on the weekly pool is the
binding constraint, not API price. The design that follows from it **minimizes Fable's context size,
never the number of Fable turns**: a thin conductor that reads structured reports instead of diffs
spends less pool per turn" (`model-economy.md:66-72`). Cache-read compounding is the dominant term
for long conducting sessions (`:165-173`).

**Quality safeguard kept.** The diff-reviewer is explicitly the replacement for the conductor's own
diff read (`diff-reviewer.md` description; `model-economy.md:47-56`) and is told "Do not take a
claimed file list, claimed gate result, or claimed test coverage at face value" and "When the report
and the diff disagree, the diff wins and the disagreement is itself a finding"
(`diff-reviewer.md:20-25`, `:49-50`).

**Precondition.** A runner that enforces a response schema, and an Opus-pinned reviewer with repo
read access (cross-model diversity against correlated self-review blind spots).

**Born from.** The 2026-07-13 ecxc session that "silently burned ~1M tokens ... doing
execution-shaped work in the main loop", flagged by Geoff rather than the conductor
(`model-economy.md:82-93`).

---

## 9. Parallel chains in one worktree each, tasks sequential inside a chain

**What.** `parallel()` over chains, each chain a dedicated git worktree and branch with exactly one
writer; inside a chain, tasks run strictly sequentially and a non-accepted task halts the chain and
marks the rest `deferred` rather than burning them.

**Encoded at.** `pass-execute-chains.js:1-6` (header), `:69-70` ("You are the ONLY writer in this
worktree. Work only here, never in the main checkout"), `:190-204` (`runChain` halt-and-defer),
`:226` (`parallel(args.chains...)`). Worktree creation with the from-scratch showcase install and the
symlink proof: `cairn-overnight-to-release.js:201-208`. Workstation `CLAUDE.md`: "**Parallel chains
where Files are disjoint**, one worktree each, with any shared port made an environment variable".

**Measured.** The counter-evidence is what cairn recorded most precisely. Identity seam ran in
parallel with chassis-B1 and the entry notes: "**Two concurrent worktree gates on this machine ...
produce one-off flaky failures in `rendered.test.ts` and `reference-coverage.test.ts` from
shared-machine contention, not real regressions; re-run the single file in isolation before chasing
a fix**" (`docs/HISTORY.md`, Identity seam entry). The plan review put a hard ceiling on it: "two
parallel chains each running a full gate consumes the entire machine ... a run is 25-40 minutes ...
the admin sweep in this pass's own inputs **could not build at all** because 'a full install +
package + showcase build on a box at load 7.4 with two gating suites would contend'"
(`docs/internal/record/2026-09-08-polish-inputs/spec-review-risk.md:39`). So polish-11a deliberately
runs **one** chain: "No parallel chains: the single end-to-end slot, `CHANGELOG.md`, and
`docs/internal/engine-rulings.md` are all contended" (`plan-11a:52-55`).

**Quality safeguard kept.** One writer per worktree means the reviewer can assert "the task's diff is
exactly the commits the implementer reports ... the worktree has no other writers"
(`chains.js:105`). Each worktree proves it builds before any task runs (readlink check,
`overnight:201-208`).

**Precondition.** Genuinely disjoint file sets, a named contended-resource list (ledger files,
CHANGELOG, a preview port), machine headroom for concurrent gates, and a per-worktree from-scratch
install so the e2e proves the branch's engine rather than `main`'s symlinked build
(`plan-11a:71-74`).

**Born from.** The workstation-wide "One executor per worktree" ruling (2026-07-14, ~1.2M duplicated
tokens), plus the identity-seam/chassis-B1 flake pair and the admin sweep that could not build.

---

## 10. Halt, checkpoint and merge hygiene in the orchestrator

**What.** Halts PREPEND to STATUS rather than rewriting it, carrying the stage, the reason, and a
verbatim resume prompt. A ceiling check runs between runs. The merge step brings `main` in first with
fixed conflict rules.

**Encoded at.** `cairn-overnight-to-release.js:93-96` ("PREPEND one paragraph to the 'Immediate next
action' section (keep every existing bullet of that section intact) ... this resume prompt
verbatim"), `:215-218` (`ceilingCheck`, halts the stage when `budget.spent()` passes the stage's
`ceilingM`), `:193` (merge rules: "docs/STATUS.md takes origin/main's version whole
(conductor-owned); docs/HISTORY.md keeps both sides with this pass's entry first and no duplicates;
CHANGELOG.md keeps both sides' bullets under Unreleased"), `:194` (merge commit, then
`git pull --ff-only` on the main checkout). Workstation `CLAUDE.md`, "Gate economy on a pass",
Orchestrator hygiene bullet, which adds: "the Workflow tool refuses a `~/.claude/workflows`
scriptPath (copy to the session scratchpad)".

**Measured.** `STATUS.md` records the avoided cost directly: "**Do not push docs to `main` while a
PR is waiting on CI (it caused two extra merge rounds on #54)**". chassis-B2's halt-and-resume
"added a second cold start" mid-ritual and is named as one of the two drivers of its 2x overrun.

**Quality safeguard kept.** The resume prompt is committed, so a fresh session resumes without a
sitting: B2's halt "resumed from the committed resume prompt and completed capture-after, records,
simplify, verify, the reviewer fan-out, and the review-fix round **with no decision from Geoff**".

**Precondition.** A STATUS file that is conductor-owned and present-tense (the workstation ledger
rule), and a budget API the script can read mid-run.

**Born from.** PR #54's two extra merge rounds; B2's mid-ritual wall-clock halt.

---

## 11. The CI-canonical baseline rule (stop chasing phantom visual failures)

**What.** Visual baselines are CI-canonical. When the branch carries a CI regen commit, a local
`CI=1` e2e may fail on exactly the files that commit rewrote; that set, and only that set, is not a
red. Never regenerate those files locally.

**Encoded at.** `cairn-overnight-to-release.js:127` ("BASELINE RULE ... name the regen commit, paste
the failing snapshot names, confirm they equal the commit's file list, and treat the e2e as green").
Workstation `CLAUDE.md` carries it as one of the rulings binding every stage. Conductor-side CI regen
step: `overnight:102-110`.

**Measured.** No token figure; the recorded failure is unbounded: "A local `CI=1` e2e run drifts
permanently from a CI baseline regen by a fixed, nameable file list ... do not try to regenerate the
drift away locally, **since that only reintroduces it against the next real CI regen**"
(chassis-B2 entry). B2's own instance: "the showcase e2e's **20** `site-home-*`/`archive2-*` baseline
diffs verified, by file-list diff against `4de378ec`, as the pass's own already-documented
CI/workstation font-rendering drift rather than a regression".

**Quality safeguard kept.** The exemption is by explicit file list, proven by diff against the named
regen commit; any failure outside the list is a real red.

**Precondition.** A CI workflow that can be dispatched with `update_snapshots=true` and pulled back
fast-forward into the worktree (`overnight:105-107`).

**Born from.** chassis-B2 and chassis-B1's baseline rounds; B1 separately recorded the inverse trap
("A baseline diff under the suite's own per-pixel threshold is not proof nothing moved").

---

## 12. Ceilings derived from a measured per-task band, with checkpoints and an 80% rule

**What.** The token ceiling is arithmetic over a measured band rather than a guess: gate-bearing
units x the band's middle, plus conductor waits. Checkpoint STATUS every three to four tasks; at 80%
of the ceiling finish the task, write STATUS, and ask one combined question.

**Encoded at.** `2026-09-08-polish-11a-pass.md:33-48`: ceiling "6.5M, re-derived here and superseding
the spec's 4.5M ... fourteen gate-bearing units ... at the measured band's middle of about 450K is
6.3M, plus about 0.2M for the two conductor CI regen waits. The band is the risk lens's measurement
for this repo, **370K to 530K per task over four comparable passes (conventions 4a about 530K,
conformance 4b about 430K, chassis-A about 367K...)**". Checkpoints at `:47-51`.
`docs/specs/2026-09-08-polish-passes-design.md:234`, `:329-331`, `:479-480`, `:561`.
Workstation `CLAUDE.md`, "Conducting a pass" (ceiling + checkpoint interval, default four tasks).

**Measured.** The band above, and the honesty note that makes it usable: "**The repo's recorded
overrun history is four for four on the last comparable passes** (`docs/HISTORY.md:343-344`, `:386`,
`:448-449`, `:547`), so at 80 percent of the ceiling the conductor finishes the task, writes STATUS,
and asks one combined question." B2's reconstruction method is recorded too: sum each turn's
`input_tokens + cache_creation_input_tokens + output_tokens`, excluding `cache_read_input_tokens`,
and treat the result as an **upper bound**, not the number of record.

**Quality safeguard kept.** The ceiling triggers a question, not a silent scope cut.

**Precondition.** A HISTORY that records per-pass spend, which is why B1's entry flags "the earlier
tasks (1 through 4) ran before that measurement was wired in and are not counted here".

**Born from.** Four consecutive overruns, and B2 at roughly 2x.

---

## 13. Plan-time adversarial lens reviews, and splitting the pass by track

**What.** Before any dispatch, each plan is reviewed through named lenses (coverage, executor,
sequencing, plannability, fold, verify) whose findings are folded into the plan. The executor lens is
the one that pays for itself in clock.

**Encoded at.** `cairn-cms/docs/internal/record/2026-09-08-polish-inputs/` — `plan-11a-review-{coverage,
executor,sequencing,fold}.md`, `plan-11a-verify.md`, and the same set for 11b and C; the earlier
`spec-review-{charter,grounding,risk}.md` and their rev3/rev4 rounds.

**Measured.** Two concrete yields. (a) The executor lens produced §1 (8 to 12 minutes off every one
of ~19 gate runs) and framed the §2 e2e ask as a ruling for Geoff rather than a plan edit
(`plan-11a-review-executor.md:243-268`). (b) The risk lens split polish-A by track before it ran:
"Sixteen nominal tasks are honestly nineteen or twenty ... this repo's measured band of 370K to 530K
per task over five comparable passes funds thirteen to sixteen tasks at 7M on a lighter gate than
this one; six of polish-A's tasks are each over 0.7M on their own deliverable counts. **Cut by track,
not by position**" (`spec-review-rev3-risk.md:45`), producing 11a (engine code, 6.5M) and 11b
(admin surface, 7.0M + 5.0M).

**Quality safeguard kept.** Lenses are read-only reviewers; findings are folded with a per-item
taken/declined record (`plan-11a-review-fold.md`).

**Precondition.** Front-loaded interaction budget (workstation `CLAUDE.md`, "Process
proportionality": "Before plan approval, probe requirements and design until the plan carries no
open readings"), and a repo with recorded per-task spend to size against.

**Born from.** poplar pass 1b's three failure modes (`model-economy.md:95-143`): a grant is not
headroom; accretion by adjacency; **splitting tasks instead of splitting the pass** — "task-splitting
... looks like sizing discipline while changing nothing about the commitment".

---

## 14. The safeguard that pays for all of the above: the pass-end reviewer fan-out

**What.** After every task's own gate is green, domain reviewers grade the whole pass diff
(`svelte-reviewer`, `web-auth-security-reviewer`, `cloudflare-workers-reviewer`,
`daisyui-a11y-reviewer` when markup moved), plus a fresh-context `visual-verifier` for paint.

**Encoded at.** `cairn-overnight-to-release.js:170-183` (parallel reviewer dispatch, then one fix
round), `:140-168` (capture-after plus the verifier loop). `2026-09-08-polish-11a-pass.md:1636-1650`
(named per what the pass touches; `daisyui-a11y-reviewer` explicitly does not run when no task
changes markup). Workstation `CLAUDE.md`: "Domain reviewers still fan out at pass end."

**Measured.** chassis-B2: "the pass-end reviewer fan-out ... returned **four blocking findings none
of Task 3 through 6's own per-task `diff-reviewer` caught**, since each read as correct in isolation
and only surfaced once the whole pass's diff was graded together" — a missing `h1` and inverted
heading order on `/archive/2`, `App.PageData` never declaring `siteName` despite nine readers, the
root layout pulling the full render adapter and three eager content globs into every `/admin`,
`/members` and 404 request, and a silently regressed site description. Identity seam: the fan-out
returned 22 findings. chassis-A: four reviewers returned ship-it with no blocking finding.

**Quality safeguard kept.** It *is* the safeguard; it is what makes trimming per-task gates safe.
Its own lesson: "bank a pass's harvest and ROADMAP routing **only after** the fan-out and verifier
have run, or their findings need a second pass at close to get filed" (B2's harvest commit predated
the verifier by nearly three hours).

**Precondition.** Reviewer agents pinned to `claude-opus-5` in frontmatter — note the identity-seam
finding that "a per-dispatch `model` override on them is rejected, not silently ignored".

**Born from.** chassis-B2's four cross-cutting defects; identity seam's masked-primary-probe
regression.

---

## Failure-mode ledger (what each measure is a scar from)

| Failure mode | Recorded in | Measure it produced |
|---|---|---|
| Gate rebuilt an unchanged `dist` thirteen times per run | `plan-11a-review-executor.md:243-256` | §1 |
| Eight per-task e2e runs caught nothing unique, at a third of each task's hour | spec `:597-600`; CLAUDE.md | §2 |
| Four of five late tasks needed a fix round; pass hit 12.7M vs 6M | B2 post-mortem, Budgets | §3, §7, §12 |
| Implementers polled their own gate with no-op commands and tripped the runaway guard | chassis-A entry | §4 |
| CLI suite absent from the gate; a reformat broke a byte-pinned literal | chassis-A entry | §6 |
| Comment claims beyond the assertion, process citations in shipped comments, missing report blocks | implementer `:86-95` | §7 |
| ~1M tokens burned doing execution-shaped work in the main loop (ecxc, 2026-07-13) | `model-economy.md:82-93` | §8 |
| Two concurrent worktree gates flaked two test files; a sweep could not build at load 7.4 | Identity-seam entry; `spec-review-risk.md:39` | §9 (its limits) |
| Two executors raced one worktree, ~1.2M duplicated tokens (2026-07-14) | workstation CLAUDE.md | §9 |
| Docs pushed to `main` while a PR waited on CI: two extra merge rounds on #54 | `STATUS.md` | §10 |
| Halt mid-ritual added a second cold start | B2 post-mortem | §10 |
| Local e2e drifts permanently from CI baselines; regenerating locally reintroduces it | B2 entry | §11 |
| Four consecutive passes overran their ceiling | `plan-11a:44-46` | §12 |
| poplar 1b: grant read as headroom, accretion by adjacency, task splits instead of a pass split | `model-economy.md:95-143` | §13 |
| Per-task diff-reviewer structurally cannot see cross-cutting defects | B2 entry | §14 |
| Harvest banked before the verifier ran; findings needed a second close pass | B2 post-mortem | §14 |
| A baseline diff under the suite's per-pixel threshold is not proof nothing moved | B1 entry | §14 (verifier tile reads) |

## Things cairn measured that did NOT become a cut

- **No check removed for cost.** "the spec forecloses that" (`plan-11a-review-executor.md:239-241`).
  Everything above is a wrapper change, a scheduling change, or a move to CI.
- **Effort defaults stay at `medium`; never lowered to `low`** (workstation CLAUDE.md, "Conducting a
  pass": "Fable 5.1 at `low` answers from memory").
- **Reviewers stay Opus-pinned** for cross-model diversity, at $3.50 a task cycle, rather than being
  downshifted to Sonnet (`model-economy.md:15-20`; `STATUS.md` cost paragraph).
