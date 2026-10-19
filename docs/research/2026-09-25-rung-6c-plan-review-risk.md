# Rung 6c plan review: domain risk (data integrity and failure)

Plan: `docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md` at `8d0eed8`
(revision 1). Lens: data integrity and failure only. Code read on master at `8d0eed8` and on
`lane-2-6b-ii-b` at `30a66ac` (read-only). Weighed against the rung 6 spec (revision 2.2) and CH.

**Counts:** 1 blocker, 6 major (one an OWNER FORK), 7 minor.

Findings are ranked by consequence, not polish. Each proposed fold is the cheapest one that removes
the risk.

---

## Blocker

### B1. A zero-valued `backup.Job` field silently disables the guard or purges the archive

**Location:** plan `:477-479` (R-6c11, "its three values become fields of `backup.Job`"), `:796`
(Task 4 Files), `:830` ("the existing 6b-i tests ... pass unchanged against the defaults").

**Defect.** Today the guard fraction, the snapshot keep, and the archive retention are constants
that cannot be absent. After Task 4 they are `backup.Job` fields, and the plan names no rule for a
field left at its zero value. Here is what each zero does in the code at `30a66ac`:

- **`guardFraction` 0.** The check `float64(files) < 0*float64(tracks)` in `guard.go:86` is never
  true, so the short-tree guard is off. A wiped or unmounted-then-empty tree that still holds one
  file passes, and the sync archives up to `--max-delete` = every catalog track.
- **`archiveRetention` 0.** In `prune.go:154`, `own.Sub(stamp) <= 0` keeps only the run's own
  archive directory and purges every older one.
- **`snapshotKeep` 0.** `othersKeep` becomes -1, and `others[othersKeep:]` panics (`prune.go:123`).
  The failure is loud, but a keep of 1 through a bypass deletes every other snapshot.

The production path is safe only because `ops/local` fills the fields from a bounds-checked
accessor. `internal/backup` itself would accept any value. The plan also asserts that the existing
6b-i tests pass "unchanged against the defaults." Those tests build `Job` literals without the new
fields, so they would pass only if `internal/backup` supplied defaults for zero values. That would
be a second, silent copy of the registry default on the data-loss path. The Sonnet implementer
would have to make that decision alone.

**Fold.**
1. `backup.Run` validates its three fields before any rclone launch and fails with a named code
   (for example `invalid-job`) that names the field: keep at least 1, fraction in (0, 1] and not
   NaN, archive retention above 0. `internal/backup` never supplies a default.
2. Existing tests set the fields explicitly through one test helper whose literals cite R-6c11.
3. Replace `:830`'s "unchanged" with "pass with the fields set to the defaults."
4. Add a Task 4 state row, "`Job` with each field at its zero value → refused naming the field; the
   stand-in records no rclone call."
5. Add a named mutation: remove the check, and the zero-fraction row runs rclone.

---

## Major

### M1. R-6c5's eager parse makes a typo in a store-homed key fail `backup` and every other verb

**Location:** `:350-357` (R-6c5), `:677-680` (Task 2 `Load` rows).

**Defect.** Today a threshold is "a declared name whose own Parse\* call resolves and validates it
at the point of use" (`internal/config/config.go`, `rejectNamespace`). `Load` never parses it, so
a bad auth value never fails `backup`. R-6c5 makes `Load` parse every set store-homed key.
`backup.go:122` returns `Load`'s error, so an env typo in any store-homed key stops timer-run
backups. At 6d that covers seven auth keys. `restore` is affected too, since it loads the same
wall. The spec's stated reason is "so a timer-run backup never fails over an auth threshold it does
not read" (`:744-745`). Backup also never applies an env value: it runs with the stored value
(criterion 8). So failing it on an env parse protects nothing and stops backups.

**Fold (one correct answer; not a fork).** Read the spec row "any | unparseable | a wall error, as
today" by today's precedent. The error is the parse error text, raised where the value is used,
and for a setting the seed rule is that use.
1. `Load` classifies a store-homed key as declared and does not parse it.
2. The seed evaluation reports an unparseable env value as a refusing outcome: serve refuses,
   `doctor` and `check` report it (under Ruling 2, they fail), and `backup` logs it and runs.
3. Rework the Task 2 `Load` row "set and not parsing" to "no error; the evaluation reports it."
   Task 3's last seed row and Task 5's "invalid" row read from the evaluation.
