# musicbox status

## Current state

Task 3 (compose.yaml, systemd units, deploy.sh) is done. Task 4a (real
provisioning) is done: server `musicbox` is up at `5.78.148.42` (CPX21,
`alma-10`, `hil`, 2 TiB/month traffic) with volume `musicbox-data` (50GB)
attached, 8 of 8 on-box checks green. `bash scripts/check.sh` exits 0.

Task 5 (Cloudflare wiring) is **blocked on a token permission**: neither
`CLOUDFLARE_API_TOKEN` nor `CF_ZT_TOKEN` carries Cloudflare Tunnel:Edit, so
`POST .../cfd_tunnel` 403s on both — tunnel creation, its token, the ingress
config PUT, and the DNS CNAMEs all wait on this. Cloudflare Access is done
independently: `inbox.907.life` has a self-hosted Access app (one-time-PIN
policy for `geoff-login@907.life`); confirmed `music.907.life` has no Access
app. Task 6's Navidrome secrets are generated and stored
(`ND_PASSWORDENCRYPTIONKEY`, `NAVIDROME_ADMIN_USER`, `NAVIDROME_ADMIN_PASS`).
Full narrative: `docs/HISTORY.md`.

**Next action (Geoff, manual, blocking)**: add the Cloudflare Tunnel:Edit
permission group to `CLOUDFLARE_API_TOKEN` (id `1d508f1ab69df49d3ef9572dc1917273`,
"Cloudflare Admin 2026-07") via the dashboard — API token self-management is
deliberately not grantable through the API itself. Once added, resume Task 5:
create tunnel `musicbox`, `secret-set.sh MUSICBOX_TUNNEL_TOKEN --stdin`, PUT
`config/cloudflared-tunnel-config.json` as the ingress config, upsert proxied
CNAMEs `music`/`inbox` → `<tunnel-id>.cfargotunnel.com`, then verify the Access
challenge on `inbox.907.life` and `https://music.907.life/rest/ping`. Then
Task 4's first real `deploy.sh` run (cloudflared needs the tunnel token to come
up) and Task 6's remaining items (R2 lifecycle/lock, healthchecks.io checks,
first backup, seed content).

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

- **Task 5 tunnel-permission blocker (Geoff, manual)**: see Next action above.
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
