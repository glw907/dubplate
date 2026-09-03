# Adversarial review: the rung 2, pass 2c plan (revision 1)

Fresh-context adversarial review of
`docs/superpowers/plans/2026-09-03-rung-2c-catalog-adversarial.md`
(revision 1, 2026-09-03), 2026-09-03. Hostile by default. Every cited
interface was checked against the shipped source at `c126d84`, not
against the plan's or the inventory's prose.

Counts: **2 BLOCKER, 12 MAJOR, 9 MINOR.**

Read alongside `2026-09-02-rung-2b-staleness-corrections.md`, the
proven method this plan must not regress, and
`docs/superpowers/research/2026-09-03-rung-2c-planning-inventory.md`,
whose findings the plan declares non-litigable.

---

## BLOCKER

### B1. A real-catalog case's `Input.Items` cannot be authored under the plan's own no-read constraint, and no tool closes the gap

**Section:** Global Constraints ("A captured response body is evidence
and is never edited, and never read into context"), and "How a
real-catalog case gets made" step 3.

**Defect.** The plan forbids an agent from opening a captured
`responses/*.json` at all ("not opened by an agent to 'see what is in
it'"; "Inspect a captured case through its derived record and its
`claims.json`, never through its `responses/`"). Step 3 then requires
that same agent to hand-author `Input.Items` describing "a plausible
real rip of that release." Those two requirements are incompatible,
and the committed corpus proves it. `corpus/cases/clean-match/case.json`
carries, per item, `mb_trackid`
(`b25f351b-206f-4297-95bd-78459b704144`), `length` to three decimals
(`173.666`), and titles with the release's own typography (`Don’t Stop`,
U+2019). None of that is inventable from an MBID and a genre; all of it
lives only in the response document. `harness capture` writes the
request log and nothing else (`cmd/harness/capture.go:177`: the seeded
case is `&cassette.Case{ID: ..., Requests: requests}`), and there is no
counterpart to `harness synth`'s "items are derived from the release by
default, then perturbed" (`corpus/README.md`, "The synth tool"). The
derived record the plan points agents at does not exist until the items
do.

The constraint's motive is sound: responses average 36 KB and
`corpus/cases/multi-disc/responses/001.json` is 433 KB, roughly 110k
tokens for one case, and the box-set axis quotas 14 more of that shape.
The constraint is right; the plan just leaves 96 cases with no way to
obey it.

**Concrete fix.** Add a Task 2.5 (before the first capture batch)
shipping one new harness verb, for example `harness seed-items --case
<dir>`, that reads the seeded case's `responses/001.json`, projects the
per-track `title` / `length` / `position` / recording id / release-level
`album` / `albumartist` / `date`, and writes a correct baseline
`Input.Items` into `case.json`, reusing `internal/harness/synth`'s
existing derive-then-perturb model plus its `overrides` vocabulary
(`field`, `missing`, `extra`) for the mistagged and partial-rip axes.
Then restate the constraint as what it actually needs to be: a response
body is never read *into agent context*, and is read only by tooling.
Absent that verb, name the sanctioned alternative explicitly (a
committed `jq`/Python projection script whose ~2 KB output an agent may
read) and put it in `corpus/README.md`. Do not dispatch Tasks 3 through
8 until one of the two exists.

### B2. The real-catalog stratum's headline finding class is unreachable by construction, and 16 cases are quotaed on it

**Section:** "Real-catalog capture batches", "A batch's standing
findings duty"; Task 2's classical/compilation quota.

**Defect.** The plan calls a confident (`strong`) recommendation on a
match release truth says is wrong "a tier-one candidate and the
sharpest thing this stratum can find," and Task 2 quotas 16 classical
and compilation cases specifically because "a confident wrong match
here is exactly what tier one exists to catch." A case cannot produce
one. The stub serves exactly the case's own captured responses
(`internal/harness/stub/stub.go:54-62`); a case seeded by `harness
capture --mbid` holds one release document (two with a pseudo-release
merge); and step 3 authors the items *to describe that release*. The
candidate pool is therefore the release the items came from, so the
only reachable verdicts are "matched its own source" or "fell through
to search." A wrong confident match requires the *correct* release to
be absent from, or a *decoy* release present in, a pool the plan never
designs. The classical rationale is worse: MusicBrainz's structural
classical gaps produce a wrong match only for material with no correct
MB entry, which is by definition material this capture loop cannot
seed.

The stratum still has real value, conformance breadth over release
shapes synthetic specs cannot reproduce, and that value is what
justifies 96 cases. The plan just sells it on a finding class its own
method excludes, and lets a quota inherit the false premise.

**Concrete fix.** Two changes to Task 2, before the dossier is
authored. First, restate the stratum's purpose as conformance breadth
over real MB document shapes, and demote the tier-one candidate to what
it honestly is: reachable only on the search-fallback minority. Second,
carve a small explicit sub-axis (6 to 8 cases) of genuine two-release
cases: capture both a decoy pressing and the correct release into one
case, author items from the correct one, and let beets choose. Those,
not the confident-band majority, are where a false confident match can
actually appear. Re-cut the classical quota's rationale accordingly or
fold part of it into the two-release sub-axis.

---

## MAJOR

### M1. Task 2 ships exactly 96 rows and no reserve; Tasks 4 through 8 depend on a reserve

**Section:** Task 2 deliverables; Tasks 4-8 behavior contract.

**Defect.** The six quotas sum to exactly 96 (16 + 14 + 14 + 12 + 14 +
26), and the deliverable count says "96 complete dossier rows." Tasks 4
through 8 then say a case that cannot be authored "is reported and
replaced from the dossier's own reserve." There is no reserve. The same
sentence is also self-contradictory: a replaced case keeps the count at
16, while "the batch's delivered count is stated honestly" anticipates a
short batch.

**Concrete fix.** Raise Task 2's deliverable to 96 primary rows plus 12
reserve rows tagged by axis (roughly one in eight, matching a realistic
capture-convergence failure rate on real data), and rewrite the Task
4-8 clause as an ordered rule: pull the same-axis reserve row first,
report the substitution, and only report a short batch when the axis
reserve is exhausted.

### M2. Task 12's mandatory set exceeds its budget, so the supplementary sample, the thing that gives ruling 3 breadth, is zero

**Section:** Task 2 ("Mark the mandatory stock siblings now"); Task 12
("Membership is deliberate, not a mirror").

**Defect.** Task 2 flags 28 mandatory rows (14 reissue/remaster + 14
multi-disc/box-set). Task 12 budgets 24 and then asks for
"**Supplementary:** fill to 24 with a stratified sample of ordinary
confident matches, which is what gives ruling 3's claim breadth."
28 > 24, so the supplementary sample is empty and ruling 3's breadth
claim gets nothing. Task 2's instruction to "rank them and name which
four are the supplementary-sample donors" is incoherent: dropping four
mandatory rows yields 24 mandatory rows and still zero supplementary
ones.

**Concrete fix.** Set Task 12 at 28 mandatory plus 8 supplementary = 36
siblings, and price the derive cost of the extra 12 against the
schedule (about 22 s per full derive at 1.83 s/case combined). If 36 is
too large for one task, split Task 12 into 12a (the 28 mandatory) and
12b (the 8 supplementary), and delete the "four donors" sentence from
Task 2 entirely.

### M3. `file_tags.json` gets no re-derive-and-diff test, breaking the precedent both other derived ledgers set

**Section:** Task 13 ("The coverage checker is gate-resident and pure
Go").

**Defect.** The corpus's two existing derived ledgers each carry two
checks: a gate-resident structural one and an oracle-tagged
re-derive-and-diff one (`TestTieLedgerMatchesTheCorpus`,
`TestComponentTableReDerivesToItsCommittedTable`, both in
`internal/harness/oracle/`). Task 13 specifies only the gate-resident
half: "Every subject case named in `file_tags.json` is a committed
case; every row's `AppliedFields` pointer resolves; and the subject set
matches a committed list." Nothing ever checks a row's *tag values*
against what beets and mediafile actually write. The ledger becomes a
write-once artifact that cannot go red, which is precisely the
"assertion that cannot fail" class the plan's own Assertion discipline
constraint forbids. It also leaves the `Manifest.MediaFile` pin inert:
a mediafile bump would move the tag shapes with nothing to notice.

**Concrete fix.** Add to Task 13's deliverables an oracle-tagged
`TestFileTagsLedgerMatchesTheCorpus` that re-derives every subject row
and diffs it, plus the `-update-ledgers` regeneration path
`corpus/README.md`'s "Regenerating a derived ledger" section documents,
and say in the README that `file_tags.json` is a derived ledger,
rewritten rather than hand-edited. Report the mutation that proved the
diff fires.

### M4. The FLAC fixture route is a repo-wide setup decision delegated to an implementer

**Section:** Task 13 ("Audio fixtures: decide and report").

**Defect.** The task lets the implementer either adopt `ffmpeg` plus
`flac` as `make derive` prerequisites or vendor a static template blob,
"the implementer decides and reports." That choice adds (or does not
add) two external binaries to what a stranger must install to run the
corpus's derivation, against CLAUDE.md's stated primary product goal of
setup by a stranger, and against `tools/venv/README.md`'s current story
that the venv is the whole derivation prerequisite. Nothing in the plan
gives an acceptance criterion to decide on, and nothing requires the
new dependency be declared anywhere (`doctor`, the venv README, the
Makefile) if it is taken. Plans specify outcomes and constraints; this
is an architecture decision at the wrong altitude.

**Concrete fix.** Rule it in the plan: vendor one small static template
FLAC blob under `internal/harness/oracle/testdata/`, committed, with
its provenance and byte size recorded in the README. `make derive`
gains no binary dependency. If the implementer finds the blob route
unworkable, that is a stop-and-report, not a decision to take inline.

### M5. Task 13's "copy into the existing scratch directory" reads as extending `oracle.Derive`, which would re-freeze every record

**Section:** Task 13 ("Copy into the existing scratch directory") and
its "**No record is re-frozen**" claim.

**Defect.** The paragraph describes `oracle.Derive`'s per-run scratch
dir (`internal/harness/oracle/oracle.go:169-173`) and then says "For
each item needing a file-backed derivation, copy the template bytes to
a fresh path under that scratch dir, point the item's `path` at the
copy." Read literally, that wires file backing into the corpus
derivation path. `derive_oracle.py:218-220` builds `applied_fields` as
`{k: jsonable(v) for k, v in item.items()}`, every field the item
carries, and `jsonable`'s own docstring records that `path` is present
and empty today ("the corpus is file-free, so the only path seen here
is the empty one"). A non-empty `path` therefore changes
`applied_fields` for every case that gets one, and the task's
"**No record is re-frozen**" proof collapses.

**Concrete fix.** State plainly that `derive_file_tags.py` is a second,
independent entry point in the shape of `derive_component.py`: it reads
a committed record's `AppliedFields`, builds beets items from those
fields, allocates its *own* scratch dir, writes and reads back there,
and is never called from `oracle.Derive`. Add an explicit invariant:
no item in the corpus derivation path ever gains a non-empty `path`,
and the proof is that `make derive` re-derives all committed records
byte-identically after the change.

### M6. "Never self-adjudicate" contradicts the dispositions Tasks 1 and 11 are told to write

**Section:** Global Constraints ("Never self-adjudicate ... Do not
write an expected verdict a human has not ruled"); Task 1 (seed
hypothesis 4 at `disposition: "oracle-correct"`); Task 11 ("the row is
`reproduced: "no"` / `oracle-correct`").

**Defect.** `oracle-correct` is one of the four adjudicated
dispositions the same plan defines, and two tasks instruct agents to
write it without a human ruling. The `adjudicator` field has no defined
value for an agent-written row. The result is a ledger in which some
rows carry machine conclusions and some carry Geoff's, with nothing in
the schema separating them, in an artifact whose entire purpose is that
a human ruling is recorded and traceable.

**Concrete fix.** Split the vocabulary. `reproduced` stays a
machine-observable field (`yes`, `no`, `not-attempted`) an agent may
set, with its evidence. `disposition` becomes owner-only, and every row
an agent writes lands at `pending-adjudication`, including hypothesis 4
and Task 11's either-way outcome. Make the gate-resident checker
enforce it: a row whose `adjudicator` is not a human name may not carry
a disposition other than `pending-adjudication`. That check ships with
a fixture proving it fires.

### M7. Task 9 updates "the two hypothesis rows" but only one of its two classes has a seeded row

**Section:** Task 1 ("Seed the four hypothesis rows now"); Task 9
deliverable count.

**Defect.** Task 1 seeds four hypotheses: the featuring-pattern false
positive; non-Latin/pseudo-release/alias mis-selection; per-pair
assignment gaps; threshold-boundary strictness. Task 9 authors two
distinct classes (`adv-ft-*` and `adv-remix-*`, the latter the
`SD_PATTERNS` parenthetical discount, #5868) and claims to update "the
two hypothesis rows." Only the ft row exists. The remix/parenthetical
class has no seeded hypothesis, so four `adv-remix-*` cases would land
either untracked or silently folded into the ft row, and Task 1's own
checker (an `adv-` case in no probe list fails the gate) would force
the fold.

**Concrete fix.** Seed five hypothesis rows at Task 1, splitting the
parenthetical/remix discount (#5868) from the featuring pattern
(`742ebdf69`, `7b59604c5`, `19d091440`), and update Task 1's
deliverable count and Task 9's "two hypothesis rows" accordingly.

### M8. The derive wall-time arithmetic omits beets' hard 0.25 s inter-request floor, which the build spec names as a rung-2 requirement

**Section:** "Wall-time budget, per batch"; Execution model.

**Defect.** The build spec's ladder entry 2 requires "the client's hard
0.25s inter-request floor in differential wall-time budgets." The plan
never mentions it. The mechanics study records the mechanism
(`2026-09-02-beets-matcher-mechanics.md:537-541`: a global
`RateLimitAdapter` floor of 0.25 s between any two requests, applying
"even at a high configured ratelimit"), and `renderProfile` does set a
high one (`ratelimit: 100`, `oracle.go:247`), so the floor is the
binding constraint on every replayed request. The plan's whole schedule
rests on 2b's 1.04 s/case corpus-derive figure, measured over 61
synthetic cases carrying one response each. Real-catalog cases average
1.4 requests, and a search-fallback case runs 6 or 7
(`corpus/cases/should-reject` has 7), so those pay an extra 1.25 to
1.5 s each on every derive, forever, not once. `derive_ties.py` starts
its own stub and pays again.

**Concrete fix.** Restate the per-case derive cost as `1.04 s + 0.25 s
x (requests - 1)`, re-run the batch projections against the dossier's
own expected request mix, and add the floor explicitly to Task 14's
version-bump differential doc, which is the artifact the spec's
requirement actually lands in.

### M9. The 8M ceiling is under-derived, and its largest unmeasured task runs before the re-derivation point

**Section:** "Token ceiling: 8M".

**Defect.** 2b spent about 5.8M on 61 synthetic cases across 14 tasks,
roughly 95k tokens per case. 2c budgets 8M for 142 new cases across 15
tasks, roughly 56k per case, and additionally carries two new ledger
formats with gate-resident checkers, a new Python derivation surface
with audio fixtures, a new profile, a schema bump, two docs, and a
96-row research dossier. The downward adjustment ("a captured case
costs far less design thought than a synthetic one") is asserted, not
measured, and B1 above argues the opposite for the authoring half. The
mitigation is to re-derive the ceiling after Task 3, but Task 2 runs
*before* Task 3 and is plausibly the single largest task in the pass:
96 releases selected by reading musicbrainz.org HTML pages, at even 15k
tokens of page content per release, is 1.4M before search pages and
writing. Finally, the 80% flag ("finish the task, write STATUS, ask one
combined question") is itself a second owner touchpoint, which
contradicts the plan's "**Owner touchpoints: exactly one**" header.

**Concrete fix.** Raise the starting ceiling to 12M, move the first
re-derivation point to after Task 2 rather than Task 3, require Task 2
to report its own spend, and either state the 80% flag as a *contingent
second* touchpoint in the header or route it through STATUS without an
ask.

### M10. Pass sizing: this is two passes

**Section:** whole plan.

**Defect.** Against 2b (14 tasks, 61 cases, one authoring loop, one
stratum), 2c carries 15 tasks, 142 new cases, six live-capture batches,
a second authoring loop, two new ledger formats with gate-resident
checkers, a new config profile, a `cassette` schema bump, an entirely
new beets/mediafile derivation surface with audio fixtures, two chore
docs, and the pass close. CLAUDE.md makes detecting accumulation the
orchestrator's duty, and names the failure mode this plan exhibits:
splitting tasks (Tasks 3 through 8 are one job cut six ways) instead of
splitting the pass, which keeps the work inside it.

**Concrete fix.** Cut after Task 12. Pass 2c carries Tasks 1, 2, 3-8,
9, 10, 11, 12 (the corpus strata) and closes with the adjudication
batch. Pass 2d carries Task 13 (file-level derivation, rung 8's
precondition and the only task needing a new Python surface, audio
fixtures, and a schema bump), Task 14 (the two chore docs), and the
whole-corpus README rewrite. 2d is small, has no network leg, and its
one dependency (seven real-catalog subjects) is satisfied by 2c's
close.

### M11. Task 2's selection mechanism is unnamed, and the facts it needs are exactly what the forbidden API answers

**Section:** Task 2 ("Selection is research, and it stays off the API").

**Defect.** Each dossier row must assert structural facts about a
release: three or more discs, a first-release date genuinely diverging
from the release date, a documented locale-specific alias in the
`artist-credit`. Those are `/ws/2` fields. The plan bans `/ws/2` for
selection and names no tool for the substitute: "Browse
musicbrainz.org's own search and release pages" does not say through
what (WebFetch is named in the same constraint block's prohibition
list, claude-in-chrome is not mentioned), at what pace, or how an agent
confirms an alias exists from a rendered HTML page. The stated
rationale is compliance, but musicbrainz.org's web pages carry the same
courtesy expectations as the API, and 96 HTML page reads is both more
tokens and weaker evidence than 96 API reads would be.

**Concrete fix.** Either name the mechanism concretely (the tool, the
pace, the per-row evidence an agent records) and accept the token cost
in the ceiling, or reopen the constraint: allow `harness capture` to
seed a candidate and let the projection tool from B1 report the
structural facts, treating an unused seeded capture as a discarded
staging directory. The second route reuses the one compliant door the
plan already blesses and produces the item baseline the case needs
anyway.

### M12. A 16-case batch with one commit has no partial-failure rule, and regresses 2b's proven cadence

**Section:** "Real-catalog capture batches (Tasks 3 through 8)".

**Defect.** 2b's case tasks averaged about six cases each. Each 2c
batch is 16, promoted and committed as one unit, with "Gate plus `make
derive` green; commit" and "clean `git status` at every task end." The
plan never says what happens when case 12 of 16 will not converge: the
tree then holds 11 promoted cases, partially regenerated `ties.json`
rows, and possibly a `proximity.json` exception row owed, with the
task's stated deliverable unmet and its commit unwritten. `make derive`
and `-update-ledgers` are O(all committed cases), so re-running after
each recovery attempt is the expensive operation, which is exactly the
argument for smaller commit units, not larger ones.

**Concrete fix.** Either cut the batches to eight cases each (12 tasks,
which also pairs with the M10 split), or keep 16 and add an explicit
partial-batch rule: a batch commits whatever converged, states the
delivered count, returns the failed rows to the dossier, and the
conductor decides whether a follow-on task picks them up. Say which
before dispatch, not during.

---

## MINOR

### m1. `profile_hashes` is an empty map nothing computes or checks

**Section:** Task 11 ("`manifest.json`'s `profile_hashes` gains an
entry rather than changing one").

`corpus/manifest.json` carries `"profile_hashes": {}`. A grep across
the tree finds `ProfileHashes` in exactly two places, both in
`cassette.go` (the struct field and a doc comment); nothing writes it,
and `LoadManifest` and `Reconcile` both ignore it. `corpus/README.md`'s
config-profile section already overclaims it ("`corpus/manifest.json`'s
`profile_hashes` is what detects it"), and the plan inherits the
overclaim as a safety argument for adding `locale.yaml`. **Fix:** have
Task 11 either populate and reconcile all three profile hashes (cheap,
and it makes the README true) or record the gap as a named carry-forward
in HISTORY, and drop the safety argument from the task's contract.

### m2. `Reconcile` is not in `oracle.go`, and it pins five components, not four

**Section:** Task 13 Files list and behavior contract.

`Reconcile` and its hardcoded `versionProbe` tuple live in
`internal/harness/oracle/reconcile.go:21-48`, and the tuple names five
components (beets, unidecode, jellyfish, lap, numpy), not "the existing
four pins." **Fix:** correct the Files list to `reconcile.go`, say five,
and name `versionProbe` explicitly as the second edit, since adding
`mediafile` to `Manifest` without adding it to the probe string yields a
reconcile that silently compares against a missing key.

### m3. A stock sibling must change its `id`, and Task 12 never says so

**Section:** Task 12 ("Sibling case directories").

`TestEveryCorpusCaseLoadsClean` (`internal/harness/corpus_test.go:87-90`)
fails any case whose `case.json` `id` differs from its directory name.
The task says a sibling gets "its own `case.json`, its own `responses/`
... `profile: "stock"`" and never mentions `id`. **Fix:** state that the
sibling's `id` is `<id>-stock`, matching its directory.

### m4. No `EvalDate` rule for captured cases

**Section:** "How a real-catalog case gets made" step 3.

`EvalDate` is called "a human's job" with no rule. The committed corpus
already carries two values (45 cases at `2026-09-02`, 23 at
`2026-09-03`), and the date is matcher input through both year-penalty
branches. **Fix:** pin one rule, either "the case's capture date" (the
smoke corpus's convention, per `corpus/README.md`) or one date for the
whole pass, and say which.

### m5. `corpus/README.md` never learns the captured-case authoring loop

**Section:** Tasks 3 and 15.

The README documents "Authoring a synthetic case" and, under "The
capture tool," `harness capture --mbid <release> --out
corpus/cases/<id>`. The plan mandates the staging-then-promote path
instead, and adds 96 cases through a loop the README does not carry.
2b's precedent is the opposite (Task 4 created the structural-inventory
section). **Fix:** add "Authoring a captured case" to Task 3's
deliverables and correct the capture section's `--out` path to
`corpus/staging/<id>`.

### m6. The `adv-` prefix coupling leaves Task 11's probe cases outside the checker

**Section:** Task 1 (checker: an `adv-`-prefixed case in no probe list
fails).

Task 11's two probe cases are `real-nonlatin-<slug>` and its `-locale`
sibling. They probe hypothesis 2 and are not `adv-`-prefixed, so the
"every adversarial case is claimed by a hypothesis" check does not
cover them. **Fix:** drop the prefix coupling in favor of the ledger's
own direction (every case id in a probe list resolves), or add a
`stratum` field to the case-level claims the checker can key on.

### m7. Corpus byte growth is unpriced

**Section:** Tasks 3-8 and 12.

The committed corpus holds 114 response files totaling 4.0 MB, with a
433 KB maximum. The plan adds roughly 134 captured responses (96 cases
at about 1.4) weighted toward the largest shapes it quotas (14 box
sets), plus 24 to 36 stock siblings that duplicate their twin's bytes
verbatim, since no cross-case reference mechanism exists. A 10 to 15 MB
committed corpus is the plausible landing point. **Fix:** state the
expected growth in the plan, and have Task 15 report the actual figure
alongside the case count.

### m8. "roughly 152 new case units" does not reconstruct

**Section:** Token ceiling paragraph.

The tasks sum to 142 new cases (96 + 10 + 10 + 2 + 24), or 154 counting
Task 13's 12 file-tag subject rows, which are readings of existing
cases rather than new ones. **Fix:** use 142 and say what a case unit
is.

### m9. The "no `curl`, no `WebFetch`" prohibition reads as absolute

**Section:** Global Constraints, second bullet.

The bullet scopes itself to `/ws/2` in its first sentence and then
lists absolute prohibitions, while Task 2 needs musicbrainz.org web
pages and Task 11 needs `gh issue view 6358`. A subagent reading the
dispatch literally, which the standing grant says they do, may refuse
both. **Fix:** rewrite as "no non-`harness capture` route to `/ws/2`,
by any tool," and list the two sanctioned non-`/ws/2` network reads
(Task 2's web pages, Task 11's `gh`) in the same bullet.

---

## Verified sound

Mechanisms the plan cites that exist and behave as claimed, checked
against source:

- **`harness capture`'s two verbs and their flags.** `--mbid` with
  `--out`, `--case`, mutual exclusion, and the
  `DUBPLATE_HARNESS_CONTACT` refusal are all as described
  (`cmd/harness/capture.go:68-128`). The maximal `inc=` release lookup
  and the backward `transl-tracklisting` pseudo-release follow-up are
  real (`releaseRequest`, `pseudoReleaseFollowUp`), as is the 1 req/s
  pacer that never delays the first request. Response bodies are written
  verbatim with `os.WriteFile`, so the "never edited" constraint has a
  mechanism behind it.
- **`harness freeze --case <dir>`** derives and writes back, and
  refuses to write when `misses.json` survives the run
  (`cmd/harness/freeze.go:59-82`), which is what makes the plan's
  "iterate to zero misses" loop a real gate rather than a convention.
  `Options.Dir` is authoritative over the `corpus/cases/<ID>` default,
  so deriving from `corpus/staging/<id>` works as the plan assumes.
- **The staging directory is gitignored** (`/corpus/staging/`), so the
  plan's staging-then-promote loop is compatible with its own "clean
  `git status` at every task end" requirement.
- **The promotion ordering and ledger-membership rules** are folded
  correctly from 2b's corrections, including `proximity.json` as an
  exception ledger and `ties.json` dispositions surviving
  `-update-ledgers`. Both match `corpus/README.md`'s "The analyses" and
  "Regenerating a derived ledger" sections.
- **The #5458 claim.** `corpus/analysis/proximity.json` does carry
  `rec-strong-medium-boundary` as `flagged` on `strong_rec_thresh` with
  a written reason, so "the proximity ledger already carries it" is
  true and the decision to author no new case for that class is
  correct. (The README's line saying the ledger is empty is stale; the
  plan is right and the README is wrong.)
- **The single-`Profile` finding and the sibling-directory conclusion.**
  `cassette.Case.Profile` and `.Expected` are single fields
  (`cassette.go:44,48`), `LoadCase` requires a non-empty profile, and
  `renderProfile` renders exactly one per derivation. No cross-case
  response reference exists, so copying bytes is the only route. The
  plan's layout follows.
- **The miss-file mechanism as the stock sibling's escape hatch.** The
  stub 404s and records any uncovered request
  (`stub.go:81-91`), `Derive` writes `misses.json` regardless of shim
  outcome, and `harness capture --case` consumes it. A stock verdict
  driving beets to a search the estate capture never triggered is
  handled exactly as claimed.
- **Task 13's five committed subject cases all exist** (`clean-match`,
  `media-uneven-split`, `va-exact`, `non-latin-pseudo-release`,
  `year-reissue-ratio-penalty`), and 5 + 7 = 12 is consistent.
- **The `AppliedFields`-is-not-a-file-tag-set premise.** The shim's own
  `applied_fields` docstring names the file-level tag set as pass 2c's
  deliverable, so the plan is executing a scope decision the code
  already records.
- **The response-status non-adoption.** All four seeded bug classes are
  distance or selection defects, the stub has no third outcome, and
  carrying the extension forward as a named non-adoption is the right
  call rather than a gap.
- **Stratum proportions against the build spec.** The spec asks for 150
  to 300 cases at about 40% structural, 50% real catalog, 10%
  adversarial. The plan lands near 210 cases at 32% / 46% / 10%, with
  the stock stratum as a separate axis. Within tolerance and defensible.
- **The sequential-dependency argument for Tasks 3 through 8.** All six
  write `ties.json`, `proximity.json`, and the README inventory, so
  parallelism would contend on all three. Correct, and correctly
  reasoned rather than asserted.
- **Grouping the adjudication batch by hypothesis** so one ruling covers
  a class. That is what keeps roughly 22 pending rows inside CLAUDE.md's
  "small and regular" bar for the owner touchpoint.
- **The gate stays offline and venv-free.** Nothing the plan adds moves
  a network or Python dependency into `scripts/check.sh`, and both new
  Task 1 checkers are pure Go over committed files, matching the three
  existing analyses.
