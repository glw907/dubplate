# Adversarial review: the rung 8 tag I/O spec

Two fresh-context reviews (2026-09-07) of
`specs/2026-09-07-rung-8-tag-io.md` at revision 1, one through the
FLAC-and-mediafile-0.17.0 fidelity lens and one through the
spec-integrity-and-scope lens, merged here and folded as revision 2 of
the spec the same day. The merged "Verified sound" list below is the
do-not-re-litigate record for every rung 8 pass plan.

---

Findings are renumbered across both reviews. Two findings that report
the same defect from opposite ends are merged into B2 and name both
readings. Every finding was re-verified before folding: encoding claims
against the pinned venv (mediafile 0.17.0, mutagen 1.48.1, beets 2.13.1
under `tools/venv/.venv`), container claims against RFC 9639 read from
the published RFC, ledger claims against `corpus/analysis/file_tags.json`
and the twelve `corpus/cases/<id>/case.json` files, and tree claims
against the worktree at `c98a916`. Live measurements ran against scratch
copies of `internal/harness/oracle/testdata/template.flac` in a temp
directory. Nothing in the repo was written by the review.

Half of one finding is rejected, one finding's evidence was wrong in its
numbers, and one finding quotes a string the spec does not contain. All
three are below.

## Rejected and corrected, with the measurement

**Rejected: the consequence half of the last-block-flag finding.** The
fidelity lens argued that a writer leaving a preserved block's last-block
flag set would pass every named proof, because the repo's reader stops at
the first block whose flag is set (`internal/engine/shadow/flac.go:100`,
`:122`) and would report the truncated chain as complete. The reader
behavior is right. The conclusion does not follow, because the spec
already names a second assertion over the same chain: "A write emits a
trailing PADDING block of 8 KiB. Proof: an assertion on the written block
chain" (spec:368-370). A reader that stops early never reaches the
appended PADDING, so that assertion fails loudly rather than passing. The
failure mode is a red test, not a shipped defect. What survives is the
contract half, which is real and unstated: the spec never says the writer
owns the flag. It folds as minor m5 rather than as a major.

**Corrected: the catalognum and barcode row counts.** The fidelity lens
reported catalognum applied `""` with a null ledger value on 6 of 12
subjects and barcode on 4 of 12. Recomputed against the committed ledger
and the twelve case files, the counts are 5 and 6. The defect the counts
support is unaffected and folds as B1.

**Corrected: a quoted unit-test expectation.** The fidelity lens quotes
the ruling 1 unit tests as hard-coding `"-7.50 dB"`. No such string
appears in the spec; `grep` over the file returns nothing. The substance
holds, because the spec does say the expected string is hard-coded rather
than derived (spec:396), and a hard-coded serialized string pins no byte
layout. The finding folds as the second half of B2 without the invented
quotation.

---

## BLOCKERS

**B1. The empty-string read-back rule is wrong for `catalognum` and
`barcode`, and the committed ledger already contradicts it.** (Fidelity
lens.)
The spec states as a contract that "An empty string writes a comment with
an empty value and reads back as `\"\"`" (spec:239-245), and the
conformance instrument's step 4 restates it as "the null versus
empty-string distinction" (spec:577-579). That holds only for a field
whose every FLAC style is writable. `MediaField.__get__` overwrites `out`
on each style and breaks only on a truthy value, so a field whose trailing
styles are read-only returns the LAST style's value when nothing is truthy
(`mediafile/fields.py:48-54`). Two of the 69 are shaped that way:
`catalognum`, whose `CATALOGNUMBER` is writable while `CATALOGID` and
`DISCOGS_CATALOG` are read-only (`mediafile/__init__.py:604-606`), and
`barcode`, whose `BARCODE` is writable while `UPC`, `EAN/UPN`, `EAN`, and
`UPN` are read-only (`:616-620`). Measured on a scratch copy of the
committed template: writing `catalognum=""` and `barcode=""` puts
`CATALOGNUMBER=` and `BARCODE=` on disk and reads both back as `None`,
while `asin=""` (one style) and `albumartist=""` (three writable styles)
both read back as `""`. The ledger agrees. `catalognum` is applied `""`
and recorded null on 5 of 12 subjects, and `barcode` on 6 of 12.
*Fix:* state the general rule in the mechanism section. A field's
read-back is the last applicable style's value when no style yields a
truthy one, so a field with trailing read-only styles reads back `None`
for an empty value even though the writer emitted an empty comment. Name
`catalognum` and `barcode` as the two members of that class. Make the
instrument's step 4 projection implement the last-style rule rather than a
first-truthy rule with an empty fallback, and pin both fields with a
ruling-1-style unit test.

**B2. Nothing anchors the writer's bytes to bytes another implementation
produced, so two whole classes of defect pass every stated gate.** (Both
lenses, from opposite ends.)
Every proof in "What the writer guarantees" and both write-side exit gates
close a loop inside `tagio`. The conformance instrument writes with
`tagio` and reads back "through an independent parse" that is the same
package's parser (spec:577-579). The block-preservation round trip, the
frame-region hash, the colliding-pair order test, and the PICTURE round
trip are all self-round-trips. The ruling 1 unit tests hard-code
serialized strings rather than block layouts (spec:396). The one committed
artifact another implementation produced is `template.flac`, which is 42
bytes of signature, header, and STREAMINFO, and carries no VORBIS_COMMENT
and no PICTURE; its two pinning tests move in at 8c (spec:673-675). Two
consequences follow, one from each lens.

