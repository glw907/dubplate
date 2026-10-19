# Adversarial review: the rung 8 pass 8c plan

Fresh-context review (2026-09-08) of
`plans/2026-09-08-rung-8c-absorb-repoint.md` revision 1, in one lens:
absorption correctness and plan integrity, reading the plan against the
rung 8 spec's 8c row and its rulings, HISTORY's 8a and 8b entries, the 8b
plan review's verified-sound list, and the tree at `6a9570d`, with every
counted claim recounted and every pointer re-resolved. Folded as revision
2 the same day; the 21-item "Verified sound" list below is the
do-not-re-litigate record for the pass, alongside the rung 8 spec's own
records.

The lens raised thirteen findings and no two of them name the same
defect, so nothing merged. The result is renumbered B1 through B3, M-a
through M-c, and m1 through m7. Every finding was re-verified by the fold
editor against the tree before folding, and every mechanism claim about
the toolchain, `tagio`'s exported surface, or the committed template's
bytes was measured on go1.27.1 rather than carried. Three reviewer
measurements did not survive that check in the form they were written and
are corrected below. None of the three retires its finding, and one of
them replaces a proposed fold's method with a workable one. No finding
was rejected outright. Three folds are conductor-owed rather than folded
silently, and each is applied provisionally in the plan with an inline
marker.

**Every `plan:` pointer below names revision 1's line numbers**, which is
the revision the review read. The folds moved those lines, and revision 2's
own pointers were re-resolved in the plan itself.

The lens agrees the plan is well grounded where it reasons from the spec
and from the landed tests. The 8c row's scope statement lands clause by
clause in a named task, all three exit-gate assertions appear as
acceptance criteria in the tasks that land the surfaces they grade, the
`buildFLAC` call-site inventory is exactly right at seven uses over three
files, no symbol in `flac.go` has a consumer outside `flac.go` and
`flac_test.go`, the independence marks are true by files, the ceiling
arithmetic is internally consistent and its basis is 8b's real report,
every one of the plan's own resolutions names what changes if it is ruled
the other way, and each of Task 3's seven expected tightening rows is a
container shape the landed reader really accepts and `tagio` really
refuses. The prose carries no em dash in 1,450 lines.

The defects concentrate in three places. R1's justification, which is the
plan's one exported-surface change, reads inherited permission out of
three records that all say the opposite. Four counted claims are wrong,
two of them holding acceptance criteria that cannot be satisfied as
written. And the plan asserts a concurrent lane 2 that does not exist, in
three places, one of which a task is told to repeat as the cause of a
slow gate.

---

## Corrections to the review's own evidence

**1. The lens cites `flac.go:194-199` for the reader's own DataSource
paragraph.** That paragraph is at `internal/engine/shadow/flac.go:203-207`;
`:194-199` is the tail of the field-splitting helper plus the opening of
`fields`'s doc comment. The content is exactly what the review quotes, so
the finding (M-a) stands unchanged and only the pointer moves. The fold
cites `:203-207`.

**2. The lens proposes locating the splice's VORBIS_COMMENT block
"through `tagio.Parse`'s `Container.Blocks`" with no block-type
constant.** Measured, that method as written cannot be built.
`blockVorbisComment` and its siblings are unexported
(`internal/engine/tagio/container.go:22-28`), so a `shadow`-side helper
cannot borrow one, and comparing `Block.Type` against a literal `4` is
precisely the block-type constant A1 forbids. A constant-free route does
exist, and the fold takes it instead: a VORBIS_COMMENT payload opens with
its own vendor string behind a little-endian length prefix
(`write.go:217-231`), and `Container.VorbisComment.Vendor` is exported, so
the helper identifies the block by that prefix, computes the block's byte
range as four plus the running sum of each preceding block's four plus
`Length`, and copies the range whole with its header included. Copying the
header rather than re-emitting one is what keeps `encoding/binary` and a
header encoder out of `shadow` altogether. The block is never the last one
in writer output, because the writer emits exactly one 8 KiB PADDING block
last (`spec:1166`), so the copied header's cleared last-block flag is
correct as it stands. M-c stands; its method sentence is the fold editor's.

**3. The lens's enumeration of `shadow.Report`'s rows omits
`ClockStarted`.** The struct carries eighteen fields at
`internal/engine/shadow/report.go:17-36`, and `ClockStarted` is the
eighteenth. The finding (m3) rests on the absence of a captured-album
total, and no row is one, so the correction leaves it standing.

---

## BLOCKERS

