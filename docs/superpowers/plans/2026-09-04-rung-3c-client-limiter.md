# Rung 3, Pass 3c (The WS/2 Client, the Query Builder, the Shared Limiter, and the Extra Channel) Implementation Plan

**Revision 2, 2026-09-04:** folds the two adversarial plan reviews,
merged in `research/2026-09-04-rung-3c-plan-review.md`: blockers B1 to
B3, majors M-a through M-h, minors m1 to m26. Four reviewer citation
claims are rejected with the evidence in that document's header, and five
items are conductor-owed rather than folded silently; each of those five
was applied provisionally, marked inline, and ruled as recommended on 2026-09-04 (revision 2.1).

The load-bearing resolutions. **The dispatcher's reserved token no longer
goes to the lowest-priority non-empty class** (B1): as written it starved
`ClassBatch` outright whenever a probe stream and the shadow were both
queued, which is the exact starvation the build spec constraint the
plan's own doc comment quotes forbids. It now round-robins across the
non-empty classes below the class strict priority would have served, and
Task 5 gains a three-class row and a matching mutation. **Both
hand-driven concurrency tests are re-specified around a registration
handshake** (B2, B3): a signal sent before `Wait` proves nothing about
enqueue, and a tick fed ahead of demand is dropped against a spare capped
at one, so the limiter grows an unexported `awaitQueued` seam and Task
6's retry test drives one tick per observed attempt. **The wire pace is
stated for the first time** (M-b): beets' `musicbrainz.org` branch pins
`rate_limit = 1.0` and the session limiter binds before the 0.25s
adapter, so C12 is reworded, C20 is added, and `NewLimiter` takes an
interval rather than an integer rate. **`Retry-After`'s HTTP-date form
becomes R13's third named divergence** (M-c), because honoring it needs a
clock read `TestEngineNeverReadsTheClock` forbids inside
`internal/engine`. **`Budget`'s cap reserves before waiting** (M-d),
because a pre-wait check is a lost update under the concurrency the type
exists for. **The search response's wire shape becomes contract C19**
(M-e), because it is an object keyed `releases` and no task said so.
Task 3 gains the two doc comments its change falsifies, `penalty.go:334-337`
and `tunables.go:138-140` (M-a). Task 2 states the derivation rule that
makes the second replay map inert for `make derive` (M-g). The
escaped-hyphen mutation's miss count is 6 rather than 5 in both tasks
that state it (M-f), and Task 4's map-order mutation becomes a
deterministic sorted-key mutant (M-h).

Four folds take an adaptation rather than the literal fix, each named
where it lands: **m2** turns the rung 3 spec's `match.searchlimit` item
into an optional clarification rather than a correction, because that
line spells legacy keys relative to the plugin namespace and its sibling
`source_weight` item does the same; **m19** unexports `VALikely` and
corrects R1's seam sentence instead of adding a consumer, on R6's own
dead-surface ground; **m8** settles the double-validate by giving
`TranslateAlbum` sole ownership rather than adding a call anywhere; and
**m21** states Task 6's deliverable count as six plainly rather than
moving `release.go`'s `wireInt` wrap into a task whose wall excludes
`internal/engine/musicbrainz`.

**Conductor rulings, 2026-09-04, all five taken as the fold recommended** (each was applied provisionally and is marked inline; see revision 2.1):
the dispatcher's reserved-token policy (B1); `Tunables.DataSourceMismatchPenalty`'s
ending, wired into `DefaultTunables` or documented as a readerless
config-load seam (M-a); the shipped wire pace and `NewLimiter`'s
signature (M-b); `corpus/README.md` as an eighth contested-surface row
(m22); and where the 80%-of-ceiling flag routes during the standing
overnight grant (m23).

**The merged review's 41-item "Verified sound" list is a second
do-not-re-litigate record** for this pass: all eighteen client and query
contracts as originally stated, the whole request census measured twice
off the tree, the empty-migration round-trip proof over all 215 cases,
the canonical-encoding probe, the two independent reproductions of all 48
committed search queries, the decode-validate-translate sweep over all
303 committed lookups, the harness API signatures, both architectural
invariants, R16's premises, and the ceiling arithmetic are cited below
rather than re-derived. An implementer who believes one of them is wrong
reports a finding and implements the record anyway.

**Revision 1, 2026-09-04.** Drafted against the rung 3 spec at **revision
2** (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`), which is
ratified on its six rulings and review-folded on all nineteen findings of
`research/2026-09-04-rung-3-spec-review.md`. That review's "Verified
sound" list is the do-not-re-litigate record: its twenty items are cited
here rather than re-derived, and an implementer who believes one of them
is wrong reports a finding and implements the record anyway. The same
standing applies to `docs/HISTORY.md`'s 3a and 3b entries, whose ten and
seventeen rediscover-guards are cited below rather than re-measured.

This plan's primary input is
`docs/superpowers/research/2026-09-04-rung-3c-planning-inventory.md`, a
research report that read pinned beets 2.13.1 in `tools/venv/.venv`, the
harness and corpus as landed at `d518e57`, and the engine as 3a and 3b
left it. Every mechanic below is cited to the tree, and the inventory's
fourteen deduplicated open questions are **resolved in this plan**, each
with its evidence and its rejected alternative, in "Resolutions this plan
takes rather than discovers".

**Two inventory figures are corrected here, off the tree, before any task
inherits them.** The inventory's §(c).2.1 says "214 of 215 committed
cases carry a `comp` key on their items"; measured at HEAD `d518e57` with
`jq 'if ([.input.items[] | has("comp")] | any) then "y" else "n" end'`
over all 215 cases, the figure is **15**, which is what its own §(a).5
already said. Fifteen cases also carry `comp: true`. §(a).5 wins and
§(c).2.1 is the losing sentence. Second, `FIELDS_TO_MB_KEYS` is **eight**
entries, not six plus two specials
(`tools/venv/.venv/lib/python3.13/site-packages/beetsplug/musicbrainz.py:47-56`):
`barcode`, `catalognum`, `country`, `label`, `media`, `year`, `tracks`,
`alias`.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section: pass-execute chain per task, sequential. The conductor consumes
> reports and never reads diffs inline. Subagents never dispatch other
> agents; `code-simplifier` is the conductor's dispatch after a task's
> review accepts. The conductor also owns the ledgers (`docs/STATUS.md`,
> `docs/HISTORY.md`, `ROADMAP.md`), the lane registry, and the batch
> commit of anything a task drafts into them, including this pass's four
> `corpus/README.md` corrections and its two `search_limit` spelling
> corrections, which an implementer drafts and never writes. Rung 3 is not
> an adversarial-cycle rung (those are 5, 12, and 16), so review scales to
> the in-chain `diff-reviewer` per task plus the pass-end gate. **This
> pass runs in the MAIN TREE**, and nothing here may be executed from a
> worktree: the spec's tree column for 3c is "Main tree: writes
> `internal/harness/conform` and `internal/harness/stub`"
> (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md:330`), and the
> standing policy is that corpus and harness work never moves to a
> worktree
> (`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`,
> lane-launch checklist item 5).

