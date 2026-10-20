# Adversarial review: the rung 8 pass 8a plan

Fresh-context review (2026-09-07) of
`plans/2026-09-07-rung-8a-tagio-writer.md` revision 1, in two
independent lenses: FLAC, mediafile and Go fidelity, verified encoding
by encoding against pinned mediafile 0.17.0 and mutagen 1.48.1 in
`tools/venv/.venv`, against INV's measured table, and against RFC 9639
as published; and plan integrity, verified against the rung 8 spec
revision 2, the spec inventory, the container-needs research, the 2.0
design's forward-hooks list, the 4a exemplar plan, `docs/HISTORY.md`'s
budget lines, and the tree at `47e80cc`. Folded as revision 2 the same
day; the merged 43-item "Verified sound" list below is the
do-not-re-litigate record for the pass, alongside the rung 8 spec's own
records and rung 4a's.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1 through B3, M-a through M-g, m1 through m11. Every
finding was re-verified by the fold editor against the tree, the venv or
the RFC before folding, and every numeric claim was re-measured rather
than carried. Five reviewer measurements did not survive that check in
the form they were written and are corrected below; none of the five
retires its finding, and no finding was rejected outright. Seven items
are conductor-owed rather than folded silently, and each is applied
provisionally in the plan with an inline marker.

Both lenses agree the plan is well grounded where it reasons from the
inventory. Every serialization rule resolves to the pinned library line
it names, the four-state model and its scale evidence are exact, the
write-every-style and last-style rules are correct including the
non-obvious reason the last style wins, the three colliding pairs and
the plural-wins resolution reproduce `sorted_fields()`, the four 2.0
hook statements are right against the streamer spec's own text, every
prior-art verdict reproduces the spec's table rather than inventing one,
the independence marks are true by files, and the test-expectation chain
has no self-derivation at its root. The prose measures clean: tellgrader
reports 0 tells over 14,444 words at a cadence CV of 0.85, and the file
carries no em dash.

The defects concentrate in four places. The golden's own input is never
written down, so three tasks assert against a citation that holds the
output. The field-name rule is stated in a task that has no input path
for it. The parse's memory bound is never stated, so Task 3 can nullify
Task 1's whole justification. And the ceiling arithmetic divides by the
wrong denominator, which schedules its own stop rule to fire.

---

## Corrections to the reviews' own evidence

**1. Review 1's M-a says three of the sixteen ledger-silent keys appear
in the golden's comment set.** Re-measured against `INV:409-427`, it is
four: `ACOUSTID_FINGERPRINT=` is in the set alongside `R128_TRACK_GAIN`,
`REPLAYGAIN_TRACK_GAIN` and `REPLAYGAIN_TRACK_PEAK`. Twelve of the
sixteen are graded by hard-coded expectations alone, not thirteen. The
finding stands with the corrected count.

**2. Review 1's M-e cites `INV:430` for the `barcode=None` sentence.**
It is `INV:429`. The finding stands; only the pointer moves.

**3. Review 1 gives two mutagen spans one line short at the start.**
`is_valid_key` is `mutagen/_vorbis.py:25-41`, not `:24-41`;
`VCommentDict.__setitem__` is `:275-297`, not `:276-297`. The spec's own
citation at `spec:362` already reads `:275-297`. Both findings that rest
on them stand, and the 0x20 to 0x7D bound is verified verbatim at
`_vorbis.py:29-30` and `:39`.

**4. Review 2's M-g says "roughly a third of about thirty-five sampled"
spec citations resolve to the wrong subject.** The fold editor ran the
sweep mechanically over all 96 `spec:`, `INV:`, `S2:` and `CN:` pointers
in the plan rather than a sample. Seventeen land on a different subject
and are listed in M-g; about a dozen more are off by one to three lines
and still open on their subject, which is not a finding. The rate claim
is directionally right and the enumerated list is now complete, which is
what the fold needs.

**5. Review 1's M-d says thirteen unlisted rows are list-valued.**
`INV:315-385` carries fourteen rows marked "one comment per value,
repeated under the writable key". Thirteen of them are absent from the
plan's exceptional-rows table; the fourteenth, `albumartists`, is
already listed there under its own shape. The finding stands exactly as
written, and the measurement is now on the record.

---

## BLOCKERS

### B1. The golden's input is recorded nowhere, "fifteen fields" cannot be reconstructed, and three assertions rest on a citation that holds the output

**Lens.** FLAC and mediafile fidelity.

**Plan lines.** Resolution R2 (`plan:638`), Task 2's first outcome
(`plan:994`), and the golden section's second bullet (`plan:571`), all
citing `INV:409-427` for the write's input. Task 7's key-by-key
comment-set criterion rests on the same citation, and so does the spec's
own 8a exit gate (`spec:1102`).

**Evidence.** `INV:409-427` is the fenced block holding the emitted
comment set, verified by line: `INV:409` and `INV:427` are the fences.
`INV:407` says only "from a fifteen-field `MediaFile.update`", and
`INV:429` adds "`barcode=None` in the same call", which makes the call
at least sixteen keys. The input dict itself appears nowhere in INV.
Reconstructed from the output, the distinct beets fields are at least
twenty: `acoustid_fingerprint`, `album`, `albumartists`, `albumtypes`,
`artist`, `barcode`, `bpm`, `comp`, `disc`, `disctotal`,
`original_day`, `original_month`, `original_year`, `r128_track_gain`,
`rg_track_gain`, `rg_track_peak`, `title`, `track`, `tracktotal`, and
`year`. The output also cannot say whether `albumtype` was set alongside
`albumtypes`, because the plural wins either way
(`mediafile/__init__.py:562-566` with `sorted_fields` at `:321-330`),
and that is exactly the fact deciding whether the golden anchors the
colliding-pair rule at all.

