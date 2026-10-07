# Rung 12 slice 2b plan review: contract and criteria lens

Target: `docs/superpowers/plans/2026-10-07-rung-12-slice-2b-oracle-harness.md` at `0ae2b54` (revision 1, pass class
`tool`, Task 11 `docs`). One lens: is every promise testable, does each criterion name its fixture and its failure
reason, can any pass vacuously, and does the ceremony fit the risk. Findings are correctness gaps ranked by
consequence, then over-ceremony ranked by cost. Polish is out of scope.

Verified against the tree at `0ae2b54`: `grade/ledger.go`, `grade/grade.go`, `grade/differ.go:160-190`,
`grade/ledger_test.go:457-477` (`TestCommittedLedger` pins owed counts by slice), `grade/testdata/seed-ledger.json`
(27 rows, 22 of them `volatile` rows on `serverVersion` and `type`), `grade/testdata/seed-report.golden`,
`scripts/check.sh:219-375`, `Makefile:134`, the seed manifest (every album shape Task 5 names is present), the seed
session's bodies (no credential URL; gonic `015` carries `replayGain: null`, `008` carries `releaseTypes: [""]`), the
pinned OpenAPI bundle (87 paths; `Child` 54 properties with `mediaType`; `AlbumID3` lacks `mediaType` and has
`isCompilation`, which `Child` lacks), the study record sections 1 and 3, and the divergence ledger record (54 L rows,
12 A rows, no escaped `|` in any cell). Measured on the workstation: the XSD fetched to the scratchpad matches the pin
(`9baee72f...0387`) and uses only `complexType`, `complexContent`, `extension`, `sequence`, one `choice`, `attribute`,
`element`, and `simpleType` restrictions; `xmllint --version` writes to **stderr** and nothing to stdout (exit 0);
validation errors exit 3, a malformed body exits 4, and an uncompilable schema exits 5. `go doc net/url.ParseQuery`
and `go doc net/url.URL.Query` quoted under C10.

Counts: 0 blockers, 6 major, 9 minor correctness findings; 1 minor over-ceremony finding. No owner fork.

## Correctness gaps

### C1 (major): the assignment's data has no source for 36 endpoints and no encoding for split rows, and no test checks a value

`plan:518-524`, `:544-553`.

**Defect.** Task 5 says every one of the 87 endpoints gets an oracle and a slice, and that "the values are the study
record's section 3 table under the rulings". That table names 51 endpoints. The other 36 in the bundle have no row
there: `getInternetRadioStations`, `getPodcasts`, `getNewestPodcasts`, `getLyricsBySongId`, `setRating`,
`getBookmarks`, `getPlayQueueByIndex`, `reportPlayback`, `jukeboxControl`, `getUsers`, and 26 more. Task 8 records two
of them against both oracles, yet the plan never says what the assignment holds for them. Several section 3 rows also
split one endpoint across oracles, and the four-value field cannot hold them. `stream` is "Both" direct and
"Navidrome" for a transcode. `getOpenSubsonicExtensions` is "Navidrome for the carve-out; contract only for the
list". `getPlayQueue` is "Navidrome for the empty answer; contract … for the rest". `getCoverArt` and the playlist
endpoints carry carve-outs too. The plan never says where a slice value comes from, since section 3 has no slice
column. The pass's own headline promise is "an assignment that never compares dubplate with an oracle the study ruled
unfair" (`:29`). The test mandate proves only the mechanism (the three-source session) and endpoint-set coverage. No
test checks any value, so a wrong `both` on `stream` passes every gate.

**Fold.**
- Name the sources. Section 3 covers the 51 endpoints it names. Study record section 1 and spec `:815-818` say the
  empty-success set answers ok, and that set grades contract only. Spec `:813-815` says every other endpoint answers
  error 0, which is `not-served-0.8` and contract only. Slice values come from the spec's slice ladder.
- Rule split rows the narrow way. The single value is the oracle that is fair for every request form of the endpoint
  (`stream` → `navidrome`, `getOpenSubsonicExtensions` → `contract-only`, `getPlayQueue` → `contract-only`). The
  wider grant goes in the reason, and the serving slice adds a per-request carve-out when it needs one. This holds the
  "never unfair" promise by construction.
- The loader refuses a `not-served-0.8` entry whose oracle is not `contract-only`. Add it to the many-bad list.
- Add a table test that pins every value Task 5's outcomes bullet names, and the not-served set. That set should equal
  the 87 endpoints minus the launch list and the empty-success set. Mutation: flip `getTranscodeDecision` to
  `navidrome` (the test fails).

