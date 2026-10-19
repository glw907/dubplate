# Adversarial review: the rung 3 Go-matcher spec

Fresh-context review (2026-09-04) of specs/2026-09-03-rung-3-go-matcher.md,
scoped to exclude the six Geoff-ratified rulings. Folded as revision 2 of
the spec the same day; the "Verified sound" list below is the
do-not-re-litigate record for every rung 3 pass plan.

---

I've completed the review. Findings below.

---

# Adversarial review: `docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`

Reviewed against the ratified build spec (rev 2.1), the rulings, the mechanics study, the dependency analysis, parallel-approach rev 2, and the tree as committed at `a878db8` (215 cases, `internal/harness/{cassette,stub,analysis,oracle}`, `corpus/analysis/*.json`). Rulings A–F are treated as settled; every finding below is about operationalization, tree fidelity, or internal consistency.

## BLOCKERS

**B1. 3a's lane placement and ruling C's wall registration both run ahead of Phase 0's ratified prerequisites.**
Spec:60 has 3a "register exactly four knobs on the wall." The contested-surface table awards the config wall to the rung 1b registry with "no lane edits `keys.go`" (`parallel-build-approach.md:150`), and rung 1b is Phase 0b, unlanded. Spec:239 then proposes "run 3a in a worktree while 2d finishes in the main tree" — the worktree *is* the second code lane, and `parallel-build-approach.md:63` states "0a, 0b, and 0c are the second code lane's prerequisites, per the ratification." `docs/STATUS.md:34` records this literally: "lane-2 (code, worktree): BLOCKED until Phase 0a/0b/0c land." The spec's only mention of 1b treats it as optional (spec:264, "in case rung 1b's wall-row registry has landed"). Phase 0a's cell list also includes `(5,3)` and `(8,3)` — the reverse-edge questions that clear rung 3 to run out of ladder order at all.
*Fix:* name Phase 0a/0b/0c as hard inbound for the worktree placement, name 1b as hard inbound for 3a's wall registration (or move registration to a later pass), and add both to the staleness pre-task checklist as blockers rather than as shape questions.

**B2. Tier 1's zero-miss rule contradicts the committed stub on `Case.Extra`.**
Spec:190: "The Go run is not required to consume one, consuming one is not a failure by itself, and every consumption is reported." As committed, `stub.New` loads only `c.Requests` into its replay map (`internal/harness/stub/stub.go:55-62`); anything else records a `Miss` and answers 404 (`stub.go:83-88`). Tier 1 fails a run "leaving a non-empty miss list" (spec:188, spec:200). So *any* Extra consumption is a hard gate failure today, and the reporting channel the spec describes does not exist. `Case.Extra` files are also validated as *not* consumed by a request (`cassette.go:168-175`), so they are structurally unreachable through the stub.
*Fix:* make the stub change a named deliverable in 3c — a second response map for `Extra` plus a distinct `Consumed()` list separate from `Misses()` — or scope tier 1's zero-miss rule to non-`Extra` requests. Either way the spec must say which, and note the change lands in `internal/harness`, not the engine.

**B3. The grading table's `oracle-wrong` row reads a field that does not exist.**
Spec:218 says a case under an `oracle_bugs` row at `oracle-wrong` "must produce the adjudicated corrected verdict." `OracleBugEntry` has no such field (`internal/harness/analysis/oraclebug.go:21-33`); `CorrectedVerdict` lives on `AdjudicationEntry` (`adjudication.go:51`) and is reachable only through `OracleBugEntry.AdjudicationRefs`. Worse, oracle-bug rows are per *hypothesis*, not per case (`oraclebug.go:17-18`), and the mapping is incomplete as committed: `bug-nonlatin-alias-selection` names four `probe_case_ids` but only two `adjudication_refs`, so two of its probe cases would have no corrected verdict to produce if the hypothesis were ruled `oracle-wrong`. The table also omits `no-correct-answer`, which is a legal `oracle_bugs` disposition (`oraclebug.go:120`), and states no precedence rule for a case that is a probe of one hypothesis and the subject of an adjudication row with a different disposition.
*Fix:* rewrite the row to resolve the corrected verdict off `adjudications.json` via `subject.case_id`; require a hypothesis ruled `oracle-wrong` to carry a corrected-verdict adjudication row *per probe case* (and name that as a checker addition, since `CheckOracleBugs` does not enforce it today); add the `no-correct-answer` row; state the precedence rule.

