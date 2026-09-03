# Rung 2, Pass 2d (File-Level Applied Tags and the Standing Chore Docs) Implementation Plan

> **For agentic workers:** workflow mode per CLAUDE.md's "Build
> process" section: pass-execute chain per task. The conductor consumes
> reports and never reads diffs inline. Subagents never dispatch other
> agents; `code-simplifier` is the conductor's dispatch after a task's
> review accepts. The conductor also owns the ledgers
> (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`) and the batch
> commit of anything a task drafts into them. Rung 2 is not an
> adversarial-cycle rung (those are 5, 12, and 16), so review scales to
> the in-chain `diff-reviewer` plus the pass-end gate. **This pass runs
> its four tasks strictly sequentially, Task 3 first; see the Execution
> model.**

Revision 2, 2026-09-03: folds the adversarial plan review
(`docs/superpowers/research/2026-09-02-greenfield-spec-reviews/rung-2d-plan-review.md`),
the BLOCKER and all five MAJORs as specified, and minors m1 through m7
where cheap (all of them). One adaptation on MAJOR 4: the review
offers two fixes — drop Task 3's independence and run it first,
sequentially, or keep it concurrent with a reduced, explicitly-named
bar — and this fold takes the first, since it is the review's own
stated preference and costs nothing against either budget. Task 3 now
runs first in execution order with no parallelism claim; see the
Execution model. The review's "Verified sound" section is binding and
not re-litigated: the pure-`mutagen` template route and its 42-byte
bar; write-then-read-back determinism across two independent runs;
JSON-clean read-back values needing no `jsonable()`-style coercion;
`Item.write()` returning `None`; the dossier's exactly-seven File-tag-
subject flags; the schema bump forcing no re-freeze; the FLAC blob
clearing the secret scan; M5 honored and strengthened; the single-
invocation shim shape; the derive wall-time budget's arithmetic; and
the gate-resident half's offline checkability against the five extant
ledger-check precedents.

Pass 2d is the tail M10 cut off pass 2c: the
file-level applied-tag derivation rung 8 is blocked on, the two
standing chore docs, and the whole-corpus README rewrite. The scope
is unchanged from 2c's Task 13 and Task 14 as originally written, and
the decisions those tasks would have made inline were pre-ruled at
2c's fold and recorded in its Task 20 carry-forward note
(`docs/superpowers/plans/2026-09-03-rung-2c-catalog-adversarial.md:1023-1072`).

**The pre-rulings are binding, not open questions.** M3, M4, M5, m2,
and M8 were ruled by the conductor at the 2c fold precisely so 2d
would not re-derive them. An implementer that finds one of them
questionable reports the finding and implements the ruling anyway;
re-litigating a pre-ruling mid-task is the failure mode the fold
exists to prevent. In short:

- **M4:** the audio fixture is one small static FLAC blob vendored
  under `internal/harness/oracle/testdata/`. `make derive` gains no
  binary dependency. `ffmpeg` and `flac` are not adopted as
  prerequisites.
- **M5:** `derive_file_tags.py` is a second, independent derivation
  entry point in the shape of `derive_component.py`. It is never
  called from `oracle.Derive`, and no item in the corpus derivation
  path ever gains a non-empty `path`.
- **M3:** `file_tags.json` needs both halves. A gate-resident
  structural check, and an oracle-tagged
  `TestFileTagsLedgerMatchesTheCorpus` that re-derives and diffs.
  Structure alone cannot go red against a `mediafile` version bump
  that moves tag shapes, which was M3's whole finding.
- **m2:** `Reconcile` lives in `internal/harness/oracle/reconcile.go`,
  and its `versionProbe` names five components today (beets,
  unidecode, jellyfish, lap, numpy). Adding `mediafile` edits both the
  Go tuple and the Python probe string in that file.
- **M8:** per-case derive cost is `1.04s + 0.25s × (requests − 1)`,
  never a flat per-case figure.

Consumes pass 2c's shipped corpus exactly as committed: the
`cassette` schema at v2, `internal/harness/stub`, `oracle.Derive`,
`derive_oracle.py`, `derive_ties.py`, `derive_component.py`,
`internal/harness/analysis`, `internal/harness/mbrequest`, `harness
capture`, `harness synth`, `harness seed-items`, `harness freeze`,
`make derive`, the `estate`, `stock`, and `locale` profiles, and every
committed case with its `claims.json` and ledger rows. 2d adds sibling
artifacts on top of that corpus. **It authors no case, captures
nothing, and edits no committed record.**

**Primary input:** the planning inventory
(`docs/superpowers/research/2026-09-03-rung-2d-planning-inventory.md`,
2026-09-03), which checks every pre-ruling against the pinned venv's
own beets source, names the exact intersection the written tag set is
drawn from (`Item._media_fields`,
`beets/library/models.py:731`), establishes that `Item.write()`
returns `None` so an independent read-back is the only signal
(`models.py:934-981`), corrects the record on the `_salvage/` FLAC
fixture (there is none; 2d authors the blob fresh), and prices the
pass's derive-time impact. Read it in full before Task 1. Two of its
findings are the pass's organizing decisions and must not be
re-litigated: that the written field set is a strict subset of
`AppliedFields` rather than all of it, and that twelve subjects in one
shim invocation is a single task's worth of derive cost, so this pass
needs no batch-task structure.

**Goal:** close rung 2 by shipping the evidence rung 8 is blocked on.
Ship the file-level applied-tag ledger derived by executing pinned
beets' own write path against a real FLAC container and reading the
result back independently, ship the two standing chore docs that keep
the corpus honest against a library bump, and rewrite the corpus
README over the whole committed corpus.

**Owner touchpoints: none.** 2d needs no ruling, no capture contact,
no live network, and no adjudication. The adjudication batch 2c
prepared is already with Geoff and is not this pass's to chase; its
rows stay `pending-adjudication` and the gate accepts them. Nothing in
this pass produces a new item for Geoff to adjudicate. If a task
believes it has found one, that is a finding to report and a row to
write, never a mid-pass ask. The standing 80%-ceiling flag CLAUDE.md
requires of every pass still applies and is contingent, not a
touchpoint this plan expects to spend. Every stop condition named
below — Task 1's escalation trigger, Task 2's two unresolvable-id
stops — resolves to the conductor, never to Geoff directly; only a
conductor decision to escalate reaches Geoff, and that is what would
spend the touchpoint above.

**Spec:** `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified). Ladder entry 2 names the version-bump
differential job and the client's hard 0.25s inter-request floor in
differential wall-time budgets as deliverables, and both land here.
Ladder entry 8 is the consumer: pure-Go FLAC tag writing "proven
against the corpus's expected applied tag sets from rung 2" is exactly
what this pass makes provable, and Task 4's handoff statement says in
writing what rung 8 may assume.

