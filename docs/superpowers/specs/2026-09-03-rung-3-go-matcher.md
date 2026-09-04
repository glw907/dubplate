# Rung 3: the Go matcher

Status: DRAFT, rulings ratified. All six open rulings below were accepted
as recommended by Geoff, 2026-09-03 (evening session), before adversarial
review. The adversarial review round and the fold happen next; the rulings
themselves are settled and are not re-litigated by that round. Written against the ratified build spec (`specs/2026-09-02-greenfield-server-build.md`, revision 2.1 as amended), the rulings (`research/2026-09-02-greenfield-server-rulings.md`), the dependency analysis (`research/2026-09-03-rung-dependency-analysis.md`), the execution model (`specs/2026-09-03-parallel-build-approach.md`, revision 2), and the mechanics study (`research/2026-09-02-beets-matcher-mechanics.md`), which is this rung's port map. Nothing here amends the build spec, the launch bar, or the ladder. Per process, this spec goes to Geoff before any adversarial review round and before any pass plan.

The corpus is moving while this is written. Every corpus figure below is stated as "the corpus as committed at 2c/2d close" and is read off the tree at execution time, never from this document.

## What rung 3 builds

A native Go reimplementation of pinned beets 2.13.1's album matcher, from the (items, MusicBrainz responses, config, evaluation date) boundary through to a recommendation and an action. It is the initiative's existential risk: every spine rung built after a failed matcher would be inventory rather than product.

Five deliverables, named in ladder entry 3:

1. **The penalty components and the scoring core**, structured per the evolvability ruling so that changing a matching rule touches one component and its cases rather than a fused loop.
2. **The WS-response-to-matcher-input translation layer**, a first-class deliverable rather than plumbing. Key normalization, per-medium track extraction, pregap and data/video-track handling, alias-driven title and artist rewriting, and release-event selection all move verdicts before any distance is computed.
3. **Album-level matching**, with track assignment pinned to the pinned-beets solver's tie behavior on corpus cases, stable candidate ordering, and the admission filters.
4. **The verified-complete tunable surface** from the mechanics study, with the evaluation date as an explicit input and never a clock read.
5. **The shared MusicBrainz rate limiter**, shipping the per-caller budget interface. Its priority policy is re-verified at rung 9 when its real callers exist.

Rung 3 stays behind the harness until the bar holds.

## Open rulings for Geoff

Every question below is one this spec cannot settle from the written record. Each carries its options and a recommendation. Nothing the record already settles appears here.

### Ruling A: defect reproduction versus fix, and the mechanism

**What the record already settles.** Ruling 2 is explicit: tier one is measured against release truth, a known-defective case enters the known-oracle-bug ledger with an adjudicated expected verdict, and **reproducing the oracle's defect fails the gate**. Conformance is to beets' behavior except where beets is documented wrong. The build spec adds that known oracle defects are the first improvement class and enter pre-launch through that ledger, while improvements never ride the port. So the headline question is answered: the Go matcher does not ship the ft-pattern discount, the parenthetical and remix discount, or any other ledgered defect as behavior.

**What the record does not settle.** The two halves of the bar collide on exactly the ledgered cases. The corpus as committed carries adversarial cases whose derived record is the defect's own arithmetic: `adv-ft-draft-colon` records distance 0.010185185 at strong/apply, where the defect-free figure derived through pinned beets with one `SD_PATTERNS` edit is 0.101851852 at medium/skip. A Go matcher that fixes the defect cannot agree with that record at 1e-12 relative. A Go matcher that agrees with it has reproduced the defect and failed ruling 2. No document states how both proofs are obtained, and the assignment-rejection-gap class is worse: its album-level verdict never moves at all, so there is no verdict-level signal to grade against, and beets has no rejection mechanism for a Go fix to be measured against.

Compounding it: every hypothesis in `oracle_bugs.json` sits at `pending-adjudication` today, `disposition` is owner-only by gate-enforced rule, and the first adjudication batch is prepared for Geoff at 2c's close. Rung 3 cannot know which defects it is authorized to fix until those rows are ruled.

**Options.**

