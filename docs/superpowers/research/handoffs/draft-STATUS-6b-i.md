<!--
Draft STATUS.md, modeled on the main tree's current docs/STATUS.md
(`git -C /var/home/glw907/Projects/dubplate show master:docs/STATUS.md`),
as it would read after 6b-i merges and lane 2 closes. The conductor
fills the bracketed fields (merge SHA, actual spend, architecture-read
routing, staged-leg counts) during the fold. Line count below (target
60, "## Current state" through EOF) excludes this comment block.
-->

# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a to 5c-i, and rung 6's **6a** (store
foundation, `8736528`, about 4.0M) and **6b-i** (ops seam and backup,
`<merge-sha>`, about `<actual>`M against 7.4M; HISTORY, top entry) are
closed. 6a's findings routed to 6b-i and the cmd cleanup pass; 6b-i's
five reads route `<findings, filled at the fold>`. Beets 2.14.1 closed;
5c-ii runs in lane 1. Visual identity is complete (Geoff, 2026-09-25,
spec `782d864`); owed: a web surface standard entry for ruling 4.

## Rulings taken (Geoff, 2026-09-24)

- **6b-i's four rulings, all yes, no split**: 7.4M ceiling; `doctor`
  probes rclone's own backend features; the `backup` verb's five exit
  codes (0, 1, 69, 75, 130); Task 1A exports serve mode's open early.
- **Beets' "ft" fix followed; the layout carries disc numbers**
  (`1-01` prefix, multi-disc only).

## Next action

1. **6b-ii plans next**: the network actor and `internal/identity`
   import, `forbidden`/`internal`, the catalog operation's first
   production caller (HISTORY's handoff). Re-derive its ceiling first.
2. **5c-ii at its flag** (lane 1; 9.4M ceiling, 7.52M flag, spend about
   7.4M). Tasks 1-7, simplifier, three reviews in. Owed: one fold task,
   Task 8's close, the reads. Carried: security M2 to rung 12; a11y 4-6
   to the identity spec.
3. **Matcher word boundary ruled yes**: a small pass after the lanes.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | OPEN on 5c-ii since 2026-09-24 |
| 2 | worktree `.claude/worktrees/lane-2-6b-i` | CLOSED; merged at `<merge-sha>` |

## Open items (carried)

- No deploy until 5d (env template selects `magic-link` without its
  nine keys): `doctor` exits 0 row, `ReadWritePaths`, trusted proxy,
  runbook rows (6a's and 6b-i's: `KillMode=mixed`, rclone v1.74.3+,
  crypt remote, exit-7 remedy, hand-remove a backup ref).
- To 6b-ii: ops interface/codes, `forbidden`/`internal`, the network
  actor, catalog's first caller, lease/staging/remote-join rules, the
  reference a restore replaces.
- To 6d: a store identity minted at create. To 6c: `SAVEPOINT` on
  `COMMIT;BEGIN;` (risk R13). To rung 7: per-collection/extension guard
  counting, a sized `--max-delete`, a remote-pruning regression floor,
  the path validator (R-6a9), archive uploads.
- To rung 9: library filter, facets, trigram search. To rung 12: no
  principal maps a non-admin onto a streamer admin.
- 99 declared coverage gaps (94 carried plus 6b-i's 5, plus 5
  review-noted limits), none blocking. Staged legs warn: `<counts,
  filled at the fold>`. Fork B: 1.0 FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`,
  mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`,
  `modernc.org/sqlite v1.59.0`, rclone v1.74.3 (measured, not on the
  box yet). Spend to date: HISTORY, per pass.