**Consequence, built as written.** Task 2's implementer is told to
reproduce "the fifteen fields INV:409-427 measured" and no such list
exists. Whatever it invents becomes the golden, and Task 7's key-by-key
assertion then grades the writer against a comment set the plan
believes it knows and does not. A wrong reconstruction is invisible: the
golden and the assertion agree with each other because one produced the
other's expectation.

**Fix folded.** Drop the "fifteen" count everywhere it appears. Task 2's
provenance record already requires "the input field by field"; that
record becomes the contract, and Task 7's key-by-key assertion reads
`golden.txt` rather than `INV:409-427`. `INV:409-427` stays cited as the
shape the input must at least cover, since the emitted set is the one
published measurement. The spec's exit-gate sentence inherits the same
correction through Task 9's drafted corrections.

### B2. The field-name rule sits in a task with no input path, the spec's fifth refusal is graded by nobody, and no task states the read-side behavior

**Lenses.** Both, independently. Review 2 found the missing input path;
review 1 found the missing read-side rule and the divergent mutagen
behaviors. Both halves are one defect and fold together.

**Plan lines.** Task 5's criterion (`plan:1328-1329`) and its mutation 9
(`plan:1351-1353`); C7's refusal list (`plan:409-419`); Task 3's refusal
outcomes (`plan:1084-1090`); Task 8's owed deviation note
(`plan:1618-1623`).

**Evidence.** `spec:555-561` lists five refusal shapes and requires "one
table-driven test per refusal", the fifth being "a field name outside
RFC 9639 section 8.6's set". C7 and Task 3 carry only the four
structural shapes plus the `fLaC` signature, so no task grades the
fifth. Task 5 carries it instead, and C1 (`plan:353-357`) fixes the tag
set at 69 fields with Go-native types, so every name the encoder emits
is a package constant and no caller can hand it an invalid one. No
task's Interfaces section produces a surface that takes a field name.
The only path by which an invalid name can reach `tagio` is a comment
read from a file, and nothing in the plan validates that.

The two possible read-side behaviors diverge sharply, and the reference
takes both: mutagen's `VComment.load` skips an invalid key silently
(`_vorbis.py:127-128`) while `validate` raises on write (`:150-151`).
RFC 9639 section 8.6 is verified as published: "The field name MUST only
consist of UTF-8 code points U+0020 through U+007E, excluding U+003D".
Under C8 the writer preserves every foreign comment, so a preserved
comment whose name carries a code point outside that range makes a file
either wholly refused or permanently un-retaggable, depending on which
reading the implementer picks, and the plan does not pick.

**Fix folded, provisional, conductor ruling owed.** The bound moves to
the parse, where the spec's other four refusals already live. Task 3
gains a refusal row: a comment name carrying a code point outside
U+0020 through U+007E returns the malformed sentinel, because `tagio`
cannot round-trip a name it may not re-emit, and a file it cannot
round-trip is malformed rather than merely odd. `=` needs no row, since
the parser splits a comment on its first `=` and a parsed name can never
contain one. Task 3 also gains the acceptance row that carries the
deviation: a comment named with U+007E parses and is preserved, which
mutagen's loader would have skipped. Task 5's criterion and its mutation
9 are deleted, and Task 8's deviation note repoints at Task 3. No
write-side validation is added: with C1's constants and Task 3's bound,
an invalid name is unreachable at the encoder, and `go-conventions`
forbids a guard against an impossible state.

### B3. The ceiling derivation divides by ten where the measurement is over nine, and the corrected expectation sits above the ceiling the plan then keeps

**Lens.** Plan integrity.

**Plan lines.** The basis paragraph (`plan:187-188`), the nine-task line
(`plan:197`), the total (`plan:203`), and the 80% flag (`plan:214`).

**Evidence.** The plan reads "4a's measured rate of about 220K per task"
from "4a-i ... chain 2.14M over ten tasks". `INV:1074` scopes 4a-i to
"Tasks 1 to 9", and `docs/HISTORY.md:745-746` lines Task 0 out
separately: "prep 2.3M, Task 0 0.07M, chain 2.14M". The measured chain
rate is 2.14M over nine tasks, about 238K, and the plan both divides by
ten and carries Task 0 again as its own 0.1M line. 8a has nine
implementer tasks, the same count, so the direct comparable is 2.14M
rather than 2.0M. The table then reads
0.8 + 0.1 + 2.14 + 0.2 + 0.4 + 0.2 + 0.25 + 0.2 = **4.29M**, against a
4M ceiling.

**Consequence, kept as written.** Prep's 0.8M is spent the moment the
plan is committed, so the 80% flag at 3.2M sits 2.4M above first
dispatch against about 3.4M of remaining forecast. The flag fires around
Task 7 by the plan's own arithmetic. `spec:1157-1160` raised the
proposed ceiling from 3M to 4M precisely so the conductor is not driven
to the stop line mid-execution, and a stop rule scheduled to fire is not
a signal.

