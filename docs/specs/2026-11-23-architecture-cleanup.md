# Spec: the architecture cleanup chain

Status: approved. Geoff approved it on 2026-11-23 ("Please do").

## Intent

Geoff, 2026-11-23: "Since agents tend to follow code around them, I want to make sure that dubplate's
code is squeaky clean and well-architectured." "Architecture should be logical and idiomatic, naming
should be clean and consistent, etc." "It should be a system that an experienced developer would be
proud of." He asked for the read to cover all 44 packages and the SvelteKit app, and it did:
`docs/research/2026-11-23-architecture-read.md` (49 readers, commit `b5211ce`) holds 37 repo-wide
entries, 354 package findings, the sites of nine patterns, and a ruling on each of the 241 functions
`go tool deadcode` finds unreachable (81 exported, 160 unexported).
Every file:line in the report and the checklists is at `b5211ce`. Pull request 27 (`46fcd28`) later
restructured ten functions and moved lines in their files, so the 58 citations into those files are older
than `master`: a task reads such a finding with `git show b5211ce:<file>` and finds the code on `master`
by its text.

He then chose the scope ("All of it"), the order ("Cleanup first": the whole chain merges before
`intake-2`), and the latitude: "Now is the point where you can change visible portions and the contract
as much as needed in service of a better, cleaner system." Stored data still needs a migration, and a
restore must still work.

### Goals

- Every finding in the report is fixed, or named in its task's `## Decisions` as not held, with the
  reason. Each task's checklist under `docs/specs/2026-11-23-architecture-cleanup/` ends with no open
  box.
- One word means one thing across the Go server, the `/api/v1` contract, the stored values, and the
  TypeScript client.
- Each package, re-read by `standards:go-architecture-reader` after its task, returns no high finding
  that the task's `## Decisions` doesn't name.
- The gocognit exclusion list in `.golangci.yml` is empty, and the `deadcode` leg of `scripts/check.sh`
  fails on a new export with no production caller.

### Non-goals

- Any new product feature, any 0.8 forward hook, and the `intake-2` milestone. The 159
  keep-for-caller exports stay, as the report rules.
- A deploy. The box stays on `d1e93e9`, at schema 0005, until a later deploy carries the chain, and
  migrations 0006 and 0007 run then.
- The visual design. The web task fixes the Svelte, accessibility, and structure findings; the visual
  identity spec (`docs/specs/2026-09-24-visual-identity-design.md`) stays as it is.

### Failure, even if it works

The chain is a failure if it trades the report's findings for new ones. Examples: a rename that
leaves the old name in comments, a split package that only makes sense beside its sibling, or a
contract change made in Go and missed in `openapi.yaml` or the TypeScript. It also fails if a task ticks
a box without reading the code. A finding a reader got wrong, fixed as worded, makes the code worse.

## What happens when it goes wrong

- A finding that doesn't hold against the code, or whose fix as worded would make the code worse,
  gets `- [~] not held: <reason>` and a line in `## Decisions`. Record 0002 lets the reviewer return
  `fix` on that call.
- A finding whose fix needs a product call (a member-visible word that record 0006 doesn't settle, or
  a wire change a client outside the repo would notice) gets `- [~] not held: needs Geoff: <question>`.
  The task builds the rest. The merging session asks Geoff, and his answer becomes a backlog item.
- A fix that turns a test red for a reason other than the rename or contract change the task made is
  reverted, never patched by loosening the test.
- If migration 0007's restore test fails, the task stops changing stored values and reports, and the
  wire renames ship without the stored ones only if `## Decisions` says why both still read cleanly.
- A task that runs past its time clause stops, pushes, and marks its open boxes in the pull request.
  Its short `task-score` stops its branch of the chain, as "Chains" step 5 says.
- The chain tick checks weekly usage before each launch, against the ceiling in "Launching the chain".

## Approaches

1. **A chain of ten tasks, one linear Go line with web beside it (the pick).** Repo-wide renames land
   first, so later tasks fix package findings in code that already uses the settled words. One
   line means no two open branches edit one Go file. Web runs beside the engine task because neither
   touches the other's files, and every wire change has landed before web starts.
2. **One task per package, 44 pull requests.** Each would pay CI's 46-minute median wait, and the
   repo-wide renames would conflict across all of them.
3. **One task.** About 2,900 minutes of work in one branch, with one review that can't hold it.

## Mechanisms and their sources

