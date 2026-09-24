# Draft STATUS body (fold into docs/STATUS.md at the merge; draft only)

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, **5b-i**, and **5b-ii** (the roster,
mail seam, rate limiters, origin check, and auth handlers) are closed;
5b-ii merges from the main tree at the tip this close commits. R4's
exclusion denominator is still UNREAD, owed before rung 15.

## Next action

1. **5b-ii closed at about 6.4M of its 7.5M ceiling**, zero planning
   misses, zero execution sittings beyond five launch rulings (ruled
   together before Task 1, not counted as a sitting). All twelve tasks
   accepted; Task 10 and Task 11 each accepted first time. The pass-end
   `web-auth-security-reviewer`
   (`research/2026-09-23-rung-5b-ii-security-review.md`) found no
   blocker, five majors, and fourteen minors against 5b-ii's own new
   surface, plus a re-check of all twenty 5b-i findings (eighteen fixed
   or fixed-with-narrower-scope, two carried to 5d). **Geoff ruled
   2026-09-23**: the five majors and nine of the fourteen minors become
   5c's security-fold Task 1; one minor (HSTS) joins 5c's header set;
   one major with two minors are 5d deploy gates; one minor is a
   recorded residual. Full disposition:
   `research/handoffs/draft-HISTORY-5b-ii.md`.
2. **5c is next**: the pages, the `member` verb, and the web gate leg,
   against the ratified web surface standard and configuration homes
   revision 3. 5c's plan is authored just in time, after this merge.
   5c's Task 1 folds the routed security majors before the add-member
   page and the link pages build on top of them.
3. **Rung 6 runs as lane 2 beside 5c/5d once ruled**; its spec
   (revision 2.2, `docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`)
   is ratified.
4. Owed in `~/.dotfiles` (held: another session had uncommitted edits
   there on 2026-09-23): the infra round's task 9, and a
   `pass-execute.js` stop-on-escalate option.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED at this merge. Reopens on 5c once its plan is ruled. |
| 2 | worktree removed, branch deleted | CLOSED |
| doc | main tree | CLOSED. Rung 6 spec ratified. |

## Open items (carried)

- **No deploy until 5d**: the env template selects `magic-link` without
  its nine required keys, so a render would refuse every server verb.
- **Carried to 5c**: rename `ErrRosterAdminChannelForbidden` when the
  CLI admin grant lands; the last-admin lockout guard; the
  session-refresh call site; the magic-link mail copy; the three
  rejoin-sentence corrections (`CLAUDE.md`, `scripts/check.sh`,
  `Makefile`); the security-fold Task 1 above.
- **Carried to 5d**: the "`doctor` exits 0" acceptance row runs after
  the bootstrap verb; the serve unit's `ReadWritePaths` grant; the env
  template's nine missing keys; the trusted-proxy setting.
- The configuration homes spec (revision 3) routes tasks into rungs 6,
  7, 9, 12, 14, 16 and the 2.0 list.
