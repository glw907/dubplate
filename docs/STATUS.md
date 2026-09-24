# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, 5b-i, 5b-ii, and **5c-i** are closed. 5c-i closed on master in
the 2026-09-23 overnight window (HISTORY, top entry). Its seven tasks were accepted, and all eight
exit-gate assertions pass. The security read found no critical and no major finding, and its minors
are folded. The spend is about 5.8M against the 6.5M ceiling, which crossed the 5.2M flag. There were
zero execution sittings. The window's document lanes applied the rung 6 errata (`b406f93`) and
brought the **6a plan to revision 2.2** (`plans/2026-09-23-rung-6a-store-foundation.md`), which
awaits approval. R4's exclusion denominator is still unread, and owed before rung 15.

## Rulings taken (Geoff, 2026-09-24)

- **The 6a plan is approved**, with all four of its rulings yes. Its eleven errata E1-E11 are applied.
- **5c-i's flag crossing is accepted.** m1 keeps the effective-principal rule, and the stricter
  shared-namespace rule is declined.
- **Follow beets on the "ft" fix.** The oracle pin moves to the 2.14 line through `dependency-upgrade`.
  The requirement is recorded in the build, rung 3, and rung 8 specs (`6d4dd05`).
- **The layout carries disc numbers**, as a `1-01` prefix only on multi-disc albums, with `Disc N/`
  subfolders available as a configured template. This is recorded in the configuration homes, build,
  and organizer specs.

## Next action

1. **Lane 2 runs 6a.** Segment A closed green (Tasks 1-3 accepted, `814a891`..`2d863cb`, one fix
   round each on 1 and 2); spend about 1.66M with Task 0. Re-derived: segment B about 2.2M, the
   simplifier round about 0.4M, segment C about 0.8M, so about 5.1M against the 5.44M flag and 6.8M
   ceiling. Segment B (Tasks 4-7) in flight; Task 0's notes 7 and 8 ride Task 7's brief.
2. **5c-ii is approved** (Geoff, 2026-09-24, all seven rulings yes): revision 2.2 at
   `plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`, 9.4M ceiling, 5c-iii split off.
   It opens in lane 1 once the beets adoption lands (one executor per tree).
3. **Beets 2.14.1 adopted** at `367564c` (the ft compat class retired, six sanctioned empty-search
   responses, 215 cases re-frozen: 209 `config_hash` only, six `adv-ft-*` now medium/skip). Review
   verdict fix, comment-only; the fix round (citation re-anchor in `tunables.go`, README, one test
   message) is in flight, then `code-simplifier`, then 5c-ii opens in lane 1.
4. The architecture cleanup pass joins the staged comment-cleanup task
   (`research/2026-09-24-rung-5c-i-architecture-reads.md`, plus the 5b-ii reads).

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | OPEN on the beets 2.14.1 adoption (Geoff, 2026-09-24: "do that in this pass"); 5c-ii follows once its plan is approved |
| 2 | worktree `.claude/worktrees/lane-2-6a`, branch `lane-2-6a` | OPEN on 6a since 2026-09-24; segment A green at `2d863cb`, segment B (Tasks 4-7) in flight |

## Open items (carried)

- No deploy until 5d: the env template selects `magic-link` without its nine keys; `main.go`'s
  `newProductionRegistry` is contested.
- To 5c-ii: the admin gate (R-5c12), the roster API (R-5c13), serve wiring, `web/`, the rejoin
  sentences (`CLAUDE.md:140`, `scripts/check.sh:5`, `Makefile:33`), security n5 on the web write path.
- To 5d: the "`doctor` exits 0" row, `ReadWritePaths`, the nine keys, the trusted proxy, and the
  runbook rows (`member add` without secrets in argv, kind mode's enumeration bound, the mail-ceiling lever).
- To rung 12: no principal maps a non-admin onto a streamer admin (Navidrome 0.63.2 matches names
  case-insensitively). To rung 6: errata row 9's `identity.go` half. To rung 7: archive uploads (`d74543d`).
- 86 declared coverage gaps, none blocking. Staged legs warn: 325 process citations, 128
  exported-unused. Fork B: 1.0 is FLAC-only; the group non-FLAC question is carried.
  Configuration homes routes tasks into rungs 6, 7, 9, 12, 14, 16 and the 2.0 list.
- From the beets adoption: Go RE2's ASCII `\b` against Python's Unicode `\b` gives a tenfold ft
  false positive on words like "Geschäft:" in both matcher modes (pre-existing, not a regression;
  needs a design call and a corpus case); about 263 beets line citations across `internal/engine`
  still point at 2.13.1 (to the comment-cleanup pass); the six `adj-adv-ft-*` rows still read
  "unreleased at 2.13.1" (Geoff's next adjudication batch); 2.14's empty-query search skip for the
  musicbrainz port; `deploy.sh`'s `BEETS_VERSION` 2.14.1 upgrades the box at the 5d deploy.
- Leftovers: four-knob wall registration; `AdminEmail`'s contact row; five 3b items; `profile_hashes`;
  response-status extension; `Case.Extra`; `slices.Clone` (rung 9); the corrupted untracked 2026-09-04 evidence doc; two review albums on
  the box; `FILEBROWSER_ADMIN_PASSWORD`; stale failed units; `DUBPLATE_SITE_NAME` re-mint.
- Pinned (2026-08-30): Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.13.1`, mediafile `0.17.0`,
  unidecode `1.4.0`, lap `0.5.13`. Spend: about 61.6M plus 9.1M (2026-09-12 window),
  6.4M (5b-ii), and about 8.3M (2026-09-23 window, 12M cap; conductor's sum of subagent tokens).
  5b-i's overage still sits on R23.
