# Rung 2, Pass 2b (Structural Stratum and Component Derivation) Implementation Plan

> **For agentic workers:** workflow mode per CLAUDE.md's "Build
> process" section: pass-execute chain per task, sequential. The
> conductor consumes reports and never reads diffs inline. Subagents
> never dispatch other agents; `code-simplifier` is the conductor's
> dispatch after a task's review accepts. Rung 2 is not an
> adversarial-cycle rung (those are 5, 12, and 16), so review scales
> to the in-chain `diff-reviewer` plus the pass-end gate.

Revision 2, 2026-09-02: folds the adversarial plan review
(`research/2026-09-02-greenfield-spec-reviews/rung-2b-plan-review.md`),
all four blockers, majors 1 through 5 and 7 through 8 as specified, and
major 6 by restructuring the case-authoring work into ten properly
sized, honestly counted tasks rather than splitting the pass into a
2b/2b' pair (an adaptation from the review's preferred fix: drafting a
second just-in-time plan document sits outside this fold's scope, and
the review's own text sanctions in-pass batching — "fixed batches of
at most eight cases with the family boundary respected" — as a valid
resolution when the count each task declares is honest, which is what
the restructuring below delivers). Minors 1 through 10 and 12 are
folded where cheap; minor 11 targets the case-inventory research
document, not this plan, and is left for the conductor to fix
separately. The review's "Verified sound" section records the
mechanisms implementers must not re-litigate: the ID-lookup premise
proven by the smoke case; the threshold/weight gap and its `max_rec`
decode hazard; the forced re-freeze and its schema-version pairing;
the 34-plus-27-equals-61 case-count arithmetic; the search-fallback
loop's shape; record determinism needing nothing added; the tie
detector's scope outside `cassette.Expected`; `incl_artist=False` as
the correct assignment objective; the component tool's shape and
package placement; the epsilon citation; the no-live-network stance;
and the Tier C exclusions.

Consumes pass 2a's shipped machinery exactly as committed: the frozen
`cassette` schema, `internal/harness/stub`, `oracle.Derive` and
`derive_oracle.py`, `cmd/harness`, `make derive`, the two config
profiles, and the seven-case live smoke corpus. 2b adds cases,
tooling, and analyses on top of that substrate; it re-opens the schema
exactly once, additively, and says so below.

**Primary input:** the case catalogue
(`docs/research/2026-09-02-beets-structural-stratum-inventory.md`,
2026-09-02), which enumerates every shape in beets' own
`test/autotag/test_distance.py`, `test_match.py`, and `test_hooks.py`
at the pinned `v2.13.1` commit, sorts them into three tiers, and
budgets 15 shape families at 61 cases. Read it in full before Task 4.
Its tier split is the pass's organizing decision and must not be
re-litigated: Tier A is component-level (no `tag_album`, no candidate
list) and lands as the component derivation tool; Tier B is
full-pipeline and lands as cassette cases; Tier C is inexpressible
under the single-source ruling and stays out.

**Goal:** turn the proven machinery into evidence. Ship the structural
stratum as synthetic cassette cases whose expected records are derived
by executing pinned beets, ship the component-level derivation tool
that pins each distance primitive independently, and ship the three
analyses that keep the corpus honest at its verdict boundaries: the
assignment-tie detector, the threshold-proximity flag, and the
per-case claims checker.

**Owner touchpoints: none.** 2b needs no ruling, no capture contact,
and no live network. Adjudication batches start at 2c, when the
real-catalog and adversarial strata and the known-oracle-bug ledger
arrive; nothing in this pass produces an item for Geoff to adjudicate.
If a task believes it has found one, that is a finding to report, not
a question to ask mid-pass.