**Revision 2.1, 2026-09-04 (conductor rulings on the fold, all five taken
as recommended):** B1, the reserved token round-robins across the non-empty
classes below the served one; M-a, `DataSourceMismatchPenalty` is wired
into `DefaultTunables`' penalty map; M-b, `NewLimiter(interval
time.Duration)` with `DefaultInterval = time.Second`; m22, `corpus/README.md`
and `corpus/manifest.json` both gain contested-surface rows, landed in the
parallel spec before launch; m23, the 80%-of-ceiling flag routes to the
conductor **for the duration of the 2026-09-04 overnight grant only**, and
CLAUDE.md's rule that it routes to Geoff stands otherwise. Every
"provisional, conductor ruling owed" mark below now reads as ruled.

## Primary inputs

Read these before Task 1, in this order:

1. **`docs/superpowers/research/2026-09-04-rung-3c-planning-inventory.md`,
   in full.** It is this plan's argument. Section (a) is the beets client
   and query mechanics with line citations against `tools/venv/.venv`;
   (b) is the harness, stub, cassette, and request logs as landed; (c) is
   the engine as landed and the proposed Go surface; (d) reconciles all
   three and lists the fourteen open questions. Do not re-derive what it
   settles; cite it and verify the line if you carry one forward. Its two
   corrected figures are named in this plan's header.
2. **`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`, revision
   2**, in full. The pass-split table's 3c row (`:330`) is this plan's
   scope statement. The conductor rulings in its header (`:45-50`), "The
   rate limiter" (`:238-242`), "`Case.Extra` and the replay channel it
   does not have yet" (`:261-263`), "The three tiers" (`:271-279`), and
   "Assertion discipline" (`:316-321`) are the sections this pass
   implements.
3. **`docs/HISTORY.md`'s "Rung 3, pass 3b" entry, in full**, then its
   "Rung 3, pass 3a" entry. 3b's guards 5 (the two-shape `position`
   decode), 8 (the `MergePseudoRelease` shallow copy), 9 (the
   `Preferred.Countries` two-regime hazard), 11 (the `\b` regime
   divergence), 14 (the browse fallback and its guard) and 16 (the
   wire-type census) are load-bearing here, as is 3a's guard 2 (the
   deferred four-knob registration) and its case-mapping sweep.
4. **`docs/superpowers/plans/2026-09-04-rung-3b-translation.md`**, for its
   sixteen translation contracts and its Global Constraints. This plan
   does not restate them; it adds the client and query contracts below.
5. The tree, read-only: `internal/engine/musicbrainz/{doc,release,
   translate,pseudo,relations}.go`; `internal/engine/matcher/{info,
   likelies,track_artist,tunables,penalty}.go`;
   `internal/harness/cassette/cassette.go` and `cassette_test.go`;
   `internal/harness/stub/stub.go` and `stub_test.go`;
   `internal/harness/conform/{doc,translate,component,defaults}.go` and
   `arch_test.go`; `internal/harness/corpus_test.go`;
   `internal/harness/mbrequest/mbrequest.go`; `cmd/harness/capture.go`;
   `internal/config/{config,keys,thresholds}.go`.
6. `CLAUDE.md`'s Build process and Gate sections, `docs/STATUS.md`'s lane
   registry, and `corpus/README.md`'s Layout paragraph (`:40-46`), its
   schema-version history (`:1959-1980`), and its recorded-fixed-point
   section (`:2775-2785`).
7. **The pinned venv itself**,
   `tools/venv/.venv/lib/python3.13/site-packages/`, abbreviated `$SP`
   below, for `beetsplug/musicbrainz.py`,
   `beetsplug/_utils/musicbrainz.py`, `beetsplug/_utils/requests.py`,
   `beets/metadata_plugins.py`, `beets/autotag/match.py`,
   `beets/autotag/distance.py`, `beets/plugins.py`,
   `beets/util/id_extractors.py`, and `urllib3/util/retry.py`. Every
   contract below cites it. Short offline probes with
   `tools/venv/.venv/bin/python` are permitted and expected. **`make
   derive` does not run in this pass at all** (Resolution R5).

## Goal

Ship the request half of the Go matcher: the code that decides what to
ask MusicBrainz for, the code that asks, the code that paces the asking,
and the harness instrument that proves the bytes are the bytes pinned
beets sent.

Land `internal/engine/musicbrainz/query.go` (the sixteen-element
`RELEASE_INCLUDES` set, `va_likely`, `get_album_criteria`'s ordered
criteria, and the Lucene rendering layer),
`internal/engine/musicbrainz/client.go` (the WS/2 client with `fmt=json`
forced, beets' retry policy, an injected base URL, a required
User-Agent, and the pseudo-release double fetch), and
`internal/engine/musicbrainz/limiter.go` (the three-class shared limiter
and the per-caller budget seam the spec's deliverable 5 names). Land the
`Case.Extra` replay channel in both halves the spec's B2 fold names: the
cassette schema half (`Extra` carrying `Request` shape) and the stub half
(`Consumed()` distinct from `Misses()`). Land the request-construction
grading in `internal/harness/conform` that replays all 351 committed
requests through the Go client against `stub.Server` at zero misses.

Nothing here decides a verdict, computes a distance, or reads a tag. The
scoring core is 3a's and is shipped; the translation layer is 3b's and is
shipped. The assignment solve, admission filters, the ID-match early
return and the search-fallback flow, recommendation, `max_rec`,
verdict-to-action, the compat switches, and tiers 1 and 2 are 3d's.

## Owner touchpoints: NONE

3c needs no ruling, no adjudication, no capture contact, and no network.

The six rulings this rung rests on were ratified on 2026-09-03, and the
conductor already ruled the one question this pass would otherwise have
asked: the cassette schema half is accepted and lands as 3c's first task
(spec `:45-50`). Adjudication batch 1 is fully decided (21 rows at 18
`oracle-wrong` / 3 `oracle-correct`, zero pending; six hypotheses at four
`oracle-wrong` / two `oracle-correct`, zero pending), and this pass
produces no adjudication row and writes no `disposition`. It registers
nothing on the config wall, because rung 1b has not landed
(`internal/config/keys.go:19-25` still carries `ownedWall` as a slice
literal, verified at HEAD) and the four-knob registration stays the named
follow-on task (Resolution R16).

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition named below resolves to the conductor. Only a
conductor decision to escalate reaches Geoff. The standing 80%-of-ceiling
flag CLAUDE.md requires still applies and is contingent, not a touchpoint
this plan expects to spend.

Two conductor edits are owed before Task 1 commits, and both are table
rows rather than rulings: `corpus/manifest.json` and `corpus/README.md`
have no contested-surface row (see "Contested surfaces" below). Task 0
reports them; the conductor writes them; neither is a launch blocker.
Five further items are conductor-owed and each is applied provisionally
with an inline marker, listed in the revision note above.

## Spec

`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md` (revision 2),
and behind it `specs/2026-09-02-greenfield-server-build.md` (revision
2.1, ratified), ladder entry 3. This pass ships deliverable 5, quoted in
full: "the shared MusicBrainz rate limiter, shipping the per-caller
budget interface. Its priority policy is re-verified at rung 9 when its
real callers exist" (spec `:63`), together with the client and query
construction the spec's tree section names
(`query.go  Lucene search construction, the inc= sets`; `client.go  the
WS/2 client: fmt=json, retries, timeouts, User-Agent`; `limiter.go  the
shared limiter and the per-caller budget interface`, spec `:180-182`),
and the `Case.Extra` replay channel's two halves (spec `:261-263`).
Deliverables 1, 2, 3 and 4 are 3a (shipped), 3b (shipped), 3d, and 3a
plus this pass's `SearchLimit` correction.

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
(`:360`), and **re-derived from the first engine task's actual report**,
meaning Task 4, `query.go`, before Task 5 dispatches.

The basis. 3b closed at **≈4.82M against 6M with zero attended
interactions** (`docs/HISTORY.md`, 3b Budgets), of which **2,106K was the
plan-prep workflow** and **2.7M was execution** across nine tasks plus a
pre-task, so ≈270K per executed task. 3a closed at ≈2.84M across nine
tasks with its own prep uncounted. 3c is eight tasks plus a pre-task: two
mechanical harness tasks over an empty migration (Tasks 1 and 2), one
small matcher task (Task 3), three engine tasks of which two are
upshifted (Tasks 4, 5, 6), one harness-grading task (Task 7), and a close
(Task 8). At 3b's measured execution rate that is ≈2.4M, plus roughly one
task's worth again for the two `opus` tasks, so ≈3.0M for execution.
3c's own plan-prep workflow is bounded by 3b's 2.1M. Expected total
**≈5.1M against 6M**, which is why 6M is held rather than lowered: the
ceiling is the stop rule's yardstick, not a forecast, and two of this
pass's tasks (the limiter's concurrency and the client's retry seam)
carry more unknown than any 3b task did.

**Flag at 80% (4.8M):** finish the task in flight, write STATUS, and one
combined question goes **to the conductor** before continuing.
*(conductor ruling 2026-09-04: taken, fold m23)* CLAUDE.md's literal
rule routes that question to Geoff. `docs/STATUS.md:29-31` records the
standing grant to "run unattended through 3c execution, 3d prep, and 3d
execution as the clock allows", and CLAUDE.md's own grants say to
continue unblocked work when Geoff is unavailable, so a question that
stops an unattended run for an answer nobody is awake to give is a defect
rather than a gate. The flag therefore routes to the conductor, STATUS is
the record, and Geoff's reading of STATUS is the touchpoint. The
conductor rules whether that holds for the duration of the overnight
grant only, or generally.

**Checkpoint: STATUS after every task.** That is 2b's, 2c's, 2d's, 3a's
and 3b's cadence; every task reports its own spend and its own gate
evidence, and the conductor sizes the next task against it.

**Phase 1's combined ceiling: 12M, and it stays a reservation rather than
a live sum.** It is 3c's 6M plus a **6M reservation for the second code
slot** (lane A, rung 5, or lane T, rung 8, one at a time).
`docs/STATUS.md`'s lane registry records lane-2 as BLOCKED until Phase 0b
and 0c land, so **no second sub-ceiling exists to sum today**, and 12M is
not a figure any spend is currently measured against. The conductor
replaces the reservation with that lane's own assigned sub-ceiling at its
launch; until then 3c's 6M is the only live ceiling. This restates 3a's
and 3b's figure unchanged rather than re-deriving it.

## Global Constraints

- **Main tree.** The spec's tree column for 3c is "Main tree: writes
  `internal/harness/conform` and `internal/harness/stub`" (spec `:330`),
  and the standing policy is that corpus and harness work never moves to
  a worktree (parallel-approach spec, lane-launch checklist item 5). The
  lane registry entry is written before the first dispatch, and **the
  live-executor check is that registry, never `pgrep -f
  Projects/dubplate`**, which matches every lane once a worktree exists.
- **Package territory, stated as a wall.** This pass writes
  **`internal/harness/cassette`** (the `Case.Extra` shape and its
  validation), **`internal/harness/stub`** (the `Extra` replay map and
  `Consumed()`), **`internal/harness/conform`** (the request-construction
  grading, plus the two now-dead data-source compensations Task 3
  removes), **`internal/harness/corpus_test.go`** (the schema constant,
  the consumed-file test, and the round-trip proof),
  **`corpus/manifest.json`** (one field, `schema_version`),
  **`internal/engine/musicbrainz`** (`query.go`, `client.go`,
  `limiter.go`, and one line in `release.go`), and
  **`internal/engine/matcher`** (three named additions and two doc-comment
  corrections: `ItemFields.Comp`, `IsVariousArtistsName`, the
  `dataSourcePenalty` fallback constant, `Tunables.SearchLimit`'s legacy
  spelling, `Tunables.DataSourceMismatchPenalty`'s role), plus
  conductor-owned `docs/` ledgers. It writes **nothing** under
  `corpus/cases/`, `corpus/analysis/`, `corpus/profiles/`,
  `internal/harness/{analysis,oracle,synth,mbrequest}`, `internal/config`,
  `cmd/`, `tools/`, `web/`, or `_salvage/`. A task that believes it needs
  one of those stops and reports.
- **The spec's tree column names two harness packages; this pass writes
  four paths under `internal/harness` plus two engine packages.** The
  extension is deliberate and each part has a reason stated at its task:
  `cassette` is the conductor-ruled first task (spec `:45-50`),
  `corpus_test.go` carries the single `wantManifestSchemaVersion`
  constant the schema change must move (spec `:373`), the engine files are
  the spec's own tree sketch (`:180-182`), and the three `matcher`
  additions are 3a's own carry addressed to 3c plus the two inbound gaps
  the query cannot be built without (Resolution R8). This mirrors 3b's
  own m14 correction; Task 8 drafts the spec's tree-column fix for the
  conductor.
- **`corpus/cases/*` is frozen input, and read-only is a real
  permission.** Loading every committed `case.json` and replaying every
  `responses/*.json` through `stub.Server` is exactly what the
  request-construction grading is for. **No agent reads a
  `responses/*.json` body into context.** Every figure about response
  content in this plan was produced mechanically.
- **`internal/config` is not edited, and neither is `keys.go`.** Ruling
  C's registration half is hard-inbound on rung 1b (Phase 0b, unlanded;
  `internal/config/keys.go:19-25` still carries `ownedWall` as a slice
  literal, verified at HEAD `d518e57` and re-verified at `f2bc19e`). The
  four-knob registration stays
  the named follow-on task for the first rung 3 pass that runs after 1b
  closes (3a HISTORY guard 2). The `.gitleaksignore` rider applies even
  though this pass does not touch the file: its single allowed finding is
  pinned by line number at
  `internal/config/keys_test.go:generic-api-key:66`, and an insertion
  above line 66 reds the gate for every lane at once. An append is
  conductor-owned, must sit strictly below line 66, and is never a rule
  disable.
- **The engine never imports the harness, and never reads the clock.**
  `TestEngineNeverImportsHarness`
  (`internal/harness/conform/arch_test.go:141-146`) and
  `TestEngineNeverReadsTheClock` (`:152-157`) both root at
  `internal/engine`, so **`internal/engine/musicbrainz/limiter.go` is
  inside both walks**. The limiter takes no clock read at all (Resolution
  R2); neither check is amended by this pass, and an implementer that
  finds itself proposing an exemption stops and reports.
- **`internal/harness/analysis` is not written by this pass.** Its
  `Case.Extra` blind spot is real (it reads `Requests` only,
  `internal/harness/analysis/album_info.go:248-264`) and it needs no
  change, because the `Extra` reshape adds a field to a type `analysis`
  never touches.
- **Assertion discipline, at the spec's own bar.** "Every checker in the
  runner ships a fixture that makes it fire, and the task that lands it
  reports the mutation that proved it can fire" (spec `:318`). The rule
  reaches this revision's additions by name: "the `Extra` replay
  channel ... ship[s] a fixture proving the failure, and each is
  hand-authored in its package's own `testdata` rather than borrowed from
  the corpus" (spec `:320`). No silent-success fakes, no vacuous
  assertions, no self-derived expectations. **Hand-authored fixtures live
  in the owning package's own `testdata`, never a corpus case**, on the
  proximity precedent.
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects and nothing else, verified at HEAD. No
  `golang.org/x/time/rate`, no `golang.org/x/text/cases`, no MusicBrainz
  client package. A task that wants one stops and reports (Resolutions R2
  and R14 each reject one by name).
- **Universality.** No estate hostname, path, email, or catalog value
  anywhere outside `internal/config`. The client's `UserAgent` is a
  required caller-supplied string the engine never builds (Resolution
  R11's second half); hand-authored fixtures are invented strings, never
  estate data.
- **No live network. Anywhere.** The gate never reaches the network and
  neither does anything else here. No `curl`, no `WebFetch`, no live
  `/ws/2` request. Every client test and every conformance run points at
  `stub.Server` on an ephemeral loopback address or at an injected
  `http.Client` transport.
- **`make derive` does not run in this pass.** No task touches oracle
  territory, no Python file changes, and the `Case.Extra` reshape's
  byte-identity proof is offline (Resolution R5). The standing invariant
  is unaffected and is restated so it does not fall off: after this pass
  `make derive` must still re-derive `corpus/analysis/components.json`,
  `ties.json`, `file_tags.json`, `album_info.json`, and every
  `corpus/cases/*/case.json` byte-identically.
- **Determinism.** No map-iteration order reaches an output. The
  extra-tag walk iterates `Tunables.ExtraTags` and looks up the MB key,
  never the reverse (contract C8), and skips a tag it has already emitted
  so a repeated tag yields one criterion as beets' own dict does. The
  limiter's dispatcher is deterministic under a hand-driven tick channel
  **only when every tick is delivered after its waiter has registered**
  (folds B2 and B3): a signal a waiter sends before calling `Wait` proves
  nothing about enqueue, and a tick fed into an idle dispatcher is held
  as the single spare and any further one is dropped. Tasks 5 and 6
  observe registration through the limiter's own `awaitQueued` seam.
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck,
  ruff over every tracked `*.py` (this pass adds none), vale over Go
  comment prose, gitleaks over the tracked worktree, `docker compose
  config`, `systemd-analyze verify`. **Never commit red. Specific files
  only.** Commit footer `Co-Authored-By: Claude
  <noreply@anthropic.com>`. **Tier-0 replay green over all 72 rows is
  part of every task's gate expectation**, because it runs inside `go
  test ./...`, and Task 3 makes it a named acceptance criterion rather
  than a background fact.
- **`go-conventions` before any Go.** The bar is the standing
  exemplary-stdlib-idiom bar (`idiomatic-beautiful-go`). Concrete types
  over interfaces, no functional options, `slices.SortStableFunc` with
  `cmp.Compare`, `for range n` where the index is unread, `errors.Join`
  for multi-error checkers, `any` not `interface{}`, no `init()`, no
  `context.Context` in a struct, errors lowercase and context-first with
  `%w` only where a caller branches. Three places need a written reason in
  the doc comment because they look like anti-patterns and are not: the
  limiter's goroutine (Resolution R2), `Budget`'s injected `wait` field
  (Resolution R2), and the engine's second, deliberately unshared copy of
  the `/ws/2` constants (Resolution R9).
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each before Task
  1 dispatches.

## Contested surfaces, and the two rows the conductor owes

Lane-launch checklist item 2 requires a written owner for every surface a
lane touches (parallel-approach spec, "Lane-launch checklist"). 3c
touches eight. **Six already carry a row in the parallel spec's
contested-surface table; two do not.**

| Surface this pass writes | Row today | Status |
| --- | --- | --- |
| `internal/harness/{cassette,stub,analysis}` | "lane E1 as shipped; rung 3 pass 3c makes the ruled `Case.Extra` schema-and-stub change, and rung 3 pass 3b adds the `AlbumInfo` ledger and the `defect_free` block to `analysis`" | written, and names **this pass's change explicitly** |
| `internal/harness/conform` | "born by rung 3 pass 3a; rung 3's lane thereafter" | written |
| `internal/harness/corpus_test.go` | "rung 3's lane while rung 3 runs; lane E1 otherwise" (conductor ruling, 2026-09-04) | written |
| `internal/engine/{matcher,musicbrainz}` | "rung 3 (the E2 lane): `matcher` born by 3a, `musicbrainz` born by 3b; no other lane writes either" (conductor ruling, 2026-09-04) | written |
| `corpus/analysis/*` and `corpus/cases/*` | "lane E1 authors ...; rung 3 adds schema only through re-derivation" | written, and **this pass writes neither** |
| `docs/` ledgers | conductor-owned by standing rule | written |
| **`corpus/manifest.json`** | **none.** The `corpus/analysis/*` and `corpus/cases/*` row does not cover the manifest, and the manifest is where the schema bump's second half lands | **conductor-owed**, one row |
| **`corpus/README.md`** | **none.** The `corpus/analysis/*` and `corpus/cases/*` row does not cover the README, and the `docs/` ledgers row names only `docs/` (fold m22) | **conductor-owed**, one row |

**The two edits this plan asks the conductor to make**, before Task 1's
commit *(conductor ruling 2026-09-04: taken) on the second, fold m22: the
alternative is a ruling that the `docs/` ledgers row already covers every
narrative document wherever it sits)*:

| Surface | Owner | Consumers |
| --- | --- | --- |
| `corpus/manifest.json` | lane E1 as the corpus's own pin file; rung 3 moves `schema_version` alone, in lockstep with `internal/harness/corpus_test.go`'s `wantManifestSchemaVersion`, and touches no other field | `cassette.LoadManifest`, `oracle.Reconcile`, `conform.CheckUnidecodeVersion` |
| `corpus/README.md` | conductor-owned, on the same standing rule the `docs/` ledgers row states; a pass drafts a correction and never writes the file | every later pass reading the corpus's own narrative record |

**It is not a launch blocker.** The `{cassette,stub,analysis}` row
already awards 3c "the ruled `Case.Extra` schema-and-stub change" by
name, and the manifest's `schema_version` is that change's mechanical
half rather than a separate claim on E1's territory: the two values are
compared against each other by one test
(`internal/harness/corpus_test.go:162-174`) and cannot move apart. The
checklist's rule blocks the *second* lane, and lane-2 is BLOCKED. Task 0
reports it; the conductor writes it.

## The client and query contracts, stated once

Every task below cites this section instead of restating it. Sources:
the pinned venv at `tools/venv/.venv/lib/python3.13/site-packages`,
abbreviated `$SP`, read at drafting time and spot-verified line by line.
Verify any line citation you carry into code.

**C1. Every request forces `fmt=json`, with no exception.**
`MusicBrainzAPI.request` does `kwargs.setdefault("params", {})` then
`kwargs["params"]["fmt"] = "json"` before delegating
(`$SP/beetsplug/_utils/musicbrainz.py:615-619`). All 351 committed
requests carry it.

**C2. The API root is `<host>/ws/2`, and the `musicbrainz.org` branch
ignores three config keys.** `api_root` is `f"{self.api_host}/ws/2"`
(`$SP/beetsplug/_utils/musicbrainz.py:609-610`). When `host ==
"musicbrainz.org"`, `api_host` and `rate_limit` are hard-coded to
`"https://musicbrainz.org"` and `1.0`, and `https`, `ratelimit` and
`ratelimit_interval` are ignored entirely (`:598-599`); any other host
honors all three (`:600-606`). The corpus's own derivation seam uses the
second branch, with `host: <stub addr>`, `https: no`, `ratelimit: 100`
appended by the Go adapter (`internal/harness/oracle/oracle.go:247`).

**C3. A release lookup carries the sixteen-element `RELEASE_INCLUDES`,
joined with a literal `+`.** `get_release` defaults
`includes=RELEASE_INCLUDES` (`$SP/beetsplug/_utils/musicbrainz.py:687-690`),
the list at `:50-67`, and `_get_resource` sets `kwargs["inc"] =
"+".join(includes)` before handing the params to the session
(`:625-637`). On the wire the `+` percent-escapes to `%2B` because
`requests` encodes params with `quote_plus`. Measured off the tree:
**exactly one distinct `inc=` value across all 303 committed lookups**,
byte-identical to `internal/harness/mbrequest/mbrequest.go:22-24`.

**C4. A search carries no `inc` at all.** `search` passes only `query`
and `limit` into `_get_resource`
(`$SP/beetsplug/_utils/musicbrainz.py:665-685`), which adds `inc` only
when `includes` is truthy (`:634-635`). Measured: `inc` appears on 303
of 351 committed requests and on none of the 48 searches.

**C5. The Lucene escape set is exactly nineteen characters.**
`LUCENE_SPECIAL_CHAR_PAT = re.compile(r'([-+&|!(){}[\]^"~*?:\\/])')`
(`$SP/beetsplug/_utils/musicbrainz.py:48`): `-`, `+`, `&`, `|`, `!`,
`(`, `)`, `{`, `}`, `[`, `]`, `^`, `"`, `~`, `*`, `?`, `:`, `\`, `/`.
Notably absent, and therefore never escaped: `'`, `,`, `.`, `;`, `<`,
`>`, `=`, `#`, `%`, `@`, `$`. The committed corpus proves both halves:
`%27` (2 apostrophes) and `%2C` (1 comma) survive unescaped, while
`%5C` appears 24 times for escaped hyphens, quotes and one colon.

**C6. Rendering is lower, strip, drop-if-empty, escape, then wrap.**
`format_search_term` (`$SP/beetsplug/_utils/musicbrainz.py:651-663`):
`if not (term := term.lower().strip()): return ""`, then
`LUCENE_SPECIAL_CHAR_PAT.sub(r"\\\1", term)`, then `f"{field}:({term})"`
when `field` is non-empty. **The wrapper is applied after the escape**,
so the wrapping `:` and parentheses are never escaped while a `:` or `(`
inside a term is. Committed evidence:
`release%3A%28frank+zappa%5C%3A+cucamonga%29`.

**C7. Terms join with a single space, and empty terms are filtered
out.** `" ".join(filter(None, starmap(self.format_search_term,
filters.items())))` (`$SP/beetsplug/_utils/musicbrainz.py:679-681`). No
literal `AND`.

**C8. `get_album_criteria` builds an insertion-ordered dict, and its
order is the query's term order.**
`$SP/beetsplug/musicbrainz.py:714-734`: `{"release": album}` first, then
`{"arid": VARIOUS_ARTISTS_ID}` when `va_likely` else `{"artist":
artist}`. `VARIOUS_ARTISTS_ID` is the literal
`89ad4ac3-39f7-470e-963a-56509c546377` (`:43`). Then one entry per
configured extra tag, in the operator's own `musicbrainz.extra_tags`
order, filtered to `FIELDS_TO_MB_KEYS` membership by
`extra_mb_field_by_tag` (`:698-712`). `FIELDS_TO_MB_KEYS` is **eight**
entries (`:47-56`): `barcode`→`barcode`, `catalognum`→`catno`,
`country`→`country`, `label`→`label`, `media`→`format`, `year`→`date`,
`tracks`→`tracks`, `alias`→`alias`. Three value rules inside the loop
(`:721-732`): `tracks` becomes `str(len(items))`; `alias` becomes the
album name; everything else becomes
`str(util.plurality(i.get(tag) for i in items)[0])`, with `catalognum`
additionally having every space removed.

**C9. The pre-filter lowercases and strips once more and drops the
empties.** `get_search_query_with_filters`
(`$SP/beetsplug/musicbrainz.py:736-753`) returns `("", {k: _v for k, v in
criteria.items() if (_v := v.lower().strip())})`. The query *text* half
of the tuple is always the empty string for MusicBrainz. So each value is
lowered and stripped **twice**, once here and once in C6.

**C10. `va_likely` is a three-way disjunction over the items.**
`$SP/beets/autotag/match.py:451-455`: `(not consensus["artist"]) or
(search_artist.lower() in VA_ARTISTS) or any(item.comp for item in
items)`. `VA_ARTISTS = ("", "various artists", "various", "va",
"unknown")` (`$SP/beets/autotag/distance.py:29`). `search_artist`
defaults to `likelies["artist"]` (`match.py:407`, `:443-446`), which
`get_most_common_tags` may already have overwritten with
`likelies["albumartist"]`, while `consensus["artist"]` is computed from
the raw per-item `artist` values and is never updated by that override.
`internal/engine/matcher/likelies.go:47-83` already reproduces both
halves, override at `:78-80`.

**C11. `search_limit` is 5, and its legacy spelling is
`musicbrainz.match.searchlimit`.** `MetadataSourcePlugin.__init__` adds
`search_limit: 5` (`$SP/beets/metadata_plugins.py:201`); `_search_api`
reads `self.config["search_limit"].get(int)` (`:415`) into `SearchParams`
(`:416`); `get_search_response` forwards `limit=params.limit`
(`$SP/beetsplug/musicbrainz.py:755-762`). The legacy read is
`self.config["match"]["searchlimit"]` (`:292-294`) where `self.config` is
`beets.config[self.name]` (`$SP/beets/plugins.py:200`), so the key
resolves as `musicbrainz.match.searchlimit`. **beets' own deprecation
message at `:297` spells it `'musicbrainz.searchlimit'`, which is where
every wrong copy of this fact came from.** All 48 committed searches
carry `limit=5`, so the corpus pins the default and exercises no
override.

**C12. The transport policy, number by number.** User-Agent set once on
the session (`$SP/beetsplug/_utils/requests.py:76`);
`Retry(total=6, backoff_factor=0.5, status_forcelist=[500, 502, 503,
504, 429])` (`:78-88`); `RateLimitAdapter(rate_limit=0.25,
max_retries=retry)` (`:89`) enforcing a 0.25s process-wide floor between
any two requests under a lock (`:128-135`); `kwargs.setdefault("timeout",
10)` then `r.raise_for_status()` (`:99-101`). **The 0.25s adapter is the
secondary floor, not the binding one** (fold M-b): for `musicbrainz.org`
the session's own limiter binds first at one request per second, and C20
states that separately so no implementer reads 0.25s as the wire pace.
The backoff schedule is
`$SP/urllib3/util/retry.py:309-326`: zero when the consecutive-error run
is one or less, then `0.5 * 2**(n-1)`, capped by `DEFAULT_BACKOFF_MAX =
120` (`:210`) with `backoff_jitter = 0.0` (`:238`), which is 0, 1, 2, 4,
8, 16 seconds. `respect_retry_after_header` defaults `True` (`:234`) and
`get_retry_after` reads the header (`:349-357`), so a `Retry-After` wins
over the computed backoff. **`parse_retry_after` accepts two forms**
(`:328-347`): a bare integer is delta-seconds, and anything else is
parsed as an HTTP-date and reduced to `retry_date - time.time()`. The
second form needs a clock read, which is why R13 records it as a
divergence rather than a port. **A sixth status is retryable in
principle**: `Retry.is_retry` also retries any status in
`RETRY_AFTER_STATUS_CODES = frozenset([413, 429, 503])` (`:202`) when the
response carries a `Retry-After` header (`:420-431`), so 413 plus
`Retry-After` is retryable in urllib3 and unreachable from MusicBrainz
(fold m3). **Retries bypass both rate limiters**, because
`Retry` lives inside the adapter's own `urlopen` loop, so only the
outermost attempt is metered.

**C13. `album_for_id` is fetch, translate, and a conditional second
fetch.** `$SP/beetsplug/musicbrainz.py:765-800`: `_extract_id` first,
returning `None` and issuing **no request** when the id does not match;
then `get_release` inside `try/except HTTPNotFoundError`, whose 404
returns `None` (`:777-781`); then `album_info(original_release)`; then,
when `original_release["status"] == "Pseudo-Release"` and the first
`release_relations` entry with `type == "transl-tracklisting"` and
`direction == "backward"` exists, a **second** `get_release` of the
linked release and `_merge_pseudo_and_actual_album` (`:785-798`). The
second `get_release` sits outside the `try`, so its 404 escapes
`album_for_id` entirely. `_extract_id`'s pattern is
`(\w{8}(?:-\w{4}){3}-\w{12})` (`$SP/beets/util/id_extractors.py:20`),
applied at `$SP/beets/metadata_plugins.py:272-278`.

**C14. A search's per-id fan-out follows the search response's order.**
`candidates` runs the search then
`filter(None, self.albums_for_ids(r["id"] for r in results))`
(`$SP/beets/metadata_plugins.py:442-446`), a generator over
`album_for_id` (`:250-259`). With one loaded source the
`as_completed` loop yields one future, so ordering is exactly the
response's (`:91-108`). All three committed profiles pin `plugins:
musicbrainz`, so the corpus is single-source by construction.

**C15. What a failure does, and what it does not raise.** 404 on a
release lookup is caught and drops the candidate silently
(`$SP/beetsplug/musicbrainz.py:777-781`); 404 on a search propagates out
of `search` and is caught by `_search_api`'s bare `except Exception`,
returning `()` (`$SP/beets/metadata_plugins.py:419-427`); exhausted
retries raise `requests.exceptions.RetryError`, which is **not** an
`HTTPError`, so `handle_http_error`
(`$SP/beetsplug/_utils/requests.py:191-202`) never maps it. **The
consequence for the conformance gate:** a stub 404 during a lookup drops
a candidate silently and a stub 404 during a search yields an empty list
silently, so the zero-miss rule is the only thing that turns either into
a visible failure.

**C16. `search_query_ascii` is a structural no-op for MusicBrainz.**
`_search_api` unidecodes `query` (`$SP/beets/metadata_plugins.py:412-413`),
the MusicBrainz plugin returns `query = ""`
(`$SP/beetsplug/musicbrainz.py:750`), and `get_search_response` reads
`params.filters` and `params.limit` and never `params.query` (`:755-762`).
It is recorded as a checked non-membership so a later reader does not add
the knob to `Tunables` for completeness.

**C17. The corpus's request census, measured off the tree at HEAD
`d518e57`.** 215 cases, 351 logged requests, all `GET`; 303 release
lookups (`/ws/2/release/<mbid>`) and 48 searches (`/ws/2/release?`); 167
cases with no search; one distinct `inc=` value, one distinct `limit=`
value (`5`), `fmt=json` on all 351; 46 distinct search `path_query`
values with exactly two filter shapes, `release`+`artist` on 43 and
`release`+`arid` on 5; **zero cases carrying an `extra` key**. Per-case
shape: 164 cases at 0 searches / 1 lookup, 3 at 0/2 (the pseudo double
fetch), 26 at 1/1, 4 at 1/2, 9 at 1/5, 9 at 1/6.

**C18. Go's `net/url` reproduces the committed canonical form byte for
byte, with no adapter.** Probed offline at drafting time against three
committed shapes (the maximal-`inc` lookup, the escaped-quote-and-comma
search, and the VA `arid` search): `url.Values.Encode()` produced the
tree's own `path_query` strings exactly. `Encode` sorts by key while
cassette sorts whole `key=value` pairs
(`internal/harness/cassette/cassette.go:214-228`), which coincides for
today's parameter names and is not guaranteed by the rule; the
coincidence is recorded rather than relied on silently.

**C19. A search response is an object, and the ids live under
`releases`** (fold M-e). `search` ends
`return self._get_resource(entity, **kwargs)[f"{normalised_entity}s"]`
(`$SP/beetsplug/_utils/musicbrainz.py:684-685`), so for `entity =
"release"` it indexes the decoded body's `"releases"` key. Measured
mechanically over all 48 committed search responses (top-level key names
only, no body read into context): every one carries exactly `count`,
`created`, `offset`, `releases`. The Go client therefore decodes

```go
struct {
    Releases []struct {
        ID string `json:"id"`
    } `json:"releases"`
}
```

and ignores `count`, `created` and `offset` at rung 3. Decoding a search
body into a bare array fails all 48 searches in Task 7, which is the
rediscovery this contract exists to prevent.

**C20. The wire pace beets actually holds against `musicbrainz.org` is
one request per second, not four** (fold M-b). The `musicbrainz.org`
branch pins `self.api_host, self.rate_limit = "https://musicbrainz.org",
1.0` (`$SP/beetsplug/_utils/musicbrainz.py:597-599`) and
`create_session` builds `LimiterTimeoutSession(per_second=self.rate_limit)`
(`:612-613`), whose `requests_ratelimiter` mixin defaults `burst: float =
1` (`$SP/requests_ratelimiter/requests_ratelimiter.py:35`) and
constructs one bucket of `per_second * burst` over `Duration.SECOND *
burst` (`:47-51`). Both limiters apply and the tighter one binds, so the
effective pace is 1/s and C12's 0.25s adapter never becomes the
constraint. Any other host derives a float instead,
`ratelimit / ratelimit_interval` (`:600-606`), which is why the shipped
Go seam takes an interval rather than an integer rate. The repo's own
live spender already holds the same line for the same reason:
`cmd/harness/capture.go:27-31`, `requestInterval = time.Second`, "at most
one request per second against the live API, regardless of MusicBrainz's
own configured rate limit". `musicbrainz.DefaultInterval` is
`time.Second`, carrying this citation.

## Resolutions this plan takes rather than discovers

The planning inventory returned fourteen deduplicated open questions
(`docs/superpowers/research/2026-09-04-rung-3c-planning-inventory.md:1067-1142`).
Each is resolved here, with its evidence and its rejected alternative. A
task that believes a resolution is wrong reports a finding and implements
the record anyway.

### R1: request construction, not the request set (inventory Q1)

**Resolved: 3c delivers request-*construction* grading and hands the
request-*set* expectation forward to 3d, and the seam is written into the
grader's own doc comment.**

The set is a function of the ID-match early return, `_recommendation`,
and `_add_candidate`, all 3d scope by the pass-split table (spec `:331`).
3c can grade "given these inputs, this exact `path_query`"; it cannot
grade "these three requests and no others".

The seam, stated exactly. For each committed case the grader reads the
case's own request log to learn **which** requests to issue: every
distinct release MBID the log names, and one search if and only if the
log carries a `/ws/2/release?` row. It then builds each of those requests
through the Go client and requires the stub to resolve every one at zero
misses. **Nothing about a request's bytes comes from the log**: the
lookup's path is built from the MBID, the `inc=` set is the engine's own
list, and the search query is built entirely from the case's own
`input.items` through `matcher.ProjectItemFields` and
`musicbrainz.AlbumSearchFor`, which resolves the likelies and the
various-artists question inside the package (fold m19; revision 1 named
`matcher.MostCommonTags` and `musicbrainz.VALikely` here, which
contradicted Task 7's own Consumes list and would have exported a
package-internal helper with no out-of-package caller). Only the
*decision to issue* is read
from the log, which is legitimate because the log is the committed
expectation and the thing under test is the bytes. 3d replaces that
decision with the real flow and inherits the same stub, the same zero-miss
rule, and the same client.

Rejected: deriving the decision from a partial Go flow in 3c (it would
build half of 3d's `_recommendation` in the wrong pass and grade it
against nothing); comparing built strings to the log's strings directly
(weaker than resolving through the stub, which is the assertion tier 1
will actually make).

### R2: the limiter's clock, and its shape (inventory Q2)

**Resolved: the limiter takes no clock read at all. It is paced by a
ticker it owns, holds one spare token so an idle limiter does not delay
the next request, and its tests replace the ticker channel with a
hand-driven `chan time.Time`. `checkNoTimeNow` is not amended, not
exempted, and not narrowed.**

`checkNoTimeNow` flags only a selector call named `Now` on the `time`
package (`internal/harness/conform/arch_test.go:99-113`) and roots at
`internal/engine` (`:152-157`), so `time.NewTicker`, `time.NewTimer` and
`time.Since` all pass as written. That is not a licence: a green walk is
not proof a file is clock-free, and this plan's answer is that the file
genuinely is.

**The walk covers `_test.go` too** (fold m10). `checkNoTimeNow` skips
only a directory named `testdata` (`arch_test.go:81-86`) and parses every
other `.go` file under the root, so `limiter_test.go` and
`client_test.go` are inside it as much as their production siblings. A
backoff-schedule or elapsed-time assertion is the natural thing to reach
for in both, and it reds the gate for every lane at once. Tasks 5 and 6
each state that their test files call `time.Now` nowhere, and each
observes timing through an injected seam instead (the hand-driven tick
channel, and `Client.sleep`).

**The same walk covers `client.go`, which is why R13 gains a third
divergence** (fold M-c). R2 answers the clock question for `limiter.go`
alone; the client has it too, through `Retry-After`'s HTTP-date form,
and R13 records the narrowing rather than leaving it to be discovered
against a red gate.

The design. One `*time.Ticker` at the wire interval feeds a dispatcher
goroutine. On each tick the dispatcher hands the token to the
highest-priority non-empty class, except that every
`reservationInterval`-th served token is reserved for the classes below
the one strict priority would have served, taken in round-robin (the
anti-starvation floor; see the correction below). When no waiter is
queued the dispatcher keeps the token as a single spare, so a caller
arriving after an idle interval proceeds immediately. That spare is what
reproduces beets' own behavior, whose `RateLimitAdapter` waits
`rate_limit - elapsed` and therefore never delays a request that follows
a long gap (C12, `$SP/beetsplug/_utils/requests.py:128-135`), and it is
`requests_ratelimiter`'s `burst=1` restated without a clock.

**The reserved token's target is corrected here, and the correction is
the pass's one blocker fold** *(conductor ruling 2026-09-04: taken), fold
B1)*. Revision 1 sent every reserved token to the **lowest**-priority
non-empty class. With `ClassUserPresent` saturated (rung 7's probe) and
`ClassBackground` non-empty (rung 4's shadow), `ClassBatch` then receives
nothing at all: it never wins the normal token, because
`ClassUserPresent` outranks it, and it never wins the reserved one,
because `ClassBackground` is lower. `ClassBatch` is this plan's own name
for rung 9's timer import, so revision 1's rule starved the exact class
the build spec's constraint protects
(`specs/2026-09-02-greenfield-server-build.md:693-694`, "the concurrency
model decision ... with the constraint that interactive work must not
starve timer imports"). The corrected rule reserves the token for the
non-empty classes strictly below the class that would otherwise be
served, cycling among them, so a saturated probe stream alternates
`ClassBatch`, `ClassBackground`, `ClassBatch`. The conductor rules
between that and two alternatives: keeping "lowest non-empty" and
recording `ClassBatch` starvation as a stated limitation rung 9
re-verifies, or a weighted fair queue, which the `reservationInterval`
comment already rejects because it makes a probe wait its weighted turn
behind a batch.

**The pace is an interval, and its default is named** *(provisional,
conductor ruling owed, fold M-b)*. `NewLimiter` takes a
`time.Duration` rather than an integer rate, because beets' own
non-`musicbrainz.org` branch derives a float
(`$SP/beetsplug/_utils/musicbrainz.py:600-606`) and an integer forbids a
deliberately conservative sub-1/s setting. `DefaultInterval` is
`time.Second`, carrying C20's citation chain. The conductor rules whether
3c ships that default at all or leaves every caller to name its own
interval; the recommendation is to ship it, because rung 4's shadow is
the first real spender and a plan that names no pace invites `4`.

The goroutine needs its written reason, because go-conventions rejects
gratuitous goroutines: a rate limiter's whole job is to make concurrent
callers wait in a defined order, and rung 7's user-present probe, rung 9's
batch import, and rung 4's shadow are three real concurrent callers named
in the specs. That reason goes on the type.

Rejected: injecting a clock seam (works, but a hand-driven tick channel is
strictly more deterministic and needs no seam at all); widening the
walk's exemption (weakens a checker that catches exactly the defect it was
built for, and would be a conductor ruling amending a landed
architectural check rather than an implementer choice);
`golang.org/x/time/rate` (a network fetch and a `tidy-check` pass for
about forty lines of hand-rolled dispatcher, against a Global Constraint
that adds no dependency).

### R3: `internal/harness/corpus_test.go`'s owner (inventory Q3)

**Resolved: written already; no conductor edit owed.** The parallel
spec's contested-surface table carries the row "`internal/harness/corpus_test.go`
(conductor ruling, 2026-09-04) | rung 3's lane while rung 3 runs; lane E1
otherwise, as the corpus's own wiring | every gate-resident ledger
check". 3c writes it three times (the schema constant, the consumed-file
test, and the new round-trip proof), all inside that grant.

### R4: the schema bump and E1's acknowledgement (inventory Q4)

**Resolved: `schema_version` moves `"3"` to `"4"`, in place, at
`internal/harness/corpus_test.go:160` and `corpus/manifest.json:2`. No
second copy of the constant is created and the symbol is not relocated.**

The precedent is recorded in two halves at `corpus/README.md:1959-1980`.
The 2b bump (`"1"`→`"2"`) moved what `cassette.Expected` carries and
forced a full re-freeze. The 2d bump (`"2"`→`"3"`) added a manifest pin
and forced none, and the README states the rule: "a schema bump forces a
re-freeze only when it moves what `Expected` itself carries, never when
it only adds a manifest pin nothing in `Expected` echoes." 3c's change is
a third shape, a `Case` field rather than an `Expected` or a `Manifest`
field, and it lands on 2d's side of the rule **for a mechanical reason
rather than a categorical one**: `Extra` is `omitempty` and every
committed case's value is nil, so the encoded bytes are identical under
both element types and the migration set is empty (C17: zero of 215 cases
carry an `extra` key). The plan says this plainly rather than claiming
`Case` is exempt.

E1's acknowledgement is already written: the `{cassette,stub,analysis}`
row names "rung 3 pass 3c makes the ruled `Case.Extra` schema-and-stub
change". The manifest row is the one conductor-owed edit, above.

The spec names a fallback if the schema change is declined (spec `:263`),
and it is written here rather than left implied: **if the conductor
declines the bump, tier 1's zero-miss rule stands exactly as written,
`Extra` stays unreachable through the stub, Tasks 1 and 2 are struck, and
the U8 recall-extension surface is recorded in HISTORY as ungraded at
rung 3 with that reason.** The conductor has already ruled the opposite
(spec `:45-50`), so this paragraph exists as a record, not as a live
branch.

Rejected: exporting `wantManifestSchemaVersion` from `cassette` as
package API (a permanent exported surface for a test-only value with one
consumer, against go-conventions, and against the two prior bumps' own
precedent).

### R5: the empty-migration proof (inventory Q5)

**Resolved: the proof is an offline round-trip test over all 215
committed cases, not a `make derive` run.**

`TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema` loads each
committed `case.json` through `cassette.LoadCase`, writes it through
`Case.Write` into `t.TempDir()`, and byte-compares the result against the
file on disk. That proves exactly the fact the bump rests on, in seconds,
with no venv and no network, and it keeps proving it for every later
schema change rather than for this one only. 2d's own precedent used a
full re-derive because its change touched `Reconcile`'s version probe,
which only `make derive` exercises; 3c's touches no derivation path at
all (`writeShimInput` marshals `{CaseID, EvalDate, Items}` and nothing
else, `internal/harness/oracle/oracle.go:280-289`, and no Python file in
the repo mentions `extra`).

Rejected: running `make derive` (≈585s to 17m50s of oracle suite for a
fact a two-second test settles, and this pass otherwise touches no oracle
territory at all).

### R6: symmetric hit tracking on the request map (inventory Q6)

**Resolved: `Consumed()` covers `Extra` only. 3c does not ship an
`Unused()` or `Served()` accessor for the request map, and the blind side
is recorded as a named carry to 3d with its one-line implementation
sketch.**

The blind side is real: nothing today reports which logged `Request`
entries a run actually consumed, so a client issuing *fewer* requests
than the log records produces a green tier 1
(`internal/harness/stub/stub.go:93-105`). It is not reachable inside 3c,
because 3c's grader enumerates the request set itself (R1) and asserts a
per-case issued count against the log's own count, so under-issuance
fails on that assertion rather than passing quietly. The checker that
would need the accessor is 3d's, where the flow decides what to issue.
Shipping an exported accessor with no consumer is dead surface under
go-conventions, and 3d can add it in four lines (a `map[key]bool` hit set
beside the response map, set in `ServeHTTP`, and one accessor returning
the unhit `Requests` entries).

Rejected: shipping it now because it is cheap (cheapness is not a reason;
the spec deliberately grades only zero-misses because a strong ID match
legally orphans search captures into `Extra`, spec `:265`, and the
symmetric rule needs 3d's flow to be stated correctly at all).

### R7: `Case.Extra`'s producer (inventory Q7)

**Resolved: `Extra` has no producer today and 3c does not build one. The
handoff is named: a producer is `cmd/harness` work, lane E1's
territory.**

Nothing writes the field: the only writes in the repo are in
`cassette_test.go`, and `cmd/harness` never sets it. The mechanism
`corpus/README.md:2775-2785` describes is a maintainer's manual
bookkeeping act, so the honest statement is stronger than "the channel is
dormant": **the channel has no producer.** A `Request`-shaped `Extra` is
the prerequisite for the capture or freeze loop ever becoming one, which
is part of why the reshape is worth landing now while the migration is
empty. Task 8 records this so a later pass does not assume the loop fills
it.

### R8: `ItemFields.Comp` and `IsVariousArtistsName` (inventory Q8)

**Resolved: both join `internal/engine/matcher`, in the shared types,
with their reasons on them. 3c writes two `matcher` files plus their
tests, and the plan names that rather than letting it read as drift.**

`va_likely` reads `item.comp` (C10) and `ItemFields` has no `Comp` field
(`internal/engine/matcher/info.go:124-143`), so the query cannot be built
without it. `ItemFields`' own doc comment says "A field outside this set
is not read by the distance path; adding one is a finding, not a
convenience" (`:111-113`), and that sentence stays true only if the
addition is named: the amended comment declares nineteen fields, eighteen
read by the distance path and `comp` read by the **search** path alone,
with the citation. Measured reach: 15 of 215 committed cases carry `comp`
on their items, all 15 with `comp: true` (this plan's header correction).

`vaArtists` is unexported at `internal/engine/matcher/track_artist.go:21`
and `internal/engine/musicbrainz` cannot read it. Export
`matcher.IsVariousArtistsName(credit string) bool` and have
`addTrackArtist` (`:33`) call it, so one list has two consumers.

Rejected: a second private item decoder in `musicbrainz` over the same
item bag (two projections of one type is exactly the duplication the
`releaseLookupPrefix` note warns about, and it would drift silently);
duplicating the `VA_ARTISTS` list in `musicbrainz` (a second copy of a
five-element literal with no independence argument behind it, unlike R9's
copy, which has one).

### R9: the `/ws/2` constants and the `releaseLookupPrefix` triplication (inventory Q9)

**Resolved: 3c does NOT unify them. The engine owns its own
`releaseIncludes` and its own path construction, `internal/harness/mbrequest`
keeps `MaximalReleaseInclude`, and the rationale is written at both
sites.**

The architectural rule permits the collapse in one direction only
(`internal/engine` never imports `internal/harness`,
`internal/harness/conform/arch_test.go:141-146`), so the only legal
shared home is the engine owning the constant and `mbrequest` importing
it. **That legal direction would vacuum the grading.** `mbrequest` is
what `cmd/harness/capture.go:279` and `internal/harness/synth/synth.go:41`
use to *produce* a committed request; the committed log is the
expectation and the engine's construction is the system under test.
Sharing one definition between them turns Task 7's whole checker into a
self-derived expectation, which go-conventions names by that phrase as an
anti-pattern. The two copies must be independent, and the conformance run
is the mechanism that proves they agree.

3b's carry says a fourth copy is the trigger to revisit
(`docs/HISTORY.md:352-359`). The trigger is met and the answer is that
the fourth is deliberate and load-bearing. Writing that down at both
sites is what stops a later `code-simplifier` pass from unifying them and
silently emptying the check. Task 8 records it as a rediscover-guard.

One cleanup is genuinely available and is **not 3c's**: folding
`capture.go:279`'s and `synth.go:41`'s inline `"/ws/2/release/"` literals
into `mbrequest`'s own constant. Task 8 files it as a note for the lane
that owns `mbrequest`.

### R10: the capture tool's pacer (inventory Q10)

**Resolved: the limiter coexists with `cmd/harness/capture.go`'s
`sleep(requestInterval)` pacer and does not subsume it.** The capture
tool is `cmd/` territory, outside 3c's declared wall, it is a live
spender with its own identified User-Agent and its own 30s timeout
(`cmd/harness/capture.go:26-41`, `:130-134`, `:312-323`), and it is
lane E1's. Recorded as a note so a later pass does not read the
duplication as an oversight; whether the capture tool ever routes through
the engine's limiter is a decision for whoever next opens `cmd/harness`.

### R11: the `search_limit` legacy spelling, and the User-Agent (inventory Q11)

**Resolved on both halves.**

**The spelling.** The truth is `musicbrainz.match.searchlimit` (C11,
probe-confirmed against `$SP/beetsplug/musicbrainz.py:292-294` and
`$SP/beets/plugins.py:200`), and the reason every copy of this fact is
wrong is that beets' own deprecation message at `:297` spells it
`'musicbrainz.searchlimit'`. **Two sites lose and one does not** (fold
m2). The two that lose are
`internal/engine/matcher/tunables.go:119-124` (code, and therefore 3c's
own edit, in Task 3, whose `:122` reads `musicbrainz.searchlimit`) and
`docs/HISTORY.md:600` (3a's correction #5, whose text is on `:600` rather
than the `:597-599` revision 1 cited, fold m13). The rung 3 spec's
tunable enumeration (`:226`) does **not** lose: it reads
"`musicbrainz.search_limit` with its legacy `match.searchlimit`
spelling", immediately followed by "`musicbrainz.data_source_mismatch_penalty`
with its legacy `source_weight` spelling", so that line spells legacy
keys relative to the plugin namespace throughout, and
`musicbrainz.` + `match.searchlimit` is the truth. Correcting one item
without the other would make the two adjacent entries inconsistent. Task
8 therefore drafts the spec line as an **optional clarification**
(spelling it out as `musicbrainz.match.searchlimit`), never as a
correction, and drops the "losing document" framing for that target
alone. 3c corrects the code and drafts the one document correction plus
the one clarification for the conductor's batch, on 3a's own M2
precedent.

**The User-Agent.** `ClientConfig.UserAgent` is a **required** string and
the engine never builds it; `NewClient` refuses to construct without one.
The universality rule puts the site-specific half in `internal/config`,
and `matcher` already refuses that import for the same reason
(`internal/engine/matcher/tunables.go:34-35`). The placeholder every test
and every conformance run passes is an invented, non-site-specific
literal (`"dubplate-conform/0 ( offline-conformance-run )"` in `conform`,
and a per-test literal elsewhere). **The real value comes from
`cmd/dubplate-server` at composition time**, built as
`dubplate-server/<version> ( <contact> )` on the harness precedent
(`cmd/harness/capture.go:130-134`), with the contact taken from
`Config.AdminEmail`, already a required wall row parsed as an RFC 5322
address (`internal/config/config.go:60-64`). So 3c mints no secret, adds
no wall row, and touches no file rung 1b owns.

**That composition departs from the precedent it cites, and the
departure is recorded rather than implied** (fold m20). `AdminEmail`'s
own doc comment says "adopters compare it, never display it"
(`internal/config/config.go:60-64`), and the harness precedent takes its
contact from a **dedicated** input, `DUBPLATE_HARNESS_CONTACT`, whose
comment says the tool "never invents one (never a literal, per CLAUDE.md's
universality rule)" (`cmd/harness/capture.go:21-26`) and which the tool
refuses to run without. Putting `AdminEmail` on the wire makes a compare-only
value a displayed one. Reusing it is still the cheaper answer for a rung
whose config wall does not exist yet, so the record is: **a dedicated
contact row is the named post-1b follow-on**, listed beside the four-knob
registration in Task 8's forward carries, and `AdminEmail` is the interim
source rather than the settled answer. Nothing in 3c depends on which one
wins, because `ClientConfig.UserAgent` is caller-supplied either way.
**The stub never
reads a request header** (`internal/harness/stub/stub.go:81-91`), so
User-Agent conformance is a unit-test claim and never a stub claim, and
Task 6 says so.

### R12: the corpus-unreachable client surfaces (inventory Q12)

**Resolved: one new declared corpus-coverage gap, two named non-ports,
and two surfaces that are 3d's rather than gaps.**

- **The `extra_tags` criteria branch is rung 3's tenth declared
  corpus-coverage gap.** No committed profile writes a `musicbrainz:`
  block at all, so `FIELDS_TO_MB_KEYS`, the `tracks` and `alias` special
  cases, the `catalognum` space-strip, and the `date:(0)` trap are all
  unreachable by any corpus case. They are graded by a hand-authored
  table in `internal/engine/musicbrainz/testdata`, with the mutation that
  proved each can fire named in Task 4's report. Together with 3a's three
  and 3b's six (`docs/HISTORY.md:101-131`), rung 3 then carries **ten**
  declared gaps, and Task 8 states the running total so 3d does not
  re-count.
- **The recording lookup and `RECORDING_INCLUDES` are a named non-port.**
  Nothing on the album path calls `get_recording`
  (`$SP/beetsplug/_utils/musicbrainz.py:692-697`), singleton matching is
  outside rung 3's five deliverables, and zero committed cases issue one.
  3c ships no `Client.Recording` and no second include list. Recorded as
  a non-port with its reason, not as a gap, because there is nothing
  built to be ungraded.
- **The browse fallback stays unported.** `internal/engine/musicbrainz/doc.go:7-16`
  refuses it and `conform`'s `checkNoReleaseReachesBrowseThreshold` guards
  the line at `>=` 500 against a corpus maximum of 470
  (`internal/harness/conform/translate.go:22-29`, `:396-443`; 3b guard
  14). 3c owns the port only if a corpus case ever crosses, and Task 0
  re-reads the maximum.
- **The explicit `search_ids` branch and `match_by_id`'s
  no-id/no-consensus branch are 3d's flow**, not 3c's construction, and
  are named in Task 8's forward carries rather than counted as 3c gaps.
- **The retry path is ungraded by the corpus and always will be.** The
  stub answers 200 or 404 and nothing else, and `cassette.Request` has no
  status field to express a third (the declined "response-status
  extension", open since 2c). Retry behavior is graded by unit tests
  against an injected `http.Client` transport, which is a different kind
  of evidence rather than a coverage gap. 3c does not take the
  response-status extension, and Task 8 records that it stays declined.

### R13: retry semantics (inventory Q13)

**Resolved: reproduce beets' policy, with exactly three named deliberate
divergences and their reasons on the code** (fold M-c raised the count
from two).

Faithful: seven attempts (one plus six retries), the same five
`status_forcelist` statuses (429, 500, 502, 503, 504), backoff of zero
before the first retry then `500ms * 2^(n-1)` capped at 120s with no
jitter, and a delta-seconds `Retry-After` honored over the computed
backoff when the response carries one (C12).

Divergence 1, **the timeout is 30s rather than beets' 10s**, matching the
capture tool's own stated reason that maximal-`inc` release lookups are
large (`cmd/harness/capture.go:38-41`); the largest committed response
body is 8,894 KB.

Divergence 2, **every retry re-acquires a limiter token.** beets gets the
opposite for free because `Retry` sits inside the adapter beneath both
limiters, so only the outermost attempt is metered (C12). A retry is a
real request against a shared wire budget, and rung 4's shadow spends a
metered budget through this limiter (spec `:240`), so a Go client with
the limiter above the loop and no per-attempt acquire would silently
burst. Reproducing beets' accident here would make the budget a lie.

Divergence 3, **`Retry-After` is honored in its delta-seconds form only,
and an HTTP-date `Retry-After` falls through to the computed backoff**
(fold M-c). urllib3 accepts both: `parse_retry_after` reads a bare
integer as seconds and otherwise parses an HTTP-date and computes
`retry_date - time.time()` (`$SP/urllib3/util/retry.py:328-347`), fed by
`get_retry_after` (`:349-357`). The second form needs the current time,
and `internal/engine/musicbrainz/client.go` sits inside
`TestEngineNeverReadsTheClock`'s walk
(`internal/harness/conform/arch_test.go:152-157`), which this plan's
Global Constraints forbid amending. `time.Since` would pass the checker
while breaking R2's own sentence that a green walk is not proof a file is
clock-free, so the honest answer is the narrowing, stated on the code.
The computed backoff is already a bounded monotone schedule, so a date
form degrades to a correct wait rather than to no wait at all.

**One further urllib3 behavior is a deliberate non-port** (fold m3).
`Retry.is_retry` also retries any status in `RETRY_AFTER_STATUS_CODES =
frozenset([413, 429, 503])` when the response carries a `Retry-After`
header (`$SP/urllib3/util/retry.py:202`, `:420-431`), which makes 413
plus `Retry-After` a sixth retryable case beyond the `status_forcelist`.
MusicBrainz's WS/2 never answers 413 to a GET, and the Go client's
five-status list is stated as exhaustive with a Task 6 sentinel test
against it, so the omission is recorded here rather than left as an
unexplained difference from a contract that reads complete.

All three divergences and the non-port are recorded in the client's doc
comment and in Task 8's forward carries, because rung 3's runs are
offline and nothing in the corpus can grade any of them.

Sentinels: `ErrNotFound` for a 404 on a lookup (callers must distinguish
a deleted release from a real failure, because beets' own flow does,
C15), and `ErrUnavailable` for retries exhausted on any retryable status
or transport error (rung 4's shadow branches on it to defer rather than
record a divergence). A rate-limit-specific sentinel is deliberately not
shipped and the omission is recorded: `%w` is an API commitment
justified by a caller that branches, and no caller branches on 429
separately from 503 today.

### R14: the query path's case mapping and its strip set (inventory Q14)

**Resolved: the 56 case-mapping divergences are recorded as a declared
divergence with a passing named test; the four-code-point strip
divergence is closed outright. No dependency is added.**

3a's sweep over all 1,112,064 valid code points found 56 where Go's
`strings.ToLower` disagrees with Python's `str.lower`; 54 are rescued by
the distance pipeline's post-transliteration `[a-z0-9]` strip and two are
not (`internal/engine/matcher/testdata/case-mapping-sweep.tsv`,
`docs/HISTORY.md:450-462`). **The 56 are two different phenomena, and
this plan says which** (fold m11). The sweep file's own header settles it
(`case-mapping-sweep.tsv:7-10`): U+0130 alone is the full-versus-simple
case-mapping difference the sweep went looking for, and the other 55 are
a Unicode version skew, Go 1.27 carrying Unicode 17.0.0 case tables
against CPython 3.13's 15.1.0. That distinction is load-bearing for the
query path, because the version-skew half moves the day either toolchain
moves while the one asserted row does not, so `lowerSearchTerm`'s doc
comment carries both halves and Task 4's fixture asserts the stable
member. **The query path has no strip**, so all
56 reach the wire. Adopting `golang.org/x/text/cases` would close it and
is rejected: it violates the no-new-dependency constraint, and the
divergence changes which candidates MusicBrainz returns rather than any
distance the conformance gate grades. Task 4 ships `lowerSearchTerm` as a
named helper citing the sweep, with a test asserting the known outcome on
`İ` (U+0130, which Go lowers to a single `i` and Python to `i` + U+0307)
so the divergence is a recorded passing assertion rather than an
unrecorded surprise. The corpus's one non-ASCII search term
(`artist%3A%28%E6%B5%9C%E5%B4%8E%E3%81%82%E3%82%86%E3%81%BF%29`) is
caseless and unaffected.

The strip divergence is different and is closed rather than declared,
because it is one line. Python's `str.strip()` treats U+001C through
U+001F as whitespace and Go's `unicode.IsSpace` does not. Task 4 ships
`trimSearchTerm` using `strings.TrimFunc` with `unicode.IsSpace(r) || (r
>= 0x1c && r <= 0x1f)` and a table test naming the four code points.

The case-mapping half is the third member of the family 3b's guard 11
opened (Python `re` and full-case semantics against Go RE2 and simple
case mapping), with the Unicode-version-skew half riding beside it as a
distinct hazard. The fourth member, `_extract_id`'s `(\w{8}(?:-\w{4}){3}-\w{12})` where Python's `\w`
is Unicode-aware and Go's is ASCII-only (C13), is recorded in Task 6 the
same way.

### R15: `dataSourcePenalty`, 3a's carry to 3c

**Resolved: closed in this pass. The fallback becomes a package constant
with the citation on it, and `Tunables.DataSourceMismatchPenalty` is
wired into `DefaultTunables`'s `DataSourcePenalties` entry so the field
keeps a reader.** *(The wiring half is provisional, conductor ruling
owed, fold M-a.)*

**Revision 1 asserted a relationship the code does not have.** It said
the field's "real job" is "populating `DataSourcePenalties["MusicBrainz"]`",
and `DefaultTunables` writes the two as independent literals
(`internal/engine/matcher/tunables.go:282-285`). Once
`dataSourcePenalty` returns the constant, the field's readers are the
`DefaultTunables` literal and one `tunables_test.go` assertion and
nothing else: `golangci-lint`'s `unused` does not report an exported
field of an exported struct, so the gate would stay green over a tunable
that configures nothing. Two endings are honest and the conductor picks
one.

- **Wire it** (recommended, and what Task 3 builds provisionally): build
  the map inside `DefaultTunables` from the field,
  `DataSourcePenalties: map[string]float64{"MusicBrainz":
  t.DataSourceMismatchPenalty}`, which makes the sentence true and is
  what beets actually does, since the per-source map entry for
  MusicBrainz **is** the plugin's own `source_weight`-first resolved
  penalty (`$SP/beets/metadata_plugins.py:182-195`). It is
  behavior-neutral on the committed record at 0.5, and 72/72 tier-0 rows
  staying green is the proof.
- **Document it as a readerless seam**: say in the doc comment that the
  field is the config-load seam for
  `musicbrainz.data_source_mismatch_penalty` / `source_weight` whose only
  consumer today is `DefaultTunables`'s own map, and carry that forward
  to the rung-1b registration task.

Either way, **the two doc comments this change falsifies are Task 3's to
fix** (fold M-a), and both are named in its Files list:
`internal/engine/matcher/penalty.go:334-337` on `AddDataSource`, which
still promises callers the fallback is
`Tunables.DataSourceMismatchPenalty`, and
`internal/engine/matcher/tunables.go:138-140` on `DataSourcePenalties`,
which still says `AddDataSource` scans it "before falling back to
DataSourceMismatchPenalty". Neither vale nor golangci-lint reads a stale
prose claim, so shipping them uncorrected would commit two
contradictions of the code on a rung whose whole premise is that a
divergence between the record and the port is the defect class.

`get_penalty` falls back to the **class literal**
`MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY`, which is
`0.5` (`$SP/beets/metadata_plugins.py:160-169`, `:180`), while the Go
port falls back to the configured value
(`internal/engine/matcher/penalty.go:404-409`). The configured value is a
different thing: it is the *plugin's own* penalty, resolved
`source_weight`-first (`:190-195`) and wired into the plugin's config at
init (`:201-202`), which is what populates the per-source map rather than
the fallback. Making the two roles explicit is faithful to beets and
makes the field's meaning legible, which the current shape does not.

**Tier 0 is blind to this today, and the fix removes the blindfold.**
`conform`'s decoder already compensates: `replayAddDataSource` sets
`tun.DataSourceMismatchPenalty = 0.5` explicitly
(`internal/harness/conform/component.go:721`), and `replayTrackDistance`
does the same inside its own branch (`:673`), each with a comment saying
the fallback "is always the class default 0.5, never the row's own
penalty field". Those two lines exist only to compensate for the
divergence being fixed here, so Task 3 removes both, and **all 72 tier-0
rows staying green afterward is the proof that the fix is behavior-neutral
on the committed record** rather than a claim.

The branch is unreachable under the greenfield server as built, because
the only candidate source is the literal `"MusicBrainz"`
(`internal/engine/musicbrainz/translate.go:13`) and the map names it,
which is why nothing caught it. The mutation that proves the fix fires is
hand-authored: set `DataSourceMismatchPenalty` and
`DataSourcePenalties["MusicBrainz"]` both to `0.7` and call
`AddDataSource("Discogs", "Bandcamp")`; beets yields `0.5`, the code as
landed yields `0.7`.

Rejected: handing it to 3d (defensible, and it would leave 3a's carry
open across two more passes for a three-line change whose own grading
instrument, tier 0, is in this pass's tree already).

### R16: the four-knob wall registration

**Resolved: still deferred, and the reason is re-verified rather than
assumed.** `internal/config/keys.go:19-25` still carries `ownedWall` as a
slice literal at HEAD `d518e57`, unchanged at `f2bc19e`, so rung 1b has
not landed. The
registration of `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh` and
`SearchLimit` stays the named follow-on task for the first rung 3 pass
that runs after 1b closes, and `internal/config/keys.go` stays the file
no lane edits until then. One fact worth carrying forward for whoever
does it: `Threshold.Default` is a string and `ParseInt` is the only
parser (`internal/config/thresholds.go:43-60`), so the three float rec
thresholds have no resolver even once 1b lands, and the registration task
needs a float variant.

## Task 0: The staleness pre-task

Model: `sonnet`, **fresh context**, dispatched before any implementer.
Reads this plan and the rung 3 spec against the tree as shipped. **Its
output is blockers with evidence, not fixes**; it writes no code and edits
no file. The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is** (fold m12). This plan's citations were verified at `d518e57`,
and two STATUS-only commits have landed since (`f2bc19e` at drafting
close). A citation that has moved is a finding for this task, not a
surprise for an implementer.

The checklist is the spec's own (`:395-406`), per pass. The first three
are blockers: a "no" stops the launch rather than producing a note.

1. **Have Phase 0a, 0b, and 0c landed?** 0a's `(5,3)` and `(8,3)` cells
   clear rung 3 to run out of ladder order; 0b owns the wall the four
   knobs register on; all three gate the second code lane. *(Expected at
   drafting time: 0a landed; **0b and 0c unlanded, which is the expected
   state and not a blocker for 3c**, since 3c registers nothing on the
   wall, writes no `internal/config`, and needs no worktree. Report the
   state; do not stop on it.)*
2. **Has rung 1b's registry shipped?** If not, which is the expected
   answer because `internal/config/keys.go:19-25` still carries
   `ownedWall` as a slice literal, confirm that the four-knob deferral is
   still recorded in 3a's HISTORY guard 2 and that `internal/config` is
   untouched by this plan (R16).
3. **Is the lane registry consistent with where this pass intends to run,
   and is the main tree actually free?** `docs/STATUS.md`'s lane registry
   is the live-executor check, never `pgrep`. *(Expected: `docs/STATUS.md`
   reads "lane-1 (code, main tree): PREP for **pass 3c**", naming the
   plan-prep workflow `wf_ede6e00d-bee`, which is the workflow this plan
   is the output of, and records no pid. Flipping lane-1 to a 3c-execution
   entry carrying a pid is a **conductor edit owed before Task 1
   dispatches**, not a finding for the pre-task to litigate. lane-2
   BLOCKED; doc lane IDLE. Report a lane-1 entry naming any workflow other
   than `wf_ede6e00d-bee`, or any live pid in this tree, as a blocker.)*
4. **Have the contested-surface rows been written for every surface this
   pass touches?** *(Expected: six of eight present at
   `docs/superpowers/specs/2026-09-03-parallel-build-approach.md`'s
   contested-surface table, with **`corpus/manifest.json` and
   `corpus/README.md` missing** (fold m22). Report both; the conductor
   writes them; neither is a launch blocker for the reason this plan's
   "Contested surfaces" section states.)*
5. Does every corpus artifact this plan names exist, at the path and
   shape named? Specifically: `corpus/manifest.json` parses as
   `cassette.Manifest` and carries `schema_version` `"3"`; every one of
   the 215 case directories loads through `cassette.LoadCase`; and
   `corpus/analysis/album_info.json` exists (3b's, read by no 3c task but
   part of the gate).
6. Does every harness and engine API this plan calls have the signature
   named? `cassette.Case`, `cassette.Request`, `cassette.LoadCase`,
   `cassette.Manifest`, `(*cassette.Case).Write`; `stub.New`,
   `(*stub.Server).Misses`, `(*stub.Server).Addr`, `(*stub.Server).Close`,
   `stub.Miss`; `matcher.ItemFields`, `matcher.ProjectItemFields`,
   `matcher.Likelies`, `matcher.MostCommonTags`, `matcher.Tunables`,
   `matcher.DefaultTunables`, `(*matcher.Distance).AddDataSource`;
   `musicbrainz.Release`, `(*musicbrainz.Release).Validate`,
   `musicbrainz.ErrMalformedRelease`, `musicbrainz.TranslateAlbum`,
   `musicbrainz.PseudoReleaseTarget`, `musicbrainz.MergePseudoRelease`;
   `conform.CheckCaseWeights`, and `conform`'s unexported `mergeProfile`
   and `releaseLookupPrefix`.
7. **Have any `oracle_bugs.json` or `adjudications.json` dispositions
   moved since this plan was drafted**, and does any movement change 3c's
   scope? *(Drafting-time snapshot: adjudications 21 rows, 18
   `oracle-wrong`, 3 `oracle-correct`, 0 pending; oracle_bugs 6
   hypotheses, 4 `oracle-wrong`, 2 `oracle-correct`, 0 pending. 3c grades
   no verdict and ships no compat switch, so a movement is a note rather
   than a scope change; report it once.)*
8. Have any `ties.json` dispositions or `proximity.json` rows moved? *(3c
   grades neither; 3d inherits both, so a movement is worth reporting
   once.)*
9. Do the corpus's committed figures match this plan's own arithmetic?
   Read and report, **mechanically, reading no response body into
   context**: `ls corpus/cases | wc -l` *(215)*; total logged requests
   *(351)*; release lookups *(303)*; searches *(48)*; cases with no search
   *(167)*; distinct `inc=` values *(1)*; distinct `limit=` values *(1,
   `limit=5`)*; cases carrying an `extra` key *(0)*; search rows carrying
   `artist%3A%28` *(43)* and `arid%3A%28` *(5)*; distinct search
   `path_query` values *(46)*; cases carrying `comp` on any item *(15)*
   and cases carrying `comp: true` *(15)*; search rows carrying an
   escaped hyphen `%5C-` *(**6**, across 6 cases and 5 distinct query
   strings: the five `arid` rows through the various-artists MBID plus
   `real-noise-dylan-bootleg`'s own `release:` clause; fold M-f)*; and
   the largest committed release's summed `media[].tracks` *(470, against
   `browseMaxTracks` of 500)*. A divergence on any of these is a finding,
   because Task 7's floors, Task 7's mutation 2 and Task 4's declared gap
   rest on them.
10. Has `internal/config` changed shape? Have `internal/engine/musicbrainz/
    {query,client,limiter}.go` appeared? *(Expected: no on all four.)*
11. Do the **one code site and the one document** R11 names still carry
    the wrong `search_limit` spelling
    (`internal/engine/matcher/tunables.go:119-124`, whose `:122` reads
    `musicbrainz.searchlimit`, and `docs/HISTORY.md:600`)? The rung 3
    spec `:226` is **not** a wrong spelling and is not checked as one
    (fold m2, fold m13): it spells legacy keys relative to the plugin
    namespace, as its own `source_weight` sibling does, and Task 8 drafts
    it as an optional clarification. If a correction already landed,
    Task 3's and Task 8's drafting halves are smaller and say so.
12. Are the two data-source compensations R15 names still present at
    `internal/harness/conform/component.go:721` and inside
    `replayTrackDistance` at `:673`? Task 3 removes both.

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
required; the task writes nothing.

---

## Task 1: The cassette schema half of the `Extra` channel

Model: `sonnet`. Depends on Task 0. **This task runs first by conductor
ruling** (spec `:45-50`): "B2's cassette schema half is accepted and
lands as pass 3c's first task, where the stub half already lives — the
channel is dormant (zero committed cases carry `Extra`), so the migration
is empty and one pass owns both halves."

**Files:**
- Modify: `internal/harness/cassette/cassette.go` (package doc `:7-9`,
  `Case.Extra` `:47`, `LoadCase`'s doc `:117-121`, `validate`'s Extra loop
  `:169-176`)
- Modify: `internal/harness/cassette/cassette_test.go` (the seven `Extra`
  sites: `:67`, `:181`, `:198`, `:218`, `:238`, `:245`, `:255`)
- Create: `internal/harness/cassette/testdata/extra-request-shape/`
  (`case.json` plus `responses/001.json` and `responses/002.json`)
- Modify: `internal/harness/corpus_test.go`
  (`TestEveryResponseFileIsConsumed` `:127-152`,
  `wantManifestSchemaVersion` `:154-160`, one new test)
- Modify: `corpus/manifest.json` (`schema_version` only)

**Interfaces:**

- Consumes: nothing new. `cassette` imports only the standard library.
- Produces:

```go
// internal/harness/cassette/cassette.go

// Case is one corpus case, loaded from its directory.
type Case struct {
    ID       string    `json:"id"`
    Input    Input     `json:"input"`   // items in pinned order; no file references at 2a
    Profile  string    `json:"profile"` // profile name under corpus/profiles ("estate", "stock")
    EvalDate string    `json:"eval_date"`
    Requests []Request `json:"requests"` // the committed request log: an expectation, not scaffolding
    // Extra names responses a case captured that no logged request
    // consumes: recall-extension fixtures for a future
    // candidate-widening improvement (U8). It carries the full request
    // shape rather than a bare filename because the replay stub keys on
    // (method, path_query) and a second response map cannot be keyed
    // against nothing (rung 3 spec, the B2 fold). Every committed case's
    // value is nil at schema version 4, so the reshape moved no bytes.
    Extra    []Request `json:"extra,omitempty"`
    Expected *Expected `json:"expected,omitempty"` // nil until derived
}
```

**The four trailing comments above are the ones `cassette.go:43-48`
already carries and must survive the reshape** (fold m14). Revision 1's
snippet dropped them, and an implementer pasting that block would have
deleted four landed comments as a side effect of a field's type
changing.

**Behavior contract:**

- **`validate`'s Extra loop applies the same four per-entry rules the
  Requests loop applies, in the same order** (`cassette.go:152-167` is the
  model): `Method` non-empty; `PathQuery` already equal to
  `canonicalizePathQuery(PathQuery)`, with the error naming both the
  stored and the canonical form; `Response` non-empty; the response file
  readable under `dir` through `checkResponseReadable`.
- **Three exclusivity rules are kept or added, and each says why.** The
  existing filename rule stays with its message unchanged
  (`"case.json: extra[%d]: %s already consumed by a request"`), because a
  response named by both a `Request` and an `Extra` would be
  double-counted by `TestEveryResponseFileIsConsumed` and the orphan
  bookkeeping would mean nothing. A **new** key rule rejects an `Extra`
  entry whose `(Method, PathQuery)` pair a `Request` already carries,
  with a distinct message (`"case.json: extra[%d]: %s %s is already a
  logged request"`), because the stub builds two maps on that key and the
  request map would silently shadow the extra entry. A **second new** key
  rule rejects an `Extra` entry whose `(Method, PathQuery)` pair another
  `Extra` entry already carries (`"case.json: extra[%d]: %s %s duplicates
  extra[%d]"`), because the stub's extra map keys on the same pair and
  would shadow one of the two silently (fold m16). That is the defect
  `TestEveryLoggedRequestResolvesToExactlyOneResponse`
  (`internal/harness/corpus_test.go:105-119`) already prevents one map
  earlier, and it covers `c.Requests` only. Three rules, three failures,
  three fixtures.
- **The package doc and `LoadCase`'s doc comment both move with the
  type.** The package doc's sentence "`[Request.Response]` and
  `[Case.Extra]` name those files" becomes accurate for a `Request`-shaped
  `Extra`, and `LoadCase`'s summary gains the two exclusivity rules by
  name.
- **`Write` is not changed.** `json.MarshalIndent` serializes the new
  element type through its own tags, and `omitempty` on a nil slice emits
  nothing, which is the whole basis of R4's empty-migration argument.
- **`internal/harness/corpus_test.go` changes in three places.**
  `TestEveryResponseFileIsConsumed` iterates `c.Extra` as
  `[]cassette.Request` and marks `extra.Response` consumed (`:134-136`).
  `wantManifestSchemaVersion` moves `"3"` to `"4"` in place, and its own
  doc comment gains one sentence naming this reshape as the bump's cause.
  A new test lands beside them:

```go
// TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema proves the
// Case.Extra reshape moved no committed byte: every case.json decodes
// through the current Case type and re-encodes, through the same
// Case.Write every `harness freeze` uses, to bytes identical to the file
// on disk. This is the schema bump's migration proof, offline and in
// seconds, in place of the full re-derivation pass 2b's bump needed
// (corpus/README.md:1959-1980). It keeps proving it for every later
// schema change rather than for this one only.
func TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema(t *testing.T) {
    root := repoRoot(t)
    forEachCase(t, root, func(t *testing.T, id, dir string, c *cassette.Case) {
        want, err := os.ReadFile(filepath.Join(dir, "case.json"))
        if err != nil {
            t.Fatalf("read %s/case.json: %v", dir, err)
        }
        out := t.TempDir()
        if err := c.Write(out); err != nil {
            t.Fatalf("case %s: Write: %v", id, err)
        }
        got, err := os.ReadFile(filepath.Join(out, "case.json"))
        if err != nil {
            t.Fatalf("case %s: read round-tripped case.json: %v", id, err)
        }
        if !bytes.Equal(want, got) {
            // The first differing offset, never a length delta: an
            // equal-length re-encode drift is the likely failure and a
            // length delta prints 0 for it (fold m4).
            at := 0
            for at < len(want) && at < len(got) && want[at] == got[at] {
                at++
            }
            lo, hi := max(0, at-20), min(len(want), at+20)
            ghi := min(len(got), at+20)
            t.Errorf("case %s: round-trip through the current schema first differs at byte %d (want %d bytes, got %d)\n want ...%q...\n  got ...%q...",
                id, at, len(want), len(got), want[lo:hi], got[min(lo, len(got)):ghi])
        }
    })
}
```

- **`corpus/manifest.json`'s `schema_version` moves to `"4"` in the same
  commit as the constant.** The two are compared by
  `TestManifestParsesAndMatchesSchemaVersion` (`corpus_test.go:162-174`)
  and cannot move apart. No other manifest field is touched.
- **The hand-authored valid fixture is committed; the invalid ones are
  built in `t.TempDir()`.** `testdata/extra-request-shape/` carries a
  two-response case whose single `Extra` entry uses the new shape, loaded
  by `LoadCase` in a test that asserts the decoded method, canonical
  `path_query` and response name. Every *failure* fixture stays
  `t.TempDir()`-built through `baseCase()`, which is the shape
  `cassette_test.go` already uses for its five negative tests, because a
  committed invalid case would confuse a reader walking `testdata`. Both
  fixtures use invented strings and neither is a corpus case.

**Named mutations this task reports:**

1. Drop the `Extra` loop's canonical-form check → a `t.TempDir()` case
   whose `Extra` entry stores `?inc=x&fmt=json` (unsorted) must fail
   naming both the stored and canonical forms; with the check removed it
   loads.
2. Drop the `Extra` loop's readability check → a case naming a
   nonexistent `responses/999.json` in `Extra` must fail.
3. Drop the filename exclusivity rule → the existing
   `TestLoadCaseRejectsExtraNamingConsumedResponse` (`:235-248`) must
   fail.
4. Drop the `Extra`-versus-`Requests` key rule → a case whose `Extra`
   entry repeats a `Request`'s exact `(Method, PathQuery)` while naming a
   different response file must fail.
5. Drop the `Extra`-versus-`Extra` key rule (fold m16) → a case whose two
   `Extra` entries share a `(Method, PathQuery)` while naming different
   response files must fail, and the message must name both indices.
6. Change `Case.Write`'s encoder to `json.Marshal` (no indent) → the new
   round-trip test must fail on all 215 cases, proving it compares bytes
   rather than values.
7. Add a non-nil `Extra` to a copy of one committed case's decoded value
   before writing → the round-trip test must fail on that case, proving
   the reshape is actually serialized.

**Deliverable count: 4** (the reshaped `Case.Extra` with its three
documented exclusivity rules and the four per-entry checks; the committed
`testdata/extra-request-shape/` fixture and the six `t.TempDir()`
failure proofs; the schema bump moved in place across the constant and
the manifest; the 215-case offline round-trip proof plus the updated
consumed-file test). Gate green plus tier-0 replay green; commit.
**Report actual token spend and the round-trip test's wall time at
`-count=1`.**

---

## Task 2: The stub half, `Consumed()` distinct from `Misses()`

Model: `sonnet`. Depends on Task 1 (the stub reads the new `Extra`
shape).

**Files:**
- Modify: `internal/harness/stub/stub.go` (package doc `:1-5`, `Server`
  `:40-49`, `New` `:51-77`, `ServeHTTP` `:79-91`, and the accessors
  `:93-105`)
- Modify: `internal/harness/stub/stub_test.go`
- Create: `internal/harness/stub/testdata/extra-channel/` (`case.json`
  plus `responses/001.json` and `responses/002.json`)

**Interfaces:**

- Consumes: `cassette.Case`, `cassette.Request` as Task 1 reshaped them.
- Produces:

```go
// internal/harness/stub/stub.go

// Server replays one [cassette.Case] over HTTP on an ephemeral loopback
// port. Close it when done; it does not stop itself.
//
// It answers from two maps. The request map is the case's committed
// log, which is an expectation: a request it does not cover is a
// [Miss], and a conformance run leaving a non-empty miss list fails.
// The extra map is the case's recall-extension fixtures
// ([cassette.Case.Extra], U8): answering from it is neither a miss nor a
// failure, and every such answer is reported through Consumed so a
// future candidate-widening improvement is graded on the shared subset
// rather than escaping the harness (rung 3 spec, the B2 fold). The
// request map is consulted first; [cassette.LoadCase] rejects a case
// whose two maps would share a key, so the order is a stated rule
// rather than a tiebreak.
type Server struct {
    listener  net.Listener
    http      *http.Server
    responses map[key][]byte
    extra     map[key]extraEntry

    mu       sync.Mutex
    misses   []Miss
    consumed []cassette.Request
}

// extraEntry is one recall-extension response and the log entry that
// names it, kept together so Consumed reports the entry rather than a
// reconstructed shape.
type extraEntry struct {
    request cassette.Request
    body    []byte
}

// Consumed returns every [cassette.Case.Extra] entry the server has
// answered so far, in the order it answered them. It is deliberately
// distinct from Misses: an Extra hit is a reportable event, never a
// failure.
func (s *Server) Consumed() []cassette.Request
```

**Behavior contract:**

- **`New` builds both maps and reads both bodies eagerly**, in `c.Requests`
  order then `c.Extra` order, with the same `os.ReadFile` and the same
  error shape the request loop already uses (`stub.go:56-62`). The error
  message distinguishes the two (`"stub: read extra response %q: %w"`) so
  a missing extra fixture is not reported as a missing request response.
- **`ServeHTTP` consults the request map, then the extra map, then
  records a miss.** An extra hit appends to `consumed` under the same
  mutex `recordMiss` uses, sets `Content-Type: application/json`, and
  writes the bytes. The 404 path is unchanged.
- **`New`'s signature does not change**, so the five existing consumers
  keep compiling untouched: `internal/harness/oracle/oracle.go:163`,
  `ties.go:65`, `album_info.go:57`, `cmd/harness/capture_test.go:544`,
  and `cmd/harness/capture.go`'s use of `stub.Miss` (fold M-g corrected
  the count from four). The task verifies that by building.
- **Compilation is not the whole check, and the derivation rule is
  stated** (fold M-g). Three of those consumers are `oracle.Derive`'s,
  which is the path pinned beets talks to while a committed record is
  produced, so a second replay map changes what that server answers: a
  request that previously drew a 404 and a recorded `Miss` would now draw
  a 200 with a body, which is a change to the oracle's own input and
  therefore to `make derive`'s byte-identity invariant. The rule that
  makes the change inert goes in `stub`'s package doc and in this
  contract: **an `Extra` entry is by construction a response the current
  flow does not request**, so the extra map cannot change a derivation,
  and a future case for which that is false has an entry belonging in
  `Requests` rather than in `Extra`. Today the rule is also arithmetic:
  zero of the 215 committed cases carry an `extra` key (C17), so the
  extra map is empty for every derivation this pass can run. Task 0 item
  9 re-measures that zero, and the day lane E1 populates the channel the
  derivation question is raised by this sentence rather than discovered
  in a byte diff.
- **`Consumed()` returns `slices.Clone(s.consumed)`** under the mutex,
  exactly as `Misses()` does (`:99-105`), so a caller cannot mutate the
  server's record.
- **The fixture is hand-authored and committed**, at
  `testdata/extra-channel/`: one `Request` naming a release lookup and
  `responses/001.json`, one `Extra` naming a search and
  `responses/002.json`, both bodies minimal invented JSON. It is not a
  corpus case, per the proximity precedent (spec `:320`). `stub_test.go`
  keeps building its other cases in `t.TempDir()` as it does today
  (`stub_test.go:17-39`); this is the package's first `testdata`
  directory and its reason is that the three-row table below needs one
  case whose two maps are both populated and stable across tests.

**The three-row table, stated exactly:**

| Request issued | `Misses()` | `Consumed()` |
| --- | --- | --- |
| the `Request` entry's `(GET, path_query)` | empty | empty |
| the `Extra` entry's `(GET, path_query)` | empty | one entry equal to the `Extra` row |

| a third `(GET, path_query)` neither map carries | one entry naming it | still one entry, from row 2 |

The table as a test, so the three rows are one function and no row can be
quietly dropped:

```go
func TestExtraChannelSeparatesConsumptionsFromMisses(t *testing.T) {
    c, err := cassette.LoadCase(filepath.Join("testdata", "extra-channel"))
    if err != nil {
        t.Fatalf("LoadCase(testdata/extra-channel): %v", err)
    }
    srv, err := stub.New(c, filepath.Join("testdata", "extra-channel"))
    if err != nil {
        t.Fatalf("New: %v", err)
    }
    defer srv.Close()

    tests := []struct {
        name         string
        pathQuery    string
        wantMisses   int
        wantConsumed int
        wantStatus   int
    }{
        {"answered by the request map", c.Requests[0].PathQuery, 0, 0, http.StatusOK},
        {"answered by the extra map", c.Extra[0].PathQuery, 0, 1, http.StatusOK},
        {"answered by neither", "/ws/2/release/00000000-0000-0000-0000-000000000000?fmt=json", 1, 1, http.StatusNotFound},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := http.Get("http://" + srv.Addr() + tt.pathQuery)
            if err != nil {
                t.Fatalf("GET %s: %v", tt.pathQuery, err)
            }
            resp.Body.Close()
            if resp.StatusCode != tt.wantStatus {
                t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
            }
            if got := len(srv.Misses()); got != tt.wantMisses {
                t.Errorf("cumulative Misses() = %d, want %d", got, tt.wantMisses)
            }
            if got := len(srv.Consumed()); got != tt.wantConsumed {
                t.Errorf("cumulative Consumed() = %d, want %d", got, tt.wantConsumed)
            }
        })
    }
    // len first, then the element: guarding the comparison on the
    // length it should be asserting lets an empty or two-entry
    // Consumed() pass silently (fold m5).
    if got := srv.Consumed(); len(got) != 1 || got[0] != c.Extra[0] {
        t.Errorf("Consumed() = %+v, want exactly the one Extra entry %+v", got, c.Extra[0])
    }
}
```

The counts are cumulative on purpose, and the prose table above says so
in row 3's own cell rather than leaving the reader to reconcile "empty"
against `wantConsumed: 1` (fold m15). The rows run in order against one
server, so a mutation that reports an extra hit as a miss fails row 2 on
both columns rather than on one.

**Named mutations this task reports:**

1. Delete the extra-map lookup from `ServeHTTP` → row 2 becomes a miss
   and both of its assertions fail.
2. `ServeHTTP` answers from the extra map without appending to
   `consumed` → row 2's `Consumed()` assertion fails while its `Misses()`
   assertion still passes, which is the mutation that proves the report is
   the thing being checked.
3. `New` loads `c.Extra` into the **request** map → row 2 is served but
   `Consumed()` stays empty and its assertion fails.
4. `New` skips `c.Extra` entirely → row 2 becomes a miss, the same
   failure as mutation 1 from a different cause; both are reported so the
   task does not claim one mutation proves two rules.

**Deliverable count: 3** (the second replay map with its documented
consult order and its distinct read error; `Consumed()` beside `Misses()`
with the clone discipline; the committed `testdata/extra-channel/`
fixture and the three-row table with its four firing proofs). Gate green
plus tier-0 replay green; commit. **Report actual token spend.**

---

## Task 3: The `matcher` inbound, and 3a's `dataSourcePenalty` carry closed

Model: `sonnet`. Depends on Task 0; independent of Tasks 1 and 2.

**Files:**
- Modify: `internal/engine/matcher/info.go` (`ItemFields` `:110-143`, the
  field-name constants `:147-166`, `ProjectItemFields` `:179-200`, and one
  new helper beside `stringField`)
- Modify: `internal/engine/matcher/info_test.go`
- Modify: `internal/engine/matcher/track_artist.go` (`vaArtists` `:18-21`,
  `addTrackArtist` `:33`)
- Modify: `internal/engine/matcher/track_artist_test.go`
- Modify: `internal/engine/matcher/penalty.go` (`dataSourcePenalty`
  `:404-409`, **and `AddDataSource`'s doc comment `:334-337`**, which
  after the change still promises callers the fallback is
  `Tunables.DataSourceMismatchPenalty`; fold M-a)
- Modify: `internal/engine/matcher/penalty_test.go`
- Modify: `internal/engine/matcher/tunables.go` (three doc comments:
  `SearchLimit` `:119-124`, `DataSourceMismatchPenalty` `:126-136`, and
  **`DataSourcePenalties` `:138-140`**, which after the change still says
  `AddDataSource` scans it "before falling back to
  DataSourceMismatchPenalty"; fold M-a. Plus, provisionally, the
  `DefaultTunables` literal `:282-285`.)
- Modify: `internal/harness/conform/component.go` (remove the two
  compensations at `:673` and `:721`, and reword `metadataSourceSet`'s doc
  comment `:564-569`)

**Interfaces:**

- Consumes: nothing new.
- Produces:

```go
// internal/engine/matcher/info.go, one new field on ItemFields and its
// name constant:

    // Comp is the beets comp flag. It is the one field here the
    // distance path never reads: va_likely's third disjunct is
    // any(item.comp for item in items) (autotag/match.py:454), which is
    // the SEARCH path, and internal/engine/musicbrainz's query builder
    // is its only consumer. It lives on this type rather than in a
    // second projection of the same item bag because two decoders over
    // one field bag drift silently.
    Comp bool

const fieldComp = "comp"

// boolField returns raw[key] as a bool, or false when key is absent or
// holds a non-bool value, reproducing beets' typed-default projection
// for a field whose declared type is bool (contract 10).
//
// It accepts a JSON bool only. beets stores comp as a types.Boolean
// integer flag, so a case carrying "comp": 1 would project false here
// and True in the shim; all 15 committed cases that carry comp carry it
// as JSON true, so nothing moves today, and this is the same shape
// intField's own comment records for the float64 arrival (info.go:
// 209-211).
func boolField(raw map[string]any, key string) bool
```

```go
// internal/engine/matcher/track_artist.go

// IsVariousArtistsName reports whether credit is one of VA_ARTISTS
// (autotag/distance.py:29), compared case-insensitively. It has two
// consumers: the track artist component, which declines when an item's
// own credit is one of these, and internal/engine/musicbrainz's
// va_likely, which treats the album's likely artist being one of these
// as evidence of a various-artists release (autotag/match.py:453). One
// list, two readers.
func IsVariousArtistsName(credit string) bool
```

```go
// internal/engine/matcher/penalty.go

// defaultDataSourceMismatchPenalty is the class-level fallback
// get_penalty resolves to when no loaded metadata source is named after,
// MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY
// (beets/metadata_plugins.py:180, returned at :160-169). It is a
// constant rather than Tunables.DataSourceMismatchPenalty because the
// two are different values in beets: the tunable is the PLUGIN's own
// penalty, resolved source_weight-first (:190-195) and wired into the
// plugin's config at init (:201-202), and it is what populates
// DataSourcePenalties. Reading the tunable here made the two identical,
// which is correct under shipped config and wrong the moment
// source_weight is set (3a's carry to 3c).
const defaultDataSourceMismatchPenalty = 0.5
```

**Behavior contract:**

- **`ItemFields`' own doc comment is amended, not left to contradict
  itself.** It says today that the eighteen fields are what the distance
  path reads and that "adding one is a finding, not a convenience"
  (`:110-113`). The amended comment declares nineteen fields: eighteen
  read by the distance path and `comp` read by the search path alone,
  with the citation and with the sentence's force preserved. An
  implementer that finds itself adding a twentieth stops and reports.
- **`ProjectItemFields` gains one line**, `Comp: boolField(raw,
  fieldComp)`, and `boolField` follows `stringField`'s exact shape
  (`info.go:202-207`): a type assertion whose failure yields the zero
  value, because a field carried under the wrong JSON type is treated the
  same as an absent one (contract 10). Its doc comment records the
  JSON-bool-only narrowing against beets' integer `comp` flag, the way
  `intField`'s own comment (`:209-211`) records the float64 arrival
  (fold m6).
- **`addTrackArtist` calls `IsVariousArtistsName(item.Artist)`** in place
  of its inline `slices.Contains(vaArtists, strings.ToLower(item.Artist))`
  (`track_artist.go:33`); `vaArtists` stays unexported and becomes the
  exported function's own data.
- **`dataSourcePenalty` returns the new constant** in its fallback branch
  (`penalty.go:408`), and **`AddDataSource`'s own doc comment `:334-337`
  is corrected in the same edit** to name the constant rather than the
  tunable (fold M-a). Leaving it would commit a doc comment that
  contradicts the function three lines below it, and neither vale nor
  golangci-lint reads a stale prose claim.
- **`Tunables.DataSourceMismatchPenalty` is given a reader**
  *(conductor ruling 2026-09-04: taken, fold M-a)*. Revision 1 said the
  field "populates `DataSourcePenalties["MusicBrainz"]` in
  `DefaultTunables`" and it does not; `:282-285` writes two independent
  literals, so after this task the field would have zero production
  readers and `golangci-lint`'s `unused` would not say so. Task 3
  therefore builds `DefaultTunables`'s map from the field,
  `DataSourcePenalties: map[string]float64{"MusicBrainz":
  t.DataSourceMismatchPenalty}` (the literal, because `matcher` carries
  no data-source name constant today and this pass adds none), which is
  what beets does (the per-source
  entry for MusicBrainz **is** the plugin's own `source_weight`-first
  resolved penalty, `$SP/beets/metadata_plugins.py:182-195`) and is
  behavior-neutral at 0.5. The field's doc comment then says that
  truthfully, and `DataSourcePenalties`' own comment `:138-140` is
  reworded to name the constant as the fallback. If the conductor rules
  the other way, the field is documented as a readerless config-load seam
  instead and the `DefaultTunables` literal is left alone; either ending
  keeps all 72 tier-0 rows green, which is the acceptance criterion
  below.
- **Both `conform` compensations are removed and tier 0 stays green.**
  `replayAddDataSource`'s `tun.DataSourceMismatchPenalty = 0.5`
  (`component.go:721`) and `replayTrackDistance`'s identical line
  (`:673`) exist only to compensate for the divergence this task closes,
  and `metadataSourceSet`'s doc comment (`:564-569`) is reworded to cite
  the new constant instead of describing a workaround. **All 72 tier-0
  rows replaying bit-identically afterward is a named acceptance criterion
  of this task**, not a background fact, and the reviewer checks the
  reported per-method counts against 3a's own record (`add` 8,
  `add_data_source` 4, `add_equality` 3, `add_expr` 2, `add_number` 4,
  `add_priority` 4, `add_ratio` 4, `add_string` 3, `generic_penalty_keys`
  1, `string_dist` 23, `track_distance` 16).
- **`Tunables.SearchLimit`'s doc comment is corrected in place** to name
  `musicbrainz.match.searchlimit` as the legacy spelling, with the
  citation chain C11 gives and one sentence recording that beets' own
  deprecation message spells it differently, which is where the wrong
  copies came from. The one document correction (`docs/HISTORY.md:600`)
  and the one optional spec clarification (`:226`, which is not wrong,
  fold m2) are Task 8's drafts for the conductor.

**Named mutations this task reports:**

1. `ProjectItemFields` drops the `Comp` line → a table row whose raw bag
   carries `"comp": true` must fail on the projected value.
2. `boolField` returns `true` for a non-bool value → a row carrying
   `"comp": "yes"` must project `false` and fires.
3. `IsVariousArtistsName` compares case-sensitively → a row with credit
   `"Unknown"` (the artist `nonva-va-artists-misflag` carries) must fail.
4. `dataSourcePenalty` reverted to `d.tun.DataSourceMismatchPenalty` →
   the hand-authored test setting both `DataSourceMismatchPenalty` and
   `DataSourcePenalties["MusicBrainz"]` to `0.7` and calling
   `AddDataSource("Discogs", "Bandcamp")` must fail with got `0.7`, want
   `0.5`.
5. `DefaultTunables` builds `DataSourcePenalties` from a literal `0.7`
   while `DataSourceMismatchPenalty` stays `0.5` → `tunables_test.go`'s
   default assertion must fail, proving the provisional wiring is really
   the map's source rather than a coincidence of two equal literals.
   *(Struck if the conductor rules the documentation-only ending.)*

**Acceptance evidence, distinct from the mutation list** (fold m7):
restoring either removed `conform` compensation must leave all 72 tier-0
rows green, in both directions. That is a neutrality check rather than a
firing proof, so it is reported here and not counted among the mutations,
keeping the assertion-discipline count honest.

**Deliverable count: 4** (`ItemFields.Comp` with `boolField`, the field
constant, and the amended type comment; `IsVariousArtistsName` with its
single consumer converted; the `defaultDataSourceMismatchPenalty`
constant closing 3a's carry, with both `conform` compensations removed
and 72/72 tier-0 rows re-verified; the four doc-comment corrections,
`SearchLimit`, `DataSourceMismatchPenalty`, `DataSourcePenalties` and
`AddDataSource`, plus the provisional `DefaultTunables` wiring). Gate green plus tier-0 replay green; commit. **Report
actual token spend and the tier-0 replay's per-method counts.**

---

## Task 4: `query.go`, the include set and the Lucene query

Model: **`opus`**. The first of two deliberate upshifts. This is the file
the whole conformance claim rests on: 351 committed requests are graded
byte for byte against what it builds, its escape set and its
double-lower-and-strip are exactly the shape that is plausibly wrong in a
way no reviewer notices, and its extra-tag branch is unreachable by any
corpus case and so is graded only by what this task hand-authors. Depends
on Task 3.

**Files:**
- Create: `internal/engine/musicbrainz/query.go` + `query_test.go`
- Create: `internal/engine/musicbrainz/testdata/query/` (the extra-tag
  and rendering fixtures)
- Modify: `internal/engine/musicbrainz/doc.go` (one sentence: the query
  construction is no longer 3c's future territory but this package's own)

**Interfaces:**

- Consumes: `matcher.Likelies`, `matcher.MostCommonTags`,
  `matcher.ItemFields`, `matcher.IsVariousArtistsName`,
  `matcher.Tunables` (`SearchLimit`, `ExtraTags`).
- Produces:

```go
// internal/engine/musicbrainz/query.go

// releaseIncludes is beets' RELEASE_INCLUDES
// (beetsplug/_utils/musicbrainz.py:50-67), in its own order, joined with
// a literal "+" at request time (:634-635).
//
// This is deliberately a SECOND definition of a string
// internal/harness/mbrequest already carries
// (mbrequest.go:22-24), and it must stay one. mbrequest is what the
// capture tool and the synth tool use to PRODUCE a committed request;
// the committed request log is the expectation and this list is the
// system under test. One shared definition would make
// internal/harness/conform's request grading a self-derived expectation,
// which is the anti-pattern the conformance run exists to catch. The two
// copies are independent on purpose and the conformance run is what
// proves they agree.
var releaseIncludes = []string{
    "artists", "media", "recordings", "release-groups", "labels",
    "artist-credits", "aliases", "recording-level-rels", "work-rels",
    "work-level-rels", "artist-rels", "isrcs", "url-rels",
    "release-rels", "genres", "tags",
}

// AlbumSearch is one album search request, built and not yet issued.
type AlbumSearch struct {
    // Query is the rendered Lucene query, escaped and wrapped. It is
    // NOT percent-encoded; the client encodes it once, at request time.
    Query string
    Limit int
}

// AlbumSearchFor builds the one album search request tag_album issues
// for items, mirroring get_album_criteria (beetsplug/musicbrainz.py:
// 714-734), get_search_query_with_filters (:736-753), format_search_term
// (beetsplug/_utils/musicbrainz.py:651-663) and search's own join
// (:679-681), in that order.
//
// It derives the search terms the way tag_album does rather than taking
// them: beets lets a caller pass explicit search_artist and search_name
// (autotag/match.py:443-446), and no rung 3 caller has any, so modelling
// the override now would be a dead parameter. 3d's flow or a later
// operator surface is where an explicit term would arrive.
func AlbumSearchFor(items []matcher.ItemFields, tun matcher.Tunables) AlbumSearch
```

Unexported, each with its own test:

```go
// vaLikely reports whether tag_album would treat items as a
// various-artists release: no artist consensus, OR the likely artist
// naming a various-artists placeholder, OR any item flagged comp
// (autotag/match.py:451-455). consensus is MostCommonTags' second
// return value; the likely artist is likelies.Artist, which
// MostCommonTags may already have overwritten with the album artist
// while leaving consensus["artist"] computed from the raw per-item
// values (util/__init__.py:842-844).
//
// It is unexported because AlbumSearchFor is this package's only caller
// and no consumer outside it exists at rung 3; 3d's flow exports it the
// day match_by_id's no-consensus branch needs it.
func vaLikely(likelies matcher.Likelies, consensus map[string]bool, items []matcher.ItemFields) bool
```

plus `albumCriteria(likelies matcher.Likelies, itemCount int, isVA bool,
tun matcher.Tunables) []criterion`; `type criterion struct{ Field, Value
string }`; `formatSearchTerm(field, term string) string`;
`escapeLucene(term string) string`; `lowerSearchTerm(s string) string`;
`trimSearchTerm(s string) string`; `variousArtistsID` (already at
`credit.go:11`, reused rather than redeclared).

**`vaLikely` is unexported and `albumCriteria` sheds two parameters**
(folds m19 and m25). Revision 1 exported `VALikely` while
`AlbumSearchFor(items, tun)` computes the likelies internally, so the
only caller in this pass sat inside the package, which is the dead
exported surface R6 rejects `Unused()` for. Revision 1 also gave
`albumCriteria` an `artist` and an `album` parameter that are
`likelies.Artist` and `likelies.Album` at every call site this pass
builds, because beets' explicit `search_artist` / `search_name` override
is deliberately not modelled (`$SP/beets/autotag/match.py:445-447`); six
parameters with two redundant is the shape go-conventions asks to
collapse, and 3d's explicit-search-term surface is where they come
back.

**Behavior contract:**

- **Criteria order is the query's term order and is fixed** (C8):
  `release` first, then `arid` when the release is various-artists else
  `artist`, then each configured extra tag. The extra-tag walk iterates
  `tun.ExtraTags` and looks up the MB key in the eight-entry
  `FIELDS_TO_MB_KEYS` map, **never the reverse**, because a Go map walk
  has no order and beets' does.
- **A tag repeated in `ExtraTags` emits one criterion, not two** (fold
  m17). beets builds a dict keyed by tag
  (`$SP/beetsplug/musicbrainz.py:697-712`), so `extra_tags: ["tracks",
  "tracks"]` yields one entry there and a naive slice walk yields two.
  The Go walk skips a tag it has already emitted, preserving first-seen
  order, and the extra-tag fixture table carries the duplicate row.
- **Each value is lowered and stripped twice** (C9 then C6), and the
  second pass is not elided as idempotent. Reproduce both, because a
  later beets change to one and not the other is exactly what the port
  must be able to see.
- **The escape runs before the field wrapper** (C6). `escapeLucene`
  prefixes each of C5's nineteen characters with one backslash; the
  wrapping `field:(` and `)` are added afterwards and are never escaped.
  The two committed shapes that prove both halves are
  `arid%3A%2889ad4ac3%5C-39f7...%29` (escaped hyphens inside unescaped
  parens) and `release%3A%28frank+zappa%5C%3A+cucamonga%29`.
- **An empty value drops the whole criterion**, at both filter points
  (C9's comprehension and C6's early return feeding C7's `filter(None,
  ...)`).
- **The six extra tags with a plurality value read `matcher.Likelies`**
  rather than recomputing one. beets computes them with `util.plurality`
  over the same items (`$SP/beetsplug/musicbrainz.py:727`,
  `$SP/beets/util/__init__.py:799-807`) that `MostCommonTags` already
  reproduces (`internal/engine/matcher/likelies.go:95-122`), so a second
  plurality pass would be a second implementation of one rule. The doc
  comment says so.
- **The three value special cases are reproduced, not fixed** (C8):
  `tracks` is `strconv.Itoa(len(items))`; `alias` is the album name;
  `catalognum` has every space removed; and `year` is
  `strconv.Itoa(likelies.Year)`, so an unset year yields the string `"0"`
  which survives the empty-value filter and ships as `date:(0)`. That
  last one gets its own fixture row and its own comment, because it looks
  like a bug and is the behavior under port.
- **`lowerSearchTerm` and `trimSearchTerm` are named helpers with their
  divergences on them** (R14). `lowerSearchTerm` is `strings.ToLower`
  with a doc comment citing
  `internal/engine/matcher/testdata/case-mapping-sweep.tsv` and recording
  that the query path has no post-transliteration strip, so all 56
  divergent code points reach the wire. The comment names the two
  phenomena separately (fold m11): U+0130 is the full-versus-simple
  case-mapping difference and will not move, and the other 55 are a
  Unicode version skew (Go 1.27's 17.0.0 tables against CPython 3.13's
  15.1.0) that moves the day either toolchain moves, which is why the
  asserted fixture row is the stable member. `trimSearchTerm` is `strings.TrimFunc` over `unicode.IsSpace(r) ||
  (r >= 0x1c && r <= 0x1f)`, closing the one divergence that is one line
  wide.
- **`AlbumSearch.Limit` is `tun.SearchLimit`**, default 5 (C11), and
  nothing in this file reads a config wall.
- **`doc.go`'s one sentence changes.** Its second paragraph says "the
  client, the Lucene query construction, and the rate limiter are 3c's
  territory" (`doc.go:3-5`); the query construction now lives here, so
  the sentence narrows to the two files Tasks 5 and 6 land, and the
  browse-fallback refusal paragraph (`:7-16`) is untouched.

**Fixture tables, each hand-authored in `testdata/query/` or inline:**

| Table | Rows |
| --- | --- |
| `escapeLucene` | one row per C5 character (19), plus one row per notably-absent character (`'`, `,`, `.`, `;`, `<`, `>`, `=`, `#`, `%`, `@`, `$`) asserting it is untouched |
| `formatSearchTerm` | empty term returns `""`; whitespace-only term returns `""`; empty field returns the bare escaped term; the wrapper's own `:` and parens survive unescaped |
| `trimSearchTerm` | the four C0 code points U+001C to U+001F, plus U+0085 and U+00A0, plus ordinary spaces |
| `lowerSearchTerm` | `İ` (U+0130) lowers to a single `i` here and to `i`+U+0307 in Python: a **recorded divergence**, asserted so it cannot change unnoticed |
| `vaLikely` | the three disjuncts one at a time and the all-false row: no artist consensus; `likelies.Artist` = `"Unknown"`; one item with `Comp: true`; and a row where none holds |
| `albumCriteria`, corpus-reachable | `release`+`artist`; `release`+`arid` under `vaLikely` |
| `albumCriteria`, the **tenth declared gap** | `extra_tags: ["tracks"]` → `tracks:(3)`; `["alias"]` → `alias:(<album>)`; `["catalognum"]` on `"ABC 123"` → `catno:(abc123)`; `["year"]` with no year → `date:(0)`; `["media","label"]` preserving the operator's order and differing from sorted MB-key order; `["tracks","tracks"]` emitting one criterion (fold m17); one tag outside `FIELDS_TO_MB_KEYS` dropped |
| `AlbumSearchFor`, end to end | the two committed shapes rendered exactly, compared against literals lifted from the tree's own `path_query` values after percent-decoding |

The end-to-end table as a test. Each `want` string is hard-coded rather
than derived from the code under test; the `release:` clauses are
percent-decoded from the tree's own committed `path_query` values and the
artist clauses are the fixture's own, and the implementer verifies each
literal against a decoded committed value before committing:

```go
func TestAlbumSearchForRendersTheCommittedShapes(t *testing.T) {
    tests := []struct {
        name  string
        items []matcher.ItemFields
        want  string
    }{
        {
            name: "release plus artist",
            items: []matcher.ItemFields{
                {Artist: "Vienna Philharmonic", Album: `Symphony No. 9, "Choral"`},
            },
            want: `release:(symphony no. 9, \"choral\") artist:(vienna philharmonic)`,
        },
        {
            name: "release plus arid, va_likely through the comp flag",
            items: []matcher.ItemFields{
                {Artist: "Frank Zappa", Album: "Frank Zappa: Cucamonga", Comp: true},
            },
            want: `release:(frank zappa\: cucamonga) arid:(89ad4ac3\-39f7\-470e\-963a\-56509c546377)`,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := AlbumSearchFor(tt.items, matcher.DefaultTunables())
            if got.Query != tt.want {
                t.Errorf("Query =\n  %q\nwant\n  %q", got.Query, tt.want)
            }
            if got.Limit != 5 {
                t.Errorf("Limit = %d, want 5 (search_limit's shipped default)", got.Limit)
            }
        })
    }
}
```

**Named mutations this task reports:**

1. Drop `genres` from `releaseIncludes` → the include-string test fails
   (and Task 7's 303 lookups miss).
2. Remove `-` from the escape set → the `arid` rendering row fails (and
   Task 7's **6** searches miss across 6 cases: the five `arid` rows
   through the various-artists MBID plus `real-noise-dylan-bootleg`'s own
   `release:` clause; fold M-f corrected the count from 5).
3. `strings.ToUpper` in place of `lowerSearchTerm` → every rendering row
   fails.
4. Reverse the criteria order (artist clause before release) → both
   end-to-end rows fail.
5. Apply `escapeLucene` after the field wrapper instead of before → the
   `frank zappa\: cucamonga` row fails, because the wrapper's own colon
   and parens would be escaped.
6. Drop the `any(item.comp)` disjunct from `vaLikely` → the `Comp: true`
   row fails.
7. `trimSearchTerm` reverted to `strings.TrimSpace` → the U+001C row
   fails.
8. Iterate `FIELDS_TO_MB_KEYS` **in sorted key order**
   (`slices.Sorted(maps.Keys(...))`) instead of `tun.ExtraTags` → the
   two-tag order row fails on every run, deterministically, because the
   fixture's `ExtraTags` is `["media","label"]`, which renders
   `format:(...) label:(...)` while sorted MB keys give `catno`,
   `country`, `date`, `format` and so on. **The mutant is sorted rather
   than bare** (fold M-h): a bare map range coincides with the expected
   order with probability `1/n!`, which falls fast and never reaches
   zero, so revision 1's "must fail on every run, enlarge the fixture
   until it is deterministic" asked for a proof that cannot exist and
   invited a false claim of determinism.
9. Emit a criterion per `ExtraTags` element without the de-duplication
   skip → the `["tracks","tracks"]` row fails with two `tracks:` terms
   (fold m17).

**Deliverable count: 4** (`releaseIncludes` with its written independence
rationale; `vaLikely`; `albumCriteria` with the eight-key extra-tag branch
and its three value special cases, graded as the tenth declared
corpus-coverage gap; the rendering layer, escape and lower and trim and
wrap and join, with its recorded case-mapping divergence and its closed
strip divergence). Gate green plus tier-0 replay green; commit. **Report
actual token spend; this is the figure the 6M ceiling is re-derived
from.**

---

## Task 5: `limiter.go`, the three-class limiter and the per-caller budget

Model: **`opus`**. The second and last deliberate upshift. This is the
pass's only concurrent code, its correctness properties (precedence
without starvation, a burst of one, context cancellation, a cap that
actually caps) are the kind a plan can state but not fully specify, and
it is the first file under `internal/engine` with a legitimate reason to
read a clock, which R2 answers by taking none. Depends on Task 0;
independent of Tasks 1 through 4.

**Files:**
- Create: `internal/engine/musicbrainz/limiter.go` + `limiter_test.go`

**Interfaces:**

- Consumes: nothing from this repo.
- Produces:

```go
// internal/engine/musicbrainz/limiter.go

// Class ranks one caller's claim on the shared MusicBrainz wire budget.
// Three classes are enough for every consumer the specs name: an
// uploader probe holds a contributor on the page through its lookups
// and must outrank a batch import (organizer design, "a user-present
// probe outranks a batch import"); interactive work must not starve
// timer imports (build spec, the same section); and rung 4's shadow
// spends deferred background budget through this limiter and never
// inline with the probe. Rung 9 re-verifies the policy against its real
// callers.
type Class int

const (
    ClassUserPresent Class = iota // a person is waiting on this request
    ClassBatch                    // an operator- or timer-triggered import
    ClassBackground               // deferred work: rung 4's shadow, and later
)

// reservationInterval is the anti-starvation floor: every
// reservationInterval-th served token is reserved for the non-empty
// classes BELOW the one strict priority would have served, taken in
// round-robin so consecutive reservations rotate through them. Strict
// priority alone satisfies the precedence requirement and violates the
// anti-starvation one; sending every reservation to the LOWEST non-empty
// class instead satisfies neither, because a saturated ClassUserPresent
// queue plus a queued ClassBackground waiter starves ClassBatch
// outright, which is the timer import the build spec's constraint
// protects; and a weighted fair queue makes a probe wait its weighted
// turn behind a batch, which is the latency the uploader journey cannot
// absorb. This one integer is the whole difference, and it is what rung
// 9 re-verifies rather than a whole policy.
const reservationInterval = 4

// ErrBudgetExhausted is returned by [Budget.Wait] when the budget's own
// limit is spent. A caller branches on it to stop asking rather than to
// retry, which is why it is a sentinel.
var ErrBudgetExhausted = errors.New("caller budget exhausted")

// ErrLimiterClosed is returned to every waiter when the limiter shuts
// down. A caller branches on it to distinguish process shutdown from
// its own budget running out.
var ErrLimiterClosed = errors.New("limiter closed")

// Limiter paces every MusicBrainz request one process issues, across
// every caller. It owns a goroutine, which is the one place in this
// package that is justified: the limiter's entire job is to make
// concurrent callers wait in a defined order, and rung 7's probe,
// rung 9's import and rung 4's shadow are three real concurrent
// callers.
//
// It reads no clock. Pacing comes from a ticker it owns, and a token is
// "the next tick", never "now minus the last request". That keeps
// internal/harness/conform's no-time.Now invariant literal rather than
// exempted, and it makes the tests deterministic: they drive the tick
// channel by hand.
type Limiter struct { /* ticker, done, waiters, three queues, spare, reservation counter */ }