### B1. R1's justification contradicts every record it cites: the spec's own erratum, 8b's coverage gap 4, and 8b's do-not-rediscover section all hand the combined-read pick to rung 9, and grant 8c the double parse instead

**Lens.** Absorption correctness and plan integrity.

**Plan lines.** R1's "Why" (`plan:492-505`), its citation at `:498`, its
handoff paragraph at `:507-514`, and deviation 4 (`plan:1366-1369`).

**Evidence.** The plan writes that R12 "named the two shapes that close
that gap ... and handed the choice to the first caller in a position to
pick (`docs/HISTORY.md:871-876`, `spec:752-756`'s erratum). 8c is that
caller." All three records name rung 9. The spec's erratum reads "the
combined shape is handed to rung 9, the first caller in a position to
pick between a second exported read entry point and an exported
projection over a parsed `Container`"
(`docs/specs/2026-09-07-rung-8-tag-io.md:766-774`). 8b's
coverage gap 4 reads "A stated deviation from `spec:729-733`, handed to
rung 9 as the first caller" (`docs/HISTORY.md:830-833`). 8b's "what a
later pass would be wrong to rediscover" reads "`Read` staying tag-only
is a stated, ruled choice (R12), not an oversight rung 9 has to notice on
its own; rung 9 picks between a second read entry point and an exported
projection over `Container`" (`docs/HISTORY.md:990-993`), and that
section is primary input 2 of this plan (`plan:88-92`). What the 8b
handoff grants 8c is the opposite of a pick: "8c's absorbed reader may
keep reading through `Parse` and `Container.Pictures` directly ... and it
owes no fold into a combined `Read`" (`docs/HISTORY.md:871-875`).

**Consequence, built as written.** R1 is already conductor-owed, so the
conductor rules it before Task 2 dispatches. It would rule on a false
premise, reading R1 as the exercise of an inherited option rather than as
a departure from a decision 8b closed. The rung 8 close then hands rung 9
a shape rung 9 was assigned to pick, with no record that anyone noticed
the reassignment, which is the class of loss `docs/HISTORY.md`'s
do-not-rediscover section exists to prevent.

**Fix folded, conductor-owed.** R1's "Why" states the departure rather
than claiming inherited permission. It says that the 8b close assigned
the pick to rung 9 and granted 8c only the double parse, that 8c
nonetheless proposes the projection shape because the shadow is the first
caller wanting one file's tags and its STREAMINFO off one parse, and that
the conductor is therefore ruling a departure from a ruled decision
rather than exercising an option. Deviation 4 carries the same sentence,
so the departure is scored at the close. The citation at `:498` becomes
the erratum's own range, `spec:766-774` (m7).

### B2. The header claims the absorbed reader routes through the seam registry, and under R1 as applied it never reaches the registry at all

**Lens.** Absorption correctness and plan integrity.

**Plan lines.** The "What 8c may assume" paragraph (`plan:29-32`) against
R1 as applied (`plan:484-490`).

**Evidence.** The header asserts that "the absorbed reader routes through
the seam registry the way `Read` and `Write` do, and `Parse` stays the
one deliberate bypass (m6), so this pass adds no second bypass." Under R1
as applied the shadow calls `tagio.Parse` once and then `tagio.Tags`, so
nothing in `shadow` ever calls `Read` or `Sniff`. `Parse` is documented
as the registry bypass at `internal/engine/tagio/seam.go:39-49` and at
`container.go:130-134`, and the registry's own proofs grade `Read` and
`Write` alone (`TestReadAndWriteRouteThroughTheRegistry` at
`seam_test.go:191`, `TestEmptyRegistryBypassesReadAndWriteButNotParse` at
`seam_test.go:135`). R1's defense at `plan:510-513` is true of `Tags` and
does not reach the path, which opens the file through `Parse`. Only R1's
fold, `Read` for the tag set beside `Parse` for the STREAMINFO, leaves
the absorbed reader a registry consumer in the handoff's sense. The
behavioral consequence is nil for FLAC, because `Parse` validates the
`fLaC` signature itself (`container.go:147-150`), so the defect is
documentary and architectural rather than a runtime hole.

**Consequence, built as written.** The plan's first statement of what it
may assume is false about the thing it builds, and the 8b handoff carries
both halves of the question in one paragraph
(`docs/HISTORY.md:861-864` asks for a registry consumer,
`docs/HISTORY.md:871-875` permits reading through `Parse` and `Container`
directly). A close written from the header would record that 8c added no
bypass and left the registry the one dispatch point for the shadow's
read, which no reader of `shadow` would be able to confirm.

**Fix folded, conductor-owed, riding R1's ruling.** The header paragraph
states both halves of the handoff and says which half each R1 answer
takes. R1 as applied says plainly that the shadow's read path consumes
`Parse` and never the registry, recorded as a position the conductor
rules rather than as an assumption the plan inherited. R1's fold keeps
the registry-consumer reading, and the plan says so at both sites, so a
reversal is a named edit.

### B3. A concurrent lane 2 is asserted as fact in three places, including a gate-cost claim derived from it, while no worktree exists and STATUS says both lanes are idle

**Lens.** Absorption correctness and plan integrity.

**Plan lines.** The agentic-workers note (`plan:65-68`), the gitleaks
constraint (`plan:366-369`), and the lane 2 paragraph (`plan:394-403`),
against Task 0 item 1's own stated expectation (`plan:713-716`).

**Evidence.** The plan states that this pass runs "concurrently with pass
5b in the lane 2 worktree at `.claude/worktrees/lane-2-5b`" and that
"Lane 2 runs concurrently, in a worktree." Measured, `git worktree list`
at `6a9570d` reports one entry, the main tree, and `.claude/worktrees` is
an empty directory. `docs/STATUS.md:24-32` reads "Lane-1 (main tree):
IDLE ... Lane-2 (worktree `lane-2-5a`): IDLE, torn down at 5a's merge ...
No lane active; next session picks one before dispatch." The plan's own
Task 0 item 1 expects exactly that state, so the plan contradicts itself
across forty lines. The dependent claim at `:366-369`, that gitleaks now
scans lane 2's worktree and adds about three minutes, was measured at 8b
with a worktree present (`docs/HISTORY.md:890-894`); with no worktree the
leg is faster.

**Consequence, built as written.** Task 0 item 1 reports a state the
plan's own header denies, which is a blocker report with nowhere to go. A
task told at `:366-369` that a slow gate is lane 2's worktree scan names
a cause that does not exist, so a real regression in the gate would be
absorbed as an expected cost. The lane 2 territory paragraph also tells
Task 0 to report any lane 2 edit under `internal/engine/shadow` or
`internal/engine/tagio`, which is sound as a check and unfounded as a
premise.

**Fix folded.** Lane 2 becomes conditional throughout. The plan says the
pass runs in lane 1 and that a lane 2 worktree may or may not be live,
with Task 0 item 1's registry read settling it before the first dispatch,
and item 1 now reports `git worktree list`'s output and
`.claude/worktrees`'s contents rather than inferring from STATUS alone. The
gitleaks cost moves from a stated fact to a measurement, taken at **Task
1's gate run** rather than at Task 0's, because Task 0 runs no gate by
design (`plan:798-801`); Tasks 2 and 3 then compare against Task 1's
baseline rather than against 8b's figure. The lane 2 territory list
survives as the definition of what would be contested if a lane 2
launches, which is what makes Task 0 item 1's check meaningful either
way.

---

## MAJORS

### M-a. The projection has eighteen rows, not nineteen, and the acceptance criterion asking for a distinguishable value per row is unsatisfiable for two of them

**Lens.** Absorption correctness and plan integrity.

**Plan lines.** A2 (`plan:420`), R2's destination table (`plan:536`),
Task 2's outcomes (`plan:977`), the projection criterion (`plan:1018`),
and mutation 3 (`plan:1053`).

**Evidence.** `internal/engine/shadow/flac.go`'s `fields` sets eighteen
`ItemFields` members, each with its mediafile citation
(`flac.go:208-229`). `matcher.ItemFields` carries nineteen members
(`internal/engine/matcher/info.go:124-152`), and the nineteenth is
`DataSource`, which the reader's own doc comment says has no row because
a grep of pinned mediafile for `data_source` returns nothing
(`flac.go:203-207`, correction 1). `Length` is the second row no
writer-built fixture can distinguish as the plan stands: `tagio.Write`
preserves the incoming STREAMINFO payload byte for byte
(`write.go:151-152`), and R3's seed is
`internal/harness/oracle/testdata/template.flac`, whose 34-byte
STREAMINFO declares a 44100 Hz rate and zero total samples, so every
writer-built fixture reads `Length` 0. Seventeen rows are left that one
writer-built fixture can vary, which is why the criterion at `:1018`
cannot be met as written. Nineteen is the right figure for the itembag
round trip alone, whose bag spells nineteen beets field names including
`data_source` (`itembag.go:14-35`).

**Consequence, built as written.** The criterion asks for a fixture
carrying a distinguishable value in each of nineteen rows. An implementer
either invents a nineteenth row in `fields`, which widens the projection
this pass exists to preserve, or reports the criterion unmeetable after
building the fixture. Mutation 3, which swaps `Label` for `Country`,
fires either way, so nothing in the mutation table catches the arithmetic.

**Fix folded.** Eighteen is the figure everywhere the projection is
meant, and nineteen survives only where the itembag round trip is meant.
The criterion at `:1018` becomes seventeen rows carrying a distinguishable
value in one writer-built fixture, plus `Length` graded by A4's
STREAMINFO table and by m4's non-zero-sample seed variant if Task 1 lands
one, plus `DataSource` asserted as the empty string with the reason it
has no row. A2 and Task 2's A2 outcome carry the eighteen-row count with
the `DataSource` reason beside it, so a reader of either does not go
looking for a missing row.

### M-b. `TestReadFLACStylePrecedence` has sixteen rows, not eighteen, and Task 2 is told to walk eighteen

**Lens.** Absorption correctness and plan integrity.

**Plan lines.** R2's destination table (`plan:537`) and Task 2's outcomes
(`plan:993`).

**Evidence.** The table lives at
`internal/engine/shadow/flac_test.go:114-199`, and its case literals
number sixteen. Counted in order: album artist first style, an empty
first style falling through, TRACK over TRACKNUMBER, a zero first style
not falling through, DISC over DISCNUMBER, disctotal's three styles,
label to PUBLISHER, catalognum to a read-only alias, barcode through the
four read-only aliases, year falling back to YEAR, an unparseable DATE
year yielding zero, keys case-insensitive, the first occurrence of a
repeated key winning, an unparseable int, a slashed pair, and a
non-numeric compilation. `grep -c 'name:'` over that range returns
sixteen.

**Consequence, built as written.** Task 2 is told to walk eighteen rows
against `tagio`'s five projection tests and to add to `tagio` any row not
already covered. An implementer walking sixteen and reporting sixteen
reads as having lost two rows, which invites either a fabricated pair or
a stop-and-report on a plan arithmetic error at the pass's most expensive
task.

**Fix folded.** Both sites read sixteen. The instruction is otherwise
sound and unchanged, so the fold is the number alone.

### M-c. A1 and Task 2's grep criterion forbid container knowledge anywhere in `shadow`, and Task 3 then adds a block-splicing helper to that package with no method named

**Lens.** Absorption correctness and plan integrity.

**Plan lines.** A1 (`plan:412-417`), Task 2's grep criterion
(`plan:1026-1027`), and Task 3's splice helper (`plan:1107`,
`plan:1155-1162`).