**Spec:** `docs/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified), ladder entry 2 (the structural stratum, the
assignment-tie detector, the threshold-proximity flag, and the
component-test derivation tooling are four of its named deliverables)
and rulings 1-3. Ruling 2 (conformance is to beets' behavior except
where beets is documented wrong) and ruling 3 (tier one also holds
over the stock-config stratum) both bind later passes; 2b's job is to
produce the evidence they are measured over, and its cases run under
the `estate` profile unless a family's contract names `stock`.

**Token ceiling:** 6M as a starting figure, re-derived after Task 4
(the first case-authoring task) reports its actual token spend for
family 1; the conductor sets the remaining case-work ceiling from that
measurement rather than treating 6M as fixed in advance (major 8's
fix). **Checkpoint:** STATUS after every task — the case-authoring
tasks are now family-sized batches, so a per-task checkpoint is the
per-family checkpoint the review asked for — and at any split or
question.

---

## Global Constraints

- go-conventions before any Go; python-conventions before any Python.
  Both legs are already gated (`golangci-lint`, `vale` over Go comment
  prose, `ruff` over tracked `*.py`).
- **No live network anywhere in this pass.** The gate never reaches
  the network, and neither does derivation: every 2b case is
  synthetic, so `oracle.Derive` replays only committed files through
  `internal/harness/stub`. `harness capture` is not invoked by any
  task here, and `DUBPLATE_HARNESS_CONTACT` is not a pass input.
- **No 2b case may depend on a failing request.** Every request pinned
  beets issues must resolve to a committed 200 response; the stub has
  no third outcome between a replayed 200 and a recorded miss, so a
  case that wants a request to fail cannot terminate under it. A case
  needing a 404 or a 503 is 2c work, once the schema can carry a
  response status.
- **The cassette surface is frozen, with exactly one additive
  exception** (Task 1): the threshold and weight echo the shim already
  emits and `oracle.Derive` currently drops. No other field is added,
  renamed, or repurposed in 2b. In particular the assignment cost
  matrix does NOT enter `cassette.Expected`: the inventory's section
  (d) establishes that a case's `Input.Items` plus its committed
  `responses/*.json` already carry everything needed to reconstruct
  it, so the tie detector is a tool, not a bigger record. The per-case
  claims block (Task 3) is likewise kept outside the schema: it lives
  in a sibling `claims.json`, never a `cassette.Expected` field, so
  this constraint's "exactly one" holds through the fold.
- **Derivation stays behind the `oracle` build tag and `make derive`.**
  The gate builds no tagged file, runs no Python, and requires no
  venv. `oracle.Derive` itself is untagged (2a shipped it that way, and
  `cmd/harness` links it), so a dev tool may derive; the gate may not.
- **Every expectation is derived by executing pinned beets code
  paths.** The structural cases are synthetic INPUTS, never
  hand-written outputs: an `Expected` record is only ever what
  `oracle.Derive` returned, written through `cassette.Case.Write`.
  This rule scopes to `corpus/cases/**` explicitly. Unit fixtures
  under `internal/harness/analysis/testdata/` are a carve-out: they
  are hand-authored `Expected` documents that exist to fire the
  proximity flag and the tie detector, are never corpus cases, never
  loaded by `corpus_test.go`, and never referenced by a ledger. beets'
  own assertion values in `test_distance.py` are targets a case is
  designed to reach, never numbers copied into a record. A derived
  record that contradicts the beets test it was modeled on is a
  finding to report — written as a row in
  `corpus/analysis/divergences.md` (Task 1 creates the file; case id,
  the beets test or source line the case was modeled on, the derived
  value, and the author's reading), not a value to adjust and not only
  a line in a task return, since a report a compacted transcript loses
  is a divergence nobody can adjudicate.
- **Universality.** Every synthetic case invents its artist, album,
  track, and MBID values. No real person's name, no estate hostname,
  path, or email, nowhere outside `internal/config`. Profiles are
  referenced by name (`estate`, `stock`), never inlined.
- **Determinism.** Every new case re-derives to its committed record
  under `make derive`. Task 1 closes the residual hole that lets an
  all-skipped re-derivation report green, so this is an enforced
  property from the pass's first task onward.
- **Assertion discipline.** No silent-success fakes, no vacuous
  assertions. The three analyses in Task 3 (the proximity flag, the
  tie detector, and the claims checker) each ship a fixture that makes
  them fire, and the task reports the mutation that proved each check
  can fail. A gate-resident check with nothing real to compare against
  is not shipped as if it had teeth (Task 3 names the one place this
  bites).
- `_salvage/` stays read-only; the box is untouched. `case.json` is
  written only through `cassette.Case.Write`; response bodies are
  written verbatim and never re-encoded. A case is authored in
  `corpus/staging/<id>/` (gitignored; Task 2 adds the ignore rule) and
  promoted to `corpus/cases/<id>/` only once its derivation records
  zero misses, its record is frozen, and its `claims.json` and both
  analysis-ledger rows are written — so `corpus/cases/` never holds a
  half-derived case and the gate is never red mid-authoring-loop.
- Gate green and clean `git status` at every task end. Never commit
  red; specific files only; commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`.

## How a synthetic case gets its MusicBrainz responses

Stated once here because every case-authoring task depends on it, and
because the stub replays committed files while the shim asserts a
single metadata source. There is no capture step anywhere in 2b.

1. **Every structural case is an ID-lookup case.** Each input item
   carries an invented `mb_albumid`, so pinned beets issues the direct
   release lookup at its maximal `inc=` and the case needs exactly one
   hand-authored response body,
   `responses/001.json`, a minimal but schema-faithful WS/2 release
   document. `internal/harness/oracle/testdata/cases/smoke/` is the
   working template: its `case.json` request log and its
   `responses/release.json` are the existence proof that a fabricated
   release drives a real derivation end to end.
2. **The search fallback is expected, and answered offline.** A
   deliberately imperfect case (most of families 1 through 15) does
   not score a strong recommendation, so pinned beets falls through to
   its text-search fallback exactly as the live `should-ask` and
   `should-reject` cases did. That request arrives as a miss, not a
   failure: `oracle.Derive` writes `<case dir>/misses.json` with the
   exact canonical `path_query`, and the author writes a response file
   for it by hand. **The default answer is an empty search result** (a
   well-formed WS/2 search document with zero releases), sourced from
   `corpus/cases/should-ask/responses/002.json`'s shape (`created`,
   `count`, `offset`, an empty `releases` array) at the
   `/ws/2/release?...` path with no trailing slash, which keeps the
   candidate set to the single looked-up release, which is what a
   structural distance case wants. **Every request pinned beets issues
   must resolve to a committed 200**; the stub has no failing outcome
   a case can rely on, so family 10's wrong-ID-then-search-recovers
   case is the one place the search response deliberately returns a
   release, and its wrong `mb_albumid` resolves to its own genuinely
   mismatching release document rather than failing — see Task 10 for
   the full shape, which also costs more response files than the
   search alone: beets fetches every search hit with a full release
   lookup, so a search response carrying a hit needs that hit's own
   release document too.
3. **Derive, read misses, author, re-derive, until zero misses, then
   promote.** The same fixed-point loop 2a documented, with
   hand-authoring where the capture tool used to sit, run inside
   `corpus/staging/<id>/` so the gate never sees a half-derived case. A
   case is ready to promote to `corpus/cases/<id>/` when a derivation
   records no misses, its record is frozen, and its `claims.json` and
   both analysis-ledger rows (`ties.json`, `proximity.json`) are
   written — the ledger-and-claims step costs one further
   `derive_ties.py` venv invocation per case, which is a real per-case
   budget line, not a free follow-on.
4. **Single source, always, for cassette cases.** A case exercising
   the `data_source` penalty (family 5) pre-sets the item's own
   `data_source` field. It never adds a second metadata-source plugin;
   that shape is Tier C and stays out of the corpus. Task 3's
   component table is the one exception, and it is not a cassette
   case: `TestDataSourceDistance`'s rows call
   `find_metadata_source_plugins` directly in-process, which is a
   narrower beets entry point that touches no stub and no response
   body.

---

### Task 1: Substrate hardening, the threshold echo, and the freeze path

Model: `sonnet`. Everything downstream depends on this task's freeze
verb, so it runs first and alone.

**Files:**
- Modify: `internal/harness/oracle/corpus_derive_test.go` (the three
  hardenings carried out of pass 2a)
- Modify: `internal/harness/cassette/cassette.go` +
  `cassette_test.go` (the additive echo fields)
- Modify: `internal/harness/oracle/oracle.go` + `oracle_test.go`
  (decode and validate the echo; reopen `Options` with a `Dir` field
  so `deriveRaw` can target a case directory outside
  `corpus/cases/<ID>`)
- Modify: `internal/harness/oracle/testdata/fakeshim` (emit a
  well-formed `thresholds` object and `weights` by default; add
  `FAKESHIM_OMIT_THRESHOLDS`)
- Modify: `Makefile` (widen the `derive` target to
  `go test -tags oracle ./internal/harness/... ./cmd/harness/...` so
  an oracle-tagged test outside package `oracle` is actually compiled
  and run)
- Create: `cmd/harness/freeze.go` + `freeze_test.go`
- Create: `corpus/analysis/divergences.md` (empty at pass end is a
  fine outcome; a missing file is not)
- Modify: the seven `corpus/cases/*/case.json` (re-frozen, never
  hand-edited), `corpus/manifest.json`,
  `internal/harness/corpus_test.go`'s schema-version pin,
  `corpus/README.md`

**Behavior contract:**

- **The three carried hardenings** (HISTORY, pass 2a's "Carried to
  2b/2c"), all in `corpus_derive_test.go`: the skip-all-green residual
  hole (a corpus whose every case skips for want of a committed record
  currently reports green; it must fail, and the count of cases
  actually re-derived is what the test asserts against, not the count
  of directories walked); the unguarded `Candidates[0]` index in
  `TestPinnedDateMovesTheRealYearPenalty` (an empty candidate list
  must fail with a named message, never panic); and the redundant
  equality branch in the same test (the `nearYear <= farYear` check
  subsumes `nearYear == farYear`, and the surviving check keeps the
  directional message). Also close the carried 2a item HISTORY records
  and this plan otherwise drops: the two live cases with response-
  number gaps (documented, harmless) — re-verify they stay harmless
  under the re-freeze below, or note explicitly in the README if they
  do not, so the item does not fall out of the ledger a second time.
- **The threshold and weight echo, additive fields.**
  `derive_oracle.py` already computes and prints `weights` and
  `thresholds`; `oracle.Derive`'s `json.Unmarshal` silently drops both
  because `cassette.Expected` has no field for them, which is the gap
  the inventory's section (d) names. Add them as plain struct fields,
  not `omitempty` ones: `encoding/json` ignores `omitempty` on a
  struct field, and once every record is re-frozen there is no zero
  case left to elide — old records simply decode with zero values
  until re-frozen, and the new validation (below) is what actually
  catches a record still missing them. **Named hazard:** the shim's
  `thresholds` object is not uniformly numeric. Three keys are floats
  (`strong_rec_thresh`, `medium_rec_thresh`, `rec_gap_thresh`) and
  `max_rec` is a mapping of recommendation name to string, so a flat
  `map[string]float64` fails to decode. Pick a shape that decodes the
  real document (a small struct, or split fields), and prove it with a
  decode test over the shim's actual output, not over a hand-trimmed
  sample.
- **`Derive` validates the echo like the others.** A record arriving
  with no thresholds is a loud failure, in the same family as the
  existing echoed-date and config-hash assertions, with a test proving
  it fires — the updated fake shim's `FAKESHIM_OMIT_THRESHOLDS` knob
  is what makes that test possible; without it every fake-shim test
  would break the moment this validation lands, with no way to prove
  the validation itself fires.
- **`harness freeze --case <dir>`**: calls `oracle.Derive` with
  `Options{Dir: <dir>}`, which `deriveRaw` prefers over its existing
  `corpus/cases/<ID>` default — the flag is authoritative for any
  directory, staging or committed. It writes the returned record back
  through `cassette.Case.Write`, leaving `misses.json` handling to
  `Derive` as today. This is the repeatable freezing path 2a never
  shipped (its seven records were frozen ad hoc), and the corpus's
  case-authoring tasks make it load-bearing for every case authored in
  `corpus/staging/`. It refuses to overwrite a committed record when
  the derivation reports misses, since a case with an unanswered
  request is not derived, it is half-derived. Tests cover the write
  path against the fake shim, the miss refusal, and freeze idempotence:
  `Case.Write` re-marshals `Input`, and an integral `ItemFields` float
  loses its `.0` spelling on re-encode, so a freeze-idempotence test
  asserts everything but `expected` is byte-identical across two
  successive freezes of the same already-derived case.
- **Re-freeze the seven 2a records through the new verb.** The echo
  fields change every committed record's content, so
  `TestCorpusReDerivesToItsCommittedRecords` is red until they are
  re-derived. They are re-derived, never edited. Per
  `corpus/README.md`'s own rule, a record-shape change is a
  corpus-revalidation event: bump `manifest.json`'s `schema_version`,
  move `wantManifestSchemaVersion` with it, and record the reason in
  the README's manifest section.

**Deliverable count: 4** (the three hardenings plus the response-
number-gap re-verification; the echo fields with their decode hazard
and `FAKESHIM_OMIT_THRESHOLDS` knob covered; `harness freeze` with its
`Options.Dir` targeting, staging support, and freeze-idempotence test;
the re-frozen smoke corpus with its schema-version bump and the empty
`corpus/analysis/divergences.md` ledger). Gate plus `make derive`
green; commit.

---

### Task 2: The synthetic case builder and the authoring loop

Model: `sonnet`. Produces the tool the ten case-authoring tasks then
use 61 times.

**Files:**
- Create: `cmd/harness/synth.go` + `synth_test.go`
- Create: `internal/harness/synth/` (the release-spec type, the
  document builder, and its tests) if the builder is more than a
  command-local helper; the implementer decides and reports which
  shape it took
- Modify: `.gitignore` (add `corpus/staging/`)
- Modify: `corpus/README.md` (a new "Authoring a synthetic case"
  section)

**Behavior contract:**

- **A compact release spec in, a valid case directory out.** The spec
  names an album, an artist, a release date, a country, a media list
  with per-medium track lists (title, length, position), and the
  handful of flags the families need (various-artists credit,
  per-track artist credits, label and catalog number, disambiguation,
  format). `harness synth --spec <file> --out corpus/staging/<id>`
  emits `case.json` (the input items in pinned order, the profile, the
  eval date, and the ID-lookup request log in canonical form) plus
  `responses/001.json` (the WS/2 release document the log names).
- **The items are derived from the release by default, then
  perturbed.** The common shape across every family is "a correct item
  set with one deliberate divergence", so the spec expresses items as
  the release's own tracks plus an explicit per-item override list
  (wrong title, wrong year, missing item, extra item, wrong
  `data_source`, and so on). A family that needs a wholly independent
  item set can still supply one. `mb_albumid` is fenced off the
  override list outside family 10: beets picks the looked-up release
  id from `current_metadata`'s most-common value across items, so
  perturbing it on enough items changes which release is fetched,
  which is unanswerable under the no-failing-request constraint above.
  An "extra item" override carries the album's `mb_albumid` like every
  other item.
- **Canonical form is produced, never hand-typed.** The emitted
  request log's `path_query` goes through the same canonicalization
  the cassette package enforces, including the `%2B`-escaped `inc=`
  join that cost 2a a full miss cycle. The round-trip test below is
  what proves it, and the wire fact must not be "fixed" back to a
  literal `+`.
- **The authoring loop is documented and exercised.** README gains the
  four-step loop from this plan's "How a synthetic case gets its
  MusicBrainz responses" section, including the empty-search-result
  default, where its document shape comes from, and the
  `corpus/staging/<id>/` working directory (`.gitignore` gains a
  `corpus/staging/` rule here). The tests generate a case into a
  scratch directory, load it through `cassette.LoadCase`, serve it
  through `internal/harness/stub`, derive it via `oracle.Derive` with
  `Options.Dir` (oracle-tagged), and re-derive it to the same record.
  The round trip runs under Task 1's widened `make derive` target;
  Task 2 reports the number of oracle-tagged tests `make derive`
  actually ran, proving they were compiled and executed rather than
  silently skipped by tag or path. A generated case that beets rejects
  for a malformed release document must fail loudly in that test, not
  silently score a distance.

**Deliverable count: 4** (the release spec; the case-directory
emitter; the perturbation overrides with `mb_albumid` fenced off
outside family 10; the generate-load-serve-derive round trip, its
reported oracle-tagged test count, and the README loop). Gate plus
`make derive` green; commit.

---

### Task 3: The tie detector, the proximity flag, the claims checker, and the component derivation tool

Model: `opus` (novel correctness-critical logic: three analyses whose
whole value is that they cannot pass vacuously — the proximity flag,
the tie detector, and the claims checker — plus a second beets entry
point, the component derivation tool). **This is the pass's single
split-risk task.** If the work grows past its budget, split it (3a:
proximity flag, tie detector, claims checker; 3b: component derivation
tool) and say so in the return rather than letting it accrete; do not
absorb case-authoring work from Task 4.

**Files:**
- Create: `internal/harness/analysis/` (the proximity flag, the
  tie-ledger checks, and the per-case claims checker — all pure Go, no
  build tag) + tests + `testdata` fixtures
- Create: `internal/harness/oracle/derive_ties.py` and its
  oracle-tagged Go driver
- Create: `internal/harness/oracle/derive_component.py`, its
  component table, and its oracle-tagged driver + a gate-resident
  schema test
- Create: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`
- Modify: `internal/harness/corpus_test.go` (wire the proximity flag,
  the tie detector, and the claims checker into the gate),
  `corpus/README.md`

