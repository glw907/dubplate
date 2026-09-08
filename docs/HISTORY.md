# dubplate history (born musicbox; entries keep the names of their day)

Per-pass ledger, newest first.

## Rung 4, pass 4a-ii (pass close, 2026-09-07)

Tasks 10 through 15 executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-05-rung-4a-shadow-runner.md` (revision
2.1), closing the fifteen-task 4a plan that pass 4a-i split at the Task
9 / Task 10 boundary. Rung 4a is now closed. The chain was
`wf_c97fd506-dda`: 16 agents, sonnet implementers and opus
`diff-reviewer`s, strictly sequential, 1.32M tokens.

### What landed

Task 10 (`09353f0`, fix `8326499`) added
`analysis.AdjudicationEntry.MatcherCommit` and `MatcherConfigHash`,
both `omitempty`, with the two `checkAdjudicationEntry` rules that
require the commit on a shadow-subject row and refuse either field on a
case-subject row, plus
`TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost`. Task 11
(`bef628e`) built `caseJSON`, `pendingJSON` and `storeArtifacts` in
`cmd/dubplate-server/shadow.go`, the marshalling that keeps
`internal/engine/shadow` clear of both harness shapes: `pendingJSON`
runs `analysis.CheckAdjudications` before it returns bytes, so a
malformed row never reaches disk. Task 12 (`81183fe`, fix `3c229f0`)
built `shadow.BuildReport` and the `shadow` subcommand's three verbs,
`observe`, `report` and `reevaluate`, in `newCheckCmd`'s shape, with
`main.matcherCommit` defaulting to `"dev"` and threaded through
`newShadowCmd`. Task 13 (`f7cd7eb`, fixes `f7f82b7` and `0fed5b3`)
built `shadow.Reevaluate`, the mechanical re-evaluation of an
observation stamped at a superseded matcher commit: it replays that
observation's own stored responses through `replayTransport` and
`matchAlbum` under `musicbrainz.Unmetered()` at the observation's own
`EvalDate`, and sorts every observation into `Retired`, `Restamped`,
`NewlyDiverged`, `Unevaluable` or `Skipped`. Task 14 (`9c956b8`, fix
`bb75cea`) built rung 4's exit gate, one end-to-end `observe` run over
a hand-authored fourteen-directory inbox. The post-chain round
(`85dcd37`) folded the Task 13 and Task 14 review leftovers.

The close-out `code-simplifier` pass (`2992bbb`, 9 files, +140/-217)
exported `shadow.StoreDir` to retire a twice-written directory literal,
folded four identical merge loops in `BuildReport` onto one helper,
collapsed `Reevaluate`'s three outcome arms onto one shared
restamp-and-write block, replaced a hand-rolled map comparison with
`maps.Equal`, and replaced collect-then-sort and membership loops with
`slices.Sorted(maps.Keys)` and `slices.Contains`. Simplify review:
accept, no blockers; the reviewer confirmed the outcome switch keeps the
write order and the agreement-to-agreement skip, that `maps.Equal` is
strictly stricter than the helper it replaced, and that the restamp
assertions now read from disk rather than the mutated pointer.

Also landed this session, outside the pass but inside this ledger's
window: the streamer 2.0 initiative. Its design spec
(`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, revision
2.1) is ratified after four reviews, `ROADMAP.md` carries the
`streamer-2-0` initiative, and CLAUDE.md carries the build-for-2.0 and
prior-art rules, the build spec's 2026-09-07 erratum, and the SvelteKit
static web rule. The repo also gained an MIT `LICENSE`.

### What the gate caught

Two tasks accepted on the first review, Task 11 and Task 12's second
round. Task 10 drew a `fix`: the `ReleaseMBID` doc comment landed only
half of owner ruling 2, naming the field as the release under dispute
and dropping "filled by whichever side proposed one," which is the
clause that answers which MBID goes in the field. Task 12 drew a `fix`
for a printed row rather than for logic: the report emitted the
`go-applies-oracle-skipped` row with its off-box note but omitted the
divergences-by-disposition row the criteria name, and the absence was
explained only in a Go doc comment an operator never sees.

Two tasks went past the one-re-dispatch rule and were ruled by the
conductor.

- **Task 13** drew a `fix` for a comment misattachment, landed as
  `f7f82b7`, then a second `fix` on a real assertion gap: nothing read
  `resolved/<obs-id>.json` back, and the reviewer proved the gap by
  swapping `OriginalCommit` with `RetiringCommit`, `OriginalClass` with
  `NewClass` and `OriginalGo` with `NewGo` all at once with the package
  still green. The conductor ruled the second fix rather than accepting
  a deliverable no test could fail on. `0fed5b3` reads the resolved
  record back, restamps a retired observation so the per-commit report
  row drains, merges the `PendingRowFor` pass-through into one exported
  function, and dates the resolved record with `time.DateOnly` to match
  the pending rows. Its review accepted, with three swap mutations run,
  each firing two assertion misses.
- **Task 14** **escalated**. Named mutations 4 and 6 did not fire, and
  the implementer had written the mutation proofs as design commentary
  at each defending assertion rather than running them. The reviewer
  ran four of the six itself and reported the two that were unfired.
  The conductor ruled a fix on `opus`, landed as `bb75cea`: all six
  mutations run, firing 7, 8, 4, 2, 3 and 2 assertion misses in order.
  The wall-time bound is a `time.AfterFunc` timer whose `Stop` return
  is the assertion, because `TestEngineNeverReadsTheClock` bars
  `time.Now` from every test under `internal/engine`. Determinism now
  builds two trees rather than one. Row 04 is un-tag-logged, so its
  review-tree collision suffix is load-bearing and the test asserts
  which reader produced its oracle signal.

**Conductor rulings taken this pass.**

1. **B7 stands.** `newShadowCmd` keeps the injected `matcherCommit`
   parameter. Resolution R3 rejects the package variable the
   alternative would have to write.
2. **Mutation proofs are run, never narrated.** A named mutation is
   applied to the source, the suite is run red, and the source is
   restored. This is now written into
   `.claude/agents/dubplate-implementer.md`, the repo's own executor,
   so it binds every later rung rather than this pass alone.
3. **The exit gate's figures are 14 directories, 3 uploaders and 13
   stored observations.** The plan's single-uploader fixture is
   unachievable alongside its own row-04 backstop
   (`reconcile.go:319-323`) and its per-uploader partial-skip flag
   (`reconcile.go:103-106`, `:325-328`). Uploader `pk` carries row 09
   and the ambiguous-basename partner, `bo` carries row 07, `ada`
   carries the other eleven.
4. **Retired observations are restamped; agreement-to-agreement
   observations are not.** The asymmetry is recorded rather than
   changed: a retire has a report row to drain, and a replay that
   agrees the way it already agreed has nothing to record.
5. **The plan's mutation-6 narration is a half-truth.** Dropping
   `buildInbox`'s `os.Chtimes` calls fails
   `TestRunIsDeterministicUnderAFrozenClock` only.
   `TestObservePassIsIdempotent` is insensitive to mtime pinning,
   because a second `Run` inside one process reads the same unpinned
   mtimes.

### The shadow contract coverage, S1 through S23

The plan's contracts, each with the Go site that implements it and the
test that proves it. A contract no end-to-end run can reach names the
hand-authored fixture instead.

| # | Contract | Go site | What proves it |
| --- | --- | --- | --- |
| S1 | `tag.Album` is the fixed entry point | `shadow/match.go` `matchAlbum` | `TestMatchAlbumTakesTheIDConsensusEarlyReturn`, `TestMatchAlbumSearchesWhenThereIsNoConsensus` |
| S2 | The estate tunables are `QuietImportTunables()` | `matcher/tunables.go`, `shadow/match.go` | `TestMatchAlbumRunsUnderFullCompatAndQuietImportTunables`; `conform`'s drift guard against `gradeTunables("estate")` |
| S3 | The oracle's signal is binary | `shadow/classify.go` | `TestClassifyVerdictReadsTheStoredAction` |
| S4 | Four outcomes, three of them divergences | `shadow/classify.go` | `TestClassifyEveryOutcome`; all three classes in `TestObservePassEndToEnd` |
| S5 | Four deferring branches, closed deliberately | `shadow/classify.go` | `TestClassifyUsesErrorsIsRatherThanEquality`; `TestObservePassEndToEnd` (unavailable, budget), `TestObservePassDefersEverythingOnAClosedLimiter` (limiter closed) |
| S6 | No error but `ErrNoCandidates` is a Go skip | `shadow/classify.go` | `TestClassifyEveryOutcome`'s deferral rows |
| S7 | The capture digest is the identity | `shadow/capture.go` | `TestCaptureDigestMovesWithEveryFileProperty`, `TestCaptureIsIdempotentUnderAnUnmovedDigest`, `TestHasObservationIsTheIdempotenceCheck` |
| S8 | A partial read is discarded, never stored | `shadow/capture.go` | `TestCaptureDiscardsAPartialRead`, `TestCaptureTreatsAnUnreadableAlbumDirectoryAsAPartialRead` |
| S9 | The grouping rule, with beets' artist fallback | `shadow/capture.go` | `TestCaptureExcludesAMultiAlbumDirectory`, `TestCaptureExcludesAFlatArtistOnlyMultiAlbumDirectory` |
| S10 | The join key is the basename, never `SourceDir` | `shadow/reconcile.go` | `TestReconcileJoinsAnImportedRowByContentNotBySourceDir`, `TestReconcileResolvesASkipFromEitherBackstop`, `TestReconcileExcludesGateOutcomes`, `TestReconcileClaimsGateRejectedOnlyThroughReviewTree` |
| S11 | Four `unreconciled` reasons | `shadow/reconcile.go` | `TestReconcileUnreconciledReasons`, `TestReconcilePartialSkipIsUnreconciledNotASkip`; all four counted in `TestObservePassEndToEnd` |
| S12 | The reports are a rolling window, and gaps are named | `shadow/oraclereport.go`, `shadow/reconcile.go` | `TestReadRunReportsSortsByStartedAt`, `TestReconcileDetectsAReportGap`, `TestReconcileNoGapWhenReportsStillPresent` |
| S13 | The store layout, rooted at one directory | `shadow/store.go` | `TestStoreRoundTripsEveryArtifact`, `TestStoreWritesNothingOutsideItsRoot`, `TestStoreLeavesNoTempFile` |
| S14 | `passes/` exists because a `Budget` is in-process state | `shadow/store.go`, `shadow/report.go` | `TestBuildReportSumsTheStore` (11, not 3 and not the maximum), `TestPassIDCollisionSecondWriteWins` |
| S15 | The budget, the pace and the contact | `shadow/config.go`, `cmd/dubplate-server/shadow.go` | `TestDefaultIntervalIsTwiceMusicbrainzDefault`; the cap asserted in `TestObservePassEndToEnd` against `observeBudgetCap` |
| S16 | Mode one records, mode two re-scores | `shadow/rescore.go` | `TestRescoreDefaultScoresUnderDefaultCompat`, `TestRescoreDefaultMarksItsVerdictAsAHypothesis` |
| S17 | Any replay client takes `Unmetered()` | `shadow/rescore.go`, `shadow/reevaluate.go` | `TestRescoreDefaultIssuesNoSearchAndSpendsNoToken`, `TestRunSpendsNoTokenOnTheRescore`, `TestReevaluateSpendsNoBudgetAndReachesNoNetwork` |
| S18 | The replay transport refuses a miss with 501 | `shadow/transport.go` | `TestReplayTransportRefusesAMissWithoutA404`, `TestRescoreDefaultReportsAMissRatherThanDroppingACandidate`, `TestReevaluateReportsUnevaluableRatherThanRetiring` |
| S19 | The engine marshals neither harness shape | `cmd/dubplate-server/shadow.go` | `TestCaseJSONRoundTripsThroughLoadCase`, `TestStoreArtifactsWritesOneFilePerObservationAndOnePerNewsRow`; `conform`'s `TestEngineNeverImportsHarness` |
| S20 | Every staged row is validated before it is written | `cmd/dubplate-server/shadow.go` `pendingJSON` | `TestPendingJSONValidatesBeforeItReturns` (four-row table), `TestPendingJSONNeverEmitsATodo` |
| S21 | The two schema additions and their rules | `analysis/adjudication.go` | Three inline table rows (`shadow-subject-no-commit`, `case-row-with-matcher-commit`, `case-row-with-matcher-config-hash-only`) plus `TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost` |
| S22 | Two clauses gate, not one | `shadow/report.go`, `cmd/dubplate-server/shadow.go` | `TestBuildReportSumsTheStore`, `TestReportPrintsThePinnedRowForm`; the second clause's disposition join is declared off-box (gap 4) |
| S23 | `disposition` stays owner-only | `cmd/dubplate-server/shadow.go` `pendingJSON` | `TestPendingJSONValidatesBeforeItReturns`'s `OracleWrong` row; the shadow writes into the committed ledger never |

### The declared coverage gaps

Nine, on one criterion: a surface 4a itself builds that no test in this
pass can reach, or a real input shape the design deliberately does not
observe. None blocks.

1. **An archive upload is invisible to the capture** (Resolution R19).
   `stageArchive` extracts a `.zip` and removes it from the inbox, so
   the album that later appears in a run report has a basename the
   shadow never saw.
2. **The re-scored `DefaultCompat` verdict is a hypothesis, not a
   measurement**, wherever the real flow would have widened the
   candidate set. That is ruling 3's own limit, and the row says so.
3. **`Budget.wait` cannot be injected per album** (Resolution R7), so
   the deferral branches are proven in two runs rather than one.
4. **The report's disposition rows cannot be computed on the box**
   (Resolution R21). Owner dispositions live in the corpus on the
   workstation, and the launch gate does that join.
5. **`atomicfile`'s temp-directory property is unproven on a
   single-filesystem workstation**, held by the doc comment plus 4b's
   `ReadWritePaths` argument.
6. **No test exercises a real library tree.** `readFLAC` runs against
   an injected reader in the reconcile tests and against built FLACs in
   the end-to-end run. A real beets-written library file is first read
   at 4b's deploy.
7. **The shadow's own timing against a live import is untested.** The
   quiescence rule is a constraint on 4b's timer interval, and 4a
   proves no part of it.
8. **"`internal/engine/shadow` imports `internal/config` never" has no
   mechanical guard.** Conductor ruling C3 keeps this a declared gap
   with the `diff-reviewer` as the meanwhile guard, rather than
   widening this pass's `conform` territory.
9. **The cross-filesystem half of `atomicfile`'s rename property is
   untested**, distinct from gap 5's temp-directory half. Nothing
   proves what `os.Rename` does across an EXDEV boundary.

Joined to rung 3's twenty-one, **the running total is thirty**. Rung 15
reads that number rather than re-counting.

### Evidence

`bash scripts/check.sh` passed clean at every commit's review point.

The exit gate, `TestObservePassEndToEnd`, drives one `observe` pass over
fourteen album directories under three uploaders and asserts parsed
structure only, with no golden file anywhere:

- 2 agreements (row 01's ID consensus, row 02's empty search).
- 1 divergence of each class: `go-skips-oracle-applied` (03),
  `go-applies-oracle-skipped` (04), `go-applies-different-release`
  (05).
- 2 excluded: 1 multi-album (06), 1 gate-rejected (07).
- 4 unreconciled reasons, 5 rows: no release id (08), partial skip
  (09), ambiguous basename (10 and its structural partner), report
  swept (11).
- 2 deferred: `ErrUnavailable` (12), `ErrBudgetExhausted` (13).
  `ErrLimiterClosed` is proven separately by
  `TestObservePassDefersEverythingOnAClosedLimiter`, per gap 3.
- 3 pending rows, one per divergence, each carrying a `release_mbid`.
- 13 stored observations: the fourteen captured directories less row
  06, which capture excludes before it ever becomes an observation.
- `started-at` holds the pass clock, and `ClockStarted` is true.

**The budget arithmetic.** `observeBudgetCap` is 12, a fixed literal
rather than a figure derived from the fixture table: rows 01 through 05
reserve one token each, and row 12's seven exhausted retry attempts
spend one each, for 5 + 7 = 12. Row 13's reserve is blocked and spends
nothing. `Budget.Spent` is asserted equal to the cap, so the cap is a
cap and not a decoration.

**Measured wall times.** `go test ./...` 2m15s to 2m27s;
`internal/harness/conform` 132.8s to 143.0s; `internal/engine/shadow`
0.175s to 0.2s; `bash scripts/check.sh` 1m6s to 3.5 minutes. The
`internal/engine/shadow` figure is two orders of magnitude inside the
five-second wall bound the exit gate asserts.

**The conform figures are untouched by this pass**, read from the 4a-i
entry rather than re-measured: tier 0 still replays 72/72, translation
still grades 50/50, the request tier still runs 215/303/48, and tier 2
still produces 215 verdicts. Nothing in `internal/engine/shadow`
reaches the conform tree.

### The mutation table

Every named mutation across Tasks 10 through 14, each applied to the
source, run red, and reverted.

| Task | Mutation | Outcome |
| --- | --- | --- |
| 10 | Drop the shadow-row `MatcherCommit` requirement | fires, `shadow-subject-no-commit` |
| 10 | Narrow the case-row guard to `MatcherCommit` only | fires, `case-row-with-matcher-config-hash-only` |
| 10 | Drop `omitempty` from `MatcherCommit` | fires, the ledger round trip |
| 11 | Skip the `CheckAdjudications` call | fires, the three refusal rows |
| 11 | Substitute `analysis.OracleWrong` for `Pending` | fires, two `checkAdjudicationEntry` rules |
| 11 | Emit a literal `TODO` for `ReleaseTruth` | fires, `TestPendingJSONNeverEmitsATodo` |
| 11 | Corrupt `PathQuery` to a non-canonical form | fires, `cassette.LoadCase` rejects |
| 12 | Let `observe` run on a `"dev"` stamp | fires, `TestObserveRefusesAnUnstampedBuild` |
| 12 | Report the last pass's budget instead of the sum | fires, `TestBuildReportSumsTheStore` |
| 12 | Call `time.Now` inside `BuildReport` | fires, `TestEngineNeverReadsTheClock` |
| 12 | Print map rows in map order | fires, `TestReportPrintsMapRowsInSortedKeyOrder` |
| 12 | Print to `os.Stdout` rather than the injected writer | fires, `TestShadowVerbsAreRegisteredAndPrintToStdout` |
| 13 | Answer a replay miss with 404 rather than 501 | fires, `TestReevaluateReportsUnevaluableRatherThanRetiring` |
| 13 | Drop the newly-diverged arm | fires, `TestReevaluateRecordsANewDivergenceOnAnObservationThatAgreed` |
| 13 | Re-run at `now` rather than the stored `EvalDate` | fires, `TestReevaluateRetiresADivergenceThatNoLongerReproduces` |
| 13 | Give the replay a metered budget | fires, `TestReevaluateSpendsNoBudgetAndReachesNoNetwork` |
| 13 | Restamp an unevaluable observation | fires, `TestReevaluateReportsUnevaluableRatherThanRetiring` |
| 13 (fix) | Swap the resolved record's commits, classes and verdict pairs, three swaps | each fires, 2 assertion misses each |
| 14 | Make `capture` skip the grouping check | fires, 7 misses |
| 14 | Make `reconcile` treat a `rejected` row as a skip | fires, 8 misses |
| 14 | Set the budget cap to zero | fires, 4 misses |
| 14 | Drop `Retry-After: 0` from the test server | fires, 2 misses, through the `AfterFunc` bound |
| 14 | Reverse the fixture's numeric prefixes | fires, 3 misses |
| 14 | Drop `buildInbox`'s `os.Chtimes` calls | fires, 2 misses, `TestRunIsDeterministicUnderAFrozenClock` only |

Four mutations named in 4a-i stay unfired, each with the guard that
holds the property instead, and none is re-litigated here: Task 1's
mutations 1 and 3, Task 5's mutation 2 and Task 8's mutation 2. The 4a-i
entry carries their reasons.

### Non-blocking leftovers, condensed

- `printCountRows` prints nothing for an empty map, so a launch-gate
  reader cannot tell "zero divergences" from "this report does not
  compute divergences." Every vocabulary is a closed constant set, so
  printing the known keys at zero is available.
- The store path literal `filepath.Join(cfg.StateDir, "shadow")` is
  written in `cmd` twice, duplicating the engine's own unexported
  `storeDirName`. An exported `shadow.StorePath(stateDir)` would keep
  the layout in one place.
- `report` calls `OpenStore`, which creates the store directory as a
  side effect of merely printing a report on a box that has never run
  `observe`.
- The printed `go-applies-oracle-skipped` count is the raw count, while
  the gated figure is the subset the owner has not disposed
  `oracle-wrong`. The parenthetical carries the caveat and the raw
  count can only over-report, so it fails safe.
- `Report.MatcherCommits` counts non-agreement observations only, so a
  commit observed under with zero divergences never gets a row.
- Any replay error in `Reevaluate` that is neither a recording miss nor
  `tag.ErrNoCandidates` aborts the whole call, discarding outcomes
  already computed. `classify` defers per observation instead.
- `stageNewlyDiverged` and the `reevaluate` verb's `RunE` have no test
  in `cmd/dubplate-server`; `shadow_test.go` asserts only that the verb
  is registered.
- The four committed `testdata/observe` bodies are byte-for-byte what
  the in-package `releaseBody` and `searchBody` helpers already
  generate, so they are a second source of truth that can drift.
- 4a-i's own leftover list stands unchanged: no directory fsync, the
  non-`.json` decode in `Observations()` and `Passes()`, the uncapped
  consumed ledger, the silent content-join drop, the unsynchronized
  `recordingTransport`, and the terminal deferral.

### Filed for other owners

- Renaming `case.json`'s `Profile` the day the corpus grows a
  `quiet-import` profile, for lane E1 (Resolution R8).
- A `KindFloat` threshold variant, for the first rung 3 follow-on after
  rung 1b lands (forward carry 3).
- The 4a-i filings stand: rung 8 absorbs the FLAC reader, rung 16
  retires the oracle report decode, and rung 1b absorbs
  `integrationClasses` as data.

### What a later pass would be wrong to rediscover

1. **Mutation proofs are run, never narrated.** A mutation described in
   a comment beside the assertion it is supposed to defend proves
   nothing, and Task 14 shows why: two of six such mutations did not
   fire, and no reader could tell from the commentary. The rule is now
   in `.claude/agents/dubplate-implementer.md`.
2. **The exit gate's uploader count is forced by the code, not by
   taste.** `reconcile.go` scopes the partial-skip flag per uploader
   and lets the first review-only capture under an uploader consume the
   gate-rejected count, and an ambiguous basename is by construction a
   collision between two uploaders. Any fixture that wants a clean
   partial skip, a clean gate rejection and an ambiguous basename in
   one pass needs three uploaders.
3. **`TestEngineNeverReadsTheClock` covers tests too.** Nothing under
   `internal/engine` may call `time.Now`, so a wall-time bound in an
   engine test is a `time.AfterFunc` timer whose `Stop` return is the
   assertion, never a stopwatch.
4. **Retire restamps, agreement-to-agreement does not.** A retired
   observation is restamped so `BuildReport`'s per-commit row drains
   and a second `reevaluate` does not re-retire it. An observation that
   agreed and still agrees is left untouched, because there is no row
   to drain. The asymmetry is deliberate.
5. **`os.ReadDir` mtime granularity is the digest's assumption.** The
   capture digest ingests modification times, so a fixture that does
   not pin mtimes with `os.Chtimes` produces a different digest on
   every run. The determinism test is what catches that, and it needs
   two separately built trees to catch it at all.
6. **A retire is only idempotent because the observation is
   restamped.** `RemovePending` tolerates `ENOENT` and `PutResolved`
   overwrites, so without the restamp a second pass rewrites the
   resolved record with a fresh date and reports the retire again.

### The rung 4 spec corrections, applied at this close

Each is marked in the spec as "(corrected 2026-09-07 at the 4a-ii
close)": `spec:1608-1609`'s "(ruling 7)" reads ruling 6, the deploy
sitting; `spec:299-301`'s pre-fold one-request-per-second sentence
follows the conductor's pace ruling; `spec:1267`'s "the four
`DUBPLATE_SHADOW_*` rows" reads six; `spec:845-848`'s byte-for-byte
claim reads as the round trip that actually holds, and its fixture form
reads as an inline Go literal variant; `spec:1616`'s exit gate names
only the deferral branches one pass can reach; and `spec:1063-1066`'s
`newShadowCmd` signature carries the matcher stamp. The `spec:441`
grouping-key correction was already applied at the plan's own fold and
needed nothing here. Rung 3d's own "rung 3 to rung 4 handoff"
section carries the matching wire-pace amendment.

### The six owner rulings, ratified

Geoff ratified all six on 2026-09-07 (`57b49fd`), each as recommended:
the integration-class wall, the release under dispute filled by
whichever side proposed it, both compat modes, the 720h floor, the
60-request cap at twice the limiter's own interval, and the deploy as an
owner sitting. **The pass's batched owner ask is therefore answered
rather than open.** What remains is mechanical: every provisional
application still carries a "provisional, owner ruling N open" marker in
the landed code, and retiring those markers is a search rather than a
re-reading. 3d's inherited ask, re-dispositioning
`bug-assignment-rejection-gap` from `oracle-wrong` to
`no-correct-answer`, rides its own stream and is not rung 4's to raise.

### The 4a to 4b handoff

What 4b's plan needs from 4a, stated rather than derived.

1. **The store's root and its footprint.** `<state dir>/shadow/` holds
   `passes/`, `observations/<obs-id>/`, `pending/`, `resolved/`,
   `started-at` and `consumed-reports.json`, and nothing is written
   anywhere else. `TestStoreWritesNothingOutsideItsRoot` proves it, so
   that root is 4b's `ReadWritePaths=` argument, static rather than
   assumed.
