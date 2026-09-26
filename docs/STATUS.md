# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a-5c-iii, and **rung 6 in full** (6a, 6b-i, 6b-ii-a, 6b-ii-b) closed.
**5c-iii** (fonts, theme, rendered leg, document head, component layer; fold `81988b1`..`8588dd0`,
fix `db1e111`; 10.8M ceiling, ~12.6M spent) and **6b-ii-b** (the offline `restore` verb,
`c12ef83..392d1cd`, ~3.5M against 7.53M, merged range accepted by a fresh `diff-reviewer`) both
closed 2026-09-25, lanes 1 and 2; HISTORY's top two entries, then 6b-ii-a and the matcher pass.
Beets 2.14.1 adopted; visual identity spec at revision 1.3 (`45b80cb`).

## Next action

- **Lane 1 runs 5c-iv's D-ii and E**: Task 0/7 accepted (`wf_9508455c-ca8`); D-ii (8, 9a, 9b)
  running, then the register gate, then E (10, reviews, 11). Ceiling ~9.2M.
- **Lane 2 runs pass 6c** in worktree `lane-2-6c` (cut at `392d1cd`, rev 2.2 `50a1cf8`, 5.4M
  ceiling), segment A (0-3) running; then the transliteration pass (3.3M, approved).

**Overnight 2026-09-25 (Geoff):** full queue 18M cap, raised to 27M for 6c. 5d (reviewed, rev 2.2)
awaits Geoff's morning approval on ceilings and the attended first-deploy sitting; 5d-i launches
after 5c-iv closes and 6c merges. **Resume:** read this file and memory
`session-2026-09-25-handoff`, then continue both lanes from their journals; box read-only,
nothing deploys until 5d (memory `overnight-window-2026-09-25`).

## Lane registry

This session conducts both code lanes; a separate docs session owns the document lane (photos).

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | OPEN: 5c-iv D-ii (8, 9a, 9b) running; then E (10, reviews, 11) |
| 2 | worktree `lane-2-6c` (cut at `392d1cd`) | OPEN: 6c segment A running; then translit |
| photos | main tree, docs session | OPEN: photo curation round one; ceiling 5.5M |

Parallel work uncapped ~4h beyond the 14.5M window (flag 13.4M, Geoff 2026-09-25); per-pass spend
in HISTORY; 0 attended interactions after approval tonight.

## Open items (carried)

- No deploy until 5d: `magic-link` env template lacks its nine keys; `doctor` exits 0,
  `ReadWritePaths`, the trusted proxy, `deploy.sh`'s clean-`web/`-build requirement, each pass's
  runbook rows (6a, 6b-i, 6b-ii-b: "runbook owed to rung 16" in each HISTORY entry).
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's svelte S13/a11y S7,
  reject-reason wording, CREDITS output. To rung 9: filter, facets, trigram search. To rung 12:
  non-admin/streamer-admin mapping, restore-before-12 losing stars/playlists, roster/Navidrome
  shared origin (M2), the streamer database's restore place. To the revoke verb: the epoch
  store's first writer.
- To 6d (unplanned): `ServeOptions.Reconcile`/`backup.Reconcile` wiring, a store identity minted
  at create, the migration-wait key, a `member:<principal>` actor form if needed. To 5c-iv:
  HISTORY's 5c-iii "5c-iv handoff". To 6c: the `SAVEPOINT` check, stale `internal/config` doc.
  To 5d-i: `NewTrustedProxyConfig` has no caller.
- Errata owed and architecture-cleanup reads, each with full file:line, all in HISTORY: 5c-iii's
  nine errata plus latent items; 6b-ii-b's three (pre-restore-copy retention, the privileged-
  fields row, R-6bii11's line); 5c-ii's, 6b-i's, 6a's, window 2's five, 5c-iii's, and 6b-ii-b's
  reads plus carried non-blocking items; matcher re-run trigger `strdist_test.go:452`.
- 99 declared coverage gaps, none blocking. Staged legs: 326 process citations, 126
  exported-unused (up one, 6b-ii-b). Fork B: 0.8/0.9 FLAC-only. Readings taken (reversible, see
  recent HISTORY): the streamer spec's seven open items, `init`/`upgrade` naming owed.
- Photo lane, after lane 1 closes: a `scripts/check.sh` integrity line, the Ruling 3 photo-column
  erratum to rung 7, a pre-release source-URL recheck, and a blob-retention call. Pinned:
  Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`, unidecode
  `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
