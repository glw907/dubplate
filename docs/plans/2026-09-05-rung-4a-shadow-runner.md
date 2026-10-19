# Rung 4, Pass 4a (The Shadow Runner) Implementation Plan

**Revision 2.2, 2026-09-07 (owner rulings ratified; 4a-ii ceiling):** Geoff accepted all six of the rung 4 spec's open rulings as recommended, so every "provisional, owner ruling N open" mark below reads as ruled and 4a-ii carries no reversal fold. Pass 4a-ii (Tasks 10 through 15) runs under its own 3M ceiling, set on 4a-i's measured rate (about 230K per task, one Opus upshift candidate, simplify, review, and a full close), with the 80% flag at 2.4M routing to Geoff per CLAUDE.md's standing rule now that the 2026-09-05 window's grant has lapsed. 4a-i's spend of about 4.99M against the original 5M is recorded in HISTORY's 4a-i entry.

**Revision 2.1, 2026-09-05 (conductor rulings on the fold, all three taken as recommended, and the pass split):** C1, the capture's grouping key is beets' own, `(cmp.Or(AlbumArtist, Artist), Album)`, and the rung 4 spec's grouping sentence is amended to match; C2, `newShadowCmd` takes `matcherCommit string` as a parameter and `newRootCmd` passes the package var; C3, the `internal/engine/shadow` never imports `internal/config` rule stays a declared gap guarded by the in-chain diff-reviewer. **The pass is split at the Task 9 / Task 10 boundary**, on CLAUDE.md's pass-sizing rule rather than on this plan's Task 4 trigger: fifteen tasks is past the roughly-eight bar, and the 2026-09-05 unattended window's 15M cap binds before the pass ceiling does (about 10M spent at this note, prep 2.3M of it). **Pass 4a-i is Tasks 0 through 9** (the engine half: atomicfile, tunables, config, the FLAC reader, the store, capture, reconcile, transports, the re-score and classification), run now in the main tree with a 3M sub-ceiling and an intermediate close carrying a written handoff. **Pass 4a-ii is Tasks 10 through 15** (the ledger fields, the marshalling, the verbs, `reevaluate`, the exit gate, the close), run by the next session against 4a-i's handoff. Every "provisional, owner ruling N open" mark stays; every "provisional, conductor ruling owed" mark reads as ruled.

**Revision 2, 2026-09-05:** folds the two adversarial plan reviews,
merged in `research/2026-09-05-rung-4a-plan-review.md`: blockers B1 to
B7, majors M-a through M-g, minors m1 to m18. Three reviewer evidence
claims are corrected with the measurement in that document's header, and
none of the three retires its finding. **Three items are conductor-owed
rather than folded silently**, each applied provisionally here and
marked inline: the capture's grouping key, `newShadowCmd`'s stamp
parameter, and whether an architectural walk guards the
`internal/config` import rule.

The load-bearing corrections. **Task 4's STREAMINFO layout named the
wrong eight bytes** (B1): the sample-rate and total-sample pack sits at
payload offsets 10 through 17, and the payload's last 64 bits are the
tail of the MD5 signature, so a reader built to revision 1's sentence
would have divided one half of a digest by the other and moved every
live verdict through the track-length penalty. The task now states the
layout by offset and cross-checks it against the committed 42-byte
template (M-a), because one task authoring both the builder and the
parser can agree with itself on a wrong format forever. **Task 10's
byte-for-byte ledger re-marshal cannot pass** (B2), for two independent
reasons the fold states rather than patches: the committed ledger orders
`corrected_verdict` after `date` while the Go struct declares it before
`reason`, and the file carries characters Go's encoder escapes by
default. **The match step iterated in observation-id order over a digest
that ingests modification times** (B3), so the exit gate's
numeric-prefix fixture could not control which album the budget cap
starved and the determinism test could not pass; the pass loop now
matches in capture order and every fixture pins its modification times.
**The capture's grouping key dropped beets' own `albumartist or artist`
fallback** (B4), so a flat contributor directory with no `ALBUMARTIST`,
the exact material the rule exists to exclude, would have been observed
as one album and scored against one of several run-report rows.

Three more blockers move real behavior. `Observation` had nowhere to put
a reconcile-time exclusion, and the exit gate asserts on twelve stored
records one of which is exactly that shape (B5). Task 1's crash-safety
test asserted on a file the failing write never touches, which is a
vacuous assertion dressed as the store's crash-safety proof (B6). And
Task 12's stamp-refusal test had no injection seam that R3 does not
itself reject (B7).

Seven majors. Task 4 could not catch a wrong container layout at all
(M-a); Task 8's 503 row cannot spend seven metered tokens in under a
second at the shadow's own two-second pace (M-b); two of Task 9's named
mutations cannot fire (M-c); `DUBPLATE_SHADOW_ENABLED=false` became a
load error naming itself (M-d); the new doctor rows break
`doctor_test.go`'s exhaustive line count and the duration row's
registration is test-order dependent (M-e); the exit gate's three
strong-scoring bodies had no recipe and no reference (M-f); and an
`imported` row whose library file carries no `mb_albumid` fabricated a
`go-applies-different-release` (M-g).

**Revision 1, 2026-09-05.** Drafted against the rung 4 spec at **revision
2** (`docs/superpowers/specs/2026-09-05-rung-4-shadow.md`), whose six
owner rulings are **OPEN**. Under the 2026-09-05 unattended grant this
plan applies each recommendation **provisionally**, marks every
application inline as "provisional, owner ruling N open", and states in
one sentence what changes if a ruling comes back the other way, so a
later fold is a named edit rather than a rediscovery. The conductor's
four fold rulings (spec `:59`) are settled and are read here as ruled:
the wire pace is `2 * musicbrainz.DefaultInterval`; both contested-surface
rows are written; rung 4's second scoring mode tracks `conform`'s
`ModeDefault` re-score and 4a retires that mode's in-code provisional
mark; `DUBPLATE_SITE_NAME` stays a template literal and the age store is
not re-keyed.

**Pass 4b is out of scope in every part.** The two unit files, the
greenfield environment template, the six `deploy.sh` edits, the two
`-ldflags` stamps, `dubplate-server version` printing both rows, the
fresh-context unit read, the deploy sitting, the SELinux verifications,
and the recorded deploy date are all 4b's (spec `:1617`). 4b's deploy
waits for the owner sitting ruling 6 recommends, and nothing in this
plan blocks on it.

This plan's primary input is
`docs/superpowers/research/2026-09-05-rung-4a-planning-inventory.md`, a
research report read against the tree at HEAD
`54e534b0e83a305fe02c58cc41dd05ec7ee573a0`. Section (a) is the spec as
this plan's contract, (b) the engine, harness and config seams as
landed, (c) the estate inputs and the process record, (d) the
reconciliation and the twenty deduplicated open questions. **Do not
re-derive what it settles; cite it, and verify the line if you carry one
forward.** Its twenty open questions are **resolved in this plan**, each
with its evidence and its rejected alternative, in "Resolutions this
plan takes rather than discovers".

**The spec's own do-not-re-litigate record is the 28-item "Verified
sound" list of `research/2026-09-05-rung-4-spec-review.md:487-676`,
summarized at inventory `:604-626`, plus the one half-finding rejected
there with its evidence (inventory `:628-635`).** An implementer who
believes one of them is wrong reports a finding and implements the
record anyway.

**Three spec defects this plan reads past rather than reproduces**, each
carried into Task 15's drafted corrections:

1. `spec:1608-1609` cites "ruling 7" for the deploy sitting. The deploy
   sitting is **ruling 6** (`spec:342-368`); revision 1's seventh ruling
   on response storage was retired rather than answered (`spec:110-123`)
   and the cross-reference was not renumbered.
2. `spec:299-301` concludes "at one request per second the run's wire
   time stays inside sixty seconds". That sentence predates the M6 fold
   twenty-five lines later, which re-derives the same run at two seconds
   a request (`spec:328-330`). Under the conductor's ruling that the
   pace is `2 * DefaultInterval`, **only the second figure is quoted
   here**, and it is stated tighter than the spec states it: sixty
   tokens at two seconds a token bounds a run at two minutes.
3. `spec:1824-1825` says the six rulings are "batched at one checkpoint
   before 4a's plan is drafted". The conductor's fold paragraph
   (`spec:59`) and `docs/STATUS.md:24-26` both record them as open with
   4a proceeding provisionally. **This plan does not read `spec:1824` as
   a gate**; the operative instruction is the spec header's.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section: the pass-execute chain per task, sequential. The conductor
