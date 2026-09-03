# Rung dependency analysis: the true graph of rungs 3 through 17

Status: analysis, 2026-09-03. Written against the ratified build spec
(`specs/2026-09-02-greenfield-server-build.md`, revision 2.1) and the
governing organizer spec (`specs/2026-09-01-dubplate-organizer-design.md`).
This document does not amend either spec. It re-derives what the ladder
sequences, separates the orderings that are real from the orderings that
are an artifact of authoring for one executor, and prices the execution
mechanics that concurrency would actually cost.

Scope: rungs 3 through 17. Rung 1 landed 2026-09-02. Rung 2 is in
flight (2c closing, 2d planned), and its machinery half is shipped.

## How edges are classified

- **HARD.** B consumes an artifact or interface that A must ship. The
  artifact is named. Without it B either cannot be built or cannot be
  proven.
- **SOFT.** B benefits from A but can build against A's contract, a
  stub, or a fixture. The integration cost of building B first is named.
- **NONE.** Adjacency in the ladder only. No artifact, no contract, no
  shared surface.

Two edge kinds sit outside this scheme and are marked where they occur:
**calendar** gates (satisfied by elapsed time, not by executor work) and
**latest-bound** gates (A must complete *before* B, where B destroys
something A needs, rather than B consuming something A produces).

## Rung 1's outbound edges are all discharged

Rung 1 is a hard predecessor of nearly every later rung: `internal/config`
and its threshold registry, the reserved-class and superseded-key
machinery, the `cmd/dubplate-server` cobra skeleton, the one-owned-prefix
rule that binds the serve design, the gate rewiring, and the deploy build
target repointed at `cmd/dubplate-server` (rung 4 depends on that last one
by name). Every one of those edges is satisfied. Rung 1 constrains no
scheduling decision and is omitted from the matrix below.

## The matrix

Rows are the predecessor, columns the successor. `H` hard, `S` soft, `—`
none. Read a cell as "does the column rung depend on the row rung".

| A \ B | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| **2** corpus | H | H | — | — | — | H | — | H | — | — | — | — | H | — | — |
| **3** matcher | | H | — | — | — | S | S | — | — | — | H | S | H | — | — |
| **4** shadow | | | — | — | — | — | — | — | — | — | — | — | H | H | S |
| **5** auth | | | | S | S | — | — | — | — | H | S | H | H | H | — |
| **6** ops+store | | | | | H | — | H | H | H | S | H | H | H | S | — |
| **7** intake | | | | | | — | S | — | — | S | S | S | H | H | — |
| **8** tag I/O | | | | | | | H | — | H | — | — | — | H | — | — |
| **9** pipeline | | | | | | | | H | H | S | H | S | H | H | — |
| **10** parity | | | | | | | | | — | — | — | — | S | H | H* |
| **11** built-ins | | | | | | | | | | S | — | — | H | — | — |
| **12** streamer | | | | | | | | | | | — | S | H | H | — |
| **13** queues | | | | | | | | | | | | S | H | — | — |
| **14** API | | | | | | | | | | | | | H | — | — |
| **15** integration | | | | | | | | | | | | | | H | H |
| **16** cutover | | | | | | | | | | | | | | | H |

`H*` at 10 → 17 is a latest-bound gate, explained below.

Totals: 34 hard edges, 21 soft, 50 none. Half the ladder's sequencing is
adjacency.

## The hard edges, with their artifacts

**2 → 3.** The harness machinery: the frozen-verdict fast gate, the
cassette player and its request-matching rule, the component-test
derivation tooling that turns beets' own distance tests into per-component
expected values, the assignment-tie detector, and the threshold-proximity
flag. Rung 3's entire acceptance story is rung 2's gate.

*The adversarial half of this edge matters.* The dependency is on rung 2's
**machinery and case format**, not on rung 2's **case count**. The
machinery shipped in 2a and 2b. The corpus's completeness gates the
evaluation of the matcher bar, which is a launch-gate item evaluated at
rung 15, not a precondition of writing penalty components. Rung 3 can
begin now against 130-plus landed cases and absorb 2c and 2d's additions
as they arrive.

**2 → 4.** The known-acceptable-divergence artifact and the adjudication
ledger format. The spec requires shadow divergences to land in the same
artifact the corpus uses, so the format is rung 2's.

**2 → 8.** The corpus's expected applied tag sets. That record is rung 8's
only acceptance evidence.

**2 → 10.** The corpus `AppliedFields` records and the pinned-beets
adapter, both consumed by the corpus-driven path-parity fixture that
renders path formats through beets and through the new engine.