- **(a) A per-defect-class compat switch.** The matcher carries a named switch per ledgered defect class. The conformance gate runs the corpus twice: once in full compat, which must agree bit-identically with every committed record including the adversarial ones, and once in default mode, where each ruled-`oracle-wrong` case must produce its adjudicated corrected verdict. Cost: a permanent compat surface, which must itself be exercised or it rots.
- **(b) Per-case exclusion.** The matcher fixes unconditionally, and ledgered cases are excluded from the distance tier. Cheapest. It also loses the only proof that the fix is the sole divergence on those cases, because a second, accidental divergence hides behind the exclusion.
- **(c) Recorded defect-free expectations.** Each ledgered case carries a machine-readable defect-free distance derived through pinned beets with the one named edit, and the Go matcher is graded against that. The corpus already computed these figures for the ten string-heuristic cases and checked them both ways. It cannot express the assignment class, where the fix has no beets counterfactual.

**Recommendation: (a) as the mechanism, with (c) as the expectation source where a defect-free figure was derived, and a scheduling rule that removes the pending-adjudication deadlock.** The rule: while a hypothesis sits at `pending-adjudication`, the Go matcher reproduces the oracle and the compat switch defaults on for that class. When Geoff rules a hypothesis `oracle-wrong` with a corrected verdict, the switch defaults off and the fix is a deliberate, recorded divergence in the improvements register, exactly as "improvements never ride the port" requires. Rung 3's scope is therefore self-scheduling against the adjudication stream rather than blocked by it, and the launch gate, which refuses an unadjudicated divergence, remains the place the question must be finished.

### Ruling B: rung 4's handoff point

The dependency analysis lists as a re-sequencing candidate that rung 4 may deploy against a functionally complete matcher rather than a bar-clearing one, trading adjudication noise for calendar time, and its risk 1 calls the shadow clock the only irreducible wall. The build spec says the matcher stays behind the harness until the bar holds. Neither statement is a ruling on the handoff, and the two read differently.

**Options.** (a) Rung 4 waits for the full bar. (b) Rung 4 deploys at functional completeness, with every divergence stamped with the matcher commit so divergences from superseded revisions are re-evaluated mechanically instead of consuming owner adjudication.

**Recommendation: (b), with the stamp as a hard requirement of rung 4's divergence row.** Launch date equals rung 4's deploy date plus the shadow floor. "Behind the harness until the bar holds" governs what the matcher is allowed to decide, and shadow decides nothing.

### Ruling C: where the tunable surface lives

The verified-complete surface is roughly thirty verdict-moving settings: three float thresholds, a `max_rec` map of penalty key to recommendation level, a twenty-key weight table, four `preferred`/`ignored`/`required` list surfaces, `ignored_media`, two track-length numbers, two track-kind booleans, `search_limit`, the data-source mismatch penalty, `extra_tags`, `timid`, `languages`, `ignored_alias_types`, and the verdict-to-action trio. `internal/config` today is env-only and its threshold registry carries a string default with an int parser. Twenty distance weights as twenty environment variables is not a design anyone would defend, and the universality rule that every tunable is a threshold with a sane default does not say the wall is the only surface a tunable may have.

This question is not rung 3's alone. Rung 7's format allowlist and theme roster, rung 12's driver config, and rung 13's custody policy all want structured configuration, and rung 1b is already reworking the wall's registry.

**Options.** (a) Everything on the env wall, with the registry gaining typed variants. (b) An engine-owned typed defaults table, with a named small subset registered on the wall, and the rest settable only in code until a rung ships a config-file surface. (c) Rung 3 designs the config-file surface now, for the whole build.

**Recommendation: (b).** Rung 3 ships `matcher.Tunables` as a typed struct carrying pinned beets' shipped defaults with their citations, registers exactly four knobs on the wall (the three recommendation thresholds and the MusicBrainz search limit), and records the rest as deliberately not operator-facing at launch. Rung 3 does not design a config-file format for other rungs. If Geoff prefers (c), it belongs in Phase 0 beside rung 1b, not inside the critical-path lane.

### Ruling D: the defaults themselves