**Behavior contract:**

- **The threshold-proximity flag is pure Go and fully gate-resident.**
  With Task 1's echo present, it needs nothing but the committed
  record: best distance `d = Candidates[0].Distance`, gap
  `Candidates[1].Distance - d` when a second candidate exists, flagged
  when `d` sits within ~1e-6 of `strong_rec_thresh` or
  `medium_rec_thresh`, or the gap within ~1e-6 of `rec_gap_thresh`. **A
  record with no candidates is unflaggable and is handled explicitly**:
  a named skip in the ledger, never an index into an empty `Candidates`
  slice — family 14's zero-track candidate release (dropped by
  `_add_candidate`) makes this reachable inside this very pass, and the
  analysis package's `testdata` carries a zero-candidate fixture
  proving the skip fires rather than panics. Thresholds come from the
  record, never from a hardcoded constant: that is what the echo
  bought. The strict-versus-inclusive asymmetry at the two comparison
  sites is already documented in the mechanics study's epsilon-regime
  finding and needs no new research.
- **A flag is not a failure; an unacknowledged flag is.** Family 15
  deliberately parks cases on boundaries, so the gate's contract is
  coverage and acknowledgment: the gate recomputes every case's flag
  from committed records and diffs it against
  `corpus/analysis/proximity.json`, failing on a case flagged but not
  listed, a listed case no longer flagged, and a listed case with no
  written reason. A stale ledger fails the gate.