**2 → 15.** The conformance gate, evaluated against the launch bar.

**3 → 4.** The matcher itself, plus the shared MusicBrainz rate limiter
with its per-caller budget interface. The shadow runner spends deferred
background budget through it.

*The adversarial half.* Rung 4 needs a matcher that **produces verdicts**,
not a matcher that has **cleared the bar**. Divergence is the shadow's
product. Deploying the shadow against a functionally complete but
not-yet-final matcher starts the calendar clock earlier, at the cost of
adjudication noise from divergences a later matcher revision fixes. See
risk 1 for the mitigation.

**3 → 13.** The candidate, distance, and penalty-breakdown types that
decision records serialize.

**3 → 15.** The merge point. The matcher wires into the pipeline here by
the spec's own design.

**4 → 15, 4 → 16.** The shadow record. Both are calendar gates as well as
artifact gates: the configurable duration floor (30-day default) must have
elapsed, and every arrived divergence must be adjudicated.

**5 → 12.** The session cookie, the verify-identity seam, the ingress
serve shell, and the header-strip set the driver's configured header name
drives. The birth requirement is explicit: the ingress flip precedes
enabling the streamer's externalized auth.

**5 → 14.** The authorization model. Header-only admin verbs enforced
through provenance, and non-browser token acquisition on the session
vocabulary, are both rung 5 surfaces.

**5 → 15, 5 → 16.** Login is a launch-bar completeness component, it is on
the rendered user path, and the ingress-before-ExtAuth ordering binds the
cutover.

**6 → 7.** The store schema for journey and session records, and the ops
vocabulary the intake verbs join.

**6 → 9.** The store and the ops seam the pipeline runs on.

**6 → 10.** The target schema and the forward-on-open migrations the
migrator writes into.

**6 → 11.** Named in the spec (U4): the per-track metadata storage
decision is made at rung 6 because rung 11's built-ins consume it.

**6 → 13.** Persisted decision records.

**6 → 14.** The ops interface and its JSON-serializable types. The wire
types are a published projection of the interface, so the API cannot be
authored before the interface exists.

**7 → 15, 7 → 16.** The uploader is on the rendered user path and is one
of the three things Geoff verifies at the cutover sitting.

**8 → 9.** The pure-Go FLAC writer. Writing tags is the pipeline's primary
side effect, which is why the spec pulls tag I/O ahead of the pipeline
rather than folding it in.

**8 → 11.** ReplayGain and fingerprint built-ins write metadata through
the tag writer.

**9 → 10.** The layout and collection-routing engine the parity diff
grades.

**9 → 11.** The named lifecycle events and the pipeline that fires them.

**9 → 13.** The producer of pending decisions, and the point where the
custody policy applies.

**9 → 16.** The pipeline is what the box runs after the swap.

**10 → 16.** The migrator and the no-move assertion, re-run at the cutover
as a gate rather than a discovery.

**10 → 17 (latest-bound).** Rung 10 captures the legacy ledger's on-disk
format as a documented fixture while `_salvage/` is still legible. Rung 17
deletes `_salvage/`. This is the ladder's only edge where A must finish
before B destroys A's input, rather than B consuming A's output. It is
already satisfied by the ladder order and would only break if rung 17 were
ever pulled forward.

**11 → 15, 12 → 15, 13 → 15, 14 → 15, 6 → 15, 8 → 15.** Launch-bar part
one is completeness. Every component rung is a hard predecessor of the
integration rung by definition of the gate.

**12 → 16.** The ingress-before-ExtAuth ordering, the Navidrome theme
configuration, and the player half of the attended sitting.

**15 → 16 → 17.** The gate, the swap, and the close-out are strictly
serial. Rung 16 to 17 additionally carries the deprecation-window calendar
gate.

## The soft edges, with their integration costs

**3 → 8.** Both rungs want one `AppliedFields` type. Rung 2's case format
already pins its shape, which discharges most of the risk. Residual cost
if 8 lands first: reconciling the corpus's applied-tag record with the
matcher's output type, a rename-and-thread pass.

**3 → 9.** The pipeline builds against a matcher interface. The spec makes
this soft deliberately by assigning the wiring to rung 15. Rung 9 also
re-verifies the rate limiter's caller-priority policy against its real
callers, which needs rung 3's limiter present. Cost if 9 runs without 3
finished: a stub matcher, and the limiter re-verification defers into the
integration rung where it becomes a discovery rather than a check.

