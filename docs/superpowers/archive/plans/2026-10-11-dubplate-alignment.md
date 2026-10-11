# Dubplate alignment with the lean pass process: implementation plan

**Design:** `docs/superpowers/specs/2026-10-11-dubplate-alignment-design.md` (approved by Geoff, 2026-10-11, with
five binding rulings). **Evidence:** `docs/superpowers/research/2026-10-11-dubplate-alignment-audit.md` (cited below
as "audit"). **Review:** the three lenses and the fold record,
`docs/superpowers/research/2026-10-11-dubplate-alignment-plan-{review-contract,review-mechanics,review-risk,fold}.md`.
The design is short enough that no separate spec exists. The lean process: `~/.claude/skills/pass-core/SKILL.md` and
the `dubplate-pass` adapter.

**Outcome.** A fresh dubplate session loads only the lean process. CLAUDE.md, the adapter, STATUS, ROADMAP, and the
project auto-memory carry no lanes, lane registry, conductor rules, errata ledgers, phase post-mortems, or
workflow-mode grants. Every live item those files carried arrives somewhere a later pass reads. The ladder keeps its
product order and gains stable names, ROADMAP carries the one numbered "Path to 0.8", and the `admin-api-1` re-plan
can start on aligned ground.

**Pass class:** `auth-data` (T7's comment-only edit in `cmd/dubplate-server/main_test.go`, by the adapter's path
map; the design's `runner` is the floor), with the `live-account` flag: the session pushes a branch and opens a PR on
`glw907/dubplate`. No task touches the box or a Cloudflare account.

**Out of scope:** rung work (`admin-api-1` is the next pass); PR test CI (ruling 3); the roughly 40 process citations
in shipped Go comments (audit item 15, a queued chore); the "boundary gate" wording at `scripts/check.sh:31, 41, 253`
and any other `check.sh` edit (gate machinery; it goes to "Found, not done"); upgrading a workstation binary (brew's
golangci-lint, rclone, libxml2), which writes outside the worktree and records itself in the dotfiles manifest, so
the PR body records it for the conducting session; deleting or editing anything outside this pass's worktree. The main
checkout's untracked files (`reports/`, `research_notes/`, the 2026-09-04 research notes) are the conducting
session's; T0 only reads the three Pi 5 reports.

**End-to-end check.** On the close head:

1. `bash scripts/gate.sh --merge` prints `gate exit: 0` (the full gate plus the branch-history gitleaks scan).
2. The retired-phrase check (Conventions) prints nothing for `CLAUDE.md`, `.claude/skills/dubplate-pass/SKILL.md`,
   `docs/STATUS.md`, `ROADMAP.md`, and the proposed auto-memory index, and the lane check prints nothing for the
   first four.
3. `wc -l < docs/STATUS.md` prints 60 or less, and `grep '^## ' docs/STATUS.md` prints exactly `## Current state`,
   `## Next action`, and `## Open decisions`, in that order.
4. This prints `admin-api-1 intake admin-api-2 pipeline streamer queues migration builtins library-web-1 integration
   cutover closeout ` (the order derived below):

   ```bash
   sed -n '/^## Path to 0\.8/,/^## Active/p' ROADMAP.md | grep -o -E '^[0-9]+\. \*\*`[a-z0-9-]+' | grep -o -E '[a-z0-9-]+$' | tr '\n' ' '
   ```
5. The inventory check (T3) prints nothing and counts every row.

**Clock estimate:** 390 task-minutes (T0 45, T1 waived, T2 45, T3 55, T4 45, T5 40, T6 30, T7 60, close 70). T2 and T3,
T4 and T6, and T5 and T7 run as pairs, so the path is about 45 + 55 + 45 + 60 + 70 = 275 minutes, plus each
pair's post-merge fast lane. Every estimate that holds a gate allows for one: recorded dubplate gates ran 10 to 19
minutes, plus up to 11 minutes of heavy-lane lock wait.

## Path to 0.8: the derived run order

T2 writes this list into ROADMAP. It comes from the hard and soft edges in
`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md` ("The matrix", "The hard edges", "The soft
edges"), the reverse-edge note `docs/superpowers/research/2026-09-04-reverse-edge-completion.md`, the re-plan
(`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, "The ladder after this spec"), the rung 7 and 14a
specs, and ruling 2.

Landed: `foundation` (rungs 1 and 1b), `corpus` (2a to 2d), `matcher` (3a to 3d), `shadow` (4a and 4b; its program was
deleted on 2026-10-05), `auth` (5a to 5d), `store` (6a to 6d), `tag-io` (8a to 8c), and `streamer-1` to `streamer-3`
(slices 1, 2a, and 2b).

1. **`admin-api-1`** (14a). Its hard inbound edges are `auth` and `store` (matrix rows 5 and 6, column 14), both
   landed. The hard edge from `pipeline`, (14,9) (`reverse-edge-completion.md:37`; the matrix itself marks it soft),
   was discharged by Geoff's per-album lease ruling (14a spec, owner ruling 1). STATUS names it as the next pass.
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
`worktree-alignment-plan` (the approved design, the audit, the planning commits, and the review and fold records).
The executing session continues that branch: its first act after `EnterWorktree` is `git merge --no-edit
worktree-alignment-plan`, so its PR carries the design, the plan, and the work together.

**Who launches.** The planning session is worktree-isolated, and its Bash guard refuses `git -C` into the main
checkout, so it cannot run this block. After its "plan ready to launch" notification, the conducting session (not
isolated) or Geoff runs it from the main checkout, after every check prints what its comment names:

```bash
git -C ~/Projects/dubplate fetch origin
git -C ~/Projects/dubplate branch --show-current                  # must print master
git -C ~/Projects/dubplate rev-list --count HEAD..origin/master   # must print 0 (not behind)
git -C ~/Projects/dubplate rev-list --count origin/master..HEAD   # must print 0 (not ahead)
git -C ~/Projects/dubplate status --porcelain --untracked-files=no   # must print nothing
git -C ~/Projects/dubplate/.claude/worktrees/alignment-plan status --porcelain   # must print nothing (skip if that worktree was removed)
git -C ~/Projects/dubplate rev-parse worktree-alignment-plan origin/worktree-alignment-plan   # two equal lines
claude agents --json | jq -r '.[] | select((.cwd | startswith("/var/home/glw907/Projects/dubplate/.claude/worktrees/dubplate-alignment")) or .name == "dubplate-alignment") | .name'   # must print nothing
cd ~/Projects/dubplate && claude --bg --model sonnet --effort medium --name dubplate-alignment \
  "$(git -C ~/Projects/dubplate show worktree-alignment-plan:docs/superpowers/plans/2026-10-11-dubplate-alignment-prompt.txt)"
```

The prompt is read with `git show` because the plan is not on `master` until this pass's PR merges. The ahead check
matters because `worktree.baseRef` is `"head"` in `~/.claude/settings.json`, so an unpushed `master` commit would ride
into the PR.

## Conventions every task follows

- **Gate.** `bash scripts/gate.sh` (the full gate) and `bash scripts/gate.sh --scope-changed --base <base>` (the fast
  lane), from the worktree root, with the Bash tool's `timeout: 600000`, re-issued on exit 75 until it prints `gate
  exit:`. T0 adds the wrapper. A worktree-isolated session's Bash guard refuses `cairn-run-gate 'bash
  scripts/check.sh'` ("what it runs cannot be shown not to be git"), and accepts a plain `bash <script> <flags>` call
  whose script execs `cairn-run-gate` (probed 2026-10-11 from the planning worktree). Until T0 lands, the only gate
  the session runs is T0's own acceptance. Never fall back to a bare `bash scripts/check.sh` for an acceptance.
- **The base, and why the fast lane names it.** `scripts/gate.sh` matches no row of `check.sh`'s class table, so it
  is `unmatched` (`check.sh:399`), and an unmatched path promotes a scoped run to the full form
  (`check.sh:458-459`: `fallback_reason="$f is $class"`). Without `--base`, a branch's change set runs from the merge
  base with master (`check.sh:307`), so on this branch every scoped run after T0 would include `scripts/gate.sh` and
  run the full gate. Every fast lane on this branch therefore passes `--base <base>`, where `<base>` is the literal
  40-hex id of the commit the task started from: the session's `HEAD` at task start, written into the clock file as a
  fourth line `base=<sha>` (the clock hook reads only `task=`, `start=`, and `estimate=` with `sed -n`, so the extra
  line is inert), or, for a pair, the dispatch commit named in the dispatch text. `check.sh` uses a supplied base only
  when it is a full hex id that resolves to itself and is an ancestor of `HEAD`, and otherwise falls back to the full
  form (`check.sh:268-294`), so a wrong value costs time, never coverage. The rule is local to this branch: once
  `scripts/gate.sh` is on `master`, a later branch's merge base excludes it, and the adapter keeps "Pass no `--base`".