### C2 (major): the ledger's `slice` column is claimed "proved by the transcription test", and nothing proves it

`plan:250-254` (R13), `:567`, `:576-577`, `:589-591`.

**Defect.** The transcription test's three proofs (round trip, accounting, fidelity) compare columns the record holds.
`slice` is not a record column. R13 derives it, and no test recomputes it. The plan names "owed to the right slice" as
a thing slice 3 consumes (`:27-28`), so a misrouted row hides from slice 3's plan. R13 is also ambiguous on real cells.
L-02's cell reads "Harness: …; rung 16: Amperfy", and a rung is not a slice. L-35's reads "Slice 5, if fork A is
yes", and that row retires. Several cells name "Harness" ahead of a slice.

**Fold.** Make R13 mechanical. The first `Slice N` token in the cell wins. "Harness" or "Contract tier" with no slice
token takes the assignment's slice for the row's endpoint. A rung is never a slice. The listed no-endpoint rows take
slice 3. The transcription test recomputes `slice` from the record's cell and the committed assignment, then compares
it with the committed value. `TestCommittedLedger`'s owed-by-slice map (`ledger_test.go:474`) is updated to the
recomputed counts. Mutation: move one row's `slice` (the test fails).

### C3 (major): "a `contract` row's pattern names one path in both tiers" is false for a list with one item

`plan:490-492`, `:192`, `:505-508`.

**Defect.** JSON contract paths come from the validator's pointer, and a list path is always indexed: a one-song
album's field sits at `album/song/0/x`. Task 4 maps an XML ordinal under `tree.go`'s rule, where only a name repeated
among siblings becomes an array. So the same field in an XML one-song album sits at `album/song/x`. A `format: any` row
at `album/song/*/x` silently misses every XML singleton, and so does an `xml` row. The seed fixture holds one-track
"Loose files", and slice 3's single-track albums will hit this. The path proof uses two songs (`:506`), so it cannot
catch it.

**Fold.** `xmlrule` already resolves each element's type by position, so it knows whether that position is declared
as a list (`maxOccurs` greater than 1). Report that flag per ordinal. Task 4 indexes a list position even when it holds
one element, which matches the JSON tier. The oracle tier's XML rule stays as 2a left it. Add a proof: a one-song
album's XML violation lands at `…/song/0/…`, and one `format: any` row suppresses the same field's finding in both
formats. Mutation: index by sibling repetition (the singleton proof fails).

### C4 (major): the residual-scan proof cannot tell the scan from xmllint

`plan:437-439`, `:468`, `:476`.

**Defect.** A misspelled attribute left in place is exactly what xmllint reports itself. The residual scan's only
added value is the case xmllint hides: anything after a parent's first unexpected child (`:206`). The `titel` proof
puts the misspelling where xmllint sees it. With the scan removed, one violation still comes out, so the mutation
"skip the residual scan" survives unless the test also asserts the violation's origin. The plan's own mutation text
("the misspelling proof passes and the test fails") reads as if the authors noticed this. That shows up in slice 3 as
an XML pass on a dubplate body that carries an error xmllint never reached, which breaks "a tier whose only pass is a
real pass".

**Fold.** Put the proof where xmllint goes blind. An `album` carries an unexpected child element first, then a `song`
with `titel`. Assert both violations, and assert the `titel` one comes from the residual scan. Keep the in-place
`titel` case for the dedupe rule (one violation, not two).

### C5 (major): Task 5's three-source grading proof can pass vacuously, depending on what the dubplate source holds

`plan:547-553`.

**Defect.** The plan does not say what the hand-built session's `dubplate` bodies are. The natural choice is a copy of
one oracle's bodies. Then every check passes without proving anything. A skipped oracle tier and a run one both give
zero findings on `getIndexes`, so "skips the other two" holds even when nothing is skipped. "Grading dubplate against
gonic compares `getMusicFolders`" gives zero findings either way. "Reports `getAlbum`'s oracle findings" needs
dubplate's body to differ from Navidrome's. The mutation "treat an unmapped endpoint as `both`" has no request to catch
it, because the session holds no unmapped endpoint. "Skip the contract tier on an unassigned request" fails only if
`getIndexes` yields a contract finding, and no one has checked that it does.