The fidelity reading: a writer that emitted the VORBIS_COMMENT vendor
length, comment count, or per-comment lengths big-endian, the exact
asymmetry the spec itself flags (spec:300-301), would read its own files
back correctly, agree with all twelve ledger rows on all 69 keys, and ship
a library no other player can read.

The integrity reading: the ledger stores mediafile's read-back VALUES, not
the comment set. `derive_file_tags.py:154` is
`values = {field: getattr(mf, field) for field in writable}`, and
`corpus/README.md:2398-2400` states the proof shape the same way. So the
instrument cannot see a violation of the spec's own
write-every-applicable-style contract (spec:247-254). A writer that emits
only `ALBUMARTIST` and skips `ALBUM ARTIST` and `ALBUM_ARTIST` reads back
identically through mediafile and through `tagio`, because both take the
first truthy style, and every subject and every key still agrees. The same
holds for `disctotal`, `tracktotal`, `track`, `disc`, `label`, and
`albumstatus`. The consequence is real rather than theoretical, since
Navidrome and other readers look for `DISCTOTAL` and `TOTALDISCS`, so a
file that passes every gate can still be under-tagged for every consumer
but `tagio`.

The spec then declines the artifact that would close the byte half: "Two
shapes are declined: ... and a new binary fixture, because the instrument
reads two files the repo already tracks" (spec:598-602).
*Fix:* commit one small golden container produced by the pinned oracle
into `internal/engine/tagio/testdata`, a mutagen-written FLAC carrying a
VORBIS_COMMENT block and a PICTURE block, generated once under the venv
the way `template.flac` was. Assert two things against it: the Go writer's
emitted VORBIS_COMMENT and PICTURE block bytes equal the oracle's for the
same input, and the Go reader parses the golden. `go-conventions` permits
a golden exactly here, at a "high-fidelity render surface where the value
is the literal bytes". Narrow the declined-fixture sentence so it declines
a second conformance corpus rather than every new binary file, and record
the narrowing. Separately, add a byte-level assertion of the emitted
comment set to 8a's exit gate, graded against `INV:409-427`'s measured
fifteen-field write, and state in "What the writer guarantees" that the
conformance instrument grades read-back values and therefore proves
nothing about which styles were written.

---

## MAJORS

**M1. The writer's VORBIS_COMMENT merge semantics are unstated, so a
rebuild-from-the-69 writer silently drops every foreign comment the oracle
preserves.** (Fidelity lens.)
Ruling 7 and the preservation guarantee cover metadata BLOCKS
(spec:193-196, spec:334-339). Nothing states what happens to comments
inside the VORBIS_COMMENT block that the 69-key tag set does not own. The
oracle preserves them. `MediaFile.update` touches only the fields present
in its dict (`mediafile/__init__.py:365-378`), and mutagen's
`VCommentDict.__setitem__` deletes only the case-insensitive matches of
the key being written before appending (`mutagen/_vorbis.py:275-297`).
Measured: a template copy carrying `COPYRIGHT=(c) me` and
`ORGANIZATION=label` still carried both after a mediafile write of
`title`. `copyright` is a MediaFile field (`__init__.py:519`) and is not
one of the 69, so beets never touches it. A writer built from this spec's
69-key value model would rebuild the block and drop them on admit and on
every rung 9 retag. The instrument cannot catch it, because `template.flac`
has no comments at all.
*Fix:* state the merge contract. A write replaces only the keys the tag set
owns, matching key names case-insensitively, deletes the read-only styles
of a deleted field, and leaves every other comment untouched and in its
original order. Add an exit-gate assertion over a fixture whose comment
block carries a foreign key.

**M2. The vendor string is missing from the container contract and from
the parsed-container type, so the writer would drop or reinvent it.**
(Fidelity lens.)
RFC 9639 section 8.6 specifies the block as a little-endian vendor length,
the vendor string, a comment count, then the comments. The spec describes
VORBIS_COMMENT only as field-name rules, the repeat rule, and the absent
framing bit (spec:310-314), and describes the parsed block as
"VORBIS_COMMENT decoded to its comments with repeats preserved"
(spec:470-473). Neither mentions the vendor. The oracle writes one:
measured, a mediafile write into a scratch template copy produced a block
whose vendor is `Mutagen 1.48.1` (`mutagen/_vorbis.py:73` sets the
default, `:185-187` writes it), and a rewrite of a file that already has a
block preserves the loaded vendor (`:104`).
*Fix:* add the vendor string to the container contract and to the
parsed-block type. Read it, preserve it on rewrite, and state what the
writer emits when it creates the block fresh. Assert vendor preservation
in the block round-trip test.

