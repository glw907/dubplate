# Spec: restructure the over-complex functions

## Intent

Geoff, 2026-11-23: "Given that new overly complex code can't happen is there any strong reason not to fix
all the old instances now?" and "Yes, write the dubplate spec". The lean-pattern pilot (pull request 26)
set gocognit's bar at 30 and excluded the 12 functions over it by name. This spec restructures them, so
the exclusion list empties instead of waiting on specs that may never touch the code.

### Goals

- Each of ten functions scores 30 or under on gocognit, which reports only scores above 30, and its
  exclusion leaves `.golangci.yml`.
- Nothing a member, an operator, or a test sees changes. Every existing test passes unedited.
- Each function is covered before it changes: 85 percent of its statements at the commit before its
  restructure.

### Non-goals

- `(*submissions).commitVerdict` (`internal/intake/accept.go:55`) and `walkZIP`
  (`internal/intake/gate/zip.go:90`). `intake-2` takes #126, whose findings I2 and I3 edit lines inside
  both, and the refactor-first rule puts their restructure in that spec's first task.
- The 241 test-only exports (#185, part 2), and #185's score ceiling for what stays excluded.
- Any printed string, doctor row, or exported identifier.

### Failure, even if it works

The scores drop because a function was cut into pieces that only make sense together, named for their
position rather than what they do. Record 0002 lets the reviewer return `fix` on awkward code, and it
applies here.

## What happens when it goes wrong

A function that can't reach 30 or under without a change in behavior or in its exported surface keeps
its exclusion, and the pull request's `## Decisions` names it with the reason. If fixture tests can't
bring `gradeCaseDefault` to 85 percent, the restructure moves each uncovered block into its helper
unchanged, and `## Decisions` gives the figure reached and the uncovered lines. No test seam is added
to reach the bar. A restructure that turns a
test red is reverted, never patched by editing the test. No function's score rises above today's.

## Approaches

1. **One task, one commit per function (the pick).** One review and one CI wait, and a single function
   reverts alone. No two of the ten share a file.
2. **One task per package.** Ten pull requests, each paying CI's 46-minute median wait, for no reduction
   in risk.
3. **Leave them to refactor-first.** Each function is restructured when a spec changes it. The six under
   `internal/harness` sit outside every 0.8 milestone and may never be touched, so their exclusions would
   stay.

## Mechanisms and their sources

- Refactoring is "a change made to the internal structure of software to make it easier to understand and
  cheaper to modify without changing its observable behavior"
  (<https://martinfowler.com/bliki/DefinitionOfRefactoring.html>). Documented.
- Tests come before the change, so a broken restructure fails a test
  (<https://martinfowler.com/bliki/SelfTestingCode.html>). Documented. The 85 percent bar is invented,
  and it sits under nine of the ten functions' coverage today.
- gocognit scores cognitive complexity (<https://github.com/uudashr/gocognit>). Documented. Probed: the
  run in done item 1 prints the 12 functions, from 31 to 50, each in the form "cognitive complexity 35
  of func `newDoctorCmd` is high (> 30)". A score of 30 passes.
- Record 0003: the task is neutral and lands no forward hook. No prior art applies, since a restructure
  takes no library.

## Assumptions

- No open 0.8 item edits the ten functions. Probed: a search of `BACKLOG.md`'s open items for each file
  and function name found #164 (strings inside `gradeCaseDefault`), #140 and #179
  (`cmd/dubplate-server/doctor.go`), and #185. The backlog sweep below rules on each.
- Today's coverage, from `go test -cover` and `go tool cover -func` with every `DUBPLATE_*` name unset:

  | Function | Score | Function coverage | Package coverage |
  | --- | --- | --- | --- |
  | `newDoctorCmd` (`cmd/dubplate-server/doctor.go:30`) | 35 | 98.5% | 90.9% |
  | `findPath` (`internal/engine/matcher/assign.go:319`) | 36 | 100% | 95.7% |
  | `parseBlocks` (`internal/engine/tagio/container.go:217`) | 43 | 96.4% | 94.0% |
  | `CheckAlbumInfoLedger` (`internal/harness/analysis/album_info.go:155`) | 35 | 94.1% | 89.2% |
  | `gradeCaseDefault` (`internal/harness/conform/case.go:519`) | 34 | 61.5% | 78.5% |
  | `(*Library).check` (`internal/harness/subsonic/fixture/manifest.go:92`) | 31 | 95.7% | 92.2% |
  | `(*assignment).problems` (`internal/harness/subsonic/grade/assignment.go:153`) | 40 | 91.2% | 93.6% |
  | `(*Session).validate` (`internal/harness/subsonic/session/session.go:338`) | 45 | 87.2% | 88.9% |
  | `(*schema).project` (`internal/harness/subsonic/xmlrule/project.go:69`) | 50 | 90.0% | 89.3% |
  | `(leaf).swapPlan` (`internal/store/swap.go:133`) | 31 | 93.9% | 86.0% |

## Files

Done item 1's command gives 12 hits: the ten files in the table and the two intake files under
Non-goals. The task also edits `.golangci.yml` (`grep -n 'text: func' .golangci.yml` gives 12 hits, one
per function), the tests beside `internal/harness/conform/case.go`, and `docs/STATUS.md`. Fixtures for
`gradeCaseDefault`'s tests are built in `t.TempDir()`. `BACKLOG.md` is not the task's: the author lands
the sweep's #164 reword when Geoff approves this spec, and the merging session lands the #185 and #140
edits.

## Commands against the denials log

Read against `~/.local/state/claude-denials.jsonl`, 2026-10-24 to 2026-11-23. The task runs `go test`,
`go tool cover`, `golangci-lint`, `bash scripts/gate.sh`, `git`, `gh pr create`, and `gh pr checks
--watch`. None was refused in a task worktree. The refusals that name them were a `kill` before a
`go test`, a push to another repo's `main`, and two `gh pr checks` that rode with a merge. The task
merges nothing.

## Task 1: `complexity-cleanup`

The task restructures each function in the table, one commit per function, in the table's order. Before
it changes a function, it reads that function's coverage. A function under 85 percent gets tests first,
in their own commit, that pin what it does today; today that is `gradeCaseDefault` alone. The
restructure then extracts unexported helpers, each named for what it does, or flattens the control flow
with early returns. It keeps every signature, every exported identifier, and every string the function
prints or returns, and it removes the function's exclusion from `.golangci.yml` in the same commit.
`docs/STATUS.md`'s "Current state" gains one line naming the work and its pull request.

`cmd/dubplate-server/**` is `auth-data` in the dubplate-build skill's risk map. Its test-first rule is met
by the doctor tests that exist today (98.5 percent of `newDoctorCmd`), since a change with no behavior
change has no new test that fails at the start commit.

Below, `B` is `$(git merge-base origin/master HEAD)`, and `U` is
`env $(env | grep -o '^DUBPLATE_[A-Z0-9_]*' | sed 's/^/-u /')`, which runs a command with every
`DUBPLATE_*` name unset.

Done when:

1. `golangci-lint run --no-config --default=none --enable=gocognit --tests=false --show-stats=false
   --output.text.print-issued-lines=false ./...` prints 12 lines today. After, it prints two, naming
   `commitVerdict` and `walkZIP`, plus one for each function the pull request's `## Decisions` names
   with its reason.
2. `grep -c 'text: func' .golangci.yml` prints `12` today, and after it prints item 1's line count.
   `golangci-lint run ./...` exits 0.
3. `git diff "$B" --numstat -- '*_test.go' | awk '$2 != 0'` prints nothing: no existing test line is
   changed or removed.
4. `git diff "$B" -U0 -- '*.go' ':!*_test.go' | grep -c -E '^[-+](func (\([^)]*\) )?|type |var |const )[A-Z]'`
   prints `0` today and after: no exported function, method, type, variable, or constant is added,
   removed, or re-signed.
5. At the commit before `gradeCaseDefault`'s restructure, `$U go test -coverprofile=c.out
   ./internal/harness/conform && go tool cover -func=c.out | grep gradeCaseDefault` prints 85.0% or more,
   or `## Decisions` gives the figure as "What happens when it goes wrong" says. It prints 61.5% today.
   The pull request quotes the line and the commit.
6. `$U go test -cover` for the ten packages prints, for each, a figure at or above the table's package
   coverage. The pull request quotes the ten lines.
7. `grep -c complexity-cleanup docs/STATUS.md` prints `0` today, and 1 or more after.
8. `bash scripts/gate.sh` passes from the worktree root, and the pull request quotes its `gate exit: 0`
   line.
9. CI passes. `gh pr checks <n>` lists `check` and `fuzz` as `pass`; `fuzz` runs because
   `internal/engine/tagio` changes.

The reviewer reads each restructure for the failure above: a helper that only makes sense beside its
caller, a name that says where code sat rather than what it does, or a control-flow change that alters
an error's wording or order.

Reviewers: `standards:diff-reviewer` for the branch, and `standards:web-auth-security-reviewer` for the
`newDoctorCmd` commit. Minutes: 200 (work 95, local gates 56 for two full runs at the
28-minute median, CI wait 46, review 15). The work figure scales pull request 25 (`structure-doctor`,
194 minutes actual), the nearest like task, from task-report's dubplate rows for 2026-10-24 to
2026-11-23. Time clause: 400 minutes. Model: Opus at high. Waits for: nothing.

Launch from `~/Projects/dubplate` by `building-from-specs` "Launching a task", as `complexity-cleanup`.
The goal names this spec, says to read it with `git -C /var/home/glw907/Projects/dubplate show
origin/master:docs/specs/2026-11-23-complexity-cleanup.md`, and is met when each done item holds as the
spec words it or is marked unmet with its reason, the `standards:diff-reviewer` review accepts, and the
pull request is ready with planPathMin 200.

## Merge order

One pull request. It touches no intake file, so `intake-2` can be specified beside it, and `intake-2`
builds after it merges.

## Backlog sweep

The scope is every open item that names a file the task changes, plus #126. Nothing folds: each fix
in those lines changes printed output, and this task changes none.

| Item | Mechanism today | Verdict |
| --- | --- | --- |
| #140 | `doctor.go:71` prints "disabled (arrives at rung %d)", asserted at `doctor_test.go:165`. | keep |
| #179 | `doctor.go:75` drops `class.Enabled`'s error, so a malformed flag prints "disabled". | keep |
| #164 | The two strings sit at `case.go:614` and `:641`, inside `gradeCaseDefault`. `component.go:380`'s beets claim is backed by `matcher/penalty.go:338`. | reword: "Reword the two `ungradedReason` strings in `gradeCaseDefault` to drop 'Resolution R1' and 'rung-3', and cite `matcher/penalty.go:338` for `conform/component.go:380`'s anchoring claim or drop 'as beets anchors it'." |
| #185 | Nothing in it is code the task changes. After the task, two exclusions remain. | reword at the merge: "12" becomes the count item 2 prints. The same edit points #140 at `newDoctorCmd`'s new lines. |
| #126 | I2 edits `zip.go:137`, inside `walkZIP`. I3 edits `accept.go:63`, inside `commitVerdict`. | keep |

The task's moves shift the line numbers in #140 and #179, so the merging session refreshes them.

## Calls Claude made

- `commitVerdict` and `walkZIP` go to `intake-2`, which edits them.
- The coverage bar is 85 percent of the function's statements, read before its restructure. Where
  fixtures can't reach it, uncovered blocks move unchanged instead of gaining a test seam.
- The architecture read running now (`docs/research/2026-11-23-architecture-read.md`) feeds a separate
  cleanup spec. This task doesn't wait on it, since the read grades package shape, not these
  functions' complexity.
- One task with one commit per function, in place of one task per package.
- Existing tests stay unedited, so they hold the behavior.
- Opus at high, since each restructure is a judgment on what a helper should hold.
