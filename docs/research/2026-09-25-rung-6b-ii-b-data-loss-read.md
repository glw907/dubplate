# Rung 6b-ii-b: targeted data-loss read of restore

A fresh-context adversarial read of pass 6b-ii-b's restore code (`685b80f^..7a56699`, worktree
`lane-2-6b-ii-b`, head `7a56699`), hunting for ways a restore loses or corrupts data the operator did
not ask to lose. The read traced every crash window, the cancel paths, concurrent openers, the
wrong-archive states, the library copy, the displaced directory, `--discard-current`, the backup
reference rewrite, and serve's `reconcile-pending` path. It found no blocker, two majors, and five
minors. One major is proven with a scratch run of the real rclone (v1.74.3). The other follows
from the code and a spec criterion. The swap, the intent record, and the concurrency guards held up
under every window traced.

Read-only on code: no Go edited, nothing committed, no deploy, no ssh. The scratch runs used fresh
directories under `/tmp/claude-1000/` and never touched the worktree.

## Findings, ranked by consequence

### M1 (major): the first backup after a restore deletes the pre-restore copy, the only copy of the rolled-away writes

- **Site:** `internal/backup/prune.go:165-192` (`pruneLocal`) and `:196-217` (`isLocalCopyName`
  matches `pre-restore-<stamp>.db`); the copy is made at `internal/engine/store/swap.go:142-143`
  and `restore.go:620-622`.
- **Scenario:** an operator restores (`--snapshot`, `--snapshot --older`, or `--from-file`), and
  serve's schedule or the `backup` verb then runs one backup.
- **Evidence:** `pruneLocal` removes every pre-restore copy whose mtime is before the run's start.
  `os.Rename` keeps a file's mtime, so the copy carries the mtime of the old store's last write, or
  of the quiesce's journal-mode switch at the latest. Both come before any later backup starts, so
  the very first successful backup after the restore deletes the copy. Spec criterion 16 (`spec
  :1018`) asks for exactly this, borrowing ruling 7's reason, "until a backup has safely captured
  the upgraded store". That reason holds for a pre-migration copy, whose rows the migrated store
  carries forward and the next backup captures. It fails for a pre-restore copy: by construction,
  its rows are the ones the restore did not bring back, so no backup ever captures them. Writes
  made after the last backup and before the restore then exist nowhere once that backup runs. The
  same goes for the whole rolled-away state after a `--from-file` rollback beyond the remote's 14
  snapshots. Restore's output says only "moved the current store's file aside to …"
  (`internal/backup/restore.go:317`), and nothing tells the operator the copy expires at the next
  backup.
- **Why not a blocker:** it follows a ratified spec criterion, the remote still holds every
  snapshot up to the pre-restore backup, and the loss is bounded by the window between the last
  backup and the restore.
- **Narrow fix (an owner ruling):** drop the `PreRestorePrefix` arm from `isLocalCopyName`, so a
  pre-restore copy is kept like `discarded-*` and an operator removes it. Or have the backup upload
  pre-restore copies into `archive/<stamp>/` before the local prune deletes them. At the least,
  restore should print that the copy is deleted at the next successful backup.

### M2 (major, proven): with `--into-nonempty`, a local file behind a symlinked directory is overwritten and never displaced

- **Site:** `internal/backup/restore.go:381-390` (the `rclone copy … --backup-dir` call) and
  `:417-436` (`checkRoot`, which looks only at the root's top-level entries and never for
  symlinks).
- **Scenario:** a library root that holds a symlink to a directory (`root/Artist -> ../other`),
  restored with `--library --into-nonempty`, where the remote holds `library/Artist/x.flac` and
  differs from the local `other/x.flac`.
- **Evidence (a scratch run, rclone v1.74.3, local backend):** rclone lists the destination without
  following the symlink ("NOTICE: Artist: Can't follow symlink without -L/--copy-links"), so it
  treats `root/Artist/x.flac` as absent and never moves it into the backup directory. The write
  then goes through the symlinked directory. After the run, `other/x.flac` held `REMOTE-VERSION`
  in place of `LOCAL-ONLY-COPY-NEWER`, and the displaced directory was never created, because
  nothing was displaced. Ruling 6's promise ("keep every local file the copy replaces") fails
  silently. The report cannot name the file either: the path has a row and a file. The target can
  sit outside the library root, anywhere the service uid can write.