2. **The unit's `ExecStart` and its refusal.** `dubplate-server shadow
   observe` errors rather than observing when `main.matcherCommit` is
   empty or `"dev"`, and creates no store on either refusal. No build
   path sets the stamp today. **4b's deploy edit 6 is what makes the
   refusal observable**: without the two `-ldflags` stamps and the
   smoke-test assertion, an unstamped binary deploys clean, every timer
   pass refuses, and the `OnFailure=` chain fires into a slug whose URL
   is unset and therefore no-ops. `reevaluate` carries the same
   refusal.
3. **The six config rows and how they render.** The three paths are
   required-when-enabled and abort `Load` when the flag is set without
   them. The flag and the two thresholds are accepted and never
   required. An explicit `DUBPLATE_SHADOW_ENABLED=false` loads clean
   with a nil `Shadow`, which is the adopter's own way to turn the
   integration off. The greenfield template carries six
   `DUBPLATE_SHADOW_*` rows, not four.
4. **The wire pace and the cap.** `shadow.DefaultInterval` is
   `2 * musicbrainz.DefaultInterval` and the cap is sixty tokens, both
   now ratified under ruling 5 and both one literal away from a
   different value. Sixty tokens at two seconds a token bounds a pass at
   two minutes, comfortably inside 4b's five-minute timer.
5. **The operator rule the unit's comment must carry.** A hand-run
   shadow verb goes through `systemctl start dubplate-shadow.service`
   or `systemd-run` with the same properties, never the bare binary as
   root. The read-only guarantee is the unit's, not the binary's.
6. **The clock's definition** (Resolution R11). `started-at` is written
   on the first pass that completes all three steps without a fatal
   error, including a pass that observes nothing, so the clock starts at
   the deploy rather than at the first upload. **4b's close records that
   date as a first-class figure**, because the launch date is that date
   plus the floor.

### Forward carries

1. **The rung 15 re-run criterion** travels forward unchanged, as it
   has from every rung 3 pass close, now joined by the shadow's own
   launch-bar clause.
2. **The declared coverage gaps are a running total of thirty**, rung
   3's twenty-one plus 4a's nine. Read it, never re-count it.
3. **The four-knob wall registration** stays the named follow-on for
   the first pass after rung 1b, and **rung 4's own `KindFloat` need
   joins it**. `Threshold.Kind` now has two values and the three float
   rec thresholds want a third. `AdminEmail`'s dedicated contact row
   sits beside both.
4. **Rung 1b absorbs `integrationClasses` as data**, not as a shape to
   invent, and the table's `Required`/`Optional` split is the semantics
   it must preserve.
5. **Rung 8 absorbs the FLAC reader** into `internal/engine/tagio` and
   repoints the shadow. The nineteen-field mapping with its per-row
   citations travels in 4a-i's entry, so rung 8 does not re-derive it.
   The deployed binary carries landed-but-unwired packages from other
   lanes, and none is reachable from the `shadow` subcommand's own call
   graph; repointing is rung 8's task, never a silent adoption.
6. **Rung 9 re-verifies the limiter's caller-priority policy against a
   measurement.** 4a is the first `ClassBackground` caller and 4b puts
   it beside a live import. 4a's own figure is the exit gate's spend of
   twelve tokens; the live one arrives after 4b's deploy.
7. **Rung 6's store supersedes none of this**, and `internal/lock`
   stays unadopted. The timer path is a systemd singleton, `reevaluate`
   writes only under `resolved/`, `report` opens nothing for writing,
   and `observe` only ever creates.
8. **The response store is a consequence of ratified ruling B rather
   than a preference.** The retention knob is named rather than built:
   the trigger is the report's own store-size row against the root
   filesystem's free space, which 4b's sitting records as a baseline.
9. **The streamer 2.0 forward hooks are binding on rungs 5 through 9
   and rung 12**, through CLAUDE.md's build-for-2.0 and prior-art rules
   rather than through this pass's own carries.

**Resume prompt for the next session.** Rung 4b is next, in the main
tree: the deploy, whose close needs an owner sitting per ruling 6. Its
plan is owed just-in-time and does not exist yet.

**Budget.** 4a-ii spent about 2.7M against its 3M ceiling (Geoff,
2026-09-07), an estimate from the agent meters: staleness pre-task and
chain-args extraction 0.26M, chain 1.32M, the Task 13 and Task 14 fix
and review rounds 0.49M, simplify 0.18M plus its review, the close
drafter 0.16M, and the conductor's own turns for the pass about 0.2M.
The 2.4M flag was crossed at the close, after every task had landed,
by the simplify and close rounds; nothing was cut. The 2.0 brainstorm's
reviews and fold (about 0.9M) ran in the same session and are not
counted here. The chain alone was 1.32M across 16 agents. Attended
interactions for 4a-ii: planning misses 2 (the exit gate's
single-uploader fixture figure, and the mutation-6 narration, both of
which a probe at planning would have caught), execution sittings 0. The
conductor asked Geoff nothing during execution.

## Rung 4, pass 4a-i (intermediate close, 2026-09-05)

Nine plan tasks executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-05-rung-4a-shadow-runner.md` (revision
2.1), which splits the fifteen-task 4a plan at the Task 9 / Task 10
boundary on CLAUDE.md's pass-sizing rule. This entry is that
intermediate close, not the pass's own close: Tasks 10 through 15 land
in pass 4a-ii, a later session.

### What landed

Task 1 (`bc21683`, fix `a1d3d6c`) re-landed `internal/atomicfile`
unchanged from `_salvage` on merit, plus the plan's crash-safety test
and a follow-up subtest pinning the temp file's directory. Task 2
(`46f4229`, fix `751bdde`) built `matcher.QuietImportTunables`, the
engine-side constructor for the four quiet-import assignments, with
`conform`'s drift guard against `gradeTunables("estate")`. Task 3
(`f20f329`, fixes `afeb007` and `8b73af8`) built the integration-class
wall (`IntegrationClass`, `Config.Shadow`, required-when-enabled
resolution), `ThresholdKind` and `ParseDuration`, and doctor's
integration and duration rows. Task 4 (`adcf66d`) built
`internal/engine/shadow`, with the provisional FLAC container reader
and the nineteen-field mapping onto `matcher.ItemFields`. Task 5
(`5966b8a`, fix `60221e9`) built the shadow store, one directory per
observation under `observations/`, `passes/`, `pending/`,
`resolved/`, every write through `atomicfile.WriteFile`. Task 6
(`d1603ec`, fix `891dc1a`) built capture, the lexical
uploader/basename walk, the grouping rule, the digest, and the
partial-read discard. Task 7 (`b0aef65`, fixes `d1e09dc` and
`6bbcbe0`) built the oracle report decode and the reconcile step,
joining captures to run reports by content and basename rather than
by `SourceDir`. Task 8 (`4c8bcc4`, fix `ffabf71`) built the recording
and replay transports, `shadow.Config`, and `matchAlbum`, plus the
wire-pace constant `DefaultInterval`. Task 9 (`817c75f`) built the
`DefaultCompat` re-score, the divergence classification, and `Run`'s
five-step pass loop. The close-out `code-simplifier` pass (`ea1d0ff`,
3 files, +12/-34) folds three collect-then-sort loops onto
`slices.Sorted`/`SortedFunc`, replaces a hand-rolled newest-capture
scan with `slices.MaxFunc`, and drops a test helper that reimplemented
`strings.Join`. Simplify review: accept, no blockers; every rewrite
verified against the go1.27.1 stdlib and the newest-capture tie rule
confirmed unchanged.

### What the gate caught

Five tasks needed a fix round from the in-chain `diff-reviewer`, two
of them twice. Task 2 attached a citation to the wrong clause. Task 3
carried a stale citation, then an unproven budget threshold. Task 5's
test comment claimed coverage the store's own naming scheme cannot
provide. Task 6 aborted the whole pass on an unreadable album
directory instead of recording a partial read. Task 7 carried a stale
provisional-ruling comment, then a gate-rejected capture claimed by
the wrong row.

Two tasks escalated to the conductor rather than taking a
`diff-reviewer` fix.

- **Task 1** flagged that the plan's own mutation-1 narrative was
  wrong: moving `os.CreateTemp` to `os.TempDir()` is not caught by
  either the leftover-temp-file test or the new crash-safety test,
  because `os.Rename` needs write permission on the destination
  directory regardless of where the temp file lives. Ruled: add one
  extra test beyond the plan's fold, pointing `TMPDIR` at an
  unwritable directory. Landed as `a1d3d6c`.
- **Task 8** escalated on two points: whether `shadowScoring` is the
  legitimate mutation site for named mutation 6, and whether the wire
  pace and the budget cap owed a materialized constant rather than a
  doc comment. Ruled: `shadowScoring` is the accepted mutation site;
  `DefaultInterval = 2 * musicbrainz.DefaultInterval` is a named
  constant with a pinning test, landed as `ffabf71`; the 60-token cap
  stays Task 3's registered threshold, and wiring it into the
  `ClassBackground` budget is Task 12's job in 4a-ii.

The reviewer found one race and one misattributed claim beyond the
review's own blocking-list shape. Task 6's reviewer traced a genuine
race with `beets`: `review.Move` renames the album directory itself,
so a walk that lists a basename can take `ENOENT` reading that
directory's own entries, and the first cut aborted the whole capture
pass on that error instead of recording a partial read. Task 7's
second review found the gate-rejected exclusion claimed by whichever
capture under an uploader matched first, so a beets-skipped capture
could steal the exclusion meant for a genuinely rejected one and the
rejected album would surface a fabricated oracle skip signal; fixed
in the post-chain round as `6bbcbe0`.

Three mutations were recorded as unprovable rather than papered over.
Task 1's mutation 1 (moving the temp file out of the target's own
directory) was unfired by the original test pair, as above, before
the extra subtest closed it. Task 5's mutation 4 (dropping the
explicit sort in `Observations`) cannot fail with this store's naming
scheme: `os.ReadDir` already returns filename order and the directory
name is the id, so no test tree can make the two orders diverge.
Task 9's mutation 5 (handing the replay client the pass budget) is
unreachable without a signature change the plan's own fold predicted;
proven only by a temporary signature change that confirmed the effect,
then reverted.

### Evidence

`bash scripts/check.sh` passed clean at every commit's review point.
`go test ./...` runs about 2m06s, dominated by `internal/harness/conform`'s
sweep (unchanged from rung 3). The conform sweep's own figures (72/72,
50/50, 215/303/48, 215/294/294/214, 215 verdicts) are untouched by this
pass; nothing in `internal/engine/shadow` reaches the conform tree.

### The six owner rulings, applied provisionally

Each is written into the landed code as "provisional, owner ruling N
open." One line each, with what reverses if Geoff rules the other way.

1. **Ruling 1** (integration-class shape): option (a), the
   `IntegrationClass` table, is built. Reversing to (b) or (c) retires
   `internal/config`'s integration-class wall and moves the shadow's
   paths onto flat accepted keys or bare cobra flags.
