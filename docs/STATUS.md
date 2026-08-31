# musicbox status

## Current state

**musicbox is deployed and live** as of 2026-08-31. `scripts/deploy.sh` has run
successfully against `5.78.148.42`; all three containers (`navidrome`,
`filebrowser`, `cloudflared`) are up with zero restarts, the tunnel reads
`healthy`, and both public hostnames serve through it correctly (Access
challenge on `inbox.907.life`, none on `music.907.life`). Six real bugs
surfaced by this first live run (a fixture-backed test suite can't catch
host-environment issues) are fixed and committed; two package-provisioning
gaps remain open, blocking `music-backup` entirely and degrading
`music-import`. Full narrative, root causes, and every fix: the 2026-08-31
"Task 4/6: first live deploy" entry in `docs/HISTORY.md`.

### Acceptance sweep (2026-08-31)

| # | Check | Result |
|---|---|---|
| 3a | `docker compose ps` + 60s settle, restart counts | **Pass** — all 3 containers running, RestartCount=0 |
| 3b | `cfd_tunnel` API status | **Pass** — `status: healthy`, connections active |
| 3c | `/rest/ping`, `music.907.life` web, `inbox.907.life` Access | **Pass** — all three behave as specified |
| 3d | Timers active; first `music-import` run | **Pass, degraded** — clean no-op run, success ping; replaygain/ffmpeg load fails (non-fatal, empty inbox); Navidrome scan-trigger 404s (stale native-API assumption, non-fatal — cron backstop covers it) |
| 3e | First `music-backup` run; R2 `state/` layout | **Blocked** — `sqlite3`/`rclone` missing on box, neither ever in cloud-init; R2 confirmed empty via read-only `lsjson` |
| 3f | Navidrome admin bootstrap + login | **Pass** — `POST /auth/createAdmin` then `POST /auth/login`, verified via API only |

## Open items

- **`ffmpeg`/`flac`/`sqlite3` fixes are committed to `cloud-init/user-data.yaml`
  but not yet applied to this already-provisioned box** (cloud-init only runs
  at first boot). Needs Geoff present to run the remediation over SSH or a
  full reprovision — a subagent's `dnf install` attempt was refused by the
  Claude Code auto-mode classifier this pass (same shape as Task 5's
  tunnel-recreate refusal).
- **`rclone`/`msmtp` are missing on the box AND were never in cloud-init at
  all** — a real spec gap. Blocks `music-backup.service` entirely and leaves
  `lib.sh`'s `notify()` unverified. Needs each package's AlmaLinux 10 install
  path looked up before adding (an unattended `set -e` first-boot script is
  the wrong place to guess).
- **`trigger_navidrome_scan()` calls a native admin endpoint that doesn't
  exist in pinned Navidrome `0.63.2`** (moved to the Subsonic API's
  `startScan` action). Non-fatal — the daily cron scan covers it — needs a
  rewrite to Subsonic query-param auth.
- **`MUSICBOX_HC_DISK_URL`/`MUSICBOX_HC_NAVIDROME_URL` are minted but
  unwired**: only `import`/`backup` have real `ping_hc` calls; Task 6 needs
  disk-space and Navidrome-heartbeat scripts + timers too.
- **cloudflared ingress is not set anywhere in compose.yaml on purpose**: a
  token-run tunnel is remotely managed; `config/cloudflared-tunnel-config.json`
  is the record of what's PUT to the tunnel's remote config, not something
  compose.yaml or deploy.sh reads.
- **Ruling**: "no `latest`, pin explicitly" governs container tags and
  package versions (beets); `dnf-automatic` and `uv` itself are exceptions.
- Task 8 (per-contributor FileBrowser accounts, Navidrome family accounts,
  seed content, out-of-band credential delivery to Geoff) has not started.

**Next action**: with Geoff present, run the ffmpeg/flac/sqlite3 remediation
on the live box (or reprovision), confirm a real `music-backup` run and a
`music-import` run with actual FLAC content, look up `rclone`/`msmtp`'s
AlmaLinux 10 install path and add both to cloud-init, then Task 8.

Plan: `~/.dotfiles/docs/superpowers/plans/2026-08-30-music-vps-build.md` — Spec:
`~/.dotfiles/docs/superpowers/specs/2026-08-30-music-vps-build-design.md`

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version | Released |
|---|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` | 2026-07-11 |
| FileBrowser Quantum | `gtstef/filebrowser` (upstream repo name `gtsteffaniak/filebrowser`) | `1.5.5-stable` | 2026-08-30 |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` | 2026-08-14 |
| beets (box, via uv) | PyPI `beets` | `2.13.1` | current at lookup |
| yq (box, cloud-init) | GitHub `mikefarah/yq` (`yq_linux_amd64`) | `v4.53.6` (sha256 `c5f056448f973ae7d39b5401949648a78f2dc1947d6a8eb65be60d5c504b9385`) | current at lookup |

Re-verify all five before the next fresh provision if stale — neither
`docker compose config` nor cloud-init's `sha256sum -c` catch a pin drifted
upstream.
