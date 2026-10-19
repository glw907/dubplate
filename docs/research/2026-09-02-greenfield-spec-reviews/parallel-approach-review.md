# Adversarial review: the parallel build approach

Status: review, 2026-09-03. Fresh context, hostile by default. Target:
`specs/2026-09-03-parallel-build-approach.md` (ratified). Evidence base
audited: `research/2026-09-03-rung-dependency-analysis.md`. Governing
document the amendment edits: `specs/2026-09-02-greenfield-server-build.md`
(revision 2.1).

Decisions 1 and 2 are Geoff's and are not re-litigated here. This review
attacks the mechanism that carries them: the dependency graph the lanes
are cut from, the merge ritual, the worktree mechanics against this
repo as it sits today, Phase 0's sufficiency, the conductor-capacity
claim against this session's own record, and the fidelity of section 5's
amendment text.

Counts: 4 blockers, 13 majors, 8 minors, 10 verified-sound items.

Measurements taken during this review (workstation, 2026-09-03):
`gitleaks detect --no-git -s .` scans 69.30 MB in 1m0s; `corpus/`
carries 62 MB across 585 tracked files at 146 cases; `tools/venv/.venv/`
is 78 MB and gitignored; `make derive` is documented at ~585s and
carries a 90m timeout.

## Blockers

### B1. The dependency matrix is one-directional, and the proposal's value comes entirely from reversing edges it never classified

The matrix header states the rule: "Rows are the predecessor, columns
the successor. Read a cell as 'does the column rung depend on the row
rung'." Every row is upper-triangular. Each rung pair is therefore
classified exactly once, in ladder order, asking only whether the LATER
rung depends on the EARLIER one.

Every re-sequencing the proposal buys its throughput from inverts a
pair: 5 and 8 ahead of 3 and 4, 14 ahead of 7, 9, 12, and 13, and 10,
11, 13 concurrent. For each of those the load-bearing question is the
one the matrix did not ask.

