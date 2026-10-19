# Rung 2, Pass 2d Plan: Adversarial Review

Fresh-context adversarial review of
`docs/plans/2026-09-03-rung-2d-file-level.md` (revision 1,
2026-09-03), read against the 2d planning inventory
(`docs/research/2026-09-03-rung-2d-planning-inventory.md`),
the pre-ruled findings M3, M4, M5, m2 and M8 in
`rung-2c-plan-review.md`, the oracle package as shipped
(`internal/harness/oracle/`), and ladder entry 8 of the build spec
(`docs/specs/2026-09-02-greenfield-server-build.md:666-673`).

**Counts: 1 BLOCKER, 5 MAJOR, 7 MINOR.**

Every claim below about beets, `mediafile`, or `mutagen` behavior was
measured against the pinned venv
(`tools/venv/.venv`, `beets==2.13.1`, `mediafile==0.17.0`,
`mutagen==1.48.1`) in a scratch directory outside the repository. No
file under `corpus/` was read or written, no derivation was run, and
nothing in the working tree was touched. The concurrent capture
workflow's outputs are cited by role only.

**On the pre-rulings.** M4 and M5 are honored faithfully, and the plan
strengthens both. M3 is honored in shape and undercut in substance by
the BLOCKER below, which is a defect in the artifact M3 asked for
rather than a re-opening of M3. m2 is honored in its Files list and
inverted in its rationale (MAJOR 2). M8's formula is carried and half
its finding is dropped (MAJOR 3). No finding here asks the conductor to
re-litigate a ruling.

---

## BLOCKER

### B1. The ledger's value-key contract is `readable_fields()`, which carries nine read-only audio properties, and Task 4 hands them to rung 8 as conformance targets

**Section:** "How a file-level record gets derived", step 5; Task 1
("The exact `_media_fields` intersection is computed and recorded");
Task 4 (the rung 8 handoff statement).

**Defect.** The plan defines the ledger's value-key contract as
`Item._media_fields`, which beets computes as
`set(MediaFile.readable_fields()) & _field_names`
(`library/models.py:731`). Measured against the pinned venv, that
intersection is 78 keys. Nine of them are not tags at all:

```
bitdepth, bitrate, bitrate_mode, channels, encoder_info,
encoder_settings, format, length, samplerate
```

These are `MediaFile`'s read-only properties, the difference between
`MediaFile.readable_fields()` and `MediaFile.fields()`. All nine sit
inside the intersection. `MediaFile.update()` iterates
`sorted_fields()`, which is the writable set only, so `Item.write()`
never writes any of them. A read-back through `MediaFile(path)`
nonetheless reports all nine, and their values come from the vendored
template blob's `STREAMINFO`, not from the case.

Measured on a 42-byte pure-`mutagen` template of the shape Task 1
mandates, a subject row would carry `format="FLAC"`,
`samplerate=44100`, `bitdepth=16`, `channels=1`, `length=0.0`,
`bitrate=0`. Those six values are identical across all twelve subjects
and independent of every case. Worse, `length=0.0` directly
contradicts the same item's `AppliedFields.length` (a real duration in
seconds), so the ledger's own row disagrees with the record it claims
to be a derivation of. In a probe that set `length=123.456` on the
item before `write()`, the read-back reported `0.0`.

Task 4's handoff statement then tells rung 8 that
`corpus/analysis/file_tags.json` "is the conformance target for
pure-Go FLAC tag writing" and names "the exact intersection, by name"
as what rung 8 may assume. Ladder entry 8's deliverable is "pure-Go
FLAC tag writing... proven against the corpus's expected applied tag
sets from rung 2". A rung-8 implementer reading the handoff literally
would build a Go writer that tries to reproduce nine values no tag
writer produces, and would discover the contradiction only by
re-deriving the study 2d exists to spare it. This is the pass's single
stated purpose, so the defect ships into the rung it was built for.

**Concrete fix.** Split the contract in Task 1 and carry the split
through Tasks 2 and 4.