## MAJOR

**M1. Tier 1's profile decoder is vacuous as committed, and its cross-check can never fire.**
Spec:202 has "a small decoder over the three committed profiles" supply `preferred`/`ignored`/`required`, `ignored_media`, `search_limit`, the track-length numbers, the track-kind booleans, and the quiet trio, "cross-checked by the echo: where the profiles and the echo overlap, they must agree." But `corpus/README.md:1914-1922` records that `estate.yaml` and `stock.yaml` set **no** `match:` and **no** `musicbrainz:` keys at all — "the whole distance/threshold surface at beets' defaults." So the decoder returns nothing for almost every setting listed, the values must come from `Tunables`' beets defaults, and the overlap between "what the profiles set" and "what the record echoes" is empty. A check that cannot fire is precisely the defect spec:224 forbids.
*Fix:* state the sourcing as `Tunables` defaults **merged with** the profile, and restate the cross-check as: the echoed `Weights` and `Thresholds` must equal that merge, for every case. That has a real, non-empty overlap and fires on any defaults transcription error — which is the check actually worth having.

**M2. `original_date` and `per_disc_numbering`: the corpus README and the mechanics study disagree, and the spec silently follows one.**
`corpus/README.md:1907-1908` keeps both in `estate.yaml` verbatim and labels them **verdict-moving**. The mechanics study lists both under "Checked and found NOT verdict-moving … application-time only" (`2026-09-02-beets-matcher-mechanics.md:461-466`). The spec's "verified-complete" enumeration (spec:157-159) carries neither. Every estate-profile case in the corpus is derived with both set, so if the README is right the Go `Tunables` are incomplete and tier 1 will diverge with no named cause.
*Fix:* name this as a second execution-time discrepancy beside the 19/20 one at spec:163, with the resolution recorded in 3a's plan and the losing document corrected.

**M3. spec:34's claim about `oracle_bugs.json` is false as committed.**
"Every hypothesis in `oracle_bugs.json` sits at `pending-adjudication` today." `bug-threshold-boundary` is at `oracle-correct`, `adjudicator: "Geoff"`, `date: 2026-09-03` (`corpus/analysis/oracle_bugs.json`). Both the README (`corpus/README.md:1863-1867`, `1875-1877`) and STATUS (`docs/STATUS.md:8-9`, "one `oracle-correct`, five pending") say so explicitly.
*Fix:* "five of six hypotheses sit at `pending-adjudication`." The ruling A deadlock argument survives unchanged; the sentence does not.

**M4. `ties.json`'s `disposition` is not owner-only and not gate-enforced.**
Ruling F rests on it (spec:82: "an agent cannot re-disposition a pinned row to `excluded` even with good evidence"), and the grading table repeats it (spec:215). The owner-only mechanism is the `AgentAdjudicator` sentinel, implemented for adjudications (`adjudication.go:140-142`) and oracle bugs (`oraclebug.go:125-127`). `TieEntry` has **no `adjudicator` field** (`ties.go:56-66`), and `checkTieEntry` accepts `excluded` and `pinned` interchangeably given any non-empty reason (`ties.go:135-140`). An agent can flip `should-ask` from `pinned` to `excluded` and the gate stays green.
*Fix:* either restate it as policy rather than mechanism, or make ruling F's escalation shape land the enforcement — an `adjudicator` field on `TieEntry` plus the sentinel check — as a named deliverable (harness territory, main tree, which affects the pass split).

