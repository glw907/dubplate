# Rung 8, Pass 8c (Absorb and Repoint) Implementation Plan

Status: **revision 2.1, 2026-09-12: the conductor-owed folds are ruled;
see the ratify-or-reverse table in
`plans/2026-09-12-overnight-window.md`.**

**Revision 2, 2026-09-08.** Revision 1 went through a fresh-context
adversarial review, one lens covering absorption correctness and plan
integrity, folded here. The record is
`docs/research/2026-09-08-rung-8c-plan-review.md`, whose
21-item "Verified sound" list is the do-not-re-litigate record for this
pass and whose three corrections to the review's own evidence carry the
measurements. Thirteen findings folded and none was rejected. Three
blockers: R1's justification reads inherited permission out of three
records that all hand the combined-read pick to rung 9; the header claims
the absorbed reader routes through the seam registry when under R1 as
applied it never reaches it; a concurrent lane 2 is asserted as fact in
three places while no worktree exists. Three majors: the projection has
eighteen rows rather than nineteen and two of them cannot carry a
distinguishable value; `TestReadFLACStylePrecedence` has sixteen rows
rather than eighteen; A1 forbids container knowledge in `shadow` and Task
3 then splices a block there with no method named. Seven minors, the
largest of which pre-authorizes Task 2 to grade three parser refusals no
test reaches today. **Three folds are conductor-owed and are marked
inline where they sit**: R1's departure ruling, including which half of
the 8b handoff R1 answers; **R2's seven reader-test destinations**; and
R4 item 2's composed denominator. R2 is conductor-owed as revision 1
wrote it, so the conductor rules R1 and R2 before Task 2 dispatches and
R4's item 2 at its own box step. **The
task count does not change: Task 0 plus four.** The ceiling holds at 2.1M
with its 80% flag at 1.68M, and the expected total moves from 1.87M to
1.92M for the rows the fold adds inside Tasks 1 and 2. Ten citation
ranges across fourteen sites were re-resolved and corrected. Three of the
review's proposed folds were rewritten rather than taken as offered, each
argued in the record.

Drafted against the rung 8 spec at revision 2 with both close errata in
its header
(`docs/specs/2026-09-07-rung-8-tag-io.md`, last landed at
`3604294`), whose fourteen conductor rulings are applied and whose four
owner rulings are ruled. The pass-split table's 8c row (`spec:1168`) is
this plan's scope statement. Every `spec:`, `docs/HISTORY.md:` and
`file:line` pointer below was resolved against the tree at `6a9570d`,
with `internal/engine/tagio` as pass 8b's simplifier round left it
(`ce0ac2e`) and `internal/engine/shadow` as pass 4b's leftovers round
left it (`a86475c`).

**What this pass is.** 8c is the last of rung 8's three passes, and it
closes the rung. It absorbs the shadow's provisional FLAC reader into
`tagio`, repoints the four sites that name it, retires the fixture
builder three test files share, and lands the one behavioral change the
rung owes the box: ruling 9's refusals now apply to the shadow's read
path (`spec:310-317`, `spec:1012-1033`). **8a landed the container
parse, the 69-key value model, the exported read projection, the writer
and the golden container. 8b landed PICTURE, the seam registry and the
conformance instrument.** Neither is reopened here. A task that believes
it needs a new `tagio` refusal, a new corpus subject, a `make derive`
run, or a box write stops and reports.

**What 8c may assume, from the 8b handoff** (`docs/HISTORY.md:856-879`).
`Container.Pictures` and `Write`'s picture argument are landed, so the
absorbed reader is written against the widened shapes rather than 8a's.
**The handoff paragraph carries two halves about the read path, and this
plan answers both rather than quoting one.** Its first half says 8c's
reader routes through the seam registry the way `Read` and `Write` do,
with `Parse` staying the one deliberate bypass (m6), so the absorbed
reader is a registry consumer rather than a second bypass
(`docs/HISTORY.md:861-864`). Its second half permits 8c's absorbed reader
to keep reading through `Parse` and `Container.Pictures` directly, since
it already holds a parsed `Container` (`docs/HISTORY.md:871-875`). **R1 as
applied takes the second half**: the shadow's read path consumes `Parse`
and never reaches the registry, so no `shadow` call site exercises
`Read` or `Sniff` after this pass. **R1's fold takes the first half**, at
the cost of a duplicate parse. Either way this pass adds no bypass of its
own, because `Tags` opens no file. Which half stands is the conductor's
ruling on R1, and it is **conductor-owed** rather than assumed here.
`TestConformanceLedgerSubjects`
(`internal/engine/tagio/conform_test.go:640`, selected by `go test
./internal/engine/tagio/... -run Conformance -count=1`) must stay green
as the shadow's reader moves, and a regression there is a defect in the
absorption rather than a reason to relax the instrument
(`docs/HISTORY.md:864-869`). The `io.ReaderAt` call is 8c's alone (R8 of
the 8b plan, `docs/HISTORY.md:869-871`), and R5 below takes it. R12's
read-side answer and the double parse it accepts are settled rather than
open, and what the handoff settles is the double parse rather than a
choice of combined shape (`docs/HISTORY.md:871-875`). The tightened
refusal set is 8a's rules, unchanged by 8b, so this pass's capture-path
work is written against `container.go:35-42` and `TestParseRefusals`
(`container_test.go:118`) as they stand.

