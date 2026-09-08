# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 1b, rung 8 pass 8a (the tagio writer core and reader),
and **rung 4 (4a and 4b) are all closed** (`docs/HISTORY.md`'s "Rung 4,
pass 4b" entry). The 4b sitting deployed the two shadow units to the box
2026-09-08: `systemctl enable --now dubplate-shadow.timer` started the
evidence clock, `started-at` reads `2026-09-08T17:30:30.116835324Z` on
the box, and the shadow's adjudication stream is now live. Geoff's
standing touchpoint is divergences under
`/var/lib/dubplate/shadow/pending`, read through `dubplate-server shadow
report` via `systemctl start` or `systemd-run`, never the bare binary
run as root. The rung 5 identity and auth spec is ratified at revision
2.1 (`c5021ce`), Geoff's six rulings applied. Streamer 2.0 (ROADMAP
`streamer-2-0`) binds rungs 5-9 and 12. The box (5.78.148.42) runs
Navidrome, cloudflared, guardian timers, and now the shadow timers; not
live, production.

## Next action

Lane-1 (main tree, IDLE): author pass 8b's plan (PICTURE and
conformance) from 8a's handoff, then run `code-simplifier` over 8a's
changed Go before Task 1 dispatches. Lane-2 (worktree, IDLE): author
pass 5a's plan from the rung 5 spec. Both plans are just-in-time.

## Open items (carried)

- Forty-three declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four, rung 1b's two, rung 8a's seven), none blocking;
  the sitting added none. The `io.ReaderAt`-injection gap (`Parse` takes
  a path only) is 8b's or 8c's call. Fork B (FLAC-only 1.0, MP3 after
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
  orphaned; stale failed units; `DUBPLATE_SITE_NAME` retires when
  re-minted.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 52.7M through pass 8a's close, plus 4b's sitting
spend of about 0.05M conductor turns.
