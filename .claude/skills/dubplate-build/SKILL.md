---
name: dubplate-build
description: >
  The dubplate facts for building work from a spec: gate, fast lane, full-suite home, worktree setup,
  CI watch, risk map, reviewers by path, and the coding conventions. Use when building, resuming, or
  reviewing a task in dubplate. Use it with `building-from-specs`, which owns the lifecycle.
---

# Dubplate build

`building-from-specs` owns the lifecycle: launch, goal, review, pull request, merge. This skill holds
the facts that process leaves to the repo. Paths are relative to the repo root. The binding
decisions are records in `docs/decisions/`, the backlog is in `BACKLOG.md`, and specs are under
`docs/specs/`.

## Gate

Run `bash scripts/gate.sh` from the worktree root with the Bash tool's `timeout: 600000`, and
re-issue it on exit 75 until it prints `gate exit:`. The wrapper builds the gate string from a fixed
flag set and hands it to `cairn-run-gate`, which keeps the lane lock and the run records. The full
gate launches Playwright's Chromium for the rendered leg, so it takes the heavy lane. Never commit
red. Never accept a task on a bare `bash scripts/check.sh`: CI runs that form by design, and the
gate is the form a task's evidence quotes. `scripts/check.sh`'s header lists every leg.

Two import-graph tests hold the OpenSubsonic layer's boundaries, and a green `go test ./...` runs
both. The import-boundary test (`internal/opensubsonic/boundary_test.go`) fails if anything under
`internal/opensubsonic` pulls in a module-internal package outside it, tests included. The no-link
test (`internal/harness/conform/arch_test.go`) fails if `go list -deps ./cmd/dubplate-server` names
a package under `internal/harness` or `internal/opensubsonic/conformance`, or any
`github.com/getkin` path, so the server never links the harness, the contract checker, or
kin-openapi. The protocol client stays linkable.

The process-citation leg fails on a shipped Go comment under `cmd` or `internal` that cites a rung,
task, resolution, or plan line. A wider pattern (contract, ruling, fold, pass, and conductor tags,
and a name) applies everywhere except `cmd/dubplate-server`, `internal/serve`, `internal/auth`, and
`internal/intake`, whose comments BACKLOG #156 clears.

