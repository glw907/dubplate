# dubplate status

## Current state

Rung 5 is closed (2026-10-05), and the navidrome-removal pass closed 2026-10-06 (HISTORY's top entry): the
shadow program, the front exemption, the fallthrough, and the streamer and front keys are gone, and the store's
`users` column and the roster, API, CLI, and members page say `username`. Master is green. The box runs only
dubplate-server: binary `d1e93e9` (doctor exit 0), the serve unit installed and disabled, the `auth`, `roster`,
`unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB); the removal pass deployed nothing.
No rollback set remains: the next `--apply` writes the first fresh set. The first roster-reading deploy takes
the runbook step (the roster has no compatibility shim). Reading 2 departs from the spec's literal default:
`ClassBackground` and `ErrUnavailable` stay, because no rung 9 spec names a consumer and the limiter's
three-class design and the client's retry contract need them; only their shadow justification was pruned.

## Next action

> **Resume prompt (launch in /var/home/glw907/Projects/dubplate; `claude --model claude-opus-5-5`, effort medium)**
>
> **Goal:** execute the document pass, `plans/2026-10-05-replan-document-pass.md` (rev 2; spec
> `specs/2026-10-05-rungs-12-16-replan-design.md`, rev 2.1). Segment A (Tasks 1-4) now, then segment B
> (Tasks 5-8), then Task 10, the close. The plan's header carries its ceiling, runner args
> (`gateLane: "heavy"`, `reviewerModel: "claude-opus-5-5"`), and conductor actions.
>
> **Run:** `pass-execute` by name, one invocation per segment, runaway guard armed, lane registry updated.
> Shared heavy-gate lock with the cairn-cms session: a heads-up via SendMessage before any heavy stretch over
> 15 minutes, and never re-run a full gate on an unchanged commit (`~/.claude/docs/pass-gate-economy.md`).
>
> **Then, not in that plan:** 6d revision 2.3 (rebased onto the removal pass, its own review and approval);
> then the slice 1 (document lane) and slice 2 (third code lane, own worktree) plans per spec ruling 5.

## Archive record (document pass Task 0, 2026-10-05)

`/root/musicbox-retired-2026-10-05.tar.gz` on the box (99,709,984 B, 0600 root, unchanged) has a verified copy at
`~/.local/share/dubplate/archive/musicbox-retired-2026-10-05.tar.gz.age` (0600, age to the store's `RECIP=`;
sha256 `de8f995d...ebec1`). Box and pulled sha256 match: `12b1d576...9044`, re-matched by the decrypted stream.
The tar holds one `srv/music/state/navidrome/navidrome.db` (1,810,432 B, with `-wal` and `-shm`): 157 schema rows,
`integrity_check` ok. R2 `music-library` also holds it: `state/navidrome.db` and 35 dated snapshots under
`_archive/state/` (2026-08-31 to 2026-10-05), which corrects HISTORY's "state/ empty". The full tar's only
off-box copy is the `.age` file; rung 16's backup remote takes a second. No plaintext remains.

## Owner questions

- 6d: rev 2.2's approval is superseded; rev 2.3 (after the removal pass, per the re-plan spec) carries its
  rulings forward for approval (`plans/2026-09-26-rung-6d-identity-serve-store.md`).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | IDLE: `navidrome-removal` closed 2026-10-06 (spend about 3.7M of 8.0M). Next: the document pass, segment A |
| photos | main tree, separate docs session | round 1 closed at `c10c0f2`; one Run-D.M.C. frame awaits Geoff's call |

## Open items (carried)

- **Handoffs from the Q&A session (Geoff, 2026-10-05), for document segment A:** commit
  `~/.local/share/dubplate/handoff/2026-10-05-client-roster-audit.md` to `docs/superpowers/research/` (the
  ruled client roster: Feishin, Rufin, Amperfy, Tempus with Navic alternate); point streamer spec rev 3's roster
  open item at it; add Melodee's intake pipeline comparison as a named slice 1 study question (one line in rev 3,
  full scope in the slice 1 plan).

- **To the rungs 12 and 16 re-plan, still live:** enabling the serve unit, the secrets in every `music`-uid
  verb's environment, exercising SMTP from the box, 6c's Ruling 2 (`deploy.sh` halting on `check`'s nonzero
  exit before restarting `serve`), 6b-ii-b's runbook lines (HISTORY 6b-ii-b), stripping group 2 names from
  a trusted peer, UNI-H2 (`/` to login), group 2 in trailers. Moot items: HISTORY 5d-ii; the removal closed
  `FrontExemption`'s dual-stack match, the configured-header carve-out, and the `SetFallthrough` items. The
  bearer bypass's first consumer owes a shell-level cookie test (HISTORY).
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
