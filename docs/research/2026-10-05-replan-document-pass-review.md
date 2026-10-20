# Re-plan document pass: plan review and fold

Target: `docs/superpowers/plans/2026-10-05-replan-document-pass.md`, revision 1 at `15be3fd` (365 lines), folded in
place to revision 2 (404 lines). Spec: `docs/specs/2026-10-05-rungs-12-16-replan-design.md`, revision
2.1 at `33b465c`, owner rulings 1 to 3 ruled yes.

One self-folding lens, `docs` class, combining contract and criteria, mechanics and feasibility, and domain risk.
Per Anthropic's Claude Code guidance, findings are limited to gaps that affect correctness or the stated
requirements and are ranked by consequence; polish is not chased.

## What was checked and held

- **Coverage.** Every item in the spec's "Document changes" (1 to 9) and the fold record's "Errata owed" (1 to 7)
  is placed: 1 to Task 1, 2 to Task 2, 3 to Task 3, 4 to Task 4, 5 to Task 5, 6 to Task 7, 7 to Task 8, 8 to Task
  6, 9 to Tasks 9 and 10. The 6d revision 2.3 (errata owed item 7) is deferred with a reason under "Not in this
  plan". Owner ruling 1's IDEAS entry, ruling 2's IDEAS entry and runbook deletion, and ruling 3's slice 8 line
  are all placed.
- **Targets resolve.** Every section and line reference the tasks cite resolves at `33b465c`: the build spec's
  "Drivers own driver knowledge", "One owned prefix", "Box safety during the build", "Shadow: the standing
  oracle", `:459-461`, and `:976-988`; the streamer spec's "Relation to the current ladder", "Cutover and
  operations", "The ladder" item 9, hooks 5, 6, 15, and 19, the two open items, and `:675-676`; the configuration
  homes spec's `:123-126` and `:221-224`; the group-scale spec's "Before 0.8", A2, A13, and "Versions this spec
  speaks to"; the rung 6 spec's "Migrations" and append-only rule.
- **Task 2's additions** beyond doc change 2 (the gonic and Navidrome oracle containers, the configuration
  properties, the bearer-bypass sentences) are grounded in the spec's own text (`:127`, `:134`, `:227-231`).
- **Workstation tooling for Task 0.** `~/.dotfiles/scripts/secrets/secret-set.sh` exists. Its recipient is the
  hard-coded `RECIP=` line (no `AGE_KEY_FILE` variable exists). The identity cache is `/dev/shm/.age-key-<uid>`,
  shared with `sync.sh`. `secret-set.sh --sync` runs `ensure_key` before its sync-only branch, so it warms a cold
  cache as the plan says. `age`, `sqlite3`, and `shred` are installed. `~/.local/share` is outside every git
  repository, and `~/.dotfiles/secrets/` is tracked, which justifies the plan's reading of "beside the secrets
  store". The round trip in step 4 proves the recipient is the right one.
- **Model and executor choices** follow the global seat table; the conductor stays thin.

## Findings and dispositions

Severity: **major** would cause a wrong result, a blocked task, or a false acceptance; **minor** would cost a
search or a fix round but not a wrong result.

