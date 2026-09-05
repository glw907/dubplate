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

- lane-1 (code, main tree): **pass 3c at its close step**, session handed
  off 2026-09-05 09:55 for a Claude Code restart. Landed and green: Tasks
  1-7 (e9364ff..18a1292, three fix rounds), Task 5b (78a6e5a, ruling rev
  2.2), the simplify pass (7809558) and the shutdown-arm proof (0082ae6).
  **Resume prompt:** (1) dispatch a `diff-reviewer` (opus) over 7809558
  and 0082ae6 (the one that was stopped mid-read; criteria in the plan's
  Execution model and rev 2.2 note); (2) run Task 8, the close, per the
  plan, feeding it `~/.claude/projects/-var-home-glw907-Projects-dubplate/
  handoff/3c-task-records.md` and `3c-nonblocking.md`, then a
  `prose-voice-reviewer` pass and a fold, then the batch commit and push;
  (3) 3d plan-prep workflow against the quiet tree (3b/3c prep scripts in
  this project's `workflows/scripts/` are the shape), then 3d execution.
  Spend so far: prep 2.55M (two runs, a usage-limit pause), Task 0 0.11M,
  chain 1.74M, 5b 0.13M, simplify 0.15M, ≈4.7M of 6M; ceiling flag routes
  to the conductor under the overnight grant. Guard scripts and the AC
  state log are in the same `handoff/` dir; size inhibitor holds to the
  clock (memory note `overnight-hold-sized-to-clock`). The two untracked
  research docs under `docs/superpowers/research/` (outreach, evidence)
  are sidebars, deliberately uncommitted.
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
