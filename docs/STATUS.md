# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 4a, and **rung 1b (Phase 0b, the config wall-row
registry) are closed**, HEAD `db040ec` (`docs/HISTORY.md`'s "Rung 1b"
entry). The config wall is now registry contents: `Load` takes a
`*config.Registry`, thresholds are one family among four, and a rung
declares its keys in its own file, never by editing
`internal/config/keys.go` (CLAUDE.md). **Rung 4, pass 4b, is still OPEN
at the sitting boundary**: Tasks 0 through 7 landed; Task 8 hands off to
Task 9, Geoff's attended deploy sitting. Read `docs/HISTORY.md`'s 4b
entry before the sitting; it carries the checklist, the ten retired
resolutions, and the conductor's `deploy.sh --check` record. Streamer
2.0 (ROADMAP `streamer-2-0`) binds rungs 5-9 and 12 via CLAUDE.md. The
box (5.78.148.42) runs Navidrome, cloudflared, and guardian timers only;
the family system is not live, and the box is production.

- lane-1 (main tree): ACTIVE, pass 8a chain running (pass-execute wf_a6f1705e-408, launched 2026-09-08 04:50; Task 0 clean; window cap 15M binds before the pass's own 4.5M ceiling, so the chain may stop at a task boundary with STATUS pointing at the resume).
lane-2 (worktree): **UNBLOCKED** (0b and 0c both landed). doc lane: pass
8a plan prep.

## Next action

Lane-1: pass 8a, plan `docs/superpowers/plans/2026-09-07-rung-8a-tagio-writer.md`
(revision 2, 4.5M ceiling), Task 0 first. Separately, whenever Geoff is
available: Task 9, the rung 4b sitting (ruling 6), runs
`docs/HISTORY.md`'s 4b checklist against the box and closes 4b.

## Open items (carried)

- Thirty-six declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four, rung 1b's two), owned in HISTORY, none blocking.
- Four-knob wall registration is unblocked, named for the first rung 3
  pass after 1b, joined by `AdminEmail`'s contact row, five 3b
  leftovers, `profile_hashes` (m1), the response-status extension
  (2b/2c), `Case.Extra`, `MergePseudoRelease`'s `slices.Clone` (rung 9).
- The 2026-09-04 agentic-coding-evidence research doc is corrupted and
  stays untracked until regenerated. `bug-assignment-rejection-gap` is
  re-dispositioned to `no-correct-answer` (Geoff, 2026-09-07; `53f0405`,
  `6aadf83`).
- Adjudication batch 1 fully ruled; rewrite post-mortem owed; two albums
  await review in `/srv/music/review/2026-08-31/geoff/`;
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

Spend to date: about 49.1M through rung 1b's close (HISTORY has
per-pass figures; the 4b sitting's spend is separate).
