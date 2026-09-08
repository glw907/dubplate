# Rung 8, Pass 8a (The Writer Core and the Reader) Implementation Plan

**Revision 2, 2026-09-07.** Folds the two-lens adversarial review at
`research/2026-09-07-rung-8a-plan-review.md`: three blockers, seven
majors and eleven minors, every one re-verified against the tree at
`47e80cc`, the pinned venv or RFC 9639 before folding, and none
rejected. The blockers were the golden's own input recorded nowhere, the
field-name rule sitting in a task with no input path, and a ceiling
derivation that divided by ten where the measurement is over nine. Seven
folds are applied as the recommendation and marked **provisional,
conductor ruling owed**; each names what changes if the ruling comes back
the other way. Seventeen `spec:` pointers that opened on a different
subject are corrected in place, and the verification stamp moves from
`8031e3b` to `47e80cc`. The ceiling moves from 4M to 4.5M with the
arithmetic shown. The task count is unchanged at Task 0 plus nine. The
conductor's eight rulings of 2026-09-08 are folded throughout, each
marked **ruled, conductor 2026-09-08** at its site; Fork B stays open and
owed to Geoff.

**Revision 1, 2026-09-07.** Drafted against the rung 8 spec at
**revision 2** (`docs/superpowers/specs/2026-09-07-rung-8-tag-io.md`,
landed at `a7e10c4`), whose fourteen conductor rulings are applied and
whose four owner rulings are ruled. One product fork stays open, fork B,
a FLAC-only 1.0 with the MP3 pass immediately after cutover. Fork B is
applied provisionally by the conductor and is owed to Geoff at the next
sitting. **8a's surface does not change under either answer**
(`spec:186-190`), so no task in this plan carries a fold marker for it.

**What this pass is.** 8a is the first of rung 8's three passes
(`spec:1092-1121`). It builds the FLAC container reader, the 69-key
value model with every serialization rule, the exported read projection,
the writer with its merge contract, the streaming write in
`internal/atomicfile`, and the golden container that anchors the writer's
bytes to bytes another implementation produced. **8b builds PICTURE, the
per-format seam, and the conformance instrument. 8c absorbs the shadow's
provisional reader and repoints its call sites.** Neither is in scope
here, in any part. A task that believes it needs the seam registry, the
picture decode, the twelve ledger subjects, or any file under
`internal/engine/shadow` stops and reports.

**Ten tasks, under the spec's split trigger.** `spec:1116-1121` says to
split 8a rather than its tasks if the plan exceeds fifteen tasks, with
the cut point between the container writer and the 69-key value model.
This plan lands **Task 0 plus nine tasks**, so **8a stays one pass** and
the conductor takes no split at authoring. The cut point is named anyway
in the execution model, because a re-derivation past the ceiling at the
Task 3 checkpoint is where the call would be taken.

