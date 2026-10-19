# Draft HISTORY entry: rung 6, pass 6b-ii-b (the offline `restore` verb)

Paste-ready for the conductor's ledger fold, once the merge ritual and the
architecture reads land. The bracketed fields need the merge's own commit
ids; everything else is final. Model: `draft-HISTORY-6b-ii-a.md` and the
top entries of `docs/HISTORY.md` (passes 5c-ii and 6b-i) are the format
this draft follows. Scope: Task 0 and Tasks 4A-7 only. The projection-side
outcomes (R-6bii1 to R-6bii5, the contract suite) are 6b-ii-a's, already
folded (`docs/HISTORY.md`'s 6b-ii-a entry).

---

## Rung 6, pass 6b-ii-b (the offline `restore` verb): closed 2026-09-25, lane 2

Executed against
`docs/superpowers/plans/2026-09-25-rung-6b-ii-restore-http.md` (Ruling 1
split the pass into 6b-ii-a, folded earlier at `6e6fd12`, and 6b-ii-b,
this entry) in the lane 2 worktree
(`.claude/worktrees/lane-2-6b-ii-b`, branch `lane-2-6b-ii-b`).

Segment B1 ran as run `wf_a047da45-f4b` (the previous session): 4A
`685b80f`, 4B `ced021e` then `c3dc967`, 4C `044a433` then `0ee7aba`.
Segment B2's first attempt (`wf_a360199a-016`) was killed mid-Task-5A by a
power loss with nothing committed; it relaunched fresh as
`wf_ed08444b-481`: 5A `30a66ac`, 5B `cd9afbf`, Task 6 `9d9c552`, zero fix
rounds, each accepted by its `diff-reviewer` on the first pass. Segment B3
ran Task 7 (`9a7532b`, run `wf_6b3f2248-285`, accepted with zero fix
rounds; SIGINT/SIGTERM/SIGHUP proven to exit 130 and a held lock to exit
75, both with real subprocesses; `ErrLockBusy` exported with the verb as
its first caller), then the `code-simplifier` round (`7a56699`, four leaf
free functions made methods on `leaf`, no behavior change), then the
targeted data-loss read (`23ee61f`, `opus`, fresh context, `0` blockers,
`2` majors, `5` minors), then two fix rounds:

- **Fix round `cba822c`**: closed M1 (a proven false premise in spec
  criterion 16 — the plan's privileged-fields row for a pre-restore copy
  said 6b-i's local prune deletes it; a pre-restore copy's rows are the
  ones the restore rolled away, so no later backup ever captures them,
  and deleting it closes the operator's only undo window; conductor
  decision: kept like a `discarded-*` file until removed by hand, with
  `RestoreSession.PreRestorePath` added as the one caller `restore`'s
  output reads to print the kept path), M2 and m3 (a symlinked directory
  or file under the library root is refused before any displacing copy,
  on both a first run and a resume), m1 (the fetched stage and
  `staging/` are fsynced before the swap can use them), and m4 (a
  relative library root is refused before the fetch and the quiesce).
  m2 (a library root that is a mount point puts the displaced directory
  on the parent filesystem, uncounted by the free-space check) was not
  fixed; it carries to the runbook below.
- **Follow-up `c18fbeb`** (conductor decision on m5): every resumed
  library copy now passes `--backup-dir`, not only a first run under
  `--into-nonempty`, since `rclone copy` skips a destination file
  identical to the source and `--backup-dir` moves only a file the copy
  would overwrite (rclone v1.74.3's `rclone help flags`), so a file an
  operator places in the root between an interrupted run and its resume
  is displaced rather than overwritten. Its `diff-reviewer` returned
  `fix`: the symlink walk descended into a root-owned `lost+found`.
- **Second fix round `876f8ea`**: `lost+found` is now skipped by name
  (the one directory `mklost+found(8)` says every fresh filesystem
  makes, root-owned at 0700); any other unreadable directory is refused
  by name instead of silently walked past; the symlink walk re-runs
  immediately before every displacing copy, so a link made after the
  first check is still caught. Test follow-up `960e760` covers
  `lost+found` itself replaced by a symlink on both the first-run and
  resume paths, and guards the `unreadable()` test helper against
  running as root, where a mode-0 directory reads clean and the
  assertions it backs would be vacuous. This fix round's `diff-reviewer`
  verified the `SkipDir` the walk returns for a refused entry cannot be
  tricked into skipping a sibling it should still check.

### The merge

6b-ii-b merges to master at `<merge-sha, filled at the fold>`: `<rebase
outcome>`, a green worktree gate, `git merge --ff-only`, a green gate on
master, `<merged-range diff-reviewer verdict, if the rebase is not a
fast-forward>`.

### The architecture reads

Three `go-architecture-reader` grades run at the merge, each with no plan
in context: `internal/engine/store`, `internal/backup`,
`cmd/dubplate-server`. `<findings and routing, filled at the fold>`.

### The gate evidence

At HEAD (`960e760`): `go test -count=1 ./internal/engine/store/...
./internal/backup/... ./cmd/dubplate-server/...`: `internal/engine/store`
**11.487s**, `internal/backup` **52.529s**, `cmd/dubplate-server`
**56.083s**. `go test ./... -count=1` at HEAD: **2m28.930s** wall
(`internal/harness/conform` at 145.916s dominates, unchanged from prior
passes; `cmd/dubplate-server` at 55.541s is next largest, and
`internal/backup` (which launches the real `rclone` binary) at 26.918s;
every other touched package finishes under 10s).

`cairn-run-gate 'bash scripts/check.sh'` at HEAD: **exit 0**, 494-line
log. gitleaks scanned about 114.95 MB in 51.5s and found no leaks.

Two staged legs stayed at warning: **326** process citations in Go
comments (unchanged from 6b-ii-a's close; this half added no new one)
and **126** exported-unused `internal/` identifiers (up one from
6b-ii-a's 125: `PreRestorePath`'s own package caller is `internal/backup`
non-test code, so it does not add to the count; the rise is a symbol this
half's own tests do not yet exercise from outside its package — named at
the merge's architecture read, not enumerated here).

### The step 3 (quiesce) measurements

- Switching the current store's journal mode to `DELETE` on the only
  connection removes `-wal` and `-shm` at that connection's close
  (`internal/engine/store/swap.go:53-55`, verified by `quiesce`'s own
  test).
- An idle foreign connection (a reader holding an old WAL snapshot, or
  another process's handle) fails the switch at once: `PRAGMA
  wal_checkpoint(TRUNCATE)` reports a nonzero busy column, or the mode
  switch itself answers `SQLITE_BUSY`, either refused before any file is
  touched (`swap.go:75-92`).

### The mode `copyto` writes, and the combination table

`--from-file` and the fetch both pre-create no stage file; `copyto`
writes it directly and the store's `VerifyStage` runs before any record.
The swap's combination rule (R-6bii8) keys the resume solely on which of
the record's named files exist in `db/` and `snapshots/`, never on the
record's own state string, so a crash between either rename is finished
from the filesystem alone. The finished state for `--discard-current`
needs only a discarded file beside the incoming store, not a discarded
database specifically (`0ee7aba`'s fix): a discard record for a `db/`
holding only a `-wal` or `-shm` sidecar (the marked database itself
already renamed away by an earlier resume) must not refuse a finished
swap.

### The under-lock record re-read, and the `canceled` meaning

Write mode and serve's shared path each re-read the intent record after
winning the exclusive or shared lock, so an opener that waited through a
restore sees the record the restore left rather than a stale in-memory
read from before it waited. `context.Canceled` reaching the transport
(a SIGINT the operator sent, or the command's own context canceled)
exits 130, distinct from every other failure's exit 1, per the source
table R-6bii9 and R-6bii12 give.

### The exit gate items this half owns

Per the plan's "The exit gate" (Ruling 1's no: 6b-ii-b owns items 2 to 8,
11, 12, and 13, plus 9 and 14 over its own tasks and exports):

- **Item 11** (`internal/engine/store` imports no `net/http`): confirmed
  by `go list -f '{{join .Imports "\n"}}' ./internal/engine/store/...`
  filtered for `net/http`, which finds nothing.
- **Item 12** (nothing creates or migrates a store outside a test, except
  restore creating a leaf and placing a verified snapshot): `OpenRestore`
  (`internal/engine/store/restore.go:111`) is the one production path;
  serve's own store creation is 6a's, unchanged this pass.
- **Item 13** (no HTTP route mounted): this half touches no HTTP surface;
  `git grep -n opshttp -- ':!internal/ops/opshttp' ':!docs'` finds
  nothing, unchanged from 6b-ii-a's close.
- **Item 9** (state rows each have a named test): R-6bii7's two tables,
  R-6bii8's combination table, R-6bii9's source table, and R-6bii11's
  report table are each asserted by Tasks 4A-4C's, 5A-5B's, and 6's own
  state tables (`swap_test.go`, `restore_test.go`, `intent_test.go`).
- **Item 14** (every exported symbol lists its production caller):
  `store.OpenRestore` called from `backup.Restore`
  (`internal/backup/restore.go:168`); `backup.Restore` called from
  `cmd/dubplate-server/restore.go:176`; `store.ErrLockBusy` called from
  `cmd/dubplate-server/restore.go:196` and matched in
  `internal/engine/store/inspect.go:200`; `RestoreSession.PreRestorePath`
  called from `internal/backup/restore.go:694`. `store.OpenServe`'s
  `Reconcile` parameter has no production caller yet — 6d's named
  exception, unchanged.

### R-6bii6 to R-6bii12, where each landed

| Resolution | What it decides | Landed |
| --- | --- | --- |
| R-6bii6 | Restore's package split, the session, and its staging rule | `internal/engine/store/restore.go` (session, staging, close-on-exec handles), Tasks 4A-4C (`685b80f`, `ced021e`, `c3dc967`, `044a433`, `0ee7aba`); the fetch/transport half in `internal/backup/{fetch,rclone}.go`, Task 5A (`30a66ac`) |
| R-6bii7 | The intent record's format, its states, and when an open reads it | `internal/engine/store/intent.go`, Task 4A (`685b80f`); write mode's and serve's re-read after the lock, Task 4A; the two-table resume rule exercised through Tasks 5B and 6 |
| R-6bii8 | The quiesce, the swap, the discard, and what is quoted or measured | `internal/engine/store/swap.go`, Task 4C (`044a433`, `0ee7aba`); step 3's measurements above |
| R-6bii9 | The sources a restore reads, and every wrong-archive state | `internal/backup/restore.go` (`--from-file` and `--snapshot` staging, the wrong-archive refusals), Task 4B (`ced021e`, `c3dc967`) and Task 5A (`30a66ac`) |
| R-6bii10 | The free-space check | `internal/backup/restore.go`, Task 5A (`30a66ac`); the mount-point line, Ruling 7 |
| R-6bii11 | The library half and the reconciliation report | `internal/backup/restore.go` (the copy, the displaced directory, `checkRoot`/`refuseSymlinks`) and `internal/backup/reconcile.go`, Task 6 (`9d9c552`), amended by the fix rounds (`cba822c`, `c18fbeb`, `876f8ea`, `960e760`) |
| R-6bii12 | The verb, its signals, and its output | `cmd/dubplate-server/restore.go`, Task 7 (`9a7532b`) |

### Mutations run, per task

- **Task 7** (the verb): SIGINT, SIGTERM, and SIGHUP each proven to exit
  130 through a real subprocess; a held store lock proven to exit 75
  through a real second holder; zero fix rounds, accepted by its
  `diff-reviewer` on the first pass.
- **The second fix round's `diff-reviewer`** verified the symlink walk's
  `SkipDir` return cannot be tricked into skipping a sibling entry it
  should still check.
- **Tasks 4A, 4B, 4C, 5A, 5B, and 6's own named-mutation counts and
  which test each fired on were not carried forward into this Task 8's
  dispatch** (the conductor's brief for this task states only Task 7's).
  This is a declared gap in this draft, not a claim that no mutations
  ran: each task's `diff-reviewer` accepted its own dispatch (Task 4C
  after one re-executed child at `0ee7aba`, 4B after one re-executed
  child at `c3dc967`, both fixes named in their own commit messages
  above; 5A, 5B, and 6 each accepted with zero fix rounds). The
  merge-time architecture reads and this repo's `git log` are the record
  a later pass should consult for the per-task detail this draft could
  not paste.

### The pass-end review

The spec's one targeted fresh-context `opus` read of the data-loss paths
(`23ee61f`) found 0 blockers, 2 majors, 5 minors; see above for M1, M2,
m1, m3, and m4 (fixed) and m2 and m5 (m5 fixed at the follow-up, m2
carried to the runbook). Its "Verified sound" section held up the swap,
the intent record, and every concurrency window it traced without a
finding.

### Both budgets

Tokens: `<the pass's actual, filled at the fold>` against the **7.53M
execution ceiling** (Ruling 1's split; 6b-ii-b's half) with the 80% flag
at **6.02M**. Segment B1 ran as `wf_a047da45-f4b` in the prior session.
Segment B2's first attempt (`wf_a360199a-016`) was lost to a power cut
mid-Task-5A with nothing committed and no tokens landed; it relaunched
fresh as `wf_ed08444b-481`. Segment B3 ran as `wf_6b3f2248-285`. `<the
targeted read's spend, the two fix rounds', the simplifier round's, and
this task's, filled at the fold>`.

Attended time: **zero attended interactions after approval**; the
power-loss relaunch needed no owner decision, since the last committed
state (Task 4C's `0ee7aba`) was intact and the killed run had committed
nothing. Fix rounds: Task 4B one
(`c3dc967`), Task 4C one (`0ee7aba`), Task 5A none, Task 5B none, Task 6
none, Task 7 none; two pass-end fix rounds (`cba822c`, `876f8ea`) plus
one conductor-decision follow-up (`c18fbeb`) after the targeted
data-loss read.

---

## The errata this pass owes

- **Spec criterion 16** (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md:1018`,
  "every local pre-migration and pre-restore snapshot older than its
  snapshot is deleted") and **line 712** ("A pre-restore copy follows
  ruling 7's retention") are both corrected by M1's fix: a pre-restore
  copy is never auto-pruned. Criterion 16 holds for pre-migration
  snapshots only; a pre-restore copy is kept, like a `discarded-*` file,
  until an operator removes it by hand.
- **The plan's privileged-fields row for a pre-restore copy**
  (`docs/superpowers/plans/2026-09-25-rung-6b-ii-restore-http.md:459`,
  "deletion by 6b-i's post-backup local prune and spec ruling 7") is
  corrected the same way: `isLocalCopyName` no longer matches the
  pre-restore prefix (`cba822c`).
- **R-6bii11's "The displaced directory" paragraph**
  (plan `:936`, "applies only with `--into-nonempty`... a resumed copy
  has no operator file to displace") and **Task 6's summary**
  (plan `:1711`, "with Ruling 6's `--backup-dir` on yes, with
  `--into-nonempty` only") are both superseded by the m5 follow-up
  (`c18fbeb`): `--backup-dir` applies on a first run with
  `--into-nonempty` and on every resume, since a resume's root can hold
  a file an operator placed between runs.
- Deviations D1-D5 are 6b-ii-a's, unchanged here. D6-D10 (this half's
  surface) landed as the plan states, with the two corrections above.

---

## The 6b-ii-b to 6c and 6d handoff

**To 6d** (serve's wiring):
- `ServeOptions.Reconcile` and `backup.Reconcile` are `serve`'s own
  wiring: `internal/engine/store.OpenServe` takes a `Reconcile` function
  parameter with no production caller yet (the named exception,
  unchanged from the plan).
- The store identity that would let restore refuse another store's
  snapshot: still unminted; a snapshot from another store's history is a
  declared 6d gap (the targeted read's "Wrong archives" section).
- The network actor's `member:<principal>` form against
  `Identity.UserID`: restore's own verb needs no actor (it runs offline,
  under the operator's own uid), so this stays a 6d question for the
  wiring, not something this half resolved.

**To rung 14** (the mount): unchanged from 6b-ii-a's handoff; this half
touches no HTTP surface.

**To rung 12**: the streamer database's place in a restore, unchanged
(0.8 forward-hook reference; streamer versions re-keyed 2026-09-25).

---

## The runbook lines owed to rung 16

- Removing a kept pre-restore copy by hand: no backup holds it (M1's
  fix), so the operator confirms the restored state is correct before
  deleting `snapshots/pre-restore-<stamp>.db`.
- A library root that is a mount point puts the displaced directory
  (`<root>.restore-displaced-<stamp>`) on the parent filesystem,
  uncounted by the free-space check (data-loss read m2, carried
  unfixed): mount the volume with headroom on the parent, or restore
  into an empty root sized for the whole displacement.
- An unreadable directory under the library root (other than
  `lost+found`) refuses a displacing copy: make it readable, or move it
  out of the root, before `--into-nonempty` or a resume.
- An empty 0700 displaced directory left by a failed resumed copy is
  harmless and is removed automatically after a successful copy leaves
  it empty; an operator does not need to remove it by hand unless a
  restore is abandoned entirely.
- Restore needs the server stopped.
- `--older` restores the store only.
- A restore before rung 12 loses stars and playlists.
- The rotation step (Ruling 12's instruction, printed at the verb's
  last step).
- The reconciliation report file's location, in the leaf.
- `--discard-current`'s `discarded-<stamp>.db` files, with their `-wal`
  and `-shm`, which nothing prunes.
- Mount the library volume before a library restore (Ruling 7).
- An interrupted restore is finished with `restore --resume`, or by
  starting serve when the verb said so.
- On Ruling 6's yes, the displaced directory beside the root: nothing
  removes it automatically while it holds files an operator must
  reconcile by hand.
- A real `rclone` killed by `SIGKILL` during the library copy leaves
  `<name>.<hex>.partial` files under the root, which the report lists as
  files without a row; the operator removes them by hand.
- On Ruling 9's yes, the manual procedure for a record resume that
  refuses: every refusal already names each file the record expects and
  whether it exists; the procedure is to make the named file's state
  match what the record expects (restore it, remove it, or repair its
  permissions), then retry `restore --resume`. There is no
  `restore --abandon` flag.

---

## Carried non-blocking items (not fixed this pass)

From the in-chain reviews across Tasks 4A-7 and the two fix rounds:

- The reconcile walk descends into `lost+found` in the reconciliation
  report path (distinct from the symlink-refusal walk, which now skips
  it) and repeats the raw path in its walk-error lines.
- `Reconcile` never checks the root is a directory before it walks.
- `makeDisplaced` creates its directory at step 6, not earlier, so a
  failure there is a step-6 failure rather than a pre-flight one.
- `sayKept` (the pre-restore-copy message) prints before `Complete`,
  so a crash between the two would print a message about a copy the
  step after it has not yet finished recording.
- The first-run `--into-nonempty` residual window: a symlink created
  between the late `checkRoot`/`refuseSymlinks` walk and rclone's own
  traversal is not caught by this pass's re-walk-before-copy fix if the
  race lands inside rclone's own directory read, not before it.
- `fetch_test.go`'s stand-in environment names are duplicated from
  `standin_test.go`'s (different compiled packages; not a collision,
  but a maintenance note).
- `cmd/dubplate-server/backup.go`'s second remote-name validator is
  untouched by this pass.

---

## What a later pass would be wrong to rediscover

- **Spec criterion 16 and the plan's own privileged-fields row were both
  wrong about pre-restore-copy retention.** A pre-migration copy's rows
  are carried forward by the migrated store and the next backup captures
  them; a pre-restore copy's rows are by construction the ones the
  restore did not bring back, so no backup ever captures them, and
  auto-pruning it on the same schedule as a pre-migration copy destroys
  the operator's only undo window. `isLocalCopyName` no longer matches
  the pre-restore prefix; a pre-restore copy is retained like a
  `discarded-*` file.
- **`rclone copy` does not follow symlinks and silently treats a
  symlinked destination path as absent** (a scratch run, rclone
  v1.74.3, local backend, cited in the data-loss read): the destination
  write then goes straight through the link, replacing whatever it
  points at with no entry in the backup directory and no row in the
  reconciliation report to name it. `checkRoot`/`refuseSymlinks` refuses
  every symlink under the root before any displacing copy, both on a
  first run and immediately before every resume's copy, because a link
  can appear in the gap between the two.
- **`--backup-dir` on a resume, not only a first run under
  `--into-nonempty`, is needed to protect a file an operator places in
  the root between an interrupted run and its resume**: deviation D9
  (resume skips the emptiness check) means the root can be non-empty at
  resume time for a reason other than the restore's own earlier copy.
- **`lost+found` is root-owned at 0700 on a fresh ext volume**
  (`mklost+found(8)`), which the emptiness check already exempts by
  name (Ruling 8) but which an unqualified symlink-refusal walk would
  refuse to even read as the library's non-root owner. The walk skips
  `lost+found` by name rather than trying to make it readable.
- **Segment B2's first attempt was lost to a power cut with nothing
  committed**, mid-Task-5A. The relaunch (`wf_ed08444b-481`) started
  fresh from the last committed state (Task 4C's `0ee7aba`); no partial
  work from the killed run survived or needed reconciling, because
  nothing from it had reached a commit.
