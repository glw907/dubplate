<!--
Draft STATUS.md, modeled on the main tree's current docs/STATUS.md
and on draft-STATUS-6b-i.md's shape, as it would read after 6b-ii-a
merges and lane 2 closes. The conductor fills the bracketed fields
(merge SHA, actual spend, architecture-read routing) during the fold.
Line count below ("## Current state" through EOF) is at most 60,
per the plan's Task 8 acceptance criterion.
-->

# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-i and 5c-ii, and rung 6's 6a
and 6b-i are closed. **6b-ii-a** (the ops seam's HTTP projection,
`<merge-sha>`, about 1.8M against 3.90M) is closed; HISTORY, top entry.
Beets 2.14.1 adopted. The visual identity spec (revision 1.2,
`782d864`) is complete.

## Next action

1. **6b-ii-b plans and runs next** (lane 2, fresh Task 0): Tasks 4A-7,
   the offline `restore` verb; ceiling 7.53M, flag 6.02M. Handoff in
   HISTORY's 6b-ii-a entry: `IntegrationValue`/`IntegrationEnabled`
   for R-6bii9's disabled-class row.
2. **Matcher word boundary and 5c-iii** continue in lane 1 per the
   unattended-window-2 plan, unaffected by this merge.
3. **To rung 14** (not yet planned): the projection's mount without a
   trailing slash on `StripPrefix`, the production session authorizer,
   the bearer channel, Ruling 3's answer (403 forbidden for a member
   lacking privilege; a no-identity caller is rung 14's own denial
   kind), and the 5c-ii security M2 question extended to the
   projection before the cookie mount lands.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | OPEN: matcher word-boundary / 5c-iii work, unaffected by this merge |
| 2 | worktree `lane-2-6b-ii-a` (branch `lane-2-6b-ii-a`, cut at `f6c8edb`) | CLOSED; merged at `<merge-sha>` |

## Open items (carried)

- No deploy until 5d (env template lacks `magic-link`'s nine keys):
  `doctor` exits 0, `ReadWritePaths`, trusted proxy, theme keys' row,
  runbook rows from 6a and 6b-i.
- To 6b-ii-b: `ServeOptions.Reconcile`/`backup.Reconcile` (6d's
  wiring), a store identity for restore to refuse another store's
  snapshot, the network actor's `member:<principal>` form.
- To rung 14: opshttp's mount, the production authorizer, the bearer
  channel, Ruling 3's answer, security L5-L7 (audit logging with actor
  and `client_ip`, `CancelJob`'s discarded actor / durable id in
  `jobs.actor`, the unauthenticated response shape), the M2 shared-
  origin question extended to the projection.
- To rung 12: the streamer database's place in a restore (0.8 forward
  hook; streamer versions re-keyed 2026-09-25, "2.0" is now 0.8).
- 99 declared coverage gaps, none blocking. Staged legs warn: 326
  process citations, 125 exported-unused (`ops.Codes`/`ops.CodeOf` now
  have real callers in `opshttp`, unmounted). Fork B: 0.8/0.9
  FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`,
  mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`,
  `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
