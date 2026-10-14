# Music VPS Build (musicbox) Implementation Plan

> **For agentic workers:** execute per-task through the implementer → diff-reviewer → gate
> chain (conductor conventions in the workstation CLAUDE.md). Steps use checkbox syntax.

**Goal:** Stand up the musicbox VPS end to end: inbox → headless beets → Navidrome at
music.907.life, R2 backup, monitoring, and the workstation repointed.

**Architecture:** One Hetzner CPX21 (Hillsboro) with a 50 GB volume at `/srv/music`. Docker
Compose runs Navidrome, FileBrowser Quantum, and cloudflared; beets and the import/backup
scripts run on the host as the `music` user under systemd timers. Cloudflare Tunnel
publishes both hostnames; Access guards the inbox. Everything reproducible from the
`~/Projects/musicbox` repo plus R2.

**Tech Stack:** AlmaLinux 10, cloud-init, hcloud CLI, Docker Compose, Navidrome, FileBrowser
Quantum, cloudflared, beets (uv), rclone, msmtp, bats-core for script tests.

**Spec:** `~/.dotfiles/docs/superpowers/specs/2026-08-30-music-vps-build-design.md`

**Token ceiling:** 900k. **Checkpoint interval:** 3 tasks (STATUS in the musicbox repo).
Model per dispatch: sonnet implementers, claude-opus-5 diff review, per convention.

## Global Constraints

- The box is cattle: nothing hand-edited on the VPS; every config file originates in the
  repo and lands via `deploy.sh`. `/srv/music` restores from R2.
- All secrets originate in the workstation age store (`secret-set.sh`, then
  `registry.md`); the box copy is `/etc/musicbox/env`, root:root 0600, written by
  `deploy.sh`. No secret value in the repo, ever.
- Image tags and package versions pinned explicitly; no `latest`, no auto-updater.
- One `music` system user (uid/gid fixed at 2000) owns `/srv/music`; containers run as
  2000:2000; inbox directories are setgid.
- SELinux stays enforcing (AlmaLinux default): every container bind mount in
  `compose.yaml` carries a `:Z` label; never `setenforce 0` as a fix.
- R2 bucket layout is pinned: `library/`, `state/`, `_archive/<tree>/<date>/`. One rclone
  invocation per tree, always `--exclude "/_archive/**"`.
- Bash scripts pass shellcheck; pure helper functions live in `scripts/lib.sh` and carry
  bats tests runnable on the workstation.
- Prose (README, onboarding doc) follows the Google developer-doc register; commits
  imperative with the standard co-author footer.

---

### Task 0: Credential batch (Geoff, manual — the single human gate)

Everything the build needs from a human, gathered once. Each value goes through
`secret-set.sh` and gets a `registry.md` row (scope + rotation).

- [ ] Hetzner Cloud account; API token with read/write → `HCLOUD_TOKEN`
- [ ] R2 API token, **read-write**, scoped to bucket `music-library` →
      `MUSICBOX_R2_ACCESS_KEY_ID` / `MUSICBOX_R2_SECRET_ACCESS_KEY`
- [ ] R2 API token, **read-only**, scoped to bucket `music-library` →
      `MUSIC_R2_RO_ACCESS_KEY_ID` / `MUSIC_R2_RO_SECRET_ACCESS_KEY`
- [ ] Fastmail app password for SMTP sending → `MUSICBOX_SMTP_PASSWORD` (user is the
      Fastmail account address)
- [ ] healthchecks.io account (free); API key → `HEALTHCHECKS_API_KEY`
- [ ] Contributor list: name, email, and role (`contributor` / `listener`) for each family
      member, for the Access policy, FileBrowser accounts, Navidrome accounts, and the
      notification map. Recorded in `musicbox/config/contributors.yaml` (names and emails
      are config, not secrets).

**Acceptance:** `~/.local/secrets` exports all seven variables; `registry.md` documents
each; `contributors.yaml` exists.

### Task 1: Repo scaffold and provisioning inputs

