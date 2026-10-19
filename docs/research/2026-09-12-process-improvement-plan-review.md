# Process improvement plan, revision 1: merged three-lens review and fold record

Three fresh-context adversarial reviews of
`docs/research/2026-09-12-process-improvement-plan.md` revision 1, merged
2026-09-12, with the disposition of every finding in revision 2. The three per-lens files
this replaces were `-review-evidence.md`, `-review-quality.md` and `-review-mechanics.md`.

**The lenses.**

- **EVIDENCE**, evidence fidelity: every number, attribution and study claim checked
  against the three cited evidence files and, where an evidence line cites a repo or
  workstation file, against that file.
- **QUALITY**, quality risk: of every item, what defect class it stops catching, and
  whether the named safeguard catches that class with a concrete scenario behind it.
  Quality is the immovable point, so a finding that an item lowers the bar is a blocker and
  its fix is a safeguard or a withdrawal.
- **MECHANICS**, implementability and cost: every encoding site the plan names, read; two
  claims checked by running the tool rather than by reading about it.

**Citation keys**, as revision 2 now keys them: `measure:N`, `harvest:N` and `research:N`
are lines in `2026-09-12-dubplate-economy-measure.md`,
`2026-09-12-cairn-economy-harvest.md` and `2026-09-12-pattern-soundness-research.md`, all
under `docs/research/`.

## Totals

| Severity | Findings | Folded as proposed | Adapted | Rejected |
| --- | --- | --- | --- | --- |
| Blocker | 5 | 3 | 2 | 0 |
| Major | 18 | 12 | 6 | 0 |
| Minor | 6 | 5 | 1 | 0 |
| **Total** | **29** | **20** | **9** | **0** |

By lens: EVIDENCE 12 (9 major, 3 minor), QUALITY 8 (4 blocker, 3 major, 1 minor),
MECHANICS 9 (1 blocker, 6 major, 2 minor).

"Adapted" means the finding was taken with a change to the reviewer's proposed fix, named
per row below. Nothing was rejected, and no finding was answered with a caveat.

## Owner-owed decisions

Six folds applied a reviewer's recommendation to a question Geoff owns. Each is marked at
its site in revision 2 as **Owner-owed** and listed in that document's header.

1. **Item 1, CI.** Stand up a CI job that runs the full gate on every push, so cairn's
   blast-radius precondition transfers intact, or accept the boundary-only recovery
   cadence? Applied as the boundary-only cadence. (QUALITY-5.)
2. **Item 1, a new gate leg.** The merge ritual's in-tree gate gains a branch-history
   `gitleaks` scan, which is a new leg in the repo `CLAUDE.md` "Gate" section. Applied as
   added. (QUALITY-1.)
3. **Items 3 and 6, shared workstation assets.** `~/.claude/workflows/pass-execute.js` and
   `~/.claude/agents/diff-reviewer.md` are shared with cairn. Applied as optional fields
   and an opt-in argument only, with a one-task dubplate smoke test before any cairn pass
   runs on the edited runner. (MECHANICS-2.)
4. **Item 4, the close's authorship.** It moves from a drafting implementer plus a folding
   agent to one fold agent. Applied with one independent `diff-reviewer` read over the
   fold's diff and the committed draft kept. (QUALITY-4, MECHANICS-7.)
5. **Item 7, the default three-lens set.** Applied as contract-and-criteria,
   mechanics-and-feasibility, and domain-risk, with staleness moved into the drafter's
   pre-flight. (QUALITY-7.)
6. **Item 14, the cross-lane review trigger.** Applied as a disjunction that adds the
   shared-package case and keeps the non-fast-forward case. (QUALITY-3.)

## Blockers

### QUALITY-1 (Item 1). The secret scan was scoped from a plan string, not from the task's diff