**M3. An empty list is a fourth input state the three-behavior rule does
not cover, and 45 graded cells carry it.** (Fidelity lens.)
The spec states three behaviors (spec:239-245). For a list-backed field
the empty value writes nothing at all: `ListStorageStyle.set_list` stores
the serialized list, so an empty list stores an empty value list and
mutagen emits zero comments
(`mediafile/storage/base.py:189-200`). Measured: `arrangers=[]` and
`genres=[]` produced no `ARRANGER` and no `GENRE` comment and read back as
`None`. This is not a corner case. Across the twelve subjects' applied
fields, 45 cells are `[]`, against 149 empty strings. An implementer
applying the stated rule uniformly writes `ARRANGER=` and reads back
`[""]`, failing those 45 cells.
*Fix:* extend the rule to four input states, stated per Go type. A string
field's empty value writes one empty comment. A list field's empty value
writes no comment and is indistinguishable on read-back from absent. State
it where the tag set type is specified, since that is where the state
model is chosen.

**M4. The date rule omits the four-digit zero padding and the year-zero
case, which is beets' own default and ungraded by the corpus.** (Fidelity
lens.)
The spec's date paragraph gives the value as `YYYY`, then `-MM`, then
`-DD`, and covers only the `None` year and the 2001/0/0 example
(spec:267-273). `_set_date_tuple` formats the year as `f"{int(year):04d}"`
and the components as `:02d`, and it deletes only on `year is None`
(`mediafile/fields.py:182-201`). A fresh beets `Item` defaults year,
month, and day to the integer 0 (measured), so the unknown-date path is
the common production path once rung 9 writes tags. Measured on a scratch
template copy, an update with every date component 0 wrote `DATE=0000`,
`YEAR=0`, and `ORIGINALDATE=0000`, and read back year 0 with month and day
`None`. Note the asymmetry: `DATE` is zero-padded to four digits and the
`YEAR` fallback is not, because the fallback style serializes the int
through `str()` (`mediafile/storage/base.py:109-123`). No subject has a
zero year; the twelve run from 1975 to 2025, so nothing in the conformance
set would catch a writer that emits `DATE=0` or deletes the field.
*Fix:* state the year encoding exactly. Four-digit zero-padded in `DATE`
and `ORIGINALDATE`, two-digit zero-padded month and day, the `YEAR`
fallback unpadded, and a zero year written rather than deleted, since only
a null year deletes. Add the zero-year and sub-1000-year cases to the
ruling 1 unit tests, because the ledger cannot grade them.

**M5. Ruling 10 states the wrong decode rule for invalid UTF-8. The
oracle replaces, it does not ignore.** (Fidelity lens.)
Ruling 10 says invalid UTF-8 in a value "is decoded with errors ignored,
the way `safe_cast` does" (spec:203-206). Both halves are off. On the read
side mutagen decodes FLAC comment bytes with `errors='replace'`,
substituting U+FFFD: `VCFLACDict.load` passes `errors='replace'`
(`mutagen/flac.py:352-353`) into `VComment.load`, which decodes the vendor
and each comment with that mode (`mutagen/_vorbis.py:104`, `:109`). On the
write side, the decode-with-ignore branches in `safe_cast` and `serialize`
fire only when the value is `bytes`
(`mediafile/utils/type_conversion.py:21-22`, `:37-38`;
`mediafile/storage/base.py:120-121`), and beets item values reaching
`MediaFile.update` are `str`, so that branch never fires in the oracle's
path. Rung 7 admits files uploaded by people the estate does not fully
trust, so a comment with invalid UTF-8 is reachable, and a Go reader
carrying the raw bytes through projects a different string than mediafile
for the same file.
*Fix:* restate the rule. On read, invalid UTF-8 in a comment is replaced
with U+FFFD per code point, matching mutagen's `errors='replace'`. On
write, values are valid UTF-8 by construction. Cite `mutagen/flac.py:352`
and `mutagen/_vorbis.py:109` rather than `safe_cast`, and add a unit test
over a comment carrying an invalid byte sequence.

**M6. PICTURE geometry is described as preserved, and the oracle zeroes
it.** (Fidelity lens.)
The spec says width, height, colour depth, and indexed-colour count "stay
at whatever the reference left them, because mediafile sets none of them"
(spec:425-430), and the types section says geometry is "carried on read
and left at the reference's values on write" (spec:476-478). mediafile's
FLAC image write builds a fresh `mutagen.flac.Picture` and sets only data,
type, mime, and desc (`mediafile/storage/flac.py:30-37`). `Picture.__init__`
initialises width, height, depth, and colors to 0 (`mutagen/flac.py:577-586`),
and `store` calls `clear_pictures()` before adding (`storage/flac.py:24-28`).
So the oracle writes zeros for all four and destroys any geometry an
incoming file carried. "Left at the reference's values on write" reads as
preserve-on-round-trip, which would diverge from the oracle on any
real-world file whose cover block has real dimensions.
*Fix:* state it as it is. A write emits zero for all four fields, and a
rewrite discards whatever the incoming picture carried, because the
reference clears and rebuilds every picture. Keep the deviation note, and
make the 8b assertion "geometry is zero after a write" rather than
"geometry is unset".

**M7. Existing PADDING blocks have no stated fate, so a retag either grows
the file 8 KiB each time or silently drops padding.** (Fidelity lens.)
Ruling 7 enumerates what is preserved and PADDING is not on the list
(spec:193-196). Ruling 6 and the guarantee say a write emits an 8 KiB
PADDING block (spec:187-192, spec:368-370). Neither says whether incoming
PADDING blocks are dropped, kept, or resized. RFC 9639 section 8.3 permits
more than one padding block, so keeping them and appending is legal and
unbounded, and rung 9's retag verb runs repeatedly over the same library
file. The oracle resizes instead of appending: measured, a mediafile save
into the 42-byte template produced a chain of STREAMINFO, VORBIS_COMMENT,
and exactly one 1024-byte PADDING block.
*Fix:* state the padding policy as one rule. The writer drops every
existing PADDING block and emits exactly one 8 KiB PADDING block as the
last metadata block, so a file retagged N times has the same padding as a
file retagged once. Assert "exactly one PADDING block, 8 KiB" in the
block-chain assertion rather than "a trailing PADDING block".

