# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 1b, rung 8 (8a, 8b), rung 4 (4a, 4b), and now **rung
5's pass 5a (the serve shell and ingress)** are closed
(`docs/HISTORY.md`'s "Rung 5, pass 5a" entry, merged at `307001b`).
`internal/serve` is born: the composition order, the ingress strip,
path normalization, prefix routing, the extension point, the deadline
policy, and the bind surface. A fresh-context
`web-auth-security-reviewer` read the whole shell; findings are at
`docs/superpowers/research/2026-09-08-rung-5a-security-review.md`. The
rung 5 spec stands at revision 2.1, errata now carrying 5a's nine
corrections. The 4b sitting deployed the shadow units to the box; the
adjudication stream is live, read through `dubplate-server shadow
report`, never the bare binary as root. Streamer 2.0 (ROADMAP
`streamer-2-0`) binds rungs 5-9 and 12. The box (5.78.148.42) is
production, runs Navidrome and the guardian and shadow timers; the
family system is not live.

## Next action

Lane-1 (main tree): IDLE. Pass 8b closed at `ce0ac2e`; pass 8c's plan
(absorb the shadow's reader into `tagio`) is next, just-in-time from
8b's handoff. Lane-2 (worktree `lane-2-5a`): IDLE, torn down at 5a's
merge. Rung 5 continues with **pass 5b (the auth core and the mail
seam)**, plan owed just-in-time from 5a's handoff (`internal/identity`,
`internal/auth`, `internal/notify` born; the roster store; the origin
check reading 5a's kept `Origin`/`Sec-Fetch-Site`; `keys.go`'s
remaining reserved-class retirement). No lane active; next session
picks one before dispatch.

## Open items (carried)

- Sixty-three declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four, rung 1b's two, rung 8a's seven, rung 8b's
  nine, rung 5a's eleven), none blocking. Fork B (FLAC-only 1.0, MP3
  after cutover) stays open, owed to Geoff; the family non-FLAC
  question is unresolved, carried.
- Four-knob wall registration unblocked; `AdminEmail`'s contact row,
  five 3b leftovers, `profile_hashes` (m1), the response-status
  extension (2b/2c), `Case.Extra`, `slices.Clone` (rung 9).
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

Spend to date: about 59.6M through pass 5a's close (55.7M plus 3.9M).
