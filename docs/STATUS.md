# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a to 5c-i, and rung 6's **6a** (store foundation, `8736528`, about
4.0M) and **6b-i** (ops seam and backup, `6f2c3e0`, about 6.0M against 7.4M; HISTORY, top entry)
are closed. 6b-i's reads route to 6b-ii (`internal/ops`, backup's rclone-launch/copy-classification
duplication) and the architecture cleanup pass (`ops/local`'s seams/nits). Beets 2.14.1 closed;
5c-ii runs in lane 1. Visual identity is complete (Geoff, 2026-09-25, `782d864`); owed: a web
surface standard entry for ruling 4.

## Rulings taken (Geoff, 2026-09-24)

- **6b-i's four rulings, all yes, no split**: 7.4M ceiling; `doctor` probes rclone's own backend
  features; the `backup` verb's five exit codes (0, 1, 69, 75, 130); Task 1A exports serve mode's
  open early.
- Beets' "ft" fix followed; the layout carries disc numbers (`1-01` prefix, multi-disc only).

## Next action

**Resume prompt:** execute `docs/superpowers/plans/2026-09-25-unattended-window.md` (5c-ii and 6b-i
closed; draft and review three plans: 5c-iii, 6b-ii, the matcher pass; no new pass executes). Read
memory `session-2026-09-25-handoff`.

1. **5c-ii close is next** (lane 1; 9.4M ceiling, 7.52M flag, spend about 7.4M). Fold A
   (`d236d03`, `cc4649c`), F-A3 (`1306ae3`), and F-B (`3acbdc0`) are all accepted; the simplifier
   over the fold is in flight. Owed: Task 8's close and three architecture reads (`internal/auth`,
   `internal/serve`, `internal/identity`). Deferred to 5c-iii: label ink values `#9d2f26`/
   `#754f00`, unpinned until a11y W4's per-surface G9 row lands. Untestable: M1's
   `O_NOFOLLOW`/stat-race guards have no mutation test. Carried: security L5, L6.
2. **6b-ii plans next**, at fold 1: the network actor and `internal/identity` import,
   `forbidden`/`internal`, the catalog's first production caller, and the routed reads.
3. **The matcher word-boundary plan**, at fold 1, ruled yes 2026-09-24
   (`research/2026-09-24-matcher-unicode-word-boundary.md`).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | OPEN on 5c-ii since 2026-09-24 |
| 2 | worktree `.claude/worktrees/lane-2-6b-i` | CLOSED; merged at `6f2c3e0` |

## Open items (carried)

- No deploy until 5d: the env template selects `magic-link` without its nine keys. Owed the
  "`doctor` exits 0" row, `ReadWritePaths`, the trusted proxy, and the runbook rows (6a's plus
  6b-i's `KillMode=mixed`, rclone v1.74.3+, crypt remote, exit-7 remedy).
- To rung 7: per-collection/extension guard counting, a sized `--max-delete`, a remote-pruning
  regression floor, archive uploads. To rung 9: library filter, facets, trigram search. To rung
  12: no non-admin maps onto a streamer admin; a restore before then loses stars and playlists.
- To 6b-ii: the ops interface's four operations/sentinels; `forbidden`/`internal`/the network
  actor with their first caller; the lease, staging, and remote-join rules a restore must
  respect; the backup reference a restore replaces; `internal/ops`'s `Seam`/catalog gap and
  sentinel duplication; backup's rclone-launch/copy-classification duplication, including
  `cmd/dubplate-server/backup.go`'s second copy of the remote-name validator.
- To the architecture cleanup pass: `internal/ops/local`'s test-only seams and nits, plus 6a's
  cmd findings. To 6d: a store identity minted at create, the migration-wait key, serve's caller.
  To 6c: the `SAVEPOINT` check on `COMMIT;BEGIN;` (risk R13).
- 99 declared coverage gaps, none blocking. Staged legs warn: 326 process citations, 127
  exported-unused. Fork B: 1.0 FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`,
  unidecode `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