- **Narrow fix:** with `--into-nonempty`, have `checkRoot` walk the root (`filepath.WalkDir` does
  not follow links) and refuse any symlink, naming each one, before any rclone run. The walk costs
  one tree read, which the reconciliation already pays.

### m1 (minor, rclone behavior not verified): the fetched stage is never fsynced before the swap

- **Site:** `internal/backup/fetch.go:111-121`; the swap renames at
  `internal/engine/store/restore.go:716` and syncs only directories (`:673`, `:720`).
- **Scenario:** a power cut within the page-cache writeback window after a `--snapshot` restore,
  including after `Complete` has removed the record.
- **Evidence:** `--from-file` syncs its stage (`restore.go:403`). The fetch relies on rclone, and
  `rclone help backend local` lists no fsync option. This read could not confirm whether rclone's
  local backend fsyncs, since strace is absent. A directory fsync does not flush the file's data.
  Under ext4 delayed allocation, the rename can land while the file's blocks do not, which leaves
  `db/dubplate.db` at zero bytes. `classify` then routes the zero-byte file through `besideCopy`
  (`open.go:505`, `:569-584`), and serve refuses, naming the pre-restore copy or the backup
  reference. So the likely outcome is a refused start, not silent use. A partly written file would
  get past that check.
- **Narrow fix:** after the copyto, open the stage and call `File.Sync` before `VerifyStage`, as
  `copyFile` does.

### m2 (minor, not measured): when the library root is a mount point, the displaced directory sits on the parent filesystem

