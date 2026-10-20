# Rung 6b-ii plan review: mechanics and feasibility

Reviewed: `docs/plans/2026-09-25-rung-6b-ii-restore-http.md`, revision 1 (uncommitted).
Lens: mechanics and feasibility only. Does each mechanism behave as the plan states?
6b-i code was read in the lane 2 worktree at `33ceff5`. Toolchain: go1.27.1 linux/amd64, modernc.org/sqlite
v1.59.0 (embeds SQLite 3.53.4), rclone v1.74.3. Probes live in the reviewer's scratch directory
(`/tmp/claude-1000/.../scratchpad/{sq,js,rc,mod}`) on tmpfs. No repo file other than this one was
written.

Counts: **1 blocker, 4 major, 8 minor.** One major is an owner fork.

---

## Blocker

### B1. A writing open honors the intent record only if the record existed before the open took the lock

- **Location:** plan `:191-194` (Goal), `:437-477` (R-6bii7), Task 4A `:870-928`. Code:
  `internal/engine/store/open.go:122` and `:140` (`OpenWrite`: `l.stat()` runs `checkIntent`, then
  `acquireShared`, and nothing checks the intent again); `:202` and `:219-234` (`OpenServe`:
  `stat` runs before `waitLock`, and the shared fast path returns a handle without a second check);
  `:361`.
- **Defect:** the record is written under the exclusive lock. A write-mode or serve-mode open that
  passed `stat()` before the restore wrote the record is already polling for the shared lock. When
  the restore dies, the lock is freed and the waiting open proceeds past a `swap` record.
  Depending on where the restore died, that open writes into the old database (which a resume then
  renames into `pre-restore-*`, losing the writes), into the half-swapped leaf, or into the
  restored database before it is reconciled. On the box this is a realistic path: a restarted
  `serve` waits up to `MigrationWait`, and a timer backup waits on `acquireShared`. The plan's Task
  4A rows test every record state an open finds *at `stat` time*. None tests a record that appears
  while the open waits.
- **Evidence:** a probe test was added to a scratch copy of the worktree's `internal/engine/store`.
  It opens a store through `OpenServe`, then takes `LOCK_EX` on `store.lock` itself. It starts
  `OpenWrite` and `OpenServe` in goroutines and waits 300 ms (both are then past `stat` and
  polling). It then writes `restore-intent` containing `{"state":"swap"}` and closes the lock.
  ```
  serve open beside a swap record: err=<nil>
  write open beside a swap record: err=<nil>
  ```
- **Proposed fold:** in Task 4A, every mode checks the intent again after it wins the store lock.
  That covers `OpenWrite` after `acquireShared`, `OpenServe`'s shared path after `waitLock`
  (`settle` already checks under the exclusive lock), and inspect. Serve's re-check accepts only
  `reconcile-pending`. Add one state row, "a record written while the open waits on the lock,"
  using the fixture above, with the refusal text asserted. As a cheap second net, resume in the
  `swap` state refuses when `-wal` or `-shm` has reappeared beside the current database before the
  first rename.

---

## Major

### M1. The `canceled` contract case cannot pass as specified, and its HTTP leg cannot observe the answer

- **Location:** Ruling 2 `:76-87`, R-6bii5 `:405-409`, Task 3 table `:835`.
- **Defect 1:** the fixture is "a writer holding the store's write lock in a second process, the
  caller's context canceled while the call waits." A cross-process wait is SQLite's busy handler,
  which ignores the context. The call returns `SQLITE_BUSY` after the full busy timeout, and no
  `context.Canceled` appears in the chain. Neither R-6bii5's chain test nor `ops/local`'s chain
  wrapping can classify the error as `canceled`, so the case answers `internal` (or whatever busy
  maps to). The store already documents this in `write.go:51-53`: "The wait on another process's
  write lock is bounded by the busy timeout alone: a deadline shorter than the busy timeout does not
  shorten it."
- **Defect 2:** over `httptest.NewServer`, a request whose client context is canceled returns a
  client-side error and no status. The HTTP leg therefore cannot assert "canceled" on the wire.
  `go doc net/http Request.Context`: "For outgoing client requests, the context controls
  cancellation."
- **Evidence:** two processes, `writeDSN`, busy timeout 5000 ms. The holder runs `BEGIN IMMEDIATE`.
  The waiter calls `BeginTx`, and its context is canceled at 200 ms:
  ```
  elapsed=5.013s err=database is locked (5) (SQLITE_BUSY) isCanceled=false
  ```
  An in-process wait on the write pool's single connection (`SetMaxOpenConns(1)`, a transaction
  held open) with the context canceled at 200 ms:
  ```
  elapsed=200ms err=context canceled isCanceled=true
  ```
  A pre-canceled context on `QueryRowContext` gives `context canceled`, `isCanceled=true`, so Task
  1's `JobStatus` criterion holds.
