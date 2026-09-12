# dubplate development process improvement plan

Revision 2.1, 2026-09-12. Author: a research lane dispatched by the conductor; revision 2
folded a three-lens adversarial review, and revision 2.1 folds two inputs that arrived after
it. Status: proposal, unruled except Item 16, which Geoff ruled on 2026-09-12.

**What changed in revision 2.1.** Two inputs, neither of which existed when revision 2 was
written. First, fresh cairn numbers from the run after its runner change, relayed by Geoff
and now at `docs/superpowers/research/2026-09-12-cairn-fresh-evidence.md`, with dubplate's
own counterpart measured the same night appended to that file. They are cited where this
document discusses the blocking runner, the reduced gate, and blast-radius scoping, and they
produce one new item, Item 15. Second, the segment-sizing read at
`docs/superpowers/research/2026-09-12-segment-sizing-research.md`, whose recommended rule
Geoff ruled on 2026-09-12 ("That works"); it lands as Item 16, the document's only ruled
item. Nothing in revision 2 was withdrawn or re-derived, and no owner-owed question was
answered or added.

**What changed in revision 2.** Three fresh-context reviewers read revision 1 against
evidence fidelity, quality risk, and implementability. They returned 29 findings: 5
blockers, 18 majors, 6 minors. All 29 are folded here, 20 as the reviewer proposed and 9
adapted; none was rejected. The merged review, finding by finding with its disposition, is
`docs/superpowers/research/2026-09-12-process-improvement-plan-review.md`. Nine of revision
1's arithmetic derivations did not follow from the figures they cited and are re-derived
below, so most expected-effect ranges in section 3 have changed.

**Owner-owed decisions.** Six folds applied a reviewer's recommendation to a question Geoff
owns. Each is marked at its site as **Owner-owed** and listed here.

1. Item 1: stand up a CI job that runs the full gate on every push, so cairn's
   blast-radius precondition transfers intact, or accept the boundary-only recovery cadence?
   Applied as the boundary-only cadence.
2. Item 1: the merge ritual's in-tree gate gains a branch-history `gitleaks` scan, a new
   leg in the repo `CLAUDE.md` "Gate" section. Applied as added.
3. Items 3 and 6: `~/.claude/workflows/pass-execute.js` and
   `~/.claude/agents/diff-reviewer.md` are shared workstation assets cairn also runs.
   Applied as optional fields and an opt-in argument only, with a one-task dubplate smoke
   test before any cairn pass runs on the edited runner.
4. Item 4: the close's authorship moves to one fold agent. Applied with one independent
   `diff-reviewer` read over the fold's diff and the committed draft kept.
5. Item 7: the default three-lens set. Applied as contract-and-criteria,
   mechanics-and-feasibility, and domain-risk, with staleness moved into the drafter's
   pre-flight.
6. Item 14: the merge ritual's cross-lane review trigger. Applied as a disjunction that
   adds the shared-package case and keeps the non-fast-forward case.

Geoff's goal, verbatim: "minimize token spend and reduce development clock time while
holding the line on quality. Ultimately quality should always be the immovable point."

Quality is the constraint, not a term in the trade. Every item below either leaves
coverage identical, moves a check to a place that still catches the same defect, or adds
a check. No item removes a check to save money. Where an item's only argument is absence
of evidence, it says so and is marked reversible.

## Citation keys

Five evidence files, copied into the repo at revision 2 and 2.1 so every pointer below still
resolves after the session that produced them ends. All five are under
`docs/superpowers/research/`:

- `measure:N` is `2026-09-12-dubplate-economy-measure.md` line N.
- `harvest:N` is `2026-09-12-cairn-economy-harvest.md` line N.
- `research:N` is `2026-09-12-pattern-soundness-research.md` line N.
- `fresh:N` is `2026-09-12-cairn-fresh-evidence.md` line N (added at revision 2.1).
- `sizing:N` is `2026-09-12-segment-sizing-research.md` line N (added at revision 2.1).

Repo and workstation files are cited by path and line. Revision 1 keyed the three evidence
files to a session-scoped scratchpad under `/tmp`, which would have left every `measure:N`,
`harvest:N` and `research:N` in a committed document resolving to nothing.

## 1. The evidence base and its limits

Three bodies of evidence sit behind this plan, and they are not of equal weight. The
dubplate numbers are the strongest: they come from the 2026-09-12 two-lane window's own
workflow journals and agent transcripts, deduplicated on `message.id` before summing and
then reduced to a calibrated unit of `(output_tokens + cache_creation_input_tokens) / 2`
(`measure:12-16`). The quantity checked against the conductor's own spend ledger is the
undivided form, that is twice this unit, and it reproduces the ledger within about 15
percent across three segments (`measure:18-22`), so a comparison against a live ledger
reading doubles the unit first. They describe fourteen tasks in two
passes on one repo at one point in its ladder, so they generalize to the next dubplate
pass and to nothing else. The cairn-cms numbers are a second repo's record and are
**probably not comparable in magnitude**, because cairn's own reconstruction method sums
`input + cache_creation + output` per turn and treats the result as an upper bound
(`harvest:368-370`) with no statement that it deduplicates per message; if it does not,
its figures run two to three times high (`measure:22-23`), so cairn's numbers are used
below for the **shape** of a measure and its preconditions, never as a dubplate forecast.
The public studies are the weakest evidence for any specific number and the strongest
evidence about direction: several are explicitly marked SECONDARY in the research brief
because they were located through a search summary rather than verified against the
primary text (`research:3-4`), four of the questions this pattern turns on have no usable
public evidence at all (`research:442-453`), and no study anywhere compares per-task
against batched quality gates for coding agents (`research:446`). Where a claim rests on
absence of published evidence rather than a positive finding, the item says so and is
written to be reversible.

Revision 2.1 adds two files, and each sits in a different place on that scale. The fresh
cairn numbers (`fresh:5-12`) are a relayed quotation of another session's own report, not a
reconstruction this document can audit, so they carry the cairn comparability warning above
in full and are used only for direction and shape. The dubplate counterpart appended to the
same file (`fresh:18-31`) is a direct count of this repo's own transcripts and carries the
weight of the section 2 baseline. The segment-sizing read (`sizing:1-595`) is a derivation
rather than a measurement: its arithmetic is checkable and its two inputs, the boundary cost
and the per-task cost, are dubplate's own measured figures, but the fault rate it runs
against is an estimated band and the file says so (`sizing:461-478`).

## 2. Baseline: dubplate's measured pass economics

All token figures are unit-tokens as defined above. The window is 09:26 to 13:20 UTC on
2026-09-12, covering pass 8c (4 tasks plus a staleness read, main tree) and pass 5b-i (10
tasks plus a staleness read, worktree).

| Measure | Value | Source |
| --- | --- | --- |
| Window total | 2,415,715 | `measure:91` |
| In-chain implementers, 12 code tasks | 892,715 (36.9%) | `measure:81` |
| In-chain reviewers, first round, 12 code tasks | 366,721 (15.2%) | `measure:82` |
| Review of any kind (first round, fix rounds, merged range) | 634,507 (26.3%) | `measure:95` |
| Fix rounds, implementer plus reviewer, 4 tasks | 198,557 (8.2%) | `measure:83` |
| Out-of-chain fix, 8c Task 2 | 27,524 (1.1%) | `measure:84` |
| Task 0 staleness reads, 2 at opus | 80,211 (3.3%) | `measure:85` |
| Close tasks 8c-4 and 5b-i-10, in chain | 302,749 (12.5%) | `measure:86` |
| Ledger folds, 2 conductor-dispatched opus | 138,568 (5.7%) | `measure:87` |
| All close-and-ledger work | 441,317 (18.3%) | `measure:96` |
| Simplifier rounds, 4 | 211,840 (8.8%) | `measure:88` |
| Merged-range diff review, 1 | 41,936 (1.7%) | `measure:89` |
| Conductor's own turns, 105 turns | 154,894 (6.4%) | `measure:90` |

**Per-task cost by model and kind.**

| Unit | Count | Chain tokens | Per unit | Source |
| --- | --- | --- | --- | --- |
| Code task (8c 1-3, 5b-i 1-9) | 12 | 1,485,516 | 123,793 | `measure:175-176` |
| Close task, docs-only (8c-4, 5b-i-10) | 2 | 302,749 | 151,374 | `measure:176-178` |
| Sonnet implementer dispatches | 16 | 863,288 | 53,956 | `measure:229` |
| Opus implementer dispatches, upshifted | 4 | 332,408 | 83,102 | `measure:230` |
| `claude-opus-5` in-chain diff-reviewer | 16 | 592,571 | 37,036 | `measure:231, :234` |
| Opus out-of-chain (Task 0, simplifier, folds, merged review) | 9 | 472,555 | 52,506 | `measure:232` |

A docs-only close task costs 22 percent more than the average code task (`measure:178`),
before the ledger fold each one still needed (`measure:179-181`).

**Fix rounds by cause.** Six fix or escalate events across fourteen tasks. One was a
genuine code defect.

| Cause | Tokens | Share of fix spend | Source |
| --- | --- | --- | --- |
| Real code defect (5b-i Task 9, epoch keyed on the raw address) | 47,684 | 18% | `measure:153, :161` |
| Vacuous or missing test evidence (half of 5b-i Task 3) | about 29,000 | 11% | `measure:154, :162` |
| Comment-only, citation-pointer, and draft-content | about 189,000 | about 71% | `measure:163-165` |
| Plus the escalating review on the 5b-i close | 42,836 | (review only) | `measure:158, :164` |

Off-by-one `file:line` pointers into a 5,000-line `docs/HISTORY.md` and a 1,300-line spec
are their own sub-class: they caused a full reviewer round on 8c Task 3 and the second
round on 8c Task 4 (`measure:166-169`). Neither close's drafts were accepted as written
(`measure:192`).

**Gate wall.**

| | 8c | 5b-i | Source |
| --- | --- | --- | --- |
| Full `check.sh` invocations in chain | 13 | 22 | `measure:120` |
| Measured gate wall | 19.4m | 29.3m (+7.1m merge ritual) | `measure:121` |
| Chain wall, first impl start to last review end | 123.8m | 201.6m | `measure:122` |
| Gate share of chain wall | 15.7% | 14.5% (18.0% with merge) | `measure:123` |

Per-task gate share ranges 7% to 32%, median about 18%, at 2.5 full gate runs per task
(`measure:125-126`). Inside a run, `gitleaks` over the tracked worktree takes 56.7s to
2m13s scanning about 119.6 MB and is the single dominant leg in most runs
(`measure:131-132`); `internal/harness/conform` takes 125.6s whenever the Go test cache
misses it (`measure:133`); everything else takes seconds (`measure:134-135`). The mean
full run is therefore 83s (48.7m over 35 runs, `measure:121, :126`), and `gitleaks` alone
is 69% to 160% of that mean.

Two gate charges are pure waste against the current shape. A docs-only close task pays
the full gate: 8c Task 4 ran it three times for 264s on two Markdown drafts and a spec
erratum, and 5b-i Task 10 ran it twice for about 146s, about seven minutes of gate on two
commits that touched no code (`measure:139-143`). A read-only staleness probe paid 205s
of gate plus `go test ./...` purely to record a baseline (`measure:144`), which is the
32% gate share on 5b-i Task 0 (`measure:49`), while 8c's Task 0 did the same job with
zero gate runs (`measure:37`).

