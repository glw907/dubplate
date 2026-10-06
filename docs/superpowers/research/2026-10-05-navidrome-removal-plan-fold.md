# Navidrome removal plan: fold record

Target: `docs/superpowers/plans/2026-10-05-navidrome-removal.md`, revision 1 at `5ec8a4a`, now revision 2.
Reviews: `docs/superpowers/research/2026-10-05-navidrome-removal-plan-review-{contract,mechanics,risk}.md`.
Spec: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1.
Method: each finding was checked against the tree at `5ec8a4a` before it was folded. A finding id appears once below.
Convergent findings were fixed once, and each gets its own row pointing at the shared fold.

**Totals:** 38 findings. 35 folded, 3 refused, 0 owner forks.

## Verified before folding

- `~/.claude/workflows/pass-execute.js:262-268` gives `tool` the class default `gateLane: "light"`, resolved at
  `:359`. `sweep` sets `reviewerModel: "sonnet"` and `coverageBlocks: false` (`:248-253`), resolved at `:569`.
- `internal/serve/prefix_test.go:162-166` holds the `TestRoute` row "a neighbor sharing the prefix's bytes without
  a segment boundary does not enter", which has its own `next` spy.
- `ErrUnavailable` has a production use: `musicbrainz/client.go:392` wraps it on retry exhaustion.
  `tag_test.go:274` asserts it. `ClassBackground` has no caller outside the shadow package (`run.go:271`).
  `limiter.go:11-41` argues the reservation in three-class terms.
- `github.com/spf13/pflag@v1.0.10/errors.go:40` contains `"unknown flag: --%s"`.
- `cmd/dubplate-server/main.go:110` defines `newProductionRegistry`, and `serve_test.go:97` defines
  `productionRegistry`. `auth_test.go:73-74` and `:239-241` exercise the `_PREVIOUS` rows.
- `corpus/README.md:2290-2303` documents the `album` plus `release_mbid` subject. `BACKLOG.md:43` names
  `--streamer-username`.
- `TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost` is a general round-trip test. Only its last two
  assertions concern the deleted fields.
- `git grep -niw streamer -- web/src web/scripts` hits `deck.ts`, `+page.svelte`, and `run.ts` only.

## PM4: the ClassBackground reading

The spec's sentence is "`musicbrainz.ClassBackground` and `ErrUnavailable`'s shadow justification stay only if rung
9's spec names a consumer. Otherwise they are pruned". It is ambiguous: "they" can mean the symbols or their
justification. The fold takes the justification reading, which leaves the cleaner architecture, for three reasons:

- **The limiter is a three-class design.** The reservation rotates round-robin through the classes below the one
  strict priority would serve. With two classes, the cursor, the modulo walk, and the rotation test have nothing to
  do. A removal pass would have to choose between keeping hollow structure and redesigning concurrency code.
- **Pruning one class buys no consumer honesty.** `ClassUserPresent` and `ClassBatch` have no non-shadow caller
  either. After the shadow goes, the whole `Limiter` waits on the import pipeline.
- **`ErrUnavailable` is a live contract.** It is the client's retry-exhaustion sentinel, and deleting it would
  change the client's error shape.

Ruling 6's "reads as if it had never had the removed part" is met by rewriting the comments without the shadow.
Reading 2 is rewritten to match. Task 2b drops the limiter mutation and its files become comment-only.

## Ruling 6 and the reader findings

Ruling 6 is Geoff's. Its text is "Every touched package is graded by a `go-architecture-reader` at the pass's merge,
as for any pass", and it excludes no touched package. OC1 (contract), PO1, and OC1 (risk) each propose dropping
readers for packages with comment-only changes. All three are refused. Reading 6 now records this.

## Dispositions

### Contract lens

