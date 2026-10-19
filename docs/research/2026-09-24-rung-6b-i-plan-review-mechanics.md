# Rung 6b-i plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-24-rung-6b-i-ops-seam-backup.md` at `b36871c`
(revision 1, draft). Lens: does every mechanism behave as the plan states, and can the task
order build. Line numbers are the plan's at `b36871c`. 6a code was read at `lane-2-6a`
(`aa3b13e`). Probes ran under `/tmp/claude-1000/rv6bi/` on 2026-09-24 against rclone v1.74.3,
go1.27.1, modernc.org/sqlite v1.59.0, and golangci-lint 2.13.2 with the repo's `.golangci.yml`.

## Counts

| Severity | Count |
| --- | --- |
| Blocker | 0 |
| Major | 5 |
| Minor | 9 |
| Owner fork | 1 |

## Claims verified as stated

These hold. No fold is needed.

- **R-6bi7 (`:405-423`).** After a first sync, one source file was removed and
  `rclone sync src remote/library --backup-dir remote/archive/s1 --max-delete 0` ran. It exited
  7 with `Got fatal error on delete: --max-delete threshold reached` and
  `Failed to sync with 2 errors: last error was: failed to delete 1 files`. `archive/` was not
  created, and `library/` kept all three files. A modified file under `--max-delete 0` logged
  `Moved (server-side)` and `Server Side Moves: 1`, and the same run still exited 7 on the
  delete. `rclone help flags`: `--max-delete int  When synchronizing, limit the number of
  deletes (default -1)`.
- **R-6bi9 (`:449-461`).** `rclone backend features` returned these `Features` values: `local`
  had Copy false and Move true. `http` had both false. `crypt` over `local` had Copy false and
  Move true. `s3` had Copy true and Move false. A name with no `_TYPE` exited 1 with
  `didn't find section in config file ("zz")`. A crypt wrapping an undefined remote exited 1
  naming `nope`. Wall time was 0.10 to 0.25 s for every offline backend.
- **`copyto -M --metadata-set` (`:953-957`).** `run_id` and `sync_completed` read back through
  `rclone lsjson -M` on a `local` remote.
- **R-6bi8's unused-constant premise (`:113-116`).** An unexported constant with no reader
  fails golangci-lint: `const snapshotKeep is unused (unused)`. JSON-tagged and untagged
  exported fields on a wire type that nothing reads did not fire `unused`. So Task 2's wire
  types are safe under `exported-fields-are-used: false`.
- **R-6bi2's `flock(2)` quote (`:344-351`).** In one process, a second `open` of the file
  answered `EWOULDBLOCK` to `LOCK_EX|LOCK_NB` while the first descriptor held the lock.
- **Inherited `flock` survives the parent's death (`:503-509`).** A parent took the lock and
  started a child that inherited the descriptor. After the parent got SIGKILL, `flock -n` on
  the file still failed.
- **`VACUUM INTO` consistency.** With another connection holding an uncommitted insert,
  `VACUUM INTO ?` finished in 2 ms, and the copy held none of the uncommitted rows.
- **`wal_checkpoint(TRUNCATE)` reports busy in its row (`:401-403`).** With a reader or writer
  open, the pragma returned `err=nil busy=1`, not a Go error.
- **`os/exec` `WaitDelay` and a custom `Cancel` (`:486-494`).** A child that ignored SIGINT got
  `signal: killed` at `WaitDelay`.
- **Task order.** Tasks 1 and 2 build alone. Task 1's exports warn under the staged
  exported-unused leg (`scripts/check.sh`, `CHECK_EXPORTED_UNUSED` defaults to warn) and never
  fail the gate. The exceptions are M3 and M4.

## Major

### M1: Task 1's copy fixture cannot be satisfied (Major)

**Location:** plan `:688-693` (Task 1 criterion 3).

**Defect.** The fixture asks for three things at once: "a writer inserting rows in a loop, and
one transaction held open across the copy". The copy must also hold "no row of a transaction
left open during it". For the open transaction to prove anything, it has to hold uncommitted
rows. That makes it a write transaction, and it holds SQLite's single write lock for as long as
it is open. So the looping writer cannot commit during the copy.