**Token ceiling: 2.5M**, as a starting figure, re-derived twice: after
Task 1 and again after Task 2. Task 1 is machinery only (one blob, one
shim, one driver, three small edits) and predicts Task 2 reasonably
but says little about Task 4, the pass's largest task by a wide
margin — it is the first to touch the corpus at scale, reading every
committed case to build the whole-corpus inventory. Re-deriving only
after Task 1 repeats 2c's own M9 defect with a different largest task;
re-deriving again after Task 2 gives the conductor its first honest
signal before Task 4 starts. The basis: 2b spent ≈5.8M across 14
tasks, ten of which were case-authoring batches, for a ≈415K per-task
average; 2d's four tasks are all of the machinery, prose, and close
character that ran above that average in 2b, and 2.5M is ≈625K per
task. The pass has no case authoring, no capture, and no network, so
the case-count scaling that dominated 2b and 2c is absent entirely.
Flag at 80% (2.0M): finish the task in flight, write STATUS, and one
combined question goes to Geoff before continuing. **Checkpoint:**
STATUS after every task, which is 2b's and 2c's cadence and is cheap
at four tasks.

**Derive wall time, budgeted.** `make derive` is O(all committed
cases) and 2c projected roughly 9 to 10 minutes per full run at its
own close. Tasks 1, 2, and 4 each end with one full run, and Task 4's
determinism evidence needs two from a clean tree, for roughly 40 to 50
minutes of cumulative derive wall time across the pass. The
file-level derivation itself adds a negligible line: twelve subjects,
one Python process, no network, so no inter-request floor applies to
it at all. Task 3 runs no derivation; see its entry.

---

## Global Constraints

- go-conventions before any Go; python-conventions before any Python.
  Both legs are already gated (`golangci-lint`, `vale` over Go comment
  prose, `ruff` over tracked `*.py`).
- **No live network anywhere in this pass.** The gate never reaches
  the network, and neither does anything else here. Every subject case
  is already committed, so `harness capture` is not invoked by any
  task, `DUBPLATE_HARNESS_CONTACT` is not a pass input, and no `/ws/2`
  request is issued by any tool. The differential and re-capture docs
  *describe* network-driving jobs; writing a doc about a job is not
  running it.
- **The gate never needs the venv, and derivation stays behind the
  `oracle` build tag and `make derive`.** `bash scripts/check.sh` and
  `go test ./...` stay offline and Python-free. Every new Python file
  and every new Go file that calls one carries `//go:build oracle`,
  matching `internal/harness/oracle/component.go` and
  `component_derive_test.go`. The gate-resident half of Task 2 is pure
  Go in `internal/harness/analysis`, which carries no build tag, per
  the precedent `analysis/component.go` already sets.
- **2c's corpus data is frozen input.** 2d adds sibling artifacts and
  never edits records. No committed `corpus/cases/*/case.json`,
  `claims.json`, or `responses/*.json` is touched, re-frozen,
  reformatted, or re-derived into a different value. No row in
  `ties.json`, `proximity.json`, `components.json`, or
  `adjudications.json` is edited. **The invariant M5 states is the
  pass's acceptance test: `make derive` re-derives every committed
  record byte-identically after 2d's change.** A record that diverges
  is a defect in 2d's work, never an expected consequence of it, and
  it stops the task.
- **The cassette surface is frozen, with exactly one additive
  exception** carried over from 2c's M10 deferral: `Manifest.MediaFile`
  plus the `schema_version` bump to `"3"`. No other field is added,
  renamed, or repurposed. **The bump forces no re-freeze**, because
  `Manifest` is corpus metadata rather than a `cassette.Expected`
  field, so no committed record's content changes. This is the one
  place 2d differs from 2b's schema-version event, and the difference
  is worth stating so nobody re-freezes 200-odd cases to satisfy a
  half-remembered rule. `internal/harness/corpus_test.go`'s
  `wantManifestSchemaVersion` moves with the manifest, and the reason
  is recorded in the README's manifest section.