The defaults discipline requires a sensible, research-backed default with its rationale recorded where the key is defined, and names match thresholds as the example whose citation is the corpus. That settles thresholds and weights. It does not settle the surface's edges, and one edge is uncomfortable: `import.languages` ships empty in beets, which makes `_preferred_alias` structurally unreachable, which is half of the alias mis-selection hypothesis the corpus already documents. Shipping beets' empty default reproduces that reachability gap by construction. Shipping a non-empty default is an improvement riding the port, which is forbidden.

**Recommendation: confirm the blanket rule that every default equals pinned beets' shipped value at launch, `languages` included, and log alias-language selection as a post-launch improvement candidate fed by the corpus.** The alternative is to name the exceptions now, each with its research, which is work the port does not need.

### Ruling E: the translated-AlbumInfo ledger

The translation layer moves verdicts before any distance is computed, and the corpus grades it only indirectly. `cassette.Candidate` records source, album id, distance, penalties, and per-track distances, and carries no strings at all, so an alias rewriting that picks the wrong display string is invisible unless it happens to move a number. The alias hypothesis's own evidence lives in applied artist, which rung 3 does not compute.

The fix is a third narrow beets entry point in the shape of `derive_component.py` and 2d's `derive_file_tags.py`: dump the plugin's own built `AlbumInfo` and `TrackInfo` per case into a derived ledger, and grade the Go translation against it directly.

**Options.** (a) Add it as a 2d task. (b) Ship it as a rung 3 deliverable in pass 3b.

**Recommendation: (b).** 2d's plan is at revision 2, adversarially reviewed, with binding pre-rulings, and re-opening it costs more than the lane constraint does. The consequence is stated plainly in the pass split: pass 3b writes `corpus/` and `internal/harness/oracle`, so it runs in the main tree under the standing policy that corpus and harness work never moves to a worktree.

### Ruling F: the tie-mirroring escalation

The corpus pins two genuinely tied smoke cases, `should-ask` and `should-reject`, and ladder entry 2 requires a tied case to be excluded or pinned. Pinned means the Go solver must reproduce `lap.lapjv`'s own choice among equally optimal assignments. That is a from-scratch Jonker-Volgenant implementation whose pivot order matches a specific C library's, and it may not be reachable within the pass band.

`disposition` is owner-only and gate-enforced, so an agent cannot re-disposition a pinned row to `excluded` even with good evidence.

**Recommendation: pre-authorize the escalation shape, not the outcome.** Two failed mirroring attempts stop the task and produce a written finding with the cost matrices and both solvers' optima, and Geoff rules the disposition. This mirrors the named escalation triggers 2c's Tasks 17 and 18 already carry.

## The bar

Rung 3's bar has two parts that are evaluated here and one that is evaluated later.

**Evaluated at rung 3.** The three-tier conformance gate below goes green over the corpus as committed at 2c/2d close, with every disposition honored and every checker proven able to fire.

**Evaluated at rung 3, reported forward.** Distance agreement at the bit-identical target, compared at 1e-12 relative, with the proximity flag guarding the boundaries. The mechanics study establishes that bit-identity is attainable: every accumulation runs in a fixed, code-determined order, all operations are IEEE-754 double add, multiply and divide with no transcendentals, and identical decimal literals parse to identical doubles in both languages. So bit-identity is the target and 1e-12 relative is the comparison. A case that passes at 1e-12 without being bit-identical is a written finding naming the component and the operation order suspected, not a silent pass. Accumulating those findings is how the port's remaining arithmetic drift stays visible.

**Evaluated at rung 15.** The launch bar's matcher clause: zero false confident matches over the whole corpus including the stock-config stratum, and zero unadjudicated cases where beets is right and Go is wrong. The dependency analysis states plainly that the corpus's completeness gates the bar evaluation at rung 15 rather than being a precondition of writing penalty components. Corpus cases landing after rung 3 closes are therefore graded by a rung 15 acceptance criterion re-running the conformance gate against the complete corpus, which is exactly the shape the parallel-approach fold already adopted for rung 8's provisional acceptance. This spec takes that precedent rather than proposing a launch-bar edit.

## Package territory and architecture

### The tree