**Evidence.** A1 states that after Task 2 the package holds "no FLAC
signature constant, no block-type constant, no metadata block walk, no
`binary.LittleEndian` or `binary.BigEndian` call," and Task 2's criterion
greps `internal/engine/shadow` for `encoding/binary`, `fLaC` and `VORBIS`
expecting nothing outside a comment. Task 3 then creates a helper in
`internal/engine/shadow/fixture_test.go` that splices a second copy of
the writer's own VORBIS_COMMENT block into writer output. Locating that
block and emitting its four-byte header is a block-type constant plus a
header encoder unless the method is named, and `tagio`'s own block-type
constants are unexported (`container.go:22-28`), so the helper cannot
borrow one. The constant-free method exists and is stated in correction 2.

**Consequence, built as written.** Task 3 either violates A1 with a
literal block type and a hand-rolled header, which the grep criterion
does not catch because it looks for names rather than numbers, or stalls
on a contract the plan states two ways. A1 is also the pass's headline
architectural claim, so a Task 3 that quietly re-imports container
knowledge into `shadow` retires the claim without anyone recording it.

**Fix folded.** A1 is scoped to the package's non-test read path, and its
last sentence says that Task 3's fixture helper is the one place in the
package that touches container bytes and that it does so without a block
type or an encoder. Task 3 gains one sentence naming the splice's method:
the helper identifies the VORBIS_COMMENT block by the vendor prefix
`Container.VorbisComment.Vendor` reports, computes its byte range as four
plus the running sum of each preceding block's four plus `Length`, and
copies that range whole with its header, so no block-type constant, no
header encoder and no `encoding/binary` import appears in `shadow`. The
grep criterion is restated to permit the splice helper's own doc comment.

