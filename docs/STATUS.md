# dubplate status

## Current state

Rungs 1b, 3, 4, 8, 5a-5c-iii, and rung 6 through 6b-ii-b closed; 5c-iii and 6b-ii-b (`restore`)
on 2026-09-25, HISTORY's top two entries. 6c is merged on master (`e90998a`), its close unfolded (5c-iii: segments A to C not recorded separately, window 2's
total about 11.0M across all its passes, about 1.6M for tonight's review round and close).

## Next action

- **Lane 1, 5c-iv** (`plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`, Tasks 7-11,
  ~9.2M): Tasks 0, 7, 8, 9a, 9b accepted (last `1fc7260`, 9b after an escalation and one fix round);
  register gate folded. Next: segment E (Task 10 render-and-read, the simplifier round with 9b's
  five stale-comment nits, three reviewers, one fold, Task 11 with architecture reads).
- **Lane 2, 6c**: merged `1fc7260..e90998a` (range-diff all `=`, disjoint files, both gates
  green; the merged-result diff-reviewer returned accept, no semantic conflict; record it in
  6c's HISTORY entry at the fold). Next: six architecture reads and the
  ledger fold (drafts `docs/superpowers/research/handoffs/draft-*-6c.md`, with the must-keep list),
  then remove worktree `lane-2-6c`. Then transliteration (`plans/2026-09-25-translit-table-at-setup.md`,
  approved, 3.3M) or 5d-i once Geoff approves 5d.

**Overnight 2026-09-25 (Geoff):** 27M cap (flag 21.6M), stopped at the flag about 03:30; 5d-i
after 5c-iv closes (6c has merged).
**Resume:** read this file and memory `overnight-window-2026-09-25`; start 5c-iv segment E in
lane 1 and 6c's architecture reads and ledger fold. No workflow is running. Box read-only;
nothing deploys until 5d.

## Owner questions

- Navidrome 0.64 bump for GHSA-82gh: open (Geoff 2026-09-25: sharing stays at zero until 0.8).
- TypeScript 6.0.3 to 7.0.2, a major (5c-iv Task 0's pin check, `rung-5c-iv-task-0-report.md`).
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- 5d (rev 2.2) awaits two morning rulings: 1, ceilings 5d-i 7.0M / 5d-ii 3.5M; 2, the sitting.
- 6c crossed its 4.32M pass flag (about 4.5M plus the close, under the 5.4M ceiling); continued.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master), session 2026-09-26 | OPEN: 5c-iv segment E, Task 10 running (`wf_06a26525-670`); then the read, reviews, Task 11 |
| 2 | worktree `lane-2-6c` (branch `lane-2-6c`) | 6c merged at `e90998a`; six architecture reads running (read-only, in this worktree); ledger fold, worktree removal next |
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
  actor. To 5c-iv: HISTORY 5c-iii "5c-iv handoff". 6c's handoff to 6d: its draft
  HISTORY. To 5d-i: `NewTrustedProxyConfig` has no caller.
- Errata and architecture-cleanup reads, file:line in HISTORY: 5c-iii's nine errata and latent
  items; 6b-ii-b's (spec criterion 16, `:712`; plan `:458`, `:459`, `:936`, `:1711`); the 5c-ii,
  6b-i, 6a, window 2, 5c-iii, 6b-ii-b reads; matcher re-run trigger `strdist_test.go:452`. 99
  coverage gaps, none blocking. Staged legs: 325 process citations, 128 exported-unused; the citation leg misses `R-6c` patterns.
- Photo lane, after lane 1: a `check.sh` integrity line, Ruling 3's photo erratum to rung 7, a
  source-URL recheck, a blob-retention call. Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`,
  beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, sqlite v1.59.0, rclone v1.74.3.