- **The tie detector splits across the tag boundary, because its raw
  material does.** Determining whether `assign_items`' optimal
  assignment is unique needs the full item-by-track cost matrix at
  `incl_artist=False`, which is a beets computation, so the derivation
  half is `derive_ties.py`. It gets its `TrackInfo` list the way beets
  does, not by re-parsing JSON: it starts the same `stub.Server`,
  renders the same profile, and calls the plugin's `album_for_id` for
  `Expected.Candidates[0].AlbumID` — the tie question is asked of the
  best candidate only, and `Candidates[0].AlbumID` is the selector for
  every case, multi-response cases (`should-ask` carries six) included.
  It then computes `track_distance(item, track, incl_artist=False)`
  over that list, building the cost matrix exactly as `assign_items`
  does, including `lap.lapjv`'s `extend_cost` padding for non-square
  matrices. Uniqueness is a definite, terminating algorithm: solve the
  matrix for optimal cost `z` and assignment `A`; for each pair `(i,
  j)` in `A`, re-solve with `c[i][j]` forbidden (a large sentinel),
  giving `z_ij`; the optimum is unique iff `min over (i,j) in A of
  z_ij > z + 1e-9` (well clear of the 1e-6 proximity regime). Record
  `z`, the best `z_ij`, and their difference in `ties.json`, not only a
  boolean. The gate-resident half is Go: every corpus case appears in
  the ledger, every tied case carries a disposition (`excluded` or
  `pinned`) with a reason, and no case is silently missing. Ladder
  entry 2's "tie cases are excluded or pinned" is exactly this
  contract.
