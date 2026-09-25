# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, 5b-i, 5b-ii, and 5c-i are closed. **Rung 6's pass 6a (the store
foundation) is closed and merged** to master at `8736528` (lane 2, `08befea`..`8736528`, all four
rulings yes); spend about 4.0M against the 6.8M ceiling (5.44M flag), zero attended interactions
after approval (HISTORY, top entry). Both touched packages graded workmanlike at the merge's
architecture reads; store findings route to 6b-i, cmd findings to the architecture cleanup pass.
Beets 2.14.1 is adopted and closed (`367564c`, `422efcd`, simplifier clean). 5c-ii runs in lane 1.
The visual identity spec (`specs/2026-09-24-visual-identity-design.md`, `ecba52a`, revision 1)
authors the web look: palettes, type, the record mark, the component layer, and the photo system.
It awaits Geoff's review; 5c-ii Task 4 and 5c-iii build to it. Owed at its ratification: a web
surface standard section 19 entry recording that its ruling 4 withdraws W1.3's illustrations.

## Rulings taken (Geoff, 2026-09-24)

- **The 6a plan is approved and closed**, all four rulings yes. Its eleven errata E1-E11 are applied.
- **5c-i's flag crossing is accepted**; the stricter shared-namespace rule is declined.
- **Follow beets on the "ft" fix** (the oracle pin moved to 2.14, `6d4dd05`); **the layout carries
  disc numbers**, a `1-01` prefix on multi-disc albums only.

## Next action

1. **6b-i is approved** (Geoff, 2026-09-24: all four rulings yes, no split; revision 2.2,
   `plans/2026-09-24-rung-6b-i-ops-seam-backup.md`; ceiling 7.4M, flag 5.92M). Lane 2 opens on it
   in worktree `.claude/worktrees/lane-2-6b-i`; Task 0 first.
2. **5c-ii runs in lane 1** (revision 2.4, all seven rulings yes; 9.4M ceiling, 7.52M flag; visual
   identity tokens amended in: `label`/`label-night` palettes, `dubplate-system` stamp, the
   `__Host-dubplate_theme` cookie). Segment A closed green (`5637ecb`, `e9f7367`, `7769d25`); Segment
   B (Tasks 3-5) is in flight.
3. **Matcher word boundary: ruled yes** (Geoff, 2026-09-24): a documented Go translation of beets'
   featuring row, a parity golden, and a corpus case, as a small matcher pass after the current lanes
   (`research/2026-09-24-matcher-unicode-word-boundary.md`).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | OPEN on 5c-ii since 2026-09-24; Segment A green, Segment B in flight |
| 2 | worktree `.claude/worktrees/lane-2-6b-i`, branch `lane-2-6b-i` | OPEN on 6b-i since 2026-09-24; Task 0 first |

## Open items (carried)

- No deploy until 5d: the env template selects `magic-link` without its nine keys.
- To 5d: the "`doctor` exits 0" row, `ReadWritePaths`, the nine keys, the trusted proxy, the runbook
  rows (`member add` without secrets in argv, kind mode's enumeration bound, the mail-ceiling lever,
  plus 6a's owed lines: root-run refusal, `runuser` fix, owner-mismatch create, FUSE warning, btrfs
  holder gap).
- To 6b-i: the seven handoff items (surface, intent-record rule, staging rule, checkpoint busy row,
  VACUUM INTO's WAL set, discarded-connection rule, runbook lines) plus the store architecture
  findings above. To 6b-ii: the intent record's format. To 6d: the migration-wait key, serve's caller,
  the unbounded settle loop, `serveOptions.logger`'s missing nil guard.
- To rung 12: no principal maps a non-admin onto a streamer admin. To rung 7: the path validator
  (R-6a9), archive uploads. 94 declared coverage gaps (86 carried plus 6a's 8), none blocking. Staged
  legs warn: 325 process citations, 128 exported-unused (plus the architecture reads' findings, not
  yet counted). Fork B: 1.0 is FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`, unidecode
  `1.4.0`, lap `0.5.13`; SQLite driver `modernc.org/sqlite v1.59.0` (6a). Spend: about 61.6M plus
  9.1M, 6.4M, about 8.3M, and about 4.0M (6a) across the windows above.
