# Rung 2, Pass 2a (Cassette Substrate and Capture) Implementation Plan

> **For agentic workers:** workflow mode per CLAUDE.md's "Build
> process" section: pass-execute chain per task, sequential. The
> conductor consumes reports and never reads diffs inline. Subagents
> never dispatch other agents; `code-simplifier` is the conductor's
> dispatch after a task's review accepts.

**Rung 2's split (the L rung's own planning, recorded here):**

- **Pass 2a (this plan):** the cassette substrate, the beets oracle
  adapter and derivation shim, the capture tool, and a proven
  end-to-end smoke corpus. Everything downstream depends on this
  machinery being trustworthy.
- **Pass 2b (planned just-in-time inside the evidence arc):** the
  structural stratum ported from beets' own test scenario shapes
  (roughly 60-120 cases), the component-test derivation tooling from
  `test_distance.py`, the assignment-tie detector, and the
  threshold-proximity flag.
- **Pass 2c (planned just-in-time):** the real-catalog and
  adversarial strata (roughly 80-150 plus 15-30 cases, adversarial
  sized against tier one and seeded from the known-oracle-bug
  classes), the stock-config stratum per ruling 3, the #6358
  alias-regression verification, the frozen-verdict fast gate joining
  the repo gate, the version-bump differential as a documented job,
  the quarterly re-capture chore doc, and the known-oracle-bug ledger
  and adjudication artifact formats with the first adjudication batch
  to Geoff.

**Goal:** Build and prove the corpus machinery: capture a MusicBrainz
response set once, freeze it, drive real pinned beets against the
frozen set, and extract the full oracle record (verdict, candidate
distances and ordering, applied tags) deterministically.

**Architecture:** `internal/harness` owns the fixture schema, the
replay stub, and the beets adapter; `cmd/harness` is the workstation
dev tool (capture, derive), never shipped to the box. The corpus
lives under `corpus/` at the repo root: per-case directories holding
raw frozen WS/2 JSON, pinned inputs, and derived expectations, under
a manifest that pins the beets, unidecode, jellyfish, and lap
versions and the evaluation date. A Python derivation shim, invoked
only through the adapter (the U7 boundary), drives pinned beets
2.13.1 against the stub and emits the oracle record as JSON. No live
network in any test; the capture tool alone touches the real WS, on
explicit invocation.

**Tech Stack:** Go 1.27 (schema, stub, adapter, tests); Python via a
uv-pinned venv holding beets==2.13.1 (derivation shim only; a
workstation dev dependency, documented, never part of the repo gate);
MusicBrainz WS/2 JSON.

**Spec:** `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified), ladder entry 2 and rulings 1-3. Primary
research inputs, read before implementing: the matcher mechanics
study (`research/2026-09-02-beets-matcher-mechanics.md`, especially
Q1 purity hazards and Q5 HTTP seam), the corpus report
(`research/2026-09-01-beets-conformance-corpus.md`), and the
evolution report's U6/U7/U8 lessons
(`research/2026-09-02-beets-evolution-lessons.md`).

**Token ceiling:** 5M (pass 2a is M). **Checkpoint:** STATUS after
Task 2, and at any split or question.

## Global Constraints

- go-conventions before any Go; python-conventions before any Python
  in the shim. The exemplary bar applies to both.
- Universality: no hostname, path, email, or personal name outside
  configuration. The capture tool's User-Agent contact is a flag or
  env value, never a literal.
- No live network in `go test` or the gate, ever. The capture tool
  hits the real WS only when a human runs it, at 1 request per
  second, with the identified User-Agent.
- `_salvage/` stays read-only. The box is untouched (rung 4 is the
  first box rung).
- Every expectation in the corpus is derived by executing pinned
  beets; nothing is hand-transcribed from test assertions or docs.
- Determinism is the product: any nondeterminism the mechanics study
  named (item order, evaluation date, dependency versions,
  multi-source interleaving) is pinned in the case or the manifest,
  and the smoke tests prove derivation twice yields identical bytes.
- Gate green and clean `git status` at every task end; never commit
  red; specific files only; commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`.

