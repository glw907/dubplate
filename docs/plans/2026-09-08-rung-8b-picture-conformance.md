# Rung 8, Pass 8b (PICTURE, the Seam, and the Conformance Instrument) Implementation Plan

**Revision 2, 2026-09-08.** Folded from two independent adversarial
reviews, merged and re-verified at
`research/2026-09-08-rung-8b-plan-review.md`. One blocker, seven majors
and twelve minors were folded; one finding and one half of another were
rejected with the measurement that retired them. The task count is
unchanged at Task 0 plus seven, so the ceiling stands at 3.25M with the
expected total at 3.06M and the 80% flag at 2.6M; the ceiling table's
simplifier line is re-scoped rather than re-figured, and Task 0 gains a
fourteenth item. Six folds were applied provisionally and marked
**conductor-owed** where they sit: M-c's CR2 answer, M-d's
trust-the-block-length question, M-e's batched simplifier round, R12's
read-side answer, m6's `Parse` bypass, and m9's leftovers basis. The
merged 42-item verified-sound list in the review is the
do-not-re-litigate record for this pass. **The conductor's 2026-09-08
rulings, taken under Geoff's phase grant, resolve all six
conductor-owed folds as applied and retire R1 through R11 from
provisional to ruled**, each marked at its own site; the ceiling stands
at 3.25M and the 80% flag at 2.6M unchanged by the rulings.

**Revision 1, 2026-09-08.** Drafted against the rung 8 spec at revision
2 with the 8a-close erratum in its header
(`docs/specs/2026-09-07-rung-8-tag-io.md`, landed at
`be27434`), whose fourteen conductor rulings are applied and whose four
owner rulings are ruled. The pass-split table's 8b row (`spec:1134`) is
this plan's scope statement. Fork B stays open and owed to Geoff; 8b's
surface is identical under either answer, because the seam carries one
entry at this rung either way, so no task below carries a fold marker
for it.

**What this pass is.** 8b is the second of rung 8's three passes. It
builds the PICTURE block's read, write and strip, the per-format seam
registry, and the conformance instrument that grades twelve corpus
subjects over 69 keys against the file-tag ledger. **8a landed the
container parse, the 69-key value model, the exported read projection,
the writer and the golden container. 8c absorbs the shadow's provisional
reader and repoints its call sites.** Neither is in scope here, in any
part. A task that believes it needs `internal/engine/shadow`,
`buildFLAC`, a new corpus subject, or a `make derive` run stops and
reports.

**What 8b may assume, from the 8a handoff** (`docs/HISTORY.md:345-364`).
`Parse` and its seven refusals are closed, and 8b adds no new container
refusal without a fresh ruling. `Read` is the package's one read entry
point and the instrument calls it rather than carrying a second
projection. `TagSet` carries the four states, with `Mapping()` and
`WriteOrder()` as its exported table accessors, both returning copies.
`Serialize` holds the encoding rules. `Write` holds the merge contract
and `atomicfile.WriteStream` the atomic path. `testdata/golden.flac`
already carries a PICTURE block and `golden.txt` records its decoded
header, so **8b regenerates no golden and runs no Python**.
`ErrUnsupportedContainer` and `ErrMalformedContainer` are landed, and
8b's registry adopts the first unchanged (R7 of the 8a plan, and the
spec's own erratum item 1).