**Fix folded, provisional, conductor ruling owed.** The divisor is
corrected and the expectation restated at 4.29M. The ceiling is proposed
at **4.5M**, on the spec's own re-derivation clause
(`spec:1163-1164`: each ceiling is re-derived from the previous pass's
actual report before the next plan is drafted). `spec:1150` reserves 8M
for the rung, so the change is recorded against that reservation: the
rung total stands at 8.5M until 8b closes, at which point 8c's 1M line
is the natural giveback, 8c being a scoped absorption that moves one
reader, repoints three call sites and deletes `buildFLAC`. The 80% flag
moves to 3.6M, and the plan now states what the conductor does when it
fires: take the Task 3 re-derivation that is already scheduled, rather
than treat a forecast the plan made as a surprise.

---

## MAJORS

### M-a. Nothing byte-anchors twelve of the sixteen ledger-silent keys, or any corpus-unreachable case

**Lens.** FLAC and mediafile fidelity.

**Evidence.** Risk 2b (`spec:1230-1237`) is the reason the golden
exists: a defect the conformance run cannot see. Of the sixteen keys
Task 8 grades, four appear in the golden's comment set
(`acoustid_fingerprint`, `r128_track_gain`, `rg_track_gain`,
`rg_track_peak`; re-measured, see correction 1). The remaining twelve,
plus the zero year, the empty list, `catalognum` and `barcode` empty,
are graded only by hard-coded expectations a human read out of Python
source. That is the same reading that produced the plan, so a misreading
is invisible to the pass. The golden already carries an input; widening
it costs no second fixture, which is R2's whole argument.

**Fix folded.** The golden's input widens to cover every ledger-silent
key and every corpus-unreachable case a mediafile write can express in
one file. Two cases cannot go in and stay out by construction: mediafile
writes values that are valid UTF-8 by construction, and mutagen's
`is_valid_key` refuses a name above 0x7D, so the invalid-UTF-8 sequence
and the U+007E name remain Task 8's hand-built cases. Task 2's
provenance record names what it could not express and why. Task 7's
key-by-key assertion points at that record.

### M-b. No task bounds the parse's memory, so Task 3 can nullify Task 1's whole justification

**Lens.** FLAC and Go fidelity.

**Evidence.** Task 1's first outcome is that content never buffers, so a
multi-gigabyte rewrite costs a fixed amount of memory (`plan:893-897`,
citing `INV:906-916`), and that is ruling 11's entire justification.
Task 3 names `internal/engine/shadow/flac.go` as the nucleus to widen,
and that nucleus reads the whole file: `data, err := os.ReadFile(path)`
at `internal/engine/shadow/flac.go:78`. Nothing in Task 3's outcomes or
criteria says the parse reads only the metadata region, and Task 7's
outcome says the frame region is "copied byte for byte" without saying
from where. Task 1's mutation 5 makes the streaming property a
reviewer's read, and that reviewer reads Task 1's diff, not Task 3's.

**Fix folded.** Task 3 gains an outcome and a reviewer criterion: the
parse reads only the metadata region and carries the frame region as an
offset and a length, never its bytes, bounds-checked against the file's
size. Task 7 gains the matching criterion: the frame region reaches the
temp file through a copy from the source at the recorded offset. "No
whole-file read on the write path" joins Task 7's named mutations as a
reviewer-read check, so both halves are graded where they are built.

### M-c. C8's case-insensitive replacement is stated as behavior and graded by no criterion

**Lens.** FLAC and mediafile fidelity.

**Evidence.** C8 (`plan:421-431`) says key names match case-insensitively
per RFC 9639 section 8.6, and the RFC confirms it: "The evaluation of
the field names MUST be case insensitive". The mechanism is
`VCommentDict.__setitem__` (`mutagen/_vorbis.py:275-297`), whose
`del self[key]` matches case-insensitively before appending. Task 7's
criteria assert foreign-key preservation and the emitted comment set,
but no row writes an owned key into a container that already holds it in
a different case. A writer that appends without the case-insensitive
delete leaves two `TITLE` comments in the file, and every `tagio`
read-back still passes, because the reader takes the first truthy style.
That is a wrong tag shipped by a fully green pass.

**Fix folded.** Task 7 gains a criterion: a fixture whose comment block
carries `title=old` in lowercase is replaced rather than duplicated by a
write of the title field, and the resulting comment sequence carries
exactly one title comment. The matching named mutation is skipping the
case-insensitive delete.

### M-d. The exceptional-rows lead sentence mis-describes thirteen list-valued rows

**Lens.** FLAC and mediafile fidelity.

**Evidence.** The plan says "Every row not listed below is a single
writable key taking one comment, read from `INV:315-385`". Re-measured,
`INV:315-385` marks fourteen rows "one comment per value, repeated under
the writable key", thirteen of which are unlisted: `albumartists_credit`,
`albumartists_sort`, `albumtypes`, `arrangers`, `artists`,
`artists_credit`, `artists_sort`, `composers`, `genres`, `lyricists`,
`mb_albumartistids`, `mb_artistids` and `remixers`. Whether a field is
list-valued decides its Go type and therefore which of C2's four states
it can hold: a string field's empty value writes an empty comment while
a list field's empty value writes nothing
(`ListStorageStyle.set_list`, `mediafile/storage/base.py:189-199`). The
plan itself calls confusing the two the twenty-keys-per-row defect.

**Fix folded.** The lead sentence now says each unlisted row takes its
write shape from INV's "Write shape" column, and a row naming the
thirteen list-valued fields joins the table, since the Go type decision
hangs on it.

### M-e. The date rules are specified only for a fully-populated triple

**Lens.** FLAC and mediafile fidelity.

