# dubplate status

## Current state

Rungs 1b, 3, 4, 8, 5a through 5c-iv, rung 6 through 6c, and the transliteration pass (2026-10-03,
code at `5530cbb`: the binary carries no Unidecode data; scoring verbs need `unidecode install`'s
table) closed; rung 5 continues at 5d. HISTORY's top entries: transliteration, 5c-iv, 6c.

## Next action

**Resume:** Read this file. Next: 5d-i in a fresh session (plan rev 2.2 approved 2026-09-26:
7.0M; its deploy is an attended sitting). The photo integrity leg gates since `b7766d6`. Push master after each close (gitleaks first). Box read-only; nothing
deploys until 5d.

**Overnight window (Geoff, 2026-10-04 evening):** run 5d-i to its close, then continue in the same
session through 5d-ii Segment C (Tasks 7, 8, the two adversarial lenses, Task 9), stopping before Task
10's attended sitting. Cap 10.5M across both halves, flag 8.4M: at the flag, finish the task in flight,
write STATUS, and stop. AC confirmed. No deploy; box read-only.

## Owner questions

- 6d plan rev 2.2 (`plans/2026-09-26-rung-6d-identity-serve-store.md`) awaits approval: three
  rulings (one pass at 11.5M; the epoch to the revoke verb; store identity to rung 16), all
  recommended yes. Reviewed by three lenses, verified, folded twice, prose-reviewed.
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | ACTIVE: 5d-i opened 2026-10-04 at `311b74c` (pass class `auth-data`, heavy gate); Task 0 ruled at `fafef77`; Segment A (Tasks 1-3) running as `wf_a9246e81-97e`; dependency sweep: the 2026-10-04 sweep (`0677934`, `b3c2578`) covers this pass |
| 2 | none | 6c-fix merged at `0e3272c`..`8950de7`; worktree `lane-2-6c` removed |
| photos | main tree (master), owned by the separate docs session | ROUND 1 CLOSED at `c10c0f2` (116 picks, ~13M against a raised 7.5M ceiling); round two unplanned: targets in `research/photo-curation-rounds/round-1.md` (Round close) and `artist-sweep-list.md`; one Run-D.M.C. frame awaits Geoff's call |

## Open items (carried)

- **Owed:** the streamer spec's seven open items (rev 2.5) to its study rung; from the site spec,
  settle `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference
  generator as a hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- No deploy until 5d (`plans/2026-09-25-rung-5d-boundary-deployment-close.md`): nine `magic-link` env
  keys, `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys' `doctor` row, Ruling
  6's `Referrer-Policy` residual, branding in both schemes, `deploy.sh`'s clean-`web/` build, 6a's and
  6b-i's runbook rows (`KillMode=mixed`, rclone v1.74.3+, crypt remote, exit-7 remedy), 6c's Ruling 2
  (`deploy.sh` halts on `check`'s nonzero exit before restarting `serve`), 5c-iv's two rows (HISTORY
  5c-iv "To 5d"), and the transliteration runbook row (HISTORY transliteration, "The deploy transition
  is staged": `--check` first, rollback by redeploy, the wheel cache as `--from` source). 5d's deploy
  gate checks 6d's two wiring items below. 6b-ii-b's runbook lines: rung 16. To 5d-i:
  `NewTrustedProxyConfig` has no caller.
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason set,
  CREDITS, picked images at stable paths an export manifest can list, 6c's `member_quota` reservations
  and default-collection preference, 5c-iv's handoff (HISTORY 5c-iv "To rung 7", the styled 404 among
  it). To rung 9: filter, facets, trigram. To rung 12: streamer-admin mapping, restore-before-12
  losing stars/playlists, shared-origin M2 and 5c-iii's CSP-scope M2, the streamer database's restore
  place. Full lists for rungs 7 and 12: HISTORY 5c-ii, 5c-iii, and 5c-iv. To the revoke verb: the
  epoch store's first writer. To rung 14: 6c's admin write path and a resolving channel for a
  `serve`-refusing `admin` row.
- To 6d: `ServeOptions.Reconcile` wiring, store identity, migration-wait key, `member:<principal>`
  actor, 6c's handoff (HISTORY 6c). Before any deploy exposing the strings: wire settings `apply` at
  `serve` start (`cmd/dubplate-server/backup.go:274`, `store.go:280-284` claim it) and open the store
  through `OpenServe` (unwired today, so `OpenWrite`'s "start serve first" refusal cannot be satisfied
  on a fresh box). To 6d Task 4: `rosterResetServer`'s baseline (HISTORY 5c-iv).
- Errata and architecture-cleanup reads, file:line in HISTORY: 5c-iii's nine errata; 5c-iv-E1 to E3;
  6b-ii-b's; 6c-E1 to 6c-E10; the 5c-ii, 6b-i, 6a, window 2, 5c-iii, 6b-ii-b, and 6c reads (the NaN
  float-parser disagreement among 6c's); 5c-iv's routings (`research/2026-09-26-rung-5c-iv-final-reads.md`,
  9b's five nits); transliteration's ROUTE items (`research/2026-09-27-translit-architecture-reads.md`)
  and the four-copy `mainTB` shim in the `TestMain` files; matcher re-run trigger `strdist_test.go:452`.
  103 coverage gaps, none blocking. Staged legs (gate at `5530cbb`): 323 process citations, 129
  exported-unused; the citation leg misses `R-6c` patterns.
- To `give-back` and `packaging`: the deleted table's git history, the process-wide holder's
  tension, the Unidecode-derived items still carried (HISTORY transliteration, "Routed").
- Workstation prerequisite: `make unidecode-wheel` once per machine before `make test`; prove
  "wheel absent" with `-count=1`. A gate failing on a `/tmp` disk-quota write is a lost run.
- Photo lane, after lane 1: `tools/photos/check.test.mjs` fails 2 of 121 (both assume an empty
  `photos/`), so no gate runs it; the integrity leg reads `photos/` from disk, not `git ls-files`;
  `check.sh`'s header omits the leg. Ruling 3's photo erratum to rung 7, a source-URL recheck, a blob-retention
  call. Pinned: Navidrome `0.63.2` (no further bumps: Geoff 2026-10-04, dubplate no longer uses it), cloudflared `2026.9.3`, beets `2.14.1`, mediafile `0.17.0`,
  unidecode `1.4.0` (sha256 pins in `unidecode/pins.go`), lap `0.5.13`, sqlite v1.60.1, rclone
  v1.74.3.
