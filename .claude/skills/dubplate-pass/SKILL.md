---
name: dubplate-pass
description: >
  The dubplate adapter for the lean pass process: gate, fast lane, full-suite home, worktree setup,
  CI watch, risk-class path map, checklist globs, and close checklist. Use when planning, executing,
  resuming, or closing a dubplate pass, or when a fresh dubplate session follows a STATUS launch
  prompt.
---

# Dubplate pass

Load `pass-core` first; it holds the lifecycle and the rules this adapter fills in. Paths are
relative to the repo root. The resume prompt is in `docs/STATUS.md` and plans live under
`docs/plans/`. The repo's standing rules are in `docs/standing-rules.md`.

## Rules the pass carries

- **One executor per tree.** Before any executor touches the repo or the box, run
  `claude-worktree-busy <path>` on the tree, then `claude agents --json` to list the running
  sessions, then `pgrep -f` with a bracket-class pattern (for example
  `pgrep -f '[P]rojects/dubplate/.claude/worktrees/<name>'`, so the check does not match itself) for
  a process in the same worktree, then `git status` for warm changes you did not author. STATUS
  records each running pass's session name and id. The box (`5.78.148.42`) is production: no writes
  without a quiet-box check, and only read-only ssh during reviews.
- **Mutation proofs are run, never narrated.** A mutation described in a comment beside the
  assertion it defends proves nothing: two of six such mutations in one task did not fire, and no
  reader could tell from the commentary. Run each mutation and record that the test failed.
- **Read a plan against the shipped code before dispatching it.** A fresh-context read of the
  plan's tasks against the machinery as shipped found five blockers before any dispatch; a stale
  plan is the dominant defect class.
- **Concurrent passes.** Passes run inside a round, one at a time unless the round design pairs
  them. Each design names the files it touches, so two passes never edit the same one. With two
  PRs open, the second to merge first merges master and reruns its close gate.
- **Plans are just-in-time.** Plan each pass against a tree no other executor is writing, never
  ahead of the work: plan staleness is the proven dominant defect class.
- **Errata.** Frozen, ratified specs get no errata, and a superseded one gets at most a one-line
  header pointer. A living reference (the streamer spec's forward hooks, PCF, the visual identity
  spec, CH, W, PS, `docs/roster-walk.md`) is edited in place by the pass that changes what it says,
  as a task in that pass. There is never an owed list.
- **Carried items.** A round's design reads BACKLOG's items tagged `#round:<its round>` and the
  "Rounds to 0.8" section of BACKLOG. A pass's plan reads the items tagged `#pass:<its pass>`.
- **Planning-miss prevention.** Every plan closes four gaps before approval. Each stdlib or library
  behavior a task relies on is quoted from `go doc` or the source, never recalled. Each state a
  check can meet (absent, empty, malformed, many-bad) names its report. Each privileged field names
  every channel that may set it. Each proof names the fixture state it needs (selected, populated)
  so the test cannot pass vacuously. The plan-review lenses check these four.
- **Findings are ranked, and a fold may refuse.** A reviewer asked for gaps will report some even
  when the work is sound, and chasing every finding over-engineers. Every reviewer dispatch asks for
  correctness gaps ranked by consequence, not polish. A fold may refuse a finding whose fix costs
  more than the risk it removes, with a one-line reason in the fold record.
- **Product touchpoints, and only these:** corpus adjudication batches, the `intake` uploader
  sitting, and the first-launch sitting. The design approval and the packet read are `pass-core`'s.
- **Dependency sweep.** The sweep runs once a round, as the first task of the round's first pass
  (finding G14), as a `dependency-upgrade` sweep over `go.mod`, `web/package.json`,
  `tools/venv/requirements.txt`, and the pinned binaries. The binaries' pins live in the `env:`
  block of `.github/workflows/check.yml` (rclone, golangci-lint, shellcheck, ruff, gitleaks, vale,
  libxml2) and in `web/package.json` (Playwright). A sweep with nothing to move is one line in the
  PR body. Minors and patches land without asking; a major waits for Geoff. Full text and the
  libxml2 bump steps: `docs/standing-rules.md`.
- **Web review passes the visual identity spec.** Dispatches to the `web/` reviewers carry
  `docs/specs/2026-09-24-visual-identity-design.md` and ask for conformance to it
  (tokens, inks, the component layer, the copy register) graded beside correctness.
