# dubplate

A group of family or friends building a well-organized library for a shared
music server. This repo is the whole estate: the `dubplate` Go binary,
the VPS provisioning (cloud-init, systemd), and the web app. The
pings Worker lives in `~/Projects/pings`. Ease of setup by a stranger is a
primary product goal: no site-specific value, path, or identity default
outside `internal/config`.

**The active initiative (2026-09-02) is the greenfield dubplate-server
build**: a fresh, universal, foundation-first build on a 17-rung ladder
(each rung has a name, and the run order is BACKLOG's "Path to 0.8")
per `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified), governed by the rulings in
`docs/superpowers/research/2026-09-02-greenfield-server-rulings.md`. The
pre-pivot `cmd/` + `internal/` + `web/` + `e2e/` trees are frozen under
`_salvage/` (git tag `pre-greenfield`) as SALVAGE TIER: read-only
adoption material (admitted on merit, exemplary-Go/idiomatic-Svelte
bars), never imported, and retired at launch. The group's system is not
live; the box runs only dubplate-server (Navidrome, the Cloudflare
tunnel, the box's beets, and the incumbent musicbox stack are retired,
Geoff 2026-10-05) until the new server launches complete (Go matcher
included).

**Two standing bars (Geoff, 2026-09-02).**
- *Exemplary Go.* Idiomatic standard-library style and architectural beauty, never mere correctness: the code
  is part of the product, and a correct but awkward implementation misses the bar the way an unclean UI does.
  A reviewer may return `fix` on awkward-but-working code, and a dispatch names "idiomatic, architecturally
  beautiful Go" as an acceptance dimension.
- *Render and read every user-path page before any attended ask.* Never ask Geoff to test a surface you have
  not seen, the auth pages in front of the app included, starting from the URL the person will type. A
  status code is not evidence; a screenshot read in the main loop is. (A deploy once went to Geoff after
  only a 302 was checked, and the login page behind it carried another organization's branding.)

**Words (Geoff, 2026-09-08 and 2026-09-24).** The people who share a library are "the group" and one person
is "a member"; "family" survives only in the product-essence line. Specs and UI copy use plain terms, with no
coined metaphor ("the crate") and no record-shop pun ("Back to the shelf"): the identity lives in the visuals.

**Versions (Geoff, 2026-09-25, re-planned 2026-10-05).** The first launch
is **0.8.0**, on the built-in streamer: no earlier pre-release and no
box swap. **0.9** adds the paired local client. **1.0** adds polish and the
remaining features, the format ladder among them. A version names what a
member can do with it. A document older than a ruling keeps its numbers
and stays as written.

**Build for 0.8.** The `streamer` rung is the built-in OpenSubsonic streamer (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`).
Rung plans weigh their surface against that spec's forward-hooks list and its prior-art list; a hook is
named in the plan and built by the rung that first consumes it, and nothing is exported ahead of a caller.
Never build 0.8 features early. Full text: `docs/standing-rules.md`.

**Current dependencies (Geoff, 2026-10-04).** Development runs on the newest production version of every
library, tool, and pinned binary; pins stay exact. Minors and patches land without asking; a major waits
for Geoff. libxml2 (`/home/linuxbrew/.linuxbrew/bin/xmllint`, 2.15.4) is held by `brew pin libxml2`: a
bump is `brew unpin libxml2 && brew upgrade libxml2`, regenerate the XML goldens, `brew pin libxml2`,
then upgrade the eight held dependents. Full text and the last sweep: `docs/standing-rules.md`,
`docs/superpowers/research/2026-10-04-dependency-sweep.md`.

## Ledgers

- `docs/STATUS.md`: read in full at session start; present tense only.
- `docs/HISTORY.md`: a frozen archive (2026-10-10); git and PR bodies are the record.
- `BACKLOG.md`: open items, tagged `#rung:<name>` for the rung that takes each, and the ladder's run order
  under "Path to 0.8".
- `IDEAS.md`: future consideration: uncommitted features and member requests,
  read on demand. An idea moves to Promoted when an owner ruling targets it.
- Initiatives live in `~/Projects/initiatives`; this repo's is `dubplate-server`, the greenfield build
  (17-rung ladder) followed by packaging. Specs and plans live in `docs/superpowers/`.

## Build process

The build's milestones, in order, are foundation, evidence, spine, surfaces, and launch; BACKLOG's "Path to 0.8"
holds the run order. Process: the `dubplate-pass` skill (`.claude/skills/dubplate-pass`), loading `pass-core`.
Product touchpoints, and only these: corpus adjudication batches, the `intake` uploader sitting, and the
first-launch sitting. The design approval and the packet read are `pass-core`'s.

## Gate

`bash scripts/gate.sh` (the wrapper that runs `check.sh` through `cairn-run-gate`) before every commit: `make check` (vet, golangci-lint, test, tidy-check) over the
whole module (`_salvage/` is invisible to the go tool), shellcheck, ruff, the photo pool check, gitleaks,
vale over Go comments, a hard-required `systemd-analyze verify`, the state-dir agreement leg
(`scripts/check-state-dir.sh`, which also admits the serve unit's one music-volume grant, the common parent of the staging and library paths), and the `web/` leg (static bundle, `svelte-check`, the static web gates,
and the rendered Playwright gates on the heavy lane). Each leg, the two OpenSubsonic import-graph tests,
and the workstation prerequisites outside the gate (`npm ci --prefix web`, `make unidecode-wheel`,
`make subsonic-xsd`, `xmllint` at libxml 2.15.4) are in `docs/standing-rules.md`. The scoped form is
`gate.sh --scope-changed`; it fails closed to the full form. A check either gates or does not exist;
never guard one with `command -v || skip`. Never commit red.

