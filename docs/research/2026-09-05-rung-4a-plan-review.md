# Adversarial review: the rung 4 pass 4a plan

Fresh-context review (2026-09-05) of
`plans/2026-09-05-rung-4a-shadow-runner.md` revision 1, in two
independent lenses: engine fidelity and the FLAC and wire seams,
verified signature by signature against `internal/engine/{tag,matcher,
musicbrainz}` and against pinned mediafile 0.17.0, mutagen and beets
2.13.1 in `tools/venv/.venv`, and harness, config, estate inputs, and
process and plan integrity, verified against the tree at
`54e534b0`. Folded as revision 2 the same day; the merged 74-item
"Verified sound" list below is the do-not-re-litigate record for the
pass, alongside the rung 4 spec review's own 28-item list and rung 3's
pass records.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1 through B7, M-a through M-g, m1 through m18. Every
finding was re-verified by the fold editor against the tree or the venv
before folding, and every numeric claim was re-measured rather than
carried. Three reviewer claims did not survive that check in the form
they were written and are corrected below with the evidence; none of the
three retires its finding. Three items are conductor-owed rather than
folded silently, and each is applied provisionally in the plan with an
inline marker.

Both lenses agree the plan is unusually well grounded. Twenty-three
contracts and twenty resolutions each resolve to the lines they name,
`QuietImportTunables` provably equals `gradeTunables("estate")` today,
contract S16's five-step re-score matches `gradeCaseDefault:568-622`
step for step, Resolution R7's mutual-exclusion argument about the two
deferral branches is exactly right against `Budget.Wait`'s
reserve-before-wait, the nine contested-surface rows are all present in
the parallel-approach spec, the ceiling arithmetic reconciles against
`docs/HISTORY.md`, and there is no placeholder, no `TODO` and no
unresolved cross-reference in 5,420 lines. The defects concentrate in
five places the plan reasons about from memory rather than from the
artifact: the FLAC container's own byte layout, the committed ledger's
own bytes, the pass loop's iteration order, beets' own grouping key, and
three pieces of literal test code that cannot do what the plan says they
do.

---

## Corrections to the reviews' own evidence

