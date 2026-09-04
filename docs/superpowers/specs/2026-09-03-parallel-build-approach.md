# Parallel build approach: a proposed revision to the execution model

Status: PROPOSAL, **revision 2 (folded)**, 2026-09-03. Revision 1 was ratified by Geoff on 2026-09-03 (decisions 1 and 2 as proposed, decision 3 reframed); revision 2 folds the adversarial review at `research/2026-09-02-greenfield-spec-reviews/parallel-approach-review.md`. Amends the execution model of the ratified build spec (`2026-09-02-greenfield-server-build.md`, revision 2.1). It changes no rung's deliverables, no acceptance criterion, and no launch bar item.

## Fold note (revision 2, 2026-09-03)

The adversarial review returned 4 blockers, 13 majors, 8 minors, and 10 verified-sound items. All 25 findings are folded. Three took an adaptation rather than the literal fix, each named below; nothing was rejected.

What changed, by weight:

- **The dependency graph is no longer trusted in one direction.** The matrix classified each rung pair exactly once, in ladder order, and every re-sequencing this proposal buys throughput from inverts a pair the matrix never asked about (B1). Completing fifteen named reverse-direction cells is now a blocking prerequisite in Phase 0, ahead of any phase-1 lane launch.
- **Phase 0 split into what it can actually be** (B2). The config wall-row registry becomes **rung 1b** (S band, its own plan, the standard chain), because it changes rung 1's ratified config-core contract. The serve mount seam becomes a requirement recorded in rung 5's spec plus a scheduling rule with no artifact, because `Mount(prefix, http.Handler)` cannot express what rung 12's ingress shell needs and the seam cannot both precede rung 5 and be built by it.
- **The merge ritual gained a merge token, an on-master gate, and a re-review trigger** (B3). A lane's accepted work could previously be invalidated by another lane's merge with nothing re-reviewing it, and reviewer isolation was the mechanism hiding the defect rather than a safeguard.
- **Rung 5's inbound edges are not zero** (B4). Mail and the design language both flow into it. The mail seam is now assigned; the theme-variable contract and the embedded font delivery are promoted into Phase 0 as a documented standard so rung 5's login pages are authored against them.
- **The gate is repriced** (M3, M4). `scripts/check.sh` runs three whole-repo legs, not one: the `./...` module gate, `ruff` over tracked Python, and `gitleaks` over the entire worktree (69.30 MB in 1m0s today, `corpus/` 62 MB of it). Document lanes and code lanes therefore do couple in a shared tree. The chain gets a fast/full split behind a scope flag, and `check.sh` takes a repo-level lock because `tidy-check` mutates `go.mod`.
- **A lane-launch checklist now exists** (M5, M6, M9, M13): runaway guard, lane registry, venv provisioning, sub-ceiling, AC power, gate-lock cap, throughput baseline.
- **CLAUDE.md is folded in the same commit as the spec amendment** (M7). Its "Build process" section still forbids by name the thing the parallel model does by design, and it is the file every subagent loads. Exact replacement text is in section 5.
- **Section 5's replacement text restored five clauses it had dropped** (M8) and now allows a phase to span more than one session with a written per-lane handoff (M10).

Adaptations, stated plainly:

