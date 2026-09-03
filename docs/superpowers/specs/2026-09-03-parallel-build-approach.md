# Parallel build approach: a proposed revision to the execution model

Status: PROPOSAL, 2026-09-03. For Geoff's review before any adversarial round, per the repo's process rule. Amends the execution model of the ratified build spec (`2026-09-02-greenfield-server-build.md`, revision 2.1). It changes no rung's content, no acceptance criterion, and no launch bar item.

Foundation: `research/2026-09-03-rung-dependency-analysis.md`, which classified all 105 ordered rung pairs from rung 2 up. Result: 34 hard edges, 21 soft, 50 none. Half the ladder's sequencing is adjacency from authoring for one executor, not dependency.

Trigger: the workstation doctrine changed on 2026-09-03. Clock time is now a watched metric, parallelizing genuinely independent work is a standing goal, and serialization is owed a named contended resource. The ladder's opening sentence, "rungs are sequential with one executor and no overlap", predates that ruling.

## 1. The revised shape

Five lanes are dependency-safe today. Contention, not dependency, caps the width, so the proposal runs **at most two code lanes plus any number of document lanes** at once.

### Phase 0: unblockers (serial, small, not rungs)

- **Config wall-row registry.** Extend rung 1's existing `Register()` threshold-registry pattern to wall rows and reserved classes, so each rung declares its keys in its own file and no lane ever edits `internal/config/keys.go`. This is the one file the whole build touches and the first thing two lanes collide on.
- **Serve mount seam.** Rung 5's spec gains a `Mount(prefix, http.Handler)` requirement. Rung 5 owns the shell, the mux, and the ingress strip; every later serve consumer adds its own file. Without this, rungs 6, 7, 12, and 14 all edit `internal/serve`, and phase 2's real width is one lane.

Both are prerequisites for a second Go lane. Also in phase 0, and cheaper: the contested-surface ownership table in section 3.

### Phase 1: evidence, with the spine's head start

Merge point **M1 is rung 4's deploy**, which starts the shadow calendar clock and ends the phase.

| Lane | Rungs | Territory |
|---|---|---|
| **E1 corpus** (doc/data) | 2c, 2d | `corpus/`, `internal/harness`, `tools/` |
| **E2 matcher** (code, protected) | 3 → 4 | `internal/engine/matcher`, `musicbrainz`, `shadow` |
| **A auth** (code) | 5 | `internal/auth`, `internal/serve` shell |
| **T tags** (code, alternate) | 8 | `internal/engine/tagio` |

Rungs 5 and 8 move into this phase. Rung 5 has zero inbound beyond the discharged rung 1. Rung 8's only inbound is rung 2's expected applied tag sets, already shipped. Their ladder positions are adjacency.

E2 is the critical path and is protected. Conductor attention diverted from `3 → 4` pushes the launch date; attention on any other lane is spent on slack. A and T share the second code slot, one at a time.

### Phase 2: spine, inside the shadow window

| Lane | Rungs | Territory |
|---|---|---|
| **C core** | 6 | `internal/ops`, `internal/engine/store` |
| **V api** | 14 | `internal/serve` (api), contract doc |
| **I intake** | 7 | `internal/intake`, `internal/reject`, `web/` |
| **P pipeline** | 9 | `internal/engine/pipeline`, `internal/notify` |

C first and alone, since six rungs consume its store schema and ops vocabulary. Then two at a time: V pairs with I, then P with whichever remains. V moves early because its inbound edges from the producing rungs are soft by the additive-within-version policy, so landing the contract early gives every later rung a fixed target to project into.

### Phase 3: surfaces

Rungs 10, 11, and 13 are mutually independent in disjoint territory: the migrator, the built-ins and hook seam, the queues. Two at a time. Rung 12 (streamer and SSO) takes a slot alongside one of them and requires rung 5 landed.

### Phase 4: launch

15 → 16 → 17, one lane, strictly serial. No parallelism exists here and none should be manufactured.

Maximum width by dependency is 5, 4, 3, 1. Proposed width by contention is two code lanes throughout, plus document lanes.

## 2. What stays serial, and the named reason

