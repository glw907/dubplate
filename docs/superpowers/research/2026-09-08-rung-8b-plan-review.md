# Adversarial review: the rung 8 pass 8b plan

Fresh-context review (2026-09-08) of
`plans/2026-09-08-rung-8b-picture-conformance.md` revision 1, in two
independent lenses: FLAC PICTURE layout and library fidelity, verified
block by block against the committed golden's bytes, against pinned
mediafile 0.17.0, mutagen 1.48.1 and filetype 1.2.0 in
`tools/venv/.venv`, and against RFC 9639 as published; and plan
integrity, verified against the rung 8 spec revision 2 with its
2026-09-08 erratum, the 8a-to-8b handoff and forward carries in
`docs/HISTORY.md`, the corpus handoff, the reverse-edge ownership rows,
the 2.0 streamer spec's forward-hooks list, and the tree at `0f2b61d`.
Folded as revision 2 the same day. The merged 42-item "Verified sound"
list below is the do-not-re-litigate record for the pass, alongside the
rung 8 spec's own records, 8a's and rung 4a's.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1, M-a through M-g, m1 through m12. Every finding was
re-verified by the fold editor against the tree, the venv or the RFC
before folding, and every numeric claim was re-measured rather than
carried. Six reviewer measurements did not survive that check in the
form they were written and are corrected below; none of the six retires
its finding. One finding was rejected outright and one half of another
was rejected, both with the measurement that retired them. Six items are
conductor-owed rather than folded silently, and each is applied
provisionally in the plan with an inline marker.

Both lenses agree the plan is well grounded where it reasons from
measurement. Every mediafile and mutagen behavior rule resolves to the
pinned line it names, P4 through P8 reproduce the reference exactly
including the non-obvious single-candidate-first ordering, the
four-state model holds against all twelve subjects on both the case side
and the ledger side, R11's positional tie is a true invariant of the
landed tree, the ceiling arithmetic is internally consistent and
re-derived from 8a's actual as the spec's clause requires, the
independence marks are true by files, and the comparison design cannot
be defeated by a mirrored bug because the ledger is external evidence.
The prose measures clean: tellgrader reports 0 tells over 13,931 words
at a cadence CV of 0.75, and the file carries no em dash.

The defects concentrate in four places. The one assertion in the pass
that anchors to bytes another implementation produced never says where
its input comes from, so it can be satisfied as a decode-and-encode
round trip. The block's own field count is stated three different ways
and none of the three matches RFC 9639. The magic-byte rules the plan
claims are cited row by row from the pinned `filetype` package are not
what that package tests. And the writer's ownership of the PICTURE block
type is never stated, which leaves both the preserve-unknown transition
and the new block's position in the chain unpinned.

---

## Corrections to the reviews' own evidence

**1. Review 1's M-c says `Cr2.match` "is tried first in the IMAGE
tuple".** It is not. `filetype/types/__init__.py:15-32` lists `Tiff()`
at index 8 and `Cr2()` at index 9, so TIFF is tried first. The CR2
answer wins because `Tiff.match` carries its own exclusion,
`not (buf[8] == 0x43 and buf[9] == 0x52)` at
`filetype/types/image.py:198`, which hands the blob down to `Cr2`. The
finding stands with the corrected mechanism, and the mechanism matters
to the fold: a Go sniff reproducing the reference needs the exclusion in
the TIFF row rather than an ordering rule.

**2. Review 1's M-d cites `container.go:344` for the VORBIS_COMMENT
length check.** Line 344 reads the size; the refusal is at `:345`, and
the check spans `:344-346`. The finding stands; only the pointer moves.

**3. Review 2's m1 puts Task 0 item 11 at `plan:854-857`.** Item 11 is
at `:858-862`; `:854-857` is item 10. Review 2's m3 puts item 6 at
`:828-835`; item 6 opens at `:826`. Both findings stand, and the two
pointers swap.

**4. Review 1's B1 cites `golden.txt:71-72` for the picture's data
literal.** Line 71 is the lead-in sentence and `:72` carries the
literal. The write input's full provenance is `golden.txt:71-74`, which
records the data, the description and the type index as the generator
passed them, and `gen_golden_flac.py:115` and `:131`, which are the
generator's own lines. The fold cites all four rather than the decoded
record alone, because the decoded record at `:203-210` is the output
side of the same fact and citing only it is what leaves the round-trip
reading open.

**5. Review 2's m10 measures "rather than" at 59 occurrences and calls
it a tic.** The count is right and the rate is not a defect. Measured
per thousand words, the plan runs 4.2, against 4.2 in the ratified
exemplar `plans/2026-09-05-rung-4a-shadow-runner.md` (173 over 41,108
words) and 3.9 in `plans/2026-09-07-rung-8a-tagio-writer.md` (74 over
18,823). The plan sits at the house rate. That half is rejected below;
the bolded-lead half survives with its own measurement.

**6. Review 2's M-b says a non-matching `-run` pattern "prints no tests
to run (or nothing) and exits 0".** Measured rather than assumed:
`go test ./internal/engine/tagio/... -run ZzzNoSuchPattern -count=1`
prints `ok ... [no tests to run]` and exits 0. The finding stands on the
measurement.

---

## BLOCKERS

### B1. Task 2's golden PICTURE byte assertion never says where its input picture comes from, so it can be satisfied as a decode-and-encode round trip

**Lens.** FLAC and library fidelity.

**Plan lines.** Task 2's first acceptance criterion (`plan:1067-1072`),
against Task 1's parallel criterion (`plan:946-951`).

**Evidence.** Task 2's criterion reads "The emitted PICTURE block's
header and payload equal the golden's, byte for byte, for a write of the
golden's own recorded picture into a copy of the golden." Task 1's
criterion is explicit where this one is silent: "The expectation is
transcribed from the provenance record, never computed from the
decoder." Nothing in Task 2 forbids the implementer from building the
input picture by calling Task 1's decode over `golden.flac`.