**Plan-prep share.** Prep is the largest single line item and it sits outside the window
total above. The pre-window plan and standard prep ran about 2.0M across 22 subagents
plus conductor turns, of which one agent, "Fold seven reviews into standard rev 2," cost
564,846 in 30 turns (`measure:100-102`). The 5b-ii plan prep ran 830,069 across 7 agents:
draft 230,218, three parallel plan reviewers 286,376, fold 152,846, the 5b-i security read
82,851, spec errata 77,778 (`measure:103-106`). One plan revision cost more than pass 8c's
entire code execution, which was 454,274 for Tasks 1 through 3 including implementers,
reviewers and fixes (`measure:107-108`). The same shape shows across the seven passes whose prep
budgets `docs/HISTORY.md` records, out of roughly eighteen closed passes: prep runs 22% to
49% of a pass and twice exceeded the whole implementer chain (`research:36-40` for the
band, `research:5-6` for the ledger's pass count).

**Conductor share.** 105 turns for 154,894, or 1,475 per turn (`measure:216-217`). The
thin-conductor discipline held: the conductor read no diffs and no gate logs, and its
share of the window was 6.4% (`measure:218-219`).

**Segmentation overhead.** Five `pass-execute` launches, four boundaries. A boundary
costs a `code-simplifier` round at 38,364 to 63,160 (mean 52,960) plus 10,000 to 31,000
of conductor turns to close, checkpoint STATUS, and relaunch, for about 65,000 to 85,000
per boundary (`measure:200-205`). Four boundaries is about 272,000, roughly 11% of the
window (`measure:207`). The marginal cost of segmenting into five runs instead of two is
the two extra simplifier rounds, about 100,000, plus the relaunch turns
(`measure:207-209`). There is no within-segment cache sharing to lose, because
`pass-execute` spawns a fresh agent per task and every implementer re-reads the plan
regardless of which run it sits in (`measure:211-214`).

**The idle-gap re-prime.** One conductor turn block at 17:33 UTC, after the roughly
four-hour idle gap between the 5b-i close and the prep phase, carries
`cache_creation = 574,539` in a single request, or 303,912 unit across 4 turns
(`measure:109-113`). That is 76,000 per turn against 1,475 per turn for the active
window (`measure:216-222`). One idle gap cost twice what four hours of active conducting
cost (`measure:222-223`).

## 3. The improvement items

Sixteen items. Each states its measure, its evidence, its expected effect as a range
with the derivation, the safeguard that keeps the bar, its precondition, where it is
encoded, its owner, and a priority. Items 1 through 14 are revision 2's; Items 15 and 16
arrived with revision 2.1, and Item 16 is the only one Geoff has ruled.

---

### Item 1. Scope `gitleaks` to the change set the gate derives for itself, full-tree at every boundary

**The measure.** `scripts/check.sh:128` runs `gitleaks detect --no-git -s .` over the
whole tracked worktree on every invocation, including a docs-only commit and a read-only
probe. Give `check.sh` a `--scope-changed` mode that derives the scanned set itself, from
`git diff --name-only` against the task's base plus `git status --porcelain` for untracked
files, prints the derived list, and scopes the gitleaks leg to it. The scope is never
supplied by the plan. Keep the full-tree scan at every segment boundary, at the merge
ritual's in-tree and on-master gates, and at pass end. The scan is not removed and no path
escapes it; the schedule changes, not the coverage.

**The scope comes from git, never from the plan.** Revision 1 encoded the scope as a
per-task `gate` string carrying the task's `Files` list. Both runners interpolate that
string at dispatch time, before the implementer has written anything
(`~/.claude/workflows/pass-execute.js:109`;
`~/.claude/workflows/pass-execute-chains.js:45`), so the scope would have been the plan
author's prediction of the diff rather than the diff itself. The two differ in practice.
`.claude/agents/dubplate-implementer.md:71-77` instructs the implementer to verify plan
assumptions against the real code, and
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:215-219` names contested
surfaces a rung-5 task amends without declaring them, `env/dubplate-server.env.template`
among them, which the repo `CLAUDE.md` records as a secrets manifest. A task declaring
`Files: internal/auth/*.go` that also appends a commented pending row to that template
would have committed it unscanned, and the implementer commits on a green gate
(`.claude/agents/dubplate-implementer.md:65-68`). A git-derived scope covers that file,
because it is in `git status --porcelain` whatever the plan predicted.

**Two mechanical requirements on the scoped form.** `gitleaks` resolves its config
relative to its source path. Run against the repo root it reports `using existing gitleaks
config .gitleaks.toml from (--source)/.gitleaks.toml`; run against a subdirectory of the
same repo it reports `no gitleaks config found in path .claude/.gitleaks.toml, using
default gitleaks config`. Both lines are from `gitleaks detect --no-git --log-level debug`
on 8.30.1, the installed version. A scoped run therefore silently drops
`.gitleaks.toml:4-16`, which sets `[extend] useDefault = true` and the `.claude/worktrees/`
allowlist, unless the mode passes `-c "$REPO_ROOT/.gitleaks.toml"` on every invocation. It
must. And `gitleaks detect` takes one `-s, --source` path per its own `--help`, so a
multi-directory change set means one process per directory, each paying rule-compile and
walk startup. The mode reduces the derived paths to their common directories and loops, and
the derivation below carries that residual.

**The history hole, and the scan that closes it.** `check.sh:128` runs with `--no-git`, so
it scans the worktree and not the branch's history. A secret committed by one task and
rewritten by a later one is invisible to every later worktree scan, boundary scans
included, while it stays in the branch permanently. The full-tree boundary scan does not
recover it. The merge ritual's in-tree gate therefore gains one `gitleaks detect` over the
branch's commit range, without `--no-git`, before the work leaves the branch. That is an
addition, and it is the only place in this document where a leg joins a gate.

**Owner-owed:** the merge ritual's in-tree gate gains a branch-history `gitleaks` scan, a
new leg in the repo `CLAUDE.md` "Gate" section, applied as added.

The same conditional applies to three ops legs that cannot change verdict on a task that
touched none of their inputs: `docker compose config` (`check.sh:137-144`),
`systemd-analyze verify` (`check.sh:146-169`), and the state-dir agreement leg
(`check.sh:171-186`). Those three take seconds (`measure:134-135`), so they are a
tidiness change rather than a clock change, and they are listed here only because the
same flag carries them.

The conform sweep needs no scoping rule. It is already blast-radius-scoped by the Go test
cache and costs 125.6s only when that cache misses (`measure:133`), which is a real input
change or a worktree's first run. The actionable half is that a fresh worktree pays it
once at lane launch, not on every task.

**Evidence.** `gitleaks` is the dominant leg at 56.7s to 2m13s per run against a mean run
of 83s (`measure:131-132`, `measure:121, :126`). Both dominant legs are
blast-radius-insensitive: the corpus under `corpus/cases/` does not change, and
`internal/harness/conform` was untouched by anything either pass wrote, yet a docs-only
close task pays both (`measure:136-139`). The workstation's own ruling already carries
this shape for browser suites ("the slow suite runs only where it can catch something,"
`~/.claude/CLAUDE.md`, "Gate economy on a pass," first bullet), and cairn's third bullet
extends it to test suites by blast radius (`harvest:165-188`). cairn has since run a pass
under that scoping and reported the clock result: with blast-radius scoping and the
comment-only reduced gate both in force, "every fix round tonight ran in under four minutes"
(`fresh:11-12`). That is one session's relayed report on another repo, so it is evidence
that the scoping holds its bar in practice rather than a dubplate forecast. The research
brief reaches
the same conclusion independently and names its own weakness: no study compares per-task
against batched gates for coding agents, so this rests on absence of evidence and should
be reversible (`research:115-137`, `research:446`).

**The precondition cairn's pattern carries and dubplate does not.** cairn's blast-radius
bullet states its own precondition at `harvest:183-184`: "A CI that runs the omitted suites
on every push, and a merge policy that blocks on red CI." Its safeguard sentence is "the
PR's own CI run is the gate that clears them. Nothing merges on a red CI"
(`harvest:179-180`), and the same section records the case that needed it, a mechanical
Prettier reformat that reflowed a pinned literal neither the plan's gate nor the dispatched
gate covered, "the reason the CLI suite is now a per-task exception" (`harvest:186-188`).
dubplate has no CI. `ls .github` returns no such directory, and
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:165` states the consequence:
"nothing in this repo is CI-enforced, so 'master is never left red' is otherwise
aspirational." In cairn the omitted runs are recovered on every push. In dubplate they are
recovered at a boundary that may sit several committed tasks later, which is why the
git-derived scope and the branch-history scan above are the conditions under which this
item is safe at all rather than refinements of it.

**Owner-owed:** stand up a CI job that runs the full gate on every push, so cairn's
precondition transfers intact, or accept the boundary-only recovery cadence? Applied as the
boundary-only cadence, with the git-derived scope and the branch-history scan as its
compensation.

**Expected token effect.** Near zero, and not separately measurable. Gate wall carries no
token line in the measurement.

**Expected clock effect: 24 to 30 minutes per pass of 5b-ii's size.** Derivation: 13 tasks
at the measured 2.5 full runs per task (`measure:126`) is about 32 runs. Five segment-end
runs plus the merge ritual's two leave about 25 scoped runs. At `gitleaks`' measured floor
of 56.7s (`measure:132`) those 25 runs carry 24 minutes of scannable wall. The ceiling is
arithmetic rather than the 2m13s upper leg figure: 25 runs at the mean full run of 83s
(`measure:121, :126`) is 34.6 minutes of gate wall in total, and the legs that remain take
seconds each plus the conform sweep on a cache miss (`measure:133-135`), so about 30
minutes is the most the scoping can remove, less the per-invocation startup the scoped form
pays. Revision 1 stated 40 minutes and capped it with 48.7m, which is the two-pass window
total, 19.4m for 8c plus 29.3m for 5b-i (`measure:121`); neither measured pass had a gate
wall above 29.3m, and deleting the gate entirely from 25 runs saves 34.6. As a share, 24 to
30 minutes against a projected 32 runs at 83s, which is 44.3 minutes, removes 54% to 68% of
in-chain gate wall, itself 15.7% and 14.5% of chain wall today (`measure:123`).

**Quality safeguard.** Coverage is identical per commit, because the scoped scan is derived
from the working tree's own change set rather than from a declaration, and it runs under
the repo's config. The full-tree scan still runs at every boundary, at both merge gates,
and at pass end, and the branch-history scan at the merge gate closes the one hole a
worktree scan structurally cannot see. This is also the answer to `CLAUDE.md`'s "a check
either gates or does not exist; never guard one with `command -v || skip`": that rule
forbids a check that silently disappears when a tool is absent, and a declared scope mode
that prints what it scanned, with a mandatory full run at each boundary, is the opposite of
a silent disappearance.

**Precondition.** `scripts/check.sh` has no scope flag, and the 2026-09-12 window recorded
that it was deliberately not built ("`check.sh` has no fast/full scope flag ... the scope
flag is **not built this window** and every task pays the full three legs,"
`docs/superpowers/plans/2026-09-12-overnight-window.md:72-77`). The mode must exist, must
default to full, must derive its own paths, must pass `-c`, and must name the scoped paths
in its own output so the reviewer can see which form ran. It also needs three edits to
`.claude/agents/dubplate-implementer.md`, whose frontmatter at `:3`, step 5 at `:65` and
report format at `:99` all name `bash scripts/check.sh` literally. A dispatch-supplied gate
string that says anything else puts the prompt in conflict with three lines of the standing
definition, and the definition is what the agent reads first. Those three lines become "the
gate string the dispatch gives you."

**Where it is encoded.** `scripts/check.sh` gains a `--scope-changed` mode that scopes the
gitleaks leg and the three ops conditionals, and the merge ritual's in-tree gate gains the
branch-history scan. `.claude/agents/dubplate-implementer.md:3, :65, :99`. The repo
`CLAUDE.md` "Gate" section records the forms and when each runs.

**Owner.** Tool change (`check.sh` and the implementer definition), then a plan-template
rule. The scope mode's design is a conductor decision; the `CLAUDE.md` edits are Geoff's to
ratify.

**Priority: next pass, in full. Nothing lands at 5b-ii's launch.** Revision 1 proposed a
narrow form for launch day: docs-only and read-only tasks running a hand-composed gate
string that omits the gitleaks and conform legs, "deliverable through the existing per-task
`gate` override with no code change at all." It is not deliverable. `scripts/check.sh`
takes no arguments and its gitleaks leg is unconditional at `:128`. The conform sweep is
not a leg of `check.sh` but a package inside `go test ./...`, reached through `make check`
at `Makefile:35` and `Makefile:17-18` and called as one unit at `check.sh:57`, so omitting
it means enumerating packages rather than dropping a leg. And the implementer definition
overrides a conflicting gate string, as the precondition above records. A hand-composed
string would duplicate `check.sh`'s internals and drift from them. The mode is built first,
as one repo task of roughly 40 to 60 lines of shellcheck-clean shell.

---

### Item 2. A reduced gate for comment-only fix rounds, docs-only fix rounds, and docs-only tasks

**The measure.** A fix round whose blocking findings are all comment-only, and whose own
fix diff proves it, runs a reduced gate rather than the full string. A task whose Files
list contains no code path runs the same reduced gate on its first attempt, not only on its
fix rounds.

**The reduced gate is a `check.sh` mode, and it compiles Go.** Revision 1 defined the
reduced set as "the comment and prose legs plus the touched packages' tests" and named vale
as "the only leg that can fail on Markdown and Go comments." Both halves are wrong. vale
reads no Markdown in this repo: `scripts/vale-comments.sh:3` describes its input as comment
text from each `.go` file and `:18` unions tracked and untracked-but-not-ignored `.go`
files, and no leg of `check.sh` reads a `*.md` file at all. And several Go comments are
directives. A `//go:build` line and a `//go:embed` line are comments to a prose linter and
load-bearing to the compiler, and `web/`'s assets ship through `go:embed` under the repo
`CLAUDE.md` stack rules. `make check` also covers golangci-lint, whose comment linters a
comment-only edit can red (`check.sh:3-4`). The reduced gate is therefore
`check.sh --reduced <packages>`: `go build ./...`, golangci-lint,
`scripts/vale-comments.sh`, and `go test` over the named packages. It is a mode of the same
script for the same reason the scope mode is, so it cannot drift from the full gate's
internals, and it is passed as an absolute string per
`docs/superpowers/plans/2026-09-12-overnight-window.md:170-177`. Revision 1 named the
reduced gate's encoding site without ever stating the string, and the string it would have
mirrored is cairn's, `npm run check:comments && npm run check:symbols && npm run
check:docs` (`~/.claude/workflows/pass-execute-chains.js:90-92`), of which dubplate has
none.

**The reduction is earned from the fix diff, never from the label.** `commentOnly` is set
by the reviewer on a finding, which is a prediction about a fix the implementer has not yet
written (`~/.claude/workflows/pass-execute-chains.js:69`), and the reviewer's re-grade lands
after the implementer has gated and committed (`.claude/agents/dubplate-implementer.md:65-69`,
steps 5 through 7). The window shows the shape that breaks it. 5b-i Task 9's real defect
was `CheckSession` and `CheckLink` keying the epoch lookup on the raw address rather than
the normalized one (`measure:153`), and 5b-i Task 10's escalate included "two doc-comment
claims that did not match the code" (`measure:158`). When the comment is right and the code
wrong, the correct fix to a comment-only finding is a code change. So the implementer runs
a mechanical precondition before choosing the reduced string: the fix's `git diff` must
touch no non-comment line in any `*.go` file, must add or remove no `//go:` directive line,
and must touch no file outside Markdown and comment text. Any failure falls back to the
full string. The reviewer's label is the trigger to attempt the reduction, never its
authority.

**Evidence.** Comment-only, citation-pointer and draft-content findings are about 71% of
all fix-round tokens plus the 42,836 escalating review on the 5b-i close, and none of them
changed a line of shipped Go (`measure:163-165`). The two docs-only close tasks ran five
full gates for about 410s between them on commits that touched no code
(`measure:139-143`). cairn encoded exactly this rule after its own fix-round
concentration and its 12.7M-against-6M overrun (`harvest:82-107`), and the workstation
carries it as the second bullet of "Gate economy on a pass." cairn's first pass run under
the encoded rule reports the reduced gate holding alongside blast-radius scoping, with every
fix round of that run finishing in under four minutes (`fresh:11-12`), which is the
strongest available evidence that the reduction costs no round trips back.

**Expected token effect: 0 to 20,000 per pass.** A reduced gate saves gate wall rather
than tokens. The small token effect is the shorter gate tail an implementer pastes into
its report, which the reviewer then reads; it is not separately measured and should not
be claimed.

**Expected clock effect: 6 to 10 minutes per pass.** Derivation, corrected. The close half:
the two docs-only close tasks' five full runs cost 264s plus about 146s
(`measure:141-143`), and 5b-ii carries one close task, so 2.4 to 4.4 minutes. The
fix-round half: the window ran six fix or escalate events across fourteen tasks
(`measure:148`), which at that rate is five to six events on a thirteen-task pass, and the
comment-only and draft-content class is about 71% of fix-round tokens (`measure:163-165`),
which puts three to four of those events inside the reduction. At the mean full run of 83s
each (`measure:121, :126`), against a reduced gate of a few seconds, that is 4.2 to 5.5
minutes. Sum, 6 to 10 minutes. The derivation assumes one gate run per fix round; at the
window's measured 2.5 full runs per task (`measure:126`) a round that re-runs the gate
twice puts the fix half at 8 to 11 minutes and the total at 11 to 15, and that assumption
is not measured. Revision 1 stated 8 to 18 minutes from a four-event count its own cited
line gives as six (`measure:148`; the four is `measure:83`, which counts fix-bearing tasks
and excludes both closes), and from a two-to-three-round arithmetic that reaches 2.8 to 4.2
minutes rather than the 4 to 11 it claimed.

**Quality safeguard.** Two things stand behind the reduction. The implementer's mechanical
precondition on the fix diff is the first, and it is the one that cannot be wrong about
what the fix did, because it reads the fix. The reduced set is the second: a comment edit
that cannot compile, cannot lint, or breaks its package's tests fails where it is made. The
full gate still runs at the segment boundary immediately after, so a round that slips past
both is caught within one segment rather than at pass end. For a Markdown-only task the
reduced gate is genuinely empty, because nothing in `check.sh` reads a `*.md` file; such a
commit is gated by the reviewer and by the full gate at the next boundary and by nothing in
between, which is why the boundary run is mandatory rather than nominal.

**Precondition.** Two. `check.sh` must carry the `--reduced` mode, which is the same repo
task that builds Item 1's scope mode and edits the three lines of
`.claude/agents/dubplate-implementer.md`. And the reviewer's output shape must carry the
per-finding flag with `pass-execute.js` routing on it: `commentOnly` is present in
`pass-execute-chains.js:45, :90-92, :103` and absent from `pass-execute.js:50-82`, and
dubplate runs `pass-execute.js`. That gap is Item 3.

**Where it is encoded.** `scripts/check.sh`'s `--reduced` mode;
`~/.claude/workflows/pass-execute.js` REVIEW_SCHEMA and `implementPrompt`; the mechanical
precondition in `.claude/agents/dubplate-implementer.md`; the repo `CLAUDE.md` "Gate"
section, which records the reduced form and when it runs; the docs-only task
classification in the plan template's per-task `gate` field.

**Owner.** Tool change, then a plan-template rule.

**Priority: next pass.** Revision 1 put this at "now" on the strength of a per-task gate
string that omits legs, which `check.sh` cannot express. It lands with Item 1's repo task.

---

### Item 3. Port `commentOnly` from `pass-execute-chains.js` into `pass-execute.js`

**The measure.** Add `commentOnly` to `pass-execute.js`'s REVIEW_SCHEMA blocking item
shape as an **optional** property, add the reviewer instruction that tells it to mark the
flag honestly, and add an `a.reducedGate` runner argument, defaulting to absent, carrying
the reduced gate string the implementer is given when every blocking finding holds the flag
and the fix diff clears Item 2's mechanical precondition.

**`pass-execute.js` is a shared workstation asset, and the port is not a mirror.** The
workstation `CLAUDE.md` names it as the runner every repo uses at six or more tasks, and
the script's own invocation example at `:12` carries `implementer: "cairn-implementer"`.
Two consequences. First, nothing this document adds may join either schema's `required`
list: IMPL_SCHEMA's `required` at `:47` and REVIEW_SCHEMA's are what every consumer must
satisfy, and `~/.claude/agents/cairn-implementer.md` never mentions mutations, so a
required field would ask a consumer for something its definition does not define. Every
field added here is optional. Second, the reduced-gate line cannot be copied across:
`pass-execute-chains.js:90-92` hardcodes `npm run check:comments && npm run check:symbols
&& npm run check:docs` plus the touched files' unit tests, and dubplate has no npm and none
of those three commands. The reduced string arrives as the runner argument instead, so a
consumer that passes none is unaffected.

**Owner-owed:** `~/.claude/workflows/pass-execute.js` and `~/.claude/agents/diff-reviewer.md`
are shared workstation assets cairn also runs. Applied as optional fields and an opt-in
argument only, with a one-task dubplate smoke test before any cairn pass runs on the edited
runner.

**Evidence.** The mechanism exists and is proven in the chains runner
(`~/.claude/workflows/pass-execute-chains.js:45, :90-92, :103`) and in the reviewer
definition (`~/.claude/agents/diff-reviewer.md:60, :67-70`). dubplate's own passes run
`pass-execute.js`, whose REVIEW_SCHEMA at `:50-82` has no such field, so the reviewer's
`[comment-only]` tag reaches the conductor as prose in SUMMARY and routes nothing. That
is why the window's comment-only fix rounds each re-ran the whole gate
(`measure:154-157`).

**Expected token effect: 0.** It is an enabler.

**Expected clock effect: 0 on its own.** It carries Item 2's 6 to 10 minutes.

**Quality safeguard.** The flag only ever reduces a gate for a round the reviewer itself
declared behavior-neutral and the implementer's own diff check confirmed, and the reviewer
re-grades that declaration on the next round.

**Precondition.** None. The reviewer definition already emits the tag.

**Where it is encoded.** `~/.claude/workflows/pass-execute.js`, REVIEW_SCHEMA at `:50-82`,
`validateArgs` for the new argument, and `implementPrompt` at `:102-120`, taking
`pass-execute-chains.js:45` and `:88-93` as the shape rather than as text to copy.

**Owner.** Tool change on a shared asset.

**Priority: now, as the schema and argument edit.** It is a small edit to a workstation
script, it changes no consumer that does not opt in, and it unblocks Item 2 once Item 1's
repo task lands the `--reduced` mode.

---

### Item 4. Rewrite the close task: numbers from a script, citations from grep, one fold with one read

**The measure.** Three changes to how a pass closes.

1. The close task's evidence numbers are produced by a script the conductor runs, not
   reconstructed by an agent. Gate wall per task, run counts, fix-round counts and
   verdicts all exist in the workflow journal and the spend ledger already.
2. Every `file:line` citation in the close's output is produced by `grep -n` and pasted
   with its matched line, never counted by hand.
3. The ledger fold is done once, by the conductor's fold agent, working from the segment
   reports and the close's numbers. The fold agent commits its draft before folding, and
   one `diff-reviewer` reads the fold's diff before it lands. The current shape is draft,
   review, fix, then a separate fold, which is the same text written twice.

**The fold keeps one independent read and one committed draft.** Revision 1 replaced the
close's in-chain `diff-reviewer` with a conductor-dispatched fold agent that carries none,
and subagents cannot dispatch agents (repo `CLAUDE.md`, build process), so the fold could
not commission its own. What that review catches is not all mechanical. 5b-i Task 10's
escalate listed seven blockers (`measure:158`); four are reachable by script or grep, and
two are not. "Two doc-comment claims that did not match the code" requires reading Go
against prose, and "a dropped 'awaiting Geoff's read' block" requires knowing what the
previous STATUS carried. Revision 1 classified the whole event as draft-content and then
reasoned as though every member of that class were mechanical. The house requirement that a
HISTORY entry carry "what a later pass would be wrong to rediscover from scratch"
(`~/.claude/CLAUDE.md`, "Project ledgers") is itself a judgment no measurement validates,
and under an unreviewed fold it would be written once and read by nobody before landing in
the ledger every future session opens. So one independent read stays: the conductor
dispatches one `diff-reviewer` over the fold's diff, at Item 2's reduced gate rather than
the full string.

The committed draft stays for the reason Item 11 depends on it.
`docs/superpowers/plans/2026-09-12-rung-5b-ii-auth-handlers.md:2953-2957` keeps the ledger
text as committed drafts because "they survive a restart and a fresh session can read
them," and Item 11 makes restarts more frequent rather than less. A single fold with no
committed draft would leave the pass's ledger text inside one agent's turn until it
returned. The fold agent commits its draft, then folds: the artifact survives and the text
is still written once.

**Owner-owed:** the close's authorship moves from a drafting implementer plus a folding
agent to one fold agent. Applied with one independent `diff-reviewer` read over the fold's
diff and the committed draft kept.

**Evidence.** Close-and-ledger work was 441,317, or 18.3% of the window
(`measure:96, :191`), and a docs-only close task cost 22 percent more than the average
code task at 151,374 against 123,793 (`measure:175-178`). Neither close's drafts were
accepted as written (`measure:192`). 8c Task 4 took two fix rounds costing 66,690, and the
second verdict was still six off-by-one line numbers the conductor then fixed by hand
(`measure:157, :185-187`). 5b-i Task 10 escalated with seven findings, costing 42,836 of
reviewer time, which the conductor corrected itself while folding; the re-dispatch was
explicitly skipped as not worth it (`measure:158, :188-190`). cairn reached the same place
from the other direction, pricing its own pass-end ritual as the concentration point of a
2x overrun (`harvest:13-17`).

**Expected token effect: 47,000 to 52,000 per pass.** Derivation, from the two measured
closes. The fix-and-escalate half is 66,690 on 8c Task 4 and 42,836 on 5b-i Task 10
(`measure:157-158`), a mean of 54,763 per pass, and measures 1 and 2 address its two named
causes directly, off-by-one pointers and missing measured wall times. The
draft-versus-fold duplication is the smaller of the close implementer's cost (86,260 and
68,154, `measure:41, :59`) and the fold's cost (58,884 and 79,684, `measure:43, :66`),
which is 58,884 and 68,154 respectively; half of each is 29,442 and 34,077, so 29,000 to
34,000. Against that, the independent read over the fold's diff costs one reviewer dispatch
at the measured mean of 37,036 (`measure:234`). Net 47,000 to 52,000, which is 1.9% to 2.2%
of the window total of 2,415,715 (`measure:91`). Revision 1 stated 85,000 to 115,000,
taking half of 58,884 and 68,154 as 30,000 to 60,000 and carrying no cost for the read.

**Expected clock effect: 15 to 25 minutes per pass.** Derivation: 8c Task 4's wall was
29.9m across an implementer, a review, and two fix rounds (`measure:41`); removing the two
fix rounds removes roughly two thirds of it. 5b-i Task 10's wall was 21.1m with no
re-dispatch (`measure:59`), so its saving is the review round trip alone, less the one read
the fold keeps. Add the three and two full gate runs those tasks made, about 6.8 minutes
(`measure:141-143`), already counted under Item 2 and not double-counted here.

**Quality safeguard.** Measures 1 and 2 strengthen the close outright. A number produced by
a script cannot be misremembered, and a citation produced by `grep -n` cannot be off by
one, which is precisely what the two closes' fix rounds were spent discovering
(`measure:157, :166-169`). Measure 3 removes a round trip and not a read: the fold's output
still meets one fresh `diff-reviewer` grading it against the close's acceptance criteria,
which is how 5b-i's equivalent claims were caught (`measure:158`), and its draft is still a
committed artifact a restart can recover.

**Precondition.** A script that emits the per-pass evidence block from the workflow journal
and the spend ledger, living in `tools/` and taking the session directory as an argument
rather than as a constant. The window proved the inputs exist, because the measurement this
plan is built on was itself reconstructed from those journals (`measure:3-8`), but no
artifact of that reconstruction is reusable as written: `scratchpad/gtime.py:2` and
`scratchpad/gates.py:2` hardcode a session-keyed transcript path, and
`scratchpad/spend-ledger.sh` is a nine-line `while true` heartbeat against the same
constant, so it cannot emit an evidence block for a pass it did not watch. The `tools/`
script is ruff-clean under `check.sh:101-108`, which lints every tracked `*.py`.

**Where it is encoded.** A `tools/` script, since the repo `CLAUDE.md` confines `scripts/`
to ops tooling. The plan template's close-task section, replacing the draft-then-fold shape
recorded at `docs/superpowers/plans/2026-09-12-overnight-window.md:470-495`. The repo
`CLAUDE.md` build-process section, which currently says nothing about how a close produces
its numbers.

**Owner.** Tool change plus a plan-template rule.

**Priority: now for the grep half; next pass for the script and the single fold.** The grep
rule costs nothing and removes the largest named fix-round sub-class. The script is a task
of its own. The single fold changes 5b-ii's Task 12 as ratified
(`docs/superpowers/plans/2026-09-12-rung-5b-ii-auth-handlers.md:2944-2960`), so it needs
a plan edit rather than a launch-time decision.

---

### Item 5. A pre-flight checklist in the implementer definition, keyed to dubplate's own fix causes

**The measure.** Add to `dubplate-implementer.md` a short pre-flight list the implementer
runs before reporting. Four entries are fix rounds the window already paid for; two are
cairn-recorded rounds adopted on shape and marked as such, because section 1 commits to
using cairn for the shape of a measure and never as a dubplate forecast.

- Every `file:line` citation in your report was produced by a command, and you paste the
  matched line beside it. (dubplate: `measure:157`.)
- Every count you state ("six named", "four rows") was produced by a command, and you
  paste the command. (dubplate: `measure:158`, a hand count stated as "(6 named)" that
  enumerated five.)
- No comment claims what its assertion does not prove. (dubplate: `measure:158`.)
- Counts found, changed, and deferred, with the deferred list named. (dubplate:
  `measure:158`.)
- No process citation (pass, plan, ruling id, task number) appears in a shipped comment.
  (cairn, on shape: `harvest:194-198`, `harvest:454`. The window's only comment-citation
  finding runs the other way: 8c Task 2's two blocking findings were doc-comment
  restorations the reviewer asked to have added, `measure:155`.)
- Re-emit the labeled report block verbatim before the gate. (cairn, on shape:
  `harvest:194-198`, `harvest:454`.)

**Evidence.** dubplate's own fix-round causes. 8c Task 3's two rounds cost 38,606 and both
concerned the enumeration table inside the task's return report, with no repo file changed
in the fix round at all (`measure:156`). 8c Task 4's second round was six off-by-one line
pointers (`measure:157`). 5b-i Task 10's escalate listed "(6 named)" enumerating five and
a phantom third simplifier round among its seven blockers (`measure:158`). Together the
comment-only, citation-pointer and draft-content class is about 71% of fix-round tokens
(`measure:163-165`). cairn built the same list from the same kind of counting and recorded
that "each one cost a full fix round on chassis-B2" (`harvest:192-216`).

**Expected token effect: a mean of 36,000 per pass.** Derivation: the citation-pointer
sub-class alone is 8c Task 3's 38,606 in full plus roughly half of 8c Task 4's 66,690
(`measure:156-157`), about 72,000 across two passes. The observed spread of a single fix
round is 38,606 to 66,690, which is the variance around that mean rather than a bracket on
it. Revision 1 stated 25,000 to 60,000 and attributed the range to a bracketing rule that
produces neither bound.

**Expected clock effect: about 3 minutes derived, up to 15 once the round trips are
counted.** Derivation: two fix rounds avoided, each carrying one full gate run at the mean
83s (`measure:121, :126`), is about 3 minutes. The rest is the two
implementer-plus-reviewer round trips, whose wall the window does not isolate. Revision 1
claimed 10 to 25 minutes from a contrast the cited rows contradict at both ends: 5b-i Task
9 took a fix round and ran 16.6m (`measure:58`), below the claimed fix-bearing floor of
22.0m, while two clean tasks ran above the claimed clean ceiling, 5b-i Task 2 at 26.1m and
Task 5 at 29.1m, both verdict accept (`measure:51, :54`). The true ranges are 16.6m to
43.5m for fix-bearing tasks and 12.1m to 29.1m for clean ones (`measure:38-42, :50-59`);
they overlap, and the two clean outliers are opus-upshifted tasks whose wall is unrelated
to fix rounds.

**Quality safeguard.** This item adds checks and removes none. The `diff-reviewer` still
grades every one of these independently (`harvest:208-210`), so a skipped pre-flight
entry fails exactly where it fails today.

**Precondition.** The repo's fix rounds are dominated by report-shape and citation
defects, which section 2 establishes by counting them. If a later pass's fix rounds are
dominated by real defects instead, this list stops paying and should be re-derived. The two
cairn entries carry no local cost and are the first candidates for that re-derivation.

**Where it is encoded.** `/var/home/glw907/Projects/dubplate/.claude/agents/dubplate-implementer.md`,
as a new section before "Report format" at `:93`. The report format at `:97-105` gains one
field under Item 6 and nothing else; the discipline is about how each field is produced.

**Owner.** Agent-definition change, owned by the conductor.

**Priority: now.** It is one edit to one agent file and it attacks the largest measured
waste class.

---

### Item 6. A mutation ledger and a finding severity as optional structured outputs

**The measure.** Two schema fields.

*The mutation ledger.* `dubplate-implementer.md:59-64` already requires the implementer to
run every named mutation, and `:102` already requires a report row per mutation. Promote
that prose row to an **optional** schema property in `pass-execute.js`'s IMPL_SCHEMA: an
array of `{mutation, site, provingTest, fired}`, never added to that schema's `required`
list at `:47`. `pass-execute.js` is the runner every repo uses and
`~/.claude/agents/cairn-implementer.md` never mentions mutations, so a required field would
ask a consumer for something its definition does not define. The reviewer is told to grade the table against
the plan's named mutations and to treat any `fired: false` row as a finding the
implementer has already surfaced rather than one the reviewer must reconstruct.

*The severity field.* Add an optional `severity` to each blocking finding in REVIEW_SCHEMA,
optional for the same shared-asset reason, since `~/.claude/agents/diff-reviewer.md` serves
both runners. Its values are `blocking-correctness`, `blocking-contract`, `comment-only`, and
`optional`. The label is advisory to the conductor and routes the reduced gate of Item 2;
it never bypasses a decision on its own.

**Evidence.** For the mutation half: MuTAP raises mutation score on synthetic buggy code
from 66% for Pynguin, a coverage-based generator, to 93.57% by forcing the kill-the-mutant
check
(https://arxiv.org/abs/2308.16557, `research:98-104`), and coverage flatters LLM-written
tests so that only a kill check separates a test that can fail from one that cannot
(`research:104-106`). dubplate's own record confirms the failure mode is live: the
reviewer repeatedly blocked on tests that could not fail (`research:20-28`, LOCAL-2), and
5b-i Task 3's fix round included exactly that finding, with acceptance criteria asserted
only against scratch registries so a wrongly declared key name could not fail the test
(`measure:154`), at about 29,000 of the fix spend (`measure:162`). The repo has also
recorded three mutations as unprovable rather than papered over
(`research:110-113`), which is the behavior this field makes legible.

For the severity half: Anthropic's guidance is to have the reviewer separate correctness
gaps from optional findings, because a reviewer prompted to find gaps will report some
even when the work is sound (https://code.claude.com/docs/en/best-practices,
`research:207-213`), and Jin and Chen find LLM reviewers frequently misclassify correct
code as defective, with prompts requesting detailed explanations increasing the error rate
(https://arxiv.org/abs/2603.00539, `research:205-209`).

**Expected token effect: 0 to 30,000 per pass.** The mutation ledger's saving is one
avoided round of the vacuous-test class, measured once at about 29,000 (`measure:162`),
and it may be zero in a pass where that class does not fire. The severity field's own
saving is zero; it is the routing key for Items 2 and 3.

**Expected clock effect: 0 to 12 minutes per pass.** One avoided fix round, at the
window's observed fix-bearing task wall premium (`measure:38-42`).

**Quality safeguard.** Both fields add information rather than removing a check. The
severity label is explicitly advisory, because no study measures LLM accuracy at severity
labeling and the research brief names this a genuine evidence gap
(`research:227-229`, `research:447-448`); the conductor still rules on every second `fix`.
A `fired: false` mutation row is a finding, exactly as `dubplate-implementer.md:64` already
requires.

**Precondition.** The plan names mutations per task, which every dubplate plan already
does (`dubplate-implementer.md:59-61`).

**Where it is encoded.** `~/.claude/workflows/pass-execute.js` IMPL_SCHEMA at `:37-48` and
REVIEW_SCHEMA at `:50-82`, as optional properties only; `~/.claude/agents/diff-reviewer.md` report format at `:52-65`;
`dubplate-implementer.md:93-105` report format.

**Owner.** Tool change plus two agent-definition changes, owned by the conductor. The
severity vocabulary is worth one line to Geoff, because it becomes the routing key for
gate reduction.

**Priority: now for the mutation ledger; now for the severity field as an advisory label.**
Neither changes a verdict rule.

---

### Item 7. Cap the plan-prep fan-out at three disjoint lenses, tighten the drafter's brief, and cap the fold

**The measure.** Three parts.

1. The default plan-review fan-out is three lenses, chosen for non-overlapping failure
   modes rather than for count. Five or more lenses are reserved for the three high-risk
   rungs (5, 12, 16).
2. The drafter is given a tighter brief: the sections it must produce, the sections it
   must not, a stated token target, and a required staleness pre-flight against committed
   state.
3. The fold is capped at one dispatch with a named budget, and a second fold round is a
   conductor decision, not a default.

**Evidence.** The ensembling result that supports a fan-out demonstrates its effect at
three: SWR-Bench reports multi-review aggregation raising F1 by up to 43.67% over
single-pass review (https://arxiv.org/abs/2509.01494), and the panel-of-judges result beat
a single GPT-4 judge at 7 to 8 times lower cost with a three-model panel, attributing its
bias advantage specifically to disjoint model families
(https://arxiv.org/abs/2404.18796); both are summarized at `research:279-297`. The
multi-agent literature is consistently negative on adding same-shaped parallel voices at
equal compute (https://arxiv.org/abs/2604.02460, https://arxiv.org/abs/2311.17371,
https://arxiv.org/abs/2503.13657, `research:293-297`). The cost fact is local: prep runs
22% to 49% of a pass and twice exceeded the whole implementer chain
(`research:36-40, :299-302`).

**The part of this the harvest overstates for dubplate, stated plainly.** The research
brief calls the lens cut "the largest single token saving available anywhere in the
pattern, on the order of 0.5M to 1.5M per pass" (`research:303-308`). For dubplate that is
only partly live, because the 5b-ii prep already ran three lenses: mechanics 89,542,
contract 101,412, security 95,422, totaling 286,376 of an 830,069 prep
(`measure:104-106`). The lens count was not dubplate's problem on that prep. The drafter
at 230,218 and the fold at 152,846 were the larger halves (`measure:104-105`), and the
pre-window standard prep's single fold agent at 564,846 in 30 turns was larger than either
(`measure:101-102`).

**Expected token effect: 0.05M to 0.2M on a prep already at three lenses; 0.3M to 1.2M on a
prep of the pre-window shape.** Derivation for the lower range: capping the fold at about
150,000 against the measured 152,846 saves nothing on 5b-ii's prep, so the saving there is
the drafter's discipline alone, which is not separately measured and is bounded by the
draft's 230,218 (`measure:104`). Derivation for the upper range: the fold cap alone saves
564,846 minus 150,000, about 0.41M, on the one measured over-running fold
(`measure:102`), and the ceiling of 1.2M is the difference between the two measured prep
shapes, 2.0M across 22 subagents against 830,069 across 7 (`measure:100-104`), reachable
only where a prep currently fans out that wide.

**Expected clock effect: 10 to 30 minutes per plan revision.** The 5b-ii prep ran 17:36 to
18:35 UTC, 59 minutes for 7 agents (`measure:103`); the pre-window prep ran 08:20 to 09:25
UTC, 65 minutes for 22 agents plus conductor (`measure:100-101`). The lens dispatches run
in parallel and contribute little wall, and the item keeps the drafter and all three
lenses, so what a cap can remove is the fold's serial tail inside a 59-minute or 65-minute
prep. Revision 1 stated 20 to 90 minutes, whose upper bound exceeds both measured preps end
to end.

**Quality safeguard.** Lenses stay read-only reviewers and findings are still folded with
a per-item taken-or-declined record (`harvest:400-402`). The three lenses are chosen for
disjointness, which is what the supporting studies actually measure. The high-risk rungs
keep the wider fan-out, so the place where a missed plan defect costs most loses nothing.
A second fold round remains available; it stops being automatic.

**Precondition, and the lens that must not be the one dropped.** A named disjoint-lens set
for this repo. The set is a contract-and-criteria lens, a mechanics-and-feasibility lens,
and a domain-risk lens (security for rungs 5 and 12, data integrity for the engine rungs).
Staleness moves into the drafter's own required pre-flight rather than spending a lens,
because it is mechanically checkable against committed state, which makes it the cheapest
of the four to move out of the fan-out. Revision 1 named contract, domain risk and
staleness, dropping the mechanics lens the 5b-ii prep actually ran (`measure:104-106`) and
leaving no lens to ask whether a plan's stated mechanism can be built. That is a live
defect class here: 8c Task 2 escalated on a plan-infeasible test outcome, an invalid-UTF-8
projection row the implementer judged unbuildable through `tagio.Write`, costing 53,948 in
chain plus 27,524 out of chain and producing the window's one commit that landed with no
`diff-reviewer` at all (`measure:155`). It is also the lens that would have caught this
document's own worst defect, because "does the runner interpolate this scope string before
or after the diff exists" is a mechanism question and not a contract, domain or staleness
one. If the staleness lens is judged load-bearing on its own, the cap rises to four rather
than mechanics being dropped.

**Owner-owed:** the default three-lens set. Applied as contract-and-criteria,
mechanics-and-feasibility, and domain-risk, with staleness moved into the drafter's
pre-flight.

**Where it is encoded.** The repo `CLAUDE.md` build-process section, which currently
states the lane and conducting loop but not a prep shape. The plan template's header,
which should record the lens set and the fold budget the way it records the token ceiling.

**Owner.** Geoff ruling for the default lens count and the named set, then a plan-template
rule for the fold cap.

**Priority: next pass.** 5b-ii's plan prep is already done, so nothing here applies to its
launch. It applies to the next plan drafted, which is 5c.

---

### Item 8. Batch the simplifier per pass or per merge, never per segment, and chunk its review by package

**The measure.** One `code-simplifier` round per pass, dispatched at the last segment
boundary before the close, over the whole pass's accepted Go. Not one per segment. The
round's review is chunked by package territory, one `diff-reviewer` dispatch per package,
so tokens per reviewed line hold at today's level.

**Evidence.** Four rounds in the window cost 211,840, or 8.8% (`measure:88`), at 38,364 to
63,160 each with a mean of 52,960 (`measure:200-201`). Those rounds are already the
correct economy relative to the per-task alternative; the 2026-09-12 window recorded the
deviation explicitly, replacing eleven interleaved dispatches with three batched rounds
(`docs/superpowers/plans/2026-09-12-overnight-window.md:163-166`), and 5b-ii's plan carries
four (`docs/superpowers/plans/2026-09-12-rung-5b-ii-auth-handlers.md:3179-3182, :3236-3242`).
The research brief judges the simplifier the least-supported element of the whole pattern:
unguarded LLM refactoring changes behavior at 6.6% to 7.4% of attempts across 180 real
refactorings (https://arxiv.org/abs/2411.04444) and at about 11.5% across 8,096 segments in
a larger 2026 study the brief records as paywalled and SECONDARY, meaning it was not
verified against the primary text (`research:240-242`, `research:3-4`); its measured benefit
here is line-neutral diffs, one round that changed nothing, and several proposals the
reviewer declined (`research:230-263, :425-441`). What keeps it is SlopCodeBench, which
finds structural erosion rising in 77% of long-horizon trajectories and, critically, that
explicit quality guidance reduces initial verbosity and erosion by up to a third without
affecting degradation rates (https://arxiv.org/abs/2603.24755, `research:250-257`).
Prompting moves the intercept; only a periodic cleanup acts on the slope. The brief's own
conclusion is to keep it, batch it per rung rather than per segment, and never run it
without its review (`research:258-263`).

**Why the review is chunked rather than stretched.** The brief's adjustment is narrower
than the batching alone: "Never drop the simplifier's own review; that review is what makes
the 7.4% figure not apply here" (`research:258-263`). A single round over four segments of
accumulated Go asks one reviewer to read a roughly four-times-larger diff, and the
behavior-change rates it guards against are per refactoring rather than per round
(`research:234-242`), so batching holds the expected number of behavior changes constant
while lowering the probability that any one is seen. Revision 1 assumed sub-linear
attention outright, at "1.5x to 2x a segment round" for a round covering four segments.
This document treats context degradation as first order everywhere else: section 4 defends
the thin conductor with needle accuracy dropping 20% to 50% between 10k and 100k tokens and
10 of 13 models falling below half their short-context baseline at 32K (`research:61-70`).
One `diff-reviewer` per package territory holds tokens per reviewed line where they are
today, at the measured reviewer mean of 37,036 (`measure:234`) against the simplifier round
mean of 52,960 (`measure:201`), so even three extra reviews leave the item net positive.

**Expected token effect: 53,000 to 80,000 on a pass of 5b-i's shape; zero on a pass that
already runs one round.** Derivation, corrected for what the measure actually cuts. The
window ran four rounds across two passes: 8c ran exactly one at 52,582 (`measure:43`) and
5b-i ran three at 159,257 (`measure:65`). One round per pass is two rounds, not one, so the
cut is 5b-i's three to one, two rounds removed, and 8c is unchanged. The single remaining
round covers three segments of accumulated Go, so at 1.5x to 2x a segment round it costs
79,000 to 106,000 against the mean 52,960 (`measure:201`), leaving 159,257 less that, which
is 53,000 to 80,000, or 2.2% to 3.3% of the window total (`measure:91`). Revision 1 stated
105,000 to 155,000 by removing three rounds and by pairing its bounds in a way that follows
from neither figure it subtracted.

**Expected clock effect: 5 to 15 minutes per pass.** Two fewer dispatches plus the full
gate run each one's commit requires, at the mean full run of 83s (`measure:121, :126`),
less the extra reviewer dispatches the chunked review adds.

**Quality safeguard.** The round still runs behind the full test suite and a diff review,
which is the guard the refactoring papers themselves recommend and the reason this repo's
record shows zero regressions across roughly a dozen rounds
(`research:244-249`, LOCAL-5 at `research:42-47`), and the per-package chunking is what
keeps that review's depth constant as the round's diff grows. Two additions, both from the
research brief. A pass that ran unattended past a long window takes its simplifier round as
non-optional, and that round's reviewer is told to look for erosion and verbosity
specifically, not only for reuse and altitude (`research:365-379`). And the round never
runs without its review, which is the clause that makes the 7.4% figure inapplicable here
(`research:261-263`).

**Precondition.** The pass's Go must be accepted before the round runs, which the segment
structure already guarantees at the last boundary. One conductor-owned dispatch per
package, since subagents cannot dispatch agents (repo `CLAUDE.md`, build process).

**Where it is encoded.** The plan template's execution-model section, replacing the
per-segment simplifier row in the segment table. The repo `CLAUDE.md` build-process
section, which currently states the conductor owns the simplifier dispatch but not its
cadence.

**Owner.** Conductor rule, stated in the plan template. Worth one line to Geoff when the
item is ruled: it widens the standing deviation from `~/.claude/CLAUDE.md`'s
run-the-simplifier-before-committing rule, which the chain already deviates from on every
task because its implementer commits its own files
(`docs/superpowers/plans/2026-09-12-overnight-window.md:167-173`).

**Priority: now, in its half-step form.** Cutting 5b-ii's four rounds to one at D2 is a
larger change than a launch decision should make, because segment A's boundary exists
partly to let later tasks build on cleaned code. Cutting four to two, one after segment C
and one at D2, is the safe form and captures about half the saving. The full cut is next
pass. The per-package chunking applies to whichever form runs.

---

### Item 9. Keep a cross-family adversarial read for rungs 5, 12, and 16

**The measure.** At the three rungs that already spend a full fresh-context adversarial
cycle (repo `CLAUDE.md`, build process: "auth (5), streamer/SSO (12), and cutover (16)
keep the full fresh-context adversarial cycle"), one of the reads is done by a model
outside the Anthropic family. This item **adds** cost.

**Evidence.** The asymmetric-review result is the cleanest number in the research brief:
across 116 LiveCodeBench hard and medium tasks in all four directions, Claude reviewing
Codex raised pass rate from 71.6% to 89.7% (p=.001), Codex self-review reached 84.5%
(p=.022), and Codex reviewing Claude *lowered* pass rate from 91.4% to 82.8% (p=.046)
(https://arxiv.org/abs/2607.21656, `research:161-169`). Opus reviewing Sonnet is on the
right side of that arrow, so the arrangement is sound and stays. The unaddressed risk is
same-family correlated blind spots, documented for self-preference with scores above the
0.5 no-bias line (https://arxiv.org/abs/2404.13076), for the familiarity mechanism that
would carry across a family (https://arxiv.org/abs/2410.21819), for same-family rating bias
directly (https://arxiv.org/abs/2508.06709), and as a rising trend across frontier models
(https://arxiv.org/abs/2502.04313); all four at `research:171-184, :404-414`. The panel
result attributes its bias advantage specifically to disjoint model families
(https://arxiv.org/abs/2404.18796, `research:284-289`).

**Expected token effect: plus 80,000 to 150,000 per high-risk rung.** Derivation: a
reviewer dispatch's measured mean is 37,036 (`measure:234`), and a full fresh-context
security read on this repo measured 82,851 (`measure:106`). The upper bound allows for a
second-vendor reviewer with no repo context needing more reading.

**Expected clock effect: plus 20 to 40 minutes per high-risk rung.** One serial read at the
pass-end ritual.

**Quality safeguard.** This item *is* a safeguard. It is the one place in this document
where spending rises, and it rises at exactly the three rungs where a missed defect is
most expensive. Nothing else changes: the in-chain Opus reviewer stays on every task.

**Precondition, and the reason this item may be blocked.** A non-Anthropic model must be
reachable from a dispatch in this harness. If it is not, the item cannot be executed as
written. The available fallback is weaker and should not be sold as equivalent: a second
fresh-context read under a deliberately different rubric and a different agent definition,
which decorrelates the prompt but not the model. The literature's claimed gain comes from
decorrelated errors, not from the second model being better (`research:412-414`), so the
fallback captures only part of it.

**Where it is encoded.** The repo `CLAUDE.md` "Reviewers" section, which currently names
`web-auth-security-reviewer` for the serve and JWT surface without naming a model family.
The rung 5, 12 and 16 plan templates' review sections.

**Owner.** Geoff ruling. It costs money at a vendor the workstation does not currently use,
and it is the one item that raises spend.

**Priority: next pass, contingent.** 5b-ii is a rung 5 pass and qualifies on subject
matter, but the precondition is unverified. The right next step is to check reachability
and report, not to plan around it.

---

### Item 10. Task 0 takes no gate; the lane's baseline is the conductor's one call

**The measure.** The staleness pre-task is read-only and runs no gate. The lane's gate
baseline is taken once by the conductor at lane launch, through `cairn-run-gate`, before
the first dispatch. Task 0's dispatch notes tell the reviewer to report `gate: "pass"` with
the conductor's baseline line quoted as its evidence, and Task 0's citations are produced
by `grep -n` under Item 5's rule, so its output is verifiable without re-reading the files.

**Why the reviewer is told what to report.** `pass-execute.js:140-146` returns `"accepted"`
only when `review.verdict === "accept"` and `review.gate === "pass"`; every other
combination falls through to `"needs-decision"`. A Task 0 that runs no gate would have its
reviewer report `gate: "not run"`, a legal enum value at `:78`, and the task would return
`needs-decision` on an `accept` verdict, land under `needsDecision` in `tally()` at
`:220-230`, and print at `:274` as a segment carrying an unresolved task. One conductor
ruling per lane launch erases the item's whole measured saving of 3,740
(`measure:37, :49`). Quoting the conductor's own baseline line as the gate evidence needs
no code change and reports a gate that genuinely ran. The alternative, a per-task
`noGate: true` that `taskStatus` honors alongside an accept verdict, is a runner edit and
is the better form if Item 3's edit is being made anyway.

**Evidence.** The two Task 0s did the same job by different means. 8c's ran zero gate runs
in 4.2m for 38,236 tokens (`measure:37`). 5b-i's ran the gate three times for 205s, a 32%
gate share on a read-only probe, in 10.6m for 41,976 (`measure:49`). The measurement calls
this out directly: 5b-i Task 0 "spent 205s more running the gate plus `go test ./...`
purely to record a baseline" (`measure:143-144`). The window's own plan already caps a
Task 0 at 0.3M and stops its lane if it exceeds that
(`docs/superpowers/plans/2026-09-12-overnight-window.md:121-124`), so the scope discipline
exists; the gate exemption does not.

**Expected token effect: 0 to 5,000 per lane.** Derivation: the difference between the two
measured Task 0s, 41,976 minus 38,236, is 3,740 (`measure:37, :49`). This is a small item
and should not be oversold.

**Expected clock effect: 3 to 6 minutes per lane.** The 205s of gate plus `go test ./...`
(`measure:144`), plus the wall difference between the two probes, 10.6m against 4.2m
(`measure:37, :49`), of which the gate is the named cause.

**Quality safeguard, and the thin-conductor exemption it needs.** The baseline still exists
and is still taken before the first implementer dispatch, which is what it is for. Moving
it to the conductor makes it one call whose result the conductor holds, rather than a
number inside a read-only agent's report the conductor must trust. The staleness verdict
itself is unchanged. The conductor reading that result does read a gate tail, because
`cairn-run-gate` "blocks to completion and prints the tail" (repo `CLAUDE.md`), while the
thin-conductor rule says the conductor reads "no source file, no diff, no gate transcript"
during execution (`~/.claude/CLAUDE.md`, "Conducting a pass";
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:234`). The exemption is
stated rather than left implicit: the conductor reads the gate's pass or fail line at lane
launch only, never during execution, and a red baseline stops the lane launch rather than
opening a diagnosis in the main loop. That boundary is what keeps the measured 6.4%
conductor share at 1,475 tokens per turn (`measure:216-219`) from drifting.

**Precondition.** A green master or a green worktree at lane launch, which the lane-launch
checklist already requires
(`docs/superpowers/plans/2026-09-12-overnight-window.md:76-78, :116-120`).

**Where it is encoded.** The plan template's Task 0 section, and the lane-launch checklist
in `docs/superpowers/specs/2026-09-03-parallel-build-approach.md` section 4.

**Owner.** Plan template plus a conductor rule.

**Priority: now.** It is a one-line change to a task's notes at launch.

---

### Item 11. Close the session at a window's end rather than re-prime after an idle gap

**The measure.** When a window's execution ends and the next work is hours away, close the
session. The next window starts fresh from STATUS, the resume prompt, and the segment
ledger. Do not hold a large conversation open across an idle gap.

**Evidence.** The single most expensive turn block in the window was the resume. One
conductor turn block at 17:33 UTC, after the roughly four-hour idle gap between the 5b-i
close and the prep phase, carries `cache_creation = 574,539` in a single request, which is
303,912 unit across 4 turns, against 154,894 for the conductor's entire 105-turn execution
window (`measure:109-113`). The prompt cache expired over the gap and a 605K-token
conversation was re-primed from scratch (`measure:113-114`). At 76,000 per turn against
1,475 per turn for active conducting, one idle gap cost twice what four hours of active
conducting cost (`measure:220-223`). The workstation rule already says this in principle:
"every turn re-reads the whole cached conversation, so a long session's meter compounds
even with disciplined steps" (`~/.claude/CLAUDE.md`, "Initiative-scoped sessions").

**Expected token effect: up to about 262,000 per idle gap avoided, an upper bound.**
Derivation: the measured re-prime was 303,912 unit (`measure:110-111`), and the saving is
that figure less a fresh session's own cold start, which is not measured. The window's
Task 0 dispatches, at 38,236 to 41,976 (`measure:37, :49`), are the nearest proxy and
under-price it: a Task 0 is a subagent with a task-scoped read list, while a fresh
conductor session reads both `CLAUDE.md` files, the memory index and its entries,
`docs/STATUS.md` in full, and a pass plan that ran to 3,344 lines for 5b-ii. So 303,912
less about 42,000, or about 262,000, is a ceiling and not a midpoint, and revision 1
presented it as a midpoint. The next window prices the cold start properly from the
conductor session's own first-turn `cache_creation`, which the transcripts carry.

**Expected clock effect: neutral.** A fresh session reads its handoff faster than a stale
one re-primes, but the difference is not measured and should not be claimed.

**Quality safeguard.** The handoff artifacts already exist and are already required. The
resume prompt is written to STATUS at every segment boundary, at every split, and at any
stop, and carries per-lane state, the workflow `runId` and its transcript directory,
committed tasks, the task in flight, the last reviewer verdict, any task committed whose
review has not returned, and the spend ledger reading
(`docs/superpowers/plans/2026-09-12-overnight-window.md:505-528`). Nothing is lost that
was not already written down; the window's own design assumed a fresh session might
resume, and recorded that `resumeFromRunId` is same-session only while the journal is the
cross-session path (`:530-539`).

**Precondition.** The resume prompt, the lane registry, and the segment ledger are
current in STATUS at the moment of the close. That is a discipline the window already
runs.

**Where it is encoded.** `~/.claude/CLAUDE.md`, "Initiative-scoped sessions," which states
the principle per initiative and should state it per idle gap. The repo `CLAUDE.md`
build-process section, whose phase rule ("Between phases: close the session, post-mortem,
STATUS handoff") should extend to a within-phase idle gap.

**Owner.** Conductor rule; a `CLAUDE.md` amendment is Geoff's to ratify.

**Priority: now.** It is a behavior, not a change to anything.

---

### Item 12. Cut only the boundaries whose sole work is a simplifier round

**The measure.** Keep every segment boundary that carries a conductor decision. Cut a
boundary whose only work is a simplifier round and a STATUS write, which Item 8 empties of
its simplifier anyway.

**Evidence, including the part that argues against the general form of this item.** A
boundary costs 65,000 to 85,000, four boundaries cost about 272,000 or 11% of the window,
and the marginal cost of segmenting into five runs instead of two was about 100,000 plus
relaunch turns (`measure:200-209`). Against that, the boundaries are not decoration.
`pass-execute` loops over `args.tasks` with no pause point and no conductor hook, so a
segment runs to its end or not at all and every conductor action lands at a boundary and
nowhere else (`docs/superpowers/plans/2026-09-12-overnight-window.md:132-140`;
`docs/superpowers/plans/2026-09-12-rung-5b-ii-auth-handlers.md:3226-3235`). 5b-ii's segment
D splits at the Task 10 and Task 11 boundary precisely so that the rule that a second `fix`
is the conductor's decision has a place to be taken, without which Tasks 11 and 12 would
already be implemented, reviewed, gated and committed against a handler set the conductor
had not accepted (`:3230-3235`). And there is no within-segment cache sharing to recover:
`pass-execute` spawns a fresh agent per task, so splitting a pass across runs costs the
ritual, not the reads (`measure:211-214`).

**Expected token effect: 0 to 85,000 per pass.** Derivation: one boundary at the measured
65,000 to 85,000 (`measure:205`), and only where a boundary's work is genuinely empty once
Item 8 lands. On 5b-ii's five segments, boundaries A, B and C each carry a simplifier plus
a named confirmation (`5b-ii:3236-3242`); A's confirmation, that the five reproductions
were observed failing before their fixes, can be moved into Task 1's acceptance criteria,
which would make A a candidate. D1 and D2 carry decisions and must not be cut.

**Expected clock effect: 5 to 15 minutes per pass.** One relaunch's conductor turns and one
fewer full gate on a simplifier commit.

**Quality safeguard.** No decision point is removed. The rule is stated as a test: a
boundary survives if and only if the next segment builds on something the conductor must
accept, rule, or measure first.

**Precondition.** Item 8 must land first, because a boundary carrying a simplifier round is
not empty.

**Where it is encoded.** The plan template's execution-model section, as the criterion for
placing a boundary.

**Owner.** Plan template.

**Priority: later.** 5b-ii's boundaries are ratified and each carries named work; moving
one changes an approved plan mid-flight for a saving the plan can capture next time.

---

### Item 13. Re-key the 80 percent flag to the segment boundary

**The measure.** The ceiling check compares spend to date against tasks remaining at every
segment boundary, rather than firing once at 80 percent of a total. The boundary is the
only place a decision can land, so it is the only place a flag can change one.

**Evidence.** The flag is crossed routinely and usually after every deliverable has landed,
and one HISTORY entry records the conductor crossing the flag inside the chain and not
acting on it, because there was no lever left to pull
(`research:342-363`, citing `docs/HISTORY.md:1688, :2173, :2580, :2914`). The research
brief's verdict is that the ceiling functions as a post-hoc accounting line more than a
live control (`research:356-359`). cairn reaches the same place from its own overrun
record, four for four on comparable passes (`harvest:351-377`). The structural reason is
the same one Item 12 names: a segment runs to its end or not at all
(`5b-ii:3226-3229`).

**Expected token effect: 0 directly.** It changes no dispatch. Its value is that it makes
every other ceiling in this document enforceable, and that a pass projected past its
ceiling can be cut at a boundary instead of discovered over at its close.

**Expected clock effect: 0.**

**Quality safeguard.** The flag triggers a question, not a silent scope cut
(`harvest:372`). That is unchanged. What changes is that the question arrives while an
answer is still possible.

**Precondition.** A spend ledger the conductor can read mid-run, which the 2026-09-12
window built (`docs/superpowers/plans/2026-09-12-overnight-window.md:45-46`). Its one
artifact is session-keyed, so Item 4's `tools/` script is what makes the reading
reproducible on the next pass.

**Where it is encoded.** The plan template's token-ceiling section, which in 5b-ii already
names a re-derivation at the segment B boundary
(`5b-ii:3337-3338`) and should generalize that to every boundary.
`~/.claude/CLAUDE.md`, "Conducting a pass," which states the 80 percent rule as a level
rather than a checkpoint.

**Owner.** Plan template; the `CLAUDE.md` wording is Geoff's.

**Priority: now.** It costs nothing and it is already half-present in 5b-ii's plan.

---

### Item 14. Fire the cross-lane review on semantic overlap as well as rebase mechanics

**The measure.** The merge ritual's `diff-reviewer` over the merged range fires when the
rebase was not a fast-forward, **or** when the two lanes touched a shared package or a
shared contested surface. The trigger becomes a disjunction; nothing is removed from it.

**Why an addition rather than a swap.** Revision 1 replaced the rebase trigger with the
overlap trigger and claimed the swap was two-sided, adding "one review where a fast-forward
hid a shared-package overlap." That case cannot occur.
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:166` states the current rule
and its reason: "A fast-forward rebase means the other lane landed nothing, so this step
skips mechanically and costs nothing." A fast-forward carries no other-lane commits, so
there is no overlap for it to hide. The swap's added half was empty and its removal half
was real. What it removed is named in the lane ownership table at `:215`, which records
`.gitleaksignore` as a contested surface and states the consequence: "a concurrent append
is a non-fast-forward rebase and triggers the merge ritual's re-review." Under the swap
that trigger would disappear whenever the lanes' package territories were disjoint, which
is the normal case the ownership table is designed to produce. Two lanes both appending an
allowed finding, one resolution keeping its own line and dropping the other's, and both
gates passing, is a defect no test catches and no reviewer would have read. The same holds
for every hand-resolved ledger conflict, where the workstation `CLAUDE.md`'s merge rules
("STATUS takes main's, HISTORY keeps both") describe resolutions nothing verifies.

**Owner-owed:** the merge ritual's cross-lane review trigger, a sentence in the repo
`CLAUDE.md` build process and in `2026-09-03-parallel-build-approach.md` section 3's step
5. Applied as a disjunction that adds the shared-package case and keeps the
non-fast-forward case.

**Evidence.** The current trigger is stated as the rebase shape: one `diff-reviewer` on the
merged result "whenever the rebase was not a fast-forward" (repo `CLAUDE.md`, build
process), and the window priced exactly one such review at 41,936 (`measure:89`). The
field evidence says something is missing from it: in a practitioner incident log of
concurrent sessions on one repo, conflict probability tracked task similarity rather than
session count (https://readysolutions.ai/blog/2026-06-06-parallel-claude-code-sessions-one-repo/,
`research:325-333`). Package overlap is therefore a risk signal the current rule does not
read, which is an argument for adding it and not for trading it against the one already
there.

**Expected token effect: 0 to plus 42,000 per lane close.** It adds one review where two
lanes shared a package or a contested surface and the rebase was a fast-forward, at the
measured 41,936 (`measure:89`). It removes none. On a two-lane window of the 2026-09-12
shape, where 8c and 5b-i were disjoint by package and the one shared directory was excluded
(`docs/superpowers/plans/2026-09-12-overnight-window.md:208-212`), it costs nothing.

**Expected clock effect: 0 to plus 10 minutes per lane close.**

**Quality safeguard.** The trigger covers strictly more of the risk it exists to catch. The
spec already records why the review is owed at all: a lane's reviewer is structurally blind
to the other lane's landed work, so isolation is the mechanism that hides a cross-lane
defect rather than a safeguard
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:231`). Package overlap and
merge-base arithmetic are each a partial proxy for that blindness, and the disjunction is
both.

**Precondition.** The lane ownership table names each lane's package territory and its
contested surfaces, which it already does
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:215` and its table).

**Where it is encoded.** The repo `CLAUDE.md` build-process section's merge-ritual
sentence, and `2026-09-03-parallel-build-approach.md` section 3's step 5.

**Owner.** Conductor rule; the `CLAUDE.md` sentence is Geoff's to ratify.

**Priority: later.** 5b-ii is a single-lane pass. This applies at the next two-lane window.

---

### Item 15. Eliminate the residual gate-log polling through the implementer definition

**The measure.** State in `.claude/agents/dubplate-implementer.md` that the gate is run
through `cairn-run-gate`, that the runner blocks to completion and prints the tail itself,
and that the implementer therefore never tails the gate log and never sleeps waiting on it.
One run, one result, no poll loop.

**The wording that licenses the polling is in the definition today.** Step 5 at
`.claude/agents/dubplate-implementer.md:65-66` reads "Run `bash scripts/check.sh` (about
three and a half minutes; the conform sweep alone is two) until it ends with `check.sh: all
checks passed.`" The definition names no runner, gives the gate's duration, and says
"until", which is an instruction to wait and check. `~/.local/bin/cairn-run-gate` already
"blocks to completion and prints the tail" (repo `CLAUDE.md`, "Gate economy on a pass"), so
the waiting has nothing to do. The fix is the same three lines Item 1 already edits, at
`:3`, `:65` and `:99`, which is why this item costs nothing to land alongside it.

**Evidence, and the reason the effect is small here.** cairn measured the whole of this
change. Before its runner landed, one task ran 75 minutes across 381 shell calls at up to
427K tokens of context per call, and another spent 263 of its 562 calls polling a log; after
it, one task made four runner calls and zero polls, and per-agent cost roughly halved across
comparable stretches, 1.73M for 14 agents against 0.75M for 15 (`fresh:5-12`). dubplate has
already taken the larger half of that: every chain agent in the 2026-09-12 window shows
`cairn-run-gate` calls, so the gate is already run through the blocking runner
(`fresh:20-22`). What remains is a residue. The heaviest task in the window, agent
`a4c457b6` in run `wf_c808ec5e-1ba`, made 118 shell calls, and the worst single agent made
fourteen `tail`-or-`sleep` calls (`fresh:23-24`), against cairn's pre-runner 381 and 562
(`fresh:5-6`, `fresh:24-25`). This item collects the residue, not the 2x.

**Expected token effect: small and not separately measured; tens of agent turns per pass.**
Derivation: fourteen poll calls is the observed per-agent ceiling and most agents are well
below it (`fresh:23-24`), so a pass of 5b-ii's size, at roughly nineteen implementer
dispatches on the window's ratio of 20 across fourteen tasks (`measure:229-230`), carries
tens of poll turns rather than
hundreds. The window's measurement carries no per-tool-call token line, so no token figure
is claimed here; what is claimed is that each removed poll removes one agent turn that
re-reads the agent's context. cairn's own before-and-after is the only magnitude evidence
and is not transferable, for the reason section 1 gives.

**Expected clock effect: small, and bounded by the sleep intervals themselves.** A poll loop
adds its own sleeps to a gate that would have returned at the same moment regardless.

**Quality safeguard.** None is needed, and that is the point: the gate runs to completion
exactly as it does today, the implementer reads the same tail, and the report format's
**Gate** field at `.claude/agents/dubplate-implementer.md:99` is unchanged. No check is
scheduled differently, scoped differently, or removed. This is the one item in the document
with no quality argument on either side.

**Precondition.** None. `cairn-run-gate` is installed and already in use.

**Where it is encoded.** `.claude/agents/dubplate-implementer.md:3, :65, :99`, the same three
lines Item 1 rewrites to let a dispatch name the gate string. Step 5's "until it ends with"
becomes a single blocking call.

**Owner.** Agent-definition change, owned by the conductor.

**Priority: now.** It is one edit to one agent file, it carries zero quality risk, and its
effect is small by construction.

---

### Item 16. Segment a pass at three to four tasks, every boundary on a green commit

**Ruled by Geoff, 2026-09-12 ("That works").** This is the document's only ruled item; every
other item remains a proposal.

**The measure.** The rule as `docs/superpowers/research/2026-09-12-segment-sizing-research.md`
states it (`sizing:488-499`): segment a pass at three to four tasks, and place every boundary
on a commit the gate has already proved green. The count, never the placement, is overridden
on an irreversible or production-touching task (boundary immediately before it), a second
`fix` verdict from the `diff-reviewer` on one task (boundary immediately), or a seam where a
task's Files or acceptance criteria are disjoint from the next task's (boundary at the seam).
The handoff across a boundary is structured state in the repo, the STATUS ledger, the plan's
task table and the git log, never conversation the next segment has to re-read.

**The derivation, and why it lands on dubplate's own numbers.** The sizing read models
expected overhead per useful task as `H(S) = δ_eff / S + S / (2M)`, where S is tasks per
segment, δ_eff is the boundary cost in task-equivalents with its expected handoff loss folded
in, and M is the mean number of tasks between faults (`sizing:437-450`). Minimizing gives
Young's checkpoint-interval formula unchanged, `S* = sqrt(2 · M · δ_eff)` (`sizing:452`). At
δ_eff about 0.5 and a fault every 8 to 12 tasks it returns 3 to 4 (`sizing:461-478`).

Both of that formula's inputs are measurable here, and section 2 already carries one of them.
A boundary costs 65,000 to 85,000 (`measure:200-205`, `measure:205`), four boundaries about
272,000 or 11% of the window (`measure:207`), against a mean code task of 123,793 through the
chain (`measure:175-176`). That is a measured δ_eff of 0.53 to 0.69 before handoff loss,
against the 0.4 to 0.5 the sizing read assumed from a wider per-task band (`sizing:461-465`).
The higher figure moves S* up rather than down: at δ_eff 0.6 and M of 8 to 12, `S*` is 3.1 to
3.8, which is the same 3 to 4.

**The one input the window does not settle, stated plainly.** M is defined as tasks between
faults "that need a redirect or a rollback" (`sizing:441-442`), and the window can be counted
two ways against that definition. Six fix or escalate events across fourteen tasks
(`measure:148`) gives M of 2.3 and an `S*` of about 1.7, which would argue for segments of
two. One real code defect across the same fourteen tasks (`measure:153, :161`) gives M of 14
and an `S*` of about 4.1. The ruled 3 to 4 sits between the two readings, and which one is
right is exactly the question the fix-round cause classification in section 6 answers, so the
next pass's classification is what recomputes this rather than settling it now. The fix
rounds that were comment, citation or draft defects, about 71 percent of fix spend
(`measure:163-165`), are the ones Items 4, 5 and 6 exist to remove, and removing them moves M
toward the upper reading without changing the rule.

**Why the exact number matters little.** Being wrong by a factor of two in either direction
costs 25% more overhead, `H(2S*) = H(S*/2) = 1.25 · H(S*)` (`sizing:454-455`). The rule is
therefore worth stating once and not worth tuning often.

**Evidence.** The sizing read is a derivation over converging principles rather than a
published number, and it says so at its top: no vendor, framework or paper publishes a tasks
per segment figure (`sizing:8-23`). What it adds to this document is that dubplate's current
practice is already at the optimum, so the rule ratifies an existing shape rather than
changing one. The clause forbidding a context-fill percentage trigger is explicit
(`sizing:518-526`): the published fill numbers disagree, the best source rejects a universal
percentage outright, and the step-count literature finds steps rather than context length to
be the driver.

**Expected token effect: 0 as a change of practice; it is a guard against drift.** The window
ran five `pass-execute` launches over fourteen tasks (`measure:200`), already inside the
band. Its value is that a later pass that segments at seven or at one now has a stated figure
to be wrong against.

**Expected clock effect: 0.**

**Quality safeguard.** The rule adds two boundaries that do not exist as rules today, before
an irreversible task and immediately on a second `fix` verdict, and removes none. The
green-commit placement is what makes a boundary a resume point at all, and it agrees with
5b-ii's segment D split, which exists precisely so the second-`fix` decision has a place to
be taken (Item 12's evidence paragraph;
`docs/superpowers/plans/2026-09-12-rung-5b-ii-auth-handlers.md:3230-3235`).

**Interaction with Item 12.** Item 12 cuts a boundary whose only work is a simplifier round.
This item sets the count those cuts are measured against. They agree: Item 12's test is that
a boundary survives if the next segment builds on something the conductor must accept, and
this item's three overrides are that test's named cases. Where they would conflict, on a
segment that Item 12's cut would stretch past four tasks, this item's count holds and Item 12
takes the shorter segment.

**The three measurements to keep, plus one diagnostic.** The sizing read names them
(`sizing:528-556`), and they join section 6's table rather than living here:

1. **δ_eff, the boundary cost in task-equivalents.** Boundary tokens divided by the mean
   task tokens of the segment it closed. A δ_eff drifting above 1.0 means the boundary ritual
   has grown and should be trimmed rather than the segment lengthened (`sizing:533-538`).
2. **M, tasks per fault.** Tasks completed divided by the count of `fix` verdicts, gate
   failures that survived the implementer, and rollbacks or re-dispatches. This is the number
   that moves; δ_eff is nearly constant (`sizing:539-542`).
3. **Fault-detection latency in tasks.** Tasks run between a defect being introduced and
   caught. A mean near S/2 means the model behaves as assumed; a mean near S means faults are
   reaching the boundary rather than the in-chain reviewer, and the cheaper fix is a better
   acceptance criterion (`sizing:543-549`).

The diagnostic is nearly free and is worth adding: log whether each fault was caught by the
in-chain `diff-reviewer`, by the gate, or by the conductor at the boundary (`sizing:550-556`).
It is the same log section 6 already requires for a different reason, recording where each
blocking finding was caught.

**Precondition.** The δ_eff and M rows need the per-pass evidence block Item 4's `tools/`
script emits; both are reconstructable from the workflow journal and the spend ledger today
but were reconstructed by hand for this document (`measure:3-8`).

**Where it is encoded.** The plan template's execution-model section, as the segment count
and the boundary-placement rule; the repo `CLAUDE.md` build-process section, whose lane and
conducting-loop paragraphs state no segment size; and section 6's metric table for the three
measurements.

**Owner.** Ruled by Geoff. What remains is a plan-template edit and a `CLAUDE.md` sentence,
owned by the conductor to draft.

**Priority: now.** 5b-ii's five segments over thirteen tasks already sit in the band, so the
rule ratifies the launch rather than changing it, and the three measurements start being
recorded at the next close.

---

## 4. What not to change

Nine elements the evidence says keep. Each is named with what would be lost.

**The thin conductor.** Its share of the window was 6.4% at 1,475 per turn across 105
turns, and the discipline demonstrably held: the conductor read no diffs and no gate logs
(`measure:216-219`). The research brief calls it the best-supported element of the whole
pattern (`research:83`), with Anthropic's guidance on keeping verbose output in subagent
context (https://code.claude.com/docs/en/best-practices, `research:53-59`) and the
context-degradation literature behind it: needle accuracy drops 20% to 50% between 10k and
100k tokens (https://www.trychroma.com/research/context-rot), and 10 of 13 models claiming
128K context fall below 50% of their short-context baseline at 32K
(https://arxiv.org/abs/2502.05167), both at `research:61-70`. The one counterweight, that
subagents lacking full context make conflicting implicit decisions
(https://cognition.com/blog/dont-build-multi-agents), bites on decision work rather than
reading work, and this pattern keeps decisions in the conductor (`research:72-81`).

**A fresh-context Opus reviewer over a Sonnet implementer.** The direction is the cleanest
number in the brief: stronger-reviews-weaker raised pass rate 71.6% to 89.7% while the
reverse direction lowered it 91.4% to 82.8% (https://arxiv.org/abs/2607.21656,
`research:161-169`). Reference-guided grading cuts a judge's *math-judging* failure rate
from 70% to 15% (https://arxiv.org/abs/2306.05685, `research:154-156`), and grading against
the task's acceptance criteria is the same mechanism applied to a different subject. cairn explicitly refused to downshift its reviewer to Sonnet
(`harvest:473-474`). Locally, the reviewer's findings are mostly substantive rather than
cosmetic in this repo's record (`research:20-28`), and the window's one real code defect,
a session epoch keyed on the raw rather than the normalized address, was caught by the
in-chain reviewer and fixed in one round (`measure:153`).

**Test-first with named mutations proved to fire.** Intrinsic self-correction frequently
fails and sometimes degrades performance (https://arxiv.org/abs/2310.01798), correction
works once the model validates against an external tool (https://arxiv.org/abs/2305.11738),
and forcing the kill-the-mutant check moves mutation score from about 66% to 93.57%
(https://arxiv.org/abs/2308.16557); all at `research:85-106`. The requirement already
stands at `dubplate-implementer.md:59-64`. Item 6 makes it legible; it does not weaken it.

**One re-dispatch on `fix`, with the second as the conductor's decision.** Self-repair gains
are modest and inconsistent once repair cost is counted (https://arxiv.org/abs/2306.09896),
and most iterative-refinement gains land in rounds one and two (`research:194-203`). The
local record says second rounds and escalations have paid, but only when the conductor
ruled rather than when the loop re-fired: in pass 4a-i two twice-fixed tasks closed real
defects and two escalations corrected the plan rather than the code
(`research:30-34, :215-219`). The rule already encodes that distinction exactly.

**The merge ritual.** Merge token, rebase, full gate in-tree, `--ff-only`, full gate on
master. The on-master gate exists because nothing in this repo is CI-enforced, so "master
is never left red" is otherwise aspirational
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:163-165`). cairn's
counter-evidence on parallel lanes is what makes the ritual load-bearing rather than
ceremonial: two concurrent worktree gates produce one-off flaky failures from shared-machine
contention, and a sweep could not build at all at load 7.4 (`harvest:264-274`). Item 14
changes one trigger inside the ritual and nothing else.

