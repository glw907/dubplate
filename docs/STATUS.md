# dubplate status

## Current state

Rungs 1b, 3, 4, 8, 5a through 5c-iv, and rung 6 through 6c closed. 5c-iv (the deck, the three
real pages, the header, the render-and-read) closed 2026-09-26 at `1a08eae`, so rung 5 continues
at 5d. 5c-iv, 6c, and 5c-iii are HISTORY's top entries; 6c-fix merged at `8950de7`.

**First look for Geoff:** the 15 final contact sheets (one per page state, all ten cells each)
are in `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/22507da8-e76c-4351-96f1-9e6d7c33ef0d/scratchpad/captures-final-1a08eae/`.
Any full gate regenerates them under `web/rendered-captures/` (ignored).

## Next action

**Overnight 2026-09-26 (Geoff):** 22M cap (flag 17.6M). Queue, serial in the main tree:
transliteration (master pushed through `d618bff`; push again at each close, gitleaks first), the photo integrity gate line, then 5d-i. The 6d plan awaits approval. Memory
`overnight-window-2026-09-26` holds the grants.

**Resume:** read this file and memory `overnight-window-2026-09-26`. Next is transliteration
(`plans/2026-09-25-translit-table-at-setup.md`, approved, 3.3M, Task 0 first), then the photo
gate line, then 5d-i (plan rev 2.2, approved). Box read-only; nothing deploys until 5d.

## Owner questions

- 6d plan rev 2.2 (`plans/2026-09-26-rung-6d-identity-serve-store.md`) awaits approval: three
  rulings (one pass at 11.5M; the epoch to the revoke verb; store identity to rung 16), all
  recommended yes. Reviewed by three lenses, verified, folded twice, prose-reviewed.
- Navidrome 0.64 bump for GHSA-82gh: open (Geoff 2026-09-25: sharing stays at zero until 0.8).
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master), overnight 2026-09-26 | OPEN: transliteration; segment A accepted (Tasks 1 `5318671` `1ca79d7`, 1b `194c01c`, 2 `e8e158d`, 3 `850cb4c` `189a6c8`); spend about 2.1M of 3.3M (flag 2.64M, segment B will cross it; continuing under the overnight grant, as 6c did); segment B (Task 4, simplifier, Task 5) next |
| 2 | none | 6c-fix merged at `0e3272c`..`8950de7`; worktree `lane-2-6c` removed |
| photos | main tree (master), owned by the separate docs session | ROUND 1 CLOSED at `c10c0f2` (116 picks, ~13M against a raised 7.5M ceiling); round two unplanned: targets in `research/photo-curation-rounds/round-1.md` (Round close) and `artist-sweep-list.md`; one Run-D.M.C. frame awaits Geoff's call |

## Open items (carried)

- **Owed:** the streamer spec's seven open items (rev 2.5) to its study rung; from the site spec,
  settle `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference
  generator as a hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- No deploy until 5d (`plans/2026-09-25-rung-5d-boundary-deployment-close.md`): the nine
  `magic-link` env keys, `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys'
  `doctor` row, Ruling 6's `Referrer-Policy` residual, the branding-in-both-schemes line,
  `deploy.sh`'s clean-`web/` build, 6a's and 6b-i's runbook rows (`KillMode=mixed`, rclone
  v1.74.3+, crypt remote, exit-7 remedy), and 6c's Ruling 2 (`deploy.sh` halts on `check`'s
  nonzero exit before restarting `serve`). 5c-iv adds two rows, the rendered leg's absence on
  the box and the unknown-address disclosure in the runbook (HISTORY 5c-iv "To 5d"). 5d's
  deploy gate checks that 6d's two wiring items below have landed. 6b-ii-b's runbook lines:
  rung 16 (its HISTORY entry). To 5d-i: `NewTrustedProxyConfig` has no caller.
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason
  set, CREDITS, picked images at stable paths an export manifest can list, 6c's `member_quota`
  reservations and default-collection preference, and 5c-iv's handoff (HISTORY 5c-iv "To rung
  7", the styled 404 among it). To rung 9: filter, facets, trigram. To rung 12: streamer-admin mapping, restore-before-12 losing
  stars/playlists, shared-origin M2 and 5c-iii's CSP-scope M2, the streamer database's restore
  place. Full lists: HISTORY 5c-ii, 5c-iii, and 5c-iv "What stayed open". To the revoke verb:
  the epoch store's first writer. To rung 14: 6c's admin write path and a resolving channel
  for an `admin` row that refuses `serve`'s start.
- To 6d: `ServeOptions.Reconcile` wiring, store identity, migration-wait key, `member:<principal>`
  actor, and 6c's handoff (HISTORY 6c "The 6c to 6d handoff"). From 6c's reads, both before any
  deploy that exposes the strings: wire settings `apply` at `serve` start (the operator strings at
  `cmd/dubplate-server/backup.go:274` and `store.go:280-284` claim it today), and make `serve`
  open the store through `OpenServe` (unwired, so `OpenWrite`'s "start serve first" refusal cannot
  be satisfied on a fresh box). To 6d Task 4: `rosterResetServer`'s baseline (HISTORY 5c-iv).
- Errata and architecture-cleanup reads, file:line in HISTORY: 5c-iii's nine errata and latent
  items; 5c-iv-E1 to E3 (after 5c-iii's 1 and 2); 6b-ii-b's (spec
  criterion 16, `:712`; plan `:458`, `:459`, `:936`, `:1711`); 6c's errata 6c-E1 to 6c-E10; the
  5c-ii, 6b-i, 6a, window 2, 5c-iii, 6b-ii-b reads; 6c's reads' non-correctness findings,
  including the NaN float-parser disagreement and 117 plus 41 process-citation lines; 5c-iv's
  cleanup routings (`research/2026-09-26-rung-5c-iv-final-reads.md`, plus the login title-argument
  and h1-fallback rendered proof, and 9b's five stale-comment nits from run `wf_5b983a74-956`,
  which the simplifier round did not take); matcher re-run trigger `strdist_test.go:452`. 103 coverage gaps (6c added 4), none
  blocking. Staged legs: 325 process citations, 128 exported-unused; the citation leg misses
  `R-6c` patterns.
- Photo lane, after lane 1: a `check.sh` integrity line, Ruling 3's photo erratum to rung 7, a
  source-URL recheck, a blob-retention call. Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`,
  beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, sqlite v1.59.0, rclone v1.74.3.