**Evidence.** The probe held an IMMEDIATE transaction with one insert open on a second
connection. A write on the first connection returned `database is locked (5) (SQLITE_BUSY)
after 2.005s` with `_busy_timeout=2000`. The store's default busy timeout is 15 s
(`cmd/dubplate-server/store.go:22`), and `update` refuses with `errBusy` once that wait ends.
The implementer can only pass the criterion by changing it, which costs a fix round, or by
quietly weakening it.

**Proposed fold.** Split the criterion into two fixtures.

- **Fixture A:** the writer loops through `update`, and the copy runs mid-loop. Every row
  committed before the copy began is present in the copy.
- **Fixture B:** a second handle holds an uncommitted insert across the copy, and no writer
  loops. The copy holds none of that insert's rows, and the copy finishes. The probe measured
  2 ms.

### M2: the backup copy's `TRUNCATE` checkpoint stalls serve's writers for the full busy timeout (Major)

**Location:** plan `:401-403` (R-6bi6), `:642-643`, `:673`.

**Defect.** 6a's `takeSnapshot` checkpoints only under serve's exclusive lock, where nothing
contends. 6b-i's backup copy runs under the shared lock with serve live. A
`wal_checkpoint(TRUNCATE)` that finds any open reader waits the whole busy timeout, and new
writers wait behind it for that whole time. The plan treats a busy result as benign ("logged,
never a failure") and misses the stall. `update`'s own doc says that "a deadline shorter than
the busy timeout does not shorten" another process's lock wait. So a serve request can hang
for up to 15 s on every backup whenever a reader is mid-transaction.

**Evidence.** The probe held a read transaction open and started `TRUNCATE` on a maintenance
connection with a 2 s busy timeout. A writer that started 100 ms later waited `1.932s`, and
the checkpoint returned `busy=1 after 2.004s`. The same probe then set
`PRAGMA busy_timeout=0` on that one connection first. The checkpoint returned `busy=1 after
0s`, and the writer finished in `0s`.

**Proposed fold.** Task 1's backup copy sets `PRAGMA busy_timeout=0` on its maintenance
connection before the `TRUNCATE` checkpoint. It holds that connection through `db.Conn` so
the pragma and the checkpoint share one connection. It stays a `TRUNCATE` checkpoint, so the
spec's wording holds and no deviation is needed. Add a row: a writer that starts during the
checkpoint completes within a stated bound, far below the busy timeout. This is spec-shaped
but costs one pragma, so fold it without an owner question.

### M3: no task wires the backup body into `ops/local`, and the file lists forbid it (Major)

**Location:** plan `:242-246` (the import table), `:899-901`, `:937-939` (Task 4 files),
`:1114-1119` (Task 6 files); spec `:198-205`.

**Defect.** The plan and the spec both say that `internal/ops/local` imports `backup`, and
Task 3 is the only task that writes `ops/local`. When Task 3 runs, `internal/backup` does not
exist yet, so the body its tests run is a test-supplied function (`:899-901`). Task 4 creates
the backup body, but its Files hold only `internal/backup/*`. Task 6's Files hold only `cmd/`
and `keys.go`. So no task has the file territory to make a production `ops/local` start run
`backup`'s body. There are two ways out.

- `cmd/dubplate-server` injects the body as a function, which contradicts the stated import
  edge.
- An implementer edits `ops/local` outside its Files, which the diff-reviewer flags as scope
  creep.

The same gap blocks Task 4's orphan row (`:1007-1011`), which needs a real runner holding the
kind lease around the backup body. That row must live in the external test package
`backup_test`. An in-package test that imports `ops/local` is an import cycle once `ops/local`
imports `backup`.

**Proposed fold.** Choose one before approval.

- **Option (a):** Task 4's Files add `internal/ops/local/jobs.go` and its test, for the
  kind-to-body binding.
- **Option (b):** state that `ops/local` takes a kind-to-body map at construction, that `cmd`
  composes it, and that the import table's `ops/local` to `backup` edge becomes
  `cmd`-composed.

Either way, name `backup_test` as the package for the orphan row.

### M4: the child's inherited descriptors have no export in Task 1, and the store is "Task 1 only" (Major)

**Location:** plan `:355-363` (R-6bi3), `:503-509` (R-6bi12), `:226` (the store is edited in
Task 1 only), `:1025-1026` (mutation 6).

**Defect.** R-6bi12 passes two descriptors to rclone through `ExtraFiles`: the kind lease's
and the store lock's. They are `*os.File` values that the store owns. `storeLock.file` is
unexported (`lock.go`), and so is `handle.lock`. R-6bi3's export list names the write-mode
open, close, the write method, the read path, the lease primitive, the id function, and the
backup copy. It names no accessor for either descriptor. Task 4 builds the transport, but the
constraints bar it from editing `internal/engine/store`. So Task 4 cannot pass the store lock
to the child without leaving its territory. The lease primitive's return type is unspecified
too, so whether it exposes its file is left to Task 1's implementer, who has no caller in view.

**Proposed fold.** Add one outcome to R-6bi3 and Task 1: the store exports what a transfer
child must inherit. Either the handle and the kind lease each gain a method returning the
descriptors, or one method returns both as `[]*os.File`. Task 1's test proves that a child
started with them keeps the kind lease busy after the parent's descriptor closes. Adding this
now keeps the store change inside Task 1.

### M5: joining the remote and its subpaths by string puts the layout in the wrong place (Major)

**Location:** plan `:952-955` (the `<remote>/library`, `<remote>/archive/<stamp>`, and
`<remote>/snapshots/<stamp>.db` forms), `:1207-1213` (Task 7 validates only the colon and the
name).

**Defect.** The plan writes each destination as `<remote>/<subpath>` and states no join rule.
`DUBPLATE_BACKUP_REMOTE` commonly has an empty path, such as `mycrypt:` or `b2bucket:`. Plain
concatenation then yields `name:/library`, which is a different path.

- **`local` and `sftp`:** the path is absolute from the filesystem root, not relative. A
  non-root run most likely fails with permission denied.
- **Bucket backends:** the first path element is the bucket. So `name:/library`,
  `name:/archive/...`, and `name:/snapshots/...` name three different buckets.

A value with a trailing slash produces `//`. No check or test names the empty-path state, which
is planning-miss item 2.

**Proposed fold.** State the join rule: split the value at the first colon, then form
`name:` + `path.Join(rest, "library")`, and the same for the other two subpaths. An empty
`rest` gives `name:library`. Add rows for `name:`, `name:dir`, and `name:dir/` to Task 4's
transport table, and pin each argv in a stand-in test. Decide whether Task 7's `doctor`
accepts an empty path; the fold recommends accepting it.

## Minor

### M6: the Setpgid proof can pass without `Setpgid` (Minor)

**Location:** plan `:1167-1172` (criterion 20's second row); Task 4 mutation 5 at `:1023-1024`.

**Defect.** The row claims that the group-signal case "fails without `Setpgid`". Without
`Setpgid`, the stand-in dies at once, and the verb writes `cancel_requested` within
milliseconds. If the runner re-reads `cancel_requested` after a transfer error, the job still
ends canceled, and that recheck is a plausible robustness choice. In that case the row passes
with or without `Setpgid`. The test also has to start the verb in its own process group, or
`kill(-pid)` answers `ESRCH`. If the test signals its own group instead, it interrupts
`go test` itself. The plan names neither requirement.

**Proposed fold.** Discriminate on the stand-in's lifetime. After the group SIGINT, assert the
stand-in is still alive at a point shorter than the cancel poll interval (for example 500 ms
against 2 s). Assert that it exits only after the runner's `Cancel`. Start the verb child with
`Setpgid: true` and signal `-pid`.

### M7: a canceled transfer's `Wait` returns the exit status, not `ctx.Err()`, and rclone exits 130 (Minor)

**Location:** plan `:958-961`, `:978-980`, `:1468-1470` ("Not verified").

**Evidence.** `go doc os/exec.Cmd`: "If the command exits with a non-success status ... Wait
and similar methods continue to return the command's usual exit status." The probes:

| Child | What `Wait` returned |
| --- | --- |
| A child SIGINT killed | `signal: interrupt`, `ExitCode()` = -1 |
| A child that exited 3 on SIGINT | `exit status 3` |
| Real rclone mid-sync | `exit status 130`, within 20 ms, with no partial file left |

**Defect.** A runner that maps a nonzero exit straight to `rclone-exit-<N>` records a
canceled job as failed with `rclone-exit-130` or `rclone-exit--1`. Criterion 20's stand-in
would catch this, so it is not a blocker. But the plan quotes neither behavior (planning-miss
item 1). It also has no row for rclone dying from a signal nobody sent through `Cancel`, such
as the OOM killer, which reads as `rclone-exit--1`.

**Proposed fold.** Quote the `Cancel` paragraph. State that the job context's error decides
canceled before the exit code does. Name a failure code for a death by signal, for example
`rclone-signal-<name>`. Move "rclone on SIGINT" from "Not verified" to measured.

### M8: `/proc/locks` names a dead pid as the orphan conflict's holder (Minor)

**Location:** plan `:850`, `:1007-1011`.

**Evidence.** A parent took a lock through `flock -x 9` and started a child that inherited
fd 9. After the parent got SIGKILL, `/proc/locks` read
`FLOCK ADVISORY WRITE 286916 00:3c:...`. That pid belongs to the process that acquired the
lock, which was dead. The child holding the lock was never named. `lockHolders`
(`lock.go`) reports that pid as-is.

**Defect.** In the orphan case, "conflict naming the holders" hands the operator a pid that no
longer exists. The live rclone goes unnamed, so the operator cannot use the pid to find it.
On btrfs the holder list is empty anyway (6a handoff item 7).

**Proposed fold.** When the kind lease is busy and no queued or running record exists, send
the fixed text ("a previous run's transfer is still running") without pids. Alternatively,
label the pids "acquired by". The orphan row asserts the text, not a pid.

### M9: a concurrent lease try can refuse on its own removal, and the constructor fails (Minor)

**Location:** plan `:636-638`, `:666-669`, `:845`.

**Evidence.** Process A and process B each open the same dead lease file. A locks it, unlinks
it, and closes it. B's `flock` on its still-open descriptor then succeeds, because it locks
the unlinked inode. B's `os.Remove` of the path then fails with `no such file or directory`
(`f2 remove path: ... no such file or directory` in the probe). The plan says a try "refuses
on any other error naming the path", and an unreadable lease refuses the constructor. So a
serve start and a `backup` start that recover at the same moment can fail one of them for no
real fault.

**Proposed fold.** After a successful lock, a not-exist error from the remove reads as free.
Optionally, compare the locked descriptor's inode with the path's before removing the file.

### M10: instance lease files are never removed on a clean exit (Minor)

**Location:** plan `:301`, `:631-635`, `:662`.

**Defect.** Recovery tries only the leases that non-terminal records and staging files name.
The plan never says `ops/local`'s close removes its own `leases/<instance>`. So every run
leaves one file behind: each `backup` verb run, each flag-off run, and each crashed process
with no open record. With a daily timer, `leases/` grows by at least one file a day, and
nothing ever removes them.

**Proposed fold.** The handle's close unlinks its instance lease while still holding the lock,
then closes it. Recovery also sweeps every `leases/<32 hex>` file it can lock. Add a row:
after a verb run, `leases/` holds only the kind lease.

### M11: the conflict answer can name a dead job as running (Minor)

**Location:** plan `:820-825`, `:849-850`.

**Defect.** A start takes the kind lease before it runs recovery. When the lease is busy,
the conflict is answered from records that recovery has not yet processed. Take an
`ops/local` built before its sibling process died, with an orphaned rclone holding the lease.
Its conflict names the dead job's handle as running. The orphan row's expected "naming
holders" therefore depends on when the fixture builds its runner.

**Proposed fold.** When the lease is busy, try each candidate record's owner lease before
naming it. If the owner is dead, fall through to the holders text. Pin the orphan row's
fixture to build its runner after the kill.

### M12: `forbidden` and `internal` are exported before any production caller (Minor)

**Location:** plan `:385-389`, `:749-751`, `:1344-1345` (exit gate 11).

**Defect.** Task 2 builds seven sentinels, but 6b-i raises only five of them. That conflicts
with the plan's own rule (`:253-256`), with the build-for-2.0 refinement ("nothing is
exported ahead of a caller"), and with exit gate 11. The exported-unused leg reads names, so a
code table in a second file of `ops` keeps the other two looking alive. The gate would pass
while the rule is broken.

**Proposed fold.** Declare five sentinels. 6b-ii adds `forbidden` and `internal` with its
contract suite. Alternatively, record the two as a named exception to exit gate 11.

### M13: the features probe's 10 s bound turns a B2 credential error into "not checked" (Minor)

**Location:** plan `:74-85` (Ruling 2), `:1229-1231`.

**Evidence.** A B2 remote with bad credentials, run through `rclone backend features b:`,
took 13.7 s and exited 1 at the default `--low-level-retries 10`. With `--low-level-retries 1`
it took 2.1 s and exited 1 with `failed to authorize account ... 401 bad_auth_token`.

**Defect.** Under the planned bound, a credential typo on B2 reads "not checked" and exit 0.
Ruling 2's case for the probe is that it catches a remote an adopter configured wrong, and
this is the probe's most useful catch.

**Proposed fold.** Run the probe with `--low-level-retries 1`. Record the B2 timing as
evidence under Ruling 2.

### M14: two smaller state gaps (Minor)

**Crypt over a plain path.** At `:1211-1212` and `:1227`, `doctor` requires a crypt remote's
`_REMOTE` to name a remote with its own `_TYPE`. rclone also accepts a plain path, and
`RCLONE_CONFIG_C_REMOTE=/tmp/x` probed as Copy false and Move true with exit 0. It also
accepts a connection string such as `:s3,...:`. `doctor` would fail both valid configurations.

- **Fold:** a `_REMOTE` value with no colon, or one starting with `:`, skips the wrapped
  `_TYPE` check and lets the features probe decide.

**Same-instance stage.** Per `:671` and `:983`, a stage whose removal failed is only logged.
The next backup in the same process then refuses with "the stage already exists", and
recovery cannot clear the stage because its instance lease is live. That process is the
verb in 6b-i, and serve from 6b-ii on.

- **Fold:** a start first removes its own instance's leftover stage, which the instance lease
  proves is its own.

**Orphan lifetime, a note only.** rclone is a Go program. `go doc os/signal`: "A write to a
broken pipe on file descriptors 1 or 2 ... will cause the program to exit with a SIGPIPE
signal." Once the verb dies, an orphaned rclone whose stderr pipes to the verb exits at its
next log line. That makes the orphan window shorter than R-6bi12 implies, and the orphan
sync ends partway. It is harmless, since the next run resumes. The HISTORY entry should
record it so that a later pass does not treat the orphan window as long-lived.

## Owner fork

### F1: which exit code a job canceled from outside the verb returns

**Location:** plan `:86-96` (Ruling 3), `:1155`.

Ruling 3 maps every canceled job to 130, which is 128 plus SIGINT. A job can also end
canceled in two ways that involve no SIGINT: SIGTERM or SIGHUP (conventionally 143 and 129),
or another process's cancel operation, where the verb received no signal at all.

- **(a)** 130 for every canceled job, as planned. Rung 16's timer and rung 9's ping read one
  code.
- **(b)** 128 plus the signal received, and a separate code for a cancel from elsewhere.
- **(c)** 130 for a signal, and 1 for a cancel from elsewhere, treating it as a failure from
  the timer's point of view.

**Recommendation:** (a). One canceled code keeps rung 16 simple, and the job record carries
who canceled. The help text should say that 130 means "canceled by any means".

## Planning-miss checklist

| Item | Result |
| --- | --- |
| 1. Library behaviors quoted | Mostly quoted. Unquoted: `Wait` after `Cancel` (M7), holder pids in `/proc/locks` (M8), and `TRUNCATE` blocking writers (M2). |
| 2. States named | Missing: an empty remote path (M5), rclone killed by a signal (M7), a race on lease removal (M9), a leaked lease file (M10), and a leftover stage in the same instance (M14). |
| 3. Privileged-field channels | Adequate. Lease file removal at close is missing (M10). The inherited descriptors are an unlisted channel (M4). |
| 4. Fixture states that rule out vacuous passes | Two gaps: Task 1's copy fixture cannot be satisfied (M1), and the Setpgid row can pass without `Setpgid` (M6). |
