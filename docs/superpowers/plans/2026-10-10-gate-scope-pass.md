# Gate Scope Pass Plan

Status: **revision 1, 2026-10-10.** Geoff approved the design on 2026-10-10. It came from the cairn-cms gate-economy
pass's lessons: an unclassifiable case runs more, never less; a task that touches the gate machinery never sizes its
own gate; the type check stays in every per-task gate; Go's test cache sizes the test leg; and every fails-if claim is
a mutation with its red quoted.

**Goal.** Size each task's gate to the diff since a runner-recorded base commit. The classification fails closed:
any path it cannot place runs the full form. The rendered leg is selected by reverse dependency, Go's test cache sizes
the conform suite, and `scripts/checks/gate-tier.mjs` lets `pass-execute` use all of it.

**Pass class:** `tool` for every task. **Ceiling:** 2.5M tokens, flag 2.0M (80%). **Clock estimate:** about 2 hours.
That covers five full gates at about 9 minutes, two scratch scoped runs, and the dry-run proofs, which take seconds.
**Checkpoint interval:** four tasks; the one checkpoint falls after Task 4, before the close. **Segments:** one
(Tasks 1 to 4, then the close). Its boundary gate yields to the close gate.

**Tree.** The pass runs on master in the main tree as lane 1, the primary code lane, which is idle. This departs from
pass-core's "never the default branch's checkout" rule for three reasons:

- CLAUDE.md puts the primary lane in the shared tree.
- Switching that tree to a branch would capture the commits of the rung 7 planning lane and the rung 14a fold, which
  both write there.
- dubplate has no CI for a draft PR to shadow, and a worktree would add the merge ritual's two full gates.

Every task commits only after a full gate passes, so master never goes red. Update the lane registry row at launch.

**Per-task gate (fixed):** `cairn-run-gate 'bash scripts/check.sh'` on the heavy lane (no `CAIRN_GATE_LANE`). On exit
75, re-issue the same call until it prints `gate exit:`. Every task edits the gate machinery, which is the risk a
computed tier misses. Hand dispatches name the heavy lane, because `pass-execute`'s `tool` class defaults to light
(`pass-execute.js:334-340`). Before each gate, run the receipt lookup. The pass has fewer than six tasks, so each runs
the Agent-tool chain: `dubplate-implementer` (Sonnet), then `diff-reviewer` on `claude-opus-5-5`, one fix round, all
tasks sequential (Tasks 1 to 3 share `check.sh`). Before Task 1, one Haiku pre-flight re-checks this plan's code
claims at HEAD (pass-core).

