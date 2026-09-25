# Rung 6b-ii plan review: domain risk (data integrity and security)

Reviewed: `docs/superpowers/plans/2026-09-25-rung-6b-ii-restore-http.md`, revision 1 (uncommitted).
The lens covers domain risk only: data integrity on the restore path and security on the HTTP
projection. I checked the plan against the rung 6 spec (revision 2.2), the configuration homes spec,
the 2.0 spec's hooks, and CLAUDE.md. I read the 6b-i code in the lane 2 worktree at `33ceff5` and
rung 5's gate on master.

Findings are ranked by consequence. Each one gives the defect, the proposed fold, and any owner fork.
A fold may refuse a finding whose fix costs more than the risk it removes.

**Counts:** 1 blocker, 8 major, 13 minor. There is one sizing owner fork, and three majors carry
owner forks.

Citation spot-check: 40 code pointers and 22 spec pointers were resolved. All of them land on the
content they name except two. First, spec `:698` does not say "0600" for the intent record. Spec
`:692` says 0600 for the stage, and the record's mode comes from `atomicfile` instead, which is
harmless. Second, Ruling 4 says that exit 75 reuses `backup`'s meaning. In `backup`, 75 is
`exitBackupConflict`, a busy kind lease (`cmd/dubplate-server/backup.go:71-74`), and a timed-out
store-lock wait there exits 1. The meaning is close (tempfail), but the text should say so (m13).

---

## Blocker

### B1. A writer that waits out a dead restore opens the store under its intent record

**Location:** plan R-6bii7 and Task 4A (`open.go` edits); code at
`.claude/worktrees/lane-2-6b-i/internal/engine/store/open.go:116-162` (`OpenWrite`) and
`:192-247` (`OpenServe`).

**Defect.** Every writing open checks the record in `leaf.stat()` (`open.go:361`), before it takes
the lock. `OpenWrite` then waits on `acquireShared` for up to its context's bound. `backup` gives
that bound as `backupVerbWait` (30 s, `backup.go:80`). After the wait, `OpenWrite` classifies the
store and opens it with no second record check. `OpenServe`'s shared path (`:221-231`) does the same
for a current store: it returns a handle and never reaches `settle`'s `checkIntent` (`:257`). Only
the exclusive path re-checks.

A restore holds the exclusive lock, so a `backup`, a `member` verb, or a restarting `serve` queues
behind it. Suppose the restore then dies, or exits 130 by design (Ruling 4), and leaves the record.
The queued writer then wins the lock and opens the store anyway:

- **`swap` state (after step 4, before step 5).** The writer writes into the old database.
  `restore --resume` then renames that database to `pre-restore-<stamp>.db`. The next backup's
  `pruneLocal` deletes that copy, since its mtime predates the run (`prune.go:172-200`). The rows
  are then gone from the live store and, soon, from the leaf, with no report.
- **`library` state (half-copied tree).** A backup runs `rclone sync` of the partial tree to
  `library/`. If the file count clears the guard fraction, rclone moves the "extra" remote files
  into `archive/<stamp>/`. `--resume` re-runs `rclone copy <remote>/library`, which no longer holds
  them. Library files go missing silently. They survive only under `archive/`, and the
  reconciliation reports them only as rows without files.
- **`reconcile` state (current version).** `serve`'s shared path opens without reconciling, and the
  record never clears through serve.

The spec requires "Every writing open checks for a `restore-intent` record first" (`:240`). That
requirement holds only if the check runs under the lock. This is 6a code, but it becomes reachable
only when 6b-ii writes records. Task 4A already edits `leaf.stat`, `checkIntent`, and `OpenServe`.

**Proposed fold.** Task 4A adds an outcome. Every writing mode re-reads the record after its lock is
won: `OpenWrite` after `acquireShared`, and `OpenServe` after each shared `waitLock`. Any record on
`OpenServe`'s shared path releases the lock and routes to the exclusive path, where `settle` decides.
That same routing is what makes Task 4A's "reconcile-pending, store current" row reachable. Add a
state row and a named mutation. In the test, a second process takes the exclusive lock, and the
writer under test starts waiting. The second process then writes a `swap` record and exits. The
writer must refuse, naming `restore --resume`. The mutation deletes the under-lock check, and the
writer then opens. Inspect's pre-lock check is read-only and can stay.

