<!--
Draft STATUS.md, modeled on the main tree's current docs/STATUS.md
(`git -C /var/home/glw907/Projects/dubplate show master:docs/STATUS.md`), as it would read after
6c merges and lane 2 closes. The conductor fills the bracketed fields (merge SHA, actual spend,
architecture-read routing) during the fold. Line count below ("## Current state" through EOF)
excludes this comment block: 33 lines, under the 60-line cap.
-->

# dubplate status

## Current state

Rungs 1b, 3, 4, 8, 5a-5c-iii, and all of rung 6 closed, **6c** (the routed schema and settings
home, `<merge-sha>`, about 4.5M plus the merge and reads against 5.4M) joining 5c-iii and
6b-ii-b as HISTORY's top entries. 6c's six architecture reads route `<findings, filled at the
fold>`. 5c-iv (Lane 1) continues.

## Next action

- **5d-i is now clear to launch**: its own rule (5d plan `593d7fd:10`, revision 2.1) held it until
  6c's merge; 6c has merged.
- **Lane 1, 5c-iv** continues (Tasks 8-11 in flight or closing; see its own HISTORY entries for
  detail once folded).
- **The transliteration pass** (`plans/2026-09-25-translit-table-at-setup.md`, approved, 3.3M)
  awaits a lane, now that 6c has freed lane 2.

## Owner questions

- Navidrome 0.64 bump for GHSA-82gh: open (sharing stays at zero until 0.8).
- TypeScript 6.0.3 to 7.0.2, a major (5c-iv Task 0's pin check).
- "Streamer username": a term admins know before 0.8?
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- 5d (rev 2.2) awaits two morning rulings: 1, ceilings 5d-i 7.0M / 5d-ii 3.5M; 2, the sitting.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master), this session | OPEN: 5c-iv |
| 2 | worktree `lane-2-6c` (branch `lane-2-6c`) | CLOSED; merged at `<merge-sha>` |

## Open items (carried)

- No deploy until 5d: the nine `magic-link` env keys, `doctor` exits 0, `ReadWritePaths`, the
  trusted proxy, the theme keys' `doctor` row, `deploy.sh`'s clean-`web/` build, the runbook rows.
  6c adds: `deploy.sh` halts on `check`'s nonzero exit before restarting `serve` (Ruling 2).
- To 6d: exporting the seed apply into `serve`, the local-keep `OpenServe` prune proof, the real
  nonce-retention/link-expiry correlated pair, the auth keys as sensitive settings, a "restart"-mode
  key's once-at-start read, the seed-vs-restored-admin-row interaction. To 5d and 6d: a window
  where `doctor` reports a seed that `serve` does not yet apply.
- To rung 14: the admin write path and a resolving channel for an `admin` row that refuses
  `serve`'s start.
- To rung 7: `member_quota` reservations, the default-collection preference, the remaining
  `member_preferences` columns.
- 99 coverage gaps carried plus 6c's 4, none blocking. Staged legs: 325 process citations, 128
  exported-unused.
