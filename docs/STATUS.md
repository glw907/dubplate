# dubplate status

## Current state

Rung 5 is closed (2026-10-05): rungs 1b, 3, 4, 8, 5a through 5d-ii, rung 6 through 6c, and the
transliteration pass are done. Master is green at `cea546e`. The box runs only dubplate-server: binary
`d1e93e9` (doctor exit 0), the serve unit installed and disabled, the `auth`, `roster`, `unidecode` leaves,
one admin member, and `/srv/music/library` (6.6 GB). No rollback set remains: the `.prev` copies rolled back
to the retired stack, so the next `--apply` writes the first fresh set. The retirement record is in HISTORY 5d-ii. HISTORY's
top entries: 5d-ii (with the rung 5 summary), 5d-i, transliteration.

## Next action

> **Resume prompt (launch in /var/home/glw907/Projects/dubplate; `claude --model claude-opus-5-5`, effort high)**
>
> **Goal:** re-plan rungs 12 and 16, the 0.8 escape hatch, and the shadow program under the 2026-10-05
> ruling (the box serves only dubplate's development; the incumbent stack is retired; the matcher is validated).
>
> **Scope:** in: the shape of rung 12 (streamer driver, SSO, `internal/serve`'s forward hooks), rung 16 (the
> cutover), the 0.8 escape hatch, and what remains of the shadow program (code in `cmd/dubplate-server/shadow.go`).
> Out: rung 6d (its plan awaits Geoff's approval separately), rung 7 onward, any code change.
>
> **Settled, do not re-brainstorm:** the ruling above; the moot list (the cloudflared address pin and
> `compose.yaml` values, serve bound to the bridge gateway, the Navidrome verify endpoint R-5d17 and its
> forward-auth doctor proof, the shadow timer and its step 7 gate); the matcher is validated, so the shadow
> floor and adjudication batches do not gate launch (corpus and conformance stay its gate); the refused
> features in CLAUDE.md.
>
> **Still open, brainstorm these:** whether rung 12's Navidrome driver survives at all; what the cutover becomes
> with no incumbent; whether `FrontExemption`, `SetFallthrough`, the driver header, and the bearer bypass
> stay in `internal/serve` or go; what the streamer-username term and mapping mean with no Navidrome
> (owner question below); the live carries listed under Open items; whether the shadow code is deleted.
>
> **Approach:** `superpowers:brainstorming`, then `superpowers:writing-plans`. Inputs: HISTORY 5d-ii ("wrong to
> rediscover"), `research/2026-10-05-rung-5d-ii-architecture-reads.md` and `-5d-i-architecture-reads.md`, the
> streamer 0.8 spec (`docs/superpowers/specs/`), and the Open items below. Bring one design for approval.

## Owner questions

- **Two plans await Geoff's approval** (spec `specs/2026-10-05-rungs-12-16-replan-design.md`, approved with
  rulings 1-3 yes): `plans/2026-10-05-navidrome-removal.md` (code, lane 1, rev 2, three-lens review folded and
  verified) and `plans/2026-10-05-replan-document-pass.md` (docs, rev 2, self-folded lens). Its Task 0 archive
  chore runs before any other box work. 6d rev 2.3 follows the removal pass; slice 1 (doc lane) and slice 2
  (third code lane, worktree) open per spec ruling 5.

- 6d plan rev 2.2 awaits approval (`plans/2026-09-26-rung-6d-identity-serve-store.md`): three rulings (one
  pass at 11.5M; the epoch to the revoke verb; store identity to rung 16), all recommended yes.
- "Streamer username": a term admins know before 0.8? (`rung-5c-iv-register-review.md` item 23).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | IDLE: rung 5 closed. `internal/serve` and `internal/auth` released to rung 12's ownership; rung 12 is being re-planned |
| photos | main tree, separate docs session | round 1 closed at `c10c0f2`; one Run-D.M.C. frame awaits Geoff's call |

## Open items (carried)

- **To the rungs 12 and 16 re-plan, still live:** enabling the serve unit, the secrets in every `music`-uid
  verb's environment, exercising SMTP from the box, 6c's Ruling 2 (`deploy.sh` halting on `check`'s nonzero
  exit before restarting `serve`), 6b-ii-b's runbook lines (HISTORY 6b-ii-b), stripping group 2 names but the
  configured header from a trusted peer, UNI-H2 (`/` to login), group 2 in trailers, `FrontExemption`'s
  dual-stack match. Moot items: HISTORY 5d-ii.
- **Bugs and chores:** BACKLOG #15 (serve's direct-mode warning fires on doctor's own probe); two dead
  `docker compose` entries in `.claude/settings.json` for Geoff to drop.
- **Owed:** the streamer spec's seven open items (rev 2.5) to its study rung; from the site spec, settle
  `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference generator as a hook
  18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- To rung 7: guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason set, CREDITS,
  picked images at stable paths an export manifest can list, 6c's `member_quota` reservations and
  default-collection preference, 5c-iv's handoff (HISTORY 5c-iv "To rung 7"). To rung 9: filter, facets,
  trigram. To rung 12: streamer-admin mapping, restore-before-12 losing stars/playlists, shared-origin M2 and
  5c-iii's CSP-scope M2, the streamer database's restore place; full lists HISTORY 5c-ii, 5c-iii, 5c-iv. To the
  revoke verb: the epoch store's first writer. To rung 14: 6c's admin write path and a resolving channel for
  a `serve`-refusing `admin` row.
- To 6d: `ServeOptions.Reconcile` wiring, store identity, migration-wait key, `member:<principal>` actor, 6c's
  handoff (HISTORY 6c). Before any deploy exposing the strings: wire settings `apply` at `serve` start
  (`cmd/dubplate-server/backup.go:274`, `store.go:280-284` claim it) and open the store through `OpenServe`
  (unwired today). To 6d Task 4: `rosterResetServer`'s baseline (HISTORY 5c-iv).
- Errata and architecture-cleanup inventory: file:line in HISTORY and `research/*-architecture-reads.md`
  (5d-ii's doctor twelve-flag OR chain first); 9b's five nits (HISTORY 5c-iv); 103 coverage gaps, none
  blocking; staged legs 317/127 (the citation leg misses `R-6c` patterns).
- To `give-back` and `packaging`: the deleted table's git history, the process-wide holder's tension, the
  Unidecode-derived items (HISTORY transliteration, "Routed"), UNI-L2's failure-alert hook, the loopback-relay
  mail class (`IDEAS.md`), the token-in-fragment move (BACKLOG #13).
- Photo lane, after lane 1: `tools/photos/check.test.mjs` fails 2 of 121 (both assume an empty `photos/`), so
  no gate runs it; the integrity leg reads `photos/` from disk, not `git ls-files`; `check.sh`'s header omits
  the leg. Ruling 3's photo erratum to rung 7, a source-URL recheck, a blob-retention call.
