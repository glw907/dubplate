# Rung 3: the Go matcher

Status: DRAFT, rulings ratified. All six open rulings below were accepted
as recommended by Geoff, 2026-09-03 (evening session), before adversarial
review. The adversarial review round ran and is folded here; the rulings
themselves are settled and were out of that round's scope. Written against the ratified build spec (`specs/2026-09-02-greenfield-server-build.md`, revision 2.1 as amended), the rulings (`research/2026-09-02-greenfield-server-rulings.md`), the dependency analysis (`research/2026-09-03-rung-dependency-analysis.md`), the execution model (`specs/2026-09-03-parallel-build-approach.md`, revision 2), and the mechanics study (`research/2026-09-02-beets-matcher-mechanics.md`), which is this rung's port map. Nothing here amends the build spec, the launch bar, or the ladder. Per process, this spec goes to Geoff before any adversarial review round and before any pass plan.

Revision 2, 2026-09-04: folds the adversarial review
(`research/2026-09-04-rung-3-spec-review.md`), blockers B1-B3, majors
M1-M9, minors m1-m7. All nineteen findings are folded; two take an
adaptation rather than the literal fix, each named inline where it
lands, and nothing is rejected. The load-bearing changes: rung 3's worktree placement now names
Phase 0a, 0b, and 0c as hard inbound and its wall registration names
rung 1b, because the lane model blocks the second code lane on all
three and the config wall belongs to a registry that has not shipped
(B1); the `Extra` replay channel, the tie ledger's `adjudicator`
field, and the machine-readable defect-free block are named as
deliverables with a pass and a tree rather than assumed to exist
(B2, M4, M6); the `oracle-wrong` grading row now resolves its
corrected verdict off `adjudications.json` by `subject.case_id`, gains
the `no-correct-answer` disposition it omitted, and states the
precedence rule between the two ledgers (B3); tier 1's profile decoder
is restated as a `Tunables`-defaults-merged-with-profile sourcing whose
cross-check against the echo has a non-empty overlap and can therefore
fire (M1); tier 2 gains the ledgers' own pending state so one
unadjudicated divergence blocks the launch gate instead of redding
master for every lane (M7); and the pass split's tree columns are
restated pass by pass against Phase 0, lane E1's primary-tree policy,
and lanes A and T (M5, M9). The review's "Verified sound" list is the
do-not-re-litigate record for this document: the twenty items it names,
including the five deliverables, the launch-bar restatement, the two
architectural invariants, the package territory, port hazards 1 and 3,
the accumulation claims, and the bit-identity argument, are settled and
are not reopened by a later round.

Phase 0a note (2026-09-04): the reverse-edge completion
(research/2026-09-04-reverse-edge-completion.md, cell (8,3)) corrects
this spec's tree-allocation claim that lane T writes no
`internal/harness` or `corpus/` — true of lane A, unproven of lane T.
Lane T runs beside rung 3 only while its plan states its conformance
instrument is read-only against both paths; a rung 8 task deriving a
new corpus record moves to the primary tree. Preferred allocation:
lane T beside 3a, lane A beside 3b-3d.

Conductor rulings on the fold (2026-09-04): B2's cassette schema half
is accepted and lands as pass 3c's first task, where the stub half
already lives — the channel is dormant (zero committed cases carry
`Extra`), so the migration is empty and one pass owns both halves.
3a's main-tree default stands; the worktree early-start remains an
explicitly costed option nobody currently needs.


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

Compounding it: five of the six hypotheses in `oracle_bugs.json` sit at `pending-adjudication` today, the sixth (`bug-threshold-boundary`) having been ruled `oracle-correct` by Geoff on 2026-09-03. `disposition` is owner-only by gate-enforced rule, and the first adjudication batch is prepared for Geoff at 2c's close. Rung 3 cannot know which defects it is authorized to fix until the five pending rows are ruled.

**Options.**

- **(a) A per-defect-class compat switch.** The matcher carries a named switch per ledgered defect class. The conformance gate runs the corpus twice: once in full compat, which must agree bit-identically with every committed record including the adversarial ones, and once in default mode, where each ruled-`oracle-wrong` case must produce its adjudicated corrected verdict. Cost: a permanent compat surface, which must itself be exercised or it rots.
- **(b) Per-case exclusion.** The matcher fixes unconditionally, and ledgered cases are excluded from the distance tier. Cheapest. It also loses the only proof that the fix is the sole divergence on those cases, because a second, accidental divergence hides behind the exclusion.
- **(c) Recorded defect-free expectations.** Each ledgered case carries a machine-readable defect-free distance derived through pinned beets with the one named edit, and the Go matcher is graded against that. The corpus already computed these figures for the ten string-heuristic cases and checked them both ways. It cannot express the assignment class, where the fix has no beets counterfactual.

**Recommendation: (a) as the mechanism, with (c) as the expectation source where a defect-free figure was derived, and a scheduling rule that removes the pending-adjudication deadlock.** The rule: while a hypothesis sits at `pending-adjudication`, the Go matcher reproduces the oracle and the compat switch defaults on for that class. When Geoff rules a hypothesis `oracle-wrong` with a corrected verdict, the switch defaults off and the fix is a deliberate, recorded divergence in the improvements register, exactly as "improvements never ride the port" requires. Rung 3's scope is therefore self-scheduling against the adjudication stream rather than blocked by it, and the launch gate, which refuses an unadjudicated divergence, remains the place the question must be finished.

**(c) has no machine-readable home yet, so rung 3 builds one (fold, M6).** The ten string-heuristic figures exist only as prose inside `adjudications.json`'s `suspected` field: `adj-adv-ft-draft-colon` carries "0.010185185" and "0.101851852" in a sentence, and no schema field anywhere holds a defect-free figure. Pass 3b adds a `defect_free` block to `AdjudicationEntry`, carrying the derived distance, recommendation, and action together with the one named `SD_PATTERNS` edit that produced them, plus the `internal/harness/oracle` entry point that derives it and the `internal/harness/analysis` check that an `oracle-wrong` row on a string-heuristic case carries one. The first population is the ten cases the corpus already computed, six `adv-ft-*` and four `adv-remix-*`. This writes `corpus/` and `internal/harness`, so it runs in the main tree with 3b's other corpus work.