// DefaultInterval is the minimum spacing between two MusicBrainz
// requests, one second. beets' musicbrainz.org branch pins rate_limit =
// 1.0 (beetsplug/_utils/musicbrainz.py:597-599) and builds its session
// as LimiterTimeoutSession(per_second=1.0) with burst 1, which binds
// before the 0.25s RateLimitAdapter floor (contract C20). The repo's
// own live spender holds the same line
// (cmd/harness/capture.go:27-31).
const DefaultInterval = time.Second

// NewLimiter returns a limiter pacing the wire at one request per
// interval and starts its dispatcher. Close it when the process shuts
// down. interval must be positive; DefaultInterval is the MusicBrainz
// value.
//
// It takes an interval rather than a requests-per-second integer
// because beets' own non-musicbrainz.org branch derives a float rate
// (ratelimit / ratelimit_interval, :600-606) and an integer forbids a
// deliberately conservative sub-1/s pace.
func NewLimiter(interval time.Duration) *Limiter

// Budget returns a named per-caller budget in class, spending at most
// limit requests; a limit of zero or less means no cap, and the budget
// is still paced and still counts what it spends. The name is what makes
// spend attributable in a log and is not otherwise interpreted.
//
// This is the "per-caller budget interface" the rung 3 spec names: a
// seam, not a Go interface. go-conventions holds an interface back until
// a second implementation exists, and one concrete type with an
// injected wait function serves both the metered and the unmetered case.
func (l *Limiter) Budget(name string, class Class, limit int) *Budget

