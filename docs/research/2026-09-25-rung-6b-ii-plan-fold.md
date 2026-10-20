# Rung 6b-ii plan fold record (fold 1)

This record folds the three lens reviews of the 6b-ii plan's revision 1 into revision 2
(`docs/plans/2026-09-25-rung-6b-ii-restore-http.md`). Each finding ID has exactly one
disposition: folded (with where it landed), refused (with a one-line reason), or owner fork (with
the ruling it became). A convergent finding is folded once, and its row lists every ID. Before
folding, the fold checked each finding against the lane 2 worktree at `6f2c3e0` (6b-i's close
drafts on top of `33ceff5`), the go1.27.1 source, and rclone v1.74.3. No finding was folded on the
reviewer's word alone.

Inputs:

- Contract lens (`2026-09-25-rung-6b-ii-plan-review-contract.md`): 0 blockers, 10 majors, and 16
  minors, plus notes on Rulings 2, 3, 5, and 6. IDs C-M1 to C-M10, C-m1 to C-m16, and C-R2, C-R3,
  C-R5, and C-R6.
- Mechanics lens (`2026-09-25-rung-6b-ii-plan-review-mechanics.md`): 1 blocker, 4 majors, and 8
  minors, all probe-backed. IDs X-B1, X-M1 to X-M4, and X-m1 to X-m8.
- Risk lens (`2026-09-25-rung-6b-ii-plan-review-risk.md`): 1 blocker, 8 majors, and 13 minors,
  plus a citation spot-check and a pass-sizing fork. IDs R-B1, R-M1 to R-M8, R-m1 to R-m13,
  R-cite, and R-size.

## Counts

There are 67 finding IDs in all: 30 from the contract lens, 13 from mechanics, and 24 from risk.

| Disposition | IDs |
| --- | --- |
| Folded | 60 |
| Refused | 1 (R-m3), plus three partial refusals inside folded rows (R-M3, R-m2, X-B1) |
| Owner fork | 6 (R-size; C-M10, X-M4, and R-M7 as one fork; R-M6 and C-R6 as one fork), plus the fork half of R-M5, which is counted as folded |

The fork IDs map to Rulings 1, 6, 7, 8, and 9. Rulings 2 to 5 predate the review. Their rows
below are folded, because the lenses changed each ruling's text and the owner still rules the
question.

