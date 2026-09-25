# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-i, and rung 6's 6a and 6b-i are closed. **5c-ii is closed** (admin gate, roster API, `web/` tree born gated; HISTORY, top entry). Beets 2.14.1 adopted. The visual identity spec (revision 1.2, `782d864`) is complete, 5c-iii's standing input.

## Next action

**Resume prompt:** execute `docs/superpowers/plans/2026-09-25-unattended-window.md`. Three plans run as document lanes, no code lane open: the matcher word-boundary plan at its final fold; 6b-ii at fold verification; the 5c-iii plan (fonts, the three real pages, the rendered gates, the render-and-read, the identity spec's component layer), drafted against the carried items below, then `spec-plan-review` to "awaiting owner approval." Then the morning question.

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
