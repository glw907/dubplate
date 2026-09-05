# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. **Rung 2 is
closed. Rung 3, passes 3a and 3b are closed.** `internal/engine/matcher` (21
distance components) and `internal/engine/musicbrainz` (WS/2 vocabulary, track
extraction, aliases and credits, release events, `TranslateAlbum`, the pseudo
merge) are green; `internal/harness/conform` replays tier 0 at 72/72 rows and
grades the translation tier at 50/50 ledger rows, both bit-identical.
`corpus/analysis/album_info.json` (19 subjects, 50 rows) and the `defect_free`
block on ten adjudication rows are committed and gate-checked. Coverage tables,
evidence, and the rediscover-guards live in `docs/HISTORY.md`'s 3a and 3b
entries; read both before 3c. The family system is NOT live until launch; the
box (5.78.148.42) runs Navidrome + cloudflared + guardian timers only.

## Rules for a new session

Read CLAUDE.md's Build process, then the parallel-approach spec (rev 2, the
two-code-lane model), then this file. The live-executor check is the lane
registry below, not `pgrep`. Box is production; quiet-box check before writes.

## Lane registry

- lane-1 (code, main tree): PREP for **pass 3c** (WS/2 client, Lucene query
  construction, the limiter, the stub's `Case.Extra` replay channel). The
  plan-prep workflow `wf_ede6e00d-bee` writes only the 3c planning
  inventory, plan, and plan review under `docs/superpowers/`; if
  interrupted, relaunch fresh or land what exists. **Standing grant (Geoff,
  2026-09-04 evening): run unattended through 3c execution, 3d prep, and 3d
  execution as the clock allows, workflow mode, /loop to completion.**
- lane-2 (code, worktree): BLOCKED until Phase 0b (rung 1b) and 0c land. Doc
  lane (main tree): IDLE; Phase 0a is complete.

## Open items (carried)

- `profile_hashes` (m1) and the response-status extension stay open, carried
  since 2b/2c. Rung 3 carries **nine declared corpus-coverage gaps**, three from
  3a and six from 3b, each owned in its own HISTORY entry; none blocks 3c.
- To 3c: `dataSourcePenalty`'s configured-versus-literal fallback;
  `MergePseudoRelease`'s shallow track-slice copy; the unported browse fallback,
  which 3c owns if a corpus case ever crosses 500 tracks.
- To 3d: `Preferred.Media`/`Preferred.Countries` config-load validation (3d's
  alone now); the compat switch's pattern-table selection and the `\b` regime
  divergence; tie-mirroring needing both `inclArtist` call shapes.
- The four-knob wall registration is the follow-on task for the first rung 3
  pass after rung 1b lands; `internal/config/keys.go` stays untouched until then.
  Five 3b `code-simplifier` leftovers sit beside it in HISTORY, none blocking.
- Adjudication batch 1 is fully ruled: 18 oracle-wrong, 3 oracle-correct, zero
  pending.
- Rewrite post-mortem owed; two albums in /srv/music/review/2026-08-31/geoff/
  await review; FILEBROWSER_ADMIN_PASSWORD orphaned; stale failed units.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |

Spend to date: ≈25.0M across rungs 2 and 3 (per-pass figures in HISTORY's Budgets lines).
