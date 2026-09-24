# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, 5b-i, and **5b-ii** (roster, mail seam, limiters, origin check,
auth handlers) are closed; 5b-ii closed on master at the ledger fold, about 6.4M of 7.5M, zero misses,
zero sittings (HISTORY). R4's exclusion denominator is still UNREAD, owed before rung 15.

## Next action

1. **5c's plan awaits Geoff's approval**: `plans/2026-09-23-rung-5c-pages-member-verb.md`, split into
   5c-i (in full) and 5c-ii (outlined); three lens reviews at `2a416b5`, fold in review. 5c-i's Task 1
   folds 5b-ii's security majors 1-4 and the relevant minors; major 5 (trusted proxy) is a 5d deploy
   gate (Geoff, 2026-09-23; `research/2026-09-23-rung-5b-ii-security-review.md`, HISTORY). An
   overnight window is planned for 5c-i's execution plus the rung 6 errata and a 6a plan draft.
2. **Rung 6's spec is ratified** (Geoff, 2026-09-23): `specs/2026-09-23-rung-6-ops-seam-store.md`
   revision 2.2, all fourteen rulings yes, five passes, about 28M; errata owed. The web surface
   standard (`94d8af3`) and configuration homes revision 3 (layout parity conditional on a rung 12
   Navidrome check) are ratified too. Rung 6 runs as lane 2 beside 5c/5d; 5 through 5d stays first.
3. **Landed in `~/.dotfiles`**: `spec-plan-review` (`8c53329`); `stopOnEscalate`, infra task 9 (`ce70b5f`).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED. 5b-ii closed on master at the ledger fold; reopens on 5c-i once its plan is approved |
| 2 | worktree removed, branch deleted | CLOSED. Reopens on rung 6 once its 6a plan is approved |
| doc | main tree | CLOSED. Rung 6 spec revision 2.2 ratified 2026-09-23 |

## Open items (carried)

- No deploy until 5d: the env template selects `magic-link` without its nine required keys, so a
  render would refuse every server verb; `newProductionRegistry` in `main.go` is a contested surface.
- Carried to 5c: rename `ErrRosterAdminChannelForbidden` with the CLI admin grant; the atomic
  last-admin lockout guard; the session-refresh call site; the magic-link mail copy; the rejoin
  sentences (`CLAUDE.md:140`, `scripts/check.sh:5`, `Makefile:33`). Carried to 5d: the "`doctor`
  exits 0" row (the 5d exit gate) after the bootstrap verb; `ReadWritePaths`; the template's nine
  keys; the trusted-proxy setting with minors 9 and 14.
- Configuration homes routes tasks into rungs 6, 7, 9, 12, 14, 16 and the 2.0 list. Seventy-nine
  declared coverage gaps (six new at 5b-ii), none blocking. Fork B ruled FLAC-only for 1.0 (ROADMAP
  `formats-1-1`); the group non-FLAC question is carried.
- Staged gate legs (warn today): 313 process citations (a comment-cleanup task flips
  `CHECK_CITATIONS=fail`); 134 exported-unused symbols (HISTORY's 5b-i process-changes note).
- Leftovers: four-knob wall registration; `AdminEmail`'s contact row; five 3b items; `profile_hashes`;
  response-status extension; `Case.Extra`; `slices.Clone` (rung 9); the corrupted untracked
  2026-09-04 evidence doc; two review albums on the box; `FILEBROWSER_ADMIN_PASSWORD`; stale failed
  units; `DUBPLATE_SITE_NAME` re-mint.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend: about 61.6M plus the 2026-09-12 window's 9.1M and 5b-ii's about 6.4M; 5b-i's overage sits on R23.
