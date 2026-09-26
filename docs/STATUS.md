# dubplate status

## Current state

Rungs 1b, 3, 4, 8, 5a-5c-iii, and rung 6 through 6c closed. 6c (the routed schema and settings
home) closed on 2026-09-26, merged at `e90998a`; it joins 5c-iii and 6b-ii-b as HISTORY's top
entries. Its architecture reads (`research/2026-09-26-rung-6c-architecture-reads.md`) route six
correctness items to 6c-fix, two wiring items to 6d, and the rest to architecture cleanup.

## Next action

- **Lane 1, 5c-iv** (`plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`, Tasks 7-11,
  ~9.2M): Tasks 0, 7, 8, 9a, 9b accepted (last `1fc7260`); register gate folded; segment E running.
- **Lane 2, 6c-fix**: the six correctness items from 6c's reads (config C1 and C2, ops/local L1
  and L2, store's nil-lock `DupFiles` panic, cmd's two store-state lists). 5d-i follows 5c-iv's close.

**Resume:** read this file. Lane 1 continues 5c-iv segment E after Task 10: a fresh-context read
of the captures, the fold, a re-read, the simplifier (with 9b's five stale-comment nits from its
fix-round review, run `wf_5b983a74-956`, and the line-number citations in 5c-iv's web comments),
three reviewers, and Task 11 with
architecture reads. Lane 2 merges 6c-fix through the merge ritual, then runs transliteration
(`plans/2026-09-25-translit-table-at-setup.md`, approved, 3.3M) or 5d-i once 5c-iv closes and Geoff rules 5d.
Box read-only; nothing deploys until 5d.

**Unattended 2026-09-26 (Geoff):** 16M cap (flag 12.8M); finish 5c-iv, merge 6c-fix, then
transliteration; 5d's ceilings deferred to the next sitting, so 5d-i does not start.

## Owner questions

- Navidrome 0.64 bump for GHSA-82gh: open (Geoff 2026-09-25: sharing stays at zero until 0.8).
- TypeScript 6.0.3 to 7.0.2, a major (5c-iv Task 0's pin check, `rung-5c-iv-task-0-report.md`).
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- 5d (rev 2.2) awaits two morning rulings: 1, ceilings 5d-i 7.0M / 5d-ii 3.5M; 2, the sitting.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master), session 2026-09-26 | OPEN: 5c-iv segment E, Task 10 running (`wf_06a26525-670`); then the read, reviews, Task 11 |
| 2 | worktree `lane-2-6c` (branch `lane-2-6c`) | 6c CLOSED (merged `e90998a`); OPEN: pass 6c-fix running (`wf_98ff134f-a42`); then its merge, then the worktree is removed or reused for transliteration |
| photos | main tree (master), owned by the separate docs session | OPEN: `plans/2026-09-25-photo-curation-round-one.md`; writes only `photos/`, `tools/photos/`, its research files; never runs `scripts/check.sh`; ceiling 5.5M (flag 4.4M) |

## Open items (carried)

- **Owed:** the streamer spec's seven open items (rev 2.5) to its study rung; from the site spec,
  settle `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference
  generator as a hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- No deploy until 5d (`plans/2026-09-25-rung-5d-boundary-deployment-close.md`): the nine
  `magic-link` env keys, `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys'
  `doctor` row, Ruling 6's `Referrer-Policy` residual, the branding-in-both-schemes line,
  `deploy.sh`'s clean-`web/` build, 6a's and 6b-i's runbook rows (`KillMode=mixed`, rclone
  v1.74.3+, crypt remote, exit-7 remedy), and 6c's Ruling 2 (`deploy.sh` halts on `check`'s
  nonzero exit before restarting `serve`). 5d's deploy gate checks that 6d's two wiring items
  below have landed. 6b-ii-b's runbook lines: rung 16 (its HISTORY entry).
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason
  set, CREDITS, picked images at stable paths an export manifest can list, 6c's `member_quota`
  reservations and default-collection preference. To rung 9: filter, facets, trigram. To rung 12:
  streamer-admin mapping, restore-before-12 losing stars/playlists, shared-origin M2 and 5c-iii's
  CSP-scope M2, the streamer database's restore place. Full list: HISTORY 5c-ii and 5c-iii "What
  stayed open". To the revoke verb: the epoch store's first writer. To rung 14: 6c's admin write
  path and a resolving channel for an `admin` row that refuses `serve`'s start.
- To 6d: `ServeOptions.Reconcile` wiring, store identity, migration-wait key, `member:<principal>`
  actor, and 6c's handoff (HISTORY 6c "The 6c to 6d handoff"). From 6c's reads, both before any
  deploy that exposes the strings: wire settings `apply` at `serve` start (the operator strings at
  `cmd/dubplate-server/backup.go:274` and `store.go:280-284` claim it today), and make `serve`
  open the store through `OpenServe` (unwired, so `OpenWrite`'s "start serve first" refusal cannot
  be satisfied on a fresh box). To 5c-iv: HISTORY 5c-iii "5c-iv handoff". To 5d-i:
  `NewTrustedProxyConfig` has no caller.
- Errata and architecture-cleanup reads, file:line in HISTORY: 5c-iii's nine errata and latent
  items; 6b-ii-b's (spec criterion 16, `:712`; plan `:458`, `:459`, `:936`, `:1711`); 6c's
  errata 6c-E1 to 6c-E10; the 5c-ii, 6b-i, 6a, window 2, 5c-iii, 6b-ii-b reads; 6c's reads'
  non-correctness findings (research file above), including the NaN float-parser disagreement
  and 117 plus 41 process-citation lines; matcher re-run trigger `strdist_test.go:452`. 103
  coverage gaps (6c added 4), none blocking. Staged legs: 325 process citations, 128
  exported-unused; the citation leg misses `R-6c` patterns.
- Photo lane, after lane 1: a `check.sh` integrity line, Ruling 3's photo erratum to rung 7, a
  source-URL recheck, a blob-retention call. Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`,
  beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, sqlite v1.59.0, rclone v1.74.3.
