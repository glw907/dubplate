# dubplate status

## Current state
Rung 5, the navidrome-removal and re-plan document passes, rung 6d, rung 12 slices 1, 2a, and 2b, and the gate scope
pass are closed (HISTORY). Master's full gate is green at `f0a2056`. Nothing is paused and nothing deploys.

The gate is classified: `check.sh --base`, fail-closed classes, `--dry-run`, the rendered leg chosen by reverse
dependency, and Go's cache sizing conform. The 0.9 home install scope and forward hooks 23 and 24 landed at `e833365`
(ROADMAP and the streamer spec).

Slice 3 waits on rung 9's catalog; it consumes 2b's XML contract tier, the machine-form ledger (52 owed rows, none
active), the per-endpoint assignment, the differ, and the music-folders evidence. The Tempus walk ran on 2026-10-10
(`1f4f232`, 11 of 11 flows, emulator needs `-gpu host`); Amperfy's walk is rung 16's attended gate.
The box is unchanged: dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`, `roster`,
`unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB); the first 6d deploy takes the runbook's 6d
upgrade. Archive: HISTORY, "Archive record".

## Next action
Rung 14a: the plan is `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md`, revision 2.1 as amended at
`f0a2056`. It awaits Geoff's approval of its task list and ceiling (11.8M, about 6.8 h, about 8.3 h with fix rounds).
After approval it launches in the worktree `.claude/worktrees/lane-1-rung-14a`, with the lane checklist run first.
Settled (do not re-brainstorm): rung 12 slices 1 to 2b are closed; forks A, B, C are ruled (streamer spec, "Revision
note, 2026-10-10").

## Owner questions
- Rung 7 spec revision 2.1 (below) awaits Geoff's read.
- R-6d21 overrides 6c's "never reads environ" rule; Geoff may reopen it (not blocking).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- Workstation: ~41 stale `DUBPLATE_*` exports in `~/.local/secrets` fail `check` outside `env -i`; prune them.

## Lane registry
| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | idle; next is rung 14a, awaiting plan approval, then the worktree `.claude/worktrees/lane-1-rung-14a` |
| rung 7 planning | docs only | spec revision 2.1 at `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md` (`102d973`) awaits Geoff's read; his seven rulings are written in; the plan is drafted after approval |
| 3 (slice 2) | `.claude/worktrees/lane-3-slice-2` | closed 2026-10-10; worktree removed |
| photos | main tree, docs session | round 1 closed at `c10c0f2`; one Run-D.M.C. frame awaits Geoff |

## Owed
- **After the rung 7 spec approval:** errata to eight documents, listed in
  `docs/superpowers/research/2026-10-10-rung-7-intake-fold.md` (SS hook 12 and the buy-link hook; CH; 14A; W8.2; VI;
  BS re-band to L; PCF; CLAUDE.md).
- **Now:** `docs/roster-walk.md` still pins Tempus 4.28.0 with the wrong F11 expectation (the walk used 4.29.0, and
  F11 used `createPlaylist(name, songId)`); the rung 14a spec's eight owed errata; the `id3-tags` copy must name a
  working strip method (`mid3v2 --delete-all` or a tagger), at the rung 7 plan's copy read; `make derive` has to run
  once before the next oracle-dependent rung (expect no diff).

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
  bearer bypass's first consumer owes a cookie test (HISTORY); BACKLOG #15, #16, #21, #23-#26, #28, #30-#34; two dead
  `docker compose` entries in `.claude/settings.json` for Geoff to drop. **Owed:** settle `init`/`upgrade` naming
  against `dubplate-server`; the configuration-reference generator as a hook 18 consumer.
- **Rung 9,** errata inventory, `give-back`, `packaging`, photo lane: HISTORY, slice 1 close, "Moved out of STATUS".