```
internal/engine/matcher/       the scoring core and the verdict flow
    tunables.go                the verified-complete surface, beets defaults, citations
    info.go                    AlbumInfo, TrackInfo, ItemFields: the matcher's vocabulary
    strdist.go                 string_dist, SD_PATTERNS, the article and & rewrites
    penalty.go                 the accumulator: Add, AddEquality, AddNumber, AddPriority,
                               AddRatio, AddString, AddDataSource, raw/max/normalized
    album_*.go                 one file per album-level component, with its own test
    track_*.go                 one file per track-level component, with its own test
    score.go                   the scoring core: the declared component order, uniformly composed
    assign.go                  the assignment solve, mirroring lapjv tie behavior
    candidate.go               admission filters, insertion order, stable sort
    verdict.go                 recommendation, max_rec downgrade, verdict-to-action
    compat.go                  the per-defect-class switches (ruling A)
internal/engine/musicbrainz/   the WS seam
    translate.go               the translation layer (normalize, media, aliases, events)
    pseudo.go                  the pseudo-release double fetch and merge
    query.go                   Lucene search construction, the inc= sets
    client.go                  the WS/2 client: fmt=json, retries, timeouts, User-Agent
    limiter.go                 the shared limiter and the per-caller budget interface
internal/harness/conform/      rung 3's addition to the harness: the conformance runner
```

Two architectural invariants, both worth a gate check.

**The engine never imports the harness.** `internal/harness` reads the corpus, drives beets, and grades; the engine is the product. The dependency runs harness to engine and never the reverse. This is why the conformance runner lives on the harness side even though rung 3 writes it.

**One package for the scoring core.** The components are files, not subpackages. A `penalty` subpackage would force the accumulator's whole surface to be exported and would put a package boundary between a component and the core that composes it, which is the opposite of what the evolvability ruling asks for. The evolvability property that matters is that a rule change touches one file and its test.

### The scoring core

The build spec's structural consequence is the design: small, named penalty components with declared weights, each carrying its own unit tests, composed by a scoring core that treats them uniformly.

A component is a name, a weight key, and a function that takes the album or track pair plus the tunables and writes into the accumulator. The core holds two slices, album components and track components, in beets' own fixed call order, and iterates them. That order is load-bearing rather than stylistic: the normalizer's denominator counts only the penalties that were actually added, so which components fire and in what order determines the distance, and a component that declines to add anything must decline in exactly the place beets declines.

The accumulator reproduces `Distance`: a map of key to penalty slice, per-key weighted distance, the raw-over-max normalization with its zero-max special case, the nonzero-keys ordering by descending weighted distance then key, and the multi-unit behavior of `add_number`, where one key can carry several penalties and inflate the denominator per unit.

Three port hazards the mechanics study names, restated here because they are easy to get wrong and expensive to find late:

- `string_dist` lowercases before `unidecode` runs inside `_string_dist_basic`, which lowercases again. The port replicates lower, then transliterate, then lower.
- No Unicode normalization happens anywhere in the distance path. The Go port must not helpfully normalize.
- The strong threshold compares with a strict `<` and the medium threshold with `<=`. This asymmetry is the whole reason the proximity flag exists.

### The translation layer

Named as a deliverable because it decides verdicts before the scoring core sees anything. Its scope:

- Key normalization: every response key's hyphens become underscores, recursively, and each generic `relations` array regroups into `{target_type}_relations` keys, sorted by target type and stable within type.
- Per-medium track extraction, including pregap inclusion, `ignore_data_tracks` and `ignore_video_tracks` filtering, and the track-level overrides of recording length, credit, and title.
- Alias-driven rewriting of candidate titles, artist names, and release titles, driven by `import.languages` and `import.ignored_alias_types`.
- Release-event selection under `preferred.countries`, which supplies both the candidate's country and its date, and so feeds the year penalty.
- The pseudo-release double fetch and the merge of the two `AlbumInfo` values.

Graded directly against the derived `AlbumInfo` ledger proposed in ruling E, and indirectly through whole-case agreement.

### The tunable surface, enumerated

