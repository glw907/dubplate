# Adversarial review: the pass 2b plan (structural stratum)

Fresh-context hostile review of
`docs/superpowers/plans/2026-09-02-rung-2b-structural-stratum.md`
(revision 1, 2026-09-02), read against pass 2a's plan, the structural
stratum inventory, the matcher mechanics study, the ratified build
spec's ladder entry 2, and the code as shipped at commit `3112877`:
`internal/harness/cassette/cassette.go`,
`internal/harness/oracle/oracle.go`,
`internal/harness/oracle/derive_oracle.py`,
`internal/harness/oracle/corpus_derive_test.go`,
`internal/harness/oracle/derive_test.go`,
`internal/harness/oracle/testdata/fakeshim`,
`internal/harness/stub/stub.go`, `internal/harness/corpus_test.go`,
`cmd/harness/capture.go`, `Makefile`, `corpus/`.

Counts: **4 blockers, 8 majors, 12 minors.**

Every interface the plan cites was checked against the shipped source.
Most of them exist as claimed; the exceptions are blocker 1 and
blocker 4.

---

## Blockers

### BLOCKER 1 — `oracle.Derive` cannot derive a case outside `corpus/cases/<ID>`, which both `harness freeze` and the Task 2 round trip require

**Section:** Task 1 (`harness freeze --case <dir>`), Task 2 (the
generate-load-serve-derive round trip), Task 4/5 (the authoring loop).

**Defect.** `Derive` does not take a case directory. It resolves one
itself through an unexported package var:

```go
var caseDir = func(c *cassette.Case) (string, error) {
    root, err := repoRoot()
    ...
    return filepath.Join(root, "corpus", "cases", c.ID), nil
}
```

`deriveRaw` uses that directory for two things the plan depends on:
`stub.New(c, dir)` reads the response bodies from it, and
`writeMissFile(dir, ...)` writes `misses.json` into it. The var is
substitutable only from inside package `oracle` (`derive_test.go`'s
`smokeCase` does exactly that). `cmd/harness/freeze.go` is package
`main` and cannot reach it, so `harness freeze --case <dir>` silently
ignores its own flag for any directory that is not
`corpus/cases/<c.ID>`. The same wall blocks Task 1's
`freeze_test.go` ("tests cover the write path against the fake shim",
which needs a scratch case dir) and Task 2's round-trip test, wherever
that test lives outside package `oracle`.

The second half of the defect is that `corpus/cases/<ID>` is not a
legal home for an in-flight case either.
`internal/harness/corpus_test.go`'s `TestEveryCorpusCaseLoadsClean`
errors when `c.Expected == nil`, so the moment `harness synth` writes a
case under `corpus/cases/`, the gate is red until that case is fully
derived and frozen. The plan therefore requires the case to live in
`corpus/cases/<ID>` (because `Derive` looks nowhere else) and forbids
it from living there (because the gate rejects an underived case).

**Failure scenario.** Task 1's implementer writes `freeze.go`, points
`--case` at `t.TempDir()/some-case` in `freeze_test.go`, and `Derive`
tries to read responses from `<repo>/corpus/cases/some-case`, which
does not exist. `stub.New` fails with `stub: read response ...: no such
file or directory`. The implementer's likely repair is to make the
test write into the real corpus tree, which then trips the gate's
`Expected == nil` check.

**Fix.** Reopen `oracle.Options` (not `cassette`, so the frozen-schema
constraint is untouched) with an explicit directory:

```go
type Options struct {
    Venv string
    Shim string
    Dir  string // the case's own directory; empty means corpus/cases/<ID>
}
```

Have `deriveRaw` prefer `opts.Dir` and fall back to the existing
`caseDir` var. State in Task 1 that `harness freeze --case <dir>`
passes `Dir: <dir>` and that the flag is authoritative. Then give the
authoring loop a staging directory outside `corpus/cases/` (for
example `corpus/staging/<id>/`, gitignored) and state that a case moves
into `corpus/cases/` only once its record is frozen and its ledger rows
are written, so the gate is never red mid-loop.

---

### BLOCKER 2 — nothing verifies a case exercises the behavior its family claims; a silently degraded case freezes green