// Close stops the dispatcher and fails every pending and future waiter
// with [ErrLimiterClosed]. It is idempotent.
func (l *Limiter) Close() error

// Budget is one named caller's claim on a [Limiter], or the unmetered
// stand-in [Unmetered] returns.
type Budget struct {
    name  string
    limit int
    wait  func(context.Context) error

    mu    sync.Mutex
    spent int
}

// Wait blocks until the budget may issue one request, returning
// ctx.Err() on cancellation, [ErrBudgetExhausted] when the limit is
// spent, and [ErrLimiterClosed] when the limiter shut down.
//
// It reserves its spend before it waits and releases the reservation on
// any error path out of the wait, so the limit caps concurrent callers
// as well as sequential ones. Checking spent against limit and only
// then waiting is a lost update: two goroutines sharing a budget with
// limit 1 both read spent == 0, both queue, and both are granted, and
// no -race run reports it because there is no data race, only a wrong
// decision.
func (b *Budget) Wait(ctx context.Context) error

// Spent reports how many requests this budget has been granted.
func (b *Budget) Spent() int

// Name reports the budget's name, for the caller that logs its own
// spend.
func (b *Budget) Name() string

// Unmetered returns a budget that never waits and never exhausts, for
// offline runs. internal/harness/conform uses it: the corpus holds 351
// committed requests and pacing them at one per second would add roughly
// six minutes to a gate that runs in 58 to 62 seconds today. It is a
// named constructor rather than a nil-limiter check, because a nil check
// at every call site is the defensive-nil pattern go-conventions
// rejects; the injected wait function is what makes the two cases one
// type with no branch.
func Unmetered() *Budget
```

**Behavior contract:**

- **The dispatcher is one goroutine reading three FIFO queues.** On each
  tick it serves the highest-priority non-empty queue, except that every
  `reservationInterval`-th served token is reserved for the non-empty
  classes strictly below the one that would otherwise be served, taken in
  round-robin so consecutive reservations rotate through them. The
  counter advances only on a served token, never on an idle tick, so an
  idle limiter does not accumulate a reservation debt. *(Provisional,
  conductor ruling owed, fold B1; R2 states the two alternatives and why
  revision 1's "lowest non-empty" rule starved `ClassBatch`.)*
- **A tick with no waiter is kept as one spare token**, and a waiter
  arriving while the spare is held proceeds immediately without waiting
  for the next tick. The spare is capped at one. This reproduces beets'
  `RateLimitAdapter`, which waits `rate_limit - elapsed` and therefore
  never delays a request following a long gap (C12), and it is
  `requests_ratelimiter`'s own `burst=1` restated without a clock.
- **`Wait` selects on the grant channel, `ctx.Done()`, and the limiter's
  `done` channel**, in one `select`, so cancellation and shutdown are
  never masked by an available token. A waiter that loses the race to
  `ctx.Done()` must not consume a token, and the test that proves it
  asserts the next waiter still gets the token that was in flight.
  **The mechanism is named rather than left to the implementer to
  invent** (fold m26): the grant channel is unbuffered, so a dispatcher
  send that no waiter receives never completes; the dispatcher therefore
  offers the grant in its own `select` against the waiter's departure
  and, when the offer is not taken, returns the token to the spare slot
  and drops the cancelled waiter from its queue. `opus` may ship a
  different shape with the same two observable properties, and it states
  which one it shipped so the reviewer has something to check the code
  against.
- **The budget reserves its spend before the wait and releases it on any
  error** (fold M-d), so an exhausted budget never occupies a queue slot
  and the cap holds under concurrent callers. `Wait` increments `spent`
  under `b.mu` and returns `ErrBudgetExhausted` when the incremented
  value exceeds `limit`; every error path out of the wait (cancellation,
  limiter close) decrements it again, so `Spent` still counts granted
  tokens only. Revision 1's "checked before the wait" was a check-then-act
  race on the one type the pass builds for concurrent use.
- **`Unmetered().wait` returns `ctx.Err()`** rather than nil, so an
  unmetered caller still honors a cancelled context. It increments
  `Spent` on success, so an offline run can still report what it issued.
- **The test seam is two unexported members**,
  `newLimiter(ticks <-chan time.Time) *Limiter`, which `NewLimiter` wraps
  with a real `time.NewTicker`, and `(*Limiter).awaitQueued(n int)`,
  which blocks until the dispatcher has registered `n` waiters. This is
  the repo's existing override convention (`doFetch`, `loadProfile`,
  `caseDir`). No exported test-only surface is added, and **neither
  `limiter.go` nor `limiter_test.go` calls `time.Now`** (fold m10:
  `checkNoTimeNow` skips only a `testdata` directory and parses every
  other `.go` file under `internal/engine`, `_test.go` included, so a
  timing assertion in the test file reds the gate for every lane).
- **No test sleeps, and a pre-`Wait` signal is not a registration
  proof** (fold B2). Every test drives ticks by hand and uses
  `context.WithCancel` for the cancellation rows. A test that needs to
  observe a waiter having queued calls `awaitQueued`, **never** a signal
  the waiter sends before calling `Wait`: between that send and the
  enqueue inside `Wait` there is an unbounded scheduling window, and a
  tick delivered inside it becomes the spare and is taken by whichever
  waiter registers first. `awaitQueued` is what makes the dispatcher's
  own state the thing the test waits on. The implementation is a counter
  the dispatcher advances on registration, under the mutex the queues
  already use, with a `sync.Cond` or a drained `registered` channel as
  the wakeup.
- **A hand-driven tick is delivered only after its waiter has
  registered.** The spare is capped at one, so a tick fed ahead of demand
  is discarded, and a test that feeds `n` ticks from an unsynchronized
  goroutine does not get `n` tokens. Every test in this task and in Task
  6 pairs each tick with an `awaitQueued` or drives it from inside the
  code under test.

The two properties that pull against each other, as one test, driven by
hand. **`awaitQueued` is what makes it deterministic** (fold B2):
registration is observed on the dispatcher's own state, never on a signal
the waiter sends before it enqueues, and no tick is fed until every
waiter this row needs is registered.

```go
func TestDispatcherPrefersUserPresentAndStillFeedsBackground(t *testing.T) {
    ticks := make(chan time.Time)
    l := newLimiter(ticks)
    defer l.Close()

    probe := l.Budget("probe", ClassUserPresent, 0)
    shadow := l.Budget("shadow", ClassBackground, 0)

    // served records the order Wait returned in, which is the property
    // under test.
    served := make(chan string, 16)
    start := func(b *Budget, label string) {
        go func() {
            if err := b.Wait(context.Background()); err != nil {
                served <- "error: " + err.Error()
                return
            }
            served <- label
        }()
    }

    start(shadow, "shadow")
    l.awaitQueued(1)
    for range 8 {
        start(probe, "probe")
    }
    l.awaitQueued(9)

    var order []string
    for range 8 {
        ticks <- time.Time{}
        order = append(order, <-served)
    }

    // Precedence: the first token goes to a probe queued after the
    // shadow waiter, not to the shadow.
    if order[0] != "probe" {
        t.Errorf("first token went to %q, want the user-present waiter", order[0])
    }
    // Anti-starvation: reservationInterval is 4, so the shadow waiter is
    // served on the fourth token and not left behind eight probes.
    if !slices.Contains(order, "shadow") {
        t.Errorf("background waiter never served across 8 tokens: %v", order)
    }
    if order[reservationInterval-1] != "shadow" {
        t.Errorf("token %d went to %q, want the reserved background waiter",
            reservationInterval, order[reservationInterval-1])
    }
}
```

**The three-class row, which revision 1 had no test for at all** (fold
B1). Two classes cannot see the starvation, because with only
`ClassUserPresent` and `ClassBackground` queued the two candidate rules
agree. The row queues one `ClassBatch` waiter and one `ClassBackground`
waiter behind a saturated probe stream, feeds `3 * reservationInterval`
ticks with `awaitQueued` before each batch of waiters, and asserts that
**every class is served at least once** and that the two reservations
land on different classes. Under revision 1's rule `ClassBatch` is served
zero times across every tick, which is what the assertion catches.

**Named mutations this task reports:**

1. Delete the reservation counter → the anti-starvation test (a steady
   stream of `ClassUserPresent` waiters plus one queued `ClassBackground`
   waiter, over twelve ticks) must fail because the background waiter is
   never served.
1b. Send every reserved token to the lowest-priority non-empty class,
   which is revision 1's rule → the three-class row must fail because
   `ClassBatch` is served zero times across `3 * reservationInterval`
   tokens (fold B1). This is the mutation that gives the middle class a
   firing proof of its own.
2. Serve the lowest-priority non-empty queue first → the precedence test
   (a `ClassBackground` waiter queued before a `ClassUserPresent` waiter,
   one tick) must fail.
3. Drop the spare token → the idle test (one tick with no waiter, then a
   waiter, then no further ticks) must fail by blocking.
4. Cap the spare at two instead of one → the burst test (two ticks with
   no waiter, then two waiters, then no further ticks) must fail because
   the second waiter is served.
5. `Wait` ignores `ctx.Done()` → the cancellation test must fail.
6. `Spent` not incremented on a grant → the cap test (a budget with
   `limit: 2` and three waits) must fail on the third wait returning nil
   instead of `ErrBudgetExhausted`.
6b. `Wait` checks the limit before the wait instead of reserving through
   it → the concurrent cap test (a budget with `limit: 2`, four waiters
   started together and released by four ticks) must fail with three or
   four successes and `Spent() > 2` (fold M-d). The row asserts exactly
   two successes, two `ErrBudgetExhausted`, and `Spent() == 2`.
7. `Unmetered()` routed through a dispatcher → the unmetered test (three
   waits with no limiter and no ticks) must fail by blocking.

**Deliverable count: 4** (the three-class dispatcher with its
round-robin reserved floor and its single spare token; the named `Budget`
with its reserve-then-release cap, its `Spent`, and the injected wait
function that makes metered and unmetered one type; `Unmetered` and
`Close` with their two sentinels; the tick-injection and `awaitQueued`
seams with the nine firing proofs). Gate green plus tier-0
replay green; commit. **Report actual token spend, and one
`go test -race -count=1 ./internal/engine/musicbrainz/...` run as
evidence.** `-race` is evidence here, not a new gate: whether it joins
`make check` is a decision for whoever owns the Makefile, and Task 8 files
it as a note.

---

## Task 6: `client.go`, the WS/2 client

Model: `sonnet`. Depends on Tasks 4 and 5.

**Files:**
- Create: `internal/engine/musicbrainz/client.go` + `client_test.go`
- Create: `internal/engine/musicbrainz/testdata/client/` (release and
  search response fixtures, hand-authored)
- Modify: `internal/engine/musicbrainz/release.go` (`wireInt`'s two
  decode errors, `:158-175`)
- Modify: `internal/engine/musicbrainz/doc.go` (the sentence Task 4
  narrowed, now naming nothing as 3c's future)

**Interfaces:**

- Consumes: `AlbumSearch` and `releaseIncludes` (Task 4); `Budget`,
  `Unmetered` and `DefaultInterval` (Task 5); `Release`,
  `ErrMalformedRelease`, `TranslateAlbum` (which owns validation, fold
  m8), `PseudoReleaseTarget`, `MergePseudoRelease`; `matcher.Tunables`,
  `matcher.AlbumInfo`.
- Produces:

```go
// internal/engine/musicbrainz/client.go