**Fold.** Name the fixture state. dubplate's `getAlbum` and `getIndexes` bodies are gonic's, so they differ from
Navidrome's. Its `getMusicFolders` body is Navidrome's, so it differs from gonic's. Add one request to an endpoint the
assignment lacks, with differing bodies. Assert a non-zero finding count wherever a comparison is claimed, and zero
oracle-tier findings plus the skip count wherever a skip is claimed. Perturb `getIndexes` with one unknown attribute so
its contract finding is known to exist.

### C6 (major): the conductor may rule Navidrome's second library unmeasured, and Task 10 then has no reachable acceptance

`plan:153-154`, `:683-686`, `:729-732`, `:818-819`.

**Defect.** The unattended run pre-authorizes "recording Navidrome's second library as unmeasured". Task 10 still
requires both oracles to answer every request, a Navidrome-against-gonic golden, Navidrome's error 70 on all ten
endpoints, and per-collection filtering on each oracle. The plan also lists, as unverified, whether the first admin
sees a new library without a grant. That failure stops readiness (`:697`), and the plan names no remedy. Either branch
turns into an `escalate` that stops a 10-hour run at segment D.

**Fold.** Write Task 10's degraded form now. If Navidrome holds one library, the session records gonic alone for the
filter requests. The test then asserts A-04 and L-19's gonic half on the bodies, with no pair golden. L-19's Navidrome
half stays slice 1's observation, recorded as an erratum. Add "grant the admin the new library, through the same
observed UI request" to Task 9's outcome if the readiness count shows the library is missing.

### C7 (minor): Task 7 makes 22 seed-ledger rows stale, and its file list omits that ledger

`plan:602-603`, `:623-624`, `:637`.

**Defect.** E1 removes the `serverVersion` and `type` value findings. Twenty-two active `volatile` rows in
`grade/testdata/seed-ledger.json` (`server-type-*` and `server-version-*`) then match nothing, and `TestSeedGrade`'s
every-row-matches check fails. `seed-ledger.json` is not in Task 7's Files, so a diff-reviewer may flag the needed edit
as out of scope.

**Fold.** Add the file. Retire the 22 rows and cite E1 in the commit message. The golden's "ledger rows matched" block
loses them.

### C8 (minor): Task 8 both asserts each prediction present and says an absent one is not a failure

`plan:656-659`.

**Fold.** The test pins each prediction's recorded state, present at a path or absent, as the report gives it. The
report reconciles each state with [contract] E6. An absent prediction is then a pinned fact, not a red test.

### C9 (minor): one escaping proof compares identical bodies, and the E6 fixtures are deferred

`plan:613-614`, `:632-633`.

**Defect.** "`&` against `&` in JSON" compares a body with itself, so it yields zero findings whatever the differ
does. The intended JSON form is the `&` escape. The E6 rows say "the task names which" body to use, and the plan
never names one.

**Fold.** Write `&` against `&`. Name the fixtures: gonic `015` (`replayGain: null`) for the null rows and gonic
`008` (`releaseTypes: [""]`) for the `[""]` row. Build the `{}`/`{"playlist":[]}` row by perturbation.

### C10 (minor): the credential scan fails open, and the `net/url` behavior it relies on is not quoted

`plan:616-619`, `:626-630`.

**Defect.** The states table says "A URL leaf that does not parse: no finding". For a guard whose invariant is "a
session never commits an echoed credential", that is the wrong direction. `url.Parse` rejects a whole string over one
bad percent escape. `go doc net/url.ParseQuery` says "Settings containing a non-URL-encoded semicolon are considered
invalid", and `URL.Query` "silently discards malformed value pairs". So `?u=a&t=x;s=y` hides `t` and `s`. The oracle
credentials are throwaway (gonic `admin`/`admin`, Navidrome a random secret), so the consequence is small. The plan
still claims the invariant.

**Fold.** Quote both docs. When parsing errs, fall back to a token scan of the raw leaf for
`(^|[?&;])(u|p|t|s|apiKey)=`, and treat any hit as a finding. Add one malformed-query row to the perturbation table.

### C11 (minor): xmllint's output streams are not quoted, and the version check rests on them

`plan:206`, `:213`, `:443-445`, `:472-473`.

**Defect.** `xmllint --version` prints `xmllint: using libxml version 21504` on stderr and leaves stdout empty. The
plan quotes `Cmd.Output`, which captures only stdout, so a runner built from these quotes reads an empty version and
refuses the real binary. A fake `xmllint` that prints on stdout would then pass the unit proof. The plan also never
says which stderr lines count as errors: the trailer `<file> fails to validate` and the success line
`<file> validates` both arrive on stderr. An uncompilable schema exits 5.

