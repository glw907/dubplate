# Musicbox architecture refinement pass: plan

**Goal:** Adversarial review of the musicbox estate as built, producing verified,
ranked findings, then a fix wave sized by triage.

**Charter:** `~/.dotfiles/docs/superpowers/specs/2026-08-30-architecture-refinement-charter.md`
(frame ratified 2026-08-30). Geoff's kickoff steers, same day: start with a
considered adversarial review; general systems setup is in scope alongside code;
workflow mode approved. Plan adversarially reviewed at kickoff (opus agent, 8
defects found); this version folds all 8.

**Budget:** Review phase ceiling 2.5M tokens; whole-pass ceiling 6M (fix wave
re-sized at triage). Checkpoints: STATUS write at review completion, at fix-wave
start, and every 4 fix-wave tasks. Runaway guard armed on the workflow.

**Concurrency note (kickoff):** a pre-existing background agent in this session is
executing the bring-up runbook against the live box (album seed rsync in flight).
The box is hands-off until it reports done; Phase B harvests its runbook-execution
results rather than re-running them. `~/.dotfiles` has dirty secrets files
(`secrets/registry.md`, `secrets/values.age`, `claude/.claude/CLAUDE.md`); those
paths are excluded from finder scope.

## Estate manifest (absolute paths; pasted into every finder dispatch)

- `~/Projects/musicbox`: `cloud-init/user-data.yaml`, `compose.yaml`,
  `env/musicbox.env.template`, `config/` (beets-config.yaml,
  cloudflared-tunnel-config.json, contributors.yaml, filebrowser.yaml, msmtprc,
  navidrome.toml, reject-email.txt), `scripts/` (check.sh, deploy.sh,
  firewall-rules.json, lib.sh, music-backup, music-beet, music-import,
  musicbox-ping-failure, provision.sh), `systemd/` (5 units/timers), `tests/`
  (lib.bats, music-import.bats), `docs/` (STATUS.md, HISTORY.md,
  bringup-runbook.md), `README.md`, `ROADMAP.md`.
- `~/Projects/pings`: `src/` (index.ts, check-do.ts, logic.ts, email.ts,
  env.d.ts), `scripts/` (add-check, remove-check, check.sh, provision.sh),
  `test/` (3 vitest suites), `tests/scripts.bats`, `wrangler.jsonc`, `README.md`.
- Workstation: `~/.dotfiles/bin/.local/bin/music-sync`.
- Spec chain: `~/.dotfiles/docs/superpowers/specs/2026-08-30-music-library-design.md`,
  `.../2026-08-30-music-vps-build-design.md`,
  `~/.dotfiles/docs/superpowers/plans/2026-08-30-music-vps-build.md`, and the
  charter above.
- Excluded: `~/.dotfiles/secrets/*` (dirty), the live box (Phase B), node_modules,
  lockfiles, `worker-configuration.d.ts`.

## Phase A: review fan-out (Workflow)

**Step 0 — Cloudflare evidence probe** (sonnet, medium): read-only API probe of
live CF state — tunnel config, Access apps on `inbox.907.life`/`music.907.life`,
R2 bucket and token scope, Worker routes/bindings for pings. Output feeds the
security and drift finders as evidence. Read-only GETs only; the box executor
touches none of this.

**Step 1 — nine finder lenses**, read-only, no ssh, no repo writes, each returning
a severity-ranked top ≤15 structured findings (dimension, artifact, file:line
where applicable, claim, evidence, severity, factual-or-judgment, fix sketch,
confidence, pass-2 flag, needs-box flag). Models: sonnet effort high, except
security and altitude on opus (cross-model diversity; the verify wave is opus, so
sonnet finders restore find/verify diversity).

1. **simplification** (sonnet) — repeated shell idioms that belong in lib.sh;
   near-dupe blocks in deploy.sh; fix-round patches that would be written
   differently designed whole; dead code and unused flags.
2. **systems-setup** (sonnet) — provisioning idiom: cloud-init shape, compose.yaml
   (pins, healthchecks, restart policy, networks, volumes), systemd units/timers
   (hardening, RandomizedDelay, Persistent), SELinux handling, dnf-automatic,
   EPEL/RPM Fusion ordering; and whether `check.sh` (both repos, written "for CI")
   is actually wired to run anywhere.
