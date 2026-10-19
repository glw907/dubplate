# Rung 2, Pass 2a (Cassette Substrate and Capture) Implementation Plan

> **For agentic workers:** workflow mode per CLAUDE.md's "Build
> process" section: pass-execute chain per task, sequential. The
> conductor consumes reports and never reads diffs inline. Subagents
> never dispatch other agents; `code-simplifier` is the conductor's
> dispatch after a task's review accepts.

Revision 2, 2026-09-02: folds the adversarial plan review
(`research/2026-09-02-greenfield-spec-reviews/rung-2a-plan-review.md`),
all five blockers, eleven majors, and eight minors. The review's
"Verified sound" section records the mechanisms implementers must not
re-litigate: the `musicbrainz.host` stub seam, the miss-recorded
capture-derive loop's termination, the test-only harness package, the
manifest quadruple stance, and the file-free extractability of the
recommendation half of the oracle.

**Rung 2's split (the L rung's own planning, recorded here):**

- **Pass 2a (this plan):** the cassette substrate, the config
  profiles, the beets oracle (adapter and shim), the capture tool,
  and a proven end-to-end smoke corpus.
- **Pass 2b (planned just-in-time inside the evidence arc):** the
  structural stratum ported from beets' own test scenario shapes
  (roughly 60-120 cases), the component-test derivation tooling from
  `test_distance.py`, the assignment-tie detector, and the
  threshold-proximity flag. The schema below already carries their
  inputs (per-component penalties, per-key distances, the assignment
  mapping, extras), so 2b consumes the record rather than rebuilding
  the shim.
- **Pass 2c (planned just-in-time):** the real-catalog and
  adversarial strata (roughly 80-150 plus 15-30 cases, adversarial
  sized against tier one and seeded from the known-oracle-bug
  classes), the stock-config stratum per ruling 3, the #6358
  alias-regression verification, the file-level applied-tag
  derivation (the stated rung 8 precondition: audio fixtures,
  copy-into-tempdir import discipline, mediafile read-back), the
  version-bump differential as a documented job, the quarterly
  re-capture chore doc, and the known-oracle-bug ledger and
  adjudication artifact formats with the first adjudication batch to
  Geoff.