---

## MINORS

**m1. `container.go:41-46` is the wrong range for ruling 9's refusal set,
and it is cited twice.** The enumeration lives in `ErrMalformedContainer`'s
doc comment and the sentinel it documents, `container.go:35-42`. Lines
41-46 are that comment's last line, the `var`, a blank line, and
`ErrUnsupportedContainer`'s own comment. The plan also uses `41-46`
correctly for a different subject, `flac.go:41-46`'s length division,
which is the likely origin of the carry. Folded by changing both
citations, at `plan:41` and `plan:315`, to `container.go:35-42`.

**m2. Two of Task 3's seven expected enumeration rows have no grading
`tagio` test today, and Task 3's file list cannot add one.**
`TestParseRefusals` (`internal/engine/tagio/container_test.go:118`)
carries seven rows: a second VORBIS_COMMENT, a STREAMINFO that is not 34
bytes, a block after a set last-block flag, a declared length past the
end, a field name above U+007E, a second SEEKTABLE, and a bad signature.
The parser refuses three further shapes with no test anywhere in
`internal/engine/tagio/*_test.go`, and a grep of that package's test
files for each message returns nothing: "STREAMINFO is not the first
block" (`container.go:198`, `:209`), "no STREAMINFO block"
(`container.go:245`), and "carries no '='" (`container.go:332`). Task 3
may modify `capture_test.go` and `fixture_test.go` alone, so it cannot
close them, and its criterion at `plan:1166-1168` demands a named
`tagio` test per row. Folded by naming rows 2 and 6 in Task 3 as expected
findings rather than surprises, and by pre-authorizing Task 2 to add the
missing refusal rows to `container_test.go`, which its file list already
permits under R2's walk clause. Adding a test for a landed refusal adds
no refusal, so the global constraint at `plan:314-319` is untouched.