4. Record the reading as D7.

### M2. A stored value that is out of bounds refuses even when the env would repair it

**Location:** `:742` (Task 3 seed row "any | stored value outside the bounds | refused"), `:875`
(the `doctor` row), R-6c9 `:438-445`.

**Defect.** The closed table's "any | stored outside bounds" row overlaps the `env-seed` rows, and
the plan states no precedence. It tests the row as unconditional. Take an upgrade that narrows a
bound, say the keep ceiling from 1000 to 100, with a stored `env-seed` 500 and an env set to 50.
The env re-seed would fix the row, but the unconditional row refuses serve's start. No env value
can clear the refusal, because the refusal fires on the stored row. The only way out is hand SQL
against the store. An `env-seed` row with the env removed (ruling 13) is stuck the same way. The
spec's prose settles it: the seed rule "validates each key's bounds and every correlated rule
against the post-seed set" (`:740-742`).

**Fold.** Validate bounds against the post-seed value. An out-of-bounds `env-seed` row that the env
replaces or removes is healed and audited. It refuses only when the stored value stays in force:
an `admin` row, or an env equal to the out-of-bounds value. Split Task 3's row into three:
- `env-seed` out of bounds with an in-bounds env → re-seeded.
- `env-seed` out of bounds with the env absent → removed.
- `admin` out of bounds → refused.

Mirror the split in Task 5's `doctor` rows. The accessor still errors on an out-of-bounds stored
row, which is correct for a consumer. Route one line to the rung 14 handoff: an `admin` row that
refuses serve's start needs a resolving channel that works while serve is down, such as a CLI verb.

### M3. The migration's local-copy keep cannot come from an accessor bound to a `Handle`

**Location:** `:444-445` (R-6c9, "The pre-seed read ... is the accessor itself"), `:470` and
`:824-825` (Task 4 criterion 13 for the local keep), `:926-927` (Task 6 handoff).

**Defect.** `ServeOptions.Keep` is an input to `store.OpenServe`, and `OpenServe` is what opens,
migrates, and returns the `Handle`. An accessor needs a `Handle` (R-6c8), so it cannot exist before
the migration that consumes `Keep`. `OpenWrite` refuses an older store (`open.go`, `writeRefusal`),
so no second handle is available either. Task 4's criterion asks for "the accessor's value passed
as `ServeOptions.Keep` to `store.OpenServe` over a store with pending migrations," which cannot be
built as written. The implementer will improvise, likely with a raw read the store is not allowed
to make. Separately, an upgrade from a store at `0003` has no `settings` or `settings_generation`
table. Under R-6c8, "A generation read that fails is an error," so an accessor there makes 6d's
serve refuse to start on the first upgrade. No task names the state "settings table absent."

**Fold.** R-6c10 already puts the raw rows in `Inspect`'s facts, and `Inspect` is read-only and
needs no handle. Make the pre-open read "resolve from facts."
1. Add an exported `internal/settings` function that resolves a key's value from `Facts` plus the
   registry. Its first production caller is `doctor`'s local-keep row. 6d's serve calls `Inspect`,
   resolves, and passes the result as `Keep`.
2. Add a state row: a store whose version predates the settings migration resolves every key to its
   default, reported as "store older than settings."
3. Rewrite the Task 4 criterion: resolve from `Inspect` over a store at `0004` holding a seeded keep
   row with `0005` pending, pass the value to `OpenServe`, and assert that the copies are pruned to
   that count.
4. Correct the Task 6 handoff to match.

### M4. The accessor has no specified read path for many rows, and its cache can go stale for good

