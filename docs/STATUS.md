# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 4a, rung 1b, and **rung 8 pass 8a (the tagio writer core
and reader) are closed** (`docs/HISTORY.md`'s "Rung 8, pass 8a" entry).
`internal/engine/tagio` exports `Parse`, `Read`, `Write`, `TagSet`,
`Mapping()`/`WriteOrder()`, `Serialize`, and two sentinels; a committed
golden container anchors the writer's bytes to another implementation's.
**Rung 4, pass 4b, is still OPEN at the sitting boundary**: Tasks 0
through 7 landed; Task 8 hands off to Task 9, Geoff's attended deploy
sitting; read `docs/HISTORY.md`'s 4b entry first. Streamer 2.0 (ROADMAP
`streamer-2-0`) binds rungs 5-9 and 12. The box (5.78.148.42) runs
Navidrome, cloudflared, and guardian timers only; not live, production.

- lane-1 (main tree): IDLE, pending pass 8b's plan (just-in-time from
  8a's handoff); `code-simplifier` owed over 8a's Go at 8b's opening.
- lane-2 (worktree): UNBLOCKED; rung 5's spec prep is the candidate.

## Next action

Author pass 8b's plan (PICTURE and conformance) from 8a's handoff
(above), then run `code-simplifier` over 8a's changed Go before Task 1
dispatches. Separately, whenever Geoff is available: rung 4b's Task 9,
the sitting (ruling 6), runs the 4b checklist against the box.

## Open items (carried)

- Forty-three declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four, rung 1b's two, rung 8a's seven), owned in
  HISTORY, none blocking; the `io.ReaderAt`-injection gap (`Parse`
  takes a path only) is 8b's or 8c's call. Fork B (FLAC-only 1.0, MP3
  after cutover) stays open, owed to Geoff; the family non-FLAC
  question is unresolved, carried.
- Four-knob wall registration unblocked, named for the first rung 3
  pass after 1b, joined by `AdminEmail`'s contact row, five 3b
  leftovers, `profile_hashes` (m1), the response-status extension
  (2b/2c), `Case.Extra`, `MergePseudoRelease`'s `slices.Clone` (rung 9).
- The 2026-09-04 agentic-coding-evidence research doc is corrupted,
  untracked until regenerated (`bug-assignment-rejection-gap`
  re-dispositioned `no-correct-answer`, Geoff 2026-09-07).
- Adjudication batch 1 fully ruled, rewrite post-mortem owed, two
  albums await review in `/srv/music/review/2026-08-31/geoff/`;
  `FILEBROWSER_ADMIN_PASSWORD` orphaned; stale failed units;
  `DUBPLATE_SITE_NAME` retires when re-minted via `secret-set.sh`.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 52.7M through pass 8a's close (4b's sitting spend
is separate).