> consumes reports and never reads diffs inline. Subagents never
> dispatch other agents; `code-simplifier` is the conductor's dispatch
> after a task's review accepts. The conductor also owns the ledgers
> (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`), the lane
> registry, the contested-surface rows, and the batch commit of anything
> a task drafts into them. Rung 4 is not an adversarial-cycle rung
> (those are 5, 12 and 16), so review scales to the in-chain
> `diff-reviewer` per task plus the pass-end gate (`spec:1800-1803`).
> **This pass runs in the MAIN TREE**, and nothing here may be executed
> from a worktree: rung 4 writes `internal/harness/analysis`, and the
> standing policy is that corpus and harness work never moves to a
> worktree (`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`,
> lane-launch checklist item 5; spec `:1619-1622`).

## Primary inputs

Read these before Task 1, in this order:

1. **`docs/superpowers/research/2026-09-05-rung-4a-planning-inventory.md`,
   in full.** It is this plan's argument.
2. **`docs/superpowers/specs/2026-09-05-rung-4-shadow.md`, revision 2**,
   in full. The pass-split table's 4a row (`:1616`) is this plan's scope
   statement. The six rulings (`:125-368`), "The observation mechanism"
   (`:370-631`), "The divergence log and the adjudication artifact"
   (`:632-913`), "The runner and its subcommand" (`:915-1132`), "Config
   keys" (`:1423-1470`), "What rung 4 does not build" (`:1472-1494`),
   and the staleness pre-task (`:1659-1712`) are the sections this pass
   implements.
3. **`docs/HISTORY.md`'s "Rung 3, pass 3d" entry, in full**, especially
   the "rung 3 to rung 4 handoff" section (`:267-302`) and the forward
   carries (`:304-394`). The handoff's item 3 records the wire pace as
   `DefaultInterval`; the conductor's ruling (`spec:59`) amends it to
   `2 * DefaultInterval`, and Task 15 drafts the HISTORY correction.
4. **`docs/superpowers/plans/2026-09-05-rung-3d-verdicts-gate.md`**, for
   its Global Constraints and its execution model. This plan does not
   restate them; it adds the shadow contracts below.
5. The tree, read-only: `internal/engine/tag/{tag,doc}.go`;
   `internal/engine/matcher/{info,candidate,verdict,compat,tunables,penalty}.go`;
   `internal/engine/musicbrainz/{client,limiter}.go`;
   `internal/harness/analysis/adjudication.go` and its test;
   `internal/harness/cassette/cassette.go`;
   `internal/harness/conform/{case,arch_test}.go`;
   `internal/config/{config,keys,thresholds}.go` and their tests;
   `cmd/dubplate-server/{main,check,doctor}.go` and their tests;
   `_salvage/internal/atomicfile/`, `_salvage/internal/importer/importer.go`,
   `_salvage/internal/beets/beets.go`, `_salvage/internal/review/review.go`.
6. `CLAUDE.md`'s Build process, Gate and Stack rules sections, and
   `docs/STATUS.md`'s lane registry.
7. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** No
   Python and no unit file in this pass, so `python-conventions` and
   `vps-conventions` are out of scope, except for the one placement rule
   Global Constraints carries below.

## Goal

Ship the shadow runner: the instrument that measures the Go matcher
against the incumbent's own live beets decisions, and the store that
makes each measurement durable, honest and mechanically re-evaluable.

Land `internal/atomicfile` (re-landed from salvage on merit),
`matcher.QuietImportTunables`, the `internal/config` integration class
with its six rows and its duration threshold, and
`internal/engine/shadow` in full: a read-only FLAC reader and its
Vorbis-key mapping, the capture with its grouping and partial-read
rules, the reconcile step with its basename join and consumed-report
ledger, the match step with its recording transport, the `DefaultCompat`
re-score, the divergence classification with its four deferring
branches, the store, the launch-gate report, and the mechanical
re-evaluator. Land `analysis.AdjudicationEntry`'s two schema fields and
their checks, the marshalling in `cmd/dubplate-server`, and the `shadow`
subcommand's three verbs.

Nothing here touches the box, deploys anything, writes `corpus/`, adds a
module dependency, or reaches the live network.

## Owner touchpoints: none

4a expects no attended sitting and raises no question. The six open
rulings are applied provisionally and marked; the deploy sitting is 4b's
(ruling 6); ruling-F-style escalation does not apply to this rung; and a
pending divergence cannot arise, because no box run happens in 4a and
every pending row this pass produces comes from hand-authored `testdata`.

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition below resolves to the conductor. Under the
2026-09-05 grant recorded at `docs/STATUS.md:24-29`, the 80%-of-ceiling
flag routes to the conductor through 2026-09-07 09:00.

Task 15 drafts an owner ask and the spec corrections for the
conductor's batch, none of them blocking.

## Spec

`docs/superpowers/specs/2026-09-05-rung-4-shadow.md` (revision 2), and
behind it `specs/2026-09-02-greenfield-server-build.md` (revision 2.1,
ratified), ladder entry 4, quoted in full at `spec:65-72`. 4a builds
deliverables 1 through 5 of the spec's six (`spec:76-97`); deliverable
6, the deployment, is 4b's.

**The rung 15 forward criterion travels in this pass's report**, so rung
15 reads it from the ledger rather than rediscovering it:

> Rung 15 re-runs `internal/harness/conform`'s three tiers over the
> complete corpus, at the matcher commit as landed, and its acceptance is
> the launch bar's matcher clause read against that run: zero false
> confident matches over every stratum including stock-config, and zero
> unadjudicated beets-right/Go-wrong cases. Any tier-2 divergence
> carrying a pending row at rung 15 is the launch gate's business.

Rung 4 adds a second forward criterion, `BS:69-71`'s launch bar part 3
read against the shadow record: zero unadjudicated divergences and zero
false confident matches against live beets decisions. The report's own
two gating rows are what rung 15 reads (contract S22 below).

## Token ceiling: 5M

As a **starting** figure, per the spec's Band and ceiling section
(`spec:1649-1657`), and **re-derived from Task 4's actual report**, Task
4 being the first `internal/engine/shadow` task and the pass's one
genuinely unbounded piece.

The basis, read off HISTORY rather than asserted. 3d closed at **≈5.15M
against a 6M ceiling with zero execution sittings**
(`docs/HISTORY.md:474-483`), of which prep was 2.51M, Task 0 0.09M, the
task chain ≈2.05M over nine tasks (≈230K per task), the simplify pass
plus its review 0.22M, the fresh-context runner review 0.14M, and the
close ≈0.15M.

4a is **fifteen tasks plus a pre-task**. At 3d's measured ≈230K per task
that is ≈3.45M, plus roughly one task's worth again for each of the two
`opus` upshifts (Tasks 4 and 9), so ≈3.9M for execution. The
`code-simplifier` pass and its review add ≈0.2M. This close is smaller
than 3d's, carrying no corpus census and no verdict-contract table, so
≈0.25M. 4a's own plan-prep workflow is bounded by 3d's 2.51M and is
expected well under it, since it runs no venv probe and no corpus
measurement; call it ≈1.5M.

**Expected total ≈5.9M against a 5M starting ceiling.** The arithmetic
already exceeds the start, and the plan says so rather than shading the
estimate: **the ceiling is the stop rule's yardstick, not a forecast**,
and the re-derivation point exists because Task 4's FLAC reader plus its
nineteen-field mapping is the piece a plan can specify but not size. On
3d's precedent the likely re-derived figure is **6.5M**, and the
conductor sets it on Task 4's number rather than on this paragraph.

**Flag at 80% of the live ceiling (4M against 5M):** finish the task in
flight, write STATUS, and one combined question goes to the conductor,
per the standing grant recorded at `docs/STATUS.md:24-29`. **That same
STATUS bullet records a window spend cap of 15M with about 7.6M already
spent**, which is the other live constraint on this window and is named
here rather than half-quoted. About 7.4M of that window remains, and an
expected total near 5.9M leaves roughly 1.5M of it, so **a re-derivation
past 6.5M is the point at which the conductor checks the window cap
rather than assuming it**.

**Checkpoint: STATUS after every task.** That is 2b's through 3d's
cadence; every task reports its own spend and its own gate evidence, and
the conductor sizes the next task against it.

**Rung 4's combined ceiling: 6.5M, and it stays a reservation.** It is
4a's 5M plus 4b's 1.5M (`spec:1652`). `docs/STATUS.md:29-30` records
lane-2 as BLOCKED until Phase 0b and 0c land, so no second sub-ceiling
exists to sum today.

## Global Constraints

- **Main tree.** The spec's tree column for 4a is "Main tree"
  (`spec:1616`), and rung 4 holds the primary tree for both passes
  (`spec:1621-1622`). The lane registry entry is written before the first
  dispatch, and **the live-executor check is that registry, never
  `pgrep -f Projects/dubplate`**, which matches every lane once a
  worktree exists (CLAUDE.md, Lanes).
- **Package territory, stated as a wall.** This pass writes
  **`internal/atomicfile`** (born this pass, one file plus its test),
  **`internal/engine/matcher`** (`QuietImportTunables` and nothing
  else), **`internal/engine/shadow`** (born this pass),
  **`internal/config`** (`keys.go`, `config.go`, `thresholds.go` and
  their tests), **`internal/harness/analysis`** (two fields and their
  checks), **`internal/harness/conform`** (one drift-guard assertion and
  one doc-comment clause), and **`cmd/dubplate-server`** (`shadow.go`
  plus its test, `doctor.go`'s threshold switch, `main.go`'s
  `AddCommand` block), plus conductor-owned `docs/` ledgers. It writes
  **nothing** under `corpus/`, `systemd/`, `scripts/`, `env/`,
  `Makefile`, `internal/harness/{cassette,stub,synth,mbrequest,oracle}`,
  `internal/engine/{musicbrainz,tag}`, `tools/`, `web/`, or `_salvage/`.
  A task that believes it needs one of those stops and reports.
- **`corpus/` is frozen and read-only in this pass.** No task loads a
  corpus case, edits `corpus/analysis/*`, or reads a `responses/*.json`
  body into context. Every fixture is hand-authored in the owning
  package's own `testdata`, on the standing rule.
- **The engine never imports the harness, and never reads the clock.**
  `TestEngineNeverImportsHarness`
  (`internal/harness/conform/arch_test.go:141-146`, walker at `:30-65`)
  and `TestEngineNeverReadsTheClock` (`:152-157`, walker at `:75-120`)
  both root at `internal/engine`, descend into nested packages
  (`:194-206`), and walk `_test.go` files too. **`internal/engine/shadow`
  is inside both walks from the moment it exists**, its own tests
  included. `now` and `evalDate` are parameters everywhere. An
  implementer that finds itself proposing an exemption stops and reports.
- **`internal/engine/shadow` imports `internal/config` never.** No
  mechanical rule stops it; the convention is written prose at
  `internal/engine/matcher/tunables.go:50-54` and
  `internal/engine/musicbrainz/client.go:86-92`, and the shadow takes
  its paths, floor, cap, User-Agent and matcher stamp as parameters on
  the same grounds (Resolution R16).
- **Assertion discipline, at the spec's own bar.** "The observation
  mechanism's own honesty gets one mutation proof per checker"
  (`spec:1816-1818`). A shadow that silently observes nothing is this
  rung's equivalent of a conformance runner that grades nothing, and it
  would cost thirty days before anyone noticed. No silent-success fakes,
  no vacuous assertions, no self-derived expectations. **Goldens are
  last resort** (`go-conventions` SKILL.md:355-360): the natural way to
  write the end-to-end assertion is a golden `observation.json`, and the
  rule says assert on parsed structure instead.
- **Module dependencies: none added.** `go.mod` carries cobra and its
  two indirects and nothing else (`go.mod:5`). No FLAC library, no
  sqlite driver, no YAML parser (`spec:1483`). A task that wants one
  stops and reports.
- **Universality.** No estate hostname, path, email, catalog value or
  personal name anywhere outside `internal/config`
  (`spec:1416-1421`, `BS:116-120`). Hand-authored fixtures are invented
  strings. The runner is adopter-optional behind
  `DUBPLATE_SHADOW_ENABLED`.
- **No live network. Anywhere.** Every test points at an
  `httptest.Server` on an ephemeral loopback address, on
  `internal/engine/tag/tag_test.go:25-51`'s own precedent. Any replay
  client takes `musicbrainz.Unmetered()` (`limiter.go:354-359`), exactly
  as `gradeClient` does (`internal/harness/conform/case.go:330-336`).
- **The shadow opens `<incumbent state dir>/import.lock` never.** The
  incumbent treats a held lock as healthy, writes a `LockHeld` report,
  pings success and returns nil (`_salvage/internal/importer/importer.go:397-413`),
  so a shadow holding it would stall the box's import while reporting
  green. `flock` on a read-only descriptor succeeds on Linux, so this is
  a code rule with its reason in the code, not a grant a unit can
  withhold (`spec:591-597`). The reconcile step reads
  `report-*.json` and `import-*.log` by name and never opens any other
  path under the incumbent's state directory.
- **The store's root is a single directory.** Everything the shadow
  writes lives under `<state dir>/shadow/`, so 4b's `ReadWritePaths=` is
  one path and the argument is static rather than assumed
  (`~/.claude/skills/vps-conventions/SKILL.md:39-45`). No task writes
  anywhere else on any filesystem outside `t.TempDir()`.
- **`internal/config` is rung 4's for the length of this pass.** Rung 1b
  has not landed (`docs/STATUS.md:29`), rung 4 is the only live code
  lane, and the contested-surface row says so
  (`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:183`).
  **The `.gitleaksignore` rider:** its single line pins a finding to
  `internal/config/keys_test.go` at **line 66**, by line number, so any
  insertion above line 66 in that file moves the finding and reds the
  gitleaks leg of `scripts/check.sh` for every lane at once. **The rule
  for every task touching `keys_test.go`: append below line 66, or
  re-fingerprint `.gitleaksignore` in the same commit.** Appending a new
  test function at the end of the file is the safe form; adding rows to
  `keys.go` shifts nothing in `keys_test.go`. **Neither this plan, any
  task's code, any comment, nor any commit message quotes the pinned
  fixture literal**, because gitleaks scans the tracked worktree with
  `--no-git` and would flag the quotation itself.
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck
  over `scripts/*`, ruff over every tracked `*.py` (this pass adds
  none), gitleaks over the tracked worktree, vale over Go comment prose,
  `docker compose config`, and a pattern-filtered `systemd-analyze
  verify`. **Never commit red. Specific files only.** Commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`. **Tier 0's 72 rows,
  the translation tier's 50, the request tier's 215/303/48, tier 1's 215
  cases and tier 2's two modes are part of every task's gate
  expectation**, because all of them run inside `go test ./...`; 3d
  measured the whole run at 2m06s and `scripts/check.sh` at about 3.5
  minutes (`docs/HISTORY.md:261-265`), so **a task's gate run is minutes,
  not seconds, and that is expected**.
- **`go-conventions` before any Go.** The bar is the standing
  exemplary-stdlib-idiom bar (`idiomatic-beautiful-go`). Concrete types
  over interfaces, no functional options, `slices.SortStableFunc` with
  `cmp.Compare` or `cmp.Or`, `sync.OnceFunc` over a `sync.Once` pair,
  `any` not `interface{}`, no `init()`, no `context.Context` in a
  struct, `ctx` first, errors lowercase and context-first with `%w` only
  where a caller branches. **The em dash is banned in Go comments**, with
  no opt-out, and vale scans a brand-new file before it is staged
  (`.vale.ini:5-11`, `scripts/vale-comments.sh:53-59`).
- **Atomic writes.** Every file the store writes goes through
  `atomicfile.WriteFile(path, data, perm)`, directories 0755, plain
  files 0644 (`go-conventions` SKILL.md:298-310). No task calls
  `os.WriteFile` on a store path.
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each before
  Task 1 dispatches.

## Contested surfaces: nine of nine written

Lane-launch checklist item 2 requires a written owner for every surface a
lane touches (`PA:197`). 4a touches eight of the spec's nine rows (the
ninth, `systemd/` and `scripts/deploy.sh` and `env/`, is 4b's). **All
nine were written into the parallel-approach spec's ownership table on
2026-09-05**, each stamped `conductor ruling, 2026-09-05`, verified at
HEAD:

| Surface this pass writes | Row |
| --- | --- |
| `internal/engine/shadow` | `PA:180` |
| `internal/atomicfile` | `PA:181` |
| `cmd/dubplate-server` | `PA:182` |
| `internal/config` (`keys.go`, `thresholds.go`) | `PA:183`, carrying the `.gitleaksignore` rider |
| `internal/harness/{cassette,stub,analysis}` | `PA:169`, amended for rung 4's `MatcherCommit`, `MatcherConfigHash` and their checks |
| `internal/engine/{matcher,musicbrainz}` | `PA:171`, amended for `QuietImportTunables` **and nothing else** |
| the state directory layout | `PA:178` |
| `corpus/analysis/adjudications.json` and the divergence schema | `PA:179` |
| `docs/` ledgers | `PA:177`, conductor-owned; a pass drafts and never commits |
| `systemd/`, `scripts/deploy.sh`, `env/` | `PA:184`, **4b's half, untouched here** |

Two surfaces this pass writes are covered by an existing row rather than
a rung 4 amendment, and both are named so a reviewer does not read them
as unowned: `internal/harness/conform` (`PA:167`, "rung 3's lane
thereafter") gains one drift-guard assertion and one doc-comment clause,
both of which the conductor's own fold ruling assigns to 4a
(`spec:59`, `spec:255-259`); and `internal/engine/matcher`'s row
(`PA:171`) already names `QuietImportTunables` by name. Task 0 reports
both; neither is a launch blocker.

**Lane-launch item 2 is therefore satisfied before 4a's launch**, and
the plan cites these rows rather than re-deriving them.

## The six open owner rulings, applied provisionally

Each row states the application, the inline marker every task carrying it
must reproduce, and **the one-sentence fold if the ruling comes back the
other way**. The fold sentence is what makes a later ruling a mechanical
edit.

| # | Question | Applied as | Fold if ruled otherwise |
| --- | --- | --- | --- |
| **1** | Where the shadow's configuration lives (`spec:125-162`) | **Option (a)**, an `IntegrationClass` table in `internal/config` beside `reservedClasses`, with required-when-enabled semantics (Task 3). Marker: "provisional, owner ruling 1 open" | Under (b) the class table collapses to a flat accepted-name set, `Load` requires nothing, and `doctor` prints one row instead of a per-key status; under (c) Task 3 is deleted whole and the six values arrive as cobra flags on `dubplate-server shadow`, with `Config.Shadow` and the `Register` call site going with it |
| **2** | How a skipped album's `Subject.ReleaseMBID` is filled (`spec:164-190`) | **Option (a)**, the release under dispute, filled by whichever side proposed one, with `AdjudicationSubject.ReleaseMBID`'s doc comment restated to say so (Tasks 9, 10, 11). Marker: "provisional, owner ruling 2 open" | Under (b) `checkAdjudicationEntry`'s pair rule relaxes to accept `Album` alone for a shadow subject, `pendingRowFor` stops filling `ReleaseMBID`, and Task 10's half-shadow fixture inverts from a firing case to an accepted one |
| **3** | Which `matcher.Compat` the shadow runs (`spec:192-259`) | **Option (c)**, `FullCompat()` for the recorded flow plus a `DefaultCompat()` re-score of that run's own candidate set, with the derived figure's limit stated in the row itself (Tasks 8, 9). Marker: "provisional, owner ruling 3 open" | Under (a) Task 9's `rescore.go` and the `Default` verdict half of `Observation.Go` are deleted and the compat-class annotation goes with them; under (b) Task 8's single flow runs `DefaultCompat()` instead of `FullCompat()` and Task 9's re-score is deleted |
| **4** | The estate's own shadow floor (`spec:261-276`) | **The shipped 30-day default**, `720h`, as `DUBPLATE_SHADOW_FLOOR`'s registered `Default` (Task 3). Marker: "provisional, owner ruling 4 open" | A different duration is one string literal in `shadowFloorThreshold`, and the doctor row and the report's floor row follow it with no other edit |
| **5** | The shadow's MusicBrainz budget cap (`spec:278-340`) | **60 requests per pass**, as `DUBPLATE_SHADOW_BUDGET`'s registered `Default`, spent at `2 * musicbrainz.DefaultInterval` (Tasks 3, 8). The pace half is the conductor's ruling and is settled. Marker: "provisional, owner ruling 5 open" | A different cap is one string literal in `shadowBudgetThreshold`; the pass record's `budget_spent` row and the report's budget arithmetic are unchanged, because both read `Budget.Spent()` rather than the cap |
| **6** | Whether the first deploy is a conductor action or an owner sitting (`spec:342-368`) | **Option (a)**, an owner sitting. **4a is unaffected either way**: no box write happens in this pass. Marker: none needed in code | Under (b) 4b's plan drops its sitting task and reports the deploy afterwards; nothing in 4a moves |

## The three conductor rulings this fold owes, applied provisionally

Revision 2 folds two adversarial reviews, and three of their findings
resolve to a decision above this plan's authority: a spec amendment, a
spec deviation on a contested surface, and a scope question about a
contested surface. **Each is applied as recommended, marked inline where
it lands, and listed here with what changes if the ruling goes the other
way.** No task blocks on any of them.

| # | Question | Applied as | Fold if ruled otherwise |
| --- | --- | --- | --- |
| **C1** | The capture's grouping key (fold, B4). `spec:441` states it as "their `albumartist` and `album` fields"; pinned beets groups on `(item.albumartist or item.artist, item.album)` (`beets/importer/stages.py:102-103`) | **Beets' own key**, `(cmp.Or(AlbumArtist, Artist), Album)`, in contract S9 and Task 6, with a firing mutation row for the fallback. Marker: "provisional, conductor ruling owed" | Task 6 drops the `cmp.Or`, its second exclusion row and mutation 2 revert to revision 1's shapes, and a flat directory with no `ALBUMARTIST` is observed as one album against several oracle rows, which Task 15 then records as a declared gap rather than a defect |
| **C2** | `newShadowCmd`'s signature (fold, B7). `spec:1063-1066` pins `(environ []string, stdout, stderr io.Writer)`; Resolution R3 rejects the package variable a stamp table would otherwise have to write | **One extra parameter**, `newShadowCmd(environ []string, matcherCommit string, stdout, stderr io.Writer)`, with `newRootCmd` passing the package variable. Marker: "provisional, conductor ruling owed" | The refusal moves into `shadow.Run`'s own configuration validation, where `Config.MatcherCommit` is already a field the engine test can vary, and Task 12's verb test asserts only that the error propagates |
| **C3** | Whether an architectural walk guards "`internal/engine/shadow` imports `internal/config` never" (fold, m14). Closing it widens this pass's `internal/harness/conform` territory past the one drift-guard assertion Global Constraints names | **Declared as a gap**, Task 15's list item 8, with the `diff-reviewer` as the meanwhile guard. Marker: the gap entry itself | Task 2 gains a third `conform` edit: a sibling of `checkNoHarnessImports` (`arch_test.go:30-64`) over a second forbidden prefix, with its own firing counter-example fixture beside the two that exist |

## The shadow contracts, stated once

Every task below cites this section instead of restating it. Sources are
the spec, the tree at HEAD, and pinned beets 2.13.1 and mediafile 0.17.0
under `tools/venv/.venv/lib/python3.13/site-packages/`.

**S1. The entry point is fixed.** `internal/engine/tag/tag.go:68-69`:

```go
func Album(ctx context.Context, src *musicbrainz.Client, items []matcher.ItemFields,
    tun matcher.Tunables, compat matcher.Compat, evalDate time.Time, opts Options) (*Proposal, error)
```

`Proposal` is five fields (`tag.go:23-29`): `CurArtist`, `CurAlbum`,
`Candidates []matcher.Candidate` (ascending distance, stable),
`Recommendation matcher.Recommendation`, `Action matcher.Action`. The
shadow passes `Options{}`: `SearchIDs` selects the explicit-ID branch
(`tag.go:78-85`) and the shadow has none. `Candidates[0]` is well
defined on both flow paths whenever the call returns without error
(inventory `:965-970`).

**S2. The estate tunables are `matcher.QuietImportTunables()`**:
`DefaultTunables()` (`tunables.go:266`) with four assignments,
`QuietFallback = "skip"`, `NoneRecAction = "skip"`, `Timid = false`,
`Quiet = true`. Two of the four already ship
(`tunables.go:312`, `:316`); `Quiet` ships `false` (`:315`) and
`NoneRecAction` ships `"ask"` (`:317`). The box's live config reduces to
exactly those (`config/beets-config.yaml:16-18` plus the `-q` flag at
`_salvage/internal/beets/beets.go:87`), and the estate's beets config
sets no `match:` and no `musicbrainz:` block, so the whole distance and
threshold surface runs at beets' shipped defaults
(`internal/harness/conform/defaults.go:16-18`).

**S3. The oracle's signal is binary, and it is the only comparable
quantity.** In quiet mode `_summary_judgment` returns `APPLY` when and
only when the recommendation is strong, otherwise resolving
`quiet_fallback` to `SKIP`
(`tools/venv/.venv/lib/python3.13/site-packages/beets/ui/commands/import_/session.py:335-341`),
mirrored exactly by `matcher.ActionFor` (`verdict.go:179-210`). Under
`QuietFallback = "skip"` the Go projection is exactly `apply` / `skip`
and can never be `asis` or `ask`. `RecMedium`, `RecLow` and `RecNone`
are indistinguishable to the oracle. Distances, per-key penalties, track
vectors and assignments are logged as evidence and never graded
(`spec:610-623`).

**S4. Four outcomes, three of them divergences** (`spec:636-649`):

| Oracle | Go | Class |
| --- | --- | --- |
| apply, release R | apply, release R | `agreement`, no row |
| skip | skip (or `tag.ErrNoCandidates`) | `agreement`, no row |
| skip | apply, release G | **`go-applies-oracle-skipped`** |
| apply, release R | skip (or `tag.ErrNoCandidates`) | **`go-skips-oracle-applied`** |
| apply, release R | apply, release G, `R != G` | **`go-applies-different-release`** |

`go-applies-oracle-skipped` is the class a false confident match hides
in and the launch bar's own clause (`BS:69-71`).

**S5. Four branches are neither, and the enumeration is closed
deliberately** (`spec:651-660`). `musicbrainz.ErrUnavailable`
(`client.go:74-84`), `musicbrainz.ErrBudgetExhausted`
(`limiter.go:45-48`) and `musicbrainz.ErrLimiterClosed`
(`limiter.go:50-53`) each **defer** the observation for a later pass;
`tag.ErrNoCandidates` (declared `tag.go:18`, returned `:139`) records
"no match", which agrees with an oracle skip and diverges from an oracle
apply. **The shadow classifies with `errors.Is`, never equality**,
because `tag.Album` wraps every client error in a message
(`tag.go:128`, `:170`).

**S6. No error other than `tag.ErrNoCandidates` may ever be recorded as
a Go skip** (`spec:662-668`). A truncated album written as
`go-skips-oracle-applied` is a fabricated divergence in the one artifact
whose whole value is that it is honest, and it consumes exactly the
owner adjudication ruling B's stamp exists to save. An unclassifiable
error defers with its own reason and is counted, never recorded as a
verdict.

**S7. The capture's digest is the identity.** A fixture records, per
album directory, the `matcher.ItemFields` read from each track, the
capture time, and a digest over the file set (names, sizes, modification
times). A directory whose digest has not moved is not re-captured, so
the capture is idempotent (`spec:405-411`). The observation id **is**
that digest, prefixed (Resolution R5).

**S8. A capture that cannot read every entry it listed is discarded**
(`spec:413-423`). Never stored, never scored, never counted; the pass
counts it as a partial read and moves on. The supersede rule does not
cover this, because once the importer moves an album out of the inbox
there is no next pass to supersede anything, and `review.Move` tries
`os.Rename` first with a copy-then-remove EXDEV fallback
(`_salvage/internal/review/move.go:24-40`), so a walk can read three
tracks and take ENOENT on the fourth.

**S9. The grouping rule, because the box's album unit is not the
directory** (`spec:425-455`). `config/beets-config.yaml:23` sets
`group_albums: yes`, with the estate's own 2026-08-31 reason at
`:19-22`; the incumbent runs one `beet import` over the whole staging
tree (`_salvage/internal/importer/importer.go:1100`); and `tag.Album`
has no grouping, calling `matcher.MostCommonTags(items)` over every item
it is handed (`tag.go:75`). **Rung 4 observes only what groups to
exactly one album, and counts the rest.** More than one group is
**excluded**, with its group count and its basename recorded, and the
report prints the excluded-multi-album count beside the gate-rejected
one.

**The grouping key carries beets' own artist fallback** (fold, B4).
Pinned beets groups on `(item.albumartist or item.artist, item.album)`
(`tools/venv/.venv/lib/python3.13/site-packages/beets/importer/stages.py:102-103`,
consumed at `:112-113`), not on `albumartist` alone. The estate's own
material is the reason it matters: `config/beets-config.yaml:19-22`
records contributor ZIPs extracting flat, and a flat directory whose
tracks carry `ARTIST` and no `ALBUMARTIST` maps every track to
`("", "<album>")` under an albumartist-only key, groups to one, and is
observed, while beets splits it by artist into several verdicts. So the
key is `(cmp.Or(AlbumArtist, Artist), Album)`. **Provisional, conductor
ruling owed**: `spec:441` states the key as "their `albumartist` and
`album` fields", so this is a spec amendment rather than a plan-internal
choice. If the ruling keeps the spec's wording, Task 6 drops the
`cmp.Or` and its mutation-2 row reverts.

**S10. The join key is the album basename, never `SourceDir`**
(`spec:476-491`). `AlbumOutcome.SourceDir` is the bare field at
`importer.go:136` and is documented "the uploader's inbox directory" on
`albumUnit.sourceDir` at `:212` (fold, m2), and all three assignment
sites set it at the uploader level: `:766` (`filepath.Dir(srcDir)`),
`:908` (`inboxUserDir`) and `:1066`
(`filepath.Join(InboxDir, uploader)`) (fold, m1);
`albumUnit.stagedDir` is the only per-album path (`:213`) and `toReport`
does not copy it (`:231-240`); `stageAllContributors` stages every
quiescent album of every contributor in one run (`:782-800`), so several
rows routinely share one `SourceDir`. `SessionID` does not rescue it
either (`:139-142`). Per outcome:

- **`imported` is an apply.** The oracle's chosen release MBID is read
  from `MUSICBRAINZ_ALBUMID` on the first file `LibraryPaths` names,
  through the shadow's own FLAC reader. Beets wrote it there under
  `write: yes` (`config/beets-config.yaml:15`), and
  `MUSICBRAINZ_ALBUMID` is mediafile's Vorbis key
  (`mediafile/__init__.py:782`). **The applied files are matched back to
  the fixture by their track titles and lengths**, because
  `LibraryPaths` is itself a positional correlation: the importer fills
  it by matching staging order against the beets database's album order,
  since "one shared beet invocation carries no per-uploader join key",
  and a count mismatch leaves the extra units empty (`:1119-1136`).
  **An `imported` row with no resolvable release id is `unreconciled`,
  never an apply** (fold, M-g; Resolution R22). `readFLAC` gives an
  absent tag the field's Go zero value, so a library file carrying no
  `MUSICBRAINZ_ALBUMID` yields `MBAlbumID == ""`, and an oracle signal
  of `apply` with an empty release would classify against every Go apply
  as `go-applies-different-release`: a divergence against a release the
  oracle never named, which is exactly the fabricated row contract S6
  forbids.
- **`review` is a skip**, resolved by the album's basename in a tag
  log's `skip ` lines (`_salvage/internal/beets/beets.go:115-135`, never
  swept) or under `<review dir>/<date>/<uploader>/<basename>`.
- **`quarantined` and `rejected` are not oracle verdicts at all**
  (`importer.go:945`, `:499-505`). The observation is excluded, counted
  and named, never scored.

**S11. Four `unreconciled` reasons, on the "report what was actually
compared" discipline** (`internal/harness/conform/case.go:109-110`).
Three are the spec's (`spec:522-528`): an ambiguous basename; an
`imported` outcome with **no resolvable release id**, which covers both
an empty `LibraryPaths`, described by the importer's own `Error` field
as "a beets-database match that never turned up" on an outcome that
still reads `imported` (`importer.go:145-149`), and a library file the
reader opens whose `MUSICBRAINZ_ALBUMID` is absent (fold, M-g); and a
fixture whose run report was already swept. **The fourth is a
partial skip** (Resolution R4): `resolveImported` marks a unit
`OutcomeReview` with `Error: "left in staging after import"` when beets
took some but not all of its files, and the function's own comment calls
it "the one shape a beets tag log can produce that this
album-as-a-unit model does not cleanly resolve" (`importer.go:1179-1199`).

**S12. The run reports are a rolling seven-and-a-half-hour window**
(`spec:530-546`). `sweepReports` runs on every import pass
(`importer.go:389`) and keeps at most `retention` of the newest
(`:533-548`), with `defaultReportRetention = 30`
(`_salvage/internal/config/config.go:172`); the timer fires `*:0/15`
(`systemd/dubplate-import.timer:10`). **The shadow records which report
files it has already consumed, so a pass never re-reads one and a gap in
the report sequence is a named, reported condition rather than
silence.** Gap detection reads `StartedAt` and `EndedAt` out of each
report rather than trusting the filename, because `RunID` is 8 random
bytes and is "not derived from `StartedAt`" by design
(`importer.go:158-161`; `:139-145` is `SessionID`'s comment, fold m3);
Resolution R6 states the rule.

**S13. The store layout** (`spec:736-747`), rooted at one directory:

```
<state dir>/shadow/
    started-at                      the clock start, written once, never rewritten
    passes/<pass-id>.json           one record per observe pass
    observations/<obs-id>/
        observation.json            the divergence record, or the agreement record
        case.json                   the cassette-shaped fixture, Expected nil
        responses/NNN.json          the captured response bodies
    pending/<obs-id>.json           the AdjudicationEntry a checkpoint batch presents
    resolved/<obs-id>.json          a divergence reevaluate retired, and the commit that retired it
    consumed-reports.json           the incumbent run reports already reconciled
```

One directory per observation, so two writers never contend for one
file. Every file is written through `atomicfile.WriteFile`. **No lock
package** (`spec:767-773`): the timer path is a systemd singleton,
`reevaluate` writes only under `resolved/`, `report` opens nothing for
writing, and `observe` only ever creates.

**S14. `passes/` exists because a `Budget` is in-process state**
(`spec:749-757`). `spent` is an int guarded by a mutex on the struct
(`limiter.go:342-344`, `:400-404`), created per call to `Limiter.Budget`
(`:139-145`), and nothing persists it. `observe` and `report` are
separate process invocations, so a `report` that built its own limiter
would print a zero that reads as a measurement of an unattended
thirty-day window. Each pass writes its own `Budget.Name()` and
`Budget.Spent()`, and `report` sums them, printing per-run and
cumulative separately.

**S15. The budget, the pace and the contact.**
`musicbrainz.NewLimiter(2 * musicbrainz.DefaultInterval)`
(`limiter.go:110-115`, `:62`) and
`Limiter.Budget("shadow", musicbrainz.ClassBackground, cap)`
(`limiter.go:139-145`, class at `:25` whose own comment names this
caller at `:16-18`). The cap ships at 60 (ruling 5, provisional). **A
limit of zero or less means no cap** (`limiter.go:130-133`), so a cap
must be positive for the exhaustion branch to be reachable at all. The
User-Agent is composed in `cmd/dubplate-server` as
`dubplate-server/<version> ( <contact> )` with the contact from
`Config.AdminEmail`, which `ClientConfig`'s doc comment specifies
verbatim (`client.go:86-92`); the engine never builds one, and the
shadow refuses to run unidentified because `NewClient` refuses a missing
`UserAgent` (`client.go:124-130`).

**S16. Mode one records; mode two re-scores** (ruling 3, provisional).
Mode one drives `tag.Album` under `FullCompat()` (`compat.go:31-33`)
through a recording `http.RoundTripper` installed as the `Transport` of
the `*http.Client` that `ClientConfig.HTTP` carries (`client.go:98` is
`HTTP *http.Client`, consumed `:137-144`; a round tripper cannot be
assigned to that field directly, fold m8). Mode two
**re-scores that run's own recorded candidate set** under
`DefaultCompat().Apply(tun)` (`compat.go:41-46`, `:51-63`), issuing no
search and driving no second `tag.Album`, in `gradeCaseDefault`'s own
five steps (`internal/harness/conform/case.go:568-622`): apply and
validate, look each recorded candidate up through a replay client,
`cands.Add` per candidate, `cands.Sorted()` then `matcher.Recommend`,
then `matcher.ActionFor`. **The derived figure's limit is stated rather
than absorbed**: a re-scored `DefaultCompat` verdict is not the
product's true one wherever the real flow would have widened the
candidate set, so the compat-class annotation is a hypothesis and the
row says so (`spec:230-237`).

**S17. Any replay client takes `musicbrainz.Unmetered()`.** The budget
sits above the transport: `doWithRetry` calls `c.budget.Wait(ctx)`
before every attempt (`client.go:346`) whatever `ClientConfig.HTTP`
holds, so a replay through the metered client would double the
`ClassBackground` spend and double `Budget.Spent()`, the figure the
report prints as honest (`spec:239-247`).

**S18. The replay transport answers a miss with a refusal, never a
404.** A 404 on a lookup is swallowed inside `lookupAndAdd`
(`tag.go:166-170`), so a silently absorbed miss would retire a
divergence that never re-ran. A newer matcher revision can take a
different branch through `tag.Album` for the same input, exactly as a
different `Compat` can, so the miss must surface (`spec:866-873`). The
refusal is HTTP 501, which `doWithRetry`'s `default:` arm turns into a
non-retryable "unexpected status" error in one attempt
(`client.go:381-382`). **The transport's own miss list, rather than the
error's shape, is the signal the caller branches on**, because the error
`doWithRetry` returns for a 501 is an unwrapped `fmt.Errorf` with no
sentinel to test: `replayTransport.misses` names every request the
recording did not carry, and a non-empty list is what makes an
observation un-re-evaluable (fold, m7; revision 1 cited Resolution R13
here, which is the incumbent's report decode struct and states no such
rule).

**S19. `internal/engine/shadow` marshals neither harness-shaped
artifact** (`spec:777-801`). The package cannot reach `cassette.Case` or
`analysis.AdjudicationEntry`. It returns typed engine-side results, and
`cmd/dubplate-server` imports both trees and writes `case.json` in
`cassette.Case`'s shape and `pending/<obs-id>.json` in
`analysis.AdjudicationEntry`'s. `observation.json` is the engine's own
type and needs no harness import.

**S20. Every staged row is validated before it is written**
(`spec:803-811`). `checkAdjudicationEntry` is unexported
(`adjudication.go:158`) and `analysis.CheckAdjudications` is exported
(`:138`); the subcommand runs it over each row before staging it, so a
malformed row fails within one pass rather than weeks later. Its
`caseIDs` argument receives `nil` on the box, which is sound because a
shadow subject never reaches the case-resolution branch
(`adjudication.go:164-167` is the only reader).

**S21. The two schema additions** (`spec:834-848`).
`analysis.AdjudicationEntry` gains

```go
MatcherCommit     string `json:"matcher_commit,omitempty"`
MatcherConfigHash string `json:"matcher_config_hash,omitempty"`
```

and `checkAdjudicationEntry` gains the matching rules: a shadow-subject
row (`Album` and `ReleaseMBID` both set, the branch at
`adjudication.go:168`) requires `MatcherCommit`; a case-subject row is
refused if it carries either field. Both are `omitempty`, and `readJSON`
sets `DisallowUnknownFields` (`ledger.go:19-24`), so the 21 committed
rows validate unchanged and neither key appears in the file.

**Byte-for-byte identity was never available, and the plan states that
rather than asserting it** (fold, B2). `spec:845-846` claims the 21 rows
"validate byte-for-byte unchanged"; two facts refuse it. The committed
ledger's key order is `id, subject, suspected, citation, release_truth,
disposition, reason, adjudicator, date, corrected_verdict[,
defect_free]`, while `AdjudicationEntry` declares `CorrectedVerdict`
before `Reason` (`adjudication.go:46-67`) and `encoding/json` emits
declaration order, so a re-marshal first diverges at line 12. And the
file carries one literal `<` and three literal `>`, which
`json.MarshalIndent` escapes; the repo records that hazard at
`internal/harness/cassette/cassette.go:9-11`. **The honest guarantee is
that the ledger round-trips through a decode and re-encode with no key
gained or lost**, which Task 10 asserts as a `map[string]any` tree
comparison, plus a raw-bytes assertion that neither new key appears.
Task 15 drafts the `spec:845-846` correction.
**`matcher_config_hash` is deliberately not named `config_hash`**:
`cassette.Expected.ConfigHash` hashes the Python shim's effective config
(`cassette.go:80`), and conflating them would make a rung 15 comparison
silently wrong.

**S22. Two clauses gate, not one** (`spec:1118-1132`). Launch bar part 3
reads "zero unadjudicated divergences **and zero false confident matches
against live beets decisions**" (`BS:69-71`). The first clause is the
pending count, because `Pending` "is legal at every gate this package
runs; the launch gate, a different gate at a different time, refuses it"
(`adjudication.go:10-12`). The second is the `go-applies-oracle-skipped`
count net of rows the owner disposed `oracle-wrong`, a join the
workstation does against the committed corpus rather than the box
(Resolution R21). The observed album count is reported and never gated
(`BS:249-253`, `BS:812-815`).

**S23. `disposition` stays owner-only.** A non-`Pending` disposition set
by `AgentAdjudicator` is a gate error (`adjudication.go:198-200`), so
every row the shadow writes is `pending-adjudication` with `adjudicator:
agent`. **The shadow writes into `corpus/analysis/adjudications.json`
never** (`spec:825-830`).

## Resolutions this plan takes rather than discovers

The inventory's twenty deduplicated open questions (`inventory:2098-2181`),
each resolved here with its evidence and its rejected alternative. A
task cites the resolution rather than re-arguing it.

### R1: the two threshold names join the integration class's `Optional` set (inventory Q1)

`rejectNamespace` exempts only `ownedNow(name)` (`config.go:135`) and
knows nothing about `registeredThresholds`, so a set
`DUBPLATE_SHADOW_FLOOR` fails `Load` with "unknown configuration key"
(`config.go:150`) even after `config.Register`. That is the exact
failure ruling 1 exists to prevent, and the spec picks neither of the
two available readings.

**`IntegrationClass` carries two key sets rather than one**:

```go
type IntegrationClass struct {
	Name     string
	Rung     int
	Flag     string
	Required []string // required when Flag is set; Load errors on each unset one
	Optional []string // accepted when Flag is set, never required; a threshold's default stands
}
```

`Required` is ruling 1's required-when-enabled half and covers the three
paths. `Optional` covers the two thresholds, whose whole point is that
they carry registered defaults, so requiring them would contradict the
registry. `rejectNamespace` accepts `Flag`, every `Required` name and
every `Optional` name when the flag resolves true, and errors naming the
flag when it does not; `Load` requires only `Required`. **Rejected:** a
fourth wall state for "accepted and never required", which is the same
semantics spelled twice, once in `keys.go` and once in the class table.

Provisional, owner ruling 1 open.

### R2: the six values reach the runner as parameters, and `internal/config` stays the sole environment reader (inventory Q2)

`config.Config` is four fields with no home for the three paths
(`config.go:56-65`), and the package doc calls itself "the single choke
point this package reads the environment through" (`config.go:67-71`).
A direct lookup in the subcommand would put a second environment reader
beside it.

**`Config` gains one pointer field**, non-nil only when the flag is set:

```go
// Shadow is rung 4's integration, resolved from the required-when-enabled
// rows of the shadow class. It is nil when DUBPLATE_SHADOW_ENABLED is
// unset, which is every adopter who has not opted in.
Shadow *ShadowIntegration

// ShadowIntegration is the shadow integration's three resolved paths.
type ShadowIntegration struct {
	ObserveDir     string
	OracleStateDir string
	ReviewDir      string
}
```

The floor and the cap stay thresholds, resolved in `cmd/dubplate-server`
through `config.ParseDuration` and `config.ParseInt`, because that is
what the registry is for and no threshold is a `Config` field today.
`internal/engine/shadow` sees only plain values, on `shadow.Config`
(R16). **Rejected:** a per-class accessor returning a name-to-value map,
which trades a typed struct for a map lookup at every call site and
gives `doctor` nothing it does not already have.

Provisional, owner ruling 1 open.

### R3: the matcher stamp is a `shadow.Config` field, and the refusal lives in the subcommand (inventory Q3)

This is the pass's sharpest under-specification. 4a owns `shadow
observe`'s refusal on an empty or `dev` stamp (`spec:886-892`) while 4b
owns the `-ldflags` change that would set it for real (`spec:1336-1348`),
`main.matcherCommit` does not exist today, and `Makefile:6-7` sets no
`-ldflags` at all.

**Three things, in this order.** 4a adds `var matcherCommit = "dev"` to
`cmd/dubplate-server/main.go` beside `version` (`main.go:29-31`), with a
doc comment naming 4b as the pass that makes a build path set it.
`shadow.Config` carries `MatcherCommit string`, and every artifact the
engine writes stamps it. **The refusal lives in
`cmd/dubplate-server/shadow.go`'s `observe` RunE**, before it calls
`shadow.Run`, because the spec assigns it to the verb rather than to the
engine and because that placement lets the engine's own end-to-end test
pass a real stamp without a build flag.

**Rejected:** an overridable package variable the test writes, which is
a global mutated by tests and reintroduces the ordering hazard R12
closes; and a `--matcher-commit` flag, which would let an operator
invent a stamp for a divergence nobody can re-evaluate.

### R4: a partial skip is a fourth named `unreconciled` reason (inventory Q4)

`resolveImported` marks a unit `OutcomeReview` with `Error: "left in
staging after import"` when beets took some but not all of its files
(`importer.go:1179-1199`). The spec's reconcile step maps `review` to a
clean oracle skip (`spec:506-511`), and recording a partly-imported
album as an oracle skip manufactures a divergence whenever Go applies,
which is exactly the fabricated-divergence failure S6 forbids. **A
`review` row carrying a non-empty `Error` is `unreconciled` with reason
`partial-skip`, never a clean skip.** The grouping rule already excludes
multi-group directories, so this is narrower than it first looks and is
still reachable for a single-group directory with one unreadable track.
**Rejected:** treating it as a clean skip and relying on the grouping
rule, which trades a counted honest gap for a silent lie.

### R5: the observation id is the capture digest (inventory Q5)

The record's `id` is `"shadow-" + <observation id>`, "deterministic from
the capture" (`spec:686`), and the capture's identity is a digest over
names, sizes and modification times (`spec:405-411`). Whether the id
**is** that digest is never said, and the choice decides whether a
re-uploaded album is a new observation.

```go
// Digest is the full hex SHA-256 over the album's file set: for each
// entry in lexical order, its name, its size in bytes, and its
// modification time in RFC 3339 nanosecond form, each NUL-separated,
// prefixed by the uploader and the basename. It is the capture's
// identity (contract S7), so a directory whose digest has not moved is
// not re-captured.
Digest string

// ID is "shadow-" plus Digest's first sixteen hex characters. It names
// the observation directory, the pending row and the resolved row.
ID string
```

Making the id the digest satisfies both halves at once: idempotence,
because a stable directory yields a stable id, and re-upload
distinction, because changed content yields a new observation.

**Two consequences of ingesting modification time, both stated rather
than discovered** (fold, B3). A digest is a function of the fixture only
when the fixture pins its mtimes, so **every fixture writer in this pass
calls `os.Chtimes` with a fixed instant** (the helper list carries the
rule once, and Tasks 6, 7, 9 and 14 inherit it); without it two fresh
`t.TempDir()` trees written at two wall-clock instants yield different
ids and Task 9's determinism test cannot pass. And **an id is a SHA-256
prefix, so id order is not capture order**: the pass loop matches in
`(uploader, basename)` capture order, never in lexical id order, and
carries the ordered slice on `PassResult` so the order is a value rather
than a filesystem property.
**Supersession is the reconcile step's**: when several stored,
unreconciled observations share one `(uploader, basename)`, the run
report's row resolves to the newest by capture time and the others are
marked superseded and counted. **Rejected:** a random or sequential id,
which cannot tell a re-uploaded album from its predecessor and turns the
idempotence check into a second index nobody maintains.

### R6: a report gap is detected from `StartedAt`, never from a filename or an mtime (inventory Q6)

Run report files are named `report-<runid>.json` and `RunID` is 8 random
bytes, "not derived from StartedAt, since several runs sharing a test's
frozen clock must not collide on one path" (`importer.go:158-161`, with
`newRunID` at `:456-462`; fold m3 corrects revision 1's `:139-145`,
which is `SessionID`'s comment). So the file names carry no order.

**`consumed-reports.json` records, per consumed report, its file name and
the `StartedAt` and `EndedAt` it decoded.** A pass computes the newest
consumed `EndedAt`, then reads every `report-*.json` present. A gap is
counted when the oldest report now present has a `StartedAt` later than
the newest consumed `EndedAt` **and** the ledger is non-empty, which
means at least one report was swept before this shadow consumed it.
**Rejected:** mtime ordering, which a restore, an rsync or a backup
rewrites, and which the incumbent's own sweep already uses for deletion
(`importer.go:533-548`), so a shadow reading the same signal would
inherit its blind spot.

### R7: the exit gate splits its three deferral branches across two runs, and the reason is mechanical (inventory Q7)

`Budget.Wait` reserves before it waits (`limiter.go:372-381`, with the
lost-update reason at `:361-371`), so `reserve` returns
`ErrBudgetExhausted` (`:383-391`) before the injected wait function is
ever called: once the cap is spent, every later album gets
`ErrBudgetExhausted` and the limiter is never consulted. In the other
order, `Limiter.Close` fails every pending **and future** waiter
(`:147-160`), so once closed every later album gets `ErrLimiterClosed`
and the cap is never reached. And `Budget.wait` is an unexported field
(`:340`), so no test outside `internal/engine/musicbrainz` can inject a
per-album failure.

**A single pass therefore cannot produce all three branches while
holding one `Budget`, which S14's `passes/` record requires it to.** The
end-to-end run covers `ErrUnavailable` and `ErrBudgetExhausted`;
`ErrLimiterClosed` is proven by a second focused test that runs one pass
against a pre-closed `musicbrainz.NewLimiter` and asserts every
observation defers with that reason. Task 13 states this in its own
report, and Task 15 drafts the exit-gate wording correction.
**Rejected:** loosening the store so a pass holds two budgets, which
would make `Budget.Spent()` unattributable and break the report's
honest sum.

### R8: `case.json` carries `Profile: "estate"` and a `YYYY-MM-DD` `EvalDate` (inventory Q8)

`cassette.Case` requires `Profile`, a profile name under
`corpus/profiles` (`cassette.go:44`), and `EvalDate` as a string
(`:45`). The committed corpus writes `eval_date` as `YYYY-MM-DD`
(verified against `corpus/cases/admit-duplicate-candidate/case.json`),
and `"estate"` is the profile whose reduction equals
`QuietImportTunables()` value for value today
(`corpus/profiles/estate.yaml:16-19`, inventory `:1056-1058`). Both are
corpus vocabulary rather than estate identity, so neither breaches the
universality rule.

**The written file is `LoadCase`-valid rather than produced by
`Case.Write`.** `Case.Write` (`cassette.go:237-252`) calls
`os.WriteFile` at `0o644` (`:248`), which is not an atomic write, and
the store's whole rule is that every file goes through
`atomicfile.WriteFile` (S13). The indent is not the reason: `Case.Write`
already marshals with two spaces (`cassette.go:242`), the same indent
`caseJSON` uses, so the two outputs differ only in how they reach disk
(fold, m6). So `cmd/dubplate-server` marshals the same
value with `json.MarshalIndent` and writes it atomically, and **the
honest check is that the written file round-trips through
`cassette.LoadCase`**, whose canonicalization rule (`cassette.go:15-25`,
`:168-170`) is exactly what a rung 15 consumer applies. Task 11 asserts
that round-trip. **Rejected:** calling `Case.Write` from the box binary,
which would put a non-atomic write inside the one directory the unit may
write.

Task 15 files "rename the profile the day the corpus grows a
`quiet-import` profile" as a note for lane E1.

### R9: basename resolution is a three-rule join, and the review tree matches on a suffix grammar (inventory Q9)

Backstop D resolves a skip "by the album's basename in a tag log's
`skip ` lines" (`spec:506-508`). The paths beets logs are **staging**
paths, since the call site passes `im.Cfg.Paths.StagingDir`
(`importer.go:1096`), and the basename survives the inbox-to-staging
move for a directory upload, because `stageUser` moves each inbox entry
to `filepath.Join(destDir, name)` with `name` the inbox entry's own name
(`importer.go:844-856`). So `filepath.Base` of a `skip ` path equals the
captured basename for the material the capture can see.

The review tree is not an equality. `moveInto` uses `uniqueDestPath`
(`_salvage/internal/review/review.go:64-73`, `:76-99`), which appends an
`-HHMMSS` suffix on a same-day collision and then `.2`, `.3`, and so on.
**The accepted grammar for basename `B` is `B`, `B-HHMMSS`, or
`B-HHMMSS.N` for `N >= 2`**, matched as `^B(-\d{6}(\.\d+)?)?$` with `B`
regexp-quoted. The layout is
`<review dir>/<YYYY-MM-DD>/<uploader>/<name>`, and the date is the
**run's** `StartedAt` (`importer.go:271`, `:499-505`), not the album's,
so the shadow reads every dated directory rather than guessing one.

**Rejected:** mirroring the incumbent's own `findUnit` plus
`deriveOrphanUnit` resolution (`importer.go:1147-1161`), which is a
second unledgered port of importer behavior inside a rung whose whole
product is the honesty of a comparison, the same reason the spec
declines mechanism B (`spec:450-455`).

### R10: `spec:1608-1609`'s "(ruling 7)" means ruling 6 (inventory Q10)

Confirmed stale. The deploy sitting is ruling 6 (`spec:342-368`);
revision 1's seventh ruling was retired rather than answered
(`spec:110-123`) and the cross-reference was not renumbered. Task 14
drafts the correction. Nothing in 4a depends on it.

### R11: "the first successful pass" is a pass that completes all three steps without a fatal error (inventory Q11)

`started-at` is written on the first successful pass (`spec:1099`,
`spec:1586-1590`), and the launch date is that date plus the floor
(`DA:525`), so the definition is load-bearing for a calendar figure the
close is required to record as first class.

**A pass is successful when capture, reconcile and match all complete
without a fatal error**, a fatal error being a store write failure, an
unreadable or uncreatable state directory, or a configuration error. **A
pass that captures nothing, reconciles nothing and observes nothing is
successful**, and that is deliberate: the clock must start at the deploy
(`DA:315`) rather than at the first upload, because the floor measures
the observation window and not the arrival rate. A deferred observation
does not make a pass unsuccessful; deferral is an ordinary branch (S5).
**Rejected:** "a pass that observed at least one album", which would
make the launch date depend on when Geoff next uploads and would silently
restart the clock's meaning at rung 15.

### R12: `config.Register` is called through `sync.OnceFunc` (inventory Q12)

`Register` is a bare append with no dedup (`thresholds.go:28-30`) over a
package-level slice (`:22`), the spec puts the call "from the root
command's own construction" (`spec:1080-1082`), and `newRootCmd()` is
called from `cmd/dubplate-server/main_test.go:14` as well as from
`main()`. A test binary that constructs the root command twice would
register the shadow thresholds twice and print duplicate doctor rows.
`doctor_test.go:98-108` builds its expectation from
`config.Thresholds()` itself, so this does not red the gate today and is
still a latent defect.

```go
// registerShadowThresholds is called once per process, from newRootCmd,
// because config.Register appends without dedup (thresholds.go:28-30)
// and newRootCmd runs more than once in a test binary.
var registerShadowThresholds = sync.OnceFunc(func() {
	config.Register(shadowFloorThreshold)
	config.Register(shadowBudgetThreshold)
})
```

**Rejected:** a dedup inside `Register`, which changes a rung 1 surface
this pass does not own the reason to change, and an `init()`, which
`go-conventions` forbids and which `thresholds.go:24-27` rejects by name.

### R13: the incumbent's report decode struct is engine-side, and it names its own retirement (inventory Q15)

`internal/engine/shadow` cannot import `_salvage`, which is outside the
module and invisible to the go tool. The reconcile logic is engine-side,
so the decode struct sits beside it in
`internal/engine/shadow/oraclereport.go`, with a doc comment citing
`_salvage/internal/importer/importer.go:134-172` as the shape it decodes
and naming rung 16 as the pass that retires it with the incumbent.
Fields are declared with the incumbent's own JSON tags, and the decoder
does **not** set `DisallowUnknownFields`, because the incumbent may add
a field and a shadow that refuses to read a newer report would go blind
rather than degrade. **Rejected:** putting the decode in
`cmd/dubplate-server` beside the other marshalling, which would split
the reconcile step across the arch wall for no gain, since the report is
not a harness type.

### R14: the two new `AdjudicationEntry` checks ship inline Go literal variants, on the package's own precedent (inventory Q14)

The spec asks for hand-authored `testdata` fixtures (`spec:846-848`).
The package's landed precedent for a **per-row** rule is the inline
table in `TestCheckAdjudications` (`adjudication_test.go:10-113`, base
entry at `:11-21`, shadow subject at `:23-24`, half-shadow at `:29-30`);
`testdata/` holds cross-ledger fixtures only (`testdata/defect_free`,
`testdata/oracle_bugs_probe_rule`, loaded at `:203-219`). **This plan
takes the package's precedent**, because a per-row rule proven by a
one-line struct variant is more legible than a JSON file a reader has to
open, and because splitting one test's fixtures across two mechanisms is
the drift the package has avoided so far. Task 15 drafts the spec
clarification. **Rejected:** two new `testdata` files, which would be the
first per-row fixtures in a directory whose whole contents are
cross-ledger.

### R15: the divergence classification is exhaustive over a closed vocabulary, and an unmapped error defers (inventory, S5 and S6)

`classify` takes the oracle signal, the Go proposal or error, and returns
exactly one of the four class constants or a deferral reason. **The
default arm is a deferral with reason `unclassified-error`, never a
verdict**, and the reason is recorded and counted. That is S6 applied to
the branch a future error type will arrive on: a matcher revision that
adds a sentinel gets a counted deferral rather than a fabricated
divergence, and the counter is what makes the omission visible.

### R16: `internal/engine/shadow`'s seam is stated, not implied (inventory Q16, Q12 of section (a))

The two architectural tests and the two transports together constrain the
signature enough that the plan states it rather than leaving it to an
implementer: no `internal/harness` import, no `internal/config` import,
no `time.Now` call, `now` and `evalDate` as parameters, the HTTP
transport supplied by the caller, and the limiter injectable so the
`ErrLimiterClosed` branch is reachable. Task 8 lands `shadow.Config` in
full; every field carries the reason it is a parameter.

### R17: the end-to-end run lives in `internal/engine/shadow`, and the marshalling test feeds `cmd` a literal (inventory Q17)

`internal/engine/shadow`'s test files may not import `internal/harness`
(`arch_test.go:30-44` walks `_test.go` files and skips only `testdata`),
and the two harness-shaped artifacts the exit gate asserts on are
marshalled in `cmd/dubplate-server`. The inventory names three ways out
and the plan picks one.

**The end-to-end `observe` run lives in `internal/engine/shadow`**,
against an `httptest.Server`, with an in-test FLAC builder, and it
asserts on parsed `Observation` structures rather than a golden.
**`cmd/dubplate-server`'s marshalling test feeds the writers a
hand-constructed `shadow.PassResult` literal**, which is the honest test
for that code: the marshalling takes engine-side results and writes two
files, so its test should hand it engine-side results directly and
assert the files round-trip through `cassette.LoadCase` and
`analysis.CheckAdjudications`. No FLAC bytes are needed in `cmd`, and no
builder is duplicated. **Rejected:** committing the eight albums' FLACs
as binary `testdata`, which the spec's "no binary fixture is vendored
twice" rule discourages and which would make the reader's mutation proof
a re-derivation rather than a byte edit.

### R18: the capture walks in lexical order, the match step follows capture order, and the fixture names are chosen against both (inventory Q18)

`os.ReadDir` returns entries sorted by filename, which is what
`listContributorDirs` already relies on (`importer.go:1238-1254`). The
capture walks `<observe dir>` in that order, then each uploader's
entries in that order.

**The match step follows the same order, and that is a fold rather than
a restatement** (fold, B3). Revision 1 had the pass loop match "in
lexical id order", and an id is `"shadow-"` plus a SHA-256 prefix
(R5), which bears no relation to the basename order the fixture chooses.
**The budget-exhaustion fixture's correctness depends on which album the
pass reaches first**, so **Task 14's** fixture names (fold m18 corrects
revision 1's "Task 13's"; Task 13 is `reevaluate`) carry an explicit
numeric prefix, the pass loop matches in `(uploader, basename)` order,
and `PassResult` carries that ordered slice so the order is stated
rather than left to alphabetical accident or to a digest's sort.
`Store.Observations` keeps its own lexical id order, which is the
store's read order and not the pass's match order.

### R19: an archive upload is invisible to the capture, and that is a declared gap (inventory Q19)

`stageArchive` extracts a `.zip` into a scratch directory and moves each
extracted top-level entry under a collision-safe name
(`importer.go:861-888`), then removes the archive from the inbox
(`:887`). The shadow's capture walks album directories, so an archive in
the inbox is never captured, and the album that eventually appears in
the run report has a basename the shadow never saw. **This is a declared
coverage gap for 4a's close, not an archive branch.** Adding one would
put a second unledgered port of importer behavior inside the rung, the
same reason S9 declines porting `group_albums`. The gap is real and
bounded: the affected album is simply never observed, and the report's
`unreconciled` count does not grow, because there is no fixture to fail
to resolve.

### R20: the exit gate also proves the two exclusion counters and each `unreconciled` reason (inventory Q20)

The spec's exit gate names one agreement, one divergence of each class,
one excluded multi-album directory, and one deferred observation of each
error branch (`spec:1616`). It does not name the gate-rejected exclusion
or the four `unreconciled` reasons, and both print in the report.
**A counter reading zero forever is indistinguishable from a broken
one**, which is this rung's own named worst outcome applied to a row
rather than to a pass. Task 13's fixture set therefore adds one
gate-rejected directory and one fixture per `unreconciled` reason, at a
cost of five directories and three report rows. **Rejected:** proving
them by unit test alone, which would leave the report's own summation
untested on exactly the rows nobody watches.

### R21: the report states what the box cannot know

The report's `divergences ... by disposition` and `false confident
matches` rows (`spec:1108-1109`) read owner dispositions, which live in
`corpus/analysis/adjudications.json` on the workstation and never on the
box. **The box's own report prints what the store knows**: divergences
by class, rows still staged under `pending/`, rows retired under
`resolved/`, and the `go-applies-oracle-skipped` count. The disposition
join is the launch gate's, run on the workstation against the committed
corpus, and the report says so in its own row labels rather than
printing a zero that reads as a measurement. Task 12 lands the wording;
Task 15 drafts the spec clarification.

### R22: an `imported` row with no resolvable release id is `unreconciled`, never an apply (fold, M-g)

`readFLAC` gives an absent tag the field's Go zero value, which is
`ProjectItemFields`' own typed-default rule
(`internal/engine/matcher/info.go:178-211`; `stringField` returns `""`
for an absent key at `:215-218`). So a library file the reader opens
successfully but that carries no `MUSICBRAINZ_ALBUMID` yields
`MBAlbumID == ""`. Revision 1 guarded only an empty `LibraryPaths`, so
that file produced `OracleSignal{Signal: "apply", ReleaseMBID: ""}`,
and `classify` read `R != G` against any Go apply and wrote
`ClassGoAppliesDifferentRelease` plus a pending row naming the Go
release. That is a divergence against a release the oracle never named,
in the one artifact contract S6 says must be honest, and it consumes
exactly the owner adjudication ruling B's stamp exists to save. The case
is reachable rather than theoretical: `LibraryPaths` is a positional
correlation whose count mismatch leaves later units empty
(`importer.go:1119-1136`), so the first named path is not guaranteed to
be the album the row means.

**The second `unreconciled` reason therefore widens** rather than a
fifth being added, so the report keeps four rows and the exit gate keeps
four fixtures: `UnreconciledNoReleaseID = "imported-with-no-release-id"`
covers an empty `LibraryPaths` and an unreadable or absent
`MUSICBRAINZ_ALBUMID` alike, and its constant's comment names both.
Task 7 carries a firing fixture for each half, because a reason with one
of its two halves untested is R20's own anti-pattern narrowed.
**Rejected:** a fifth reason with its own end-to-end fixture directory,
which grows the exit gate by a fourteenth album to distinguish two
shapes that produce the identical honest outcome.

## Test helpers named in this plan

Every test below calls helpers that live in the same package's `_test.go`
file. They are named here once so no code block reads as a reference to
something no task defines.

- `internal/atomicfile`: none; the salvage test file is re-landed as it
  stands.
- `internal/engine/matcher`: none added. The `QuietImportTunables` test
  is a plain field comparison against `DefaultTunables()`.
- `internal/config`: `withRegistry(t *testing.T)` (exists,
  `internal/config/thresholds_test.go:11-16`; it saves and restores
  `registeredThresholds` through `t.Cleanup`, which is the semantics
  this plan wants). Revision 1 named a `resetThresholds` that does not
  exist and an `envWith` builder no task uses; both are dropped (fold,
  m5 and m13), and Task 3's tests build their environments with
  `append(slices.Clone(base), tt.extra...)` as written.
- `internal/engine/shadow`: `evalDate` and `passNow` (package-level
  fixed `time.Time` literals the tests share, never a clock read);
  `buildFLAC(t *testing.T, tags []string, samples uint64, rate uint32) []byte`
  (Task 4, the in-test FLAC builder; tags are ordered `KEY=value`
  strings, because Vorbis comment order is load-bearing for the
  first-value-wins rule); `writeAlbum(t, dir, uploader,
  basename string, tracks [][]string)` (Task 6, writes one inbox album
  from `buildFLAC`, one ordered tag list per track); `writeRunReport(t, stateDir string, r
  oracleRunReport)` (Task 7); `writeTagLog(t, stateDir string, skips
  []string)` (Task 7); `writeReviewDir(t, reviewDir, date, uploader,
  name string)` (Task 7); `releaseServer(t *testing.T, bodies
  map[string][]byte, status map[string]int) *httptest.Server` (Task 8,
  on `internal/engine/tag/tag_test.go:25-66`'s precedent, sending
  `Retry-After: 0` beside any retryable status so a 503 row collapses
  its backoff instead of spending 31 real seconds); `releaseBody(albumID,
  albumTitle, artist string, tracks []string) []byte` (Task 8, a minimal
  WS/2 release document **adapted from
  `internal/engine/tag/testdata/early-return/`'s committed body**, which
  `internal/engine/tag/tag_test.go:141-158` already proves scores
  `matcher.RecStrong` through the identical `tag.Album` path; fold M-f);
  `searchBody(ids ...string) []byte` (Task 8). Task 5 adds
  `openTempStore(t)`; Task 4 adds `writeTemp(t, data)`;
  Task 6 adds `corrupt(t, path)` (which overwrites a track with a
  truncated container, the shape a file renamed out from under the walk
  presents to `readFLAC`); Task 14 adds `buildInbox(t) inboxTree` (the
  thirteen-directory fixture with its reports, tag log, review tree,
  ledger and budget cap) and `indexByBasename(t, obs)`.
- `internal/harness/analysis`: `base()` (exists, the table's base entry
  at `adjudication_test.go:11-21`).
- `cmd/dubplate-server`: `passResultFixture(t *testing.T) *shadow.PassResult`
  (Task 11, the hand-constructed literal Resolution R17 names),
  `completeRow()` and `rowWithout(t, field)` (Task 11, the pending-row
  literals), and `writeResponse(t, dir, name)` (Task 11, a body file so
  `cassette.LoadCase`'s readability check at `cassette.go:225` has
  something to read).

**Every helper that writes a file into a fixture tree pins that file's
modification time** (fold, B3). `writeAlbum`, the library-file writer
Task 7 uses, and `buildInbox` each call
`os.Chtimes(path, fixedModTime, fixedModTime)` with a package-level
`fixedModTime` literal beside `evalDate` and `passNow`. The capture
digest ingests modification time (Resolution R5), so without this a
digest is a function of when a test ran and no id is reproducible across
two temp trees. Task 6's touched-file mutation moves that literal rather
than sleeping.

`internal/engine/tag`'s `taggingServer`, `statusServer` and `clientFor`
are read as precedent and never imported; they are unexported in their
own package.

---

## Task 0: The staleness pre-task

Model: `sonnet`, **fresh context**, dispatched before any implementer.
Reads this plan and the rung 4 spec against the tree as shipped. **Its
output is blockers with evidence, not fixes**; it writes no code and
edits no file. The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's citations were verified at
`54e534b0e83a305fe02c58cc41dd05ec7ee573a0`. A citation that has moved is
a finding for this task, not a surprise for an implementer.

The checklist is the spec's own (`spec:1670-1712`), adapted to 4a. **The
first three are blockers: a "no" stops the launch rather than producing
a note.**

1. **Has rung 1b landed?** *(Expected: no.
   `internal/config/keys.go:19-25` still carries `ownedWall` as a slice
   literal and `reservedClasses` as a table at `:57-142`.)* **If yes,
   the integration class and the duration threshold register through
   that registry and never by editing `keys.go`, and ruling 1's option
   (a) is re-expressed in the registry's own terms.** That is a blocker
   for Task 3 and rewrites it.
2. **Is the lane registry consistent with where this pass intends to
   run, and is the primary tree free?** `docs/STATUS.md`'s lane registry
   is the live-executor check, never `pgrep -f Projects/dubplate`, which
   matches every lane once a worktree exists. *(Expected: lane-1 reads
   "rung 4 spec landed ... next action: pass 4a plan prep, then 4a
   execution"; flipping it to a 4a-execution entry is a conductor edit
   owed before Task 1 dispatches. lane-2 BLOCKED; doc lane IDLE.)*
   Report any live pid in this tree, or a lane-1 entry naming another
   workflow, as a blocker.
3. **Is `.gitleaksignore`'s single line still
   `internal/config/keys_test.go:generic-api-key:<N>`, and is line `N`
   still the row it names?** *(Expected: `N` is 66, the first row of
   `TestReservedClassOf`'s table literal, whose rows run `:66-72` inside
   a table declared `:60-73`; the file is 123 lines.)* **Report the
   number and the row's identity without quoting the fixture literal**,
   because gitleaks scans the tracked worktree with `--no-git` and would
   flag the quotation. A moved pin is a blocker for Task 3.
4. **Have the contested-surface rows been written for every surface this
   pass touches?** *(Expected: all nine present in
   `docs/superpowers/specs/2026-09-03-parallel-build-approach.md:168-184`,
   at rows `:169`, `:171`, `:177`, `:178`, `:179`, `:180`, `:181`,
   `:182`, `:183`, plus 4b's at `:184`.)* **Check the row paths, not the
   stamp count**: a grep for `conductor ruling, 2026-09-05` returns
   **13**, not nine, because rows `:168`, `:172`, `:176` and `:177`
   carry the same stamp from 3d's own fold (fold, m15). Report the count
   as 13 and the nine rows by path. Also confirm that
   `internal/harness/conform`'s existing row (`PA:167`) and
   `internal/engine/matcher`'s (`PA:171`) cover 4a's two small edits
   there. None is a launch blocker.
5. **Does `config/beets-config.yaml` still set `group_albums: yes`?**
   *(Expected: yes, `:23`, with the estate's reason at `:19-22`.)* The
   capture's grouping rule and its excluded-directory count are written
   against it.
6. **Are `defaultReportRetention` still 30 and `defaultRetentionDays`
   still 60** (`_salvage/internal/config/config.go:172`, `:155`)? The
   reconcile step's report-gap handling is written against both.
7. **Does the incumbent's quiescence default still sit at ten minutes**
   (`_salvage/internal/config/config.go:152`, applied at
   `importer.go:790-798`) **and its import timer at `*:0/15` with
   `RandomizedDelaySec=300`** (`systemd/dubplate-import.timer:9-11`)?
   Informational for 4a and load-bearing for 4b's timer.
8. **Is `tag.Album`'s signature unchanged, and are `Proposal`'s five
   fields unchanged?** *(Expected: `internal/engine/tag/tag.go:68-69`
   and `:23-29`, matching contract S1.)*
9. **Are `analysis.AdjudicationEntry`, `AdjudicationSubject` and
   `checkAdjudicationEntry` unchanged since 3d, and does
   `corpus/analysis/adjudications.json` still carry zero shadow-subject
   rows?** *(Expected: subject three fields at `adjudication.go:35-39`,
   entry ten plus `DefectFree` at `:46-67`, the pair rule at `:161-174`,
   the five non-empty prose checks at `:176-190` with `Citation` not
   among them, the agent rule at `:198-200`; 21 rows, 0 with an `album`
   subject.)* Count the rows mechanically; read no row body into context
   beyond its subject shape.
10. **Are `musicbrainz.ClassBackground`, `Limiter.Budget`, `NewLimiter`,
    `Close`, `Unmetered`, `DefaultInterval`, `Budget.Spent`,
    `Budget.Name`, `ErrUnavailable`, `ErrBudgetExhausted`,
    `ErrLimiterClosed` and `ClientConfig` unchanged?** *(Expected:
    `limiter.go:25`, `:139-145`, `:110-115`, `:147-160`, `:354-359`,
    `:62`, `:400-404`, `:408`; `client.go:84`, `limiter.go:48`, `:53`;
    `client.go:93-99`.)* **Confirm two facts the plan's tests rest on**:
    a `Budget` limit of zero or less means no cap (`limiter.go:130-133`,
    `:386`), and `ClientConfig.Timeout` is ignored entirely when `HTTP`
    is non-nil (`client.go:137-144`).
11. **Has `internal/config` changed shape?** *(Expected: `Config` four
    fields at `config.go:56-65`; `rejectNamespace` `:133-152`; `Load`
    `:159-175`; `Threshold` three string fields at `thresholds.go:13-17`
    with `ParseInt` the only parser at `:43-60`; `Register` a bare
    append at `:28-30`; zero registered thresholds in production.)*
    Confirm that `internal/config/thresholds_test.go` still compares
    `Threshold` values with `!=`, which is why `Kind` must be a
    comparable type. *(Expected: the one comparison is `:34`,
    `got[0] != first || got[1] != second`; `:101` compares an `int`.
    Revision 1 cited `:35` and `:45`, neither of which compares
    anything; fold, m4.)* Confirm too that the registry helper is
    `withRegistry(t *testing.T)` at `:11-16` and that no
    `resetThresholds` exists (fold, m5).
12. **Have `internal/atomicfile`, `internal/engine/shadow`,
    `cmd/dubplate-server/shadow.go` or `main.matcherCommit` appeared?**
    *(Expected: no on all four.)* And is
    `_salvage/internal/atomicfile/atomicfile.go` still 42 lines with one
    exported function at `:15-41`, and its test 95 lines with the
    four-case table at `:9-70`, the leftover-temp assertion at `:72-87`
    and the missing-directory case at `:89-95`?
13. **Do `docs/HISTORY.md` and `docs/STATUS.md` still describe the pass
    state this plan assumes?** *(Expected: 3d closed; the 3d handoff at
    `docs/HISTORY.md:267-302` recording the wire pace as
    `DefaultInterval`, which the conductor's ruling amends; twenty-one
    declared corpus-coverage gaps carried; the window cap and the
    2026-09-07 09:00 flag route at `docs/STATUS.md:24-29`.)*
14. **Is `go test ./...`'s wall time still what this plan sizes gate runs
    against?** Run it once at `-count=1` and report the total and the
    `internal/harness/conform` figure. *(3d's close measured `go test
    ./...` at 2m06s, `internal/harness/conform` at 123.7s, and
    `bash scripts/check.sh` at about 3.5 minutes.)*

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
beyond item 14's timing measurement; the task writes nothing.

---

## Task 1: `internal/atomicfile`, re-landed from salvage on merit

Model: `sonnet`. Depends on Task 0.

Everything durable in the store writes through this package, so it is
born first and nothing else in the pass may call `os.WriteFile` on a
store path.

**Files:**
- Create: `internal/atomicfile/atomicfile.go`
- Create: `internal/atomicfile/atomicfile_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:

```go
// internal/atomicfile/atomicfile.go

// Package atomicfile writes files atomically, so a crash or a concurrent
// reader never observes a partial write.
package atomicfile

// WriteFile writes data to path atomically: it creates a temp file in
// path's directory, writes and syncs data to it, then renames it over
// path. The rename is atomic on the same filesystem, so readers see
// either the old content or the new content, never a partial file.
func WriteFile(path string, data []byte, perm os.FileMode) error
```

**The merit bar, stated so the admission is a judgment rather than a
copy.** Salvage tier is read-only adoption material admitted on merit
(CLAUDE.md), and this is the file that passes: 42 lines, one exported
function, a temp file created in the **target's own directory** so a
single-directory `ReadWritePaths` grant suffices
(`_salvage/internal/atomicfile/atomicfile.go:16-17`), a
`defer os.Remove(tmpPath)` cleanup at `:22`, and a `Chmod` **after**
`Close` and before `Rename` at `:35-37`, which is what makes the
permission survive the process umask. `go-conventions`
(SKILL.md:298-310) names this exact signature as the canonical one. The
implementer reads the salvage file, confirms each of those five
properties against the modern stdlib, and re-lands it. **Nothing is
rewritten for style.** A modern-idiom pass touches nothing beyond
confirming `os.CreateTemp` usage.

- [ ] **Step 1: Write the failing test**

Port `_salvage/internal/atomicfile/atomicfile_test.go` whole: the
four-case table over create, overwrite-shorter, permissions and empty
content (`:9-70`), the leftover-temp-file assertion (`:72-87`), and the
missing-directory failure case (`:89-95`). Add one case the salvage test
does not carry, because it is the property the store depends on and the
salvage suite proves only indirectly:

**Revision 1's form of this test was vacuous and is replaced** (fold,
B6). It wrote `"first"` to `dir/state.json`, then failed a write at
`dir/blocked/state.json`, then asserted `dir/state.json` still read
`"first"`. Those are two different paths: `WriteFile` creates its temp
in `filepath.Dir(path)` and renames within it
(`_salvage/internal/atomicfile/atomicfile.go:16-17`, `:38`), so
`dir/state.json` is never opened on any code path and the assertion
holds for every possible implementation, including one that writes
nothing. The failing write has to target the path whose content is being
asserted:

```go
func TestWriteFileLeavesTheOldContentWhenTheWriteFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions, so the write cannot be made to fail this way")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := WriteFile(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// os.CreateTemp writes into path's own directory
	// (atomicfile.go:16-17), so removing that directory's write bit
	// fails the write before the rename and leaves path untouched. That
	// is the crash-safety property the shadow store rests on: a reader
	// sees the old content or the new content, never neither.
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	if err := WriteFile(path, []byte("second"), 0o644); err == nil {
		t.Fatal("WriteFile into an unwritable directory returned nil; want an error")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "first" {
		t.Errorf("original content = %q, want %q", got, "first")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/atomicfile/ -count=1 -v`
Expected: FAIL, the package does not exist.

- [ ] **Step 3: Land the package**

Copy `_salvage/internal/atomicfile/atomicfile.go` to
`internal/atomicfile/atomicfile.go` unchanged, confirming the five
properties above. Do not import it from `_salvage`; the salvage tree is
outside the module and invisible to the go tool.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/atomicfile/ -count=1 -v`
Expected: PASS, every case.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/atomicfile && git commit`

**Acceptance criteria:**

- `WriteFile` has exactly the signature above and no second exported
  symbol.
- The temp file is created in the target's own directory, and the test
  proving no temp file survives a successful write passes.
- The `Chmod` runs after `Close` and before `Rename`, and the
  permissions case proves it against a non-zero umask.
- `go.mod` is unchanged; `make tidy-check` is green.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Move `os.CreateTemp` from `dir` to `os.TempDir()` →
   `TestWriteFileLeavesNoTempFile` still passes, and
   `TestWriteFileLeavesTheOldContentWhenTheWriteFails` **fails**,
   because a temp created outside `dir` is unaffected by `dir`'s write
   bit and the write succeeds where an error is wanted. The
   cross-filesystem half of the property, that `os.Rename` takes EXDEV
   when temp and target sit on different filesystems, is **a separate
   claim this single-filesystem workstation cannot discriminate**; that
   half is held by the doc comment and by 4b's `ReadWritePaths`
   argument, and the task says so rather than reporting a
   discrimination it did not observe (fold, B6).
2. Drop the `Chmod` → the permissions case fails, reporting the umasked
   mode rather than 0644.
3. Drop the `Sync` → nothing in the suite fails. **Reported as a
   deliberately unfired mutation**, and recorded at the call site: the
   durability the sync buys is observable only across a power loss, and
   naming it as unfired is more honest than inventing a test that
   cannot see it.

**Deliverable count: 1** (the package plus its test). Gate green; commit.
**Report actual token spend and the three mutations' outcomes, including
mutation 1's environment dependence.**

---

## Task 2: `matcher.QuietImportTunables`, the drift guard, and the retired provisional mark

Model: `sonnet`. Depends on Task 0. Independent of Task 1.

Three small edits across two packages, together because they are one
seam: the engine gains the estate's import tunables as a named
constructor, and the harness gains the assertion that keeps its own
profile-driven copy in step.

**Files:**
- Create: `internal/engine/matcher/quiet.go`
- Modify: `internal/engine/matcher/tunables_test.go` (append)
- Modify: `internal/harness/conform/defaults_test.go` (append the drift
  guard)
- Modify: `internal/harness/conform/case.go:152-153` (the `GradeMode`
  doc comment's last clause)

**Interfaces:**
- Consumes: `matcher.DefaultTunables` (`tunables.go:266`).
- Produces:

```go
// internal/engine/matcher/quiet.go

// QuietImportTunables returns DefaultTunables with the four assignments
// an unattended quiet import makes, which is the configuration rung 4's
// shadow observes under and rung 9's pipeline will import under.
//
// It lives here rather than in the harness because internal/engine may
// import no harness package (TestEngineNeverImportsHarness,
// internal/harness/conform/arch_test.go:141-146) and conform's
// gradeTunables (defaults.go:72-79) is on the wrong side of that wall.
// It adds no new surface concept: Quiet, QuietFallback, NoneRecAction
// and Timid are all existing Tunables fields.
//
// Two of the four already ship at these values (Timid at tunables.go:312,
// QuietFallback at :316) and are assigned anyway, so the set reads as one
// configuration rather than as a diff a reader has to compute.
func QuietImportTunables() Tunables {
	tun := DefaultTunables()
	tun.Quiet = true          // the -q flag the incumbent passes beet import
	tun.QuietFallback = "skip" // config/beets-config.yaml:16
	tun.NoneRecAction = "skip" // config/beets-config.yaml:17
	tun.Timid = false          // config/beets-config.yaml:18
	return tun
}
```

- [ ] **Step 1: Write the failing tests**

In `internal/engine/matcher/tunables_test.go`:

```go
func TestQuietImportTunablesAppliesTheFourImportAssignments(t *testing.T) {
	def := DefaultTunables()
	got := QuietImportTunables()

	if !got.Quiet {
		t.Error("Quiet = false, want true")
	}
	if got.QuietFallback != "skip" {
		t.Errorf("QuietFallback = %q, want %q", got.QuietFallback, "skip")
	}
	if got.NoneRecAction != "skip" {
		t.Errorf("NoneRecAction = %q, want %q", got.NoneRecAction, "skip")
	}
	if got.Timid {
		t.Error("Timid = true, want false")
	}

	// Nothing else moves. The whole distance and threshold surface runs
	// at beets' shipped defaults, because the estate's beets config sets
	// no match: block and no musicbrainz: block
	// (internal/harness/conform/defaults.go:16-18).
	got.Quiet, got.QuietFallback, got.NoneRecAction, got.Timid =
		def.Quiet, def.QuietFallback, def.NoneRecAction, def.Timid
	if !reflect.DeepEqual(got, def) {
		t.Error("QuietImportTunables moved a field other than the four import assignments")
	}
}
```

In `internal/harness/conform/defaults_test.go`, the drift guard. The
harness may read the engine, which is the permitted direction:

```go
func TestQuietImportTunablesMatchesTheEstateProfile(t *testing.T) {
	// Two definitions of one configuration is the drift
	// internal/engine/matcher/info.go:148-150 warns about, and they are
	// both live: gradeTunables must stay profile-driven, because the
	// stock profile overlays nothing (translate.go:205-206), while
	// matcher.QuietImportTunables is the engine-side constructor rung 4's
	// shadow needs and may not reach across the arch wall for. This
	// one-line assertion is what fires the day either moves.
	want := gradeTunables(t, "estate")
	if got := matcher.QuietImportTunables(); !reflect.DeepEqual(got, want) {
		t.Error("matcher.QuietImportTunables and gradeTunables(\"estate\") have drifted")
	}
}
```

The implementer reads `gradeTunables`' real signature
(`internal/harness/conform/defaults.go:72-79`) and its profile-merge
path (`translate.go:203-226`) and adapts the call shape; the assertion
is what matters, not the helper's arity.

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/matcher/ ./internal/harness/conform/ -run 'QuietImport' -count=1 -v`
Expected: FAIL, `QuietImportTunables` undefined.

- [ ] **Step 3: Write `quiet.go`**

As above.

- [ ] **Step 4: Retire the provisional mark**

`internal/harness/conform/case.go:152-153` ends `GradeMode`'s doc
comment with "Revision 2, fold B2; provisional, conductor ruling owed".
The conductor ruled it (`spec:59`, `docs/HISTORY.md:426-428`), so
**delete that clause and name the ruling in its place**, in one
sentence: the ruling is recorded at the rung 4 spec's conductor fold and
at 3d's HISTORY entry, and rung 4's second scoring mode tracks this
design. Change nothing else in the paragraph; the flow-change argument
it states is what contract S16 quotes.

- [ ] **Step 5: Run the tests, the gate, and commit**

Run: `go test ./internal/engine/matcher/ ./internal/harness/conform/ -count=1`
Then: `bash scripts/check.sh`
Then: `git add internal/engine/matcher/quiet.go internal/engine/matcher/tunables_test.go internal/harness/conform/defaults_test.go internal/harness/conform/case.go && git commit`

**Acceptance criteria:**

- `QuietImportTunables` moves exactly four fields and nothing else,
  proven by the `reflect.DeepEqual` restore.
- The drift guard passes today and is a one-line failure the day either
  definition moves.
- `case.go`'s provisional clause is gone and the ruling is named; the
  paragraph's argument is otherwise byte-identical.
- `TestEngineNeverImportsHarness` still passes with the new engine file
  inside its walk.
- `bash scripts/check.sh` green, including tier 1's and tier 2's sweeps.

**Named mutations this task reports:**

1. Set `NoneRecAction = "ask"` (the shipped default) →
   `TestQuietImportTunablesAppliesTheFourImportAssignments` fails on
   that field, **and so does the drift guard**, which is the point: the
   guard is the cheaper detector because it fires on either side moving.
   Inventory `:1049-1053` records `none_rec_action` as the one of the
   three `withImportKeys` overlays whose value actually differs from the
   shipped default, so this is the only one of the four whose mutation
   the guard can see independently.
2. Return `DefaultTunables()` unchanged → both tests fail, the unit test
   on `Quiet` first.
3. Change one `DistanceWeights` entry inside `QuietImportTunables` → the
   `reflect.DeepEqual` restore fails, which is what proves the "nothing
   else moves" half is not vacuous.

**Deliverable count: 3** (`QuietImportTunables`, the conform drift guard,
the retired provisional mark). Gate green; commit. **Report actual token
spend and the three mutations' outcomes.**

---

## Task 3: The config integration class, the duration threshold, the six rows, and the `Register` call site

Model: `sonnet`. Depends on Task 0.

**Provisional, owner ruling 1 open.** The whole task is ruling 1's option
(a). If the ruling returns (b) or (c), the fold is the row in "The six
open owner rulings" above.

**Files:**
- Modify: `internal/config/keys.go` (the `IntegrationClass` type, the
  table, the lookup)
- Modify: `internal/config/config.go` (`Config.Shadow`, the flag
  resolution, `rejectNamespace`, `Load`)
- Modify: `internal/config/thresholds.go` (`ThresholdKind`,
  `Threshold.Kind`, `ParseDuration`)
- Modify: `internal/config/keys_test.go` (**append at the end of the
  file only**, below the pinned line; see Global Constraints)
- Modify: `internal/config/config_test.go`,
  `internal/config/thresholds_test.go`
- Create: `cmd/dubplate-server/shadow.go` (the two `Threshold` values and
  the `sync.OnceFunc` registration; the subcommand itself is Task 12's)
- Modify: `cmd/dubplate-server/main.go` (call the registration from
  `newRootCmd`)
- Modify: `cmd/dubplate-server/doctor.go` (`thresholdStatus`'s kind
  switch, and the integration-class rows)
- Modify: `cmd/dubplate-server/doctor_test.go`

**Interfaces:**
- Consumes: `config.Threshold`, `config.Register`, `config.Thresholds`,
  `config.Load`, `config.OwnedWall`, `config.ReservedClasses`.
- Produces:

```go
// internal/config/keys.go

// IntegrationClass is one named integration the wall accepts only when
// its Flag is set: an adopter-optional capability rather than an owned
// one (build spec U4). Setting any of its keys with the flag unset is a
// load error naming the flag, which is what keeps a half-configured
// integration from running unattended and observing nothing.
//
// Required and Optional differ in what Load does with them, not in what
// rejectNamespace accepts. Load requires every Required key when the
// flag is set. An Optional key is accepted and never required, because
// it is a registered threshold whose Default stands when it is unset
// (thresholds.go:37-42).
type IntegrationClass struct {
	Name     string
	Rung     int
	Flag     string
	Required []string
	Optional []string
}

// IntegrationClasses returns the integration classes in table order.
func IntegrationClasses() []IntegrationClass

// Enabled reports whether c's flag resolves true in environ. An unset or
// empty flag is off. A value strconv.ParseBool refuses is reported as an
// error rather than silently read as off, because a typo that disables
// an unattended observer is the failure this class exists to prevent.
func (c IntegrationClass) Enabled(environ []string) (bool, error)

// internal/config/config.go

// ShadowIntegration is the shadow integration's resolved paths, the
// three required-when-enabled rows of the shadow class. The floor and
// the request cap are registered thresholds rather than fields here,
// because both carry defaults and Config carries no defaulted value.
type ShadowIntegration struct {
	ObserveDir     string
	OracleStateDir string
	ReviewDir      string
}

// internal/config/thresholds.go

// ThresholdKind names the parser a threshold's value is read through.
// KindInt is the zero value, so a Threshold declared before this field
// existed keeps its meaning.
type ThresholdKind int

const (
	KindInt ThresholdKind = iota
	KindDuration
)

// ParseDuration resolves t as a duration against environ, in ParseInt's
// own shape (thresholds.go:43-60): t.Default's parse when t.Key is
// unset, environ's value when set, an error naming t.Key and t.Doc on a
// bad value, and an error naming t.Key alone on a bad default, which is
// a registration bug rather than an environment one.
func ParseDuration(environ []string, t Threshold) (time.Duration, error)
```

`Threshold` gains one field, placed **after** `Key` so the existing
literal order stays readable:

```go
type Threshold struct {
	Key     string
	Kind    ThresholdKind
	Default string
	Doc     string
}
```

`ThresholdKind` is an `int`, so `internal/config/thresholds_test.go`'s
`!=` comparisons at `:35` and `:45` keep compiling.

**The table**, per `spec:1436-1443`:

```go
// integrationClasses is every adopter-optional capability the wall
// accepts behind a flag. Rung 1b's registry absorbs this table as data
// rather than re-deriving it (rung 4 spec, risk 6).
var integrationClasses = []IntegrationClass{
	{
		Name: "shadow",
		Rung: 4,
		Flag: "DUBPLATE_SHADOW_ENABLED",
		Required: []string{
			"DUBPLATE_SHADOW_OBSERVE_DIR",     // the inbox the incumbent consumes
			"DUBPLATE_SHADOW_ORACLE_STATE_DIR", // where the incumbent's run reports and tag logs live
			"DUBPLATE_SHADOW_REVIEW_DIR",      // where a skipped album is filed
		},
		Optional: []string{
			"DUBPLATE_SHADOW_FLOOR",  // threshold, duration, default 720h
			"DUBPLATE_SHADOW_BUDGET", // threshold, int, default 60
		},
	},
}
```

`DUBPLATE_LIBRARY_ROOT` is reused rather than re-minted: it is an owned
row (`keys.go:20`) and is where an applied album's files land
(`spec:1445-1447`). `DUBPLATE_BEETS_DB` stays retired and unamended
(`keys.go:178`), because the shadow never opens the beets database.
`DUBPLATE_SHADOW_REVIEW_DIR` is deliberately distinct from
`DUBPLATE_REVIEW_DIR`, which is reserved for rung 13 at `keys.go:130`,
and both appear in one wall.

**The two `Threshold` values and the registration**, in
`cmd/dubplate-server/shadow.go`:

```go
// shadowFloorThreshold is the evidence floor rung 15 measures the
// shadow window against: a configurable duration with a 30-day default
// (build spec ruling 1). Provisional, owner ruling 4 open.
var shadowFloorThreshold = config.Threshold{
	Key:     "DUBPLATE_SHADOW_FLOOR",
	Kind:    config.KindDuration,
	Default: "720h",
	Doc:     "a Go duration, for example 720h",
}

// shadowBudgetThreshold caps one observe pass's MusicBrainz spend. An
// album costs one lookup on the ID-consensus early return
// (internal/engine/tag/tag.go:96-118) and, on the search path, one
// search plus SearchLimit lookups plus one pseudo-release fetch per
// candidate (internal/engine/musicbrainz/client.go:320-333), so the
// real per-album bound is eleven rather than seven. Sixty tokens at two
// seconds a token bounds a pass at two minutes.
// Provisional, owner ruling 5 open.
var shadowBudgetThreshold = config.Threshold{
	Key:     "DUBPLATE_SHADOW_BUDGET",
	Kind:    config.KindInt,
	Default: "60",
	Doc:     "an integer request count per pass",
}

// registerShadowThresholds runs once per process, from newRootCmd
// (Resolution R12): config.Register appends without dedup
// (internal/config/thresholds.go:28-30) and newRootCmd runs more than
// once in a test binary.
var registerShadowThresholds = sync.OnceFunc(func() {
	config.Register(shadowFloorThreshold)
	config.Register(shadowBudgetThreshold)
})
```

`newRootCmd` calls `registerShadowThresholds()` as its first statement,
before the `AddCommand` block, so registration order stays visible in
the caller's own code (`thresholds.go:24-27`).

- [ ] **Step 1: Write the failing tests**

`internal/config/config_test.go`, the acceptance and requirement halves:

```go
func TestLoadShadowIntegration(t *testing.T) {
	base := []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/library",
		"DUBPLATE_STATE_DIR=/var/lib/x",
		"DUBPLATE_SITE_NAME=Example",
		"DUBPLATE_ADMIN_EMAIL=admin@example.test",
	}
	shadowPaths := []string{
		"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox",
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR=/var/lib/incumbent",
		"DUBPLATE_SHADOW_REVIEW_DIR=/srv/review",
	}

	tests := []struct {
		name      string
		extra     []string
		wantErr   []string // substrings every one of which must appear
		wantShadow bool
	}{
		{
			name:  "flag unset, no shadow keys: loads with a nil Shadow",
			extra: nil,
		},
		{
			name:    "a shadow key set with the flag unset names the flag",
			extra:   []string{"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ENABLED"},
		},
		{
			name:    "flag set with no paths reports every missing row at once",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=true"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ORACLE_STATE_DIR", "DUBPLATE_SHADOW_REVIEW_DIR"},
		},
		{
			name:       "flag set with every path resolves Shadow",
			extra:      append([]string{"DUBPLATE_SHADOW_ENABLED=true"}, shadowPaths...),
			wantShadow: true,
		},
		{
			name:       "a threshold row is accepted and never required",
			extra:      append([]string{"DUBPLATE_SHADOW_ENABLED=true", "DUBPLATE_SHADOW_FLOOR=48h"}, shadowPaths...),
			wantShadow: true,
		},
		{
			name:    "an unparseable flag is an error rather than off",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=maybe"},
			wantErr: []string{"DUBPLATE_SHADOW_ENABLED"},
		},
		// The two rows fold M-d adds. An adopter who turns the
		// integration off explicitly must be able to start the
		// server, and a shadow key set beside an off flag must
		// still name the flag.
		{
			name:  "an explicit false loads clean with a nil Shadow",
			extra: []string{"DUBPLATE_SHADOW_ENABLED=false"},
		},
		{
			name:    "an explicit false with a shadow key still names the flag",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=false", "DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ENABLED"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(append(slices.Clone(base), tt.extra...))
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if (cfg.Shadow != nil) != tt.wantShadow {
					t.Fatalf("Shadow non-nil = %v, want %v", cfg.Shadow != nil, tt.wantShadow)
				}
				if tt.wantShadow && cfg.Shadow.ObserveDir != "/srv/inbox" {
					t.Errorf("ObserveDir = %q, want %q", cfg.Shadow.ObserveDir, "/srv/inbox")
				}
				return
			}
			if err == nil {
				t.Fatal("Load returned nil; want an error")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error does not name %s: %v", want, err)
				}
			}
		})
	}
}
```

The unparseable-flag row additionally asserts that `Load` reports
**exactly one** error naming that key rather than two: `Load`'s own
parse error is the single report, and `rejectNamespace`'s
`name == class.Flag` guard is what removes the duplicate (fold, m12).

`internal/config/thresholds_test.go`:

```go
func TestParseDuration(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_SHADOW_FLOOR", Kind: KindDuration, Default: "720h", Doc: "a Go duration"}
	tests := []struct {
		name    string
		environ []string
		want    time.Duration
		wantErr bool
	}{
		{name: "unset takes the default", want: 720 * time.Hour},
		{name: "empty takes the default", environ: []string{"DUBPLATE_SHADOW_FLOOR="}, want: 720 * time.Hour},
		{name: "set wins", environ: []string{"DUBPLATE_SHADOW_FLOOR=48h"}, want: 48 * time.Hour},
		{name: "a bad value errors", environ: []string{"DUBPLATE_SHADOW_FLOOR=soon"}, wantErr: true},
	}
	for _, tt := range tests { /* t.Run, compare, check the error names the key and the Doc */ }
}