- **Proposed fold:** change the fixture to an in-process wait. An `Update` callback on the runner's
  own handle blocks, then `CancelJob` or `StartBackup` runs with a context canceled mid-wait.
  Drive the HTTP leg with `handler.ServeHTTP(httptest.NewRecorder(), req.WithContext(canceledCtx))`
  instead of over the wire. If Ruling 2 should also cover the cross-process wait, the seam boundary
  must test the caller's `ctx.Err()` on any error return, not only the chain. Say which in R-6bii5.
  The wait still lasts the full busy timeout either way.

### M2. The swap-resume table has no row for a restore with no current store, the disaster-recovery case

- **Location:** R-6bii8 `:505-510`, and Task 4B's rows `:966` and `:973`.
- **Defect:** the four combinations assume that a pre-restore copy is owed. Take a fresh leaf (a
  lost box, or a lost database), where the record names no pre-restore copy (`:966`). A crash
  after step 4 finds current absent, copy not owed, and incoming present. After the rename it finds
  current present, copy not owed, and incoming absent. Neither state is in the table, so both fall
  to "any other combination: refused". `--resume` cannot finish the most important restore.
- **Evidence:** read from the table against R-6bii7's field "the pre-restore copy's name (empty
  when no current store existed)" (`:446-447`).
- **Proposed fold:** add two rows keyed on an empty pre-restore name. Absent, not owed, present
  means do the second rename, then the fsyncs. Present, not owed, absent means do the fsyncs only.
  Test both with Task 5's step-hook crash on a fresh-leaf fixture.

### M3. `--resume` in the `library` state dead-ends on the non-empty-root refusal

- **Location:** R-6bii7 `:453` ("re-runs step 6's copy"), R-6bii11 `:571-573`, Task 6 `:1097`,
  Task 7 `:1147`.
- **Defect:** step 6 refuses a non-empty root without `--into-nonempty`. Once a copy has failed, or
  SIGINT or a crash has interrupted it, the root holds a partial tree. A resume that re-runs step 6
  as written refuses the root the restore itself filled. The restore can then never finish, and
  every writing open keeps refusing. Neither Task 5's crash table nor Task 6 or Task 7 pairs
  "interrupted during the copy" with a resume.
- **Evidence:** read from the plan. The record carries `--into-nonempty` (`:447`), but no row says
  that resume uses the recorded decision instead of checking the root again.
- **Proposed fold:** the emptiness check runs once, before the record enters `library`. Resume keeps
  the absent-root check (re-resolving the root and comparing it with the recorded root) and skips the
  emptiness check. Add a Task 6 row, "rclone killed mid-copy, then `--resume`," which completes the
  copy with every file present.

### M4. OWNER FORK: the "unmounted mount point" guard does not exist

- **Location:** Review focus 8 `:627-628` ("refused, never populated. Pinned in Task 6"), R-6bii11
  `:569-572`.
- **Defect:** an unmounted mount point is not absent. It is the existing, usually empty, directory
  a filesystem would attach to. mount(8): the kernel attaches the filesystem "at the directory dir".
  The absent-root check and the non-empty check both pass it, and `rclone copy` fills the
  underlying filesystem, often the box's OS disk. The free-space check stops this only when the
  library is larger than that disk's free space. Task 6's table has no row that could pin the claim.
- **Options:**
  - **(a)** Drop the claim from Review focus 8. Print the resolved root's device and filesystem
    type before the copy, and let the runbook (rung 16) say "mount the library volume first".
    Cost: none.
  - **(b)** Refuse an empty root that shares `st_dev` with its parent directory unless the operator
    passes a confirming flag. Cost: a new flag. It also false-positives on every same-filesystem
    library, which is the default layout.
  - **(c)** Compare the root against `/proc/self/mountinfo` and `/etc/fstab`, and refuse when an
    fstab entry covers the root but is not mounted. Cost: parser code and fixtures.
    It misses systemd `.mount` units and cloud-init mounts outside fstab.
- **Recommendation:** (a). The universal-setup goal rules out a heuristic that fires on the default
  layout, and (c) is incomplete.

---

## Minor

### m1. Step 3's busy column misses an idle foreign connection; the `-shm` question is answered