- **Pair subagents** are `general-purpose`. The dispatch gives the adapter's worktree setup and fast
  lane (`bash scripts/gate.sh --scope-changed`). For Go it names the `go-conventions` skill and
  makes "idiomatic, architecturally beautiful Go" an acceptance dimension. For `web/` it names
  `svelte-conventions`, `ts-conventions`, and the visual identity spec.

## Conventions every task follows

A plan cites this section instead of restating it (G15). Each subagent dispatch carries it.

### Gate

Run `bash scripts/gate.sh` (full) or the fast lane from the worktree root, with the Bash tool's
`timeout: 600000`, and re-issue on exit 75 until it prints `gate exit:`. A worktree session's Bash
guard refuses the quoted `cairn-run-gate '<string>'` form and accepts the wrapper. Never fall back
to bare `bash scripts/check.sh` for an acceptance; CI runs it bare by design.

### The base, and why the fast lane names it

`check.sh` classes `go.mod`, `go.sum`, `web/package*.json`, `scripts/check-state-dir.sh`, and
`Makefile` as gate machinery, which promotes a scoped run to full. The pass whose first task is the
round's sweep passes `--base <task start>` (a full 40-hex id) on every later fast lane. `check.sh`
uses a base only when it resolves to itself and is an ancestor of `HEAD`, so a wrong value costs
time, never coverage. Other passes pass no `--base`.

### Setup

Run `npm ci --prefix web` in the session's worktree and in each subagent's worktree before its
first gate.

### Test-first

Every `auth-data` task writes its acceptance tests first and records each one failing at the task's
start commit in the PR notes.

### Mutation proofs are run

Apply each mutation one at a time, run the named test, and record the failure in the PR notes with
the test name. A mutation that does not fail is a finding, fixed in the same task. Narrate none in
a comment.

### Fixture states

Every proof names the fixture state it needs (populated, selected, renamed, removed). A `web/` proof
is a named rendered row (`web/scripts/rendered/run.ts`, host `cmd/dubplate-server/rendered_test.go`)
that says what it intercepts and what it reaches for real; `web/` has no component test runner.

### Layering

The plan names each import edge its packages must not gain, and a test or the two import-graph
tests (`docs/standing-rules.md`, "Gate legs") holds each.

### No export ahead of a caller

Every new export names its production caller in the task that adds it.

### Config keys

Declare each key through the registry from the pass's own file, with one `newProductionRegistry`
line (`cmd/dubplate-server/main.go:99`). Never edit `internal/config/keys.go`.

### Skills

Go work runs under `go-conventions`, with "idiomatic, architecturally beautiful Go" an acceptance
dimension of every Go task. `web/` work runs under `svelte-conventions`, `ts-conventions`,
`daisyui`, and the visual identity spec, which every `web/` task reads. Copy says "the group" and
"a member".

### Per-task review

On each `auth-data` task, run `diff-reviewer` at task scope on the task's commit or the merged
pair's range. `accept` continues. `fix` gets one fix commit quoting each finding verbatim, then one
re-read; a second `fix` or any `escalate` is a stop. Each verdict appends
`review T<n>=<verdict> <sha>` to the PR notes.

### Commits

One commit per task, the task id first in the subject, imperative mood, specific files only, with
the session's trailer. A pair half merges with `git merge --no-edit <branch>`.

### Records

PR notes live at `<git dir>/pr-notes.md`, never committed, and survive a restart. The first line is
`launch-base=<sha>`, the branch head at the executor's start. Before each dispatch the session
appends `dispatch T<a>[+T<b>] sha=<40-hex>`. Each task appends its test-first failures, mutation
results, review line, and out-of-scope finds. A subagent writes nothing outside its worktree and
returns its records in its report.

### Clock

The clock runs from the first execution commit to PR-ready (`pass-core` `executor.md`, "Score"). A
task past twice its estimate gets one `## Decisions` line with the estimate and the actual, and the
run goes on.

### Dispatches

A pair is two `general-purpose` Agent subagents with `isolation: "worktree"`, dispatched after the
session commits, `sonnet` at `medium` unless the plan names `opus` for novel correctness-critical
logic. Each gets its task's full text, the plan sections it cites, and this section. Its first act
is `git merge-base --is-ancestor <dispatch sha> HEAD`, a failure being a stop. It commits and never
pushes. The session merges, runs the fast lane, runs any per-task review, and pushes.

### Stops