- **`file_tags.json` is a derived ledger and is generated, never hand
  written.** It lands at `corpus/analysis/file_tags.json` alongside
  its three siblings and is written only through the oracle-tagged
  test's `-update-ledgers` path, exactly as `components.json` and
  `ties.json` are. An agent that edits it by hand has broken the same
  rule as an agent that hand-writes an `Expected` record.
- **Every expectation is still derived by executing pinned beets.**
  The file-level values in the ledger are what `Item.write()` actually
  landed in a real FLAC container and what an independent
  `mediafile.MediaFile(path)` read back, never what anyone expected
  either to be. `Item.write()` returns `None`, so the read-back is the
  only signal there is; this is not a stylistic preference about
  trusting return values, there is no return value.
- **Universality.** The vendored FLAC blob carries no real artist,
  album, or identity metadata, and no estate hostname, path, or email
  appears anywhere outside `internal/config`. The blob is silent audio
  with an empty tag block; every tag in it at read-back time was put
  there by the derivation under test.
- **Assertion discipline.** No silent-success fakes, no vacuous
  assertions. Every new gate-resident check ships a fixture that makes
  it fire, and the task reports the mutation that proved it can fail.
  A check that passes over an empty ledger is a check with no teeth,
  and the coverage floor is what stops that.
- **A sub-second `make derive` is a cache hit and proves nothing.**
  `make derive` is a fixed recipe with no `-count` hook of its own, so
  force it at the environment: `GOFLAGS=-count=1 make derive` for any
  derivation reported as evidence. Report wall time broken down
  (corpus derive, ties derive, component derive, and now file-tags
  derive), never a bare "green".
- `_salvage/` stays read-only and holds no FLAC fixture to adopt; the
  inventory checked and the tree is empty of audio. The box is
  untouched. Gate green and clean `git status` at every task end;
  never commit red; specific files only; commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`.

## How a file-level record gets derived

Stated once here because Tasks 1 and 2 both depend on it, and because
the shape is easy to get subtly wrong in a way that still produces a
plausible-looking ledger.

1. **The input is a committed record, not a case directory.**
   `derive_file_tags.py` reads a subject case's frozen
   `Expected.AppliedFields`, which is the field-level record of beets'
   in-memory `apply_metadata()` step. It starts no stub, issues no
   request, renders no profile, and loads no plugin, matching
   `derive_component.py`'s minimalism.
2. **`BEETSDIR` is set before beets is imported.** `Item.write()`
   resolves `beets.config["id3v23"]` and fires the `write` plugin
   event before it opens the file (`models.py:961-970`), which means
   importing beets without isolation reads the runner's own
   `~/.config/beets` and can silently change a value or raise on a
   non-bool `id3v23` setting the runner happens to have configured.
   The failure is machine-dependent, so it can pass clean on this
   workstation and fail on a stranger's, which is exactly the setup
   goal CLAUDE.md states. Set `BEETSDIR` to an empty scratch directory
   first, matching `derive_component.py:448-452`'s "an empty
   `BEETSDIR` is what puts the run on beets' shipped defaults," and
   assert in the derivation that `os.environ["BEETSDIR"]` points into
   the scratch directory at the moment `Item` is imported, so the
   isolation is proven rather than assumed.
3. **Build items from those fields.** The `build_items` shape already
   in `derive_oracle.py:156-165` is the template. Every subject's
   items are reconstructed from the record, so the ledger is a
   derivation *of* the corpus rather than a second, parallel corpus
   that could drift from it. Assert faithfulness cheaply while
   building: every `_media_fields` value read off a subject's
   `AppliedFields` must already be a JSON primitive or a list of them,
   never a value that needed `AppliedFields`' own `str(value)` last
   resort (`derive_oracle.py:105-121`) to serialize. Measured against
   the pinned venv every intersection field is backed by a primitive
   beets type today, so the fallback cannot fire — but a future field
   type could, silently turning a value into its stringified repr, and
   this assertion is what keeps that from shipping unnoticed.
4. **Each subject gets its own fresh file.** Copy the committed
   template FLAC bytes to `<scratch>/<case-id>-<item-index>.flac`
   inside one `tempfile.TemporaryDirectory` per process, matching
   `derive_oracle.py:229` and `derive_component.py:451`. `item.write()`
   mutates the file at its path in place, so a shared path would let
   one subject's write leak into the next subject's read-back. The
   context manager cleans up whatever the outcome, so a crashed run
   leaves no residue for the next.
5. **Write, then read back independently.** Call `item.write()` into
   the copy, then open a fresh `mediafile.MediaFile(path)` against the
   bytes on disk. The read-back must be a new parse, not the
   `MediaFile` object `write()` already had open; this is exactly what
   `Item.read()` does internally at `models.py:917`, and the ledger
   wants the raw per-field values `mediafile` reports rather than a
   second `Item`.
6. **The value keys split into a writable set and a read-only
   remainder, and only the writable set is a per-row value.**
   `Item.write()` filters to `Item._media_fields` before touching the
   file (`models.py:967`), which is
   `set(MediaFile.readable_fields()) & _field_names` (`models.py:731`)
   — but `MediaFile.update()`, which `write()` actually calls, iterates
   `sorted_fields()`, the *writable* set only. Measured against the
   pinned venv, `_media_fields` is 78 keys; nine of them
   (`bitdepth`, `bitrate`, `bitrate_mode`, `channels`, `encoder_info`,
   `encoder_settings`, `format`, `length`, `samplerate`) are
   `MediaFile`'s read-only audio properties, never written by
   `Item.write()`, and their read-back values come from the template
   blob's `STREAMINFO`, not from the case — measured concretely,
   `length` reads back `0.0` regardless of what an item's own
   `AppliedFields.length` says, which is a direct contradiction if a
   read-only key is ever treated as a per-subject value. Compute both
   sets from the pinned venv rather than hardcoding either: the
   69-key writable set is each row's value dict; the nine-key
   read-only remainder is recorded once, in a `template` block shared
   by the whole ledger, since its values are a property of the blob
   and never of a subject. Bookkeeping columns such as `id`,
   `album_id`, `mtime`, and `path` are `Item` columns outside
   `_media_fields` entirely and never reach a file at all; a row
   carrying one is a separate defect from a row carrying a read-only
   property, and both are checked.
7. **The emitted document carries the key lists as a header, not just
   the rows.** `derive_file_tags.py` writes the writable set and the
   read-only remainder into the document itself, once, so the ledger
   is self-describing and no key list is hand-copied into a second,
   independently maintained home. The pure-Go gate-resident checker
   (Task 2) reads both lists from this header rather than carrying its
   own copy — one derived fact, one home.
8. **One process for all twelve subjects, one document out.**
   `derive_component.py`'s "no arguments, one document out" contract
   is the model, not `derive_oracle.py`'s per-case invocation. The
   ledger is a single derived document keyed by case and item index,
   with the header block from step 7 alongside the rows.

---

### Task 1: The template FLAC fixture, the file-level derivation, and the sixth version pin

Model: `sonnet`. **Named escalation trigger in place of an upshift:**
two failed attempts to produce a template blob that `mediafile` opens
as a FLAC, or a `mediafile` API that will not do what step 4 above
describes, stops the task and reports rather than spending a third
attempt.

**Files:**
- Create: `internal/harness/oracle/testdata/template.flac` (the
  vendored blob) and the one-time authoring script that produced it
- Create: `internal/harness/oracle/derive_file_tags.py` and
  `internal/harness/oracle/file_tags.go` (`//go:build oracle`, the
  driver, in the shape of `component.go`'s `DeriveComponentTable`)
