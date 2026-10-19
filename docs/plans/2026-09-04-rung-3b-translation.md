# Rung 3, Pass 3b (The Translation Layer, the AlbumInfo Ledger, and the `defect_free` Block) Implementation Plan

**Revision 2, 2026-09-04:** folds the two adversarial plan reviews,
merged in `research/2026-09-04-rung-3b-plan-review.md`: blockers B1 to
B5, majors M-a through M-l, minors m1 to m20. Four reviewer claims are
rejected with the evidence in that document's header, and four items are
conductor-owed rather than folded silently; each of those four was applied
provisionally and **all four were ruled as recommended by the conductor on
2026-09-04 before launch** (see the marks inline).

The load-bearing resolutions. **Task 2's derivation is rewritten from
what it must actually compute** (B1): `DefectFreeSubject` is declared,
the input envelope is the case's own release-lookup response plus
`input.items` plus `eval_date`, the four beets steps are named, the
assignment is re-solved under the substituted table rather than read off
the record, and the monkeypatch target is the module object because
`beets.autotag.distance` resolves to a function. Its precision contract
splits in two: the block stores the derivation's own unrounded float64,
the plan's table is the prose cross-check, and the both-ways check
compares bit-identically against `case.json` (M-a). **R15's
release-event measurement was false and is replaced by the measured
figures at all three sites it appeared** (B2). **`Track.Position` becomes
a two-shape decode**, because 297 committed track positions are JSON
strings and three ledger subjects break on `int` (B3). **Task 8's
profile decoder becomes a profile-name-to-overlay table plus a key-name
drift guard**, because no YAML parser exists and the plan forbids adding
one (B4). **The `"Media"` sentinel has two routes, not three**; a single
null-format medium yields `None`, and Task 5's third fixture now asserts
that non-sentinel outcome (B5). `CheckAlbumInfoLedger`'s check 6 stops
deriving its own expectation and reads a named
`AlbumInfoPseudoSubjects` list (M-b). `Validate` grows from three rules
to five, closing two latent panics on external input (M-d, M-e). The
`defect_free` write-back names its mechanism, a targeted Python writer,
because a Go re-serialize would reorder `corrected_verdict` on eighteen
rows (M-g).

Three folds take an adaptation rather than the literal fix, each named
where it lands: **m3** keeps `AlbumInfoSubject.Why` and gives it a reader
rather than deleting the field on the `FileTagSubject` precedent;
**B4** keeps the unknown-key rule as a text drift guard rather than
dropping it with the YAML parse; and **M-j** downgrades the engine
contested-surface row from a launch blocker to a conductor-owed note,
because the phase-1 lane table already awards lane E2 that territory.

**Conductor rulings, 2026-09-04, all taken as the fold recommended** (each
marked inline; listed in the fold's `conductor_rulings_needed`): the nineteenth ledger subject
`real-noise-dylan-bootleg` and the row floor moving 45 to 50 (M-h); the
three contested-surface rows the table owes (M-j); the spec's 3b tree
column correction (m14); and `docs/STATUS.md`'s 3a spend figure, which
disagrees with `docs/HISTORY.md`'s itemized one (m1).

**Revision 2.1, 2026-09-04 (conductor ruling on Task 5's escalation):**
Task 5's named mutation 4 (dropping the `preferredAlias` guard on the
track-title override) is provably unobservable through `TrackInfo.Title`,
because `keyWithPreferredAlias` resolves the same aliases under the same
tunables immediately afterward; the implementer declared it rather than
fabricating a proof and the `diff-reviewer` reproduced the non-firing
independently. Ruling: the guard stays, mirroring beets' shape at
`musicbrainz.py`'s title override, its non-effect is stated in
`tracks.go`'s doc comment, and mutation 4 is struck from Task 5's proof
set and recorded in the pass report as a declared unobservable mutation,
not a coverage gap. Task 5 is accepted at `251ab14` with no code change.

**The merged review's 26-item "Verified sound" list is a second
do-not-re-litigate record** for this pass: all sixteen translation
contracts, the four measured beets tables, the ten `defect_free` figures
and both pattern tables, R6's and R7's arithmetic, the alias-path
leverage figures, R12's predicate, the harness API signatures, the gate
coverage claims, both architectural invariants, and the plan's structural
discipline are cited below rather than re-derived. An implementer who
believes one of them is wrong reports a finding and implements the record
anyway.

**Revision 1, 2026-09-04.** Drafted against the rung 3 spec at **revision
2** (`docs/specs/2026-09-03-rung-3-go-matcher.md`), which is
ratified on its six rulings and review-folded on all nineteen findings of
`research/2026-09-04-rung-3-spec-review.md`. That review's **"Verified
sound" list is the do-not-re-litigate record**: its twenty items are cited
here rather than re-derived, and an implementer who believes one of them is
wrong reports a finding and implements the record anyway. The same standing
applies to 3a's own review record and to `docs/HISTORY.md`'s 3a entry,
whose ten rediscover-guards are cited below rather than re-measured.

This plan's primary input is
`docs/research/2026-09-04-rung-3b-planning-inventory.md`, a
research report that read pinned beets 2.13.1 in `tools/venv/.venv`, the
harness and corpus as landed at `f40a669`, and the engine as 3a left it.
Every mechanic below is cited to the tree, and the inventory's twenty-one
deduplicated open questions are **resolved in this plan**, each with the
evidence and the rejected alternative, in the "Resolutions this plan takes
rather than discovers" section.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section: pass-execute chain per task, sequential. The conductor consumes
> reports and never reads diffs inline. Subagents never dispatch other
> agents; `code-simplifier` is the conductor's dispatch after a task's
> review accepts. The conductor also owns the ledgers (`docs/STATUS.md`,
> `docs/HISTORY.md`, `ROADMAP.md`), the lane registry, and the batch commit
> of anything a task drafts into them, including this pass's four
> `corpus/README.md` corrections, which an implementer drafts and never
> writes. Rung 3 is not an adversarial-cycle rung (those are 5, 12, and
> 16), so review scales to the in-chain `diff-reviewer` per task plus the
> pass-end gate. **This pass runs in the MAIN TREE**, and nothing here may
> be executed from a worktree: the spec's tree column for 3b is "Main tree:
> writes `corpus/` and `internal/harness/oracle`"
> (`docs/specs/2026-09-03-rung-3-go-matcher.md:329`), and the
> standing policy is that corpus and harness work never moves to a
> worktree (`docs/specs/2026-09-03-parallel-build-approach.md`,
> lane-launch checklist item 5).

## Primary inputs

Read these before Task 1, in this order:

1. **`docs/research/2026-09-04-rung-3b-planning-inventory.md`,
   in full.** It is this plan's argument. Sections (a) 1–14 are the beets
   mechanics with line citations against `tools/venv/.venv`; (b) 1–11 is
   the harness and corpus as landed; (c) 1–3 is the engine as landed and
   the proposed Go surface; (d) reconciles all three. Do not re-derive what
   it settles; cite it and verify the line if you carry one forward.
2. **`docs/specs/2026-09-03-rung-3-go-matcher.md`, revision
   2**, in full. The pass-split table's 3b row (`:329`) is this plan's
   scope statement; ruling A (`:71-88`), ruling D (`:109-113`), ruling E
   (`:115-124`), "The translation layer" (`:210-220`), and "Assertion
   discipline" (`:316-321`) are the sections this pass implements.
3. **`docs/HISTORY.md`'s "Rung 3, pass 3a" entry, in full**: the component
   coverage table, the ten rediscover-guards, and the two items its
   "Carried" list addresses to 3b. Guard 3 (the `SD_PATTERNS`-as-a-value
   seam and its two measured figures) is load-bearing for Task 2.
4. **`docs/plans/2026-09-04-rung-3a-scoring-core.md`**, for its
   twelve arithmetic contracts and its Global Constraints. This plan does
   not restate them; it adds the translation contracts below.
5. The tree, read-only: `internal/engine/matcher/info.go`,
   `tunables.go`, `strdist.go`; `internal/harness/analysis/adjudication.go`,
   `file_tags.go`, `oraclebug.go`, `ledger.go`;
   `internal/harness/oracle/oracle.go`, `ties.go`, `file_tags.go`,
   `component.go`; `internal/harness/conform/defaults.go`, `arch_test.go`,
   `doc.go`; `internal/harness/cassette/cassette.go`;
   `internal/harness/corpus_test.go`.
6. **`internal/harness/oracle/derive_file_tags.py` and
   `internal/harness/oracle/derive_ties.py`, in full, before Task 1.**
   `derive_file_tags.py` is the subject-list-driven shape; `derive_ties.py`
   is the one narrow entry point that needs the stub and a rendered
   profile, which is what Task 1's entry point also needs. Reading oracle
   territory is permitted and expected; 3b also **writes** it, which 3a did
   not.
7. `CLAUDE.md`'s Build process and Gate sections, `docs/STATUS.md`'s lane
   registry, and `corpus/README.md`'s adversarial-stratum and
   config-profile sections.
8. **The pinned venv itself**, `tools/venv/.venv/lib/python3.13/site-packages/`,
   for `beetsplug/musicbrainz.py`, `beetsplug/_utils/musicbrainz.py`,
   `beets/autotag/hooks.py`, and `beets/autotag/distance.py`. Every
   translation contract below cites it. Short offline probes with
   `tools/venv/.venv/bin/python` are permitted and expected; `make derive`
   is not, except where a task names it.

## Goal

Ship the translation half of the Go matcher, graded field by field. Land
`internal/engine/musicbrainz`, which holds the WS/2 release vocabulary, key
normalization and the `relations` regrouping, per-medium track extraction
with pregap and the two track-kind filters and the track-level overrides,
alias-driven rewriting of titles and artist credits, release-event
selection feeding country and date, and the pseudo-release merge. Land
also the two derived-corpus deliverables the spec's rulings E and A/M6
name: the `AlbumInfo` ledger (`corpus/analysis/album_info.json`) with its
Python entry point, its gate-resident checkers, its oracle-tagged
re-derive test, and the field-by-field grading of the Go translation
against it; and the `defect_free` block on `AdjudicationEntry` with its
derivation, its checkers, and its first population of the ten
string-heuristic cases.

Nothing here decides a verdict, issues a request, or computes a distance.
The WS/2 client, the Lucene query construction, the limiter, and the
stub's `Extra` replay channel are 3c's. The assignment solve, admission
filters, recommendation, `max_rec`, verdict-to-action, the compat
switches, and tiers 1 and 2 are 3d's. The scoring core is 3a's and is
shipped.

## Owner touchpoints: NONE

3b needs no ruling, no adjudication, no capture contact, and no network.

The six rulings this rung rests on were ratified on 2026-09-03. The first
adjudication batch is fully decided: as read at drafting time,
`corpus/analysis/adjudications.json` carries 21 rows at 18 `oracle-wrong`
/ 3 `oracle-correct` with none pending, and
`corpus/analysis/oracle_bugs.json` carries six hypotheses at four
`oracle-wrong` / two `oracle-correct` with none pending, every row
adjudicated by `Geoff` on 2026-09-04 (inventory §(b).8). That is precisely
why the `defect_free` block's ten seed figures are settled data rather
than work waiting on a batch: ruling A's scheduling rule
(spec `:85`) has nothing left pending on the two string-heuristic
hypotheses this pass populates.

**`disposition` stays owner-only, and this pass writes none.** The
`defect_free` block is **derived data, never a disposition**: it is
computed through pinned beets and written by `make derive`, it carries no
verdict about whether the oracle is wrong, and `checkAdjudicationEntry`'s
existing sentinel rule (`internal/harness/analysis/adjudication.go:140-142`)
is untouched by this pass. **Task 2's** population adds only the ten
rows' `defect_free` field and leaves all ten existing fields
byte-identical: `id`, `subject`, `suspected`, `citation`,
`release_truth`, `disposition`, `corrected_verdict`, `reason`,
`adjudicator`, and `date` (`internal/harness/analysis/adjudication.go:45-56`;
fold, M-f and m5). An implementer that finds itself editing any of those
ten fields stops and reports.

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition named below resolves to the conductor. Only a
conductor decision to escalate reaches Geoff. The standing 80%-of-ceiling
flag CLAUDE.md requires still applies and is contingent, not a touchpoint
this plan expects to spend.

## Spec

`docs/specs/2026-09-03-rung-3-go-matcher.md` (revision 2), and
behind it `specs/2026-09-02-greenfield-server-build.md` (revision 2.1,
ratified), ladder entry 3. This pass ships deliverable 2, quoted in full:
"the WS-response-to-matcher-input translation layer, a first-class
deliverable rather than plumbing. Key normalization, per-medium track
extraction, pregap and data/video-track handling, alias-driven title and
artist rewriting, and release-event selection all move verdicts before any
distance is computed" (spec `:60`). Ruling E's derived `AlbumInfo` ledger
and ruling A / M6's `defect_free` block ship with it. Deliverables 1, 3,
and 5 are 3a (shipped), 3d, and 3c.

**The rung 15 forward criterion travels in this pass's report**, per the
spec's m7 fold, so rung 15's plan reads it from the ledger rather than
rediscovering it:

> Rung 15 re-runs `internal/harness/conform`'s three tiers over the
> complete corpus, at the matcher commit as landed, and its acceptance is
> the launch bar's matcher clause read against that run: zero false
> confident matches over every stratum including stock-config, and zero
> unadjudicated beets-right/Go-wrong cases. Rung 3's own acceptance is
> provisional against the corpus at its then-current size. Any tier-2
> divergence carrying a pending row at rung 15 is the launch gate's
> business, not rung 15's to wave through.

## Token ceiling: 6M

As a **starting** figure, per the spec's Band and ceiling section
(`:360`), and **re-derived from the first Go translation task's actual
report**, meaning Task 3, which births `internal/engine/musicbrainz`,
before Task 4 dispatches.

The basis. 3a spent **≈2.84M against its 6M starting ceiling** with zero
attended interactions across nine tasks plus a pre-task
(`docs/HISTORY.md`, 3a entry, Budgets), which is ≈316K per task.
**`docs/STATUS.md:66` records that spend as ≈2.26M and the two ledgers
disagree** (fold, m1). This plan takes HISTORY's figure because HISTORY
itemizes it to 2,836K across nine named dispatches where STATUS carries a
bare number, and STATUS is read in full at every session start. The
correction is conductor-owed and is not a plan edit; it is listed with
the other three provisional items. 2d, the
closest derivation-shaped precedent, spent ≈1.6M against 2.5M. 3b is nine
tasks in the same shape: two derivation tasks in 2d's cost shape (Tasks 1
and 2, Python entry points plus committed ledgers), five Go tasks on a
package materially smaller than `matcher`'s twenty-nine files, one
harness-grading task, and a close. At 3a's own measured per-task rate that
is ≈2.85M, and 6M carries roughly a two-times margin for the two tasks
this plan flags as the hard ones (Task 4's alias precedence and Task 1's
first ledger derivation). **6M is therefore held rather than lowered**:
3a's underspend is one data point, not a trend, and the ceiling is the
stop rule's yardstick rather than a forecast.

**Flag at 80% (4.8M):** finish the task in flight, write STATUS, and one
combined question goes to Geoff before continuing.

**Checkpoint: STATUS after every task.** That is 2b's, 2c's, 2d's, and
3a's cadence; every task reports its own spend and its own gate evidence,
and the conductor sizes the next task against it.

**Phase 1's combined ceiling: 12M, and it stays a reservation rather than
a live sum.** It is 3b's 6M plus a **6M reservation for the second code
slot** (lane A, rung 5, or lane T, rung 8, one at a time). `docs/STATUS.md`'s
lane registry records lane-2 as BLOCKED until Phase 0b and 0c land, so
**no second sub-ceiling exists to sum today**, and 12M is not a figure any
spend is currently measured against. The conductor replaces the
reservation with that lane's own assigned sub-ceiling at its launch; until
then 3b's 6M is the only live ceiling. This restates 3a's figure
unchanged rather than re-deriving it.

## Global Constraints

- **Main tree.** The spec's tree column for 3b is "Main tree: writes
  `corpus/` and `internal/harness/oracle`" (spec `:329`), and the standing
  policy is that corpus and harness work never moves to a worktree
  (parallel-approach spec, lane-launch checklist item 5). The lane registry
  entry is written before the first dispatch, and **the live-executor check
  is that registry, never `pgrep -f Projects/dubplate`**, which matches
  every lane once a worktree exists.
- **Package territory, stated as a wall.** This pass writes
  **`internal/engine/musicbrainz`** (new, born here),
  **`internal/harness/analysis`** (the `defect_free` block and the
  `AlbumInfo` ledger schema), **`internal/harness/oracle`** (the fourth and
  fifth Python entry points and their Go drivers),
  **`internal/harness/conform`** (the translation-tier grading, the
  profile overlay and its drift guard, the pattern-table name guard, and
  the widened clock check), **`internal/harness/corpus_test.go`** (wiring the
  new gate-resident checkers to the real corpus, exactly as
  `TestFileTagLedgerIsWellFormed` at `internal/harness/corpus_test.go:270`
  already does), **`corpus/analysis/album_info.json`** (new) and
  **`corpus/analysis/adjudications.json`** (the `defect_free` field only),
  and **`internal/engine/matcher/tunables.go`** (two named additions:
  `VAName` and `SDPatternsNoBracketDiscount`), plus conductor-owned `docs/`
  ledgers. It writes **nothing** under `corpus/cases/`,
  `internal/harness/{cassette,stub,synth,mbrequest}`, `internal/config`,
  `cmd/`, `tools/`, or `_salvage/`. A task that believes it needs one of
  those stops and reports.
- **`corpus/cases/*` is frozen input, and read-only is a real permission.**
  Reading every committed `case.json` and every `responses/*.json` from Go
  and from Python is exactly what the ledger derivation and the
  translation grading are for. **No agent reads a `responses/*.json` body
  into context**. The corpus discipline is unchanged, and every figure
  about response content in this plan was produced mechanically.
- **`internal/config` is not edited, and neither is `keys.go`.** Ruling C's
  registration half is hard-inbound on rung 1b (Phase 0b, unlanded;
  `internal/config/keys.go:19-25` still carries `ownedWall` as a slice
  literal). The four-knob registration stays the named follow-on task for
  the first rung 3 pass that runs after 1b closes (3a HISTORY guard 2).
  The `.gitleaksignore` rider applies even though this pass does not touch
  the file: its single allowed finding is pinned by line number at
  `internal/config/keys_test.go:generic-api-key:66`, and an insertion above
  line 66 reds the gate for every lane at once. An append is
  conductor-owned, must sit strictly below line 66, and is never a rule
  disable.
- **The engine never imports the harness.** `internal/engine/musicbrainz`
  inherits the invariant free: `TestEngineNeverImportsHarness`
  (`internal/harness/conform/arch_test.go:138-144`) already walks all of
  `internal/engine`. Task 8 widens `TestMatcherNeverReadsTheClock`
  (`arch_test.go:146-152`) from `internal/engine/matcher` to
  `internal/engine` for the same reason release-event date selection is
  where a `time.Now` would creep in.
- **`internal/harness/analysis` never imports `internal/engine`.** It is
  the ledger-and-schema tier that `internal/harness/corpus_test.go` wires
  to the corpus, and it imports only `internal/harness/cassette` today
  (`internal/harness/analysis/file_tags.go:10`). Every comparison against
  an engine value lives in `internal/harness/conform`, whose own charter
  (`internal/harness/conform/doc.go:1-23`) is exactly the
  harness-drives-engine direction. A task that reaches for a `matcher`
  import inside `analysis` stops and reports.
- **Assertion discipline, at the spec's own bar.** "Every checker in the
  runner ships a fixture that makes it fire, and the task that lands it
  reports the mutation that proved it can fire" (spec `:318`). The rule
  reaches this revision's additions by name: the `defect_free` block's
  checker is called out explicitly at spec `:320`. No silent-success fakes,
  no vacuous assertions, no self-derived expectations. **Hand-authored
  fixtures live in the owning package's own `testdata`, never a corpus
  case**, on the proximity precedent.
- **The spec's tree sketch names `translate.go` and `pseudo.go`; this pass
  ships eight files.** The rung 3 spec's package tree (`:177-182`) lists
  `translate.go  the translation layer (normalize, media, aliases,
  events)` as one line, which is a sketch of the package's job rather than
  a file-count mandate. 3a shipped twenty-nine files against the same
  tree's single `album_*.go` line for the same reason. The evolvability
  ruling asks that a rule change touch one file and its test (spec
  `:193`), and the layer's five bullets are five independently changeable
  rules, so they get `release.go`, `relations.go`, `alias.go`,
  `credit.go`, `tracks.go`, `events.go`, `translate.go`, and `pseudo.go`.
  `translate.go` is the composition, exactly as `score.go` is on the
  matcher side.
- **`make derive` runs in this pass, and only where a task names it.**
  Tasks 1 and 2 run oracle-tagged tests (`make derive` is `go test
  -tags oracle -timeout 90m ./internal/harness/... ./cmd/harness/...`,
  `Makefile:38-39`). Every other task is barred. **The byte-identity
  invariant is absolute**: after this pass, `make derive` must re-derive
  every previously committed record byte-identically:
  `corpus/analysis/components.json`, `ties.json`, `file_tags.json`, and
  every `corpus/cases/*/case.json`. This is 2d's own pre-ruling (spec
  `:374`) and it is the reason the `SD_PATTERNS` edit Task 2's derivation
  needs is applied by monkeypatching `beets.autotag.distance.SD_PATTERNS`
  inside the new script's own process, never by editing
  `derive_oracle.py`, `derive_component.py`, or the venv.
