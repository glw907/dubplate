# Rung 6b-ii plan fold verification (fold 1)

This is the independent verification read of the 6b-ii plan's revision 2
(`docs/superpowers/plans/2026-09-25-rung-6b-ii-restore-http.md`, committed at `7fd9e78`). It checks
the plan against its fold record (`2026-09-25-rung-6b-ii-plan-fold.md`) and the three lens reviews.
The reader had no part in the fold. Code was read on master, where 6b-i is now merged. Quotes were
checked against go1.27.1 and the installed rclone v1.74.3. The architecture reads
(`2026-09-25-rung-6b-i-architecture-reads.md`, uncommitted working copy) were read for question 4.

**Verdict: the fold holds, with 0 blockers, 2 majors, and 10 minors.** Every blocker and major the
lenses raised closed at its cited location. The two majors are new, and each needs one sentence.
The first is the step hook's point between step 5's two renames, which no package can reach as
written. The second is the Ruling 6 displaced directory on a `library` resume, which can dead-end
the resume in the same way C-M8 did.

## 1. Did each blocker and major close where the fold says?

| Finding | Closed? | Where it landed, and what was checked |
| --- | --- | --- |
| X-B1 / R-B1, lock-then-recheck | **Yes** | R-6bii7 `:634-646` and Task 4A (three lock-wait rows at `:1222-1224`, mutations 5 and 6 at `:1259-1261`). Checked against master's `open.go`: `OpenWrite` calls `l.stat()` at `:122`, then `acquireShared` at `:140` with no re-read. `OpenServe`'s shared fast path returns at `:219-231`. The fold's routing is sound: on the exclusive path, `settle` calls `checkIntent` first. The second net (a `-wal` or `-shm` beside `db/` in a `swap` record) is R-6bii8's combination row 2. The inspect half was refused on good grounds: `tryLock` never waits. The fixture's determinism is a minor (m1). |
| C-M1 / X-M1, the canceled fixture | **Yes** | R-6bii5 `:542-559`, Task 1 rows `:1012-1014`, Task 3 row `:1165`. Checked: `write.go:50-53` says what the plan quotes. `sql.go:1320-1331` (the expired-ctx check) and `:1370-1375` (`case <-ctx.Done():`) match. `Request.Context`'s "For outgoing client requests, the context controls cancellation" matches. Keying the wrap on `ctx.Err()` is correct, since `classifyErr` wraps with `%v` (`conn.go:222`). |
| R-M1 / X-M2, the fresh-box swap | **Yes** | R-6bii8's table keys on what the record owes (`:742-752`), with two no-copy rows. Task 4C adds mutation 4, and Task 5 adds the fresh-leaf crash row (`:1451`). |
| C-M8 / X-M3, the library resume | **Yes** | R-6bii7's `library` row, R-6bii11 `:838-840` (D9), Task 6's resume rows, mutation 4, and Task 7's signal-then-resume row. See major 2 for the displaced directory, which this fold did not reach. |
| R-M2, the version re-check | **Yes** | R-6bii7 `:655-659` (D10), Task 4C's resume-version row, Task 5's two binary-change rows, and Task 5 mutation 5. |
| R-M3 / X-m7, `--discard-current` ordering | **Yes** | D8, R-6bii8 `:724-731`, both discard rows in the combination table, and Task 4C mutation 5. Checked: `isLocalCopyName` (`prune.go:196-220`) and `isCopy` (`open.go:~503`) never match `discarded-<stamp>.db`. So nothing prunes it, and `besideCopy` never names it. The refusal of a widened `besideCopy` is sound, since the record now exists first. |
| R-M4, the one-device refusal | **Yes** | R-6bii6 `:585-589`, and Task 4B's device row with a helper table test and mutation 4. Checked: `/tmp` is device 60 and `/dev/shm` is 27 on this workstation, as the fold measured. |
| R-M5, a new restore beside a record | **Yes** (the rule); **fork** (Ruling 9) | R-6bii6 `:583-584`, Task 4B's record rows, mutation 3, and Task 7's record row. |
| R-M8 / C-M3 / C-M4, the privilege table and actor channel | **Yes** | R-6bii1 gives the literal values. The Task 1 lookup rows and the separate `Catalog` literal assertion break the circularity. R-6bii2 pins `authorize(ctx, op, p) (ops.Actor, error)`. R-6bii3 discards the caller's actor. Task 2 adds the planted-actor row and mutation 6. `ops.Codes` has a caller in the constructor's table check, and the guard iterates it. Task 3 adds the member-only row. The privileged-field table (`:411-412`) names every channel. |
| C-M5, criterion 6's checkpoint mutation | **Yes** | Task 4C's criterion `:1384-1388` proves the rows are WAL-only with a bare-file copy. Mutation 1 is now "skip step 3 and the leftover check", and the text says why skipping the checkpoint alone is not a killing mutation. |
| C-M6, 0600 on the `--snapshot` path | **Yes** | R-6bii9 `:776-781` records that the path pre-creates no stage and does not rely on `-M`. Task 5 checks this under umask 022 (`:1473-1475`), with mutation 6. |
| C-M7, the reference-rewrite channel | **Yes** | R-6bii6 `:591-599` adds a session method that shares one leaf-level body with `WriteBackupRef`. Checked: `WriteBackupRef` uses only `h.leaf` (`snapshot.go:366-397`). The write happens after the swap while the record is still `swap`, and every resume redoes it (Task 5 `:1438-1439`, `:1449-1450`). Its ownership under Ruling 5's no is a minor (m6). |
| C-M9, the resume completion tuple | **Yes** | R-6bii7 `:661-666`, plus a global constraint (`:388-389`) and Task 5 `:1460`. The tuple does not cover the discard case (m4). |
| C-M2, the blocking body | **Yes** | Task 3 `:1137-1144`. Checked: `backup.go:69` resolves rclone through `exec.LookPath`. |
| C-M10 / X-M4 / R-M7, R-M6 | **Forked** | Rulings 6, 7, and 8. The false claim in focus item 8 is struck (`:918-919`) and declared in "Not verified". |