**3 → 14, 9 → 14, 12 → 14, 13 → 14, 7 → 14.** These are soft **by
policy**, not by luck. The additive-within-version compatibility rule
means any verb landing after the contract document is a cheap addition.
Cost per rung: one contract-document amendment and one wire-type addition.
This is the most consequential softness in the graph, and it means rung 14
is not naturally last. See the re-sequencing note below.

**4 → 17.** The shadow runner outlives its rung as the standing
migration-shadowing capability. Close-out reconciles its documentation.

**5 → 6.** Two couplings. First, `internal/serve`: rung 5 builds the
minimal serve shell under the one owned prefix, and rung 6's contract
suite exercises the ops interface over HTTP. Second, the identity type
that ops request types may carry. Cost if 6 lands first: rung 6 stands up
an `httptest` server against its own handler and later remounts into rung
5's shell. Mitigation that removes the cost entirely: rung 5's plan
defines a `Mount(prefix, http.Handler)` registration seam, and every later
serve consumer adds its own file rather than editing the shell.

**5 → 7.** Rung 7 needs the **verify-identity seam's contract**, not rung
5 landed. A test double satisfies the build. Two real costs. First, the
roster loader is assigned twice: rung 5's birth requirements own roster
validation, the `streamer_username` split, and the collision checks, while
rung 7's ladder entry says "roster loading with the birth-requirement
validation". Whichever rung runs first must be given ownership explicitly,
or the second one duplicates it. Second, rung 7's attended visual check
wants a logged-in page. Screenshots of an unauthenticated uploader would
have to be retaken, and Geoff's before/after is one of only two attended
moments in the whole build. **Conclusion: rung 7 may build before rung 5,
but rung 7's owner touchpoint must not.**

**5 → 13.** The admin verbs need admin identity provenance. The minimal
CLI answering path can run on a service-token stub. Cost: one wiring pass.

**6 → 12.** The album deep link needs an album record. Buildable against a
fixture. Cost: one integration test.

**6 → 16.** Backup and restore are exercised at the cutover. Mostly
transitive through 10 and 15.

**7 → 9.** The intake-to-pipeline handoff record, and the journey
vocabulary. Note the ladder assigns the "imported, no player configured"
vocabulary amendment to rung 9 while rung 7 builds the journey records
themselves, so the vocabulary's owner is rung 7 and rung 9 adds one state.
Cost if 9 runs first: a stub producer of "session ready for import".

**7 → 12.** The post-login destination's terminal state, and the theming
coordination. Rung 7 authors the custom DaisyUI gruvbox theme; rung 12
sets Navidrome's built-in Gruvbox Dark. The pair must be authored against
the same published palette standard. Cost if 12 runs first: rung 7's theme
is authored to match what is already configured, which is the cheaper
direction.

**7 → 13.** The appeal action lives on the push-back page. Cost: it ships
as a follow-on edit to rung 7's page rather than as part of rung 13.

**9 → 12.** Post-import scan-trigger wiring.

**10 → 15.** Layout parity is launch-bar part four, first evaluated at
rung 10 and reported at rung 15.

**11 → 12.** The post-scan hook event is named at rung 11 and fired by
rung 12's driver. Cost: the event contract is defined once, wired later.

## Answers to the questions posed

**Does rung 5 depend on anything before it?** Only rung 1, which is
discharged. Rungs 2, 3, and 4 contribute nothing to auth. Rung 5 can start
today, concurrently with the evidence track. This is the single largest
piece of ordering-by-habit in the ladder.

**Are 5 and 6 mutually independent?** Substantially yes. The spec built
them that way on purpose: rung 5's durable nonces and per-email epoch live
in the **state directory**, and rung 4's divergence log is likewise "durable
in the state directory; the store arrives later". Both phrasings are
deliberate decouplings from rung 6's SQLite store. The residual coupling
is the serve shell and the identity type, and pre-assigning the serve
shell to rung 5 removes it.

**Does 7 need 5 landed, or only the seam's contract?** Only the contract,
with the roster-loader ownership resolved and the attended visual moment
held until rung 5 lands.

**Does 8 need anything beyond rung 2's corpus artifacts?** No. Rung 8
needs the expected applied tag sets and nothing else. It does not need
rung 3, rung 6, or rung 7. The reverse constraint the spec states, that
rung 7's format allowlist is bounded by rung 8's tag I/O support, is
discharged by the spec fixing rung 8's scope at FLAC-only. It is an
invariant to hold, not a build edge. Rung 8 is the most under-recognized
parallel opportunity in the ladder.

