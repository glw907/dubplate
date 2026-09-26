# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-i and 5c-ii, and rung 6's 6a, 6b-i, and **6b-ii-a** are closed. 6b-ii-a (the ops seam's HTTP projection, `6e6fd12`, about 1.95M against 3.90M) and the **matcher word-boundary pass** (the featuring row's ASCII RE2 rewrite, U+0130 declared and pinned, about 1.25M against 2.75M) both closed 2026-09-25; HISTORY, top two entries. Beets 2.14.1 adopted. The visual identity spec (revision 1.2, `782d864`) is complete, 5c-iii's standing input.

## Next action

- **5c-iii: segments A, B, C done (Tasks 1-5 accepted, last `610ad61`); its review round ran 2026-09-25
  (reviews in `research/2026-09-25-rung-5c-iii-review-*.md`, a fold in progress); Task 6 close next.** Plan `plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`,
  conductor actions 4 and 5: one `code-simplifier` over Tasks 1-5 (paste
  `.claude/instructions/simplifier-brief.md`), `svelte-reviewer` + `daisyui-a11y-reviewer` (given the VI
  spec) + `web-auth-security-reviewer` (head insertions, new routes), one fold; then Task 6 (architecture
  reads for `internal/serve`, `cmd/dubplate-server`, `web`; the ledger fold; the refresh read of Tasks 7-11).
  Every full gate is heavy (Chromium, about 3-4 min, about 560 MiB).
- **6b-ii-b (restore) runs in lane 2** (worktree `lane-2-6b-ii-b`), segment B1 (Tasks 4A-4C) accepted (run
  `wf_a047da45-f4b`); B2 (5A, 5B, 6) running as `wf_a360199a-016`; B3 (7, close, merge) follows. If this session ended mid-segment: journal first, then relaunch (memory
  `pass-execute-cross-session-resume`).
- **Approved and waiting for a lane:** the transliteration-at-setup pass (`plans/2026-09-25-translit-table-at-setup.md`,
  3.3M, `unidecode install`; keeps dubplate MIT).

**Resume prompt:** read this file and memory `session-2026-09-25-handoff`, then run 5c-iii's review round
and close in lane 1, and continue 6b-ii-b in lane 2 from its journal. Box read-only; nothing deploys until 5d.

Readings taken 2026-09-25 (conductor; reversible), in full in HISTORY at each pass's close: the light gate lane
until 5c-iii Task 3a; G9 keeps DL R2.6's direct-text check (ratified standard over an ambiguous plan line);
two conductor-ruled second fix rounds (5c-iii Tasks 2 and 5); 3b's link-order mutation replaced by
"branding inside `@layer base`"; Task 4's icon mutation strengthened (SvelteKit rewrites icon links on
hydration); OpenSubsonic wire types hand-written (the survey found generators unfit). Geoff 2026-09-25:
Navidrome sharing stays at its default (zero shares) until 0.8; the Navidrome 0.64 bump for GHSA-82gh is
an open question. The visual identity spec is at revision 1.3 (its three errata landed). Owed: the streamer
spec's seven open items (revision 2.5, "Open items") go to its study rung. From the project site spec: settle
`init`/`upgrade` naming against `dubplate-server`, and add the configuration-reference generator as a consumer
of hook 18.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | OPEN: 5c-iii segments A and B accepted (3a accepted; 3b `591fe53`+`3db2708`); every full gate now runs heavy (Chromium; about 179s, about 560 MiB peak); segment C (Tasks 4, 5) next |
| 2 | worktree `lane-2-6b-ii-b` (branch `lane-2-6b-ii-b`, cut at master after `6528321`) | OPEN: 6b-ii-b restore (Geoff, 2026-09-25: extra budget beyond the window cap, about 9M with the two document lanes); ceiling 7.53M, flag 6.02M; heavy gates queue behind lane 1 |
| photos | main tree (master), document lane | OPEN: photo curation round one, `plans/2026-09-25-photo-curation-round-one.md`; writes only `photos/`, `tools/photos/`, and its research files; never runs `scripts/check.sh`; ceiling 5.5M (flag 4.4M; raised by Geoff 2026-09-25) |

