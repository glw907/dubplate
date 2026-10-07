# dubplate status

## Current state

Rung 5 is closed (2026-10-05). The navidrome-removal and re-plan document passes are closed (HISTORY's entries
below slice 1's): the shadow program, front exemption, fallthrough, and streamer and front keys are gone, the roster,
store, API, CLI, and members page say `username`, and the specs, `CLAUDE.md`, `ROADMAP.md`, and `IDEAS.md` carry
the rungs 12 and 16 re-plan. Rung 12 slice 1 (the study) is closed and landed: nine records and streamer spec
revision 3.1. Master is green. The box runs only dubplate-server: binary `d1e93e9` (doctor exit 0), the serve unit
installed and disabled, the `auth`, `roster`, `unidecode` leaves, one admin member, and `/srv/music/library`
(6.6 GB). No rollback set remains. The first roster-reading deploy takes the runbook's re-create step. The retired
musicbox archive's copy is `~/.local/share/dubplate/archive/musicbox-retired-2026-10-05.tar.gz.age` (HISTORY,
"Archive record").

## Next action

> **Resume prompt (launch in /var/home/glw907/Projects/dubplate; `claude --model claude-opus-5-5`, effort medium)**
> **Goal:** run the three-lane window (Geoff, 2026-10-07): lane 1 runs 6d rev 2.5 (`plans/2026-09-26-rung-6d-identity-serve-store.md`,
> `auth-data`, 10.5M, flag 8.4M), segment B (Tasks 3-5) next; slice 2a rev 1.2 (`plans/2026-10-06-rung-12-slice-2-oracle-harness.md`,
> `tool`, 9.0M, flag 7.2M) runs in its worktree and merges only after 6d closes. **Settled:** merge order 6d then 2a;
> slice 1's first oracle pin wins and its pin list goes to 2a; the live client walks are conductor-run. **Approach:** invoke `pass-core`;
> run every lane's launch checklist (repo CLAUDE.md "Lanes"); `pass-execute` by name, one segment per invocation;
> never re-run a full gate on an unchanged commit. **Window:** 10+ hours, nobody attending; each lane stops at its
> own flag, and a blocker or genuine fork parks its lane with STATUS and one combined question; push master to origin at each merge (no force, `gitleaks` first); AC power; `claude-wf-guard` and `/loop`
> armed; nothing deploys, the box stays read-only. The full grant text is in HISTORY, slice 1 close, "Moved out of STATUS".

## Owner questions
- R-6d21 overrides 6c's "never reads environ" rule; Geoff may reopen it (not blocking).
- Fork B: 0.8/0.9 launch FLAC-only; the group non-FLAC question is still open.
- Slice 1's three study forks (`research/2026-10-06-slice-1-study-record.md`), one answer each: **A** the
  `transcoding` extension at 0.8 (recommend no, revisit at 1.0); **B** the protocol module's license and NOTICE
  (recommend MIT, the OpenAPI file loaded as caller input, never vendored; the OpenSubsonic repo has no NOTICE file
  today); **C** folder browsing at launch (recommend no).
- VT-x appears disabled in firmware (no vmx/svm flag, no `/dev/kvm`), so the Android emulator cannot run; enabling it
  unblocks Tempus walks and slice 2's roster walk.

## Lane registry
| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | RUNNING: segment A closed (Tasks 0-2: `beda726`, `9f8bd41`, `6bd91bb`, `acad661`, ~1.3M spent); segment B (Tasks 3-5) next |
| slice 1 | document lane | closed 2026-10-07, landed on master (`00d4b1a`, `the B-land commit after it`) |
| slice 2 | `.claude/worktrees/lane-3-slice-2`, branch `lane-3-slice-2` from `beda726`; contests `go.mod`/`go.sum` (6d merges first) | RUNNING: segment A closed at `7b54596` (Tasks 1-3, boundary gate green, ~1.25M with Task 0); segment B running |
| photos | main tree, docs session | round 1 closed at `c10c0f2`; one Run-D.M.C. frame awaits Geoff |

## Open items (carried; landings from the re-plan spec's "Where the carried items land")

- **Rung 16:** enabling the serve unit, the secrets in every `music`-uid verb's environment, SMTP exercised from
  the box, 6c's Ruling 2 (`deploy.sh` halts on `check`'s nonzero exit), 6b-ii-b's runbook lines (HISTORY
  6b-ii-b), the archive's second copy (the `.age` file to the backup remote), per-collection relocation, store
  identity if 6d's Ruling 3 is yes. **Slice 3:** UNI-H2 (`/` to login), the prefix-gap and percent-encoding items.
  **Slice 6:** the streamer-admin mapping, offboarding deprovision, CSP-scope and shared-origin M2, username
  principal uniqueness. **Slice 7:** the streamer database in a restore. **Slice 8:** group 2 in trailers,
  `CAP_NET_BIND_SERVICE`, command naming.
- **Bugs and chores:** the bearer bypass's first consumer owes a shell-level cookie test (HISTORY); BACKLOG #15
  (direct-mode warning fires on doctor's probe) and #16 (a username equal to another username-holder's
  address); two dead `docker compose` entries in `.claude/settings.json` for Geoff to drop.
- **Owed:** from the site spec, settle `init`/`upgrade` naming against `dubplate-server` and add the
  configuration-reference generator as a hook 18 consumer. Reading taken: OpenSubsonic wire types hand-written.
- **To 6d** (the 6d plan carries them), **rung 7, rung 9, revoke verb, rung 14**, the errata and architecture
  inventory, `give-back` and `packaging`, and the photo lane: HISTORY, slice 1 close, "Moved out of STATUS".