### Ruling B: rung 4's handoff point

The dependency analysis lists as a re-sequencing candidate that rung 4 may deploy against a functionally complete matcher rather than a bar-clearing one, trading adjudication noise for calendar time, and its risk 1 calls the shadow clock the only irreducible wall. The build spec says the matcher stays behind the harness until the bar holds. Neither statement is a ruling on the handoff, and the two read differently.

**Options.** (a) Rung 4 waits for the full bar. (b) Rung 4 deploys at functional completeness, with every divergence stamped with the matcher commit so divergences from superseded revisions are re-evaluated mechanically instead of consuming owner adjudication.

**Recommendation: (b), with the stamp as a hard requirement of rung 4's divergence row.** Launch date equals rung 4's deploy date plus the shadow floor. "Behind the harness until the bar holds" governs what the matcher is allowed to decide, and shadow decides nothing.

### Ruling C: where the tunable surface lives

The verified-complete surface is roughly thirty verdict-moving settings: three float thresholds, a `max_rec` map of penalty key to recommendation level, a twenty-key weight table, four `preferred`/`ignored`/`required` list surfaces, `ignored_media`, two track-length numbers, two track-kind booleans, `search_limit`, the data-source mismatch penalty, `extra_tags`, `timid`, `languages`, `ignored_alias_types`, and the verdict-to-action trio. `internal/config` today is env-only and its threshold registry carries a string default with an int parser. Twenty distance weights as twenty environment variables is not a design anyone would defend, and the universality rule that every tunable is a threshold with a sane default does not say the wall is the only surface a tunable may have.

This question is not rung 3's alone. Rung 7's format allowlist and theme roster, rung 12's driver config, and rung 13's custody policy all want structured configuration, and rung 1b is already reworking the wall's registry.

**Options.** (a) Everything on the env wall, with the registry gaining typed variants. (b) An engine-owned typed defaults table, with a named small subset registered on the wall, and the rest settable only in code until a rung ships a config-file surface. (c) Rung 3 designs the config-file surface now, for the whole build.

**Recommendation: (b).** Rung 3 ships `matcher.Tunables` as a typed struct carrying pinned beets' shipped defaults with their citations, registers exactly four knobs on the wall (the three recommendation thresholds and the MusicBrainz search limit), and records the rest as deliberately not operator-facing at launch. Rung 3 does not design a config-file format for other rungs. If Geoff prefers (c), it belongs in Phase 0 beside rung 1b, not inside the critical-path lane.

**Rung 1b is hard inbound for the registration half (fold, B1).** The contested-surface table awards the config wall to rung 1b's registry with "no lane edits `keys.go`", and rung 1b is Phase 0b, unlanded. Rung 3 therefore registers its four knobs *through* that registry, in its own file, and never by editing `internal/config/keys.go`. Two consequences the pass plan carries. First, the four-knob registration cannot land before rung 1b closes; if 3a runs first, `Tunables` carries all four defaults with their citations and the registration lands as a named follow-on task in the first rung 3 pass that runs after 1b, which is a plan-header item and not a discovery. Second, rung 1b's `.gitleaksignore` rider matters to rung 3 too: the single allowed finding is pinned by line number at `internal/config/keys_test.go:generic-api-key:66`, so any insertion above line 66 reds the gate for every lane at once.

### Ruling D: the defaults themselves

The defaults discipline requires a sensible, research-backed default with its rationale recorded where the key is defined, and names match thresholds as the example whose citation is the corpus. That settles thresholds and weights. It does not settle the surface's edges, and one edge is uncomfortable: `import.languages` ships empty in beets, which makes `_preferred_alias` structurally unreachable, which is half of the alias mis-selection hypothesis the corpus already documents. Shipping beets' empty default reproduces that reachability gap by construction. Shipping a non-empty default is an improvement riding the port, which is forbidden.

**Recommendation: confirm the blanket rule that every default equals pinned beets' shipped value at launch, `languages` included, and log alias-language selection as a post-launch improvement candidate fed by the corpus.** The alternative is to name the exceptions now, each with its research, which is work the port does not need.

### Ruling E: the translated-AlbumInfo ledger

The translation layer moves verdicts before any distance is computed, and the corpus grades it only indirectly. `cassette.Candidate` records source, album id, distance, penalties, and per-track distances, and carries no display strings at all: `Source` and `AlbumID` are strings, but they are identity keys rather than rendered metadata, so an alias rewriting that picks the wrong display string is invisible unless it happens to move a number. The alias hypothesis's own evidence lives in applied artist, which rung 3 does not compute.

The fix is a third narrow beets entry point in the shape of `derive_component.py` and 2d's `derive_file_tags.py`: dump the plugin's own built `AlbumInfo` and `TrackInfo` per case into a derived ledger, and grade the Go translation against it directly.

**Options.** (a) Add it as a 2d task. (b) Ship it as a rung 3 deliverable in pass 3b.

**Recommendation: (b).** 2d's plan is at revision 2, adversarially reviewed, with binding pre-rulings, and re-opening it costs more than the lane constraint does. The consequence is stated plainly in the pass split: pass 3b writes `corpus/` and `internal/harness/oracle`, so it runs in the main tree under the standing policy that corpus and harness work never moves to a worktree.

### Ruling F: the tie-mirroring escalation

The corpus pins two genuinely tied smoke cases, `should-ask` and `should-reject`, and ladder entry 2 requires a tied case to be excluded or pinned. Pinned means the Go solver must reproduce `lap.lapjv`'s own choice among equally optimal assignments. That is a from-scratch Jonker-Volgenant implementation whose pivot order matches a specific C library's, and it may not be reachable within the pass band.

**Recommendation: pre-authorize the escalation shape, not the outcome.** Two failed mirroring attempts stop the task and produce a written finding with the cost matrices and both solvers' optima, and Geoff rules the disposition. This mirrors the named escalation triggers 2c's Tasks 17 and 18 already carry.