// DefaultBaseURL is the live MusicBrainz host beets hard-codes for the
// musicbrainz.org branch (beetsplug/_utils/musicbrainz.py:598-599).
const DefaultBaseURL = "https://musicbrainz.org"

// ErrNotFound reports that a lookup's release does not exist. beets'
// own flow branches on it: album_for_id catches HTTPNotFoundError and
// drops the candidate silently rather than failing the import
// (beetsplug/musicbrainz.py:777-781), so callers must be able to tell a
// deleted release from a real failure.
var ErrNotFound = errors.New("release not found")

// ErrUnavailable reports that every retry was spent on a retryable
// status or a transport error. It is the shape of beets' own
// RetryError, which is deliberately NOT an HTTPError and so is never
// mapped to a beets exception (beetsplug/_utils/requests.py:78-89).
// Rung 4's shadow branches on it to defer a lookup rather than record a
// divergence.
//
// No rate-limit-specific sentinel ships beside it: %w is an API
// commitment justified by a caller that branches, and no caller branches
// on 429 separately from 503 today.
var ErrUnavailable = errors.New("musicbrainz unavailable")

// ClientConfig configures a [Client]. UserAgent is required and the
// engine never builds one: MusicBrainz requires an identified client
// and the identity is site-specific, so it belongs to internal/config
// and reaches here through cmd/dubplate-server's composition
// (dubplate-server/<version> ( <contact> ), on cmd/harness/capture.go's
// own precedent, with the contact from Config.AdminEmail). Budget is
// required too; offline callers pass [Unmetered].
type ClientConfig struct {
    BaseURL   string        // empty means DefaultBaseURL
    UserAgent string        // required
    Budget    *Budget       // required
    Timeout   time.Duration // zero means 30s
    HTTP      *http.Client  // zero value means one built from Timeout
}