**M8. The streaming atomic write does not say what file mode the renamed
file gets, and the default narrows a library file to 0600.** (Fidelity
lens.)
The write path names the steps (spec:529-533) and ruling 11 puts the new
function "beside `WriteFile`" (spec:207-209). `WriteFile` takes perm
explicitly and chmods the temp file before the rename
(`internal/atomicfile/atomicfile.go:15`, `:35-37`) precisely because
`os.CreateTemp` creates 0600. The streaming function's signature and mode
contract are unstated, so the natural implementation renames a 0600 file
over a 0644 library file. On the box the library is read by the streaming
service rather than by the writer's user, so a retag would make the track
unplayable while leaving the audio bytes perfect. No test in the exit gates
would see it.
*Fix:* state that the streaming function takes the mode as a parameter the
way `WriteFile` does, and that rung 9 passes the target's existing mode by
stat-before-write so a retag preserves it. Add the mode assertion to the
`atomicfile` test extension the proof already names.

**M9. The conformance instrument reads the twelve case files, which its
own contested-surface row and gate paragraph do not name.** (Fidelity
lens.)
Step 2 takes "the subject's frozen `Expected.AppliedFields` from the
committed case" (spec:571-574). Those values live in
`corpus/cases/<case_id>/case.json` under `expected.applied_fields`, keyed
by item index. The ledger's rows carry only `case_id`, `item_index`, and
`values`, verified against the committed file. The twelve rows name twelve
distinct case ids, so the instrument reads fourteen tracked files: twelve
case files, the ledger, and the template. The contested-surfaces row names
two of them (spec:818), and the gate paragraph declines a binary fixture
"because the instrument reads two files the repo already tracks"
(spec:602). A reviewer checking the ownership table would read the
case-file access as out of territory.
*Fix:* amend the contested-surface row and the gate paragraph to name
`corpus/cases/<id>/case.json` as a third read-only input. Record alongside
it that all twelve subjects' applied fields cover all 69 writable keys,
verified, which is what makes step 2 viable at all.

**M10. Ruling 9's strictness lands on the shadow's live read path at 8c,
against a stated no-change invariant the gate cannot test.** (Integrity
lens.)
Ruling 9 refuses more than one VORBIS_COMMENT block, a STREAMINFO that is
not first, or a declared length running past the end (spec:198-202). 8c
repoints the shadow's read path onto `tagio` while asserting "The shadow's
behavior does not change, and its tests prove that" (spec:677-680), gated
by "the shadow's own seven tests passing unchanged in behavior"
(spec:802). The landed reader tolerates exactly what ruling 9 refuses. It
switches on block type with no ordering and no uniqueness requirement
(`internal/engine/shadow/flac.go:110-121`) and merges a repeated key
first-seen-wins (`:174-177`). So a real inbox FLAC with two VORBIS_COMMENT
blocks or an out-of-order STREAMINFO moves from captured and scored to
discarded whole, on a box where the shadow is live from rung 4's deploy.
No test can distinguish the two levels, because every shadow test builds
its container with `buildFLAC` (`flac_test.go:22`), which emits a
well-formed container.
*Fix:* state whether `tagio`'s reader carries the writer's strictness. If
it does, say that 8c deliberately tightens capture, name the effect on a
live shadow run, and add a capture-path test asserting the new refusal. If
it does not, name the read path's leniency explicitly, including the
first-seen-wins rule for a repeated key that the shadow's projection must
preserve.

**M11. Deliverable 1, the 69-key reader, is not assigned to any pass in
the split table.** (Integrity lens.)
The deliverables open with the FLAC reader widened to all 69 writable keys
(spec:57-61), and the caller-facing surface includes a read entry point
(spec:480-481). The pass table gives 8a "the 69-key value model with every
serialization rule", where every listed rule is write-side (spec:800). 8b
gets the instrument, whose step 4 is the read-back projection
(spec:577-579, spec:801). 8c absorbs the shadow's nineteen-field reader
(spec:802). The spec calls that projection "a full reimplementation of
`MediaField.__get__` for the 69 fields, including first-truthy-style
precedence, `safe_cast`, the date tuple, and the Q-number division"
(spec:583-587), the largest single unbudgeted item in the rung. A plan
author cannot tell whether it is 8a's package code or 8b's
instrument-local code.
*Fix:* name the pass that builds the 69-key read projection and say
whether it is the package's exported read entry point or code local to the
instrument. If it is 8a's, add it to 8a's scope cell and its ceiling. If
it is 8b's, say that 8a's read-back agreement guarantee is not provable
until 8b.

