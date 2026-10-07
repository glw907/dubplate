# Rung 6d plan fold: revision 2.3 to 2.4

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md`, revision 2.3 at `4b2cddf`,
pass class `auth-data`. Folded: the contract review (C*, O*), the mechanics review (M*, m*, O*), and the
risk review (R*), all dated 2026-10-06. The rebase record is `2026-10-06-rung-6d-plan-rebase.md`. Fold
agent: Opus 5.5, high effort.

Method. Each finding was checked against the tree before a disposition. `git diff --name-only
0cb33ae..HEAD` touches only documents, so the plan's code citations and the reviewers' citations
were read in the working tree. Checked by reading: `reconcile.go:166-167`, `keys.go:126-127`,
`admin_test.go:123-131`, `runbook :216-250` and `:334-345`, `unidecode.go:258`/`:269` and its tests,
`write.go:44`, `conn.go:146`, `doctor.go:48-52`/`:198-201`/`:369-391`, `restore_test.go:747-775`,
the six `CREATE TABLE`s in `0005_routed.sql`, `inspect.go:195-247`, `ServeOptions` (`open.go:85-116`),
`lock.go:99-108`, `identity.go:203-211`, `main.go:126-135`, the seven `config.Load` callers, `git show
d1e93e9:cmd/dubplate-server/member.go:101`, and the store package's test names the plan now cites.
The mechanics review's `storeFile` probe was not re-run; its output names the `0005` tables the
migration file creates, which matches the file.

## Dispositions

Twenty-six distinct findings after merging duplicates: **25 folded, 1 refused, 0 owner forks.** Every
review ID appears once below.

### Majors

| IDs | Disposition | Where in the plan |
| --- | --- | --- |
| C1 | Folded. Verified: `doctor` prints `statuses[...]` only for its own rows, so a superseded key's message never prints and `doctor` exits "configuration is incomplete" naming nothing. Fixed at the root: `doctor` prints one line per load error no row claimed, only when one exists. The fix covers every superseded, reserved, or unknown key, not only this one | Task 5 Files and Outcomes; criterion 4; mutations |
| C2 | Folded. Verified by trace: the rendered leg's add-member row writes before the first reset, so a first-POST snapshot equals the named set | Task 4 criterion (`rosterResetServer` alone); mutation retargeted |
| M1, contract O2, R10 | Folded together. Task 1 proves what `serve` decides. The keep decision becomes a pure function, table-tested. One wiring row per path: a `newer` refusal for the no-shell mapping, one in-process lock row for `MigrationWait`, and the foreign owner. D10 is proven end to end through a cmd older-store helper. The store tests are cited by name for the rest, and the re-executed lock child is dropped | Task 1 Files (the helper), "What this task proves", both state tables, criteria, mutations; Not verified |
| M2, R3 | Folded. Verified at `runbook :340-345` | R-6d18 "The settings rows section"; Task 8 Outcomes and render-and-read |
| R1 | Folded. Verified: `d1e93e9`'s `member add` takes `--streamer-username`, so the `0cb33ae` re-create section fails under the rollback binary. The rollback step moves the aside file back | R-6d18 rollback bullet; Task 8 render-and-read |
| R2, m1 | Folded. Verified at `reconcile.go:166-167`. Under Ruling 4's yes the line names `member add ADDRESS --admin`; under its no it stays | Task 1 Files, Outcomes, criterion, mutation; 6d-R11 restated; Ruling 4's "What yes builds" |

### Minors

| IDs | Disposition | Where |
| --- | --- | --- |
| C3 | Folded: the keep-row-of-1 states name the keep key unset | Task 1 keep table and start table |
| C4 | Folded in its proportionate form: Task 2 proves a refusal after the open releases the handle (an exclusive open succeeds). The verb's own `defer` is diff-reviewer-checked by name. Declined the alternative of running the verb's listener on a reserved port, since process exit releases the `flock` in production | Task 1 criteria; Task 2 criterion and mutation |
| C5 | Folded: the body-cap reader joins the failing-reader list, the cheaper option | Task 7 criteria and mutations |
| C6, m7 | Folded: three refusals, each named | Task 8 state table |
| C7, m4 (first half) | Folded at the root: R-6d21 and the members row key on `Report.Version`, never on `Facts`. A `0001` fixture is added to Task 4's goldens and Task 6's table, since the helper makes it one row | R-6d13, R-6d21; Task 4 members-row table and fixture state; Task 6 table and fixture state |
| C8 (a), m3 | Folded: `fixtureAdmin` named | Task 4 Files |
| C8 (b), m8 | Folded: `resolveAuthConfig` and its `auth_test.go` table go in Task 6 | Task 6 Files and interim read; Task 7 Files |
| C9, mechanics O2 | Folded into Task 8, not routed. `unidecode.go:269`'s hint is false for R-6d15's reason, and the Goal is operator strings that are true. Cost: one string and two test substrings. The criterion becomes a repo-wide `grep` that finds no `runuser` | Task 8 Files, Outcomes, criterion, mutation; Contested surfaces |
| C10 | Folded: the seven loading leaves are named, and the test walks `newRootCmd`'s tree against them plus the named non-loaders | Task 5 criterion 4 |
| contract O1, mechanics O1 | Folded in the mechanics form: the state-dir leg moves to `scripts/check-state-dir.sh`, called from `check.sh` on `vale-comments.sh`'s precedent (`check.sh:514`). Each state then runs in seconds. **Refused in part:** contract O1's cut to three states. Once a state costs seconds, the three states that exercise existing checks are what prove the move kept them | Task 8 Files and criterion; Task 9's `CLAUDE.md` line; Contested surfaces |
| m2 | Folded: `keys.go:126`'s message names `member add ADDRESS --admin`. It edits a message and declares no key | Task 5 Files and grep criterion; Contested surfaces |
| m4 (second half) | Folded: R-6d10 now credits non-empty usernames to R-6d22's write path, not `NOT NULL`, and notes `Valid` fails closed on an empty `Principal` (verified, `identity.go:208-210`) | R-6d10 |
| m5 | Folded: the lock row accepts either `busyError` form | Task 1 criteria |
| m6, R6 | Folded: `restore` named as a `users` channel, and the grep widened to insert, replace, update, and delete | Privileged fields; R-6d22; Review focus; Task 4 criterion |
| m9 | Folded: distinct `created_at`, admin first | R-6d10 `List` bullet; Task 4 Files |
| R4 | **Refused** (the mechanism). Both proposed fixes change `ServeOptions.Keep`'s contract (a callback evaluated under the lock, or a second `Inspect` that `OpenServe`'s internal wait gives no place for). That is new store surface for a state no 6d start meets: no store on the box, `0005` the last migration, and the keep falling to the default only under a busy or intent pre-lock read. What the fold keeps: R-6d4 states the limit, and Task 9 routes it to the plan of the first migration after `0005` | R-6d4's closing paragraphs; Task 9 routed items |
| R5 | Folded, narrowed. On an **older** store, a refusal the pre-open evaluation already shows stops the start before `OpenServe`. Narrowed from "any state with facts" because, on a current store, a pre-open check would pre-empt every Task 2 seed-refusal row. `serve`'s handling of the seed apply's own refusal would then go untested, and a current store migrates nothing anyway. Recorded as reading 6d-R15 | R-6d2 "The pre-open check"; R-6d4; Review focus; Task 2 Outcomes, state, mutation; 6d-R15 |
| R7 | **Not an owner fork; resolved by an existing schedule.** The configuration homes spec (ratified) files "the removal verb and this data cleanup" at rung 14, tombstoning the user row (CH `:576-583`), and lists "member removal as the checklist above" in rung 14's admin spine (CH `:732`). The re-plan leaves rung 14 unchanged except its private gate (`:108`). STATUS's slice 6 "offboarding deprovision" is the streamer half of the same offboarding. Rung 14 precedes rung 16's enabling of `serve`, so the window in which removal is a raw statement has nothing live | "What this pass is not"; R-6d10's last-admin bullet; 6d-R3; Task 9 routed items |
| R8 | Folded: the aside name (`roster.json.pre-username`, the re-plan `:328`), the source of step 4's fields, and the re-run state | Owner touchpoints steps 3 and 4; R-6d18 upgrade bullet |
| R9 | Folded: the `Update` row quotes "one IMMEDIATE transaction" and cites `conn.go:146` | Behaviors table |

## Rulings

Unchanged: Rulings 1, 3, and 4 stand as revision 2.3 posed them. No finding changes a ruling's
question or its answers. Ruling 4's "What yes builds" gains the reconciler line (R2), which is
consequence, not a new fork.

## Owed errata (ratified documents not edited)

- **Rung 6 spec `:732-733`** ("the report says to run `member import`"): under Ruling 4's yes, the
  restore report names `member add ADDRESS --admin`. Carried as the restated 6d-R11, which Task 9's
  errata step applies.
- **Rung 6 spec `:891-892`** (the start order): on an older store, a refusal the pre-open evaluation
  already shows comes before the open. Carried as the new 6d-R15, which Task 9 applies.
- No erratum is owed to the configuration homes spec or the re-plan. R7's routing restates CH's
  schedule, and R8's aside name is the re-plan's own.

## Measures

**New-mechanism findings: 5 folded, 1 refused.**

| Finding | Disposition | Source that runs it | Measured defect it answers |
| --- | --- | --- | --- |
| R4 (a lock-held keep read) | Refused | none named | none reachable at 6d; routed |
| M1 (the cmd older-store helper) | Folded | `restore_test.go:747`'s `storeFile`, the repo's own lowering helper | the mechanics probe: re-migrating a `storeFile` store fails with "table member_overrides already exists" |
| C1 (`doctor`'s unclaimed-error line) | Folded | `doctor`'s own print-on-failure roster row (`doctor.go:87-114`) | the trace: no `doctor` row prints a superseded key's message (`doctor.go:51-201`, `:369-391`) |
| O1 (the state-dir leg as a script) | Folded | `scripts/vale-comments.sh`, called from `check.sh:514` | the leg is inline (`check.sh:657-737`), so each of six states would cost a full heavy-lane gate (about 106s each) |
| C10 (the command-tree walk) | Folded | `TestDoctorCmd_LineCountFollowsRegistry`, which derives its count rather than hand-listing it | `newRootCmd` holds two non-loading verbs (`main.go:126-135`), so a hand-picked list passes when a verb is missed |
| R5 (the pre-open refusal) | Folded | R-6d4's own pre-open `Inspect` and `EvaluateFacts`, whose result it reuses | the trace: after Task 6, an auth refusal moves after the migration, stranding the rollback binary on a newer store |

**Plan line count:** 1694 before, 1856 after (+162). The growth is Task 1's two state tables and its
cited-tests paragraph, the corrected runbook bullets, `doctor`'s line, the reset-server criterion, and
the helper's description. About eight cmd subtests and a re-executed child are cut from the build,
which the line count does not show.

**Ceiling:** unchanged at 10.5M, flag 8.4M (and the variants). The fold moves no task's rate (the plan's
"Token ceiling" says why). Task count unchanged at ten; deliverables unchanged at 28 across Tasks 1 to 8.

## Second fold

The verification read of revision 2.4 (`2026-10-06-rung-6d-plan-fold-verification.md`) returned one
major and four minors. This fold takes all five, and the plan moves to revision 2.5. The ceiling,
the task count, and the rulings are unchanged.

| Finding | Disposition | Evidence |
| --- | --- | --- |
| V1 (major) | Folded. Task 1's foreign-owner start row is dropped, and its cited-tests paragraph gains `TestOpenModes_OwnerAndFilesystemRefused`. Task 4's criterion cites the same test, and the verbs' mapping rides their existing `OpenWrite` refusal rows. | `leaf.go:63` holds the store's unexported `geteuid`; cmd's own is `unidecode.go:46`. `open_test.go:622-662` has a "leaf owned by another uid" row through `withGeteuid`, and it asserts refusal from both `OpenWrite` and `OpenServe`. |
| V2 (minor) | Folded. `Report.Version` reads `Report.StoreVersion` in all five places. | `inspect.go:93-98` |
| V3 (minor) | Folded. R-6d21 says that from Task 6, `serve`'s pre-open check reads its result below `0004`. R-6d4 qualifies the sentence by version. R-6d2 names a `0001` store among the states with no facts. | `evaluate.go:145-150`, `inspect.go:240-242` |
| V4 (minor) | Folded. The appendix gains "Task 1 under no", which drops the reconciler criterion and mutation. Task 3's block adds `member_imports` to the helper's drop list. The conductor's action 1 reads "Tasks 1 and 3 to 5". | plan Task 1 criteria, appendix |
| V5 (minor) | Folded. The script call stays inside the non-reduced block, so `--reduced` still skips the leg. | `check.sh:626-737` |

No item is owed from this fold.
