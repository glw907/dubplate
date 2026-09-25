# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a to 5c-i, and **rung 6's pass 6a** (the store foundation, merged at
`8736528`, about 4.0M against 6.8M; HISTORY, top entry) are closed. 6a's architecture findings route
to 6b-i (store) and the architecture cleanup pass (cmd).
Beets 2.14.1 is adopted and closed (`367564c`, `422efcd`, simplifier clean). 5c-ii runs in lane 1.
The visual identity spec (`specs/2026-09-24-visual-identity-design.md`, `ecba52a`, revision 1)
authors the web look: palettes, type, the record mark, the component layer, and the photo system.
It awaits Geoff's review; 5c-ii Task 4 and 5c-iii build to it. Owed at its ratification: a web
surface standard section 19 entry recording that its ruling 4 withdraws W1.3's illustrations.

## Rulings taken (Geoff, 2026-09-24)

- **Follow beets on the "ft" fix** (the oracle pin moved to 2.14, `6d4dd05`); **the layout carries
  disc numbers**, a `1-01` prefix on multi-disc albums only.

## Next action

1. **6b-i runs in lane 2** (approved 2026-09-24, all four rulings yes, no split; ceiling 7.4M).
   Tasks 1S, 1A, 1B accepted (`4079422`..`089afe2`; 1B's escalation was the gate-string rule only,
   reproduced green by its reviewer, accepted by the conductor). A tidy task 1T (lease-kind and
   reference-remote validation) and Task 2 close segment A; spend about 1.47M so far.
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
| 2 | worktree `.claude/worktrees/lane-2-6b-i`, branch `lane-2-6b-i` | OPEN on 6b-i since 2026-09-24; segment A (1T, 2) in flight |

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
- To rung 9: the library filter consumes the typed album columns; facet counts plus a page under
  100 ms p95 at 10,000 albums as indexed SQL, search on the search3 trigram index with a prefix path
  under 3 characters (`research/2026-09-24-library-filtering-prior-art.md`, `f84ceea`). Gaps with
  no owner yet: genre (rung 11), bit depth and sample rate, saved views' visibility (and 6c's
  member-preferences table). To rung 12: no principal maps a non-admin onto a streamer admin. To rung 7: the path validator
  (R-6a9), archive uploads. 94 declared coverage gaps (86 carried plus 6a's 8), none blocking. Staged
  legs warn: 325 process citations, 128 exported-unused (plus the architecture reads' findings, not
  yet counted). Fork B: 1.0 is FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`, unidecode
  `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`. Spend to date: HISTORY, per pass.