**m3. R4's denominator is not a field the shadow report prints.** R4 item
2 asks the conductor to read "the number of captured albums the shadow
holds" through `dubplate-server shadow report`. `shadow.Report`
(`internal/engine/shadow/report.go:17-36`) carries eighteen fields, and
none of them is a captured-album total: `ClockStart`, `ClockStarted`,
`Floor`, `Elapsed`, `ObservedApplied`, `ObservedSkipped`, `Excluded`,
`Unreconciled`, `ReportGaps`, `Deferred`, `Divergences`,
`GoAppliesOracleSkipped`, `StagedPending`, `Resolved`, `MatcherCommits`,
`Passes`, `BudgetTotal` and `StoreBytes`. The denominator has to be
composed from several rows or accepted as a proxy. Folded provisionally
as conductor-owed, because the conductor owns the reading: R4 item 2 says
the denominator is a proxy composed from `ObservedApplied`,
`ObservedSkipped` and the `Unreconciled` counts, and the conductor names
in the close which rows it actually summed and what the sum does and does
not count. A conductor who reads the report and finds a better composition
records that instead, which is why the fold names the rows rather than
fixing the arithmetic.

**m4. The end-to-end length join goes vacuous once every fixture comes
from the zero-sample seed, and Task 4's gap list does not carry it.**
`reconcile.go:242` joins on title and length,
`math.Abs(item.Length-fields.Length) < lengthTolerance`. Today's external
`buildFLAC` fixtures declare 44100 samples at 44100 Hz
(`capture_test.go:34`, `:173`, `run_test.go:107`), so both sides carry a
non-zero length. R3's seed declares zero total samples and `tagio.Write`
preserves STREAMINFO (`write.go:151-152`), so every writer-built fixture
reads `Length` 0 and the join's length term compares zero to zero on both
the capture side and the library side. `reconcile_test.go` grades the join
with fabricated `ItemFields` and is unaffected, so no test fails and
nothing records the loss. The plan asks Task 1 to "say so explicitly"
(`plan:896-899`) and never routes it anywhere a later reader would look.
Folded two ways, and Task 1 reports which route it took: the vacuous
length term is added to Task 4's expected declared-gap list, and Task 1 is
permitted to carry a second seed variant whose STREAMINFO pack declares a
non-zero total-sample count, taken as a byte edit of the committed literal
rather than as a container builder, which closes the join instead and also
gives M-a's `Length` row a distinguishable value.

**m5. Task 3 cites `capture.go:130` for the symbol it consumes, and the
partial-read rule it means belongs to `captureAlbum`.**
`internal/engine/shadow/capture.go:78` declares `capture`; line 130 is the
last line of `captureAlbum`'s doc comment and `captureAlbum` is declared
at `:131`. The partial returns the plan means are at `:133-135`,
`:145-147` and `:159-161`, all inside `captureAlbum`. Folded by citing
`captureAlbum` (`capture.go:131`) for the partial-read rule at both
`plan:1111` and `plan:1120`, and by naming `capture` (`capture.go:78`)
separately where the walk itself is meant.