**1. Review 1's m3 gives the `RunID` comment span as `importer.go:156-160`.**
Re-measured, it is `:158-161` (`:158` opens "RunID identifies this
report's file", `:161` closes "path."; the field itself is `:162`).
Review 2's `:158-162` is right. The finding stands: the plan's
`:139-145` is `SessionID`'s comment, not `RunID`'s. The correct citation
folded is `:158-161`.

**2. Review 2's m7 calls `checkNoHarnessImports` "parameterized on a
`forbiddenImportPrefix`".** It is not a parameter;
`forbiddenImportPrefix` is a package-level constant and the walker takes
`root string` alone (`internal/harness/conform/arch_test.go:30`). The
finding stands: no mechanical guard exists for "`internal/engine/shadow`
imports `internal/config` never", and a walk is the repo's own precedent
for a rule this load-bearing. Only the "few lines" estimate rests on the
corrected reading.

**3. Review 2's m1 attributes the `thresholds_test.go:35`, `:45`
citation to "Task 0 item 11 and R1".** R1 names no such citation; the two
sites are Task 0 item 11 (`plan:1253-1255`) and Task 3's `ThresholdKind`
note (`plan:1696-1697`). The finding stands and is measured: the only
`Threshold` comparison is `thresholds_test.go:34`
(`got[0] != first || got[1] != second`); `:101` compares an `int`, and
there is no comparison at `:35` or `:45`.

---

## BLOCKERS

### B1. Task 4's STREAMINFO layout is wrong. The 64 bits carrying sample rate and total samples sit at payload bytes 10 through 17, and the payload's last 64 bits are the tail of the MD5 signature

**Lens.** Engine fidelity.

**Plan line.** Task 4, "The container format, stated so the implementer
builds to it rather than to a memory" (`plan:2104-2117`): "`STREAMINFO`
is type 0 with a 34-byte payload, whose last 64 bits pack sample rate
(20 bits), channels (3), bits per sample (5) and total samples (36);
`Length` is total samples divided by sample rate".

**Evidence.** STREAMINFO is 16 + 16 + 24 + 24 + 20 + 3 + 5 + 36 + 128
bits, 272 bits, 34 bytes. Pinned mutagen writes it in that order at
`tools/venv/.venv/lib/python3.13/site-packages/mutagen/flac.py:233-257`:
two bytes of min blocksize, two of max blocksize, three of min frame
size, three of max frame size (`:235-238`), then the packed eight bytes
(`:240-252`), then sixteen bytes of MD5 (`:253-257`). Its reader
confirms the same offsets at `:208-231`. So the pack occupies
`payload[10:18]` and `payload[18:34]` is the MD5. The payload's last 64
bits, `payload[26:34]`, are the MD5's second half.

**Consequence, built as written.** The implementer takes
`payload[26:34]`, reads a sample rate and a total-sample count out of an
MD5 digest, and divides one by the other. Every real inbox FLAC then
yields a garbage `Length`, `Length` feeds the track-length penalty, and
every live verdict moves, so the shadow's whole comparison is silently
wrong for the thirty-day window. The error survives the task, because
`buildFLAC` is authored in the same task by the same implementer from
the same sentence, so the builder and the parser agree on the wrong
offsets. `TestReadFLACLengthFromStreaminfo` passes, and named mutation 3
still fires against the wrong bytes. The only cited external source,
`internal/harness/oracle/testdata/gen_template_flac.py:64-75`, delegates
the payload to `StreamInfo.write()` at `:61` and states no offsets.

**Fix.** State the layout by offset rather than by "last", and pair it
with M-a's cross-check, because a corrected sentence with no external
detector is one transcription away from the same failure.

### B2. Task 10's byte-for-byte re-marshal test cannot pass, for two independent reasons

**Lenses.** Both, independently. Review 1 found the field-order cause and
the escaping cause; review 2 found the escaping cause with a quoted
probe. Review 1's evidence is the stronger of the two and is what the
fold takes; review 2's probe is kept because it names the repo's own
landed precedent for the hazard.

**Plan lines.** Task 10's
`TestTheCommittedLedgerRemarshalsUnchangedWithTheNewFields`
(`plan:4168-4175`); its acceptance criterion (`plan:4206`); contract S21
(`plan:662-675`); named mutation 3 (`plan:4215-4218`).

**Evidence, cause one: field order.**
`internal/harness/analysis/adjudication.go:46-67` declares the fields
`ID, Subject, Suspected, Citation, ReleaseTruth, Disposition,
CorrectedVerdict, Reason, Adjudicator, Date, DefectFree`, and
`encoding/json` emits struct fields in declaration order. Every one of
the 21 committed entries carries the key order `id, subject, suspected,
citation, release_truth, disposition, reason, adjudicator, date,
corrected_verdict[, defect_free]` (re-measured: 10 rows with
`defect_free`, 8 with `corrected_verdict` and no `defect_free`, 3 with
neither). Re-serializing the decoded ledger in Go field order first
diverges from the file at line 12.

**Evidence, cause two: HTML escaping.** The file carries one literal `<`
and three literal `>` (re-measured over
`corpus/analysis/adjudications.json`, one of them at `:188`), which
`json.Marshal` and `json.MarshalIndent` escape. Nothing under
`internal/` sets `SetEscapeHTML(false)` today, so no landed precedent
hands the implementer the answer, though the repo already records the
hazard at `internal/harness/cassette/cassette.go:9-11`.

**Consequence, built as written.** The test fails for two reasons that
have nothing to do with `omitempty`, and both available repairs are
wrong. Reordering `AdjudicationEntry`'s fields is a schema change to a
shared corpus type outside this pass's declared scope. Relaxing the
comparison to a semantic one destroys the guarantee the test exists to
prove and takes mutation 3's only detector with it.

**Fix.** Replace the byte-for-byte re-marshal with the guarantee that
actually holds: decode into `AdjudicationLedger`, re-encode through a
`json.Encoder` with `SetEscapeHTML(false)` and `SetIndent("", "  ")`,
and compare the original and re-encoded bytes as decoded
`map[string]any` trees, which is order-independent and still catches a
dropped or added key. Assert additionally that the raw file contains
neither `matcher_commit` nor `matcher_config_hash`, which is mutation
3's real detector. Correct contract S21 and draft the `spec:845-846`
correction: the committed ledger is not in Go struct field order, so
byte identity was never available.

### B3. The match step iterates in observation-id order over a digest that ingests modification times, so Task 14's numeric-prefix fixture cannot control which album the cap starves and Task 9's determinism test cannot pass

**Lenses.** Both. Review 1 found the ordering half and the determinism
half; review 2 found the determinism half with the helper signature that
proves no fixture pins an mtime. Both are folded.

**Plan lines.** Task 9's pass loop step 4 (`plan:3892-3893`), "For each
newly reconciled observation, **in lexical id order**"; `Observations`'
contract (`plan:2628`); Resolution R5 (`plan:816-828`), the digest over
"its name, its size in bytes, and its **modification time** in RFC 3339
nanosecond form" and "`ID` is `shadow-` plus `Digest`'s first sixteen
hex characters"; Resolution R18 (`plan:1063-1071`); Task 9's
`TestRunIsDeterministicUnderAFrozenClock` (`plan:3972-3976`); Task 14's
mutation 5 and its `res.Pass.Budget.Spent != tree.budgetCap` assertion
(`plan:4987-4989`, `:5104-5106`).

**Evidence.** Two facts read together. First, an id is a SHA-256 prefix,
so lexical id order bears no relation to the `01-` through `13-`
basename order R18 chose. The capture walks lexically and is correctly
grounded in `os.ReadDir`, but the match step, which is where the budget
is spent, does not. Second, the digest ingests each file's `ModTime`,
and `writeAlbum(t, dir, uploader, basename string, tracks [][]string)`
(`plan:1132-1134`) takes no time argument; nothing in Task 6, Task 9 or
Task 14 pins a modification time with `os.Chtimes`, and Task 6's
acceptance only requires that the digest *move* on a touch. Two fresh
`t.TempDir()` trees written at different wall times therefore carry
different mtimes, different digests and different ids.

**Consequence, built as written.**
`TestRunIsDeterministicUnderAFrozenClock` fails on its first assertion,
for a reason unrelated to the code under test. In
`TestObservePassEndToEnd` the album whose first request reserves past
the cap is whichever digest happens to sort last, so `13-defer-budget`
is starved only by luck, `12-defer-unavailable` may be starved before it
can spend its seven attempts, and `res.Pass.Budget.Spent ==
tree.budgetCap` becomes flaky. Mutation 5 cannot fire at all. This is
the pass's exit gate, so a flaky failure here is the most expensive
kind. The obvious local repair, taking mtime out of the digest,
contradicts Task 6's mutation 4 and weakens the identity the whole
idempotence contract rests on.

**Fix.** Two edits. Make the pass loop match in capture order,
`(uploader, basename)`, say so in Task 9's step 4, in `Store.Observations`'
contract and in R18, and carry the ordered slice on `PassResult` so the
order is a value rather than a filesystem property. And require every
fixture writer to pin each written file's modification time with
`os.Chtimes` to a fixed instant, stated once in the helper list so
Tasks 6, 7, 9 and 14 all inherit it, so a digest is a function of the
fixture alone.

### B4. The capture's grouping key drops beets' `albumartist or artist` fallback, so the one directory shape the rule exists to exclude is observed instead

**Lens.** Engine fidelity.

**Plan lines.** Contract S9 (`plan:498-509`): "Group a directory's items
by `(AlbumArtist, Album)`. One group is observed." Task 6's `capture`
doc comment and Step 3 (`plan:2858-2865`, `:3000-3002`). Task 6's
mutation 2 (`plan:3036-3040`).

**Evidence.** Pinned beets 2.13.1 groups with the artist fallback, at
`tools/venv/.venv/lib/python3.13/site-packages/beets/importer/stages.py:102-103`:

```python
def group(item: library.Item) -> tuple[str | None, str | None]:
    return (item.albumartist or item.artist, item.album)
```

consumed by `sorted(task.items, key=group)` and
`itertools.groupby(sorted_items, group)` at `:112-113`. The box enables
that stage at `config/beets-config.yaml:23`, and the estate's own reason
at `:19-22` names the material: contributor ZIPs that extract flat.

**Consequence, built as written.** A flat directory whose tracks carry
`ARTIST` and no `ALBUMARTIST` is the common case for that material, and
many rippers write no `ALBUMARTIST` at all. Under the plan's key every
such track maps to `("", "<album>")`, so the directory groups to one and
is observed; under beets' key it splits by artist and produces several
verdicts. The shadow then scores one Go verdict against one of several
run-report rows and records it as an observation, which is a fabricated
comparison in the artifact whose whole value is that it is honest, and
the exact failure the grouping rule was added to prevent (spec fold B1).
Task 6's own mutation 2 makes it concrete: "two artists under one album
title" with no `ALBUMARTIST` is a single group under `(AlbumArtist,
Album)`, so the mutation's own scenario is mishandled by the unmutated
code.

**Fix.** Group by `(cmp.Or(AlbumArtist, Artist), Album)`, citing
`stages.py:102-103` beside the `group_albums` citation. Rewrite Task 6's
mutation-2 row to carry two `ARTIST` values, one `ALBUM` and no
`ALBUMARTIST`, and make the mutation "drop the artist fallback".
**Conductor ruling owed**: the spec states the key as "their
`albumartist` and `album` fields" (`spec:441`), so this is a spec
amendment rather than a plan-internal choice.

### B5. `Observation` has no field for a reconcile-time exclusion, and Task 14's acceptance criterion is therefore unsatisfiable

**Lens.** Harness and process.

**Plan lines.** Task 5's type block (`plan:2537-2556`), whose doc clause
reads "Exactly one of Class, Deferred and Unreconciled is set on a
stored record"; Task 7's `reconciliation` carrying a separate `Excluded
string` (`plan:3150-3156`); Task 9's pass loop step 3 (`plan:3888-3891`),
"write the oracle signal, the unreconciled reason, **the exclusion** or
the supersession onto the stored observations"; Task 14's acceptance
(`plan:5077-5079`).

**Evidence.** The exclusion vocabulary is a third, distinct constant
group (`plan:2473-2477`) summed into `PassRecord.Excluded`
(`plan:2585`). Two of the three exclusions happen at capture time and
never become observations. The third does not: fixture row 7,
`07-gate-rejected` (`plan:4886`), is a single-album directory with one
track, so `capture` stores it as an observation, and only `reconcile`
later learns from the run report's `rejected` row that it never reached
beets. The plan's own count confirms it: "stored observations = 12
(thirteen directories minus the excluded multi-album one)"
(`plan:5014-5015`) includes `07-gate-rejected`. That record has no
`Class`, no `Deferred` and no `Unreconciled`, and `Observation` offers
no fourth field.

**Consequence, built as written.** An implementer building Tasks 5 and 9
exactly as written commits a stored schema in Task 5, then discovers at
Task 14 that one of thirteen fixtures has nowhere to land. Adding a
field after Task 5 shipped is a schema change mid-pass; deleting the
record makes the stored count 11 against the plan's asserted 12.

**Fix.** Add `Excluded string \`json:"excluded,omitempty"\`` to
`Observation` in Task 5 beside `Unreconciled`, restate the invariant in
four terms, and restate Task 14's acceptance in those four terms.
`BuildReport` already carries `Excluded map[string]int`
(`plan:4476`), so the report row needs no change.

