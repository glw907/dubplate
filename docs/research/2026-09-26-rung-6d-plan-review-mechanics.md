# Rung 6d plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md` at `5afe46b` (revision 1).
Lens: does every mechanism the plan relies on behave as stated? Each finding was checked against the
source at `5afe46b`, `go doc`, a man page, or a probe run in a scratch copy of the committed tree
(`git archive HEAD go.mod go.sum internal`, with probe tests added only in the copy). No web/ file was
read.

**Counts: 0 blockers, 6 major, 9 minor.** No owner fork: every finding has one clearly correct fold.

## What is sound

- **`OpenServe` as the plan wires it.** The nil-logger refusal is the first line (`open.go:220-222`),
  so R-6d16's "confirm it" item is already true. The refusal matrix in Task 1's table matches the
  code: intents other than reconcile-pending refuse before any file is touched (`:232-234`), a newer
  store refuses in both `openShared` and `settle`, an absent or interrupted create is completed under
  the exclusive lock, and `besideCopy` keeps E11's lost-database refusal. `waitLock`'s expiry error
  wraps `busyError`, which names `held by pid N` when `/proc/locks` matches (`lock.go:105-112`), so
  criterion 11's "naming the holder" holds on the tmpfs `t.TempDir()` uses here (`df -T /tmp`:
  tmpfs). `migrate` refuses `keep < 1` (`migrate.go:103-105`), and the setting's floor of 1 covers it.
- **The reconciler's root is reachable.** A reconcile-pending record keeps its `Library` half, since
  `Advance` copies the record whole (`advanced := *r`, `internal/engine/store/restore.go`), so the
  store can hand `intentRecord.Library.Root` to a widened `Reconcile` signature. Nothing in production
  calls `ServeOptions.Reconcile` yet, so the signature change is free.
- **The seed apply export.** `apply` is one `Update` with the read inside the transaction, and
  `writeChanges` writes rows, audit rows, and one generation bump (`apply.go`). Exporting it and
  calling it after `OpenServe` is mechanically trivial. Task 2's state rows match `outcome`'s cases
  (`evaluate.go:292-345`), including the admin-row refusal naming both values.
- **Correlated rules.** `evaluate` runs every rule over the values in force after the whole seed and
  stamps the wrapped refusal on every key the rule names (`evaluate.go:246-270`), so "refused naming
  both keys" and "raised together passes" hold as Task 6 states them. `AddCorrelatedRule` refuses a
  key that is not a registered setting, so the rules must register after the ten settings.
- **`AddSuperseded`.** The load error renders as `KEY: <msg>` (`config.go` `rejectNamespace`), and a
  key removed from `magicLinkSource.Required` no longer collides in the claim map. The template-keys
  test classifies a superseded name as known, so the Task 5 to Task 8 gap cannot fail it.
- **`systemd-run` (R-6d15).** Probed on systemd 259 (the box runs 257, per the 5d plan):
  `systemd-run --user --wait --pipe -q -p EnvironmentFile=<file> sh -c '...; exit 3'` loaded both
  variables and returned `rc=3`. A missing `EnvironmentFile=` failed with `rc=1`, and a `-`-prefixed
  one ran with `rc=0`. systemd-run(1): "If systemd-run waits for the service to terminate, the return
  value will be propagated from the service." So `deploy.sh` can halt on the owner-run `check`. The
  runuser(1) quote is verbatim, and it is the right reason to deviate from the spec.
- **Unit and ownership.** 5d's unit is `Type=exec` (5d plan `:503`), so a 5-minute lock wait meets no
  start timeout. `deploy.sh` already chowns `/var/lib/dubplate` to `music` (`deploy.sh:505`), and
  `leaf.create` requires the state directory to be owned by the process uid (`leaf.go:71-77`), so a
  pre-created 0700 `store` leaf plus `serve` as `music` creates cleanly. The create path `chmod`s the
  leaf to 0700 even if it already exists.
- **Schema.** `users` has `id BLOB PK`, `address UNIQUE`, and `streamer_username UNIQUE COLLATE NOCASE`
  (`0003_catalog.sql:1-10`). The roster's usernames are ASCII by `isASCIIPrintable` (loader) or the
  pattern (write path), so NOCASE and `strings.EqualFold` agree. `member_imports` with no `kind` column
  and a name not ending `_events`, `_audit`, or `_journal` escapes invariant 3
  (`invariant_test.go:329-351`), and a BLOB `id` satisfies invariant 6.
