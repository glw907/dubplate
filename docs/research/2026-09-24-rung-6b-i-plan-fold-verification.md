# Rung 6b-i plan fold verification

Target: `docs/plans/2026-09-24-rung-6b-i-ops-seam-backup.md`,
revision 2 (working tree), against revision 1 (`b36871c`) and the fold
record `docs/research/2026-09-24-rung-6b-i-plan-fold-record.md`.
This is a fresh-context read by an agent that took no part in the fold. 6a's
code was read at `lane-2-6a`, which now equals master at `8736528` after the
6a rebase. The store tree there is identical to the plan's `aa3b13e` pointer
base (`git diff aa3b13e 0b6da0f -- internal/engine/store` is empty). Probes
ran under `/tmp/claude-1000/verify6bi/`.

## Verdict

The fold is sound. The blocker (C1) and every major closed at the cited
location, with one exception: the Setpgid half of C6/M6 is still vacuous
(V1). The fold introduced no task order that cannot build. Each task's Files
list covers its outcomes, including the C3/M3 body binding (Task 4 edits
`internal/ops/local/jobs.go`) and the M4 descriptor export (Task 1A,
`lock.go`). All four rulings can be ruled on the evidence given, with one
outcome per "no". Ruling 1's evidence carries two small errors (V5).

| Severity | Count |
| --- | --- |
| Blocker | 0 |
| Major | 1 (V1) |
| Minor | 7 (V2 to V8) |
| Low | 2 (V9, V10) |

## Question 1: did the blocker and the majors close?

| ID | Closed? | Where, as read |
| --- | --- | --- |
| C1 | Yes | Ruling 4 (plan:115-131), R-6bi3 last bullet (:463-464), global constraints (:311-318), exit gate 11 (:1901-1905). The claim that a `storetest` package needs the same export holds: `openServe`, `serveOptions`, `create`, and the migration runner are unexported at `lane-2-6a`. |
| C2 | Yes | Criterion 17 (:1314-1319) and mutations 6 and 11. |
| C3 / M3 | Yes | R-6bi17 (:777-795), Task 3's body contract (:1217-1221), Task 4 Files (:1375-1376). `backup_test` is named for the orphan row. |
| C4 / M7 | Yes | R-6bi12 precedence (:646-664), Task 4 table (:1430-1433), mutation 10. |
| C5 / M1 | Yes | Task 1B fixtures A and B (:1062-1071). |
| C6 | Mostly | Two stand-in modes, the pid file, and `Setpgid` on the verb child are in place (:1408-1411, :1670-1679). Rows 1 to 3 now exercise `WaitDelay`. The group row's discrimination is vacuous (V1). |
| C7 | Yes | R-6bi15 (:737-757), Task 5 (:1526-1530), the clock row (:1574-1576). |
| C8 | Yes | The failed-run row (:1584-1587). |
| C9 | Yes | Privileged-field rows (:362-364) and Task 3 (:1231-1233). |
| C10 | Yes | Criterion 6's queued record is owned by the child's held lease (:1297-1300), and the cancel-first row covers the rest (:1302-1305). |
| C11 | Yes | R-6bi5 (:485-496), exit gate 11 (:1901-1905), and the deviation (:141-145). |
| M2 | Yes | R-6bi6 (:517-527), Task 1B's checkpoint row, and mutation 2. |
| M4 | Yes, with V2 | R-6bi3 (:457-459) and Task 1A (:938-940, :976-979). |
| M5 | Yes | R-6bi14 (:719-735) and Task 4's join rows. |
| R1 | Yes | R-6bi16 (:759-775) and Task 1B's reference rows and mutation 5. The mechanism matches 6a's guard: `besideCopy` calls `newestCopy`, which today admits only `pre-migrate-v<N>.db` and `pre-restore-*.db` regular files (`open.go:474-500`). `restore --snapshot <stamp>` is a spec verb (spec :680). |

## Question 3: mechanisms stated in the fold