### B6. Task 1's crash-safety test asserts on a file the failing write never touches

**Lens.** Harness and process.

**Plan lines.** `plan:1336-1362`, the added
`TestWriteFileLeavesTheOldContentWhenTheWriteFails`, given as literal
code and described as "the property the store depends on and the salvage
suite proves only indirectly".

**Evidence.** The test writes `"first"` to `dir/state.json`, then calls
`WriteFile(filepath.Join(blocked, "state.json"), ...)` where `blocked`
is `dir/blocked` and `dir/blocked/state.json` is a directory, then reads
`dir/state.json` back and asserts it still holds `"first"`. Those are
two different paths. `_salvage/internal/atomicfile/atomicfile.go:16-17`
creates the temp file in `filepath.Dir(path)`, which is `dir/blocked`,
and `:38` renames within `dir/blocked`; `dir/state.json` is not opened,
renamed over, or referenced on any code path. The assertion holds for
every possible implementation of `WriteFile`, including one that never
writes anything, so it is exactly the vacuous assertion CLAUDE.md's Go
rules forbid, dressed as the store's crash-safety proof.

The defect propagates into the mutation table. Mutation 1
(`plan:1400-1408`) claims that moving `os.CreateTemp` to `os.TempDir()`
makes the new test fail. Under that mutation the blocked write still
fails and `dir/state.json` still reads `"first"`, so the new test still
passes, and the implementer would report a mutation outcome the code
cannot produce.

**Fix.** Point the failing write at the same path whose content is being
asserted: write `"first"` to `p`, then make `os.CreateTemp` fail by
removing `filepath.Dir(p)`'s write permission (`os.Chmod(dir, 0o500)`,
skipped under `os.Geteuid() == 0`), assert the error, and assert `p`
still reads `"first"`. Restate mutation 1 as what it is, a property this
single-filesystem workstation cannot discriminate, held by the doc
comment and 4b's `ReadWritePaths` argument.

### B7. Task 12's stamp-refusal test has no injection seam, and R3 rejects the only one the plan's signatures allow

**Lenses.** Both, independently; review 1 filed it as a major and review
2 as a blocker. Promoted to a blocker on the two-lens agreement and on
R3's own words.

**Plan lines.** Resolution R3 (`plan:770-791`) places `var matcherCommit
= "dev"` in `cmd/dubplate-server/main.go` and the refusal in
"`cmd/dubplate-server/shadow.go`'s `observe` RunE", then rejects "an
overridable package variable the test writes, which is a global mutated
by tests and reintroduces the ordering hazard R12 closes". Task 12
declares `newShadowCmd(environ []string, stdout, stderr io.Writer)
*cobra.Command` (`plan:4509`) and specifies
`TestObserveRefusesAnUnstampedBuild` as a two-row table over stamps
(`plan:4587-4594`). Mutation 1 depends on the same test.

**Evidence.** With the stamp read from a package-level `var` inside RunE
and no stamp on the constructor, a table over stamps has exactly one
mechanism available: assigning `matcherCommit` in the test, which is the
rejected option. The plan's other rejected alternative, a
`--matcher-commit` flag, is rejected for a separate and good reason. The
constructor shape the plan pins matches the landed
`newCheckCmd(environ []string, stdout, stderr io.Writer)`
(`cmd/dubplate-server/check.go:16`) and `newDoctorCmd`
(`doctor.go:23`), neither of which carries a stamp.

**Consequence.** The implementer either writes the global from a test,
contradicting the resolution the plan says it is implementing, or
changes a signature the plan and the spec both pinned. Either way the
pass's sharpest under-specification, in R3's own words, is resolved by
an implementer mid-task. This is also the refusal 4b's whole
deploy-time assertion is the second line of defense for, so a weakened
or deleted test here is expensive later.

**Fix.** Give the constructor the stamp:
`newShadowCmd(environ []string, matcherCommit string, stdout, stderr io.Writer) *cobra.Command`,
called from `newRootCmd` as
`newShadowCmd(os.Environ(), matcherCommit, os.Stdout, os.Stderr)`. The
package `var` stays unwritten by any test, R3's rejection stands, and
the table becomes a plain argument. **Conductor ruling owed**: the spec
pins the constructor shape at `spec:1063-1066`, so extending it by one
parameter is a spec deviation on a contested surface.

---

## MAJORS

### M-a. Nothing in Task 4 can catch a wrong container layout, because one task authors both the fixture builder and the parser and no test reads a real FLAC

**Lens.** Engine fidelity.

**Plan lines.** Task 4, Step 1 (`plan:2199-2201`): "`flac_test.go` builds
its fixtures in Go. **No binary fixture is vendored**", and the spec's
own reasoning at `spec:1024-1030`.

**Evidence.** Every one of the five Task 4 tests reads bytes produced by
`buildFLAC`, so any shared misreading of the format is invisible. The
tree does hold one real, valid FLAC:
`internal/harness/oracle/testdata/template.flac`, re-measured at 42
bytes, a signature plus one last-block STREAMINFO carrying mono,
44100 Hz, 16-bit, zero samples
(`internal/harness/oracle/testdata/gen_template_flac.py:38`, `:52-61`,
`:64-70`).

**Consequence.** B1 is the instance; this is the reason B1 would have
shipped. The task's mutation set is honest about everything except the
one property no test can see.

**Fix.** Add two assertions. Embed the 42 committed template bytes as a
Go byte literal inside `flac_test.go` (which keeps the engine from
reading a harness path and leaves `TestEngineNeverImportsHarness`
untouched), parse them, and assert sample rate 44100, channels 1, bits
per sample 16 and total samples 0. And assert `buildFLAC`'s own
STREAMINFO block byte for byte against a hand-written 34-byte
expectation derived from B1's offsets, so the builder is pinned to the
format rather than to the parser.

### M-b. Task 8's 503 test cannot both spend seven metered tokens and finish in under a second, and the acceptance criterion names the wrong cause

**Lens.** Engine fidelity.

**Plan lines.** Task 8's
`TestMatchAlbumSpendsOneTokenPerAttemptIncludingRetries`
(`plan:3600-3607`); the acceptance criterion (`plan:3664-3666`);
mutation 4 (`plan:3687-3689`).

**Evidence.** `internal/engine/musicbrainz/client.go:346` calls
`c.budget.Wait(ctx)` before every attempt including retries. A metered
budget's wait routes to `Limiter.acquire` (`limiter.go:293-321`), which
holds at most one spare token (`:301-304`) and otherwise blocks until
the dispatcher's next tick; `NewLimiter` drives that from
`time.NewTicker(interval)` (`:110-114`). Seven attempts therefore
require at least six ticks. Contract S15 fixes the shadow's interval at
`2 * musicbrainz.DefaultInterval`, two seconds (`limiter.go:62`), so the
row takes at least twelve seconds. The precedent the plan cites,
`internal/engine/tag/tag_test.go:52-64`, escapes this only because
`clientFor` hands the client `musicbrainz.Unmetered()`, which this
task's mutation 4 rules out.

**Consequence.** An implementer building the stated `Config` gets a
twelve-second test that violates the task's own acceptance criterion,
and the plan gives no instruction that would resolve it. Task 14 shows
the plan knows the answer (`Interval: time.Millisecond`), but Task 8
lands first.