- Modify: `internal/harness/cassette/cassette.go` +
  `cassette_test.go` (the `MediaFile` field)
- Modify: `internal/harness/oracle/reconcile.go` (the Go `pinned`
  tuple and the Python `versionProbe` string, both)
- Modify: `corpus/manifest.json`, `internal/harness/corpus_test.go`'s
  `wantManifestSchemaVersion`, `corpus/README.md`'s manifest section

**Behavior contract:**

- **The template blob is authored by the pure-`mutagen` route.**
  `mutagen` is already a pinned transitive dependency of `mediafile`
  in the venv and needs no new pin, and the route removes the
  `ffmpeg`/`flac` presence question from even the one-time authoring
  step. The blob is a valid, minimal, silent FLAC that `mediafile`
  opens: mutagen's own reader dispatches on the `fLaC` signature and a
  well-formed `STREAMINFO` block
  (`mediafile/__init__.py:174`), which is the bar the blob must clear.
  Prove it clears the bar in a test, never by inspection.
- **The authoring script is committed as the blob's provenance.** It
  is ruff-gated like every other tracked `*.py`, is never invoked by
  `make derive` or the gate, and says in its own docstring that it
  runs once and its output is the artifact. `corpus/README.md` records
  the blob's provenance, its byte size, and the command that produced
  it, per M4. A vendored binary with no reproduction path is the thing
  M4's ruling was trying to avoid.
- **`derive_file_tags.py` implements the eight steps above exactly,
  `BEETSDIR` isolation included.** One process, all twelve subjects,
  one JSON document to stdout keyed by case id and item index, with
  the writable/read-only header block from step 7. It takes the
  subject list as input rather than discovering it, so Task 2 owns
  which cases are subjects and this task owns the mechanism. The
  `BEETSDIR`-points-into-scratch assertion (step 2) ships as part of
  this same deliverable: a shim that writes correct tags on this
  workstation only because this workstation's `~/.config/beets` is
  empty is not the isolated derivation the plan requires.
- **The writable set and the read-only remainder are both computed
  and recorded, not assumed, and kept distinct.** Read
  `sorted(Item._media_fields)` off the pinned venv (beets computes it
  at `models.py:731` as `set(MediaFile.readable_fields()) &
  _field_names`; take the attribute rather than recomputing the
  expression) — that is 78 keys, not the ledger's value-key contract.
  The contract is the 69-key subset `MediaFile.update()` actually
  writes (`sorted_fields()`); the remaining nine
  (`bitdepth`, `bitrate`, `bitrate_mode`, `channels`, `encoder_info`,
  `encoder_settings`, `format`, `length`, `samplerate`) are read-only
  audio properties that a read-back reports from the template blob's
  `STREAMINFO` regardless of the subject, never from the case, and
  belong only in the ledger's single shared `template` block. Record
  both lists, distinguished, in `corpus/README.md`'s file-tags section,
  and report both in the task return. Prove both exclusions
  concretely: assert in a test that no bookkeeping column (`id`,
  `album_id`, `mtime`, `path`) appears in a derived row, and a second
  test that no read-only property appears in a row's value dict
  (only in the shared `template` block) — so a future `mediafile`
  release that starts reporting one, or an implementer who reaches for
  the wrong set, fails loudly instead of quietly widening the ledger
  or handing rung 8 a value no writer can reproduce.
