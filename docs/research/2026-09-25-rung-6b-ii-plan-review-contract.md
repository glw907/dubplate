# Rung 6b-ii plan review: contract and criteria lens

Reviewed: `docs/plans/2026-09-25-rung-6b-ii-restore-http.md`, revision 1 (uncommitted).
Lens: are the promises testable? Does each criterion name its fixture and its failure reason? Can a
criterion or a named mutation pass vacuously? Checked against spec
`2026-09-23-rung-6-ops-seam-store.md` (6b-ii criteria `:1035-1060`, restore flow `:674-714`,
contract suite `:600-605`), the CLAUDE.md planning-miss rules, and 6b-i's code in
`.claude/worktrees/lane-2-6b-i` at `33ceff5`.

Counts: **0 blockers, 10 majors, 16 minors.** One owner fork (M10).

The plan is strong on state tables and names its fixtures more consistently than most. The majors
fall into three groups. Some fixtures contradict the code they run against (M1, M2, M7). Some
proofs or mutations cannot fail (M3, M4, M5). Some promises have no assertion (M6, M8, M9, M10).

---

## Majors

### M1. The `canceled` contract fixture contradicts the store's documented write contract

- **Location:** plan `:835` (the contract row), `:405-409` (R-6bii5); against
  `internal/engine/store/write.go:51-54` in the worktree.
- **Defect:** the row says to cancel the caller's context "while the call waits" on a write lock
  that a second process holds. `Handle.Update` documents the opposite: "The wait on another
  process's write lock is bounded by the busy timeout alone: a deadline shorter than the busy
  timeout does not shorten it." After the busy timeout, the error is `errBusy`, and `classifyErr`
  wraps it with `%v` (`conn.go:221-222`). So the chain holds no `context.Canceled`, and the
  classifier answers `internal`. The case fails, or someone "fixes" it by loosening the classifier.