---

## Major

### M1. Resume cannot finish a swap when no current store existed, which is the disaster-recovery case

**Location:** R-6bii8's four-combination table; Task 4B "Current store | absent" row; Task 5 crash
table.

**Defect.** The table assumes that a current database existed. On a fresh box (leaf created by
restore, R-6bii6), the record names no pre-restore copy. After step 4, `db/` is absent and the
incoming file is present. After the rename, `db/` is present and the incoming file is absent. Both
states fall under "any other combination → refused," so a crash anywhere in step 5 of a fresh-box
restore leaves an unresumable record. Write mode and serve then refuse forever (see M5). Criterion 7's
fixture ("a current store with rows") never meets this case.

**Proposed fold.** Branch the table on whether the record names a pre-restore copy, and add the two
no-current rows: "absent, n/a, present → rename, fsync" and "present, n/a, absent → fsync." Repeat
Task 5's crash rows over a no-current fixture. Add a mutation that drops the branch, so the
no-current resume refuses.

### M2. A resume under a different binary acts on a version verdict the old binary made

**Location:** R-6bii7 state table, Task 5 "Resume".

**Defect.** The record carries "the incoming store's version," and the state (`reconcile` or
`reconcile-pending`) is chosen against the binary that wrote it. Consider an operator who upgrades
between the crash and `--resume`:

- A `reconcile` record under a newer binary runs step 7 against a store below that binary's
  version. Either the reconciliation queries an unmigrated schema, or its handle open refuses as
  older, and resume is stuck.
- A `swap` record under an older binary swaps in a file above the running binary's version, because
  the version check ran only at step 2. The spec's refusal of a newer snapshot (`:691-692`) is then
  bypassed.

The format-number rule protects the record's shape, not these semantics.

**Proposed fold.** Resume re-derives the version decision from the running binary before it acts.
For `swap`, it re-verifies that the staged file is at or below the head, and otherwise refuses,
naming both versions. For `reconcile` with a store now below the head, it advances to
`reconcile-pending` and names "start serve." Add two rows to Task 5 through the synthetic migration
set (`migrate.go:91`), which is already used for criterion 10.

### M3. `--discard-current` moves the database away before any record exists, to an unnamed place

**Location:** R-6bii8 and Task 4B, "`--discard-current` move … each named"; spec flow step 3.

**Defect.** The plan names neither the destination nor the names, and the move happens at step 3,
before step 4 writes the record. A crash between the move and the record leaves `db/` empty with no
record. Three things can follow:

- Whether `serve` (from 6d) refuses depends on `besideCopy` (`open.go:487-500`). It refuses only
  when the moved-aside name matches `isCopy` or a backup reference exists. Otherwise serve creates
  an empty store and clears the staged incoming file.
- If the moved-aside file takes the `pre-restore-<stamp>.db` name, three problems follow. It
  collides with step 5's own rename. `besideCopy` then tells the operator to `restore --from-file`
  a file known to be unopenable. `pruneLocal` deletes the `.db` and orphans its `-wal` and `-shm`,
  since `isLocalCopyName` matches `.db` only (`prune.go:204-222`).
- A discarded `-wal` may hold the only committed copy of recent rows.

**Proposed fold.** Write the record before the discard: step 4 first, with a `discard` field naming
the three destination names. Name them distinctly, for example `snapshots/discarded-<stamp>.db`,
`.db-wal`, and `.db-shm`. Make `besideCopy` treat a discarded name as a copy, and keep `pruneLocal`
from ever deleting one. Resume's combination table then covers the discard. Add crash rows after
the discard and before the swap.

### M4. The swap renames across `db/`, `staging/`, and `snapshots/`, which may sit on different filesystems

