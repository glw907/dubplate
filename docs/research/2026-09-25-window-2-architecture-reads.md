# Window 2 architecture reads (2026-09-25)

One `go-architecture-reader` per package touched by the matcher word-boundary pass and pass
6b-ii-a, each run at master `6e6fd12` with no plan in context. The findings route to the
architecture cleanup pass unless a line below says otherwise.

## internal/ops/local: sound with nits

Density 0.26 against `database/sql` at 0.42.

- **Test-only seams.** `insert` takes a `body` parameter that production always passes as nil
  (`jobs.go:83`, `local.go:182`); drop it and let the test set `j.body`. The `bindBackup binder`
  field and the second constructor `open` exist only for tests (`local.go:63,81,86`). A
  package-level `var bindBackup = (*Runner).backupJob` that tests swap collapses `New` and
  `open`.
- **Duplication.** `markRunning` and `finish` share one guarded-UPDATE, re-read, and log body
  (`jobs.go:141-167`, `:198-225`); one `transition` helper. `Catalog` restates the privilege
  table positionally and patches `entries[0]` (`local.go:126-141`); `ops.BuildCatalog` should
  fill privileges itself. The live and terminal state sets are SQL literals that must agree with
  `ops.JobState.Terminal` by hand (`local.go:218`, `recovery.go:34,44`, `jobs.go:95`), and the
  not-found construction repeats (`jobs.go:289,323`).
- **File split (nit).** `conflict` (`local.go:207`) belongs beside the other queries in
  `jobs.go`; the binder types belong beside `start`.
- **Comments (trim).** `wrapCanceled`'s 12-line doc (`local.go:266-277`); `local.go:120-122`
  rebuts an alternative nobody wrote.
- **Tests (nit).** `jobs_test.go:146` computes its expectation with the encoder under test.
- **Best passages.** `storeContext` (`jobs.go:76-78`), `outcomeOf`'s precedence
  (`jobs.go:178-193`), and `wrapCanceled` (`local.go:278-283`).

## internal/ops/opshttp: sound with nits

Density 0.313 against `net/http/pprof` at 0.318. The package exports nothing, which is correct
before a caller exists.

- **Duplication.** The three denial causes and the `denials` slice list one set twice
  (`authorize.go:53-76`); an unexported `denial` string type read through `errors.AsType`
  removes the list. The four route bodies differ only by type, and each op name is written
  twice (`opshttp.go:72-77`, `:105-139`); a generic `serve[Req, Resp]` adapter binds it once.
  Reason extraction is written twice, and `validHandle`'s fallback text is unreachable
  (`opshttp.go:176-186`, `:214-217`); one `reasonOf`.
- **Test-only seam.** `checkStatuses` takes a `table` parameter only a test varies
  (`opshttp.go:37,68`); move the invariant into a test over `ops.Codes()`, which also catches a
  code newly added to `ops` (the current `TestStatuses` cannot).
- **Test discipline.** `contract_test.go` holds the third copy of the rclone stand-in
  `TestMain`, `runChild`, and `syncBuffer` (the others are in `internal/backup` and
  `internal/ops/local`). The job-path and cross-process rows test `ops/local` behavior; the
  two-leg code-agreement table is what this package owns. One shared test helper.
- **Comments.** `doc.go:6-19` repeats the decorator's type comment, and its last paragraph is
  build status that goes stale at mount.
- **Forward note, to rung 14.** `authorizer` has an unexported method, so only a type inside
  `opshttp` can implement it. The mounting rung decides on purpose whether the real authorizer
  lives in `opshttp` or outside (then `Authorize`, or a func type).
- **Best passages.** Exact-path dispatch with its reason for not using `ServeMux`
  (`opshttp.go:46-49`), the refusal of a JSON `null` body before decode (`:155-161`; a zero
  `start_backup` request is valid and privileged), and `admit` (`authorize.go:95-112`).

## internal/ops: sound with nits

Density 0.60 against `io/fs` at 0.57.