| Mechanism | Status |
| --- | --- |
| The `.ref` read by the lost-database guard | Consistent with `newestCopy` and `besideCopy` on `lane-2-6a`. `pruneSnapshots` touches only exact `pre-migrate-v<N>.db` names, so it never deletes a reference. |
| `busy_timeout=0` on the checkpoint's connection | The mechanics probe measured both sides (review M2). `openMaintenanceConn` sets `SetMaxOpenConns(1)`, and its DSN carries `_busy_timeout`, so the pragma override is needed and sticks to the one connection. |
| `path.Join` after a first-colon split | Probed. `name:`, `name:dir`, `name:dir/`, and `name:/abs` produce exactly the table's destinations. |
| Microsecond stamp `20060102T150405.000000Z` | Probed on go1.27.1. It round-trips, and parsing rejects missing, short, or long fractions and any zone other than a literal `Z`. `Format` on a non-UTC time writes local digits with a literal `Z`, so the job must call `.UTC()`. The plan says UTC, so this is a note and not a finding. |
| The prune skip rule | Sound as logic. A backward clock step accumulates snapshots rather than deleting the run's own. |
| `--low-level-retries 1` | `rclone backend features --low-level-retries 1` is accepted on v1.74.3 (probed). The B2 timings come from the mechanics probe (review M13). |
| Lease removal on clean close | Follows from `flock(2)` and the M9 probe. It is safe because a cleanly closed runner has no non-terminal record left, so a racing try that reads "free" has nothing to interrupt. |
| Serve-mode export with a nil-logger guard | Needed. `openServe` dereferences `opts.logger` on a probe warning, and `waitLock` logs through it. |
| `lsjson` exit 3 on a missing directory | Re-probed: exit 3, `directory not found`, including a nested path. |
| `dup(2)` descriptors | Stated from memory, and wrong about close-on-exec (V2). |

## Question 4: folds that avoided owner forks

- **C11 (deferral).** Internal only, and within 6b-i's authority.
- **R9 (runbook and a 6d store id).** Records a residual and a forward item.
  It changes no 6b-i behavior. 6d's plan still decides.
- **R10 (rung 7).** Within authority. Nothing changes at 6b-i.
- **R1 (`.ref`).** It restores the reach of erratum E11, which is already
  ratified, so it does not override an owner ruling. It is still
  adopter-visible: it adds a new operator artifact, a refusal, and a runbook
  procedure ("remove the reference to start fresh").
- **R-6bi15 (the prune skip).** Qualifies ruling 10, safely.

Both R-6bi15 and R-6bi16 are listed under the deviations, so approval does
cover them. The rulings preamble, though, says the spec's rulings "are not
reopened" (V8).

## Question 5: can Geoff rule the four rulings?

Yes, each with one outcome per "no".

- **Ruling 1.** "No" holds 5.9M as a total and puts the flag on the plan.
  Two evidence errors need correcting before Geoff reads it (V5).
- **Ruling 2.** "No" builds the static table.
- **Ruling 3.** "No" means 2, 3, and 4.
- **Ruling 4.** "No" builds the leaf template. That branch has no Files
  wiring (V6), but the conductor can amend the plan at Task 0 item 12.

## Findings

### V1 (Major): criterion 20's group row cannot fail without `Setpgid`

- **Location:** plan:1670-1679. Also Task 4 mutation 5 (:1490-1491) and
  review focus 5 (:827-829).
- **Defect:** All four rows run "against the SIGINT-ignoring stand-in". The
  group row asserts that the stand-in is still alive 500 ms after a group
  SIGINT. An ignoring stand-in survives a group SIGINT whether or not
  `Setpgid` is set, so the row passes with `Setpgid` dropped. M6's point is
  lost. Task 4 mutation 5 defers the terminal case to this row, so no test
  pins `Setpgid`. A regression would make Ctrl-C at a terminal kill a real
  rclone first: it exits 130 before the runner's poll cancels the context,
  so the job ends `failed` (`rclone-exit-130`) and the verb exits 1 instead
  of 130. The obvious fix, switching to the SIGINT-honoring stand-in with a
  500 ms "alive" check, would flake: the runner's 2 s poll can land inside
  500 ms of the cancel write, about one run in four.
- **Proposed fold:** Discriminate by signal count, not timing. The ignoring
  stand-in appends one line to its record file per SIGINT received. The
  group row asserts exactly one SIGINT: the runner's `Cancel`. Without
  `Setpgid`, the terminal's group signal adds a second line. Task 4 mutation
  5 then names this row as the one that fires.

### V2 (Minor): a `dup(2)` descriptor is not close-on-exec

- **Location:** plan:679-680, R-6bi3 (:457-459), Task 1A (:938-940).
- **Defect:** The plan argues that "nothing but `ExtraFiles` reaches the
  child" because Go opens files close-on-exec, then specifies `dup(2)`
  duplicates. A probe on go1.27.1 shows that `syscall.Dup` returns a
  descriptor without `FD_CLOEXEC`, while `fcntl(F_DUPFD_CLOEXEC)` returns
  one with it. At 6b-i the only exec in the verb is the transfer, so nothing
  leaks yet. From 6d, serve runs backup in-process. There, any concurrent
  `exec` inherits a stray kind-lease and store-lock descriptor, and holds
  both for its whole life.
- **Proposed fold:** Say `fcntl(F_DUPFD_CLOEXEC)`. `os/exec` clears the
  flag on the `ExtraFiles` slots in the child. Add a Task 1A row that
  asserts the returned descriptors carry `FD_CLOEXEC`.