Under `match`: `strong_rec_thresh`, `medium_rec_thresh`, `rec_gap_thresh`, `max_rec`, `distance_weights`, `preferred.countries`, `preferred.media`, `preferred.original_year`, `ignored`, `required`, `ignored_media`, `track_length_grace`, `track_length_max`, `ignore_data_tracks`, `ignore_video_tracks`.

Outside `match`: `musicbrainz.search_limit` with its legacy `match.searchlimit` spelling, `musicbrainz.data_source_mismatch_penalty` with its legacy `source_weight` spelling, `musicbrainz.extra_tags`, `import.timid`, `import.languages`, `import.ignored_alias_types`, the loaded metadata source set, and the verdict-to-action trio `import.quiet`, `quiet_fallback`, and `none_rec_action`.

Plus the evaluation date, which is an explicit input to the year penalty and never a clock read.

One discrepancy to resolve at execution rather than inherit: the mechanics study's Q4 calls `distance_weights` nineteen keys while its Q2 lists twenty. The pass plan counts them off `beets/config_default.yaml` at the pinned tag and records the number it found.

The verdict-to-action layer belongs here and not to rung 13. `cassette.Expected` grades `Action`, so rung 3 must compute it to be graded at all. dubplate's custody policy, with its strict default and its owner-queue and contributor-picks settings, sits above that mapping at rung 13 and never inside it.

### The rate limiter

Home is settled by the build spec's tree: `internal/engine/musicbrainz`. It ships the per-caller budget interface with priority classes, because the organizer spec requires a user-present probe to outrank a batch import, and rung 9 re-verifies that policy against its real callers. The 3 to 4 edge is hard and names the limiter by artifact: shadow lookups spend deferred background budget through it.

Rung 3's own conformance runs are offline and spend no budget. The limiter and the live client are built here because rung 4 consumes them, and because the MusicBrainz budget is the one contended resource file ownership cannot fix: rung 2's capture tool, rung 3's live tests, and rung 4's shadow all draw on it.

### What rung 3 does not build

Stated because the boundary is easy to drift across.

- **Applied fields.** The apply-metadata field mapping is rung 8's, and the `AppliedFields` type is pinned by rung 2's case format with rungs 8 and 9 as its consumers. Rung 3 grades no applied field and computes none.
- **Tag I/O.** No file is read or written.
- **The pipeline wiring.** The matcher wires into the pipeline at rung 15 by the spec's own design; rung 9 builds against the matcher interface.
- **A beets config parser for general use.** See the conformance section.

## The conformance gate

### Where it lives, and its offline discipline

`internal/harness/conform`, gate-resident under `go test ./...`, with no venv, no network, no Python, and no shim process. This is rung 2's frozen-verdict fast gate extended to its consumer, and it inherits the corpus discipline wholesale: the gate never reaches the network and never needs the venv, a captured response body is evidence and is never edited, and no agent reads `responses/*.json` into context.

Each case replays through `stub.Server`, which already takes a `cassette.Case` and its directory and answers an uncovered request with a 404 it records. That choice is deliberate rather than convenient. U8 makes MusicBrainz query construction conformance surface, so the Go client's own requests must be graded, and the stub is the instrument that grades them: a run leaving a non-empty miss list fails the gate naming the unmatched `PathQuery`.

`Case.Extra` files are recall-extension fixtures. The Go run is not required to consume one, consuming one is not a failure by itself, and every consumption is reported, so a future candidate-widening improvement is graded on the shared subset rather than escaping the harness.

Two carried facts the runner must respect. A strong ID match legally orphans search captures into `Extra`, so the committed request set is a recorded fixed point rather than a stable set. A sub-second run is a cache hit and proves nothing, so anything reported as evidence runs with `-count=1` and reports broken-down wall time.

If the whole-corpus run outgrows the module gate's tolerance, it becomes its own leg of `scripts/check.sh` at full scope, next to ruff and gitleaks. It does not become an optional target. A check either gates or does not exist.

### The three tiers

