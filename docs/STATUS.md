# dubplate status

## Current state

Rung 5 and the navidrome-removal and re-plan document passes are closed (HISTORY). Rung 12 slice 1 (the study)
is closed and landed: nine records and streamer spec revision 3.1. 6d's build is on master (`beda726..4de73e7`), and
slice 2a's harness waits on its branch. Master is green. The box is unchanged: dubplate-server binary `d1e93e9`, the
serve unit installed and disabled, the `auth`, `roster`, `unidecode` leaves, one admin member, and
`/srv/music/library` (6.6 GB); the first 6d deploy takes the runbook's 6d upgrade. The archive is in HISTORY,
"Archive record".

## Next action

> **Resume prompt (launch in /var/home/glw907/Projects/dubplate; `claude --model claude-opus-5-5`, effort medium)**
> **Goal:** close 6d, then merge and close slice 2a. Both parked 2026-10-07 at 10:15 when 6d crossed its 8.4M flag
> (about 9.5M spent of 10.5M). **6d** (`plans/2026-09-26-rung-6d-identity-serve-store.md`) has only Task 9 left (the
> HISTORY entry, the errata, STATUS and the `CLAUDE.md` gate line); every build task, both security reads and folds,
> the simplifier, seven architecture reads and their folds (`5079b38`, `4de73e7`), and a 7/7 live auth smoke are
> done. Task 9's inputs are `research/2026-10-07-overnight-conductor-notes.md` (carries, rulings, spend, deviations)
> and the commit log `beda726..4de73e7`. **Slice 2a** (`plans/2026-10-06-rung-12-slice-2-oracle-harness.md`, worktree
> `.claude/worktrees/lane-3-slice-2`, branch at `25c442a`, about 6.0M of 9.0M) has only Task 9 steps 4, 6, and 7 left:
> the merge ritual after 6d closes, then its ledgers. **Settled:** merge order 6d then 2a; the 2a rulings (2 s seeds,
> Two Sides as the multi-disc album) are in those notes. **Approach:** invoke `pass-core`; nothing deploys;
> push master (no force, `gitleaks` first) at each close.

## Owner questions
- R-6d21 overrides 6c's "never reads environ" rule; Geoff may reopen it (not blocking).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- Slice 1's three study forks (`research/2026-10-06-slice-1-study-record.md`), one answer each: **A** the
  `transcoding` extension at 0.8 (recommend no, revisit at 1.0); **B** the protocol module's license and NOTICE
  (recommend MIT, the OpenAPI file loaded as caller input, never vendored; the OpenSubsonic repo has no NOTICE file
  today); **C** folder browsing at launch (recommend no).
- 6d crossed its 8.4M flag at about 9.5M (its architecture fold alone was 2.6M); close both lanes over the flag?
  6d's Task 9 is docs, about 0.4M, inside the 10.5M ceiling.
- Workstation: ~41 stale `DUBPLATE_*` exports in `~/.local/secrets` fail `check` outside `env -i`; prune them.
- VT-x appears disabled in firmware (no vmx/svm flag, no `/dev/kvm`), so the Android emulator cannot run; enabling it
  unblocks Tempus walks and slice 2's roster walk.

## Lane registry
| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | PARKED at its flag 2026-10-07 10:15: 6d Tasks 0-8, both security folds, simplifier, architecture folds, live auth smoke done (through `4de73e7`, ~9.5M); Task 9 left |
| slice 1 | document lane | closed 2026-10-07, landed on master (`00d4b1a`, `d11f853`) |
| slice 2 | `.claude/worktrees/lane-3-slice-2`, branch `lane-3-slice-2` from `beda726`; contests `go.mod`/`go.sum` (6d merges first) | PARKED for the merge: Tasks 1-8 and close steps 1-3 and 5 done (`25c442a`, ~6.0M); merges after 6d's Task 9 |
| photos | main tree, docs session | round 1 closed at `c10c0f2`; one Run-D.M.C. frame awaits Geoff |

## Open items (carried; landings from the re-plan spec's "Where the carried items land")

- **Rung 16:** enabling the serve unit; secrets in every `music`-uid verb's environment; SMTP from the box; 6c's
  Ruling 2; 6b-ii-b's runbook lines; the archive's second copy; per-collection relocation; store identity.
  **Rung 14:** the removal verb bumps the address's revocation epoch as it tombstones (6d security read B, Low 3).
  **Slice 3:** UNI-H2, the prefix-gap and percent-encoding items. **Slice 6:** the streamer-admin mapping,
  offboarding deprovision, CSP-scope and shared-origin M2, username uniqueness. **Slice 7:** the streamer database
  in a restore. **Slice 8:** group 2 in trailers, `CAP_NET_BIND_SERVICE`, command naming.
- **Bugs and chores:** the bearer bypass's first consumer owes a cookie test (HISTORY); BACKLOG #15, #16, #21-#26;
  two dead `docker compose` entries in `.claude/settings.json` for Geoff to drop.
- **Owed:** settle `init`/`upgrade` naming against `dubplate-server`; the configuration-reference generator as a
  hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- **To 6d** (the 6d plan carries them), **rung 7, rung 9, revoke verb, rung 14**, the errata and architecture
  inventory, `give-back` and `packaging`, and the photo lane: HISTORY, slice 1 close, "Moved out of STATUS".