func TestParseDurationRejectsABadDefaultAsARegistrationBug(t *testing.T) {
	_, err := ParseDuration(nil, Threshold{Key: "K", Kind: KindDuration, Default: "not a duration"})
	if err == nil || !strings.Contains(err.Error(), "K") {
		t.Errorf("ParseDuration on a bad default = %v; want an error naming the key", err)
	}
}
```

`internal/config/keys_test.go`, **appended at the end of the file**, so
nothing above line 66 moves:

```go
func TestIntegrationClassesCoverEveryShadowKey(t *testing.T) {
	classes := IntegrationClasses()
	if len(classes) != 1 {
		t.Fatalf("IntegrationClasses len = %d, want 1", len(classes))
	}
	c := classes[0]
	if c.Name != "shadow" || c.Rung != 4 || c.Flag != "DUBPLATE_SHADOW_ENABLED" {
		t.Errorf("class = %+v, want the rung 4 shadow class", c)
	}
	if len(c.Required) != 3 || len(c.Optional) != 2 {
		t.Errorf("Required/Optional = %d/%d, want 3/2", len(c.Required), len(c.Optional))
	}
	// Every name the class claims resolves back to it, which is what
	// rejectNamespace's acceptance path depends on.
	for _, name := range slices.Concat([]string{c.Flag}, c.Required, c.Optional) {
		if got, ok := integrationClassOf(name); !ok || got.Name != c.Name {
			t.Errorf("integrationClassOf(%s) = %v, %v; want the shadow class", name, got.Name, ok)
		}
	}
}
```

`cmd/dubplate-server/doctor_test.go`: a case asserting the duration row
renders as `720h0m0s` rather than `invalid:`, and a case asserting the
integration row renders `disabled` with the flag unset and `enabled`
with it set.

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/config/ ./cmd/dubplate-server/ -count=1 -v`
Expected: FAIL, `IntegrationClass`, `ParseDuration` and `Config.Shadow`
undefined; the doctor duration row renders `invalid:`.