| Id | Severity | Finding | Disposition |
|---|---|---|---|
| R1 | major | The gate paragraph claimed no lighter per-task check exists for a docs-only diff. `scripts/check.sh --reduced` exists and its header calls it "the comment-only and docs-only form". The conclusion still holds, because the same header says it is "never the final gate on master" and "a change that reaches master has had the full form run over it", and every task here commits to master. | Folded: the paragraph quotes `CLAUDE.md` and the script's header and states why the full form applies. |
| R2 | major | Under `passClass: 'docs'`, `pass-execute` reduces a fix round whose findings are all comment-only to its class default, "the repo's type check plus only the test files this fix round touched". Every docs finding is comment-only, this repo defines no such gate, and the fix commit lands on master without the full form. | Folded: the execution-mode line sets `reducedGate` to the full-gate string, using the runner's documented arg. |
| R3 | major | The shared-tree rule dispatched a document task "while lane 1 is between tasks". A `pass-execute` segment runs its tasks back to back, so the conductor has no between-task window inside a lane 1 segment. The plan also assumed lane 1 commits on master; if the removal plan used a branch in the main tree, segment A's commits would land on it. | Folded: segment A runs before lane 1's first segment or at a lane 1 segment boundary. The pre-flight agent reports whether the removal plan commits on master; if it branches, segment A finishes first. "The merge" is defined for segment B. |
| R4 | major | Task 0's integrity proof could pass vacuously. SQLite opens a zero-byte or missing file as an empty database, and `PRAGMA integrity_check` on it prints `ok`. | Folded: each extracted copy also needs a nonzero size and `SELECT count(*) FROM sqlite_master` above zero. Neither reads member data. |
| R5 | major | Task 2's grep acceptance could not be met. The "Cutover and operations" heading is cited by four other documents, so under the plan's own rule it keeps its name and hits `cutover`. The shadow-length open item is marked moot in place and still hits `shadow period`/`cutover`. Two passages outside the named sections also hit: "Hosting and setup" `:553` ("The group's own cutover uses this command") and the prior-art cell "Cutover import" `:797`. | Folded: the two passages are rewritten to the first-launch import (ruling 2), and the grep allows the kept heading and the moot open item. |
| R6 | major | Task 7's grep contradicted its own first bullet. The bullet requires a reading-rule clause naming older documents' "0.x launch and a rung 16 cutover"; the grep required no `0.x` and no `cutover` anywhere in `CLAUDE.md`. | Folded: the grep allows that clause, and the report lists every hit. |
| R7 | minor | Task 0 step 5 pointed at the estate inventory for the R2 bucket, which names no dubplate bucket. The repo's `.claude/instructions/api-access.md` names `music-library`; the old backup wrote `state/` and `_archive/state/<date>/`; the read-only credential `MUSIC_R2_RO_*` survived the retirement; HISTORY records an earlier empty `state/`. | Folded: the step names the bucket, prefixes, credential, and prior evidence. |
| R8 | minor | Task 0 assumed one `navidrome.db` in the tar. The old stack staged a backup snapshot beside the live database, so the archive may hold more than one copy. | Folded: step 1 records every match, step 4 checks each, and none present is a stop. |
| R9 | minor | Task 0 named the box access vaguely and did not size the pull against the scratchpad. | Folded: `root@5.78.148.42` with key auth, per `api-access.md`. The scratchpad is tmpfs (`/tmp`, about 7 GB free), so plaintext never reaches disk; the chore stops if the archive does not fit twice over. |
| R10 | minor | The plan did not say whether the `.age` copy is backed up. The workstation runs no backup job, so until rung 16 the copies are the box's plaintext tar and the workstation `.age` file. The spec gives the second copy to rung 16's backup remote, but no task carried that. | Folded: Task 0 states the posture, and Task 10's STATUS step adds the rung 16 carry. |
| R11 | minor | Task 6 required a repo-wide `grep` for `forward-auth-topology` to return nothing. The plan itself and the research records contain the string. | Folded: the grep is scoped outside `docs/` and `docs/HISTORY.md`. |
| R12 | minor | Task 8 kept all of `streamer-0-8`'s text in ROADMAP ("nothing is deleted") while its grep required no `escape hatch` outside Done entries. The entry's "Sequenced after launch (rung 17); the driver seam stays and Navidrome remains the escape hatch" sentence is superseded. | Folded: superseded passages leave ROADMAP, quoted in the report, and Task 10 moves them into HISTORY's re-plan entry. |
| R13 | minor | Task 1's "`--numstat` shows deletions only on that line" is not something numstat shows. Task 3's errata count sits in a wrapped status paragraph, so updating it may reflow more than one line. | Folded: Task 1 checks one deletion and names it from `git diff`; Task 3 allows the status paragraph's reflow. |
| R14 | over-ceremony | The full heavy gate (about 106 s, Chromium) runs on every docs commit, eight or more times. | Refused: `CLAUDE.md` requires `bash scripts/check.sh` "before every commit", and the script's own contract bars `--reduced` as the final gate on master. Moving the lane off master would break `CLAUDE.md`'s rule that document lanes share the main tree. The cost is priced into the ceiling. |
| R15 | over-ceremony | Task 0's quiet-box check asks for "no upload in progress" on a box whose serve unit is disabled and which holds no store, so that clause cannot fail. | Refused: the check costs one read-only command and the lane registry clause is a real check. Trimming saves nothing worth a revision. |

## Planning-miss items

- **Quoted behavior.** Task 0 relies on `secret-set.sh --sync` warming the key cache (verified in the source:
  `ensure_key` runs before the sync-only branch) and on SQLite's integrity check (R4 closes its empty-file case).
- **Every state a check can meet.** Task 0's tar listing now names absent and many (R8). The R2 listing names
  present, absent, and no credential. The scratchpad names too small (R9).
- **Privileged fields.** None in a docs pass.
- **Fixture state for proofs.** R4: the integrity proof now needs a populated database.

## Counts

| Severity | Folded | Refused | Owner fork |
|---|---|---|---|
| Major | 6 | 0 | 0 |
| Minor | 7 | 0 | 0 |
| Over-ceremony | 0 | 2 | 0 |
| Total | 13 | 2 | 0 |

No owner fork. The one question with any choice in it, where the encrypted copy lives and whether it gets an
earlier off-workstation copy, was answered by the spec (rung 16's backup remote takes the second copy), so the
plan records the posture and carries it rather than asking.

Plan length: 365 lines before, 404 after.