**Section:** Global Constraints ("Every expectation is derived by
executing pinned beets code paths"), Task 4 and Task 5 behavior
contracts, Task 6 (the README inventory).

**Defect.** The corpus's only behavioral check is
`TestCorpusReDerivesToItsCommittedRecords`, which asserts that today's
venv reproduces the committed blob. That proves reproducibility. It
proves nothing about *what* was reproduced. Combined with the plan's
correct rule that a record is only ever what `Derive` returned, this
makes every case in Tasks 4 and 5 a test that cannot fail: whatever
beets scores becomes the expectation, and the claim that the case
exercises `unmatched_tracks`, or the VA short-circuit, or the
`data_source` penalty, lives only in a README row.

This is not hypothetical, because the degradation path is short and
quiet. Every structural case is an ID-lookup case whose request log is
written by `harness synth`, not by observing beets. If synth's
`inc=` value or its escaping disagrees by one byte with what pinned
beets requests (the exact defect that cost 2a a full miss cycle and is
recorded in HISTORY as "do not fix it back"), the release lookup 404s,
`album_for_id` returns nothing, beets falls through to the text search,
the author answers the search miss with the plan's default empty
document, the derivation reports zero misses, and the case freezes with
`candidates: []`, `recommendation: none`. The loop terminates. The gate
is green. The README says the case exercises the `mediums` component.

The same hole swallows the plan's own instruction that "the disc cases
additionally assert applied `track` and `tracktotal` under
`per_disc_numbering`". There is no mechanism in the corpus for a case
to assert anything.

**Failure scenario.** Task 4 lands `media-uneven-disc-split`. Its
`case.json` freezes with an empty candidate list because the ID lookup
never matched. `make derive` is green, `make check` is green, the
README row reads "uneven disc split across two media", and rung 3's Go
matcher is graded against a case that scores nothing.

**Fix.** Add a per-case claims block, checked gate-resident over the
committed record, and make it a Task 3 deliverable (it is the third
analysis, and it is worth more than either of the two the plan already
carries). Minimum shape, in `case.json` or a sibling `claims.json`:

- `min_candidates` (default 1) and `best_album_id` equal to the
  release the case was synthesized from, which kills the degradation
  path outright.
- `penalty_keys_nonzero` and `penalty_keys_zero`: the components the
  family exists to move, checked against
  `Expected.Candidates[0].Penalties`.
- `recommendation`, and optionally `extra_items_len` /
  `extra_tracks_len` / `assignment` for the assignment family.
- `applied_fields` spot values for the `per_disc_numbering` cases.

Every value in the block is read off the derived record by the author
(never invented), but writing it down converts "the record is what
beets said" into "the record is what beets said AND it is the shape
this case was built to produce". Prove the check can fail by mutating
one committed claim.

---

### BLOCKER 3 — the authoring loop cannot terminate for any case that needs a request to fail, and family 10 is written as exactly that case

**Section:** "How a synthetic case gets its MusicBrainz responses",
steps 2 and 3; Task 5 deliverable 2 (the wrong-ID-then-search-recovers
case).

**Defect.** `stub.Server.ServeHTTP` has two outcomes: replay a logged
response with 200, or record a `Miss` and answer 404. There is no third
state. A case that *wants* a request to fail has no way to say so: the
failing request is recorded as a miss on every derivation, forever, so
the plan's termination criterion ("a case is done when a derivation
records no misses") is unreachable. The cassette surface is frozen for
this pass except for the threshold echo, so a per-request status field
is not available as a repair.

Task 5's family 10 case is described as "the wrong-ID-then-search-
recovers path". The natural reading, and the one an implementer will
take, is that the wrong ID does not resolve. Under this stub it cannot
be expressed.

**Failure scenario.** The implementer authors items carrying a
fabricated wrong `mb_albumid`, derives, sees a miss for
`/ws/2/release/<wrong-id>?...`, has nothing sensible to write for it,
and either loops forever, or hand-writes an `Expected` record (which
the plan forbids), or quietly drops the case and leaves family 10 at
one case, breaking the 61 arithmetic without a report.

**Fix.** Add a Global Constraint: **no 2b case may depend on a failing
request; every request pinned beets issues must resolve to a committed
200 response.** Then respecify family 10's case: the wrong `mb_albumid`
resolves to a *different, genuinely mismatching* release document (its
own `responses/NNN.json`), which scores poorly, does not reach a strong
recommendation, and lets the search fallback return the correct
release. That path is expressible with 200s only and still exercises
`match.py:432-442`. Note in the same constraint that a case needing a
404 or a 503 is 2c work, once the schema can carry a response status.

---

### BLOCKER 4 — the oracle-tagged tests Task 2 and Task 3 rely on would never run

**Section:** Task 2 ("the tests generate a case, load it through
`cassette.LoadCase`, serve it through `internal/harness/stub`, derive
it (oracle-tagged), and re-derive it to the same record"), Task 3 (the
component table's "re-derive-and-diff assertion is oracle-tagged").

**Defect.** The `derive` target is scoped to one package:

```make
derive:
	go test -tags oracle ./internal/harness/oracle/...
```

Task 2's files are `cmd/harness/synth.go` and `internal/harness/synth/`.
A `//go:build oracle` test in either package is built by nothing:
`make check` excludes it by tag, `make derive` excludes it by path. It
is not a weak test, it is a test that never executes, and the plan
names it as the proof that synth's canonical form matches what beets
requests, which is precisely the byte-level agreement that failed in
2a.

**Failure scenario.** Task 2 reports the round trip green because
`make derive` printed `ok`; the tagged test was never compiled. The
canonicalization disagreement surfaces in Task 4 as 34 mysterious
misses.

**Fix.** Either extend the target to
`go test -tags oracle ./internal/harness/... ./cmd/harness/...`, or
require the tagged round-trip test to live in package `oracle` (which
also lets it substitute `caseDir` and sidesteps blocker 1's test half).
State the choice in the plan rather than leaving it to the
implementer, and have Task 2 report the number of tagged tests
`make derive` actually ran.

---

## Majors

### MAJOR 1 — Task 1's file list omits `testdata/fakeshim`, which the new validation immediately breaks

**Section:** Task 1, "Derive validates the echo like the others."

`internal/harness/oracle/testdata/fakeshim` emits a fixed document with
no `thresholds` and no `weights`. The moment `Derive` rejects a record
with no thresholds, every fake-shim test in `oracle_test.go` fails, and
there is no knob to prove the new validation fires (the shim has
`FAKESHIM_DATE`, `FAKESHIM_OMIT_HASH`, `FAKESHIM_CALL_LOG`,
`FAKESHIM_ENV_LOG`, and nothing else).

**Fix.** Add `internal/harness/oracle/testdata/fakeshim` to Task 1's
file list: emit a well-formed `thresholds` object (three floats plus
the `max_rec` mapping) and a `weights` object by default, and add
`FAKESHIM_OMIT_THRESHOLDS` so the new validation has a test that can
fail, in the same family as `FAKESHIM_OMIT_HASH`.

### MAJOR 2 — the proximity flag indexes `Candidates[0]` unguarded, the exact defect Task 1 is fixing one task earlier

**Section:** Task 3, "The threshold-proximity flag is pure Go and fully
gate-resident."

The contract reads "best distance `d = Candidates[0].Distance`" with no
empty-list branch. Zero-candidate records are reachable in this very
pass: family 14's zero-track candidate release is dropped by
`_add_candidate`, and the shim already handles `best = None`
(`derive_oracle.py` l.261-268). A gate-resident test that panics is
worse than one that fails.

**Fix.** State the contract: a record with no candidates is
unflaggable and must be handled explicitly, with a named skip in the
ledger rather than an index. Add a zero-candidate fixture to the
analysis package's testdata.

### MAJOR 3 — Task 3's fixtures contradict the plan's own derived-only rule

**Section:** Global Constraints ("an `Expected` record is only ever
what `oracle.Derive` returned") versus Task 3 ("A record parked exactly
on `strong_rec_thresh` for the flag").

A record whose best distance is exactly 0.04 cannot be derived on
demand; it has to be hand-written. As drafted, the plan forbids the
fixture it requires.

**Fix.** Scope the derived-only rule to `corpus/cases/**` explicitly,
and add a carve-out: unit fixtures under
`internal/harness/analysis/testdata/` are hand-authored `Expected`
documents that exist to fire the analysis and are never corpus cases,
never loaded by `corpus_test.go`, and never referenced by a ledger.
Say so in Global Constraints so a reviewer does not have to
re-litigate it.

### MAJOR 4 — the tie detector's uniqueness test is under-specified to the point of being unimplementable

**Section:** Task 3, "The tie detector splits across the tag boundary."

"Reports uniqueness by perturbing a near-tied entry and re-solving with
`lap.lapjv`" names neither which entry, nor the perturbation
magnitude, nor the rule that turns the re-solve's outcome into a
verdict, nor a termination argument. The inventory offers a second,
different heuristic ("duplicate row/column minima"), which is not
equivalent and is not sound as a uniqueness test. A subagent will
invent one of these and report it as done.

**Fix.** Specify a definite, terminating algorithm. The standard one
costs `n` extra solves and is exact:

1. Solve the cost matrix `C` as `assign_items` does, giving optimal
   cost `z` and assignment `A`.
2. For each pair `(i, j)` in `A`, re-solve with `c[i][j]` forbidden
   (set to a large sentinel), giving `z_ij`.
3. The optimum is unique iff `min over (i,j) in A of z_ij > z + eps`,
   with `eps` stated in the plan (1e-9 is well clear of the 1e-6
   proximity regime).

Record `z`, the best `z_ij`, and their difference in `ties.json`, so a
later reader can see how close the case sat rather than only a boolean.
State that the tool must build its matrix exactly as `assign_items`
does, including `lap.lapjv`'s `extend_cost` padding for non-square
matrices; a differently padded matrix is a different problem.

### MAJOR 5 — "re-parses a case's committed response into `TrackInfo` objects" opens a second, divergent parse path

**Section:** Task 3, the tie detector's derivation half.

The corpus's whole premise is that one parse path (beets' own
MusicBrainz plugin) turns committed bytes into `AlbumInfo`. A
hand-rolled JSON-to-`TrackInfo` reader in `derive_ties.py` is a second
path that can silently disagree with the first, most obviously on the
pseudo-release merge (two documents, `_merge_pseudo_and_actual_album`)
and on artist-credit joinphrases. The instruction is also ambiguous for
multi-response cases (`should-ask` carries six) and never says which
candidate's track list is the subject.

**Fix.** Have `derive_ties.py` obtain its `TrackInfo` list the way
beets does: start the same `stub.Server`, render the same profile, and
call the plugin's `album_for_id` for
`Expected.Candidates[0].AlbumID`, then compute
`track_distance(item, track, incl_artist=False)` over that list. State
that the tie question is asked of the best candidate only, and name
`Candidates[0].AlbumID` as the selector.

### MAJOR 6 — Tasks 4 and 5 are 34 and 27 hand-authored cases wearing a "deliverable count: 4" label

**Section:** Task 4 and Task 5 headers; Execution model.

Both tasks declare four deliverables. Both are, in fact, dozens of
independent fixtures, each requiring a hand-written WS/2 release
document, a synth spec, one or more derive-and-answer iterations, two
ledger rows, and a README row. The declared count defeats the very
guard it is quoting: a reviewer reading "4" has no signal that Task 4
is larger than Tasks 1 through 3 combined. No single implementer
dispatch holds 34 cases; the dispatch will run out of context and
either truncate the family or degrade quality across the tail, and the
degradation is invisible under blocker 2.

This is also the shape CLAUDE.md names as a failure mode: splitting the
case work into two tasks keeps it inside one pass, which feels like
discipline while changing nothing.

**Fix.** One task per family (15 tasks), or fixed batches of at most
eight cases with the family boundary respected, and split the pass at
the Task 3/Task 4 seam: 2b ships the substrate, the tools, and the
ledgers; 2b' ships the 61 cases against a frozen toolchain. State the
cut point, what each half carries, and the follow-up number, per the
sizing rule. If the pass is not split, at minimum stop declaring "4"
for a 34-item task and declare the real number.

### MAJOR 7 — the plan anticipates oracle divergences three times and ships nowhere to record them

**Section:** Global Constraints ("A derived record that contradicts the
beets test it was modeled on is a finding to report"), Owner
touchpoints ("If a task believes it has found one, that is a finding to
report"), Task 4 deliverable 4 ("a divergence from it is a finding to
report").

The known-oracle-bug ledger is 2c's deliverable, so a divergence found
in 2b has no artifact. A task return is not an artifact: it is consumed
by the conductor and lost to compaction, and CLAUDE.md's own rule is
that an unadjudicated divergence blocks the launch gate. A divergence
that exists only in a report is an unadjudicated divergence nobody can
find.

**Fix.** Create `corpus/analysis/divergences.md` (or
`docs/superpowers/research/2026-09-02-corpus-divergences.md`) in Task
1, with a stated four-field row shape: case id, the beets test or
source line the case was modeled on, the derived value, and the
author's reading. Task 6 carries the file forward to 2c's ledger and
adjudication batch. An empty file at pass end is a fine outcome; a
missing file is not.

### MAJOR 8 — the ceiling and the checkpoint interval are not derived from this pass's work

**Section:** header ("Token ceiling: 6M. Checkpoint: STATUS every two
tasks (after Task 2 and Task 4)").

2a spent about 1.75M subagent tokens against a 5M ceiling for five
tasks, seven live cases, and all the machinery. 2b carries a schema
reopen, three new tools, two ledgers, and 61 hand-authored fixtures,
and 6M is only a 20% increase over 2a's ceiling for several times the
bulk. More importantly the checkpoint interval is measured in tasks
while the work is concentrated inside two of them: "after Task 2 and
Task 4" means the opus task and the 34-case task each run to completion
with no STATUS, which is the opposite of where a checkpoint is worth
buying.

**Fix.** Re-derive the ceiling from a measured number: have Task 4
report actual tokens for its first family, and set the case-work
ceiling from that. Move the checkpoint from "every two tasks" to "after
every task, and after every family (or eight-case batch) inside the
case tasks". If the pass splits per major 6, give each half its own
ceiling.

---

## Minors

1. **A 2a carried item is silently dropped.** HISTORY's "Carried to
   2b/2c" list names four things: the three `corpus_derive_test.go`
   hardenings (Task 1 takes them), the line-pinned `.gitleaksignore`
   fingerprint and git-history gitleaks scanning (Task 6 carries them
   to 2c), and **"response-number gaps in two cases (documented,
   harmless)"**, which appears nowhere in the plan. Either close it in
   Task 1's re-freeze or add it to Task 6's carry list; do not let it
   fall out of the ledger.

2. **`omitempty` is a no-op on a struct field.** Task 1 says the echo
   fields are "additive and `omitempty`", while also (correctly)
   recommending a struct for `thresholds` because `max_rec` is a
   mapping of name to string. `encoding/json` ignores `omitempty` on
   structs. Say what is actually meant: old records decode with zero
   values, and once every record is re-frozen the tag buys nothing, so
   either use a pointer or drop the tag and rely on the validation.

3. **Family 10's search response costs more than a search response.**
   beets fetches every search hit with a full release lookup:
   `should-ask` logs `002.json` (the search) followed by `006`-`009`
   (four release lookups). The plan's "the search response deliberately
   returns a release" therefore implies at least one further
   hand-authored release document. Say so, so the case is budgeted
   honestly.

4. **The empty-search-result document has an unsourced shape.** The
   plan says "a well-formed WS/2 search document with zero releases"
   without saying where the shape comes from. It comes from
   `corpus/cases/should-ask/responses/002.json`:
   `{"created": ..., "count": ..., "offset": 0, "releases": [...]}`.
   Cite it the way the plan already cites the smoke fixture for the
   release document, and note the search path is `/ws/2/release?...`
   with no trailing slash (the canonicalization rule preserves that
   distinction deliberately).

5. **A Global Constraint contradicts Task 2.** "No case directory is
   written except through the schema writer" is violated by
   `harness synth` writing `responses/001.json`, exactly as
   `harness capture` already does with `os.WriteFile`. Reword to what
   is meant: `case.json` is written only through `cassette.Case.Write`;
   response bodies are written verbatim and never re-encoded.

6. **`harness freeze` rewrites the whole `case.json`, not just the
   record.** `Case.Write` re-marshals `Input`, and `ItemFields` values
   round-trip through `float64`, so an integral float loses its
   spelling (`"length": 210.0` in
   `internal/harness/oracle/testdata/cases/smoke/case.json` becomes
   `210`). The seven live corpus cases happen to be safe (their
   integral lengths are already stored without a fractional part), but
   a synth-emitted case will mutate on its first freeze. Add a
   freeze-idempotence test: everything but `expected` is byte-identical
   before and after.

7. **Task 5 deliverable 3 lists four `track_index` shapes against a
   three-case budget** ("matching absolutely, matching per-disc,
   mismatched, and absent" at 3 cases). The inventory does the same.
   Say which two shapes share a case, or move the count to 4 and
   restate the family arithmetic.

8. **The single-source rule reads pass-wide but is not.** "How a
   synthetic case gets its MusicBrainz responses", item 4, says a case
   "never adds a second metadata-source plugin", while Task 3's table
   coverage deliberately monkeypatches `find_metadata_source_plugins`
   for the `TestDataSourceDistance` rows. Both are right; scope item 4
   to cassette cases explicitly so the component tool's rows do not
   read as a violation.

9. **The ledger-joins-as-it-lands rule has an unstated cost and an
   unstated red window.** Every new case needs a `derive_ties.py` run
   (a venv invocation per case) before `ties.json` can carry its row,
   and the gate fails between adding a case and updating both ledgers.
   State it in the authoring loop so the per-case budget is honest and
   the implementer knows the tree is legitimately red mid-case.

10. **The component table's gate-resident check has no floor.** "Schema
    validity, row-id uniqueness, and non-emptiness" passes over a
    one-row table. Add a per-method coverage set (every `add_*` method,
    `string_dist`, `track_distance`) and a row-count floor, so the
    table cannot silently shrink between rungs. The plan is right that
    the check cannot be a conformance check at 2b; that is a reason to
    make its coverage assertion sharp, not to leave it soft.

11. **The inventory's own numbering is off.** Section (a) points at
    "family 16 in (b)"; the table stops at 15, and family 15 is the
    recommendation boundary. The plan reads it correctly; fix the
    inventory so a later reader does not go hunting.

12. **The perturbation override list needs `mb_albumid` fenced off.**
    beets picks the looked-up release id from `current_metadata`'s
    most-common value across items, so a perturbation that changes or
    omits `mb_albumid` on enough items changes which release is fetched
    (and, per blocker 3, can make a request unanswerable). Task 2's
    override list should name `mb_albumid` as off-limits outside family
    10, and Task 4's "extra item" case should carry the album's id like
    every other item.

---

## Verified sound

These mechanisms were checked against the shipped source and found
correct. The fold should not re-litigate them.

- **The ID-lookup premise.** `internal/harness/oracle/testdata/cases/smoke/`
  is a genuine existence proof: fabricated `mb_albumid`
  `3b6b0a4e-...-0001` on both items, one logged release lookup, one
  hand-authored `responses/release.json`, and `TestDeriveRunsPinnedBeets`
  asserts a strong recommendation with exactly one candidate carrying
  that album id. A synthetic ID-lookup case does drive a real
  derivation end to end, exactly as the plan claims.

- **The threshold/weight gap is real and correctly diagnosed.**
  `derive_oracle.py` emits `record["weights"]` and
  `record["thresholds"]` (l.253-259); `cassette.Expected` has no field
  for either; `oracle.Derive`'s `json.Unmarshal` drops both. Every
  committed `case.json` confirms it: the `expected` objects carry
  `action`, `applied_fields`, `assignment`, `candidates`,
  `config_hash`, `echoed_date`, `extra_items`, `extra_tracks`,
  `recommendation`, and nothing else. Adding the fields is the right
  fix and the inventory's recommendation (ii) is sound.

- **The `max_rec` decode hazard is real.** `derive_oracle.py` writes
  `"max_rec": config["match"]["max_rec"].get(dict)`, a mapping of
  penalty-key name to recommendation *name*, beside three floats. A
  flat `map[string]float64` would indeed fail. The plan's insistence on
  a decode test over the shim's actual output, not a trimmed sample, is
  the right guard.

- **The re-freeze is genuinely forced, and the two knobs named are the
  right ones.** `TestCorpusReDerivesToItsCommittedRecords` compares
  with `reflect.DeepEqual` over the decoded `*cassette.Expected`, so
  populated echo fields on the fresh derivation against zero values on
  the committed record is a hard failure until every record is
  re-derived. `corpus/manifest.json`'s `schema_version` and
  `corpus_test.go`'s `wantManifestSchemaVersion` are exactly the pair
  that must move together.

- **Case-count arithmetic adds up.** Task 4: 8 + 7 + 7 + (5 + 2 + 5) =
  34. Task 5: 4 + (2 + 3 + 2) + (4 + 3 + 2) + (2 + 5) = 27. Total 61,
  matching the inventory's per-family budget row for row and sitting at
  the floor of the spec's 60-120 band. Task 6's "68 committed cases" (7
  smoke plus 61) is right.

- **The search-fallback story is correct and evidenced.** A non-strong
  ID-lookup recommendation does make pinned beets fall through to its
  text search (`should-ask` and `should-reject` prove it live), an
  empty search document does keep the candidate set to the single
  looked-up release, and `oracle.Derive` does write the exact canonical
  `path_query` of the missing request into `<case dir>/misses.json` for
  the author to answer. The loop's shape is right; only its
  termination edge (blocker 3) and its degradation edge (blocker 2)
  need work.

- **Record determinism needs nothing added.** `PYTHONHASHSEED=0` is
  set by `runShim` and enforced by the shim's own guard;
  `stable_json` pins key order and float spelling at `%.17g`; one shim
  process per case pins the config cache; extras sort after a set
  difference under the fixed seed. The plan correctly adds no new
  determinism machinery for the synthetic cases.

- **The tie detector belongs outside `cassette.Expected`.** The
  inventory's claim checks out: `Candidate.Tracks` is
  `dist.tracks[track]` for matched pairs only, computed with
  `incl_artist=album_info.va`, which is the final-scoring objective,
  not the assignment objective. Off-diagonal costs appear nowhere in
  the record, and `Input.Items` plus the committed responses genuinely
  do carry everything needed to rebuild the matrix. Keeping the
  assignment cost matrix out of the schema is right.

- **`incl_artist=False` is the correct assignment objective**, and its
  divergence from the final `incl_artist=va` scoring on VA albums is
  the documented reason ties matter at all (mechanics study, threats
  section, item 2). The plan cites it correctly.

- **The component tool's shape is right.** No stub, no `tag_album`, no
  response body, no rendered profile, and therefore not a
  `cassette.Case`. Living in `internal/harness/oracle` keeps the U7
  boundary ("a beets bump edits only this package") true as written.
  The plan's refusal to dress the gate-resident half up as a
  conformance check it cannot be at 2b is exactly the right call, and
  the honesty is worth preserving through the fold.

- **The epsilon citation is real.** The mechanics study's threats
  section, item 3, does document the ~1e-6 flag regime and the
  strict-`<` versus `<=` asymmetry at `match.py:285-288`. No new
  research is needed, as the plan says.

- **The no-live-network stance holds.** `harness capture` is invoked by
  no task, `DUBPLATE_HARNESS_CONTACT` is not a pass input, and
  `oracle.Derive` reaches only the loopback stub. `Derive` is indeed
  untagged in the shipped source, so a `cmd/harness` verb can link it
  while the gate stays Python-free.

- **The Tier C exclusions are correctly drawn and correctly deferred.**
  Multi-source ordering, `preferred.original_year: yes`,
  `match.required`/`match.ignored`, file-level applied tags, and
  further synthetic pseudo-release variants are all genuinely out of
  reach under the two shipped profiles and the single-source ruling,
  and Task 6's carry list names each one. Neither profile carries a
  `match:` block, confirming the inventory's claim that both run at
  `config_default.yaml`'s thresholds.

- **The spec sanctions the two analyses as designed.** Ladder entry 2
  names "the assignment-tie detector (perturb and re-solve; tie cases
  are excluded or pinned)" and "the threshold-proximity flag" in the
  plan's own terms. The objections above are to under-specification
  (major 4) and to an unguarded index (major 2), never to the
  approach.