**m6. Task 4 calls `itembag.go` a fifth site, and the spec's own paragraph
already names six references to the reader.** `spec:978-985` reads "Three
call sites reference the reader: `shadow/run.go:209`,
`shadow/capture.go:158`, and the doc comment on `shadow/reconcile.go:35`.
Two more mention it in prose (`capture_test.go:47`, `:203`). A sixth is
the reader's own doc comment at `flac.go:62-70`." Both prose mentions
still resolve after this pass, because R7 keeps `readFLAC`'s name, so no
edit is owed at either. Folded by saying that the spec's paragraph names
six references, that the two prose mentions survive unedited under R7,
and that `itembag.go:11-13` is the one reference the spec never named.

**m7. Six citation ranges overshoot their subject.** Five were raised by
the review and the sixth was found while folding. All were re-resolved
against the tree by the fold editor. Ruling 14 runs `spec:338-342`, and
`spec:343` is ruling 15, the product fork; the plan cites `spec:338-343`
at `:83` and `:321`. The R12 erratum is `spec:766-774`, and
`spec:752-756` is the sentence the erratum corrects, so the plan's
"`spec:752-756`'s erratum" at `:498` names the wrong end of the pair
(B1). The committed template literal is `flac_test.go:244-251`, and
`:252-254` are a blank line and the next test's comment; the plan cites
`244-254` at `:753` and `:848`. 8b's forward carries run
`docs/HISTORY.md:880-898`, with `:899` blank and `:900` the next heading;
the plan cites `:880-899` at `:89`. 8b's coverage gap 4 runs
`docs/HISTORY.md:830-833`, with `:834` the opening of gap 5; the plan
cites `:830-834` at `:502`. The sixth, found while folding, is R3's
seed-literal citation: the provenance comment plus the literal run
`flac_test.go:237-251`, and the plan cites `:237-254`. Content matches in
every case and only the ranges moved, which is the drift Task 0 item 14
exists to catch. Folded by correcting all six. Four more ranges move under
m1, m5, B1 and B2, so the plan's whole correction is ten ranges across
fourteen sites: `container.go:41-46` to `:35-42` twice (m1),
`capture.go:130` to `capture.go:78` and `:131` (m5), `spec:752-756` to
`spec:766-774` (B1), `docs/HISTORY.md:871-876` to `:871-875` twice and
`:860-864` to `:861-864` (B1 and B2), and m7's own six.

---

## Rejected findings

None retired. Every finding in the review survived re-verification against
the tree. Three proposed folds were rewritten by the fold editor rather
than taken as offered. M-c's splice method changed, because the block-type
constants it would have borrowed are unexported (correction 2). m4 is
folded as a gap entry plus a permitted seed variant rather than as one or
the other, so Task 1 reports which route it took, and M-a's `Length` row
is written to hold under either route. B3's gate measurement moved from
Task 0 to Task 1, because Task 0 runs no gate by design and adding one
would change its deliverable rather than correct a fact.

---

## Verified sound

Do not re-litigate these. Each was checked by the fold editor against the
named artifact, and every mechanism claim was measured on go1.27.1.

1. The `buildFLAC` call-site count is right. Counted independently, three
   external uses at `capture_test.go:34`, `capture_test.go:173` and
   `run_test.go:107`, and four inside its own file at `flac_test.go:85`,
   `:202`, `:226` and `:288`. `blockHeader` (`:51`) and `writeTemp`
   (`:63`) have no user outside `flac_test.go` either, so Task 2's
   deletion of all four symbols with the file is self-contained.
2. No symbol from `flac.go` is used anywhere else in the package. A grep
   for `flacSignature`, `streamInfo`, `flacFile`, `parseFLAC`, `castInt`,
   `castBool`, `dateYear`, `firstBytes` and `vorbisField` outside
   `flac.go` and `flac_test.go` returns nothing, so deleting the file
   breaks nothing Task 2 does not already list.
3. Task 1's edit list is complete by files. Only `capture_test.go`,
   `observe_test.go` and `run_test.go` call `writeAlbum`, `albumTracks`,
   `writeLibraryFile` or `windroseTracks`, and the inline `[][]string`
   literals are exactly the five the plan names, at `capture_test.go:61`,
   `:92`, `:119`, `:196` and `observe_test.go:111`. Changing
   `albumTracks` and `writeAlbum` to carry a `tagio.TagSet` leaves
   `observe_test.go`'s other call sites compiling unchanged, which is why
   one line of that file is the whole edit.