**Assertion discipline is the pass's own risk control.** Risk 2b
(`spec:1230-1237`) is a defect the later conformance instrument cannot
see by construction: a wrong endianness or a skipped write style agrees
with all twelve ledger subjects on all 69 keys and ships a library no
other player reads correctly. Everything in this plan that asserts on
emitted bytes exists for that risk. A writer that reaches 8b with wrong
bytes has already passed the pass that was supposed to catch it.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section, the pass-execute chain per task, sequential. The implementer
> is `cairn-implementer`, instructed to read and follow
> `.claude/agents/dubplate-implementer.md` before anything else. The
> conductor consumes reports and never reads diffs inline. Subagents
> never dispatch other agents; `code-simplifier` is the conductor's
> dispatch after a task's review accepts. The conductor owns the ledgers
> (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`), the lane registry,
> the contested-surface rows, and the batch commit of anything a task
> drafts into them. Rung 8 is not an adversarial-cycle rung (those are 5,
> 12 and 16), so review scales to the in-chain `diff-reviewer`
> (`claude-opus-5`) per task plus the pass-end gate (`spec:1292-1299`).
> **This pass runs in the MAIN TREE.** No task runs from a worktree.

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/superpowers/specs/2026-09-07-rung-8-tag-io.md`, revision 2,
   in full.** The pass-split table's 8a row (`:1097-1102`) is this plan's
   scope statement. The sections this pass implements are "Conductor
   rulings, applied" (`:192-289`), "The write side of beets, as a
   contract" (`:291-415`), "The container contract" (`:417-473`), "What
   the writer guarantees" (`:475-575`), "The sixteen the ledger cannot
   grade" (`:577-606`), "The package and its types" (`:653-715`), "The
   write path" (`:753-785`), "Config keys" (`:850-866`), "What rung 8
   does not build" (`:868-890`), and the staleness pre-task
   (`:1166-1212`).
2. **`docs/superpowers/research/2026-09-07-rung-8-spec-inventory.md`
   (INV), sections 1, 2, 5 and 7, plus its twelve surprises.** Section
   2's 69-key table (`INV:315-385`) is the row-by-row source for the
   mapping. Section 2.1 (`INV:219-308`) is the mechanism behind every
   serialization rule. Section 5 (`INV:715-806`) is the format. Section 7
   (`INV:884-956`) is the atomic write path as landed. **Do not
   re-derive what INV settles; cite it, and verify a line you carry
   forward.**
3. **`docs/superpowers/research/2026-09-07-rung-8-container-needs.md`
   (CN)**, its library table and its "Forward hooks 13 and 14" section.
   CN's per-format recommendations belong to 8b and to the follow-on
   passes; 8a reads it for the FLAC-only launch argument and the hook 13
   tag names.
4. **`docs/superpowers/plans/2026-09-05-rung-4a-shadow-runner.md`**, its
   Task 4 (`:2338-2775`), which is the closest shape in the repo to Tasks
   3 and 6 below, and whose corrected STREAMINFO offset table is the
   layout this pass reuses. **Read it for that shape alone.** This plan's
   own Global Constraints stand on their own and delegate nothing: 4a's
   list carries rules that are false here, among them "`internal/config`
   is rung 4's for the length of this pass" and the shadow's
   `import.lock` rule.
5. **`corpus/README.md`'s handoff section** (`:2354` onward), which
   reproduces the 69 writable field names from the ledger's
   `writable_fields` header (`:2381`). **It is the source of record for
   the 69 names** that C1 and Task 4 build to.
6. The tree, read-only: `internal/atomicfile/atomicfile.go` and its test;
   `internal/engine/shadow/flac.go` (313 lines, the nucleus this pass
   widens, absorbed at 8c and untouched here);
   `internal/engine/tag/doc.go`; `internal/harness/conform/arch_test.go`;
   `internal/harness/oracle/testdata/gen_template_flac.py` and
   `template.flac`; `scripts/check.sh`; `ruff.toml`; `.gitleaksignore`.
7. **`CLAUDE.md`** in full, and `docs/STATUS.md`'s lane registry.
8. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** Task 2
   writes one Python file, so `python-conventions` applies to that task
   alone. No unit file and no deploy path exists in this pass, so
   `vps-conventions` is out of scope (`spec:1284-1286`).

## Goal

Ship `internal/engine/tagio`'s writer core and its reader: a pure-Go FLAC
tag library that reproduces every serialization rule mediafile 0.17.0
applies to a FLAC file, over a container whose every other block and
every comment the tag set does not own survives the write untouched, with
the write itself atomic, mode-preserving and crash-safe.

Land the block-chain parse with its refusals and its frame-region
offsets, the 69-key tag set with its four input states and its
deterministic write order, the encodings, the exported 69-key read
projection, the writer with its VORBIS_COMMENT merge contract and its
vendor string, exactly one 8 KiB PADDING block per write, the streaming
write in `internal/atomicfile` with its mode parameter and its directory
fsync, and the committed golden container that anchors the emitted bytes
to the pinned oracle's.

Nothing here touches the box, deploys anything, writes `corpus/`, adds a
module dependency, reaches the live network, declares a configuration
key, or edits `internal/engine/shadow`.

## Owner touchpoints: none expected

8a raises no attended question. Fork B is Geoff's and is already owed at
the next sitting; 8a's surface is identical under either answer
(`spec:186-190`), so the pass does not wait on it. Rung 8 declares no
configuration key (ruling 13, `spec:850-866`), so no template gains a
name and no default reaches an adopter through this pass.

**What would become an owner touchpoint, so a task recognizes one rather
than absorbing it.**

- **Any key added to `env/dubplate.env.template`.** That file is a
  secrets manifest rendered by `deploy.sh` from the age store, and adding
  a name there requires minting the secret (CLAUDE.md, "Facts agents get
  wrong"). Rung 8 adds none. A task that finds itself wanting one stops
  and reports.
- **Any adopter-visible behavior**: a reject wording, an accepted-format
  decision, a configuration default, or a change to what the intake gate
  admits. All of those belong to rung 7's string authority and rung 7's
  config, bounded by what this rung supports (`spec:737-745`).
- **A divergence that survives investigation and needs an adjudication.**
  8a produces none, because no ledger grading happens until 8b. If one
  appears anyway, it joins the standing adjudication stream through
  ruling 2's route (`spec:831-838`) rather than opening a new question.

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition below resolves to the conductor. Under the
2026-09-07 unattended-window grant, the 80%-of-ceiling flag routes to the
conductor rather than to Geoff.

## Spec

`docs/superpowers/specs/2026-09-07-rung-8-tag-io.md` (revision 2), and
behind it `specs/2026-09-02-greenfield-server-build.md` (revision 2.1,
ratified), ladder entry 8, quoted in full at `spec:58-70`. 8a builds
deliverables 1, 2 and 6 of the spec's seven (`spec:72-112`), plus the
streaming write ruling 11 assigns to this pass. Deliverables 3, 4 and 5
are 8b's; deliverable 7 is 8c's.

**Hook 14 is recorded rather than built.** The spec's pass-split table
says 8a lands hooks 13 and 14. Hook 13 is Task 4's, and **hook 14 is the
ID3 version ruling, already recorded by owner ruling 4 as ID3v2.4**
(`spec:131-137`, `spec:1009`). Rung 8 writes no ID3, so no task builds
it, and this sentence exists so the build-for-2.0 sweep reads a decline
rather than an omission.

**The forward criterion this pass hands rung 15**, so rung 15 reads it
from the ledger rather than rediscovering it:

> Rung 15 re-runs 8b's conformance instrument over the twelve ledger
> subjects and reads the named deviation note Task 8 produces. An
> unadjudicated divergence blocks the launch gate. The golden container
> at `internal/engine/tagio/testdata` is a fixture rung 15 reads and
> never regenerates.

## Token ceiling: 4.5M (ruled, conductor 2026-09-08)

The spec proposes 4M for 8a on INV's budget evidence (`spec:1141-1164`,
`INV:1061-1135`). The arithmetic is re-derived here rather than carried,
because a ceiling quoted without its derivation is a number nobody can
check. Re-derived, it does not support 4M, and this section proposes 4.5M
with the correction shown.

**The basis, read off `docs/HISTORY.md` through INV's table.** Prep,
meaning plan authoring plus the staleness check, ran 0.25M to 2.55M
across seven passes. Execution through the chain ran 1.3M to 2.2M for a
nine-to-fifteen-task pass. Simplify plus its review is consistently about
0.2M. The close draft plus its review and fold is 0.15M to 0.5M. The
closest comparables are 4a-i (M band, pure Go against an external
contract, 4.99M at prep 2.3M and chain 2.14M) and 4a-ii (2.7M at prep
0.26M, chain 1.32M across sixteen agents for six tasks, fix and review
rounds 0.49M, simplify 0.18M, close 0.16M, conductor 0.2M).

**The per-task rate, corrected.** Revision 1 read 4a-i's chain as 2.14M
over ten tasks and got 220K. `INV:1074` scopes 4a-i to Tasks 1 to 9, and
`docs/HISTORY.md:745-746` lines Task 0 out separately at 0.07M, outside
the 2.14M. The measured chain rate is 2.14M over **nine** tasks, about
**238K**, and 8a carries nine implementer tasks, so 4a-i's chain figure
transfers whole rather than through a rate.

**The arithmetic for 8a.**

| Line | Basis | Figure |
| --- | --- | --- |
| Prep, this plan's authoring workflow | Below 4a-i's 2.3M: no venv probe, no corpus measurement, and a ratified spec plus a 1,338-line inventory doing the close-read work | 0.8M |
| Task 0, staleness, opus at zero context | 4a-ii's staleness plus extraction was 0.26M for a longer checklist | 0.1M |
| Tasks 1 through 9 through the chain | Nine implementer tasks against 4a-i's nine-task chain of 2.14M, about 238K each | 2.14M |
| One `opus` upshift (Task 6) | Roughly one task's worth again | 0.2M |
| Fix rounds and `diff-reviewer` re-dispatches | 4a-ii measured 0.49M; this pass has fewer contested seams | 0.4M |
| `code-simplifier` and its review | Flat across seven passes | 0.2M |
| The close, Task 9 plus the conductor's fold | Smaller than 4a's: no corpus census, no verdict table | 0.25M |
| Conductor overhead | 4a-ii measured 0.2M | 0.2M |
| **Expected total** | | **4.29M** |

**The ceiling is proposed at 4.5M, because 4M schedules its own stop rule
to fire.** Prep's 0.8M and Task 0's 0.1M are spent before Task 1
dispatches. Walk the table forward from there: the chain and its upshift
reach 3.24M by the end of Task 9, and the fix rounds carry it past 3.2M
somewhere around Task 8. An 80% flag at 3.2M therefore fires inside the
chain by the plan's own arithmetic, which is not a signal. `spec:1157-1160`
raised the proposed ceiling from 3M to 4M for exactly that reason, and
the same section grants the move: "Each ceiling is re-derived from the
previous pass's actual report before the next plan is drafted"
(`spec:1162-1164`). At 4.5M the flag lands at 3.6M, past the chain and
inside the close, where a stop is cheap. The 0.21M of headroom over the
expectation is the smallest band that keeps the flag meaningful.

**Ruled, conductor 2026-09-08: the ceiling holds at 4.5M.** What would
have changed had it been ruled at 4M instead: Task 6 would lose its
`opus` upshift and run at `sonnet`, freeing 0.2M and bringing the
expectation to 4.09M. That was the only cut available that did not
remove a deliverable, and it lands on the task the spec calls the
largest in the pass, which is why the recommendation was the ceiling.
Under the ruling, Task 6 keeps its `opus` upshift.

**Recorded against the rung reservation. Ruled, conductor 2026-09-08.**
`spec:1150` reserves 8M for rung 8 (4M, 3M, 1M). At 4.5M the rung total
stands at **8.5M** until 8b closes. The named giveback is 8c's 1M line,
re-derived when 8b closes: 8c is a scoped absorption that moves one
reader, repoints three call sites and deletes `buildFLAC`, and the
spec's own re-derivation clause already schedules that call for when
8b's report lands. This plan does not take it, because 8c's ceiling is
not 8a's to set.

**Flag at 80%, which is 3.6M.** Finish the task in flight, write STATUS,
and send one combined question to the conductor, under the 2026-09-07
unattended-window grant. **The conductor's action when it fires is the
re-derivation this plan already schedules**, since the flag is a forecast
the plan made rather than a surprise. The arithmetic to check at that
point: 3.6M against a 4.29M expectation leaves 0.69M of forecast against
0.9M of headroom, so the close normally fits. If Task 3's re-derivation
already moved the expectation, that number is what the conductor
compares.

**Re-derive the ceiling from Task 3's actual report**, Task 3 being the
first `internal/engine/tagio` behavior task and the pass's one
binary-format parse. If the re-derivation passes 5.5M, the conductor
checks the window cap before continuing and names the split at the Task
3 / Task 4 boundary.

**Checkpoint: STATUS after every four tasks**, so after Task 3 and after
Task 7, plus one at the close. Every task reports its own spend, its
mutation outcomes, and its measured gate wall time, and the conductor
sizes the next task against it.

## Global Constraints

- **Main tree.** The spec's tree column for 8a is "Main tree"
  (`spec:1102`). The lane registry entry is written before the first
  dispatch, and **the live-executor check is that registry, never
  `pgrep -f Projects/dubplate`**, which matches every lane once a
  worktree exists (CLAUDE.md, Lanes).
- **Package territory, stated as a wall.** This pass writes
  **`internal/engine/tagio`** (born this pass, including its `testdata`)
  and **`internal/atomicfile`** (one added exported function and its
  tests, and nothing else). It writes **nothing** under `corpus/`,
  `internal/engine/{shadow,matcher,musicbrainz,tag}`, `internal/harness/`,
  `internal/config`, `cmd/`, `systemd/`, `scripts/`, `env/`, `Makefile`,
  `tools/`, `web/`, or `_salvage/`, and no `docs/` file outside Task 9's
  conductor-owned drafts. A task that believes it needs one of those
  stops and reports.
- **`corpus/` and `internal/harness/` are read-only in this pass**, which
  is the ownership row `RE` cells (8,3) and (8,4) already wrote
  (`research/2026-09-04-reverse-edge-completion.md:161-192`). One task
  reads `internal/harness/oracle/testdata/template.flac`, and it reads it
  **at fixture-generation time only** (Resolution R3). No Go test in this
  pass opens a path outside `internal/engine/tagio`.
- **`tagio` never imports `internal/engine/matcher`** (ruling 3,
  `spec:214-219`). `tagio` exposes its own tag set type and no consumer's
  type appears in its signature. `internal/engine/tag/doc.go:8-11` states
  the split in words and stays true.
- **The engine never imports the harness, and never reads the clock.**
  `TestEngineNeverImportsHarness`
  (`internal/harness/conform/arch_test.go:141`) and
  `TestEngineNeverReadsTheClock` (`:152`) both root at `internal/engine`,
  descend into nested packages, and walk `_test.go` files outside a
  `testdata` subtree. **`internal/engine/tagio` is inside both walks from
  the moment it exists.** A tag write stamps nothing, so the clock rule
  costs this pass nothing.
- **No configuration key, stated in those words** (ruling 13). No task
  edits `internal/config`, and Task 0 item 4 confirms it. A reader who finds no rung 8 reserved class
  reads a decision (`spec:858-866`).
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects. No FLAC library joins it: `mewkiz/flac` is rung 7's for the
  decode-integrity check and exports no block encoder, and every other
  candidate is declined in the spec's prior-art table (`spec:1017-1047`).
  A task that wants a dependency stops and reports.
- **Assertion discipline, at the spec's own bar.** No silent-success
  fakes, no vacuous assertions, no expectations derived from the code
  under test. Every hard-coded expectation in this pass comes from INV's
  measurement, from the RFC, or from the committed golden's bytes.
  **Goldens are last resort** (`go-conventions` SKILL.md), and this pass
  commits exactly one, for the reason `spec:499-508` states: it is the
  rung's only anchor to bytes another implementation produced, and
  `go-conventions` permits a golden at a high-fidelity surface whose
  value is the literal bytes.
- **Universality.** No estate hostname, path, email, catalog value or
  personal name anywhere. Fixture strings are invented. The rung
  implements a published format and a pinned library's serialization,
  both identical for every adopter (`spec:1308-1316`).
- **No live network, and no venv in the gate.** The gate never requires
  uv or beets by design (`spec:844-849`). One task runs one offline
  Python script under `tools/venv/.venv` to author a fixture, once, and
  the gate never runs it (Resolution R1). No task runs `make derive`, and
  no task derives a new corpus subject (ruling 1, `spec:598-606`).
- **`go-conventions` before any Go.** The bar is the standing
  exemplary-stdlib-idiom bar (`idiomatic-beautiful-go`). Concrete types
  over interfaces, and **no interface at all in `tagio` until the MP3
  pass brings the second implementation** (ruling 14, `spec:719-735`).
  `any` not `interface{}`, no `init()`, guard clauses over nesting,
  errors lowercase and context-first with `%w` only where a caller
  branches. **The em dash is banned in Go comments**, and vale scans a
  brand-new file before it is staged (`.vale.ini:5-11`).
- **Atomic writes.** Every file this package writes goes through
  `internal/atomicfile`. No task calls `os.WriteFile` or `os.Create` on a
  target path. Directories 0755, plain files 0644, and a library file's
  existing mode passed by the caller (ruling 11).
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck
  over `scripts/*`, ruff `check` and `format --check` over every tracked
  `*.py` (**this pass adds the ninth**, Task 2's generator), gitleaks over
  the tracked worktree with `--no-git`, vale over Go comment prose,
  `docker compose config`, and a pattern-filtered `systemd-analyze
  verify`. **Never commit red. Specific files only.** Commit footer
  `Co-Authored-By: Claude <noreply@anthropic.com>`.
- **The `.gitleaksignore` rider still binds.** The file is one line,
  `internal/config/keys_test.go:generic-api-key:66`, which pins a finding
  to that test file at its line 66. No task in this pass edits that file.
  **If gitleaks flags the committed golden** (a binary fixture is new
  territory for this gate), the fix is a new fingerprint **appended below
  the existing line**, and the task reports it (Task 2's acceptance
  criteria).
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each one before
  Task 1 dispatches.

## Contested surfaces

The spec's own table (`spec:1122-1139`) is the source. The conductor
writes these into the parallel-approach spec's ownership table before the
lane launches, per lane-launch checklist item 2.

| Surface | Owner | Consumers | 8a's position |
| --- | --- | --- | --- |
| `internal/engine/tagio` (new) | rung 8, created by it; no other lane writes it | rung 7's allowlist bound, rung 9's pipeline, rung 15's re-run | **8a creates it.** 8b adds PICTURE, the seam and the instrument; 8c adds the absorbed reader's doc comment. |
| `internal/engine/tagio/testdata`'s golden container | rung 8, generated once under the pinned venv the way `template.flac` was, then committed and never regenerated in the gate | rung 15 reads it as a fixture | **8a generates it, once** (Task 2), carrying both the VORBIS_COMMENT and the PICTURE block so 8b regenerates nothing. |
| `internal/atomicfile` (amends the existing row) | born by rung 4 pass 4a; **rung 8 adds the streaming write and nothing else** | rungs 6, 9, 13 | **8a adds it** (Task 1). `WriteFile` is left alone (ruling 12). |
| `corpus/` and `internal/harness/` | rung 2 and rung 3 | rung 15 | **Read-only, and read once**, at fixture-generation time only (R3). |
| `internal/engine/shadow` (amends the existing row) | rung 4, created by it; rung 8 removes `flac.go` and repoints three call sites | rung 15's re-run, rung 17's docs | **Untouched by 8a.** That work is 8c's, whole. |
| `docs/` ledgers | conductor-owned | every rung | A pass drafts and never commits. |

**Two passes precede this one in the same tree, and their files are
disjoint from 8a's.** Pass 4b writes `scripts/deploy.sh`, `systemd/`,
`env/`, the `Makefile`, `cmd/dubplate-server`, and comment lines under
`internal/engine/shadow`. Rung 1b writes `internal/config`. Neither
writes `internal/engine/tagio` or `internal/atomicfile`. **The staleness
pre-task names what to re-read** rather than assuming the disjointness
holds: 1b changes `internal/config`'s shape and 8a declares no key, so
that check is a confirmation, and 4b's Task 6 retires the shadow's
provisional marks in comments only, which 8a does not read.

## The package contract, as outcomes

Every task below cites this section rather than restating it. Each item
names the ruling or the spec section it applies. **The plan states
outcomes; the implementer writes the code.**

**C1. The tag set is `tagio`'s own vocabulary.** A named exported type
carrying the 69 writable fields with their Go-native types. It is the
intersection of `Item._media_fields` with `MediaFile.sorted_fields()`,
computed against the pinned venv and matching `corpus/README.md`'s
handoff key for key (`INV:159-172`). No consumer's type appears in its
signature (ruling 3).

**The 69 names have one source of record, and it is
`corpus/README.md`'s handoff section** (`:2354` onward, the names from
`:2381`). `INV:315-385` is the per-row mapping source: it gives each
field its Vorbis keys, its write shape and its mediafile line. **If the
two disagree on a name, the handoff wins and the disagreement is a
finding**, because the handoff is what rung 2 froze and what rung 15
re-reads. Task 0 checks both, and Task 4 builds its hard-coded list from
the handoff.

**C2. Four input states, chosen by the type.** A value, an empty string,
an empty list, and absent. `MediaFile.update` deletes on `None` and sets
on everything else, so a null removes the key and reads back as absent, a
string field's empty value writes a comment with an empty value, and a
list field's empty value writes no comment at all and is indistinguishable
on read-back from absent (`spec:311-327`, `INV:387-403`). **A Go string's
zero value and a Go slice's zero value are the same in Go and are not the
same on disk**, so the type carries the presence bit rather than
inferring it. Row 0 of the ledger carries ten nulls and ten empty
strings, which is why getting this backwards fails twenty keys per row
while every non-empty value matches.

**C3. The write order is part of the contract.** Iteration for a write is
in a fixed order that resolves the three colliding pairs the way
`sorted_fields()` resolves them, with the plural winning (`spec:401-410`,
INV surprise 4). A Go writer that iterates a map is non-deterministic on
`RELEASETYPE`, `MUSICBRAINZ_ALBUMTYPE`, `MUSICBRAINZ_ARTISTID` and
`MUSICBRAINZ_ALBUMARTISTID`.

**C4. A set writes every applicable style; a read takes the first truthy
one.** `MediaField.__set__` writes every applicable style that is not
read-only, and `MediaField.__get__` walks the styles in declaration order
and breaks on the first truthy value (`INV:225-235`). So `albumartist`
writes three Vorbis keys and `disctotal` writes three, while a reader
that emitted only `ALBUMARTIST` would still pass every read-back check
and leave the file under-tagged for Navidrome and every other consumer
(`spec:486-495`).

**C5. Read-back takes the LAST applicable style when nothing is truthy.**
`MediaField.__get__` overwrites its result on each style and breaks only
on a truthy value (`mediafile/fields.py:48-54`), so a field whose
trailing styles are read-only returns absent for an empty value even
though the writer emitted an empty comment. Two of the 69 are shaped that
way: `catalognum` behind read-only `CATALOGID` and `DISCOGS_CATALOG`, and
`barcode` behind read-only `UPC`, `EAN/UPN`, `EAN` and `UPN`
(`spec:327-352`). Both are measured, and the committed ledger already
records the consequence.

**C6. The reader is the exported 69-key projection.** One entry point
reads a file and projects all 69 keys, with first-truthy-style
precedence, the last-style rule of C5, `safe_cast`'s numeric behavior,
the date tuple, and the Q-number division (`spec:698-707`). It is the
package's read entry point, and 8b's conformance instrument calls it
rather than carrying a second projection, because two decoders over one
field bag drift silently.

**C7. The reader parses the metadata blocks and refuses what it does not
fully understand.** The parsed container carries the block chain in
order, each block with its type, its declared length and its payload;
STREAMINFO decoded; VORBIS_COMMENT decoded to its vendor string and its
comments with repeats and order preserved; and the offset and length of
the frame region (`spec:686-692`). **Ruling 9's refusals apply to the
reader as well as the writer** (`spec:257-264`). Six shapes are errors
and never a best-effort read, five matching the spec's own list at
`spec:555-561` and the sixth added by ruling, below:

1. More than one VORBIS_COMMENT block (RFC 9639 section 8.6).
2. A STREAMINFO that is not first or not 34 bytes (section 8.2).
3. A block appearing after a header whose last-block flag is set
   (section 8.1).
4. A declared length running past the end of the file (section 8.1).
5. **A comment field name carrying a code point outside U+0020 through
   U+007E** (section 8.6).
6. **More than one SEEKTABLE block** (section 8.5).

The fifth is folded from the review and **stands, ruled by the conductor
2026-09-08**. Its reason: `tagio` preserves every foreign comment
verbatim (C8), so a name it may not re-emit is a file it cannot
round-trip, and a file it cannot round-trip is malformed rather than
merely odd. `=` needs no rule, because the parser splits a comment on its
first `=` and a parsed name can never contain one.

**The sixth is reversed from the draft reading, ruled by the conductor
2026-09-08.** RFC 9639 section 8.5 allows at most one SEEKTABLE block per
stream (`INV:779`). The draft reading had C10's unconditional copy-through
of unknown blocks cover a second SEEKTABLE as a deliberate non-refusal;
the ruling holds that keeping ruling 9's strictness consistent across the
reader means two SEEKTABLE blocks are refused rather than re-emitted, the
same way the other five shapes are. **This pass's close drafts the
matching correction to ruling 9**, since ruling 9's own text still reads
as the non-refusal the draft took. Task 3 grades this refusal alongside
the other five; Task 8's deviation note no longer carries it, since it is
a refusal rather than a divergence.

**C7a. The parse reads the metadata region only.** The parser
opens the file, reads forward through the block chain, and records the
frame region as an offset and a length rather than as bytes. Every
declared length is bounds-checked against the file's size before a slice
is taken. **A whole-file read here nullifies C13's streaming write**: the
shadow's reader does exactly that (`internal/engine/shadow/flac.go:78`
is `os.ReadFile(path)`), and widening it without changing that line would
make a multi-gigabyte rewrite cost a multi-gigabyte read.