- **Quoted behaviors.** `sql.Row.Scan` (`go doc`), `store.Row.Scan` (`write.go:178-179`), `Update`
  (`write.go:44-63`), `NewID` (`id.go`), and the `ReadWritePaths=` `-` prefix (systemd.exec(5):
  "may be prefixed with "-", in which case they will be ignored when they do not exist") all match
  the source.
- **Ruling premises.** `grep -rn '\.Bump(' internal cmd` finds no non-test caller. The demote grep
  finds only the two comments (`admin.go:299`, `roster.go:754`).
- **Arithmetic.** 0.15 + 7.04 + 0.70 + 1.02 = 8.91. 8.91 / 0.8 = 11.14, which rounds up to 11.5, flag
  9.2. The variants (12.5 / 10.0 and 13.5 / 10.8) and the thirty deliverables check out.

## Major

### M1. R-6d4's keep value: the natural extraction gives the wrong number, and out-of-bounds rows pass

- **Location:** plan `:293-298` (R-6d4); Task 1 states `:544-546`.
- **Defect:** R-6d4 says `serve` passes "the local keep's value in force (the stored row, or the
  registry default ...)" and that "a stored keep row outside its bounds refuses the start." Two
  mechanisms get in the way:
  - `settings.EvaluateFacts`'s `Outcome.Value` is the value after this start's seed, not the stored
    row's value.
  - `outcome` heals an out-of-bounds `env-seed` row the environment re-seeds or removes, so no
    refusal is raised for it.

  So neither `Outcome.Value` nor `Outcome.Err` gives R-6d4's number or refusal. Task 1's first row (an
  `env-seed` keep row of 1, backup off, one copy survives) fails if the implementer uses
  `Outcome.Value`. A raw `Outcome.Stored.Value` of 500 passes `migrate`'s `keep >= 1` check unbounded.
  The plan also names no keep for the `Inspect` states that carry no `Facts`:
  - restore in progress, which is exactly Task 1's reconcile-pending row;
  - busy, refused, and not a store;
  - absent, and older below `0002`.
- **Evidence:** probe over `EvaluateFacts`, with the keep setting registered as `store.go` declares
  it:
  ```text
  env-seed row 1, env unset: Value="3" Stored="1" Change=remove Err=<nil>
  env-seed row 500 (out of bounds), env 2: Value="2" Stored="500" Change=reseed Err=<nil>
  admin row 1, env bad: Value="" Stored="1" Change=unchanged Err=environment seed refused: ...
  ```
  `evaluate.go` `outcome`: "an out-of-bounds environment-seeded row that the environment re-seeds or
  removes is healed rather than refused." `inspect.go` `Inspect` returns early on an intent
  (`ls.intent.refusal`) and on `StateBusy`, with `Facts` nil.
- **Fold:** state the extraction in R-6d4, on `settingsOutcomeStatus`'s precedent (`store.go`):
  - With `Report.Facts` non-nil and `SettingsAbsent` false, find the keep row in `Facts.Settings`.
    When a row exists, `storeLocalSnapshotKeepSetting.Validate(row.Value)` must pass or the start
    refuses naming the key and bound. When it passes, the row's value is `Keep`.
  - With no row, or any state without facts, `Keep` is the registry default, and `OpenServe` itself
    decides whether the state refuses.
  - Ignore `Outcome.Err`/`Change` for this read: an unparseable env is the seed's refusal after the
    open, per spec `:411-414`.

  Add a Task 1 state row: "an `env-seed` keep row out of bounds, env in bounds: refused before
  migration".

### M2. R-6d7 misses `evaluate.go`'s `parse` and `equal`: every restart after a string seed refuses `serve`

- **Location:** plan `:322-328` (R-6d7); Task 6 `:785-797`.
- **Defect:** R-6d7 changes registration, the validator, and the accessor. It does not mention
  `settings.parse`, which has no `KindString` arm (`default: "setting kind %v has no parser"`,
  `evaluate.go:367-378`), or `equal`, which calls it. `outcome` calls `parse(s.Kind, stored.Value)`
  on every stored row (`evaluate.go:285-287`). So once the first start seeds
  `DUBPLATE_UNKNOWN_ADDRESS_ANSWER`, every later evaluation refuses that row. Every later `serve`
  start refuses, and `check` and `doctor` report a refusal.