**The owner-only rule this ruling rests on is policy today, not mechanism, so rung 3 lands the mechanism (fold, M4).** The owner-only enforcement is the `AgentAdjudicator` sentinel, implemented for adjudications and for oracle bugs. `TieEntry` has no `adjudicator` field at all, and `checkTieEntry` accepts `excluded` and `pinned` interchangeably given any non-empty reason, so an agent could flip a pinned row to `excluded` and leave the gate green. Pass 3d adds the field and the check: `TieEntry.Adjudicator`, required whenever `Disposition` is set, and the sentinel rule from `checkAdjudicationEntry`. This writes `internal/harness/analysis` and the tie driver in `internal/harness/oracle`, main tree.

One adaptation from the review's literal fix, and its reason. The sentinel check applies to `excluded` only, not to `pinned`. `pinned` keeps a case graded; `excluded` stops grading it, and only the second can hide a defect behind a disposition an agent chose. Requiring a human name on `pinned` too would force a new owner touchpoint immediately, because the two committed pinned rows were authored by an agent and would red the gate the moment the field landed. The asymmetry is recorded in the field's own doc comment so a later reader does not mistake it for an oversight.

## The bar

Rung 3's bar has two parts that are evaluated here and one that is evaluated later.

**Evaluated at rung 3.** The three-tier conformance gate below goes green over the corpus as committed at 2c/2d close, with every disposition honored and every checker proven able to fire.

**Evaluated at rung 3, reported forward.** Distance agreement at the bit-identical target, compared at 1e-12 relative, with the proximity flag guarding the boundaries. The mechanics study establishes that bit-identity is attainable: every accumulation runs in a fixed, code-determined order, all operations are IEEE-754 double add, multiply and divide with no transcendentals, and identical decimal literals parse to identical doubles in both languages. So bit-identity is the target and 1e-12 relative is the comparison. A case that passes at 1e-12 without being bit-identical is a written finding naming the component and the operation order suspected, not a silent pass. Accumulating those findings is how the port's remaining arithmetic drift stays visible.

**Evaluated at rung 15.** The launch bar's matcher clause: zero false confident matches over the whole corpus including the stock-config stratum, and zero unadjudicated cases where beets is right and Go is wrong. The dependency analysis states plainly that the corpus's completeness gates the bar evaluation at rung 15 rather than being a precondition of writing penalty components. Corpus cases landing after rung 3 closes are therefore graded by a rung 15 acceptance criterion re-running the conformance gate against the complete corpus, which is exactly the shape the parallel-approach fold already adopted for rung 8's provisional acceptance. This spec takes that precedent rather than proposing a launch-bar edit.

**The rung 15 criterion, written rather than assumed (fold, m7).** Rung 8's re-run criterion is written down in the parallel-approach spec; rung 3's is not written anywhere yet, and a criterion nobody recorded is one a later plan rediscovers from scratch. Rung 3's is therefore stated here as an explicit forward requirement that rung 15's plan inherits:

> Rung 15 re-runs `internal/harness/conform`'s three tiers over the
> complete corpus, at the matcher commit as landed, and its acceptance
> is the launch bar's matcher clause read against that run: zero false
> confident matches over every stratum including stock-config, and zero
> unadjudicated beets-right/Go-wrong cases. Rung 3's own acceptance is
> provisional against the corpus at its then-current size, exactly as
> rung 8's is. Any tier-2 divergence carrying a pending row at rung 15
> is the launch gate's business, not rung 15's to wave through.

Every rung 3 pass close carries this sentence forward into its report, so rung 15's plan reads it from the ledger rather than deriving it.

## Package territory and architecture

### The tree

```
internal/engine/matcher/       the scoring core and the verdict flow
    tunables.go                the verified-complete surface, beets defaults, citations
    info.go                    AlbumInfo, TrackInfo, ItemFields: the matcher's vocabulary
    strdist.go                 string_dist, SD_PATTERNS, the article and & rewrites
    penalty.go                 the accumulator: Add, AddEquality, AddExpr, AddNumber,
                               AddPriority, AddRatio, AddString, AddDataSource,
                               raw/max/normalized
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
internal/engine/tag/           the matcher's public entry point, born in pass 3d as the
                               one package that may see both matcher and musicbrainz
                               (Resolution R18, the package-cycle reason)
internal/harness/conform/      rung 3's addition to the harness: the conformance runner,
                               all three tiers, including tier 0's typed decoder
```

`AddExpr` is in the adder list because `add_expr` is in `componentMethods`, the coverage set every generated component table has to reach, with two committed rows (`add_expr/0`, `add_expr/1`). Tier 0 will demand the Go primitive, so the deliverable list names it (fold, m4).

Two architectural invariants, both worth a gate check.

**The engine never imports the harness.** `internal/harness` reads the corpus, drives beets, and grades; the engine is the product. The dependency runs harness to engine and never the reverse. This is why the conformance runner lives on the harness side even though rung 3 writes it.

**One package for the scoring core.** The components are files, not subpackages. A `penalty` subpackage would force the accumulator's whole surface to be exported and would put a package boundary between a component and the core that composes it, which is the opposite of what the evolvability ruling asks for. The evolvability property that matters is that a rule change touches one file and its test.

### The scoring core

The build spec's structural consequence is the design: small, named penalty components with declared weights, each carrying its own unit tests, composed by a scoring core that treats them uniformly.

A component is a name, a weight key, and a function that takes the album or track pair plus the tunables and writes into the accumulator. The core holds two slices, album components and track components, in beets' own fixed call order, and iterates them. That order is load-bearing rather than stylistic: the normalizer's denominator counts only the penalties that were actually added, so which components fire and in what order determines the distance, and a component that declines to add anything must decline in exactly the place beets declines.

The accumulator reproduces `Distance`: a map of key to penalty slice, per-key weighted distance, the raw-over-max normalization with its zero-max special case, the nonzero-keys ordering by descending weighted distance then key, and the multi-unit behavior of `add_number`, where one key can carry several penalties and inflate the denominator per unit.

Four port hazards, restated here because they are easy to get wrong and expensive to find late. Hazards 1, 3, and 4 are the mechanics study's; hazard 2 is an assertion rung 3 verifies rather than a claim the study makes (fold, m6).