Revision 1 rested its coverage claim on the scoped scan covering exactly the paths a commit
changed, while encoding the scope as a per-task `gate` string. Both runners interpolate that
string at dispatch time, before the implementer has written anything
(`pass-execute.js:109`, `pass-execute-chains.js:45`), so the scope was the plan author's
prediction. A task declaring `Files: internal/auth/*.go` that also appends to
`env/dubplate-server.env.template`, which
`2026-09-03-parallel-build-approach.md:215-219` entitles a rung-5 task to do, would commit
an unscanned secrets manifest on a green gate. `--no-git` then means a later rewrite leaves
the secret in the branch history where no worktree scan can see it.

**Adapted.** The git-derived scope, the printed path list and the `-c` requirement are
folded as proposed, and they close the failure scenario, because a stray edit is in
`git status --porcelain` whatever the plan predicted. The reviewer's separate proposal to
keep `gitleaks` unscoped before every commit would empty the item of its whole saving, since
every in-chain run precedes a commit; in its place revision 2 adds a branch-history
`gitleaks detect` at the merge ritual's in-tree gate, which closes the history hole the
reviewer named and which a worktree scan structurally cannot. Item 1's expected effect,
priority, precondition and safeguard are all rewritten.

### QUALITY-2 (Items 2, 3, 6). The reduced gate was earned from a label, and its reduced set could not compile Go

`commentOnly` is set by the reviewer on a finding, which is a prediction about a fix not yet
written (`pass-execute-chains.js:69`), and the re-grade lands after the implementer has
gated and committed (`dubplate-implementer.md:65-69`). When a doc comment is right and the
code wrong, the correct fix to a comment-only finding is a code change, which is the shape
the window produced twice (`measure:153`, `measure:158`). The reduced set also omitted
`go build` and golangci-lint, while `//go:build` and `//go:embed` lines are comments to a
prose linter and load-bearing to the compiler.

**Folded as proposed.** Revision 2 has the implementer run a mechanical precondition on the
fix's own `git diff` before choosing the reduced string, with any failure falling back to
the full string, and the reviewer's label becomes the trigger to attempt the reduction
rather than its authority. `go build ./...` and golangci-lint join the reduced set.

### QUALITY-3 (Item 14). The trigger was replaced rather than extended, and the case it claimed to add cannot occur

A fast-forward rebase carries no other-lane commits
(`2026-09-03-parallel-build-approach.md:166`), so it cannot hide a shared-package overlap.
The added half was empty and the removal half was real: the ownership table at `:215`
records that a concurrent `.gitleaksignore` append is a non-fast-forward rebase and triggers
the re-review, which the swap would have dropped whenever package territories were disjoint.

**Folded as proposed.** The trigger becomes a disjunction, the expected token effect becomes
0 to plus 42,000, and the false sentence is replaced with the reasoning above.

### QUALITY-4 (Item 4). The single fold removed the close's independent read

Five of 5b-i Task 10's seven blockers are reachable by script or grep; two are not, namely
two doc-comment claims that did not match the code and a dropped "awaiting Geoff's read"
block (`measure:158`). A conductor-dispatched fold agent carries no `diff-reviewer` and
cannot commission one, and the HISTORY requirement that an entry carry what a later pass
would be wrong to rediscover is a judgment no measurement validates.

**Folded as proposed.** One independent read stays: the conductor dispatches one
`diff-reviewer` over the fold's diff, at Item 2's reduced gate. Its cost, one reviewer
dispatch at the measured mean 37,036 (`measure:234`), is carried in the item's re-derived
range.

### MECHANICS-1 (section 5, Items 1 and 2). The narrow forms are not deliverable

`scripts/check.sh` takes no arguments and its gitleaks leg is unconditional at `:128`; the
conform sweep is a package inside `make check` (`Makefile:35`, `:17-18`; `check.sh:57`)
rather than a leg to drop; and `dubplate-implementer.md:3, :65, :99` name
`bash scripts/check.sh` literally, so a dispatch-supplied gate string conflicts with the
standing definition the agent reads first. Two of section 5's seven entries were therefore
not adoptable.