- **Fold:** use the wait that honors ctx instead. That wait is for the write pool's one connection
  in this process (`write.go:50-52`: "ctx ... bounds the wait for the write pool's one connection
  inside this process"). Fixture: a test goroutine holds `Handle.Update` open on the runner's own
  handle. A `cancel_job` call then waits for that connection, and the test cancels its context.
  Quote `database/sql`'s connection-wait branch (`sql.go`, `conn`: `case <-ctx.Done(): ... return
  nil, ctx.Err()`) from go1.27.1 source. State that the HTTP leg runs through
  `httptest.NewRecorder` with the request context canceled. Over `httptest.Server`, the client sees
  only its own context error and never reads a 503, so that leg would be vacuous.

### M2. The blocking job body is left as a mid-pass conductor ruling, but the answer is already knowable

- **Location:** plan `:841-847`, "Not verified" `:1374-1375`; against
  `internal/ops/local/local.go:76-101`.
- **Defect:** `bindBackup` and `open` are unexported. `opshttp`'s contract test cannot reach them,
  and spec `:604` forbids "a test-only constructor." The plan's own candidate is to run the real
  backup through `local.New` with a blocking rclone stand-in found through `PATH`. That is how
  `internal/backup/standin_test.go` already drives `ops/local` (`backup.Job` resolves
  `exec.LookPath("rclone")`, `backup.go:69`). Leaving this open puts an ambiguity after approval
  that planning can close now. Three rows depend on it: `conflict`, criterion 3, and criterion 4.
- **Fold:** rule it in the plan. Task 3 uses the stand-in pattern, installed on `PATH` in
  `TestMain`, beside the re-exec dispatch. Criterion 3's "the body observed its context done"
  becomes "the stand-in recorded one SIGINT", as in `standin_test.go`. Criterion 4's child runs the
  same stand-in. Delete the item from "Not verified."

### M3. The sentinel-coverage guard has no stated enumeration source, and the obvious one is vacuous

- **Location:** plan `:824`, `:852`, mutation `:857`; against `internal/ops/errors.go:36-42`
  (`codes` is unexported).
- **Defect:** spec criterion 1 says the suite "fails when a sentinel has none." The `opshttp` test
  cannot read `ops.codes`. A guard that iterates a hard-coded list in the test still passes when an
  eighth sentinel is added to `ops`, so the guard only proves the list matches itself. Mutation 1
  (delete the `internal` case) also passes against a hard-coded list if the list is edited along
  with the case.
- **Fold:** `ops` exports the code set, for example `ops.Codes()`, with a production caller: the
  projection's code-to-status table checks itself at construction and fails on a code it does not
  map. That caller is real and needs no exception under exit gate 13. The guard iterates
  `ops.Codes()`. Add the mutation "add an eighth sentinel to `ops` with no contract case → the
  guard fails, naming it."

### M4. The privilege agreement test is circular, and mutation 3 survives

- **Location:** plan `:709`, `:716-717`, mutation `:728`.
- **Defect:** R-6bii1 makes `Catalog` build its entries from the lookup, so the test compares the
  lookup with itself. If one entry is dropped, both sides go missing, or `Catalog` errors, depending
  on how it handles the gap. Nothing fixes which. A mutation that turns `start_backup` into a
  member operation passes every Task 1 row. This table is the authorization source for the whole
  projection. The repo bar ("no expectation derived from the code under test") applies here.
- **Fold:** the lookup row asserts literal values: `start_backup`, `job_status`, and `cancel_job`
  are admin, and `catalog` is member. Cite 6b-i's catalog as shipped (`local.go:121-129`, before
  R-6bii1 moves it). Add a Task 3 row: a member-only actor calling `start_backup` gets `forbidden`
  on both legs. Replace mutation 3 with "flip `start_backup` to member → the literal row and the
  member-forbidden row both fail." State what `Catalog` does with a name the lookup lacks: a
  construction error, never a zero privilege.

### M5. Criterion 6's mutation 2 survives, and the fixture does not prove the rows are WAL-only

- **Location:** plan `:980-983`, mutation `:999`.
- **Defect:** skipping `wal_checkpoint(TRUNCATE)` loses nothing. Leaving WAL mode checkpoints on
  its way out, and "When the last connection to a database closes, that connection does one last
  checkpoint" (the plan's own quote, `:488`). The pre-restore copy still holds the three rows, so
  the mutant lives. The rows are lost only when step 3 is skipped entirely: the database file is
  renamed while `db/dubplate.db-wal` stays behind. That leftover WAL would then sit beside the
  swapped-in file, which is the worse outcome. Also, "`-wal` is non-empty" does not prove the rows
  are absent from the main file.
- **Fold:** add the fixture proof. Before the restore, copy the bare `db/dubplate.db` (no WAL) to a
  temp path and open it read-only. The three rows must be absent there. Replace mutation 2 with
  "skip step 3 and the leftover check → the pre-restore copy lacks the rows and `-wal` remains in
  `db/`." Keep the explicit checkpoint for its busy column (M-minor 3), and say that is why it is
  there.

### M6. Criterion 13 is never asserted on the path where it can fail

- **Location:** plan `:990` (4B asserts mode on a `--from-file` stage it pre-creates at 0600),
  `:1052` (Task 5 only "reports" the mode of the file `copyto` writes), `:527-529`.
- **Defect:** the one stage whose mode is not under the store's control is rclone's rename-into-place
  on the `--snapshot` path. No criterion asserts 0600 after a `--snapshot` restore. 4B's mutation 5
  runs against a stage the test itself makes 0644, which never exercises the rclone path.
- **Fold:** Task 5 adds a criterion: a `--snapshot` restore through real rclone with umask 022, and
  the swapped-in `db/dubplate.db` is 0600 per `os.Stat`. Mutation: skip the chmod. If the measured
  rclone mode is already 0600, the mutant survives. Report that and declare it.

### M7. Ruling 5's reference rewrite has no channel the restore can reach

- **Location:** plan `:108-117`, `:312`, `:1028`; against `internal/engine/store/snapshot.go:366`.
- **Defect:** `WriteBackupRef` is a `*Handle` method. A `Handle` comes from a write-mode or
  serve-mode open. Write mode refuses beside an intent record (`open.go:365-376`), and restore
  holds the exclusive lock. Task 5's file list (`:1014-1018`) excludes `internal/engine/store`, so
  it cannot add the method. Nothing states whether the reference is written before or after the
  record is removed, or which crash row covers it. The privileged-field rule requires every
  channel to be named.
- **Fold:** 4B's restore session gets the reference write, through the same body and prune as
  `WriteBackupRef`, factored so the two share code. Name it in the privileged-field row. Place it
  between the swap and step 7, while the record still exists, so resume redoes it idempotently.
  Add the Task 5 row "crash after the reference write → resume completes, one reference names the
  restored stamp."

### M8. Resume after a partial library copy has no row, and it can deadlock

- **Location:** plan `:453` (`library`: "re-runs step 6's copy"), Task 5 crash table
  `:1032-1040`, Task 7 signal row `:1147`.
- **Defect:** a crash or SIGINT during `rclone copy` leaves the library root non-empty. If resume
  repeats step 6's emptiness check, every `--resume` refuses unless `--into-nonempty` was recorded.
  Resume takes no flags, so the operator is stuck. The plan never says whether resume re-checks.
  Task 7's signal row stops at "exit 130; record `library`" and never resumes.
- **Fold:** state in R-6bii7 that resume skips the emptiness check because the record proves the
  root was checked. Add a Task 5 or Task 7 row: interrupt during step 6 with a stand-in that has
  written one file, then `--resume` completes and the tree matches the remote.

### M9. "`--resume` completes" is never defined as an end state

- **Location:** plan `:1044-1045` (criterion 7), `:1035-1039`.
- **Defect:** each crash row asserts that resume "completes" but names no observable result. A
  resume that returns nil after removing the record passes. This is the pass's data-destructive
  path.
- **Fold:** define one completion tuple and assert it in every resume row. `db/dubplate.db` holds
  the incoming snapshot's rows. `snapshots/pre-restore-<stamp>.db` holds the prior store's rows.
  `staging/` is empty. No `-wal` or `-shm` sits beside the pre-restore copy. The report file exists
  (or the record is `reconcile-pending`). The record is gone. The reference follows Ruling 5.

### M10. Review focus 8 promises a refusal that no task builds (OWNER FORK)

- **Location:** plan `:627-628` ("A library root that is an unmounted mount point. Expected:
  refused, never populated. Pinned in Task 6."), `:568-573`, Task 6 table `:1091-1097`; against
  `internal/backup/guard.go:23-42` (no mount detection).
- **Defect:** an unmounted mount point is an existing, empty directory. R-6bii11 refuses only an
  absent root. An empty root is the valid target for criterion 12, so restore would copy the whole
  library onto the parent filesystem. No Task 6 row covers this case. The reverse case also fails:
  a freshly formatted ext4 volume's root holds `lost+found`, so it reads as non-empty and needs
  `--into-nonempty` for its first restore.
- **Owner fork: what should a restore into a mount point do?**
  - (a) Detect a mount point with the parent's `st_dev` equal to the root's, and refuse an empty
    root on the same device as its parent unless a flag is given. This is precise for the unmounted
    case, but it refuses a legitimate library directory on the root filesystem without the flag.
  - (b) Refuse nothing new. Print the resolved root, its filesystem, and free space before the copy
    (R-6bii10 already reads free space), and rewrite focus 8 as a declared gap.
  - (c) Option (b), plus ignoring `lost+found` in the emptiness check.
  - **Recommendation: (c).** It keeps criterion 12 literal, costs one line of output and one
    exemption, and states the gap honestly. Option (a) is worth doing only if the box's library
    will sit on a separate mount at launch. That is a rung 16 fact the plan does not have.

---

## Minors

1. **Crash rows miss step 3, and criterion 7's literal wording cannot hold for steps 1 to 3.**
   (`:1034`) No record exists before step 4, yet spec `:1051` says each crash "leaves the intent
   record." Declare it as D7. Add the row "after step 3: absent; the current store is in DELETE
   mode; a write-mode open proceeds and reads WAL." Name the hook point between step 5's two
   renames, since the spec's hook sits between steps. Replace "`swap` or its successor" (`:1037`)
   with the one state the flow writes. Name the "after step 1 or 2" fixture's current store as
   present, since with an absent leaf, write mode answers "start serve first," not "proceeds."
2. **Criterion 5 tests a writer, and the spec says a reader.** (`:978`) Add an inspect-mode holder,
   the literal "real reader" that `doctor` and `check` use.
3. **The "reader outside the lock" fixture state decides which refusal fires.** (`:968-969`) An open
   read transaction makes the checkpoint busy. An idle connection leaves `-shm` behind. Name both
   fixtures, one per row.
4. **The actor case's body names `job_id`, but the actor column is written by `start_backup`,
   which has no fields.** (`:837`) Use `{"actor":"local:root"}` and `{"ACTOR":...}` on
   `start_backup`, and assert `jobs.actor`.
5. **The channel-mismatch fixture has only one possible form.** (`:772`) `ChannelCookie` is the only
   network channel, so the mismatch has to be an authorizer returning a `LocalActor`
   (command-line). Say so.
6. **4A's mutation 4 depends on a fixture that does not exist yet.** (`:921-922`) It points to
   "Task 4B's criterion 8 fixture," but criterion 8 is Task 5's and 4A runs first. 4A builds its
   own fixture: a hand-written `swap` record plus a staged file, then a serve open and a write open,
   after which the file is still present.
7. **4B's mutation 4 names the wrong killer.** (`:1001`) Resume's combination rows use hand-built
   fixtures, so reversing the forward rename order does not touch them. Criterion 6's pre-restore
   assertion kills it instead, because the incoming file replaces the current database.
8. **Task 7's mutation 1 cannot fire.** (`:1164`) In the signal row the record is already `library`,
   so no record write follows the signal. Either name a write that follows a cancel or drop the
   `WithoutCancel` rule (`:599`).
9. **Task 7's end-to-end fixture leaves the library root populated after the backup.** (`:1153-1156`)
   `--library` would then refuse. Name an emptied root. Assert rows by value: the rows changed after
   the backup are absent and the backed-up rows are present.
10. **The zero-byte row names the wrong report.** (`:544`) `verifyCopy` says "An empty file passes
    the integrity check" (`snapshot.go:108-109`). Under R-6bii9's widened version rule, a
    zero-byte file reads version 0 and is refused on `application_id`. Fix the expected report, or
    refuse version 0 explicitly. Also state that the widening is a new parameter and that backup's
    call keeps exact equality, with a test.
11. **Criterion 11's fixture should be the real set's prefix, not a synthetic set.** (`:908-910`)
    Serve then migrates with the embedded set, and version 2 of the real set (files 0001 and 0002)
    holds `jobs`. A synthetic v2 would test migration 0003 against a schema that never ships.
12. **Exit gate 13 leans on the exported-unused leg, which cannot prove it.** (`:1228-1231`) That leg
    passes a symbol referenced from any non-test file in its own package. List every new export with
    a `git grep` for its cross-package caller.
13. **Exit gate 8 omits five tables.** (`:1219-1220`) It leaves out Task 2's, 4B's, Task 5's crash
    table, Task 6's, and Task 7's.
14. **D4 misses the `-journal` sibling.** (`:150-152`) A `--from-file` source with a hot rollback
    journal beside it is torn the same way. Refuse it too.
15. **R-6bii5 defers a behavior it can quote now.** (`:411-413`) `database/sql` returns `errDBClosed`
    ("sql: database is closed") from a closed `*DB`, and `classifyErr` passes non-sqlite errors
    through (`conn.go:216-218`). Quote both.
16. **Smaller items.** The 4 KiB cap has no boundary pair: add 4096 accepted and 4097 refused. The
    uid half of criterion 13 cannot fail under the test's own uid: declare it. A corrupt intent
    record is refused by every mode and by resume, which is a dead end. Its report should name the
    manual recovery, which is to inspect the named files and remove the record.

---

## The six rulings: can each be ruled on this evidence?

1. **Ceiling and split.** Rulable. The arithmetic is shown, and the split point shares no file.
2. **`canceled` meaning.** Rulable on the semantics. Whichever way it is ruled, the "yes" fixture as
   written cannot pass (M1). That flaw is a plan fix, not a reason to answer no.
3. **403 over 404.** Rulable for a member, and the catalog argument holds. The evidence says nothing
   about a caller the authorizer cannot identify at all. That caller never saw the catalog, so a 403
   tells them a route exists. Spec criterion 2 needs only the code `forbidden`, not the status.
   Recommend yes, with a sentence that the unauthenticated answer (401 or 404) is rung 14's, where
   the mount and the production authorizer land.
4. **Exit codes.** Rulable. It is consistent with `backup`.
5. **Reference follows the restore.** Rulable on meaning, but "yes" cannot be built until M7 names
   the channel. State one more consequence: with `--older`, the reference then names an older
   stamp than the remote's newest.
6. **`rclone copy` for `--into-nonempty`.** Rulable, with two fixes. My recollection is that rclone
   compares size and modification time, and on a modtime-only difference with matching hashes
   updates the modtime instead of copying. I did not check this against v1.74.3's help, so quote it
   before the ruling text claims "replacing any file whose ... modification time differs."
   Add a third option: `rclone copy --update` ("Skip files that are newer on the destination"). It
   keeps a local file a member edited after the backup. Recommendation stays plain `copy`, since
   the backup is the record, but the owner should see the option.
