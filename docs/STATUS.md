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

1. Execute the Go rewrite:
   `~/.dotfiles/docs/superpowers/plans/2026-08-30-dubplate-go-rewrite.md`
   (fifteen tasks, pass-execute workflow — workflow mode authorized by
   Geoff 2026-08-31; spec beside it). Verify no other executor holds the
   repo first.
2. The uploader and packaging initiatives queue behind it — see ROADMAP.md.

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
| yq (box, cloud-init) | `mikefarah/yq` | `v4.53.6` |

FileBrowser (`gtstef/filebrowser 1.5.5-stable`) retires with the uploader.
Re-verify pins before any fresh provision.

**Next action**: run the Go rewrite plan (rule 1).