**Tier 0, per component.** Every row of `corpus/analysis/components.json` replays against the Go primitive it names. The row's `Args` and `Output` stay raw JSON in the harness by design, and the corpus README says outright that rung 3 is what gives them their Go types, so the typed decoder is a rung 3 deliverable. Coverage today spans every adder, the raw/max/normalized arithmetic including the zero-weight denominator rule and the key added at 0.0, the `generic_penalty_keys` exclusion, `string_dist`'s full pipeline, the `SD_PATTERNS` relative-weight rows, and the whole data-source parameter space, against a row-count floor of 72. This tier is where a component bug is found for the price of a unit test.

**Tier 1, per case, mechanical.** For every committed case: the Go client's request set resolves against the stub with zero misses; the candidate list matches the record's candidates in order by source and album id; each candidate's distance and per-key penalty map match; the per-track distances match; the assignment, extra items, and extra tracks match; the recommendation and action match.

The run configures itself from the case's own echoed `Weights` and `Thresholds` rather than from a re-parsed profile, which keeps beets' YAML parsing out of the conformance claim. The settings the record does not echo, the `preferred` and `ignored` and `required` surfaces, `ignored_media`, `search_limit`, the track-length numbers, the track-kind booleans, and the quiet trio, come from a small decoder over the three committed profiles. That decoder is cross-checked by the echo: where the profiles and the echo overlap, they must agree, so a decoder drift fails rather than silently changing a run's meaning.

**Tier 2, the bar.** Verdict agreement read through ruling 2, evaluated over the corpus as committed at rung 3's close and re-evaluated at rung 15 over the complete corpus.

### What the corpus's own dispositions mean at this gate

| Corpus signal | Rung 3 gate meaning |
| --- | --- |
| `proximity` clear | Distance compared at 1e-12 relative. Verdict must agree. |
| `proximity` flagged | Distance must be bit-identical. Verdict must agree. No epsilon, no adjudication: the flag exists because a comparison-direction disagreement admits no tolerance. |
| `proximity` skipped | The record carries no candidate. The Go run must also produce none. |
| `ties` unique | The assignment must match the record exactly. |
| `ties` pinned | The assignment must match the record exactly. This is the lapjv tie-mirroring requirement, and it is live today on two smoke cases. |
| `ties` excluded | The assignment and the album distance are not graded; the verdict is. The excluded count is reported at every run and may not grow without an owner ruling, because `disposition` is owner-only. |
| `ties` skipped | No assignment to grade. |
| `oracle_bugs` disposition `pending-adjudication` | The Go matcher reproduces the oracle. The case grades like any other. |
| `oracle_bugs` disposition `oracle-wrong` | The Go matcher must produce the adjudicated corrected verdict. The oracle's own record becomes the negative expectation, and reproducing it fails the gate. |
| `oracle_bugs` disposition `oracle-correct` | Grades like any other case. |
| `adjudications` disposition `no-correct-answer` | The verdict is not graded. The distance still is. |

### Assertion discipline

Every checker in the runner ships a fixture that makes it fire, and the task that lands it reports the mutation that proved it can fire. This is the repo's standing rule and it is load-bearing here: a conformance runner that silently passes because it graded nothing is the single most expensive defect this rung can ship. The proximity flag's own precedent is the model, where the flag's proof that it fires lives in hand-authored records parked on a boundary that are never corpus cases.

## The pass split

Rung 3 is an L rung, so it splits into multiple passes at its own planning. Four passes, each M band, each with a real exit gate.

| Pass | Scope | Inbound | Tree |
| --- | --- | --- | --- |
| **3a** Scoring core | The tunable struct with cited defaults; the accumulator; `string_dist` and `SD_PATTERNS`; every album and track component with its own test; the scoring core; the typed `components.json` decoder and tier 0. | Rung 1's config; rung 2's `components.json` (shipped). | Worktree-eligible: engine packages only. |
| **3b** Translation | The translation layer and the pseudo-release merge; the derived `AlbumInfo` ledger, its Python entry point, its gate-resident checker, and its oracle-tagged re-derive test (ruling E). | 3a's info types; the venv; 2d closed. | Main tree: writes `corpus/` and `internal/harness/oracle`. |
| **3c** Client and limiter | The WS/2 client with its retry, timeout, and User-Agent behavior; the Lucene query construction and the `inc=` sets; the shared limiter and the per-caller budget interface; request-construction grading against the committed request logs. | 3a for the likely-tags helper the query is built from; rung 2's request logs and stub. | Main tree: the grading half writes `internal/harness/conform`. |
| **3d** Verdicts and the gate | The assignment solve with tie mirroring; extras and their sort orders; admission filters, insertion order, stable sort, the ID-match early return and the search fallback flow; recommendation, `max_rec` downgrade, verdict-to-action; the compat switches; the whole-corpus conformance runner and its tier 1 and tier 2 gates; the matcher's public entry point. | 3a, 3b, 3c. | Main tree. |