---

### Task 1: The cassette schema and corpus manifest

**Files:**
- Create: `internal/harness/cassette/cassette.go`,
  `internal/harness/cassette/cassette_test.go`, plus a package doc
  carrying the schema as the contract of record
- Create: `corpus/README.md` (the schema in prose, the layout, the
  request-matching rule), `corpus/manifest.json`

**Interfaces (later tasks and passes build against these exactly):**

```go
package cassette

// Case is one corpus case, loaded from its directory.
type Case struct {
    ID        string
    Input     Input           // items in pinned order
    Config    string          // beets config profile name ("estate", "stock")
    EvalDate  string          // ISO date the year penalty is computed against
    Responses []Response      // raw frozen WS/2 exchanges
    Expected  *Expected       // nil until derived
}

// Response is one frozen HTTP exchange. Matching is exact on method
// plus canonical path plus sorted query string (the U8 request-
// matching rule); recall-extension responses (requests pinned beets
// never issued) are legal members.
type Response struct {
    Method string
    URL    string // canonical: sorted query parameters
    Status int
    Body   json.RawMessage // raw WS/2 JSON, maximal inc= capture
}

// Expected is the oracle record the derivation shim emits.
type Expected struct {
    Verdict     string          // strong | medium | low | none
    Action      string          // apply | skip (quiet-mode mapping)
    Candidates  []Candidate     // distances, in beets' final order
    AppliedTags map[string]any  // the tag set beets wrote, flattened
}

func LoadCase(dir string) (*Case, error)
func (c *Case) Write(dir string) error
func LoadManifest(path string) (*Manifest, error)
```

- `Manifest` pins: beets version and commit, unidecode, jellyfish,
  and lap versions, the shim's own version, and per-case capture
  dates. Loading a corpus whose manifest disagrees with the venv's
  installed versions is an error, not a warning.
- The schema supports recall-extension cases from day one (U8): a
  case may carry responses for requests the oracle never made, and
  `Expected.Candidates` is keyed so a future retrieval-widening
  matcher is graded on the shared candidate subset.

**Behavior contract (tests first, against stubs, proven able to
fail):** round-trip Write-then-LoadCase yields identical bytes;
LoadCase rejects a case directory with a missing member, an
unparseable response body, or an eval date absent; response lookup by
(method, canonical URL) finds exact matches and errors on near
misses (unsorted query, trailing slash); manifest version mismatch
errors name the disagreeing component.

**Deliverable count: 3** (schema types and IO; matching rule;
manifest). Gate; commit.

---

### Task 2: The replay stub and the beets oracle adapter

The pass's correctness-critical core: the adapter and shim ARE the
oracle, and an error here poisons every expectation silently. Model:
`opus` for this dispatch (novel correctness-critical logic; the plan
cannot fully specify the beets-side extraction).

**Files:**
- Create: `internal/harness/stub/stub.go` + test (an
  `httptest`-style server serving a Case's responses per the matching
  rule; unknown request returns 404 AND is recorded as a miss the
  test can assert on—a silent 404 is how replay drift hides)
- Create: `internal/harness/oracle/oracle.go` + test (the adapter:
  the ONLY code that knows how to invoke beets—the U7 boundary)
- Create: `tools/derive_oracle.py` (the shim) plus
  `tools/README.md` and the uv environment spec pinning
  beets==2.13.1 with unidecode, jellyfish, lap versions recorded
  into the manifest

**The adapter contract:**

```go
package oracle

// Derive runs pinned beets against the case's frozen responses and
// returns the oracle record. It is deterministic: the same case
// yields byte-identical Expected on every run.
func Derive(ctx context.Context, c *cassette.Case, venv string) (*cassette.Expected, error)
```

- The adapter starts the stub on a loopback port, renders a beets
  config from the case's profile with `musicbrainz.host` pointed at
  the stub (`https: no`, ratelimit raised—the mechanics study's
  documented seam), materializes the case's items as files or item
  fixtures per the shim's input contract, and invokes the shim inside
  the venv.
