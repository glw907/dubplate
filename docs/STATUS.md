# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 1b, rung 8 (pass 8a and now **pass 8b, PICTURE and the
conformance instrument**), and rung 4 (4a and 4b) are all closed
(`docs/HISTORY.md`'s "Rung 8, pass 8b" entry). The 4b sitting deployed
the two shadow units to the box 2026-09-08; the shadow's adjudication
stream is live. Geoff's standing touchpoint is divergences under
`/var/lib/dubplate/shadow/pending`, read through `dubplate-server
shadow report` via `systemctl start` or `systemd-run`, never the bare
binary as root. The rung 5 identity and auth spec is ratified at
revision 2.1 (`c5021ce`). Streamer 2.0 (ROADMAP `streamer-2-0`) binds
rungs 5-9 and 12. The box (5.78.148.42) is production and runs
Navidrome, cloudflared, and the guardian and shadow timers; the family
system is not live.

## Next action

Lane-1 (main tree): IDLE. Pass 8b closed at `ce0ac2e`; pass 8c's plan
(absorb the shadow's reader into `tagio`, repoint its call sites) is
next, authored just-in-time from 8b's handoff. Lane-2 (worktree
`.claude/worktrees/lane-2-5a`, branch `lane-2-5a` from `8ac38d7`):
ACTIVE, pass 5a (wf_115f4cec-9c0) in its security review and close; the
merge ritual follows. One executor; phase cap 10M, flag to the
conductor (Geoff, 2026-09-08).

## Open items (carried)

- Fifty-two declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four, rung 1b's two, rung 8a's seven, rung 8b's
  nine), none blocking. The `io.ReaderAt`-injection gap (`Parse` takes
  a path only) is 8c's call alone. Fork B (FLAC-only 1.0, MP3 after
  cutover) stays open, owed to Geoff; the family non-FLAC question is
  unresolved, carried.
- Four-knob wall registration unblocked, named for the first rung 3 pass
  after 1b, joined by `AdminEmail`'s contact row, five 3b leftovers,
  `profile_hashes` (m1), the response-status extension (2b/2c),
  `Case.Extra`, `MergePseudoRelease`'s `slices.Clone` (rung 9).
- The 2026-09-04 agentic-coding-evidence research doc is corrupted,
  untracked until regenerated (`bug-assignment-rejection-gap`
  re-dispositioned `no-correct-answer`, Geoff 2026-09-07). Rewrite
  post-mortem owed; two albums await review in
  `/srv/music/review/2026-08-31/geoff/`; `FILEBROWSER_ADMIN_PASSWORD`
  orphaned; stale failed units; `DUBPLATE_SITE_NAME` retires when re-minted.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 55.7M through pass 8b's close (8a's 52.7M plus
4b's sitting spend of about 0.05M plus 8b's 2.9M).