- [ ] **Step 3: Implement**

`rejectNamespace` gains one arm, **before** the reserved-class test,
because a name can be in both tables only by mistake and the
integration's own message is the more specific one:

```go
	if class, ok := integrationClassOf(name); ok {
		// The flag names itself, so it is never "set without" itself.
		// Load already reports an unparseable flag value; reporting it
		// again here would emit two errors for one typo, one of them
		// self-contradictory.
		if name == class.Flag {
			continue
		}
		if enabled, err := w.integrationEnabled(class); err != nil || !enabled {
			w.errs = append(w.errs, fmt.Errorf("%s: set without %s, the %s integration's enable flag", name, class.Flag, class.Name))
		}
		continue
	}
```

**The `name == class.Flag` guard is load-bearing, not tidiness** (fold,
M-d and m12). `integrationClassOf` resolves the flag itself, which
Task 3's own `keys_test.go` case asserts by iterating
`slices.Concat([]string{c.Flag}, c.Required, c.Optional)`; `Load` walks
every set `DUBPLATE_*` name; and `wall.get` treats only an empty value
as unset (`config.go:92-98`). Without the guard,
`DUBPLATE_SHADOW_ENABLED=false` reaches this arm with `enabled ==
false` and produces "DUBPLATE_SHADOW_ENABLED: set without
DUBPLATE_SHADOW_ENABLED, the shadow integration's enable flag", so an
adopter who explicitly turns the integration off cannot start the
server. That is the config wall refusing a legal configuration, which is
the failure class ruling 1 exists to prevent.

`Load` gains, after the four owned requires and before
`w.rejectNamespace()`:

```go
	for _, class := range integrationClasses {
		enabled, err := w.integrationEnabled(class)
		if err != nil {
			w.errs = append(w.errs, err)
			continue
		}
		if !enabled {
			continue
		}
		for _, key := range class.Required {
			w.require(key)
		}
	}
```

with the shadow's three resolved into `cfg.Shadow` from the same values.
`integrationEnabled` is a `wall` method so the environment is still read
through `get` alone (`config.go:67-71`), and `IntegrationClass.Enabled`
is its exported wrapper for `doctor`.

`doctor.go` prints the integration rows between the reserved classes and
the thresholds. **Two edits in `cmd/dubplate-server/doctor_test.go` ride
that, and both are named here rather than invented at a red gate**
(fold, M-e). First, `doctor_test.go:110-112` closes its walk with
`if li != len(lines) { ... "extra line(s) beyond owned+reserved+threshold
rows" }` after iterating `config.OwnedWall()`,
`config.ReservedClasses()` and `config.Thresholds()` (`:73-108`), so a
fourth arm over `config.IntegrationClasses()` joins the walk **in the
same position `doctor.go` prints it**, between the reserved and the
threshold arms. Second, `config.Thresholds()` reads the package-level
`registeredThresholds` (`thresholds.go:22-35`) and the only registration
is `newRootCmd`'s; Go runs a package's tests in source order across
files sorted by name, so `doctor_test.go` runs before `main_test.go`,
whose `newRootCmd()` call at `main_test.go:14` would otherwise be the
first construction. **Every doctor test that asserts a threshold row
calls `registerShadowThresholds()` as its first statement**, which is
idempotent by construction and is exactly why R12 wraps it in
`sync.OnceFunc`; without it, `go test -run TestDoctorCmd` and a full
package run see different registry states.

`thresholdStatus` takes a kind switch:

```go
func thresholdStatus(environ []string, t config.Threshold) string {
	switch t.Kind {
	case config.KindDuration:
		d, err := config.ParseDuration(environ, t)
		if err != nil {
			_, reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		return d.String()
	default:
		v, err := config.ParseInt(environ, t)
		if err != nil {
			_, reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		return strconv.Itoa(v)
	}
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/config/ ./cmd/dubplate-server/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Confirm the gitleaks pin, then run the gate and commit**

Before staging, confirm `internal/config/keys_test.go`'s line 66 is the
row `.gitleaksignore` names, by line number and by the row's identity,
**without quoting the literal**. Report the result. If any insertion
landed above it, re-fingerprint `.gitleaksignore` in this same commit
and say so.

Run: `bash scripts/check.sh`
Then: `git add internal/config cmd/dubplate-server/shadow.go cmd/dubplate-server/main.go cmd/dubplate-server/doctor.go cmd/dubplate-server/doctor_test.go && git commit`

**Acceptance criteria:**

- A set shadow key with the flag unset is a load error **naming the
  flag**, not "unknown configuration key".
- With the flag set, every unset `Required` key is reported, all at once,
  in `Load`'s report-everything shape (`config.go:154-158`).
- A `DUBPLATE_SHADOW_FLOOR` set with the flag on loads clean and is
  never required.
- An unparseable flag value is an error rather than a silent off.
- `Config.Shadow` is nil when the flag is unset, and every existing
  `internal/config` test passes unchanged.
- **`DUBPLATE_SHADOW_ENABLED=false` alone loads clean with a nil
  `Shadow`**, and set beside a shadow key it errors naming the flag.
- **An unparseable flag value produces exactly one error for that key.**
- `doctor` prints the shadow class with its flag state and both
  thresholds with their effective values, the duration one as a duration.
- **`cmd/dubplate-server/doctor_test.go` walks the integration rows in
  the position `doctor.go` prints them**, and every doctor test
  asserting a threshold row registers the thresholds itself, so the
  package's tests pass in any order and under `-run`.
- **Constructing `newRootCmd` twice adds no threshold rows beyond the
  first registration**, which is order-independent and is what
  `sync.OnceFunc` buys.
- `keys_test.go` grew only below its pinned line, or `.gitleaksignore`
  was re-fingerprinted in the same commit.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Make `rejectNamespace` accept every integration key unconditionally,
   flag or no flag → `TestLoadShadowIntegration`'s "a shadow key set
   with the flag unset" row fails. This is ruling 1's own stated failure
   mode, a half-configured shadow running unattended and observing
   nothing.
2. Move the two threshold names from `Optional` into `Required` → the "a
   threshold row is accepted and never required" row still passes (it
   sets one), but the "flag set with every path" row fails, reporting
   both thresholds missing. That is the contradiction R1 rejects, made
   visible.
3. Read an unparseable flag as off → the last table row fails.
4. Drop the `Kind` switch from `thresholdStatus` → the doctor duration
   case fails, reporting `invalid:` where `720h0m0s` is wanted.
5. Call `config.Register` directly from `newRootCmd` instead of through
   `sync.OnceFunc` → the double-construction case fails, reporting more
   threshold rows after the second construction than after the first.
6. Drop the `name == class.Flag` guard from `rejectNamespace` → both new
   `DUBPLATE_SHADOW_ENABLED=false` rows fail, the first with a load
   error where a clean load is wanted (fold, M-d).

**Deliverable count: 4** (the integration class with its acceptance and
requirement semantics; the duration threshold kind with its parser and
the doctor switch; the six rows; the `Register` call site under
`sync.OnceFunc`). Gate green; commit. **Report actual token spend, the
six mutations' outcomes, and the `.gitleaksignore` pin's state.**

---

## Task 4: The FLAC reader, its mapping table, and the item-bag round trip

Model: **`opus`** (see Execution model). Depends on Task 0. Independent
of Tasks 1 to 3.

This is the pass's hardest single piece and its one binary-format
surface. It is also **provisional by design**: the reader is
shadow-local, unexported, and rung 8's plan carries the task of
absorbing it into `internal/engine/tagio` and repointing the shadow
(`spec:1007-1012`, `spec:1476-1478`). It says so in its own doc comment.

**Files:**
- Create: `internal/engine/shadow/doc.go`
- Create: `internal/engine/shadow/flac.go` + `flac_test.go`
- Create: `internal/engine/shadow/itembag.go` (the reverse projection)

**Interfaces:**
- Consumes: `matcher.ItemFields` (`info.go:124-152`),
  `matcher.ProjectItemFields` (`info.go:189-211`) in the test only.
- Produces (all unexported except the two the later tasks call):

```go
// internal/engine/shadow/flac.go

// readFLAC reads one FLAC file's matcher input: the nineteen fields
// matcher.ItemFields carries (info.go:124-152), read one way, for
// scoring only.
//
// It is a provisional reader. Rung 8 owns tag I/O and the full field
// surface in internal/engine/tagio (internal/engine/tag/doc.go:8-11);
// rung 4 builds this because waiting on that lane would spend calendar
// on the one rung whose calendar IS the launch date, and rung 8's plan
// carries the task of absorbing it. It writes no tag, reads no format
// but FLAC, and reads exactly two of the container's metadata blocks:
// STREAMINFO (type 0) for the track length, and VORBIS_COMMENT (type 4)
// for everything else. Every other block is skipped by its declared
// length.
//
// An absent or unparseable tag takes the field's Go zero value, which is
// ProjectItemFields' own typed-default rule (info.go:178-188) and beets'
// Item.get behavior.
func readFLAC(path string) (matcher.ItemFields, error)

// internal/engine/shadow/itembag.go

// itemBag projects fields back into the beets field-name bag that
// cassette.Case's Input.Items carries, which is ProjectItemFields' own
// input shape (matcher/info.go:189-211).
//
// The names are spelled here because matcher's own field-name constants
// are unexported (info.go:156-176). Two decoders over one field bag
// drift silently (info.go:148-150), so the round-trip test below is
// what keeps these nineteen names in step with that table rather than a
// comment asking a reader to check.
func itemBag(fields matcher.ItemFields) map[string]any
```

**The container format, stated so the implementer builds to it rather
than to a memory.** A FLAC file is the four-byte signature `fLaC`,
then a run of metadata blocks. Each block header is four bytes: one byte
whose high bit is the last-block flag and whose low seven bits are the
block type, then a three-byte big-endian payload length. `STREAMINFO`
is type 0 with a 34-byte payload, **stated by offset rather than by
"last", because revision 1 named the wrong eight bytes** (fold, B1):

| Offset | Width | Field |
| --- | --- | --- |
| `[0:2]` | 16 bits | minimum block size |
| `[2:4]` | 16 bits | maximum block size |
| `[4:7]` | 24 bits | minimum frame size |
| `[7:10]` | 24 bits | maximum frame size |
| `[10:18]` | 64 bits | the pack: sample rate (bits 0-19), channels minus one (20-22), bits per sample minus one (23-27), total samples (28-63), big-endian across the eight bytes |
| `[18:34]` | 128 bits | the MD5 signature of the unencoded audio |

Pinned mutagen writes exactly that order at
`tools/venv/.venv/lib/python3.13/site-packages/mutagen/flac.py:233-257`
(the four sizes at `:235-238`, the pack at `:240-252`, the MD5 at
`:253-257`) and reads it back at `:208-231`. **The payload's last 64
bits are the MD5's second half, not the pack**, so a reader built to
revision 1's sentence divides one half of a digest by the other.
`Length` is total samples divided by sample rate, as a `float64`, and is
zero when the sample rate is zero. `VORBIS_COMMENT` is type 4 and is
**little-endian**: a uint32 vendor length, the vendor string, a uint32
comment count, then per comment a uint32 length and that many bytes of
`KEY=value` UTF-8. The committed generator states the same framing for
the two-field case at
`internal/harness/oracle/testdata/gen_template_flac.py:64-75`, and its
42-byte output is a signature plus one 34-byte STREAMINFO block.

**Keys fold to upper case**, because mutagen's Vorbis lookup is
case-insensitive, and **the first occurrence of a key wins**, because
`StorageStyle.fetch` takes `mutagen_file[key][0]`, the first value of a
multi-valued comment
(`tools/venv/.venv/lib/python3.13/site-packages/mediafile/storage/base.py:77-84`).

**The mapping table, derived from pinned mediafile 0.17.0**
(`tools/venv/.venv/lib/python3.13/site-packages/mediafile/__init__.py`,
the version `corpus/manifest.json` pins). **Seventeen Vorbis-key rows,
one STREAMINFO derivation, and one field with no file source.** The spec
says "the Vorbis-comment key for each of the nineteen fields"
(`spec:1014-1018`) and restates it as "nineteen fields read one way"
(`spec:1011-1012`); two of the nineteen have no Vorbis key at all, and
the table says so rather than inventing one.

**A field with more than one style takes the FIRST style that yields a
truthy value**, because `MediaField.__get__` walks the styles in
declaration order and breaks on the first truthy `out`
(`mediafile/fields.py:48-54`). **An absent or empty-string value falls
through to the next key; a `"0"` does not**, because the truthiness test
runs on the raw fetched string before `safe_cast` sees it, and `"0"` is
a truthy Python string (fold, m9). So `TRACK=0` beside `TRACKNUMBER=5`
yields 0, never 5. The order below is load-bearing rather than
decorative. Read-only styles are marked; they are read here
because this reader only reads.

| `ItemFields` member | Go type | Vorbis key(s), in mediafile's declaration order | mediafile line |
| --- | --- | --- | --- |
| `Title` | string | `TITLE` | `:392` |
| `Artist` | string | `ARTIST` | `:398` |
| `Album` | string | `ALBUM` | `:410` |
| `AlbumArtist` | string | `ALBUM ARTIST`, then `ALBUM_ARTIST`, then `ALBUMARTIST` | `:542-544` |
| `Track` | int | `TRACK`, then `TRACKNUMBER` | `:463-464` |
| `Disc` | int | `DISC`, then `DISCNUMBER` | `:480-481` |
| `DiscTotal` | int | `DISCTOTAL`, then `DISCC`, then `TOTALDISCS` | `:488-490` |
| `Comp` | bool | `COMPILATION` | `:535` |
| `Label` | string | `LABEL`, then `PUBLISHER` | `:574-575` |
| `CatalogNum` | string | `CATALOGNUMBER`, then read-only `CATALOGID`, then read-only `DISCOGS_CATALOG` | `:604-606` |
| `Barcode` | string | `BARCODE`, then read-only `UPC`, `EAN/UPN`, `EAN`, `UPN` | `:616-620` |
| `Country` | string | `RELEASECOUNTRY` | `:659` |
| `Media` | string | `MEDIA` | `:673` |
| `AlbumDisambig` | string | `MUSICBRAINZ_ALBUMCOMMENT` | `:680` |
| `MBTrackID` | string | `MUSICBRAINZ_TRACKID` | `:764` |
| `MBAlbumID` | string | `MUSICBRAINZ_ALBUMID` | `:782` |
| `Year` | int | `DATE`'s year component, with `YEAR` as the fallback style | `:685-693` |
| `Length` | float64 | **none.** STREAMINFO's total samples over its sample rate | `:892-895` |
| `DataSource` | string | **none.** No `MediaField` exists in mediafile | see below |

**`Year` is not a plain field.** `year = date.year_field()`
(`:693`) over a `DateField` whose Vorbis style is `DATE` (`:688`) with
`year=(StorageStyle("YEAR"),)` as a fallback (`:690`). `_get_date_tuple`
strips anything from a `T` or a space onward, splits on `-` or `/`,
truncates or pads to three components, and **uses the year style only
when the first component is falsy** (`fields.py:156-172`). So
`DATE=1994-05-01` yields 1994; `DATE=` with `YEAR=1994` yields 1994; a
`DATE` whose year component does not parse as an integer yields the zero
value.

**`DataSource` has no file source at all.** A grep of pinned mediafile
for `data_source` or `DATA_SOURCE` returns nothing, because beets keeps
it as a flexible attribute in its database rather than in the file. The
reader therefore always produces `DataSource == ""` for an inbox file,
which `ProjectItemFields`' typed-default rule makes correct behavior:
`info.go:115-123` records that this field's true beets default is `None`
and models it as `""` anyway.

**Numeric parsing follows mediafile's `safe_cast`**
(`mediafile/utils/type_conversion.py:15-27`): for an int, take the
leading `[+-]?[0-9]+` of the trimmed string and return 0 when there is
none, so `TRACKNUMBER=3/12` is 3 and `TRACKNUMBER=A` is 0. For a bool,
`bool(int(val))` with a `ValueError` falling to false
(`:29-34`), so `COMPILATION=1` is true and `COMPILATION=yes` is false.

**The reader refuses the rest, explicitly.** It reads no other block
type, no other container, no other tag, and it never writes. A file
whose signature is not `fLaC`, whose block header runs past the end, or
whose VORBIS_COMMENT payload is truncated is an error rather than a
partial read, because contract S8 turns any unreadable entry into a
discarded capture and a silently short field bag would score an album
the shadow never actually read.

- [ ] **Step 1: Write the failing tests**

`flac_test.go` builds its fixtures in Go. **No binary fixture is
vendored**, so the mutation that proves the reader fires is a byte edit
(`spec:1024-1030`).

```go
// buildFLAC assembles a minimal FLAC container in memory: the fLaC
// signature, a STREAMINFO block carrying rate and samples, and a
// VORBIS_COMMENT block carrying tags in the order given. It is the
// fixture builder the whole package shares, and it is a builder rather
// than a vendored file so a mutation is a byte edit rather than a
// re-derivation (rung 4 spec, the FLAC reader's own cost section).
func buildFLAC(t *testing.T, tags []string, samples uint64, rate uint32) []byte

func TestReadFLACMapsEveryVorbisKey(t *testing.T) {
	tags := []string{
		"TITLE=Sub Rosa", "ARTIST=Cabinet", "ALBUM=Windrose",
		"ALBUMARTIST=Cabinet Ensemble",
		"TRACKNUMBER=3/12", "DISCNUMBER=2", "TOTALDISCS=2",
		"COMPILATION=1", "PUBLISHER=Nightwood", "CATALOGNUMBER=NW-004",
		"BARCODE=5901234123457", "RELEASECOUNTRY=GB", "MEDIA=CD",
		"MUSICBRAINZ_ALBUMCOMMENT=deluxe",
		"MUSICBRAINZ_TRACKID=11111111-1111-1111-1111-111111111111",
		"MUSICBRAINZ_ALBUMID=22222222-2222-2222-2222-222222222222",
		"DATE=1994-05-01",
	}
	path := writeTemp(t, buildFLAC(t, tags, 441000, 44100))

	got, err := readFLAC(path)
	if err != nil {
		t.Fatalf("readFLAC: %v", err)
	}
	want := matcher.ItemFields{
		Title: "Sub Rosa", Artist: "Cabinet", Album: "Windrose",
		AlbumArtist: "Cabinet Ensemble",
		Track: 3, Disc: 2, DiscTotal: 2, Comp: true,
		Label: "Nightwood", CatalogNum: "NW-004", Barcode: "5901234123457",
		Country: "GB", Media: "CD", AlbumDisambig: "deluxe",
		MBTrackID: "11111111-1111-1111-1111-111111111111",
		MBAlbumID: "22222222-2222-2222-2222-222222222222",
		Year: 1994, Length: 10,
		// DataSource stays "": beets keeps it as a flexible attribute
		// rather than in the file, so no Vorbis key exists.
	}
	if got != want {
		t.Errorf("readFLAC =\n%+v\nwant\n%+v", got, want)
	}
}

func TestReadFLACStylePrecedence(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want matcher.ItemFields
	}{
		{
			name: "album artist takes the first declared style",
			tags: []string{"ALBUM ARTIST=First", "ALBUM_ARTIST=Second", "ALBUMARTIST=Third"},
			want: matcher.ItemFields{AlbumArtist: "First"},
		},
		{
			name: "an empty first style falls through to the next",
			tags: []string{"ALBUM ARTIST=", "ALBUM_ARTIST=Second", "ALBUMARTIST=Third"},
			want: matcher.ItemFields{AlbumArtist: "Second"},
		},
		{
			name: "track prefers TRACK over TRACKNUMBER",
			tags: []string{"TRACKNUMBER=9", "TRACK=4"},
			want: matcher.ItemFields{Track: 4},
		},
		{
			name: "catalognum falls through to a read-only alias",
			tags: []string{"CATALOGID=RO-1"},
			want: matcher.ItemFields{CatalogNum: "RO-1"},
		},
		{
			name: "barcode falls through the four read-only aliases in order",
			tags: []string{"EAN=E-1", "UPN=U-1"},
			want: matcher.ItemFields{Barcode: "E-1"},
		},
		{
			name: "year falls back to YEAR when DATE has no year component",
			tags: []string{"DATE=", "YEAR=1979"},
			want: matcher.ItemFields{Year: 1979},
		},
		{
			name: "keys are case-insensitive",
			tags: []string{"title=lowercase key"},
			want: matcher.ItemFields{Title: "lowercase key"},
		},
		{
			name: "the first occurrence of a repeated key wins",
			tags: []string{"TITLE=first", "TITLE=second"},
			want: matcher.ItemFields{Title: "first"},
		},
		{
			name: "an unparseable int takes the zero value",
			tags: []string{"TRACKNUMBER=A-side"},
			want: matcher.ItemFields{},
		},
		{
			name: "a non-numeric compilation is false",
			tags: []string{"COMPILATION=yes"},
			want: matcher.ItemFields{},
		},
	}
	// Each row builds a FLAC at 44100 Hz with samples 0, so Length
	// stays 0 and the fixture stays a container pinned mutagen would
	// open: a STREAMINFO sample rate of 0 is refused outright
	// (mutagen/flac.py:222-223, "A sample rate value of 0 is
	// invalid"), so a rate-0 fixture is a file the oracle itself could
	// not read (fold, m10). Each row reads the file and compares the
	// whole struct.
}

func TestReadFLACLengthFromStreaminfo(t *testing.T) {
	// 441000 samples at 44100 Hz is exactly 10 seconds; 0 Hz is 0,
	// never a division by zero.
}

// The two assertions that make the container format checkable at all
// (fold, M-a). Every other test in this file reads bytes buildFLAC
// produced, so the builder and the parser can agree on a wrong layout
// forever; these two pin each of them to something outside the task.
func TestReadFLACParsesTheCommittedTemplateBytes(t *testing.T) {
	// The 42 bytes of internal/harness/oracle/testdata/template.flac,
	// embedded here as a Go byte literal rather than read from that
	// path: it is harness territory and stays unread by the engine
	// (rung 4 spec, the FLAC reader's fixture section). Its provenance
	// is internal/harness/oracle/testdata/gen_template_flac.py, which
	// builds it through mutagen's own StreamInfo.write (:52-61) as
	// mono, 44100 Hz, 16-bit, zero samples.
	//
	// Assert sample rate 44100, channels 1, bits per sample 16 and
	// total samples 0 off the parsed block, which is a real FLAC no
	// code in this package produced.
}

func TestBuildFLACEmitsTheStreaminfoLayoutByteForByte(t *testing.T) {
	// A hand-written 34-byte expectation derived from the offset table
	// above, compared against buildFLAC's own STREAMINFO payload. This
	// pins the BUILDER to the format rather than to the parser, so a
	// shared misreading fails here rather than shipping.
}

func TestReadFLACRefusesAMalformedContainer(t *testing.T) {
	tests := []struct{ name string; data []byte }{
		{"not a FLAC", []byte("OggS....")},
		{"truncated block header", /* fLaC plus two bytes */},
		{"block length past the end", /* a header claiming more than remains */},
		{"truncated vorbis comment payload", /* a comment length past the block */},
		{"streaminfo shorter than 34 bytes", /* a type-0 block of 12 bytes */},
	}
	// Each row asserts a non-nil error, and that readFLAC returns the
	// zero ItemFields rather than a partly filled one.
}