- Compute both sets: `MediaFile.fields()` (writable, 69 keys inside
  the intersection) and the nine-key read-only remainder. Take them
  from the pinned venv rather than hardcoding either.
- Make the ledger's per-row value dict the writable set only, and give
  the nine read-only properties their own place if they are wanted at
  all: one `template` block recorded once for the whole ledger, since
  their values are a property of the blob and never of a subject.
- Have the gate-resident checker assert the negative directly: no row
  carries any key from the read-only remainder. That check has teeth
  the plan's current bookkeeping-column assertion does not, because
  the nine keys really are inside `_media_fields` today.
- Rewrite Task 4's handoff sentence to name the writable set as the
  conformance target and to state the nine exclusions by name, with
  the reason (`update()` iterates `sorted_fields()`, so a writer
  cannot write them and a reader gets them from `STREAMINFO`).
- Record both key lists in `corpus/README.md`, distinguished, rather
  than one 78-key list labelled "the ledger's value-key contract".

---

## MAJOR

### M1. `derive_file_tags.py` is given no config isolation, and `Item.write()` resolves `beets.config`

**Section:** "How a file-level record gets derived", step 1; Task 1
("`derive_file_tags.py` implements the six steps above exactly").

**Defect.** Step 1 enumerates what the new shim does not do: "It
starts no stub, issues no request, renders no profile, and loads no
plugin, matching `derive_component.py`'s minimalism." That reads as a
complete isolation list, and it omits the one isolation step that
matters most for a write. `Item.write()` calls
`beets.config["id3v23"].get(bool)` (`library/models.py:961-962`) and
fires `plugins.send("write", ...)` (`models.py:970`) before it opens
the file. Resolving `beets.config` makes confuse materialize the
config, which reads the runner's own user config directory unless
`BEETSDIR` is set first.

