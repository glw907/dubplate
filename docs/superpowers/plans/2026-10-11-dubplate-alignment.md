# Dubplate alignment with the lean pass process: implementation plan

**Design:** `docs/superpowers/specs/2026-10-11-dubplate-alignment-design.md` (approved by Geoff, 2026-10-11, with
five binding rulings). **Evidence:** `docs/superpowers/research/2026-10-11-dubplate-alignment-audit.md` (cited below
as "audit", by section and item). The design is short enough that no separate spec exists. The lean process:
`~/.claude/skills/pass-core/SKILL.md` and the `dubplate-pass` adapter.

**Outcome.** A fresh dubplate session loads only the lean process. CLAUDE.md, the adapter, STATUS, ROADMAP, and the
project auto-memory carry no lanes, lane registry, conductor rules, errata ledgers, phase post-mortems, or
workflow-mode grants. The ladder keeps its product order and gains stable names, ROADMAP carries the one numbered
"Path to 0.8", and the `admin-api-1` re-plan can start on aligned ground.

**Pass class:** `runner` (the design's class: T0, T6, and the close touch `scripts/` and `.claude/`). T7 moves
lessons into doc comments under `internal/serve/` and `cmd/dubplate-server/`, which the adapter's path map classes
`auth-data`, so T7 is `auth-data` and the pass takes `auth-data` (pass-core: "The design's class is a floor"). Every
T7 Go edit is comment-only, and its acceptance proves that. No task touches the box, a Cloudflare account, or npm. The
session pushes a branch and opens a PR on `glw907/dubplate`.

**Out of scope:** rung work (`admin-api-1` is the next pass); PR test CI (ruling 3); the roughly 40 process citations
in shipped Go comments (audit item 15, a queued chore); the "boundary gate" wording at `scripts/check.sh:31, 41, 253`
(audit item 14; `check.sh` is gate machinery, so it goes to "Found, not done"); deleting or editing anything outside
this pass's worktree. The main checkout's untracked files (`reports/`, `research_notes/`, two 2026-09-04 research
notes, four untracked agent-memory files) and its modified `.claude/agent-memory/dubplate-implementer/MEMORY.md` are
the conducting session's. T1 only reads the three Pi 5 reports.

**End-to-end check.** On the close head:

1. `bash scripts/gate.sh --merge` prints `gate exit: 0` (the full gate plus the branch-history gitleaks scan).
2. `grep -n -i -E 'lane registry|pass-execute|conducting loop|conductor ruling|errata owed|workflow mode|post-mortem'
   CLAUDE.md .claude/skills/dubplate-pass/SKILL.md docs/STATUS.md ROADMAP.md` prints nothing. The same grep over the
   proposed auto-memory index (T4 writes it to `$CLAUDE_JOB_DIR/tmp/MEMORY.md.proposed`) prints nothing.
3. `wc -l < docs/STATUS.md` prints 60 or less.
4. ROADMAP's "Path to 0.8" names every remaining rung (`admin-api`, `intake`, `pipeline`, `streamer`, `queues`,
   `migration`, `builtins`, `integration`, `cutover`, `closeout`) in the order derived below.

**Clock estimate:** 305 task-minutes (T0 15, T1 10, T2 40, T3 45, T4 40, T5 25, T6 30, T7 45, close 55). T2 and T3,
T4 and T6, and T5 and T7 run as pairs, so the path is about 15 + 10 + 45 + 40 + 45 + 55 = 210 minutes.

## Path to 0.8: the derived run order

T2 writes this list into ROADMAP. It comes from the hard and soft edges in
`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md` ("The matrix", "The hard edges", "The soft
edges"), the re-plan (`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, "The ladder after this spec"),
the rung 7 and 14a specs, and ruling 2.

Landed: `foundation` (rungs 1 and 1b), `corpus` (2a to 2d), `matcher` (3a to 3d), `shadow` (4a and 4b; its program was
deleted on 2026-10-05), `auth` (5a to 5d), `store` (6a to 6d), `tag-io` (8a to 8c), and `streamer-1` to `streamer-3`
(slices 1, 2a, and 2b).

1. **`admin-api-1`** (14a). Its hard inbound edges are `auth` and `store` (matrix rows 5 and 6, column 14), both
   landed. The hard edge from `pipeline` (14,9) was discharged by Geoff's per-album lease ruling (14a spec, owner
   ruling 1). STATUS names it as the next action.
