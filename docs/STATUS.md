# dubplate status

## Current state

Rungs 1b, 3, 4, 8, 5a-5c-iii, and all of rung 6 closed; 5c-iii and 6b-ii-b (`restore`) on
2026-09-25, HISTORY's top two entries (5c-iii: segments A to C not recorded separately, window 2's
total about 11.0M across all its passes, about 1.6M for tonight's review round and close).

## Next action

- **Lane 1, 5c-iv** (`plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`, Tasks 7-11,
  ~9.2M): Tasks 0 and 7 accepted; the register gate already ran and was folded (`1f189b7`,
  `c12ef83`); D-ii (8, 9a, 9b) runs as `wf_5b983a74-956`, then segment E (10, reviews, 11).
- **Lane 2, pass 6c** (rev 2.2 `50a1cf8`, 5.4M): segment A (0-3) runs as `wf_62218e14-e00`; then
  transliteration (`plans/2026-09-25-translit-table-at-setup.md`, approved, 3.3M, awaits a lane).

**Overnight 2026-09-25 (Geoff):** 27M cap (flag 21.6M); 5d-i after 5c-iv closes and 6c merges.
**Resume:** read this file and memory `session-2026-09-25-handoff`, then continue lane 1
(`wf_5b983a74-956`) and lane 2 (`wf_62218e14-e00`); relaunch fresh from their journals if the
session ended. Box read-only; nothing deploys until 5d.

## Owner questions

- Navidrome 0.64 bump for GHSA-82gh: open (Geoff 2026-09-25: sharing stays at zero until 0.8).
- TypeScript 6.0.3 to 7.0.2, a major (5c-iv Task 0's pin check, `rung-5c-iv-task-0-report.md`).
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- 5d (rev 2.2) awaits two morning rulings: 1, ceilings 5d-i 7.0M / 5d-ii 3.5M; 2, the sitting.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master), this session | OPEN: 5c-iv D-ii (Tasks 8, 9a, 9b) as `wf_5b983a74-956`; then E |
| 2 | worktree `lane-2-6c` (branch `lane-2-6c`, cut at `392d1cd`) | OPEN: 6c segment A accepted (Tasks 0-3, `20bd27a`..`257617b`, one fix round, about 2.9M); segment B (Tasks 4, 5) running as `wf_a5a9ec89-e5e`, then simplifier, Task 6, merge; re-derived about 5.0M: crosses the 4.32M pass flag, under the 5.4M ceiling (continued under the overnight grant; morning item) |
| photos | main tree (master), owned by the separate docs session | OPEN: `plans/2026-09-25-photo-curation-round-one.md`; writes only `photos/`, `tools/photos/`, its research files; never runs `scripts/check.sh`; ceiling 5.5M (flag 4.4M) |

## Open items (carried)

- **Owed:** the streamer spec's seven open items (rev 2.5) to its study rung; from the site spec,
  settle `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference
  generator as a hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- No deploy until 5d (`plans/2026-09-25-rung-5d-boundary-deployment-close.md`): the nine
  `magic-link` env keys, `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys'
  `doctor` row, Ruling 6's `Referrer-Policy` residual, the branding-in-both-schemes line,
  `deploy.sh`'s clean-`web/` build, 6a's and 6b-i's runbook rows (`KillMode=mixed`, rclone
  v1.74.3+, crypt remote, exit-7 remedy). 6b-ii-b's runbook lines: rung 16 (its HISTORY entry).
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason
  set, CREDITS, picked images at stable paths an export manifest can list. To rung 9: filter,
  facets, trigram. To rung 12: streamer-admin mapping, restore-before-12 losing stars/playlists,
  shared-origin M2 and 5c-iii's CSP-scope M2, the streamer database's restore place. Full list:
  HISTORY 5c-ii and 5c-iii "What stayed open". To the revoke verb: the epoch store's first writer.
- To 6d: `ServeOptions.Reconcile` wiring, store identity, migration-wait key, `member:<principal>`
  actor. To 5c-iv: HISTORY 5c-iii "5c-iv handoff". To 6c: `SAVEPOINT` check, stale `internal/config`
  doc. To 5d-i: `NewTrustedProxyConfig` has no caller.
- Errata and architecture-cleanup reads, file:line in HISTORY: 5c-iii's nine errata and latent
  items; 6b-ii-b's (spec criterion 16, `:712`; plan `:458`, `:459`, `:936`, `:1711`); the 5c-ii,
  6b-i, 6a, window 2, 5c-iii, 6b-ii-b reads; matcher re-run trigger `strdist_test.go:452`. 99
  coverage gaps, none blocking. Staged legs: 326 process citations, 125 exported-unused.
- Photo lane, after lane 1: a `check.sh` integrity line, Ruling 3's photo erratum to rung 7, a
  source-URL recheck, a blob-retention call. Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`,
  beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, sqlite v1.59.0, rclone v1.74.3.