The concrete instance that bites inside the proposal's own schedule is
mail. The matrix classifies 7 -> 9 as SOFT (the intake-to-pipeline
handoff record and the journey vocabulary). It never asks whether rung 7
depends on rung 9. It does: rung 7 ships the reject copy contract whose
strings "the uploader page and emails share verbatim" (repo CLAUDE.md,
and the build spec's rung 7 entry), and the only mail seam in the tree
shape is `internal/notify/`, assigned to rung 9 ("the notifications and
monitoring seams as optional integrations, built immediately ahead of
the pipeline that consumes them"). Phase 2 of the proposal runs "V pairs
with I, then P with whichever remains", so lanes I (rung 7) and P (rung
9) are adjacent or concurrent, and `internal/notify` appears only in
lane P's territory column. Two lanes, one unassigned surface, and the
contested-surface ownership table in section 3 does not list it.

Rung 5 makes it three consumers (B4).

This is the catastrophic class the review was asked to hunt: a
claimed-NONE cell that is a real edge, discovered only because the
schedule inverts it.

**Fix.** Before either phase-1 lane launches, complete the matrix's
lower triangle for exactly the pairs the proposal reorders: (5,2), (5,3),
(5,4), (5,6), (5,7), (5,9), (8,3), (8,4), (14,7), (14,9), (14,12),
(14,13), (12,13), (12,10), (12,11). Fifteen cells, one research
dispatch. Every cell that comes back HARD or SOFT is either a new
ownership-table row or a lane-ordering constraint. Additionally, assign
`internal/notify` explicitly. The cheapest assignment is to rung 5,
which is the first rung that must send mail and is already in phase 1;
rung 7 and rung 9 then consume the seam rather than inventing it.

### B2. Phase 0's serve mount seam is defined on both sides of the gate it is supposed to open

Section 1 says two things about the same artifact:

> **Serve mount seam.** Rung 5's spec gains a `Mount(prefix,
> http.Handler)` requirement. Rung 5 owns the shell, the mux, and the
> ingress strip.

and

> Both are prerequisites for a second Go lane.

Ratification carries the second reading: "Phase 0 (the config wall-row
registry and the serve mount seam) as the second lane's prerequisite."
If the seam is a prerequisite for the second code lane, and rung 5 is
itself lane A in phase 1, then the seam must be built before rung 5's
spec, plan, and full fresh-context adversarial cycle. Rung 5 is one of
the three rungs the build spec holds to that cycle. Building its serve
shell as an unplanned "serial, small, not a rung" item strips the
adversarial review from the ingress strip, which is the single artifact
the whole trust boundary rests on ("Serve strips every inbound
identity-header spelling on ingress, before routing, for every path").

The other reading, that Phase 0 contains only the SPEC AMENDMENT and the
seam's code lands with rung 5, is coherent for phase 2 and makes the
sentence "Both are prerequisites for a second Go lane" false for phase 1.

The seam also does not do the job its own justification claims. Section
1 motivates it with "Without this, rungs 6, 7, 12, and 14 all edit
`internal/serve`". `Mount(prefix, http.Handler)` cannot express what
rung 12 needs. Rung 12 is the ingress reverse proxy, and by the one
owned prefix rule the streamer owns "everything else", so rung 12 is the
root mux's fallthrough rather than a mounted subtree. Its birth
requirements are all shell-level: header-strip set driven by the
driver's configured header name, "path normalization rejects
raw/decoded mismatches and `..` segments before prefix routing", and
"per-route write-timeout handling for SSE and audio streams". None of
those are reachable through a per-consumer `Mount` call.

The proposal's own phase 3 accidentally saves this, because rungs 10,
11, and 13 do not write `internal/serve`. The justification is still
unsound, and it becomes a live blocker the moment the schedule shifts.

**Fix.** Split Phase 0 into what it can actually be. The config
wall-row work becomes rung 1b (S band) with a plan and the standard
chain, because it changes rung 1's ratified config-core contract. The
serve seam becomes a REQUIREMENT recorded in rung 5's spec plus a
scheduling rule with no artifact: no two lanes write `internal/serve`
until rung 5 closes. Rung 5's spec requirement must name the root-level
composition point (ingress strip, then path normalization, then prefix
routing, then fallthrough) and declare the extension point the driver's
proxy occupies, or state plainly that rung 12 owns the shell after rung
5 and never runs concurrently with another serve-writing lane.

### B3. A lane's accepted work can be invalidated by another lane's merge, and nothing re-reviews it

Section 3's merge ritual is: the lane's gate goes green in its own tree,
the lane rebases onto master, re-runs `bash scripts/check.sh`, and a
trivially additive merge lands.

Two holes.

First, semantic conflict with no textual conflict. Lane A's
`diff-reviewer` accepted a diff against a tree state that lane B's merge
then changes. A config default lane B moved, a registry semantic lane B
altered, an interface lane B widened: `git rebase` exits zero, the gate
may still pass, and the reviewer verdict that accepted lane A's work was
rendered against a tree that no longer exists. The proposal states no
re-review trigger anywhere. Section 4 makes this worse by treating
reviewer isolation as a safeguard: "Review independence is automatic.
Every `diff-reviewer` dispatch is a fresh zero-context agent, so lanes
cannot correlate through a shared reviewer context." A fresh reviewer
inside lane B's worktree cannot see lane A's landed work. Here isolation
is the mechanism that HIDES the cross-lane defect.

Second, a time-of-check to time-of-use gap. The post-rebase gate runs in
the lane's own tree. The merge into master is a later, separate act. If
the other lane merges in between, master's post-merge state was never
gated by anything, and the "Master is never left red" invariant is
aspirational. No CI enforces it in this repo.

**Fix.** Three additions to the ritual. (1) Serialize merges: exactly
one lane holds the merge token at a time, and a lane that cannot take it
waits. (2) After the rebase, merge with `--ff-only` and re-run the full
gate ON MASTER, not only in the lane tree. (3) Any rebase that is not a
fast-forward triggers one `diff-reviewer` dispatch on the merged result
against BOTH lanes' acceptance criteria before the merge is announced
done. A fast-forward rebase, where the other lane landed nothing, skips
step 3 mechanically and costs nothing.

### B4. "Rung 5 has zero inbound beyond the discharged rung 1" is false in two places, and phase 1's whole reshape rests on it

The analysis states it twice ("Only rung 1, which is discharged", "This
is the single largest piece of ordering-by-habit in the ladder") and the
proposal repeats it. Two inbound edges exist.

**Mail.** Rung 5's deliverables include "SMTP send". The tree shape
assigns `internal/notify/` ("optional mail seam") and the build spec
assigns the notifications seam to rung 9. See B1.

**The design language.** Rung 5 ships "the login and consumption pages
(Go-templated, no JavaScript, shared design language)". The theming
ruling in rung 7 is explicit about its reach: "binding here and on every
later web surface: theming is data, never code. Every component is
authored against DaisyUI theme variables from day one (no hardcoded
colors)". It also fixes the customization API as the theme-variable
contract, the two-axis match/palette composition model, and the font
delivery mechanism ("fonts ship embedded (woff2 via `go:embed`, never a
CDN)"). Rung 7 is also where "the `web/` legs rejoin the repo gate".

Moving rung 5 into phase 1 makes its login pages the FIRST rendered
user-path surface in the entire build, authored before the theme
contract that the ratified ruling binds every surface to, in a tree
whose web gate legs do not yet exist. The outcomes are unstyled pages
that rung 7 re-authors, or rung 5 inventing the theme contract without
rung 7's brief. Both cost the thing the build spends attention on, since
login is a rendered user-path page under the standing
`see-it-before-geoff-tests-it` bar and is one of the three things Geoff
verifies at the rung 16 sitting.

The analysis's 5 -> 7 discussion covers the roster loader and the
screenshot timing. It does not cover the direction where rung 7 is the
producer.

**Fix.** Either (a) pull the theme-variable contract and the embedded
font delivery out of rung 7 into Phase 0 as a small documented standard
(the theme-collection brief already exists as a standing input), so rung
5 authors against it, or (b) keep rung 5 in phase 1 but scope its pages
to semantic markup with no styling decisions, and record explicitly that
rung 7 owns their visual authoring and their before/after. Option (a) is
better, because it also fixes the same problem for rung 12's driver
pairing.

## Majors

### M1. The analysis's own totals do not reconcile with its own matrix

Stated: "34 hard edges, 21 soft, 50 none", summing to 105.

Counting the matrix as printed gives 120 cells (15 + 14 + ... + 1) over
rungs 2 through 17. Excluding row 2 gives 105, which is the population
the totals describe. Over those 105 cells the counts are 36 hard, 19
soft, 50 none. Independent corroboration for the soft figure: the "soft
edges, with their integration costs" section enumerates exactly 19 edges
(3->8, 3->9, 3->14, 9->14, 12->14, 13->14, 7->14, 4->17, 5->6, 5->7,
5->13, 6->12, 6->16, 7->9, 7->12, 7->13, 9->12, 10->15, 11->12).

Two edges are being reported as SOFT that the matrix marks HARD. Because
34 + 21 + 50 lands exactly on 105, the totals read as reconciled to the
population rather than counted from the cells. The headline claim ("half
the ladder's sequencing is adjacency", 50/105) survives. The hard/soft
split does not, and hard/soft is the axis every lane cut is made on.

**Fix.** Recount the matrix cell by cell, publish the per-row tallies,
and name the two edges the prose and the totals disagree about.

### M2. Rung 8's inbound is not "already shipped"

The proposal: "Rung 8's only inbound is rung 2's expected applied tag
sets, already shipped." The analysis, on the same edge: "The corpus's
expected applied tag sets. That record is rung 8's ONLY acceptance
evidence."

The corpus stands at 146 cases against a 150 to 300 target, with the
real-catalog stratum (50% of the target) still landing through 2c and
2d. The analysis makes a careful argument for rung 3 that the dependency
is on rung 2's machinery and case FORMAT rather than case COUNT. That
argument does not transfer to rung 8. Rung 3's deliverable is penalty
components with their own unit tests; rung 8's deliverable is a FLAC
writer whose only proof is per-case applied-field agreement. The 61
structural synthetic cases shipped in 2b are precisely the stratum least
likely to exercise real tag diversity.

**Fix.** Either schedule rung 8 after 2d closes, or state in its plan
that its acceptance is provisional against the corpus at its then-current
size and that rung 15 re-runs it against the complete corpus as a
launch-bar item. The second is cheap and honest.

### M3. The gate is mispriced, and the shared tree is coupled beyond compilation

Section 3's central claim: "`scripts/check.sh` runs the module gate over
`./...`, so a second Go lane's in-progress package reds the first lane's
gate even when the file sets are disjoint. That is the whole difference,
and it is why the second code lane gets its own tree."

`scripts/check.sh` today also runs two whole-repo legs the analysis and
the proposal both omit:

- `ruff check` and `ruff format --check` over `git ls-files '*.py'`,
  repo-wide tracked Python.
- `gitleaks detect --no-git -s .`, which scans the entire worktree
  including untracked non-ignored files.

Measured on this workstation: the gitleaks leg scans 69.30 MB in 1m0s
today, and `corpus/` is 62 MB of that at 146 cases. At the 300-case
target the leg roughly doubles. That cost is paid by every lane, on
every task in the chain, and again on every re-dispatch and every
post-rebase re-run.

The coupling consequence is worse than the cost. Corpus cases are
"verbatim wire data from a live service" (check.sh's own comment), the
repo already carries one gitleaks false positive in `.gitleaksignore`,
and gitleaks reads the whole tree. In the shared tree the E1 corpus lane
and a code lane therefore DO couple: a new case whose response body
trips a gitleaks rule reds the code lane's gate. "Document lanes do not
compile, and that is the whole difference" is not true of this gate.

**Fix.** Three parts. (1) Correct the mechanism sentence in section 3 to
name all three whole-repo legs. (2) Give the chain a fast gate and a full
gate, with gitleaks and ruff on the full gate at task close and the
module gate per implementer iteration; check.sh keeps one entry point
and takes a scope flag, so the rule that a check either gates or does not
exist is preserved. (3) Put the corpus lane and a code lane in separate
trees too, or accept and document the coupling.

### M4. `make tidy-check` writes to the shared tree

`tidy-check` runs `go mod tidy` and then `git diff --exit-code -- go.mod
go.sum`. The first half MUTATES two tracked files. Two gate runs in one
shared tree write the same two files concurrently, and a code lane can
observe a `go.mod` mid-rewrite. Every lane's chain runs the full gate,
document lanes included, so this is not confined to Go-on-Go pairs.

The analysis's option (b) treats go.mod contention as self-healing
("`tidy-check` is in the gate, so a post-merge `go mod tidy` resolves it
mechanically"). That covers merge conflicts. It does not cover a gate leg
with a write side effect running twice at once.

**Fix.** Make `tidy-check` non-mutating: run `go mod tidy` against a
scratch copy, or take a repo-level gate lock so only one `check.sh` runs
per tree at a time. The lock is one line and also fixes concurrent
`golangci-lint` cache contention.

### M5. A worktree does not carry the venv, and the gap is silent

`/tools/venv/.venv/` is gitignored (78 MB). `git worktree add` checks out
tracked files only, so a second lane's tree has `tools/venv/README.md`
and `requirements.txt` and no venv.

`make derive` is the oracle-tagged differential run, documented at ~585s
and carrying a 90m timeout, and it is DELIBERATELY outside `make check`
("the gate protects the frozen corpus records and never touches
Python"). So a worktree lane's gate goes green while the only leg that
would have exercised the beets oracle never ran and never complained.
Any lane that touches `corpus/`, `internal/harness`, or `cmd/harness`
from a worktree is affected.

Each worktree also checks out 62 MB of corpus that most lanes have no use
for, and pays the M3 scan on it.

**Fix.** Add venv provisioning to the lane-launch checklist (`uv sync`
against `tools/venv/requirements.txt`), and make the checklist state
which lanes need it. Better: state as policy that corpus and harness work
never moves to a worktree, so lane E1 stays in the primary tree, which
is what the proposal already schedules.

### M6. The mandated live-executor check breaks under two trees

Both CLAUDE.md files require it. Global: "verify no live executor is
already working it: `pgrep -f <worktree path>`". Repo: "Before any
executor touches this repo or the box: check for live executors (`pgrep
-f Projects/dubplate`)". Git worktrees live under `.claude/worktrees/`
INSIDE the repo, so once a second lane exists `pgrep -f
Projects/dubplate` matches both lanes, permanently and by design.

A safety rail that returns a false positive every time is a safety rail
that gets ignored. This one exists because two workflows once raced one
worktree and burned ~1.2M duplicated tokens.

**Fix.** Replace the check with a lane registry: each lane launch writes
its name, worktree path, and pid to a known file, and the check becomes
"no other executor in THIS worktree". Put the registry write and the
teardown on the lane-launch checklist next to the runaway guard.

### M7. Repo CLAUDE.md still forbids the model, and it is the file every subagent loads

The amendment edits the spec. Repo CLAUDE.md's "Build process" section
still reads:

> The build runs in session-scale ARCS, never one session per rung
> ... **The conducting loop, per rung:** ... prep the next rung's plan
> in a workflow only against a quiet tree (never launch agents into a
> mid-execution tree)

The parenthetical is the exact rule the parallel model violates by
design. The spec's own arc paragraph names CLAUDE.md as authoritative:
"CLAUDE.md's 'Build process' section carries the operational form every
session and subagent loads." Every implementer, reviewer, and merge agent
dispatched into a lane will read a file that forbids what it was
dispatched to do.

**Fix.** Fold CLAUDE.md's Build process section in the same commit as the
spec amendment. It needs the phase re-key, the lane vocabulary, the
lane-launch checklist (guard, registry, venv, ceiling), and a
replacement for the quiet-tree rule that says what actually holds: never
launch a second executor into the SAME tree.

### M8. Section 5's replacement drops five things the arc paragraph carried

Comparing the original Process arc paragraph against the replacement,
these are present in the original and absent from the replacement and
from section 6:

1. Plans are "drafted and ADVERSARIALLY REVIEWED in the shadow of the
   previous rung's execution". The replacement keeps only "drafted".
   This review exists because of the dropped clause, and plan-level
   adversarial review is the instrument that found the staleness defect
   class the same sentence cites.
2. "Workflow mode is a standing grant for execution and prep across the
   whole build."
3. "with human questions batched".
4. "and unblocked work continuing when the owner is unavailable."
5. "CLAUDE.md's 'Build process' section carries the operational form
   every session and subagent loads."

Items 2 through 4 are Geoff's standing grants of 2026-09-02. They survive
in CLAUDE.md, and the spec paragraph is the canonical record CLAUDE.md
points back to, so dropping them here breaks the loop in both directions.

**Fix.** Restore all five into the replacement text. The adversarial
review clause needs one adaptation for lanes: a lane's next plan is
drafted AND adversarially reviewed in the shadow of that lane's own
execution.

### M9. Two ceilings now contradict each other

Section 4: "Each phase carries one combined ceiling across its lanes, not
a ceiling per lane."

The ladder preamble, which section 5 amends only in its first sentence,
still says "each plan header carries a token ceiling and checkpoint
interval". Section 6 keeps "Both budgets, scored at every pass close
against a per-phase ceiling."

A lane running a workflow cannot see the other lanes' spend, so the
global 80%-of-ceiling stop rule ("At 80% of the ceiling, finish the task,
write STATUS, and ask one combined question") has nothing to evaluate
itself against.

**Fix.** Give each lane a sub-ceiling that sums to the phase ceiling, and
have the conductor reallocate at checkpoints. Then the plan-header rule
and the phase rule both hold.

### M10. The re-keyed evidence phase is larger than any arc the process has run

The replacement text: "evidence (2, 3, 4, with lanes A and T carrying
rungs 5 and 8 alongside), closing at rung 4's deploy ... One conducting
session per phase, no clears or re-brainstorms inside one."

That phase now carries three L-band rungs (2, 3, 5) plus two M-band (4,
8). Rung 2 alone has consumed 2a, 2b (~5.8M), and 2c (8M ceiling) so
far, and 2d is planned. Adding rungs 3, 5, 4, and 8 to the same session
with no clears is a large bet on compaction holding across the whole
thing.

**Fix.** Either allow a phase to span more than one session with a
written lane handoff, or split the evidence phase at rung 4's deploy
into evidence-critical (2, 3, 4) and evidence-parallel (5, 8) and accept
two sessions.

### M11. Rung 14 before rung 9 freezes a contract while an owner ruling is still open

The proposal moves V (rung 14) into phase 2 ahead of I and P: "V moves
early because its inbound edges from the producing rungs are soft by the
additive-within-version policy, so landing the contract early gives every
later rung a fixed target to project into."

The additive policy makes later VERBS cheap. It does not make later
CHANGES TO EXISTING TYPES cheap. Rung 9 carries "the concurrency model
decision (per-album lease or explicit queue, with the constraint that
interactive work must not starve timer imports), settled in this rung's
spec", and it is one of the two owner touchpoints the build spec names
ahead of time. That decision shapes the ops job model, and rung 14's wire
types are "a published projection of the interface" (the analysis's own
6 -> 14 justification). If the ruling changes a job's states or
preemption semantics, `/api/v1` either takes a breaking change before it
has a single consumer or carries a documented breach into launch.

**Fix.** Take rung 9's concurrency ruling at phase 2's opening
checkpoint, before lane V drafts its plan. It is already a known-ahead
owner ruling, so this costs no new attended moment. Alternatively mark
the contract document provisional until rung 9 closes, and say so in the
contract itself.

### M12. The box contention list is incomplete

Section 2: "The box. Rung 4's additive unit and rung 10's read-only
listing are compatible. Rung 16 needs a quiet box and overlaps neither."

Rung 12 also reaches the box. Its ladder entry: "set Navidrome's default
UI theme to its built-in Gruvbox Dark via config", plus
"the externalized-auth config names verified against the pinned
Navidrome version" and "provisioning choreography with the first-admin
`doctor` preflight". The build spec's own box-safety section contradicts
this by claiming "Rungs 1 through 3, 5 through 9, and 11 through 15 touch
the box not at all", which puts rung 12 in the never-touches set.

Phase 3 runs rung 12 alongside rung 10, and rung 10's deliverable is the
first read-only no-move diff against the LIVE library. The blast radius
is small. The defect is that the proposal enumerates box contention as a
named safety item and the enumeration is demonstrably incomplete on the
one production resource in the build.

**Fix.** Reconcile the two lists. Enumerate rung 12's box writes
explicitly, decide whether the Navidrome theme and externalized-auth
config land at rung 12 or defer to rung 16, and if they land at 12,
require the quiet-box check and forbid overlap with rung 10's diff
window.

### M13. This session's record is cited as proof and is partly a counterexample

Section 3: "Shared-tree concurrency is already proven in this session,
but only for document lanes."

What the record actually shows for 2026-09-03:

- STATUS: "BATTERY STAND-DOWN (2026-09-03 ~13:10): the 11% watchdog
  fired mid-quad-3; the workflow was stopped cleanly." A three-lane day
  exhausted the host's power budget and stopped a lane mid-quad.
- STATUS: "Batch 11's implementer was killed mid-run (its half-written
  README edit reverted; tree clean)." A mid-run kill left a dirty tree
  that someone had to inspect and revert. That is the conductor doing
  exactly the work the thin rule forbids, in the shared tree, under the
  proven-safe mode.
- Commit timeline: between 10:34 and 11:03 the conductor landed 12 doc
  commits (theming and format rulings). Over the same window the corpus
  lane's batch cadence fell from roughly one per 25 minutes (05:58 to
  07:21, four batches) to roughly one per 50 minutes (10:19, 11:17,
  12:09). The correlation is not proof of causation, and it is the only
  throughput measurement the session offers, and the proposal offers
  none.

The unnamed contended resource is the workstation itself: power, and CPU
for concurrent `golangci-lint`, `go test ./...`, a 60-second gitleaks
scan, and a 585-second `make derive`. The doctrine requires naming the
contended resource when serializing. The converse obligation, when
parallelizing, is showing the resource is not contended. Today it was.

**Fix.** Add to the lane-launch checklist: AC power required for any
multi-lane session, and a stated cap on concurrent gate runs (the M4
gate lock delivers this for free). Record a throughput baseline per lane
so the next multi-lane phase can be scored rather than asserted.

## Minors

**m1.** The analysis's 5 -> 14 justification says "non-browser token
acquisition on the session vocabulary" is a rung 5 surface. The build
spec's supersession list says the opposite: "the seam itself is born at
rung 5; non-browser token acquisition stays at rung 14."

**m2.** Both documents say rungs 5 through 14 are "roughly four L-band
and six M-band passes". The ladder bands are 3 L (rungs 5, 9, 12) and 7 M
(6, 7, 8, 10, 11, 13, 14). The sentence also conflates rungs with passes,
since an L rung splits into several.

**m3.** "classified all 105 ordered rung pairs" is the phrasing that
makes B1 easy to miss. The matrix classifies each UNORDERED pair once,
in ladder order. Say so.

**m4.** The merge ritual never states who decides "trivially additive"
versus "non-trivial conflict". Make it mechanical so no judgment is
required: a rebase exiting nonzero dispatches the merge agent, and the
conductor never inspects.

**m5.** `.gitleaksignore` pins its single allowed finding by line number:
`internal/config/keys_test.go:generic-api-key:66`. Phase 0's wall-row
work edits that file. Any insertion above line 66 shifts the fingerprint,
the ignore stops matching, and the gate reds for every lane at once.
Re-fingerprint as part of the Phase 0 task, and note the fragility in
the file.

**m6.** Section 4 moves the domain fan-out to each lane close rather than
each phase. That multiplies reviewer dispatches by the lane count.
Section 7's cost list ("Merge overhead: one rebase, one full gate run,
and possibly one merge-agent dispatch per lane close") omits it, and
section 4's "Parallel lanes cost the same tokens" is then not quite
true.

**m7.** The compaction instruction is single-lane ("Preserve the plan
path and pass number; the task ledger"). STATUS is also single-lane in
shape and already well past the 60-line cap, with a workflow resume path
pasted into it. N lanes need a defined per-lane ledger block, or STATUS
grows N times and is re-read in full at every session start.

**m8.** "Review independence is automatic" is stated as a safeguard in
section 4. Under B3 it is the opposite. Reword it to say what is true:
reviewer contexts cannot correlate, AND a lane's reviewer is structurally
blind to the other lane's landed work, which is why the merge ritual owes
a re-review step.

## Verified sound

Checked against the specs cited, and correct.

1. **The three adversarial rungs really are serial by dependency.**
   5 -> 12 is hard (session cookie, verify-identity seam, ingress shell,
   header-strip set, and the explicit ingress-before-ExtAuth ordering),
   12 -> 16 is hard, 5 -> 16 is hard. The dismissal of concurrency-form
   correlated-reviewer risk is sound, and the residual across-time risk
   is correctly identified and already mitigated by rung 12's
   fresh-context mandate.

2. **10 -> 17 as a latest-bound gate is real and unviolated.** Rung 10
   captures the legacy ledger format "while `_salvage/` is legible";
   rung 17 retires `_salvage/`. No proposed reordering pulls 17 forward.
   The edge kind is correctly distinguished from a consumption edge.

3. **The rung 7 attended-visual hold is correctly reasoned, and the
   proposal's own schedule makes it moot.** Rung 5 moves to phase 1 and
   rung 7 sits in phase 2, so rung 7 never builds ahead of rung 5 under
   the proposal. Keeping the rule written is still right, because it is
   the constraint that survives a schedule change.

4. **The MusicBrainz budget is correctly named as the one contended
   resource file ownership cannot fix.** Verified against all three
   claimed drawers: rung 2's capture tool with "the client's hard 0.25s
   inter-request floor", rung 3's shared limiter with its per-caller
   budget interface, and rung 4's "deferred background MusicBrainz
   budget through the shared rate limiter".

5. **The `./...` argument for a separate tree per Go lane is correct.**
   Verified in the Makefile: `check: vet lint test tidy-check`, and vet,
   lint, and test all run `./...`. A half-landed package in one lane does
   red the other lane's gate in a shared tree.

6. **3 -> 4 reads the edge correctly.** Rung 4 needs a matcher that
   produces verdicts, not one that has cleared the bar; the bar is a
   launch-gate item evaluated at rung 15. The build spec supports this,
   and the mitigation (stamp every divergence with the matcher commit so
   superseded revisions re-evaluate automatically) is the right shape.

7. **6 -> 14 is correctly hard.** "The wire types are a published
   projection of the interface, so the API cannot be authored before the
   interface exists" matches rung 6's deliverable ("one Go interface plus
   JSON-serializable request and response types") and rung 14's.

8. **15 -> 16 -> 17 refuses manufactured parallelism, correctly.** All
   three edges are hard, and 16 -> 17 additionally carries the
   deprecation-window calendar gate.

9. **The honest headline is honest.** "The shadow floor sets the launch
   date, so spine concurrency buys slack, not schedule" is correct and it
   reduces the proposal's own value claim. Section 7 then routes the
   slack to quality rather than to more concurrency, which is the right
   call and is the strongest paragraph in the document.

10. **Decision 3's ratified reframing follows from the graph.** The only
    calendar consumers of the shadow record are rungs 15 and 16 (4 -> 15,
    4 -> 16), nothing between 5 and 14 is gated on it, and the shadow
    accumulates from rung 4 onward. So any plausible floor is satisfied
    before the spine completes, and the floor is an ordinary config
    choice under the defaults discipline.

## What to do before phase 1 launches

Ordered by what blocks what.

1. Complete the 15 reverse-direction matrix cells (B1). One research
   dispatch.
2. Assign `internal/notify` and the theme-variable contract in the
   contested-surface table (B1, B4).
3. Resolve Phase 0's shape: config work becomes rung 1b with a plan; the
   serve seam becomes a rung 5 spec requirement plus a scheduling rule
   (B2).
4. Rewrite the merge ritual with the merge token, the on-master gate, and
   the post-merge re-review trigger (B3).
5. Correct the gate mechanism sentence and add the gate lock and the
   fast/full split (M3, M4).
6. Write the lane-launch checklist: runaway guard, lane registry, venv,
   sub-ceiling, AC power (M5, M6, M9, M13).
7. Fold CLAUDE.md's Build process section together with the spec
   amendment, and restore the five dropped clauses to section 5's
   replacement text (M7, M8).
8. Recount the matrix totals and publish per-row tallies (M1).