- **`Manifest.MediaFile` is the pass's single additive schema change.**
  Add the field, pin `mediafile==0.17.0` (already resolved in the venv
  as a `beets==2.13.1` transitive dependency, `tools/venv/requirements.txt:19`
  — no install needed), bump `schema_version` to `"3"`, and move
  `wantManifestSchemaVersion` with it. Record the reason in the README
  per the corpus-revalidation rule. **No case record is re-frozen**,
  and the task's own evidence that none was is a `make derive` run
  reporting every committed case re-derived byte-identically.
- **`Reconcile` gains its sixth component, on both sides, and the test
  covers the direction that is actually silent.** `reconcile.go`
  carries a Go `pinned` tuple and a Python `versionProbe` string, and
  both name the same five components today. Edit both. The loop at
  `reconcile.go:50-58` iterates `pinned`: a `pinned` tuple naming
  `mediafile` against a `versionProbe` that omits it is loud (the
  comparison finds `installed["mediafile"]` empty and reports a
  mismatch), and a genuinely absent package is loud too
  (`md.version(n)` raises `PackageNotFoundError`). The silent
  direction is the other one — a `versionProbe` naming `mediafile`
  while `pinned` stays at five keys means the probe reports six and
  the loop only ever compares five, so the sixth is never examined.
  Prove the pairing with a test that drives `Reconcile` against a
  `pinned` set and a probe output disagreeing in *membership* in that
  direction (probe has a key `pinned` lacks), report which direction
  the mutation exercised, and add the loud direction as a second,
  cheap assertion rather than the one doing the proving.

**Deliverable count: 4** (the template blob with its committed
authoring script, provenance record, and open-as-FLAC proof;
`derive_file_tags.py` and its oracle-tagged Go driver implementing the
eight-step contract including `BEETSDIR` isolation; the computed
writable set and read-only remainder recorded distinctly with both
exclusion proofs; `Manifest.MediaFile` with the schema bump, the
`Reconcile`/`versionProbe` sixth pin on both sides proved in the
silent direction, and the no-re-freeze evidence). Gate plus
`make derive` green; commit. Report actual token spend; it re-derives
the ceiling for Task 2.

---

### Task 2: The file-tag ledger, its gate-resident checker, and its re-derive test

Model: `sonnet`. Depends on Task 1.

**Files:**
- Create: `corpus/analysis/file_tags.json` (generated via
  `-update-ledgers`, never hand-written)
- Create: `internal/harness/analysis/file_tags.go` +
  `file_tags_test.go` + `testdata` fixtures (pure Go, no build tag,
  in the shape of `analysis/component.go`)
- Create: `internal/harness/oracle/file_tags_derive_test.go`
  (`//go:build oracle`)
- Modify: `internal/harness/corpus_test.go` (wire the gate-resident
  checker in, alongside `TestComponentTableIsWellFormed`)
- Modify: `corpus/README.md` (the "Regenerating a derived ledger"
  section)

**Behavior contract:**

- **The subject set is twelve, and it is verified against the tree,
  not taken on faith.** Five come from the committed corpus and were
  named in 2c's plan review: `clean-match`, `media-uneven-split`,
  `va-exact`, `non-latin-pseudo-release`, `year-reissue-ratio-penalty`.
  The other seven, pre-extracted from 2c's Task 3 selection dossier
  (`docs/superpowers/research/2026-09-03-real-catalog-selection.md`,
  File-tag subject column, lines 84-88 for the row format), are:
  `real-classical-karajan-beethoven-89`, `real-reissue-nevermind-2011`,
  `real-boxset-james-brown-star-time`,
  `real-nonlatin-utada-first-love`, `real-noise-dylan-bootleg`,
  `real-genre-public-enemy-nation-millions`,
  `real-decoy-beethoven-symphony-5`. Task 2's job on this half is
  reduced to confirming each of these twelve resolves to a committed
  case, not to re-deriving the set from the dossier. **A named id that
  does not resolve to a committed case is a finding to report and a
  stop, never a gap to pad or trim.** The ids came from a review
  document written before 2c executed, and the corpus is 2c's product;
  checking is cheap and the alternative is a ledger that silently
  covers fewer than twelve cases.
- **The ledger is generated.** Promote nothing by hand: run
  `go test -tags oracle ./internal/harness/oracle -run FileTags
  -update-ledgers`, commit what it wrote, and re-run without the flag
  to prove the committed file is what a fresh derivation produces.