- The shim drives beets' autotag API directly (not `beet import`
  CLI): it must emit verdict, recommendation, per-candidate distances
  in final order, and the applied tag set as one JSON document. The
  mechanics study's pipeline map is the guide; the shim is small,
  auditable, and version-stamped, because it is the half of the
  oracle that runs inside Python.
- **Evaluation-date pinning is settled inside this task** (the
  mechanics study's one true impurity): prefer `faketime` around the
  shim; if unavailable on the workstation, an equivalent mechanism
  chosen and documented. Acceptance is behavioral: a smoke case
  exercising the year penalty derives identical Expected under two
  different wall-clock dates.
- **Determinism proof:** `Derive` twice on one case yields
  byte-identical output; the test runs it and diffs.

**Deliverable count: 4** (stub with miss-recording; adapter; shim +
venv; date pinning). Gate; commit.

---

### Task 3: The capture tool

**Files:**
- Create: `cmd/harness/main.go` (dev tool root; package doc states it
  is workstation tooling, never deployed), `cmd/harness/capture.go` +
  test for the request-planning logic (the network call itself is
  exercised only manually)

**Behavior contract:**
- `harness capture --mbid <release> --out corpus/cases/<id>` fetches
  the release with the maximal practical `inc=` set (U6), honors the
  1 request-per-second anonymous floor, sends the identified
  User-Agent (`dubplate-harness/<version> ( <contact> )`, contact
  from flag or env, never a literal), follows the pseudo-release
  double fetch (a `Pseudo-Release` status with a backward
  transl-tracklisting relation triggers capture of the linked actual
  release too), and writes raw bodies untouched into the case
  directory with canonical URLs.
- Search-request capture: the tool also captures the search
  responses the oracle will issue for a case's items (the shim's
  request log from a dry derivation names them; the capture loop is:
  capture search, derive, capture any recorded misses, re-derive
  until the miss list is empty). That loop is the documented
  procedure in `corpus/README.md`.
- A batched owner item, recorded not asked mid-task: whether to
  register a MusicBrainz account application for higher-than-
  anonymous rate limits; anonymous 1 rps is sufficient for pass 2a's
  smoke corpus.

**Deliverable count: 3** (capture verb; pseudo-release rule; the
capture-derive loop doc). Gate; commit.

---

### Task 4: The smoke corpus and the end-to-end proof

**Files:**
- Create: `corpus/cases/<id>/...` for 5 to 10 smoke cases;
  `corpus/README.md` gains the case inventory table
- Create: `internal/harness/corpus_test.go` (the fast-gate seed: a
  Go test that walks `corpus/`, validates every case against the
  schema, verifies manifest pins, and replays each case's responses
  through the stub asserting zero misses against the recorded
  request log)
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (pass close per the
  ledger discipline)

**The smoke set, chosen for coverage of the machinery (not the
matcher):** one clean confident match; one should-ask; one
should-reject; one various-artists album; one multi-disc release; one
non-Latin release with a pseudo-release follow-up; one case
exercising the year penalty (for the date-pinning proof). Every
expectation derived by the adapter, never by hand.

**Acceptance criteria:** the corpus test is in the gate and green; a
second full derivation of every smoke case yields byte-identical
Expected (run it, record it in the report); the pseudo-release case's
directory visibly contains both releases' responses; STATUS stays
within its line cap and points at pass 2b's just-in-time planning.

**Deliverable count: 4** (smoke cases; corpus gate test; derivation
determinism proof; ledger close). Gate; commit.

---

## Execution model

Workflow mode per the standing grant: pass-execute chain, sequential
(each task consumes the previous task's interfaces). Task 2 dispatches
on `opus`; the rest on `sonnet`; `diff-reviewer` per task as always.
The conductor dispatches `code-simplifier` over each accepted task's
changed Go (and the shim) after review, before the pass close.
Runaway guard armed at launch. Budget: 5M ceiling, spend scored at
close. Owner touchpoints: zero planned; the MusicBrainz registration
question rides the next checkpoint bundle.