Edges: 3a precedes everything. 3b and 3c are mutually independent and may swap order. 3d needs all three. The 3 to 4 handoff is at 3d's close, subject to ruling B.

**Lane placement, honestly.** Only 3a is cleanly worktree-eligible, because everything else writes harness or corpus territory, and the standing policy is that corpus and harness work never moves to a worktree. Rung 3 therefore wants the main tree for three of its four passes, and lane E1's 2d work wants the main tree too. The cheap sequencing follows: run 3a in a worktree while 2d finishes in the main tree, then take the main tree for 3b through 3d once 2d closes. Rung 3 is the protected critical path, so it takes the primary tree the moment it is free.

**Band and ceiling.** Four M passes. Recent evidence: 2a spent about 1.75M against 5M, 2b about 5.8M against 6M, 2c about 6.5M against 8M through Task 17. Propose a 6M starting ceiling for 3a, re-derived from 3a's actual report before 3b's plan is drafted, per the 2c precedent of refining a ceiling from the first task's measurement rather than asserting one for the whole rung. Each pass plan header carries its own sub-ceiling under the phase's combined ceiling.

**Review.** Rung 3 is not one of the three full adversarial-cycle rungs, so review scales to the in-chain `diff-reviewer` per task plus the pass-end gate and the conductor's `code-simplifier` dispatch. The one place to consider an exception is 3d's conformance runner, because it is the instrument every later claim about the matcher rests on, and a defective instrument passes silently. The pass plan should say whether it takes a fresh-context review of the runner alone.

## What rung 3 needs from rung 2, and what it does not need from 2d

**Hard inbound from rung 2, all shipped.** The frozen-verdict fast gate's machinery, `cassette`'s case format and loader, `stub.Server`'s replay and miss recording, the cassette request-matching rule, the component derivation table, the assignment-tie detector and its ledger, the threshold-proximity flag and its ledger, the claims checker, and the known-oracle-bug and adjudication ledger formats. The dependency is on rung 2's machinery and case format, not on its case count.

**From 2d, nothing blocking.** Precisely:

- **File-level applied-tag derivation is rung 8's blocker, not rung 3's.** 2d's Task 13 work, `derive_file_tags.py`, the vendored template FLAC, `file_tags.json`, and the `mediafile` version pin, produces evidence for the pure-Go FLAC writer. Rung 3 computes no applied field and writes no file. It consumes none of it.
- **The cassette schema bump to `"3"` and `Manifest.MediaFile`** are absorbed by the harness packages rung 3 consumes. Rung 3 pins no schema version of its own and adds no second copy of `wantManifestSchemaVersion`.
- **2d's own invariant protects rung 3's expectations.** Its pre-ruling requires that `make derive` re-derives every committed record byte-identically after its change, and that no item in the corpus derivation path ever gains a non-empty path. So no committed `Expected` moves under rung 3's feet.
- **`Reconcile` and its `versionProbe` tuple**, five components today and six after 2d adds `mediafile`, are derivation-time machinery. Rung 3 never calls `Reconcile`.
- **The version-bump differential and re-capture chore docs** are 2d's Task 3. Rung 3 does not depend on them. Rung 3's measured whole-corpus conformance wall time is a number the differential doc will want, and passing it forward is the same courtesy 2b passed to 2c and 2c passed to 2d.
- **The whole-corpus README rewrite** lands twice, at 2c's close and again at 2d's. It is a read target for the staleness pre-task, not an artifact rung 3 consumes.

**What must be re-read at execution time, because the tree is moving.**