**The pass's own risk control.** Risk 5 (`spec:1282-1285`) is that
PICTURE ships with no corpus evidence at all, because `images` and `art`
are `MediaFile` fields and not beets `Item` fields, so neither is in the
writable 69 (`INV:432-441`). Everything in Tasks 1 and 2 that asserts on
the golden's own PICTURE bytes exists for that gap. Risk 1
(`spec:1247-1252`) is a value model subtly wrong on many keys at once,
and Tasks 4 and 5 are the pass the spec split off to catch it.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section, the pass-execute chain (`~/.claude/workflows/pass-execute.js`)
> per task, sequential. The implementer is `cairn-implementer`,
> **instructed in its dispatch to read and follow
> `.claude/agents/dubplate-implementer.md`** before anything else. The
> conductor consumes reports and never reads diffs inline. Subagents
> never dispatch agents; `code-simplifier` is the conductor's dispatch.
> The conductor owns the ledgers (`docs/STATUS.md`, `docs/HISTORY.md`,
> `ROADMAP.md`), the lane registry, the contested-surface rows, and the
> batch commit of anything a task drafts into them. Rung 8 is not an
> adversarial-cycle rung (those are 5, 12 and 16), so review scales to
> the in-chain `diff-reviewer` (`claude-opus-5`) per task plus the
> pass-end gate (`spec:1323-1331`). **This pass runs in the MAIN TREE.**
> No task runs from a worktree.

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/specs/2026-09-07-rung-8-tag-io.md`, revision 2,
   its header erratum included.** The 8b row of the pass-split table
   (`spec:1134`) is the scope statement. The sections this pass
   implements are "PICTURE" (`spec:639-682`), "The per-format seam"
   (`spec:748-782`), "The conformance instrument" (`spec:818-879`), and
   the four rulings it applies: ruling 2 (`spec:239-244`), ruling 3
   (`spec:245-249`), ruling 13 (`spec:314`), and ruling 14
   (`spec:315-319`). Read "What rung 8 does not build"
   (`spec:899-921`) for the boundary and "Risks" (`spec:1245-1305`)
   for what each task's assertions exist to catch.
2. **`docs/HISTORY.md:5-410`, pass 8a's entry**, and inside it the 8a to
   8b handoff (`:345-364`) and the forward carries (`:365-401`). The
   handoff names what 8b inherits; the carries name what 8b decides. The
   declared coverage gaps (`:263-301`) are the running list this pass
   adds to rather than restarts.
3. **`docs/research/2026-09-07-rung-8-spec-inventory.md`
   (INV), sections 2.4, 3.4 and 5.** Section 2.4 (`INV:432-478`) is the
   read of mediafile's FLAC image layer, key by key. Section 3.4
   (`INV:563-591`) is the five-step procedure the instrument reproduces.
   Section 5's PICTURE paragraph (`INV:770-776`) is RFC 9639 section
   8.8's field order. **Do not re-derive what INV settles; cite it, and
   verify a line you carry forward.**
4. **`corpus/README.md`'s handoff section** (`:2354` onward), and inside
   it the two-altitude rule (`:2361-2377`), the 69 writable names
   (`:2378-2400`), and the nine read-only names with the sentence saying
   why a subject never carries them (`:2401-2412`). Read
   "`file_tags.json`, the file-level applied-tag ledger" (`:2171-2235`)
   for the twelve subjects and the ruled twelfth substitution, and "The
   adjudication artifacts" (`:2255-2280`) for what
   `corpus/analysis/adjudications.json` is and who may write it.
5. **`internal/harness/oracle/derive_file_tags.py`**, its module
   docstring and `derive_row` (`:128-161`). That function is what the
   instrument reproduces in Go, and its per-subject scratch path is a
   rule rather than a detail.
6. **The tree, read-only:** `internal/engine/tagio/doc.go`,
   `container.go`, `tagset.go`, `encode.go`, `read.go`, `write.go`,
   `deviations.go`, and `testdata/golden.txt`;
   `internal/harness/analysis/adjudication.go` and its test;
   `internal/harness/analysis/file_tags.go`;
   `internal/harness/conform/arch_test.go`; `scripts/check.sh`;
   `.gitleaksignore`.
7. **`docs/research/2026-09-04-reverse-edge-completion.md`**,
   cells (8,3) and (8,4) (`:161-199`) and the ownership rows at
   `:177-182`. Those rows are why the instrument is read-only against
   `corpus/` and `internal/harness`, and why the divergence record has a
   home at all.
8. **`CLAUDE.md`** in full, and `docs/STATUS.md`'s lane registry.
9. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** This
   pass writes no Python, no shell, no unit and no deploy path, so
   `python-conventions` and `vps-conventions` are both out of scope
   (`spec:1319-1321`).

## Goal

Ship the three deliverables the spec assigns to 8b: PICTURE, the seam,
and the grading.

Land the PICTURE block decoded into its RFC 9639 section 8.8 fields with
all three declared lengths bounds-checked, the front-cover selection rule, the
media-type sniff from the picture's own bytes, the write that replaces
the whole picture set with zero geometry, the strip, the registry keyed
by container that exposes the supported set and dispatches by func
value, and the conformance instrument that writes each of the twelve
`file_tags.json` subjects into a scratch copy of the committed template,
reads it back through `tagio.Read`, and grades all 69 keys against the
ledger row.

Land the divergence record's shape and its route into
`corpus/analysis/adjudications.json`, so a disagreement that survives
investigation has a home before the pass that can produce one runs
(`spec:1303-1305`).

Nothing here touches the box, deploys anything, writes under `corpus/`,
adds a module dependency, reaches the live network, runs the venv,
declares a configuration key, regenerates the golden, or edits
`internal/engine/shadow`.

## Owner touchpoints: none expected

8b raises no attended question by design. Fork B is Geoff's and is
already owed at the next sitting; the seam carries one entry under
either answer, so the pass does not wait on it. Rung 8 declares no
configuration key (ruling 13, `spec:314`), so no template gains a name
and no default reaches an adopter through this pass.

**What would become an owner touchpoint, so a task recognizes one rather
than absorbing it.**

- **A divergence that survives investigation.** This is the pass that
  can produce one, and it is the reason ruling 2 exists. The route is
  Task 6's: the instrument stages a pending row, the conductor folds it
  into `corpus/analysis/adjudications.json`, and it joins the standing
  adjudication stream Geoff already reads
  (`spec:1332-1337`). **A divergence is not an ask.** It becomes a
  sitting only if its resolution would change what an adopter sees,
  which no value in the writable 69 does on its own.
- **Any key added to `env/dubplate.env.template`.** That file is a
  secrets manifest rendered by `deploy.sh` from the age store, and
  adding a name there requires minting the secret (CLAUDE.md, "Facts
  agents get wrong"). Rung 8 adds none.
- **Any adopter-visible behavior**: a reject wording, an accepted-format
  decision, a configuration default, or a change to what the intake gate
  admits. Those belong to rung 7's string authority and rung 7's config,
  bounded by what this rung supports (`spec:768-777`). 8b supplies the
  bound and the container name, never the wording.
- **Whether the pipeline embeds cover art by default.** Rung 9's knob by
  owner ruling 3, and rung 8 mints no key for it (`spec:158-161`).

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition below resolves to the conductor. The 80%-of-ceiling
flag routes to the conductor rather than to Geoff (Geoff, 2026-09-08).

## Spec

`docs/specs/2026-09-07-rung-8-tag-io.md` (revision 2, with
the 2026-09-08 erratum), and behind it
`specs/2026-09-02-greenfield-server-build.md` (revision 2.1, ratified),
ladder entry 8. 8b builds deliverables 3, 4 and 5 of the spec's seven
(`spec:72-112`). Deliverables 1 and 2 landed at 8a, and so did
deliverable 6's golden container; its PICTURE byte assertion is Task 2's
here, because 8a had no picture encoder to grade (`spec:131-138`).
Deliverable 7 is 8c's.

**The forward criterion this pass hands rung 15**, so rung 15 reads it
from the ledger rather than rediscovering it:

> Rung 15 re-runs 8b's conformance instrument over the twelve ledger
> subjects (`spec:1020-1023`). The re-run criterion is that all twelve
> subjects agree with `corpus/analysis/file_tags.json` on all 69
> writable keys, or that every disagreement carries an adjudicated
> divergence row in `corpus/analysis/adjudications.json`. **An
> unadjudicated divergence blocks the launch gate.** The re-run is `go
> test ./internal/engine/tagio/... -run Conformance -count=1`, which
> selects `TestConformanceLedgerSubjects` by name. **Rung 15 checks the
> reported test count rather than the exit status alone**, because a
> `-run` pattern matching nothing prints "no tests to run" and exits 0.
> The re-run needs no venv and no network, and reads the same tracked
> files the pass committed against. The golden container is a fixture
> rung 15 reads and never regenerates.

## Token ceiling: 3.25M

The spec proposes 3M for 8b on INV's budget evidence
(`spec:1181-1196`). The arithmetic is re-derived here from 8a's actual
report rather than carried, because a ceiling quoted without its
derivation is a number nobody can check, and because the spec's own
clause requires it: "Each ceiling is re-derived from the previous pass's
actual report before the next plan is drafted" (`spec:1194-1196`).

**The basis, read off `docs/HISTORY.md:403-410`, pass 8a's spend.** 8a
closed at about **3.6M against a 4.5M ceiling** for Task 0 plus nine
implementer tasks. Itemized: plan preparation 0.87M; Task 0, 0.1M; the
chain 2.27M across 24 agents; the leftovers round 0.2M; conductor turns
about 0.15M. The 24-agent figure over nine tasks means the chain line
**already includes** every `diff-reviewer` pass and every fix
re-dispatch, so the measured per-task rate is 2.27M over nine, about
**252K**, with review and fixes inside it. A separate fix-round line
would double-count.

**The arithmetic for 8b.**

| Line | Basis | Figure |
| --- | --- | --- |
| Prep, this plan's authoring workflow | Below 8a's 0.87M: the spec is unchanged, the handoff is written, and this plan reads a subset of 8a's inputs | 0.5M |
| One batched `code-simplifier` round over 8b's Go, plus its review | Flat at about 0.2M across seven passes. Forward carry 1's round over 8a's Go landed ahead of this plan at `0f2b61d` and its actual is reported at the close against this same line | 0.2M |
| Task 0, staleness, `opus` at zero context | 8a measured 0.1M for a thirteen-item checklist | 0.1M |
| Tasks 1 through 7 through the chain | Seven implementer tasks at 8a's measured 252K, reviews and fix rounds included | 1.76M |
| One `opus` upshift (Task 5, the grading) | Roughly one task's worth again | 0.2M |
| A leftovers round, if one is needed | 8a measured 0.2M for six findings over nine tasks; seven tasks at the same rate is about 0.156M (**ruled, conductor 2026-09-08, as applied, m9:** the line stays 0.15M on this basis; booking the full 0.2M instead would have put the expected total at 3.11M and the headroom at 0.14M) | 0.15M |
| Conductor overhead | 8a measured 0.15M | 0.15M |
| **Expected total** | | **3.06M** |

**The ceiling is proposed at 3.25M, because 3M schedules its own stop
rule to fire.** Prep and Task 0 spend 0.6M before Task 1 dispatches.
Walking forward, Tasks 1 through 6 plus Task 5's upshift reach 2.31M,
the batched simplifier round takes it to 2.51M, and Task 7 closes at
2.76M. An 80% flag against a 3M ceiling lands
at 2.4M, which falls inside Task 6, and a flag that fires mid-chain by
the plan's own arithmetic is not a signal. At 3.25M the flag lands at
**2.6M**, still inside the chain but at Task 7, the close, where a stop
is cheap and the conductor's action is a re-derivation the plan already
schedules. The 0.19M of headroom over the expectation is the smallest
band that keeps the flag meaningful.

**Recorded against the rung reservation.** `spec:1181` reserves 8M for
rung 8 (4M, 3M, 1M). Ceilings now stand at 4.5M, 3.25M and 1M, or
**8.75M**. The actuals tell a different story: 8a spent 3.6M against its
4.5M, so the rung's measured trajectory is 0.9M under its own ceiling
before 8b starts. 8c's 1M line is re-derived when this pass closes, per
the spec's clause, and Task 7 drafts that re-derivation.

**Flag at 80%, which is 2.6M.** Finish the task in flight, write STATUS,
and send one combined question to the conductor. **The conductor's
action when it fires is the re-derivation this plan already schedules**,
since the flag is a forecast the plan made rather than a surprise.

**Re-derive the ceiling from Task 2's actual report.** Task 2 is the
pass's first byte-level write task and its first golden assertion, which
is the shape 8a measured most precisely. If the re-derivation passes
4M, the conductor checks the window cap before continuing and names the
split at the **Task 3 / Task 4 boundary**: Tasks 0 to 3 would be 8b
(PICTURE and the seam), and Tasks 4 to 7 would be 8b-ii (the instrument,
the divergence record and the close). That is the boundary where the
risk changes shape from bytes to semantics, and it is the only cut in
this pass that leaves both halves shippable, because PICTURE plus the
seam is a complete surface and the instrument is a complete grader.

**Checkpoint: STATUS after every four tasks**, so after Task 4 and at
the close. Every task reports its own spend, its mutation outcomes, and
its measured gate wall time, and the conductor sizes the next task
against it.

## Global Constraints

- **Main tree.** The spec's tree column for 8b is "Main tree"
  (`spec:1134`). The lane registry entry is written before the first
  dispatch, and **the live-executor check is that registry, never
  `pgrep -f Projects/dubplate`**, which matches every lane once a
  worktree exists (CLAUDE.md, Lanes).
- **Package territory, stated as a wall.** This pass writes
  **`internal/engine/tagio`** (extending it, including its `testdata`)
  and **`internal/harness/analysis/adjudication.go`** plus its test and
  testdata, and nothing else. It writes **nothing** under `corpus/`, any
  other file in `internal/harness/`, `internal/engine/{shadow,matcher,
  musicbrainz,tag}`, `internal/atomicfile`, `internal/config`,
  `internal/serve`, `cmd/`, `systemd/`, `scripts/`, `env/`, `Makefile`,
  `tools/`, `web/`, or `_salvage/`, and no `docs/` file outside Task 7's
  conductor-owned drafts.
- **Read-only means the instrument, and the schema addition is the named
  exception.** `corpus/` is read-only to this pass without qualification.
  `internal/harness/` is read-only **to the conformance instrument**,
  which is the ownership row `RE` cells (8,3) and (8,4) already wrote
  (`RE:161-199`). The one write into `internal/harness` this pass makes
  is **Task 6's applied-field divergence block in
  `internal/harness/analysis/adjudication.go`**. `spec:1165` grants the
  schema addition rather than the file: its words are "rung 8 adds its
  applied-field divergence rows as a schema addition, never by
  hand-editing a derived record". **The file follows by reading**, since
  the divergence schema is defined in `adjudication.go` and a schema
  addition has nowhere else to land. That reading is stated here rather
  than quoted, so a reviewer grades an argument instead of a
  misquotation. A task
  other than Task 6 that finds itself editing anything under
  `internal/harness/` stops and reports.
- **The engine never imports the harness, and never reads the clock.**
  `TestEngineNeverImportsHarness`
  (`internal/harness/conform/arch_test.go:141`) and
  `TestEngineNeverReadsTheClock` (`:152`) both root at
  `internal/engine`, descend into nested packages, and walk `_test.go`
  files outside a `testdata` subtree. **The conformance instrument is
  inside both walks.** It therefore decodes `file_tags.json` and every
  `case.json` into **its own local types**, never
  `internal/harness/analysis`'s, and it **stamps no date**: every date
  and commit value in a staged divergence row is a caller-supplied
  parameter (R7).
- **`tagio` never imports `internal/engine/matcher`** (ruling 3,
  `spec:245-249`). `tagio` exposes its own tag set type and no
  consumer's type appears in its signature.
  `internal/engine/tag/doc.go:8-11` states the split in words and stays
  true.
- **No interface, exported or not, inside `tagio`** until the MP3 pass
  brings the second implementation (ruling 14, `spec:315-319`). The
  registry's value is a func value. The go-conventions rule is about
  existence rather than export, so an unexported single-implementation
  interface is the same anti-pattern as an exported one.
- **No configuration key, stated in those words** (ruling 13). No task
  edits `internal/config`, and Task 0 item 10 confirms it.
- **Module dependencies: none added.** `go.mod` carries cobra and its
  two indirects. No image library and no tag library joins it: every
  candidate is declined in the spec's prior-art table
  (`spec:1048-1078`), `dhowden/tag` most pointedly, and CN's
  recommendation is to sniff by magic bytes in-repo. A task that wants a
  dependency stops and reports.
- **No Python, no venv, no derivation.** The gate never requires uv or
  beets by design (`spec:874-879`). 8a's Task 2 ran the one offline
  generator this rung needs, and **8b runs none**: the golden is
  committed and `golden.txt` records what it holds
  (`docs/HISTORY.md:345-349`). No task runs `make derive`, and no task
  derives a new corpus subject (ruling 1, `spec:899-921`).
- **Assertion discipline, at the spec's own bar.** No silent-success
  fakes, no vacuous assertions, and **no expectation derived from the
  code under test**. Every hard-coded expectation in this pass comes
  from RFC 9639, from the pinned mediafile and mutagen sources read at
  their committed versions, from INV's measurement, from
  `testdata/golden.txt`, or from `corpus/analysis/file_tags.json`. The
  instrument's own ability to fail is exit-gate content
  (`spec:1134`), and Task 5 carries it as a test rather than as a
  one-off demonstration.
- **Universality.** No estate hostname, path, email, catalog value or
  personal name anywhere. Fixture bytes and fixture strings are
  invented. The rung implements a published format and a pinned
  library's serialization, both identical for every adopter
  (`spec:1339-1348`).
- **`go-conventions` before any Go.** The bar is the standing
  exemplary-stdlib-idiom bar (`idiomatic-beautiful-go`). Concrete types
  over interfaces, `any` not `interface{}`, no `init()`, guard clauses
  over nesting, receivers named for their type, errors lowercase and
  context-first with `%w` only where a caller branches. **The em dash is
  banned in Go comments**, and vale scans a new file before it is
  staged.
- **Atomic writes.** Every file this package writes goes through
  `internal/atomicfile`. No task calls `os.WriteFile` or `os.Create` on
  a target path. A test's scratch file under `t.TempDir()` is not a
  target path and may use either.
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck
  over `scripts/*`, ruff `check` and `format --check` over every tracked
  `*.py` (**this pass adds none**), gitleaks over the tracked worktree
  with `--no-git`, vale over Go comment prose, `docker compose config`,
  and a pattern-filtered `systemd-analyze verify`. **Never commit red.
  Specific files only.** Commit footer `Co-Authored-By: Claude
  <noreply@anthropic.com>`.
- **`.gitleaksignore` is tracked and empty**, measured at zero bytes at
  `0f2b61d`; commit `3a35558` emptied it when the config key fixture was
  renamed. No task in this pass edits it. This pass commits **one new
  tracked data file**, Task 6's `divergence-example.json`, which carries
  invented values and no key-shaped field name. If gitleaks flags it,
  the fix is the file's first fingerprint line, and the task reports the
  fingerprint it wrote.
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each one before
  Task 1 dispatches.

## Contested surfaces

The spec's own table (`spec:1153-1170`) is the source. The conductor
writes these into the parallel-approach spec's ownership table before
the lane launches, per lane-launch checklist item 2.

| Surface | Owner | Consumers | 8b's position |
| --- | --- | --- | --- |
| `internal/engine/tagio` | rung 8, created by pass 8a; no other lane writes it | rung 7's allowlist bound, rung 9's pipeline, rung 15's re-run | **8b extends it**: `Container` and `Write` widen for pictures, the registry and the instrument are added. 8c adds the absorbed reader. |
| `internal/engine/tagio/testdata`'s golden container | rung 8, generated once under the pinned venv, then committed and never regenerated in the gate | rung 15 reads it as a fixture | **8b reads it and never regenerates it.** It already carries the PICTURE block (`docs/HISTORY.md:345-349`). |
| Rung 8's conformance instrument | rung 8, under `internal/engine/tagio` with its own `testdata`, **read-only** against `corpus/analysis/file_tags.json`, the twelve `corpus/cases/<id>/case.json` files, and `internal/harness/oracle/testdata/template.flac` | rung 15's re-run | **8b builds it** (Tasks 4 and 5). It pins no schema version of its own and never calls `Reconcile`. |
| `corpus/` | rung 2 | rungs 3, 8, 15 | **Read-only, whole.** No task writes a byte under it, the adjudication ledger included; Task 6 stages a row for the conductor to fold. |
| `internal/harness/` | rungs 2 and 3 | rung 15 | **Read-only to the instrument.** The one write is Task 6's schema addition in `internal/harness/analysis/adjudication.go`, granted by `spec:1165`. |
| `corpus/analysis/adjudications.json` and the divergence schema | lane E1 authored it; rung 4 added the shadow rows and the matcher-commit stamp; **rung 8 adds its applied-field divergence rows** as a schema addition, never by hand-editing a derived record | rungs 3, 8 and 15 read it; `disposition` stays owner-only | **8b adds the shape** (Task 6). It adds a row only if the instrument finds a divergence, and the row is staged for the conductor. |
| `internal/atomicfile` | born by rung 4 pass 4a; rung 8 added the streaming write at 8a | rungs 6, 9, 13 | **Untouched by 8b.** 8a's `WriteStream` is complete. |
| `internal/engine/shadow` | rung 4, created by it | rung 15's re-run, rung 17's docs | **Untouched by 8b.** That work is 8c's, whole. |
| `docs/` ledgers | conductor-owned | every rung | A pass drafts and never commits. |

**Lane 2 runs concurrently, in a worktree, and its territory is
disjoint.** Lane 2 authors and then executes pass 5a from the rung 5
identity and auth spec (`docs/STATUS.md`'s lane registry, and
`c5021ce`). Its package territory is `internal/serve` and `web/`, and it
holds its own git worktree because `scripts/check.sh` runs the module
gate over `./...`. **Nothing 8b writes is anything 5a writes**, and
nothing 8b reads is anything 5a writes. Task 0 item 1 verifies that
rather than assuming it, and reports any file under
`internal/engine/tagio`, `internal/harness/analysis` or `corpus/` that
the lane 2 worktree has touched.

## The 8b package contract, as outcomes

Every task below cites this section rather than restating it. Items are
numbered P1 through P12 to keep them distinct from 8a's C1 through C16,
which stand unchanged. **The plan states the contract; the implementer
writes the code.**

**P1. The picture is its own exported type**, carrying the eight values
RFC 9639 section 8.8's Table 12 defines, in the order it defines them:
the picture type index, the media type string, the description, the
width, the height, the colour depth, the indexed-colour count, and the
data (`INV:770-776`). **The block itself carries eleven fields**, the
eight above plus the three `u(32)` lengths that prefix the media type,
the description and the data; the lengths are framing rather than
values, so the type carries none of them. Geometry is carried on read
and written as zero (P6).
No consumer's type appears in its signature, the same rule ruling 3 puts
on `TagSet`.

**P2. `Container` carries the pictures it parsed**, in on-disk order,
decoded from every PICTURE block rather than only the first. RFC 9639
section 8.8 allows more than one PICTURE block per stream, unlike
SEEKTABLE and VORBIS_COMMENT, so a second PICTURE block is not a
refusal. 8a's R5 named this widening as 8b's, "one refactor inside
`tagio` with no consumer wired yet".

**P3. Every declared length in a PICTURE block is bounds-checked before
a slice is taken.** The block carries three length fields, each a
big-endian `u(32)`: the media type length, the description length and
the picture data length (`INV:770-776`,
`mutagen/flac.py:612-623`). The data length is the one an attacker
controls most directly. This is the exact defect `dhowden/tag` open
issue #116 carries, where a 4-byte length reaches `make([]byte,
dataLen)` and a 49-byte crafted FLAC commits a roughly 2 GB allocation
(`spec:1061`). A length running past the block's own declared payload is
`ErrMalformedContainer` under 8a's existing sentinel, so the sentinel is
not a new refusal class.

**The condition, however, diverges from the reference, and Task 2
records it.** Mutagen sets `_distrust_size = True` on `Picture`
(`mutagen/flac.py:575`) and on `VCFLACDict` (`:350`), and
`FLAC.__read_metadata_block` therefore ignores the 24-bit block length
and parses each from the file object using the block's internal lengths
(`:712-724`, whose comment cites quodlibet/mutagen issue 106 for the
real-world shape). A library file with a wrong block length is one beets
can tag and `tagio` refuses, and the refusal reaches rung 7's intake
gate as a rejection. 8a already applies the same policy to
VORBIS_COMMENT (`container.go:344-346`) and never recorded it. **The
hardening stands; the silence does not.** Task 2 adds the row to
`deviations.go` covering both block types. **Ruled, conductor
2026-09-08, as applied (M-d):** tagio trusting the block header length
where mutagen sets `_distrust_size` is discharged by the stated
deviation, with no new refusal condition. The `deviations.go` row
covers VORBIS_COMMENT as well as PICTURE, and records the same trust
8a already applied and left unrecorded.

**P4. The picture type index is the ID3v2 APIC enumeration**, 0 through
20, with front cover at 3. An out-of-range index degrades to 0, `other`,
on read, which is what `Image.__init__` does when it is handed an
integer index outside `list(ImageType)`
(`mediafile/utils/image.py:40-47`, read at the pinned version). An unset
type is 0 (`:54-60`).

**P5. Front-cover selection is a query over a parsed picture set**: the
single candidate when there is exactly one, otherwise the first with
type front, otherwise the first of any type. That is
`CoverArtField.guess_cover_image` (`mediafile/fields.py:253-260`,
`INV:466-470`). **The single-candidate rule comes first**, so a lone
back cover is the cover.

**P6. A write emits zero for width, height, colour depth and
indexed-colour count**, and a rewrite discards whatever geometry the
incoming picture carried. mediafile's FLAC image write builds a fresh
`mutagen.flac.Picture` and sets only data, type, mime and desc
(`mediafile/storage/flac.py:30-37`), and `Picture.__init__` initialises
all four to 0 (`mutagen/flac.py:577-586`). A Go writer that fills them
from a decoded image writes more than the reference did
(`spec:663-676`). **The rung's close records this as a stated behavior**,
so a later rung that wants real dimensions changes a statement rather
than discovering an accident.

**P7. The media type is sniffed from the picture's own bytes on write,
and read from the block on read.** mediafile calls
`filetype.guess_mime(data)` through `Image.mime_type`
(`mediafile/utils/image.py:12-14`, `:49-52`) and ignores whatever the
caller declared. `tagio` sniffs by magic bytes in-repo, which is also
CN's recommendation for container detection and the reason
`dhowden/tag` is declined (`spec:1061`). The recognized set and the
no-match behavior are R2.

**P8. A picture store replaces the whole set; a delete clears it.**
mediafile's FLAC image style calls `clear_pictures()` and then
`add_picture()` per image (`mediafile/storage/flac.py:24-28`), and its
delete is `clear_pictures()` alone (`:39-41`), which is owner ruling 3's
strip. **Preservation is therefore the caller's**: a caller that wants a
back cover to survive passes it back. The three states a caller
expresses are R1.

**PICTURE therefore joins VORBIS_COMMENT as a block type the writer owns
and re-plans**, and it leaves the preserve-unknown set ruling 7 governs.
8a's `planBlocks` carries every non-PADDING block through in place
(`write.go:69-105`), which is right for a block the writer does not
decode and wrong for one it does: a store that both preserves the
incoming blocks and emits the new set writes the picture twice, and a
strip then removes nothing. **Position is part of the ownership.** A
rewritten picture replaces its own block in place. A picture written
into a container carrying none goes after the VORBIS_COMMENT block and
before the trailing PADDING, which is where the oracle put it in the
golden (`golden.txt:120-124`) and which leaves C9a's fresh-comment
placement intact.

**P9. The seam is a registry keyed by container**, with one entry, a
func-value dispatch, an exported supported set, and no interface
(ruling 14, `spec:748-766`). Sniffing routes to the FLAC implementation
or returns `ErrUnsupportedContainer`, the sentinel 8a landed and 8b
adopts unchanged. **Rung 7 reads its allowlist bound from the supported
set** rather than restating it as a config default, which is what makes
a later container a registration rather than an edit in two places
(`spec:768-777`). Per-format reject wording is rung 7's; 8b's
contribution is the sentinel and the container name.

**P10. The conformance instrument reproduces `derive_file_tags.py`'s
five steps in the same order** (`spec:836-857`, `INV:563-591`): copy the
committed 42-byte template to a scratch path **per subject**, take the
subject's frozen applied fields from its committed case filtered to the
ledger's own `writable_fields` header, write them with the four-state
rule, read the container back through `tagio.Read` and project each of
the 69 keys to the ledger's JSON shape, and compare against the row.
**The nine read-only keys are never compared per subject**, and the
ledger's shared `template` block describes the fixture rather than any
case (`corpus/README.md:2401-2412`).

**P11. The instrument calls `tagio.Read` rather than carrying a second
projection**, because two decoders over one field bag drift silently
(`spec:731-736`, `docs/HISTORY.md:345-347`). It pins no schema version
and never calls `Reconcile`.

**P12. A divergence is first a bug in the Go writer.** When it survives
investigation and its resolution is an owner question, it becomes a
divergence record staged beside the conformance run and routed into
`corpus/analysis/adjudications.json` as a schema addition, never by
hand-editing a derived record (ruling 2, `spec:239-244`). `file_tags.json`
is the derived record that is never hand-edited;
`adjudications.json` is hand-written by design
(`corpus/README.md:2255-2258`), and folding a staged row into it is the
sanctioned route. An unadjudicated divergence blocks the launch gate.

## Resolutions this plan takes rather than discovers

Each was applied as the recommendation and marked provisional at
revision 1. The conductor's 2026-09-08 rulings, taken under Geoff's
phase grant, retire R1 through R11 from provisional to ruled, each
marked **ruled, conductor 2026-09-08, as applied** at its own site.
Each still names what changes if it is ruled the other way, so a
reversal is a named edit rather than a rediscovery.

### R1: the picture set has three caller states, matching C2's idiom

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** `Write`'s picture argument distinguishes three states the
way `TagSet`'s list members already do (C2). A **nil** slice leaves the
file's existing pictures untouched. A **non-nil empty** slice strips
every picture, which is owner ruling 3's strip. A **non-empty** slice
replaces the whole set. The reference has the same three behaviors
through a different surface: `MediaFile.update` touches only the fields
its dict carries, `mf.images = [...]` stores, and `del mf.images`
deletes.

**Why.** Without a leave-alone state, every tag-only write strips the
cover, and rung 9's retag verb would have to read and re-pass the
picture set on every call to avoid destroying art. The nil versus
non-nil-empty distinction is idiomatic Go and is already the package's
own vocabulary for exactly this ambiguity.

**Fold if ruled otherwise.** `Write` keeps a pictures-always-replace
contract and gains a separate strip entry point. The cost is a second
exported verb and a caller that must read before it writes.

### R2: the recognized media types, and what an unsniffable blob does

**Ruled, conductor 2026-09-08, as applied.** M-c's CR2 answer stands as
applied: a refusal plus a `deviations.go` row recording that pinned
`filetype` reports `image/x-canon-cr2` where `tagio` refuses.

**Applied.** The sniff recognizes the raster formats a cover
realistically carries, each reproducing the whole magic-byte test the
pinned `filetype` package applies rather than its leading prefix. Read
at `filetype/types/image.py`, at the pinned 1.2.0:

| Row | The test, in full | Source |
| --- | --- | --- |
| JPEG | `FF D8 FF` | `:22-27` |
| PNG | `89 50 4E 47` | `:103-109` |
| GIF | `47 49 46` | `:124-129` |
| WEBP | at least 14 bytes, `RIFF` at 0, `WEBP` at 8, **and `VP` at 12** | `:144-155` |
| TIFF | at least 10 bytes, `49 49 2A 00` or `4D 4D 00 2A`, **and not `CR` at 8** | `:193-199` |
| BMP | `42 4D` | `:215-219` |

**The two emphasised clauses are the ones a leading-prefix
reimplementation gets wrong**, and both change the answer. A RIFF
container that is not a WebP bitstream is refused by the reference and
would be written as `image/webp` by a prefix-only sniff. A Canon raw
file carries TIFF's own prefix, and `Tiff.match`'s `not (buf[8] == 0x43
and buf[9] == 0x52)` clause is what routes it to `Cr2` and
`image/x-canon-cr2`; tuple order does not, since `Tiff()` precedes
`Cr2()` in `IMAGE` (`filetype/types/__init__.py:15-32`). The tuple order
still governs the APNG-before-PNG case R3 declines.

**The CR2 collision is a stated deviation, ruled, conductor 2026-09-08,
as applied:** `tagio` refuses a CR2-prefixed blob where the reference
writes it as `image/x-canon-cr2`, and Task 2 records the row in
`deviations.go`. A seventh sniff row that reproduces the reference
would have been the fold had it been ruled the other way, at the cost
of a format no cover realistically carries.

**A blob matching no row is refused** with a wrapped error rather than
written with a guessed or empty media type. The reference has no
fallback at all: `Image.mime_type` returns `None` when
`filetype.guess_mime` fails, and that `None` reaches
`mutagen.flac.Picture.mime` (`mediafile/storage/flac.py:34`), so
mediafile raises rather than writing a picture with no type. Refusing is
the closest honest reproduction, and it keeps a garbage blob out of a
library file.

**No new sentinel.** No caller branches on it yet, so it is a wrapped
context error under the package's existing rule (C14). Rung 9 is the
caller that would want a sentinel, and it can have one when it can use
it.

**Fold if ruled otherwise.** The no-match case writes
`application/octet-stream` and the deviation note records the divergence
from a reference that raises.

### R3: APNG discrimination is not reproduced

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** `filetype` tries APNG before PNG and distinguishes them by
walking the chunk table for an `acTL` chunk appearing before `IDAT` or
`IEND` (`filetype/types/image.py:51-88`). `tagio` matches PNG's
signature and reports `image/png` for both. An animated cover is not a
case this estate or any adopter's library realistically holds, and a
chunk-table walk over caller-supplied bytes is new attack surface for no
benefit.

**This is a stated deviation and a declared coverage gap**, recorded in
`deviations.go` by Task 2 and counted by Task 7, rather than an
omission a later reader has to discover.

**Fold if ruled otherwise.** Task 2 adds the chunk walk with its own
bounds checks and its own refusal for a malformed chunk table, and the
deviation row is removed.

### R4: picture fixtures are built in-test, and no new binary is committed

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** Every picture fixture in Tasks 1 and 2 is built in the test
from a hand-written magic-byte prefix plus invented payload bytes. **No
image file joins `testdata`.** The one oracle-produced picture this pass
grades against is the golden's, which 8a committed and `golden.txt`
records field by field.

**Why.** A committed image would be a second binary fixture whose bytes
no reviewer can read from a diff, for a property a twelve-byte literal
proves as well. It would also enter the gitleaks leg for nothing.
`go-conventions` calls goldens a last resort, and this pass already has
the one golden it needs.

**Fold if ruled otherwise.** A small invented PNG is generated once and
committed with a provenance record beside it, on 8a Task 2's precedent.

### R5: the instrument reads its tracked inputs by relative path

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** `go test` runs with the working directory set to the
package directory, so the instrument names its three inputs as path
constants relative to `internal/engine/tagio`: the ledger, the case
directory, and the template. The constants sit in one block, each with a
comment naming the rung that owns the file and the fact that it is read
and never written. A missing file fails the test loudly, naming the
expected path, rather than skipping.

**No conditional skip.** A check either gates or does not exist. An
absent ledger is a broken tree, not a reason to pass.

**Fold if ruled otherwise.** A helper resolves the module root from `go
env GOMOD` and joins from there, at the cost of a subprocess in a test
the arch walk already constrains.

### R6: the subject list comes off the ledger's own rows

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** The instrument derives its twelve subjects from
`file_tags.json`'s own `rows`, each row's `case_id` and `item_index`,
rather than carrying a second hard-coded copy of
`analysis.FileTagSubjects`. It asserts the row count against the ledger's
own floor of twelve and asserts each `case_id` resolves to a committed
case directory.

**Why.** The instrument cannot import `internal/harness/analysis`, so a
hard-coded list would be a third copy of a list that already exists
twice, and a subject added to the ledger would be silently ungraded. The
ledger is the conformance target, so the ledger names the subjects.

**Fold if ruled otherwise.** The twelve ids are hard-coded in the test
with `corpus/README.md:2182-2192` as their citation, and a count
assertion catches a ledger that grew.

### R7: the divergence route, and where the schema addition lives

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** Three pieces, and no new verb.

1. **The schema addition** is an optional `applied_field` block on
   `analysis.AdjudicationEntry`, in
   `internal/harness/analysis/adjudication.go`. It carries the item
   index, the field name, the ledger's value, the writer's read-back
   value, and the `tagio` commit the row was observed at. It is legal
   only on a case-subject row. **Nothing existing changes**: every field
   the checker already requires stays required, and an applied-field row
   fills `release_truth` with the ledger's own value and its citation,
   because for this row class the ledger is the truth.
2. **The emitter** is the instrument's own, in `internal/engine/tagio`,
   over a local struct that mirrors the shape. It **takes the date and
   the commit as parameters** and stamps neither itself, because
   `TestEngineNeverReadsTheClock` walks this package's test files and
   because a commit is not a fact a test can read. The instrument passes
   empty strings; the failure message names the two fields a human
   fills.
3. **The route** is the pattern rung 4 already proved: the engine
   produces a pending row, something outside the engine marshals and
   validates it, and the owner folds it. Rung 4's marshaller is
   `cmd/dubplate-server`'s `pendingJSON`
   (`cmd/dubplate-server/shadow.go:151-186`), which sets `disposition`
   to pending and `adjudicator` to the agent sentinel as constants
   rather than reading them off the row. **8b adds no verb**, because
   there is no running system to stage into: the instrument prints the
   row and writes it under `t.TempDir()`, and the conductor folds it
   into `corpus/analysis/adjudications.json` by hand, which is how that
   hand-written ledger has always been written.
4. **The anchor against drift** is one committed example row,
   `internal/engine/tagio/testdata/divergence-example.json`, carrying
   invented values and a fixed date. The `tagio` emitter test asserts
   its output equals that file byte for byte for a synthetic divergence.
   The `internal/harness/analysis` test reads the same file, decodes it
   into `AdjudicationEntry`, and asserts `CheckAdjudications` accepts
   it. One literal, two graders, no second copy to drift.

**Fold if ruled otherwise.** A `dubplate-server conform` verb runs the
grading outside `go test`, stages rows the way the shadow does, and
carries its own cobra wiring. The cost is a subcommand rung 8 said it
would not add (`spec:928-929`) and a second execution path for the same
comparison.

### R8: the `io.ReaderAt` gap is declined at 8b

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** Forward carry 2 (`docs/HISTORY.md:369-371`) leaves it to 8b
or 8c whether `Parse` grows a second, injectable entry point. **8b
declines it.** No task in this pass needs to inject a failing reader, no
consumer arrives until rung 9, and widening a signature to make one
untestable error path testable buys a coverage gap's closure at the cost
of a permanent second entry point.

**The gap stays declared** and stays on the running list Task 7 carries
forward, now named as 8c's call rather than 8b's or 8c's.

**Fold if ruled otherwise.** Task 1 widens `Parse` to accept an
`io.ReaderAt` with the path-taking form as a thin wrapper, and the
mid-chain I/O error gets its own refusal test.

### R9: the registry's row carries the magic prefix beside its func values

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** One unexported map keyed by an exported `Format` string
type. Its value is a plain struct of func values plus the container's
magic prefix, so registering a container is one map entry rather than an
edit in a sniffer and two dispatch tables. `Read` and `Write` sniff,
look up, and call the func value, which is what makes the dispatch real
rather than decorative. **A struct of func values is not an interface**,
and ruling 14's bar is the absence of an interface.

**Fold if ruled otherwise.** Two parallel maps, one per operation, plus
a third table for the magic prefixes, and the sniff walks the third.

### R10: the instrument compares typed values, and a float difference is a finding

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** Comparison runs per key over decoded JSON values on both
sides: the ledger's row value as decoded, and the read-back tag set
marshalled and decoded back through the same encoder so both sides carry
`encoding/json`'s own number type. Equality is exact. **No tolerance is
introduced**, on any float or any key.

**Why.** Both sides originate as decimal text written by the same
serialization rules, so a representation difference is a real divergence
in the writer or the projection. A tolerance added to make a test pass
would hide exactly the defect the instrument exists to find.

**If a float disagreement appears anyway**, it is a finding for the
conductor and a Task 6 divergence row, never a loosened comparator. The
task reports the field, both values, and the number of digits at which
they part.

**Fold if ruled otherwise.** The comparator gains a per-key epsilon
table with each entry's evidence, and the deviation note records it.

### R11: the instrument fills and projects through `TagSet`'s JSON tags

**Ruled, conductor 2026-09-08, as applied.**

**Applied.** `TagSet`'s 69 members gain JSON tags naming the ledger's own
field names, which are already the values of the `Field` constants
(`tagset.go:15` onward). The instrument then fills a `TagSet` by
unmarshalling the filtered applied-fields object, and projects a
read-back `TagSet` by marshalling it. The four states fall out of the
member types exactly as C2 defines them: `null` leaves a `*string` nil
and a `[]string` nil, `""` allocates a pointer to the empty string, `[]`
produces a non-nil zero-length slice, and any other value is a value.

**No member's tag carries `omitempty`.** Every row in
`corpus/analysis/file_tags.json` carries all 69 keys, absent values
included as explicit nulls, so a tag written the idiomatic way would
drop absent members from the projection and fail Task 5's 69-key count
on all twelve subjects. The tag names the key and nothing else.

**Why.** The alternative is a 69-case fill switch and a 69-case
projection switch inside a test file, whose own defects would be
invisible to every assertion in the pass. The tags also state a real
contract worth having: `TagSet`'s JSON encoding is the ledger's
vocabulary, which is what rung 9 will want when it logs a tag set.

**The hazard, named.** A wrong tag is invisible on its own, because the
fill and the projection cancel: a tag that maps `album` to the artist
member writes ARTIST and reads it back under `album`, and the comparison
agrees while the file on disk is wrong. **The mitigation is mandatory**,
and it joins the table the package already carries. `tagset_test.go:95`
holds `fieldStructName`, a hand-written `Field`-to-member map that
`TestTagSetMembersTieToMapping` (`:173`) grades by name. Task 4 asserts,
member by member through `fieldStructName`, that each `TagSet` member's
JSON tag equals the `Field` constant that names it, which holds the
tags, the members and the mapping rows together by **one literal**
rather than by struct order. A second and cheaper assertion keeps the
positional check: each tag equals `Mapping()`'s row at the same index,
which fails loudly if the declaration order ever drifts from the table.
Both tie back to `TestFieldCountAndNames`, which already grades the
mapping against a hard-coded list from `corpus/README.md`. Without them
this resolution is not safe, and Task 4 does not pass review without
them.

**Fold if ruled otherwise.** The instrument carries its own local
69-field struct with the tags on it, plus an explicit conversion each
way, and the tie test grades the conversion instead. The cost is roughly
140 lines of hand-written assignment whose errors the pass cannot see.

### R12: `Read` stays tag-only, and the picture read is reached through `Parse`

**Ruled, conductor 2026-09-08, as applied.** `Read` stays tag-only. The
picture read is reached through `Parse`'s `Container`. The double parse
is a stated deviation and a declared coverage gap. The combined entry
point is handed to rung 9.

**Applied.** The spec describes two caller entry
points, one that "reads a file's tag set and its pictures" and one that
"writes a tag set and its pictures over a file" (`spec:729-733`), and
puts both in 8a. 8a landed `Read` returning `TagSet` alone
(`read.go:14`), so the read side's picture half is unbuilt and revision
1 of this plan took no position on it. Task 2 widens `Write`; **no task
widens `Read`.** That is the answer, stated rather than left open: `Read`
stays tag-only at 8b, `Container` carries the pictures it parsed (P2),
and a caller wanting both parses through `Parse` and reads
`Container.Pictures`.

**The cost, named.** A caller wanting the tags and the cover from one
file parses it twice, through `Read` and again through `Parse`. That is
a stated deviation from `spec:729-733` and a declared coverage gap Task
7 records, not a defect: the metadata region is bounded and the parse is
cheap, and rung 9 is the first caller that will notice.

**Why not fix it here.** The shapes that fix it are a second exported
read entry point returning both, or an exported projection over a parsed
`Container`. Each is a new exported symbol at a rung whose ruling 14 is
about keeping this package's surface minimal until a second
implementation forces it, and neither has a caller until rung 9. **The
call site should pick the shape.** Task 1 names R12 in its outcomes so a
`diff-reviewer` grades the surface deliberately, and Task 7 drafts the
spec correction.

**Fold if ruled otherwise.** Task 1 adds the sibling entry point
returning the tag set and the pictures from one parse, `Read` becomes a
thin wrapper over it, and the deviation row and the coverage gap both
come out.

---

## Task 0: The staleness pre-task

Model: **`opus`**, **fresh context, zero prior context**, dispatched by
the conductor before any implementer. Reads the rung 8 spec, pass 8a's
HISTORY entry, and this plan against the tree as shipped. **Its output
is blockers with evidence, not fixes**; it writes no code and edits no
file. The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's `spec:`, `INV:`, `RE:` and `docs/HISTORY.md:`
pointers were resolved at `bc4016b`. A citation that has moved since is
a finding for this task rather than a surprise for an implementer.

The checklist is the spec's own (`spec:1197-1243`), adapted to 8b.
**The first three are blockers: a "no" stops the launch.**

1. **Is the lane registry consistent with where this pass intends to
   run, and is the primary tree free?** `docs/STATUS.md`'s lane registry
   is the live-executor check, never `pgrep -f Projects/dubplate`, which
   matches every lane once a worktree exists. *(Expected at drafting:
   lane-1 main tree IDLE, next action "author pass 8b's plan"; lane-2
   worktree IDLE, next action "author pass 5a's plan".)* **Report any
   live executor in this tree, any uncommitted change under
   `internal/engine/tagio` or `internal/harness/analysis`, and any file
   under `internal/engine/tagio`, `internal/harness/` or `corpus/` that
   the lane 2 worktree has modified.** Lane 2's territory is
   `internal/serve` and `web/`; a lane 2 edit inside 8b's territory is a
   blocker. **Confirm in the same item that no lane holds
   `internal/harness/analysis`**, which rung 3 owns and Task 6 writes;
   lane 2's territory does not reach it, and rung 3's would.
2. **Confirm the `code-simplifier` round over 8a's Go landed at
   `0f2b61d`**, "Simplify pass 8a's tagio package and WriteStream",
   which changed one file, `internal/atomicfile/atomicfile.go`, by 1
   insertion and 5 deletions. That discharges forward carry 1
   (`docs/HISTORY.md:366-367`). **The check is that it touched no file
   under `internal/engine/tagio`**, because every `file:line` citation
   in Tasks 1 through 3 was resolved before it ran. A `tagio` file in
   that commit's stat is a blocker for Task 1; anything else is a
   confirmation.
3. **Is `corpus/analysis/file_tags.json` unchanged?** *(Expected:
   `shim_version` `1`, `beets_version` `2.13.1`, 69 `writable_fields`,
   9 `readonly_fields`, twelve rows.)* **A changed ledger is a changed
   conformance target**, and this is 8b's blocker where it was only a
   confirmation at 8a. **Count mechanically; read no row body into
   context.** Cross-check the header's `writable_fields` count and names
   against `corpus/README.md:2378-2400`'s reproduction of the same list.
4. **Do the twelve `corpus/cases/<id>/case.json` files named by the
   ledger's rows all exist, and does each carry an
   `expected.applied_fields` object keyed by item index?** *(Expected:
   twelve directories; `clean-match`'s item `"0"` carries 99 keys.)*
   Report the twelve ids and the per-case key count of item 0, and
   confirm every one of the ledger's 69 `writable_fields` names appears
   in each. **A subject whose case is missing a writable key is a
   blocker for Task 4**, because the spec's claim that the applied
   fields cover all 69 is what makes the instrument viable at all
   (`spec:829-835`).
5. **Is `internal/harness/oracle/testdata/template.flac` still the same
   42 bytes**, and is `corpus/manifest.json` still pinning
   `mediafile==0.17.0`? Every serialization rule in this plan reads from
   that pin. **Do not run any derivation and do not touch the venv.**
6. **Is `internal/engine/tagio`'s exported surface still what 8a's close
   recorded?** *(Expected: `Parse`, `Read`, `Write`, `TagSet`,
   `Mapping()`, `WriteOrder()`, `Serialize`, `EmptyString`,
   `EmptyList`, `Container`, `Block`, `StreamInfo`, `Comment`,
   `VorbisComment`, `KeyWrite`, `Field` and its 69 constants,
   `FieldMapping`, `Style`, `ErrUnsupportedContainer`,
   `ErrMalformedContainer`, `Deviations`;
   `docs/HISTORY.md:82-89`.)* **Report `Write`'s exact current signature
   and `Container`'s exact current field list**, because Tasks 1 and 2
   widen both.
7. **Does `testdata/golden.txt` still record the PICTURE block as type 6,
   declared length 78, with type 3, `image/png`, desc `cover`, zero
   geometry and 32 data bytes?** *(Expected: `golden.txt:123` and
   `:203-210`.)* **Report the arithmetic**: 4 + 4 + 9 + 4 + 5 + 16 + 4 +
   32 = 78, which is what makes the block's byte layout checkable from
   the record alone. A record that no longer adds up is a blocker for
   Task 1.
8. **Is `internal/harness/analysis/adjudication.go`'s
   `AdjudicationEntry` still the field list Task 6 extends**, and does
   `CheckAdjudications` still require `suspected`, `release_truth`,
   `reason`, `adjudicator` and `date` on every row? *(Expected:
   `adjudication.go:49-81` and `:200-214`.)* Report the checker's
   current rule list, because Task 6 adds to it and removes nothing.
9. **Do `TestEngineNeverImportsHarness` and
   `TestEngineNeverReadsTheClock` still walk `internal/engine`
   including test files outside `testdata`?** *(Expected:
   `arch_test.go:141` and `:152`.)* The instrument's local decode types
   and its parameterised date both exist because of that walk.
10. **Has rung 1b or pass 4b changed `internal/config`'s shape or
    `scripts/check.sh`'s leg list?** 8b registers no key either way, so
    this is a confirmation. Report the leg list as it stands and confirm
    `internal/config/keys.go` still carries no rung 8 reserved class.
11. **Is `.gitleaksignore` still tracked and empty?** *(Expected: zero
    bytes, emptied by `3a35558`.)* Report its size. This pass adds no
    fingerprint, so a non-empty file is a note for Task 6 rather than a
    blocker. **Do not quote any fixture literal**, because gitleaks
    scans the tracked worktree and would flag the quotation.
12. **Does `internal/engine/tag/doc.go` still describe `tagio` as
    knowing nothing about matching?** *(Expected: `:8-11`.)* Ruling 3
    preserves that sentence.
13. **Is `go test ./... -count=1`'s wall time still what this plan sizes
    gate runs against?** Run it once and report the total, the
    `internal/engine/tagio` figure, and the `internal/harness/conform`
    figure. *(8a measured `internal/engine/tagio` at 59 tests in 0.121s,
    `go test ./...` at 2m20s, and `bash scripts/check.sh` at 1m26s with
    gitleaks the dominant cost;
    `docs/HISTORY.md:315-322`.)*
14. **Does every `spec:`, `INV:`, `RE:` and `docs/HISTORY.md:` pointer
    this plan carries land on its subject?** **Resolve each distinct
    pointer by grepping for the subject rather than by reading the line
    number**, and report every one that does not. Revision 1 sent six
    `dhowden/tag` citations to the `Eyevinn/mp4ff` row and three more
    one to three rows past their entries; those are corrected in this
    revision and the check exists so the next drift is caught here
    rather than by an implementer. Stale pointers are this build's named
    dominant defect class, so this item is a report and never a fix.

**Deliverable count: 1** (the blocker report, each item with its
evidence and each blocker with the file and line that proves it). No
gate run beyond item 13's timing measurement; the task writes nothing.

**2.0 hooks: neutral.** The pre-task builds no surface.

---

## The `code-simplifier` rounds: one landed, one owed

**Forward carry 1 is discharged.** `code-simplifier` over 8a's Go
landed at `0f2b61d` ahead of this plan, changing one file,
`internal/atomicfile/atomicfile.go`, by 1 insertion and 5 deletions. It
touched no file under `internal/engine/tagio`, so **every `file:line`
this plan cites in Tasks 1 through 3 still resolves where it was
written**, and the premise revision 1 carried about the round moving a
citation is settled. Task 0 item 2 confirms the commit.

**8b's own round is the one this pass owes, and it is batched.** Rung
8's chain commits per task, and CLAUDE.md forbids instructing an
implementer to dispatch an agent, so a conductor-owned simplifier
dispatch cannot reach a point before a chain task's commit. **The two
rules cannot both be honored inside the chain**, and 8a's practice
resolved it the same way: one round over the pass's Go rather than seven
over its tasks. This pass states the departure rather than inheriting
it. The round runs **after Task 6 is accepted and before Task 7's
close**, over the Go Tasks 1 through 6 landed, and the conductor applies
the refinements, runs `bash scripts/check.sh`, and commits them on their
own so Task 7's ledger draft describes a settled tree.
**Ruled, conductor 2026-09-08, as applied (M-e):** one batched round
over 8b's Go after Task 6 and before Task 7, dispatched by the
conductor, with the stated departure from CLAUDE.md's before-committing
rule.

**Budget line:** 0.2M, the flat figure across seven passes, in the
ceiling table above. The 8a round's actual is reported at the close
against that same line.

---

## Task 1: The picture type, the PICTURE block decode, and front-cover selection

Model: `sonnet`. Depends on Task 0. **Not independent**: it widens
`Container`, which Tasks 2 and 3 both build on.

**Files:**
- Create: `internal/engine/tagio/picture.go` + `picture_test.go`
- Modify: `internal/engine/tagio/container.go` (the `Container` type and
  the block walk)

**Interfaces:**
- Consumes: 8a's `Parse`, `Block`, `Container` and
  `ErrMalformedContainer`; `testdata/golden.flac`, `golden.txt`, and
  `testdata/gen_golden_flac.py`, which is the golden's provenance record
  on the input side and holds the picture literal at `:115` and the
  `mf.images` call at `:131`.
- Produces: the exported picture type (P1), `Container`'s picture list
  (P2), and the front-cover selection query (P5). Task 2 writes what
  this task reads; Tasks 4 and 5 use neither.

**Outcomes:**

- P1's type, with the eight RFC 9639 section 8.8 values in the order
  Table 12 defines them, and the block's three declared lengths as
  framing rather than members (`INV:770-776`). Every numeric field and
  every length is `u(32)` **big-endian, unlike VORBIS_COMMENT's
  little-endian framing**, which is the single asymmetry in this block
  and the one a mutation must catch.
- P2: every PICTURE block in the chain decodes, in on-disk order. A
  second PICTURE block is legal and is not a refusal.
- P3: all three declared lengths, the media type's, the description's
  and the data's, are bounds-checked against the block's own payload
  extent before a slice is taken, and a length running past it returns
  `ErrMalformedContainer` with the block's index named.
- P4: the type index degrades to 0 when it is above 20.
- P5: front-cover selection, single candidate first.
- The parse stays inside 8a's C7a contract: the metadata region only,
  and no whole-file read.
- R12's read-side answer, stated in the package documentation rather than
  left to a reader: `Read` returns the tag set alone and `Container`
  carries the pictures, so a caller wanting both parses through `Parse`.
  The doc comment names the double parse as the accepted cost and names
  rung 9 as the caller that may fold it. **No new read entry point
  lands here.**

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- The golden's PICTURE block decodes field by field to the values
  `golden.txt:203-210` records: type index 3, media type `image/png`,
  description `cover`, width, height, depth and colours all 0, and 32
  data bytes beginning with PNG's eight-byte signature. The expectation
  is transcribed from the provenance record, never computed from the
  decoder.
- The block's own arithmetic is asserted: the decoded field widths
  sum to the block's declared length of 78. That is the assertion that
  catches a field read at the wrong offset while every individual value
  still looks plausible.
- A table-driven refusal test covers each of the three length fields
  independently: a media-type length, a description length, and a data
  length each declaring more bytes than the block holds, plus a block
  shorter than 32 bytes, which is the sum of the eight `u(32)` fields
  every picture block carries. **Those eight are not contiguous**: the
  media type string sits between its own length and the description's,
  so 32 bytes is a minimum block length rather than a header to read in
  one struct. Every row asserts `errors.Is(err, ErrMalformedContainer)`
  and asserts the returned container is the zero value.
- A crafted block declaring a data length near the three-byte block
  header's own maximum is refused without allocating, asserted by the
  refusal firing rather than by a memory measurement. This is the
  `dhowden/tag` #116 class, and the criterion is that the length is
  checked before the allocation, verified by the reviewer's read of the
  diff.
- A fixture carrying two PICTURE blocks parses both, in order, and
  neither is a refusal.
- A type-index table covers 0, 3, 20, 21 and 255, asserting the degrade
  to 0 above 20 and no degrade at 20.
- A front-cover selection table covers: no pictures (nothing selected),
  one back cover (selected, the single-candidate rule), a back cover
  then a front cover (the front cover), two front covers (the first),
  and three non-front pictures (the first). Every expected value is
  named in the table row.
- `Container`'s existing fields and every 8a assertion over them are
  unchanged; `go test ./internal/engine/tagio/... -count=1` reports the
  8a test count plus this task's, and the task states both numbers.
- `go.mod` unchanged. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Read the `u(32)` fields little-endian → the golden's decode fails on
   the type index and on the media-type length, and the sum-to-78
   assertion fails.
2. Skip the bounds check on the data-length field → the crafted-length
   refusal row fails, and the task reports whether the process
   allocated.
3. Degrade an out-of-range type index to 3 rather than 0 → the
   type-index table's 21 and 255 rows fail.
4. Return the first picture rather than the first front cover when
   several exist → the back-cover-then-front-cover row fails.
5. Apply the front-cover rule before the single-candidate rule → the
   lone-back-cover row fails, which is the ordering `guess_cover_image`
   fixes and a reimplementation gets backwards.
6. Read the description to the end of the block rather than to its own
   declared length → the golden's decode fails on the description and on
   every field after it: the four geometry fields and the data length.
7. Decode only the first PICTURE block → the two-block fixture fails.

**2.0 hook: advances hook 3** (`S2:363-364`, `spec:1041`). The art blob
recorded at admit depends on an embedded picture being extractable once
rather than at serve time, and the decode is the half that makes
extraction possible. Task 2 lands the hook.

**Prior art:** declines `dhowden/tag`, whose `readPictureBlock` is the
named defect this task's bounds checks exist to avoid, and whose fix PR
#115 has been unmerged since 2026-08-18 (`spec:1061`). Declines
`go-flac/flacpicture` v2 with its base: its endianness is correct and it
offers no delete or replace, and its exported constants are
SCREAMING_SNAKE (`spec:1060`). Reproduces mutagen's and mediafile's
**behavior only**, read at the pinned versions, with no code carried.
Adopts RFC 9639 section 8.8 as the field order's authority.

**Deliverable count: 3** (the picture type, the block decode with its
bounds checks, the front-cover query). Gate green; commit.

---

## Task 2: The PICTURE write, the strip, and the golden's picture bytes

Model: `sonnet`. Depends on Tasks 0 and 1. **Not independent**: it
widens `Write`, which Tasks 4 and 5 call.

**Files:**
- Modify: `internal/engine/tagio/picture.go` + `picture_test.go` (the
  sniff and the encode)
- Modify: `internal/engine/tagio/write.go` + `write_test.go` (the
  argument list and the block planning)
- Modify: `internal/engine/tagio/deviations.go` (P6's stated behavior
  and R3's APNG deviation)

**Interfaces:**
- Consumes: Task 1's picture type and decode; 8a's `Write`, its block
  planning, its PADDING policy and its last-block-flag rule (C10, C11);
  the golden, `golden.txt` and `testdata/gen_golden_flac.py`, the last
  two being the provenance record the byte assertion transcribes its
  input from.
- Produces: `Write`'s widened argument list, which Tasks 4 and 5 call
  with a nil picture set.

**Outcomes:**

- `Write`'s argument list gains the ordered picture set, between the tag
  set and the file mode. **No other 8a signature changes.**
- R1's three caller states: nil leaves the file's pictures alone, a
  non-nil empty set strips every picture, and a non-empty set replaces
  the whole set (P8).
- P7: the media type written is sniffed from the picture's own data and
  overrides whatever the caller declared. R2's recognized set, and R2's
  refusal for a blob matching no row.
- P6: width, height, colour depth and indexed-colour count are written
  as zero, and a rewrite discards the incoming geometry.
- The emitted PICTURE block's framing is big-endian `u(32)` throughout,
  under a re-emitted 4-byte block header carrying the three-byte
  big-endian length and a cleared last-block flag, which is 8a's C11
  asymmetry applied to a new block type.
- The PADDING policy is unchanged: exactly one 8 KiB block, last,
  carrying the last-block flag, regardless of how many pictures the
  write emits.
- P8's ownership transition: PICTURE leaves the preserve-unknown set and
  `planBlocks` stops carrying incoming PICTURE blocks through from
  `c.Blocks`. A rewritten picture replaces its own block in place; a
  picture written into a container carrying none goes after the
  VORBIS_COMMENT block and before the trailing PADDING. **Without the
  transition the writer emits the picture twice and the strip removes
  nothing.**
- `deviations.go` gains four rows, each in the package-documentation
  shape 8a established and each citing the pinned source line it
  diverges from or reproduces: P6's zero geometry, R3's APNG deviation,
  R2's CR2 refusal against the reference's `image/x-canon-cr2`, and
  P3's block-length trust, which covers VORBIS_COMMENT as well as
  PICTURE and names `mutagen/flac.py:350`, `:575` and `:712-724` plus
  the rung 7 intake consequence.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- **The emitted PICTURE block's header and payload equal the golden's,
  byte for byte**, for a write of the golden's own recorded picture into
  a copy of the golden (8a's R4: the byte assertion writes into a copy of
  the golden, never into the template). **The picture handed to `Write`
  is constructed in the test from values hand-transcribed from the
  provenance record**, `golden.txt:71-74` and `gen_golden_flac.py:115`
  and `:131`, and **never from `Parse` or Task 1's decoder**. The
  `diff-reviewer` checks that neither the `want` nor the write input is
  produced by calling into the package. Built the other way the
  assertion is a decode-and-encode round trip, and any symmetric error,
  both sides little-endian or both sides at the same wrong offset,
  reproduces the golden's bytes exactly. **This is the pass's answer to
  risk 5**, and it is the only place in rung 8 where a picture's bytes
  are graded against bytes another implementation produced.
- A sniff table covers each of R2's six rows by its **whole** test
  rather than its leading prefix, with each row's expected media type
  hard-coded from `filetype`'s own `MIME` constant. Four rows exist for
  the cases a prefix-only sniff gets wrong: a RIFF blob carrying `WEBP`
  at 8 and `VP` at 12 (accepted as `image/webp`), a RIFF blob carrying
  `WEBP` at 8 and not `VP` (refused), a TIFF-prefixed blob without `CR`
  at 8 (accepted as `image/tiff`), and a CR2-prefixed blob (refused,
  R2's stated deviation). One further row is a blob matching nothing.
  Every refusing row asserts the write is refused and asserts the target
  file is unchanged on disk.
- A picture whose declared media type disagrees with its bytes (a
  PNG blob declared `image/jpeg`) writes `image/png`, asserted by
  reading the block back. That is the criterion that proves the sniff
  overrides rather than defaults.
- A picture carrying non-zero geometry writes zeros, asserted on the
  block's bytes, and the round trip reports the geometry as zero. That
  is P6, and it is the deviation the close records.
- The three caller states each have a row: a nil set over a file with a
  picture leaves the block untouched **byte for byte**; a non-nil empty
  set over the same file leaves no PICTURE block in the chain; a
  two-picture set over the same file leaves exactly two blocks, in the
  order passed, with both type indexes intact.
- A write with a non-nil empty picture set leaves every VORBIS_COMMENT
  behavior unchanged, asserted against 8a's own merge-contract
  assertions, so the strip is proven not to disturb the tags.
- The whole emitted block chain's order is asserted, for two
  containers rather than one: a write into a copy of the golden, which
  already carries a picture, reads back STREAMINFO, VORBIS_COMMENT,
  PICTURE, PADDING with the picture replaced in place; a write into a
  copy of the conformance template, which carries none, reads back the
  same order with the picture newly placed after VORBIS_COMMENT. The
  chain after any picture write still reads exactly one PADDING block,
  8 KiB, last, and the frame region's hash is equal before and after.
- The renamed file carries the caller's mode, and a refusal never
  half-writes: the no-match row asserts the target's bytes are
  unchanged.
- Every 8a assertion in `write_test.go` still passes unchanged, and the
  task states how many call sites the widened signature touched.
- `go.mod` unchanged. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Write the caller's declared media type rather than the sniffed one →
   the declared-disagrees row fails and the golden byte assertion fails.
2. Write the incoming geometry rather than zero → the golden byte
   assertion fails and the non-zero-geometry row fails.
3. Treat a nil picture set as a strip → the leave-alone row fails,
   reporting the golden's picture gone after a tag-only write.
4. Treat a non-nil empty set as leave-alone → the strip row fails,
   reporting a surviving PICTURE block.
5. Append pictures rather than replacing the set → the two-picture row
   fails, reporting three blocks where two are wanted.
6. Leave PICTURE in `planBlocks`'s preserve-unknown path while also
   emitting the new set → the golden's chain-order row fails, reporting
   two PICTURE blocks where one is wanted, and the strip row fails,
   reporting a surviving block. That is P8's ownership transition, and
   it is the mutation a writer inherits by doing nothing.
7. Emit the picture block's `u(32)` fields little-endian → **the golden
   byte assertion fails and every read-back assertion still passes.**
   That asymmetry is the whole reason the golden exists, and the task
   reports both halves.
8. Accept an unsniffable blob and write an empty media type → the
   no-match refusal row fails.
9. Sniff WEBP on the `RIFF` and `WEBP` prefixes alone, or TIFF without
   the `CR` exclusion → the RIFF-without-`VP` row fails and the CR2 row
   fails. Those two rows exist because a prefix-only sniff passes every
   other row in the table.
10. Emit the PICTURE block before VORBIS_COMMENT, or after the trailing
    PADDING block → the chain-order assertion fails on the template row
    for the first, and on the golden row plus the last-block-flag
    assertion for the second, because a reader that stops at the flag
    never reaches the picture.
11. Drop any one of `deviations.go`'s four rows → **no test fails**, and
    the task reports it unfired with the reason: the note's completeness
    is a reviewer's read against P3, P6, R2 and R3, the same shape 8a's
    Task 8 reported.

**2.0 hook: lands hook 3** (`S2:363-364`, `spec:1041`). With the read at
Task 1 and the write here, an embedded cover is extractable once at
admit rather than at serve time, which is what the 2.0 streamer's art
blob depends on. Rung 9 owns whether the pipeline embeds by default.

**Prior art:** adopts CN's in-repo magic-byte sniff over a dependency,
which is the same recommendation that declines `dhowden/tag` for
container detection (`spec:1061`). Declines `go-flac/flacpicture` v2,
whose absence of a delete or replace is exactly the store-replaces
semantics this task builds (`spec:1060`). Declines `gcottom/audiometa`
on license and adoption (`spec:1065`). Reproduces mediafile's
`FlacImageStorageStyle` and mutagen's `Picture` **behavior only**, read
at the pinned versions.

**Deliverable count: 4** (the media-type sniff, the picture encode, the
three-state write with its strip, the deviation note). Gate green;
commit.

---

## Task 3: The per-format seam registry

Model: `sonnet`. Depends on Tasks 0, 1 and 2. **Not independent**: it
routes `Read` and `Write`, which Task 2 just widened.

**Files:**
- Create: `internal/engine/tagio/seam.go` + `seam_test.go`
- Modify: `internal/engine/tagio/read.go` and `write.go` (the dispatch,
  one call each)

**Interfaces:**
- Consumes: 8a's `ErrUnsupportedContainer`, `Read` and `Write`.
- Produces: the exported format type, the supported set, and the sniff.
  **Rung 7 reads the supported set as its allowlist bound**; no consumer
  is wired at this rung.

**Outcomes:**

- P9's registry: one unexported map keyed by an exported format type,
  its value R9's struct of func values plus the container's magic
  prefix, and **no interface anywhere in the package**, exported or not
  (ruling 14).
- The supported set is exported and returns a copy, on `Mapping()`'s
  own precedent (`docs/HISTORY.md:323-326`): a caller mutating the
  returned slice never reorders the registry.
- The sniff reads the leading magic bytes and returns the format, or
  `ErrUnsupportedContainer` for anything that matches no row. **The
  sentinel is adopted unchanged**, never redefined or reclassified
  (`docs/HISTORY.md:352-357`, and the spec's erratum item 1).
- `Read` and `Write` dispatch through the registry rather than calling
  the FLAC implementation directly, so a container that is registered is
  a container that works and a container that is not is one sentinel
  from either.
- **`Parse` stays the FLAC-specific entry point and deliberately does
  not route**, documented as such in its own doc comment. It returns a
  FLAC `Container`, and a registry whose later entries return other
  containers has nothing to hand it; 8c's absorbed reader and the shadow
  both want the FLAC parse by name. **The bypass is stated so a reader
  finds a decision rather than an oversight**, and the emptied-registry
  test below asserts it. **Ruled, conductor 2026-09-08, as applied
  (m6):** `Parse` stays the FLAC-specific entry point outside the
  registry, with the emptied-registry subtest as its proof.
- The registry carries exactly one entry, FLAC, and the package
  documentation says the MP3 pass adds the second implementation and the
  interface at the same time.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- A sniff table covers: the golden (FLAC), an ID3v2-prefixed blob, an
  MPEG frame-sync blob, an `OggS` blob, an `ftyp`-boxed blob, a
  three-byte file, and an empty file. Every non-FLAC row asserts
  `errors.Is(err, ErrUnsupportedContainer)`, which is rung 7's branch.
- A file beginning `fLaX` sniffs as unsupported, which is the row
  that proves the magic comparison covers all four bytes rather than
  three.
- The supported set has exactly one entry, and a test mutating the
  returned slice asserts a second call returns the original.
- **`Read` and `Write` both route through the registry**, asserted by a
  test that empties the registry for the duration of one subtest and
  observes both entry points return the unsupported sentinel for the
  golden. **The same subtest asserts `Parse` still parses the golden**,
  which is what makes the bypass a graded decision rather than an
  untested gap. Restoring the registry restores all three.
- No interface type is declared in the package, exported or not,
  asserted by the reviewer's read of the diff against ruling 14. The
  registry's value is a struct whose fields are func values.
- The package documentation names what a second registration costs: one
  map entry, and the interface introduced at the same moment as the
  second implementation.
- Every 8a and Task 1 and 2 assertion still passes unchanged.
- `go.mod` unchanged. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Empty the registry permanently → `Read` and `Write` return
   `ErrUnsupportedContainer` for the golden, and every read and write
   test in the package fails. That is the proof the dispatch is real
   rather than decorative, and the task reports the failure count.
2. Compare the magic against a three-byte prefix → the `fLaX` row fails.
3. Return the registry's backing slice from the supported set → the
   caller-mutation row fails on the second call.
4. Return a wrapped error rather than the sentinel for an unknown
   container → every `errors.Is` row fails, which is the branch rung 7
   depends on.
5. Sniff by file extension rather than by bytes → the `fLaX` row and the
   empty-file row both fail, and the task reports that no fixture in the
   package has a filename with a meaningful extension, which is why the
   mutation is observable at all.

**2.0 hook: neutral.** The seam serves rung 7's allowlist bound
(`BS:630-638`, `spec:768-777`), which is a 1.0 surface. Its shape is
judged against the two libraries the follow-on passes register,
`bogem/id3v2` for MP3 and `abema/go-mp4` for MP4, so a registry that
cannot hold them is a finding now rather than at the MP3 pass.

**Prior art:** declines `dhowden/tag` for container detection on its
security bar, and adopts CN's in-repo magic-byte sniff instead
(`spec:1061`). Declines an interface until the second implementation
arrives (ruling 14). Names `bogem/id3v2` (conditionally adopted for the
MP3 pass, pinned) and `abema/go-mp4` (adopted as the MP4 pass's box
layer) as the two future registrants the registry's shape is measured
against (`spec:1062-1063`). Adopts pinned mediafile's own organization
as the argument for keying by tag standard: `mediafile/storage/` holds
exactly `flac.py`, `mp3.py`, `mp4.py`, `vorbis.py`, `ape.py` and
`afs.py` (`spec:750-755`).

**Deliverable count: 3** (the registry with its func-value dispatch, the
exported supported set, the magic-byte sniff). Gate green; commit.

---

## Task 4: The conformance instrument's inputs

Model: `sonnet`. Depends on Tasks 0 and 2. **Not independent**: it calls
the widened `Write` and it edits `tagset.go`, which nothing else in this
pass edits but everything else in the package reads.

This task lands steps 1 through 3 of P10 and proves them on their own.
Task 5 lands steps 4 and 5. **The split is deliberate**: an input-side
defect and a writer defect are indistinguishable once they are graded
together, and risk 1 is a value model wrong on twenty keys at once.

**Files:**
- Modify: `internal/engine/tagio/tagset.go` (JSON tags, R11) +
  `tagset_test.go` (the tie assertion)
- Create: `internal/engine/tagio/conform_test.go` (the input half). **The
  conformance test function is named `TestConformanceLedgerSubjects`**,
  because rung 15's launch-gate re-run selects it with `-run
  Conformance` and a pattern matching nothing exits 0 with "no tests to
  run". Any helper test in the file that is not part of that run carries
  a name the pattern does not match.

**Interfaces:**
- Consumes: `TagSet`, `Mapping()`, `Write`;
  `corpus/analysis/file_tags.json`, the twelve
  `corpus/cases/<id>/case.json` files, and
  `internal/harness/oracle/testdata/template.flac`, all read-only.
- Produces: the per-subject filled `TagSet` and the per-subject scratch
  file Task 5 reads back.

**Outcomes:**

- R11's JSON tags on all 69 `TagSet` members, each naming the ledger's
  own field name, with the tie assertion R11 makes mandatory.
- R5's path constants, one block, each naming the rung that owns the
  file.
- Local decode types for the ledger and for a case, declared in the test
  file, because the arch walk forbids importing
  `internal/harness/analysis` (`spec:692-697`).
- R6's subject derivation from the ledger's rows.
- Step 2: each subject's applied fields taken from its committed case at
  the row's item index, **filtered to the ledger's own
  `writable_fields` header** rather than to a second hard-coded list,
  which is the rule the shim follows (`spec:843-847`).
- Step 3: the filtered object fills a `TagSet` under the four-state rule,
  and the filled set is written through `Write` into a **per-subject**
  scratch copy of the template under `t.TempDir()`. The per-subject path
  is a rule, not a detail: a shared path lets one subject's write leak
  into the next subject's read-back
  (`internal/harness/oracle/derive_file_tags.py:132-138`).
- The picture argument is nil at every call: the instrument grades tags,
  and the corpus grades no picture behavior at all (`INV:432-441`).
- Nothing is written under `corpus/` or `internal/harness/`. The
  template is copied out; the copy is what is written.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- **The tie assertion, in two halves.** By name: for each entry in
  `fieldStructName` (`tagset_test.go:95`), the named `TagSet` member's
  JSON tag equals the `Field` constant that names it, which holds the
  tags, the members and the mapping rows to one literal. By position: for
  each index, the member's JSON tag equals `Mapping()`'s row at that
  index, and the member count equals 69, which fails loudly if the
  declaration order ever drifts. R11 does not pass review without both,
  and the reviewer checks that each reads the tag through reflection
  rather than restating the list.
- No member's JSON tag carries `omitempty`, asserted over all 69
  tags, because every ledger row carries all 69 keys with absent values
  as explicit nulls and an omitted member would fail Task 5's count on
  every subject.
- The subject list derives from the ledger's rows, its length is
  asserted against the ledger's twelve, and each `case_id` is asserted
  to resolve to a committed case directory. The twelve ids are reported
  in the task's return so the conductor can check them against
  `corpus/README.md:2182-2192`.
- The filter drops every non-writable key, asserted by naming at
  least four of them that a case carries and the ledger does not:
  `data_source`, `album_id`, `id` and `bitrate`. The count after
  filtering is asserted to be 69 for every subject, which is also the
  check that a case is not missing a writable key.
- A four-state table over hand-transcribed values from one named
  subject, `clean-match` at item 0, asserting: a string value
  (`album` is `Rumours`), an empty string (`acoustid_fingerprint` is
  present and empty), a list value (`albumartists` holds one entry), an
  empty list (`arrangers` is present and empty), a JSON null mapping to
  absent, a numeric zero mapping to a present zero (`bpm`), and a false
  boolean mapping to a present false (`comp`). **Every expected value is
  transcribed from `corpus/cases/clean-match/case.json`**, never
  computed.
- The write goes to a per-subject path under `t.TempDir()`, asserted by
  the paths being distinct across the twelve, and the committed template
  is asserted unchanged in size after the run.
- The instrument imports nothing under `internal/harness`, asserted by
  the existing arch test staying green, and reads no clock.
- `bash scripts/check.sh` green, with the arch test's result named
  explicitly in the task's report.

**Named mutations this task reports:**

1. Change one `TagSet` member's JSON tag → the tie assertion fails,
   naming the member and the mapping row. **This is R11's safety net**,
   and the task reports the exact failure message.
2. Add `omitempty` to one member's tag → the no-`omitempty` assertion
   fails, and Task 5's 69-key count would have failed on every subject
   carrying that key as null.
3. Fill from the case's applied fields without the writable filter → the
   count assertion fails at 99 rather than 69, and the named-leftovers
   row fails on `data_source`.
4. Treat a JSON null as an empty string → the four-state row fails on
   the null-valued key.
5. Treat a JSON `[]` as absent → the `arrangers` row fails.
6. Treat a numeric 0 as absent → the `bpm` row fails, which is the trap
   `safe_cast` and Python truthiness make easy to fall into.
7. Share one scratch path across the twelve subjects → the distinct-path
   assertion fails. **Task 5 carries the same mutation** as a value-level
   leak, and the deliberate redundancy is what makes the leak observable
   at both altitudes.
8. Hard-code the twelve subject ids rather than reading the ledger's
   rows → **no test fails**, and the task reports it unfired with the
   reason: a hard-coded list in the same order passes every assertion
   here, so R6 is held by the reviewer's read of the diff.

**2.0 hooks: neutral.** The instrument grades a 1.0 surface. The JSON
tags are a contract rung 9 will read when it logs a tag set, and that is
a convenience rather than a hook.

**Prior art:** adopts `internal/harness/oracle/derive_file_tags.py` as
the procedure to reproduce, step for step, and its per-subject scratch
path as a rule. Declines shelling out to the venv, which `spec:874-879`
already settles: the gate never requires uv or beets by design. Declines
a second conformance corpus, which the same section settles.

**Deliverable count: 4** (the JSON tags with their tie assertion, the
path constants and local decode types, the subject derivation with its
filter, the four-state fill and the per-subject write). Gate green;
commit.

---

## Task 5: The conformance grading, and the proof the instrument can fail

Model: **`opus`**. Depends on Tasks 0, 2 and 4.

This is the pass's largest single item and the one risk 1 exists for. It
lands steps 4 and 5 of P10, and it lands the exit gate's requirement
that the instrument be shown able to fail (`spec:1134`).

**Files:**
- Modify: `internal/engine/tagio/conform_test.go` (the grading half)

**Interfaces:**
- Consumes: Task 4's filled subjects and scratch files; `tagio.Read`
  (P11); the ledger's rows.
- Produces: the per-subject, per-key comparison result Task 6's emitter
  turns into a staged row.

**Outcomes:**

- Step 4: each subject's scratch file is read back through `tagio.Read`,
  and the resulting `TagSet` is projected to the ledger's JSON shape
  through R11's tags. **No second projection exists** (P11).
- Step 5: all 69 writable keys are compared for all twelve subjects,
  under R10's exact typed comparison. **The nine read-only keys are
  never compared per subject**, and the ledger's `template` block is
  never a per-subject expectation
  (`corpus/README.md:2401-2412`).
- A failure names the subject, the key, the ledger's value and the
  read-back value, and it names **every** disagreement rather than the
  first, so one run produces the whole divergence set.
- **The comparator is proven able to fail**, by its own table-driven test
  over synthetic rows rather than by a one-off demonstration: an
  agreeing row, a row differing on a string, a row differing on a list's
  order, a row where one side is null and the other an empty string, a
  row differing on a float's last digit, and a row missing a key
  entirely. Each names the expected failure text.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- **All twelve subjects agree on all 69 keys**, or every disagreement is
  reported with its subject, key and both values, and the task's return
  carries the full list for the conductor. **A disagreement is not a
  test to relax.** The gate stays red until the conductor rules, and the
  task reports rather than adjusts.
- The comparison covers exactly 69 keys per subject, asserted by count,
  so a key silently absent from a row cannot pass by not being compared.
- A key present in the read-back and absent from the ledger row is a
  failure, and so is the converse. Both directions have a row in the
  comparator's own test.
- The comparator's own test fires on every synthetic row that should
  fire, and the task reports each row's observed failure text. This is
  the exit gate's "proof that the instrument can fail", discharged as a
  standing test rather than as a mutation that leaves no trace
  (`spec:1134`).
- The nine read-only key names are asserted absent from every
  comparison, by name, so a future reader cannot mistake their omission
  for an oversight.
- No expectation is derived from the code under test: the ledger is
  the expectation, the comparator's synthetic rows carry hard-coded
  values, and the reviewer checks that no `want` is computed by calling
  `tagio.Read`, `Serialize` or the projection.
- The test count and the wall time for `go test
  ./internal/engine/tagio/... -run Conformance -count=1` are both
  reported, along with the full package's wall time, because rung 15
  re-runs exactly this and a pattern matching nothing exits 0. A count
  of zero is the failure this criterion exists to catch.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Change one expected value in the comparator's **own synthetic
   fixture** → the comparator's table fires. **The corpus ledger is
   never mutated**, and the task states that it read
   `corpus/analysis/file_tags.json` and wrote nothing under `corpus/`.
2. Compare only the keys the ledger row happens to carry, rather than
   all 69 → the count assertion fails and the missing-key row fails.
3. Skip the write and read the untouched template copy back → every
   subject diverges on every non-empty key; the task reports the total
   divergence count as the instrument's own sensitivity figure.
4. Treat a null and an empty string as equal → the null-versus-empty row
   fails, which is the distinction C2 and C5 both turn on.
5. Compare a list ignoring order → the list-order row fails.
6. Share one scratch path across the twelve subjects → subject N's
   values appear in subject N+1's read-back, and the task reports which
   keys leaked. Task 4 carries the path-level half of this mutation.
7. Introduce a float tolerance → **no test fails**, and the task reports
   it unfired with the reason: R10 forbids a tolerance, and the
   criterion that holds it is the reviewer's read of the comparator's
   diff.

**2.0 hooks: neutral.** The grading proves a 1.0 surface. It is what
rung 15's launch gate re-runs (`spec:1020-1023`).

**Prior art:** adopts `derive_file_tags.py`'s read-back-through-an-
independent-parse discipline: the ledger records what a fresh
`MediaFile` parse reported, never what the writer believed
(`derive_file_tags.py:151-157`). Declines a venv-shelling conformance
run (`spec:874-879`). Declines any comparison helper from outside the
package, since the arch walk forbids the harness's own.

**Deliverable count: 3** (the read-back projection through `tagio.Read`,
the 69-key comparison over twelve subjects, the comparator's
proof-of-failure test). Gate green; commit.

---

## Task 6: The divergence record and its schema addition

Model: `sonnet`. Depends on Tasks 0 and 5. **Independent of Tasks 1
through 3 by files**, and the only task in this pass that writes outside
`internal/engine/tagio`.

This closes the gap `RE:192` recorded: rung 3 has a tier 2 pending-row
escape and rung 8 has none (`spec:1303-1305`).

**Files:**
- Modify: `internal/harness/analysis/adjudication.go` +
  `adjudication_test.go`
- Create: `internal/engine/tagio/testdata/divergence-example.json`
- Modify: `internal/engine/tagio/conform_test.go` (the emitter)

**Interfaces, split by package**, because the arch walk makes the split
load-bearing rather than tidy:
- The `internal/harness/analysis` half consumes `AdjudicationEntry`,
  `Pending` and `AgentAdjudicator` directly, and extends the first.
- **The `internal/engine/tagio` half consumes none of them.**
  `TestEngineNeverImportsHarness` walks `internal/engine` and skips only
  a `testdata` subtree, so it covers `_test.go` files
  (`arch_test.go:37-42`, `:141`), and an import of `analysis` from
  `conform_test.go` turns the gate red. The emitter works over a local
  struct (R7 point 2) and spells `"pending-adjudication"` and `"agent"`
  as literals (`adjudication.go:18`, `:30`). **Their only tie to the
  constants is `divergence-example.json`**, which both packages grade
  against.
- Consumes, in both halves: Task 5's comparison result.
- Produces: the `applied_field` block on `AdjudicationEntry`, the
  committed example row, and the instrument's emitter.

**Outcomes:**

- R7's schema addition: an optional `applied_field` block on
  `AdjudicationEntry`, carrying the item index, the field name, the
  ledger's value, the writer's read-back value, and the `tagio` commit.
  **Additive only.** No existing field's requirement changes, no
  existing rule is removed, and rung 3's and rung 4's rows validate
  unchanged.
- The checker's new rules, each reporting rather than returning at the
  first failure, matching the file's existing style: the block is legal
  only on a case-subject row; when present, the field name, the item
  index and the commit are all required; and a row carrying the block
  carries no matcher stamp, which the case-subject rule already
  enforces.
- `release_truth` stays required and stays meaningful for this row
  class: an applied-field row fills it with the ledger's own value and
  its citation, because for a writer divergence the ledger is the truth.
- R7's committed example row, with invented values and a fixed date, as
  the single literal both sides grade against.
- The instrument's emitter, over a local struct, taking the date and the
  commit as parameters and stamping neither (the clock walk). On a
  divergence the instrument writes the row under `t.TempDir()` and
  prints it in the failure message, naming the two fields a human fills
  and naming `corpus/analysis/adjudications.json` as its destination.
- The instrument writes nothing under `corpus/`. The fold is the
  conductor's, into a hand-written ledger
  (`corpus/README.md:2255-2258`).
- The package documentation on the new block states the route in one
  place: staged by the instrument, folded by the conductor, adjudicated
  by the owner, and blocking at rung 15 while it stays pending.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- The committed example decodes into `AdjudicationEntry` and passes
  `CheckAdjudications`, asserted in `internal/harness/analysis`'s own
  test, which reads the example by path and **supplies its own
  `caseIDs` slice**, since `CheckAdjudications` resolves a case-subject
  row against the caller's list (`adjudication.go:156-159`, `:183-185`)
  rather than against the corpus.
- The instrument's emitter reproduces the example byte for byte
  from a synthetic divergence and the example's own date and commit,
  asserted in `internal/engine/tagio`. One literal, two graders.
- A malformed-row table in `internal/harness/analysis` covers: an
  applied-field block on a shadow-subject row, a block with an empty
  field name, a block with an empty commit, and a block on a row whose
  disposition is not pending with the agent adjudicator. Each row
  asserts the specific error text.
- Every existing adjudication test passes unchanged, and the task
  states the count. `corpus/analysis/adjudications.json`'s existing rows
  validate unchanged, asserted by the gate's own
  `internal/harness/analysis` check staying green.
- The instrument emits a row only when a subject disagrees, asserted by
  a synthetic agreeing run producing none.
- `corpus/analysis/adjudications.json` is unchanged by this task,
  unless Task 5 found a real divergence, in which case the staged row is
  handed to the conductor in the task return and **the implementer does
  not fold it**.
- `bash scripts/check.sh` green, and gitleaks' result on the new tracked
  JSON file is named.

**Named mutations this task reports:**

1. Emit a disposition other than pending → the byte-equality assertion
   against `divergence-example.json` fails in `tagio`, and the
   malformed-row table fires in `internal/harness/analysis`. **The
   emitter cannot call `CheckAdjudications`**: the arch walk forbids
   `internal/engine` importing the harness, which is why the example row
   is the two packages' only shared literal.
2. Emit the applied-field block on a shadow-subject row → the new check
   fires with its own message.
3. Drop the commit from the block → the new check fires.
4. Change the emitted row's key order or indentation → the byte-equality
   assertion against the example fails, which is what keeps the two
   graders on one literal.
5. Emit a row for an agreeing subject → the no-divergence-no-row
   assertion fails.
6. Make the block legal on a shadow row by relaxing the new rule →
   the malformed-row table's first row fails.
7. Stamp the date from the clock inside the instrument → **the arch
   test fails**, `TestEngineNeverReadsTheClock` naming the file. The task
   reports it as fired, since it is the constraint R7's parameterisation
   exists for.

**2.0 hooks: neutral.** The divergence route is a 1.0 launch-gate
mechanism.

**Prior art:** adopts rung 4's pending-row pattern whole: an engine-local
row type, a marshaller outside the engine that sets disposition and
adjudicator as constants rather than reading them off the row, and
validation before the bytes reach disk
(`internal/engine/shadow/run.go:349-385`,
`cmd/dubplate-server/shadow.go:151-186`). Declines a new
`dubplate-server` subcommand, which `spec:928-929` already rules out for
this rung.

**Deliverable count: 3** (the schema addition with its checker rules,
the committed example row, the instrument's emitter). Gate green;
commit.

---

## Task 7: The pass close, the 8b to 8c handoff, and the spec corrections

Model: `sonnet`. Depends on Tasks 0 through 6.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit; **the implementer never commits a ledger**)
- Draft, handed to the conductor in the task return: the spec
  corrections, the deviation summary, and the 8c handoff

**Outcomes:**

- The HISTORY entry in the house shape: what landed, what the gate
  caught, and what a later pass would be wrong to rediscover.
- **The P1 through P12 coverage table**, each item with its Go site,
  what proves it, and, for a rule no test can reach, the stated reason.
- **The twelve Resolutions listed with their markers**, each retired
  from provisional to ruled or folded, each naming the file and symbol
  the marker sits on, and each naming what would have changed had it
  been ruled the other way.
- **The declared coverage gaps recorded as gaps, not as passing checks**,
  on one criterion: a surface 8b itself builds that no test in this pass
  can reach. The expected list, to be corrected against what execution
  found:
  1. **APNG is reported as PNG** (R3). A stated deviation, held by
     `deviations.go` and by this entry.
  2. **The picture set has no oracle-produced multi-picture fixture.**
     The golden carries one picture, so the two-picture and
     front-cover-selection rows are all hand-built.
  3. **The sniff covers six formats and the reference covers
     seventeen** (`filetype/types/__init__.py:15-32`). A cover in a
     format outside the six is refused where the reference would have
     written it.
  4. **No real library file has been graded.** Every conformance subject
     writes into a 42-byte silent template; a real beets-written library
     file is first written at rung 9. Inherited from 8a's gap 6.
  5. **The `io.ReaderAt` gap stays open** and is now 8c's call alone
     (R8), moved from "8b's or 8c's".
  6. **`Read` returns no pictures** (R12), so a caller wanting the tags
     and the cover parses the file twice. A stated deviation from
     `spec:729-733`, handed to rung 9 as the first caller.
  7. **A CR2-prefixed blob is refused** where the reference writes it as
     `image/x-canon-cr2` (R2). A stated deviation, held by
     `deviations.go`.
  8. **A block whose declared length disagrees with its contents is
     refused** where mutagen parses it anyway (P3). A stated deviation
     covering VORBIS_COMMENT as well as PICTURE, and the one gap this
     pass records against 8a's landed code as well as its own.
  9. **The divergence schema's field name is not checked against the 69
     writable keys**, because the checker does not load the ledger. A
     row naming a field that is not writable passes.
  10. Anything execution adds.
  **The count is stated and joined to the running total of forty-three**
  (rung 3's twenty-one, rung 4a's nine, rung 4b's four, rung 1b's two,
  rung 8a's seven; `docs/STATUS.md`'s carried list), so rung 15 does not
  re-count.
- **The spec corrections the pass owes**, drafted for the conductor:
  - **The erratum's own pointer.** Item 1 cites `spec:1103` for the
    clause "the seam registry with its supported set, its func-value
    dispatch, and its unsupported-container sentinel". That clause is at
    `spec:1134`, the 8b pass-split row; line 1103 today sits inside the
    F1 launch clause block quote, and the erratum's insertion into the
    header is what moved it. `docs/HISTORY.md:356` already flags the
    pointer with "(corrected, below)", so the correction the pass owes is
    the repoint itself, plus any further wording the clause needs once
    the registry has landed and the sentinel is confirmed adopted
    unchanged.
  - Any correction Tasks 1 and 2 forced on the spec's PICTURE section,
    which was written from INV's read of the pinned source rather than
    from a Go implementation. **A disagreement between the spec's
    PICTURE description and what the golden's bytes actually hold is a
    correction, never a silent adjustment.**
  - The `spec:1181-1196` ceiling table's 8b line, corrected to this
    plan's 3.25M with the derivation, and **8c's 1M line re-derived from
    this pass's actual**, which the spec's own clause schedules for
    exactly this moment.
  - CLAUDE.md's gate sentence names the tracked Python files by role;
    this pass adds none, so the sentence stands. State that it was
    checked.
  - The `docs/HISTORY.md:442-447` forward-carry-5 correction the spec's
    errata section already names, if 8a's close did not apply it.
- **The 8b to 8c handoff**, naming what 8c inherits: the widened
  `Container` and `Write` signatures its absorbed reader must satisfy,
  the registry its sniff routes through and the `Parse` bypass that
  deliberately does not route (m6), the instrument that must stay green
  as the shadow's reader moves, **named as
  `TestConformanceLedgerSubjects` beside the `-run Conformance` command
  rung 15 uses**, the `io.ReaderAt` call that is now 8c's alone, R12's
  read-side answer and the double parse it accepts, and the fact that
  the tightened refusal set is unchanged by this pass so 8c's
  capture-path work is written against 8a's rules.
- **The carries this pass does not decide**, so the close does not have
  to rediscover them. Forward carry 3, fork B, stays open and owed to
  Geoff. **Forward carry 4, the group non-FLAC question**
  (`docs/HISTORY.md:375-377`), stays carried and is not 8b's: this pass
  builds one container's picture layer and the seam that would hold a
  second, and neither answers whether a member's non-FLAC library
  gets a rung 8-shaped pass.
- The deviation summary rung 15 reads: `deviations.go`'s rows as they
  stand after Task 2, named, so the launch gate reads a statement rather
  than a file it has to interpret.
- The evidence, not a bare "green": the golden's PICTURE block
  arithmetic; the mutation table across Tasks 1 through 6 with every
  outcome, including the deliberately unfired ones; the twelve subjects'
  agreement result key by key in summary form; the measured wall times
  for the conformance run, `internal/engine/tagio`, `go test ./...` and
  `bash scripts/check.sh`; and confirmation of which existing suites
  changed, since this pass touches `internal/harness/analysis` as well
  as `tagio`.
- Both budgets scored: tokens against the 3.25M ceiling, and attended
  time as two counts. The target is zero planning misses and zero
  execution sittings.

**Acceptance criteria:**

- Every gap in the list is a gap this pass created or inherited, with
  the task and the mutation that names it.
- The mutation table has a row per named mutation across every task, and
  every unfired mutation carries its reason.
- Every Resolution is retired, and a folded one names the edit that
  followed.
- `docs/STATUS.md` stays present tense and under 60 lines after the
  edit, with the pass ledger moved to HISTORY.
- The 8c handoff names what 8c may assume and what it must not, in
  the shape 8a's handoff used, so 8c's plan reads one section rather
  than the whole entry.
- No ledger is committed by the implementer.

**2.0 hook: neutral.** The close records hook 3 as landed by this pass,
hooks 13 and 14 as 8a's, and hook 11 as enabled at 8a.

**Deliverable count: 4** (the HISTORY entry, the STATUS edit, the
drafted spec corrections, the 8c handoff).

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain
(`~/.claude/workflows/pass-execute.js`) per task, **strictly
sequential**. Nothing here runs in parallel, and no second executor
enters this tree. Every task runs `cairn-implementer`, **instructed in
its dispatch to read and follow `.claude/agents/dubplate-implementer.md`**,
then `diff-reviewer` (`claude-opus-5`), then the gate, with one
re-dispatch on `fix` and a second `fix` as the conductor's decision. The
conductor dispatches **one batched `code-simplifier` round** over 8b's
Go after Task 6 is accepted and before Task 7's close, per the
simplifier section above, and owns every ledger edit, the lane registry,
the drafted corrections and the batch commit. **The runaway guard is
armed at launch. Subagents never dispatch agents.**

**Independence.** Task 6 is independent of Tasks 1 through 3 by files.
Every other pair is contended: Tasks 1, 2 and 3 converge on
`container.go`, `write.go` and `read.go`, and Tasks 4 through 6 build on
the `Write` signature Task 2 widens. **This pass therefore marks no
task pair for parallel execution**, and the sequential chain is the mode
regardless, because one tree runs one executor and the module gate runs
over `./...`.

**Models.** `sonnet` by default. **One named upshift to `opus`:**

- **Task 5, the conformance grading.** It is the pass the spec split off
  to catch a value model wrong on twenty keys at once
  (`spec:1247-1252`), its comparator has to be right about the
  null-versus-empty distinction that C2 and C5 both turn on, and its own
  ability to fail is exit-gate content. A comparator that is subtly
  permissive ships a green pass that proves nothing.

**Task 0 also runs at `opus`**, fresh context: a staleness check that
misses a moved citation costs more than the model difference.

**One named runner-up**, upshifted on a condition rather than by policy.
**Task 2**, the PICTURE write, if Task 1's review returns `fix`. Task 2
composes the sniff, the encode, the three-state semantics and the golden
byte assertion in one place, and its correctness rests on the decode
Task 1 landed. RFC 9639 section 8.8 and `golden.txt` fully specify the
byte layout, so Task 2 stays `sonnet` unless that evidence weakens.

**Review.** Rung 8 is not one of the three full adversarial-cycle rungs
(5, 12 and 16), so review scales to the in-chain `diff-reviewer` per
task plus the pass-end gate (`spec:1323-1331`). The golden's byte
assertions carry the defects the conformance run cannot see, which is
why Tasks 1 and 2 precede Tasks 4 and 5 rather than riding beside them.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- **The batched `code-simplifier` round over 8b's Go runs after Task 6
  and before Task 7**, and its result is committed on its own. Forward
  carry 1's round over 8a's Go already landed at `0f2b61d`.
- Task 1 → Task 2 → Task 3. The decode, then the write, then the
  dispatch that routes both.
- Task 2 → Task 4. The instrument calls the widened `Write`.
- Task 4 → Task 5 → Task 6.
- Tasks 0 through 6 → Task 7.

**Where this pass splits, if it splits.** Task 0 plus seven tasks is
well inside the spec's fifteen-task trigger (`spec:1144-1151`), so **the
recommendation is to keep one pass.** The cut point, if the Task 2
re-derivation puts the ceiling past 4M, is the **Task 3 / Task 4
boundary**: Tasks 0 to 3 stay 8b and Tasks 4 to 7 become 8b-ii. Both
halves ship something a consumer can use, which is the property 8a's own
split analysis found missing in its alternative. The argument against
taking it is that hook 3 and the grading are the two things the spec
put in this pass together, and a full close plus a full staleness
pre-task between them buys nothing the checkpoint does not.

**Budget:** 3.25M ceiling, re-derived after Task 2. Rung 8's ceilings
stand at 8.75M against the spec's 8M reservation, with 8a's 0.9M
underspend as the offset and 8c's line re-derived at this pass's close.
Flag at 80%, which is **2.6M**, routed to the conductor. STATUS after
every four tasks, so after Task 4 and at the close. Spend is scored at
close against the ceiling and against **zero attended interactions**,
which is this pass's target. **A divergence found by Task 5 is not an
execution sitting**: it is a staged row that joins the standing
adjudication stream.