**Evidence.** The plan's serialization list and Task 5's date table cover
all components present, a truthy month with a zero day, a zero month
with a truthy day, every component zero, and an absent year. C2's
four-state model lets year, month and day each be absent independently.
mediafile's `DateItemField.__set__` (`mediafile/fields.py:225-228`) is a
read-modify-write over the file's current `DATE`, and
`MediaFile.sorted_fields` orders them `date0`/`date1`/`date2`
(`__init__.py:305-330`), so `update({"month": 3})` against a file
holding `DATE=2001-01-01` yields `DATE=2001-03-01`. The plan states no
rule for a partially-populated triple, Task 5 has no row for it, and
Task 4's write-order criterion asserts only the three colliding pairs.
An implementer must invent the rule and the `diff-reviewer` has no
criterion to grade it against.

**Fix folded, provisional, conductor ruling owed.** The triple composes
as a unit. `tagio`'s write entry point takes a whole tag set rather than
a partial update, so it never read-modify-writes a file's current
`DATE`: an absent year deletes `DATE` and its `YEAR` fallback, and an
absent month or day with a present year is treated as falsy. That
reproduces mediafile's end state for every fully-populated case and
diverges from it only for a partial update `tagio` has no surface for,
which Task 8's deviation note records. Task 5 gains a row for a present
year with an absent month.

### M-f. Task 0 drops the spec's handoff check and demotes the ledger check, leaving Task 4's 69-name list with two sources and no rule

**Lens.** Plan integrity.