- **The claims checker is the third analysis, and it is what keeps a
  degraded case from freezing green.** Every case that reaches
  `corpus/cases/` carries a sibling `claims.json` (never a
  `cassette.Expected` field, so the frozen-schema constraint stays
  intact), written by the case's author off the derived record, never
  invented: `min_candidates` (default 1) and `best_album_id`; the
  family's `penalty_keys_nonzero` / `penalty_keys_zero`, checked
  against `Candidates[0].Penalties`; `recommendation`; and, for the
  assignment family, `assignment` / `extra_items_len` /
  `extra_tracks_len`; and, for the `per_disc_numbering` cases,
  `applied_fields` spot values. The gate-resident checker re-reads
  every committed case's record and its `claims.json` and fails on any
  mismatch, any case missing `claims.json`, or a claim key no parser
  recognizes. It ships with a fixture proving the check can fail: a
  mutated `claims.json` against a real committed record.
- **The proximity flag and the tie detector each ship a fixture that
  makes them fire**, and the task reports the mutation used to prove
  each check can fail: a record parked exactly on `strong_rec_thresh`
  for the flag, a constructed two-track case whose cost matrix is
  genuinely tied for the detector.
- **The component derivation tool is a second, narrower beets entry
  point.** `derive_component.py` lives in the oracle package (same U7
  boundary: a beets bump edits only this package), starts no stub,
  calls no `tag_album`, and takes no MusicBrainz response at all. It
  reads a table of literal argument tuples mirroring
  `test_distance.py`'s parametrize tables, calls the real
  `Distance.add_*` method, `string_dist`, or `track_distance` per row,
  and emits row id, method, arguments, and observed output as a small
  committed JSON table. It is deliberately NOT a `cassette.Case`: there
  is no request log, no response body, and no rendered profile for a
  call that touches neither stub nor config file.