// NewClient validates cfg and returns a client. It fails rather than
// defaulting on a missing UserAgent or Budget, because both are
// correctness requirements rather than conveniences.
func NewClient(cfg ClientConfig) (*Client, error)

// Release fetches one release at beets' own maximal inc= set, mirroring
// get_release (beetsplug/_utils/musicbrainz.py:687-690). It returns
// ErrNotFound on a 404, ErrUnavailable when the retries are spent, and
// an error wrapping ErrMalformedRelease when the body does not decode.
// mbid is run through the same id extraction beets applies first
// (beets/util/id_extractors.py:20), so a malformed id costs no request
// at all, exactly as album_for_id costs none (beetsplug/musicbrainz.py:
// 771-773).
//
// It decodes and does not validate. TranslateAlbum owns validation and
// calls Release.Validate as its first act (translate.go:46-48), so a
// second call here would validate every body twice and split ownership
// of one rule across two files.
func (c *Client) Release(ctx context.Context, mbid string) (*Release, error)

// SearchReleases issues one release search and returns the release ids
// in response order, mirroring search (beetsplug/_utils/musicbrainz.py:
// 665-685) and candidates' own use of the result
// (beets/metadata_plugins.py:442-446). It returns ids rather than
// releases: the per-id fan-out and the candidate ordering are 3d's.
//
// A search body is an OBJECT, and the hits live under "releases":
// search ends by indexing the decoded body at f"{entity}s" (:684-685),
// and all 48 committed search responses carry exactly count, created,
// offset and releases (contract C19). Decoding into a bare array fails
// every committed search.
//
// Its errors are its own, and they do not mirror beets. beets swallows
// a search failure into an empty result inside _search_api's bare
// except (beets/metadata_plugins.py:419-427, which re-raises first when
// raise_on_error is set at :422-423); this client returns the failure
// instead, because swallowing it here would make Task 7's zero-miss
// rule unable to see a search that never happened. A 404 returns
// ErrNotFound, exhausted retries return ErrUnavailable, and a body that
// does not decode returns an error wrapping ErrMalformedRelease. The
// caller that wants beets' shape maps them to an empty candidate list,
// which is 3d's flow.
func (c *Client) SearchReleases(ctx context.Context, s AlbumSearch) ([]string, error)