**Evidence.** `spec:1175` makes the first two checks blockers, and the
spec's second is `corpus/analysis/file_tags.json` unchanged
(`spec:1181-1184`). The plan makes its own item 2, `internal/atomicfile`'s
shape, the second blocker (`plan:802-811`) and files the ledger check as
item 11, "a finding that 8b's plan needs rather than a blocker here",
instructing it to count mechanically (`plan:851-856`). The spec's last
check, `corpus/README.md`'s handoff section and its 69-key list
(`spec:1211-1212`), is dropped from Task 0 entirely. That handoff is
load-bearing here: C1 defines the tag set as the intersection "matching
`corpus/README.md`'s handoff key for key" (`plan:353-357`), and Task 4's
central mechanical assertion compares "the count and the names against a
hard-coded list taken from the corpus handoff" (`plan:1209-1212`). The
handoff exists and holds the list, verified at
`corpus/README.md:2354` ("Handoff to rung 8: what pure-Go FLAC tag
writing may assume") with the 69 names reproduced from the
`writable_fields` header at `:2381` onward. It is not in the plan's
primary inputs, so the implementer's only listed source for those names
is `INV:315-385`. Two sources, one hard-coded list, no rule for a
disagreement, and no staleness check on either one's names.

**Fix folded.** The handoff check is restored to Task 0 as a blocker,
because it is 8a's source of record for the 69 names. The ledger-header
check stays as its own item with the reason stated in one sentence: 8a
reads the handoff's name list rather than the ledger's rows, so a
changed header is 8b's blocker, and the item now cross-checks that the
handoff list and the header's count agree, which is where the two would
drift. Task 4 names `corpus/README.md`'s handoff as the source of record
and INV's table as the per-row mapping source, with the handoff winning
on a name disagreement, and `corpus/README.md` joins the primary inputs.

### M-g. Seventeen spec and INV pointers resolve to a different subject, against the plan's own verify-what-you-carry rule

**Lens.** Plan integrity.

**Evidence.** The fold editor resolved all 96 `spec:`, `INV:`, `S2:` and
`CN:` pointers in the plan mechanically at `47e80cc`. Seventeen open on
a different subject:

| Plan line | Cited | Actual | Subject |
| --- | --- | --- | --- |
| 33, 972 | `spec:1263-1272` | `spec:1230-1237` | Risk 2b |
| 161 | `spec:1003` | `spec:1009` | Hook 14's row |
| 226 | `spec:1148` | `spec:1150` | The 8M rung total |
| 233 | `spec:1097` | `spec:1102` | 8a's pass-split row |
| 374 | `spec:389-401` | `spec:401-410` | The three colliding pairs |
| 401, 1373 | `spec:694-703` | `spec:698-707` | The read projection |
| 411 | `spec:672-687` | `spec:686-692` | The container type |
| 454 | `spec:527-531` | `spec:538-541` | Audio untouched, MD5 never recomputed |
| 462 | `spec:533-545` | `spec:543-553` | The atomic write path |
| 955 | `spec:1134` | `spec:1132` | `internal/atomicfile`'s consumers |
| 961, 1573 | `spec:1024` | `spec:1028` | `go-flac` v2's verdict |
| 1082 | `spec:678-682` | `spec:688-690` | VORBIS_COMMENT as a sequence |
| 1100 | `spec:1022` | `spec:1030` | `dhowden/tag`'s issue #116 |
| 1158, 1459 | `spec:1020` | `spec:1026` | Adopt the repo's own reader |
| 1251 | `spec:1004` | `spec:1008` | Hook 13's row |
| 1259 | `spec:1041` | `spec:1035` | Picard's tables |
| 1569 | `spec:1006` | `spec:1014` | Hook 12's row |
| 1576 | `spec:1026` | `spec:1029` | `flacvorbis`'s verdict |

The four `S2:` hook pointers are exact, and about a dozen further
pointers are off by one to three lines while still opening on their
subject, which is not a finding. The plan instructs its readers to "cite
it, and verify a line you carry forward" (`plan:74-75`) and charges Task
0 with treating a moved citation as a finding (`plan:794-796`), so this
rate either burns Task 0's budget or passes the errors through to nine
implementers.

**Fix folded.** All seventeen are corrected in place. The plan's
verification stamp moves from `8031e3b` to `47e80cc`, and Task 0's
citation item now says the sweep was run mechanically at that HEAD, so
the task re-runs it rather than rediscovering it.

---

## MINORS

**m1. Task 1's mutation 3 predicts an outcome the landed code
contradicts.** The mutation reads "Remove the temp file on success
instead of on failure -> the overwrite row fails, because the rename
target is gone." `WriteFile` already removes the temp file on success:
`defer func() { _ = os.Remove(tmpPath) }()` at
`internal/atomicfile/atomicfile.go:22` runs after the rename at `:38`,
where it is a harmless no-op. The mutation as worded changes nothing
observable. Folded as "remove the temp file before the rename", which is
the edit that breaks the overwrite row.

**m2. Task 6's mutation 5 is unfired by construction, and the plan can
say so from evidence rather than asking the task to discover it.**
`safe_cast` to float matches the leading numeric prefix,
`[+-]?([0-9]+\.?[0-9]*|[0-9]*\.[0-9]+)`, at
`mediafile/utils/type_conversion.py:52`, so `"-7.50 dB"` casts to -7.5
whether or not the suffix was stripped first. `" dB"` is the only suffix
on any FLAC-applicable style (`__init__.py:836` and `:846` for the
gains; the peaks at `:856` and `:864` carry `float_places=6` and no
suffix), so the ordering is unobservable across all 69 keys. The plan
already hedges, and its stated outcome ("the rows fail, reporting 0")
contradicts the reason it gives in the same sentence. Folded as unfired
by construction with the citation, keeping one round of task budget.

**m3. C7 cites RFC 9639 section 8.5 for a refusal it does not carry.**
C7 says sections 8.1, 8.2, 8.5 and 8.6 make each refusal a spec
violation, and none of the four derives from 8.5. Section 8.5 caps
SEEKTABLE at one per stream (`INV:779`), and C10 has the writer preserve
unknown blocks unconditionally, so a container with two SEEKTABLE blocks
is re-emitted with both. Folded provisionally as: drop the 8.5 citation,
and state the copy-through as a deliberate non-refusal in Task 8's
deviation note. Refusing it would reject files other players read, and
rulings 7 and 10 already commit to copying unknown blocks through.

**m4. Task 7's byte assertion does not say whether the block header is
compared.** The criterion reads "The emitted VORBIS_COMMENT block's
bytes equal the golden's". If the parser returns payload only, the
four-byte header falls outside the comparison, and the plan's own C11
calls that big-endian header beside the little-endian body "the
asymmetry that catches implementers". Folded as "the header and the
payload", which pins the header's length encoding and its cleared
last-block flag in the same assertion.

**m5. `buildContainer` re-creates the shape ruling 4 deletes, against
the plan's own rule.** The test-helpers section says no helper builds a
container the writer also builds. Task 7 then requires a round-trip
fixture carrying one block of every type plus one reserved type, which
neither the writer nor the golden can produce. The exception is
defensible, because Task 3's golden parse independently anchors the
header layout, but the plan reads as contradicting itself. Folded by
stating the exception where the rule is stated.

**m6. Two citation drifts a fold should correct.** R6 cites
`internal/engine/shadow/flac.go:74` for the reader taking no `ctx`;
`readFLAC`'s signature is at `:77`, the doc comment running to `:76`.
The `.gitleaksignore` instruction says "nothing is inserted above line
66", and the file is a single line whose fingerprint references
`internal/config/keys_test.go:generic-api-key:66`, so the phrasing reads
as a rule about the ignore file's own numbering. Both folded, the second
as "append the new fingerprint below the existing line", which is what
`.claude/agents/dubplate-implementer.md` already says.

**m7. R7 crosses `spec:1103`'s 8b row without naming it.** `spec:1103`
scopes "the seam registry with its supported set, its func-value
dispatch, and its unsupported-container sentinel" to 8b. R7 lands both
sentinels in 8a and argues it from 8b's registry adopting the first
unchanged, citing only C14. The plan's own scope statement
(`plan:20-22`) forbids a task that believes it needs the seam registry,
so a reader cannot tell whether R7 is a considered crossing or an
oversight. The fold R7 takes is right; folded by naming what it folds
against, in one sentence citing `spec:1103`.

**m8. Two of Task 2's criteria are over a committed binary under a
header promising diff verification.** The section is headed "Acceptance
criteria a diff-reviewer verifies from the diff", and two entries grade
`golden.flac`'s block chain and vendor string, which a reviewer reading
a diff cannot check without a decoder. Task 3's first criterion does
close it one task later. Folded by stating in Task 2 that the reviewer
grades the provenance record's internal consistency and Task 3's parse
assertion proves the record against the bytes, which also tells the
conductor a Task 2 accept is provisional until Task 3 lands.

**m9. Where a freshly created VORBIS_COMMENT block sits in the re-emitted
chain is unresolved.** C9 gives the writer a fresh vendor and C11 gives
it the header re-emission, but no contract, criterion or resolution says
where a newly created block goes. R4 makes every byte assertion a write
into a copy of the golden, which already carries a block, and Task 7's
only fresh-block criterion is the vendor prefix. Folded provisionally as
a contract item: a freshly created VORBIS_COMMENT block is written
immediately after STREAMINFO, which is where the oracle places it
(`spec:531-533`: a mediafile save into the 42-byte template produced
STREAMINFO, VORBIS_COMMENT, and one PADDING block). Task 2's provenance
record is the evidence, and a record that contradicts it is a Task 9
correction.

**m10. The primary-input delegation to 4a's Global Constraints is
inaccurate and inherits stale rules.** `plan:84-85` says this plan does
not restate 4a's constraints. Its own Global Constraints do restate most
of them. What 4a carries and 8a does not includes constraints now false
for this pass, among them "`internal/config` is rung 4's for the length
of this pass" and the shadow's `import.lock` rule, both verified present
in 4a's list. Folded by dropping the delegation and letting 8a's
constraints stand alone; the 4a plan stays a primary input for its Task
4 shape and its STREAMINFO offset table, which is why it is there.

**m11. The bolded-lead-phrase bullet is at nearly three times the
exemplar's rate.** Measured, 94 of 199 top-level bullets open with a
bolded lead phrase (47%), against 55 of 307 (18%) in the ratified
exemplar `plans/2026-09-05-rung-4a-shadow-runner.md`. The writing-voice
router's document-shape rule reads "Do not open every bullet or
paragraph with a bolded lead phrase. Sparingly it signposts; by reflex
it is the machine list default." The rest of the prose is clean:
tellgrader reports 0 tells over 14,444 words at a cadence CV of 0.85,
and the file carries no em dash. Folded by unbolding the lead phrases in
the per-task Outcomes, Acceptance-criteria and mutation lists, keeping
the bold in Global Constraints, the package contract and the
Resolutions, where the lead phrase is the item's name.

---

## Rejected findings

None. Every finding in both reviews survived re-verification against the
tree, the venv or the RFC. Five measurements were corrected without
retiring their findings, and those corrections are recorded above.

---

## Verified sound

Do not re-litigate these. Each was checked by the fold editor against
the named artifact.

1. The float-with-string-target rule: formats to `float_places` decimals
   and a suffix appends after formatting
   (`mediafile/storage/base.py:109-130`).
2. A bool becomes `"1"` or `"0"` (`storage/base.py:117-119`).
3. `" dB"` is the only suffix on any FLAC-applicable style, on
   `rg_track_gain` (`__init__.py:836`) and `rg_album_gain` (`:846`);
   both peaks carry `float_places=6` and no suffix (`:856`, `:864`).
4. The R128 pair shifts by 2^8 (`QNumberField`, `fields.py:272-293`,
   `fraction_bits=8` at `__init__.py:874` and `:881`), so
   `r128_track_gain=-1.5` writes -384.
5. The four-state model (C2). `MediaFile.update` deletes on `None` and
   sets on everything else (`__init__.py:365-378`). An empty list
   reaches `ListStorageStyle.set_list`, which stores `[]` and emits no
   comment (`storage/base.py:189-199`); an empty string emits an
   empty-valued comment. The plan's insistence that the Go type carries
   the presence bit rather than inferring it from a zero value is
   correct and load-bearing.
6. The write-every-style rule (C4). `MediaField.__set__` writes every
   applicable non-read-only style (`fields.py:56-61`) and
   `MediaField.__get__` takes the first truthy one (`:48-54`).
   `albumartist`'s three writable FLAC styles at `__init__.py:542-544`;
   `albumartists`' two writable plus read-only `ALBUM ARTISTS` at
   `:554-556`; `disctotal` and `tracktotal`'s three each per INV's table.
7. The last-style rule (C5). `MediaField.__get__` overwrites `out` on
   every style and breaks only on a truthy value, so the last applicable
   style's fetch survives when nothing is truthy (`fields.py:48-54`).
   `catalognum`'s trailing `CATALOGID` and `DISCOGS_CATALOG` are
   read-only (`__init__.py:604-606`) and `barcode`'s `UPC`, `EAN/UPN`,
   `EAN` and `UPN` are read-only (`:616-620`), so both read back absent
   for an empty written value while `asin` and `albumartist` read back
   `""`.
8. `read_only` blocks `set` and not `delete`
   (`storage/base.py:56-58`, `:60-65`, `:132-135`), so Task 5's
   "deleting `barcode` removes all five keys" is correct.
9. The three colliding pairs and the plural-wins resolution.
   `sorted_fields` is lexicographic with `DateItemField` names rewritten
   to `date0`/`date1`/`date2` (`__init__.py:305-330`), and `update`
   iterates it, so `albumtypes`, `mb_artistids` and `mb_albumartistids`
   each land after their singular partner. C3's insistence that the
   write order is contract rather than implementation detail, and Task
   4's criterion asserting each pair member's position relative to its
   partner, are the correct proof shape for a Go map's randomized
   iteration.
10. `safe_cast`'s read behavior (`utils/type_conversion.py:4-60`):
    `None` returns `None` for every type; `int` takes the leading
    `[+-]?[0-9]+` (`:26`) and yields 0 otherwise; `float` takes the
    leading numeric prefix (`:52`) and yields 0.0; `bool` is
    `bool(int(val))` falling to False on `ValueError`. The plan
    reproduces all four, including "an absent key reads back absent
    rather than a typed zero".
11. The U+007E deviation is real, one code point wide. RFC 9639 section
    8.6 as published requires U+0020 through U+007E excluding U+003D,
    and mutagen's `is_valid_key` stops at 0x7D (`_vorbis.py:25-41`).
    `VComment.load` skips an invalid key silently (`:127-128`) while
    `validate` raises on write (`:150-151`), exactly as the spec states.
12. The container layout: a four-byte header with a last-block flag, a
    seven-bit type and a three-byte big-endian length, against a
    little-endian VORBIS_COMMENT body, verified against RFC 9639
    sections 8.1 and 8.6 through `INV:723-760` and against mutagen's own
    write. C11's naming of that asymmetry, and Task 3's mutations 1 and
    2 for each half, are the right pair.
13. The STREAMINFO offset table Task 3 builds to. The 4a plan's
    corrected table matches RFC 9639 section 8.2 exactly: four size
    fields in `[0:10]`, the 64-bit pack in `[10:18]`, the MD5 in
    `[18:34]`. Task 3's mutation 3 reads inside the MD5, so the
    predicted failure is real.
14. The container invariants. STREAMINFO must be first and there must be
    no more than one; PADDING may appear more than once, which is what
    makes the drop-all-and-write-one policy of rulings 5 and 6 necessary
    rather than optional (`INV:736-780`).
15. The crash-safety of the write path. C13's sequence has no window in
    which the target holds partial content, and the source stays intact
    until the rename. R10's decision to leave `WriteFile` alone is
    correct: it already takes and applies a perm
    (`internal/atomicfile/atomicfile.go:35`), so no rung 4 call site
    changes behavior, and the new function is a sibling rather than an
    edit.
16. No interface is introduced before a second implementation. Ruling
    14's registry is 8b's, Task 1's new function takes a func value
    rather than an interface, and R5 declines to ship a picture
    parameter 8a cannot fill. The plan carries no Go code at all, so it
    specifies outcomes throughout.
17. Territory and reverse edges. R1's generator copies
    `internal/harness/oracle/testdata/template.flac` (verified 42 bytes)
    to a scratch path and writes only into
    `internal/engine/tagio/testdata`, so cells (8,3) and (8,4) stay
    read-only. R3's "no Go test opens a path outside
    `internal/engine/tagio`" holds.
18. The architectural walks are correctly cited:
    `TestEngineNeverImportsHarness` at
    `internal/harness/conform/arch_test.go:141` and
    `TestEngineNeverReadsTheClock` at `:152`, both skipping only
    `testdata` subtrees, so `internal/engine/tagio` enters both from
    creation.
19. Task 0's checklist expectations hold at `47e80cc`: `atomicfile.go`
    carries one exported function at `:15` in a 42-line file;
    `internal/engine/tag/doc.go:8-11` still says `tagio` knows nothing
    about matching; `corpus/analysis/file_tags.json` still carries
    `shim_version` 1, `beets_version` 2.13.1, 69 `writable_fields`, 9
    `readonly_fields` and 12 rows; `.gitleaksignore` is a single line
    pinning `internal/config/keys_test.go:generic-api-key:66`;
    `git ls-files '*.py'` returns eight, so Task 2's generator is the
    ninth as the plan says; `internal/engine/shadow/flac.go` is 313
    lines.
20. The 2.0 hook accounting, verified against the streamer spec's own
    text with all four `S2:` pointers exact. Hook 13's sort tags
    (`S2:388-389`) are all in the writable 69, so Task 4 lands it; hook
    11 (`S2:383-384`) is enabled by the ReplayGain and R128
    serializations; hook 12 (`S2:385-387`) is honored by writing the
    applied value verbatim; hook 1 (`S2:359-360`) is constrained by
    ruling 8's frame-region extent, which Task 3 exposes; hook 14 is
    recorded as a decline rather than an omission, matching owner ruling
    4 (`spec:131-137`).
21. R4's byte assertion is sound, for a non-obvious reason worth
    keeping: writing the same field set into a copy of the golden
    reproduces the golden's comment order, because mutagen's
    `__setitem__` deletes every case-insensitive match and appends
    (`_vorbis.py:275-297`), so a second pass over the same keys in the
    same order lands them in the same sequence. Writing into a copy of
    the 42-byte template would change both the vendor and the block
    length, as the plan says.
22. The test-expectation chain has no self-derivation at its root. The
    golden's comment set comes from the oracle, Task 3's parse is graded
    against the golden rather than against `buildContainer`, and Task
    6's 69 projected values are graded against the provenance record.
    Task 7's block extraction uses the parser under test, but the parser
    is independently anchored by the golden parse, so the chain does not
    close on itself.
23. The ceiling table's arithmetic is internally correct as printed:
    0.8 + 0.1 + 2.0 + 0.2 + 0.4 + 0.2 + 0.25 + 0.2 = 4.15M. Only the
    2.0M line's derivation is wrong (B3).
24. Every basis figure the ceiling section quotes matches INV's budget
    table: prep 0.25M to 2.55M across seven passes, chain 1.3M to 2.2M
    for a nine-to-fifteen-task pass, simplify about 0.2M, close 0.15M to
    0.5M (`INV:1084-1089`); 4a-ii's six lines sum to 2.61M against the
    recorded 2.7M (`INV:1075`).
25. Every conductor ruling 8a applies is named in a task or a global
    constraint: ruling 1 (Task 8), 3 (constraint plus Task 4), 5 and 6
    (Task 7), 7 (C10 and Task 7), 8 (C15 and Task 3), 9 (C7 and Task 3),
    10 (Task 5), 11 and 12 (Task 1), 13 (constraint plus Task 0 item 3),
    14 (constraint, no interface until the MP3 pass). Ruling 2's
    divergence route is placed in the owner-touchpoints section, and
    ruling 4 is correctly left whole to 8c.
26. The prior-art work is done rather than gestured at. Each task's
    entry reproduces the verdict the spec's table records and none
    invents one: adopt and extend the repo's own reader
    (`spec:1026`), decline `mewkiz` for tags because no block encoder is
    exported (`spec:1027`), decline `go-flac` v2 as a dependency and
    read `saveInPlace` as a design reference (`spec:1028`), decline
    `flacvorbis` with its base (`spec:1029`), decline `dhowden/tag` on
    the unbounded allocation in issue #116 (`spec:1030`), decline
    `gcottom` on license and adoption (`spec:1034`), and adopt Picard's
    tables for hook 13's per-container names in the follow-on passes
    (`spec:1035`). Only the line numbers were off (M-g).
27. The independence marks are true by files. Task 1 writes only
    `internal/atomicfile`; Task 2 only `internal/engine/tagio/testdata`;
    Task 4 only `tagset.go`, which no other task touches. The execution
    model's claim that Tasks 1, 2 and 4 are mutually independent holds,
    and the plan is right that they still run sequentially because one
    tree runs one executor and the module gate runs over `./...`.
28. The staleness pre-task names both moving pieces the brief calls for:
    rung 1b at item 3 with the confirmation framing the spec asks for
    since 8a declares no key (`spec:1188-1190`), and pass 4b's Task 6 at
    item 4 with the right escalation rule that a behavior change under
    `internal/engine/shadow` is a finding while a comment change is not.
29. The pass stays inside 8a's scope. No task decodes a picture, builds
    the seam registry, writes the conformance instrument, or edits
    `internal/engine/shadow`, and the plan states the stop condition for
    each. The golden carrying a PICTURE block is spec-sanctioned rather
    than scope creep: deliverable 6 requires it (`spec:100-107`) and the
    contested-surface row says it is generated once and never
    regenerated (`spec:1131`).
30. R4's scoping argument is right. The oracle reserves 1 KiB of padding
    and ruling 6 reserves 8 KiB, and the oracle's vendor is
    `Mutagen 1.48.1` while a fresh `tagio` block names `dubplate`, so a
    whole-container comparison would fail by design. The plan states
    that reasoning rather than leaving a reviewer to discover the
    mismatch.
31. The serialization contract is faithful to the pinned library at
    every point checked against INV: the three-writable-style fields and
    their key order (`INV:234-237`), `albumartists`' read-only
    `ALBUM ARTISTS` (`INV:429`), deleting `barcode` removing all five
    keys (`INV:238-240`), the one-element list through `single_field`
    for the five named fields (`INV:271-275`), the Q8 shift
    (`INV:292-295`), the date decomposition with its truthiness rule and
    unpadded `YEAR` fallback (`INV:277-290`), and the plural winning all
    three colliding pairs (`INV:297-308`).
32. The four-state model's scale evidence is carried correctly. Row 0's
    ten nulls against ten empty strings comes from `INV:387-403`, and
    the 45 empty-list cells against 149 empty strings in Task 4's
    mutation 4 comes from `spec:321-324`. Splitting the two empty-form
    mutations rather than folding them is the right response to the
    defect revision 1 of the spec carried.
33. Task 8's sixteen ledger-silent keys match `spec:580-584` name for
    name, and the plan keeps `bpm`'s test for the reason the spec gives.
34. The mutation proofs name what fails, and the unfired ones are
    labeled as unfired with a reason rather than dressed up: the
    directory fsync's crash property, the streaming property held by a
    reviewer's read, the golden truncation no test sees until Task 3,
    the read-only mark whose consequence is a write, the intermittent
    map-order mutations that report their run count, and the
    deviation-note completeness. This is the assertion discipline the
    pass says is its own risk control.
35. The prose meets the developer-docs standard on every mechanical
    measure: tellgrader reports 0 tells over 14,444 words, cadence CV
    0.85, no em dash anywhere, no connector openers. Only the bolded
    lead-phrase rate is off (m11).
36. `WriteFile`'s current shape supports R10 exactly: 42 lines, one
    exported function at `:15`, the temp file in the target's own
    directory at `:16-17`, `Chmod` after `Close` and before `Rename` at
    `:35-38`.
37. C16's precedent is real: `internal/engine/shadow/flac.go:209-228`
    carries the eighteen-row citation block the mapping's shape copies.
38. The corpus handoff exists and holds what C1 and Task 4 read from it:
    `corpus/README.md:2354` opens "Handoff to rung 8", `:2381` names the
    `writable_fields` header reproduced there, and the nine read-only
    keys are named with the reason they are never a writer's target.
39. `spec:1102`'s exit gate and the plan's Task 7 and Task 8 criteria
    correspond item for item, with one exception: the exit gate's
    "asserted key by key against `INV:409-427`" inherits B1's
    correction.
40. Deliverable counts are stated per task and none exceeds four. Ten
    entries including the pre-task and the close sit inside the spec's
    fifteen-task split trigger, and the plan names its cut point with a
    two-reason argument against taking it, which is what CLAUDE.md's
    pass-sizing rule asks of an orchestrator.
41. The ten Resolutions each carry an application, a fold if ruled
    otherwise, and, where one applies, a prior-art line. None is left
    open.
42. The plan carries no placeholder, no `TBD`, and no unresolved
    cross-reference in 1,827 lines.
43. `go.mod` carries cobra and its two indirects, so the
    no-new-dependency constraint is checkable mechanically at every task
    end, and no task's Interfaces section names an import that would
    break it.