- `string_dist` lowercases before `unidecode` runs inside `_string_dist_basic`, which lowercases again. The port replicates lower, then transliterate, then lower. (Study, the `string_dist` pipeline.)
- No Unicode normalization happens anywhere in the distance path. The Go port must not helpfully normalize. The study documents `unidecode` transliteration and says nothing about NFC or NFD either way, so this is rung 3's own assertion, and 3a verifies it against pinned beets and records the check in its report rather than citing the study for it.
- The strong threshold compares with a strict `<` and the medium threshold with `<=`. This asymmetry is the whole reason the proximity flag exists. (Study, from distance to verdict.)
- The added-key map and the nonzero-key view are different objects, and conflating them is a silent defect (fold, m5). `derive_oracle.py` emits `{key: dist[key] for key in dist._penalties}`, so `Candidate.Penalties` grades key *presence at 0.0* and 272 committed candidates carry one. `max_rec`'s downgrade pass instead reads `keys()`, which yields nonzero penalties only, so a 0.0 penalty never triggers a downgrade. A port that grades against `_penalties` but downgrades against `_penalties` passes tier 1 and mis-downgrades; a port that does the reverse fails tier 1 on every case carrying a zero-valued key. The accumulator therefore exposes both views under distinct names, and each has its own unit test.

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

**`distance_weights` is twenty keys, resolved off the tree (fold, m2).** The mechanics study's Q4 calls it nineteen while its Q2 lists twenty, and the pinned `beets/config_default.yaml` is outside this repo, so the count is settled where the corpus already settled it: every committed record echoes twenty keys, `corpus/cases/clean-match/case.json`'s `expected.weights` among them, and tier 1 configures from that echo. Twenty it is, and 3a's plan records the study's Q4 as the losing document.

**One discrepancy that does remain open: `original_date` and `per_disc_numbering` (fold, M2).** `corpus/README.md`'s estate-profile transform table keeps both verbatim and labels them **verdict-moving** ("controls which year the year penalty evaluates against"; "affects multi-disc track distance and assignment"). The mechanics study lists both under "Checked and found NOT verdict-moving ... application-time only". The enumeration above carries neither. Every estate-profile case in the corpus is derived with both set, so if the README is right the Go `Tunables` are incomplete and tier 1 diverges with no named cause. The pass plan resolves this at execution time, against pinned beets rather than against either document, records the resolution in 3a's plan, and corrects the losing document in the same pass. This is a blocking pre-task item for 3a, not a note.

**Three surfaces no corpus fixture can reach, and how they are graded anyway (fold, M8).** `match.required`, `match.ignored`, and `preferred.original_year` stay unreachable under all three committed profiles, carried as an open item since 2b. The admission filters that consume `required` and `ignored` land in 3d and the preference surface lands in 3a, and the standing rule is that every checker ships a fixture that makes it fire. A corpus case cannot be that fixture here. The route is therefore stated rather than discovered: these three surfaces are graded by table-driven component unit tests against `Tunables` in the packages that own them, with the fixture hand-authored in the package's own `testdata` on the proximity precedent, and each test's report names the mutation that proved it can fire. Their absence from tier 1 is recorded in the pass report as a declared corpus-coverage gap, not as a passing check. A fourth profile that would reach them is a corpus-lane decision and is not rung 3's to take.

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

**`Case.Extra` and the replay channel it does not have yet (fold, B2).** `Case.Extra` files are recall-extension fixtures, and the intent is that the Go run is not required to consume one, that consuming one is not a failure by itself, and that every consumption is reported, so a future candidate-widening improvement is graded on the shared subset rather than escaping the harness. As committed, none of that is reachable. `stub.New` loads only `c.Requests` into its replay map, so anything else records a `Miss` and answers 404, and tier 1 fails a run leaving a non-empty miss list. So any `Extra` consumption would be a hard gate failure today. The deeper obstacle is schema rather than stub: `Case.Extra` is a bare list of response filenames carrying no method and no `path_query`, because the request key is exactly what the capture loop drops when it orphans a response, and `LoadCase` additionally validates that no `Extra` entry names a file a `Request` already consumes. A second response map cannot be keyed against nothing.

The channel is therefore a named two-part deliverable, and this is the second adaptation of a review fix. The cassette half, `Case.Extra` carrying `Request` shape (method, canonical `path_query`, response file) rather than a filename, is `internal/harness/cassette` territory, main tree, and it is hard inbound for the stub half; it is a schema change, so it moves the single `wantManifestSchemaVersion` constant rather than adding a second copy of it, and it is coordinated with lane E1 because that constant and the corpus's derivation path are E1's. The stub half lands in **3c**: an `Extra` replay map alongside the request map, and a `Consumed()` list distinct from `Misses()`, so tier 1's zero-miss rule reads against the union of both maps while every `Extra` consumption is reported rather than failed. Both halves are cheap today and get dearer with every case: **no committed case carries an `Extra` entry** at the corpus as read for this revision, so the migration is empty and the channel is dormant. Its assertion-discipline proof is therefore a hand-authored case in the harness's own `testdata`, never a corpus case, on the proximity precedent. If lane E1 declines the schema change, the fallback is explicit and must be written into 3c's plan rather than left implied: tier 1's zero-miss rule stands as written, `Extra` is unreachable through the stub, and the U8 recall-extension surface is recorded as ungraded at rung 3 with the reason.

Two carried facts the runner must respect. A strong ID match legally orphans search captures into `Extra`, so the committed request set is a recorded fixed point rather than a stable set. A sub-second run is a cache hit and proves nothing, so anything reported as evidence runs with `-count=1` and reports broken-down wall time.

If the whole-corpus run outgrows the module gate's tolerance, it becomes its own leg of `scripts/check.sh` at full scope, next to ruff and gitleaks. It does not become an optional target. A check either gates or does not exist.

### The three tiers

**Tier 0, per component.** Every row of `corpus/analysis/components.json` replays against the Go primitive it names. The row's `Args` and `Output` stay raw JSON in the harness by design, and the corpus README says outright that rung 3 is what gives them their Go types, so the typed decoder is a rung 3 deliverable. Coverage today spans every adder, the raw/max/normalized arithmetic including the zero-weight denominator rule and the key added at 0.0, the `generic_penalty_keys` exclusion, `string_dist`'s full pipeline, the `SD_PATTERNS` relative-weight rows, and the whole data-source parameter space, against a row-count floor of 72. This tier is where a component bug is found for the price of a unit test.