**M5. Tier 0's home contradicts 3a's tree column, the harness-side invariant, and the lane territory table.**
Spec:232 puts "the typed `components.json` decoder and tier 0" in 3a, whose Tree column reads "Worktree-eligible: engine packages only." But spec:125 states the conformance runner lives on the harness side, and spec:234 has 3c be the first pass to write `internal/harness/conform`. Separately, `internal/harness` is lane **E1**'s territory (`parallel-build-approach.md:71`), not E2's (`:72`), and the contested-surface table carries no row for `internal/harness` or `corpus/` — which lane-launch checklist item 2 requires *before* the lane launches (`:178`).
*Fix:* say where the tier-0 decoder and its runner live; if harness, correct 3a's tree column and its worktree eligibility. Add ownership rows for `internal/harness/conform`, `internal/harness/oracle`, and `corpus/analysis/*` covering the E1↔E2 handover.

**M6. Ruling A's expectation source (c) has no machine-readable home, and nobody is assigned to build one.**
Spec:40 says "the corpus already computed these figures for the ten string-heuristic cases." It did — but only as prose inside `adjudications.json`'s `suspected` field (`adj-adv-ft-draft-colon` carries "0.010185185 … 0.101851852" in a sentence). Grep confirms no schema field anywhere carries a defect-free figure. The recommendation adopts (c) "as the expectation source" with no deliverable creating one. (The ten cases do check out: six `adv-ft-*` plus four `adv-remix-*`.)
*Fix:* name the schema addition — a `defect_free` block on the adjudication row, or a per-case file — as a deliverable, and note it writes `corpus/`, which places it in the main tree and likely in 3b rather than 3d.

**M7. The gate-resident conformance gate has no pending state, so one divergence reds every lane with no agent-resolvable path.**
Spec:194 makes the whole-corpus run a `check.sh` leg; risk 5 (spec:296) treats "a new case grades the matcher on the next `check.sh`" as the mitigation. But when Go and the oracle disagree, deciding which is right is owner adjudication (spec:42), and disposition is owner-only. Tiers 1 and 2 admit no pending state, so a corpus case landing from lane E1 that the matcher diverges on hard-reds master for every lane — against "never commit red" and "master is never left red" (`parallel-build-approach.md:131`) — until Geoff rules. The ledgers already solved exactly this: `adjudication.go:8-11`, "Pending is legal at every gate this package runs (`bash scripts/check.sh`); the launch gate, a different gate at a different time, refuses it."
*Fix:* give tier 2 the same escape — a recorded pending divergence passes `check.sh` and blocks the launch gate — and name who owns a red that a corpus-lane commit causes.

**M8. Three tunable surfaces are unreachable under every committed profile, yet the spec grades them.**
`docs/STATUS.md:49-51`: "`match.required`/`match.ignored`, and `preferred.original_year` stay unreachable under every profile, carried since 2b." Spec:235 puts the admission filters (which consume `required`/`ignored`) in 3d, spec:202 lists those surfaces as decoder-supplied, and spec:90/224 demand "every checker proven able to fire." No corpus fixture can prove them.
*Fix:* state the coverage route explicitly — component unit tests against `Tunables`, a fourth profile, or a declared-ungraded surface with the reason recorded — rather than leaving an implementer to discover the gap at 3d.

**M9. The pass split's lane reasoning omits lanes A and T entirely.**
Spec:239 and risk 4 (spec:294) reason only about E1's 2d contending for the main tree. Phase 1 also carries lane A (rung 5) or lane T (rung 8) as a code lane (`parallel-build-approach.md:69-78`). Once 3b–3d take the primary tree, the model's slot assignment (primary tree = primary code lane + all document lanes; worktree = second code lane) has to be restated, and E1's "primary tree as policy (M5)" plus 3b's own `corpus/` and `internal/harness/oracle` writes make three claims on one tree.
*Fix:* state the full phase-1 tree allocation pass by pass for 3a–3d, naming where A/T sit in each configuration.

## MINOR

