# dubplate status

## Current state

The estate is live, seeded, and monitored (Navidrome at music.907.life,
tunnel healthy). The bring-up runbook completed 2026-08-31: seven albums
imported (105 tracks), post-seed backup verified in R2 (`library/` 112
objects / 2.794 GiB, `state/` both DB snapshots). The architecture
refinement fix wave is committed AND deployed to the box (commit 17cd7fa +
interim deploy 2026-08-31): the failure-ping path is fixed (POST /fail)
and all four check timers are enabled. The pings repo's hardening is
committed (`40fdfa7`). Remote: github.com/glw907/dubplate (private).

## Rules for a new session in this repo

1. The Go rewrite is CLOSED on the box and in the repo (2026-08-31
   evening): the estate runs entirely on the dubplate binary (all four
   jobs green on supervised first runs; doctor fully green), box paths
   are /etc/dubplate + /opt/dubplate, pings checks are dubplate-*
   slugs, the shell layer and msmtp are deleted everywhere, and the
   repo lives at ~/Projects/dubplate. Post-hoc soak: overnight timer
   evidence (backup fires ~11:37 UTC) confirms unattended operation;
   the dead-man checks alert if not. Then the pass post-mortem.
   Kept identifiers: README's list. Spend ≈5.0M of the approved 5.5M.
2. Uploader plan APPROVED (Geoff 2026-08-31, two adversarial rounds):
   `docs/superpowers/plans/2026-08-31-dubplate-uploader.md`.
   Gate 3 (iPad picker) already CLOSED documentarily — see the plan's
   Gate outcomes; gates 1–2 run post-Pass-A against the real serve.
3. The uploader and packaging initiatives queue behind it — see ROADMAP.md.

## Open items (carried)

- Two albums review-queued from the seed, awaiting Geoff's manual review
  in `/srv/music/review/2026-08-31/geoff/`: Aim "Hinterland (Remastered)",
  illuminati hotties "POWER" (plus a TV Girl zine PDF).
- SSH firewall restriction: DECLINED 2026-08-31 — the workstation has no
  stable IP, so an allowlist would strand access on rotation. tcp/22 stays
  open, key-only auth. Durable option if ever wanted: SSH over the
  existing cloudflared tunnel.
- Two stale `musicbox-ping-failure@` units show failed from the pre-fix
  404 era; cosmetic. Clear with:
  `ssh root@5.78.148.42 systemctl reset-failed` (box write, Geoff or an
  approved session).
- STATUS/runbook called musicbox-navidrome a "pull" check; KV shows all
  four checks are push — the box's navidrome-check timer drives it.
- pings known gap: a check with a permanently missing route: record is
  undeletable via remove-check (no --force escape); backlog.
- Navidrome family accounts still owed (per-contributor intake itself is
  superseded by the uploader initiative).

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |

FileBrowser (`gtstef/filebrowser 1.5.5-stable`) retires with the uploader.
Re-verify pins before any fresh provision.

**Next action**: verify overnight timer evidence, then the rewrite
post-mortem. Uploader Pass A (U1-U8) is executing via pass-execute in
this repo; gates 1-2 run against the deployed serve after U12a.
