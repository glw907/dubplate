# Draft STATUS body (fold into docs/STATUS.md at the merge; draft only)

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, **5b-i**, and **5b-ii** (roster,
mail seam, rate limiters, origin check, auth handlers) are closed;
5b-ii merges at the tip this close commits. R4's exclusion denominator
is still UNREAD, owed before rung 15.

## Next action

1. **5b-ii closed at about 6.4M of its 7.5M ceiling**, zero planning
   misses, zero execution sittings beyond five launch rulings. All
   twelve tasks accepted; Tasks 10-11 first time. The pass-end
   `web-auth-security-reviewer` found no blocker, five majors, fourteen
   minors, plus a re-check of all twenty 5b-i findings (eighteen fixed
   or narrowed, two carried to 5d). **Geoff accepted the conductor's
   recommendation 2026-09-23**: majors 1-4 and the relevant minors
   become 5c's security-fold Task 1; major 5 is a 5d deploy gate. **The
   granular per-minor split is the conductor's own ruling of
   2026-09-23 under that acceptance**, not a separate owner decision;
   full breakdown in `research/handoffs/draft-HISTORY-5b-ii.md`.
2. **5c is next**: pages, the `member` verb, the web gate leg, against
   the ratified web surface standard and configuration homes revision
   3, plan authored just in time; Task 1 folds the routed security
   majors before the add-member and link pages build on it. Rung 6
   runs as lane 2 beside 5c/5d once ruled (spec revision 2.2, ratified).
   Owed in `~/.dotfiles`: the infra round's task 9 and a
   `pass-execute.js` stop-on-escalate option.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED at this merge. Reopens on 5c once its plan is ruled. |
| 2 | worktree removed, branch deleted | CLOSED |
| doc | main tree | CLOSED. Rung 6 spec ratified. |

## Open items (carried)

- **No deploy until 5d**: `magic-link` lacks its nine keys, refusing every
  server verb; `newProductionRegistry` in `main.go` is the contested surface.
- **Carried to 5c**: rename `ErrRosterAdminChannelForbidden` on the CLI
  admin grant; the last-admin lockout guard; the session-refresh call
  site; the magic-link mail copy; the rejoin-sentence corrections
  (`CLAUDE.md:140`, `scripts/check.sh:5`, `Makefile:33`); the
  security-fold Task 1 above. **Carried to 5d**: the "`doctor` exits 0"
  row after the bootstrap verb; the `ReadWritePaths` grant; the env
  template's nine keys; the trusted-proxy setting.
- Configuration homes (rev. 3) routes tasks into rungs 6, 7, 9, 12, 14,
  16, and the 2.0 list. **Coverage-gap total moves seventy-three to
  seventy-nine** (six newly declared, detail in the HISTORY draft).
- **Unchanged from the live ledger, carried verbatim at the fold**:
  staged gate legs (313 process citations, 134 exported-unused, both
  warn-only); the leftovers list (four-knob wall registration through
  `DUBPLATE_SITE_NAME` re-mint); Pinned versions (Navidrome,
  cloudflared, beets, mediafile, unidecode, lap; 2026-08-30); Fork B
  FLAC-only for 1.0; the group non-FLAC carry; and the spend line
  (about 61.6M plus the window's 9.1M and this pass's about 6.4M,
  5b-i's overage still on R23).