**C8. The writer replaces only the keys the tag set owns.** Every other
comment in the VORBIS_COMMENT block survives untouched and in its
original order (`spec:357-369`). `MediaFile.update` touches only the
fields present in its dict, and `VCommentDict.__setitem__` deletes only
the case-insensitive matches of the key being written before appending. A
writer that rebuilds the block from the 69-key value model drops a real
upload's tagger-specific keys, on admit and on every rung 9 retag.

**Key names match case-insensitively, per RFC 9639 section 8.6**, which
says the evaluation of field names MUST be case insensitive. The
mechanism is `VCommentDict.__setitem__` (`mutagen/_vorbis.py:275-297`),
whose `del self[key]` matches every case before appending. **A writer
that appends without that delete ships a wrong tag through a green
pass**: a container already holding `title=old` in lowercase ends up with
two title comments, and every `tagio` read-back still passes, because the
reader takes the first truthy style. Task 7 grades it.

**C9. The vendor string is part of the block and part of the type.**
`tagio` reads the vendor, preserves it on rewrite, and writes a vendor
naming itself when it creates the block fresh (`spec:451-461`). Measured,
a mediafile write into a scratch template copy produced a block whose
vendor is `Mutagen 1.48.1`, and a rewrite of a file that already has a
block preserves the loaded vendor. A container type that does not carry
the vendor cannot round-trip it.

**C9a. A freshly created VORBIS_COMMENT block is written immediately
after STREAMINFO.** The oracle places it there: measured, a mediafile
save into the 42-byte template produced STREAMINFO, VORBIS_COMMENT, and
one PADDING block (`spec:531-533`). **Stands, ruled by the conductor
2026-09-08.** Task 2's provenance record is the evidence: if the recorded
chain contradicts this, the writer follows the record and Task 9 corrects
this item. Without it, two implementers place a fresh block in two places
and both pass the pass, because R4's byte assertions all write into a
copy of the golden, which already carries a block.

**C10. Every block the writer does not own survives.** STREAMINFO's
payload is copied byte for byte. SEEKTABLE, APPLICATION, CUESHEET and
every reserved type are copied through by declared length, in their
original order, under **re-emitted** headers (ruling 7,
`spec:236-244`). PADDING is the one exception: every incoming PADDING
block is dropped, and one 8 KiB PADDING block is written last, so a file
retagged N times carries the same padding as one retagged once (rulings 5
and 6).

**C11. The last-block flag is the writer's responsibility.** A writer
re-emits every block header rather than copying one through, clears the
flag on all but the final block, and sets it on the block it writes last
(`spec:428-439`). The header's three-byte length is big-endian while the
VORBIS_COMMENT body's lengths are little-endian, which is the asymmetry
that catches implementers.

**C12. The audio is untouched.** The frame region is copied byte for
byte, and STREAMINFO's MD5 field is never recomputed
(`spec:538-541`). The MD5 covers unencoded audio and a tag write moves no
sample (RFC 9639 section 8.2).

**C13. A tag write is atomic or it does not happen, and it preserves the
target's mode.** The write goes through the new streaming path in
`internal/atomicfile`: a temp file in the target's own directory, the new
head, the copied frame region, sync, close, chmod to the mode the caller
passed, rename, then an fsync of the directory (rulings 11 and 12,
`spec:543-553`). **The mode is a parameter because `os.CreateTemp`
creates 0600**, and a rename of that over a 0644 library file makes the
track unreadable to the streaming service while leaving its audio bytes
perfect.

**C14. Refusals are sentinels where a caller branches.** Rung 7's gate
branches on "this container is not supported" and rung 9's repair verbs
branch on "this file is malformed" (`spec:712-715`). Those two are
sentinels; every other error is wrapped context.

**C15. The frame region is the content hash's extent** (ruling 8,
`spec:245-256`). `tagio` exposes the offsets its parse already computes,
so rung 6's schema and rung 9's retag verb agree on one answer. A
whole-file hash goes stale the first time rung 9's retag verb runs.

**C16. The mapping's citations live in the code.** The 69-key table is
reproduced as the mapping's own citation block, one mediafile line
reference per row, in the shape `internal/engine/shadow/flac.go:209-228`
already uses for eighteen rows (`spec:412-415`). **INV's table
(`INV:315-385`) is the row-by-row source**, and this plan reproduces only
the exceptional rows below rather than all 69.

### The exceptional rows, stated here so a task builds to them

Every row not listed below takes its write shape from `INV:315-385`'s
"Write shape" column. **Most are a single writable key taking one
comment, and thirteen are not**: the list-valued row below names them,
because whether a field is list-valued decides its Go type and therefore
which of C2's four states it can hold. A string field's empty value
writes an empty comment; a list field's empty value writes nothing
(`ListStorageStyle.set_list`, `mediafile/storage/base.py:189-199`).
Confusing the two is the twenty-keys-per-row defect C2 exists to
prevent.