**Fold.** Add these to the quotes table. Read the version from combined output or from stderr, and make the fake print
on stderr. Parse only lines that match `<file>:<line>: Schemas validity error : `. Under exit 3, the trailer is
expected, and any other unmatched line is a tool failure. Also quote `encoding/xml.EscapeText`'s newline escape (or
count emitted newlines), since a text node must not shift the line-to-ordinal table.

### C12 (minor): smaller fixture gaps in Task 3 and Task 4

`plan:463-465`, `:466-467`, `:475`, `:504-507`.

- The 49-additions measurement "names any count that differs", so it asserts nothing. Pin the computed per-type
  additions as a golden, and let the report reconcile the golden with 49.
- The mutation "key the scope on the element name" survives an implementation that unions additions across every type
  sharing a name. Under that union, `mediaType` is dropped from a `Child`-typed `album` and the proof passes. Add a
  case where `isCompilation`, an `AlbumID3` addition absent from `Child`, on an `album` under `albumList` is a residual
  violation.
- The XSD has four types (`Captions`, `AudioTrack`, `VideoConversion`, `ArtistInfoBase`) with no same-name OpenAPI
  schema and no rename entry. State that such a type has no additions, so any extra name surfaces as a residual.
- Task 4's XML twin asserts "exactly `xmlrule`'s violations for that body". If that list is empty, the
  skip-XML mutation survives. Perturb the body with one known residual (`titel`) and assert the count is non-zero.

### C13 (minor): L-51's `test` text keeps the rejected branch of fork C

`plan:569-571`, `:589-591`.

**Defect.** The fidelity exemption covers only the `dubplate` column. L-51's "Test owed" reads "Harness: folder rows
graded by gonic, only if fork C is yes", and the plan carries it verbatim into a row ruled the other way.

**Fold.** Give L-51's `test` the ruled text too: the contract tier and the spec, an empty `getIndexes`, error 70 from
`getMusicDirectory`, slice 3. Add it to the test's named exemptions.

### C14 (minor): Task 11 lets a pass claim go without evidence, and its class sets the wrong review bar

`plan:290-292` (R25), `:745-747`, `:773-775`.

**Defect.** R25 says each step records pass or fail "with a screenshot path". Task 11's Files keep screenshots "only
if the record cites them", so a step can read "pass" with nothing a reviewer can check. The `docs` class renders the
register chain as the review bar (pass-core's table). The record is evidence, not published prose, and the review the
plan actually wants is a coverage and evidence read.

**Fold.** Every step cites a screenshot, or a DOM capture over Feishin's debugging port. The diff-reviewer checks each
pass claim against its capture. Keep the `--reduced` gate. Either label the task `tool` with no test mandate, or keep
`docs` and state that the register chain does not run.

### C15 (minor): Task 0 assumes a fast-forward the branch cannot have

`plan:336-337`. The branch already carries `0ae2b54`, and will carry the review folds, so "a fast-forward, since the
branch has no commits yet" is false. The rebase is trivial, since the plan commit touches only docs. Say "a rebase of
the plan commits" so the conductor does not read a non-fast-forward as a stop signal.

## Over-ceremony

### O1 (minor): Task 11 walks each desktop client against both oracles

`plan:759-761`. Four GUI walks, driven through screenshots, form the costliest per-step work in the pass (the plan
prices a 0.3M premium). Slice 1 already walked Feishin and Rufin live. 2b's deliverable is the script, and one oracle
proves a script runs. Gonic changes the expected answer for only one step on the 0.8 surface, the folder view.
Recommendation: walk each client against Navidrome, and against gonic only for the folder-view step. That halves the
GUI time. The cost of being wrong is small, because every later slice walks the same scripts against dubplate.

## Not found

- **Pass class.** `tool` fits Tasks 1 to 10, including the gate edit (Task 1), whose shell evidence follows CLAUDE.md
  ("Shell … no behavior tests"), and the credential guard (Task 7), which the Opus diff-reviewer covers.
- **Ceiling arithmetic.** It holds (8.55M with one fix round of headroom).
- **Seed-golden stability.** Task 5's "an oracle pair never consults the assignment" leaves the seed golden alone,
  because that golden reads `seed-ledger.json`, not `data/ledger.json`. Task 5 does have to update
  `TestCommittedLedger`'s owed counts (`ledger_test.go:474`), and that file falls inside its "its tests" scope.
- **Album shapes.** Task 5's album-shape checks are satisfiable: the seed manifest carries a multi-disc album, a
  `Various Artists` album, a track with a list-valued `artist`, and three non-ASCII albums.