2. **`intake`**, as `intake-1` to `intake-3` (the rung 7 spec's 7a, 7b, and 7c). Ruling 2 puts it strictly after
   `admin-api-1`, which also settles the rung 7 spec's option of running 7a segments A and B beside 14a: they wait.
   Its hard inbound edge, `store` (6 → 7), is landed. It ends with Geoff's uploader sitting (rung 7 owner ruling 4).
3. **`admin-api-2`** (14b). It follows the first rung that writes a collection row (14a spec, "The split"), and
   `intake` creates the default collection (rung 7 spec, "The default collection"; "14b follows rung 7").
4. **`pipeline`** (rung 9, L, split into passes at its own planning). Its hard edges, `store` and `tag-io` (6 → 9,
   8 → 9), are landed. Its soft edge from `intake` (7 → 9: the handoff record and the `accepted` state rung 7
   defines) costs a stub producer if `pipeline` goes first, so it follows `intake`. It shares no files with
   `admin-api-2`, so the two may run at once (ruling 1).
5. **`streamer`**, slices 3 to 8 as `streamer-4` onward (a slice may take more than one pass, and each pass takes the
   next free number). Slice 3 waits on `pipeline`'s catalog (re-plan, "Rung 12"). It goes first among the rungs that
   open after `pipeline` because it is the longest (L, six slices) and 0.8 ships on it. Slice 7 consumes the decision
   queue (streamer spec hook 16), so `queues` lands before slice 7.
6. **`queues`** (rung 13). Hard on `matcher`, `store`, and `pipeline` (3, 6, 9 → 13). Ahead of `migration` and
   `builtins` because streamer slice 7 waits on it.
7. **`migration`** (rung 10, migration and layout parity). Hard on `corpus`, `store`, and `pipeline`
   (2, 6, 9 → 10).
8. **`builtins`** (rung 11). Hard on `store`, `tag-io`, and `pipeline` (6, 8, 9 → 11). Mutually independent of
   `migration` and `queues` (dependency analysis, "Wave C").
9. **`library-web-1`**, a web pass outside the ladder: the album page, the library grid and list, filtering, and buy
   links, which rung 7 owner ruling 2 moved to "a web pass after rung 9". It reads `pipeline`'s catalog.
10. **`integration`** (rung 15). Hard on every component rung by the launch bar's completeness part.
11. **`cutover`** (rung 16): the first launch, 0.8.0, ending in Geoff's first-launch sitting.
12. **`closeout`** (rung 17). Latest-bound after `migration` (10 → 17: the legacy ledger fixture is captured before
    `_salvage/` is deleted).

Items 5 to 9 open together once `pipeline` lands, so up to three of them run at once (ruling 1); the numbers set the
order in which they start. Items 10 to 12 are strictly serial (dependency analysis, "Wave D").

## Launch

Prompt file: `docs/superpowers/plans/2026-10-11-dubplate-alignment-prompt.txt`. The plan lives on branch
`worktree-alignment-plan` (the approved design, the audit, and the planning commits). The executing session continues
that branch: its first act after `EnterWorktree` is `git merge --no-edit worktree-alignment-plan`, so its PR carries
the design, the plan, and the work together. Launch from the main checkout, after every check prints what its comment
names:

```bash
git -C ~/Projects/dubplate fetch origin
git -C ~/Projects/dubplate branch --show-current                  # must print master
git -C ~/Projects/dubplate rev-list --count HEAD..origin/master   # must print 0 (not behind)
git -C ~/Projects/dubplate rev-list --count origin/master..HEAD   # must print 0 (not ahead)
git -C ~/Projects/dubplate status --porcelain --untracked-files=no   # only " M .claude/agent-memory/dubplate-implementer/MEMORY.md"
git -C ~/Projects/dubplate rev-parse worktree-alignment-plan origin/worktree-alignment-plan   # two equal lines
claude agents --json                                              # no live session with cwd under ~/Projects/dubplate/.claude/worktrees/dubplate-alignment
cd ~/Projects/dubplate && claude --bg --model sonnet --effort medium --name dubplate-alignment \
  "$(git -C ~/Projects/dubplate show worktree-alignment-plan:docs/superpowers/plans/2026-10-11-dubplate-alignment-prompt.txt)"
```

The prompt is read with `git show` because the plan is not on `master` until this pass's PR merges. The ahead check
matters because `worktree.baseRef` is `"head"` in `~/.claude/settings.json`, so an unpushed `master` commit would ride
into the PR. The one modified tracked file is the conducting session's and does not reach a worktree branched from
`HEAD`.

## Conventions every task follows

- **Gate.** `bash scripts/gate.sh` (the full gate) and `bash scripts/gate.sh --scope-changed` (the fast lane), from
  the worktree root, with the Bash tool's `timeout: 600000`, re-issued on exit 75 until it prints `gate exit:`. T0
  adds the wrapper. A worktree-isolated session's Bash guard refuses `cairn-run-gate 'bash scripts/check.sh'` ("what
  it runs cannot be shown not to be git"), and accepts a plain `bash <script> <flags>` call whose script execs
  `cairn-run-gate` (probed 2026-10-11 from the planning worktree: `--scope-changed --dry-run` printed `gate exit: 0`).
  Until T0 lands, the only gate the session runs is T0's own acceptance. Never fall back to a bare
  `bash scripts/check.sh` for an acceptance. A task's acceptance is the fast lane; the close runs the full form.
- **Setup.** `npm ci --prefix web` in each new worktree, the session's and each pair's, before the first gate (the
  gate refuses a worktree without `web/node_modules`).
- **The main checkout stays untouched.** Every command runs from the task's own worktree. T1 reads three files in
  `~/Projects/dubplate/reports/` and writes nothing there. No task writes, moves, or deletes a file outside its
  worktree, and that includes the auto-memory directory
  `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/`, which T4 only reads.
- **Names.** The rungs are `foundation`, `corpus`, `matcher`, `shadow`, `auth`, `store`, `intake`, `tag-io`,
  `pipeline`, `migration`, `builtins`, `streamer`, `queues`, `admin-api`, `integration`, `cutover`, and `closeout`
  (ruling 5), for rungs 1 to 17 in order. A pass is `<name>-<n>`: `admin-api-1` is 14a, `admin-api-2` is 14b,
  `intake-1` to `intake-3` are 7a to 7c, and `streamer-1` to `streamer-3` are slices 1, 2a, and 2b. A landed pass with
  a letter id (5c, 6d) keeps that id in git and in frozen documents; a live document that cites one writes the rung
  name and what the pass did ("the config registry pass", not "rung 1b"). Frozen specs keep their old numbers, and no
  errata sweep follows. Only ROADMAP's "Path to 0.8" uses a number for order.
- **Voice.** Prose follows the `writing-voice` output style, with Go comments under `go-conventions`, shell comments
  plain, and no em dash in any comment. No new process citation (rung, task, ruling, or plan line) enters a shipped Go
  comment; the gate's process-citation leg warns on one.
- **Commits.** One commit per task, subject starting with the task id (`T2: Add the Path to 0.8 and the rung names to
  ROADMAP`), imperative mood, specific files only (`git add <path>...`, never `-A`), with the trailer the session's
  system reminder names.
- **PR notes.** Every task that owes the PR body a record (T3's dispositions, T4's dry run, T5's triage) appends it to
  `$CLAUDE_JOB_DIR/tmp/pr-notes.md` in the session's job directory; a pair subagent puts it in its final report and the
  session appends it. The close builds the PR body from that file.
- **Clock.** At each task start the session writes `pass-task-clock` in its pass worktree's git dir
  (`git rev-parse --absolute-git-dir`) with `task=<id>`, `start=<epoch>`, `estimate=<minutes>`, and removes it at the
  task's end. For a pair it writes one file (`task=T2+T3`, the larger estimate) before dispatch, checks the overrun
  when both return, then removes it. Pair worktrees hold no clock file.
- **Pairs.** Each pair is two Agent-tool subagents, `subagent_type: "general-purpose"`, `isolation: "worktree"`,
  model `sonnet`, effort `medium`, dispatched after the session commits, so `worktree.baseRef: "head"` branches them
  from that commit. Each gets its task's full text and this section verbatim, runs the setup, commits, and never
  pushes. T7's dispatch also names the `go-conventions` skill, and T5's names the visual identity spec and the
  `python-conventions` skill (for its one possible docstring). The session merges both halves, runs the fast lane, and pushes.
- **Stops.** A red the session cannot fix in two attempts, a merge conflict, a clock stop, or a finding that would
  change the approved design: write `docs/STATUS.md` on the branch (task in flight, what failed, one question with a
  recommendation), run `notify-send -u normal "dubplate alignment: stopped"`, and wait. A pair subagent whose gate is
  not green after two attempts returns its work uncommitted, and the session applies this rule before merging either
  half.
- **Scope.** A defect found outside a task goes to the PR body's "Found, not done" list.

## Tasks

### T0. The gate wrapper and the session record

- **Outcome.** `scripts/gate.sh` runs `bash scripts/check.sh` through `cairn-run-gate`, passing its own arguments
  through (`--scope-changed`, `--merge`, `--dry-run`), from the repo root, on the default (heavy) lane, since the full
  gate's rendered leg launches Chromium. A header comment says why it exists (the worktree guard refuses the quoted
  form) and that its exit codes are `cairn-run-gate`'s, so 75 still means re-issue; `~/.dotfiles/scripts/gate.sh` is
  the model. The adapter's Gate, Fast lane, and Close checklist lines name the wrapper (T6 rewrites the rest of the
  adapter; T0 edits only those three lines). CLAUDE.md's Shell bullet (`CLAUDE.md:89-92`) lists `gate.sh` beside the
  other ops scripts. STATUS gains one line under "Next action" recording this session's name
  (`dubplate-alignment`) and id as in flight; T3 carries it into the rewrite.