- **m1. Spend arithmetic understates the base by ~26%.** Spec:241 cites "2c about 6.5M against 8M through Task 17." 2c *closed* at ≈8.2M against 8M with an approved overrun (`docs/STATUS.md:64-65`; `docs/HISTORY.md:207-209`). Sizing 3a's 6M off the mid-pass figure will mislead the conductor. Quote the close figure. The spec also never names the phase-1 combined ceiling its four sub-ceilings must sum into (`parallel-build-approach.md:172`).
- **m2. The 19-vs-20 discrepancy is already resolved by the tree.** Spec:163 is right that Q4:409 says nineteen while Q2:184-188 lists twenty, but it sends the plan to `beets/config_default.yaml` at the pinned tag — outside the repo. Every committed record echoes twenty keys (`corpus/cases/clean-match/case.json`, `expected.weights`), and tier 1 configures from that echo. Resolve it off the tree.
- **m3. "carries no strings at all" is false.** Spec:70. `cassette.Candidate` carries `Source` and `AlbumID`, both strings (`cassette.go:96-102`). The intended claim is "no *display* strings." Ruling E's argument is unaffected; the sentence is checkable and wrong.
- **m4. The adder list omits `AddExpr`.** Spec:106 enumerates `Add, AddEquality, AddNumber, AddPriority, AddRatio, AddString, AddDataSource`. `add_expr` is in the required coverage set (`component.go:23-35`) with two committed rows (`add_expr/0`, `add_expr/1`), so tier 0 will demand a Go primitive the deliverable list does not name.
- **m5. A fourth port hazard belongs beside the three.** The added-key map and the nonzero view are different objects. `derive_oracle.py:176` emits `{key: dist[key] for key in dist._penalties}`, so `Candidate.Penalties` grades key *presence at 0.0* (272 committed candidates carry one), while `max_rec`'s downgrade reads `keys()`, which yields nonzero only (study:327-328, "a 0.0 penalty never triggers a downgrade"). A port conflating them passes tier 1 and mis-downgrades, or the reverse. This is the same invariant spec:133 states in prose; make it a hazard.
- **m6. The second port hazard is not in the cited source.** Spec:137 says "Three port hazards the mechanics study names," but the study never addresses Unicode normalization; it documents `unidecode` transliteration (study:144-151) and nothing about NFC/NFD. Hazards 1 and 3 check out exactly (study:155-157 and study:311-320). Drop the attribution for hazard 2 and mark it an assertion rung 3 verifies against pinned beets.
- **m7. The rung 15 re-run criterion is asserted, not written.** Spec:94 "takes the precedent" of rung 8, whose criterion was actually written down (`parallel-build-approach.md:76`). Nothing yet records rung 3's. Carry it as an explicit forward requirement so the rung 15 plan inherits it rather than rediscovering it.

## Verified sound — do not re-litigate