// AlbumForID is album_for_id's whole body (beetsplug/musicbrainz.py:
// 765-800): fetch, translate, and, when the release is a pseudo-release
// with a backward transl-tracklisting relation, fetch the linked release
// and merge. Validation is TranslateAlbum's, called once there, never
// again here.
func (c *Client) AlbumForID(ctx context.Context, mbid string, tun matcher.Tunables) (*matcher.AlbumInfo, error)
```

Unexported: `extractReleaseID(s string) (string, bool)` with its compiled
pattern; `doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error)`;
`backoff(attempt int) time.Duration`; and the `sleep func(context.Context,
time.Duration) error` field the tests override.

**Behavior contract:**

- **Every request carries `fmt=json`** (C1), built through `url.Values`
  and `Encode()`, which reproduces the committed canonical form byte for
  byte with no adapter (C18). The path is `<base>/ws/2/release/<mbid>` for
  a lookup and `<base>/ws/2/release` for a search (C2), and the client
  appends `/ws/2` itself rather than requiring the caller to.
- **The lookup carries `inc`, the search carries none** (C3, C4).
- **The retry policy is C12's, with R13's three named divergences.**
  Seven attempts; retryable statuses 429, 500, 502, 503, 504; backoff
  zero before the first retry then `500ms * 2^(n-1)` capped at 120s with
  no jitter; **a delta-seconds `Retry-After` honored over the computed
  backoff, and an HTTP-date `Retry-After` falling through to it** (fold
  M-c: urllib3 accepts both forms, `$SP/urllib3/util/retry.py:328-347`,
  and the date form needs `time.Now`, which
  `TestEngineNeverReadsTheClock` forbids inside `internal/engine`);
  timeout 30s by default rather than beets' 10s; **and every attempt,
  retries included, acquires a limiter token before it is issued.** All
  three divergences carry their reason in the doc comment, as does the
  413-plus-`Retry-After` non-port (fold m3).
- **`client.go` and `client_test.go` call `time.Now` nowhere** (fold
  m10). The architectural walk covers both, the backoff schedule is
  observed through the injected `sleep` seam, and an elapsed-time
  assertion is the natural reach that reds the gate for every lane.
- **`extractReleaseID` is ported and its regex regime is recorded.** The
  pattern is `(\w{8}(?:-\w{4}){3}-\w{12})`; Python's `\w` is
  Unicode-aware and Go's RE2 `\w` is ASCII-only, which is the fourth
  member of the family 3b's guard 11 opened. No committed case carries a
  non-ASCII mb_albumid, so nothing moves today, and the divergence is
  named at the site.
- **`AlbumForID`'s error map is stated, including where it diverges.** A
  404 on the first lookup returns `ErrNotFound`, and the caller (3d) maps
  it to "no candidate", reproducing `album_for_id`'s `return None`. A 404
  on the **second**, pseudo-release lookup returns a wrapped error rather
  than a dropped candidate, because beets' second `get_release` sits
  outside the try and its `HTTPNotFoundError` escapes to
  `maybe_handle_plugin_error`, which logs and drops the plugin's whole
  candidate list (C13, C15). The shapes differ; the outcome for the caller
  is the same candidate loss, and the doc comment says which is which.
  Zero committed cases 404 on a pseudo second fetch.
- **An absent `status` key is a plain non-pseudo release, deliberately.**
  beets indexes `original_release["status"]` directly
  (`$SP/beetsplug/musicbrainz.py:785`), so a body with no `status` raises
  `KeyError` inside `album_for_id` and is swallowed, dropping the whole
  candidate list. `Release.Status` is a plain string, so an absent key
  decodes to `""` and `PseudoReleaseTarget` reports false. Measured: **all
  303 committed release-lookup bodies carry a top-level `status`** (the
  48 that do not are search responses), so the divergence is unreachable
  in the corpus. It is recorded as a deliberate choice rather than
  absorbed: reproducing an accident of direct indexing is not
  conformance, and the reader is told where the accident is.
- **`AlbumForID` closes 3b's `MergePseudoRelease` aliasing carry by
  construction**, and says so at the call site: both `AlbumInfo` values
  are built inside the call, only the merged one escapes, and the pseudo
  is dropped, so the shallow copy's aliased track slice
  (`docs/HISTORY.md:265-272`, `internal/engine/musicbrainz/pseudo.go:49-58`)
  has no second retaining reader. The hazard stays live for 3d and rung 9,
  so the carry changes owner rather than disappearing, and the one-line
  `slices.Clone` inside `MergePseudoRelease` is the fix the day a caller
  retains both.
- **`wireInt`'s two decode errors wrap `ErrMalformedRelease`**
  (`release.go:158-175`), closing 3b's `code-simplifier` leftover: the
  sentinel's own comment calls it "the one sentinel every boundary failure
  in this package wraps" (`release.go:10-13`), and 3c is the pass whose
  client makes that boundary real. `encoding/json` returns a custom
  unmarshaler's error unchanged for a non-`UnmarshalTypeError`, so
  `errors.Is` survives `json.Unmarshal`, and the test asserts exactly
  that rather than assuming it. **`wireInt` gains no second sibling and
  its census stands** (3b guard 16): `position` is the only field this
  payload spells two ways, 6,763 integers and 297 strings across the
  committed corpus, and the string form comes from the synth tool
  (`internal/harness/synth/release.go:128`) rather than from the wire, so
  this client meets real captures only and needs no other custom
  unmarshaler.
- **The transport is injected, and so is the wait.** `ClientConfig.HTTP`
  takes a `*http.Client` whose `Transport` a test replaces with a scripted
  `http.RoundTripper`; the `sleep` field is overridden so backoff tests
  observe the schedule without spending it. No test sleeps and no test
  reaches the network.
- **User-Agent conformance is a unit-test claim.** The stub never reads a
  request header (`internal/harness/stub/stub.go:81-91`), so the only
  proof is a scripted transport asserting the header on each attempt,
  retries included (R11).

The retry seam as a test, covering the schedule, the per-attempt token,
and the header on every attempt, without spending a second of wall time.
Its fixture table carries both `Retry-After` forms, the delta-seconds
row and the HTTP-date row, so the narrowing is a recorded passing
assertion rather than an unrecorded gap (fold M-c). Three test-local
helpers it uses are this task's to write beside it:
`roundTripperFunc`, a one-method `http.RoundTripper` adapter;
`releaseFixture`, the hand-authored release body under
`testdata/client/`; and `releaseFixtureID`, that body's own release id.

```go
func TestReleaseRetriesOnTransientStatusAndMetersEveryAttempt(t *testing.T) {
    var attempts int
    var agents []string
    rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
        attempts++
        agents = append(agents, req.Header.Get("User-Agent"))
        if attempts < 3 {
            return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))}, nil
        }
        return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(releaseFixture))}, nil
    })

    // One tick per observed attempt, driven from inside the transport
    // rather than fed ahead of demand: the spare is capped at one, so a
    // goroutine sending three ticks into an idle dispatcher has two of
    // them dropped and the call deadlocks on attempt 2 (fold B3).
    ticks := make(chan time.Time)
    l := newLimiter(ticks)
    defer l.Close()
    budget := l.Budget("test", ClassBatch, 0)
    go func() {
        for range 3 {
            l.awaitQueued(1)
            ticks <- time.Time{}
        }
    }()

    c, err := NewClient(ClientConfig{
        BaseURL:   "http://stub.invalid",
        UserAgent: "dubplate-test/0 ( test )",
        Budget:    budget,
        HTTP:      &http.Client{Transport: rt},
    })
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }
    var slept []time.Duration
    c.sleep = func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil }

    if _, err := c.Release(context.Background(), releaseFixtureID); err != nil {
        t.Fatalf("Release: %v", err)
    }
    if attempts != 3 {
        t.Errorf("attempts = %d, want 3", attempts)
    }
    if got := budget.Spent(); got != 3 {
        t.Errorf("budget.Spent() = %d, want 3: every retry is a real request against the wire budget", got)
    }
    if want := []time.Duration{0, time.Second}; !slices.Equal(slept, want) {
        t.Errorf("backoff schedule = %v, want %v", slept, want)
    }
    for i, ua := range agents {
        if ua != "dubplate-test/0 ( test )" {
            t.Errorf("attempt %d User-Agent = %q, want the configured one", i, ua)
        }
    }
}
```

**Named mutations this task reports:**

1. `NewClient` accepts an empty `UserAgent` → the constructor test fails.
2. `NewClient` accepts a nil `Budget` → the constructor test fails.
3. The retry loop acquires one token for the whole call instead of one
   per attempt → the scripted-transport test (503, 503, 200 against a
   hand-driven limiter) sees one token taken instead of three and fails.
4. Exhausted retries return `ErrNotFound` instead of `ErrUnavailable` →
   the sentinel test fails.
5. `backoff` computes `2^n` instead of `2^(n-1)` → the schedule test
   (0, 1, 2, 4, 8, 16 seconds) fails.
6. `Retry-After: 3` ignored → the delta-seconds header test fails.
6b. An HTTP-date `Retry-After` treated as seconds, or parsed at all →
   the date-form row fails, which asserts the computed backoff is used
   and that the client reads no clock (fold M-c).
7. `extractReleaseID` removed → the junk-id test, which asserts the
   scripted transport received **zero** requests, fails.
8. `wireInt`'s decode error not wrapping `ErrMalformedRelease` → the
   `errors.Is` test through `json.Unmarshal` fails.
9. `AlbumForID` skips the pseudo second fetch → the pseudo fixture's
   merged `Country` differs from the expected merged value and fails.

**Deliverable count: 6, and this task is the one that passes roughly
four** (fold m21; CLAUDE.md asks that it be said plainly rather than
counted down to fit). `ClientConfig` and `NewClient` with required
UserAgent and Budget and injected base URL; `Release` and
`SearchReleases` with `extractReleaseID`, the include set, `fmt=json`,
C19's search decode, and the decode boundary wrapping
`ErrMalformedRelease`; the retry, backoff, timeout and `Retry-After`
transport with `ErrNotFound` and `ErrUnavailable` and the per-attempt
token acquire; `AlbumForID`'s composition with the pseudo double fetch,
the merge, and its recorded divergences; `release.go`'s `wireInt`
sentinel wrap, which is a behavior change with its own mutation; and
`doc.go`'s sentence. The `wireInt` wrap could move to Task 3, and does
not, because Task 3's declared package territory excludes
`internal/engine/musicbrainz` and the wrap's whole justification is that
this task's client makes the decode boundary real. The conductor may
still split the task at the `client.go` / `release.go` line if Task 4's
report re-derives the ceiling upward. Gate green plus tier-0 replay green; commit.
**Report actual token spend.**

---

## Task 7: Request-construction grading in `internal/harness/conform`

Model: `sonnet`. Depends on Tasks 2, 4 and 6 (and, through them, 1, 3 and
5).

**Files:**
- Create: `internal/harness/conform/request.go` + `request_test.go`
- Modify: `internal/harness/conform/doc.go` (the package charter gains
  the request tier and the reason it drives a stub)

**Interfaces:**

- Consumes: `cassette.LoadCase`, `cassette.Case`, `cassette.Request`;
  `stub.New`, `(*stub.Server).Addr`, `Misses`, `Consumed`, `Close`;
  `musicbrainz.NewClient`, `ClientConfig`, `Unmetered`, `Release`,
  `SearchReleases`, `AlbumSearchFor`; `matcher.ProjectItemFields`,
  `matcher.DefaultTunables`; `conform`'s own `mergeProfile`
  (`translate.go:193-204`) and `releaseLookupPrefix` (`:20`), both reused
  rather than re-declared.
- Produces:

```go
// internal/harness/conform/request.go

// gradeUserAgent identifies the offline conformance client. The stub
// never reads a request header, so this value is a NewClient
// requirement rather than a graded fact; it is an invented,
// non-site-specific literal per CLAUDE.md's universality rule.
const gradeUserAgent = "dubplate-conform/0 ( offline-conformance-run )"

// Floors for the request census, measured off the tree at 2026-09-04.
// They are floors rather than exact counts because lane E1 adds cases:
// a shrink means a case or a request went missing, which is the failure
// worth catching, while growth is the corpus doing its job.
const (
    RequestCaseFloor   = 215
    RequestLookupFloor = 303
    RequestSearchFloor = 48
)

// RequestGradeSummary reports what one sweep actually compared, not what
// it visited.
type RequestGradeSummary struct {
    Cases         int
    Lookups       int
    Searches      int
    CasesNoSearch int
    ExtraConsumed []cassette.Request
}

