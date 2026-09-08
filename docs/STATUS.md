# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d) is closed at the rung 3d figures. **Rung 4a is closed**:
passes 4a-i and 4a-ii are both in the tree and green.
`internal/engine/shadow` carries the whole shadow runner,
`cmd/dubplate-server` carries the `shadow` verb's three subcommands,
and rung 4's exit gate runs one end-to-end `observe` pass over a
fourteen-directory inbox. Rung 4's six owner rulings are ratified
(Geoff, 2026-09-07, all as recommended). Read HISTORY's 4a-ii entry
before touching `internal/engine/shadow`.

Streamer 2.0 is a ratified design (ROADMAP `streamer-2-0`); its
forward hooks bind rungs 5 through 9 and 12 via CLAUDE.md.

The family system is not live; the box (5.78.148.42) runs Navidrome +
cloudflared + guardian timers only, and it is production. A new session
reads CLAUDE.md's Build process, then the parallel-approach spec (rev 2),
then this file; the live-executor check is the lane registry.

## Next action

Pass 4b, the rung 4 deploy, is executing in the main tree under the
2026-09-07 unattended window (grants in memory
`unattended-window-2026-09-07`: 8M cap, flag to the conductor,
read-only `deploy.sh --check` allowed, no box write). Plan:
`docs/superpowers/plans/2026-09-07-rung-4b-deploy.md` (revision 2,
3M ceiling, flag 2.4M, checkpoint after Task 4). Task 0 ran clean on
2026-09-07. Checkpoint 2026-09-07 23:40: Tasks 2, 1, 3, 4 landed (6f698b1,
two fix rounds), Tasks 5 and 6 next; the conductor then
runs the read-only `--check`, dispatches Task 7's fresh-context unit
read, and Task 8 writes this file's handoff. **Task 9, the sitting, is
Geoff's** (ruling 6); the pass stays OPEN until it runs, and its close
records `started-at`, the fact the launch gate's 720h floor check
reads.

Rung 8's spec is ratified at revision 2
(`docs/superpowers/specs/2026-09-07-rung-8-tag-io.md`, a7e10c4; fourteen
conductor rulings; fork B, a FLAC-only 1.0 with MP3 as the first
follow-on pass, applied provisionally and owed to Geoff); pass 8a's
plan is in prep. After 4b's Task 8: rung 1b in the main tree (plan
`docs/superpowers/plans/2026-09-07-rung-1b-config-registry.md`,
revision 2, 2.5M ceiling, flag 2.0M, eight tasks, KindString added
from the 0c standard), then rung 8 if the window cap allows, then a
rung 5 spec draft as a doc lane. Phase 0c is complete:
`docs/superpowers/specs/2026-09-07-design-language-standard.md`
(revision 2, six conductor rulings applied; one owner note rides to
the rung 5 sitting, the CSP `style-src 'unsafe-inline'` relaxation).

## Lane registry

- lane-1 (main tree): ACTIVE, pass 4b chain (pass-execute), then
  rung 1b. One executor; nothing else edits code here.
- lane-2 (worktree): BLOCKED until rung 1b lands (0c is done).
- doc lane: IDLE; candidates queued: the rung 5 spec draft, the 2.0
  study rung.

## Open items (carried)

- Thirty declared coverage gaps, rung 3's twenty-one plus rung 4a's
  nine, owned in HISTORY, none blocking. Read the total, never
  re-count it.
- Four-knob wall registration follows rung 1b, now joined by rung 4's
  `KindFloat` need; `AdminEmail` and five 3b leftovers, none blocking.
  `profile_hashes` (m1) and the response-status extension open since
  2b/2c; `Case.Extra` and `MergePseudoRelease`'s `slices.Clone` to rung 9.
- Five retireable "provisional" ruling marks (three in
  `internal/engine/shadow`, two in `cmd/dubplate-server/shadow.go`)
  retire in 4b Task 6; `flac.go:62` stays as a rung 8 handoff mark.
- 17 unledgered tied candidates filed for lane E1.
  `bug-assignment-rejection-gap` is re-dispositioned to
  `no-correct-answer` (Geoff, 2026-09-07; 53f0405, 6aadf83): compat
  mode still grades its six probe cases, default mode does not.
- `docs/superpowers/research/2026-09-04-agentic-coding-public-evidence.md`
  is corrupted (escaped newlines, a leaked transcript tail) and stays
  untracked until Geoff regenerates it; the outreach doc stays
  untracked by design.
- Adjudication batch 1 fully ruled. Rewrite post-mortem owed; two albums
  await review in `/srv/music/review/2026-08-31/geoff/`;
  FILEBROWSER_ADMIN_PASSWORD orphaned; stale failed units.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 42.8M through 4a-ii (HISTORY has per-pass figures).
