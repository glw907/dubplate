# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a and **5b-i** (the auth core,
`internal/auth` and `internal/identity`) are closed; 5b-i merged at
`1d4ed29`. R4's exclusion denominator is UNREAD (box unreachable at
03:14 AKDT), owed before rung 15. The overnight window
(`plans/2026-09-12-overnight-window.md`) closed 8c and 5b-i at 9.1M of 10M.

## Next action

1. **Rule 5b-ii's seven conductor-owed folds** (plan header,
   `plans/2026-09-12-rung-5b-ii-auth-handlers.md`, revision 2 at
   `f36b10c`; the ceiling question first), run its Task 0 at `opus`,
   then launch it in the main tree in four segments. The 5b-i security
   read landed (`research/2026-09-12-rung-5b-i-security-review.md`: one
   blocker, five majors; Tasks 1 and 2 fold them). Then **5c**, and the
   owed box read.

Awaiting Geoff's read: `specs/2026-09-12-web-surface-standard.md`
(revision 2.1, `49a03b4`, rulings first in section 16) and
`specs/2026-09-11-configuration-homes-design.md` (revision 3,
`07f4e01`). Grant reading, flagged: the twelve conductor-owed folds were
ruled by recommendation on "then start" (2026-09-12 00:44); the
ratify-or-reverse table is in the overnight plan.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED. Awaiting the conductor's next dispatch |
| 2 | worktree removed, branch deleted | CLOSED. 5b-i merged at `1d4ed29` (runs `wf_61f3cf53-665`, `wf_769302f4-cef`, `wf_be7f90c5-281`); worktree to be removed by the conductor |

## Open items (carried)

- **No deploy until 5d**: the env template selects `magic-link` without
  its nine required keys, so a render would refuse every server verb,
  shadow included (merged-range review, 2026-09-12). `newProductionRegistry`
  in `main.go` is the one registration line, a contested surface for the
  next two-lane run.
- The configuration homes spec (revision 3) routes tasks into rungs 6,
  7, 9, 12, 14, 16 and the 2.0 list; the rung 5 spec's eighteen errata
  land with 5b-ii's plan prep (HISTORY). Seventy-three declared
  coverage gaps (rungs 3, 4a, 4b, 1b, 8a, 8b, 5a, 8c, 5b-i's six), none
  blocking. Fork B ruled FLAC-only for 1.0 (ROADMAP `formats-1-1`); the
  group non-FLAC question is carried.
- Leftovers: four-knob wall registration; `AdminEmail`'s contact row;
  five 3b items; `profile_hashes`; response-status extension; `Case.Extra`;
  `slices.Clone` (rung 9); the corrupted untracked 2026-09-04 evidence
  doc; two review albums on the box; `FILEBROWSER_ADMIN_PASSWORD`;
  stale failed units; `DUBPLATE_SITE_NAME` re-mint.

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
