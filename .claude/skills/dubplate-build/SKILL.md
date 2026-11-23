---
name: dubplate-build
description: >
  The dubplate facts for building work from a spec: gate, fast lane, full-suite home, worktree setup,
  CI watch, risk map, reviewers by path, and the coding conventions. Use when building, resuming, or
  reviewing a task in dubplate, or when a fresh dubplate session follows a STATUS launch prompt. Use
  it with `building-from-specs`, which owns the lifecycle.
---

# Dubplate build

`building-from-specs` owns the lifecycle: launch, goal, review, pull request, merge. This skill holds
the facts that process leaves to the repo. Paths are relative to the repo root. The standing rules
are in `docs/standing-rules.md`, the backlog in `BACKLOG.md`, and specs under `docs/specs/`.

## Gate

Run `bash scripts/gate.sh` from the worktree root with the Bash tool's `timeout: 600000`, and
re-issue it on exit 75 until it prints `gate exit:`. The wrapper builds the gate string from a fixed
flag set and hands it to `cairn-run-gate`, which keeps the lane lock and the run records. The full
gate launches Playwright's Chromium for the rendered leg, so it takes the heavy lane. Never commit
red. Never accept a task on a bare `bash scripts/check.sh`: CI runs that form by design, and the
gate is the form a task's evidence quotes.

## Fast lane

Use `bash scripts/gate.sh --scope-changed` between steps. On a branch it measures from the merge
base with master. `--dry-run` prints the scope and runs nothing.

- An empty change set falls back to the full form.
- `check.sh` classes `go.mod`, `go.sum`, `web/package*.json`, `scripts/check-state-dir.sh`, and
  `Makefile` as gate machinery, and a scoped run that touches one becomes full. A task whose first
  step moves one of them passes `--base <sha>` on every later fast lane, where `<sha>` is the full
  40-hex id of the commit the task started from. `check.sh` uses a base only when it resolves to
  itself and is an ancestor of `HEAD`, so a wrong value costs time and never coverage. A shorter id
  silently runs the full form.

## Full-suite home

CI. `.github/workflows/check.yml` runs `bash scripts/check.sh` on every push to a pull request. Before
a pull request is marked ready, run `bash scripts/gate.sh --merge` locally: it is the full gate plus
one gitleaks scan over `master..HEAD` with git history, the only leg CI cannot run. The scoped form
narrows gitleaks to the change set, so a secret one commit added and a later one removed is visible
only to `--merge`.

When the task touches `cmd/harness` or `internal/harness`, also run
`go vet -tags oracle ./cmd/harness/... ./internal/harness/...`. The gate never builds the
`//go:build oracle` files, so nothing else reads them.

`RUN_GATE_IF_BUSY=defer` is allowed except on an `auth-data` task.

## Worktree setup

Run `npm ci --prefix web`, then `cd web && npx svelte-kit sync`, in the new worktree before the
first gate. The gate refuses a worktree without `web/node_modules` (`make web: web/node_modules
absent`), and `npm run check --prefix web` fails with `File '$app/tsconfig' not found` until
`svelte-kit sync` has generated `.svelte-kit`.

The beets venv (`tools/venv/`) is gitignored and `make derive` sits outside the gate, so a worktree's
gate goes green without the one leg that exercises beets. A task that needs the venv builds it in the
worktree with the two commands in `tools/venv/README.md`
(`uv venv --python 3.13 tools/venv/.venv`, then
`uv pip install --python tools/venv/.venv/bin/python -r tools/venv/requirements.txt`) and runs
`make derive` there. `corpus/` is tracked, so the worktree has it. The OpenSubsonic harness
(`internal/harness/subsonic`, `internal/opensubsonic`, `cmd/harness`) needs no venv, no beets, and no
`corpus/`.

## CI watch

After each push, run `gh pr checks <n> --watch --fail-fast` as a background Bash task, bare. The
fallback is `gh run watch <run id> --exit-status`.

## Risk map

`auth-data`: `internal/auth/**`, `internal/serve/**`, `internal/webapp/**`, `internal/ops/**`,
`internal/store/migrations/**`, `cmd/dubplate-server/**`, `internal/intake/**`,
`web/src/routes/{login,members}/**`, and `systemd/**`. `internal/ops/**` holds the `/api/v1`
authorization layer (`opshttp`) and the admin seam that writes roster audits. `internal/intake/**` is
the tus wrapper, the gate over member-supplied bytes, and the worker: the upload path's member gate.