**M12. 8a's 3M ceiling is below what the inventory's own budget table
predicts for a pass that includes plan authoring.** (Integrity lens.)
The spec reproduces INV's figures faithfully (spec:833-842) and then
proposes 3M/3M/1M. INV's table shows every pass whose spend included plan
authoring closing between 2.84M and 5.15M (`INV:1069-1074`), and the one
pass that closed at 2.7M, 4a-ii, carried no plan-authoring prep at all,
with "staleness and extraction 0.26M" (`INV:1075`). The closest
comparable, 4a-i, spent 4.99M with prep 2.3M and chain 2.14M for the same
M band and the same pure-Go-against-an-external-contract shape. 8a is the
largest of the three passes yet carries the same ceiling as 8b. Under
CLAUDE.md the conductor must stop and ask Geoff one combined question at
80% of the ceiling, which would manufacture exactly the owner interruption
spec:962-966 says the rung does not expect.
*Fix:* raise 8a, or state that plan-authoring prep sits outside the
per-pass ceiling and say where it is counted. Keep the promise that 8b and
8c are re-derived from the previous pass's actual report.

**M13. The absorption's disposition of the STREAMINFO layout test
contradicts ruling 4.** (Integrity lens.)
Ruling 4 is "`buildFLAC` does not move", with the shadow's three external
uses switching to fixtures built by the real writer (spec:180-183). The
absorption paragraph then says "Two tests pin the format to something
outside the package and move with the reader: the committed-template parse
and the byte-for-byte STREAMINFO layout assertion" (spec:673-675). The
second of those asserts `buildFLAC`'s own output rather than the reader's:
`TestBuildFLACEmitsTheStreaminfoLayoutByteForByte` calls
`buildFLAC(t, nil, 441000, 44100)` and compares the emitted 34-byte
payload against a hand-written expectation (`flac_test.go:273-295`). Its
doc comment states why the pair exists: "the builder and the parser could
agree on a wrong layout forever; these two pin each of them to something
outside this package" (`flac_test.go:253-258`). Moved to `tagio` as
written it exercises a function that stayed in `shadow`. Left in `shadow`
after the three external uses switch to writer-built fixtures and the
reader's own tests leave, `buildFLAC` has no caller and the writer-side
external pin is silently dropped.
*Fix:* say that the writer inherits the external STREAMINFO pin as its own
byte-for-byte assertion on an emitted block, and that `buildFLAC` and its
layout test are deleted once nothing calls them. Name that as a task item
rather than leaving it inside "the two format-pinning tests move".

---

## MINORS

**m1. RFC 9639 has no section 8.6.4; the field-name rule is in section
8.6.** The spec cites "section 8.6.4" three times, at ruling 10
(spec:203), the container contract (spec:312), and the refusal guarantee
(spec:356). RFC 9639's section 8.6 has exactly two subsections, 8.6.1
Standard Field Names and 8.6.2 Channel Mask, and the sentence "The field
name MUST only consist of UTF-8 code points U+0020 through U+007E,
excluding U+003D, which is the = character" sits in 8.6 itself. Verified
against the published RFC. INV inherits the same citation, so the
correction belongs in both documents.

**m2. The oracle's field-name bound is stricter than the RFC's, so an
RFC-legal name is silently dropped on read-back.** Ruling 10 and the
refusal contract set the writer's bound at U+0020 through U+007E excluding
U+003D. mutagen's `is_valid_key` accepts only 0x20 through 0x7D excluding
`=` (`mutagen/_vorbis.py:25-41`), and `VComment.load` silently skips any
comment whose key fails that test (`:127-128`) while `validate` raises on
write (`:150-151`). The only differing code point is U+007E, and the
consequence is asymmetric: a comment the Go writer preserves under the RFC
bound vanishes from mediafile's read of the same file with no error
anywhere. Pick one bound and record the choice.

**m3. `bpm` is not one of the ledger-silent keys; the set is sixteen and
`bpm` is graded at 0 on all twelve rows.** The spec lists seventeen keys
"empty or absent on all twelve" and includes `bpm` (spec:378-386).
Computed over the committed ledger, `bpm` is the integer 0 on all twelve,
which the oracle writes as `BPM=0` and reads back as 0, a real graded
value. Excluding `bpm`, sixteen keys are empty, null, or the empty list on
all twelve. So 53 of the 69 carry a graded non-empty value, not 52. The
mitigation does not change. Only the claim that the ledger proves nothing
about these keys, which is untrue of `bpm`.

**m4. The shadow reader's mapping carries eighteen rows, not nineteen.**
The spec says the provisional reader "covers nineteen fields
(`internal/engine/shadow/flac.go:208-229`)" (spec:57-61) and repeats "the
nineteen" twice more. The struct literal at `flac.go:209-228` has eighteen
rows. `matcher.ItemFields` has nineteen members
(`internal/engine/matcher/info.go:124-152`); `DataSource` has no row by
design, and the reader's own doc comment says so (`flac.go:203-207`).
HISTORY's forward carry 5 uses the same phrase, and the close is already
correcting that entry.

**m5. The last-block flag is never made the writer's responsibility.**
The container contract names the flag once, descriptively (spec:298-301).
The write path then says blocks are copied through by declared length in
their original order, and that a write emits a trailing PADDING block. A
writer that copies each preserved block's four-byte header verbatim leaves
a previously-last block's flag set, and a conformant decoder stops there.
The consequence half of this finding is rejected above, because the
block-chain assertion catches it loudly. The contract half stands: say
that the writer re-emits every block header rather than copying it, clears
the flag on all but the final block, and sets it on the block it writes
last, and that the parser rejects any block appearing after a header whose
flag is set.