Both existing shims set it. `derive_oracle.py:124-132`'s
`prepare_config` docstring states the reason in as many words:
"reading the workstation's own ~/.config/beets would silently change
every verdict." `derive_component.py:448-452` sets an *empty*
`BEETSDIR` with its own comment ("An empty BEETSDIR is what puts the
run on beets' shipped defaults"). The plan cites
`derive_component.py:451` for its `TemporaryDirectory` pattern and
never mentions the `BEETSDIR` assignment on the adjacent line.

The failure is machine-dependent, which is what makes it dangerous.
This workstation's `~/.config/beets` is an empty directory, so the
derivation would pass here and the defect would ship green. A stranger
with a real beets config gets a different `id3v23` resolution, or a
`ConfigTypeError` out of `.get(bool)` if their config sets a non-bool,
and `TestFileTagsLedgerMatchesTheCorpus` fails with nothing pointing
at the cause. CLAUDE.md's stated primary product goal is setup by a
stranger, and the plan's own constraint says every expectation is
derived by executing pinned beets, not pinned beets plus whatever the
runner happens to have in `~/.config`.

**Concrete fix.** Add a seventh step to the "How a file-level record
gets derived" contract: `derive_file_tags.py` sets `BEETSDIR` to an
empty scratch directory before importing beets, matching
`derive_component.py:448-452`, so the write runs on beets' shipped
defaults. Add it to Task 1's behavior contract as a named deliverable
with a proof: assert in the derivation that `os.environ["BEETSDIR"]`
points inside the scratch directory at the moment `Item` is imported.
`derive_component.py`'s comment is the register to match.

### M2. The `Reconcile` pairing rationale is inverted, so the prescribed test proves the loud direction and leaves the silent one uncovered

**Section:** Task 1 ("`Reconcile` gains its sixth component, on both
sides").

**Defect.** The plan states the failure mode as: "A probe that reports
five keys against a `pinned` tuple expecting six compares against a
missing key and says nothing, which is the silent failure m2 named."
Read against `reconcile.go:50-58`, that is backwards. The loop
iterates `pinned`. When `pinned` names `mediafile` and `installed` has
no such key, `installed["mediafile"]` is the empty string, the
`p.manifest != installed[p.component]` branch fires, and `Reconcile`
returns `mediafile: manifest pins 0.17.0, venv has `. That is loud.
`versionProbe` uses `md.version(n)`, which raises
`PackageNotFoundError` on a genuinely absent package, so that path is
loud too.

The silent direction is the other one. Adding `mediafile` to the
`versionProbe` string while leaving the `pinned` tuple at five means
the probe reports six keys, the loop compares five, and the sixth is
never examined. That is the failure m2 actually named, and it is the
one the plan's prescribed test ("Prove the pairing with a test that
fails when only one side is edited") will not find, because an
implementer following the stated rationale will write the
five-probe-versus-six-pinned case, watch it fail loudly, and call the
pairing proven.

**Concrete fix.** Correct the sentence to name the silent direction:
a `versionProbe` naming a component the `pinned` tuple omits is never
compared. Require the test to cover that direction specifically, by
driving `Reconcile` against a `pinned` set and a probe output that
disagree in membership rather than in value, and require the task to
report which direction its mutation exercised. The reverse direction
is worth one assertion too, since it is cheap, but it is not the one
that needs proving.

### M3. Task 3's wall-time section carries M8's formula and drops half of M8's finding

**Section:** Task 3 ("The wall-time section uses the M8 formula at the
corpus's real size").

**Defect.** M8's defect paragraph names two costs. The first is the
per-case floor, `1.04s + 0.25s × (requests − 1)`, which the plan
carries correctly and repeats in its pre-ruling summary. The second is
the sentence that follows it: "`derive_ties.py` starts its own stub
and pays again." That is verifiable in the shipped code:
`DeriveTies` calls `stub.New` at `ties.go:65` and runs its own Python
process at `ties.go:101`, entirely separately from `Derive`'s stub at
`oracle.go:163`. A full `make derive` therefore replays every case's
requests twice and pays the inter-request floor twice.

Task 3's behavior contract mentions only the per-case formula and "the
committed case count and the per-case request-count distribution". A
conscientious implementer computes one corpus-wide pass and writes it
down. The differential doc is then declared the single home for that
number, and the re-capture doc is told to cite it rather than restate
it. One number with one home is the right rule, and it makes an
understated number worse rather than better, because the understatement
propagates with the citation.

**Concrete fix.** State in Task 3's contract that a full derive pays
the floor once in the corpus pass and again in the ties pass, cite
`ties.go:65` and `ties.go:101` as the evidence, and require the
wall-time section to present the corpus leg, the ties leg, and the
total separately. The plan's own "Derive wall time, budgeted" section
already asks every derive report to break down corpus, ties, and
component legs, so the doc should be held to the same shape as the
reports that feed it.

### M4. Task 3's independence buys nothing on either budget and costs the gate

**Section:** Execution model ("Task 3 is INDEPENDENT of Tasks 1 and
2"); the "one honest caveat" paragraph.

**Defect.** The plan justifies parallelism with disjoint file sets and
then concedes the real coupling: "parallel tasks share a working tree,
so Task 3's gate run may observe Task 1's in-flight changes." That
concession is larger than the plan treats it.

`bash scripts/check.sh` runs `make check`, which is `go vet ./...`,
`golangci-lint run ./...`, `go test ./...`, and `tidy-check` over the
whole module, plus `ruff check` and `ruff format --check` over every
tracked `*.py`. While Task 1 is mid-edit on `cassette.go`,
`reconcile.go`, `file_tags.go`, and `derive_file_tags.py`, none of
those can be green. Task 3's stated close is "Gate green; commit", and
its Global Constraint is "Gate green and clean `git status` at every
task end". Under parallel execution, both are unreachable for whichever
task finishes first. The plan pre-excuses the failure by reassigning
it ("a gate failure in Go or Python during Task 3 belongs to Task 1's
report"), which leaves Task 3 committing into a tree it cannot
demonstrate is green. CLAUDE.md's rule is "Never commit red", without
an attribution clause.

The `diff-reviewer` half is worse, and the plan does not mention it. The
in-chain reviewer reads the task's diff. In a shared worktree with two
live tasks, the working diff carries both, so Task 3's reviewer is
handed Task 1's half-written Go and Task 1's reviewer is handed two new
markdown files it has no acceptance criteria for. That is contention,
in the plan's own sense of the word, and the disjoint-file-sets
argument does not address it.

Against what does the parallelism pay? CLAUDE.md names two budgets,
tokens and attended time, and states plainly that clock time is not one
of them: "prefer the serial, cheaper path." Running Task 3 concurrently
saves neither budget. It saves wall time the conductor is not spending
attended anyway, and it spends a re-dispatch the first time a reviewer
gets a mixed diff.

**Concrete fix.** Drop the independence claim and run Task 3
sequentially, in whichever slot is convenient: before Task 1, since it
depends on nothing from the machinery track and its M8 arithmetic
depends only on the corpus 2c froze. Ordering it first also removes
the placeholder handoff, because Task 4 fills the file-tags line
either way. If the conductor still wants it concurrent, then say so
explicitly and change its acceptance to match reality: no full gate
run, a named markdown-only bar, no commit of its own, and its files
folded into the conductor's batch commit after Task 1 lands.

### M5. The ceiling's re-derivation point is placed after the task that says least about the pass's largest one

**Section:** "Token ceiling: 2.5M"; Task 4.

**Defect.** The ceiling is 2.5M for four tasks, re-derived after Task
1 reports its spend. Task 1 is machinery: one binary blob, one Python
shim, one Go driver, three small Go edits. Its spend predicts Task 2's
reasonably well and predicts Task 4's hardly at all.

Task 4 is the pass's largest read by a wide margin. It rewrites
`corpus/README.md` over the whole committed corpus, "One inventory
covering every committed case, each row naming its id, stratum, family
or axis, profile, and the behavior it exercises", with the case total
read off the tree. 2c's own close projection puts that total near 164.
Naively executed, that is 164 `case.json` reads pulled into context,
each carrying an `Input`, an `Expected` record with candidates,
weights, thresholds, and `AppliedFields`, plus a `claims.json`. Task 4
also runs two full derives, drafts HISTORY and STATUS, and writes the
rung 8 handoff. It is entirely plausible for Task 4 alone to approach
the plan's 625K per-task average several times over, and the plan's
re-derivation point gives the conductor no warning before it starts.

This is the shape of 2c's own M9 ("The 8M ceiling is under-derived,
and its largest unmeasured task runs before the re-derivation point"),
re-instantiated with a different largest task. The plan's basis
paragraph asserts that 2d's four tasks "are all of the machinery,
prose, and close character that ran above that average in 2b" without
citing a single per-task figure from 2b or 2c.

**Concrete fix.** Two changes, both cheap.

- Re-derive the ceiling after Task 2 as well as after Task 1. Task 2
  is the first task that touches the corpus at scale, so it is the
  first honest signal about Task 4.
- Constrain how Task 4 builds the inventory. Require a mechanical
  extraction of `id`, `profile`, `eval_date`, stratum, and family from
  the committed cases into a compact intermediate, and require the
  prose to be written from that intermediate. Reading 164 full case
  records into context to write a table of five columns is the
  avoidable half of the cost, and the plan already forbids the
  analogous waste elsewhere.

---

## MINOR

### m1. The value-key contract has no machine-readable home

Task 1 records the intersection in `corpus/README.md` prose. Task 2's
gate-resident checker must assert "every value key sits inside the
recorded `_media_fields` intersection", in pure Go, with no venv. The
Go checker therefore needs its own copy of the key list, and the plan
never names where it lives or which copy is authoritative. That is two
hand-maintained copies of one derived fact, which is precisely the
failure the plan holds Task 3 to avoid ("Two independently maintained
copies of an arithmetic result is how one of them goes stale").
**Fix:** have the oracle-tagged derivation emit the key list into
`file_tags.json` itself as a header block, have the Go checker read it
from there, and let the README quote it as documentation rather than
as the contract.

### m2. The 2c selection dossier, Task 2's one hard external input, is never cited by path

Task 2 must read "2c's Task 3 selection dossier via its
file-tag-subject flag" and stop if the flagged set is not exactly
seven. Every other source in the plan carries a `path:line` citation.
A zero-context subagent has to hunt for this one. The dossier is
`docs/research/2026-09-03-real-catalog-selection.md`; the
row format is declared at its lines 84-88 and the flag is column 6,
"File-tag subject". Counted directly, it carries exactly seven `yes`
rows: `real-classical-karajan-beethoven-89`,
`real-reissue-nevermind-2011`, `real-boxset-james-brown-star-time`,
`real-nonlatin-utada-first-love`, `real-noise-dylan-bootleg`,
`real-genre-public-enemy-nation-millions`,
`real-decoy-beethoven-symphony-5`. **Fix:** pre-extract all seven ids
into the plan, keep the stop condition, and reduce Task 2's job to
confirming each id resolves to a committed case.

### m3. Task 3's two deliverables get no gate coverage at all

Task 3's contract says `bash scripts/check.sh` "is the whole bar".
Read against the gate as shipped, that bar does not touch Task 3's
files. `scripts/vale-comments.sh` lints Go comment prose only, ruff
covers `*.py`, shellcheck covers `scripts/*`, and nothing looks at
`docs/*.md`. "Gate green" for Task 3 means "Tasks 1 and 2 did not
break anything", which is not a statement about Task 3's work.
**Fix:** say so plainly in the contract, and name the real bar: the
`diff-reviewer` pass against the acceptance criteria, plus the
`writing-voice` standard the two docs are written to.

### m4. `-count=1` is demanded of `make derive`, which has no way to accept it

The Global Constraints require forcing `-count=1` "for any derivation
reported as evidence". `make derive` is a fixed recipe
(`go test -tags oracle -timeout 90m ...`) with no variable hook, and
the Makefile appears in no task's Files list. **Fix:** name the actual
invocation, `GOFLAGS=-count=1 make derive`, or authorize the one-line
Makefile change under Task 1.

### m5. The rebuild from `AppliedFields` carries no faithfulness assertion

Step 2 rebuilds beets items by assigning `AppliedFields` values
straight onto a fresh `Item`. `AppliedFields` is a `jsonable()`
projection whose documented last resort is `str(value)`
(`derive_oracle.py:105-121`). A field that took that fallback would be
written into the FLAC as a stringified repr, and nothing in the plan
would notice. Measured against the pinned venv, all 78 intersection
fields are backed by primitive beets types (String, DelimitedString,
Integer, PaddedInt, ScaledInt, Boolean, NullFloat, DurationType,
MusicalKey), so the fallback cannot fire today. **Fix:** assert it,
cheaply, so it stays true: the derivation fails if any
`_media_fields` value in a subject's `AppliedFields` is not a JSON
primitive or a list of them.

### m6. `corpus/README.md` is written three times, and Task 4's instruction contradicts itself

Task 1 writes its manifest and file-tags sections, Task 2 writes its
regeneration section, and Task 4 then performs "the whole-corpus
README rewrite". Task 4's own sentence calls that rewrite "an
extension rather than a replacement", which is the opposite
instruction. **Fix:** pick one. Either Task 4 extends a file Tasks 1
and 2 already shaped, in which case drop the word "rewrite", or Task 4
authors the file whole, in which case Tasks 1 and 2 should write their
material into a staging note and stop editing the README.

### m7. "Owner touchpoints: none" sits next to three mandated hard stops

The header states the pass expects no owner contact. Task 1 carries a
named escalation trigger, and Task 2 carries two stop conditions (an
unresolvable case id, a flagged set whose size is not seven). Each
resolves to the conductor rather than to Geoff, which the plan never
says. **Fix:** one sentence stating that every stop in this pass
returns to the conductor, and that only a conductor decision escalates
to Geoff.

---

## Verified sound

Checked and holding, so no later reader spends the same effort twice.

**The pure-`mutagen` template route works, and the plan's stated bar is
the right one.** Constructing a `mutagen.flac.StreamInfo(None)`, setting
44100/mono/16-bit with `total_samples=0`, and prefixing the encoded
block with the `fLaC` signature and a last-block header yields a
42-byte file. `mutagen.flac.FLAC()` opens it, and `MediaFile()` opens it
and reports `type(mgfile).__name__ == "FLAC"`, which is exactly the
dispatch the plan cites at `mediafile/__init__.py:174`. No `ffmpeg` and
no `flac` binary is involved. M4's ruling is implementable as written,
and Task 1's escalation trigger will not fire on this route.

**The write-then-read-back round trip is deterministic.** Two
independent runs against fresh copies of the same template, with the
same item fields, produced byte-identical FLAC files and identical
read-back dicts across all 78 intersection keys. The determinism
worry the review brief raised about generated fixtures does not
materialize here, and the inventory's reasoning about why is correct:
the blob is committed once and never regenerated, so only the write
step needs to be deterministic, and it is.

**The read-back values are JSON-clean.** No value in a measured
78-key read-back was outside `str`, `int`, `float`, `bool`, `None`, or
a list of strings. Nothing needs `jsonable()`-style coercion on the
read-back side, and the `-update-ledgers` template the plan adopts
round-trips them: `decodeAny` on both sides of
`TestComponentTableReDerivesToItsCommittedTable`
(`component_derive_test.go:61`) compares decoded values, so Go's
integer-versus-float JSON formatting cannot produce a false
divergence.

**`Item.write()` returns `None`.** Confirmed at
`library/models.py:934-939` and by execution. The plan's insistence
that an independent read-back is the only signal is not a stylistic
preference, exactly as it says.

**The seven dossier flags are exactly seven.** Counted directly off
`2026-09-03-real-catalog-selection.md`: 108 `real-` rows, 7 carrying
`yes` in the File-tag subject column. Task 2's stop condition will not
fire on the dossier itself. Whether all seven were captured and
committed by 2c is a separate question the stop condition still
covers.

**The schema bump forces no re-freeze.** `schema_version` appears only
in `Manifest`, in `cassette.LoadManifest`'s validation, in
`corpus_test.go:160`'s `wantManifestSchemaVersion`, and in unrelated
ledger test fixtures. No case record carries it, so the plan's claim
that bumping to `"3"` changes no committed record's content is
correct.

**A committed FLAC blob does not trip the secret scan.** `gitleaks
detect --no-git` over a directory holding both a bare template blob
and a tag-written copy reports no leaks. Task 1 can vendor the blob
without a `.gitleaksignore` entry.

**M5 is honored, and strengthened.** The plan states the invariant as
the pass's acceptance test, requires a full `make derive` as the
evidence for it in Tasks 1 and 2, and forbids the corpus derivation
path from ever giving an item a non-empty `path`. Nothing in the
described design reaches `oracle.Derive`.

**The single-invocation shim shape is right.** `DeriveComponentTable`
calls `runPython` with no arguments (`component.go:16`), and
`runPython` sets `PYTHONHASHSEED=0` (`ties.go:129`), which the
file-tags derivation inherits for free. One process for twelve
subjects, keyed by case and item index, matches the
`components.json` precedent rather than `derive_oracle.py`'s per-case
shape.

**The derive wall-time budget reconstructs.** 2c's Execution model
projects "roughly 9-10 minutes per task by the loop's end and at Task
20's determinism re-derivation". 2d budgets four full runs (Tasks 1,
2, and two in Task 4) at that rate for 40 minutes, and states 40 to 50.
The Makefile's `~585s` comment landed mid-2c and is consistent with the
same projection. The arithmetic is sound and the 90m per-run timeout
has headroom.

**The gate-resident half is genuinely gate-checkable offline.**
`corpus_test.go` already loads every committed case through
`cassette.LoadCase` and already wires five ledger checks the same way
(`TestProximityLedgerAcknowledgesEveryFlaggedCase`,
`TestTieLedgerCoversEveryCorpusCase`,
`TestComponentTableIsWellFormed`, and the two adjudication ledgers).
A sixth in that shape needs no venv, no network, and no new
machinery. `analysis.ComponentRowFloor` (`analysis/component.go:15`)
is a real precedent for the row-count floor the plan asks for.
