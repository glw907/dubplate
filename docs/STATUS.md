# musicbox status

## Current state

Task 3 (compose.yaml, systemd units, deploy.sh) is done: `compose.yaml`, `config/`
(navidrome.toml, filebrowser.yaml, msmtprc, beets-config.yaml,
cloudflared-tunnel-config.json), the `music-import`/`music-backup` systemd units plus a
shared `musicbox-ping-failure@.service`, `env/musicbox.env.template`, and
`scripts/deploy.sh` all land. `bash scripts/check.sh` is green: shellcheck, the 16-test
bats suite, `docker compose config -q`, and a pattern-filtered `systemd-analyze verify`
(plain `--recursive-errors=no` doesn't suppress the workstation's missing-ExecStart-binary
noise; check.sh filters exactly that message and fails on anything else — confirmed
against a real injected syntax error). No infrastructure exists yet; `deploy.sh --check`
was exercised against an unreachable host instead. Full rationale for every Task 3
decision: `docs/HISTORY.md`'s Task 3 entry.

Next task: Task 4 (provision and deploy — real infrastructure).

Plan: `~/.dotfiles/docs/superpowers/plans/2026-08-30-music-vps-build.md`
Spec: `~/.dotfiles/docs/superpowers/specs/2026-08-30-music-vps-build-design.md`

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version | Released |
|---|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` | 2026-07-11 |
| FileBrowser Quantum | `gtstef/filebrowser` (upstream repo name `gtsteffaniak/filebrowser`) | `1.5.5-stable` | 2026-08-30 |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` | 2026-08-14 |
| beets (box, via uv) | PyPI `beets` | `2.13.1` | current at lookup |

Re-verify all four before Task 4's real deploy if this pass is more than a few days
stale — `docker compose config` never checks that a pinned tag still exists upstream.

## Open items

- **Task 0 (Geoff, manual, blocking Task 4+)**: credential batch not landed. `deploy.sh`
  already fails correctly, listing every missing env var by name.
- **CPX21 US traffic allowance**: unconfirmed; record the console-shown quota before
  Task 4 (if 1 TB, default family clients to transcoded streams).
- **`alma-10` image availability**: unconfirmed on Hetzner; `provision.sh` falls back to
  `alma-9` and says so loudly. Record which one actually ran once Task 4 runs.
- **cloudflared ingress is not set anywhere in compose.yaml on purpose**: a token-run
  tunnel is remotely managed, so a local ingress file would be silently ignored.
  `config/cloudflared-tunnel-config.json` is the exact API body Task 5 must PUT to the
  tunnel once it exists — skipping that step leaves cloudflared connected but serving
  nothing.
- **MUSICBOX_HC_DISK_URL / MUSICBOX_HC_NAVIDROME_URL are named but unwired**: only
  `import` and `backup` have real `ping_hc` calls today. Task 6 needs new disk-space and
  Navidrome-heartbeat check scripts + timers, not just the healthchecks.io checks
  themselves.
- **Ruling**: "no `latest`, pin explicitly" governs container tags and installed package
  versions (beets); host RPMs (`dnf-automatic`) and `uv` itself are intentional
  exceptions, not violations.
