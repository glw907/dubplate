# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-i, and rung 6's 6a and 6b-i are closed. **5c-ii is closed** (admin gate, roster API, `web/` tree born gated; HISTORY, top entry). Beets 2.14.1 adopted. The visual identity spec (revision 1.2, `782d864`) is complete, 5c-iii's standing input.

## Next action

The 2026-09-25 unattended window is complete, at about 6.8M of its 12M cap (flag not reached).
Three plans are reviewed and await approval, each executable under every answer to its rulings:

- Matcher word boundary, `plans/2026-09-25-matcher-word-boundary.md` rev 3: Rulings 1-3.
- 6b-ii restore and HTTP projection, `plans/2026-09-25-rung-6b-ii-restore-http.md` rev 4: Rulings 1-9.
- 5c-iii pages, fonts, rendered gates, `plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`
  rev 4: Rulings 1-10.

**Morning question (Geoff), one sitting:**
1. Each plan's "Rulings awaiting owner" section. All 22 recommend yes, except where a ruling
   says otherwise. 6b-ii Ruling 1 recommends no, which splits the pass into 6b-ii-a (3.90M) and
   6b-ii-b (7.53M). 5c-iii Ruling 1 recommends yes, which splits it into 5c-iii (10.8M) and
   5c-iv (8.5M).
2. Rung 12: the roster API shares an origin with the Navidrome fallthrough (5c-ii security M2).
   Accept as a rung 12 threat-model entry, or act sooner?
3. Order: the matcher pass and 5c-iii both need lane 1 (the matcher touches corpus). The
   recommendation is matcher first, then 5c-iii; 6b-ii takes lane 2 in parallel. From 5c-iii's
   Task 3 on, every full gate launches Chromium (heavy lane).
4. Readings to ratify or reverse (conductor, 2026-09-25): one more targeted round on 5c-ii
   Fold A after its second `fix` (F-A3, accepted); the plans drafted alongside the fold rather
   than after both closes; the architecture reads' routing (ops and backup findings to 6b-ii,
   the rest to a cleanup pass; forward hooks removed there and re-added by the rung that
   consumes them); the 6b-i CLAUDE.md Gate line without its false "one validator" clause.

Also flagged: auth's revocation epoch has no production writer until the revoke verb; the
visual identity spec takes errata (Permanent Marker is Apache-2.0; R5.4 cites should be R5.5);
the workstation was on battery at the window's start.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | CLOSED on 5c-ii; free for the next code lane |
| 2 | worktree `lane-2-6b-i` | CLOSED; merged at `6f2c3e0` |

## Open items (carried)

- No deploy until 5d: `magic-link` env template lacks its nine keys. Owed: `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys' `doctor` row, Ruling 6's old-browser `Referrer-Policy` residual, the branding-in-both-schemes line, `deploy.sh`'s clean-`web/`-build requirement, 6a's and 6b-i's runbook rows (`KillMode=mixed`, rclone v1.74.3+, crypt remote, exit-7 remedy).
- To rung 7: guard counting, `--max-delete`, a remote-pruning regression floor, archive uploads. To rung 9: library filter, facets, trigram search.
- To rung 12: no non-admin maps onto a streamer admin; a restore before then loses stars/playlists; roster API/Navidrome-fallthrough shared origin (security M2, a question for Geoff); Ruling 7's availability question; re-validate every roster row's `streamer_username`; the header step's fallthrough reconciliation. To the revoke verb: the epoch store's first writer.
- To 5c-iii: the handoff in HISTORY's 5c-ii entry, the label-ink gate row (a11y W4), WCAG 1.4.11 on night-primary fills, S11, S12, I2.
- To 6b-ii: the ops interface's operations/sentinels; `forbidden`/`internal`/the network actor with their first caller; the lease/staging/remote-join rules a restore must respect; the backup reference a restore replaces; `internal/ops`'s `Seam`/catalog gap and sentinel duplication; backup's rclone-launch/copy-classification duplication (incl. `cmd/dubplate-server/backup.go`'s second remote-name validator).
- To the architecture cleanup pass: the 5c-ii reads (`internal/auth`, `internal/serve`, `internal/identity`), 6b-i's ops/local findings, 6a's cmd findings. To 6d: a store identity minted at create, the migration-wait key, serve's caller. To 6c: the `SAVEPOINT` check on `COMMIT;BEGIN;` (risk R13). L5/L6 (roster-POST throttle, n5 whole-roster doctor row): no pass assigned.
- 99 declared coverage gaps, none blocking. Staged legs warn: 326 process citations, 128 exported-unused. Fork B: 1.0 FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