**Location:** R-6bii8 (step 5 renames), R-6bii10 ("the swap renames and the pre-restore copy costs
no space").

**Defect.** Erratum E6 already contemplates a symlinked `db/` whose target is on another filesystem
(spec, E6: "statfs follows a symlinked `db/` to its target"). Across devices, `os.Rename` fails with
`EXDEV`. The failure comes after step 4 has written a `swap` record, and every resume fails the same
way, so write mode and serve refuse until someone edits the leaf by hand. R-6bii10's claim that the
swap costs no space also holds only on one device.

**Proposed fold.** At step 1, the session open compares `st_dev` of `db/`, `staging/`, and
`snapshots/`, after resolving symlinks. On any mismatch it refuses, naming the directories, before
any stage or record exists. It checks free space on `staging/`'s filesystem. Add one row to Task 4B;
the `statfs` seam cannot fake a device, so use a bind-mount-free fixture: a `db/` symlink to a tmpfs
path when one is available, or a device-comparison unit test over a seam.

### M5. No path out of an unresumable record, and a new restore beside a record is not stated

**Location:** R-6bii6, R-6bii7, Task 7 flag table.

**Defect.** The plan does not say what `restore --snapshot` or `restore --from-file` does when a
record already exists. If it proceeds, it overwrites the record and the staging the record names. If
it refuses, three states leave the operator with no verb that proceeds: R-6bii8's "any other
combination → refused," M1's no-current states, and M4's `EXDEV`. Write mode, serve, resume, and a
new restore all refuse. The one way out is hand-deleting `restore-intent`, which CLAUDE.md's safety
rails forbid calling done.

**Proposed fold.** State the rule: a new restore beside any record refuses, naming `--resume`. Add a
Task 7 row for it and a mutation that proceeds instead.

**OWNER FORK.** How does an operator leave a record that resume refuses?

- (a) A `restore --abandon` flag, a spec deviation. It removes the record only when `db/` holds a
  store that verifies. When `db/` does not, it first renames the named pre-restore copy back. It
  never deletes the staged file.
- (b) A runbook line owed to rung 16 naming the manual steps. The refusal text lists every file the
  record names and whether each exists, which the plan already requires.
- Recommendation: (b) at 6b-ii. M1, M3, and M4 remove the reachable unresumable states, so what
  remains is corruption outside the flow, and a verb for that is not worth a spec deviation now.
  Revisit (a) if rung 16's cutover rehearsal meets one.

### M6. `--into-nonempty` with plain `rclone copy` overwrites newer local files with no copy kept (Ruling 6)

**Location:** Ruling 6; R-6bii11; Task 6 "non-empty, with `--into-nonempty`" row.

**Defect.** Ruling 6's rationale says the transfer "never deletes." That is true, but `rclone copy`
replaces any destination file whose size or modtime differs, whichever side is newer. It does not
consult `--update`. A FLAC retagged, or uploaded, after the last backup is overwritten by the
backup's older bytes, with no `--backup-dir`, so its local version is unrecoverable. The
reconciliation report cannot name it, because the path still has both a row and a file. The ruling's
text presents a destructive option as the safe one.

**Proposed fold.** Rewrite Ruling 6's reason to state the overwrite consequence, and add a Task 6
row: a local file that differs from the remote, with its outcome asserted.

**OWNER FORK.** With `--into-nonempty`, what happens to a local file that differs from the backup?

- (a) `rclone copy`: overwritten and lost, as drafted.
- (b) `rclone copy --ignore-existing`: kept, and the backup's version is not restored.
- (c) `rclone copy --backup-dir <leaf-sibling>/restore-displaced-<stamp>`: overwritten, and the
  displaced local version is kept. rclone's condition from spec `:629-631` applies: same remote, no
  overlap with the destination. On the `local` backend that means a directory on the root's own
  filesystem outside the root. The report names the displaced directory.
- Recommendation: (c). It costs one flag and one free-space term, and it is the only option in which
  a mistaken `--into-nonempty` destroys nothing.

### M7. Review focus 8 claims a test for an unmounted mount point that no task has

**Location:** "Review focus" item 8; R-6bii11; Task 6's table.

**Defect.** Focus item 8 says an unmounted mount point is "refused, never populated. Pinned in Task
6." An unmounted mount point is an existing, empty directory. Task 6 refuses only an absent root or
a non-empty root without the flag, so the empty underlying directory is populated on the root
filesystem. The library free-space check then runs against the wrong filesystem, and a small library
passes it. The step 7 reconciliation reads the underlying directory and reports clean. When the
volume mounts, it shadows the restored tree, and the catalog's files vanish. The plan asserts
coverage that does not exist, so the targeted data-loss reader and the owner would both believe the
case is closed.

**Proposed fold.** Either strike focus item 8's claim and declare the gap in "Not verified," with a
runbook line owed to rung 16, or add a cheap signal. The cheap signal: before the copy, print the
resolved root's mount point and filesystem type from `/proc/self/mountinfo`, and refuse an empty
root that is not itself a mount point while its parent is on the same device as `/` without
`--into-nonempty`. That rule may be too clever; declaring the gap honestly is the minimum.
Recommendation: declare the gap and print the mount line.

### M8. The authorizer's input and the actor's single channel are not specified across the two legs

**Location:** R-6bii2 and R-6bii3; privileged-fields table row "An HTTP caller's actor"; D1.

**Defect.** R-6bii2 says the authorizer "takes the request and the operation's privilege." The
authorizing layer, however, is an `ops.Seam` decorator (R-6bii3). Its methods receive
`(ctx, actor, req)`, not an `*http.Request`, and the in-process contract leg (D1) has no HTTP
request at all. The plan therefore never fixes three things:

- what the authorizer actually receives;
- what the HTTP layer passes as the `actor` argument into the decorator;
- that the decorator discards its caller's `actor` and substitutes the authorizer's answer.

A decorator that forwards its caller's actor when the authorizer admits opens a second channel for
the most privileged field on the surface. The actor row in Task 3 would not catch that decorator if
the HTTP layer happens to pass the same value.

**Proposed fold.** Pin the shape in R-6bii2. The authorizer takes `(ctx, op name, privilege)` and
returns `(ops.Actor, error)`. The verified identity reaches it only through `ctx`, which rung 14's
session middleware sets, the way `internal/auth/admin.go`'s `adminKey` does, and tests set in both
legs. The decorator ignores the `actor` parameter it receives and passes only the authorizer's
actor downstream. The HTTP layer passes the zero `Actor`. Add a Task 2 row and a mutation: the
decorator forwards the caller's actor, and a test passing a planted `LocalActor` through the
decorator finds it in the `jobs.actor` column.

---

## Minor

- **m1. Shared origin with the Navidrome fallthrough is missing from the rung 12 and 14 handoff.**
  Task 8's handoff. STATUS carries 5c-ii security M2 ("the roster API shares an origin with the
  fallthrough") to rung 12. `originCheck` admits `Sec-Fetch-Site: same-origin`
  (`internal/serve/origin.go:13-24`), so script running in the proxied Navidrome UI on the same
  origin can POST JSON with the session cookie. The fallthrough strips the cookie on the way to
  Navidrome, but that does not stop a same-origin script. Once rung 14 mounts it, the projection is
  a second admin-privileged, cookie-authenticated, state-changing API under that question. Fold:
  one handoff line. Before the cookie mount lands, rung 14 resolves M2 for the projection too, by an
  origin split or a per-request token, and the bearer mount goes behind `addBearerBypass` only.
- **m2. Ruling 3 conflates an anonymous caller with an under-privileged member.** Its rationale
  (the member-visible catalog already discloses the operations) holds only for an authenticated
  member. For an anonymous caller, a 403 against the shell's 404 discloses the mount. Fold: extend
  Ruling 3's question. The recommendation stays yes for a verified member lacking the privilege,
  and the authorizer's "no identity" denial answers what `http.NotFound` answers, as rung 5's gate
  does. This needs two denial kinds in the authorizer contract now, because rung 14 inherits the
  shape.
- **m3. Body refusals answer before the authorizer.** R-6bii4 decides Content-Type, size, JSON
  shape, and handle refusals "before any authorizer call," so an anonymous caller gets distinct
  400s per operation. The consequence is low, since only operation names leak, but it inverts rung
  5's gate-first order. Fold: authorize right after the path match, then decode. Otherwise record
  the order as deliberate.
- **m4. "With its reason" should name the field.** The body must carry `(*ops.Error).Reason` only,
  never `err.Error()`, because a wrapper around a coded error may add a path. Fold: pin it in
  R-6bii4, and add a Task 2 mutation that encodes `err.Error()` for a coded error wrapped by a
  path-naming `fmt.Errorf`.
- **m5. Ruling 2's wrap keys on the error chain, not on `ctx.Err()`.** A wait interrupted inside
  SQLite's busy handler may surface as a driver interrupt error without `context.Canceled` in its
  chain, and the plan quotes no modernc behavior for it. The contract case would then classify as
  `internal`. Fold: wrap when the caller's `ctx.Err()` is non-nil at return, or quote the driver
  per the planning-miss rule.
- **m6. The stage is not ledger-checked.** `verifyCopy` checks integrity, version, and marker
  (`snapshot.go:113-151`) but not the ledger that `classifyOpen` enforces (`checkLedger`). A stage
  whose ledger disagrees (a divergent build) swaps in, and only then refuses at every open. It is
  recoverable through the pre-restore copy, but it passes the "refused before the record" promise.
  Fold: run `checkLedger` on the stage for versions at or below the head, and add a wrong-archive
  row.
- **m7. `--from-file` source checks are thin.** Also refuse a `-journal` sibling (a hot rollback
  journal gives the same torn copy D4 guards against), and decide "is the live database" by
  `os.SameFile` after `filepath.EvalSymlinks`, not by path string.
- **m8. The record's names contradict each other.** R-6bii7 says the names are leaf-relative
  (`snapshots/…`, `staging/…`) and also refuses "a separator." Fold: store only the stamp and derive
  both names, or store basenames whose directory each field fixes.
- **m9. Runbook line missing: the next backup deletes the pre-restore copy.** Its mtime is the old
  database's last write, so `pruneLocal` removes it on the next successful run. The undo window for a
  mistaken restore therefore closes at the next nightly backup. Add the line to Task 8's runbook
  list.
- **m10. Ruling 5 with `--from-file` keeps a reference to the state being rolled back from.** If the
  database is lost later, `besideCopy` names `restore --snapshot <newer stamp>`, which silently
  brings back the rejected state. Add the case to Ruling 5's text for Geoff. Separately, and cheaply:
  warn when `DUBPLATE_BACKUP_REMOTE` differs from the remote the reference names. That is the one
  wrong-archive signal available before 6d's store identity.
- **m11. The report's mode is not stated.** It lists every library path. Fold: write it through
  `atomicfile.WriteStream` at 0600, like the record.
- **m12. How restore gets a handle under its own exclusive lock is unspecified.** The step 7 query
  and Ruling 5's `WriteBackupRef` (a `Handle` method, `snapshot.go:365`) need a handle while the
  session holds `LOCK_EX`. A second `flock` from the same process on a new descriptor conflicts
  with that lock. Fold: the session hands out a handle that takes no lock. Also make the reference
  rewrite idempotent on every resume path from `swap` on, since the record has no state for it.
- **m13. Crash-table and ruling-text gaps.** Task 5's crash table has no "after step 3" row: the
  current store has been switched to DELETE mode and no record exists. Add it. Also fix Ruling 4's
  "reuses backup's 75" wording, as noted in the citation check above.

---

## Pass sizing (OWNER FORK)

The arithmetic checks: 0.11 + 0.30 + 2.91 + 0.75 + 0.40 + 1.77 = 6.24M; plus 0.48M gives 6.72M; and
6.72M / 0.8 gives 8.4M. The split figures (3.6M + 6.0M, about 1.2M more) also check. The growth
from the spec's 5 tasks and 4.4M is real and explained: two deliverables at four or fewer
deliverables per task. Still, 8.4M is 1.9 times the spec's figure, and the plan counts three task
splits beyond the spec (4A and 4B, 5 and 6, 2 and 3). CLAUDE.md's sizing rule treats a second task
split as the prompt to split the pass.

From the risk lens, the two halves share no file. Each half also has its own review class: the
security read for the projection, and the targeted data-loss read for restore. As drafted, the
data-destructive half runs last. The 80% flag and the conductor's heaviest context both land during
Tasks 5 to 7. The folds above add rows to Tasks 4A, 4B, 5, and 6. The 0.40M reserve for restore fix
rounds is also thin next to 6b-i, whose verb segment alone took two fix rounds.

**Recommendation: Ruling 1 "no" (split).** Run 6b-ii-a (projection) and then 6b-ii-b (restore)
with its own ceiling. Re-derive 6b-ii-b after this review's folds, likely above 6.0M. Its close
then carries the targeted data-loss read alone. The extra cost of about 1.2M buys a fresh
conductor context for the one path that can lose the group's library.