**Adapted.** Revision 2 moves Items 1 and 2 out of launch day entirely and rewrites section
5 into four tiers, with tier D naming the one repo task both are blocked on. The reviewer's
alternative of building the flag was taken; the reduced gate is folded as a `check.sh
--reduced` mode rather than a hand-composed string, which also answers MECHANICS-6's
complaint that the string was never stated.

## Majors

### EVIDENCE-1 (sections 1 and 6). The ledger scale check was attributed to the wrong unit

`measure:18-22` checks the undivided `output + cache_creation`, twice the reported unit,
against the ledger. Section 6 told the next pass to compare its unit totals against a live
ledger reading, which is out by a factor of two in the direction that reads as a 50 percent
saving. **Folded as proposed** in both places, with the doubling stated explicitly.

### EVIDENCE-2 (Item 2). Four fix or escalate events where the cited line states six

`measure:148` reads six across fourteen tasks, and the plan's own section 2 agrees. The four
is `measure:83`, a different bucket counting fix-bearing tasks and excluding both closes.
**Folded as proposed:** six events, scaled to five or six at thirteen tasks, with the
denominator named.

### EVIDENCE-3 (Item 2). The clock range did not follow from its own derivation

Two to three rounds at 83s is 2.8 to 4.2 minutes, not the 4 to 11 claimed, and one close
task is 2.4 to 4.4 minutes rather than 6.8. **Adapted.** Revision 2 states 6 to 10 minutes
rather than the reviewer's 5 to 9, because the corrected six-event count from EVIDENCE-2
puts three to four events inside the reduction; the single-gate-run assumption is stated,
and the two-run alternative is priced at 11 to 15 and marked unmeasured.

### EVIDENCE-4 (Item 1). The 40-minute ceiling exceeded the total gate wall of the runs it scoped

25 runs at the 83s mean is 34.6 minutes in total, so deleting the gate entirely saves 34.6;
and 48.7m is the two-pass window total, 19.4m plus 29.3m (`measure:121`), presented as one
pass's ceiling. **Folded as proposed:** 24 to 30 minutes, the per-pass walls cited
separately, and the share sentence re-derived from 50% to 85% to 54% to 68%.

### EVIDENCE-5 (Item 5). Two cairn checklist entries were presented as fix rounds dubplate paid for

Four entries map onto `measure:150-158`; "no process citation in a shipped comment" and
"re-emit the labeled report block" are cairn's (`harvest:194-198`, `harvest:454`), and the
window's only comment-citation finding runs the other way (`measure:155`). **Folded as
proposed:** each entry now carries its source, and the framing sentence distinguishes the
four local rounds from the two adopted on shape.

### EVIDENCE-6 (Item 5). Both effect ranges contradicted the rows they cited

Neither 25,000 nor 60,000 is a bound of the stated bracket, and the wall contrast fails at
both ends: 5b-i Task 9 took a fix round at 16.6m (`measure:58`) and two clean tasks ran at
26.1m and 29.1m (`measure:51, :54`). **Adapted.** The token effect becomes the mean 36,000
with the single-round spread quoted, as proposed. The clock headline, which rested entirely
on the false contrast, is restated as about 3 minutes derived and up to 15 once the
unisolated round trips are counted, which the reviewer did not specify.

### EVIDENCE-7 (Item 7). The clock saving exceeded the entire wall of either measured prep

90 minutes exceeds both preps end to end, at 59 and 65 minutes (`measure:100-103`), and the
item keeps the drafter and all three lenses. **Folded as proposed:** 10 to 30 minutes, the
fold's serial tail, with the note that lens dispatches run in parallel.

### EVIDENCE-8 (Item 4). The duplication saving doubled the larger of its two halves