- **No live network. Anywhere.** The gate never reaches the network and
  neither does anything else here. No `curl`, no `WebFetch`, no HTTP
  client, no `/ws/2` request. The derivation entry points reach
  `stub.Server` on an ephemeral loopback address, exactly as
  `DeriveTies` already does (`internal/harness/oracle/ties.go:63-66`).
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck,
  ruff (`check` and `format --check`) over every tracked `*.py` (this pass
  adds two, and they are linted the moment they are tracked with no gate
  edit, `scripts/check.sh:85-92`), vale over Go comment prose, gitleaks
  over the tracked worktree, `docker compose config`, `systemd-analyze
  verify`. **Never commit red. Specific files only.** Commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`. **Tier-0 replay green
  over all 72 rows is part of every task's gate expectation**, because it
  runs inside `go test ./...`.
- **`internal/harness/analysis`'s loader sets `DisallowUnknownFields`**
  (`internal/harness/analysis/ledger.go:20`). The mechanism is
  **one-directional** (fold, m4): a `defect_free` key in the committed
  JSON with no Go field reds the gate immediately, while a Go field with
  no JSON key stays at its zero value and the gate stays green. The rule
  it enforces is unchanged and matters more for it: **the Go type and the
  committed JSON move in the same commit**, every time, because only one
  of the two orders fails loudly.
- **`go-conventions` before any Go; `python-conventions` before either
  Python entry point.** The bar is the standing exemplary-stdlib-idiom bar
  (`idiomatic-beautiful-go`). Concrete types over interfaces, no functional
  options, `slices.SortStableFunc` with `cmp.Compare` never
  `sort.SliceStable`, `for range n` where the index is unread,
  `errors.Join` for multi-field checkers, `any` not `interface{}`, no
  `init()`, no `context.Context` in a struct, errors lowercase and
  context-first with `%w` only where a caller branches. Two places need a
  written reason in the doc comment because they look like anti-patterns
  and are not: the `*T` ledger fields (the JSON null/empty distinction the
  skill's `[]T{}` rule carves out) and `GroupRelations` returning a map
  whose key order is unobservable.
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects. No YAML, JSON-path, or MusicBrainz client package is adopted;
  a task that wants one stops and reports.
- **Universality.** No estate hostname, path, email, or catalog value
  anywhere outside `internal/config`. Hand-authored fixtures are invented
  strings, not estate data.
- **Determinism.** No map-iteration order reaches an output. Every ordering
  the port reproduces is beets' own, and the derivation entry points set
  `PYTHONHASHSEED=0` through `runPython`
  (`internal/harness/oracle/ties.go:129`).
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each before Task 1
  dispatches.

## Contested surfaces, and the one row the conductor owes

Lane-launch checklist item 2 requires a written owner for every surface a
lane touches (parallel-approach spec, "Lane-launch checklist"). 3b touches
eight. **Five already carry a row in the parallel spec's contested-surface
table; three do not, and all three are conductor-owed edits** (fold,
M-j), in the same shape as 3a's B4 ruling on `tools/unidecode-table/`.

| Surface this pass writes | Row today | Status |
| --- | --- | --- |
| `internal/harness/conform` | "born by rung 3 pass 3a; rung 3's lane thereafter" | written |
| `internal/harness/oracle` | "lane E1 through 2d's close (landed); rung 3 pass 3b adds the derived-AlbumInfo entry point" | written, and names 3b explicitly |
| `corpus/analysis/*` and `corpus/cases/*` | "lane E1 authors; the adjudication stream commits dispositions; rung 3 adds schema only through re-derivation (3b's `defect_free`)" | written, and names 3b's `defect_free` explicitly |
| `internal/harness/{cassette,stub}` | "lane E1 as shipped; rung 3 pass 3c makes the ruled `Case.Extra` schema-and-stub change" | written, and 3b writes neither |
| `docs/` ledgers | conductor-owned by standing rule | written |
| **`internal/harness/analysis`** | the same row's Owner column carves out exactly one rung-3 exception and it is **3c's** `Case.Extra` change | **conductor-owed**: 3b's `analysis/album_info.go` and its `AdjudicationEntry` change are not covered |
| **`internal/harness/corpus_test.go`** | **none**; it sits in the parent `internal/harness` package, outside `{cassette,stub,analysis}`, and a Consumers column is not an ownership grant | **conductor-owed**: 3b writes it twice (Tasks 1 and 2) |
| **`internal/engine/{matcher,musicbrainz}`** | **none in the contested-surface table**, though the phase-1 lane table already awards lane E2 "`internal/engine/matcher`, `musicbrainz`, `shadow`" (`docs/specs/2026-09-03-parallel-build-approach.md:80`) | **conductor-owed restatement, not a launch stop** |

**The three edits this plan asks the conductor to make** (**Conductor ruling, 2026-09-04: taken as recommended**; all three
landed in the parallel spec's table before launch):

| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/engine/{matcher,musicbrainz}` | rung 3 (the E2 lane): `matcher` born by 3a, `musicbrainz` born by 3b; no other lane writes either | rung 4's shadow, rung 9's pipeline, rung 15's re-run |
| `internal/harness/corpus_test.go` | rung 3's lane while rung 3 runs; lane E1 otherwise, as the corpus's own wiring | every gate-resident ledger check |
| `internal/harness/{cassette,stub,analysis}` (amend the Owner column) | add "and rung 3 pass 3b adds the `AlbumInfo` ledger and the `defect_free` block to `analysis`" beside the existing 3c carve-out | rungs 8, 9, 15 |

Why they are owed rather than assumed. The spec's own contested-surface
proposal (spec `:335-343`) enumerates four rows and names only harness
and corpus paths, because at authoring time `internal/engine` did not
exist. It does now (`internal/engine/matcher`, twenty-nine non-test
files). **None of the three is a launch blocker**, and this is the fold's
one downgrade of a reviewer's own severity: the lane table at `:80`
already awards lane E2 the engine territory, the checklist's rule is that
an unwritten surface "blocks the second lane", and 3b is the only live
code lane. Task 0 reports all three as missing and the conductor writes
them; a missing row does not stop the launch.

## The translation contracts, stated once

Every task below cites this section instead of restating it. Sources: the
pinned venv at `tools/venv/.venv/lib/python3.13/site-packages`, abbreviated
`$SP`, read at drafting time by the planning inventory. Verify any line
citation you carry into code.

**T1. There is exactly one album-side translation function, and the
pseudo merge sits inside it.** `MusicBrainzPlugin.album_info`
(`$SP/beetsplug/musicbrainz.py:631-695`) is the whole layer. It is reached
only through `album_for_id` (`:765-800`), and **search candidates reach it
too**: `SearchApiMetadataSourcePlugin.candidates` maps each hit's id
through `albums_for_ids` (`$SP/beets/metadata_plugins.py:442-446`), a
generator over `album_for_id` (`:250-259`). So every `AlbumInfo` the
matcher ever scores has been through `album_info` **and** through the
pseudo-release check. The matcher itself never touches raw WS data:
`$SP/beets/autotag/distance.py` reads only `AlbumInfo`/`TrackInfo`
attributes.

**T2. Key normalization is two rules, and only one of them survives into
Go.** `MusicBrainzAPI._normalize_data`'s dict handler
(`$SP/beetsplug/_utils/musicbrainz.py:758-775`) does two things: it
rewrites every key's hyphens to underscores recursively, indiscriminately,
including data keys (`{"attribute-ids": {"guest-vocal": "x"}}` becomes
`{"attribute_ids": {"guest_vocal": "x"}}`), and it regroups a generic
`relations` array by `target-type` into `<target_type>_relations` keys,
`sorted` then `groupby` on that key, dropping `target-type` from each
entry, then recursing into the group. **The hyphen half is discharged by
Go struct tags** naming the wire key directly, and needs no runtime pass.
**The regrouping half is structural and lands as one named function with
its own test.** Three sub-rules a port must carry: grouping keys appear in
ascending target-type order and entries keep response order within a group
(Python's sort is stable); a `relations` key destructively collides with a
pre-existing `<type>_relations` key, later-wins in both directions; and a
relation with no `target-type` **raises** in beets rather than being
skipped (`AttributeError` at `_utils/musicbrainz.py:769`, or `TypeError`
from `sorted` at `:763` with two or more mixed types).

**T3. `mediums` counts every medium; track extraction counts only the
surviving ones.** `album_info:659` reads `len(release["media"])`, the
**unfiltered** list, while `:638-640` filters against
`match.ignored_media` for track extraction. Verified on
`corpus/cases/media-baseline`: `ignored_media=[]` gives `(mediums=2,
tracks=6, media='CD')`; `ignored_media=['CD']` gives `(mediums=2,
tracks=0, media='Media')`. `ignored_media` membership is exact and
case-sensitive (`'cd'` does not match `'CD'`).

**T4. `media` can become the literal string `"Media"`, and it is truthy.**
`:654-658` is `medias.pop() if len(medias) == 1 else "Media"` over `{t.media
for t in track_infos}`. Zero surviving tracks makes the set empty and
**two or more distinct** per-track formats make it larger than one, so
both land on the four-character sentinel. **Ten of the 303 committed
release lookups produce `media == "Media"`, five of those with zero
tracks.** The sentinel is truthy, so the media penalty at
`distance.py:455` **fires** on it and scores `"Media"` as if it were a
media name.

**A `None` format is one of the distinct values, not a third route to the
sentinel** (fold, B5). A single medium whose `format` is null gives the
set `{None}`, whose length is 1, so `pop()` returns `None` and
`AlbumInfo.media` is `None`, which is falsy and declines the media
penalty. Measured against pinned beets: `[None]` gives `None`, `['CD']`
gives `'CD'`, `['CD','Vinyl']` and `[None,'CD']` both give `'Media'`, and
`[]` gives `'Media'`. So `[None, "CD"]` **is** a sentinel route and is
worth its own fixture, and the single-null-format case is a fixture for
the **non**-sentinel outcome.

**T5. Per-medium track extraction has a fixed assembly order and two
filters.** `get_tracks_from_medium` (`$SP/beetsplug/musicbrainz.py:555-600`).
Assembly, `:556-563`: `pregap` first, **unconditionally, with no config
gate**; then `medium["tracks"]`; then `medium["data_tracks"]` **only when
`ignore_data_tracks` is false**. All three read through `.get`, so a medium
missing any contributes nothing rather than raising. Filters, `:571-580`,
applied to the assembled list: drop any track whose
`recording["title"] == "[data track]"`, **regardless of
`ignore_data_tracks`**; and drop any track with `recording["video"]` true
when `ignore_video_tracks` is set. The measured truth table over a medium
carrying a pregap, one audio, one video, one data, and one
`[data track]`-titled track:

| `ignore_data_tracks` | `ignore_video_tracks` | titles yielded |
|---|---|---|
| true | true | `['Pregap', 'A']` |
| false | true | `['Pregap', 'A', 'D']` |
| true | false | `['Pregap', 'A', 'V']` |
| false | false | `['Pregap', 'A', 'V', 'D']` |

**T6. Track-level overrides are three rules on a shallow copy of the
recording** (`:582-590`): `recording["length"] = track["length"] or
recording["length"]`, so a track length of `0` or `None` falls back;
`recording["artist_credit"] = track["artist_credit"] or
recording["artist_credit"]`, so an empty list falls back; and
`recording["title"] = track["title"]` **only if** `track["title"]` is
truthy **and** `_preferred_alias(recording["aliases"])` returns nothing.
The last rule means `_preferred_alias` runs **twice per track**, at `:589`
on the recording's aliases and again inside `track_info` at `:443` via
`_key_with_preferred_alias`. Measured precedence, with a Japanese primary
alias on the recording:

| `import.languages` | recording aliases | `track["title"]` | resulting `TrackInfo.title` |
|---|---|---|---|
| `['ja']` | ja alias present | `'TrackTitle'` | `'AliasTitle'` |
| `[]` | ja alias present | `'TrackTitle'` | `'TrackTitle'` |
| `['ja']` | none | `'TrackTitle'` | `'TrackTitle'` |
| `['ja']` | none | `''` | `'RecordingTitle'` |

Per-medium fields written at `:565-570` and `:593-598`: `medium` (from
`medium["position"]`), `medium_total` (`track_count`), `disctitle`
(`title`), `media` (`format`, `str | None`), `medium_index`
(`int(track["position"])`, an explicit cast), `release_track_id`
(`track["id"]`), `track_alt` (`track["number"]`). `index` is assigned
afterwards, 1-based **across all valid media** (`:645-646`).

**That cast is not decoration, and the Go port must carry it** (fold,
B3). beets writes `int(track["position"])` because the wire value is not
always an integer. Counted mechanically over
`corpus/cases/*/responses/*.json`, **297 track positions across 76 case
directories are JSON strings**, including three ledger subjects,
`media-baseline` (6), `media-uneven-split` (6) and `va-exact` (3). The
cause is the synth tool: `internal/harness/synth/release.go:128` declares
`Position string` and `:200-203` writes `strconv.Itoa(tr.Position)`, so
every synthetic case's response carries a string where a real capture
carries an integer. A Go `Position int` field fails the whole decode
(`json: cannot unmarshal string into Go struct field .position of type
int`), which would break three of the nineteen ledger subjects on the
first run and 76 case directories for 3c and 3d.

A full wire-type census over all 399 committed release documents says
`position` is the **only** field that needs the two-shape treatment.
`medium.position` is `int` 576 and `null` 130, and a JSON null decoded
into an `int` is a no-op leaving 0, which matches beets' falsy `None` at
`distance.py:421`. `medium.title` is `str` 576 / `null` 130,
`medium.format` `str` 693 / `null` 13, `release.date` `str` 379 / `null`
20, `release.status` `str` 394 / `null` 5, `labelinfo.catalog-number`
`str` 311 / `null` 32, `labelinfo.label` `dict` 340 / `null` 3, and
`track.title`, `recording.title`, `track.id` and `recording.id` are
always strings. Every one of those decodes into the declared shape
without a custom unmarshaler.

**T7. `_preferred_alias` is an outer loop over configured locales and an
inner loop over aliases in response order** (`:128-155`). It accepts the
first alias where `al["locale"] == locale` **and** `al["primary"]` is
truthy **and** `(al["type"] or "").lower()` is not in the ignored set. It
reads `import.ignored_alias_types` at `:136-138`, lowercased into a set,
and `import.languages` at `:141`. Both ship empty
(`$SP/beets/config_default.yaml:32` and `:58`). Measured behaviors:

| `languages` | `ignored_alias_types` | selected alias |
|---|---|---|
| `[]` | `[]` | `None` |
| `['ja']` | `[]` | first ja primary |
| `['ja']` | `['artist name']` | falls through to the **next ja primary alias**, not to `None` |
| `['ja']` | `['ARTIST NAME']` | same (the ignored set is lowercased) |
| `['JA']` | `[]` | matches only a `JA` locale, **case-sensitive** |
| `['en','ja']` | `[]` | the en alias (locale order wins over response order) |
| `['ja','en']` | `[]` | the ja alias |
| `['ja']` | `['artist name','search hint']` | `None` |

The explicit `languages` parameter overrides config but **no call site in
the plugin ever passes it**, so that branch is out of the port's scope.
`_key_with_preferred_alias` (`:158-162`) wraps it for three title fields:
release title (`:648`), release-group title (`:487-489`), recording title
(`:443`).

**T8. Artist credits flatten with a trailing joinphrase, and the credit
name differs from the entity name.** The function is
`MusicBrainzPlugin._parse_artist_credits` (`:301-357`); **no symbol named
`_flatten_artist_credit` exists at the pinned version**, and the port map's
name for it is wrong. Per credit element (`:327-346`):
`artists_ids.append(el["artist"]["id"])`; `alias =
_preferred_alias(el["artist"].get("aliases", []))`; `artist_object = alias
or el["artist"]`, so an alias supplies both `name` and `sort_name`;
`credit_artist_object = alias if (alias and self.aliases_as_credits) else
el`, so the credit string uses the **credited** name unless
`aliases_as_credits` is on. For each name, `parts.extend([name,
joinphrase])`. **The joinphrase is appended after the last artist too**:
`"".join(artist_parts)` at `:349` includes the final element's joinphrase,
so joins `[' & ', '']` give `'A & B'`, a single credit with join
`' feat. '` gives `'A feat. '` trailing space and all, and joins
`[' & ', ' & ']` give `'A & B & '`. `artist_id = artists_ids[0]` at `:350`
**raises `IndexError` on an empty credit list**, so `artist` is never
`None` and in practice never `""`.

**Three call sites reach that raise, not one** (fold, M-d): the release
credit (`:650`), the recording credit inside `track_info` (`:457`), and
the track-level override that swaps a track credit in for a recording
credit (`:586-588`). A boundary check that validates only the release
credit leaves the Go port to panic on the other two, so `Validate` walks
the track level too (R10).

**T9. The VA override replaces six artist fields and is not
verdict-moving.** `:673-681` replaces `artist`, `artist_sort`, `artists`,
`artists_sort`, `artist_credit` and `artists_credit` with
`config["va_name"]` whenever `artist_id == VARIOUS_ARTISTS_ID` (`:43`).
The album artist penalty is skipped entirely when `album_info.va`
(`distance.py:447-448`) and the track artist penalty reads
`track_info.artist`, which the override never touches. It is
**ledger-moving**, so ruling E's ledger records it and `Tunables` must
carry `VAName` to reproduce it.

**T10. Release-event selection is exact, case-sensitive list membership,
and it returns the configured string.** `_preferred_release_event`
(`:165-177`) reads `match.preferred.countries` at `:170`, loops over
configured countries outermost and `release.get("release_events", [])`
innermost, and returns `(country, event["date"])` for the first configured
country with any event whose `area["iso_3166_1_codes"]` contains it;
otherwise `(release.get("country"), release.get("date"))`. Measured
against events for `GB` and `US`: `['GB']` returns `('GB',
'1999-01-02')`; `['gb']` returns the fallback `('XW', '1980-05-06')`;
`['US','GB']` returns `('US', '2001-03-04')` because configured order
wins. **The returned country is the configured string, not the area's own
code.** `event["area"]` may be `None` and is skipped by the walrus at
`:174`. **This is the port hazard the tree does not yet name**: the same
`Preferred.Countries` field is compiled as an anchored case-insensitive
regex list by the country distance component
(`internal/engine/matcher/album_country.go:31-40` →
`CompilePatternOption`, `penalty.go:368-378`) and matched by plain
membership here. One config list, two matching regimes, in two packages.

**T11. Date parsing splits on `-` and swallows what it cannot parse.**
`_get_date` (`:180-200`) maps each part through an int parse returning
`None` on `ValueError`. Measured: `''` gives `(None,None,None)`; `'1999'`
gives `(1999,None,None)`; `'1999-05'` gives `(1999,5,None)`; `'1999-XX-06'`
gives `(1999,None,6)`, a `None` month with a real day; `'199x'` gives all
`None`; `'1999-05-06-07'` gives `(1999,5,6)`, extra parts discarded
silently. The fallback at `:685-689` tests **truthiness** of
`release_date`, not `is not None`, so an empty-string date falls back to
`(original_year, original_month, original_day)` just as a missing one
does.

**T12. The pseudo merge is a deepcopy of the pseudo plus exactly fifteen
keys from the actual.** Trigger, `:785-798`: `original_release["status"]
== "Pseudo-Release"` (exact string) plus the first `release_relations`
entry with `type == "transl-tracklisting"` **and** `direction ==
"backward"`. That relation key exists only because T2's regrouping ran.
`_merge_pseudo_and_actual_album` (`:203-240`): `merged = pseudo.copy()`,
which is `AttrDict.copy` and therefore a **deepcopy**
(`$SP/beets/autotag/hooks.py:93-95`), tracks included; then
`merged.update(from_actual)` with `media`, `mediums`, `country`,
`catalognum`, `year`, `month`, `day`, `original_year`, `original_month`,
`original_day`, `label`, `barcode`, `asin`, `style`, `genres` (`:221-238`).
**Seven of the fifteen are in `matcher.AlbumInfo`**: `media`, `mediums`,
`country`, `catalognum`, `year`, `original_year`, `label`. `Album`,
`Artist`, `VA`, `AlbumDisambig`, `AlbumID`, `DataSource`, and `Tracks`
stay the pseudo's. Corpus evidence over the four committed pseudo cases:
the merge moved `asin, barcode, catalognum, country, label, media` on
`non-latin-pseudo-release`; `barcode, catalognum, country, label, media`
on `real-nonlatin-girls-generation`; `catalognum, country, label` on
`real-nonlatin-zemfira` and its stock twin. **`year`, `month`, `day`,
`original_*`, `mediums`, `style` and `genres` were already equal on every
committed case**, so the merge's year path has no corpus signal at all and
its Go test must be hand-authored.

**T13. No Unicode normalization happens anywhere in the translation
layer.** No `unicodedata.normalize` call exists in either
`beetsplug/musicbrainz.py` or `beetsplug/_utils/musicbrainz.py`. This
extends 3a's contract 2 from the distance path to translation: **the port
must not helpfully normalize.** The only case folding the layer does is
`_preferred_alias` lowercasing the ignored alias types (`:137`),
`_parse_release_group` lowercasing `primary_type` and every
`secondary_type` (`:476`, `:484`), and `_parse_label_infos` comparing the
label name against the exact literal `"[no label]"` (`:504`,
case-sensitive, no strip). `_preferred_release_event` does no case folding
at all. Search-query lowercasing (`_utils/musicbrainz.py:656`,
`beetsplug/musicbrainz.py:752`) is 3c's territory.

**T14. Config reads inside the translation layer, with their shipped
defaults.** Five are verdict-moving and already in `Tunables`:
`match.ignored_media` `[]` (`config_default.yaml:200`, read `:255-256`),
`match.ignore_data_tracks` `yes` (`:201`, read `:259-260`),
`match.ignore_video_tracks` `yes` (`:202`, read `:263-264`),
`match.preferred.countries` `[]` (`:195`, read `:170`),
`import.languages` `[]` (`:32`, read `:141`), and
`import.ignored_alias_types` `[]` (`:58`, read `:136-138`). Two are
ledger-moving and absent from `Tunables` today: `va_name` `"Various
Artists"` (`config_default.yaml:106`, read `:675`) and
`musicbrainz.aliases_as_credits` `false` (`beetsplug/musicbrainz.py:287`,
read `:267-268`). **Ruling D's blanket rule holds: every default equals
pinned beets' shipped value at launch.**

**T15. Five of the plugin's config reads are `@cached_property` on the
instance** (`:246-268`): `ignored_media`, `ignore_data_tracks`,
`ignore_video_tracks`, `aliases_as_credits`, `genres_field`. They are read
once per instance per process. **The derivation harness must construct a
fresh plugin after any config change**; a probe that mutates config
between two calls on one instance silently reads the stale value.
`_preferred_alias` and `_preferred_release_event` read config fresh on
every call, so the two families behave differently under config mutation.

**T16. The browse fallback's exact trigger, and why 3b does not port it.**
`_ensure_complete_recordings` (`$SP/beetsplug/musicbrainz.py:602-629`)
fires when `sum(len(m.get("tracks", [])) for m in release["media"]) >
BROWSE_MAXTRACKS`, where `BROWSE_MAXTRACKS = 500` (`:67`). The sum counts
only `medium["tracks"]`, not pregap and not data tracks, across **all**
media, unfiltered by `ignored_media`. It then issues paginated
`browse_recordings` requests in chunks of `BROWSE_CHUNKSIZE = 100`. The
largest committed release is `real-boxset-rachmaninov-complete-stock` at
470 tracks, thirty short of the line, and no committed cassette carries a
browse request, so the stub would answer one with a recorded 404 miss.
**`TranslateAlbum` issues no requests, so the fallback is 3c's client
work**, and 3b's obligation is a guard rather than a port (Resolution R9).

## Resolutions this plan takes rather than discovers

The planning inventory returned twenty-one deduplicated open questions
(`docs/research/2026-09-04-rung-3b-planning-inventory.md:1932-2090`).
Each is resolved here, with its evidence and its rejected alternative. A
task that believes a resolution is wrong reports a finding and implements
the record anyway.

### R1: the absent-source-field question (3a's first carried item)

**Resolved: `matcher.AlbumInfo` and `matcher.TrackInfo` keep plain-string
fields, the translation maps a beets `None` to the type's zero value, and
the proof that the divergence is unreachable ships as a gate-resident
check rather than an assumption. `addAlbumAlbum` is not changed and
neither are the `info.go` types.**

3a carried this: `addAlbumAlbum` passes `&likelies.Album`/`&info.Album`,
both always non-nil, so a beets `album_info.album` of `None` scores `0.0`
where beets scores `string_dist("", None) = 1.0` at weight 3.0
(`docs/HISTORY.md`, 3a "Carried"). The evidence, from pinned beets:

- **Only three `add_string` call sites on the whole distance path are
  unguarded by a truthiness test on the info side**, so only three can
  ever see a `None`: `dist.add_string("album", likelies["album"],
  album_info.album)` (`$SP/beets/autotag/distance.py:451`, unguarded);
  `dist.add_string("artist", ...)` (`:448`, guarded only by `not
  album_info.va`); and `dist.add_string("track_title", item.title,
  track_info.title)` (`:402`, unguarded). Every other component is guarded
  on both sides (`:495-518`), and for those `None`, `""` and `0` are
  indistinguishable.
- **All three are `str` by construction through the MusicBrainz
  translation path.** `album` is
  `_key_with_preferred_alias(release, key="title")` →
  `alias["name"] if alias else obj["title"]`
  (`$SP/beetsplug/musicbrainz.py:154-157`), and both are non-optional `str`
  in the client's own typing (`_utils/musicbrainz.py:523`, `:140`).
  `artist` is `"".join(artist_parts)` (`:349`), a `str` by construction,
  overwritten with `va_name` on a VA release (`:676`). `track_info.title`
  is the same shape (`:443`), and the track-level override only ever
  replaces it with a truthy `track["title"]` (`:585-586`).
- **The corpus agrees.** Across 294 committed candidates the largest
  `album` penalty is 0.169, nowhere near the 3/61 a raw 1.0 would produce,
  so no committed case exercises the divergence today.

Modelling it would cost the whole vocabulary: `*string` on
`AlbumInfo.Album`, `.Artist`, and `TrackInfo.Title` ripples into all eight
`AddString` call sites, every component test, and 3d's assignment work, to
represent a state the only shipped source cannot produce.

**The proof is a check, not an assumption**, and it is Task 8's
nullability policy (Resolution R11's table). The residual is recorded
rather than absorbed: the proof is source-scoped, and a second metadata
source, or a beets bump that lets an alias carry a null name, reopens it.
The strict-nil half of the policy is the thing that would catch it.

Rejected: pointer fields on `AlbumInfo` (correct, but pays the full
vocabulary cost for an unreachable state and turns eight simple call sites
into nil-juggling); a presence bitset (an idiom nobody in this repo uses,
invisible to `go doc`, needing its own decoder); a sentinel string (it can
collide with real data and every reader must know it).

### R2: per-track `Media`

**Resolved: `matcher.TrackInfo` does NOT gain a `Media` field.
`translate.go` keeps an internal per-track record carrying the medium's
`format` alongside the matcher's eight fields, computes `AlbumInfo.Media`
from the set of those formats, then builds the final `[]TrackInfo`.**

`TrackInfo`'s own doc comment (`internal/engine/matcher/info.go:69-72`)
declares it "the eight fields the track-level distance path reads off
beets' `TrackInfo` ... and nothing else", and per-track media is never
read by the distance path. Adding a fifteenth field to the vocabulary to
carry a derivation input would falsify that sentence and would put a
translation-only concern in the scoring package. The derivation itself is
T4's, and the sentinel must be reproduced exactly.

### R3: the `"Media"` sentinel and the all-ignored-media release

**Resolved: in scope for 3b, and proved by hand-authored fixtures in
`internal/engine/musicbrainz/testdata`, not by the corpus releases that
happen to exercise it.** The sentinel is truthy and the media penalty
fires on it (T4), so it is verdict-moving.

**Two routes to the sentinel, not three** (fold, B5): zero surviving
tracks, and two or more distinct per-track formats, where a `None` counts
as one of the distinct values. A single-medium release whose `format` is
null yields beets' `None`, which this port maps to `""` under R1, and
which declines the media penalty. Task 5 therefore ships three fixtures:
one per sentinel route, plus one asserting the non-sentinel outcome on a
single null-format medium.

**Which of the corpus's ten instances actually grade this** (fold, M-h).
Recomputed under `DefaultTunables`, exactly ten committed release
documents land on the sentinel, five of them with zero surviving tracks,
and they live in six cases: `admit-zero-track-dropped`,
`real-nonlatin-hamasaki-ayumi` (four rows),
`real-noise-bowie-us1967`, `real-noise-dylan-bootleg` (two rows),
`real-reissue-doolittle-2008`, and `real-reissue-doolittle-2008-stock`.
Under R7's subject list as drafted, only `real-nonlatin-hamasaki-ayumi`
was a subject, so **four of the ten reached the ledger and all four were
the zero-track route**, leaving the mixed-format route graded by no
ledger row at all. R7 therefore gains `real-noise-dylan-bootleg` as a
nineteenth subject, **five rows**, so both routes are graded against
beets rather than only against a hand-authored fixture. The case carries
five release-lookup requests, counted off the tree, so the ledger's row
floor moves 45 to 50. **Conductor ruling, 2026-09-04: taken as recommended**: the
rejected alternative is to keep eighteen subjects and record the mixed-format
route as fixture-graded only, which costs nothing and grades less.

### R4: `preferred.countries` carrying two matching semantics

**Resolved: one `Tunables.Preferred.Countries` field, consumed two ways,
with the divergence named in a doc comment at the translation site.
`internal/engine/musicbrainz` compiles no regex at all: release-event
selection uses plain, case-sensitive `slices.Contains` (T10).** A second
derived view would put a "which one do I use" question at every call site
and would invite a reader to assume the two agree. The doc comment on the
release-event selector names `internal/engine/matcher/album_country.go:31-40`
as the other consumer and states plainly that it anchors and folds case
while this one does neither, citing `$SP/beetsplug/musicbrainz.py:174`.
Task 6 ships a fixture proving the difference: a `Preferred.Countries`
entry of `"gb"` selects nothing here while the country component would
match `"GB"`.

### R5: the preferred-regex validation question (3a's second carried item)

**Deferred to 3d, with the reason stated.** 3a carried that
`compileMediaPatterns` and `compileCountryPatterns` panic on a malformed
user regex (`docs/HISTORY.md`, 3a "Carried"). `internal/config` has no
validation hook and no typed variant today
(`internal/config/thresholds.go:13-59`), `keys.go` is frozen until rung 1b
lands, and **3b compiles no regex at all**, so this pass has nowhere to
put the check and nothing in it to check. 3d is the pass that already
touches the verdict and admission surface where a wrapped config error
would be raised, and it is the first rung 3 pass that could plausibly run
after rung 1b. The item stays on the carried list, re-addressed to 3d
alone rather than "3b/3d".

### R6: the ledger's keying and its field scope

**Resolved: one row per **release lookup**, keyed by `(case_id,
release_mbid)`, carrying the pre-merge `AlbumInfo` plus, on a
pseudo-release, the merged one. The graded field set is exactly the
fourteen `matcher.AlbumInfo` fields and the eight `matcher.TrackInfo`
fields.**

Keying: `case_id` alone is not unique. Four committed cases carry a
pseudo-release and therefore two release lookups whose translations
differ, and six other subjects carry search-fallback lookups. Measured at
drafting time, the nineteen named subjects carry **50 release-lookup
requests** between them (45 across the eighteen drafted subjects, plus
five for the nineteenth that M-h adds).

Field scope: rung 3 grades no applied field and computes none (spec
`:248`). `artist_sort`, `artist_credit`, `artists`, `artists_sort`,
`release_group_title`, `script`, `language`, `disctitle`, `medium_total`,
`track_alt`, `release_track_id` and the external-id fields are rung 8's
territory. **The alias hypothesis's verdict-facing evidence is already
inside the vocabulary**: under `languages=['ja']` the album `artist`
changes on 90 of 303 releases and track `artist` on 3,272 of 6,919 tracks,
and both are graded fields. A ledger recording all forty-one `AlbumInfo`
keys would grade rung 8's surface a rung early, at roughly triple the
committed size, with no verdict-moving field added.

### R7: the ledger's subjects and its row floor

**Resolved: a named subject list in `FileTagSubjects`' own shape
(`internal/harness/analysis/file_tags.go:39-52`), nineteen cases,
expanding to 50 release-lookup rows, with `AlbumInfoLedgerRowFloor`
measured by Task 1 and its measurement recorded in the constant's doc
comment.**

Every candidate of every case is 303 release lookups and 6,919 tracks,
several megabytes of committed JSON on a tree where gitleaks' full-worktree
scan is already the gate's dominant cost by a wide margin (3a HISTORY
guard 10: 99.6s on one run, 3m13s on the next, with no tree change). The
named subject list is the file_tags precedent, and its bias is stated
rather than implicit: each subject is there because a named
translation-layer branch runs through it.

| Subject | Lookups | Why it is a subject |
| --- | --- | --- |
| `clean-match` | 1 | the plain baseline; single medium, single credit, no alias, one release event taken through the fallback |
| `media-baseline` | 1 | T3's own measured case: `mediums=2` unfiltered against filtered extraction |
| `media-uneven-split` | 1 | multi-medium index assignment across media (T6's 1-based `index`) |
| `va-exact` | 1 | T9's VA override and the `VAName` read |
| `various-artists` | 1 | VA, plus the corpus's one release whose two `release-events` carry **different** dates (`2014-07-28` XE and `2014-08-29` DE) |
| `should-reject` | 6 | search fallback (six lookups) plus multiple `release-events` |
| `non-latin-pseudo-release` | 2 | pseudo merge; six fields moved (T12) |
| `real-nonlatin-zemfira` | 2 | pseudo merge; three fields moved |
| `real-nonlatin-zemfira-stock` | 2 | the stock-profile twin of the same merge |
| `real-nonlatin-girls-generation` | 6 | pseudo merge **and** an `oracle-wrong` alias probe **and** three video recordings |
| `real-nonlatin-anna-vissi-essential` | 1 | `oracle-wrong` alias probe (`adj-adv-alias-vissi-native-greek`) |
| `real-nonlatin-alla-pugacheva-vol1` | 1 | `oracle-correct` alias probe: entity name and release credit coincide |
| `real-nonlatin-x-japan-art-of-life` | 1 | `oracle-correct` alias probe, single-track release |
| `real-nonlatin-utada-first-love` | 6 | `bug-locale-alias-ignored`'s A half, under `estate` |
| `real-nonlatin-utada-first-love-locale` | 6 | the B half, and **the corpus's only `locale` profile case**: the one place `import.languages` is non-empty |
| `real-nonlatin-hamasaki-ayumi` | 5 | 138 `video: true` recordings; `ignore_video_tracks` empties a medium and the candidate is dropped |
| `real-genre-metallica-master-of-puppets` | 1 | one `release-event` whose date equals the top-level `date`, so the fallback supplies both country and date |
| `real-compilation-queen-greatest-hits-ii` | 1 | two `release-events`, both dated equal to the top-level `date`: the same fallback under a second stratum |
| `real-noise-dylan-bootleg` | 5 | the **mixed-format** route to the `"Media"` sentinel, two Hybrid SACD layers on one release (fold, M-h; conductor ruling 2026-09-04: taken) |

Total **50 rows** at drafting time. `real-boxset-rachmaninov-complete-stock`
(470 tracks) is **deliberately excluded**: its value is T16's browse-fallback
headroom, which Task 6 guards by assertion, and its 470 track rows would be
roughly a third of the ledger's whole size for one measurement a guard makes
better.

**Three why-strings are re-derived off the tree, and one claim they
implied is retired** (fold, M-l). `clean-match` was described as having
no release events and carries one, `('1984-04-24', ['GB'])`.
`real-genre-metallica-master-of-puppets` was described as carrying two or
more events whose dates differ from the top-level date and carries one,
equal to it. `real-compilation-queen-greatest-hits-ii` was described the
same way and carries two, both equal to it. More generally, with
`preferred.countries` empty every case takes `_preferred_release_event`'s
fallback (`$SP/beetsplug/musicbrainz.py:176`), so **no subject grades
release-event selection at all**: that rule is graded by Task 6's
hand-authored fixtures, and the subjects with events prove only that the
fallback is what runs. These strings land verbatim in committed Go doc
data, which is why a wrong one is worth a fold rather than a note.

### R8: `va_name` and `aliases_as_credits`

**Resolved: `Tunables` gains `VAName` (default `"Various Artists"`,
`config_default.yaml:106`) and does NOT gain `AliasesAsCredits`.**

`VAName` is written into `AlbumInfo.Artist` on a VA release (T9), which the
ledger grades field by field, so hard-coding the literal in `translate.go`
would put a beets default outside the one struct that carries beets'
defaults with citations. `AliasesAsCredits` selects only which object
supplies `artist_credit` (T8), and `artist_credit` is neither in
`matcher.AlbumInfo` nor in the ledger's graded set (R6), so the
translation never reads it. Measured confirmation that it does not leak
into `artist`: under `languages=['ja']` the album `artist` is the alias
name whether `aliases_as_credits` is true or false (inventory §(a).5's
table). Its exclusion is named in `Tunables`' doc comment, exactly as 3a
named `album_disambig_fields` and `singleton_disambig_fields`
(3a plan, Task 1's m10 fold).

**Neither is echoed by any committed case**, so the M1 defaults
cross-check (`conform.CheckCaseWeights` / `CheckCaseThresholds`,
`internal/harness/conform/defaults.go:21,43`) cannot cover `VAName`. Its
coverage is a unit test in `tunables_test.go` plus Task 4's
`applyVariousArtists` tests and Task 6's VA whole-release test, and the
gap is named in the pass report.

### R9: the browse fallback

**Resolved: not ported, declared with a named refusal, and guarded.**
`TranslateAlbum(rel, tun)` takes a decoded release and issues no requests;
pagination is client work, so the entry point belongs to 3c or later. The
refusal is recorded in `translate.go`'s package doc with T16's exact
trigger. **The guard is a gate-resident assertion in Task 6**: over every
committed case's release-lookup responses, `sum(len(medium.tracks))` must
stay at or below `BROWSE_MAXTRACKS = 500`, failing and naming the case,
the release, and the count when a future corpus case crosses the line.
Without it, a larger box set would make the Go translation silently
diverge from an oracle that fetched pages the cassette never carried.

Rejected: an injected fetcher parameter left nil in 3b (dead parameter,
untested branch, and it invites a nil-check that go-conventions forbids
for internal code); deferring the whole entry point to 3c (the spec scopes
the translation layer to 3b, and a client without a translation to feed is
not a shippable half).

### R10: the sites where beets raises

**Resolved, per site, with the Go behavior named.** `TranslateAlbum`
validates at the boundary and returns an error wrapping a single sentinel,
`ErrMalformedRelease`, so 3c can branch once on "this response is
unusable" without a hierarchy of error types.

Revision 1 called this "the five sites" and the enumeration was neither
complete nor correctly scoped (fold, M-d, M-e, m15). Two sites are added
as **boundary rules** because the Go port panics on them rather than
returning an error, and the table's closing row now says plainly that the
remaining direct indexes are a class rather than a list.

| beets site | beets behavior | Go behavior |
| --- | --- | --- |
| `artists_ids[0]` (`musicbrainz.py:350`), reached from the **release** credit (`:650`) | `IndexError` | boundary error: `%w` on `ErrMalformedRelease`, naming the release id and "artist-credit is empty" |
| `artists_ids[0]` reached from the **recording** credit (`:457`) or the track-level credit override (`:586-588`) | `IndexError` | boundary error naming the medium position and the track position (fold, M-d). Without this rule the Go port indexes element zero of an empty slice and panics inside `internal/engine/musicbrainz`. The corpus carries no instance, zero release-level empty credits and zero tracks whose track and recording credits are both empty, so the fixture is hand-authored on the proximity precedent and 3c's live client is the first thing that could meet it. |
| `rel["release"]["id"]` in `album_for_id`'s pseudo branch (`:794`) | `KeyError` | boundary error naming the relation index and the missing release target (fold, M-e). `Relation.Release` is a pointer and is nil on most relations: the corpus carries `url` 351, `release` 36 and `artist` 285 by target type, and an artist relation carries no `release` key at all. No committed backward `transl-tracklisting` relation lacks the object, so this too is a latent panic with a hand-authored fixture. |
| `target_type.replace` on an untyped relation (`_utils/musicbrainz.py:769`) | `AttributeError`, or `TypeError` from `sorted` | boundary error naming the relation index and the empty target type; `GroupRelations` may then assume well-formed input and keeps its simple map signature |
| `recording["aliases"]` (`musicbrainz.py:589`) | `KeyError` | **deliberate divergence**: a missing `aliases` key decodes to a nil slice and is treated as an empty alias list, which is what `_key_with_preferred_alias` itself does through `obj.get("aliases", [])` (`:161`). `RELEASE_INCLUDES` always requests aliases, so the raising branch is unreachable through a well-formed capture, and reproducing a Python crash as a Go crash buys nothing. Recorded in the function's doc comment. |
| `actual[k]` in the merge (`:220`) | `KeyError` | not reachable: `MergePseudoRelease` takes two total `*matcher.AlbumInfo` values, so there is no absent key. Stated in the doc comment. |
| every remaining direct index on a decoded response: `album_info`'s fifteen constructor indexes (`:662-671`), `release["media"]` and `m["format"]` in the `valid_media` filter (`:638-639`), `medium_data`'s four indexes on `position`, `track_count`, `title` and `format` (`:565-570`), and `track_info`'s `recording["length"]` and `recording["disambiguation"]` | `KeyError` | inert in Go: absent JSON keys decode to zero values, which is the same result every truthiness-guarded consumer sees (R1). **This row is a class, not a list** (fold, m15). One member is worth naming because its Go behavior differs rather than merely being inert: an absent `media` array makes beets raise while Go produces a zero-track candidate carrying the `"Media"` sentinel, harmless only because `_add_candidate` discards zero-track candidates (`$SP/beets/autotag/match.py:344-347`), which is 3d's code. Stated in `Release`'s doc comment. |

### R11: where the ledger's checker lives

**Resolved: a three-package split, matching 2d's file_tags precedent
exactly.**

- **`internal/harness/analysis/album_info.go`**: the ledger types, the
  subject list, the row floor, `LoadAlbumInfoLedger`, and
  `CheckAlbumInfoLedger` (schema, coverage, subject-set equality,
  case resolution). This package must not import `internal/engine`
  (Global Constraints), and it does not need to: the checks here are
  schema and coverage only, exactly as `CheckFileTagLedger`
  (`internal/harness/analysis/file_tags.go:100`) is.
- **`internal/harness/oracle/album_info.go` + `derive_album_info.py`**:
  the derivation. Oracle-tagged, venv-driven, `make derive` only.
- **`internal/harness/conform/translate.go`**: the field-by-field
  comparison of the Go translation against the committed ledger. This is
  the one package whose charter is harness-drives-engine
  (`internal/harness/conform/doc.go:1-23`), and where
  `CheckUnidecodeVersion` already reaches into `matcher`
  (`defaults.go:67-72`).

**It lands beside tier 0, as the translation tier**, and the spec's
three-tier numbering is unchanged: like tier 0 it grades a component of
the port against a derived table, rather than a case's verdict (tier 1) or
the bar (tier 2). The pass report names it "the translation tier" so 3d's
plan does not renumber anything.

The invariant argument, stated once so a reviewer can check it: the
dependency runs harness to engine and never the reverse (spec `:191`).
Putting the comparison in `analysis` would make the ledger tier import the
product; putting it in `internal/engine/musicbrainz` would invert the
dependency outright and trip
`TestEngineNeverImportsHarness`. `conform` is the only remaining home, and
it is the one the spec's M5 fold already chose for tier 0 for the same
reason (spec `:273`).

### R12: the `defect_free` block's population predicate

**Resolved: the predicate is derived from `oracle_bugs.json`, never a
hardcoded case-id list.** A case is **string-heuristic** when it is a
probe case of a hypothesis in a named set of two: `bug-ft-pattern` and
`bug-remix-parenthetical`. Both are defects **of `SD_PATTERNS`**, which is
the only surface a `PatternTable` swap can express. The other two
`oracle-wrong` hypotheses are excluded by construction and the constant's
doc comment says why: `bug-nonlatin-alias-selection` is a translation
defect with no `SD_PATTERNS` counterfactual, and
`bug-assignment-rejection-gap` is the class ruling A names as having no
beets counterfactual at all ("It cannot express the assignment class,
where the fix has no beets counterfactual", spec `:83`), so **the six
`adv-assign-*` rows owe no block**.

That predicate resolves, at drafting time, to exactly the ten cases the
spec names: six `adv-ft-*` under `bug-ft-pattern` and four `adv-remix-*`
under `bug-remix-parenthetical` (inventory §(b).8).

**The predicate is enforced in both directions** (fold, m12). Revision 1
required a block on the ten and forbade one on a non-`OracleWrong` row,
which left the six `adv-assign-*` rows unguarded: they are `oracle-wrong`
and they owe no block, so a fabricated block on one would pass every
check. `CheckDefectFree` therefore also fails a block sitting on any row
whose case is not a probe of a `StringHeuristicHypotheses` hypothesis,
naming the case and the hypothesis it belongs to instead.

### R13: `defect_free` versus `CorrectedVerdict`'s prose

**Resolved: the block carries structured figures **alongside** the prose;
the prose is not retired, and the gate cross-checks the two on their
categorical half only.**

`corrected_verdict` and `suspected` are owner-authored fields on an
owner-only row; retiring them would delete Geoff's own words to make a
checker's life easier. The block is derived data beside them.

The cross-check has two halves, split by what a gate can do honestly:

- **Standing, mechanical:** `CheckDefectFree` asserts that a block's
  `recommendation` and `action` appear verbatim in the row's
  `corrected_verdict`. All ten rows read "medium, skip", so the overlap is
  non-empty and the check fires on a transcription error in either
  direction. Parsing a float out of prose is not attempted, because a
  regex over an owner's sentence is a checker that breaks when the sentence
  is edited.
- **One-time, numeric, at population:** **Task 2** asserts each derived
  `distance` equals the figure the row's prose already carries, at the
  precision the prose states, and **reports each pair**. The ten expected
  figures are written into the plan (Task 2's table) so the assertion is
  against a value the task owns rather than one it computes. **The prose
  figure is the cross-check, never the stored value** (fold, M-a): the
  block stores the derivation's own unrounded `float64`, because the
  committed records themselves carry full precision
  (`corpus/cases/adv-remix-both-sides/case.json` reads
  `0.023999999999999997` against the prose's `0.024000000`).
- **Standing, numeric:** the oracle-tagged re-derive test. Re-deriving the
  ten figures through pinned beets and diffing against the committed
  blocks is what actually guards the numbers, on the same mechanism ruling
  E's ledger uses. A perturbation test gives it teeth.

### R14: which ledger file changes, and how disposition stays owner-only

**Resolved: `corpus/analysis/adjudications.json` changes, and only the
`defect_free` field on ten of its twenty-one rows.** `oracle_bugs.json`
is read, never written. `corpus/analysis/album_info.json` is a new file.
No `corpus/cases/*` file changes at all.

Disposition stays owner-only by three mechanisms, none of them new:
`checkAdjudicationEntry`'s sentinel rule
(`internal/harness/analysis/adjudication.go:140-142`) is untouched; the
`defect_free` block carries no disposition, no adjudicator, and no verdict
about who is right; and **Task 2's** acceptance criterion is that the ten
rows' other ten fields are byte-identical before and after. A `git diff`
over `adjudications.json` showing a change to any field but `defect_free`
is a task failure. That criterion is restated inside Task 2's own
behavior contract rather than left here, because a criterion parked in a
resolution nobody re-reads at execution time is a criterion nobody
applies (fold, M-f).

**The write-back's mechanism, named rather than left to the implementer**
(fold, M-g). `adjudications.json` is **not** in the derived set:
`corpus/README.md:2214-2218` lists `ties.json`, `components.json`, and
`file_tags.json` as the files rewritten rather than hand-edited, so
neither a `-update-ledgers` re-serialize nor a hand edit is safe by
default. Two measurements settle it. A Python round trip, `json.load`
then `json.dumps(indent=2, ensure_ascii=False) + "\n"`, reproduces the
committed file byte-identically today, 49,252 bytes both ways. A Go
re-serialize through `AdjudicationEntry` would not: the file's per-entry
key order ends `... reason, adjudicator, date, corrected_verdict`, while
the struct emits `corrected_verdict` immediately after `disposition`,
which reorders eighteen rows and trips this rule on the first run. So the
population is a targeted **Python** writer inside Task 2's derivation,
setting `entry["defect_free"]` on exactly the ten resolved ids so the new
key appends last, written with the measured spelling. Rejected: a Go
writer (reorders), and a `-update-ledgers` branch over the whole ledger
(re-serializes owner-authored rows nothing asked it to touch).

### R15: three more declared corpus-coverage gaps

**Resolved: 3b declares pregap, data tracks, and the preferred-country
release-event branch as M8-style corpus-coverage gaps**, in the same shape
3a declared `match.required`, `match.ignored`, and
`preferred.original_year` (spec `:234`; 3a HISTORY, "The three
corpus-unreachable surfaces"). Each is measured, not assumed:

- **Pregap: zero cases.** No response anywhere carries a non-null
  `pregap`, so the pregap-inclusion branch (`musicbrainz.py:557-558`) has
  no corpus fixture.
- **Data tracks: zero cases.** No response anywhere carries a non-empty
  `data-tracks` array, so both the `ignore_data_tracks` branch
  (`:562-563`) and the `"[data track]"` title filter (`:573`) have none.
- **The preferred-country branch: unreached, and reachable the moment any
  profile sets `match.preferred.countries`** (fold, B2). Revision 1
  claimed no release event anywhere in the corpus carries a non-null
  `area` with `iso_3166_1_codes`. That measurement was false. Counted
  over the 303 release-lookup responses: 298 top-level `release-events`
  objects, **290 of them coded, across 208 case directories**; counted
  recursively over all 351 responses so search-hit releases are included,
  413 events and 405 coded. `corpus/cases/clean-match`'s single event
  reads `('1984-04-24', ['GB'])`. The gap's one true cause is the other
  half of the sentence: **no committed profile sets any `match:` key**
  (`corpus/profiles/estate.yaml`, `stock.yaml`, `locale.yaml`, all read
  directly), so every case takes the fallback at `musicbrainz.py:176`. A
  fourth profile setting `match.preferred.countries: [GB]` would reach
  the branch on 208 case directories immediately, and `various-artists`
  is the concrete target: its two events carry **different** dates
  (`2014-07-28` XE and `2014-08-29` DE), so a preferred country would
  move both the country and the date on that case.

Each is graded by a hand-authored table-driven test in
`internal/engine/musicbrainz/testdata`, with the mutation that proved it
fires named in the task's report, and each is recorded in the pass report
as a **declared corpus-coverage gap, never a passing check**. A fourth
profile that would reach the preferred-country branch is a corpus-lane
decision and is not rung 3's to take, and the measurement above is what
that lane should read rather than re-derive.

### R16: who corrects `corpus/README.md`

**Resolved: 3b drafts the corrections and the conductor lands them in the
batch commit**, exactly as 3a handled its own two README cells. Four
corrections, each verified against the tree at drafting time:

1. `corpus/README.md:1096` says every adversarial case carries an
   `adjudications.json` row "at `pending-adjudication`". Zero rows are
   pending as of 2026-09-04.
2. `corpus/README.md:1341` calls `bug-locale-alias-ignored`
   "`pending-adjudication`, owner-only". It is `oracle-correct`.
3. `corpus/README.md:2317-2318` says "Every hypothesis but
   `bug-threshold-boundary` sits at `disposition:
   "pending-adjudication"`". Four are `oracle-wrong` and two are
   `oracle-correct`.
4. `corpus/README.md:2467-2468` says placeholder substitution is "a no-op
   on `stock.yaml` and `locale.yaml`, which carry neither". (Revision 1
   cited `:2469-2470`, which is the following sentence; fold, m2.)
   `corpus/profiles/locale.yaml:13-14` carries both
   `{{DUBPLATE_LIBRARY_DIR}}` and `{{DUBPLATE_LIBRARY_DB}}`, and has to,
   because it is rebased on `estate.yaml`. Only `stock.yaml` carries
   neither.

3b is the right pass for all four because it is the pass that rewrites the
adjudication schema and the pass that reads `locale.yaml` as its one
non-empty-`languages` fixture. A separate document lane would write the
same file.

### R17: the pseudo-release case count

**Resolved against the tree: four pseudo-release cases, three distinct
pseudo releases.** The inventory's §(a).8 (four cases) and §(b).7 (three
case directories) disagree, and §(a) is right. Measured mechanically over
`corpus/cases/*/responses/*.json`, exactly four case directories carry a
`transl-tracklisting` relation: `non-latin-pseudo-release`,
`real-nonlatin-girls-generation`, `real-nonlatin-zemfira`, and
`real-nonlatin-zemfira-stock`. §(b).7's count used a two-release-lookup
heuristic, and `real-nonlatin-girls-generation` carries six lookups
because it also has a search fallback. Task 0 verifies the count and
treats a fifth as a finding.

### R18: the remix pattern table

**Resolved: 3b lands a third `Tunables.StringDistPatterns` variant,
`SDPatternsNoBracketDiscount`.** 3a landed the seam
(`internal/engine/matcher/tunables.go:189-196`) with two tables,
`BeetsSDPatterns` (`:316-323`) and `SDPatternsWordBoundary` (`:333-340`).
The word-boundary variant covers the six `adv-ft-*` figures by assignment.
**No table covering the remix edit exists**, so the four `adv-remix-*`
figures have nothing to derive against. The mechanical form is recorded in
the corpus README: "the two 0.3-weight bracket patterns removed, which is
arithmetically the same as charging a parenthetical at full weight"
(`corpus/README.md:1107-1108`). The new table is `BeetsSDPatterns` with
the `\(.*?\)` and `\[.*?\]` rows dropped, four entries, with a doc comment
citing `#5868` and that README line.

### R19: the derivation's shim shape

**Resolved: `derive_album_info.py` takes `derive_ties.py`'s shape, not
`derive_component.py`'s.** Ruling E asks for "a third narrow beets entry
point in the shape of `derive_component.py` and 2d's `derive_file_tags.py`"
(spec `:119`), and **the narrowness that sentence means is preserved**:
one script, one `stable_json` document out, no `tag_album`, no verdict
path, no record validation. What cannot be preserved is the
no-stub-no-profile half, because the raw material here is a real release
response rather than a set of scalar arguments. It needs the replay stub, a
rendered profile, and the case's responses, because it must build the
`AlbumInfo` the way beets builds it rather than re-parsing a committed
body. That is exactly why `derive_ties.py` is the one narrow entry point
that already needs all three
(`internal/harness/oracle/ties.go:63-66`). It therefore
reuses `stub.New`, `renderProfile`, and `runPython` on the Go side
(`internal/harness/oracle/ties.go:63-66`, `:76`, `:100`) and imports
`prepare_config`, `pin_date`, and `stable_json` from `derive_oracle` on the
Python side (`derive_ties.py:28`).

`derive_defect_free.py` takes `derive_component.py`'s **config** shape:
no stub, no rendered profile, an empty scratch `BEETSDIR` so the run sits
on beets' shipped defaults
(`internal/harness/oracle/derive_component.py:443-447`), beets' stdout
redirected to stderr, one `stable_json` document out.

**It does read a response body, and revision 1's "no response ... touches
no corpus case" was unbuildable** (fold, B1). An album distance is
`distance(items, album_info, item_info_pairs)`
(`$SP/beets/autotag/distance.py:429-432`): `album_info` is the candidate's
whole translated release and `item_info_pairs` comes from `assign_items`
(`$SP/beets/autotag/match.py:53`). The candidate side exists nowhere in
the committed record. `corpus/cases/adv-ft-craft-titles/case.json`'s
`expected.candidates[0]` carries `source`, `album_id`, `distance`, a
per-key `penalties` map and three per-track distances, and no candidate
title, artist, length or track list at all. Nor is the arithmetic uniform
across the ten: the penalty the substitution moves is `album` on seven
rows, `artist` on `adv-ft-theft-artist`, and `tracks` on
`adv-ft-craft-titles` and `adv-remix-track-titles`, and the two `tracks`
rows route three per-track `track_title` distances per matched pair
through a cost matrix the substitution itself changes.

The buildable shape, verified end to end at fold time against pinned
beets:

1. The Go driver hands the script, per subject, the case's own
   release-lookup response path, the case's `input.items`, and its
   `eval_date`.
2. The script calls `pin_date(eval_date)` (see below), decodes the
   response, and runs `MusicBrainzAPI._normalize_data`
   (`$SP/beetsplug/_utils/musicbrainz.py:746-774`).
3. It builds the candidate with `MusicBrainzPlugin.album_info` and the
   items with `derive_oracle.build_items`' own shape
   (`derive_oracle.py:156-166`).
4. It **re-solves** the assignment with `match.assign_items` under the
   substituted table rather than reading `expected.assignment`, then
   calls `distance.distance`.

Measured this way, all four classes reproduce: `adv-ft-draft-colon`
`0.010185185185185186` and `adv-ft-craft-titles` `0.00917075163398693`
match their committed `case.json` distances exactly, and the substituted
tables produce `0.10185185185185186` and `0.09170751633986929`;
`adv-remix-both-sides` reads `0.023999999999999997` against `0.08` and
`adv-remix-track-titles` `0.02456140350877193` against
`0.08187134502923976`. Review 1 measured the optimal mapping unchanged on
all ten cases; the task **verifies** that rather than assuming it, and a
mapping change under the substituted table is a finding, not an expected
outcome.

**One mechanic the implementer would otherwise rediscover:**
`beets.autotag.distance` resolves to a **function** in the
`beets.autotag` package namespace, so the monkeypatch target is the
module object, reached through
`importlib.import_module("beets.autotag.distance")` or
`sys.modules["beets.autotag.distance"]`. A naive `from beets.autotag
import distance as D; D.SD_PATTERNS` raises `AttributeError`.

**The empty scratch `BEETSDIR` is arithmetically safe here, and the
reason is worth stating** (fold, m17). All ten cases derive under
`estate`, which sets `import.none_rec_action: skip` where beets ships
`ask` (`$SP/beets/config_default.yaml:34`, `corpus/profiles/estate.yaml`),
and sets no `match:` key at all, so no weight, threshold, or pattern
differs. The action half resolves identically because `quiet_action` sets
`import.quiet: True` and a medium recommendation then resolves through
`quiet_fallback`, which is `skip` in both
(`internal/harness/oracle/derive_oracle.py:193-206`).

### R20: the byte-identity invariant

**Resolved and stated as a task acceptance criterion, not a hope.** 2d's
pre-ruling requires that `make derive` re-derives every committed record
byte-identically after a change (spec `:374`). Tasks 1 and 2 each carry
it: after the task, `make derive` must leave
`corpus/analysis/components.json`, `ties.json`, `file_tags.json`, and every
`corpus/cases/*/case.json` unchanged, verified by `git status --porcelain`
being empty for those paths, and `TestShimStdoutIsByteIdentical`
(`internal/harness/oracle/derive_test.go:119`) must stay green. The
mechanism that makes this achievable is that both new scripts are **new
files** that import from `derive_oracle` without editing it, and that the
`SD_PATTERNS` edit is a runtime monkeypatch inside
`derive_defect_free.py`'s own process.

### R21: the widened clock check

**Resolved: Task 8 widens `TestMatcherNeverReadsTheClock`
(`internal/harness/conform/arch_test.go:147-153`) from
`internal/engine/matcher` to `internal/engine`.** Release-event date
selection is exactly where a `time.Now` would creep in, and the walk
already skips `testdata` subtrees (`arch_test.go:79-81`, the Task 6b
hardening). The widening needs its own firing proof, because the existing
fixture proves the checker fires at its root, not that the walk descends
into a sibling package; Task 8 ships a nested fixture for that.

**The walk's own words widen with it** (fold, m8). `checkNoTimeNow`'s
error text hardcodes "calls time.Now, which internal/engine/matcher must
never do" (`arch_test.go:107`) and its doc comment (`:70-71`) is
matcher-specific. A widened check reporting the old package name is a
message that sends the next reader to the wrong tree, so both move with
the test name.

---

## Task 0: The staleness pre-task

Model: `sonnet`, **fresh context**, dispatched before any implementer.
Reads this plan and the rung 3 spec against the tree as shipped. **Its
output is blockers with evidence, not fixes**; it writes no code and edits
no file. The conductor rules each blocker before Task 1 dispatches.

The checklist is the spec's own (`:395-406`), per pass. The first three
are blockers: a "no" stops the launch rather than producing a note.

1. **Have Phase 0a, 0b, and 0c landed?** 0a's `(5,3)` and `(8,3)` cells
   clear rung 3 to run out of ladder order; 0b owns the wall the four
   knobs register on; all three gate the second code lane. *(Expected at
   drafting time: 0a landed at `1b4ed9d`/`8545e71`/`0833589`; **0b and 0c
   unlanded, which is the expected state and not a blocker for 3b**. 3b
   registers nothing on the wall, writes no `internal/config`, and needs no
   worktree. Report the state; do not stop on it.)*
2. **Has rung 1b's registry shipped, and is the registration seam what
   this plan calls?** If not, which is the expected answer because
   `internal/config/keys.go:19-25` still carries `ownedWall` as a slice
   literal, confirm that the four-knob deferral is still recorded in 3a's
   HISTORY guard 2 and that `internal/config` is untouched by this plan.
3. **Is the lane registry consistent with where this pass intends to run,
   and is the main tree actually free?** `docs/STATUS.md`'s lane registry
   is the live-executor check, never `pgrep`. *(Expected, restated
   against the tree at fold time; fold, M-k: `docs/STATUS.md:29-33` still
   reads **lane-1 (code, main tree): PREP for 3b**, naming the plan-prep
   workflow `wf_5b24de07-ef0` as "running", which is the workflow this
   plan is the output of. It records **no pid at all**, where the
   lane-launch checklist requires "lane name, worktree path, pid"
   (`parallel-build-approach.md:194`). So the honest answer is "PREP, no
   pid recorded", and flipping lane-1 to a 3b-execution entry carrying a
   pid is a **conductor edit owed before Task 1 dispatches**, not a
   finding for the pre-task to litigate. lane-2 BLOCKED; doc lane IDLE.
   Report a lane-1 entry naming any workflow **other** than
   `wf_5b24de07-ef0`, or any live pid in this tree, as a blocker.)*
4. **Have the contested-surface rows been written into the parallel
   spec's table for every surface this pass touches?** *(Expected, fold
   M-j: **five rows present** and **three missing** at
   `docs/specs/2026-09-03-parallel-build-approach.md`'s
   contested-surface table. Missing: `internal/engine/{matcher,musicbrainz}`;
   `internal/harness/corpus_test.go`, which sits outside the
   `{cassette,stub,analysis}` row; and 3b's `analysis` additions, which
   that row's Owner column does not name because its one rung-3 carve-out
   is 3c's. **None of the three is a launch blocker**: the phase-1 lane
   table at `:80` already awards lane E2 `internal/engine/matcher`,
   `musicbrainz`, `shadow`, the checklist rule blocks the *second* lane,
   and 3b is the only live code lane. Report all three; the conductor
   writes them.)*
5. Does every corpus artifact this plan names exist, at the path and shape
   named? Specifically: `corpus/analysis/adjudications.json` parses as
   `analysis.AdjudicationLedger`; `corpus/analysis/oracle_bugs.json` as
   `analysis.OracleBugLedger`; every one of the **nineteen named ledger
   subjects** (R7's table) is a directory under `corpus/cases`; and
   `corpus/analysis/album_info.json` does **not** yet exist.
6. Does every harness API this plan calls have the signature named?
   `analysis.LoadAdjudications`, `analysis.CheckAdjudications`,
   `analysis.AdjudicationEntry`, `analysis.OracleBugEntry`,
   `analysis.LoadOracleBugs`, `analysis.AgentAdjudicator`,
   `analysis.OracleWrong`, `cassette.LoadCase`, `cassette.Case`,
   `stub.New`, `oracle.Options`, `oracle.DeriveFileTags`,
   `conform.CheckCaseWeights`, `matcher.AlbumInfo`, `matcher.TrackInfo`,
   `matcher.Tunables`, `matcher.DefaultTunables`,
   `matcher.BeetsSDPatterns`, `matcher.SDPatternsWordBoundary`,
   `matcher.SDPattern`.
7. **Have any `oracle_bugs.json` or `adjudications.json` dispositions
   moved since this plan was drafted**, and does any movement change 3b's
   scope? Report both disposition tables as read. *(Drafting-time
   snapshot: adjudications 21 rows, 18 `oracle-wrong`, 3 `oracle-correct`,
   0 pending; oracle_bugs 6 hypotheses, 4 `oracle-wrong`
   (`bug-ft-pattern`, `bug-remix-parenthetical`,
   `bug-nonlatin-alias-selection`, `bug-assignment-rejection-gap`), 2
   `oracle-correct` (`bug-threshold-boundary`,
   `bug-locale-alias-ignored`), 0 pending. A hypothesis moving off
   `oracle-wrong` on either string-heuristic class changes R12's predicate
   and is a real finding.)*
8. Have any `ties.json` dispositions or `proximity.json` rows moved?
   *(Drafting time: ties 215 rows, 212 unique, 2 `pinned`, 1 `skipped`;
   proximity 2 entries, 1 `skipped`, 1 `flagged`. 3b grades neither; 3d
   inherits both, so a movement is worth reporting once.)*
9. Do the corpus's committed case count and this plan's own measured
   figures match the tree? Read and report: `ls corpus/cases | wc -l`
   *(215)*; the number of release-lookup requests across the nineteen
   named subjects, counted as
   `jq '[.requests[] | select(.path_query | startswith("/ws/2/release/"))] | length'`
   per subject *(50 in total: 45 across the eighteen drafted subjects
   plus five for `real-noise-dylan-bootleg`)*; the number of case
   directories carrying a
   `transl-tracklisting` relation *(4, per R17)*; the number of case
   directories carrying a `video` recording *(2)*; the number carrying a
   non-empty `data-tracks` array *(0)*; the number carrying a non-null
   `pregap` *(0)*; and, because revision 1 got this one wrong and the
   corrected figure is what R15's declared gap now rests on (fold, B2),
   **the number of release events carrying a non-null `area` with
   `iso-3166-1-codes`** *(290 of 298 top-level events across the 303
   release-lookup responses, in 208 case directories; 405 of 413 counted
   recursively over all 351 responses)*. **Count mechanically; read no
   response body into context.**
10. Has `internal/config` changed shape? Has `internal/engine/musicbrainz`
    appeared? *(Expected: `internal/engine` holds `matcher` and nothing
    else.)*
11. Do `corpus/README.md`'s four stale cells still read as R16 records
    them (`:1096`, `:1341`, `:2317-2318`, `:2467-2468`)? If a correction
    already landed, Task 9's drafting half is smaller and says so.
12. Is the venv reconciled? Run `Reconcile`
    (`internal/harness/oracle/reconcile.go:36`) or read
    `corpus/manifest.json`'s six pins against `tools/venv/.venv`.
    *(Drafting time: beets 2.13.1, unidecode 1.4.0, jellyfish 1.2.1, lap
    0.5.13, numpy 2.5.2, mediafile 0.17.0.)* **3b runs `make derive`, so a
    venv drift is a blocker here in a way it was not for 3a.**

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
required; the task writes nothing.

---

## Task 1: The derived `AlbumInfo` ledger (ruling E)

Model: **`opus`**. The upshift is deliberate and is this pass's first of
two: the ledger is the written specification every later Go task in this
pass is graded against, it is derived through a beets entry point nobody
has driven before, and a ledger that silently records the wrong thing
produces five green tasks and a wrong translation layer. Depends on Task 0.

**This task runs first on purpose.** The ledger has zero dependency on the
Go translation, and once it is committed an implementer of Tasks 3 through
7 reads a committed row as the expected output of the function being
written.

**Files:**
- Create: `internal/harness/analysis/album_info.go` + `album_info_test.go`
- Create: `internal/harness/analysis/testdata/album_info/` fixtures for the
  firing proofs
- Create: `internal/harness/oracle/derive_album_info.py`
- Create: `internal/harness/oracle/album_info.go` (`//go:build oracle`) +
  `album_info_derive_test.go` (`//go:build oracle`)
