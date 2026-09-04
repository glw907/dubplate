# Rung 3, Pass 3a (The Scoring Core, `string_dist`, and Tier 0) Implementation Plan

**Revision 2, 2026-09-04:** folds the adversarial plan review
(`research/2026-09-04-rung-3a-plan-review.md`), blockers B1–B4 (B4 by
conductor ruling: the fifth contested-surface row), majors M-a through
M-i, minors m1–m15. The load-bearing resolutions: **Task 6's decoder
contract is rewritten from `derive_component.py`'s actual emission** —
dispatch on `method`, with the add-key-versus-weights shape split inside
method `add` (eleven methods, twelve shapes), the row `weights` map as an
**override on `Tunables.DistanceWeights`** rather than a complete table,
and the `get_penalty` stub model wired through `Tunables`' loaded-source
set (B1–B3). **The `SD_PATTERNS` tier-0 safety claim is deleted** — the
word-boundary guard is provably invisible to all 72 rows — and replaced by
Task 4's byte-identity table test plus a discriminating fixture carrying
both measured figures (M-a). Three corpus-reading assertions move out of
engine tasks into Task 6 (M-b, M-c, M-d). Contract 7 splits: regex
**options** anchor, `SD_PATTERNS` never do (M-f). Conform owns its own
decoder-method set (M-g). The lane-T pairing and the 12M phase-1 figure are
stated as conditional and as a reservation (M-h). Contract 8's fourteen
album components ship in thirteen files, `album_extras.go` named as the
exception (M-i). The minors are folded as editorial; m15 is
conductor-owed, not a plan edit, and landed at `889d71f`. **The review's
"Verified sound" list is a second do-not-re-litigate record**: its
twenty-two items — all twelve arithmetic contracts, the eight adder
semantics, both component orders, every track fire condition and album
branch, the twenty weights, every `config_default.yaml` citation, the five
`unidecode` measurements, the corpus figures, the M2 and m2 resolutions,
the harness API signatures, the gate coverage claim, and the split-point
reasoning — are cited below rather than re-derived, and an implementer who
believes one of them is wrong reports a finding and implements the record
anyway.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section: pass-execute chain per task, sequential. The conductor consumes
> reports and never reads diffs inline. Subagents never dispatch other
> agents; `code-simplifier` is the conductor's dispatch after a task's
> review accepts. The conductor also owns the ledgers (`docs/STATUS.md`,
> `docs/HISTORY.md`, `ROADMAP.md`), the lane registry, and the batch commit
> of anything a task drafts into them — including this pass's one
> `corpus/README.md` correction, which an implementer drafts and never
> writes. Rung 3 is not an adversarial-cycle rung (those are 5, 12, and
> 16), so review scales to the in-chain `diff-reviewer` per task plus the
> pass-end gate. **This pass runs in the MAIN TREE.** The worktree
> early-start remains an explicitly costed option nobody currently needs
> (rung 3 spec, revision 2, "Lane placement"; fold, m1); nothing here may
> be executed from a worktree.

Drafted 2026-09-04 against the rung 3 spec at **revision 2**
(`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`), which is
ratified on its six rulings and review-folded on all nineteen findings of
`research/2026-09-04-rung-3-spec-review.md`. That review's **"Verified
sound" list is the do-not-re-litigate record**: its twenty items — the five
deliverables and their framing, the launch-bar restatement, the two
architectural invariants, the package territory, port hazards 1 and 3, the
four accumulation claims, the bit-identity argument, the component row
floor of 72, the proximity precedent, and the rest — are cited here rather
than re-derived, and an implementer who believes one of them is wrong
reports a finding and implements the record anyway. This plan folds every
spec note that touches 3a: ruling C's fallback (rung 1b has **not** landed —
`internal/config/keys.go` still carries `ownedWall` as a slice literal — so
`Tunables` carries the beets defaults with citations and the four-knob wall
registration is a **named follow-on task for the first rung 3 pass that
runs after 1b closes**, and 3a does not edit `internal/config` at all);
port hazards 1 through 4; the M1 merge cross-check, landed here in the one
form that can fire today; the M2 execution-time discrepancy, **resolved
below against pinned beets rather than against either document**; m2's
twenty distance-weight keys, resolved off the committed records; and m4's
`AddExpr` in the adder list.

**Phase 0a note.** `research/2026-09-04-reverse-edge-completion.md`'s
permit/forbid section governs the lane pairing: **lane T beside 3a** is the
preferred allocation *if a second code lane can run at all*, because 3a
writes `internal/engine/matcher` and `internal/harness/conform` and touches
neither `internal/harness/oracle` nor `corpus/`, which is the smallest
collision surface in phase 1. **The pairing is conditional, not the
committed state** (fold, M-h): `reverse-edge-completion.md`'s own
precondition is that Phase 0a, 0b, and 0c have all landed, and
`docs/STATUS.md`'s lane registry records lane-2 as BLOCKED until they do.
Task 0 item 1 expects 0b and 0c unlanded, so as committed **no lane T runs
beside 3a**, and this plan neither assumes one nor depends on one. The
preference is a constraint on this plan either way: nothing here may write
`corpus/` or `internal/harness/{cassette,stub,analysis,oracle}`.
`internal/harness/conform` is **new territory 3a births**, and it is the
only harness path this pass creates or writes.

**Consumes rung 2's shipped substrate exactly as committed**, read-only:
`corpus/analysis/components.json` (72 rows, `beets` `2.13.1`,
`shim_version` `1`), `internal/harness/analysis`'s `ComponentTable`,
`ComponentRow`, `LoadComponentTable`, and `ComponentRowFloor`,
`internal/harness/cassette`'s `Case`, `Expected`, `Thresholds`,
`Candidate`, and `Manifest`, the 215 committed cases and their echoed
records, the three profiles, and `corpus/manifest.json`'s six version pins.
3a adds Go packages beside all of it and changes none of it.

## Primary inputs

Read these before Task 1, in this order:

1. **`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`, revision 2**,
   in full. The pass-split table's 3a row is this plan's scope statement.
2. **`docs/superpowers/research/2026-09-02-beets-matcher-mechanics.md`** —
   the port map. Every arithmetic contract below comes from it, cited by
   heading (Q2's "String distance", "The Distance object", "Track
   distance", "Album distance"; Q4's tunable enumeration; the float-regime
   section). **Treat its line numbers as possibly stale**: a 2d task
   already caught line-citation drift in a planning document, and this
   plan's own citations were re-verified against the pinned venv at
   drafting time (see the two resolutions below). Verify any line citation
   you carry forward.
3. **`docs/superpowers/research/2026-09-04-rung-3-spec-review.md`** — the
   Verified sound list, cited and not re-derived.
4. **`docs/superpowers/research/2026-09-04-reverse-edge-completion.md`**,
   the "What the results permit and forbid in phase 1" section.
5. The tree: `internal/harness/analysis/component.go`,
   `corpus/analysis/components.json`, `internal/harness/cassette/cassette.go`,
   `internal/harness/corpus_test.go`, `internal/config`, and
   `docs/HISTORY.md`'s 2c and 2d entries. **`repoRoot` is read for its
   shape, never called** (fold, m12): it is unexported and test-only
   (`corpus_test.go:19-36`), so `internal/harness/conform` reimplements the
   same walk-up-to-`go.mod` helper in its own test file rather than
   importing anything.
6. `CLAUDE.md`'s Build process and Gate sections, `docs/STATUS.md`'s lane
   registry, and `corpus/README.md`'s "Config profiles" section.
7. **`internal/harness/oracle/derive_component.py`, in full, before Task
   6** — the emitter every `components.json` row comes from, and the
   authority on the decoder's shapes. Reading oracle territory is
   permitted and expected; writing it is not. Revision 1's Task 6 was
   written from the row ids instead and was unexecutable as a result
   (fold, B1–B3).

## Goal

Ship the arithmetic half of the Go matcher, graded. Land
`internal/engine/matcher` — the verified-complete tunable surface with
pinned beets' shipped defaults and their citations, the `Distance`
accumulator with both of its penalty views, `string_dist` with `SD_PATTERNS`
and a transliteration table generated from the pinned `unidecode`, every
album-level and track-level penalty component with its own test, and the
scoring core that composes them in beets' own fixed order — and land
`internal/harness/conform` with the typed `components.json` decoder and
tier 0, replaying **all 72 committed rows green** against the Go
primitives they name.

Nothing here decides a verdict. Recommendation, `max_rec`, verdict-to-action,
the assignment solve, admission filters, the compat switches, and tiers 1
and 2 are 3d's. The translation layer is 3b's. The client and limiter are
3c's.

## Owner touchpoints: NONE

3a needs no ruling, no adjudication, no capture contact, and no network.
The six rulings this rung rests on were ratified on 2026-09-03 and the
first adjudication batch has been decided: as read at drafting time,
`corpus/analysis/oracle_bugs.json` carries four hypotheses at
`oracle-wrong` (`bug-ft-pattern`, `bug-remix-parenthetical`,
`bug-nonlatin-alias-selection`, `bug-assignment-rejection-gap`) and two at
`oracle-correct` (`bug-threshold-boundary`, `bug-locale-alias-ignored`),
all with `adjudicator: "Geoff"`, and `adjudications.json` carries 21 rows
at 18 `oracle-wrong` / 3 `oracle-correct` with none pending. **Task 0
re-reads that state off the tree and reports it**; the numbers above are
this plan's drafting-time snapshot, not a pin.

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition named below resolves to the conductor. Only a
conductor decision to escalate reaches Geoff. The standing 80%-of-ceiling
flag CLAUDE.md requires of every pass still applies and is contingent, not
a touchpoint this plan expects to spend.

## Spec

`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md` (revision 2), and
behind it `specs/2026-09-02-greenfield-server-build.md` (revision 2.1,
ratified), ladder entry 3. Deliverable 1 ("the penalty components and the
scoring core, structured per the evolvability ruling so that changing a
matching rule touches one component and its cases rather than a fused
loop") and deliverable 4 ("the verified-complete tunable surface, with the
evaluation date as an explicit input and never a clock read") are what this
pass ships. Deliverables 2, 3, and 5 are 3b, 3d, and 3c respectively.

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

As a **starting** figure, per the spec's Band and ceiling section, and
**re-derived from the first substantive task's actual report** — Task 2,
the accumulator — before Task 3 dispatches, and again after Task 4 if
`string_dist` overruns. The basis: 2a spent ≈1.75M against 5M, 2b ≈5.8M
against 6M, and **2c closed at ≈8.2M against its 8M re-derived ceiling with
the flag overrun approved**; 2d spent ≈1.6M against 2.5M. 3a is nine tasks
of machinery with no case authoring, no capture, no network, and no
`make derive`, which is 2d's cost shape at **2.25 times** 2d's task count
(fold, m7: the earlier "three times" was wrong, and 3 × 1.6M would have
argued for 4.8M, not 6M). **6M is the rung 3 spec's own proposal** (spec,
"Band and ceiling"), and the per-task arithmetic supports it rather than
deriving it: ≈665K per task against 2b's ≈415K per-task average, on a
harder surface with two `opus` tasks in it.

**Flag at 80% (4.8M):** finish the task in flight, write STATUS, and one
combined question goes to Geoff before continuing.

**Checkpoint: STATUS after every task.** That is 2b's, 2c's, and 2d's
cadence, and at nine tasks it is what the conductor sizes the next task
against — every task reports its own spend and its own gate evidence.

**Phase 1's combined ceiling: 12M, and it is a reservation rather than a
live sum** (fold, M-h). It is named here because the spec's m1 fold
requires 3a's plan to state it and because the 80% stop rule has nothing to
evaluate against inside a lane that cannot see the others' spend. It is
3a's 6M plus a **6M reservation for the second code slot** (lane T, rung 8,
M band, *if* it launches). The lane registry records lane-2 as BLOCKED
until Phase 0a/0b/0c land, so **no second sub-ceiling exists to sum
today**, and 12M is not a figure any spend is currently measured against.
The conductor replaces the reservation with lane T's own assigned
sub-ceiling at that lane's launch, and until then 3a's own 6M is the only
live ceiling.

## Two resolutions this plan takes rather than discovers