- **Be honest about what the gate can assert over that table.** The Go
  matcher does not exist until rung 3, so there is nothing at 2b to
  compare a row's output against. The gate-resident test asserts
  schema validity, row-id uniqueness, and non-emptiness; the
  re-derive-and-diff assertion is oracle-tagged and runs under `make
  derive`. Do not dress the gate-resident half up as a conformance
  check it cannot be. Its consumer is rung 3, and the table's own doc
  says so.
- **Table coverage:** every `TestDistance` adder row, the raw/max/
  normalized arithmetic including the zero-penalty-key denominator
  rule, the `generic_penalty_keys` exclusion, `string_dist`'s full
  pipeline (casing, punctuation, article rotation, `&` to `and`,
  bracket and `(EP)` strips, diacritic transliteration), the
  `SD_PATTERNS` relative-weight rows, and the `TestDataSourceDistance`
  parameter space the tool can reach in-process by setting
  `distance_weights` and monkeypatching `find_metadata_source_plugins`
  exactly as beets' own test does. The gate-resident schema test
  additionally asserts a per-method coverage set (every `add_*`
  method, `string_dist`, `track_distance` each appear at least once)
  and a row-count floor, so the table cannot silently shrink between
  rungs.

**Deliverable count: 4** (the proximity flag with its zero-candidate
guard and ledger contract; the tie detector with its concrete
uniqueness algorithm, beets-native `TrackInfo` derivation, and ledger
contract; the component derivation tool with its table and coverage
floor; the per-case claims checker with its mutation-proof fixture).
Gate plus `make derive` green; commit.

---

## Case-authoring tasks (families 1 through 15)

Ten tasks, Task 4 through Task 13, each scoped to one family or a
small group of families whose combined count stays at or under eight
cases, family boundaries never split across a task (the review's major
6 fix). Each task's deliverable count is its real case count, not a
flat "4" — a reviewer reading the count now has a true signal of the
task's size. **Case budget:** 61 across all ten tasks is the floor of
the spec's 60-to-120 band, and it is the number this plan holds. If a
family proves cheaper to author than budgeted, the extension levers
are string-distance edges and the recommendation boundary, in that
order, and any extension is reported rather than absorbed silently. Do
not compress below the per-family counts stated below. Every case in
every task below is generated through `harness synth` into
`corpus/staging/<id>/`, derived through `harness freeze --case
corpus/staging/<id>`, iterated to zero misses per the authoring loop,
and promoted to `corpus/cases/<id>/` with its `claims.json` and both
analysis-ledger rows written — never a hand-written record. A tied
case is dispositioned in `ties.json`, not quietly rewritten. Read the
inventory's section (b) in full before Task 4; its per-family notes
name the beets source lines each family targets.

### Task 4: Structural stratum, family 1 (assignment and ordering)