- **Evidence:** in the scratch copy, with a minimal closed-set `KindString` arm added to
  `config.buildValidator` (R-6d7 as written):
  ```text
  env equal to stored env-seed row: Change=unchanged Err=stored env-seed value refused: %!w(<nil>)
  no row, env generic: Change=seed Err=<nil> Value="generic"
  ```
  The first start seeds, and the second refuses. The `%!w(<nil>)` is `s.Validate` returning nil
  inside a wrap whose trigger was `parse`'s error.
- **Fold:**
  - R-6d7 adds `KindString` to `parse`, as the identity parse (return the string), so `equal` becomes
    exact byte equality.
  - Task 6's states table gains the row "a stored `env-seed` string row, env equal: unchanged, no
    refusal".
  - Name the mutation "drop the `parse` arm (the second start refuses)". Criterion 10's
    seed-table-second-row subtest would catch this, but the plan should not leave it to a fix round.

### M3. Criterion 9's "then B at a new start" contradicts Task 2's refusal row

- **Location:** Task 7 acceptance `:847-852`.
- **Defect:** the fixture is "env A seeded and then an `admin` row B written in-package ... the
  limiter uses A until a new start, then B." At that new start the env still holds A beside an admin
  row B. The seed rule refuses the pair: Task 2's own fourth row, and `outcome`'s
  `case o.EnvSet && !equal(...)` ("the environment seeds %s but an administrator set %s; unset %s").
  So `serve` never starts, and "then B" cannot be observed.
- **Evidence:** `evaluate.go:333-335`, and plan `:596`.
- **Fold:** the restart half of the fixture unsets the key in the env, or sets it to B, before the
  second start, and the plan says so. Add the refusal itself as a named sub-case: env A beside admin B
  at restart refuses naming both.

### M4. Criterion 9's "raised cap admits it" is vacuous or fails as a verification-only test

- **Location:** Task 7 acceptance `:849-851` ("in both directions ... raised above admits it"). The
  spec's criterion 9 (spec `:1109-1111`) does not ask for the raised direction.
- **Defect:** `VerifySession` refuses at `AnchoredAt+cap` and also at `ExpiresAt`. `MintSession` and
  `RefreshSession` clamp `ExpiresAt` to `AnchoredAt+cap` using the cap in force at mint or refresh
  (`session.go:147-152`, `:219-233`). There are two cases:
  - A session older than the old cap already fails `ExpiresAt`, so raising the cap cannot admit it.
  - A session younger than the old cap is admitted either way, so a test built on it passes
    vacuously.
- **Evidence:** probe in `internal/auth` (lifetime 10h, cap A 2h, cap B 4h):
  ```text
  minted under capA: ExpiresAt=+2h0m0s
  verify at +3h with raised capB, no refresh: err=auth: session expired
  refreshed at +1h under capB (ExpiresAt=+4h0m0s); verify at +3h: err=<nil>
  ```
- **Fold:** define the raised direction through a refresh. A session minted under cap A is refreshed
  after the admin row raises the cap to B, the refreshed cookie's expiry passes `AnchoredAt+A`, and a
  verification between the two admits it. Name the fixture state (refresh after the write, verify
  past A). Alternatively, drop "both directions", since the spec requires only enforcement.

### M5. Task 1's D10 prune proof: the "second run" cannot prune, and the default case is vacuous

- **Location:** Task 1 states `:544-545`; acceptance `:560-561`.
- **Defect:**
  - The acceptance says "a second run with the row removed equals the default." After the first run
    the store is current, and `migrate` returns before any snapshot or prune when
    `current == head` (`migrate.go:121-123`). A second run never prunes.
  - The table's keep-row-absent row is "older, same", which is two pre-migration copies. With the new
    `pre-migrate-v4.db`, that makes three, equal to the default of 3. `pruneSnapshots` returns when
    `len(versions) <= keep` (`snapshot.go:236`), so the proof passes with prune deleted.
  - The acceptance's "more copies than either value" contradicts the table.
- **Evidence:** `migrate.go:102-140`, `snapshot.go:225-245`.
- **Fold:** two separate `0004` fixtures. The keep-row-1 fixture holds two copies (three after the
  new one) and is pruned to 1. The keep-row-absent fixture holds `pre-migrate-v1..v3` (four after
  the new one) and is pruned to 3. Drop "a second run". Add the mutation "skip the prune" and require
  both rows to fail under it.

### M6. The Task 6 to Task 7 seam is not independently green, and the interim is unspecified