Under that reading the assertion becomes decode-then-encode against the
same bytes, and any symmetric error survives it. Both sides
little-endian, both sides reading at the same wrong offset, both sides
carrying the same off-by-one description length: each reproduces the
golden's bytes exactly. Task 2's own mutation 6 (`plan:1119-1122`,
"Emit the picture block's `u(32)` fields little-endian → the golden byte
assertion fails") is true only under the hand-transcribed reading, and
no named mutation in the pass covers the both-halves case.

**Consequence, built as written.** This is the pass's only anchor to
bytes another implementation produced (`plan:1070-1072`,
`spec:1282-1285` risk 5), and it is the whole of the plan's answer to
the risk that PICTURE ships with no corpus evidence. A round-trip
reading leaves the pass with no external anchor for the picture encoder
at all.

**The material for the correct route exists and the plan never cites
it.** `gen_golden_flac.py:115` holds the data literal,
`b"\x89PNG\r\n\x1a\n" + b"invented-picture-payload"`, 32 bytes;
`:131` holds the `mf.images` call with `desc="cover"` and
`type=ImageType.front`; `golden.txt:71-74` reproduces all three as the
write's recorded input. The fold editor confirmed the block against the
file's actual bytes: PICTURE at offset 1112, declared length 78,
decoding to type 3, `image/png`, description `cover`, geometry all zero,
32 data bytes. The arithmetic closes: 4 + 4 + 9 + 4 + 5 + 16 + 4 + 32 =
78.

**Folded** by adding one sentence to Task 2's criterion: the picture
handed to `Write` is constructed in the test from values hand-transcribed
from `golden.txt:71-74` and `gen_golden_flac.py:115` and `:131`, never
from `Parse` or Task 1's decoder, and the `diff-reviewer` checks that no
`want` and no write input is produced by calling into the package.
`gen_golden_flac.py` joins Tasks 1 and 2's input lists, since it is the
provenance record and revision 1 never named it.

---

## MAJORS

### M-a. The PICTURE block's field count is stated three ways, and none of the three is RFC 9639's

**Lens.** Both, independently. Review 2 carried it at major on the
length-count contradiction; review 1 carried the field-count half at
minor. Merged at major, because Task 1's implementer is told to
bounds-check a field that does not exist.

**Plan lines.** `plan:122`, P1 (`plan:412-417`), P3 (`plan:427-435`),
Task 1's outcome (`plan:936-939`) and Task 1's refusal criterion
(`plan:954-958`).

**Evidence.** P1 promises "the seven fields RFC 9639 section 8.8 defines
in the order it defines them" and then lists eight: the picture type
index, the media type string, the description, the width, the height,
the colour depth, the indexed-colour count, and the data. P3 says "The
block carries four length fields, each a big-endian `u(32)`", and Task 1
repeats it: "every one of the four declared lengths is bounds-checked".

`INV:770-776`, the authority the plan cites, records Table 12 in order:
picture type `u(32)`, media type length `u(32)`, media type string,
description length `u(32)`, description, width `u(32)`, height `u(32)`,
colour depth `u(32)`, number of colours `u(32)`, picture data length
`u(32)`, and the picture data. That is eight `u(32)` fields and exactly
three declared lengths. Mutagen writes the same eight,
`struct.pack('>2I', ...)` then `'>I'` then `'>5I'`
(`mutagen/flac.py:612-623`). Task 1's own refusal criterion already
enumerates three and no more: "a media-type length, a description
length, and a data length each declaring more bytes than the block
holds". The block header's own length is `u(24)`
(`write.go:208-210`, `maxBlockPayload = 1 << 24`), so it is not the
missing fourth.

**A second error rides in the same criterion.** Task 1 asks for "a block
shorter than the fixed 32-byte header the seven fields need before any
variable-length data". Thirty-two bytes is right as a minimum block
length, and it is the sum of the eight `u(32)` fields. It is not a
header: the media type string sits between the media-type length and the
description length, so the eight fixed fields are not contiguous, and no
implementation can read them as a header struct.

**Consequence, built as written.** An implementer told to bounds-check
four lengths either invents a fourth or writes three and contradicts the
plan on review. An implementer told to read a 32-byte header before any
variable-length data writes a decoder that is wrong about the layout in
the one way the golden's own arithmetic would catch.

**Folded** by correcting P1 to eight fields with the three lengths named,
correcting P3 and Task 1's outcome to three declared lengths, scoping
the big-endian claim to the eight `u(32)` fields rather than to all of
them, and restating Task 1's minimum-length row as "a block shorter than
32 bytes, the sum of the eight `u(32)` fields every picture block
carries", with the non-contiguity stated. Task 1's refusal table stands
unchanged: its three length rows plus the short-block row already match
RFC 9639.

### M-b. The rung 15 forward criterion pins a `-run Conformance` command that no task's test name has to match, and a non-matching pattern exits 0

**Lens.** FLAC and conformance-instrument fidelity.

**Plan lines.** The rung 15 forward criterion (`plan:194-198`), Task 4's
file creation (`plan:1263`), Task 5's file modification (`plan:1379`)
and its wall-time criterion (`plan:1431-1433`).

**Evidence.** The forward criterion states that rung 15's re-run is
`go test ./internal/engine/tagio/... -run Conformance -count=1`, and
Task 5 reports wall time for the same command. No task states that the
conformance test function's name must match that pattern. Task 4 creates
`conform_test.go` and Task 5 extends it, both without a naming
requirement.

Measured on the tree: `go test ./internal/engine/tagio/... -run
ZzzNoSuchPattern -count=1` prints `ok ... [no tests to run]` and exits
0. A conformance test named, say, `TestLedgerSubjects` would leave rung
15's launch-gate re-run passing while grading nothing.

**Consequence, built as written.** That is the same vacuous-green class
the pass's own exit gate exists to close (`spec:1134`), reached through
the launch gate rather than through the instrument.

**Folded** by stating the test function name as an outcome in Task 4,
`TestConformanceLedgerSubjects`, requiring Task 5 to report the test
count the `-run Conformance` invocation actually executed alongside its
wall time, and carrying the name beside the command in Task 7's handoff
and in the rung 15 forward criterion.

### M-c. R2's WEBP and TIFF rows do not reproduce pinned filetype 1.2.0, so the "cited row by row" claim is false and the CR2 collision goes unrecorded

**Lens.** FLAC and library fidelity.

**Plan lines.** R2 (`plan:546-551`), Task 2's sniff-table criterion
(`plan:1074-1078`), and Task 2's deviation outcome (`plan:1061-1063`).

**Evidence.** R2 states that the recognized set is taken "each by the
magic-byte test the pinned `filetype` package applies, cited row by
row", and gives WEBP as "`RIFF` at 0 with `WEBP` at 8" and TIFF as
"`49 49 2A 00` or `4D 4D 00 2A`". Read at the pinned version in
`tools/venv/.venv/lib/python3.13/site-packages/filetype/types/image.py`:

`Webp.match` (`:144-155`) requires `len(buf) > 13` and, beyond RIFF at 0
and WEBP at 8, `buf[12] == 0x56 and buf[13] == 0x50`, which is `VP` at
offset 12. A RIFF container that is not a WebP bitstream is refused.

`Tiff.match` (`:193-199`) requires `len(buf) > 9` and, beyond the two
four-byte prefixes, `not (buf[8] == 0x43 and buf[9] == 0x52)`.
`Cr2.match` (`:171-177`) claims the same two prefixes plus `CR` at
offset 8. The exclusion is what routes a CR2 blob to
`image/x-canon-cr2`; tuple order does not, since `Tiff()` precedes
`Cr2()` in `IMAGE` (`types/__init__.py:15-32`).

**Consequence, built as written.** A Go sniff built to R2's stated rules
returns `image/webp` for a RIFF blob the reference refuses, and
`image/tiff` for a CR2 blob the reference calls `image/x-canon-cr2`.
Under R2's own no-fallback argument the reference then hands mutagen a
`None` mime for the RIFF case and raises, so the divergence is a written
file where the reference wrote none. Task 2's sniff-table criterion
builds each row from the plan's own prefixes, so neither divergence
fails a test, and Task 2's deviation list records only P6 and R3.

**Folded** by correcting R2's WEBP row to RIFF at 0 plus WEBP at 8 plus
`VP` at 12 with a fourteen-byte minimum, correcting its TIFF row to the
two four-byte prefixes with the CR2 exclusion at offset 8 and a ten-byte
minimum, and giving Task 2's sniff table a row for a RIFF blob without
`VP` and a row for a CR2-prefixed blob, each asserting refusal.
**Conductor-owed:** the CR2 answer is applied provisionally as a
refusal plus a `deviations.go` row recording that the reference writes
`image/x-canon-cr2` where `tagio` refuses. Adding a seventh sniff row
for a raw camera format no cover realistically carries is the fold if it
is ruled the other way.

### M-d. P3's bounds-check refusal contradicts mutagen's `_distrust_size` behavior, and the plan neither rules it nor records it

**Lens.** FLAC and library fidelity.

**Plan lines.** P3 (`plan:427-435`), the plan's own header rule
(`plan:25-26`), and Task 2's deviation outcome (`plan:1061-1063`).

**Evidence.** P3 makes any declared length running past "the block's own
declared payload" an `ErrMalformedContainer`. The reference does the
opposite. `Picture._distrust_size = True` (`mutagen/flac.py:575`), and
`FLAC.__read_metadata_block` therefore ignores the 24-bit block size and
parses the picture straight from the file object using the block's
internal lengths (`:712-724`). The comment there cites quodlibet/mutagen
issue 106 for exactly this real-world shape: "Some jackass is writing
broken Metadata block length ... and the FLAC reference implementation
can parse them (mostly by accident), so we have to too."

The same policy already applies one block type over. `VCFLACDict`
carries `_distrust_size = True` (`mutagen/flac.py:350`), and 8a's
`vorbisField` refuses on the block's declared extent
(`container.go:344-346`, "field declares %d bytes, %d remain"), which is
the same divergence landed and unrecorded in `deviations.go`.

**Consequence, built as written.** A library file with a wrong PICTURE
or VORBIS_COMMENT block length is one beets can tag and `tagio` refuses
outright, and the refusal reaches rung 7's intake gate as a rejection of
a file the user's other players read. The plan's header states "8b adds
no new container refusal without a fresh ruling", and P3 asserts this is
"not a new refusal class", which is true of the sentinel and not of the
condition.

**The hardening itself is right.** Trusting caller-controlled internal
lengths past the block extent is the `dhowden/tag` #116 shape the pass
exists to avoid. It is the silence that is not right.

**Folded** provisionally as a `deviations.go` row in Task 2 stating that
`tagio` trusts the block header length where mutagen distrusts it, for
both VORBIS_COMMENT and PICTURE, naming `mutagen/flac.py:350`, `:575`
and `:712-724` and the intake consequence, and joined to Task 7's
deviation summary and coverage-gap list. **Conductor-owed:** whether the
condition needs the fresh ruling the plan's own header requires, or
whether the stated deviation discharges it.

### M-e. Per-task `code-simplifier` dispatches are unbudgeted, and they are scheduled after each task's commit

**Lens.** Plan integrity.

**Plan lines.** The Execution model (`plan:1682-1685`), the ceiling
table's simplifier line (`plan:224`), and the "Gate green; commit"
closing of Tasks 1 through 6.

**Evidence.** The Execution model reads "The conductor dispatches
`code-simplifier` over each accepted task's changed Go after review and
before the pass close", which is seven dispatches. The ceiling table's
only simplifier line is scoped elsewhere: "The `code-simplifier` round
over 8a's Go, plus its review | Forward carry 1; flat at about 0.2M
across seven passes | 0.2M". The measured per-task rate the plan reuses
(2.27M over nine tasks, about 252K) cannot contain a simplifier cost
either, because forward carry 1 records that 8a's round never ran
(`docs/HISTORY.md:366-367`). Expected total is 3.06M against a 3.25M
ceiling, so the unbudgeted headroom for seven rounds is 0.19M.

The scheduling half is structural rather than arithmetic. Every task
ends "Gate green; commit", so a dispatch "after review" lands after that
task's commit, against CLAUDE.md's rule that `code-simplifier` runs over
changed code before committing. The two cannot both be honored inside
the chain: the implementer commits per task, and CLAUDE.md's build
process forbids instructing an implementer to dispatch an agent, so the
conductor's dispatch cannot reach a point before a chain task's commit.
That is why 8a's practice was one batched round owed at the next pass's
opening.

**Folded** by scoping the Execution model to one batched
`code-simplifier` round over 8b's Go, dispatched by the conductor after
Task 6 is accepted and before Task 7's close and committed on its own,
with the departure from CLAUDE.md's before-committing rule stated in the
same place and its reason given. The ceiling's simplifier line is
re-scoped from 8a's round to 8b's, which also disposes of m2, and the
expected total and the 80% flag are unchanged at 3.06M and 2.6M.
**Conductor-owed:** the batched reading rather than a per-task one, and
the stated departure.

### M-f. Nine prior-art citations point past their rows in the spec's table, including the dhowden defect P3 rests on

**Lens.** Both, independently, with identical row arithmetic. Merged at
review 2's severity.

**Plan lines.** `plan:433`, `:468`, `:1010`, `:1136`, `:1235` (all
`spec:1064` for `dhowden/tag`); `plan:1239` (`spec:1065-1066` for
`bogem/id3v2` and `abema/go-mp4`); `plan:1139` (`spec:1067` for
`gcottom/audiometa`).

**Evidence.** The spec's prior-art table rows were resolved
mechanically by printing the first cell of every line from 1055 to 1068:
1060 `go-flac/flacvorbis` and `flacpicture`, 1061 `dhowden/tag`, 1062
`bogem/id3v2`, 1063 `abema/go-mp4`, 1064 `Eyevinn/mp4ff`, 1065
`gcottom/audiometa`, 1066 Picard, 1067 taglib. Six plan citations send a
reader looking for `dhowden/tag` to `Eyevinn/mp4ff`; `spec:1065-1066`
sends a reader looking for `bogem` and `abema` to `gcottom` and Picard;
`spec:1067` sends a reader looking for `gcottom` to taglib. Every
claim's substance checks out against the table, so only the pointers are
wrong. The `go-flac/flacpicture` citation at `spec:1060` is correct, so
the offset is not uniform and cannot be corrected by a single shift.

**Consequence, built as written.** P3 rests on `dhowden/tag` #116 as its
whole justification, and an implementer or reviewer who follows the
pointer reads an MP4 streaming library instead. Task 0's staleness
framing catches only "a citation that has moved" (`plan:779-781`); these
were wrong at `bc4016b`, so no Task 0 item re-resolves them.

**Folded** by repointing the six dhowden citations to `spec:1061`,
`spec:1065-1066` to `spec:1062-1063`, and `spec:1067` to `spec:1065`,
and by adding Task 0 item 14: resolve each distinct `spec:`, `INV:`,
`RE:` and `docs/HISTORY.md:` pointer this plan carries by grepping for
the subject rather than by reading the line number, and report every one
that does not land on its subject. The plan names stale pointers as this
build's dominant defect class, so the check belongs in the pre-task
rather than in its preamble.

### M-g. The plan leaves unresolved whether the exported read entry point returns pictures

**Lens.** Plan integrity.

**Plan lines.** P2 (`plan:419-425`), Task 1's outcomes
(`plan:926-943`), and Task 2's widening of `Write` (`plan:1043-1046`).

**Evidence.** The spec's caller-surface paragraph is explicit: "The two
entry points a caller uses. One reads a file's tag set and its pictures
... One writes a tag set and its pictures over a file" (`spec:729-733`).
The plan attaches pictures only to `Container`: P2 says "`Container`
carries the pictures it parsed", and `Read` returns `TagSet` alone
(`read.go:14`) with no plan item widening it. Task 2 widens `Write` and
no task widens `Read`. No resolution R1 through R11 takes the question.

**Consequence, built as written.** A caller at rung 9 that wants the
tags and the cover parses the file twice, through `Read` and again
through `Parse`. That is the surface hook 3 depends on: the 2.0
streamer's art blob requires that "embedded pictures are extracted once,
never at serve time" (`S2:363-364`). A double parse at admit is not a
correctness defect, and it is exactly the kind of shape the plan is
supposed to choose deliberately rather than inherit.

**Folded** as a new resolution R12, applied provisionally. `Read` stays
tag-only at 8b and the picture read is reached through `Parse`'s
`Container`; the double parse is named as the accepted cost, recorded as
a stated deviation from `spec:729-733` and as a declared coverage gap,
and the question of a combined entry point is handed to rung 9, the
first caller, because the shape should be chosen by the call site rather
than invented one rung ahead of it. R12 is named in Task 1's outcomes so
a `diff-reviewer` can grade the surface, and in Task 7's spec
corrections. **Conductor-owed.** The fold if it is ruled the other way
is a sibling entry point returning the tag set and the pictures from one
parse, which is a new exported symbol at a rung whose ruling 14 is about
keeping the surface minimal.

---

## MINORS

**m1. `.gitleaksignore` is a zero-byte tracked file, not the single-line
rider the plan describes.** `stat -c%s .gitleaksignore` returns 0 and
`od -c` shows no bytes; the file is tracked and was emptied by `3a35558`,
"Rename the config key fixture field to retire gitleaks' generic-api-key
finding". Global Constraints assert it as fact (`plan:365-369`, "If
gitleaks flags it, the fix is a new fingerprint appended below the
existing line"), and Task 0 item 11 (`plan:858-862`) asks for a line
number that does not exist. Folded by restating the constraint as: the
file is empty, this pass adds no fingerprint, and if gitleaks flags Task
6's JSON the fix is the file's first fingerprint line, which the task
reports. Task 0 item 11 becomes a confirmation that the file is empty.

**m2. The conductor's opening simplifier round has already landed, and
it touched no `tagio` file.** HEAD is `0f2b61d`, "Simplify pass 8a's
tagio package and WriteStream"; `git show --stat` reports one file
changed, `internal/atomicfile/atomicfile.go`, 1 insertion and 5
deletions. The plan describes the round in the future tense as "The
conductor's opening action, before Task 1" (`plan:882-904`) and
`docs/STATUS.md:23-25` still records it as owed. Task 0 item 2 catches
it, so no task stalls, but the plan's premise that the round may move a
cited `file:line` in Tasks 1 through 3 is now settled: it did not touch
`internal/engine/tagio` at all. Folded by recording the landing and its
scope in that section, reducing Task 0 item 2 to confirming `0f2b61d`
and that no `tagio` citation moved, and re-scoping the ceiling's
simplifier line to 8b's own round per M-e.

**m3. Ruling 13's confirmation cites the wrong Task 0 item.** Global
Constraints (`plan:318`) read "No task edits `internal/config`, and Task
0 item 6 confirms it." Item 6 (`plan:826-835`) is the `tagio`
exported-surface check; the `internal/config` check is item 10
(`plan:854-857`). Folded by repointing to item 10.

**m4. The spec row cited for the `internal/harness` write does not grant
it "in those words".** Global Constraints (`plan:290-296`) call Task 6's
write into `internal/harness/analysis/adjudication.go` one "which the
spec's own contested-surface row grants in those words (`spec:1165`)".
Read at the line: `spec:1165` is the `corpus/analysis/adjudications.json`
and divergence-schema row, whose words are "rung 8 adds its
applied-field divergence rows as a schema addition, never by
hand-editing a derived record". It names the JSON ledger and the schema,
never the Go file or `internal/harness`, and `RE`'s ownership rows put
`internal/harness` under rungs 2 and 3, with the constraint sentence
"Lane T runs beside rung 3 only while its plan states that its
conformance instrument is read-only against `corpus/` and
`internal/harness`". Folded by restating the grant as a reading rather
than a quotation: the divergence schema is defined in
`adjudication.go`, so a schema addition lands there. Task 0 item 1 gains
a clause confirming no lane holds `internal/harness/analysis` before
Task 6 dispatches, since item 1 covers lane 2's territory and not rung
3's.

**m5. R11 does not forbid `omitempty`, and its tie assertion points past
the table that already exists.** R11 (`plan:735-748`) adds JSON tags to
all 69 `TagSet` members and projects a read-back set by marshalling it,
then compares against a ledger row carrying all 69 keys. Measured: every
row in `corpus/analysis/file_tags.json` has exactly 69 values, with
absent keys carried as explicit nulls. A tag written idiomatically as
`json:"album,omitempty"` drops every absent member, so the projected
object would carry fewer than 69 keys and Task 5's count criterion
(`plan:1415-1416`) would fire on all twelve subjects. Separately, R11
specifies the tie positionally ("the `TagSet` member's JSON tag equals
`Mapping()`'s row at that index", `plan:1300-1303`). That is buildable
today, since the struct's 69 members are in mapping's exact order, but
`tagset_test.go:95` already carries `fieldStructName`, a hand-written
Field-to-member map that `TestTagSetMembersTieToMapping` (`:173`) grades
by name. Folded by stating in R11 that no member's tag carries
`omitempty`, and by joining the tie assertion to `fieldStructName` so
the tags, the members and the mapping rows are held together by one
literal, with the positional check kept as a second and cheaper
assertion.

**m6. Task 3 never says whether the exported `Parse` routes through the
registry, so the seam has a bypass and its own proof-test is partial.**
Task 3's outcome is that "`Read` and `Write` dispatch through the
registry" (`plan:1178-1181`), and its proof empties the registry and
observes both entry points return the sentinel (`plan:1197-1200`).
`Parse` is a third exported read entry point: it checks the `fLaC`
signature itself and returns `ErrUnsupportedContainer` directly
(`container.go:123-136`). It is what 8c's absorbed reader and the shadow
will use, and the plan says nothing about it. With the registry emptied,
`Parse` still parses the golden, so the emptied-registry test does not
establish that every container entry point is one sentinel from an
unregistered format. Folded provisionally: `Parse` stays the
FLAC-specific entry point that deliberately does not route, documented
as such, because it returns a FLAC `Container` and a registry whose
other entries return other containers has nothing to give it. The
emptied-registry subtest asserts that `Parse` still parses the golden,
so the bypass is proven deliberate rather than discovered.
**Conductor-owed.**

**m7. The writer's ownership of the PICTURE block type is never stated,
which leaves both the preserve-unknown transition and the new block's
position unpinned.** Two reviewers reached this from opposite ends and
the fold is one item. `Container.Blocks` carries every block with its
raw payload so an undecoded block "still survives a write unchanged
(C10)" (`container.go:48-63`), and `planBlocks` preserves every
non-PADDING block in place (`write.go:69-105`). Ruling 7 is "The writer
preserves unknown metadata blocks unconditionally." Once Task 1 decodes
PICTURE into a typed list, the writer must stop preserving those blocks
from `c.Blocks` or a store emits both the originals and the new set, and
the strip removes nothing. Task 2's criteria catch the symptom (a
two-picture set leaves exactly two blocks; a non-nil empty set leaves
none), and neither P2, P8 nor Task 2's outcomes state the transition, so
a reviewer grading the contract has nothing to grade against. The
position is unstated for the same reason: `planBlocks` creates a missing
VORBIS_COMMENT immediately after STREAMINFO on C9a's rule, and the
golden's chain is STREAMINFO, VORBIS_COMMENT, PICTURE, PADDING
(`golden.txt:120-124`), but a write of a picture into a container
carrying none (the conformance template, and rung 9's admit path) has no
stated position. An implementation inserting before VORBIS_COMMENT would
break C9a silently, since the byte assertion compares only the PICTURE
block. Folded by adding the ownership sentence to P8 (PICTURE joins
VORBIS_COMMENT as a block type the writer owns and re-plans, leaving the
preserve-unknown set ruling 7 governs), by adding a Task 2 outcome
fixing the position (pictures follow the VORBIS_COMMENT block and
precede the trailing PADDING; an existing PICTURE block is replaced in
place), and by adding a Task 2 criterion asserting the whole emitted
chain order for both the golden and the template.

**m8. Task 6's Interfaces line reads as an import the arch test forbids,
and its first mutation names a failure that import ban makes
unreachable.** Two reviewers, one boundary. Task 6 lists "Consumes: ...
`analysis.AdjudicationEntry`, `Pending` and `AgentAdjudicator`"
(`plan:1490-1492`) while R7 point 2 requires the `tagio` emitter to work
"over a local struct that mirrors the shape" (`plan:654-660`).
`TestEngineNeverImportsHarness` walks `internal/engine` and skips only a
`testdata` subtree, so it covers `_test.go` files
(`internal/harness/conform/arch_test.go:37-42`, `:141`). An implementer
following the Interfaces line imports `analysis` from `conform_test.go`
and turns the gate red. The same ban makes Task 6's mutation 1 ("Emit a
disposition other than pending → `CheckAdjudications` rejects the
emitted row, and the emitter's validation row fails") unreachable from
`tagio`: nothing in the engine can call `CheckAdjudications`. The
mutation is still observable, through the byte-equality assertion
against `divergence-example.json`, which is mutation 4's mechanism.
Folded by splitting Task 6's Interfaces line by package, saying
explicitly that the `tagio` half consumes none of those symbols and
reproduces `"pending-adjudication"` and `"agent"`
(`adjudication.go:18`, `:30`) as literals whose only tie is
`divergence-example.json`; by restating mutation 1's proof as the
byte-equality assertion in `tagio` plus the malformed-row table in
`internal/harness/analysis`; and by noting that the analysis-side test
supplies its own `caseIDs` slice, since `CheckAdjudications` resolves a
case-subject row against the caller's list
(`adjudication.go:156-159`, `:183-185`).

**m9. The leftovers budget line books 0.15M against a stated basis of
0.2M.** The ceiling table (`plan:228`) reads "A leftovers round, if one
is needed | 8a measured 0.2M for six findings | 0.15M". Every other line
takes its figure from its basis or explains the departure in the basis
cell, as the prep line does. This one does neither, and the 0.05M it
removes is a quarter of the 0.19M headroom the plan calls "the smallest
band that keeps the flag meaningful" (`plan:240`). Folded by giving the
basis cell its reason: seven tasks against 8a's nine, so 8a's 0.2M
scales to about 0.156M on the same rate. The figure and the arithmetic
stand. **Conductor-owed:** booking the full 0.2M instead, which puts the
expected total at 3.11M and the headroom at 0.14M.

**m10. "Deliverables 1, 2 and 6 landed at 8a" overstates deliverable 6,
and forward carry 4 is never disposed of.** `plan:183` reads
"Deliverables 1, 2 and 6 landed at 8a; deliverable 7 is 8c's."
Deliverable 6 is the golden container "carrying a VORBIS_COMMENT block
and a PICTURE block. The writer's emitted bytes for those two blocks are
asserted equal to the oracle's for the same input" (`spec:131-138`). 8a
had no picture encoder, so the PICTURE half of that byte assertion is
this pass's, and Task 2 carries it as "the only place in rung 8 where a
picture's bytes are graded against bytes another implementation
produced". Separately, `docs/HISTORY.md:375-377` carries forward carry
4, the family non-FLAC question, which the plan's Primary Inputs point
at ("the carries name what 8b decides") and no section disposes of.
Folded by saying deliverable 6's container landed at 8a and its PICTURE
byte assertion is Task 2's, and by adding forward carry 4 to Task 7's
carry list as still carried and not 8b's, so the close does not have to
rediscover it.

**m11. An inherited stale pointer is repeated rather than corrected.**
Task 7 lists a spec correction the pass owes: "`spec:1103`'s remaining
scope, if the erratum's item 1 needs any further wording"
(`plan:1608`). The spec's own erratum uses that pointer too
(`spec:36`). Line 1103 today reads "stand: pure-Go per-format readers
and writers under permissive", inside the F1 launch clause block quote;
the clause the erratum quotes, "the seam registry with its supported
set, its func-value dispatch, and its unsupported-container sentinel",
is at `spec:1134`, the 8b pass-split row. The erratum's own insertion
into the header is what moved it. `docs/HISTORY.md:356` already flags
the pointer with "(corrected, below)". Folded by repointing to
`spec:1134` and stating in Task 7 that the erratum's own reference is
the stale one being corrected, which is the correction the pass owes
rather than a note beside it.

**m12. The bolded-lead-phrase bullet runs at roughly twice the ratified
rate.** Measured, 58 of 156 top-level bullets open with a bolded lead
phrase (37%), against 38 of 212 (18%) in `plans/2026-09-07-rung-8a-
tagio-writer.md` after its own fold, and 55 of 307 (18%) in
`plans/2026-09-05-rung-4a-shadow-runner.md`. In some blocks it is every
bullet in the block, Global Constraints (`plan:272-374`) among them. The
writing-voice router's document-shape rule reads "Do not open every
bullet or paragraph with a bolded lead phrase. Sparingly it signposts;
by reflex it is the machine list default." The rest of the prose is
clean: tellgrader reports 0 tells over 13,931 words at a cadence CV of
0.75, no em dash, no connector opener and no "not X but Y" frame. Folded
by unbolding the lead phrases in the per-task Outcomes, Acceptance-
criteria and mutation lists, keeping the bold in Global Constraints, the
package contract and the Resolutions, where the lead phrase is the
item's name. The reviewer's second half, the "rather than" count, is
rejected below.

---

## Rejected findings

**Review 2's m6, "Hook 3's landing is credited to the write, but the
hook's requirement is extraction."** The finding argues that Task 2
"lands hook 3" on the strength of the write, when extraction is Task 1's
decode. Read at the lines, the plan already says what the finding asks
for. Task 1 reads "**2.0 hook: advances hook 3** ... the decode is the
half that makes extraction possible. Task 2 lands the hook"
(`plan:1000-1005`). Task 2 reads "With the read at Task 1 and the write
here, an embedded cover is extractable once at admit"
(`plan:1142-1146`). The spec makes the hook read-and-write in one
sentence: "**Landed by pass 8b.** The PICTURE read and write is what
makes extract-once possible" (`spec:1041`). Marking the later of two
tasks "lands" and the earlier "advances" is the plan's own convention
across every hook statement, and Task 7 records the hook against the
pass rather than a task. No edit follows, so no edit was made.

**Review 2's m10, the "rather than" half.** Measured at 59 occurrences
over 13,931 words, which is 4.2 per thousand. The ratified exemplar
`plans/2026-09-05-rung-4a-shadow-runner.md` runs 173 over 41,108 words,
also 4.2; `plans/2026-09-07-rung-8a-tagio-writer.md` runs 74 over
18,823, or 3.9. The plan sits at the house rate for a document class
whose whole job is stating what a thing is instead of what a reader
might assume. Varying a connective that is already at the ratified rate
buys nothing. The bolded-lead half of the same finding survives as m12
on its own measurement.

---

## Verified sound

Do not re-litigate these. Each was checked by the fold editor against
the named artifact.

1. P5's front-cover rule is exactly `CoverArtField.guess_cover_image`:
   single candidate first, then the first with type front, then the
   first of any type (`mediafile/fields.py:252-259`). The
   single-candidate-first ordering holds, and mutation 5's
   lone-back-cover row is the right probe.
2. P6's zero geometry is right and both citations are exact.
   `FlacImageStorageStyle.serialize` sets only data, type, mime and desc
   (`mediafile/storage/flac.py:30-37`), and `Picture.__init__`
   initialises type, mime, desc, width, height, depth, colors and data
   (`mutagen/flac.py:577-586`). The golden's own block carries geometry
   0, 0, 0, 0, confirmed by decoding `golden.flac` directly.
3. P8's replace-and-clear semantics are exact: store calls
   `clear_pictures()` then `add_picture()` per picture
   (`storage/flac.py:24-28`), and delete is `clear_pictures()` alone
   (`:39-41`).
4. P7 holds. mediafile computes the MIME from the bytes through
   `Image.mime_type` into `filetype.guess_mime`
   (`mediafile/utils/image.py:12-14`, `:49-52`) and never reads a
   caller-declared type. `mime_type` returns `None` both for an
   unsniffable blob and for empty data, and that `None` reaches mutagen,
   so R2's refusal is the closest honest reproduction.
5. P4 is exact. mediafile's `ImageType` runs `other=0` through
   `publisher_logo=20` with `front=3` (`constants.py:21-44`),
   `Image.__init__` degrades an out-of-range integer index to `other`
   (`utils/image.py:41-47`), and `type_index` returns 0 for an unset
   type (`:54-60`). Index 21 and above are the degrade cases, and 20 is
   not.
6. R3's APNG description is exact. `filetype` tries `Apng` before `Png`
   in the IMAGE tuple (`types/__init__.py:15-32`), and `Apng.match`
   walks the chunk table from offset 8, returning False at IDAT or IEND
   and True at acTL (`types/image.py:51-88`). JPEG (`FF D8 FF`), PNG
   (four-byte `89 50 4E 47`), GIF (`47 49 46`) and BMP (`42 4D`) are all
   as the plan states.
7. The golden's PICTURE block matches every claim Task 0 item 7 and Task
   1 make: block type 6 at offset 1112, not last, declared length 78,
   decoding to type 3, `image/png`, description `cover`, geometry all
   zero, and 32 data bytes opening with PNG's eight-byte signature.
   4 + 4 + 9 + 4 + 5 + 16 + 4 + 32 = 78 exactly. The full 32-byte data
   literal is recoverable from `golden.txt:72` and
   `gen_golden_flac.py:115`, so R4's no-new-binary-fixture position is
   workable.
8. P2 is right that multiple PICTURE blocks are legal where SEEKTABLE
   and VORBIS_COMMENT are capped (`INV:770-780`), and mutagen's write
   order matches RFC 9639 section 8.8 field for field
   (`mutagen/flac.py:612-623`).
9. Task 4's viability premise holds on measurement. All twelve ledger
   subjects' committed `case.json` `expected.applied_fields` carry all
   69 `writable_fields` (the clean-match item 0 carries 99 keys total),
   so the filter-to-69 criterion is provable and no subject is missing a
   writable key.
10. Task 4's four-state table transcribes correctly from
    `corpus/cases/clean-match/case.json` item 0: `album` is `Rumours`,
    `acoustid_fingerprint` is present and empty, `albumartists` holds
    one entry, `arrangers` is the empty list, `bpm` is the integer 0,
    `comp` is false, and seven writable keys are JSON null
    (`initial_key` and the six gain and peak fields).
11. R11's unmarshal fill and marshal projection are type-safe across the
    whole corpus. Every non-null value in both the twelve cases' applied
    fields and the twelve ledger rows matches the Go kind of its
    `TagSet` member: no float where an int is declared, no string where
    a list is, `comp` always a bool. `json.Unmarshal` of `[]` into a
    `[]string` member yields the non-nil zero-length slice R11 depends
    on, and of `""` into a `*string` yields a pointer to the empty
    string.
12. R11's positional tie assertion is satisfiable. `TagSet`'s struct
    field order (`tagset.go:115-180`) matches `mapping`'s row order
    exactly (`:243-305`), including the date-triple substitution that
    places Year, Month and Day after Country rather than alphabetically.
13. The grading cannot pass vacuously at subject granularity. Every
    ledger row carries between 37 and 51 non-empty values, so Task 5's
    mutation 3 fails all twelve subjects rather than merely some. No
    ledger value is a non-null empty list and none is a float, so the
    empty-list-versus-absent ambiguity (`deviations.go` item 3) and
    R10's float-tolerance hazard do not arise in this corpus.
14. The instrument's step 2 and step 3 shape is faithful even though the
    plan omits the shim's beets `Item` construction:
    `derive_file_tags.py` builds an `Item` from the applied fields
    (`derive_oracle.py:156-165`, whose `item[key] = value` normalizes
    through beets' type system) before `item.write()`. Measured across
    all twelve subjects, applied value and ledger read-back differ on
    only eleven keys, and every one is a mediafile write-read semantic
    8a already models: an empty list writes no comment and reads back
    absent, an empty `catalognum` or `barcode` reads back absent through
    the last-applicable-style rule, and a zero month or day drops out of
    the composed DATE. No beets normalization altered a value, so
    filling from the raw applied fields is the right procedure for this
    corpus.
15. P9 and R9 satisfy ruling 14 as written. A map keyed by an exported
    format type whose value is a struct of func values plus a magic
    prefix declares no interface, exported or not (`spec:314-319`), and
    the sentinel is adopted rather than redefined, which is what the 8a
    handoff hands forward (`docs/HISTORY.md:352-357`).
16. The territory and arch claims check out.
    `TestEngineNeverImportsHarness` and `TestEngineNeverReadsTheClock`
    both root at `internal/engine` and walk `_test.go` files outside a
    `testdata` subtree (`arch_test.go:37-42`, `:82-87`, `:141`,
    `:152`), so the instrument's local decode types and R7's
    parameterised date are both required rather than defensive.
17. Task 6's schema addition is compatible with the existing checker.
    `CheckAdjudications` requires `suspected`, `release_truth`,
    `reason`, `adjudicator` and `date` on every row and resolves a
    case-subject row against a caller-supplied `caseIDs` slice
    (`adjudication.go:156-214`), so an invented example row validates
    without touching the corpus, and no existing field's requirement
    needs to change.
18. The ceiling arithmetic is internally consistent and re-derived from
    8a's actual as the spec's clause requires (`spec:1194-1196`).
    0.5 + 0.2 + 0.1 + 1.76 + 0.2 + 0.15 + 0.15 = 3.06M. The chain plus
    the upshift reaches 2.76M at the end of Task 7, an 80% flag against
    3M lands at 2.4M inside Task 6, and the 3.25M flag lands at 2.6M
    inside Task 7. Headroom is 0.19M.
19. The 8a spend basis is quoted exactly from `docs/HISTORY.md:403-410`:
    about 3.6M against 4.5M, prep 0.87M, Task 0 0.1M, the chain 2.27M
    across 24 agents, leftovers 0.2M, conductor turns about 0.15M. The
    reasoning that the 24-agent chain line already contains every
    `diff-reviewer` pass and fix re-dispatch, so a separate fix line
    would double-count, holds.
20. The rung reservation arithmetic is right: 4.5 + 3.25 + 1 = 8.75M
    against the spec's 8M reservation at `spec:1181`, with 8a's 0.9M
    underspend as the stated offset.
21. The plan's cited anchor tests exist. `TestFieldCountAndNames` is at
    `tagset_test.go:39`, and `TestTagSetMembersTieToMapping` at `:173`,
    tying members and mapping rows by name through `fieldStructName`
    (`:95`).
22. Every value Task 4 hard-codes from clean-match item 0 is transcribed
    correctly, and the four named non-writable leftovers
    (`data_source`, `album_id`, `id`, `bitrate`) are all present in the
    case and all absent from `writable_fields`, so the filter assertion
    is provable.
23. Task 0 item 3's expected ledger header is exact: `shim_version` 1,
    `beets_version` 2.13.1, 69 `writable_fields`, 9 `readonly_fields`,
    twelve rows, and the twelve case ids match
    `corpus/README.md:2182-2192`.
24. Task 0 item 6's expected exported surface matches the tree,
    `Deviations` included. `Write`'s current signature is
    `Write(path string, ts TagSet, mode os.FileMode) error`
    (`write.go:46`), so Task 2's widening between the tag set and the
    mode is well defined.
25. Task 0 item 8's expectations are right, and the existing case-row
    rule that rejects `matcher_commit` on a case subject does not
    collide with R7, because R7 puts the `tagio` commit inside the
    applied-field block rather than in the top-level stamp.
26. Task 0 item 1's expected lane registry matches `docs/STATUS.md:23-26`
    verbatim: lane-1 main tree idle with "author pass 8b's plan", lane-2
    worktree idle with "author pass 5a's plan".
27. Task 7's coverage-gap total is right: 21 + 9 + 4 + 2 + 7 = 43,
    matching `docs/STATUS.md:30-32`.
28. Global Constraints' dependency claim is right: `go.mod` carries
    cobra with mousetrap and pflag as its two indirects and nothing
    else. Nothing outside `internal/engine/tagio` calls `tagio.Write`,
    `tagio.Read` or `tagio.Parse`, so Task 2's signature widening is
    package-local.
29. The quote the plan attributes to 8a's R5 is verbatim from the 8a
    plan at its lines 880-884, including "That is one refactor inside
    `tagio` with no consumer wired yet", and it does name the
    `Container` widening and the picture argument as 8b's.
30. R7's rung-4 precedent citations resolve. `pendingJSON`
    (`cmd/dubplate-server/shadow.go:158-186`) sets `Disposition` and
    `Adjudicator` as constants rather than reading them off the row, and
    `PendingRowFor` is at `internal/engine/shadow/run.go:349` onward.
31. `RE` cells (8,3) and (8,4) say what the plan says: the instrument is
    read-only against `corpus/` and `internal/harness`, it pins no
    schema version and never calls `Reconcile`, and `RE:192`'s paragraph
    records that rung 8 has no pending-row escape where rung 3 has one,
    which is the gap Task 6 closes.
32. The corpus handoff supports the plan's use of it: the 69 writable
    names at `corpus/README.md:2378-2400`, the nine read-only names and
    the never-compare-per-subject rule at `:2401-2412`, and the
    hand-written status of `adjudications.json` at `:2255-2258`.
33. The pass stays inside 8b. No task touches `internal/engine/shadow`,
    `buildFLAC`, a new corpus subject, `make derive`, `internal/config`
    or `go.mod`, and the package-territory wall names every excluded
    tree explicitly.
34. Both named carries from the 8a handoff are dispositioned. The
    simplifier round has its own conductor section and a Task 0 blocker
    item, and the `io.ReaderAt` gap is declined at R8 with the gap kept
    declared, carried by Task 7, and reassigned from "8b's or 8c's" to
    8c's alone.
35. The staleness pre-task names both things it was required to name:
    lane 2's worktree reach into 8b's territory, with lane 2's own
    territory stated as `internal/serve` and `web/`, and the simplifier
    round, marked a blocker for Task 1 rather than for Task 0.
36. The independence marks are true. Only Task 6 is file-independent of
    Tasks 1 through 3, and it still depends on Task 5 by data. Every
    other pair is contended through `container.go`, `write.go`,
    `read.go` or the widened `Write` signature, and the plan marks no
    pair for parallel execution.
37. The mutation proofs each name a failing assertion, and the three
    deliberately unfired ones are honest rather than hidden. Task 2's
    deviation-note mutation, Task 4's hard-coded-subject-list mutation
    and Task 5's float-tolerance mutation each say "no test fails" and
    name the reviewer's read of the diff as the criterion that holds
    them.
38. The comparison design cannot be defeated by a mirrored bug. The
    ledger is external evidence produced by pinned beets and read back
    by an independent mediafile parse, so a `tagio` writer defect that a
    `tagio` reader defect would cancel still diverges from the ledger
    row. R10's refusal of any tolerance follows from both sides
    originating as decimal text.
39. The resolution count matches: R1 through R11 is eleven, which is
    what Task 7's close is told to retire. Revision 2 adds R12, so the
    close retires twelve.
40. The deliverable mapping is right against `spec:106-140`: deliverable
    3 is PICTURE, 4 is the per-format seam, 5 is the conformance
    instrument, and 7 is the absorption that belongs to 8c.
41. The load-bearing spec, HISTORY and corpus README citations resolve.
    `spec:1134` is the 8b pass-split row, `spec:239-249` rulings 2 and
    3, `spec:314-319` rulings 13 and 14, `spec:1041` hook 3,
    `spec:1245-1252` risk 1, `spec:1282-1285` risk 5, `spec:1303-1305`
    risk 8, `spec:1197` the staleness pre-task, `spec:1339`
    Universality; `HISTORY:345` the handoff, `:365` the forward carries,
    `:403` the spend, `:82-89` the exported surface;
    `corpus/README.md:2354` the handoff.
42. The prose meets the developer-docs standard on the mechanical
    checks: zero em dashes, no connector openers, no "not X but Y"
    contrast frames, and tellgrader reports 0 tells at a cadence CV of
    0.75.