**m6. Fork A reads as an engineering ruling escalated to the owner.**
Fork A asks what the content hash covers (spec:111-129). The surface
belongs to rung 6's schema and rung 7's gate, no adopter configures it,
and no 1.0 feature exposes it, because the drift check that reads it is a
2.0 hook. The spec's own argument makes the recommendation dominant rather
than a matter of taste, since a whole-file hash "goes stale the first time
rung 9's retag verb runs", and the declined half is recoverable later by a
separate tag-region digest (spec:637-639). CLAUDE.md reserves attended
time for taste, priorities, and product forks. Rule it as a conductor
ruling with the same recommendation and the same cost sentence, leaving
fork B as the one owner question.

**m7. Ruling 9 decides what an adopter may upload, with no consequence
stated and no reject path named.** Fork B is a fork "because the answer
decides which formats an adopter may upload at 1.0" (spec:131-137). Ruling
9 applies the same kind of decision at a finer grain: a FLAC with more than
one VORBIS_COMMENT block is refused, and rung 7's intake gate is bounded by
rung 8's support (`BS:633-636`), so such a file bounces at intake. Mutagen
and the deployed shadow reader both accept it. The spec offers no
prevalence evidence and says nothing about the reject copy, though it
records elsewhere that reject wording is contract and that per-format
wording is rung 7's (spec:518-520). Add one sentence naming the adopter
consequence and the wording rung 7 owes.

**m8. The staleness pre-task's `buildFLAC` question is false as written
and would report a blocker at every launch.** The question asks whether
`buildFLAC` is still used by `capture_test.go:34`, `capture_test.go:173`,
and `run_test.go:107`, "and only those" (spec:874-875). It is also used
four times inside its own file, at `flac_test.go:85`, `:202`, `:226`, and
`:288`. A fresh-context agent answering literally reports the mismatch,
and the first two staleness questions are declared blockers whose "no"
stops the launch. Scope the question to uses outside `flac_test.go`.

**m9. The absorption inventory misses the one comment that becomes false
when 8c lands.** The spec names three call sites and two prose mentions
(spec:665-668). It does not name the reader's own doc comment,
`internal/engine/shadow/flac.go:62-70`, which says "It is a provisional
reader. Rung 8 owns tag I/O and the full field surface in
`internal/engine/tagio` ... and rung 8's plan carries the task of
absorbing it." That comment travels with the reader and is wrong the
moment the absorption lands. Add it to 8c's repoint list, with the note
that the provisional framing and the rung 4 rationale are rewritten as
`tagio`'s own doc comment rather than carried over.

**m10. "No exported interface" is the wrong scope for the go-conventions
rule, and the registry's value type is unstated.** Ruling 14 says the seam
carries "one entry and no exported interface at rung 8" (spec:214-216),
repeated at spec:500. The rule it invokes is about existence rather than
export: go-conventions says "Unnecessary interfaces. Don't define an
interface until a second consumer exists." The prose two lines later
states it correctly (spec:501-504), so the document contradicts itself,
and as ruled an implementer may define an unexported single-implementation
interface and satisfy it. The registry's value shape is also unstated, and
the obvious reading of "a registry keyed by container that dispatches to a
tag-standard implementation" is a map whose value type is an interface,
which is the thing the ruling means to forbid.

**m11. Fork B's evidence omits that the ratified build spec already
answers it in the recommended direction.** Fork B is framed against F1's
"FLAC, MP3, and MP4/AAC at launch" (`S2:447`) and cites `BS:630-638` only
for the allowlist bound (spec:136-137). The passage it cites already
reads: "The allowlist is BOUNDED by rung 8's tag I/O support (pure-Go FLAC
at launch): the gate never admits a format the server cannot tag, so a
permissive adopter's wider allowlist arrives with post-launch tag I/O
extensions (ID3 and friends), not by loosening intake alone"
(`BS:633-636`). So the two ratified documents disagree, and the one the
fork does not quote agrees with CN's recommendation. Cite `BS:633-636`
inside fork B so the question reads as reconciling two ratified documents.

**m12. The blanket provisional marking sweeps in the two product forks,
and the counts disagree.** The section opens "Each is applied as written,
marked **provisional, conductor ruling owed**" (spec:160-162) over a list
of fifteen items whose items 8 and 15 are the two product forks
(spec:197, spec:217). Those are Geoff's to rule. The header separately
counts "Thirteen further rulings" (spec:9-11), so the opening sentence and
the list length disagree by two. Exempt the fork items in the opening
sentence.

**m13. The prior-art table omits chromaprint, and rung 8 never states that
it computes no fingerprint.** CLAUDE.md's build-for-2.0 rule names the
prior-art list a rung must adopt or decline with a reason, and `S2:421`
carries a chromaprint row reading "Adopt for 2.1 as a subprocess only.
Never link the library". The rung 8 table (spec:731-745) has no chromaprint
row. This matters beyond bookkeeping, because `acoustid_fingerprint` and
`acoustid_id` are two of the keys the writer must handle, and "What rung 8
does not build" states the parallel boundary for art but states none for
fingerprints. Add the row and the boundary.

