# dubplate

A group of family or friends building a well-organized library for a shared
music server. This repo is the whole estate: the `dubplate` Go binary,
the VPS provisioning (cloud-init, compose, systemd), and the web app. The
pings Worker lives in `~/Projects/pings`. Ease of setup by a stranger is a
primary product goal: no site-specific value, path, or identity default
outside `internal/config`.

**The active initiative (2026-09-02) is the greenfield dubplate-server
build**: a fresh, universal, foundation-first build per
`docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md`, governed
by the rulings in
`docs/superpowers/research/2026-09-02-greenfield-server-rulings.md`. The
existing `internal/` + `web/` code is SALVAGE TIER: it stays in-tree as
adoption material (admitted on merit, exemplary-Go/idiomatic-Svelte bars)
and retires at launch. The family system is not live; the box runs only
Navidrome + guardian timers until the new server launches complete (Go
matcher included). Two standing bars, from memory: exemplary Go
(`idiomatic-beautiful-go`) and render-and-read every user-path page before
any attended ask (`see-it-before-geoff-tests-it`).

## Ledgers

- `docs/STATUS.md` — read in full at session start; present tense only.
- `docs/HISTORY.md` — per-pass ledger, read on demand.
- `ROADMAP.md` — initiative queue. Current: Go rewrite → uploader →
  packaging (specs and plans live in `docs/superpowers/`, migrated into the
  repo at T14; the dotfiles copies are pointers only).

## Gate

`bash scripts/check.sh` before every commit — `make check` (build, vet,
golangci-lint, test, tidy-check), shellcheck, vale (Go comment prose),
compose config, systemd-analyze verify. The `web/` build joins when the
uploader lands. A check either gates or does not exist; never guard one
with `command -v || skip`. Never commit red.

## Stack rules (mandatory skills by area)

- **Go** (`cmd/`, `internal/`): invoke the `go-conventions` skill BEFORE
  writing any Go. Module `github.com/glw907/dubplate`. cobra CLI, wrapped
  errors with sentinels via `errors.Is`, `ctx` first parameter, every
  subprocess under `exec.CommandContext`, atomic writes via
  `internal/atomicfile`, table-driven tests under the assertion-discipline
  rules (no silent-success fakes, no vacuous assertions — prove tests can
  fail).
- **Shell** (`scripts/`, migration window only): behavior is pinned by
  `tests/*.bats`; `lib.sh` is the shared layer; shellcheck clean. The shell
  layer is being replaced per-subcommand — never add capability to it, only
  port-parity fixes.
- **Web** (`web/`): Svelte 5 runes, standalone (NO SvelteKit), DaisyUI 5,
  Vite to static assets embedded via `go:embed`. Invoke `svelte-conventions`
  and `ts-conventions` before writing components or TS. `tus-js-client`
  pinned 4.3.x. Node is workstation build-time only; nothing Node ships to
  the box.
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

## Reviewers

`diff-reviewer` per plan task. At pass end, fan out by surface:
`svelte-reviewer` + `daisyui-a11y-reviewer` for `web/`,
`web-auth-security-reviewer` for the serve/JWT surface,
`cloudflare-workers-reviewer` only in the pings repo. Uploader UI work
additionally gates on the visual-fidelity chain (Navidrome reference
screenshots, fresh-context `visual-verifier`, Geoff's before/after).

## Facts agents get wrong

- Failure pings are POST to `<url>/fail`; success is GET. The Worker 404s a
  GET to `/fail` — this was a live bug once.
- Lock files live in `/var/lib/musicbox/` (persistent), NOT `/run/musicbox`
  (RuntimeDirectory, deleted at unit stop). The Go layer uses NO runtime
  directory at all: the scan URL builds inline and the roster loads into
  memory (the retired shell layer's temp-file pattern died with it).
- The beets database is the only import success record: the tag log records
  skips only, and beets renames every file. Never infer success from the
  filesystem.
- The album is the verdict unit: one failing track rejects the batch. Never
  admit a partial album.
- Reject wording is contract (`internal/reject` / `friendly_reason`): the
  uploader page and emails share it verbatim; upload-check reuses it.
- `env/dubplate.env.template` is a secrets MANIFEST rendered by deploy.sh
  from the age store — inline values in it are not defaults and adding a
  name there requires minting it (`secret-set.sh`, registry).
- Collections route via a beets flexible attribute with query-prefixed path
  formats; the default collection keeps the unprefixed layout. No second
  beets database, ever.

## Safety rails

- Before any executor touches this repo or the box: check for live
  executors (`pgrep -f Projects/dubplate`, `git status` for warm changes
  you did not author). The box (`5.78.148.42`) is production — no writes
  without a quiet-box check; read-only ssh only during reviews.
- Never commit secrets, `.env` values, or `~/.local/secrets` content. The
  gate runs gitleaks.
- Deploys go through `scripts/deploy.sh` (idempotent, `--check` first for
  anything risky). Never edit the box by hand and call it done — the repo
  is the source of truth.