Half of 58,884 is 29,442 and half of 68,154 is 34,077, so the stated rule yields 29,000 to
34,000 rather than 30,000 to 60,000. **Folded as proposed**, and carried through the item's
sum, which then meets QUALITY-4's reviewer cost to give 47,000 to 52,000.

### EVIDENCE-9 (Item 8). The net range did not follow from the two figures it subtracted

Consistent pairing of 115,000 to 190,000 against a 17,000 to 52,000 add-back gives 98,000 to
138,000; the stated 105,000 to 155,000 is neither that nor the widest defensible pairing,
and 1.5 times 52,960 is 79,440 rather than 70,000. **Adapted, and superseded in part by
QUALITY-6**, which showed the measure removes two rounds rather than three. Revision 2
derives from the mean as the reviewer proposed, over the corrected count: 53,000 to 80,000
on a pass of 5b-i's shape, zero on a pass already running one round.

### QUALITY-5 (Item 1). cairn's blast-radius pattern was imported without the precondition that makes it safe

`harvest:183-184` states the precondition as a CI running the omitted suites on every push
with a merge policy blocking on red CI, and `harvest:186-188` records the case that needed
it. dubplate has no CI: `ls .github` returns no such directory, and
`2026-09-03-parallel-build-approach.md:165` names that as why the on-master gate exists at
all. **Folded as proposed:** the precondition difference is stated in Item 1, the CI
alternative is named and priced as the owner-owed question, and the compensating cadence is
the git-derived scope plus the boundary and branch-history scans.

### QUALITY-6 (Item 8). A four-times-larger diff was assumed to be read at twice the budget

The item assumed sub-linear attention outright while the document treats context degradation
as first order elsewhere (`research:61-70`), and the behavior-change rates it guards against
are per refactoring rather than per round (`research:234-242`). The reviewer also showed the
measure removes two rounds, not three: 8c already ran exactly one (`measure:43`) and 5b-i
ran three (`measure:65`). **Adapted.** The per-package review chunking is folded as
proposed and the count correction is taken as the stronger of the two lenses' arithmetic,
superseding EVIDENCE-9's three-round derivation; both corrections are recorded in the item.

### QUALITY-7 (Item 7). The named triad dropped the mechanics lens

The 5b-ii prep's actual lenses were mechanics, contract and security (`measure:104-106`),
and the proposed triad replaced mechanics with staleness, leaving no lens asking whether a
plan's mechanism can be built. 8c Task 2 escalated on exactly that class, costing 53,948 in
chain plus 27,524 out of chain and producing the window's one unreviewed commit
(`measure:155`). **Folded as proposed:** the triad becomes contract-and-criteria,
mechanics-and-feasibility, and domain-risk, with staleness moved into the drafter's
pre-flight and the raise-to-four alternative named.

### MECHANICS-2 (Items 3 and 6). `pass-execute.js` is a shared asset, and the port is not a mirror

A required field would be asked of `cairn-implementer.md`, which never mentions mutations,
and the chains runner's reduced-gate line hardcodes three npm commands dubplate does not
have (`pass-execute-chains.js:90-92`). **Folded as proposed:** every new field is optional,
the reduced gate arrives as an `a.reducedGate` argument defaulting to absent, and the
one-task smoke test before any cairn pass is stated as the owner-owed condition.

### MECHANICS-3 (Item 10). A Task 0 with no gate returns `needs-decision`

`pass-execute.js:140-146` returns `accepted` only on `verdict === "accept"` and
`gate === "pass"`; anything else falls through, lands in `tally()` at `:220-230` and prints
at `:274` as an unresolved task. One conductor ruling per lane launch erases the item's
3,740 saving. **Folded as proposed,** taking the reviewer's first option: Task 0's notes
tell the reviewer to report `gate: "pass"` with the conductor's baseline line quoted as its
evidence, which needs no code and reports a gate that genuinely ran. The `noGate: true`
alternative is named as the better form if Item 3's runner edit is being made anyway.