3. **drift** (sonnet) — factual claims a fresh reader would act on, across README,
   STATUS, HISTORY, runbook, the spec chain, and `music-sync`'s own header docs,
   checked against the repos; phantom fields, stale package lists, doc-pinned
   versions vs file-pinned. Claims needing the box get the needs-box flag, not a
   guess.
4. **security** (opus) — the surface as one system: secrets flow (env template,
   deploy.sh transport, msmtprc, beets config, wrangler secrets), systemd
   hardening not yet applied (ProtectSystem, NoNewPrivileges), Access/tunnel
   boundary and R2 token scope judged against the CF probe evidence (not the
   repo's description of it), credential handling in music-import's Subsonic call
   and FileBrowser bootstrap, pings auth model.
5. **altitude** (opus) — what in musicbox is estate-level and vice versa;
   music-sync placement; musicbox vs pings provision.sh duplication; what belongs
   in shared tooling.
6. **test-quality** (sonnet) — bats + vitest suites: redundancy, fixtures that no
   longer earn their runtime, stub fidelity vs the real box, coverage gap classes.
7. **operational-truth** (sonnet) — disjoint from lens 3 by carve-out: not doc
   facts, but operational claims — the known-limitation list re-examined for
   entries that quietly became load-bearing (Navidrome cron backstop, non-fatal
   replaygain failure), recovery paths that have never run, single points of
   failure the docs shrug at. Box-dependent claims flagged needs-box.
8. **failure-modes** (sonnet) — runtime correctness of the shell layer: prove live
   bugs (falsy-zero, fail-open, set -e-in-pipeline, quoting, signal/trap,
   partial-failure states) in `music-import`, `lib.sh`, `music-backup`,
   `deploy.sh`, `provision.sh` — the bug in the code, not the missing test.
9. **workers** (`cloudflare-workers-reviewer` agent, sonnet) — pings domain
   review: `src/` (Durable Object alarm semantics, storage, retry in check-do.ts),
   wrangler.jsonc, bindings, limits, edge-runtime gotchas, `add-check`/
   `remove-check` scripts.

**Step 2 — verify wave** (opus, effort high, ≤8 agents): findings grouped by
artifact (bin-packed so one verifier owns all lenses' findings on an artifact and
merges conceptual dupes). Rubric split by claim type: factual claims are
confirmed or refuted against the code; judgment claims (altitude, design-whole,
simplification) are endorsed or confidence-downgraded with a stated
counter-argument, never killed; box-dependent claims get verdict needs-box for
Phase B. Seeded finding entering the wave directly: `music-sync`'s header
documents a PikaPods+R2 topology, not the Hetzner box (found during plan review).

## Phase B: box reality probe

After the bring-up agent reports done and `pgrep` shows no executor on the box:
one agent, read-only ssh, resolves every needs-box verdict (installed packages vs
cloud-init, unit/timer states, SELinux mode, compose ps, live tunnel config vs
`config/cloudflared-tunnel-config.json`, R2 layout) and harvests the bring-up
agent's runbook-execution results as the charter's "runbooks executed
skeptically" evidence.

## Phase C: triage and gate

Main loop ranks confirmed + endorsed findings by severity × fix cost × ownership
(this pass / pass-2 / wontfix), writes STATUS, presents the ranked list plus a
proposed fix-wave task list to Geoff as the pass's combined question. Fix wave
runs as implementer → diff-reviewer → gate chains, checkpoint every 4 tasks.

## STATUS checkpoint — Phase A complete (2026-08-30)

Phase A ran as workflow `wf_c6eee9bd-5ed`: 18 agents, ~1.77M subagent tokens
(ceiling 2.5M), zero errors. 72 raw findings → 40 confirmed, 12 endorsed, 20
downgraded, 5 duplicates, 0 refuted, 0 needs-box. Full findings with verdicts:
`2026-08-30-musicbox-arch-refinement-findings.json` (this directory, committed).
CF probe notable: `music.907.life` has no Access app (Navidrome auth is the only
gate) — decision item, likely intentional for Subsonic clients. Phase B shrinks
to harvesting the bring-up agent's runbook execution plus a post-fix live
failure-ping and DR-drill verification (no needs-box verdicts remained).
musicbox `docs/STATUS.md` deliberately not touched yet: the pre-clear bring-up
agent may still edit it. Next: triage question to Geoff, then fix wave.

## Out of scope

New features, upload-check work, Immich, anything merely nicer rather than
simpler, truer, or safer (charter). No edits to any repo during Phase A/B.
