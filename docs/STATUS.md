# musicbox status

## Current state

Task 3 (compose.yaml, systemd units, deploy.sh) is done. Task 4a (real
provisioning) is done: server `musicbox` is up at `5.78.148.42` (CPX21,
`alma-10`, `hil`, 2 TiB/month traffic) with volume `musicbox-data` (50GB)
attached, 9 of 9 on-box checks green. Import pipeline's quality gate (spec
step 4), `ping_hc` hardening, a flock-held-lock success ping, beets config
sync, and an Opus-review fix round (quoting bug, mv-collision safety, CRB)
landed this pass; `bash scripts/check.sh` exits 0 (35 bats cases).

Task 5 (Cloudflare wiring) is **done at the control-plane level**. Geoff added
Cloudflare Tunnel:Edit to `CLOUDFLARE_API_TOKEN` 2026-08-30; on 2026-08-31 the
tunnel `musicbox` (id `66fa450c-5abb-459d-a1f8-b80d8d19e07c`, `config_src:
cloudflare`) was created, its token stored as `MUSICBOX_TUNNEL_TOKEN` (via
`secret-set.sh --stdin`, piped directly from the API response), its ingress
PUT from `config/cloudflared-tunnel-config.json` and verified by GET
(`music.907.life` → `http://navidrome:4533`, `inbox.907.life` →
`http://filebrowser:80`, 404 catch-all), and DNS CNAMEs `music.907.life` /
`inbox.907.life` created proxied → `66fa450c-5abb-459d-a1f8-b80d8d19e07c.cfargotunnel.com`
on zone `907.life`. Cloudflare Access is done independently (`inbox.907.life`
has a self-hosted Access app, one-time-PIN for `geoff-login@907.life`,
re-confirmed as the only Access app on the musicbox surface; `music.907.life`
confirmed Access-free). Task 6's Navidrome secrets are generated and stored.
`cloudflared` is not yet running on the box, so both hostnames 530 until the
first real deploy — expected, not a defect. Full narrative: `docs/HISTORY.md`.

**Next action**: Task 4's first real `deploy.sh` (which starts `cloudflared`
with `MUSICBOX_TUNNEL_TOKEN` and should clear the 530s), then verify Access on
`inbox.907.life` and `https://music.907.life/rest/ping`, then Task 6's
remainder (R2 lifecycle/lock, healthchecks.io checks, first backup, seed
content).

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

Re-verify all five before Task 4's deploy if stale — neither `docker compose config`
nor cloud-init's `sha256sum -c` catch a pin drifted upstream.

## Open items

- **Resolved: Task 5 tunnel-permission blocker.** Cloudflare Tunnel:Edit added
  to `CLOUDFLARE_API_TOKEN` 2026-08-30; tunnel created and wired 2026-08-31
  (see Current state above).
- **Resolved: flac/ffmpeg's missing repos.** Ruling, 2026-08-30: enable EPEL
  and RPM Fusion free in cloud-init's runcmd (key imported, no `--nogpgcheck`)
  since both parse untrusted uploads and must stay on dnf-automatic's path.
- **cloudflared ingress is not set anywhere in compose.yaml on purpose**: a
  token-run tunnel is remotely managed. `config/cloudflared-tunnel-config.json`
  is the API body already PUT to the tunnel's remote configuration (Task 5,
  2026-08-31); it stays in the repo as the record of that config, not as
  something compose.yaml or deploy.sh reads.
- **MUSICBOX_HC_DISK_URL / MUSICBOX_HC_NAVIDROME_URL are named but unwired**:
  only `import`/`backup` have real `ping_hc` calls; Task 6 needs disk-space
  and Navidrome-heartbeat scripts + timers too.
- **Ruling**: "no `latest`, pin explicitly" governs container tags and package
  versions (beets); `dnf-automatic` and `uv` itself are intentional exceptions.
