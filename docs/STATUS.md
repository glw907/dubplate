# dubplate status

## Current state

Rungs 1b, 3, 4, and 8, rung 5's 5a, and now **5b-i** (the auth core:
`internal/auth`, `internal/identity`) are closed; 5b-i merged at
`1d4ed29`. R4's exclusion denominator is **UNREAD** (box unreachable
03:14 AKDT), owed before rung 15 treats R4 as final. The overnight
window (`plans/2026-09-12-overnight-window.md`) closed both its passes,
8c and 5b-i, under the 10M cap at about 9.2M.

## Next action

1. **The rung 5 security read for 5b-i** (R-5b16): the fresh-context
   `web-auth-security-reviewer` cut from the window, owed before 5b-ii.
2. **5b-ii's plan**, just in time after it (roster, limiters, notify,
   origin check, verify seam, handlers), ceiling re-derived from 5b-i's
   5.5M against R23's live 15.75M. Then **5c**, and the owed box read.

Awaiting Geoff's read: `specs/2026-09-12-web-surface-standard.md`
(revision 2.1, `49a03b4`, rulings first in section 16) and
`specs/2026-09-11-configuration-homes-design.md` (revision 3,
`07f4e01`). Grant reading, flagged (overnight plan, G1): the twelve
conductor-owed folds were ruled by recommendation on "then start"
(Geoff, 2026-09-12 00:44); the ratify-or-reverse table is there.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED. Awaiting the conductor's next dispatch |
| 2 | `.claude/worktrees/lane-2-5b` | CLOSED. 5b-i merged at `1d4ed29` (runs `wf_61f3cf53-665`, `wf_769302f4-cef`, `wf_be7f90c5-281`); worktree to be removed by the conductor |

## Open items (carried)

- The configuration homes spec (revision 3) routes tasks into rungs 6,
  7, 9, 12, 14, 16 and the 2.0 list; the rung 5 spec's eighteen errata
  land with 5b-ii's plan prep (HISTORY). Seventy-three declared
  coverage gaps (rungs 3, 4a, 4b, 1b, 8a, 8b, 5a, 8c, 5b-i's six), none
  blocking. Fork B ruled FLAC-only for 1.0 (ROADMAP `formats-1-1`); the
  group non-FLAC question is carried.
- Four-knob wall registration unblocked; `AdminEmail`'s contact row,
  five 3b leftovers, `profile_hashes` (m1), response-status extension
  (2b/2c), `Case.Extra`, `slices.Clone` (rung 9); corrupted untracked
  2026-09-04 agentic-coding-evidence doc; two albums awaiting review in
  `/srv/music/review/2026-08-31/geoff/`; `FILEBROWSER_ADMIN_PASSWORD`
  orphaned; stale failed units; `DUBPLATE_SITE_NAME` re-mint.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend: about 61.6M plus the window's 9.2M; 5b-i's overage sits on R23.
