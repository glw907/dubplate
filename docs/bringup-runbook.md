# Bring-up runbook: from reprovisioned box to seeded library

Self-contained: an agent or session with no prior context executes this top to
bottom. State assumptions: server `musicbox` freshly recreated (2026-08-30) from the
corrected cloud-init (all five package fixes); volume `musicbox-data` survived with
Navidrome's DB (admin geoff@907.life) and the /srv/music tree; every secret exists in
`~/.local/secrets` (source it per shell); repo clean at or past `f054b91`. Never
print a secret value; name-and-length checks only. On-box work over SSH as root
(key auth); IP from `hcloud server ip musicbox`.

1. **Gate on first boot.** `cloud-init status` must be `done` (poll; two reboots are
   designed in). Then verify: `getenforce` = Enforcing; docker active; volume
   mounted; `flac ffmpeg sqlite3 rclone msmtp yq jq` all present; tree ownership
   2000:2000 with inbox setgid. Any miss: stop and report; do not patch the box.
2. **Deploy.** `bash scripts/deploy.sh --check`, review, then `bash scripts/deploy.sh`.
3. **Acceptance.** Containers healthy with zero restarts after 60s; tunnel healthy
   (read-only GET on cfd_tunnel b6781175-000b-40c3-a54a-05906ab25037); from the
   workstation, music.907.life serves Navidrome (no Access) and inbox.907.life
   challenges via Access; both timers active; a manual `systemctl start
   music-import` no-ops cleanly with a success ping; beets loads all three plugins
   (replaygain must load now that ffmpeg exists).
4. **First backup.** `systemctl start music-backup`; verify success in the journal,
   then from the workstation (RO token) `rclone lsjson` shows the pinned layout
   (`state/` holding both DB snapshots). Backup ping green.
5. **Seed.** rsync the nine ZIPs from the workstation `~/Downloads/` into
   `/srv/music/inbox/geoff/` (as root, then chown 2000:2000). Wait for the timer (or
   start music-import once). Track per-album outcomes: imported (visible via
   Subsonic API search on music.907.life with admin creds from env, never printed) vs
   review-queued (with reasons). Rejection or review emails are expected for any
   MusicBrainz misses; that path working is a pass, not a failure.
6. **Post-seed backup.** Run music-backup again; verify `library/` objects appear in
   R2; note total object count and approximate size.
7. **Monitoring truth.** All three push checks green (backup, import, disk pings
   observed in journals); `musicbox-navidrome` pull check now has a live endpoint —
   verify the pings Worker's next cron ticks record probe success (read-only
   `wrangler tail` briefly or KV state read of the pull check's ping:/alert: keys by
   name).
8. **Ledgers.** Update docs/STATUS.md (seeded state, per-album table, next actions:
   Task 7 restore drill, Task 8 accounts + onboarding, pings P2c drill) and
   HISTORY.md; commit with the standard footer
   (`Co-Authored-By: Claude <noreply@anthropic.com>`).

Report shape: per-step results; per-album outcome table; anything degraded;
couldNotDo.
