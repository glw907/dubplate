# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d), rung 1b, rung 8 (8a, 8b), rung 4 (4a, 4b), and now **rung
5's pass 5a (the serve shell and ingress)** are closed
(`docs/HISTORY.md`'s "Rung 5, pass 5a" entry, merged at `307001b`).
`internal/serve` is born: the composition order, the ingress strip,
path normalization, prefix routing, the extension point, the deadline
policy, and the bind surface. A fresh-context
`web-auth-security-reviewer` read the whole shell; findings are at
`docs/superpowers/research/2026-09-08-rung-5a-security-review.md`. The
rung 5 spec stands at revision 2.1, errata now carrying 5a's nine
corrections. The 4b sitting deployed the shadow units to the box; the
adjudication stream is live, read through `dubplate-server shadow
report`, never the bare binary as root. Streamer 2.0 (ROADMAP
`streamer-2-0`) binds rungs 5-9 and 12. The box (5.78.148.42) is
production, runs Navidrome and the guardian and shadow timers; the
group system is not live.

## Next action

Session 2026-09-11/12 was a design sitting, not a lane: the
configuration homes spec landed at revision 3 (`07f4e01`,
`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`),
ratified in conversation and awaiting Geoff's read of the file; its
routed tasks are in the carried list below and its errata are in the
ratified docs. No code moved. The lane plan is unchanged: both next
plans are drafted,
reviewed, folded to revision 2, and committed at `5377d9f`, unruled:

- `docs/superpowers/plans/2026-09-08-rung-5b-auth-core.md` (pass
  **5b-i**: identity, config contract, token, session, state records;
  Task 0 plus ten; 5.0M ceiling, 4.0M flag). The drafter took the
  spec's fifteen-task split, so **5b-ii** (roster, limiters, notify,
  origin check, verify seam, handlers; provisional 4.5M) follows.
  Review record `research/2026-09-08-rung-5b-plan-review.md`: 28
  findings folded, none rejected, nine conductor-owed folds marked
  inline, each with the fold editor's recommendation (rule all nine as
  applied). Runs in lane 2 (worktree `.claude/worktrees/lane-2-5b`).
- `docs/superpowers/plans/2026-09-08-rung-8c-absorb-repoint.md` (Task 0
  plus four; 2.1M ceiling against the spec's 1M, 1.68M flag). Review
  `research/2026-09-08-rung-8c-plan-review.md`: 13 folded, none
  rejected, three conductor-owed folds (R1 as a stated departure from
  R12's rung 9 assignment; the `Parse` read path; R4's composed
  denominator). Runs in lane 1 (main tree).

**Resume prompt:** rule the twelve conductor-owed folds (mark each at
its site, commit the plans as revision 2.1); write the rung 5 spec's
contested-surface rows and 8c's `internal/engine/shadow` amendment into
`specs/2026-09-03-parallel-build-approach.md`'s ownership table; run
each plan's Task 0 (fresh context, `opus`); then launch pass-execute
for 8c in the main tree and 5b-i in the worktree, guards armed
(`~/.claude/docs/unattended-work-guards.md`). The prep workflow was
`wf_0663b27f-52a` (seven agents, 1.99M); its cache is session-bound,
so nothing resumes from it.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | RUNNING 8c segment 1A (run `wf_c808ec5e-1ba`): Task 1 accepted `9be7e43`; Task 2 `3a299ac`+`6de7d55` escalated on R6 (ruled: byte-splice fixture row, fix dispatch at the 1A boundary); Task 3 in flight. Window plan: `plans/2026-09-12-overnight-window.md` |
| 2 | `.claude/worktrees/lane-2-5b` (branch `lane-2-5b`, created from `7ba3cc1`) | RUNNING 5b-i: segment 2A closed at `eddeadd` (Tasks 1-3 accepted, one fix round on Task 3, run `wf_61f3cf53-665`, 0.79M); simplifier round then segment 2B (Tasks 4-7) |

Grant reading, flagged (overnight plan, G1): the twelve conductor-owed
folds were ruled by recommendation on "then start" (Geoff, 2026-09-12
00:44); the ratify-or-reverse table is in the overnight plan.

## Open items (carried)

- Configuration homes spec (revision 3, 2026-09-12) routes tasks the
  just-in-time plans must carry: rung 6 (settings/overrides/preferences
  tables, collection-row attributes, device credentials, quota counters,
  move journal), rungs 7 and 9 (template grammar and root-confined
  writes, audio key by decode plus file hash, break-link rule, intake
  caps, same-device check), rung 12 (capability report extensions and
  three pinned-build verifications), rung 14 (two mounts, admin
  endpoints incl. seize and removal, re-auth on sensitive settings),
  rung 16 (single-link migration, restore-size runbook line), and the
  2.0 projector conformance list. Full list: the spec's "What this spec
  routes elsewhere". Build spec, organizer spec, 2.0 spec, ROADMAP, and
  the methodology study carry dated errata.
- Sixty-three declared coverage gaps (rung 3's twenty-one, rung 4a's
  nine, rung 4b's four, rung 1b's two, rung 8a's seven, rung 8b's
  nine, rung 5a's eleven), none blocking. Fork B is ruled (Geoff, 2026-09-08: FLAC-only 1.0, the format ladder at 1.1, ROADMAP `formats-1-1`); the group non-FLAC question is unresolved, carried.
- Four-knob wall registration unblocked; `AdminEmail`'s contact row,
  five 3b leftovers, `profile_hashes` (m1), the response-status
  extension (2b/2c), `Case.Extra`, `slices.Clone` (rung 9).
- The 2026-09-04 agentic-coding-evidence doc is corrupted, untracked;
  two albums await review in `/srv/music/review/2026-08-31/geoff/`;
  `FILEBROWSER_ADMIN_PASSWORD` orphaned; stale failed units;
  `DUBPLATE_SITE_NAME` retires when re-minted.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 61.6M through the 5b-i and 8c plan prep (59.6M plus 1.99M).