- **Location:** R-6bii8 `:481-496`, Task 4B rows `:968-969`, "Not verified" `:1368`.
- **Evidence:** a WAL left by a child killed with SIGKILL (`-wal` 20632 bytes, `-shm` 32768 bytes),
  opened on a `maintenanceDSN`-equivalent connection:
  - **Sole connection:** `wal_checkpoint(TRUNCATE)` returns busy 0, and `-wal` drops to 0 bytes.
    `journal_mode=DELETE` returns `"delete"`, and both `-wal` and `-shm` are **gone before the
    close**. The header reads 1/1 after the close. This settles the plan's open item: SQLite 3.53.4
    removes `-shm` at the switch.
  - **Another process has the database open, idle, with no transaction:** the checkpoint returns
    **busy 0**. `journal_mode=DELETE` fails at once (0 ms, despite busy_timeout) with
    `database is locked (5) (SQLITE_BUSY)`. `-wal` (0 bytes) and `-shm` remain, and the header
    stays WAL.
  - **Another process holds a read transaction:** the checkpoint returns busy 1, and `-wal` stays
    at 20632 bytes. `journal_mode=DELETE` fails with `SQLITE_BUSY`.
- **Proposed fold:** the refusal for "a reader outside the lock" comes from the `journal_mode`
  statement's `SQLITE_BUSY` (through `classifyErr`). The busy column covers only an open read
  transaction. Also require the returned mode to equal `delete`. Split row `:968` into the two
  holder states. Row `:969` ("`-shm` remains") needs a named fixture: a sole connection never leaves
  one. A feasible fixture is a second process opening on the WAL DSN between the switch and the
  close.

### m2. On SIGINT, rclone removes its own `.partial`; only SIGKILL leaves one

- **Location:** Task 5 crash table `:1040`.
- **Evidence:** `rclone copyto --bwlimit 2M` of a 20 MB object into `staging/restore-x.db`.
  Mid-transfer, the file is `restore-x.db.a80e31de.partial` at mode 0644. After SIGINT, rclone
  exits 130 and `staging/` is **empty**. After SIGKILL, the `.partial` remains. The transport
  cancels with SIGINT (`rclone.go:69`).
- **Proposed fold:** the row expects no leftover after a SIGINT cancel. Prove the "a new restore
  clears the leftover" behavior with a SIGKILL'd rclone or a planted `<name>.<hex>.partial`.

### m3. What `copyto` gives the staged file; do not rely on `-M`

- **Location:** R-6bii9 `:527-529`, "Not verified" `:1370`.
- **Evidence:** the remote object was uploaded as backup uploads it (`-M`, source 0600). Fetched
  without `-M` under umask 022, the file is **0644**. Fetched without `-M` under umask 077, it is
  0600. Fetched with `-M` on the local backend, it is 0600, because the mode comes from the
  object's metadata (`"mode":"100600"`). A pre-created 0600 stage is **replaced**, with a new inode
  at 0644. The leaf directories are 0700 (`leaf.go:19`), so nothing is exposed in the window.
- **Proposed fold:** keep the chmod. Do not pre-create the stage for the rclone path, since rclone
  replaces it. Do not add `-M` to the fetch: mode metadata is a local-backend feature.

### m4. rclone does not name a missing object; an absent remote `library/` is an unnamed state

- **Location:** R-6bii9 row `:538`, R-6bii10 `:560-563`, R-6bii11.
- **Evidence:** `copyto` of a missing object retries three times and exits 3. Its last stderr line
  is `NOTICE: Failed to copyto: directory not found`, with no path. `lsjson` of the missing file
  path also exits 3. `rclone size --json` of an absent `library/` exits 3. On a present tree it
  prints `{"count":2,"bytes":20004096,"sizeless":0}`.
- **Proposed fold:** decide "no such object" from the `snapshots/` listing, which `latest` and the
  free-space size already need, before any `copyto`, and build the message with the remote path.
  Name the state "`--library` with no `library/` on the remote" and give its report. `sizeless` in
  the JSON counts exactly the objects that made the library check a floor, so print that count.

### m5. A `null` body passes as a JSON object

- **Location:** R-6bii4 `:385-392`, Task 2 row `:768`.
- **Evidence:** go1.27.1 `json.Decoder.Decode` into a struct: `null` gives the zero request and
  `err=<nil>`. `[]`, `"x"`, and `1` all error. `{"job_id":"a","job_id":"b"}` gives `b` with no
  error. `{"JOB_ID":"b"}` matches case-insensitively, as quoted. The trailing-data check
  (`dec.Token()` expecting `io.EOF`) works.
- **Proposed fold:** check that the first token is `{`, or add `null` to Task 2's non-object row, so
  that `null` to `start_backup` is refused instead of starting a backup.