Model: `sonnet`. **8 cases** (`assign-*`).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 8
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`
  (every new case joins both ledgers on promotion), `corpus/README.md`
  (inventory rows)

**Behavior contract:** scrambled item order corrected by content; an
unmatched leading, an unmatched middle, and an unmatched trailing
candidate track, each becoming `extra_tracks`; an unmatched item
becoming `extra_items`, carrying the album's own `mb_albumid` per Task
2's fencing rule; one deliberately large (10 to 14 track) case whose
titles and lengths are all substantially wrong yet still assigns
correctly, proving the solver's optimum survives large per-pair costs;
and the group's remaining ordering and assignment edges per the
inventory's family 1 row, to reach 8.

**Deliverable count: 8 cases.** Gate plus `make derive` green; commit.
Report actual token spend for this task; it sets the re-derived
ceiling for Tasks 5 through 13.

---

### Task 5: Structural stratum, family 2 (album completeness)

Model: `sonnet`. **7 cases** (`complete-*`).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 7
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** candidate one and two tracks short of the items
(`unmatched_tracks`, weight 0.6); candidate one and two tracks beyond
the items (`missing_tracks`, weight 0.9); the combined shape; and one
case crossing the `max_rec` downgrade boundary, which is where
completeness stops being a distance nudge and becomes a verdict.

**Deliverable count: 7 cases.** Gate plus `make derive` green; commit.

---

### Task 6: Structural stratum, family 3 (various artists and artist handling)

Model: `sonnet`. **7 cases** (`va-*`).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 7
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** VA exact; VA with a mismatched track artist (the
album-level artist component is skipped while the per-track one still
fires); VA with genuinely absent per-track credits on both sides
scoring zero; non-VA correct and wrong artist; and a single-artist
album whose credit string collides with the `VA_ARTISTS` set.

**Deliverable count: 7 cases.** Gate plus `make derive` green; commit.

---

### Task 7: Structural stratum, families 4 and 5 (string distance, data source)

Model: `sonnet`. **7 cases** (`strdist-*`, `datasrc-*`; 5 + 2).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 7
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** the string-distance edges ported to the full
pipeline (article rotation, `(EP)` strip, `feat.` strip, `&` versus
`and`, one diacritic-transliteration pair; family 4, 5 cases); and the
single-source `data_source` penalty, where the item is pre-tagged with
a foreign source and the penalty fires because "before is set" alone
suffices (family 5, 2 cases).

**Deliverable count: 7 cases.** Gate plus `make derive` green; commit.

---

### Task 8: Structural stratum, family 6 (media and disc structure)

Model: `sonnet`. **5 cases** (`media-*`).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 5
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** correct two-medium baseline; disctotal mismatch;
media-format mismatch; uneven disc split; and one further disc-
structure edge per the inventory's family 6 row, to reach 5. The disc
cases additionally assert applied `track` and `tracktotal` under
`per_disc_numbering`; beets' `test_hooks.py` is the reference for what
those applied values should be, and a divergence from it is a finding
to report as a row in `corpus/analysis/divergences.md` (Task 1).

**Deliverable count: 5 cases.** Gate plus `make derive` green; commit.

---

### Task 9: Structural stratum, family 7 (year-penalty branches)

Model: `sonnet`. **4 cases** (`year-*`).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 4
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** the branches the live `year-penalty` smoke case
does not reach, notably a candidate with no `original_year` taking the
flat 1.0 branch, and an item year equal to the release year rather
than the original year. `preferred.original_year: yes` is out of reach
under both shipped profiles and belongs to 2c; do not add a profile
here.

**Deliverable count: 4 cases.** Gate plus `make derive` green; commit.

---

### Task 10: Structural stratum, families 8, 9, and 10 (the richly-tagged candidate)

Model: `sonnet`. **7 cases** (`rich-*`; 2 + 3 + 2).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 7
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** country code mismatch and one-sided presence
(family 8, 2 cases); label, catalog number, and album disambiguation
present on both sides and on one (family 9, 3 cases); the `album_id`
component (family 9); and family 10's 2 cases, the pass's only shapes
whose search response deliberately returns a release. Family 10's
wrong `mb_albumid` resolves to a different, genuinely mismatching
release document (its own `responses/NNN.json`) rather than failing —
every request pinned beets issues must resolve to a committed 200 per
the Global Constraint above. The mismatching release scores poorly,
does not reach a strong recommendation, and lets the search fallback
return the correct release, exercising `match.py:432-442`. That
recovery costs more response files than the search document alone:
beets fetches every search hit with its own full release lookup, so
this case needs at least one further hand-authored release document
beyond the wrong-ID release and the search document itself. The
inventory's compression lever applies across this task: a case may
carry several of these components at once, as long as the group's
contract stays legible from the case id and the README row.

**Deliverable count: 7 cases.** Gate plus `make derive` green; commit.

---

### Task 11: Structural stratum, family 11 (track length)

Model: `sonnet`. **4 cases** (`track-length-*`).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 4
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** `track_length` within grace; beyond grace but
under max; beyond max; and with the candidate length absent.

**Deliverable count: 4 cases.** Gate plus `make derive` green; commit.

---

### Task 12: Structural stratum, families 12 and 13 (track index, track id)

Model: `sonnet`. **5 cases** (`track-index-*`, `track-id-*`; 3 + 2).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 5
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** `track_index` matching absolutely and matching
per-disc (two cases), and mismatched and absent covered together by
one case (a candidate whose items include both a track with an absent
index and a track with a mismatched one) — three cases against the
four shapes the inventory names, the pairing stated here so a later
reader does not need to rediscover it (family 12); and `track_id`
matched and mismatched (family 13, 2 cases).

**Deliverable count: 5 cases.** Gate plus `make derive` green; commit.

---

### Task 13: Structural stratum, families 14 and 15 (candidate admission and the recommendation boundary)

Model: `sonnet`. **7 cases** (`admit-*`, `rec-*`; 2 + 5).

**Files:**
- Create (via `corpus/staging/`, promoted to `corpus/cases/<id>/`): 7
  cases
- Modify: `corpus/analysis/ties.json`, `corpus/analysis/proximity.json`,
  `corpus/README.md`

**Behavior contract:** the two admission filters reachable without
profile config (a duplicate `(source, album_id)` pair inside one
search response, and a zero-track candidate release; family 14, 2
cases — the zero-track release is what Task 3's proximity-flag
zero-candidate guard exercises); and the recommendation and threshold
boundary family, a single-candidate `low`, and an explicit `max_rec`
downgrade distinct from family 2's (family 15, 5 cases). beets ships
no unit test for `_recommendation` or `_summary_judgment` at all, so
this family is written from the source and the config defaults rather
than ported, and it is the group most likely to trip the proximity
flag by design.

**Deliverable count: 7 cases.** Gate plus `make derive` green; commit.

---

### Task 14: The pass close

Model: `sonnet`.

**Files:**
- Modify: `corpus/README.md` (the full case inventory)
- Modify: `corpus/analysis/divergences.md` (reviewed; carried forward
  to 2c's ledger and first adjudication batch, populated or empty)
- Create: the pass's derivation record (the determinism evidence,
  reported in the task return and summarized in the README)
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit)

**Behavior contract:**

- **README inventory.** One table covering all 68 committed cases (the
  7 live smoke cases plus 61 structural), each row naming its id,
  family, profile, and the exact behavior it exercises, in the register
  the existing smoke-corpus table already sets. The structural section
  states plainly that its cases are synthetic inputs with derived
  expectations, and points at the authoring loop.
- **Determinism re-derivation.** Run the committed corpus test over
  the whole corpus (`make derive`) a second time from a clean tree and
  record the result: every case re-derives to its committed record,
  and the count of cases actually re-derived is reported, not merely
  "green" (Task 1's hardening is what makes that number meaningful).
  Report the wall-time cost too, since 68 cases at one shim process
  each is now a real budget line for 2c.
- **Ledger notes for the conductor.** Draft the HISTORY entry in the
  house shape: what landed, what the gate caught, and what a later
  pass would be wrong to rediscover. Carry forward to 2c, explicitly:
  the Tier C exclusions (multi-source ordering and most of the
  `data_source` parameter space, blocked by the single-source ruling);
  `preferred.original_year: yes`, which neither shipped profile can
  reach and which may need a third named profile in 2c's stock-config
  work; `match.required` and `match.ignored` candidate dropping, same
  shape; file-level applied tags, already named as 2c's deliverable
  and rung 8's precondition; synthetic pseudo-release parsing
  variants, better owned by 2c's real-catalog stratum; the two items
  2a carried that 2b deliberately did not take (the line-pinned
  `.gitleaksignore` fingerprint, and git-history gitleaks scanning);
  and `corpus/analysis/divergences.md`'s rows (or its empty state) for
  2c's adjudication batch. Update STATUS's next action to point at 2c.
- **STATUS stays present tense** and under 60 lines; anything
  historical moves to HISTORY rather than being summarized harder.

**Deliverable count: 4** (README inventory; determinism record;
HISTORY draft with the 2c carry list; STATUS next action). Gate plus
`make derive` green; commit.

---

## Execution model

Workflow mode per the standing grant: pass-execute chain, sequential.
The pass is sequential by dependency, not by habit: Task 1 ships the
freeze verb and the echo every later task writes records through, Task
2 ships the builder Tasks 4 through 13 call 61 times, and Task 3's
ledgers and claims checker are joined by every case those tasks
author. Tasks 4 through 13 are the ten case-authoring tasks (one per
family or small family group, per the review's batching fix, each
capped at 8 cases and each declaring its real case count rather than a
flat "4"); none of the ten can run in parallel with another, since all
ten write the same `corpus/analysis/ties.json`,
`corpus/analysis/proximity.json`, and README inventory and would
contend on all three.

Task 3 dispatches on `opus`; the rest on `sonnet`; `diff-reviewer` per
task. The conductor dispatches `code-simplifier` over each accepted
task's changed Go and Python after review, before the pass close.
Runaway guard armed at launch. Budget: 6M ceiling as a starting
figure, re-derived after Task 4 reports its actual per-family token
spend (major 8's fix); spend scored at close against the re-derived
ceiling and against zero attended interactions, which is the target
for this pass.