**Does 3 need anything beyond the frozen corpus?** It needs rung 2's
machinery and rung 1's config, and the corpus's **format** rather than its
**completeness**. The complete corpus gates the bar evaluation at rung 15.

**Owner-sitting gates.** The spec names exactly two attended moments, and
the question's premise needs one correction.

- **Rung 7:** the attended visual check, Geoff's uploader before/after,
  mandatory under the visual-fidelity chain.
- **Rung 16:** the one attended gate, a single sitting verifying login,
  upload, and player on his devices. The family launch is on his yes.
- **Rung 12 has no owner sitting.** It is a full fresh-context adversarial
  review rung carrying the densest security surface and three carried
  blockers. Its gate is review capacity, not owner attention.
- **Standing adjudication stream:** corpus divergences at rungs 2 and 3,
  shadow divergences from rung 4 onward. These do not gate their own
  rungs. They gate **rung 15**, because an unadjudicated divergence blocks
  the launch gate. Batches are therefore on the critical path and must
  stay small and regular.
- **Known-ahead owner ruling:** rung 9's concurrency model, at spec
  authoring rather than as a sitting.

**Shadow calendar gates.** Rung 4 starts the clock at deploy. Rungs 15 and
16 consume the record. Nothing between 5 and 14 is gated on the shadow.
That is precisely why the ladder puts the spine after rung 4: the spine is
designed to be built *inside* the shadow window.

## Lane structure

### Wave A, executable now

| Lane | Rungs | Package territory | Inbound |
|---|---|---|---|
| **E1 corpus** | 2 remainder (2c, 2d) | `corpus/`, `internal/harness`, `cmd/harness`, `tools/` | in flight |
| **E2 matcher** | 3 | `internal/engine/matcher`, `internal/engine/musicbrainz` | 2's machinery (shipped) |
| **T tags** | 8 | `internal/engine/tagio` | 2's applied-tag records (shipped) |
| **A auth** | 5 | `internal/auth`, `internal/serve` (shell) | rung 1 only |
| **C core** | 6 | `internal/ops`, `internal/engine/store` | rung 1 only |

Five lanes are dependency-safe. Whether five are *advisable* is a separate
question, answered in the risks.

### Merge point M1: rung 4

Needs E2. Deploys to the box. Starts the calendar clock. Everything after
this point has slack; everything before it does not.

### Wave B, after A and C land

| Lane | Rungs | Package territory | Inbound |
|---|---|---|---|
| **I intake** | 7 | `internal/intake`, `internal/reject`, `web/` | 6 hard, 5 soft |
| **P pipeline** | 9 | `internal/engine/pipeline`, `internal/notify` | 6 and 8 hard, 7 and 3 soft |
| **X streamer** | 12 | `internal/streamer`, `internal/serve` (proxy) | 5 hard, 6/7/9/11 soft |
| **V api** | 14 | `internal/serve` (api), contract doc | 5 and 6 hard |

I, X, and V all write `internal/serve`. Without the Mount seam from rung
5, Wave B's real width is two lanes, not four.

### Wave C, after P lands

Rungs 10, 11, and 13 are mutually NONE and run concurrently in disjoint
territory: the migrator, the built-ins and hook seam, and the queues.
Rung 13's appeal action reaches into `web/`, so it contends with lane I if
lane I is still open.

### Wave D, strictly serial

15 → 16 → 17. No parallelism exists here and none should be manufactured.

### Maximum width by wave

Five, then four (two without the Mount seam), then three, then one.

## Shared-resource constraints, per lane pair

Ranked by how much damage the collision does.

1. **`internal/config` wall rows.** Every rung adds keys. This is the one
   file the whole build touches, and it is the first thing two lanes
   collide on. The threshold registry already solves the same problem for
   thresholds with `Register()`. **Recommendation: extend the registry
   pattern to wall rows and reserved classes, so each rung declares its
   keys in its own file and no lane ever edits `keys.go`.** This is a
   small rung-5-or-6 change that buys every subsequent lane pair.
2. **`internal/serve`.** Contended by rungs 5, 6, 7, 12, and 14.
   Mitigation as above: rung 5 owns the shell, the mux, and the ingress
   strip, and publishes a mount seam.