Tier 0's home is `internal/harness/conform`, decoder and runner both (fold, M5). The typed decoder reads a harness artifact and drives engine primitives, which is the harness-to-engine direction the first architectural invariant requires, and putting it in the engine would invert that dependency for a test-only convenience. The pass split's tree column follows from this rather than the other way around; see the tree allocation below.

**Tier 1, per case, mechanical.** For every committed case: the Go client's request set resolves against the stub with zero misses; the candidate list matches the record's candidates in order by source and album id; each candidate's distance and per-key penalty map match; the per-track distances match; the assignment, extra items, and extra tracks match; the recommendation and action match.

**How a tier 1 run is configured (fold, M1).** The run's `Weights` and `Thresholds` come from the case's own echoed record rather than from a re-parsed profile, which keeps beets' YAML parsing out of the conformance claim. Everything else comes from `Tunables`' beets defaults **merged with** the case's profile, in that order, through a small decoder over the three committed profiles. The merge is the load-bearing word. `estate.yaml` and `stock.yaml` set no `match:` and no `musicbrainz:` keys at all, running "the whole distance/threshold surface at beets' defaults", and `locale.yaml` is estate plus exactly `import.languages` and `import.ignored_alias_types`. A decoder that read only what a profile sets would return nothing for almost every setting the run needs, and a cross-check over "where the profiles and the echo overlap" would have an empty overlap and could never fire, which is precisely the defect the assertion discipline forbids.

The cross-check is therefore stated against the merge, not against the profile: **the echoed `Weights` and `Thresholds` must equal the merged result, for every case.** That overlap is non-empty on every case in the corpus, and it fires on any defaults transcription error in `Tunables`, which is the check actually worth having. A profile-decoder drift and a defaults typo both surface as the same failure, naming the key and both values.

**Tier 2, the bar.** Verdict agreement read through ruling 2, evaluated over the corpus as committed at rung 3's close and re-evaluated at rung 15 over the complete corpus.

### What the corpus's own dispositions mean at this gate

| Corpus signal | Rung 3 gate meaning |
| --- | --- |
| `proximity` clear | Distance compared at 1e-12 relative. Verdict must agree. |
| `proximity` flagged | Distance must be bit-identical. Verdict must agree. No epsilon, no adjudication: the flag exists because a comparison-direction disagreement admits no tolerance. |
| `proximity` skipped | The record carries no candidate. The Go run must also produce none. |
| `ties` unique | The assignment must match the record exactly. |
| `ties` pinned | The assignment must match the record exactly. This is the lapjv tie-mirroring requirement, and it is live today on two smoke cases. |
| `ties` excluded | The assignment and the album distance are not graded; the verdict is. The excluded count is reported at every run and may not grow without an owner ruling, because `disposition` is owner-only, enforced from 3d by `TieEntry.Adjudicator` and the sentinel check. |
| `ties` skipped | No assignment to grade. |
| `oracle_bugs` disposition `pending-adjudication` | The Go matcher reproduces the oracle for that defect class: the compat switch defaults on. The case grades like any other. |
| `oracle_bugs` disposition `oracle-wrong` | The compat switch for that defect class defaults off. Each probe case must produce the corrected verdict recorded on **its own `adjudications.json` row**, resolved by `subject.case_id`. The oracle's own record becomes the negative expectation, and reproducing it fails the gate. |
| `oracle_bugs` disposition `no-correct-answer` | The class has no correct answer to converge on. The switch defaults on, no probe case's verdict is graded, and each probe case's distance is still graded at the tier its own `proximity` row sets. |
| `oracle_bugs` disposition `oracle-correct` | Grades like any other case. |
| `adjudications` disposition `no-correct-answer` | The verdict is not graded. The distance still is. |

**Where a corrected verdict actually lives (fold, B3).** `OracleBugEntry` has no corrected-verdict field. `CorrectedVerdict` lives on `AdjudicationEntry`, and `checkAdjudicationEntry` already requires it whenever a row is `oracle-wrong`. Oracle-bug rows are per *hypothesis*, not per case, and the mapping from hypothesis to adjudication rows is incomplete as committed: `bug-nonlatin-alias-selection` names four `probe_case_ids` against two `adjudication_refs`, and `bug-locale-alias-ignored` names two against one. Resolved by subject at the corpus as read for this revision, three probe cases under still-pending hypotheses (`real-nonlatin-alla-pugacheva-vol1`, `real-nonlatin-x-japan-art-of-life`, `real-nonlatin-utada-first-love`) have no adjudication row at all, and so would have no corrected verdict to produce if their hypotheses were ruled `oracle-wrong`. A fourth, `rec-strong-medium-boundary`, is unaffected: its hypothesis is the one already ruled `oracle-correct`. The count is read off the tree at execution time like every other corpus figure here. Pass 3d therefore lands a checker addition alongside its tier 2 work: `CheckOracleBugs` gains the rule that **every probe case of a hypothesis at `oracle-wrong` resolves to an adjudication row by `subject.case_id`**, which needs the adjudication ledger's subject index rather than today's `adjudicationIDs map[string]bool`, so the signature changes with it. A probe case with no row is a gate failure naming the missing row, never a silently ungraded case. This is `internal/harness/analysis`, main tree.

**Precedence, when both ledgers speak.** A case can be a probe of a hypothesis and the subject of an adjudication row whose disposition differs, and the two ledgers answer different questions. **The hypothesis row decides behavior; the case row supplies the expectation.** `oracle_bugs.disposition` alone sets whether the compat switch for that defect class defaults on, because a switch is per defect class and cannot be half on. `adjudications.disposition`, with `corrected_verdict`, alone supplies what a case must produce once that switch is off. So a case row ruled `oracle-wrong` under a hypothesis still at `pending-adjudication` changes no behavior yet; it is the expectation waiting for its hypothesis. And a probe case whose own row is `oracle-correct` or `no-correct-answer` under a hypothesis ruled `oracle-wrong` is a case that is simply not evidence of the class defect: the switch is off and the case grades against its own row.

### Pending divergence, and who owns a red

