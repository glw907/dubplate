# dubplate status

## Current state
Rung 5, the navidrome-removal and re-plan document passes, rung 6d, and rung 12 slices 1, 2a, and 2b are closed
(HISTORY). Slice 2b's code is on master through `be9abf5` (`e561c32..be9abf5`) and its ledgers landed in the commit
after it. Master's full gate is green at `313c7ff` (heavy lane, 2026-10-10); the closing unexport `8dbe59a` is test-only and passed the scoped gate. Nothing is paused and nothing deploys.

Slice 3 is ready to plan: it is the first slice with a dubplate side and consumes 2b's XML contract tier, the
machine-form ledger (52 owed rows, none active), the per-endpoint assignment, the differ with its encoding-quirk
rows, and the music-folders evidence (both oracles recorded over two collections, the `musicFolderId` filter graded).
The Tempus walk is deferred until `/dev/kvm` exists, and Amperfy's walk is rung 16's attended gate.
The box is unchanged: dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`, `roster`,
`unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB); the first 6d deploy takes the runbook's 6d
upgrade. Archive: HISTORY, "Archive record".

## Next action
> **Resume prompt (start a fresh brainstorm-and-plan session in /var/home/glw907/Projects/dubplate; `claude --model
> claude-opus-5-5`, effort high)**
> **Goal.** Start from `docs/superpowers/research/2026-10-10-next-pass-planning-brief.md`: it recommends rung 14
> (slice 3 waits on rung 9's catalog), lists three questions for Geoff (rung 9's concurrency ruling first), and
> proposes the scope, decisions, segments, and ceiling. Ask the questions, then spec and plan rung 14.
> **Settled (do not re-brainstorm):** rung 12
> slices 1 to 2b are closed; forks A, B, C are ruled (streamer spec, "Revision note, 2026-10-10"); the lane registry
> below. **Approach:** invoke `pass-core`; open the pass with the `dependency-upgrade` sweep, which now covers libxml2
> (held by `brew pin libxml2`); write the plan to `docs/superpowers/plans/`, run it through `spec-plan-review`, and take
> it to Geoff for approval before any executor launches. This closed session does not draft the plan.

## Owner questions
- R-6d21 overrides 6c's "never reads environ" rule; Geoff may reopen it (not blocking).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- Workstation: ~41 stale `DUBPLATE_*` exports in `~/.local/secrets` fail `check` outside `env -i`; prune them.
- VT-x appears disabled in firmware (no `/dev/kvm`): enabling it unblocks the Android emulator and the Tempus walk.

## Lane registry
| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | idle: 6d closed 2026-10-07 (code through `4de73e7`, close commit `0aa8d18`) |
| 3 (slice 2) | `.claude/worktrees/lane-3-slice-2` | closed 2026-10-10; worktree removed; code on master through `be9abf5` |
| photos | main tree, docs session | round 1 closed at `c10c0f2`; one Run-D.M.C. frame awaits Geoff |

## Open items (carried; landings from the re-plan spec's "Where the carried items land")
- **Revoke verb and slice 6:** spec criteria 7 and 8 (the dual-key epoch, R-6d20; `EpochStore.Bump` has no caller).
  **Rung 14:** the last-admin guard with the removal verb; the removal verb bumps the address's revocation epoch as it
  tombstones; the network actor keyed on `UserID`; the admin write path's resolving channel for a `serve`-refusing
  `admin` row. **First migration after `0005`:** R-6d4's keep limit (a busy or intent start prunes to the default)
  and the older-store test helper's table map.
- **Rung 16:** enabling the serve unit; criterion 14's `runuser` and holder clauses with 6c's Ruling 2 (`deploy.sh`
  halts on `check` before restarting `serve`); secrets in every `music`-uid verb's environment; SMTP from the box;
  the `systemd-run` sandboxed `member add` and `doctor` (the 6d smoke did not drive them); the remaining runbook
  lines; the archive's second copy; per-collection relocation; store identity (R9's option b); Amperfy's roster walk.
  **Slice 3:** UNI-H2, the prefix-gap and percent-encoding items; the owed ledger rows L-20, L-21, and L-51.
  **Slice 5:** L-17. **Slice 6:** the streamer-admin mapping, the `duplicate_principal` code and BACKLOG #16,
  offboarding deprovision, CSP-scope and shared-origin M2, username uniqueness. **Slice 7:** the streamer database in
  a restore. **Slice 8:** group 2 in trailers, `CAP_NET_BIND_SERVICE`, command naming.
- **Lane 1's next `web/` pass:** R-6d17's note and BACKLOG #23 (HISTORY, 6d, "Routed items"). **Bugs and chores:** the
  bearer bypass's first consumer owes a cookie test (HISTORY); BACKLOG #15, #16, #21, #23-#26, #28, #30-#33; two dead
  `docker compose` entries in `.claude/settings.json` for Geoff to drop. **Owed:** settle `init`/`upgrade` naming
  against `dubplate-server`; the configuration-reference generator as a hook 18 consumer.
- **Rung 7, rung 9,** errata inventory, `give-back`, `packaging`, photo lane: HISTORY, slice 1 close, "Moved out of STATUS".