`runner`: `scripts/`, `.claude/`, and `.github/`. `ordinary`: everything else.

The box (`5.78.148.42`) is production: no writes without a quiet-box check, and only read-only ssh
during reviews.

## Reviewers by path

`standards:diff-reviewer` reads every task. Add these by path:

- `web/**`: `standards:svelte-reviewer` and `standards:daisyui-a11y-reviewer`. Their briefs carry
  `docs/specs/2026-09-24-visual-identity-design.md` and ask for conformance to it (tokens, inks, the
  component layer, the copy register), graded beside correctness.
- Every `auth-data` path above: `standards:web-auth-security-reviewer`.
- `*.go`: the `go-conventions` skill, with "idiomatic, architecturally beautiful Go" an acceptance
  dimension.

Ask every reviewer for correctness gaps ranked by consequence, never polish. A reviewer asked for gaps
reports some even when the work is sound, and chasing every one over-engineers. A fix may refuse a
finding whose cost exceeds the risk it removes, with a one-line reason in the pull request's
`## Decisions`.

## Conventions

- Test-first on `auth-data` work. Write the acceptance tests first and record each one failing at
  the task's start commit in the pull request.
- Mutation proofs are run. Apply each mutation one at a time, run the named test, and record that
  it failed, with the test's name. A mutation that does not fail is a finding, fixed in the same task.
  A mutation described in a comment proves nothing: two of six such mutations in one task never fired,
  and no reader could tell from the commentary.
- Fixture states. Every proof names the state it needs (populated, selected, renamed, removed), so
  it cannot pass vacuously. A `web/` proof is a named rendered row (`web/scripts/rendered/run.ts`,
  host `cmd/dubplate-server/rendered_test.go`) that says what it intercepts and what it reaches for
  real. `web/` has no component test runner.
- Layering. A spec names each import edge its packages must not gain, and a test or the two
  import-graph tests (`docs/standing-rules.md`, "Gate legs") holds each.
- No export ahead of a caller. Every new export names its production caller in the task that adds
  it.
- Config keys. Declare each setting in `config.go` of the package that reads it and add it to that
  package's `Register` function. `newProductionRegistry` in `cmd/dubplate-server/main.go` gains a call
  only when a new package registers for the first time. Never edit `internal/config/keys.go`.
  `internal/settings/layering_test.go` fails when any other file in those packages imports
  configuration.
- Skills. Go work runs under `go-conventions`. `web/` work runs under `svelte-conventions`,
  `ts-conventions`, `daisyui`, and the visual identity spec, which every `web/` task reads. Copy says
  "the group" and "a member".
- Errata. A frozen, ratified spec gets no errata, and a superseded one gets at most a one-line
  header pointer. A living reference (the streamer spec's forward hooks, the visual identity spec,
  `docs/roster-walk.md`) is edited in place by the task that changes what it says. No list of owed
  edits exists.
- A spec closes four gaps before approval. Each stdlib or library behavior a task relies on is
  quoted from `go doc` or the source, never recalled. Each state a check can meet (absent, empty,
  malformed, many-bad) names its report. Each privileged field names every channel that may set it.
  Each proof names the fixture state it needs. A stale spec is the dominant defect class, so a spec is
  read against the code as shipped before a task launches from it.
- Dependency sweep. The first task of each milestone's first spec is a `dependency-upgrade` sweep
  over `go.mod`, `web/package.json`, `tools/venv/requirements.txt`, and the pinned binaries. The
  binaries' pins live in the `env:` block of `.github/workflows/check.yml` (rclone, golangci-lint,
  shellcheck, ruff, gitleaks, vale, libxml2) and in `web/package.json` (Playwright). A sweep with
  nothing to move is one line in the pull request body. Minors and patches land without asking, and a
  major waits for Geoff. Full text and the libxml2 bump steps: `docs/standing-rules.md`.
- Product sittings. Geoff takes three, and no others: corpus adjudication batches, the `intake`
  uploader sitting, and the first-launch sitting.
