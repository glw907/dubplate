# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. **Rung 2 is
closed. Rung 3, passes 3a, 3b and 3c are closed.** `internal/engine/matcher`
(21 distance components, `Comp`, `IsVariousArtistsName`) and
`internal/engine/musicbrainz` (WS/2 vocabulary, `TranslateAlbum`, the pseudo
merge, `query.go`, `client.go`, `limiter.go`) are green. `internal/harness/conform`
replays tier 0 at 72/72, translation at 50/50, request construction at 215
cases/303 lookups/48 searches, all bit-identical or zero-miss. Ledger and
gap evidence live in `docs/HISTORY.md`'s 3a/3b/3c entries; read all three
before 3d. Family system NOT live; box (5.78.148.42) runs Navidrome +
cloudflared + guardian timers only.

## Rules for a new session

Read CLAUDE.md's Build process, the parallel-approach spec (rev 2), then this
file. Live-executor check is the lane registry, not `pgrep`; box is production.

## Lane registry

- lane-1 (main tree): **3d executing** under the 2026-09-05 unattended
  grant; plan `docs/superpowers/plans/2026-09-05-rung-3d-verdicts-gate.md`
  rev 2.1, ceiling 6M re-derived after Task 2, chain over Tasks 1-8 in
  the main tree. Flag routes to the conductor through 2026-09-07 09:00;
  ruling F stops only its own task; window cap 15M, ≈3.2M spent.
  lane-2 (worktree): BLOCKED until Phase 0b/0c land. Doc lane: IDLE.

## Open items (carried)

- Rung 3 carries **ten declared corpus-coverage gaps**, owned in HISTORY; none
  blocks 3d. `dataSourcePenalty` (R15) CLOSED; `MergePseudoRelease`'s aliasing
  hazard closed by construction at 3c's call site, the `slices.Clone` fix
  re-owned to 3d/rung 9; browse fallback a non-port past 500 tracks.
- To 3d: `Preferred.Media`/`Countries` config-load validation; the compat
  switch's pattern-table selection and the `\b` regime divergence; tie-
  mirroring needing both `inclArtist` shapes; 3c's R1 seam, R6's stub blind
  side, R7's `Extra` producer (`cmd/harness`, lane E1), R11's `AdminEmail`.
- Four-knob wall registration (`StrongRecThresh`, `MediumRecThresh`,
  `RecGapThresh`, `SearchLimit`) is the follow-on for the first rung 3 pass
  after rung 1b, `internal/config/keys.go` untouched until then; `AdminEmail`
  contact row and five 3b `code-simplifier` leftovers beside it, none blocking.
- `profile_hashes` (m1) and the response-status extension stay open since
  2b/2c, the latter declined a third pass running; adjudication batch 1
  fully ruled (18 oracle-wrong, 3 oracle-correct, zero pending). Rewrite
  post-mortem owed; two albums in /srv/music/review/2026-08-31/geoff/
  await review; FILEBROWSER_ADMIN_PASSWORD orphaned; stale failed units.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |

Spend to date: ≈29.9M across rungs 2 and 3 (per-pass figures in HISTORY).
