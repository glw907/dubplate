# Rung 6d plan fold verification: revision 2.4

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md`, revision 2.4 at `dec779d`
(pass class `auth-data`). Fold record: `2026-10-06-rung-6d-plan-fold.md`. Reader: Opus 5.5, high
effort, no part in the fold. Findings are correctness gaps only, ranked by consequence.

**Verdict: needs second fold (narrow).** One major, four minors. Every blocker and major the fold
lists closed; the major below is a fixture mechanism the fold kept and named from memory.

## Question 1: did each blocker and major close?

| ID | Closed? | Evidence |
| --- | --- | --- |
| C1 | Yes | `doctor.go:48-51` prints `statuses[row.Name]` only for its own rows; `rowStatuses` (`:369-391`) keys on the text before `": "`, and `rejectNamespace` writes `KEY: message` (`config.go:312-313`). Task 5's unclaimed-error line closes the silent exit. Criterion 4 and a mutation cover it. |
| C2 | Yes | Task 4's own `rosterResetServer` test: the first POST answers 2, the second answers 3 under the named set. A first-POST snapshot holds 2 and answers 2 on the second POST, so the mutation dies. |
| M1, O2, R10 | Yes, except V1 | The keep decision is a pure function with a table. The store tests Task 1 cites exist. The older-store helper works (probe below). The foreign-owner row is V1. |
| M2, R3 | Yes | R-6d18's settings-rows bullet and Task 8's render-and-read both target runbook `:340-345`, whose "not yet true" text I confirmed. |
| R1 | Yes | `git show d1e93e9:cmd/dubplate-server/member.go` line 101 registers `--streamer-username`. The rollback step moves `roster.json.pre-username` back and names no `member add`. |
| R2, m1 | Yes | `reconcile.go:167` reads "the users table is empty: run member import". Task 1's Outcomes, criterion, and mutation reword it under Ruling 4's yes. V4 covers the no branch. |

**Probe: the older-store helper (Q3).** In a scratch copy (`git archive HEAD`, never the repo tree),
a test in `internal/engine/store` did the following: created a store with `OpenServe`, dropped the
tables each migration above the target created, and ran
`DELETE FROM migrations WHERE version > N; PRAGMA user_version = N`. It then called `Inspect` and
`OpenServe` again:

```
4: inspect state=older version=4 facts=true  -> reopen version 5, snapshots [pre-migrate-v4.db]
3: inspect state=older version=3 facts=true  -> reopen version 5, snapshots [pre-migrate-v3.db]
2: inspect state=older version=2 facts=true  -> reopen version 5, snapshots [pre-migrate-v2.db]
1: inspect state=older version=1 facts=false -> reopen version 5, snapshots [pre-migrate-v1.db]
```

A populated `jobs` row was inserted before the cut at targets 2 to 4. The helper mechanism is sound
at all four versions the plan uses. The `0001` store carries no facts, which matches R-6d13.

**`doctor`'s unclaimed-error line (Q3)** is grounded in source, as cited under C1. The command-tree
walk (C10) matches `main.go:126-135` and the seven `config.Load` callers.

## Question 2: contradictions and build order

**Rulings numbering is consistent.** The Rulings section poses 1, 3, and 4. It says that revision
2.2's Ruling 2 is now R-6d20 and that "its number is not reused" (plan `:100-102`). The labels
paragraph says "Ruling N" alone means this plan's ruling (`:69-70`). Every bare "Ruling 1", "Ruling 3",
and "Ruling 4" matches that section. Each "Ruling 2" is qualified as revision 2.2's (`:102`, `:827`) or
6c's (`:820`, `:1476`, `:1664`). "Ruling 4's yes" refers to the third of the three questions, which
is numbered 4. No reference is stale.

The task order builds: Task 1 creates the helper, and Tasks 2, 4, and 6 then reuse it.

## Findings

### V1 (major): the foreign-owner rows name a `geteuid` override that `cmd` cannot reach

**Location:** plan `:1057` (Task 1's start table: "the state directory owned by another uid (the
`geteuid` override)"). Plan `:1310-1311` has the same defect (Task 4: "the foreign-owner row uses the
`geteuid` override `leaf.go:57-63` provides"). The Task 4 row predates the fold, and the fold kept the
Task 1 row ("and the foreign owner", fold record, M1 row).

**Defect.** `leaf.create` and `checkOwner` read `internal/engine/store`'s unexported
`var geteuid = os.Geteuid` (`leaf.go:63`). `cmd/dubplate-server` has its own, separate `geteuid`
(`unidecode.go:46`), which `doctor.go:294` and `unidecode.go:265` use. Swapping the cmd variable never
reaches `OpenServe` or `OpenWrite`. Neither row can be built as specified without one of two moves.
The implementer could export a test seam from `store`, but nothing is exported ahead of a caller. Or it
could run against a root-owned directory, which fails when the gate runs as root. Either way the
result is a fix round or a new surface. The mechanics review proposed citing the store test instead
of keeping the row. That test, `TestOpenModes_OwnerAndFilesystemRefused`
(`open_test.go:622-662`), already asserts the refusal through both `OpenWrite` and `OpenServe`.

**Proposed fold.** Drop the foreign-owner row from Task 1's start table. Add
`TestOpenModes_OwnerAndFilesystemRefused` to "What this task proves, and what it cites"; the `newer`
row already proves the refusal-to-no-shell mapping. In Task 4, cite the same test for the member
verbs' owner row in place of the override. Task 4 keeps only a mapping assertion, which any
`OpenWrite` refusal it already has (older or newer) proves.

### V2 (minor): `Report.Version` is not a field; the field is `Report.StoreVersion`

**Location:** plan `:715`, `:854`, `:1300`, `:1314`, `:1492` (C7's fold).

**Defect.** `store.Report` carries `BinaryVersion` and `StoreVersion`
(`inspect.go:93-98`). It has no `Version` field. The rule is right, but the name is wrong in five
places that an implementer and a reviewer will grep.

**Proposed fold.** Replace `Report.Version` with `Report.StoreVersion` throughout.

### V3 (minor): R-6d21 says it changes no `serve` behavior, but R5's fold makes `serve` depend on it

**Location:** plan `:858` ("it changes no stored row and no `serve` behavior"), against `:497-503`
(R-6d2's pre-open check "over the stored rows or (from Task 6) R-6d21's empty row set") and `:546-547`
(R-6d4: "an unparseable env value for the keep ... On an older store, it refuses through R-6d2's
pre-open check").

**Defect.** These three passages disagree. Task 6 changes `EvaluateFacts` (Files: `evaluate.go`), and
`serve`'s keep read calls it, so after Task 6 the pre-open check reaches `0002` and `0003` stores. That
is a `serve` behavior. Before Task 6, `evaluateTableAbsent` "never reads environ"
(`evaluate.go:145-150`), so R-6d4's sentence is false below `0004` until Task 6 lands. At `0001` (no
facts) it stays false, because R-6d2 exempts states with no facts while R-6d21 evaluates them in
`check` and `doctor`. No such store exists anywhere, so this is a text defect, not an exposure.

**Proposed fold.** At `:858`, say that `serve`'s pre-open check also reads R-6d21's result on a store
with facts below `0004`, from Task 6. At `:546`, qualify the sentence with "at `0004`, and from Task 6
at `0002` and `0003`". Add one sentence to R-6d2 saying that a `0001` store, which has no facts, gets
no pre-open check. No test row is owed.

### V4 (minor): Ruling 4's "no" path misses two Task 1 consequences

**Location:** plan `:1701-1703` (the conductor folds the appendix "into Tasks 3 to 5"). Also Task 1's
reconciler criterion and mutation (`:1080-1081`, `:1092`), and the helper (`:1030-1039`).

**Defect.** The criterion and the mutation that require the reconciler line to name `member add
ADDRESS --admin` are written without a condition. Under the no answer the line stays (6d-R11), so the
criterion contradicts the reading. Under no, Task 3 also adds migration `0006`, and the helper's drop
list then lacks `member_imports`, so every helper-built fixture fails "table already exists" from
Task 3 onward. The failure is loud, not vacuous, but no task owns the extension.

**Proposed fold.** In the appendix, add a "Task 1 under no" note: the reconciler criterion and
mutation drop, and Task 3 adds `0006`'s table to the helper's list. In the conductor's action 1,
change "into Tasks 3 to 5" to "into Tasks 1 and 3 to 5".

### V5 (minor): the state-dir script's call site changes reduced mode

**Location:** plan `:1580-1583` ("`check.sh` calls it unconditionally, as it calls
`scripts/vale-comments.sh`").

**Defect.** Today the leg sits inside `if [ "$mode" != "reduced" ]` (`check.sh:626-737`), together
with `systemd-analyze verify`. An unconditional call makes `--reduced` run a leg it skips today. The
leg reads no variable set outside its block, so the move itself is sound.

**Proposed fold.** Say that the call replaces the leg inside the existing non-reduced block.

## Question 3: new mechanisms stated from memory

- The older-store helper is proven by the probe above.
- `doctor`'s unclaimed line is quoted from `doctor.go` and `config.go`.
- The script move is grounded in `check.sh:514`, with V5's nuance.
- Task 2's release row ("an exclusive open ... succeeds") holds. `OpenRestore` takes the lock
  "exclusively without waiting, refusing a held lock" (`restore.go:96-102`).
- The `geteuid` override (V1) is the one mechanism stated from memory, and it is wrong.

## Question 4: R7's routing and R4's refusal

**R7 is accurate.**

- CH `:576-583` reads, in part, "The removal verb and this data cleanup file at rung 14", with the
  user row "tombstoned with a removal time".
- CH `:732` lists "member removal as the checklist above" in rung 14's admin spine.
- The re-plan `:108` reads "Rung 14 is unchanged except its private gate".
- STATUS's slice 6 names "offboarding deprovision". No demote or remove path exists in the tree.

**R4's refusal is sound.** Every store a 6d binary creates is at its own last migration. The box has
no store before the 6d upgrade's step 2. So a busy or intent pre-lock read during a migration cannot
occur until the next migration lands, and Task 9 routes the limit to that migration's plan. This
still holds under Ruling 4's no, where `0006` also lands inside 6d.

## Question 5: can Geoff rule on the evidence given?

Yes. Each of Rulings 1, 3, and 4 states a recommendation, reasons, and what yes and no build, and each
carries a ceiling variant.

- Ruling 1's basis is 6c's actual rate.
- Ruling 3 cites the re-plan's `:422` placement.
- Ruling 4's premise checks out at `roster.go:340`
  (`json.Unmarshal(data, &file, json.RejectUnknownMembers(true))`) and in the re-plan's roster step
  `:323-331`.

No ruling depends on V1 to V5.

## Counts

Blockers 0, majors 1 (V1), minors 4 (V2 to V5). **Verdict: needs second fold.** It is a narrow fold
of V1, with the minors folded alongside at the fold agent's discretion.
