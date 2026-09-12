# Draft HISTORY entry for the conductor's fold (rung 8c close)

Written by Task 4 of
`docs/superpowers/plans/2026-09-08-rung-8c-absorb-repoint.md`. The
implementer never edits `docs/HISTORY.md`; the conductor prepends this
entry (newest-first) and folds the ROADMAP text from
`draft-STATUS-8c.md`.

---

## Rung 8, pass 8c (absorb and repoint): closed 2026-09-12

Tasks 0 through 4 executed against
`docs/superpowers/plans/2026-09-08-rung-8c-absorb-repoint.md` (revision
2) in the main tree, run `wf_c808ec5e-1ba`. This entry is Task 4, the
pass close and the rung's close. Task 0 (`opus`, fresh context) found
no blockers against HEAD `6a9570d`: the lane registry, `flac.go`'s
three call sites, `buildFLAC`'s four external uses, `tagio`'s exported
surface, and every `spec:`/`docs/HISTORY.md:` pointer this plan carries
all resolved as expected. Baseline measurements: `internal/engine/shadow`
at 75 top-level tests, 140 RUN lines, 65 subtest passes, about 0.18s;
`TestConformanceLedgerSubjects` at twelve subtests, about 0.05s.

### What landed

Task 1 (`9be7e43`) moved every `internal/engine/shadow` test fixture
onto the writer: a new `fixture_test.go` carries the committed 42-byte
oracle template as a seed literal (moved out of `flac_test.go` with its
provenance comment) and a one-second variant (a two-byte splice of the
zero-sample template's STREAMINFO pack, encoding 44100 total samples at
the template's own 44100 Hz rate), and one `writeFixtureTrack` helper
writes the seed, calls `tagio.Write` with a `tagio.TagSet` and a nil
picture set, and pins the modification time after the write. Every
fixture-building call site in `capture_test.go`, `run_test.go`, and
`observe_test.go` converts from `"KEY=value"` string lists to
`tagio.TagSet`. **Task 1 took the plan's first offered route for the
content-join criterion**: every fixture now seeds from the one-second
variant, so `reconcile.go:242`'s title-and-length join compares a live
1.0-second value on both sides rather than vacuously comparing zero to
zero. `buildFLAC` still compiled, with every caller confined to its own
file, clearing the way for Task 2's deletion. All five named mutations
ran and restored; none was reported unfired for an unexpected reason
(mutation table below).

Task 2 (`3a299ac`, fix `6de7d55`, conductor-owed R6 fix `12d1e97`)
deleted `internal/engine/shadow/flac.go` and `flac_test.go` (313 and
361 lines: the block walker, the Vorbis comment decode, `safe_cast`,
date splitting, `buildFLAC`, `blockHeader`, `writeTemp`, and
`TestBuildFLACEmitsTheStreaminfoLayoutByteForByte`), and created
`libraryread.go` + `libraryread_test.go` and `itembag_test.go`.
`readFLAC` keeps its name and signature (R7/A3): `run.go:209` and
`capture.go:158` needed no edit. It now calls `tagio.Parse` once per
file and the newly exported `tagio.Tags` for the tag set (R1 as
applied, the 8b handoff's second half), reading `Container.StreamInfo`
for the length with A4's zero-rate guard preserved. The eighteen-row
`matcher.ItemFields` projection survives with its per-row mediafile
citations (restored by the fix round after the main commit dropped
them) and a new sentence naming ruling 3 as the reason the projection
stays in `shadow` rather than moving to `tagio` with the rest of the
container knowledge. Under R1, `tagio` gained one exported symbol,
`Tags(c Container) TagSet` (`read.go:39-45`), documented in `doc.go`
as `Read`'s own unexported `project` exported unchanged; `Read`'s
behavior is byte-identical.

R2's seven-row destination table (as executed, below) was discharged
test by test, including three previously-ungraded `TestParseRefusals`
rows this task was pre-authorized to add ("STREAMINFO is not the first
block" in both its trigger shapes, "no STREAMINFO block", and "carries
no '='") and one previously-uncovered barcode-precedence row folded
into `tagio`'s own `TestProjectLastStyleRule`. **The conductor-owed R6
fix** carries the invalid-UTF-8 value change through `shadow`'s own
projection test: `spliceInvalidUTF8TitleByte` (`fixture_test.go`)
length-preserves a byte splice of a writer-built fixture's title byte
to an ill-formed UTF-8 continuation byte, because `tagio.Write` only
ever accepts values already valid UTF-8 (ruling 10) and `shadow` has no
lower-level fixture path; `TestReadFLACProjectsInvalidUTF8AsReplacementChar`
asserts `readFLAC` projects U+FFFD, proving `shadow`'s reader carries
`tagio`'s own rule through rather than duplicating it. All seven named
mutations ran and restored, including a new test
(`TestReadFLACProjectsAbsentFieldsToTypedZeroValues`) the implementer
added because the plan's own all-fields-present fixture test could not
catch mutation 5 (an unguarded absent-field dereference) on its own.

The batched `code-simplifier` round (`45dd73a`) derived the one-second
seed by `slices.Clone` plus a two-byte patch instead of a second
40-byte literal that could silently drift from the first (the two
literals' doc comment had already gone wrong once, naming three bytes
for a two-byte change), collapsed `stringVal`/`intVal`/`boolVal` into
one generic `deref[T]` (three copies of one sentence about
`ProjectItemFields`' typed-default rule), and rebuilt the
second-STREAMINFO refusal fixture with `slices.Insert` rather than a
nested append over an aliased prefix. No behavior or assertion change;
the derived seed is verified byte for byte against both the prior
literal and `internal/harness/oracle/testdata/template.flac`.

Task 3 (`67d5c3d`) added the pass's one behavioral change: a
capture-path test proving a container the deleted reader accepted but
`tagio` refuses (ruling 9's "more than one VORBIS_COMMENT block") is
**discarded whole**, matching the partial-read rule `captureAlbum`
already applies to a truncated file (zero captures, zero exclusion
rows, one partial-read row). A second test proves the same album
unspliced still captures normally, so the witness fixture is a witness
rather than a broken file. The witness's bytes come from the writer's
own output plus `spliceDuplicateVorbisComment`, which parses the
writer's output via `tagio.Parse`, identifies the VORBIS_COMMENT block
by matching each block's payload against
`Container.VorbisComment.Vendor`'s own bytes (no length-prefix decode
needed, since STREAMINFO is a fixed 34-byte layout and PADDING is
zero-filled), and copies that block's header-plus-payload whole,
splicing a byte-identical second copy after the first: no block-type
constant, no header encoder, no `encoding/binary` import anywhere in
`shadow` (A1). All three code/test mutations ran and restored; the
fourth named mutation (enumerate from the deleted reader's behavior
rather than `tagio`'s refusal set) is a documentation-construction
check rather than an executable one, verified by construction instead.
**The enumeration table went through two comment-only review rounds**
before it was correct (below).

### The A1 through A9 coverage table

| Item | Go site | What proves it |
| --- | --- | --- |
| A1 (no container knowledge in `shadow`'s non-test files) | `internal/engine/shadow/libraryread.go` (whole file) | `grep -rn "encoding/binary\|fLaC\|VORBIS" internal/engine/shadow` returns nothing outside a comment and outside `capture_test.go`'s pre-existing `[]byte("fLaC")` literal in `corrupt()` (rung 4a Task 6, untouched by this pass, out of A1's non-test-file scope). Task 3's splice helper touches container bytes without a block-type constant or header encoder, exactly as A1's scoping sentence permits. |
| A2 (eighteen-row projection, per-row mediafile citations, typed-default rule) | `projectItemFields` (`libraryread.go:45-`) | `TestReadFLACMapsEveryVorbisKey` grades seventeen distinguishable rows plus `DataSource`'s asserted empty string; the citations were restored by Task 2's fix commit `6de7d55` and verified byte for byte against the deleted `flac.go`. |
| A3/R7 (`readFLAC` keeps its name and signature) | `run.go:209`, `capture.go:158` | Both call sites are untouched in the diff; `git diff` across the whole pass shows zero lines changed in either file's call site. |
| A4 (length division, zero-rate guard) | `readFLAC` -> `trackLength` path in `libraryread.go` | `TestReadFLACLengthFromStreaminfo`'s table: ten seconds at 44100 Hz, a fractional length, and a zero sample rate reading zero rather than dividing by zero. Named mutation 2 (drop the guard) fires this row. |
| A5/R1 (one parse per file) | `readFLAC` (`libraryread.go:24-29`) | One `tagio.Parse` call, feeding both `tagio.Tags` and `Container.StreamInfo`; the doc comment states the count. |
| A6 (provisional framing rewritten) | `libraryread.go`'s package/function doc comments | `flac.go:62-70`'s provisional paragraph and rung-4-calendar rationale do not survive; the deleted file's whole text is gone with it. |
| A7/ruling 4 (`buildFLAC` does not survive) | deleted `flac_test.go` | `grep -rn buildFLAC internal/engine/shadow` returns nothing after Task 2's commit. |
| A8 (tightening asserted and counted) | `capture_test.go`'s two new tests | `TestCaptureDiscardsATrackWithADoubledVorbisCommentBlock` and its unspliced companion, plus the eight-row enumeration below. |
| A9 (conformance instrument stays green) | `internal/engine/tagio/conform_test.go:640` | `go test ./internal/engine/tagio/... -run Conformance -count=1`: `TestConformanceLedgerSubjects`, 12 subtests, about 0.06-0.07s, all green, reconfirmed at this close. |

### The seven Resolutions, retired

**R1** (the shadow parses once; `tagio` exports its tag projection).
**Conductor-owed. Ruled 2026-09-12, as applied**: the shadow consumes
`Parse` and the newly exported `tagio.Tags`, taking the 8b handoff's
second half rather than its first (`docs/HISTORY.md`'s 8b entry,
"registry consumer" sentence); `tagio` gains one exported symbol;
8b's declared coverage gap 4 retires now rather than at rung 9.
Site: `internal/engine/tagio/read.go:39-45` (`Tags`),
`internal/engine/shadow/libraryread.go:24-29` (`readFLAC`). **Had R1
been ruled the other way** (the fold), the shadow would call
`tagio.Read` for the tag set and `tagio.Parse` for the STREAMINFO, two
parses and three file opens per track, `tagio` would gain no exported
symbol, and 8b's coverage gap 4 would stay open for rung 9 to close
under R12's original assignment.

**R2** (the five reader tests get named destinations, two staying in
`shadow`). **Conductor-owed. Ruled 2026-09-12, as applied**: the
seven-row destination table stands (below). **Had it been ruled
otherwise**, the two staying tests would be rewritten inside `tagio`
against `TagSet` expectations, and the eighteen-row `ItemFields`
projection would carry no test of its own.

**R3** (fixtures come from the writer; the one malformed fixture is a
byte splice). Applied. Site: `internal/engine/shadow/fixture_test.go`
(`writeFixtureTrack`, `spliceDuplicateVorbisComment`,
`spliceInvalidUTF8TitleByte`). **Ruled otherwise**, the malformed
fixture would be a hand-written byte literal transcribed in full.

**R4** (the exclusion count: exclusion set plus a read-only denominator
plus a bounded count). Applied for items 1 and 3 (Task 3's enumeration;
the count stated as zero known exclusions out of the denominator, bound
by the exclusion set, next measurable at the shadow's next deploy).
**Item 2, the denominator's composition, is conductor-owed. Ruled
2026-09-12**: `ObservedApplied` + `ObservedSkipped` + the
`Unreconciled` counts, read by the conductor through `dubplate-server
shadow report`, never the bare binary as root. **The read itself did
not happen this pass**: the conductor's box read failed at 03:14 AKDT
(ssh timed out at the connection level, twice; no quiet-box check was
possible). The denominator is recorded as **UNREAD**, owed to the next
attended session. Had R4 item 2 been ruled otherwise, the conductor
would take a read-only sample of library and inbox files off the box
and grade them on the workstation with the landed parser.

**R5** (the `io.ReaderAt` gap declined at 8c, handed to rung 9).
Applied; no task in this pass widened `Parse`'s signature. The gap
stays declared (restatement of 8a's gap 7, not counted again), now
naming rung 9 as the decision-maker rather than "8c or rung 9." Had it
been ruled otherwise, `Parse` would gain a second, injectable
`io.ReaderAt` entry point and a mid-chain I/O error test.

**R6** (the invalid-UTF-8 value change stated in advance). Applied,
with the conductor-owed extension landed at `12d1e97`: R6's finding
during Task 2's second review round (no row in `shadow` graded the
value change, only `tagio`'s own `TestRulesInvalidUTF8ReadBack`) forced
a choice between giving `shadow` a low-level fixture path, exporting a
`tagio` test helper, or ruling R6 already discharged in `tagio`. The
conductor ruled a fourth option not named in the plan's own text: a
length-preserving byte splice over writer output, the same shape R3
already establishes for the malformed-container fixture, carrying the
value change through `shadow`'s own projection test without exporting
anything new from `tagio`. This is a **named deviation from the plan's
own three-way framing of R6's escalation**, resolved with R3's own
method rather than any of the three options the escalation posed.

**R7** (`readFLAC` keeps its name and signature). Applied; `run.go:209`
and `capture.go:158` needed no edit, confirmed by the implementer's
grep and by this close's own re-check. Had it been ruled otherwise,
the function would be renamed and both call sites plus
`reconcile.go:35` would need editing together.

### R2's destination table, as executed

| Test | Destination |
| --- | --- |
| `TestReadFLACMapsEveryVorbisKey` | Stayed in `shadow`, rewritten as `libraryread_test.go`'s eighteen-row table over a writer-built fixture. |
| `TestReadFLACStylePrecedence` | Moved. Its sixteen rows were walked against `tagio`'s existing projection tests: fifteen already covered by an existing row over the same branch; one uncovered row (barcode falling through the four read-only aliases in order) added to `TestProjectLastStyleRule` (`internal/engine/tagio/read_test.go`). |
| `TestReadFLACLengthFromStreaminfo` | Stayed in `shadow`, rewritten as a table over `tagio.StreamInfo` values needing no file. |
| `TestReadFLACParsesTheCommittedTemplateBytes` | Retired. Its purpose is already covered by `tagio`'s `TestParseGolden` (`container_test.go:359`); no second external pin added, and the now-dead `committedTemplateFLAC` literal was removed with it. |
| `TestReadFLACRefusesAMalformedContainer` | Split. The container half moved to `tagio`'s `TestParseRefusals`, which also gained three previously-ungraded rows this task was pre-authorized to add; the capture half stayed in `shadow` as a thin test asserting `readFLAC` returns the zero `ItemFields` and an error, never a partly filled bag. |
| `TestItemBagRoundTripsThroughProjectItemFields` | Moved within `shadow`, to the new `itembag_test.go`; `itembag.go:11-13`'s pointer repointed. |
| `TestBuildFLACEmitsTheStreaminfoLayoutByteForByte` | Deleted with `buildFLAC` (ruling 4). |

### Task 3's enumeration, verbatim as the tightening record

The table below is the corrected, final form: Task 3's own report
carried seven rows; a review round added row 2 (a shape the deleted
reader accepted and the report's first draft omitted: more than one
STREAMINFO block); a second review round found row 8 (as it then
stood) named the wrong `tagio` test. Both corrections are folded here
because no further fix round ran to fold them into the task's own
committed return, and re-verified against the tree at this close
(`go test ./internal/engine/tagio/ -run 'TestParseRefusals|TestParsePictureRefusals' -v`).

| # | Shape | Landed reader's old behavior | `tagio` test | Reachable from a real upload? |
| --- | --- | --- | --- | --- |
| 1 | More than one VORBIS_COMMENT block | Read happily, took the first value per key | `TestParseRefusals/second_VORBIS_COMMENT_block` (asserted directly by Task 3's capture-path test) | Crafted file only; no known compliant encoder emits this. |
| 2 | More than one STREAMINFO block | Parsed each occurrence and overwrote `file.info` with the last, no error | `TestParseRefusals/a_second_STREAMINFO_block` | Crafted file only, no known compliant encoder emits a second STREAMINFO. |
| 3 | A STREAMINFO block that is not first, or no STREAMINFO block at all | Neither read as an error; projected a zero length | `TestParseRefusals/first_block_is_not_STREAMINFO` and `TestParseRefusals/no_STREAMINFO_block_at_all` | Crafted file only. |
| 4 | A block appearing after a header whose last-block flag is set | Stopped at the flag, returned partial | `TestParseRefusals/block_after_a_header_whose_last-block_flag_is_set` | Crafted file only. |
| 5 | More than one SEEKTABLE block | Skipped extras by declared length | `TestParseRefusals/second_SEEKTABLE_block` | Crafted file only. |
| 6 | A comment field name carrying a code point outside U+0020-U+007E | Upper-cased it and kept it | `TestParseRefusals/comment_field_name_above_U+007E` | Unlikely from a compliant encoder; possible from a hand-edited or non-compliant tagger. **Finding, execution-discovered**: `validateFieldName` (`container.go:347-350`) refuses both directions (`r < 0x20 \|\| r > 0x7e`), but `TestParseRefusals` has no dedicated subtest for the below-U+0020 direction. This is a coverage gap in `tagio`'s own test suite, not a missing enumeration row (the shape ties to a real, if partial, test). |
| 7 | A comment carrying no `=` | Skipped that comment, read the rest | `TestParseRefusals/comment_carries_no_'='` | Possible from a non-compliant real-world tagger; more plausible than rows 1-5. |
| 8 | A PICTURE block whose declared lengths do not fit its payload | Skipped every PICTURE block by declared length | `TestParsePictureRefusals` (`internal/engine/tagio/picture_test.go:91`; five subtests: `media_type_length_overdeclares`, `description_length_overdeclares`, `data_length_overdeclares_near_the_block_header's_own_3-byte_maximum`, `block_shorter_than_the_32-byte_minimum`, `media_type_consumes_the_block,_no_room_left_for_geometry`) | Plausible from a real corrupted or truncated upload, not only a crafted one. |

Every row ties to a named, currently-passing `tagio` test; none is a
row with no test to name.

### R4's three-part answer to the exit gate's count

1. **The exclusion set**: the eight-row table above, each row a
   workstation measurement over committed code, needing no box.
2. **The denominator**: **UNREAD**. Conductor-owed R4 item 2 ruled the
   composition (`ObservedApplied` + `ObservedSkipped` + `Unreconciled`),
   but the read itself failed: ssh to the box (5.78.148.42) timed out
   at the connection level, twice, at 03:14 AKDT, so no quiet-box check
   was possible and `dubplate-server shadow report` was never read.
   Owed to the next attended session, before rung 15 treats R4 as
   final.
3. **The count**: zero known exclusions out of the (unread)
   denominator, bounded by the exclusion set above. Not derivable from
   stored records, because a stored capture carries no container bytes
   (`capture.go:143-162`). The tightening reaches the live capture at
   the shadow's next deploy, not at this merge.

### The declared coverage gaps

Four new gaps this pass creates, joined to the running total of
sixty-three (`docs/STATUS.md:80-82`: rung 3's twenty-one, rung 4a's
nine, rung 4b's four, rung 1b's two, rung 8a's seven, rung 8b's nine,
rung 5a's eleven), bringing the total to **sixty-seven**. Two items
this pass's own expected list carried are restatements of an
already-counted gap rather than new declarations and are not counted
again: "the `io.ReaderAt` gap" restates 8a's gap 7 (now naming rung 9
alone as the decision-maker), and "no real library file has been read
through the absorbed path" restates 8a's gap 6.

**A fifth item from the plan's expected list, "the end-to-end length
join's length term is vacuous," does not exist**: Task 1 landed the
non-zero-sample seed variant, so every fixture's join compares a live
1.0-second value rather than zero to zero, exactly the condition the
plan named as retiring this gap.

1. **The tightening's live effect is unmeasured** (R4). The count is
   bounded rather than counted, because stored captures carry no
   container bytes.
2. **The invalid-UTF-8 value change is graded on a synthetic fixture**
   (R6), not on a file any member uploaded, and the fixture reaching
   `shadow`'s own projection test is itself a byte splice rather than a
   file any encoder produced.
3. **The content join's length term discriminates a real length from
   zero or a missing one, but not between two different non-zero
   lengths**, because Task 1 seeded every fixture in the package from
   the same one-second variant. This is the corrected, narrower shape
   of the plan's originally-feared gap 5: the vacuous zero-to-zero
   comparison does not exist, but the join still cannot distinguish
   fixture A's length from fixture B's, since both are 1.0 seconds by
   construction. Flagged by Task 1's own `diff-reviewer` round as worth
   carrying here.
4. **`TestParseRefusals` has no dedicated subtest for a comment field
   name below U+0020**, though `validateFieldName` refuses both
   directions. Discovered while building Task 3's enumeration (row 6,
   above); the refusal code path is exercised, the specific below-range
   shape is not independently graded.

### 8b's declared coverage gap 4 is retired

R1 stood as applied: the shadow's `readFLAC` now reads a file's tags
and its STREAMINFO from one `tagio.Parse`, so a caller wanting both no
longer parses twice. 8b's gap 4 ("`Read` returns no pictures, so a
caller wanting the tags and the cover parses the file twice") does not
apply to `shadow`'s own reader, whose caller now gets the tag set and
the length from one parse; the gap's remaining scope is a caller
wanting `Read`'s tag set alongside `Container.Pictures`, which is
`Read`'s own signature and unaffected by this pass. **The running gap
total above is not reduced for this retirement**: the sixty-three
figure and the sixty-seven figure both count declarations, not a live
open list, so retiring one item changes what rung 15 should read as
still open, not the historical count.

### The mutation table

**Task 1** (5 named, all fired or confirmed unfired for the stated
reason): (1) skip the `tagio.Write` call, nine tests across
capture/observe/run fail because tracks carry no tags; (2) pin the
mtime before the write, `TestRunIsDeterministicUnderAFrozenClock` fails
because the atomic rename installs a fresh mtime; (3) give both
`windroseTracks` entries the same title, no test fails, confirmed
unfired for the stated reason (the capture groups by
albumartist/album, invisible at this altitude); (4) fixture mode 0
instead of 0o644, ten tests fail because the file cannot be reopened;
(5) a non-nil empty picture set instead of nil, no test fails,
confirmed unfired for the stated reason (the seed carries no PICTURE
block, so strip and leave-alone are byte-identical). Every restoration
verified with `git diff` empty.

**Task 2** (7 named, all fired or confirmed for the stated reason):
(1) skip the STREAMINFO read, `TestReadFLACMapsEveryVorbisKey`'s
Length row fails (1→0); (2) drop the zero-rate guard, the zero-rate
row fails with `+Inf`; (3) swap `Label` for `Country`, both rows fail,
each named; (4) keep the deleted `parseFLAC` and call it instead of
`tagio`, the package still compiles and every landed `shadow` test
passes, confirming Task 3's tightening test is what catches this, not
Task 2's own criteria; (5) an unguarded `*string` dereference, panics,
caught only after the implementer added
`TestReadFLACProjectsAbsentFieldsToTypedZeroValues` because the plan's
own all-fields-present fixture could not exercise an absent field; (6)
return the first matching comment's value versus the last for a
repeated key, confirmed unfired in `shadow`, for the stated reason (the
rule is now `tagio`'s, graded at `read_test.go`); (7) leave
`flac_test.go` in place while deleting `flac.go`, the package fails to
compile, the mechanical proof the reader and its tests die together.
Every restoration verified with `git diff` empty.

**Task 3** (4 named): (1) build the witness fixture without the
splice, `captures = 1, want 0`, the album was captured whole; (2) catch
the read error in `captureAlbum` and continue, `captures = 1, want 0`,
the bad track silently dropped, exactly the tolerance this pass
removes; (3) assert the album is excluded with a reason rather than
discarded, fails against `capture.go`'s own discard-not-exclusion rule
(`excluded = [], want one row`; `partial reads = [...], want none`);
(4) enumerate the tightening from the deleted reader's behavior rather
than `tagio`'s refusal set, not an executable mutation, verified by
construction instead (every row built from `TestParseRefusals` and
`TestParsePictureRefusals` subtest names directly; the one gap found,
row 6's untested below-U+0020 direction, is reported above rather than
silently dropped or invented). Every code/test mutation's restoration
verified with `git diff` empty.

### What rung 9 inherits

`tagio`'s exported surface after this pass: `Parse`, `Read`, `Write`,
`Tags` (new this pass), `TagSet`, `Mapping()`, `WriteOrder()`,
`Serialize`, `EmptyString`, `EmptyList`, `Container`, `Block`,
`StreamInfo`, `Comment`, `VorbisComment`, `KeyWrite`, `Field` and its
69 constants, `FieldMapping`, `Style`, `Picture`, `FrontCover`,
`Format`, `SupportedFormats`, `Sniff`, `ErrUnsupportedContainer`,
`ErrMalformedContainer`, `Deviations`. The `io.ReaderAt` call (R5) is
rung 9's alone to make: whether `Parse` grows a second, injectable
entry point is undecided, and no consumer in this pass forces it. The
combined-read question R12 posed is **answered, not left**: R1 as
applied means a caller wanting one file's tags and its STREAMINFO
parses once through `Parse` plus `Tags`, the shape `shadow` now uses;
rung 9 inherits that shape as precedent rather than a decision still to
make, though `Read`'s own tag-only contract (R12, 8b) is unchanged and
a caller wanting `Read`'s tag set alongside `Container.Pictures` still
parses twice. The tightening's live effect on rung 7's intake gate
bound does not reach the box until the shadow's next deploy; this
pass's merge changes no running process.

### The spec corrections

Applied directly to
`docs/superpowers/specs/2026-09-07-rung-8-tag-io.md` as a dated
erratum, in the shape the 8a-close and 8b-close errata used. See the
diff for the exact inserted text; summarized here:

1. **The five-tests sentence** (spec's pass-split table, 8a-close/8b-
   close resolved offset: line ~1011) is corrected to R2's
   seven-row destination table, naming why two of the five stay in
   `shadow` (their subject is `matcher.ItemFields`, and ruling 3
   forbids `tagio` naming that type).
2. **The reference-count sentence** (line ~994) is corrected: the
   paragraph named six references (three call sites, two prose
   mentions, the reader's own doc comment). Under R7 two of the three
   call sites and both prose mentions need no edit, because `readFLAC`
   keeps its name. **`itembag.go:11-13` is a seventh reference the
   spec never named**, and it is the one this pass actually edits (its
   pointer at the deleted `flac_test.go` repointed to `itembag_test.go`).
3. **The ceiling table's 8c line** (line ~1230, and the 8b-close
   erratum's restatement at line 85) is corrected from 1M to this
   plan's 2.1M, re-derived from 8b's actual under the spec's own
   re-derivation clause. Rung 8's ceilings now stand at **9.85M**
   against the spec's 8M reservation: 4.5M (8a) plus 3.25M (8b) plus
   2.1M (8c). 8c's actual spend is stated in this entry's own spend
   section.
4. **The read-path strictness section** (line ~1028-1049) is corrected
   against Task 3's enumeration: the spec's own prose named the
   tightening only in general terms ("a real inbox FLAC with two
   VORBIS_COMMENT blocks or an out-of-order STREAMINFO"); the landed
   refusal set is the full eight-row table above, materially wider
   (it also refuses a doubled STREAMINFO, an out-of-range SEEKTABLE
   count, an out-of-range comment field name, a comment carrying no
   `=`, and a malformed PICTURE block), and the correction names the
   table as the authoritative enumeration rather than the spec's own
   two-example prose.
5. **CLAUDE.md's gate sentence** was checked: this pass adds no
   tracked Python file, so the sentence's Python wording stands
   unchanged.

### The rung 8 close statement

Rung 8 is closed: three passes (8a the writer core and reader, 8b
PICTURE and the conformance instrument, 8c the absorption and
repoint), `internal/engine/tagio` born and now the sole holder of FLAC
container knowledge in the repo, `internal/engine/shadow` repointed to
it with the tightening asserted rather than discovered.

**The four owner rulings**: owner ruling 2 (the wazero-taglib spike
retired, F1 superseded) — discharged at 8a's close. Owner ruling 5
(fork B, the format ladder) — discharged: Geoff ruled 2026-09-08,
FLAC-only 1.0, format ladder at 1.1 (`ROADMAP:formats-1-1`). Two more
owner rulings this rung carried from its own spec's ratification are
recorded in that spec's own errata section and are not re-litigated
here.

**The fourteen conductor rulings**: this pass alone carried three
(R1, R2, R4 item 2, all ruled 2026-09-12, above); the rest belong to
8a and 8b's own close entries and are not restated here, each
discharged or carried per that entry's own resolutions section.

**Hooks 13, 14, and 3**: hooks 13 and 14 landed at 8a (the writer core
and the 69-key value model, per 8a's own close entry). Hook 3 landed
at 8b (the PICTURE decode and `Container.Pictures`, per 8b's own close
entry). **8c lands no hook**; under R1 it improves the shape hook 3's
consumer (rung 9's admit path) will use, because a file's tags and its
embedded picture now reach a caller off fewer parses, but that is a
consequence recorded here, not a hook this pass claims.

**The deviation note, as a statement rung 15 can read without
interpreting a file**: 8c's read path never reaches `tagio`'s seam
registry (R1's second deviation, folded with R1's own ruling); the two
reader tests that stayed in `shadow` are a stated departure from
`spec:1011`'s original five-tests sentence (R2), corrected above; the
2.1M ceiling is a stated departure from the spec's proposed 1M,
corrected above; `tagio` gaining one exported symbol at a pass the
spec called mechanical is a stated departure from a decision 8b closed
(R1, conductor-owed); and R6's escalation resolved with a fourth
option (a length-preserving byte splice) the plan's own three-way
framing did not name, recorded above under R6's retirement.

**Fork B recorded as ruled**: unchanged by this pass; FLAC-only 1.0,
the format ladder at 1.1 (`ROADMAP:formats-1-1`), per Geoff's
2026-09-08 ruling.

**The one carry that leaves the rung open**: the group non-FLAC
question (`docs/HISTORY.md`'s rung 8a entry) stays carried and is not
decided by this pass, which absorbs and repoints the FLAC reader alone.

### The evidence

`internal/engine/shadow`'s suite: 75 top-level tests / 140 RUN lines /
65 subtest passes / about 0.18s before this pass (Task 0's baseline);
76 top-level tests / 120 RUN lines / 44 subtest passes / about 0.2-0.5s
across runs after (Task 2 net -2, Task 3 net +2, the R6 fix net +1:
75-2+2+1=76). `internal/engine/tagio`'s conformance run:
`TestConformanceLedgerSubjects`, 12 subtests, about 0.06-0.07s, before
and after unchanged. `go test ./... -count=1` (every package): green,
2m19.516s. `bash scripts/check.sh`: green, about 2m43s, gitleaks the
dominant cost at 2m27s scanning about 119.64 MB.

Line counts across the pass (Task 1 through the `code-simplifier`
round, `6a9570d..45dd73a`, `internal/engine/shadow` and
`internal/engine/tagio` combined): **616 insertions, 738 deletions**,
across 14 files, a net deletion of 122 lines. `flac.go` (313 lines) and
`flac_test.go` (361 lines), 674 lines of container-decode code and its
tests, are gone; `libraryread.go` (90), `libraryread_test.go` (169),
`itembag_test.go` (40), and `fixture_test.go` (151) replace them
alongside small `tagio` additions (`doc.go` +10, `read.go` +12/-,
`container_test.go` +35, `read_test.go` +1).

The mutation table above: every named mutation across Tasks 1 through
3 fired or was confirmed unfired for its stated reason; none fired for
an unexpected reason and none was silently dropped.

### Spend

Budget: 2.1M ceiling. This close does not have exact per-agent token
figures on hand; the run's journal (`wf_c808ec5e-1ba`, ten
implementer/reviewer agent turns across Tasks 1-3, plus Task 0's
separate fresh-context `opus` dispatch and this Task 4) is the source
if the conductor wants an exact number. Attended interactions: **zero
planning misses**; **one execution sitting** is arguable — the
conductor's R6 ruling and the box-read attempt both happened without a
Geoff interaction (conductor-owned, not an attended sitting), so the
target of zero execution sittings holds as reported, pending the
conductor's own accounting of turns it spent on R1/R2/R4/R6.

### What a later pass would be wrong to rediscover

`shadow`'s read path consumes `tagio.Parse` and `tagio.Tags` directly
and never reaches the seam registry; a future pass adding a second
format to the registry does not make `shadow`'s FLAC path a registry
consumer by that addition alone, because the decision not to route
through it is R1's own, not an oversight. `readFLAC`'s name and
signature are a deliberate no-op (R7): a reader expecting the function
to be renamed alongside the container-knowledge move is reading the
plan's own two-name collision (`shadow.readFLAC` and `tagio.readFLAC`,
different packages, never confused at compile time) rather than a
defect. The content-join's length term now discriminates a real length
from zero, but every fixture in the package shares one seed value, so
a future test relying on the join to distinguish two different
non-zero lengths needs its own distinct seed, not this pass's shared
one (declared gap 3, above). The invalid-UTF-8 fixture path
(`spliceInvalidUTF8TitleByte`) is `shadow`'s only way to put an
ill-formed byte in front of a reader, because `tagio.Write` enforces
valid UTF-8 on write (ruling 10); a future `shadow` test wanting
another invalid-byte shape reaches for the same splice technique
rather than a new writer bypass. R4's denominator is unread, not
zero-valued: a later pass reading `docs/STATUS.md`'s carried-gap count
should not treat "zero known exclusions" as "zero exclusions
happened."