func TestItemBagRoundTripsThroughProjectItemFields(t *testing.T) {
	// Every one of the nineteen fields carries a distinguishable
	// non-zero value, so a name this table spells wrong shows up as a
	// dropped field rather than as a coincidence.
	want := matcher.ItemFields{ /* all nineteen populated */ }
	if got := matcher.ProjectItemFields(itemBag(want)); got != want {
		t.Errorf("round trip =\n%+v\nwant\n%+v", got, want)
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: FAIL, the package does not exist.

- [ ] **Step 3: Write `doc.go`, `flac.go` and `itembag.go`**

`doc.go` states the package's one reason to exist, in the shape
`internal/engine/musicbrainz/doc.go` uses: this package observes the
incumbent's live decisions and compares them against the Go matcher's,
it imports no harness package and no `internal/config`, it reads no
clock, and every path, knob and stamp reaches it as a parameter.

`flac.go` parses the container as described, folding keys to upper case
into a `map[string][]string` that preserves per-key order, then walking
the mapping table's style lists in declaration order and taking the
first truthy value. Guard clauses over nesting; every bounds check
returns a wrapped error naming the offset.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow && git commit`

**Acceptance criteria:**

- All seventeen Vorbis-mapped fields are read, and the whole-struct
  comparison in `TestReadFLACMapsEveryVorbisKey` is one equality rather
  than seventeen.
- Style precedence follows declaration order with an empty value falling
  through, proven per multi-style field.
- `Year` reads `DATE`'s year component with a `YEAR` fallback, and a
  `DATE` with no parseable year yields 0.
- `Length` is total samples over sample rate, read from payload bytes
  `[10:18]`, and a zero rate yields 0 rather than a panic.
- **The committed 42-byte template parses to 44100 Hz, 1 channel, 16
  bits and 0 samples**, and `buildFLAC`'s own STREAMINFO payload matches
  a hand-written 34-byte expectation byte for byte. Together these are
  the only assertions in the task that a shared misreading of the
  container cannot satisfy.
- Every style-precedence fixture is built at 44100 Hz, so no fixture is
  a container the oracle could not open.
- `DataSource` is `""` for every input, with the reason in the doc
  comment.
- Every malformed container is an error, and no partial `ItemFields` is
  ever returned.
- The reader is unexported, is named provisional in its doc comment, and
  names rung 8 as the pass that absorbs it.
- `itemBag` round-trips through `matcher.ProjectItemFields` for all
  nineteen fields.
- `TestEngineNeverImportsHarness` and `TestEngineNeverReadsTheClock`
  both pass with the new package inside their walks. **The tests in this
  package call `time.Now` nowhere**; `internal/engine/tag/tag_test.go:18-20`
  is the pattern to copy.
- `go.mod` unchanged.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Reverse `AlbumArtist`'s style order to `ALBUMARTIST` first →
   `TestReadFLACStylePrecedence`'s first row fails, reporting "Third"
   where "First" is wanted.
2. Take the last occurrence of a repeated key instead of the first →
   the repeated-key row fails.
3. Read `Length` as total samples rather than samples over rate →
   `TestReadFLACLengthFromStreaminfo` fails, reporting 441000 where 10
   is wanted. This is the mutation that matters most downstream: `Length`
   feeds the track-length penalty, so a wrong unit moves every verdict.
4. Drop the `YEAR` fallback → the year-fallback row fails.
5. Return the fields read so far on a truncated VORBIS_COMMENT instead
   of an error → `TestReadFLACRefusesAMalformedContainer`'s truncated
   row fails, and the failure names the field bag that would otherwise
   have scored an album nobody read.
6. Misspell one name in `itemBag`, for example `mb_album_id` for
   `mb_albumid` → `TestItemBagRoundTripsThroughProjectItemFields` fails
   on that field alone, which is the drift `info.go:148-150` warns about
   made mechanical.
7. Read the pack from `payload[26:34]` instead of `[10:18]`, which is
   revision 1's own wrong sentence → the two external assertions both
   fail (the template's sample rate reads out of MD5 bytes, and
   `buildFLAC`'s payload no longer matches the hand-written
   expectation), while every test that reads only `buildFLAC`'s output
   still passes. That asymmetry is the reason both assertions exist
   (fold, B1 and M-a).

**Deliverable count: 4** (the package doc; the FLAC container parser
with its two external layout assertions; the nineteen-field mapping with
its precedence rules; `itemBag` and its round trip). Gate green; commit.
**Report actual token spend, the seven mutations' outcomes, and the
measured wall time of the package's test run. This task's report is where the conductor re-derives the pass
ceiling.**

---

## Task 5: The store, its types, and the consumed-report ledger

Model: `sonnet`. Depends on Tasks 1 and 4.

Everything the shadow writes lives under one directory, so 4b's
`ReadWritePaths=` is one path and the argument is static rather than
assumed. This task lands the layout, the types every later task fills,
and the atomic write path.

**Files:**
- Create: `internal/engine/shadow/store.go` + `store_test.go`
- Create: `internal/engine/shadow/types.go`

**Interfaces:**
- Consumes: `atomicfile.WriteFile` (Task 1), `matcher.ItemFields`.
- Produces:

```go
// internal/engine/shadow/types.go

// Divergence classes and the agreement, the closed vocabulary contract
// S4 fixes. A class outside this set is a bug, never a new row.
const (
	ClassAgreement                 = "agreement"
	ClassGoAppliesOracleSkipped    = "go-applies-oracle-skipped"
	ClassGoSkipsOracleApplied      = "go-skips-oracle-applied"
	ClassGoAppliesDifferentRelease = "go-applies-different-release"
)

// Deferral reasons. Each names a musicbrainz sentinel a caller branches
// on to STOP ASKING rather than to retry (limiter.go:45-53,
// client.go:74-84). DeferUnclassified is the closed enumeration's
// default arm: an error the classifier does not recognize defers and is
// counted, never recorded as a verdict (Resolution R15).
const (
	DeferUnavailable     = "musicbrainz-unavailable"
	DeferBudgetExhausted = "budget-exhausted"
	DeferLimiterClosed   = "limiter-closed"
	DeferUnclassified    = "unclassified-error"
)

// Unreconciled reasons, contract S11's four.
//
// UnreconciledNoReleaseID covers both halves of one honest gap: an
// imported outcome whose LibraryPaths is empty, and one whose named
// library file carries no MUSICBRAINZ_ALBUMID. Either way the oracle
// named no release, and recording an apply against an empty release
// would classify as go-applies-different-release against any Go apply,
// which is a divergence against a release the oracle never named
// (Resolution R22).
const (
	UnreconciledAmbiguousBasename = "ambiguous-basename"
	UnreconciledNoReleaseID       = "imported-with-no-release-id"
	UnreconciledReportSwept       = "report-swept"
	UnreconciledPartialSkip       = "partial-skip"
)

// Exclusion reasons, counted and never scored. Each names a directory
// the shadow saw and deliberately did not observe, so the report prints
// a count with a reason rather than silence.
const (
	ExcludedMultiAlbum   = "multi-album-directory"
	ExcludedGateRejected = "gate-rejected"
	ExcludedNoTracks     = "no-tracks"
)

// A CapturedFile is one file the capture listed, in the digest's own
// terms.
type CapturedFile struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// A Capture is one album directory as the shadow saw it before the
// incumbent imported it. Digest is the identity (contract S7).
type Capture struct {
	Uploader   string               `json:"uploader"`
	Basename   string               `json:"basename"`
	Digest     string               `json:"digest"`
	Files      []CapturedFile       `json:"files"`
	Items      []matcher.ItemFields `json:"items"`
	CapturedAt time.Time            `json:"captured_at"`
}

// An OracleSignal is the incumbent's own verdict for one album, and the
// artifact it was read from. Signal is "apply" or "skip" and nothing
// else: the oracle is binary (contract S3).
type OracleSignal struct {
	Signal        string `json:"signal"`
	ReleaseMBID   string `json:"release_mbid,omitempty"`
	Uploader      string `json:"uploader"`
	AlbumBasename string `json:"album_basename"`
	ReadFrom      string `json:"read_from"`
}

// A GoVerdict is one scoring mode's answer. Hypothesis marks the
// DefaultCompat re-score, whose verdict is not the product's true
// DefaultCompat verdict wherever the real flow would have widened the
// candidate set (contract S16).
type GoVerdict struct {
	Recommendation string   `json:"recommendation"`
	Action         string   `json:"action"`
	AlbumID        string   `json:"album_id,omitempty"`
	Distance       float64  `json:"distance"`
	NonzeroKeys    []string `json:"nonzero_keys"`
	Hypothesis     bool     `json:"hypothesis,omitempty"`
}

// GoVerdicts is the pair the two scoring modes produce. Default is nil
// when the re-score could not run, which is honest rather than a zero
// verdict a reader would compare against.
type GoVerdicts struct {
	Compat  *GoVerdict `json:"compat"`
	Default *GoVerdict `json:"default,omitempty"`
}

// A BudgetRecord is one pass's MusicBrainz spend, persisted because a
// Budget is in-process state (contract S14).
type BudgetRecord struct {
	Name  string `json:"name"`
	Spent int    `json:"spent"`
}

// An Observation is observations/<obs-id>/observation.json: one album,
// the two verdicts, and the class. Exactly one of Class, Deferred,
// Unreconciled and Excluded is set on a stored record; an observation
// still awaiting its oracle signal carries none of the four.
//
// Excluded is the fourth because one of the three exclusions arrives
// after the record already exists: a multi-album directory and a
// directory with no tracks are excluded at capture and never become
// observations at all, while a gate-rejected album is a single-album
// directory the capture stores and only the run report later reveals as
// quarantined or rejected (fold, B5).
type Observation struct {
	ID                string       `json:"id"`
	Album             string       `json:"album"`
	MatcherCommit     string       `json:"matcher_commit"`
	MatcherConfigHash string       `json:"matcher_config_hash"`
	EvalDate          time.Time    `json:"eval_date"`
	Capture           Capture      `json:"capture"`
	Oracle            *OracleSignal `json:"oracle,omitempty"`
	Go                *GoVerdicts  `json:"go,omitempty"`
	Class             string       `json:"class,omitempty"`
	Deferred          string       `json:"deferred,omitempty"`
	Unreconciled      string       `json:"unreconciled,omitempty"`
	Excluded          string       `json:"excluded,omitempty"`
	Superseded        bool         `json:"superseded,omitempty"`
	Budget            BudgetRecord `json:"budget"`
	Requests          []RecordedRequest `json:"requests,omitempty"`
}

// A RecordedRequest is one captured wire request and the file its body
// was stored in, in cassette.Request's own field order so
// cmd/dubplate-server's marshalling is a field copy rather than a
// translation.
type RecordedRequest struct {
	Method    string `json:"method"`
	PathQuery string `json:"path_query"`
	Response  string `json:"response"`
}

// A PassRecord is passes/<pass-id>.json. WireInterval is recorded
// because Budget.Spent alone cannot reconstruct the pace a run used,
// and the pace is a conductor ruling a later reader will want to check.
type PassRecord struct {
	ID            string         `json:"id"`
	StartedAt     time.Time      `json:"started_at"`
	EndedAt       time.Time      `json:"ended_at"`
	MatcherCommit string         `json:"matcher_commit"`
	WireInterval  string         `json:"wire_interval"`
	Captured      int            `json:"captured"`
	PartialReads  int            `json:"partial_reads"`
	Superseded    int            `json:"superseded"`
	Observed      int            `json:"observed"`
	Agreements    int            `json:"agreements"`
	Divergences   map[string]int `json:"divergences"`
	Deferred      map[string]int `json:"deferred"`
	Unreconciled  map[string]int `json:"unreconciled"`
	Excluded      map[string]int `json:"excluded"`
	ReportGaps    int            `json:"report_gaps"`
	Budget        BudgetRecord   `json:"budget"`
}

// A ConsumedReport is one incumbent run report this shadow has already
// reconciled. StartedAt and EndedAt are decoded from the report rather
// than read off the file, because RunID carries no order
// (Resolution R6).
type ConsumedReport struct {
	File      string    `json:"file"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// internal/engine/shadow/store.go

// A Store is the shadow's durable state under one directory, rooted at
// <state dir>/shadow. It is the unit's only writable path, so every
// method here writes inside root and nowhere else, and every write goes
// through atomicfile.WriteFile: a reader never sees a partial file, and
// the observe pass and a hand-run report never contend.
type Store struct {
	root string
}

// OpenStore creates root and its four subdirectories if they do not
// exist and returns a Store over them. Directories are 0755 and files
// 0644, per go-conventions.
func OpenStore(root string) (*Store, error)

// ClockStart reports the recorded clock start and whether one exists.
func (s *Store) ClockStart() (time.Time, bool, error)

// StartClock writes started-at if and only if no clock start exists.
// It reports whether this call wrote it. The clock start is written
// once and never rewritten: launch date is this date plus the floor, so
// a rewrite would silently move the launch date.
func (s *Store) StartClock(now time.Time) (bool, error)

// PutObservation writes obs to observations/<obs.ID>/observation.json.
func (s *Store) PutObservation(obs *Observation) error

// Observations returns every stored observation, in lexical id order.
// That is the STORE's read order and not the pass's match order: an id
// is a digest prefix (Resolution R5), so id order says nothing about
// capture order, and Run matches in (uploader, basename) capture order
// instead (Resolution R18).
func (s *Store) Observations() ([]*Observation, error)

// HasObservation reports whether an observation directory exists for id,
// which is the capture's idempotence check (contract S7).
func (s *Store) HasObservation(id string) bool

// PutResponse writes one recorded body to
// observations/<obsID>/responses/<n>.json, named %03d from 1, matching
// the corpus's own case-directory convention.
func (s *Store) PutResponse(obsID string, n int, body []byte) error

// Response reads one recorded body back, for the re-score and for
// reevaluate.
func (s *Store) Response(obsID, name string) ([]byte, error)

// PutCaseFile writes the marshalled cassette-shaped fixture.
// internal/engine/shadow may import no harness package (contract S19),
// so it takes bytes cmd/dubplate-server produced rather than a
// cassette.Case.
func (s *Store) PutCaseFile(obsID string, data []byte) error

// PutPending, RemovePending and PutResolved manage the two staged-row
// directories. PutPending takes marshalled bytes for the same reason
// PutCaseFile does.
func (s *Store) PutPending(obsID string, data []byte) error
func (s *Store) RemovePending(obsID string) error
func (s *Store) PutResolved(obsID string, data []byte) error
func (s *Store) PendingIDs() ([]string, error)
func (s *Store) ResolvedIDs() ([]string, error)

// PutPass writes one pass record.
func (s *Store) PutPass(p *PassRecord) error

// Passes returns every pass record, in lexical id order.
func (s *Store) Passes() ([]*PassRecord, error)

// ConsumedReports and PutConsumedReports read and replace the ledger
// that makes a swept report a named condition rather than silence
// (contract S12).
func (s *Store) ConsumedReports() ([]ConsumedReport, error)
func (s *Store) PutConsumedReports(rs []ConsumedReport) error

// Size returns the store's own footprint in bytes, the report's sizing
// row (rung 4 spec, risk 5).
func (s *Store) Size() (int64, error)
```

**The pass id** is `now.UTC().Format("20060102T150405.000000000Z")`, a
sortable, collision-free-in-practice name derived from the pass clock
rather than from a random source, so `Passes()`' lexical order is
chronological order and `report` needs no secondary sort. `now` is a
parameter (contract S16's no-clock rule), so a test's frozen clock makes
the id deterministic.

**`started-at`'s on-disk format** is RFC 3339 nanoseconds plus a
trailing newline, so `cat started-at` on the box reads as a date rather
than as an integer nobody can check.

- [ ] **Step 1: Write the failing tests**

```go
func TestStartClockWritesOnceAndNeverRewrites(t *testing.T) {
	s := openTempStore(t)
	first := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	wrote, err := s.StartClock(first)
	if err != nil || !wrote {
		t.Fatalf("StartClock = %v, %v; want true, nil", wrote, err)
	}
	later := first.Add(72 * time.Hour)
	wrote, err = s.StartClock(later)
	if err != nil {
		t.Fatalf("StartClock: %v", err)
	}
	if wrote {
		t.Error("StartClock reported a second write; the clock start is written once")
	}
	got, ok, err := s.ClockStart()
	if err != nil || !ok {
		t.Fatalf("ClockStart = %v, %v, %v", got, ok, err)
	}
	if !got.Equal(first) {
		t.Errorf("ClockStart = %v, want %v; a rewrite moves the launch date silently", got, first)
	}
}

func TestStoreRoundTripsEveryArtifact(t *testing.T) {
	// One table over PutObservation/Observations, PutPass/Passes,
	// PutResponse/Response, PutConsumedReports/ConsumedReports,
	// PutPending/PendingIDs/RemovePending, PutResolved/ResolvedIDs.
	// Each row asserts the value read back equals the value written,
	// field for field, and that Observations and Passes come back in
	// lexical id order rather than in filesystem order.
}

func TestStoreWritesNothingOutsideItsRoot(t *testing.T) {
	// Open a store under a subdirectory of t.TempDir(), write one of
	// every artifact, then walk the whole temp dir and assert every
	// created path is under root. This is 4b's ReadWritePaths argument
	// made mechanical: the unit's only writable path is this one.
}

func TestStoreLeavesNoTempFile(t *testing.T) {
	// After a write of every artifact, no entry under root matches
	// "*.tmp-*". atomicfile creates its temp in the target's own
	// directory (atomicfile.go:16-17), so a leaked temp would sit
	// inside the store and be read back by Observations.
}

func TestHasObservationIsTheIdempotenceCheck(t *testing.T) {
	// False before, true after PutObservation, for the same id.
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ -run 'Store|Clock|HasObservation' -count=1 -v`
Expected: FAIL, `OpenStore` undefined.

- [ ] **Step 3: Write `types.go` and `store.go`**

Every write is `atomicfile.WriteFile(path, data, 0o644)` after an
`os.MkdirAll(dir, 0o755)`. Every read is `os.ReadFile` plus
`json.Unmarshal`. `Observations` and `Passes` sort with `slices.SortFunc`
and `cmp.Compare` on the id rather than trusting `os.ReadDir`, so the
order is stated rather than inherited.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow && git commit`

**Acceptance criteria:**

- `started-at` is written once and a second call neither rewrites it nor
  errors.
- Every artifact round-trips field for field.
- Every write goes through `atomicfile.WriteFile`; no `os.WriteFile`
  appears anywhere in the package.
- Nothing is written outside the store root, proven by a walk.
- No temp file survives any write.
- `Observations` and `Passes` return lexical id order, asserted against
  a deliberately out-of-order insertion.
- The pass id derives from the `now` parameter, so two passes at one
  frozen clock collide and the test says so rather than the code
  pretending otherwise. **If they collide, the second write wins and the
  test asserts that**, which is the honest behavior for a store whose
  ids are clock-derived.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Let `StartClock` overwrite → `TestStartClockWritesOnceAndNeverRewrites`
   fails on both the report flag and the read-back value.
2. Replace one `atomicfile.WriteFile` with `os.WriteFile` → nothing in
   this suite fails, because a successful write looks identical.
   **Reported as a deliberately unfired mutation**, with the guard named
   at the call site: the property is held by `TestStoreLeavesNoTempFile`
   only indirectly, and the real guard is the package rule in Global
   Constraints plus the `diff-reviewer`'s read.
3. Write an observation outside root, for example by joining an
   attacker-shaped id containing `..` → `TestStoreWritesNothingOutsideItsRoot`
   fails. **The store rejects any id that is not a single clean path
   segment**, and the test carries that row: ids are digest-derived
   (R5), so a non-hex id is a bug rather than input, and the check makes
   it a loud one.
4. Sort `Observations` by `os.ReadDir` order instead of by id → the
   out-of-order insertion row fails on at least one platform, and the
   test states the order it wants rather than the order it observed.

**Deliverable count: 3** (the store's layout and atomic write path; the
observation, pass, capture and verdict types; the consumed-report
ledger). Gate green; commit. **Report actual token spend and the four
mutations' outcomes, including mutation 2's non-firing.**

---

## Task 6: The capture, its grouping rule, its digest, and the partial-read discard

Model: `sonnet`. Depends on Tasks 4 and 5.

**Files:**
- Create: `internal/engine/shadow/capture.go` + `capture_test.go`

**Interfaces:**
- Consumes: `readFLAC` (Task 4), `Capture`, `CapturedFile`,
  `ExcludedMultiAlbum` (Task 5).
- Produces:

```go
// internal/engine/shadow/capture.go

// A captureResult is one pass's walk of the observe directory.
type captureResult struct {
	Captures     []Capture
	Excluded     []excludedDir // multi-album directories, with their group counts
	PartialReads []string      // "<uploader>/<basename>" of every discarded capture
}

// An excludedDir is a directory the capture could not resolve to one
// album, kept so the report prints a count with a reason rather than
// silence.
type excludedDir struct {
	Uploader string
	Basename string
	Reason   string
	Groups   int // for ExcludedMultiAlbum, the number of (albumartist, album) groups found
}

// capture walks observeDir/<uploader>/<basename>/ and reads one fixture
// per album directory.
//
// The walk is lexical, matching os.ReadDir's own order and the
// incumbent's listContributorDirs (_salvage/internal/importer/importer.go:1238-1254),
// so a budget-capped pass reaches albums in a stated order rather than
// an accidental one.
//
// Three rules, each of which costs an album rather than risking a lie:
//
// A capture that cannot read every entry it listed is DISCARDED
// (contract S8). The importer moves an album out of the inbox with
// os.Rename or a copy-then-remove fallback (_salvage/internal/review/move.go:24-40),
// so a walk can list four tracks and take ENOENT on the fourth, and
// once the album is gone there is no next pass to supersede a partial
// fixture with a whole one.
//
// A directory whose items group to more than one album is EXCLUDED with
// its group count (contract S9). The box runs group_albums: yes
// (config/beets-config.yaml:23), so beets is free to split one uploaded
// directory into several verdicts, while tag.Album takes items
// []matcher.ItemFields and calls MostCommonTags over all of them
// (internal/engine/tag/tag.go:75). Rung 4 observes only what groups to
// exactly one album, and counts the rest.
//
// The key is (cmp.Or(AlbumArtist, Artist), Album), which is beets' own:
// group(item) returns (item.albumartist or item.artist, item.album)
// (beets/importer/stages.py:102-103, consumed at :112-113). The
// fallback is what handles the estate's real material: a contributor
// ZIP extracts flat (config/beets-config.yaml:19-22) and its tracks
// often carry ARTIST and no ALBUMARTIST, which an albumartist-only key
// collapses into one group and observes as one album while beets splits
// it into several verdicts. Provisional, conductor ruling owed: the
// spec states the key as albumartist and album alone (spec:441).
//
// A directory whose digest already names a stored observation is
// SKIPPED, which is what makes the capture idempotent (contract S7).
func capture(observeDir string, has func(id string) bool, now time.Time) (captureResult, error)

// digestFor computes a capture's identity: the hex SHA-256 over the
// uploader, the basename, and each file's name, size and modification
// time in lexical name order, NUL-separated.
func digestFor(uploader, basename string, files []CapturedFile) string

// observationID is "shadow-" plus a digest's first sixteen hex
// characters (Resolution R5). It names the observation directory, the
// pending row and the resolved row.
func observationID(digest string) string
```

**The extension filter.** The capture reads only files whose extension
is `.flac`, case-insensitively, because the reader reads no other format
and a cover image in the directory is not a track. **Non-FLAC entries
still enter the digest**, because a changed cover means a changed upload
and the digest is the upload's identity, and because excluding them
would make a re-uploaded album with new art look unchanged. A directory
holding no FLAC at all is excluded with reason `ExcludedNoTracks`,
counted and named. A nested directory is not
descended into; the incumbent's consumed layout is
`<inbox>/<uploader>/<top-level entry>` (`importer.go:839-857`,
`:1238-1254`), and one level is what the shadow observes.

- [ ] **Step 1: Write the failing tests**

```go
func TestCaptureReadsOneAlbumPerDirectory(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "windrose", [][]string{
		{"ALBUMARTIST=Cabinet", "ALBUM=Windrose", "TITLE=One", "TRACKNUMBER=1"},
		{"ALBUMARTIST=Cabinet", "ALBUM=Windrose", "TITLE=Two", "TRACKNUMBER=2"},
	})

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 1 {
		t.Fatalf("captures = %d, want 1", len(got.Captures))
	}
	c := got.Captures[0]
	if c.Uploader != "ada" || c.Basename != "windrose" {
		t.Errorf("capture = %s/%s, want ada/windrose", c.Uploader, c.Basename)
	}
	if len(c.Items) != 2 {
		t.Errorf("items = %d, want 2", len(c.Items))
	}
	if !c.CapturedAt.Equal(passNow) {
		t.Errorf("CapturedAt = %v, want the pass clock %v", c.CapturedAt, passNow)
	}
}

func TestCaptureExcludesAMultiAlbumDirectory(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "contributor-zip", [][]string{
		{"ALBUMARTIST=Cabinet", "ALBUM=Windrose", "TITLE=One", "TRACKNUMBER=1"},
		{"ALBUMARTIST=Cabinet", "ALBUM=Windrose", "TITLE=Two", "TRACKNUMBER=2"},
		{"ALBUMARTIST=Thicket", "ALBUM=Understory", "TITLE=One", "TRACKNUMBER=1"},
		{"ALBUMARTIST=Thicket", "ALBUM=Understory", "TITLE=Two", "TRACKNUMBER=2"},
	})

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 0 {
		t.Errorf("captures = %d, want 0; a multi-album directory is excluded, never scored", len(got.Captures))
	}
	if len(got.Excluded) != 1 || got.Excluded[0].Reason != ExcludedMultiAlbum || got.Excluded[0].Groups != 2 {
		t.Errorf("excluded = %+v, want one multi-album row with 2 groups", got.Excluded)
	}
}

func TestCaptureDiscardsAPartialRead(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "half-gone", [][]string{
		{"ALBUMARTIST=Cabinet", "ALBUM=Windrose", "TITLE=One", "TRACKNUMBER=1"},
		{"ALBUMARTIST=Cabinet", "ALBUM=Windrose", "TITLE=Two", "TRACKNUMBER=2"},
		{"ALBUMARTIST=Cabinet", "ALBUM=Windrose", "TITLE=Three", "TRACKNUMBER=3"},
	})
	// Replace one track's bytes with a truncated container, which is
	// what a file renamed out from under the walk looks like to
	// readFLAC: the entry was listed and cannot be read whole.
	corrupt(t, filepath.Join(dir, "ada", "half-gone", "02.flac"))

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 0 {
		t.Fatalf("captures = %d, want 0; a partial read is never stored, never scored, never counted", len(got.Captures))
	}
	if len(got.PartialReads) != 1 || got.PartialReads[0] != "ada/half-gone" {
		t.Errorf("partial reads = %v, want [ada/half-gone]", got.PartialReads)
	}
}

func TestCaptureIsIdempotentUnderAnUnmovedDigest(t *testing.T) {
	// Two capture calls over one directory yield the same digest and the
	// same id; the second, run with a has() that reports the id stored,
	// yields zero captures.
}

func TestCaptureDigestMovesWithEveryFileProperty(t *testing.T) {
	tests := []struct{ name string; mutate func(t *testing.T, dir string) }{
		{"a renamed file", /* ... */},
		{"a resized file", /* ... */},
		{"a touched file", /* ... */},
		{"an added cover image", /* ... */},
	}
	// Each row asserts the digest differs from the baseline. The cover
	// row is why non-FLAC entries enter the digest: a re-uploaded album
	// with new art is a new upload.
}

func TestCaptureWalksInLexicalOrder(t *testing.T) {
	// Three uploaders with three albums each, created in reverse order,
	// come back sorted by uploader then basename. Task 13's budget
	// fixture depends on this (Resolution R18).
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ -run Capture -count=1 -v`
Expected: FAIL, `capture` undefined.

- [ ] **Step 3: Write `capture.go`**

Guard clauses over nesting. Group with a `map[[2]string][]int` over
`(cmp.Or(AlbumArtist, Artist), Album)` and count the keys; a `map` is
legal here because only its **size** reaches an output, never its
iteration order. **Provisional, conductor ruling owed** on the artist
fallback (fold, B4).

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow && git commit`

**Acceptance criteria:**

- One single-album directory yields one capture with every track's
  fields and the pass clock as `CapturedAt`.
- A two-group directory yields zero captures and one excluded row
  carrying the group count and the basename, **for both group shapes**:
  two `ALBUMARTIST` values over two albums, and two `ARTIST` values
  under one album with no `ALBUMARTIST` at all.
- A directory with one unreadable track yields zero captures and one
  partial-read row, and **stores nothing**.
- The digest moves on a rename, a resize, a touch and an added non-FLAC
  file, **with the touch expressed as an `os.Chtimes` to a second fixed
  instant rather than as a sleep**, since `writeAlbum` pins every
  written file's modification time to `fixedModTime` (helper list) and a
  digest that is a function of wall-clock time is not reproducible
  across two temp trees (Resolution R5).
- A stored digest is not re-captured.
- The walk is lexical at both levels, asserted against a reverse-order
  creation.
- No `time.Now` call anywhere in the file or its test.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Store a partial read instead of discarding it →
   `TestCaptureDiscardsAPartialRead` fails with one capture where zero
   is wanted. This is the mutation that matters: a fixture missing a
   track scores a different album distance and manufactures a divergence.
2. Drop the artist fallback, grouping by `(AlbumArtist, Album)` alone →
   **the second exclusion row fails**, reporting one group where two are
   wanted. That row carries four tracks with two distinct `ARTIST`
   values, one `ALBUM`, and **no `ALBUMARTIST` on any track**, which is
   the shape a flat contributor ZIP produces and the shape beets splits
   at `stages.py:102-103`. Revision 1's mutation, grouping by `Album`
   alone, was already caught by the first row and left the fallback with
   no detector (fold, B4).
3. Exclude non-FLAC entries from the digest →
   `TestCaptureDigestMovesWithEveryFileProperty`'s cover row fails.
4. Drop modification time from the digest → the touched-file row fails.
5. Walk with `filepath.WalkDir` and descend → the lexical-order test
   still passes and the single-album test fails, because a nested
   directory's tracks join the parent's item list and change the group
   count.

**Deliverable count: 3** (the lexical walk with its extension filter and
one-level rule; the grouping rule with its artist fallback and its
excluded count; the digest over pinned modification times, the id and
the partial-read discard). Gate green; commit. **Report actual token
spend and the five mutations' outcomes.**

---

## Task 7: The reconcile step, the oracle report decode, the basename join, and the gap ledger

Model: `sonnet`; **the named runner-up for an `opus` upshift** (see
Execution model). Depends on Tasks 4, 5 and 6.

This is where the shadow reads the incumbent's own verdict. Every rule
here costs an observation rather than risking a fabricated one.

**Files:**
- Create: `internal/engine/shadow/oraclereport.go` + `oraclereport_test.go`
- Create: `internal/engine/shadow/reconcile.go` + `reconcile_test.go`

**Interfaces:**
- Consumes: `readFLAC` (Task 4), the store and its types (Task 5), the
  captures (Task 6).
- Produces:

```go
// internal/engine/shadow/oraclereport.go

// The incumbent's run-report shape, decoded rather than imported.
// _salvage/ is outside the module and invisible to the go tool, so
// these declarations mirror _salvage/internal/importer/importer.go:134-172
// field for field and tag for tag. They retire at rung 16 with the
// incumbent binary; nothing else in the greenfield server reads them.
//
// The decoder deliberately does NOT set DisallowUnknownFields: the
// incumbent may add a field, and a shadow that refused to read a newer
// report would go blind rather than degrade.
type oracleRunReport struct {
	Version   int                 `json:"version"`
	RunID     string              `json:"run_id"`
	StartedAt time.Time           `json:"started_at"`
	EndedAt   time.Time           `json:"ended_at"`
	LockHeld  bool                `json:"lock_held"`
	Error     string              `json:"error,omitempty"`
	Albums    []oracleAlbumOutcome `json:"albums"`
}

// oracleAlbumOutcome carries no per-album path, which is the whole basis
// of the basename join (contract S10): SourceDir is the UPLOADER's inbox
// directory (the field at importer.go:136, documented on
// albumUnit.sourceDir at :212, assigned at the uploader level at :766,
// :908 and :1066) and several rows of one report routinely share it.
type oracleAlbumOutcome struct {
	Uploader     string   `json:"uploader"`
	SourceDir    string   `json:"source_dir"`
	Outcome      string   `json:"outcome"`
	SessionID    string   `json:"session_id,omitempty"`
	LibraryPaths []string `json:"library_paths,omitempty"`
	Collection   string   `json:"collection,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// The outcome vocabulary, importer.go:107-125. imported is an apply and
// review is a skip; quarantined and rejected are GATE outcomes that
// never reached beets (importer.go:945, :499-505), so an observation
// carrying one is excluded rather than counted.
const (
	outcomeImported    = "imported"
	outcomeReview      = "review"
	outcomeQuarantined = "quarantined"
	outcomeRejected    = "rejected"
)

// readRunReports reads every report-*.json under stateDir, newest last
// by StartedAt. It opens no other path under stateDir, and in
// particular it opens import.lock NEVER: the incumbent treats a held
// lock as healthy and returns nil without importing
// (importer.go:397-413), so a shadow holding it would stall the box's
// own import while every signal read green.
func readRunReports(stateDir string) ([]oracleRunReport, error)

// readTagLogSkips reads every import-*.log under stateDir and returns
// the basenames beets skipped. Tag logs are written atomically
// (_salvage/internal/beets/beets.go:104) and are never swept, which is
// why they are the skip half's durability backstop. The logged paths
// are STAGING paths (importer.go:1096) and stageUser preserves the
// inbox entry's own name (importer.go:844-856), so filepath.Base of a
// skip line is the captured basename.
func readTagLogSkips(stateDir string) (map[string]bool, error)

// reviewedBasenames reads <reviewDir>/<date>/<uploader>/ and returns,
// per uploader, the album basenames filed there. The match is a suffix
// grammar rather than an equality, because uniqueDestPath appends
// -HHMMSS on a same-day collision and then .2, .3
// (_salvage/internal/review/review.go:76-99). The date component is the
// RUN's StartedAt (importer.go:271), not the album's, so every dated
// directory is read.
func reviewedBasenames(reviewDir string) (map[string][]string, error)

// internal/engine/shadow/reconcile.go

// A reconciled observation and its oracle signal, or the reason it
// could not be resolved.
type reconciliation struct {
	Capture      Capture
	Oracle       *OracleSignal // nil when Unreconciled or Excluded is set
	Unreconciled string
	Excluded     string
	Superseded   bool
}

// reconcile resolves each album outcome in the newly arrived run
// reports to a captured fixture and records the oracle signal
// (contract S10).
//
// It returns the reconciliations, the reports it consumed, and the
// number of gaps it detected. A gap means at least one report was swept
// before this shadow consumed it: the incumbent keeps at most thirty
// reports (importer.go:389, :533-548,
// _salvage/internal/config/config.go:172) and its timer fires *:0/15,
// which is roughly seven and a half hours of history rather than a
// durable ledger.
func reconcile(in reconcileInput) (reconcileResult, error)

type reconcileInput struct {
	Captures       []Capture      // every stored, still-unreconciled fixture
	Reports        []oracleRunReport
	Consumed       []ConsumedReport
	TagLogSkips    map[string]bool
	Reviewed       map[string][]string
	LibraryReadFn  func(path string) (matcher.ItemFields, error) // readFLAC, injected so the test needs no library tree
}

type reconcileResult struct {
	Reconciliations []reconciliation
	Consumed        []ConsumedReport
	Gaps            int
}
```

**The join, per outcome, stated as three rules.**

**`imported` is an apply, resolved by content.** `LibraryPaths` is a
positional correlation rather than a join: the importer fills it by
matching staging order against the beets database's album order, since
"one shared beet invocation carries no per-uploader join key", and a
count mismatch leaves the extra units empty (`importer.go:1119-1136`).
So the shadow reads the **first** file `LibraryPaths` names through
`readFLAC`, takes `MBAlbumID` as the oracle's chosen release, and joins
back to a capture whose uploader matches the row's and whose track
**titles and lengths** match the library file's. The length comparison
is a tolerance rather than an equality, because beets rewrites tags but
not audio: two lengths match when they differ by less than one second.
**An `imported` row with no resolvable release id is `unreconciled`
with reason `imported-with-no-release-id`, never an apply** (Resolution
R22). Two shapes reach it. An empty `LibraryPaths`, which the importer's
own `Error` field documents as "a beets-database match that never turned
up" on an outcome that still reads `imported` (`importer.go:145-149`).
And a named library file the reader opens whose `MUSICBRAINZ_ALBUMID` is
absent, which `readFLAC` returns as `MBAlbumID == ""` under
`ProjectItemFields`' typed-default rule
(`internal/engine/matcher/info.go:178-211`): recording that as an apply
would classify against any Go apply as `go-applies-different-release`,
a divergence against a release the oracle never named, which is exactly
the fabricated row contract S6 forbids.

**`review` is a skip, resolved by basename.** A capture's basename is
resolved when the tag log carries a `skip ` line whose `filepath.Base`
equals it, or when the review tree carries a directory under that
uploader matching `^<basename>(-\d{6}(\.\d+)?)?$`. **A `review` row
carrying a non-empty `Error` is `unreconciled` with reason
`partial-skip`** (Resolution R4), never a clean skip, because
`resolveImported` marks a partly-imported unit `OutcomeReview` with
`Error: "left in staging after import"` (`importer.go:1179-1199`) and
recording that as an oracle skip fabricates a divergence whenever Go
applies.

**`quarantined` and `rejected` are excluded**, counted with reason
`gate-rejected`, and never scored. Gate-rejected material never reaches
beets (`importer.go:945`, `:499-505`), so an observed-album count that
included them would lie.

**Ambiguity and supersession.** When one row resolves to more than one
capture, or one capture matches more than one row, every candidate is
`unreconciled` with reason `ambiguous-basename`. The one exception is
supersession: when several unreconciled captures share one
`(uploader, basename)`, the row resolves to the newest by `CapturedAt`
and the others are marked `Superseded` and counted rather than reported
as ambiguous (Resolution R5). A capture whose own report was already
swept is `unreconciled` with reason `report-swept`, detected by the gap
rule below.

**The gap rule** (Resolution R6). A report is consumed when this pass
reconciled it; the ledger records its file name and its decoded
`StartedAt` and `EndedAt`. A gap is counted when the ledger is non-empty
**and** the oldest report now present has a `StartedAt` later than the
newest consumed `EndedAt`. Every capture older than that boundary and
still unreconciled becomes `report-swept` on the pass that detects the
gap, so a swept oracle signal is a named condition rather than a fixture
that sits unresolved forever. A `LockHeld` report carries no albums
(`importer.go:162-164`) and is consumed like any other, so it advances
the ledger and never counts as a gap.

- [ ] **Step 1: Write the failing tests**

```go
func TestReconcileJoinsAnImportedRowByContentNotBySourceDir(t *testing.T) {
	// Two captures under one uploader, "alpha" and "beta", both with
	// distinct track titles. One run report with TWO imported rows
	// sharing one source_dir, which is the shape stageAllContributors
	// produces (importer.go:782-800) and which a source_dir join maps
	// every fixture to every outcome. Each row's library_paths names a
	// library FLAC carrying that album's own titles and its own
	// MUSICBRAINZ_ALBUMID.
	//
	// Assert: two reconciliations, each carrying the right release
	// mbid, and zero unreconciled.
}

func TestReconcileResolvesASkipFromEitherBackstop(t *testing.T) {
	tests := []struct {
		name     string
		tagLog   []string
		review   []string // directory names under <review>/<date>/<uploader>
		wantOK   bool
	}{
		{name: "tag log alone", tagLog: []string{"/srv/staging/ada/windrose"}, wantOK: true},
		{name: "review tree alone", review: []string{"windrose"}, wantOK: true},
		{name: "review tree with an HHMMSS collision suffix", review: []string{"windrose-142530"}, wantOK: true},
		{name: "review tree with a numeric collision suffix", review: []string{"windrose-142530.2"}, wantOK: true},
		{name: "a different album's directory does not resolve it", review: []string{"windrose-live"}, wantOK: false},
		{name: "neither backstop", wantOK: false},
	}
	// The last two rows are what make the suffix grammar a rule rather
	// than a prefix match: "windrose-live" shares a prefix and is a
	// different album.
}

func TestReconcilePartialSkipIsUnreconciledNotASkip(t *testing.T) {
	// A review row carrying Error: "left in staging after import".
	// Assert reason == UnreconciledPartialSkip, and that no OracleSignal
	// is produced. Recording it as a clean skip would fabricate a
	// divergence whenever Go applies (contract S6).
}

func TestReconcileExcludesGateOutcomes(t *testing.T) {
	// One quarantined row and one rejected row; both excluded with
	// reason ExcludedGateRejected, neither counted as observed.
}

func TestReconcileUnreconciledReasons(t *testing.T) {
	tests := []struct{ name, want string }{
		{"an imported row with no library paths", UnreconciledNoReleaseID},
		{"an imported row whose library file carries no album id", UnreconciledNoReleaseID},
		{"two uploaders with an identically named album in one report", UnreconciledAmbiguousBasename},
		{"a capture whose report was swept", UnreconciledReportSwept},
		{"a review row carrying an error", UnreconciledPartialSkip},
	}
	// Each of the four reasons has its own firing fixture, and the
	// widened one has a fixture for each of its two halves (Resolution
	// R22). A reason with no fixture is a counter that reads zero
	// forever, and a half with no fixture is the same defect narrowed.
	// The second row is the one that would otherwise have produced a
	// go-applies-different-release against a release the oracle never
	// named.
}

func TestReconcileSupersedesAnOlderCaptureOfOneAlbum(t *testing.T) {
	// Two captures of ada/windrose with different digests and different
	// CapturedAt. One imported row. Assert the newer reconciles, the
	// older carries Superseded, and neither is ambiguous.
}

func TestReconcileDetectsAReportGap(t *testing.T) {
	// A ledger whose newest consumed EndedAt precedes the oldest present
	// report's StartedAt. Assert Gaps == 1 and that a still-unreconciled
	// capture older than the boundary comes back UnreconciledReportSwept.
}

func TestReconcileNeverReadsTheImportLock(t *testing.T) {
	// Put an unreadable-by-design marker at <stateDir>/import.lock (mode
	// 0000 is not portable, so the test asserts on the paths the
	// functions opened rather than on an error): readRunReports and
	// readTagLogSkips are given a stateDir instrumented through a
	// recording os.DirFS-shaped seam, and the recorded path list must
	// not contain "import.lock". Risk 3's whole mechanism is that a
	// shadow holding that lock makes the box look green while it imports
	// nothing (importer.go:397-413).
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ -run 'Reconcile|RunReport' -count=1 -v`
Expected: FAIL, `reconcile` undefined.

- [ ] **Step 3: Write `oraclereport.go` and `reconcile.go`**

The lock rule is a named constant and a guarded read, with the reason in
the code:

```go
// importLockName is the incumbent's lock file. It is named here so the
// reader functions can refuse it by name: flock on a read-only
// descriptor succeeds on Linux, and the incumbent treats a held lock as
// healthy, writing a LockHeld report, pinging SUCCESS and returning nil
// without importing (_salvage/internal/importer/importer.go:397-413).
// A shadow holding it would stall the box's own import while every
// signal read green, so this package opens it never.
const importLockName = "import.lock"
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow && git commit`

**Acceptance criteria:**

- Two `imported` rows sharing one `source_dir` resolve to their own
  captures, and the fixtures are authored so a `source_dir` join would
  be ambiguous. **If they are not, the join's correctness is untested**,
  and the task says so rather than passing.
- All four backstop shapes for a skip resolve, and the two negative rows
  do not.
- Each of the four `unreconciled` reasons has its own firing fixture,
  and `imported-with-no-release-id` has one for each half: an empty
  `LibraryPaths`, and a named library file carrying no
  `MUSICBRAINZ_ALBUMID`.
- Gate outcomes are excluded and counted, never observed.
- Supersession resolves to the newest capture and marks the rest.
- A gap is detected from `StartedAt` ordering, and a capture older than
  the boundary becomes `report-swept`.
- `import.lock` is opened nowhere, proven by the recorded path list.
- The decoder tolerates an unknown field in a report.
- No `time.Now` call anywhere.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Join `imported` rows on `SourceDir` →
   `TestReconcileJoinsAnImportedRowByContentNotBySourceDir` fails with
   two ambiguous rows, which is exactly fold B2's finding made
   mechanical.
2. Map a `review` row with a non-empty `Error` to a clean skip →
   `TestReconcilePartialSkipIsUnreconciledNotASkip` fails.
3. Match the review tree by `strings.HasPrefix` instead of the suffix
   grammar → the `windrose-live` row fails, resolving a different album.
4. Count `quarantined` as a skip → `TestReconcileExcludesGateOutcomes`
   fails, and the observed count grows by one album that never reached
   beets.
5. Detect a gap from file modification times instead of `StartedAt` →
   `TestReconcileDetectsAReportGap` fails once the fixture writes its
   reports in an order that disagrees with their `StartedAt`, which the
   fixture does deliberately.
6. Read the import lock, for example by globbing `*` under the state
   directory instead of `report-*.json` →
   `TestReconcileNeverReadsTheImportLock` fails.
7. Take an empty `MBAlbumID` as the oracle's chosen release rather than
   as an unresolvable one → the second `imported-with-no-release-id` row
   fails, producing an `OracleSignal` with an empty `release_mbid` that
   the classifier would later turn into a fabricated
   `go-applies-different-release` against any Go apply (fold, M-g).

**Deliverable count: 4** (the report and tag-log decode with its
lock refusal; the three-rule basename join with its no-release-id rule;
the four `unreconciled` reasons and the two exclusions; the
consumed-report ledger and its gap rule). Gate green; commit. **Report
actual token spend and the seven mutations' outcomes.**

---

## Task 8: The transports, the client, and the match step

Model: `sonnet`. Depends on Tasks 5, 6 and 7.

**Files:**
- Create: `internal/engine/shadow/transport.go` + `transport_test.go`
- Create: `internal/engine/shadow/match.go` + `match_test.go`
- Create: `internal/engine/shadow/config.go`

**Interfaces:**
- Consumes: `tag.Album` (contract S1), `musicbrainz.NewClient`,
  `musicbrainz.NewLimiter`, `musicbrainz.Limiter.Budget`,
  `musicbrainz.Unmetered`, `matcher.QuietImportTunables` (Task 2),
  `matcher.FullCompat`.
- Produces:

```go
// internal/engine/shadow/config.go

// Config is one shadow run's inputs. Every value is a plain parameter:
// this package imports internal/config never, on the same grounds
// internal/engine/matcher's tunables (tunables.go:50-54) and
// internal/engine/musicbrainz's identity (client.go:86-92) state, and
// it reads no clock, so now and evalDate arrive as parameters too.
type Config struct {
	// StateDir is the greenfield state directory. The store lives at
	// StateDir/shadow and nothing outside it is ever written.
	StateDir string

	// ObserveDir, OracleStateDir and ReviewDir are the incumbent's inbox,
	// its own state directory, and its review tree. All three are read
	// only, and OracleStateDir's import.lock is opened never.
	ObserveDir     string
	OracleStateDir string
	ReviewDir      string

	// UserAgent is composed in cmd/dubplate-server, because the identity
	// is site-specific (musicbrainz/client.go:86-92). NewClient refuses
	// an empty one, so the shadow cannot run unidentified.
	UserAgent string

	// MatcherCommit stamps every artifact, so a divergence from a
	// superseded matcher revision is re-evaluated mechanically rather
	// than consuming owner adjudication (ratified ruling B). The
	// subcommand refuses to observe when it is empty or "dev".
	MatcherCommit string

	// BudgetLimit caps one pass's MusicBrainz spend. A limit of zero or
	// less means NO CAP (musicbrainz/limiter.go:130-133), so a caller
	// that means "cap" must pass a positive number.
	BudgetLimit int

	// Interval is the wire pace. The shadow runs at
	// 2 * musicbrainz.DefaultInterval, because the box's own beet import
	// pins its own rate independently and shares the box's single egress
	// IP, where MusicBrainz's limit actually applies (conductor ruling,
	// rung 4 spec).
	Interval time.Duration

	// BaseURL is empty for musicbrainz.DefaultBaseURL. A test points it
	// at an httptest.Server.
	BaseURL string

	// Transport is the round tripper the recorder wraps; nil means
	// http.DefaultTransport.
	Transport http.RoundTripper

	// Timeout is the HTTP client's timeout. It is set here rather than
	// on ClientConfig because ClientConfig.Timeout is ignored entirely
	// when HTTP is non-nil (musicbrainz/client.go:137-144), and the
	// shadow always sets HTTP to install its recorder.
	Timeout time.Duration

	// Limiter is the wire pacer. Nil means Run builds one at Interval
	// and closes it when the pass ends. A test supplies a pre-closed
	// limiter so the ErrLimiterClosed branch is reachable at all, which
	// it otherwise is not: Budget.Wait reserves before it waits
	// (limiter.go:372-381), so a spent cap never consults the limiter
	// and the two deferral branches are mutually exclusive within one
	// pass (Resolution R7).
	Limiter *musicbrainz.Limiter
}

// internal/engine/shadow/transport.go

// recordingTransport records every request's canonical path and query
// and its response body, in the order tag.Album issues them, so an
// observation's stored responses are exactly what the run saw. The
// recorded path_query is already in cassette's canonical form by
// construction, because the client builds every query through
// url.Values.Encode, which sorts by key
// (musicbrainz/client.go:183-200).
type recordingTransport struct {
	base     http.RoundTripper
	requests []RecordedRequest
	bodies   [][]byte
}

// replayTransport answers from an observation's stored responses and
// REFUSES a miss rather than answering 404.
//
// A 404 on a lookup is swallowed inside tag.Album's lookupAndAdd
// (internal/engine/tag/tag.go:166-170), so a miss answered 404 would
// silently drop a candidate and retire a divergence that never actually
// re-ran. The refusal is 501, which doWithRetry's default arm turns
// into a non-retryable error in one attempt
// (musicbrainz/client.go:381-382), and the miss list below is the
// signal callers branch on rather than the error's shape. A silently
// retired divergence is the one outcome worse than an un-retired one
// (rung 4 spec, reevaluate).
type replayTransport struct {
	bodies map[string][]byte // keyed by request URI
	misses []string
}

// internal/engine/shadow/match.go

// matchAlbum drives tag.Album once against a capture's items under
// FullCompat, through a recording transport, spending one token per
// request from budget.
//
// It returns the proposal, the recorded requests and bodies, and any
// error tag.Album returned, unwrapped classification being the caller's
// (contract S5: classify with errors.Is, never equality, because
// tag.Album wraps every client error).
func matchAlbum(ctx context.Context, cfg Config, budget *musicbrainz.Budget,
	items []matcher.ItemFields, evalDate time.Time) (*tag.Proposal, []RecordedRequest, [][]byte, error)

// configHash hashes the Tunables and Compat pair a run scored under.
//
// It is deliberately not named config_hash at the JSON boundary:
// cassette.Expected.ConfigHash is a hash of the Python shim's effective
// config (internal/harness/cassette/cassette.go:80), and conflating the
// two would make a rung 15 comparison silently wrong. The hash runs over
// the JSON rendering of both values, so a field added to either type
// moves the hash without anyone remembering to update this function.
func configHash(tun matcher.Tunables, compat matcher.Compat) string
```

**The client the shadow builds**, stated once so Tasks 9, 12 and 13
build the identical one:

```go
	limiter := cfg.Limiter
	if limiter == nil {
		limiter = musicbrainz.NewLimiter(cfg.Interval)
		defer limiter.Close()
	}
	budget := limiter.Budget("shadow", musicbrainz.ClassBackground, cfg.BudgetLimit)

	rec := &recordingTransport{base: cmp.Or(cfg.Transport, http.RoundTripper(http.DefaultTransport))}
	client, err := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   cfg.BaseURL,
		UserAgent: cfg.UserAgent,
		Budget:    budget,
		HTTP:      &http.Client{Transport: rec, Timeout: cmp.Or(cfg.Timeout, 30*time.Second)},
	})
```

`ClassBackground`'s own comment names this caller: "rung 4's shadow
spends deferred background budget through this limiter and never inline
with the probe" (`musicbrainz/limiter.go:16-18`). The `Timeout` is set
on the `http.Client` rather than on `ClientConfig`, because a non-nil
`HTTP` makes `ClientConfig.Timeout` dead (`client.go:137-144`) and the
zero value of `http.Client.Timeout` is no timeout at all, which on an
unattended timer is a hung pass rather than a failed one.

- [ ] **Step 1: Write the failing tests**

```go
func TestRecordingTransportRecordsInFlowOrder(t *testing.T) {
	// A search returning two ids, then two lookups. Assert three
	// recorded requests, in issue order, each path_query already sorted
	// by key, and three bodies matching what the server sent byte for
	// byte.
}

func TestReplayTransportRefusesAMissWithoutA404(t *testing.T) {
	// A replay client over one recorded release. Ask for a different
	// release id. Assert: the miss list names the request, the error is
	// neither musicbrainz.ErrNotFound nor musicbrainz.ErrUnavailable,
	// and exactly ONE request was attempted, which is what proves 501 is
	// non-retryable and the call does not spend seven attempts.
}

func TestMatchAlbumTakesTheIDConsensusEarlyReturn(t *testing.T) {
	// Items carrying a unanimous mb_albumid, a release body scoring
	// strong. Assert: one recorded request, Proposal.Action is
	// matcher.ActionApply, Candidates[0].Info.AlbumID is the id, and
	// budget.Spent() == 1.
}

func TestMatchAlbumSearchesWhenThereIsNoConsensus(t *testing.T) {
	// Items with no mb_albumid. Assert one search plus one lookup per
	// returned id, in response order, and budget.Spent() equal to the
	// request count.
}

func TestMatchAlbumSpendsOneTokenPerAttemptIncludingRetries(t *testing.T) {
	// TWO collapses are needed, and revision 1 named only one (fold,
	// M-b). The server sends Retry-After: 0, which collapses the retry
	// backoff (retryWait prefers a delta-seconds header,
	// musicbrainz/client.go:404-412; without it maxAttempts of 7 with
	// waits 0+1+2+4+8+16 spends 31 real seconds). And cfg.Interval is
	// time.Millisecond, which collapses the LIMITER's pacing: the
	// budget is metered here (mutation 4 forbids Unmetered), so every
	// attempt calls budget.Wait before it fires
	// (musicbrainz/client.go:346), Limiter.acquire holds at most one
	// spare token and otherwise blocks until the next tick
	// (limiter.go:293-321), and NewLimiter drives that from a ticker at
	// the configured interval (:110-114). At contract S15's production
	// pace of 2 * DefaultInterval, seven attempts need six ticks and
	// the row takes at least twelve seconds.
	//
	// Assert the error wraps musicbrainz.ErrUnavailable,
	// budget.Spent() == 7, and the whole row completes in under a
	// second.
}

func TestMatchAlbumRunsUnderFullCompatAndQuietImportTunables(t *testing.T) {
	// Revision 1 proposed asserting that Action is apply or skip and
	// never asis or ask, on the reasoning that NoneRecAction at its
	// shipped "ask" would not project to skip. That reasoning is wrong
	// (fold, m11): under Quiet, ActionFor returns apply on RecStrong
	// and otherwise the parsed QuietFallback, never reaching Timid or
	// NoneRecAction (matcher/verdict.go:179-210), and QuietFallback
	// already ships "skip" (tunables.go:316). The assertion separates
	// quiet from non-quiet and nothing finer.
	//
	// Assert instead that the observation's stored MatcherConfigHash
	// equals configHash(matcher.QuietImportTunables(),
	// matcher.FullCompat()). That is a whole-configuration identity: it
	// moves the day either input moves, it distinguishes
	// QuietImportTunables from DefaultTunables with Quiet set, and it
	// is machinery this task already builds.
}

func TestConfigHashMovesWithEitherInput(t *testing.T) {
	base := configHash(matcher.QuietImportTunables(), matcher.FullCompat())
	if configHash(matcher.QuietImportTunables(), matcher.DefaultCompat()) == base {
		t.Error("the hash does not move with Compat")
	}
	tun := matcher.QuietImportTunables()
	tun.DistanceWeights["album"] = 4.0
	if configHash(tun, matcher.FullCompat()) == base {
		t.Error("the hash does not move with Tunables")
	}
	if configHash(matcher.QuietImportTunables(), matcher.FullCompat()) != base {
		t.Error("the hash is not stable across two calls with equal inputs")
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ -run 'Transport|Match|ConfigHash' -count=1 -v`
Expected: FAIL, `matchAlbum` undefined.

- [ ] **Step 3: Write `config.go`, `transport.go` and `match.go`**

The recording transport buffers each response body with `io.ReadAll`,
closes the original, and hands the client a fresh
`io.NopCloser(bytes.NewReader(body))`, so the client's own
`io.ReadAll` (`client.go:368`) sees the same bytes the recorder stored.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow && git commit`

**Acceptance criteria:**

- Recorded requests are in flow order with canonical path and query, and
  the bodies round-trip byte for byte.
- A replay miss is recorded, is not a 404, and costs exactly one
  attempt.
- The early return issues exactly one request; the search path issues
  one search plus one lookup per returned id.
- `budget.Spent()` counts every attempt including retries, and the 503
  row completes in under a second **because two things collapse
  together**: `Retry-After: 0` collapses the retry backoff
  (`client.go:404-412`) and `Interval: time.Millisecond` collapses the
  limiter's pacing (`limiter.go:110-114`, `:293-321`). Both are asserted
  rather than assumed, and the wall-time assertion is what proves it.
- **The run's configuration identity is asserted as a hash**, not as a
  verdict shape: the stored `MatcherConfigHash` equals
  `configHash(matcher.QuietImportTunables(), matcher.FullCompat())`.
- The HTTP client carries a non-zero timeout.
- `configHash` moves with either input and is stable across calls.
- No `time.Now` call anywhere, `_test.go` files included.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Answer a replay miss with 404 →
   `TestReplayTransportRefusesAMissWithoutA404` fails, and the failure
   names the candidate `tag.Album` would have dropped silently.
2. Build the client with `ClientConfig{Timeout: ...}` and a non-nil
   `HTTP` → nothing in this suite fails, because the zero timeout only
   bites on a hung server. **Reported as a deliberately unfired
   mutation**, with the guard being the doc comment and the explicit
   `cmp.Or` default; the honest detector is a test with a server that
   never answers, which would spend the test's own timeout and is
   declined for that reason.
3. Record the response body without rewinding it → every match test
   fails, because the client reads an exhausted body and translation
   yields no candidate.
4. Pass `musicbrainz.Unmetered()` instead of the capped budget →
   `TestMatchAlbumSpendsOneTokenPerAttemptIncludingRetries` fails,
   reporting a spend of 0.
5. Hash only `Tunables` → `TestConfigHashMovesWithEitherInput`'s first
   assertion fails, which is the assertion that keeps a compat change
   from being invisible to a rung 15 re-evaluation.
6. Score under `matcher.DefaultTunables()` with `Quiet` set instead of
   `matcher.QuietImportTunables()` →
   `TestMatchAlbumRunsUnderFullCompatAndQuietImportTunables` fails on
   the hash. Revision 1's action-shape assertion could not see this
   difference at all (fold, m11).
7. Leave `cfg.Interval` at contract S15's production pace in the 503 row
   → the wall-time assertion fires at roughly twelve seconds, which is
   the limiter's six ticks rather than the retry backoff (fold, M-b).

**Deliverable count: 4** (`shadow.Config`; the recording transport; the
replay transport with its refusal; the match step and `configHash`).
Gate green; commit. **Report actual token spend and the seven mutations'
outcomes, including mutation 2's non-firing.**

---

## Task 9: The `DefaultCompat` re-score, the divergence classification, and the pass loop

Model: **`opus`** (see Execution model). Depends on Task 8.

This is where the artifact's honesty is decided. Every branch here
either records a true divergence or defers; nothing invents one.

**Provisional, owner ruling 3 open**, for the re-score half. **Provisional,
owner ruling 2 open**, for `PendingRow.ReleaseMBID`.

**Files:**
- Create: `internal/engine/shadow/rescore.go` + `rescore_test.go`
- Create: `internal/engine/shadow/classify.go` + `classify_test.go`
- Create: `internal/engine/shadow/run.go` + `run_test.go`

**Interfaces:**
- Consumes: everything Tasks 5 through 8 produced.
- Produces:

```go
// internal/engine/shadow/rescore.go

// rescoreDefault re-scores one observation's own recorded candidate set
// under DefaultCompat and returns the second verdict.
//
// It drives NO second tag.Album and issues NO search. That is what
// conform's own ModeDefault does, and the reason is that a corrected
// verdict changes the FLOW: the ID-consensus early return fires only at
// matcher.RecStrong (internal/engine/tag/tag.go:96-118, condition at
// :106), Compat.Apply swaps StringDistPatterns
// (internal/engine/matcher/compat.go:51-63), which moves the distance
// and therefore the recommendation, so an album that took the
// one-lookup early return under FullCompat falls through to
// src.SearchReleases (tag.go:125-126) under DefaultCompat, a request
// the recording never captured.
//
// The five steps mirror gradeCaseDefault
// (internal/harness/conform/case.go:568-622): apply and validate, look
// each recorded candidate up through a replay client, cands.Add per
// candidate, cands.Sorted then matcher.Recommend, then matcher.ActionFor.
//
// The replay client takes musicbrainz.Unmetered()
// (musicbrainz/limiter.go:354-359), because the budget sits ABOVE the
// transport: doWithRetry calls budget.Wait before every attempt
// (client.go:346) whatever ClientConfig.HTTP holds, so a replay through
// the metered client would double the ClassBackground spend and double
// the figure the report prints as honest.
//
// The verdict it returns is marked Hypothesis. A re-scored DefaultCompat
// verdict is not the product's true DefaultCompat verdict wherever the
// real flow would have widened the candidate set, and conform's own
// exactness argument (case.go:149-152, the medium rung decided by the
// minimum distance alone) covers ten pinned corpus cases and does not
// generalize to an arbitrary live album.
func rescoreDefault(ctx context.Context, recorded []RecordedRequest, bodies [][]byte,
	albumIDs []string, items []matcher.ItemFields, evalDate time.Time) (*GoVerdict, error)

// internal/engine/shadow/classify.go

// classify returns the divergence class for one reconciled observation,
// or the reason it defers.
//
// The oracle is binary (contract S3), so the comparable quantity is a
// binary projection of Proposal.Action and nothing finer: RecStrong is
// exactly attested, and RecMedium, RecLow and RecNone are
// indistinguishable to the oracle.
//
// NO ERROR OTHER THAN tag.ErrNoCandidates MAY EVER BE RECORDED AS A GO
// SKIP. A truncated album written as go-skips-oracle-applied is a
// fabricated divergence in the one artifact whose whole value is that
// it is honest, and it consumes exactly the owner adjudication ruling
// B's stamp exists to save. Every other error defers, and an error this
// function does not recognize defers with DeferUnclassified rather than
// falling into a verdict arm: a matcher revision that adds a sentinel
// gets a counted deferral, and the counter is what makes the omission
// visible.
//
// Every branch tests with errors.Is, never equality, because tag.Album
// wraps every client error in a message (tag.go:128, :170).
func classify(oracle OracleSignal, p *tag.Proposal, err error) (class, deferred string)

// internal/engine/shadow/run.go

// Run executes one observe pass: capture, reconcile, match.
//
// now is the pass clock and every timestamp the pass writes derives from
// it, which is what TestEngineNeverReadsTheClock requires
// (internal/harness/conform/arch_test.go:152-157) and what makes a
// whole pass deterministic under test.
//
// A pass is SUCCESSFUL when all three steps complete without a fatal
// error, a fatal error being a store write failure, an unusable state
// directory, or an invalid configuration. A pass that captures nothing,
// reconciles nothing and observes nothing is successful: the clock must
// start at the deploy rather than at the first upload, because the floor
// measures the observation window and not the arrival rate. A deferred
// observation is an ordinary branch and never makes a pass unsuccessful.
func Run(ctx context.Context, cfg Config, now time.Time) (*PassResult, error)

// PassResult is one pass's record plus the two artifacts
// cmd/dubplate-server marshals. This package may import no harness
// package (contract S19), so the cassette-shaped fixture and the
// AdjudicationEntry-shaped row are described here in engine types and
// written there.
type PassResult struct {
	Pass         PassRecord
	// Observations are in CAPTURE order, (uploader, basename), which
	// is the order the match step ran them in and therefore the order
	// the budget was spent in. It is a value here rather than a
	// property of the store, because an observation id is a digest
	// prefix (Resolution R5) and id order says nothing about capture
	// order (Resolution R18).
	Observations []*Observation
	Fixtures     []CaseFixture
	Pending      []PendingRow
	ClockStarted bool
}

// A CaseFixture is what cmd/dubplate-server needs to write
// observations/<obs-id>/case.json in cassette.Case's shape, with
// Expected nil.
//
// Profile is "estate", the corpus profile whose reduction equals
// matcher.QuietImportTunables value for value (corpus/profiles/estate.yaml:16-19).
// EvalDate is YYYY-MM-DD, the form every committed case carries.
type CaseFixture struct {
	ObservationID string
	Profile       string
	EvalDate      string
	Items         []map[string]any
	Requests      []RecordedRequest
}

// A PendingRow is what cmd/dubplate-server needs to write
// pending/<obs-id>.json in analysis.AdjudicationEntry's shape.
// Disposition and Adjudicator are not fields here: they are constants
// the marshaller supplies, because disposition is owner-only and every
// row the shadow writes is pending-adjudication by adjudicator agent
// (internal/harness/analysis/adjudication.go:198-200).
type PendingRow struct {
	ID                string
	Album             string
	ReleaseMBID       string
	MatcherCommit     string
	MatcherConfigHash string
	Suspected         string
	Citation          string
	ReleaseTruth      string
	Reason            string
	Date              string
}
```

**The classification table, as the code must implement it** (contract
S4, provisional ruling 3 for the second verdict's annotation):

| Oracle `Signal` | Go outcome | Result |
| --- | --- | --- |
| `apply`, release R | `ActionApply`, `Candidates[0].Info.AlbumID == R` | `ClassAgreement` |
| `skip` | `ActionSkip`, or `errors.Is(err, tag.ErrNoCandidates)` | `ClassAgreement` |
| `skip` | `ActionApply`, release G | `ClassGoAppliesOracleSkipped` |
| `apply`, release R | `ActionSkip`, or `errors.Is(err, tag.ErrNoCandidates)` | `ClassGoSkipsOracleApplied` |
| `apply`, release R | `ActionApply`, release G with `R != G` | `ClassGoAppliesDifferentRelease` |
| any | `errors.Is(err, musicbrainz.ErrUnavailable)` | defer, `DeferUnavailable` |
| any | `errors.Is(err, musicbrainz.ErrBudgetExhausted)` | defer, `DeferBudgetExhausted` |
| any | `errors.Is(err, musicbrainz.ErrLimiterClosed)` | defer, `DeferLimiterClosed` |
| any | any other non-nil error | defer, `DeferUnclassified` |

**`ReleaseMBID` on a pending row** is the release under dispute, filled
by whichever side proposed one (ruling 2, provisional): the oracle's
release for `go-skips-oracle-applied` and `go-applies-different-release`,
the Go matcher's `Candidates[0].Info.AlbumID` for
`go-applies-oracle-skipped`. Every divergence has exactly one release
under dispute, and the observation record says which side proposed it,
so nothing is lost and the pair rule stays a check that can fire.

**The four prose fields are rendered from the observation, never emitted
as `TODO`** (`spec:707`). `Suspected` names the class and the two
verdicts; `ReleaseTruth` states which release is under dispute and that
truth is unruled; `Reason` names the matcher commit and the pass; and
`Citation` is the observation id and its class, which is a convention
rather than a check, since `checkAdjudicationEntry` never tests
`Citation` (`adjudication.go:176-190`; this is the one correction rung 4
does not inherit from the 3d handoff's wording).

**The pass loop**, in order, so a later reader does not have to infer
it:

1. Open the store. Read the consumed-report ledger and every stored,
   still-unreconciled observation.
2. **Capture.** Walk the observe directory, skipping any digest the
   store already holds. Store each new capture as an observation with no
   oracle signal and no verdict.
3. **Reconcile.** Read the new run reports, the tag logs and the review
   tree; resolve each row; write the oracle signal, the unreconciled
   reason, **the exclusion (`Observation.Excluded`, which a
   gate-rejected album reaches only here, after the capture already
   stored it)** or the supersession onto the stored observations;
   replace the ledger.
4. **Match.** For each newly reconciled observation, **in capture
   order, `(uploader, basename)`** (Resolution R18; revision 1 said
   lexical id order, which is a digest sort and bears no relation to the
   order the capture walked or the order a fixture chooses), drive
   `matchAlbum`, then `rescoreDefault`, then `classify`.
   Store the responses, the verdicts and the class. Build a
   `CaseFixture` for every matched observation and a `PendingRow` for
   every news-class divergence.
5. Write the pass record. Start the clock if it is unstarted.

**"News" is every one of the three divergence classes**, with one
annotation rather than one exclusion: a divergence where `DefaultCompat`
agrees with the oracle while `FullCompat` does not is a known
compat-class divergence, and the pending row's `Suspected` says so and
names the hypothesis. It still gets a row, because the annotation is a
hypothesis rather than a ruling, and suppressing a row on a hypothesis
is exactly the silent loss ruling 3's own limit paragraph warns about.

- [ ] **Step 1: Write the failing tests**

```go
func TestClassifyEveryOutcome(t *testing.T) {
	const (
		relR = "11111111-1111-1111-1111-111111111111"
		relG = "22222222-2222-2222-2222-222222222222"
	)
	apply := func(id string) *tag.Proposal { /* ActionApply with Candidates[0].Info.AlbumID == id */ }
	skip := func() *tag.Proposal { /* ActionSkip */ }

	tests := []struct {
		name         string
		oracle       OracleSignal
		proposal     *tag.Proposal
		err          error
		wantClass    string
		wantDeferred string
	}{
		{"same release agrees", OracleSignal{Signal: "apply", ReleaseMBID: relR}, apply(relR), nil, ClassAgreement, ""},
		{"both skip agrees", OracleSignal{Signal: "skip"}, skip(), nil, ClassAgreement, ""},
		{"no candidates agrees with a skip", OracleSignal{Signal: "skip"}, nil, fmt.Errorf("tag: %w", tag.ErrNoCandidates), ClassAgreement, ""},
		{"go applies where the oracle skipped", OracleSignal{Signal: "skip"}, apply(relG), nil, ClassGoAppliesOracleSkipped, ""},
		{"go skips where the oracle applied", OracleSignal{Signal: "apply", ReleaseMBID: relR}, skip(), nil, ClassGoSkipsOracleApplied, ""},
		{"no candidates where the oracle applied", OracleSignal{Signal: "apply", ReleaseMBID: relR}, nil, fmt.Errorf("tag: %w", tag.ErrNoCandidates), ClassGoSkipsOracleApplied, ""},
		{"different release", OracleSignal{Signal: "apply", ReleaseMBID: relR}, apply(relG), nil, ClassGoAppliesDifferentRelease, ""},
		{"unavailable defers", OracleSignal{Signal: "apply", ReleaseMBID: relR}, nil, fmt.Errorf("tag: looking up x: %w", musicbrainz.ErrUnavailable), "", DeferUnavailable},
		{"budget exhausted defers", OracleSignal{Signal: "apply", ReleaseMBID: relR}, nil, fmt.Errorf("tag: %w", musicbrainz.ErrBudgetExhausted), "", DeferBudgetExhausted},
		{"limiter closed defers", OracleSignal{Signal: "skip"}, nil, fmt.Errorf("tag: %w", musicbrainz.ErrLimiterClosed), "", DeferLimiterClosed},
		{"an unrecognized error defers rather than becoming a skip", OracleSignal{Signal: "apply", ReleaseMBID: relR}, nil, errors.New("truncated album"), "", DeferUnclassified},
	}
	// Each row asserts both returns. The last row is contract S6 made
	// mechanical: an unrecognized error must never produce
	// ClassGoSkipsOracleApplied.
}

func TestClassifyUsesErrorsIsRatherThanEquality(t *testing.T) {
	// The three sentinels, each double-wrapped, in tag.Album's own two
	// wrapping shapes ("tag: looking up %s: %w" at tag.go:170 and
	// "tag: searching: %w" at :128). An equality test passes the bare
	// sentinel rows and fails every wrapped one, which is why the rows
	// above are all wrapped.
}

func TestRescoreDefaultIssuesNoSearchAndSpendsNoToken(t *testing.T) {
	// One recorded search plus two recorded lookups. rescoreDefault over
	// the two recorded album ids. Assert: the replay transport saw
	// exactly two requests, both lookups, and the miss list is empty.
	//
	// The no-token half is asserted at Run level rather than here,
	// because rescoreDefault takes no budget, client or Config and
	// there is no pass budget in this scope to read (fold, M-c):
	// TestRunSpendsNoTokenOnTheRescore runs one pass containing one
	// re-scored observation and asserts PassRecord.Budget.Spent equals
	// mode one's own request count.
}

func TestRunSpendsNoTokenOnTheRescore(t *testing.T) {
	// One pass containing exactly one re-scored observation. Assert
	// PassRecord.Budget.Spent equals mode one's own recorded request
	// count, so a replay through the metered client would show up as a
	// doubled figure. This is the assertion rescoreDefault's own test
	// cannot make, because that function takes no budget and no Config
	// (fold, M-c).
}

func TestRescoreDefaultMarksItsVerdictAsAHypothesis(t *testing.T) {
	// The returned GoVerdict carries Hypothesis true. The annotation is
	// the honest statement of ruling 3's own limit.
}

func TestRescoreDefaultReportsAMissRatherThanDroppingACandidate(t *testing.T) {
	// Recorded bodies covering one of two candidate ids. Assert an
	// error naming the missing id, and that no verdict is returned:
	// a partial re-score is a verdict computed over a candidate set
	// the run never had.
}

func TestRunIsDeterministicUnderAFrozenClock(t *testing.T) {
	// Two Run calls over two fresh temp trees with identical inputs and
	// the same now yield byte-identical observation ids, pass ids and
	// stored records.
	//
	// This is reachable only because every fixture writer pins its
	// files' modification times with os.Chtimes to the package-level
	// fixedModTime (helper list): the capture digest ingests
	// modification time (Resolution R5), so two temp trees written at
	// two wall-clock instants would otherwise yield different digests
	// and different ids, and the test would fail on its first
	// assertion for a reason unrelated to the code under test (fold,
	// B3).
}

func TestRunMatchesInCaptureOrder(t *testing.T) {
	// Three albums under one uploader whose digest-derived ids sort in
	// a different order than their basenames. Assert
	// res.Observations comes back in (uploader, basename) order, and
	// that the recorded request sequence follows it. The exit gate's
	// budget-cap fixture depends on this: a cap bites the album the
	// pass reaches last, and a digest sort makes that album arbitrary
	// (Resolution R18).
}

func TestRunStartsTheClockOnAPassThatObservesNothing(t *testing.T) {
	// An empty observe directory and no run report. Assert the pass
	// succeeds, ClockStarted is true, and started-at holds now
	// (Resolution R11).
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ -run 'Classify|Rescore|Run' -count=1 -v`
Expected: FAIL, `classify` undefined.

- [ ] **Step 3: Write `rescore.go`, `classify.go` and `run.go`**

`classify` is a `switch` with guard clauses and no `else`. The deferral
tests come **first**, because an error means there is no verdict to
compare, and the default arm is `DeferUnclassified`.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow && git commit`

**Acceptance criteria:**

- All eleven classification rows pass, each sentinel wrapped in
  `tag.Album`'s own wrapping shape.
- An unrecognized error defers and never produces a class.
- The re-score issues no search, spends no token from the pass budget,
  marks its verdict as a hypothesis, and errors rather than scoring a
  partial candidate set.
- `Run` is deterministic under a frozen clock, and a pass observing
  nothing is successful and starts the clock.
- Every stored observation carries exactly one of `Class`, `Deferred`,
  `Unreconciled` and `Excluded`, or none while it awaits its oracle
  signal, asserted as an invariant over the store after a run.
- **`Run` matches in capture order**, asserted against a fixture whose
  digest-derived ids sort differently from its basenames.
- No `time.Now` call anywhere.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Make the default arm return `ClassGoSkipsOracleApplied` instead of
   deferring → the last classification row fails. This is contract S6's
   own failure, and it is the mutation whose absence would cost owner
   adjudication on a fabricated divergence.
2. Compare with `==` instead of `errors.Is` →
   `TestClassifyUsesErrorsIsRatherThanEquality` fails on every wrapped
   row, and `TestClassifyEveryOutcome`'s eight wrapped rows fail with
   it.
3. Classify on `Recommendation != RecNone` instead of on the action →
   a table row carrying a `RecMedium` proposal against an oracle skip
   fails, reporting `go-applies-oracle-skipped` where `agreement` is
   wanted. **Revision 1's mutation, comparing recommendations instead of
   actions, cannot fire and is replaced** (fold, M-c): under the estate
   tunables `ActionFor` returns apply if and only if `rec == RecStrong`
   and otherwise the pinned `skip` fallback
   (`matcher/verdict.go:179-210`), so `Action == ActionApply` and
   `Recommendation == RecStrong` are the same predicate and no input
   can separate them. `RecMedium` against `RecNone` is a predicate the
   estate tunables genuinely do separate.
4. Drive a second `tag.Album` in the re-score instead of re-scoring →
   `TestRescoreDefaultIssuesNoSearchAndSpendsNoToken` fails with a
   search the recording never captured, which is fold B3's finding made
   mechanical.
5. Hand the replay client the pass budget instead of `Unmetered()` →
   **a `Run`-level row fails on `PassRecord.Budget.Spent`**, reporting
   roughly twice mode one's request count where the recorded request
   count is wanted. The assertion belongs at `Run` rather than inside
   `TestRescoreDefaultIssuesNoSearchAndSpendsNoToken`, because
   `rescoreDefault`'s signature takes no budget, client or `Config` and
   there is no pass budget in that test's scope to assert on (fold,
   M-c). `TestRescoreDefaultIssuesNoSearchAndSpendsNoToken` keeps its
   two reachable assertions: the replay transport saw exactly the
   recorded lookups and no search, and its miss list is empty.
6. Drop the `Hypothesis` mark → `TestRescoreDefaultMarksItsVerdictAsAHypothesis`
   fails, and the row would present a derived figure as a measurement.
7. Suppress a pending row when the `DefaultCompat` re-score agrees with
   the oracle → a run test row fails with zero pending rows where one is
   wanted, which is the silent loss ruling 3's limit paragraph warns
   about.

**Deliverable count: 4** (the `DefaultCompat` re-score with its
`Unmetered` replay client and its hypothesis mark; the classification
with its four deferring branches and its closed default; the pass loop
with its five steps and its success definition; `CaseFixture` and
`PendingRow` as the marshalling boundary's engine-side shapes). Gate
green; commit. **Report actual token spend and the seven mutations'
outcomes.**

---

## Task 10: `analysis.AdjudicationEntry`'s two fields and their checks

Model: `sonnet`. Depends on Task 0. Independent of Tasks 1 to 9.

**Provisional, owner ruling 2 open**, for the `ReleaseMBID` doc-comment
restatement.

**Files:**
- Modify: `internal/harness/analysis/adjudication.go` (the two fields,
  the two rules, one doc-comment restatement)
- Modify: `internal/harness/analysis/adjudication_test.go` (append rows
  to `TestCheckAdjudications`' inline table)

**Interfaces:**
- Consumes: nothing new.
- Produces: two fields on an existing exported type.

```go
// MatcherCommit is the Go matcher revision a shadow divergence was
// observed under. It is required on a shadow-subject row and refused on
// a case-subject row: a divergence from a superseded matcher revision is
// re-evaluated mechanically rather than consuming owner adjudication
// (ratified ruling B), and a corpus case has no matcher revision to
// stamp, because the corpus records the ORACLE's answer.
MatcherCommit string `json:"matcher_commit,omitempty"`

// MatcherConfigHash is a hash over the Tunables and Compat pair the run
// scored under. It is deliberately not named config_hash: 
// cassette.Expected.ConfigHash hashes the Python shim's effective
// config (cassette.go:80), and conflating the two would make a rung 15
// comparison silently wrong.
MatcherConfigHash string `json:"matcher_config_hash,omitempty"`
```

**The `ReleaseMBID` restatement** (ruling 2, provisional). Its enclosing
doc comment already reads "an album plus release MBID for a live shadow
divergence rung 4 will write" (`adjudication.go:32-34`). Rung 4 adds one
sentence naming what the MBID **is**: the release under dispute, filled
by whichever side proposed one, the oracle's when the oracle applied and
the Go matcher's when the oracle skipped and Go applied.

**The two rules**, in `checkAdjudicationEntry`, placed with the pair rule
they extend (`adjudication.go:161-174`):

```go
	if e.Subject.Album != "" && e.Subject.ReleaseMBID != "" {
		if e.MatcherCommit == "" {
			errs = append(errs, fmt.Errorf("adjudications: %s: matcher_commit: missing on a shadow-divergence row", e.ID))
		}
	} else if e.MatcherCommit != "" || e.MatcherConfigHash != "" {
		errs = append(errs, fmt.Errorf("adjudications: %s: matcher_commit and matcher_config_hash belong to a shadow-divergence row, not a case row", e.ID))
	}
```

- [ ] **Step 1: Write the failing tests**

Append rows to `TestCheckAdjudications`' inline table
(`adjudication_test.go:10-113`), which is the package's own landed
precedent for a per-row rule (Resolution R14):

```go
		{
			name: "a shadow-subject row with no matcher commit is refused",
			entry: func() AdjudicationEntry {
				e := base()
				e.Subject = AdjudicationSubject{Album: "Windrose", ReleaseMBID: "1111"}
				return e
			}(),
			wantErr: "matcher_commit",
		},
		{
			name: "a shadow-subject row carrying a matcher commit is accepted",
			entry: func() AdjudicationEntry {
				e := base()
				e.Subject = AdjudicationSubject{Album: "Windrose", ReleaseMBID: "1111"}
				e.MatcherCommit = "abc1234"
				e.MatcherConfigHash = "deadbeef"
				return e
			}(),
		},
		{
			name: "a case-subject row carrying a matcher commit is refused",
			entry: func() AdjudicationEntry {
				e := base()
				e.MatcherCommit = "abc1234"
				return e
			}(),
			wantErr: "matcher_commit",
		},
		{
			name: "a case-subject row carrying a matcher config hash alone is refused",
			entry: func() AdjudicationEntry {
				e := base()
				e.MatcherConfigHash = "deadbeef"
				return e
			}(),
			wantErr: "matcher_config_hash",
		},
```

The implementer reads the table's real field names and helper shape
(`adjudication_test.go:10-30`) and adapts; the four rows are what
matters.

And one round-trip test proving the committed corpus is unmoved:

**Revision 1 asked for a byte-for-byte re-marshal. That cannot pass,
for two reasons neither of which is `omitempty`** (fold, B2). The
committed ledger's key order is `id, subject, suspected, citation,
release_truth, disposition, reason, adjudicator, date,
corrected_verdict[, defect_free]`, while `AdjudicationEntry` declares
`CorrectedVerdict` before `Reason` (`adjudication.go:46-67`) and
`encoding/json` emits declaration order, so a re-marshal first diverges
at line 12. And the file carries one literal `<` and three literal `>`,
which Go's encoder escapes by default; `internal/harness/cassette/cassette.go:9-11`
already records that hazard. Reordering the struct is a schema change to
a shared corpus type outside this pass's scope, and relaxing the
comparison to a bare semantic one takes mutation 3's only detector with
it. **The test asserts the guarantee that actually holds, plus the
absence mutation 3 needs:**

```go
func TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost(t *testing.T) {
	// Load corpus/analysis/adjudications.json's own bytes, decode into
	// AdjudicationLedger, re-encode through a json.Encoder with
	// SetEscapeHTML(false) and SetIndent("", "  ") (Encode appends the
	// trailing newline the file carries), then compare the original
	// bytes and the re-encoded bytes as decoded map[string]any trees.
	//
	// A tree comparison is order-independent, which the file's own key
	// order requires, and it still catches a key gained or lost, which
	// is what omitempty buys. Byte identity was never available and the
	// plan says so rather than asserting it.
	//
	// Then assert on the RAW bytes that they contain neither
	// "matcher_commit" nor "matcher_config_hash". That is mutation 3's
	// real detector: dropping omitempty adds two empty keys to all 21
	// rows, which both the tree comparison and this substring check
	// see.
	//
	// This is a read of the ledger's own bytes, not of a corpus case,
	// and it writes nothing.
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/harness/analysis/ -count=1 -v`
Expected: FAIL, `MatcherCommit` undefined.

- [ ] **Step 3: Implement**

Add the two fields beside the existing ones (`adjudication.go:46-67`),
the two rules, and the `ReleaseMBID` sentence.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/harness/... -count=1`
Expected: PASS, including `internal/harness/corpus_test.go`'s
gate-resident ledger checks over the 21 committed rows.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/harness/analysis && git commit`

**Acceptance criteria:**

- A shadow-subject row without `MatcherCommit` is refused, naming the
  field.
- A shadow-subject row with it is accepted.
- A case-subject row carrying either field is refused, and **both fields
  have their own firing row**, so neither rule rides on the other.
- The 21 committed rows validate, round-trip with no key gained or
  lost, and the file's raw bytes carry neither new key. **Byte-for-byte
  identity is not asserted**, because the ledger is not in Go struct
  field order and Go escapes the `<` and `>` it carries; Task 15 drafts
  the `spec:845-846` correction.
- `internal/harness/corpus_test.go` is untouched and green.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Drop the shadow-row requirement → the "no matcher commit" row fails.
2. Check only `MatcherCommit` on the case-row refusal → the
   `matcher_config_hash` alone row fails, which is why that row exists.
3. Drop `omitempty` from either tag →
   `TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost` fails twice
   over, on the tree comparison and on the raw-bytes absence check, with
   two empty keys on all 21 rows.

**Deliverable count: 2** (the two fields with their `omitempty` tags; the
two checks with a firing row each, plus the `ReleaseMBID` restatement).
Gate green; commit. **Report actual token spend and the three mutations'
outcomes.**

---

## Task 11: The marshalling in `cmd/dubplate-server`

Model: `sonnet`. Depends on Tasks 9 and 10.

`internal/engine/shadow` may import no harness package, so the two
harness-shaped artifacts are written here (contract S19).

**Files:**
- Modify: `cmd/dubplate-server/shadow.go` (Task 3 created it; this task
  adds the two writers and the validator)
- Create: `cmd/dubplate-server/shadow_test.go`

**Interfaces:**
- Consumes: `shadow.PassResult`, `shadow.CaseFixture`,
  `shadow.PendingRow`, `shadow.Store` (Tasks 5 and 9);
  `cassette.Case`, `cassette.Input`, `cassette.ItemFields`,
  `cassette.Request`; `analysis.AdjudicationEntry`,
  `analysis.AdjudicationSubject`, `analysis.AdjudicationLedger`,
  `analysis.CheckAdjudications`, `analysis.Pending`,
  `analysis.AgentAdjudicator`.
- Produces:

```go
// cmd/dubplate-server/shadow.go

// caseJSON renders one shadow fixture in cassette.Case's shape, with
// Expected nil. internal/engine/shadow may import no harness package
// (internal/harness/conform/arch_test.go:141-146), so the engine names
// the fields and this function names the type.
//
// It marshals rather than calling cassette.Case.Write, because Write
// uses os.WriteFile (cassette.go:237-252) and every file the shadow
// store holds goes through atomicfile. What matters is that the written
// file loads through cassette.LoadCase, whose canonicalization rule
// (cassette.go:15-25, :168-170) is exactly what a rung 15 consumer
// applies to this output.
func caseJSON(f shadow.CaseFixture) ([]byte, error)

// pendingJSON renders one staged adjudication row and VALIDATES it
// before returning it.
//
// analysis.CheckAdjudications is the exported entry (adjudication.go:138)
// and its caseIDs argument receives nil, which is sound because a
// shadow-subject row never reaches the case-resolution branch
// (adjudication.go:164-167 is that argument's only reader). Validating
// here means a malformed row fails on the box within one pass rather
// than weeks later when the owner tries to commit it.
//
// Disposition and Adjudicator are constants rather than inputs:
// disposition is owner-only, and a non-Pending disposition set by the
// agent adjudicator is a gate error (adjudication.go:198-200).
func pendingJSON(row shadow.PendingRow) ([]byte, error)

// storeArtifacts writes every fixture and every pending row a pass
// produced into the store.
func storeArtifacts(s *shadow.Store, res *shadow.PassResult) error
```

- [ ] **Step 1: Write the failing tests**

The test feeds the writers a hand-constructed `shadow.PassResult`
literal (Resolution R17), which is the honest test for code whose whole
job is turning engine-side results into two files.

```go
func TestCaseJSONRoundTripsThroughLoadCase(t *testing.T) {
	dir := t.TempDir()
	f := shadow.CaseFixture{
		ObservationID: "shadow-0123456789abcdef",
		Profile:       "estate",
		EvalDate:      "2026-09-10",
		Items: []map[string]any{
			{"title": "One", "artist": "Cabinet", "album": "Windrose", "track": 1, "length": 210.0},
		},
		Requests: []shadow.RecordedRequest{
			{Method: "GET", PathQuery: "/ws/2/release/1111?fmt=json&inc=artists", Response: "responses/001.json"},
		},
	}
	data, err := caseJSON(f)
	if err != nil {
		t.Fatalf("caseJSON: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "case.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	// LoadCase checks every path_query is already canonical
	// (cassette.go:168-170) and that every named response file is
	// readable (:225), so the fixture writes a body too.
	writeResponse(t, dir, "responses/001.json")

	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase on the shadow's own output: %v", err)
	}
	if c.Expected != nil {
		t.Error("Expected is non-nil; a shadow fixture carries no oracle record")
	}
	if c.Profile != "estate" || c.EvalDate != "2026-09-10" {
		t.Errorf("profile/eval_date = %q/%q, want estate/2026-09-10", c.Profile, c.EvalDate)
	}
	if len(c.Requests) != 1 || c.Requests[0].Response != "responses/001.json" {
		t.Errorf("requests = %+v", c.Requests)
	}
}

func TestPendingJSONValidatesBeforeItReturns(t *testing.T) {
	tests := []struct {
		name    string
		row     shadow.PendingRow
		wantErr string
	}{
		{
			name: "a complete row loads and validates",
			row:  completeRow(),
		},
		{
			name:    "an empty release mbid is refused by the pair rule",
			row:     rowWithout(t, "ReleaseMBID"),
			wantErr: "album and release_mbid",
		},
		{
			name:    "an empty matcher commit is refused by rung 4's own rule",
			row:     rowWithout(t, "MatcherCommit"),
			wantErr: "matcher_commit",
		},
		{
			name:    "an empty reason is refused",
			row:     rowWithout(t, "Reason"),
			wantErr: "reason",
		},
	}
	// Every failing row asserts the error names the field. The first row
	// additionally decodes the output and asserts Disposition is
	// analysis.Pending and Adjudicator is analysis.AgentAdjudicator,
	// which the row itself never sets.
}

func TestPendingJSONNeverEmitsATodo(t *testing.T) {
	// The four prose fields are rendered from the observation
	// (rung 4 spec, the staged pending row's table). Assert the
	// marshalled bytes contain no "TODO" and that suspected,
	// release_truth, reason and citation are all non-empty.
}

func TestStoreArtifactsWritesOneFilePerObservationAndOnePerNewsRow(t *testing.T) {
	// A PassResult with three observations, two of them divergences.
	// Assert three case.json files, two pending rows, and that the
	// agreement observation has no pending row.
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./cmd/dubplate-server/ -count=1 -v`
Expected: FAIL, `caseJSON` undefined.

- [ ] **Step 3: Implement**

`caseJSON` marshals with `json.MarshalIndent(c, "", "  ")` plus a
trailing newline, the same indent
`internal/harness/conform/case.go:478` uses for the printed row, so the
adjudication stream reads one shape whether a row came from the corpus
runner or from the box.

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/dubplate-server/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add cmd/dubplate-server && git commit`

**Acceptance criteria:**

- A written `case.json` loads through `cassette.LoadCase` with
  `Expected` nil, its `path_query` accepted as canonical without a call
  to the unexported canonicalizer.
- A pending row fails `analysis.CheckAdjudications` **before** it is
  written, with a firing fixture per rule.
- `Disposition` is `analysis.Pending` and `Adjudicator` is
  `analysis.AgentAdjudicator` on every row, set by the marshaller rather
  than by the caller.
- No `TODO` reaches a staged row.
- `internal/engine/shadow` gains no harness import;
  `TestEngineNeverImportsHarness` stays green.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Skip the validation call → `TestPendingJSONValidatesBeforeItReturns`
   fails on all three refusal rows, and a malformed row would sit on the
   box for weeks.
2. Supply `analysis.OracleWrong` in place of `analysis.Pending` → the
   first row's decode assertion fails, and `CheckAdjudications` refuses
   it besides, because a non-`Pending` disposition set by
   `AgentAdjudicator` is a gate error (`adjudication.go:198-200`).
   Revision 1's wording, "set `Disposition` from the row", names a field
   `shadow.PendingRow` deliberately does not have (fold, m17).
3. Emit `TODO` for `ReleaseTruth` → `TestPendingJSONNeverEmitsATodo`
   fails, and the row would consume owner attention to say nothing.
4. Write `case.json` with a non-canonical `path_query`, for example by
   re-encoding the query with unsorted keys →
   `TestCaseJSONRoundTripsThroughLoadCase` fails inside `LoadCase`.

**Deliverable count: 3** (`caseJSON` with its round trip; `pendingJSON`
with its validation and its two constants; `storeArtifacts`). Gate
green; commit. **Report actual token spend and the four mutations'
outcomes.**

---

## Task 12: The `shadow` subcommand's three verbs and the launch-gate report

Model: `sonnet`. Depends on Tasks 3, 9 and 11.

**Files:**
- Create: `internal/engine/shadow/report.go` + `report_test.go`
- Modify: `cmd/dubplate-server/shadow.go` (the cobra command and the
  three verbs)
- Modify: `cmd/dubplate-server/main.go` (the `AddCommand` block, and
  `matcherCommit`)
- Modify: `cmd/dubplate-server/shadow_test.go`

**Interfaces:**
- Consumes: `shadow.Run`, `shadow.Store`, `config.Load`,
  `config.ParseDuration`, `config.ParseInt`.
- Produces:

```go
// internal/engine/shadow/report.go

// A Report is the launch-gate report's computed rows. It is computed
// here and PRINTED in cmd/dubplate-server, in the pinned "%-32s %s"
// row form doctor already uses (cmd/dubplate-server/doctor.go:64-66),
// so an operator reads one table shape across the binary.
//
// Two rows the spec's table names are deliberately absent, because the
// box cannot know them: divergences by owner disposition, and false
// confident matches net of oracle-wrong rulings. Owner dispositions
// live in corpus/analysis/adjudications.json on the workstation, which
// the box never reads. The report prints what the store knows, and the
// disposition join is the launch gate's own, run on the workstation
// against the committed corpus (Resolution R21).
type Report struct {
	ClockStart      time.Time
	ClockStarted    bool
	Floor           time.Duration
	Elapsed         time.Duration
	ObservedApplied int
	ObservedSkipped int
	Excluded        map[string]int
	Unreconciled    map[string]int
	ReportGaps      int
	Deferred        map[string]int
	Divergences     map[string]int
	GoAppliesOracleSkipped int
	StagedPending   int
	Resolved        int
	MatcherCommits  map[string]int
	Passes          []BudgetRecord
	BudgetTotal     int
	StoreBytes      int64
}

// BuildReport computes the report from the store alone. It reads no
// clock: now is the parameter the elapsed row is measured against.
func BuildReport(s *Store, now time.Time, floor time.Duration) (*Report, error)
```

**The three verbs** (`spec:1069-1072`), in
`cmd/dubplate-server/shadow.go`, matching the shape `newCheckCmd`
(`check.go:16-32`) and `newDoctorCmd` (`doctor.go:23-61`) set, **plus
one parameter**: `SilenceUsage` and `SilenceErrors` both true and
`cmd.SetErr(stderr)`, leaving `main` the sole stderr writer
(`main.go:33-38`).

```go
// newShadowCmd builds the shadow subcommand and its three verbs:
// observe runs one capture, reconcile and match pass and is the timer's
// ExecStart; report prints the launch-gate report; reevaluate runs the
// mechanical re-evaluation ratified ruling B's adjudication-load
// reduction rests on.
//
// matcherCommit is a parameter rather than a package variable read
// inside RunE, on the same injection grounds environ already uses: the
// observe verb refuses an empty or "dev" stamp, and a test that cannot
// vary the stamp cannot prove the refusal. newRootCmd passes the
// package variable (main.go), so production behavior is unchanged and
// no test writes a global.
func newShadowCmd(environ []string, matcherCommit string, stdout, stderr io.Writer) *cobra.Command
```

**The extra parameter is a deliberate extension of the landed shape, and
it is conductor-owed** (fold, B7). Resolution R3 rejects "an overridable
package variable the test writes" by name, and with the stamp read from
a package `var` inside RunE and no stamp on the constructor, writing
that global is the only mechanism `TestObserveRefusesAnUnstampedBuild`'s
table has. Injecting it is the third option R3's ordered list should
have carried: the package variable stays unwritten by any test, R3's
rejection stands, and the table becomes a plain argument.
**Provisional, conductor ruling owed**: `spec:1063-1066` pins the
constructor as `(environ []string, stdout, stderr io.Writer)`, so
extending it by one parameter is a spec deviation on a contested
surface. If the ruling keeps the spec's signature, the refusal moves
into `shadow.Run`'s own configuration validation, where
`Config.MatcherCommit` is already a field and the engine test can vary
it, and the verb's own test asserts only that the error propagates. A
`diff-reviewer` reading the extra parameter as drift should read this
paragraph.

Registration adds one line to `newRootCmd`'s single `AddCommand` block
(`main.go:48-52`), reading
`newShadowCmd(os.Environ(), matcherCommit, os.Stdout, os.Stderr)`, and
`main.go`'s doc comment already reserves the verb at `:11`.

**`main.matcherCommit`** (Resolution R3):

```go
// matcherCommit is the Go matcher revision this build carries, set via
// -ldflags "-X main.matcherCommit=..." at build time. It stays "dev" for
// workstation builds, and shadow observe REFUSES to run on "dev" or on
// an empty value, because an unstamped divergence cannot be
// mechanically re-evaluated and is therefore worse than no observation.
//
// Pass 4b lands the build path that sets it: today no build path does,
// so this value is "dev" everywhere and the refusal fires everywhere,
// which is the correct behavior for a binary whose divergences would be
// unretireable.
var matcherCommit = "dev"
```

**The observe verb's own contract**, in order: load config; refuse when
`cfg.Shadow` is nil, naming `DUBPLATE_SHADOW_ENABLED`; **refuse when
`matcherCommit` is empty or `"dev"`**; resolve the floor and the cap
through the two thresholds; compose the User-Agent as
`"dubplate-server/" + version + " ( " + cfg.AdminEmail + " )"`; open the
store under `filepath.Join(cfg.StateDir, "shadow")`; call `shadow.Run`
with `time.Now()`, which is legal here because `cmd/` is outside the
engine's no-clock walk; write the artifacts through `storeArtifacts`;
print a one-line summary.

**The report's printed rows**, each in `"%-32s %s"` form:

| Row label | Value |
| --- | --- |
| `clock start` | the recorded start, or `not started` |
| `floor` | the configured duration |
| `elapsed` | `now` minus the clock start, and `elapsed / floor` as a percentage |
| `observed albums` | `<applied> applied, <skipped> skipped` |
| `excluded` | one line per reason (`gate-rejected`, `multi-album-directory`, `no-tracks`), with its count |
| `unreconciled` | one line per reason, with its count |
| `report gaps` | the count |
| `deferred` | one line per reason, with its count |
| `divergences` | one line per class, with its count |
| `go-applies-oracle-skipped` | the count, labeled **"owner disposition joined off-box"** |
| `staged pending` | the count. **Must be zero at the launch gate** |
| `resolved` | the count `reevaluate` retired |
| `matcher commits` | one line per commit, with its divergence count |
| `budget` | per-pass `name spent`, then the cumulative total |
| `store size` | the state directory's footprint |

- [ ] **Step 1: Write the failing tests**

```go
func TestBuildReportSumsTheStore(t *testing.T) {
	// A store seeded with: two agreements (one applied, one skipped),
	// one divergence of each class, one deferral of each reason, one
	// excluded row of each reason, one unreconciled row of each of the
	// four reasons, one report gap, three passes with spends 4, 7 and 0,
	// and two observations under different matcher commits.
	//
	// Assert every row, by value. The budget total is 11, not 3 and not
	// the largest, which is what proves the sum is a sum.
}

func TestBuildReportElapsedIsMeasuredAgainstTheParameter(t *testing.T) {
	// Clock start frozen, two different now values, two different
	// elapsed values. No time.Now anywhere in report.go.
}

func TestBuildReportOnAnEmptyStore(t *testing.T) {
	// Every count zero, clock start reported as not started, and no
	// error. A fresh box's first report must be readable rather than an
	// error.
}

func TestObserveRefusesAnUnstampedBuild(t *testing.T) {
	tests := []struct{ name, stamp, want string }{
		{"dev refuses", "dev", "matcher"},
		{"empty refuses", "", "matcher"},
	}
	// Each row constructs newShadowCmd with that stamp as its second
	// argument, runs the observe verb, and asserts a non-nil error
	// naming the stamp, and that the store was NOT created, so a
	// refused pass leaves no half-started clock. The stamp is an
	// argument rather than a package variable the test writes, which
	// Resolution R3 rejects by name (fold, B7).
}

func TestObserveRefusesWhenTheIntegrationIsDisabled(t *testing.T) {
	// A wall with no DUBPLATE_SHADOW_ENABLED. Assert the error names the
	// flag rather than a nil dereference.
}

func TestShadowVerbsAreRegisteredAndPrintToStdout(t *testing.T) {
	// newShadowCmd has exactly three subcommands named observe, report
	// and reevaluate; newRootCmd carries shadow; SilenceUsage and
	// SilenceErrors are both true; and report writes its table to the
	// stdout writer rather than to os.Stdout.
}

func TestReportPrintsThePinnedRowForm(t *testing.T) {
	// Every printed line matches the doctor's own "%-32s %s" shape,
	// asserted by checking each line's first 32 bytes are the label
	// left-padded. The pinned form is what makes one table shape across
	// the binary.
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ ./cmd/dubplate-server/ -count=1 -v`
Expected: FAIL, `BuildReport` and `newShadowCmd` undefined.

- [ ] **Step 3: Implement**

`BuildReport` walks the store once. Map-valued rows are printed in
sorted key order, so a report is reproducible rather than
map-iteration-ordered.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ ./cmd/dubplate-server/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow cmd/dubplate-server && git commit`

**Acceptance criteria:**

- Every report row is asserted by value against a seeded store, and the
  budget row is a sum rather than a maximum or a last-wins.
- `elapsed` is measured against the `now` parameter, and `report.go`
  calls `time.Now` nowhere.
- An empty store reports cleanly.
- `observe` refuses on `dev` and on an empty stamp, naming it, and
  creates no store when it refuses.
- `observe` refuses when the integration is disabled, naming the flag.
- The three verbs are registered, both silence flags are set, and every
  writer is the injected one.
- Every printed line matches `"%-32s %s"`.
- Map-valued rows print in sorted key order.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Let `observe` run on a `dev` stamp →
   `TestObserveRefusesAnUnstampedBuild` fails. This is the refusal that
   keeps an unretireable divergence from ever being written, and 4b's
   deploy-time stamp assertion is its second line of defense.
2. Report the last pass's spend instead of the sum →
   `TestBuildReportSumsTheStore` fails with 0 where 11 is wanted, which
   is the figure the report prints as honest.
3. Read `time.Now()` inside `BuildReport` →
   `TestBuildReportElapsedIsMeasuredAgainstTheParameter` fails, and
   `TestEngineNeverReadsTheClock` fails with it.
4. Print map rows in iteration order → the printed-form test fails
   intermittently, so **assert sorted order explicitly** rather than
   relying on the flake.
5. Print to `os.Stdout` instead of the injected writer →
   `TestShadowVerbsAreRegisteredAndPrintToStdout` fails with an empty
   buffer.

**Deliverable count: 4** (`BuildReport` with its store walk; the three
verbs in the established subcommand shape; the stamp and integration
refusals; the printed table in the pinned row form). Gate green; commit.
**Report actual token spend and the five mutations' outcomes.**

---

## Task 13: `reevaluate`

Model: `sonnet`. Depends on Tasks 8, 9 and 12.

This is the mechanism ratified ruling B's adjudication-load reduction
rests on (`R3:93`, `DA:534-535`). A stamp with no re-evaluator hands the
owner exactly the load the ruling promised to avoid.

**Files:**
- Create: `internal/engine/shadow/reevaluate.go` + `reevaluate_test.go`
- Modify: `cmd/dubplate-server/shadow.go` (the verb's wiring)

**Interfaces:**
- Consumes: `replayTransport` (Task 8), `Store` (Task 5),
  `classify` (Task 9), `musicbrainz.Unmetered`.
- Produces:

```go
// internal/engine/shadow/reevaluate.go

// Reevaluate re-runs every observation stamped with a commit other than
// cfg.MatcherCommit against that observation's own stored responses.
//
// It spends no MusicBrainz budget, reaches no network, and reads no
// clock: the replay client takes musicbrainz.Unmetered()
// (limiter.go:354-359), the transport answers only from disk, and each
// observation is re-run at its OWN stored eval date, so a re-evaluation
// reproduces the original run exactly rather than re-scoring it against
// today.
//
// Four outcomes, and the fourth is why the replay transport refuses a
// miss rather than answering 404:
//
//   - RETIRED: the divergence no longer reproduces. resolved/<obs-id>.json
//     records the retiring commit and both verdicts, and the staged
//     pending row is removed.
//   - RESTAMPED: the divergence still reproduces, so the row is
//     restamped with the current commit and stays staged.
//   - NEWLY DIVERGED: an observation that previously agreed now
//     diverges. That is evidence the launch bar wants, and a
//     retire-only design would silently lose it.
//   - UNEVALUABLE: the replay asked for a request the recording does not
//     carry. The row stays stamped and staged. A newer matcher revision
//     can take a different branch through tag.Album for the same input,
//     exactly as a different Compat can, and a silently retired
//     divergence is the one outcome worse than an un-retired one.
func Reevaluate(ctx context.Context, cfg Config, s *Store, now time.Time) (*ReevaluateResult, error)

// A ReevaluateResult names what moved, by observation id, so the verb
// prints a record rather than a count.
type ReevaluateResult struct {
	Retired       []string
	Restamped     []string
	NewlyDiverged []string
	Unevaluable   []string
	Skipped       []string // already at the current commit
}
```

**What `resolved/<obs-id>.json` carries**: the observation id, the
retiring commit, the original commit, both classes (the original and the
re-run's `agreement`), both verdict pairs, and the date, rendered from
the `now` parameter. It is the engine's own type and needs no harness
import.

- [ ] **Step 1: Write the failing tests**

```go
func TestReevaluateRetiresADivergenceThatNoLongerReproduces(t *testing.T) {
	// A store holding one go-skips-oracle-applied observation stamped
	// "old", whose recorded responses now score an apply on the oracle's
	// own release under the current tunables. Assert: the id is in
	// Retired, resolved/<id>.json exists naming both commits, and the
	// staged pending row is gone.
}

func TestReevaluateRestampsADivergenceThatStillReproduces(t *testing.T) {
	// Assert the id is in Restamped, the observation's matcher_commit is
	// the current one, and the pending row is still there.
}

func TestReevaluateRecordsANewDivergenceOnAnObservationThatAgreed(t *testing.T) {
	// An agreement stamped "old" whose replay now diverges. Assert the id
	// is in NewlyDiverged and that a pending row now exists where none
	// did. A retire-only design loses exactly this.
}

func TestReevaluateReportsUnevaluableRatherThanRetiring(t *testing.T) {
	// The fixture's class and candidate count are named rather than
	// left to the implementer, because mutation 1 only fires for one
	// shape (fold, m16): a go-applies-oracle-skipped observation with
	// exactly TWO admitted candidates, whose stored responses cover one
	// of the two lookups the replay issues, and whose remaining
	// candidate does not score an apply. Dropping the missing candidate
	// silently then leaves no apply, the re-run agrees with the oracle
	// skip, and the row RETIRES. On any other shape a dropped candidate
	// can leave the divergence in place and produce Restamped, and the
	// pass's most consequential mutation would not fire.
	//
	// Assert: the id is in Unevaluable, the observation is still
	// stamped "old", the pending row survives, and NOTHING is written
	// under resolved/. This is the test that proves the 501 refusal is
	// load-bearing.
}

func TestReevaluateSpendsNoBudgetAndReachesNoNetwork(t *testing.T) {
	// cfg.Transport is a round tripper that fails every call, so a
	// network read is a test failure rather than a slow test. cfg's
	// BudgetLimit is 1; assert the pass budget is untouched, which the
	// Unmetered replay client guarantees.
}

func TestReevaluateSkipsObservationsAtTheCurrentCommit(t *testing.T) {
	// One observation already at the current commit. Assert it is in
	// Skipped and that its stored record is byte-identical afterwards.
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/engine/shadow/ -run Reevaluate -count=1 -v`
Expected: FAIL, `Reevaluate` undefined.

- [ ] **Step 3: Implement**

Reuse `replayTransport` and the re-score's own five steps rather than
declaring a second replay path. The one difference is that `Reevaluate`
re-runs the **whole flow** through `tag.Album` (a newer revision may
take a different branch, which is the point), while `rescoreDefault`
re-scores a recorded candidate set. Both drive the same transport and
both refuse a miss the same way.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow cmd/dubplate-server && git commit`

**Acceptance criteria:**

- All four outcomes have their own firing fixture, plus the skip case.
- A retired divergence writes `resolved/` and removes the pending row;
  nothing else does.
- An un-re-evaluable observation stays stamped and staged and writes no
  `resolved/` row.
- No budget is spent, no network is reached, and no clock is read.
- Each observation re-runs at its own stored `eval_date`, asserted by a
  fixture whose verdict depends on the date.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Answer a replay miss with 404 instead of 501 →
   `TestReevaluateReportsUnevaluableRatherThanRetiring` fails: the
   candidate is dropped silently inside `lookupAndAdd`
   (`tag.go:166-170`), the re-run agrees with the oracle skip, and the
   row retires. This is the single most consequential mutation in the
   pass, because a retired divergence leaves no trace to notice, and it
   fires **only** against the two-candidate `go-applies-oracle-skipped`
   fixture the test names (fold, m16).
2. Drop the newly-diverged arm and retire only →
   `TestReevaluateRecordsANewDivergenceOnAnObservationThatAgreed` fails,
   and the launch bar loses evidence it wanted.
3. Re-run at `now` instead of the stored `eval_date` → the date-
   dependent fixture fails, and every re-evaluation would silently
   re-score against a different year penalty.
4. Hand the replay client the metered budget → the budget test fails on
   `Spent()`.
5. Restamp an un-re-evaluable observation → its assertion on the
   surviving `old` stamp fails, and a later pass would think it had
   been re-evaluated.

**Deliverable count: 3** (the four re-evaluation outcomes; the
`resolved/` record; the verb's wiring). Gate green; commit. **Report
actual token spend and the five mutations' outcomes.**

---

## Task 14: The exit gate, one end-to-end `observe` run against a hand-authored inbox

Model: `sonnet`; **the named runner-up for an `opus` upshift** if Task 9's
review returns `fix`. Depends on Tasks 1 through 13.

This is the pass's exit gate, quoted from the spec's pass-split table
(`spec:1616`): "`bash scripts/check.sh` green, plus one end-to-end
`observe` run against a hand-authored inbox and a hand-authored run
report in the pass's own `testdata`, producing one agreement, one
divergence of each class, one excluded multi-album directory, and one
deferred observation of each error branch."

**A shadow that silently observes nothing is this rung's equivalent of a
conformance runner that grades nothing, and it would cost thirty days
before anyone noticed** (`spec:1816-1822`). The instrument is proven
able to fire before it is deployed, and that is what this task is for.

**Files:**
- Create: `internal/engine/shadow/observe_test.go`
- Create: `internal/engine/shadow/testdata/` (the WS/2 response bodies
  the server serves; every album's FLAC bytes are built in the test)

**The fixture set.** Thirteen inbox directories under one uploader,
`ada`, plus the artifacts that resolve them. **Names carry a numeric
prefix so the walk order is stated rather than alphabetical accident**
(Resolution R18), and the two budget-capped albums sort last so the cap
bites where the fixture intends. **That works only because the pass loop
matches in capture order** (Task 9, fold B3): an observation id is a
digest prefix, so a match step iterating in id order would starve
whichever album's digest happened to sort last and make the budget
assertion flaky. **And `buildInbox` pins every written file's
modification time** with `os.Chtimes(path, fixedModTime, fixedModTime)`,
because the digest ingests modification time and an id that moves with
the wall clock is not reproducible across two runs.

| # | Basename | Inbox tags | Server | Run report row | Backstop | Expected |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | `01-agree-apply` | unanimous `MUSICBRAINZ_ALBUMID=R1`; titles and lengths close to R1 | `release-R1.json`, scoring strong | `imported`, `library_paths` naming a library FLAC tagged `MUSICBRAINZ_ALBUMID=R1` with matching titles and lengths | none | `agreement`. **One request**, the ID-consensus early return |
| 2 | `02-agree-skip` | no album id | a search body with an empty `releases` array | `review` | a tag-log `skip ` line | `agreement`, Go returning `tag.ErrNoCandidates` |
| 3 | `03-go-skips` | no album id | an empty search | `imported`, library FLAC tagged `R2` | none | `go-skips-oracle-applied` |
| 4 | `04-go-applies` | unanimous `G1` | `release-G1.json`, strong | `review` | a review directory named `04-go-applies-142530`, the `-HHMMSS` collision form | `go-applies-oracle-skipped`; the pending row's `release_mbid` is `Candidates[0].Info.AlbumID` |
| 5 | `05-go-different` | unanimous `G2` | `release-G2.json`, strong | `imported`, library FLAC tagged `R3` | none | `go-applies-different-release`. `R3` needs no response body, only the library tag |
| 6 | `06-multi-album` | four tracks over two `(albumartist, album)` pairs | none | none | none | **excluded, multi-album directory**; never reaches the match step |
| 7 | `07-gate-rejected` | one track | none | `rejected` | none | **excluded, gate-rejected** |
| 8 | `08-no-library-paths` | one track | none | `imported` with an empty `library_paths` and the importer's own `Error` text | none | `unreconciled`, `imported-with-no-release-id` (Resolution R22; the second half of that reason, a library file carrying no `MUSICBRAINZ_ALBUMID`, is proven by Task 7's own unit fixture rather than by a fourteenth directory) |
| 9 | `09-partial-skip` | one track | none | `review` with `Error: "left in staging after import"` | none | `unreconciled`, `partial-skip` |
| 10 | `10-ambiguous` | one track | none | two `imported` rows whose library files carry identical titles and lengths | none | `unreconciled`, `ambiguous-basename` |
| 11 | `11-swept` | one track, captured on the seeded earlier pass | none | absent; the ledger names a report file that no longer exists and the oldest present report starts after the newest consumed one ends | none | `unreconciled`, `report-swept`, and one report gap |
| 12 | `12-defer-unavailable` | unanimous `U1` | 503 with `Retry-After: 0` for `/ws/2/release/U1` | `imported`, library FLAC tagged `U1` | none | **deferred**, `musicbrainz-unavailable`, after seven attempts |
| 13 | `13-defer-budget` | unanimous `B1` | `release-B1.json`, never served | `imported`, library FLAC tagged `B1` | none | **deferred**, `budget-exhausted` |

**The budget is sized so the cap bites exactly at 13.** Albums 1, 3, 4
and 5 spend one or two tokens each on the early return or the empty
search; album 12 spends seven, one per attempt including retries
(`client.go:346`). The fixture sets `BudgetLimit` to the sum of every
spend through album 12, so album 13's first request is the one that
reserves past the cap. **The test asserts that arithmetic rather than
guessing it**: it reads `budget.Spent()` off the pass record and
compares it against the cap.

**`ErrLimiterClosed` is a second, focused test**, not a row in this run
(Resolution R7). `Budget.Wait` reserves before it waits
(`limiter.go:372-381`), so once the cap is spent the limiter is never
consulted, and `Limiter.Close` fails every future waiter
(`:147-160`), so once closed the cap is never reached. The two branches
are mutually exclusive within one pass's budget. The focused test runs
one pass against a pre-closed `musicbrainz.NewLimiter` supplied through
`Config.Limiter` and asserts every observation defers with
`DeferLimiterClosed`.

- [ ] **Step 1: Write the failing end-to-end test**

```go
func TestObservePassEndToEnd(t *testing.T) {
	tree := buildInbox(t)            // the thirteen directories above
	srv := releaseServer(t, bodies, statuses)
	defer srv.Close()

	cfg := Config{
		StateDir:       tree.stateDir,
		ObserveDir:     tree.inbox,
		OracleStateDir: tree.oracleState,
		ReviewDir:      tree.review,
		UserAgent:      "dubplate-test/0 ( test )",
		MatcherCommit:  "abc1234",
		BudgetLimit:    tree.budgetCap,
		Interval:       time.Millisecond, // a real limiter, off the wall clock
		BaseURL:        srv.URL,
	}

	res, err := Run(context.Background(), cfg, passNow)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Assert on parsed structure, never on a golden: goldens catch THAT
	// output changed, not WHAT should have changed
	// (go-conventions SKILL.md:355-360).
	byID := indexByBasename(t, res.Observations)

	wantClass := map[string]string{
		"01-agree-apply":  ClassAgreement,
		"02-agree-skip":   ClassAgreement,
		"03-go-skips":     ClassGoSkipsOracleApplied,
		"04-go-applies":   ClassGoAppliesOracleSkipped,
		"05-go-different": ClassGoAppliesDifferentRelease,
	}
	for basename, want := range wantClass {
		if got := byID[basename].Class; got != want {
			t.Errorf("%s: class = %q, want %q", basename, got, want)
		}
	}

	wantDeferred := map[string]string{
		"12-defer-unavailable": DeferUnavailable,
		"13-defer-budget":      DeferBudgetExhausted,
	}
	for basename, want := range wantDeferred {
		obs := byID[basename]
		if obs.Deferred != want {
			t.Errorf("%s: deferred = %q, want %q", basename, obs.Deferred, want)
		}
		if obs.Class != "" {
			t.Errorf("%s: class = %q on a deferred observation; a deferral is never a verdict", basename, obs.Class)
		}
	}

	wantUnreconciled := map[string]string{
		"08-no-library-paths": UnreconciledNoReleaseID,
		"09-partial-skip":     UnreconciledPartialSkip,
		"10-ambiguous":        UnreconciledAmbiguousBasename,
		"11-swept":            UnreconciledReportSwept,
	}
	// ... each asserted, with Class empty on every one.

	if got := res.Pass.Excluded[ExcludedMultiAlbum]; got != 1 {
		t.Errorf("excluded multi-album = %d, want 1", got)
	}
	if got := res.Pass.Excluded[ExcludedGateRejected]; got != 1 {
		t.Errorf("excluded gate-rejected = %d, want 1", got)
	}
	if res.Pass.ReportGaps != 1 {
		t.Errorf("report gaps = %d, want 1", res.Pass.ReportGaps)
	}
	if res.Pass.Budget.Spent != tree.budgetCap {
		t.Errorf("budget spent = %d, want the cap %d", res.Pass.Budget.Spent, tree.budgetCap)
	}

	// Three pending rows, one per divergence class, and none for an
	// agreement or a deferral.
	if len(res.Pending) != 3 {
		t.Fatalf("pending rows = %d, want 3", len(res.Pending))
	}
	for _, row := range res.Pending {
		if row.MatcherCommit != "abc1234" {
			t.Errorf("%s: matcher_commit = %q, want the run's stamp", row.ID, row.MatcherCommit)
		}
		if row.ReleaseMBID == "" {
			t.Errorf("%s: release_mbid empty; every divergence has exactly one release under dispute", row.ID)
		}
	}

	// The store is what an operator and rung 15 read, so assert it too.
	s, err := OpenStore(filepath.Join(tree.stateDir, "shadow"))
	if err != nil {
		t.Fatal(err)
	}
	stored, err := s.Observations()
	if err != nil {
		t.Fatal(err)
	}
	// Thirteen directories minus the multi-album one, which is excluded
	// at capture and never becomes an observation. The gate-rejected
	// one IS stored, and carries Observation.Excluded.
	if len(stored) != 12 {
		t.Errorf("stored observations = %d, want 12", len(stored))
	}
	if start, ok, err := s.ClockStart(); err != nil || !ok || !start.Equal(passNow) {
		t.Errorf("clock start = %v, %v, %v; want the pass clock", start, ok, err)
	}
}

func TestObservePassDefersEverythingOnAClosedLimiter(t *testing.T) {
	// One pass over the same inbox with Config.Limiter set to a
	// musicbrainz.NewLimiter that has already been Closed. Assert every
	// reconciled observation defers with DeferLimiterClosed, and that
	// none carries a class (Resolution R7).
}

func TestObservePassIsIdempotent(t *testing.T) {
	// Run twice over the same tree at the same clock. Assert the second
	// pass captures nothing, matches nothing, spends no budget, and
	// leaves every stored observation byte-identical, which is contract
	// S7's idempotence measured end to end rather than per function.
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/engine/shadow/ -run Observe -count=1 -v`
Expected: FAIL until every earlier task has landed; if it fails for a
reason other than a missing symbol, that failure is the finding this
task exists to surface.

- [ ] **Step 3: Author the fixture tree**

`buildInbox` builds every FLAC with `buildFLAC` (Task 4), pins every
written file's modification time to `fixedModTime`, writes the library
files the `imported` rows name, writes the run reports, the tag log and
the review tree, and seeds the consumed-report ledger for the gap row.
**Every string is invented**; no estate path, name or catalog value
appears (Global Constraints, universality). The WS/2 response bodies are
the only committed `testdata`, and they are minimal hand-authored
documents rather than captured ones.

**Author `releaseBody` by adapting a committed body that is already
proven to score strong, rather than by iterating against the penalty
surface** (fold, M-f). Rows 1, 4 and 5 each need a release document
scoring `RecStrong`, and row 1 additionally needs the ID-consensus early
return, which fires only at `rec == matcher.RecStrong`
(`internal/engine/tag/tag.go:106`) and therefore needs an album distance
at or below `StrongRecThresh = 0.04`
(`internal/engine/matcher/tunables.go:268`) across a twenty-weight
surface (`:275-296`). Whether a "minimal" hand-authored body clears that
bar is discovered by iterating, not reasoned to, and three of the five
verdict outcomes rest on it. The tree already holds the answer:
`internal/engine/tag/testdata/early-return/release-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.json`
is a committed body that `internal/engine/tag/tag_test.go:141-158`
asserts scores `matcher.RecStrong` through the identical `tag.Album`
path, with `medium-id/` beside it for a body that must not. It is engine
testdata rather than corpus, so adapting it breaches no rule: copy it
into this package's own `testdata` and change only ids, titles and
lengths. **A row failing to reach `RecStrong` is a fixture defect, not a
matcher finding**, and the task reports it as one rather than chasing it
into the engine.

- [ ] **Step 4: Run the test and report its wall time**

Run: `go test ./internal/engine/shadow/ -count=1 -v`
Expected: PASS. **Report the measured wall time of the package's test
run**, and the `internal/engine/shadow` figure separately, so 4b's plan
and rung 15 size against a measurement rather than an assumption.

- [ ] **Step 5: Run the gate and commit**

Run: `bash scripts/check.sh`
Then: `git add internal/engine/shadow && git commit`

**Acceptance criteria:**

- One agreement of each shape, one divergence of each of the three
  classes, one excluded multi-album directory, one excluded
  gate-rejected directory, one `unreconciled` row of each of the four
  reasons, one report gap, and one deferral of `ErrUnavailable` and of
  `ErrBudgetExhausted`, all in one `observe` run.
- `ErrLimiterClosed` is proven by the second focused test, with the
  mutual-exclusion reason stated in the test's own comment.
- Exactly three pending rows, each stamped and each carrying a
  `release_mbid`.
- The store holds twelve observations, `started-at` holds the pass
  clock, and every stored record carries exactly one of `Class`,
  `Deferred`, `Unreconciled` and `Excluded`. **`07-gate-rejected` is the
  record that needs the fourth term**: the capture stores it as an
  ordinary single-album directory, and only the run report's `rejected`
  row later excludes it, so it carries `Excluded` and none of the other
  three (fold, B5).
- The budget spend equals the cap, asserted rather than assumed, **and
  the album the cap starves is `13-defer-budget` because the match step
  runs in capture order**, not because a digest happened to sort it
  last.
- The 503 row completes without spending real backoff time, because the
  server sends `Retry-After: 0`; **the whole package test run stays
  under thirty seconds**, and the task reports the measured figure.
- A second pass over the same tree captures nothing and spends nothing.
- No golden file anywhere; every assertion is on parsed structure.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Make `capture` skip the grouping check → the multi-album directory
   becomes a fourteenth observation, the excluded count reads 0, and the
   stored count reads 13.
2. Make `reconcile` treat a `rejected` row as a skip → the gate-rejected
   count reads 0 and an extra agreement appears, inflating the observed
   count with material that never reached beets.
3. Set the budget cap to zero → **every album succeeds**, because a
   limit of zero or less means no cap (`limiter.go:130-133`), and the
   `13-defer-budget` row fails. This is the mutation that proves the cap
   is a cap rather than a decoration.
4. Drop the `Retry-After: 0` header from the test server → the run
   spends 31 real seconds on album 12 (`maxAttempts` 7 with waits
   0+1+2+4+8+16, `client.go:42`, `:418-423`) and the wall-time
   assertion fires.
5. Reverse the fixture's numeric prefixes so `13-defer-budget` sorts
   first → the budget row fails, and the failure names the capture order
   the fixture depends on, which is Resolution R18 made visible. **This
   mutation is only reachable because the match step follows capture
   order**; under revision 1's lexical id order it could not fire at
   all, since a numeric prefix does not move a SHA-256 prefix (fold,
   B3).
6. Drop `buildInbox`'s `os.Chtimes` calls →
   `TestObservePassIsIdempotent` and
   `TestRunIsDeterministicUnderAFrozenClock` both fail, because a digest
   that ingests the wall clock is a different digest on every run.

**Deliverable count: 3** (the thirteen-directory fixture tree with its
reports, tag log, review tree and ledger; the end-to-end assertions on
parsed structure; the closed-limiter and idempotence tests). Gate green;
commit. **Report actual token spend, the six mutations' outcomes, the
measured wall time of the package and of `go test ./...`, and the
budget arithmetic the fixture rests on.**

---

## Task 15: The pass close, and the 4a to 4b handoff

Model: `sonnet`. Depends on Tasks 0 through 14.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit)
- Draft, handed to the conductor in the task return: three rung 4 spec
  corrections; one HISTORY correction; two filed notes; one owner ask

**Behavior contract:**

- **The HISTORY entry is drafted in the house shape**: what landed, what
  the gate caught, and what a later pass would be wrong to rediscover.
- **The shadow contract coverage table, as a table a reviewer can
  check.** Per contract S1 through S23: its Go site, what proves it, and,
  for a rule no end-to-end run can reach, which hand-authored fixture
  proves it and what mutation proved that fixture fires.
- **The six provisional ruling applications are listed with their
  markers**, so the fold that follows an owner ruling is a search rather
  than a re-reading. Each row names the file and the symbol the marker
  sits on.
- **The declared coverage gaps are recorded as gaps, not as passing
  checks.** 4a's list, on one criterion: a surface **4a itself builds**
  that no test in this pass can reach, or a real input shape the design
  deliberately does not observe.

  1. **An archive upload is invisible to the capture** (Resolution R19):
     `stageArchive` extracts a `.zip` and removes it from the inbox
     (`importer.go:861-888`), so the album that later appears in a run
     report has a basename the shadow never saw.
  2. **The re-scored `DefaultCompat` verdict is a hypothesis, not a
     measurement**, wherever the real flow would have widened the
     candidate set (ruling 3's own limit, `spec:230-237`).
  3. **`Budget.wait` cannot be injected per album**, so the deferral
     branches are proven in two runs rather than one (Resolution R7).
  4. **The report's disposition rows cannot be computed on the box**
     (Resolution R21): owner dispositions live in the corpus on the
     workstation, and the launch gate does that join.
  5. **`atomicfile`'s cross-filesystem property is unproven on a
     single-filesystem workstation** (Task 1's mutation 1), and is held
     by the doc comment plus 4b's `ReadWritePaths` argument.
  6. **No test exercises a real library tree**: `readFLAC` on the
     applied file is driven through an injected reader in the reconcile
     tests, and through built FLACs in the end-to-end run. A real
     beets-written library file is first read at 4b's deploy.
  7. **The shadow's own timing against a live import is untested**: the
     quiescence rule is a constraint on 4b's timer interval, and 4a
     proves no part of it.
  8. **"`internal/engine/shadow` imports `internal/config` never" has no
     mechanical guard** (fold, m14). Resolution R16 states the rule and
     rests it on written prose, while the repo's own precedent for a
     rule this load-bearing is a walk: `checkNoHarnessImports` matches a
     package-level `forbiddenImportPrefix` over a rooted tree
     (`internal/harness/conform/arch_test.go:30-64`), and a sibling
     walker with a second prefix would be a short file. **The gap is
     declared rather than closed**, because closing it widens this
     pass's `internal/harness/conform` territory past the one
     drift-guard assertion and one doc-comment clause Global Constraints
     names, and the `diff-reviewer` reads every `internal/engine/shadow`
     diff meanwhile. **Provisional, conductor ruling owed.**
  9. **The cross-filesystem half of `atomicfile`'s rename property is
     untested**, distinct from gap 5's temp-directory half: Task 1's
     replacement test proves the temp file lives in the target's own
     directory, and nothing proves what `os.Rename` does across an EXDEV
     boundary.

  **The count is nine, stated and joined to rung 3's running total of
  twenty-one**, so rung 15 does not re-count.
- **The evidence, not a bare "green".** The end-to-end run's thirteen
  outcomes with their counts; the budget arithmetic; the measured wall
  times for `internal/engine/shadow`, `internal/harness`,
  `go test ./...` and `bash scripts/check.sh`; and confirmation that
  tier 0 still replays 72/72, translation still grades 50/50, the
  request tier still runs 215/303/48, tier 1 still grades 215 cases and
  tier 2 still runs both modes, all unchanged by this pass.
- **The mutation table**, every named mutation across Tasks 1 through 14
  with its measured outcome, including the four deliberately unfired
  ones (Task 1's mutations 1 and 3, Task 5's mutation 2, Task 8's
  mutation 2), each with the guard that holds the property instead.
- **The 4a to 4b handoff, stated as what 4b's plan needs from 4a:**
  1. **The store's root and its footprint.** `<state dir>/shadow/` with
     `passes/`, `observations/<obs-id>/`, `pending/`, `resolved/`,
     `started-at` and `consumed-reports.json`, and **nothing written
     anywhere else**, proven by Task 5's walk test. That is 4b's
     `ReadWritePaths=` argument, static rather than assumed.
  2. **The unit's `ExecStart` and its refusal.** `dubplate-server shadow
     observe` errors rather than observing when `main.matcherCommit` is
     empty or `"dev"`, and today no build path sets it. **4b's deploy
     edit 6 is what makes the refusal observable**: without the two
     `-ldflags` stamps and the smoke-test assertion, an unstamped binary
     deploys clean, every timer pass refuses, and the `OnFailure=` chain
     fires into a slug whose URL is unset and therefore no-ops.
  3. **The six config rows and how they render.** The three paths are
     required-when-enabled and abort `Load` when the flag is set without
     them; the flag and the two thresholds are accepted and never
     required; and **an explicit `DUBPLATE_SHADOW_ENABLED=false` loads
     clean with a nil `Shadow`**, which is the adopter's own way to turn
     the integration off (fold, M-d). **The greenfield template carries six `DUBPLATE_SHADOW_*`
     rows and not four**, which is a spec correction 4a drafts:
     `spec:1267` says "the four `DUBPLATE_SHADOW_*` rows" and then
     enumerates five values, while the key table lists six including the
     enable flag.
  4. **The wire pace and the cap.** `2 * musicbrainz.DefaultInterval`
     and sixty tokens, both provisional on ruling 5, and both one
     literal away from a different value. Sixty tokens at two seconds a
     token bounds a pass at two minutes, comfortably inside 4b's
     five-minute timer.
  5. **The operator rule the unit's comment must carry**: a hand-run
     shadow verb goes through `systemctl start dubplate-shadow.service`
     or `systemd-run` with the same properties, never the bare binary as
     root, because the read-only guarantee is the unit's and not the
     binary's (`spec:578-589`).
  6. **The clock's definition** (Resolution R11): `started-at` is
     written on the first pass that completes all three steps without a
     fatal error, including a pass that observes nothing, so the clock
     starts at the deploy rather than at the first upload. **4b's close
     records that date as a first-class figure**, because launch date
     equals it plus the floor.
- **Forward carries, each stated so 4b and rung 15 read them rather than
  derive them:**
  1. **The rung 15 re-run criterion**, quoted verbatim from the Spec
     section above, carried forward as every rung 3 pass close carried
     it, now joined by the shadow's own launch-bar clause.
  2. **Rung 3's twenty-one declared corpus-coverage gaps** are read as a
     running total, never re-counted, and 4a's own gaps are added to it
     with their own count.
  3. **The four-knob wall registration** (`StrongRecThresh`,
     `MediumRecThresh`, `RecGapThresh`, `SearchLimit`) stays the named
     follow-on for the first pass after rung 1b, and **rung 4's own
     `KindFloat` need joins it**: `Threshold.Kind` now has two values
     and the three float rec thresholds want a third, which
     `docs/HISTORY.md:678-685` already records as rung 3's follow-on.
     `AdminEmail`'s dedicated contact row sits beside both.
  4. **Rung 1b absorbs `integrationClasses` as data**, not as a shape to
     invent, and the table's two-set split (`Required` and `Optional`)
     is the semantics it must preserve.
  5. **Rung 8 absorbs the FLAC reader** into `internal/engine/tagio` and
     repoints the shadow. The reader is unexported, shadow-local and
     named provisional in its own doc comment; the nineteen-field
     mapping with its per-row citations and its precedence rules is in
     Task 4 of this plan and travels into HISTORY, so rung 8 does not
     re-derive it. **The deployed binary carries landed-but-unwired
     packages from other lanes, and none of them is reachable from the
     `shadow` subcommand's own call graph** (`spec:1576-1582`);
     repointing at `tagio` is rung 8's task, never a silent adoption.
  6. **Rung 9 re-verifies the limiter's caller-priority policy against a
     measurement**, because 4a is the first `ClassBackground` caller and
     4b puts it beside a live import. 4a's own figure is the end-to-end
     run's budget spend; the live one arrives after 4b's deploy.
  7. **Rung 6's store supersedes none of this** (`spec:732-735`), and
     `internal/lock` stays unadopted: the timer path is a systemd
     singleton, `reevaluate` writes only under `resolved/`, `report`
     opens nothing for writing, and `observe` only ever creates.
  8. **The response store is a consequence of ratified ruling B rather
     than a preference**, and the retention knob is named rather than
     built: the trigger is the report's own store-size row against the
     root filesystem's free space, which 4b's sitting records as a
     baseline.
- **The drafted corrections, each one line for the conductor's batch:**
  - `spec:1608-1609`, "(ruling 7)" reads ruling 6, the deploy sitting
    (Resolution R10).
  - `spec:299-301`, ruling 5's pre-fold sentence "at one request per
    second the run's wire time stays inside sixty seconds", which the M6
    fold twenty-five lines later supersedes and which the conductor's
    pace ruling makes stale.
  - `spec:1267`, "the four `DUBPLATE_SHADOW_*` rows", which enumerates
    five values against a key table of six.
  - `spec:846-848`, the two new `AdjudicationEntry` rules' fixture form:
    the package's landed precedent for a per-row rule is an inline Go
    literal variant, and `testdata/` holds cross-ledger fixtures only
    (Resolution R14).
  - `spec:1616`, the exit gate's "one deferred observation of each error
    branch", which one pass cannot produce for mechanical reasons
    (Resolution R7).
  - `spec:845-846`, "so the 21 committed corpus rows validate
    byte-for-byte unchanged". Byte identity was never available: the
    committed ledger orders `corrected_verdict` after `date` while
    `AdjudicationEntry` declares it before `Reason`
    (`adjudication.go:46-67`), and the file carries a literal `<` and
    three `>` that Go's encoder escapes. The guarantee that holds is a
    round trip with no key gained or lost, which Task 10 asserts (fold,
    B2).
  - `spec:441`, the capture's grouping key. Pinned beets groups on
    `(item.albumartist or item.artist, item.album)`
    (`beets/importer/stages.py:102-103`), and the spec's
    `albumartist`-only reading observes as one album exactly the flat
    contributor directory the rule exists to exclude (fold, B4). **This
    one carries a conductor ruling rather than a note**, because 4a
    implements the fallback provisionally.
  - `spec:1063-1066`, the subcommand constructor's shape. `newShadowCmd`
    takes the matcher stamp as a fourth parameter, because Resolution
    R3 rejects the package variable a stamp table would otherwise have
    to write (fold, B7). **Conductor ruling rather than a note.**
  - `docs/HISTORY.md:287-293`, the 3d handoff's item 3, whose recorded
    wire pace of `DefaultInterval` the conductor's ruling amends to
    `2 * DefaultInterval`.
- **Two notes filed, neither this pass's work:** renaming
  `case.json`'s `Profile` the day the corpus grows a `quiet-import`
  profile, for lane E1 (Resolution R8); and a `KindFloat` threshold
  variant, for the first rung 3 follow-on after rung 1b (carry 3).
- **One owner ask drafted, batched rather than raised**: the six open
  rulings' answers, presented as the six provisional applications with
  their fold sentences, so a sitting reads what changes rather than
  re-deriving each question. **3d's own inherited ask rides the same
  stream and is not rung 4's to raise**: re-dispositioning
  `bug-assignment-rejection-gap` from `oracle-wrong` to
  `no-correct-answer` (`docs/HISTORY.md:388-393`).
- **STATUS stays present tense and under 60 lines**; anything historical
  moves to HISTORY. The next action points at pass 4b, and the lane
  registry entry is updated.
- **Both budgets are scored**: token spend against the re-derived
  ceiling, and attended interactions as two counts, planning misses and
  execution sittings, against a target of zero for both.

**Deliverable count: 4** (the HISTORY draft in the house shape with its
eight forward carries; the shadow contract coverage table and the seven
declared gaps joined to rung 3's running total; the evidence record with
the end-to-end run's outcomes, the mutation table and the measured wall
times; the STATUS rewrite plus the 4a to 4b handoff, the six drafted
corrections, the two filed notes and the one owner ask). Gate green;
commit.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain per task,
**strictly sequential**. Nothing here runs in parallel, and if a second
code slot opens at all it belongs to lane A or lane T in its own
worktree, never to a second executor in this tree. `docs/STATUS.md`'s
lane registry records lane-2 as BLOCKED until Phase 0b and 0c land, so
the expected state through this pass is one code lane. Every task runs
the repo's Sonnet implementer (`cairn-implementer`, which commits its
own task's specific files once `bash scripts/check.sh` is green), then
`diff-reviewer`, then the gate, with one re-dispatch on `fix` and a
second `fix` as the conductor's decision. The conductor dispatches
`code-simplifier` over each accepted task's changed Go after review and
before the pass close, and owns every ledger edit, the lane registry,
the six drafted corrections and the batch commit. **Runaway guard armed
at launch**, per the standing requirement. **Subagents never dispatch
agents.**

**Models.** `sonnet` by default. **Two named upshifts to `opus`:**

- **Task 4**, the FLAC reader and its mapping. It is the pass's one
  binary-format surface, its nineteen-field mapping has seven
  multi-style fields whose read order is load-bearing and invisible to a
  single-key test, and two of the nineteen have no Vorbis key at all. A
  wrong `Length` unit moves every verdict downstream, and nothing in the
  pass would catch it except that field's own test.
- **Task 9**, the re-score and the classification. It is where the
  artifact's honesty is decided: a default arm that falls into a verdict
  instead of deferring fabricates divergences in the one record whose
  whole value is that it is honest, and it consumes exactly the owner
  adjudication ruling B's stamp exists to save. It also carries the
  re-score, whose correct mechanics are fold B3's own correction of the
  spec's first revision.

**Two named runners-up**, each upshifted on a condition rather than by
policy. **Task 7**, the reconcile step, if Task 6's review returns
`fix`: the basename join is the pass's trickiest non-parsing logic and
its failure mode is a silently wrong oracle signal. **Task 14**, the
exit gate, if Task 9's review returns `fix`: the instrument's own
fires-proof sits on the classification Task 9 lands.

**Review.** Rung 4 is not one of the three full adversarial-cycle rungs
(5, 12 and 16), so review scales to the in-chain `diff-reviewer` per
task plus the pass-end gate (`spec:1800-1803`). **The spec's two named
exceptions are both 4b's**: the fresh-context read of the unit files and
the deploy path, and the deploy sitting. 4a takes neither, and the
spec's second exception, "one mutation proof per checker"
(`spec:1816-1822`), is discharged by the named mutations in every task
above rather than by a separate review.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- **Task 1 → Task 5.** The store writes through `atomicfile`.
- **Task 4 → Task 5.** `Capture.Items` is the reader's output.
- **Tasks 4, 5 → Task 6.** The capture reads files and stores fixtures.
- **Tasks 4, 5, 6 → Task 7.** The reconcile step reads the library file
  through the reader and joins against stored captures.
- **Tasks 5, 6, 7 → Task 8.** The match step consumes a reconciled
  capture.
- **Task 8 → Task 9.** The re-score replays the recorded responses and
  the pass loop composes everything.
- **Tasks 9, 10 → Task 11.** The marshalling needs both the engine
  shapes and the two schema fields.
- **Tasks 3, 9, 11 → Task 12.** The subcommand needs the config rows,
  the runner and the writers.
- **Tasks 8, 9, 12 → Task 13.** `reevaluate` reuses the replay transport
  and the classification, and hangs off the third verb.
- **Tasks 1 to 13 → Task 14.**
- **Tasks 0 to 14 → Task 15.**
- **Task 2 depends on Task 0 alone**, and **Task 10 depends on Task 0
  alone**; both are independent of every other task. **They are still
  run sequentially**, because one tree runs one executor and the module
  gate runs over `./...` regardless. The conductor may reorder either
  earlier if a dependency stalls.

**Where this pass splits, if it splits.** Fifteen tasks plus a pre-task
is well past the roughly-eight bar, so the split point is named rather
than the scope packed. **The split is at the Task 9 / Task 10
boundary**: Tasks 0 to 9 are the engine half (`internal/atomicfile`,
`internal/config`, `internal/engine/{matcher,shadow}`) and write no
`internal/harness` and no `cmd` surface beyond Task 3's registration;
Tasks 10 to 15 are the harness schema, the command surface, the exit
gate and the close.

**The recommendation is to keep one pass.** Three reasons, in order of
weight. **The engine half alone ships an unreachable package**: no verb
reaches it, no artifact is written in the shape rung 15 reads, and no
run has ever produced a divergence, which is exactly the "instrument
nobody points at anything" shape the spec's own review scaling exists to
prevent. **4b is blocked on the whole runner rather than on its engine
half**: 4b's unit `ExecStart`s `dubplate-server shadow observe`, its
smoke test asserts a stamp the subcommand refuses on, and its
`ReadWritePaths` argument is Task 5's walk test, so a split would put a
full close and a full staleness pre-task between 4b and everything it
needs. **And the two halves are cheap in opposite places**: the engine
half carries both `opus` upshifts, while the second half carries the
exit gate and the close, so splitting would put a second close and a
second gate cycle between the two expensive things rather than between a
cheap thing and an expensive one.

**Take the split only if Task 4's actual spend re-derives the ceiling
past 8M.** The conductor takes the call at the Task 4 checkpoint, on the
number rather than on a feeling. **At 8M the window cap is the binding
constraint rather than the pass ceiling**, with about 7.6M of a 15M
window already spent, and the conductor's combined question at that
checkpoint names both.

**Budget:** 5M starting ceiling, re-derived after Task 4; rung 4's
combined ceiling 6.5M as a reservation; flag at 80% of the live ceiling
(4M against 5M), routed to the conductor under the 2026-09-05 grant
through 2026-09-07 09:00; STATUS after every task. Spend is scored at
close against the re-derived ceiling and against **zero attended
interactions**, which is this pass's target. The six open rulings are
applied provisionally and marked rather than blocking, so none counts as
an execution sitting.
