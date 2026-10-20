# Rung 2, Pass 2d Planning Inventory

Research report, 2026-09-03, prepared for pass 2d's plan authoring.
Covers 2d's ratified scope per the 2c plan's fold (M10): the file-level
applied-tag derivation (rung 8's stated precondition), the
version-bump differential documented job, the quarterly re-capture
chore doc, and the whole-corpus README rewrite deferred from 2c's Task
20. A capture workflow is running in `corpus/` concurrently; this
report does not read `corpus/cases/`, `corpus/analysis/`, or
`corpus/README.md` for current state, and cites them by role only,
quoting only what the 2c plan and its adversarial review already
quoted from them (both closed, static documents).

Every source claim below is checked against the file cited, not
against another document's paraphrase of it, except where the source
is itself the 2c plan or its review (both frozen text, safe to quote).

---

## (a) What the file-level derivation must produce

**Not a `cassette.Expected` field.** The 2a plan settled this at
authoring time, and 2c's blocker M5 and the 2c-plan-review's
"Verified sound" section both re-confirm it: `AppliedFields` is the
field-level record of beets' in-memory `apply_metadata()` step, proven
to need no files, no importer, no library
(`docs/plans/2026-09-02-rung-2a-cassette-substrate.md:205-211`).
`derive_oracle.py`'s own docstring on `applied_fields()` states the
boundary explicitly: "This is the field-level record only. What
`item.write()` would put in a file is pass 2c's deliverable"
(`internal/harness/oracle/derive_oracle.py:214-216`) — 2d, after the
M10 fold. The corpus derivation path (`Derive` in `oracle.go`) never
gives an item a non-empty `path`
(`internal/harness/oracle/derive_oracle.py:109-111`'s `jsonable`
docstring: "the corpus is file-free, so the only path seen here is
the empty one"), and the file-level derivation must not change that:
M5's invariant is that `make derive` re-derives every committed record
byte-identically after 2d's change.

**The artifact is a second, independent derivation entry point in the
shape of `derive_component.py`** (M5's concrete fix, restated in the
2c plan's Task 20 carry-forward note,
`docs/plans/2026-09-03-rung-2c-catalog-adversarial.md:1034-1042`):
a script that reads a *committed record's* `AppliedFields`, builds
beets `Item` objects from those fields (the same `build_items` shape
already in `derive_oracle.py:156-165`), writes each item to a fresh
FLAC copy on disk via `item.write()`, and reads the result back
independently via `mediafile.MediaFile(path)` — never trusting
`item.write()`'s own return value (it returns `None`; see (c) below).
It allocates its own scratch directory (`tempfile.TemporaryDirectory`,
matching `derive_oracle.py:229` and `derive_component.py:451`'s own
pattern) and is never called from `oracle.Derive`.

**Precisely what gets written is a subset of `AppliedFields`, not all
of it.** Verified in the pinned venv's beets source
(`tools/venv/.venv/lib/python3.13/site-packages/beets/library/models.py`):

- `Item._media_fields = set(MediaFile.readable_fields()) & _field_names`
  (line 731) — the intersection of every tag `mediafile` can read and
  every column beets' `Item` model defines.
- `Item.write()` filters to exactly that set before calling
  `mediafile.update()`: `item_tags = {k: v for k, v in
  item_tags.items() if k in self._media_fields}` (line 967).
- `Item.read()` (the read-back precedent `derive_file_tags.py`
  mirrors) opens a fresh `MediaFile(syspath(read_path))` and copies
  every `_media_fields` key back onto the item (lines 904-929) — this
  is the exact mechanism M3 wants exercised and diffed.

So the output ledger (`file_tags.json`) is, per subject case: the case
id, the item index, and a dict keyed by the `_media_fields`
intersection — a strict subset of the record's own `AppliedFields`
keys (bookkeeping fields like `id`, `album_id`, `mtime`, `path` itself
are `Item` columns but not `MediaFile.readable_fields()`, so they
never reach the file and must not appear in `file_tags.json`).
Confirming the exact intersection for the estate's FLAC subjects (via
`MediaFile.readable_fields()` against the pinned `mediafile==0.17.0`)
is 2d's own first-task work, not something to hardcode from memory
into the plan.

**The gate-resident half and the oracle-tagged half, both required**
(M3, folded verbatim into the 2c plan's Task 20 carry-forward note):

- Gate-resident, pure Go, no venv: every subject case named in
  `file_tags.json` is a committed case; every row's pointer resolves
  into that case's `AppliedFields`; the subject set matches a
  committed list. This alone is not enough — M3's defect finding was
  precisely that a structural-only check "cannot go red" against a
  `mediafile` version bump moving tag shapes.
- Oracle-tagged: `TestFileTagsLedgerMatchesTheCorpus`, in the exact
  shape of the two extant precedents —
  `TestTieLedgerMatchesTheCorpus` (`internal/harness/oracle/ties_derive_test.go:54`)
  and `TestComponentTableReDerivesToItsCommittedTable`
  (`internal/harness/oracle/component_derive_test.go:22-64`). The
  latter's shape is the closer template: derive fresh, decode both
  the fresh and the committed JSON to `any` and `reflect.DeepEqual`
  them (not a byte compare, since the committed file is indented for
  review and the shim prints one line), fail with "re-run with
  `-update-ledgers`" on divergence, and honor the package's
  `*updateLedgers` flag to rewrite the committed file. `file_tags.json`
  needs the same `-update-ledgers` regeneration path documented in
  `corpus/README.md`'s "Regenerating a derived ledger" section
  (cited by role in the M3 finding and the m5 finding; its content is
  stable per CLAUDE.md's ledger-format conventions and is not read
  here since the section is not one the concurrent capture workflow
  is editing — 2d's plan should still open the live file itself
  before writing, since this report deliberately avoids that read).

**`Reconcile` and `Manifest` both need editing, and the edit is
already fully specified** (m2, folded verbatim): `Reconcile` lives in
`internal/harness/oracle/reconcile.go` (confirmed: lines 21-63), not
`oracle.go`. Its `versionProbe` names five components today — beets,
unidecode, jellyfish, lap, numpy (`reconcile.go:21-22`, confirmed by
direct read) — not four; 2d's Task adds `mediafile` as a sixth,
editing both the Go tuple literal (`reconcile.go:39-48`'s `pinned`
slice) and the Python `versionProbe` string in the same file, or a
reconcile run silently compares against a missing map key. The
`Manifest.MediaFile` field itself does not exist yet
(`grep -n MediaFile internal/harness/cassette/cassette.go` finds
nothing under the current `type Manifest struct` at
`cassette.go:105-`); it is the one additive change to
`internal/harness/cassette` this pass is permitted (2c's Global
Constraints froze the cassette surface with exactly this one named
exception, `2026-09-03-rung-2c-catalog-adversarial.md:213-218`),
alongside the `schema_version` bump to `"3"`.

**Version to pin: `mediafile==0.17.0`**, already resolved in the
pinned venv as a transitive dependency of `beets==2.13.1`
(`tools/venv/requirements.txt:19`, confirmed by direct read — no new
`pip`/`uv` install needed; the package is already present at
`tools/venv/.venv/lib/python3.13/site-packages/mediafile/`).

---

## (b) Audio fixture generation for FLAC in a venv-only world

**The route is already ruled, not 2d's to choose** (M4, folded
verbatim into Task 20's carry-forward note): vendor one small static
template FLAC blob under `internal/harness/oracle/testdata/`,
committed, with its provenance and byte size recorded in the README.
`make derive` gains no binary dependency; `ffmpeg`/`flac` are not
adopted as prerequisites. The review's stated reason still holds
verified: CLAUDE.md's stated primary product goal is "Ease of setup
by a stranger... no site-specific value, path, or identity default
outside `internal/config`" (repo CLAUDE.md), and
`tools/venv/README.md`'s whole story today is that the venv is the
derivation prerequisite — adding a binary codec toolchain requirement
would be a second, silent one.

**No existing salvage fixture to adopt.** The 2c plan's Global
Constraints section names "Task 14's `_salvage` FLAC-fixture
precedent" as already ruled
(`2026-09-03-rung-2c-catalog-adversarial.md:285-287`), but a direct
search (`find _salvage -iname '*flac*'`, `find _salvage -iname
'*.mp3'`) turns up nothing — `_salvage/e2e/testdata/` holds only Go
stub binaries (`beet`, `rclone`, `sqlite3`), no audio. Read plainly,
that sentence refers to the *ruling itself* (M4's decision) carrying
forward as precedent, not to a real file sitting in `_salvage/` for
2d to copy. **2d authors the blob fresh; there is nothing to adopt.**

**Generation tool: `ffmpeg` and `flac` are both present on this
workstation** (Homebrew/linuxbrew,
`/home/linuxbrew/.linuxbrew/bin/{ffmpeg,flac}`, confirmed via `which`)
— available for the one-time authoring step, never wired into
`make derive` or the gate. The straightforward generator is `ffmpeg`'s
`lavfi` null-audio source piped through its FLAC encoder, e.g.:

```
ffmpeg -f lavfi -i anullsrc=r=44100:cl=mono -t 0.1 -c:a flac out.flac
```

producing a few-hundred-byte silent mono FLAC with a valid `fLaC`
marker and `STREAMINFO` block — the minimum mutagen's `FLAC` reader
needs to open successfully
(`mediafile/__init__.py:174`: `type(self.mgfile).__name__ == "FLAC"`
dispatches on mutagen's own parse, which requires the signature and
STREAMINFO). A pure-Python alternative exists if a stranger's
workstation lacks `ffmpeg`: `mutagen.flac.FLAC()` can construct an
empty FLAC container from a hand-built minimal STREAMINFO block
without any audio codec at all, since mutagen only needs the metadata
container to be well-formed — `mutagen` is already a transitive
dependency of `mediafile` in the pinned venv
(`tools/venv/requirements.txt`, `# via mediafile`) and needs no new
pin. Either route is a one-time authoring script, run once by whoever
authors the fixture, with its output committed; 2d's plan should
prefer the pure-`mutagen` route specifically because it removes the
`ffmpeg`/`flac` presence question entirely from even the authoring
step, leaving the committed blob as the only artifact anyone else
ever needs.

**Determinism of the generated fixture is not a derive-time
concern.** Because the blob is committed once and never regenerated
by `make derive` or any gate step (per M4), byte-for-byte
reproducibility of the *generation command* across ffmpeg versions is
irrelevant — the corpus does not re-run the generator. What must be
deterministic is the *test's own file-writing step*: `item.write()`
against a fresh copy of the committed template, run once per subject
per test invocation, in a fresh scratch directory each time (mirroring
`derive_oracle.py`'s and `derive_component.py`'s own
`tempfile.TemporaryDirectory` pattern), so two runs of the ledger
test never contend on one path and a crashed run leaves no residue for
the next.

---

## (c) Copy-into-tempdir import discipline and mediafile read-back

**The discipline, confirmed against pinned beets source**
(`beets/library/models.py`, `tools/venv/.venv/lib/python3.13/site-packages/beets/library/models.py`):

1. **Write:** `Item.write(path=None, ...)` (`models.py:934-981`).
   Defaults `path` to `self.path`; filters the item's own dict to
   `_media_fields` only (line 967); opens `MediaFile(syspath(path),
   id3v23=...)` (line 975); calls `mediafile.update(item_tags)` then
   `mediafile.save()` (lines 978-982). **`write()` returns `None`** —
   there is no success value to trust or distrust; the only signal is
   whether it raised (`ReadError`/`WriteError`) or not. This is the
   exact "never trusting the write call's own return value" M5's
   carry-forward note names: the discipline is not about a return
   value being untrustworthy, it is that there is no return value at
   all, so the *only* way to know what actually landed in the file is
   an independent read.
2. **Read-back:** `Item.read(read_path=None)` (`models.py:903-929`).
   Opens a **fresh** `MediaFile(syspath(read_path))` (line 917,
   independent of the `MediaFile` object `write()` used — a new
   `mutagen.File` parse of the bytes on disk, not the in-memory
   object `write()` already had open) and copies every
   `_media_fields` key from it back onto the item (lines 924-926).
   `derive_file_tags.py` should call `MediaFile(path)` directly
   (matching the plan's own phrasing) rather than route through
   `Item.read()`, since the ledger wants the raw per-field values
   `mediafile` reports, not a second `Item` object; either is
   source-equivalent, since `Item.read()` is exactly that call plus a
   field copy.
3. **The copy-into-tempdir step is what makes this safe to run
   per-case.** Every subject needs its own fresh file: `item.write()`
   mutates the file at `path` in place, so re-using one path across
   subjects (or across ledger-test re-runs) would let subject N's
   write leak into subject N+1's read-back if anything ever failed to
   overwrite a field. The pattern both existing shims already
   establish — one `tempfile.TemporaryDirectory` per process,
   never a shared fixture directory — extends directly: copy the
   committed template FLAC bytes to `<scratch>/<case-id>-<item-index>.flac`
   fresh per subject, `write()` into that copy, `MediaFile()` it back,
   and let the `TemporaryDirectory` context manager clean up
   regardless of outcome.
4. **Plugin hooks fire during write and are inert under the estate
   config**, consistent with the mechanics study's own finding for
   the read/match side: `plugins.send("write", ...)` and
   `plugins.send("after_write", ...)` both fire (`models.py:966`,
   `981`), but the estate's non-MB plugins (`fetchart`, `embedart`,
   `replaygain`) register neither event name — grep over
   `beetsplug/{fetchart,embedart,replaygain}.py` in the pinned venv
   confirms no `register_listener("write", ...)` or
   `register_listener("after_write", ...)` in any of the three
   (`embedart.py` listens for `art_set` and `import_task_files`;
   `replaygain.py` for `import_begin` and `import`; `fetchart.py` for
   `import_task_files` only). `derive_file_tags.py` loads no plugins at all
   (matching `derive_component.py`'s "no stub, no `tag_album`, no
   rendered profile" minimalism), so this is confirmatory, not a
   dependency the code needs to handle.

**beets 2.13.1 citation for the class:** `mediafile==0.17.0` per (a)
above; `MediaFile` class at
`mediafile/__init__.py:140` (`__init__`), `:233` (`path` property),
`:254` (`save`); `MediaFile.readable_fields()` is the classmethod
`Item._media_fields` intersects against (`library/models.py:731`).

---

## (d) The version-bump differential job and quarterly re-capture chore docs

**Both are unwritten**, confirmed absent: no file matching
`*differential*` or `*re-capture*` or `*revalidat*` exists under
`docs/` today (the 2c plan's Task 20 carry-forward note
states this plainly: "The version-bump differential and quarterly
re-capture chore docs are unwritten, deferred with Task 13",
`2026-09-03-rung-2c-catalog-adversarial.md:1059-1063`).

**What the differential doc must contain, per U7**
(`docs/research/2026-09-02-beets-evolution-lessons.md:396-412`,
confirmed by direct read):

- The standing job drives pinned beets as a subprocess with an estate
  config, and the risk it exists to catch is specifically a beets
  major version (3.0.0 is beets' one active break-shedding plan:
  legacy plugin layer TODO-marked for removal, `docs/dev/plugins/
  autotagger.rst` promising removal in v3.0.0) breaking the
  **invocation contract** (config keys, plugin enablement, CLI
  behavior) independently of any verdict change.
- The falsifiable acceptance criterion U7 states outright: "when the
  differential is pointed at a beets 3.x, only the adapter package
  should need edits" — i.e. `internal/harness/oracle`'s two Go files
  (`oracle.go`, `reconcile.go`) plus the Python shims, never a harness
  rewrite. The doc should state this as its own success condition,
  not merely describe the job.
- **The M8 inter-request-floor formula belongs in this doc's wall-time
  section, per the 2c plan's own instruction**
  (`2026-09-03-rung-2c-catalog-adversarial.md:1059-1063`: "The
  differential doc's wall-time section must use the M8
  inter-request-floor formula (`1.04s + 0.25s × (requests − 1)`) at
  the corpus's real size once 2d closes, not a flat per-case
  figure"). The build spec independently requires this: ladder entry
  2 names "the client's hard 0.25s inter-request floor in
  differential wall-time budgets" as an explicit deliverable
  (`docs/specs/2026-09-02-greenfield-server-build.md:508-511`).
  2d's plan must pull the corpus's actual committed case count and
  request-count distribution (from 2c's close, not from this
  report's stale arithmetic) to compute the real figure — this
  report does not attempt that number, since the corpus is mid-write
  under the concurrent capture workflow and any figure taken now
  would be stale by the time 2d executes.
- U6 (capture breadth) is the differential job's second standing
  concern, restated for context: the capture tool records the
  maximal practical `inc=` response rather than the minimal matcher
  view specifically so a future matcher improvement wanting another
  field never forces a full re-capture
  (`beets-evolution-lessons.md:387-395`). The differential doc should
  note this as the reason a beets bump's blast radius is scoped to
  scoring/verdict changes, not response-shape changes.

**What the quarterly re-capture chore doc must contain, from U6/U8
plus the corpus's own stable revalidation rules** (cited by role,
since `corpus/README.md`'s live content is mid-write; the rules
themselves are load-bearing conventions the 2c plan and its review
both quote verbatim, which is what makes them safe to cite here
without opening the file):

- **U8's additive-recall framing**
  (`beets-evolution-lessons.md:413-`): the highest-leverage open
  beets deficiencies widen candidate retrieval (embedded-MBID
  seeding, release-group lookup, better search queries) rather than
  change scoring, and an improvement of that shape "adds candidates
  and leaves the beets-equivalent subset's scores untouched" — which
  is exactly why the fixture format is required to admit
  recall-extension cases from day one (fixtures for lookups beets
  never issued, conformance still checked on the shared candidate
  subset). The chore doc should state the re-capture trigger this
  enables: a post-launch candidate-retrieval improvement is graded by
  the same harness rather than escaping it, per the build spec's own
  ladder-entry-2 language
  (`greenfield-server-build.md:500-509`), so a "quarterly" cadence is
  a floor, not the only trigger — a beets version bump or a
  documented MB schema change are event-driven triggers on top of it.
- **The manifest version quadruple (now quintuple with `mediafile`) is
  what a re-capture event re-pins**, per `Reconcile`'s own doc
  comment: "unidecode's transliteration tables, jellyfish's edit
  distance, and lap's assignment solver all sit directly in the
  distance path... a silent bump there would change verdicts with
  nothing in the corpus to show for it"
  (`internal/harness/oracle/reconcile.go:15-20`, confirmed by direct
  read). The chore doc's procedure is exactly `Reconcile`'s own
  contract: run it before any derivation is trusted, and treat any
  disagreement as the corpus-revalidation event both the mechanics
  study and U4 already name as "already the standing plan"
  (`beets-matcher-mechanics.md:588-591`: "Library-version pins are
  corpus metadata... a beets bump is a corpus-revalidation event
  (already the standing plan)").
- **The wall-time section reuses the differential doc's M8 formula
  rather than restating it** — one number, cited from the other doc,
  scaled to "re-derive the whole committed corpus," which by 2d's
  close is 2c's ~164 cases plus 2d's own file-tag subjects.
- The doc should name its own trigger conditions explicitly, matching
  the "never guard a check with `command -v || skip`" discipline
  CLAUDE.md states for the shell gate generally (a chore doc that says
  "run this quarterly, or sooner if X" without naming X is exactly
  the kind of soft guidance the repo's gate philosophy argues
  against): a beets point release, a `mediafile` bump (now that 2d
  adds it to the pin set), a MusicBrainz schema change noticed via
  U6's response-shape drift, or the calendar quarter, whichever comes
  first.

---

## (e) Derive-time impact: which cases need file-level records

**Twelve subject cases, membership already fixed** by the 2c plan's
own carry-forward note, not 2d's to re-derive:

> "`internal/harness/oracle/derive_file_tags.py` and its Go driver,
> reading back independently through `mediafile.MediaFile(path)`
> after `item.write()`, never trusting the write call's own return
> value, over 12 subject cases (5 from the committed corpus, 7
> flagged by this pass's dossier)."
> (`2026-09-03-rung-2c-catalog-adversarial.md:1026-1029`)

The 5 committed subjects are named explicitly in the review's
"Verified sound" section and confirmed to exist as committed cases
already: `clean-match`, `media-uneven-split`, `va-exact`,
`non-latin-pseudo-release`, `year-reissue-ratio-penalty`
(`rung-2c-plan-review.md:553-555`, "Task 13's five committed subject
cases all exist"). The other 7 are real-catalog cases flagged during
2c's Task 3 selection dossier via its file-tag-subject flag
(`2026-09-03-rung-2c-catalog-adversarial.md:576-583`: "whether it is a
file-tag subject (a flag this pass still records for 2d's benefit,
even though 2d authors those cases)"); 2d's plan reads that flag off
the dossier and the closed corpus at 2c's actual close, not off this
report, since the concurrent capture workflow has not finished
producing them.

**All 12, not a further-reduced subset** — the plan review's phrasing
("5 from the committed corpus, 7 flagged by this pass's dossier")
gives an exact count with no "at least" or "up to" qualifier, and
nothing in M3, M4, or M5's fixes narrows it. 12 is a small, bounded
subject list chosen specifically so the write-side derivation adds
real per-case import work (file creation, `write()`, independent
`MediaFile()` read-back) without scaling with the corpus's 164+ case
total the way `make derive`'s existing O(all cases) passes already
do.

**Per-subject cost, estimated from the write/read-back mechanics
above, not measured**: one `tempfile` copy of the committed template
(microseconds, in-memory filesystem cache), one `MediaFile()` open
plus `update()` plus `save()` (a `mutagen` FLAC container rewrite —
sub-millisecond for a file this small, since there is no audio
re-encode, only metadata block rewriting), one fresh `MediaFile()`
open for read-back. This is negligible against the per-case network
and inter-request-floor costs the M8 formula prices for the
network-driving derive path — the file-level derivation never touches
the network or the stub, so it adds no inter-request floor. The real
cost driver is process count: whether `derive_file_tags.py` runs once
per subject (12 Python process starts) or once for all 12 subjects in
a single process (1 start, matching `derive_component.py`'s
single-invocation shape). 2d's plan should specify one process for
all 12, mirroring `derive_component.py`'s "no arguments, one document
out" contract rather than `derive_oracle.py`'s per-case invocation —
the file-tag ledger, like `components.json`, is a single derived
document keyed by subject, not 12 independent shim calls the Go
driver launches separately.

**What this means for 2d's task sizing**: the file-level derivation's
wall-time is not the pass's binding cost the way `make derive`'s
corpus-wide passes are for 2c. The binding costs for 2d are
authorship (the fixture blob, the second shim, the Go driver, the
gate-resident checker, the oracle-tagged re-derive test) and the two
chore docs, not case-count-driven derive time. 2d's plan should budget
accordingly: no batch-task structure like 2c's twelve capture batches
is warranted here, since 12 subjects processed in one shim invocation
is a single task's worth of derive cost, not twelve.

---

## Summary table

| Question | Answer | Primary citation |
|---|---|---|
| (a) artifact form | Second independent derived ledger (`file_tags.json`), keyed by case+item, values restricted to the `_media_fields` intersection of `AppliedFields`; never a `cassette.Expected`/`Manifest` field beyond the version pin | `derive_oracle.py:214-216`; `library/models.py:731,967`; M3/M5 |
| (b) fixture generation | Vendor one static committed FLAC blob under `oracle/testdata/`; author it via `mutagen` directly (pure-Python) or `ffmpeg`'s lavfi+flac (both present on this workstation but neither becomes a `make derive` dependency); no `_salvage/` precedent actually exists to adopt | M4; `which ffmpeg flac`; `_salvage` search (empty) |
| (c) import discipline | `item.write(path)` returns `None`; independent `mediafile.MediaFile(path)` read-back is the only trustworthy signal; copy-into-tempdir per subject, matching the two existing shims' `tempfile.TemporaryDirectory` pattern | `library/models.py:934-981,903-929`; `mediafile/__init__.py:140-` |
| (d) chore docs | Differential doc: U7's beta-3.0 invocation-contract risk plus its own falsifiable criterion, M8's inter-request formula computed fresh at 2d's close. Re-capture doc: U6/U8's triggers (recall-widening improvements, `Reconcile`'s five-now-six-component quadruple, calendar quarter as floor not ceiling) | U6/U7/U8, `beets-evolution-lessons.md:387-`; `reconcile.go:15-20`; build spec ladder entry 2 |
| (e) subject scope | Exactly 12: 5 committed (`clean-match`, `media-uneven-split`, `va-exact`, `non-latin-pseudo-release`, `year-reissue-ratio-penalty`) + 7 from 2c's dossier flag, read at 2c's actual close. One shim invocation for all 12, not per-subject | `2026-09-03-rung-2c-catalog-adversarial.md:1026-1029`; `rung-2c-plan-review.md:553-555` |
