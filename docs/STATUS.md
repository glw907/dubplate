# dubplate status

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-i and 5c-ii, and rung 6's 6a, 6b-i, and **6b-ii-a** are closed. 6b-ii-a (the ops seam's HTTP projection, `6e6fd12`, about 1.95M against 3.90M) and the **matcher word-boundary pass** (the featuring row's ASCII RE2 rewrite, U+0130 declared and pinned, about 1.25M against 2.75M) both closed 2026-09-25; HISTORY, top two entries. Beets 2.14.1 adopted. The visual identity spec (revision 1.2, `782d864`) is complete, 5c-iii's standing input.

## Next action

- **5c-iii is in progress in lane 1.** Task 0 launched at `6e6fd12`: no blockers, no pin bumps
  owed, Playwright chromium-1243 present. Segment A (Tasks 1, 2) is now running.
  `plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`: split into 5c-iii (10.8M) then
  5c-iv (8.5M). From Task 3 on, every full gate launches Chromium (heavy lane).
- **6b-ii-b (restore) is next for a code lane.** `plans/2026-09-25-rung-6b-ii-restore-http.md`
  split B1 to B3; ceiling 7.53M. Handoff in HISTORY's 6b-ii-a entry: R-6bii9's disabled-class row
  keys on `IntegrationEnabled`, not `IntegrationValue`'s emptiness.

**Resume prompt:** execute `docs/superpowers/plans/2026-09-25-unattended-window-2.md` (Geoff's grants
inside: 5c-iii segments continue in lane 1 through its Task 6 close; 6b-ii-b does not start in
window 2; cap raised to 14.5M, flag 13.4M (Geoff, 2026-09-25, attended); read-only box).
Read memory `session-2026-09-25-handoff`. Rung 12 owns the roster API's shared origin
with the Navidrome fallthrough (5c-ii security M2) as a threat-model entry (Geoff, 2026-09-25).

Readings taken (conductor, 2026-09-25; reversible): one more targeted round on 5c-ii Fold A after
its second `fix`; the plans drafted alongside the fold; the architecture reads' routing; the 6b-i
CLAUDE.md Gate line without its false "one validator" clause; the light gate lane on both window-2
lanes until a browser leg exists; L3 (job `FailureMessage` text, admin-only) refused; the reviewer
note flagging the matcher corpus case's non-hex mbid refused; mutation 4 on 6b-ii-a Task 1
recorded reviewer-read (architecturally unfireable today). Owed: the visual identity spec's
errata (Permanent Marker is Apache-2.0; R5.4 cites read R5.5; the header focus halo). Owed to the streamer (0.8) spec's next revision: the prior-art survey's eight errata (`research/2026-09-25-streamer-prior-art-survey.md`: the XML/XSD rule, the `transcoding` extension, ranges on an in-flight transcode, CORS, AAC container, short trigram queries, the "collections" name clash, stale table pins).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | OPEN: 5c-iii Task 0 launched (`6e6fd12`), segment A (Tasks 1, 2) running; ceiling 10.8M |
| 2 | (free) | CLOSED: 6b-ii-a merged at `6e6fd12`; 6b-ii-b next |

Window 2 spend against the 14.5M cap (raised from 11M by Geoff, 2026-09-25, to run 5c-iii through its close; stop-don't-ask flag 13.4M): about 4.9M (matcher pass about 1.25M against 2.75M; 6b-ii-a
about 1.95M against 3.90M, including five merge-side architecture reads split across both passes).
Attended interactions after approval: 0 for both passes. AC online at launch.

## Open items (carried)

- No deploy until 5d: `magic-link` env template lacks its nine keys. Owed: `doctor` exits 0, `ReadWritePaths`, the trusted proxy, the theme keys' `doctor` row, Ruling 6's old-browser `Referrer-Policy` residual, the branding-in-both-schemes line, `deploy.sh`'s clean-`web/`-build requirement, 6a's and 6b-i's runbook rows (`KillMode=mixed`, rclone v1.74.3+, crypt remote, exit-7 remedy).
- To rung 7: guard counting, `--max-delete`, a remote-pruning regression floor, archive uploads. To rung 9: library filter, facets, trigram search.
- To rung 12: no non-admin maps onto a streamer admin; a restore before then loses stars/playlists; roster API/Navidrome-fallthrough shared origin (security M2, a threat-model entry, Geoff 2026-09-25); Ruling 7's availability question; re-validate every roster row's `streamer_username`; the header step's fallthrough reconciliation; the streamer database's place in a restore (0.8 forward hook). To the revoke verb: the epoch store's first writer.
- To 5c-iii: the handoff at `docs/superpowers/research/handoffs/draft-HISTORY-5c-ii.md:368`, the label-ink gate row (a11y W4), WCAG 1.4.11 on night-primary fills, S11, S12, I2.
- To 6b-ii-b: `ServeOptions.Reconcile`/`backup.Reconcile` (6d's wiring), a store identity for restore to refuse another store's snapshot, the network actor's `member:<principal>` form, R-6bii9 keyed on `IntegrationEnabled`; the lease/staging/remote-join rules a restore must respect; the backup reference a restore replaces; backup's rclone-launch/copy-classification duplication (incl. `cmd/dubplate-server/backup.go`'s second remote-name validator).
- To rung 14 (the mount): the projection's mount without a trailing slash on `StripPrefix`, the production authorizer (`(Actor, Privilege, error)`, never a zero or unchecked grant), the bearer channel, Ruling 3's answer (403 forbidden for a member lacking privilege; a no-identity caller is rung 14's own denial kind), security L5-L7 (audit logging with actor and `client_ip`, `CancelJob`'s discarded actor / durable id in `jobs.actor`, the unauthenticated response shape), the M2 shared-origin question extended to the projection, `NetworkActor` exported ahead of its caller, and the authorizer-implementability forward note (unexported method, so only a type inside `opshttp` can implement it).
- To the architecture cleanup pass: the 5c-ii reads (`internal/auth`, `internal/serve`, `internal/identity`), 6b-i's ops/local findings, 6a's cmd findings, and this window's five reads (`internal/ops`, `internal/ops/local`, `internal/ops/opshttp`, `internal/engine/matcher`, `internal/harness/conform`), leading the queue: `conform`'s dropped `ModeDefault` near misses, and matcher's false `Distance.Update` doc plus the drifted `quiet_fallback`/`none_rec_action` legality checks. To 6d: a store identity minted at create, the migration-wait key, serve's caller. To 6c: the `SAVEPOINT` check on `COMMIT;BEGIN;` (risk R13). L5/L6 (roster-POST throttle, n5 whole-roster doctor row): no pass assigned.
- A later Go toolchain bump re-runs the matcher's Unicode class sweep (`wantGoOnlyWordCodePoints`, `strdist_test.go:452`) and the U+0130 case-mapping sweep; both fail loudly with the re-read procedure if the count moves.
- 99 declared coverage gaps, none blocking. Staged legs warn: 326 process citations, 125 exported-unused. Fork B: 0.8/0.9 FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