- **Docs-only pair halves.** `check.sh` classes `docs/*`, `.claude/*`, `ROADMAP.md`, `BACKLOG.md`, `IDEAS.md`, and
  `CLAUDE.md` as `noeffect`, "gitleaks only" (`check.sh:395-397`; header `:109-110`). A pair half whose files are all
  in that class (T2, T3, T4, T6) skips `npm ci` and its own gate; its greps still run, and the session's post-merge
  fast lane is the gate for both halves. T5 and T7 touch `web/`, Go, and ops paths, so each runs its own setup and
  fast lane.
- **Setup.** `npm ci --prefix web` in the session's worktree and in T5's and T7's, before their first gate (the gate
  refuses a worktree without `web/node_modules`).
- **Retired-phrase check.** For each file, `test -s <file>` succeeds, then `tr '\n' ' ' < <file> | tr -s ' ' | grep
  -o -i -E 'lane registry|pass-execute|conducting loop|conductor ruling|errata owed|workflow mode|post-mortem'` prints
  nothing. Joining lines catches a phrase that wraps (today it finds the adapter's "lane registry", split across
  `SKILL.md:18-19`, which a per-line grep misses).
- **Lane check.** For each file, this prints nothing (today it prints "Document lanes", "photo lane", "dubplate
  lanes", and others; a probe shows it keeps "heavy-lane" and "fast lane" and still prints "every lane"):

  ```bash
  tr '\n' ' ' < <file> | tr -s ' ' | grep -o -i -E '\b([[:alnum:]]+[ -])?lanes?\b' | grep -v -i -E '^(heavy|light|fast)[ -]lanes?$'
  ```
- **Diff checks.** Every diff check reads `git diff -U0 <base>..HEAD -- <paths>` with the literal base, and first
  asserts the diff is non-empty (`... | grep -c '^@@'` prints 1 or more). The comment-only filter, for T5 and T7, is
  this pipeline, which must print nothing:

  ```bash
  git diff -U0 <base>..HEAD -- <paths> | awk '/^diff --git / { f = $NF; sub(/^b\//, "", f); hdr = 1; next }
  /^@@/ { hdr = 0; next }
  hdr { next }
  /^[+-]/ { b = substr($0, 2)
    if (b ~ /^[ \t]*$/) next
    if (b ~ /\/\/go:|\/\/line |nolint|@ts-|eslint-|#[ \t]*shellcheck|#[ \t]*noqa/) { print f ": directive: " $0; next }
    if (f ~ /\.go$/ && (b ~ /^[ \t]*\/\// || b ~ /^[ \t]*\/\*([^*]|\*+[^*\/])*\*+\/[ \t]*$/)) next
    if (f ~ /\.ts$/ && b ~ /^[ \t]*\/\//) next
    if (f ~ /\.css$/ && b ~ /^[ \t]*\/\*([^*]|\*+[^*\/])*\*+\/[ \t]*$/) next
    if (f ~ /\.(sh|py)$/ && b ~ /^[ \t]*#/ && b !~ /^#!/) next
    print f ": " $0 }'
  ```

  It allows Go `//` and a line that is exactly one `/* */` comment, TS `//` only (a TSDoc `/** */` block is flagged,
  so a TS comment is `//`), CSS exactly-one `/* */` only, and shell and Python `#` (not `#!`); it rejects directive
  comments and skips header lines only before a file's first `@@`. Probed 2026-10-11 on two scratch diffs: it flagged
  all 14 planted lines of the first (`*p = v`, a removed `return nil`, `//go:build`, `// @ts-ignore`, a hunk line
  `---flag`, a CSS `* {` and `//` line, `# shellcheck`, `#!`, `local x=1`, `-++counter`, `+--counter`, `# noqa`, a
  Python call) and passed its 6 plain comments; on the verification read's diff it flags code wrapped in two block
  comments (`/* x */ y := 2 /* z */`, `/* a */ --color-primary: red; /* b */`), which the earlier `.*` form passed. A `#` line inside a shell heredoc would pass; `scripts/provision.sh`
  has no heredoc (`grep -n -E '<<-?[[:space:]]*["'\'']?[A-Za-z_]' scripts/provision.sh` prints nothing; its `<<<` are
  here-strings), and T7's acceptance reruns that grep.
- **The main checkout stays untouched.** Every command runs from the task's own worktree. T0 reads three files in
  `~/Projects/dubplate/reports/` and writes nothing there. No task writes, moves, or deletes a file outside its
  worktree, and that includes the auto-memory directory
  `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/`, which T4 only reads. The session writes no
  auto-memory of its own.
- **Names.** The rungs are `foundation`, `corpus`, `matcher`, `shadow`, `auth`, `store`, `intake`, `tag-io`,
  `pipeline`, `migration`, `builtins`, `streamer`, `queues`, `admin-api`, `integration`, `cutover`, and `closeout`
  (ruling 5), for rungs 1 to 17 in order. A pass is `<name>-<n>`: `admin-api-1` is 14a, `admin-api-2` is 14b,
  `intake-1` to `intake-3` are 7a to 7c, and `streamer-1` to `streamer-3` are slices 1, 2a, and 2b. A landed pass with
  a letter id (5c, 6d) keeps that id in git and in frozen documents; a live document that cites one writes the rung
  name and what the pass did ("the config registry pass", not "rung 1b"). Frozen specs keep their old numbers, and no
  errata sweep follows. Only ROADMAP's "Path to 0.8" uses a number for order. A rung-number grep is
  case-insensitive and plural-aware: `grep -n -i -E '\brungs? [0-9]'`.
- **Voice.** Prose follows the `writing-voice` output style, with Go comments under `go-conventions`, shell comments
  plain, and no em dash in any comment. No new process citation (rung, task, ruling, or plan line) enters a shipped Go
  comment; the gate's process-citation leg only warns (`CHECK_CITATIONS=warn`, `check.sh:127-130`), so T7 greps for
  one itself.
- **Commits.** One commit per task, subject starting with the task id (`T2: Add the Path to 0.8 and the rung names to
  ROADMAP`), imperative mood, specific files only (`git add <path>...`, never `-A`), with the trailer the session's
  system reminder names. The session merges a pair half with `git merge --no-edit <branch>`.
- **Records.** The PR notes live at `<git dir>/pr-notes.md`, where `<git dir>` is `git rev-parse --absolute-git-dir`
  in the session's worktree: never committed, and it survives a session restart, unlike `$CLAUDE_JOB_DIR`. Its first
  line, written right after the planning-branch merge, is `launch-base=<sha>` (that `HEAD`), which Resume reads. A
  write into the worktree git dir with one plain command has precedent under this guard: the isolated cairn sessions
  `lean-cutover` and `engine-pass-b` wrote `pass-task-clock` there with a single `printf ... > <path>`, and T0's clock
  write is this pass's first live proof (a refusal there is a stop before any work). Every task
  that owes the PR body a record (T2's removal map, T3's map and dispositions, T4's dry run, T5's
  triage, T7's lessons table) gets it there. A pair subagent writes nothing outside its worktree: it returns its
  records in its final report, with its worktree path and branch, and the session appends them. The session writes
  with the Write tool or one plain command per write, since the guard refuses a compound command that writes outside
  the worktree. T4's proposed auto-memory index goes to `<git dir>/memory-index.proposed.md`, written by the session
  from T4's report.
- **Clock.** At each task start the session writes `pass-task-clock` in its worktree's git dir with `task=<id>`,
  `start=<epoch>`, `estimate=<minutes>`, and `base=<sha>`, and removes it at the task's end. For a pair it writes one
  file (`task=T2+T3`, the larger estimate, the dispatch commit as `base`) before dispatch, checks the overrun when both
  return, then removes it. Pair worktrees hold no clock file.
- **Pairs.** Each pair is two Agent-tool subagents, `subagent_type: "general-purpose"`, `isolation: "worktree"`,
  model `sonnet`, effort `medium` (T3 runs on `opus` at `medium`), dispatched after the session commits, so
  `worktree.baseRef: "head"` branches them from that commit. The dispatch names that commit's full id; the
  subagent's first act is `git merge-base --is-ancestor <dispatch sha> HEAD`, and a failure is a stop. Each gets its
  task's full text and this section verbatim, commits, and never pushes. T7's dispatch names `go-conventions`,
  `ts-conventions`, and `vps-conventions`; T5's names the visual identity spec and `python-conventions`. The session
  merges both halves, runs the fast lane with `--base <dispatch sha>`, and pushes.
- **Stops.** A red the session cannot fix in two attempts, a merge conflict, a clock stop, an auto-mode classifier
  refusal, or a finding that would change the approved design: write `docs/STATUS.md` on the branch (task in flight,
  what failed, one question with a recommendation), commit and push it, run `notify-send -u normal "dubplate
  alignment: stopped"`, and wait. Never work around a refusal with another tool. A pair subagent whose gate is not
  green after two attempts returns its work uncommitted, and the session applies this rule before merging either
  half.
- **Resume.** After a crash, an API drop, or a restart, the session (the same one, or a fresh one launched with the
  same prompt, which enters the existing worktree with EnterWorktree's `path`) does not redo work. It first rewrites
  the clock file's `start=` to now (pass-core: "A resume rewrites `start=`"), since a surviving file would trip the
  clock hook on the first tool call. It reads this pass's task commits with `git fetch origin`, then `git log
  --no-merges --format='%h %s' <launch base>..HEAD ^origin/master`, where `<launch base>` is the `HEAD` the session
  recorded right after merging the planning branch (the first line of the PR notes, `launch-base=<sha>`); the
  `^origin/master` term drops another pass's `T<n>:` commits that the Close's merge of `origin/master` brings in. For
  pairs it lists the worktrees from its own tree (`git worktree list`, which needs no `-C`) and, for each pair
  branch, `git log --format=%s <dispatch sha>..<branch>`: a branch with its `T<n>:` commit is merged; a branch with no
  `T<n>:` commit is a stop, since its worktree may hold warm uncommitted work the session cannot read (the guard
  refuses `git -C` into another tree; global CLAUDE.md: warm code is stop-and-investigate), unless the same session
  can resume that subagent with `SendMessage`. It then resumes at the first task whose acceptance fails, never
  re-applying a committed task. A merged task whose record is missing from the PR notes gets it rebuilt from its
  commit.
- **Scope.** A defect found outside a task goes to the PR body's "Found, not done" list.

## Tasks

### T0. The gate wrapper, the Pi 5 research, and the session record

- **Outcome.**
  - `scripts/gate.sh` runs `bash scripts/check.sh` through `cairn-run-gate` from the repo root. It accepts only
    `--scope-changed`, `--merge`, `--dry-run`, and `--base` followed by a value matching `^[0-9a-f]{40}$`, refuses
    anything else with exit 2 and a usage line (starting nothing), and builds the gate string from the validated words
    with no trailing space (`"bash scripts/check.sh${args:+ $args}"`). A plain argument forwarded through `$*` would
    reach `bash -c` as shell (`cairn-run-gate:335`: `exec bash -c "$2"`), which is the hole the guard closes. The
    default lane is heavy, since the full gate's rendered leg launches Chromium; with `--dry-run` the wrapper sets
    `CAIRN_GATE_LANE=light` (a dry run "runs no leg, takes no lock", `check.sh:75-76`; the light lane is for "a gate
    that launches no browser", `cairn-run-gate` header). The header comment says why the wrapper exists and that its
    exit codes are `cairn-run-gate`'s, so 75 still means re-issue. `~/.dotfiles/scripts/gate.sh` is the model.
  - The adapter names the wrapper in the Gate line (`SKILL.md:58-59`), the Fast lane line (`:61-64`: `bash
    scripts/gate.sh --scope-changed`; it keeps "Pass no `--base`"), the Full-suite home line (`:66-69`), and the Close
    checklist's `--merge` item (`:92-95`). CLAUDE.md's Gate section opens with `bash scripts/gate.sh` (`CLAUDE.md:70`)
    and its Shell bullet (`:89-92`) lists `gate.sh`. T6 rewrites the rest of the adapter; T0 edits only these lines.
  - `.gitignore` gains `docs/superpowers/research/2026-09-04-visual-collaborator-outreach.md`, with a one-line comment:
    the file holds personal email addresses and stays uncommitted (Geoff's ruling, recorded only in the auto-memory
    file `unattended-window-2026-09-07.md`, which T4 classes for deletion). Today only the main checkout's local
    `.git/info/exclude:18` keeps it out. It lands here rather than in T4 because `.gitignore` is gate machinery
    (`check.sh:367`), and T0's gate is already the full form.
  - Geoff's direction (2026-10-11, relayed by the conducting session, to which the design's Out of scope gives
    `reports/`): the three Pi 5 stereo client reports written 2026-10-10 are committed, byte for
    byte, as live input to the 0.9 local client, in `docs/superpowers/research/2026-10-10-pi5-client/`:
    `proven-build.md` (from "Pi 5 stereo client proven build.md", the current decision), `verified-parts.md` (from "Pi
    5 stereo client verified parts.md"), and `hardware.md` (from "Pi 5 stereo client hardware.md"). The planning
    session read the lines a personal-data scan flagged: "EEPROM address" (`verified parts.md:49`) and "street
    pricing" (`hardware.md:7`, `:19`), with no personal data. ROADMAP's `local-client-0-9` entry
    (`ROADMAP.md:117-142`) gains one line pointing to `proven-build.md`. If the auto-mode classifier refuses the copy,
    that is a stop.
  - STATUS gains one line under "Next action" recording this session's name (`dubplate-alignment`) and id
    (`$CLAUDE_CODE_SESSION_ID`) as in flight, visible on this branch only; T3 carries it into the rewrite.
- **Files.** `scripts/gate.sh` (new, executable), `.gitignore`, `.claude/skills/dubplate-pass/SKILL.md` (the four
  lines above), `CLAUDE.md` (`:70`, `:89-92`), `docs/STATUS.md`, `ROADMAP.md` (one line), and the three new research
  files.
- **Risk class.** `runner`.
- **Acceptance.**
  - `bash scripts/gate.sh --bogus` exits 2 and prints no `gate exit:`; so does `bash scripts/gate.sh --scope-changed
    --base abc123`.
  - `bash scripts/gate.sh --scope-changed --dry-run` prints `check.sh: scope=full fallback="<path> is <class>"` naming
    a T0 path (`.gitignore is machinery` or `scripts/gate.sh is unmatched`) and `gate exit: 0`. That fallback is the
    expected result, not a failure.
  - `bash scripts/gate.sh --merge --dry-run` prints `check.sh: scope=full+history` (`check.sh:537`) and `gate exit:
    0`.
  - `bash scripts/gate.sh` (the full form, which also runs shellcheck over the wrapper) prints `gate exit: 0`.
  - `test -x scripts/gate.sh`; `git check-ignore -v docs/superpowers/research/2026-09-04-visual-collaborator-outreach.md`
    prints a line starting `.gitignore:` (a plain `-q` would pass vacuously: the worktree shares the main repo's
    `info/exclude`, which already matches it); `cmp` of each Pi 5 copy against its original under
    `/var/home/glw907/Projects/dubplate/reports/` (absolute paths, quoted) exits 0; and `stat -c %Y` on the three
    originals, run once before the copy (the values go into the PR notes) and once after, prints the same three values.
- **Test.** The acceptance runs: the refusals prove the allowlist, and the full run proves the wrapper reaches
  `cairn-run-gate` past the guard.
- **Estimate.** 45 minutes (one full gate plus lock wait). The session runs it inline, first.

### T1. The dependency sweep (waived)

Geoff waived the adapter's per-pass sweep for this pass on 2026-10-11, in the conducting session: it changes
only docs, skills, memory, and comments, and `admin-api-1`, the next code pass, runs the sweep. No task here
touches `go.mod`, `go.sum`, `web/package*.json`, or `tools/venv/requirements.txt`. T3 writes the held TypeScript 7
major (inventory row I66) as its own BACKLOG item, with the text "TypeScript 7" and `#rung:integration`.

### T2. ROADMAP: the Path to 0.8, the names, and a lean shape

- **Outcome.**
  - A new section, "Path to 0.8", directly under the header block: the twelve numbered items of this plan's derived
    order, each as `N. **\`<name>\`**` with the old number in parentheses, one line of what it delivers, and its
    inbound edges in a short clause. Below the list, one sentence: items whose edges are met may run at once, up to
    three passes (ruling 1); the numbers set start order and are rewritten when the order changes. Each item names
    where its carried work lives: BACKLOG items tagged `#rung:<name>` (T3 writes them); for `intake`, the rung 7 spec's
    "Owed records" (T6 reframes it); for `admin-api-1`, its planning prompt (T3 writes it).
  - A mapping table under it: old rung number, name, landed passes (old ids), and the new pass ids, covering all 17
    rungs plus `library-web-1`. One line above it says frozen documents keep the old numbers and this table bridges
    them.
  - Active: the greenfield entry (`ROADMAP.md:14-65`) and the rung 12 entry (`:67-113`) are trimmed to goal and next
    step: the goal, the launch bar in one sentence, "next: `admin-api-1`", and a pointer to "Path to 0.8". Landed
    status leaves (git and the mapping table carry it). The standing 0.8 rule at `:97-101` goes; its home is
    `docs/standing-rules.md` "Build for 0.8", which CLAUDE.md points to. The TypeScript 7 recheck (`:111-113`) leaves
    ROADMAP; T3 writes it into BACKLOG from its own base.
  - **The removal map.** Before deleting any sentence of `:14-113`, `:214-215`, `:221-225`, or `:292-306`, T2 finds
    its home: a spec or doc line found by a `grep` of a distinctive phrase (file and line), a BACKLOG or IDEAS item
    (name it), or "kept" (it stays in ROADMAP, shortened). A sentence with no home is kept. Geoff's 2026-09-24 web
    player requirement (`:104-109`: "a strong play-queue interface ... now playing, up next, reorder, and add from
    album, track, or filter results", designed from a prior-art study in the visual identity spec's language) has no
    other home (`grep -rn -i 'up next' docs IDEAS.md` finds no requirement), so it stays as one line in the streamer
    entry, containing "up next" in lower case (row I70's check is case-sensitive). The report carries the map, one row per removed sentence or range.
  - Planned: `pr-test-ci` gains "after `admin-api-1`" (ruling 3). `prod-polish` loses the owed post-mortem
    (`:214-215`). The folded organizer entry (`:221-225`) goes. `:276`'s "Two owed dubplate lanes" is reworded without
    "lanes". Other Planned entries keep their substance (the `give-back` entry's lawyer, Unidecode, history, and holder
    items included; T3's inventory checks them); rung numbers become names.
  - The Done section (`:292-306`) goes.
- **Files.** `ROADMAP.md`.
- **Risk class.** `ordinary`.
- **Acceptance.** The retired-phrase and lane checks print nothing for ROADMAP. End-to-end check 4 prints the
  expected order. `grep -n '^## ' ROADMAP.md` prints, in order, the Path section, Active, Planned, and Someday, and no
  Done. `grep -n -i -E '\brungs? [0-9]' ROADMAP.md` matches only the mapping table and the Path items' parenthesized
  old numbers. `grep -c 'up next' ROADMAP.md` prints 1 or more. T0's Pi 5 pointer survives (`grep -c
  'pi5-client/proven-build.md' ROADMAP.md` prints 1). No gate in the half (docs only).
- **Test.** A read of the Path list against this plan's derived order, item by item, and the removal map.
- **Pair.** Runs as a pair with T3 (disjoint files).
- **Estimate.** 45 minutes.

### T3. STATUS, BACKLOG, the dispositions, and the admin-api-1 planning prompt

- **Outcome.**
  - `docs/STATUS.md` rewritten to pass-core's three parts, present tense, 60 lines or fewer:
    - Current state: master's last green full gate; nothing deploys; the box holds dubplate-server `d1e93e9` with the
      serve unit installed and disabled, the `auth`, `roster`, and `unidecode` leaves, one admin member, and
      `/srv/music/library` (6.6 GB) (`STATUS.md:9-10`); this pass in flight (T0's line, visible on the branch only).
    - Next action: the ordered post-merge list, run by whoever merges this pass's PR (the conducting session or
      Geoff):
      1. `git -C ~/Projects/dubplate pull --ff-only`, under pass-core's conditions (the checkout is on `master`, has
         no uncommitted change in a file the merge touches, and has no live session per `claude agents --json`);
         otherwise report the checkout stale and stop.
      2. The auto-memory prune: `tar czf ~/.local/share/dubplate/memory-backup-2026-10-11.tar.gz -C
         ~/.claude/projects/-var-home-glw907-Projects-dubplate memory`; re-list the directory (`ls -l
         --time-style=+%s`) and keep any file modified after T4's start epoch (the PR body names it); write the new
         `MEMORY.md` from the PR body's proposed index; show the delete list from the PR body's dry-run table, then
         delete those files; re-list, and run the retired-phrase check over the new `MEMORY.md`.
      3. Trash the three Pi 5 originals: `gio trash` on each file in `~/Projects/dubplate/reports/` that T0 copied.
      4. Launch the `admin-api-1` re-plan, an unattended Opus planning session from the 14a spec (revision 2.1,
         approved by Geoff 2026-10-10, its "Rulings for Geoff" section "None"): `cd ~/Projects/dubplate && claude --bg
         --model opus --effort high --name admin-api-1-plan "$(cat
         docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt)"`. Its first STATUS edit replaces this
         list with its own in-flight line.
    - Open decisions: R-6d21 (`STATUS.md:22`), the group non-FLAC question (`:23`), the stale `DUBPLATE_*` exports in
      `~/.local/secrets` (`:24`), and the one Run-D.M.C. frame (`commons:139197098`) with the five maybes in
      `photos/.candidates/` (`:31`, and the auto-memory `photo-round-one-closed.md`). The rung 7 spec question (`:21`)
      leaves (T6 replaces that gate).
    - The lane registry (`:26-31`), the Owed list (`:33-40`), and the open items (`:42-60`) leave.
  - **The inventory.** Every live item reachable from `STATUS.md:5-60`, with every HISTORY pointer followed to its
    leaf, is one row below. The planning session traced the chain: `STATUS.md:60` to `HISTORY.md:588`, whose carries
    at `:612-627` point on to `:697-708` ("re-plan close, Moved out of STATUS"), to `:1316-1336` (transliteration,
    "Routed, not decided here"), and to `:1557-1580` (5c-iv, "To rung 7"); `STATUS.md:56` to `HISTORY.md:443-464`
    (6d, "Routed items"); `STATUS.md:57`'s cookie test to `HISTORY.md:802-803`. T3 writes each row whose file is `B`
    into BACKLOG as an item in its own words, tagged `#rung:<name>` for the rung that consumes it (a rung-free chore
    keeps its existing tags; a streamer item names its slice, "streamer slice 6"), and containing the row's token
    verbatim. Related rows may share one item. An item already in BACKLOG is tagged, not duplicated. Rows whose file
    is `S` stay in STATUS. Rows whose file is `R`, `I`, or `7` are already carried there, and the check confirms they
    survive. T3 runs on `opus`.

    Files: `B` is `BACKLOG.md`, `S` is `docs/STATUS.md`, `R` is `ROADMAP.md`, `I` is `IDEAS.md`, `7` is
    `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`.

    | Row | Token | File | Item (source) |
    |---|---|---|---|
    | I1 | `R-6d21` | S | R-6d21 overrides 6c's "never reads environ" rule; Geoff may reopen it (STATUS:22) |
    | I2 | `non-FLAC` | S | the group non-FLAC question (STATUS:23) |
    | I3 | `DUBPLATE_` | S | about 41 stale exports in `~/.local/secrets` fail `check` outside `env -i` (STATUS:24) |
    | I4 | `photos/.candidates` | S | the Run-D.M.C. frame and five maybes await Geoff (STATUS:31; photo memory) |
    | I5 | `d1e93e9` | S | the box's state (STATUS:9-10) |
    | I6 | `6d upgrade` | B | the first deploy takes the runbook's 6d upgrade, `#rung:cutover` (STATUS:10-11) |
    | I7 | `machine-form ledger` | B | slice 3's inputs: 2b's XML contract tier, the machine-form ledger (52 owed rows), the per-endpoint assignment, the differ, the music-folders evidence (STATUS:6-7) |
    | I8 | `-gpu host` | B | `docs/roster-walk.md`: Tempus pin 4.28.0 to 4.29.0, F11 uses `createPlaylist(name, songId)`, and the emulator needs `-gpu host`, against the doc's `--disable-gpu` (`roster-walk.md:35`); edited by `streamer-4` (STATUS:7-8, 37-38) |
    | I9 | `Amperfy` | B | Amperfy's roster walk is `cutover`'s attended gate (STATUS:8, 51) |
    | I10 | `make derive` | B | run once before the next oracle-dependent pass, expect no diff; tag the first Path rung that touches the matcher or beets oracle (expected `#rung:pipeline`) (STATUS:39-40) |
    | I11 | `mid3v2` | B | the `id3-tags` copy names a working strip method (`mid3v2 --delete-all` or a tagger), `#rung:intake` (STATUS:38-39) |
    | I12 | `EpochStore.Bump` | B | the revoke verb and streamer slice 6: criteria 7 and 8, the dual-key epoch (R-6d20), the rung 5 review's minor 14, erratum row 10; `EpochStore.Bump` has no caller (STATUS:43; HISTORY:445-447) |
    | I13 | `last-admin guard` | B | the last-admin guard and criterion 6's concurrent self-demotion clause, with the removal verb (CH:576-583, :732), `#rung:admin-api` (STATUS:44; HISTORY:448-449) |
    | I14 | `tombstones` | B | the removal verb bumps the address's revocation epoch as it tombstones (STATUS:44-45) |
    | I15 | `UserID` | B | the network actor keyed on `UserID` (STATUS:45) |
    | I16 | `resolving channel` | B | the admin write path's resolving channel for a `serve`-refusing `admin` row (STATUS:45-46; HISTORY:616) |
    | I17 | `R-6d4` | B | the first migration after `0005`: R-6d4's keep limit (STATUS:46-47; HISTORY:453-455) |
    | I18 | `table map` | B | the older-store test helper's table map, same migration (STATUS:47; HISTORY:455-456) |
    | I19 | `enabling the serve unit` | B | `#rung:cutover` (STATUS:48) |
    | I20 | `runuser` | B | criterion 14's `runuser` and holder clauses with 6c's Ruling 2 (STATUS:48-49) |
    | I21 | `secrets in every` | B | secrets in every `music`-uid verb's environment (STATUS:49) |
    | I22 | `SMTP from the box` | B | (STATUS:49) |
    | I23 | `systemd-run` | B | the sandboxed `member add` and `doctor` the 6d smoke did not drive (STATUS:50) |
    | I24 | `runbook lines` | B | the remaining runbook lines (STATUS:50-51) |
    | I25 | `second copy` | B | the archive's second copy (STATUS:51) |
    | I26 | `per-collection relocation` | B | (STATUS:51) |
    | I27 | `store identity` | B | R9's option b (STATUS:51) |
    | I28 | `UNI-H2` | B | slice 3 (STATUS:52) |
    | I29 | `percent-encoding` | B | slice 3: the prefix-gap and percent-encoding items (STATUS:52) |
    | I30 | `L-51` | B | slice 3: the owed ledger rows L-20, L-21, and L-51 (STATUS:52) |
    | I31 | `L-17` | B | slice 5 (STATUS:53) |
    | I32 | `streamer-admin mapping` | B | slice 6 (STATUS:53) |
    | I33 | `duplicate_principal` | B | slice 6: the code, with BACKLOG #16 (STATUS:53; HISTORY:460) |
    | I34 | `offboarding` | B | slice 6: offboarding deprovision (STATUS:54) |
    | I35 | `shared-origin M2` | B | slice 6: CSP scope and shared-origin M2 (STATUS:54) |
    | I36 | `username uniqueness` | B | slice 6 (STATUS:54) |
    | I37 | `streamer database` | B | slice 7: the streamer database in a restore (STATUS:54-55) |
    | I38 | `group 2 in trailers` | B | slice 8 (STATUS:55) |
    | I39 | `CAP_NET_BIND_SERVICE` | B | slice 8 (STATUS:55) |
    | I40 | `command naming` | B | slice 8 (STATUS:55) |
    | I41 | `R-6d17` | B | the next `web/` pass: the roster line's blank-username conditionals, the unreachable `duplicate_principal` mapping and copy, `rendered/run.ts`'s comments; with BACKLOG #23 (STATUS:56; HISTORY:461-463) |
    | I42 | `cookie test` | B | the bearer bypass's first consumer owes a shell-level test that a bypassed route never reads the session cookie (STATUS:57; HISTORY:802-803) |
    | I43 | `upgrade naming` | B | settle `init`/`upgrade` naming against `dubplate-server` (STATUS:58-59) |
    | I44 | `configuration-reference generator` | B | a hook 18 consumer (STATUS:59) |
    | I45 | `trigram` | B | `pipeline`: filter, facets, trigram (HISTORY:616) |
    | I46 | `btn-rest-fg` | B | 5c-iv's "To rung 7" list for `intake-3`: Ruling 5's four items, W12.4, W4.10's 44px promotion, placeholder opacity, the outline, dash, and status-ghost text fix (`--btn-rest-fg`), 5c-iii's three latent items, the favicon paths, the toggle under `system`, the masked "Sign in as" line, the phone gutter `clamp()` (HISTORY:1557-1580; the rung 7 spec's 7c cites it at a stale line, `:898`) |
    | I47 | `stable paths` | B | `intake`: 5c-iii's S13 and a11y S7, picked images at stable paths (HISTORY:613) |
    | I48 | `--max-delete` | 7 | the backup guard's counting and `--max-delete` (rung 7 spec, "Moved out of rung 7") |
    | I49 | `## The reject set` | 7 | the reject-reason set |
    | I50 | `CREDITS` | 7 | the CREDITS generator (7c) |
    | I51 | `member_quota` | 7 | 6c's `member_quota` reservations (HISTORY:614, :1926-1930) |
    | I52 | `twelve-flag` | B | the architecture cleanup: the routed findings in `research/*-architecture-reads.md` (5d-ii's doctor twelve-flag OR chain first), the translit ROUTE items, the 6d "To file" list, BACKLOG #24 to #26, 5c-iv 9b's five stale-comment nits, the roster outcome classes with no oracle cross-check, the login title and W6 h1 fallback with no rendered proof (HISTORY:624-625, :1333-1335, :464, :1393, :1582-1586) |
    | I53 | `103 coverage gaps` | B | none blocking (HISTORY:625) |
    | I54 | `check.test.mjs` | B | the photo check fails 2 of 121 (both assume an empty `photos/`), so no gate runs it (HISTORY:701-702) |
    | I55 | `from disk` | B | the photo integrity leg reads `photos/` from disk, not `git ls-files`, and `check.sh`'s header omits the leg (HISTORY:702-703) |
    | I56 | `photo column` | 7 | Ruling 3's photo erratum (7c's photo column) |
    | I57 | `source-URL recheck` | B | photos (HISTORY:703) |
    | I58 | `blob-retention` | B | photos: a blob-retention call (HISTORY:704) |
    | I59 | `lawyer` | R | `give-back`: the Unidecode-derived items, the deleted table's history, the lawyer's confirmation (HISTORY:704-705, :1316-1332) |
    | I60 | `holder is process-wide` | R | `give-back`: the process-wide holder's tension (HISTORY:705) |
    | I61 | `UNI-L2` | B | `packaging`: UNI-L2's failure-alert hook (HISTORY:705-706) |
    | I62 | `Unauthenticated local mail relay` | I | the loopback-relay mail class (HISTORY:706) |
    | I63 | `URL fragment` | B | the token-in-fragment move, BACKLOG #13 (HISTORY:706) |
    | I64 | `R-6c` | B | the citation leg misses `R-6c` patterns (staged legs 317/127) (HISTORY:707) |
    | I65 | `nowPlaying` | B | the seed lacks lyrics, `nowPlaying`, and radio requests; their contract flags wait for a later recording, `#rung:streamer` (HISTORY:296-297, :311-312) |
    | I66 | `TypeScript 7` | B | before the 0.8 release, recheck TypeScript 7 (7.0.2 then; `sveltejs/language-tools` #2891), `#rung:integration` (`ROADMAP.md:111-113` at T3's base) |
    | I67 | `5.0.0-pre3` | B | BACKLOG #8's watch item names 5.0.0-pre3 (2026-09-30) (rung 7 spec, "Other records", `:1259`) |
    | I68 | `instance lease` | B | `serve`'s job runner, instance lease, and job recovery go to the first rung whose page starts a job over HTTP (14a spec `:638`; 14a fold "Conductor-owed") |
    | I69 | `C14` | B | C14's check lists each live member whose epoch record is later than the restored snapshot's stamp, `#rung:cutover` (14a fold `:91`) |
    | I70 | `up next` | R | Geoff's web-player play-queue requirement, kept by T2 (`ROADMAP.md:104-109`) |

    Retired, with no row: the lane registry and the lean-cutover next action; STATUS:18's settled forks (they live
    in the streamer spec's "Revision note, 2026-10-10"); STATUS:58's two `docker compose` entries (T6 removes them);
    the "To 6d" carry (landed: `applySettingsSeed` at `cmd/dubplate-server/serve.go:300` and `store.OpenServe` at
    `:391`); the archive-upload carry (the rung 7 spec retires it); the errata inventory (frozen documents get no
    errata); the resume prompts and lane rows HISTORY quotes.
  - Also into BACKLOG: BACKLOG #8's version fix (row I67) and BACKLOG #1 (the Go-rewrite post-mortem) closed with the
    note "closed under lean ruling 8: history is git and PR bodies".
  - A disposition line for each STATUS errata item, in T3's report for the PR notes:
    - Rung 7's eight documents (`STATUS.md:34-36`): SS, CH, W, VI, PCF, and CLAUDE.md are living references, so their
      edits stay in the rung 7 spec's "Owed records" as tasks of the `intake` passes (T6); 14A and BS are frozen, so
      those two are moot.
    - The 14a spec's eight owed errata (14a spec fold, "Owed errata" items 1 to 8): items 1 to 4 and 8 (R6), 6 (PA),
      and 7 (BS) are moot, frozen documents; item 5 (CH:171-172) and the PS member-cap note are living-reference edits
      that move into `admin-api-1` through its planning prompt. The plan fold's contested-surfaces erratum is moot (the
      tables retire).
    - `docs/roster-walk.md`'s Tempus pin: row I8. The `id3-tags` copy: row I11 and the rung 7 spec's PCF item.
      `make derive`: row I10.
    - The 14a fold's open question on rung 7 and a streamer slice both writing `internal/serve`: moot; each design
      names its files, and with two PRs open the second merges master and reruns its close gate.
  - `docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt`, shaped like this pass's planning prompt
    (`docs/superpowers/plans/2026-10-11-dubplate-alignment-planning-prompt.txt`): EnterWorktree name
    `admin-api-1-plan`; load `dubplate-pass`; read the 14a spec in full as the approved design, the old plan
    `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md` as evidence only, and BACKLOG's `#rung:admin-api` items;
    write `docs/superpowers/plans/<date>-admin-api-1.md` in the lean format (this plan is the model, archived by its Close at `docs/superpowers/archive/plans/2026-10-11-dubplate-alignment.md`), class
    `auth-data`, with two living-reference edits as tasks: CH:171-172 (14a fold, "Owed errata" item 5) and the PS
    member-cap note with its erratum to PS:145-147, whose source is the old plan
    (`2026-10-10-rung-14a-admin-api.md:829-833`); the re-plan decides whether the member cap stays in scope. Run
    `spec-plan-review` with all three lenses, the fold, and the verification read; write the executor's launch prompt
    beside it; a finding that changes the approved design is the one stop (STATUS question, `notify-send -u normal`);
    record the session's name and id in STATUS, replacing the post-merge list; commit and push to `glw907/dubplate`,
    never `master`, never merge, never open a PR; notify `"admin-api-1: plan ready to launch"` and end.
- **Files.** `docs/STATUS.md`, `BACKLOG.md`, `docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt`
  (new).
- **Risk class.** `ordinary`.
- **Acceptance.**
  - The inventory check prints nothing, and its row count is 70:

    ```bash
    awk -F' \\| ' '/^ *\| I[0-9]+ \| /{ sub(/^ +/, ""); t = $2; gsub(/`/, "", t); print $3 "\t" t }' docs/superpowers/plans/2026-10-11-dubplate-alignment.md | while IFS="	" read -r k t; do case "$k" in B) p=BACKLOG.md ;; S) p=docs/STATUS.md ;; R) p=ROADMAP.md ;; I) p=IDEAS.md ;; 7) p=docs/superpowers/specs/2026-10-10-rung-7-intake-design.md ;; *) p=/nonexistent ;; esac; grep -q -F -- "$t" "$p" || echo "MISSING $k $t"; done
    grep -c -E '^ *\| I[0-9]+ \| ' docs/superpowers/plans/2026-10-11-dubplate-alignment.md
    ```

    (The `IFS` value is one literal tab; the table is indented, hence `^ *`.) Probed on the planning head: the count
    prints 70 and the check prints 56 `MISSING` lines, every one a `B` row or the `S` row I4 that T3 writes, so it can
    fail; the 14 rows it passes are the carried ones.
  - End-to-end check 3 holds; the retired-phrase and lane checks print nothing for STATUS. `grep -c 'HISTORY'
    docs/STATUS.md` prints 0, and `grep -c 'HISTORY' BACKLOG.md` prints 2 or less (today's count is 2).
  - `grep -c '#rung:' BACKLOG.md` prints a nonzero count (the inventory check, not a fixed floor, proves coverage, since related rows may share one item), and BACKLOG #1 is checked.
  - The prompt file names the 14a spec path, `auth-data`, the three lenses, `admin-api-1-plan`, `CH:171-172`,
    `PS:145-147`, `2026-10-10-rung-14a-admin-api.md:829`, and the push target `glw907/dubplate` (`grep -c -F` on each
    prints 1 or more).
  - No gate in the half (docs only).
- **Test.** The inventory check, and the side-by-side map in T3's report: each old STATUS line and each HISTORY leaf
  maps to a STATUS line, an inventory row, a disposition, or "retired".
- **Pair.** Runs as a pair with T2. After merging the pair, the session reruns the inventory check (it catches a
  ROADMAP-carried row T2 removed) and runs the fast lane, which must print `check.sh: scope=changed`: the first live
  proof that `--base` narrows.
- **Estimate.** 55 minutes.

### T4. CLAUDE.md, the implementer memory, and the auto-memory dry run

- **Outcome.**
  - CLAUDE.md (T0 has edited `:70` and `:89-92`, so T4 finds each section by content, not line):
    - The initiative paragraph uses the rung names, says "the ladder (ROADMAP, Path to 0.8)", and states the two
      standing bars outright instead of "from memory". Exemplary Go: idiomatic standard-library style and
      architectural beauty, never mere correctness; a reviewer may return `fix` on awkward-but-working code. Render and
      read every user-path page before any attended ask. Each carries its one-line why from its memory file
      (`idiomatic-beautiful-go.md`, `see-it-before-geoff-tests-it.md`).
    - Product rulings that live only in auto-memory become clauses where they execute in CLAUDE.md: "the group" and "a
      member", never "family", except in the product-essence line (`group-not-family.md`); plain terms in specs and UI
      copy, with no coined metaphor or record-shop pun (`plain-terms-no-coined-jargon.md`); the Svelte and DaisyUI idiom
      bar in the Web bullet: the stack's idiom wins over a finding's literal fix text, un-idiomatic code adjacent to a
      change is in scope, and the exemplar is `~/Projects/cairn-cms/src/lib/admin-toolkit/` and
      `src/lib/components/` (`idiomatic-svelte-daisyui.md`). Version numbering and the own visual identity are already
      in CLAUDE.md and need no clause.
    - The errata-avoidance prose (the Versions section's last sentences) becomes one sentence: documents older than a
      ruling keep their numbers and stay as written.
    - "Build for 0.8" names the `streamer` rung instead of "Rung 12", and keeps its pointer to
      `docs/standing-rules.md`.
    - Ledgers: HISTORY is "a frozen archive (2026-10-10); git and PR bodies are the record"; the ROADMAP line drops
      the T14 migration note and names "Path to 0.8".
    - Build process: the phase names (foundation, evidence, spine, surfaces, launch) as ROADMAP milestones, "Process:
      the `dubplate-pass` skill, loading `pass-core`", and the product touchpoints: corpus adjudication batches, the
      `intake` uploader sitting, and the first-launch sitting. The design approval and the PR read are pass-core's. No
      pointer to the build spec's Process section or the parallel-build spec remains.
    - The Go bullet gains the test-echoes-code clause: a test's expected value is a literal or an independently
      computed value, never built with the helper the code under test calls
      (`.claude/agent-memory/dubplate-implementer/feedback_test-echoes-code.md`).
    - Remaining rung numbers become names, per the Names convention.
    - Safety rails: before any executor touches a worktree, `claude agents --json` for a live session whose cwd is that
      worktree, then `pgrep -f` with a pattern whose literal text never appears in the command (a bracket class,
      `pgrep -f 'worktrees/<nam>[e]'`, since the Bash tool's wrapper carries the command text;
      `~/.claude/docs/bluefin-admin.md`), then `git status` for warm changes you did not author. STATUS records each
      running pass's session name and id. The box rules stay.
  - `.claude/agent-memory/dubplate-implementer/MEMORY.md` and `feedback_test-echoes-code.md` are removed with `git rm`
    (the agent was deleted at `8269d27`).
  - The auto-memory dry run. T4 reads every file in `~/.claude/projects/-var-home-glw907-Projects-dubplate/memory/`
    (41 files plus `MEMORY.md`) in full and classes each as one of: delete, retired process; delete, session or window
    log; delete, duplicate (naming the repo file and line that now carries it, confirmed by `grep`); extract, then
    delete (its last live fact moved into the repo this pass; name where); or keep, a preference or a ruling's why
    that no file holds. Each row also fills a "facts not carried" column: every ruling, requirement, or lesson in the
    file body (not only its index line) that no repo file now carries. A row whose column is not empty is classed keep,
    or its facts are extracted first: into CLAUDE.md (T4's file), into BACKLOG as a new item (T4 may append; T3's pair
    has merged), or named in "Found, not done" for a home T4 cannot write (the adapter is T6's). A "duplicate" whose
    `grep` finds nothing is classed keep, with the missing ruling named in its reason and in "Found, not done". The
    provisional classes, which T4 confirms or corrects with a reason:
    - Retired process: `greenfield-arc-sessions`, `pass-execute-cross-session-resume`,
      `pass-execute-dispatch-mechanics`, `two-lane-run-mechanics`, `ledger-fold-must-keep-list` (its lesson lands in
      the adapter through T6), `overnight-hold-sized-to-clock` (the inhibitor is tool-held now), and
      `spec-prose-review-rule` (the `spec-plan-review` skill now runs no prose review on a spec or plan).
    - Session or window logs: `unattended-window-2026-09-05`, `session-2026-09-07-4a-ii-close`,
      `session-2026-09-08-wind-down`, `session-2026-09-12-config-homes-close`, `unattended-window-2026-09-12`,
      `session-2026-09-12-infra-round` (its `pgrep -f` self-match lesson is carried by T4's Safety rails clause),
      `overnight-window-2026-09-23`, `session-2026-09-25-handoff`, `overnight-window-2026-09-25`,
      `unattended-window-2026-09-26`, `overnight-window-2026-09-26`.
    - Extract, then delete: `unattended-window-2026-09-07` (the outreach doc's privacy ruling, now in `.gitignore`
      through T0) and `photo-round-one-closed` (the five maybes, now an open decision in STATUS through T3).
    - Duplicates once T4 lands: `version-numbering`, `build-for-2-0`, `greenfield-server-pivot`, `own-visual-identity`,
      `group-not-family`, `plain-terms-no-coined-jargon`, `idiomatic-beautiful-go`, `idiomatic-svelte-daisyui`,
      `see-it-before-geoff-tests-it`; and, each only if `grep` finds its ruling in the named spec,
      `configuration-homes-rulings` (CH), `group-scale-principles-rulings` (PS), `rung-5-owner-rulings` (the rung 5
      spec), `fork-b-formats-ruling` and `give-back-go-libraries` (ROADMAP), `web-design-standards-sources` (the web
      surface standard).
    - Keep: `irreversible-means-bound-not-defer`, `no-calendar-schedule`, `no-ceremony-rulings`,
      `photo-older-is-better` (or, if T4 judges it a curation rule, the visual identity spec's "Curation rules",
      VI:277, named in "Found, not done" since VI is T5's), `process-framing-not-geoffs-rules`,
      `workstation-vs-project-scope-by-survey`.
  - T4's final report carries the dry-run table (file, class, reason, facts not carried) and the proposed index, the
    kept files' lines from the current `MEMORY.md` and nothing else; the session writes both (Records). T4 deletes
    nothing and writes nothing outside its worktree.
- **Files.** `CLAUDE.md`, `BACKLOG.md` (append only, if an extraction needs it),
  `.claude/agent-memory/dubplate-implementer/MEMORY.md` (removed),
  `.claude/agent-memory/dubplate-implementer/feedback_test-echoes-code.md` (removed).
- **Risk class.** `runner` (`.claude/`).
- **Acceptance.** The retired-phrase and lane checks print nothing for CLAUDE.md. This prints nothing:

  ```bash
  grep -n -i 'from memory\|parallel-build-approach\|at T14' CLAUDE.md
  ```

  `grep -n -i -E '\brungs? [0-9]' CLAUDE.md` matches only the Versions sentence about older documents. `git ls-files
  .claude/agent-memory/dubplate-implementer` prints nothing. `wc -l < CLAUDE.md` stays at or under 200. The session,
  after the pair returns: `test -s` on `<git dir>/memory-index.proposed.md`, then the retired-phrase check over it
  prints nothing; the dry-run table's row count equals the directory listing at T4's start (`ls
  ~/.claude/projects/-var-home-glw907-Projects-dubplate/memory | grep -v -x MEMORY.md | wc -l`); and `find
  ~/.claude/projects/-var-home-glw907-Projects-dubplate/memory -newermt "@<pair start epoch>"` is read, not required
  empty (probed: a recent epoch printed 0 lines, `@1000` printed all 43 entries). Other dubplate sessions share that
  directory, so each printed path goes into the PR notes as written during the pass and is kept by the post-merge
  re-list; it is a stop only if T4 classed that path delete. No gate in the half (docs only).
- **Test.** The row count against the directory listing, and the "facts not carried" column read.
- **Pair.** Runs as a pair with T6 (disjoint files).
- **Estimate.** 45 minutes.

### T5. The site-implementer memory to its executing homes

- **Outcome.** `.claude/agent-memory/site-implementer/` (11 tracked lessons plus `MEMORY.md`) goes, since the adapter
  stops naming `site-implementer` (T6) and no other agent loads it. Each lesson lands where it executes or is deleted
  with a one-line reason in the PR notes:
  - The six DaisyUI lessons (`daisyui-*.md`): a lesson about a theme token becomes a one-line `/* */` comment beside
    that token in `web/src/lib/themes/*.css` (`daisyui-semantic-theme-tokens`, `daisyui-soft-pastel-vs-saturated`,
    `daisyui-soft-reads-color-token-as-text`); a lesson about a component's markup becomes a sentence in the visual
    identity spec's "Component layer" section (`docs/superpowers/specs/2026-09-24-visual-identity-design.md:226`), a
    living reference (`daisyui-size-field-touch-targets`, `daisyui-alert-grid-single-wrapper`,
    `daisyui-horizontal-steps-baseline`). Where the theme files have no such token, the lesson goes to the spec.
  - `admin-identity-in-copy`: a sentence in the visual identity spec's "Copy" section (`:314`) if the display-name
    field still exists (it does today: `internal/auth/handlers.go:218`, `web/src/lib/api/auth.ts:26`); otherwise
    deleted as stale.
  - Deleted with reasons: `commit-only-named-paths` (pass-core and the global git policy carry it),
    `app-svelte-no-store-di` (`App.svelte` is salvage-era; `web/src` has none), `systemd-python-subprocess` (the box
    runs no beets), `beets-dry-run` (if the oracle shims under `internal/harness/oracle/*.py` already document their
    beets invocation; otherwise a `#` comment above that invocation, under `python-conventions`).
- **Files.** `.claude/agent-memory/site-implementer/*` (removed), `web/src/lib/themes/*.css` (comments only),
  `docs/superpowers/specs/2026-09-24-visual-identity-design.md`, and possibly one `internal/harness/oracle/*.py`
  comment.
- **Risk class.** `runner` (`.claude/`).
- **Acceptance.** `git ls-files .claude/agent-memory/site-implementer` prints nothing. If any CSS or Python file
  changed, the comment-only filter over `web/src/lib/themes internal/harness/oracle` prints nothing, after the
  non-empty check. `git diff -U0 <base>..HEAD | grep '^+' | grep -c $'\xe2\x80\x94'` prints 0 (no em dash). Each of the
  11 lessons has a row (home or deletion reason) in the report. The fast lane, `--base <dispatch sha>`, prints `gate
  exit: 0` (its `web/` leg rebuilds the bundle).
- **Test.** The comment-only filter.
- **Pair.** Runs as a pair with T7 (disjoint files).
- **Estimate.** 40 minutes.

### T6. The adapter, permissions, the frozen-spec headers, and the rung 7 gate

- **Outcome.**
  - `.claude/skills/dubplate-pass/SKILL.md`:
    - "One executor per tree" (`:18-22`) becomes the check T4 writes into CLAUDE.md's Safety rails: `claude agents
      --json`, then `pgrep -f` with a bracket-class pattern, then `git status` for warm changes; STATUS records each
      running pass's session name and id. The registry sentence goes. The box rules stay.
    - "Lanes" (`:29-31`) becomes "Concurrent passes": up to three dubplate passes at once, each its own session,
      worktree, and PR (Geoff, 2026-10-11, restating his 2026-10-05 three-code-lane cap). Each design names the files
      it touches. With two PRs open, the second to merge first merges master and reruns its close gate.
    - "Plans are just-in-time" (`:32-33`) speaks of passes, not lanes.
    - "Owner touchpoints" (`:43-44`) becomes the product touchpoints (corpus adjudication batches, the `intake`
      uploader sitting, the first-launch sitting), with the design approval and the PR read named as pass-core's.
    - "Pair subagents" (`:53-54`): `general-purpose`, given the adapter's setup and fast lane (`bash scripts/gate.sh
      --scope-changed`) and the `go-conventions` skill for Go (with "idiomatic, architecturally beautiful Go" named as
      an acceptance dimension) or `svelte-conventions`, `ts-conventions`, and the visual identity spec for `web/`.
    - A new rule, "Errata": frozen, ratified specs get no errata, and a superseded one gets at most a one-line header
      pointer. A living reference (the streamer spec's forward hooks, PCF, the visual identity spec, CH, W, PS,
      `docs/roster-walk.md`) is edited in place by the pass that changes what it says, as a task in that pass. There
      is never an owed list.
    - A new rule, "Carried items": a pass's design reads BACKLOG's items tagged `#rung:<its rung>` and the ROADMAP
      "Path to 0.8" entry for its rung.
    - The close checklist's STATUS item gains: a STATUS rewrite moves each item to BACKLOG or ROADMAP and never drops
      an open decision.
    - The worktree setup's last sentence (`:75-77`) says "its pass", not "its lane".
  - `.claude/settings.json`: the `docker compose config*` rule and the two box rules for `docker compose ... restart`
    and `systemctl restart` go (ruling 4). The `systemctl reset-failed` rule stays.
  - One-line superseded headers. Under the build spec's "## Process" heading
    (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:1143`): "Superseded 2026-10-11 by the lean pass
    process (`dubplate-pass`, `pass-core`); the ladder above still governs product order." Directly under the
    parallel-build spec's status line (`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:3`), after a
    blank line: "Superseded 2026-10-11: execution follows the lean pass process (`dubplate-pass`); the run order is
    ROADMAP's Path to 0.8."
  - The rung 7 spec (`docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`): the sentence ending on line 8, "It
    goes to Geoff before any plan is written", becomes "It is agent input: when `intake`'s turn comes, a one-page
    design goes to Geoff, and this spec carries the detail behind it." The "Owed records" heading (`:1233`) keeps its
    title, since `:165` and `:877` cite it by name; one line under it says these are living-reference edits the
    `intake` passes make as tasks, under the adapter's Errata rule. The 14A and BS items are struck with "moot
    (frozen; lean ruling 8)". "Other records" (`:1257-1262`): BACKLOG #8's fix is T3's (row I67), and the discharged
    note stays. No other line changes.
- **Files.** `.claude/skills/dubplate-pass/SKILL.md`, `.claude/settings.json`,
  `docs/superpowers/specs/2026-09-02-greenfield-server-build.md` (one line),
  `docs/superpowers/specs/2026-09-03-parallel-build-approach.md` (one line),
  `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`.
- **Risk class.** `runner`.
- **Acceptance.**
  - The retired-phrase and lane checks print nothing for the adapter (today the retired-phrase check prints "lane
    registry" for it). `grep -c 'site-implementer'` on the adapter prints 0. This prints 8 (today too):

    ```bash
    grep -c -E '^\*\*(Gate|Fast lane|Full-suite home|Worktree setup|CI watch|Risk-class path map|Checklist globs|Close checklist)\*\*:' .claude/skills/dubplate-pass/SKILL.md
    ```
  - `python3 -I -c 'import json; a = json.load(open(".claude/settings.json"))["permissions"]["allow"]; print(len(a),
    [r for r in a if "docker compose" in r or "systemctl restart" in r], "Bash(ssh root@5.78.148.42 systemctl
    reset-failed *)" in a)'` prints `14 [] True` (today `17`, three entries, `True`).
  - `git diff --numstat <base>..HEAD --` on each frozen spec prints `1 0` or `2 0` (one line, plus at most one blank).
  - Every `@@` header in `git diff -U0 <base>..HEAD -- docs/superpowers/specs/2026-10-10-rung-7-intake-design.md`
    starts its old range between 7 and 9 or between 1233 and 1262.
  - No gate in the half (docs only).
- **Test.** The JSON assertion, the eight adapter slots, and the hunk ranges.
- **Pair.** Runs as a pair with T4.
- **Estimate.** 30 minutes.

### T7. Lessons to their executing homes

- **Outcome.** `docs/standing-rules.md` "Lessons the code cannot carry" (`:113-170`) empties into the places that
  execute each lesson, and the section goes. Each becomes a Go doc comment or a file comment under `go-conventions`
  (no process citation, no em dash, plain paragraphs: gofmt rewrites doc-comment lists and headings, and the gate's
  golangci-lint runs gofmt as a formatter). Each destination's comment carries its token verbatim:
  - Oracle harness (`:118-136`): the Navidrome multi-library login (token `Music Library`; `X-ND-Authorization` already appears in `stack.go:541` and `:549`, so it cannot prove arrival, while `Music Library` prints 0 there today) and the gonic sub-path
    limit (`GONIC_PROXY_PREFIX`) into the `internal/harness/subsonic/stack` package doc (`stack/stack.go`); the
    `collection:` key resolution (`artist-set`) into `session/session.go`'s package doc; gonic's error 10
    (`getMusicDirectory`), the per-track cover, disc-subfolder splitting (`Borealis`), the two-second duration floor,
    and the seed's missing lyrics, `nowPlaying`, and radio (`nowPlaying`) into `fixture/doc.go`; "the golden grades
    the grader" (`golden grades the grader`) and the ledger-row rule into `grade/finding.go`'s package doc. T7 adds no
    `doc.go`.
  - The `jsonpointer` and `swag` note (`:131-132`) is dropped: `go mod tidy` owns those `go.mod` lines, and the
    tidy-check leg enforces them.
  - Build for 0.8's stable hook numbers (`:138-139`) moves up into the "Build for 0.8" section of the same file
    (`hook numbers are stable`).
  - Serve routing (`:141-143`) is dropped as already carried: `Shell.Register`'s doc says a path no subtree claims
    answers 404 (`internal/serve/serve.go:64-67`), and `Shell.Handler`'s says the origin check covers every
    registered subtree (`:140-148`). T7 confirms both before dropping.
  - Settings (`:145-148`) into `Handle.Update`'s doc comment (`internal/engine/store/write.go:64`; `reads a setting`).
  - Web (`:150-153`) into a comment beside the kit options in `web/vite.config.ts`, where `adapter-static` is
    configured (`:68`; no `svelte.config.js` exists) (`paths.relative`), as a `//` comment like the options beside it (the filter flags a TSDoc block).
  - Matcher and corpus (`:155-160`) into `internal/engine/musicbrainz/doc.go` (`retry path`).
  - Go tests (`:162-166`): the cobra usage-writer fact beside the first `SetOut` wiring in
    `cmd/dubplate-server/main_test.go` (`:24`; the comment contains `os.Stderr`). Source: cobra v1.10.2 `UsageFunc`
    writes through `c.OutOrStderr()` (`command.go:454`), which is `c.getOut(os.Stderr)` (`:398-399`). The `doctor` row
    rule is dropped as already carried: the doctor tests build their counts from `reg.OwnedWall()`,
    `reg.ReservedClasses()`, `reg.IntegrationClasses()`, and `reg.Thresholds()` (`doctor_test.go:289-292`) plus two
    documented one-line constants (`:22-33`). T7 confirms before dropping.
  - Shell (`:168-170`): a comment above `ssh_cidrs_json` in `scripts/provision.sh:123`, the function whose callers
    run it in a command substitution (`:143`, `:161`) and the subshell BACKLOG #5 names (`errexit`).
  - The standing-rules header line (`:3`) drops "plus lessons harvested from the frozen HISTORY".
- **Files.** `docs/standing-rules.md`, `internal/harness/subsonic/{stack/stack.go,session/session.go,fixture/doc.go,
  grade/finding.go}` (doc comments only), `internal/engine/store/write.go`, `internal/engine/musicbrainz/doc.go`,
  `cmd/dubplate-server/main_test.go`, `web/vite.config.ts`, `scripts/provision.sh`.
- **Risk class.** `auth-data` by path (`cmd/dubplate-server/**`), with comment-only edits.
- **Acceptance.**
  - Before the gate: `gofmt -l` on the touched `.go` files prints nothing.
  - The comment-only filter over `'*.go' '*.ts' '*.sh'` prints nothing, after the non-empty check; the heredoc grep
    over `scripts/provision.sh` (Conventions) prints nothing.
  - No new process citation: `git diff -U0 <base>..HEAD -- '*.go' | grep '^+' | grep -E 'ruled, conductor|Resolution
    R|rung [0-9]|Task [0-9]|\(C[0-9]|plans/.*\.md:[0-9]|_salvage/.*:[0-9]|Ruled, '` prints nothing (the leg's own
    pattern, `check.sh:781`). No em dash: `git diff -U0 <base>..HEAD | grep '^+' | grep -c $'\xe2\x80\x94'` prints 0.
  - Arrival: `grep -c -F` of each token above in its destination file prints 1 or more (`os.Stderr` on a `//` line in
    `main_test.go`: `grep -c -E '//.*os\.Stderr'`).
  - `grep -n 'Lessons the code cannot carry' docs/standing-rules.md` prints nothing, and every lesson in the old
    section has a row (home, or "dropped" with its reason) in the report.
  - The fast lane, `--base <dispatch sha>`, prints `gate exit: 0`; its vale-over-Go-comments leg reads the new
    comments.
  - `diff-reviewer` at task scope reads the T7 commit (pass-core: the `auth-data` read) and returns `accept`; the session runs this read after merging the pair, never the pair itself.
- **Test.** The comment-only filter. Test-first does not apply: no behavior changes.
- **Pair.** Runs as a pair with T5.
- **Estimate.** 60 minutes.

### Close

Per pass-core "Close" and the adapter's close checklist, in order:

1. `git fetch origin` and merge `origin/master` into the branch (a conflict is a stop). Rerun `git ls-files
   .claude/agent-memory`: it prints nothing, or a lesson a concurrent session added is routed as T5 did. Then `bash
   scripts/gate.sh --merge` prints `gate exit: 0` on the close head (the full-suite home, with the branch-history
   gitleaks scan). T7 touched `internal/harness` comments, so `go vet -tags oracle ./cmd/harness/...
   ./internal/harness/...` runs too.
2. The end-to-end checks 2 to 5.
3. `diff-reviewer` at whole-branch scope reads the branch against this plan and the design. It loads the
   `go-conventions` skill (the changed `*.go`), and the checklist globs bring `web-auth-security-reviewer` for T7's
   `cmd/dubplate-server/` path and `svelte-reviewer` and `daisyui-a11y-reviewer` for `web/**` (T5's CSS comments and
   T7's `vite.config.ts` comment; the visual identity spec's path goes with them). The security reviewer's dispatch
   says the diff on its paths is comment-only, asks it to confirm that from the comment-only filter, and to read the
   comments for any false security claim. One batched fix chain follows if any returns `fix`, each finding quoted
   verbatim as the fix's acceptance, then the gate again.
4. `git mv docs/superpowers/plans/2026-10-11-dubplate-alignment.md docs/superpowers/archive/plans/`. The two prompt
   files stay. STATUS's in-flight line becomes "The alignment PR is open for Geoff's read; the next action starts when
   it merges." STATUS stays 60 lines or fewer.
5. Push `glw907/dubplate`, open the PR, and write its body from the PR notes:
   - what landed (task id, commit), T7's `diff-reviewer` verdict, the close review's verdicts;
   - the dependency sweep's line or table, with any held major and any newer workstation binary for the conducting
     session;
   - T2's removal map, T3's map and errata dispositions, T5's triage, and T7's lessons table;
   - **After merge**, the post-merge list exactly as STATUS's next action states it (the pull, the memory backup,
     index, and deletion, the Pi 5 originals' trash, the `admin-api-1` launch), with T4's start epoch;
   - T4's dry-run table and the proposed index;
   - "Found, not done": the "boundary gate" wording in `scripts/check.sh`; the roughly 40 process citations in Go
     comments; anything the tasks found;
   - the score rows (pass-core "Score"), with gate and lock time summed over `cairn-run-gate --records <toplevel>
     <branch>` for the session's worktree and for T5's and T7's (their reports name each worktree and branch).
6. `gh pr view <n> -R glw907/dubplate --web`, then `notify-send -u normal` with the PR link. Geoff reads and merges.

- **Estimate.** 70 minutes (one `--merge` gate of up to 19 minutes plus lock wait, the review, and the PR).