3. **`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the build spec.**
   Every lane wants to append. Mitigation is already repo practice: only
   the conductor writes these, batched after an executor finishes. Lanes
   never touch them.
4. **The MusicBrainz budget.** Rung 2's capture tool, rung 3's live
   tests, and rung 4's shadow all draw on one rate-limited external
   service. This is the least obvious shared resource and the only one
   that cannot be resolved by file ownership. Mitigation: the shared
   limiter is a rung 3 deliverable, and captures are scheduled off the
   shadow's active hours until it exists.
5. **`go.mod` / `go.sum`.** Two lanes adding dependencies produce
   conflicting sums. Self-healing: `tidy-check` is in the gate, so a
   post-merge `go mod tidy` resolves it mechanically.
6. **`Makefile` and `scripts/check.sh`.** One-time contention when rung 7
   rejoins the `web/` legs.
7. **The box.** Rung 4 (additive unit) and rung 10 (read-only listing) are
   compatible. Rung 16 requires a quiet box and must never overlap either.

## Execution mechanics, priced honestly

Concurrent `pass-execute` workflows committing to one master branch race
each other. Four options, with real costs.

**(a) Worktree per lane, merge ritual at the lane close.** Satisfies the
one-executor-per-worktree rule by construction. Each lane's gate runs
against its own tree, so a half-landed package in lane B never reds lane
A. Cost: the merge is conductor work, and resolving a real conflict makes
the conductor read diffs, which the thin-conductor rule forbids. Viable
only if the shared-surface list above is neutralized first, so that every
merge is trivially additive.

**(b) Disjoint-package lanes sharing one tree, interleaved commits.**
Cheapest, and the mode this session already proved: the corpus workflow
writes `corpus/` and `internal/harness`, the 2d planning workflow writes
`docs/superpowers/plans` and `research`, and this analysis writes one
research file, with no overlap and no incident. **The proof covers
document lanes only.** Document lanes do not compile, and that is the
whole difference. `scripts/check.sh` runs the module gate over `./...`, so
a Go lane's in-progress package reds the other Go lane's gate even when
the file sets are disjoint. Cost of (b) for Go lanes: every commit must
leave the whole tree green, which serializes the two lanes at the gate
even though their code is independent, plus an index race whenever two
implementers stage at once. Mitigations: `git commit <specific files>`
(already repo policy), staggered commit windows, and stub-first landing so
a new package compiles from its first commit.

**(c) A single integration branch per wave.** Lanes branch off it, merge
into it, and it merges to master at the wave close. Cost: one extra merge
level. Benefit: master stays green, and the wave's conflicts are resolved
once by one agent rather than N times by the conductor. Worth it at three
or more concurrent code lanes, over-engineering at two.

**(d) Serialize the executor, parallelize only prep and research.** Zero
code contention, because doc lanes touch no compiled file. This is the
mode the build spec already grants and this session already runs. It buys
plan readiness, not build throughput.

**Recommendation.** Use (d) always. Use (b) for one code lane plus any
number of doc lanes. Use (a) for the second and third **code** lane, and
only after the config-wall registry and the serve mount seam exist. Skip
(c) unless three code lanes run at once.

## Conductor and review capacity

**Conductor.** The thin conductor consumes structured reports and decides
accept, re-dispatch, split, upshift, or stop. Per-lane cost is roughly
linear. The failure mode is not report volume, it is the conductor
becoming the merge-conflict resolver, at which point the thin rule is
broken and the lane structure has cost more than it bought.

**Runaway guards.** The guard is per-workflow. N concurrent lanes need N
guards, or one guard polling N transcript directories. This is mandatory
for anything past roughly 30 minutes, and it is easy to forget on lane
two.

**Review capacity and correlated-reviewer risk.** The three full
adversarial-cycle rungs are 5, 12, and 16. They are **strictly serial by
dependency**: 5 → 12 is HARD, 12 → 16 is HARD, 5 → 16 is HARD. Two of
them can never run concurrently, so the concurrency form of
correlated-reviewer risk never materializes on this build. The residual
risk is correlation across *time*: rungs 5 and 12 present related auth
surfaces to the same reviewer agent type. The spec's existing mitigation
holds, because rung 12's review is fresh-context by mandate and its plan
walks the login review texts as an independent checklist instrument.

**Token pool.** Parallel lanes do not reduce total spend. They raise peak
burn against the weekly cap, and the workstation doctrine is explicit that
clock time is not a budget. Parallelism therefore needs a justification
that is not "faster", and on this build there is exactly one: the shadow
window is measured in calendar days, so pulling rung 4's deploy date
earlier converts executor time into launch date. That is the only place
concurrency buys something real.

## The critical path

Executor time: `2 (in flight) → 3 (L) → 4 (M)`, then the calendar wall,
then `15 (M) → 16 (M) → 17 (S)`.

Everything from rung 5 to rung 14 is roughly four L-band and six M-band
passes. At the pace rungs 1, 2a, 2b, and 2c have demonstrated, the spine
fits inside a 30-day shadow window with room. **The shadow floor, not the
spine, sets the launch date, provided rung 4 deploys early.** Two
consequences follow.

First, any conductor attention diverted from the `3 → 4` chain pushes the
launch date by that amount, while attention spent on spine lanes is spent
on slack. Front-load the matcher.

Second, if the spine finishes well before the floor expires, the slack is
better spent on quality than on more concurrency: the adversarial rungs,
the visual-fidelity chain, and the post-launch improvement candidates the
corpus feeds.

## Re-sequencing candidates the graph exposes

- **Rung 5 moves to Wave A.** Zero inbound beyond rung 1. It is the
  longest spine rung (L band) and the gate for rungs 12, 14, and 16, so it
  is the spine's own critical path and belongs as early as possible.
- **Rung 8 moves to Wave A.** Inbound is rung 2's applied-tag records
  alone. The ladder places it at position 8 purely by adjacency to the
  pipeline it feeds.
- **Rung 14 can move earlier**, to just after 5 and 6. Its inbound edges
  from the producing rungs are soft by the additive-within-version policy.
  Landing the contract document and the authorization model early gives
  every later rung a fixed target to project into, rather than a document
  written last against nine finished surfaces.
- **Rung 4 may deploy against a functionally complete matcher** rather
  than a bar-clearing one, trading adjudication noise for calendar. See
  risk 1.

## Owner questions this analysis raises

One question has more effect on the launch date than every scheduling
decision in this document combined: **what shadow floor does the estate
itself run?** Ruling 1 set a configurable duration floor with a 30-day
default and moved the evidentiary weight onto the corpus's real-catalog
and adversarial strata. The default is the adopter-facing shipped value.
The estate's own setting is a separate choice, and it is the launch date's
dominant term. Recommend batching this at the next checkpoint rather than
discovering it at rung 15.

## Top three risks

**1. The shadow clock is the only irreducible wall, and it starts late.**
Launch date equals rung 4's deploy date plus the floor, and nothing in
rungs 5 through 14 changes that. The risk is spending concurrency on spine
lanes, which is slack, while the `3 → 4` chain runs at single-lane speed.
The secondary risk is a clean but empty record: the wind-down retired
every contributor intake path, so the shadow observes only what Geoff adds
for his own listening, and 30 days of zero albums is a weak record even
though ruling 1 makes it non-gating. Mitigations: front-load `3 → 4` and
protect it from lane contention; deploy the shadow against a functionally
complete matcher and stamp every divergence with the matcher commit, so
divergences from superseded revisions are re-evaluated automatically
instead of consuming owner adjudication; and consider deliberately feeding
the box a listening backlog during the window.

**2. Parallel lanes reintroduce plan staleness, the build's own named
dominant defect class.** The process bans front-loaded plans on evidence
from the review rounds. N concurrent lanes require N plans prepped ahead
of their execution, which is the same failure wearing a different hat, and
it lands hardest on the lanes with soft edges (5 → 6 on the serve shell,
5 → 7 on the roster loader, 7 → 9 on the journey vocabulary) where a plan
written early encodes an assumption the other lane then invalidates.
Mitigation: cap concurrency at two code lanes, pre-assign every contested
surface **in writing before either plan is drafted** (the config-wall
registry, the serve mount seam, the roster loader, the journey vocabulary,
the `AppliedFields` type, the theme pairing), and treat any contested
surface without a written owner as a blocker on launching the second lane.

**3. Shared-surface contention turns the thin conductor into a merge
resolver.** `internal/config/keys.go` and `internal/serve` are touched by
almost every rung, `docs/STATUS.md` and `HISTORY.md` by every pass, and
the MusicBrainz budget by three rungs at once. The session's existing
parallel-lane proof covers document lanes only, where nothing compiles and
the module gate is not in play. The first Go-on-Go lane pair is untested
territory, and its most likely failure is a merge conflict in the config
wall that the conductor resolves by reading diffs. Mitigations, in order:
land the wall-row registry and the serve mount seam before the second code
lane starts; keep STATUS, HISTORY, ROADMAP, and the spec conductor-only
and batched; and prefer a worktree per code lane, so a conflict surfaces
at a merge ritual the conductor can delegate rather than mid-gate where it
cannot.