- **Exported ahead of a caller.** `NetworkActor` has no non-test caller (`ops.go:47`); its
  first production caller is rung 14's session authorizer. The plan specified it, so the pass
  kept it. The cleanup pass or rung 14 either unexports it until then or lands the caller. The
  package doc (`doc.go:8-10`) and `Actor`'s doc (`ops.go:21-23`) disagree about whether this
  package builds the network kind; one of them is wrong today.
- **Ceremony.** `Catalog`, `BuildCatalog`, and `Entries` have one caller, which unwraps the
  result on the next line, and `BuildCatalog`'s zero-privilege check cannot fire
  (`catalog.go:103-121`); one ordered operations table in `ops` and a `CatalogEntries` builder.
  Seven one-line error constructors differ only by sentinel (`errors.go:71-94`); one exported
  `NewError(sentinel, reason)`. `ParseHandle`'s value is discarded by both callers
  (`ops.go:76-84`); a `Handle.Valid()` method or `UnmarshalText`, with the two identical error
  branches folded.
- **Test-only seam.** `Privilege.UnmarshalText` has no production decoder (`catalog.go:51`).
- **Across packages.** Privilege ordering lives in `opshttp`'s `reaches`; a
  `Privilege.Covers` method belongs on the type. `JobState.Terminal` is restated as SQL in
  `local/jobs.go:95`.
- **File split.** `KindBackup` (`catalog.go:72`) belongs in `ops.go` beside `JobStatus`.
- **Comments.** A process citation at `errors.go:9-11`; caller-business comments at
  `ops.go:30-31,63-65` and `errors.go:123-127`. Test comments carry process citations
  (`ops_test.go:48,222,235`, `catalog_test.go:9,76`, `errors_test.go:11,97`); `errors_test.go:58`
  says five where the table holds seven; `errors_test.go:30-44` keeps a self-checking literal
  map.
- **Best passages.** `Seam` takes the actor as a parameter and no request carries identity
  (`ops.go:164-189`); `Privilege`'s invalid zero value and failing `MarshalText`
  (`catalog.go:5-46`); `CodeOf`'s precedence (`errors.go:108-121`).

## internal/harness/conform: workmanlike

Density 0.363 against `testing/fstest` at 0.153. The package predates this window; the matcher
pass touched only its floors and case-count comments, so every finding below is standing debt.

- **One reporting gap, first in the cleanup queue.** `GradeCases`' fold discards `ModeDefault`
  near misses: the `continue` at `case.go:231` skips the `NearMisses` append at `:244`, so a
  `defect_free` distance that passes only within tolerance is never reported. One
  `(*caseGradeSummary).add` that the fold and `gradeFixtureCase` (`case_test.go:218-222`) share,
  with `NearMisses` outside the mode gate.
- **Two float comparers disagree.** `case.go:855` scales by `max(|want|,|got|)`;
  `component.go:828` scales by `max(|want|,1)`, which is absolute below 1, and hard-codes `1e-12`
  beside the named `nearMissTolerance`. One comparer, one tolerance.
- **Exported surface.** Nothing imports the package, yet 30 identifiers are exported
  (the floors, `GradeCases`, `ReplayTable`, the `Check*` functions, and the rest). Unexport all;
  update the comment references at `internal/engine/matcher/compat.go:26` and
  `internal/engine/shadow/rescore.go:40`.
- **Duplication.** The replay prologue three times (`case.go:352-398`, `:553-603`,
  `request.go:114-145`); the corpus directory walk four times; the component method set in three
  hand-kept lists (`component.go:24-36,112-137,199-207`); the compat classes three times
  (`case.go:993-999,703-708,525-527`); `fieldNullability` restated at its call sites
  (`translate.go:70-96` vs `:136-148,174-181`), so the coverage test counts a table production
  does not use.
- **Test-only seams.** `tiesAwaitingRuling` is always empty and its guard runs only from a test
  (`case.go:105,1008`); `CaseGradeSummary.Mode` is never read; `ReplaySummary.TotalRows` echoes
  `len(t.Rows)`.