- **Location:** Task 6 `:779-824`, Task 7 `:828-863`; R-6d8.
- **Defect:** Task 6 moves the ten keys from `AddThreshold` to `AddSetting`. `serve` still resolves
  them in `resolveAuthConfig` through `config.ParseDuration`/`ParseInt`, which take a
  `config.Threshold` (`go doc`: `func ParseDuration(environ []string, t Threshold)`), and Task 7
  deletes `resolveAuthConfig`. A key registered as both a threshold and a setting panics in `claim`
  (`registry.go:126-131`). So Task 6 must either:
  - keep unregistered `Threshold` values that read the env directly, which bypasses the store the
    task just made the home;
  - or read the ten values once through the accessor at start, which is half of Task 7.

  The plan names neither. An implementer picks one, and the diff-reviewer grades it against criteria
  that do not mention it.
- **Evidence:** `auth.go:514-546` (`resolveAuthConfig`), `store.go`'s `storeBusyTimeout` precedent for
  threshold reads, and `settings/doc.go` ("this package exports no reader of a setting's environment
  value" is on `config.Setting`).
- **Fold:** Task 6's outcomes gain "serve reads the ten values once at start through the accessor
  over its handle (the restart-mode interim), replacing `resolveAuthConfig`'s env parse". Task 7 then
  converts the "immediately" and "next operation" values to per-use readers and deletes `authConfig`.
  Alternatively, merge the seam into Task 7 and move `authConfig`'s deletion there, which it already
  names.

## Minor

### m1. The settle loop's re-entry warning would fire on every normal create and migration

- **Location:** R-6d16 `:421-424`; Task 1 outcome 4.
- **Defect:** `OpenServe`'s loop always re-enters once after a successful `settle`: shared wait, then
  `openShared` returns the handle (`open.go:248-269`). "A warning line per re-entry" as written logs
  on every first start and every upgrade.
- **Fold:** warn only when `openShared` again finds work after a `settle` already ran in this call
  (the second and later settle). Add a Task 1 assertion that a plain migration logs no such line.

### m2. `check`'s holder line would list `check`'s own pid

- **Location:** Task 8 states `:890-897`.
- **Defect:** `Inspect` holds its own shared lock while it reads (`inspect.go`, `acquireSharedNonblocking`
  then `defer lock.release()`). `/proc/locks` lists every flock holder, the caller included. The rows
  "older, no other holder: no holder line" and "current, held shared" would then print `check`'s own
  pid.
- **Evidence:** probe with two shared holders on one file: `/proc/locks` listed both `READ 1772723`
  (self) and `READ 1772707`.
- **Fold:** Task 8 excludes `os.Getpid()` from the holders it reports for a shared-held store, or
  reads them before taking its own lock. Name it in the states table.

### m3. Task 3's `Valid` change breaks `internal/ops` test fixtures the plan does not list

- **Location:** Task 3 `:635-636`, Files `:624-628`; R-6d14 ("6d changes nothing in `internal/ops`");
  "Contested surfaces".
- **Defect:** `ops.NetworkActor` panics unless `id.Valid()` (`ops.go:43-48`). Two fixtures build a
  cookie identity with no id: `validIdentity()` (`internal/ops/ops_test.go:38-45`) and `ann`
  (`internal/ops/opshttp/helpers_test.go:22-28`). Once `Valid` refuses a zero `UserID`, those tests
  panic or fail.
- **Fold:** add both test files to Task 3's Files and to "Contested surfaces", as test-only edits.

### m4. The members facts need version guards, or `Inspect` refuses an older store

- **Location:** R-6d13; Task 4 `:682-703`.
- **Defect:** `readFacts` runs for any store at version 2 or above (`inspect.go`, `jobsVersion`). A
  member count needs `users` (version 3 or above), and the import time needs `member_imports`
  (version 6 or above). An unguarded query turns every older store below 6 into `StateRefused`. That
  fails `check`'s store row, and so `deploy.sh`'s pre-restart halt. It would also break Task 1's
  `0004` fixture in any `doctor` test.
- **Fold:** guard both reads on constants beside `settingsVersion` (the `SettingsAbsent` precedent),
  and add a members-row state "older below `0006`: count shown, import record not yet recorded".

### m5. The store's own owner refusal recommends `runuser`, which R-6d15 shows cannot load the env

- **Location:** R-6d15; `internal/engine/store/leaf.go:113`.
- **Defect:** `checkOwner` answers "run as that user: runuser -u %s -- dubplate-server <command>". After
  6d every store-opening verb needs the rendered env, which `runuser` does not load (quoted in the
  plan). The refusal routes an operator to a command that then fails `config.Load`.
- **Fold:** Task 8, which touches the operator path, rewords the hint to the `systemd-run` form or to
  "run as <name> with the rendered environment". Add `leaf.go` to Task 8's Files.

### m6. The import's validator is unnamed, and Task 4 may delete what Task 5 needs

- **Location:** Task 5 `:729-731`, `:742-753`; Task 4 `:684-685`, `:693`.
- **Defect:** `parseRoster` and `validateRoster` apply the loader's rules: any ASCII-printable
  username, spaces and a leading `.` included, and a blank display name (`roster.go:575-600`). The
  write path refuses both (`validateWriteCandidate`, `:602-625`). Nothing says which rule an imported
  entry meets. Separately, Task 4 deletes "the file-backed constructor and its helpers", and
  `parseRoster`, `validateRoster`, and `memberRecord` are those helpers, which Task 5 then exports.
- **Fold:** Task 5 validates each entry with the write path's rules, since the rows it writes are
  write-path rows, and adds a state "an entry the write path's pattern refuses: refused naming the
  entry". Task 4 keeps the parse and validate helpers for Task 5.

### m7. Two write-path states are unnamed

- **Location:** Task 3 states `:639-655`; Task 5 states `:742-753`.
- **Defect:**
  - `users.address` is `UNIQUE` whether or not a row is tombstoned (`0003_catalog.sql:3`). An import
    entry on a tombstoned address matches no live row. The insert then hits a raw constraint error
    rather than a named report.
  - `AddMember`'s "the roster written must name at least one admin" rule (`roster.go:754-757`) has no
    row for a non-admin CLI add on an empty `users` table, which is 6d's fresh-box state.
- **Fold:** name both. The import reactivates a tombstoned address as `AddMember` does, or refuses
  naming it. A non-admin add to an empty table refuses naming `member add --admin`.

### m8. "The same report body `restore` writes" needs more than the root

- **Location:** R-6d16 `:416-420`.
- **Defect:** `restorer.report` prefixes the body with `restore <stamp>`, a `source:` line from the
  intent's remote, snapshot, or file, and a displaced line built from the stamp
  (`backup/restore.go:629-650`). A `Reconcile` that receives only the handle and the root cannot
  reproduce that head.
- **Fold:** say whether serve's report carries only `report.render`'s body, or pass the stamp and
  source too. `finishRestore` already has the stamp.

### m9. Every write and reader needs a context the plan does not place

- **Location:** R-6d10, R-6d9; Task 7's nonce readers.
- **Defect:** `Update` refuses a context without a deadline (`errNoDeadline`, `write.go:69-71`).
  Request contexts carry none, so the roster API's add through `serve`'s handle fails unless someone
  adds a timeout. `NonceStore.Insert(nonce, now, expiresAt)` and `Sweep(now)` take no context, yet a
  retention reader needs one.
- **Fold:** name the layer that adds the write deadline (the roster's `AddMember`, on `backupVerbWait`'s
  precedent), and give `Insert` and `Sweep` a context (and `RunSweep` its caller's) in Task 7's Files.

## Planning-miss checks

1. **Quoted behaviors.** The quoted rows are accurate. Six load-bearing behaviors are relied on but
   not quoted, and each is a finding above:
   - `EvaluateFacts`'s post-seed `Value` and its heal rule (M1);
   - `Inspect` returning no facts for restore-in-progress and busy (M1);
   - `parse` having no string arm (M2);
   - the `ExpiresAt` clamp (M4);
   - `/proc/locks` listing the caller (m2);
   - the loop's routine re-entry (m1).
2. **States named.** The tables are thorough. The gaps are the keep's no-facts states (M1), the
   stored string row at restart (M2), the members read on older stores (m4), the tombstoned import
   address and the first non-admin add (m7), and the write-pattern import entry (m6).
3. **Privileged-field channels.** The table is complete and matches the code's admin rule
   (`checkAddition` refuses the grant off the command line). `ServeOptions.Keep` names its channel,
   but not the extraction (M1).
4. **Fixture states.** Most proofs name theirs. Three would pass vacuously or cannot pass as worded:
   the raised cap (M4), the restart with env A beside admin B (M3), and the prune-to-default fixture
   (M5).