- **`3 → 4`, the calendar wall, then `15 → 16 → 17`.** The critical path: hard edges throughout plus two calendar gates.
- **`5 → 12`.** Ingress-before-ExtAuth is a birth requirement. The session cookie, verify-identity seam, and header-strip set are rung 5 artifacts.
- **`6 → 7, 9, 10, 11, 13, 14`.** The store schema and ops vocabulary. C runs alone.
- **`8 → 9`** (tag writing is the pipeline's primary side effect) and **`9 → 10`** (the layout engine is what the parity diff grades).
- **The three adversarial-cycle rungs, 5, 12, and 16.** Already strictly serial by hard dependency. The concurrency form of correlated-reviewer risk never materializes here, and nothing extra is owed to prevent it.
- **The two owner touchpoints.** Rung 7's before/after and rung 16's sitting stay ordered and never overlap. Correction the analysis surfaced: rung 12 has no owner sitting; its gate is review capacity.
- **Rung 7's attended visual moment.** Rung 7 may *build* against a verify-identity test double before rung 5 lands. Its screenshots may not be taken before rung 5 lands, or they get retaken, and that is one of only two attended moments in the build.
- **The box.** Rung 4's additive unit and rung 10's read-only listing are compatible. Rung 16 needs a quiet box and overlaps neither.
- **The MusicBrainz budget.** Rung 2's capture tool, rung 3's live tests, and rung 4's shadow draw on one rate-limited service. Captures schedule off the shadow's active hours until rung 3's shared limiter exists. This is the one contended resource file ownership cannot fix.

## 3. Execution mechanics

**Chosen: one shared tree for the primary code lane and all document lanes; a git worktree for the second code lane.**

Shared-tree concurrency is already proven in this session, but only for document lanes, and document lanes do not compile. `scripts/check.sh` runs the module gate over `./...`, so a second Go lane's in-progress package reds the first lane's gate even when the file sets are disjoint. That is the whole difference, and it is why the second code lane gets its own tree.

**Merge ritual, at lane close only.** The lane's gate goes green in its own tree, the lane rebases onto master and re-runs `bash scripts/check.sh`, and a trivially additive merge lands. A non-trivial conflict is dispatched to a merge agent with both task specs, never resolved by the conductor reading diffs. A conductor resolving conflicts inline has broken the thin rule, and the lane structure has then cost more than it bought.

**Gate discipline.** Every lane runs the full gate in its own chain, never in the main loop. Master is never left red. `git commit <specific files>` stays policy, which also defuses the index race. New packages land stub-first so they compile from the first commit. `go.mod` conflicts self-heal through `tidy-check` and a post-merge `go mod tidy`.

**Contested-surface ownership, written before either lane's plan is drafted.** The config wall (registry, phase 0), the serve shell (rung 5), the roster loader (assigned twice today, to rungs 5 and 7; award it to 5), the journey vocabulary (rung 7 owns it, rung 9 adds one state), the `AppliedFields` type (rung 2's case format pins it), the theme pairing (rung 12 configures Navidrome first, rung 7 authors to match). A contested surface without a written owner blocks the second lane.

**Rejected alternatives.**

- *Both code lanes in one shared tree.* The `./...` module gate serializes them at the gate anyway, so the sharing buys nothing and adds an index race.
- *An integration branch per wave.* Over-engineering below three concurrent code lanes. Reconsider only if the cap is raised.
- *Serialize all execution, parallelize only prep and research.* Insufficient. It is the current mode and the standing grant, it buys plan readiness rather than build throughput, and it stays in force underneath everything above.

## 4. Conductor and reviewer capacity

- **Two code lanes maximum per conducting session**, plus any number of document lanes. The cap comes from plan staleness and merge load, not report volume. Raising it is a separate decision with evidence.
- **One runaway guard per lane**, or one guard polling every lane's transcript directory. Mandatory past roughly 30 minutes, and easy to forget on lane two, so it joins the lane-launch checklist.
- **Review independence is automatic.** Every `diff-reviewer` dispatch is a fresh zero-context agent, so lanes cannot correlate through a shared reviewer context. Domain fan-out runs at each *lane* close against that lane's surface, not batched per phase.
- **The conductor stays thin** in every lane: no source file, no diff, no gate transcript. The failure mode to watch is the conductor becoming the merge-conflict resolver.
- **Peak burn, not total spend.** Parallel lanes cost the same tokens and raise burn against the weekly cap. Each phase carries one combined ceiling across its lanes, not a ceiling per lane.