### V3 (Minor): the instance lease has two owners in the text

- **Location:** Task 1A (:926-930 "held for the handle's life", :975 "A
  handle's clean close leaves `leases/` without its instance file"),
  against Task 3 (:1206-1210, where the constructor takes and releases the
  lease) and the privileged-field row (:368, "its owning runner").
- **Defect:** A Task 1A implementer could bind the instance lease to the
  store handle's open. Task 3's constructor then takes a second one, and
  records name only one of them. The receiver of the descriptor method (the
  kind lease or the handle) is also unstated, although it needs both
  descriptors.
- **Proposed fold:** State the ownership in one sentence. The lease
  primitive returns lease values, the store handle takes none, and
  `ops/local`'s runner owns the instance lease. The descriptor method sits on
  the kind-lease value, which was taken through a handle, so it can
  duplicate that handle's lock too.

### V4 (Minor): the stamp layout has no named home

- **Location:** R-6bi8 (:573-577) and Task 1B (:1021-1024, "refuses a stamp
  that does not parse through the layout").
- **Defect:** The store must parse stamps for the reference and for
  6b-ii's `pre-restore-<stamp>.db`. `internal/backup` must parse the
  remote's listings. `backup` imports `store`, so the layout lives in the
  store, or it is duplicated and drifts. R-6bi3's export list omits it.
- **Proposed fold:** The store owns the layout and exports it, or a
  format and parse pair. Add it to R-6bi3 with `internal/backup` as its
  caller.

### V5 (Minor): Ruling 1's evidence misstates two numbers

- **Location:** plan:77-80.
- **Defect:** "6a's measured record prices the close side at about 2.1M"
  contradicts "Not verified" (:2042) and the token section (:235-236),
  which both call 2.1M a projection. "About a third of a task above the
  projection" is 0.36M against a 0.40M task, which is about 0.9 of a task.
- **Proposed fold:** "6a's projected close side, about 2.1M", and "0.36M,
  just under one task, above the projection".

### V6 (Minor): Ruling 4's "no" has no Files wiring

- **Location:** plan:127-131 and Task 1A Files (:910-913).
- **Defect:** On "no", Task 1A needs `testdata/leaf/`, a regeneration
  test, and a copy-and-chmod helper that three packages use. git keeps
  none of the leaf's modes. No Files line and no helper home is named.
- **Proposed fold:** One sentence. On "no", the conductor amends Task 1A's
  Files before dispatch and names the helper's home. The chmod step is
  required because git does not keep the leaf's modes.

### V7 (Minor): the pointer base is now unreachable

- **Location:** plan:23-32 and Task 0 item 2 (:860-870).
- **Defect:** The 6a merge rebased `lane-2-6a`. `aa3b13e`, `3b962aa`, and
  `f4f5ac4` are now reachable only through the reflog, and `git gc` will
  eventually prune them. Their content survives as `0b6da0f`, `8736528`,
  and `f85e7be`, with an identical store tree.
- **Proposed fold:** Repoint the header and Task 0's diff base to
  `0b6da0f`.

### V8 (Minor): "not reopened" understates two deviations

- **Location:** plan:71-72 against R-6bi15 (:166-169) and R-6bi16
  (:170-173).
- **Defect:** R-6bi15 qualifies ruling 10: a clock stepped back skips
  pruning, so snapshots accumulate past 14. R-6bi16 adds an operator-facing
  refusal and a "remove the reference to start fresh" procedure next to
  ruling 7. Both are adopter-visible and safe. The preamble says no ruling
  is reopened.
- **Proposed fold:** Add one sentence to the preamble saying that approval
  also approves the listed deviations, and that R-6bi15 and R-6bi16 qualify
  rulings 10 and 7.

### V9 (Low): an exit 0 racing a cancel

- **Location:** R-6bi12 rule 1 (:657, "whatever the exit") against Task 3
  (:1239-1242, "a job that finishes before it sees the request keeps its
  own state").
- **Defect:** A cancel that lands just after the send exits 0 records
  `canceled`. The snapshot has landed, but no reference is written. This is
  harmless, but the two sentences disagree.
- **Proposed fold:** Rule 1 applies to a nonzero exit. An exit 0 continues,
  and the next progress callback observes the cancel.

### V10 (Low): the SIGPIPE-bounded orphan claim assumes stderr writes

- **Location:** plan:686-691.
- **Defect:** At the default log level, rclone can go the whole sync
  without writing to stderr, so the claim that an orphan ends "short in
  practice" is optimistic. This is a note in the HISTORY entry, and no
  criterion rests on it.
- **Proposed fold:** Change it to "exits at its next stderr write, which at
  the default log level may be the sync's end".
