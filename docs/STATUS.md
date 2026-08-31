# musicbox status

## Current state

Task 3 (compose.yaml, systemd units, deploy.sh) is done; `bash scripts/check.sh` is green.
Task 4a (real provisioning) is done: server `musicbox` is up at `5.78.148.42` (CPX21,
`alma-10`, `hil`, 2 TiB/month traffic) with volume `musicbox-data` (50GB) attached, built
from `cloud-init/user-data.yaml` at commit `3fa905d`. Full narrative: `docs/HISTORY.md`.

**Task 4a evidence, re-verified 2026-08-31 against `root@5.78.148.42` — 8 of 8 green:**

| Check | Command | Result |
|---|---|---|
| cloud-init | `cloud-init status --wait` | `status: done` — PASS |
| volume mount | `mountpoint -q /srv/music` | mounted — PASS |
| tree ownership | `ls -lnd /srv/music/{library,inbox,review,state,staging}` | all `2000 2000`; `inbox` is `drwxrwsr-x` (2775, setgid) — PASS |
| music user | `id music` | `uid=2000(music) gid=2000(music) groups=2000(music)` — PASS |
| docker | `systemctl is-active docker` | `active` — PASS |
| SELinux | `getenforce` | `Enforcing` — PASS |
| sshd password auth | `sshd -T \| grep -i '^passwordauthentication'` | `passwordauthentication no` — PASS |
| yq / jq | `/usr/local/bin/yq --version` / `jq --version` | `v4.53.6` / present (`jq-` is a cosmetic upstream version-string quirk on this AlmaLinux 10 build, not a functional defect) — PASS |
| fstab | `grep srv/music /etc/fstab` | exactly one line: `UUID=6614e3dd-9014-4949-9f4c-effcd317bd3a /srv/music ext4 defaults,nofail,x-systemd.device-timeout=30s 0 2` — PASS |

Deployed sha256 (box `/var/lib/cloud/instance/user-data.txt`) equals committed sha256
(`cloud-init/user-data.yaml` at commit `3fa905d`), both
`196e76fbc51d5a5828c6a378761d480b790160cc9fd07fa7a1f41ab1557dfa5a`. `bash scripts/check.sh` exits 0.

Next task: Task 4/5 — deploy the compose stack (`scripts/deploy.sh`) and wire the
Cloudflare tunnel, once Task 0's credentials land (see Open items below).

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

Re-verify all five before Task 4's deploy if this is stale — neither `docker compose
config` nor cloud-init's `sha256sum -c` catch a pin drifted upstream.

## Open items

- **Task 0 (Geoff, manual, blocking Task 4+)**: credential batch not landed. `deploy.sh`
  already fails correctly, listing every missing env var by name.
- **cloudflared ingress is not set anywhere in compose.yaml on purpose**: a token-run
  tunnel is remotely managed, so a local ingress file would be silently ignored.
  `config/cloudflared-tunnel-config.json` is the exact API body Task 5 must PUT to the
  tunnel once it exists — skipping that leaves cloudflared connected but serving nothing.
- **MUSICBOX_HC_DISK_URL / MUSICBOX_HC_NAVIDROME_URL are named but unwired**: only
  `import`/`backup` have real `ping_hc` calls today. Task 6 needs new disk-space and
  Navidrome-heartbeat check scripts + timers, not just the healthchecks.io checks.
- **Ruling**: "no `latest`, pin explicitly" governs container tags and installed package
  versions (beets); host RPMs (`dnf-automatic`) and `uv` itself are intentional
  exceptions, not violations.
