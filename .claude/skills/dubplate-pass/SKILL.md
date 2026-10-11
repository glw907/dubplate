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
`docs/superpowers/plans/`. The repo's standing rules are in `docs/standing-rules.md`.

## Rules the pass carries

- **One executor per tree.** Before any executor touches the repo or the box, check the lane
  registry for a live executor in the same tree and run `git status` for warm changes you did not
  author. `pgrep -f Projects/dubplate` matches every lane once a worktree exists, so the registry is
  the check. The box (`5.78.148.42`) is production: no writes without a quiet-box check, and only
  read-only ssh during reviews.
- **Lanes.** A lane is one rung's execution with a named package territory. Document lanes and the
  primary code lane share the main tree; each further code lane takes its own worktree, because the
  gate runs the module over `./...`. Up to three code lanes run at once (Geoff, 2026-10-05).
- **Plans are just-in-time.** Plan each lane's next rung against a tree no other executor is
  writing, never ahead of the work: plan staleness is the proven dominant defect class.
- **Planning-miss prevention.** Every plan closes four gaps before approval. Each stdlib or library
  behavior a task relies on is quoted from `go doc` or the source, never recalled. Each state a
  check can meet (absent, empty, malformed, many-bad) names its report. Each privileged field names
  every channel that may set it. Each proof names the fixture state it needs (selected, populated)
  so the test cannot pass vacuously. The plan-review lenses check these four.
- **Findings are ranked, and a fold may refuse.** A reviewer asked for gaps will report some even
  when the work is sound, and chasing every finding over-engineers. Every reviewer dispatch asks for
  correctness gaps ranked by consequence, not polish. A fold may refuse a finding whose fix costs
  more than the risk it removes, with a one-line reason in the fold record.
- **Owner touchpoints, and only these:** rung-spec rulings, corpus adjudication batches, the rung 7
  uploader before and after, and the first-launch sitting.
- **Dependency sweep.** Every pass opens with a `dependency-upgrade` sweep over `go.mod`,
  `web/package.json`, `tools/venv/requirements.txt`, and the pinned binaries (rclone,
  golangci-lint, Playwright, libxml2). A sweep with nothing to move is one line in the PR body.
  Minors and patches land without asking; a major waits for Geoff. Full text and the libxml2 bump
  steps: `docs/standing-rules.md`.
- **Web review passes the visual identity spec.** Dispatches to the `web/` reviewers carry
  `docs/superpowers/specs/2026-09-24-visual-identity-design.md` and ask for conformance to it
  (tokens, inks, the component layer, the copy register) graded beside correctness.
- **Pair subagents** use the workstation `site-implementer` definition, or `general-purpose` with the
  `go-conventions` skill named in the dispatch.

## Adapter

**Gate**: `cairn-run-gate 'bash scripts/check.sh'`. The full gate launches Playwright's Chromium for
the rendered leg, so it takes the heavy lane. Never commit red.

**Fast lane**: `bash scripts/check.sh --scope-changed` through `cairn-run-gate`. Pass no `--base`:
on a branch it measures from the merge base with master. A `--base` that is not a full 40-hex commit
id silently runs the full form. An empty change set, or any path under `scripts/checks/`, also falls
back to the full form. `--dry-run` prints the scope and runs nothing.

**Full-suite home**: Local, because the repo has no PR CI (ruling 11). Run one `bash scripts/check.sh`
on the close head. When the pass touched `cmd/harness` or `internal/harness`, also run
`go vet -tags oracle ./cmd/harness/... ./internal/harness/...`, since the gate never builds the
`//go:build oracle` files.

**Worktree setup**: `npm ci --prefix web` in the new worktree; the gate refuses a worktree without
`web/node_modules` (`make web: web/node_modules absent`). The beets venv (`tools/venv/`) is
gitignored and `make derive` sits outside the gate, so a worktree's gate goes green without the one
leg that exercises beets. Work that needs the venv, `make derive`, or `corpus/` (the matcher corpus
and the beets oracle) stays in the main tree. The OpenSubsonic harness (`internal/harness/subsonic`,
`internal/opensubsonic`, `cmd/harness`) needs no venv, no beets, and no `corpus/`, so its lane runs
in a worktree.

**CI watch**: None. There is no PR CI, so no watch command runs.

**Risk-class path map**: `auth-data` for `internal/auth/**`, `internal/serve/**`,
`internal/engine/store/migrations/**`, `cmd/dubplate-server/**`, `web/src/routes/{login,members}/**`,
and `systemd/**`. `runner` for `scripts/` and `.claude/`. `ordinary` otherwise.

**Checklist globs**: The close review loads `svelte-reviewer` and `daisyui-a11y-reviewer` for
`web/**`; `web-auth-security-reviewer` for every `auth-data` path above; and the `go-conventions`
skill for `*.go`.

**Close checklist**: Beyond `pass-core` "Close":

- The full local gate on the close head, with the harness `go vet` line above where it applies.
- The scoped form narrows gitleaks to the change set, so the close owes one branch-history scan.
  Run `bash scripts/check.sh --merge` through `cairn-run-gate` as the close run: it is the full gate
  plus one gitleaks pass over `master..HEAD` with git history, the only leg that sees a secret one
  task committed and a later one rewrote.
- `docs/STATUS.md` in present tense, 60 lines or fewer.
- Archive the plan with `git mv docs/superpowers/plans/<this-pass>.md docs/superpowers/archive/plans/`.
- The PR body carries the score.
- Commit specific files only, one commit per task with the task id first in the subject.