1. **The five deliverables and their framing** match ladder entry 3 (`greenfield-server-build.md:548-568`) item for item, including the translation layer as a first-class deliverable, the bit-identical target compared at 1e-12, the evaluation date as explicit input, and the limiter's rung-9 re-verification.
2. **The launch-bar restatement** (spec:94) is faithful to the bar (`:60-68`): zero false confident matches over the whole corpus including the stock-config stratum, zero unadjudicated beets-right/Go-wrong cases.
3. **The two architectural invariants.** Harness→engine and never the reverse is correct and matches the committed dependency direction; one package for the scoring core is the right reading of the evolvability ruling (`:192-204`).
4. **The package territory** matches the build spec's tree exactly: `internal/engine/matcher` and `internal/engine/musicbrainz` ("WS client, rate limiter, response translation", `:310-311`). The limiter's home is settled.
5. **Port hazards 1 and 3** are verbatim-correct against the study: lower→transliterate→lower (study:144-157) and strict `<` for strong versus `<=` for medium (study:311-320, and `proximity.go:13-20` cites the same).
6. **The accumulation claims** are all correct: denominator counts only added penalties (study:196-197), nonzero-key ordering by descending weighted distance then key (study:201-202), `add_number`'s multi-unit denominator inflation (study:212-214), the zero-max special case (study:194).
7. **The bit-identity argument** (spec:92) is a faithful reading of study:583-600, including the identical-decimal-literal parsing claim.
8. **The 3→4 edge and ruling B's framing** are accurate against the dependency analysis (`:96-105`, `:507-509`), including "shadow lookups spend deferred background budget through" the limiter.
9. **The corpus-completeness-gates-at-15 claim** is stated in the dependency analysis twice (`:77-81`, `:293-295`), and rung 15's ladder entry already evaluates the conformance gate (`:765-770`).
10. **The two carried facts about the request set** are exactly the README's (`corpus/README.md:2236-2251` on the recorded fixed point and legally orphaned `Extra`; the sub-second cache-hit rule).
11. **`stub.Server`'s described shape** is correct: takes a case plus its directory, 404s and records an uncovered request (`stub.go:54-97`), and `Miss` carries method plus canonical `PathQuery`.
12. **The proximity precedent** (spec:224) is real and quoted accurately — `proximity_test.go:14-18`: hand-authored `testdata/records/*` fixtures that are never corpus cases, with a one-mutation pair proving the flag distinguishes.
13. **Component row floor 72** is exact (`component.go:15`; `components.json` carries exactly 72 rows), and the named coverage (`generic_penalty_keys` exclusion, `string_dist` pipeline, `SD_PATTERNS` rows, the whole data-source parameter space) is present.
14. **The 2d claims all check out** against the tree and the plan: `versionProbe` is five components today (`reconcile.go:19-21`), schema is `"2"` today, and `wantManifestSchemaVersion` lives once, in `internal/harness/corpus_test.go:160`, so "add no second copy" is correct advice.
15. **The verdict-to-action assignment** (spec:165) is right: `cassette.Expected.Action` is graded (`cassette.go:71`), so rung 3 must compute it, and rung 13 owns only the custody policy above it (ladder `:751-757`).
16. **The three committed profiles** are exactly three today: `estate.yaml`, `stock.yaml`, `locale.yaml`.
17. **Review scaling** is correct: rung 3 is not one of the three full adversarial-cycle rungs (5, 12, 16 — `parallel-build-approach.md:112`), and raising 3d's runner as the candidate exception is well-argued.
18. **The staleness pre-task** is correctly motivated and its checklist is the right shape; the 2b precedent is real (`research/2026-09-02-rung-2b-staleness-corrections.md`).
19. **Ruling A's scheduling rule** genuinely dissolves the pending-adjudication deadlock and is consistent with "improvements never ride the port" (`:100-110`).
20. **No contradiction with the rung 10 amendment.** Rung 3 touches no path-format surface, and nothing in it re-imports the killed beets DSL.

## Verdict

**Needs a revision — but a fold, not a rewrite.** The architecture is sound: package territory, the accumulation model, the port hazards that are cited correctly, the three-tier structure, and ruling A's self-scheduling rule all survive adversarial reading, and roughly two thirds of the document is directly checkable against the tree and checks out. What fails is the layer where the spec meets the *committed harness and the ratified lane model*: three blockers would each send an implementer to build something the tree cannot support — a wall registration whose registry does not exist yet in a lane the model has blocked (B1), a tier-1 rule the stub contradicts on every `Extra` file (B2), and a grading row pointing at a field that lives on a different type (B3) — and the major findings cluster in the same seam, where a decoder cross-check cannot fire (M1), an owner-only rule is claimed for a ledger that does not implement it (M4), and the gate has no pending state for the one outcome only Geoff can resolve (M7). All of it is repairable by editing this document plus naming three small harness deliverables (the `Extra` replay channel, the tie-ledger `adjudicator` field, the machine-readable defect-free block), which is a fold of maybe a few hours of drafting. The pass split's shape is right; its lane placements and tree columns need restating against Phase 0 and lanes A/T before 3a's plan is drafted. Fix the blockers and M1–M9, correct the seven minors, and this is ready to carry the rung.