- **The gate-resident half is structural and honest about it.** In
  `internal/harness/analysis`: every subject named in the ledger is a
  committed case; every row's case-and-item pointer resolves into that
  case's `AppliedFields`; every value key in a row sits inside the
  ledger's own writable-set header block (Task 1 step 7) — never a
  second, hand-maintained Go copy of that list; **no row carries any
  key from the read-only remainder** (the negative assertion the
  BLOCKER fix requires, and the one with real teeth, since those nine
  keys really do sit inside `_media_fields` today); the subject set
  matches the twelve committed ids above; and a row-count floor stops
  the ledger silently shrinking between rungs, per
  `analysis.ComponentRowFloor`'s precedent. **Do not dress this half
  up as a conformance check.** The Go matcher it will eventually be
  compared against does not exist until rung 3, and M3's finding was
  precisely that structure alone cannot go red against a `mediafile`
  bump.
- **The oracle-tagged half is the one with teeth**, and its template
  is `TestComponentTableReDerivesToItsCommittedTable`
  (`internal/harness/oracle/component_derive_test.go:22-64`), not a
  fresh design. `TestFileTagsLedgerMatchesTheCorpus` derives fresh,
  decodes both the fresh and the committed JSON to `any`, and
  `reflect.DeepEqual`s them, because the committed file is indented
  for review while the shim prints one line. It fails with the
  "re-run with `-update-ledgers`" message the two extant ledger tests
  already use, and it honors the package's existing `*updateLedgers`
  flag rather than adding a second one.
- **Both halves ship a fixture that makes them fire**, and the task
  reports the mutation that proved each can fail: a ledger row
  pointing at a case id that is not committed, a row carrying a value
  key outside the header's writable set, and — the fixture with the
  real teeth — a row carrying one of the nine read-only keys, are the
  gate-resident half's three. For the oracle-tagged half, a one-value
  perturbation of the committed ledger that the re-derivation
  contradicts.
- **The README's regeneration section gains `file_tags.json`.** Open
  `corpus/README.md`'s "Regenerating a derived ledger" section and
  extend it in the register it already sets; do not write a new
  section from this plan's description of it. The section must say
  which command regenerates the file, that hand-editing it is not a
  path, and that a regeneration is expected to be a no-op unless the
  venv's pins moved.
- **The M5 invariant is proven here, not assumed.** Report the
  `make derive` run showing every committed record re-derived
  byte-identically with the file-level derivation in the tree, and
  report the count of cases actually re-derived rather than "green".

**Deliverable count: 4** (the generated `corpus/analysis/file_tags.json`
over a verified twelve-subject set, its per-row value dicts holding
only the writable set with the read-only remainder confined to the
shared `template` block; the pure-Go gate-resident checker wired into
`corpus_test.go` with its three firing fixtures, reading the key lists
from the ledger's own header; `TestFileTagsLedgerMatchesTheCorpus` on
the component-table template with `-update-ledgers` support; the
README regeneration section and the byte-identical re-derivation
evidence). Gate plus `make derive` green; commit. Report actual token
spend; it re-derives the ceiling for Tasks 3 and 4, per the token
ceiling section above — Task 2 is the first task to touch the corpus
at scale and the first honest signal about Task 4's cost.

---

### Task 3: The version-bump differential job and the re-capture chore docs

Model: `sonnet`. **No dependencies, and it runs first, sequentially**;
see the Execution model. Prose only, no code, no derivation, and its
M8 arithmetic depends only on the corpus 2c already froze, so nothing
in the machinery track blocks it.

**Files:**
- Create: `docs/corpus-version-bump-differential.md`
- Create: `docs/corpus-recapture-chore.md`

**Behavior contract:**

- **Both are standing operational docs**, so they sit beside
  `docs/bringup-runbook.md` rather than under `docs/superpowers/`,
  which holds specs, plans, and pass research. They are written for
  whoever runs this job in a year, who may be neither Geoff nor an
  agent that has read this plan.
- **The differential doc's subject** is the standing job that drives
  pinned beets as a subprocess against an estate config and a local
  stub, and the risk it exists to catch is a beets major version
  breaking the **invocation contract** (config keys, plugin
  enablement, CLI behavior) independently of any verdict change. beets
  3.0.0 is the concrete instance: it is beets' one active
  compatibility-shedding plan, with the legacy plugin layer
  TODO-marked for removal and `docs/dev/plugins/autotagger.rst`
  promising removal in v3.0.0 (U7,
  `docs/superpowers/research/2026-09-02-beets-evolution-lessons.md:396-412`).
- **The differential doc states its own falsifiable success
  condition**, in U7's words: when the differential is pointed at a
  beets 3.x, only the adapter package should need edits. Name the
  package concretely so the criterion can actually be checked:
  `internal/harness/oracle`'s Go files plus its Python shims, and
  never a harness rewrite. A doc that describes a job without stating
  what would falsify its premise is a doc nobody can act on.
- **The wall-time section uses the M8 formula at the corpus's real
  size, and it prices the whole pass, not just the corpus leg.** M8
  named two costs: the per-case floor, `1.04s + 0.25s × (requests −
  1)`, never a flat per-case figure; and the fact that a full
  `make derive` pays that floor twice, because `DeriveTies` starts its
  own stub (`ties.go:65`) and runs its own Python process
  (`ties.go:101`) entirely separately from `Derive`'s stub
  (`oracle.go:163`), replaying every case's requests a second time.
  Read the committed case count and the per-case request-count
  distribution off the tree at execution time, and present the corpus
  leg, the ties leg, and the total separately — the same shape the
  plan's own derive-report constraint already asks every reported run
  to use. The committed corpus is fixed at 2c's close and 2d adds no
  case, so this figure does not wait on Tasks 1 and 2. **The file-tags
  derive line is the one number this task cannot supply**; leave a
  named placeholder for it and Task 4 fills it from its own
  measurement.