| Shape | Fields | Vorbis keys, in mediafile declaration order |
| --- | --- | --- |
| List-valued, one comment per value under the writable key | `albumartists_credit`, `albumartists_sort`, `albumtypes`, `arrangers`, `artists`, `artists_credit`, `artists_sort`, `composers`, `genres`, `lyricists`, `mb_albumartistids`, `mb_artistids`, `remixers` | each field's own key, from `INV:315-385` |
| Three writable styles | `albumartist` | `ALBUM ARTIST`, `ALBUM_ARTIST`, `ALBUMARTIST` |
| Three writable styles | `disctotal` | `DISCTOTAL`, `DISCC`, `TOTALDISCS` |
| Three writable styles | `tracktotal` | `TRACKTOTAL`, `TRACKC`, `TOTALTRACKS` |
| Two writable styles | `track`, `disc`, `comments`, `label`, `albumstatus`, `encoder` | `TRACK`/`TRACKNUMBER`; `DISC`/`DISCNUMBER`; `DESCRIPTION`/`COMMENT`; `LABEL`/`PUBLISHER`; `RELEASESTATUS`/`MUSICBRAINZ_ALBUMSTATUS`; `ENCODEDBY`/`ENCODER` |
| Two writable, one read-only | `albumartists` | `ALBUMARTISTS`, `ALBUM_ARTISTS`, read-only `ALBUM ARTISTS` |
| One writable behind read-only styles (C5's last-style rule) | `catalognum` | `CATALOGNUMBER`, read-only `CATALOGID`, `DISCOGS_CATALOG` |
| One writable behind read-only styles (C5's last-style rule) | `barcode` | `BARCODE`, read-only `UPC`, `EAN/UPN`, `EAN`, `UPN` |
| Colliding pair, plural wins | `albumtype` / `albumtypes` | `RELEASETYPE`, `MUSICBRAINZ_ALBUMTYPE` |
| Colliding pair, plural wins | `mb_artistid` / `mb_artistids` | `MUSICBRAINZ_ARTISTID` |
| Colliding pair, plural wins | `mb_albumartistid` / `mb_albumartistids` | `MUSICBRAINZ_ALBUMARTISTID` |
| Single value into a list-backed key | `albumtype`, `catalognum`, `language`, `mb_artistid`, `mb_albumartistid` | a one-element list through `single_field()` |
| Date, with an unpadded `YEAR` fallback | `year`, `month`, `day` | `DATE`, plus `YEAR` |
| Date, no fallback | `original_year`, `original_month`, `original_day` | `ORIGINALDATE` |
| Float, two places, `" dB"` suffix | `rg_track_gain`, `rg_album_gain` | `REPLAYGAIN_TRACK_GAIN`, `REPLAYGAIN_ALBUM_GAIN` |
| Float, six places, no suffix | `rg_track_peak`, `rg_album_peak` | `REPLAYGAIN_TRACK_PEAK`, `REPLAYGAIN_ALBUM_PEAK` |
| Q8 shift, integer on disk | `r128_track_gain`, `r128_album_gain` | `R128_TRACK_GAIN`, `R128_ALBUM_GAIN` |
| Bool as `"1"` or `"0"` | `comp` | `COMPILATION` |

**The serialization rules, exactly** (`spec:371-399`, `INV:236-296`).

- A float with a string target formats to `float_places` decimals, 2 by
  default and 6 on both ReplayGain peaks.
- A bool becomes `"1"` or `"0"`, never `true` or `false`.
- A suffix appends after formatting, and `" dB"` on the two ReplayGain
  gain fields is the only suffix in the FLAC table.
- The R128 pair shifts by Q8, so `r128_track_gain = -1.5` writes
  `R128_TRACK_GAIN=-384`.
- A multi-valued field repeats its key, and a single-valued field
  reaching a list-backed key writes a one-element list.
- **Dates decompose, and the year is zero-padded to four digits.** The
  value is `YYYY`, then `-MM` when the month is truthy, then `-DD` only
  when the month and the day are both truthy. `DATE` also writes a
  bare-year `YEAR` fallback, unpadded. Only a `None` year deletes the
  field and its fallback together. **A zero year is written rather than
  deleted**, and a fresh beets `Item` defaults year, month and day to the
  integer 0, so the unknown-date path is the common production path once
  rung 9 writes tags. Measured: every component 0 wrote `DATE=0000`,
  `YEAR=0` and `ORIGINALDATE=0000`, and read back year 0 with month and
  day absent. A year of 2001 with a zero month and day produces
  `DATE=2001` and `YEAR=2001`, and reads back as year 2001 with month and
  day absent rather than zeros.
- **The date triple composes as a unit, and a partial triple has a
  stated rule.** C2 lets year, month and day each be absent
  independently, so the writer needs an answer for a present year beside
  an absent month. The rule: an absent year deletes `DATE` and its
  `YEAR` fallback together, and an absent month or day with a present
  year is treated as falsy. **Folded from the review, stands as applied,
  ruled by the conductor 2026-09-08.** It reproduces mediafile's end
  state for every fully-populated case and diverges only where mediafile
  has no equivalent: `DateItemField.__set__` (`mediafile/fields.py:225-228`)
  is a read-modify-write over the file's current `DATE`, so
  `update({"month": 3})` against a file holding `DATE=2001-01-01` yields
  `DATE=2001-03-01`. `tagio`'s write entry point takes a whole tag set
  rather than a partial update, so it has no surface that can express
  that call. Task 8's deviation note records the divergence.
- **On read**, `deserialize` strips the suffix when present, then
  `safe_cast` runs: to `int`, take the leading `[+-]?[0-9]+` of the
  trimmed string and yield 0 when there is none; to `float`, take the
  leading numeric prefix and yield 0.0; to `bool`, `bool(int(val))` with
  a parse failure falling to false. `safe_cast` of an absent value
  returns absent for every type (`INV:253-260`).
- **Invalid UTF-8 in a comment is replaced with U+FFFD per code point on
  read** (ruling 10, `spec:265-273`), matching mutagen's
  `errors='replace'`. Values are valid UTF-8 by construction on write.
- **Field names follow RFC 9639 section 8.6**: U+0020 through U+007E
  excluding U+003D, evaluated case-insensitively. **The only surface that
  can receive an invalid name is the parse**, because C1 fixes the
  writable names as package constants and no caller hands the encoder a
  name, so the bound lives in C7's refusal list and Task 3 grades it. No
  write-side check is added: with C1's constants and C7's refusal, an
  invalid name is unreachable at the encoder, and `go-conventions`
  forbids a guard against an impossible state. `tagio`'s bound is the
  RFC's, one code point wider than mutagen's, and the deviation note
  records it
  (`spec:463-473`).

## The golden container and its byte assertions

**The question this section answers.** How does 8a prove byte-for-byte
agreement with what mediafile writes, without importing the harness and
without a conformance instrument that does not exist until 8b?

**The answer is a committed golden container, and the spec requires it**
(deliverable 6, `spec:100-107`; the guarantee at `spec:499-508`; the
contested-surface row at `spec:1131`). One small FLAC written by the
pinned oracle is committed to `internal/engine/tagio/testdata`, generated
once under the venv the way `template.flac` was, then committed and never
regenerated in the gate. **The alternative is narrower.** Ruling 1's
unit tests derived from the mediafile source grade the serialization
rules the corpus never reaches, and they cannot catch a wrong endianness
or a skipped write style, because a writer with either defect reads its
own files back correctly.
Both mechanisms ship, and the division of labor is stated here.

**The golden's input is written down by the generator that uses it.**
`INV:409-427` is the fenced block holding the *emitted comment set*.
`INV:407` describes the call that produced it as "fifteen-field",
`INV:429` adds `barcode=None` to the same call, and the input dict
itself appears nowhere in INV. Reconstructed from the output it runs to
at least twenty distinct beets fields, and the output still cannot say
whether `albumtype` was set alongside `albumtypes`, because the plural
wins either way. So revision 1's instruction to reproduce "the fifteen
fields `INV:409-427` measured" names a list that does not exist.
**Task 2's generator declares the input, `golden.txt` records it field by
field, and that record is the contract every later assertion reads.**
`INV:409-427` stays cited as the emitted shape the input must at least
cover, since it is the one published measurement of what a mediafile
write puts on disk.

**The input covers every ledger-silent key and every corpus-unreachable
case a mediafile write can express.** Risk 2b (`spec:1230-1237`) is why
the golden exists, and four of Task 8's sixteen ledger-silent keys appear
in `INV:409-427`'s set (`acoustid_fingerprint`, `r128_track_gain`,
`rg_track_gain`, `rg_track_peak`). Widening the input to reach the other
twelve, plus the zero year, the empty list, and `catalognum` and
`barcode` empty, costs no second fixture, which is R2's whole argument.
**Two cases stay out by construction**: mediafile writes values that are
valid UTF-8, and mutagen's `is_valid_key` (`_vorbis.py:25-41`) refuses a
name above 0x7D, so the invalid-UTF-8 sequence and the U+007E field name
remain Task 8's hand-built cases. **A widened input may also change a
value INV's measurement set**, since one file holds one date triple and
the zero-date case needs one of the two triples to carry zeros. The
record names which key it moved and why, so the emitted set stays
readable against `INV:409-427` rather than merely different from it.
Task 2's record names anything else it could not express, with the
reason.

**What the golden proves.**

- The writer's emitted VORBIS_COMMENT block bytes equal the golden's for
  the same input, which pins the little-endian body framing, the comment
  count, the per-comment lengths, and the on-disk comment order.
- The emitted comment set is asserted key by key against the comment
  sequence `golden.txt` records, which pins the write-every-style rule,
  the empty-versus-absent rule, and every serialization rule the widened
  input reaches.
- The reader parses the golden, which pins the parse against bytes
  another implementation produced rather than against the writer's own.

**What the golden does not prove, and why the assertion is scoped to
blocks rather than to the whole file.** The oracle reserves 1 KiB of
padding and `tagio` reserves 8 KiB by ruling 6, and the oracle's vendor
is `Mutagen 1.48.1` while a fresh `tagio` block names `dubplate`. **A
whole-container byte comparison would therefore fail by design.** The
assertion is over the VORBIS_COMMENT block's bytes, taken from a write
into a copy of the golden itself, so the vendor is the preserved one and
the comparison is exact (Resolution R4).

**The on-disk comment order is measured rather than assumed.** The
expected order is field order through `sorted_fields()`, then style
declaration order within a field, then list order within a style. That is
a hypothesis from `INV:297-308` and the mutagen source, and INV's printed
comment set is displayed alphabetically, so it is not evidence of the
on-disk sequence. **Task 2 records the golden's decoded comment sequence
in on-disk order beside the golden**, and Task 7 implements to that
record. If the recorded order contradicts the hypothesis, the writer
follows the golden and Task 9 corrects this plan's sentence in the close.

**The golden carries a PICTURE block as well as a VORBIS_COMMENT block**,
because the contested-surface row says it is generated once and never
regenerated, and 8b's PICTURE byte assertion reads the same file. 8a
parses the PICTURE block's header and preserves its payload by declared
length; **8a decodes no picture and asserts nothing about its fields.**

## Resolutions this plan takes rather than discovers

Every open reading is resolved here, applied as the recommendation. **All
ten stand as applied, ruled by the conductor 2026-09-08.** Each row keeps
its stated fold for the record, though none of the folds is taken.

### R1: the golden's generator is a committed script beside the golden, run once, offline

**Ruled, conductor 2026-09-08: stands.**

**Applied.** `internal/engine/tagio/testdata/gen_golden_flac.py` is a
hand-run script in the exact shape of
`internal/harness/oracle/testdata/gen_template_flac.py`: a module
docstring that is the fixture's provenance record, a copy of the
committed 42-byte template into a scratch path, a `MediaFile` write of
the declared input plus one small picture, and the result written next
to the script. **The script's input dict is the declaration**: the module
docstring names every field and value it writes, and `golden.txt`
reproduces it, so no later task reconstructs the input from the output. **The gate never runs it**, the same way `make derive`
never runs `gen_template_flac.py`. It is a tracked `*.py`, so it joins
the ruff leg and `python-conventions` applies.

**Who runs it.** Task 2's implementer runs it once, under an exception
this plan states in the task: **it is one offline Python invocation
against the pinned venv, it is not `make derive`, it reaches no network,
and it writes only `internal/engine/tagio/testdata`.** The implementer's
standing rules forbid `make derive` and network access, and neither is
engaged here.

**Fold if ruled otherwise.** The conductor generates the golden out of
chain and hands the committed bytes to Task 2 as an input, and Task 2
becomes a review of the fixture plus the provenance record. Nothing else
in the plan moves.

**Prior art.** Adopts the repo's own `gen_template_flac.py` precedent
(`spec:1017-1047` scores the repo's own FLAC work as "adopt and extend").

### R2: one golden, whose declared input reaches every rule a mediafile write can express

**Ruled, conductor 2026-09-08: stands.** The widened input carries.

**Applied.** The golden is written from an input the generator declares
and `golden.txt` records, sized to cover `INV:409-427`'s emitted shape
plus every ledger-silent key and every corpus-unreachable case mediafile
can produce. **One fixture anchors the block-byte assertion, the
key-by-key comment-set assertion, and twelve keys that would otherwise be
graded only by a human's reading of Python source.** Adding a second
fixture would be a second thing to keep in step with the first, and the
input costs nothing to widen.

**Revision 2 changed this resolution.** Revision 1 named the input as
"the exact fifteen-field input `INV:409-427` measured", and no such list
exists: that citation holds the output. The correction is B1 in the plan
review.

**Fold if ruled otherwise.** The golden keeps the narrow input, a second
golden is generated by the same script with the wider one, and Task 7's
comment-set assertion points at the second. The byte assertion is
unaffected.

### R3: `internal/harness/oracle/testdata/template.flac` is read at generation time only

**Ruled, conductor 2026-09-08: stands.**

**Applied.** The generator copies the committed template; the committed
golden is thereafter self-contained. **No Go test in this pass opens a
path outside `internal/engine/tagio`.** The architectural walk forbids an
import rather than a file read, so this is a territory rule rather than a
mechanical one, and it keeps `RE` cell (8,3)'s read-only condition
trivially satisfied.

**Fold if ruled otherwise.** A test may read the template by relative
path, and the reader's "parses a real oracle-written container"
assertion gains a second subject.

### R4: the byte assertion writes into a copy of the golden, not into the template

**Ruled, conductor 2026-09-08: stands.**

**Applied.** Writing the golden's own recorded input into a copy of the
42-byte template would produce a `dubplate` vendor and a different block
length, so the emitted bytes could not equal the golden's. Writing into a copy of
the golden preserves the loaded vendor (C9) and makes the comparison
exact.

**Fold if ruled otherwise.** The assertion excludes the vendor field and
compares the comment region alone, which is weaker: it stops pinning the
vendor length prefix.

### R5: the exported surface names, and what 8b widens

**Ruled, conductor 2026-09-08: stands.**

**Applied.** 8a exports a tag set type, a parsed-file type carrying the
vendor and the frame-region extent, one read entry point over a path, one
write entry point over a path taking the tag set and the file mode, and
two sentinels. **8a's signatures carry no picture parameter.** 8b widens
the parsed-file type with pictures and adds the picture argument to the
write entry point. That is one refactor inside `tagio` with no consumer
wired yet, and it is cheaper than shipping a picture parameter 8a cannot
fill.

**Fold if ruled otherwise.** 8a lands the wider signatures with an empty
picture slice, and 8b fills them. The cost is a parameter that means
nothing for one pass, which the assertion-discipline rule reads as a
silent-success shape.

### R6: no `ctx` on `tagio`'s entry points

**Ruled, conductor 2026-09-08: stands.**

**Applied.** CLAUDE.md's `ctx`-first rule exists for subprocesses and
requests. `tagio` spawns nothing and calls no host, and the shadow's own
reader takes no `ctx` (`internal/engine/shadow/flac.go:77`). A `ctx` a
function never selects on is a parameter that lies about cancellation.

**Fold if ruled otherwise.** Both entry points take `ctx` first, and the
write path checks it before the rename. Rung 9 is the caller that would
want it, and it can be added when a caller can use it.

### R7: both sentinels land in 8a

**Ruled, conductor 2026-09-08: stands.** Both sentinels land in 8a; 8b
adopts the unsupported-container one unchanged. Task 9's close records
the `spec:1103` scope correction this reading crosses, below.

**Applied.** `tagio` defines the unsupported-container sentinel and the
malformed-container sentinel in this pass. A file whose signature is not
`fLaC` returns the unsupported sentinel from the first byte, so 8b's
registry adopts it unchanged rather than reclassifying an error that
already exists. Rung 7 branches on the first and rung 9 on the second
(C14).

**What this crosses, named rather than left for a reader to infer.**
`spec:1103` scopes "the seam registry with its supported set, its
func-value dispatch, and its unsupported-container sentinel" to 8b. **8a
takes the sentinel and leaves the registry, the dispatch and the
supported set to 8b**, so the pass's own scope statement above still
holds: no task here believes it needs the seam registry.

**Fold if ruled otherwise.** 8a defines only the malformed sentinel and
returns a wrapped error for a bad signature; 8b introduces the
unsupported sentinel with the registry and rewrites the signature check.

### R8: the four-state representation is the implementer's, under two stated properties

**Ruled, conductor 2026-09-08: stands.**

**Applied.** Both a pointer-field struct and a value struct paired with a
present-key set satisfy the contract, and `go-conventions` prefers
neither. The plan states the properties and declines to pick the shape.
**Property one: an absent field and an empty field are distinguishable
without a second lookup.** **Property two: iteration for a write yields
the 69 keys in the fixed order of C3.** The `diff-reviewer` grades
against those two, and the four-state table test in Task 4 proves them.

**Fold if ruled otherwise.** The named representation replaces the
implementer's, and Task 4's tests are unchanged, because they assert on
behavior rather than on layout.

### R9: the fresh vendor is asserted by prefix, and the version comes from the build

**Ruled, conductor 2026-09-08: stands.**

**Applied.** The writer emits `dubplate ` followed by the main module's
build version when it creates a block fresh, resolved once from the
build information the toolchain already embeds, falling back to a
development marker when none is available. **The round-trip test asserts
the prefix rather than the exact string**, because a test binary and a
release build report different versions and an exact assertion would
green only one of them.

**Fold if ruled otherwise.** The vendor is a package constant naming the
module without a version, and the assertion becomes exact. The cost is a
file whose vendor string cannot say which dubplate wrote it.

### R10: `WriteFile` is left alone, and the new function is its sibling

**Ruled, conductor 2026-09-08: stands.**

**Applied.** Ruling 11 says an added exported function beside `WriteFile`
rather than a new package, and ruling 12 says the new function fsyncs the
containing directory while `WriteFile` does not change. The shadow's
small JSON records need neither the streaming path nor the directory
fsync, and changing `WriteFile` would change behavior under every rung 4
call site for no stated reason.

**Fold if ruled otherwise.** `WriteFile` gains the directory fsync too,
and rung 4's existing tests are re-run as the evidence that nothing
regressed.

## Test helpers named in this plan

Every test below calls helpers that live in the same package's `_test.go`
files. They are named here once so no acceptance criterion reads as a
reference to something no task defines.

- `internal/atomicfile`: the landed test file's shape is the model. Task
  1 adds `writeErr(t)` (a writer callback that fails at a stated byte
  offset, so the failure path is a real failure rather than an early
  return) and reuses the existing leftover-temp and old-content-intact
  assertions.
- `internal/engine/tagio`: `buildContainer(t, blocks)` (Task 3, an
  in-test container assembler over an ordered block list, so a malformed
  fixture is a one-field edit rather than a hand-typed byte string);
  `golden(t)` (Task 3, reads the committed golden into memory);
  `writeTemp(t, data)` (Task 3, a container on disk under `t.TempDir()`);
  `comments(t, container)` (Task 3, the decoded comment sequence in
  on-disk order, which every comment-set assertion compares against);
  `goldenTagSet(t)` (Task 7, the tag set matching the input `golden.txt`
  records, field for field); `frameHash(t, path)` (Task 7, a digest over the frame region
  alone).

**No helper builds a container the writer also builds.** `buildContainer`
exists to make malformed and exotic fixtures, and every well-formed
round-trip fixture comes from the writer or from the golden. A fixture
builder that duplicates the writer is a second implementation of the
thing under test, which is exactly what ruling 4 retires from the shadow
at 8c.

**The rule has one stated exception, and it is Task 7's exotic
round-trip fixture.** A container carrying one block of every type plus
one reserved type is something neither the writer nor the golden can
produce, so `buildContainer` makes it. **What keeps that fixture honest
is Task 3's golden parse**, which anchors the header layout against bytes
the oracle wrote, so a `buildContainer` that encodes a header wrongly
fails Task 3 before Task 7 ever runs. The exception is named here rather
than left to read as a contradiction.

---

## Task 0: The staleness pre-task

Model: **`opus`**, **fresh context, zero prior context**, dispatched by
the conductor before any implementer. Reads the rung 8 spec and this plan
against the tree as shipped. **Its output is blockers with evidence, not
fixes**; it writes no code and edits no file. The conductor rules each
blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's `spec:`, `INV:`, `S2:` and `CN:` pointers were
resolved mechanically at `47e80cc` during the revision 2 fold, all 96 of
them, and seventeen that opened on a different subject were corrected. A
citation that has moved since is a finding for this task rather than a
surprise for an implementer.

The checklist is the spec's own (`spec:1166-1212`), adapted to 8a. **The
first three are blockers: a "no" stops the launch.** The spec's second
blocker is `corpus/analysis/file_tags.json`; 8a substitutes the corpus
handoff that derives from it, for the reason item 12 gives.

1. **Is the lane registry consistent with where this pass intends to run,
   and is the primary tree free?** `docs/STATUS.md`'s lane registry is
   the live-executor check, never `pgrep -f Projects/dubplate`. *(Expected
   at drafting: lane-1 ACTIVE on pass 4b then rung 1b; lane-2 BLOCKED;
   doc lane IDLE.)* **Report any live executor in this tree, and any
   uncommitted change under `internal/engine/tagio` or
   `internal/atomicfile`, as a blocker.**
2. **Are `corpus/README.md`'s handoff section and its 69-key list
   unchanged since 2d?** *(Expected: the section opens at `:2354`,
   "Handoff to rung 8: what pure-Go FLAC tag writing may assume", and the
   69 names reproduced from the ledger's `writable_fields` header run
   from `:2381`.)* **This is 8a's source of record for the 69 names**
   (C1), and Task 4 hard-codes its list from it. **Report the count, and
   any name that differs from `INV:315-385`'s table**, because a
   disagreement between those two is the defect this check exists to
   catch, and the handoff wins.
3. **Is `internal/atomicfile` still one exported function taking
   `[]byte`?** *(Expected: `atomicfile.go`, `WriteFile(path string, data
   []byte, perm os.FileMode) error` at `:15`, 42 lines of
   implementation.)* Ruling 11's shape depends on it. Report the test
   file's case list, since Task 1 extends it.
4. **Has rung 1b landed, and did it change `internal/config`'s shape?**
   *(Expected: possibly yes; the plan is
   `plans/2026-09-07-rung-1b-config-registry.md`.)* **8a registers no key
   either way**, so this is a confirmation rather than a blocker. Report
   the doctor table's shape, because that is what a reviewer checks a
   no-key claim against, and confirm `internal/config/keys.go` still
   carries no rung 8 reserved class.
5. **Has pass 4b's Task 6 retired the shadow's provisional marks?**
   *(Expected: comment-only edits under `internal/engine/shadow`.)* 8a
   touches no shadow file, so report the state and move on. A change to
   `internal/engine/shadow/flac.go`'s **behavior** rather than its
   comments is a finding, because 8c's absorption is written against the
   landed reader.
6. **Is `internal/harness/oracle/testdata/template.flac` still the same
   42 bytes**, and does `gen_template_flac.py` still describe it the way
   R1's generator copies? *(Expected: 42 bytes; the script's docstring at
   `:1-27`.)*
7. **Is `corpus/manifest.json` still pinning `mediafile==0.17.0`**, and
   is `tools/venv/.venv` present with beets 2.13.1 and mediafile 0.17.0
   importable? Every serialization rule in this plan is read from that
   pin, and Task 2 runs one script against it. **Report the two versions
   by name, and do not run any derivation.**
8. **Do `TestEngineNeverImportsHarness` and `TestEngineNeverReadsTheClock`
   still walk `internal/engine` including test files outside
   `testdata`?** *(Expected: `arch_test.go:141` and `:152`.)*
   `internal/engine/tagio` enters both walks the moment it exists.
9. **Does `internal/engine/tag/doc.go` still describe `tagio` as knowing
   nothing about matching?** *(Expected: `:8-11`.)* Ruling 3 preserves
   that sentence.
10. **Has `scripts/check.sh` gained or lost a leg**, and does the ruff leg
    still glob `git ls-files '*.py'`? *(Expected: eight tracked `*.py`
    today, nine after Task 2. CLAUDE.md says six, which the spec's errata
    already record as a one-line correction owed at some close.)* Pass 4b
    owns the gate during this window, so report the leg list as it stands.
11. **Is `.gitleaksignore` still a single line pinning
    `internal/config/keys_test.go` by line number?** **Report the number
    and the row's identity without quoting the fixture literal**, because
    gitleaks scans the tracked worktree with `--no-git` and would flag
    the quotation.
12. **Are `corpus/analysis/file_tags.json`'s headers unchanged?**
    *(Expected: `shim_version` 1, `beets_version` 2.13.1, 69
    `writable_fields`, 9 `readonly_fields`, twelve rows.)* This is not a
    blocker here, and the one sentence why: **8a reads the handoff's name
    list rather than the ledger's rows**, so a changed header is 8b's
    blocker, not 8a's. **Cross-check that the header's
    `writable_fields` count equals item 2's handoff count**, which is
    where the two would drift apart unnoticed. **Count mechanically; read
    no row body into context.**
13. **Is `go test ./...`'s wall time still what this plan sizes gate runs
    against?** Run it once at `-count=1` and report the total and the
    `internal/harness/conform` figure. *(4a measured `go test ./...` at
    about 2m06s and `bash scripts/check.sh` at about 3.5 minutes.)*

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
beyond item 13's timing measurement; the task writes nothing.

**2.0 hooks: neutral.** The pre-task builds no surface.

---

## Task 1: `internal/atomicfile`'s streaming write

Model: `sonnet`. Depends on Task 0. **Independent** of Tasks 2 through 9
by files: it writes only `internal/atomicfile`.

Every tag write in this package goes through this function, so it is born
first and Task 7 has no reason to invent a second temp-file convention.

**Files:**
- Modify: `internal/atomicfile/atomicfile.go`
- Modify: `internal/atomicfile/atomicfile_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: one added exported function beside `WriteFile`, taking the
  destination path, the file mode, and a caller-supplied function that
  writes the new content to an `io.Writer`. It creates a temp file in the
  target's own directory, invokes the caller's function, syncs, closes,
  chmods to the passed mode, renames over the target, then opens the
  containing directory and fsyncs it. `WriteFile` is unchanged.

**Outcomes:**

- The content never buffers in memory. The caller streams into the
  writer, so a multi-gigabyte FLAC rewrite costs one file's worth of
  disk and a fixed amount of memory. That is the defect `go-flac`'s v1
  `Save` carries, and the estate's library is where multi-gigabyte files
  live (`INV:906-916`).
- The mode is a parameter (ruling 11). `os.CreateTemp` creates 0600,
  and a rename of that over a 0644 library file makes the track
  unreadable to the streaming service while leaving its audio bytes
  perfect, so no integrity check would notice.
- The containing directory is fsynced after the rename (ruling 12). A
  library file that survives a power cut while its directory entry does
  not is the worse failure.
- A failed write leaves the old file intact and no temp file behind.
  An error from the caller's write function propagates wrapped, and the
  temp file is removed.
- `WriteFile` is untouched (R10), so every rung 4 call site keeps its
  behavior.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The new function is exported, sits beside `WriteFile` in the same file,
  and adds no new package.
- Its doc comment states the streaming property, the mode parameter's
  reason, and the directory fsync's reason, and states nothing the
  signature already says.
- A table test covers creation, overwrite of an existing file, the passed
  mode surviving the rename, empty content, and a write function
  returning an error.
- **The renamed file's mode is asserted against the passed mode, and the
  passed mode is not 0600 in any row**, so a test cannot pass by
  inheriting `os.CreateTemp`'s default.
- The failure row asserts three things: the error propagates and
  names the caller's error through `errors.Is`, the target still holds
  its old content, and the target's directory holds no leftover temp
  file. The failing writer fails partway through rather than before its
  first byte, so the assertion covers a partial write rather than a
  no-op.
- The temp file is created in the target's own directory, asserted by the
  same shape the landed test uses.
- `go.mod` unchanged. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Drop the chmod → the mode row fails, reporting 0600 where the passed
   mode is wanted.
2. Drop the directory fsync → **no test fails**, and the task reports it
   as a deliberately unfired mutation. A power-cut property is not
   observable from a test on this workstation, and the doc comment plus
   ruling 12 is what holds it. Report the mutation as unfired rather than
   inventing an assertion that cannot see it.
3. Remove the temp file before the rename instead of after a failure →
   the overwrite row fails, because the rename target is gone. **The
   obvious wording of this mutation, "remove on success instead of on
   failure", fires nothing**: `WriteFile` already removes on success,
   `defer func() { _ = os.Remove(tmpPath) }()` at
   `internal/atomicfile/atomicfile.go:22` running after the rename at
   `:38`, where it is a harmless no-op.
4. Return early on the caller's error without removing the temp file →
   the leftover-temp assertion in the failure row fails.
5. Buffer the caller's writes and call `WriteFile` internally → the tests
   still pass, and the task reports it unfired with the reason: the
   streaming property is a memory property, and the acceptance criterion
   that holds it is the reviewer reading the diff for an
   `io.ReadAll`-shaped buffer. **This is the one place a reviewer's read
   is the proof**, and the task says so in its report.

**2.0 hook: neutral, and consumed later.** Rungs 6, 9 and 13 consume this
function (`spec:1132`). It lands no hook itself.

**Prior art:** adopts the repo's own `internal/atomicfile`, extended
rather than replaced, which is what `S2`'s prior-art table already
commits to (`INV:917-923`). Declines `go-flac/go-flac` v2 as a
dependency and reads its `saveInPlace` as a design reference for what not
to do (`spec:1028`, `INV:938-950`).

**Deliverable count: 2** (the function, its tests). Gate green; commit.

---

## Task 2: The golden container and its provenance record

Model: `sonnet`. Depends on Task 0. **Independent** of Task 1 by files.

This is the pass's only anchor to bytes another implementation produced,
and the reason it exists is risk 2b (`spec:1230-1237`).

**Files:**
- Create: `internal/engine/tagio/testdata/gen_golden_flac.py`
- Create: `internal/engine/tagio/testdata/golden.flac`
- Create: `internal/engine/tagio/testdata/golden.txt` (the provenance
  record: the exact input, the pinned versions, the decoded block chain
  with each block's type and declared length, and **the decoded comment
  sequence in on-disk order**)

**The named exception this task runs under.** The task runs one offline
Python invocation against `tools/venv/.venv`. **It is not `make derive`,
it reaches no network, it reads no corpus record, and it writes only
`internal/engine/tagio/testdata`.** Every other standing rule holds. If
the implementer cannot run it, that is a report to the conductor and
Resolution R1's fold applies.

**Outcomes:**

- The script copies the committed 42-byte template to a scratch path,
  writes a declared input through `MediaFile.update`, adds one small
  picture, and saves the result beside itself as `golden.flac`.
- **The input is declared by the script and recorded by `golden.txt`,
  field by field.** `INV:409-427` holds the emitted comment set rather
  than the input, so nothing reconstructs the input from the output. The
  input covers `INV:409-427`'s emitted shape and widens it to reach
  **every one of Task 8's sixteen ledger-silent keys** and **every
  corpus-unreachable case a mediafile write can express**: the zero year,
  the empty list, and `catalognum` and `barcode` empty. Two cases stay
  out by construction, and the record says so: mediafile writes valid
  UTF-8, and mutagen's `is_valid_key` (`_vorbis.py:25-41`) refuses a name
  above 0x7D, so the invalid-UTF-8 sequence and the U+007E field name are
  Task 8's hand-built cases.
- The record names anything else the input could not express, with
  the reason, so a later reader sees a bounded fixture rather than an
  arbitrary one.
- The script is provenance, not a build step. Its module docstring
  says so, in the shape `gen_template_flac.py:1-27` uses, and names the
  command that produced the committed bytes.
- The provenance record states the pinned beets and mediafile versions,
  the input field by field, the golden's byte size, the block chain in
  order with each type and declared length, and the comment sequence in
  on-disk order. **That sequence is what Task 7 implements to.**
- The picture is a small invented image with a sniffable magic prefix.
  8a asserts nothing about it beyond its block surviving a parse; 8b
  grades it.

**Acceptance criteria.** Two of these are over a committed binary, which
a reviewer reading a diff cannot decode. **The reviewer grades the
provenance record's internal consistency and the script's declared input
against it; Task 3's parse assertion is what proves the record against
the bytes.** A Task 2 accept is therefore provisional until Task 3 lands,
and the conductor holds it that way.

- The script passes `ruff check` and `ruff format --check` as a tracked
  `*.py`, and follows `python-conventions`: a PEP 257 docstring stating
  intent, no comment that paraphrases the next line.
- **The provenance record's input matches the script's declared input
  field for field**, which a reviewer verifies from the diff alone, since
  both are in it.
- **The record's field list covers all sixteen ledger-silent keys and the
  four expressible corpus-unreachable cases**, or names each omission
  with its reason. Verifiable from the diff against Task 8's list.
- The committed golden's block chain matches the provenance record
  exactly, including the block count, each declared length, and the
  padding block's size. *(Closed by Task 3's first criterion, not by this
  reviewer.)*
- The golden's vendor string is recorded verbatim in the provenance
  record. *(Expected `Mutagen 1.48.1`, `spec:453-456`; the record states
  what was actually produced. Closed by Task 3.)*
- **The recorded block chain says where the oracle placed its
  VORBIS_COMMENT block relative to STREAMINFO**, which is C9a's evidence.
  A chain that contradicts C9a is a finding for the conductor and a
  Task 9 correction.
- The golden's size is reported, and it is well under 10 KiB.
- The comment sequence in the record is the decoded on-disk order,
  and the record says whether it matches this plan's hypothesis of
  `sorted_fields` order, style declaration order, then list order. A
  mismatch is a finding for the conductor and a correction Task 9
  carries, never a silent adjustment.
- The gate is green with a binary fixture committed, which is new for
  this repo's gitleaks leg. If gitleaks flags the golden, the fingerprint
  is **appended below** `.gitleaksignore`'s existing line and the task
  reports the addition.
- No Go file is added by this task.

**Named mutations this task reports:**

1. Regenerate the golden with one field changed and re-run the recorded
   comment sequence check → the record and the file disagree, which is
   the check that the record is derived from the committed bytes rather
   than typed by hand.
2. Truncate the golden by one byte → **no test fails yet**, and the task
   reports it unfired with the reason: Task 3 is the first task that
   parses it, and Task 3's own mutation list covers the same edit. This
   is the honest statement that a fixture-only task cannot prove itself.

**2.0 hook: neutral.** The fixture lands no hook. It carries the PICTURE
block 8b's hook 3 work grades.

**Prior art:** adopts `gen_template_flac.py`'s hand-run generator
precedent. Declines a second conformance corpus, which
`spec:844-849`'s narrowed decline already settles.

**Deliverable count: 3** (the generator, the golden, the provenance
record). Gate green; commit.

---

## Task 3: The container reader, its refusals, and the frame region

Model: `sonnet`. Depends on Tasks 0 and 2.

This is the pass's binary-format parse. It is the shadow reader's
mechanism widened from two block types to the whole chain, with ruling
9's refusals added and the frame region exposed.

**Files:**
- Create: `internal/engine/tagio/doc.go`
- Create: `internal/engine/tagio/container.go` + `container_test.go`

**Interfaces:**
- Consumes: the golden and its provenance record (Task 2).
- Produces: a parsed-container type carrying the block chain in order
  (each block's type, declared length and payload), the decoded
  STREAMINFO, the decoded VORBIS_COMMENT (its vendor string and its
  comments with repeats and order preserved), and the frame region's
  offset and length. Plus the two sentinels of R7. Tasks 6 and 7 consume
  all of it.

**Outcomes:**

- The chain is parsed, not sampled. Every block is carried by type,
  declared length and payload, including reserved types the parser has
  never seen (C10). The shadow's reader skips what it does not read;
  this one keeps it, because the writer has to re-emit it.
- VORBIS_COMMENT decodes to a sequence, not a map. Repeats and order
  are preserved, which the shadow's `map[string]string` cannot hold
  (`spec:688-690`). The vendor string is a field of the decoded block
  (C9).
- The refusals of C7 are errors, and the file is never partly read.
  All six of C7's shapes plus the signature check: more than one
  VORBIS_COMMENT block; a STREAMINFO that is not first or not 34 bytes; a
  block after a header whose last-block flag is set; a declared length
  running past the end; **a comment field name carrying a code point
  outside U+0020 through U+007E**; **more than one SEEKTABLE block**; and
  a signature that is not `fLaC`. The first six return the malformed
  sentinel; the signature returns the unsupported sentinel (R7). The
  field-name row is the spec's fifth refusal (`spec:555-561`), which
  revision 1 left with no home. The SEEKTABLE row is ruled, conductor
  2026-09-08, reversing the draft reading's non-refusal.
- A comment named with U+007E parses and is preserved. That is the
  deviation from mutagen, whose `is_valid_key` stops at 0x7D
  (`_vorbis.py:25-41`) and whose loader skips such a comment silently
  (`:127-128`). `tagio`'s bound is the RFC's. Task 8 records it.
- The frame region's offset and length are exposed (C15). The parse
  already computes them.
- The parse reads the metadata region, never the whole file (C7a).
  The parser reads forward through the block chain and records the frame
  region as an offset and a length. **Widening the shadow's reader
  without changing its first line would defeat Task 1's whole purpose**:
  `internal/engine/shadow/flac.go:78` is `os.ReadFile(path)`, and a
  parse that buffers a multi-gigabyte file makes the streaming write
  cosmetic.
- STREAMINFO is decoded by offset, using the layout the 4a plan
  states at `plans/2026-09-05-rung-4a-shadow-runner.md:2404-2416`: the
  four size fields, then the eight-byte pack at payload offsets 10
  through 18, then the 16-byte MD5. Revision 1 of that plan named the
  wrong eight bytes, and the corrected table is the one to build to.
- A length is never trusted into an allocation. Every declared length
  is bounds-checked against the file's remaining extent before any slice
  is taken. `dhowden/tag`'s open issue #116 is exactly this bug, and it is
  why that library is declined (`spec:1030`).
- `doc.go` states the package's one reason to exist: it reads and
  writes FLAC tags, it knows nothing about matching, it imports no
  harness package and reads no clock, and every path and mode reaches it
  as a parameter.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The committed golden parses, and its block chain, each declared
  length, its vendor string and its comment sequence equal the provenance
  record's, field by field. This is the assertion a shared misreading of
  the format cannot satisfy, because the record came from the oracle.
- Every block type in the golden is represented in the parsed chain,
  including the PICTURE block, carried by declared length with its
  payload intact.
- A table test covers each refusal, one row per shape, each fixture
  mutating one byte, one length, one comment name, or one duplicated
  SEEKTABLE block from a well-formed container built by `buildContainer`.
  **Seven rows**, matching the outcome list above.
- The parse never reads the whole file, asserted by the reviewer's
  read of the diff: no `os.ReadFile`, no `io.ReadAll` over the target,
  and the frame region reaching the parsed container as two numbers
  rather than as bytes. **This is one of the pass's two reviewer-read
  proofs**, and the task names it in its report.
- A comment named with U+007E parses and its value is preserved,
  which is the deviation row Task 8's note carries.
- Each refusal returns the stated sentinel through `errors.Is`, and the
  parse returns a zero container rather than a partly filled one.
- The frame region's offset and length are asserted against a container
  whose frame bytes are a known invented run, and the slice named by
  those two numbers equals that run exactly.
- A reserved block type the parser has never seen survives the parse with
  its payload intact.
- A repeated comment key is carried twice, in order, rather than merged.
  **The shadow's first-seen-wins merge is a projection rule, not a parse
  rule**, and it stays on the shadow side until 8c.
- `TestEngineNeverImportsHarness` and `TestEngineNeverReadsTheClock` both
  pass with the new package inside their walks. The package's tests call
  `time.Now` nowhere.
- `go.mod` unchanged. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Read the metadata block header's three-byte length as little-endian →
   the golden parse fails, because the second block's length lands past
   the end of the file. This is the endianness asymmetry C11 names.
2. Read the VORBIS_COMMENT vendor length as big-endian → the golden's
   vendor string assertion fails, reporting garbage.
3. Read the STREAMINFO pack from payload offsets 26 through 34 instead of
   10 through 18 → the golden's sample rate reads out of MD5 bytes and
   the assertion fails. Every test that reads only `buildContainer`'s
   output would still pass, which is why the golden assertion exists.
4. Accept a second VORBIS_COMMENT block → that refusal row fails.
5. Continue past a header whose last-block flag is set → that refusal row
   fails.
6. Take the payload slice before bounds-checking the declared length →
   the past-the-end refusal row fails, and the failure names the
   allocation the check prevents.
7. Merge a repeated comment key first-seen-wins → the repeated-key
   assertion fails, reporting one comment where two are wanted.
8. Accept a comment name carrying a code point above U+007E → the
   field-name refusal row fails.
9. Reject a comment name containing U+007E → the U+007E preservation row
   fails, which is the half of the bound that proves it is the RFC's
   rather than mutagen's.
10. Accept a container carrying two SEEKTABLE blocks → the SEEKTABLE
    refusal row fails, reporting both blocks re-emitted.
11. Read the whole file into memory and parse from the buffer → **no test
    fails**, and the task reports it unfired: this is C7a's property, and
    the acceptance criterion that holds it is the reviewer's read.

**2.0 hook: lands hook 1's mechanism.** Ruling 8 names the frame region
as the content hash's extent, and this task exposes the offsets rung 6's
schema and rung 9's retag verb both read (`S2:359-360`,
`spec:245-256`).

**Prior art:** adopts the repo's own `internal/engine/shadow/flac.go` as
the nucleus, widened (`spec:1026`). Declines `mewkiz/flac` for tags,
because no block encoder is exported and retagging through `NewEncoder`
means re-encoding every frame. Declines `dhowden/tag` on the unbounded
allocation, which this task's bounds rule answers directly. Declines
`go-flac/flacvorbis` with its base.

**Deliverable count: 3** (the package doc, the container parse with its
frame region, the refusal set). Gate green; commit. **Report actual token
spend, the eleven mutations' outcomes, and the package's measured test
wall time. This task's report is where the conductor re-derives the pass
ceiling.**

---

## Task 4: The tag set, the 69-key mapping, and the write order

Model: `sonnet`. Depends on Task 0. **Independent** of Tasks 1, 2 and 3
by files.

**Files:**
- Create: `internal/engine/tagio/tagset.go` + `tagset_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: the exported tag set type of C1, its four-state model (C2),
  its fixed write order (C3), and the mapping table with its per-row
  citations (C16). Tasks 5, 6 and 7 consume it; 8b's instrument and rung
  9's pipeline consume it after that.

**Outcomes:**

- The 69 fields, with their Go-native types, matching
  `corpus/analysis/file_tags.json`'s `writable_fields` header key for
  key. The nine read-only audio properties are never a writer's target
  and are not on the type.
- The four states are explicit (C2), under R8's two properties.
- The mapping table is in the code with its citations (C16): each
  field's Vorbis keys in mediafile declaration order, each style marked
  writable or read-only, and the mediafile line each row was read from.
  The exceptional rows are in the package contract above; the rest come
  from `INV:315-385`.
- Iteration for a write is in one fixed order (C3), and the three
  colliding pairs resolve with the plural last. **The order is the
  contract rather than an implementation detail**, and it is stated in
  the type's doc comment with its reason.
- Sort tags are in the set. `albumartist_sort`, `artist_sort`,
  `albumartists_sort`, `artists_sort` and `composer_sort` are five of the
  69, with their keys from INV's table.
- The thirteen list-valued fields carry a list type, not a string
  type. They are named in the package contract's exceptional-rows table,
  and the distinction decides which of C2's four states each field can
  hold.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The field count is 69, asserted mechanically rather than by reading
  the struct. A test enumerates the set and compares the count and the
  names against a hard-coded list taken from **`corpus/README.md`'s
  handoff section** (`:2354` onward, names from `:2381`), which C1 names
  as the single source of record. `INV:315-385` supplies each row's
  Vorbis keys and its mediafile citation, not the names. **If the two
  disagree on a name, the handoff wins and the task reports the
  disagreement**; Task 0 item 2 checks for it before this task
  dispatches.
- Every field's mapping row carries at least one Vorbis key and a
  mediafile line citation, asserted by a test that walks the table.
- The read-only styles are marked, and the marks match the
  exceptional-rows table for `albumartists`, `catalognum` and `barcode`.
- The write order is deterministic across runs, asserted by iterating
  twice and comparing, **and** by asserting the position of each member
  of the three colliding pairs relative to its partner.
- The four states are distinguishable for a string field, a list field
  and a numeric field, asserted per state with a table.
- The type's doc comment states why the presence bit exists, citing the
  delete-on-null rule, and says nothing the field names already say.
- No import of `internal/engine/matcher`, asserted by the absence in the
  file and covered by the reviewer's read (ruling 3).
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Drop one field from the set → the count-and-names test fails naming
   that field.
2. Reverse the three colliding pairs' relative order → the pair-position
   assertions fail, one per pair.
3. Collapse the empty-string state into the absent state → the four-state
   table fails on the string row.
4. Collapse the empty-list state into the absent state → the four-state
   table fails on the list row. **These two mutations are separate on
   purpose**: revision 1 of the spec folded the two empty forms together,
   and the ledger's 45 empty-list cells against 149 empty strings is why
   they are distinct.
5. Iterate the set from a Go map → the determinism assertion fails
   intermittently, and the task reports how many runs it took to fire.
   **If it does not fire within the runs attempted, the task says so**,
   and the position assertions of mutation 2 remain the real proof.
6. Drop a read-only mark from `barcode`'s `UPC` → no test in this task
   fails, and the task reports it unfired: the mark's consequence is a
   write, which Task 7 asserts. Naming it here keeps the gap visible.

**2.0 hook: lands hook 13.** The five sort-tag fields are in the writable
69, and retrofitting means re-tagging the library (`S2:388-389`,
`spec:1008`). **Enables hook 11**: six of the 69 are the ReplayGain and
R128 values rung 6's schema holds and rung 9 computes (`S2:383-384`).

**Prior art:** the mapping is read from pinned mediafile 0.17.0 through
INV's table. **Picard's tag-mapping tables are adopted for hook 13's
per-container names in the follow-on passes** and are not consulted for
this row set, because the FLAC names come from the pinned library the
corpus was derived with rather than from a general mapping
(`spec:1035`).

**Deliverable count: 3** (the type with its states, the mapping table
with its citations, the write order). Gate green; commit.

---

## Task 5: The encodings

Model: `sonnet`. Depends on Task 4.

Every serialization rule mediafile applies to a FLAC file, reproduced
exactly. The rules are fully specified by the package contract's
serialization list and `INV:236-296`, so this task implements a stated
table rather than deriving one.

**Files:**
- Create: `internal/engine/tagio/encode.go` + `encode_test.go`

**Interfaces:**
- Consumes: the tag set (Task 4).
- Produces: the field-to-comment serialization for all 69 keys, and the
  comment-to-field deserialization Task 6's projection calls.

**Outcomes:**

- Every rule in the package contract's serialization list: float
  places, the bool encoding, the `" dB"` suffix, the Q8 shift, list
  repetition, the one-element list for a single value into a list-backed
  key, and the date decomposition with its four-digit year and its
  unpadded `YEAR` fallback.
- The four states serialize to three behaviors: a value writes a
  comment, an empty string writes a comment with an empty value, an empty
  list writes nothing, and absent removes the key (C2).
- Deleting a field deletes its read-only styles too, so removing
  `barcode` removes `UPC`, `EAN/UPN`, `EAN` and `UPN`. `read_only` blocks
  a set and not a delete (`INV:238-240`).
- `safe_cast`'s behavior on read, including the leading-numeric-prefix
  rule and the zero fallbacks, and the suffix strip before the cast.
- Invalid UTF-8 is replaced with U+FFFD per code point on read
  (ruling 10). Values are valid UTF-8 by construction on write.
- No field-name validation lives here. The bound is the parse's, in
  Task 3's refusal table, because C1 fixes the writable names as package
  constants and this task's surface takes no field name from a caller.
  Revision 1 put the criterion here with no input path; the correction is
  B2 in the plan review.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- A table test per rule, each row's expected string **hard-coded** and
  each row citing the mediafile line it reproduces. No expectation is
  computed from the code under test.
- `r128_track_gain = -1.5` writes `R128_TRACK_GAIN=-384`, taken from
  INV's measurement.
- `rg_track_gain` writes two decimal places with `" dB"` appended, and
  `rg_track_peak` writes six decimals with no suffix.
- `comp` writes `"1"` and `"0"`, never `true` or `false`.
- The date table covers: all components present; a truthy month with a
  zero day; a zero month with a truthy day; every component zero; an
  absent year; and **a present year with an absent month**, which asserts
  the compose-as-a-unit rule of the serialization list, treating the
  absent month as falsy and writing `DATE=YYYY` with its `YEAR`
  fallback. **The every-component-zero row asserts `DATE=0000`,
  `YEAR=0` and `ORIGINALDATE=0000`**, which is INV's measurement and the
  common production path once rung 9 writes tags.
- The absent-year row asserts that `DATE` and its `YEAR` fallback are
  both removed.
- A two-element list writes its key twice, in list order, under every
  writable style of that field.
- An empty list writes nothing, asserted against an empty comment
  sequence rather than against an empty string.
- Deleting `barcode` removes all five keys.
- A comment carrying an invalid UTF-8 sequence reads back with U+FFFD per
  offending code point, asserted against a hard-coded expected string.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Format floats with `%g` → the ReplayGain rows fail on both the gain
   and the peak.
2. Drop the `" dB"` suffix → the two gain rows fail and the two peak rows
   still pass, which is the asymmetry the suffix rule predicts.
3. Shift the R128 pair by 2^7 instead of 2^8 → the R128 row fails,
   reporting -192 where -384 is wanted.
4. Pad the year with `%d` instead of `%04d` → the every-component-zero
   row fails on `DATE`, and the `YEAR` fallback still passes, because the
   fallback is unpadded by design.
5. Write `-DD` when the day is truthy and the month is zero → the
   zero-month row fails.
6. Delete on an empty string instead of writing an empty comment → the
   empty-string row fails, and the empty-list row still passes, which
   proves the two states are handled separately.
7. Write an empty comment for an empty list → the empty-list row fails.
8. Ignore invalid UTF-8 instead of replacing it → the replacement row
   fails, reporting the raw bytes.
9. Omit `DATE` entirely when the year is present and the month is absent
   → the present-year-with-absent-month row fails, reporting no `DATE`
   where `DATE=YYYY` and `YEAR=YYYY` are wanted. That row is what pins
   the compose-as-a-unit rule, which treats an absent month as falsy
   rather than as a reason to skip the field.

**2.0 hook: enables hook 11.** The ReplayGain and R128 serializations are
what let rung 9 compute those values and rung 6's schema carry them
(`S2:383-384`).

**Prior art:** read from pinned mediafile 0.17.0 (`storage/base.py`,
`fields.py`) through INV section 2.1. No library is adopted or declined
here, because no Go library implements this mapping (`spec:1042-1047`).

**Deliverable count: 2** (the serialization for all 69 keys, the
deserialization Task 6's projection calls). Gate green; commit.

---

## Task 6: The exported 69-key read projection

Model: **`opus`** (see Execution model). Depends on Tasks 3, 4 and 5.

The spec calls this the largest single item in the pass
(`spec:703-705`). It is a full reimplementation of `MediaField.__get__`
for the 69 fields, and it is 8a's rather than 8b's because 8a's
read-back agreement guarantee is otherwise unprovable until 8b.

**Files:**
- Create: `internal/engine/tagio/read.go` + `read_test.go`

**Interfaces:**
- Consumes: the parsed container (Task 3), the tag set (Task 4), the
  deserialization (Task 5).
- Produces: the package's exported read entry point over a path, and the
  projection over a parsed container that the entry point calls. 8b's
  conformance instrument calls the exported one rather than carrying a
  second projection.

**Outcomes:**

- All 69 keys, projected (C6), with first-truthy-style precedence,
  **the last-style rule of C5** for a non-truthy read, `safe_cast`, the
  date recomposition, the Q-number division, and the list shape.
- The truthiness test runs on the raw fetched string before the cast,
  so `TRACK=0` beside `TRACKNUMBER=5` yields 0 rather than 5. A `"0"` is
  a truthy Python string, which is the trap the shadow's reader already
  documents.
- **`catalognum` and `barcode` read back absent for an empty written
  value** (C5), because their trailing styles are read-only. Both carry
  their own test under ruling 1, and the committed ledger already records
  the consequence on 5 and 6 of 12 subjects.
- A comment's first value is the one a single-valued field reads, and
  a list-valued field reads the whole value list for its first non-empty
  style.
- An absent key reads back absent rather than a typed zero, which is
  `safe_cast(t, None)`'s behavior and the four-state rule's read half.
- An empty list is indistinguishable from absent on read-back, which
  is the measured behavior and is stated in the projection's doc comment
  rather than left for a caller to discover.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- **The golden's 69 projected values are asserted against a hard-coded
  expectation derived from the provenance record**, not from the writer
  and not from the projection itself.
- A precedence table per multi-style field, each row covering: the first
  style truthy; the first style empty and a later one truthy; and a
  `"0"` in the first style beside a real value in a later one.
- The two last-style fields have their own rows, asserting absent for an
  empty written value **and** a value for a non-empty one, so the rule is
  proven to apply only when nothing is truthy.
- The date rows recompose `DATE=2001` to year 2001 with month and day
  absent, and `DATE=0000` to year 0 with month and day absent.
- The R128 rows divide by 2^8, and the ReplayGain gain rows strip
  `" dB"` before the cast.
- A whole-set comparison against the golden is one equality rather than
  69 assertions.
- The exported entry point returns the two sentinels of C14 unchanged
  from the parse, asserted through `errors.Is`.
- The doc comment states the last-style rule and the empty-list
  ambiguity, and states nothing the signature says.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Break on the first non-empty style instead of the first truthy one →
   the `"0"` precedence rows fail.
2. Return the first style's value when nothing is truthy instead of the
   last → the `catalognum` and `barcode` empty rows fail, reporting an
   empty string where absent is wanted.
3. Read a list-valued field's first value only → the list rows fail,
   reporting one element where several are wanted.
4. Divide the R128 pair by 2^7 → the R128 rows fail.
5. Strip the suffix after the cast instead of before → **unfired by
   construction, and the plan says so rather than asking the task to
   discover it.** `safe_cast` to float matches the leading numeric
   prefix, `[+-]?([0-9]+\.?[0-9]*|[0-9]*\.[0-9]+)`, at
   `mediafile/utils/type_conversion.py:52`, so `"-7.50 dB"` casts to -7.5
   either way, and `" dB"` is the only suffix on any FLAC-applicable
   style (`__init__.py:836` and `:846`; both peaks carry `float_places=6`
   and no suffix at `:856` and `:864`). The ordering is unobservable
   across all 69 keys. The task reports it unfired with this citation and
   spends no round on it.
6. Project an absent key as its typed zero → the absent rows fail across
   every type.
7. Reverse two field names in the projection, for example `disctotal` and
   `tracktotal` → the golden's whole-set comparison fails on both, which
   is the drift a per-field test would hide.

**2.0 hook: neutral, and it enables hook 11's read half.** The projection
is what lets rung 9 read the gain values back after computing them.

**Prior art:** adopts the shadow reader's eighteen-row projection as the
nucleus and widens it to 69 (`spec:1026`). No external library is
consulted, because none implements mediafile's precedence rules.

**Deliverable count: 3** (the exported entry point, the 69-key
projection, the precedence and last-style rules). Gate green; commit.
**Report the measured test wall time and the mutation outcomes.**

---

## Task 7: The writer, the merge contract, and the golden byte assertions

Model: `sonnet`. Depends on Tasks 1, 2, 3, 4 and 5.

**Files:**
- Create: `internal/engine/tagio/write.go` + `write_test.go`

**Interfaces:**
- Consumes: the streaming write (Task 1), the parsed container (Task 3),
  the tag set (Task 4), the encodings (Task 5), the golden (Task 2).
- Produces: the package's exported write entry point over a path, taking
  the tag set and the file mode.

**Outcomes:**

- The merge contract of C8. A write replaces only the keys the tag
  set owns, and every other comment survives in place and in order. Key
  matching is case-insensitive per RFC 9639 section 8.6.
- Every writable style is written (C4), so `albumartist` emits three
  comments and `disctotal` emits three.
- The vendor is preserved on rewrite and named fresh on creation
  (C9, R9).
- Every block the writer does not own survives (C10), under
  **re-emitted** headers, with the last-block flag owned by the writer
  (C11).
- Exactly one 8 KiB PADDING block, written last, on every write
  (rulings 5 and 6). Every incoming PADDING block is dropped.
- The audio is untouched (C12), and STREAMINFO's payload and its MD5
  are copied byte for byte.
- The write is a full rewrite through Task 1's streaming path (C13),
  with the mode passed in and the directory fsynced.
- **A malformed container is refused before the target is opened for
  writing**, so a refusal never half-writes a file.
- The on-disk comment order follows Task 2's recorded sequence, which
  is what makes the byte assertion possible.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- **The header and the payload of the emitted VORBIS_COMMENT block equal
  the golden's**, for a write of the golden's own recorded input into a
  copy of the golden (R4). Naming the header pins the three-byte
  big-endian length and the cleared last-block flag in the same
  assertion, which is C11's asymmetry. This is the pass's answer to risk
  2b.
- **The emitted comment set is asserted key by key against the comment
  sequence `golden.txt` records**, including the absence of a key for an
  absent value and the absence of `ALBUM ARTISTS` for a read-only style.
  `INV:409-427` is the published shape that record must cover, not the
  expectation itself.
- A round-trip fixture carrying one block of every type, plus one
  reserved type, asserts **byte equality of every preserved payload** and
  the original block order.
- A fixture whose comment block carries a foreign key that is not one of
  the 69 asserts that key intact, with its value and its position,
  after a write. `COPYRIGHT` and `ORGANIZATION` are the measured case
  (`spec:360-364`).
- **A fixture whose comment block carries `title=old` in lowercase is
  replaced rather than duplicated** by a write of the title field, and
  the resulting comment sequence carries exactly one title comment. That
  is C8's case-insensitive rule, and without this row a writer that
  appends without the delete ships two title comments through a fully
  green pass, because the reader takes the first truthy style.
- The block chain after a write reads **exactly one PADDING block, 8 KiB,
  last**. Writing twice over the same file yields the same padding as
  writing once.
- The frame region's hash is equal before and after a write, on a
  fixture carrying real frame bytes rather than the 42-byte template
  alone.
- **The frame region reaches the temp file by a copy from the source at
  the recorded offset**, never through a buffer, asserted by the
  reviewer's read of the diff. **This is the pass's second reviewer-read
  proof**, and it is what makes Task 1's streaming write worth having:
  Task 3 keeps the parse off the frame region and this criterion keeps
  the write off it.
- The renamed file carries the mode the caller passed, asserted with a
  mode that is not 0600 and not the temp file's default.
- One refusal test per malformed shape, each asserting **that the target
  file is unchanged on disk**, so a refusal is proven not to half-write.
- A rewrite preserves the loaded vendor exactly; a write that creates the
  block fresh emits a vendor with the `dubplate ` prefix (R9).
- A write into a container carrying no VORBIS_COMMENT block places the
  new block immediately after STREAMINFO (C9a), asserted against the
  emitted block chain. The fixture is a `buildContainer` product carrying
  STREAMINFO and nothing else, since R3 keeps every Go test inside
  `internal/engine/tagio` and the golden already has a block.
- Writing both members of each colliding pair in one call leaves the
  plural's comments, which is INV's own measurement (`INV:306-308`).
- `go.mod` unchanged. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Emit the vendor length, the comment count or a per-comment length
   big-endian → **the golden byte assertion fails and every read-back
   test still passes.** That asymmetry is the whole reason the golden
   exists, and the task reports both halves.
2. Write only the last style of a multi-style field → the comment-set
   assertion fails on `ALBUM ARTIST`, `ALBUM_ARTIST`, `DISCC` and
   `TOTALDISCS`, while the read-back projection still agrees, which is
   the second defect the golden catches.
3. Rebuild the comment block from the tag set instead of merging → the
   foreign-key fixture fails, reporting the dropped key.
4. Keep incoming PADDING blocks and append a new one → the
   exactly-one-block assertion fails, and the write-twice assertion
   reports a file 8 KiB larger.
5. Set the last-block flag on the VORBIS_COMMENT block instead of on the
   PADDING block → the block-chain assertion fails, because a reader that
   stops early never reaches the padding.
6. Recompute STREAMINFO's MD5 → the preserved-payload assertion fails on
   STREAMINFO.
7. Copy the frame region with a re-encode or a truncation → the
   frame-region hash assertion fails.
8. Open the target for writing before validating the container → the
   refusal rows' "target unchanged" assertion fails.
9. Pass `os.CreateTemp`'s default mode instead of the caller's → the mode
   assertion fails, reporting 0600.
10. Write the tag set in map order → the golden byte assertion fails on
    the comment sequence, and the colliding-pair assertion fails
    intermittently. The task reports how many runs the intermittent half
    took to fire.
11. Append an owned key without the case-insensitive delete → the
    lowercase-`title` row fails, reporting two title comments where one
    is wanted, while every read-back assertion still passes.
12. Read the source file whole before rewriting it → **no test fails**,
    and the task reports it unfired: the criterion that holds it is the
    reviewer's read for a whole-file read on the write path.

**2.0 hook: honors hook 12, and is otherwise neutral.** The writer writes
the applied value verbatim, so the unsanitized title stays in the tag,
which is what hook 12's portable layout requires of rung 7
(`S2:385-387`, `spec:1014`).

**Prior art:** declines `go-flac/go-flac` v2 as a dependency and reads
its `saveInPlace` as the design reference for the padding-shift problem
this task deliberately does not take (`spec:1028`). Declines
`go-flac/flacvorbis` v2, whose endianness is correct and which offers no
delete or replace, which is the merge contract this task's whole risk
sits in (`spec:1029`). Declines `gcottom/audiometa` on license and
adoption.

**Deliverable count: 4** (the writer with its merge contract, the block
re-emit with its padding policy and last-block flag, the atomic
full-rewrite path, the golden byte assertions). Gate green; commit.
**Report the twelve mutations' outcomes, naming the two unfired ones with
their reasons.**

---

## Task 8: The ruling 1 unit tests and the deviation note

Model: `sonnet`. Depends on Tasks 5, 6 and 7.

The conformance instrument 8b builds grades 53 of the 69 keys. This task
grades the rest, and the cases the twelve subjects never reach.

**Files:**
- Create: `internal/engine/tagio/rules_test.go`
- Create: `internal/engine/tagio/deviations.go` (the named deviation note
  as package documentation rung 15 reads)

**Interfaces:**
- Consumes: the encodings (Task 5), the projection (Task 6), the writer
  (Task 7).
- Produces: the deviation note.

**Outcomes:**

- One test per serialization rule for the sixteen ledger-silent keys:
  `acoustid_fingerprint`, `acoustid_id`, `comments`, `encoder`, `genres`,
  `grouping`, `initial_key`, `lyrics`, `r128_album_gain`,
  `r128_track_gain`, `remixers`, `rg_album_gain`, `rg_album_peak`,
  `rg_track_gain`, `rg_track_peak` and `subtitle` (`spec:577-596`). Those
  sixteen carry the float formatting, the suffix and the Q8 shift, which
  is the trickiest quarter of the surface.
- `bpm` keeps a test, even though it is graded: it is the integer 0
  on all twelve rows, the oracle writes `BPM=0` and reads back 0, and a
  zero that must survive a write is worth pinning.
- The cases the corpus cannot reach: the zero year, the empty list,
  `catalognum` and `barcode` empty, an invalid UTF-8 sequence, and the
  U+007E field name (ruling 1, `spec:201-207`).
- The deviation note lists what the corpus does not cover and every
  place `tagio` deliberately differs from the oracle, so rung 15 reads a
  statement rather than inferring one from a passing suite. It names at
  least:
  1. The U+007E field-name divergence from mutagen's bound, and the
     refusal above U+007E that Task 3 lands (C7).
  2. The date triple composed as a unit, which diverges from mediafile's
     read-modify-write only for a partial update `tagio` has no surface
     for.
  3. The empty list being indistinguishable from absent on read-back.
  4. The PADDING reservation at 8 KiB against mutagen's 1 KiB.
  5. The fresh vendor string.
  6. The sixteen keys the ledger cannot grade.

  A second SEEKTABLE block is no longer on this list: ruled, conductor
  2026-09-08, it is a refusal Task 3 grades (C7), not a divergence this
  note carries.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Every test's expected string is **hard-coded** and cites the mediafile
  line it reproduces. No expectation is derived from the code under test.
- Each of the sixteen keys has at least one row exercising its own rule,
  and the four ReplayGain keys and two R128 keys have rows for the
  formatting, the suffix and the shift.
- The zero-year row asserts `DATE=0000`, `YEAR=0` and `ORIGINALDATE=0000`
  on write and year 0 with month and day absent on read-back.
- The `catalognum` and `barcode` empty rows assert an empty comment on
  disk **and** absent on read-back, which is the pair C5 exists for.
- The deviation note is package documentation rather than a Markdown file
  under `docs/`, so it travels with the code and vale lints it.
- The note names each deviation with the ruling or measurement behind it,
  and claims no coverage the tests do not have.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Change one hard-coded expectation to be computed from the encoder →
   the reviewer rejects it, and the task reports the shape as the
   self-derived-expectation anti-pattern it is. **This mutation is a
   review check rather than a test run**, and the task says so.
2. Drop the `" dB"` suffix → the two gain rows fail here as well as in
   Task 5, which is the intended redundancy: these rows grade the keys
   the ledger never reaches.
3. Write `BPM` as absent when the value is 0 → the `bpm` row fails.
4. Remove one key from the deviation note → no test fails, and the task
   reports it unfired. The note's completeness is the reviewer's read
   against this task's list, and the task names that as the mechanism.

**2.0 hook: neutral.** The note is what rung 15 reads at the launch gate.

**Prior art:** ruling 1 declines new derived corpus subjects, because
deriving them is main-tree `make derive` work under the pinned venv and
reopens a frozen corpus record for keys no MusicBrainz match supplies
(`spec:598-606`, `RE` cell (8,3)).

**Deliverable count: 3** (the ledger-silent key tests, the
corpus-unreachable case tests, the deviation note). Gate green; commit.

---

## Task 9: The pass close, and the 8a to 8b handoff

Model: `sonnet`. Depends on Tasks 0 through 8.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit; the implementer never commits a ledger)
- Draft, handed to the conductor in the task return: the spec
  corrections, the filed notes, and the 8b handoff

**Outcomes:**

- The HISTORY entry in the house shape: what landed, what the gate
  caught, and what a later pass would be wrong to rediscover.
- The package contract coverage table, C1 through C16 including C7a
  and C9a, each with its Go site, what proves it, and, for a rule no test
  can reach, the stated reason.
- The ten Resolutions are listed with their markers, so the fold that
  follows a conductor ruling is a search rather than a re-reading. Each
  row names the file and the symbol the marker sits on.
- **The declared coverage gaps are recorded as gaps, not as passing
  checks**, on one criterion: a surface 8a itself builds that no test in
  this pass can reach. The expected list, to be corrected against what
  execution found:
  1. **The directory fsync's crash property is unproven** (Task 1,
     mutation 2). It is held by the doc comment and ruling 12.
  2. **The streaming property is held by a reviewer's read** rather than
     by an assertion, at three sites: Task 1's buffered-write mutation,
     Task 3's whole-file-read mutation (C7a), and Task 7's whole-file
     read on the write path. **All three hold one property**, that a
     multi-gigabyte rewrite costs a fixed amount of memory, and none of
     them is a test.
  3. **`atomicfile`'s cross-filesystem rename behavior is untested** on a
     single-filesystem workstation, inherited from 4a's own gap list.
  4. **No test grades against the twelve ledger subjects.** That is 8b's
     whole pass, and 8a's read-back agreement rests on the golden and on
     ruling 1's unit tests until it runs.
  5. **PICTURE is parsed and preserved, never decoded.** 8b owns its
     fields and its geometry rule.
  6. **No real library file has been written.** Every fixture is the
     golden, a `buildContainer` product, or a writer product; a real
     beets-written library file is first written at rung 9.
  **A container carrying two SEEKTABLE blocks is no longer a gap on this
  list**: ruled, conductor 2026-09-08, it is refused and Task 3 grades
  it, so it moves from a coverage gap to a tested refusal.
  **The count is stated and joined to the running total** of thirty
  (rung 3's twenty-one plus rung 4a's nine), so rung 15 does not
  re-count.