**m14. 8c's exit gate is not checkable as written.** The gate requires
"the shadow's own seven tests passing unchanged in behavior" (spec:802).
The seven tests in `flac_test.go` are the reader's own, and the pass moves
the reader out, so they cannot pass unchanged in `shadow`. One of the
seven, `TestItemBagRoundTripsThroughProjectItemFields` (`flac_test.go:333`),
is an itembag projection test that must stay in `shadow` under ruling 3,
so the count also implies a disposition it does not intend. State the gate
as the shadow package's full suite green with the projection's behavior
unchanged, and name where each of the seven lands.

**m15. The bolded paragraph lead is used reflexively in three sections.**
61 paragraphs open with a bolded lead phrase across 8,775 words, 7.0 per
thousand against 5.3 in the ratified peer
(`specs/2026-09-05-rung-4-shadow.md`, 80 in 15,155). In "What the writer
guarantees" all eight paragraphs open that way (spec:327-376). In "The
write side of beets, as a contract" seven of eight do (spec:224-284). In
"The write path" all five do (spec:529-556). The writing-voice shape rule
names the reflexive bolded lead as the machine-list default. Sentence-level
register is clean, so this is the only voice exception. Keep the bold where
it names a contract a reviewer will check against, and drop it elsewhere.

---

## Verified sound

Both lenses' do-not-re-litigate records, merged. Every item below was
checked and holds.

### The encoding and container surface

- The writable set is exactly 69 keys and the read-only remainder exactly
  9. Recomputed against the pinned venv as
  `set(Item._media_fields) & set(MediaFile.sorted_fields())` and diffed
  against INV's 69-row table: identical sets. The nine read-only names
  match, and the list matches `corpus/README.md`'s handoff enumeration key
  for key.
- Every row of INV's 69-key table is right. The FLAC-applicable styles for
  all 69 fields were regenerated from mediafile 0.17.0 and compared: key
  names, declaration order, which styles are read-only, which are
  list-backed, `float_places` (2 by default, 6 on both ReplayGain peaks),
  and the `" dB"` suffix on exactly `REPLAYGAIN_TRACK_GAIN` and
  `REPLAYGAIN_ALBUM_GAIN` all match. The per-row line citations match
  `grep` of `__init__.py` exactly.
- A set writes every applicable non-read-only style and a read takes the
  first truthy one (`mediafile/fields.py:48-61`). `albumartist` writes
  three keys, `disctotal` and `tracktotal` three each, and `track`,
  `disc`, `comments`, `label`, `albumstatus`, and `encoder` two each, as
  stated.
- Deleting a field also deletes its read-only styles, because `read_only`
  blocks `set` and not `delete`
  (`mediafile/storage/base.py:56-58`, `:132-135`), so removing `barcode`
  removes `UPC`, `EAN/UPN`, `EAN`, and `UPN`.
- `MediaFile.update` deletes on `None` and sets on everything else,
  iterating `sorted_fields()` (`mediafile/__init__.py:365-378`,
  `:321-330`).
- Exactly three writable pairs collide on a Vorbis key, and the plural
  sorts last in every one. The full key-to-owner map over the 69 was
  computed; the only other multi-owner keys are `DATE`/`YEAR` and
  `ORIGINALDATE`, which the date rule already governs. The warning that a
  Go writer iterating a map is non-deterministic on those three keys is
  sound.
- Serialization: a bool becomes `"1"` or `"0"`, a float with a string
  target formats to `float_places` decimals, the suffix appends after
  formatting, and the R128 pair shifts by Q8, so `r128_track_gain = -1.5`
  writes `-384`.
- The date decomposition rule holds: `YYYY`, then `-MM` when the month is
  truthy, then `-DD` only when month and day are both truthy; `DATE`
  carries a `YEAR` fallback and `ORIGINALDATE` carries none; a `None` year
  deletes the field and its fallback together; a truthy year with zero
  month and day reads back as the year with `None` month and day. The
  three subjects with month 0 and day 0 confirm it against the ledger.
- Every container claim against RFC 9639: the block type table, the
  four-byte header with a big-endian 24-bit length against the block
  body's little-endian lengths, STREAMINFO present, first, once, and 34
  bytes, its MD5 over unencoded audio, at most one VORBIS_COMMENT, at most
  one SEEKTABLE, more than one PADDING and more than one PICTURE
  permitted, PICTURE big-endian throughout, and no framing bit in FLAC's
  Vorbis comment (`mutagen/flac.py:352-356` implements the last).
- The PICTURE behaviors: a store clears every picture and re-adds, delete
  clears all, the MIME type comes from sniffing via `filetype.guess_mime`,
  the type index is the ID3v2 APIC enumeration 0 through 20 with front
  cover at 3, an unset type is 0, an out-of-range index degrades to
  `other`, and front-cover selection takes the single candidate, then the
  first typed front, then the first of any type.
- The three-state significance and its scale: on ledger row 0 exactly ten
  keys are null and ten are the empty string, matching the named lists,
  and the type census over the twelve rows is 421 strings, 152 nulls, 123
  lists, 120 ints, and 12 booleans.
- The 8 KiB PADDING reservation is comfortably sized. A full 69-key write
  of subject `clean-match` into a scratch template copy produced a
  1844-byte VORBIS_COMMENT block over 67 comments, so "runs well past a
  kilobyte" is right and 8 KiB against mutagen's 1 KiB is defensible.

