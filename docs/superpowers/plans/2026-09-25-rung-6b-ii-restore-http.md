# Rung 6, Pass 6b-ii (The Projection and Restore) Implementation Plan

Status: **reviewed; awaiting owner approval (Rulings 1-9).** Revision 3, 2026-09-25. Three lenses
reviewed revision 1 (`docs/superpowers/research/2026-09-25-rung-6b-ii-plan-review-contract.md`,
`-mechanics.md`, and `-risk.md`), and an independent read verified revision 2's fold
(`docs/superpowers/research/2026-09-25-rung-6b-ii-plan-fold-verification.md`). The fold record
(`docs/superpowers/research/2026-09-25-rung-6b-ii-plan-fold.md`) carries one disposition per
finding of both rounds; its "Fold 2" section covers the verification read. Fold 2 splits Task 5
into 5A and 5B. Nine rulings for Geoff are open ("Rulings awaiting owner"). Ruling 1 recommends
splitting the pass, and both answers are executable as written.

Sources: the ratified rung 6 spec (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`,
revision 2.2 with errata E1 to E11); its fold record
(`docs/superpowers/research/2026-09-23-rung-6-spec-fold-record.md`); the approved 6b-i plan
(`docs/superpowers/plans/2026-09-24-rung-6b-i-ops-seam-backup.md`, revision 2.2) and 6b-i's close
drafts (`docs/superpowers/research/handoffs/draft-HISTORY-6b-i.md` and `draft-STATUS-6b-i.md`, in
the lane 2 worktree); the approved 6a plan's R-6a7 and its handoff
(`docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md`, `docs/HISTORY.md`'s 6a entry);
the 2.0 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, "Forward hooks"
and "Prior art"); the configuration homes spec
(`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`); and rung 5's admin gate as
shipped on master (`internal/auth/admin.go`, `internal/identity/identity.go`,
`internal/serve/origin.go`).

**Code pointers are resolved against two trees, and Task 0 re-resolves both.** Every pointer into
`internal/ops`, `internal/ops/local`, `internal/backup`, `internal/engine/store`,
`internal/atomicfile`, and `cmd/dubplate-server/backup.go` is read in the lane 2 worktree
(`.claude/worktrees/lane-2-6b-i`, branch `lane-2-6b-i`) at `33ceff5`, because 6b-i is finished but
not yet merged and that tree is what merges. The fold re-read the pointers it relies on at the
worktree's `6f2c3e0` (6b-i's close drafts on top of `33ceff5`). Every pointer into `internal/auth`,
`internal/identity`, and `internal/serve` is read on master at `f0eec0d`. 5c-ii is committing in
the main tree while this plan is drafted, so those master pointers are the likeliest to move.
`cmd/dubplate-server/main.go` changes at 6b-i's merge, since both lanes add a registration line, so
this plan cites it by symbol (`newProductionRegistry`, `newRootCmd`'s `AddCommand` block, `main`'s
`exitCodeErr` branch), never by line.

**What this pass is.** Pass 6b-ii builds the two deliverables the spec's pass table assigns it
(spec `:152-153`). The first is `internal/ops/opshttp`, the unexported HTTP projection of the ops
seam with its authorizer and the contract suite (spec criteria 1 to 4, `:1037-1043`). The second
is the offline `restore` verb: the restore-intent record's format, the one-file swap, library
restore, and the reconciliation report (spec criteria 5 to 13, `:1044-1060`). It also declares the
two sentinels and the network actor 6b-i deferred to their first caller (6b-i's R-6bi4 and
R-6bi5), and it teaches serve mode the one intent state it accepts, `reconcile-pending` (spec
`:302-303`, R-6a7).

**What this pass is not.** No HTTP route is mounted anywhere, and no `ops/client` is written: the
projection stays unexported until rung 14 owns the authorization model and mounts it (spec
`:164-166`). No production authorizer exists at 6b-ii; the contract suite's authorizers are test
implementations. No settings table, no seed rule, and no re-homed constant: those are 6c's. No
`serve` wiring, no roster import, and no store identity: those are 6d's. Nothing in production
creates or migrates a store in this pass, and `restore` can create a store only from a snapshot,
which cannot exist on a box before 6d (spec `:186-189`). No `systemd/`, `scripts/`, `env/`, `web/`,
or `internal/config/keys.go` change, and **no deploy of any kind.** A task that believes it needs
any of these stops and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the pass-execute chain per
> task, sequential, in the segments "Execution model" names. The implementer is
> `dubplate-implementer`. The conductor consumes structured reports and never reads a diff, a test
> log, or a gate transcript inline. Subagents never dispatch agents, so `code-simplifier` is the
> conductor's dispatch. The conductor owns `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the
> lane registry, and the contested-surface rows. **This pass runs in lane 2, a fresh git worktree
> cut from master after 6b-i merges, beside lane 1 in the main tree.** The live-executor check is
> the lane registry, never `pgrep -f Projects/dubplate`.

---

## Rulings awaiting owner

Nine yes-or-no questions. The spec's fourteen rulings, 6a's four, and 6b-i's four stand and are not
reopened. Approving this plan also approves the spec deviations listed after the rulings.

**1. Does 6b-ii run as one pass under a 10.30M execution ceiling?** Recommendation: **no, split
it.** Reason: the spec priced 6b-ii at five tasks and 4.4M (spec `:1139`). Revision 1 needed eight
build tasks to keep each at four deliverables or fewer, fold 1 added a ninth (Task 4C) and rows to
Tasks 3, 4A, 5, and 6, and fold 2 splits Task 5 into 5A and 5B. That is five task splits against
the spec's count (2 and 3; 4A, 4B, and 4C; 5A and 5B; 5B and 6). The workstation's sizing rule
makes a third task split the point where a pass split is overdue. The two halves share no file, and each has its own review class: the security
read for the projection, the targeted data-loss read for restore. As one pass, the
data-destructive half runs last, under the conductor's heaviest context, where the 80% flag lands.
Revision 1 recommended yes; the risk lens's split recommendation and the fold's growth change it.
What each answer builds:

- **Yes:** one pass, one close, ceiling 10.30M, flag 8.24M, the four segments of "Execution
  model."
- **No:** **6b-ii-a** (Task 0, Tasks 1 to 3, and a close; ceiling 3.90M, flag 3.12M), then
  **6b-ii-b** (a fresh Task 0, Tasks 4A to 7, and a close; ceiling 7.53M, flag 6.02M). The
  second Task 0, close, and merge cost about 1.1M more in total. "Token ceiling" carries both
  derivations, and "Execution model" names each half's segments and close.

**2. Does the `canceled` wire code mean that the caller's own context was canceled before the seam
answered?** Recommendation: yes. Reason: 6b-i declared `ErrCanceled` and gave it a wire code
(`internal/ops/errors.go:18`, `:29`), but no production path raises it (a `git grep` of
`NewCanceledError` finds only its declaration and its test). A job that ends canceled is a status
value, never a failed operation: `CancelJob` on a running job answers its state
(`internal/ops/local/local.go:228-246`). The spec's contract case, "canceled (canceling a running
job)" (`:603`), has no operation that returns the sentinel today. What each answer builds: yes
classifies a seam method's error as `canceled` when the caller's `ctx.Err()` is
`context.Canceled` at the method's return (R-6bii5). The contract case becomes "a request whose
context is canceled while it waits for the write pool's one connection in this process." A wait on
another process's write lock also classifies as `canceled`, but only after the full busy timeout,
since SQLite's busy handler ignores the context (`internal/engine/store/write.go:51-53`; measured
by the mechanics lens at 5.013 s for a 5 s timeout). No drops `canceled` from the vocabulary as a
spec deviation, leaving six sentinels, and the contract suite's per-sentinel rule counts six.

**3. Does the projection answer a verified member who lacks an operation's privilege with HTTP 403
and the code `forbidden`, where rung 5's admin gate answers a non-admin with the same 404 as an
unregistered path?** Recommendation: yes. Reason: the catalog operation, a member privilege, already
lists every operation and its privilege to any member (`internal/ops/local/local.go:118-135`), so a
403 tells a member nothing the catalog did not. The admin gate's 404 hides a route that exists
(`internal/auth/admin.go:98-112`), and hiding it here would contradict the catalog. The argument
holds only for a verified member: a caller with no identity never saw the catalog. That caller's
answer (401 or 404) is rung 14's, where the mount and its session middleware land; `opshttp` is
unexported, so rung 14 can add a second denial kind at no cost to any caller. Task 8's handoff
carries it. What each answer builds: yes answers 403 with the JSON error body, and spec criterion
2's "returns forbidden over HTTP" holds literally. No answers exactly what `http.NotFound` answers
for a denial, as the admin gate does, and the contract suite's forbidden case asserts the 404 and
the absence of a job record instead.

**4. Does the `restore` verb exit with the codes below?** 0 restored (or resumed to completion), 1
refused or failed, 75 the store lock is held by another process, and 130 interrupted by a signal,
with the intent record left for `restore --resume`. Recommendation: yes. Reason: 1 and 130 carry
`backup`'s meanings (6b-i Ruling 3, `cmd/dubplate-server/backup.go:67-74`). `backup`'s 75 means a
busy kind lease ("a backup is already running," `backup.go:101`), and a timed-out store-lock wait
there exits 1. Restore takes its lock non-blocking, so its 75 carries the same "try again later"
meaning for its own lock, and an operator reads one table for both verbs. What each answer builds:
yes fixes the four codes in the verb's help and adds no new branch to `main`, which already honors
`exitCodeErr`. No exits 1 on every failure, and the help text says to check for a restore-intent
record after an interrupt.

**5. After a restore, does the backup reference follow the restored state?** With
`--snapshot <stamp>`, the reference is rewritten to name that stamp and the remote. With
`--from-file`, the existing reference is left unchanged. Recommendation: yes. Reason: the
reference is the leaf's memory that a store existed and where to get it back (6b-i R-6bi16,
`internal/engine/store/snapshot.go:353-397`). After a snapshot restore, the restored stamp is the
state the operator chose. After a `--from-file` rollback, the remote still holds the newer backup,
and keeping the reference keeps the lost-database guard alive. Two consequences to weigh: with
`--older`, the reference names a stamp older than the remote's newest; after a `--from-file`
rollback, a later lost database makes serve name `restore --snapshot <newer stamp>`, which brings
back the state the operator rolled away from. What each answer builds: yes adds one reference
write to the snapshot path, through the restore session (R-6bii6), and none to the file path. No
leaves the reference unchanged on both paths, so a lost database after a snapshot restore names the
pre-restore backup stamp, which the operator had rejected.