- The spec corrections the pass owes, drafted for the conductor:
  - The errata `spec:1049-1090` already names, if the close applies them
    at this pass rather than at 8c: the build spec's ladder entry 8
    taglib erratum (owner ruling 2), and the 2.0 spec's F1 launch clause
    (contingent on fork B).
  - CLAUDE.md's gate section says the ruff leg covers six tracked Python
    files. It covers eight today and **nine after Task 2**. The one-line
    correction is drafted here.
  - Any correction Task 2's provenance record forced on this plan's
    comment-order hypothesis or on C9a's fresh-block position.
  - **The spec's 8a exit gate (`spec:1102`) says the emitted comment set
    is asserted key by key against `INV:409-427`.** That citation holds
    the emitted output rather than the input, and 8a asserts against
    `golden.txt`'s recorded sequence instead. The one-line correction is
    drafted here, with B1 of the plan review as its evidence.
  - **The `spec:1103` scope correction R7 crosses**, ruled by the
    conductor 2026-09-08: `spec:1103` scopes the unsupported-container
    sentinel to 8b alongside the seam registry, but R7 lands that
    sentinel in 8a and 8b adopts it unchanged, so `spec:1103` is corrected
    to scope the registry and its dispatch to 8b while naming the
    sentinel as already landed.
  - **The seven folds of revision 2 the conductor ruled on 2026-09-08**,
    each with the ruling recorded, whether execution confirmed it, and
    (where the ruling reversed the draft reading) the edit that followed:
    the 4.5M ceiling, C7's field-name refusal, C7's SEEKTABLE refusal
    (reversed from the draft's non-refusal, with ruling 9's matching
    correction), C9a's fresh-block position, the date triple's
    compose-as-a-unit rule, R2's widened golden input, and R7's sentinel
    crossing.
