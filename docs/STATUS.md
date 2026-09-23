# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a and **5b-i** (the auth core) are closed; 5b-i merged at `1d4ed29`. R4's
exclusion denominator is UNREAD (box unreachable at 03:14 AKDT), owed before rung 15. The overnight window
(`plans/2026-09-12-overnight-window.md`) closed 8c and 5b-i at 9.1M of 10M.

## Next action

1. **5b-ii is executing** from `plans/2026-09-12-rung-5b-ii-auth-handlers.md` (revision 2.1, 7.5M ceiling,
   flag 6.0M), segment A (Tasks 1-3) first. Launch rulings sit in the plan header (2026-09-23): upshifts
   dispatch as `claude-opus-5-5`, `resolveAuthConfig` returns a struct, the seven folds ratified. Task 0's
   evidence at `f75b0d9` holds because the Go tree is unchanged. Then segments B, C, D1, D2, then **5c**.
2. The Claude infra round's task 9 (the `pass-execute.js` edit) is still owed in `~/.dotfiles/docs/STATUS.md`.

Awaiting Geoff's read: `specs/2026-09-12-web-surface-standard.md` (revision 2.1, `49a03b4`, rulings first in
section 16) and `specs/2026-09-11-configuration-homes-design.md` (revision 3, `07f4e01`).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | ACTIVE. 5b-ii segment A (Tasks 1-3) launched 2026-09-23 from an Opus 5.5 conductor session |
| 2 | worktree removed, branch deleted | CLOSED. 5b-i merged at `1d4ed29` |

## Open items (carried)

- **No deploy until 5d**: the env template selects `magic-link` without its nine required keys, so a render
  would refuse every server verb, shadow included (merged-range review, 2026-09-12). `newProductionRegistry`
  in `main.go` is the one registration line, a contested surface for the next two-lane run.
- The configuration homes spec (revision 3) routes tasks into rungs 6, 7, 9, 12, 14, 16 and the 2.0 list;
  the rung 5 spec's eighteen errata land with 5b-ii's plan prep (HISTORY). Seventy-three declared coverage
  gaps (rungs 3, 4a, 4b, 1b, 8a, 8b, 5a, 8c, 5b-i's six), none blocking. Fork B ruled FLAC-only for 1.0
  (ROADMAP `formats-1-1`); the group non-FLAC question is carried.
- Staged gate legs (warn today): 313 process citations in shipped Go comments (a comment-cleanup task flips
  `CHECK_CITATIONS=fail`); 134 exported-unused symbols (before `CHECK_EXPORTED_UNUSED=fail`; detail in
  HISTORY's 5b-i process-changes note).
- Leftovers: four-knob wall registration; `AdminEmail`'s contact row; five 3b items; `profile_hashes`;
  response-status extension; `Case.Extra`; `slices.Clone` (rung 9); the corrupted untracked 2026-09-04
  evidence doc; two review albums on the box; `FILEBROWSER_ADMIN_PASSWORD`; stale failed units;
  `DUBPLATE_SITE_NAME` re-mint.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend: about 61.6M plus the window's 9.1M; 5b-i's overage sits on R23.