| Id | Disposition | Where |
| --- | --- | --- |
| PC1 | Folded. The segment-boundary test is deleted with its reason. The `TestRoute` row and the bare-`HasPrefix` mutation are named. The terminator test is renamed to a behavior pin, with the "`Terminator` answers 200" mutation. The Flush test gains a reached check, with the "delete `Flush` and `Unwrap`" mutation | Task 1 dispositions table and named mutations |
| PC2 | Folded. `gateLane: "heavy"` and `reviewerModel: "claude-opus-5-5"` are added to the runner args, citing the runner lines | Header, "Execution mode" |
| PC3 | Folded. Task 4b becomes `auth-data`, and Task 4a stays `sweep` with an Opus reviewer through PC2 | Header "Pass class"; Task 4b |
| PC4 | Folded. The method uses `env -i`, scratch paths, and valid sample values, and iterates until `configuration: ok`. Keys are classified as template literal, generated, or operator-typed, and the count is the operator-typed class. Task 3 and the close compare the key set | Task 0; Task 3 acceptance; close step 7 |
| PC5 | Folded. The test lives in `cmd/dubplate-server` against `newProductionRegistry`, with the mutation "re-add `registerShadowIntegration`" | Task 2a acceptance |
| PC6 | Folded. The uncapped standalone leg is named as the method, and the expected `internal/serve` list is cited literally | Reading 1; Task 1 acceptance |
| PC7 | Folded. Reading 7 extends to `serve_test.go`, and the test is renamed `..._NonceSweepStopsWithTheGivenContext` | Reading 7; Task 1 dispositions table |
| PC8 | Folded. `IdentitySource.Optional` is named out of scope. The `registry.go:197-300`, `keys.go:172-202`, and `config.go:21-28` ranges are added | Reading 3; Task 2a files and outcomes |
| PC9 | Folded. `git grep -niw streamer -- web/src web/scripts` is added | Task 4b acceptance |
| PC10 | Folded. The conductor fetches the A-Z entry in segment B's pre-flight and puts it in the dispatch | Task 4b outcomes; conductor action 4 |
| PC11 | Folded, refined. `corpus/README.md` joins Task 2b. The round-trip test is kept, because only its two deleted-field assertions lost their subject, and those two go | Task 2b files and outcomes |
| PC12 | Folded. pflag's error text is quoted. The SQLite `RENAME COLUMN` quote is the conductor's at a `0006` re-cut, the only path where it matters | Behaviors table; preconditions section |
| OC1 (contract) | Refused: ruling 6 is an owner ruling, and its text excludes no touched package | Reading 6 |

### Mechanics lens

| Id | Disposition | Where |
| --- | --- | --- |
| PM1 | Folded. Source `git grep` criteria now sit beside the `go doc` ones for `internal/serve`, `internal/auth`, and `internal/config`, with named survivors. The unexported stale comments are listed in the outcome | Task 1 outcomes and acceptance; Task 3 acceptance |
| PM2 | Folded through PC1. The terminator test is kept as an honestly named pin with a mutation that fails it, so it is not deleted | Task 1 dispositions table |
| PM3 | Folded. Task 2b carries a disposition table for the twelve names, each kept with its consumer or the inference. Reading 1 adds `ReadFrom`. HISTORY records the real delta, and the wrong `tagio` cause is replaced | Task 2b; Reading 1; Task 5 |
| PM4 | Folded on the justification reading, recorded above | Reading 2; Task 2b |
| PM5 | Folded through PC6 | Reading 1; Task 1 acceptance |
| PM6 | Folded. Rows are added for `cookieValue`, `cookieLines`, and the `newTestShell` parameter | Task 1 dispositions table |
| PM7 | Folded through PC7 | Reading 7 |
| PM8 | Folded through PC5. The report shows the failure under the mutation | Task 2a acceptance |
| PM9 | Folded through PC8 | Task 2a files |
| PM10 | Folded. Task 4b's grep excludes `internal/engine/store`, and the precondition no longer stalls inside a segment | Task 4b acceptance; "Independence" |
| PM11 | Folded. `BACKLOG.md:43` joins Task 4b's files and grep | Task 4b; "Contested surfaces" |
| PM12 | Folded. The end shape is stated as one package-level identity set, with `Strip.identity`, `trailerScrubbingBody.keys`, and the `keys` parameters removed | Task 1 outcomes |
| PO1 | Refused: ruling 6, as for OC1 (contract) | Reading 6 |
| PO2 | Folded as a consequence of PM4. The limiter rewrite and its mutation leave Task 2b | Task 2b |

### Risk lens