- The 8b handoff, naming what 8b inherits: the exported reader it
  calls rather than reimplementing, the golden it grades PICTURE against
  without regenerating, the tag set its instrument fills, and the
  sentinels its registry adopts.
- The evidence, not a bare "green": the golden's size and block
  chain; the mutation table across Tasks 1 through 8 with every outcome,
  including the deliberately unfired ones; the measured wall times for
  `internal/engine/tagio`, `go test ./...` and `bash scripts/check.sh`;
  and confirmation that no existing suite changed, since this pass
  touches no landed package but `internal/atomicfile`.
- Both budgets scored: tokens against the 4.5M ceiling, and attended
  time as two counts. The target is zero planning misses and zero
  execution sittings.

**Acceptance criteria:**

- Every gap in the list is a gap this pass created or inherited, with the
  task and the mutation that names it.
- The mutation table has a row per named mutation across every task, and
  every unfired mutation carries its reason.
- `docs/STATUS.md` stays present tense and under 60 lines after the edit,
  with the pass ledger moved to HISTORY.
- No ledger is committed by the implementer.

**2.0 hook: neutral.** The close records hooks 13 and 11 as landed and
enabled by this pass, and hook 3 as 8b's.

**Deliverable count: 4** (the HISTORY entry, the STATUS edit, the drafted
corrections, the 8b handoff).

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain
(`~/.claude/workflows/pass-execute.js`) per task, **strictly sequential**.
Nothing here runs in parallel, and no second executor enters this tree.
Every task runs `cairn-implementer`, **instructed in its dispatch to read
and follow `.claude/agents/dubplate-implementer.md`**, then `diff-reviewer`
(`claude-opus-5`), then the gate, with one re-dispatch on `fix` and a
second `fix` as the conductor's decision. The conductor dispatches
`code-simplifier` over each accepted task's changed Go after review and
before the pass close, and owns every ledger edit, the lane registry, the
drafted corrections and the batch commit. **The runaway guard is armed at
launch. Subagents never dispatch agents.**

