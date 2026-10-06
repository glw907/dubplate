# dubplate status

## Current state

Rung 5 is closed (2026-10-05). The navidrome-removal and re-plan document passes are closed (HISTORY's top two
entries): the shadow program, front exemption, fallthrough, and streamer and front keys are gone, the roster,
store, API, CLI, and members page say `username`, and the specs, `CLAUDE.md`, `ROADMAP.md`, and `IDEAS.md` carry
the rungs 12 and 16 re-plan. Master is green. The box runs only dubplate-server: binary `d1e93e9` (doctor exit
0), the serve unit installed and disabled, the `auth`, `roster`, `unidecode` leaves, one admin member, and
`/srv/music/library` (6.6 GB). No rollback set remains. The first roster-reading deploy takes the runbook's
re-create step. The retired musicbox archive's copy is
`~/.local/share/dubplate/archive/musicbox-retired-2026-10-05.tar.gz.age` (HISTORY, "Archive record").

## Next action

> **Resume prompt (launch in /var/home/glw907/Projects/dubplate; `claude --model claude-opus-5-5`, effort medium)**
> **Goal:** 6d revision 2.3: rebase `plans/2026-09-26-rung-6d-identity-serve-store.md` onto the removal pass, run
> its own `spec-plan-review`, then bring it to Geoff for approval. **Then:** the slice 1 (document lane) and slice 2 (third code lane, own worktree) plans, just in time, per the
> re-plan spec (`specs/2026-10-05-rungs-12-16-replan-design.md`, rev 2.1) ruling 5. Never re-run a full gate on
> an unchanged commit (`~/.claude/docs/pass-gate-economy.md`).

## Owner questions
- 6d: rev 2.2's approval is superseded; rev 2.3 carries its rulings forward for approval.
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.

## Lane registry
| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | IDLE: removal pass and document pass closed 2026-10-06. Next: 6d rev 2.3 |
| slice 1 | document lane | not launched; plan unwritten |
| slice 2 | own worktree | not launched; plan unwritten |
| photos | main tree, docs session | round 1 closed at `c10c0f2`; one Run-D.M.C. frame awaits Geoff |

## Open items (carried; landings from the re-plan spec's "Where the carried items land")

- **Rung 16:** enabling the serve unit, the secrets in every `music`-uid verb's environment, SMTP exercised from
  the box, 6c's Ruling 2 (`deploy.sh` halts on `check`'s nonzero exit), 6b-ii-b's runbook lines (HISTORY
  6b-ii-b), the archive's second copy (the `.age` file to the backup remote), per-collection relocation, store
  identity if 6d's Ruling 3 is yes. **Slice 1:** API base URL measurement. **Slice 3:** UNI-H2 (`/` to login),
  the prefix-gap and percent-encoding items. **Slice 6:** the streamer-admin mapping, offboarding deprovision,
  CSP-scope and shared-origin M2, username principal uniqueness. **Slice 7:** the streamer database in a
  restore. **Slice 8:** group 2 in trailers, `CAP_NET_BIND_SERVICE`, command naming.
- **Bugs and chores:** the bearer bypass's first consumer owes a shell-level cookie test (HISTORY); BACKLOG #15
  (direct-mode warning fires on doctor's probe) and #16 (a username equal to another username-holder's
  address); two dead `docker compose` entries in `.claude/settings.json` for Geoff to drop.
- **Owed:** the streamer spec's open items to its study rung (slice 1); from the site spec, settle
  `init`/`upgrade` naming against `dubplate-server` and add the configuration-reference generator as a hook 18
  consumer. Reading taken: OpenSubsonic wire types hand-written.
- **To rung 7:** guard counting, `--max-delete`, archive uploads, 5c-iii's S13/a11y S7, reject-reason set,
  CREDITS, picked images at stable paths, 6c's `member_quota` reservations, 5c-iv's handoff (HISTORY 5c-iv "To
  rung 7"). **Rung 9:** filter, facets, trigram. **Revoke verb:** the epoch store's first writer. **Rung 14:**
  6c's admin write path and a resolving channel for a `serve`-refusing `admin` row.
- **To 6d:** `ServeOptions.Reconcile` wiring, store identity, migration-wait key, `member:<principal>` actor,
  6c's handoff (HISTORY 6c), re-validating existing rows' usernames. Before any deploy exposing the strings: wire
  settings `apply` at `serve` start (`cmd/dubplate-server/backup.go:274`, `store.go:280-284` claim it) and open
  the store through `OpenServe` (unwired today). 6d Task 4: `rosterResetServer`'s baseline (HISTORY 5c-iv).
- Errata and architecture-cleanup inventory: HISTORY and `research/*-architecture-reads.md` (5d-ii's doctor
  twelve-flag OR chain first); 9b's five nits (HISTORY 5c-iv); 103 coverage gaps, none blocking.
- To `give-back` and `packaging`: the carries listed in HISTORY, re-plan close, "Moved out of STATUS".
- Photo lane, after lane 1: three owed checks and a retest (HISTORY, re-plan close, "Moved out of STATUS").