## Stack rules (mandatory skills by area)

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
- **Web** (`web/`): SvelteKit in static mode (Geoff, 2026-09-07;
  supersedes the standalone rule): Svelte 5 runes, DaisyUI 5,
  `adapter-static` with client-side rendering only, built to static
  assets embedded via `go:embed` and served by the Go binary, the only
  server. Server-only SvelteKit files (`+page.server.*`, `+server.*`,
  `hooks.server.*`, form actions) are forbidden and the gate checks that
  none exists. Clean, idiomatic SvelteKit is the bar, as in the cairn
  family. Invoke `svelte-conventions` and `ts-conventions` before writing components or TS. `tus-js-client` pinned 4.3.x, from `intake`. Node is workstation
  build-time only; nothing Node ships to the box. The `auth` rung's login page is
  the first page built under this rule. Push hard toward the stack's own idiom
  (Geoff, 2026-09-01): reach for the DaisyUI component class and its documented
  modifiers before a hand-rolled utility stack, never fight a component's layout
  model, take colors only from theme tokens, prefer `$derived` to an effect that
  mirrors state, and keep `$effect` for real side effects with cleanup. The idiom
  wins over a finding's literal fix text, un-idiomatic code adjacent to a change is
  in scope for it, and the exemplar is `~/Projects/cairn-cms/src/lib/admin-toolkit/`:
  imitating it beats interpreting the rule. The look is the visual identity
  spec (`docs/superpowers/specs/2026-09-24-visual-identity-design.md`,
  Geoff, 2026-09-25): every `web/` task reads it, and it outranks any
  styling habit.
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
- **Python** (`internal/harness/oracle/*.py`, `internal/engine/matcher/testdata/gen_featuring_golden.py`,
  `tools/unidecode-table/generate.py`):
  invoke the `python-conventions` skill BEFORE writing any Python. Python
  stays confined to workstation-only derivation and generation tooling;
  `generate.py` only re-derives the table pin (its output is never
  committed; the binary builds the table at setup with
  `dubplate-server unidecode install`);
  nothing Python ships to the box. The oracle shim drives pinned beets
  under a uv-managed venv (`tools/venv/`) that the gate never requires and
  never runs (derivation is `make derive`, tagged `oracle`, workstation
  only). ruff (`ruff check` and `ruff format --check`) joins `scripts/check.sh`
  for every tracked `*.py` file, so the shim's bar is enforced, not
  aspirational, even though the gate can't execute it.

## Reviewers

Web UI work matches no streamer: dubplate replaces Navidrome at 0.8 and carries its own visual identity
(Geoff, 2026-09-24), so no reference screenshots and no visual-fidelity chain. Every user-path page is
rendered and read before Geoff sees it, and the `intake` uploader is Geoff's design review. Review at
pass end fans out by surface and passes the visual identity spec's path to the `web/` reviewers.

## Refused features

dubplate is for a group sharing music its members hold, and is built to be easy for that and
hard to turn into a distribution hub (ratified: `docs/superpowers/specs/2026-09-25-group-scale-principles-design.md`).
Never build, and never accept a contribution that adds:
- anything that fetches music: search-and-download, torrent, Usenet, or Soulseek integrations,
  or "fetch missing albums" automation (metadata sources are a separate question);
- federation, or any link between two servers' catalogs;
- a public library: anonymous browsing, public collections, or a link that opens a collection;
- a `download` larger than one album (song or album ids only; other kinds answer error 50);
- a share outside the six bounds (one item; lossy stream only; clamped expiry; the listener
  pause; an unguessable, unindexed link; bounds enforced server-side with the owner from login).

## Facts agents get wrong

- Failure pings are POST to `<url>/fail`; success is GET. The Worker 404s a
  GET to `/fail`; this was a live bug once.
- dubplate-server's state lives under `/var/lib/dubplate/` (persistent),
  never in a RuntimeDirectory (deleted at unit stop). The server uses NO
  runtime directory at all. Upload staging is the exception by design: it
  lives on the music volume beside the library, so a rename into the
  library never crosses a device.
- The album is the verdict unit: one failing track rejects the batch. Never
  admit a partial album. Separate tracks are the other kind: each file gets
  its own verdict, and the passing ones are accepted.
- Reject wording is contract (`friendly_reason`): the uploader page and
  emails share it verbatim; upload-check reuses it. The set lives in
  `internal/reject`.
- `env/dubplate-server.env.template` is the render manifest deploy.sh
  turns into `/etc/dubplate/server.env`: a bare `NAME=` row resolves from
  the age store, so adding one requires minting it (`secret-set.sh`,
  registry); a `NAME=value` row is a literal.
- A collection is a store row whose root the server allocates from its
  id, with no two roots nested, per
  `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`.
- Config keys are declared through the registry in each rung's own file,
  never by editing `internal/config/keys.go` (the config registry pass).

## Safety rails

- Before any executor touches a worktree, run `claude agents --json` for a live session whose cwd is that
  worktree, then `pgrep -f` with a pattern whose literal text never appears in the command (a bracket class,
  `pgrep -f 'worktrees/<nam>[e]'`, because the Bash tool's wrapper carries the command text; see
  `~/.claude/docs/bluefin-admin.md`), then `git status` for warm changes you did not author. STATUS records
  each running pass's session name and id.
- The box (`5.78.148.42`) is production: no writes without a quiet-box check; read-only ssh only during reviews.
- Never commit secrets, `.env` values, or `~/.local/secrets` content.
  The gate's gitleaks leg scans for them, but it misses a value that
  looks ordinary; review a diff by eye before staging anything that might
  carry one.
- Deploys go through `scripts/deploy.sh` (idempotent; it checks by
  default, and `--apply` writes, only from a clean committed tree). Never edit the box by hand and call it done: the repo
  is the source of truth.
