# dubplate status

## Current state
Rung 5, the navidrome-removal and re-plan document passes, rung 12 slice 1 (the study), and rung 6d (identity in the
store, the auth settings, and `serve`'s store) are closed (HISTORY). 6d's code is on master (`beda726..4de73e7`).
Slice 2a's harness waits on its branch for its merge. Master is green. The box is unchanged: dubplate-server binary
`d1e93e9`, the serve unit installed and disabled, the `auth`, `roster`, `unidecode` leaves, one admin member, and
`/srv/music/library` (6.6 GB); the first 6d deploy takes the runbook's 6d upgrade. Archive: HISTORY, "Archive record".

## Next action
> **Resume prompt (launch in /var/home/glw907/Projects/dubplate; `claude --model claude-opus-5-5`, effort medium)**
> **Goal:** merge and close rung 12 slice 2a (`plans/2026-10-06-rung-12-slice-2-oracle-harness.md`, worktree
> `.claude/worktrees/lane-3-slice-2`, branch `lane-3-slice-2` at `25c442a` from `beda726`; about 6.0M of 9.0M). Tasks 1
> to 8, the simplifier, the vacuity fold, the oracle-tag vet, and the architecture reads and their fold are done; Task 9
> steps 4, 6, 7 are left. **Step 4:** the merge ritual (repo `CLAUDE.md` "Lanes"; plan Task 9 step 4 for the
> `go.mod`/`go.sum` and kin-openapi handling), then one `diff-reviewer` over the merged range. **Step 6:** three edits
> in one batch: `CLAUDE.md` (R1's wording with the `go vet -tags oracle` check; the gate paragraph naming the boundary
> and no-link tests); HISTORY's 2a entry, from the slice 2a sections of `research/2026-10-07-overnight-conductor-notes.md`
> (golden counts, the 586 s baseline, budgets, the 2 s seed and Two Sides rulings, the light-lane slip, architecture
> reads before the merge); this STATUS (lane 3's row, 2b's wait on its plan). Also add the plan erratum for the two
> rulings and file the `--scope-changed` refinement for `scripts/check.sh`. **Step 7:** hold the worktree for 2b.
> **Approach:** invoke `pass-core`; pass `gateLane: "heavy"`; nothing deploys; push master (no force, `gitleaks` first).

## Owner questions
- R-6d21 overrides 6c's "never reads environ" rule; Geoff may reopen it (not blocking).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- Slice 1's three study forks (`research/2026-10-06-slice-1-study-record.md`), one answer each: **A** the
  `transcoding` extension at 0.8 (recommend no); **B** the protocol module's license and NOTICE (recommend MIT, the
  OpenAPI file loaded as caller input, never vendored); **C** folder browsing at launch (recommend no).
- Workstation: ~41 stale `DUBPLATE_*` exports in `~/.local/secrets` fail `check` outside `env -i`; prune them.
- VT-x appears disabled in firmware (no `/dev/kvm`): enabling it unblocks the Android emulator, Tempus walks, and
  slice 2's roster walk.

## Lane registry
| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | 6d closed 2026-10-07 (code through `4de73e7`, close commit `0aa8d18`; ~9.5M of 10.5M) |
| slice 1 | document lane | closed 2026-10-07, landed on master (`00d4b1a`, `d11f853`) |
| slice 2 | `.claude/worktrees/lane-3-slice-2`, branch `lane-3-slice-2` from `beda726`; contests `go.mod`/`go.sum` | PARKED for the merge: Tasks 1-8, close steps 1-3 and 5 done (`25c442a`, ~6.0M) |
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
  bearer bypass's first consumer owes a cookie test (HISTORY); BACKLOG #15, #16, #21-#26; two dead `docker compose`
  entries in `.claude/settings.json` for Geoff to drop. **Owed:** settle `init`/`upgrade` naming against
  `dubplate-server`; the configuration-reference generator as a hook 18 consumer.
- **Rung 7, rung 9,** errata inventory, `give-back`, `packaging`, photo lane: HISTORY, slice 1 close, "Moved out of STATUS".