## 2. Contradictions, build order, deliverable counts, and the split

**Build order.** The chain 4A → 4B → 4C → 5 → 6 → 7 builds, with one exception, major 1. No task has
more than four counted deliverables. Task 5 carries uncounted work: the transport generalization,
the shared listing, and, if m2 is taken, the listing's move onto the transport. It is the pass's
densest task, with 14 crash rows and 7 mutations. That is a sizing note, not a defect.

**The split after Task 3 is clean by file** under both answers of Ruling 1. 6b-ii-a writes
`internal/ops`, `internal/ops/local`, and `internal/ops/opshttp`. 6b-ii-b writes
`internal/engine/store`, `internal/backup`, and `cmd/dubplate-server`. Each half's close is named
in conductor action 3 and "Execution model" A2 and B3: the simplifier round, the security review or
the targeted data-loss read, Task 8, the merge, and three architecture reads each. There is one
coupling without a shared file. Task 3's contract suite runs the real backup body, and 6b-ii-b's
Task 5 edits `internal/backup/rclone.go`, so 6b-ii-b's gate reruns the suite. That is how it
should work. The split's bookkeeping has two small gaps (m8).

**No contradictions between sections were found**, apart from the ruling-conditional wording in m6
and m7.

## 3. New mechanisms: quoted or from memory?

Every new mechanism checked is quoted or probed, and the quotes match their sources:

- **Go stdlib (go1.27.1).** `MaxBytesReader`'s `*MaxBytesError` sentence, `os.CreateTemp`'s "0o600
  (before umask)", `WalkDir`'s "does not follow symbolic links", `File.Sync`, `os.Rename`'s
  "replaces", `Context.Err`, `Unmarshal`'s case-insensitive match and ignored keys,
  `Request.Context`, and `errors.Is`. Also `sql.go:1318` (`errDBClosed`), `:1320-1331`, and
  `:1370-1375`, and `server.go:2751-2762` (the mux's plain-text 405 and `NotFoundHandler`).
- **The store source.** `classifyErr`'s pass-through at `:216-218` and `%v` wrap at `:221-222`.
  `maintenanceDSN` omits `_journal_mode`. `writeDSN` sets `_journal_mode=WAL` at `:133`.
  `verifyCopy`'s empty-file comment. `WriteBackupRef`'s shape. `checkLedger`'s signature.
- **The session's read handle.** An additional check the fold did not make: `readDSN` carries
  `_journal_mode=WAL` and `_query_only=1`. modernc v1.59.0's apply phase runs `journal_mode` before
  `query_only` ("query_only is applied last"). So a read handle on the swapped-in DELETE-mode file
  converts it to WAL and does not fail. Step 7 is safe.
- **rclone v1.74.3.** `rclone help flags` lists `--backup-dir` ("Make backups into hierarchy based
  in DIR"), `--ignore-existing`, and `-u, --update` as quoted. `rclone copy --help` does not list
  `--backup-dir`, as the fold says. The `copy`, `copyto`, `--inplace`, `--partial-suffix`, and
  `size` quotes match. **The plan cites the probe**: Ruling 6 (`:148-152`) points to the fold
  record's "Probes". This reader re-ran it. `dst/a.flac` read `backup`, `displaced/a.flac` read
  `local-newer`, the identical and local-only files were unchanged, and an overlapping backup
  directory was refused with the quoted message. One addition: rclone creates a missing backup
  directory at 0755 under umask 022, so the plan's pre-creation at 0700 is doing real work.
- **Not checked.** The `mklost+found(8)` quote, since the man page is not installed here.

## 4. Does the plan build on 6b-i surfaces the architecture reads found unsound?

- **`prune.list` and `runRclone`.** These launch rclone without a process group, without the SIGINT
  and `WaitDelay` cancel, and without inherited locks. The conductor's routing sends the fix to
  6b-ii, but the plan does not absorb it. R-6bii9 (`:763-764`) and Task 5 (`:1425-1426`) reuse
  `prune.list` as sound for restore's `snapshots/` listing, and R-6bii10's `rclone size --json`
  names no launcher. The transport captures no stdout, so `size` would have to go through the
  unsound path as well. The consequence is small. These calls are read-only, so an orphan writes
  nothing, and `copyto` and `copy` do run through the transport with the exclusive lock inherited.
  What is wrong is observable, though. A SIGTERM during the listing or the size call kills rclone
  with SIGKILL. `list` then returns a message that does not wrap `ctx.Err()`, so the verb exits 1
  where Ruling 4 says 130. This is **minor m2**: Task 5 should absorb the fix.
- **`ops.Seam` and the catalog having no consumer.** The plan does not build on anything wrong here.
  The authorizing layer is `ops.Seam`'s first consumer, and `Runner.Catalog` gains its caller
  through the seam. The other `internal/ops` findings routed to 6b-ii are carried by no task (m3).

## 5. Can the owner rule each ruling on the evidence, and does the plan execute under each answer?

All nine rulings can be ruled on the evidence given. Each states a recommendation, a reason with a
source, and what each answer builds. The plan executes under every "yes". It executes under every
"no" except for the wording in m6 and m7, where a task's rows are written for the yes answer and an
implementer would have to translate them. Ruling 1 carries derivations for both answers. Ruling 9's
"no" names Task 7B with its size.

## Findings, ranked by consequence

### Major

**M1. No package can reach the step hook's point between step 5's two renames as written.**

- **Location:** R-6bii8 `:738-739`; Task 4C outcome `:1361-1362` ("the step hook's named point
  between the two renames"); Task 5 outcome `:1430-1431` ("an unexported function variable called
  after each step and between step 5's two renames"); spec `:283-284`.
- **Defect:** D3 puts the renames inside a store session method and the hook variable in
  `internal/backup`. A `backup` package variable cannot fire between two renames inside one
  `store` method. The 4C implementer then has two options. One is a store-level hook variable, a
  fourth test seam, which the spec's three-seam list forbids and which the fold itself refused
  (R-M4). The other is a design Task 5 cannot call. Task 5's "between step 5's two renames" crash
  row (`:1448`) depends on the answer. This is the planning-miss class that a fix round pays for.
- **Proposed fold:** one sentence in R-6bii8. The session exposes step 5 as separately callable
  steps: the first rename (or the discard's moves), then the second rename with the fsyncs. The
  flow in `internal/backup` calls the hook between them. The hook stays the spec's one variable, in
  `internal/backup`. Task 4C's outcome becomes "step 5 as two session calls", and its combination
  rows stay hand-built fixtures.

**M2. With Ruling 6's displaced directory, a `library` resume can refuse its own directory.**

- **Location:** Ruling 6 `:152-155` ("creates the displaced directory at 0700 before the copy (a
  creation failure refuses before any rclone call) … removes it when the copy displaced nothing");
  R-6bii11 `:840-843`; Task 6's resume rows `:1532-1533`.
- **Defect:** the directory is `<root>.restore-displaced-<stamp>`, and the stamp comes from the
  record, so a resumed step 6 creates the same path. An implementer following the text literally
  calls `os.Mkdir`, which fails with `EEXIST` on resume. That counts as a "creation failure", so
  the resume refuses before any rclone call. The operator is then stuck at `library` with Ruling
  9's manual procedure, which is the dead-end C-M8 was folded to remove. "Removes it when the copy
  displaced nothing" is also ambiguous across a resume. The first run may have displaced files
  while the resumed run displaces none, and removing the directory then loses the kept files if
  anything but an empty-directory remove is used. Task 6's resume row uses a stand-in, so it never
  runs `--backup-dir`. It also does not say `--into-nonempty`. A further ambiguity: Ruling 6 asks
  about `--library --into-nonempty`, while R-6bii11 and Task 6 add `--backup-dir` to every library
  copy.
- **Proposed fold:** in R-6bii11, state three rules. `--backup-dir` applies with `--into-nonempty`
  only, or state that it applies to every library copy. A resume in `library` accepts an existing
  displaced directory, checking that it is a directory owned by the caller at 0700. The directory
  is removed only when empty, by `os.Remove`, after the copy completes. Add a Task 6 row: an
  `--into-nonempty` restore interrupted after displacing one file resumes, keeps that file, and
  completes.

### Minor

- **m1. The lock-wait fixture has no deterministic "past stat" signal** (Task 4A `:1233-1237`).
  "The test asserts it has not returned" does not prove the open is past `stat`. If the goroutine
  has not yet reached `stat`, the pre-lock read refuses. The row then passes under mutation 5,
  which survives by timing. Serve has a signal already: `waitLock` logs the holders on its first
  busy attempt (`open.go:312-313`), so a recording logger works. Write mode's `acquireBounded` has
  none. The fold: serve rows gate on the holders log line; the write-mode row uses a settle delay
  and reports mutation 5's kill rate across repeated runs, or is named a reviewer-read.
- **m2. The restore listing and `rclone size` ride the unsound launch path** (section 4). The fold:
  in Task 5's `rclone.go` and `prune.go` edits, take the architecture read's fix. `transport.run`
  takes a stdout writer, `prune` holds a transport, and `runRclone` is deleted. The listing, `size`,
  and the fetch then share one launch path. Add a Task 7 row: SIGTERM during the `snapshots/`
  listing exits 130.
- **m3. The other architecture-read findings routed to 6b-ii are carried by no task.** Task 1 must
  edit `TestImportsStandardLibraryOnly` ("lands with 6b-ii"), since `internal/ops` gains
  `internal/identity`. Its new sentinels resolve `errors.go:21-23`'s dangling doc links. The
  four-times-written sentinel list becomes collapsible once the classifier reads `Code`. Name the
  first two in Task 1 so the diff-reviewer expects them, and route the third to the cleanup pass
  explicitly. The exit-gate claim that `ops.Codes` and the classifier have a "production caller"
  is also weak, since that caller is the unmounted projection. Say so in exit gate 14 instead of
  claiming reachability.
- **m4. The completion tuple leaves out the discard** (R-6bii7 `:661-666`). With
  `--discard-current`, no pre-restore copy exists, and the tuple should require
  `snapshots/discarded-<stamp>.db`, plus any `-wal` or `-shm` that existed. Otherwise Task 4C's
  discard rows can finish on a nil error and an absent file.
- **m5. Resume's version re-check can re-create the stage's `-wal` and `-shm`** (R-6bii7
  `:655-658`; R-6bii6 `:613-616`). An inspect-DSN open of a WAL-header stage creates both files,
  which the mechanics lens measured. If resume opens the stage without the remove-after-verify
  step, the swap leaves `staging/restore-<stamp>.db-wal` behind, and "`staging/` is empty" fails.
  The fold: resume's re-check calls the same verify-then-remove session method.
- **m6. Task 4B builds the session's reference write unconditionally** (`:1294-1296`). On Ruling
  5's no, that is an export with no caller, which exit gate 14 forbids. Condition it on Ruling 5's
  yes, as Task 5 does.
- **m7. Some rows are written for the yes answer only.** Under Ruling 2's no, Task 1's "seven
  codes" and `Codes` rows (`:1006`, `:1018`) become six, and R-6bii4's `canceled` status row
  goes. Under Ruling 3's no, Task 3's mutation 3 ("map `forbidden` to 404") is the intended
  behavior, not a mutation. Under Ruling 7's no, the new flag has no name and no place in Task 7's
  flags deliverable, and criterion 12's empty-root fixture needs it. A sentence per ruling in the
  owning task closes each.
- **m8. The split's bookkeeping.** The exit-gate split (`:1667-1668`) gives items 10 (the
  `opshttp` and `ops` import walls) and 14 (the exports list) to 6b-ii-b. Both cover 6b-ii-a's
  surface, and item 14 must also list 6b-ii-b's exports. Give item 10 to 6b-ii-a, and item 14 to
  both halves over their own exports. Both halves' Task 8 also write the same draft paths
  (`draft-STATUS-6b-ii.md`, `draft-HISTORY-6b-ii.md`). Suffix them `-6b-ii-a` and `-6b-ii-b`.
- **m9. Task 7's "signal, then resume" row mixes fixtures without naming the switch** (`:1590`).
  The interrupted run needs the blocking stand-in, and the resume needs real rclone to make "the
  tree matches the remote" hold. Name the `PATH` switch. Separately, a real rclone killed with
  SIGKILL during the library copy leaves `<name>.<hex>.partial` files under the library root,
  which were measured for `copyto`. The report then lists them as files without a row. Add that
  to Task 8's runbook lines, or clear `*.partial` under the root on a `library` resume.
- **m10. Task 5's deliverable count understates its load.** It counts four deliverables but also
  carries the transport generalization and the listing share, plus m2 if taken. Either count them
  under "the fetch" explicitly, or give the conductor a pre-named re-dispatch split: the flow and
  resume, then the fetch and the reference.