**Fix.** State in Task 8's test that the `Config` carries `Interval:
time.Millisecond`, and reword the acceptance criterion to name both
terms: `Retry-After: 0` collapses the retry backoff
(`client.go:388-389`, `retryWait` at `:404-412`) and a millisecond
interval collapses the limiter's pacing (`limiter.go:110-114`).

### M-c. Two of Task 9's named mutations cannot fire

**Lens.** Engine fidelity.

**Plan lines.** Task 9, mutation 3 (`plan:4032-4036`) and mutation 5
(`plan:4041-4043`), together with
`TestRescoreDefaultIssuesNoSearchAndSpendsNoToken` (`plan:3951-3958`).

**Evidence.** For mutation 3, `matcher.ActionFor` under `Quiet` returns
`ActionApply` if and only if `rec == RecStrong`, and otherwise
`ParseAction(QuietFallback)`, which `QuietImportTunables` pins to
`skip` (`internal/engine/matcher/verdict.go:179-210`). Action is
therefore a bijection of `rec == RecStrong` under the estate tunables,
which is contract S3's own statement, so a classifier comparing
`Recommendation == RecStrong` and one comparing `Action == ActionApply`
agree on every input and no table row can produce the shape the plan
instructs the implementer to add.

For mutation 5, the signature the same task declares
(`plan:3755-3756`) takes `recorded`, `bodies`, `albumIDs`, `items` and
`evalDate`, and no budget, client or `Config`, so there is no pass
budget in scope to hand the replay client and none to assert on.

**Fix.** Replace mutation 3 with one the estate tunables can detect,
classifying on `Recommendation != RecNone`, which does diverge from the
action projection on a `RecMedium` verdict. Move mutation 5 and its
assertion up to a `Run`-level test that reads `PassRecord.Budget.Spent`
after a pass containing one re-scored observation, and assert the
recorded spend equals mode one's request count.

### M-d. `DUBPLATE_SHADOW_ENABLED=false` becomes a load error, with a self-contradictory message

**Lens.** Config and estate inputs.

**Plan lines.** Task 3's `rejectNamespace` arm, given as literal code
(`plan:1927-1933`).

**Evidence.** `integrationClassOf` must resolve the flag itself: Task 3's
own test asserts it, iterating
`slices.Concat([]string{c.Flag}, c.Required, c.Optional)` and requiring
every one to resolve back to the class (`plan:1900-1906`).
`internal/config/config.go:133-152` walks every set `DUBPLATE_*` name,
and `w.get` treats only an empty value as unset (`config.go:92-98`). So
an environment carrying `DUBPLATE_SHADOW_ENABLED=false` reaches the arm
with `name == class.Flag`, `enabled == false`, and appends
"DUBPLATE_SHADOW_ENABLED: set without DUBPLATE_SHADOW_ENABLED, the
shadow integration's enable flag". None of the six table rows in
`TestLoadShadowIntegration` (`plan:1792-1827`) covers a set-and-false
flag, so the gate stays green and the defect ships.

**Consequence.** An adopter who explicitly turns the integration off
cannot start the server. That is the config wall refusing a legal
configuration, the failure class ruling 1 exists to prevent.

**Fix.** Skip the flag in the rejection arm (`if name == class.Flag {
continue }` before the enabled test), and add two rows to the table:
`DUBPLATE_SHADOW_ENABLED=false` alone loads clean with a nil `Shadow`,
and `DUBPLATE_SHADOW_ENABLED=false` plus a shadow path errors naming the
flag.

### M-e. Adding doctor rows breaks `doctor_test.go`'s exhaustive line count, and the new duration row's registration is test-order dependent

**Lens.** Config and estate inputs.

**Plan lines.** Task 3, Step 3 (`plan:1960-1961`) and the test direction
(`plan:1910-1913`); the acceptance criterion "every existing
`internal/config` test passes unchanged" (`plan:2010`).

**Evidence, two halves.** `cmd/dubplate-server/doctor_test.go:110-112`
closes its walk with `if li != len(lines) { t.Errorf("stdout has %d
extra line(s) beyond owned+reserved+threshold rows: ...") }` after
iterating `config.OwnedWall()`, `config.ReservedClasses()` and
`config.Thresholds()` (`:73-108`). Inserting integration rows makes that
assertion fire, and the acceptance criterion names `internal/config`
only, so the plan never says the loop needs a fourth arm over
`config.IntegrationClasses()`.

Second, `config.Thresholds()` returns the package-level
`registeredThresholds` (`internal/config/thresholds.go:22-35`), and the
plan registers only from `newRootCmd` through `sync.OnceFunc`
(`plan:1764-1772`). Go runs a package's tests in source order across
files sorted by name, so `doctor_test.go` runs before `main_test.go`,
whose `newRootCmd()` call at `main_test.go:14` is the only construction
in the binary today. A doctor test asserting the duration row therefore
sees an empty registry, and `go test -run TestDoctorCmd
./cmd/dubplate-server/` sees a different registry state than a full
package run. That is exactly the ordering hazard R12 says it closes.

**Fix.** Name both edits in Task 3: add the `config.IntegrationClasses()`
arm to `doctor_test.go`'s walk in the same position `doctor.go` prints
it, and have every doctor test that asserts a threshold row call
`registerShadowThresholds()` as its first statement, idempotent by
construction. Restate the double-construction acceptance criterion as
"constructing `newRootCmd` twice adds no rows beyond the first
registration", which is order-independent.

### M-f. The exit gate's verdict rows need release documents that score `RecStrong`, and the plan names no recipe and no reference body

**Lens.** Harness and estate inputs.

**Plan lines.** Task 14's fixture table (`plan:4880-4892`), rows 1, 4 and
5, each requiring a body "scoring strong"; row 1 additionally requiring
"One request, the ID-consensus early return". The only construction
guidance is `releaseBody(albumID, albumTitle, artist string, tracks
[]string) []byte`, "a minimal WS/2 release document" (`plan:1141-1143`).

**Evidence.** The early return fires only at `rec == matcher.RecStrong`
(`internal/engine/tag/tag.go:106`), and `RecStrong` needs an album
distance at or below `StrongRecThresh = 0.04`
(`internal/engine/matcher/tunables.go:268`) across a twenty-weight
penalty surface (`tunables.go:275-296`). Whether a "minimal"
hand-authored body clears that bar is discovered by iterating, not
reasoned to, and rows 1, 4 and 5 carry three of the five verdict
outcomes the exit gate exists to prove.

The repo already holds the answer and the plan does not point at it:
`internal/engine/tag/testdata/early-return/release-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.json`
is a committed body that `internal/engine/tag/tag_test.go:141-158`
asserts scores `matcher.RecStrong` through the identical `tag.Album`
path, with `search-fallback/` and `medium-id/` beside it. It is engine
testdata rather than corpus, so reading and adapting it breaches no rule.

**Fix.** Add one sentence to Task 14's Step 3: author `releaseBody` by
adapting `internal/engine/tag/testdata/early-return/`'s committed
release document, changing only ids, titles and lengths, and adapt
`medium-id/`'s body for a row that must not score strong. State that a
row failing to reach `RecStrong` is a fixture defect rather than a
matcher finding.

### M-g. An `imported` row whose library file carries no `mb_albumid` fabricates a `go-applies-different-release`

**Lens.** Harness and estate inputs.

**Plan lines.** S10's apply rule (`plan:521-531`): "the shadow reads the
**first** file `LibraryPaths` names through `readFLAC`, takes
`MBAlbumID` as the oracle's chosen release". The only guarded case is an
empty `LibraryPaths`, which becomes `UnreconciledNoLibraryPaths`. The
classification table (`plan:3853-3857`) then compares
`oracle.ReleaseMBID` against `Candidates[0].Info.AlbumID`.

**Evidence.** `readFLAC`'s own contract is that an absent tag takes the
field's zero value (`plan:2084-2086`), which is `ProjectItemFields`'
typed-default rule, verified at `internal/engine/matcher/info.go:178-211`
(`MBAlbumID: stringField(raw, fieldMBAlbumID)`, and `stringField`
returns `""` for an absent key). So a library file the reader opens
successfully but that carries no `MUSICBRAINZ_ALBUMID` yields
`MBAlbumID == ""`, and the reconcile step records `OracleSignal{Signal:
"apply", ReleaseMBID: ""}`. `classify` then sees `R != G` and writes
`ClassGoAppliesDifferentRelease` plus a pending row whose
`subject.release_mbid` is the Go id. That is a divergence against a
release the oracle never named: a fabricated row in the artifact
contract S6 says must be honest, consuming exactly the owner
adjudication ruling B's stamp exists to save.

The case is reachable rather than theoretical. `LibraryPaths` is a
positional correlation the importer fills by matching staging order
against beets database order, "and a count mismatch leaves the extra
units empty" (`_salvage/internal/importer/importer.go:1119-1136`), so
the first named path is not guaranteed to be the album the row means,
and the plan's own titles-and-lengths join is what catches a mismatch,
not the tag's presence.

**Fix.** Make an unresolvable release id on an `imported` row an
`unreconciled` outcome rather than an apply, by widening the
second reason to "an `imported` outcome with no resolvable release id"
(covering both an empty `LibraryPaths` and an empty `MBAlbumID`),
renaming its constant value to `imported-with-no-release-id`, and adding
a firing fixture row in Task 7, since R20 makes the no-fixture case the
pass's named anti-pattern.

---

## MINORS

**m1. `sourceDir`'s three assignment sites.** Contract S10
(`plan:513-514`) cites `:769`, `:905-907`, `:1066-1068`. Re-measured,
they are `_salvage/internal/importer/importer.go:766` (`filepath.Dir(srcDir)`),
`:908` (`inboxUserDir`) and `:1066` (`filepath.Join(InboxDir, uploader)`),
and `:905-907` carries no assignment. The spec has them right
(`spec:480`), so this is drift the plan introduced. Both lenses.

**m2. `importer.go:136` is the bare field, not the doc.** Contract S10
and Task 7's `oracleAlbumOutcome` comment (`plan:3096-3097`) cite `:136`
for `SourceDir` being "documented the uploader's inbox directory".
Re-measured, `:136` is `SourceDir string \`json:"source_dir"\`` inside
`AlbumOutcome`; the wording is on `albumUnit.sourceDir` at `:212`. Cite
`:212` for the doc and `:136` for the field. Both lenses.

**m3. The `RunID` comment is `:158-161`.** Contract S12 (`plan:562-563`)
and Resolution R6 (`plan:842-845`) cite `:139-145` for `RunID` being
"not derived from `StartedAt`". Re-measured, `:139-145` is
`AlbumOutcome.SessionID`'s comment; `RunID`'s runs `:158-161`. The
companion citation `:456-462` for `newRunID` is correct. Both lenses.

**m4. The `Threshold` `!=` comparison is at `:34` alone.** Task 0 item
11 (`plan:1253-1255`) and Task 3's `ThresholdKind` note
(`plan:1696-1697`) cite `thresholds_test.go:35` and `:45`. Re-measured,
the only `Threshold` comparison is `:34`; `:101` compares an `int`, and
neither `:35` nor `:45` compares anything.

**m5. `resetThresholds` does not exist.** The helper list
(`plan:1125-1126`) names "`resetThresholds` (exists,
`thresholds_test.go:13-15`)". The landed helper is
`withRegistry(t *testing.T)` at `internal/config/thresholds_test.go:11-16`,
which saves and restores `registeredThresholds` through `t.Cleanup` and
is exactly the semantics the plan wants.

**m6. `Case.Write`'s indent is two spaces, not one.** R8
(`plan:894-898`) rejects calling `Case.Write` partly because it
"marshals with a single-space indent".
`internal/harness/cassette/cassette.go:242` is
`json.MarshalIndent(c, "", "  ")`. The rejection stands on its other
half, the non-atomic `os.WriteFile` at `:248`, so only the sentence
needs repair.

**m7. S18's "(Resolution R13)" names no such rule.** Contract S18
(`plan:642-643`) ends "the transport's own miss list is the signal the
caller branches on (Resolution R13)". R13 (`plan:989-1003`) is the
incumbent's report decode struct; no resolution states the miss-list
rule. State the rule inline in S18.

**m8. A round tripper cannot be installed as `ClientConfig.HTTP`.**
Contract S16 (`plan:613-614`) describes "a recording
`http.RoundTripper` **installed as** `ClientConfig.HTTP`
(`client.go:98`)". That field is `HTTP *http.Client`
(`internal/engine/musicbrainz/client.go:98`). Task 8's own code block
does it correctly with `&http.Client{Transport: rec, ...}`. Reword S16
to "installed as the `Transport` of the `*http.Client` that
`ClientConfig.HTTP` carries".

**m9. "An empty or zero value" falls through is half wrong.** Task 4
(`plan:2138-2139`) states "An empty **or zero** value therefore falls
through to the next key". `MediaField.__get__` tests the raw fetched
value, not the cast one
(`tools/venv/.venv/lib/python3.13/site-packages/mediafile/fields.py:48-54`),
and `"0"` is a truthy Python string. So `TRACK=0` with `TRACKNUMBER=5`
yields 0, and an implementer who falls through on a zero-valued parse
yields 5.

**m10. Ten style-precedence fixtures are containers the oracle could not
open.** Task 4's `TestReadFLACStylePrecedence` (`plan:2304-2305`) builds
every row "with samples 0 and rate 0 so `Length` stays 0". A STREAMINFO
with sample rate 0 is refused outright by pinned mutagen
(`mutagen/flac.py:222-223`, "A sample rate value of 0 is invalid").
Build those rows at 44100 Hz with zero samples, which keeps `Length` at
0 and keeps the fixture a file beets could read.

**m11. Task 8's tunables test's stated reason is wrong.**
`TestMatchAlbumRunsUnderFullCompatAndQuietImportTunables`
(`plan:3609-3616`) argues "with `NoneRecAction` at its shipped `ask`, a
no-candidate outcome would not project to skip". Under `Quiet == true`,
`matcher.ActionFor` returns apply on `RecStrong` and otherwise the
parsed `QuietFallback`, never reaching `Timid` or `NoneRecAction`
(`verdict.go:179-210`), and `QuietFallback` already ships `"skip"`
(`tunables.go:316`). The assertion distinguishes quiet from non-quiet
and nothing finer, so it cannot see `QuietImportTunables` against
`DefaultTunables` with `Quiet` set. Replace it with an assertion that
the stored `MatcherConfigHash` equals
`configHash(matcher.QuietImportTunables(), matcher.FullCompat())`, which
moves the day either input moves and is already Task 8's own machinery.

**m12. The unparseable-flag path emits a second, self-contradictory
error.** With `DUBPLATE_SHADOW_ENABLED=maybe`, Task 3's `Load` loop
(`plan:1940-1952`) appends `Enabled`'s parse error and
`rejectNamespace`'s arm appends "set without DUBPLATE_SHADOW_ENABLED"
for the flag itself. The table row only checks that the joined error
names the key, so it passes on the wrong message. M-d's `name ==
class.Flag` guard removes the duplicate.

**m13. `envWith` is named but never used.** The helper list
(`plan:1125-1126`) introduces "`envWith` (a `[]string` builder taking
`k=v` pairs)"; Task 3's test code builds its environment with
`append(slices.Clone(base), tt.extra...)` instead. Drop it, or use it.

**m14. "`internal/engine/shadow` imports `internal/config` never" has no
mechanical guard.** R16 (`plan:265-270`) says so explicitly and rests
the rule on written prose. The repo's own precedent for a rule this
load-bearing is a walk: `checkNoHarnessImports`
(`internal/harness/conform/arch_test.go:30-64`) walks a root and matches
a package-level `forbiddenImportPrefix`, and a sibling walker with a
second prefix is a short file. Either extend it, which widens this
pass's `internal/harness/conform` territory, or record the absence as a
declared gap in Task 15's list rather than leaving it implied.

**m15. Task 0 item 4's stamp count is ambiguous.** It expects "all nine
present in ...`:168-184`, each stamped `conductor ruling, 2026-09-05`;
grep the file for that stamp and report the count" (`plan:1206-1212`).
Re-measured, the grep returns 13, because rows `:168`, `:172`, `:176`
and `:177` carry the same stamp from 3d's fold. State the expected
count, or key the check on the row paths.

**m16. Task 13's mutation 1 needs its fixture's class and candidate
count named.** Mutation 1 (`plan:4827-4831`) fires only for a fixture
whose class collapses to `agreement` once the missing candidate is
dropped, which means a `go-applies-oracle-skipped` observation whose only
admitted candidate is the missing one. With the plan's unspecified
"responses cover one of the two requests" (`plan:4768`), dropping a
candidate can leave a divergence in place and produce `Restamped`, so
the most consequential mutation in the pass would not fire.

**m17. Task 11's mutation 2 cannot be applied.** "Set `Disposition` from
the row instead of from the constant" (`plan:4421-4423`) has no target:
`shadow.PendingRow` (`plan:3834-3845`) declares no `Disposition` field
and its doc comment says so deliberately. Reword the mutation to "supply
`analysis.OracleWrong` in place of `analysis.Pending`", which the
existing decode assertion catches through `adjudication.go:198-200`.

**m18. R18 names the wrong task.** R18 (`plan:1068-1071`) says "**Task
13's** fixture names carry an explicit numeric prefix". The
thirteen-directory fixture is Task 14's; Task 13 is `reevaluate`.

---

## Verified sound

Every item below was checked against the tree at HEAD `54e534b0` or
against the pinned venv by at least one lens, and re-checked by the fold
editor wherever a fold depended on it. This is the do-not-re-litigate
record for pass 4a.

1. Contract S1's `tag.Album` signature and parameter order, verbatim at `internal/engine/tag/tag.go:68-69`.
2. `Proposal`'s five fields at `tag.go:23-29`, `Candidates` ascending and stable.
3. `Candidates[0]` is well defined on both flow paths on a nil error: the early return guards `cands.Len() > 0` (`tag.go:106`), the normal path returns `ErrNoCandidates` on an empty sort (`tag.go:138-140`).
4. `tag.ErrNoCandidates` declared at `tag.go:18` and returned at `:139`.
5. `tag.Album`'s two wrapping shapes, `tag.go:128` ("tag: searching: %w") and `:170` ("tag: looking up %s: %w"), so contract S5's `errors.Is` rule is required rather than stylistic.
6. The 404 swallow inside `lookupAndAdd` via `errors.Is(err, musicbrainz.ErrNotFound)` at `tag.go:166-170`, which is contract S18's whole justification for a 501 refusal.
7. The ID-consensus early return fires only at `matcher.RecStrong`, condition at `tag.go:106`; the search fallthrough at `tag.go:125-126`.
8. `matcher.QuietImportTunables`' four assignments against the shipped defaults: `Timid: false` at `tunables.go:312`, `Quiet: false` at `:315`, `QuietFallback: "skip"` at `:316`, `NoneRecAction: "ask"` at `:317`; `DefaultTunables` at `:266`, `SearchLimit: 5` at `:309`.
9. Task 2's drift guard holds today. `mergeProfile(base, "estate")` is `withImportKeys` (`internal/harness/conform/translate.go:207-208`, `:221-226`) and `gradeTunables` forces `Quiet = true` (`defaults.go:72-79`), so `gradeTunables("estate")` equals `QuietImportTunables()` field for field, maps included. `mergeProfile` reads no YAML, so the guard touches no corpus file.
10. `DefaultTunables()` builds `DistanceWeights` fresh per call (`tunables.go:266-325`), so Task 8's `configHash` test mutating one weight corrupts no package state.
11. `Tunables` carries only JSON-marshalable field types, including `StringDistPatterns []SDPattern` (`tunables.go:61-244`, `SDPattern` fully exported at `:334-337`), so `configHash` over a JSON rendering is deterministic.
12. `matcher.FullCompat` at `compat.go:31-33`, `DefaultCompat` at `:41-46`, `Compat.Apply` swapping `StringDistPatterns` at `:51-63`, with all four combinations naming a real table.
13. Contract S16's five re-score steps match `gradeCaseDefault` step for step: `compatFor(ModeDefault, policy).Apply(tun)` plus `Validate` at `case.go:568-572`, the per-candidate `client.AlbumForID` at `:593`, `cands.Add` at `:604`, `cands.Sorted()` and `matcher.Recommend` at `:612-613`, `matcher.ActionFor` at `:618`. The cited span `:568-622` is exact.
14. `gradeCaseDefault` issues no search and drives no `tag.Album`, and its doc comment at `case.go:517-527` says so; `GradeMode`'s own argument, including the provisional clause Task 2 retires, is at `case.go:141-153`.
15. `gradeClient` builds its replay client with `Unmetered()` at `case.go:330-336`, the precedent contract S17 cites.
16. The budget sits above the transport: `doWithRetry` calls `c.budget.Wait(ctx)` before every attempt at `client.go:346`, whatever `ClientConfig.HTTP` holds. Contract S17's doubling argument is correct.
17. `maxAttempts = 7` at `client.go:42`; the retryable set at `:51-57`; the `default:` arm returning a non-retryable "unexpected status" in one attempt at `client.go:381-382`, so contract S18's 501 refusal costs exactly one attempt.
18. `ClientConfig.Timeout` is dead when `HTTP` is non-nil (`client.go:137-144`), and Task 8 both states this and sets the timeout on the `http.Client` it builds.
19. `NewClient` refuses a missing `UserAgent` or `Budget` at `client.go:124-130`; `ClientConfig`'s doc comment specifies the User-Agent composition verbatim at `client.go:86-92`.
20. Request URLs are built through `url.Values.Encode` at `client.go:183-200`, so a recorded `RequestURI()` is already in `cassette`'s canonical form for both the two-key lookup and the three-key search.
21. `MergePseudoRelease`'s second fetch is per candidate (`client.go:320-333`), which is what makes ruling 5's eleven-request worst case right.
22. `musicbrainz.ClassBackground` at `limiter.go:25` with its own comment naming rung 4 at `:16-18`; `DefaultInterval` at `:62`; `NewLimiter` at `:110-115`; `Limiter.Budget` at `:139-145`; `Close` failing every pending and future waiter at `:147-163`; "a limit of zero or less means no cap" at `:130-133`.
23. `Unmetered` at `limiter.go:354-359`, `Budget.Spent` at `:400-404`, `Budget.Name` at `:408`, the `wait` field unexported at `:340`, `spent` guarded at `:342-343`.
24. Resolution R7's mutual-exclusion argument is exactly right. `Budget.Wait` reserves before it waits (`limiter.go:372-381`, lost-update reason `:361-371`), `reserve` returns `ErrBudgetExhausted` without consulting the limiter (`:383-391`), and a closed limiter's `acquire` returns `ErrLimiterClosed` before the reservation can stick. One pass holding one budget cannot produce both branches, and `Spent()` equals the cap exactly.
25. `ErrUnavailable` at `client.go:84`, `ErrBudgetExhausted` at `limiter.go:48`, `ErrLimiterClosed` at `limiter.go:53`, all sentinels reaching `tag.Album`'s caller wrapped and therefore reachable by `errors.Is`.
26. Both architectural walks cover `internal/engine/shadow` from the moment it exists: `TestEngineNeverImportsHarness` at `arch_test.go:141-146` with `checkNoHarnessImports` at `:30-64`, `TestEngineNeverReadsTheClock` at `:152-157` with `checkNoTimeNow` at `:75-120`, the nested-descent proof at `:194-206`, and both walkers skipping only `testdata`. `_test.go` files are inside both walks, which the plan states.
27. The FLAC mapping table's seventeen Vorbis rows resolve line for line in pinned mediafile 0.17.0: `TITLE:392`, `ARTIST:398`, `ALBUM:410`, albumartist `:542-544`, track `:463-464`, disc `:480-481`, disctotal `:488-490`, comp `:535`, label `:574-575`, catalognum `:604-606`, barcode `:616-620`, country `:659`, media `:673`, albumdisambig `:680`, `MUSICBRAINZ_TRACKID:764`, `MUSICBRAINZ_ALBUMID:782`, date `:685-693` with the `YEAR` fallback at `:690`. `length` is a property over `mgfile.info.length` at `:892-895` with no `MediaField`, and a grep of the whole package for `data_source` returns nothing, so the two "none" rows are right.
28. The precedence rule's mechanism is correct: `MediaField.__get__` walks styles in declaration order and breaks on the first truthy value (`mediafile/fields.py:48-54`); `StorageStyle.fetch` takes `mutagen_file[key][0]` (`storage/base.py:77-84`); `ListStorageStyle.get` returns the first list element (`base.py:154-163`). Only the "or zero" clause is wrong (m9).
29. `_get_date_tuple`'s behavior is stated correctly and cited exactly: the `T`-or-space strip, the `-`/`/` split, the truncate-or-pad to three, and the year-style fallback only when the first component is falsy, all at `mediafile/fields.py:156-172`. `DATE=1994-05-01` yields 1994 and `DATE=` with `YEAR=1994` yields 1994.
30. `safe_cast`'s numeric rules are cited exactly: the leading `[+-]?[0-9]+` with 0 on no match at `utils/type_conversion.py:15-27`, and `bool(int(val))` with `ValueError` falling to false at `:29-34`. `TRACKNUMBER=3/12` is 3, `TRACKNUMBER=A` is 0, `COMPILATION=yes` is false.
31. `matcher.ItemFields` is nineteen fields at `info.go:124-152`, the field-name constants unexported at `:156-176` (which is why `itemBag` must spell them), `ProjectItemFields` at `:189-211` with the typed-default rule at `:178-188`, the two-decoders-drift warning at `:148-150`, and `DataSource`'s None-versus-"" note at `:115-123`.
32. `intField` and `floatField` accept both `int` and `float64` (`info.go:237-259`) and `boolField` accepts only `bool` (`:229-232`), so `itemBag` round-trips through `ProjectItemFields` provided `comp` is emitted as a Go bool.
33. `Candidate.Info.AlbumID` (`candidate.go:12-16`, `info.go:57`), `Distance.Distance()` (`penalty.go:96`) and `Distance.NonzeroKeys()` (`penalty.go:166`) all exist.
34. `matcher.ActionFor` at `verdict.go:179-210` mirrors beets' `_summary_judgment` quiet branch at `beets/ui/commands/import_/session.py:335-341`, so contract S3's binary projection is exact.
35. `cassette.Case` at `:41-58` with `Profile` at `:44` and `EvalDate` at `:45`, `cassette.Request` at `:69-73` field for field and tag for tag with the plan's `RecordedRequest`, `Expected.ConfigHash` at `:80`, `Expected` a nil-able pointer at `:55`, `LoadCase` at `:132` with its validation at `:152-176`, `checkResponseReadable` at `:224-229`. `filepath.Join(dir, name)` makes `responses/001.json` a legal response reference.
36. `corpus/cases/admit-duplicate-candidate/case.json` carries `eval_date` `2026-09-03` and `profile` `estate`, confirming R8's two format choices.
37. `analysis` surfaces are all where the plan says: `Pending` at `:18`, `AgentAdjudicator` at `:30`, `AdjudicationSubject` at `:32-39`, `AdjudicationEntry` at `:46-67`, `CheckAdjudications` exported at `:138`, `checkAdjudicationEntry` unexported at `:158`, the shadow-subject branch at `:168`, the five non-empty prose checks at `:176-190` with `Citation` correctly absent, the agent-disposition rule at `:198-200`, `readJSON`'s `DisallowUnknownFields` at `ledger.go:19-24`.
38. Contract S20's claim that `caseIDs` may be `nil` holds: `known` is read only inside the `CaseID` branch at `adjudication.go:164-167`.
39. `corpus/analysis/adjudications.json` holds 21 rows, every one a case subject with no `album` or `release_mbid`, counted mechanically; the file is two-space indented with a trailing newline. Task 10's two new rules cannot red the gate.
40. `internal/harness/corpus_test.go:377` runs `analysis.CheckAdjudications` in the gate, so Task 10's new rules are exercised against all 21 rows.
41. The incumbent's report shape is mirrored field for field and tag for tag: `AlbumOutcome` at `importer.go:134-150`, `RunReport` at `:156-172`, `toReport` at `:231-241` not copying `stagedDir` (`:213`), the outcome vocabulary at `:107-125`, `LibraryPaths` populated only for `imported` at `:129-133`, the `Error` field's "a beets-database match that never turned up" at `:146-149`.
42. `SourceDir` is uploader-level at all three assignment sites and `albumUnit.stagedDir` is the only per-album path, so the basename join is required and a `SourceDir` join would be ambiguous. Only the line numbers are wrong (m1, m2).
43. `stageAllContributors` stages every quiescent uploader in one run, quiescence checked per uploader inbox at `:790-798`; `stageUser` preserves each inbox entry's own name into staging at `:844-856`; one `beet import` runs over the whole staging tree at `:1100`, so `filepath.Base` of a `skip ` line equals the captured basename.
44. `LibraryPaths` is a positional correlation whose count mismatch leaves later units empty (`:1119-1136`), which is why the apply join is by content.
45. `resolveImported` marks a partly-imported unit `OutcomeReview` with `Error: "left in staging after import"` (`:1196-1199`), with the "one shape ... this album-as-a-unit model does not cleanly resolve" doc at `:1180-1184`.
46. `acquireOrRecordHeld` writes a `LockHeld` report, pings success and returns nil without importing (`:405-413`), which is risk 3's whole mechanism.
47. `sweepReports` runs on every pass (`:389`) and keeps the newest `retention` by ascending mtime (`:540-550`); `defaultReportRetention = 30` (`_salvage/internal/config/config.go:172`), `defaultRetentionDays = 60` (`:155`), `defaultQuiescentMin = 10` (`:152`). R6's rejection of mtime ordering is right, since the incumbent's own sweep already keys on it.
48. Resolution R9's review-tree grammar `^B(-\d{6}(\.\d+)?)?$` matches `uniqueDestPath` exactly (`_salvage/internal/review/review.go:76-100`), `filerFor` builds `<ReviewDir>/<date>/<uploader>` at `importer.go:499-505` with `dateStr` the run's `StartedAt` at `:271`, and `windrose-live` correctly fails to match.
49. `review.Move` tries `os.Rename` and falls back to copy-then-remove on EXDEV (`_salvage/internal/review/move.go:24-40`), which is the partial-read hazard contract S8 names.
50. `config/beets-config.yaml` sets `move: yes` and `write: yes` at `:14-15`, the three import keys at `:16-18`, and `group_albums: yes` at `:23` with the estate's 2026-08-31 reason at `:19-22`. The `-q` flag is at `_salvage/internal/beets/beets.go:87`, the atomic tag-log write at `:104`, `parseSkips` at `:115-135`.
51. `corpus/profiles/estate.yaml:17-19` carries the three import keys, and `mergeProfile`'s `estate` arm applies exactly those.
52. `internal/atomicfile`'s merit claims all hold: 42 lines, one exported function at `:15-41`, temp file in the target's own directory at `:16-17`, `defer os.Remove` at `:22`, `Chmod` after `Close` and before `Rename` at `:35-38`; its test is 95 lines.
53. `cmd/dubplate-server` shapes: `version` at `main.go:29-31`, the single `AddCommand` block at `:48-52`, `newVersionCmd` printing one line, the `shadow` verb reserved at `main.go:11`, `newCheckCmd` at `check.go:16-32`, `newDoctorCmd` at `doctor.go:23-61`, both with `SilenceUsage`/`SilenceErrors` and `cmd.SetErr(stderr)`, `printRow`'s pinned `"%-32s %s"` at `doctor.go:63-67`, `thresholdStatus`'s unconditional `ParseInt` at `:100-107`. Calling `time.Now()` in `cmd/` is legal, since both walks root at `internal/engine`.
54. `internal/config`: `Config`'s four fields at `config.go:56-65`; the single-choke-point doc at `:67-71`; `wall.get` treating empty as unset at `:92-98`; `rejectNamespace` at `:133-152` with `ownedNow` its only exemption at `:135`; `Load` at `:159-175` in its report-everything shape.
55. `Threshold` is three string fields at `thresholds.go:13-17`; `Register` is a bare append with no dedup at `:28-30`; `ParseInt` is the only parser at `:43-60`; the "explicit setup call rather than an init function" rule at `:24-27`; production registers none.
56. A `ThresholdKind` backed by `int` keeps `thresholds_test.go`'s comparison compiling.
57. `.gitleaksignore` is the single line pinning a finding to `internal/config/keys_test.go` at line 66; `keys_test.go` is 123 lines; line 66 is the first row of `TestReservedClassOf`'s table, whose rows run `:66-72` inside a literal declared `:60-73`. Appending at end of file is below the pin, and `keys.go` edits shift nothing in `keys_test.go`. The plan quotes the fixture literal nowhere, and `gitleaks detect --no-git` over the plan file reports no leaks.
58. No `DUBPLATE_SHADOW_*` name collides with a reserved class or a superseded key; `DUBPLATE_REVIEW_DIR` is reserved for rung 13 at `keys.go:130` and is distinct from `DUBPLATE_SHADOW_REVIEW_DIR`; `DUBPLATE_BEETS_DB` stays retired at `:178`.
59. `go.mod` declares `go 1.27.0` with cobra as the only direct requirement, so `cmp.Or` is available and Task 8's `cmp.Or(cfg.Transport, http.RoundTripper(http.DefaultTransport))` compiles and cannot panic.
60. `internal/engine/tag/tag_test.go:25-51`'s `taggingServer` keys on `r.URL.Path` rather than the full request URI, and `statusServer` at `:52-64` sends `Retry-After: 0`, so `releaseServer`'s stated design and the 31-second claim both hold; `retryWait` prefers a delta-seconds `Retry-After` at `client.go:404-412` and `backoff` is urllib3's schedule at `:415-423`.
61. `CaseGradeSummary`'s "what one sweep actually compared, never what it visited" at `case.go:109-110`.
62. `gradeCaseVerdict` marshals the pasted row with `json.MarshalIndent(row, "", "  ")` at `case.go:478`, the indent Task 11 matches.
63. `gradeTunables` is on the harness side of the arch wall (`defaults.go:72-79`), which is why `QuietImportTunables` must live in `matcher`.
64. All nine contested-surface rows exist in `docs/superpowers/specs/2026-09-03-parallel-build-approach.md` as the plan's table claims: shadow `:180`, atomicfile `:181`, `cmd/dubplate-server` `:182`, `internal/config` `:183` with the `.gitleaksignore` rider, harness `{cassette,stub,analysis}` `:169`, `internal/engine/{matcher,musicbrainz}` `:171` naming `QuietImportTunables` "and nothing else", the state directory layout `:178`, the adjudications schema `:179`, `docs/` ledgers `:177`, and 4b's row `:184`; `internal/harness/conform` is covered by `:167`.
65. The `internal/harness/conform` edit is assigned to 4a by the conductor's own fold paragraph at `spec:59`, so Task 2's two small edits are not unowned.
66. The staleness pre-task's expected answers match `docs/STATUS.md`: lane-1's next action at `:24-26`, lane-2 BLOCKED and the doc lane IDLE at `:29`, the 2026-09-07 09:00 flag route and the 15M window with about 7.6M spent at `:27-28`.
67. The rung 3 to rung 4 handoff is at `docs/HISTORY.md:267-302`, its wire-pace item 3 at `:287-294`, and its "all five prose fields non-empty" wording at `:295-302`, which the plan correctly declines to inherit since `checkAdjudicationEntry` does not check `Citation`.
68. The ceiling arithmetic reconciles: 3d's close records prep 2.51M, Task 0 0.09M, the task chain 2.05M over nine tasks, simplify 0.15M, its review 0.07M, the runner review 0.14M, the close 0.15M, "about 5.15M against the 6M ceiling" (`docs/HISTORY.md:474-483`); 15 tasks at 230K is 3.45M, plus two upshifts, plus 0.2M, plus 0.25M, plus 1.5M prep is 5.9M; 15M minus 7.6M is 7.4M, leaving 1.5M.
69. Deliverable counts are stated per task and none exceeds four; the pass is sixteen entries, past the roughly-eight bar, and the plan names its split point at the Task 9/10 boundary with a three-reason argument and a numeric trigger, which is what CLAUDE.md's pass-sizing rule asks of an orchestrator.
70. All six owner rulings carry a provisional application, an inline marker, and a one-sentence fold; ruling 6 correctly records that 4a is unaffected either way.
71. The plan carries no placeholder, no `TBD`, and no unresolved cross-reference; the four `TODO` occurrences are all references to the string Task 11 forbids in a staged row.
72. The three spec defects the plan reads past are real: `spec:1608` does say "(ruling 7)" where ruling 6 is meant, `spec:299-301`'s one-request-per-second sentence is superseded by the M6 fold at `:328-330`, and `spec:1824` does read as a gate the conductor's own fold paragraph at `:59` has already lifted.
73. `internal/harness/analysis`'s inline-table precedent for a per-row rule is real (`adjudication_test.go:10-113`), and `testdata/` holds cross-ledger fixtures only, so R14's choice is the package's own.
74. `os.ReadDir` returns entries sorted by filename, which `listContributorDirs` already relies on (`importer.go:1238-1254`), so R18's capture-walk order is correct as stated. Only its consumption by the match step is wrong (B3).
