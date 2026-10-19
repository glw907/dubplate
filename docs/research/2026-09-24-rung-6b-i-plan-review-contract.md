# Rung 6b-i plan review: contract and criteria lens

Target: `docs/plans/2026-09-24-rung-6b-i-ops-seam-backup.md` at
`b36871c` (revision 1, draft). Lens: is every promise testable, does each
criterion name its fixture and why it would fail, can any pass vacuously,
and the four planning-miss items (behaviors quoted, states named, privileged
channels named, fixture states named). Sources read: the plan, the rung 6
spec (`:533-672`, `:919-937`, `:990-1033`), the 6a plan (R-6a2, R-6a15), and
6a's code on `lane-2-6a` (head `aa3b13e`): `write.go`, `lock.go`, `id.go`,
`leaf.go`, `open.go`, `snapshot.go`, `migrate.go`, `0002_jobs.sql`, and
`cmd/dubplate-server/store_test.go`. Behaviors checked against `go doc` at
go1.27.1 and `man rclone` at v1.74.3.

Findings are ranked by consequence. Line numbers are the plan's unless a
file is named.

Counts: 1 blocker, 10 major, 18 minor. Two owner forks (C1, C11).

---

## Blocker

### C1. No channel builds a current store outside `internal/engine/store` (OWNER FORK)