The spec names both as execution-time work for 3a. Both were resolved
against **pinned beets 2.13.1 in `tools/venv/.venv`** while this plan was
drafted, with the evidence below. Task 1 re-verifies them against the venv
at execution time and reports the evidence; it does not re-open them.

### M2: `original_date` and `per_disc_numbering` are NOT verdict-moving

`corpus/README.md`'s config-profile transform table calls both
**verdict-moving** (`original_date` "controls which year the year penalty
evaluates against"; `per_disc_numbering` "affects multi-disc track distance
and assignment"). The mechanics study's Q4 lists both under "Checked and
found NOT verdict-moving … application-time only". **The mechanics study
wins.** The evidence, read off the pinned venv:

- `original_date` is read in exactly one place in the whole tree:
  `beets/autotag/hooks.py`'s `AlbumInfo.item_data` (`config["original_date"].get(bool)`),
  which overrides `year`/`month`/`day` in the metadata a match *applies*.
- `per_disc_numbering` is read in `TrackInfo.raw_data` (`medium_index` and
  `tracktotal`), in `beets/library/models.py`'s display path, and in the
  importer's display code — nowhere else.
- `raw_data` and `item_data` are consumed by three call sites outside
  `hooks.py`, all in `beets/autotag/match.py`: `Match.merged_pairs`,
  `AlbumMatch.apply_album_metadata` (`:199-201`, a **sibling** of
  `apply_metadata` rather than a call inside it), and
  `TrackMatch.apply_metadata` (`:239`). All three run **after** the
  verdict, so the conclusion is unchanged (fold, m5).
- The distance path never touches either. `track_index_changed` reads
  `track_info.medium_index` and `track_info.index` as attributes, not
  through `raw_data`; `distance()` and `track_distance()` read no config
  key outside `match:` and the data-source penalty.

**`corpus/README.md` already proves this against itself.** Its
stock-sibling stratum finding says plainly that both keys "are each read in
exactly one place: `AlbumInfo.item_data` … Neither key is read anywhere in
`beets/autotag/distance.py` or `beets/autotag/match.py`", and that "the
divergence lands in `applied_fields` instead, on all 28 without exception."
So the correction is a consistency fix inside one document, not a new
claim: **two cells of the config-profile transform table** change from
"verdict-moving: …" to "**applied-field-moving, never verdict-moving**:
read only in `AlbumInfo.item_data` / `TrackInfo.raw_data`
(`beets/autotag/hooks.py`), consumed only by `apply_metadata`
(`beets/autotag/match.py`); kept verbatim because it moves every stock
sibling's `applied_fields`, which is rung 8's conformance target — see the
stock-config stratum finding above." The consequence for the port:
**`Tunables` carries neither key**, and the enumeration in the spec stands
unamended.

**The README sentence quoted just above is itself imprecise, and the
drafted correction says so** (fold, m5): `per_disc_numbering` is read in
`TrackInfo.raw_data` (`hooks.py:421-452`), not in `AlbumInfo.item_data`.
The replacement text already names both accessors correctly; Task 1's draft
adds one clause noting that the earlier sentence's single-accessor phrasing
is what the correction tightens, so the conductor lands a fix rather than a
contradiction.

The correction is drafted by Task 1 and **landed by the conductor** in its
batch commit, not written by an implementer: this plan writes nothing under
`corpus/`, and the house pattern already routes prose drafts through the
conductor (2d Task 4's ledger drafts are the precedent).

### m2: `distance_weights` is twenty keys

Confirmed twice over, so the mechanics study's Q4 ("19 keys") is the losing
sentence and its own Q2 list is right. Off the tree: **all 215 committed
cases echo one weight set, and it carries twenty keys** (`album`,
`album_id`, `albumdisambig`, `artist`, `catalognum`, `country`,
`data_source`, `label`, `media`, `medium`, `mediums`, `missing_tracks`,
`track_artist`, `track_id`, `track_index`, `track_length`, `track_title`,
`tracks`, `unmatched_tracks`, `year`); all 215 echo one threshold set
(`0.04` / `0.25` / `0.25`, `max_rec` `{missing_tracks: medium,
unmatched_tracks: medium}`). Off the pinned venv:
`beets/config_default.yaml` lines **174–193** are exactly twenty entries
(`:173` is the `distance_weights:` header — fold, m4). No
document edit is owed — the study is research, not a spec — but 3a's report
records Q4's count as superseded so a later reader does not re-derive it.

### What the ruled dispositions mean for 3a

Ruling A's compat switches live in `compat.go` at **3d**, not here. But two
of the four `oracle-wrong` hypotheses — `bug-ft-pattern` and
`bug-remix-parenthetical` — are defects *of `SD_PATTERNS`*, which is 3a's
file. The corpus's tier-0 `string_dist` rows exercise the featuring,
parenthetical, and bracket discounts (`relative/3` is the featuring pair,
`relative/1` and `relative/2` the parenthetical and bracket ones), but
**they do not exercise the defect**: in
`string_dist("My Song feat Someone", "My Song")` the pattern fires on the
standalone word "feat", which is the correct behavior, and the row's
`0.06470588235294118` is the same under the guarded table. The adjudicated
defect is `ft` matching *inside another word* — `adj-adv-ft-draft-colon`,
where "Draft:" is eaten as a featuring credit. So:

- **3a ships beets' behavior exactly**, defect included.
- **Tier 0 cannot catch a "fixed" pattern table, and this plan does not
  claim it can** (fold, M-a). Measured against pinned beets at revision-2
  drafting time: applying the word-boundary guard named in
  `adj-adv-ft-draft-colon` changes **zero of the 72 rows**. The guard is
  invisible to the corpus, so a well-meaning implementer who "corrects"
  `BeetsSDPatterns` ships a green pass and a wrong matcher. **The guard
  against that is Task 4's, explicit and byte-level**, not tier 0's by
  construction: a test asserting `BeetsSDPatterns` is byte-identical to
  beets' six shipped literals, plus a discriminating fixture that separates
  the two tables. The discriminating pair and both measured figures, for
  Task 4's fixture and for 3b's `defect_free` derivation:
  `string_dist("Draft: Ember Wharf", "Draft: Northern Lines")` =
  **`0.061111111111111116`** under `BeetsSDPatterns` and
  **`0.6111111111111112`** under `SDPatternsWordBoundary`. Both figures
  travel in the pass report.
- **The fix must be a table swap, not a code edit.** The adjudication rows
  name the correction as a word-boundary guard on the featuring pattern
  (upstream `19d091440`, unreleased at pinned 2.13.1), and 3b derives each
  `defect_free` figure "through pinned beets with the one named
  `SD_PATTERNS` edit". 3a therefore exposes the pattern set as a **named
  value** — `Tunables.StringDistPatterns`, defaulting to
  `BeetsSDPatterns`, with `SDPatternsWordBoundary` as a second exported
  table carrying the guarded featuring pattern and a doc comment citing
  `adj-adv-ft-draft-colon` — so 3d's compat switch selects a table and
  touches no component. This is the one place 3a builds *for* a later pass
  rather than only for itself, and it is cheap now and expensive later.

## Global Constraints

- **Main tree.** The conductor's fold ruling: 3a's main-tree default
  stands; the worktree early-start is an explicitly costed option nobody
  needs. The lane registry entry is written before the first dispatch, and
  the live-executor check is that registry, never `pgrep -f Projects/dubplate`.
  Lane T may hold the worktree beside this pass only under its own
  read-only-against-`corpus/`-and-`internal/harness` statement.
- **Package territory, stated as a wall.** This pass writes
  **`internal/engine/matcher`**, **`internal/harness/conform`** (new,
  born here), and **`tools/unidecode-table/`** (Task 3's one-time
  generator), plus conductor-owned `docs/` ledgers. It writes **nothing**
  under `corpus/`, `internal/harness/{cassette,stub,analysis,oracle,synth,mbrequest}`,
  `internal/harness/corpus_test.go`, `internal/config`, `cmd/`, or
  `_salvage/`. A task that believes it needs one of those stops and
  reports. **`tools/` is otherwise lane E1's** — the parallel spec's lane
  table gives it the whole directory — so `tools/unidecode-table/` is a
  carve-out held by a written contested-surface row (Task 0 item 4;
  conductor ruling, 2026-09-04, fold B4), and **nothing else under
  `tools/` is 3a's**, `tools/venv/` emphatically included: this pass reads
  the venv and never writes it.
- **`.gitleaksignore` is conductor-owned and conditionally permitted**
  (fold, m11). It sits on neither side of the wall above, and Task 3 may
  need an append if the generated table trips a rule. An append is
  **drafted by the task and landed by the conductor**, must be strictly
  **below** the existing `internal/config/keys_test.go:generic-api-key:66`
  line, and is never a rule disable. An insertion above line 66 reds the
  gate for every lane at once.
- **`internal/config` is not edited, and neither is `keys.go`.** Ruling C's
  registration half is hard-inbound on rung 1b (Phase 0b, unlanded — the
  tree still carries `ownedWall` as a slice literal). `Tunables` carries
  all four knobs' defaults with their citations meanwhile, and the
  four-knob registration is a **named follow-on task for the first rung 3
  pass that runs after 1b closes**. This is a plan-header item, not a
  discovery. The `.gitleaksignore` rider matters even though we do not
  touch the file: its single allowed finding is pinned by line number at
  `internal/config/keys_test.go:generic-api-key:66`, and an insertion above
  line 66 reds the gate for every lane at once.
- **The engine never imports the harness.** `internal/harness` reads the
  corpus, drives beets, and grades; the engine is the product. This is why
  tier 0's decoder and runner live in `internal/harness/conform` even
  though 3a writes them (spec, M5 fold). **The invariant ships with a gate
  check** (Task 6): a stdlib-only source walk over `internal/engine/...`
  that fails on any import path under `github.com/glw907/dubplate/internal/harness`,
  and on any `time.Now` call inside `internal/engine/matcher` (the
  evaluation date is an input, never a clock read). Both halves ship a
  `testdata` fixture proving they fire.
- **Assertion discipline, at the spec's own bar.** Every checker ships a
  fixture that makes it fire, and the task that lands it **reports the
  mutation that proved it can fire**. No silent-success fakes, no vacuous
  assertions. A conformance runner that silently passes because it graded
  nothing is the single most expensive defect this rung can ship (spec,
  risk 2). Hand-authored fixtures live in the owning package's own
  `testdata`, on the proximity precedent — **never a corpus case**.
- **No `make derive`, anywhere in this pass.** No task touches oracle
  territory, so no task derives. Two tasks read the venv (Task 1 reads
  beets' source to re-verify the M2 resolution; Task 3 runs a committed
  one-time generator against `unidecode` 1.4.0), and neither is a
  derivation: the gate never invokes either, and neither writes `corpus/`.
  A task that believes it needs `make derive` stops and reports.
- **No live network. Anywhere.** The gate never reaches the network and
  neither does anything else here. No `curl`, no `WebFetch`, no second HTTP
  client, no `/ws/2` request. The one Python run is offline against a
  vendored package already in the venv.
- **Corpus data is frozen input, and read-only is a real permission.**
  Reading `corpus/analysis/components.json` and every committed
  `case.json` is exactly what tier 0 and the defaults cross-check are for.
  Writing any of it is out of scope for the pass.
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck,
  ruff over every tracked `*.py` (Task 3 adds one), vale over Go comment
  prose, gitleaks over the tracked worktree, `docker compose config`,
  `systemd-analyze verify`. **Never commit red. Specific files only.**
  Commit footer `Co-Authored-By: Claude <noreply@anthropic.com>`.
  From Task 6 on, **tier-0 replay green over all 72 rows** is part of every
  subsequent task's gate expectation, because it runs inside `go test ./...`.
- **`go-conventions` before any Go; `python-conventions` before Task 3's
  generator.** The bar is the standing exemplary-stdlib-idiom bar
  (`idiomatic-beautiful-go`) and the ops-seam-shaped small-package
  discipline. The scoring core is **one package**: components are files,
  not subpackages (spec, "One package for the scoring core"). The one
  permitted exception is data, not behavior — Task 3's transliteration
  table is an embedded blob inside package `matcher`, not a second package.
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects today. No third-party transliteration, Levenshtein, or regex
  package is adopted; a task that wants one stops and reports (Task 3's
  contract states why).
- **Universality.** No estate hostname, path, email, or catalog value
  anywhere outside `internal/config`. Test fixtures are invented strings,
  not estate data.
- **Determinism.** No map-iteration order reaches an output. Every
  accumulation runs in a fixed, code-determined order; every ordering the
  port reproduces is beets' own. `_penalties` is an **insertion-ordered**
  structure in Python and must be an insertion-ordered structure in Go
  (a key slice beside the map, not `maps.Keys`).
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each before
  Task 1 dispatches.

## The arithmetic contracts, stated once

Every task below cites this section instead of restating it. Sources: the
mechanics study's Q2 ("String distance", "The Distance object", "Track
distance", "Album distance") and its float-regime section, all re-verified
against the pinned venv at drafting time.

**1. The `string_dist` pipeline is lower → transliterate → lower.**
`string_dist` lowercases both inputs *before* calling
`_string_dist_basic`, which transliterates through `unidecode` and then
lowercases again. The port replicates all three steps in that order (port
hazard 1; study, "String distance"; Verified sound item 5). Both `None`
→ `0.0`; exactly one `None` → `1.0`. Then article rotation over
`SD_END_WORDS = ["the", "a", "an"]` — a trailing `", <word>"` moves to the
front — then `&` → `and`, then the pattern loop, then `base_dist + penalty`.

**2. No Unicode normalization happens anywhere in the distance path.**
The Go port must not helpfully normalize: no NFC, no NFD, no
`golang.org/x/text/unicode/norm`. This is **rung 3's own assertion, not the
study's** (fold, m6): the study documents transliteration and says nothing
about NFC/NFD either way. Task 4 verifies it against pinned beets and
records the check in its report rather than citing the study for it.

**3. The strong threshold compares with a strict `<` and the medium
threshold with `<=`.** This asymmetry is the whole reason the proximity
flag exists (port hazard 3; study, "From distance to verdict"). 3a does not
compute a recommendation, so 3a's obligation is narrower and still real:
`Tunables`' threshold fields carry the asymmetry in their doc comments, and
no 3a code introduces a comparison helper that flattens it for 3d to
inherit.

**4. The added-key map and the nonzero-key view are different objects.**
Port hazard 4, and the spec's m5 fold. `derive_oracle.py` emits
`{key: dist[key] for key in dist._penalties}`, so `cassette.Candidate.Penalties`
grades **key presence at 0.0** and 272 committed candidates carry one.
`max_rec`'s downgrade instead reads `keys()`, which yields **nonzero
penalties only**, so a 0.0 penalty never triggers a downgrade. A port that
grades against one and downgrades against the other passes tier 1 and
mis-downgrades; a port that does the reverse fails tier 1 on every case
carrying a zero-valued key. **The accumulator exposes both views under
distinct names, and each has its own unit test.**

**5. The normalizer's denominator counts only penalties that were added.**
`raw = Σ_keys sum(penalties[key]) × weight[key]`;
`max = Σ_keys len(penalties[key]) × weight[key]`;
`distance = raw / max`, **and `0.0` when `max` is zero** (study, "The
Distance object"; Verified sound item 6). A per-key value is
`sum(penalties[key]) × weight[key] / max`, `0.0` when `max` is zero. The
nonzero view sorts by `(-weighted_dist, key)`. `add_number` appends
`abs(n1-n2)` separate `1.0` penalties (or one `0.0`), so **one key can
carry several penalties and inflate the denominator per unit**.
`components.json`'s `aggregate/*` rows pin all of this, including
`aggregate/all-zero-weights` (max 0 with a key present) and
`aggregate/empty`.

**6. The adders, with their exact semantics** (`beets/autotag/distance.py`,
re-read at drafting time):

| Adder | Contract |
| --- | --- |
| `Add` | append one penalty; **error outside `[0,1]`** (beets raises `ValueError`) |
| `AddEquality` | `0.0` if any option equals the value, else `1.0`; a compiled-regex option **matches** rather than compares |
| `AddExpr` | `true` → `1.0`, `false` → `0.0` (m4: named because `add_expr` is in `componentMethods` with two committed rows) |
| `AddNumber` | `abs(n1-n2)` separate `1.0` penalties, or one `0.0` |
| `AddPriority` | `i × (1/len(options))` for the first matching option index, `1.0` if none match; `len(options) or 1` guards the divisor |
| `AddRatio` | `float(max(min(n1, n2), 0)) / n2`, `0.0` when `n2` is falsy |
| `AddString` | `string_dist` |
| `AddDataSource` | adds `get_penalty(after)` **only when** `before != after` **and** (`before` is set **or** more than one metadata source is loaded); a fresh single-source import adds nothing at all, not even a `0.0`, so the key stays out of the denominator |

**6a. `AddEquality` and `AddPriority` wrap a scalar option into a
one-element list** (fold, M-e). `distance.py:287-288` and `:328-329` are
`if not isinstance(options, (list, tuple)): options = [options]`, and that
is not a defensive nicety — it is how `distance()` itself calls both:
`add_equality("media", album_info.media, likelies["media"])` (`:465`) and
`add_equality("album_id", likelies["mb_albumid"], album_info.album_id)`
(`:522-524`) each pass a bare scalar. Tier 0 exercises both forms:
`add_priority/2`'s first call passes a scalar and its second a list. The Go
signature accepts one option or many, and `AddPriority`'s
`1.0 / (len(options) or 1)` divisor is computed **after** the wrap, so a
scalar option yields a unit of `1.0`.

**7. Regex OPTIONS anchor. `SD_PATTERNS` never do.** These are two
different regex regimes and conflating them destroys the pattern loop
(fold, M-f).

- **`Distance._eq`'s options anchor.** A compiled-pattern option to
  `AddEquality` or `AddPriority` is matched with Python's
  `re.Pattern.match` (`distance.py:265-267`), which anchors at position 0.
  The Go port anchors explicitly (`^(?:…)`), never `MatchString` on a bare
  pattern. Related: `re.I` becomes `(?i)`; the media patterns are compiled
  as `(\d+x)?(<pat>)` case-insensitive and the country patterns as
  `<pat>` case-insensitive.
- **`SD_PATTERNS` are `re.sub`, unanchored and global**
  (`distance.py:99-100`). The loop deletes **every non-overlapping match
  over the whole string** — Go's `Regexp.ReplaceAllString` with an
  unanchored pattern — and anchoring one at `^` would silently break the
  parenthetical, bracket, featuring, and part discounts. The one pattern
  that carries `^` carries it in beets' own literal (`^the `), and the port
  copies the literal rather than adding an anchor of its own.

Python's `$` matches before a trailing newline while Go's does not; no
committed pattern in either regime uses `$`, and a port that adds one is a
finding.

**8. Component order is load-bearing, not stylistic.** Which components
fire and in what order determines the denominator, so a component that
declines to add anything must decline in exactly the place beets declines.
The fixed track order is `track_length`, `track_title`, `track_artist`,
`track_index`, `track_id`, `medium`, `data_source`. The fixed album order
is `artist` (skipped entirely when `album_info.va`), `album`, `media`,
`mediums`, `year`, `country`, `label`, `catalognum`, `albumdisambig`,
`album_id`, `tracks`, `missing_tracks`, `unmatched_tracks`, `data_source`
— **fourteen album components, shipped in thirteen files**, with
`missing_tracks` and `unmatched_tracks` sharing `album_extras.go` as Task
7's one named exception to one-component-one-file (fold, M-i).

**9. The evaluation date is an explicit input, never a clock read.** Both
year branches divide by a span ending at the evaluation year. `Tunables`
does not carry it; it is a parameter on the album-distance call, and the
gate check in Task 6 proves no `time.Now` reaches the package.

**10. Item fields project through beets' typed defaults.** `likelies`
comes from `get_most_common_tags`, which reads thirteen fields off every
item via `item.get(field)` — and a beets `Item` returns its **typed
default** for an unset field (empty string for a string field, `0` for an
int), not nil. `derive_oracle.py`'s `build_items` constructs real `Item`s
from each case's `ItemFields`, so the corpus's records already encode that
projection. The Go `ItemFields` accessor must reproduce it; a Go port that
treats a missing field as "absent" rather than as `""` changes what
`plurality` counts and therefore what `likelies` says.

**`data_source` is the one exception, and it is not a counterexample**
(fold, m6). It lives in `Item._types`, not `Item._fields`
(`library/models.py:66-70`), so an unset read returns `None`
(`dbcore/db.py:452-468`) rather than `""`. The outcomes are identical
through both consumers — `add_data_source` treats `None` as unset, and
`plurality` counts `None` as its own value — so the port may model it
either way, but **it is not the example a typed-default test should
use**. Also: `get_most_common_tags` filters falsy items before counting
(`values = [item.get(field) for item in items if item]`,
`util/__init__.py:838`), which the port reproduces.

**11. `plurality` is `Counter.most_common(1)`: ties break by first-seen
insertion order**, i.e. by item order, which is matcher input. The Go port
counts in item order, records each value's first-seen index, and picks the
highest count with the lowest first-seen index. An albumartist consensus
(every item agreeing, non-empty) overrides `likelies["artist"]`.

**12. Bit-identity is the target; 1e-12 relative is the comparison.**
Every accumulation is IEEE-754 double `+`, `×`, `÷` with no
transcendentals, and identical decimal literals parse to identical doubles
in both languages. **A tier-0 row that passes at 1e-12 without being
bit-identical is a written finding naming the component and the suspected
operation order, not a silent pass.** Tier 0's own comparison is exact
equality first, with the tolerance as a reported fallback.

---

## Task 0: The staleness pre-task

Model: `sonnet`, **fresh context**, dispatched before any implementer.
Reads this plan and the rung 3 spec against the tree as shipped. **Its
output is blockers with evidence, not fixes**; it writes no code and edits
no file. The conductor rules each blocker before Task 1 dispatches.

The checklist is the spec's own, per pass. The first three are blockers: a
"no" stops the launch rather than producing a note.

1. **Have Phase 0a, 0b, and 0c landed?** 0a's `(5,3)` and `(8,3)` cells
   clear rung 3 to run out of ladder order; 0b owns the wall the four knobs
   register on; all three gate the second code lane. *(Expected at drafting
   time: 0a landed at `1b4ed9d`/`8545e71`/`0833589`; 0b and 0c unlanded.
   0b unlanded is **not** a blocker for 3a under ruling C's fallback — it
   is the reason the four-knob registration is deferred — and Task 0 should
   report it as the expected state rather than as a stop.)*
2. **Has rung 1b's registry shipped, and is the registration seam what this
   plan calls?** If not — the expected answer — confirm that the deferral
   is recorded and that `internal/config` is untouched by this plan.
3. **Is the lane registry consistent with where this pass intends to run,
   and is the main tree actually free?** The live-executor check is the
   registry, never `pgrep`.
4. **Have the FIVE contested-surface rows been written into the parallel
   spec's table by the conductor?** Lane-launch checklist item 2 requires a
   written owner for every surface a lane touches, and this pass touches
   five. The four the rung 3 spec proposes:
   `internal/harness/conform`, `internal/harness/oracle`,
   `internal/harness/{cassette,stub,analysis}`, and `corpus/analysis/*.json`
   plus `corpus/cases/*`. **The fifth is `tools/unidecode-table/`**, which
   the review raised as blocker B4 because
   `parallel-build-approach.md`'s lane table gives all of `tools/` to lane
   E1. **Conductor ruling, 2026-09-04: the generator keeps its location** —
   `tools/unidecode-table/` is owned by rung 3 pass 3a as the one-time
   transliteration-table generator, and lane E1 keeps the rest of `tools/` —
   **and the row is written into the parallel spec's table before this lane
   launches**, not discovered at a merge. *(Expected at revision-2 drafting
   time: all five rows landed at `889d71f`, "Write the five
   contested-surface rows"; the fifth is
   `parallel-build-approach.md:171`. Task 0 verifies rather than assumes,
   and a missing row is a blocker.)*
5. Does every corpus artifact this plan names exist, at the path and shape
   named? Specifically: `corpus/analysis/components.json` parses as
   `analysis.ComponentTable` with **≥ 72 rows** and covers all eleven
   `componentMethods`; every committed `case.json` carries an `expected`
   block with `weights` and `thresholds`.
6. Does every harness API this plan calls have the signature named?
   `analysis.LoadComponentTable`, `analysis.ComponentRowFloor`,
   `cassette.LoadCase`, `cassette.Expected`, `cassette.Thresholds`.
7. **Have any `oracle_bugs.json` or `adjudications.json` dispositions moved
   since this plan was drafted**, and does any movement change 3a's scope?
   Report the disposition table as read. The drafting-time snapshot is
   above; the ruled-`oracle-wrong` state of `bug-ft-pattern` and
   `bug-remix-parenthetical` is what makes the `SD_PATTERNS`-as-a-value
   seam load-bearing, so a movement there is a real finding. *(The review's
   m15 caught `docs/STATUS.md`'s ruled-hypothesis line stale against
   `oracle_bugs.json`. That was **conductor-owed and is landed** at
   `889d71f`, not a plan edit; Task 0 confirms STATUS and the JSON now
   agree rather than re-reporting the staleness.)*
8. Have any `ties.json` dispositions or `proximity.json` rows moved?
   (3a grades neither, but 3d inherits both, and a movement is worth
   reporting once.)
9. Do the committed case count and the weight/threshold echo match this
   plan's arithmetic? Read `ls corpus/cases | wc -l` and the distinct
   echoed weight sets off the tree. *(Drafting time: 215 cases, one weight
   set of twenty keys, one threshold set, across all 215.)*
10. Has `internal/config` changed shape? Has `internal/engine` appeared?
11. Does `corpus/README.md`'s config-profile transform table still carry
    the two "verdict-moving" cells the M2 resolution corrects? If the
    correction already landed, Task 1's drafting half is a no-op and says so.

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
required; the task writes nothing.

---

## Task 1: The matcher's vocabulary and the tunable surface

Model: `sonnet`. Births `internal/engine/matcher`.

**Files:**
- Create: `internal/engine/matcher/tunables.go` + `tunables_test.go`
- Create: `internal/engine/matcher/info.go` + `info_test.go`
- Draft (not write): the two corrected cells of `corpus/README.md`'s
  config-profile transform table, handed to the conductor in the task return

**Behavior contract:**

- **`Tunables` is the verified-complete surface from the mechanics study's
  Q4**, as a typed struct with pinned beets' shipped defaults and **a
  citation per field**. Under `match`: `StrongRecThresh` 0.04,
  `MediumRecThresh` 0.25, `RecGapThresh` 0.25
  (`config_default.yaml:167-169`), `MaxRec`
  `{missing_tracks: medium, unmatched_tracks: medium}` (`:170-172`),
  `DistanceWeights` — **twenty keys at `:174-193`; `:173` is the
  `distance_weights:` header** (fold, m4; see the m2 resolution),
  `Preferred.Countries` `[]`, `Preferred.Media` `[]`,
  `Preferred.OriginalYear` `false` (`:194-197`), `Ignored` `[]`,
  `Required` `[]`, `IgnoredMedia` `[]` (`:198-200`), `IgnoreDataTracks`
  `true`, `IgnoreVideoTracks` `true` (`:201-202`), `TrackLengthGrace` 10,
  `TrackLengthMax` 30 (`:203-204`). Outside `match`: `SearchLimit` 5
  (`metadata_plugins.py:201`; the legacy spelling named in the doc comment
  is **`musicbrainz.searchlimit`**, the metadata-source plugin's own
  `match.searchlimit` subkey rather than beets' top-level `match:` —
  `beetsplug/musicbrainz.py:290-298` reads it and deprecates it by that
  name, fold m3), `DataSourceMismatchPenalty` 0.5
  (**the default is `MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY`
  at `metadata_plugins.py:180`, wired into the plugin config at `:202`**;
  `:190-195` is the `source_weight`-first resolution property, whose legacy
  spelling the doc comment also names — fold m2), `ExtraTags` `[]`
  (**no `config_default.yaml` citation exists**: it is the musicbrainz
  plugin's own `extra_tags: []`, `beetsplug/musicbrainz.py:286`, and the
  doc comment says so rather than leaving the one uncited field — fold m9),
  `Timid` `false` (`config_default.yaml:27`), `Languages` `[]` (`:32`),
  `IgnoredAliasTypes` `[]` (`:58`), the loaded metadata source set, and the
  verdict-to-action trio `Quiet` `false` (`:28`), `QuietFallback` `"skip"`
  (`:33`), `NoneRecAction` `"ask"` (`:34`). **Verify every line citation
  against the venv before writing it**; the study's line numbers are
  research, not a pin. Every citation other than the four corrected above
  verified exactly at review time (Verified sound; fold, m4).
- **`Tunables.MaxRec`'s Go type is `map[string]string`** (fold, m14), for
  the same reason `cassette.Thresholds.MaxRec` is
  (`cassette.go:88-93`): the value is a recommendation-level name, not a
  number, so a `map[string]float64` cannot hold it. Task 6's defaults
  cross-check compares the two directly, so the types must match.
- **The `match:` surface carries two keys `Tunables` deliberately omits**
  (fold, m10): `album_disambig_fields` and `singleton_disambig_fields`
  (`config_default.yaml:205-206`). Both are display-only, read through
  `Match.disambig_fields` (`match.py:113-142`) and never on the distance
  path. **Name them in `Tunables`' doc comment as deliberately excluded**,
  so a reader checking the "verified-complete" claim against
  `config_default.yaml` finds the answer rather than a falsification.
- **Ruling D's blanket rule holds: every default equals pinned beets'
  shipped value at launch, `languages` included.** `Languages: []` makes
  `_preferred_alias` structurally unreachable, which is half of the alias
  mis-selection hypothesis; shipping a non-empty default would be an
  improvement riding the port, which is forbidden. Record the reachability
  gap in the field's doc comment as a post-launch improvement candidate.
- **`Tunables` carries neither `original_date` nor `per_disc_numbering`**,
  per the M2 resolution above. Re-verify the resolution against the venv
  (the three greps and the two call sites named there), report the
  evidence, and draft the two corrected README cells for the conductor.
  This is a **blocking pre-task item resolved, not a note**: if the venv
  contradicts the resolution, the task stops and reports rather than
  amending `Tunables` on its own authority.
- **`SD_PATTERNS` is a value, not a literal buried in a function.**
  Exported as `BeetsSDPatterns` (the shipped six, in beets' fixed order,
  with `SD_END_WORDS` and `SD_REPLACE` beside them) and referenced through
  `Tunables.StringDistPatterns`, defaulting to it. A second exported table,
  `SDPatternsWordBoundary`, carries the guarded featuring pattern with a
  doc comment citing `adj-adv-ft-draft-colon` and upstream `19d091440`, and
  is **used by nothing in 3a**: it exists so 3b's `defect_free` derivation
  and 3d's compat switch swap a table rather than edit a component. Task 4
  implements against `Tunables.StringDistPatterns` and never against the
  package-level table directly.
- **The four wall knobs are named and not registered.** A doc comment on
  `Tunables` names the four (`StrongRecThresh`, `MediumRecThresh`,
  `RecGapThresh`, `SearchLimit`) as the operator-facing subset, states that
  registration lands through rung 1b's registry in a later rung 3 pass and
  never by editing `internal/config/keys.go`, and records the rest as
  deliberately not operator-facing at launch. **No `internal/config` import
  appears in this package.**
- **The three surfaces no corpus fixture can reach** — `match.required`,
  `match.ignored`, `preferred.original_year` — are named in `Tunables`'
  doc comments as **graded by table-driven component unit tests in the
  package that owns them**, with the fixture hand-authored in the package's
  own `testdata` on the proximity precedent (spec, M8 fold).
  `preferred.original_year` lands in Task 7's year component; `required`
  and `ignored` land in 3d. Their absence from tier 1 is a **declared
  corpus-coverage gap** carried into the pass report, never a passing check.
- **`info.go` is the matcher's vocabulary**: `AlbumInfo`, `TrackInfo`, and
  `ItemFields`, carrying exactly the fields the distance path reads —
  album: `Artist`, `Album`, `VA`, `Media`, `Mediums`, `Year`,
  `OriginalYear`, `Country`, `Label`, `CatalogNum`, `AlbumDisambig`,
  `AlbumID`, `DataSource`, `Tracks`; track: `Title`, `Artist`, `Length`,
  `Index`, `MediumIndex`, `Medium`, `TrackID`, `DataSource`. **No applied
  field, no path, no tag I/O**: rung 3 computes no applied field and writes
  no file. 3b's translation layer is what populates these; 3a defines them
  and their zero values.
- **`ItemFields` is enumerated, not gestured at** (fold, M-d). The distance
  path reads **eighteen** item fields, and the struct carries exactly those
  with beets' own types:
  - the thirteen `get_most_common_tags` reads, in beets' order
    (`util/__init__.py:822-836`): `artist` (string), `album` (string),
    `albumartist` (string), `year` (int), `disctotal` (int), `mb_albumid`
    (string), `label` (string), `barcode` (string), `catalognum` (string),
    `country` (string), `media` (string), `albumdisambig` (string),
    `data_source` (string, and the `_types`-not-`_fields` exception of
    contract 10's m6 note);
  - the seven `track_distance` reads (`distance.py:394-424`): `length`
    (float), `title` (string), `artist` (string — the same field as the
    first list's, counted once), `track` (int), `mb_trackid` (string),
    `disc` (int), `data_source` (string — likewise counted once).
  Eighteen distinct fields, and `artist` and `data_source` are the two the
  two lists share. A field outside this set is not read by the distance
  path, and adding one is a finding rather than a convenience.
- **`ItemFields` projects through beets' typed defaults** (contract 10).
  Ship the accessor with a table-driven test proving an unset string field
  reads `""` and an unset int field reads `0` — **`data_source` is not the
  field to use as the example** (fold, m6). **The fixture is hand-authored
  in this package's own `testdata`; this task reads no corpus case**
  (fold, M-d): the cross-check that the projection matches what a committed
  case's `Input.Items` map produces for an omitted field is **Task 6's**,
  where reading `corpus/` is the harness's own business. This test is the
  one with teeth: report the mutation (an accessor returning a sentinel for
  "absent") that proves it fires.
- **Every exported identifier carries a Go Doc Comment** that states intent
  and cites its beets source; vale gates the prose.

**Deliverable count: 4** (`Tunables` with its roughly twenty-six fields as
enumerated above — not "thirty-odd", fold m9 — each
default carrying a verified citation, and the four-knob deferral recorded;
the `SD_PATTERNS`-as-a-value seam with its second table for 3b/3d; `info.go`'s
vocabulary, including `ItemFields`' eighteen enumerated fields with beets'
types and the typed-default projection proved by a hand-authored firing
fixture; the M2 re-verification evidence and the two drafted README cells).
Gate green; commit. **Report actual token spend.**

---

## Task 2: The accumulator

Model: **`opus`**. The upshift is deliberate: this is novel,
correctness-critical arithmetic on which every later number in the rung
depends, and its two hardest properties — the dual penalty views and the
denominator's add-count rule — are exactly the ones a plausible-looking
implementation gets silently wrong. Depends on Task 1.

**Files:**
- Create: `internal/engine/matcher/penalty.go` + `penalty_test.go`
- Create: `internal/engine/matcher/testdata/` fixtures for the firing proofs

**Behavior contract:**

- **`Distance` reproduces beets' `Distance` object**: an insertion-ordered
  map of key to penalty slice, plus the per-track sub-distances the album
  component fills in (`tracks`). Insertion order is preserved explicitly
  (a key slice beside the map), because `raw_distance` and `max_distance`
  sum over it and Go map iteration is randomized.
- **`tracks`' key type is stated here, not assumed by Task 7** (fold,
  m13). Python's `dist.tracks` is `dict[TrackInfo, Distance]`
  (`distance.py:131`, filled at `:527-530`), keyed by **object identity**,
  because `TrackInfo` defines no `__hash__` of its own. Go has no value
  equivalent: `TrackInfo` is a struct and using it as a map key would key
  by value and collide on two identical candidate tracks. Ship
  `map[*TrackInfo]*Distance` **or** an index-keyed slice parallel to the
  pairing — the implementer picks, states the choice in the doc comment,
  and the choice is load-bearing for Task 7's `tracks` component. A
  value-keyed map is wrong; say so in the comment.
- **Eight adders**, exactly as contract 6 specifies: `Add`, `AddEquality`,
  `AddExpr`, `AddNumber`, `AddPriority`, `AddRatio`, `AddString`,
  `AddDataSource`. **`AddExpr` is in the list** (fold, m4) because
  `add_expr` is in `componentMethods` with two committed rows, and tier 0
  will demand the primitive.
- **`Add` rejects a penalty outside `[0,1]`.** beets raises; the Go port
  returns an error or panics with a message naming the key and the value —
  choose the one `go-conventions` supports for a programming error in an
  internal package, and state the choice in the doc comment. It is not
  silently clamped.
- **The three aggregates** (`RawDistance`, `MaxDistance`, `Distance`) with
  the **zero-max special case** returning `0.0`, per contract 5.
- **Both views, under distinct names, each with its own unit test** (port
  hazard 4). `AddedKeys()` / `PerKey()` — the `_penalties` view, including
  keys added at `0.0`, which is what `cassette.Candidate.Penalties` grades
  — and `NonzeroItems()` / `NonzeroKeys()` — the `items()`/`keys()` view,
  nonzero only, sorted by `(-weighted, key)`, which is what `max_rec` reads
  at 3d. The doc comment on each names the other and says which consumer
  reads it, so a 3d implementer cannot pick the wrong one by accident.
- **`GenericPenaltyKeys`** reproduces beets' derived view: for each key
  with a **nonzero** weighted distance, `strings.ReplaceAll` of `album_`
  → ``, then `track_` → ``, then `_` → ` `, in that order. The
  zero-penalty exclusion is part of the contract and gets its own
  hand-authored case.
- **`AddNumber`'s multi-unit behavior is a named test**, not an incidental
  consequence: one key, several `1.0` penalties, a denominator inflated by
  weight per unit.
- **`AddDataSource`'s three-way condition is a named test**: match →
  nothing added at all (not even `0.0`, so the key stays out of the
  denominator); mismatch with a `before` → the penalty; no `before` with a
  single source → nothing; no `before` with multiple sources → the
  penalty.
- **Regex options anchor, and only options do** (contract 7).
  `AddEquality` and `AddPriority` accept an option that is either a literal
  value or a compiled anchored pattern; the type that expresses that choice
  is the implementer's, and the doc comment states that a pattern
  **matches from position 0** and cites Python's `.match`. **Both adders
  wrap a scalar option into a one-element list before iterating**
  (contract 6a), and `AddPriority`'s unit divisor is computed after the
  wrap.
- **Assertion discipline, against this task's own hand-authored fixtures**
  (fold, M-b). Every arithmetic property above ships a table-driven test,
  and the task reports **the mutation that proved each can fire**. The four
  named mutations, each stated against a fixture this task authors rather
  than a committed row it may not read:
  1. a denominator that counts a fixed weight table instead of added
     penalties — must fail the fixture that adds one key at `0.0` beside
     two scored keys, where the `0.0` key still enlarges the denominator;
  2. a nonzero view that includes a `0.0`-weighted key — must fail the
     fixture whose only key carries weight `0.0`, whose `items` view is
     **empty** while its added-key view is not;
  3. a sort keyed on `(dist, key)` instead of `(-dist, key)` — must fail
     the two-key fixture whose weighted distances differ, and the
     equal-distance fixture that pins the key tiebreak;
  4. a zero-max path returning `NaN` instead of `0.0`.
  The committed rows that pin the same four properties
  (`aggregate/zero-added-key`, `aggregate/all-zero-weights`,
  `aggregate/sorted-penalties`, `aggregate/sorted-penalties-tied`,
  `aggregate/empty`) are **Task 6's** mutation list, not this task's.
- **This task does not read `components.json`.** Its tests are
  hand-authored from the contracts; Task 6 is what replays the committed
  rows. Two independent expressions of the same arithmetic is the point:
  if they disagree, one of them is wrong and the gate says so.

**Deliverable count: 4** (the insertion-ordered `Distance` with its three
aggregates and the zero-max case; the eight adders at contract 6's exact
semantics; both penalty views plus `GenericPenaltyKeys`, each with its own
test and its cross-reference doc comment; the firing-fixture suite with the
four named mutations reported). Gate green; commit. **Report actual token
spend — this is the figure the 6M ceiling is re-derived from.**

---

## Task 3: The transliteration table

Model: `sonnet`. **Named escalation trigger in place of an upshift:** two
failed attempts to produce a Go table that reproduces the pinned
`unidecode`'s output over a full codepoint sweep stops the task and reports,
rather than spending a third. Depends on Task 1 (for the package), not on
Task 2.

**Why this is its own task.** `_string_dist_basic` transliterates through
`unidecode` before stripping to `[a-z0-9]`, and the corpus reaches well
past Latin-1: the real-catalog stratum carries non-Latin cases (X Japan,
Utada, Alla Pugacheva) whose titles transliterate through CJK and Cyrillic
tables. Go has no `unidecode` in its standard library, and **no third-party
port is admissible**, because fidelity to a *specific* `unidecode` version
is the whole requirement and a port tracking a different upstream revision
diverges silently on exactly the characters nobody tests. The corpus
manifest pins `unidecode` `1.4.0` precisely so this is checkable.

**Files:**
- Create: `tools/unidecode-table/generate.py` (the one-time generator, ruff-gated, never invoked by the gate or `make derive`)
- Create: `internal/engine/matcher/translit.go` + `translit_test.go`
- Create: `internal/engine/matcher/unidecode-1.4.0.tbl` (the generated, embedded table)

**Behavior contract:**

- **The table is generated from the pinned venv, and the generator is
  committed as its provenance.** 2d's `template.flac` authoring script is
  the precedent: the script runs once, its output is the artifact, its
  docstring says so, and it is ruff-gated like every other tracked `*.py`.
  It lives under `tools/` rather than `internal/harness/oracle` because
  this plan writes no oracle territory.
- **Sized honestly, measured at drafting time:** `unidecode` 1.4.0 ships
  190 block files carrying **48,597 entries**, of which 41,379 are
  non-empty and 5,477 are `None`, for **157,450 bytes of mapped output**.
  A block-indexed table plus a concatenated string arena is a few hundred
  kilobytes embedded through `go:embed` — cheap to ship, cheap to compile,
  and exact. A giant Go source literal is not the shape; an embedded data
  blob decoded once at package init (or through `sync.OnceValue`) is.
- **`None` and empty are different.** `unidecode` distinguishes an
  unmapped codepoint from one that maps to the empty string; the table
  encodes both, and the decoder reproduces `unidecode`'s own behavior for
  each (including the >0xeffff cutoff and the surrogate case). Read the
  pinned package's `__init__.py` rather than assuming.
- **The fidelity proof is a sweep, not a spot check.** The generator emits,
  beside the table, a **verification digest**: for every codepoint in the
  table's range, the transliteration `unidecode` produced. `translit_test.go`
  recomputes the digest from the embedded table and compares. If the
  digest is too large to commit, commit the hash and the sweep parameters
  and state that in the test's doc comment — the property that matters is
  that a table edit or a regeneration against a different `unidecode`
  fails loudly.
- **The table carries its own exported version stamp; the manifest
  comparison lands at Task 6** (fold, M-c). This task exports the stamp —
  the `unidecode` version the generator ran against, recorded in the table
  and surfaced from package `matcher` — and tests it against the embedded
  table locally. **It does not read `corpus/manifest.json`**: every other
  engine task is barred from reading the corpus, `cassette.Manifest` is a
  harness type, and importing it here would trip the very architecture
  check Task 6 lands. The comparison of the stamp against
  `corpus/manifest.json`'s `unidecode` field is a **conform-side test at
  Task 6**, where reading the corpus is the harness's own business, and it
  is Task 6 that reports the mutation proving it fires (a stamp of `1.3.9`
  against the manifest's `1.4.0`). A `unidecode` bump is a
  corpus-revalidation event; the seam is split across two packages so that
  the check exists without inverting the dependency.
- **Levenshtein rides along.** `jellyfish.levenshtein_distance` over the
  stripped ASCII strings is plain integer edit distance with unit costs;
  implement it in this file or in `strdist.go` at Task 4's discretion, with
  a table-driven test including the empty-string and equal-length cases.
  No third-party package.
- **If gitleaks flags the generated table**, that is a finding to report,
  and the resolution is a `.gitleaksignore` **append** naming the finding
  in the pass report (per Phase 0a's `(5,2)` row) — never a rule disable,
  and never an edit that inserts a line above
  `internal/config/keys_test.go:66`. **The file is conductor-owned**
  (fold, m11, and the Global Constraints bullet above): this task drafts
  the appended line and the conductor lands it in the batch commit.

**Deliverable count: 4** (the committed one-time generator with its
provenance docstring; the embedded table with its `None`/empty
distinction; the full-sweep fidelity proof; the exported version stamp
whose manifest comparison lands at Task 6, plus Levenshtein). Gate green;
commit.

---

## Task 4: `string_dist` and `SD_PATTERNS`

Model: **`opus`**. The second deliberate upshift: the pattern loop's
baseline-shift arithmetic is the subtlest code in the pass, it is the
surface two ruled `oracle-wrong` hypotheses live on, and every string
component's number flows through it. Depends on Tasks 1, 2, and 3.

**Files:**
- Create: `internal/engine/matcher/strdist.go` + `strdist_test.go`
- Create: `internal/engine/matcher/testdata/` fixtures for the hazard proofs

**Behavior contract:**

- **`stringDistBasic`**: transliterate each input, lowercase, strip every
  character not in `[a-z0-9]`; both empty after stripping → `0.0`; else
  `levenshtein(a, b) / max(len(a), len(b))` as a float division. Note that
  length is measured in **bytes of the stripped ASCII string**, which after
  stripping is the same as runes — say so in the doc comment so a later
  reader does not "fix" it to `utf8.RuneCountInString`.
- **`StringDist`** implements contract 1's pipeline in order: nil handling,
  lowercase both, article rotation over `SD_END_WORDS` (a trailing
  `", <word>"` becomes `"<word> <prefix>"`, applied to each string
  independently and to **each of the three words in order**), `&` → `and`,
  then the pattern loop, then `base + penalty`.
- **`SD_PATTERNS` are never anchored** (contract 7, fold M-f). The pattern
  loop is `re.sub` in Python — unanchored, all-non-overlapping deletion
  over the whole string — and the Go port is `Regexp.ReplaceAllString` on
  the pattern **exactly as beets writes it**. Do not add `^`, do not
  translate the loop into a match-at-position-0 test, and do not reuse
  whatever anchoring helper `AddEquality`'s options need: contract 7's
  anchoring rule belongs to `Distance._eq`'s options and to nothing else.
  The only `^` in the table is beets' own, in the `^the ` literal.
- **`BeetsSDPatterns` is byte-identical to beets' six shipped literals, and
  a test says so** (fold, M-a). Tier 0 **cannot** catch a well-meaning
  "fix" to the table: applying the word-boundary guard named in
  `adj-adv-ft-draft-colon` changes zero of the 72 committed rows, so the
  corpus is blind to exactly the edit most likely to be made. The guard is
  therefore explicit and lives here:
  1. a test asserting each of the six pattern strings and its weight is
     byte-identical to the literal in
     `beets/autotag/distance.py:35-42` at the pinned version. The six
     literals are written into the test itself, so an edit to
     `Tunables`' table has to be made twice to pass and cannot drift
     silently;
  2. a **discriminating fixture** that separates the two tables:
     `string_dist("Draft: Ember Wharf", "Draft: Northern Lines")` is
     **`0.061111111111111116`** under `BeetsSDPatterns` and
     **`0.6111111111111112`** under `SDPatternsWordBoundary`, both measured
     against pinned beets. The test asserts the first against the shipped
     table and the second against the guarded one, which is what makes
     `SDPatternsWordBoundary` a tested value rather than dead weight.
  **Both figures travel in the pass report**, because 3b derives each
  `defect_free` figure through pinned beets with the one named
  `SD_PATTERNS` edit and this is the first measured pair it can check
  itself against.
- **The pattern loop is the load-bearing part.** For each `(pattern,
  weight)` in `Tunables.StringDistPatterns`, in order: delete every
  non-overlapping match from **both** strings; if **either** changed,
  recompute `caseDist` on the stripped pair; `caseDelta = max(0,
  baseDist - caseDist)`; **if the delta is zero, continue without adopting**;
  otherwise **adopt the stripped strings as the new baseline**, set
  `baseDist = caseDist`, and add `weight × caseDelta` to the running
  penalty. The adopt-only-on-positive-delta rule is what makes the six
  patterns non-commutative, and it is the single most likely thing to get
  wrong.
- **Port hazard 1 is proved, not assumed.** A test drives a pair whose
  result differs under lower→transliterate→lower versus
  transliterate→lower alone, and reports it as the mutation that proves
  the ordering is real.
- **Port hazard 2 is verified against pinned beets and recorded in this
  task's report** (fold, m6): the study does not make this claim, so 3a
  makes it. The evidence is a grep over `beets/autotag/` and the MB plugin
  for `unicodedata`, `normalize`, `NFC`, `NFD` returning nothing on the
  distance path, plus a test pairing a composed and a decomposed form of
  the same string and asserting the Go result equals what pinned beets
  produces for it. **The Go implementation must not import
  `golang.org/x/text/unicode/norm` or call any normalization**, and the
  test is what stops a later "helpful" fix.
- **Go/Python case-mapping difference is checked once and reported.**
  Python's `str.lower()` is full case mapping (which can expand one rune
  into several); Go's `strings.ToLower` is simple case mapping. The strip
  to `[a-z0-9]` absorbs most of the difference, but not by construction.
  Sweep the case-mapping-divergent codepoints against the venv, confirm the
  post-strip results agree, and **report the sweep's size and outcome**. A
  disagreement is a finding and a stop, not a silent tolerance.
- **`components.json`'s 23 `string_dist` rows are the arithmetic's own
  acceptance**, but this task does not read the file — Task 6 replays it.
  This task's tests are hand-authored from the pipeline, including at
  minimum: the article rotations for all three words; `&` → `and`; the
  `(EP)` zero-weight pattern; the featuring discount **as beets computes
  it, defect included**; the parenthetical and bracket discounts; the
  `part`/`pt.` pattern; and the **three** solo cases that pin the loop's
  behavior when one side empties — `"The "` vs `""`, `"(EP)"` vs `"(EP)"`,
  and `", An"` vs `""` (fold, m8: `string_dist/solo/*` is three rows, and
  the `(EP)` pair is one of them rather than a separate item).
- **Regex compilation happens once**, at package scope or through
  `sync.OnceValue`, never per call, and the patterns compile at init so a
  malformed table is a startup failure rather than a per-call error.

**Deliverable count: 5** (`stringDistBasic` with its strip and empty-pair
rules; `StringDist`'s full pipeline including the article rotation and the
`&` rewrite; the pattern loop with adopt-on-positive-delta, unanchored
`ReplaceAll` semantics, and its hand-authored acceptance suite; the
byte-identity table test and the discriminating Draft fixture with both
measured figures; port hazards 1 and 2 proved and the case-mapping sweep
reported). Gate green; commit. **Re-derive the ceiling against Task 2's
figure and report both.**

---

## Task 5: The track-level components and `track_distance`

Model: `sonnet`. Depends on Tasks 1, 2, and 4.

**Files:**
- Create: `internal/engine/matcher/track_length.go`, `track_title.go`,
  `track_artist.go`, `track_index.go`, `track_id.go`, `track_medium.go`,
  `track_data_source.go`, each with its own `_test.go`
- Create: `internal/engine/matcher/track.go` (the fixed-order composition)

**Behavior contract:**

- **One component, one file, one test** — the evolvability ruling's whole
  point, and the spec's own tree. A component is a name, a weight key, and
  a function taking the pair plus the tunables and writing into the
  accumulator.
- **The fixed order is contract 8's**, and the composition iterates a
  declared slice rather than calling seven functions inline, because Task 8
  composes the album side the same way and the two must be uniform.
- **Each component's fire condition is its own hardest part**, and each
  gets a test proving the component **declines** as well as one proving it
  adds:
  - `track_length` (2.0): only when `track_info.Length` is truthy;
    `diff = |item.Length − info.Length| − TrackLengthGrace`, then
    `AddRatio("track_length", diff, TrackLengthMax)`. Negative diff clamps
    to 0; diff beyond max clamps to 1.
  - `track_title` (3.0): `AddString` **always**.
  - `track_artist` (2.0): only when `inclArtist` **and** `track_info.Artist`
    **and** `strings.ToLower(item.Artist)` is not in `VA_ARTISTS`
    (`""`, `"various artists"`, `"various"`, `"va"`, `"unknown"`).
  - `track_index` (1.0): only when `info.Index` and `item.Track` are both
    truthy; `AddExpr` of `item.Track ∉ {info.MediumIndex, info.Index}` —
    tolerating either per-disc or absolute numbering.
  - `track_id` (5.0): only when `item.MBTrackID` is set; `AddExpr` of
    inequality against `info.TrackID`.
  - `medium` (1.0): only when `info.Medium` and `item.Disc` are both
    truthy; `AddExpr` of inequality.
  - `data_source` (2.0): `AddDataSource(item.DataSource, info.DataSource)`,
    at contract 6's three-way condition.
- **`inclArtist` is a parameter, not a config read.** The assignment solve
  calls `track_distance` with `inclArtist=false` while the final per-track
  distances pass `album_info.va`; the two objectives diverge exactly on VA
  albums (study, "Track assignment"). 3a does not build the solve, but the
  parameter must exist and be documented now, because 3d's tie-mirroring
  work depends on the distinction being explicit rather than rediscovered.
- **The `components.json` `track_distance` rows are 16** and cover the whole
  data-source parameter space plus `identical`, `tolerate-va`,
  `different-artist`, `different-title`. This task does not read them;
  Task 6 does. Hand-author the equivalents.

**Deliverable count: 4** (seven component files each with its own
fire-and-decline test; the fixed-order composition through a declared
slice; the `inclArtist` parameter with its VA-divergence doc comment; the
data-source three-way behavior proved in all four states). Gate green;
commit.

---

## Task 6: The typed `components.json` decoder, tier 0, and the architecture checks

Model: `sonnet`. **Runner-up upshift candidate**: if Tasks 2 or 4 report
that an arithmetic contract above was under-specified, the conductor
upshifts this task, because tier 0 is the instrument every later claim in
the pass rests on. Depends on Tasks 1 through 5. **Births
`internal/harness/conform`.**

**Files:**
- Create: `internal/harness/conform/doc.go` (the package comment naming the
  harness→engine direction and why the runner lives here)
- Create: `internal/harness/conform/component.go` + `component_test.go`
  (the typed decoder and the tier-0 replay)
- Create: `internal/harness/conform/defaults.go` + `defaults_test.go`
  (the M1 cross-check in the form that can fire today, plus the
  `unidecode` manifest pin check and the `ItemFields` corpus cross-check
  the engine tasks are barred from making)
- Create: `internal/harness/conform/arch_test.go` + `testdata/` (the
  engine-never-imports-harness and no-`time.Now` checks with their firing
  fixtures)
- Create: `internal/harness/conform/testdata/` fixtures for the decoder's
  own firing proofs

**Behavior contract:**

- **Tier 0's home is `internal/harness/conform`, decoder and runner both**
  (fold, M5). The typed decoder reads a harness artifact and drives engine
  primitives, which is the harness→engine direction the first architectural
  invariant requires; putting it in the engine would invert that dependency
  for a test-only convenience.
- **The decoder dispatches on `method`, and inside method `add` on the
  argument shape: eleven methods, twelve shapes** (fold, B1). The row
  **id** is a label, not a discriminator, and reading it as one is what
  revision 1 got wrong: the seven `aggregate/*` rows carry
  `"method": "add"`, because `derive_component.py:294` and `:312` emit
  `row(row_id, "add", …)` for them. The committed method counts, recomputed
  off the tree, are `add` **8** (`add/basic` plus the seven `aggregate/*`),
  `add_data_source` 4, `add_equality` 3, `add_expr` 2, `add_number` 4,
  `add_priority` 4, `add_ratio` 4, `add_string` 3, `generic_penalty_keys`
  1, `string_dist` 23, `track_distance` 16 — 72 in total. So method `add`
  carries **two incompatible shapes**, and the decoder splits them on
  `args.key` (the adder shape) versus `args.weights` (the aggregate shape).
  A row under method `add` carrying neither, or both, is a failure naming
  the row id. Twelve shapes across eleven methods, in six families —
  seven methods share the adder family, and the other five each have their
  own. Read off the committed table and off `derive_component.py` at
  revision-2 drafting time:
  - **the adder shape** — method `add` with `args.key`, plus
    `add_equality`, `add_expr`, `add_number`, `add_priority`, `add_ratio`,
    `add_string`: args are `{"key": "<name>", "calls": [[...], ...]}`,
    output `{"penalties": [float, ...]}`. Each row is a **replay of the
    full call prefix against a fresh accumulator** (`derive_component.py:241-254`
    rebuilds `Distance()` per row and replays `calls[0..i]`), and the
    output is that key's penalty slice in append order. **`args.key` is
    the literal penalty key the adder was called with, not a beets weight
    key**: the adder tables use `"equality"`, `"expr"`, `"number"`,
    `"priority"`, `"ratio"`, `"string"`, and `add/basic` uses `"add"`.
    None is in `distance_weights`, and none needs to be — the adder shape's
    output is unweighted;
  - **the aggregate shape** — method `add` with `args.weights`: args are
    `{"weights": {...}, "calls": [[key, dist], ...]}`, output
    `{"distance", "raw_distance", "max_distance", "per_key", "items"}` —
    the aggregates and **both views**, which is what makes tier 0 the place
    port hazard 4 is caught for the price of a unit test;
  - **`generic_penalty_keys`**: args are the aggregate shape's
    `{"weights", "calls"}`, output `{"keys": [string, ...]}`;
  - **`string_dist`**: args are a **bare two-element JSON array** of
    strings — not an object — either of which may be `null`; output
    `{"distance": float}`;
  - **`track_distance`**: args are `{"incl_artist", "item", "track_info"}`
    plus optional `weights`, `data_source_mismatch_penalty`, and
    `multiple_data_sources`; output `{"distance", "penalties"}` where
    `penalties` is a map of key to **raw penalty slice**, not weighted
    values;
  - **`add_data_source`**: args are `{"before", "after", "weights",
    "data_source_mismatch_penalty", "multiple_data_sources"}`, where
    `before` and `after` may each be `null`; output `{"penalties": {...}}`
    — **an empty object when nothing was added**, which is the case that
    proves the key stays out of the denominator.
  **Read every shape off the committed table at execution time**; the list
  above is a drafting-time reading, not a pin. A row whose method the
  decoder does not know, or whose shape it cannot resolve, is a **failure
  naming the row id**, never a skip.
- **A row's `weights` map is an OVERRIDE on `Tunables.DistanceWeights`,
  not a complete weight table** (fold, B2). `derive_component.py:168-190`
  is `Distance._weights.update(overrides)` with a restore afterward, so
  **every key the row does not name keeps beets' shipped default**. Two
  committed rows prove it decisively, and both are in the mutation list
  below:
  - `aggregate/normalized` supplies `{album: 4, data_source: 2, medium: 2}`
    and then adds two **`media`** penalties. Its `max_distance` of `6` is
    `1×4 + 2×1.0`, which needs `media`'s shipped default of `1.0`. A
    decoder using the row's map verbatim computes `3/4 = 0.75` against a
    committed `0.5`, and fails with no named cause;
  - `track_distance/data_source/mismatch-high-weight` supplies
    `{data_source: 5}`; its committed `distance` of `0.3125` is
    `2.5/8`, which needs `track_title`'s shipped `3.0`.
  One key in the table is not a beets weight at all: `zero_weighted`,
  introduced by `aggregate/all-zero-weights` and **removed again rather
  than restored** by the shim, exists only for that row and only to make a
  zero denominator reachable. The decoder therefore layers the row's map
  over `Tunables.DistanceWeights` key by key and tolerates a key beets does
  not ship.
- **`get_penalty` and the two-plugin stub are modeled explicitly**
  (fold, B3). Sixteen rows — the four `add_data_source/*` and the twelve
  `track_distance/data_source/*` — are unreplayable without it, because
  beets resolves the mismatch penalty through
  `metadata_plugins.get_penalty(after)` (`metadata_plugins.py:159-169`),
  which scans the **loaded** metadata source plugins for one whose
  `data_source` name equals `after` and falls back to
  `MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY` = `0.5`
  (`:180`). The derivation's stub (`derive_component.py:193-226`) installs
  exactly two plugins, named so that `data_source` derives from the class
  name: **`OriginalPlugin`**, carrying the class default `0.5`, and
  **`OtherPlugin`**, carrying the row's own
  `data_source_mismatch_penalty`. When the row's `multiple_data_sources`
  is `false`, **only `OtherPlugin` is loaded**. So the decoder builds the
  row's loaded-source set as `{"Original": 0.5, "Other": <row penalty>}`,
  or `{"Other": <row penalty>}` when `multiple_data_sources` is false, and
  resolves `after` against it with `0.5` as the fallback. Three committed
  rows pin the fallback and the single-source branch:
  `track_distance/data_source/mismatch-unknown` (`after` = `"unknown"`,
  no such plugin, `0.5`), `.../mismatch-no-info` (`after` = `null`,
  likewise `0.5`), and `.../match-no-original-but-single-source`
  (`before` = `null`, one source loaded, **nothing added**).
  **This wires through `Tunables`' loaded metadata source set**, the field
  Task 1 already puts on the struct: the decoder populates it per row
  rather than inventing a parallel mechanism, which is also what makes the
  field's real shape testable before 3c has a client.
- **A compiled-regex option is encoded as an object** (fold, M-e).
  `derive_component.py:35-51` serializes a `Pattern` as
  `{"regex": "<pattern source>", "ignorecase": <bool>}`, and the decoder
  must recognize it wherever an `add_equality` or `add_priority` option
  can appear — `add_equality/2`'s third call and `add_priority/2`'s fourth
  option are the committed instances. `ignorecase: true` compiles as
  `(?i)`, and the option **anchors at position 0** per contract 7. An
  option that is not such an object is a literal value, and a **scalar
  option is wrapped into a one-element list** before iterating
  (contract 6a): `add_priority/2` passes a scalar first and a list second,
  so the decoder cannot assume either form.
- **Conform owns its own decoder-method set** (fold, M-g).
  `analysis.componentMethods` is **unexported**
  (`internal/harness/analysis/component.go:23-35`) and behind the territory
  wall, so this task neither reads it nor exports it, and the "one derived
  fact, one home" argument that works for `ComponentRowFloor` cannot work
  here. Table-side coverage is already enforced on the analysis side by
  `analysis.CheckComponentTable`, which rejects a table naming a method
  outside its list. **Conform's own obligation is the mirror of it, stated
  as two assertions**: every method the decoder knows has **at least one
  replayed row**, and every row's method is one the decoder knows. A
  decoder method with zero rows is a failure naming the method — that is
  what catches a table that quietly loses a method — and an unknown row
  method is a failure naming the row id. The two lists being separately
  maintained is the point: they are two independent statements of the same
  fact, and a disagreement is a gate failure rather than a silent skip.
- **The row floor is enforced through `analysis.ComponentRowFloor`**, not a
  second copy of 72: one derived fact, one home. A table below the floor
  fails.
- **Every one of the 72 rows replays green.** Comparison is **exact float
  equality first**; a row that passes only at 1e-12 relative is reported as
  a written finding naming the component and the suspected operation order
  (contract 12), and the pass report carries the list — empty is the
  expected answer and is worth stating.
- **A skipped row is a failure.** The runner reports the count it actually
  replayed, per method, and the test fails if that count is not the table's
  row count. This is the "instrument that can pass silently" mitigation in
  its concrete form.
- **The M1 defaults cross-check, in the form that can fire today.** Tier 1
  is 3d's, but the cross-check the M1 fold names — "the echoed `Weights`
  and `Thresholds` must equal the merged result, for every case" — has a
  non-empty overlap **now**, because all three committed profiles set no
  `match:` keys and every committed record echoes beets' defaults. So this
  task lands the half that needs no tier-1 runner: a gate-resident test
  reading every committed `case.json` (read-only) and asserting that its
  echoed `Weights` equals `Tunables`' twenty defaults and its echoed
  `Thresholds` equals the three thresholds and `max_rec`. **It fires on any
  defaults transcription error in `Tunables`**, which is the check actually
  worth having, and it names the key and both values when it does. 3d
  extends it with the profile merge; 3a proves the defaults.
- **The `unidecode` manifest pin check lands here, not in the engine**
  (fold, M-c). Task 3 exports the transliteration table's own version
  stamp from package `matcher`; this task reads `corpus/manifest.json`'s
  `unidecode` field (read-only, through `cassette.Manifest`) and fails if
  the two differ, naming both values. Reading the corpus from the harness
  is exactly the permitted direction; reading it from the engine would have
  tripped this task's own architecture check. **Report the mutation that
  proved it fires** (a stamp of `1.3.9` against the manifest's `1.4.0`).
- **The `ItemFields` corpus cross-check lands here too** (fold, M-d). Task
  1 proves the typed-default projection against its own hand-authored
  fixture; this task proves the same projection agrees with the corpus, by
  reading a committed `case.json` and asserting that a field its
  `Input.Items` map omits reads as the beets typed default through the Go
  accessor rather than as an absent value. Same reason, same direction:
  the engine defines the projection, the harness checks it against the
  data.
- **The architecture checks, stdlib only.** `arch_test.go` walks
  `internal/engine/...` with `filepath.WalkDir` and `go/parser` in
  `ImportsOnly` mode and fails on any import path under
  `github.com/glw907/dubplate/internal/harness`; a second walk fails on any
  `time.Now` selector inside `internal/engine/matcher`. Both run against a
  `testdata` fixture tree carrying a violating file, and **the task reports
  both mutations**.
- **Every checker in this task ships a fixture that makes it fire**, and
  the report names each mutation. At minimum: a row whose `Output` is
  perturbed by one ULP (the replay must fail); a row whose method is
  unknown (must fail, not skip); a decoder method with no replayed row
  (must fail naming the method); a table one row below the floor; a
  `Tunables` weight perturbed from `0.9` to `0.09` (the defaults
  cross-check must name `missing_tracks` and both values); a stamp of
  `1.3.9` against the manifest's `1.4.0`; a violating import; a `time.Now`
  call. **Fixtures are hand-authored in this package's own `testdata`,
  never a corpus case.**
- **The four arithmetic mutations Task 2 could not state against the
  committed table are stated here** (fold, M-b). Task 2 proved each
  against its own hand-authored fixtures; this task proves the same four
  against the rows that pin them, and reports both halves agreeing:
  a denominator counting a fixed weight table instead of added penalties
  must fail `aggregate/zero-added-key`; a nonzero view including a
  zero-weighted key must fail `aggregate/all-zero-weights`'s empty
  `items`; a sort keyed on `(dist, key)` instead of `(-dist, key)` must
  fail `aggregate/sorted-penalties` and `aggregate/sorted-penalties-tied`;
  and a zero-max path returning `NaN` must fail `aggregate/empty` and
  `aggregate/all-zero-weights`. **Two independent expressions of the same
  arithmetic disagreeing is a finding, not a tolerance.**
- **Wall time is reported broken down**, with `-count=1`, per the standing
  rule. If the tier-0 leg plus the 215-case defaults read outgrows the
  module gate's tolerance, that is a finding for the conductor — it becomes
  its own leg of `scripts/check.sh` at full scope, never an optional target.

**Deliverable count: 6** (the typed decoder dispatching on method with the
add-shape split, the weights-as-override rule, the `get_penalty` stub
model, and the regex-option encoding, with its unknown-method and
unknown-shape failures; the tier-0 replay green over all 72 rows with
exact-equality comparison and a per-method replayed count; the defaults
cross-check over every committed case with its firing mutation; the
`unidecode` manifest pin check and the `ItemFields` corpus cross-check
carried here from Tasks 3 and 1; the two architecture checks with their
fixtures; the package's own doc comment recording why the runner lives on
the harness side). Gate green — **and from this task on, "gate green"
includes tier-0 replay green over all 72 rows**; commit.

**This is the task the review found unexecutable as written**, and its
whole failure was reading the row *ids* rather than the row *methods* and
the derivation shim. `internal/harness/oracle/derive_component.py` is a
**primary input for this task**, read in full before the decoder is
designed; the contracts above are a reading of it, not a substitute for it.

---

## Task 7: The album-level components and `likelies`

Model: `sonnet`. Depends on Tasks 1 through 6.

**Files:**
- Create: `internal/engine/matcher/likelies.go` + `likelies_test.go`
- Create: `internal/engine/matcher/album_artist.go`, `album_album.go`,
  `album_media.go`, `album_mediums.go`, `album_year.go`,
  `album_country.go`, `album_label.go`, `album_catalognum.go`,
  `album_disambig.go`, `album_id.go`, `album_tracks.go`,
  `album_extras.go`, `album_data_source.go`, each with its own `_test.go`

**Behavior contract:**

- **`likelies` is a deliverable, not plumbing.** `distance()` calls
  `get_most_common_tags(items)` itself, so the plurality computation is
  inside 3a's scope and is easy to miss. Port the thirteen fields in beets'
  own order (`artist`, `album`, `albumartist`, `year`, `disctotal`,
  `mb_albumid`, `label`, `barcode`, `catalognum`, `country`, `media`,
  `albumdisambig`, `data_source`), the consensus map beside it, and the
  **albumartist-consensus override** of `likelies["artist"]`. Tie-breaking
  is contract 11's first-seen rule, and it gets its own test with a
  deliberately tied item set. **Falsy items are filtered before counting**
  (`util/__init__.py:838`, contract 10's m6 note), which also means
  `consensus[field]`'s denominator is the filtered count, not the input
  count.
- **Fourteen components in thirteen files, with one named exception**
  (fold, M-i). Contract 8's album order enumerates fourteen components;
  `missing_tracks` and `unmatched_tracks` share `album_extras.go`, because
  they are one counting rule read from both ends of the same pairing and
  splitting them would put half a rule in each file. That is a **deliberate
  departure from "one component, one file"**, stated here and restated in
  `album_extras.go`'s doc comment, not an accident of the file list. Every
  other component is its own file. Both still get their own fire and
  decline tests, and both appear separately in the declared component
  slice, so the order test of Task 8 still sees fourteen names.
- **One component, one file, one test** — with the exception just named —
  in contract 8's fixed order, each with a fire test and a decline test:
  - `artist` (3.0): **skipped entirely** when `album_info.VA`;
  - `album` (3.0): `AddString`, always;
  - `media` (1.0): only when `album_info.Media`; with `Preferred.Media`
    patterns, `AddPriority` against `(\d+x)?(<pat>)` compiled
    case-insensitive and **anchored** (contract 7); else when
    `likelies["media"]`, `AddEquality`;
  - `mediums` (1.0): only when both `likelies["disctotal"]` and
    `album_info.Mediums`; `AddNumber` — each disc of difference is a full
    weighted unit;
  - `year` (1.0): three branches, and **the evaluation date is a
    parameter** (contract 9). With `Preferred.OriginalYear` and
    `album_info.Year`: `original = OriginalYear or 1889`,
    `AddRatio(|Year − original|, |evalYear − original|)`. Else with both
    `likelies["year"]` and `album_info.Year`: `0.0` when the item year
    equals the release or original year; else with an original year,
    `AddRatio(|likelies − Year|, evalYear − OriginalYear)`; else a flat
    `1.0`. **`preferred.original_year` is one of the three
    corpus-unreachable surfaces** (spec, M8), so its branch is graded by a
    hand-authored table test in this package's `testdata`, and the task
    reports the mutation that proved it fires;
  - `country` (0.5): with `Preferred.Countries`, `AddPriority` against
    anchored case-insensitive patterns; else when both present, `AddString`
    on the country codes;
  - `label`, `catalognum`, `albumdisambig` (0.5 each): `AddString` when
    both sides are present;
  - `album_id` (5.0): when `likelies["mb_albumid"]` is truthy,
    `AddEquality` against `album_info.AlbumID`;
  - `tracks` (2.0): per matched pair, compute the track distance with
    `inclArtist = album_info.VA` and add **that track's normalized
    distance** as one penalty under `tracks`, recording the sub-distance in
    `Distance.tracks`;
  - `missing_tracks` (0.9) / `unmatched_tracks` (0.6): one `1.0` penalty
    per candidate track beyond the matched pairs, and one per item beyond
    them;
  - `data_source` (2.0): `AddDataSource(likelies["data_source"], album_info.DataSource)`.
- **The pairs are an input, not a solve.** The album components take the
  item/track pairing as a parameter; producing it is 3d's assignment work.
  A 3a that reaches for a solver is out of scope and stops.
- **`extras` counts, not contents.** Extras never contribute to the
  distance — only their counts do — and their sort orders are 3d's. Say so
  in `album_extras.go`'s doc comment so 3d's implementer finds the boundary
  where the code is.

**Deliverable count: 4** (`likelies` with the consensus override and the
first-seen tie rule proved; the fourteen album components in thirteen
files with one test each, in beets' fixed order, `album_extras.go` named
as the deliberate exception; the year component's three branches
with the evaluation date as a parameter and the unreachable
`preferred.original_year` branch graded by a hand-authored fixture; the
anchored case-insensitive pattern handling for `media` and `country`).
Gate green plus tier-0 replay green; commit.

---

## Task 8: The scoring core

Model: `sonnet`. Depends on Tasks 1 through 7.

**Files:**
- Create: `internal/engine/matcher/score.go` + `score_test.go`

**Behavior contract:**

- **The core holds two slices — album components and track components — in
  beets' own fixed call order, and iterates them.** It treats them
  uniformly: a component is a name, a weight key, and a function. The
  order is load-bearing rather than stylistic (contract 8), and the core's
  doc comment says why in one sentence a later reader can act on.
- **The core adds nothing of its own.** Every penalty comes from a
  component; the core composes, normalizes through the accumulator, and
  returns a `*Distance`. A rule change touches one component file and its
  test — that is the evolvability property the spec asks for, and the
  test that proves it is a component-order test asserting the declared
  slice equals beets' order by name.
- **The album entry point takes `(items, albumInfo, pairs, tunables,
  evalDate)`** and returns the `Distance` with its `tracks` map filled.
  No clock read, no config read, no I/O, no harness import.
- **Whole-distance tests, hand-authored**, that exercise the composition
  end to end on a small synthetic album: one clean pairing (every component
  declining or scoring zero), one with a missing track and an unmatched
  item (proving both counting components and the denominator inflation),
  and one VA album (proving the `artist` skip and the `inclArtist=true`
  path through `tracks`). These are **not** corpus cases and do not read
  `corpus/`; whole-case agreement is tier 1's, at 3d.
- **Report the mutation that proved the order test fires** (two components
  transposed in the declared slice).

**Deliverable count: 3** (the two declared component slices in beets' fixed
order with the order test; the album entry point with the evaluation date
as a parameter and no clock or config read; the three hand-authored
whole-distance composition tests). Gate green plus tier-0 replay green;
commit.

---

## Task 9: The pass close

Model: `sonnet`. Depends on Tasks 0 through 8.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the conductor's
  batch commit)
- Draft: the 3b handoff note and the follow-on task registrations, in the
  HISTORY entry

**Behavior contract:**

- **The HISTORY entry is drafted in the house shape**: what landed, what
  the gate caught, and what a later pass would be wrong to rediscover.
- **The coverage record, stated as a table a reviewer can check.** Per
  component: its file, its weight key, its fire condition, whether a
  tier-0 row grades it, and — for the three unreachable surfaces — which
  hand-authored fixture grades it and what mutation proved that fixture
  fires. **The three unreachable surfaces are recorded as a declared
  corpus-coverage gap, not as a passing check** (spec, M8).
- **The evidence, not a bare "green".** Tier-0 replay: rows replayed per
  method, the total, the exact-equality result, and the (expected empty)
  list of rows that passed only at 1e-12. Defaults cross-check: cases read
  and the result. `bash scripts/check.sh` with broken-down wall time at
  `-count=1`. **No `make derive` ran, and the report says so.**
- **Forward carries, each stated so 3b/3c/3d read them rather than derive
  them:**
  1. **The rung 15 re-run criterion**, quoted verbatim from the Spec
     section above. Every rung 3 pass close carries it.
  2. **The four-knob wall registration** as a named follow-on task for the
     first rung 3 pass that runs after rung 1b closes, with the four knobs
     named and `internal/config/keys.go` named as the file no lane edits.
  3. **The `SD_PATTERNS`-as-a-value seam**, naming `BeetsSDPatterns`,
     `SDPatternsWordBoundary`, and `Tunables.StringDistPatterns`, so 3b's
     `defect_free` derivation and 3d's compat switch swap a table rather
     than edit a component. **With it, the two measured figures and the
     warning that produced them**: the word-boundary guard changes **zero
     of the 72 tier-0 rows**, so tier 0 is blind to a "fixed" table and
     Task 4's byte-identity test is the only thing standing between the
     port and a silent improvement. The discriminating pair,
     `string_dist("Draft: Ember Wharf", "Draft: Northern Lines")`, reads
     **`0.061111111111111116`** under `BeetsSDPatterns` and
     **`0.6111111111111112`** under `SDPatternsWordBoundary`; 3b checks its
     own `defect_free` derivation against that pair before trusting it
     (fold, M-a).
  4. **The M2 resolution and the two drafted `corpus/README.md` cells**,
     for the conductor to land; and the note that the mechanics study's Q4
     "19 keys" is superseded by the tree and by
     `config_default.yaml:174-193` (the twenty entries; `:173` is the
     header — fold, m4).
  5. **The four corrected citations**, so no later pass re-derives them:
     `DataSourceMismatchPenalty`'s default is `metadata_plugins.py:180`
     wired at `:202`, not `:190-195`; the `SearchLimit` legacy spelling is
     `musicbrainz.searchlimit`, not a top-level `match:` key; `ExtraTags`
     lives on the musicbrainz plugin and has no `config_default.yaml`
     citation; and `match:` carries `album_disambig_fields` /
     `singleton_disambig_fields`, display-only and deliberately outside
     `Tunables`.
  6. **The decoder contract as built** — dispatch on method, the
     add-shape split, weights-as-override, the `get_penalty` stub model,
     and the regex-option encoding — because 3c and 3d extend
     `internal/harness/conform` and would otherwise re-derive it from
     `derive_component.py` a second and third time.
  7. **The `inclArtist` divergence** between the assignment objective and
     the final per-track distances, for 3d's tie-mirroring task.
  8. **The `unidecode` pin check**, split across `matcher`'s exported
     version stamp and conform's manifest comparison, so a manifest bump
     reds the Go side too.
  9. **Any row that passed at 1e-12 without bit-identity**, with the
     component and suspected operation order named.
  10. **The measured tier-0 and gate wall time**, which the version-bump
     differential doc will want, the same courtesy 2b passed to 2c and 2c
     passed to 2d.
- **STATUS stays present tense and under 60 lines**; anything historical
  moves to HISTORY. The next action points at 3b or 3c (mutually
  independent, may swap order), with 3d needing all three.
- **Both budgets are scored**: token spend against the re-derived ceiling,
  and attended interactions against zero, which is this pass's target.

**Deliverable count: 4** (the HISTORY draft in the house shape with its
ten forward carries; the component coverage table with the declared
corpus-coverage gap; the evidence record with broken-down wall time and the
bit-identity findings list; the STATUS rewrite with the next action). Gate
green plus tier-0 replay green; commit.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain per task,
**strictly sequential** — nothing here runs in parallel, and if a second
code slot opens at all it belongs to lane T in its own worktree, never to a
second executor in this tree. **Lane T is blocked until Phase 0a, 0b, and
0c land** (fold, M-h), so the expected state through this pass is one code
lane. Every task runs implementer → `diff-reviewer` → gate. The
conductor dispatches `code-simplifier` over each accepted task's changed Go
and Python after review and before the pass close, and owns every ledger
edit, the lane registry, the `corpus/README.md` correction, and the batch
commit. **Runaway guard armed at launch**, per the standing requirement.
Subagents never dispatch agents.

**Models.** `sonnet` by default. **Two named upshifts to `opus`:** Task 2
(the accumulator) and Task 4 (`string_dist`), both novel
correctness-critical arithmetic where a plausible implementation is
silently wrong and the cost of finding it at 3d is a whole pass. Task 6 is
the runner-up: the conductor upshifts it if Task 2's or Task 4's report
says an arithmetic contract above was under-specified. Task 3 carries a
**named escalation trigger in place of an upshift** — two failed table
attempts stop the task — matching 2c's Tasks 17 and 18 and 2d's Task 1.

**Review.** Rung 3 is not one of the three full adversarial-cycle rungs (5,
12, 16), so review scales to the in-chain `diff-reviewer` per task plus the
pass-end gate. **This pass does not take a fresh-context review of its own
instrument.** The spec raises 3d's conformance runner as the one candidate
exception, and 3a's tier 0 is materially cheaper to trust: it grades 72
committed rows whose expected outputs were derived by executing pinned
beets, its skipped-row and unknown-method failures make silence impossible,
and every checker ships a firing fixture. The conductor may still take one;
this plan states that it is not required and why, so the decision is
recorded rather than defaulted.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer dispatch.**
  Its blockers are ruled by the conductor before Task 1.
- **Task 1 → everything.** It births the package and defines the vocabulary
  and the tunables every later task reads.
- **Task 2 and Task 3 are independent of each other**, both depending only
  on Task 1. They still run sequentially (one tree, one gate), Task 2 first,
  because Task 2's spend is what the ceiling is re-derived from.
- **Tasks 2, 3 → Task 4.** `string_dist` needs the accumulator's `AddString`
  seam and the transliteration table.
- **Tasks 2, 4 → Task 5.** Track components need the adders and `string_dist`.
- **Tasks 1–5 → Task 6.** Tier 0's `track_distance` rows need the track
  components; its adder, aggregate, and `string_dist` rows need Tasks 2
  through 4.
- **Task 6 → Task 7 → Task 8 → Task 9.**

**Where this pass splits, if it splits.** Nine tasks plus a pre-task is
past the roughly-eight bar, so the split point is proposed rather than the
scope packed. **The split is at the Task 6 / Task 7 boundary**, and it is
unusually clean: `components.json` carries no album-distance rows at all
(its eleven methods are the eight adders, `generic_penalty_keys`,
`string_dist`, and `track_distance`), so **tier 0 grades all 72 rows
without the album components or the scoring core**. A 3a-i of Task 0
through Task 6 plus a short close is therefore a complete, gate-green unit
that ships graded primitives; a 3a-ii of its own staleness pre-task plus
Tasks 7, 8, and 9 ships the album side and the core on top of them.

**The recommendation is to keep one pass**, and to take the split only if
Task 2's actual spend re-derives the ceiling past 6M: the album components
are mechanical once the accumulator and `string_dist` are green, splitting
costs a second staleness pre-task and a second close, and 3b's inbound is
"3a's info types", which Task 1 already satisfies either way. The
conductor takes the call at the Task 2 checkpoint, on the number rather
than on a feeling.

**Budget:** 6M starting ceiling, re-derived after Task 2 and again after
Task 4; phase 1's combined ceiling 12M as stated; flag at 4.8M; STATUS
after every task. Spend is scored at close against the re-derived ceiling
and against **zero attended interactions**, which is this pass's target.

---
