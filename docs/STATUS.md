# dubplate status

## Current state

Rungs 1b, 3, 4, 8, 5a through 5c-iv, rung 6 through 6c, the transliteration pass, and 5d-i (2026-10-05:
the trusted-proxy class, doctor's four boundary rows, the probe route, `deploy.sh` check-by-default,
the serve unit, the env template's magic-link rows) closed; rung 5 continues at 5d-ii. HISTORY's top
entries: 5d-i, transliteration, 5c-iv.

## Next action

**Resume:** Read this file. 5d-ii Tasks 7 to 10 are done: Segment C closed at `8108605`, and the
attended first-deploy sitting deployed `ad094f9` to the box on 2026-10-05 (see "Deployed" under Open
items). Next: **Task 11, the rung 5 close**, in a fresh session (plan rev 2.2, "## Task 11"), then the
brainstorm that re-plans rungs 12 and 16 under the 2026-10-05 Navidrome and Cloudflare ruling.

Task 11 owes: the once-per-pass simplifier over 5d-ii's Go (9B, 9C), the architecture reads for
`internal/serve`, `cmd/dubplate-server`, and `internal/notify`, the 5d-ii HISTORY entry, the
`CLAUDE.md` gate-description edit (state-dir leg now covers every `dubplate-server` unit and
`DUBPLATE_SERVER_LEAVES`; `systemd-analyze` hard-required), and both budgets. Box read-only until the
sitting.

## Owner questions

- 6d plan rev 2.2 (`plans/2026-09-26-rung-6d-identity-serve-store.md`) awaits approval: three
  rulings (one pass at 11.5M; the epoch to the revoke verb; store identity to rung 16), all
  recommended yes. Reviewed by three lenses, verified, folded twice, prose-reviewed.
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.

- **Owner ruling (Geoff, 2026-10-05):** the box now only serves dubplate's development. Navidrome, all
  the Cloudflare infrastructure (the tunnel, Access), and the box's beets (the incumbent importer) may
  be removed whenever the conductor likes. Not yet planned: it reshapes rung 12 (the Navidrome
  driver), rung 16 (the cutover with Navidrome streaming), the 0.8 "escape hatch", and the shadow
  program. Geoff also ruled the matcher validated: the shadow evidence floor and its adjudication
  batches no longer gate launch, so the box's beets importer can go with no evidence cost; the corpus
  and conformance tests stay the matcher's gate, and the workstation's pinned beets oracle under
  `tools/venv` stays for derivation. The next brainstorm re-plans them. Until then the box serves through the tunnel. The streamer username
  (`glw907`, Geoff's) maps to nothing until then.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | IDLE: 5d-i closed; 5d-ii through Task 10 done (deployed `ad094f9`, 2026-10-05). Next: Task 11, the rung 5 close |
| 2 | none | 6c-fix merged at `0e3272c`..`8950de7`; worktree `lane-2-6c` removed |
| photos | main tree (master), owned by the separate docs session | ROUND 1 CLOSED at `c10c0f2` (116 picks, ~13M against a raised 7.5M ceiling); round two unplanned: targets in `research/photo-curation-rounds/round-1.md` (Round close) and `artist-sweep-list.md`; one Run-D.M.C. frame awaits Geoff's call |

## Open items (carried)

- **Owed:** the streamer spec's seven open items (rev 2.5) to its study rung; from the site spec,
  settle `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference
  generator as a hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- **Deployed 2026-10-05 (5d-ii Task 10 sitting, attended):** binary `ad094f9` (`pre-greenfield-985`)
  replaces `be27434`; `server.env` carries the magic-link rows (21 variables); the `auth`, `roster`
  (0700) and `unidecode` (0755) leaves; the table installed; the serve unit installed and disabled,
  started once under supervision (doctor exit 0, `boundary probe` refused), then stopped; beets 2.13.1
  to 2.14.1; cloudflared recreated at 2026.9.3; one admin member. The rollback set stays on the box
  until the next attended deploy replaces it: the `.prev` binary, `server.env`, `env`, contributors,
  tmpfiles and beets config, `/opt/dubplate.prev`, and `/root/dubplate-units.prev` (runbook "Roll
  back"); a binary rollback returns beets to 2.13.1 with it. The next deploy runs `deploy.sh --apply`
  from a fresh worktree after a `--check` read. The two HMAC keys are in the age store and registry
  (dotfiles `c463de1`).
- To Task 11: the plan's step 7 timing gate almost never qualifies (the shadow timer's 5-minute
  cadence lands after the 10-minute floor); the sitting ran one manual shadow pass after the import
  instead, which met the gate's intent. 9B's direct-mode warning fires on doctor's own probe while
  serve runs, a false alarm for a healthy doctor. A first `member list` exited 1 with stderr discarded;
  the identical rerun exited 0.
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason set,
  CREDITS, picked images at stable paths an export manifest can list, 6c's `member_quota` reservations
  and default-collection preference, 5c-iv's handoff (HISTORY 5c-iv "To rung 7", the styled 404 among
  it). To rung 9: filter, facets, trigram. To rung 12: `FrontExemption`'s dual-stack raw-string peer match (rebuild on `TrustedProxyConfig.trusts`; `research/2026-10-05-rung-5d-i-architecture-reads.md`), streamer-admin mapping, restore-before-12
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