**The guards on long unattended work.** The runaway guard, the sleep inhibitor and the
battery watchdog stay (`~/.claude/CLAUDE.md`, "Multi-agent workflows"). The research brief
adds a reason beyond runaway containment: quality erosion appears in 77% of long-horizon
trajectories and gates do not detect it, because eroded code passes tests
(https://arxiv.org/abs/2603.24755, `research:365-379`). That is why Item 8 makes the
simplifier round non-optional after a long unattended run rather than optional.

**No check removed for cost.** Every measure in this document is a scheduling change, a
scoping change that preserves per-commit coverage, or an addition, and revision 2 is where
that became true. Revision 1 broke the invariant twice: Item 4 removed the close's
independent read, and Item 14 replaced a review trigger with a narrower one. Both now keep
what they had, and Item 1 adds a branch-history `gitleaks` scan the current gate does not
run at all. cairn holds the same
line: "No check is dropped for cost; the spec forecloses that" (`harvest:467-470`), and the
repo `CLAUDE.md` states it as "A check either gates or does not exist."

**Effort stays at `medium`, never lowered to `low`.** Fable 5.1 at `low` answers from memory
(`~/.claude/CLAUDE.md`, "Conducting a pass"; `harvest:471-472`). No item here proposes an
effort change.

**The domain reviewer fan-out at pass end.** It is what makes trimming per-task gates safe
(`harvest:432-435`). cairn's chassis-B2 fan-out returned four blocking findings that none
of four tasks' own per-task reviewers caught, because each read as correct in isolation and
only surfaced once the whole pass diff was graded together (`harvest:424-430`). For 5b-ii
that is the fresh-context `web-auth-security-reviewer` at pass end
(`5b-ii:3185-3186, :3279-3284`), which stays.

## 5. First pass to apply this

Revision 1 listed seven items as adoptable at 5b-ii's launch with no code change. Two of
them, Items 1 and 2 in their narrow forms, are not deliverable. `scripts/check.sh` takes no
arguments, its gitleaks leg is unconditional at `:128`, the conform sweep is a package
inside `make check` rather than a leg of the script, and
`.claude/agents/dubplate-implementer.md:3, :65, :99` all name `bash scripts/check.sh`
literally, so a dispatch-supplied gate string conflicts with the standing definition the
agent reads first. What survives is below, in four tiers. Revision 2.1 adds two entries to
tier A and changes no other tier.

**A. Adoptable at 5b-ii's launch with no code change.**

1. **Item 5, the pre-flight checklist.** One new section in
   `.claude/agents/dubplate-implementer.md`, before `:93`, with the two cairn-derived
   entries marked as adopted on shape rather than on a local cost. It adds checks and
   removes none.
2. **Item 10, Task 0 without a gate.** One line in Task 0's dispatch notes plus the
   conductor taking the baseline itself at launch, with the reviewer told to report the
   conductor's baseline line as its gate evidence so the task does not land in
   `needs-decision`, and with the thin-conductor exemption stated.
3. **Item 13, the ceiling check at every boundary.** 5b-ii already re-derives at the
   segment B boundary (`5b-ii:3337-3338`); extend the same check to A, C, D1 and D2.
4. **Item 4's grep half.** Task 12's acceptance criteria already require the resolution
   table to be produced by `grep -rn 'ruled, conductor 2026-09-12, R-5b'` rather than
   reconstructed (`5b-ii:2978-2983`), which is exactly this rule; extend it to every
   citation and every count in that task's output.
5. **Item 11, closing the session at an idle gap.** It is a behavior, not a change to
   anything.
6. **Item 15, the no-polling line in the implementer definition.** One rewritten step in
   `.claude/agents/dubplate-implementer.md:65-66`, replacing "until it ends with" with a
   single blocking `cairn-run-gate` call. It carries no quality risk in either direction.
7. **Item 16, the three-to-four-task segment and the green-commit boundary.** Ruled, and
   5b-ii's five segments over thirteen tasks already satisfy it, so what lands at launch is
   the recording of δ_eff, M and fault-detection latency at each boundary.

**B. Adoptable at launch with one edit to a shared workstation script.**

6. **Item 3, `commentOnly` plus the `reducedGate` argument in `pass-execute.js`.** Optional
   properties only, never in either schema's `required` list, and an argument defaulting to
   absent, so a consumer that does not opt in is unaffected. Smoke-tested on one dubplate
   task before any cairn pass runs on the edited runner.
7. **Item 6, the mutation ledger and the severity label.** Optional properties in
   IMPL_SCHEMA and REVIEW_SCHEMA (`pass-execute.js:37-82`) plus the matching lines in
   `diff-reviewer.md:52-65`. The mutation requirement already stands in prose at
   `dubplate-implementer.md:59-64, :102`, so this adds no obligation to the implementer.

**C. Adoptable at launch with one plan edit.**

8. **Item 8, half-step.** Cut 5b-ii's four simplifier rounds to two, after segment C and at
   D2 (`5b-ii:3236-3242`), with each round's review chunked by package territory. Captures
   roughly half of Item 8's 53,000 to 80,000.

**D. Blocked on one repo task, and Items 1 and 2 are blocked on the same one.**
`scripts/check.sh` needs a `--scope-changed` mode and a `--reduced` mode, plus the three
`.claude/agents/dubplate-implementer.md` edits that let a dispatch name the gate string.
That is roughly 40 to 60 lines of shellcheck-clean shell and one diff review. Item 4's
script half needs a second task, a `tools/` evidence-block emitter taking the session
directory as an argument. Neither is a launch-day decision, and Item 3's edit routes nothing
until the first one lands.

Not for 5b-ii's launch: Item 7 (its prep is done), Item 9 (precondition unverified),
Item 12 (its boundaries are ratified and each carries named work), Item 14 (single lane).

## 6. Scoring the next pass

The next pass measures itself against this baseline using the same unit and the same
method, so the two are comparable.

**The unit.** Deduplicate every `usage` object on `message.id` first, then apply
`(output_tokens + cache_creation_input_tokens) / 2` (`measure:12-16`). A naive sum over
`usage` objects double-to-quintuple-counts every multi-block turn (`measure:12-14`). Record
the scale check against the conductor's own spend ledger, and double the unit before
comparing: what reproduced the ledger within about 15 percent on the window is the undivided
`output + cache_creation`, that is twice this unit (`measure:18-22`). A unit-token total
compared against a live ledger reading without that doubling is out by a factor of two, in
the direction that reads as a 50 percent saving. Do not compare any figure to a cairn number
without first establishing whether cairn's method deduplicates (`measure:22-23`).

**Where the numbers come from.** The metric table below is emitted by Item 4's `tools/`
script, which takes the session directory as an argument rather than carrying it as a
constant. Three of its rows have no instrumentation today and need it first.
`~/.local/bin/cairn-run-gate` prints the gate's exit status and its last 60 lines and emits
no elapsed time, and `dubplate-implementer.md:97-105` requires the gate's final line and
`go test` wall time rather than the gate's own, so "Gate wall per pass", "Gate wall per
task" and "Full gate runs per task" are otherwise reconstructable only by walking
transcript timestamps around Bash calls. One elapsed-seconds line in `cairn-run-gate`'s
output and one field in the implementer's report format make them reported numbers
instead.

**What to record, and the baseline each is scored against.**

| Metric | Baseline | Source |
| --- | --- | --- |
| Chain tokens per code task | 123,793 | `measure:175-176` |
| Chain tokens per close task | 151,374 | `measure:176` |
| Close-and-ledger share of pass | 18.3% | `measure:96` |
| Review share of pass | 26.3% | `measure:95` |
| Simplifier share of pass | 8.8% over 4 rounds | `measure:88` |
| Conductor share, and tokens per conductor turn | 6.4%, 1,475 | `measure:90, :216-217` |
| Full gate runs per task | 2.5 | `measure:126` |
| Gate wall per task, median share | about 18% | `measure:125` |
| Gate wall per pass | 19.4m (8c), 29.3m (5b-i) | `measure:121` |
| Fix-round count per task | 6 events over 14 tasks | `measure:148` |
| Fix-round tokens by cause | 18% real defect, 11% vacuous test, 71% comment or citation or draft | `measure:161-165` |
| Prep share of pass, from seven recorded budgets | 22% to 49% | `research:36-40` |
| Prep tokens by component | draft, lenses, fold, each separately | `measure:104-106` |
| Segment boundary cost | 65,000 to 85,000 | `measure:205` |
| Idle-gap re-prime, if any occurs | 303,912 over 4 turns | `measure:110-111` |
| δ_eff, boundary cost in task-equivalents | 0.53 to 0.69 | `measure:205, :175-176`; `sizing:533-538` |
| M, tasks per fault | 2.3 counting every fix or escalate, 14 counting redirects only | `measure:148, :153, :161`; `sizing:539-542` |
| Fault-detection latency in tasks | not measured on the window | `sizing:543-549` |
| Who caught each fault: reviewer, gate, or conductor | not measured on the window | `sizing:550-556` |

**The fix-round cause classification is the most important single metric.** It is what
tells the next pass whether Items 4, 5 and 6 worked, because all three attack the same 71
percent. Classify every fix and escalate event into the six classes the window established:
real code defect, vacuous or missing test evidence, comment-only, citation pointer,
draft content, and plan-infeasible outcome (`measure:150-158`). A pass whose fix rounds
shift toward real defects has succeeded even if its total fix spend is unchanged, because
the spend is then buying something. The classification is a conductor-dispatched read at
pass close carrying its own budget line in Item 4's script, not an unowned side task
competing with the close for the conductor's attention at the moment the close is being
trimmed.

**Two guards on the scoring itself.** A saving claimed here is not a saving observed; every
range in section 3 carries its derivation so the next pass can check the arithmetic against
what it actually paid. And a quality regression would show first as a defect caught late,
at the pass-end fan-out or at the merge rather than at the task, so record where each
blocking finding was caught, not only that it was.