**Test mandate (replaces the `tool` row's `go-conventions`).** CLAUDE.md's Shell rule bars behavior tests for scripts,
so no task commits a test. Each fails-if claim is instead a recorded run in a scratch worktree outside the repo
(`git worktree add <scratchpad>/<name> HEAD`, removed afterward). The run states the planted change and quotes two
lines: the scope line under the task's code, and the red line under a named mutation of the rule being tested. The
mutation is committed in the scratch tree below the base, so the planted change is the only path in the change set.
The record goes under "Record" below. Both `check.sh` and the shim are shellcheck clean or `node --check` clean.

**Task 0 (conductor, on master).** Run the `dependency-upgrade` sweep over `go.mod`, `web/package.json`,
`tools/venv/requirements.txt`, and the pinned binaries, per CLAUDE.md. If nothing moves, record that in one line. If
anything moves, it lands as its own full-gated commit before Task 1. Rung 14a's own Task 0 then finds the tree current.

## What `check.sh` reads today (evidence)

- The scoped base is `HEAD~1` on master and `git merge-base master HEAD` elsewhere (`check.sh:177-193`). A
  multi-commit task on master is scoped to its last commit only.
- The conform scope is a hand list (`:223-235`). So is the rendered scope (`:240-256`), which names three packages and
  seven serve-shell files.
- The rendered leg's test binary links 14 module packages. The command is `go list -deps -test -tags rendered -f
  '{{if not .Standard}}{{.Dir}}{{end}}' ./cmd/dubplate-server`, run at `dd161b3`: `cmd/dubplate-server`,
  `internal/{atomicfile, auth, backup, config, engine/matcher/unidecode, engine/store, identity, notify, ops,
  ops/local, serve, settings}`, `web`. The hand list misses nine of them.
- `internal/ops/opshttp` is not in that set until rung 14a mounts it.
- `go help list`: "-test flag causes list to report not only the named packages but also their test binaries"; "-deps
  flag causes list to iterate over not just the named packages but also all their dependencies."
- The only embeds are `web/embed.go` (`//go:embed all:build`, built from `web/`) and
  `internal/engine/store/migrations.go` (`migrations/*.sql`). Mapping each file to its nearest enclosing package
  covers both.
- With conform and rendered both in scope, the scoped form runs `make web` and then `make check`, so it builds the
  bundle twice (`:332-335`, `:371-374`; rung 14a plan, lane launch, step 8).
- An empty change set reads as a narrowed scope with `files=0` and goes green.

## Global constraints

- Classification lives in `check.sh` only. The shim chooses no tier and no lane.
- A fallback promotes the run to the full form and prints one line naming the reason. `--merge` and the bare full
  form are unchanged.
- `--scope-changed` and `--reduced` share one derivation and one class table. Each form prints its scope line as it
  does today, now carrying the base, where the base came from (caller, merge base, or `HEAD~1`), and any fallback
  reason.
- Any `go list` or `git` failure during classification promotes the run to full.

## The class table (Task 1 lands it; Tasks 2 and 3 fill in the effects)

Rows are matched in this order. The table is the acceptance criterion; the paths come from `git ls-files` top-level
counts and the legs in `check.sh` and the Makefile.

| Class | Paths | `--scope-changed` | `--reduced` |
| --- | --- | --- | --- |
| Gate machinery | `scripts/check.sh`, `scripts/check-state-dir.sh`, `scripts/vale-comments.sh`, `scripts/checks/**`, `Makefile`, `.golangci.yml`, `.gitleaks.toml`, `.gitleaksignore`, `.gitignore`, `.vale.ini`, `.vale/**`, `ruff.toml`, `go.mod`, `go.sum`, `web/package.json`, `web/package-lock.json`, `web/scripts/gates/**`, `web/scripts/rendered/**`, `tools/photos/**` | full | full |
| Web | `web/**` | rendered leg runs | `npm run check --prefix web` runs |
| Go package | the nearest enclosing directory that `go list -f '{{.Dir}}' ./...` names, for any file (testdata, embeds, and `.py` included). A path whose own directory is gone maps to its nearest existing ancestor package; none at all is unmatched | rendered leg runs when the package is in the rendered set | a `*.go` file only; anything else is full |
| Corpus | `corpus/**` | conform is decided by Task 3 | full |
| Photos | `photos/**` | the integrity leg already runs | the integrity leg already runs |
| Ops | `systemd/**`, `cloud-init/**`, `env/**`, `scripts/deploy.sh`, `scripts/provision.sh`, `scripts/firewall-rules.json` | their legs already run whole | full |
| Python tooling | `tools/unidecode-table/**` | ruff already runs whole | full |
| No effect | `docs/**`, `.claude/**`, `_salvage/**`, `tools/venv/**`, `README.md`, `ROADMAP.md`, `IDEAS.md`, `BACKLOG.md`, `CLAUDE.md`, `LICENSE` | gitleaks only | gitleaks only |
| Unmatched | anything else | full | full |

`tools/venv/requirements.txt` is not a gate input. `check.sh` never reads `tools/venv/`, and `make derive` sits
outside `check` (`Makefile:136-146`).

## Task 1: The caller's base and fail-closed classification

**Files:** `scripts/check.sh`. **Gate:** the per-task gate.

- `--base <sha>` is accepted with `--scope-changed` and `--reduced`. It is refused (exit 2) with the full and merge
  forms.
- The base is used only when all three hold: it is a full SHA in the repo's object format (sha1, 40 hex),
  `git rev-parse --verify "<sha>^{commit}"` returns that same string, and `git merge-base --is-ancestor <sha> HEAD`
  holds. Anything else (an abbreviation, a ref name, a missing object, a non-ancestor, or shell metacharacters) falls
  back to full.
- Without `--base`, the base derivation is unchanged. The change set is still git's (`git diff --name-only <base>`
  plus untracked files) in every case.
- An empty change set falls back to full.
- The class table's matching and its two promotions to full (gate machinery and unmatched) land here.
- `--dry-run` works with any form. It derives the scope, prints the scope line, runs no leg, and exits 0. The proofs
  and a reviewer's reproduction use it.