**Models.** `sonnet` by default. **One named upshift to `opus`:**

- Task 6, the exported read projection. It is a full reimplementation
  of `MediaField.__get__` for 69 fields, its last-style rule fires only
  when nothing is truthy and is invisible to a single-key test, and the
  `"0"` truthiness trap sits inside the precedence walk. The spec calls
  it the largest single item in the pass.

**Task 0 also runs at `opus`**, fresh context, as the brief for this pass
requires: a staleness check that misses a moved citation costs more than
the model difference.

**One named runner-up**, upshifted on a condition rather than by policy.
**Task 7**, the writer, if Task 5's review returns `fix`. Task 7 composes
the byte layout, the merge contract, the padding policy and the atomic
path in one place, and its correctness rests on encodings Task 5 landed.
The VORBIS_COMMENT framing and the float encodings are fully specified by
INV's table, so Task 7 stays `sonnet` unless that evidence weakens.

**Review.** Rung 8 is not one of the three full adversarial-cycle rungs
(5, 12 and 16), so review scales to the in-chain `diff-reviewer` per task
plus the pass-end gate (`spec:1292-1299`). The golden's byte assertions
sit at 8a rather than 8b for the defects the conformance run cannot see.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- Task 1 → Task 7. The writer writes through the streaming path.
- Task 2 → Task 3, Task 6, Task 7. The golden is the external anchor
  for the parse, the projection and the emitted bytes.