1. **M1 (matrix totals).** The review asked to name the two edges the prose and the totals disagree about. They cannot be named: the prose's soft-edge enumeration matches the matrix cell for cell at 19 edges, so the totals were reconciled to a population rather than counted, and no individual edge is in dispute. The review also inferred the population was 105 by excluding row 2; row 2 is a legitimate predecessor row carrying five hard edges, so the true population is the 120 unordered pairs among rungs 2 through 17. Recounted per-row tallies are published below. The headline improves: 60 of 120 is exactly half.
2. **M2 (rung 8's provisional acceptance).** The review's second option routes rung 8's full-corpus re-run to "a launch-bar item". The launch bar is ratified at four parts and this proposal changes none of them, so the re-run lands as a rung 15 acceptance criterion instead. Same effect, no launch-bar edit.
3. **B4 (design language).** Option (a) adopted, with one clause of option (b) kept alongside: the theme-variable contract and font delivery move into Phase 0, **and** it is recorded explicitly that rung 7 still owns the visual authoring of rung 5's pages and their before/after. Belt and braces, both cheap.

Two findings produce corrections the build spec owes but that this fold does not make, because the amendment's scope is the ladder preamble and the Process arc paragraph only. Both are recorded here as errata for the spec's next revision: the box-safety section's claim that rungs "5 through 9, and 11 through 15 touch the box not at all" is false for rung 12 (M12), and the supersession list's ruling on rung 5 versus rung 14 token acquisition contradicts the dependency analysis's 5 → 14 justification, with the spec winning (m1).

## Foundation

`research/2026-09-03-rung-dependency-analysis.md` classifies each **unordered** rung pair among rungs 2 through 17 exactly once, in ladder order, asking only whether the later rung depends on the earlier one. That is 120 pairs, not 105, and the direction it does not ask is the subject of B1 below.

Recounted from the matrix cell by cell: **41 hard edges, 19 soft, 60 none.** The published totals (34 / 21 / 50) do not reconcile with the cells and are superseded. Per-row tallies, hard / soft / none:

| Row | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| H | 5 | 3 | 2 | 4 | 7 | 2 | 3 | 5 | 2 | 1 | 2 | 1 | 1 | 2 | 1 |
| S | 0 | 3 | 1 | 3 | 2 | 4 | 0 | 2 | 1 | 1 | 1 | 1 | 0 | 0 | 0 |
| — | 10 | 8 | 10 | 5 | 2 | 4 | 6 | 1 | 4 | 4 | 2 | 2 | 2 | 0 | 0 |

Exactly half the ladder's sequencing is adjacency from authoring for one executor, not dependency. The hard/soft split moved, and hard/soft is the axis every lane cut is made on, so section 1's lane cuts are stated against the recount rather than against the published totals.

Trigger: the workstation doctrine changed on 2026-09-03. Clock time is now a watched metric, parallelizing genuinely independent work is a standing goal, and serialization is owed a named contended resource. The ladder's opening sentence, "rungs are sequential with one executor and no overlap", predates that ruling.

## 1. The revised shape

Five lanes are dependency-safe today. Contention, not dependency, caps the width, so the proposal runs **at most two code lanes plus any number of document lanes** at once.

### Phase 0: unblockers

Phase 0 is three items of different kinds, and the review's blocker was that revision 1 treated them as one.

**0a. The reverse-edge completion (research, blocking).** The matrix asked each pair once, in ladder order. Every re-sequencing below inverts a pair, so the load-bearing question for each is the one the matrix did not ask. Before either phase-1 lane launches, one research dispatch classifies these fifteen cells: (5,2), (5,3), (5,4), (5,6), (5,7), (5,9), (8,3), (8,4), (14,7), (14,9), (14,12), (14,13), (12,13), (12,10), (12,11). Every cell returning HARD or SOFT becomes either a row in the contested-surface table or a lane-ordering constraint. This is the cheapest item in Phase 0 and the one that gates the rest, because a claimed-NONE cell that is a real edge is the catastrophic class.

**0b. Rung 1b: the config wall-row registry** (S band, its own plan, the standard implementer/`diff-reviewer`/gate chain). Extend rung 1's existing `Register()` threshold-registry pattern to wall rows and reserved classes, so each rung declares its keys in its own file and no lane ever edits `internal/config/keys.go`. This is the one file the whole build touches and the first thing two lanes collide on. It is a rung rather than an unplanned chore because it changes rung 1's ratified config-core contract, and contract changes get plans. Two riders: the work must re-fingerprint `.gitleaksignore`, whose single allowed finding is pinned by line number (`internal/config/keys_test.go:generic-api-key:66`) so any insertion above line 66 reds the gate for every lane at once, and it must note that fragility in the file.

**0c. The design-language standard** (document, small). The theme-variable contract, the two-axis match/palette composition model, and the embedded font delivery mechanism (woff2 via `go:embed`, never a CDN) are pulled out of rung 7 into a documented standard, derived from the ratified rung 7 theming ruling and the existing `research/2026-09-03-theme-collection-brief.md` standing input. This is a promotion, not a content move: rung 7 still owns the roster, the visual authoring of every surface including rung 5's login pages, and the attended before/after. Without it, rung 5's login pages are the first rendered user-path surface in the build and would be authored ahead of the contract every surface is bound to. It also fixes the same problem for rung 12's driver pairing.

**The serve mount seam is not a Phase 0 artifact.** `Mount(prefix, http.Handler)` cannot express rung 12's needs: rung 12 is the ingress reverse proxy and by the one-owned-prefix rule the streamer owns everything else, so it is the root mux's fallthrough rather than a mounted subtree, and its birth requirements (the header-strip set driven by the driver's configured header name, path normalization rejecting raw/decoded mismatches and `..` segments before prefix routing, per-route write-timeout handling for SSE and audio streams) are all shell-level. The seam is therefore two things with no code ahead of rung 5:

- **A requirement recorded in rung 5's spec.** Rung 5 owns the serve shell, the mux, and the ingress strip, and its spec must name the root-level composition order (ingress strip, then path normalization, then prefix routing, then fallthrough), declare the extension point every later consumer adds a file at, and state which of those points the driver's proxy occupies. Building the ingress strip as an unplanned Phase 0 chore would strip the full fresh-context adversarial cycle from the single artifact the whole trust boundary rests on.
- **A scheduling rule.** No two lanes write `internal/serve` until rung 5 closes. Phase 1's second code slot is lane A (rung 5) or lane T (rung 8), and rung 8 does not touch serve, so the rule costs phase 1 nothing. In phase 3, rung 12 owns the shell after rung 5 and never runs concurrently with another serve-writing lane.

0a, 0b, and 0c are the second code lane's prerequisites, per the ratification.
**Phase 0a is complete (2026-09-04):**
`research/2026-09-04-reverse-edge-completion.md` classifies all
fifteen cells (fourteen SOFT with paste-ready ownership rows and
constraint sentences, one HARD: lane V waits on rung 9's
concurrency ruling at phase 2's opening checkpoint). Its rows and
its "permit and forbid" section govern lane launches by
incorporation; a launch checklist reads that document beside this
spec's table. The contested-surface ownership table in section 3 is also written here.

### Phase 1: evidence, with the spine's head start

Merge point **M1 is rung 4's deploy**, which starts the shadow calendar clock and ends the phase.

| Lane | Rungs | Territory |
|---|---|---|
| **E1 corpus** (doc/data, primary tree) | 2c, 2d | `corpus/`, `internal/harness`, `cmd/harness`, `tools/` |
| **E2 matcher** (code, protected) | 3 → 4 | `internal/engine/matcher`, `musicbrainz`, `shadow` |
| **A auth** (code) | 5 | `internal/auth`, `internal/serve` shell, `internal/notify` |
| **T tags** (code, alternate) | 8 | `internal/engine/tagio` |

Rungs 5 and 8 move into this phase. Rung 5's inbound is rung 1 (discharged) plus the two edges B4 surfaced, both now discharged by Phase 0: the mail seam it births itself (see the ownership table) and the design-language standard it authors against. Rung 8's inbound is rung 2's expected applied tag sets, which are shipped in machinery and format but **not in population**: the corpus stands at 146 cases against a 150 to 300 target with the real-catalog stratum still landing through 2c and 2d, and rung 8's only acceptance evidence is per-case applied-field agreement. The 61 structural synthetic cases shipped in 2b are the stratum least likely to exercise real tag diversity. Rung 8's plan therefore states that its acceptance is provisional against the corpus at its then-current size, and rung 15 carries an acceptance criterion re-running rung 8's applied-field agreement against the complete corpus.

E2 is the critical path and is protected. Conductor attention diverted from `3 → 4` pushes the launch date; attention on any other lane is spent on slack. A and T share the second code slot, one at a time. E1 stays in the primary tree as policy (M5).

### Phase 2: spine, inside the shadow window

| Lane | Rungs | Territory |
|---|---|---|
| **C core** | 6 | `internal/ops`, `internal/engine/store` |
| **V api** | 14 | `internal/serve` (api), contract doc |
| **I intake** | 7 | `internal/intake`, `internal/reject`, `web/` |
| **P pipeline** | 9 | `internal/engine/pipeline` |

C first and alone, since six rungs consume its store schema and ops vocabulary. Then two at a time: V pairs with I, then P with whichever remains. V moves early because its inbound edges from the producing rungs are soft by the additive-within-version policy, so landing the contract early gives every later rung a fixed target to project into.

**Rung 9's concurrency ruling is taken at phase 2's opening checkpoint**, before lane V drafts its plan. The additive-within-version policy makes later verbs cheap; it does not make later changes to existing types cheap. Rung 9's per-album-lease-versus-explicit-queue decision shapes the ops job model, and rung 14's wire types are a published projection of that interface, so a ruling arriving after V lands leaves `/api/v1` taking a breaking change before it has a consumer or carrying a documented breach into launch. This costs no new attended moment: the ruling is already one of the touchpoints the build spec names ahead of time.

### Phase 3: surfaces

Rungs 10, 11, and 13 are mutually independent in disjoint territory: the migrator, the built-ins and hook seam, the queues. Two at a time. Rung 12 (streamer and SSO) takes a slot alongside one of them, requires rung 5 landed, and is the only serve-writing lane in the phase.

**Rung 12 writes the box**, contrary to the build spec's box-safety section (erratum, M12): it sets Navidrome's default UI theme to its built-in Gruvbox Dark via config, verifies the externalized-auth config names against the pinned Navidrome version, and carries provisioning choreography with the first-admin `doctor` preflight. Those land at rung 12 rather than deferring to 16, because rung 16 is the cutover and should discover nothing. Rung 12 therefore carries the quiet-box check and must not overlap rung 10's read-only diff window against the live library.

### Phase 4: launch

15 → 16 → 17, one lane, strictly serial. No parallelism exists here and none should be manufactured.

Maximum width by dependency is 5, 4, 3, 1. Proposed width by contention is two code lanes throughout, plus document lanes.

## 2. What stays serial, and the named reason

- **`3 → 4`, the calendar wall, then `15 → 16 → 17`.** The critical path: hard edges throughout plus two calendar gates.
- **`5 → 12`.** Ingress-before-ExtAuth is a birth requirement. The session cookie, verify-identity seam, and header-strip set are rung 5 artifacts.
- **`6 → 7, 9, 10, 11, 13, 14`.** The store schema and ops vocabulary. C runs alone.
- **`8 → 9`** (tag writing is the pipeline's primary side effect) and **`9 → 10`** (the layout engine is what the parity diff grades).
- **`internal/serve`.** No two lanes write it until rung 5 closes, and rung 12 owns it thereafter without a concurrent serve-writing lane.
- **The three adversarial-cycle rungs, 5, 12, and 16.** Already strictly serial by hard dependency. The concurrency form of correlated-reviewer risk never materializes here, and nothing extra is owed to prevent it.
- **The two owner touchpoints.** Rung 7's before/after and rung 16's sitting stay ordered and never overlap. Correction the analysis surfaced: rung 12 has no owner sitting; its gate is review capacity.
- **Rung 7's attended visual moment.** Rung 7 may *build* against a verify-identity test double before rung 5 lands. Its screenshots may not be taken before rung 5 lands, or they get retaken, and that is one of only two attended moments in the build. The proposal's own schedule makes this moot (rung 5 is phase 1, rung 7 is phase 2); the rule stays written because it is the constraint that survives a schedule change.
- **The box.** Rung 4's additive unit and rung 10's read-only listing are compatible. Rung 12 writes Navidrome config and must not overlap rung 10's diff window. Rung 16 needs a quiet box and overlaps neither.
- **The MusicBrainz budget.** Rung 2's capture tool, rung 3's live tests, and rung 4's shadow draw on one rate-limited service. Captures schedule off the shadow's active hours until rung 3's shared limiter exists. This is the one contended resource file ownership cannot fix.
- **The workstation itself.** Power and CPU. See the lane-launch checklist in section 4.

## 3. Execution mechanics

**Chosen: one shared tree for the primary code lane and all document lanes; a git worktree for the second code lane.**

**The gate mechanism, stated correctly.** `scripts/check.sh` runs three whole-repo legs, not one. The module gate (`vet`, `golangci-lint`, `test`, `tidy-check`) runs over `./...`, so a second Go lane's in-progress package reds the first lane's gate even when the file sets are disjoint. `ruff check` and `ruff format --check` run over `git ls-files '*.py'`, repo-wide tracked Python. `gitleaks detect --no-git -s .` scans the entire worktree including untracked non-ignored files: 69.30 MB in 1m0s as measured on 2026-09-03, of which `corpus/` is 62 MB at 146 cases, roughly doubling at the 300-case target, and paid by every lane on every task, every re-dispatch, and every post-rebase re-run.

The consequence revision 1 missed: **document lanes and code lanes do couple.** Corpus cases are verbatim wire data from a live service, the repo already carries one gitleaks false positive in `.gitleaksignore`, and gitleaks reads the whole tree, so a new case whose response body trips a rule reds the code lane's gate. "Document lanes do not compile, and that is the whole difference" is not true of this gate.

**Gate discipline.**

- **A fast gate and a full gate, one entry point.** `check.sh` takes a scope flag: the module gate per implementer iteration, the full three-leg gate at task close, at lane close, and on master. The rule that a check either gates or does not exist is preserved, because both scopes are the same script and the full gate always runs before anything merges.
- **A repo-level gate lock.** Only one `check.sh` runs per tree at a time. `make tidy-check` runs `go mod tidy` before `git diff --exit-code`, and the first half mutates two tracked files, so two concurrent gate runs in one tree write `go.mod` and `go.sum` at once and a code lane can observe a `go.mod` mid-rewrite. Every lane's chain runs the full gate, document lanes included, so this is not confined to Go-on-Go pairs. Make `tidy-check` non-mutating as well (tidy against a scratch copy). The lock doubles as the cap on concurrent gate runs and fixes `golangci-lint` cache contention.
- Master is never left red. `git commit <specific files>` stays policy, which also defuses the index race. New packages land stub-first so they compile from the first commit.

**Merge ritual, at lane close only.**

1. The lane's gate goes green in its own tree.
2. The lane takes the **merge token**. Exactly one lane holds it at a time; a lane that cannot take it waits. Serialized merges are what make the remaining steps meaningful.
3. The lane rebases onto master and re-runs the full `bash scripts/check.sh` in its own tree.
4. The merge into master is `--ff-only`, and the **full gate re-runs on master** after it. Without this the post-rebase gate is a time-of-check to time-of-use gap: nothing in this repo is CI-enforced, so "master is never left red" is otherwise aspirational.
5. **If the rebase was not a fast-forward**, one `diff-reviewer` dispatch reads the merged result against *both* lanes' acceptance criteria before the merge is announced done. A fast-forward rebase means the other lane landed nothing, so this step skips mechanically and costs nothing.
6. The token is released.

Step 5 exists because a lane's accepted work can be invalidated by another lane's merge with no textual conflict at all: a config default moved, a registry semantic altered, an interface widened. `git rebase` exits zero, the gate may still pass, and the verdict that accepted the work was rendered against a tree that no longer exists.

The trigger for a merge agent is mechanical, so no judgment is required and the conductor never inspects: a rebase exiting nonzero dispatches the merge agent with both task specs. A conductor resolving conflicts inline has broken the thin rule, and the lane structure has then cost more than it bought.

**Contested-surface ownership, written before either lane's plan is drafted.** A contested surface without a written owner blocks the second lane.

| Surface | Owner | Consumers |
|---|---|---|
| The config wall | the registry (rung 1b); no lane edits `keys.go` | every rung declares in its own file |
| `internal/serve` shell, mux, ingress strip | rung 5 (spec requirement); rung 12 thereafter | 6, 7, 12, 14 add files at the declared extension point |
| `internal/notify` (the mail send seam) | **rung 5** — it must send mail to deliver magic links, and its deliverables already carry SMTP send | rung 7's reject copy, rung 9's notifications-and-monitoring integration |
| The theme-variable contract and font delivery | Phase 0c standard | rung 5's login pages, rung 7 (visual authoring, roster, before/after), rung 12's driver pairing |
| The roster loader | rung 5 (assigned twice today, to 5 and 7; award it to 5) | rung 7 |
| The journey vocabulary | rung 7 | rung 9 adds one state |
| The `AppliedFields` type | rung 2's case format pins it | rungs 8, 9 |
| The theme pairing | rung 12 configures Navidrome first | rung 7 authors to match |
| Session vocabulary and token acquisition | the seam is born at rung 5; non-browser token acquisition stays at rung 14 (the build spec's supersession list governs, over the analysis's 5 → 14 note) | rung 14 |
| `internal/harness/conform` (added 2026-09-04, per the rung 3 spec's fold) | born by rung 3 pass 3a; rung 3's lane thereafter | the tier 0-2 conformance runners; rung 15's re-run |
| `internal/harness/oracle` | lane E1 through 2d's close (landed); rung 3 pass 3b adds the derived-AlbumInfo entry point | `make derive`; 3b's ledger |
| `internal/harness/{cassette,stub,analysis}` | lane E1 as shipped; rung 3 pass 3c makes the ruled `Case.Extra` schema-and-stub change, and rung 3 pass 3b adds the `AlbumInfo` ledger and the `defect_free` block to `analysis` (conductor ruling, 2026-09-04) | every conformance consumer |
| `internal/harness/corpus_test.go` (conductor ruling, 2026-09-04) | rung 3's lane while rung 3 runs; lane E1 otherwise, as the corpus's own wiring | every gate-resident ledger check |
| `internal/engine/{matcher,musicbrainz}` (conductor ruling, 2026-09-04; restates the phase-1 lane table's E2 award) | rung 3 (the E2 lane): `matcher` born by 3a, `musicbrainz` born by 3b; no other lane writes either | rung 4's shadow, rung 9's pipeline, rung 15's re-run |
| `corpus/analysis/*` and `corpus/cases/*` | lane E1 authors; the adjudication stream commits dispositions; rung 3 adds schema only through re-derivation (3b's `defect_free`) | rungs 3, 8, 15 |
| `tools/unidecode-table/` (conductor ruling, 2026-09-04) | rung 3 pass 3a — the one-time transliteration-table generator; lane E1 owns the rest of `tools/` | the matcher's embedded table's provenance |

**Rejected alternatives.**

- *Both code lanes in one shared tree.* The `./...` module gate serializes them at the gate anyway, so the sharing buys nothing and adds an index race.
- *An integration branch per wave.* Over-engineering below three concurrent code lanes. Reconsider only if the cap is raised.
- *Serialize all execution, parallelize only prep and research.* Insufficient. It is the current mode and the standing grant, it buys plan readiness rather than build throughput, and it stays in force underneath everything above.

## 4. Conductor and reviewer capacity

- **Two code lanes maximum per conducting session**, plus any number of document lanes. The cap comes from plan staleness and merge load, not report volume. Raising it is a separate decision with evidence.
- **Review independence is not a safeguard against cross-lane defects.** What is true: every `diff-reviewer` dispatch is a fresh zero-context agent, so lanes cannot correlate through a shared reviewer context. What is equally true: a lane's reviewer is structurally blind to the other lane's landed work, so isolation is the mechanism that *hides* a cross-lane defect. That is why the merge ritual owes step 5.
- Domain fan-out runs at each *lane* close against that lane's surface, not batched per phase, which multiplies reviewer dispatches by the lane count. That cost is real and is named in section 7.
- **The conductor stays thin** in every lane: no source file, no diff, no gate transcript. The failure mode to watch is the conductor becoming the merge-conflict resolver.
- **Peak burn, and a small real increase in total spend.** Each phase carries one combined ceiling across its lanes, and **each lane's plan header carries a sub-ceiling that sums to the phase ceiling**, so the 80%-of-ceiling stop rule has something to evaluate itself against inside a lane that cannot see the others' spend. The conductor reallocates sub-ceilings at checkpoints.
- **Per-lane ledgers.** STATUS gains one block per live lane (lane name, tree path, rung, plan path, task ledger, sub-ceiling and spend, next task) rather than growing N times in prose, and the compaction instruction preserves the per-lane blocks rather than a single plan path and pass number.

**Lane-launch checklist.** Every item on it exists because something has already gone wrong without it.

1. **Reverse edges cleared** for any rung this lane runs out of ladder order (Phase 0a).
2. **Contested surfaces written** for every surface this lane touches.
3. **Runaway guard armed** — one per lane, or one guard polling every lane's transcript directory. Mandatory past roughly 30 minutes and easy to forget on lane two.
4. **Lane registry entry written**: lane name, worktree path, pid, to a known file. This *replaces* `pgrep -f Projects/dubplate` as the live-executor check, because git worktrees live under `.claude/worktrees/` inside the repo, so once a second lane exists that pattern matches both lanes permanently and by design. A safety rail that returns a false positive every time is a safety rail that gets ignored, and this one exists because two workflows once raced one worktree for ~1.2M duplicated tokens. The check becomes "no other executor in *this* worktree". Teardown removes the entry.
5. **Venv provisioned if the lane needs it**: `uv sync` against `tools/venv/requirements.txt`. `tools/venv/.venv/` is gitignored at 78 MB and `git worktree add` checks out tracked files only, so a worktree lane has the README and the requirements and no venv. `make derive` (~585s, 90m timeout) is deliberately outside `make check`, so the lane's gate goes green while the only leg exercising the beets oracle never ran and never complained. **Policy: corpus and harness work never moves to a worktree**, which is why lane E1 stays in the primary tree.
6. **Sub-ceiling assigned** from the phase ceiling.
7. **AC power** for any multi-lane session, and the gate lock in force as the cap on concurrent gate runs.
8. **Throughput baseline recorded** for the lane, so the next multi-lane phase is scored rather than asserted.

**On the claim that shared-tree concurrency is proven.** It is proven for document lanes only, and the 2026-09-03 record is partly a counterexample. The 11% battery watchdog fired mid-quad-3 and a lane was stopped cleanly; batch 11's implementer was killed mid-run and left a half-written README edit for the conductor to inspect and revert, which is exactly the work the thin rule forbids, in the shared tree, under the mode being called proven. Over the window where the conductor landed twelve doc commits, the corpus lane's batch cadence fell from roughly one per 25 minutes to roughly one per 50. Correlation is not causation, and it is the only throughput measurement the session offers while this proposal offers none. The doctrine requires naming the contended resource when serializing; the converse obligation, when parallelizing, is showing the resource is not contended, and on 2026-09-03 the workstation was.

## 5. Amendment text for the spec

**Replace** the ladder preamble's first sentence ("Rungs are sequential with one executor and no overlap.") with:

> Rungs are numbered by dependency, not by schedule (Geoff, ratified
> 2026-09-03). The dependency graph is
> `research/2026-09-03-rung-dependency-analysis.md` and the execution
> model reading it is `specs/2026-09-03-parallel-build-approach.md`,
> revision 2, which carries the full model. A rung whose inbound edges
> are satisfied may run concurrently with any other such rung, subject
> to that document's two-code-lane cap, its contested-surface
> ownership table, and its lane-launch checklist; a lane's plan-header
> ceiling is a sub-ceiling of its phase's combined ceiling.

**Replace** the Process section's arc paragraph with:

> Sessions run in phases (Geoff's arc ruling of 2026-09-02, re-keyed
> and ratified 2026-09-03; the full model is
> `specs/2026-09-03-parallel-build-approach.md`, revision 2). A phase
> is bounded by a merge point rather than by a range of rung numbers,
> because lanes cross the old arc boundaries: foundation (rung 1,
> closed); evidence (2, 3, 4, with lanes A and T carrying rungs 5 and
> 8 alongside), closing at rung 4's deploy; spine (6, 14, 7, 9);
> surfaces (10, 11, 12, 13); launch (15, 16, 17, one lane). A phase
> runs with no clears or re-brainstorms inside it and closes with a
> post-mortem and a STATUS handoff; a phase whose lanes outrun one
> session may span more than one, each intermediate close carrying a
> written per-lane handoff in place of the post-mortem. Plans stay
> just-in-time per lane: a lane's next plan is drafted and
> adversarially reviewed in the shadow of that lane's own execution,
> never a phase's worth in advance. Front-loading plans stays refused
> on the evidence that plan staleness was the review rounds' dominant
> defect class, and running N lanes is exactly the condition under
> which that defect returns. Workflow mode is a standing grant for
> execution and prep across the whole build, with human questions
> batched and unblocked work continuing when the owner is
> unavailable. CLAUDE.md's "Build process" section carries the
> operational form every session and subagent loads.

**And, in the same commit, replace** repo `CLAUDE.md`'s "Build process" section. The spec's arc paragraph names that section as the operational form every session and subagent loads, and it currently forbids by name ("never launch agents into a mid-execution tree") the thing the parallel model does by design. Every implementer, reviewer, and merge agent dispatched into a lane would read a file forbidding what it was dispatched to do. Replacement:

> ## Build process (greenfield initiative; ratified 2026-09-02,
> re-keyed 2026-09-03)
>
> The build runs in session-scale PHASES bounded by merge points, not
> one session per rung: foundation (rung 1, closed), evidence (2, 3,
> 4, with rungs 5 and 8 alongside), spine (6, 14, 7, 9), surfaces (10,
> 11, 12, 13), launch (15, 16, 17). Inside a phase: no clears, no
> re-brainstorm. Between phases: close the session, post-mortem,
> STATUS handoff. A phase whose lanes outrun one session may span
> more than one, each intermediate close carrying a written per-lane
> handoff. Canonical record: the build spec's Process section and
> `docs/superpowers/specs/2026-09-03-parallel-build-approach.md`.
>
> **Standing grants (Geoff, 2026-09-02, extended 2026-09-03):**
> workflow mode for the whole build, execution and prep both; batch
> human questions into checkpoint bundles; when Geoff is unavailable,
> continue unblocked work rather than idling; up to two code lanes
> plus any number of document lanes at once.
>
> **Lanes.** A lane is one rung's execution with a named package
> territory. The primary code lane and all document lanes share the
> tree; a second code lane gets its own git worktree, because
> `scripts/check.sh` runs the module gate over `./...` and two
> whole-repo legs besides. Corpus and harness work never moves to a
> worktree. Every lane launch runs the checklist: reverse edges
> cleared, contested surfaces written, runaway guard armed, lane
> registry entry written, venv provisioned if needed, sub-ceiling
> assigned, AC power, throughput baseline. **Never launch a second
> executor into the SAME tree** — the live-executor check is the lane
> registry, not `pgrep -f Projects/dubplate`, which matches every lane
> once a worktree exists.
>
> **The conducting loop, per lane:** execute via the pass-execute
> chain with the runaway guard armed; prep the next rung's plan in a
> workflow against a tree no other executor is writing; land queued
> repo edits in one batch after the executor finishes; plans are
> just-in-time per lane, never front-loaded (plan staleness is the
> proven dominant defect class). A lane closes through the merge
> ritual: merge token, rebase, full gate in-tree, `--ff-only`, full
> gate on master, and one `diff-reviewer` on the merged result
> whenever the rebase was not a fast-forward. Review scales to rung
> risk: mechanical rungs ride the in-chain diff-reviewer; auth (5),
> streamer/SSO (12), and cutover (16) keep the full fresh-context
> adversarial cycle. Subagents cannot dispatch agents: never instruct
> an implementer to run `code-simplifier`; the conductor owns that
> dispatch.

Do arcs survive? Yes, as session boundaries. What changes is the key: the boundary is the merge point, not the arc's last rung number.

## 6. What does not change

The quality bar (exemplary Go, idiomatic Svelte, render-and-read every user-path page before any attended ask). The chain per task: implementer, `diff-reviewer`, full gate, one re-dispatch on `fix`. Both budgets, scored at every pass close against a per-phase ceiling with per-lane sub-ceilings summing to it. Every owner touchpoint: rung 7's before/after, rung 16's sitting, the standing adjudication stream, rung 9's concurrency ruling at spec authoring, now taken at phase 2's opening checkpoint. Nothing here adds an attended moment. The launch bar, all four parts, and the shadow floor's role in it. The just-in-time plan rule, including its adversarial-review clause, adapted per lane as above. Box safety, the quiet-box check, and `--check` first. The gate: never commit red, in any lane, and every check either gates or does not exist.

## 7. What this buys, and what it costs

**The honest headline: the shadow floor sets the launch date, so spine concurrency buys slack, not schedule.** Launch equals rung 4's deploy date plus the floor. Rungs 5 through 14 are three L-band rungs (5, 9, 12) and seven M-band (6, 7, 8, 10, 11, 13, 14), and since an L rung splits into several passes that is more than ten passes, which fit inside a 30-day window at the pace rungs 1 through 2c have demonstrated. The one place parallelism moves the launch date is pulling rung 4's deploy earlier, and that comes from protecting the `3 → 4` chain rather than from running more lanes.

What actually collapses from serial to concurrent:

- **The evidence phase.** Rungs 5 (L) and 8 (M) move off the spine's serial length and ride alongside 2, 3, and 4. Rung 5 is the spine's own critical path, since it gates 12, 14, and 16.
- **The spine phase.** Rung 14 pairs with rung 7 instead of waiting behind nine finished surfaces.
- **The surfaces phase.** Rungs 10, 11, and 13 run two at a time.

If the spine finishes before the floor expires, the slack is better spent on the adversarial rungs, the visual-fidelity chain, and the post-launch improvement candidates the corpus feeds. Not on more concurrency.

Costs, stated plainly. Merge overhead per lane close: one rebase, one in-tree full gate, one on-master full gate, waiting on the merge token, possibly one merge-agent dispatch, and possibly one `diff-reviewer` re-review of the merged result. Domain fan-out at every lane close rather than every phase close, which multiplies those dispatches by the lane count — so parallel lanes cost *slightly more* total tokens, not the same, on top of raising peak burn against the weekly cap. Plan staleness, the build's own named dominant defect class, returning through the back door, with the two-lane cap, the ownership table, and the per-lane adversarial review as the mitigation. One runaway guard per lane. And untested territory: the first Go-on-Go lane pair has no precedent in this repo, its likeliest first failure is a config-wall merge conflict, and Phase 0 exists to prevent exactly that.

## Decisions this proposal needs

1. **Adopt or reject the two-code-lane model**, as amended in section 5. *Ratified 2026-09-03.*
2. **Phase 0 as a prerequisite**, blocking the second code lane until the wall registry and mount seam land. Confirm or waive. *Ratified 2026-09-03; the fold splits Phase 0 into 0a, 0b, 0c and converts the mount seam into a rung 5 spec requirement plus a scheduling rule, which preserves the ratified intent that the second code lane is blocked until they land.*
3. **The estate's own shadow floor.** *Reframed 2026-09-03 rather than answered: an ordinary configuration variable under the defaults discipline.*
4. **New in revision 2, for confirmation at the next checkpoint:** rung 1b as a real rung with its own plan (0b), and the Phase 0c design-language standard promoted out of rung 7. Neither changes a rung's deliverables; both add a small unit of work ahead of phase 1.

## Ratification (Geoff, 2026-09-03)

Decisions 1 and 2 are confirmed as proposed: the two-code-lane model
as amended in section 5, with Phase 0 (the config wall-row registry
and the serve mount seam) as the second lane's prerequisite. Decision
3's framing is corrected rather than answered with a number: the
shadow floor is an ordinary configuration variable, and since the
shadow accumulates from rung 4 through the whole remaining build, any
plausible floor is satisfied before the spine completes - it does not
gate the launch date, which follows from completeness and a clean
shadow record. The estate's value gets set like any other config
choice under the defaults discipline; nothing pre-commits now.
Next per process: the adversarial review round, then the fold of
section 5's amendment text into the build spec.

The review is at
`research/2026-09-02-greenfield-spec-reviews/parallel-approach-review.md`
and this revision folds it. Section 5's spec amendment, as revised by
the fold, is landed in the build spec. **Section 5's CLAUDE.md
replacement is drafted here and not yet applied**, and it must land in
the same commit as the spec amendment: until it does, master carries a
spec that permits lanes and an instruction file that forbids them, and
the instruction file is the one every subagent loads.