**6. With `--library --into-nonempty`, does restore keep every local file the copy replaces, by
moving it into a sibling directory `<root>.restore-displaced-<stamp>` through `rclone copy
--backup-dir`?** Recommendation: yes. Reason: plain `rclone copy` never deletes ("Doesn't delete
files from the destination," `rclone copy --help`, v1.74.3), but it replaces any destination file
that differs, "testing by size and modification time or MD5SUM," whichever side is newer. A FLAC
retagged or uploaded after the last backup is then overwritten and lost, and the reconciliation
report cannot name it, since the path has both a row and a file. `--backup-dir` is listed in
`rclone help flags`'s Sync group as "Make backups into hierarchy based in DIR"; `rclone copy
--help` does not list it, so the fold probed it (v1.74.3, local backend, fold record "Probes"): a
differing destination file moved into the backup directory, an identical file and a local-only
file stayed untouched, and a backup directory inside the destination was refused with "destination
and parameter to --backup-dir mustn't overlap." What each answer builds: yes, with
`--into-nonempty` only, creates the displaced directory at 0700 before the copy (a creation failure
refuses before any rclone call), passes `--backup-dir`, names the directory and its file count in
the report, and removes it only when it is empty after the copy. A resume accepts the directory
its own interrupted run created (R-6bii11). No runs plain `rclone copy`, as revision 1 drafted. Two other answers exist:
`--ignore-existing` ("Skip all files that exist on destination") keeps every local file and never
restores the backup's version of it, and `--update` ("Skip files that are newer on the
destination") keeps only local files newer than the backup's.

**7. Does restore treat an unmounted mount point as a declared gap, printing before the copy
whether the resolved library root is itself a mount point, with a runbook line owed to rung 16?**
Recommendation: yes. Reason: an unmounted mount point is an existing, usually empty directory, so
neither the absent-root nor the non-empty refusal fires, and `rclone copy` fills the filesystem
underneath. Every refusal rule found either fires on the default layout (an empty root on its
parent's device is also every same-filesystem library's first restore) or misses mounts that
`/etc/fstab` does not list (systemd `.mount` units, cloud-init mounts). What each answer builds:
yes prints the resolved root, whether its device differs from its parent directory's (a mount
point) or not, and the free space the library check read, and strikes revision 1's claim that the
case is refused. No refuses an empty root on its parent's device unless a new flag names it, a
spec deviation that fires on every default layout's first restore.

**8. Does step 6's emptiness check ignore a `lost+found` directory at the library root?**
Recommendation: yes. Reason: "There is normally a lost+found directory in the root directory of
each file system" (`mklost+found(8)`), so a library on a freshly made ext volume reads as
non-empty, and its first restore would need `--into-nonempty`, the flag that disarms the guard.
What each answer builds: yes skips one directory entry named `lost+found` in the check, with a
state row. No counts it, and such a root needs `--into-nonempty`.

**9. Does 6b-ii leave a record that `restore --resume` refuses to a runbook procedure owed to rung
16, rather than a `restore --abandon` verb?** Recommendation: yes. Reason: the fold removes every
reachable unresumable state it found (a fresh-box swap, a cross-device rename, a discard before the
record, a version verdict made by another binary), so what remains is damage from outside the
flow. Every refusal already lists each file the record names and whether it exists. What each answer
builds: yes adds the runbook line to Task 8, and the refusal of an unreadable or unresumable record
names that procedure. No adds `--abandon`, a spec deviation: it removes the record only when `db/`
holds a store that verifies, first renames the named pre-restore copy back when it does not, and
never deletes the staged file. That is a fifth deliverable for Task 7, so it becomes its own task
(7B, `opus`, about 0.5M added to the restore half's ceiling).

### Spec deviations that take effect on approval

These change the ratified spec's meaning. Approving this plan approves them; Task 8 drafts the
erratum text. D2 to D6 and D10 are clarifying or additive; D1, D7, D8, and D9 change what a sentence
of the spec says.

- **D1 (R-6bii3).** The contract suite's in-process leg runs against `ops/local` wrapped in the
  same authorizing layer the projection composes, not against `ops/local` bare. `ops/local`
  enforces no privilege (6b-i R-6bi4), so a bare in-process call can never answer `forbidden`, and
  the spec's "every case passes in-process" (`:1037-1038`) cannot hold for that one sentinel
  otherwise.
- **D2 (R-6bii7).** Serve mode reconciles through a function its caller passes in `ServeOptions`.
  The store cannot import `internal/backup`, which owns the reconciliation queries (spec
  `:167-169`, `:204`), so "serve mode ... writes the report" (`:1055-1056`) is serve mode calling
  that function and writing what it returns.
- **D3 (R-6bii6).** Restore's flow is split across the package boundary. The store owns every
  operation on the leaf's files (the exclusive lock, staging, the checkpoint, the intent record,
  the swap, the reference write, and the report file); `internal/backup` owns the fetch, the
  library copy, the reconciliation, the step hook, and the order of the steps. The spec's package
  table puts "restore" in `internal/backup` (`:204`), and `internal/backup` cannot reach the leaf's
  unexported paths. The session exposes step 5 as two calls, so the flow calls its one step hook
  between the two renames and the spec's three test seams stay three (R-6bii8).
- **D4 (R-6bii9).** A `--from-file` source whose `-wal` or `-journal` sibling exists is refused.
  The spec lists no such state; a byte copy of a database beside a live WAL or a hot rollback
  journal is the torn copy 2.0 hook 10 forbids (spec `:619-620`).
- **D5 (R-6bii7).** The intent record carries a state for every flow step it must resume, not only
  the one state the spec names, and write mode's refusal names the action that state needs:
  `restore --resume` for every state but `reconcile-pending`, which names "start serve".
- **D6 (R-6bii2).** `internal/ops/opshttp` imports `internal/identity` beside `internal/ops`, for
  the mount's channel type. The spec's table (`:203`) names `ops` alone.
- **D7 (Task 5B).** Criterion 7's "a crash ... after each flow step leaves the intent record"
  (`:1047-1048`) holds from step 4 on. The record is written at step 4, so a crash after step 1, 2,
  or 3 leaves none: the current store is untouched or in rollback-journal mode, and a write-mode
  open proceeds and restores WAL through its DSN.
- **D8 (R-6bii8).** `--discard-current` moves the unopenable files aside after the record is
  written, as step 5's first rename, to `snapshots/discarded-<stamp>.db` with its `-wal` and `-shm`
  beside it. Spec step 3 (`:693-697`) moves them before step 4's record, which leaves a crash
  window with no database and no record, in which serve would create an empty store.
- **D9 (R-6bii11).** `restore --resume` in the `library` state skips step 6's emptiness check,
  because the record proves the check passed before the copy began. Spec step 6 (`:702-703`)
  states the refusal without an exception, and without one a resumed partial copy refuses the tree
  the restore itself wrote.
- **D10 (R-6bii6, R-6bii7).** Three refusals the spec does not list: restore refuses before any
  stage when `db/`, `staging/`, and `snapshots/` do not share one device; a new restore refuses
  beside any intent record, naming `restore --resume`; and resume re-checks the staged or restored
  store's version against the running binary before it acts.

---

## Primary inputs

Read these before Task 1, in this order.

1. **The rung 6 spec, revision 2.2.** The sections this pass implements: "What rung 6 builds" and
   "does not build" (`:141-189`), "Package territory" (`:193-221`), "Layout, ownership, and the
   lock" and "Three open modes" (`:227-310`), "The ops seam" (`:533-605`), "The restore verb"
   (`:674-714`), the privileged-field table (`:926-937`), and 6b-ii's thirteen criteria
   (`:1035-1060`). Line numbers are as read at `f0eec0d`; Task 0 re-reads them.
2. **6b-i's handoff** (`draft-HISTORY-6b-i.md`, "The 6b-i to 6b-ii handoff," items 1 to 7) and its
   plan's R-6bi2, R-6bi4, R-6bi5, R-6bi6, R-6bi12, R-6bi14, and R-6bi16.
3. **6a's R-6a7** (the intent record is presence only at 6a, and 6b-ii refines it) and 6a handoff
   items 5 and 6 (`docs/HISTORY.md`, the 6a entry): a `VACUUM INTO` copy is a rollback-journal
   database, and an interrupted file-backed connection is discarded by the driver.
4. **The code as it merges**: `internal/ops/{ops,errors,catalog}.go`,
   `internal/ops/local/{local,jobs,recovery}.go`, `internal/engine/store/{open,leaf,lock,lease,
   snapshot,stamp,conn,inspect,write,migrate}.go`, `internal/backup/{backup,rclone,remote,guard,prune}.go`,
   `internal/atomicfile/atomicfile.go`, and `cmd/dubplate-server/backup.go`, all in the worktree;
   `internal/auth/admin.go`, `internal/identity/identity.go`, and `internal/serve/origin.go` on
   master.
5. **This plan's quoted behaviors.** Each is quoted from `go doc` at `go1.27.1 linux/amd64`, the Go
   toolchain's source (`$(go env GOROOT)/src/net/http/server.go`), rclone v1.74.3's installed help
   (`rclone <command> --help`, `rclone help flags`), sqlite.org (read 2026-09-25), or a probe
   the mechanics lens or the fold ran at modernc.org/sqlite v1.59.0 (SQLite 3.53.4) and rclone
   v1.74.3, cited by its review or by the fold record's "Probes." A task that relies on a behavior
   this plan does not quote confirms it by a gate-resident test and reports what it read.

---

## Goal

After this pass, a test can drive every ops operation over HTTP through a projection that consults
an authorizer before it dispatches, and an operator can bring a store back from a remote snapshot
or a local copy with `dubplate-server restore`. A restore that dies at or after step 4 leaves a
record that every writing open honors, including an open already waiting on the lock when the
record appears, and that `restore --resume` finishes. One that dies earlier leaves the current
store usable (deviation D7). A restore of an older snapshot hands its last step to serve mode,
which migrates, reconciles, and clears the record before any other writer proceeds.

---

## Owner touchpoints

The nine rulings above, at the plan-approval gate. Nothing else is expected. What would become
one: a new module dependency (this pass adds none); a change to what the box runs; a restore step
whose measured SQLite or rclone behavior contradicts this plan's quotes or probes. The mechanics
lens settled revision 1's one named candidate: SQLite 3.53.4 removes `-wal` and `-shm` at the
switch to DELETE on the only connection (R-6bii8).

---

## Token ceiling: 10.30M execution as one pass, or 3.90M and 7.53M split (Ruling 1)

The basis is 6b-i's recorded spend (`draft-HISTORY-6b-i.md`, "Both budgets"): about 5.4M through
the simplifier round for nine build tasks and Task 0, so about 0.46M per `opus` task with its fix
rounds, and about 0.75M a task for the verb segment (Tasks 6 and 7 there, two fix rounds). 6b-i's
close is not yet recorded; its 6a-based estimate of 0.92M plus 0.1M per extra architecture read is
used. Planning, this fold included, is recorded against the rung and outside these ceilings, as
6a's and 6b-i's Ruling 1 set.

**Fold 1's growth, priced.** Task 4C is new (`opus` with a re-executed child: 0.51M). The fold's
rows add about 0.05M each to Tasks 3, 4A, 5, and 6 and about 0.03M and 0.04M to Tasks 1 and 2.
The restore fix reserve rises from 0.40M to 0.60M, since 6b-i's verb segment alone took two fix
rounds and the fold added crash rows to the data-destructive path.

**Fold 2's growth, priced.** Task 5 splits into 5A (`opus`, the one rclone launch path and the
fetch: 0.46M + 0.05M) and 5B (`opus` with a re-executed child, the flow, resume, and the
reference: 0.51M + 0.05M), which adds one task's rate, 0.51M. Task 6 gains the displaced
directory's resume rows (0.03M). Task 7's listing-signal row and Task 1's two named edits ride
their tasks' rates.

**One pass (Ruling 1 yes).**

1. Task 0 at `opus`, at 6b-i's actual: **0.11M**.
2. Task 1 (`sonnet`): 0.30M + 0.03M = **0.33M**.
3. Task 2 (`opus`): 0.46M + 0.04M = **0.50M**.
4. Task 3 (`opus`, a re-executed child): 0.51M + 0.05M = **0.56M**.
5. Tasks 4A (0.46M + 0.05M), 4B (0.51M), 4C (0.51M), 5A (0.46M + 0.05M), 5B (0.51M + 0.05M),
   and 6 (0.46M + 0.05M + 0.03M): **3.14M**.
6. Task 7, the verb, at the 6b-i verb segment's rate: **0.75M**.
7. Fix rounds beyond that rate: **0.60M**.
8. The close side: the simplifier round with its review (0.35M), Task 8 (0.13M), the merge ritual
   and a possible merged-range review (0.11M), six architecture reads at 0.1M each
   (`internal/ops`, `internal/ops/local`, `internal/ops/opshttp`, `internal/engine/store`,
   `internal/backup`, `cmd/dubplate-server`: 0.60M), the spec's targeted `opus` read of the
   data-loss paths (0.30M, spec `:1166-1168`), and `web-auth-security-reviewer` over the
   projection (0.28M): **1.77M**.
9. **The projection is 7.76M.**

In one line: 7.76M projection + 0.48M (one task) = 8.24M flag target; 8.24M / 0.8 = 10.30M, so a
**10.30M** ceiling with its 80% flag at **8.24M**.

**Split (Ruling 1 no).** Each half re-prices the shared close items at its own share.

- **6b-ii-a, the projection:** Task 0 0.11M, Task 1 0.33M, Task 2 0.50M, Task 3 0.56M, a fix
  reserve of 0.15M, and a close of 0.99M (the simplifier round 0.20M, its Task 8 0.10M, the merge
  0.11M, three architecture reads 0.30M, and the security review 0.28M). Projection 2.64M; flag
  2.64M + 0.48M = **3.12M**; ceiling 3.12M / 0.8 = **3.90M**.
- **6b-ii-b, restore:** a fresh Task 0 0.11M, Tasks 4A to 6 3.14M, Task 7 0.75M, a fix reserve of
  0.45M, and a close of 1.09M (the simplifier round 0.25M, its Task 8 0.13M, the merge 0.11M,
  three architecture reads 0.30M, and the targeted data-loss read 0.30M). Projection 5.54M; flag
  5.54M + 0.48M = **6.02M**; ceiling 6.02M / 0.8 = 7.525M, so **7.53M**.
- Together 11.43M against one pass's 10.30M: about 1.1M more, for the second Task 0, a second
  close, and a second merge. Ruling 9's "no" adds about 0.5M to 6b-ii-b (Task 7B).

Checkpoint: STATUS at every segment boundary and at the close. Every task reports its spend, its
mutation outcomes, and its measured gate wall time.

Deliverable count: thirty-eight across Tasks 1 to 8, plus Task 0's report. Tasks 1, 2, 4A, 7, and
8 carry four each (20); Tasks 3, 4B, 4C, 5A, 5B, and 6 carry three each (18). No task passes four.
On Ruling 2's no, Task 1 carries three.

---

## Global constraints

- **Lane 2, a fresh git worktree.** The conductor cuts it from master after 6b-i merges, before
  Task 0, and writes the lane registry row. Every task, gate, and commit runs inside it. After
  lane 1's 5c-ii lands on master, the worktree's first rebase needs `npm ci --prefix web` before
  its in-tree gate (6b-i's global constraints).
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`, and no task opens ssh to
  the box.
- **Package territory, stated as a wall.** This pass creates `internal/ops/opshttp` (every file
  and its tests). It edits `internal/ops` (Task 1), `internal/ops/local` (Task 1: the catalog's
  privilege source and the seam boundary's cancel classification), `internal/engine/store`
  (Tasks 4A, 4B, and 4C), and `internal/backup` (Tasks 5A, 5B, and 6). It creates
  `cmd/dubplate-server/restore.go` and its test, and edits `main.go` (one `AddCommand` entry)
  and `store.go` (the restore-in-progress row's reason, Task 4A). **It writes nothing under
  `internal/auth`, `internal/identity`, `internal/serve`, `internal/config`,
  `internal/atomicfile`, `internal/notify`, `internal/harness`, the rest of `internal/engine`,
  `web/`, `systemd/`, `scripts/`, `env/`, `config/`, `Makefile`, `CLAUDE.md`, `corpus/`,
  `tools/`, or `_salvage/`,** and no `doctor.go` or `check.go`. It writes no `docs/` file outside
  Task 8's drafts.
- **Imports follow the spec's table** (`:198-205`): `internal/ops` gains `internal/identity`, its
  one permitted internal import; `internal/ops/opshttp` imports `ops`, `identity`
  (deviation D6), and the standard library; `internal/backup` imports `store`, `atomicfile`, and
  the standard library. The engine has no HTTP knowledge (`internal/engine/store` imports no
  `net/http`), and no package this pass touches reads the environment except
  `cmd/dubplate-server`.
- **No new module dependency.** rclone stays a subprocess (spec prior art).
- **The store's rules carry over** (6b-i's global constraints): no process that holds a store
  connection opens the database, `-wal`, or `-shm` file through `os`; every DSN comes from the one
  builder (`internal/engine/store/conn.go:187`); times are integer UTC microseconds; every store
  write goes through `Handle.Update`; new code matches no raw SQLite result code outside the
  store's existing classifiers. Restore's renames and removals of those files run with no store
  connection open in the process: the session closes the quiesce and verification connections
  before any rename, and opens its read handle only after the swap.
- **Nothing is exported ahead of a cross-package caller.** At the close every new export has a
  production caller outside its own package and tests, with the exceptions exit gate 14 names:
  everything in `opshttp` (unexported by construction, spec `:595-596`), the network actor's
  constructor (its first production caller is rung 14's authorizer), `ServeOptions.Reconcile` and
  `backup.Reconcile` (first production caller 6d's `serve`), and serve mode's open, already
  6b-i's named exception. `ops.Codes` and the classifier have a caller outside `ops` in the
  projection's non-test code (R-6bii5), but that code is unmounted, so no binary reaches either
  until rung 14. Exit gate 14 says so rather than claiming reachability.
- **Every restore test names its fixture's crash, partial-write, concurrent-process, and
  wrong-archive state**, and each state's report is asserted by text, never by "an error
  occurred". A crash is a real process death: the step hook exits a re-executed test binary (spec
  `:283-284`). Every row in which a resume finishes asserts R-6bii7's completion tuple, never
  a nil error alone.
- **Assertion discipline at the repo's bar.** No silent-success fakes, no vacuous assertions, no
  expectation derived from the code under test. Every hard-coded expectation cites the spec, a
  quoted document, or a measurement this plan records.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar. `ctx` first where a call can
  block. The em dash is banned in Go comments. No process citation in a shipped comment.
- **No `t.Parallel()` in `internal/engine/store`** (6a's umask rule). A test that signals a
  re-executed process starts it with `SysProcAttr.Setpgid`.
- **A real `rclone` is a gate prerequisite** (6b-i Task 4). A test that needs it fails, never
  skips, when `exec.LookPath("rclone")` fails.
- **The gate is `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`,**
  run with the worktree as its working directory and re-issued on exit 75 until it prints
  `gate exit:`. Anything reported as evidence runs with `-count=1` and reports wall time. Specific
  files only. Commit footer `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` for an
  `opus` task and the dispatch's footer otherwise.

---

## Privileged fields and their channels

| Field | Channels that may set it at 6b-ii | What refuses every other channel |
| --- | --- | --- |
| An HTTP caller's actor | The authorizer the projection was built with, from whatever its `ctx` carries (R-6bii2). The authorizing layer passes only the authorizer's actor downstream. | No request type has an actor field; the body decoder ignores one (spec `:540-541`, the contract suite's actor case). The HTTP layer passes the zero `Actor` into the layer, and the layer discards the `actor` argument it receives (R-6bii3). The projection never reads a header, cookie, or query parameter for identity. |
| The privilege an operation requires | `internal/ops`'s one privilege table (R-6bii1), read by both the catalog and the authorizing layer. | No request carries a privilege; the catalog cannot drift from what the layer enforces, and the table's values are asserted as literals. |
| A job's `cancel_requested` over HTTP | The cancel operation, after the authorizer admits the caller at the operation's privilege. | A denied request reaches no seam method (criterion 2). |
| The restore-intent record | `restore` under the exclusive lock (write, each state change, removal); serve mode under the exclusive lock (removal of a `reconcile-pending` record only, after its report lands). | Write mode, serve's shared path, and inspect only read it, and write mode and serve re-read it after winning the lock (R-6bii7); a new restore refuses beside it; `backup` and every `member` verb refuse beside it. |
| The database file at `db/dubplate.db` | Serve mode's create and migrations (6a); restore's swap, under the exclusive lock, with an intent record naming both files. | Write mode never creates or replaces it. |
| A pre-restore copy (`snapshots/pre-restore-<stamp>.db`) | Restore's swap, by renaming the current database; deletion by 6b-i's post-backup local prune and ruling 7 (`internal/backup/prune.go:168-200`). | No other path names it. |
| A discarded store (`snapshots/discarded-<stamp>.db`, `-wal`, `-shm`) | Restore's swap with `--discard-current`, under the exclusive lock, after the record names all three (deviation D8). | Nothing deletes them: 6b-i's local prune matches only pre-migration and pre-restore names (`isLocalCopyName`, `prune.go:204-222`). |
| Restore's staging files (`staging/restore-<stamp>.db`, rclone's `.partial` beside it, and a verified stage's `-wal` and `-shm`) | Restore under the exclusive lock, which clears `staging/` of every entry when no intent record exists (R-6bii6). | Serve's cleanup (R-6a7) and a backup's stage removal (`staging/backup-*.db` only) are the other two removers; neither runs beside an intent record. |
| The reconciliation report (`restore-report-<stamp>.txt` in the leaf, 0600) | Restore's final step, or serve mode's `reconcile-pending` path, through the store (R-6bii7), with `atomicfile.WriteStream`. | Nothing else writes it; nothing deletes it at 6b-ii. |
| The backup reference after a restore | Restore, on Ruling 5's yes, with `--snapshot` only, through the restore session's reference write, which shares `WriteBackupRef`'s body (R-6bii6). On Ruling 5's no, the session has no reference write. | 6b-i's channels otherwise (`snapshot.go:353-397`). |
| The library root's contents | `restore --library` through `rclone copy`, into a root that exists; refused into a non-empty root without `--into-nonempty`. | Restore never creates the library root and never deletes a file under it; on Ruling 6's yes, with `--into-nonempty`, a replaced file moves to the displaced directory, which restore removes only when it is empty (R-6bii11). |
| The remote a restore reads | `DUBPLATE_BACKUP_REMOTE` through the backup integration class, parsed by `backup.ParseRemote` (`internal/backup/remote.go:21-35`). | No flag names a remote; restore reads the one the backups wrote to, and warns when it differs from the remote the backup reference names. |

---

## Resolutions this plan takes rather than discovers

Each resolution names its site. A task landing the surface a resolution decides writes nothing
about the resolution in a shipped comment; the close maps each resolution to its landing commit.

### R-6bii1: one privilege table in `internal/ops`

6b-i put each operation's privilege inside `ops/local`'s `Catalog` method
(`internal/ops/local/local.go:121-129`), so a second reader would restate it. The projection's
authorizing layer must consult "each operation's declared privilege" (spec `:597-598`), and it
imports only `ops`. So `internal/ops` gains one unexported table from operation name to privilege
and one exported lookup, and `ops/local`'s `Catalog` builds its entries from that lookup. The
table covers exactly the four names in `internal/ops/catalog.go:63-69`, with the values 6b-i
shipped (`local.go:121-129`): `start_backup`, `job_status`, and `cancel_job` are admin, and
`catalog` is member. Tests assert those four values as literals, never by comparing the lookup with
itself. An unknown name has no privilege, and the layer answers `not_found` for it before any
authorizer call. `Catalog` returns an error when the lookup lacks one of its four names; it never
lists a zero privilege.

### R-6bii2: the network actor, the authorizer, and the channel

`internal/ops` gains the network kind of `Actor` (6b-i R-6bi4): built from an
`identity.Identity`, refused unless `Identity.Valid()` holds, and rendering as
`member:<principal>`. `Valid` already refuses `ChannelCommandLine`
(`internal/identity/identity.go:204-211`), so a command-line identity can never become a network
actor. The actor keeps the identity's channel, and its log form names it.

The authorizer is an unexported interface in `opshttp` with one method:
`authorize(ctx context.Context, op string, p ops.Privilege) (ops.Actor, error)`. It receives no
`*http.Request`, so the in-process leg and the HTTP leg call it the same way. A verified identity
reaches it only through `ctx`: rung 14's session middleware will set it there, as rung 5's admin
gate passes its verified identity under a context key (`internal/auth/admin.go:84-91`,
`adminKey`), and the contract suite's test authorizers read a test key the suite sets on both legs.
A nil error admits; any error denies. The projection's constructor takes the authorizer and the
mount's channel (`identity.Channel`), and fails on a nil authorizer and on a channel that is not
`Valid` or is `ChannelCommandLine`. The layer refuses, as `forbidden`, an actor whose channel
differs from the mount's, so an authorizer wired to the wrong mount fails closed. `ChannelCookie`
is the only network channel `internal/identity` declares today, and the bearer channel is rung
14's (CH `:107-110`), so the one mismatch a test can build is an authorizer that returns
`ops.LocalActor`, whose channel reads as command-line.

The spec's table lets `opshttp` import `ops` only (`:203`), but a constructor parameter of type
`identity.Channel` must name that type, so `opshttp` also imports `internal/identity` (deviation
D6). `internal/identity` imports no other internal package and forbids gaining one
(`internal/identity/identity.go:1-8`), so the edge pulls in no server package, which is what the
table guards against (spec `:207`, "`internal/ops` imports no server package").

### R-6bii3: the authorizing layer is a seam decorator

The projection composes two layers. An unexported authorizing layer implements `ops.Seam` over
another `ops.Seam`: for each method it looks up the privilege (R-6bii1), calls the authorizer, and
on a denial returns `forbidden` without calling through. It discards the `actor` argument its own
caller passes and calls through with the authorizer's actor only, so no caller of the layer can
choose the actor a job records. The HTTP layer decodes the request, calls the authorizing layer
with the zero `Actor`, and encodes the answer. The layer is the one place authorization happens,
which is why R-6bii4's request refusals come first: a gate ahead of the decode would be a second
authorization site that the in-process leg never exercises. The contract suite's in-process leg
runs against `ops/local` wrapped in the authorizing layer (deviation D1), and its HTTP leg runs
through the whole projection over `httptest`. A denial calls no seam method, so it writes no job
record, which criterion 2 checks by counting `jobs` rows.

### R-6bii4: the wire mapping

Every operation is `POST /<name>`, where `<name>` is the catalog name
(`start_backup`, `job_status`, `cancel_job`, `catalog`), relative to wherever rung 14 mounts the
handler. The body is the operation's request type as one JSON object, and the response body is the
response type. `internal/serve/origin.go` documents that "GET, HEAD and OPTIONS are always allowed"
by the origin check and that "a handler must therefore never change state on one of them", so no
operation answers a GET. One POST shape for all four keeps the catalog name the route with no
second table.

The projection dispatches by exact path match itself rather than through `http.ServeMux`
patterns, because the mux answers a method mismatch and a miss with plain-text bodies of its own
(`net/http/server.go:2751-2762` at go1.27.1: `Error(w, StatusText(StatusMethodNotAllowed), ...)`
and `NotFoundHandler()`), and every answer here carries the JSON error shape.

Error answers carry `{"error": {"code": "<code>", "reason": "<reason>"}}`. The status per code:
`invalid_request` 400, `forbidden` 403 (Ruling 3), `not_found` 404, `conflict` 409, `unavailable`
503, `canceled` 503 (Ruling 2), and `internal` 500. On Ruling 2's no, the table has no `canceled`
row. On Ruling 3's no, `forbidden` answers exactly what `http.NotFound` answers, with no JSON
body, and the constructor's table check treats that entry as a status like any other. An
`internal` answer carries no reason: the
underlying error is logged with the operation name and never written to the body, since a store
error can name a path. A coded answer's reason is `(*ops.Error).Reason` read through
`errors.As`, never `err.Error()`, because a wrapper around a coded error may add a path. A non-POST
answers 405 with `Allow: POST` and code `invalid_request`. An
unknown name answers 404 with `not_found`. Request refusals, all `invalid_request` and all decided
before any authorizer call: a `Content-Type` whose media type is not `application/json`
(`mime.ParseMediaType`, as `internal/auth/admin.go:237-241` does), a body over the cap, an empty
body, a body that is not one JSON object, trailing data after the object, and any query string.
"One JSON object" is decided by the first token being the delimiter `{`: the mechanics lens
measured go1.27.1's `Decoder.Decode` of `null` into a struct answering the zero request with a nil
error, so a `null` body would otherwise start a backup. Trailing data is decided by a following
`Decoder.Token` returning `io.EOF` (measured by the same lens). A duplicated key keeps its last
value (measured), which changes nothing here, since no request field is privileged. Unknown
object keys are ignored, as spec `:540-541` requires; `encoding/json`'s `Unmarshal`
documents that "object keys which don't have a corresponding struct field are ignored (see
Decoder.DisallowUnknownFields for an alternative)" and that it accepts "a case-insensitive match",
so `{"ACTOR": ...}` is ignored the same way. The body cap is a named unexported constant, 4 KiB,
applied with `http.MaxBytesReader`, which "returns a non-nil error of type *MaxBytesError for a
Read beyond the limit": a 4096-byte body is read and a 4097-byte body is refused. A handle is
validated by `ops.ParseHandle` (`internal/ops/ops.go:52-60`).

### R-6bii5: the `internal` and `canceled` sentinels, one classifier, and the code set

`internal/ops` declares `ErrForbidden` and `ErrInternal` with codes `forbidden` and `internal`
(completing the spec's seven, `:543-546`), one exported function that returns the wire code of any
error, and one exported function, `Codes`, that returns every code the sentinel table holds. The
classifier answers the code of the first `*ops.Error` in the chain (`errors.As`), else `canceled`
when the chain holds `context.Canceled` (Ruling 2's yes), else `internal`. The projection encodes
only the classifier's answer. The contract suite asserts that the in-process error and the HTTP
answer classify to the same code for every case.

`Codes` has a production caller: the projection's constructor checks its code-to-status table
against `Codes` and fails, naming the code, when a code has no status. The contract suite's
coverage guard iterates the same function, so an eighth sentinel added to `ops` without a contract
case fails the guard by name. A guard over a list the test hard-codes would only prove that the list
matches itself.

On Ruling 2's yes, each `ops/local` seam method wraps its error as `canceled` when the caller's
`ctx.Err()` is `context.Canceled` at the method's return. It keys on `ctx.Err()` rather than on the
error's chain, because a wait inside SQLite's busy handler returns `SQLITE_BUSY` with no
`context.Canceled` in its chain (the mechanics lens measured `database is locked (5)` after the full
5 s busy timeout, with the context canceled at 200 ms), and `classifyErr` wraps with `%v`
(`internal/engine/store/conn.go:221-222`). `context.Context.Err` documents: "If Done is closed, Err
returns a non-nil error explaining why: DeadlineExceeded if the context's deadline passed, or
Canceled if the context was canceled for some other reason." An in-process caller then matches
`ops.ErrCanceled` with `errors.Is`, which "reports whether any error in err's tree matches target."
A job's own context is separate from the caller's (`internal/ops/local/jobs.go:108`), so this
wrapping never touches a job's terminal state.

The in-process wait that does honor the context is the write pool's one connection:
`Handle.Update`'s "ctx ... bounds the wait for the write pool's one connection inside this
process" (`internal/engine/store/write.go:50-52`). `database/sql`'s connection request, at go1.27.1
(`$(go env GOROOT)/src/database/sql/sql.go:1370-1375`), selects on `case <-ctx.Done():` and returns
the context's error. The mechanics lens measured it: a waiter canceled at 200 ms returned `context
canceled` at 200 ms. The contract case uses that wait.

The `internal` contract case is spec `:604`'s: the store handle is closed under a live runner, and
a status read then fails. `database/sql` answers a closed `*DB` with `errDBClosed`,
`errors.New("sql: database is closed")` (`sql.go:1318`, returned by `DB.conn` at `:1322-1325`), and
`classifyErr` passes an error that is not SQLite's through unchanged
(`internal/engine/store/conn.go:216-218`), so the classifier answers `internal`. The mechanics lens
measured that string through `QueryRowContext`.

### R-6bii6: restore's package split, the session, and its staging rule

The store exports one restore session, opened under the exclusive lock, whose methods are the
leaf operations of spec steps 1 to 5, the reference write (Ruling 5's yes only), the
reconciliation handle, and the
report write of step 7: take the lock, create the leaf when absent, clear staging, stage a local
file, verify a staged file, quiesce the current store, write and advance the intent record, swap in two calls (R-6bii8),
write the backup reference, open a read handle on the restored store, write the report, and remove
the record. It also hands out close-on-exec duplicates of the exclusive lock's descriptor, as the
kind lease does (`internal/engine/store/lease.go:220`, `lock.go:176-194`). `internal/backup` owns
the flow's order, the step hook, the fetch, the library copy, and the reconciliation (deviation
D3).

**The session open's refusals, before any stage exists.** The uid check (spec `:248-252`); the
leaf's creation when absent, under 6a's E9 state-directory owner rule (`leaf.go:70-99`); the lock,
taken non-blocking; the filesystem probe; and two refusals the spec does not list (deviation D10).
First, a new restore (`--snapshot` or `--from-file`) beside any intent record refuses, naming
`restore --resume`; only `--resume` proceeds beside one. Second, `db/`, `staging/`, and
`snapshots/`, each resolved with `filepath.EvalSymlinks`, must share one device, read as
`st_dev` from `info.Sys().(*syscall.Stat_t)` as `checkOwner` already reads the owner
(`leaf.go:105`). `os.Rename` across devices fails, so a mismatch found after step 4 would leave a
record every resume fails on in the same way; spec E6 already contemplates a symlinked `db/` on
another filesystem. A refusal names each directory and its device.

**The session's store handle.** Step 7's query and Ruling 5's reference write need the leaf while
the session holds `LOCK_EX`, and a second lock from the same process conflicts with it. flock(2),
as the mechanics lens quotes it: "An attempt to lock the file using one of these file descriptors
may be denied by a lock that the calling process has already placed via another file descriptor."
So the session never calls `OpenWrite` or anything that takes the store lock. It opens a read
handle on the restored store that holds no lock of its own (the session's lock covers it), at the
binary's version only. On Ruling 5's yes, the reference write is a session method, and
`WriteBackupRef`'s body and earlier-reference prune move into one unexported leaf-level function
that both `(*Handle).WriteBackupRef` and the session call, so the two cannot drift. On Ruling 5's
no, neither the method nor the move exists, since the method would be an export with no caller.

The transport generalizes to any source of inherited descriptors, so a restore's rclone child
inherits the exclusive store lock the way a backup's inherits the kind lease and shared lock
(`internal/backup/rclone.go:42-78`). An orphaned restore transfer then keeps every other writer
out until it exits, the property 6b-i R-6bi12 established for backup. The mechanics lens confirmed
the inheritance: the store lock is `flock`, not `fcntl` (`lock.go:88`), and flock(2) says that
"duplicate file descriptors (created by, for example, fork(2) or dup(2)) refer to the same lock."

The transport also becomes the package's one rclone launch path, as 6b-i's `internal/backup`
architecture read directs (`2026-09-25-rung-6b-i-architecture-reads.md`). Today `prune.list`
(`prune.go:248-278`) and `runRclone` (`prune.go:280-294`) call `exec.CommandContext` directly,
with no process group, no SIGINT-then-`WaitDelay` cancel, and no inherited descriptors, though
`internal/backup/doc.go` promises all three for rclone. So `transport.run` takes a stdout writer,
`prune` holds a transport, and `runRclone` is deleted. Restore's `snapshots/` listing and its
`rclone size --json` then run under the same cancel as its `copyto` and `copy`, and a SIGTERM
during either exits 130 (Ruling 4), where the direct launch would kill rclone and return an error
that does not wrap `ctx.Err()`.

**Restore's staging rule** (6b-i handoff item 5): a restore that holds the exclusive lock and finds
no intent record removes every entry in `staging/`, as serve's cleanup does
(`internal/engine/store/open.go:543-554`). The exclusive lock proves no other process stages: a
backup's orphaned transfer inherits the shared lock and would have refused the exclusive one. With
an intent record present, restore removes nothing in `staging/` except through the steps the
record names. After each verification of a stage, the session removes the stage's `-wal` and
`-shm` siblings, refusing if the `-wal` is not empty: the mechanics lens measured an inspect-DSN
open of a byte copy of a cleanly closed WAL database creating a 0-byte `-wal` and a 32768-byte
`-shm` that stay after the close.

### R-6bii7: the intent record's format, its states, and when an open reads it

This settles STATUS's carried item "To 6b-ii: the intent record's format." The record is
`restore-intent` in the leaf (`internal/engine/store/leaf.go:53`), written only through
`atomicfile.WriteStream` at 0600, which writes a temp file created with "mode 0o600 (before
umask)" (`os.CreateTemp`), syncs it, renames it, and fsyncs the directory
(`internal/atomicfile/atomicfile.go:55-89`). Its body is one JSON object decoded with
`DisallowUnknownFields`, carrying a format number, the state, the restore's stamp, the source (a
remote stamp with the remote, or a local path), the incoming staged file's basename, the
pre-restore copy's basename (empty when no current store existed), whether `--discard-current`
applies, whether the library half is owed with its resolved root and `--into-nonempty`, and the
incoming store's version. Every name inside the leaf is a basename whose directory its field fixes:
the incoming file lives in `staging/`, the pre-restore copy in `snapshots/`, and the discarded
names are derived from the stamp (`snapshots/discarded-<stamp>.db` and its `-wal` and `-shm`). A
name holding a separator or `..`, or an empty name where the state needs one, is refused.

**Every writing open reads the record after it wins the lock, not only before.** 6a's opens read it
in `leaf.stat()` before waiting (`open.go:122`, `:202`, `:361`), and nothing reads it again on
`OpenWrite`'s path after `acquireShared` (`:140`) or on `OpenServe`'s shared path after `waitLock`
(`:219-231`). Both lenses found the consequence, and the mechanics lens proved it: an open waiting
on the lock while a restore writes its record then proceeds past the record once the restore dies
(both opens returned a nil error beside a `swap` record). A `backup` waits up to `backupVerbWait`
(30 s, `cmd/dubplate-server/backup.go:80`) and a restarting serve up to `MigrationWait`, so this is
reachable on the box. So `OpenWrite` re-reads the record after `acquireShared`, and `OpenServe`'s
shared path re-reads it after each shared `waitLock`; a record found there releases the lock and
takes the exclusive path, where `settle`'s check decides. That routing is also what makes serve's
"`reconcile-pending`, store current" row reachable, since the shared path would otherwise open a
current store without reconciling. Inspect keeps its one pre-lock read: it takes the lock
non-blocking, so it never waits through a restore.

| State | Written after | What `restore --resume` does | Write mode | Serve mode |
| --- | --- | --- | --- | --- |
| `swap` | step 4 | re-checks the version, finishes step 5 from R-6bii8's combination table, and writes the reference (Ruling 5) | refuses, naming `restore --resume` | refuses, naming `restore --resume` |
| `library` | step 5 and the reference write, with `--library` | re-resolves the root, refuses when it differs from the recorded root or is absent, skips the emptiness check (deviation D9), and re-runs step 6's copy | refuses, naming `restore --resume` | refuses, naming `restore --resume` |
| `reconcile` | step 5 or 6, restored version equal to the binary's | re-reads the restored version: equal runs step 7; below advances to `reconcile-pending` and names "start serve"; above refuses naming both versions | refuses, naming `restore --resume` | refuses, naming `restore --resume` |
| `reconcile-pending` | step 5 or 6, restored version below the binary's | refuses, naming "start serve to finish the restore" | refuses, naming "start serve" | accepts: migrates, reconciles, writes the report, removes the record |

**Resume re-derives every version verdict from the running binary** (deviation D10), because an
operator may change binaries between a crash and `--resume`. In `swap`, before any rename, the
staged file (or, after the swap, `db/`'s store) must be at or below the running binary's version,
else resume refuses naming both versions and the binary that wrote the record. Resume reads the
staged file's version through the same session method that verifies a stage and then removes its
`-wal` and `-shm` (R-6bii6), since an inspect-DSN open creates both, and a re-check that skipped
the removal would leave `staging/` non-empty after the swap. In `reconcile`, the table's third row
applies. The recorded version is kept for the report and the refusal text only.

**The completion tuple.** A restore or a resume has finished only when all of these hold, and every
test in which one finishes asserts them all: `db/dubplate.db` holds the incoming snapshot's rows;
`snapshots/pre-restore-<stamp>.db` holds the prior store's rows (when a current store existed
and `--discard-current` did not apply), with no `-wal` or `-shm` beside it; with
`--discard-current`, `snapshots/discarded-<stamp>.db` exists, each `-wal` or `-shm` the fixture
planted sits beside it under the discarded name, and none of the planted files remains in `db/`;
`staging/` is empty; the report file exists at 0600, or the
record is `reconcile-pending`; the record is gone, or is `reconcile-pending`; and the backup
reference follows Ruling 5.

Every mode meets these record states too, each with its report:

| Record | Report, every mode |
| --- | --- |
| absent | no intent; each mode proceeds as today |
| written while the open waits on the lock | write mode refuses naming `restore --resume`; serve takes the exclusive path and `settle` decides by state |
| empty file | refused, naming the path and "restore --resume"; resume refuses it as unreadable, naming the path and the manual procedure (Ruling 9) |
| malformed JSON, an unknown key, or an unknown format number | the same refusal; resume refuses naming the parse error and the manual procedure |
| an unknown state | the same refusal; resume refuses naming the state and the manual procedure |
| a name with a separator or `..`, or an empty required name | the same refusal; resume refuses naming the field and the manual procedure |
| an `atomicfile` temp file (`restore-intent.tmp-*`) beside no record | ignored by every mode: a crash before the rename left no record |

Inspect keeps reporting `StateRestoreInProgress` (`internal/engine/store/inspect.go:25`), and its
`Reason` names the record's state and the action that finishes it.

Serve mode's `reconcile-pending` path: holding the exclusive lock, it clears staging (the swap has
already renamed the incoming file into `db/`, so no file the record names is in `staging/`),
migrates as today, calls `ServeOptions.Reconcile` with a handle on the migrated store, writes the
returned report text through `atomicfile.WriteStream` at 0600, and removes the record, fsyncing the
leaf. A nil `Reconcile` beside a `reconcile-pending` record refuses, naming the missing reconciler,
and leaves the record. A reconciler error refuses and leaves the record, so the next start retries.
The store writes the report; `internal/backup` never learns the leaf's paths (deviation D2). The
report lists every library path, so restore's own report write uses the same call and mode.

### R-6bii8: the quiesce, the swap, the discard, and what is quoted or measured

Step 3 opens the current store on a maintenance-class connection that skips the version check
(`maintenanceDSN`, `conn.go:160`, which "omits _journal_mode"), runs `PRAGMA
wal_checkpoint(TRUNCATE)`, then `PRAGMA journal_mode=DELETE`, closes, and verifies that neither
`-wal` nor `-shm` remains. sqlite.org, quoted: TRUNCATE "works the same way as RESTART with the
addition that the WAL file is truncated to zero bytes upon successful completion"
(`pragma.html`); "the journal_mode cannot be changed while a transaction is active"
(`pragma.html`); "When the last connection to a database closes, that connection does one last
checkpoint and then deletes the WAL and its associated shared-memory file" (`wal.html`); and "PRAGMA
journal_mode=WAL is persistent" (`wal.html`).

**Measured by the mechanics lens** (SQLite 3.53.4 through modernc v1.59.0, a WAL left by a child
killed with SIGKILL):

- **The only connection.** The checkpoint answers busy 0 and truncates `-wal` to 0 bytes;
  `journal_mode=DELETE` answers `delete`, and `-wal` and `-shm` are gone before the close. This
  settles revision 1's open question.
- **Another process has the database open, idle.** The checkpoint answers busy 0;
  `journal_mode=DELETE` fails at once with `SQLITE_BUSY`, whatever the busy timeout; `-wal` and
  `-shm` remain.
- **Another process holds a read transaction.** The checkpoint answers busy 1 and `-wal` keeps its
  bytes; `journal_mode=DELETE` fails with `SQLITE_BUSY`.

So step 3 refuses in three distinct ways, each naming the database and "another process may have
the database open": a nonzero busy column (an open read transaction), an `SQLITE_BUSY` from the
switch through `classifyErr` (an idle connection), and a switch answering any mode but `delete`.
The explicit checkpoint stays for its busy column, which names the read-transaction case before
the switch fails less specifically. The leftover-file check after the close is the net under all
three; its fixture is a store already in DELETE mode with a planted `-shm`, which the switch never
touches.

**The discard** (deviation D8). A current store that cannot open is refused, naming
`--discard-current`. With the flag, step 3 opens nothing, and the record written at step 4 carries
the discard. Step 5's first renames then move `db/dubplate.db`, `-wal`, and `-shm`, each when
present and in that order, to `snapshots/discarded-<stamp>.db`, `-wal`, and `-shm`. The names are
distinct from `pre-restore-*`, so step 5 never collides with them, 6b-i's local prune never deletes
them (`isLocalCopyName` matches only pre-migration and pre-restore names, `prune.go:204-222`), and
serve's lost-database guard never suggests restoring an unopenable file. Nothing opens the leaf in
the window, since the record already exists.

**The swap.** Step 5 renames the current database to `snapshots/pre-restore-<stamp>.db` (or runs
the discard), then the staged file to `db/dubplate.db`, then fsyncs `snapshots/`, `db/`, and
`staging/` through the existing helper (`snapshot.go:175-188`). `os.Rename` "replaces" an existing
non-directory `newpath`, and fails across devices, which the session open's device check rules out
(R-6bii6). The pre-restore name uses `store.PreRestorePrefix` and the stamp layout, which `isCopy`
already recognizes (`open.go:33`, `:503-506`).

**Step 5 is two session calls.** The first renames the current database to its pre-restore name,
runs the discard's moves, or does nothing when no current store exists. The second renames the
staged file into `db/` and runs the fsyncs. The flow in `internal/backup` calls the step hook
between the two calls, so the hook stays the spec's one unexported function variable in
`internal/backup` (spec `:283-284`), and the store gains no hook of its own, which would be a
fourth test seam. The store's tests reach the in-between states through R-6bii8's hand-built
combination fixtures; Task 5B's crash row reaches them through a real process death. Resume after a
crash inside step 5 finds one of these, keyed first on what the record owes:

| Record owes | Current at `db/` | Pre-restore or discarded copy | Incoming in `staging/` | Resume's action |
| --- | --- | --- | --- | --- |
| a pre-restore copy | present, no `-wal` or `-shm` | absent | present | both renames, then the fsyncs |
| a pre-restore copy | present, with `-wal` or `-shm` | absent | present | refused, naming each: a process wrote after the record |
| a pre-restore copy | absent | present | present | the second rename, then the fsyncs |
| a pre-restore copy | present | present | absent | the fsyncs; the swap had finished |
| no copy (no current store) | absent | none owed | present | the rename, then the fsyncs |
| no copy (no current store) | present | none owed | absent | the fsyncs; the swap had finished |
| a discard | any of the three files still in `db/` | any subset moved | present | move each file still in `db/`, then the rename and the fsyncs |
| a discard | present | the database moved | absent | the fsyncs; the swap had finished |
| any | any other combination | | | refused, naming every file the record names, whether each exists, and the manual procedure (Ruling 9) |

The next open restores WAL through the write DSN's `_journal_mode=WAL` (`conn.go:133`), which spec
`:701` relies on and 6a handoff item 5 flags. The mechanics lens measured it: a rollback-journal
file (header bytes 1/1) opened on the write DSN reads 2/2 and `journal_mode=wal`. Task 4C still
proves it on the swapped-in file.

### R-6bii9: the sources a restore reads, and every wrong-archive state

`--snapshot <stamp>` reads `snapshots/<stamp>.db` from the remote; `--snapshot latest` lists
`snapshots/` and takes the greatest stamp that parses through `store.ParseStamp`
(`stamp.go:31`), reusing 6b-i's listing, launched through the transport (R-6bii6), which reads a
missing directory as empty
(`internal/backup/prune.go:248`; rclone's `lsjson` exits 3 on one, 6b-i R-6bi15). Every
`--snapshot` restore lists `snapshots/` first, and "no such object" is decided from that listing,
before any `copyto`: the mechanics lens measured `copyto` of a missing object retrying three times,
exiting 3, and ending on `NOTICE: Failed to copyto: directory not found`, which names no path. The
same listing supplies `latest`, the `--older` rule's latest stamp, and the object's `Size` for
R-6bii10. Destinations are built by the remote's join (`remote.go:36-38`), never by concatenation
(6b-i handoff item 6).

The fetch is `rclone copyto` into the stage path. rclone's help, quoted: `copyto` will "copy it to
dst, overwriting an existing file if it exists", and by default it downloads through a temporary
file, since `--inplace` means "Download directly to destination file instead of atomic download to
temp/rename" and `--partial-suffix` is "Add partial-suffix to temporary file name when --inplace is
not used (default \".partial\")". **Measured by the mechanics lens:** the file `copyto` renames
into place is a new inode at 0644 under umask 022 (0600 under umask 077), and a pre-created 0600
stage is replaced by it, so the `--snapshot` path creates no stage of its own. The mode `-M` would
carry comes from local-backend metadata, so the fetch does not rely on `-M`. The restore chmods the
stage to 0600 before any open, and the leaf's 0700 directories (`leaf.go:19`) expose nothing in the
window. Mid-transfer, rclone's temporary file is `<name>.<hex>.partial`; after SIGINT rclone exits
130 and removes it, and only a SIGKILL leaves it (measured, `rclone.go:69` cancels with SIGINT).

`--from-file PATH` copies the file into a pre-created 0600 stage with `io.Copy` and `File.Sync`,
which "commits the current contents of the file to stable storage"; the source is never renamed,
since it is often a copy in `snapshots/` the operator may need again. "The live database" is
decided by `os.SameFile` on the two paths after `filepath.EvalSymlinks`, never by comparing
strings.

| Source | State | Report |
| --- | --- | --- |
| `--snapshot <stamp>` | the stamp does not parse | refused before any rclone call, naming the stamp and the layout |
| `--snapshot <stamp>` | no such object in the `snapshots/` listing | refused before any `copyto`, naming the remote path |
| `--snapshot latest` | `snapshots/` empty or absent | refused, naming the remote path |
| `--snapshot <stamp>` | not the latest, without `--older` | refused, naming the latest stamp and `--older` (ruling 11) |
| `--older` | with `--library` | refused before any rclone call (ruling 11) |
| `--snapshot` | backup integration off, or the remote invalid | refused, naming `DUBPLATE_BACKUP_ENABLED` or the parse error |
| `--snapshot` | the configured remote differs from the one the backup reference names | proceeds, with a printed warning naming both remotes (the one wrong-archive signal before 6d's store identity) |
| either | the fetch fails (rclone nonzero) | refused, naming rclone's exit and last stderr line; the stage removed; no record written |
| either | a truncated or non-SQLite file | refused, naming `integrity_check`'s answer or "not a database"; the stage removed |
| either | a zero-byte file | refused naming `application_id` 0: `verifyCopy` says "An empty file passes the integrity check" (`snapshot.go:108-109`), and a zero-byte file reads version 0, which the widened version rule accepts |
| either | a SQLite file without the store marker | refused, naming its `application_id` |
| either | `user_version` above the binary's | refused, naming both versions (spec `:691-692`) |
| either | a store whose migration ledger disagrees with the binary's first `user_version` migrations | refused, naming `checkLedger`'s answer (`migrate.go:190`) |
| `--from-file` | the path is absent, a directory, or unreadable | refused, naming the path |
| `--from-file` | a `-wal` or `-journal` file exists beside it | refused, naming it (deviation D4) |
| `--from-file` | the path is the live database, or inside `staging/` | refused, naming the path |
| either | a snapshot from another dubplate store | **not detectable at 6b-ii**; the store identity is 6d's forward item (`draft-STATUS-6b-i.md`); a declared gap |

The verification reuses the store's copy verifier (`verifyCopy`, `snapshot.go:113-151`) with a new
parameter that accepts any version at or below the binary's, since an older snapshot is valid
input. Backup's call keeps exact equality, and a test proves it still refuses a copy one version
below. The ledger check runs `checkLedger` on the stage for its own version, so a stage from a
divergent build is refused before the record rather than at every open after the swap.

### R-6bii10: the free-space check

Step 1's check (spec `:690`) counts the staged store only. The swap renames within one device,
which the session open's device check guarantees (R-6bii6), so the pre-restore copy costs no space.
The size is the remote object's `Size` from the `snapshots/` listing, or the local file's `os.Stat`
size, plus the WAL bound the store already adds (`snapshot.go:156-172`). With `--library`, it also
counts the remote tree's bytes from `rclone size --json`, launched through the transport with its
stdout captured (R-6bii6), against the library root's filesystem.
rclone's help says "Some backends do not always provide file sizes ... and count them in as empty
files in the output of the size command," so the library check is a floor. The mechanics lens
measured the JSON as `{"count":2,"bytes":20004096,"sizeless":0}`, so the verb prints the
`sizeless` count beside the floor. An object whose `Size` is -1 skips the store check with a
printed line naming it. `rclone size --json` of an absent remote `library/` exits 3 (measured), so
`--library` lists the remote root first and refuses a remote without `library/`, naming the remote
path, before any size or copy call.

### R-6bii11: the library half and the reconciliation report

Step 6 resolves the library root once with `filepath.EvalSymlinks`, as backup does
(`internal/backup/guard.go:23`, 6b-i R-6bi18). An absent root refuses, naming the path: restore
never creates it, and `rclone copy` would create it ("If dest:path doesn't exist, it is created").
A non-empty root without `--into-nonempty` refuses naming the flag (spec `:702-703`); on Ruling 8's
yes, one directory entry named `lost+found` does not count. The emptiness check runs once, before
the record enters `library`. A resume in `library` re-resolves the root, refuses when it is absent
or differs from the recorded root, and skips the emptiness check (deviation D9), since a copy the
restore itself interrupted leaves the root non-empty. The copy is `rclone copy <remote>/library
<root>` under the transport with the exclusive lock inherited.

**The displaced directory** (Ruling 6's yes). `--backup-dir <root>.restore-displaced-<stamp>`, a
sibling of the resolved root, applies only with `--into-nonempty`: without that flag the root was
empty when the record entered `library`, so every file in it came from this restore's own copy,
and a resumed copy has no operator file to displace. Three rules keep the directory safe across a
resume, since its stamp comes from the record and a resumed step 6 names the same path:

- A first run creates the directory with `os.Mkdir` at 0700 before the copy, and any failure,
  `EEXIST` included, refuses before any rclone call. The pre-creation does real work: the
  verification read probed rclone v1.74.3 creating a missing backup directory at 0755 under umask
  022 (`2026-09-25-rung-6b-ii-plan-fold-verification.md`, section 3).
- A resume in `library` accepts the path when it already exists, after an `os.Lstat` shows a
  directory (never a symlink), owned by the caller's uid, at mode 0700. Any other state refuses,
  naming the path and what it found. An absent directory is created as on a first run.
- After the copy completes, the directory is removed by `os.Remove` only when it holds no entry.
  A directory that holds files is kept and named in the report with its file count, whichever run
  displaced them. Nothing ever removes it recursively.

**An unmounted mount point is a declared gap** (Ruling 7's yes). It is an existing, usually empty
directory, so neither refusal fires, and the copy fills the filesystem underneath. Before the copy
the verb prints the resolved root, whether its device differs from its parent directory's (it is a
mount point) or not, and the free space the library check read. Task 8 owes rung 16 the runbook
line "mount the library volume before a library restore." On Ruling 7's no, step 6 refuses an
empty root on its parent's device unless `--root-not-a-mount` is given, naming that flag. The flag
is valid only with `--library`. The check runs once, beside the emptiness check, and a resume in
`library` skips it for the same reason (deviation D9).

The reconciliation (step 7) reads every track's collection `root_path` and `rel_path` through a
read-pool query (the store exports a multi-row read beside `Handle.QueryRow`) on the session's
read handle (R-6bii6), joins each under the resolved root as bytes, and `os.Lstat`s it: absent or
not a regular file is a row without a file. It walks the root with `filepath.WalkDir`, which "does
not follow symbolic links", and every regular file whose root-relative bytes match no track's
joined path is a file without a row. The report file carries every entry, one per line, and a
header naming the stamp, the source, both counts, and on Ruling 6's yes the displaced directory
with its file count; stdout carries the counts and at most the first 20 of each. A report with an
empty `users` table adds the line spec `:713-714` requires, "run member import." The mechanics
lens measured a non-UTF-8 library name round-tripping byte-exact through a local-to-local `copy`.
States:

| Catalog and tree | Report |
| --- | --- |
| no tracks, empty tree | clean; both counts 0 |
| no tracks, populated tree | every file listed as without a row |
| tracks, empty tree | every track listed as without a file |
| a planted row without a file and a planted file without a row | each named once (criterion 12) |
| the root absent at step 7 (no `--library`) | the report says the root is absent and lists every track; restore still succeeds |
| a walk error on one subtree | the report names the subtree and the error; the rest is still reported |

### R-6bii12: the verb, its signals, and its output

`restore` opens nothing through `ops/local`: it is not an operation (spec `:591`). It runs on the
context `executeWithSignalContext` cancels on SIGINT or SIGTERM
(`cmd/dubplate-server/main.go`, by symbol), and it also treats SIGHUP as a cancel, as `backup`
does (6b-i R-6bi11). A cancel interrupts rclone through the transport, leaves the record at the
last state a completed step wrote, and exits 130 (Ruling 4). No record write follows a cancel: every
state write sits at a step boundary, after the step's work, so a canceled step never reaches one.
The stage removal after a failed or canceled fetch is a file operation that takes no context. The
verb runs the uid check before anything (spec `:248-252`), and the lock is taken non-blocking: a
held lock exits 75, naming the holders as `busyError` renders them (`lock.go:105-111`), which is how
criterion 5's refusal reads. It prints ruling 12's instruction last: mint new values for
`DUBPLATE_SESSION_SECRET` and `DUBPLATE_MAGIC_LINK_SECRET` and leave both `_PREVIOUS` keys unset,
so every outstanding session and link fails (the four names are registered at
`cmd/dubplate-server/auth.go:48-59` on master). The instruction names the keys and never an estate
tool.

---

## Review focus

These are the input classes most likely to bite, each pinned by a test in the task that owns it.

1. **A request body that names an actor or a privilege, or a caller of the authorizing layer that
   passes its own actor.** Expected: ignored, and the job records the authorizer's actor. Pinned in
   Tasks 2 and 3.
2. **A denying or mis-mounted authorizer, or a member calling an admin operation.** Expected:
   `forbidden`, no seam call, no job record. Pinned in Tasks 2 and 3.
3. **An internal error's text, or a path a wrapper adds around a coded error.** Expected: logged,
   never in the body. Pinned in Task 2.
4. **A restore while `serve`, a `backup`, an inspecting `doctor`, or a backup's orphaned rclone
   holds the store.** Expected: exit 75 naming the holders, nothing changed. Pinned in Tasks 4B and
   7.
5. **A writer already waiting on the lock when a restore writes its record and then dies.**
   Expected: the writer refuses, naming `restore --resume`. Pinned in Task 4A.
6. **A process killed at every restore step, on a leaf with and without a current store.**
   Expected: from step 4 on, a record every writing open honors and a resume that reaches the
   completion tuple; before step 4, a usable current store. Pinned in Tasks 4C and 5B.
7. **A snapshot that is wrong for this store in any detectable way, or a binary changed between a
   crash and `--resume`.** Expected: refused before the record is written, or refused by resume
   naming both versions. Pinned in Tasks 4B, 4C, 5A, and 5B.
8. **An older snapshot restored under a newer binary.** Expected: no writer but serve proceeds
   until serve has migrated and reconciled. Pinned in Task 4A.
9. **A restore interrupted during the library copy.** Expected: `--resume` finishes the copy into
   the root it partly filled, and keeps every file the first run displaced. Pinned in Task 6.

A library root that is an unmounted mount point is not refused (Ruling 7); revision 1's claim that
Task 6 pins a refusal is struck, and the case is a declared gap in "Not verified."

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched by the conductor after 6b-i merges and the lane 2
worktree is cut. **Its output is blockers with evidence, not fixes.** It writes no code and edits
no file. It grades a moved pointer by content: content that survives at any offset or under a new
name is a note naming the new site, and content that is gone is a blocker.

**Blockers: a "no" stops the launch.**

1. **Has 6b-i merged, and is the worktree fresh from master?** Report `git worktree list`, the
   lane registry rows, and `git status` in both trees.
2. **Do the 6b-i surfaces this plan cites match the merge?** Report, by function, any difference
   between `33ceff5` and HEAD in `internal/ops`, `internal/ops/local`, `internal/backup`,
   `internal/engine/store`, and `cmd/dubplate-server/backup.go`, above all in `Seam`, `Actor`,
   `Error`, `codes`, `Runner.Catalog`, `requestCancel`, `OpenServe`, `settle`, `leaf.stat`,
   `checkIntent`, `clearStaging`, `verifyCopy`, `checkFreeSpace`, `syncDir`, `WriteBackupRef`,
   `KindLease.DupFiles`, `transport`, and `prune.list`, and the fold's additions: `OpenWrite`'s
   `acquireShared` path, `OpenServe`'s shared path, `checkLedger`, `classifyErr`, `isLocalCopyName`,
   `besideCopy`, `checkOwner`, and `Handle.Update`'s wait contract (`write.go:50-53`).
3. **Does `internal/ops/opshttp` not exist, and does no file declare `ErrForbidden`,
   `ErrInternal`, or a restore verb?**
4. **Is `ErrCanceled` still raised nowhere in production code?** Report `git grep -n
   'NewCanceledError\|ErrCanceled' -- '*.go' ':!*_test.go'`. A production raise changes Ruling 2's
   premise.
5. **Has lane 1 touched `internal/identity`'s `Identity`, `Channel`, or `Valid`?** 5c-ii writes
   `internal/identity` (6b-i R-6bi1). A changed field or channel set is a blocker for Task 1.
6. **Does `internal/identity` still import no other internal package?** Deviation D6 rests on it.

**Notes, not blockers:**

7. Report `newRootCmd`'s `AddCommand` block and `newProductionRegistry` at HEAD, and whether
   `IntegrationValue` answers the remote while the backup class is disabled
   (`cmd/dubplate-server/backup.go:59-65`), which R-6bii9's disabled-class row depends on.
8. Report the master pointers into `internal/auth/admin.go`, `internal/serve/origin.go`, and
   `cmd/dubplate-server/auth.go:48-59` (the four secret names) at HEAD.
9. Report the spec pointers this plan cites.
10. Report `rclone version`; a version other than v1.74.3 is a note, and Tasks 5A and 6 re-run
    the probes this plan cites against it (`lsjson` exit 3, `copyto`'s mode and SIGINT cleanup,
    `size --json`'s shape, and, on Ruling 6's yes, `copy --backup-dir`).
11. Report the conductor's recorded outcome of Rulings 1 to 9.
12. On Ruling 1's no, 6b-ii-b's fresh Task 0 re-runs items 1, 2, 9, and 10 against the tree
    6b-ii-a merged, and reports any file 6b-ii-a changed outside `internal/ops` and
    `internal/ops/opshttp`, since the halves were drawn to share no file.

**Deliverable count: 1** (the report).

---

## Task 1: `internal/ops`: the last two sentinels, the classifier, the network actor, and the privilege table

Model: `sonnet`. Depends on Task 0. **Independent by files of Tasks 4A to 7.** **Lands hook 15's
channel half at the seam** (the actor carries the mount's channel, which rung 14's two mounts
read); otherwise neutral.

**Files:**
- Modify: `internal/ops/errors.go`, `ops.go`, `catalog.go`, and their tests
- Modify: `internal/ops/local/local.go` (the catalog reads the table; the seam boundary's
  canceled wrapping) and its tests

**Outcomes:**

- **The two sentinels, the classifier, and the code set** (R-6bii5): `ErrForbidden` and
  `ErrInternal` with codes `forbidden` and `internal`, in the same table as the five
  (`internal/ops/errors.go:35-41`), so the "every sentinel has a code" test covers seven; one
  exported function from any error to its wire code; and `Codes`.
- **The network actor** (R-6bii2): built from an `identity.Identity`, refused unless it is `Valid`,
  rendering `member:<principal>` and exposing its channel. The local actor's channel reads as
  command-line.
- **The privilege table** (R-6bii1) and its lookup, with 6b-i's four values; `ops/local`'s
  `Catalog` builds from it and errors on a name the lookup lacks.
- **The seam boundary's `canceled` classification**, on Ruling 2's yes: each `Runner` seam method
  wraps its error as `canceled` when the caller's `ctx.Err()` is `context.Canceled` at return
  (R-6bii5).

Two edits the diff-reviewer should expect, both from 6b-i's `internal/ops` architecture read
(`2026-09-25-rung-6b-i-architecture-reads.md`). `TestImportsStandardLibraryOnly`
(`internal/ops/ops_test.go:185`) says it "lands with 6b-ii": it becomes a test that admits
`internal/identity` as the one internal import and still fails on any other. The doc comment at
`internal/ops/errors.go:21-23` links `[ErrForbidden]` and `[ErrInternal]`, which do not exist yet,
so it is rewritten once they do. The same read's collapse of the sentinel list, now written four
times, into a self-coding sentinel type goes to the architecture cleanup pass, not here.

**On Ruling 2's no**, this task drops `canceled` from `internal/ops`: `ErrCanceled`,
`CodeCanceled`, and `NewCanceledError` are deleted, the per-sentinel table and `Codes` cover six
literal codes, the classifier's `context.Canceled` row answers `internal`, and the seam boundary
outcome, its three rows, and mutation 4 do not exist. The task then carries three deliverables.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Classifier | an `*ops.Error` of each of the seven codes, wrapped by `fmt.Errorf` | its code |
| Classifier | `context.Canceled` in the chain, no `*ops.Error` | `canceled` (Ruling 2 yes) |
| Classifier | `context.DeadlineExceeded` alone | `internal` |
| Classifier | nil | not called; the projection answers success |
| Classifier | any other error | `internal` |
| `Codes` | the sentinel table | exactly the seven literal codes |
| Network actor | a `Valid` cookie identity | `member:<principal>`, channel cookie |
| Network actor | the zero identity, or a command-line identity | construction refuses |
| Privilege lookup | `start_backup`, `job_status`, `cancel_job` | admin, as literals |
| Privilege lookup | `catalog` | member, as a literal |
| Privilege lookup | any other name | not found |
| Seam boundary | a caller context already canceled | an error matching `ops.ErrCanceled` (Ruling 2 yes) |
| Seam boundary | a caller context canceled while `CancelJob` waits for the write pool's one connection | an error matching `ops.ErrCanceled` (Ruling 2 yes) |
| Seam boundary | a caller context alive, the store answering an error | that error's own classification, never `canceled` |

**Acceptance criteria:**

- The per-sentinel table covers seven codes with literal expected strings, and fails when a
  sentinel has no code.
- The lookup's four values are asserted as literals, citing 6b-i's shipped catalog
  (`local.go:121-129`). A second test asserts `Catalog`'s entry privileges equal those literals,
  never the lookup's output.
- `go list -deps ./internal/ops` shows `internal/identity` as its only internal import.
- On Ruling 2's yes: `Runner.JobStatus` called with an already-canceled context returns an error
  matching `ops.ErrCanceled`. **Fixture state:** a current store with one job record, so the call
  would otherwise succeed. And `Runner.CancelJob` whose context is canceled while a test goroutine
  holds `Handle.Update` open on the runner's own handle returns an error matching
  `ops.ErrCanceled` well before the busy timeout (the elapsed time is asserted below it).
  **Fixture state:** a current store with one running job record.
- `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Classify an unknown error as `not_found` → the "any other error" row fails.
2. Let the network actor accept a command-line identity → its refusal row fails.
3. Flip `start_backup` to member in the table → the literal row and the catalog row both fail.
4. Key the canceled wrapping on the error chain instead of `ctx.Err()` → report whether the
   write-pool row still passes, or name it reviewer-read.

**Prior art:** none; the classifier is the seam's own vocabulary.

**Deliverable count: 4** (the sentinels with the classifier and the code set, the network actor,
the table with `ops/local`'s catalog edit, the seam boundary's classification).

---

## Task 2: `internal/ops/opshttp`: the projection

Model: `opus`. Depends on Task 1. **Lands hook 9's projection half and hook 15's mount-channel
seam** (the constructor takes the mount's channel, which rung 14's cookie and bearer mounts pass).

**Files:**
- Create: `internal/ops/opshttp/doc.go`, `opshttp.go`, `authorize.go`, and their tests

**Outcomes:**

- **The authorizing layer** (R-6bii2, R-6bii3): an unexported `ops.Seam` decorator over an
  authorizer of the pinned shape, which looks up each operation's privilege, consults the
  authorizer, refuses a channel mismatch, discards its caller's actor, and calls through with the
  authorizer's actor only when admitted.
- **The HTTP layer** (R-6bii4): exact-path dispatch on the four names, POST only, the request
  refusals in their fixed order, the JSON error shape, and the status per code.
- **The constructor**, unexported, taking the seam, the authorizer, the mount's channel, and a
  logger; it fails on a nil seam, a nil authorizer, a nil logger, an invalid or command-line
  channel, and a code from `ops.Codes` with no status in its table.
- **The error-body discipline:** an `internal` answer's body carries no reason, a coded answer's
  reason is `(*ops.Error).Reason` alone, and the logged line carries the operation and the error.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Method | not POST | 405, `Allow: POST`, `invalid_request` |
| Path | not one of the four names, a trailing slash, or a subpath | 404, `not_found`; no authorizer call |
| Query string | any | 400, `invalid_request`; no authorizer call |
| Content-Type | absent, unparseable, or not `application/json` | 400, `invalid_request` |
| Body | empty | 400, `invalid_request` |
| Body | exactly 4096 bytes, one valid object | read and dispatched |
| Body | 4097 bytes | 400, `invalid_request` |
| Body | `null`, an array, a string, or a number | 400, `invalid_request`; no seam call |
| Body | trailing data after one object | 400, `invalid_request` |
| Body | an unknown key, `actor` and `ACTOR` among them | ignored |
| Handle | malformed | 400, `invalid_request` |
| Authorizer | denies | 403, `forbidden` (Ruling 3 yes); no seam call |
| Authorizer | member-only (admits `member`, denies `admin`), called for `start_backup` | 403, `forbidden`; no seam call |
| Authorizer | admits with `ops.LocalActor`, whose channel is command-line, on a cookie mount | 403, `forbidden`; no seam call |
| Authorizer | returns an error | 403, `forbidden`; the error logged |
| Layer | its caller passes a planted `ops.LocalActor` and the authorizer admits a network actor | the seam receives the authorizer's actor |
| Seam | returns each coded sentinel | its status and code, with `(*ops.Error).Reason` as the reason |
| Seam | returns a coded sentinel wrapped by a `fmt.Errorf` that names a path | its status and code; the path absent from the body |
| Seam | returns an uncoded error | 500, `internal`, no reason in the body |
| Constructor | a nil authorizer, seam, or logger, or a bad channel | fails, naming which |

**Acceptance criteria:**

- Each row above is its own test case through `httptest`, with a recording fake seam whose call
  count each refusal row asserts is zero. The recording fake records calls and the actor each call
  received, and never answers success on a path it was not configured for.
- The internal row asserts the planted error's text (it contains a path) is absent from the body
  and present in the captured log.
- `go list -deps ./internal/ops/opshttp` shows `internal/ops` and `internal/identity` as its only
  internal imports.
- Nothing in the package is exported; the exported-unused leg reports nothing for it.
- Gate green.

**Named mutations this task reports:**

1. Call through before the authorizer → the denial row's call count is 1.
2. Write the internal error's text into the body → the internal row fails.
3. Accept a GET for `job_status` → the method row fails.
4. Skip the channel comparison → the mismatch row fails.
5. Decode with `DisallowUnknownFields` → the actor row answers 400.
6. Forward the layer's caller's actor instead of the authorizer's → the planted-actor row fails.
7. Encode `err.Error()` for a coded error → the wrapped-path row fails.
8. Drop the first-token check → the `null` row dispatches.

**On Ruling 3's no**, the four authorizer rows answer exactly what `http.NotFound` answers (404,
its plain-text body) with no seam call. The mutations are unchanged.

**Prior art:** RFC 9457 (Problem Details for HTTP APIs) is consulted for the error body and
declined: its `type` member is a URI, and the seam's seven codes are the whole vocabulary a client
branches on, so a problem document adds a URI space nothing resolves. `go-chi/chi` is declined as
2.0's prior art already records (Navidrome's router, "architecture 2.0 chose against"); the
dispatch is four exact paths. The OpenSubsonic error shape stays declined (6b-i Task 2).

**Deliverable count: 4** (the authorizing layer, the HTTP layer, the constructor, the error-body
discipline).

---

## Task 3: The contract suite

Model: `opus`. Depends on Tasks 1 and 2. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/ops/opshttp/contract_test.go`, with a `TestMain` that re-executes the test
  binary as a second process (the shape of `internal/ops/local/process_test.go:40`) and installs a
  blocking rclone stand-in on `PATH`

**How the backup body blocks.** `ops/local.New` binds the real backup body (`local.go:76-83`), and
`bindBackup` and `open` are unexported, so the external contract suite cannot supply a body, and
spec `:604` forbids "a test-only constructor." The real body resolves rclone through
`exec.LookPath("rclone")` (`internal/backup/backup.go:69`), which is how 6b-i's
`internal/backup/standin_test.go` already drives `ops/local`. So `TestMain` puts a stand-in named
`rclone` first on `PATH`: it blocks until signaled and records each signal it receives. The
conflict, criterion 3, and criterion 4 rows run the real backup over a `local` remote with this
stand-in, and the re-executed child of criterion 4 inherits the same `PATH`.

**Outcomes:**

- **One case per sentinel**, each run twice: in-process against `ops/local` wrapped in the
  authorizing layer, and through the projection; each asserts both legs classify to the same code
  (spec criterion 1).
- **The job path over HTTP:** start, poll, cancel (criterion 3), and a cancel through the
  projection of a job running in another `ops/local` process (criterion 4).
- **A coverage guard over `ops.Codes`** that fails, naming the code, when a code has no case.

**States each check meets, and its report, with the fixture each needs:**

| Sentinel or criterion | Fixture state | Expected on both legs |
| --- | --- | --- |
| `unavailable` | a current store; backup flag off | `unavailable`, reason "backup integration disabled"; no job record |
| `not_found` | a current store; a well-formed handle with no record | `not_found` |
| `conflict` | a current store; flag on; a backup running against the blocking stand-in | `conflict` naming its handle |
| `invalid_request` | a malformed body on HTTP; a malformed handle in-process | `invalid_request` |
| `forbidden` | a denying authorizer; flag on | `forbidden`; `jobs` holds zero rows after |
| `forbidden`, by privilege | a member-only authorizer calling `start_backup`; flag on | `forbidden`; `jobs` holds zero rows after |
| `canceled` | a current store with one running job; a test goroutine holds `Handle.Update` open on the runner's own handle; the caller's context canceled while `cancel_job` waits for the write pool's one connection | `canceled` (Ruling 2 yes). The HTTP leg calls the handler's `ServeHTTP` with an `httptest.NewRecorder` and a request whose context the test cancels, and reads 503 from the recorder: over `httptest.Server` the client sees only its own context error, since "for outgoing client requests, the context controls cancellation" (`go doc net/http Request.Context`) |
| `internal` | a live runner whose store handle is then closed | `internal`; HTTP body without a reason. `database/sql` answers `sql: database is closed` (R-6bii5) |
| actor | `start_backup` with the body `{"actor":"local:root"}`, then `{"ACTOR":"local:root"}`, with an admitting authorizer | the `jobs.actor` column holds the authorizer's actor both times |
| criterion 3 | flag on; the blocking stand-in | started, polled `running`, canceled over HTTP, ends `canceled`, and the stand-in recorded one SIGINT |
| criterion 4 | a re-executed child running `ops/local` over the same store with a job running against the stand-in | the cancel over HTTP answers its state, the child's job ends `canceled`, and the child's stand-in recorded one SIGINT |

**Acceptance criteria:**

- Spec criteria 1 to 4, each by the rows above.
- The guard iterates `ops.Codes`, never a list the test hard-codes.
- Gate green, with the suite's wall time reported.

**Named mutations this task reports:**

1. Delete the `internal` case → the guard fails, naming `internal`.
2. Add an eighth sentinel to `ops` with no contract case → the guard fails, naming it.
3. Map `forbidden` to 404 in the projection → both legs disagree and the case fails. On Ruling
   3's no, 404 is the intended answer, so the mutation is the reverse: answer 403 with the JSON
   body → the forbidden case's 404 assertion fails.
4. Let the layer call through on denial → the `forbidden` row's `jobs` count is 1.

On Ruling 2's no, the `canceled` row does not exist and the guard counts six codes. On Ruling 3's
no, the HTTP leg of both `forbidden` rows asserts `http.NotFound`'s answer and the zero `jobs`
count, while the in-process leg still asserts `forbidden`.

**Prior art:** none.

**Deliverable count: 3** (the sentinel cases with the guard, the job path, the cross-process
cancel).

**Pre-named split point:** on Ruling 1's no, 6b-ii-a closes here.

---

## Task 4A: The store: the intent record, its refusals, and serve mode's `reconcile-pending`

Model: `opus`. Depends on Task 0. **Independent by files of Tasks 1 to 3.** **Neutral on 2.0
hooks** (hook 7's migration runner is reused, not changed).

**Files:**
- Create: `internal/engine/store/intent.go` and its test, with a re-executed child for the
  lock-wait race
- Modify: `internal/engine/store/open.go` (`leaf.stat`, `checkIntent`, `OpenWrite`, `OpenServe`,
  `settle`), `inspect.go`, `conn.go` (the read-pool multi-row query), and their tests
- Modify: `cmd/dubplate-server/store.go` (the restore-in-progress row renders the new reason) and
  its test

**Outcomes:**

- **The intent record** (R-6bii7): its type, its encode and strict decode, its four states, its
  basename fields, and its write through `atomicfile.WriteStream` at 0600.
- **The refusals per state, before and after the lock** in inspect, write mode, and serve mode, per
  R-6bii7's two tables: `OpenWrite` re-reads the record after `acquireShared`, and `OpenServe`'s
  shared path re-reads it after each shared `waitLock` and takes the exclusive path on any record.
- **Serve mode's `reconcile-pending` path** with `ServeOptions.Reconcile` (deviation D2), the
  report file write at 0600, and the record's removal with a leaf fsync.
- **The read-pool multi-row query** beside `Handle.QueryRow` (`conn.go:108-110`), for the
  reconciliation.

**States each check meets, and its report:** R-6bii7's two tables, each row its own test, plus:

| Check | State | Report |
| --- | --- | --- |
| Lock-wait race, write mode | a second process holds the exclusive lock; `OpenWrite` is waiting; the second process writes a `swap` record and exits | refused, naming `restore --resume` |
| Lock-wait race, serve mode | the same, with `OpenServe` waiting on its shared lock | refused, naming `restore --resume`, from `settle` on the exclusive path |
| Lock-wait race, serve mode | the same, with a `reconcile-pending` record and a current-version store | reconciles, writes the report, removes the record, and opens |
| Serve, `reconcile-pending`, store older | a pre-migration copy is taken, the store migrates, the reconciler runs, the report lands, the record goes | opens current |
| Serve, `reconcile-pending`, store current | a crash after migrating left the record | reconciles and removes the record without migrating |
| Serve, `reconcile-pending`, nil reconciler | | refused, naming the missing reconciler; the record stays |
| Serve, `reconcile-pending`, reconciler fails | | refused, naming the error; the record stays; no report file |
| Serve, `reconcile-pending`, a leftover in `staging/` | | removed before the migration |
| Serve, `swap` record and a staged incoming file | a serve open, then a write open | both refuse; the staged file is still present, listed by name |
| Report write | the report file already exists at that stamp | replaced atomically; mode 0600 |

The lock-wait fixture is the mechanics lens's probe, made a gate test: the child takes `LOCK_EX`
on `store.lock`, signals readiness, and waits; the parent starts the open in a goroutine and waits
until the open is past `stat`; the child then writes the record through the record's own writer
and exits, releasing the lock. The test asserts the refusal text. "Past `stat`" needs a real
signal, since an open that has not reached `stat` would refuse on the pre-lock read and pass the
row under mutation 5. The serve rows have one: `waitLock` logs the holders on its first busy
attempt (`open.go:312-313`), so they pass a recording logger and wait for that line. Write mode's
`acquireBounded` (`lock.go:63-78`) logs nothing, so the write-mode row waits a settle delay
instead, and the task reports mutation 5's kill rate across repeated `-count` runs. If the
mutation survives any run, the row is named a reviewer-read.

**Acceptance criteria:**

- Spec criterion 11: an older-version store with a `reconcile-pending` record; a write-mode open
  refuses naming "start serve"; serve mode migrates it, calls a recording reconciler once with a
  handle whose version is the binary's, writes the report, and removes the record; a write-mode
  open then succeeds. **Fixture state:** a store at version 2, made through the migration runner's
  `fs.FS` parameter (`migrate.go:91`) with the real set's first two files (`0001_ledger.sql`,
  `0002_jobs.sql`), so serve migrates the real `0003_catalog.sql`; it holds one populated job row
  that must survive, which version 2's `jobs` table carries.
- Every intent-record row of R-6bii7's second table, by its own fixture file.
- The `doctor` and `check` render of each record state, in `cmd/dubplate-server/store.go`'s
  existing golden shape.
- Gate green.

**Named mutations this task reports:**

1. Let serve mode accept every record state → the `swap` row opens instead of refusing.
2. Remove the record before the report lands → the reconciler-fails row leaves no record.
3. Decode the record without `DisallowUnknownFields` → the unknown-key row reads as valid.
4. Clear staging while a `swap` record exists → the serve-then-write row's staged file is gone.
5. Delete `OpenWrite`'s re-read after `acquireShared` → the write-mode lock-wait row opens.
6. Let `OpenServe`'s shared path return a handle beside a record → the serve lock-wait rows open
   without refusing or reconciling.

**Prior art:** `internal/atomicfile`, adopted for the record and the report (spec prior art).
SQLite's own backup API (`conn.go:1208-1219` in modernc v1.59.0, as spec `:622-623` records) stays
declined for restore: the swap moves a verified file, and the API would copy into a live store.

**Deliverable count: 4** (the record, the refusals before and after the lock, serve's path, the
read query).

---

## Task 4B: The store: the restore session, staging, and verification

Model: `opus`. Depends on Task 4A. **Neutral on 2.0 hooks** (hook 10's consistency is backup's;
restore consumes its output).

**Files:**
- Create: `internal/engine/store/restore.go` and its test, with a re-executed child for the
  lock-holder fixtures
- Modify: `internal/engine/store/snapshot.go` (the verifier's version parameter; on Ruling 5's
  yes, `WriteBackupRef`'s body moved into the shared leaf-level function) and its test

**Outcomes:**

- **The session open** (R-6bii6): the uid check, leaf creation when absent (with 6a's E9
  state-directory owner rule, `leaf.go:70-99`), the exclusive lock taken non-blocking, the
  filesystem probe, the one-device check, the refusal of a new restore beside a record, and the
  staging rule. A leaf with a lost database beside a copy or a reference opens: restore is the
  remedy `besideCopy` names (`open.go:487-500`).
- **Staging and verification** (R-6bii9): a pre-created 0600 stage for `--from-file`, a local-file
  copy with sync, the stage chmod for a stage rclone writes, the verifier's version parameter, the
  ledger check, the removal of a verified stage's `-wal` and `-shm`, and the wrong-archive rows of
  R-6bii9 that the store can decide.
- **The session's handles** (R-6bii6): close-on-exec duplicates of the exclusive lock's descriptor
  for a transfer child, the lockless read handle on the restored store, and, on Ruling 5's yes
  only, the reference write sharing `WriteBackupRef`'s body. On Ruling 5's no, the reference write
  and its row do not exist, since the method would be an export with no caller (exit gate 14).

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Lock | held shared by a live write-mode `Handle` in a second process | refused, "held by", no file changed |
| Lock | held shared by an inspect-mode open in a second process (the "real reader" `doctor` and `check` use) | refused the same way |
| Lock | held by a descriptor a child inherited from a killed parent | refused the same way |
| Leaf | absent; state directory owned by the caller | created |
| Leaf | wrong owner | refused, naming the owner and `runuser` |
| Record | present, with `--snapshot` or `--from-file` | refused, naming `restore --resume`; nothing changed |
| Record | present, with `--resume` | opens |
| Devices | `db/` a symlink to a directory under `/dev/shm`, the leaf under the test's temp directory | refused, naming each directory and its device, before any stage |
| Staging | no record; leftovers including a planted `<name>.<hex>.partial` | every entry removed |
| Stage | created by the test at 0644, as rclone leaves it | chmodded to 0600 before any open |
| Stage | a byte copy of a cleanly closed WAL-header store | verified; its `-wal` and `-shm` removed after the verification |
| Stage | each wrong-archive row R-6bii9 gives the store (truncated, zero-byte, unmarked, newer, ledger mismatch) | refused with that row's report, before any record |
| Verifier | backup's call with a copy one version below the binary's | refused, exact equality kept |
| Read handle | opened on the restored store while the session holds `LOCK_EX` | reads without waiting on the lock |
| Reference write (Ruling 5 yes) | twice, with two stamps | one reference, naming the second stamp |

The device row's test asserts first that the two directories' devices differ and fails, naming
both, when they do not; the fold measured `/tmp` and `/dev/shm` on this workstation as separate
tmpfs mounts with devices 60 and 27. The device comparison itself is a helper over device numbers
with its own table test, so the integration row is not the check's only proof.

**Acceptance criteria:**

- Spec criterion 5: a second process holds a write-mode `Handle`, and in a second case an
  inspect-mode open; the session open refuses and the database file's bytes and mode are
  unchanged. **Fixture state:** a current store with one job row.
- Spec criterion 9's store half: a newer, a corrupt (a truncated copy), a zero-byte, and an
  unmarked SQLite file are each refused before any record is written.
- Spec criterion 13's mode half on the `--from-file` path: the stage is 0600 per `os.Stat`. The uid
  half cannot fail under the test's own uid; it is declared a reviewer-read, since the session
  never chowns and the leaf-owner check refuses any other uid.
- Gate green.

**Named mutations this task reports:**

1. Take the lock blocking → criterion 5's test hangs to its deadline instead of refusing (report
   the timing).
2. Skip the stage chmod → the 0644 row fails.
3. Let a new restore proceed beside a record → the record row overwrites it.
4. Drop the device check → the device row stages a file.
5. Call the verifier with exact equality on the restore path → an older stage is refused.

**Deliverable count: 3** (the session open, staging and verification, the session's handles).

---

## Task 4C: The store: the quiesce, the swap, the discard, and resume's combinations

Model: `opus`. Depends on Task 4B. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/engine/store/swap.go` and its test, with a re-executed child for the
  crash-left-WAL and foreign-connection fixtures
- Modify: `internal/engine/store/restore.go` (the session's quiesce and swap methods) and its test

**Outcomes:**

- **The quiesce** (R-6bii8): step 3 with its three refusals (a nonzero busy column, an
  `SQLITE_BUSY` from the switch, a mode other than `delete`) and the leftover-file net.
- **The swap and the discard** (R-6bii8, deviation D8): step 5 as two session calls, the first
  holding the pre-restore rename or the discard's moves after the record, the second holding the
  incoming rename and the fsyncs. The store declares no hook; Task 5B's flow calls the hook
  between the two calls, and this task's combination rows stay hand-built fixtures.
- **Resume's combination rule**: R-6bii8's table, keyed on what the record owes, with the version
  re-check that opens every `swap` resume (R-6bii7).

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Current store | absent | no quiesce; the record names no pre-restore copy |
| Current store | opens; WAL holds rows | checkpointed, switched to DELETE, no `-wal` or `-shm` remains |
| Current store | a second process holds a read transaction | busy column 1; refused, naming the database and "another process may have the database open" |
| Current store | a second process holds an idle connection | the switch answers `SQLITE_BUSY`; refused the same way |
| Current store | already in DELETE mode, with a planted `-shm` | refused, naming the leftover |
| Current store | newer than the binary | quiesced and swapped (the rollback path) |
| Current store | cannot open | refused, naming `--discard-current` |
| Current store | cannot open, with `--discard-current` | after the record, the database, `-wal`, and `-shm` moved to the `discarded-<stamp>` names, each named |
| Swap | each row of R-6bii8's combination table | as the table states |
| Resume, `swap` | the staged file above the running binary's version | refused, naming both versions |
| Swapped-in file | after the swap | owned by the leaf owner at 0600; reads WAL after a write-mode open |

**Acceptance criteria:**

- Spec criterion 6: a re-executed child commits three rows through `Handle.Update` and is
  SIGKILLed. Before the restore, the test asserts `-wal` is non-empty and that a byte copy of the
  bare `db/dubplate.db`, opened read-only in a temp directory without its `-wal`, lacks the three
  rows, so they live only in the WAL. The swap then moves one file, `-wal` and `-shm` are absent,
  and the pre-restore copy, opened read-only, holds the three rows.
- Spec criterion 9's current-store half: an unopenable current store is refused without
  `--discard-current`.
- Spec criterion 10: a current store at the binary's version plus one (the synthetic migration
  set), a `pre-migrate-v<N>.db` at the binary's version; the restore from that file swaps, and a
  write-mode open of the result succeeds under this binary.
- Every combination row by its own hand-built fixture, including both no-copy rows and both
  discard rows.
- Gate green.

**Named mutations this task reports:**

1. Skip step 3 and the leftover check → criterion 6's pre-restore copy lacks the rows and `-wal`
   remains in `db/`. (Skipping the explicit checkpoint alone is not a killing mutation: leaving WAL
   mode checkpoints on the way out, per the `wal.html` quote. The checkpoint stays for its busy
   column's report.)
2. Skip the version re-check on a `swap` resume → the resume-version row swaps a newer file in.
3. Rename the incoming file first → criterion 6's pre-restore copy holds the incoming rows instead
   of the three.
4. Drop the no-copy branch → both no-copy combination rows refuse.
5. Discard before the record → a step-hook crash after the discard leaves no record (report the
   leaf state it leaves).
6. Accept a `delete` switch that answers another mode → report which fixture, if any, reaches it,
   or name it reviewer-read.

**Deliverable count: 3** (the quiesce, the swap with the discard, resume's combination rule).

---

## Task 5A: `internal/backup`: the one rclone launch path and the fetch

Model: `opus`. Depends on Task 4C. **Neutral on 2.0 hooks.**

Fold 2 split revision 2's Task 5 here, because absorbing the architecture read's launch fix took
it past four deliverables (verification m2 and m10). 5A is the remote side; 5B is the flow.

**Files:**
- Create: `internal/backup/fetch.go` and its test
- Modify: `internal/backup/rclone.go` (the transport takes a stdout writer and any source of
  inherited descriptors, R-6bii6), `prune.go` (`prune` holds a transport, `list` launches through
  it, `runRclone` is deleted, and the listing is shared with the fetch), and their tests

**Outcomes:**

- **The one launch path** (R-6bii6): `transport.run` with a stdout writer and any source of
  inherited descriptors; `prune.list`, `deletefile`, and `purge` launched through it; `runRclone`
  deleted. 6b-i's backup and prune tests pass unchanged, which is the proof that backup's behavior
  did not move.
- **The listing and its rules** (R-6bii9, R-6bii10): the `snapshots/` listing first, `latest`
  resolution, the absent-object and `--older` rules, the remote-mismatch warning, the object's
  `Size` for the store's free-space check, and `rclone size --json` for the library's.
- **The fetch** (R-6bii9): `copyto` into the stage path with no stage pre-created, the chmod to
  0600, and the stage's removal after a failed or canceled fetch.

**States each check meets, and its report:** R-6bii9's table rows the listing and the fetch decide,
R-6bii10's rows, and:

| Check | State | Report |
| --- | --- | --- |
| Listing | canceled while the blocking stand-in runs | the stand-in records one SIGINT to its own process group; the error wraps `ctx.Err()` |
| Listing | names out of order, one name that does not parse as a stamp | `latest` is the greatest parsed stamp; the unparseable name is ignored |
| Fetch | a stage rclone leaves at 0644 under umask 022 | 0600 per `os.Stat` before any open |
| Size | an object whose `Size` is -1 | the store check skipped, with a printed line naming the object |
| Size | the measured `{"count":2,"bytes":20004096,"sizeless":0}` shape | the floor and the `sizeless` count parsed |

**Acceptance criteria:**

- Spec criterion 9's remote half: an absent stamp, `latest` on an empty remote, and an older stamp
  without `--older` are each refused against a real `rclone` `local` remote; the refusal names
  the remote path or the latest stamp, and the absent-stamp refusal makes no `copyto` call.
- `git grep -n 'exec.Command' -- internal/backup ':!*_test.go'` finds one launch site, in
  `rclone.go`.
- Gate green, with the new tests' wall time.

**Named mutations this task reports:**

1. Resolve `latest` by listing order instead of parsed stamp → the out-of-order row picks the wrong
   one.
2. Build a destination by string concatenation → the empty-path remote row (`name:`) fetches from
   `name:/snapshots` and fails.
3. Launch the listing with a bare `exec.CommandContext` again → the canceled-listing row records no
   SIGINT, and its error does not wrap `ctx.Err()`.
4. Skip the stage chmod → the 0644 row fails.

**Prior art:** rclone as a probed subprocess, adopted (spec).

**Deliverable count: 3** (the launch path, the listing and its rules, the fetch).

---

## Task 5B: `internal/backup`: the restore flow, the step hook, resume, and the reference

Model: `opus`. Depends on Task 5A. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/backup/restore.go` and its test, with a re-executed child for the step-hook
  crashes

**Outcomes:**

- **The flow**: spec steps 1 to 8's order, with the step hook, an unexported function variable in
  `internal/backup` called after each step and between the session's two step 5 calls (R-6bii8,
  spec `:283-284`), and the record advanced at each boundary R-6bii7 names.
- **Resume**: reads the record, re-derives each version verdict from the running binary through
  the verify-then-remove session method (R-6bii7), and finishes from its state, never re-fetching
  after step 4.
- **The reference rewrite** on Ruling 5's yes: after the swap and before the record leaves `swap`,
  through the session's reference write, redone by every resume from `swap`. On Ruling 5's no,
  this outcome and the two reference crash rows do not exist, the "after step 5" row completes to
  the next state, and the task carries two deliverables.

**States each check meets, and its report:**

| Crash point (step hook exits the child) | Record after | Write-mode open | `--resume` |
| --- | --- | --- | --- |
| after step 1 or 2 (a current store present) | absent | proceeds (the stage is a leftover) | refuses: no record; a new restore clears staging |
| after step 3 (a current store present) | absent | proceeds; the store, left in DELETE mode, reads WAL after the open | refuses: no record |
| after step 4 | `swap` | refuses naming `--resume` | completes the swap |
| between step 5's two calls | `swap` | refuses | the third combination row (the second rename, then the fsyncs) |
| between step 5's two calls, with `--discard-current` | `swap` | refuses | the first discard row, then completes with the discard's completion tuple |
| after step 5, before the reference write | `swap` | refuses | the fsyncs, the reference, the next state |
| after the reference write | `swap` | refuses | completes; one reference names the restored stamp |
| after step 4, 5's first call, or 5, on a fresh leaf (no current store) | `swap` | refuses | the matching no-copy row, then completes |
| after step 4, then a binary whose version is below the staged file's | `swap` | refuses | refuses, naming both versions |
| after step 5 with `--library` | `library` | refuses | re-runs the copy |
| after step 6 (or 5 without `--library`) | `reconcile` or `reconcile-pending` | refuses | completes, or names "start serve" |
| `reconcile`, then a binary one version newer (the synthetic migration set) | `reconcile` | refuses | advances to `reconcile-pending` and names "start serve" |
| after step 7's report, before removal | `reconcile` | refuses | rewrites the report, removes the record |
| rclone interrupted mid-fetch by SIGINT to the child | absent | proceeds | `staging/` already empty: rclone removed its `.partial` (measured) |
| rclone SIGKILLed mid-fetch | absent | proceeds | a new restore clears the `.partial` leftover |

Each row in which `--resume` completes asserts R-6bii7's completion tuple.

**Acceptance criteria:**

- Spec criterion 7, as deviation D7 states it: each crash row above, a real process death through
  the step hook, followed by a write-mode open refusing where the table says and `--resume`
  reaching the completion tuple.
- Spec criterion 8: a crash after step 4, then a write-mode open, leaves `staging/`'s incoming
  file in place (listed by name), and `--resume` completes the swap. **Fixture state:** a current
  store with rows, a verified stage, the `swap` record.
- Spec criterion 13 on the `--snapshot` path: a restore through real rclone under umask 022 leaves
  `db/dubplate.db` at 0600 per `os.Stat`. Mutation: skip the chmod, and the mode reads 0644, as
  the mechanics lens measured.
- `staging/` holds no `-wal` or `-shm` after a `swap` resume whose version re-check opened the
  stage.
- Gate green, with the new tests' wall time.

**Named mutations this task reports:**

1. Advance the record before the swap's fsyncs → report whether the step-5 crash row still resumes,
   or name it reviewer-read.
2. Let resume re-fetch after step 4 → the resume row records an rclone call on the stand-in.
3. Trust the recorded version on resume → the newer-binary `reconcile` row reconciles an
   unmigrated store.
4. Skip the chmod on the `--snapshot` path → the criterion 13 row reads 0644.
5. Swap before verifying → R-6bii9's truncated-file row reaches a `swap` record. The flow, not
   the store, owns the order of verify and swap, so this mutation is Task 5B's.
6. Re-check the version on resume by a plain inspect open → the `staging/` criterion finds a
   `-wal` or `-shm`.

**Prior art:** the salvage restore (`_salvage/internal/backup`, spec prior art row) is read for
its flow order only; nothing imported.

**Deliverable count: 3** (the flow with the hook, resume, the reference rewrite).

---

## Task 6: `internal/backup`: the library half and the reconciliation report

Model: `opus`. Depends on Task 5B. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/backup/reconcile.go` and its test
- Modify: `internal/backup/restore.go` (step 6 and its resume) and its test

**Outcomes:**

- **Step 6** (R-6bii11): the remote `library/` presence check, the root's resolution, the
  absent-root and non-empty refusals (with Ruling 8's `lost+found` exemption on yes), the
  mount-point line (Ruling 7's yes), `rclone copy` under the transport (with Ruling 6's
  `--backup-dir` on yes, with `--into-nonempty` only, under R-6bii11's three displaced-directory
  rules), and resume's re-resolution without the emptiness check (deviation D9).
- **The reconciliation** (R-6bii11) as an exported function over a handle and a root, returning
  a report with its rendered text: the join, the `Lstat`, and the walk.
- **The report's two renders**: the file body (every entry) and the stdout summary (counts and
  the first 20).

**States each check meets, and its report:** R-6bii11's table, each row its own test, and:

| Check | State | Report |
| --- | --- | --- |
| Remote | no `library/` under the remote, with `--library` | refused, naming the remote path; no size or copy call |
| Library root | absent | refused, naming the path; no rclone call |
| Library root | a symlink to an empty directory | resolved, then copied into |
| Library root | an empty directory on its parent's device | copied into; the printed line says it is not a mount point (Ruling 7 yes). On Ruling 7's no: refused naming `--root-not-a-mount`, and copied into with it |
| Library root | holding only `lost+found` | copied into on Ruling 8's yes; refused naming the flag on no |
| Library root | non-empty, no `--into-nonempty` | refused, naming the flag; no rclone call |
| Library root | non-empty, with `--into-nonempty`, a local-only file | copied; the local-only file kept and reported as a file without a row |
| Library root | non-empty, with `--into-nonempty`, a local file differing from the remote's | on Ruling 6's yes, the remote's version in the root and the local version under the displaced directory, named in the report; on no, the remote's version in the root and the local version gone |
| Copy | rclone fails | the record stays `library`; the verb exits 1 naming rclone's exit |
| Resume | a stand-in that wrote one file of two was SIGKILLed, the record `library` | `--resume` skips the emptiness check, completes the copy, and the tree matches the remote by relative path and bytes |
| Resume | the recorded root now absent, or resolving elsewhere | refused, naming the recorded and the current root |
| Resume, `--into-nonempty` (Ruling 6 yes) | a hand-built interrupted state: the record `library` with `--into-nonempty`, the displaced directory at 0700 holding one displaced file, the root holding one of the remote's two files and one differing local file | `--resume` accepts the directory, completes the copy with real rclone, keeps the first run's displaced file, adds the second, and the report names the directory with a count of 2 |
| Resume, `--into-nonempty` (Ruling 6 yes) | the displaced path exists as a symlink, a file, at 0755, or under another uid | refused, naming the path and what was found; no rclone call |
| Displaced directory (Ruling 6 yes) | `--into-nonempty`, every local file identical to the remote's or local-only | removed after the copy; the report names none |
| Displaced directory (Ruling 6 yes) | a first run whose path already exists | refused before any rclone call, naming the path |

**Acceptance criteria:**

- Spec criterion 12: a tree restore from a real `rclone` `local` remote into an empty root yields
  every backed-up file, compared by relative path and bytes; into a non-empty root without
  `--into-nonempty` it refuses naming the flag; and the report names a planted track row whose file
  is absent and a planted file with no row, each once. **Fixture state:** a skeleton with one
  collection, one album, and two tracks, one of whose files is deleted after the copy; one extra
  file planted.
- The empty-`users` line appears in a report over a store whose `users` table is empty.
- The report file is 0600 per `os.Stat`.
- Gate green.

**Named mutations this task reports:**

1. Use `rclone sync` instead of `copy` → the local-only-file row loses its file.
2. Follow symlinks in the walk → report which row changes, or name it reviewer-read.
3. Compare paths as strings through a UTF-8 conversion → a fixture file with a non-UTF-8 name is
   misreported.
4. Re-run the emptiness check on resume → the resume row refuses its own partial tree.
5. On Ruling 6's yes, drop `--backup-dir` → the differing-file row finds no displaced copy.
6. On Ruling 6's yes, create the displaced directory with `os.Mkdir` on resume too → the
   `--into-nonempty` resume row refuses its own directory.
7. On Ruling 6's yes, remove the displaced directory with `os.RemoveAll` → the resume row loses
   the first run's displaced file.

**Deliverable count: 3** (step 6 with its resume, the reconciliation, the renders).

---

## Task 7: The `restore` verb

Model: `opus`. Depends on Tasks 5B and 6. **Neutral on 2.0 hooks.**

**Files:**
- Create: `cmd/dubplate-server/restore.go` and its test
- Modify: `cmd/dubplate-server/main.go` (one `AddCommand` entry in `newRootCmd`) and its test

**Outcomes:**

- **The flags** (spec `:676-686`) with their mutual exclusions: exactly one of `--snapshot`,
  `--from-file`, or `--resume`; `--older` only with `--snapshot`; `--library` never with
  `--older` or `--from-file`; `--into-nonempty` only with `--library`. On Ruling 7's no, the flags
  add `--root-not-a-mount`, only with `--library` (R-6bii11), and spec criterion 12's end-to-end
  empty-root fixture passes it, since the test's root sits on its parent's device.
- **The exit codes** (Ruling 4) through the existing `exitCodeErr` path.
- **Signals** (R-6bii12): SIGINT, SIGTERM, and SIGHUP cancel, leave the record at its last written
  state, and exit 130.
- **The output**: the free-space lines with the `sizeless` count, the remote-mismatch warning, the
  mount-point line (Ruling 7's yes), the reconciliation summary, the report file's path, and ruling
  12's rotation instruction.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Flags | none of the three sources, or two | usage error naming them; exit 1 |
| Flags | each forbidden combination | refused naming both flags; exit 1 |
| Config | the config fails to load | exit 1, the load error |
| Lock | held | exit 75, naming the holders |
| Record | present, with `--snapshot` | exit 1, naming `restore --resume` |
| Signal | SIGINT during the library copy | exit 130; record `library`; rclone ended |
| Signal | SIGTERM during the `snapshots/` listing, the stand-in blocking | exit 130; no record; `staging/` empty; the stand-in recorded one SIGINT from the transport |
| Signal, then resume | SIGINT during the library copy with the blocking stand-in first on `PATH`, then `restore --resume` with `PATH` switched to put the real rclone first | exit 0; the tree matches the remote |
| Success | reconcile path | exit 0; the summary, the report path, the rotation instruction |
| Success | `reconcile-pending` path | exit 0; "start serve to finish the restore"; the rotation instruction |

**Acceptance criteria:**

- An end-to-end run of the built verb against a real `rclone` `local` remote holding a backup the
  6b-i `backup` verb wrote restores the store and library, exits 0, and prints the rotation
  instruction naming both secret keys. **Fixture state:** a current store with two tracks and a
  backup run to completion; then one row added and one row changed after the backup, and the
  library root emptied. The test asserts by value that the added row is absent, the changed row
  holds its backed-up value, and every backed-up file is present by relative path and bytes.
- The signal rows with a stand-in rclone that blocks, as 6b-i's `standin_test.go` does, started
  with `Setpgid`.
- The verb's help lists the four exit codes.
- Gate green.

**Named mutations this task reports:**

1. Let a cancel during step 6 advance the record → the signal row's record reads `reconcile`, not
   `library`.
2. Allow `--older` with `--library` → the combination row passes.
3. Omit SIGHUP → report whether the stand-in survives the verb, or name it reviewer-read.

**Deliverable count: 4** (the flags, the exit codes, the signals, the output).

On Ruling 9's no, Task 7B adds `--abandon` (its flag, its session method, its refusals, and its
state rows), `opus`, depending on Task 7, with its own four deliverables and about 0.5M.

---

## Task 8: The pass close, and the 6b-ii to 6c and 6d handoff

Model: `sonnet`. Depends on Tasks 0 to 7 and the simplifier round. On Ruling 1's no, each half
runs its own Task 8 over its own tasks, and 6b-ii-a's handoff names 6b-ii-b as the next pass.
Each half writes drafts suffixed with its own name, so the second half never overwrites the
first's.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/superpowers/research/handoffs/draft-STATUS-6b-ii.md` (on a split,
  `draft-STATUS-6b-ii-a.md` or `draft-STATUS-6b-ii-b.md`)
- Create: `docs/superpowers/research/handoffs/draft-HISTORY-6b-ii.md` (on a split,
  `draft-HISTORY-6b-ii-a.md` or `draft-HISTORY-6b-ii-b.md`)

**Outcomes:**

- **The HISTORY draft**: what landed per task with commit ids, the gate evidence with `-count=1`
  wall times, R-6bii1 to R-6bii12 with where each landed, the nine rulings' outcomes, the declared
  coverage gaps, both budgets, and "what a later pass would be wrong to rediscover," carrying at
  minimum the step 3 measurements (the switch removes `-shm` on the only connection; an idle
  foreign connection fails the switch at once), the mode `copyto` writes, the combination table,
  the under-lock record re-read, and the `canceled` meaning.
- **The errata owed**, paste-ready: deviations D1 to D10, each ruling-dependent erratum the fold
  record lists, and the spec's "Not verified" rows this pass settles.
- **The handoff**: to 6d, `ServeOptions.Reconcile` and `backup.Reconcile` as `serve`'s wiring, the
  store identity that would let restore refuse another store's snapshot, and the network actor's
  `member:<principal>` form against `Identity.UserID`; to rung 14, the projection's mount, the
  production authorizer, the bearer channel, the answer to a caller with no identity (Ruling 3),
  and, before the cookie mount lands, the 5c-ii security M2 question for the projection too (the
  roster API shares an origin with the Navidrome fallthrough, and `originCheck` admits
  `Sec-Fetch-Site: same-origin`, `internal/serve/origin.go:13-24`), with the bearer mount going
  behind `addBearerBypass` only; to rung 12, the streamer database's place in a restore.
- **The runbook lines owed to rung 16**: restore needs the server stopped; `--older` restores the
  store only; a restore before rung 12 loses stars and playlists; the rotation step; the report
  file's location; `--discard-current`'s `discarded-<stamp>` files, which nothing prunes; the next
  successful backup deletes the local pre-restore copy (`pruneLocal`, `prune.go:168-200`), which
  closes the local undo window; mount the library volume before a library restore (Ruling 7); an
  interrupted restore is finished with `restore --resume`, or by starting serve when the verb said
  so; on Ruling 6's yes, the displaced directory beside the root, which nothing removes while it
  holds files; a real rclone killed by SIGKILL during the library copy leaves
  `<name>.<hex>.partial` files under the root, which the report lists as files without a row and
  the operator removes by hand; and, on Ruling 9's yes, the manual procedure for a record resume
  refuses.

**Acceptance criteria:** every count in the drafts was produced by a command the report pastes;
every commit id resolves; the STATUS draft is at most 60 lines. Gate green as a formality.

**Deliverable count: 4** (the HISTORY draft, the errata, the handoff with the runbook lines, the
STATUS draft).

---

## The exit gate

The gate green in the worktree and again on master after the merge, plus every assertion below.
On Ruling 1's no, 6b-ii-a's gate is items 1, 10, 12, and 13, with items 9 and 14 over its own
tasks and exports; 6b-ii-b's is items 2 to 8, 11, 12, and 13, with items 9 and 14 over its own
tasks and exports. Item 10 is 6b-ii-a's alone, since the import walls it checks are that half's
surface.

1. Spec criteria 1 to 4: Task 3 (with Task 2's rows).
2. Spec criterion 5: Tasks 4B and 7.
3. Spec criteria 6 and 10: Task 4C.
4. Spec criteria 7 and 8: Task 5B, with criterion 7 as deviation D7 states it.
5. Spec criterion 9: Tasks 4B, 4C, and 5A.
6. Spec criterion 11: Task 4A.
7. Spec criterion 12: Task 6.
8. Spec criterion 13: Task 4B (the `--from-file` stage), Task 4C (the swapped-in file), and Task 5B
   (the `--snapshot` path), the uid half declared a reviewer-read.
9. Every state row has a named test: R-6bii7's three tables, R-6bii8's combination table, R-6bii9's
   source table, R-6bii11's report table, and the state tables of Tasks 1, 2, 3, 4A, 4B, 4C, 5A,
   5B, 6, and 7.
10. `internal/ops/opshttp` imports only `internal/ops` and `internal/identity` among internal
    packages; `internal/ops` imports only `internal/identity`.
11. `internal/engine/store` imports no `net/http`.
12. Nothing in the pass creates or migrates a store outside a test, except `restore` creating a
    leaf and placing a verified snapshot.
13. No HTTP route is mounted: `git grep -n opshttp -- ':!internal/ops/opshttp'` finds nothing but
    docs.
14. Every symbol the pass exports is listed with a `git grep` showing its production caller
    outside its own package, pasted, since the exported-unused leg passes a symbol referenced from
    any non-test file in its own package. The named exceptions are exactly these: the network
    actor's constructor (rung 14), `ServeOptions.Reconcile` and `backup.Reconcile` (6d), and serve
    mode's open (6b-i's exception, unchanged). `ops.Codes` and the classifier are listed with their
    caller in `opshttp`'s non-test code and the note that no binary reaches it until rung 14 mounts
    the projection; the gate does not claim they are reachable.

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's ownership table before the lane
launches.

| Surface | Owner | 6b-ii's position | Contention |
| --- | --- | --- | --- |
| `internal/ops`, `internal/ops/local`, `internal/backup`, `internal/engine/store` | rung 6 | Tasks 1, 4A to 6 (5A and 5B included) | none |
| `internal/ops/opshttp` (new) | rung 6 | Tasks 2, 3 | none |
| `internal/identity` | rung 5 | read only; imported by `internal/ops` | lane 1 may edit it; Task 0 item 5 |
| `cmd/dubplate-server/main.go` | rung 1 | Task 7: one `AddCommand` entry | lane 1 edits `newProductionRegistry`; a textual conflict keeps both in landing order, and the second lander's rebase is non-fast-forward |
| `cmd/dubplate-server/store.go`, `restore.go` (new) | rung 6 | Tasks 4A, 7 | none |
| `cmd/dubplate-server/doctor.go`, `check.go` | rung 1's skeleton | untouched | none |
| `scripts/check.sh`, `Makefile`, `web/` | rung 5 | untouched | after 5c-ii merges, the rebase needs `npm ci --prefix web` |
| The state directory layout | rung 4 | the leaf gains `restore-intent` (format), `restore-report-<stamp>.txt`, and `snapshots/discarded-<stamp>.db` with its siblings | 6d grants the leaf on the box |
| The library root's parent | the operator's layout | on Ruling 6's yes, restore creates `<root>.restore-displaced-<stamp>` beside the root | 6d grants the path on the box |
| `docs/` ledgers | conductor | Task 8 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: confirm 6b-i merged and its architecture reads' findings are routed; record the
   planning spend against the rung; cut the lane 2 worktree from master; write the lane registry
   row and the contested-surface rows; arm the runaway guard; check AC power.
2. Before Task 1: rule Task 0's blockers and record Rulings 1 to 9.
3. At segment A's boundary (after Task 3): STATUS; re-derive the ceiling. On Ruling 1's no, 6b-ii-a
   runs its close here (actions 5 to 8 over Tasks 1 to 3, with the security review and three
   architecture reads) and 6b-ii-b starts as a fresh session from the STATUS resume prompt. On yes,
   take the split here anyway when the re-derivation projects past the ceiling.
4. At segment B's boundary (after Task 4C) and segment C's (after Task 6): STATUS; re-derive.
5. After Task 7: one `code-simplifier` round over the pass's accepted Go, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, its review chunked one `diff-reviewer` per
   package.
6. The pass-end fan-out: `web-auth-security-reviewer` over `internal/ops/opshttp` and the
   `internal/ops` actor and classifier, briefed that the projection is unmounted and that rung 5's
   admin gate is the comparison surface; and the spec's one targeted fresh-context `opus` read of
   the data-loss paths (the migration runner, the lock procedure, the snapshot, and restore),
   briefed with the spec fold record's risk model and this plan's fold record. Each dispatch asks
   for correctness gaps ranked by consequence, not polish; a fold may refuse a finding whose fix
   costs more than the risk it removes, with a one-line reason.
7. The merge ritual, with one `diff-reviewer` on the merged result whenever the rebase is not a
   fast-forward, then **one `go-architecture-reader` per touched Go package**
   (`internal/ops`, `internal/ops/local`, `internal/ops/opshttp`, `internal/engine/store`,
   `internal/backup`, `cmd/dubplate-server`; on a split, each half reads its own three), no plan in
   context.
8. Record 6b-ii's ceiling against the actual, and re-derive 6c's figure before its plan is drafted.
9. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task (`dubplate-implementer`, then `diff-reviewer` at
`claude-opus-5-5`, then the gate), **sequential**, in segments of at most four tasks, each
boundary on a task's own green committed gate. One re-dispatch on `fix`; a second `fix` is the
conductor's decision. Every reviewer dispatch asks for correctness gaps ranked by consequence.

**One pass (Ruling 1 yes):**

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 0, 1, 2, 3 | STATUS; re-derive |
| **B** | 4A, 4B, 4C | STATUS; re-derive |
| **C** | 5A, 5B, 6 | STATUS; re-derive |
| **D** | 7, then the simplifier round, the fan-out, Task 8 | the close drafts, the merge ritual, the architecture reads, the ledger fold |

**Split (Ruling 1 no):**

| Pass | Segment | Tasks | The boundary's own work |
| --- | --- | --- | --- |
| 6b-ii-a | **A1** | 0, 1, 2, 3 | STATUS; re-derive |
| 6b-ii-a | **A2** | the simplifier round, the security review, Task 8 | the merge ritual, three architecture reads, the ledger fold |
| 6b-ii-b | **B1** | 0, 4A, 4B, 4C | STATUS; re-derive |
| 6b-ii-b | **B2** | 5A, 5B, 6 | STATUS; re-derive |
| 6b-ii-b | **B3** | 7, then the simplifier round, the targeted data-loss read, Task 8 | the merge ritual, three architecture reads, the ledger fold |

Segment B (or B1) closes after Task 4C because 4A to 4C are the pass's first data-destructive code,
and a measurement there that contradicts this plan's probes needs a ruling before Task 5A builds
on it.

**Models.** `sonnet` for Tasks 1 and 8. **`opus` for Tasks 2 to 7, dispatched as
`claude-opus-5-5`, unconditionally**: Task 2 is the security surface; Task 3 carries cross-process
cases a single-process test passes; Tasks 4A, 4B, 4C, 5A, 5B, and 6 are the data-destructive path; Task
7 carries signals and exit codes whose 6b-i analog took two fix rounds. Task 0 runs at `opus` in
fresh context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1.
- Task 1 → Task 2 → Task 3.
- Task 0 → Task 4A → Task 4B → Task 4C → Task 5A → Task 5B → Task 6 → Task 7.
- Tasks 0 to 7 and the simplifier round → Task 8.
- **Independent by files:** Tasks 1 to 3 of Tasks 4A to 7. They still run sequentially, because
  one tree runs one executor and the module gate runs over `./...`, and lane 1 holds the other
  code lane.

**Budget:** the execution ceilings and 80% flags in "Token ceiling" (Ruling 1), planning recorded
separately, STATUS at each boundary, spend scored at close against the ceiling and against zero
attended interactions after approval.

---

## The 2.0 forward hooks, per task

Per CLAUDE.md's build-for-2.0 rule. The 2.0 spec's forward-hooks list (`:450-519`) was read in
full.

| Task | Hook |
| --- | --- |
| 1 | Lands hook 15's channel half at the seam: the network actor carries the channel rung 14's two mounts assert. |
| 2 | Lands hook 9's projection half (the seam a later projector joins is now reachable over HTTP) and hook 15's mount-channel parameter. |
| 3 | Neutral. |
| 4A, 4B, 4C | Neutral: hook 7's runner is reused unchanged; hook 10 is backup's, and restore consumes its copies. |
| 5A, 5B, 6 | Neutral. The spec's "the streamer's state lives under the existing state directory, so backup covers it" (2.0 spec, "Cutover and operations") is rung 12's to add to the restore set, recorded in Task 8's handoff. |
| 7 | Neutral. |

No new hook was discovered.

---

## Prior art consulted

Scored on CLAUDE.md's four criteria. The spec's table (`:895-915`) carries the evidence for the
rows it scored; this pass adopts nothing new, so no new license, maintenance, or adoption evidence
is required.

| Entry | Verdict for 6b-ii |
| --- | --- |
| rclone v1.74.3, subprocess | **Adopt** (spec) for the fetch (`copyto`), the library (`copy`), and, on Ruling 6's yes, `--backup-dir`; quoted or probed in R-6bii9, R-6bii11, and Ruling 6. |
| `internal/atomicfile` | **Adopt** for the intent record and the report (spec `:698`, spec prior art). |
| SQLite `VACUUM INTO` output as restore input | **Adopt** (6b-i's copies); verified with the store's own verifier and ledger check. |
| SQLite backup API | **Decline** for restore (Task 4A). |
| `net/http`, `encoding/json`, `mime` (stdlib) | **Adopt** for the projection. |
| `go-chi/chi` | **Decline** (2.0 prior art's recorded decline; four exact paths). |
| RFC 9457 problem details | **Consulted, declined** (Task 2). |
| OpenSubsonic error codes | **Declined** as a model (6b-i Task 2), unchanged. |
| `_salvage/internal/backup` restore | **Read for flow order only**; import nothing. |
| Navidrome, gonic, taglib, mewkiz/flac, Picard, chromaprint, goose, golang-migrate, modernc's fallback | No 6b-ii surface. |

---

## Not verified in this revision

Fold 1 settled four of revision 1's items by measurement or quote: the `-shm` question (R-6bii8),
the mode `copyto` writes (R-6bii9), the closed `*sql.DB` error (R-6bii5), and the blocking job body
(Task 3). What remains:

- Whether `IntegrationValue` returns the remote while the backup class is disabled; Task 0 item 7.
- Whether a restored snapshot came from this store rather than another; undetectable until 6d's
  store identity (R-9 option b in 6b-i's risk model). The remote-mismatch warning is the one signal
  before then.
- `rclone size`'s accuracy on backends that omit sizes, which its help says count as empty; the
  library free-space check is a floor, and the verb prints the `sizeless` count (R-6bii10).
- An unmounted mount point as the library root: not refused, a declared gap on Ruling 7's yes
  (R-6bii11).
- A library root that is itself a mount point puts Ruling 6's displaced directory on its parent's
  filesystem. The fold probed `copy --backup-dir` across two tmpfs mounts and found the displaced
  file copied across; Task 6 re-runs that probe, and a mount-point root's displaced directory
  costs free space on the parent's filesystem that the library check does not count.
- The master pointers into `internal/auth`, `internal/identity`, and `internal/serve`, which lane
  1 is editing while this plan is drafted.
- 6b-i's close spend, which the ceiling's close-side figure estimates from 6a's actual.