4. The independence marks are true by files. Task 1 touches the three
   test files plus the new `fixture_test.go`, Task 2 deletes
   `flac_test.go`, which still holds `buildFLAC`'s remaining callers, and
   Task 3 edits `capture_test.go` and `fixture_test.go` against Task 2's
   absorbed path. No pair is independent, exactly as the plan says.
5. A9 holds as measured. `go test ./internal/engine/tagio/... -run
   Conformance -count=1` selects `TestConformanceLedgerSubjects`
   (`conform_test.go:640`) and reports twelve subtests in 0.052s on
   go1.27.1, matching `docs/HISTORY.md:954-957`. The instrument is named
   as a must-stay-green criterion inside Task 2, the task that could break
   it, rather than only at the close.
6. The exclusion count is sourced read-only and never as an implementer
   box write. R4 item 2 assigns the denominator to the conductor under the
   quiet-box check, R4 item 3 explains why the count is not derivable from
   stored records, and the sentence that no implementer task ssh's appears
   at `plan:336`, `:594` and `:1176`. Task 3's own deliverables are a test
   and a table, with the box reading held outside the chain.
7. The lane and implementer names are right. The plan runs in lane 1, the
   main tree, which is the spec's tree column for 8c (`spec:1168`).
   `.claude/agents/dubplate-implementer.md` exists and
   `~/.claude/workflows/pass-execute.js` exists.
8. Task 0 item 4's expected `tagio` surface is complete and accurate
   against the tree. `go doc -all` reports `Parse`, `Read`, `Write`,
   `TagSet`, `Mapping`, `WriteOrder`, `Serialize`, `EmptyString`,
   `EmptyList`, `Container`, `Block`, `StreamInfo`, `Comment`,
   `VorbisComment`, `KeyWrite`, `Field`, `FieldMapping`, `Style`,
   `Picture`, `FrontCover`, `Format`, `SupportedFormats`, `Sniff`,
   `ErrUnsupportedContainer`, `ErrMalformedContainer` and `Deviations`,
   and nothing exported is missing from the plan's list.
9. Task 0's other expectations check out against the tree at `6a9570d`.
   `flac.go` is 313 lines last touched by `a86475c`, `tagio` was last
   touched by `ce0ac2e`, `.gitleaksignore` is tracked and zero bytes,
   `internal/harness/oracle/testdata/template.flac` is 42 bytes,
   `corpus/analysis/file_tags.json` carries `shim_version` 1,
   `beets_version` 2.13.1, 69 writable fields, 9 read-only fields and
   twelve rows, the arch tests sit at
   `internal/harness/conform/arch_test.go:141` and `:152` and root at
   `internal/engine`, and `internal/engine/tag/doc.go:8-11` still says
   `tagio` knows nothing about matching. The `shadow` package's own suite
   reports 75 top-level tests in 0.241s, which is Task 0 item 6's baseline.
10. Every call-site pointer in Tasks 1 through 3 resolves: `run.go:209`,
    `capture.go:158`, `reconcile.go:35`, `itembag.go:11-13`'s
    `flac_test.go` pointer, `capture_test.go`'s `writeAlbum` at `:25` and
    `windroseTracks` at `:60`, `run_test.go`'s `albumTracks` at `:85` and
    `writeLibraryFile` at `:103`, `observe_test.go:111`, and
    `capture_test.go:194-215` for the partial-read discard. Every
    `flac.go` pointer resolves too: `:13`, `:41-46`, `:62-70`, `:72-76`,
    `:77`, `:117-121`, `:170-177` and `:263-307`.
11. Every `tagio` test named in R2 and in the mutations resolves to the
    test it claims: `TestParseRefusals` at `container_test.go:118`,
    `TestParseGolden` at `:359`, `TestWriteGoldenBytes` at
    `write_test.go:144`, `TestProjectStylePrecedence` at
    `read_test.go:144`, `TestProjectLastStyleRule` at `:199`,
    `TestProjectDates` at `:222`, `TestProjectNumericCasts` at `:249`,
    `TestProjectComparesNamesCaseInsensitively` at `:300`, and
    `read_test.go:277`, which is precisely the row asserting that a
    single-valued field over a repeated key reads the first value.
