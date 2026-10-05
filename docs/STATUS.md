# dubplate status

## Current state

Rung 5 is closed (2026-10-05): rungs 1b, 3, 4, 8, 5a through 5d-ii, rung 6 through 6c, and the
transliteration pass are done. The box runs only dubplate-server (deployed record below). HISTORY's top
entries: 5d-ii (with the rung 5 summary), 5d-i, transliteration.

## Next action

**Resume:** Read this file. Lane 1 is idle and the tree is clean after the rung 5 close commit. Two
things wait, neither blocking the other:

1. **The brainstorm that re-plans rungs 12 and 16** (and the 0.8 escape hatch and the shadow program)
   under the 2026-10-05 ruling. Inputs: the ruling below; `research/2026-10-05-rung-5d-ii-architecture-reads.md`
   (the `internal/serve` forward-hook removals: `FrontExemption`, `SetFallthrough`, the driver header,
   the bearer bypass); HISTORY 5d-ii's "wrong to rediscover" (which carries became moot and which stay
   live). Rung 12's Navidrome driver may no longer exist; decide that first.
2. **6d plan approval** (`plans/2026-09-26-rung-6d-identity-serve-store.md`, rev 2.2), under "Owner questions".

Executor conditions: fresh session, `dubplate-implementer` per task, push master after each close
(gitleaks first), heavy-lane full gate. Box writes only through `deploy.sh --apply` from a fresh worktree
after a `--check` read, attended.

## Owner questions

- 6d plan rev 2.2 awaits approval: three rulings (one pass at 11.5M; the epoch to the revoke verb; store
  identity to rung 16), all recommended yes. Reviewed by three lenses, verified, folded twice,
  prose-reviewed.
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- **Owner ruling (Geoff, 2026-10-05):** the box serves only dubplate's development. Navidrome,
  cloudflared and the tunnel, Cloudflare Access, the box's beets, the shadow program's box side, Docker,
  and the incumbent `dubplate` binary and units are retired (done the same day). The matcher is
  validated: the shadow evidence floor and its adjudication batches no longer gate launch; the corpus
  and conformance tests stay the matcher's gate, and the workstation's pinned beets oracle under
  `tools/venv` stays for derivation. Rungs 12 and 16, the 0.8 escape hatch, and the shadow program are
  unplanned under it. The streamer username (`glw907`) maps to nothing until then.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | IDLE: rung 5 closed (5d-ii, HISTORY). `internal/serve` and `internal/auth` released to rung 12's ownership; rung 12 is being re-planned |
| 2 | none | 6c-fix merged at `0e3272c`..`8950de7`; worktree `lane-2-6c` removed |
| photos | main tree (master), owned by the separate docs session | ROUND 1 CLOSED at `c10c0f2` (116 picks, ~13M against a raised 7.5M ceiling); round two unplanned: targets in `research/photo-curation-rounds/round-1.md` and `artist-sweep-list.md`; one Run-D.M.C. frame awaits Geoff's call |

## Open items (carried)

- **Deployed 2026-10-05** (Task 10 sitting, then the retirement): binary `d1e93e9` (doctor exit 0), the
  15-row `server.env`, the serve unit (installed, disabled), the `auth`, `roster`, `unidecode` leaves,
  one admin member, `/srv/music/library` (6.6 GB). No rollback set remains. History archived at
  `/root/musicbox-retired-2026-10-05.tar.gz` (0600). The retired stack's external pieces are gone (tunnel,
  `music.907.life`, four pings checks, 42 age-store names). `.claude/settings.json` keeps two dead
  `docker compose` permission entries for Geoff to drop.
- **To the rungs 12 and 16 re-plan:** moot under the ruling: the cloudflared address pin and
  `compose.yaml` trusted-proxy values, serve bound to the bridge gateway, the Navidrome verify endpoint
  (R-5d17) and its forward-auth doctor proof, the shadow timer, the 0.8 escape hatch. Still live: enabling
  the serve unit, the secrets in every `music`-uid verb's environment, exercising SMTP from the box, 6c's
  Ruling 2 (`deploy.sh` halting on `check`'s nonzero exit before restarting `serve`), stripping group 2
  names but the configured header from a trusted peer, UNI-H2 (`/` to login) and group 2 in trailers.
  Also the `internal/serve` hook removals in the 5d-ii reads and `FrontExemption`'s dual-stack match.
- **Owed:** the streamer spec's seven open items (rev 2.5) to its study rung; from the site spec, settle
  `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference generator as a
  hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason set,
  CREDITS, picked images at stable paths an export manifest can list, 6c's `member_quota` reservations
  and default-collection preference, 5c-iv's handoff (HISTORY 5c-iv "To rung 7", the styled 404 among
  it). To rung 9: filter, facets, trigram. To rung 12: streamer-admin mapping, restore-before-12 losing
  stars/playlists, shared-origin M2 and 5c-iii's CSP-scope M2, the streamer database's restore place;
  full lists HISTORY 5c-ii, 5c-iii, 5c-iv. To the revoke verb: the epoch store's first writer. To rung
  14: 6c's admin write path and a resolving channel for a `serve`-refusing `admin` row.
- To 6d: `ServeOptions.Reconcile` wiring, store identity, migration-wait key, `member:<principal>` actor,
  6c's handoff (HISTORY 6c). Before any deploy exposing the strings: wire settings `apply` at `serve`
  start (`cmd/dubplate-server/backup.go:274`, `store.go:280-284` claim it) and open the store through
  `OpenServe` (unwired today). To 6d Task 4: `rosterResetServer`'s baseline (HISTORY 5c-iv).
- Errata and architecture-cleanup reads, file:line in HISTORY and research: 5c-iii's nine errata;
  5c-iv-E1 to E3; 6b-ii-b's; 6c-E1 to 6c-E10; the 5c-ii, 6b-i, 6a, window 2, 5c-iii, 6b-ii-b, 6c reads;
  5c-iv's routings (`research/2026-09-26-rung-5c-iv-final-reads.md`); transliteration's ROUTE items
  (`research/2026-09-27-translit-architecture-reads.md`) and the four-copy `mainTB` shim; matcher re-run
  trigger `strdist_test.go:452`; the 5d reads (`research/2026-10-05-rung-5d-i-architecture-reads.md`,
  `-5d-ii-architecture-reads.md`: doctor's twelve-flag OR chain first). 103 coverage gaps, none blocking.
  Staged legs (gate at the rung 5 close): 317 process citations, 127 exported-unused; the citation leg misses
  `R-6c` patterns. BACKLOG #11 (the docker leg) is retired with the leg.
- To `give-back` and `packaging`: the deleted table's git history, the process-wide holder's tension,
  the Unidecode-derived items (HISTORY transliteration, "Routed"), UNI-L2's failure-alert hook, the
  loopback-relay mail class (`IDEAS.md`), the token-in-fragment move (BACKLOG).
- Workstation prerequisite: `make unidecode-wheel` once per machine before `make test`; prove "wheel
  absent" with `-count=1`. A gate failing on a `/tmp` disk-quota write is a lost run.
- Photo lane, after lane 1: `tools/photos/check.test.mjs` fails 2 of 121 (both assume an empty
  `photos/`), so no gate runs it; the integrity leg reads `photos/` from disk, not `git ls-files`;
  `check.sh`'s header omits the leg. Ruling 3's photo erratum to rung 7, a source-URL recheck, a
  blob-retention call. Pinned: beets `2.14.1` (workstation oracle), mediafile `0.17.0`, unidecode `1.4.0`
  (sha256 pins in `unidecode/pins.go`), lap `0.5.13`, sqlite v1.60.1, rclone v1.74.3.
