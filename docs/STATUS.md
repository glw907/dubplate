# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d) and rung 4a are closed. **Rung 4, pass 4b, is OPEN at the
sitting boundary**: Tasks 0 through 7 landed in the main tree (HEAD
`89920ba`); Task 8 (this file) hands off to Task 9, Geoff's attended
deploy sitting. Read `docs/HISTORY.md`'s 4b entry before the sitting; it
carries the checklist, the ten retired resolutions with file:line, and
the conductor's `deploy.sh --check` record. Streamer 2.0 (ROADMAP
`streamer-2-0`) binds rungs 5-9 and 12 via CLAUDE.md. The box
(5.78.148.42) runs Navidrome, cloudflared, and guardian timers only; the
family system is not live, and the box is production.

**Lane registry.** lane-1 (main tree): pass 4b OPEN at the sitting
boundary, ACTIVE on rung 1b next. lane-2 (worktree): BLOCKED until rung
1b lands. doc lane: pass 8a plan prep.

## Next action

**Task 9, the sitting, is Geoff's** (ruling 6): run `docs/HISTORY.md`'s
4b checklist against the box, in order. Its close records `started-at`,
the SELinux verdicts, the free-space baseline, and the shadow's
first-pass journal, then writes the pass-close HISTORY entry. Until
then the conductor works unblocked: rung 1b next (plan
`docs/superpowers/plans/2026-09-07-rung-1b-config-registry.md`, revision
2, 2.5M ceiling, eight tasks, `KindString` added), then pass 8a (plan in
prep), then a rung 5 spec draft as a doc lane.

## Open items (carried)

- Thirty-four declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four), owned in HISTORY, none blocking.
- Four-knob wall registration follows rung 1b, joined by rung 4's
  `KindFloat`; `AdminEmail`, five 3b leftovers, `profile_hashes` (m1),
  the response-status extension (2b/2c), `Case.Extra`, and
  `MergePseudoRelease`'s `slices.Clone` (rung 9).
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

Spend to date: about 45.4M through 4b's unattended half (HISTORY has
per-pass figures; the sitting's spend is separate).