- Task 3 → Task 6, Task 7.
- Task 4 → Task 5 → Task 6, Task 7.
- Tasks 5, 6, 7 → Task 8.
- Tasks 0 to 8 → Task 9.
- Tasks 1, 2 and 4 are mutually independent by files (`internal/atomicfile`,
  `internal/engine/tagio/testdata`, `internal/engine/tagio/tagset.go`).
  **They still run sequentially**, because one tree runs one executor and
  the module gate runs over `./...` regardless. The conductor may reorder
  them if a dependency stalls.

**Where this pass splits, if it splits.** Ten tasks including the
pre-task and the close is inside the spec's fifteen-task trigger
(`spec:1116-1121`), so **the recommendation is to keep one pass.** The
cut point, if the Task 3 re-derivation puts the ceiling past 5.5M, is the
**Task 3 / Task 4 boundary**, which is where the risk changes shape from
bytes to semantics, exactly as the spec names it. Tasks 0 to 3 would be
8a (the streaming write, the golden, the container reader), and Tasks 4
to 9 would be 8a-ii (the value model, the encodings, the projection, the
writer, the rules and the close). **Two reasons argue against taking it.**
The first half alone ships a package that reads containers and writes
nothing, which is a surface no consumer can use and no gate can grade.
And the writer's golden assertion is the pass's answer to risk 2b, so
splitting would put a full close and a full staleness pre-task between
the golden and the assertion that uses it.

**Budget:** 4.5M ceiling, ruled by the conductor 2026-09-08, re-derived
after Task 3. Rung 8's combined ceiling stands at 8.5M against the spec's
8M reservation, the difference recorded for 8c to give back, re-derived
when 8b closes. Flag at 80%,
which is 3.6M, routed to the conductor under the 2026-09-07
unattended-window grant; STATUS after every four tasks. Spend is scored
at close against the ceiling and against **zero attended interactions**,
which is this pass's target. The ten Resolutions stand as applied,
ruled by the conductor 2026-09-08, so none counts as an execution
sitting.