### MECHANICS-4 (citation keys, Item 4, section 6). Every input lived in a session-scoped `/tmp` directory

The three evidence files and the measurement scripts are keyed to this session's UUID;
`gtime.py:2` and `gates.py:2` hardcode the transcript path and `spend-ledger.sh` is a live
heartbeat rather than a per-pass emitter. **Adapted.** The three evidence files are copied
into `docs/research/` under dated names at revision 2 and the citation block is
re-keyed to them, so every pointer resolves; the generalized `tools/` emitter is folded as
Item 4's stated precondition and as section 6's named source rather than built here.

### MECHANICS-5 (Item 1). A scoped scan runs a different configuration than a full one

Verified by running the tool. Against the repo root, `using existing gitleaks config
.gitleaks.toml from (--source)/.gitleaks.toml`; against a subdirectory, `no gitleaks config
found in path .claude/.gitleaks.toml, using default gitleaks config` (gitleaks 8.30.1). A
scoped run silently drops `.gitleaks.toml:4-16`. `gitleaks detect` also takes one `-s` path,
so a multi-directory change set means one process per directory. **Folded as proposed:**
`-c "$REPO_ROOT/.gitleaks.toml"` is required on every scoped invocation, the multi-path
shape is specified, and the per-invocation residual is named in the derivation.

### MECHANICS-6 (Item 2). vale reads no Markdown, and the reduced gate string was never stated

`vale-comments.sh:3` and `:18` take comment text from `.go` files only, and no leg of
`check.sh` reads a `*.md` file, so the item's safeguard sentence was false and a docs-only
reduced gate is empty. The item named its encoding site without stating the string, and the
shared definition at `diff-reviewer.md:67-70` names three legs of which dubplate has one.
**Adapted.** The string is stated, as a `check.sh --reduced <packages>` mode rather than the
reviewer's free-standing command, so it cannot drift from the full gate's internals. The
false safeguard sentence is replaced with the true one: a Markdown-only commit is gated by
the reviewer and by the boundary's full gate and by nothing between, which is why the
boundary run is mandatory.

### MECHANICS-7 (Item 4, Item 11). The draft is the restart artifact, not duplication

`2026-09-12-rung-5b-ii-auth-handlers.md:2953-2957` keeps the ledger text as committed drafts
because they survive a restart, which is the property Item 11 depends on while making
restarts more frequent. **Folded as proposed,** taking the reviewer's second alternative:
the fold agent commits its draft before folding, so the artifact survives and the text is
still written once. The interaction is stated in Item 4.

## Minors

### EVIDENCE-10 (sections 2 and 6). "Eighteen closed passes" describes the ledger, not the sample

The 22 to 49 percent band rests on the seven recorded prep budgets at `research:36-39`; the
eighteen is the ledger's pass count at `research:5-6`. **Folded as proposed** in the prose
and in the section 6 table row.

### EVIDENCE-11 (multiple). Six citations pointed beside their claim

`research:154-156` not `:161-164` for reference-guided grading; `research:161-169` not
`:169-177` for the Xiang pass rates, in both places; `research:446` not `:444` for the
no-study bullet and `research:447-448` not `:445-446` for the severity gap;
`overnight-window.md:45-46` not `:47-49` for the spend ledger and `:163-166` not `:167-177`
for the simplifier ruling; `5b-ii:3185-3186, :3279-3284` not `:3285-3290` for the pass-end
reviewer. **Folded as proposed.** Every one is corrected, and the same-family bias pointer
at `research:171-184` is corrected with them. The finding is the document meeting its own
Item 4 rule.

### EVIDENCE-12 (sections 4, Items 6 and 8). Three study claims dropped a scope qualifier

Reference-guided grading is measured on math judging; the 11.5 percent study is recorded as
paywalled and SECONDARY, which section 1 promises to flag; MuTAP's 93.57 percent is on
synthetic buggy code against Pynguin. **Folded as proposed.** All three qualifiers are
restored.