Geoff, 2026-09-25 (attended): beyond the 14.5M window, parallel work runs uncapped for about four hours (6b-ii-b in lane 2, the errata document lane, the transliteration plan). Window 2 spend against the 14.5M cap (raised from 11M by Geoff, 2026-09-25, to run 5c-iii through its close; stop-don't-ask flag 13.4M): about 11.0M at segment B's boundary (incl. about 1.3M of attended extras: two prior-art surveys, the group-scale principles spec and its review) (matcher pass about 1.25M against 2.75M; 6b-ii-a
about 1.95M against 3.90M, including five merge-side architecture reads split across both passes).
Attended interactions after approval: 0 for both passes. AC online at launch.

## Open items (carried)

- No deploy until 5d: `magic-link` env template lacks its nine keys. Owed: `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys' `doctor` row, Ruling 6's old-browser `Referrer-Policy` residual, the branding-in-both-schemes line, `deploy.sh`'s clean-`web/`-build requirement, 6a's and 6b-i's runbook rows (`KillMode=mixed`, rclone v1.74.3+, crypt remote, exit-7 remedy).
- To rung 7: guard counting, `--max-delete`, a remote-pruning regression floor, archive uploads. From the project site spec: an enumerable reject-reason set with one canonical wording per parameterized reason, and the CREDITS output (with pack grantors' name-only lines) and picked images at stable paths an export manifest can list. To rung 9: library filter, facets, trigram search.
- To rung 12: no non-admin maps onto a streamer admin; a restore before then loses stars/playlists; roster API/Navidrome-fallthrough shared origin (security M2, a threat-model entry, Geoff 2026-09-25); Ruling 7's availability question; re-validate every roster row's `streamer_username`; the header step's fallthrough reconciliation; the streamer database's place in a restore (0.8 forward hook). To the revoke verb: the epoch store's first writer.
- To 5c-iii: the handoff at `docs/superpowers/research/handoffs/draft-HISTORY-5c-ii.md:368`, the label-ink gate row (a11y W4), WCAG 1.4.11 on night-primary fills, S11, S12, I2.
- To 6b-ii-b: `ServeOptions.Reconcile`/`backup.Reconcile` (6d's wiring), a store identity for restore to refuse another store's snapshot, the network actor's `member:<principal>` form, R-6bii9 keyed on `IntegrationEnabled`; the lease/staging/remote-join rules a restore must respect; the backup reference a restore replaces; backup's rclone-launch/copy-classification duplication (incl. `cmd/dubplate-server/backup.go`'s second remote-name validator).
- To rung 14 (the mount): the projection's mount without a trailing slash on `StripPrefix`, the production authorizer (`(Actor, Privilege, error)`, never a zero or unchecked grant), the bearer channel, Ruling 3's answer (403 forbidden for a member lacking privilege; a no-identity caller is rung 14's own denial kind), security L5-L7 (audit logging with actor and `client_ip`, `CancelJob`'s discarded actor / durable id in `jobs.actor`, the unauthenticated response shape), the M2 shared-origin question extended to the projection, `NetworkActor` exported ahead of its caller, and the authorizer-implementability forward note (unexported method, so only a type inside `opshttp` can implement it).
- To the architecture cleanup pass: the 5c-ii reads (`internal/auth`, `internal/serve`, `internal/identity`), 6b-i's ops/local findings, 6a's cmd findings, and this window's five reads (`internal/ops`, `internal/ops/local`, `internal/ops/opshttp`, `internal/engine/matcher`, `internal/harness/conform`), leading the queue: `conform`'s dropped `ModeDefault` near misses, and matcher's false `Distance.Update` doc plus the drifted `quiet_fallback`/`none_rec_action` legality checks. To 6d: a store identity minted at create, the migration-wait key, serve's caller. To 6c: the `SAVEPOINT` check on `COMMIT;BEGIN;` (risk R13). L5/L6 (roster-POST throttle, n5 whole-roster doctor row): no pass assigned.
- A later Go toolchain bump re-runs the matcher's Unicode class sweep (`wantGoOnlyWordCodePoints`, `strdist_test.go:452`) and the U+0130 case-mapping sweep; both fail loudly with the re-read procedure if the count moves.
- 99 declared coverage gaps, none blocking. Staged legs warn: 326 process citations, 125 exported-unused. Fork B: 0.8/0.9 FLAC-only, group non-FLAC carried.
- Queued for the batch after lane 1 closes (photo lane): a `scripts/check.sh` line running `node tools/photos/check.mjs integrity`, and the 5c-iii plan's Ruling 3 erratum routing the photo column to rung 7.
- Photo lane, before the first release that embeds photos: recheck every committed source URL, and decide whether withdrawn images' blobs may stay in git history once the repo is public (plan fold N2).
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