- **Location:** plan `:360-362` (R-6bi3: "Serve mode, the migration runner,
  and the pre-migration snapshot stay unexported until 6d"); consumers at
  `:865-896` (Task 3 criteria 3 to 8, 17), `:987-1011` (Task 4 criteria 9 to
  14 and the orphan row), `:1160-1176` (Task 6 criterion 20 and the exit-code
  table).
- **Defect:** every one of those criteria names "a current store" (Task 4
  adds "a store with two tracks") as its fixture, and Task 1 adds `leases/`
  to the leaf, which only serve mode's create makes. The store's only
  exported function at `lane-2-6a` is `Inspect` (`inspect.go:118`). The write
  mode Task 1 exports refuses an absent store (`:41-42`). 6a already
  recorded the wall: "`cmd/dubplate-server` can construct an inspect report
  (exported data) but cannot build an older or newer store without the
  unexported runner" (6a plan R-6a15, `:720-721`), and `store_test.go` says
  the same in its doc comment. `internal/ops/local`, `internal/backup`, and
  `cmd/dubplate-server` tests therefore have no way to reach the fixture
  state their criteria name. The first implementer to reach Task 3 stops.
- **Fold:** name the fixture channel in R-6bi3 and in Global constraints.
  This is an exception to the standing "nothing is exported ahead of a
  caller" ruling (CLAUDE.md, Geoff 2026-09-12), so it is Geoff's call.
- **OWNER FORK. Options:**
  - (a) A test-support package, `internal/engine/store/storetest`, that
    exports one function creating a current leaf (with `leases/`) in a
    `t.TempDir()`, imported only by `_test.go` files. The exported-unused
    leg flags it (its references are test files only), so the plan names it
    as the leg's one standing exception and exit gate 11 lists it.
  - (b) Export serve mode's create now, ahead of 6d's `serve` caller, with
    tests as its only 6b-i caller. Pulls 6d surface forward and weakens exit
    gate 10's claim that nothing outside a test creates a store.
  - (c) Each consumer package builds a leaf from a checked-in fixture
    database plus the leaf layout written by hand. Duplicates the store's
    layout knowledge (modes, lock file, `leases/`) in three packages and
    goes stale at every migration.
  - **Recommendation: (a).** It keeps the production surface honest, puts
    the layout knowledge in one place, and the exception is visible in the
    gate rather than hidden in duplicated test code.

---

## Major

### C2. Criterion 17's fixture expects the rule's opposite

- **Location:** `:891-894` against Task 3's outcome `:835-837` and spec
  `:578-580`.
- **Defect:** the fixture plants a succeeded record at 100 days, a failed
  record at 100 days, and a succeeded record at 10 days, and expects the
  next start to delete "only the 100-day failed record." The rule deletes
  terminal records past 90 days "except the newest succeeded record of each
  kind." The newest succeeded `backup` record is the 10-day one, so the
  100-day succeeded record is past retention and not protected; it must be
  deleted too. As written, an implementation that never deletes a succeeded
  record passes, and the correct implementation fails the criterion. Named
  mutation 5 (`:914-915`) does not catch the wrong variant either.
- **Fold:** expected outcome: both 100-day records deleted, the 10-day
  succeeded record kept. Keep the second fixture (a lone 100-day succeeded
  record survives). Add one row for an interrupted record past 90 days
  (depends on C9).

### C3. The seam between `ops/local` and `backup` is owned by no task

- **Location:** Task 3 `:810-813`, `:899-901`; Task 4 files `:937-940` and
  `:943-945`; Task 6 files `:1114-1119`; Task 2 `:746-748`; import table
  `:242-245`.
- **Defect:** Task 3 runs "a test-supplied function" as the job body. Task 4
  creates only `internal/backup` files. Task 6 touches only `cmd` and
  `internal/config`. No task's file list includes the edit that makes
  `ops/local`'s start run `backup`'s body, although the import table says
  `ops/local` imports `backup`. The body's contract with the runner is also
  undefined. Task 4 says "each phase is a progress point where the runner
  reads `cancel_requested`," and Task 2's status type carries "the phase,"
  but no task defines how a body reports a phase or who writes the
  `progress` column. Named mutation 7 in Task 3 (`:918-919`, "poll only at
  phase boundaries") cannot be built until a phase boundary exists in Task
  3's own test body. The implementer who needs the wiring must edit outside
  its file list and so must stop.
- **Fold:** Task 3 defines the body type: `ctx` plus a progress callback
  that writes `progress` and is a cancel-poll point. State which task wires
  production: either Task 4 adds `internal/ops/local` to its Modify list, or
  Task 6 passes `backup`'s body into the constructor and the import table
  drops `ops/local -> backup`. Task 3's criterion 6 fixture names a body that
  reports one phase, then blocks for a bounded time longer than the 2 s poll
  interval, then returns success. With that fixture, mutation 7 ends
  succeeded and does not hang.

### C4. `os/exec`'s return after cancel is not quoted, and the canceled and failed rows overlap

- **Location:** Task 4 states table `:977-982`; R-6bi12 `:486-494`; Task 7
  rows `:1230-1231`.
- **Defect:** the job's outcome depends on how `Wait` reports a canceled
  child, and the plan never quotes it. `go doc os/exec.Cmd` (go1.27.1), on
  `Cancel`: "If the command exits with a non-success status, or Cancel
  returns an error that wraps os.ErrProcessDone, Wait and similar methods
  continue to return the command's usual exit status." A SIGINTed rclone (or
  stand-in) returns an `*ExitError`, so a runner that classifies by `Wait`'s
  error writes `rclone-exit-<N>` for a canceled job. The table lists "Sync |
  canceled | canceled" and "Sync | any other nonzero | rclone-exit-<N>"
  without saying which wins, and a canceled sync is also a nonzero exit.
  The same overlap exists in Task 7 between "the probe exits nonzero" and
  "the probe times out."
- **Fold:** quote the passage in R-6bi12. Rule the precedence: the job's
  context error is checked before the exit status, so any exit after the
  context ends records canceled. Doctor's probe applies the same rule: a
  deadline hit means "not checked", whatever the exit. Add the precedence to
  each states table.

### C5. Criterion 9's store-half fixture is self-contradictory

- **Location:** Task 1 `:688-693`.
- **Defect:** the fixture asks for "a writer inserting rows in a loop, and
  one transaction held open across the copy." Every write goes through the
  write method, which runs "one IMMEDIATE transaction on the write pool"
  (`write.go`, `update`'s doc). An open write transaction holds SQLite's
  single write lock, so the looping writer blocks for the busy timeout and
  then errors. Nothing commits during the copy. "While a second handle
  commits in a loop" then passes vacuously, or the loop's busy error makes
  the test flaky. "Every row committed before the copy began" also needs a
  barrier the fixture does not name.
- **Fold:** split it into two fixtures. (a) The looping writer runs across
  the copy. The test records the committed count at a barrier just before
  the copy starts, and asserts the copy holds at least that many rows and
  passes `integrity_check`. (b) An uncommitted write transaction holds one
  marked row across the copy, with no looping writer, and the copy lacks
  that row. Neither half can pass unless a copy actually ran concurrently.

### C6. The SIGINT and `WaitDelay` proofs depend on a stand-in behavior nobody specified

- **Location:** Task 6 criterion 20 `:1167-1172`; Task 4 mutation 7
  `:1027-1028`; review focus 5 `:541-542`.
- **Defect:** "no stand-in process remains after `WaitDelay`" exercises
  `WaitDelay` only if the stand-in survives the SIGINT that `Cancel` sends.
  A "sleeping stand-in" that keeps SIGINT's default action dies at once, so
  leaving `WaitDelay` unset changes nothing and mutation 7 cannot fire. The
  fixture also names no channel for the test to learn the grandchild's pid,
  which it needs to assert that the stand-in is gone. The process-group row
  sends SIGINT to the verb's group, and it does not say that the test starts
  the verb in its own group (`Setpgid`). Without that, the group signal
  reaches the test binary itself.
- **Fold:** define two stand-in modes: one that honors SIGINT, and one that
  ignores it, so only `Kill` after `WaitDelay` ends it. The stand-in writes
  its pid to a file the test names. The test starts the verb with `Setpgid`.
  Criterion 20 runs the SIGINT-ignoring mode, and mutation 7 then leaves the
  stand-in alive (or hangs `Wait` past a test deadline). Either result is a
  failure the test reports.

### C7. "The latest snapshot is never pruned" rests on an unstated clock assumption

- **Location:** Task 5 `:1051-1055`; privileged table `:303`.
- **Defect:** the plan says the run's own snapshot "is the newest, so it is
  never pruned" and keeps the 14 newest by stamp. That holds only if no
  remote stamp is later than this run's. Suppose the remote holds 14 or more
  stamps from the future: a clock stepped back, a box whose RTC reset, or a
  restored box writing to the same remote. Keep-14-newest then deletes the
  snapshot this run just sent, which breaks ruling 10 on the only path that
  deletes remote data. Criterion 15's fixture seeds only past stamps, so
  nothing tests this.
- **Fold:** exclude the run's own object by name, then keep the 13 newest of
  the rest. Add a fixture row: 15 seeded stamps one day in the future. The
  run's own snapshot survives. Name the mutation "keep 14 newest including
  own → the future-stamp row deletes own."

### C8. The failed-run prune criterion is vacuous on its remote half

- **Location:** Task 5 `:1093-1094`, mutation 2 `:1100`.
- **Defect:** the failed run uses the stand-in (exit 7) as rclone. Any prune
  the mutation wrongly runs (`lsjson`, `deletefile`, `purge`) also goes to
  the stand-in, which changes no remote file. So "leaves every remote file
  as it found them" holds with or without the mutation. Only the local half
  (`os.Remove` of pre-migration copies) can fire.
- **Fold:** assert that the stand-in recorded no invocation after the failed
  sync, meaning no `lsjson`, `deletefile`, or `purge`. Or drive the failure
  with the real binary through the criterion 13 shape (zero tracks,
  `--max-delete 0`) and a remote seeded with 15 old snapshots, then assert
  all 15 survive.

### C9. `finished_at` and the other timing and progress columns have no named channel

- **Location:** privileged table `:292-305`; recovery `:814-819`; cancel
  `:828-829`; retention `:835-837`; R-6bi13 `:519`; schema `0002_jobs.sql`
  (`finished_at INTEGER`, nullable).
- **Defect:** retention selects on finish time, and `doctor`'s last-success
  age reads the newest succeeded record's finish time. The plan never says
  that recovery's move to interrupted, or cancel's move from queued to
  canceled, sets `finished_at`. If either leaves it NULL, `finished_at < ?`
  never matches, and those terminal records are kept forever. That is a
  state the retention check meets with no report named. `progress`,
  `started_at`, and `updated_at` also appear in no row of the
  privileged-field table.
- **Fold:** add one row per column naming its writers. Each transition into
  a terminal state (runner, recovery, cancel of a queued record) sets
  `finished_at` and `updated_at` in the same statement as `state`. The
  runner alone sets `started_at` and `progress`. Add a retention row for an
  interrupted record past 90 days, which recovery wrote.

### C10. Criterion 6's queued-cancel fixture is erased by the canceller's own recovery

- **Location:** Task 3 criterion 6 `:879-882`; constructor `:810-813`;
  recovery `:814-819`.
- **Defect:** "cancel on a queued record ends it canceled" runs from a
  second `ops/local`, and that constructor runs recovery first. A planted
  queued record whose owner lease is free or absent is marked interrupted
  before cancel runs. Cancel then reports a terminal state and writes
  nothing, and the criterion fails for a reason that is not a defect. The
  fixture must name a queued record whose owner lease is held (by the
  re-executed child, or by a descriptor the test holds). A second state has
  no report either. The owning runner's queued-to-running transition can
  run after a cross-process cancel has written canceled. That transition
  matches no row, and the plan says only "refused and logged at warning."
  It does not say that the body then must not run.
- **Fold:** name the held-owner fixture state. Add a row: "owner's
  queued-to-running matches no row because the record is canceled: the body
  never runs; logged at info, not warning." Add a matching mutation that
  runs the body anyway, which the stand-in or a body-ran flag detects.

### C11. Exit gate 11 cannot pass for surfaces R-6bi4 and R-6bi5 build (OWNER FORK)

- **Location:** exit gate 11 `:1344-1345`; Global constraints `:253-256`;
  R-6bi4 `:365-376`; R-6bi5 `:378-389`.
- **Defect:** exit gate 11 requires "a production caller outside its own
  package" for every export. At 6b-i, none of the following has one: the
  `forbidden` and `internal` sentinels ("6b-i raises five of them"), the
  `member:` actor kind built from `identity.Identity` (its caller is 6b-ii's
  projection), and the "list the catalog" operation (nothing in 6b-i reads
  the catalog through the interface; `doctor` reads the flag). The close
  gate fails as written, or it gets waived at close without a ruling.
- **Fold / OWNER FORK. Options:**
  - (a) Defer the three to 6b-ii, their first caller. `internal/ops` ships
    five sentinels, one actor kind (`local`), and three operations. The
    `internal/identity` import waits for 6b-ii too, which removes R-6bi1's
    5c-ii coupling.
  - (b) Keep the spec's full vocabulary at 6b-i and record a named exception
    list in exit gate 11.
  - **Recommendation: (a).** It follows the 2026-09-12 refinement ("nothing
    is exported ahead of a caller"). Spec criteria 1 and 2 still pass with
    two privileges and a local actor.

---

## Minor

### C12. R-6bi3's export list omits what Tasks 3 and 4 need, and `ExtraFiles` is unquoted

`:353-363`, `:503-509`. Task 3 must tell a busy kind lease (conflict) from
any other lease error. That needs an exported busy sentinel or typed error,
and `errLockBusy` is unexported (`lock.go`). R-6bi12 passes "the kind
lease's and the store lock's descriptors" to rclone, which needs an export
yielding an `*os.File`. A raw file handed out of the store lets a caller
`Close` it and drop the store lock. Mutation 6 (`:1025-1026`) also relies on
Go opening files close-on-exec by default, which the plan never quotes.
**Fold:** list both exports in R-6bi3, and hand out a dup (`syscall.Dup`
into a new `os.File`) that the caller owns. Quote `go doc os/exec.Cmd`'s
`ExtraFiles` ("additional open files to be inherited by the new process")
and the close-on-exec source.

### C13. Criterion 21's store-fact rows have no state rows

`:1214-1216`, `:1221-1236`, `:1239-1242`. The criterion asks for "one golden
line per row state above," and the table has no rows for WAL size, free
space, or local snapshot count and size. Spec criterion 21 (`:1033`) names
all three. Their states are also unnamed: zero copies, an absent `-wal`,
and a `statfs` failure. Say also whether any of them can fail `doctor`.
**Fold:** add a row per state with its rendered text and exit.

### C14. A succeeded job's "message" has no channel

`:1063-1064`, `:1079` against the privileged table's `failure_message` row
(`:299`, "the transition to failed"). The two contradict each other, and
the diff-reviewer will flag whichever column the implementer picks.
**Fold:** name the column (for example `progress`, or `failure_message` on
succeeded) and add that channel to the table.

### C15. Instance lease files are never removed on a clean exit

`:631-638`, `:814-819`. Recovery tries only instances that a non-terminal
record or a stage names. A clean run leaves `leases/<instance>` behind, one
file per run, forever. **Fold:** the handle's close removes its own lease
file while still holding it, and a Task 1 row proves it.

### C16. Criterion 16's fixture name is not a stamp

`:1089-1092` uses `pre-restore-x.db`. The table's pattern is
`pre-restore-<stamp>.db` and "any other name → untouched" (`:1078`), so a
strict parser keeps it and the criterion fails. **Fold:** use a valid stamp.

### C17. Criterion 4's terminal record must share an owner with a recovered record

`:870-874`, mutation 2 `:907-908`. Suppose recovery selects only
non-terminal rows, but updates by owner without a state guard. Then a
terminal record owned by some other instance is never touched, and the
mutation cannot fire. **Fold:** the fixture's terminal record names the same
dead owner as the free-lease queued record.

### C18. The rename-failure fixture is undone by `create`

`:696-697`. `leaf.create` chmods every leaf directory back to 0700 because
"MkdirAll is a no-op on a directory that already exists" (`leaf.go`). A
test that makes `snapshots/` unwritable and then goes through serve's open
has its fixture repaired before the rename runs. **Fold:** name the entry
point: after create, through the unexported migration runner.

### C19. Two runs in one second share a stamp

`:446-447`, `:962-963`. The stamp layout `20060102T150405Z` has second
resolution. A second run in the same second overwrites the first run's
`snapshots/<stamp>.db` and archives into the same `archive/<stamp>/`.
Criterion 11's back-to-back runs can hit this in the gate. **Fold:** name
the state. Either refuse when `snapshots/<stamp>.db` already exists, or add
sub-second digits that still sort lexicographically.

### C20. SIGTERM and SIGHUP have outcomes but no proof

`:1138-1140` promises all three signals, but criterion 20 drives SIGINT
only, and mutation 3 (`:1185-1186`) only reports. **Fold:** add SIGTERM and
SIGHUP rows to criterion 20's table. They are the same fixture with a
different signal, so they cost little.

### C21. The snapshot's metadata has no gate-resident proof

`:953-956` and spec `:616-617`. The plan measured `run_id` and
`sync_completed` through `lsjson -M` in a probe, but no criterion re-proves
it. The claim "a backend that drops it loses only the metadata"
(`:1463-1465`) is recalled: the rclone manual read here does not say
whether an unsupporting backend drops the flag or errors. **Fold:** add a
metadata read-back to criterion 10. Mark the backend claim as unverified,
or make a metadata failure non-fatal.

### C22. The crypt and empty-name states are incomplete

`:1207-1213`, `:1224-1227`. Three states are unnamed. A `crypt` remote may
have `_REMOTE` unset. It may have `_REMOTE` holding a bare path with no
colon, which is valid rclone config, and the check as worded refuses it.
And an empty name (`:path`) is refused in the outcomes but has no table
row. **Fold:** add these three rows.

### C23. `rclone-missing` and the library root have no owning layer or channel

`:958-959`, `:969`. The rclone path is "a parameter the caller resolves with
`exec.LookPath`," yet the job records `rclone-missing`. The plan does not
say whether the verb refuses before start (exit 1, no record) or the job
fails, and no criterion covers either. The library root's channel
(`DUBPLATE_LIBRARY_ROOT`, `config.go:350`) is also absent from the
constructor's parameters (`:810-813`) and from the privileged table.
**Fold:** name both.

### C24. The deliverable arithmetic is off

`:211-213`. Five tasks at four, one at two, and two at three sum to 28, not
25.

### C25. The pointer base has already moved

`:20-24`. `lane-2-6a` is at `aa3b13e`, not `dedc1b7`. `f4f5ac4` ("make
leaf-first free functions methods on leaf") reshaped `createLeaf` and
`statLeaf`, which Task 0 item 2 (`:572-575`) names as blockers. The
staleness lens owns this; it is noted here because Task 0 will block on it
as written.

### C26. Ruling 3's 130 also covers a cancel no signal caused

`:86-95`. A cancel issued from another `ops/local` also ends the verb
canceled, and it exits 130, the "killed by SIGINT" convention. **Fold:**
state this in the ruling text so Geoff rules on it knowingly.

### C27. The `http` doctor row's network use is unquoted

`:1241-1242`, `:1466-1467`. The plan does not name the fixture URL. It also
does not quote whether `rclone backend features` on an `http` remote makes
a request, and the gate may run offline. **Fold:** name a loopback URL and
record the probe's command.

### C28. Four Task 1 states have reports and no proof

`:670-673`. The four are free space short, stage already exists, fails
verification, and checkpoint busy. **Fold:** name a proof for each, or
declare each reviewer-read. Verification failure has no production-real
trigger.

### C29. A "naming holders" assertion depends on the filesystem

`:665`, `:1007-1011`. `lockHolders` returns nil on btrfs (`lock.go`,
`busyError`'s doc), so the orphan row's "conflicts naming holders" reads
"held by another process" there. **Fold:** name the fixture's filesystem, or
accept either form and assert the stand-in's pid only when `/proc/locks`
matches.
