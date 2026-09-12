# Draft STATUS.md replacement: after the 5b-i close

Draft for the conductor to fold into `docs/STATUS.md` at the merge.
This is a full proposed replacement for the file's body (based on
master's current text at the time of dispatch), not a diff; the
conductor reconciles it against whatever else lands on master first.
Kept to the house shape and under sixty lines.

---

# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 1b, rung 4 (4a, 4b), rung 8 in full, and now **rung 5's
pass 5b-i (the auth core)** are closed (`docs/HISTORY.md`'s "Rung 5,
pass 5b-i" entry). `internal/auth` and `internal/identity` are born:
the config core requires every owned-now wall row generically; the
token and the session cookie share one wire shape; the nonce store and
the revocation epoch are the two state records. The fresh-context
security review for 5b-i is owed before 5b-ii's merge (R-5b16). R4's
exclusion denominator is **UNREAD**, owed before rung 15 treats it as
final. Streamer 2.0 (ROADMAP `streamer-2-0`) binds rungs 5-9 and 12.
The box (5.78.148.42) is production, runs Navidrome and the guardian
and shadow timers; the group system is not live.

## Next action

Lane 2's worktree closes and is removed once master is green. **5b-ii**
(roster, limiters, notify, origin check, verify seam, handlers) is
next: re-derive its ceiling from 5b-i's actual 5.5M against the live
15.75M reservation, and its plan prep owes the fresh-context security
review covering both halves (R-5b16). Lane 1 is free.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED. Awaiting the conductor's next dispatch |
| 2 | `.claude/worktrees/lane-2-5b` | CLOSED at merge; worktree removed once master is green |

## Open items (carried)

- Configuration homes spec (revision 3, 2026-09-12) routes tasks into
  rungs 6, 7, 9, 12, 14, and 16 and the 2.0 projector conformance list
  (full list: the spec's "What this spec routes elsewhere").
- Seventy-three declared coverage gaps (rungs 3, 4a, 4b, 1b, 8a, 8b,
  5a, 8c, 5b-i), none blocking. Fork B is ruled (Geoff, 2026-09-08:
  FLAC-only 1.0, format ladder at 1.1, ROADMAP `formats-1-1`); the
  group non-FLAC question is unresolved, carried.
- Four-knob wall registration unblocked; `AdminEmail`'s contact row,
  five 3b leftovers, `profile_hashes` (m1), response-status extension
  (2b/2c), `Case.Extra`, `slices.Clone` (rung 9).
- The 2026-09-04 agentic-coding-evidence doc is corrupted, untracked;
  two albums await review in `/srv/music/review/2026-08-31/geoff/`;
  `FILEBROWSER_ADMIN_PASSWORD` orphaned; stale failed units;
  `DUBPLATE_SITE_NAME` retires when re-minted.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 61.6M through the 5b-i and 8c plan prep, plus
8c's own execution against its 2.1M ceiling, plus 5b-i's own
execution at roughly 5.5M against its 5.0M ceiling (overage recorded
against R23's reservation, not funded from another pass).