- **Files.** `scripts/gate.sh` (new, executable), `.claude/skills/dubplate-pass/SKILL.md` (lines 58-64 and 93),
  `CLAUDE.md` (lines 89-92), `docs/STATUS.md`.
- **Risk class.** `runner`.
- **Acceptance.** `bash scripts/gate.sh --scope-changed --dry-run` prints a `check.sh: scope=changed` line and
  `gate exit: 0`. `bash scripts/gate.sh --scope-changed` prints `gate exit: 0`; its scope includes `scripts/gate.sh`,
  so shellcheck reads it. `test -x scripts/gate.sh` succeeds.
- **Test.** The acceptance runs are the test: the wrapper either reaches `cairn-run-gate` past the guard or it does
  not.
- **Estimate.** 15 minutes. The session runs it inline, first.

### T1. The Pi 5 client research

- **Outcome.** Geoff's direction (2026-10-11, relayed by the conducting session): the three Pi 5 stereo client
  reports written 2026-10-10 are committed as live input to the 0.9 local client. Copies, byte for byte, land in
  `docs/superpowers/research/2026-10-10-pi5-client/`: `proven-build.md` (from "Pi 5 stereo client proven build.md",
  the current decision), `verified-parts.md` (from "Pi 5 stereo client verified parts.md"), and `hardware.md` (from
  "Pi 5 stereo client hardware.md"), the last two its evidence. ROADMAP's `local-client-0-9` entry
  (`ROADMAP.md:117-142`) gains one line pointing to `proven-build.md`.
- **Files.** The three new files, `ROADMAP.md` (one line).
- **Risk class.** `ordinary`.
- **Acceptance.** `cmp` of each copy against its original under `/var/home/glw907/Projects/dubplate/reports/`
  (absolute paths, quoted, since the names hold spaces) exits 0, and the originals' modification times are unchanged
  (`stat -c %Y` before and after). The fast lane prints `gate exit: 0`.
- **Test.** The `cmp` runs.
- **Estimate.** 10 minutes. The session runs it inline after T0, before the first pair, because T2 rewrites ROADMAP.

### T2. ROADMAP: the Path to 0.8, the names, and a lean shape