- **The differential doc records why a beets bump's blast radius is
  scoped** (U6,
  `beets-evolution-lessons.md:387-395`): the capture tool records the
  maximal practical `inc=` response rather than the minimal matcher
  view, specifically so a later matcher improvement wanting another
  field never forces a full re-capture. That is why a bump is expected
  to move scoring and verdicts rather than response shapes.
- **The re-capture chore doc's framing is additive recall** (U8,
  `beets-evolution-lessons.md:413-`): the highest-leverage open beets
  deficiencies widen candidate retrieval rather than change scoring,
  and an improvement of that shape adds candidates while leaving the
  beets-equivalent subset's scores untouched. This is why the fixture
  format admits recall-extension cases from day one, and why a
  post-launch retrieval improvement is graded by the same harness
  rather than escaping it.
- **The re-capture doc names its trigger conditions explicitly.**
  "Quarterly, or sooner if things change" is exactly the soft guidance
  the repo's gate philosophy rejects. Name them: a beets point
  release; a `mediafile` bump, which 2d has just made a pinned
  component; a MusicBrainz schema change noticed as response-shape
  drift; or the calendar quarter, whichever comes first. Quarterly is
  the floor, not the ceiling.
- **The re-capture doc's procedure is `Reconcile`'s own contract**,
  and it names the now-six-component pin set a re-capture event
  re-pins. Quote `reconcile.go:15-20`'s reasoning about why
  transliteration tables, edit distance, and the assignment solver all
  sit directly in the distance path, and state that any disagreement
  is the corpus-revalidation event the mechanics study already names
  as the standing plan
  (`2026-09-02-beets-matcher-mechanics.md:588-591`).
- **The re-capture doc cites the differential doc's wall-time number
  rather than restating it.** One number, one home, scaled to
  re-deriving the whole committed corpus. Two independently maintained
  copies of an arithmetic result is how one of them goes stale.
- **The gate runs, but it is not the bar for this task's own content.**
  `bash scripts/check.sh` touches no `docs/*.md` file at all —
  `vale-comments.sh` lints Go comment prose, `ruff` covers `*.py`,
  shellcheck covers `scripts/*` — so a green gate here says only that
  this task did not break Go or Python, nothing about whether these
  two docs are correct or well-written. Say that plainly in the
  return. The real bar for this task's own content is the in-chain
  `diff-reviewer` against this contract's acceptance criteria, plus
  the `writing-voice` standard the two docs are written to. Run the
  gate anyway, and report it, since it is free and still worth
  knowing; just do not call it evidence of this task's quality.

**Deliverable count: 4** (the differential doc's job description and
U7 invocation-contract risk; its falsifiable acceptance criterion
naming the adapter package concretely; its wall-time section computed
with the M8 formula at the real corpus size, corpus leg and ties leg
broken out and summed, with the file-tags line left for Task 4; the
re-capture chore doc with its U6/U8 framing, explicit trigger list,
six-component pin set, and `Reconcile`-based procedure). Gate green
(reported, not treated as this task's quality bar — see above);
commit.

---

### Task 4: The pass close and the rung 8 handoff

Model: `sonnet`. Depends on Tasks 1, 2, and 3.

**Files:**
- Modify: `corpus/README.md` (adds the whole-corpus case-inventory
  section deferred from 2c, plus the file-tags section; Tasks 1 and 2
  already own the manifest and regeneration sections — this task
  extends the file with new sections, it does not rewrite what they
  wrote)
- Modify: `docs/corpus-version-bump-differential.md` (the file-tags
  derive line Task 3 left named)
- Create: the rung 8 handoff statement (a section in
  `corpus/README.md`, so rung 8's implementer finds it where the
  evidence is, rather than a fifth loose document)
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit)

**Behavior contract:**

- **The whole-corpus case-inventory section**, which 2c's Task 20
  explicitly deferred here as an extension rather than a replacement
  — this task adds sections to `corpus/README.md`, it does not rewrite
  the file whole, matching the manifest and regeneration sections
  Tasks 1 and 2 already wrote into it. One inventory covering every
  committed case, each row naming its id, stratum, family or axis,
  profile, and the behavior it exercises. Read the committed-case
  total off the tree rather than from any plan's arithmetic. **Build
  it from a mechanical extraction, not from reading every case in
  full.** Extract `id`, `profile`, `eval_date`, stratum, and family
  from each committed case into a compact intermediate (a script pass
  or a single structured read, not a prose-writing pass over each
  `case.json`), and write the inventory prose from that intermediate.
  Reading roughly 164 full case records — each carrying an `Input`, an
  `Expected` record with candidates, weights, thresholds, and
  `AppliedFields`, plus a `claims.json` — into context to produce a
  five-column table is the avoidable half of this task's cost, and the
  token-ceiling section above treats this task as the pass's largest
  precisely because of it. The new file-tags section states which
  twelve cases are subjects, what the ledger's value keys are (both
  the writable set and the read-only remainder, distinguished), where
  the template blob came from, and that the ledger is generated rather
  than authored.
