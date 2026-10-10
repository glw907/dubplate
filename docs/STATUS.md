# dubplate status

## Current state
Rung 5, the navidrome-removal and re-plan document passes, rung 12 slice 1 (the study), rung 6d (identity in the
store, the auth settings, and `serve`'s store), and rung 12 slice 2a (the oracle harness) are closed (HISTORY). 6d's
code is on master (`beda726..4de73e7`) and 2a's is merged at `70bcc30` (`d87ecc5..70bcc30`). Master is green.
**Slice 2b is paused mid-pass** (Geoff, 2026-10-07: wind down, resume next week). On branch `lane-3-slice-2`, not yet
merged: Tasks 1 to 8 accepted (`9f878c3..f5c916d`, plan revision 1.3 with errata E1 to E6), Task 0 clean (nothing
moved upstream), and the Navidrome library request observed and replayed (record at `5d00313`). Segment A and B
boundary gates were green (`f4ba980`, `5d00313`); Tasks 7 and 8 passed their scoped light gates only, so the full
gate has not run at `2f1e3a2`. Spend about 3.5M of the 9.0M ceiling. The box is unchanged: dubplate-server binary
`d1e93e9`, the serve unit installed and disabled, the `auth`, `roster`, `unidecode` leaves, one admin member, and
`/srv/music/library` (6.6 GB); the first 6d deploy takes the runbook's 6d upgrade. Archive: HISTORY, "Archive record".

## Next action
> **Resume prompt (launch in /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-3-slice-2; `claude --model
> claude-opus-5-5`, effort medium)**
> **Goal.** Finish rung 12 slice 2b: Task 9 (two collections), then segment D (Tasks 10 and 11), then the close
> (Task 12). **Settled (do not re-brainstorm):** the plan at revision 1.3 and its Errata E1 to E6 (E5 binds Task 9:
> the normal form, the repeated `--collection name=manifest` flag, the `stack.Start` order, the fake server's new
> routes); Ruling 1 yes (ceiling 9.0M, about 3.5M spent); forks A, B, C ruled. **Approach:** invoke `pass-core`;
> check the lane is quiet (`git status` clean at `2f1e3a2`); launch `pass-execute` for Task 9 alone, with segment C's
> args (in the plan's "Ceiling" section, plus `commonNotes` pointing at the Errata), then run segment C's boundary
> (full heavy gate, segment D pre-flight, STATUS), then segment D (Task 11 `passClass: docs`, `check.sh --reduced`;
> Tempus deferred until `/dev/kvm` exists), then Task 12's close. Before each launch, `cmp` the persisted runner
> against `~/.claude/workflows/pass-execute.js`; if they differ, copy the committed file to the session scratchpad
> and launch with `scriptPath`. Arm `claude-wf-guard`, a `/loop` fallback wake-up, and the lid-switch hold at
> launch. Nothing deploys. The close merges to master by the ritual and pushes (no force, `gitleaks` first).
> **After 2b:** the ladder's next rung (rung 14, after refreshing ROADMAP's stale "6d is next").

## Owner questions
- R-6d21 overrides 6c's "never reads environ" rule; Geoff may reopen it (not blocking).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- Workstation: ~41 stale `DUBPLATE_*` exports in `~/.local/secrets` fail `check` outside `env -i`; prune them.
- VT-x appears disabled in firmware (no `/dev/kvm`): enabling it unblocks the Android emulator, Tempus walks, and
  slice 2's roster walk.

## Lane registry
| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | idle: 6d closed 2026-10-07 (code through `4de73e7`, close commit `0aa8d18`; ~9.5M of 10.5M) |
| slice 1 | document lane | closed 2026-10-07, landed on master (`00d4b1a`, `d11f853`) |
| 3 (slice 2) | `.claude/worktrees/lane-3-slice-2`, branch `lane-3-slice-2` | **running** (overnight 2026-10-10 from 00:01): segment C's Task 9 in `pass-execute` run `wf_c273e213-588`, then segment D and the close; contested: `scripts/check.sh`, `Makefile`, `cmd/harness/main.go` |
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
  lines; the archive's second copy; per-collection relocation; store identity (R9's option b).
  **Slice 3:** UNI-H2, the prefix-gap and percent-encoding items. **Slice 6:** the streamer-admin mapping, the
  `duplicate_principal` code and BACKLOG #16, offboarding deprovision, CSP-scope and shared-origin M2, username
  uniqueness. **Slice 7:** the streamer database in a restore. **Slice 8:** group 2 in trailers,
  `CAP_NET_BIND_SERVICE`, command naming.
- **Lane 1's next `web/` pass:** R-6d17's note and BACKLOG #23 (HISTORY, 6d, "Routed items"). **Bugs and chores:** the
  bearer bypass's first consumer owes a cookie test (HISTORY); BACKLOG #15, #16, #21-#27; two dead `docker compose`
  entries in `.claude/settings.json` for Geoff to drop. **Owed:** settle `init`/`upgrade` naming against
  `dubplate-server`; the configuration-reference generator as a hook 18 consumer.
- **Slice 2a's carry for 2b:** the seed lacks lyrics, `nowPlaying`, and radio requests, so the study's expected
  contract flags for them are owed to 2b's recording; BACKLOG #22 and #27 are its other routed items.
- **Rung 7, rung 9,** errata inventory, `give-back`, `packaging`, photo lane: HISTORY, slice 1 close, "Moved out of STATUS".