**The pass's own risk control.** Risk 6 (`spec:1320-1330`) is that the
absorption changes the shadow, whose reader is live on the box from rung
4's deploy and whose fixture builder is shared by three test files.
Everything in Tasks 2 and 3 that grades the projection row by row, and
every criterion naming the shadow package's own suite, exists for that
risk. Risk 2b (`spec:1294-1301`) is a defect the conformance instrument
cannot see, and the instrument staying green is a criterion in the task
that could break it rather than a hope at the close.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section, the pass-execute chain (`~/.claude/workflows/pass-execute.js`)
> per task, sequential. The implementer is `dubplate-implementer`
> (`.claude/agents/dubplate-implementer.md`), read and followed before
> anything else. The conductor consumes reports and never reads diffs
> inline. Subagents never dispatch agents; `code-simplifier` is the
> conductor's dispatch. The conductor owns the ledgers
> (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`), the lane
> registry, the contested-surface rows, the read-only box evidence step,
> and the batch commit of anything a task drafts into them. Rung 8 is
> not an adversarial-cycle rung (those are 5, 12 and 16), so review
> scales to the in-chain `diff-reviewer` (`claude-opus-5`) per task plus
> the pass-end gate (`spec:1356-1364`). **This pass runs in LANE 1, the
> main tree.** Whether a lane 2 worktree is live alongside it is not
> assumed either way: at drafting there is none, `git worktree list`
> reports the main tree alone, and `docs/STATUS.md:59-64` records both
> lanes idle with pass 5b owed. **Task 0 item 1's registry read settles
> it before the first dispatch.** No task in this pass runs from a
> worktree, and no task reads or writes anything in lane 2's territory
> whether or not a lane 2 is running.

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/specs/2026-09-07-rung-8-tag-io.md`, revision 2,
   both header errata included.** The 8c row of the pass-split table
   (`spec:1168`) is the scope statement. The sections this pass
   implements are "What the absorption costs" (`spec:978-1011`), "The
   read path's strictness is a deliberate change, and 8c owns it"
   (`spec:1012-1033`), and the closing paragraph on the projection
   (`spec:1034-1038`). The rulings it applies are ruling 3
   (`spec:268-272`), ruling 4 (`spec:273-279`), ruling 9
   (`spec:310-317`), ruling 13 (`spec:337`) and ruling 14
   (`spec:338-342`). Read "The container contract" (`spec:471-528`) for
   what a refusal means in format terms, "What rung 8 does not build"
   (`spec:932-955`) for the boundary, "Edges to later rungs"
   (`spec:1039-1063`) for what rungs 9 and 15 inherit, and "Risks"
   (`spec:1278-1338`) for what each task's assertions exist to catch.
2. **`docs/HISTORY.md:544-1012`, pass 8b's entry**, and inside it the 8b
   to 8c handoff (`:856-879`), the forward carries (`:880-898`), what
   landed (`:557-672`), the declared coverage gaps (`:810-855`), and what
   a later pass would be wrong to rediscover (`:974-994`). The handoff
   names what 8c inherits; the carries name what 8c does not decide.
3. **`docs/HISTORY.md:1013-1420`, pass 8a's entry**, and inside it what
   landed (`:1022-1113`), the declared coverage gaps (`:1271-1309`,
   gap 7 at `:1301-1309` being the `io.ReaderAt` gap R5 takes), and what
   a later pass would be wrong to rediscover (`:1331-1352`). 8a's refusal
   set and the reader's shape are the rules this pass tightens the
   capture path to.
4. **`docs/research/2026-09-08-rung-8b-plan-review.md`**, and
   inside it the merged 42-item "Verified sound" list (`:683` onward).
   That list is the do-not-re-litigate record for rung 8's plan shape,
   and it covers the assertion-discipline bar, the mutation-reporting
   shape, and the ceiling derivation this plan reuses.
5. **The tree, read-only, before any edit:**
   `internal/engine/shadow/flac.go` in full (313 lines),
   `flac_test.go` in full (361 lines), `capture.go:123-186`,
   `capture_test.go:19-59`, `run.go:190-215`, `run_test.go:80-116`,
   `reconcile.go:28-40`, `itembag.go` in full, and
   `observe_test.go:105-125`; then
   `internal/engine/tagio/doc.go`, `container.go`, `read.go`,
   `seam.go`, `write.go:34-60`, `tagset.go:128-198`, and
   `container_test.go:1-120`. **Count the `buildFLAC` call sites
   yourself** rather than trusting this plan's count, and report what you
   find (Task 0, item 3).
6. **`docs/specs/2026-09-07-streamer-2-0-design.md`'s forward
   hooks** (`S2:354-391`). Rung 8's own hook positions are the spec's
   table (`spec:1065-1080`): hooks 13 and 14 are recorded, hook 3 landed
   at 8b, and 8c lands none. Every task below states its position anyway.
7. **`docs/specs/2026-09-03-parallel-build-approach.md`'s
   ownership table** (`:175-220`), whose `internal/engine/shadow` row
   (`:199`) says no lane but rung 4 writes that package. Rung 8's own
   contested-surfaces table amends it (`spec:1197`), and **the conductor
   writes the amendment before this lane launches**, per lane-launch
   checklist item 2.
8. **`CLAUDE.md`** in full, and `docs/STATUS.md`'s lane registry
   (`docs/STATUS.md:59-64`).
9. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** This
   pass writes no Python, no shell, no systemd unit and no deploy path,
   so `python-conventions` and `vps-conventions` are both out of scope
   (`spec:1351-1353`).

## Goal

Ship the one deliverable the spec assigns to 8c: the absorption, with
the tightening it lands stated rather than discovered.

Land the shadow reading one FLAC file through `internal/engine/tagio`
and through nothing else, with `internal/engine/shadow` keeping only its
projection to `matcher.ItemFields` (ruling 3, `spec:268-272`). Land the
four sites that name the provisional reader repointed or rewritten.
Land the three external `buildFLAC` uses on fixtures the real writer
built, and `buildFLAC` with its STREAMINFO layout test deleted (ruling
4, `spec:273-279`). Land each of `flac_test.go`'s seven tests at a named
destination, with the itembag round trip staying in `shadow`. Land a
capture-path test that asserts ruling 9's tightened refusal, and an
enumeration of every shape the tightening adds, each naming the `tagio`
test that grades it.

Nothing here touches the box, deploys anything, writes under `corpus/`
or `internal/harness/`, adds a module dependency, reaches the live
network, runs the venv, declares a configuration key, regenerates the
golden, or adds a `tagio` refusal.

## Owner touchpoints: none expected

8c raises no attended question by design. Fork B was Geoff's and is
ruled (2026-09-08, the spec's owner ruling 5): 1.0 ships FLAC-only and
the other formats target 1.1. 8c's surface is identical under either
answer, because this pass registers no format and removes one duplicate
reader; the close applies the F1 erratum as ruled rather than
provisionally.
Rung 8 declares no configuration key (ruling 13, `spec:337`), so no
template gains a name and no default reaches an adopter through this
pass.

**What would become an owner touchpoint, so a task recognizes one rather
than absorbing it.**

- **A divergence between the shadow's landed projection and the
  projection over `tagio`.** This pass's whole exit gate is that the
  projection's behavior does not change, so a row that cannot be made to
  agree is a finding for the conductor, never a test to relax and never
  a value to adjust. R6 below names the one value change this pass
  accepts in advance, and any second one stops the task.
- **Any adopter-visible behavior**: a reject wording, an accepted-format
  decision, a configuration default, or a change to what the intake gate
  admits. Those belong to rung 7's string authority and rung 7's config,
  bounded by what this rung supports (`spec:801-807`).
- **Any box write.** The box (`5.78.148.42`) is production. This pass
  deploys nothing, and the tightening reaches the live shadow only at
  the next deploy of the shadow units, which is not this pass's and not
  rung 8's (`spec:1057-1059`). The one box contact this pass allows is
  the conductor's read-only evidence step in Task 3, described there.
- **Any key added to `env/dubplate.env.template`.** That file is a
  secrets manifest rendered by `deploy.sh` from the age store, and
  adding a name there requires minting the secret (CLAUDE.md, "Facts
  agents get wrong"). Rung 8 adds none.

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition below resolves to the conductor. The 80%-of-ceiling
flag routes to the conductor rather than to Geoff (Geoff, 2026-09-08).

## Spec

`docs/specs/2026-09-07-rung-8-tag-io.md` (revision 2, with
the 2026-09-08 errata), and behind it
`specs/2026-09-02-greenfield-server-build.md` (revision 2.1, ratified),
ladder entry 8. 8c builds deliverable 7 of the spec's seven
(`spec:129`); deliverables 1 and 2 landed at 8a, and 3, 4 and 5 at
8b, with 6's golden container landed at 8a and graded at both.

**The forward criterion this pass hands rung 15**, so rung 15 reads it
from the ledger rather than rediscovering it:

> Rung 15 re-runs 8b's conformance instrument unchanged, `go test
> ./internal/engine/tagio/... -run Conformance -count=1`, which selects
> `TestConformanceLedgerSubjects` by name, and checks the reported test
> count rather than the exit status alone. 8c adds nothing to that
> re-run. What 8c adds to rung 15's reading is the deviation note as it
> stands after this pass and the tightening enumeration Task 3 produces:
> the shadow's capture path refuses what ruling 9 refuses, so a rung 15
> re-read of the shadow's evidence is read against one reader rather
> than two.

## Token ceiling: 2.1M

The spec proposes 1M for 8c (`spec:1214`). The arithmetic is
re-derived here from 8b's actual report rather than carried, because a
ceiling quoted without its derivation is a number nobody can check, and
because the spec's own clause requires it: "Each ceiling is re-derived
from the previous pass's actual report before the next plan is drafted"
(`spec:1226-1228`).

**The basis, read off `docs/HISTORY.md:1004-1012`, pass 8b's spend.** 8b
closed at about **2.9M against a 3.25M ceiling** for Task 0 plus seven
implementer tasks. Itemized: plan preparation 0.87M; Task 0, 0.11M; the
chain 1.51M across 16 agents; the leftovers round 0.18M; the batched
`code-simplifier` round 0.13M; conductor turns about 0.1M. The 16-agent
figure over seven tasks means the chain line **already includes** every
`diff-reviewer` pass and every fix re-dispatch, so the measured per-task
rate is 1.51M over seven, about **216K**, with review and fixes inside
it. A separate fix-round line would double-count.

**The arithmetic for 8c.**

| Line | Basis | Figure |
| --- | --- | --- |
| Prep, this plan's authoring workflow plus one adversarial review lens and its fold | Below 8b's 0.87M, which merged two lenses. The spec is unchanged, the handoff is written, and this plan reads a subset of 8b's inputs | 0.5M |
| Task 0, staleness, `opus` at zero context | 8b measured 0.11M for a fourteen-item checklist | 0.11M |
| Tasks 1 through 4 through the chain | Four implementer tasks at 8b's measured 216K, reviews and fix rounds included | 0.86M |
| One batched `code-simplifier` round over 8c's Go, plus its review | 8b's round measured 0.13M against a 0.2M line | 0.15M |
| A leftovers round, if one is needed | 8b measured 0.18M over seven tasks; four tasks at the same rate is about 0.1M | 0.1M |
| The conductor's read-only box evidence step | One quiet-box check and one `dubplate-server shadow report` reading, with no file read into context beyond the counts | 0.05M |
| Conductor overhead | 8b measured 0.1M | 0.1M |
| The review fold's own added rows | Two refusal rows in `container_test.go` (m2), one permitted seed variant and one gap entry (m4), and the restated criteria, all inside Tasks 1 and 2 | 0.05M |
| **Expected total** | | **1.92M** |

**The ceiling is proposed at 2.1M, because 1M is a close-shaped figure
for a four-task chain.** The spec's 1M line sits inside the budget
table's close-shaped range of 0.15M to 0.5M plus a small chain, and it
was set before any 8c task list existed. Prep and Task 0 spend 0.61M
before Task 1 dispatches, which is more than half the spec's whole
proposal, so an 80% flag against 1M would fire before the first
implementer ran. Walking forward from 0.61M: Tasks 1 through 3 reach
1.26M, the review fold's added rows and the batched simplifier round take
it to 1.46M, Task 4 closes the chain at 1.68M, and the leftovers round,
the box step and conductor turns land the expectation at 1.92M. An 80%
flag against 2.1M lands at **1.68M**, which falls at Task 4's dispatch,
where a stop is cheap and the conductor's action is a re-derivation this
plan already schedules. The 0.18M of headroom over the expectation is a
narrow band, and it stays narrow rather than moving the ceiling, because
the fold added rows inside existing tasks rather than a task.

**Recorded against the rung reservation.** `spec:1214` reserves 8M for
rung 8 (4M, 3M, 1M). Ceilings now stand at 4.5M, 3.25M and 2.1M, or
**9.85M**. The actuals tell the other half: 8a spent 3.6M against 4.5M
and 8b spent 2.9M against 3.25M, so the rung is 1.25M under its own
ceilings with one pass left, and the expected rung total is 6.5M plus
1.92M, or **8.42M against the 8M reservation**. The overshoot is
prep the reservation never priced, and Task 4 drafts the spec correction
that records it rather than leaving the spec's 1M line to read as a miss.

**Flag at 80%, which is 1.68M.** Finish the task in flight, write
STATUS, and send one combined question to the conductor. **The
conductor's action when it fires is the re-derivation this plan already
schedules**, since the flag is a forecast the plan made rather than a
surprise.

**Re-derive the ceiling from Task 2's actual report.** Task 2 is the
pass's one substantive change and its only multi-file edit, which is the
shape 8b measured most precisely. If the re-derivation passes 3M, the
conductor checks the window cap before continuing and names the split at
the **Task 2 / Task 3 boundary**: Tasks 0 to 2 would be 8c (the
absorption), and Tasks 3 and 4 would be 8c-ii (the tightening test, the
evidence and the close). That cut leaves both halves shippable, because
the absorption is complete without the tightening's own test and the
tightening is a test plus a report. The argument against taking it is
that the tightening is the reason 8c is a pass rather than a chore
(`spec:1173-1175`), and closing the rung without it would leave the
rung's one behavioral change ungraded.

**Checkpoint: STATUS after every four tasks**, so after Task 4, which is
the close. Every task reports its own spend, its mutation outcomes, and
its measured gate wall time, and the conductor sizes the next task
against it.

## Global constraints

- **Main tree, lane 1.** The spec's tree column for 8c is "Main tree"
  (`spec:1168`). The lane registry entry is written before the first
  dispatch, and **the live-executor check is that registry, never
  `pgrep -f Projects/dubplate`**, which matches every lane once a
  worktree exists (CLAUDE.md, Lanes).
- **Package territory, stated as a wall.** This pass writes
  **`internal/engine/shadow`** and, under R1, at most one exported
  symbol in **`internal/engine/tagio`**. It writes **nothing** under
  `corpus/`, `internal/harness/`, `internal/engine/{matcher,
  musicbrainz,tag}`, `internal/atomicfile`, `internal/config`,
  `internal/serve`, `internal/identity`, `internal/auth`,
  `internal/notify`, `cmd/`, `systemd/`, `scripts/`, `env/`, `Makefile`,
  `tools/`, `web/`, or `_salvage/`, and no `docs/` file outside Task 4's
  conductor-owned drafts.
- **`tagio` never imports `internal/engine/matcher`** (ruling 3,
  `spec:268-272`). The absorption moves container knowledge into
  `tagio` and leaves the projection to `matcher.ItemFields` in
  `shadow`. `matcher.ItemFields` never appears in a `tagio` signature,
  and `internal/engine/tag/doc.go:8-11` stays true.
- **`tagio` adds no refusal.** Ruling 9's set is 8a's, closed at
  `container.go:35-42` and graded by `TestParseRefusals`
  (`container_test.go:118`). A shape this pass finds that `tagio`
  tolerates and should refuse is a finding for the conductor and a rung
  9 or rung 15 item, never a refusal added here
  (`docs/HISTORY.md:1360-1362`). **Adding a test for a refusal `tagio`
  already performs adds no refusal**, which is what lets Task 2 close the
  two ungraded shapes m2 names.
- **No interface, exported or not, inside `tagio`** until the MP3 pass
  brings the second implementation (ruling 14, `spec:338-342`). Under
  R1 this pass adds a function over a concrete type, never a seam.
- **No configuration key, stated in those words** (ruling 13,
  `spec:337`). No task edits `internal/config`, and Task 0 confirms it.
- **Module dependencies: none added.** `go.mod` carries cobra and its
  two indirects. `internal/engine/shadow` gains one intra-module import,
  `internal/engine/tagio`, and nothing else. A task that wants a
  dependency stops and reports.
- **No Python, no venv, no derivation.** The gate never requires uv or
  beets by design (`spec:905-908`). No task runs `make derive`, and no
  task derives a new corpus subject (ruling 1, `spec:932-955`).
- **No box write, and no deploy.** Rung 8 touches no box surface, no
  systemd unit and no deploy path (`spec:1057-1059`). The conductor's
  read-only evidence step in Task 3 is the pass's only box contact, it
  runs under the quiet-box check, and **no implementer task may ssh at
  all**.
- **Assertion discipline, at the spec's own bar.** No silent-success
  fakes, no vacuous assertions, and **no expectation derived from the
  code under test**. Every hard-coded expectation in this pass comes
  from the landed shadow tests' own transcribed values, from RFC 9639,
  from the pinned mediafile and mutagen sources read at their committed
  versions, or from `internal/harness/oracle/testdata/template.flac`'s
  committed bytes.
- **Universality.** No estate hostname, path, email, catalog value or
  personal name in any fixture or any comment. The shadow's existing
  fixture vocabulary (invented artists and albums) is kept.
- **`go-conventions` before any Go.** The bar is the standing
  exemplary-stdlib-idiom bar (`idiomatic-beautiful-go`). Concrete types
  over interfaces, `any` not `interface{}`, no `init()`, guard clauses
  over nesting, receivers named for their type, errors lowercase and
  context-first with `%w` only where a caller branches. **The em dash is
  banned in Go comments**, and vale scans a changed file before it is
  staged.
- **A deletion is a deliverable, and it is graded like one.** This pass
  removes about 230 lines of reader (`flac.go:13-193` plus its helpers)
  and about 300 lines of test. A criterion that says a file is gone also
  says what still proves the behavior it held.
- **The gate is `bash scripts/check.sh`, green at every task end.** `make
  check` (vet, golangci-lint, test, tidy-check) over `./...`, shellcheck
  over `scripts/*`, ruff `check` and `format --check` over every tracked
  `*.py` (**this pass adds none**), gitleaks over the tracked worktree
  with `--no-git`, vale over Go comment prose, `docker compose config`,
  a pattern-filtered `systemd-analyze verify`, and the state-dir
  agreement leg. **Never commit red. Specific files only.** The commit
  footer is CLAUDE.md's, with the session's own attribution line.
- **The gate's wall time is measured before it is explained.** 8b found
  that gitleaks from the main tree also scans a live lane 2 worktree,
  adding about three minutes to that leg
  (`docs/HISTORY.md:890-894`), and that cost applies only while such a
  worktree exists. Task 0 runs no gate, so **Task 1's gate run is this
  pass's baseline**: it reports the total wall time and the leg
  breakdown, read against Task 0 item 1's finding on whether a lane 2
  worktree exists at all. Tasks 2 and 3 compare against Task 1's
  measurement rather than against 8b's figure. A task reporting a slow
  gate names the leg that grew and says whether a worktree is present,
  rather than attributing the cost to a worktree it has not confirmed
  exists.
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each one before
  Task 1 dispatches.

## Contested surfaces

The spec's own table (`spec:1186-1204`) is the source, and it amends the
parallel-approach spec's rows. The conductor writes these into that
spec's ownership table before the lane launches, per lane-launch
checklist item 2.

| Surface | Owner | Consumers | 8c's position |
| --- | --- | --- | --- |
| `internal/engine/shadow` | rung 4, created by it; the parallel spec's row (`:199`) says no other lane writes it, and rung 8's table amends it: **rung 8 removes `flac.go` and repoints its call sites** (`spec:1197`) | rung 15's re-run, rung 17's docs | **8c writes it**, and it is the amendment the conductor lands before the lane launches. |
| `internal/engine/tagio` | rung 8, created by pass 8a; no other lane writes it | rung 7's allowlist bound, rung 9's pipeline, rung 15's re-run | **8c adds at most one exported symbol** (R1) and changes no landed behavior in `Parse`, `Read` or `Write`. |
| `internal/engine/matcher` | rung 3 (the E2 lane), except rung 4's one addition | rung 4's shadow, rung 9, rung 15 | **Untouched.** `ItemFields` is read and never widened, and no field is added to it. |
| `internal/engine/tagio/testdata`'s golden container | rung 8, generated once under the pinned venv, then committed and never regenerated | rung 15 reads it as a fixture | **8c neither reads nor regenerates it.** The shadow's fixtures use the committed 42-byte oracle template's own bytes (R3). |
| Rung 8's conformance instrument | rung 8, under `internal/engine/tagio` | rung 15's re-run | **8c does not edit it and must not break it.** Task 2 carries the criterion. |
| `corpus/` and `internal/harness/` | rung 2 and rung 3 | rungs 3, 8, 15 | **Read-only, whole.** No task writes a byte under either. |
| The box, its units and its shadow state | rung 4, deployed at 4b | rung 15's floor check | **Read-only, and only the conductor.** No task deploys, and the tightening reaches the live shadow at the next deploy rather than this pass. |
| `docs/` ledgers | conductor-owned | every rung | A pass drafts and never commits. |

**A lane 2 may or may not run alongside this pass, and either way its
territory is disjoint.** At drafting there is no lane 2 worktree: `git
worktree list` reports the main tree alone, `.claude/worktrees` is empty,
and `docs/STATUS.md:59-64` records lane-1 idle with 8c's plan next and
lane-2 idle with `lane-2-5a` torn down at 5a's merge and pass 5b owed.
Whichever session launches lane 2 next, its territory is pass 5b's from
the rung 5 identity and auth spec: `internal/identity`, `internal/auth`,
`internal/notify`, `internal/serve`, `web/`, and
`internal/config/keys.go`'s remaining reserved-class retirement
(`docs/STATUS.md:33-41`). **Nothing 8c writes is anything 5b writes**,
and nothing 8c reads is anything 5b writes. **Task 0 item 1 settles
whether a lane 2 is live rather than either half being assumed**, and
reports any file under `internal/engine/shadow` or
`internal/engine/tagio` that a lane 2 worktree has touched.

## The 8c contract, as outcomes

Every task below cites this section rather than restating it. Items are
numbered A1 through A9 to keep them distinct from 8a's C1 through C16
and 8b's P1 through P12, which stand unchanged. **The plan states the
contract; the implementer writes the code.**

**A1. `internal/engine/shadow`'s read path holds no container
knowledge.** After Task 2 the package's non-test files contain no FLAC
signature constant, no block-type constant, no metadata block walk, no
`binary.LittleEndian` or `binary.BigEndian` call, no Vorbis comment
decode, no `safe_cast` reimplementation and no date splitting. Everything
`flac.go:13-193` and `flac.go:263-307` hold today is `tagio`'s, and the
file is deleted. **The scope is the read path rather than the whole
package**, because Task 3's fixture helper is the one place in `shadow`
that touches container bytes, and it does so without a block-type
constant, without a header encoder and without importing
`encoding/binary`. Task 3 states the method that makes that true.

**A2. The projection to `matcher.ItemFields` stays in `shadow`** (ruling
3, `spec:268-272`, `spec:1034-1038`). It keeps its **eighteen** rows, its
mediafile citations, and its documented rule that an absent or
unparseable value takes the field's Go zero value (`flac.go:72-76`).
Eighteen rows cover a nineteen-member `ItemFields`
(`internal/engine/matcher/info.go:124-152`) because `DataSource` has no
row at all: a grep of pinned mediafile for `data_source` returns nothing,
so an inbox file always projects the empty string
(`flac.go:203-207`). Nineteen is the right figure only for the itembag
round trip, whose bag spells nineteen beets field names including
`data_source` (`itembag.go:14-35`). `ItemFields` carries no presence bit,
so an absent `*string` and a present empty string both project to the
empty string, and that is the behavior the landed tests already assert.

**A3. The seam the shadow injects keeps its name and its signature.**
`reconcileInput.LibraryReadFn` is `func(path string) (matcher.ItemFields,
error)` (`reconcile.go:35`), `run.go:209` passes `readFLAC`, and
`capture.go:158` calls it. R7 keeps that name, so the two call sites need
no edit and the doc comment at `reconcile.go:35` is the one that
repoints.

**A4. The track length comes from the parsed container's STREAMINFO**,
divided the way `flac.go:41-46` divides it today, with the zero-rate
guard preserved so a stream declaring no sample rate reads zero rather
than dividing by zero.

**A5. One parse per file.** Under R1 the shadow parses a file once and
projects twice from the same `Container`, once to a `TagSet` and once to
the length. Under R1's fold it parses twice and the plan says so; either
way the count is stated in the close rather than left to a reader.

**A6. The provisional framing is rewritten, never carried.**
`flac.go:62-70` calls the reader provisional and says rung 8's plan
carries the task of absorbing it. That paragraph is false the moment
this pass lands (`spec:981-986`), and its replacement documents what the
surviving function does rather than what an earlier rung deferred. The
rung 4 calendar rationale is dropped.

**A7. `buildFLAC` does not survive, and nothing replaces it in general
form** (ruling 4, `spec:273-279`). Every well-formed fixture in the
`shadow` package is built by `tagio.Write` over a seed carrying the
committed oracle template's own STREAMINFO bytes. The one malformed
fixture the tightening test needs is a byte splice over writer output
(R3), which emits only containers the writer refuses to produce.

**A8. The tightening is asserted and counted, never discovered.** A
capture-path test proves that a container the landed reader accepted and
`tagio` refuses now discards its album whole, and the pass report
enumerates every shape the tightening adds, each naming the `tagio` test
that grades it (`spec:1012-1033`).

**A9. The conformance instrument stays green.** `go test
./internal/engine/tagio/... -run Conformance -count=1` selects
`TestConformanceLedgerSubjects` and reports twelve subtests. A run
reporting zero tests is the failure the criterion exists to catch, and a
regression there is a defect in the absorption
(`docs/HISTORY.md:864-869`).

## Resolutions this plan takes rather than discovers

Each is applied as the recommendation and marked **provisional** at this
revision. **Three are marked conductor-owed** where they change an
exported surface, depart from a ruled decision, or pick a value the plan
cannot derive: R1, R2, and R4's item 2. The conductor rules each before
the task that acts on it dispatches, which for R4 item 2 is its own
read-only box step. Each names what changes if it is ruled the other way,
so a reversal is a named edit rather than a rediscovery.

### R1: the shadow parses once, and `tagio` exports its tag projection

**Conductor-owed. Ruled before Task 2 dispatches. What the conductor is
ruling is a departure from a decision 8b closed, and this resolution says
so rather than claiming an inherited option.**
**Ruled, conductor 2026-09-12:** **As applied**, stated to break
`8c:611`'s inverted use of "fold": the shadow consumes `Parse` and
`tagio.Tags`, taking the 8b handoff's **second half**; `tagio` gains one
exported symbol; one parse per file; 8b's coverage gap 4 retires; **Task
2 stays `sonnet`**. Deviation 5 is ruled with this and needs no separate
mark (`8c:1601-1602`).

**Applied.** `internal/engine/tagio` exports its projection over an
already-parsed container, `Tags(c Container) TagSet`, which is
`read.go:58`'s unexported `project` exported and documented. The shadow
calls `tagio.Parse` once per file, reads `Container.StreamInfo` for the
length, and calls `tagio.Tags` for the 69-key tag set it projects to
`matcher.ItemFields`. `Read` (`read.go:17`) stays the path-taking entry
point and stays the registry's dispatch, unchanged.

**Why, stated as the departure it is.** The shadow needs a file's tags
and its STREAMINFO from one file, and the two live on different sides of
`tagio`'s current surface: `Read` returns a `TagSet` and never a
`Container`, and `Parse` returns a `Container` and never a `TagSet`. R12
named the two shapes that close that gap, a second exported read entry
point or an exported projection over a parsed `Container`, and **it
handed that pick to rung 9, not to 8c.** Three records say so in the same
words: the spec's own erratum names "rung 9, the first caller in a
position to pick" (`spec:766-774`); 8b's coverage gap 4 is "handed to rung
9 as the first caller" (`docs/HISTORY.md:830-833`); and 8b's
do-not-rediscover section reads "rung 9 picks between a second read entry
point and an exported projection over `Container`"
(`docs/HISTORY.md:990-993`). What the 8b handoff grants 8c is the double
parse rather than the pick: 8c's reader "may keep reading through `Parse`
and `Container.Pictures` directly ... and it owes no fold into a combined
`Read`" (`docs/HISTORY.md:871-875`). **8c proposes taking the projection
shape anyway**, because the shadow is the first caller that actually wants
one file's tags and its STREAMINFO off one parse, and because the shape is
the smaller of R12's two: it opens no file, sniffs nothing, is no entry
point at all, and it turns 8b's declared coverage gap 4
(`docs/HISTORY.md:830-833`) from a standing cost into a retired one. The
cost of taking it is that rung 9 loses a decision that was assigned to
it, and the conductor is ruling that trade rather than confirming a
permission. Deviation 4 carries the same sentence, so the departure is
scored at the close rather than read later as an option 8c exercised.

**The handoff's two halves, and which one R1 answers.** The handoff's
first half says 8c's reader routes through the seam registry the way
`Read` and `Write` do, with `Parse` staying the one deliberate bypass, so
the absorbed reader is a registry consumer rather than a second bypass
(`docs/HISTORY.md:861-864`). Its second half permits 8c's absorbed reader
to read through `Parse` and `Container` directly
(`docs/HISTORY.md:871-875`). **R1 as applied takes the second half, and
the consequence is stated plainly rather than softened**: the shadow's
read path consumes `Parse`, so it never reaches the registry and no
`shadow` call site exercises `Read` or `Sniff` after this pass. `Tags`
itself adds no bypass, because a bypass is an entry point that skips the
sniff and `Tags` performs no I/O, and `Parse`'s own bypass is 8a's ruled
decision graded at `seam_test.go:135` and `:191`. The behavioral
consequence for FLAC is nil, because `Parse` validates the `fLaC`
signature itself (`container.go:147-150`), so what the conductor is ruling
here is a documentary and architectural position. **R1's fold takes the
first half instead**, and the header paragraph states both readings so a
reversal is a named edit.

**Fold if ruled otherwise.** The shadow calls `tagio.Read` for the tag
set and `tagio.Parse` for the STREAMINFO, two parses and three file
opens per track, and `tagio` gains no exported symbol at this pass. The
cost is the duplicate parse on the shadow's hot path and 8b's coverage
gap 4 staying open for rung 9. The gain is that the absorbed reader stays
a registry consumer in the handoff's own sense, because `Read` routes
through the registry, and that rung 9 keeps the pick R12 assigned it. A2
through A4 are unchanged under either answer, and only A5's count and the
header paragraph's second sentence change.

### R2: the five reader tests get named destinations, and two of them stay in `shadow`

**Conductor-owed. Ruled before Task 2 dispatches, because it departs
from the spec's own words.**
**Ruled, conductor 2026-09-12:** As applied: two of the five spec-named
tests stay in `shadow` because their subject is `matcher.ItemFields` and
ruling 3 forbids `tagio` naming that type; the seven-row destination
table stands. The alternative costs a projection with eighteen rows and
no test. **8c's execution model makes this a precondition of Task 2's
dispatch**, which the list at `8c:20-22` omitted.

**Applied.** The spec says five reader tests move to `tagio`
(`spec:995-998`). Two of the five cannot: their subject is the
projection to `matcher.ItemFields`, and ruling 3 forbids `tagio` naming
that type. Each of `flac_test.go`'s seven tests therefore gets a
destination, and the pass reports the table rather than a count:

| Test | Subject | Destination |
| --- | --- | --- |
| `TestReadFLACMapsEveryVorbisKey` (`flac_test.go:73`) | the eighteen-row projection to `ItemFields` over a real container | **Stays in `shadow`**, rewritten over a writer-built fixture (ruling 3). |
| `TestReadFLACStylePrecedence` (`:108`) | mediafile's style precedence and `safe_cast` semantics | **`tagio`'s**, already graded by `TestProjectStylePrecedence` (`read_test.go:144`), `TestProjectLastStyleRule` (`:199`), `TestProjectDates` (`:222`), `TestProjectNumericCasts` (`:249`) and `TestProjectComparesNamesCaseInsensitively` (`:300`). Task 2 walks its **sixteen** rows against those tests and adds to `tagio` any row not already covered; sixteen is the count `grep -c 'name:'` returns over `flac_test.go:114-199`. |
| `TestReadFLACLengthFromStreaminfo` (`:213`) | the seconds division and the zero-rate guard | **Stays in `shadow`** as a table over `tagio.StreamInfo` values, needing no file at all (A4). |
| `TestReadFLACParsesTheCommittedTemplateBytes` (`:259`) | a decode pinned to bytes another implementation produced | **`tagio`'s.** Task 2 checks it against `TestParseGolden` (`container_test.go:359`), which pins the decode to the oracle's own golden, and either adds the 42-byte template as a second external pin or records the coverage. |
| `TestReadFLACRefusesAMalformedContainer` (`:296`) | refusal | **Split.** The container half is `tagio`'s, graded row by row against `TestParseRefusals` (`container_test.go:118`); the capture half becomes Task 3's tightening test. |
| `TestItemBagRoundTripsThroughProjectItemFields` (`:333`) | the nineteen beets field names | **Stays in `shadow`** (`spec:999-1001`), moved to a surviving file, with `itembag.go:11-13`'s pointer at `flac_test.go` repointed to its new home. |
| `TestBuildFLACEmitsTheStreaminfoLayoutByteForByte` (`:273`) | `buildFLAC`'s own output | **Deleted** with `buildFLAC` (ruling 4, `spec:1002-1010`). Its external pin's purpose already lives in `tagio`'s byte assertions, `TestWriteGoldenBytes` (`write_test.go:144`) and `TestParseGolden` (`container_test.go:359`). |

**Why.** A moved test whose expectation is an `ItemFields` value cannot
compile inside `tagio`, so "move the five" resolves either to moving
their subjects or to deleting coverage. Naming destinations keeps every
assertion's home explicit and lets the close record what moved, what was
already covered, and what stayed.

**Fold if ruled otherwise.** The two staying tests are rewritten inside
`tagio` against `TagSet` expectations rather than `ItemFields`
expectations, and `shadow` keeps no test of its own projection. The cost
is a projection with eighteen rows and no test, which is the coverage
this pass would be removing while claiming the behavior is unchanged.

### R3: fixtures come from the writer, and the one malformed fixture is a byte splice

**Applied.** Every well-formed fixture the `shadow` package builds after
this pass is produced by `tagio.Write` over a seed file carrying the
committed 42-byte oracle template's bytes, which already live in the
package as a byte literal (`flac_test.go:237-251`, the provenance comment plus the literal) and move to a
surviving test file. The fixture helper writes the seed, calls
`tagio.Write` with a `TagSet` and a nil picture set, and then pins the
modification time, in that order, because the writer's atomic rename
replaces the file.

The tightening test's malformed fixture is built by splicing a second
copy of the writer's own VORBIS_COMMENT block into writer output. No
general container builder returns: the helper emits only containers the
writer refuses to produce, and it has no well-formed users.

**Why.** Ruling 4's reason is that a hand-rolled container builder
inside a test is a second implementation of the thing under test
(`spec:273-279`). A splice over real writer output is not a second
implementation, because every byte in it came from the writer.

**Fold if ruled otherwise.** The malformed fixture is a hand-written
byte literal transcribed in full, at the cost of about forty bytes no
reviewer can read as a block chain.

### R4: the exclusion count is discharged as a bounded statement plus a read-only denominator

**Applied.** The exit gate asks the pass report to name how many
captured albums the tightened reader would have excluded (`spec:1168`).
The honest answer has three parts, and Task 3 produces all three.

1. **The exclusion set**, enumerated: every container shape the landed
   shadow reader accepted and `tagio` refuses, each naming the `tagio`
   test that grades it. This is a workstation measurement over committed
   code, and it needs no box.
2. **The denominator**, read-only, and stated as a proxy because the
   report prints no captured-album total. **Conductor-owed.**
   `shadow.Report` carries eighteen fields
   (`internal/engine/shadow/report.go:17-36`), and none of them is a
   count of captured albums, so the denominator is composed rather than
   read off a row. **Applied provisionally as `ObservedApplied` plus
   `ObservedSkipped` plus the `Unreconciled` counts**, read by the
   **conductor** through `dubplate-server shadow report` under the
   quiet-box check, never the bare binary as root
   (`docs/STATUS.md:14-16`). The conductor names in the close which rows
   it actually summed and what the sum does and does not count, and a
   better composition found while reading the report replaces this one
   rather than being forced into it. No implementer task ssh's.
   **Ruled, conductor 2026-09-12:** As applied by the fold editor at
   that site: `ObservedApplied` + `ObservedSkipped` + the `Unreconciled`
   counts, read by the conductor through `dubplate-server shadow
   report`. The site's own escape hatch stands: a better composition
   found while reading the report replaces it.
3. **The count itself**, stated with its bound: **it is not derivable
   from the stored records**, because a stored capture carries file
   names, sizes, modification times and projected `ItemFields`
   (`capture.go:143-162`) and carries no container bytes. Re-grading a
   captured album needs its source files re-read with the new reader,
   which is a box read of audio this pass does not take and a binary
   this pass does not deploy. The report therefore states zero known
   exclusions out of the denominator, names the exclusion set as the
   bound, and names the next deploy of the shadow units as the moment
   the tightening reaches the live capture.

**The count cannot be taken from workstation-held captures**, and the
plan says so plainly rather than leaving a reader to look: the shadow
runs on the box alone, no capture record is held on the workstation, and
`corpus/` holds JSON cases rather than audio.

**Why.** A number invented from stored records that cannot support it
would be worse than a bound stated with its method. The evidence a rung
15 reader actually needs is the exclusion set, because that is what
tells them which live shapes changed meaning.

**Fold if ruled otherwise.** The conductor takes a read-only sample of
library and inbox files off the box and grades them on the workstation
with the landed parser. The cost is copying whole files, since a
truncated prefix manufactures false refusals when a declared block
length runs past the copied extent.

### R5: the `io.ReaderAt` gap is declined at 8c and handed to rung 9

**Applied.** 8a's coverage gap 7 (`docs/HISTORY.md:1301-1309`) and the
8b handoff (`docs/HISTORY.md:869-871`) leave it to 8c whether `Parse`
grows a second, injectable entry point. **8c declines it.** No task in
this pass needs to inject a failing reader, the absorption adds no
consumer that does, and widening a signature to make one untestable
error path testable buys a coverage gap's closure at the cost of a
permanent second entry point at the pass that closes the rung.

**The gap stays declared** and moves to rung 9's call, named in Task 4's
close, because rung 9 is the first rung with a repair verb that branches
on a malformed file (`spec:1046-1051`).

**Fold if ruled otherwise.** Task 2 widens `Parse` to accept an
`io.ReaderAt` with the path-taking form as a thin wrapper, and the
mid-chain I/O error gets its own refusal test.

### R6: the invalid-UTF-8 value change is stated in advance

**Applied.** The landed shadow reader passes a comment's bytes through
unchanged (`flac.go:170-177`). `tagio` replaces each maximal ill-formed
subpart with U+FFFD on read, matching mutagen's `errors='replace'`
(`read.go:240-245`, ruling 10 at `spec:318-326`). A library file
carrying invalid UTF-8 in a comment therefore projects a different
string value after this pass than before it. **This is the one value
change the pass accepts**, it is graded by a row in the projection test,
and it is recorded as a deviation and a declared coverage gap at the
close. Any second value change stops the task and reports.

**Why.** The exit gate says the projection's behavior is unchanged. A
change found and stated is a deviation; a change found and smuggled is
the defect the gate exists to catch. The replacement is also the
correct behavior, because the shadow's scoring compares strings the
oracle produced through mutagen's own replacement.

**Fold if ruled otherwise.** The shadow's projection re-derives the raw
bytes, which `tagio` does not expose, so the fold is really a request to
widen `tagio`'s read surface and is refused for the same reason R5 is.

### R7: `readFLAC` keeps its name and its signature

**Applied.** The shadow's `readFLAC` (`flac.go:77`) keeps both, so
`run.go:209` and `capture.go:158` need no edit and the injected seam's
type is unchanged (A3). What changes is its body, its file and its doc
comment. The spec's 8c row says those two sites "repoint or are
rewritten" (`spec:1168`), and this resolution takes the first half by
making the repoint a no-op the close states rather than an edit nobody
needed.

**Why.** Renaming would touch two call sites, one struct field comment
and one injected test double for no behavioral gain, and the name still
describes what a caller gets: one FLAC file's matcher input.

**A note for the reviewer**, because the collision is real and reads
oddly in a diff: `tagio` has its own unexported `readFLAC`
(`read.go:31`), the registry's FLAC row for `Read`. The two are in
different packages and never meet, and `shadow`'s doc comment says which
one it is.

**Fold if ruled otherwise.** The function is renamed to name the
projection rather than the format, and Task 2 edits `run.go:209`,
`capture.go:158` and `reconcile.go:35` together.

---

## Task 0: The staleness pre-task

Model: **`opus`**, **fresh context, zero prior context**, dispatched by
the conductor before any implementer. Reads the rung 8 spec, passes 8a's
and 8b's HISTORY entries, and this plan against the tree as shipped.
**Its output is blockers with evidence, not fixes**; it writes no code
and edits no file. The conductor rules each blocker before Task 1
dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's pointers were resolved at `6a9570d`. A citation
that has moved since is a finding for this task rather than a surprise
for an implementer.

The checklist is the spec's own (`spec:1230-1277`), adapted to 8c.
**The first three are blockers: a "no" stops the launch.**

1. **Is the lane registry consistent with where this pass intends to
   run, is the primary tree free, and has no other lane written
   `internal/engine/shadow` or `internal/engine/tagio`?**
   `docs/STATUS.md:59-64` is the live-executor check, never `pgrep -f
   Projects/dubplate`, which matches every lane once a worktree exists.
   *(Expected at drafting: lane-1 main tree IDLE with pass 8c's plan as
   its next action; lane-2 IDLE, worktree `lane-2-5a` torn down, pass 5b
   owed.)* **Report any live executor in this tree, any uncommitted
   change under either package, and any file under either package that a
   lane 2 worktree has modified.** Lane 2's territory is
   `internal/{identity,auth,notify,serve}`, `web/` and
   `internal/config/keys.go`; a lane 2 edit inside 8c's territory is a
   blocker. **Report plainly whether a lane 2 worktree exists at all**,
   naming `git worktree list`'s output and `.claude/worktrees`'s contents,
   because this plan does not assume either answer and Task 1's gate
   baseline is read against it.
2. **Is `internal/engine/shadow/flac.go` unchanged, and are its three
   call sites still `run.go:209`, `capture.go:158` and
   `reconcile.go:35`?** *(Expected: 313 lines, last touched by
   `a86475c`, pass 4b's leftovers.)* Pass 4b edits comment lines under
   `internal/engine/shadow`, so this is a live question rather than a
   formality (`spec:1255-1259`). Report the file's current length and
   each call site's current line.
3. **Outside `flac_test.go`, is `buildFLAC` still used by
   `capture_test.go:34`, `capture_test.go:173` and `run_test.go:107`,
   and only those?** Its four uses inside its own file
   (`flac_test.go:85`, `:202`, `:226`, `:288`) are expected and are not
   a finding (`spec:1260-1264`). **Count them yourself and report the
   list**, because Task 1's whole scope is that list.
4. **Is `internal/engine/tagio`'s exported surface still what 8b's close
   recorded?** *(Expected: `Parse`, `Read`, `Write`, `TagSet`,
   `Mapping()`, `WriteOrder()`, `Serialize`, `EmptyString`, `EmptyList`,
   `Container`, `Block`, `StreamInfo`, `Comment`, `VorbisComment`,
   `KeyWrite`, `Field` and its 69 constants, `FieldMapping`, `Style`,
   `Picture`, `FrontCover`, `Format`, `SupportedFormats`, `Sniff`,
   `ErrUnsupportedContainer`, `ErrMalformedContainer`, `Deviations`;
   `docs/HISTORY.md:663-672`.)* **Report `Parse`'s, `Read`'s and
   `Write`'s exact current signatures and `Container`'s exact current
   field list**, because R1 and Tasks 1 and 2 all build on them.
5. **Is `TestConformanceLedgerSubjects` still at `conform_test.go:640`,
   and what does `go test ./internal/engine/tagio/... -run Conformance
   -count=1` report?** Run it once and report **the test count and the
   wall time**, not the exit status alone, because a `-run` pattern
   matching nothing prints "no tests to run" and exits 0. *(8b measured
   twelve subtests in about 0.05s; `docs/HISTORY.md:954-957`.)* A count
   of zero is a blocker for Task 2.
6. **What does `go test ./internal/engine/shadow/... -count=1` report
   now?** Report the test count and the wall time. That is the baseline
   the absorption must preserve, and Tasks 2 and 3 are graded against
   it.
7. **Is `internal/harness/oracle/testdata/template.flac` still the same
   42 bytes, and does `flac_test.go:244-251`'s literal still equal it
   byte for byte?** R3's seed depends on both. Compare mechanically and
   report the sizes; **do not run any derivation and do not touch the
   venv**.
8. **Is `corpus/analysis/file_tags.json` unchanged?** *(Expected:
   `shim_version` `1`, `beets_version` `2.13.1`, 69 `writable_fields`,
   9 `readonly_fields`, twelve rows.)* A changed ledger is a changed
   conformance target, and item 5 would then be green for the wrong
   reason. **Count mechanically; read no row body into context.**
9. **Do `TestEngineNeverImportsHarness` and `TestEngineNeverReadsTheClock`
   still walk `internal/engine` including test files outside
   `testdata`?** *(Expected: `internal/harness/conform/arch_test.go:141`
   and `:152`.)* Confirm that no arch rule forbids `internal/engine/shadow`
   importing `internal/engine/tagio`, which is the one new import this
   pass adds.
10. **Does `internal/engine/tag/doc.go` still describe `tagio` as
    knowing nothing about matching?** *(Expected: `:8-11`.)* Ruling 3
    preserves that sentence, and the absorption is the pass most able to
    break it.
11. **Has `scripts/check.sh` gained or lost a leg, and has
    `internal/config` changed shape?** 8c registers no key either way,
    so this is a confirmation. Report the leg list as it stands, confirm
    `internal/config/keys.go` carries no rung 8 class, and report
    whether lane 2 is holding `keys.go`.
12. **Is `.gitleaksignore` still tracked and empty?** *(Expected: zero
    bytes.)* This pass adds no tracked data file and no fingerprint, so
    a non-empty file is a note rather than a blocker. **Do not quote any
    fixture literal**, because gitleaks scans the tracked worktree and
    would flag the quotation.
13. **Confirm 8b's own closing commits landed and touched nothing under
    `internal/engine/shadow`:** the batched `code-simplifier` round
    `ce0ac2e` and the leftovers round `a70b346`
    (`docs/HISTORY.md:640-672`). A `shadow` file in either commit's stat
    is a blocker for Task 2, because every `file:line` in Tasks 1 and 2
    was resolved before them.
14. **Does every `spec:` and `docs/HISTORY.md:` pointer this plan carries
    land on its subject?** **Resolve each distinct pointer by grepping
    for the subject rather than by reading the line number**, and report
    every one that does not. Stale pointers are this build's named
    dominant defect class, so this item is a report and never a fix.
15. **Is the box quiet, on the record rather than by ssh?** Read
    `docs/STATUS.md:14-20` and confirm what is deployed and that no task
    in this plan deploys. **Do not ssh.** The one box reading in this
    pass is the conductor's, in Task 3.

**Deliverable count: 1** (the blocker report, each item with its
evidence and each blocker with the file and line that proves it). No
gate run beyond items 5 and 6's timing measurements; the task writes
nothing.

**2.0 hooks: neutral.** The pre-task builds no surface.

---

## The `code-simplifier` round

**One batched round, dispatched by the conductor after Task 3 is
accepted and before Task 4's close.** Rung 8's chain commits per task,
and CLAUDE.md forbids instructing an implementer to dispatch an agent,
so a conductor-owned simplifier dispatch cannot reach a point before a
chain task's commit. The two rules cannot both be honored inside the
chain, and 8a and 8b both resolved it the same way: one round over the
pass's Go rather than one per task. This pass states the departure
rather than inheriting it silently. The conductor applies the
refinements, runs `bash scripts/check.sh`, and commits them on their own
so Task 4's ledger draft describes a settled tree.

**The round's own subject is unusual here and worth naming:** most of
this pass's Go is deleted rather than written, so the round grades the
surviving projection and the new fixture helpers. A round that reports
no change is a legitimate outcome, and the close records the actual.

**Budget line:** 0.15M, against 8b's measured 0.13M.

---

## Task 1: The external fixtures move to the writer

Model: `sonnet`. Depends on Task 0. **Not independent**: it must land
before Task 2, because `buildFLAC` cannot be deleted while a caller
outside its own file remains.

**Files:**
- Create: `internal/engine/shadow/fixture_test.go` (the seed literal and
  the writer-built fixture helper)
- Modify: `internal/engine/shadow/capture_test.go` (`writeAlbum` at
  `:25`, `windroseTracks` at `:60`, the inline tag lists at `:92`,
  `:119`, `:196`, and the nested track at `:173`)
- Modify: `internal/engine/shadow/run_test.go` (`albumTracks` at `:85`,
  `writeLibraryFile` at `:103`)
- Modify: `internal/engine/shadow/observe_test.go` (the inline tag list
  at `:111`)

**Interfaces:**
- Consumes: `tagio.Write` (`write.go:39`), `tagio.TagSet`
  (`tagset.go:128`) and its four states, and `flac_test.go:244-251`'s
  committed template literal.
- Produces: the fixture helper every `shadow` test uses from here on,
  and a `buildFLAC` whose only remaining callers are inside
  `flac_test.go`. Task 2 deletes both.

**Outcomes:**

- The seed literal moves out of `flac_test.go` into a file that survives
  Task 2, with its provenance comment intact: it is
  `internal/harness/oracle/testdata/template.flac`'s 42 bytes, embedded
  rather than read, because that path is harness territory the engine
  does not read.
- One fixture helper writes the seed to a path, calls `tagio.Write` with
  a `TagSet` and a nil picture set and the caller's mode, and pins the
  modification time **after** the write, because the writer's atomic
  rename replaces the file and a mtime pinned first would not survive.
- Every fixture-building helper takes a `tagio.TagSet` rather than a
  `"KEY=value"` string list. No test parses a tag string, because a
  parser inside a test is the second implementation ruling 4 is about.
- The four-state vocabulary is used where the landed fixtures meant it:
  `albumTracks("", ...)` (`run_test.go:85-95`) writes an empty
  MUSICBRAINZ_ALBUMID today, and the fixture keeps whichever state
  reproduces the album-with-no-release-id shape the observe rows depend
  on.
- `buildFLAC` (`flac_test.go:22`) still exists and still compiles, with
  every caller inside its own file. **Task 1 deletes nothing.**
- No behavior in `internal/engine/shadow`'s non-test files changes. This
  task edits tests alone.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- `grep -rn buildFLAC internal/engine/shadow` names `flac_test.go` and
  no other file, and the task reports the grep's output.
- Every test in the package passes unchanged in intent: the task reports
  `go test ./internal/engine/shadow/... -count=1`'s test count and wall
  time against Task 0 item 6's baseline, and names any test whose
  assertions it had to touch and why.
- The capture digest tests still pass, which is what proves the
  modification time is pinned after the writer's rename rather than
  before it. The task names the tests that would have failed.
- A fixture track read back through the landed `readFLAC` projects the
  `ItemFields` its test expects, so the switch of fixture producer
  changed no expectation. Any expectation that did change is reported
  with its cause rather than adjusted quietly.
- The multi-album exclusion rows still produce two groups, which is what
  proves the fixtures still carry distinct albumartist and album values
  through the writer's own serialization.
- `writeLibraryFile`'s fixture and its capture-side counterpart still
  join on title and length under `reconcile.go:242`'s tolerance. **The
  landed fixtures declare 44100 samples at 44100 Hz and R3's seed declares
  zero, and `tagio.Write` preserves the incoming STREAMINFO payload
  (`write.go:151-152`), so both sides of that join read `Length` 0 and the
  length term compares zero to zero.** The task takes one of two routes
  and reports which: it carries a **second seed variant** whose STREAMINFO
  pack declares a non-zero total-sample count, taken as a byte edit of the
  committed literal rather than as a container builder, so the join keeps
  a live length term and A2's `Length` row gains a distinguishable value;
  or it leaves every fixture on the zero-sample seed and reports the
  vacuous length term as a finding, which Task 4 then carries in the
  declared-gap list. Silently letting the term go vacuous is the outcome
  this criterion exists to prevent, because `reconcile_test.go` grades the
  join over fabricated `ItemFields` and would stay green either way.
- No test file reads `internal/engine/tagio/testdata` or any path under
  `internal/harness`. The seed is the embedded literal.
- `go.mod` unchanged. `bash scripts/check.sh` green. **This task's gate
  run is the pass's timing baseline**: the task reports the total wall
  time and the per-leg breakdown, and names whether a lane 2 worktree
  exists per Task 0 item 1, so Tasks 2 and 3 compare against a measurement
  taken in this tree rather than against 8b's figure.

**Named mutations this task reports:**

1. Write the seed and skip the `tagio.Write` call → every capture test
   asserting an album's items fails, because the tracks carry no tags.
   That is the proof the writer is actually producing the fixture.
2. Pin the modification time before the write rather than after → the
   digest-stability assertion fails, because the atomic rename installs
   a fresh mtime. The task names the failing test.
3. Give the two `windroseTracks` entries the same title → no test fails,
   and the task reports it unfired with the reason: the capture groups
   by albumartist and album, so a title collision is invisible at this
   altitude and is graded downstream by the matcher's own tests.
4. Pass the fixture mode as 0 rather than 0o644 → the capture discards
   the album, because the file cannot be opened, which is the same path
   a partial read takes.
5. Write the fixture with a non-nil empty picture set rather than nil →
   no test fails, and the task reports it unfired with the reason: the
   seed carries no PICTURE block, so a strip and a leave-alone produce
   identical bytes (`docs/HISTORY.md:571-576`).

**2.0 hooks: neutral.** The fixture switch builds no surface.

**Prior art:** adopts the repo's own writer as the fixture source,
which is ruling 4's own instruction (`spec:273-279`) and the reason the
spec's prior-art table calls the in-repo reader the rung's nucleus
(`spec:1090`). Declines a vendored FLAC fixture file, on 8b's R4
precedent that a committed binary is bytes no reviewer can read from a
diff. Declines `go-flac/go-flac` v2 as a fixture builder on the same
grounds the spec declines it as a dependency (`spec:1092`).

**Deliverable count: 3** (the seed literal moved with its provenance and,
at the task's option, its non-zero-sample variant, the writer-built
fixture helper, the call sites converted). The variant is part of the seed
deliverable rather than a fourth, because it is the same committed bytes
with one field edited. Gate green; commit.

---

## Task 2: The absorption

Model: `sonnet`, with a conditional `opus` upshift named in the
execution model. Depends on Tasks 0 and 1. **Not independent**: it
deletes what Task 1 emptied and everything after it reads the result.

**Files:**
- Delete: `internal/engine/shadow/flac.go`,
  `internal/engine/shadow/flac_test.go`
- Create: `internal/engine/shadow/libraryread.go` +
  `libraryread_test.go` (the projection and its tests),
  `internal/engine/shadow/itembag_test.go` (the round trip that stays)
- Modify: `internal/engine/shadow/reconcile.go:35` (the doc comment),
  `internal/engine/shadow/itembag.go:11-13` (the pointer at
  `flac_test.go`)
- Modify, under R1 only: `internal/engine/tagio/read.go` (export the
  projection) and `internal/engine/tagio/doc.go` (say what the new
  function is for)
- Modify, if R2's walk finds an uncovered row:
  `internal/engine/tagio/read_test.go` or
  `internal/engine/tagio/container_test.go`
- Modify, **pre-authorized rather than conditional**:
  `internal/engine/tagio/container_test.go`, to add rows to
  `TestParseRefusals` for the three refusals the parser performs and no
  test grades. Verified at drafting by grepping the package's test files
  for each message: "STREAMINFO is not the first block"
  (`container.go:198`, `:209`), "no STREAMINFO block" (`container.go:245`),
  and "carries no '='" (`container.go:332`). Task 3's enumeration rows 2
  and 6 need two of the three, and Task 3 may edit only `shadow` files, so
  this task closes them. Adding a test for a landed refusal adds no
  refusal.

**Interfaces:**
- Consumes: `tagio.Parse` (`container.go:135`), `Container.StreamInfo`
  (`container.go:104-118`), `tagio.TagSet` (`tagset.go:128`), and under
  R1 the newly exported `tagio.Tags`; `matcher.ItemFields` and
  `matcher.ProjectItemFields`.
- Produces: `shadow.readFLAC` with its name and signature unchanged
  (A3, R7), and a `shadow` package that holds no container knowledge
  (A1).

**Outcomes:**

- A1: `internal/engine/shadow`'s non-test files hold no FLAC signature
  constant, no block-type constant, no block walk, no `encoding/binary`
  import, no Vorbis decode, no `safe_cast` reimplementation, and no date
  splitting. Task 3's fixture helper is the one place in the package that
  touches container bytes, and A1's scoping sentence says why that is not
  a hole.
- A2: the projection to `matcher.ItemFields` keeps its **eighteen** rows,
  its per-row mediafile citations, and its typed-default rule. `DataSource`
  is the nineteenth `ItemFields` member and has no row, for the reason
  `flac.go:203-207` records, and the surviving doc comment carries that
  reason forward. Its input is now `tagio`'s own types rather than a
  private map.
- A3 and R7: `readFLAC` keeps its name and signature, so `run.go:209`
  and `capture.go:158` need no edit. The task states that as a finding
  rather than leaving two of the spec's three named call sites looking
  untouched by accident.
- A4: the length division and its zero-rate guard survive, reading
  `Container.StreamInfo`'s sample rate and total sample count.
- A5 and R1: one `tagio.Parse` per file, with the tag set and the length
  read off the same `Container`. Under R1's fold, two parses, stated.
- A6: `flac.go:62-70`'s provisional paragraph does not survive. The
  surviving doc comment says what the projection does, names ruling 3 as
  the reason the projection lives in `shadow`, and names `tagio` as
  where the container knowledge went.
- R2's destination table, discharged test by test, with the walk of
  `TestReadFLACStylePrecedence`'s **sixteen** rows against `tagio`'s own
  projection tests reported row by row.
- The three ungraded parser refusals gain rows in `TestParseRefusals`, so
  Task 3's enumeration rows 2 and 6 have a `tagio` test to name. The task
  reports each row it added and each message it could not reproduce.
- `reconcile.go:35`'s doc comment repointed, and `itembag.go:11-13`'s
  pointer at `flac_test.go` repointed to the round trip's new home.
- R6's invalid-UTF-8 value change, carried as a test row rather than a
  surprise: a comment carrying an ill-formed sequence projects U+FFFD
  where the landed reader projected the raw bytes.
- `buildFLAC`, `blockHeader`, `writeTemp` and
  `TestBuildFLACEmitsTheStreaminfoLayoutByteForByte` are all deleted
  with `flac_test.go`, and anything Task 1's helper still needs from
  them lives in `fixture_test.go` instead.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- **The `shadow` package's full suite is green, with the projection's
  behavior unchanged.** The task reports `go test
  ./internal/engine/shadow/... -count=1`'s test count and wall time
  against Task 0 item 6's baseline, names every test whose body changed,
  and names every assertion whose expected value changed. **The expected
  count of changed expected values is one**, R6's, and any second one
  stops the task and is reported to the conductor rather than absorbed.
- **`TestConformanceLedgerSubjects` stays green** (A9). The task runs
  `go test ./internal/engine/tagio/... -run Conformance -count=1` and
  reports the test count and the wall time, because a pattern matching
  nothing exits 0.
- **The eighteen-row projection has a test that fails when any one row is
  swapped**, on `TestItemBagRoundTrips...`'s own precedent
  (`flac_test.go:333-337`'s rationale). The grading splits three ways
  because two rows cannot carry a distinguishable value in a writer-built
  fixture. **Seventeen rows carry a distinguishable value in one
  writer-built fixture.** `Length` is graded by A4's `StreamInfo` table,
  and additionally by a distinguishable value in the fixture if Task 1
  landed its non-zero-sample seed variant. **`DataSource` is asserted as
  the empty string**, with the test's own comment naming the reason it has
  no row (`flac.go:203-207`). A criterion demanding a distinguishable value
  for all eighteen is unsatisfiable, which is why it is not the criterion.
- The length table covers ten seconds at 44100 Hz, a fractional length,
  and a zero sample rate, each over a `tagio.StreamInfo` value rather
  than a file, and each expected value hand-transcribed from
  `flac_test.go:213-235`'s landed rows.
- `grep -rn "encoding/binary\|fLaC\|VORBIS" internal/engine/shadow`
  returns nothing outside a comment, and the task reports the grep. Task
  3's splice helper will later add a comment naming VORBIS_COMMENT, which
  this criterion permits by design; what it forbids is an import, a
  signature constant or a block-type literal in code.
- `internal/engine/shadow` imports `internal/engine/tagio` and nothing
  else new. `internal/engine/tagio` imports nothing new at all, and
  `matcher.ItemFields` appears in no `tagio` file, asserted by the
  reviewer's read and by the arch tests staying green.
- Under R1, the exported projection has a doc comment naming its caller
  and its relationship to `Read`, and `Read`'s own behavior is
  byte-identical to what it was: the task reports that `read.go`'s
  registry row still calls the same projection and that every `tagio`
  read test passes unchanged.
- R2's table is discharged in the task's return, test by test, naming
  for each one where it went and, for a deleted one, which `tagio` test
  now carries its subject.
- `reconcile.go:35` and `itembag.go:11-13` both name a file that exists.
  A pointer to a deleted file is a defect this criterion exists to
  catch.
- `go.mod` unchanged. `bash scripts/check.sh` green, with vale's result
  on every new or rewritten doc comment named explicitly.

**Named mutations this task reports:**

1. Read the tag set and skip the STREAMINFO read → every length is zero,
   and the length table's ten-second row fails.
2. Drop the zero-rate guard → the zero-rate row fails with a division
   result rather than zero. That is the guard `flac.go:41-46` documents
   and the one the absorption most easily loses.
3. Swap two projection rows, `Label` for `Country` → the eighteen-row
   projection test fails on both rows, naming each.
4. Keep the deleted `parseFLAC` and call it instead of `tagio` → the
   package still compiles and every landed test passes, and **Task 3's
   tightening test is what fails**. The task reports that dependency
   explicitly, because it is the reason Task 3 exists as a separate
   task rather than as a criterion here.
5. Project an absent `*string` member as the literal string "nil", or
   dereference it unguarded → the absent-key row fails or panics.
6. Return the first matching comment's value versus the last for a
   repeated key → no test fails **in `shadow`**, and the task reports it
   unfired with the reason: the rule is now `tagio`'s and is graded at
   `read_test.go:277`.
7. Leave `flac_test.go` in place while deleting `flac.go` → the package
   does not compile, which is the mechanical proof that the reader's
   tests and the reader die together.

**2.0 hooks: neutral.** This task lands no hook. Under R1 it improves
the shape hook 3's consumer will use, because rung 9's admit path can
read a file's tags and its embedded picture from one parse
(`S2:363-364`), which retires 8b's declared coverage gap 4. That is a
consequence to record at the close rather than a hook this pass lands.

**Prior art:** adopts the repo's own FLAC reader as the thing being
absorbed, which is the spec's prior-art verdict for it
(`spec:1090`: adopt and extend, the rung's nucleus). Declines
`dhowden/tag` for the read path on its security bar (`spec:1094`).
Declines `mewkiz/flac`, which the spec adopts for rung 7's integrity
check alone and declines for tags (`spec:1091`). Reproduces mediafile's
and mutagen's behavior only, already landed in `tagio` at 8a, with no
code carried and no new library consulted.

**Deliverable count: 5**, which is above CLAUDE.md's roughly-four line
and is stated plainly rather than hidden: the projection over `tagio`,
the deletion of the reader, the deletion of the fixture builder and its
layout test, the four repoint sites, and R2's seven destinations. They
are one change in substance, and splitting them further leaves the tree
red at a task boundary, because the reader, its tests and its fixture
builder compile as one unit. **Revision 2 added the three
`TestParseRefusals` rows to this task, and the count stays 5**, because
those rows sit inside R2's own walk clause, which already sent uncovered
rows here. Naming the addition rather than absorbing it is CLAUDE.md's
pass-sizing rule, and the addition is three table rows. Gate green;
commit.

---

## Task 3: The tightening, asserted and enumerated

Model: `sonnet`. Depends on Tasks 0, 1 and 2.

This is the pass's one behavioral change and the reason 8c is a pass
rather than a chore (`spec:1173-1175`). It closes risk 6
(`spec:1320-1330`) by asserting the tightening rather than discovering
it in a divergence.

**Files:**
- Modify: `internal/engine/shadow/capture_test.go` (the tightening test
  and its spliced fixture)
- Modify: `internal/engine/shadow/fixture_test.go` (the splice helper)

**Interfaces:**
- Consumes: Task 1's writer-built fixture helper, Task 2's absorbed read
  path, `capture`'s directory walk (`capture.go:78`) and
  `captureAlbum`'s partial-read discard rule (`capture.go:131`).
- Consumes, for the splice alone: `tagio.Parse`, `Container.Blocks`'s
  `Length` field, and `Container.VorbisComment.Vendor`.
- Produces: the enumeration Task 4's report names, as a table in the
  task's return.

**Outcomes:**

- A capture-path test proving that an album holding a container the
  landed reader accepted and `tagio` refuses is **discarded whole**:
  no capture stored, no exclusion row written, which is the
  partial-read rule `captureAlbum` already implements at its three
  returns (`capture.go:133-135`, `:145-147`, `:159-161`) and
  `capture_test.go:194-215` already asserts for a truncated file.
- The witness fixture is a writer-built container with its own
  VORBIS_COMMENT block spliced in twice (R3), which is ruling 9's
  "more than one VORBIS_COMMENT block" refusal (`spec:310-317`) and a
  shape the landed reader read happily, taking the first value per key
  (`flac.go:117-121`, `:175-177`).
- **The splice's method is named here rather than left to the
  implementer**, because A1 forbids a block-type constant, a header
  encoder and an `encoding/binary` import inside `shadow`, and `tagio`'s
  own block-type constants are unexported (`container.go:22-28`) so none
  can be borrowed. The helper parses the writer's output, identifies the
  VORBIS_COMMENT block as the one whose payload opens with the vendor
  string `Container.VorbisComment.Vendor` reports (the payload's first
  field is that string behind a length prefix, `write.go:217-231`),
  computes the block's byte range as four plus the running sum of each
  preceding block's four plus `Length`, and **copies that range whole with
  its four-byte header included**. Copying the header rather than emitting
  one is what keeps an encoder out of `shadow`, and the copied header's
  cleared last-block flag is correct as it stands, because the writer emits
  exactly one 8 KiB PADDING block last (`spec:1166`) so VORBIS_COMMENT is
  never the last block. The helper's doc comment states the method and
  states that it emits only containers the writer refuses to produce.
- **The enumeration**, produced as a table: every container shape the
  landed reader accepted and `tagio` refuses, each with the `tagio`
  test that grades it. The expected rows, to be corrected against what
  execution finds:
  1. More than one VORBIS_COMMENT block (`TestParseRefusals`).
  2. A STREAMINFO block that is not first, or no STREAMINFO block at
     all. The landed reader read neither as an error and projected a
     zero length. **Expected finding rather than a surprise**: neither
     message had a `TestParseRefusals` row at drafting
     (`container.go:198`, `:209`, `:245`), and Task 2 is pre-authorized to
     add them, so this row names the row Task 2 added.
  3. A block appearing after a header whose last-block flag is set. The
     landed reader stopped at the flag and returned what it had.
  4. More than one SEEKTABLE block. The landed reader skipped it by
     declared length.
  5. A comment field name carrying a code point outside U+0020 through
     U+007E. The landed reader upper-cased it and kept it.
  6. A comment carrying no `=`. The landed reader skipped that comment
     and read the rest. **Expected finding rather than a surprise**: this
     message had no `TestParseRefusals` row at drafting either
     (`container.go:332`), and Task 2 is pre-authorized to add it.
  7. A PICTURE block whose declared lengths do not fit its payload,
     which 8b's decode added (`docs/HISTORY.md:559-566`). The landed
     reader skipped every PICTURE block by declared length.
- Each row names whether a shape is reachable from a real inbox upload
  or only from a crafted file, so a reader of the close knows which
  rows carry live consequence.
- The task states, as a finding, that **no fixture in the landed shadow
  suite carried any of these shapes**, because every fixture was
  `buildFLAC`'s well-formed output (`spec:1020-1022`). That is why the
  tightening was invisible before this test.

**Acceptance criteria a `diff-reviewer` verifies from the diff:**

- The tightening test asserts zero captures **and** zero exclusion rows,
  because a partial read is discarded rather than excluded, and it names
  that rule in its own comment.
- The witness fixture's bytes come from the writer plus a splice, and
  the reviewer checks that no general container builder returned. The
  splice helper's own doc comment says it emits only containers the
  writer refuses to produce.
- **The splice helper carries no block-type literal, no header encoder and
  no `encoding/binary` import**, which the reviewer checks against A1's
  scoping sentence and against the method the outcomes name. A `grep -rn
  "encoding/binary" internal/engine/shadow` still returns nothing, and the
  task reports it.
- A second row proves the fixture is a witness rather than a broken
  file: the same album without the splice is captured normally. Without
  that row the test passes for any fixture the reader dislikes.
- The enumeration's every row names a `tagio` test by name, and the
  task reports any row it could not tie to one. A row with no test is a
  finding for the conductor, not a row to drop. Rows 2 and 6 name the
  `TestParseRefusals` rows Task 2 added, and a missing one is a Task 2
  regression this criterion surfaces rather than a Task 3 gap.
- Every existing capture test passes unchanged, and the task reports the
  package's test count and wall time.
- `bash scripts/check.sh` green, with the wall time compared against Task
  1's baseline measurement rather than against 8b's figure.

**The conductor's read-only evidence step, which is not this task's.**
R4's denominator is read by the conductor, after this task is accepted:
one quiet-box check, then `dubplate-server shadow report` read-only,
never the bare binary as root (`docs/STATUS.md:14-16`). **The report
prints no captured-album total**, so the conductor composes the
denominator from the rows R4 item 2 names and records which rows it summed
(conductor-owed). **No implementer task ssh's, and no task writes to the
box.** If the box cannot be read quietly, the conductor records the
denominator as unread with the reason, and the exit gate is still
discharged, because R4's answer to "how many" does not depend on it.

**Named mutations this task reports:**

1. Build the witness fixture without the splice → the refusal row fails,
   because the album is captured. That is the proof the fixture carries
   the shape.
2. Catch the read error in `captureAlbum` and continue → the refusal row
   fails, reporting a stored capture, which is the tolerance the landed
   reader had and this pass removes.
3. Assert the album is excluded with a reason rather than discarded →
   the assertion fails against `capture.go`'s own rule, which is the
   distinction `capture_test.go:210-215` already draws.
4. Enumerate the tightening from the deleted reader's behavior rather
   than from `tagio`'s refusal set → a row appears with no `tagio` test
   to name, which the criterion above catches.

**2.0 hooks: neutral.** The tightening is a 1.0 correctness change on
the shadow's evidence path.

**Prior art:** adopts the gate-rejected exclusion's own precedent for
counting an exclusion in the pass report rather than discovering it in a
divergence (`spec:1022-1026`). Declines any new library; the assertion
is over the repo's own capture path.

**Deliverable count: 2** (the capture-path tightening test with its
witness fixture, the enumeration). Gate green; commit.

---

## Task 4: The pass close, the rung close, and the spec corrections

Model: `sonnet`. Depends on Tasks 0 through 3 and on the batched
`code-simplifier` round.

**Files:**
- Draft, in the main tree:
  `docs/research/handoffs/draft-STATUS-8c.md` and
  `docs/research/handoffs/draft-HISTORY-8c.md`, the second
  carrying this pass's `ROADMAP.md` text (the ledger text as drafts;
  **the implementer never edits or commits a ledger**)
- Draft, handed to the conductor in the task return: the spec
  corrections, the deviation summary, and the rung 8 close statement

**Ruled, conductor 2026-09-12 (M11,
`plans/2026-09-12-overnight-window.md`):** The ledger text is written as
drafts at `docs/research/handoffs/draft-STATUS-8c.md` and
`docs/research/handoffs/draft-HISTORY-8c.md` in the main
tree, the second carrying this pass's `ROADMAP.md` text, never to
`docs/STATUS.md`, `docs/HISTORY.md` or `ROADMAP.md` themselves. The
conductor folds the drafts into the real ledgers at the merge and
commits them path-limited. The drafts are committed artifacts, so they
survive a restart and a fresh session can read them.

**Outcomes:**

- The HISTORY entry in the house shape: what landed, what the gate
  caught, and what a later pass would be wrong to rediscover.
- **The A1 through A9 coverage table**, each item with its Go site, what
  proves it, and, for a rule no test can reach, the stated reason.
- **The seven Resolutions listed with their markers**, each retired from
  provisional to ruled or folded, each naming the file and symbol the
  marker sits on, and each naming what would have changed had it been
  ruled the other way. **R1, R2 and R4 item 2 are the three conductor-owed
  folds**, and each additionally names the conductor's own ruling and its
  date. R1's entry records the ruling as a departure from R12's assignment
  to rung 9, and it names which half of the 8b handoff the ruling took.
- **R2's destination table as executed**, so rung 15 and rung 9 read
  where each of the seven tests went rather than searching for them.
- **Task 3's enumeration**, verbatim, as the tightening record the rung
  hands rung 15.
- **R4's three-part answer to the exit gate's count**, with the
  denominator if the conductor read it and the reason if not.
- **The declared coverage gaps recorded as gaps, not as passing checks**,
  on one criterion: a surface 8c itself builds or inherits that no test
  in this pass can reach. The expected list, to be corrected against
  what execution found:
  1. **The `io.ReaderAt` gap stays open** (R5), inherited from 8a's gap
     7 and now handed to rung 9 rather than to a rung 8 pass.
  2. **The tightening's live effect is unmeasured** (R4): the count is
     bounded rather than counted, because stored captures carry no
     container bytes.
  3. **No real library file has been read through the absorbed path.**
     Every fixture is writer-built or a committed literal; a real
     beets-written library file is first read at rung 9. Inherited from
     8a's gap 6 rather than newly declared.
  4. **The invalid-UTF-8 value change is graded on a synthetic
     fixture** (R6), not on a file any member uploaded.
  5. **The end-to-end length join's length term is vacuous**, if Task 1
     left every fixture on the zero-sample seed. `reconcile.go:242` joins
     on title and length, R3's seed declares zero total samples, and
     `tagio.Write` preserves STREAMINFO (`write.go:151-152`), so both
     sides read zero and only the title term discriminates.
     `reconcile_test.go` grades the join over fabricated `ItemFields` and
     stays green regardless, which is why the loss needs declaring rather
     than detecting. **This gap does not exist if Task 1 landed its
     non-zero-sample seed variant**, and the close records which happened.
  6. Anything execution adds.
  **The count is stated and joined to the running total of sixty-three**
  (`docs/STATUS.md:80-82`), so rung 15 does not re-count. A gap that
  restates an already-counted 8a or 8b gap is named as a restatement
  rather than counted again, on 8b's own precedent
  (`docs/HISTORY.md:812-819`).
- **8b's declared coverage gap 4 is retired if R1 stood**, because a
  caller wanting tags and pictures now parses once. The close says so
  explicitly, because a retired gap that nobody records stays on rung
  15's list forever.
- **The spec corrections the pass owes**, drafted for the conductor:
  - **The five-tests sentence** (`spec:995-998`) is corrected to R2's
    destination table, naming why two of the five stay in `shadow`.
  - **The reference-count sentence** (`spec:978-985`) is corrected to
    name what actually changed. That paragraph already names six
    references: three call sites (`run.go:209`, `capture.go:158`,
    `reconcile.go:35`), two prose mentions (`capture_test.go:47`, `:203`),
    and the reader's own doc comment (`flac.go:62-70`). Under R7 two of the
    three call sites need no edit and both prose mentions survive
    unedited, because `readFLAC` keeps its name. **`itembag.go:11-13` is a
    seventh reference and the one the spec never named**, and it is the
    reference this pass actually edits.
  - **The ceiling table's 8c line** (`spec:1214`), corrected from
    1M to this plan's 2.1M with the derivation, and rung 8's totals
    restated: ceilings 9.85M against the 8M reservation, actuals about
    8.42M expected.
  - **The read-path strictness section** (`spec:1012-1033`), corrected
    against Task 3's enumeration wherever the spec's four-shape list is
    narrower than what the landed refusal set actually refuses.
  - CLAUDE.md's gate sentence names the tracked Python files by role;
    this pass adds none, so the sentence stands. State that it was
    checked.
- **The rung 8 close statement**, because 8c is the rung's last pass:
  the four owner rulings and the fourteen conductor rulings each named
  as discharged or carried, hooks 13, 14 and 3 recorded with the pass
  that landed each, the deviation note handed to rung 15 as a statement
  rather than a file to interpret, fork B recorded as ruled, and the one
  carry that leaves the rung open (the group non-FLAC question).
- **What rung 9 inherits**, in one section it can read without reading
  the whole entry: `tagio`'s exported surface as it now stands, the
  `io.ReaderAt` call now rung 9's, the combined-read question R1 either
  answered or left, the tightening's effect on the intake gate rung 7
  bounds, and the fact that the shadow's live behavior changes at the
  next deploy rather than at this merge.
- The evidence, not a bare "green": the `shadow` package's test count
  and wall time before and after; the conformance run's test count and
  wall time; `go test ./... -count=1` and `bash scripts/check.sh` wall
  times; the mutation table across Tasks 1 through 3 with every outcome,
  including the deliberately unfired ones; and the line counts deleted
  and added, because a pass whose main deliverable is a deletion should
  say how much it deleted.
- Both budgets scored: tokens against the 2.1M ceiling, and attended
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
  edit, with the pass ledger moved to HISTORY and the lane registry left
  pointing at the next action.
- `ROADMAP.md`'s rung 8 entry moves to closed, and the next rung's entry
  is left as the initiative queue's next item.
- The rung close names every ruling once, and a ruling that was not
  exercised by any pass is named as unexercised rather than as
  discharged.
- No ledger is committed by the implementer.

**2.0 hook: neutral.** The close records hook 3 as 8b's, hooks 13 and 14
as 8a's and the spec's, and 8c as landing none.

**Deliverable count: 4** (the HISTORY entry, the STATUS and ROADMAP
edits, the drafted spec corrections, the rung 8 close statement with the
rung 9 handoff).

---

## Exit-gate coverage

The spec's 8c exit gate (`spec:1168`) has three assertions, and each is
an acceptance criterion in the task that lands it. This table exists so
a reviewer can check the mapping without reading every task.

| Exit-gate assertion | Task that lands it | The criterion |
| --- | --- | --- |
| `check.sh` green | every task | The last criterion in Tasks 1, 2 and 3, each reporting the gate's wall time. |
| The `shadow` package's full suite green, with the projection's behavior unchanged | Task 2 | The first criterion, graded against Task 0 item 6's baseline, with exactly one expected value changed (R6) and any second one stopping the task. |
| The pass report naming how many captured albums the tightened reader would have excluded | Task 3 produces it, Task 4 reports it | Task 3's enumeration criterion plus R4's three-part answer, with the conductor's read-only denominator. |

The 8c row's scope statement maps the same way: the reader moves and the
projection stays (Task 2), the four repoint sites (Task 2), the three
external `buildFLAC` uses (Task 1), `buildFLAC` and its layout test
deleted (Task 2), the five reader tests and the itembag test (Task 2,
under R2), and the capture-path test (Task 3).

## Deviations this plan takes

Five, each stated here rather than discovered by a reviewer. Revision 2
added the fifth, which revision 1 had asserted as an inherited assumption
in its header rather than as a departure.

1. **The `code-simplifier` round is batched after Task 3**, against
   CLAUDE.md's before-committing rule, because the chain commits per
   task and a subagent may not dispatch an agent. 8a and 8b both took
   the same departure, and this plan states it rather than inheriting it
   silently.
2. **Two of the five reader tests stay in `shadow`** (R2), against
   `spec:995-998`, because their subject is the projection ruling 3
   keeps there. Task 4 drafts the spec correction.
3. **The ceiling is 2.1M against the spec's proposed 1M**, re-derived
   from 8b's actual under the spec's own re-derivation clause
   (`spec:1226-1228`). Task 4 drafts the correction, and the rung's
   ceilings then stand at 9.85M against an 8M reservation with an
   expected actual of about 8.42M.
4. **`tagio` gains one exported symbol at a pass the spec calls
   mechanical** (R1), because 8c is the first caller needing a file's
   tags and its STREAMINFO from one parse. **This is a departure from a
   decision 8b closed, not the exercise of an inherited option**: R12's
   pick between a second read entry point and an exported projection over
   `Container` was handed to rung 9 in three places (`spec:766-774`,
   `docs/HISTORY.md:830-833`, `docs/HISTORY.md:990-993`), and what the 8b
   handoff granted 8c was the double parse
   (`docs/HISTORY.md:871-875`). Taking the pick here costs rung 9 a
   decision that was assigned to it, and the gain is that 8b's coverage
   gap 4 retires now rather than at rung 9. It is **conductor-owed**, and
   its fold adds no symbol at the cost of a duplicate parse per track.
   **Ruled, conductor 2026-09-12:** **As applied**, stated to break
   `8c:611`'s inverted use of "fold": the shadow consumes `Parse` and
   `tagio.Tags`, taking the 8b handoff's **second half**; `tagio` gains
   one exported symbol; one parse per file; 8b's coverage gap 4 retires;
   **Task 2 stays `sonnet`**. Deviation 5 is ruled with this and needs
   no separate mark (`8c:1601-1602`).
5. **The shadow's read path consumes `Parse` and never the seam
   registry**, under R1 as applied, so the 8b handoff's registry-consumer
   sentence (`docs/HISTORY.md:861-864`) is answered by its own second half
   (`docs/HISTORY.md:871-875`) rather than satisfied. `Parse` validates
   the `fLaC` signature itself (`container.go:147-150`), so nothing about
   FLAC behavior changes, and the deviation is documentary and
   architectural. **Conductor-owed**, ruled with R1; R1's fold retires
   this deviation entirely.

## Execution model

Workflow mode per the standing grant: the pass-execute chain
(`~/.claude/workflows/pass-execute.js`) per task, **strictly
sequential**. Nothing here runs in parallel, and no second executor
enters this tree. Every task runs `dubplate-implementer`
(`.claude/agents/dubplate-implementer.md`), then `diff-reviewer`
(`claude-opus-5`), then the gate, with one re-dispatch on `fix` and a
second `fix` as the conductor's decision. The conductor dispatches **one
batched `code-simplifier` round** after Task 3 and before Task 4, owns
every ledger edit, owns the lane registry and the ownership-table
amendment, owns the read-only box reading, and owns the batch commit of
anything a task drafts. **The runaway guard is armed at launch.
Subagents never dispatch agents.**

**The pass's deliverable count is 15**: Task 0's blocker report, Task
1's three, Task 2's five, Task 3's two, and Task 4's four. Task 2 is the
one task above CLAUDE.md's roughly-four line, argued at its own site.
Anything added after dispatch is named at the checkpoint rather than
absorbed.

**Independence.** No task pair in this pass is independent. Task 1 must
precede Task 2, because `buildFLAC` cannot be deleted while an external
caller remains. Task 3 must follow Task 2, because the tightening is
unobservable until the absorbed path is live. Task 4 follows everything.
The sequential chain is the mode regardless, because one tree runs one
executor and the module gate runs over `./...`.

**Models.** `sonnet` by default. **Task 0 runs at `opus`**, fresh
context: a staleness check that misses a moved citation costs more than
the model difference.

**One named conditional upshift.** **Task 2 to `opus`** if Task 0
reports any drift in `flac.go`, its call sites or `tagio`'s exported
signatures, or if the conductor rules R1's fold rather than R1 as
applied. Task 2 composes a deletion, a projection rewrite and a
seven-test disposition in one place, and its correctness rests on two
projections agreeing row for row. It stays `sonnet` while the tree
matches what this plan read, because every row it must preserve is
written down in a landed test.

**Review.** Rung 8 is not one of the three full-adversarial-cycle rungs
(5, 12 and 16), so review scales to the in-chain `diff-reviewer` per
task plus the pass-end gate (`spec:1356-1364`). The `shadow` package's
own suite is this pass's golden: it was written against the reader being
deleted, and it is the only thing that can prove the projection's
behavior did not change.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- **The conductor rules R1 and R2 before Task 2 dispatches.**
- Task 1 → Task 2 → Task 3.
- **The batched `code-simplifier` round runs after Task 3 and before
  Task 4**, and its result is committed on its own.
- **The conductor's read-only box reading happens after Task 3 and
  before Task 4's draft**, so the close has the denominator or the
  reason it does not.
- Tasks 0 through 3 → Task 4.

**Where this pass splits, if it splits.** Task 0 plus four tasks is well
inside the spec's fifteen-task trigger (`spec:1177`), so **the
recommendation is to keep one pass.** The cut point, if the Task 2
re-derivation puts the ceiling past 3M, is the **Task 2 / Task 3
boundary**, with Tasks 3 and 4 becoming 8c-ii. The argument against
taking it is that the tightening is the rung's one behavioral change and
closing the rung without it would ship an ungraded change to a live
evidence path.

**Budget:** 2.1M ceiling, re-derived after Task 2. Rung 8's ceilings
stand at 9.85M against the spec's 8M reservation, with the measured
actuals 1.25M under those ceilings through 8b. Flag at 80%, which is
**1.68M**, routed to the conductor. STATUS after every four tasks, so at
the close. Spend is scored at close against the ceiling and against
**zero attended interactions**, which is this pass's target. **A
projection row that will not agree is not an execution sitting**: it is
a finding the conductor rules, and only a change to what an adopter sees
would make it Geoff's.