- **Determinism re-derivation.** Run `GOFLAGS=-count=1 make derive`
  twice from a clean tree and report the count of cases actually
  re-derived alongside broken-down wall time: corpus derive, ties
  derive, component derive, and now file-tags derive. A sub-second run
  is a cache hit and proves nothing. This number is rung 3's budget
  line exactly as 2c's was 2d's. Write the file-tags line into the
  differential doc's placeholder as part of the same measurement.
- **The rung 8 handoff statement, in writing, stating what rung 8 may
  now assume.** At minimum: that the corpus carries expected applied
  tags at two altitudes, `Expected.AppliedFields` on every case (the
  in-memory `apply_metadata()` record) and
  `corpus/analysis/file_tags.json` on twelve subjects (what beets'
  own write path landed in a real FLAC container, read back
  independently); that the second's **writable-set rows** (69 keys)
  are the conformance target for pure-Go FLAC tag writing, that its
  own **read-only `template` block** (nine keys — `bitdepth`,
  `bitrate`, `bitrate_mode`, `channels`, `encoder_info`,
  `encoder_settings`, `format`, `length`, `samplerate`) is never a
  per-subject target since no writer produces them and a reader gets
  them from the container's `STREAMINFO`, and that the first altitude
  (`AppliedFields`) is not a conformance target at all since fields
  outside the writable set never reach a file; both key lists, by
  name; that `mediafile==0.17.0` is pinned in the manifest and
  `Reconcile` fails loudly on a bump; and where the template blob and
  its authoring script live. State the non-scope just as plainly: the
  write-side close read of the apply-metadata field mapping and the
  mediafile layer is rung 8's own plan-opening work per ladder entry
  8, and the wazero-taglib spike is rung 8's too. 2d ships the
  evidence, not the mapping study.
- **Ledger notes for the conductor.** Draft the HISTORY entry in the
  house shape: what landed, what the gate caught, and what a later
  pass would be wrong to rediscover. Carry forward explicitly: the
  rung 8 handoff statement's location; 2c's `profile_hashes` gap (m1);
  the response-status extension, still deliberately not taken; the
  Tier C exclusions and profile gaps 2b and 2c both carried and
  neither closed; and the adjudication rows still pending as the
  launch gate's open item. Update STATUS's next action to point at
  rung 3 (the Go matcher), which is what rung 2's completion unblocks.
- **STATUS stays present tense** and under 60 lines; anything
  historical moves to HISTORY rather than being summarized harder.
  Rung 2 is complete at this task's end, and STATUS should read that
  way rather than narrating how it got there.

**Deliverable count: 4** (the whole-corpus case-inventory section built
from a mechanical extraction, plus the file-tags section distinguishing
the writable set from the read-only remainder; the determinism record
with broken-down wall time, including the file-tags line written back
into the differential doc; the rung 8 handoff statement with its named
non-scope; the HISTORY and STATUS drafts with rung 2's remaining open
items carried forward). Gate plus `make derive` green; commit.

---

## Execution model

Workflow mode per the standing grant: pass-execute chain. Four tasks,
run strictly sequentially — nothing in this pass runs in parallel.
(Revision 1 marked Task 3 independent and ran it concurrently with
Tasks 1 and 2; MAJOR 4 of the plan review found that the shared
working tree makes the gate and the in-chain `diff-reviewer`
unreliable for whichever task finishes first, for a saving that pays
neither of CLAUDE.md's two budgets, and this fold drops the claim
rather than carrying a bar the plan itself could not meet.)

**Dependencies and order, stated exactly:**

- **Task 3 runs first.** It depends on nothing from the machinery
  track — its two docs and its M8 wall-time arithmetic depend only on
  the corpus 2c already froze — so there is nothing to wait for.
  Running it first also means Task 4 fills the file-tags placeholder
  from its own measurement regardless of ordering, so nothing is lost
  by going first instead of last.
- **Task 1 → Task 2.** Task 2 generates a ledger with Task 1's shim
  and checks it against Task 1's recorded intersection. There is
  nothing to generate before Task 1 lands.
- **Task 3, Task 1, and Task 2 → Task 4.** The close reports over all
  of them, in that execution order.

Every task dispatches on `sonnet`. Nothing in this pass is novel
correctness-critical logic the plan leaves unspecified: the shim's
shape is `derive_component.py`, the re-derive test's shape is
`TestComponentTableReDerivesToItsCommittedTable`, the gate-resident
checker's shape is `analysis/component.go`, and the two docs are
prose over sources this plan cites by line. Task 1 carries a named
escalation trigger in place of an upshift, matching 2c's Tasks 17 and
18. `diff-reviewer` per task. The conductor dispatches
`code-simplifier` over each accepted task's changed Go and Python
after review, before the pass close, and owns every ledger edit and
the batch commit. Runaway guard armed at launch, per the standing
requirement.

Budget: 2.5M ceiling as a starting figure, re-derived after Task 1 and
again after Task 2 report their actual token spend; roughly 40 to 50
minutes of cumulative derive wall time across Tasks 1, 2, and 4. Spend
is scored at close against the re-derived ceiling and against zero
attended interactions, which is the target for this pass.