- **Outcome.**
  - A new section, "Path to 0.8", directly under the header block: the twelve numbered items of this plan's derived
    order, each with its rung name, the old number in parentheses, one line of what it delivers, and its inbound
    edges in a short clause. Below the list, one sentence: items whose edges are met may run at once, up to three
    passes (ruling 1); the numbers set start order and are rewritten when the order changes. Each item names where
    its carried work lives: BACKLOG items tagged `#rung:<name>` (T3 writes them); for `intake`, the rung 7 spec's
    living-reference list (T6 reshapes it); for `admin-api-1`, its planning prompt (T3 writes it).
  - A mapping table under it: old rung number, name, landed passes (old ids), and the new pass ids, covering all 17
    rungs plus `library-web-1`. One line above it says frozen documents keep the old numbers and this table bridges
    them.
  - Active: the greenfield entry (`ROADMAP.md:14-65`) and the rung 12 entry (`:67-113`) are trimmed to goal and next
    step: the goal, the launch bar in one sentence, "next: `admin-api-1`", and a pointer to "Path to 0.8". Landed
    status leaves (git and the mapping table carry it). The organizer spec's folded revisions (`:49-65`) and the
    streamer's settled list (`:78-110`) collapse into pointers to the organizer and streamer specs. The standing 0.8
    rule at `:97-101` goes; its home is `docs/standing-rules.md` "Build for 0.8", which CLAUDE.md points to. The
    TypeScript 7 recheck (`:111-113`) moves to BACKLOG as one item tagged `#rung:integration` (T3 owns BACKLOG, so T2
    hands the text to T3 through the session: T2's report quotes it).
  - Planned: `pr-test-ci` gains "after `admin-api-1`" (ruling 3). `prod-polish` loses the owed post-mortem
    (`:214-215`). The folded organizer entry (`:221-225`) goes. Other Planned entries keep their substance; rung
    numbers become names.
  - The Done section (`:292-306`) goes.
- **Files.** `ROADMAP.md`.
- **Risk class.** `ordinary`.
- **Acceptance.** End-to-end checks 2 and 4 hold for ROADMAP. `grep -n '^## ' ROADMAP.md` prints, in order, the Path
  section, Active, Planned, and Someday, and no Done. `grep -n -E 'rung [0-9]+' ROADMAP.md` matches only the mapping
  table and the Path items' parenthesized old numbers. T1's pointer line survives. The fast lane prints
  `gate exit: 0`.
- **Test.** A read of the Path list against this plan's derived order, item by item.
- **Pair.** Runs as a pair with T3 (disjoint files).
- **Estimate.** 40 minutes.

### T3. STATUS, BACKLOG, the errata dispositions, and the admin-api-1 planning prompt

- **Outcome.**
  - `docs/STATUS.md` rewritten to pass-core's three parts, present tense, 60 lines or fewer:
    - Current state: master's last green full gate, nothing deployed, the box as now described (`STATUS.md:9-11`),
      and this pass in flight (T0's line).
    - Next action: the `admin-api-1` re-plan, an unattended Opus planning session from the 14a spec (revision 2.1,
      approved by Geoff 2026-10-10, its "Rulings for Geoff" section "None"), with the exact launch line
      `cd ~/Projects/dubplate && claude --bg --model opus --effort high --name admin-api-1-plan "$(cat
      docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt)"`, run after this pass's PR merges.
    - Open decisions: R-6d21 (`STATUS.md:22`), the group non-FLAC question (`:23`), the stale `DUBPLATE_*` exports in
      `~/.local/secrets` (`:24`), and the one Run-D.M.C. photo frame (`:31`). The rung 7 spec question (`:21`) leaves
      (T6 replaces that gate).
    - The lane registry (`:26-31`), the Owed list (`:33-40`), and the open items (`:42-60`) leave.
  - Every open item at `STATUS.md:42-60` becomes a BACKLOG item, one per distinct item, tagged with the rung that
    consumes it (`#rung:admin-api`, `#rung:cutover`, `#rung:streamer`, `#rung:intake`, `#rung:pipeline`, and so on;
    a rung-free chore keeps its existing tags). Every pointer into the frozen HISTORY (`STATUS.md:4, 11, 56, 57, 60`)
    is resolved by reading the cited HISTORY section ("Moved out of STATUS at this close" at `HISTORY.md:316, 511,
    588`; "Routed items" at `:443`; the bearer-bypass cookie test) and writing each live item into BACKLOG in its own
    words, with no HISTORY pointer left. Items already in BACKLOG are not duplicated; the existing number is tagged
    instead. Also into BACKLOG: `make derive` once before the next oracle-dependent pass (`STATUS.md:39-40`); the
    14a fold's routing of `serve`'s job runner, instance lease, and job recovery to the first rung whose page starts a
    job (14a spec fold, "Conductor-owed"); C14's restore check, tagged `#rung:cutover`; and T2's TypeScript 7 item.
    BACKLOG #1 (the Go-rewrite post-mortem) is closed with the note "closed under lean ruling 8: history is git and
    PR bodies".
  - A disposition line for each STATUS errata item, appended to the PR notes:
    - Rung 7's eight documents (`STATUS.md:34-36`): SS, CH, W, VI, PCF, and CLAUDE.md are living references, so their
      edits move into the `intake` passes (T6 reshapes the rung 7 spec's list); 14A and BS are frozen, so those two
      are moot.
    - The 14a spec's eight owed errata (14a spec fold, "Owed errata" items 1 to 8): items 1 to 4 and 8 (R6), 6 (PA),
      and 7 (BS) are moot, frozen documents; item 5 (CH:171-172) and the PS member-cap note are living-reference edits
      that move into `admin-api-1` through its planning prompt. The plan fold's contested-surfaces erratum is moot (the
      tables retire).
    - `docs/roster-walk.md`'s Tempus pin (`STATUS.md:37-38`): a living reference, edited by `streamer-4`, the next
      pass that walks it (a BACKLOG item tagged `#rung:streamer`).
    - The `id3-tags` copy (`:38-39`): part of the PCF edits `intake` makes. `make derive` (`:39-40`): BACKLOG.
    - The 14a fold's open question on rung 7 and a streamer slice both writing `internal/serve`: moot; each design
      names its files, and with two PRs open the second merges master and reruns its close gate.
  - `docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt`, shaped like this pass's planning prompt
    (`docs/superpowers/plans/2026-10-11-dubplate-alignment-planning-prompt.txt`): EnterWorktree name
    `admin-api-1-plan`; load `dubplate-pass`; read the 14a spec in full as the approved design, the old plan
    `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md` as evidence only, and BACKLOG's `#rung:admin-api` items;
    write `docs/superpowers/plans/<date>-admin-api-1.md` in the lean format (this plan is the model), class
    `auth-data`, with the CH:171-172 and PS member-cap edits as tasks; run `spec-plan-review` with all three lenses,
    the fold, and the verification read; write the executor's launch prompt beside it; a finding that changes the
    approved design is the one stop (STATUS question, `notify-send -u normal`); commit and push to `glw907/dubplate`,
    never `master`, never merge, never open a PR; notify `"admin-api-1: plan ready to launch"` and end.
- **Files.** `docs/STATUS.md`, `BACKLOG.md`, `docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt`
  (new).
- **Risk class.** `ordinary`.
- **Acceptance.** End-to-end checks 2 and 3 hold for STATUS. `grep -n 'HISTORY' docs/STATUS.md` prints nothing.
  `grep -c '#rung:' BACKLOG.md` is at least the number of distinct items at `STATUS.md:42-60` (the subagent states
  its count). BACKLOG #1 is checked. The prompt file names the 14a spec path, `auth-data`, the three lenses, and the
  push target `glw907/dubplate`. The fast lane prints `gate exit: 0`.
- **Test.** A side-by-side read: each old STATUS line maps to a new STATUS line, a BACKLOG item, a disposition, or
  "retired" (the lane registry, the lean-cutover next action); the subagent's report carries that map.
- **Pair.** Runs as a pair with T2.
- **Estimate.** 45 minutes.

### T4. CLAUDE.md, the implementer memory, and the auto-memory dry run

- **Outcome.**
  - CLAUDE.md:
    - The initiative paragraph (`:10-24`) uses the rung names, says "the ladder (ROADMAP, Path to 0.8)", and states
      the two standing bars outright instead of "from memory". Exemplary Go: idiomatic standard-library style and
      architectural beauty, never mere correctness. Render and read every user-path page before any attended ask.
      Each carries its one-line why from its memory file (`idiomatic-beautiful-go.md`,
      `see-it-before-geoff-tests-it.md`).
    - Product rulings that live only in auto-memory become clauses where they execute in CLAUDE.md: "the group" and
      "a member", never "family", except in the product-essence line (`group-not-family.md`); plain terms in specs and
      UI copy, with no coined metaphor or record-shop pun (`plain-terms-no-coined-jargon.md`); the Svelte and DaisyUI
      idiom bar, where the stack's idiom wins over a finding's literal fix text (`idiomatic-svelte-daisyui.md`), in
      the Web bullet. Version numbering and the own visual identity are already in CLAUDE.md (`:26-33`, `:103-106`,
      `:133-136`) and need no clause.
    - The errata-avoidance prose (`:29-33`) becomes one sentence: documents older than a ruling keep their numbers
      and stay as written.
    - "Build for 0.8" (`:35-38`) names the `streamer` rung instead of "Rung 12", and keeps its pointer to
      `docs/standing-rules.md`.
    - Ledgers (`:47-56`): HISTORY is "a frozen archive (2026-10-10); git and PR bodies are the record"; the ROADMAP
      line drops the T14 migration note and names "Path to 0.8".
    - Build process (`:58-66`): the phase names (foundation, evidence, spine, surfaces, launch) as ROADMAP
      milestones, "Process: the `dubplate-pass` skill, loading `pass-core`", and the product touchpoints: corpus
      adjudication batches, the `intake` uploader sitting, and the first-launch sitting. The design approval and the
      PR read are pass-core's. No pointer to the build spec's Process section or the parallel-build spec remains.
    - The Go bullet (`:82-88`) gains the test-echoes-code clause: a test's expected value is a literal or an
      independently computed value, never built with the helper the code under test calls
      (`.claude/agent-memory/dubplate-implementer/feedback_test-echoes-code.md`).
    - Remaining rung numbers become names (`:101-102`, `:135`, `:170`), per the Names convention.
    - Safety rails (`:174-177`): before any executor touches a worktree, `pgrep -f <worktree path>` and `git status`
      for warm changes you did not author; STATUS records each running pass's session name and id. The box rules stay.
  - `.claude/agent-memory/dubplate-implementer/MEMORY.md` and `feedback_test-echoes-code.md` are removed with
    `git rm` (the agent was deleted at `8269d27`); the directory goes with them in this worktree.
  - The auto-memory dry run. T4 reads every file in `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/`
    (41 files plus `MEMORY.md`) and classes each as one of: delete, retired process; delete, session or window log;
    delete, duplicate (naming the repo file and line that now carries it, confirmed by `grep`); or keep, a preference
    or a ruling's why that no file holds. The provisional classes, which T4 confirms or corrects with a reason:
    - Retired process: `greenfield-arc-sessions`, `pass-execute-cross-session-resume`,
      `pass-execute-dispatch-mechanics`, `two-lane-run-mechanics`, `ledger-fold-must-keep-list`,
      `overnight-hold-sized-to-clock` (the inhibitor is tool-held now), and `spec-prose-review-rule` (the
      `spec-plan-review` skill now runs no prose review on a spec or plan).
    - Session or window logs: `unattended-window-2026-09-05`, `session-2026-09-07-4a-ii-close`,
      `unattended-window-2026-09-07`, `session-2026-09-08-wind-down`, `session-2026-09-12-config-homes-close`,
      `unattended-window-2026-09-12`, `session-2026-09-12-infra-round`, `overnight-window-2026-09-23`,
      `session-2026-09-25-handoff`, `overnight-window-2026-09-25`, `unattended-window-2026-09-26`,
      `overnight-window-2026-09-26`, `photo-round-one-closed`.
    - Duplicates once T4 lands: `version-numbering`, `build-for-2-0`, `greenfield-server-pivot`, `own-visual-identity`,
      `group-not-family`, `plain-terms-no-coined-jargon`, `idiomatic-beautiful-go`, `idiomatic-svelte-daisyui`,
      `see-it-before-geoff-tests-it`; and, each only if `grep` finds its ruling in the named spec,
      `configuration-homes-rulings` (CH), `group-scale-principles-rulings` (PS), `rung-5-owner-rulings` (the rung 5
      spec), `fork-b-formats-ruling` and `give-back-go-libraries` (ROADMAP), `web-design-standards-sources` (the web
      surface standard).
    - Keep: `irreversible-means-bound-not-defer`, `no-calendar-schedule`, `no-ceremony-rulings`,
      `photo-older-is-better`, `process-framing-not-geoffs-rules`, `workstation-vs-project-scope-by-survey`.
  - T4 writes the dry run to the PR notes as a table (file, class, reason) and writes the proposed index, the kept
    files' lines from the current `MEMORY.md` and nothing else, to `$CLAUDE_JOB_DIR/tmp/MEMORY.md.proposed`, then
    appends that text to the PR notes. It deletes nothing and writes nothing in the memory directory.
- **Files.** `CLAUDE.md`, `.claude/agent-memory/dubplate-implementer/MEMORY.md` (removed),
  `.claude/agent-memory/dubplate-implementer/feedback_test-echoes-code.md` (removed).
- **Risk class.** `runner` (`.claude/`).
- **Acceptance.** End-to-end check 2 holds for CLAUDE.md and for `$CLAUDE_JOB_DIR/tmp/MEMORY.md.proposed`.
  `grep -n -i 'from memory\|parallel-build-approach\|at T14\|lane' CLAUDE.md` prints nothing except a "heavy lane" or
  "light lane" gate reference. `git ls-files .claude/agent-memory/dubplate-implementer` prints nothing. Every file in
  the memory directory appears once in the dry-run table. `wc -l < CLAUDE.md` stays at or under 200. A `find` of the
  memory directory with `-newer` a stamp written at task start prints nothing. The fast lane prints `gate exit: 0`.
- **Test.** The dry-run table's row count equals `ls` of the memory directory, minus `MEMORY.md`.
- **Pair.** Runs as a pair with T6 (disjoint files; T0's CLAUDE.md line is already committed).
- **Estimate.** 40 minutes.

### T5. The site-implementer memory to its executing homes

- **Outcome.** `.claude/agent-memory/site-implementer/` (11 tracked lessons plus `MEMORY.md`) goes, since the adapter
  stops naming `site-implementer` (T6) and no other agent loads it. Each lesson lands where it executes or is deleted
  with a one-line reason in the PR notes:
  - The six DaisyUI lessons (`daisyui-*.md`): a lesson about a theme token becomes a short comment beside that token
    in `web/src/lib/themes/*.css` (`daisyui-semantic-theme-tokens`, `daisyui-soft-pastel-vs-saturated`,
    `daisyui-soft-reads-color-token-as-text`); a lesson about a component's markup becomes a sentence in the visual
    identity spec's "Component layer" section
    (`docs/superpowers/specs/2026-09-24-visual-identity-design.md:226`), a living reference
    (`daisyui-size-field-touch-targets`, `daisyui-alert-grid-single-wrapper`, `daisyui-horizontal-steps-baseline`).
    Where the theme files have no such token, the lesson goes to the spec instead.
  - `admin-identity-in-copy`: a sentence in the visual identity spec's "Copy" section (`:314`) if `grep` shows no
    display-name field still exists; otherwise deleted as stale.
  - Deleted with reasons: `commit-only-named-paths` (pass-core and the global git policy carry it),
    `app-svelte-no-store-di` (`App.svelte` is salvage-era; `web/src` has none), `systemd-python-subprocess` (the box
    runs no beets), `beets-dry-run` (if the oracle shim under `internal/harness/oracle/` already documents its beets
    invocation; otherwise it becomes that shim's module docstring sentence under `python-conventions`).
- **Files.** `.claude/agent-memory/site-implementer/*` (removed), `web/src/lib/themes/*.css` (comments only),
  `docs/superpowers/specs/2026-09-24-visual-identity-design.md`, and possibly one `internal/harness/oracle/*.py`
  docstring.
- **Risk class.** `runner` (`.claude/`).
- **Acceptance.** `git ls-files .claude/agent-memory/site-implementer` prints nothing. `git diff --stat` against the
  task base shows only comment lines changed in each CSS file (`git diff -U0 -- web/src/lib/themes | grep -E
  '^[+-][^+-]' | grep -v -E '^[+-]\s*(/\*|\*|//)'` prints nothing). Each of the 11 lessons has a row (home or
  deletion reason) in the PR notes. The fast lane prints `gate exit: 0` (its `web/` leg rebuilds the bundle).
- **Test.** The comment-only grep above.
- **Pair.** Runs as a pair with T7 (disjoint files).
- **Estimate.** 25 minutes.

### T6. The adapter, permissions, the frozen-spec headers, and the rung 7 gate

- **Outcome.**
  - `.claude/skills/dubplate-pass/SKILL.md`:
    - "One executor per tree" (`:18-22`) becomes the global check: `pgrep -f <worktree path>` and `git status` for
      warm changes; STATUS records each running pass's session name and id. The registry sentence goes. The box
      rules stay.
    - "Lanes" (`:29-31`) becomes "Concurrent passes": up to three dubplate passes at once, each its own session,
      worktree, and PR (Geoff, 2026-10-11, restating his 2026-10-05 three-code-lane cap). Each design names the files
      it touches. With two PRs open, the second to merge first merges master and reruns its close gate.
    - "Plans are just-in-time" (`:32-33`) speaks of passes, not lanes.
    - "Owner touchpoints" (`:43-44`) becomes the product touchpoints (corpus adjudication batches, the `intake`
      uploader sitting, the first-launch sitting), with the design approval and the PR read named as pass-core's.
    - "Pair subagents" (`:53-54`): `general-purpose`, given the adapter's setup and fast lane (`bash scripts/gate.sh
      --scope-changed`) and the `go-conventions` skill for Go or `svelte-conventions`, `ts-conventions`, and the
      visual identity spec for `web/`.
    - A new rule, "Errata": frozen, ratified specs get no errata, and a superseded one gets at most a one-line header
      pointer. A living reference (the streamer spec's forward hooks, PCF, the visual identity spec, CH, W, PS,
      `docs/roster-walk.md`) is edited in place by the pass that changes what it says, as a task in that pass. There
      is never an owed list.
    - A new rule, "Carried items": a pass's design reads BACKLOG's items tagged `#rung:<its rung>` and the ROADMAP
      "Path to 0.8" entry for its rung.
    - The worktree setup's last sentence (`:75-77`) says "its pass", not "its lane".
  - `.claude/settings.json`: the `docker compose config*` rule and the two box rules for `docker compose ... restart`
    and `systemctl restart` go (ruling 4). The `systemctl reset-failed` rule stays.
  - One-line superseded headers. Under the build spec's "## Process" heading
    (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:1143`): "Superseded 2026-10-11 by the lean pass
    process (`dubplate-pass`, `pass-core`); the ladder above still governs product order." Directly under the
    parallel-build spec's status line (`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:3`): "Superseded
    2026-10-11: execution follows the lean pass process (`dubplate-pass`); the run order is ROADMAP's Path to 0.8."
  - The rung 7 spec (`docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`): line 7's "It goes to Geoff before
    any plan is written" becomes "It is agent input: when `intake`'s turn comes, a one-page design goes to Geoff, and
    this spec carries the detail behind it." The "Owed records" section (`:1240-1262`) is retitled "Living-reference
    edits the intake passes make" and keeps the SS, CH, W, VI, PCF, and CLAUDE.md items as tasks for those passes;
    the 14A and BS items are struck with "moot (frozen; lean ruling 8)". No other line changes.
- **Files.** `.claude/skills/dubplate-pass/SKILL.md`, `.claude/settings.json`,
  `docs/superpowers/specs/2026-09-02-greenfield-server-build.md` (one line),
  `docs/superpowers/specs/2026-09-03-parallel-build-approach.md` (one line),
  `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`.
- **Risk class.** `runner`.
- **Acceptance.** End-to-end check 2 holds for the adapter. `grep -n -i 'lane' .claude/skills/dubplate-pass/SKILL.md`
  matches only "heavy lane", "light lane", or "fast lane". `grep -c 'site-implementer'` on the adapter prints 0.
  `python3 -I -c 'import json; d = json.load(open(".claude/settings.json")); print(d["permissions"]["allow"])'`
  parses and lists no `docker compose` or `systemctl restart` entry. `git diff --stat` shows one changed line in each
  of the two frozen specs. The fast lane prints `gate exit: 0`.
- **Test.** The JSON parse, and a read of the adapter against pass-core's adapter slots (gate, fast lane, full-suite
  home, worktree setup, CI watch, path map, checklist globs, close checklist), all still present.
- **Pair.** Runs as a pair with T4.
- **Estimate.** 30 minutes.

### T7. Lessons to their executing homes

- **Outcome.** `docs/standing-rules.md` "Lessons the code cannot carry" (`:113-170`) empties into the places that
  execute each lesson, and the section goes. Each becomes a Go doc comment or a file comment under `go-conventions`
  (no process citation, no em dash):
  - Oracle harness (`:119-136`): the Navidrome multi-library login and the gonic sub-path limit into the
    `internal/harness/subsonic/stack` package doc; the `collection:` key resolution into the session package's doc;
    gonic's error 10, the per-track cover, disc-subfolder splitting, and the two-second duration floor into
    `internal/harness/subsonic/fixture/doc.go`; the seed's missing lyrics, `nowPlaying`, and radio into the same doc;
    "the golden grades the grader" and the ledger-row rule into the `grade` package doc. The package doc comments
    live today in `stack/stack.go`, `session/session.go`, `grade/finding.go`, and `fixture/doc.go`; T7 extends
    those and adds no `doc.go`.
  - The `jsonpointer` and `swag` note (`:131-132`) is dropped: `go mod tidy` owns those `go.mod` lines, and the
    tidy-check leg enforces them.
  - Build for 0.8's stable hook numbers (`:138-139`) moves up into the "Build for 0.8" section of the same file.
  - Serve routing (`:141-143`) into the doc comment of the origin-wrapped mux in `internal/serve` (`serve.go:59` or
    `origin.go:78`, whichever builds the routed mux).
  - Settings (`:145-148`) into `Handle.Update`'s doc comment (`internal/engine/store/write.go:64`).
  - Web (`:150-153`) into a comment beside the kit options in `web/vite.config.ts`, where `adapter-static` is
    configured (no `svelte.config.js` exists).
  - Matcher and corpus (`:155-160`) into `internal/engine/musicbrainz/doc.go`.
  - Go tests (`:162-166`): the cobra usage-writer fact beside the first `SetOut` wiring in
    `cmd/dubplate-server/main_test.go`; the `doctor` row rule into `cmd/dubplate-server/doctor_test.go`'s file
    comment or the test that derives rows.
  - Shell (`:168-170`): a comment above `ssh_cidrs_json` in `scripts/provision.sh:123`, the function whose callers
    run it in a command substitution (`:143`, `:161`) and the subshell BACKLOG #5 names.
  - The standing-rules header line (`:3`) drops "plus lessons harvested from the frozen HISTORY".
- **Files.** `docs/standing-rules.md`, `internal/harness/subsonic/{stack,session,fixture,grade}/*.go` (doc comments
  only), `internal/serve/serve.go` or `origin.go`, `internal/engine/store/write.go`,
  `internal/engine/musicbrainz/doc.go`, `cmd/dubplate-server/main_test.go`, `cmd/dubplate-server/doctor_test.go`,
  `web/vite.config.ts`, `scripts/provision.sh`.
- **Risk class.** `auth-data` by path (`internal/serve/**`, `cmd/dubplate-server/**`), with comment-only edits.
- **Acceptance.**
  - Comment-only: `git diff -U0 <task base> -- '*.go' '*.ts' '*.sh' | grep -E '^[+-][^+-]' | grep -v -E
    '^[+-]\s*(//|#|\*|/\*)'` prints nothing.
  - `grep -n 'Lessons the code cannot carry' docs/standing-rules.md` prints nothing, and every lesson in the old
    section has a row (home, or "dropped" with its reason) in the subagent's report.
  - `go doc ./internal/harness/subsonic/fixture` and `go doc ./internal/engine/musicbrainz` print the moved text.
  - The fast lane prints `gate exit: 0`; its vale-over-Go-comments leg reads the new comments.
  - `diff-reviewer` at task scope reads the T7 commit (pass-core: the `auth-data` read) and returns `accept`.
- **Test.** The comment-only grep. Test-first does not apply: no behavior changes.
- **Pair.** Runs as a pair with T5.
- **Estimate.** 45 minutes.

### Close

Per pass-core "Close" and the adapter's close checklist, in order:

1. `git fetch origin` and merge `origin/master` into the branch (a conflict is a stop). Then
   `bash scripts/gate.sh --merge` prints `gate exit: 0` on the close head (the full-suite home, with the branch-history
   gitleaks scan). No pass task touches `cmd/harness` or `internal/harness` code, only comments, so the oracle-tagged
   `go vet` line is run once anyway: `go vet -tags oracle ./cmd/harness/... ./internal/harness/...`.
2. The end-to-end checks 2 to 4.
3. `diff-reviewer` at whole-branch scope reads the branch against this plan and the design. It loads the
   `go-conventions` skill (the changed `*.go`), and the checklist globs bring `web-auth-security-reviewer` for the T7
   paths under `internal/serve/` and `cmd/dubplate-server/`, and `svelte-reviewer` and `daisyui-a11y-reviewer` for
   `web/**` (T5's CSS comments and T7's `vite.config.ts` comment; the visual identity spec's path goes with them). One
   batched fix chain follows if any returns `fix`, each finding quoted verbatim as the fix's acceptance, then the gate
   again.
4. `git mv docs/superpowers/plans/2026-10-11-dubplate-alignment.md docs/superpowers/archive/plans/`. The two prompt
   files stay. STATUS's in-flight line for this pass becomes "PR open for Geoff's read"; STATUS stays 60 lines or
   fewer.
5. Push `glw907/dubplate`, open the PR, and write its body from the PR notes:
   - what landed (task id, commit), T7's `diff-reviewer` verdict, the close review's verdicts;
   - the errata dispositions (T3), the site-implementer triage (T5), and the lessons table (T7);
   - **After merge, the conducting session** (the pass session does none of these): (a) settle the main checkout's
     modified `.claude/agent-memory/dubplate-implementer/MEMORY.md` and its four untracked siblings, since the merge
     deletes the tracked directory and a pull will refuse over the local edit; (b) delete the auto-memory files the
     dry-run table lists, after showing the list (global CLAUDE.md: "Destructive ops: Show a dry run"); (c) write the
     proposed index as `MEMORY.md` and rerun end-to-end check 2's grep over it; (d) trash the three untracked Pi 5
     originals in `~/Projects/dubplate/reports/`; (e) launch the `admin-api-1` re-plan with STATUS's line;
   - the dry-run table and the proposed index (T4);
   - "Found, not done": the "boundary gate" wording in `scripts/check.sh`; the roughly 40 process citations in Go
     comments; anything the tasks found;
   - the score rows (pass-core "Score"; `cairn-run-gate --records <toplevel> <branch>` for gate and lock time).
6. `gh pr view <n> -R glw907/dubplate --web`, then `notify-send -u normal` with the PR link. Geoff reads and merges.

- **Estimate.** 55 minutes.