- **Site:** `internal/backup/restore.go:408-410` (`displacedPath` is a sibling of the root) and
  `:445-473` (the free-space check reads only the root's filesystem).
- **Scenario:** `--into-nonempty` into a root that is itself a mount point. Each displaced file is
  then a cross-device move, which rclone's local backend answers with "Can't move … trying copy",
  a full copy onto the parent filesystem. The parent is often the system filesystem that holds the
  store leaf.
- **Consequence:** no file is lost, since rclone deletes the source only after the copy. But a large
  displacement can fill the filesystem the store lives on, the copy then fails, and the record
  stays in `library`. The check prints "a mount point" (`:461-464`), but no check accounts for the
  space.
- **Narrow fix:** when `mount` is true and `--into-nonempty` is set, refuse, or require the displaced
  directory's filesystem to hold the root's used bytes.

### m3 (minor, proven): a symlinked file in the library root is replaced by a regular file and never displaced

- **Site:** as M2.
- **Evidence (a scratch run):** `root/y.flac -> ../other/y.flac` came out as a regular `y.flac`
  holding the remote bytes. The target stayed intact, and the displaced directory stayed empty.
  The symlink itself is lost, and nothing reports it. M2's `checkRoot` refusal covers this case
  too.

### m4 (minor): a relative `DUBPLATE_LIBRARY_ROOT` is refused only after the fetch and the quiesce

- **Site:** `internal/backup/guard.go:27` (`filepath.EvalSymlinks` keeps a relative path relative),
  `internal/backup/restore.go:156-160` and `:187`; the refusal comes from
  `internal/engine/store/intent.go:152-153` inside `Record` (`restore.go:626`).
- **Consequence:** no data loss. A `--library` restore fetches, verifies, and quiesces before
  `Record` refuses the non-absolute root. The quiesced store reopens in WAL mode, and the next open
  clears the stage. The config loader does not require an absolute path
  (`internal/config/config.go:350`).
- **Narrow fix:** in `Restore`, refuse a resolved root that is not absolute (or take
  `filepath.Abs` of it) before `OpenRestore`.

### m5 (minor): a resume without `--into-nonempty` copies with no `--backup-dir`

- **Site:** `internal/backup/restore.go:383-389`.
- **Scenario:** a restore stops during the library copy. The operator hand-places files in the root
  before `restore --resume`, for example to recover a few files. The resume skips the emptiness
  check (deviation D9) and overwrites any placed file that differs from the remote.
- **Narrow fix:** on a resume, pass `--backup-dir` whatever `IntoNonempty` says. On a root the
  restore alone filled, the directory stays empty and `removeIfEmpty` removes it.

## Verified sound

- **Crash before the record (steps 1 to 3).** A crash after the fetch or stage leaves only
  `staging/` debris, which the next restore (`restore.go:154-160`) or serve's `settle`
  (`open.go:330`) clears under the exclusive lock. A crash after the quiesce leaves the store in
  rollback-journal mode, and the next write-DSN open returns it to WAL. The current store is usable
  in every window before step 4.
- **Every window in step 5, including the one between a rename and its fsync.** `SwapOut` syncs
  `db/` and `snapshots/` before `SwapIn` renames the stage into place (`restore.go:670-676`), so
  the incoming file never lands where a lost set-aside rename would leave the old store.
  `swapPlan` (`swap.go:133-213`) decides every combination from the files present. A lost rename
  from a power cut only moves the leaf back to an earlier row of R-6bii8's table. The plan never
  renames over an existing pre-restore or discarded name: a present target with a present source
  is refused (`swap.go:166-182`, `:226-229`). It refuses before the incoming store lands beside any
  sidecar it does not move (`:193-211`), so a stale WAL is never applied to the restored store.
  Stamps run to the microsecond (`stamp.go:11`), so two restores never share a set-aside name.
- **Record states and the resume.** The record is written through `atomicfile.WriteStream` with a
  file and directory fsync, validated before every write (`intent.go:214-223`). Its removal syncs
  the leaf (`:227-232`). `Advance` re-derives the version from `db/` on every step
  (`restore.go:834-850`), and a resume in `swap` re-verifies the stage without removing it on
  failure (`restore.go:427-428`, `:684-697`). An unreadable record refuses both a new restore and a
  resume.
- **SIGINT mid-copy or mid-fetch.** The transport sends SIGINT to rclone in its own process group
  and kills it after 10 s (`rclone.go:76-83`). A cancel during the fetch removes the stage, and no
  record exists yet. A cancel during the library copy leaves the record in `library`, and the resume
  reruns the copy. Files already displaced stay in the displaced directory, since `removeIfEmpty`
  never removes a non-empty directory (`restore.go:548-560`). SIGHUP is caught
  (`cmd/dubplate-server/restore.go:172`), and any error after a cancel exits 130.
- **Concurrent serve, backup, a second restore, or an orphaned rclone.** The lock is taken
  non-blocking and exclusive (`restore.go:137`). Serve, write mode, a backup, and the orphaned rclone
  of a backup or a restore (through the inherited flock descriptor) all hold the lock and make the
  restore exit 75. An opener that waits through a restore re-reads the record once it holds the
  lock (`open.go:161-164`, `:277-281`, `:308-315`).
- **Wrong archives.** A truncated file, a non-SQLite file, a zero-byte file, a file without the
  marker, a newer file, and a file with a divergent ledger are each refused before the record, and
  the stage is removed (`restore.go:421-487`, `snapshot.go:140-177`). An older schema goes to
  `reconcile-pending`. A snapshot from another store is the declared 6d gap. rclone's own
  size and hash check, plus `integrity_check`, covers a partial fetch.
- **Full disk.** The stage space is checked first (`fetch.go:103-109`, `restore.go:346`). A failed
  copyto or `copyFile` removes the stage. A full disk during the library copy fails with the record
  in `library` and loses nothing.
- **`--discard-current`.** The move opens nothing and runs only after the record names it. It moves
  `-wal` and `-shm` with the database under matching names, so the discarded store stays readable
  with its WAL. No prune or guard matches the `discarded-` prefix. A `-journal` is refused before the
  record (`restore.go:548-551`).
- **The backup reference rewrite.** It is atomic and runs after the swap. Its prune removes every
  other reference, so an `--older` stamp takes effect. A crash before the prune leaves two
  references, and the resume rewrites and prunes again.
- **`--from-file` sources.** The live database is detected by `os.SameFile` after resolving
  symlinks, which catches hard links and bind mounts (`restore.go:372-378`). A source inside
  `staging/` is refused before the clear (`:203-235`). A source beside a `-wal` or `-journal` is
  refused.
- **Serve's `reconcile-pending` path.** It migrates, reconciles, writes the report, and only then
  removes the record (`open.go:342-371`). A crash anywhere in that sequence leaves the record, and
  the next start redoes it at head. A nil reconciler refuses and leaves the record. OpenServe has no
  production caller yet (6d), so this path runs only in tests today, as the plan's exit gate 14
  states.
- **An empty or zero-byte `db/` after a failed swap never becomes a fresh store.** `besideCopy`
  refuses while a pre-restore copy or a reference exists (`open.go:569-584`).