### QUALITY-8 (Item 10). A gate read moves into the conductor that the thin-conductor rule forbids

`cairn-run-gate` prints the gate's tail, and the rule says no gate transcript during
execution (`~/.claude/CLAUDE.md`; `2026-09-03-parallel-build-approach.md:234`). **Folded as
proposed:** the exemption is stated rather than left implicit, scoped to the pass-or-fail
line at lane launch only, with a red baseline stopping the launch rather than opening a
diagnosis in the main loop.

### MECHANICS-8 (Item 11). The cold-start proxy is a lighter job than the thing it prices

A Task 0 dispatch is a subagent with a task-scoped read list; a fresh conductor session
reads both `CLAUDE.md` files, the memory index, STATUS in full, and a 3,344-line plan.
**Adapted.** Rather than the reviewer's two options, revision 2 states the 262,000 as an
upper bound rather than a midpoint, marks the cold start unmeasured, and names the
first-turn `cache_creation` as the next window's way to price it.

### MECHANICS-9 (section 6). Two metrics have no instrumentation and one has no owner

`cairn-run-gate` emits no elapsed time and the implementer report requires the gate's final
line rather than its wall, so three gate rows are reconstructable only from transcript
timestamps; and the fix-round cause classification, called the most important single metric,
had neither a dispatch nor a budget. **Folded as proposed:** one elapsed-seconds line in
`cairn-run-gate` and one report field are named as preconditions, and the classification
becomes a conductor-dispatched read at pass close with its own budget line in Item 4's
script.

## What all three lenses checked and found sound

Every figure in the plan's two baseline tables reproduces its cited line exactly, including
the ten window buckets, the window total of 2,415,715, and all six per-unit divisions to the
last digit (`measure:81-91, :95-96, :175-178, :229-234`). The fix-round cause table is
internally consistent against its base of 265,247, and the 42,836 escalate review is
correctly held outside that base. The gate figures hold: run counts of 13 and 22, walls of
19.4m and 29.3m, shares of 15.7 and 14.5 percent, the 83s mean over 35 runs, the `gitleaks`
range across about 119.6 MB, and the 125.6s conform figure. The prep, segmentation,
conductor and idle-gap figures check out, including the claim that one idle gap cost 1.96
times four hours of active conducting.

Section 1's characterization of the evidence base is honest: the cairn comparability warning
reproduces `harvest:368-370`, no cairn magnitude is imported into a dubplate forecast
anywhere in section 3, and the one place the plan could have leaned on cairn's 0.5M to 1.5M
lens saving it instead names the overstatement. All nine cairn attributions match their
harvest lines, and the remaining study characterizations are faithful to the research brief.

On quality risk, Item 5, Item 6's mutation half, Item 9, Item 11, Item 12 and Item 13 drew
no finding. Item 9 is the plan's one net addition of coverage and is correctly scoped; Item
11's handoff was checked for conductor knowledge living only in the conversation and the one
class expected to be missing, an unreviewed commit, is covered verbatim at
`overnight-window.md:518`. Section 4's nine retained elements each hold, the pass-end domain
fan-out most strongly (`harvest:424-430`).

On mechanics, the per-task gate override exists and reaches both agents; `commentOnly` is
present in the chains runner exactly where the plan says and absent from `pass-execute.js`;
the mutation requirement Item 6 promotes already stands in prose; Item 12's reading of
`pass-execute`'s missing mid-run hook matches the loop at `pass-execute.js:258-269`; and
Item 13's claim that 5b-ii already re-derives at the segment B boundary is accurate. Item
8's tension with the workstation Git Conventions rule is a standing stated deviation
recorded at `overnight-window.md:167-173`, widened rather than created, and revision 2
carries it as one line to Geoff at ruling time rather than as a finding.

## Revision 2.1 addendum