**Pass inputs (required before Task 5 can run):** a MusicBrainz
contact for the identified User-Agent, supplied by Geoff as
`DUBPLATE_HARNESS_CONTACT` (house pattern: `~/.local/secrets` or
`~/.bashrc`; a reachable address or URL per MusicBrainz's terms), and
his answer on registering a MusicBrainz application versus staying at
the anonymous 1 request per second. One batched ask covers both; the
plan does not run Task 5 without the contact.

**Goal:** Build and prove the corpus machinery: capture a MusicBrainz
response set once, freeze it byte-for-byte, drive real pinned beets
against the frozen set, and extract the full oracle record
(recommendation, action, per-candidate distances and penalties,
ordering, assignment, applied fields) deterministically.

**Architecture:** `internal/harness/cassette` owns the fixture
schema and IO; `internal/harness/stub` the replay server;
`internal/harness/oracle` the beets oracle—adapter AND shim
together, because the shim is the code that actually knows beets and
U7's criterion ("a beets 3.x bump edits only the adapter package")
must be true as written. `cmd/harness` is the workstation dev tool,
never shipped. The corpus lives under `corpus/`: per-case directories
with verbatim response files, a committed request log, pinned inputs,
and the derived record; `corpus/profiles/` holds the two beets config
profiles; `corpus/manifest.json` pins versions. Derivation runs only
behind a build tag via `make derive`—the gate protects the frozen
records and never touches Python.

**Tech Stack:** Go 1.27 (schema, stub, adapter, tests); Python via a
uv-pinned venv (beets==2.13.1 with unidecode, jellyfish, lap, and
numpy versions recorded into the manifest; the shim only; never in
the repo gate); ruff as the Python lint leg; MusicBrainz WS/2 JSON.

**Spec:** `docs/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified), ladder entry 2 and rulings 1-3. Primary
research, read before implementing: the mechanics study
(`research/2026-09-02-beets-matcher-mechanics.md`; Q1 purity, Q5
seam), the corpus report, and U6/U7/U8 in the evolution report.

**Token ceiling:** 5M. **Checkpoint:** STATUS after Task 3, and at
any split or question.

## Global Constraints

- go-conventions before any Go; python-conventions before any Python.
  ruff joins the gate for tracked `*.py` (Task 3), so the shim's bar
  is enforced, not aspirational.
- Universality: no hostname, path, email, or personal name outside
  configuration; the User-Agent contact comes from
  `DUBPLATE_HARNESS_CONTACT`, never a literal.
- No live network in `go test` or the gate, ever. The capture tool
  hits the real WS only when a human runs it, at 1 request per
  second, identified. The gate never requires uv, the venv, or beets;
  derivation tests carry `//go:build oracle` and run via
  `make derive` only.
- **Oracle process contract:** one shim process per (case, profile);
  `PYTHONHASHSEED=0` in the shim environment; every profile loads
  exactly one metadata source (`plugins: musicbrainz`), asserted by
  the adapter; the shim echoes into its output the effective match
  weights and thresholds it ran under, its observed evaluation date,
  and a hash of its effective config—and the adapter fails loudly
  when any echo disagrees with the rendered profile or
  `Case.EvalDate`. beets caches config per process (the study's
  Q1.8), so the per-process rule is a correctness contract, not a
  convention.
- Every expectation is derived by executing pinned beets code paths;
  nothing is hand-transcribed from test assertions or docs. Where
  beets' own function is importable (the quiet-mode action mapping),
  the shim calls it rather than restating it.
- Response bodies and the request log are stored as verbatim files,
  byte-for-byte as the wire delivered them; no JSON re-encoding of
  captured bytes anywhere (Go's encoder compacts and HTML-escapes
  `RawMessage`, which the review proved breaks byte identity).
- `_salvage/` stays read-only; the box is untouched; derivation never
  writes inside `corpus/` case directories except through the schema
  writer.
- Gate green and clean `git status` at every task end (`.gitignore`
  grows the venv, `__pycache__`, and derivation-scratch lines in
  Task 3). Never commit red; specific files only; commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`.

---

### Task 1: The cassette schema, corpus layout, and manifest

**Files:**
- Create: `internal/harness/cassette/cassette.go`,
  `internal/harness/cassette/cassette_test.go`
- Create: `corpus/README.md`, `corpus/manifest.json`

**The complete interface block (later tasks and passes build against
these exactly; this revision closes the review's blocker 4):**

```go
package cassette

// Case is one corpus case, loaded from its directory.
type Case struct {
    ID       string
    Input    Input      // items in pinned order; no file references at 2a
    Profile  string     // profile name under corpus/profiles ("estate", "stock")
    EvalDate string     // ISO date the year penalty is evaluated against
    Requests []Request  // the committed request log: an expectation, not scaffolding
    Extra    []string   // response files marked recall-extension (U8)
    Expected *Expected  // nil until derived
}

// Input is the matcher's input: synthetic beets items, file-free.
type Input struct {
    Items []ItemFields
}

// ItemFields is the field bag one item carries into matching
// (title, artist, album, track, disc, length, year, mb ids, ...),
// serialized as a flat JSON object; length is set directly.
type ItemFields map[string]any

// Request is one logged oracle request in canonical form.
type Request struct {
    Method    string
    PathQuery string // path + "?" + query, canonicalized per the README rule
    Response  string // response filename within the case dir
}

// Expected is the oracle record the shim emits, decoded.
type Expected struct {
    Recommendation string          // strong | medium | low | none
    Action         string          // apply | skip; executed via beets' own quiet-mode mapping
    EchoedDate     string          // the date the shim observed; must equal EvalDate
    ConfigHash     string          // hash of the shim's effective config
    Candidates     []Candidate     // in beets' final order
    Assignment     map[int]int     // item index -> track index, best candidate
    ExtraItems     []int           // unmatched item indexes, in beets' sort order
    ExtraTracks    []int           // unmatched track indexes, in beets' sort order
    AppliedFields  map[string]any  // post-apply in-memory item fields, best candidate
}

// Candidate is one scored release candidate.
type Candidate struct {
    Source    string             // metadata source name ("MusicBrainz")
    AlbumID   string             // the U8 shared-subset key with Source
    Distance  float64
    Penalties map[string]float64 // per-key weighted distances
    Tracks    []float64          // per-track distances, assignment order
}

// Manifest pins everything the corpus's meaning depends on.
type Manifest struct {
    SchemaVersion string
    Beets         string // version and commit
    Unidecode     string
    Jellyfish     string
    Lap           string
    Numpy         string
    ShimVersion   string
    ProfileHashes map[string]string // profile name -> content hash
}

func LoadCase(dir string) (*Case, error)
func (c *Case) Write(dir string) error
func LoadManifest(path string) (*Manifest, error)
```

- **AppliedFields is field-level, by decision** (the review's blocker
  1): it is the output of beets' apply step on in-memory items, which
  the study verified needs no files, no importer, no library. The
  file-level tag set (what `item.write()` puts in a FLAC via
  mediafile) is explicitly NOT in this record; it is pass 2c's
  deliverable and rung 8's stated precondition. Rung 8 inherits a
  named debt, not a surprise.
- **Bodies live as verbatim files** (`responses/NNN.json`), written
  with `os.WriteFile` and never re-encoded; `Request.Response` and
  `Case.Extra` reference them by name. Byte-identity round-trip
  applies to our own metadata files via a stable encoder and to
  response files trivially.
- **The canonicalization rule** lives in `corpus/README.md` and the
  package doc, and it is encoding-preserving: sort the raw encoded
  `key=value` pairs byte-wise; never decode-and-re-encode (the `+` in
  `inc=artists+media` and beets' escaped Lucene queries do not
  survive a re-encode). The README records the three wire facts the
  rule must survive: `fmt=json` is forced on every request, `inc=`
  values join with `+`, search queries carry Lucene escapes.
- `LoadManifest` parses and validates shape only; reconciling against
  a live venv is derivation-time work and lives in Task 3's
  `Reconcile` (the review's major 8 split).

**Behavior contract (tests first, against stubs, proven able to
fail):** Write-then-LoadCase round-trips byte-identically; LoadCase
rejects a missing member, an unreadable response file, an eval date
absent, a `Request.Response` naming a nonexistent file, and an
`Extra` entry naming a consumed response; near-miss canonical lookups
(unsorted pairs, trailing slash) fail; manifest parsing errors name
the missing field. Plus the negative fixture: a deliberately
corrupted case under `testdata/` that the loader must reject.

**Deliverable count: 4** (schema types and IO; verbatim-file body
store; canonicalization rule; manifest parsing). Gate; commit.

---

### Task 2: The replay stub, the config profiles, and the adapter's Go side

Proven against a fake shim so the Go contract lands reviewable on its
own (the review's major 10 split). Model: `sonnet`.

**Files:**
- Create: `internal/harness/stub/stub.go` + test
- Create: `corpus/profiles/estate.yaml`, `corpus/profiles/stock.yaml`,
  and the derivation notes section in `corpus/README.md`
- Create: `internal/harness/oracle/oracle.go` + test (adapter Go
  side), `internal/harness/oracle/testdata/fakeshim` (a script
  emitting a fixed, well-formed oracle JSON document)

**Interfaces:**
- Consumes: the cassette package exactly as Task 1 shipped it.
- Produces:

```go
package oracle

// Derive runs the oracle for one case under one profile: it renders
// the profile, starts the stub, launches ONE shim process with
// PYTHONHASHSEED=0, decodes and validates the record (echoed date,
// config hash, single-source assertion), and returns it.
func Derive(ctx context.Context, c *cassette.Case, opts Options) (*cassette.Expected, error)

// Options carries the venv path and the shim override the fake-shim
// tests use.
type Options struct {
    Venv string
    Shim string // path; empty means the package's real shim
}
```

**The profile artifact (the review's blocker 2):**
- `estate.yaml` derives from `config/beets-config.yaml` under a
  documented transform, recorded in `corpus/README.md`: every
  verdict-moving key is preserved verbatim (the `match:` and
  `musicbrainz:` surfaces, the quiet trio, `original_date`,
  `per_disc_numbering`); box paths are replaced by placeholders the
  adapter substitutes at derivation time; `plugins:` reduces to
  `musicbrainz` alone (the single-source constraint; fetchart,
  embedart, and replaygain are matching-inert per the study and
  import-active, so they cannot ride); `group_albums` is dropped
  (the corpus tests below the importer's grouping boundary). Each
  stripped key is listed with its reason.
- `stock.yaml` is the defaults profile: `plugins: musicbrainz` and
  nothing else.
- A profile edit is a corpus-revalidation event exactly like a beets
  bump; the manifest's `ProfileHashes` is what detects it.

**Stub behavior:** serves a Case's response files per the canonical
rule; an unknown request returns 404 AND is recorded in a miss list
the caller reads—the miss record is what makes the capture-derive
loop safe (review, Verified sound). Test covers hit, near-miss
rejection, miss recording, and the ephemeral-port lifecycle.

**Adapter behavior (against the fake shim):** renders the profile
with the stub's host substituted (`musicbrainz: {host, https: no,
ratelimit: 100}`—the study's documented seam; note the client's
unconditional 0.25s inter-request floor still applies and is budgeted
in derivation wall-time, not assumed away); launches exactly one
process per call; validates the record (EchoedDate equals
`Case.EvalDate`, ConfigHash present, exactly one metadata source in
the rendered profile) and fails loudly on each; decodes into
`cassette.Expected`. Tests use the fake shim to prove every
validation can fire.

**Deliverable count: 4** (stub; profiles + transform doc; adapter
against fake shim; validation contract). Gate; commit.

---

### Task 3: The real shim, the venv, and the determinism proof

The pass's correctness-critical core; the shim IS the oracle. Model:
`opus` (novel correctness-critical logic the plan cannot fully
specify).

**Files:**
- Create: `internal/harness/oracle/derive_oracle.py` (the shim lives
  INSIDE the oracle package—the U7 boundary is "the oracle package,
  shim included", so a beets 3.x bump edits only this package),
  `internal/harness/oracle/derive_test.go` (`//go:build oracle`),
  `tools/venv/README.md` and the uv environment spec (pinned
  beets==2.13.1; unidecode, jellyfish, lap, numpy versions captured
  into `corpus/manifest.json`)
- Modify: `.gitignore` (venv dir, `__pycache__`, derivation scratch),
  `Makefile` (a `derive` target running the tagged tests; `check`
  unchanged), `scripts/check.sh` (a ruff leg over tracked `*.py`;
  ruff availability is a hard requirement like shellcheck, per "a
  check either gates or does not exist")
- Create: `(*cassette.Manifest).Reconcile` in the oracle package
  (derivation-time venv-version reconciliation, error names the
  disagreeing component)

**The shim's contract:**
- Drives beets' autotag API for the recommendation half
  (candidates, distances, per-key penalties, per-track distances,
  assignment, extras—the study verified all of it is extractable
  without an importer), calls beets' own quiet-mode mapping function
  for `Action` (executed, never restated; if that import proves
  unstable at implementation time, the fallback is dropping `Action`
  from the shim's record and deriving it in Go from Recommendation
  plus the profile, documented as a dubplate-side derivation—the
  implementer reports which branch was taken), and applies the match
  to the in-memory items for `AppliedFields`.
- **Date pinning is internal and self-verifying** (the review's
  major 2 inversion): the shim patches Python's date inside its own
  process to `Case.EvalDate`—no faketime, no system layering—and
  reports the date it actually observed; the adapter's equality
  assertion is what makes the acceptance test non-circular. The
  pinning must not freeze monotonic time (the client's 0.25s floor
  waits on elapsed time; a frozen clock deadlocks it).
- Emits one JSON document to stdout; deterministic serialization
  (sorted keys, fixed float formatting) is the shim's own duty.
- Echoes effective weights, thresholds, observed date, and the
  config hash per the global process contract.

**The determinism proof (the review's major 1 form):** derive one
smoke-shaped fixture case twice in two separate processes and diff
the shim's RAW STDOUT bytes—not the Go-normalized record—so
Python-side ordering or formatting drift can actually fail the test.
Runs under the `oracle` build tag via `make derive`.

**Deliverable count: 4** (shim; venv + Reconcile; date pinning;
raw-stdout determinism proof). Gate (which never runs the tagged
tests) plus `make derive` green; commit.

---

### Task 4: The capture tool

**Files:**
- Create: `cmd/harness/main.go`, `cmd/harness/capture.go` + tests

**Behavior contract:**
- Two verbs, matching the loop's two halves: `harness capture --mbid
  <release> --out corpus/cases/<id>` seeds a case (maximal practical
  `inc=` per U6, the pseudo-release double fetch, verbatim body
  writes); `harness capture --case <dir>` consumes a derivation's
  miss list and captures exactly those requests (the review's minor
  5: the loop is case-directory-driven and needs its own verb).
- Wire discipline: 1 request per second, identified User-Agent
  `dubplate-harness/<version> ( $DUBPLATE_HARNESS_CONTACT )`; the
  tool refuses to run without the contact set.
- The capture-derive loop is documented in `corpus/README.md` as
  producing a recorded fixed point, not a stable request set (a
  successful ID lookup with a strong recommendation retires the
  search path, orphaning earlier captures into `Extra` legally—the
  review's caveat).
- Tests cover request planning, canonical-form writing, and the
  round-trip the review's minor 7 demands: a response set written by
  the tool's writer loads through `cassette.LoadCase` and matches
  the stub's canonical lookup, so a canonicalization disagreement
  surfaces here, not as a mysterious miss after live captures.

**Deliverable count: 3** (two verbs; wire discipline; writer/loader
round-trip test). Gate; commit.

---

### Task 5: The smoke corpus and the pass close

Blocked on the pass input: `DUBPLATE_HARNESS_CONTACT` must exist
before any live capture.

**Files:**
- Create: `corpus/cases/<id>/...` for 5 to 10 smoke cases;
  `corpus/README.md` gains the case inventory
- Create: `internal/harness/corpus_test.go` (the gate-resident
  static consistency check, stated in its true form per the review's
  major 7: every logged request resolves to exactly one response;
  every response is consumed by the log or listed in `Extra`; the
  manifest parses and matches the schema version; every case loads
  clean—plus the corrupted-fixture negative case)
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (pass close);
  `scripts/check.sh` header comment (its "build, vet" claim is
  already stale against the Makefile—fix in passing);
  `CLAUDE.md` (the Python stack-rule line: shim under
  python-conventions, ruff in the gate); and add the gitleaks leg to
  `scripts/check.sh` (the binary is installed, the corpus is the
  repo's first large data directory, and a check either gates or
  does not exist—this makes the safety rail real).

**The smoke set** (chosen for machinery coverage, not matcher
coverage): one clean confident match; one should-ask; one
should-reject; one various-artists album; one multi-disc release; one
non-Latin release—with a pseudo-release follow-up if a qualifying
release (status Pseudo-Release with a backward transl-tracklisting
relation) is found within the task's budget, otherwise a plain
non-Latin case now and the pseudo-release case filed to 2c's
inventory explicitly; one case exercising the year penalty (the
date-pinning proof rides it). Every expectation derived by the
adapter; a second full derivation of every case diffs byte-identical
at the shim-stdout level, recorded in the report.

**Deliverable count: 4** (smoke cases; corpus gate test; determinism
record; ledger close + gate housekeeping). Gate; commit.

---

## Execution model

Workflow mode per the standing grant: pass-execute chain, sequential
(each task consumes the previous task's interfaces). Task 3 dispatches
on `opus`; the rest on `sonnet`; `diff-reviewer` per task. The
conductor dispatches `code-simplifier` over each accepted task's
changed Go (and ruff-checked Python) after review, before the pass
close. Runaway guard armed at launch. Budget: 5M ceiling, spend
scored at close. Owner touchpoints: exactly one, batched—the
MusicBrainz contact plus the registration question, required before
Task 5.