- Create: `corpus/analysis/album_info.json` (generated, committed)
- Modify: `internal/harness/corpus_test.go` (one new gate-resident test,
  in `TestFileTagLedgerIsWellFormed`'s shape at `:270`)

**Interfaces:**

- Consumes: `matcher.AlbumInfo` and `matcher.TrackInfo`
  (`internal/engine/matcher/info.go:10-105`) as the field list to mirror,
  **by name only; `analysis` does not import `matcher`**;
  `cassette.LoadCase`, `cassette.Case`; `stub.New`; `oracle.Options`;
  `renderProfile` and `runPython`
  (`internal/harness/oracle/oracle.go:232`, `ties.go:120`);
  `derive_oracle`'s `prepare_config`, `pin_date`, `stable_json`
  (`internal/harness/oracle/derive_oracle.py:73` and siblings).
- Produces:

```go
// internal/harness/analysis/album_info.go

// AlbumInfoLedgerRowFloor is the row count corpus/analysis/album_info.json
// must not fall below: one row per release-lookup request across the
// nineteen named AlbumInfoSubjects, measured off the tree at 2026-09-04.
// A shrink here means the derivation quietly dropped a lookup rather
// than an edit anyone would notice in review.
const AlbumInfoLedgerRowFloor = 50

// AlbumInfoSubject names one committed case the AlbumInfo ledger derives
// rows for. Every release-lookup request in the case's own log becomes a
// row; a case is a subject because a named translation-layer branch runs
// through it, and Why carries that reason beside the id rather than in a
// plan nobody reads at execution time. CheckAlbumInfoLedger requires Why
// to be non-empty, so a subject added without a rationale fails the gate
// (fold, m3: FileTagSubject puts its rationale in the slice's own doc
// comment, and this type diverges on purpose, with a reader, rather than
// carrying a field nothing reads).
type AlbumInfoSubject struct {
    CaseID string
    Why    string
}

// AlbumInfoPseudoSubjects names the committed cases whose release set
// contains a pseudo-release, measured off the tree at 2026-09-04:
// exactly four case directories carry a relation with type
// "transl-tracklisting" and direction "backward"
// (beetsplug/musicbrainz.py:785-798). It exists so CheckAlbumInfoLedger's
// Merged check compares against a named expectation instead of deriving
// one from the rows it is checking, which would pass on a ledger
// carrying no Merged block at all.
var AlbumInfoPseudoSubjects = []string{
    "non-latin-pseudo-release",
    "real-nonlatin-girls-generation",
    "real-nonlatin-zemfira",
    "real-nonlatin-zemfira-stock",
}

// AlbumInfoSubjects is the fixed nineteen-case subject set. Every
// release-lookup request in each case's own log becomes a row, which is
// 50 rows at 2026-09-04.
var AlbumInfoSubjects = []AlbumInfoSubject{
    {"clean-match", "the plain baseline: single medium, single credit, no alias, one release event taken through the fallback"},
    {"media-baseline", "mediums counts unfiltered media while extraction filters them (musicbrainz.py:659 against :638-640)"},
    {"media-uneven-split", "index assignment 1-based across all valid media (musicbrainz.py:645-646)"},
    {"va-exact", "the various-artists override and the va_name read (musicbrainz.py:673-681)"},
    {"various-artists", "the VA override beside the corpus's one release whose two events carry different dates"},
    {"should-reject", "search fallback: six release lookups in one case, plus multiple release events"},
    {"non-latin-pseudo-release", "pseudo-release merge; six fields move"},
    {"real-nonlatin-zemfira", "pseudo-release merge; three fields move"},
    {"real-nonlatin-zemfira-stock", "the same merge under the stock profile"},
    {"real-nonlatin-girls-generation", "pseudo-release merge, an oracle-wrong alias probe, and three video recordings at once"},
    {"real-nonlatin-anna-vissi-essential", "oracle-wrong alias probe (adj-adv-alias-vissi-native-greek)"},
    {"real-nonlatin-alla-pugacheva-vol1", "oracle-correct alias probe: entity name and release credit coincide"},
    {"real-nonlatin-x-japan-art-of-life", "oracle-correct alias probe on a single-track release"},
    {"real-nonlatin-utada-first-love", "bug-locale-alias-ignored's A half, under the estate profile"},
    {"real-nonlatin-utada-first-love-locale", "the B half, and the corpus's only case running a non-empty import.languages"},
    {"real-nonlatin-hamasaki-ayumi", "138 video recordings: ignore_video_tracks empties a medium and _add_candidate drops the candidate"},
    {"real-genre-metallica-master-of-puppets", "one release event whose date equals the top-level date, so the fallback supplies country and date"},
    {"real-compilation-queen-greatest-hits-ii", "two release events, both equal to the top-level date: the same fallback under a second stratum"},
    {"real-noise-dylan-bootleg", "the mixed-format route to the \"Media\" sentinel: two Hybrid SACD layers on one release (musicbrainz.py:654-658)"},
}

// AlbumInfoLedger is corpus/analysis/album_info.json: the translated
// AlbumInfo every candidate the matcher scores is built from, dumped
// through pinned beets so the Go translation layer is graded against
// beets' own answer field by field rather than only through whole-case
// agreement (ruling E).
type AlbumInfoLedger struct {
    ShimVersion  string               `json:"shim_version"`
    BeetsVersion string               `json:"beets_version"`
    Rows         []AlbumInfoLedgerRow `json:"rows"`
}

// AlbumInfoLedgerRow is one release lookup's translation. Merged is
// non-nil only on a pseudo-release whose transl-tracklisting merge
// fired, and then Album carries the pre-merge value, so the merge is
// gradeable on its own rather than only through its result.
type AlbumInfoLedgerRow struct {
    CaseID       string           `json:"case_id"`
    ReleaseMBID  string           `json:"release_mbid"`
    Album        LedgerAlbumInfo  `json:"album"`
    Merged       *LedgerAlbumInfo `json:"merged,omitempty"`
}

// LedgerAlbumInfo mirrors matcher.AlbumInfo field for field. Every
// scalar is a pointer so a beets None survives the round trip as a JSON
// null and is distinguishable from the Go zero value: this is the JSON
// boundary case go-conventions carves out, not defensive nil checking,
// and internal/harness/conform's nullability policy is what reads the
// distinction.
type LedgerAlbumInfo struct {
    Artist        *string           `json:"artist"`
    Album         *string           `json:"album"`
    VA            *bool             `json:"va"`
    Media         *string           `json:"media"`
    Mediums       *int              `json:"mediums"`
    Year          *int              `json:"year"`
    OriginalYear  *int              `json:"original_year"`
    Country       *string           `json:"country"`
    Label         *string           `json:"label"`
    CatalogNum    *string           `json:"catalognum"`
    AlbumDisambig *string           `json:"albumdisambig"`
    AlbumID       *string           `json:"album_id"`
    DataSource    *string           `json:"data_source"`
    Tracks        []LedgerTrackInfo `json:"tracks"`
}

// LedgerTrackInfo mirrors matcher.TrackInfo field for field, on
// LedgerAlbumInfo's own pointer convention.
type LedgerTrackInfo struct {
    Title       *string  `json:"title"`
    Artist      *string  `json:"artist"`
    Length      *float64 `json:"length"`
    Index       *int     `json:"index"`
    MediumIndex *int     `json:"medium_index"`
    Medium      *int     `json:"medium"`
    TrackID     *string  `json:"track_id"`
    DataSource  *string  `json:"data_source"`
}

// LoadAlbumInfoLedger reads corpus/analysis/album_info.json from path.
func LoadAlbumInfoLedger(path string) (*AlbumInfoLedger, error)

// CheckAlbumInfoLedger validates the ledger's schema and its coverage
// against the committed corpus, reporting every failure rather than the
// first.
func CheckAlbumInfoLedger(cases map[string]*cassette.Case, l *AlbumInfoLedger) error
```

```go
// internal/harness/oracle/album_info.go   //go:build oracle

// DeriveAlbumInfo runs derive_album_info.py against one committed case
// and returns the rows it prints, one per release-lookup request in the
// case's own log, unparsed.
func DeriveAlbumInfo(ctx context.Context, c *cassette.Case, opts Options) ([]byte, error)
```

**Behavior contract:**

- **`CheckAlbumInfoLedger` does six things**, each with its own firing
  fixture, on `CheckFileTagLedger`'s model
  (`internal/harness/analysis/file_tags.go:100-106`):
  1. `shim_version` and `beets_version` are non-empty.
  2. Every row's `case_id` resolves to a committed case in `cases`.
  3. The set of `case_id`s across all rows equals `AlbumInfoSubjects`'
     ids, **in both directions**. A missing subject and an extra case are
     both failures. Every subject's `Why` is non-empty (fold, m3).
  4. Every row's `release_mbid` appears in that case's own request log as
     a `/ws/2/release/<mbid>` path, and every such request in a subject
     case has exactly one row. This is the check that catches a derivation
     that silently skipped a lookup.
  5. `len(l.Rows) >= AlbumInfoLedgerRowFloor`.
  6. `Merged` is non-nil on exactly the rows whose case is named in
     `AlbumInfoPseudoSubjects` and whose release is the pseudo half.
     **Both directions are checked against that named list**: no case
     outside it carries a `Merged` block, and every case in it has
     exactly one row that does. Revision 1 derived the pseudo set from
     the rows themselves and then asserted that no case outside the
     derived set carried one, which is true by construction: a
     derivation emitting no `Merged` block at all would leave the set
     empty, the check green, and the merge silently ungraded, which is
     the "instrument that can pass silently" shape the spec's risk 2
     names (fold, M-b). `analysis` cannot resolve the pseudo half
     honestly any other way, because it must not import `internal/engine`
     and the predicate lives in the response body.
- **`derive_album_info.py` takes `derive_ties.py`'s shape** (R19),
  invoked as `derive_album_info.py <rendered-config.yaml> <input.json>`.
  Its input envelope carries `case_id`, `eval_date`, and the list of
  `release_mbid` lookups read off the case's own request log by the Go
  driver. **`is_search` is not carried** (fold, m6): nothing consumes it,
  T1 establishes that a search candidate and a direct id both reach
  `album_info` through `album_for_id`, and all 50 lookups are mbid paths
  because the search endpoint is `/ws/2/release?` and the driver's own
  filter excludes it. R9 rejected a dead parameter for the same reason.
  For each lookup it fetches the release through the
  plugin's own API against the stub, calls
  `MusicBrainzPlugin.album_info(release)` for the pre-merge value. When
  the release is a pseudo-release with a backward `transl-tracklisting`
  relation it also calls `album_for_id` for the merged value. It writes one `stable_json` document to stdout, beets' own stdout
  redirected to stderr (`derive_component.py:448-451`'s pattern).
- **The script asserts its own isolation and its own seed.** `BEETSDIR`
  isolation is asserted at the moment beets is first imported, exiting
  loudly if the runner's own config leaked in
  (`derive_file_tags.py:74-81`'s pattern); `PYTHONHASHSEED` must be `0`
  (`derive_ties.py:136-137`'s pattern), because the ledger records set-derived
  values (`media`) whose iteration order is hash-dependent.
- **A fresh plugin instance per config** (T15). The five `@cached_property`
  reads are per instance, so the script constructs the plugin after
  `prepare_config` and never mutates config afterwards.
- **The projection is total and typed.** Every one of the twenty-two
  fields is written for every row; a beets `None` is written as JSON
  `null`, never omitted, never coerced to `""`. The script **refuses any
  value that is not already a JSON primitive**, exiting rather than
  falling back to `str(value)`. `derive_file_tags.py:92-109` is the
  precedent, and it exists because a silent `str()` is how a ledger ships a
  wrong type.
- **The oracle-tagged re-derive test has the house three-part shape**
  (`internal/harness/oracle/file_tags_derive_test.go:65-111`):
  `venvOptions(t)`, re-derive every subject; run
  `analysis.CheckAlbumInfoLedger` over the **fresh** derivation first,
  failing with "the freshly derived ledger does not satisfy its own
  schema" before any diff runs; honor the shared `-update-ledgers` flag
  (`internal/harness/oracle/ties_derive_test.go:26`) by rewriting the
  committed file and returning; otherwise diff **decoded values, never
  bytes**, through `decodeAny` (`component_derive_test.go:66`), with the
  same one-sentence reason the two existing tests carry.
- **The perturbation test is the teeth**, on
  `TestFileTagsLedgerReDerivationCatchesAPerturbation`'s model
  (`file_tags_derive_test.go:119`): perturb one row's `album.album` in a
  copy of the committed bytes and assert the comparison fires.
- **`internal/harness/corpus_test.go` gains
  `TestAlbumInfoLedgerIsWellFormed`**, gate-resident, wiring
  `LoadAlbumInfoLedger` and `CheckAlbumInfoLedger` to the real corpus in
  `TestFileTagLedgerIsWellFormed`'s shape (`:270-284`).
- **The regeneration command joins `corpus/README.md`'s list** (`:2214-2218`)
  as a drafted correction for the conductor:
  `go test -tags oracle ./internal/harness/oracle -run AlbumInfo -update-ledgers`.
- **The byte-identity invariant** (R20): after this task, `make derive`
  leaves `components.json`, `ties.json`, `file_tags.json`, and every
  `case.json` unchanged, and `TestShimStdoutIsByteIdentical` stays green.
  Verified with `git status --porcelain corpus/` and reported.

**Named mutations this task reports:**

1. A committed ledger with `shim_version` deleted → `CheckAlbumInfoLedger`
   fails naming the field.
2. A row whose `case_id` is `no-such-case` → fails naming the id as
   unresolvable.
3. A ledger with `real-nonlatin-hamasaki-ayumi`'s rows removed → fails
   naming the missing subject, proving check 3's first direction.
4. A ledger with a row for `real-boxset-mahler-complete` (a real case,
   not a subject) → fails, proving check 3's second direction.
5. A ledger with one of `should-reject`'s six rows removed → fails naming
   the release mbid whose request has no row, proving check 4 (this is the
   silent-skip mutation, and it is the one that matters most).
6. A ledger one row below `AlbumInfoLedgerRowFloor` → fails.
7. A `Merged` block added to `clean-match`'s row → fails naming the case,
   proving check 6's first direction.
7b. `non-latin-pseudo-release`'s `Merged` block removed → fails naming
   that case, proving check 6's second direction. Revision 1 carried only
   mutation 7, which a self-derived pseudo set passes (fold, M-b).
8. One `album.album` perturbed in a copy of the committed bytes → the
   oracle-tagged re-derive diff fires (the assertion-discipline proof that
   the ledger is compared at all).

**Deliverable count: 4** (the ledger schema, both subject lists, loader
and six-part checker in `analysis` with its firing fixtures; the Python
entry point and its Go driver in `oracle`; the committed
`corpus/analysis/album_info.json` over 19 subjects and 50 rows; the
oracle-tagged re-derive and perturbation tests plus the gate-resident
`corpus_test.go` wiring). Gate green plus tier-0 replay green; `make
derive` byte-identity verified; commit. **Report actual token spend and
the derivation's wall time.**

---

## Task 2: The `defect_free` block, its derivation, and its checkers (ruling A / M6)

Model: `sonnet`. Depends on Task 1 (for the shim conventions it copies),
not on any Go translation work.

**Files:**
- Modify: `internal/engine/matcher/tunables.go` + `tunables_test.go` (one
  addition: `SDPatternsNoBracketDiscount`)
- Modify: `internal/harness/analysis/adjudication.go` +
  `adjudication_test.go`
- Create: `internal/harness/analysis/testdata/defect_free/` fixtures
- Create: `internal/harness/oracle/derive_defect_free.py`
- Create: `internal/harness/oracle/defect_free.go` (`//go:build oracle`) +
  `defect_free_derive_test.go` (`//go:build oracle`)
- Modify: `corpus/analysis/adjudications.json` (**the `defect_free` field
  on ten rows and nothing else**, R14)
- Modify: `internal/harness/corpus_test.go` (one new gate-resident test)

**Interfaces:**

- Consumes: `analysis.AdjudicationEntry`, `analysis.OracleBugEntry`,
  `analysis.LoadAdjudications`, `analysis.LoadOracleBugs`,
  `analysis.OracleWrong` (`internal/harness/analysis/adjudication.go:17`,
  `oraclebug.go`); `matcher.SDPattern`, `matcher.BeetsSDPatterns`,
  `matcher.SDPatternsWordBoundary`
  (`internal/engine/matcher/tunables.go:316-340`).
- Produces:

```go
// internal/engine/matcher/tunables.go

// SDPatternsNoBracketDiscount carries BeetsSDPatterns' six discounts
// with the two 0.3-weight bracket patterns removed. The corpus README
// calls that "arithmetically the same as charging a parenthetical at
// full weight" (corpus/README.md:1107-1108), which holds only while no
// later pattern fires: a full-weight discount still shifts str1, str2
// and base_dist for the rest of the loop where a removed pattern does
// not (beets/autotag/distance.py:97-117). The two agree on all four
// adv-remix-* pairs, so the committed figures stand; the table is
// defined by what it removes, not by that equivalence. It is the table
// the four adv-remix-* cases' defect_free figures are derived under,
// and the counterpart to SDPatternsWordBoundary for the adjudicated
// bug-remix-parenthetical defect (#5868). Nothing on the distance path
// selects it; 3d's compat switch assigns it to
// Tunables.StringDistPatterns.
var SDPatternsNoBracketDiscount = []SDPattern{
    {Pattern: `^the `, Weight: 0.1},
    {Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
    {Pattern: `[\[\(]?(featuring|feat|ft)[\. :].+`, Weight: 0.1},
    {Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
}
```

```go
// internal/harness/analysis/adjudication.go

// StringHeuristicHypotheses names the oracle-bug hypotheses whose
// defect is a defect OF SD_PATTERNS, and therefore the only ones whose
// corrected expectation a pattern-table swap can express. The other two
// oracle-wrong hypotheses are excluded on purpose:
// bug-nonlatin-alias-selection is a translation defect with no
// SD_PATTERNS counterfactual, and bug-assignment-rejection-gap is the
// class ruling A names as having no beets counterfactual at all, so its
// six adv-assign-* probe cases owe no block.
var StringHeuristicHypotheses = []string{"bug-ft-pattern", "bug-remix-parenthetical"}

// DefectFree is the distance a case scores once the named defect is
// removed, derived through pinned beets with one SD_PATTERNS
// substitution and nothing else. It is derived data beside the owner's
// own corrected_verdict prose, never a disposition and never a verdict
// about who is right: it is the expectation an oracle-wrong case is
// graded against once its hypothesis's compat switch defaults off
// (ruling A / M6).
type DefectFree struct {
    Distance       float64 `json:"distance"`
    Recommendation string  `json:"recommendation"`
    Action         string  `json:"action"`
    PatternTable   string  `json:"pattern_table"`
}

// AdjudicationEntry gains, as its last field:
//     DefectFree *DefectFree `json:"defect_free,omitempty"`
// A pointer because absence is meaningful and omitempty does not omit a
// struct value.

// CheckDefectFree proves the cross-ledger rule the per-row checker
// cannot see: every probe case of a StringHeuristicHypotheses
// hypothesis at OracleWrong resolves to an adjudication row carrying a
// defect_free block, and every block's recommendation and action appear
// verbatim in its row's corrected_verdict. Reports every failure rather
// than the first.
func CheckDefectFree(bugs *OracleBugLedger, adj *AdjudicationLedger) error
```

```go
// internal/harness/oracle/defect_free.go   //go:build oracle

// DefectFreeSubject is one string-heuristic case the defect-free
// derivation runs over. ResponsePath and Items are what the script
// needs to rebuild the candidate and the item set: an album distance
// takes the candidate's whole track list and a solved assignment
// (autotag/distance.py:429-432, autotag/match.py:53), neither of which
// the committed record carries. PatternTable names the substitution the
// figures are derived under, and is one of the two legal values
// DefectFree.PatternTable carries.
type DefectFreeSubject struct {
    CaseID         string
    AdjudicationID string
    ReleaseMBID    string
    ResponsePath   string
    EvalDate       string
    PatternTable   string
}

// DeriveDefectFree runs derive_defect_free.py over the ten
// string-heuristic subjects and returns the figures it prints,
// unparsed: one document carrying {case_id, distance, recommendation,
// action, pattern_table} per subject.
func DeriveDefectFree(ctx context.Context, opts Options, subjects []DefectFreeSubject) ([]byte, error)
```

**Behavior contract:**

- **`checkAdjudicationEntry` gains three within-row rules**, appended to
  its existing list (`adjudication.go:100-146`) and reporting rather than
  returning early:
  1. A `defect_free` block on a row whose disposition is not
     `OracleWrong` is a failure. The block is the expectation a
     corrected verdict is graded against, and a row with no corrected
     verdict has nothing to express.
  2. A block with an empty `recommendation`, `action`, or `pattern_table`
     is a failure naming the field.
  3. A block whose `distance` is outside `[0, 1]` is a failure naming the
     value. beets' own `Distance.add` raises outside that range
     (3a plan, contract 6), so a figure outside it is a derivation
     defect, not a datum.
- **`CheckDefectFree` is the cross-ledger half**, and it lives in its own
  function rather than inside `checkAdjudicationEntry` because that
  checker sees one entry and cannot resolve a hypothesis. It resolves each
  `StringHeuristicHypotheses` hypothesis at `OracleWrong` to its
  `ProbeCaseIDs`, resolves each probe case to an adjudication row by
  `Subject.CaseID`, and requires a `defect_free` block on it. **A probe
  case with no row, or a row with no block, is a gate failure naming the
  case**, never a silently ungraded case. It then asserts the block's
  `Recommendation` and `Action` appear verbatim in the row's
  `CorrectedVerdict` (R13's mechanical half).
- **`PatternTable` names the table, never the regex.** Its two legal
  values today are `"SDPatternsWordBoundary"` for the six `adv-ft-*` rows
  and `"SDPatternsNoBracketDiscount"` for the four `adv-remix-*` rows.
  `CheckDefectFree` asserts the value is one of the two and names both on
  failure, so a hand-edit to a third string fails loudly rather than
  becoming an unreproducible figure.
- **`derive_defect_free.py` takes `derive_component.py`'s config shape and
  its own input envelope** (R19 as folded, B1): no stub and no rendered
  profile, an empty scratch `BEETSDIR` so the run sits on beets' shipped
  defaults (`derive_component.py:443-447`), beets' stdout redirected to
  stderr, one `stable_json` document out, `SHIM_VERSION` pinned. It
  **does** read the case's own release-lookup response, because an album
  distance cannot be computed from scalars. Per subject it runs four
  steps in order: `pin_date(eval_date)` and
  `MusicBrainzAPI._normalize_data` over the decoded body;
  `MusicBrainzPlugin.album_info`; `match.assign_items` over the case's
  `input.items`, **re-solved under the substituted table** rather than
  read from `expected.assignment`, because a changed `SD_PATTERNS`
  changes every `track_title` distance and therefore the cost matrix; and
  `distance.distance`. The optimal mapping is expected to be unchanged on
  all ten cases and the task **verifies** that rather than assuming it; a
  mapping change is a finding.
- **The monkeypatch target is the module object, not the attribute.**
  `beets.autotag.distance` resolves to a function in the `beets.autotag`
  package namespace, so the script reaches the table through
  `importlib.import_module("beets.autotag.distance")` or
  `sys.modules["beets.autotag.distance"]`. `from beets.autotag import
  distance as D; D.SD_PATTERNS` raises `AttributeError`.
- **`pin_date(eval_date)` is called even though no case reads the date**
  (fold, m16). `distance()` divides by `datetime.date.today().year` in
  both year branches (`$SP/beets/autotag/distance.py:477`, `:488`), all
  ten cases take `dist.add("year", 0.0)` at `:483`, and
  `derive_oracle.pin_date` (`internal/harness/oracle/derive_oracle.py:52-69`)
  exists for exactly this reason. A later `defect_free` population on a
  reissue case would otherwise drift with the wall clock.
- **The empty scratch `BEETSDIR` is arithmetically equivalent to the
  `estate` profile here, stated rather than left to be rediscovered**
  (fold, m17). See R19 for the reasoning: no `match:` key differs, and
  the action half resolves through `quiet_fallback`, `skip` in both.
- **The `SD_PATTERNS` substitution is a runtime monkeypatch**, applied to
  `beets.autotag.distance.SD_PATTERNS` inside the script's own process and
  restored afterwards, exactly as `derive_component.py:168-190` swaps
  `Distance._weights` and restores. **It never edits beets, never edits
  `derive_oracle.py`, and never edits the venv** (R20). The two
  substitution tables are written in the script as the same literals
  `matcher.SDPatternsWordBoundary` and
  `matcher.SDPatternsNoBracketDiscount` carry, so a divergence between the
  Go table and the derivation's table is visible in review rather than
  hidden in a monkeypatch. **One divergence the identical literals cannot
  close is recorded as a forward carry to 3d** (fold, m19):
  `SDPatternsWordBoundary`'s `\b` is Unicode-aware in Python's `re` on
  `str` and ASCII-only in Go's RE2, so on non-ASCII input the derived
  figure and 3d's compat-switch result can differ. All six `adv-ft-*`
  pairs are ASCII, so nothing moves today. It travels beside 3a's
  `SD_PATTERNS`-as-a-value seam (HISTORY guard 3).
- **The derivation is checked both ways, as the corpus README already
  requires** (`corpus/README.md:1109-1112`): run with `SD_PATTERNS`
  untouched, the script must reproduce each case's committed distance
  **bit-identically against `case.json`'s own value**, and the test
  asserts that before it asserts anything about the defect-free figure. A
  script that cannot reproduce the committed number has not proved its
  counterfactual half.
- **Two comparisons at two precisions, and conflating them reds the
  gate** (fold, M-a). The committed records carry full `float64`:
  `corpus/cases/adv-remix-both-sides/case.json` reads
  `0.023999999999999997` and `adv-ft-theft-artist` reads
  `0.00909090909090909`, against the table's `0.024000000` and
  `0.009090909`. So the both-ways check compares against `case.json`, not
  against the table; the **block stores the derivation's own unrounded
  `float64`**, which is what makes mutation 7's one-ULP perturbation
  meaningful; and the table below is the one-time cross-check against the
  owner's prose, at the precision the prose states.
- **The first population is the ten cases**, with the figures below
  written into this task's acceptance rather than computed by it. Each
  derived `distance` must equal the figure the row's own `suspected`
  prose already carries, at the precision the prose states, and **the task
  reports each pair** (R13's one-time numeric half). The **write-back
  mechanism is R14's**, a targeted Python writer, not a whole-ledger
  re-serialize:

| Row id | Case | Derived distance | Defect-free distance | `pattern_table` |
| --- | --- | --- | --- | --- |
| `adj-adv-ft-draft-colon` | `adv-ft-draft-colon` | 0.010185185 | **0.101851852** | `SDPatternsWordBoundary` |
| `adj-adv-ft-left-colon` | `adv-ft-left-colon` | 0.011764706 | **0.117647059** | `SDPatternsWordBoundary` |
| `adj-adv-ft-gift-space` | `adv-ft-gift-space` | 0.010784314 | **0.107843137** | `SDPatternsWordBoundary` |
| `adj-adv-ft-soft-onesided` | `adv-ft-soft-onesided` | 0.030095238 | **0.086666667** | `SDPatternsWordBoundary` |
| `adj-adv-ft-craft-titles` | `adv-ft-craft-titles` | 0.009170752 | **0.091707516** | `SDPatternsWordBoundary` |
| `adj-adv-ft-theft-artist` | `adv-ft-theft-artist` | 0.009090909 | **0.090909091** | `SDPatternsWordBoundary` |
| `adj-adv-remix-both-sides` | `adv-remix-both-sides` | 0.024000000 | **0.080000000** | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-onesided` | `adv-remix-onesided` | 0.021428571 | **0.071428571** | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-bracket` | `adv-remix-bracket` | 0.027777778 | **0.092592593** | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-track-titles` | `adv-remix-track-titles` | 0.024561404 | **0.081871345** | `SDPatternsNoBracketDiscount` |

  All ten carry `recommendation: "medium"` and `action: "skip"`. **A
  mismatch between a derived figure and its prose figure is a stop and a
  finding, not a silent overwrite of either.**

- **R14's byte-identity criterion is this task's acceptance line, stated
  here rather than left in a resolution** (fold, M-f). After the
  population, a `git diff` over `corpus/analysis/adjudications.json` must
  show added `defect_free` blocks on exactly ten rows and **no change to
  any of the ten existing fields on any row**. The task reports `git
  diff --stat` and a per-field diff over the ten rows as its evidence. A
  change to `disposition`, `adjudicator`, `date`, `corrected_verdict`,
  `suspected`, `release_truth`, `reason`, `id`, `subject`, or `citation`
  is a task failure, not a fix.
- **`CheckDefectFree` is enforced in both directions** (fold, m12): a
  block is required on the ten string-heuristic probe cases **and
  forbidden on any row whose case is not a probe of a
  `StringHeuristicHypotheses` hypothesis**, which is what stops a
  fabricated block on one of the six `adv-assign-*` rows from passing
  every check.

- **`SDPatternsNoBracketDiscount` gets a discriminating test**, on the
  precedent 3a's own Task 4 set (3a plan, Task 4's fold M-a): a fixture pair
  whose `StringDist` differs measurably under `BeetsSDPatterns` and under
  the new table, asserted against both, so the third table is a tested
  value rather than dead weight. Use `adv-remix-both-sides`'s own pair,
  whose `string_dist` reads 0.144000 discounted against a true 0.480000
  (`corpus/README.md:1186-1189`; verify off the tree).
- **The byte-identity invariant** (R20), verified and reported.
- **`internal/harness/corpus_test.go` gains
  `TestDefectFreeBlocksAreWellFormed`**, gate-resident, loading both
  ledgers and running `CheckDefectFree`, in
  `TestOracleBugsLedgerIsWellFormed`'s shape (`:313`).

**Named mutations this task reports:**

1. `adj-adv-ft-draft-colon`'s `defect_free` block deleted from a copy of
   the committed ledger → `CheckDefectFree` fails naming the case.
2. A `defect_free` block added to `adj-real-nonlatin-alla-pugacheva-vol1`
   (an `oracle-correct` row) → `checkAdjudicationEntry` rule 1 fails.
3. A block with `pattern_table: ""` → rule 2 fails naming the field.
4. A block with `distance: 1.5` → rule 3 fails naming the value.
5. A block with `pattern_table: "BeetsSDPatterns"` → `CheckDefectFree`
   fails naming both legal values.
6. A block with `recommendation: "strong"` on a row whose
   `corrected_verdict` reads "medium, skip" → `CheckDefectFree`'s
   categorical cross-check fails.
7. One committed `defect_free.distance` perturbed by one ULP → the
   oracle-tagged re-derive diff fires.
8. The monkeypatch removed from the derivation (leaving `SD_PATTERNS` at
   beets' shipped table) → the script's both-ways check still passes but
   every defect-free figure equals its derived figure, and the re-derive
   diff fires on all ten rows.

**Deliverable count: 4** (`SDPatternsNoBracketDiscount` with its
discriminating test; the `DefectFree` schema plus the three within-row
rules and `CheckDefectFree`'s cross-ledger rule, with their firing
fixtures; the Python entry point, its Go driver, and the both-ways
reproduction check; the first population of the ten rows with every
derived figure cross-checked against its prose figure and reported). Gate
green plus tier-0 replay green; `make derive` byte-identity verified;
commit.

---

## Task 3: `internal/engine/musicbrainz`, the release vocabulary, normalization, and the boundary

Model: `sonnet`. **Births `internal/engine/musicbrainz`.** Depends on
Task 0; independent of Tasks 1 and 2, and sequenced after them so the
committed ledger is available as every later task's expected output.

**Files:**
- Create: `internal/engine/musicbrainz/doc.go`
- Create: `internal/engine/musicbrainz/release.go` + `release_test.go`
- Create: `internal/engine/musicbrainz/relations.go` + `relations_test.go`
- Create: `internal/engine/musicbrainz/testdata/` fixtures
- Modify: `internal/engine/matcher/tunables.go` + `tunables_test.go` (one
  addition: `VAName`)

**Interfaces:**

- Consumes: `matcher.Tunables`, `matcher.DefaultTunables`
  (`internal/engine/matcher/tunables.go:42-276`).
- Produces:

```go
// internal/engine/matcher/tunables.go, one new field on Tunables:

    // VAName is the va_name config key: the artist string every
    // artist field is overwritten with on a various-artists release
    // (beetsplug/musicbrainz.py:673-681). Default "Various Artists"
    // (config_default.yaml:106). It moves no distance, because the
    // album artist component is skipped entirely when AlbumInfo.VA
    // (album_artist.go:13-15). It does move AlbumInfo.Artist, which
    // the ruling-E ledger grades field by field, so it is a tunable
    // here rather than a literal in the translation layer. No
    // committed case echoes it, so the M1 defaults cross-check cannot
    // cover it; its coverage is this package's own unit test and the
    // translation layer's VA test.
    //
    // musicbrainz.aliases_as_credits is deliberately NOT here: it
    // selects only which object supplies artist_credit
    // (musicbrainz.py:311-313), and artist_credit is neither in
    // AlbumInfo nor in the ledger's graded set, so the translation
    // never reads it.
    VAName string
```

```go
// internal/engine/musicbrainz/release.go

// ErrMalformedRelease is the one sentinel every boundary failure in
// this package wraps, so a caller (3c's client) branches once on "this
// response is unusable" rather than on a hierarchy of error types.
var ErrMalformedRelease = errors.New("malformed musicbrainz release")

// Release is the subset of a WS/2 release beets' own album_info reads.
// Several of its fields feed applied metadata rather than a distance
// (medium title and track count, track id and number, alias sort names,
// barcode, the release-group type fields), so nothing in 3b reads them;
// they are decoded here because the wire shape is decoded once and
// rung 8 should not re-derive it. That is the deliberate counterpart to
// dropping credits.SortName (fold, M-i): the wire vocabulary names the
// payload, and the computed vocabulary produces only what the matcher
// reads. The struct tags name the wire keys directly, which discharges the
// hyphen-to-underscore half of the client's own _normalize_data
// (beetsplug/_utils/musicbrainz.py:774) with no runtime pass: that
// rewrite exists so a Python dict can be read with underscore keys, and
// a Go struct does not need it. The relations regrouping half is
// structural and lives in GroupRelations.
type Release struct {
    ID             string         `json:"id"`
    Title          string         `json:"title"`
    Status         string         `json:"status"`
    Disambiguation string         `json:"disambiguation"`
    Country        *string        `json:"country"`
    Date           string         `json:"date"`
    Barcode        *string        `json:"barcode"`
    Aliases        []Alias        `json:"aliases"`
    ArtistCredit   []CreditPart   `json:"artist-credit"`
    ReleaseEvents  []ReleaseEvent `json:"release-events"`
    LabelInfo      []LabelInfo    `json:"label-info"`
    Media          []Medium       `json:"media"`
    ReleaseGroup   ReleaseGroup   `json:"release-group"`
    Relations      []Relation     `json:"relations"`
}

// Alias is one MusicBrainz alias. Primary is a pointer because the wire
// value is bool | null and null is falsy in beets' own read
// (_utils/musicbrainz.py:138-155).
type Alias struct {
    Name     string  `json:"name"`
    SortName string  `json:"sort-name"`
    Locale   *string `json:"locale"`
    Primary  *bool   `json:"primary"`
    Type     *string `json:"type"`
}

// CreditPart is one element of an artist-credit array.
type CreditPart struct {
    Name       string `json:"name"`
    JoinPhrase string `json:"joinphrase"`
    Artist     Artist `json:"artist"`
}

// Artist is a credit element's artist entity.
type Artist struct {
    ID       string  `json:"id"`
    Name     string  `json:"name"`
    SortName string  `json:"sort-name"`
    Aliases  []Alias `json:"aliases"`
}

// ReleaseEvent is one dated release event. Area is a pointer because
// the wire value is nullable and beets' walrus at
// beetsplug/musicbrainz.py:174 skips a null one.
type ReleaseEvent struct {
    Date string `json:"date"`
    Area *Area  `json:"area"`
}

// Area carries the ISO codes release-event selection tests membership
// against.
type Area struct {
    ISO31661Codes []string `json:"iso-3166-1-codes"`
}

// LabelInfo is one label/catalog-number pair.
type LabelInfo struct {
    CatalogNumber *string `json:"catalog-number"`
    Label         *Label  `json:"label"`
}

// Label is a label entity.
type Label struct {
    Name string `json:"name"`
}

// ReleaseGroup carries the original-release date and the type fields.
type ReleaseGroup struct {
    ID               string   `json:"id"`
    Title            string   `json:"title"`
    Disambiguation   string   `json:"disambiguation"`
    FirstReleaseDate string   `json:"first-release-date"`
    PrimaryType      *string  `json:"primary-type"`
    SecondaryTypes   []string `json:"secondary-types"`
    Aliases          []Alias  `json:"aliases"`
}

// Medium is one disc.
type Medium struct {
    Position   int     `json:"position"`
    Title      string  `json:"title"`
    Format     *string `json:"format"`
    TrackCount int     `json:"track-count"`
    Pregap     *Track  `json:"pregap"`
    Tracks     []Track `json:"tracks"`
    DataTracks []Track `json:"data-tracks"`
}

// Track is one track entry on a medium. Position is a wireInt because
// the wire value is an integer in real captures and a string in every
// synthetic case, and beets casts it for the same reason
// (int(track["position"]), beetsplug/musicbrainz.py:592;
// internal/harness/synth/release.go:128 writes the string form).
type Track struct {
    ID           string       `json:"id"`
    Position     wireInt      `json:"position"`
    Number       string       `json:"number"`
    Title        string       `json:"title"`
    Length       *float64     `json:"length"`
    ArtistCredit []CreditPart `json:"artist-credit"`
    Recording    Recording    `json:"recording"`
}

// Recording is a track's underlying recording.
type Recording struct {
    ID           string       `json:"id"`
    Title        string       `json:"title"`
    Length       *float64     `json:"length"`
    Video        bool         `json:"video"`
    Aliases      []Alias      `json:"aliases"`
    ArtistCredit []CreditPart `json:"artist-credit"`
}

// wireInt is an integer the MusicBrainz WS/2 payload spells either as a
// JSON number or as a JSON string. Only track positions need it: a
// census over all 399 committed release documents found 6,763 integer
// and 297 string track positions and no other mismatched field, so this
// type is deliberately narrow rather than applied to every number.
type wireInt int

// UnmarshalJSON accepts both spellings and rejects anything else.
func (n *wireInt) UnmarshalJSON(data []byte) error

// Relation is one entry of a release's generic relations array. Release
// is nil on any relation whose target is not a release, which is most of
// them: the corpus carries 351 url, 285 artist, and 36 release
// relations.
type Relation struct {
    TargetType string          `json:"target-type"`
    Type       string          `json:"type"`
    Direction  string          `json:"direction"`
    Release    *RelationTarget `json:"release"`
}

// RelationTarget is a relation's target entity, named rather than
// anonymous so `go doc` shows it and so the caller that must test it can
// name the type (fold, m13).
type RelationTarget struct {
    ID string `json:"id"`
}

// Validate reports whether rel can be translated at all, wrapping
// ErrMalformedRelease and naming the offending element. It is the
// boundary check TranslateAlbum runs first, so the rest of this package
// may assume well-formed input: a WS response is external input, and
// go-conventions puts validation at the boundary rather than a nil check
// at every read.
func (rel *Release) Validate() error
```

```go
// internal/engine/musicbrainz/relations.go

// GroupRelations splits rels by target type, mirroring the client's own
// regrouping of a generic relations array into <target_type>_relations
// keys (beetsplug/_utils/musicbrainz.py:759-772). The sort is stable so
// within-type order survives, which Python's sorted() gives for free.
//
// The returned map's key iteration order is unobservable and callers
// must not depend on it: every beets consumer reads a single named key
// through .get, so no ordering of the map itself is part of the
// contract. Callers pass a release Validate has accepted, so a relation
// with an empty target type cannot reach here.
func GroupRelations(rels []Relation) map[string][]Relation
```

**Behavior contract:**

- **`doc.go` states the package's charter and its two exclusions**: this
  package translates already-decoded WS/2 responses into the matcher's
  vocabulary and issues no requests; the client, the query construction,
  and the limiter are 3c's; and the browse fallback
  (`_ensure_complete_recordings`, `BROWSE_MAXTRACKS = 500`, T16) is
  deliberately not ported, with the reason and the Task 6 guard named
  (R9).
- **The hyphen half is discharged by struct tags, and the doc comment
  says so** (T2). No runtime hyphen-rewriting pass exists in this package,
  and a reader looking for beets' `_normalize_data` finds the answer on
  `Release`. Rejected alternative, recorded in the same comment: decoding
  to `map[string]any`, running a faithful recursive normalizer, and
  re-decoding reproduces a normalizer for hundreds of keys nothing reads,
  loses every type, and makes the decoder testable only through goldens,
  which go-conventions calls last-resort.
- **`GroupRelations` uses `slices.SortStableFunc` with `cmp.Compare` on
  `TargetType`**, then a single group walk. Never `sort.SliceStable`.
- **The destructive-collision rule does not apply and the doc comment says
  why** (T2's third sub-rule): beets' collision arises because it writes
  regrouped keys into the same dict a pre-existing `<type>_relations` key
  may occupy. `Release` has no such field, because the only relations
  field is the generic array, so the Go port cannot collide. Stating it stops a
  later reader from "fixing" a rule that has no expression here.
- **`Validate` checks exactly five things**, each wrapping
  `ErrMalformedRelease` (R10 as folded): a relation with an empty
  `target-type` (naming the index); an empty release-level
  `artist-credit` array (naming the release id); **an empty resolved
  track credit** on any surviving track of any medium, the track's own
  when non-empty and the recording's otherwise, naming the medium
  position and the track position (fold, M-d); and **a backward
  `transl-tracklisting` relation whose `release` target is absent**,
  naming the relation index (fold, M-e). Both additions exist because the
  Go port panics rather than erring on them: `parseArtistCredits` indexes
  element zero of the resolved credit list, and `PseudoReleaseTarget`
  dereferences `Relation.Release`. Neither has a corpus instance, so both
  fixtures are hand-authored on the proximity precedent and the pass
  report says so. A medium whose `tracks`, `pregap`, and `data-tracks`
  are all absent **is legal**, not an error, because beets reads all
  three through `.get` and a medium missing all three contributes nothing
  (T5). That last one is a named non-check so a reviewer does not read
  its absence as an oversight.
- **`Tunables` gains `VAName` only.** `DefaultTunables()` sets it to
  `"Various Artists"`, and `tunables_test.go` gains one row asserting the
  default. The `AliasesAsCredits` exclusion is a doc comment, not a field
  (R8).
- **Table-driven tests, one per rule**, in this package's own `testdata`
  where a fixture is more legible than a literal. At minimum:
  `GroupRelations` over an interleaved `url`/`artist`/`url`/`artist` array
  yields keys `artist` and `url` with intra-group order preserved;
  `Validate` accepts a well-formed release; and each of the two
  `Validate` failures names its element.

**Named mutations this task reports:**

1. `GroupRelations` switched to `slices.SortFunc` (unstable) → the
   intra-group-order test must fail. If it does not, the fixture has too
   few entries per group and the test is theatre; enlarge it until the
   mutation fires and report the size that was needed.
2. `Validate`'s empty-target-type check removed → the fixture with an
   untyped relation must fail.
3. `Validate`'s release-level empty-credit check removed → the
   empty-credit fixture must fail.
3b. `Validate`'s track-level credit check removed → a fixture whose track
   and recording credits are both empty must fail, and with the check
   restored the same fixture must return `ErrMalformedRelease` rather
   than panicking (fold, M-d).
3c. `Validate`'s pseudo-target check removed → a fixture carrying a
   backward `transl-tracklisting` relation with no `release` object must
   fail, and `PseudoReleaseTarget` over the validated form must not
   panic (fold, M-e).
3d. `Position` declared `int` instead of `wireInt` → decoding
   `media-baseline`'s committed response must fail, proving the census
   claim rather than asserting it (fold, B3).
4. `DefaultTunables().VAName` changed to `""` → the tunables test must
   fail naming the field.

**Deliverable count: 4** (the package with its charter doc naming both
exclusions; the `Release` vocabulary with wire-key tags discharging the
hyphen rewrite, `wireInt`'s two-shape decode, and its rejected
alternative recorded; `GroupRelations` with its stable sort and its
unobservable-key-order contract; `Validate`'s five-rule boundary with
`ErrMalformedRelease`, plus `Tunables.VAName`).
Gate green plus tier-0 replay green; commit. **Report actual token spend;
this is the figure the 6M ceiling is re-derived from.**

---

## Task 4: Aliases and artist credits

Model: **`opus`**. The second and last deliberate upshift. The alias path
is the highest-leverage part of the layer and the least covered by the
corpus: under `languages=['ja']` it changes the album `artist` on 90 of
303 releases and track `artist` on 3,272 of 6,919 tracks, yet **only one
committed case runs under a profile that sets `import.languages` at all**,
so tier 1 will see almost none of it. Its precedence rules interlock
across three call sites and a plausible implementation is silently wrong.
Depends on Tasks 1 and 3.

**It runs before track extraction on purpose.** Both of the track-level
overrides that produce a `matcher.TrackInfo`, the title guard and the
artist credit, call into this task's functions, so landing aliases and
credits first removes a deferral rather than creating one.

**Files:**
- Create: `internal/engine/musicbrainz/alias.go` + `alias_test.go`
- Create: `internal/engine/musicbrainz/credit.go` + `credit_test.go`
- Create: `internal/engine/musicbrainz/testdata/` alias fixtures

**Interfaces:**

- Consumes: `Alias`, `CreditPart`, `Artist` (Task 3);
  `matcher.Tunables` (`Languages`, `IgnoredAliasTypes`, `VAName`).
- Produces:

```go
// internal/engine/musicbrainz/alias.go

// preferredAlias returns the first alias matching the configured
// locales, mirroring _preferred_alias (beetsplug/musicbrainz.py:128-155):
// an outer loop over tun.Languages in configured order and an inner loop
// over aliases in response order, accepting the first alias whose locale
// matches exactly (case-sensitively), whose primary flag is truthy, and
// whose lowercased type is not in tun.IgnoredAliasTypes. It returns nil
// when nothing matches, which is every case under beets' shipped empty
// languages default (Ruling D).
//
// An ignored type falls through to the NEXT matching alias rather than
// to nil; the ignored set is lowercased while the locale comparison is
// not.
func preferredAlias(aliases []Alias, tun matcher.Tunables) *Alias

// keyWithPreferredAlias returns the alias name when preferredAlias
// finds one and fallback otherwise, mirroring
// _key_with_preferred_alias (musicbrainz.py:158-162). beets calls it at
// three sites: the release title (:648), the release-group title
// (:487-489), and the recording title (:443). This port has two
// callers, because release_group_title is not in matcher.AlbumInfo and
// is rung 8's applied-field territory.
func keyWithPreferredAlias(aliases []Alias, fallback string, tun matcher.Tunables) string
```

```go
// internal/engine/musicbrainz/credit.go

// credits carries the artist strings _parse_artist_credits produces
// that this port needs, and no others. Artist reaches
// matcher.AlbumInfo and matcher.TrackInfo; ArtistID is here because the
// various-artists override is keyed on the first artist's id
// (musicbrainz.py:673-681). artist_sort is deliberately NOT produced:
// beets computes it at musicbrainz.py:337, but R6 puts it in rung 8's
// applied-field territory and out of the ruling-E ledger's graded set,
// so carrying it here would be tested dead data of exactly the shape
// 3a's close had to remove (fold, M-i).
type credits struct {
    Artist   string
    ArtistID string
}

// parseArtistCredits flattens an artist-credit array, mirroring
// _parse_artist_credits (beetsplug/musicbrainz.py:301-357). The
// joinphrase is appended after the LAST element too, so a malformed
// credit whose final joinphrase is non-empty leaks a trailing separator
// straight into the weight-3.0 album artist penalty; that is beets'
// behavior and the port reproduces it.
//
// Callers pass a credit list Validate has accepted, at the release level
// and at every surviving track's resolved level both, so the empty-list
// case that raises IndexError at musicbrainz.py:350 cannot reach here.
// Revision 1 claimed that guarantee against a Validate that checked the
// release level only, which left this function indexing element zero of
// an empty slice on track and recording credits (fold, M-d).
func parseArtistCredits(parts []CreditPart, tun matcher.Tunables) credits

// applyVariousArtists overwrites c's artist strings with tun.VAName
// when c names the various-artists entity, mirroring
// musicbrainz.py:673-681, and reports whether it fired. The report is
// AlbumInfo.VA, which the album artist component reads to skip itself
// entirely (matcher/album_artist.go:13-15).
func applyVariousArtists(c credits, tun matcher.Tunables) (credits, bool)

// variousArtistsID is MusicBrainz's own special-purpose artist entity
// for various-artists releases (beetsplug/musicbrainz.py:43).
const variousArtistsID = "89ad4ac3-39f7-470e-963a-56509c546377"
```

**Behavior contract:**

- **`preferredAlias` reproduces T7's table row for row**, and the test is
  that table verbatim, eight rows, each asserting the selected alias by
  name. The locale comparison is `==` on the raw strings; the ignored-type
  comparison lowercases the configured set once and lowercases each
  alias's type at comparison, treating a null type as `""`.
- **`Primary` is `*bool` and a null is falsy** (`_utils/musicbrainz.py:138-155`),
  so the guard is `a.Primary != nil && *a.Primary`. A test row carries a
  null primary and asserts it is not selected.
- **The explicit-`languages`-parameter branch is not ported**, and the doc
  comment says why: no call site in the plugin ever passes it (T7), so it
  is dead code in the port's scope. Naming it stops a later reader adding
  a parameter nothing uses.
- **`parseArtistCredits` reproduces T8's trailing-joinphrase behavior**,
  with three named test rows: joins `[" & ", ""]` → `"A & B"`; a single
  credit with join `" feat. "` → `"A feat. "`, trailing space asserted;
  joins `[" & ", " & "]` → `"A & B & "`.
- **The alias supplies the entity name; the credit string uses the
  credited name.** `artist_object = alias or el["artist"]` and
  `credit_artist_object = el` under the shipped
  `aliases_as_credits: false`. The A/B table from T8 is the test, with
  its `SortName` column dropped because `credits` no longer carries one
  (fold, M-i):

| `Languages` | `Artist` | credit string |
| --- | --- | --- |
| `[]` | `RealName` | `CreditedName` |
| `["ja"]` | `AliasName` | `CreditedName` |

  The second row is the one that proves the split: the alias moves
  `Artist` and does not move the credit. (`aliases_as_credits` is not a
  field, per R8, so the two `true` rows of the inventory's own four-row
  table are not portable, and their absence is noted in the test's doc
  comment.)
- **`applyVariousArtists` fires on `artist_id == variousArtistsID`
  only**, and both branches get a test, one per `VAName` value, asserting
  that the returned flag is what `AlbumInfo.VA` will carry.
- **The alias path's coverage gap is declared, not hidden** (R6, R7).
  Under the shipped empty `Languages` default, `preferredAlias` returns nil
  on 214 of 215 committed cases, so **the corpus grades this path through
  one case**. The mitigation is stated in the pass report as a declared
  coverage limit: `real-nonlatin-utada-first-love-locale` is the only
  ledger subject running under `locale.yaml`, its A-half twin
  `real-nonlatin-utada-first-love` is a subject so the pair is gradeable
  in both directions, and everything else about the path is graded by this
  task's hand-authored tables.
- **The second consequence of the empty default is easy to miss and gets
  its own test** (inventory §(b).7): the track-title override at
  `musicbrainz.py:589-590` is guarded by `not
  preferredAlias(recording.Aliases)`, so on 214 of 215 cases the guard is
  unconditionally true and `recording.Title = track.Title` fires whenever
  the track carries a title. This task asserts both sides of that guard at
  the `preferredAlias` level; Task 5 asserts it through the override
  itself.

**Named mutations this task reports:**

1. The locale comparison lowercased → T7's `['JA']` row must fail.
2. The ignored-type fall-through changed to return nil → T7's `['ja']` /
   `['artist name']` row must fail (it expects the **next** ja primary, not
   nil).
3. The two loops transposed (aliases outermost) → T7's `['en','ja']` row
   must fail, because response order would beat configured order.
4. The final joinphrase dropped from the join → the `" & "`-terminated row
   must fail.
5. `credit_artist_object` switched to the alias → the `["ja"]` row's
   credit assertion must fail. If no assertion covers it, the credit
   string is untested and the task adds one.
6. `applyVariousArtists` keyed on the artist **name** rather than the id →
   a fixture whose artist is named "Various Artists" but carries an
   ordinary id must show the flag false.

**Deliverable count: 4** (`preferredAlias` at T7's eight-row table with
the null-primary and fall-through rules; `keyWithPreferredAlias` with its
two Go callers and the third beets call site named as out of scope;
`parseArtistCredits` with the trailing-joinphrase behavior, the
alias-versus-credit split, and `artist_sort` named as deliberately not
produced; `applyVariousArtists` with `variousArtistsID`,
both branches, and the declared alias coverage gap). Gate green plus
tier-0 replay green; commit. **Re-derive the ceiling against Task 3's
figure and report both.**

---

## Task 5: Per-medium track extraction and the media derivation

Model: `sonnet`. Depends on Tasks 1, 3, and 4.

**Files:**
- Create: `internal/engine/musicbrainz/tracks.go` + `tracks_test.go`
- Create: `internal/engine/musicbrainz/testdata/` fixtures for the three
  declared gaps and the sentinel

**Interfaces:**

- Consumes: `Release`, `Medium`, `Track`, `Recording` (Task 3);
  `preferredAlias`, `parseArtistCredits`, `keyWithPreferredAlias` (Task
  4); `matcher.TrackInfo`, `matcher.Tunables`
  (`internal/engine/matcher/info.go:73-105`, `tunables.go:42-197`).
- Produces:

```go
// internal/engine/musicbrainz/tracks.go

// builtTrack is the matcher's TrackInfo plus the one field the
// translation needs and the vocabulary deliberately does not carry:
// the medium's own format string, from which AlbumInfo.Media is derived
// (beetsplug/musicbrainz.py:654-658). TrackInfo declares itself "the
// eight fields the track-level distance path reads ... and nothing
// else" (matcher/info.go:69-72), and per-track media is never read by
// that path, so it stays here rather than widening the vocabulary.
type builtTrack struct {
    Info   matcher.TrackInfo
    Format *string
}

// tracksFromMedium yields one builtTrack per surviving track on m,
// mirroring get_tracks_from_medium (beetsplug/musicbrainz.py:555-600):
// pregap first and unconditionally, then tracks, then data tracks only
// when tun.IgnoreDataTracks is false; then the two filters, the
// "[data track]" title drop that runs regardless of IgnoreDataTracks
// and the video drop under tun.IgnoreVideoTracks; then the three
// track-level overrides. Index is left zero here and assigned by the
// caller across all valid media.
func tracksFromMedium(m Medium, tun matcher.Tunables) []builtTrack

// extractTracks runs tracksFromMedium over every medium of rel whose
// format is not in tun.IgnoredMedia, then assigns Index 1-based across
// all surviving tracks in medium order (musicbrainz.py:645-646).
//
// It does not compute AlbumInfo.Mediums: that is len(rel.Media), the
// UNFILTERED count (musicbrainz.py:659), and it is deliberately not
// derived here, because a reader who saw both numbers produced by one
// function would reasonably assume they agree.
func extractTracks(rel *Release, tun matcher.Tunables) []builtTrack

// deriveMedia returns AlbumInfo.Media for the given built tracks: the
// single format when exactly one distinct format survives, and
// otherwise the literal "Media" (musicbrainz.py:654-658). Zero tracks
// and mixed formats both land on the sentinel, which is truthy and
// which the media distance component therefore scores as if it were a
// media name (autotag/distance.py:455).
func deriveMedia(tracks []builtTrack) string
```

**Behavior contract:**

- **Assembly order and both filters are T5's, exactly.** The `[data
  track]` title drop is compared against the literal string, no strip, no
  case fold, and it runs whether or not data tracks were admitted.
- **The three track-level overrides are T6's, exactly**, including the
  double `preferredAlias` read: `recording.Title` is replaced by
  `track.Title` **only if** `track.Title` is non-empty **and**
  `preferredAlias(recording.Aliases, tun)` returns nothing; the resulting
  `TrackInfo.Title` is then
  `keyWithPreferredAlias(recording.Aliases, recording.Title, tun)`, which
  is the second read. **The `or` fallbacks are truthiness, not
  nil-ness**: a track length of `0` falls back to the recording length
  just as a `null` does, and an empty artist-credit slice falls back just
  as a nil one does. A Go port testing `!= nil` is wrong, and the test
  table carries `0`, `null`, and a real value as three separate rows.
- **`TrackInfo.Artist` comes from `parseArtistCredits` over the resolved
  credit list**, the track's own when non-empty and the recording's
  otherwise (T6's second override). That is the same function the album
  artist goes through, and it is why this task depends on Task 4.
- **`TrackInfo.Length` is seconds**, converted from the wire's
  milliseconds, matching `hooks.py:481`'s own units; a fixture with a
  200000 ms recording asserts `200.0`.
- **A missing `aliases` key is an empty alias list, not an error** (R10's
  third row), with the reason in the doc comment.
- **`IgnoredMedia` membership is exact and case-sensitive** (T14), matched
  against the medium's `format` string. A medium with a null format is
  never ignored, because `None` is not in any configured list.
- **Three declared corpus-coverage gaps get hand-authored fixtures**
  (R15). Pregap: a medium carrying a pregap plus one ordinary track, under
  both `IgnoreDataTracks` values, asserting the pregap survives both
  (unconditional). Data tracks: T5's four-row truth table, hand-authored,
  covering both booleans in both states. The `[data track]` title filter:
  a track carrying that title inside `medium.Tracks` (not
  `medium.DataTracks`) with `IgnoreDataTracks` false, asserting it is
  still dropped.
- **The `"Media"` sentinel gets three fixtures, two routes and one
  non-route** (R3 as folded, B5): zero surviving tracks (a single-medium
  release whose only track is a dropped video) and two distinct formats
  across two media both yield `"Media"`; a **single** medium whose
  `format` is `null` yields `""`, the R1 zero-value mapping of beets'
  `None`, which is falsy and declines the media penalty at
  `distance.py:455`. A fourth row is worth carrying because it is the
  route a reader least expects: `[null, "CD"]` across two media **is** a
  sentinel route, because `None` counts as one of the distinct values.
  `deriveMedia`'s `string` return cannot express `None` and does not need
  to.
- **`medium_index` decodes both wire spellings** (fold, B3): the test
  table carries a JSON string position beside a JSON number one, because
  297 committed track positions across 76 case directories are strings
  and `media-baseline`, `media-uneven-split`, and `va-exact` are all
  ledger subjects that carry them.

**Named mutations this task reports:**

1. The pregap moved behind the `IgnoreDataTracks` gate → the pregap
   fixture must fail under `IgnoreDataTracks: true`.
2. The `[data track]` title filter gated on `IgnoreDataTracks` → the title
   fixture must fail.
3. `track.Length` fallback changed from truthiness to `!= nil` → the
   `length: 0` row must fail (this is the mutation the plan expects to be
   hardest to notice by reading).
4. The title override's `preferredAlias` guard dropped → T6's table's
   first row must fail, because an alias-carrying recording would take the
   track title.
5. `deriveMedia` returning `""` instead of `"Media"` on zero tracks → the
   zero-track fixture must fail.
5b. `deriveMedia` returning `"Media"` on a single null-format medium →
   the non-route fixture must fail, which is the direction revision 1 had
   backwards (fold, B5).
6. `extractTracks` assigning `Index` per medium instead of across all
   media → the `media-uneven-split` shape fixture must fail on the second
   medium's first track.
7. `IgnoredMedia` matched case-insensitively → a fixture with
   `IgnoredMedia: ["cd"]` and a `"CD"` medium must show the medium
   surviving.

**Deliverable count: 4** (`tracksFromMedium` at T5's assembly order and
both filters; `extractTracks` with the ignored-media filter, the
across-media index assignment, and the stated non-derivation of
`Mediums`; the three track-level overrides at T6's truthiness semantics,
including the double alias read, the artist credit, and the two-spelling
position decode; `deriveMedia` with both routes to the sentinel and the
single-null-format non-route, plus the three declared-gap fixtures).
Gate green plus tier-0 replay green; commit.

---

## Task 6: Release events, dates, and `TranslateAlbum`'s composition

Model: `sonnet`. Depends on Tasks 1, 3, 4, and 5.

**Files:**
- Create: `internal/engine/musicbrainz/events.go` + `events_test.go`
- Create: `internal/engine/musicbrainz/translate.go` + `translate_test.go`
- Create: `internal/engine/musicbrainz/testdata/` fixtures for the
  preferred-country gap and the date table

**Interfaces:**

- Consumes: everything Tasks 3, 4, and 5 produce; `matcher.AlbumInfo`,
  `matcher.Tunables`.
- Produces:

```go
// internal/engine/musicbrainz/events.go

// preferredReleaseEvent returns the country and date the translation
// writes, mirroring _preferred_release_event
// (beetsplug/musicbrainz.py:165-177): an outer loop over
// tun.Preferred.Countries in configured order and an inner loop over
// rel.ReleaseEvents, returning the first configured country any event's
// area lists, together with that event's date. Otherwise it returns the
// release's own top-level country and date.
//
// The membership test is a literal, case-sensitive slices.Contains
// against the area's ISO codes, and the country returned is the
// CONFIGURED string rather than the area's own code. The same
// Tunables.Preferred.Countries field is compiled as an anchored,
// case-insensitive regex list by the country distance component
// (matcher/album_country.go:31-40), so one config list carries two
// matching regimes in two packages. Reusing that compiler here would be
// silently wrong on any pattern-shaped entry.
func preferredReleaseEvent(rel *Release, tun matcher.Tunables) (country string, date string)

// parseDate splits a MusicBrainz date on "-" and returns the parts it
// can parse, mirroring _get_date (beetsplug/musicbrainz.py:180-200): an
// unparsable part becomes zero without failing the whole parse, and
// extra parts beyond the third are discarded silently.
func parseDate(s string) (year, month, day int)
```

```go
// internal/engine/musicbrainz/translate.go

// TranslateAlbum builds the matcher's candidate vocabulary from one
// decoded WS/2 release, mirroring MusicBrainzPlugin.album_info
// (beetsplug/musicbrainz.py:631-695). It issues no request and reads no
// clock: the evaluation date is the distance path's input, not this
// layer's (3a contract 9), and internal/harness/conform's widened
// no-time.Now walk is what enforces that.
//
// It returns an error wrapping ErrMalformedRelease when rel cannot be
// translated at all; every other beets-raising site is handled as a
// zero value and the reasons are recorded on Release and on the
// functions concerned.
func TranslateAlbum(rel *Release, tun matcher.Tunables) (*matcher.AlbumInfo, error)
```

**Behavior contract:**

- **`preferredReleaseEvent` is T10's exactly**, and its test is T10's
  measured table: `["GB"]` → `("GB", "1999-01-02")`; `["gb"]` → the
  fallback `("XW", "1980-05-06")`; `["US","GB"]` → `("US", "2001-03-04")`.
  A null `area` is skipped rather than failing. A release with neither
  `release-events` nor top-level country and date yields `("", "")`.
- **The preferred-country branch is a declared corpus-coverage gap**
  (R15 as folded), graded by this task's hand-authored fixtures, with the
  mutation named. Its unreachedness is measured, not assumed, and the
  measurement is not the one revision 1 carried (fold, B2): **290 of the
  303 release-lookup responses' 298 top-level release events do carry a
  non-null `area` with `iso-3166-1-codes`, across 208 case directories**.
  The branch is unreached only because no committed profile sets a
  `match:` key, and it would be reachable on 208 case directories the
  moment one set `match.preferred.countries`.
- **`parseDate` is T11's exactly**, and its test is T11's measured table,
  six rows: `""` → `(0,0,0)`; `"1999"` → `(1999,0,0)`; `"1999-05"` →
  `(1999,5,0)`; `"1999-XX-06"` → `(1999,0,6)`, **a zero month with a real
  day**, which is the row a "stop at the first failure" implementation
  gets wrong; `"199x"` → `(0,0,0)`; `"1999-05-06-07"` → `(1999,5,6)`.
  Beets' `None` maps to Go's `0` here, which is the R1 convention applied
  to the three int fields; `AlbumInfo.Year` and `.OriginalYear` are both
  truthiness-guarded at their consumers
  (`autotag/distance.py:472,480,475,484`), so the mapping is faithful.
- **`TranslateAlbum` composes in `album_info`'s own order** (T1's step
  table), and the order is load-bearing in exactly one place the doc
  comment names: the release-event assignment overwrites `country`,
  `year`, `month` and `day` **after** the release-group parse supplies the
  original-date fallback (`musicbrainz.py:684-689`), and the fallback tests
  **truthiness** of the resolved date, not nil-ness (T11). The composition:
  1. `rel.Validate()`, returning early on failure.
  2. `tracks := extractTracks(rel, tun)`; `media := deriveMedia(tracks)`.
  3. `mediums := len(rel.Media)`, **the unfiltered count** (T3).
  4. `album := keyWithPreferredAlias(rel.Aliases, rel.Title, tun)`.
  5. `c := parseArtistCredits(rel.ArtistCredit, tun)`; then
     `c, va := applyVariousArtists(c, tun)`.
  6. Scalars: `AlbumID` from `rel.ID`; `AlbumDisambig` from
     `rel.Disambiguation`; `DataSource` the literal `"MusicBrainz"`
     (`beets/metadata_plugins.py:182-188`, `cls.__name__.replace("Plugin","")`).
  7. Release group: `OriginalYear` from `parseDate(rel.ReleaseGroup.FirstReleaseDate)`.
  8. Labels, and the two fields are guarded differently (fold, m20).
     `CatalogNum` is `rel.LabelInfo[0].CatalogNumber` read
     **unconditionally**, a null yielding the empty string; `Label` is
     read only when the label object is present and its name differs from
     the exact literal `"[no label]"` (`$SP/beetsplug/musicbrainz.py:497-506`,
     case-sensitive, no strip). An absent or empty `label-info` yields
     both empty. The corpus carries 32 null catalog numbers and 3 null
     labels, and a fixture row carries a `"[no label]"` label beside a
     real catalog number so the asymmetry is asserted rather than
     assumed.
  9. Release event: `country, date := preferredReleaseEvent(rel, tun)`;
     `Year/Month/Day` from `parseDate(date)` **when `date` is non-empty**,
     falling back to the release group's original date otherwise.
  10. `Tracks` from the built tracks, dropping the per-track format.
- **`AlbumInfo.Media` is `deriveMedia`'s result**, so a zero-track release
  carries the `"Media"` sentinel and the media component fires on it.
- **The browse-fallback guard lands here** (R9): a gate-resident test in
  `internal/harness/conform`, **not in this package**, because it reads
  the corpus. It asserts that no committed case's release-lookup response
  carries `sum(len(medium.tracks)) > 500`, failing and naming the case,
  the release, and the count. It is stated in this task because it is
  T16's obligation, and it is **implemented in Task 8** with the rest of
  the conform work. This task's own obligation is the recorded refusal in
  `doc.go` (Task 3) and a `translate.go` doc-comment cross-reference to
  the guard's test name.
- **Whole-release tests, hand-authored**, exercising the composition end
  to end on small synthetic releases: a plain single-medium release; a
  two-medium release proving `mediums` counts both while `IgnoredMedia`
  drops one medium's tracks (T3's measured pair); a VA release proving the
  `VAName` overwrite and `VA: true`; and a release with no
  `release-events` and no top-level date, proving the original-date
  fallback. **These are not corpus cases and this package reads no
  `corpus/` path**; whole-case agreement against the ledger is Task 8's.

**Named mutations this task reports:**

1. `preferredReleaseEvent` compiled through
   `matcher.CompilePatternOption` → T10's `["gb"]` row must fail, because
   a case-insensitive regex would match.
2. `preferredReleaseEvent` returning the area's own code instead of the
   configured string → a fixture whose configured entry and area code
   differ in spelling must fail.
3. The two loops transposed → T10's `["US","GB"]` row must fail.
4. `parseDate` returning early on the first unparsable part → the
   `"1999-XX-06"` row must fail.
5. `mediums` computed from the filtered media → T3's measured pair must
   fail (`ignored_media=['CD']` must still report `mediums=2`).
6. The release-event assignment moved before the release-group parse → the
   no-events fixture must fail, because the original-date fallback would
   be overwritten by empties.

**Deliverable count: 4** (`preferredReleaseEvent` at T10's exact
membership semantics with the two-regimes doc comment and the declared
preferred-country gap; `parseDate` at T11's six-row table; `TranslateAlbum`'s
ten-step composition in `album_info`'s own order; the four hand-authored
whole-release tests). Gate green plus tier-0 replay green; commit.

---

## Task 7: The pseudo-release merge

Model: `sonnet`. Depends on Tasks 1, 3, and 6.

**Files:**
- Create: `internal/engine/musicbrainz/pseudo.go` + `pseudo_test.go`
- Create: `internal/engine/musicbrainz/testdata/` fixtures for the
  merge's corpus-unsignalled fields

**Interfaces:**

- Consumes: `Release`, `Relation`, `GroupRelations` (Task 3);
  `matcher.AlbumInfo`.
- Produces:

```go
// internal/engine/musicbrainz/pseudo.go

// PseudoReleaseTarget reports the release id whose AlbumInfo must be
// merged into rel's, and whether rel is a pseudo-release at all,
// mirroring album_for_id's own check (beetsplug/musicbrainz.py:785-798):
// rel.Status is exactly "Pseudo-Release" and some release relation
// carries type "transl-tracklisting" with direction "backward". The
// relation key exists only because GroupRelations regrouped it, so this
// predicate depends on that pass having run.
//
// The FETCH of that release is 3c's client work; 3b exports the
// predicate so 3c calls it, and merges over two already-built AlbumInfo
// values.
func PseudoReleaseTarget(rel *Release) (releaseID string, ok bool)

// MergePseudoRelease folds an actual release's own fields into a
// pseudo-release candidate, mirroring _merge_pseudo_and_actual_album
// (beetsplug/musicbrainz.py:203-240): the pseudo supplies identity and
// titles, the actual supplies the physical and date fields. It returns a
// new value; beets starts from pseudo.copy(), which is a deepcopy
// (autotag/hooks.py:93-95), and nothing here mutates either argument.
//
// Beets moves fifteen keys. Seven of them are in matcher.AlbumInfo and
// are the merge's entire verdict-moving effect: Media, Mediums,
// Country, CatalogNum, Year, OriginalYear, Label. The other eight
// (month, day, original_month, original_day, barcode, asin, style,
// genres) are outside the vocabulary. Album, Artist, VA, AlbumDisambig,
// AlbumID, DataSource and Tracks stay the pseudo's.
//
// Both arguments are total values, so the KeyError beets can raise on
// actual[k] (musicbrainz.py:220) has no expression here.
func MergePseudoRelease(pseudo, actual *matcher.AlbumInfo) *matcher.AlbumInfo
```

**Behavior contract:**

- **`PseudoReleaseTarget` reads the regrouped relations**, through
  `GroupRelations(rel.Relations)["release"]`, and matches the **first**
  entry with both `type == "transl-tracklisting"` and `direction ==
  "backward"`. Both comparisons are exact, and `Status ==
  "Pseudo-Release"` is exact (T12). A release satisfying the status but
  carrying no such relation returns `ok == false`, and so does the
  reverse; both get a test row.
- **`MergePseudoRelease` copies exactly the seven vocabulary fields**, and
  the doc comment lists them, because that list is the function's entire
  contract and a reader checking it against beets' fifteen needs the map.
- **The merge's `Year`/`OriginalYear`/`Mediums` path has no corpus signal
  at all** (T12), because those fields were already equal on all four
  committed pseudo cases, so those three are graded **only** by hand-authored
  fixtures here, and the pass report names it as a declared gap in the
  same shape as R15's three. A fixture whose pseudo and actual differ in
  all seven fields, asserting all seven move and that the other seven
  vocabulary fields plus `Tracks` do not, is the one test that catches a
  transcription slip in either direction.
- **The pseudo's `Tracks` survive**, and a fixture with different track
  counts on the two sides asserts it. On the committed corpus the counts
  are equal on all four cases, so this too is hand-authored rather than
  corpus-graded.
- **Detection stays out of the merge.** `MergePseudoRelease` never inspects
  a `Release`; it takes two built `AlbumInfo` values, which is what makes
  it testable without a client and what lets 3c call the predicate and the
  merge separately.

**Named mutations this task reports:**

1. `direction == "forward"` accepted → the forward-relation fixture must
   fail.
2. `Status` compared case-insensitively → a `"pseudo-release"` fixture
   must show `ok == false`.
3. Any one of the seven merged fields dropped → the all-seven-differ
   fixture must fail naming that field.
4. `AlbumID` added to the merged set → the same fixture must fail, proving
   the negative half.
5. `Tracks` taken from the actual → the differing-track-count fixture must
   fail.

**Deliverable count: 3** (`PseudoReleaseTarget` with its exact status and
relation predicate, both negative branches tested; `MergePseudoRelease`
with the seven-field contract and the eight-field negative, proved in both
directions; the hand-authored fixtures for the year, mediums, and track
paths the corpus cannot signal, declared as a gap). Gate green plus tier-0
replay green; commit.

---

## Task 8: The translation tier, grading the Go translation against the ledger

Model: `sonnet`. **Runner-up upshift candidate**: if Task 4 or Task 6
reports that a translation contract above was under-specified, the
conductor upshifts this task, because this is the instrument every claim
about the translation layer rests on. Depends on Tasks 1 through 7.

**Files:**
- Create: `internal/harness/conform/translate.go` + `translate_test.go`
  (the comparison, the nullability policy, the profile-name overlay and
  its key-name drift guard, and the browse-fallback guard)
- Create: `internal/harness/conform/testdata/translate/` fixtures
- Create: `internal/harness/conform/testdata/arch/time-now-nested/` fixture
- Modify: `internal/harness/conform/arch_test.go` (the widened clock walk,
  its widened error text and doc comment, and its nested firing fixture)
- Modify: `internal/harness/conform/defaults_test.go` (the pattern-table
  name resolution guard, m11)

**Interfaces:**

- Consumes: `analysis.LoadAlbumInfoLedger`, `analysis.AlbumInfoLedger`,
  `analysis.LedgerAlbumInfo`, `analysis.LedgerTrackInfo` (Task 1);
  `musicbrainz.TranslateAlbum`, `musicbrainz.MergePseudoRelease`,
  `musicbrainz.PseudoReleaseTarget`, `musicbrainz.Release` (Tasks 3, 6,
  7); `matcher.AlbumInfo`, `matcher.DefaultTunables`;
  `cassette.LoadCase`, `cassette.Case`.
- Produces:

```go
// internal/harness/conform/translate.go

// CheckTranslatedAlbum compares one ledger row's recorded AlbumInfo
// against the AlbumInfo the Go translation built, field by field, and
// reports every disagreement rather than the first. caseID and mbid
// name the row in every message so a failure identifies the release
// without a second lookup.
func CheckTranslatedAlbum(caseID, mbid string, want analysis.LedgerAlbumInfo, got *matcher.AlbumInfo) error

// nullPolicy says what a JSON null in the ledger means for one field.
// It exists because beets' own consumers do not treat them alike: three
// add_string call sites on the whole distance path are unguarded by a
// truthiness test on the info side, so only those three can turn a None
// into arithmetic, and for every other field None, "" and 0 are
// indistinguishable at the consumer.
type nullPolicy int

const (
    // nullStrict: a null here is a gate failure. These are the three
    // unguarded add_string sites (autotag/distance.py:451, :448, :402),
    // the only shapes in which a real divergence could arrive. The
    // MusicBrainz translation cannot produce one today, and this is
    // the check that would catch a source that could.
    nullStrict nullPolicy = iota

    // nullNeverExpected: the derivation cannot produce a null here, so
    // one means the derivation changed rather than that the value is
    // absent.
    nullNeverExpected

    // nullEqualsZero: a null compares equal to the Go zero value,
    // because beets' own truthiness guard at the consumer makes None,
    // "" and 0 indistinguishable.
    nullEqualsZero
)
```

**Behavior contract:**

- **The runner reads every committed ledger row, rebuilds each release's
  `AlbumInfo` through the Go translation, and compares.** For each row it
  resolves the case through `cassette.LoadCase`, finds the request whose
  `path_query` is `/ws/2/release/<mbid>...`, reads that request's
  `response` file, decodes it into `musicbrainz.Release`, and calls
  `TranslateAlbum` with `matcher.DefaultTunables()` merged with the case's
  own profile. **No stub, no client, no network**: the stub is 3c's
  business, and reading a committed response file from the harness is
  exactly the permitted direction.
- **The profile merge is `Tunables`' defaults merged with the case's
  profile, in that order**, over the three committed profiles, which is
  the same rule the spec's M1 fold states for tier 1 (spec `:277`).
  **It is a profile-name-to-overlay table, never a YAML parse** (fold,
  B4). `go.mod` carries cobra and two indirects, no Go file in the tree
  imports a YAML package, and this plan's own Global Constraints say a
  task that wants a dependency stops and reports. `cassette.Case.Profile`
  (`internal/harness/cassette/cassette.go:44`) is a profile **name**, so
  the decoder is three names against one overlay: `stock` and `estate`
  overlay nothing the translation reads, and `locale` sets
  `Languages: ["ja"]` and `IgnoredAliasTypes: []`. **An unrecognized
  profile name is the failure.**
- **The unknown-key rule survives as a text drift guard, and this is the
  fold's second adaptation** (B4). Dropping it with the YAML parse would
  let a future `match:` key land in a profile without ever reaching the
  translation, which is the silent-stop the rule existed to catch.
  Revision 1's version could not be built and could not have passed
  anyway: the three profiles differ in far more than two keys (`estate`
  sets `directory`, `library`, `import.quiet_fallback`,
  `import.none_rec_action`, `import.timid`, `original_date`,
  `per_disc_numbering`, `plugins`; `locale` is estate plus two), and five
  of those have no `Tunables` home by design, two of them because 3a's M2
  resolution deliberately keeps them out. So the guard is a line-oriented
  **key-name** scan over each committed profile's own text, in
  `countMetadataSources`' shape
  (`internal/harness/oracle/oracle.go:254-268`), failing and naming any
  top-level or `import:` key the written allowlist does not carry. The
  allowlist names all ten keys the three profiles use today, each with
  either its `Tunables` field or the reason it has none.
- **The nullability policy is a table in this file, one line of reason per
  field** (R1, R11), so a reader sees why `media` tolerates null and
  `album` does not without reading beets:

| Field | Policy | Reason |
| --- | --- | --- |
| `album.album` | `nullStrict` | `distance.py:451`, unguarded `add_string` |
| `album.artist` | `nullStrict` | `distance.py:448`, guarded only by `not va` |
| track `title` | `nullStrict` | `distance.py:402`, unguarded `add_string` |
| `album.va` | `nullNeverExpected` | a Python bool, never None |
| `album.mediums` | `nullNeverExpected` | `len(...)`, may be 0, never None |
| `album.album_id` | `nullNeverExpected` | `release["id"]`, always present |
| `album.data_source` | `nullNeverExpected` | the literal `"MusicBrainz"` |
| track `index`, `medium_index`, `medium`, `track_id`, `data_source` | `nullNeverExpected` | all written unconditionally by `track_info` |
| `album.media` | `nullEqualsZero` | truthiness-guarded, `distance.py:455`; `medium["format"]` is `str | None` (`_utils/musicbrainz.py:468`), so a single-medium null format really does yield None |
| `album.year`, `original_year` | `nullEqualsZero` | truthiness-guarded, `distance.py:472,480,475,484` |
| `album.country` | `nullEqualsZero` | truthiness-guarded, `distance.py:498,501` |
| `album.label`, `catalognum`, `albumdisambig` | `nullEqualsZero` | truthiness-guarded, `distance.py:505,509,515` |
| track `length` | `nullEqualsZero` | truthiness-guarded, `distance.py:397` |
| track `artist` | `nullEqualsZero` | truthiness-guarded, `distance.py:404-410` |

  Twenty-one fields, three strict, nine never-expected, nine
  zero-equivalent. **The row count is asserted against
  `reflect.TypeOf(matcher.AlbumInfo{}).NumField() - 1 +
  reflect.TypeOf(matcher.TrackInfo{}).NumField()`**, with the `- 1`
  named in a comment as `AlbumInfo.Tracks`, which is compared as a list
  rather than through the policy. `AlbumInfo` declares 14 fields and
  `TrackInfo` 8, so revision 1's assertion compared 21 against 22 and
  failed as written (fold, M-c). Adding a fifteenth vocabulary field
  without extending the policy still fails.
- **Track lists compare element by element and length first**, naming the
  index and the field on any disagreement. A length mismatch names both
  counts and stops that row's track comparison, because a shifted list
  would otherwise report every track as wrong.
- **The pseudo rows grade the merge too.** For a row carrying `Merged`,
  the runner additionally builds the actual release's `AlbumInfo`, calls
  `MergePseudoRelease`, and compares against the `Merged` block. That is
  the only place `MergePseudoRelease` is graded against beets rather than
  against a hand-authored fixture.
- **`errors.Join` over per-field errors**, matching `CheckCaseWeights`'
  shape (`internal/harness/conform/defaults.go:21-37`). Every
  disagreement is reported, never just the first.
- **A skipped row is a failure.** The runner reports the count it actually
  compared and the test fails if that count is not the ledger's row count.
  That is the same "instrument that can pass silently" mitigation tier 0
  carries (3a plan, Task 6).
- **The browse-fallback guard lands here** (R9, T16):
  `TestNoCommittedReleaseReachesTheBrowseThreshold` reads every committed
  case's release-lookup responses, sums `len(medium.tracks)` across media,
  and fails naming the case, the release, and the count when any reaches
  `BROWSE_MAXTRACKS = 500`. The threshold is a named constant in this file
  with its citation. **This is the one check in the pass that reads
  outside the nineteen ledger subjects**, and it does so because a corpus
  case crossing the line is exactly the event nobody would otherwise
  notice.
- **The clock walk widens** (R21): `TestMatcherNeverReadsTheClock`
  (`arch_test.go:147-153`) changes its root from
  `internal/engine/matcher` to `internal/engine`, and is renamed
  `TestEngineNeverReadsTheClock` so the name states what it now checks.
  **`checkNoTimeNow`'s error text (`arch_test.go:107`) and its doc
  comment (`:70-71`) widen with it** (fold, m8): both name
  `internal/engine/matcher` today, and a widened check reporting the old
  package name sends the next reader to the wrong tree.
  Its existing firing fixture stays; a **new** nested fixture
  (`testdata/arch/time-now-nested/musicbrainz/events.go`) proves the walk
  descends into a subdirectory rather than only scanning the root, with
  `checkNoTimeNow` given `testdata/arch/time-now-nested` as its root.
- **The two `pattern_table` names get a resolution guard** (fold, m11).
  `DefectFree.PatternTable`'s two legal values are the Go identifier
  names `"SDPatternsWordBoundary"` and `"SDPatternsNoBracketDiscount"`,
  asserted as string literals inside `internal/harness/analysis`, which
  must not import `internal/engine`. Nothing checks they still resolve,
  so a rename in `matcher` would orphan the ledger silently. A
  `conform`-side test asserts both names resolve to a non-empty
  `[]matcher.SDPattern`, in `CheckUnidecodeVersion`'s shape
  (`conform/defaults.go:67-72`), which is the same
  harness-reads-both-sides pattern the unidecode pin already uses.
- **Wall time is reported broken down**, with `-count=1`. The runner
  decodes 50 response files, some of them large, so if this leg plus tier
  0 plus the 215-case defaults read outgrows the module gate's tolerance,
  that is a finding for the conductor: it becomes its own leg of
  `scripts/check.sh` at full scope, never an optional target.

**Named mutations this task reports:**

1. A hand-authored ledger fixture with one row's `"album": null` → the
   strict half fires naming the case and the field.
2. A hand-authored fixture with one row's `"media": null` against a Go
   `""` → the lenient half does **not** fire (a negative assertion, and
   the one that proves the policy is not uniformly strict).
3. A fixture with one row's `"va": null` → the never-expected half fires.
4. A fixture with one row's `album.label` perturbed by one character →
   `CheckTranslatedAlbum` fails naming `label` and both values.
5. A fixture with one row's track list one element short → the length
   mismatch fires naming both counts.
6. The row counter hardcoded to the ledger length instead of counting
   comparisons → a fixture whose row is skipped must fail.
7. `BROWSE_MAXTRACKS` lowered to 400 in a test-local copy → the guard
   fires naming **both** `real-boxset-rachmaninov-complete` and
   `real-boxset-rachmaninov-complete-stock`, each at 470, proving the
   guard can see the real corpus rather than passing vacuously. Revision
   1 named only the stock twin; both sum to 470 (fold, m7).
8. A `time.Now` call added to `testdata/arch/time-now-nested/musicbrainz/events.go`
   with the walk narrowed back to the root → the nested fixture test must
   fail, proving the widening is real.

**All fixtures are hand-authored in this package's own `testdata`, never a
corpus case**, on the spec's proximity precedent (spec `:318`).

**Deliverable count: 5, and the plan says so plainly rather than letting
it pass as four** (fold, m9). CLAUDE.md's pass-sizing rule asks that a
task's count be stated at dispatch and that a count past roughly four be
named. This is the pass's only such task, and the two riders are the
reason: the browse-fallback guard and the widened clock walk are each a
handful of lines that belong beside the rest of the `conform` work rather
than in a task of their own, and moving them to Task 9 would put gate
code in the close. The five: `CheckTranslatedAlbum`'s field-by-field
comparison with `errors.Join` and the compared-row count; the
twenty-one-row nullability policy with its per-field reason and its
corrected field-count assertion; the profile-name overlay with its
unrecognized-name failure, its key-name drift guard, and the pattern-table
name guard; the pseudo-merge grading against the `Merged` block; and the
browse-fallback guard plus the widened clock walk with its nested
fixture. If the conductor prefers four, the cut is the last item into its
own task. Gate green plus tier-0 replay green; commit.

---

## Task 9: The pass close

Model: `sonnet`. Depends on Tasks 0 through 8.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the conductor's
  batch commit)
- Draft: the four `corpus/README.md` corrections (R16), the ledger
  regeneration command (Task 1), and **one rung 3 spec correction** (fold,
  m14), handed to the conductor in the task return

**Behavior contract:**

- **The HISTORY entry is drafted in the house shape**: what landed, what
  the gate caught, and what a later pass would be wrong to rediscover.
- **The translation coverage record, stated as a table a reviewer can
  check.** Per translation rule (T2 through T14): its file, its function,
  whether a ledger row grades it, and, for a rule no corpus case can
  reach, which hand-authored fixture grades it and what mutation proved
  that fixture fires. **The declared corpus-coverage gaps are recorded as
  gaps, not as passing checks** (spec, M8). This pass declares five:
  1. **Pregap**: zero committed cases carry a non-null `pregap`.
  2. **Data tracks**: zero committed cases carry a non-empty
     `data-tracks` array; both the `ignore_data_tracks` branch and the
     `"[data track]"` title filter are ungraded by the corpus.
  3. **The preferred-country release-event branch**: no committed profile
     sets a `match:` key, so every case takes
     `_preferred_release_event`'s fallback. **The branch is reachable
     data-side and unreached only for that reason**: 290 of the 298
     top-level release events in the 303 release-lookup responses carry a
     non-null `area` with `iso-3166-1-codes`, across 208 case
     directories, so a fourth profile setting
     `match.preferred.countries` would reach it immediately.
     `various-artists` is the concrete target, its two events differing
     in date. (Revision 1 recorded the opposite and it was false; a
     rediscover-guard is the expensive place for a wrong record.)
  4. **The merge's `year`, `original_year`, and `mediums` path**: equal
     on both sides of all four committed pseudo cases, so the merge's
     year path has no corpus signal at all.
  5. **The alias path's breadth**: `import.languages` is non-empty in
     exactly one committed profile and one committed case, so 214 of 215
     cases run with `preferredAlias` returning nil unconditionally.
  6. **Two boundary rules with no corpus instance**: no committed release
     carries an empty resolved track credit, and no committed backward
     `transl-tracklisting` relation lacks its `release` target, so both
     of `Validate`'s new rules (M-d, M-e) are graded by hand-authored
     fixtures only. Both close a Go panic rather than an error, so the
     absence is worth recording rather than celebrating.
  Together with 3a's three (`match.required`, `match.ignored`,
  `preferred.original_year`), rung 3 now carries **nine** declared gaps,
  and the entry states the running total so 3d's close does not re-count.
- **The evidence, not a bare "green".** The translation tier: rows
  compared, fields compared, the result, and the compared-row count
  against the ledger's row count. The ledger: subjects, rows, and the
  derivation's wall time. The `defect_free` block: the ten derived figures
  beside the ten prose figures, pair by pair. `make derive`: which tests
  ran, the byte-identity result over `components.json`, `ties.json`,
  `file_tags.json`, and every `case.json`. `bash scripts/check.sh` with
  broken-down wall time at `-count=1`, and tier-0 replay still green over
  all 72 rows.
- **Forward carries, each stated so 3c and 3d read them rather than derive
  them:**
  1. **The rung 15 re-run criterion**, quoted verbatim from the Spec
     section above. Every rung 3 pass close carries it.
  2. **The four-knob wall registration** stays the named follow-on task
     for the first rung 3 pass that runs after rung 1b closes, with the
     four knobs named and `internal/config/keys.go` named as the file no
     lane edits. Unchanged by this pass; restated so it does not fall off.
  3. **The three `SD_PATTERNS` tables and what each is for**:
     `BeetsSDPatterns` (shipped, defective on purpose),
     `SDPatternsWordBoundary` (the six `adv-ft-*` figures), and
     `SDPatternsNoBracketDiscount` (the four `adv-remix-*` figures, born
     here). 3d's compat switch selects a table by assignment to
     `Tunables.StringDistPatterns` and edits no component. 3a's warning
     travels with them: the word-boundary guard changes zero of the 72
     tier-0 rows, so tier 0 is blind to a "fixed" table.
  4. **The `defect_free` block as built**: its schema, the
     `StringHeuristicHypotheses` predicate and why the six `adv-assign-*`
     rows are excluded, the two legal `pattern_table` values, and the
     split between `CheckDefectFree`'s standing categorical cross-check
     and the one-time numeric one, so 3d's compat-switch work reads the
     expectation rather than re-deriving it.
  5. **The `AlbumInfo` ledger as built**: its `(case_id, release_mbid)`
     keying, its nineteen subjects and 50 rows, its graded field scope and
     the reason applied fields are excluded, and its `Merged` block, so
     rung 8 does not read it as an applied-field ledger and 3d does not
     re-key it.
  6. **The nullability policy** and R1's resolution: plain-string
     vocabulary, a source-scoped unreachability proof, and the strict-nil
     check as the thing that would catch a second metadata source
     reopening it. **3a's first carried item is closed by this pass**, and
     the entry says so explicitly.
  7. **The `Preferred.Countries` two-regime hazard**: plain
     case-sensitive membership in `internal/engine/musicbrainz`, an
     anchored case-insensitive regex in `internal/engine/matcher`. One
     config list, two rules, named at both sites.
  8. **The browse fallback's refusal and its guard**, with T16's exact
     trigger and the 470-track headroom, so 3c knows it owns the port if a
     corpus case ever crosses the line.
  9. **`Tunables.VAName` landed and `AliasesAsCredits` deliberately
     excluded**, with R8's reason, so a later pass does not add the second
     one "for symmetry".
  10. **3a's second carried item is re-addressed to 3d alone** (R5): the
      `Preferred.Media` / `Preferred.Countries` compile panics want a
      wrapped error at config-load time, `internal/config` has no
      validation hook, and 3b compiles no regex, so 3b was never its home.
  11. **The `\b` regime divergence** (m19): `SDPatternsWordBoundary`'s
      `\b` is Unicode-aware in Python's `re` and ASCII-only in Go's RE2,
      so 3d's compat switch and the `defect_free` derivation can differ
      on non-ASCII input. All six `adv-ft-*` pairs are ASCII, so nothing
      moves today. It travels beside 3a's `SD_PATTERNS`-as-a-value seam.
  12. **The wire-type census** (B3): `position` is the only field the
      MusicBrainz payload spells two ways in this corpus, 6,763 integers
      and 297 strings, and `wireInt` is the narrow answer. 3c's live
      client meets real captures only, where the string form does not
      occur, so the type is carried for the corpus rather than for the
      wire.
  13. **The measured wall times**: the derivation, the translation tier,
      tier 0, `go test ./...`, and `scripts/check.sh`, which the
      version-bump differential doc will want, the same courtesy 2b passed
      to 2c, 2c to 2d, and 2d to 3a.
- **The drafted spec correction** (fold, m14). The rung 3 spec's pass-split
  table gives 3b the tree column "Main tree: writes `corpus/` and
  `internal/harness/oracle`" (spec `:329`). This pass also writes
  `internal/harness/{analysis,conform}` and `internal/harness/corpus_test.go`,
  and it births `internal/engine/musicbrainz`. The plan justifies each
  under R11 and Global Constraints, but an uncorrected spec is the next
  pass's stale input, and this pass drafts four `corpus/README.md`
  corrections for exactly that reason. One drafted line goes into the
  conductor's batch. **Conductor ruling, 2026-09-04: taken as recommended**; the conductor landed
  the spec's tree-column correction before launch, so Task 9 has nothing
  to draft for it.
- **STATUS stays present tense and under 60 lines**; anything historical
  moves to HISTORY. The next action points at 3c (3b and 3c are mutually
  independent and may swap order; 3d needs all three), and the lane
  registry entry is updated.
- **Both budgets are scored**: token spend against the re-derived ceiling,
  and attended interactions against zero, which is this pass's target.

**Deliverable count: 4** (the HISTORY draft in the house shape with its
thirteen forward carries; the translation coverage table with the six
declared corpus-coverage gaps and the running total of nine; the evidence
record with broken-down wall times, the byte-identity result, and the ten
figure pairs; the STATUS rewrite with the next action, plus the four
drafted `corpus/README.md` corrections and the drafted regeneration
command, plus the drafted spec correction). Gate green plus tier-0 replay
green; commit.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain per task,
**strictly sequential**. Nothing here runs in parallel, and if a second
code slot opens at all it belongs to lane A or lane T in its own worktree,
never to a second executor in this tree. `docs/STATUS.md`'s lane registry
records lane-2 as BLOCKED until Phase 0b and 0c land, so the expected
state through this pass is one code lane. Every task runs implementer →
`diff-reviewer` → gate. The conductor dispatches `code-simplifier` over
each accepted task's changed Go and Python after review and before the
pass close, and owns every ledger edit, the lane registry, the four
`corpus/README.md` corrections, and the batch commit. **Runaway guard
armed at launch**, per the standing requirement. **Subagents never
dispatch agents.**

**Models.** `sonnet` by default. **Two named upshifts to `opus`:**

- **Task 1**, the `AlbumInfo` ledger. It is the written specification every
  later Go task in this pass is graded against, it drives a beets entry
  point nobody has driven before, and a ledger that records the wrong
  thing produces five green tasks and a wrong translation layer.
- **Task 4**, aliases and artist credits. The alias path changes 90 of 303
  album `artist` values and 3,272 of 6,919 track `artist` values under a
  non-empty `languages`, yet exactly one committed case runs under a
  profile that sets it, so the corpus cannot catch a subtly wrong
  precedence rule. Its three interlocking call sites, the recording-title
  guard, the credit-name split, and the locale-order-beats-response-order
  loop nesting, are each independently easy to get plausibly wrong.

**Task 8 is the runner-up**: the conductor upshifts it if Task 4's or Task
6's report says a translation contract above was under-specified, because
it is the instrument every claim about this pass rests on.

**Review.** Rung 3 is not one of the three full adversarial-cycle rungs
(5, 12, 16), so review scales to the in-chain `diff-reviewer` per task plus
the pass-end gate. **This pass does not take a fresh-context review of its
own instrument**, and the reason is recorded rather than defaulted: the
spec raises 3d's conformance runner as the one candidate exception (spec
`:362`), and 3b's translation tier is materially cheaper to trust than
3d's will be. It grades 50 rows whose expected values were produced by
executing pinned beets, its skipped-row failure and its
field-count-versus-policy assertion make silence impossible, and every
checker ships a firing fixture with a named mutation. The conductor may
still take one.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer dispatch.**
  Its blockers are ruled by the conductor before Task 1.
- **Task 1 → everything.** The committed ledger is the expected output
  Tasks 3 through 7 are written against, and Task 8 grades against it.
- **Task 2 depends on Task 1** for the shim conventions it copies, and on
  nothing else. It is the last task that touches Python.
- **Task 3 → Tasks 4, 5, 6, 7.** It births the package and defines the
  release vocabulary every later task decodes into.
- **Task 3 → Task 4 → Task 5.** Aliases and credits come before track
  extraction because both track-level overrides that produce a
  `matcher.TrackInfo` call into them: the title guard reads
  `preferredAlias` and the artist reads `parseArtistCredits`. Ordering
  them the other way would defer two of Task 5's tests into Task 4 and
  make the deferral the thing a reviewer has to notice, which is what
  3a's `AddString` escalation cost a revision note (3a HISTORY, revision
  2.1).
- **Tasks 3, 4, 5 → Task 6.** `TranslateAlbum` composes all three.
- **Tasks 3, 6 → Task 7.** The merge grades against `TranslateAlbum`'s
  output.
- **Tasks 1–7 → Task 8 → Task 9.**

**Where this pass splits, if it splits.** Nine tasks plus a pre-task is
past the roughly-eight bar, so the split point is named rather than the
scope packed. **The split is at the Task 2 / Task 3 boundary**, and it is
clean in one direction: Tasks 0–2 are the corpus and derivation half
(`corpus/`, `internal/harness/{analysis,oracle}`, and two `matcher`
table additions) and write no `internal/engine/musicbrainz` at all; Tasks
3–8 are the engine half plus its grading.

**The recommendation is to keep one pass.** The halves are each
incomplete without the other: the ledger is an ungraded 50-row JSON file
until `conform` compares something to it, and the translation layer is
ungraded against beets until the ledger exists, so a 3b-i that shipped
Tasks 0–2 would ship exactly the "instrument nobody points at anything"
shape the spec's risk 2 names. Splitting also costs a second staleness
pre-task, a second close, and a second `make derive` cycle at ~585s for
the oracle package. Take the split only if Task 3's actual spend
re-derives the ceiling past 6M. **The conductor takes the call at the Task
3 checkpoint, on the number rather than on a feeling.**

**Budget:** 6M starting ceiling, re-derived after Task 3; phase 1's
combined ceiling 12M as a reservation; flag at 4.8M; STATUS after every
task. Spend is scored at close against the re-derived ceiling and against
**zero attended interactions**, which is this pass's target.

---