Revision 2.1 folds two inputs that arrived after this review was merged. Neither is a review
finding, so neither changes a disposition above; both are recorded here so the plan's
revision history stays readable from one file.

**Input 1: fresh cairn evidence and its dubplate counterpart.**
`docs/research/2026-09-12-cairn-fresh-evidence.md`, keyed `fresh:N` in the plan.
Its first half is Geoff's verbatim relay of the cairn session's own report on the run after
its runner change: a pre-runner task at 381 shell calls with up to 427K tokens of context per
call and another spending 263 of 562 calls polling a log, against a post-runner task at four
runner calls and zero polls, per-agent cost roughly halved across comparable stretches (1.73M
for 14 agents against 0.75M for 15), and every fix round of that run finishing in under four
minutes with the comment-only reduced gate and blast-radius scoping both in force
(`fresh:5-12`). Its second half is the dubplate counterpart measured the same night and
appended to the same file: every chain agent in the 2026-09-12 window already calls
`cairn-run-gate`, with a residue of up to fourteen `tail`-or-`sleep` calls in one agent and
118 shell calls in the heaviest task, agent `a4c457b6` in run `wf_c808ec5e-1ba`
(`fresh:18-31`).

Where it landed in the plan:

- Citation keys, as `fresh:N`, alongside a note that the file was added at revision 2.1.
- Section 1, as a paragraph placing each half on the evidence-weight scale: the relayed
  cairn half carries the existing comparability warning in full, the dubplate half carries
  the section 2 baseline's weight.
- Item 1's evidence paragraph, as the field result for blast-radius scoping.
- Item 2's evidence paragraph, as the field result for the comment-only reduced gate.
- **Item 15, new**: eliminate the residual gate-log polling through the implementer
  definition. Priority now, zero quality risk in either direction, expected effect small by
  construction (tens of agent turns per pass). The item names the wording that licenses the
  polling, step 5's "until it ends with" at `.claude/agents/dubplate-implementer.md:65-66`,
  and lands on the same three lines Item 1 already rewrites.

**Input 2: the segment-sizing read.**
`docs/research/2026-09-12-segment-sizing-research.md`, keyed `sizing:N`. Geoff
ruled its recommended rule on 2026-09-12 ("That works"): segments of three to four tasks,
every boundary on a green commit, the count overridden only for an irreversible task, a
second `fix` verdict, or a disjoint Files seam, with the handoff carried as structured repo
state.

Where it landed in the plan:

- Citation keys, as `sizing:N`.
- Section 1, as the note that the file is a derivation rather than a measurement.
- **Item 16, new and the document's only ruled item**: the rule, its derivation from the
  checkpoint-interval arithmetic (`H(S) = δ_eff / S + S / (2M)` minimized as
  `S* = sqrt(2 · M · δ_eff)`, `sizing:437-452`), and the 25% flatness result at a factor of
  two (`sizing:454-455`). The item cross-references section 2's segmentation-overhead
  baseline: a boundary at 65,000 to 85,000 (`measure:200-205`) against a mean code task of
  123,793 (`measure:175-176`) is a measured δ_eff of 0.53 to 0.69, higher than the read's
  assumed 0.4 to 0.5 and therefore moving `S*` up to 3.1 to 3.8 rather than down. It states
  plainly that M is the input the window does not settle, 2.3 counting every fix or escalate
  event and 14 counting real defects only, and routes the question to section 6's fix-cause
  classification. It also states its interaction with Item 12: Item 12's cut is measured
  against Item 16's count, and Item 16's count holds where they conflict.
- Section 6's metric table, which gains four rows: δ_eff, M under both readings,
  fault-detection latency, and the who-caught-it diagnostic.

**What did not change.** No revision 2 derivation was re-run, no item was withdrawn, and the
six owner-owed decisions listed in the plan's header are unchanged in number and in
disposition. Item 16 is ruled rather than owner-owed, so it adds nothing to that list. The
plan's item count moves from fourteen to sixteen.
