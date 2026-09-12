# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 1b, rung 4 (4a, 4b), rung 5's pass 5a, and now **rung 8
in full (8a, 8b, 8c)** are closed. `internal/engine/tagio` is the sole
holder of FLAC container knowledge in the repo:
`internal/engine/shadow` holds none, and its read path routes through
`tagio.Parse` plus the newly exported `tagio.Tags`, one parse per file
(`docs/HISTORY.md`'s "Rung 8, pass 8c" entry). `internal/serve` is
born from 5a, and the rung 5 spec stands at revision 2.1. R4's
exclusion denominator is **UNREAD**: the box read failed at 03:14 AKDT
and the read is owed to the next attended session, before rung 15
treats R4 as final; the bound stands at zero known exclusions against
8c's eight-row exclusion set. Streamer 2.0 (ROADMAP `streamer-2-0`)
binds rungs 5-9 and 12. The box (5.78.148.42) is production, runs
Navidrome and the guardian and shadow timers; the group system is not
live.

## Next action

Lane 2 continues 5b-i (`plans/2026-09-08-rung-5b-auth-core.md`, Task 0
plus ten, 5.0M ceiling, 4.0M flag; **5b-ii** follows with roster,
limiters, notify, origin check, verify seam, handlers, provisional
4.5M). Lane 1 is free: rung 5b-ii's plan, or a fresh assignment, per
the conductor's next dispatch. The window plan is
`docs/superpowers/plans/2026-09-12-overnight-window.md`.

Awaiting Geoff's read:

- The web surface standard,
  `docs/superpowers/specs/2026-09-12-web-surface-standard.md` (revision
  2.1, commit `49a03b4`); its rulings-first list is in section 16.
- The configuration homes spec,
  `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`
  (revision 3, `07f4e01`), ratified in conversation; its routed tasks
  are in the carried list below.

Grant reading, flagged (overnight plan, G1): the twelve conductor-owed
folds were ruled by recommendation on "then start" (Geoff, 2026-09-12
00:44); the ratify-or-reverse table is in the overnight plan.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED. 8c landed on master (runs `wf_c808ec5e-1ba` then `wf_5d503f3a-ebd`); rung 8 closed in full. Awaiting the conductor's next dispatch |
| 2 | `.claude/worktrees/lane-2-5b` (branch `lane-2-5b`, created from `7ba3cc1`) | RUNNING 5b-i: segments 2A (`wf_61f3cf53-665`, Tasks 1-3, closed `eddeadd`+`52b6a20`) and 2B (`wf_769302f4-cef`, Tasks 4-7 all accepted, closed `ad6cb0c`) done; simplifier round then segment 2C (Tasks 8-10), then the merge ritual |

## Open items (carried)

- Configuration homes spec (revision 3, 2026-09-12) routes tasks the
  just-in-time plans must carry, into rungs 6, 7, 9, 12, 14, and 16
  and the 2.0 projector conformance list. Full list: the spec's "What
  this spec routes elsewhere". Build spec, organizer spec, 2.0 spec,
  ROADMAP, and the methodology study carry dated errata.
- Sixty-seven declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four, rung 1b's two, rung 8a's seven, rung 8b's
  nine, rung 5a's eleven, rung 8c's four), none blocking. 8b's gap 4 is
  retired by 8c's R1, which lowers what rung 15 reads as open without
  lowering the declaration count. Fork B is ruled (Geoff, 2026-09-08:
  FLAC-only 1.0, the format ladder at 1.1, ROADMAP `formats-1-1`); the
  group non-FLAC question is unresolved, carried.
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

Spend to date: about 61.6M through the 5b-i and 8c plan prep (59.6M
plus 1.99M), plus 8c's own execution against its 2.1M ceiling.
