# dubplate

This repo holds the `dubplate-server` Go binary, its VPS provisioning (cloud-init and systemd), and
the web app; the pings Worker lives in `~/Projects/pings`.

## Decisions

Accepted records in `docs/decisions/` bind every change. Before editing, list the folder and read
each accepted record whose title touches the work. A proposed record is context, and a superseded
one is history. A change of mind is a new record that supersedes the old one, never an edit to it
(record 0001).

## Ledgers

- `docs/STATUS.md`: read in full at session start; present tense only.
- `docs/HISTORY.md`: a frozen archive (2026-10-10); git and PR bodies are the record.
- `BACKLOG.md`: open items, tagged `#milestone:<name>` for the milestone that takes each, with the
  milestone order under "Milestones to 0.8".
- `IDEAS.md`: future consideration: uncommitted features and member requests, read on demand. An
  idea moves to Promoted when an owner ruling targets it.
- Initiatives live in `~/Projects/initiatives`; this repo's is `dubplate-server`. Specs live in
  `docs/specs/`; frozen plans in `docs/plans/` and `docs/archive/plans/`.

## Building

Work is built from specs by `building-from-specs`, with the `dubplate-build` skill
(`.claude/skills/dubplate-build`) holding this repo's gate, fast lane, worktree setup, CI watch, risk map,
reviewers, and conventions.

## Gate

`bash scripts/gate.sh` before every commit; never commit red. The scoped form is `gate.sh --scope-changed`, and
it fails closed to the full form. A check either gates or does not exist; never guard one with
`command -v || skip`. `scripts/check.sh`'s header lists every leg; the skill's "Gate" section has the
lanes, the import-graph tests, and the workstation prerequisites outside the gate.

## Stack rules (mandatory skills by area)

The web and Python rules load from `.claude/rules/web.md` and `.claude/rules/python.md` when a session reads a
matching path.

- **Go** (`cmd/`, `internal/`): invoke the `go-conventions` skill BEFORE
  writing any Go. Module `github.com/glw907/dubplate`. cobra CLI, wrapped
  errors with sentinels via `errors.Is`, `ctx` first parameter, every
  subprocess under `exec.CommandContext`, atomic writes via
  `internal/atomicfile`, table-driven tests under the assertion-discipline
  rules (no silent-success fakes, no vacuous assertions; prove tests can
  fail). A test's expected value is a literal or an independently computed
  value, never built with the helper the code under test calls: such a test
  cannot fail when the code regresses, because both sides drift together.
- **Shell** (`scripts/`): ops tooling only (`deploy.sh`, `provision.sh`,
  `check.sh`, `gate.sh`, `check-state-dir.sh`, `vale-comments.sh`), shellcheck clean,
  no behavior tests of its own. The box-side import/backup shell layer is fully retired
  (Go rewrite, 2026-08-31); never add subcommand logic here.
- **systemd / cloud-init**: invoke the `vps-conventions` skill BEFORE
  touching `systemd/`, `cloud-init/`, or deploy/provision scripts. Units
  are code, hardened (`NoNewPrivileges`,
  `PrivateTmp`, `ProtectSystem` with explicit `ReadWritePaths`), verified by
  the gate. A `ReadWritePaths` entry must exist on the host before the unit
  starts (deploy.sh pre-creates). cloud-init runs at first boot only;
  converging a live box is deploy.sh's job.
- **Cloudflare**: API-first with `$CLOUDFLARE_API_TOKEN` (script must
  `source ~/.local/secrets`). The tunnel and its DNS are deleted (Geoff,
  2026-10-05). See `.claude/instructions/api-access.md`.

## Facts

- Failure pings are POST to `<url>/fail`; success is GET. The Worker 404s a
  GET to `/fail`; this was a live bug once.
- `env/dubplate-server.env.template` is the render manifest deploy.sh
  turns into `/etc/dubplate/server.env`: a bare `NAME=` row resolves from
  the age store, so adding one requires minting it (`secret-set.sh`,
  registry); a `NAME=value` row is a literal.

## Safety rails

- Before a session enters a worktree, it runs `claude-worktree-busy <path>`.
- The box (`5.78.148.42`) is production: no writes without a quiet-box check; read-only ssh only during reviews.
- Never commit secrets, `.env` values, or `~/.local/secrets` content.
  The gate's gitleaks leg scans for them, but it misses a value that
  looks ordinary; review a diff by eye before staging anything that might
  carry one.
- Deploys go through `scripts/deploy.sh` (idempotent; it checks by
  default, and `--apply` writes, only from a clean committed tree). Never edit the box by hand and call it done: the repo
  is the source of truth.