**Why so little was refused.** Nearly every folded finding cost a state row, a fixture sentence,
or a quote. These all sit on the pass's data-destructive path or its authorization surface. The
repo's planning-miss rule (CLAUDE.md, 2026-09-23) exists because every 5b-ii fix round was an
ambiguity its plan left open, so a row added now is cheaper than a fix round later. The fold
refused where a fix added a second mechanism with no reachable failure: a second authorization
site (R-m3), a guard state no open can reach while a record exists (R-M3's `besideCopy` half), a
denial kind for a caller that cannot reach the unmounted projection (R-m2's second half), and a
lock re-check for an open that never waits (X-B1's inspect half). The fold also declined the one
proposal that needed a fourth test seam. The spec declares exactly three seams (`:278-284`), so
R-M4's device check is proven by a helper table and a real two-device fixture.

## Convergent findings, folded once at the root

| IDs | Finding | Disposition |
| --- | --- | --- |
| X-B1, R-B1 | A writing open that passed `stat()` before a restore wrote its record proceeds past the record once the restore dies. | **Folded.** Verified in `open.go`: `OpenWrite` calls `l.stat()` (`:122`) and never re-checks after `acquireShared` (`:140`), and `OpenServe`'s shared path returns a handle (`:219-231`). R-6bii7 gains "Every writing open reads the record after it wins the lock." Serve's shared path takes the exclusive path on any record. Task 4A gains three lock-wait race rows from a real second process, plus mutations 5 and 6. R-6bii8's combination table adds X-B1's second net, which refuses when `-wal` or `-shm` reappears beside the current database. **Partial refusal:** X-B1 also asked for an inspect re-check. Inspect takes the lock non-blocking, so it never waits through a restore, and its one pre-lock read stays. |
| C-M1, X-M1, R-m5, C-R2 | The `canceled` fixture waits on another process's write lock, which ignores the context. Over `httptest.Server`, the client never reads a 503. | **Folded.** Verified: `write.go:51-53` says the cross-process wait is "bounded by the busy timeout alone." R-6bii5 now keys the wrap on `ctx.Err()`, quoting `context.Context.Err`. It quotes `database/sql`'s `case <-ctx.Done():` connection wait (`sql.go:1370-1375`) for the in-process wait. Task 1 adds the write-pool row and mutation 4. Task 3's `canceled` row uses the write-pool wait, and its HTTP leg uses `httptest.NewRecorder`, quoting `Request.Context`'s "For outgoing client requests, the context controls cancellation." Ruling 2's text is updated to match. |
| R-M1, X-M2 | Resume's combination table has no row for a restore with no current store (the disaster-recovery case). | **Folded.** R-6bii8's table is keyed on what the record owes, with two no-copy rows. Task 4C adds a row per combination and mutation 4. Task 5 adds a fresh-leaf crash row. |
| C-M8, X-M3 | Resume in `library` re-runs the emptiness check and refuses the partial tree the restore itself wrote. | **Folded.** R-6bii7's `library` row and R-6bii11 skip the check on resume, while re-resolving and comparing the root. Deviation D9. Task 6 adds the resume rows and mutation 4, and Task 7 adds a signal-then-resume row. |
| C-M10, X-M4, R-M7 | Review focus 8 promises that an unmounted mount point is refused, and no task builds it. | **Owner fork: Rulings 7 and 8.** Revision 1's claim is struck in "Review focus" whatever the ruling. It is declared in "Not verified." R-6bii11 and Task 6 carry both answers. C-M10's option (c) ignores `lost+found`, and it became Ruling 8 because it is a separate yes-or-no. |
| R-M3, X-m7 | `--discard-current` moves files before the record exists, to unnamed places. | **Folded.** Deviation D8 moves the discard after the record, as step 5's first renames, to `snapshots/discarded-<stamp>.db` and its siblings. Verified: `isLocalCopyName` (`prune.go:204-222`) never matches that name, so the files are never pruned. Task 4C adds the discard rows and mutation 5. **Partial refusal:** R-M3 asked `besideCopy` to treat a discarded name as a copy. With the record written first, no open reaches `besideCopy` while a discarded file sits beside an absent database. Widening the guard would also make serve suggest restoring an unopenable file. |
| C-M7, X-m6, R-m12 | Ruling 5's reference write and step 7's query need a `Handle`, which restore cannot open under its own exclusive lock. | **Folded.** Verified: `WriteBackupRef` is a `*Handle` method that uses only `h.leaf` (`snapshot.go:366-397`). R-6bii6 gives the session a lockless read handle, quoting flock(2) on self-denial. It also gives the session a reference write that shares `WriteBackupRef`'s body through one leaf-level function. The reference write happens after the swap, while the record is still `swap`, so every resume redoes it. Task 4B's third deliverable and Task 5's reference rows hold it. |
| R-M2, X-m8 (second half) | A resume under a different binary acts on a version verdict the old binary made. | **Folded.** R-6bii7's "Resume re-derives every version verdict from the running binary." Deviation D10. Task 4C adds the resume-version row, and Task 5 adds two binary-change rows and mutation 5. |
| C-m14, R-m7 | A `--from-file` source beside a `-journal` file is torn, as is one beside a `-wal`. Detecting the live database by comparing strings is weak. | **Folded.** D4 now covers `-journal`. R-6bii9 decides the live database with `os.SameFile` after `EvalSymlinks`. |
| C-m1, R-m13 (first half) | The crash rows miss step 3. Criterion 7's "leaves the intent record" cannot hold before step 4. | **Folded.** Deviation D7. Task 5 adds the step 3 row, a named hook point between step 5's renames, and a single state per row. The "after step 1 or 2" fixture names a present current store. |
| C-m3, X-m1 | Step 3's refusal for "a reader outside the lock" depends on the holder's state. The `-shm` question is answered. | **Folded.** R-6bii8 quotes the mechanics lens's three measurements and names three refusals. The switch must answer `delete`. Task 4C splits the holder row in two. The leftover-file net uses a planted `-shm` on a DELETE-mode store instead of a timing race. The `-shm` item leaves "Not verified." |
| C-M6, X-m3 | Criterion 13 is never asserted on the `--snapshot` path. `copyto` writes 0644 under umask 022 and replaces a pre-created stage. | **Folded.** R-6bii9 quotes the measurement: the `--snapshot` path pre-creates no stage and does not rely on `-M`. Task 5 adds the criterion 13 row under umask 022 and mutation 6. |
| R-M6, C-R6 | With `--into-nonempty`, plain `rclone copy` overwrites a newer local file and keeps no copy. | **Owner fork: Ruling 6.** The fold probed `copy --backup-dir` (see "Probes") because `rclone copy --help` does not list the flag. The ruling offers `--ignore-existing` and `--update` as other answers. Task 6 carries both answers' rows. |
| R-m2, C-R3 | Ruling 3's reasoning holds for a verified member, not for a caller with no identity. | **Folded** into Ruling 3's text: it is scoped to a verified member, and the answer for a caller with no identity is rung 14's (Task 8 handoff). **Partial refusal:** R-m2 asked for two denial kinds in the authorizer contract now. `opshttp` is unexported and unmounted, so rung 14 can add the second kind at no cost to any caller, and nothing at 6b-ii can reach it. |

## The remaining findings

### Contract lens

| ID | Disposition |
| --- | --- |
| C-M2 | **Folded.** Verified: `backup.go:69` resolves rclone through `exec.LookPath`, and `standin_test.go` already drives `ops/local` this way. Task 3's "How the backup body blocks" rules the matter, criteria 3 and 4 assert the stand-in's recorded SIGINT, and the item leaves "Not verified." |
| C-M3 | **Folded.** R-6bii5 adds `ops.Codes`, whose production caller is the projection constructor's status-table check. Task 3's guard iterates it. Mutation 2 adds an eighth sentinel. |
| C-M4 | **Folded.** R-6bii1 names the literal values (verified at `local.go:121-129`). Task 1 asserts them as literals, and mutation 3 flips `start_backup`. Task 2 adds a member-only row and Task 3 a forbidden-by-privilege row. `Catalog` errors on a missing name. |
| C-M5 | **Folded.** Task 4C's criterion 6 proves the rows live only in the WAL with a bare-file copy. Mutation 1 skips step 3 and says why skipping the checkpoint alone is not a killing mutation. R-6bii8 states the checkpoint's reason. |
| C-M9 | **Folded.** R-6bii7 defines the completion tuple. A global constraint and every resume row assert it. |
| C-m2 | **Folded.** Task 4B adds an inspect-mode holder row, and criterion 5 runs both holders. |
| C-m4 | **Folded.** Task 3's actor row uses `start_backup` with `actor` and `ACTOR`. |
| C-m5 | **Folded.** R-6bii2 and Task 2 name `ops.LocalActor` as the one buildable mismatch. |
| C-m6 | **Folded.** Task 4A builds its own `swap`-record and staged-file fixture (the serve-then-write row). |
| C-m7 | **Folded.** Task 4C's mutation 3 is killed by criterion 6's pre-restore assertion. |
| C-m8 | **Folded.** R-6bii12 drops the `WithoutCancel` rule, since no record write follows a cancel, and states why. Task 7's mutation 1 is replaced. |
| C-m9 | **Folded.** Task 7's end-to-end fixture empties the library root and asserts rows by value. |
| C-m10 | **Folded.** R-6bii9's zero-byte row names `application_id` 0, quoting `verifyCopy`'s comment. The version rule is a new parameter, and backup keeps exact equality under a test (Task 4B). |
| C-m11 | **Folded.** Verified: the real set is `0001_ledger.sql`, `0002_jobs.sql`, and `0003_catalog.sql`. Task 4A's criterion 11 fixture is the real set's first two files. |
| C-m12 | **Folded.** Exit gate 14 lists every export with a pasted `git grep`. |
| C-m13 | **Folded.** Exit gate 9 names every state table. |
| C-m15 | **Folded** with C-M1: R-6bii5 quotes `errDBClosed` (`sql.go:1318`) and `classifyErr`'s pass-through (`conn.go:216-218`), and the item leaves "Not verified." |
| C-m16 | **Folded.** Task 2 adds the 4096/4097 boundary pair. Task 4B declares criterion 13's uid half a reviewer-read. R-6bii7's refusals of an unreadable record name the manual procedure (Ruling 9). |
| C-R5 | **Folded** into Ruling 5's text: the `--older` consequence. |

### Mechanics lens

| ID | Disposition |
| --- | --- |
| X-m2 | **Folded.** R-6bii9 quotes the measurement: SIGINT removes the `.partial` file and SIGKILL leaves it. Task 5 splits the row, and Task 4B clears a planted `.partial`. |
| X-m4 | **Folded.** R-6bii9 decides "no such object" from the listing before any `copyto`. R-6bii10 refuses a remote without `library/` and prints `sizeless`. Task 6 adds the remote row. |
| X-m5 | **Folded.** R-6bii4 checks the first token. Task 2 adds the `null` row and mutation 8. |
| X-m8 (first half) | **Folded.** R-6bii6 removes a verified stage's `-wal` and `-shm`, refusing a non-empty `-wal`. Task 4B adds the row. |

### Risk lens

| ID | Disposition |
| --- | --- |
| R-M4 | **Folded.** R-6bii6 adds the one-device check (deviation D10). Task 4B adds the device row, with `db/` symlinked into `/dev/shm`, and mutation 4. The fold measured `/tmp` and `/dev/shm` here as separate devices (60 and 27). A helper table test covers the comparison, so no fourth test seam is needed. |
| R-M5 | **Folded** (the rule): a new restore beside a record refuses, naming `--resume` (R-6bii6). Task 4B adds the row and mutation 3, and Task 7 adds a row. **Owner fork** (the way out of an unresumable record): Ruling 9. |
| R-M8 | **Folded.** R-6bii2 pins the authorizer's signature. R-6bii3 has the layer discard its caller's actor, and the HTTP layer pass the zero `Actor`. Task 2 adds the planted-actor row and mutation 6, and the privileged-field row is updated. |
| R-m1 | **Folded.** Task 8's rung 14 handoff carries the same-origin question and the bearer mount's bypass. |
| R-m3 | **Refused.** Authorizing before the decode would add a second authorization site that the in-process contract leg never runs. The gain is hiding which operation names exist, which the member catalog already lists. The deliberate order is recorded in R-6bii3. |
| R-m4 | **Folded.** R-6bii4 reads `(*ops.Error).Reason` through `errors.As`. Task 2 adds the wrapped-path row and mutation 7. |
| R-m6 | **Folded.** R-6bii9 adds the ledger row through `checkLedger` (verified at `migrate.go:190`). Task 4B adds a stage row. |
| R-m8 | **Folded.** R-6bii7 stores basenames and fixes each field's directory. |
| R-m9 | **Folded.** Task 8 adds the runbook line. Verified: `pruneLocal` removes a local copy modified before the run (`prune.go:168-200`), and nothing uploads it. |
| R-m10 | **Folded.** Ruling 5's text names the `--from-file` consequence. R-6bii9 and Task 7 add the remote-mismatch warning. |
| R-m11 | **Folded.** The report is written through `atomicfile.WriteStream` at 0600 (R-6bii7, the privileged-field row, and Task 6's criterion). |
| R-m13 (second half) | **Folded.** Ruling 4's text now says that `backup`'s 75 is a busy kind lease (verified at `backup.go:101`). |
| R-cite | **Folded.** R-6bii7 no longer cites spec `:698` for the record's 0600. The mode comes from `atomicfile`. |
| R-size | **Owner fork: Ruling 1.** The recommendation changes to split. "Token ceiling" and "Execution model" carry both answers. |

## Rulings awaiting owner, as consolidated

1. Does 6b-ii run as one pass under 9.63M? The recommendation is no: split into 6b-ii-a (3.90M) and
   6b-ii-b (6.85M).
2. Does `canceled` mean the caller's own context ended before the seam answered? Yes.
3. Is a verified member lacking the privilege answered 403 `forbidden`, with the no-identity
   answer left to rung 14? Yes.
4. Are restore's exit codes 0, 1, 75, and 130? Yes.
5. Does the backup reference follow a `--snapshot` restore and stay put after `--from-file`? Yes.
6. Does `--into-nonempty` keep each displaced local file through `rclone copy --backup-dir`? Yes.
7. Is an unmounted mount point a declared gap, with a printed mount-point line and a runbook line?
   Yes.
8. Does the emptiness check ignore `lost+found`? Yes.
9. Is an unresumable record handled by a runbook procedure rather than `restore --abandon`? Yes.

## Errata owed to the rung 6 spec

Revision 2 changes the ratified spec's meaning in these places. None is edited here, and Task 8
drafts the text.

| # | Spec site | Change | When |
| --- | --- | --- | --- |
| D1 | `:1037-1038` | The in-process contract leg runs through the authorizing layer. | on approval (carried from revision 1) |
| D2 to D6 | as revision 1 listed | Serve's reconciler, restore's package split, the `--from-file` sibling refusal, the intent states, and `opshttp`'s import of `identity`. D4 now also names `-journal`. | on approval |
| D7 | `:1047-1048` | Criterion 7's "leaves the intent record" holds from step 4 on. | on approval |
| D8 | `:693-697` | `--discard-current` moves the files after the record, to `discarded-<stamp>` names. | on approval |
| D9 | `:702-703` | Resume in `library` skips the emptiness check. | on approval |
| D10 | step 1 and "Three open modes" | The one-device refusal, the refusal of a new restore beside a record, and resume's version re-check. | on approval (additive) |
| R2 | `:603` | The `canceled` contract case becomes a canceled caller waiting on the write pool. | if Ruling 2 is yes; if no, the vocabulary drops to six |
| R6 | `:702-703` | Step 6 keeps displaced local files under `<root>.restore-displaced-<stamp>`. | if Ruling 6 is yes |
| R8 | `:702-703` | A root holding only `lost+found` counts as empty. | if Ruling 8 is yes |
| R9 | the flag table, `:676-686` | A `--abandon` flag. | if Ruling 9 is no |

The spec's "every writing open checks for a `restore-intent` record first" (`:240`) is not
changed. The under-lock re-read is what makes that sentence true for an open that waits.

## Ceiling and task list after the fold

The one-pass ceiling rises from 8.4M to 9.63M (flag 7.70M). The split ceilings are 3.90M
(6b-ii-a, flag 3.12M) and 6.85M (6b-ii-b, flag 5.48M), about 1.1M more in total. The growth is
Task 4C (0.51M), about 0.27M of added rows across Tasks 1, 2, 3, 4A, 5, and 6, and a restore fix
reserve raised from 0.40M to 0.60M.

| Task | Model | Deliverables |
| --- | --- | --- |
| 0 | `opus` | 1 (the report) |
| 1 | `sonnet` | 4 (the sentinels with the classifier and `Codes`, the network actor, the privilege table, the seam boundary's `canceled` classification) |
| 2 | `opus` | 4 (the authorizing layer, the HTTP layer, the constructor, the error-body discipline) |
| 3 | `opus` | 3 (the sentinel cases with the guard, the job path, the cross-process cancel) |
| 4A | `opus` | 4 (the record, the refusals before and after the lock, serve's `reconcile-pending` path, the read query) |
| 4B | `opus` | 3 (the session open, staging and verification, the session's handles) |
| 4C (new) | `opus` | 3 (the quiesce, the swap with the discard, resume's combination rule) |
| 5 | `opus` | 4 (the flow with the hook, the fetch, resume, the reference rewrite) |
| 6 | `opus` | 3 (step 6 with its resume, the reconciliation, the renders) |
| 7 | `opus` | 4 (the flags, the exit codes, the signals, the output) |
| 8 | `sonnet` | 4 (the HISTORY draft, the errata, the handoff with the runbook lines, the STATUS draft) |

That is 36 deliverables, and no task has more than four. On Ruling 9's no, a Task 7B (`opus`, four
deliverables, about 0.5M) joins the restore half.

## Probes

The fold ran these probes itself. The mechanics lens's probes are cited from its review.

- One device, rclone v1.74.3, local backend. The source held `a.flac`
  (`backup`) and `b.flac` (`same`). The destination held `a.flac` (`local-newer`), a timestamp-kept
  copy of `b.flac`, and `c.flac` (`extra`). After `rclone copy src dst --backup-dir displaced`,
  `dst/a.flac` read `backup`, `dst/b.flac` and `dst/c.flac` were unchanged, and
  `displaced/a.flac` read `local-newer`. With the backup directory inside the destination, rclone
  refused: "destination and parameter to --backup-dir mustn't overlap."
- Two devices. The destination was under `/dev/shm` (device 27) and the backup
  directory under `/tmp` (device 60). The differing file was displaced into the backup directory,
  and the destination read the backup's version.
- The flag listing. `rclone copy --help` does not list `--backup-dir`. `rclone help flags` lists it
  in the "Sync" flag group as "Make backups into hierarchy based in DIR", beside `-u, --update`
  ("Skip files that are newer on the destination") and `--ignore-existing` ("Skip all files that
  exist on destination").
- Quotes read for the fold: `context.Context.Err`'s doc (go1.27.1); `database/sql`'s
  `errDBClosed` and its `case <-ctx.Done():` connection wait (`sql.go:1318`, `:1320-1331`, and
  `:1370-1375`); `net/http`'s `Request.Context` doc; and `mklost+found(8)`: "There is normally a
  lost+found directory in the root directory of each file system."

## Fold 2

Input: the verification read of revision 2
(`2026-09-25-rung-6b-ii-plan-fold-verification.md`, committed at `f21c417`): 0 blockers, 2
majors, and 10 minors. Each finding was checked against the plan and against the code on master
before it was folded. The fold is narrow: no ruling's substance changed, and all nine stay yes/no
with their recommendations. The result is revision 3, status "reviewed; awaiting owner approval
(Rulings 1-9)."

### Checks made before folding

- M1: Spec `:278-285` declares exactly three test seams, the step hook among them as "an
  unexported function variable called between flow steps." Fold 1 already refused a fourth seam
  for R-M4. A hook variable in `internal/backup` cannot fire inside one `store` method, so the
  finding holds.
- M2: Ruling 6 and R-6bii11 derived the displaced directory's name from the record's stamp and
  said only "created at 0700 before the copy." A resumed step 6 names the same path, so an
  `os.Mkdir` on resume fails with `EEXIST`. The finding holds.
- m1: `waitLock` logs the holders on its first busy attempt (`open.go:312-313` on master), and
  `acquireBounded` (`lock.go:63-78`) logs nothing. Confirmed.
- m2: `prune.list` (`prune.go:248`) and `runRclone` (`prune.go:282-294`) call
  `exec.CommandContext` directly with no `SysProcAttr`, `Cancel`, or `WaitDelay`, while
  `transport.run` (`rclone.go:51-72`) sets all three. `git grep` finds three launch sites in
  `internal/backup`. Confirmed.
- m3: `TestImportsStandardLibraryOnly` is at `internal/ops/ops_test.go:185`, and
  `errors.go:21-23` links `[ErrForbidden]` and `[ErrInternal]`, which do not exist. Confirmed.

### Dispositions

| Finding | Disposition |
| --- | --- |
| M1 | **Folded as proposed.** R-6bii8 gains "Step 5 is two session calls": the first holds the pre-restore rename, the discard's moves, or nothing on a fresh leaf; the second holds the incoming rename and the fsyncs. The flow in `internal/backup` calls the hook between them, so the seams stay three. D3, R-6bii6's method list, Task 4C's swap outcome, and Task 5B's flow outcome and crash rows say the same. Task 5B's in-between crash row now names the third combination row; revision 2 named the second, which is the `-wal` refusal. A discard crash row between the two calls was added. |
| M2 | **Folded.** R-6bii11 gains "The displaced directory" with three rules. `--backup-dir` applies with `--into-nonempty` only, since without it every file in the root came from the restore's own copy. A first run creates the directory with `os.Mkdir` at 0700 and refuses on any failure, `EEXIST` included. A resume accepts an existing directory only when `os.Lstat` shows a directory, the caller's uid, and mode 0700. The directory is removed by `os.Remove` only when it holds no entry, and never recursively. The pre-creation cites the verification read's probe (rclone v1.74.3 creates a missing backup directory at 0755 under umask 022). Ruling 6's "what yes builds" sentence and the privileged-field row were aligned without changing the question. Task 6 gains four rows (the hand-built interrupted `--into-nonempty` resume, a wrong-state displaced path, the empty-directory removal, and a first run beside an existing path) and mutations 6 and 7. Review focus 9 names the kept files. |
| m1 | **Folded.** Task 4A's fixture paragraph: the serve rows wait for the holders log line through a recording logger. The write-mode row waits a settle delay, and the task reports mutation 5's kill rate across repeated runs; any survival makes it a reviewer-read. |
| m2 | **Folded, into Task 5A.** R-6bii6 now states the architecture read's fix: `transport.run` takes a stdout writer, `prune` holds a transport, and `runRclone` is deleted. R-6bii9's listing and R-6bii10's `rclone size --json` name the transport. Task 5A carries the launch path as a deliverable, a canceled-listing row, mutation 3, and a `git grep` criterion finding one launch site. Task 7 gains the SIGTERM-during-listing row (exit 130). |
| m3 | **Folded.** Task 1 names both edits for the diff-reviewer: `TestImportsStandardLibraryOnly` admits `internal/identity` alone, and the `errors.go:21-23` doc links are rewritten. The sentinel-list collapse is routed to the architecture cleanup pass. The global constraint and exit gate 14 now say that the `Codes` and classifier caller is unmounted projection code, which no binary reaches until rung 14. |
| m4 | **Folded.** The completion tuple gains the discard case: `snapshots/discarded-<stamp>.db` exists, each planted `-wal` or `-shm` sits beside it under the discarded name, and none remains in `db/`. The pre-restore clause now excludes `--discard-current`. |
| m5 | **Folded.** R-6bii7: resume reads the staged file's version through the session's verify-then-remove method. Task 5B gains a criterion (no `-wal` or `-shm` in `staging/` after such a resume) and mutation 6. |
| m6 | **Folded.** The reference write, and the move of `WriteBackupRef`'s body into a shared function, are built on Ruling 5's yes only, in R-6bii6, the privileged-field row, Task 4B's files, its handles outcome, and its row. Task 5B states what its rows become on the no. |
| m7 | **Folded.** Ruling 2's no: Task 1 deletes `ErrCanceled`, `CodeCanceled`, and `NewCanceledError`; the tables count six; the classifier's `context.Canceled` row answers `internal`; the seam boundary outcome, rows, and mutation go; Task 1 carries three deliverables. R-6bii4 drops the `canceled` status, and Task 3 drops its row. Ruling 3's no: R-6bii4, Task 2's authorizer rows, and Task 3's HTTP leg answer what `http.NotFound` answers, and Task 3's mutation 3 is reversed. Ruling 7's no: R-6bii11 names the flag `--root-not-a-mount` (only with `--library`, checked once, skipped on resume), Task 6's same-device row refuses without it, and Task 7's flags carry it into criterion 12's empty-root fixture. |
| m8 | **Folded.** Exit gate item 10 is 6b-ii-a's alone. Item 14 applies to both halves, each over its own exports, as item 9 does over each half's tasks. Task 8's draft files are suffixed `-6b-ii-a` and `-6b-ii-b` on a split. |
| m9 | **Folded in part.** Task 7's signal-then-resume row names the `PATH` switch: the blocking stand-in first for the interrupted run, the real rclone first for the resume. Task 8's runbook lines name the `.partial` files a SIGKILLed library copy leaves, which the report lists as files without a row. **Refused:** clearing `*.partial` under the root on a `library` resume. It deletes files under the library root, which the privileged-field table rules out, to save one manual step after a SIGKILL. |
| m10 | **Folded by a task split.** Task 5, with m2 absorbed, carried six outcomes. It splits into 5A (the one launch path, the listing with its rules, the fetch: three deliverables) and 5B (the flow with the hook, resume, the reference rewrite: three). The dependency chain, segments C and B2, the exit gate, review focus, the models paragraph, and the hooks table are updated. |

### Ceilings after fold 2

Task 5A is priced at 0.51M and Task 5B at 0.56M, against revision 2's single 0.56M, which adds
one task's rate of 0.51M. Task 6's displaced-directory rows add 0.03M.

- **One pass (Ruling 1 yes):** projection 7.76M; flag 7.76M + 0.48M = **8.24M**; ceiling 8.24M /
  0.8 = **10.30M** (was 9.63M).
- **Split (Ruling 1 no):** 6b-ii-a is unchanged at **3.90M** (flag 3.12M). 6b-ii-b's projection
  is 5.54M; flag 5.54M + 0.48M = **6.02M**; ceiling 6.02M / 0.8 = 7.525M, so **7.53M** (was
  6.85M). Together 11.43M, about 1.1M above the one-pass ceiling, as before.

That is 38 deliverables. Tasks 1, 2, 4A, 7, and 8 carry four each, and Tasks 3, 4B, 4C, 5A, 5B,
and 6 carry three each. Ruling 1 now counts five task splits against the spec's task count, and
its recommendation to split the pass is unchanged.

## Prose fold

The prose review (`2026-09-25-rung-6b-ii-plan-prose-review.md`, over revision 3 at `3df0fd0`)
raised 1 blocker, 7 warnings, and 4 suggestions. Each was checked against the repo before it was
applied. The plan is now revision 4, and `tellgrader --register docs` reports 0 findings on it. No
task scope, ceiling, or ruling's substance changed.

| Finding | Disposition |
| --- | --- |
| 1 (blocker) | **Folded.** Verified: `backup.go:72` declares exit 69, and its help lists 69 at `:100` and 75 at `:101`. Ruling 4's reason now says that `backup` also exits 69 (disabled) and 75 (a backup already running), that restore has no 69, and that the disabled-integration row differs between the verbs. The optional "rule that row separately" offer is added. |
| 2 | **Folded as proposed.** The first discard row now reads "none, some, or all of the three files still in `db/`", with "the rest moved", so a crash with every file already moved matches it rather than the refusing row. Task 5B's discard crash row names that state: the rename, the fsyncs, then the discard's completion tuple. |
| 3 | **Folded, with one correction.** STATUS's lane registry shows both lanes closed, and `git worktree list` shows only the main tree. The sources, the pointer paragraph, primary input 4, the ceiling basis, the lane constraint, the agentic-workers note, Task 0's launch line and item 5, the conductor's action 1, the contested-surface rows, the dependency note, and "Not verified" were updated. **Corrected:** master does not contain `33ceff5`, because 6b-i was rebased before its merge. The pointer paragraph says instead that the cited paths are byte-identical between `33ceff5` and `6f2c3e0` and unchanged on master since (checked with `git diff`). The `internal/auth`, `internal/identity`, and `internal/serve` pointers were checked the same way: only `internal/serve`'s `csp.go` and `document.go` moved since `f0eec0d`, and the plan cites neither. Both "Not verified" bullets are settled and removed. |
| 4 | **Folded.** Exit gate item 9 reads "R-6bii7's two tables", matching `:676`, `:704`, and Task 4A. |
| 5 | **Folded.** The spec's rulings read "spec ruling 7", "spec ruling 11" (twice), and "spec ruling 12" (twice). |
| 6 | **Folded as proposed.** Ruling 9's question is rewritten, and `--abandon` is called a flag, matching Task 7B. |
| 7 | **Folded as proposed.** Verified against the installed rclone v1.74.3: `rclone copy --help` carries both quotes and does not list `--backup-dir`; `rclone help flags` lists it under "flag group Sync". The "whichever side is newer" gloss is replaced, the long sentence is split, and the close names the two alternatives as answers instead of yes or no, each changing Task 6's differing-file row (revision 4 `:1700`). |
| 8 | **Folded as proposed.** Ruling 7 is one question. Its "no" names `--root-not-a-mount`, the flag R-6bii11 and Task 7 already carry, and the "strikes revision 1's claim" clause is dropped (R-6bii11 still records it). |
| 9 | **Folded.** Ruling 1 states the ten build tasks, five past the spec's count. |
| 10 | **Folded.** Verified: `write.go:51-52` is the pool wait, `:52-54` the busy timeout, and `isLocalCopyName` is `prune.go:202-220`. Ruling 2, R-6bii5, Task 0 item 2 (`:51-54`, the whole wait contract), the privileged-field row, and R-6bii8 now cite those ranges. |
| 11 | **Folded.** R-6bii4's error paragraph is split into the status per code, the six request refusals as a list, and the decoding rules as a list. The completion tuple is a bullet list. No rule was added or dropped. |
| 12 | **Folded.** The spliced lines are rewrapped, along with the paragraphs this fold edited. |

Refused: none.
