# dubplate

A group of family or friends building a well-organized library for a shared
music server. This repo is the whole estate: the `dubplate` Go binary,
the VPS provisioning (cloud-init, compose, systemd), and the web app. The
pings Worker lives in `~/Projects/pings`. Ease of setup by a stranger is a
primary product goal: no site-specific value, path, or identity default
outside `internal/config`.

**The active initiative (2026-09-02) is the greenfield dubplate-server
build**: a fresh, universal, foundation-first build on a 17-rung ladder
per `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified), governed by the rulings in
`docs/superpowers/research/2026-09-02-greenfield-server-rulings.md`. The
pre-pivot `cmd/` + `internal/` + `web/` + `e2e/` trees are frozen under
`_salvage/` (git tag `pre-greenfield`) as SALVAGE TIER: read-only
adoption material (admitted on merit, exemplary-Go/idiomatic-Svelte
bars), never imported, and retired at launch. The group's system is not
live; the box runs only Navidrome + guardian timers until the new server
launches complete (Go matcher included). Two standing bars, from memory:
exemplary Go (`idiomatic-beautiful-go`) and render-and-read every
user-path page before any attended ask (`see-it-before-geoff-tests-it`).

**Versions (Geoff, 2026-09-25).** The ladder's rung 16 cutover, with
Navidrome still streaming, is a 0.x pre-release. **0.8** is the first
production release: the built-in streamer replaces Navidrome. **0.9** adds
the paired local client. **1.0** adds polish and the remaining features,
the format ladder among them. Documents written before this ruling say
"1.0" for the launch, "1.1" for the formats, and "2.0" for the streamer,
and they stay as written.

**Build for 0.8 (Geoff, 2026-09-07, re-keyed 2026-09-25).** The next
initiative after launch is dubplate-server 0.8, the built-in OpenSubsonic
streamer that subsumes Navidrome (ROADMAP `streamer-0-8`; the spec under
`docs/superpowers/specs/` still says 2.0). Every rung plan authored from
now on weighs its surface against that spec's forward-hooks list and
states per task whether it lands a hook or is neutral; a rung 6 schema or
a rung 12 driver seam built without 0.8 in view is rework. Route newly
discovered hooks into that list; never build 0.8 features early. The same spec carries a prior-art list (the
OpenSubsonic specification, Navidrome and gonic for behavior only under
GPL, taglib via wazero, modernc sqlite, goose, mewkiz/flac, Picard's tag
tables, chromaprint); a rung plan names which entries it consulted and
adopts or declines each with a reason, so no rung rebuilds what a
license-compatible library already does well. An adoption is judged on
license, code quality against the house bar, maintenance (release
cadence, open-issue age, bus factor), and adoption (importers, who
depends on it), with the evidence recorded in the plan, never on
familiarity alone. **Refined (Geoff, 2026-09-12):** a 0.8 forward hook is
named in the rung plan and built by the rung that first consumes it,
and nothing is exported ahead of a caller.

## Ledgers

- `docs/STATUS.md` — read in full at session start; present tense only.
- `docs/HISTORY.md` — per-pass ledger, read on demand.
- `IDEAS.md` — future consideration: uncommitted features and member requests,
  read on demand. An idea moves to Promoted when an owner ruling targets it.
- `ROADMAP.md` — initiative queue. Current: the greenfield dubplate-server
  build (17-rung ladder) → packaging (specs and plans live in
  `docs/superpowers/`, migrated into the repo at T14; the dotfiles copies
  are pointers only).

## Build process (greenfield initiative; ratified 2026-09-02, re-keyed 2026-09-03)

The build runs in session-scale PHASES bounded by merge points, not
one session per rung: foundation (rung 1, closed), evidence (2, 3,
4, with rungs 5 and 8 alongside), spine (6, 14, 7, 9), surfaces (10,
11, 12, 13), launch (15, 16, 17). Inside a phase: no clears, no
re-brainstorm. Between phases: close the session, post-mortem,
STATUS handoff. A phase whose lanes outrun one session may span
more than one, each intermediate close carrying a written per-lane
handoff. Canonical record: the build spec's Process section and
`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`.

**Standing grants (Geoff, 2026-09-02, extended 2026-09-03):**
workflow mode for the whole build, execution and prep both; batch
human questions into checkpoint bundles; when Geoff is unavailable,
continue unblocked work rather than idling; up to two code lanes
plus any number of document lanes at once.

**Lanes.** A lane is one rung's execution with a named package
territory. The primary code lane and all document lanes share the
tree; a second code lane gets its own git worktree, because
`scripts/check.sh` runs the module gate over `./...` and two
whole-repo legs besides. Corpus and harness work never moves to a
worktree. Every lane launch runs the checklist: reverse edges
cleared, contested surfaces written, runaway guard armed, lane
registry entry written, venv provisioned if needed, sub-ceiling
assigned, AC power, throughput baseline. **Never launch a second
executor into the SAME tree**: the live-executor check is the lane
registry, not `pgrep -f Projects/dubplate`, which matches every lane
once a worktree exists.

**The conducting loop, per lane:** execute via the pass-execute
chain with the runaway guard armed; prep the next rung's plan in a
workflow against a tree no other executor is writing; land queued
repo edits in one batch after the executor finishes; plans are
just-in-time per lane, never front-loaded (plan staleness is the
proven dominant defect class). A lane closes through the merge
ritual: merge token, rebase, full gate in-tree, `--ff-only`, full
gate on master, and one `diff-reviewer` on the merged result
whenever the rebase was not a fast-forward. Review scales to rung
risk: mechanical rungs ride the in-chain diff-reviewer; auth (5),
streamer/SSO (12), and cutover (16) keep the full fresh-context
adversarial cycle. Subagents cannot dispatch agents: never instruct
an implementer to run `code-simplifier`; the conductor owns that
dispatch.

**Planning-miss prevention (conductor, 2026-09-23).** Every 5b-ii fix
round was an ambiguity its plan left open, so every rung plan closes
these before approval: each stdlib or library behavior a task relies on
is quoted from `go doc` or the source, never recalled; each state a
check can meet (absent, empty, malformed, many-bad) names its report;
each privileged field names every channel that may set it; and each
proof names the fixture state it needs (selected, populated) so the test
cannot pass vacuously. The plan-review lenses check these four.

**Findings are ranked, and a fold may refuse (conductor, 2026-09-23).**
Anthropic's Claude Code guidance warns that "a reviewer prompted to find
gaps will usually report some, even when the work is sound" and that
"chasing every finding leads to over-engineering"; the rung 6 spec folded
about 200 findings with none refused. Every reviewer dispatch asks for
correctness gaps ranked by consequence, not polish, and every fold
dispatch may refuse a finding whose fix costs more than the risk it
removes, with a one-line reason in the fold record.

**Expected owner touchpoints, and only these:** rung-spec rulings
(known ahead: rung 9's concurrency model), corpus and shadow
adjudication batches (small and regular; an unadjudicated divergence
blocks the launch gate), the rung 7 uploader before/after, and the
rung 16 cutover sitting.

## Gate

`bash scripts/check.sh` before every commit — `make check` (vet,
golangci-lint, test, tidy-check) over the whole module (`./...`;
`_salvage/` is invisible to the go tool), shellcheck over `scripts/*`,
ruff (`check` and `format
--check`) over every tracked `*.py` (the oracle shim and its derivation
siblings under `internal/harness/oracle/`, that package's FLAC template
generator, tagio's golden-FLAC generator under
`internal/engine/tagio/testdata/`, and `tools/unidecode-table/generate.py`),
gitleaks over
the tracked worktree (`--no-git`; the corpus under `corpus/cases/` is the repo's
first large data directory), vale (Go comment prose, `_salvage/`
excluded), `docker compose config` (skipped when Docker isn't present),
a pattern-filtered `systemd-analyze verify`, and the state-dir agreement
leg (the env template, the shadow unit, and `deploy.sh` name the same
state directory).
The `web/` leg runs today, first in `make check`'s prerequisite chain:
Node and npm required and named on failure, a stale or drifted
`npm ci --prefix web` refused, the SvelteKit static bundle built, then
`svelte-check` under strict TypeScript and the static web gates (no
server-only file, no hardcoded color, no external origin, theme
variable completeness, theme contrast, and the built bundle's stamped
theme values). The rendered (browser) gates join this leg at rung
5c-iii. Rung 6b-i's merge adds an `rclone`-dependent leg: `internal/backup`'s
tests exercise the real `rclone` binary (`internal/backup`'s share of
`go test ./...` runs about 8.3s of it). A check either
gates or does not exist; never guard one with `command -v || skip`.
Never commit red. The process-citation leg
(no rung, task, resolution, or plan-line citation in shipped Go comments)
and the exported-unused leg (an exported `internal/` symbol whose only
references are its own file and `_test.go` files) are staged: they warn
today and gate at the comment-cleanup task.

## Stack rules (mandatory skills by area)

- **Go** (`cmd/`, `internal/`): invoke the `go-conventions` skill BEFORE
  writing any Go. Module `github.com/glw907/dubplate`. cobra CLI, wrapped
  errors with sentinels via `errors.Is`, `ctx` first parameter, every
  subprocess under `exec.CommandContext`, atomic writes via
  `internal/atomicfile`, table-driven tests under the assertion-discipline
  rules (no silent-success fakes, no vacuous assertions — prove tests can
  fail).
- **Shell** (`scripts/`): ops tooling only (`deploy.sh`, `provision.sh`,
  `check.sh`, `vale-comments.sh`), shellcheck clean, no behavior tests
  of its own. The box-side import/backup shell layer is fully retired
  (Go rewrite, 2026-08-31); never add subcommand logic here.
- **Web** (`web/`): SvelteKit in static mode (Geoff, 2026-09-07;
  supersedes the standalone rule): Svelte 5 runes, DaisyUI 5,
  `adapter-static` with client-side rendering only, built to static
  assets embedded via `go:embed` and served by the Go binary, the only
  server. Server-only SvelteKit files (`+page.server.*`, `+server.*`,
  `hooks.server.*`, form actions) are forbidden and the gate checks that
  none exists. Clean, idiomatic SvelteKit is the bar, as in the cairn
  family. Invoke `svelte-conventions` and `ts-conventions` before writing
  components or TS. `tus-js-client` pinned 4.3.x, from rung 7. Node is workstation
  build-time only; nothing Node ships to the box. Rung 5's login page is
  the first page built under this rule. The look is the visual identity
  spec (`docs/superpowers/specs/2026-09-24-visual-identity-design.md`,
  Geoff, 2026-09-25): every `web/` task reads it, and it outranks any
  styling habit.
- **systemd / cloud-init**: invoke the `vps-conventions` skill BEFORE
  touching `systemd/`, `cloud-init/`, or deploy/provision scripts. Units
  are code — hardened (`NoNewPrivileges`,
  `PrivateTmp`, `ProtectSystem` with explicit `ReadWritePaths`), verified by
  the gate. A `ReadWritePaths` entry must exist on the host before the unit
  starts (deploy.sh pre-creates). cloud-init runs at first boot only;
  converging a live box is deploy.sh's job.
- **Cloudflare**: API-first with `$CLOUDFLARE_API_TOKEN` (script must
  `source ~/.local/secrets`). The tunnel is remote-managed:
  `config/cloudflared-tunnel-config.json` is the record of what gets PUT to
  the API, not a file compose reads. See `.claude/instructions/api-access.md`.
- **Python** (`internal/harness/oracle/*.py`, `tools/unidecode-table/generate.py`):
  invoke the `python-conventions` skill BEFORE writing any Python. Python
  stays confined to workstation-only derivation and generation tooling;
  nothing Python ships to the box. The oracle shim drives pinned beets
  under a uv-managed venv (`tools/venv/`) that the gate never requires and
  never runs (derivation is `make derive`, tagged `oracle`, workstation
  only). ruff (`ruff check` and `ruff format --check`) joins `scripts/check.sh`
  for every tracked `*.py` file, so the shim's bar is enforced, not
  aspirational, even though the gate can't execute it.

## Reviewers

`diff-reviewer` per plan task. At pass end, fan out by surface:
`svelte-reviewer` + `daisyui-a11y-reviewer` for `web/`, each dispatch
passing the visual identity spec's path and asking for conformance to it
(tokens, inks, the component layer, the copy register) graded alongside
correctness, since those reviewers are workstation-wide,
`web-auth-security-reviewer` for the serve/JWT surface,
`cloudflare-workers-reviewer` only in the pings repo. Web UI work
matches no streamer: dubplate replaces Navidrome at 0.8 and carries its
own visual identity (Geoff, 2026-09-24), so no reference screenshots and
no visual-fidelity chain. Every user-path page is rendered and read
before Geoff sees it, and the rung 7 uploader is Geoff's design review.
At each pass's merge the conductor dispatches one `go-architecture-reader`
per touched Go package, with no plan in context; it grades the package,
which is the unit the per-task `diff-reviewer` cannot see. Every
`code-simplifier` dispatch pastes `.claude/instructions/simplifier-brief.md`
verbatim; the four decidable patterns it names replace a clarity brief.

## Facts agents get wrong

- Failure pings are POST to `<url>/fail`; success is GET. The Worker 404s a
  GET to `/fail` — this was a live bug once.
- Lock files live in `/var/lib/musicbox/` (persistent), NOT `/run/musicbox`
  (RuntimeDirectory, deleted at unit stop). The Go layer uses NO runtime
  directory at all: the scan URL builds inline and the roster loads into
  memory (the retired shell layer's temp-file pattern died with it).
- On the interim estate (retired at rung 16), the beets database is the only
  import success record: the tag log records skips only, and beets renames
  every file. Never infer success from the filesystem. The greenfield
  server's store is its own record.
- The album is the verdict unit: one failing track rejects the batch. Never
  admit a partial album.
- Reject wording is contract (`internal/reject` / `friendly_reason`): the
  uploader page and emails share it verbatim; upload-check reuses it.
- `env/dubplate.env.template` is a secrets MANIFEST rendered by deploy.sh
  from the age store — inline values in it are not defaults and adding a
  name there requires minting it (`secret-set.sh`, registry).
- On the interim estate, collections route via a beets flexible attribute
  with query-prefixed path formats, the default collection unprefixed, and
  there is no second beets database. The greenfield server does not inherit
  this: a collection is a store row whose root the server allocates from its
  id, with no two roots nested, per
  `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`. Rung 16
  relocates the migrated collections once, conditional on rung 12's
  Navidrome check.
- Config keys are declared through the registry in each rung's own file,
  never by editing `internal/config/keys.go` (rung 1b).

## Safety rails

- Before any executor touches this repo or the box: check for live
  executors (`pgrep -f Projects/dubplate`, `git status` for warm changes
  you did not author). The box (`5.78.148.42`) is production — no writes
  without a quiet-box check; read-only ssh only during reviews.
- Never commit secrets, `.env` values, or `~/.local/secrets` content.
  No mechanism in the gate scans for them; review a diff by eye before
  staging anything that might carry one.
- Deploys go through `scripts/deploy.sh` (idempotent, `--check` first for
  anything risky). Never edit the box by hand and call it done — the repo
  is the source of truth.
