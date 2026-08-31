# dubplate status

## Current state

The estate is live and serving (Navidrome at music.907.life, inbox behind
Access, tunnel healthy). The architecture refinement pass is at close-out:
the 11-task fix wave plus lock follow-ups (M2b/M2c) are **accepted but
uncommitted in this working tree**, owned by the refinement session's close
ritual. The bring-up runbook execution (separate agent) is finishing steps
5–7 on the box: first real backup reached R2 (state/ layout verified), the
3.3 GB seed import is running, monitoring-truth findings pending. The pings
repo's hardening is committed (`40fdfa7`). Remote: github.com/glw907/dubplate
(private); committed history pushed.

## Rules for a new session in this repo

1. **Do not touch scripts/, systemd/, tests/, or config/ until the
   refinement close-out commit lands on master** — the uncommitted tree
   belongs to that ritual (wave + bring-up attribution + simplifier).
2. Then execute the Go rewrite:
   `~/.dotfiles/docs/superpowers/plans/2026-08-30-dubplate-go-rewrite.md`
   (fifteen tasks, pass-execute workflow; spec beside it). Verify no other
   executor holds the repo first.
3. The uploader and packaging initiatives queue behind it — see ROADMAP.md.

## Open items (carried)

- Reprovision/remediation of the live box's package set: partially overtaken
  by the bring-up run (bsdtar hand-installed 2026-08-31, cloud-init updated,
  uncommitted); ffmpeg/flac/sqlite3/rclone/msmtp state verified during that
  run's steps.
- Live application of the SSH firewall restriction (repo-side done in M11;
  apply with the box quiet).
- Musicbox docs drift sweep + DR drill: superseded — the drill becomes
  `dubplate restore` (rewrite T10b); the doc sweep folds into rewrite T13/T14.
- SELinux one-time check on /var/lib/musicbox at first real deploy (M2b
  reviewer note).
- pings known gap: a check with a permanently missing route: record is
  undeletable via remove-check (no --force escape); backlog.
- Task 8 of the original build plan (per-contributor accounts, credential
  delivery) — superseded by the uploader initiative for intake; Navidrome
  family accounts still owed.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| yq (box, cloud-init) | `mikefarah/yq` | `v4.53.6` |

FileBrowser (`gtstef/filebrowser 1.5.5-stable`) retires with the uploader.
Re-verify pins before any fresh provision.

**Next action**: refinement session lands the close-out commit; then a fresh
session runs the rewrite plan.