### The conformance procedure and the corpus

- All twelve subjects' `Expected.AppliedFields` cover all 69 writable keys,
  checked case by case, so step 2 reproduces what `derive_file_tags.py`
  wrote. The shim builds a fresh `Item`, sets those fields, and calls
  `Item.write`, which writes the full media-field dict; because the applied
  fields already carry beets' defaults for every unsupplied key, the two
  paths agree and nothing falls back to a default.
- The five conformance steps match `derive_file_tags.py`'s own order,
  including the per-subject scratch copy, the fresh `MediaFile` parse for
  the read-back rather than the object `write()` held, the template block
  read once from an unwritten copy, and the rule that the nine read-only
  keys are never a per-subject target.
- The twelve rows name twelve distinct case ids, so no subject shares a
  case with another.

### The spec's structure, scope, and citations

- The four owner rulings are recorded exactly as given (spec:86-104), each
  with its downstream consequence named. None is reopened or softened
  anywhere else in the document.
- The provisional-ruling numbering is INV's, item for item, with the
  product forks at 8 and 15, leaving thirteen conductor rulings. Each of
  the thirteen carries INV's own recommendation unchanged in substance.
- Fork B carries CN's recommendation and its erratum wording verbatim, and
  it carries CN's second owner question about whether the family has ever
  wanted to upload a non-FLAC.
- The 2.0 hook table is complete against the spec it cites. Hooks 1, 2, 3,
  11, 12, 13, and 14 are each given a position, and all seven line
  citations resolve exactly. Hooks 4 through 10 are rung 5, 6, and 12
  surfaces that rung 8 does not touch, so "every other hook neutral" holds.
- F1 is quoted and applied correctly. Declining `dhowden/tag` is what F1's
  "scoring `dhowden/tag` for reading" asks for rather than a reversal of a
  ruled item, and the decline carries checkable evidence.
- The territory and read-only rules match the reverse-edge document. The
  instrument lives under `internal/engine/tagio` with its own `testdata`,
  pins no schema version, and never calls `Reconcile`; rung 8 derives no
  new corpus subject; the unwired-`tagio` sentence is carried; and the
  divergence-home gap `RE:192` names is closed by ruling 2 on rung 4's
  schema-addition pattern.
- The absorb-and-repoint task names every code call site. A grep over
  `internal/engine/shadow` finds `readFLAC` referenced outside its own file
  and test only at `run.go:209`, `capture.go:158`, and `reconcile.go:35`'s
  doc comment, and `capture_test.go:47` and `:203` are the only prose
  mentions.
- Every tree fact the spec leans on checks out: `flac.go` is 313 lines;
  `flac_test.go` is 361 lines with seven test functions; `buildFLAC`'s
  external uses are the three named; `internal/atomicfile` has exactly one
  exported function taking `[]byte` at `atomicfile.go:15`;
  `arch_test.go:141` and `:152` are `TestEngineNeverImportsHarness` and
  `TestEngineNeverReadsTheClock` and both walk `internal/engine` including
  test files; `internal/engine/tag/doc.go:8-11` carries the split sentence
  ruling 3 preserves; and `internal/config/keys.go:57-140` carries reserved
  classes for rungs 5, 6, 7, 9, 12, and 13 and none for 8.
- The build-spec citations are exact: the ladder entry is quoted verbatim
  from `BS:694-701`, `BS:330` is the tagio tree line, `BS:348` is the
  tag-versus-tagio disambiguation, and `BS:630-638` is the
  allowlist-bounded-by-rung-8 passage. The rung 4 spec's "repointing is
  rung 8's task, never a silent adoption" sits at `:1591-1597`.
- Both record corrections the close applies are true. HISTORY's forward
  carry 5 does misplace the citation table, and `git ls-files '*.py'`
  returns eight files against CLAUDE.md's stated six.
- All twelve of INV's surprises land somewhere in the spec, as the header
  claims.
- The seam does not require an interface at rung 8, and reading rung 7's
  allowlist bound from the registry rather than restating it as a config
  default is a genuine second consumer, so the registry itself is not
  speculative.
- Prose register holds at the sentence level. `tellgrader` on the docs
  register reports zero tells, the file carries no em dash, and its
  tricolons are genuine enumerations of measured facts.
- Universality holds. No estate path, identity, or site-specific default
  appears; the rung declares no configuration key; and the one
  estate-shaped fact nearby is correctly assigned to rung 7's config
  bounded by rung 8's support.
- The staleness pre-task covers the live-contention risk this window
  carries: pass 4b editing comment lines under `internal/engine/shadow` and
  owning `scripts/check.sh` are both asked about by name, and the
  live-executor check is correctly stated as the lane registry rather than
  `pgrep`.
- The pass split's reasoning is well founded. The risk is a value model
  wrong on many keys at once, and grading belongs in its own pass: the
  write-every-style rule alone puts 20 keys behind three misreadings, and
  the ten-null-plus-ten-empty shape of row 0 is exactly the failure that
  looks small and is large.
- Product fork A's reasoning is sound even though its classification is
  not (m6). A tag write moves no audio sample and STREAMINFO's MD5 covers
  unencoded audio, so a whole-file hash goes stale on the first retag and a
  frame-region hash survives.