- The header's derivation rule is reconciled in the comment beside the code. A base commit is a fact, unlike a file
  list: the runner records it with `git rev-parse HEAD` before the implementer's first dispatch
  (`pass-execute.js:778-784`), and git still derives every file. The validation catches a malformed or unrelated
  commit. A later commit passed by mistake would narrow the scope, but every boundary, merge, and close runs the full
  form.
- **Proofs** (scratch worktree, `--dry-run`, red under the named mutation):
  - P1, invalid base: `--base` given an abbreviated SHA, `HEAD~1`, a missing 40-hex SHA, and `abc;true`, each falls
    back to full (mutation: skip the shape and resolve checks).
  - P2, non-ancestor base: a commit on a side branch falls back to full (mutation: drop the ancestry check).
  - P3, empty change set: `--base` set to HEAD on a clean tree falls back to full (mutation: drop the empty rule).
  - P4, `go.mod`: a one-line `go.mod` change runs full (mutation: drop `go.mod` from the machinery list and list it
    as no-effect; the unmatched promotion alone also sends it to full).
  - P5, `check.sh`: a comment line in `scripts/check.sh`, committed after the base, runs full (mutation: drop
    `check.sh` from the list and list it as no-effect, for the same reason).
  - P6, unmatched path: a new top-level `probe.cfg` runs full (mutation: classify unmatched paths as no-effect).

## Task 2: The rendered leg by reverse dependency, and `svelte-check` in the reduced form

**Files:** `scripts/check.sh`. **Gate:** the per-task gate.

- The hand path list (`check.sh:240-256`) is deleted.
- In `--scope-changed`, the rendered leg runs when any changed path is in the Web class, or is in the Go package class
  for a package whose directory appears in `go list -deps -test -tags rendered -f '{{.Dir}}' ./cmd/dubplate-server`.
  The leg's own machinery (`Makefile`, `web/scripts/rendered/**`) is already full.
- The scope line names the paths that selected the leg.
- In `--reduced`, `npm run check --prefix web` runs whenever the change set holds a `web/` path. Today the form only
  builds.
- **Proofs:**
  - P7, `internal/ops` only (in the rendered set, absent from the hand list): the rendered leg runs (mutation: restore
    the hand list).
  - P8, `opshttp` only: first commit a blank import of `internal/ops/opshttp` into `cmd/dubplate-server` below the
    base, standing in for rung 14a's mount. Then an `opshttp`-only change runs the rendered leg (mutation: the hand
    list). Without the planted import, the same change skips the leg, which is correct today; quote the `go list`
    line.
  - P9, `web/src/` only: the leg runs. Under `--reduced`, a real run's transcript shows `npm run check` (mutation:
    drop that rule).
  - P10, `docs/` only: the leg is skipped. This is the economy line, with no mutation.

## Task 3: Go's test cache sizes the conform suite

**Files:** `scripts/check.sh`, `Makefile`. **Gate:** the per-task gate.

- The probe runs first, in a scratch worktree, with `go test ./internal/harness/conform` run twice to reach
  `(cached)`. It then records three cases:
  - Touching one corpus case file reruns the suite (not `(cached)`).
  - A one-line change under `internal/engine/` reruns it.
  - A `docs/` change keeps it `(cached)`.
- Quote `go help test`: "Tests that open files within the package's module or that consult environment variables
  only match future runs in which the files and environment variables are unchanged." `hashOpen` in
  `cmd/go/internal/test/test.go` hashes a file's size and mtime, and a directory's listing with each entry's stat.
- **If the probe confirms invalidation** (planning expects it), the conform hand scoping (`:219-235`, `:348-356`) is
  deleted. `--scope-changed` then runs `make web`, `make vet lint test tidy-check`, and `make web-rendered` only when
  Task 2 selects it. It never calls `make check`, so the double bundle build goes away.
- **If the probe does not confirm it,** conform gets Task 2's rule with `go list -deps -test ./internal/harness/conform`
  plus `corpus/`, and the record says why.
- **Subprocess tests run uncached.** Two tests read source through a `go list` subprocess:
  `TestServerNeverLinksTheHarnessOrChecker` (`internal/harness/conform/arch_test.go:240`) and
  `TestLayerImportsNothingInternal` (`internal/opensubsonic/boundary_test.go:58`). The cache cannot see those reads,
  so either test can replay a stale pass in every form today. A planning probe (a scratch module) proved the class: a
  file read by `os.ReadFile` invalidated the cache and failed, but the same change read through `exec` replayed
  `ok (cached)` until `-count=1`.
