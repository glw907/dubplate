# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-i, and rung 6's 6a and 6b-i are closed. **5c-ii is closed** (admin gate, roster API, `web/` tree born gated; HISTORY, top entry). Beets 2.14.1 adopted. The visual identity spec (revision 1.2, `782d864`) is complete, 5c-iii's standing input.

## Next action

Three plans are approved (Geoff, 2026-09-25), their rulings recorded in each plan:

- Matcher word boundary, `plans/2026-09-25-matcher-word-boundary.md`: lane 1, first; 2.75M.
- 6b-ii restore and HTTP projection, `plans/2026-09-25-rung-6b-ii-restore-http.md`: split into
  6b-ii-a (the projection, 3.90M) then 6b-ii-b (restore, 7.53M); lane 2, in parallel.
- 5c-iii pages, fonts, rendered gates, `plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`:
  split into 5c-iii (10.8M) then 5c-iv (8.5M); lane 1 after the matcher pass. From its Task 3 on,
  every full gate launches Chromium (heavy lane).

**Resume prompt:** execute `docs/superpowers/plans/2026-09-25-unattended-window-2.md` (Geoff's grants
inside: matcher and 6b-ii-a in parallel, then 5c-iii segments in lane 1; 11M cap; read-only box).
Read memory `session-2026-09-25-handoff`. Rung 12 owns the roster API's shared origin
with the Navidrome fallthrough (5c-ii security M2) as a threat-model entry (Geoff, 2026-09-25).

Readings taken (conductor, 2026-09-25; reversible): one more targeted round on 5c-ii Fold A after
its second `fix`; the plans drafted alongside the fold; the architecture reads' routing; the 6b-i
CLAUDE.md Gate line without its false "one validator" clause. Owed: the visual identity spec's
errata (Permanent Marker is Apache-2.0; R5.4 cites read R5.5; the header focus halo).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | OPEN: matcher word-boundary pass (window 2, 2026-09-25); ceiling 2.75M, flag 2.2M; then 5c-iii segments |
| 2 | worktree `lane-2-6b-ii-a` (branch `lane-2-6b-ii-a`, cut at `f6c8edb`) | OPEN: 6b-ii-a (window 2, 2026-09-25); ceiling 3.90M, flag 3.12M |

Window 2 spend against the 11M cap: 0 at launch. AC online at launch; `npm ci` done in the worktree.

## Open items (carried)

- No deploy until 5d: `magic-link` env template lacks its nine keys. Owed: `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys' `doctor` row, Ruling 6's old-browser `Referrer-Policy` residual, the branding-in-both-schemes line, `deploy.sh`'s clean-`web/`-build requirement, 6a's and 6b-i's runbook rows (`KillMode=mixed`, rclone v1.74.3+, crypt remote, exit-7 remedy).
- To rung 7: guard counting, `--max-delete`, a remote-pruning regression floor, archive uploads. To rung 9: library filter, facets, trigram search.
- To rung 12: no non-admin maps onto a streamer admin; a restore before then loses stars/playlists; roster API/Navidrome-fallthrough shared origin (security M2, a threat-model entry, Geoff 2026-09-25); Ruling 7's availability question; re-validate every roster row's `streamer_username`; the header step's fallthrough reconciliation. To the revoke verb: the epoch store's first writer.
- To 5c-iii: the handoff in HISTORY's 5c-ii entry, the label-ink gate row (a11y W4), WCAG 1.4.11 on night-primary fills, S11, S12, I2.
- To 6b-ii: the ops interface's operations/sentinels; `forbidden`/`internal`/the network actor with their first caller; the lease/staging/remote-join rules a restore must respect; the backup reference a restore replaces; `internal/ops`'s `Seam`/catalog gap and sentinel duplication; backup's rclone-launch/copy-classification duplication (incl. `cmd/dubplate-server/backup.go`'s second remote-name validator).
- To the architecture cleanup pass: the 5c-ii reads (`internal/auth`, `internal/serve`, `internal/identity`), 6b-i's ops/local findings, 6a's cmd findings. To 6d: a store identity minted at create, the migration-wait key, serve's caller. To 6c: the `SAVEPOINT` check on `COMMIT;BEGIN;` (risk R13). L5/L6 (roster-POST throttle, n5 whole-roster doctor row): no pass assigned.
- 99 declared coverage gaps, none blocking. Staged legs warn: 326 process citations, 128 exported-unused. Fork B: 0.8/0.9 FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