### m6. The restore session's `*store.Handle` source is unstated, and flock self-denies within a process

- **Location:** R-6bii6 `:417-423`, R-6bii11 `:575-576`, Task 5's reference rewrite `:1028`
  (`WriteBackupRef` is a `Handle` method, `snapshot.go:365`), and Task 6 `:1084`.
- **Defect:** step 7 needs a `Handle` for the read query, and Ruling 5 needs one for the reference.
  Either one obtained through `OpenWrite`, or anything that calls `tryLock`, is denied by the
  session's own exclusive lock. flock(2): "An attempt to lock the file using one of these file
  descriptors may be denied by a lock that the calling process has already placed via another file
  descriptor." B1's probe showed this: the same process's `OpenWrite` waited on the exclusive lock
  the test held.
- **Proposed fold:** Task 4B's session hands out a lockless handle (`lock` nil, `version` set, and
  `leaf` set), or exposes the reference write and the read itself. Name that export in exit gate
  13's accounting.

### m7. `--discard-current` moves files before the record exists

- **Location:** Task 4B `:953` and `:971-972`, spec step 3 against step 4.
- **Defect:** a crash after the move-aside and before step 4 leaves the database absent with no
  record. Unless the moved-aside names match `isCopy` (`open.go:503-506`), `serve`'s `besideCopy`
  sees nothing (with no backup reference or copy present), creates an empty store, and clears the
  verified stage.
- **Proposed fold:** name the moved-aside database with `PreRestorePrefix`, so `besideCopy` refuses
  beside it, or do the move as step 5's first rename after the record, carrying the `-wal` and
  `-shm` siblings.

### m8. Verification on a WAL-header stage leaves `-wal` and `-shm` in `staging/`, and resume trusts a stale state split

- **Evidence:** a byte copy of a cleanly closed WAL store (header 2/2, no `-wal`) was opened on
  `inspectDSN` and ran `integrity_check`, which returned `"ok"`. The open **created
  `walhdr.db-wal` (0 bytes) and `walhdr.db-shm` (32768 bytes)**, and both remain after the close.
  `VACUUM INTO` output reads header 1/1 (rollback), so remote snapshots are unaffected. A
  write-DSN open converts a 1/1 file to WAL (header 2/2), confirming `:512-514`.
- **Proposed fold:** after the swap, the `--from-file` path removes the stage's `-wal` and `-shm`
  siblings, or the plan accepts them as leftovers the next clear removes. Separately: when the
  binary is upgraded between a crash and `--resume` of a `reconcile` record, resume should compare
  the store's version with the binary's again and fall to `reconcile-pending`. It should not
  reconcile an older schema. R-6bii7 `:454` keys the split on the version at write time.

---

## Plan claims confirmed by measurement or quote

- `net/http/server.go:2751-2762` at go1.27.1: the mux's `Error(w, StatusText(StatusMethodNotAllowed), ...)`
  and `NotFoundHandler()`, as quoted.
- `MaxBytesReader`: "returns a non-nil error of type *MaxBytesError for a Read beyond the limit", as
  quoted. `encoding/json`'s case-insensitive match and its "keys ... ignored" text appear in the
  go1.27.1 doc.
- A closed `*sql.DB` through `QueryRowContext` returns `"sql: database is closed"`, type
  `*errors.errorString`. The classifier yields `internal`. This closes "Not verified" `:1372`.
- flock is tied to the open file description, and duplicates share it (flock(2): "duplicate file
  descriptors (created by, for example, fork(2) or dup(2)) refer to the same lock ... released ...
  when all such file descriptors have been closed"). The exclusive lock passes to rclone through
  `dupCloexec` and `ExtraFiles` exactly as the kind lease does (`lease.go:226-236`). The store lock
  is flock, not fcntl (`lock.go:88`), so exec inheritance holds.
- rclone v1.74.3 help text matches the plan's quotes for `copyto` ("overwriting an existing file if
  it exists"), `copy` ("Doesn't delete files from the destination", "If dest:path doesn't exist, it
  is created"), `--inplace`, `--partial-suffix`, `--ignore-existing`, and `size`'s sizeless notice.
  `copy` compares "by size and modification time or MD5SUM", which Ruling 6 paraphrases correctly.
  `lsjson` exits 3 on a missing directory. A non-UTF-8 library name round-trips byte-exact through
  local-to-local `sync` and `copy`: rclone logs "Replacing invalid UTF-8 characters" but restores
  the original bytes.
- The swapped-in rollback file reads WAL after a write-DSN open (probe: header 1/1 becomes 2/2, and
  `journal_mode=wal`).