// GradeRequests replays every committed case's request construction
// through the Go client against a stub built from that case's own log,
// and reports what it compared.
//
// It grades request CONSTRUCTION, never the request SET. The set is a
// function of the ID-match early return, _recommendation and
// _add_candidate, which are 3d's; this sweep reads each case's log to
// learn WHICH requests to issue and builds every issued request from the
// case's own items and this engine's own constants, so nothing about a
// request's bytes comes from the expectation it is graded against. 3d
// replaces the decision and inherits the stub, the zero-miss rule, and
// the client unchanged.
func GradeRequests(root string) (RequestGradeSummary, error)
```

**Behavior contract:**

- **Per case, in this order.** Load the case; start `stub.New` and close
  it before the next case so peak memory is one case's bodies rather than
  the corpus's (the largest single body is 8,894 KB and the corpus holds
  77.6 MB across 351 files); build a client with `BaseURL: "http://" +
  srv.Addr()`, `UserAgent: gradeUserAgent`, `Budget:
  musicbrainz.Unmetered()`; resolve the case's tunables through
  `mergeProfile(matcher.DefaultTunables(), c.Profile)`.
- **The lookup half.** Collect the distinct release MBIDs the log names,
  in log order, by cutting `releaseLookupPrefix` and then the `?`, which
  is the discriminator `internal/harness/analysis`'s own `releaseLookups` already documents (a
  search path is `/ws/2/release?` and never carries the trailing slash,
  `internal/harness/analysis/album_info.go:249-251`). Call
  `client.Release(ctx, mbid)` for each and require no error. Across the
  corpus that is 303 requests.
- **The search half.** If and only if the case's log carries a
  `/ws/2/release?` row, project the case's items through
  `matcher.ProjectItemFields`, build the query through
  `musicbrainz.AlbumSearchFor(items, tun)`, call
  `client.SearchReleases(ctx, s)`, and require no error. Across the
  corpus that is 48 requests over 48 cases; the other 167 assert nothing
  about whether they should have searched, because that is 3d's question.
- **The zero-miss rule reads the union of both stub maps**, which is what
  the second map is for (spec `:263`): a run leaving a non-empty
  `Misses()` fails naming every unmatched `PathQuery`, and a run that
  answered from `Extra` neither fails nor is silent.
- **`Consumed()` is reported, not asserted, and the plan says why.** Zero
  committed cases carry an `Extra` entry today (C17), so there is no
  expectation to assert against; the checker being proved here is the miss
  rule, and `Consumed()`'s own firing proof is Task 2's three-row table in
  `stub`'s testdata. `GradeRequests` accumulates the consumptions into the
  summary and the test logs them with `t.Logf`. That is a report and is
  named as one, not a vacuous assertion.
- **Two count assertions, at two scales.** Per case, the sweep asserts it
  issued exactly the number of lookups and searches the case's own log
  carries, which is the anti-silent-skip check and fires on the 25 cases
  carrying more than one lookup. Across the corpus, the three floors
  above catch a whole-corpus shrink. Together they are `gradeLedger`'s
  compared-count discipline (`internal/harness/conform/translate.go:318-324`)
  applied to requests.
- **`doc.go`'s charter gains the request tier.** The package doc says
  today that nothing here writes any harness package and that it reads
  the corpus read-only (`doc.go:20-22`); that stays true, and the new
  sentence records that this tier is the first place `conform` drives
  `stub.Server`, which `translate.go:301` already anticipated in writing
  ("the stub is 3c's business").

The gate-resident entry point, which is what a reviewer reads to check
that the sweep asserts anything at all:

```go
func TestRequestConstructionGradesCleanly(t *testing.T) {
    root := repoRoot(t)
    summary, err := GradeRequests(root)
    if err != nil {
        t.Fatalf("GradeRequests: %v", err)
    }
    if summary.Cases < RequestCaseFloor {
        t.Errorf("graded %d cases, floor is %d: a case went missing", summary.Cases, RequestCaseFloor)
    }
    if summary.Lookups < RequestLookupFloor {
        t.Errorf("issued %d release lookups, floor is %d", summary.Lookups, RequestLookupFloor)
    }
    if summary.Searches < RequestSearchFloor {
        t.Errorf("issued %d searches, floor is %d", summary.Searches, RequestSearchFloor)
    }
    // The identity holds only because tag_album issues at most one
    // search per case, which is true of every committed case today
    // (measured: 164 at 0 searches, 3 at 0, 26+4+9+9 at exactly 1). A
    // case carrying two searches is itself the finding, so the message
    // says so rather than reading as an accounting bug (fold m24).
    if summary.Cases != summary.Searches+summary.CasesNoSearch {
        t.Errorf("case accounting does not close: %d graded, %d with a search, %d without; a case logging two searches would break this identity and is itself the finding",
            summary.Cases, summary.Searches, summary.CasesNoSearch)
    }
    // Reported, not asserted: no committed case carries an Extra entry
    // today, so there is no expectation to compare against. Consumed()'s
    // own firing proof lives in internal/harness/stub's testdata.
    for _, r := range summary.ExtraConsumed {
        t.Logf("extra channel answered %s %s from %s", r.Method, r.PathQuery, r.Response)
    }
}
```

`GradeRequests` returns a joined error naming every failure rather than
the first, so one wrong include set reports all 303 misses in one run and
the mutation counts below are readable off the failure.

**Wall-time expectation, stated so the report can be graded against it.**
The sweep reads all 77.6 MB of committed bodies once per run and decodes
each of the 303 lookup responses into the full `musicbrainz.Release`,
plus 351 loopback round trips at well under a millisecond each. **That is
a heavier decode than 3b's browse-fallback guard does over the same
files** (fold m9): the guard decodes into a two-field anonymous struct
(`internal/harness/conform/translate.go:423-428`), so its 2.15 to 2.38s
reported and ≈3.4s on a reviewer's machine is a floor rather than the
same work. The full decode was measured directly at review time, all 303
bodies into `musicbrainz.Release` plus `Validate`, at 1.89s with zero
decode and zero validate errors, which is why the estimate survives.
**Expect 3 to 8 seconds at `-count=1`**, against
`go test ./...` at 8.6s and `bash scripts/check.sh` at 58s to 3m25s with
gitleaks dominant. If the measured figure exceeds roughly 30 seconds, the
task reports it as a finding and the conductor decides whether the sweep
becomes its own leg of `scripts/check.sh` at full scope, which the spec
already names as the escape (spec `:267`) and which is explicitly **not**
an optional target.

**Named mutations this task reports, each with its measured miss count:**

1. Drop `genres` from `releaseIncludes` → **all 303 lookups miss.**
2. Remove `-` from the Lucene escape set → **6 searches miss, over 6
   cases.** Five are the `arid` rows, whose escaped hyphens come from the
   various-artists MBID, and the sixth is
   `real-noise-dylan-bootleg`'s own `release:` clause,
   `release%3A%28bob+dylan+%5C-+the+bootleg%29`. Those 6 rows carry 5
   distinct `path_query` values, because the two Utada cases share one
   (fold M-f corrected the count from 5).
3. `strings.ToUpper` in place of `lowerSearchTerm` → **all 48 searches
   miss.**
4. Omit `limit` from the search parameters → **all 48 searches miss.**
5. Return only the first lookup MBID per case from the extractor → the
   per-case issued-count assertion fires on the 25 cases carrying more
   than one lookup, proving the anti-silent-skip check rather than the
   miss rule.

**Deliverable count: 4** (the per-case sweep with its lookup and search
halves and its one-stub-at-a-time discipline; the zero-miss rule over the
union of both maps plus the `Consumed()` report; the per-case
issued-count assertion and the three corpus floors; the five named
mutations with their measured miss counts, and the measured wall time at
`-count=1`). Gate green plus tier-0 replay green; commit. **Report actual
token spend, the sweep's wall time, and the per-half compared counts.**

---

## Task 8: The pass close

Model: `sonnet`. Depends on Tasks 0 through 7.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit)
- Draft, handed to the conductor in the task return: four
  `corpus/README.md` corrections; one `search_limit` spelling correction
  and one optional spec clarification; one rung 3 spec tree-column
  correction; two contested-surface rows; two filed notes

**Behavior contract:**

- **The HISTORY entry is drafted in the house shape**: what landed, what
  the gate caught, and what a later pass would be wrong to rediscover.
- **The request coverage record, stated as a table a reviewer can check.**
  Per contract C1 through C20: its Go site, what grades it, and, for a
  rule no corpus case can reach, which hand-authored fixture grades it and
  what mutation proved that fixture fires.
- **The declared corpus-coverage gaps are recorded as gaps, not as
  passing checks** (spec, M8). This pass declares **one**: the
  `extra_tags` criteria branch, including `tracks`, `alias`, the
  `catalognum` space-strip and the `date:(0)` trap, unreachable because no
  committed profile writes a `musicbrainz:` block. **Together with 3a's
  three and 3b's six, rung 3 now carries ten**, and the entry states the
  running total so 3d's close does not re-count. Three further surfaces
  are recorded as **non-ports** rather than gaps, each with its reason:
  the recording lookup and `RECORDING_INCLUDES`; the browse fallback;
  beets' explicit `search_artist`/`search_name` override; and urllib3's
  413-plus-`Retry-After` sixth retryable case.
- **The evidence, not a bare "green".** The request tier: cases, lookups,
  searches, compared counts, the zero-miss result, the `Consumed()` count,
  and the measured wall time at `-count=1`. The schema bump: the
  round-trip proof over all 215 cases and the confirmation that no
  committed byte moved. Tier 0: 72/72 rows, per method, still
  bit-identical after the `dataSourcePenalty` fix and the removal of both
  `conform` compensations. `go test ./...` and `bash scripts/check.sh`
  with broken-down wall time. The limiter's `-race` run.
- **Forward carries, each stated so 3d and rung 4 read them rather than
  derive them:**
  1. **The rung 15 re-run criterion**, quoted verbatim from the Spec
     section above. Every rung 3 pass close carries it.
  2. **The four-knob wall registration** stays the named follow-on task
     for the first rung 3 pass that runs after rung 1b closes, with the
     four knobs named, `internal/config/keys.go` named as the file no lane
     edits, and R16's new fact recorded: `Threshold.Default` is a string
     and `ParseInt` is the only parser, so the three float rec thresholds
     need a float variant the registry does not yet have.
  3. **3c grades request construction, not the request set**, and the
     seam is R1's: 3d owns the ID-match early return, `_recommendation`,
     `_add_candidate`, the explicit `search_ids` branch, and
     `match_by_id`'s no-id/no-consensus branch, and inherits the stub, the
     zero-miss rule and the client unchanged.
  4. **The stub's blind side and its four-line close** (R6): nothing
     reports which logged `Request` entries a run consumed, so a client
     issuing fewer requests than the log records leaves zero misses. 3c's
     per-case issued-count assertion covers it inside 3c; 3d needs the hit
     set.
  5. **`Case.Extra` has no producer** (R7), and a producer is
     `cmd/harness` work in lane E1's territory. The channel stays
     hand-authored until E1 builds one.
  6. **The engine's `/ws/2` constants are a deliberate second copy**
     (R9), and unifying them with `internal/harness/mbrequest` would turn
     the request grading into a self-derived expectation. 3b's "worth
     revisiting if a fourth appears" trigger is met and answered.
  7. **The client's three deliberate transport divergences** (R13): a
     30s timeout rather than beets' 10s; a limiter token acquired per
     attempt rather than per call; and `Retry-After` honored in its
     delta-seconds form only, the HTTP-date form falling through to the
     computed backoff because parsing it needs a clock read
     `TestEngineNeverReadsTheClock` forbids. Beside them, one non-port:
     urllib3's 413-plus-`Retry-After` sixth retryable case, unreachable
     from MusicBrainz. None is gradeable by the corpus.
  7b. **The wire pace is one request per second and it is now written
     down** (C20): beets' `musicbrainz.org` branch pins `rate_limit =
     1.0` and its session limiter binds before the 0.25s adapter floor,
     `musicbrainz.DefaultInterval` is `time.Second`, and rung 4's shadow
     is the first real spender that inherits it.
  7c. **The `AdminEmail` User-Agent contact is interim, not settled**
     (R11): `Config.AdminEmail`'s own doc comment says adopters compare
     it and never display it, and the harness precedent this plan cites
     uses a dedicated `DUBPLATE_HARNESS_CONTACT` input instead. A
     dedicated contact row is the post-1b follow-on, beside the four-knob
     registration.
  8. **The retry path is ungraded by the corpus and always will be**: the
     stub answers 200 or 404 only, `cassette.Request` carries no status
     field, and the response-status extension stays declined, now for the
     third pass running.
  9. **Three regime divergences now travel together** (R14): 3b's `\b`
     (Unicode-aware in Python `re`, ASCII-only in Go RE2), the query
     path's full-case-versus-simple-case lowering with all 56 sweep code
     points reaching the wire, and `_extract_id`'s `\w`. The strip
     divergence is closed rather than carried.
  10. **`AlbumForID` closes 3b's `MergePseudoRelease` aliasing carry by
      construction**, and the hazard changes owner to 3d and rung 9 rather
      than disappearing; the `slices.Clone` is the fix the day a caller
      retains both.
  11. **3a's `dataSourcePenalty` carry is CLOSED** (R15), with the two
      `conform` compensations removed and 72/72 tier-0 rows re-verified as
      the proof.
  12. **An absent release `status` is a plain non-pseudo release here and
      a dropped candidate in beets**, unreachable in the corpus (303 of
      303 lookups carry one), recorded as a deliberate choice.
  13. **The limiter's one knob is `reservationInterval`**, and rung 9
      re-verifies that integer against its real callers rather than
      re-deriving the policy. Rung 9 also inherits the reserved token's
      **round-robin over the classes below the served one**, which
      replaced revision 1's "lowest non-empty" rule because that rule
      starved `ClassBatch` outright whenever a probe stream and the
      shadow were both queued.
  13b. **`Budget`'s cap reserves before it waits**, so it holds under
      concurrent callers; a pre-wait check is a lost update that no
      `-race` run reports.
  13c. **`vaLikely` is unexported at rung 3** and 3d exports it the day
      `match_by_id`'s no-consensus branch needs it; the same rule kept
      `stub`'s `Unused()` accessor out of this pass.
  14. **The measured wall times**: the request tier, the round-trip proof,
      tier 0, `go test ./...`, and `scripts/check.sh`, which the
      version-bump differential doc will want, the same courtesy 2b passed
      to 2c, 2c to 2d, 2d to 3a, and 3a to 3b.
- **The drafted corrections**, each one line for the conductor's batch:
  - `corpus/README.md:40-46`, the Layout paragraph: `Case.Extra` now
    carries `Request` shape, not a list of response filenames.
  - `corpus/README.md:1959-1980`, the schema-version history: a third
    entry for `"3"`→`"4"`, recording that it forced no re-freeze and
    stating the mechanical reason (an `omitempty` field empty on every
    committed case) rather than extending 2d's categorical rule.
  - `corpus/README.md:2389`, which names `schema_version 3` beside the
    `mediafile` pin.
  - `corpus/README.md:2775-2785`, the recorded-fixed-point narrative:
    an orphaned response becomes an `Extra` **entry** carrying its own
    method and canonical `path_query`, and the channel still has no
    automatic producer.
  - `docs/HISTORY.md:600`, 3a's correction #5 (the sentence is on
    `:600`, not the `:597-599` revision 1 cited, fold m13): the
    `SearchLimit` legacy spelling is `musicbrainz.match.searchlimit`, and
    beets' own deprecation message at
    `$SP/beetsplug/musicbrainz.py:297` is why the wrong spelling
    propagated.
  - The rung 3 spec `:226`, the tunable enumeration: an **optional
    clarification**, never a correction (fold m2). That line spells
    legacy keys relative to the plugin namespace, as its adjacent
    `source_weight` item does, so `match.searchlimit` is right as it
    stands; spelling it `musicbrainz.match.searchlimit` removes an
    ambiguity and the conductor may decline it without leaving anything
    wrong on the page.
  - The parallel-approach spec's contested-surface table: the
    `corpus/manifest.json` row and the `corpus/README.md` row this plan's
    "Contested surfaces" section drafts (fold m22).
  - The rung 3 spec `:330`, the 3c tree column: this pass also writes
    `internal/harness/{cassette,corpus_test.go}`, `corpus/manifest.json`,
    `internal/engine/musicbrainz` and three named `internal/engine/matcher`
    additions, each justified in this plan's Global Constraints. An
    uncorrected spec is the next pass's stale input, which is exactly why
    3b drafted its own m14 correction.
- **Two notes filed, neither this pass's work:** folding
  `cmd/harness/capture.go:279`'s and `internal/harness/synth/synth.go:41`'s
  inline `"/ws/2/release/"` literals into `mbrequest`'s own constant, for
  the lane that owns `mbrequest` (R9); and whether `-race` joins
  `make check`, for whoever owns the Makefile (Task 5).
- **STATUS stays present tense and under 60 lines**; anything historical
  moves to HISTORY. The next action points at 3d (3d needs 3a, 3b and 3c,
  all now shipped), and the lane registry entry is updated.
- **Both budgets are scored**: token spend against the re-derived ceiling,
  and attended interactions against zero, which is this pass's target.

**Deliverable count: 4** (the HISTORY draft in the house shape with its
eighteen forward carries; the request coverage table with the one new
declared gap, the four non-ports, and the running total of ten; the
evidence record with broken-down wall times, the round-trip result, the
72/72 tier-0 re-verification and the `-race` run; the STATUS rewrite with
the next action, plus the nine drafted corrections and rows and the two
filed notes). **The task also records which of the five conductor rulings
this pass applied provisionally were taken as recommended and which were
overruled**, in the same shape 3b's close used. Gate green plus tier-0 replay green; commit.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain per task,
**strictly sequential**. Nothing here runs in parallel, and if a second
code slot opens at all it belongs to lane A or lane T in its own
worktree, never to a second executor in this tree. `docs/STATUS.md`'s
lane registry records lane-2 as BLOCKED until Phase 0b and 0c land, so
the expected state through this pass is one code lane. Every task runs
the repo's Sonnet implementer (`cairn-implementer`, which commits its own
task's specific files once `bash scripts/check.sh` is green) →
`diff-reviewer` → gate, with one re-dispatch on `fix` and a second `fix`
as the conductor's decision. The conductor dispatches
`code-simplifier` over each accepted task's changed Go after review and
before the pass close, and owns every ledger edit, the lane registry, the
contested-surface row, the seven drafted corrections, and the batch
commit. **Runaway guard armed at launch**, per the standing requirement.
**Subagents never dispatch agents.**

**Models.** `sonnet` by default. **Two named upshifts to `opus`:**

- **Task 4**, `query.go`. It is the canonical request construction, and
  351 committed requests are graded byte for byte against what it builds.
  Its escape set, its double lower-and-strip, and its criteria ordering
  are each independently easy to get plausibly wrong in a way that passes
  a casual read, and its extra-tag branch is unreachable by any corpus
  case, so the only thing standing between a wrong rule and a green gate
  is the quality of the fixtures this task hand-authors.
- **Task 5**, `limiter.go`. It is the pass's only concurrent code; its
  five correctness properties (precedence, the anti-starvation floor that
  reaches all three classes, the single spare token, a cap that holds
  under concurrent callers, and cancellation that does not swallow a
  token) are the kind a plan states and cannot fully specify; and it is
  the first file under `internal/engine` with a legitimate reason to read
  a clock, where the wrong answer amends a landed architectural check.
  **Three of this pass's three blockers landed in this file's design and
  its tests** (B1, B2, and B3's tick discipline), which is the upshift
  earning itself before the dispatch.

**Task 6 is the runner-up**: the conductor upshifts it if Task 4's report
says a rendering contract above was under-specified, because the client
is where a query becomes bytes and the retry seam is the one part of it
no corpus fixture can grade.

**Review.** Rung 3 is not one of the three full adversarial-cycle rungs
(5, 12, 16), so review scales to the in-chain `diff-reviewer` per task
plus the pass-end gate. **This pass does not take a fresh-context review
of its own instrument**, and the reason is recorded rather than
defaulted: the spec raises 3d's conformance runner as the one candidate
exception (spec `:362`), and Task 7's request tier is materially cheaper
to trust than 3d's will be. It grades 351 requests against bytes pinned
beets actually sent, its per-case issued-count assertion and its three
floors make a silent skip fail rather than pass, and every one of its five
mutations has a measured miss count rather than a claim. The conductor may
still take one.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- **Task 1 → Task 2.** The stub reads the reshaped `Case.Extra`. Task 1
  is also the conductor-ruled first task of the pass (spec `:45-50`).
- **Task 3 → Task 4.** The query cannot be built without
  `ItemFields.Comp` and `matcher.IsVariousArtistsName`.
- **Task 4 → Task 6.** The client issues what the query builder builds.
- **Task 5 → Task 6.** `ClientConfig.Budget` is Task 5's type, and the
  per-attempt token acquire is tested against Task 5's hand-driven
  limiter.
- **Tasks 2, 4, 6 → Task 7.** The sweep needs the stub's second map, the
  query builder and the client together.
- **Tasks 0–7 → Task 8.**
- Tasks 1 and 2 are independent of Task 3, and Task 5 is independent of
  Tasks 1 through 4. **They are still run sequentially**, because one
  tree runs one executor and the module gate runs over `./...` regardless.

**Where this pass splits, if it splits.** Eight tasks plus a pre-task is
past the roughly-eight bar, so the split point is named rather than the
scope packed. **The split is at the Task 2 / Task 3 boundary**: Tasks 0–2
are the harness half (`internal/harness/{cassette,stub}`,
`corpus_test.go`, `corpus/manifest.json`) and write no
`internal/engine` at all; Tasks 3–7 are the engine half plus its grading.

**The recommendation is to keep one pass.** Three reasons, in order of
weight. The harness half is the pass's cheapest work by a wide margin: an
empty migration, one extra map, one accessor, and two committed fixtures,
which on 3b's measured per-task rate is roughly 500K against a second
staleness pre-task, a second close, and a second full-gate cycle. The
`Extra` channel shipped alone would be a channel with no producer, no
committed consumer, and no run pointing at it, which is the "instrument
nobody points at anything" shape the spec's risk 2 names. And 3d is
blocked on all of 3a, 3b and 3c (spec `:333`), so a 3c-i and a 3c-ii
would serialize the critical path for no throughput gain while lane-2
stays BLOCKED and nothing else can take the freed slot. **Take the split
only if Task 4's actual spend re-derives the ceiling past 6M.** The
conductor takes the call at the Task 4 checkpoint, on the number rather
than on a feeling.

**Budget:** 6M starting ceiling, re-derived after Task 4; phase 1's
combined ceiling 12M as a reservation; flag at 4.8M; STATUS after every
task. Spend is scored at close against the re-derived ceiling and against
**zero attended interactions**, which is this pass's target.

---