2. **Ruling 2** (a skipped album's `ReleaseMBID`): filled by whichever
   side proposed a release, oracle's or Go's. Reversing to option (b)
   relaxes `checkAdjudicationEntry`'s pair rule for shadow subjects
   specifically.
3. **Ruling 3** (which `Compat` the shadow runs): both, `FullCompat`
   scored and `DefaultCompat` re-scored from the same candidate set.
   Reversing to (a) or (b) drops the re-score step and Task 9's
   classification collapses to one verdict pair.
4. **Ruling 4** (the estate's shadow floor): the shipped 720h default.
   Reversing changes only the registered threshold's default; no code
   branches on the value.
5. **Ruling 5** (the MusicBrainz budget cap): 60 requests per run at
   `2 * musicbrainz.DefaultInterval`. Reversing to `DefaultInterval`
   moves the shadow's wire pace back to one request per second and
   reopens the shared-quota mitigation question ruling 5 itself
   resolves.
6. **Ruling 6** (the first deploy, sitting or conductor action): an
   owner sitting. This ruling touches only 4b, out of this pass's
   scope; nothing in 4a-i depends on its answer.

### Conductor rulings taken

C1 through C3 landed at the plan's fold, before any task dispatched.
C1, the capture's grouping key is beets' own,
`cmp.Or(AlbumArtist, Artist)` and `Album`. C2, `newShadowCmd` takes
`matcherCommit string` as a parameter and `newRootCmd` passes the
package var. C3, `internal/engine/shadow` importing no
`internal/config` stays a declared gap, guarded by the in-chain
reviewer rather than an architectural test. The wire pace ruling is
not one of C1 through C3. It was taken at the spec's own fold, ahead
of this plan, and amends the 3d handoff's recorded `DefaultInterval`
to `2 * musicbrainz.DefaultInterval`.

Nine further rulings came out of the task records themselves. Task 1,
add the TMPDIR subtest. Task 2, accept. Task 3's first round, re-point
the citation and pin the budget threshold. Task 4, accept, with a TDD
ordering deviation noted. Task 5, accept, mutation 4 unfireable and
recorded. Task 6, accept. Task 7's second round, restrict the
gate-rejected claim to review-tree-only captures. Task 8, accept
`shadowScoring` and materialize `DefaultInterval`. Task 9, accept,
mutation 5 unreachable without a signature change.

### Non-blocking leftovers, condensed

- Atomic file writes carry no directory fsync, so the crash-safety
  claim covers torn content but not the rename's own durability
  (inherited from `_salvage`, unchanged).
- The shadow store's `Observations()` and `Passes()` decode any
  non-`.json` entry a crashed `atomicfile` write might leave behind;
  filtering on the `.json` suffix would make both read paths
  crash-tolerant.
- `capture.go`'s uploader-level `os.ReadDir` tolerance (extended past
  the review's ask, to the album level alone) has no test of its own.
- `reconcile.go`'s consumed ledger only ever grows; nothing in this
  diff caps it, and at one oracle import run every fifteen minutes
  (`dubplate-import.timer`, `*:0/15`) that is roughly 35k entries a
  year.
- `reconcile.go`'s content join drops an imported row with empty
  `LibraryPaths` silently when the uploader has zero or more than one
  unclaimed capture, with no counter recording the drop.
- `match.go`'s 503 test margin (a 1s to 2s deadline) is modest on a
  loaded machine; `recordingTransport` is unsynchronized, safe only
  because requests are issued sequentially today.
- `run.go`'s `Observation.EvalDate` is the capture pass's clock while
  scoring uses the match pass's clock; nothing tests the divergent
  case because every test captures and matches in one pass.
- `run.go`'s deferral is terminal: a deferred observation carries an
  oracle signal and is never re-matched in a later pass.

### Filed for other owners

- Rung 8's `internal/engine/tagio` absorbs the shadow's provisional
  FLAC reader; `flac.go`'s own doc comment names the retirement.
- Rung 16 retires the oracle report decode (`oraclereport.go`), the
  same wind-down that retires the box-side shell import layer's other
  remnants.
- Rung 1b's registry inherits the shadow's integration class as a
  shape to absorb rather than a shape to invent, per ruling 1's own
  recommendation.

### What a later pass would be wrong to rediscover

1. `os.Rename` needs write permission on the **destination**
   directory to link the new entry, regardless of where the temp file
   was created; a chmod'd target directory fails the rename even when
   the temp file sits in `os.TempDir()`. Proving the temp-in-path
   property needs a separate `TMPDIR`-pointed-elsewhere test, not the
   crash-safety fold.
2. `os.ReadDir` already returns entries in filename order, and the
   shadow store's directory and file names are themselves the ids, so
   an explicit sort in `Observations()` or `Passes()` states a
   guarantee no test in this store can fail on removal.
3. A gate-rejected album (quarantined or rejected) never reaches
   beets and therefore never appears in a tag log; any code that
   resolves a gate-rejected capture through a backstop shared with
   beets-skipped captures must exclude tag-log matches explicitly, or
   first-match order can steal the exclusion from the wrong capture.
4. `review.Move` renames the album directory itself, so a capture
   walk can list a directory and then take `ENOENT` reading its own
   entries; that is an ordinary race to tolerate as a partial read,
   never a pass-fatal error.
5. `musicbrainz.ErrBudgetExhausted` and `musicbrainz.ErrLimiterClosed`
   are mutually exclusive within one budget: a budget that has
   already refused further spending is closed, not merely exhausted,
   and a caller branches on which sentinel it received rather than
   treating both as one retryable condition.

### 4a-ii handoff

Tasks 10 through 15 inherit the following landed surface, unchanged
unless a task's own criteria says otherwise.

- **`shadow.Config`** (`internal/engine/shadow/config.go`) and
  **`shadow.DefaultInterval`**, the wire-pace constant Task 12 wires
  into the `ClassBackground` budget alongside the 60-token cap
  (`DUBPLATE_SHADOW_BUDGET`, Task 3's registered threshold).
- **The store's layout**
  (`internal/engine/shadow/store.go`): `observations/<id>/`,
  `passes/<pass-id>.json`, `pending/<id>.json`, `resolved/<id>.json`,
  `started-at`, `consumed-reports.json`, every write through
  `atomicfile.WriteFile`.
- **`Observation`** and **`PendingRow`** (`types.go`, `run.go`), the
  engine-side shapes Task 10's marshalling into `cassette.Case` and
  `analysis.AdjudicationEntry` reads from; both stay in
  `cmd/dubplate-server` because `internal/engine/shadow` may import no
  harness package.
- **The classification's four classes**
  (`internal/engine/shadow/classify.go`): agreement (no row),
  `go-applies-oracle-skipped`, `go-skips-oracle-applied`,
  `go-applies-different-release`, plus the closed deferral set
  (`ErrUnavailable`, `ErrBudgetExhausted`, `ErrLimiterClosed`,
  `DeferUnclassified`).
- **`matchAlbum`/`shadowScoring`** (`match.go`): the accepted mutation
  site for the `FullCompat`/`QuietImportTunables` pair; `matchAlbum`'s
  signature does not change.
- **The declared gap on the config-import rule** (C3): guarded by the
  in-chain reviewer, not an architectural test. Any task that adds a
  second file importing `internal/config` from `internal/engine/shadow`
  needs the same reviewer-level scrutiny, not a free pass because no
  test enforces it.
- **The mutation sites reviewers accepted as non-literal**: Task 5's
  ReadDir-order mutation (structurally unfireable, recorded rather
  than fixed), Task 8's `shadowScoring` (the mutation site for named
  mutation 6, not `matchAlbum`'s own body), and Task 9's mutation 5
  (unreachable without a signature change).

**Resume prompt for the next session.** Task 10 is next, in the main
tree, via the pass-execute chain with this entry's common notes, after
a fresh Task 0-style staleness check against the new HEAD.

**Budget.** 4a-i spent about 4.99M: prep 2.3M, Task 0 0.07M, chain
2.14M, fix round 0.08M, simplify 0.15M, simplify review about 0.05M,
close about 0.2M (estimate). Execution against the 3M sub-ceiling was
about 2.69M, prep excluded. The plan's 5M ceiling covers all of 4a, so
4a-ii starts with roughly no headroom and needs its own ceiling before
it dispatches. That call belongs to the next conductor, not to this
close. Attended interactions for 4a-i: planning misses 1 (Task 1's
mutation-1 narrative, which a probe at planning would have caught),
execution sittings 0.

## Rung 3, pass 3d (Verdicts, the assignment solve, admission, recommendation, and the whole-corpus conformance gate): `internal/engine/tag` born, tier 1 and tier 2 grade the whole 215-case corpus (2026-09-05)

Nine plan tasks executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-05-rung-3d-verdicts-gate.md` (revision
2.1). Commits in order: `e7ff0bc` (Task 1), `e552d0a` (Task 2),
`1101a6c` plus `73af4d4` and `a78d6be` (Task 3, two fix rounds),
`19e03f7` plus `bbb65ac` (Task 4, an escalation ruled), `d89cb6f` (Task
5), `755efbd` (Task 6), `0b0db6f` (Task 7), `a69dee5` (Task 8), and
`e462210` (the close-out `code-simplifier` pass, 4 files, +61/-61).
Simplify review: accept, no blockers. The reviewer confirmed the
dropped `minD = d[final]` re-read as a no-op by stating the invariant
independently. No exported surface, contract, assertion or fixture
changed.

What landed. Task 1 built `solveLAP` in
`internal/engine/matcher/assign.go`: a hand-written dense
Jonker-Volgenant solve, structured as lap 0.5.13's own three routines,
padded to `n = max(rows, cols)` with zero, returning one entry per
column (the column-to-row orientation of fold B1). `AssignItems` builds
every cost cell with `inclArtist` false and walks the solve in ascending
track index. `ExtraItems` and `ExtraTracks` sort with
`slices.SortStableFunc`. Task 2 found and fixed a real defect in that
solver (below) and froze `testdata/lapjv-optima.json`, 404 rows mirroring
lap's own tie choice. Task 3 built `candidate.go`: `Candidates` with the
five admission gates in beets' order, insertion-order tracking, and a
stable distance sort, plus `Tunables.Validate`'s `Required` and
`Preferred` halves. Task 4 built `verdict.go` (`Recommend`'s ladder, the
`max_rec` downgrade, `ActionFor`) and `compat.go` (`Compat` as two
independent switches, `DefaultCompat`, `Apply` selecting one of four
`SD_PATTERNS` tables). Task 5 built `internal/engine/tag`, the matcher's
public entry point: `Album` ports beets' `tag_album`, with the ID-match
early return, the search fallback, and the `search_ids` branch. Task 6
added `TieEntry.Adjudicator` to `internal/harness/analysis`, required
whenever a tie row is dispositioned, refused as the `agent` sentinel on
`excluded` and accepted on `pinned`; `CheckOracleBugs` gained a
probe-case rule requiring every `oracle-wrong` hypothesis's probe cases
to resolve to an adjudication row. Task 7 built tier 1
(`internal/harness/conform/case.go`'s `GradeCases`), replaying all 215
committed cases through `tag.Album` and grading nine of
`cassette.Expected`'s eleven fields; `stub.Server` gained `Unused()`.
Task 8 wired tier 2's verdict grading into `ModeCompat` and built
`ModeDefault` as a re-score of each case's own committed candidate set
under the corrected pattern table, plus the pending-divergence path and
`CheckDefaultCompatMatchesTheLedger`.

What the gate caught. Task 1's reviewer escalated on a plan-reading
question (whether `testdata/assign/` was Task 1's or Task 2's to
create); ruled that Task 2 owns the directory and the inline Go-literal
matrices in Task 1 stand. **Task 2's mirroring test caught Task 1's
transliterated solver being genuinely suboptimal, not merely
differently tied**, on the pinned 14x14 `should-reject` matrix:
`5.283068783068783` against the recorded optimum `5.282407407407407`.
The cause was one wrong line in `findPath`'s final dual update, which
priced settled columns off the scan list's head rather than off the
augmenting path's own ending column; the two agree only while the path
ends on a column the last scan admitted. Task 1's own property test
enumerated only `n <= 6` small-integer matrices, where the defect never
fires, so the frozen 404-row table does real correctness work and must
never be thinned. No downstream figure had been derived from the
defective solver before the fix landed.
Ruling F was **not triggered**. One mirroring attempt reproduced lap
0.5.13's choice on both pinned cases and all 404 frozen rows, so no
finding document exists and `corpus/analysis/ties.json` stays
untouched by this pass.

Task 3's reviewer blocked twice. The first round found two named
mutations shipped as claimed-but-unproven. The gate-1/gate-2 admission
order is unobservable through `Add` for any input; both gates return
`(false, nil)` with no side effect, so a swap cannot be caught.
`slices.SortFunc` never reorders an all-equal-key slice at any length,
because Go's pdqsort takes a zero-swap fast path on ties, so the
three-element tie fixture could never discriminate a stable-sort
mutation. The fix added a fixture of thirteen candidates alternating
between two distance values, which does discriminate, resting on
Go's own twelve-element insertion-sort threshold (`slices.SortFunc`
diverges from `slices.SortStableFunc` starting at n=13). The gate
order itself is now held only by the `match.py:345-352` citation, with
no test able to
discriminate it; the comments were corrected to say so rather than
claim a discrimination that does not exist. The second fix round closed
a residual: a `t.Error` message still asserted the false gate-order
claim after the doc comment above it had already been corrected.

Task 4 escalated on the `max_rec` downgrade's key-resolution order. The
plan's own literal fixture keyed `tun.MaxRec` on `"nosuchpenalty"`, a
key no candidate carries, which is only detectable if `downgrade`
resolves every configured entry eagerly rather than beets' own lazy
walk over the candidate's key union (`match.py:304-322`). The conductor
ruled option (a): restore the lazy walk and change the fixture's key to
`"album"`, which the candidate fixture does carry, so the comment
becomes true as written rather than describing a divergence the port
never needed.

Task 7 escalated on scope and two undisclosed reversals. First, the
criteria's eleven-field graded-field table shipped with the verdict
pair (`Expected.Recommendation`, `Expected.Action`) unwired: `VerdictFloor`,
the verdict counters, and `casePolicy.GradeVerdict`/`VerdictFrom` were
all written and never read. Ruled: verdict grading belongs to Task 8,
which was already dispatched to consume exactly those fields. Second,
`AssignmentFloor` shipped at 214 with `TiesAwaitingRuling` empty, not
the plan's stated 212 with two pending rows; both pinned tie cases now
reproduce after Task 2's fix, so the stricter, measured 214 is accepted
and the plan's 212 figure is corrected here. Third, the tier 1 sweep
measured 59.9s against the plan's 45s trigger; ruled as a carried
finding rather than a task defect (see Forward carries below).
Mutation counts were also corrected against the real measurement; the
table below is the reconciled record. A fresh-context reviewer (opus)
re-ran the whole runner in a clean clone and accepted, reproducing
every floor, all eleven graded fields, and all five mutations
independently.

Task 8 escalated on a plan-wording mismatch (the criteria named "four
zero-instance verdict rows"; the plan's own Step 1 test code names
exactly two, `pending` and `unadjudicated`) and the same sweep-cost
finding Task 7 raised. Ruled: the plan never enumerates four rows,
Task 7's sixteen-row `policyFor` fixture set already gives every
zero-instance disposition its own firing fixture, and Task 8's two
fixtures are exactly the ones its own tests call; the "four" is a plan
wording error, not an owed deliverable. The combined two-mode sweep
costs 113.5s. That pushed `go test ./...` to 2m06s and
`scripts/check.sh` to about 3.5 minutes, both well past the plan's own
"roughly 45 seconds" trigger. Ruled that the sweep stays inside
`go test ./...`; spec `:267` remains the escape to a separate leg and
is not exercised this pass. The cost carries forward as a finding for
the first rung 3 follow-on or rung 9. Its profile is regexp
backtracking under `sdStrip` at 43% cumulative, `matcher.Levenshtein`
17.5%, `Transliterate` 7.7%.

### The verdict-path contract coverage table

Twenty contracts, V1 through V20, each with its Go site and what grades
it. A committed case grades a contract when tier 1's sweep drives
`tag.Album` or tier 2's re-score and compares the result against the
case's own record; a rule no corpus case can reach is graded by a
hand-authored fixture, with the mutation that proves the fixture fires.

| Contract | Go site | Graded by | Corpus reach |
| --- | --- | --- | --- |
| V1 normalized cost cell, `inclArtist` false | `matcher.DistanceForTrack` (`track.go:45`) via `AssignItems` | all 294 candidates' track vectors | Yes |
| V2 the solve returns a track-indexed array | `solveLAP` (`assign.go`), column-to-row orientation (fold B1) | 214 assignments; the frozen 404-row table | Yes plus fixture |
| V3 pairs in ascending track order | `AssignItems` (`assign.go`) | mutation 2, 16 cases fail (item order) | Yes |
| V4 the leftover set is one-sided | `AssignItems`'s extras | `TestAssignItemsMatchesEveryShapeOneSidedly`, all 20 shapes 0x0-3x4 | Fixture (7/14 corpus cases carry one side) |
| V5 rectangular extension is `max(rows,cols)` zero-padded | `solveLAP`'s pad step | the frozen 404-row table; mutation (pad max+1), 32 rectangular rows fail | Fixture |
| V6 extras sorted by `(disc,track,title)`/`(index,title)` | `AssignItems`'s two `slices.SortStableFunc` calls | mutation 4 (title-only sort), fires the reordered case | Fixture (total order is a corpus fact, not proven) |
| V7 five admission gates in beets' fixed order | `Candidates.Add` (`candidate.go`) | gates 3-5 fire on real cases; gates 1-2's order is unobservable through `Add` for any input, held by the `match.py:345-352` citation alone | Partial, one order pair unprovable |
| V8 `match.ignored` reads the nonzero view | `NonzeroKeys` wired at `candidate.go`'s gate 5 | hand-authored `ignored-zero-media.json` fixture | Fixture only (0 of 215) |
| V9 `match.required` tests Go zero value | `candidate.go` gate 3, thirteen-name table | `TestAddAppliesRequiredAgainstTheZeroValue` | Fixture (declared gap 3, R6) |
| V10 candidate identity is `(source,id)` | `cassette.Candidate.Source/AlbumID` | all 294 candidates | Yes |
| V11 insertion order then stable distance sort | `Candidates.Sorted` | the 13-element alternating fixture (n=13 threshold); `should-ask`/`should-reject`'s ties thinly | Thin plus fixture |
| V12 the recommendation ladder, strict-strong/inclusive-medium | `matcher.Recommend` (`verdict.go`) | mutations 1-2, `TestRecommendWalksTheLadder` | Yes, plus the hand-authored boundary ladder rows (declared gap 11) |
| V13 no-conclusion returns before the downgrade | `Recommend`'s early return | mutation 3, four committed cases (`rec-none-small-gap` and three others) | Yes |
| V14 downgrade key set is nonzero view unioned with track keys | `downgrade` (`verdict.go`) | mutations 4-5; the max_rec downgrade fires on 6 of 215 cases | Yes, thin on the track-union half (declared gap 5-6) |
| V15 `min(rec,max_rec)` over the ordered enum | `matcher.Recommendation`, `downgrade`'s lazy walk | `TestRecommendReturnsBeforeTheDowngradeOnNoConclusion` after the fix round | Yes |
| V16 the verdict-to-action mapping, quiet-first | `matcher.ActionFor` (`verdict.go`) | mutation 6 (quiet outranks timid); 215 cases, 167 apply / 48 skip | Yes |
| V17 `Quiet: true` overlaid at the grading call site | `gradeTunables`'s overlay (`conform/translate.go`) | all 215 actions; two of three `import.*` assertions are vacuous today (Timid/QuietFallback unreachable under quiet) | Yes for action, thin for the overlay's own three keys |
| V18 `tag_album`'s flow, the unsorted early return | `tag.Album` (`tag.go`) | `TestAlbumTakesTheIDMatchEarlyReturnAndIssuesNoSearch`; mutations 1-2, 4-5 | Yes, mutation 3 (sort before early return) unfirable, at most one candidate reaches that check |
| V19 `match_by_id`'s two early branches | `tag.Album`'s consensus lookup | declared gaps 2-3 (no-id, no-consensus), 0 of 215 | Fixture only |
| V20 the tracks/missing/unmatched summation order | `matcher.DistanceForAlbum`, fed `AlbumPairsFor`'s track-ordered pairs | all 294 track vectors, 55 near misses at 1-4 ULP | Yes |

### The graded-field table

Task 7's tier 1 grades nine of `cassette.Expected`'s eleven fields
(`case.go:343-436`); Task 8 adds the tenth and eleventh, the verdict
pair. Rung 15's re-run reads this table rather than deriving it again.

| Field | Grading rule | Populated by |
| --- | --- | --- |
| `Recommendation` | bit-exact string match, mode-gated by `policy.GradeVerdict` | `gradeCaseVerdict` (Task 8, `case.go:452`) |
| `Action` | bit-exact string match, same gate | `gradeCaseVerdict` (Task 8) |
| `EchoedDate` | exact string equality against the case's own `eval_date` | `gradeCase` (`case.go:408`) |
| `ConfigHash` | **not graded**: no Go analog computes a config hash | recorded non-grade |
| `Weights` | `CheckCaseWeights` against `tun.DistanceWeights` | `gradeCase` (`case.go:411`) |
| `Thresholds` | `CheckCaseThresholds` against `tun` | `gradeCase` (`case.go:414`) |
| `Candidates` | identity, distance (epsilon or bit-identical per policy), penalty view, per-track vector | `gradeCandidates` (`case.go:706`) |
| `Assignment` | item-to-track pairs against `AssignItems`'s solve, gated by `policy.GradeAssignment` | `gradeAssignment` (`case.go:783`) |
| `ExtraItems` | exact index-list match | `gradeAssignment` |
| `ExtraTracks` | exact index-list match | `gradeAssignment` |
| `AppliedFields` | **not graded**: no Go analog computes post-apply in-memory item fields | recorded non-grade |

### The declared corpus-coverage gaps

**Eleven new gaps, taking rung 3's running total to twenty-one** (3a's
three, 3b's six, 3c's one, this pass's eleven). Rung 4 and rung 15
should read this total rather than re-count it; the plan's own prose
states "ten new" in one place but enumerates eleven items and the
running-total arithmetic (10 + 11 = 21) requires eleven, so eleven is
what this entry records.

1. `import.search_ids` and the explicit-ID branch (0 of 215).
2. `match_by_id`'s no-id branch (0 of 215).
3. `match_by_id`'s no-consensus branch (0 of 215).
4. `import.timid`'s gate on the ID-match early return (0 of 215).
5. The `max_rec` downgrade read against a key added at 0.0 (0 of 294
   candidates carry `missing_tracks`/`unmatched_tracks` at exactly
   0.0, though 272 carry some key at 0.0).
6. `_recommendation`'s per-track key union in the downgrade set: the
   loop runs and never changes an outcome under the shipped `max_rec`.
7. `import.quiet = false` and the whole `timid`/`none_rec_action`/"ask"
   half of `_summary_judgment` (the shim forces quiet).
8. `quiet_fallback: asis` and the `ASIS` action entirely.
9. Every non-default value of the four recommendation knobs: all 215
   cases echo the shipped defaults, so the tier 1 cross-check proves
   the transcription, never the behavior.
10. The assignment objective's `inclArtist=false` on a VA release: no
    tied or assignment-sensitive VA case exists.
11. The recommendation ladder's strict-versus-inclusive asymmetry: no
    committed candidate sits exactly on a threshold (a sweep over all
    215 cases finds none at 0.04 or 0.25); the hand-authored ladder
    rows are the sole detector.

Three items are explicitly not counted. `match.required` and
`match.ignored` are 3a's gaps 1 and 2 already; 3d builds their
consumers and a fixture for each, which closes them by unit test rather
than adding new gaps. "`strong` reached after a text search" is a
corpus fact, not a code path this pass builds. `profile_hashes`
covering one profile of three is a corpus-lane item open since 2b,
carried below rather than counted here.

**Two surfaces are non-ports, not gaps.** `bug-assignment-rejection-gap`'s
fix needs a per-pair rejection mechanism beets does not have.
`bug-nonlatin-alias-selection`'s fix needs a corrected applied artist
string, which rung 3 computes none of. Both join the improvements
register.

### The mutation table, measured counts

Several plan-stated mutation counts were corrected against the real
measurement.

| Task | Mutation | Plan's figure | Measured |
| --- | --- | --- | --- |
| 3 | swap gate 1 (zero-track) and gate 2 (duplicate) | fires | **unobservable**: both gates return `(false, nil)` with no side effect, for any input |
| 3 | `slices.SortFunc` for `slices.SortStableFunc`, 3-candidate tie | fires | **does not fire**: pdqsort takes a zero-swap path on an all-equal-key slice at any length |
| 3 | same mutation, 13-candidate alternating-distance fixture | (not in plan) | **fires**, resting on Go's 12-element insertion-sort threshold |
| 7 | pairs in item order | 23 cases | **16 cases** (uniform per-track distances hide the permutation on the `adv-assign-*` family) |
| 7 | insertion order rather than distance order | 21 cases | **18 cases**, 60 candidate-identity mismatches |
| 7 | truncate the search fan-out | (unspecified) | **48 cases fail with `Unused()` in place, 22 without** |

### Evidence

The three earlier tiers are unchanged. Tier 0 still replays 72/72
bit-identical, translation still grades 50/50, and the request tier
still runs 215/303/48 at zero misses. Tier 1
(`GradeCases`, `ModeCompat`): 215 cases, 294 candidates, 294 album
distances, 294 track vectors, 214 assignments compared, 55 near misses
(1-4 ULP, max relative drift `4.61e-16`), zero failing. Tier 2:
`VerdictFloor` 215; `ModeDefault`'s ten `defect_free` rows reproduce
bit-identically; `ModeDefault` moved zero non-probe verdicts (see
`docs/superpowers/research/2026-09-05-rung-3d-default-mode-verdict-moves.md`);
eight probe cases `Ungraded` (`ProbeCaseFloor` 20, minus the ten
`defect_free` rows, minus the two `ProbeCaseOracleCorrectRows`);
`TiesExcluded` 0 against its ceiling. The corpus totals were
independently recomputed off the tree and matched every floor (215
cases, 294 candidates with 272 carrying a 0.0-valued key, 23 non-identity
assignments, 21 multi-candidate cases, 7/14 non-empty extras).

Measured wall times: tier 1 sweep 60.3s (70.3s package) at `-count=1`;
tier 2's combined two-mode sweep 113.5s (`ModeCompat` 60.3s shared via
`sync.OnceValues`, `ModeDefault` 53.2s); `internal/harness/conform`
123.7s; `go test ./...` 2m06s; `bash scripts/check.sh` about 3.5
minutes.

### The rung 3 to rung 4 handoff

What rung 4's spec needs from 3d, ruling B option (b): rung 4 deploys
at functional completeness with every divergence stamped with the
matcher commit.

1. **The entry point.** `tag.Album(ctx, src *musicbrainz.Client, items
   []matcher.ItemFields, tun matcher.Tunables, compat matcher.Compat,
   evalDate time.Time, opts tag.Options) (*tag.Proposal, error)`, in
   `internal/engine/tag`. `ctx` first, `evalDate` a parameter and never
   a clock read, `opts.SearchIDs` the only optional input. `Proposal`
   carries `CurArtist`, `CurAlbum`, `Candidates` (ascending distance,
   stable), `Recommendation` and `Action`.
2. **The divergence stamp.** Every shadow divergence row rung 4 writes
   carries the matcher commit, so a divergence from a superseded
   matcher revision is re-evaluated mechanically rather than consuming
   owner adjudication. It joins `analysis.AdjudicationEntry` beside the
   existing fields; the shadow's subject shape (`{album, release_mbid}`,
   `adjudication.go:35`) is already modeled and carries zero committed
   rows, so rung 4 is its first producer.
3. **The shadow's budget class.** `musicbrainz.NewClient` requires a
   `Budget` and every request `tag.Album` issues spends one token, so
   rung 4 caps its own MusicBrainz spend by the budget it builds the
   client with. The shadow's class is `ClassBackground`, deferred
   behind rung 7's user-present probe, and the wire pace is
   `shadow.DefaultInterval = 2 * musicbrainz.DefaultInterval`
   (corrected 2026-09-07 at the 4a-ii close; this handoff recorded the
   limiter's own `DefaultInterval = time.Second`, 3c contract C20,
   which the conductor's pace ruling at rung 4's spec fold doubles for
   the shadow's `ClassBackground` caller). Rung 4 branches
   on `musicbrainz.ErrUnavailable` to defer rather than record a
   divergence, and on `tag.ErrNoCandidates` to record "no match".
4. **The pending-divergence file format.** A row a human commits into
   `corpus/analysis/adjudications.json` carries `id`, `subject.case_id`
   (or, for the shadow, `subject.album` plus `subject.release_mbid`),
   `suspected`, `citation`, `release_truth`,
   `disposition: pending-adjudication`, `reason`, `adjudicator: agent`
   and `date`, all five prose fields non-empty. The runner prints
   exactly that row and writes nothing; rung 4's shadow inherits the
   same rule, and the launch gate refuses any pending row.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, unchanged, travels forward from every
   rung 3 pass close (quoted in 3c's entry above).
2. **The four-knob wall registration** stays the named follow-on task
   for the first rung 3 pass after rung 1b lands; `AdminEmail`'s
   interim User-Agent contact sits beside it, both unmoved this pass.
3. **`ties.json` measures `Candidates[0]` only**, so 80 of the 294
   committed candidate assignments were never asked the tie question,
   and **17 of them are tied with no disposition** (re-measured by
   Task 7's own sweep, the first run able to solve every candidate),
   across `real-noise-dylan-bootleg`, `real-noise-pink-floyd-de1994`,
   `real-nonlatin-girls-generation`, `real-nonlatin-utada-first-love`,
   `real-nonlatin-utada-first-love-locale`, and the non-best candidates
   of `should-ask` and `should-reject`. Five change tier 1's graded
   `Tracks` vector on an alternative optimum and two also change the
   `tracks` penalty sum. **Filed for lane E1** as a derivation ask;
   tier 1 grades all 294 vectors meanwhile.
4. **The `lapjv` extension model is behavioral, not source.**
   `n = max(rows, cols)` with zero fill, established against `lap`
   0.5.13 on 2634 random rectangular matrices and reconfirmed on 262 at
   this plan's drafting; the classic `(rows+cols, max+1)` recipe
   `derive_ties.py`'s own docstring assumes disagrees on 30% of
   rectangular ties. Frozen as
   `internal/engine/matcher/testdata/lapjv-optima.json` with its
   provenance header. The day `corpus/manifest.json`'s `lap` pin moves,
   that table must be regenerated, and the oracle-tagged differential
   test this pass declined becomes worth building.
5. **`MergePseudoRelease`'s aliasing hazard is now rung 9's alone.**
   `AlbumForID` closes it by construction because only the merged value
   escapes, and `tag.Album` inherits that same call site unchanged. The
   `slices.Clone` is owed the day a caller retains both values, which is
   rung 9's pipeline.
6. **`vaLikely` stays unexported, permanently.** `match_by_id`'s
   no-consensus branch is corpus-unreachable and the search path
   computes `vaLikely` inside `AlbumSearchFor`, so no caller outside
   `musicbrainz` ever needs it. 3c's carry 13c is closed as "stays
   unexported", recorded so a later pass does not export it for
   symmetry.
7. **The retry path is ungraded by the corpus and always will be**: the
   stub answers 200 or 404 only, `cassette.Request` carries no status
   field, and the response-status extension stays declined, now for the
   fourth pass running.
8. **Three regime divergences travel together**, unchanged from 3c:
   3b's `\b` in `SDPatternsWordBoundary` and now in
   `SDPatternsGuardedNoBracketDiscount`, the query path's
   full-versus-simple case lowering, and `_extract_id`'s `\w`. All three
   are Unicode-aware in Python and ASCII-only in Go RE2, and all three
   are benign on the committed corpus.
9. **`Case.Extra` still has no producer** (3c guard 5), zero of 215
   cases carry an entry, and `Unused()` deliberately excludes `Extra`
   because a strong ID match legally orphans search captures into it.
10. **`profile_hashes` covers `locale` alone** (open since 2b) and the
    response-status extension stays declined (open since 2c). Neither
    moved this pass.
11. **The measured wall times**, listed in Evidence above, join the
    version-bump differential doc the same courtesy every rung 3 pass
    close has passed forward.
12. **The tier 1/tier 2 sweep cost** (60.3s/53.2s, conform 123.7s,
    `go test ./...` 2m06s) is carried, with its CPU profile (regexp
    backtracking under `sdStrip` 43% cumulative, `matcher.Levenshtein`
    17.5%, `Transliterate` 7.7%), as the first rung 3 follow-on or
    rung 9's, and `scripts/check.sh` gains no separate leg for it this
    pass (spec `:267` stays the future escape).
13. `VerdictFrom` is gated on mode; a mode-blind read fired on 18 real
    cases in an intermediate build and was corrected before Task 8's
    close.
14. `nearMissTolerance` (1e-12 relative) is about 4500 ULP against the
    4 ULP maximum drift this pass measured; the tolerance is spec-
    mandated (`:141`), not a hole.
15. `cassette`'s two `Request` rows sharing `(Method, PathQuery)` are
    not rejected, and `Unused()` would mark both hit on one issue;
    zero committed cases are affected, pre-existing from 3c.
16. The `ties skipped` policy arm (`case.go:585-590`) reads wider than
    the spec row's own sentence (it also sets `ExpectNoCandidate`);
    harmless, cannot ungrade anything real.
17. Two of the three `import.*` overlay assertions in
    `translate_test.go` are vacuous today (`QuietFallback` and `Timid`
    equal their shipped defaults); only `NoneRecAction` discriminates.
18. The plan's fixture-count wording ("four zero-instance verdict
    rows") is never enumerated anywhere in the plan; ruled a wording
    error, not an owed deliverable (Task 8's escalation).
19. The 17 unledgered tied candidate assignments (item 3 above) are
    filed for lane E1 as a `derive_ties.py` extension ask (R19).
20. **Owner ask, batched rather than raised** (Resolution R1):
    re-dispositioning `bug-assignment-rejection-gap` from `oracle-wrong`
    to `no-correct-answer`, whose semantics the spec's own grading
    table already describes as exactly the behavior this pass
    implements by hand. `disposition` is owner-only and this pass
    writes none.

### Filed for other owners

- **The oracle-tagged `lapjv` differential test**, for the lane that
  owns `internal/harness/oracle`, to be built the day the `lap` pin
  moves (item 4 above).
- **Extending `derive_ties.py` to ask the tie question of every
  candidate rather than `Candidates[0]`**, for lane E1 (item 3 above).

### Non-blocking leftovers, condensed

- `internal/engine/matcher/assign.go`'s dual-update fix left one
  comment reading as though the augmenting path's ending column carries
  information the running minimum does not; harmless, but worth a
  clearer sentence the next time that function is touched.
- `internal/harness/analysis/ties.go:140-145`: a unique (untied) row
  carrying a stray `Adjudicator` with no `Disposition` is accepted
  silently, unlike the symmetric skipped-row check; not required by the
  criteria.
- `internal/harness/conform/case.go:495`: `compareFloat` has no direct
  unit test; its epsilon path rides on the corpus's 55 near misses
  alone, and the flagged-row bit-identity refusal is proven by mutation
  only, not by an isolated table test.
- `internal/harness/conform/case.go:602`: `gradeCaseDefault` sets
  `res.graded = true`, a field `GradeCases` never reads on the
  `ModeDefault` path; dead but harmless.

### Ruled

**Revision 2.1's three conductor rulings were all taken as
recommended.**

1. B2, `ModeDefault` re-scores each case's own committed candidate set
   under the corrected table and issues no search; the request-set rule
   stays `ModeCompat`-only.
2. m16, the entry point keeps the package name `internal/engine/tag`,
   with both doc comments naming the `tagio` pairing, and the build
   spec's tree listing gained the distinguishing clause this pass
   drafted.
3. m15, the `docs/` ledgers row is written into the parallel-approach
   spec's contested-surface table.

**Every execution-time ruling, in dispatch order.** Task 1's escalation
(fixture location): Task 2 owns `testdata/assign/`; the inline
Go-literal matrices in Task 1 stand. Task 3's fix round (twice): both
findings accepted and fixed, no further escalation. Task 4's escalation
(the `max_rec` key-resolution order): option (a), restore beets' lazy
walk, fixture key `"nosuchpenalty"` to `"album"`. Task 7's escalation
(verdict fields, `AssignmentFloor`, sweep cost): verdict grading is
Task 8's; `AssignmentFloor = 214` with an empty `TiesAwaitingRuling` is
accepted as the measured truth over the plan's stated 212; the sweep
cost is carried as a finding. Task 8's escalation (fixture count,
sweep cost): the plan's "four zero-instance rows" is a wording error,
two fixtures stand; the combined sweep stays inside `go test ./...`.

### Carried

- **To 3d, now resolved.** `Preferred.Media`/`Preferred.Countries`
  config-load validation is **CLOSED** (Task 3): `Tunables.Validate`
  compiles both lists and returns a wrapped `ErrInvalidTunable`. The
  compat switch's pattern-table selection is **CLOSED** (Task 4,
  Resolution R2): both hypotheses default off, proven by the fourth
  `SDPatternsGuardedNoBracketDiscount` table and `ModeDefault`'s ten
  `defect_free` rows. Tie mirroring needing both `inclArtist` call
  shapes is **CLOSED** (Tasks 1-2). 3c's R1 seam (the ID-match early
  return, `_recommendation`, `_add_candidate`, the explicit
  `search_ids` branch) is **closed for the first three, declared as
  gaps 1-3 for the last two**. R6's stub blind side is **CLOSED**
  (Task 7, `stub.Unused()`, a hard failure in tier 1). `vaLikely`'s
  export question is **CLOSED as "stays unexported"** (item 6 above).
  `MergePseudoRelease`'s aliasing hazard is **carried to rung 9** (item
  5 above). R7's `Case.Extra` producer, R11's `AdminEmail` interim
  contact, the four-knob wall registration, the retry path, and
  `profile_hashes` are all **carried unchanged** (items 2, 7, 9-10
  above).
- Five 3b `code-simplifier` leftovers still sit in HISTORY, none
  blocking, none touched this pass.
- Adjudication batch 1 stays fully ruled, unaffected by this pass; no
  new adjudication row was produced (ruling F never fired).

Budgets: prep 2.51M, Task 0 0.09M, run A 0.32M, run B 1.34M, fix round
0.07M, runner review 0.14M, run C 0.32M, simplify 0.15M, simplify
review about 0.07M, close ≈0.15M (estimate), summing to about 5.15M
against the 6M ceiling re-derived after Task 2 to an expected 4.8M and
not raised. Attended interaction points: planning misses 2 (Task 1's
`testdata/assign/` fixture location and Task 8's "four zero-instance
rows" fixture count, both ambiguities a planning question would have
caught); execution sittings 0, every fork ruled by the conductor under
the 2026-09-05 grant, ruling F and the pending-divergence path never
firing.

## Rung 3, pass 3c (The client, the limiter, and request-construction grading): `internal/engine/musicbrainz`'s `query.go`, `client.go` and `limiter.go` born, 351 requests graded byte for byte (2026-09-05)

Eight plan tasks executed sequentially in the main tree against
`docs/superpowers/plans/2026-09-04-rung-3c-client-limiter.md`
(revision 2.2). Commits in order: `e9364ff` plus `e90bdad` (Task 1),
`8e7dc21` (Task 2), `f3c3edf` plus `208c5af` (Task 3), `4e7bb04`
(Task 4), `807dc45` (Task 5, escalated), `f088b1a` (the conductor's
plan revision 2.2 ruling), `dca3d17` plus `18ea5da` (Task 6), `18a1292`
(Task 7), `78a6e5a` plus `0082ae6` (Task 5b, the escalation's fix, two
commits closing the fourth seam), `7809558` (the close-out
`code-simplifier` pass, 9 files, +47/-38), and this entry (Task 8).

What landed. Task 1 reshaped `cassette.Case.Extra` from `[]string` to
`[]Request` in `internal/harness/cassette/cassette.go`. The four
per-entry checks the Requests loop already applies now cover it, plus
three exclusivity rules keyed through a new unexported `requestKey`.
The task bumped `wantManifestSchemaVersion` and
`corpus/manifest.json`'s `schema_version` from `"3"` to `"4"`, and
added `TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema`,
proving the migration empty over all 215 committed cases. Task 2 gave
`internal/harness/stub` a second replay map: `extra
map[key]extraEntry`, consulted after the request map. `Consumed()
[]cassette.Request` reports distinct from `Misses()`, and the
derivation-inertness rule is stated on both doc comments (zero of 215
cases carry an `Extra` key today). Task 3 added `ItemFields.Comp` and
`boolField`, and exported `IsVariousArtistsName`. It closed 3a's
`dataSourcePenalty` carry with `defaultDataSourceMismatchPenalty =
0.5`, replacing the tunable fallback and removing both `conform`
compensations. It also corrected the
`SearchLimit`/`DataSourceMismatchPenalty`/`DataSourcePenalties`/
`AddDataSource` doc comments. Task 4 built `query.go`: the sixteen-
element `releaseIncludes`, unexported `vaLikely`, and `albumCriteria`.
`albumCriteria` walks `tun.ExtraTags` with the
`tracks`/`alias`/`catalognum`/`date:(0)` value rules (the tenth
declared corpus-coverage gap). The rendering layer (`escapeLucene`,
`formatSearchTerm`, `lowerSearchTerm`, `trimSearchTerm`) reproduces
beets' double lower-and-strip and its escape-before-wrap order. Task 5
built `limiter.go`: the three-class dispatcher with
`reservationInterval = 4` and a round-robin reserved token over the
non-empty classes below the one served. It also added `Budget`'s
reserve-then-release cap, `Unmetered()`, and `Close()`.
`NewLimiter(interval time.Duration)` carries `DefaultInterval =
time.Second` and C20's citation chain. Task 5b closed the escalation
(below). Task 6 built `client.go`: `ClientConfig` and `NewClient`,
requiring a `UserAgent` and a `Budget`. `Release` and `SearchReleases`
force `fmt=json` through `url.Values.Encode`, backed by
`extractReleaseID`. The seven-attempt retry transport uses urllib3's
backoff schedule and both `Retry-After` forms, with a per-attempt
limiter token. `AlbumForID`'s pseudo double fetch and merge closes 3b's
`MergePseudoRelease` aliasing carry for this pass's own call site.
Task 7 added `internal/harness/conform/request.go`. `GradeRequests`
opens one stub per case. It issues one `client.Release` per distinct
logged release mbid, and one `client.SearchReleases` when the case
logs a search. It joins every `srv.Misses()` entry into the case's
error. It asserts the three corpus floors (215 cases, 303 lookups, 48
searches) plus a per-case issued-count check that a silent skip cannot
pass.

What the gate caught. `bash scripts/check.sh` stayed green at every
task end. Three tasks took a fix round from the in-chain
`diff-reviewer`, and Task 5 escalated past a fix into a plan
correction. Task 1's reviewer found only five of the plan's six named
Extra-loop failure proofs landed, with the method-non-empty and
response-non-empty per-entry checks unfired by any test. The fix
commit `e90bdad` added both, each verified by a targeted mutation.
Task 3's reviewer blocked because the conductor's ruling (fold M-a)
required `DefaultTunables` to build `DataSourcePenalties["MusicBrainz"]`
from the `DataSourceMismatchPenalty` field. The first attempt instead
shared one local variable between both, leaving the field with zero
production readers despite its doc comment's claim. `208c5af` rewrote
`DefaultTunables` to build the map from the field directly, so a
mutation of the field's literal now provably moves the map too. Task
6's reviewer blocked on an undelivered behavior-contract bullet: the
deliberate divergence on an absent release `status` key (a plain
non-pseudo release here, a dropped candidate via beets' `KeyError` in
`beetsplug/musicbrainz.py:785`) was nowhere in the tree. `18ea5da`
recorded it as a comment at the `PseudoReleaseTarget` call site with
the 303-of-303 corpus measurement that makes it unreachable today, and
a matching doc comment on `Release.Status`. Task 5 escalated rather
than blocked. The plan's cancelled-waiter test cancelled and joined
the departing waiter before feeding a tick, so `depart()` had already
dequeued it. `serve`'s refused-offer arms (`case <-w.gone:`, `case
<-l.done:`) carried zero coverage. That left unproven the
distinguishing half of the chosen cancellation shape, which tries the
next waiter rather than parking the token. The conductor ruled option
(a) at plan revision 2.2: a fourth unexported hook, `holdBeforeOffer`.
It lets a test block `serve` between `pick()` and the offer, cancel
the picked waiter there, and assert the next waiter takes that same
tick's token. Task 5b landed the fix in two commits: `78a6e5a` (the
in-flight refusal, both mutations of the refusal path killed) and
`0082ae6` (the shutdown arm, `case <-l.done:`). That arm had never
been exercised either; deleting it hangs the test to its timeout. The
seam count is **four** (`newLimiter`, `awaitQueued`, `awaitTicks`,
`holdBeforeOffer`), not the plan's originally named two. A final
`diff-reviewer` (opus) read `7809558` and `0082ae6` together against a
fresh checkout and returned accept, confirming the seam count, the
`-race -count=20` limiter replay, and both new assertions' mutation
proofs.

### The request coverage table

Twenty contracts, C1 through C20, each with its Go site and what
grades it. A committed case grades a contract when `GradeRequests`
(Task 7) issues a real request through it and compares the resulting
bytes against the case's own log; a rule no corpus case can reach is
graded by a hand-authored fixture in the owning package's own
`testdata`, with the mutation that proves the fixture fires.

| Contract | Go site | Graded by | Corpus reach |
| --- | --- | --- | --- |
| C1 `fmt=json` forced, no exception | `client.go` `releaseRequestURL`/`searchRequestURL` via `url.Values.Encode` | all 351 requests, Task 7's sweep | Yes |
| C2 API root `<base>/ws/2`, three config keys ignored | `client.go` `wsPathPrefix`, `ClientConfig.BaseURL` | the injected `BaseURL` is the only root the Go client models; the three ignored keys (`https`, `ratelimit`, `ratelimit_interval`) have no Go analog to grade, recorded as a non-membership | n/a, structural |
| C3 sixteen-element `RELEASE_INCLUDES`, one `inc=` value | `query.go` `releaseIncludes` | all 303 lookups, one distinct `inc=`; `TestReleaseIncludesJoinsAllSixteen` (hand literal) plus mutation 1 (drop `genres`, 303 misses measured) | Yes |
| C4 a search carries no `inc` | `client.go` `searchRequestURL` | all 48 searches carry none | Yes |
| C5 the nineteen-character Lucene escape set | `query.go` `escapeLucene` | `TestEscapeLucene`'s 30-row table; mutation 2 (drop `-`) fires 6 cases | Thin: `%27`/`%2C` unescaped and `%5C` escaped both appear in the corpus, but most of the nineteen ride on the fixture table alone |
| C6 lower, strip, drop-if-empty, escape, then wrap | `query.go` `formatSearchTerm` | end-to-end query rows verified against committed `path_query`; mutation 5 (escape after wrap) fires both wrapper rows | Yes |
| C7 space-joined terms, empty terms filtered | `query.go` `AlbumSearchFor` | `TestAlbumSearchForDropsEmptyTerms` (hand fixture, the plan's extra row) | Thin: no committed case has an empty album name |
| C8 insertion-ordered criteria, eight-key `FIELDS_TO_MB_KEYS`, three value rules | `query.go` `albumCriteria`, `fieldsToMBKeys` | `extra_tags.json` (the tenth declared gap); the corpus-reachable `release`/`artist`/`arid` rows | Partial: the base three criteria are reached, the eight extra-tag keys are not (declared gap) |
| C9 the pre-filter's second lower-and-strip | `query.go` (folded into `formatSearchTerm`'s own lower/strip; C9's dict comprehension has no separate Go site since the tuple's query-text half is always `""`) | the double-lower-and-strip end-to-end rows | Yes, jointly with C6 |
| C10 `va_likely`'s three-way disjunction, the `search_artist` override | `query.go` `vaLikely` | mutation 6 (drop the `comp` disjunct) fires the `comp` row and the arid row | Yes |
| C11 `search_limit` 5, legacy spelling `musicbrainz.match.searchlimit` | `query.go` (`AlbumSearch.Limit = tun.SearchLimit`); `tunables.go`'s corrected doc comment | all 48 searches carry `limit=5` | Yes, pins the default only |
| C12 the transport policy (retry count, backoff schedule, timeout, `Retry-After`) | `client.go` `doWithRetry`, `retryWait`, `backoff` | `TestBackoffSchedule`'s six-row table; `TestRetryAfterDeltaSecondsHonored`; mutations 5-6b | Yes, through hand-authored fixtures; no committed case's stub answers a non-200/404 status |
| C13 fetch, translate, conditional second fetch | `client.go` `AlbumForID`, `PseudoReleaseTarget` | mutation 9 (skip the pseudo second fetch, `Country` assertion fails) | Yes, the 3 pseudo cases |
| C14 per-id fan-out follows search-response order | `client.go` `SearchReleases` (`[]string` in response order) | `TestSearchReleasesDecodesTheObjectShapeAndPreservesOrder` | Yes, implicit in every multi-hit search |
| C15 a 404 drops a lookup candidate silently, propagates from search as an empty list | `client.go` `Release`/`SearchReleases` mapping to `ErrNotFound` | the zero-miss rule (Task 7); no test asserts the 404-to-`ErrNotFound` mapping directly (recorded gap, not a plan mutation) | No committed case's stub answers 404 |
| C16 `search_query_ascii` is a structural no-op | `doc.go`'s stated non-check | nothing to grade; the absence is the contract | n/a |
| C17 the corpus's request census (215/303/48/167) | `internal/harness/conform/request.go` `RequestCaseFloor`/`RequestLookupFloor`/`RequestSearchFloor` | `TestRequestConstructionGradesCleanly`, exactly matching the plan's measured figures | Yes |
| C18 `net/url` reproduces the canonical form byte for byte | `client.go`'s `url.Values`-built URLs | all 351 requests' `path_query` compared byte for byte | Yes |
| C19 a search response is an object, ids under `releases` | `client.go` `SearchReleases`'s decode struct | all 48 searches; mutation 4 (bare-array decode) fails all 48 | Yes |
| C20 the wire pace is 1/s, not 4/s | `limiter.go` `DefaultInterval = time.Second` | `TestNewLimiterPacesFromItsOwnTicker`; no corpus case measures wall-clock pacing (Unmetered budget in Task 7's sweep) | n/a, a live-spender fact rather than a byte comparison |

**The one new declared corpus-coverage gap: `albumCriteria`'s
`extra_tags` branch**, including `tracks`, `alias`, the `catalognum`
space-strip and the `date:(0)` trap on an unset year, unreachable
because no committed profile writes a `musicbrainz:` block setting
`extra_tags`. **Together with 3a's three and 3b's six, rung 3 now
carries ten declared gaps**; 3d's close should read that running total
rather than re-count it.

**Four surfaces are non-ports, recorded with their reasons rather than
as gaps.**

1. The recording lookup and `RECORDING_INCLUDES`. This pass ships only
   the release lookup and release search beets' album path uses.
2. The browse fallback (`_ensure_complete_recordings`). Carried
   forward from 3b, still unported.
3. Beets' explicit `search_artist`/`search_name` override on a
   user-forced search. No committed profile sets it, and the plan
   scopes 3c to the automatic album-match path.
4. urllib3's 413-plus-`Retry-After` sixth retryable case (`fold m3`).
   Unreachable from MusicBrainz, which never answers 413.

### Evidence

The request tier, `TestRequestConstructionGradesCleanly`: **215 cases
graded, 303 lookups, 48 searches, 167 cases with no search**, zero
misses, in 3.34s at `-count=1` (measured at close; the task chain's own
figure was 3.32-3.54s across runs). `Consumed()` accumulates zero
entries across the sweep, matching C17's zero committed `Extra` keys.
The schema bump's round-trip proof,
`TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema`, ran 1.870s
at `-count=1` over all 215 cases at close (2.050s to 2.662s across the
task chain), confirming no committed byte moved. Tier 0:
`TestComponentTableReplaysAllRows` still 72/72 rows, bit-identical, in
0.015s at close, per-method counts unchanged after the
`dataSourcePenalty` fix and the removal of both `conform`
compensations (add 8, add_data_source 4, add_equality 3, add_expr 2,
add_number 4, add_priority 4, add_ratio 4, add_string 3,
generic_penalty_keys 1, string_dist 23, track_distance 16). `go test
./...` ran 15.081s wall at close (`internal/harness` 11.255s and
`internal/harness/conform` 11.230s dominating, tier 0 and the request
tier both running inside the latter). `bash scripts/check.sh` ran
1m25.932s wall at close, gitleaks 1m4.3s of that scanning ~114.4 MB,
consistent with the task chain's 1m6.7s-3m00.7s range across tasks.
The limiter's `-race` run: `go test -race -count=1
./internal/engine/musicbrainz/...` 2.090s at close (1.162s measured
under Task 5 alone); the pass-end reviewer additionally ran the ten
limiter tests under `-race -count=20` with zero failures, and the
Task 5b review ran `-race -count=1` over both
`internal/engine/musicbrainz` and `internal/harness` (101s and 123s)
plus a clean `golangci-lint` over `internal/...`.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, verbatim from the plan's Spec
   section, travels forward from every rung 3 pass close: "Rung 15
   re-runs `internal/harness/conform`'s three tiers over the complete
   corpus, at the matcher commit as landed, and its acceptance is the
   launch bar's matcher clause read against that run: zero false
   confident matches over every stratum including stock-config, and
   zero unadjudicated beets-right/Go-wrong cases. Rung 3's own
   acceptance is provisional against the corpus at its then-current
   size. Any tier-2 divergence carrying a pending row at rung 15 is the
   launch gate's business, not rung 15's to wave through."
2. **The four-knob wall registration** stays the named follow-on task
   for the first rung 3 pass that runs after rung 1b closes:
   `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, `SearchLimit`,
   `internal/config/keys.go` the file no lane edits until then, and
   `Threshold.Default` is a string with `ParseInt` the only parser, so
   the three float rec thresholds need a float variant the registry
   does not yet have.
3. **3c grades request construction, not the request set.** The seam
   is R1's: 3d owns the ID-match early return, `_recommendation`,
   `_add_candidate`, the explicit `search_ids` branch, and
   `match_by_id`'s no-id/no-consensus branch, and inherits the stub,
   the zero-miss rule, and the client unchanged.
4. **The stub's blind side and its four-line close** (R6). Nothing
   reports which logged `Request` entries a run consumed, so a client
   issuing fewer requests than the log records leaves zero misses.
   3c's per-case issued-count assertion covers it inside 3c; 3d needs
   the hit set.
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
7c. **The `AdminEmail` User-Agent contact is interim** (R11).
    `Config.AdminEmail`'s own doc comment says adopters compare it and
    never display it, and the harness precedent this plan cites uses a
    dedicated `DUBPLATE_HARNESS_CONTACT` input instead. A dedicated
    contact row is the post-1b follow-on, beside the four-knob
    registration.
8. **The retry path is ungraded by the corpus and always will be**:
   the stub answers 200 or 404 only, `cassette.Request` carries no
   status field, and the response-status extension stays declined,
   now for the third pass running.
9. **Three regime divergences now travel together** (R14): 3b's `\b`
   (Unicode-aware in Python `re`, ASCII-only in Go RE2), the query
   path's full-case-versus-simple-case lowering with all 56 sweep code
   points reaching the wire, and `_extract_id`'s `\w`. The strip
   divergence is closed rather than carried.
10. **`AlbumForID` closes 3b's `MergePseudoRelease` aliasing carry by
    construction**, and the hazard changes owner to 3d and rung 9
    rather than disappearing; the `slices.Clone` is the fix the day a
    caller retains both.
11. **3a's `dataSourcePenalty` carry is CLOSED** (R15), with both
    `conform` compensations removed and 72/72 tier-0 rows re-verified
    as the proof.
12. **An absent release `status` is a plain non-pseudo release here
    and a dropped candidate in beets**, unreachable in the corpus (303
    of 303 lookups carry one), recorded as a deliberate choice at
    `client.go`'s `PseudoReleaseTarget` call site.
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
14. **The measured wall times**: the request tier 3.34s, the
    round-trip proof 1.87s, tier 0 0.015s, `go test ./...` 15.08s, and
    `scripts/check.sh` 1m25.9s (all at `-count=1`, measured at close),
    which the version-bump differential doc will want, the same
    courtesy 2b passed to 2c, 2c to 2d, 2d to 3a, 3a to 3b, and 3b to
    3c.
15. Rung 9 reads four seams, not the plan's two, when it re-verifies
    the dispatcher against its real callers.
16. **The cancellation mechanism beets never specifies, and what this
    pass chose.** A refused offer (the picked waiter departed between
    `pick()` and the offer) tries the next waiter under the same tick.
    Parking the token in the spare slot would strand already-queued
    waiters, who only re-check the spare at registration.

### Non-blocking leftovers, condensed

- `internal/harness/cassette/cassette_test.go:325` names both forms in
  the mutation-1 message; the assertion checks only "not canonical",
  matching an existing sibling test's style rather than tightening it.
- `internal/harness/stub/stub.go:11` and `cassette.go:214`: two doc
  links that could not resolve, fixed in the simplify pass (`7809558`)
  by naming the package in prose instead of a bracketed link.
- `internal/engine/matcher/info.go:229`: `boolField`'s doc comment
  cites a pre-edit line range for `intField`'s comment; worth a
  conductor call since this task exists partly to correct stale
  citations, not yet fixed.
- `internal/engine/musicbrainz/query_test.go:95,111,130`: three
  rendering-table tests carry no row-count guard, so a dropped or
  renamed testdata key would pass vacuously; a `len` check per table
  would close it.
- `internal/engine/musicbrainz/query_test.go:230`: the tenth-gap table
  asserts through `AlbumSearchFor`'s rendered query rather than the
  `[]criterion` slice the plan's table header names; coverage is equal
  or stronger.
- `internal/engine/musicbrainz/client.go:329,376`: `doWithRetry`'s
  signature deviates from the plan's named shape (`requestURL string`
  in place of `*http.Request`), defensible since the body is read and
  closed in one place; no test covers the 404-to-`ErrNotFound`
  mapping.
- `internal/engine/musicbrainz/client.go:404`: `backoff`'s 120s cap is
  unreachable, since the client only sleeps for attempts 1 through 6
  (max 16s); a correct, dead port of urllib3's schedule.
- `internal/harness/conform/request.go:186-220`: `distinctReleaseLookupMBIDs`
  dedups by mbid while `countLookupRequests` counts raw log rows; a
  future case fetching one mbid twice under two `inc` sets would false-
  fire the per-case assertion. No committed case does this today.
- `internal/harness/conform/request.go:105-115`: `summary.Cases`
  increments even on a load or stub-start error, so a broken case still
  counts toward `RequestCaseFloor`; harmless because the gate test
  fatals on any error first, fixed toward the doc comment's promise in
  the simplify pass.

### Filed for other owners

- **The `/ws/2/release/` literal fold**, for the lane owning
  `mbrequest`: `cmd/harness/capture.go:279` and
  `internal/harness/synth/synth.go:41` each spell
  `"/ws/2/release/"` inline rather than reading it from
  `internal/harness/mbrequest`'s own constant. 3b's own guard 6 records
  the deliberate three-copy split between `oracle`, `analysis`, and
  `conform`'s `releaseLookupPrefix`; this is a fourth copy in a
  different package pair, worth folding the day the lane that owns
  `mbrequest` next touches either file.
- **Whether `-race` joins `make check`**, for whoever owns the
  Makefile. This pass's only concurrent code (`limiter.go`) is proven
  race-clean by hand at every task boundary and by the pass-end review,
  but `scripts/check.sh`'s `go test ./...` does not run under `-race`
  today, so a future concurrency bug in `internal/engine/musicbrainz`
  or a later package would not be caught by the gate itself.

### Ruled

**Revision 2.1's five conductor rulings were all taken as recommended.**

1. B1, the reserved token round-robins across the non-empty classes
   below the served one.
2. M-a, `DataSourceMismatchPenalty` is wired into `DefaultTunables`'
   penalty map (Task 3's fix round proved the wiring is real, not a
   coincidence of two equal literals).
3. M-b, `NewLimiter(interval time.Duration)` with `DefaultInterval =
   time.Second`.
4. m22, `corpus/README.md` and `corpus/manifest.json` both carry
   contested-surface rows, confirmed already landed in the parallel
   spec before this pass launched (no-op at this close).
5. m23, the 80%-of-ceiling flag routed to the conductor for the
   2026-09-04 overnight grant's duration only.

**Revision 2.2's ruling** (Task 5's escalation): option (a), the fourth
unexported hook `holdBeforeOffer`, landed as Task 5b, is recorded above
under "What a later pass would be wrong to rediscover" item 15.

The optional `searchlimit` spec clarification at the rung 3 spec `:226`
is **declined**: that line spells legacy keys relative to the plugin
namespace, as its adjacent `source_weight` item does, so
`match.searchlimit` is right as it stands and the ambiguity the
clarification would remove is not worth the edit. 3a's HISTORY
correction #5 (the `SearchLimit` legacy spelling) was already corrected
at `:600` before this pass's close, folded in at the plan's own drafting
(fold m13); no edit was owed here.

### Carried

- **To 3c, now resolved.** `dataSourcePenalty`'s configured-versus-
  literal fallback is **CLOSED** (R15): `defaultDataSourceMismatchPenalty`
  replaces the tunable read and both `conform` compensations are
  removed, with 72/72 tier-0 rows re-verified bit-identical.
  `MergePseudoRelease`'s shallow track-slice copy is **re-owned to 3d
  and rung 9** rather than closed: `AlbumForID` closes the aliasing
  hazard by construction for this pass's own call site, but the
  `slices.Clone` fix is still owed the day a caller retains both the
  merged and the pseudo candidate (item 10 above). The browse fallback
  is a **non-port** this pass, unowned until a corpus case crosses 500
  tracks (unchanged from 3b).
- **To 3d:** the `Preferred.Media`/`Preferred.Countries` config-load
  validation (3d's alone, unchanged from 3b); the compat switch's
  pattern-table selection and the `\b` regime divergence; the tie-
  mirroring work needing both `inclArtist` call shapes; 3c's own R1
  seam (item 3 above), R6's stub blind side (item 4), R7's `Extra`
  producer (item 5), and R11's interim `AdminEmail` contact (item 7c).
- The four-knob wall registration stays the named follow-on task for
  the first rung 3 pass after rung 1b lands (item 2 above); five 3b
  `code-simplifier` leftovers still sit beside it in HISTORY, none
  blocking.
- `profile_hashes` (m1) and the response-status extension, carried
  since 2b and 2c, are unaffected by this pass and stay open.
- Adjudication batch 1 stays fully ruled, 18 `oracle-wrong` and 3
  `oracle-correct` with zero pending; this pass produced no new
  adjudication row.

Budgets: prep 2.55M (two runs, a usage-limit pause), Task 0 0.11M, the
Tasks 1 to 7 chain 1.74M, Task 5b 0.13M, the simplify pass 0.15M, the
final fresh-context review 0.07M, close ≈0.15M (estimate) against the
6M ceiling, not re-derived after Task 4 (the conductor's Task 4
checkpoint found no cause to raise it). Attended interaction points:
planning misses zero; execution sittings one, a combined five-item
grant exchange on 2026-09-05 that covered the unattended overnight
window rather than any 3c content decision.

## Rung 3, pass 3b (The translation layer, the `AlbumInfo` ledger, and the `defect_free` block): `internal/engine/musicbrainz` born, 50/50 ledger rows graded bit-identical (2026-09-04)

Nine plan tasks plus a staleness pre-task, executed sequentially in the
main tree against
`docs/superpowers/plans/2026-09-04-rung-3b-translation.md` (revision
2.1). Commits in order: `946cc21` (Task 1), `a67d5a0` plus `ae2b56e`
(Task 2), `7fcb5b4` plus `7092090` (Task 3), `b8460e0` (Task 4),
`340a983` (the conductor's Task 4 checkpoint), `251ab14` (Task 5),
`cf5599e` plus `40f78be` (Task 6), `8ba42a5` (Task 7), `0621aa3` (the
conductor's plan revision 2.1), `f4eb389` (Task 8), and `096a881` (the
close-out `code-simplifier` pass, 15 files, +115/-157).

What landed. Task 1 built the derived `AlbumInfo` ledger ruling E asks
for: `internal/harness/analysis/album_info.go` with the nineteen named
subjects, the `AlbumInfoLedgerRowFloor` of 50, the six-part
`CheckAlbumInfoLedger`, and the pointer-per-scalar ledger types;
`derive_album_info.py` in `derive_ties.py`'s shape with its
oracle-tagged Go driver; and `corpus/analysis/album_info.json` at 19
subjects, 50 rows, 4 merged blocks, 262,377 bytes. Task 2 landed the
`defect_free` block on `AdjudicationEntry` with `CheckDefectFree`, the
three within-row rules, `matcher.SDPatternsNoBracketDiscount` as the
third pattern table, `derive_defect_free.py` with its module-object
monkeypatch, and the first population of the ten string-heuristic rows.
Task 3 birthed `internal/engine/musicbrainz` with `doc.go`'s charter,
the WS/2 release vocabulary, `wireInt`'s two-shape position decode,
`Validate`'s boundary rules, and `GroupRelations`; it also added
`matcher.Tunables.VAName`. Task 4 added `preferredAlias`,
`keyWithPreferredAlias`, `parseArtistCredits`, and
`applyVariousArtists`. Task 5 added `tracksFromMedium`, `buildTrack`,
`extractTracks`, and `deriveMedia` with both `"Media"` sentinel routes.
Task 6 added `preferredReleaseEvent`, `parseDate`, and `TranslateAlbum`,
the one album-side composition. Task 7 added `PseudoReleaseTarget` and
`MergePseudoRelease`. Task 8 landed the translation tier in
`internal/harness/conform`: `CheckTranslatedAlbum`'s field-by-field
comparison over `errors.Join`, `gradeLedger` with its compared-row
count, the twenty-one-row nullability policy, the profile-name overlay
with its key-drift guard, the browse-fallback guard, and the clock walk
widened from `internal/engine/matcher` to all of `internal/engine`.

What the gate caught: `bash
scripts/check.sh` stayed green at every task end and found nothing
novel. That is 2c's, 2d's, and 3a's record repeating. Three tasks took
a fix round and one escalated, all four from the in-chain
`diff-reviewer` rather than from the gate. Task 2's reviewer blocked on
an unimplemented acceptance clause: `derive_defect_free.py` solved the
optimal item-to-track mapping under the substituted table and then
discarded it, so the plan's "verify rather than assume the mapping is
unchanged" went unproven; `ae2b56e` emits the mapping as a row field and
asserts it against each case's committed `Expected.Assignment`, and all
ten resolve to `{0:0, 1:1, 2:2}`. Task 3's reviewer blocked because the
four `Validate` failure tests asserted only `errors.Is` and never read
the message, so stripping the index and id from every format string
would have left the suite green against a doc comment promising to name
the offending element; `7092090` added the four substring assertions and
proved them by stripping the identifiers. Task 6's reviewer blocked
because `TranslateAlbum`'s doc comment dropped the plan's error-contract
paragraph, leaving an exported function that returns an error with no
godoc about when it does; `40f78be` restored it. Task 5's reviewer
escalated rather than blocking: the plan's named mutation 4, dropping
the `preferredAlias` guard on the track-title override, is provably
unobservable through `TrackInfo.Title`, because `keyWithPreferredAlias`
resolves the same aliases under the same tunables immediately
afterward. The implementer declared the non-firing instead of
fabricating a proof. The reviewer reproduced it independently. The
conductor ruled the guard stays, its non-effect is stated in `tracks.go`,
and the mutation is struck from the proof set as a declared unobservable
mutation rather than a coverage gap (plan revision 2.1, `0621aa3`); Task
5 was accepted at `251ab14` with no code change.

### Translation coverage table

Sixteen contracts, T1 through T16, each with its Go site and what grades
it. A ledger row grades a contract when one of `album_info.json`'s 50
rows exercises it through `TranslateAlbum`; the rules the corpus cannot
reach are graded by hand-authored fixtures in the owning package's own
`testdata`, per the plan's fixture-proximity rule.

| Contract | Go site | Graded by | Corpus reach |
| --- | --- | --- | --- |
| T1 one album-side function, merge inside | `translate.go` `TranslateAlbum`, `pseudo.go` `MergePseudoRelease` | all 50 ledger rows; the 4 `Merged` blocks | Yes |
| T2 key normalization, one rule surviving | `release.go` struct tags; `alias.go` `keyWithPreferredAlias` | all 50 ledger rows; `alias_test.go`'s table | Yes |
| T3 `mediums` counts every medium | `translate.go` (`len(rel.Media)`), `tracks.go` `extractTracks` | ledger rows; `translate_test.go`'s two-medium release under `IgnoredMedia` | Yes |
| T4 the `"Media"` sentinel, two routes | `tracks.go` `deriveMedia` | 4 zero-track rows (`real-nonlatin-hamasaki-ayumi`), 2 mixed-format rows (`real-noise-dylan-bootleg`); 4 sentinel fixtures | Yes, both routes |
| T5 assembly order and the two filters | `tracks.go` `tracksFromMedium` | `gap_pregap_medium.json`, `gap_data_tracks_medium.json`, `gap_bracket_title_in_tracks.json` | No: pregap and data tracks are declared gaps 1 and 2 |
| T6 the three track-level overrides | `tracks.go` `buildTrack` | all 50 ledger rows' track walks; the truthiness-fallback tests | Yes, except the title guard (declared unobservable) |
| T7 `_preferred_alias` | `alias.go` `preferredAlias` | `preferred_aliases.json`, T7's eight rows | Thin: declared gap 5, one profile and one case |
| T8 credit flattening | `credit.go` `parseArtistCredits` | `artist_credit_alias.json`; ledger rows | Yes |
| T9 the VA override | `credit.go` `applyVariousArtists` | `translate_test.go`'s VA release; ledger rows | Yes |
| T10 release-event selection | `events.go` `preferredReleaseEvent` | 4 preferred-country fixtures | No: declared gap 3, the fallback only |
| T11 date parsing | `events.go` `parseDate` | T11's six-row table, `"1999-XX-06"` proving no early return | Yes, through the fallback path |
| T12 the pseudo merge, seven fields moved | `pseudo.go` `MergePseudoRelease`, `PseudoReleaseTarget` | the 4 `Merged` blocks; 5 `pseudo_target_*.json` fixtures | Partly: declared gap 4 on year and mediums |
| T13 no Unicode normalization | `doc.go` (a stated non-check) | nothing to grade; the absence is the contract | n/a |
| T14 config reads inside the layer | `matcher/tunables.go` (`VAName` and the rest) | `TestDefaultTunables_VAName`; the defaults cross-check | Yes |
| T15 the plugin's cached config reads | `derive_album_info.py`, `derive_defect_free.py` (one plugin built after config, never mutated) | the derivation's own shape asserts | n/a, derivation-side |
| T16 the browse fallback, not ported | `doc.go`'s refusal; `conform`'s `checkNoReleaseReachesBrowseThreshold` | `TestNoCommittedReleaseReachesTheBrowseThreshold` over all 215 cases | Yes, as a guard |

**Six declared corpus-coverage gaps, recorded as gaps rather than as
passing checks** (spec, M8).

1. **Pregap**: zero committed cases carry a non-null `pregap`.
2. **Data tracks**: zero committed cases carry a non-empty
   `data-tracks` array, so both the `IgnoreDataTracks` branch and the
   `"[data track]"` title filter are ungraded by the corpus.
3. **The preferred-country release-event branch**: the branch is
   reachable data-side and unreached only for that reason. Of the 303
   release-lookup responses, 290 of 298 top-level release events carry
   a non-null `area` with `iso-3166-1-codes`, across 208 case
   directories. A fourth profile setting `match.preferred.countries`
   would reach the branch immediately, and `various-artists` is the
   concrete target.
4. **The merge's `year`, `original_year`, and `mediums` path**: equal
   on both sides of all four committed pseudo cases, so the corpus
   gives that path no signal.
5. **The alias path's breadth**: `import.languages` is non-empty in
   exactly one committed profile and one committed case, so 214 of 215
   cases run with `preferredAlias` returning nil unconditionally.
6. **Two boundary rules with no corpus instance**: no committed
   release carries an empty resolved track credit and no committed
   backward `transl-tracklisting` relation lacks its `release` target,
   so both of `Validate`'s rules are graded by hand-authored fixtures
   alone. Each closes a Go panic rather than an error. That is why the
   absence is worth recording.

Together with 3a's three
(`match.required`, `match.ignored`, `preferred.original_year`), **rung 3
now carries nine declared gaps**; 3d's close should read that total
rather than re-count it.

**Task 5's struck mutation is not a gap.** The `preferredAlias` guard on
the track-title override is provably redundant with respect to any
observable `Title`, so no fixture can make its removal fire. The guard
is faithful to beets, it stays, and `tracks.go` documents its
non-effect at the site.

### Evidence

The `AlbumInfo` ledger: 19 subjects, exactly 50 rows, 4 merged blocks,
262,377 bytes, `shim_version` `"1"`, `beets_version` `2.13.1`. Its
derivation ran 25.19s for the re-derive test and 25.26s for the
perturbation test at `-count=1`. Row ids are the deduplicated
`/ws/2/release/<mbid>` lookups of each subject's own case, and no
subject case logs a duplicate lookup.

The translation tier, `TestTranslationLedgerGradesCleanly` against the
real `corpus/analysis/album_info.json`: **50 of 50 rows graded, zero
disagreements, on the first pass**, in 364 to 390ms across runs
(≈7.3 to 7.8ms per row). The compared-row count equals the ledger's row
count, so a silently skipped row fails. 561 of the 650 album-level
scalars in the ledger are non-null, so the grading is not vacuous. All four
`Merged` blocks graded `MergePseudoRelease` as well as the translation.

The `defect_free` block, ten rows, derived float64 beside the prose
figure each row's `corrected_verdict` carries:

| Row | Derived distance | Prose figure | Table |
| --- | --- | --- | --- |
| `adj-adv-ft-draft-colon` | 0.10185185185185186 | 0.101852 | `SDPatternsWordBoundary` |
| `adj-adv-ft-left-colon` | 0.11764705882352941 | 0.117647 | `SDPatternsWordBoundary` |
| `adj-adv-ft-gift-space` | 0.10784313725490197 | 0.107843 | `SDPatternsWordBoundary` |
| `adj-adv-ft-soft-onesided` | 0.08666666666666667 | 0.086667 | `SDPatternsWordBoundary` |
| `adj-adv-ft-craft-titles` | 0.09170751633986929 | 0.091708 | `SDPatternsWordBoundary` |
| `adj-adv-ft-theft-artist` | 0.0909090909090909 | 0.090909 | `SDPatternsWordBoundary` |
| `adj-adv-remix-both-sides` | 0.08 | 0.080000 | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-onesided` | 0.07142857142857142 | 0.071429 | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-bracket` | 0.0925925925925926 | 0.092593 | `SDPatternsNoBracketDiscount` |
| `adj-adv-remix-track-titles` | 0.08187134502923976 | 0.081871 | `SDPatternsNoBracketDiscount` |

The both-ways check ran the same ten subjects twice in one process, once
with the pattern table untouched and once with it substituted. The
baseline leg reproduced every case's committed `case.json` distance
bit-identically. That is what makes the substituted leg's figures
trustworthy. All ten re-solved mappings are `{0:0, 1:1, 2:2}`, matching
each case's committed `Expected.Assignment`, so the substitution moves
distances and never the assignment. The write-back added exactly ten
`defect_free` blocks and changed no pre-existing field on any row.

`make derive` and the byte-identity invariant: the oracle-tagged suite
ran twice in full, at 983.4s under Task 1 and at 17m50.8s under Task 2,
both exit 0, re-deriving `components.json`, `ties.json`,
`file_tags.json`, `album_info.json`, and every `case.json`
byte-identically. The `code-simplifier` reviewer re-ran the
`AlbumInfo|DefectFree` subset at 64.9s with `corpus/` unchanged
afterward.

The browse-fallback guard scans all 215 cases in 2.15 to 2.38s as
reported and ≈3.4s per scan on the reviewer's machine, where it
dominated the `conform` package's 7.7s. It runs twice. Nothing reaches
the threshold: the two `real-boxset-rachmaninov-complete` twins are the
corpus maximum at 470 tracks against `BROWSE_MAXTRACKS` of 500, and
lowering the test-local threshold to 400 fires naming both.

Gate and suite wall times, all at `-count=1`. `bash scripts/check.sh`
ran green at every task end, 57.7s to 1m2.3s across tasks, with gitleaks
self-reporting 52.9s to 1m8.8s over ~113.8 MB and dominating as it did
in 3a. `go test ./...` ran 8.563s and 8.640s on two tasks; the
`internal/...` subset ran 16.6s on a third. Tier-0 replay stayed green
over all 72 rows throughout, since it runs inside `go test ./...`.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, verbatim from the plan's Spec section,
   travels forward from every rung 3 pass close: "Rung 15 re-runs
   `internal/harness/conform`'s three tiers over the complete corpus, at
   the matcher commit as landed, and its acceptance is the launch bar's
   matcher clause read against that run: zero false confident matches
   over every stratum including stock-config, and zero unadjudicated
   beets-right/Go-wrong cases. Rung 3's own acceptance is provisional
   against the corpus at its then-current size. Any tier-2 divergence
   carrying a pending row at rung 15 is the launch gate's business, not
   rung 15's to wave through."
2. **3a's R1 carry is closed, and this is its resolution.** An absent
   source field becomes the plain-string zero value, never a nil-bearing
   vocabulary. The proof is source-scoped: `internal/engine/musicbrainz`
   is the only translation source rung 3 ships, and every field it
   writes comes from a Go decode that yields `""` rather than a null. The
   thing that would reopen the question is a second metadata source, and
   the translation tier's three `nullStrict` fields are what would catch
   it, because a ledger null against a Go zero fails there rather than
   passing silently.
3. **The pointer-per-scalar ledger convention.** `LedgerAlbumInfo`'s and
   `LedgerTrackInfo`'s scalars are `*T` because a JSON null and an empty
   string are different facts about what beets produced, and the
   `null-and-empty` testdata fixture exists to prove the two decode
   differently. The convention is deliberate against the
   go-conventions no-pointer default and its reason is on the type.
4. **The nullability policy is twenty-one rows, three regimes.**
   `conform`'s `fieldNullability` splits the fourteen `AlbumInfo` fields
   (less `Tracks`) plus the eight `TrackInfo` fields into 3 `nullStrict`,
   9 `nullNeverExpected`, and 9 `nullEqualsZero`, one reason per row, and
   `TestNullabilityPolicyCoversEveryField` asserts the count against
   `reflect.TypeFor` so a new vocabulary field cannot slip past. The
   table documents; the comparison passes its policy as a literal at each
   call site. Nothing catches a drift between the two. The limit is
   known and accepted.
5. **`Track.Position` is a two-shape decode and has to be.** The
   MusicBrainz payload spells `position` two ways in this corpus, 6,763
   integers and 297 strings, and it is the only field that does.
   `wireInt` is the narrow answer, proven directly against
   `corpus/cases/media-baseline/responses/001.json`, whose six positions
   are all JSON strings. 3c's live client meets real captures only, where
   the string form does not occur, so the type is carried for the corpus
   rather than for the wire.
6. **The profile overlay is a name-to-tunables table, not a YAML
   parse.** No YAML parser exists in the module and the plan forbids
   adding one, so `conform`'s `mergeProfile` overlays by profile name
   (`stock` and `estate` no-ops, `locale` setting `Languages` to
   `["ja"]`) and fails on any other name. A line-oriented drift guard
   scans `corpus/profiles/*.yaml` against a ten-key allowlist, five keys
   with a `Tunables` home and five without, so a profile that grows a key
   the overlay does not model reds the gate instead of being silently
   ignored.
7. **The monkeypatch target is the module object.**
   `beets.autotag.distance` resolves to a function, so
   `derive_defect_free.py` reaches `SD_PATTERNS` through
   `importlib.import_module` on the module object and restores it in a
   `finally`. `string_dist` is uncached and reads its module global at
   call time. That is why one process can carry all ten subjects with a
   per-subject substitution and no leak. Editing `derive_oracle.py`,
   `derive_component.py`, or the venv would break the byte-identity
   invariant. Do not do it.
8. **The merge moves exactly seven fields.** `MergePseudoRelease` copies
   the pseudo by value and overwrites `Media`, `Mediums`, `Country`,
   `CatalogNum`, `Year`, `OriginalYear`, and `Label` from the actual;
   `Artist`, `Album`, `VA`, `AlbumDisambig`, `AlbumID`, `DataSource`, and
   `Tracks` stay the pseudo's. All fourteen are asserted. The copy is
   shallow, so `merged.Tracks` aliases the pseudo's backing array, and a
   3c caller that mutates a merged track element writes through to the
   pseudo candidate.
9. **The `Preferred.Countries` two-regime hazard.** One config list feeds
   two different rules. `internal/engine/musicbrainz` does plain
   case-sensitive membership; `internal/engine/matcher` compiles an
   anchored case-insensitive regex. Both sites name the divergence, and
   a mutation compiling the membership test case-insensitively fires.
10. **The three `SD_PATTERNS` tables and what each is for.**
    `BeetsSDPatterns` is shipped and defective on purpose;
    `SDPatternsWordBoundary` carries the six `adv-ft-*` figures;
    `SDPatternsNoBracketDiscount` carries the four `adv-remix-*` figures
    and is born here as `BeetsSDPatterns` less its two 0.3-weight bracket
    rows. 3d's compat switch selects one by assignment to
    `Tunables.StringDistPatterns` and edits no component. 3a's warning
    travels with them: the word-boundary guard changes zero of the 72
    tier-0 rows, so tier 0 is blind to a "fixed" table.
11. **The `\b` regime divergence.** `SDPatternsWordBoundary`'s `\b` is
    Unicode-aware in Python's `re` and ASCII-only in Go's RE2, so 3d's
    compat switch and the `defect_free` derivation can differ on
    non-ASCII input. All six `adv-ft-*` pairs are ASCII, so nothing moves
    today. It travels beside 3a's `SD_PATTERNS`-as-a-value seam.
12. **The `defect_free` block as built.** It carries `distance`,
    `recommendation`, `action`, and `pattern_table`; the derivation's own
    unrounded float64 is the stored value and the prose figure is the
    cross-check. `StringHeuristicHypotheses` is the predicate selecting
    which adjudication rows must carry one, and the six
    `adv-assign-*` rows are excluded because their hypothesis is not a
    string-heuristic one. The two legal `pattern_table` values are the
    named tables above. `CheckDefectFree` enforces the categorical
    cross-check both ways, including the forbidden direction: a block on
    a row whose hypothesis is not a string-heuristic one is a failure.
    The numeric check was one-time, at population.
13. **The `AlbumInfo` ledger as built.** It is keyed on
    `(case_id, release_mbid)`, carries nineteen subjects and 50 rows, and
    grades the translated fields only. Applied fields are deliberately
    out of scope, so **rung 8 must not read it as an applied-field
    ledger**; `file_tags.json` is that. Its `Merged` block appears on
    exactly the four `AlbumInfoPseudoSubjects` cases and check 6 reads
    that named list rather than deriving its own expectation. 3d should
    not re-key it.
14. **The browse fallback's refusal and its guard.** `doc.go` states that
    `_ensure_complete_recordings` is not ported;
    `BROWSE_MAXTRACKS` is 500 and the corpus maximum is 470, a 30-track
    headroom on the two rachmaninov twins. `conform`'s guard uses `>=`
    rather than beets' `>`, deliberately, because its purpose is to catch
    a corpus case approaching the line. **3c owns the port** if a case
    ever crosses it.
15. **`Tunables.VAName` landed and `AliasesAsCredits` did not.** The
    exclusion is deliberate per R8 and is stated on `VAName`'s own doc
    comment, so a later pass does not add the second knob for symmetry.
16. **The wire-type census, restated for 3c.** `position` is the only
    two-shape field; see guard 5. Nothing else in the WS/2 vocabulary
    this pass models needs a custom unmarshaller.
17. **Measured wall times for the version-bump differential doc**, the
    courtesy 2b passed to 2c, 2c to 2d, and 2d to 3a: the `AlbumInfo`
    derivation 25.2s, the full oracle suite 983.4s and 17m50.8s on two
    runs, the translation tier 0.37 to 0.39s for 50 rows, the browse scan
    2.2 to 3.4s per pass over 215 cases, `go test ./...` 8.6s, and
    `scripts/check.sh` 57.7s to 1m2.3s with gitleaks the dominant leg.

### Carried

- **To 3c:** `dataSourcePenalty` falls back to
  `Tunables.DataSourceMismatchPenalty` where beets falls back to the
  class literal `0.5`; identical under shipped config, diverging only if
  `source_weight` is set (3a's carry, unchanged). New from this pass:
  `MergePseudoRelease`'s shallow copy aliases the pseudo's track slice,
  so a 3c caller that mutates a merged track writes through; and the
  browse fallback stays unported until a corpus case crosses 500 tracks.
- **To 3d:** the `Preferred.Media` and `Preferred.Countries` compile
  panics want a wrapped error at config-load time. **3a addressed this to
  3b or 3d; it is now 3d's alone**, because `internal/config` has no
  validation hook and 3b compiles no regex, so 3b was never its home.
  Also 3d's: the four-knob wall registration
  (`StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, `SearchLimit`),
  which stays the named follow-on task for the first rung 3 pass that
  runs after rung 1b closes, with `internal/config/keys.go` the file no
  lane edits until then; the compat switch's table selection and the
  `\b` regime divergence (guards 10 and 11); and the tie-mirroring work
  that needs both `inclArtist` call shapes (3a's guard 7).
- 3a's first carried item is closed by this pass. R1's absent-field
  resolution is guard 2 above, and no divergence is owed.
- Five `code-simplifier` leftovers, each seen and left deliberately:
  `wireInt`'s two decode errors do not wrap `ErrMalformedRelease`, while
  that sentinel's comment calls it the one sentinel every boundary
  failure wraps; `releaseLookupPrefix` now exists in three copies
  (`oracle`, `analysis`, `conform`), deliberately unshared so three
  independent walks never share a decision, which is worth revisiting if
  a fourth appears; `derive_album_info.py`'s `ALBUM_FIELDS` comment says
  "fourteen" over a thirteen-entry tuple; `Validate`'s data-track walk is
  observed by no test, a pre-existing gap that flipping its
  `withDataTracks` argument leaves green; and the m11 pattern-table guard
  now checks only the ledger-to-matcher direction.
- `profile_hashes` (m1) and the response-status extension, carried since
  2b and 2c, are unaffected by this pass and stay open.
- Adjudication batch 1 stays fully ruled, 18 `oracle-wrong` and 3
  `oracle-correct` with zero pending, and this pass produced no new
  adjudication row. **Four `corpus/README.md` cells that still described
  the batch as pending are corrected in this close**, along with the
  `stock.yaml`/`locale.yaml` placeholder claim and the
  regeneration-command list, which gains `album_info.json`'s command.

Budgets: ≈4.82M subagent tokens (the plan-prep workflow 2,106K across
three inventory readers, synthesis, a draft, two reviews, and the fold;
Task 0 109K; the Tasks 1 to 8 chain 2,057K across 22 agents;
`code-simplifier` 164K and its review 61K; the close draft 171K, its
prose review 62K, and the fold 90K) against the 6M ceiling. The 4.8M
flag was crossed by the fold, the pass's final dispatch, after every
deliverable had landed; the plan-prep workflow is the 2.1M the 3a
Budgets line never counted, and without it execution closed at 2.7M. Attended interaction points: zero. Planning
questions zero, since the plan was approved under the standing grants;
execution sittings zero; planning misses zero. Task 5's struck mutation
was a plan defect. Execution catches that shape of defect and no owner
question would have reached it, because the defect was invisible
without the code in front of it. The four conductor rulings taken at plan
landing (the nineteenth ledger subject, the three contested-surface rows,
the spec's tree-column fix, and STATUS's 3a spend figure) cost no
attended time.

## Rung 3, pass 3a (The scoring core, `string_dist`, and tier 0): 21 distance components, `internal/harness/conform` born (2026-09-04)

Nine plan tasks executed sequentially against
`docs/superpowers/plans/2026-09-04-rung-3a-scoring-core.md` (revision
2.1). Commits in order: `28396a7` (Task 1), `73d5321` (Task 2), a
same-session model restart pause at `7a21dcf`, `b5de09e` (Task 3),
`0592a32` (Task 4), `46c2eb6` (Task 5), `c5230ae` (Task 6), `f5b3243`
plus `ca50ff6` (Task 7), `91a16a4` plus `57bc094` (Task 8), `8d5a793`
(Task 6b, scope added mid-pass), and `05575d2` (the close-out
`code-simplifier` pass, 18 files, +205/-223).

What landed: `internal/engine/matcher`, a new package holding
`Tunables` (pinned beets 2.13.1's shipped `match:` defaults with a
citation per field) and the `AlbumInfo`/`TrackInfo`/`ItemFields`
vocabulary (Task 1); the `Distance` accumulator with seven of the
eight adders, both penalty views (`AddedKeys`/`PerKey` versus
`NonzeroItems`/`NonzeroKeys`), and the insertion-ordered summation
contract 5 requires (Task 2); the embedded `unidecode` 1.4.0
transliteration table via a committed one-time generator,
`tools/unidecode-table/generate.py` (Task 3); `StringDist` and the
eighth adder, `AddString`, plus the `SD_PATTERNS` pattern loop and its
byte-identity guard, which tier 0 cannot supply (Task 4); the seven
track-level components and `DistanceForTrack` (Task 5);
`internal/harness/conform`, a new package this pass creates,
with the typed `components.json` decoder (tier 0), the M1 defaults
cross-check, the `unidecode` manifest pin check, and the
engine-never-imports-harness / no-`time.Now` architecture checks
(Task 6); the fourteen album-level components (thirteen files,
`album_extras.go` the named one-file-two-component exception),
`MostCommonTags`, and `score.go`'s `albumComponents` composition
(Tasks 7 and 8); five reviewer-hardening follow-ups to the decoder
(malformed-row and mistyped-field failures, an add-shape coverage
test, a thresholds mutation test, and testdata-subtree exclusions for
both arch-check walks), landed as Task 6b, scope the conductor added
mid-pass from the Task 6 reviewer's non-blocking list rather than
carried to 3b (Task 6b).

AddString landed as a revision-2.1 editorial fold. The plan's Task 4
deliverable count moved 5 -> 6; no new work entered the pass.
Task 2's `diff-reviewer` escalated that the eighth adder was deferred
to Task 4 (`strdist.go`, alongside `StringDist` itself) but the plan's
Task 4 section never named it; the conductor accepted the escalation
and ruled the deferral correct (a placeholder would have been the
assertion-discipline rule's forbidden silent-success fake), then added
the two-line revision note at the plan's top, landed in this pass close
alongside the ledgers.

What the gate caught: nothing novel in `bash scripts/check.sh` itself
across all nine tasks; the pass's own fixtures are what proved a
mutation fires, the same discipline 2c and 2d both record. Two
findings worth a future pass's attention surfaced from hand-authored
mutation testing rather than the gate: `component_test.go`'s
`aggregate/sorted-penalties-tied` fixture cannot discriminate a
`(distance, key)`-ascending sort from a dropped sort at all (both an
insertion-ordered album and a key-ascending sort tie on that row's own
data), so the mutation the plan's M-b fold specified does not fire;
the implementer substituted a full key-order reversal
(`cmp.Compare(b.Key, a.Key)` in place of `cmp.Compare(a.Key, b.Key)`),
which does discriminate, and the `diff-reviewer` confirmed the
substitution independently rather than taking the report's word for
it. Task 4's Go/Python case-mapping sweep, run over every one of
the 1,112,064 valid Unicode code points against pinned beets in
`tools/venv/.venv`, found 56 where Go's simple-case `strings.ToLower`
disagrees with Python's full-case `str.lower`; 54 agree again once the
pipeline's post-transliteration `[a-z0-9]` strip runs, and two,
U+A7CB and U+A7DC, do not: both are unassigned in CPython 3.13's
Unicode 15.1 but lowercase under Go 1.27's Unicode 17.0, and
`unidecode` 1.4.0 maps their lowered forms to `"u"`/`"l"` where the
unlowered originals strip to nothing. This is a toolchain-Unicode-
version dependence the study never claimed and the port now carries
explicitly (`internal/engine/matcher/testdata/case-mapping-sweep.tsv`
names both rows so a third divergence reds the gate rather than
passing silently).

### Component coverage table

Twenty-one components: seven track-level (`internal/engine/matcher/track.go`'s
fixed order) and fourteen album-level (`score.go`'s fixed order).
`components.json`'s eleven tier-0 methods carry no album-distance rows
at all and no `track_distance` row that sets a length, index, MB track
ID, or disc number, so "tier-0 row?" is `No` for eighteen of the
twenty-one; every `No` still carries its own fire-proving and
decline-proving unit test in the component's own file, per contract 8
and the plan's assertion-discipline rule, distinct from the three
M8-declared corpus-coverage gaps below.

| Component | File | Weight key | Fire condition | Tier-0 row? |
| --- | --- | --- | --- | --- |
| track_length | `track_length.go` | `track_length` | `info.Length != 0`; `AddRatio` on `abs(item.Length-info.Length)-Grace` over `Max` | No (own fire/decline test) |
| track_title | `track_title.go` | `track_title` | always | Yes (`track_distance/*`, e.g. `identical`) |
| track_artist | `track_artist.go` | `track_artist` | `inclArtist && info.Artist != "" &&` item artist not in `VA_ARTISTS` (case-insensitive) | Yes (`track_distance/different-artist`, `tolerate-va`, `identical`) |
| track_index | `track_index.go` | `track_index` | `info.Index != 0 && item.Track != 0`; tolerates per-disc or absolute numbering | No (own fire/decline test) |
| track_id | `track_id.go` | `track_id` | `item.MBTrackID != ""` | No (own fire/decline test) |
| medium (track) | `track_medium.go` | `medium` | `info.Medium != 0 && item.Disc != 0` | No (own fire/decline test) |
| data_source (track) | `track_data_source.go` | `data_source` | always calls `AddDataSource` (its own three-way condition, contract 6) | Yes (`track_distance/data_source/*`, 10 rows) |
| artist | `album_artist.go` | `artist` | `!info.VA` | No (own fire/decline test) |
| album | `album_album.go` | `album` | always | No (own fire/decline test) |
| media | `album_media.go` | `media` | `info.Media != ""`; priority branch when `Preferred.Media` set, else equality against `likelies.Media` | No (own fire/decline test; anchor proved by `ca50ff6`'s dedicated rows) |
| mediums | `album_mediums.go` | `mediums` | `likelies.DiscTotal != 0 && info.Mediums != 0` | No (own fire/decline test) |
| year | `album_year.go` | `year` | three branches in order (preferred-original, likely-vs-release, decline); see below | No (own fire/decline test; branch 1 is a declared gap) |
| country | `album_country.go` | `country` | priority branch when `Preferred.Countries` set and `info.Country != ""`, else equality when both `likelies.Country` and `info.Country` set | No (own fire/decline test; anchor proved by `ca50ff6`'s dedicated rows) |
| label | `album_label.go` | `label` | `likelies.Label != "" && info.Label != ""` | No (own fire/decline test) |
| catalognum | `album_catalognum.go` | `catalognum` | `likelies.CatalogNum != "" && info.CatalogNum != ""` | No (own fire/decline test) |
| albumdisambig | `album_disambig.go` | `albumdisambig` | `likelies.AlbumDisambig != "" && info.AlbumDisambig != ""` | No (own fire/decline test) |
| album_id | `album_id.go` | `album_id` | `likelies.MBAlbumID != ""` | No (own fire/decline test) |
| tracks | `album_tracks.go` | `tracks` | always runs the pairing loop (zero pairs adds nothing) | No (own whole-distance tests, `91a16a4`) |
| missing_tracks | `album_extras.go` | `missing_tracks` | `len(info.Tracks) > len(pairs)`, one penalty per excess candidate track | No (own fire/decline test) |
| unmatched_tracks | `album_extras.go` | `unmatched_tracks` | `len(items) > len(pairs)`, one penalty per excess item | No (own fire/decline test) |
| data_source (album) | `album_data_source.go` | `data_source` | always calls `AddDataSource` | No (own fire/decline test) |

**The three corpus-unreachable surfaces are a declared corpus-coverage
gap (spec, M8). No fixture covers them, and none is owed in 3a.**
`match.required` and `match.ignored` carry their defaults and
citations on `Tunables` (`tunables.go:80-92`). The admission filters
that read them are 3d's, and 3a lands no consuming component for
either. Their coverage is 3d's table-driven unit tests, named in
`Tunables`' own doc comment. `preferred.original_year` is 3a's own:
it is `addAlbumYear`'s branch 1 (`album_year.go`), graded by a
hand-authored table
(`internal/engine/matcher/testdata/album_year_preferred_original.json`,
three rows) because no committed corpus case sets
`preferred.original_year`. The mutation that proves it fires:
`TestAddAlbumYearPreferredOriginalYearTakesPrecedence` asserts branch
1 runs, not branch 2's zero, when `Preferred.OriginalYear` is set even
though branch 2 would also apply; `f5b3243`'s own commit message
separately names the table's own proof, "a short-circuit mutation
that fails all three of its rows" (a mutant that returns after
checking only `info.Year != 0` reds the whole fixture).

### Evidence

Tier-0 replay (`go test -count=1 -v ./internal/harness/conform/...
-run TestComponentTableReplaysAllRows`), measured now: 72 of 72
committed rows replayed, all bit-identical, zero near-misses at the
1e-12 fallback. Per method: `add` 8, `add_data_source` 4,
`add_equality` 3, `add_expr` 2, `add_number` 4, `add_priority` 4,
`add_ratio` 4, `add_string` 3, `generic_penalty_keys` 1, `string_dist`
23, `track_distance` 16. Total 72, matching `ComponentRowFloor`
exactly. Defaults cross-check
(`TestCommittedCasesEchoTunablesDefaults`), measured now:
weights and thresholds cross-checked against all 215 committed cases,
zero divergences (no committed profile sets a `match:` key). Full
suite (`go test -count=1 ./...`): 12 packages, all green, real 13.4s.
`bash scripts/check.sh`, measured now over two consecutive runs: total
wall time 1m45.6s (gitleaks self-reporting 1m39.6s / 99.6s) and, on a
second run minutes later, 3m24.6s (gitleaks self-reporting 3m13s).
Gitleaks' full-text scan of the tracked worktree (`--no-git`,
~112.9 MB, `corpus/cases/` the dominant contributor) is both the
gate's dominant cost by a wide margin and its most variable one,
roughly doubling between two runs with no tree change in between. The
version-bump differential doc's next update should budget against the
higher figure rather than the first sample. The remaining ~6-12s
across both runs covers `make check` (go vet, golangci-lint 0 issues,
`go test ./...` cached from the fresh 13.4s run above, `go mod tidy`
clean), shellcheck, ruff (6 tracked `*.py` files, all formatted, all
passing), vale over Go comment prose, `docker compose config`, and the
pattern-filtered `systemd-analyze verify`, combined; none of those
self-report an individual wall time in check.sh's own output. **No
`make derive` ran anywhere in this pass**: no task touched oracle
territory, Task 1's venv read re-verified the M2 resolution and
Task 3's venv read drove the one-time `unidecode` generator, and
neither is a derivation.

### What a later pass would be wrong to rediscover

1. The rung 15 re-run criterion, verbatim from the plan's Spec
   section, travels forward from every rung 3 pass close: "Rung 15
   re-runs `internal/harness/conform`'s three tiers over the complete
   corpus, at the matcher commit as landed, and its acceptance is the
   launch bar's matcher clause read against that run: zero false
   confident matches over every stratum including stock-config, and
   zero unadjudicated beets-right/Go-wrong cases. Rung 3's own
   acceptance is provisional against the corpus at its
   then-current size. Any tier-2 divergence carrying a pending row at
   rung 15 is the launch gate's business, not rung 15's to wave
   through."
2. **The four-knob wall registration is a named follow-on task** for
   the first rung 3 pass that runs after rung 1b closes:
   `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, and
   `SearchLimit` are named on `Tunables` today with citations but
   registered on no operator wall, because `internal/config/keys.go`
   still carries `ownedWall` as a slice literal. No lane may edit that
   file until rung 1b lands.
3. **The `SD_PATTERNS`-as-a-value seam**: `BeetsSDPatterns`,
   `SDPatternsWordBoundary`, and `Tunables.StringDistPatterns`
   (defaulting to `BeetsSDPatterns`) let 3b's `defect_free` derivation
   and 3d's compat switch swap a table rather than edit a component.
   Both measured figures travel with the warning that produced them:
   the word-boundary guard changes **zero of the 72 tier-0 rows**, so
   tier 0 is blind to a "fixed" table. Task 4's byte-identity test
   against beets' six shipped literals is the only check that catches
   a silent improvement to the table.
   `string_dist("Draft: Ember Wharf", "Draft: Northern Lines")` reads
   `0.061111111111111116` under `BeetsSDPatterns` and
   `0.6111111111111112` under `SDPatternsWordBoundary`.
4. **The M2 resolution and its two drafted `corpus/README.md` cells**
   are landed in this close: `original_date` and `per_disc_numbering`
   change from "verdict-moving: ..." to "applied-field-moving, never
   verdict-moving: ..." in the config-profile transform table
   (`corpus/README.md`), and the stock-config stratum finding's own
   single-accessor sentence is corrected to name both accessors
   (`AlbumInfo.item_data` for `original_date`, `TrackInfo.raw_data`
   for `per_disc_numbering`). Also landed: the mechanics study's Q4
   "19 keys" claim is superseded. The tree and
   `config_default.yaml:174-193` (twenty entries; `:173` is the
   header) agree on twenty `distance_weights` keys, matching all 215
   committed cases' own echoed weight sets.
5. **Four corrected citations**, so no later pass re-derives them:
   `DataSourceMismatchPenalty`'s default is `metadata_plugins.py:180`
   wired at `:202`, not `:190-195`; the `SearchLimit` legacy spelling
   is `musicbrainz.match.searchlimit` (corrected at 3c's plan fold), not
   a top-level `match:` key;
   `ExtraTags` lives on the musicbrainz plugin with no
   `config_default.yaml` citation; and `match:` carries
   `album_disambig_fields` / `singleton_disambig_fields`, display-only
   and deliberately outside `Tunables`.
6. The decoder contract as built, for 3c and 3d to extend rather
   than re-derive from `derive_component.py` a second and third time:
   `internal/harness/conform`'s decoder dispatches on `method`, not
   row id, across eleven methods; method `add` splits on an
   add-key-versus-weights shape (twelve total shapes across the
   eleven methods); a row's `weights` map is an override on
   `Tunables.DistanceWeights`, never a complete table; the
   `get_penalty` stub is wired through a loaded-source set built from
   `Tunables` (`metadataSourceSet`); and a compiled-regex option
   decodes from `{"regex", "ignorecase"}` into a `matcher.PatternOption`,
   never from a bare string.
7. **The `inclArtist` divergence**: `DistanceForTrack` takes
   `inclArtist` as a parameter, not a config read, because beets' two
   call sites disagree. The assignment solve (3d) calls
   `track_distance` with `inclArtist=false` for every track, while the
   final per-track distance a user sees passes `album_info.va`. The
   two objectives diverge exactly on VA albums; 3d's tie-mirroring
   task needs both call shapes.
8. **The `unidecode` pin check is split across two packages**:
   `matcher.UnidecodeVersion` (`translit.go:20`) is the exported stamp
   the embedded table carries; `conform.CheckUnidecodeVersion`
   (`defaults.go:67`) compares it against `corpus/manifest.json`'s
   pinned value, so a manifest bump reds the Go side too, not only the
   Python derivation path.
9. No row passed at 1e-12 without bit-identity. The near-miss list
   is empty across all 72 rows; there is nothing to name.
10. Tier-0 replay ran 0.02s standalone; the full `go test -count=1
    ./...` suite 13.4s. `bash scripts/check.sh` ran 1m45.6s and, on a
    second run minutes later with no tree change, 3m24.6s (both
    dominated by gitleaks' full-worktree scan, 99.6s and 3m13s
    respectively). The version-bump differential doc's next update
    should budget against the higher figure for the matcher and
    conform packages.

Two further findings sit outside the ten: the pointer ruling and the
Key-field ruling. `StringDist` and
`AddString` both take `*string` pairs where a nil pointer is Python's
`None` (Task 4's ruled API,
`StringDist(tun Tunables, a, b *string)`); every caller in `track_title.go`
and the album files passes a typed default's address, never a literal
nil, because `AlbumInfo`/`TrackInfo`/`ItemFields` fields are all plain
strings (Task 1), which is also the root of the `album_album`
divergence below. `trackComponent`/`albumComponent`'s `Key` field
sat unread from Task 5 through the start of Task 8; Task 8 gave it a
reader, `TestComponentKeysAreWeighted`, asserting every declared key
names an entry in `Tunables.DistanceWeights`, closing what would
otherwise have been dead struct data.

### Carried

- **To 3b (translation layer):** `addAlbumAlbum` passes
  `&likelies.Album`/`&info.Album`, both always non-nil, so a beets
  `album_info.album` of `None` scores `0.0` here where beets itself
  scores `string_dist("", None) = 1.0`. Root cause is Task 1's
  plain-string `AlbumInfo` fields; 3b needs to decide what an absent
  source field becomes before this divergence reaches a real case.
- **To 3b/3d (config validation):** `compileMediaPatterns` and
  `compileCountryPatterns` panic on a malformed user regex in
  `Preferred.Media`/`Preferred.Countries`, matching beets' own
  `re.compile` raising, but a wrapped error at config-load time is the
  better home once `internal/config` owns validation.
- **To 3c:** `dataSourcePenalty` falls back to
  `Tunables.DataSourceMismatchPenalty` (configured) where beets falls
  back to the class literal `0.5`; identical under shipped config,
  diverges only if `source_weight` is set.
- **Inert deviation, recorded and reviewer-verified, no action
  owed:** `MostCommonTags` counts every item with `len(items)` as the
  denominator, where beets' `if item` filter is a no-op for real
  `Item`s (no `__bool__`/`__len__`); `MostCommonTags` on an empty item
  set reports unanimity where beets asserts non-empty, unreachable
  today.
- `profile_hashes` (m1) and the response-status extension, 2d's own
  carry-forward, are unaffected by this pass and stay open. The
  adjudication batch was already fully ruled before this pass started:
  18 oracle-wrong, 3 oracle-correct, zero pending
  (`corpus/analysis/adjudications.json`, ruled Geoff 2026-09-04), and
  3a produced no new adjudication row. Rung 2's Tier C exclusions
  (`match.required`/`match.ignored`, `preferred.original_year`) are
  now the M8-declared gap documented above rather than an open
  unknown.

Budgets: ≈2.84M subagent tokens (T2 review 74K, batch 1 445K, batch 2
561K, batch 3 938K, `code-simplifier` 238K and its review 63K, the
close draft 193K, its prose review 63K, and the fold 261K) against the
6M starting ceiling, well under; conductor context on top. Attended
interaction points: zero. The one owner touch was a model-default
change mid-close, unrelated to the pass.

## Rung 2, pass 2d (File-level applied tags, the version-bump differential, and the whole-corpus README close): file_tags.json over 12 subjects, rung 8 unblocked (2026-09-04)

Four plan tasks executed sequentially against
`docs/superpowers/plans/2026-09-03-rung-2d-file-level.md` (revision 2).
Commits in order: `8d181fe` (Task 3, which ran first per the plan's
sequencing), then `4d38717` (Task 1), then `f5abba0` (Task 2), then
`412471f` plus `cc400ad` (Task 4's close and its same-session
forward-reference fix).

What landed: the template FLAC fixture and its provenance script
(`internal/harness/oracle/testdata/template.flac`, 42 bytes, silent,
untagged, authored via the pure-`mutagen` route with no `ffmpeg`/`flac`
dependency), `derive_file_tags.py` and its Go driver implementing the
write-then-independent-read-back contract with `BEETSDIR` isolation
proved rather than assumed, and `Manifest.MediaFile` pinning
`mediafile==0.17.0` at `schema_version` `3` with `Reconcile`'s `pinned`
tuple and `versionProbe` string both gaining their sixth component
(Task 1); `corpus/analysis/file_tags.json`, generated over a verified
twelve-subject set, its pure-Go gate-resident checker
(`internal/harness/analysis/file_tags.go`) wired into `corpus_test.go`
with three firing fixtures (an unresolved case id, a value key outside
the writable-set header, and — the fixture with real teeth — a row
carrying one of the nine read-only keys), and
`TestFileTagsLedgerMatchesTheCorpus` on the
`TestComponentTableReDerivesToItsCommittedTable` template (Task 2); the
version-bump differential job doc and the quarterly re-capture chore
doc, both standing operational docs beside `docs/bringup-runbook.md`,
the differential doc's wall-time section priced at the real 215-case
corpus with the file-tags line left as a named placeholder for this
task (Task 3); the whole-corpus case-inventory section (215 rows
across five stratum tables, built by a mechanical script pass over
each case's own `id`/`profile`/`eval_date` cross-referenced against
the family/task/adversarial/stock-config tables already committed,
never by re-reading a case body) and the rung 8 handoff statement,
both new `corpus/README.md` sections; the file-tags line written back
into the differential doc's placeholder (Task 4). Corpus total: still
**215 cases**, unchanged from 2c's close — 2d authors no case, captures
nothing, and edits no committed record, exactly as its own global
constraint required.

**The twelfth file-tags subject is a ruled substitution, and this
pass's own staleness catch.** The plan's Task 2 named
`real-decoy-beethoven-symphony-5` as the two-release sub-axis's
file-tags subject, drawn from the planning inventory dossier written
before 2c executed. That id was never captured: 2c's own Task 4 had
already proved the decoy unreachable inside beets' 5-result search
window with the axis's sole reserve spent, closing the axis at 7-of-8
(`corpus/README.md`, "The two-release sub-axis finding"). Task 2's
implementer caught this with a STOP rather than padding or trimming
the twelve-subject set, and the conductor ruled a substitution:
`real-decoy-mozart-requiem`, chosen from the axis's four other
committed decoy cases by eligibility (all four carry a non-empty
`applied_fields` record regardless of `action`, since
`match.apply_metadata()` runs whenever a candidate exists at all) and
then by the most distinctive applied-tag shape among the eligible rows
(24 items, two movement-numbering conventions, two distinct
`albumartist_credit` strings). Recorded with its eligibility evidence
in `corpus/README.md`'s file-tags section. **The standing lesson: a
plan drafted before a prior pass closes must re-verify its id lists
against the tree at execution time, not trust the dossier that
predated the pass's own findings.** This is not a new failure mode —
2c's own Task 17 alias-duplicate incident carries the sibling lesson
("check `corpus/cases/` before any capture") — but this pass's version
of it is a plan citing a specific id the immediately prior pass had
already retired, which a same-pass cross-reference (rather than a
later implementer's own tree check) could in principle have caught
earlier. It did not block the pass; the implementer's own verification
step, required by Task 2's contract regardless, is what caught it.

What the gate caught: nothing novel — `bash scripts/check.sh` and
`GOFLAGS=-count=1 make derive` stayed green through all four tasks,
and the pass's own fixtures (Task 1's silent-direction `Reconcile`
proof, Task 2's three gate-resident checker fixtures plus its
re-derive-catches-a-perturbation fixture) are what proved each new
check has teeth rather than the gate itself surfacing a defect in
existing code.

Determinism: two independent uncached (`GOFLAGS=-count=1`) full
`make derive` runs against the whole `internal/harness/...` and
`cmd/harness/...` tree with the file-level derivation in the tree, 215
and 215 cases actually re-derived of 215 committed
(`TestCorpusReDerivesToItsCommittedRecords`'s own per-case subtests),
both runs exiting 0 with zero failures across every package, both
byte-identical to the committed records (corpus, ties, and component
ledgers, plus the twelve-subject file-tags ledger — `git status`
reported no diff under `corpus/` after either run). Wall time: run 1
980s (16m20s) total (corpus derive 462.94s, ties derive 503.90s,
component derive 0.75s, file-tags derive 0.59s); run 2 974s (16m14s)
total (corpus derive 541.20s, ties derive 420.06s, component derive
0.71s, file-tags derive 0.55s) — this is rung 3's budget line. 2d's
pair (980s, 974s) ran materially faster than 2c's own pair
(~1729.58s/1582.53s) on the identical 215-case corpus, roughly 40
percent faster; the spread is attributable to run-to-run variance in
request pacing rather than any corpus change, and the file-tags leg's
sub-second cost (0.55s-0.59s) is negligible under either figure. The
file-tags leg issues no MusicBrainz
request (twelve subjects, one Python process reading already-committed
`AppliedFields`), so it adds well under one second against the corpus
and ties legs' own multi-hundred-second cost; the measured figure
(0.55s-0.59s) is written into
`docs/corpus-version-bump-differential.md`'s previously-placeholder
row, alongside a new finding that doc records: the corpus and ties
legs' own M8-formula floor (257.6s each) is roughly half the real
measured wall time (462-541s and 420-504s respectively) — the formula
prices only the client's inter-request wait, not the Python/beets/
response-parsing overhead sitting on top of it, so a future
version-bump differential run should budget against the measured
figure, not the formula table.

What a later pass would be wrong to rediscover:

- **The rung 8 handoff statement lives in `corpus/README.md`** ("##
  Handoff to rung 8: what pure-Go FLAC tag writing may assume"),
  positioned immediately after "## The analyses" so rung 8's
  implementer finds it beside the evidence rather than in a fifth
  loose document. It states the two altitudes of expected applied
  tags (`Expected.AppliedFields` on every case, `file_tags.json` on
  twelve subjects), names the 69-key writable set as the only
  conformance target and the 9-key read-only remainder as never one,
  and states its own non-scope explicitly: the apply-metadata field
  mapping close read and the wazero-taglib spike are rung 8's own
  plan-opening work, not something 2d evaluated or prototyped.
- **2c's `profile_hashes` gap (m1) is still open**: `corpus/manifest.json`
  carries a `profile_hashes` entry per profile, but nothing in the gate
  or the derivation path computes or checks it for any profile yet.
  2d did not touch this; it is still scoped follow-on work, not
  something the manifest's own presence implies is built.
- **The response-status extension is still deliberately not taken.**
  Carried from 2c: the stub has no third outcome and the schema
  carries no response status, since no known-oracle-bug class probed
  so far needs a 404 or 503 to reproduce. Available to whichever later
  rung first has a case that genuinely needs it.
- **The Tier C exclusions and profile gaps 2b and 2c both carried,
  neither pass closed, carry forward again unchanged**: multi-source
  ordering and most of the `data_source` parameter space stay blocked
  by the single-source ruling; `match.required`/`match.ignored`
  candidate-dropping stays unprobed; `preferred.original_year: yes`
  stays unreachable under `estate`, `stock`, and `locale` alike (a
  fourth profile or a `locale.yaml` extension is still owed if a later
  rung wants this branch probed). 2d's own scope never touched any of
  these; nothing new closed here either.
- **The pending adjudication batch is still the launch gate's open
  item**, unchanged by this pass: nineteen rows across five hypotheses
  sit at `disposition: "pending-adjudication"` in
  `corpus/analysis/adjudications.json`
  (`docs/superpowers/research/2026-09-03-adjudication-batch-1.md`).
  2d's own gate accepts them indefinitely, per its own "Owner
  touchpoints: none" scoping; it is the launch gate, later, that
  refuses an unadjudicated divergence. 2d produced no new
  adjudication row of its own.
- **The `schema_version` `2` -> `3` bump forced no re-freeze**, the one
  place this pass's schema event differs from 2b's: `Manifest` is
  corpus metadata rather than an `Expected` field, so `Manifest.MediaFile`
  landing needed no re-derivation of any committed case's own record,
  only the two-clean-run byte-identical proof above.

Carried:

- Rung 2 is complete at this pass's close: the smoke, structural,
  real-catalog, adversarial, and stock-config strata (215 cases), the
  file-level applied-tag ledger, and the two standing chore docs are
  all shipped. Rung 3 (the Go matcher) is what rung 2's completion
  unblocks, gated on Geoff reading its revision-2 spec before
  execution starts (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`).
- Every item 2c's own carry-forward left open (`profile_hashes`, the
  response-status extension, Tier C exclusions, the pending
  adjudication batch) stays open into rung 3 and beyond, restated
  above rather than re-derived.

Budgets: ≈1.6M subagent tokens (Tasks 3/1: 787k with the first
Task 2 STOP; the ruled resume of Tasks 2/4: 728k; code-simplifier:
94k) against the 2.5M starting ceiling, well under. Attended
interaction points: zero — the twelfth-subject STOP resolved to the
conductor, exactly as the plan's owner-touchpoints section required.

## Rung 2, pass 2c (Real-catalog and adversarial strata): 147 new cases, six oracle-bug hypotheses, first adjudication batch (2026-09-03)

Twenty plan tasks executed across a single sequential pass-execute
chain against
`docs/superpowers/plans/2026-09-03-rung-2c-catalog-adversarial.md`
(revision 2). Commits `3bbdfbd` → `890bb70` (Task 20 close at `f5bb1bd`; the
locale re-base and 19b prose fix at `890bb70`).

What landed: the adjudication and known-oracle-bug ledger formats
(Task 1); the `harness seed-items` projection and `--facts` tools
(Task 2); the real-catalog selection dossier, 96 primary rows plus 12
reserve rows across seven axes (Task 3); twelve capture batches (Tasks
4-15) landing 94 real-catalog primary cases against real MusicBrainz
document shapes (two two-release sub-axis rows lost to the mechanism's
own escape hatch, real captures over MB-noise, non-Latin script,
classical/compilation, genre/era spread, reissue/remaster, and
VA/box-set releases, one `ignore_video_tracks` finding); the
adversarial stratum (Tasks 16-18), synthetic cases reproducing two
known string-heuristic beets defects (the `"ft"` word-boundary false
positive and the parenthetical/remix near-duplicate discount, ten
cases, both moving strong/apply to medium/skip) plus six
per-pair track-assignment rejection-gap cases (the album verdict never
moves even while `assignment`/`applied_fields.title` are provably
wrong) plus the alias-mis-selection hypothesis probed against two of
Task 4's own real-nonlatin cases, plus Task 18's `locale` profile and
its A/B regression-check case pair for `#6358`; the stock-config
stratum (Tasks 19a/19b), 36 siblings (28 mandatory + 8 supplementary)
proving zero verdict-level divergence between `estate` and `stock`
across every real-catalog axis, with the divergence landing only in
`applied_fields` (`year` for reissues, `tracktotal` for box sets).
Corpus total: **215 cases** (7 smoke + 61 structural + 94 real-catalog
primary + 17 adversarial + 36 stock-config), read off the tree
directly. Determinism: two independent uncached (`-count=1`) full
`make derive` runs against the whole `internal/harness/...` and
`cmd/harness/...` tree, 215 and 215 cases actually re-derived of 215
committed (`TestCorpusReDerivesToItsCommittedRecords`'s own per-case
subtests, the count that proves a case was actually re-derived rather
than skipped for want of a committed record), both runs exiting 0 with
zero failures across every package, both byte-identical to the
committed records. Wall time: run 1 1729.58s (28m50s) total (corpus
derive 927.91s, ties derive 786.04s, component derive 1.22s); run 2
1582.53s (26m23s) total (corpus derive 812.66s, ties derive 754.48s,
component derive 1.27s) — this is rung 3's budget line, exactly as
2b's was 2c's own (2b's own pair ran ~2m13s each on 68 cases; 2c's
real-catalog and box-set growth moved that budget line by roughly
13x). The heaviest single cases in either derive phase are the
box-set axis's largest releases: ties derive's own re-solve-per-pair
algorithm (`internal/harness/oracle/derive_ties.py`) pays for
`real-boxset-rachmaninov-complete`'s 470 tracks and its stock twin
individually, each taking on the order of a minute alone, the same
box sets that dominate the byte-growth figure below. Corpus byte
growth (m7): the committed corpus held 4.0 MB across 114 response
files before this pass; it now holds **81.3 MB across 351 response
files**, well past the plan's projected 10-15 MB
landing point. The overshoot is a response-*size* story, not a
response-*count* one (351 files is close to what the plan's own count
math expected): a handful of the VA/box-set axis's largest releases
serialize their entire tracklist inline in one WS/2 document
(`real-boxset-rachmaninov-complete`, 470 tracks, 9.1 MB alone;
`real-boxset-scriabin-complete`, 321 tracks, 3.1 MB;
`real-boxset-mahler-complete`, 133 tracks, 1.7 MB), and Task 19a's
mandatory-sibling `cp -r` rule doubles every one of them for its stock
twin.

What the gate caught: nothing new at Task 20 itself (README-only and
analysis-doc changes, no code); across the pass, the recurring catch
was a case's own `claims.json` failing to notice a degraded record
(the claims-checker's whole design point) and, once, a stratification
undercount in Task 19b's own README prose (fixed same-pass, commit
`9915907`).

What a later pass would be wrong to rediscover:

- **File-level derivation is 2d's Task 13, unchanged in scope**:
  `internal/harness/oracle/derive_file_tags.py` and its Go driver,
  reading back independently through `mediafile.MediaFile(path)` after
  `item.write()`, never trusting the write call's own return value,
  over 12 subject cases (5 from the committed corpus, 7 flagged by
  this pass's dossier).
- **The audio-fixture route is ruled** (M4): vendor one small static
  template FLAC blob under `internal/harness/oracle/testdata/`,
  committed, with its provenance and byte size recorded in the README.
  `make derive` gains no binary dependency; `ffmpeg`/`flac` are not
  adopted as prerequisites.
- **`derive_file_tags.py` is a second, independent entry point** (M5),
  shaped like `derive_component.py`: it reads a committed record's
  `AppliedFields`, builds beets items from those fields, allocates its
  own scratch directory, writes and reads back there, and is never
  called from `oracle.Derive`. The invariant: no item in the corpus
  derivation path ever gains a non-empty `path`, proved by `make
  derive` re-deriving all committed records byte-identically after 2d's
  change.
- **`file_tags.json` needs an oracle-tagged re-derive-and-diff test**
  (M3), `TestFileTagsLedgerMatchesTheCorpus`, matching
  `TestTieLedgerMatchesTheCorpus`/`TestComponentTableReDerivesToItsCommittedTable`'s
  precedent, plus the `-update-ledgers` regeneration path documented in
  `corpus/README.md`. A gate-resident structural check alone (case ids
  resolve, pointers resolve) is not enough; nothing else would catch a
  `mediafile` version bump moving the tag shapes.
- **`Reconcile` lives in `internal/harness/oracle/reconcile.go`, not
  `oracle.go`, and its `versionProbe` tuple names five components
  today** (m2: beets, unidecode, jellyfish, lap, numpy), not four. 2d's
  Task 13 must edit `reconcile.go` and `versionProbe` explicitly when
  it adds `mediafile` to `Manifest`, or a reconcile silently compares
  against a missing key.
- **The version-bump differential and quarterly re-capture chore docs
  are unwritten**, deferred with Task 13. The differential doc's
  wall-time section must use the M8 inter-request-floor formula
  (`1.04s + 0.25s × (requests − 1)`) at the corpus's real size once 2d
  closes, not a flat per-case figure.
- **The whole-corpus README rewrite happens again at 2d's close**,
  once, to add the file-tags section; this pass's rewrite is not
  wasted work, since 2d's version extends it rather than replacing it.
- **The response-status extension was deliberately not taken this
  pass**: the stub has no third outcome and the schema carries no
  response status, because all four string/assignment/alias
  known-oracle-bug classes this pass probed are distance or selection
  defects, not transport defects — none needs a 404 or 503 to
  reproduce. It carries forward as a named non-adoption, available to
  whichever later rung first has a case that genuinely needs it.
- **The `profile_hashes` gap (m1)**, distinct from the 2d file-tags
  items above: `corpus/manifest.json` gained a `locale` entry under
  `profile_hashes` at Task 18, but nothing in the gate or the
  derivation path computes or checks it yet for *any* profile. A
  profile edit does not fail loudly the way a beets version pin
  already does; this is scoped follow-on work, not something the
  table's presence implies is built.
- **Tier C exclusions and profile gaps 2b carried that 2c did not
  close**: multi-source ordering and most of the `data_source`
  parameter space stay blocked by the single-source ruling (unchanged
  from 2b); `match.required`/`match.ignored` candidate-dropping stays
  unprobed (same shape, still open); `preferred.original_year: yes`
  stays unreachable under all three profiles now committed (`estate`,
  `stock`, and Task 18's own `locale`, which only toggles
  `import.languages`/`import.ignored_alias_types`) — a fourth profile
  or an extension to `locale.yaml` is still owed if a later rung wants
  this branch probed.
- **The alias-duplicate incident (Task 17) and its standing rule**: an
  earlier pass of Task 17 captured four byte-identical `adv-alias-*`
  cases duplicating releases the corpus had already committed under
  Task 4's `real-nonlatin-*` rows, because case selection never
  checked the committed tree first. The conductor's escalation ruling
  folded the four duplicates back into their existing twins (commits
  `e805540`, `c685894`); nothing was lost, but the rework was pure
  waste. **The rule a later pass must inherit: check `corpus/cases/`
  before any capture**, real or synthetic, that might already exist
  under a different id.
- **The `_preferred_alias` unreachability finding and Task 18's
  resolution of it**: Task 17's alias-mis-selection probe found
  `_preferred_alias` (`beetsplug/musicbrainz.py:128-155`) structurally
  unreachable under both `estate` and `stock`, since both leave
  `import.languages` empty, so its generator yields nothing regardless
  of what a release document carries. This was left as a scoping
  finding, not a reproduction, and deferred to Task 18, which added the
  `locale` profile and an A/B case pair
  (`real-nonlatin-utada-first-love` / `-locale`) to probe it directly.
  Task 18's first cut based `locale.yaml` on `stock.yaml` (as the plan
  itself prescribed), which the review caught as a confound: the pair
  carried the whole estate/stock delta besides the locale keys. The
  conductor ruled the clean experiment and the profile was re-based
  same pass (commit `890bb70`): `locale.yaml` is now `estate.yaml`
  plus exactly `import.languages: [ja]` and
  `import.ignored_alias_types: []`, so every divergence in the frozen
  pair is attributable by construction. The result: the mechanism *is*
  reachable and *does* fire once `import.languages` names a matching
  locale, moving the four sort-name fields it owns (`artist_sort`,
  `albumartist_sort`, `artists_sort`, `albumartists_sort`) on all 12
  items, plus one genuine bonus finding reported verbatim: the
  MusicBrainz "Various Artists" *entity* itself carries a primary
  `locale=ja` alias, so `_preferred_alias` also fires for two
  VA-credited losing candidates and moves their scoring inputs
  slightly (no verdict effect). Everything else in the pair is
  byte-identical — evidence against `#6358`'s report, pending Geoff's
  own ruling (`bug-locale-alias-ignored`, adjudication batch 1).

Carried:

- **The pending adjudication batch is the launch gate's open item.**
  Nineteen rows across five hypotheses sit at
  `disposition: "pending-adjudication"` in
  `corpus/analysis/adjudications.json`
  (`docs/superpowers/research/2026-09-03-adjudication-batch-1.md` is
  the artifact prepared for Geoff's ruling). This pass's own gate
  accepts them indefinitely; it is the **launch gate**, a different
  gate at a later time, that refuses an unadjudicated divergence.
  `bug-threshold-boundary` (`#5458`) is the corpus's one hypothesis
  already ruled (`oracle-correct`, Geoff, 2026-09-03) and is not part
  of the pending batch.
- `corpus/analysis/divergences.md` stays empty at this pass's close,
  reviewed and carried forward: no case across the real-catalog,
  adversarial, or stock-config strata contradicted a specific beets
  test assertion it was modeled on (findings of that shape instead
  became adjudication-batch rows or plain README prose).
- The two-release sub-axis closed at 4-of-6 primary rows delivered (2
  discarded via the mechanism's own escape hatch, the sole reserve
  spent by Task 4); a later two-release row anywhere in the corpus
  should budget for the same WS/2 search-ranking risk with no reserve
  behind it.
- Every item from 2b's own carry-forward that 2c did not separately
  close above (Tier C, `match.required`/`match.ignored`,
  `preferred.original_year`) stays open into 2d and beyond.

Budgets: ≈8.2M subagent tokens total (≈6.5M through Task 17's fold,
1.4M for Tasks 18-20's chain, 0.27M for the close fixes) against the
8M post-Task-3 re-derived ceiling (12M was the plan's pre-re-derivation
starting figure; finishing past the 80% flag approved by Geoff at the
flag, 2026-09-03). Attended interaction points: one mid-pass (the
combined 80%-flag question); the adjudication batch is delivered, not
yet spent.

## Rung 2, pass 2b (Structural stratum): analyses, synth, 61 synthetic cases (2026-09-02/03)

Fourteen plan tasks plus one staleness pre-task, executed overnight
across seven pass-execute workflows against
`docs/superpowers/plans/2026-09-02-rung-2b-structural-stratum.md`
(revision 2). Commits cec3c12 → 2219b19. All tasks accepted; five
needed one fix round; four escalations, each ruled by the conductor
and recorded in
`research/2026-09-02-rung-2b-staleness-corrections.md`.

What landed: the machinery half (threshold/weight echo with schema v2
re-freeze; `harness freeze` with `Options.Dir`; `harness synth` and
the authoring loop; the proximity flag, tie detector, per-case claims
checker, and 72-row component derivation table; the shared
`mbrequest` package), then the 61-case structural stratum across all
15 shape families, every record derived through real pinned beets,
plus the claims schema's zero-candidate mode and `penalty_keys_absent`
field. Corpus total: 68 cases, all re-deriving byte-exact across two
independent uncached runs (~2m13s each: corpus 70.5s, ties 54s,
component 0.7s; this is 2c's budget line).

What a later pass would be wrong to rediscover: the staleness review
pattern paid for itself (a fresh-context read of the case tasks
against the machinery as shipped found five blockers before any
dispatch; plan staleness remains the dominant defect class); the
"combined" completeness shape is unexpressible (`lapjv`'s leftover
set is one-sided); the VA-misflag has NO record-level observable
under the empty-search convention, so its witness is the frozen
request log's `arid:` query, which `make derive` mechanically
enforces; the response hand-edit escape hatch (edit generated
responses BEFORE first freeze, named in the README row) is the
sanctioned route for shapes `ReleaseSpec` cannot express; a
sub-second `make derive` is a cache hit and proves nothing, so force
`-count=1` for evidence.

Carried: two smoke cases tied (pinned; rung 3's solver must mirror
lapjv tie-breaking); family 15 parks cases on thresholds with
acknowledged proximity rows; the candidate-absent country direction
documented as declined; claims formatting drift cosmetic. 2c's
inventory is unchanged from the 2a plan's split section, plus the
determinism budget line above.

Budgets: ≈5.8M subagent tokens against the 6M ceiling (Geoff's 12M
overnight grant untouched); human interaction: two batched bedtime
questions (ceiling, 2c scope), zero mid-execution.

## Rung 2, pass 2a (Cassette substrate and capture): schema, stub, shim, capture tool, smoke corpus (2026-09-02)

Executed under two pass-execute workflows (wf_5ed5b773, wf_fe30d704)
plus one escalation follow-up, against
`docs/superpowers/plans/2026-09-02-rung-2a-cassette-substrate.md`
(revision 2). Commits 79420d3 → 69a1804. Five plan tasks, all accepted;
one fix round (task 5); one escalation, ruled in-scope and closed by
follow-up commit 6d45c28.

What landed: `internal/harness/cassette` (the frozen schema, verbatim
response-file store, byte-wise canonicalization rule);
`internal/harness/stub` (replay server with a miss list) and
`internal/harness/oracle` (adapter proven against a fake shim, then the
real `derive_oracle.py` driving pinned beets 2.13.1 in a uv venv, with
internal self-verifying date pinning and a raw-stdout determinism
proof); the two config profiles with their documented transform;
`cmd/harness` (capture tool: `--mbid` seeding with maximal `inc=` and
the pseudo-release double fetch, `--case` miss consumption, 1 req/s
identified wire discipline, contact refusal); and a seven-case live
smoke corpus (clean-match, should-ask, should-reject, various-artists,
multi-disc, non-latin pseudo-release with its transl-tracklisting
follow-up, year-penalty), every record derived through the real shim,
all seven re-derivable from the committed tree
(`TestCorpusReDerivesToItsCommittedRecords`) with the date-pinning
proof riding the real year-penalty case. Gate grew ruff and gitleaks
legs; `make derive` runs the oracle-tagged tests.

What a later pass would be wrong to rediscover: the `inc=` join is
`%2B` on the wire, never a literal `+` (requests percent-escapes it;
every live case missed its primary lookup until the capture tool's
canonicalization matched, and the original cassette doc's "literal +"
wire fact was corrected in 6d45c28; do not "fix" it back); the stub
must key on `r.URL.EscapedPath()`, since `r.URL.Path` arrives
percent-decoded; the miss loop's contract is `<case dir>/misses.json`,
written by `Derive`, consumed by `harness capture --case`; a
strong ID match legally orphans search captures into `Extra` (recorded
fixed point, not a stable request set).

Carried to 2b/2c: three non-blocking hardenings on
`corpus_derive_test.go` (skip-all-green residual hole, unguarded
`Candidates[0]`, redundant equality branch); the line-pinned
`.gitleaksignore` fingerprint; git-history gitleaks scanning (deferred
as a heavier job); response-number gaps in two cases (documented,
harmless).

Budgets: ≈1.75M subagent tokens against the 5M ceiling, plus conductor
context. Human interaction: one batched ask (the MusicBrainz contact,
minted as `DUBPLATE_HARNESS_CONTACT` in `~/.bashrc` via dotfiles
3a1c551; registration declined, anonymous 1 req/s stands).

## Rung 1 (Foundation): the repo pivot, config core, check/doctor (2026-09-02)

The greenfield build's first rung, run under the pass-execute workflow
(implementer/diff-reviewer/gate chain per task, owner opt-in). Landed
against `docs/superpowers/plans/2026-09-02-rung-1-foundation.md` and the
ratified build spec (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`,
revision 2.1): the three Open rulings that spec's review round raised
(shadow floor, oracle-bug policy, stock-config gating) were all resolved
by Geoff the same day, closing the item STATUS carried as "Open rulings
1-3 await Geoff."

Five tasks: the pivot commit tagged `pre-greenfield` and moved `cmd/`,
`internal/`, `web/`, and `e2e/` under `_salvage/`, freezing them
read-only, and stood up `cmd/dubplate-server` as the module's only Go
package; the ops audit removed `config/filebrowser.yaml` as an orphan
and made `scripts/deploy.sh`'s binary step additive (rsyncs to the new
`/usr/local/bin/dubplate-server` path, never touching the incumbent
`/usr/local/bin/dubplate` the four guardian units still exec); the
config core (`internal/config`) implemented the derived wall as code:
owned, reserved-for-a-rung, and superseded key classes, plus a
threshold registry, all contract-tested from a stub before
implementation; `check`
and `doctor` scaffolds gave the wall an observable surface with a pinned
line format; this ledger close reconciled ROADMAP, CLAUDE.md, and
`.claude/settings.json` against the pivoted tree.

What the gates caught and a later rung must not rediscover:
`scripts/vale-comments.sh`'s file-list union needed an explicit
`_salvage/` exclusion, because git's `*.go` pathspec matches at any
depth and would otherwise re-admit all 76 salvage Go files into the
prose-comment scan; `go mod tidy` pruned the module to cobra's set once
the salvage tree's other dependencies left the build; the additive
deploy boundary (new path beside the untouched incumbent binary) is the
mechanism that lets rung 1 land without any box write, and every rung
before the launch cutover (rung 16) must preserve it. No box access
occurred this rung, per the spec's box-safety section.

Post-close fix trail: task 4's review escalated a silent-error defect
(SilenceErrors set with no stderr print in main). Three commits closed
it under review—a531d69 (main prints the error once; check stops
double-printing), 1f3ef3f (doctor returns "configuration is
incomplete" instead of writing the raw load error), baac9f4 (the
no-usage-block contract gains a live test, mutation-proved)—and
db3d11b landed the conductor-dispatched code-simplifier pass. Two more
lessons for later rungs: cobra's usage writer falls back to os.Stderr
when only SetErr is wired, so command tests must wire SetOut to
observe a leaked usage block; and doctor's row expectations must
derive from config's own tables (owned, reserved, Thresholds()),
never a literal count. Budget: about 1.0M subagent tokens for the
workflow (12 agents, five tasks) plus about 0.3M across the fix and
verify dispatches, against the 5M ceiling; zero attended interactions
during execution. Verdict trail: T1-T3 accept first-pass; T5 accept
after one fix round; T4 escalate plus two fix rounds.

## Uploader Pass B close + the greenfield pivot (2026-09-01/02)

Pass B built and hardened the uploader web app end to end, then Geoff
superseded its launch with the greenfield dubplate-server ruling. What
landed in code (all on master, salvage tier): U9-U11 web app; the U13
review workflow (57 agents, 45 confirmed serious findings from 5
reviewers with adversarial verification) and its complete fold F1-F8
(TagProbe built and wired -- the probe never existed and no main-lane
upload could reach the library; serve security majors; reject-copy
contract; semantic theme tokens; App.svelte structure; component fixes;
saturated palette + soft surfaces; state-text lightness + rejected
CTA); a three-round visual-fidelity chain ending PASS; code-simplifier
consolidation (AlertCard, Store access discipline). Deployed to the box
2026-09-02 with "907 Music" minted, then wound down the same day.

What the gates caught that later work must not rediscover: DaisyUI 5
theme blocks define NO semantic color defaults for custom themes
(navidrome-* themes shipped without --color-error et al. and every
status surface rendered gray); DaisyUI 5 .alert is display:grid (flex
utilities are inert; single wrapper child required); -soft variants
color text with the raw token, so dark themes need Material-400-class
state tokens; the beets probe cannot use --pretend (it skips MusicBrainz
entirely) -- the real dry-run shape is in internal/beets/probe.go; the
box serve unit needed AF_NETLINK debate but not the family (glibc falls
back); find -newermt with a bare relative string fails silently in
guard scripts (use -mmin).

The pivot: Geoff's device check hit the shared-Access-org ASC branding
on the login page (a page never rendered before asking him -- now the
see-it-before-geoff-tests-it memory). The fix path escalated by
ruling into: magic-link login + Navidrome SSO (spec
2026-09-02-magic-link-login-design.md; two adversarial reviews in
docs/superpowers/research/2026-09-02-login-spec-reviews/ found 6
security blockers + 3 universality criticals -- all now birth
requirements), then the full greenfield ruling set
(docs/superpowers/research/2026-09-02-greenfield-server-rulings.md).
Interim estate wound down 2026-09-02: FileBrowser gone (container,
compose, env, DNS, Access app), uploader serve + nudge units disabled,
uploader tunnel rule + both Access apps deleted; box runs Navidrome +
cloudflared + guardian timers only. The 37 deferred U13 minors live in
docs/superpowers/research/2026-09-01-u13-findings.json and matter only
to salvaged components. The dubplate-organizer initiative spec was
brainstormed and four-review hardened the same nights (80f0f5f + later
amendments); the organizer spec + rulings doc are the greenfield
build's inputs.

## Uploader Pass A close: U12a live, gates 1–2 passed (2026-09-01)

The uploader went live at https://music.907.life/upload behind a
path-scoped Access app (Geoff relocated it mid-sitting from its own
hostname; the AUD survived the domain edit, so the minted secret held).
serve mounts its whole tree under /upload via StripPrefix with tusd's
BasePath threaded through Options.TusBasePath, because tusd bakes the
configured base verbatim into the absolute Location clients resume
against. Gates 1–2 were measured through the full production path by
logging in through Access with an emailed code read over JMAP and
driving fetch from the authenticated page: SSE heartbeats at exact 15s
cadence with no edge buffering, Last-Event-ID resume in 120ms, tus
offset persistence through an aborted PATCH. The gate found one defect
(tusd stamping http:// behind the tunnel; fixed same hour with
RespectForwardedHeaders).

What a later pass would be wrong to rediscover: Docker's host-gateway
magic resolves to docker0's gateway, never a named bridge's, so the
compose alias pins 172.30.0.1 literally; a compose network recreate
reattaches live containers WITHOUT service aliases (502'd Navidrome
until force-recreate; now a vps-conventions row); Access service tokens
pass Access but die at serve's roster-gated JWT check by design, so
machine verification goes through the email-code flow; the Access login
page wears the ASC org's default branding (Pass B fixes the app's own
name/logo). Pass B standing inputs live in
docs/superpowers/research/passb-{stack-brief,copy-deck}.md; the copy
deck carries seven pre-U9 revision items.

## Go rewrite close: box cutover and estate rename, T12–T14 (2026-08-31)

Geoff's no-users ruling collapsed the plan's staged migration (additive
env window, per-subcommand flips, soak-gated T13) into one evening: the
estate has a library worth protecting (backed up, verified) but no users
a botched cutover could hurt, so the windows bought safety nobody
needed. One continuous run: T12–T14 repo work through the chain (two
review rounds each; the reviewers caught a ReadWritePaths grant that
would have failed every backup — fixed by clearing staging CONTENTS so
the leaf grant holds — and a compose project-name float that would have
orphaned every container, fixed by pinning `name: musicbox`); one
deploy shipping the binary, nine units, and the 36-name
single-generation env at /etc/dubplate; doctor fully green first try;
all four jobs green on supervised first runs; old units/scripts/paths
removed from the box; pings checks re-registered as dubplate-* inside
the grace window (no false alert); `~/Projects/musicbox` →
`~/Projects/dubplate`. The soak runs post-hoc: overnight timers prove
unattended operation with the dead-man checks as witness.

What a later pass would be wrong to rediscover:

- Compose derives its project name from --project-directory's basename;
  moving the directory without pinning `name:` makes `up -d` orphan
  every running container and start duplicates. The pin carries a
  never-remove-without-migration comment in compose.yaml.
- A ReadWritePaths LEAF grant cannot survive os.RemoveAll on the
  granted directory itself (needs parent write). dubplate-backup clears
  contents instead; the unit comment carries the contract.
- Kept identifiers (Hetzner names, tunnel, compose project,
  /var/lib/musicbox, /srv/music) are documented in README; the estate
  rename is complete EXCEPT those, deliberately.
- `Paths.RunDir`/`DUBPLATE_RUN_DIR` is required by config.Load but read
  by nothing — dead config, escalated to the uploader pass's U1 rather
  than patched around the frozen inventory.
- Slug swap ordering that avoids false dead-man alerts: add new checks,
  re-mint URLs, redeploy env, prove one ping, THEN remove old slugs —
  all inside the shortest check's period+grace.

## Go rewrite, repo phase T1–T11 (2026-08-31)

The shell layer's replacement landed as one `dubplate` binary: sixteen
internal packages plus cmd and an e2e harness with stub beet/rclone/
sqlite3, all through the implementer/opus-diff-reviewer/gate chain via
two pass-execute workflows (wf_edc843fb, wf_048446d1) plus follow-up
dispatches. All eleven tasks accepted; consolidation added the
code-simplifier sweep, poplar's vale comment-prose gate (vendored
overlay, hard-fail), and tidy-check. Commits 6ec7aab → 4489c99. Spend
≈4.3M of the raised 5.5M ceiling (original 4M overrun approved by Geoff
at the 3.2M checkpoint); human interaction: plan approval (pre-session),
one combined overage/SMTP question, one per-step-sanction standing
grant. The chain caught real defects at every level: a config env-scan
panic, a vacuous TLS assertion, ping-fail exit semantics, two importer
shell-parity divergences, and a contract-wording regression (T5).

What a later pass would be wrong to rediscover:

- An implementer agent can DIE mid-task (API drop) leaving compiling but
  untested, unreviewed code; the workflow journal records nothing for it
  and neighbor tasks inherit red gates that look like their own
  failures. Signature: a task with no journal record + innocent
  escalations citing out-of-scope lint. Recovery: complete + review the
  orphan task explicitly; re-review the neighbors after.
- Dispatches carrying a reviewer's verbatim LOCATION/FINDING/FIX text as
  acceptance criteria went through at zero fix rounds; paraphrased
  criteria averaged one. Pre-extract findings, never summarize them.
- Rejection wording is per-extension via a path-aware builder inside
  RenderEmail (Friendly() cannot carry the extension; its signature is
  contract). The shell's ungrammatical "an WAV" is pinned deliberately.
- `internal/beets` carries `subsonicClient = "musicbox"` matching the
  live shell; it flips in T14's rename sweep, never earlier.
- review.Filer's dirs are direct destinations, not tree roots
  (documented at review.go:22): importer builds two differently
  configured Filers, one for filing, one for sweeping.
- The e2e ENOTDIR blocked-path fixture (a plain file where a dir is
  expected) is deliberate: chmod fixtures pass trivially under root.

A separate agent executed the bring-up runbook against the live box; the
close-out session verified steps 5–7 read-only from the box and committed
the outcomes. Seed result: nine ZIPs in, seven albums imported (Angel Du$t,
Bully, Cheekface, De La Soul, Nightmares on Wax, TV Girl, Wax Tailor — 7
albums / 105 items in the beets DB), two review-queued with reasons
(Aim "Hinterland (Remastered)", illuminati hotties "POWER") plus a TV Girl
zine PDF routed to review. Post-seed backup succeeded 18:06 UTC:
`library/` in R2 at 112 objects / 2.794 GiB, `state/` holding both DB
snapshots. Monitoring truth: import and backup success pings green; the
failure-ping path is confirmed dead on the box (curl 404 in both
`musicbox-ping-failure@` journals — the known pre-wave GET-/fail bug) and
the disk/navidrome check timers were never enabled by the pre-wave deploy;
both fixed by the interim deploy this entry accompanies.

What a later pass would be wrong to rediscover:

- Contributor ZIPs extract FLAT (loose `Artist - Album - NN Track.flac`),
  so beets needs `group_albums: yes`; without it the whole seed read as
  one 135-track candidate and skipped.
- Since beets 2.4, MusicBrainz is a metadata-source plugin: an explicit
  `plugins:` line without `musicbrainz` leaves the autotagger with zero
  candidate sources and every album skips.
- Hetzner Cloud blocks outbound 25 and 465 by default on new accounts;
  Fastmail submission works on 587 STARTTLS from the box.
- `safe_extract` needs bsdtar on the host; it was missing from cloud-init
  packages and broke the first seeded import (hand-installed on the box,
  cloud-init now carries it).

## Architecture refinement pass (2026-08-30/31)

Adversarial review of the estate as built, then a fix wave. Phase A: nine
finder lenses plus an opus verify wave over both repos, the spec chain, and
live Cloudflare state — 72 findings, 40 confirmed, 12 endorsed, 0 refuted
(ledger: ~/.dotfiles/docs/superpowers/plans/2026-08-30-musicbox-arch-refinement-findings.json).
Fix wave: 11 tasks + lock follow-ups through implementer/diff-reviewer/gate
chains, all accepted (commit 087df6c); pings hardened in its own repo
(40fdfa7). Spend ~6.2M tokens against a 6M ceiling (3% over, driven by an
uploader design brainstorm folded into the same session); human
interaction: one combined triage question, approvals otherwise.

What a later pass would be wrong to rediscover:

- A lock file in a systemd RuntimeDirectory is a lifecycle bug (deleted at
  unit stop, unlinked-inode races across units). Locks live in
  /var/lib/musicbox; per-run scratch keeps /run. Doctrine now in the
  vps-conventions skill — invoke it before touching units.
- ping_hc fail is POST; the Worker 404s GET /fail. Every failure ping ever
  sent before this pass was dropped.
- bash cannot re-arm errexit inside a command-substitution subshell; the
  M3 first attempt guarded notify() with a subshell set -e that did
  nothing. Best-effort is enforced by if-guarding at the definition.
- The beets database is the only import success record (tag log = skips
  only; beets renames everything). The uploader design builds on this.
- Review standard that caught real bugs: verify test non-vacuity by
  mutation (M10's reviewer), and re-run implementer claims yourself
  (M2b's reviewer caught a misdescribed unit).
- The estate renamed musicbox → dubplate (2026-08-31); repo/paths converge
  at rewrite T14. The rewrite, uploader, and packaging specs were authored
  and adversarially reviewed this same session.

Carried out of the pass: SELinux one-time check on /var/lib/musicbox at
first deploy; provision.sh's ssh_cidrs_json exits only its substitution
subshell on malformed input (diagnostic degrades to a jq error); pings
remove-check has no --force for a permanently orphaned route record.

## Spec gaps closed: rclone/msmtp provisioning, Subsonic scan trigger (2026-08-31)

Closed the two real spec gaps the first live deploy exposed (previous entry below),
without touching the live box.

- **`rclone`/`msmtp` were never in `cloud-init/user-data.yaml` at all.** Confirmed both
  are in EPEL for AlmaLinux 10 via a container check
  (`podman run --rm almalinux:10 ...dnf -q list rclone msmtp`) before adding them,
  rather than guessing an unattended `set -e` first-boot script into a package that
  doesn't exist: `rclone.x86_64 1.74.3-1.el10_2` and `msmtp.x86_64 1.8.25-6.el10_1`,
  both from the `epel` repo alone, no RPM Fusion or CRB dependency. Installed
  immediately after `epel-release` in the existing EPEL/RPM-Fusion runcmd block,
  ahead of the RPM Fusion key import flac/ffmpeg need, so a failed RPM Fusion step
  can't take them down too.
- **`trigger_navidrome_scan()` rewritten from the dead native `/api/scanner/scan`
  endpoint to the Subsonic API's `GET /rest/startScan`** (query-param auth:
  `u`/`t`/`s`/`v`/`c`, `t=md5(password+salt)`, a fresh random `s` per call). The
  credential-bearing params never touch argv: the full URL is written to a curl
  `--config` file (mode 0600) in a new `RUNTIME_DIR` (same directory as the import
  lock file -- `/run/musicbox` on the box, the fixture's `WORK_DIR` in tests), and
  the config file is removed via a function-scoped `RETURN` trap. Salt is generated
  via `od -An -N8 -tx1 /dev/urandom | tr -d ' \n'` rather than a `tr | head`
  pipeline, which would risk a SIGPIPE-driven pipefail failure. The existing
  failure-tolerant shape is unchanged: a failed request logs to stderr and returns
  0, never killing the run (the daily Navidrome cron scan is still the backstop).
  `tests/music-import.bats`'s shared curl stub now also captures `--config` file
  contents to a separate log (the real file is gone by the time a test could
  inspect it otherwise), plus a `CURL_FAIL_STARTSCAN` switch to exercise the
  failure path; two new tests cover the call shape and the failure tolerance.

Neither fix has reached the already-provisioned live box yet (cloud-init is
first-boot-only); see STATUS's "reprovision pending" open item.

## Task 4/6: first live deploy (2026-08-31)

`scripts/deploy.sh --check` then a real run against the provisioned box. Two env-template
vars didn't resolve as assumed going in: `NAVIDROME_ADMIN_PASSWORD` was stored under the
older name `NAVIDROME_ADMIN_PASS` (renamed in the age store to match, same value, no
rotation), and `MUSICBOX_HC_NAVIDROME_URL` had never been minted at all. Minted it via
`~/Projects/pings/scripts/add-check musicbox-navidrome --period 60 --grace 15` (mirroring
`musicbox-disk`'s already-accepted pre-minted-but-unwired state) so the render gate could
pass; it sits armed and unpinged until Task 6 builds the Navidrome-heartbeat script.

The deploy itself succeeded, but the acceptance sweep surfaced real bugs a fixture-backed
test suite can't see — **wrong to rediscover**, each one below:

1. **FileBrowser crash-looped**: `cacheDir` defaults to a `tmp` folder relative to the
   process's working directory, which isn't part of any bind mount and isn't writable by
   UID 2000. Fixed by pointing `server.cacheDir` at the already-writable, already-persistent
   `/home/filebrowser/data/cache` (`config/filebrowser.yaml`).
2. **FileBrowser's own first-run admin bootstrap then failed**: its built-in default admin
   password is shorter than our own `auth.methods.password.minLength: 8` policy, so account
   creation errored and the container kept crash-looping. Minted a new secret
   `FILEBROWSER_ADMIN_PASSWORD` (FileBrowser's native env var name) and wired it into
   `compose.yaml`'s `filebrowser` service; Task 8 replaces this bootstrap account with
   per-contributor ones.
3. **music-import.service failed on `/run/lock/musicbox-import.lock: Permission denied`**:
   `/run/lock` is root:root `0755` on Alma (unlike Debian's world-writable `1777`), so the
   `User=music` service couldn't create its flock file there. Added
   `RuntimeDirectory=musicbox` to the unit (systemd creates `/run/musicbox` owned by
   `music:music` every boot) and pointed the script's lock-file default there.
4. **music-import then failed on `curl: (7) ... Failed to connect to localhost port 4533`**:
   `trigger_navidrome_scan()` runs on the host as `music`, not inside the compose network,
   but `compose.yaml`'s `navidrome` service only used `expose` (container-network-only).
   Switched to `ports: - "127.0.0.1:4533:4533"` — loopback-only, so Navidrome stays exactly
   as unreachable from the internet as before (still only `cloudflared` reaches it
   externally), while becoming reachable from the box itself.
5. **Same call's error handling was also incomplete**: the login `curl` had no failure
   guard, so `set -e`/`pipefail` killed the *entire* import run on any Navidrome-unreachable
   condition, not just the graceful "skip the scan trigger" path the function's comments
   already described for a failed login. Added `|| true` to match.
6. **The native admin scan-trigger endpoint doesn't exist in the pinned Navidrome
   `0.63.2`**: confirmed against upstream source (`server/subsonic/library_scanning.go`,
   `server/subsonic/api.go` line ~181) — `StartScan` is registered as the *Subsonic* API's
   admin-only `startScan` action (query-param auth: `u`/`p`/`v`/`c`), not a native
   `/api/scanner/scan` JWT-bearer endpoint. Every `music-import` run now logs a 404 there
   after a real login success. **Not fixed this pass** — left as a known, non-fatal bug
   (Navidrome's own `ND_SCANNER_SCHEDULE: "0 3 * * *"` cron backstop still keeps the library
   current) rather than rush an unverified Subsonic-auth rewrite. Follow-up: rewrite
   `trigger_navidrome_scan()` to call `GET /rest/startScan` with Subsonic auth params.

**Two gaps could not be closed this pass, both blocked by the Claude Code auto-mode
classifier refusing an on-box `dnf install` from a subagent** (same shape as Task 5's
DELETE/create-tunnel refusal above — mutating infra calls need Geoff present, not a
subagent acting alone):

- **`ffmpeg`/`flac` are still absent on this box.** The EPEL + RPM Fusion `runcmd` block
  that installs them (`cloud-init/user-data.yaml`) was added to the repo *after* this box
  was provisioned (Task 4a), so it never ran here — a cloud-init-only fix never converges
  an already-running box. beets' `replaygain` plugin fails to load on every `music-import`
  invocation as a result (logged, non-fatal against an empty inbox; will break real
  imports needing the quality gate's `flac -t`/`metaflac` and replaygain analysis).
- **`music-backup.service` cannot run at all**: `sqlite3` and `rclone` are both missing on
  the box, and neither was ever listed in `cloud-init/user-data.yaml` in the first place (a
  real spec gap, not just an unapplied one). Added `sqlite` to the `packages:` list (base
  AlmaLinux repo, confirmed by lookup) for future re-provisions; `rclone` and `msmtp` (also
  missing, also never listed — `lib.sh`'s `notify()` needs it, untested this pass since an
  empty-inbox run never reaches it) are left as an open item rather than guessed into an
  unattended `set -e` first-boot script without a confirmed AlmaLinux 10 install path.
  Confirmed via a read-only `rclone lsjson` against `music-library` that `state/` is empty,
  consistent with no backup ever having run.

Acceptance sweep otherwise green: all three containers running with zero restarts after a
60s settle; the tunnel reads `status: healthy` with active connections; `music.907.life`
serves Navidrome's web UI with no Access interstitial and answers `/rest/ping`;
`inbox.907.life` presents the Cloudflare Access challenge before any FileBrowser content;
both timers active; the Navidrome admin account was created via `POST /auth/createAdmin`
and its login verified via `POST /auth/login` (API only, no secret values ever printed —
checked by name, length, and boolean fields throughout this whole pass).

## Task 5: Cloudflare tunnel wiring, plus a same-day secret rotation (2026-08-31)

Geoff added Cloudflare Tunnel:Edit to `CLOUDFLARE_API_TOKEN` 2026-08-30, unblocking Task 5.
An agent created tunnel `musicbox` (`POST .../cfd_tunnel`), then displayed that call's full
JSON response in a transcript while reporting progress. **Wrong to rediscover**: a
`cfd_tunnel` CREATE response body carries the tunnel's `TunnelSecret` and its connector
`token` alongside the id — displaying or logging the raw body exposes both, not just an
identifier. The fix at the call site is mechanical: always pipe the response through
`jq -r '.result.id'` (or whichever single field is actually needed) and never let the full
body reach a transcript, a log, or a report.

Once flagged, the tunnel was treated as compromised under standing incident doctrine (a
secret printed into a transcript is compromised, full stop — no risk-assessment exception).
Remediation was delete-and-recreate, the cheap definitive path since `cloudflared` was not
yet deployed anywhere against the exposed token:

1. Deleted tunnel `66fa450c-5abb-459d-a1f8-b80d8d19e07c` (`DELETE .../cfd_tunnel/{id}`).
2. Created a new tunnel under the same name `musicbox`, this time piping the response
   through `jq -r '.result.id'` only: id `b6781175-000b-40c3-a54a-05906ab25037`.
3. Fetched its token (`GET .../cfd_tunnel/{id}/token`) and piped it directly into
   `secret-set.sh MUSICBOX_TUNNEL_TOKEN --stdin`, overwriting the old value. Verified
   name-only (`grep -c` on `~/.local/secrets`, count 1) and by length only (240 chars) —
   never printed.
4. Re-PUT the ingress from `config/cloudflared-tunnel-config.json` to the new tunnel id and
   verified with GET: `music.907.life` → `http://navidrome:4533`, `inbox.907.life` →
   `http://filebrowser:80`, 404 catch-all.
5. Repointed both DNS CNAMEs (`music.907.life`, `inbox.907.life`, zone `907.life`) to
   `b6781175-000b-40c3-a54a-05906ab25037.cfargotunnel.com`, proxied.

**Process note**: the DELETE and the tunnel-recreate POST were both refused by the Claude
Code auto-mode classifier when a subagent attempted them (destructive/mutating infra calls
without the user present); read-only GETs against the same account worked fine throughout.
The rotation completed once the conductor ran those two calls from the main session with
Geoff present, then handed the new tunnel id back for the remaining wiring and docs.

Cloudflare Access was untouched throughout (confirmed read-only via `CF_ZT_TOKEN`:
`inbox.907.life` is the only Access app on the musicbox surface; `music.907.life` stays
Access-free). `cloudflared` is not deployed on the box yet, so both hostnames 530 until
Task 4's first real `deploy.sh` — expected, not chased.

## Task 2 follow-up: quality gate, ping hardening, config sync (2026-08-30)

TDD, red-first for every case, against the spec's freshly amended import-pipeline
section (step 4). Four items, one commit each:

- **`quality_gate <dir>`** (`scripts/lib.sh`): walks every audio file under dir
  (case-insensitive flac/mp3/m4a/ogg/opus/wav/aiff extensions; anything else is out
  of scope). A non-FLAC extension fails outright; a FLAC must pass `flac -t`'s full
  decode test and carry >=44.1kHz/16-bit (`metaflac --show-sample-rate
  --show-bps`). Prints "path\treason" per failure, exit 0 only if everything
  passed. Wired into `music-import` as `quarantine_quality_failures`, between the
  free-space abort and `beet import`: any failing file takes its **whole**
  containing album directory to `review/<date>/<user>/` (a partial import of just
  the good tracks never happens), never just the offending file.
- **Uploader copy is templated, not generated**: `config/reject-email.txt` is
  Geoff's verbatim consumer-facing copy with three literal placeholders (`{first
  name}`, `{album or file name}`, `{one reason line per file, from the mapping
  below}`), filled by plain bash substring replacement in
  `render_reject_email()` — no sed, no metacharacter risk. `friendly_reason()`
  translates `quality_gate`'s raw technical reason into one of three fixed human
  sentences (non-FLAC extension, failed decode, below CD quality); Geoff's own
  copy of the notice stays technical (raw path + reason), same asymmetry as the
  existing beets-skip notice. First name comes from `uploader_name()`, the first
  word of the matched `contributors.yaml` `name` field (mirrors `uploader_email`'s
  lookup). Default template path is `/opt/musicbox/config/reject-email.txt` —
  `config/` is already rsynced wholesale to that tree by `deploy.sh`, so this
  needed no new deploy step; only `contributors.yaml` and `beets-config.yaml` get
  dedicated ones, for permissions/install-location reasons this file doesn't
  share.
- **`ping_hc` hardened**: `--retry 3 --max-time 10` added to its `curl` call, so a
  slow or flaky healthchecks.io can never be the reason an otherwise-successful
  run hangs or gets misreported. The empty-vs-unset URL guard already existed;
  only a new bats case (env var set but empty) was added to pin it down.
- **Flock-held early exit now pings success before `exit 0`** (spec decision):
  a run that finds the lock already held (another scheduled run, or an
  interactive `music-beet` session) is the healthy case this check exists for,
  not a failure — healthchecks.io has no third state between success and
  failure. Accepted residual: a genuinely wedged import would also report
  healthy via this path until someone notices stale music in the library.
- **Beets config synced to the workstation's** (`~/.dotfiles/beets/.config/beets/config.yaml`):
  `original_date: yes`, `per_disc_numbering: yes`, and the `replaygain` plugin
  (backend `ffmpeg`), added to `config/beets-config.yaml` alongside the existing
  quiet-mode section. flac (the quality gate) and ffmpeg (replaygain's analysis
  backend) are installed on the host via a dedicated `cloud-init/user-data.yaml`
  runcmd step, not the top-level `packages:` list (that module runs before
  runcmd, too early to enable a repo first): `dnf -y install epel-release`,
  `dnf config-manager --set-enabled crb` (both EPEL and RPM Fusion need it),
  then the RPM Fusion free EL10 GPG key is fetched and `rpm --import`ed and
  its release package installed against it (gpgcheck on throughout, no
  `--nogpgcheck`), then `dnf install -y flac ffmpeg`. Ruling: both binaries
  parse untrusted contributor uploads, so they stay on dnf-automatic's
  security-update path rather than a pinned static build, which would freeze
  a codec-parsing attack surface instead of patching it.
- **Known limitation, accepted**: `quality_gate`'s output channel is
  tab/newline-delimited ("path\treason" per line), so a filename containing a
  literal tab or newline would misparse downstream in
  `quarantine_quality_failures`. The `find -type f -print0` walk itself is
  correct (NUL-delimited, handles any filename); the limitation is only in
  what quality_gate prints afterward. Not expected in practice for
  contributor uploads; revisit if it ever is.

Fixed a docs nit found while touching STATUS.md: Task 4a's on-box checks table
(above, Task 4a's own section) lists nine checks but was labeled "8 of 8"; now "9
of 9" in both STATUS.md and here.

Full test suite: `tests/lib.bats` (4 new `quality_gate` cases + 1 `ping_hc` case)
and `tests/music-import.bats` (1 new flock-ping case, 1 new quality-gate
end-to-end case; three existing fixtures switched from placeholder `.flac` text to
real ffmpeg+flac-encoded audio, since the gate now actually decodes them).
`bash scripts/check.sh` exits 0.

## Task 2 follow-up, fix round: Opus review of 5da3844..d43972b (2026-08-30)

Six items landed, TDD where a test was named:

- **Blocking: unquoted `&` in `render_reject_email`'s `${var//pat/repl}`
  substitutions.** Confirmed on this box's bash (5.3.9): an unquoted
  replacement containing a literal `&` gets expanded to the matched pattern
  text, exactly like sed — an album named "Sturm & Drang" rendered as "Sturm
  {album or file name} Drang" instead of the real name. Fixed by quoting
  every replacement value; bats case covers the ampersand literally.
- **Blocking: an `mv` into an already-occupied `review/<date>/<user>/<album>`
  aborted the whole run under `set -e`** (a same-day repeat quarantine of an
  identically named album is the common case, not an edge case). New
  `lib.sh` helpers: `unique_dest_path` (appends a `-HHMMSS` suffix on
  collision, then `.2`/`.3`/... if even that collides) and
  `safe_mv_to_review` (uses it, and warns-and-returns-nonzero instead of
  letting a failed move take the run down). Applied to both
  `quarantine_quality_failures` and `move_skips_and_notify`, which shared the
  same unguarded-`mv` shape. bats case: two identically named "Bad Album"
  quarantines in one test run both survive and beet import still runs both
  times.
- **A missing `reject-email.txt` template previously sent the uploader an
  empty-bodied email** (`render_reject_email`'s failure was silently
  swallowed by command substitution). Now an explicit `[ -f ... ]` check
  before the `notify` call skips the uploader-facing email and logs why;
  Geoff's technical summary still covers the album either way.
- **Each quality-gate reason line is now prefixed with its file's
  basename** in the uploader-facing email, so an album with more than one
  failing file reads as N per-file lines instead of running together.
- **Added the 8-bit-FLAC branch's missing test coverage** (`ffmpeg -acodec
  pcm_u8` piped into `flac`), the one `quality_gate` branch that had no
  case before this round.
- **CRB enablement** added to cloud-init's EPEL/RPM Fusion runcmd block
  (`dnf config-manager --set-enabled crb`, right after `epel-release`):
  both repos need packages CRB carries, and ffmpeg's dependency chain fails
  without it.

Follow-up verification round found one more leak in the `mv`-collision fix:
`safe_mv_to_review` failing for a **non-collision** reason (permissions, a
bad mount) still `continue`d out of the loop with no Geoff notice and the
album still sitting in staging — exactly what `beet import`'s unconditional
recursive scan of `$STAGING_DIR` would then ingest, silently defeating the
whole quality gate. Fixed with a second `safe_mv_to_review` attempt to
`$MUSICBOX_ROOT/quarantine-failed/<date>/<user>`, a holding dir outside both
staging and review's own tree (so a problem isolated to `review_root`
doesn't also block the fallback); Geoff's summary gets a `could not move;
excluded from import` marker either way, and a distinct all-caps marker for
the (untested, believed-unreachable without a fully read-only filesystem)
case where even the fallback fails. The holding dir's root is a top-level
constant (`QUARANTINE_FAILED_DIR`) beside `REVIEW_DIR` and the rest of the
tree, deliberately not created up front: a healthy box never has one.

Also fixed in `safe_mv_to_review` itself: its `mkdir -p "$dest_dir"` call
was unguarded, so a destination that couldn't even be created would have
aborted the whole run under `set -e` — the exact failure mode the function
exists to prevent. Added a matching
`[ -n "$geoff_summary" ] || return 0` guard before the final Geoff `notify`,
defense-in-depth alongside the pre-loop `album_order` check (every branch of
the loop already appends something, so this is currently unreachable, but
decouples "should Geoff be emailed" from "was there at least one album_key"
for whatever a future change to the loop might do). TDD: bats case forces
the primary move to fail via `chmod 0555` on the review dir and asserts the
album lands in `quarantine-failed`, never in staging, with the marker in
Geoff's email and no uploader email.

Also fixed, while adding coverage for the above: three existing
`tests/music-import.bats` assertions used `! grep ...` as a non-final line —
under bats' `set -e`, a command negated with a leading `!` never triggers
errexit regardless of its actual result, so a failing negated assertion that
isn't the test's last line silently does nothing. Converted to bats' `run` +
`[ "$status" -ne 0 ]` idiom throughout the file for safety, whether or not a
given instance was actually latent.

## Task 5/6: Cloudflare Access + Navidrome secrets, tunnel blocked (2026-08-30)

Control-plane pass, no box changes. Findings and what landed:

- **Tunnel creation blocked, not a code problem**: `POST
  /accounts/{a}/cfd_tunnel` 403s `{"code":10000,"message":"Authentication
  error"}` on both `CLOUDFLARE_API_TOKEN` and `CF_ZT_TOKEN` — neither carries the
  Cloudflare Tunnel:Edit permission group (confirmed: both tokens list/read
  tunnels fine, i.e. GET `.../cfd_tunnel` is 200; only the POST 403s). This
  blocks Task 5's steps 1-4 (create tunnel, fetch token, PUT ingress config, DNS
  CNAMEs to `<tunnel-id>.cfargotunnel.com`) and, downstream, Task 4's first real
  `deploy.sh` run (cloudflared needs `MUSICBOX_TUNNEL_TOKEN` to come up).
  Resolution needs Geoff to add Cloudflare Tunnel:Edit to one token via the
  dashboard (API self-management of token scopes is deliberately locked out).
  Detail in registry.md's `CF_ZT_TOKEN` entry.
- **Cloudflare Access, done**: confirmed no Access application exists on any
  `907.life` hostname (checked the full account app list — zero matches), so
  `music.907.life` correctly has none. Created a self-hosted Access application
  for `inbox.907.life` (`CF_ZT_TOKEN`, which — unlike `CLOUDFLARE_API_TOKEN` —
  can read Access identity providers) restricted to the account's existing
  `onetimepin` identity provider (`e33c7c2d-facd-446c-96de-8215f64cde76`,
  `auto_redirect_to_identity: true` since it's the only allowed IdP), with an
  allow policy for `geoff-login@907.life` (policy
  `92425260-87f6-4a3a-b66e-d71bb8d9d6a8`). Family emails join this policy at
  Task 8 per plan. **Not yet verified end to end**: the "Access challenge
  appears before FileBrowser's login" acceptance check needs the tunnel serving
  traffic first, so it waits on the tunnel-permission fix above.
- **Task 6 Navidrome secrets, done**: generated and stored
  `ND_PASSWORDENCRYPTIONKEY` (openssl rand -hex 32), `NAVIDROME_ADMIN_PASS`
  (openssl rand -hex 16), `NAVIDROME_ADMIN_USER` (`admin`) via `secret-set.sh`.
  No value printed; lengths only (64/32/5 chars) confirmed against
  `~/.local/secrets`.
- **registry.md, done**: added rows + per-secret sections for
  `MUSICBOX_TUNNEL_TOKEN` (marked not-yet-minted with the blocker noted),
  `ND_PASSWORDENCRYPTIONKEY`, `NAVIDROME_ADMIN_USER/PASS`, and the four R2 pairs
  already in the store (`MUSICBOX_R2_*` — musicbox-vps token, Object Read &
  Write, bucket `music-library`; `MUSIC_R2_RO_*` — claude-code-thinkpad-x1
  token, Object Read only, same bucket); noted `HCLOUD_TOKEN`'s 2026-08-30
  rotation. Committed to `~/.dotfiles` as two commits (registry.md text, then
  `values.age` separately) — a third file in the working tree,
  `bin/.local/bin/secret-receive`, was mid-edit from an unrelated session and
  deliberately left untouched/uncommitted.

Decision a later pass would be wrong to rediscover from scratch: **use
`CF_ZT_TOKEN`, not `CLOUDFLARE_API_TOKEN`, for anything touching Access
identity providers** — the general-purpose token gets `auth.forbidden` reading
`.../access/identity_providers`, even though it can read/write Access apps and
policies fine.

## Task 4a: AlmaLinux 10 image-defect fix round, rebuild to 8/8 green (2026-08-31)

Two deterministic image defects blocked the recreated box at 5/8 green (see the prior
round's entry below): Docker's missing `xt_addrtype` kernel module
([docker/docker-ce-packaging#1340](https://github.com/docker/docker-ce-packaging/issues/1340))
and the Hetzner `alma-10` image shipping `SELINUX=permissive` instead of the distro
default of enforcing. Both fixes landed in `cloud-init/user-data.yaml`'s `runcmd` plus a
new top-level `power_state` key:

- **Docker/kernel-module fix**: a new `runcmd` item installs
  `kernel-modules-extra-$(uname -r)` — the exact NVR paired to the boot's running
  kernel — before the Docker CE install, falling back to `dnf update kernel &&  dnf
  install kernel-modules-extra` only if the mirrors have moved past this image's shipped
  kernel (that fallback needs the reboot below to take effect). `systemctl enable --now
  docker` became `systemctl enable docker` — starting it before the reboot would still run
  under the old kernel and fail identically.
- **SELinux fix**: `sed` flips `/etc/selinux/config` to `SELINUX=enforcing` and `touch
  /.autorelabel` schedules a full relabel on next boot (the safe path — `setenforce 1`
  alone would trust every file on a vendor image already carries correct labels, which
  permissive mode never enforced).
- **Reboot ordering**: a single `power_state: {mode: reboot}` module — not a `reboot` call
  inside `runcmd` — runs after every other cloud-init module completes and has its
  once-per-instance completion recorded, covering both fixes in one reboot. The
  autorelabel triggers its own second, self-contained reboot; the acceptance run must
  tolerate SSH dropping twice.

First attempt at this round left the fix uncommitted while re-verifying against the live
box — caught by a review pass that diffed the deployed `/var/lib/cloud/instance/user-data.txt`
sha256 against `HEAD`'s copy of `cloud-init/user-data.yaml` and found them different
(`196e76f...` deployed vs `db0b633...` at HEAD). Lesson for any future fix-round: verifying
green against a live box proves nothing about the repo until the fix is committed and that
commit's own blob hash is re-checked against the deployed file — a worktree edit and an
unpushed commit look identical to `ssh`.

Re-verified after commit `3fa905d`: `sha256sum cloud-init/user-data.yaml` at HEAD equals
the deployed `/var/lib/cloud/instance/user-data.txt` hash
(`196e76fbc51d5a5828c6a378761d480b790160cc9fd07fa7a1f41ab1557dfa5a`). `bash
scripts/check.sh` exits 0.

**9 of 9 on-box checks green, re-verified 2026-08-31 against `root@5.78.148.42`:**

| Check | Command | Result |
|---|---|---|
| cloud-init | `cloud-init status --wait` | `status: done` — PASS |
| volume mount | `mountpoint -q /srv/music` | mounted — PASS |
| tree ownership | `ls -lnd /srv/music/{library,inbox,review,state,staging}` | all `2000 2000`; `inbox` is `drwxrwsr-x` (2775, setgid) — PASS |
| music user | `id music` | `uid=2000(music) gid=2000(music) groups=2000(music)` — PASS |
| docker | `systemctl is-active docker` | `active` — PASS |
| SELinux | `getenforce` | `Enforcing` — PASS |
| sshd password auth | `sshd -T \| grep -i '^passwordauthentication'` | `passwordauthentication no` — PASS |
| yq / jq | `/usr/local/bin/yq --version` / `jq --version` | `v4.53.6` / present (`jq-` is a cosmetic upstream version-string quirk on this AlmaLinux 10 build, not a functional defect) — PASS |
| fstab | `grep srv/music /etc/fstab` | exactly one line: `UUID=6614e3dd-9014-4949-9f4c-effcd317bd3a /srv/music ext4 defaults,nofail,x-systemd.device-timeout=30s 0 2` — PASS |

## Task 4a: real provisioning, forced-root-password fix round (2026-08-30/31)

`scripts/provision.sh --check` matched expectations (alma-10 available, no fallback
needed), then `scripts/provision.sh` created firewall `musicbox-fw` (tcp/22 only, both
address families), volume `musicbox-data` (50GB, `hil`), and server `musicbox` (CPX21,
`alma-10`, `hil`). Server IPv4 `5.78.148.42`, IPv6 `2a01:4ff:1f0:97c6::1`. CPX21 traffic
allowance at `hil`: 2199023255552 bytes = 2 TiB/month (`hcloud server-type describe cpx21
-o json`), overage $1.20/TB — generous for a family Navidrome, no need to default clients
to transcoded streams.

A fix round closed three blocking findings from a forced-root-password blocker, and the
server was recreated to pick them up:

1. `scripts/provision.sh` registers an `hcloud ssh-key` named `musicbox-workstation`
   idempotently and passes `--ssh-key musicbox-workstation` to `hcloud server create`.
   Confirmed on the recreate: no "Root password: ..." line appeared in the create output.
2. `cloud-init/user-data.yaml`'s `runcmd` carries a defense-in-depth guard — `chage -M -1
   -I -1 -E -1 -d "$(date +%Y-%m-%d)" root` — as the first item, so any future path that
   creates the server without a key can't reintroduce this blocker. Confirmed on-box:
   `chage -l root` shows `Password expires: never`.
3. The prior round's one-time-manual-SSH-unblock instruction was removed; nothing was
   hand-applied to a live box — the box was deleted and recreated from the fixed repo
   instead (volume `musicbox-data` survived and reattached, `blkid` guard prevented
   reformat, box got IPv4 `5.78.148.42` again).

Re-verification after that round found a new, unrelated blocker: `cloud-init status
--wait` returned `status: error` because Docker failed to start
(`docker/docker-ce-packaging#1340`, see the entry above), which also aborted the
downstream yq-install `runcmd` item. `getenforce` also reported `Permissive`, traced to
the Hetzner `alma-10` image's own shipped default, not to anything cloud-init touched.
Both became the next fix round, entry above.

## Task 3: service configuration — compose, systemd units, deploy (2026-08-30)

Landed `compose.yaml` (Navidrome, FileBrowser Quantum, cloudflared — no service
publishes a host port; only cloudflared is reachable from outside), `config/
navidrome.toml`, `config/filebrowser.yaml`, `config/msmtprc`, `config/beets-config.yaml`,
`config/cloudflared-tunnel-config.json`, the four `systemd/music-{import,backup}.
{service,timer}` units plus a shared `systemd/musicbox-ping-failure@.service` template,
`env/musicbox.env.template`, and `scripts/deploy.sh`. Extended `scripts/check.sh` with
`docker compose config -q` and a pattern-filtered `systemd-analyze verify`. All pinned
tags looked up and dated in `docs/STATUS.md`.

Two Task 2/Task 3 interface mismatches found and fixed (Task 3's own acceptance criteria
named the authoritative env var scheme, so Task 2's code was the thing that moved):
- **healthchecks.io slugs**: `music-import`/`music-backup` called `ping_hc
  musicbox-import` / `ping_hc musicbox-backup`, which resolves (per `lib.sh`'s own
  uppercase-and-dash-to-underscore rule) to `MUSICBOX_HC_MUSICBOX_IMPORT_URL` /
  `MUSICBOX_HC_MUSICBOX_BACKUP_URL` — doubled "MUSICBOX", not the
  `MUSICBOX_HC_IMPORT_URL` / `MUSICBOX_HC_BACKUP_URL` this task's env template requires.
  Fixed both scripts to call `ping_hc import` / `ping_hc backup`; updated
  `tests/lib.bats` and `tests/music-import.bats` to match (still 16 green). Monitoring
  would otherwise have silently no-op'd forever — `ping_hc` treats an unset URL var as a
  no-op, not an error, precisely so a monitoring bug never takes down the pipeline it
  monitors, which also means a naming bug like this one would never have surfaced as a
  failure anywhere.
- **Navidrome scan-trigger credentials**: `music-import`'s `trigger_navidrome_scan()`
  read `NAVIDROME_USER`/`NAVIDROME_PASSWORD`; this task's env template is required to
  name `NAVIDROME_ADMIN_USER`/`NAVIDROME_ADMIN_PASSWORD`. Renamed in the script and in
  `tests/music-import.bats`.

Two review findings fixed after the first pass (both would have broken the first real
deploy, not just this repo's own gate):
- **`/etc/musicbox/env`'s parent directory didn't exist.** Nothing in `provision.sh` or
  `cloud-init/user-data.yaml` creates `/etc/musicbox` (only `/srv/music` is built at
  provision time), so `install -m 600 ... /etc/musicbox/env` died with "cannot create
  regular file" on a fresh box, after the scripts and units had already landed. Fixed by
  adding `-D` to that `install` call, which creates the missing parent and is a no-op on
  every later re-deploy.
- **`config/msmtprc` was never actually installed to `/etc/msmtprc`.** The file's own
  header already claimed deploy.sh put it there, but deploy.sh only rsynced `config/` to
  `/opt/musicbox/config` for the record — `lib.sh`'s `notify()` calls plain `msmtp`, which
  resolves its account from `/etc/msmtprc` and would have found nothing there, aborting
  `music-import` under `set -euo pipefail` on the first skipped-track notification. Added
  an explicit install step (rsync to a temp path, then `install -m 0644` to
  `/etc/msmtprc`), with a matching diff-and-report line in `--check`.

Decisions a later pass would be wrong to rediscover from scratch:
- **Navidrome config split**: non-secret baseline settings (`MusicFolder`, `DataFolder`,
  `LogLevel`) live in `config/navidrome.toml`, bind-mounted read-only at
  `/data/navidrome.toml` alongside the read-write `/data` mount for the DB — Docker lets
  a single-file bind mount overlay a path inside an already-mounted directory. Scanner
  settings and the encryption key live in `compose.yaml`'s `environment:` instead (ND_
  env vars override the toml), specifically so the two option names the task called out
  by name (`ND_SCANNER_SCHEDULE`, `ND_SCANNER_WATCHERWAIT`) are literally grep-able in
  compose.yaml. `ND_CONFIGFILE=/data/navidrome.toml` is set explicitly rather than relied
  on implicitly, per navidrome.org's own Docker install guidance.
- **FileBrowser Quantum 1.5.x has no static per-user YAML config** — verified against the
  actual `v1.5.5-stable` tag's Go source (`backend/common/settings/structs.go`), not just
  the docs site (which mixes in v2.0 beta content). Users and their scopes live in its own
  SQLite DB, created via CLI/API, not this repo. `config/filebrowser.yaml` instead
  configures one source (`/srv/music/inbox`) with `createUserDir: true` +
  `defaultUserScope: "/"`, so creating an account named `alice` automatically scopes her
  to exactly `/srv/music/inbox/alice` — no per-user YAML needed, and the username IS the
  inbox directory name music-import already matches against `contributors.yaml`. Task 8
  creates the real accounts (one per `contributor`-role entry) once Task 0 lands real
  names.
- **cloudflared runs `tunnel --no-autoupdate run` with no local `--config`.** A
  token-driven tunnel is remotely managed; a local ingress file is silently ignored
  (verified against upstream `cloudflared` issues, not assumed). The actual ingress
  mapping is `config/cloudflared-tunnel-config.json`, a ready-to-PUT API body for Task 5
  to apply once the named tunnel exists — see STATUS.md's Open items, this is a real gap
  Task 5 must not skip.
- **msmtp's `passwordeval` reads `$MUSICBOX_SMTP_PASSWORD` from its own inherited process
  environment** (`printenv MUSICBOX_SMTP_PASSWORD`), not by re-reading
  `/etc/musicbox/env` from disk. That variable is already present in music-import's
  environment via its unit's `EnvironmentFile=`; having msmtp open the file a second time
  would require it group-readable by the unprivileged `music` user, widening every OTHER
  secret in that one file to the same process for no actual reduction in blast radius
  (that process already receives all of them via `EnvironmentFile=`). `config/msmtprc`
  itself holds no secret and installs world-readable at `/etc/msmtprc`.
- **Deploy target directory**: `deploy.sh` rsyncs `compose.yaml` + `config/` to
  `/opt/musicbox` on the box (not `/srv/music`, which is data, restored from R2, and
  never holds repo-sourced config). `docker compose` runs with `--project-directory
  /opt/musicbox` so the compose file's relative `./config/...` volume paths resolve
  regardless of the caller's cwd.
- **`deploy.sh --check` never prints secret contents**, even when diffing the rendered
  `/etc/musicbox/env` against what's already on the box: it reports only whether the
  content differs (`diff -q` via process substitution, nothing touches local disk).
  Exercised in this pass against an intentionally-unreachable host and against
  `127.0.0.1` (connection refused) — both paths report cleanly and exit 0 without a real
  box to deploy to.
- **beets on the box**: installed via `uv tool install --force beets==2.13.1` as the
  `music` user (`HOME=/srv/music`), then symlinked `/usr/local/bin/beet ->
  /srv/music/.local/bin/beet` — `uv tool install`'s default bin dir is not on systemd's
  default `PATH`, so without the symlink `music-import`'s `beet import` call would fail
  with "command not found" the first time a real timer fires. `config/beets-config.yaml`
  mirrors the workstation's `~/.dotfiles/beets` layout and plugins (fetchart, embedart),
  repointed at `/srv/music`, plus a quiet-mode section (`quiet_fallback: skip`,
  `none_rec_action: skip`) so `beet import -q` sends anything it can't confidently tag to
  review instead of guessing or blocking on input that will never arrive.

## Task 2: import, backup, and wrapper scripts (2026-08-30)

Landed `scripts/lib.sh` (`quiescent`, `safe_extract`, `free_space_ok`, `notify`,
`ping_hc`), TDD with `tests/lib.bats` written and confirmed red before any
implementation existed. Landed `music-import`, `music-backup`, `music-beet`, and
`scripts/check.sh` as the repo gate (shellcheck over every `scripts/*` bash file,
then the full bats suite). `tests/music-import.bats` runs the importer end to end
against a fixture `/srv/music` tree with `beet`, `curl`, and `msmtp` stubbed on
PATH: happy path, a beets skip filed to review with both notifications, a
non-quiescent inbox left untouched, and a free-space abort that fires the fail
ping before beet ever runs.

Two one-line review cleanups from Task 1 landed alongside: the volume glob guard
in `cloud-init/user-data.yaml` now also rejects a non-block-device match (a
zero-match glob previously fell through as a false positive on some shells), and
`provision.sh`'s firewall-rule check now fails with its own message when the
`hcloud firewall describe` call itself errors, instead of reporting a network
blip as a missing rule.

Decisions a later pass would be wrong to rediscover from scratch:
- **Contributor matching**: `music-import` matches an inbox directory name
  (`/srv/music/inbox/<user>`) against contributors.yaml's `name` field verbatim;
  there is no separate username/slug field. If Task 3/8's FileBrowser account
  provisioning wants a different inbox directory naming scheme, `contributors.yaml`
  needs a dedicated field and `uploader_email()` needs updating to match on it.
- **Navidrome scan trigger**: implemented against the native admin API (`POST
  /auth/login` for a JWT, then `POST /api/scanner/scan` with `x-nd-authorization:
  Bearer <token>`), not the Subsonic API, because Subsonic has no scan-trigger
  endpoint. This is the UI's internal API rather than a documented stable
  contract; verify it against the pinned Navidrome image tag once Task 3 chooses
  one, since a trigger failure is only best-effort (logged, not fatal) precisely
  because this endpoint is not guaranteed stable.
- **State DB paths default to** `$MUSICBOX_ROOT/state/library.db` (beets) and
  `$MUSICBOX_ROOT/state/navidrome/navidrome.db` (Navidrome), both overridable via
  env var. Task 3's `compose.yaml` must either mount Navidrome's data dir there or
  the env var overridden in `/etc/musicbox/env`.
- **Archive cap default**: `MUSICBOX_MAX_ARCHIVE_BYTES` defaults to 2 GiB in
  `music-import`; revisit if a real Bandcamp download exceeds it.
- New host dependency: `yq` (mikefarah/go, v4) for parsing `contributors.yaml`,
  needed on the box alongside `beet`/`rclone`/`msmtp`/`bsdtar`. Add it to Task 3's
  `deploy.sh` package list.