- **File split.** Follows the pass history; a `profile.go`, `corpus.go`, `policy.go`, and
  `compare.go` split reads as concerns.
- **Comments.** `doc.go` is a pass history; process citations recur across the package (the
  staged citation leg will gate them). `case.go:475-480` contradicts its function (it checks
  `Pending`, not `GradeVerdict`); `defaults.go:15-20` is stale; `case_test.go:112` says sixteen
  rows against fifteen, and its guard at `:140` cannot fire.
- **Best passages.** `decodeOption` refusing a missing `"regex"` (`component.go:378-404`), the
  unadjudicated divergence printing the exact adjudication row from the real type
  (`case.go:481-497`), and the three-way `nullPolicy` (`translate.go:31-56,102-121`).

## internal/engine/matcher: workmanlike

Density 0.760 against `regexp` at 0.291. The word-boundary pass added one exported field
(`SDPattern.Replacement`), as its exit gate allowed; the rest is standing debt from rung 3.

- **Wrong comments, first in the cleanup queue.** `Distance.Update` (`penalty.go:202`) has no
  production caller, and its doc (`:197-201`) claims the album components fold through it, which
  they do not (`album_tracks.go:28`); delete it and `TestUpdate`. `likelies.go:34-39` says the
  consensus map exists for parity, but `musicbrainz/query.go:80` and `tag/tag.go:75` read it.
- **Duplicated legality checks that have drifted.** `quiet_fallback` and `none_rec_action` are
  checked at `tunables.go:481-490` and `verdict.go:184-206`; verdict.go's runtime error strings
  cite beets line numbers the pinned venv no longer has (`session.py` `_summary_judgment` sits at
  365, the `as_choice` calls at 377 and 383). One parser each, no source-line cite in an error.
  The required-field check repeats at `candidate.go:121-125` and `tunables.go:451-455`.
- **Exported surface.** About seventeen exports have no outside caller: `AssignItems`,
  `AlbumPairsFor`, `AlbumPair`, `DistanceForAlbum`, `SDEndWords`, `SDReplacement`, `SDReplace`,
  `BeetsSDPatterns`, `ParseRecommendation`, `ParseAction`, `LegalQuietFallbacks`,
  `LegalNoneRecActions`, `Transliterate`, `Levenshtein`, and `ErrInvalidTunable` (lower rank).
  `SDPattern` and `SDPatternsNoBracketDiscount` are reached only by a conform test;
  `Tunables.StringDistPatterns` is written only by `Compat.Apply`. Keep `Compat` as the one
  public selector.
- **Test-only seam.** The mutex-guarded foreign-pattern cache (`strdist.go:161-170,185-192`)
  serves a path nothing drives; compile each regexp inside its row.
- **Duplication.** `compileMediaPatterns` and `compileCountryPatterns` differ by one wrapper and
  are recovered from a panic in `Validate` (`compilePanicAsError`, `tunables.go:504`); use a
  Compile/MustCompile split. Component `Name` and `Key` are equal in all 21 rows and restated by
  each adder's literal. `SDPatternsNoBracketDiscount` restates four rows by hand; derive it as the
  test already does. `UnidecodeVersion` repeats the table's own version field.
- **File split.** 21 component files follow the task history; `album.go` and `track.go` would
  hold the two ports. `Levenshtein` belongs in `strdist.go`, as do the SD tables; the package doc
  belongs in `doc.go`.
- **Comments.** At least 40 process citations and 36 "contract N" references in shipped
  comments; stale line cites at `quiet.go:14` and `tunables.go:498`; rebuttals of unwritten
  alternatives; the `inclArtist` divergence told three times.
- **Best passages.** The pattern loop's `caseDelta == 0` comment (`strdist.go:41-63`),
  `solveLAP` reproducing lapjv's pivot order (`assign.go:149-192,384-391`), and `Add`'s
  NaN-failing range test that panics rather than clamps (`penalty.go:223-231`).