| Id | Disposition | Where |
| --- | --- | --- |
| PR1 | Folded. The re-confirmation runs before segment B launches, and a failure re-cuts Task 4a first. The date and an output summary go into Task 4a's dispatch. HISTORY records the order | Preconditions section; Task 4a; conductor action 4; Task 5 |
| PR2 | Folded. Slice 2 merges only at a lane 1 boundary with no document segment in flight. The conductor holds the token and runs the full gate on master before the next launch. Task 0 warns before a shared-binary bump | "Where it runs"; "Contested surfaces"; conductor action 5 |
| PR3 | Folded. Task 0 counts as a lane 1 segment for the sharing rules | "Where it runs"; Task 0; conductor action 2 |
| PR4 | Folded through PC4. The exact `env -i` command and the review's 14-key, three-round measurement are recorded | Task 0; Task 3 acceptance |
| PR5 | Folded. A stopped archive chore sends Task 4a down the `0006` path, with its grep acceptance amended | Preconditions section; Task 4a |
| PR6 | Folded. Non-key inputs (`member add` arguments, `unidecode install`) are listed beside the count, uncounted | Task 0 |
| PR7 | Folded. The close's steps 1 to 4 and their folds count as an in-flight lane 1 segment. **The sibling plan's "Segments" line gets the matching change:** "the merge" now counts the close's simplifier and review-fold commits | "Where it runs"; Task 5; `2026-10-05-replan-document-pass.md` "Segments" |
| PR8 | Folded through PC1 | Task 1 dispositions table |
| PR9 | Folded through PC3. The security reviewer's dispatch adds Task 4b's `roster.go` and `admin.go` diff | Header "Review chain"; Task 5 step 2 |
| PR10 | Folded through PC8. Acceptance names the existing `_PREVIOUS` tests in `auth_test.go` | Task 2a acceptance |
| OC1 (risk) | Refused: ruling 6, as for OC1 (contract) | Reading 6 |

## The coordinator's added duty

The shared heavy lock with the cairn-cms session is folded as one conductor duty in the header, with a matching
step in conductor action 3. Before a heavy stretch expected to run over 15 minutes, the conductor sends the
cairn-cms session one line through `SendMessage`. It never re-runs a full gate on an unchanged commit. Source:
`~/.claude/docs/pass-gate-economy.md`, "Sessions sharing the heavy lock coordinate".

## Measures

- **New-mechanism findings:** 0 proposed, 0 folded, 0 refused. Every fold uses existing machinery: existing
  runner args, the plan's own `0006` path, the existing exported-unused leg run uncapped, and the existing tests.
- **Over-ceremony findings:** 3 refused (ruling 6) and 1 folded (PO2, through PM4).
- **Plan line count:** 571 before, 684 after.
- **Ceiling:** 8.0M, unchanged, with the flag still at 6.4M. Removing the limiter rewrite (PM4) offsets the
  heavier `auth-data` bar on Task 4b. The ten architecture readers stay.
- **Sibling plan:** one line changed (the "Segments" definition of "the merge").

## Verification minors

- PV1: folded. Task 2b's Files add `limiter.go:68`, `client.go:340`, and `limiter_test.go` (comments, label rename to `"background"`); the outcome says assertions, not tests, are unchanged.
- PV2: folded. Task 2a's mutation is now "run the new test at the pre-deletion commit and show it fails there".
- PV3: folded. Segment A's pre-flight records the uncapped list at `HEAD`; the conductor passes it in Task 2b's notes; Task 2b's own before-run is dropped.
- PV4: folded. Conductor action 4 and Task 3's acceptance carry Task 0's classified set and sample environment in Task 3's dispatch notes.
- PV5: folded. The doctor order test pins `pipeline` before `monitoring` with the named mutation "sort by name" (verified: `keys.go` has both at rung 9, alphabetical order reverses them).
- PV6: folded. Task 5 states it runs before document segment B launches or after its Tasks 5 to 8 invocation ends, never during it.
- PV7: folded. Two quoted rows added: `go doc -src net/http.ResponseController.Flush` (source switch) and `golangci-lint run --help` (`--tests` default true, 2.14.0), with the verifying command.
- PV8: folded. Task 5's STATUS outcome carries one line naming Reading 2's departure from the spec's default and why.
- PV9: folded. `registry.go:289` dropped from Task 3's coverage claim, with a note that Task 2a removes it.
- Over-engineering Q4 item 2: folded. Close step 7 dropped as a repeat of Task 3's key check; the HISTORY outcome now cites Task 3's comparison.