- Refactoring keeps observable behavior
  (<https://martinfowler.com/bliki/DefinitionOfRefactoring.html>). Documented. This chain changes
  names, wire spellings, and printed text by Geoff's ruling, so a test may change to follow a
  rename or a contract change the task made, and for no other reason.
- gopls's rename "renames a symbol" through the LSP rename request, with the type checker's view of
  every reference (<https://go.dev/gopls/features/transformation#rename>). Documented. Tasks rename Go identifiers with
  it. Three `sed -i` and Python rewrites of Go files in dubplate worktrees were refused
  (`~/.local/state/claude-denials.jsonl`, 2026-11-16 and 2026-11-17).
- Go package naming: "Good package names are short and clear", and a name should not repeat in its
  contents (<https://go.dev/blog/package-names>). Documented. It grounds the `local` to `jobs` rename and
  the stutter findings.
- depguard enforces import rules from `.golangci.yml`
  (<https://golangci-lint.run/docs/linters/configuration/>). Documented.
- Error strings "should not be capitalized ... or end with punctuation" (Go Code Review Comments,
  <https://go.dev/wiki/CodeReviewComments#error-strings>). Documented. That page says nothing on a
  package prefix, and the standard library does both: `os.ErrNotExist` reads "file does not exist",
  and `bufio.ErrBufferFull` reads "bufio: buffer full". Bare text is the report's ruling and the repo's
  majority: 143 of the 241 `errors.New` strings outside tests carry no prefix today. Probed.
- `id.ID` implements `encoding.TextMarshaler`, "the interface implemented by an object that can
  marshal itself into a textual form", and `encoding.TextUnmarshaler` (`go doc encoding`). Documented.
- SQLite's `RENAME COLUMN` changes the name "both within the table definition itself and also within
  all indexes, triggers, and views that reference the column"
  (<https://www.sqlite.org/lang_altertable.html>, section 3). Documented. `owner` and `uploader` hold
  their foreign keys in their own table's definition. `internal/store/migrate.go:288` turns
  `foreign_keys` off for the whole migration, so migration 0007's test checks `PRAGMA
  foreign_key_check` after it runs.
- A decision record is immutable and a change of mind supersedes it (Nygard,
  <https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions>). Documented. Record 0001
  adopts it.
- The checklists are invented: one line per finding, quoted to its first sentence, generated from the
  report.
- The minutes per finding are invented, anchored on pull request 25 (`structure-doctor`, 194 minutes
  actual, four fix rounds). Each cross-package entry costs about 7 minutes and a large one 20; a local
  finding 2; a mechanical move 1. Each task adds local gates 56 (two runs at the 28-minute median), CI
  46, and review 13 to 38, from task-report's dubplate rows for 2026-10-24 to 2026-11-23.
- Record 0003: every task is neutral and lands no forward hook. No prior art applies.

## Assumptions

- `complexity-cleanup` (`7032924c`) is ready or merged before task 1 launches, and task 1 builds on
  it. After it, gocognit excludes only `commitVerdict` and `walkZIP`. Probed: its spec's done item 1.
- The report's sites are at `b5211ce`. `complexity-cleanup` moves lines in ten functions, so a site
  may sit a few lines off. Each task reads the code at the site before it fixes the finding.
- No client outside the repo reads `/api/v1`. Probed: `web/src/lib/api` is the only consumer in the
  repo, and the box runs no client but the bundled web app (dubplate `docs/STATUS.md`, "The box is
  unchanged").
- Today's counts, at `origin/master` on 2026-11-23, are in each task's done items. Each command is
  `git grep`, which reads tracked files only, so a worktree under `.claude/` never counts.

## Files

Each task's checklist names its files by site. By task: the vocabulary and boundaries tasks edit
`internal/{auth,identity,ops,settings,serve,config,notify,intake,backup,store}` and
`cmd/dubplate-server`. Each new package (`internal/member`, `internal/settings` as the setting
reader's home, `internal/fsys`) gets a line in the dubplate-build skill's layering rule naming the
internal packages it must not import, held by `internal/webapp/layering_test.go` until task 8 moves
the rules to depguard. The wire task edits those, `internal/store/migrations/` (new `0007`),
`internal/ops/opshttp/openapi.yaml`, `web/src/lib/api/*.ts`, the web files that switch on a renamed
value, `docs/specs/2026-09-11-configuration-homes-design.md`, and `docs/research/passb-copy-final.md`.
The web task edits `web/**` except `web/embed.go`, which task 4 owns. The engine task edits
`internal/engine/**` and `internal/opensubsonic/**`, and moves `internal/engine/matcher/unidecode` to
`internal/unidecode`. The harness tasks edit
`internal/harness/**`, `cmd/harness`, `tools/close-evidence`, and `.golangci.yml`. Task 10 edits
`scripts/check.sh`. Records go in `docs/decisions/`, numbered one past the highest on the task's
branch.

No task edits `docs/STATUS.md` or `BACKLOG.md`; the merging session does. A task's pull request lists
the backlog items it fixed under `## Backlog`, and the merging session moves them to `## Done`.

## Commands against the denials log

Read against `~/.local/state/claude-denials.jsonl`, 2026-10-24 to 2026-11-23. Tasks run `go`,
`gopls rename`, `golangci-lint`, `npm` in `web/`, `bash scripts/gate.sh`, `git`, `gh pr create`, and
`gh pr checks --watch`. None was refused. The refusals in dubplate worktrees were bulk `sed -i` and
Python rewrites of Go source and one edit to a test file through the shell, so tasks edit through
the Edit tool and `gopls rename`. A download into a temp directory was refused once; no task downloads.

## Done items every task carries

Below, `T` is the task's name and `L` is `docs/specs/2026-11-23-architecture-cleanup/T.md`. `P` is
the parent's branch, from the goal's opening lines, or `master` for a task with no parent. `B` is
`$(git merge-base origin/P HEAD)`, the base the review reads, so a diff from `B` holds only this task's
work.

1. `grep -c '^- \[ \]' L` prints the task's count today and `0` after. A box is ticked `- [x]` with
   the short hash of the commit that fixed it, or `done in <task>` when an earlier task's change already
   fixed it.
2. `grep '^- \[x\]' L | grep -cvE '[0-9a-f]{7}|done in '` prints `0`. `grep -c '^- \[~\]' L` prints the
   number of findings `## Decisions` names as not held, and each such line carries its reason.
3. `comm -23 <(git diff "$B" -M -U0 -- '*_test.go' | grep -oE '^-func (Test|Fuzz)\w+' | cut -c7- | sort -u) <(git diff "$B" -M -U0 -- '*_test.go' | grep -oE '^\+func (Test|Fuzz)\w+' | cut -c7- | sort -u)`
   prints only test names that `## Decisions` lists: a renamed test as `old -> new`, and a removed one
   with the reason it went.
4. `go tool deadcode ./... | wc -l` prints no more after than at `B` (241 today, exported and
   unexported). A new export names its production caller in the commit that adds it (record 0003).
5. `git diff "$B" --name-only -- docs/STATUS.md BACKLOG.md` prints nothing.
6. `standards:go-architecture-reader` re-reads each Go package the task owns after its last fix
   commit. The pull request quotes each verdict, and every high finding is fixed or named in
   `## Decisions`.
7. `bash scripts/gate.sh` passes from the worktree root, and the pull request quotes its
   `gate exit: 0` line.
8. CI passes: `gh pr checks <n>` lists every check as `pass`.
9. The pull request's `## Backlog` lists each item the "Backlog sweep" folds into the task, with the
   short hash of the commit that fixed it.

Tasks 7 to 10 also carry item 10: `git diff "$B" --name-only -- web internal/ops/opshttp/openapi.yaml`
prints nothing.

On an `auth-data` path, a change in behavior gets an acceptance test that fails at `B`, recorded in
the pull request, and one mutation proof per changed behavior, run as the dubplate-build skill's
Conventions say. A rename alone has no new behavior to test.

## Task 1: `vocabulary`

The task fixes the ten vocabulary entries in its checklist. It renames with `gopls rename`. It adds
decision record 0016, which gives each verb one meaning (Parse, Validate, Check, Verify) and sets
sentinel text as a bare lowercase phrase.

Done when the common items hold (checklist count 10), and:

1. `git grep -nE '^type [A-Za-z]*Seam\b' -- 'internal/*.go' ':!*_test.go' | wc -l` prints `7` today
   and `0` after.
2. `git grep -nE '^type (Actor|Author) struct' -- 'internal/*.go' | wc -l` prints `3` today and `1`
   after, in `internal/identity`.
3. `git grep -n 'errors.New(".*no valid session")' -- 'internal/*.go' ':!*_test.go' | wc -l` prints
   `3` today and `1` after. `git grep -nE 'Err(NotSignedIn|Unauthenticated)\s*=\s*errors\.New' --
   'internal/*.go' ':!*_test.go' | wc -l` prints `4` today and `1` after: one sentinel for a request
   with no valid session.
4. `git grep -n ClientAddress -- 'internal/*.go' 'cmd/*.go' | wc -l` prints `9` today and `0` after.
5. `test -d internal/ops/local; echo $?` prints `0` today and `1` after; the package is `jobs`.
6. `git grep -nE 'errors\.New\("[a-z/]+: ' -- 'internal/*.go' 'cmd/*.go' ':!*_test.go' | wc -l` prints `98`
   today and `0` after.
7. `ls docs/decisions | wc -l` prints `15` today and `16` after.

Risk: `auth-data`. Reviewers: `standards:diff-reviewer`, `standards:web-auth-security-reviewer`.
Minutes: 212 (work 95, gates 56, CI 46, review 15, re-read inside work). Time clause: 424 minutes.
Model: Opus at high. Parent: `complexity-cleanup`, the open pull request of
`docs/specs/2026-11-23-complexity-cleanup.md`, while it is open. Once it has merged, task 1 has no parent
and launches from `master`.

## Task 2: `wire-contract`

The task fixes the eight wire entries as one contract revision: Go, `openapi.yaml`, the TypeScript
under `web/src/lib/api` and every web file that switches on a renamed value, and the two living specs
the kebab spellings came from. Every wire and stored enum becomes snake_case. A setting is `key` and
its sensitivity is `sensitivity`. `id.ID` gains MarshalText and UnmarshalText, so wire structs carry
typed ids. Foreign keys to `users` are `user_id`. `JobStatus` and `JobStatusRequest` lose their JSON
tags, since no operation serves them yet (record 0003).

Migration 0007 rewrites every stored value the renames touch and renames the `owner` and `uploader`
columns. Its test builds a store at 0006, seeded with every old spelling, migrates it, reads each
new spelling back, and finds `PRAGMA foreign_key_check` empty. Two restore tests each restore a
committed backup archive and open it on the new schema, reading the new spellings back. One archive
is written at `B`. The other is written at `d1e93e9`, the box's commit and schema 0005: the task
builds that commit's `dubplate-server` in a detached worktree (`git worktree add --detach <dir>
d1e93e9`), seeds a store with every old spelling, runs its `backup` command, and gives the commands
in the pull request. A restore of the box's real backups must work after the next deploy. The task
adds decision record 0017: wire and stored enums
are snake_case, and a foreign key is named for the table it references.

Done when the common items hold (checklist count 8), and:

1. `git grep -cE 'Code = "[a-z]+(-[a-z]+)+"' -- internal/reject/reject.go` prints
   `internal/reject/reject.go:26` today and nothing after.
2. `git grep -nE 'enum: \[.*-' -- internal/ops/opshttp/openapi.yaml | wc -l` prints `2` today (lines
   426 and 456, `env-seed`) and `0` after.
3. `git grep -nE '^\s+mark:' -- internal/ops/opshttp/openapi.yaml | wc -l` prints `1` today and `0` after.
4. `awk '/^type JobStatus(Request)? struct/,/^}/' internal/ops/ops.go | grep -c 'json:"'` prints `11`
   today and `0` after.
5. `git grep -nE 'func (parseID|wireID|parseUserID|handleOf)\b' -- 'internal/*.go' | wc -l` prints `5`
   today and `0` after.
6. `ls internal/store/migrations | wc -l` prints `6` today and `7` after. The migration test and the
   two restore tests above fail at `B` and pass after; the pull request quotes all three.
7. `git grep -nE "['\"]env-seed['\"]" -- web/src | wc -l` prints `1` today
   (`web/src/lib/copy/deck.ts:264`) and `0` after. No reject code appears in `web/src` today.
8. `npm run check --prefix web` and `npm run build --prefix web` pass.

Risk: `auth-data`. Reviewers: `standards:diff-reviewer`, `standards:web-auth-security-reviewer`,
`standards:svelte-reviewer` for the web files. Minutes: 212 (work 95, gates 56, CI 46, review 15).
Time clause: 424 minutes. Model: Opus at high. Parent: `vocabulary`.

## Task 3: `server-boundaries`

The task fixes the eight boundary entries. Setting readers move to `internal/settings` as one
`settings.Reader[T]`. The roster leaves auth for `internal/member`, which owns the `users` and
`member_preferences` tables (the ownership list at `internal/store/owners_test.go:35` changes in the
same commit). Each package
reports its own doctor rows through one `doctor.Row` type, and cmd ranges over them. auth stops
importing serve. `auth.ConfigFrom` resolves auth's values once, and the SMTP keys move to notify. One
`internal/fsys` holds the free-bytes probe. "Rung" leaves operator output. The task adds decision
record 0018 for the doctor row contract, which builds on record 0013.

Done when the common items hold (checklist count 8), and:

1. `git grep -n 'type Setting\[' -- 'internal/*.go'` prints `internal/auth/handlers.go:75` today, and
   after it prints nothing; `git grep -n 'type Reader\[' -- internal/settings` prints one line.
2. `go list -deps ./internal/auth | grep -c 'dubplate/internal/serve$'` prints `1` today and `0` after.
3. `test -d internal/member; echo $?` prints `1` today and `0` after.
4. `git grep -lE 'func [A-Za-z]*RowStatus\(' -- 'internal/*.go' ':!*_test.go' | wc -l` prints `6`
   today and `0` after.
5. `git grep -nE '^func ([fF]reeBytes|rootSpace)\(' -- 'internal/*.go' ':!internal/fsys' | wc -l`
   prints `3` today and `0` after.
6. `git grep -nE '^[^/]*"[^"]*rung' -- cmd/dubplate-server internal/config ':!*_test.go' | wc -l`
   prints `2` today (`cmd/dubplate-server/doctor.go:71`, `internal/config/keys.go:167`) and `0` after.
   Comments that say "rung" are task 4's and task 9's process-citation findings.
7. `ls docs/decisions | wc -l` prints `17` at the task's start and `18` after.

Risk: `auth-data`. Reviewers: `standards:diff-reviewer`, `standards:web-auth-security-reviewer`.
Minutes: 230 (work 115, gates 56, CI 46, review 13). Time clause: 460 minutes. Model: Opus at high.
Parent: `wire-contract`.

## Task 4: `server-access`

The task fixes every package finding, pattern site, and move-to-test or delete ruling for
`internal/{auth,identity,notify,serve,webapp,ops,ops/admin,ops/opshttp}`, `cmd/dubplate-server`, and
`web/embed.go`, and the entry "Two package comments for serve". The `web/embed.go` rows rename
`web.Assets`, whose callers are in `cmd/dubplate-server` and `internal/webapp`.

Done when the common items hold (checklist count 132). Risk: `auth-data`. Reviewers:
`standards:diff-reviewer`, `standards:web-auth-security-reviewer`. Minutes: 410 (work 280, gates 56,
CI 46, review 28). Time clause: 820 minutes. Model: Opus at high. Parent: `server-boundaries`.

## Task 5: `server-data`

The task fixes every package finding, pattern site, and ruling for
`internal/{atomicfile,backup,config,id,intake,intake/gate,intake/staging,ops/local,reject,settings,store}`
(`ops/local` is `jobs` after task 1). It also restructures `(*submissions).commitVerdict`
(`internal/intake/accept.go`) and `walkZIP` (`internal/intake/gate/zip.go`) to a gocognit score of
30 or under, with no change in behavior, as `complexity-cleanup` did for the other ten.

Done when the common items hold (checklist count 133), and:

1. `grep -c 'text: func' .golangci.yml` prints `2` at the task's start and `0` after.
2. `golangci-lint run ./...` exits 0.

Risk: `auth-data`. Reviewers: `standards:diff-reviewer`, `standards:web-auth-security-reviewer`.
Minutes: 440 (work 300, gates 56, CI 46, review 38). Time clause: 880 minutes. Model: Opus at high.
Parent: `server-access`.

## Task 6: `web`

The task fixes every finding in the web, web:svelte, web:a11y, and web:structure tables and their
pattern sites. It reads the visual identity spec first and keeps to it. It renders each changed route
and reads the screenshot, as the dubplate-build skill's "Reviewers by path" says.

Done when the common items hold (checklist count 78; item 6 reads no package, since `web/embed.go`
is task 4's), and:

1. `git diff "$B" --name-only | grep -vE '^(web/|docs/specs/2026-11-23-architecture-cleanup/)'`
   prints nothing, and `git diff "$B" --name-only -- web/embed.go` prints nothing.
2. `npm run check --prefix web` and `npm run build --prefix web` pass, and `make web-rendered` passes.

Risk: `auth-data` for `web/src/routes/{login,members}`, `ordinary` for the rest. Reviewers:
`standards:diff-reviewer`, `standards:svelte-reviewer`, `standards:daisyui-a11y-reviewer` (each given
the visual identity spec), and `standards:web-auth-security-reviewer` for the two auth-data routes.
Minutes: 340 (work 210, gates 56, CI 46, review 28). Time clause: 680 minutes. Model: Opus at high.
Parent: `server-data`.

## Task 7: `engine`

The task fixes every package finding, pattern site, and ruling for `internal/engine/**`,
`internal/opensubsonic`, and `internal/opensubsonic/client`. The matcher takes its transliterator as
a dependency, so `SetTransliterator` and its global go. `unidecode` moves to `internal/unidecode`, so
the task also edits its importers: `cmd/dubplate-server/{doctor,unidecode,unidecode_test}.go`,
`cmd/harness/fetchunidecode{,_test}.go`, `internal/engine/matcher/testmain_test.go`,
`internal/engine/tag/testmain_test.go`, `unidecodetest/unidecodetest.go`,
`internal/harness/conform/{defaults,defaults_test,testmain_test}.go`, and the comment at
`scripts/deploy.sh:310`. The `cmd/harness` and `internal/harness/conform` edits land before task 8
moves those files, which the linear chain allows.

Done when the common items hold (checklist count 94) with item 10, and:

1. `git grep -n SetTransliterator -- '*.go' | wc -l` prints `13` today and `0` after.
2. `test -d internal/unidecode; echo $?` prints `1` today and `0` after.
3. `gh pr checks <n>` lists `fuzz` as `pass`.

Risk: `ordinary`, and `auth-data` for the `cmd/dubplate-server` files. Reviewers:
`standards:diff-reviewer`, and `standards:web-auth-security-reviewer` for the cmd commit. Minutes:
325 (work 205, gates 56, CI 46, review 18). Time clause: 650 minutes. Model: Opus at high. Parent:
`server-data`.

## Task 8: `harness-shape`

The task fixes the eight harness-shape entries. The harness tree splits into
`internal/harness/corpus/...` and `internal/harness/subsonic/...`, with a matching `harness corpus` and
`harness subsonic` cobra tree. `internal/opensubsonic/conformance` moves to
`internal/harness/subsonic/jsonrule`. cassette exports the case-directory operations. Import rules
become depguard rules, plus one `archtest` package where a rule needs code. The recorder leaves
package main, and one `RepoRoot` helper replaces the copies. The moves also reach the gate: the
`UNCACHED_TESTS` list and the conform run at `Makefile:22-27`, the comment at `scripts/check.sh:725`,
and the no-link test's description in `.claude/skills/dubplate-build/SKILL.md:26-30`, a living
reference edited in place. #162 names `corpus/README.md`, which this task reshapes, so it checks it.

Done when the common items hold (checklist count 8) with item 10, and:

1. `grep -c depguard .golangci.yml` prints `0` today and 1 or more after.
2. `test -d internal/opensubsonic/conformance; echo $?` prints `0` today and `1` after.
3. `git grep -n 'func repoRoot' -- '*.go' | wc -l` prints `7` today and `0` after; one exported
   `RepoRoot` remains.
4. `make test` passes, so the `UNCACHED_TESTS` names resolve at their new paths.

Risk: `runner` for `Makefile` and `scripts/check.sh`, `ordinary` for the rest. Reviewers: `standards:diff-reviewer`. Minutes: 195 (work 80, gates 56, CI 46,
review 13). Time clause: 390 minutes. Model: Opus at high. Parent: `engine`.

## Task 9: `harness-corpus`

The task fixes every package finding, pattern site, and ruling for the corpus side:
`internal/harness`, analysis, cassette, conform (70 move-to-test rulings), oracle, stub, synth, and
`cmd/harness`, at their paths after task 8. It also takes the entry "Comments that cite process
artifacts".

Done when the common items hold (checklist count 168) with item 10. Risk: `ordinary`. Reviewers:
`standards:diff-reviewer`. Minutes: 390 (work 260, gates 56, CI 46, review 28). Time clause: 780
minutes. Model: Opus at medium. Parent: `harness-shape`.

## Task 10: `harness-subsonic`

The task fixes every package finding, pattern site, and ruling for the subsonic side: fixture, grade,
session, stack, xmlrule, jsonrule, and `tools/close-evidence`. It then makes the `deadcode` leg of
`scripts/check.sh` fail: the findings left after the chain go in a named exclusion file that only
shrinks, one per line, keyed by package path and function name so a line move doesn't break it, and
`CHECK_EXPORTED_UNUSED` goes. Decision record 0019 supersedes
record 0003, keeps its decision word for word, and changes its Consequences to the failing leg.

Done when the common items hold (checklist count 72) with item 10, and:

1. `grep -c CHECK_EXPORTED_UNUSED scripts/check.sh` prints `4` today and `0` after.
2. With one unused exported function added to `internal/id`, `bash scripts/check.sh` fails on the
   deadcode leg; without it, the leg passes. The pull request quotes both runs.
3. The exclusion file's line count equals `go tool deadcode ./... | wc -l` after, and no more than 241.
4. `grep -c '^Superseded by 0019' docs/decisions/0003-build-for-0-8.md` prints `0` today and `1` after.

Risk: `runner` for `scripts/check.sh`, `ordinary` for the rest. Reviewers: `standards:diff-reviewer`.
Minutes: 305 (work 190, gates 56, CI 46, review 13). Time clause: 610 minutes. Model: Opus at medium.
Parent: `harness-corpus`.

## Launching the chain

The chain runs by `building-from-specs` "Chains". It starts from `~/Projects/dubplate`, on Sonnet at
medium, with `/loop 15m Run one chain tick for docs/specs/2026-11-23-architecture-cleanup.md, by
building-from-specs "Chains", with the usage ceiling in the spec's "Launching the chain".` Geoff set
this chain's ceiling on 2026-11-23: a tick launches while `Current week (all models)` is under 95
percent until the weekly reset at 23:00 on 2026-11-27 (America/Anchorage), and under 85 percent after
it. Each task launches as its name. A goal holds 4,000 characters, so it cites the done items by
section and carries none of their text. After a child's opening lines, it reads:

```text
Build task <N>, `<T>`, of docs/specs/2026-11-23-architecture-cleanup.md in dubplate. Read the spec
with `git -C /var/home/glw907/Projects/dubplate show
origin/master:docs/specs/2026-11-23-architecture-cleanup.md`, and work the checklist
docs/specs/2026-11-23-architecture-cleanup/<T>.md. `P` is <parent branch, or master>. The goal is
met when every item under "Done items every task carries" and under "Task <N>" holds as the spec
words it or is marked unmet with its reason in the pull request, the standards:diff-reviewer
review accepts, and the pull request is ready with planPathMin <minutes>. Or stop after <clause>
minutes.
```

## Merge order

`complexity-cleanup` first, then `vocabulary`, `wire-contract`, `server-boundaries`, `server-access`,
`server-data`, `engine`, `harness-shape`, `harness-corpus`, `harness-subsonic`, and `web` last. Each
child merges its parent's branch before it starts. `web` and `engine` share no file, but `web` doesn't
hold the engine line's commits, so the "Current" check in "Merging" fails for it once that line has
merged. One refresh task then merges `master` into `web` and reruns its gate, CI, and review: about
115 minutes more (gates 56, CI 46, review 13). `intake-2` builds after `web` merges.

## Backlog sweep

The scope was every open item under `## Open` and `## Later` that names a file under `internal/`,
`cmd/`, `web/`, `tools/`, `scripts/check.sh`, or `.golangci.yml`, plus #126 and #185: 80 items, read
at `origin/master` by one Sonnet reader. A full sweep is not due: 104 of 117 open items carry a stamp
from the last 90 days. An item folds only where a report finding already rewrites its lines.

| Item | Mechanism today | Verdict | Task, fix, minutes |
| --- | --- | --- | --- |
| #24 | `internal/auth/nonce.go:130`: `RunSweep` logs `err.Error()` | fold | 4: the fixed-reason log, in the edit for the finding at `nonce.go:122`; 15 |
| #32 | `fixture/manifest.go:129-137`: `checkPath` tracks exact paths only | fold | 10: refuse a path that prefixes another, in the edit for `manifest.go:150`; 15 |
| #146 | `store/intent.go:174`, `store/restore.go:189` format the cause with `%v` | fold | 5: the merged decoder wraps with `%w` (`restore.go:180`); 10 |
| #155 | the members and server-settings pages, `AdminPage.svelte:40-59`, `WriteNotice.svelte:35` | fold | 6: one admin-page controller owns refresh, retry, and refusal focus (`members/+page.svelte:77`, `:161`); 40 |
| #164 | `conform/case.go:614`, `:641` cite "Resolution R1" and "rung-3" | fold | 9: the strings lose the tags (`conform/component.go:19`); 10 |
| #179 | `cmd/dubplate-server/doctor.go:75` drops `class.Enabled`'s error | fold | 3: print the refusal, in the "rung" rewrite (`doctor.go:71`); 10 |
| #180 | `doctorrows.go:105-122` echoes an environment value | fold | 3: the shared row type never prints a sensitive value; 10 |
| #182 | `unidecode/install.go:65-116`: three table probes, no size cap, no `Lstat` | fold | 7: one table-read helper with `Lstat` and a cap (`install.go:89`); 25 |
| #183 | `backup/doctor.go:165`, `:168` print rclone's stderr | fold | 5: a fixed phrase and the exit status (`backup/doctor.go:149`); 10 |
| #184 | `member.go:51-57` and a test comment state the wrong bound | fold | 4: corrected in the edits for `backup.go:63` and `serve.go:158`; 5 |
| #185 | the `deadcode` warning leg and the gocognit exclusions | reword, then fold | 10: the text in `backlog-sweep.md`; task 5 empties the exclusions |
| #77 | "api request admitted" is gone from `internal/` | done | `internal/ops/opshttp/opshttp.go:215-217` (`cleanup-ops`) |
| #121, #123, #140, #160, #165, #167, #181 | each clause is a report site | duplicate | move to `## Done` at approval, each naming the checklist that holds its finding |
| #47, #57, #66, #73, #82, #88, #124, #125, #126, #149, #156, #159, #166, #168, #177 | part is a report site, or a line moved | reword | the text in `docs/specs/2026-11-23-architecture-cleanup/backlog-sweep.md` |

The rest of the 80 items are kept: the mechanism holds, and no finding rewrites its lines. They are
#8, #13 to #16, #18, #20, #28, #31, #48, #52, #64, #72, #74, #79 to #81, #83, #84, #87, #89, #90, #92,
#96, #120, #132, #137, #141, #142, #145, #148, #150 to #153, #158, #161, #163, #170, #171, #176, and
#178. The reader called #178 a duplicate, but no finding bounds `member list`'s roster read
(`cmd/dubplate-server/member.go:210`), so it stays kept.
The reader proposed folding #31 into task 8. It stays kept, because its fix adds a digest check inside
`stack.Start`, a line no finding rewrites.

Approving this spec lands the rewords, the stamps on kept items, and the moves of #77 and the
duplicates to `## Done` in one `BACKLOG.md` commit. The merging session moves each task's folded items
to `## Done` when it merges that task.

## Calls Claude made

- Ten tasks: one per repo-wide theme where the change crosses packages (vocabulary, wire, boundaries,
  harness shape), then one per package group. Cross-package renames land first, so package fixes are
  made in code that already uses the settled words.
- Web runs after every wire change (task 2 and the server tasks) and beside the engine task, so no two
  open branches share a file, and it merges last, after one refresh. Backlog moves wait for the merge,
  so the two open branches never both edit `BACKLOG.md`.
- Task 1 builds on the open `complexity-cleanup` pull request, so the chain can start before Geoff
  merges it.
- Every done-item count is a `git grep`, since this shell's `grep` is ugrep with different defaults.
- The restore tests include an archive written at the box's own commit, `d1e93e9`, since the box's
  real backups sit at schema 0005.
- Sentinel text is bare, by the report's ruling and the repo's majority.
- `web/embed.go`'s rows go to task 4, which edits its Go callers.
- Snake_case for every wire and stored enum, since most of the ops contract already uses it. The
  stored values change through migration 0007, with a restore test, since Geoff allowed contract
  changes and asked that stored data still migrate and restore.
- `opensubsonic/conformance` moves to `harness/subsonic/jsonrule`. The report notes a tension with
  publishing opensubsonic as its own module. No spec plans that module, and moving the grader back is
  one move.
- `commitVerdict` and `walkZIP` are restructured in task 5, since cleanup now runs before `intake-2`.
  This empties the gocognit exclusions, so #185's score ceiling has nothing left to hold.
- Each task re-reads its packages with `standards:go-architecture-reader`, so the chain is measured
  against the report's bar and not only against the report's list.
- Opus at high for the server, wire, and engine tasks, which carry `auth-data` or cross-package
  judgment, and Opus at medium for the two harness package tasks, which are mostly moves and local
  fixes.
- The minutes total 3,059, with time clauses at twice each task's minutes, plus about 115 for
  `web`'s refresh. They include 150 minutes for the ten folds.