- So the Makefile's `test` target also runs those two tests with `-count=1`, the idiom `go help test` names. The
  target fails if that run executes neither named test, so a rename cannot make it vacuous.
- **Proofs:**
  - P11: the three cache cases above.
  - P12: plant an `internal/harness` import in `cmd/dubplate-server`. `go test ./internal/harness/conform` replays
    `ok (cached)` (the red, which is today's behavior), and `make test` fails, naming the import. The red
    reproduces because conform's own file walks open only `internal/engine` (`arch_test.go:144`, `:155`).
  - P13: renaming either named test makes `make test` fail.
  - P14, end to end: in the scratch worktree with `npm ci --prefix web`, run `cairn-run-gate 'bash scripts/check.sh
    --scope-changed --base <sha>'` on the heavy lane twice. With an `internal/ops` change, the rendered leg runs, the
    transcript shows one `make web` and no `make check`, and the gate is green and timed. With a `docs/` change, the
    leg is skipped, and the gate is green and timed.

## Task 4: The `pass-execute` shim

**Files:** `scripts/checks/gate-tier.mjs` (new; the runner fixes this path). **Gate:** the per-task gate.

- **Interface.** The runner calls `node scripts/checks/gate-tier.mjs --range <base>..HEAD [--paint yes|no] [--pin
  <tier>]` (`pass-execute.js:449`, `:823`). On every decision, the shim exits 0 and prints exactly one stdout line:
  `bash <root>/scripts/check.sh --scope-changed --base <base>`. `<root>` is the shim's own repo root, so a worktree's
  shim names that worktree's `check.sh`.
  *Note: the runner line numbers cited in this paragraph are stale. The current sites, recorded under Record, Task 4,
  are `:513`, `:922`, `:887`, `:904-905`, and `:937-939`.*
- **Fallback.** The shim prints the full form, `bash <root>/scripts/check.sh`, in four cases: the base is not
  40-hex, the right side is not `HEAD`, the base is empty (the runner's failed base probe yields `--range ..HEAD`), or
  the pin is `full`. Shape-checking the base before embedding it is the shim's one check, and it exists for shell
  safety. `check.sh` does every other validation.
- **Refusal.** Only a missing or malformed `--range` exits 2 with nothing on stdout. The runner then falls back to the
  plan's gate (`pass-execute.js:835-837`).
- **Flags.** `--paint` is accepted and ignored, with one stderr line: dubplate has no paint tier, and any `web/` change
  already runs the rendered leg. `--pin` with any value other than `full` prints the full form and notes it on stderr,
  because no narrower pin exists. Today the runner never sends `--pin`: a pinned task skips the classifier on both
  sides (`:461-467`, `:816-818`).
- Stderr carries the decision; stdout carries only the gate string. The comments follow `ts-conventions` (JSDoc).
- **Proofs:**
  - P15: each case above, showing stdout, stderr, and the exit code. `--paint yes` and no flag give byte-identical
    stdout.
  - P16: the runner-shaped call `node scripts/checks/gate-tier.mjs --range <base>..HEAD` in P14's scratch tree prints
    the string P14 ran.

## Close

1. **Simplify.** Run `code-simplifier:code-simplifier` once, over `scripts/checks/gate-tier.mjs` (the one JS file),
   with `.claude/instructions/simplifier-brief.md` pasted verbatim. Bash and the Makefile take no simplifier.
2. **Docs.**
   - Rewrite `check.sh`'s header (the form list, the derivation rule and its `--base` reconciliation, `--dry-run`, and
     the class table in brief).
   - Update CLAUDE.md's Gate section: the scoped gate paragraph, the conform and rendered selection, and the two
     uncached tests.
   - Add one line to CLAUDE.md's Shell stack rule: the shim is the one JS file under `scripts/`, at the path the
     runner fixes.
3. **Rung 14a plan amendment.** Do this only after the 14a fold's commit lands and the file shows clean in
   `git status`. If the fold has not committed, carry the amendment in STATUS.
   - Per-task gate notes (`:67-78`): the runner's classifier sizes each task from its recorded base. The plan's
     `--scope-changed` string stays as the runner's fallback, and the heavy lane stays.
   - Lane launch, step 8 (`:333-336`): both BACKLOG items it files are resolved by this pass, so the step goes.
   - The close step (`:797-798`): the score's clock comes from `cairn-run-gate --records <toplevel> <branch>`.
4. **Records feature.** Check that `--records` exists by grepping the script (`grep -q -- '--records' "$(command -v
   cairn-run-gate)"`). Never invoke an unknown flag: `cairn-run-gate` has no flag parser beyond `--receipt` and runs
   any other first argument as a gate string. If the feature is present, read the clock from it; if absent, record
   "unavailable" and sum the gate lines by hand.
5. **Implementer memory.** Re-read `.claude/agent-memory/dubplate-implementer/feedback_gate-vanished-status.md`
   against `cairn-run-gate` today. A vanish now prints `gate vanished; re-issue starts a fresh run` and exits 75 with
   no `gate exit:` line. The third vanish in a row prints `gate exit: 1 (vanished 3 times)`, which is a red to report.
   Correct the file's quoted text and its "ignore the exit code" advice. The file predates this pass and stays
   uncommitted like its siblings.
6. **Full gate.** Run `cairn-run-gate 'bash scripts/check.sh'` on the heavy lane. A receipt match on the final tree
   stands in for the run. On a red, revert the simplifier commit first.
7. **Architecture read.** Run one `go-architecture-reader` only if a Go package changed. The plan expects none,
   because Task 3's fix lives in the Makefile.
8. **Ledgers.**
   - STATUS: the lane 1 row back to idle, and the next action unchanged (rung 14a).
   - HISTORY: what landed, every red quoted in Record, what a later pass would be wrong to rediscover (the subprocess
     cache hole; `opshttp` joins the rendered set at 14a's mount with no `check.sh` edit), and the pass score.
   - The pass score: tokens against 2.5M, attended time, and clock against 2 hours (total, gate time, lock wait,
     rework).
   - File one BACKLOG line through `log-issue`: a scoped run that skips the rendered leg launches no browser and
     could take the light lane, but the lane is chosen before `check.sh` classifies.
9. **Commit** specific files, then pre-bake the handoff (pass-core).

## Rulings for Geoff

None.

## Record (filled during execution)

### Task 1

Method: scratch worktree per proof (detached at HEAD 2af4404, removed after), the clean or mutated `check.sh` committed
below the base, the base `B` taken from that commit, then the planted change after it. Every run is
`bash scripts/check.sh --scope-changed --dry-run --base $B`; the shellcheck run on `check.sh` is clean.

- **P1, invalid base.** Planted: `docs/probe.md` edit. Scope lines under the code, one per input:
  `fallback="--base 'da067dad2857' is not a full 40-hex commit id"`, `fallback="--base 'HEAD~1' is not a full 40-hex
  commit id"`, `fallback="--base deadbeef... does not resolve to that commit"`, `fallback="--base 'abc;true' is not a
  full 40-hex commit id"`. Red (shape and resolve checks skipped): abbreviated SHA gives `scope=changed
  base=ba7a61c7aaf6 source=caller conform=skipped rendered=skipped gitleaks=changed files=1: docs/probe.md`. The other
  three still fall back under that mutation (`HEAD~1` through `scripts/check.sh is machinery`, because the mutated
  file sits in `HEAD~1..tree`; the missing SHA and `abc;true` through the ancestry check), so only the abbreviation
  shows the red.
- **P2, non-ancestor.** Planted: a side commit (`docs/side.md`) on the below-base commit, then `docs/probe.md`. Scope
  line: `fallback="--base 4d89... is not an ancestor of HEAD"`. Red (ancestry check dropped): `scope=changed
  base=9fab3d61... source=caller ... files=2: docs/probe.md docs/side.md`.
- **P3, empty change set.** Planted: nothing, `--base` the below-base HEAD. Scope line: `fallback="the change set is
  empty" base=c87fa5c1... source=caller`. Red (empty rule dropped): `scope=changed base=69335ba7... conform=skipped
  rendered=skipped gitleaks=changed files=0:`.
- **P4, `go.mod`.** Planted: one `// probe` line in `go.mod`, committed after the base. Scope line: `fallback="go.mod is
  machinery"`. Red under the named mutation (drop `go.mod` from the machinery list) did NOT fire: `fallback="go.mod is
  unmatched"`, because the root directory is no Go package and the unmatched promotion catches it. Red under the
  stronger mutation (drop it and list it as no-effect): `scope=changed base=0f563547... files=1: go.mod`.
- **P5, `check.sh`.** Planted: a `# probe` line in `scripts/check.sh`, committed after the base. Scope line:
  `fallback="scripts/check.sh is machinery"`. The named mutation did NOT fire: `fallback="scripts/check.sh is
  unmatched"`. Red under the stronger mutation (also listed as no-effect): `scope=changed base=9be1f576... files=1:
  scripts/check.sh`.
- **P6, unmatched path.** Planted: untracked top-level `probe.cfg`. Scope line: `fallback="probe.cfg is unmatched"`.
  Red (unmatched classified as no-effect): `scope=changed base=4da6101e... files=1: probe.cfg`.

Finding: the machinery list is defense in depth with the unmatched promotion for `go.mod` and `scripts/check.sh`
(neither lies in a Go package, so the named P4 and P5 mutations cannot fire alone). The plan text for P4 and P5 has since been amended to the strong form (drop from the list and list as
no-effect), so the finding is resolved: the strong mutations are now the named ones.

Fix round (review of bfd33c5). The change set now comes from `git diff --name-only --no-renames -z <base>` and `git ls-files
--others --exclude-standard -z`, each checked on its own, and `go list ./...` is captured to a file with its exit status
checked. The P4 and P5 mutations in the plan text were amended to "drop from the machinery list and list as no-effect"
(the strong form run above is now the named one).

- **P1b, `go list` failure.** Planted: untracked `internal/engine/store/zz/zz.go` importing a missing module. Scope line:
  `fallback="go list could not name the module's packages"`. Red (old process-substitution form): `scope=changed
  base=0b87e558... conform=full rendered=skipped ... files=1: internal/engine/store/zz/zz.go`.
- **P1c, `git diff` failure.** Triggered with a `git` shim on `PATH` that exits 1 for `diff` and execs git otherwise
  (a corrupt index also fails `git status`, which would mask the red). Planted: `ROADMAP.md` edit committed after the
  base, plus untracked `docs/u.md`. Scope line: `fallback="git diff could not list the tracked changes"`. Red (old
  piped form): `scope=changed base=10181e9c... files=1: docs/u.md`, the tracked change silently dropped.
- **P1d, rename.** Planted: `git mv .gitleaksignore docs/gitleaksignore.bak`. Scope line: `fallback=".gitleaksignore is
  machinery"`. Red (`--no-renames` dropped): `scope=changed base=8557d52e... files=1: docs/gitleaksignore.bak`.

### Task 2

Same method. `rendered=` is `run` or `skipped`; `rendered_by=` names the selecting paths. The hand-list mutation
replaces the selection test with the deleted list's pattern.

- **P7, `internal/ops` only.** Planted: a `// probe` line in `internal/ops/doc.go`. Scope line: `rendered=run
  rendered_by=internal/ops/doc.go`. Red (hand list restored): `rendered=skipped` for the same change.
- **P8, `opshttp` only.** Planted: a `// probe` line in `internal/ops/opshttp/doc.go`, with
  `cmd/dubplate-server/opshttp_mount.go` (a blank import of `internal/ops/opshttp`) committed below the base. Scope line:
  `rendered=run rendered_by=internal/ops/opshttp/doc.go`. Red (hand list, same import): `rendered=skipped`. Without the
  import the same change gives `rendered=skipped rendered_by=none`, which is correct today. The `go list` lines at
  HEAD: `262:.../internal/ops` and `263:.../internal/ops/local`, with no `opshttp`.
- **P9, `web/src/` only.** Planted: a `// probe` line in `web/src/hooks.client.ts`. Scoped line: `rendered=run
  rendered_by=web/src/hooks.client.ts`; red (Web selector dropped): `rendered=skipped rendered_by=none`. Reduced line:
  `scope=reduced ... rendered=skipped web_check=1`; red (rule dropped): `web_check=0`. The real run
  (`CAIRN_GATE_LANE=light cairn-run-gate 'cd <scratch> && bash scripts/check.sh --reduced --base <sha>'`, no browser
  in the reduced form) printed `gate exit: 0` and the transcript holds `85:== npm run check --prefix web
  (svelte-check) ==` and `91:... COMPLETED 315 FILES 0 ERRORS 0 WARNINGS 0 FILES_WITH_PROBLEMS`. The scratch tree
  carried a copy of `web/node_modules` (no network).
- **P10, `docs/` only (a tracked `ROADMAP.md` edit, class No effect).** Scope line: `rendered=skipped
  rendered_by=none`. No mutation.

### Task 3

**Branch chosen: the probe confirmed invalidation, so conform's hand scoping is deleted.** `go help test`: "The idiomatic
way to disable test caching explicitly is to use -count=1. Tests that open files within the package's module or that
consult environment variables only match future runs in which the files and environment variables are unchanged."
The names and lines at HEAD: `internal/harness/conform/arch_test.go:239:func TestServerNeverLinksTheHarnessOrChecker`
and `internal/opensubsonic/boundary_test.go:57:func TestLayerImportsNothingInternal`. The conform walks are rooted at
`internal/engine` (`arch_test.go:144` and `:155`, `root := filepath.Join(repoRoot(t), "internal", "engine")`).

- **P11, the three cache cases** (scratch worktree, `go test ./internal/harness/conform` run twice to reach `(cached)`):
  - `touch corpus/cases/admit-duplicate-candidate/case.json` (a file the suite reads): `ok ... 153.005s`, a rerun.
    (`touch .../claims.json`, which the suite never opens, stays `(cached)`, as it should.)
  - One line appended to `internal/engine/matcher/album_album.go`: `ok ... 183.665s`, a rerun, then `(cached)`.
  - `docs/probe-note.md` added and `ROADMAP.md` edited: `ok ... (cached)`.
  - A cold run of the suite costs about 150s, which is what the hand scoping used to avoid on unrelated changes.
- **Change.** `--scope-changed` runs `make web`, `make vet lint test tidy-check`, then `make web-rendered` only when
  Task 2 selects it. `derive_conform_affected`, `scope_conform` and the `conform=` field are gone. The Makefile's `test`
  target runs the two named tests again with `-count=1 -v`, and fails unless both report `--- PASS`.
- **P12, planted harness import.** Planted: `cmd/dubplate-server/planted_harness.go` blank-importing
  `internal/harness/cassette`. `go test ./internal/harness/conform` replays `ok ... (cached)` (the red, today's
  behavior). `make test` fails: `arch_test.go:250: cmd/dubplate-server links packages it must not:
  github.com/glw907/dubplate/internal/harness/cassette` / `--- FAIL: TestServerNeverLinksTheHarnessOrChecker`, exit 2.
  The scratch tree needed `make web` (and a `cp -a` of `web/node_modules`) first, because `go test ./...` fails on the
  embedded placeholder otherwise.
- **P13, rename.** Renaming `TestServerNeverLinksTheHarnessOrChecker` gives `make test: TestServerNeverLinksTheHarnessOrChecker
  did not run and pass`; renaming `TestLayerImportsNothingInternal` gives `make test: TestLayerImportsNothingInternal did not
  run and pass` (and `ok ... [no tests to run]` above it). Both exit 2; the unrenamed tree passes.
- **P14, end to end** (heavy lane, scratch tree, base committed below the change):
  - `internal/ops/doc.go` change: scope line `scope=changed ... rendered=run rendered_by=internal/ops/doc.go
    gitleaks=changed files=1: internal/ops/doc.go`; one `== make web ==`, no `make check` (the Go section ran
    `go vet ./...`, `golangci-lint run ./...`, `go test ./...`, then `running make web-rendered`); `gate exit: 0`, 1321
    lines. `internal/harness/conform` reran (`ok ... 149.294s`) because the scratch tree had just had files in that
    package restored by `git checkout` during P13; its dependency closure holds no `internal/ops`. The first attempt of
    this run was red on a gofmt error in the plant itself (a comment glued to `package ops`), which shows the lint leg
    still runs in the scoped form.
  - `ROADMAP.md` change: `rendered=skipped rendered_by=none files=1: ROADMAP.md`; one `make web`, no `make check`;
    `ok ... conform (cached)`; `skipping the rendered leg`; `gate exit: 0`, 441 lines.
  - Timing: the runner printed no run time, and both runs waited on the shared heavy lock (the NOTE for the second:
    "this gate waited 496s for another gate to release the machine lock", the first 204s), so wall time is unreliable.
    The log difference is 1321 lines against 441 and the skipped rendered leg; conform alone is 149s cold against
    cached. The Go section header read `== make check (Go) ==` in these transcripts; it now reads `== make vet lint test
    tidy-check (Go) ==` in the scoped form (label only, committed after the runs).

### Task 4

`scripts/checks/gate-tier.mjs` (new, uncommitted; the close gates and commits it). `node --check` clean. The runner's
call sites at HEAD of `~/.claude/workflows/pass-execute.js` (the plan's `:449`, `:823` have moved): `:513`
`const classifierCmd = `node scripts/checks/gate-tier.mjs --range ${baseSha}..HEAD${paintFlag}${classFlag}${pinFlag}`;`
and `:922` `const cmd = `node scripts/checks/gate-tier.mjs --range ${baseSha}..HEAD${paintFlag}${classFlag}`;`. The runner
also sends `--class <name>` (`:511`) and, only when `ciActive`, `--protected` (`:887`); the shim accepts both and ignores them.

- **P15** (base `95b5379...`, stdout / stderr / exit):
  - `--range B..HEAD`: `bash <root>/scripts/check.sh --scope-changed --base B` / `scoped to the range` / 0.
  - `--paint yes` and `--paint no`: same stdout, byte-identical to the no-flag run (`cmp` silent); stderr adds the `--paint ignored` line; 0.
  - `--pin full`: `bash <root>/scripts/check.sh` / `pinned` / 0. `--pin fast`: same stdout / `--pin fast has no narrower tier here; using the full form` then `pinned` / 0.
  - `--range ..HEAD`: full form / `empty base` / 0. `--range <12 hex>..HEAD`: full form / `base is not a 40-hex commit` / 0. `--range B..HEAD~1`: full form / `range does not end at HEAD` / 0.
  - `--range` with no value, no args, and `--range nodots`: empty stdout / `--range <base>..HEAD is required` / 2.
- **P16** (scratch worktree, base commit then an `internal/ops/doc.go` change): runner-shaped
  `node scripts/checks/gate-tier.mjs --range <base>..HEAD` printed
  `bash <scratch>/scripts/check.sh --scope-changed --base <base>`, naming the worktree's own `check.sh`; that string's
  `--dry-run` printed `check.sh: scope=changed base=<base> source=caller rendered=run rendered_by=internal/ops/doc.go gitleaks=changed files=1: internal/ops/doc.go`.
  Worktree removed afterward.

## Review

One lens, mechanics and proportionality, applied in a fresh reread on 2026-10-10. Each finding was verified by a
source quote or a probe and fixed inline.

1. **`opshttp` is not in the rendered set today.** `go list` at `dd161b3` omits it, so the approved proof ("an
   opshttp-only change must select the rendered leg") would fail as written. **Changed the design:** P7 uses
   `internal/ops`, which is in the set and absent from the hand list. P8 plants 14a's mount below the base and also
   proves today's correct skip.
2. **The conform suite holds a cache-unsound test, and so does `opensubsonic`.** Both run `go list` as a subprocess;
   the scratch-module probe replayed a stale `ok (cached)`. Deleting the hand scoping on the strength of Go's cache
   needs these two tests uncached. **Changed the design:** Task 3 adds the `-count=1` run, P12, and P13.
3. **`cairn-run-gate --help` is not a flag.** It ran `--help` as a gate string and queued it on the heavy lane behind
   a cairn-cms gate; it was killed and its state removed. **Changed the design:** close step 4 probes the source by
   grep. This is a friction item for the dotfiles owner.
4. **The `tool` class defaults to the light lane, but `check.sh` launches Chromium.** Fixed inline: the header pins
   the heavy lane, and the 14a amendment keeps it. The shim picks no lane, which matches the cairn-cms classifier's
   header note.
5. **Mutating `check.sh` inside a scratch tree would put `check.sh` in the change set,** which forces full and masks
   every red. Fixed inline: mutations are committed below the base.
6. **Proportionality: `--dry-run` is new surface.** Kept. It turns twelve of the 16 proofs (P1 to P8, P10, P15, P16,
   and P9's scoped half) from gate runs into seconds, and it lets a reviewer reproduce a classification. P14 proves
   the real legs follow the scope line.
7. **Proportionality: the `svelte-check` item stood alone.** Folded into Task 2, because a single-deliverable task
   costs about thirty minutes of fixed overhead (pass-gate-economy).
8. **Task 3 removes the double bundle build** noted in 14a's lane launch step 8 as a by-product of no longer calling
   `make check` in the scoped form. No extra task was needed.
9. **The `--pin` passthrough is dead in the current runner** (`:461-467`, `:816-818`). The shim handles it
   defensively; no further work.
10. **P12 could have been vacuous.** If a conform test opened `cmd/` files, the planted import would invalidate the
    cache and the red would not reproduce. Verified: the walks are rooted at `internal/engine` (`arch_test.go:144`,
    `:155`). Fixed inline: P12 says so.
