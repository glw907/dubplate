# Rung 6d plan review: contract and criteria lens

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md` at `5afe46b`
(revision 1). Lens: is every promise testable, does each criterion name its fixture and why it
would fail, can any pass vacuously, and the four planning-miss items. Line numbers are the plan's
at `5afe46b`. Code facts were checked against the tree at `5afe46b`.

**Counts:** 0 blockers, 10 major, 9 minor. No owner fork: every finding below has one clearly
correct answer, proposed in its fold.

## What is sound

- The behaviors table is accurate where checked: `Handle.Update`'s deadline and version rules
  (`write.go:40-64`), `store.Row.Scan`'s `ErrNoRows` carve-out (`write.go:178-180`), and
  `OpenServe`'s nil-logger refusal (`open.go:219-222`). The `NOCASE`/`EqualFold` agreement holds on
  the import path too, because `memberRecord.validate` admits ASCII-printable usernames only
  (`roster.go:589-592`).
- The `Lookup` state "store closed returns the store's error" preserves a real security property:
  the login handler already refuses to treat a non-not-found error as an unknown address
  (`handlers.go:211-222`), so a store error never produces the kind answer for a member.
- The cookie channel's `Update` deadline is already satisfied: `rosterAPI.add` wraps the request in
  `context.WithTimeout` (`admin.go:255`), and `member add` does the same (`member.go:95`).
- Task 7's criterion 9 fixture is well specified: both cap directions, a live session minted under
  A, and the limiter read-once check with its matching mutation.
- Task 1's handle-close mutation (a later `restore` fails on `ErrLockBusy`) and Task 3's
  reactivation byte-for-byte id assertion are genuine failing proofs.
- Task 6's correlated-rule rows prove the both-keys report and the raise-together case.

## Major

### M1. No criterion can observe the listener, so the order mutations are uncaught

Location: plan `:553-557`, `:566-567`, `:596`, `:608-609`.

Task 1's criterion asserts the store version "when the function that precedes `serve.Run`
returns." That does not detect a listener opened inside that function. The named mutation "open the
listener before `OpenServe` (the order test fails)" names an order test no criterion defines. Task
2's refusal row says "no listener" and its mutation "call the apply after `serve.Run` starts (the
refusal row opens a listener)" has the same gap. Task 1's outcome "opens no listener on any
refusal" (`:534`) has no criterion at all.

Fold: name the seam in Task 1. The start path takes its listen step (or `serve.Run`) as an
injected function, and the test's recorder asserts, at the moment it is called, that the store is
at head and the seed's rows are committed. Every refusal row in Task 1's and Task 2's state tables
asserts the recorder was never called. The two mutations then fail on those assertions.

### M2. The lock-wait proof covers only an exclusive holder; the designed-for case is a shared one

Location: plan `:550`, `:558-559`; R-6d3 `:283-287`; spec criterion 11 (`spec:1110-1112`).

R-6d3 sizes the migration wait around a backup's shared lock blocking a migrating start. The state
table and criterion test only "the lock held exclusive by another process." A migrating start that
fails to wait on shared holders, or that waits unbounded on them, passes every named check. The
plan also relies on the `LOCK_SH` quote ("a current-version start never waits on a running
backup") with no proof.

Fold: add two rows with the second process holding the lock shared. Against an older store with
`MigrationWait` 1s, the start exits nonzero naming the holder's pid, and the store stays at its old
version with no pre-migration copy written. Against a current store, the start proceeds without
waiting. Mutation: take the exclusive lock unconditionally (the current-store row blocks).

### M3. The prune proof is vacuous as fixtured

Location: plan `:544-545`, `:560-561`.

The table's fixture holds two pre-migration copies. Migration adds one, so three exist, and the
registry default is 3 (`cmd/dubplate-server/store.go:74`). The "pruned to the registry default"
row therefore prunes nothing and passes under a never-prune mutation. The criterion's "second run
with the row removed" is worse: the store is already current, so the second run does not migrate,
and the runner prunes only after a migration (spec `:409-415`). The criterion's own "more copies
than either value" contradicts the table.

Fold: the fixture holds four pre-migration copies (five after migration). The default case uses its
own fresh `0004` fixture with no keep row, never a second run. Both rows state backup disabled
(`ServeOptions.Backup` false). Add the mutation "never prune," which the default row must then
catch.

### M4. Reactivation can re-grant admin through the cookie channel

Location: R-6d10 `:363-366`, privileged fields `:241`, Task 3 table `:650`.

"Reactivated under the same id with the new fields" does not say what happens to `users.admin`.
An implementation that updates only address-independent fields and keeps the old `admin = 1`
lets a cookie-channel add of a removed admin's address restore admin rights. The cookie channel is
exactly the channel the privileged-fields table says must never set the flag.

Fold: R-6d10 states that reactivation writes `admin` from the validated candidate, which the channel
rule has already forced to false on the cookie channel. Add a Task 3 row: "a tombstoned admin row,
re-added on the cookie channel, reactivates with admin 0." Add the mutation "keep the stored admin
flag on reactivation." Add a `users.admin` line to the privileged-fields table naming reactivation.

### M5. The cookie actor's admin check and the at-least-one-admin rule are dropped

Location: R-6d10 `:362-367`, Task 3 table `:641-655`.

Today `AddMember` also refuses a cookie-channel actor who does not hold the admin flag in the copy
re-read under the lock, and refuses any write that leaves no admin (`roster.go:751-758`,
`roster.go:795-807`). R-6d10 carries over only "the admin flag rule," and the state table has no
row for either check. A store-backed `AddMember` that drops the actor check lets any signed-in
member reach the write, since only `RequireAdmin` in front of it would still guard it.

Fold: R-6d10 names both checks, read inside the `Update` callback (the read-pool doc at
`conn.go:107-111` says a read a write decides on belongs there). Add rows: a cookie actor with no
live row, a non-admin row, or a tombstoned admin row refuses with `errActorNotAdmin`. A
command-line add of a non-admin into a table with no live admin refuses. Mutations: read the actor
through `Handle.QueryRow` outside the transaction; count a tombstoned admin.

### M6. Tombstoned rows still hold both `UNIQUE` columns, and no row covers it

Location: R-6d10 `:363-365`, Task 3 table `:650-651`; schema `0003_catalog.sql:3,5`.

`users.address` is `UNIQUE` and `users.streamer_username` is `UNIQUE COLLATE NOCASE` across every
row, tombstoned ones included. Three consequences go unstated. First, a reactivation that reuses
its own old username must not collide with itself. The reactivation test passes vacuously if its
fixture picks a new username, and the real re-add fails. Second, a new member asking for a
tombstoned member's username must refuse by name in Go. Otherwise the raw constraint error reaches
`additionOutcome` and the API answers 500. Third, an import entry whose address matches a tombstoned
row has no stated outcome.

Fold: add Task 3 rows for a reactivation with the same username (succeeds, same id), and a new
address requesting a tombstoned row's username in a different ASCII case (refused naming the
username, never a constraint error). Add one API-level test through the roster API for the second
row, asserting the duplicate-username outcome and not 500. That test also covers criterion 6's
"the page refuses ... by name" half, which no task currently tests after the store swap. Add a Task
5 row: an entry matching a tombstoned address is refused naming the entry. Import is not a
reactivation channel.

### M7. Criterion 1's admin proof passes when the flag is dropped

Location: plan `:756-758`; spec criterion 1.

"The imported admin then runs `member add --admin` for a new member successfully" proves nothing
about the import. The command-line channel checks no member's flag: its actor is an OS user
(`resolveActor`), and `addTo` checks the actor's admin flag only off the command line
(`roster.go:796-804`). The assertion holds even if the import writes every entry as non-admin.
"Writes the record" also names no assertion on the record's content.

Fold: assert per-row flags: 1 on the admin entry and 0 on the two others, so both "all admin" and
"none admin" mutations fail. Prove the grant path through the cookie channel. The imported admin's
session adds a member through the roster API. An imported non-admin's session gets the gate's 404.
Assert the record's hash equals `sha256.Sum256` of the fixture file's bytes, its count is 3, and its
actor is `resolveActor()`'s value.

### M8. Two of Task 5's named mutations have no failing fixture

Location: plan `:769-770`, `:759-762`.

"Let `--dry-run` enter `Update` without rolling back" fails only against a dry run over a valid
file. Every refusal fixture rolls back anyway, and no criterion names a valid-file dry run. "Write
the record outside the transaction" fails only if a statement fails inside the transaction after
Go-side validation passes. No fixture forces that.

Fold: add a criterion for a dry run over a valid three-entry file. It exits 0 and reports the three
additions. `users`, `member_imports`, and the settings generation are unchanged. Add a fault
fixture: the test creates a SQLite trigger on its temp store (`CREATE TRIGGER ... BEFORE INSERT
ON users ... SELECT RAISE(ABORT, ...)`) on the third insert. The test asserts the import fails,
`member_imports` holds no row, and `users` holds none of the file's entries.

### M9. The import's field validation has no state, and it admits usernames the write path refuses

Location: Task 5 table `:742-753`, R-6d11; `roster.go:537-596`.

The import parses through `validateRoster`, whose field check is `memberRecord.validate`. That
check admits any ASCII-printable username, such as `a b` or `a/b`, while the write path demands
`streamerUsernamePattern` (`roster.go:52`, `:602-620`). After 6d every `users` row is a principal
rung 12 injects into the proxy header, so an import channel with a looser shape than the add
channel splits one field's contract. The table also has no row for an entry that fails field
validation (bad address, refused rune, overlong field, blank display name). Finally, spec rows say
"naming both entries," while `validateRoster` deliberately names positions only and never an
address (`roster.go:534-536`).

Fold: the import validates each entry with the write path's rules (`validateWriteCandidate`),
which is the one correct answer: one field, one shape. Add a row: "an entry failing a field rule is
refused naming its position and the rule." State that "naming both entries" means both positions,
under the no-address-in-errors rule. Quote `validateRoster`'s doc in the behaviors table.

### M10. The `deploy.sh` promises have no proof mechanism

Location: Task 8 `:880-882`, `:902-904`, `:909-911`; HISTORY 6c handoff item 5; repo CLAUDE.md
("Shell ... no behavior tests of its own").

Criterion 14 says `deploy.sh --check` "names the store leaf's pre-create ... and shows the owner-run
`check` before the restart step." The repo rules that shell carries no behavior tests, and no gate
leg runs `deploy.sh`. The criterion therefore has no stated mechanism and passes by assertion. The
halt on `check`'s nonzero exit, handoff item 5, has no proof at all. It also depends on
`systemd-run --wait` propagating the exit status, which the plan does not quote. The named mutation
"grant the leaf with a `-` prefix" needs the state-dir leg to refuse a `-`-prefixed entry and to
check the pre-create array. Neither is in Task 8's outcomes.

Fold: make the proof explicit. Extend the state-dir leg in `scripts/check.sh` to assert three
things. The unit's `ReadWritePaths` names the store leaf without a `-` prefix. `deploy.sh`'s
pre-create array names the store leaf and no roster leaf. The owner-run `check` line appears before
the `serve` restart and is followed by a halt on failure. The mutation list names each. Quote
systemd-run(1), EXIT STATUS: "If systemd-run waits for the service to terminate, the return value
will be propagated from the service." Where a promise stays a read (the `systemd-run` property
set), the plan says so and routes it to the diff-reviewer by name.

## Minor

- **m1, reconcile root fixture (`:562-563`, `:567-568`).** The mutation "pass the configured root
  when the record names one" fails only if the record's root differs from the configured root and
  the planted row resolves differently under each. Name that fixture state.
- **m2, the `doctor` string proof (`:603-605`).** Assert the pre-start render first: the stored
  value in force plus the "serve applies it at its next start" note. Otherwise a `doctor` that
  already printed the env value passes. `backup.go:274`'s string gets no test. Name one, or state
  that it shares `doctor`'s evaluation path and quote that path.
- **m3, Task 2 state rows (`:591-597`).** Add the spec's "stored value outside the binary's bounds:
  `serve` fails naming the key and the bound" row (spec `:761`). Then state what the admin-conflict
  refusal names as its resolving action: spec `:759` requires one. At 6d that action is unsetting
  the env value, since spec row `:760` lets the admin value stand.
- **m4, `Identity.Valid` (`:636`).** "`Valid` refuses a zero id on the cookie channel" has no
  criterion or mutation. Add a table row and the mutation "`Valid` ignores `UserID`."
- **m5, Task 4 fixtures (`:698-715`).** The mutation "count tombstoned rows as members" needs a
  tombstoned admin in the "members but no admin" fixture, and none is named. "Two members, so an
  order ... bug shows" names no order now that file order is gone, so state one (for example
  `created_at`, then `id`). Add rows for `List` excluding tombstoned rows and for a store older than
  `0003` (no `users` table). `checkRosterDirOwner` retires, so keep its foreign-uid refusal as a
  `member add` row through the store's owner check, and quote `leaf.statLeaf`
  (`open.go:425-442`), the check the plan's "its uid rule" relies on.
- **m6, Task 5 dry-run aggregation (`:761-762`).** "Three different refusals" should span layers:
  one file-level, one store collision, and one of "no admin" or "record present." Otherwise
  `validateRoster`'s existing `errors.Join` passes it without proving the store checks aggregate.
  The mode-0000 fixture reads fine as root, so name the euid assumption without a skip guard. State
  whether a real import reports a differing display name or keeps it silently.
- **m7, Task 7 reader errors (`:855-856`).** Name which reader fails, one per apply mode at least.
  The nonce sweep has no request to fail, so name its report (log the key and skip the sweep, no
  default).
- **m8, unquoted behaviors (planning-miss item 1).** `Handle.QueryRow`/`Query` (`conn.go:107-121`)
  back `Lookup` and `List`. The store's owner check (`leaf.go`, `open.go:425-442`) is behind "its
  uid rule." `systemd-run`'s exit status (M10) and `validateRoster`'s shape (M9) are relied on and
  not quoted.
- **m9, criterion 4 scope (`:765-766`).** The spec says "every verb"; the plan tests four. Either
  quote that `AddSuperseded`'s rejection runs in the one wall load every verb shares, or iterate the
  registered command tree in one table test.