**Files:** create `~/Projects/musicbox/`: `README.md`, `.gitignore`, `docs/STATUS.md`,
`cloud-init/user-data.yaml`, `scripts/provision.sh`, `config/contributors.yaml`
(placeholder committed only if Task 0 hasn't landed).

**Produces:** a repo any later task deploys from; `provision.sh` is the only thing that
talks to hcloud.

- [ ] Git repo with README stating the architecture in one paragraph and pointing at the
      spec; STATUS per the three-ledger convention.
- [ ] `cloud-init/user-data.yaml`: SSH key (workstation pubkey), Docker CE install from
      its RHEL repo, dnf-automatic (security-only), `music` user (uid 2000, no shell
      login), mount unit for the
      volume at `/srv/music`, base directory tree (`library`, `inbox`, `review`, `state`,
      `staging`) owned 2000:2000 with setgid on `inbox`.
- [ ] `scripts/provision.sh`: idempotent hcloud calls creating firewall `musicbox-fw`
      (inbound: 22/tcp only), 50 GB volume `musicbox-data` (location `hil`), server
      `musicbox` (CPX21, `alma-10` — confirm via `hcloud image list`, fall back to
      `alma-9` and record in STATUS — location `hil`, the firewall, the volume, the
      user-data). Prints the server IP. Reads `HCLOUD_TOKEN` from sourced secrets.
- [ ] Constraint: record the console-confirmed US traffic allowance for CPX21 in STATUS
      (spec calls for it before family onboarding).

**Acceptance:** shellcheck clean; `cloud-init schema --config-file` validates; dry-run
mode (`provision.sh --check`) lists intended resources without creating any.

### Task 2: Import, backup, and wrapper scripts with tests

**Files:** create `scripts/music-import`, `scripts/music-backup`, `scripts/music-beet`,
`scripts/lib.sh`, `tests/lib.bats`.

**Interfaces (produced, relied on by Tasks 3 and 6):**
- `lib.sh`: `quiescent <dir> <minutes>` (0 if no file newer than N minutes, recursive);
  `safe_extract <archive> <destdir> <max_bytes>` (bsdtar, rejects path traversal and
  over-cap archives); `free_space_ok <path> <min_gb>`; `notify <email> <subject> <body>`
  (msmtp); `ping_hc <check-slug> [fail]` (healthchecks, reads URL from env).
- `music-import`: flock on `/run/lock/musicbox-import.lock`; per-contributor inbox scan →
  quiescence → extract to `staging/` → free-space abort → `beet import -q -l <logfile>` →
  move logged skips to `review/<date>/` → notify uploader+Geoff on skips → POST Navidrome
  scan API → success/fail ping. 60-day review sweep at end.
- `music-backup`: sqlite3 `.backup` of beets and Navidrome DBs into `state/`-staging →
  two rclone syncs per the pinned layout with `--backup-dir` → ping.
- `music-beet`: interactive `beet` under the same flock.

**Acceptance (TDD; bats first for every lib function):** bats suite covers quiescence
(nested-write case explicitly), traversal rejection (`../` member), size cap, free-space
threshold; all red before implementation, green after; shellcheck clean; `music-import`
run against a fixture inbox on the workstation (stub `beet`, `rclone`, `curl`, `msmtp` on
PATH) exercises the happy path and the skip path end to end. Commit per script.

### Task 3: Service configuration (compose, systemd, deploy)

**Files:** create `compose.yaml`, `config/navidrome.toml`, `config/filebrowser.yaml`,
`config/msmtprc`, `systemd/music-import.{service,timer}`,
`systemd/music-backup.{service,timer}`, `env/musicbox.env.template`, `scripts/deploy.sh`.

**Interfaces (consumes Task 2 script names/paths verbatim).**

- [ ] `compose.yaml`: Navidrome (pinned tag, user 2000:2000, library read-only mount,
      `ND_SCANNER_SCHEDULE` daily backstop, watcher disabled, encryption key from env),
      FileBrowser Quantum (pinned tag, per-contributor scoped roots under `inbox/`),
      cloudflared (tunnel token from env; ingress `music.907.life` → navidrome:4533,
      `inbox.907.life` → filebrowser port, 404 catch-all).
- [ ] Timers: import every 15 min; backup daily 11:30 UTC. Both `User=music`,
      `OnFailure=` ping-fail unit.
- [ ] `deploy.sh`: rsync repo config to the box, render `/etc/musicbox/env` from the
      template + sourced age-store values, install systemd units, `docker compose up -d`,
      install beets on the box via uv (pinned version, same config layout as the beets
      stow package plus quiet-mode section).
- [ ] `env/musicbox.env.template` names every variable Task 0 minted plus
      `ND_PASSWORDENCRYPTIONKEY` and the four healthchecks URLs (created in Task 6).

**Acceptance:** `docker compose config` validates; `systemd-analyze verify` passes on the
units; shellcheck clean; `deploy.sh --check` diffs without applying.

### Task 4: Provision and deploy

Real infrastructure; conductor watches reports, not shells.

- [ ] `brew install hcloud`; run `provision.sh`; record IP in STATUS.
- [ ] First `deploy.sh` run; verify from the workstation: SSH up, volume mounted at
      `/srv/music` with correct ownership, all three containers healthy, beets answers
      `beet version` as the `music` user, timers listed in `systemctl list-timers`.
- [ ] Record CPX21 traffic allowance from the console into STATUS.

**Acceptance:** every check above green, captured in the task report; nothing hand-edited
on the box (re-running `deploy.sh` is a no-op diff).

### Task 5: Cloudflare wiring (tunnel, DNS, Access)

All on zone `907.life` (id `a7c2b9103ec7d835d72f356489072e5b`), account
`120c269ad6d3dfbe6d63a0bb53758ca0`. MCP token is read-only for Access/tunnel writes; use
curl with `CLOUDFLARE_API_TOKEN` per the standing rule.

- [ ] Create named tunnel `musicbox` (API), store its token via `secret-set.sh`
      (`MUSICBOX_TUNNEL_TOKEN`), redeploy so cloudflared connects.
- [ ] DNS: proxied CNAMEs `music.907.life` and `inbox.907.life` → the tunnel.
- [ ] Access application on `inbox.907.life`: one-time-PIN policy for the emails in
      `contributors.yaml`; verify an Access challenge appears before FileBrowser's login.
- [ ] Verify `music.907.life` serves Navidrome with no Access interstitial (Subsonic
      clients must reach the API directly).

**Acceptance:** curl from outside shows Access on inbox, Navidrome login page on music;
`https://music.907.life/rest/ping` (Subsonic endpoint) answers.

### Task 6: R2 layout, lifecycle, lock, monitoring, first backup

- [ ] Check whether R2 object versioning has shipped; if yes, record in STATUS and prefer
      it (spec allows dropping `--backup-dir`); if no, proceed.
- [ ] Lifecycle rule: expire `_archive/` after 90 days. Bucket lock rule: retention on
      `_archive/` prefix so the box's RW token cannot delete it early.
- [ ] Create the four healthchecks.io checks via API (`musicbox-backup`,
      `musicbox-import`, `musicbox-disk`, `musicbox-navidrome`); store ping URLs via
      `secret-set.sh`; redeploy env.
- [ ] Run `music-backup` once by hand; verify objects under `library/` (empty is fine
      pre-content), `state/` holds both DB backups, ping registered.
- [ ] Seed content: rsync the nine Bandcamp ZIPs from the workstation's ~/Downloads
      into the geoff inbox and let the full pipeline process them (quiescence →
      extraction → headless import → review queue for any MusicBrainz misses →
      Navidrome scan → next backup includes them).
- [ ] Transcoding bring-up (Geoff, 2026-08-30): verify Navidrome's default Opus
      transcoding profile exists (enable ND_ENABLETRANSCODINGCONFIG only for the
      admin session that confirms it, then disable); record in the onboarding doc
      that Symfonium and Amperfy should stream original quality on wifi and Opus
      ~128 kbps on cellular, and that ReplayGain should be enabled in both.

**Acceptance:** all four checks green on the healthchecks dashboard; the seeded album
streams in Navidrome's web player; R2 layout matches the pinned prefixes exactly.

### Task 7: Restore drill (acceptance gate for "the box is cattle")

- [ ] `provision.sh` a scratch server `musicbox-drill` (same user-data, no volume needed
      at drill scale); deploy; restore `library/` and `state/` from R2 with the RO token;
      start Navidrome; confirm the seeded album, the admin account, and a test playlist
      survive.
- [ ] Destroy the scratch server and its DNS leftovers; record the drill result and
      elapsed steps in STATUS (this is the runbook for a real loss).

**Acceptance:** drill passes with zero reference to the live box; scratch resources gone.

### Task 8: Workstation repoint, accounts, onboarding

**Files:** modify `~/.dotfiles/bin/.local/bin/` (delete `music-sync`, add `music-pull`),
`~/.dotfiles/docs/superpowers/specs/` cross-reference note; create
`musicbox/docs/onboarding.md`; update the `music-library-setup` memory.

- [ ] `music-pull`: `rclone copy` (never sync) of `library/` → `~/Music` using the RO
      token env vars; shellcheck + stub-rclone test; `stow -R bin`.
- [ ] Navidrome accounts per `contributors.yaml` (admin creates via API); FileBrowser
      accounts for contributors; record who-has-what in STATUS, passwords delivered by
      Geoff out of band with the unique-password note from the spec.
- [ ] `docs/onboarding.md`: per-platform client setup (Symfonium, Amperfy, the iPad
      stereo setup), how to upload, what the review email means. Editor-copy register
      (Microsoft), since the audience is family.
- [ ] Update memory and dotfiles STATUS/HISTORY per the ledger convention; commit both
      repos.

**Acceptance:** `music-pull` mirrors the seeded album locally; onboarding doc passes
Vale; a family member could go from doc to playing music without asking anything.

---

## Self-review notes

Spec coverage checked section by section: infrastructure (T1/T4), permissions model
(T1/T3), inbox+Access (T3/T5), import pipeline including unzip/quiescence/flock/notify
(T2), Navidrome scan handling (T2/T3), backup layout, SQLite staging, lifecycle+lock
(T2/T6), restore drill (T7), monitoring (T6), secrets (T0, threaded through), workstation
changes (T8), costs/traffic-allowance check (T1/T4), out-of-scope respected (no playlist
tooling, no transcoded mirror). Interface names used in T3/T6 match T2's produced names.
No placeholder steps; the two config values unknowable before build time (pinned image
tags, traffic allowance) are named as explicit record-at-build constraints rather than
gaps.