Tier 2 gets the ledgers' own pending state (fold, M7). When Go and the oracle disagree, deciding which is right is owner adjudication, and `disposition` is owner-only, so a gate that admitted no pending state would hard-red master for every lane the moment a corpus case landed that the matcher diverged on. That runs straight into "never commit red" and "master is never left red", and it puts a lane's unblocking inside an attended queue. The adjudication ledger already solved exactly this, in its own words: pending "is legal at every gate this package runs (`bash scripts/check.sh`); the launch gate, a different gate at a different time, refuses it."

The rule, stated for the conformance gate:

- A tier 2 verdict disagreement whose resolution is an owner question rather than a Go defect is recorded as a pending row in `adjudications.json`: `subject.case_id`, `disposition: pending-adjudication`, `adjudicator: agent`, which both ledger checkers already accept. While that row stands the case's verdict is not graded, `check.sh` stays green, and the **launch gate refuses it**, which is where an unadjudicated divergence has always had to be finished.
- The escape is tier 2 only. Tier 0, tier 1, and any `proximity`-flagged case admit no pending state at all, because a mechanical disagreement is a Go defect until an owner says otherwise, and the flagged cases are exactly the ones where no tolerance is admissible.
- The pending count is reported at every run and read at every checkpoint. It may not grow without the growth being named in the pass report, with the case and the suspicion.

**Who owns a red a corpus-lane commit causes.** A case landing from lane E1 that the matcher fails is E1's red to clear and E2's defect to fix, and those are not the same obligation. E1's obligation is that master is never left red: either the case lands with the pending row above, or the commit is reverted and re-landed once it can. E2's obligation is to clear the Go half of any pending row its own matcher caused. The conductor rules which of the two a given divergence is, since that is a judgment call and not a diff read, and it costs no attended moment.

### Assertion discipline

Every checker in the runner ships a fixture that makes it fire, and the task that lands it reports the mutation that proved it can fire. This is the repo's standing rule and it is load-bearing here: a conformance runner that silently passes because it graded nothing is the single most expensive defect this rung can ship. The proximity flag's own precedent is the model, where the flag's proof that it fires lives in hand-authored records parked on a boundary that are never corpus cases.

The rule reaches the additions this revision names, not only the tiers: the `Extra` replay channel, the `TieEntry.Adjudicator` sentinel check, the `CheckOracleBugs` probe-case rule, and the `defect_free` block's checker each ship a fixture proving the failure, and each is hand-authored in its package's own `testdata` rather than borrowed from the corpus.

## The pass split

Rung 3 is an L rung, so it splits into multiple passes at its own planning. Four passes, each M band, each with a real exit gate.

| Pass | Scope | Inbound | Tree |
| --- | --- | --- | --- |
| **3a** Scoring core | The tunable struct with cited defaults; the accumulator, both penalty views; `string_dist` and `SD_PATTERNS`; every album and track component with its own test; the scoring core; the typed `components.json` decoder and tier 0; the four wall knobs, registered through rung 1b's registry. | Rung 1's config; rung 1b for the wall registration; rung 2's `components.json` (shipped); Phase 0a's `(5,3)` and `(8,3)` cells if 3a runs out of ladder order. | Main tree, because tier 0 writes `internal/harness/conform`. Worktree only in the engine-only configuration below. |
| **3b** Translation | The translation layer and the pseudo-release merge; the derived `AlbumInfo` ledger, its Python entry point, its gate-resident checker, and its oracle-tagged re-derive test (ruling E); the `defect_free` block, its derivation, and its checker (ruling A / M6). | 3a's info types; the venv; 2d closed. | Main tree: writes `corpus/`, `internal/harness/{oracle,analysis,conform}`, and `internal/harness/corpus_test.go`, and births `internal/engine/musicbrainz` (tree column corrected by conductor ruling, 2026-09-04, at 3b's plan fold, m14). |
| **3c** Client and limiter | The WS/2 client with its retry, timeout, and User-Agent behavior; the Lucene query construction and the `inc=` sets; the shared limiter and the per-caller budget interface; request-construction grading against the committed request logs; the stub's `Extra` replay map and `Consumed()` list (B2), against the cassette schema half or its written fallback. | 3a for the likely-tags helper the query is built from; rung 2's request logs and stub; the `Case.Extra` schema change, or the recorded decision not to make it. | Main tree: writes `internal/harness/{conform,stub,cassette}` and `internal/harness/corpus_test.go`, `corpus/manifest.json` (the `schema_version` field alone), `internal/engine/musicbrainz` (born this pass), and three named `internal/engine/matcher` additions (`ItemFields.Comp`, `IsVariousArtistsName`, the `dataSourcePenalty` fallback constant) (tree column corrected by conductor ruling, 2026-09-05, at 3c's plan close). |
| **3d** Verdicts and the gate | The assignment solve with tie mirroring; extras and their sort orders; admission filters, insertion order, stable sort, the ID-match early return and the search fallback flow; recommendation, `max_rec` downgrade, verdict-to-action; the compat switches; the whole-corpus conformance runner and its tier 1 and tier 2 gates including the pending-divergence path; `TieEntry.Adjudicator` and its sentinel check (M4); `CheckOracleBugs`'s probe-case rule (B3); the matcher's public entry point. | 3a, 3b, 3c. | Main tree: writes `internal/engine/{matcher,tag}`, `internal/harness/{conform,analysis,stub,oracle}`, `internal/harness/corpus_test.go`, and `corpus/analysis/ties.json`'s `adjudicator` field by hand-edit (tree column corrected by conductor ruling, 2026-09-05, at 3d's plan close). |

Edges: 3a precedes everything. 3b and 3c are mutually independent and may swap order. 3d needs all three. The 3 to 4 handoff is at 3d's close, subject to ruling B.

**Contested surfaces rung 3 needs written before 3a launches.** Lane-launch checklist item 2 requires a written owner for every surface a lane touches, and the contested-surface table carries no row for `internal/harness` or `corpus/` at all. Rung 3 proposes these four rows, to be written into that table by the conductor before the lane launches rather than discovered at a merge:

| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/harness/conform` | rung 3 (lane E2); created by it, no E1 writes | rung 15's re-run |
| `internal/harness/oracle` | lane E1 through 2d's close, then rung 3 for 3b's `AlbumInfo` and `defect_free` entry points | never written by two lanes at once; the handover is 2d's close |
| `internal/harness/{cassette,stub,analysis}` | lane E1 owns the schema through 2d's close; rung 3 adds the `Extra` channel (3c), `TieEntry.Adjudicator` (3d), and the `CheckOracleBugs` probe rule (3d) after it | rungs 8, 9, 15 |
| `corpus/analysis/*.json` and `corpus/cases/*` | lane E1 authors cases and derives ledgers; `disposition` stays owner-only | rung 3 reads them, and writes only the schema additions above through re-derivation; it never authors a case |

**Lane placement and the phase-1 tree allocation (fold, B1, M5, M9).** Three facts constrain this and none of them were stated in revision 1. Lane E1 stays in the primary tree as policy, and corpus and harness work never moves to a worktree. Three of rung 3's four passes write harness or corpus territory, and after M5 so does 3a. And the second code lane is blocked until Phase 0a, 0b, and 0c land, which `docs/STATUS.md`'s lane registry records literally as "lane-2 (code, worktree): BLOCKED until Phase 0a/0b/0c land"; 0a's cell list includes `(5,3)` and `(8,3)`, the reverse edges that clear rung 3 to run out of ladder order at all. Phase 1 also carries lane A (rung 5) or lane T (rung 8) in the second code slot, one at a time, which revision 1 omitted from its reasoning entirely.

The allocation, pass by pass:

| Configuration | Primary tree | Worktree (second code lane) |
| --- | --- | --- |
| 3a before 2d closes, **only if Phase 0a/0b/0c have landed** | lane E1 finishing 2c/2d, plus lane A or T holding the primary code slot | 3a **engine-only**: `internal/engine/matcher` and nothing else |
| 3a after 2d closes (the default) | 3a in full, tier 0 included | lane A or T |
| 3b | 3b: `corpus/`, `internal/harness/oracle`, engine | lane A or T |
| 3c | 3c: `internal/harness/{conform,stub}`, engine | lane A or T |
| 3d | 3d: `internal/harness/{conform,analysis}`, engine | lane A or T |

Read plainly: **rung 3 holds the primary tree from 2d's close through 3d**, and lanes A and T take the worktree for that whole window, one at a time. Neither writes `internal/harness`, `corpus/`, or `internal/engine/{matcher,musicbrainz}`, and rung 5's `internal/serve` work touches nothing rung 3 owns, so the split is clean in both directions.

The early-start configuration has a cost that the pass plan states rather than absorbs: an engine-only 3a in a worktree cannot write tier 0, so its components ship ungraded by tier 0 until the tier-0 decoder and runner land as the named first task of the next pass to hold the primary tree. That is a real interval with real exposure, and taking it is the conductor's call against how long 2d still has to run. The default is to wait.

**Band and ceiling.** Four M passes. Recent evidence: 2a spent about 1.75M against 5M, 2b about 5.8M against 6M, and 2c **closed at about 8.2M against its 8M re-derived ceiling, with the flag overrun approved** (fold, m1; sizing off 2c's mid-pass 6.5M figure understates the base by roughly a quarter and would mislead the conductor). Propose a 6M starting ceiling for 3a, re-derived from 3a's actual report before 3b's plan is drafted, per the 2c precedent of refining a ceiling from the first task's measurement rather than asserting one for the whole rung. Each pass plan header carries its own sub-ceiling, and those sub-ceilings sum into **phase 1's combined ceiling, which is not yet named anywhere and must be stated at 3a's plan** alongside the sub-ceilings of every other lane live in the phase. Without it the 80%-of-ceiling stop rule has nothing to evaluate itself against inside a lane that cannot see the others' spend.

**Review.** Rung 3 is not one of the three full adversarial-cycle rungs, so review scales to the in-chain `diff-reviewer` per task plus the pass-end gate and the conductor's `code-simplifier` dispatch. The one place to consider an exception is 3d's conformance runner, because it is the instrument every later claim about the matcher rests on, and a defective instrument passes silently. The pass plan should say whether it takes a fresh-context review of the runner alone.

## What rung 3 needs from rung 2, and what it does not need from 2d

**Hard inbound from rung 2, all shipped.** The frozen-verdict fast gate's machinery, `cassette`'s case format and loader, `stub.Server`'s replay and miss recording, the cassette request-matching rule, the component derivation table, the assignment-tie detector and its ledger, the threshold-proximity flag and its ledger, the claims checker, and the known-oracle-bug and adjudication ledger formats. The dependency is on rung 2's machinery and case format, not on its case count.

**Hard inbound from Phase 0, none shipped.** 0a's reverse-edge completion, specifically the `(5,3)` and `(8,3)` cells, clears rung 3 to run out of ladder order. 0b, rung 1b's wall-row registry, is where rung 3's four knobs register. 0c, the design-language standard, does not touch rung 3 but does gate the second code lane, so it gates the worktree configuration above. All three are blockers on the staleness checklist, not shape questions.

**From 2d, nothing blocking.** Precisely:

- **File-level applied-tag derivation is rung 8's blocker, not rung 3's.** 2d's Task 13 work, `derive_file_tags.py`, the vendored template FLAC, `file_tags.json`, and the `mediafile` version pin, produces evidence for the pure-Go FLAC writer. Rung 3 computes no applied field and writes no file. It consumes none of it.
- **The cassette schema bump to `"3"` and `Manifest.MediaFile`** are absorbed by the harness packages rung 3 consumes. Rung 3 pins no schema version of its own and adds no second copy of `wantManifestSchemaVersion`. If 3c's `Case.Extra` change lands, it moves that one constant rather than duplicating it, and it is coordinated with lane E1 as the schema's owner.
- **2d's own invariant protects rung 3's expectations.** Its pre-ruling requires that `make derive` re-derives every committed record byte-identically after its change, and that no item in the corpus derivation path ever gains a non-empty path. So no committed `Expected` moves under rung 3's feet.
- **`Reconcile` and its `versionProbe` tuple**, five components today and six after 2d adds `mediafile`, are derivation-time machinery. Rung 3 never calls `Reconcile`.
- **The version-bump differential and re-capture chore docs** are 2d's Task 3. Rung 3 does not depend on them. Rung 3's measured whole-corpus conformance wall time is a number the differential doc will want, and passing it forward is the same courtesy 2b passed to 2c and 2c passed to 2d.
- **The whole-corpus README rewrite** lands twice, at 2c's close and again at 2d's. It is a read target for the staleness pre-task, not an artifact rung 3 consumes.

**What must be re-read at execution time, because the tree is moving.**

1. `corpus/README.md` in full, as rewritten at 2d's close. Case counts, strata membership, and the adversarial stratum's defect-free figures all come from the tree.
2. `ls corpus/cases | wc -l` and the per-stratum membership, read off the tree. This document pins no count.
3. `corpus/analysis/*.json`, especially `oracle_bugs.json` and `adjudications.json`, whose dispositions determine rung 3's defect-fix scope under ruling A. Also `ties.json`'s dispositions and `proximity.json`'s flagged rows, which determine per-case grading.
4. `internal/harness/analysis`'s exported constants and checkers, and `internal/harness/cassette`'s schema, as shipped.
5. `internal/config`, **and whether rung 1b has landed**. Its registry is the seam rung 3's four knobs register through, so this is a blocker check rather than a shape question.
6. The adjudication batch document and whatever Geoff ruled from it.
7. `CLAUDE.md`'s Build process section, `docs/STATUS.md`'s lane registry, and the build spec's ladder entry 3.

## The staleness pre-task

Every rung 3 pass opens with a staleness pre-task before its first implementer dispatch: one fresh-context agent reads this spec and the pass plan against the tree as shipped, and reports blockers.

The pattern is not a formality. Pass 2b's staleness review found five blockers before any dispatch, plan staleness is named as the review rounds' dominant defect class, and the parallel-approach spec identifies concurrent lanes as exactly the condition under which that defect returns. Rung 3 carries more staleness exposure than 2b did: the corpus and its README are being rewritten while this spec is written, 2d lands mid-rung, rung 1b may change the config core, and the adjudication batch may move dispositions this spec's grading table reads.

The pre-task's checklist, per pass. The first three are blockers: a "no" stops the launch rather than producing a note.

- **Have Phase 0a, 0b, and 0c landed?** 0a's `(5,3)` and `(8,3)` cells clear rung 3 to run out of ladder order; 0b owns the wall the four knobs register on; all three gate the second code lane and therefore the worktree configuration.
- **Has rung 1b's registry shipped, and is the registration seam what this plan calls?** If not, the wall registration is deferred to a named follow-on task and `Tunables` carries the defaults meanwhile.
- **Is the lane registry consistent with where this pass intends to run**, and is the tree this pass wants actually free? The live-executor check is the registry, never `pgrep -f Projects/dubplate`, which matches every lane once a worktree exists.
- Does every corpus artifact this plan names exist, at the path and shape named?
- Does every harness API this plan calls have the signature named?
- Have any `oracle_bugs.json` or `adjudications.json` dispositions moved since the plan was drafted, and does any movement change the pass's defect scope?
- Have any `ties.json` dispositions or `proximity.json` rows moved, and does any movement change a case's grading?
- Do the corpus's committed case count and strata membership match what the plan's wall-time and coverage arithmetic assumed?
- Has `internal/config` changed shape?

The pre-task's output is blockers with evidence, not fixes. The conductor rules each one before the first task dispatches.

## Risks

**1. The tie-mirroring requirement may not be reachable.** Two committed cases are pinned tied, and mirroring a specific C solver's choice among equally optimal assignments is the hardest single piece of the port. Mitigation: ruling F's pre-authorized escalation, and 3d's plan carries the named trigger of two failed attempts.

**2. The conformance runner is an instrument that can pass silently.** A runner that grades fewer fields than it claims produces a green gate that means nothing, and every later claim about the matcher rests on it. Mitigation: the assertion discipline above, a mutation proof per checker including the four ledger and stub additions this revision names, and an explicit graded-field table in the pass plan that a reviewer can check field by field against `cassette.Expected`.

**3. The defect question decides scope, and it is not yet ruled.** Rung 3 cannot size its own compat surface until ruling A returns and until the first adjudication batch is decided. Mitigation: ruling A's recommendation makes the rung self-scheduling, so a pending hypothesis costs nothing but a switch left at its default.

**4. Main-tree contention with lane E1, and with lanes A and T.** All four of rung 3's passes now want the tree 2d also needs, and the second code slot is A's or T's throughout phase 1. Mitigation: the tree allocation above, the four contested-surface rows written before 3a launches, and the standing rule that rung 3 is the protected critical path. The residual exposure is the early-start configuration, whose cost is stated rather than absorbed.

**5. Corpus growth after rung 3 closes.** Handled by the rung 15 re-run criterion, now written down rather than asserted, and by the fact that the conformance gate is gate-resident, so a new case grades the matcher on the next `check.sh` rather than at a scheduled re-evaluation. That immediacy is also the risk: without the pending-divergence path, one unadjudicated case would red master for every lane at once. Mitigation: the tier 2 pending state and the named ownership of a corpus-lane red.

**6. Phase 0 does not land in time.** Rung 3's worktree configuration, its wall registration, and its authority to run out of ladder order all depend on it. Mitigation: the default sequencing waits for 2d and takes the primary tree, which needs 0a's two cells and 0b's registry but no worktree at all, so a late 0c costs rung 3 nothing.

## Process

Each pass gets its own plan via writing-plans, drafted just in time in the shadow of the previous pass's execution and never front-loaded. Plans carry a token ceiling and a checkpoint interval. Execution runs the implementer, `diff-reviewer`, gate chain per task, with workflow mode at six or more tasks. `go-conventions` before any Go, `python-conventions` before the 3b shim. The gate is `bash scripts/check.sh`, green at every task end, never commit red, specific files only. Both budgets are scored at every pass close.

Owner touchpoints rung 3 expects: the six rulings above, batched at one checkpoint before 3a's plan is drafted, and the standing corpus adjudication stream, which rung 3 consumes rather than produces. This revision adds none. The `TieEntry.Adjudicator` asymmetry is chosen precisely so the two committed pinned rows need no backfill sitting; the `defect_free` block and the `CheckOracleBugs` probe-case rule ride the adjudication stream that already exists; and the pending-divergence path routes a red to the conductor rather than to Geoff.

---