## 5. Amendment text for the spec

**Replace** the ladder preamble's first sentence ("Rungs are sequential with one executor and no overlap.") with:

> Rungs are numbered by dependency, not by schedule. The true dependency graph is `research/2026-09-03-rung-dependency-analysis.md`, and the execution model reading it is `specs/2026-09-03-parallel-build-approach.md`. A rung whose inbound edges are satisfied may run concurrently with any other such rung, subject to that document's two-code-lane cap and its contested-surface ownership table. Magnitude bands are unchanged.

**Replace** the Process section's arc paragraph with:

> Sessions run in phases (Geoff's arc ruling of 2026-09-02, re-keyed 2026-09-03). A phase is bounded by a merge point rather than by a range of rung numbers, because lanes cross the old arc boundaries: foundation (rung 1, closed); evidence (2, 3, 4, with lanes A and T carrying rungs 5 and 8 alongside), closing at rung 4's deploy; spine (6, 14, 7, 9); surfaces (10, 11, 12, 13); launch (15, 16, 17, one lane). One conducting session per phase, no clears or re-brainstorms inside one, a close-out between. Plans stay just-in-time **per lane**: a lane's next plan is drafted in the shadow of that lane's own execution, never a phase's worth in advance. Front-loading plans stays refused on the evidence that plan staleness was the review rounds' dominant defect class, and running N lanes is exactly the condition under which that defect returns.

Do arcs survive? Yes, as session boundaries. What changes is the key: the boundary is the merge point, not the arc's last rung number.

## 6. What does not change

The quality bar (exemplary Go, idiomatic Svelte, render-and-read every user-path page before any attended ask). The chain per task: implementer, `diff-reviewer`, full gate, one re-dispatch on `fix`. Both budgets, scored at every pass close against a per-phase ceiling. Every owner touchpoint: rung 7's before/after, rung 16's sitting, the standing adjudication stream, rung 9's concurrency ruling at spec authoring. Nothing here adds an attended moment. The launch bar, all four parts, and the shadow floor's role in it. The just-in-time plan rule, adapted per lane as above. Box safety, the quiet-box check, and `--check` first. The gate: never commit red, in any lane.

## 7. What this buys, and what it costs

**The honest headline: the shadow floor sets the launch date, so spine concurrency buys slack, not schedule.** Launch equals rung 4's deploy date plus the floor. Rungs 5 through 14 are roughly four L-band and six M-band passes, which fit inside a 30-day window at the pace rungs 1 through 2c have demonstrated. The one place parallelism moves the launch date is pulling rung 4's deploy earlier, and that comes from protecting the `3 → 4` chain rather than from running more lanes.

What actually collapses from serial to concurrent:

- **The evidence phase.** Rungs 5 (L) and 8 (M) move off the spine's serial length and ride alongside 2, 3, and 4. That is roughly two passes' worth, and rung 5 is the spine's own critical path since it gates 12, 14, and 16.
- **The spine phase.** Rung 14 pairs with rung 7 instead of waiting behind nine finished surfaces.
- **The surfaces phase.** Rungs 10, 11, and 13 run two at a time.

If the spine finishes before the floor expires, the slack is better spent on the adversarial rungs, the visual-fidelity chain, and the post-launch improvement candidates the corpus feeds. Not on more concurrency.

Costs, stated plainly. Merge overhead: one rebase, one full gate run, and possibly one merge-agent dispatch per lane close. Plan staleness, the build's own named dominant defect class, returning through the back door, with the two-lane cap and the ownership table as the whole mitigation. One runaway guard per lane. Peak token burn against the weekly cap, with no reduction in total spend. And untested territory: the first Go-on-Go lane pair has no precedent in this repo, its likeliest first failure is a config-wall merge conflict, and phase 0 exists to prevent exactly that.

## Decisions this proposal needs

1. **Adopt or reject the two-code-lane model**, as amended in section 5.
2. **Phase 0 as a prerequisite**, blocking the second code lane until the wall registry and mount seam land. Confirm or waive.
3. **The estate's own shadow floor.** Ruling 1 set a configurable floor defaulting to 30 days for adopters. The estate's own value is a separate choice, and it has more effect on the launch date than every scheduling decision above combined. Worth answering now rather than discovering it at rung 15.
