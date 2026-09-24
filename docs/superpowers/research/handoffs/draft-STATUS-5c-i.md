# Draft STATUS body (fold into docs/STATUS.md at the merge; draft only)

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, 5b-i, 5b-ii, and **5c-i** (the
security fold, the CLI admin grant, the `member` verb) are closed;
5c-i merges at the tip this close commits. R4's exclusion denominator
is still unread, owed before rung 15.

## Next action

1. **5c-i closed at about 5.3M of its 6.5M ceiling**, zero planning
   misses, zero execution sittings. All seven tasks accepted, most
   after one fix round. The security review found no critical or
   major finding, confirmed all four 5b-ii majors fixed, and raised
   five new minors and five nits, folded across two tasks; one nit
   (n5) refused, noted for 5c-ii. Detail: `research/handoffs/draft-HISTORY-5c-i.md`.
2. **Geoff approves the rung 6a plan (rev 2.2)**, rules its four
   rulings (the 6.8M ceiling; exit criteria for `doctor`/`check`;
   serve's create-beside-snapshot refusal; the busy-timeout bound).
   Eleven errata owed on approval. `plans/2026-09-23-rung-6a-store-foundation.md`.
3. **5c-ii's plan is next**: the admin gate, roster API, serve wiring,
   and the `web/` tree, against a tree where 5c-i's security fold has
   merged and been reviewed. Rung 6 (6a first) opens lane 2 once ruled,
   beside 5c-ii. **Morning question**: plan 5c-ii before or after 6a.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED at this merge. Reopens on 5c-ii once ruled. |
| 2 | worktree removed, branch deleted | CLOSED. 6a plan (rev 2.2) awaits approval. |

## Open items (carried)

- No deploy until 5d: `magic-link` lacks its nine keys, refusing every
  server verb; `newProductionRegistry` (`main.go`) is contested.
- Carried to 5c-ii: the admin gate (R-5c12), the roster API (R-5c13),
  serve wiring, the `web/` tree, rejoin-sentence corrections
  (`CLAUDE.md:140`, `scripts/check.sh:5`, `Makefile:33`). Carried to
  rung 12: a principal must not map a non-admin member onto a streamer
  admin account; Navidrome v0.63.2 matches usernames case-insensitively
  (evidence recorded, not the value).
- Carried to 5d: the "`doctor` exits 0" row after the bootstrap verb;
  `ReadWritePaths`; the env template's nine keys; the trusted-proxy
  setting; the mail-ceiling lockout-lever runbook sentence.
- Seven declared coverage gaps this pass (stderr-only CLI audit; a
  hand edit leaves no audit record; a lockless hand edit inside a
  write's window is overwritten; the per-address map's untracked
  admission past 10,000 members; no promote path; the two rung-12 items
  above): detail in the HISTORY draft.
- Staged gate legs (warn today): 325 process citations, 128
  exported-unused symbols (both flip to `fail` at comment-cleanup).
  Configuration homes routes tasks into rungs 6, 7, 9, 12, 14, 16 and
  the 2.0 list; seventy-nine coverage gaps carried from 5b-ii plus the
  seven above. Fork B FLAC-only for 1.0; group non-FLAC carried.
- Pinned (2026-08-30): Navidrome `0.63.2`, cloudflared `2026.8.2`,
  beets `2.13.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`.
  Spend: about 61.6M plus the 2026-09-12 window's 9.1M, 5b-ii's about
  6.4M, and 5c-i's about 5.3M; 5b-i's overage still sits on R23.