1. `corpus/README.md` in full, as rewritten at 2d's close. Case counts, strata membership, and the adversarial stratum's defect-free figures all come from the tree.
2. `ls corpus/cases | wc -l` and the per-stratum membership, read off the tree. This document pins no count.
3. `corpus/analysis/*.json`, especially `oracle_bugs.json` and `adjudications.json`, whose dispositions determine rung 3's defect-fix scope under ruling A. Also `ties.json`'s dispositions and `proximity.json`'s flagged rows, which determine per-case grading.
4. `internal/harness/analysis`'s exported constants and checkers, and `internal/harness/cassette`'s schema, as shipped.
5. `internal/config`, in case rung 1b's wall-row registry has landed and changed the registration seam.
6. The adjudication batch document and whatever Geoff ruled from it.
7. `CLAUDE.md`'s Build process section, `docs/STATUS.md`'s lane registry, and the build spec's ladder entry 3.

## The staleness pre-task

Every rung 3 pass opens with a staleness pre-task before its first implementer dispatch: one fresh-context agent reads this spec and the pass plan against the tree as shipped, and reports blockers.

The pattern is not a formality. Pass 2b's staleness review found five blockers before any dispatch, plan staleness is named as the review rounds' dominant defect class, and the parallel-approach spec identifies concurrent lanes as exactly the condition under which that defect returns. Rung 3 carries more staleness exposure than 2b did: the corpus and its README are being rewritten while this spec is written, 2d lands mid-rung, rung 1b may change the config core, and the adjudication batch may move dispositions this spec's grading table reads.

The pre-task's checklist, per pass:

- Does every corpus artifact this plan names exist, at the path and shape named?
- Does every harness API this plan calls have the signature named?
- Have any `oracle_bugs.json` or `adjudications.json` dispositions moved since the plan was drafted, and does any movement change the pass's defect scope?
- Have any `ties.json` dispositions or `proximity.json` rows moved, and does any movement change a case's grading?
- Do the corpus's committed case count and strata membership match what the plan's wall-time and coverage arithmetic assumed?
- Has `internal/config` changed shape?
- Is the lane registry consistent with where this pass intends to run?

The pre-task's output is blockers with evidence, not fixes. The conductor rules each one before the first task dispatches.

## Risks

**1. The tie-mirroring requirement may not be reachable.** Two committed cases are pinned tied, and mirroring a specific C solver's choice among equally optimal assignments is the hardest single piece of the port. Mitigation: ruling F's pre-authorized escalation, and 3d's plan carries the named trigger of two failed attempts.

**2. The conformance runner is an instrument that can pass silently.** A runner that grades fewer fields than it claims produces a green gate that means nothing, and every later claim about the matcher rests on it. Mitigation: the assertion discipline above, a mutation proof per checker, and an explicit graded-field table in the pass plan that a reviewer can check field by field against `cassette.Expected`.

**3. The defect question decides scope, and it is not yet ruled.** Rung 3 cannot size its own compat surface until ruling A returns and until the first adjudication batch is decided. Mitigation: ruling A's recommendation makes the rung self-scheduling, so a pending hypothesis costs nothing but a switch left at its default.

**4. Main-tree contention with lane E1.** Three of rung 3's four passes want the tree that 2d also needs. Mitigation: the sequencing above, and the standing rule that rung 3 is the protected critical path.

**5. Corpus growth after rung 3 closes.** Handled by the rung 15 re-run criterion, on the rung 8 precedent, and by the fact that the conformance gate is gate-resident, so a new case grades the matcher on the next `check.sh` rather than at a scheduled re-evaluation.

## Process

Each pass gets its own plan via writing-plans, drafted just in time in the shadow of the previous pass's execution and never front-loaded. Plans carry a token ceiling and a checkpoint interval. Execution runs the implementer, `diff-reviewer`, gate chain per task, with workflow mode at six or more tasks. `go-conventions` before any Go, `python-conventions` before the 3b shim. The gate is `bash scripts/check.sh`, green at every task end, never commit red, specific files only. Both budgets are scored at every pass close.

Owner touchpoints rung 3 expects: the six rulings above, batched at one checkpoint before 3a's plan is drafted, and the standing corpus adjudication stream, which rung 3 consumes rather than produces.

---