A stop is a question `advisor.md` "What always stops" lists, or a pass with no task left. Commit and
push the work, open the draft PR if none exists, write the stop under `## Stop` at the end of
`docs/STATUS.md` on the branch (the question, what was tried, the choices), push, and end. Never ask
Geoff in chat, never run `notify-send`, and never work around a refusal with another tool. End
background commands through `TaskStop`.

### Resume

The same session, or a fresh one launched with the same prompt, enters the worktree with
`EnterWorktree` `path`. It reads its task commits with
`git log --no-merges --format='%h %s' <launch-base>..HEAD`, settles the last `dispatch` line by the
worktree list (merged, merge now, or a stop for a candidate with no `T<n>:` commit), and resumes at
the first task whose acceptance fails. Never re-apply a committed task.

### Scope

A defect found outside a task goes to BACKLOG through `standards:log-issue` and the PR body's
"Found, not done".

## Adapter

**Gate**: `bash scripts/gate.sh`, which runs `check.sh` through `cairn-run-gate`. The full gate
launches Playwright's Chromium for the rendered leg, so it takes the heavy lane. Never commit red.

**Fast lane**: `bash scripts/gate.sh --scope-changed`. On a branch it measures from the merge base
with master. A pass whose first task moves `go.mod`, `go.sum`, or `web/package*.json` passes
`--base <task start>`, the 40-hex id of the commit the task started from, so later scoped runs
exclude the sweep; any other pass passes no `--base`. A `--base` that is not a full 40-hex commit
id silently runs the full form. An empty change set, or any path under `scripts/checks/`, also falls
back to the full form. `--dry-run` prints the scope and runs nothing.

**Full-suite home**: CI. The PR check (`.github/workflows/check.yml`) runs `bash scripts/check.sh`
on every push to a PR. The close still runs `bash scripts/gate.sh --merge` locally, for the
branch-history gitleaks scan, the one leg CI cannot run. When the
pass touched `cmd/harness` or `internal/harness`, also run
`go vet -tags oracle ./cmd/harness/... ./internal/harness/...`, since the gate never builds the
`//go:build oracle` files. `RUN_GATE_IF_BUSY=defer` is allowed except on an `auth-data` task.

**Worktree setup**: `npm ci --prefix web` in the new worktree; the gate refuses a worktree without
`web/node_modules` (`make web: web/node_modules absent`). The beets venv (`tools/venv/`) is
gitignored and `make derive` sits outside the gate, so a worktree's gate goes green without the one
leg that exercises beets. Work that needs the venv, `make derive`, or `corpus/` (the matcher corpus
and the beets oracle) stays in the main tree. The OpenSubsonic harness (`internal/harness/subsonic`,
`internal/opensubsonic`, `cmd/harness`) needs no venv, no beets, and no `corpus/`, so its pass runs
in a worktree.

**CI watch**: After each push, run `gh pr checks <n> --watch --fail-fast` as a background Bash
task, bare. The fallback is `gh run watch <run id> --exit-status`.

**Risk-class path map**: `auth-data` for `internal/auth/**`, `internal/serve/**`,
`internal/engine/store/migrations/**`, `cmd/dubplate-server/**`, `internal/intake/**`,
`web/src/routes/{login,members}/**`, and `systemd/**`. `runner` for `scripts/`, `.claude/`, and `.github/`. `ordinary`
otherwise. `internal/intake/**` joined on 2026-10-12: the tus wrapper, the gate over member-supplied bytes, and the
worker are the upload path's member gate, and the `intake-1` security reads found what four task reviews had not.

**Checklist globs**: The close review loads `standards:svelte-reviewer` and `standards:daisyui-a11y-reviewer` for
`web/**`; `standards:web-auth-security-reviewer` for every `auth-data` path above; and the `go-conventions`
skill for `*.go`.

**Close checklist**: Beyond `pass-core` "Close":

- The full local gate on the close head, with the harness `go vet` line above where it applies.
- The scoped form narrows gitleaks to the change set, so the close owes one branch-history scan.
  Run `bash scripts/gate.sh --merge` as the close run: it is the full gate
  plus one gitleaks pass over `master..HEAD` with git history, the only leg that sees a secret one
  task committed and a later one rewrote.
- `docs/STATUS.md` in present tense, 60 lines or fewer. A STATUS rewrite moves each item to BACKLOG
  and never drops an open decision.
- Archive the plan with `git mv docs/plans/<this-pass>.md docs/archive/plans/`.
- The PR body carries the score.
- Commit specific files only, one commit per task with the task id first in the subject.