Three workstation prerequisites sit outside the gate, because no test opens a network connection:
`npm ci --prefix web`, `make unidecode-wheel` (the verified Unidecode 1.4.0 wheel in the user cache,
which the unidecode tests and the install verb's tests read), and `make subsonic-xsd` (the pinned
Subsonic 1.16.1 XSD in the user cache, read by the harness's XML tier and never committed). The XML
tier also needs `xmllint` at libxml 2.15.4 on `PATH`, held by `brew pin libxml2` because its
messages land in goldens. Eight installed formulae depend on it and are held with it: chafa, ffmpeg,
gdk-pixbuf, imagemagick, librsvg, libxkbcommon, shared-mime-info, and xdotool.

A libxml2 bump (record 0004) is `brew unpin libxml2 && brew upgrade libxml2`, then regenerate the XML
goldens, `brew pin libxml2`, and upgrade the eight held dependents.

## Fast lane

Use `bash scripts/gate.sh --scope-changed` between steps. On a branch it measures from the merge
base with master. `--dry-run` prints the scope and runs nothing.

- An empty change set falls back to the full form.
- `check.sh` classes `go.mod`, `go.sum`, `web/package*.json`, `scripts/check-state-dir.sh`, and
  `Makefile` as gate machinery, and a scoped run that touches one becomes full. A task whose first
  step moves one of them (a dependency sweep) commits that step, then passes `--base <sha>` on every
  later fast lane, where `<sha>` is the full 40-hex id of the commit that ends the sweep step. Later
  scoped runs then exclude the sweep. `check.sh` diffs the working tree against the base, so a base
  from before the sweep keeps `go.mod` in the change set and every fast lane runs full.
- `gate.sh` refuses a `--base` that is not exactly 40 lowercase hex characters, prints its usage, and
  exits 2. `check.sh` uses a 40-hex base only when it resolves to itself and is an ancestor of
  `HEAD`. Otherwise the run promotes to full and its `scope=` line names the reason. A later commit
  that is still an ancestor of `HEAD` passes validation and narrows the scope, so a wrong base can
  hide a change from a fast lane. `gate.sh --merge` before the pull request is ready runs the full
  form and restores the coverage.

## Full-suite home

CI. `.github/workflows/check.yml` runs `bash scripts/check.sh` on every push to a pull request, then
the race detector as its "Run the race detector" step (`make test-race`), which no local gate runs.
`.github/workflows/fuzz.yml` runs `make fuzz FUZZTIME=1m` on a pull request that touches
`internal/engine/tagio`, `internal/intake/gate`, the workflow, or the `Makefile`. Before
a pull request is marked ready, run `bash scripts/gate.sh --merge` locally: it is the full gate plus
one gitleaks scan over `master..HEAD` with git history, the only leg CI cannot run. The scoped form
narrows gitleaks to the change set, so a secret one commit added and a later one removed is visible
only to `--merge`.

When the task touches `cmd/harness` or `internal/harness`, also run
`go vet -tags oracle ./cmd/harness/... ./internal/harness/...`. The gate never builds the
`//go:build oracle` files, so nothing else reads them.

`RUN_GATE_IF_BUSY=defer` is allowed except on an `auth-data` task.

## Worktree setup

Run `npm ci --prefix web`, then `(cd web && npx svelte-kit sync)`, in the new worktree before the
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
- `*.go`: the `go-conventions` skill, with record 0002's bar an acceptance dimension.

Web UI work matches no streamer: dubplate replaces Navidrome at 0.8 and carries its own visual
identity (Geoff, 2026-09-24), so there are no reference screenshots and no visual-fidelity chain. A
task's review fans out by surface and passes the visual identity spec's path to the `web/`
reviewers. The `intake` uploader is Geoff's design review.

Every user-path page is rendered and read before any attended ask (Geoff, 2026-09-02). Never ask
Geoff to test a surface you have not seen, the auth pages in front of the app included, starting
from the URL the person will type. A status code is not evidence;
a screenshot read in the main loop is. A deploy once went to Geoff after only a 302 was checked, and
the login page behind it carried another organization's branding.

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
  import-graph tests ("Gate") holds each.
- Exports follow record 0003.
- Config keys follow record 0010; `newProductionRegistry` is in `cmd/dubplate-server/main.go`.
- Skills. Go work runs under `go-conventions`. `web/` work runs under `svelte-conventions`,
  `ts-conventions`, `daisyui`, and the visual identity spec, which every `web/` task reads.
- Copy words follow record 0006.
- Errata. A frozen, ratified spec gets no errata, and a superseded one gets at most a one-line
  header pointer. A living reference is edited in place by the task that changes what it says. The
  living references are the streamer spec's forward hooks, the visual identity spec, the uploader
  copy contract (`docs/research/passb-copy-final.md`),
  `docs/specs/2026-09-11-configuration-homes-design.md`,
  `docs/specs/2026-09-25-group-scale-principles-design.md`,
  `docs/specs/2026-09-12-web-surface-standard.md`, and `docs/roster-walk.md`. No list of owed edits
  exists.
- A spec closes four gaps before approval. Each stdlib or library behavior a task relies on is
  quoted from `go doc` or the source, never recalled. Each state a check can meet (absent, empty,
  malformed, many-bad) names its report. Each privileged field names every channel that may set it.
  Each proof names the fixture state it needs. A stale spec is the dominant defect class, so a spec is
  read against the code as shipped before a task launches from it.
- Dependency sweep. Record 0004 holds the policy and when the sweep runs. The sweep is a
  `dependency-upgrade` run over `go.mod`, `web/package.json`, `tools/venv/requirements.txt`, and the
  pinned binaries. The binaries' pins live in the `env:` block of `.github/workflows/check.yml`
  (rclone, golangci-lint, shellcheck, ruff, gitleaks, vale, libxml2) and in `web/package.json`
  (Playwright). The libxml2 bump steps are under "Gate". The same task runs
  `make fuzz FUZZTIME=10m`.
  It also takes any open Dependabot pull request (`.github/dependabot.yml` watches the GitHub Actions
  pins alone) through `standards:dependency-upgrade`.
- Product sittings. Geoff takes three, and no others: corpus adjudication batches, the `intake`
  uploader sitting, and the first-launch sitting.