**Location:** `:421-429` (R-6c8), `:436-445` (R-6c9, `backup` evaluating "from its write-mode
handle"), `:228-231` (the only `database/sql` behavior quoted).

**Defect.** `Handle` exposes one read, `QueryRow` on the read pool (`conn.go:108`), which returns at
most one row. The reload in R-6c8 and `backup`'s evaluation, which needs orphan rows, both need
every row. The plan names no store addition for this, and Task 3's Files and the territory wall
(`:178-183`) list none. The implementer is left with three options: stop, read under
`Handle.Update`, or improvise. Reading under `Update` takes the write lock for a read on every
backup start and every reload.

The plan also does not order the generation read against the row reads. Each `QueryRow` is its own
autocommit snapshot. In the sequence "read rows (old) → another process commits and bumps → read
generation (new)," the accessor caches old values under the new generation and serves them until
the next write. Criterion 10's test cannot catch this, because it serializes the write between two
reads. That is exactly the stale cross-process read that criterion 10 is meant to rule out.

**Fold.** State one of two rules in R-6c8:
- (a) Add `store.Handle.Query`, or a read-only transaction method on the read pool, exported with
  `internal/settings` as its first caller. Every reload reads the generation and the rows in one
  read transaction. Add the store file to Task 3's Files.
- (b) Read the generation first, then the rows, and cache under the generation read first. This
  never caches older values under a newer generation.

Recommend (a), because it also gives `backup`'s evaluation its orphan rows. Quote the WAL snapshot
rule from sqlite.org/wal.html and cite it. Add a reviewer checklist line: "no reload reads rows
before the generation outside one transaction."

### M5. Task 4 may change a retention default on its own research

**Location:** `:460-462` ("unless Task 4's research gives a recorded reason to change one"),
`:1091-1092`.

**Defect.** The defaults for `DUBPLATE_BACKUP_SNAPSHOT_KEEP` and `DUBPLATE_BACKUP_ARCHIVE_RETENTION`
decide how much remote history the next backup run deletes, and that deletion is irreversible.
Lowering either default on the implementer's own research is a product decision that would take
effect on the first run after deploy, with no owner in the loop. CLAUDE.md reserves such calls for
the owner, and "no ceremony" applies only where one answer is clearly correct.

**Fold.** Task 4 carries every 6b-i value unchanged and cites its source, R-6bi8 or the recorded
research. If a research finding argues for a change, the task reports it as a decision it did not
make, and the conductor routes it to a ruling. The only default Task 4 may set is the local keep's,
which is new (R-6c11).

### M6. OWNER FORK: the bound floors allow a mis-seed that deletes backup history

**Location:** `:467-469` (R-6c11 bounds table).

**Defect.** The floors are the only barrier against an operator typo at 6d's seed and an admin write
at rung 14. Three of them sit where a mis-set value destroys recovery material:
- **Keep of 1.** One run after an upgrade deletes every pre-upgrade remote snapshot. That is the
  rollback path in rulings 7 and 10 (`:94`, `:430`).
- **Archive retention of 24h.** Files that a bad sync moved into `archive/` are purged within a
  day.
- **Guard fraction just above 0**, for example `0.01`. The short-tree guard is effectively off.

The bounds cite no source (spec `:768-771` asks for researched defaults, and the bounds carry the
same weight).

**Options.**
- (a) Keep the floors as written. This gives maximum operator latitude.
- (b) Protective floors: keep at least 3, archive retention at least 168h (7 days, longer than a
  weekly check-in), guard fraction at least 0.25.
- (c) Floors of (b) for admin writes only, with env seeds free. This splits the policy by channel,
  and the registry cannot express it at 6c.

**Recommendation:** (b). Each floor still admits a deliberately small deployment, and each one
closes a one-typo path to losing the rollback snapshot or the archive. The ceiling side is harmless
either way.

---

## Minor

### m1. Task 0 blocker 3 misses the site-name grammar if it is routed into 5c-iv

**Location:** `:115-122`, `:549-552` (Task 0 blocker 3), `:973-974` (contested rows).

**Defect.** The plan argues that security L2 (`DUBPLATE_SITE_NAME`'s grammar at
`internal/config/config.go:352`) lands "before 6c's worktree is cut." `81988b1` folded only the
component-layer findings, and L2 is still open. The review ties L2 to "Task 7," which is a 5c-iv
task. If the fold routes L2 there, lane 1 edits `config.go` while 6c's Task 2 edits it in lane 2.
Blocker 3 checks only commits and uncommitted edits at the cut, not a planned 5c-iv edit. The
lane condition holds in substance only once L2 is settled. The consequence is a textual rebase
conflict, which the merge ritual catches, so it is not a data risk.

**Fold.** Blocker 3 also reads the 5c-iii security fold record's L2 disposition. If L2 is routed
into 5c-iv, the conductor either lands it as a standalone commit before the cut or adds a
contested row for `config.go` and holds 5c-iv's L2 edit until 6c merges. With that fold, the
contested rows are otherwise sufficient for 5c-iv, which writes none of 6c's other files.

### m2. The apply must read rows through the transaction, or a rung 14 admin write can be lost

**Location:** `:438-443` (R-6c9: the evaluation "takes ... the stored rows").

**Defect.** An evaluation that takes the rows as a parameter invites the apply to read them on the
read pool and then write inside `Update`. That is a time-of-check-to-time-of-use gap: an admin
write landing between the two is overwritten by a re-seed and lost without a trace. At 6c no admin
writer exists, but 6d exports the apply.

**Fold.** State in R-6c9 that the apply reads every row through `tx.Query` inside the callback
(spec `:355`), and add it to the Task 3 reviewer checklist.

### m3. The privileged-fields table omits `restore` as a channel

**Location:** `:280-290`.

**Defect.** `restore` (6b-ii-b) replaces every settings row, `admin` rows included, along with the
audit history and the generation. It is a channel that writes source `admin` without an audit row
in the current timeline. The generation can move backward, which is safe only because restore's
exclusive lock excludes every live handle and its cache. The plan should record that reasoning.

**Fold.** Add a table row for `restore` covering its consequence and the lock argument. Add a 6d
handoff line: after a restore, the seed may refuse on a restored `admin` row, and `env-seed` rows
re-seed. Record the ruling 13 residual for data-loss keys in HISTORY: when an env line is dropped,
the key returns to a lower default, and the next backup prunes the difference.

### m4. Exit gate 12's grep can pass vacuously

**Location:** `:957-958`.

**Defect.** The grep covers `'internal/**/*.go'` only, not `cmd/`. It matches only a SQL literal
`'admin'`, so a source bound as a parameter escapes it.

**Fold.** Make the apply's source a closed unexported type whose only value is `env-seed`. Grep
`'admin'` and `"admin"` across `internal/` and `cmd/`, excluding tests, and triage every hit.

### m5. The guard's failure text still says "half"

**Location:** Task 4 (`guard.go:87`: "fewer than half the catalog's %d tracks").

**Defect.** A configured fraction makes this message wrong, and the operator reads it when a backup
aborts.

**Fold.** The message names the fraction in use, and `job_test.go:547` changes with it.

### m6. Ruling 2's deploy protection misses the deploys that bring a migration

**Location:** `:84-99` (Ruling 2), `:877` (`doctor` row for an older store), R-6c10 `:449-453`.

**Assessment.** The review agrees with Ruling 2's "yes."

**Defect.** `check` shows "store not checked (older)" exactly when a deploy brings migrations, so a
refusing seed is found only after serve has migrated. At that point the old binary refuses the
newer store. The `doctor` row "the default or the env value" is also ambiguous and can present an
env value as the one in force.

**Fold.** Evaluate whenever the store's version holds the settings table, older included. On a
store older than that, show the default with "env `<v>` seeds at serve's next start."

### m7. States that name no report

**Location:** Tasks 3 and 4.

**Defect.** Three states have no named report, and one apply mode has no stated rule:
- The `settings_generation` row absent (`Scan` returns `ErrNoRows`). It should be an error naming
  the table, never treated as 0.
- Two refusing keys in one evaluation. A test should show that `errors.Join` names both, since
  `:240-241` relies on it.
- A backup key stored out of bounds at backup start. It fails the start, but only the retention
  key has a row.
- A `restart`-mode key served by a reloading accessor would apply live. No 6c key uses that mode,
  so this is a line for the 6d handoff.

**Fold.** Add the three rows and the handoff line.

---

## Checked and sound

- **The savepoint (R-6c12).** A self-terminating file fails at `RELEASE` before the version write
  and the ledger row. The prefix committed by `COMMIT; BEGIN;` stays behind, as the plan states,
  and the gate's scratch store stops such a file from ever merging.
- **The audit.** Audit rows are written in the same transaction as the change, so none can be lost
  or duplicated.
- **Typed equality (R-6c7).** It prevents re-seed churn.
- **Upgrade from `0003`.** The populated fixtures migrated to head prove it.
- **The two-handle criterion 10 fixture.** Sound.
- **Criterion 8's fixture.** Sound.
- **Criterion 13's admin fixture.** Sound: 6 files against 10 tracks passes at 0.5 and aborts at
  0.9.
- **Lane 1.** 5c-iv's plan excludes every other 6c file.
