# dubplate status

## Current state

The overnight window (`plans/2026-09-23-overnight-window.md`, 12M cap, ends 05:00 AKDT) is running.

- **5c-i segment A is closed.** All three tasks were accepted with every gate green:
  - Task 1 at `36f528d`.
  - Task 2 at `428665d`, `81cd77b`, and `49363bf`, after one fix round.
  - Task 3 at `74919a6`.
  - The spend was about 1.29M in subagents for three tasks, under the 0.533M basis, so the 6.5M ceiling
    stands.
  - Task 0 returned LAUNCH (`research/2026-09-23-rung-5c-i-task-0-report.md`). Minor 6's refusal
    (R-5c18) goes into the close's scope ledger.
- **Task 1's named mutation M9 has no proving test**, as the plan expects. The reviewer verified it,
  and the pass-end security read covers the stat-and-read order. M8 fires on the generic rows only.
- **The rung 6 errata are applied** (`b406f93`). Row 9's `identity.go` half waits for 5c-i to merge.
- **The 6a plan is at revision 2.2 and awaits Geoff's approval** (`plans/2026-09-23-rung-6a-store-foundation.md`, `e1fb273`).
  - The review ran three lenses, a fold, a verification read, a narrow second fold, and a prose fold.
    The record is `research/2026-09-23-rung-6a-plan-fold-record.md`.
  - Four rulings, each recommended yes: the 6.8M ceiling, `doctor` and `check` exiting 0 only for a
    store that is absent, older, or current, serve refusing a create beside a snapshot copy, and the
    busy timeout alone bounding the cross-process wait.
  - Eleven errata are owed to the rung 6 spec on approval.
- R4's exclusion denominator is still unread, and owed before rung 15.

## Next action

1. Segment B: Tasks 4 to 7. Then the simplifier, the mail-copy register gate, and the pass-end
   `web-auth-security-reviewer`.
2. Segment C: the close.
3. In the morning, Geoff approves 6a and rules its four rulings. Lane 2 then opens on 6a in a worktree.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | OPEN on 5c-i: segment A closed, segment B next |
| 2 | worktree removed, branch deleted | CLOSED. The 6a plan (rev 2.2) awaits approval |

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