12. The spec pointers all land on their subjects: the 8c row at
    `spec:1168`, ruling 3 at `268-272`, ruling 4 at `273-279`, ruling 9 at
    `310-317`, ruling 10 at `318-326`, ruling 13 at `337`, the absorption
    paragraph at `978-985`, the five-tests bullet at `995-998`, the
    itembag bullet at `999-1001`, the layout-test bullet at `1002-1010`,
    the strictness section at `1012-1033`, the closing projection
    paragraph at `1034-1038`, the hooks table at `1065-1080`, the
    mechanical-except-the-tightening sentence at `1173-1175`, the
    contested-surfaces table at `1186-1204` with the shadow row at
    `1197`, the ceiling proposal at `1214`, the re-derivation clause at
    `1226-1228`, the staleness checklist at `1230-1277`, risk 2b at
    `1294-1301`, risk 6 at `1320-1330`, the conventions sentence at
    `1351-1353` and the review-scaling paragraph at `1356-1364`. The
    parallel-approach ownership table is at `:175-207` with the
    `internal/engine/shadow` row at `:199`.
13. The HISTORY pointers land: the 8b to 8c handoff at `:856-879`, the
    registry sentence at `:861-864`, the conformance sentence at
    `:864-869`, the `io.ReaderAt` sentence at `:869-871`, R12's settled
    answer at `:871-875`, the restatement precedent at `:812-819`, the
    gitleaks carry at `:890-894`, the exported-surface record at
    `:662-672` and the spend at `:1004-1012`. Only the two ranges m7
    corrects overshoot.
14. The ceiling arithmetic is internally consistent and its basis is
    real. 8b closed at about 2.9M against 3.25M with prep 0.87M, Task 0
    0.11M, chain 1.51M across 16 agents, leftovers 0.18M, simplifier
    0.13M and conductor 0.1M (`docs/HISTORY.md:1006-1011`). Dividing the
    chain by seven is right, because 8b carried Task 0 plus Tasks 1
    through 7 with Task 7 as the close, the shape 8c's Task 4 takes.
    Eighty per cent of 2.1M is 1.68M, the three ceilings sum to 9.85M,
    and the measured underspend of 1.25M is right. Only the expected rung
    total reads 8.38M where 6.5 plus 1.87 is 8.37, which the plan hedges
    with "about".
15. The 8b plan review's merged verified-sound list begins at `:683` and
    holds exactly forty-two numbered items, as the plan's primary input 4
    claims.
16. Nothing outside the spec's 8c row entered the plan except R1's one
    exported symbol, which the plan declares as deviation 4 and routes to
    the conductor before Task 2. The exclusions each carry a stop
    condition: a new `tagio` refusal, a new corpus subject, a `make
    derive` run, a box write, a module dependency, a configuration key
    and a golden regeneration all stop the task and report.
17. All three exit-gate assertions from `spec:1168` appear as acceptance
    criteria. `check.sh` green is the last criterion in Tasks 1, 2 and 3.
    The `shadow` suite green with the projection unchanged is Task 2's
    first criterion, graded against Task 0 item 6's baseline with exactly
    one expected value permitted to change. The exclusion count is Task
    3's enumeration criterion plus R4's three-part answer reported by
    Task 4.
18. A2's claim that `ItemFields` carries no presence bit, so an absent
    `*string` and a present empty string both project to the empty
    string, survives the change of reader. `tagio`'s last-style rule
    leaves `catalognum` and `barcode` absent where the landed reader left
    them empty, and both project to the empty string through the shadow's
    typed-default rule, so the collapse absorbs the difference rather
    than hiding a second value change beside R6's.
19. The fixture-producer switch cannot break a pinned digest. The
    capture digest is computed over uploader, basename and each file's
    name, size and modification time (`capture.go:191-205`), no shadow
    test pins a real digest literal, and the writer's vendor string is a
    constant prefix plus the main module's build version
    (`write.go:233-244`), so writer output is byte-stable within a run.
20. Task 3's enumeration rows 1 through 7 are each genuinely a shape the
    landed reader accepted and `tagio` refuses. Verified in the parser:
    more than one VORBIS_COMMENT at `container.go:217-220`, STREAMINFO
    not first at `:198` and `:209`, no STREAMINFO at `:245`, a block after
    a set last-block flag through the last-flag check, a second SEEKTABLE
    at `:229`, a field name outside U+0020 through U+007E, a comment
    carrying no `=` at `:332`, and the PICTURE decode's own refusals. The
    landed reader tolerates every one of them, switching on block type
    with no ordering or uniqueness requirement (`flac.go:110-121`) and
    skipping unknown blocks by declared length.
21. The plan carries no em dash. A grep for the character over the whole
    file returns zero.
