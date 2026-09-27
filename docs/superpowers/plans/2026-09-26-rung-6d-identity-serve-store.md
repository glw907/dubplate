# Rung 6, Pass 6d (Identity, the Handoff, and `serve`'s Store) Implementation Plan

Status: **revision 2, 2026-09-26; reviewed (contract, mechanics, and risk lenses) and folded, no
ruling answered.** History: revision 1 at `5afe46b`; revision 2 folds the three lens reviews per
`docs/superpowers/research/2026-09-26-rung-6d-plan-fold.md`. It awaits Geoff's approval gate. It is drafted
ahead of its execution: 6d executes only after rung 5 closes (R-6d1), and 5d-i and 5d-ii will rewrite
most of the files this plan names first. **Task 0 therefore re-resolves every pointer against the tree
5d leaves, and a pointer whose content is gone is a blocker, not a note.**

**Ceiling (Ruling 1):** 11.5M execution, flag 9.2M; planning recorded against the rung and outside it.
**Checkpoint interval:** each segment boundary (at most three tasks apart), and the close.
**Segments:** A (Task 0, Tasks 1 and 2), B (Tasks 3, 4, and 5, then the security read of B), C (Tasks
6, 7, and 8, then the security read of C, the simplifier round, and the merge), then Task 9, the close.
**Tasks:** ten (Task 0, eight build tasks, and the close), at the spec's ten-task line (spec `:1149`)
and not past it. A "no" on Ruling 2 or a "yes" on Ruling 3 adds a task and splits the pass (Ruling 1).

Sources, in order of authority, cited at `20f4a7c` on master:

- The ratified rung 6 spec (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`, revision
  2.2 with errata E1 to E11, "the spec"): the 6d pass section and criteria 1 to 14 (`:1086-1117`),
  "The settings home" and "The seed handoff" (`:718-800`), "Identity in the store" (`:802-874`),
  "Three open modes" and "Migrations" (`:290-431`), "Package territory" (`:193-215`), "The 2.0 forward
  hooks" (`:877-893`), "Prior art" (`:895-917`), "Planning-miss prevention" (`:919-939`), "Band and
  ceilings" (`:1121-1150`), "Contested surfaces" (`:1152-1161`), and "Not verified" (`:1251-1265`).
- The configuration homes spec (`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`,
  CH), above all the `sensitive` paragraph (CH `:143-150`) and the routed list.
- What earlier passes handed 6d: `docs/STATUS.md`'s "To 6d" line; `docs/HISTORY.md`'s 6c entry ("The
  6c to 6d handoff" items 1 to 6, "To 5d and 6d", errata 6c-E5, 6c-E6, 6c-E10); 6b-ii-b's "To 6d"
  (`ServeOptions.Reconcile`, the store identity, the network actor); 6b-i's "To 6d" (the store
  identity, R9 option b); 6a's open items (serve's settle loop, the logger guard; the migration wait
  registering at 6d, E1).
- The 6c architecture reads (`docs/superpowers/research/2026-09-26-rung-6c-architecture-reads.md`):
  `internal/engine/store` finding 1 (`OpenServe` has no production caller; `OpenWrite`'s "start serve
  first" cannot be satisfied on a fresh box) and `internal/settings` finding 1 (the apply is unexported
  and uncalled while `cmd/dubplate-server/backup.go:274` and `store.go:280-284` promise serve applies
  settings at its next start). 6d wires both.
- The 0.8 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, which still says
  2.0): "Forward hooks the current build owes" (`:683-768`) and "Prior art" (`:770-806`).
- The 5d plan (`docs/superpowers/plans/2026-09-25-rung-5d-boundary-deployment-close.md`, revision 2.2,
  approved): its serve unit, `deploy.sh` rework (Task 4D), runbook (Task 7, R-5d15), and the sitting's
  step 13 (`:1370-1376`), which already treats the pre-6d seed string as a stop.
- The repo `CLAUDE.md`, the 6c plan (`2026-09-25-rung-6c-routed-schema-settings-home.md`) for shape.

**What this pass is.** It makes `serve` open the store through `store.OpenServe`, creating or
migrating it, finishing a reconcile-pending restore, and then applying the seed rule before the
listener opens (spec criteria 11 and 12, the two wiring items the 6c reads route here). It moves the
roster into the store: `internal/auth`'s roster reads and writes the `users` table, the resolution
point returns `Identity.UserID`, and `member import` moves a roster file in once (criteria 1 to 6).
It re-homes rung 5's ten auth thresholds as store-homed settings with accessor reads at every call
site (criteria 9 and 10). It lands the deployment side: the store leaf's grant, `deploy.sh`, the
state-dir leg, the template, and the runbook's install and upgrade order (criteria 13 and 14).

**What this pass is not.** It deploys nothing. It builds no HTTP route, no admin write path (rung 14),
no member removal, demotion, or revoke verb, no web page change, no `ops/local` runner inside `serve`,
and, under Rulings 2 and 3's recommended answers, neither the dual-key epoch nor the store identity.

---

## Rulings

Three questions, ordered by how much build each answer changes. Every other decision is a Resolution
below, taken on its architectural merit and reported.

**1. Does 6d run as one pass under an 11.5M execution ceiling, flag 9.2M, in three segments?**
Recommendation: yes. Reason: the ceiling is derived from 6c's actual rate, which ran at about 1.9 times
the spec's priced rate, and ten tasks sit at the spec's split line without crossing it ("Token
ceiling"). What yes builds: one pass, three segments, one close. What no builds: the conductor re-cuts
at the B/C seam into 6d-i (Tasks 0 to 5 and a close) and 6d-ii (Tasks 6 to 8 and a close), about 0.5M
more for the second Task 0 and close. If Ruling 2 is no or Ruling 3 is yes, the task count passes ten
and that split happens regardless, with each half's ceiling re-derived from this table.

**2. Does the dual-key epoch (spec criteria 7 and 8) move to the revoke verb, its first writer, rather
than land in 6d?** Recommendation: yes. Reason: `EpochStore.Bump` has no production caller today
(`grep -rn '\.Bump(' --include=*.go internal cmd` returns only tests), so no epoch record has ever been
written on any box, and 6d's tests would prove a revocation path nothing in production can reach; the
repo's rule is that a hook is built by its first consumer. Ruling 9's "same id on re-add" also means
the address-to-id resolution never changes in production, which the revoke verb's own plan can weigh
when it decides whether it needs the address key at all. What yes builds: 6d leaves `epoch.go`
untouched; criteria 7 and 8, rung 5's review minor 14, and erratum row 10's "bumps to both keys" route
to the revoke verb's plan by name. What no builds: Task 5b (the dual-key read, both-key bumps, and a
cross-process advisory lock on the epoch directory with a monotonic write), about 0.9M, and the pass
splits under Ruling 1.

**3. Does the store identity (6b-i's R9 option b: a random id minted at create, a remote marker, and
backup and restore refusing another store's remote or snapshot) route to rung 16, where backups first
run on a timer, rather than land in 6d?** Recommendation: yes. Reason: nothing binds it to create,
because a migration can mint the id for an existing store at any later version, and its exposure (two
servers pointed at one remote) starts when backup runs unattended, which rung 16's timer begins; its
consumers are `internal/backup`'s, which 6d otherwise does not touch beyond one exported reconciler.
What yes builds: nothing in 6d; the forward item moves from STATUS's "To 6d" to "To rung 16" with R9's
text. What no builds: a Task 8b (a singleton identity row minted by a migration, the backup-run marker
check refusing a foreign remote, and restore refusing a foreign snapshot), about 0.9M, and the pass
splits under Ruling 1.

---

## Goal

After 6d, a fresh box's `serve` creates the store on its first start; an upgrade's `serve` migrates it
behind a pre-migration copy; the seed rule runs before any request; every member lives in `users`
with a surrogate id the resolution point returns; the roster file is an import source read once by
`member import`; the ten auth thresholds are settings read at their point of use; and the deploy files
grant and pre-create the store leaf. Every operator string that says "serve applies it at its next
start" is true.

## Owner touchpoints

The plan-approval gate and its three rulings, and nothing else. No attended sitting: nothing deploys in
6d. The 6d upgrade (drop the roster row, restart `serve`, `member import --dry-run`, `member import`)
rides the next ruled deploy after 6d merges, in the runbook's words Task 8 writes.

## Token ceiling: 11.5M execution (Ruling 1)

**Basis: 6c's actual** (`docs/HISTORY.md`, 6c "Both budgets"): about 4.5M before the merge for Task 0
(about 0.1M), five build tasks, four fix rounds, and the simplifier round with its review; so (4.5 −
0.1) / 5 ≈ **0.88M per build task**, fix rounds and the simplifier share included. The merge ritual
and six architecture reads cost about 0.75M, about **0.11M per read**.

1. Task 0 (`opus`, fresh context, a wider staleness surface than 6c's): **0.15M**.
2. Tasks 1 to 8 at 0.88M: **7.04M**.
3. Two `web-auth-security-reviewer` reads (after segments B and C) and their fold rounds, the spec's
   own allowance (spec `:1141`): **0.70M**.
4. The close: Task 9's drafts (0.15M), the merge ritual (0.10M), seven architecture reads
   (`internal/engine/store`, `internal/settings`, `internal/config`, `internal/backup`,
   `internal/auth`, `internal/identity`, `cmd/dubplate-server`) at 0.11M (0.77M): **1.02M**.
5. **Projection 8.91M.** 8.91 / 0.8 = 11.14, rounded up to the next 0.5M: **11.5M**, flag at 80%,
   **9.2M**.

Variants: one added task (Ruling 2 no, or Ruling 3 yes) projects 9.79M, ceiling 12.5M, flag 10.0M;
both project 10.67M, ceiling 13.5M, flag 10.8M. Either variant passes ten tasks and splits.

The spec's own row was 6.6M for nine tasks (spec `:1141`), priced at 0.46M per task; 6c's actual
rate is the better evidence and is what CLAUDE.md's "re-derived from the previous pass's actual"
requires (spec `:1147-1149`).

Deliverable count: thirty across Tasks 1 to 8 (Tasks 1, 3, 4, 5, 6, 7, and 8 carry four; Task 2
carries two), plus Task 0's report and Task 9's drafts. No task passes four.

Revision 2 adds no task and no deliverable, so the ceiling stands. It does add state rows and
proofs inside Tasks 1, 3, 4, 5, 7, and 8 (the fold record lists them). The 0.88M rate already carries
6c's fix rounds, and the fold's rows close ambiguities that would otherwise have become fix rounds, so
the projection is unchanged. Task 0's re-derivation from 5d's actual is where a heavier rate shows.

---

## Primary inputs

### Code pointers (at `20f4a7c`; Task 0 re-resolves every one)

- `cmd/dubplate-server/serve.go`: `newServeCmd` (resolves the wall, then `newServeShell`, then
  `serve.Run`), `newServeShell` (opens `auth.Open`, the nonce and epoch stores, `auth.OpenRoster`
  from `DUBPLATE_ROSTER_FILE`, the limiter, the mailer, the seam, the handlers), `authHandlersConfig`,
  `rosterFileKey`. Nothing here references `store` or `settings`.
- `cmd/dubplate-server/auth.go`: `magicLinkSource` (`Required` lists `DUBPLATE_ROSTER_FILE` first),
  the ten auth thresholds (`magicLinkExpiryThreshold` through `unknownAddressAnswerThreshold`), their
  bound constants, `validateMagicLinkSource` (the environ-based bounds and the two relational checks:
  retention at least the expiry; lifetime at most the absolute cap), `authConfig`, `resolveAuthConfig`.
- `cmd/dubplate-server/store.go`: `storeBusyTimeoutThreshold`, `jobRetentionSetting`,
  `storeLocalSnapshotKeepSetting`, `registerStoreConfig`, `storeBusyTimeout`, the settings doctor rows
  and `settingShown` (the sensitive-withholding branch at `:331-336`).
- `cmd/dubplate-server/backup.go`: `logSettingsEvaluation` (`:262-275`, "serve applies it at its next
  start"), `runBackup`'s `store.OpenWrite` and `settings.NewAccessor`.
- `cmd/dubplate-server/member.go`: `member add` (`--name`, `--streamer-username`, `--admin`),
  `member list`, `addMember` (`auth.OpenRoster`/`auth.CreateRoster`), `checkRosterDirOwner`,
  `resolveActor`. `cmd/dubplate-server/doctor.go:169-173` (the roster row). `cmd/dubplate-server/web.go`
  `privateFiles.roster` (`:112-116`, `:691-693`).
- `internal/engine/store/open.go`: `ServeOptions` (`BusyTimeout`, `MigrationWait`, `Keep`, `Backup`,
  `Logger`, `Reconcile func(ctx, *Handle) ([]byte, error)`), `OpenServe` (refuses a nil logger;
  loops shared wait → `openShared` → exclusive wait → `settle`), `settle`, `finishRestore`, `waitLock`
  (bounded per wait by `MigrationWait`, logging holders). `inspect.go`: `Report` (`Holders` filled for
  a busy store), `Facts` (no member counts). `write.go`: `Handle.Update`, `Tx`, `Row.Scan`.
  `id.go`: `NewID`. Migrations `0001` to `0005`; `users` is in `0003_catalog.sql`.
- `internal/settings`: `apply` and `writeChanges` (unexported, `apply.go`), `Evaluate`,
  `EvaluateFacts`, `Accessor` (`Int`, `Duration`, `Float`; no string read).
- `internal/config/settings.go`: `Setting` (`Kind` refuses `KindString`; `Floor` and `Ceiling`
  required; `Sensitive` documented as "a value a report must not print"), `buildValidator`,
  `CorrelatedRule`. `registry.go`: `AddSetting`, `AddCorrelatedRule`, `AddSuperseded`.
- `internal/backup/reconcile.go`: `reconcile(ctx, h, root) (*report, error)` and `report.render`,
  unexported; `restore.go:629-660` (`restorer.report`: root is the record's `LibraryRoot`, else the
  configured root).
- `internal/auth`: `roster.go` (`OpenRoster(path)`, `CreateRoster`, `Lookup(address)`, `List()`,
  `AddMember(ctx, ...)`, `validateWriteCandidate`, `principalCollision` with `strings.EqualFold`,
  `Member.principal` falling back to the address), `seam.go` (`NewMagicLinkSeam(roster, epoch, ...,
  lifetime, absoluteCap)`, `VerifyIdentity` → `roster.Lookup`), `handlers.go` (`HandlersConfig`'s
  `LinkExpiry`, `SessionLifetime`, `SessionAbsoluteCap`, `BodyCap`, `GenericAnswer`; `Lookup` at
  `:210`, `:366`), `admin.go` (`AdminConfig.BodyCap`, `rosterAPI.add`), `nonce.go`
  (`OpenNonceStore(store, retention, linkExpiryCeiling, now)`), `limit.go` (`NewLimiter(Limits)`).
- `internal/identity/identity.go`: `Identity{Address, Principal, DisplayName, Source, Channel}`,
  `Valid`. `internal/ops/ops.go`: `NetworkActor` ("member:<principal>"), which panics unless
  `id.Valid()` (`ops.go:43-48`). Two test fixtures build a cookie identity with no id:
  `validIdentity()` (`internal/ops/ops_test.go:38-45`) and `ann`
  (`internal/ops/opshttp/helpers_test.go:22-28`).
- `cmd/dubplate-server/rendered_test.go` calls `newServeShell` directly (`:420-454`), so every change
  to that function's signature (Tasks 1, 4, 6, and 7) reaches it.
- `internal/engine/store/leaf.go`: `create` refuses unless the process uid owns the state directory
  (`:66-77`), and `checkOwner`'s refusal names `runuser -u <name> -- dubplate-server <command>`
  (`:113`), a hint four store tests assert (`leaf_test.go`, `inspect_test.go`, `restore_test.go`,
  `open_test.go`).
- Deploy files (5d rewrites all of them): `scripts/deploy.sh`, `systemd/` (5d's serve unit; the
  name is 5d's, `dubplate-server-serve.service` in 5d's sitting text), `env/dubplate-server.env.template`,
  `scripts/check.sh`'s state-dir agreement leg (`:645-660` at HEAD), `docs/server-runbook.md` (5d
  Task 7 creates it).
- `web/src/routes/members/+page.svelte` (read at committed `HEAD` only): the streamer-username input
  is already `required` with a pattern (`:216-224`); 6d changes no `web/` file (R-6d17).

### Behaviors this plan relies on, quoted

| Behavior | Quote | Where 6d relies on it |
| --- | --- | --- |
| `database/sql` `Row.Scan` (`go doc database/sql Row.Scan`) | "If more than one row matches the query, Scan uses the first row and discards the rest. If no row matches the query, Scan returns ErrNoRows." | `Lookup` on an absent or tombstoned address |
| `store.Row.Scan` (`internal/engine/store/write.go`) | "Scan copies the row's columns into dest, as [sql.Row.Scan] does, and records any error but [sql.ErrNoRows], which it returns." | A collision read inside `Update` that finds nothing does not poison the transaction |
| `store.Handle.Update` (`write.go:44-63`) | "commits only when fn returns nil and no statement fn issued through tx failed"; "ctx must carry a deadline"; "The first statement reads user_version, and Update refuses without calling fn when it differs from the binary's version." | Every write: import, add, the seed apply |
| `store.NewID` (`id.go`) | "returns 16 random bytes for a new entity's primary key ... It reports no error because [rand.Read] never returns one and always fills its buffer entirely." | The surrogate id's one mint |
| `crypto/sha256.Sum256` (`go doc`) | "Sum256 returns the SHA256 checksum of the data." | The import record's file hash |
| SQLite `NOCASE` (sqlite.org/datatype3.html) | "the 26 upper case characters of ASCII are folded to their lower case equivalents before the comparison is performed. Note that only ASCII characters are case folded." | `users.streamer_username ... UNIQUE COLLATE NOCASE` agrees with `principalCollision`'s `strings.EqualFold` only because the write path's pattern admits ASCII alone; Task 3 keeps the pattern check before the insert |
| `flock(2)`, `LOCK_SH` (`man 2 flock`) | "More than one process may hold a shared lock for a given file at a given time." | A current-version `serve` start never waits on a running backup; only a migrating start (exclusive) does |
| `flock(2)` conversion (spec `:361-365`) | "Converting a lock ... is not guaranteed to be atomic" | Already honored by `OpenServe`; 6d adds no conversion |
| `systemd.exec(5)`, `ReadWritePaths=` | "Paths in ReadWritePaths= ... may be prefixed with "-", in which case they will be ignored when they do not exist." | The store leaf is granted unprefixed, so `deploy.sh` must pre-create it (CLAUDE.md's standing fact) |
| `runuser(1)` | "runuser defaults to not changing the current directory and to setting only the environment variables HOME and SHELL (plus USER and LOGNAME if the target user is not root)." | `runuser` loads no environment file, so a store-opening verb run from `deploy.sh` needs the rendered env supplied another way (R-6d15) |
| `config.Setting` (`internal/config/settings.go`) | "A string has no bounds, so KindString is refused at registration." `buildValidator` panics on any other kind; `boundedValidator` panics when a floor or ceiling does not parse. | `DUBPLATE_UNKNOWN_ADDRESS_ANSWER` and `DUBPLATE_SESSION_LIFETIME` cannot register as settings unchanged (R-6d7, R-6d8) |
| `Registry.AddSuperseded` (`registry.go:530-544`) | registers "a pre-pivot key with no meaning in the greenfield wall, reason being the message rejectNamespace reports for it"; `requirePrefix` demands `DUBPLATE_` | Retiring `DUBPLATE_ROSTER_FILE` (R-6d12) |
| CH, `sensitive` (CH `:143-150`) | "Some settings are dangerous to change live even though they are not secrets ... The registry marks these `sensitive`. A sensitive write requires fresh re-authentication ... A settings key never holds a credential." | R-6d6 |
| `settings` `outcome` (`internal/settings/evaluate.go`) | "A bound applies to the value in force after the seed, so an out-of-bounds environment-seeded row that the environment re-seeds or removes is healed rather than refused." `Outcome.Value` is the value after this start's seed. | R-6d4: the keep read cannot use `Outcome.Value` for an in-bounds row |
| `settings` `parse` (`evaluate.go:367-378`) | `default: return nil, fmt.Errorf("setting kind %v has no parser", kind)`; `outcome` calls `parse(s.Kind, stored.Value)` on every stored row | R-6d7: a closed-set string needs a `parse` arm |
| `store.Inspect` (`inspect.go`) | returns early with `Facts` nil on a restore intent and on `StateBusy`; `readFacts` runs for any store at `jobsVersion` (2) or above, and reads settings only at `settingsVersion` (4) or above | R-6d4's no-facts states; R-6d13's version guards |
| `Handle.QueryRow` (`conn.go:107-111`) | "A read on the read pool sees the last committed state but holds no lock against a write that follows it, so it serves reports; a read that a write decides on runs inside that write's [Handle.Update] callback instead." | `Lookup` and `List` read the pool; every read `AddMember` and `member import` decide on runs inside `Update` |
| `users` schema (`0003_catalog.sql:1-10`) | `address TEXT NOT NULL UNIQUE`, `streamer_username TEXT NOT NULL UNIQUE COLLATE NOCASE`, `tombstoned_at INTEGER`, `created_at INTEGER NOT NULL` | Both uniqueness constraints reach tombstoned rows (R-6d10, R-6d11) |
| `readRoster` (`internal/auth/roster.go:318-320`) | refuses a file whose `perm&0o022 != 0` with `errRosterWritableByOthers` | R-6d11: the import keeps the refusal |
| `validateRoster` (`roster.go:529-535`) | "A record's own address or principal never appears in a returned error, only its position." Its field check, `memberRecord.validate`, admits any ASCII-printable `streamer_username` and a blank display name; `validateWriteCandidate` adds "the display name is not blank, streamer_username is required, and its shape is the narrower pattern" | R-6d11: import entries meet the write path's rules |
| `AddMember` (`roster.go`, doc above `AddMember`) | "On any other channel the actor must also hold the admin flag in the disk copy re-read under the lock ... Whatever the channel, the roster written must name at least one admin." (`errActorNotAdmin`, `errNoAdmin`) | R-6d10 carries both into the `Update` callback |
| `RefreshSession`, `clampExpiry`, `VerifySession` (`internal/auth/session.go:138-233`) | `clampExpiry` "sets s.ExpiresAt to the earlier of s.IssuedAt plus lifetime and s.AnchoredAt plus absoluteCap"; `VerifySession` checks "the MAC first, then the absolute cap, then the expiry" | Task 7: a raised cap can admit a session only after a refresh under it |
| `NonceStore.linkExpiryCeiling` (`nonce.go:69-76`) | "mtimeFallback answers with max(retention, linkExpiryCeiling) rather than retention alone, so a record this store can neither read nor parse cannot be swept, and its nonce replayed, before every token that could have produced it has already expired." | Task 7: the argument stays the registered ceiling |
| `/proc/locks` (mechanics lens probe) | two shared `flock` holders on one file list both pids, the reading process's own included | Task 8: `check`'s holder line excludes its own pid |
| `systemd-run(1)`, EXIT STATUS | "If systemd-run waits for the service to terminate, the return value will be propagated from the service." (probed on systemd 259: `--wait --pipe -p EnvironmentFile=` loaded the file and returned the command's `rc=3`) | R-6d15: `deploy.sh` halts on the owner-run `check` |
| SQLite `RAISE(ABORT, ...)` in a `BEFORE INSERT` trigger (fold probe, Python's `sqlite3`) | a trigger `WHEN (SELECT count(*) FROM users) >= 2` raised on the third insert inside `BEGIN IMMEDIATE`; after `ROLLBACK`, both `users` and a record table written earlier in the transaction held 0 rows | Task 5's fault fixture |
| `systemctl(1)`, `try-restart` | "Stop and then start one or more units specified on the command line if the units are running. This does nothing if units are not running." | Task 8: a deploy never starts the serve unit 5d keeps stopped |

Task 0 re-quotes the in-repo rows from the tree 5d leaves; a changed quote is a note naming the new
text, a removed behavior is a blocker.

---

## Global constraints

- **Where it runs.** The main tree (lane 1), after rung 5 closes (R-6d1). The conductor writes the lane
  registry row before Task 0. No other code lane may edit a file in "Contested surfaces" while 6d runs.
- **Go** follows the `go-conventions` skill: `ctx` first, wrapped errors with sentinels matched by
  `errors.Is`, table-driven tests under the assertion-discipline rules (no silent-success fakes, no
  vacuous assertion; each task names its mutations and reports each one failing).
- **Deploy files** follow the `vps-conventions` skill (Task 8).
- **Config keys** are declared through the registry in the rung's own file
  (`cmd/dubplate-server/store.go` for store keys, `auth.go` for auth keys), never in
  `internal/config/keys.go`.
- **Nothing exported ahead of a caller.** Each new export names its production caller in the task
  that adds it; the diff-reviewer checks it. `internal/auth` imports `store`, never `settings`
  (spec `:208-211`); `cmd/dubplate-server` injects each setting as a reader function.
- **The gate**: `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`, the heavy
  lane (the web leg launches Chromium). On exit 75, re-issue the same command until it prints
  `gate exit:`. Never commit red.
- **No deploy, no box write.** Box reads are read-only ssh, and only if Task 0 needs a fact.
- **No `web/` edit** (R-6d17). `cmd/dubplate-server/rendered_test.go` is edited only at its
  `newServeShell` call sites, to follow that function's signature; nothing else in it changes.
- **Commits**: each task commits its own files, Conventional Commits, the repo's attribution footer.
- **Implementers never dispatch agents**; the conductor owns `code-simplifier` and every reviewer.

## Privileged fields and their channels

| Field | Every channel that may set it | Never |
| --- | --- | --- |
| `users.id` (the surrogate id) | the store's `NewID`, in the write path (`AddMember`) and in `member import` | a request body, the roster file, a CLI flag |
| `users.admin` | `member add --admin` (the CLI grant, 5b-ii's rule); `member import`, from the file's `admin` field, once, from a file no other account can write; a reactivation writes it from the validated candidate, so a cookie-channel reactivation writes 0 | the roster API and the add-member page (cookie channel), which refuse it as today; a stored flag surviving a reactivation |
| `users.streamer_username` | the add path on every channel (page, API, CLI), and `member import` | no edit path exists; a reactivation (R-6d10) takes the new value as the add would |
| `users.tombstoned_at` | no channel sets it at 6d (no removal path); the write path's reactivation clears it | any request |
| the `member_imports` row | `member import` alone | anything else |
| `Identity.UserID` | the resolution point (`Lookup`) alone, from `users.id` | a session field, a cookie, a request |
| `Identity.Principal` | the resolution point, from `users.streamer_username` | the session's address |
| settings source `env-seed`, and `default` by removal | `serve`'s seed apply at start alone (now wired) | `check`, `doctor`, `backup`, the member verbs (read-only evaluation) |
| settings source `admin` | none at 6d (rung 14); tests write one in-package | any 6d code path |
| `ServeOptions.Keep` | the registry value resolved through `store.Inspect` and `settings.EvaluateFacts` before `OpenServe` | the env directly |
| `ServeOptions.MigrationWait` | `DUBPLATE_STORE_MIGRATION_WAIT` (wall) | a setting |

---

## Resolutions this plan takes rather than discovers

### R-6d1: placement, after rung 5 closes

6d executes after 5d-ii closes rung 5, in the main tree. The spec gates it on 5d's merge
(spec `:804-805`, `:1086`), and 5d-ii's Task 9 folds write Go in `internal/auth` and
`cmd/dubplate-server`, the files 6d rewrites. The 5d sitting deploys with `serve`'s store unopened:
its unit stays disabled and unrouted, the template sets none of the store-homed keys, and its step 13
already stops on any "seeds at serve's next start" row (5d plan `:1370-1376`). So the 6c reads' rule,
"6d must land before any deploy that exposes the strings," holds if **6d lands before any deploy that
enables the serve unit or routes traffic to it.** STATUS's line "5d's deploy gate checks that 6d's
two wiring items have landed" contradicts the approved 5d plan and the spec's order; the conductor
rewrites it to that rule at plan approval.

### R-6d2: `serve`'s start order

`newServeCmd` runs: the wall load; the serve and store thresholds; the local-keep read (R-6d4);
`store.OpenServe` (create, complete, or migrate; finish a reconcile-pending restore); the seed apply
(R-6d5); an accessor over the handle; the auth construction, where every restart-mode setting is read
once (R-6d9); then `serve.Run`. The handle lives until `Run` returns and is closed then, releasing the
shared lock. Everything before `serve.Run` stays inside the one preparation function the verb calls
(today `newServeShell`, which gains the store open and the seed apply), and `serve.Run` stays the
only call in the verb that opens a listener (`internal/serve`'s `Listen`). The start-order proofs
therefore test that function: its return is the moment the listener could first open, and a refusal
returns an error with no shell, which leaves `serve.Run` nothing to run. No test-only listen seam is
added. The spec's "then recovery" (spec `:872-874`) reads as `OpenServe`'s own staging cleanup
and restore finish; job-record recovery is `ops/local`'s, and `serve` runs no job until rung 14 mounts
the first job route, so `serve` builds no runner at 6d. A store with an empty `users` table starts:
every login then answers as an unknown address, which is spec ruling 1's stated outage window.

### R-6d3: the migration wait

`DUBPLATE_STORE_MIGRATION_WAIT` registers as a wall threshold in `cmd/dubplate-server/store.go`
beside the busy timeout (spec `:795-796`, E1). Default **5m**, floor 1s, ceiling 24h. Reason: a
current-version start waits only on an exclusive holder (a restore), since shared locks coexist
(quoted); a migrating start waits on shared holders, a backup's sync above all, which on a first run
lasts hours. Five minutes outlasts every member verb and every snapshot copy a group-scale store
takes, and is far shorter than a first sync, whose holder `deploy.sh` names before the restart
(Task 8). On expiry `serve` exits nonzero naming the holders, and the unit's `Restart=` policy (Task
0 reads 5d's value) retries. The bounds are range checks in the registering file, on
`storeBusyTimeout`'s precedent.

### R-6d4: the local keep reaches `OpenServe` through the facts

Per 6c-E10 and handoff item 1: before `OpenServe`, `serve` runs `store.Inspect` and
`settings.EvaluateFacts`, and passes `cfg.IntegrationEnabled` of the backup class as
`ServeOptions.Backup`. `ServeOptions.Keep` is extracted by one rule, on `settingsOutcomeStatus`'s
precedent in `store.go`, and never from `Outcome.Value` for an in-bounds row, because that is the
post-seed value (quoted):

- **Facts present and a keep row stored, in bounds** (`storeLocalSnapshotKeepSetting.Validate`
  passes): `Keep` is the row's value, even when the env differs. An env value that differs reaches
  the keep after this start's seed, the spec's stated order (spec `:411-414`).
- **Facts present and a keep row stored, out of bounds:** the key's outcome decides, under 6c's
  R-6c7 and 6c-E8. When the outcome refuses (an `admin` row, which stays in force), the start
  refuses naming the key and the bound, before any migration. When the outcome heals the row (the
  env re-seeds it in bounds, or an absent env removes it), `Keep` is that healed value, which is the
  value in force once this start's seed commits. So a later binary that narrows the bound never
  strands `serve`.
- **No keep row, or no facts** (absent, below `0004`, a restore intent, busy, refused, not a store):
  `Keep` is the registry default, and `OpenServe` decides whether the state itself refuses.
- An unparseable env value for the keep does not refuse here; it is the seed's refusal after the
  open (Task 2).

### R-6d5: the seed apply is exported and called once

`internal/settings` exports the apply (its production caller is `serve`), keeping its one-transaction
contract. A refusal (an `admin` row disagreeing with the env, a stored value in force outside bounds,
an unparseable env value, a correlated rule failing) ends `serve`'s start nonzero with the joined
refusal, and the listener never opens. Each change logs one line after the commit, as `apply` does
today. With it wired, `backup.go:274`'s and `store.go:280-284`'s strings become true; Task 2 re-reads
both and changes a word only if one is still false. Handoff item 6 (a restored `admin` row may refuse
the seed) needs no code: the refusal names the conflict, and its resolving channel is rung 14's.

### R-6d6: `sensitive` means CH's marker, and the value is shown

CH defines `sensitive` as "dangerous to change live even though they are not secrets," gated by
re-authentication, and rules that "a settings key never holds a credential." 6c documented
`Setting.Sensitive` as "a value a report must not print" and withholds it in `doctor` and in
`Outcome.Report`. 6d marks both lifetimes and the absolute cap sensitive (spec `:796-797`); under 6c's
meaning, `doctor` would hide an operator's session lifetime for no reason. So Task 6 re-documents the
field to CH's meaning and removes the withholding, and `doctor` shows the value and the marker. This
is erratum 6d-E1 to 6c's R-6c3.

### R-6d7: a closed-set string setting

`DUBPLATE_UNKNOWN_ADDRESS_ANSWER` (`kind` or `generic`) is store-homed under the spec's table, but
`Setting` refuses `KindString` because a string has no floor or ceiling. Task 6 admits `KindString`
only with a non-empty list of allowed values, which is its bound: registration panics on an empty list
or a default outside it, the validator refuses any other value naming the set, and `Floor`/`Ceiling`
must be empty for it. `internal/settings`' `parse` gains a `KindString` arm returning the string
itself, so `equal` is exact byte equality; without it, `outcome` refuses every stored string row
(quoted), and the second `serve` start after a string seed refuses (the mechanics lens reproduced
this). The accessor gains a typed string read, whose caller is the handlers' answer
reader. Leaving the key on the wall was weighed and declined: it is a tunable with a default, which CH
homes in the store (CH `:66`).

### R-6d8: bounds and both correlated rules move into the registry

Every auth setting's bounds become `Floor`/`Ceiling` data, from the constants `auth.go` holds today.
`DUBPLATE_SESSION_LIFETIME` has no ceiling today; it takes the absolute cap's ceiling (two years),
since the lifetime can never usefully exceed the cap and the correlated rule enforces that anyway.
Both relational checks (`validateMagicLinkSource`'s retention-at-least-expiry and
lifetime-at-most-cap) become registry correlated rules, each error naming both keys, and 6c's
test-only rule pair (the 6c reads' config finding 4) is replaced by the real pair; spec criterion 5 is
re-run over it (6c-E6, handoff item 2). `validateMagicLinkSource` keeps only the source keys it owns
(secrets, public URL, SMTP). Consequence, as the spec rules: a bad auth value no longer fails every
verb's load; `doctor` and `check` report it, `backup` logs it, and `serve` refuses at the seed.

### R-6d9: apply modes, per the spec's table

`DUBPLATE_SESSION_ABSOLUTE_CAP` ("immediately") is read at every verification, in both directions.
The link expiry, the session lifetime, the nonce retention, the body cap, and the unknown-address
answer ("next operation") are read when the operation starts, before any write it makes (6c's
settings-read-before-the-write rule). The four rate limits ("restart") are read once at start, after
the seed, and never again (handoff item 4); the limiter is built from those values. A read that fails
at request time fails the request through the handler's existing server-error path and logs the key;
it never falls back to a default.

Three orderings follow from that rule and are fixed here. In the link request, the answer and the
body cap are read before `Lookup`, so a failed read answers alike for a member and a stranger. In
`consume`, the lifetime and the cap are read before `Nonces.Insert`, so a failed read never burns a
link without minting a session. The nonce sweep has no request to fail: a failed retention read logs
the key and skips that sweep, never defaulting. `NonceStore.Insert` and `Sweep` gain `ctx` (the
accessor's reads take one), and `RunSweep` passes its own. `OpenNonceStore`'s `linkExpiryCeiling`
stays the registered `Ceiling` of `DUBPLATE_MAGIC_LINK_EXPIRY`, never the value in force, because
`mtimeFallback` relies on it outliving every token (quoted).

### R-6d10: the users-backed roster

`internal/auth`'s roster keeps its type and methods and gains a constructor over a `*store.Handle`
(the spec's one-constructor substitution, R5S `:194-196`). Its methods take `ctx` first, which the
spec's "without touching a call site" did not foresee: a database read needs a context
(`go-conventions`), and every caller already holds a request or command context. This is reading
6d-R2 in the deviations table.

- `Lookup` reads the live row (`tombstoned_at IS NULL`) by normalized address; no row is
  `ErrMemberNotFound`; a store error is returned as itself, never as not-found.
- `AddMember` runs in one `Update`. Before it, `checkAddition` runs as today (the admin grant's
  channel, the actor's shape, `validateWriteCandidate`). Every read the write decides on then runs
  inside the callback, never through `Handle.QueryRow` (quoted), in this order:
  1. **The actor's authority** (today's `addTo` rule): on any channel but the command line, the
     actor's live row must exist and hold the admin flag, or the write refuses with
     `errActorNotAdmin`. A tombstoned admin row is no authority.
  2. **The address:** a live row refuses as today's duplicate. A tombstoned row is **reactivated
     under the same id** (spec ruling 9), with every field written from the validated candidate,
     `admin` included, so a cookie-channel reactivation of a removed admin writes `admin` 0.
  3. **The username:** any other row, live or tombstoned, whose `streamer_username` matches
     case-insensitively refuses naming the username, because the `UNIQUE COLLATE NOCASE` constraint
     reaches tombstoned rows (quoted) and a raw constraint error would answer 500. The reactivated
     row's own old username is not a collision.
  4. **At least one admin** (today's `errNoAdmin`): the table after the write must hold a live
     admin, so a non-admin `member add` into an empty table refuses naming `member add --admin`.

  A new row gets `store.NewID`. The admin grant rule carries over unchanged (5b-ii: the cookie
  channel refuses it). The audit record stays the log record (`logAddition`). The deadline `Update`
  requires is already set by both callers (`admin.go:255`, `member.go:95`).
- **No last-admin guard is built.** The spec's "5c builds the last-admin guard, and 6d re-establishes
  it in the write transaction" (spec `:182-183`) presumes a demotion or removal path; none exists
  (`grep -rn -i 'demot\|RemoveMember' internal/auth cmd` finds only comments), so the guard and
  criterion 6's "concurrent self-demotion of the last two admins" route to the first verb that can
  remove or demote. Reading 6d-R3.
- `List` returns live rows only, ordered by `created_at` and then `id`, since the file's order is
  gone.
- `Member` gains the id; `Member.principal`'s address fallback stays for no row, since every `users`
  row carries a username.

### R-6d11: the import record and the import's states

Migration `0006` adds `member_imports`: a `BLOB` `id`, `at`, the file's SHA-256 as a `BLOB`, the
member count, and the OS actor, `STRICT`, no `kind` column (so invariant 3 does not apply), no path
column (the file's path is not stored), and no change to invariant 6's exempt list. "An import record
already present" means any row. The match rule is the spec's: an entry whose address, streamer
username, and admin flag equal an existing live row is that member and changes nothing; the display
name is not part of the match, and a differing one is reported by both `--dry-run` and the real
import and left as stored. The check for an existing record runs inside the `Update` callback, so two
concurrent imports serialize on the write lock and exactly one succeeds.

The import reads the file the way `readRoster` does: it opens and stats it, refuses a mode with
`0o022` set naming the file and the mode (quoted: the import is an admin-flag channel, and another
local account able to write the file could otherwise add itself as an admin), and hashes the same
bytes it parses. Each entry then meets the write path's rules (`validateWriteCandidate`), since the
rows it writes are write-path rows and rung 12 injects the username into a proxy header: one field,
one shape. Every report names an entry by its position and the field, never its address or username,
under `validateRoster`'s rule (quoted), so "naming both entries" means both positions. Task 4 keeps
`parseRoster`, `validateRoster`, and `memberRecord` for this verb.

Four states join the spec's nine. **A file listing no members is refused naming the file**: an empty
import would write the record and block every later import while adding no one. **A file writable by
group or others is refused.** **An entry failing a field rule is refused naming its position and the
rule.** **An entry whose address matches a tombstoned row is refused naming its position**: the
import is not a reactivation channel, and the `UNIQUE` address constraint would otherwise surface as a
raw store error. Reading 6d-R4.

### R-6d12: `DUBPLATE_ROSTER_FILE` retires through `AddSuperseded`

In the same commit as the verb (spec `:815-817`), the key leaves `magicLinkSource.Required` and
registers through `AddSuperseded` with the message "superseded: the roster lives in the store; move a
roster file in once with dubplate-server member import --roster PATH". `member import` takes the path
as a flag, never from the env. `web.go`'s `privateFiles.roster` entry goes with the key, since the
store leaf sits under the state directory, which the same check already refuses.

### R-6d13: `doctor`'s roster row becomes a members row

Read through `store.Inspect`'s facts (a member count, an admin count, and the import record's time
when present), never by opening the store for writing. Each read is guarded by a version constant
beside `settingsVersion`, on `SettingsAbsent`'s precedent: the counts need `users` (version 3), the
import time needs `member_imports` (the new migration's version). An unguarded query would turn every
older store into `StateRefused` (quoted), failing `check`'s store row and so `deploy.sh`'s
pre-restart halt. Its states are Task 4's table.

### R-6d14: the network actor stays `member:<principal>`

`ops.NetworkActor` has no production caller until rung 14 mounts the projection, so 6d changes
nothing in `internal/ops`. The routed note for rung 14: once an admin can edit a streamer username,
a job's actor should key on `UserID`, because a renamed principal breaks the actor's correlation.

### R-6d15: store-opening verbs from `deploy.sh`

The spec says `deploy.sh` runs every store-opening verb "through `runuser -u <owner>`" (spec `:251-252`,
criterion 14). `runuser` sets only `HOME` and `SHELL` (quoted) and loads no environment file, while
the verbs need the rendered env. So `deploy.sh` uses the operator form 5d's runbook already uses:
`systemd-run --wait --pipe --uid=<owner> --gid=<owner> -p EnvironmentFile=<env>` with the unit's
filesystem properties, which makes the same uid change and loads the env. `--wait` propagates the
verb's exit status (quoted, and probed), which is what lets `deploy.sh` halt on `check`. The store's
own owner refusal (`leaf.go:113`) names `runuser`, which would route an operator to a command that
then fails the wall load; Task 8 rewords it to "run as <name> with the rendered environment" and the
four store tests that assert the hint follow. Reading 6d-R5.

### R-6d16: `OpenServe`'s loop and the reconciler

The reconciler: `internal/backup` exports one function whose production caller is `serve`, over the
root the restore-intent record names when it names one and the configured library root otherwise,
matching `restore.go:633`. The report file carries the same body `restore` writes and the same head:
`restorer.report` prefixes the stamp, a `source:` line built from the intent's remote, snapshot, or
file, and the displaced line (`restore.go:629-650`). So `ServeOptions.Reconcile` receives the handle,
the stamp `finishRestore` already holds, and the `store.RestoreIntent` (library root and source);
the signature may change, as it has no production caller yet. The settle loop (6a's open item): each
round needs another process to win a lock conversion by completing a restore swap, which is an
operator action, and each wait is already bounded. The loop re-enters once after every successful
settle by design (`open.go:248-269`), so 6d warns only on a second or later settle within one
`OpenServe` call, naming the reason; a plain create or migration logs no such line. No iteration
bound is added, since a bound would refuse a legitimate operator sequence. The nil
logger guard already exists (`OpenServe`'s first line); Task 0 confirms it. Both 6a items close.

### R-6d17: no `web/` change

The add-member page already requires a streamer username with the write path's pattern (committed
`HEAD`, `web/src/routes/members/+page.svelte:216-224`), and the roster API's JSON shape does not
change. The page's comment that "an older member's `streamer_username` can be blank" (`:66-68`)
becomes stale once every member carries one; it routes to lane 1's next `web/` pass as a one-line
note, since `web/` is not 6d's territory.

### R-6d18: the runbook

5d's Task 7 creates `docs/server-runbook.md`, although the rung 6 spec expected rung 16 to author the
runbook. A runbook telling an operator to `member add` before `serve` starts is false after 6d, so
Task 8 updates its install order and adds the 6d upgrade section, and records the remaining owed lines
for rung 16. The upgrade section keeps the roster file until 5d's `.prev` copies are superseded by a
later attended deploy, since rolling back past 6d needs the old template and the old file (spec
`:834-836`); only then does it tell the operator to delete the file.

### R-6d19: `deploy.sh` never starts the serve unit

5d installs the serve unit disabled and stopped, and its `deploy.sh` never enables or starts it (5d
plan R-5d10, Task 4's outcomes). `systemctl restart` starts a stopped unit, which would put `serve`
on the production box without a ruling, holding the shared lock and answering logins against an
empty `users` table. So `deploy.sh` restarts `serve` with `systemctl try-restart` (quoted: it "does
nothing if units are not running"). Its steps run in this order: render the env, install the binary,
run the owner-run `check` (R-6d15), halt on a nonzero exit, then `try-restart`. `check` must see the
new binary and the new env, or the superseded `DUBPLATE_ROSTER_FILE` in a stale env would halt the
deploy, and a bad new env would slip through. Until rung 16 enables the unit, `try-restart` is a
no-op on the box, and the holder line still prints from `check`.

---

## Spec deviations and readings that take effect on approval

| # | Spec text | 6d reading | Reason |
| --- | --- | --- | --- |
| 6d-R1 | STATUS "5d's deploy gate checks that 6d's two wiring items have landed" | 6d lands before any deploy that enables or routes to the serve unit | R-6d1 |
| 6d-R2 | "The store moves behind the roster's interface without touching a call site" (`:853-855`) | The constructor is the one substitution; the methods gain `ctx` | R-6d10 |
| 6d-R3 | Criterion 6's last-admin clause; `:182-183` | Routed to the first demote or remove verb | R-6d10 |
| 6d-R4 | The import's "nine states" (`:820-831`) and criterion 2's "eight import refusal states" | Thirteen: an empty file, a file writable by others, an entry failing a field rule, and a tombstoned address each refuse; "naming both entries" names positions | R-6d11 |
| 6d-R5 | Criterion 14's "through `runuser`" | `systemd-run --uid` with the env file | R-6d15 |
| 6d-R6 | "then recovery" (`:872-874`) | `OpenServe`'s staging cleanup and restore finish; job recovery with rung 14 | R-6d2 |
| 6d-E1 | 6c's `Setting.Sensitive` ("must not print") | CH's marker; the value is shown | R-6d6 |
| 6d-R7 | Criteria 7 and 8 (Ruling 2 yes) | Routed to the revoke verb | Ruling 2 |
| 6d-R8 | STATUS "To 6d: store identity" (Ruling 3 yes) | Routed to rung 16 | Ruling 3 |
| 6d-R9 | Criterion 9's "with the env holding A and an in-package admin row holding B ... a limiter built at the next start uses B" | At the next start the env is unset or set to B; env A beside admin row B refuses the start (the seed rule, spec `:759`), and that refusal is its own sub-case | Task 7 |
| 6d-R10 | Criterion 14's "names the holder before restarting `serve`" | `check` names the holder; `deploy.sh` then runs `try-restart`, a no-op until rung 16 enables the unit | R-6d19 |
| 6d-R11 | "Restoring a pre-6d snapshot under a 6d binary leaves the users table empty, and the report says to run `member import`" (`:715-716`) | Not built: no pre-6d store or snapshot can exist, since `OpenWrite` refuses an absent store and 5d never starts `serve` | Fold record |
| 6d-R12 | Criterion 1's "an imported admin can grant admin through the CLI" | Proved as the imported flags per row plus the imported admin's authority on the cookie channel; the CLI grant checks no member's flag, so the literal clause proves nothing about the import | Task 5 |

---

## Review focus

The `diff-reviewer` reads each task against its criteria. The conductor dispatches
`web-auth-security-reviewer` (Opus) after segment B (Tasks 3 to 5) and after segment C (Tasks 6 and
7), each briefed with this plan's privileged-fields table and asked for correctness gaps ranked by
consequence, not polish; a fold task may refuse a finding whose fix costs more than the risk it
removes, with a one-line reason. Checklist lines for both:

- `Identity.UserID` has one producer (the resolution point); `Principal` is never the id.
- The admin flag reaches `users` only through the CLI grant and the import; a cookie-channel
  reactivation writes 0; the import refuses a file writable by group or others.
- `AddMember` re-reads the cookie actor's live admin flag and enforces at-least-one-admin inside the
  `Update` callback, never through the read pool.
- A tombstoned address re-adds under its old id; a live one refuses; a username collision refuses by
  name, case-insensitively, before the insert.
- The seed apply runs after `OpenServe` and before `serve.Run`, and a refusal leaves no listener.
- The absolute cap is read at every verification; the four rate limits are read once, after the seed.
- No auth reader falls back to a default on a read error; the answer is read before `Lookup`, and the
  lifetime and cap before `Nonces.Insert`.
- `internal/auth` imports `store` and never `settings`.
- The generation-first read and the settings-read-before-the-write rule (6c HISTORY) are untouched.
- A migration column's meaning: no path column, no source id under another name (the spec's standing
  reviewer line, `:463-464`).

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context. **Its output is blockers with evidence.** It writes no code and edits
no file. It grades a moved pointer by content: content that survives under a new name or offset is a
note naming the new site, and content that is gone is a blocker.

**Blockers: a "no" stops the launch.**

1. **Has rung 5 closed?** Report the 5d-ii HISTORY entry's commit, `git log -1` on master, the lane
   registry, and `git status` (no warm change under any file in "Contested surfaces").
2. **Did 5d land what 6d edits?** Report, by symbol: 5d's serve unit (file name, `Type=`, `Restart=`,
   `RestartSec=`, `ReadWritePaths=`, `EnvironmentFile=`), `deploy.sh`'s leaf pre-creates (the array
   the 5d plan says 6d extends by one word, 5d plan `:544`), its check-by-default and `--apply`
   modes, whether it runs `check` before restarting `serve` and halts on a nonzero exit (6c's Ruling
   2, owed by 5d), how it restarts `serve` if it does at all (R-6d19), the order of its env render,
   binary install, and restart steps, the state-dir leg in `scripts/check.sh`, the template's `DUBPLATE_ROSTER_FILE`
   row, and `docs/server-runbook.md`'s install-order section. A missing 6c Ruling 2 halt is not a
   blocker; it becomes a Task 8 outcome, reported.
3. **Do `serve.go`, `auth.go`, `member.go`, `doctor.go`, and `web.go` still carry every symbol the
   "Code pointers" list names?** 5d-i adds the trusted proxy and probe rows; report each changed
   signature, above all `newServeShell`'s and `authHandlersConfig`'s.
4. **Is the highest migration `0005`?** Report `ls internal/engine/store/migrations/`; Task 3 numbers
   from it.
5. **Do `OpenServe`, `ServeOptions`, `settings.apply`, and `EvaluateFacts` match the quoted
   contracts?** Quote `OpenServe`'s nil-logger refusal and `ServeOptions.Reconcile`'s doc.
6. **Is `Bump` still without a production caller, and does no demote or remove path exist?** Report
   both greps (R-6d10, Ruling 2's premise). A production `Bump` caller is a blocker under Ruling 2's
   yes.

**Notes, not blockers:**

7. Re-quote every in-repo row of "Behaviors this plan relies on."
8. Report the translit pass's edits to `deploy.sh` and `doctor.go` if it merged.
9. Report 5d's and 6c-fix's actual spend; the conductor re-derives the ceiling before Task 1.
10. Report the conductor's recorded outcome of Rulings 1 to 3.

**Deliverable count: 1** (the report).

---

## Task 1: `serve` opens the store

Model: `sonnet`, effort `high`. Depends on Task 0. **0.8 hooks:** consumes hooks 7 and 9 (the
migration runner and the one writer); lands none.

**Files:**
- Modify: `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/store.go` (the migration-wait
  threshold and the keep read), and their tests
- Modify: `internal/engine/store/open.go` (the reconciler's root, the re-entry log line) and tests
- Modify: `internal/backup/reconcile.go` or `restore.go` (the exported reconciler) and tests
- Modify: `cmd/dubplate-server/rendered_test.go`, only its `newServeShell` call sites

**Outcomes:**
- `DUBPLATE_STORE_MIGRATION_WAIT`, R-6d3, with its `doctor` threshold row.
- The preparation function opens the store through `OpenServe` per R-6d2 and R-6d4; the verb closes
  the handle when `serve.Run` returns; any refusal returns an error and no shell.
- The exported reconciler (R-6d16), whose one production caller is `serve`.
- The settle loop's warning on a second or later settle (R-6d16).

**States each check meets, and its report:**

| Store state at start | Report |
| --- | --- |
| absent (no leaf) | created at the binary's version |
| the leaf present and empty (`deploy.sh`'s pre-create, owner the process uid, 0700) | created at the binary's version |
| an interrupted create (zero-byte database, no `-wal`) | completed empty |
| older: at `0004`, an `env-seed` keep row of 1, two pre-migration copies, backup off | migrated; `pre-migrate-v4.db` written; pruned to one copy |
| older: at `0004`, no keep row, `pre-migrate-v1` to `-v3`, backup off | migrated; pruned from four copies to the registry default of 3 |
| older, an `admin` keep row outside its bounds | refused naming the key and bound; the store stays at `0004`, no copy written |
| older, an `env-seed` keep row outside its bounds, the env re-seeding it in bounds | migrated, pruned to the env's value (the healed value, R-6d4) |
| newer than the binary | refused naming both versions |
| a `reconcile-pending` intent over an older store | migrated, the report written into the leaf over the record's root with restore's head, the intent removed |
| any other intent | refused naming `restore --resume` |
| an older store, the lock held shared by another process past the wait | exits nonzero naming the holder's pid; the store stays at its version, no copy written |
| a current store, the lock held shared by another process | opens without waiting |
| the lock held exclusive by another process past the wait | exits nonzero naming the holder's pid |
| `snapshots/` holds a copy and the database is gone | refused naming the newest copy (E11, unchanged) |

**Acceptance criteria:**
- Spec criterion 11: a test calls the preparation function against an absent store and against the
  older-store fixtures above, and asserts the store is at the binary's version when it returns
  (R-6d2: that return is the first moment a listener could open). **Fixture state:** the older store
  carries a populated row in `jobs`, so a migration that silently recreated the store would lose it,
  and the assertion checks it. Every refusal row asserts an error, a nil shell, and the store
  unchanged.
- The three lock rows, with the lock held by a second process (the re-executed test binary the
  store's criterion 14 test uses), so the named pid differs from the test's own; `MigrationWait` set
  to 1s.
- D10's prune proof: the two backup-off fixtures above, each its own `0004` store with more copies
  than the value it prunes to; never a second run, since a current store does not migrate and so
  does not prune (`migrate.go:121-123`).
- The reconcile row asserts the report file carries restore's head (the stamp and the `source:` line)
  and a body naming a planted row without a file, and that the intent is gone. **Fixture state:** the
  record's library root differs from the configured root, and the planted row resolves differently
  under each.
- A plain migration logs no settle warning.
- The gate green.

**Named mutations:** pass a fixed `Keep` of 3 (the keep-row-1 case fails); skip the prune (both
prune rows fail); take `Keep` from `Outcome.Value` for an in-bounds row (the keep-row-1 case,
with the env unset, prunes to the default instead); take the lock exclusive unconditionally (the current-store shared
row blocks); pass the configured root when the record names one (the reconcile row fails); drop the
handle's close (a test that runs `restore` after the verb returns fails on `ErrLockBusy`).

**Deliverable count: 4** (the threshold, the open, the reconciler, the re-entry line).

---

## Task 2: `serve` applies the seed before listening

Model: `sonnet`, effort `high`. Depends on Task 1. **0.8 hooks:** completes hook 15's seed half
(6c built it; `serve` now applies it).

**Files:**
- Modify: `internal/settings/apply.go`, `apply_test.go` (the export)
- Modify: `cmd/dubplate-server/serve.go`, `serve_test.go`; re-read `backup.go:262-275` and
  `store.go:280-284`, editing only a string still false

**Outcomes:**
- The apply exported, called once after `OpenServe` and before `serve.Run` (R-6d5).
- A refusal ends the start nonzero with the joined refusal, before any listener.

**States:**

| Env and stored rows at start | Report |
| --- | --- |
| `DUBPLATE_JOB_RETENTION=720h`, no row | one row, source `env-seed`, one audit row (actor `system`, channel `env-seed`), generation +1 |
| env equal to the stored `env-seed` row | nothing written, generation unchanged |
| env absent, an `env-seed` row present | the row removed, audited |
| an in-package `admin` row, env different | refused naming both values and the resolving action `outcome` already prints ("unset KEY, or set it to V"); no row changed; no shell |
| an in-package `admin` row outside the binary's bounds, env unset | refused naming the key and the bound (spec `:761`, under 6c-E8's value-in-force reading) |
| an env value that does not parse | refused naming the key and the accepted form |

**Acceptance criteria:**
- Spec criterion 12's start half: each row by its own subtest, through the preparation function
  (R-6d2); the request half lands in Task 7. **Fixture state:** a current store, created by Task 1's
  path, with the row each case names; the unchanged row asserts the generation value, never only a
  nil error. When the function returns a shell, the seed's rows are already committed.
- The `doctor` and `backup` strings: a test renders `doctor`'s settings row for an env value that
  differs from the stored row and asserts the pre-start render first (the stored value in force and
  the "serve applies it at its next start" note), runs the preparation function, and asserts the row
  then reads the env's value in force. `backup.go:274`'s line gets the same before-and-after
  assertion through `logSettingsEvaluation`.
- The gate green.

**Named mutations:** move the apply out of the preparation function to after it returns (the seeded
row is absent when the function returns); swallow the apply's error (the refusal row returns a
shell).

**Deliverable count: 2** (the export and call, the refusal path).

**Segment A ends here.** Checkpoint: STATUS (ledger, spend, next task).

---

## Task 3: Members in the store, and `Identity.UserID`

Model: `opus`, effort `high` (the identity surface). Depends on Task 2. **0.8 hooks:** **lands hook
5**: the surrogate id minted at every add, the resolution point returning it as `UserID`, and the
streamer username required and unique. Hook 20 weighed: `member_imports` records no serve event;
neutral.

**Files:**
- Create: `internal/engine/store/migrations/0006_member_imports.sql` (number from Task 0 item 4), its
  populated fixture under `testdata/fixtures/`
- Modify: `internal/auth/roster.go`, `roster_test.go`, `seam.go`, `seam_test.go`
- Modify: `internal/identity/identity.go`, `identity_test.go`
- Modify (test fixtures only): `internal/ops/ops_test.go` (`validIdentity`) and
  `internal/ops/opshttp/helpers_test.go` (`ann`), which gain a `UserID` once `Valid` requires one
- Modify: `internal/auth/admin_test.go` (the roster API's duplicate-username case, below)

**Outcomes:**
- The migration (R-6d11), passing the invariant test with its fixture.
- The users-backed roster (R-6d10): constructor over a handle; `Lookup`, `List`, `AddMember` with
  `ctx`; the file-backed constructor and its file I/O deleted once Task 4 moves the callers (this task
  may keep them only until then, and the diff-reviewer checks Task 4 deletes them).
- `Identity.UserID`: comparable, its zero value meaning unresolved, set only by the seam from
  `Lookup`; `Valid` refuses a zero id on the cookie channel. `Principal` is the streamer username.
- `AddMember`'s write rules inside one `Update`.

**States:**

| Check | State | Report |
| --- | --- | --- |
| `Lookup` | a live row | the member with its id |
| `Lookup` | no row | `ErrMemberNotFound` |
| `Lookup` | only a tombstoned row | `ErrMemberNotFound` |
| `Lookup` | an address that does not normalize | `ErrMemberNotFound`, the address not named (today's rule) |
| `Lookup` | the store closed | the store's error, not `ErrMemberNotFound` |
| `AddMember` | a new address and username | a row with a fresh 16-byte id |
| `AddMember` | a live row on the address | refused as today's duplicate |
| `AddMember` | a tombstoned row on the address, re-added with its own old username | reactivated: same id, new fields, `tombstoned_at` cleared; no collision with itself |
| `AddMember` | a tombstoned admin row, re-added on the cookie channel | reactivated with `admin` 0 |
| `AddMember` | another live row's username, differing only in ASCII case | refused naming the username |
| `AddMember` | a new address asking for a tombstoned row's username, differing in ASCII case | refused naming the username, never a constraint error |
| `AddMember` | no username | refused (`errMissingStreamerUsername`) |
| `AddMember` | admin requested on the cookie channel | refused (5b-ii) |
| `AddMember` | admin on the command-line channel | admin row written |
| `AddMember` | a cookie actor with no live row, a non-admin live row, or only a tombstoned admin row | refused with `errActorNotAdmin`; nothing written |
| `AddMember` | a command-line non-admin add into a table with no live admin | refused with `errNoAdmin`, naming `member add --admin` |
| `List` | an empty table | an empty list, no error |
| `List` | two live rows and one tombstoned row | the two live rows, in `created_at`, then `id`, order |
| `Valid` | a cookie-channel identity with a zero `UserID` | false |

**Acceptance criteria:**
- Spec criterion 5: a verified identity's `UserID` equals the row's id and its `Principal` the row's
  username. **Fixture state:** a member whose address differs from its username, so a `Principal`
  built from the address fails.
- Spec criterion 6, minus the last-admin clause (6d-R3): each `AddMember` row by its own subtest.
  **Fixture state:** tombstoned rows are written directly by the test (no removal path exists); the
  reactivation asserts the id byte for byte and reuses its own old username, so a fixture that picks
  a new username cannot pass it vacuously.
- The page's half of criterion 6 after the swap: one test through the roster API adds a new address
  asking for an existing username in a different case, and asserts the duplicate-username outcome,
  never a 500.
- The invariant test passes with `0006`'s fixture; the table refuses a `NULL` actor.
- The gate green.

**Named mutations:** mint a new id on reactivation; keep the stored admin flag on reactivation;
compare usernames case-sensitively; limit the username collision read to live rows (the tombstoned
username row answers a constraint error); count a tombstoned admin as the actor's authority; drop the at-least-one-admin
check; drop the `tombstoned_at IS NULL` filter in `Lookup`; set `Principal` from the address; make
`Valid` ignore `UserID`. Reading the actor's flag outside the `Update` (through the read pool) is a
race no deterministic test catches; the diff-reviewer checks every read `AddMember` decides on sits
inside the callback, by name.

**Deliverable count: 4** (the migration, the roster over the store, `UserID`, the write rules).

---

## Task 4: The roster's callers move to the store

Model: `sonnet`, effort `high`. Depends on Task 3. **0.8 hooks:** consumes hook 5; neutral.

**Files:**
- Modify: `cmd/dubplate-server/serve.go` (the roster over the serve handle), `member.go` (`add` and
  `list` over `store.OpenWrite`), `doctor.go` (the members row), `web.go` (`privateFiles.roster`),
  and their tests
- Modify: `internal/engine/store/inspect.go` (member and admin counts and the import record's time in
  `Facts`), tests
- Modify: `internal/auth/roster.go` (delete the file-backed constructor, the file write, and the
  file lock once no caller remains; keep `parseRoster`, `validateRoster`, and `memberRecord`, which
  Task 5's import uses)
- Modify: `cmd/dubplate-server/rendered_test.go`, only its `newServeShell` call sites

**Outcomes:**
- `serve` builds the roster over its handle. `member add` and `member list` open write mode, under
  the store's shared lock and its owner rule (`leaf.create` and `checkOwner`, quoted in the code
  pointers; `checkRosterDirOwner` retires). `member add` on an absent store refuses with
  `OpenWrite`'s "start serve first", which is now satisfiable.
- The members row (R-6d13), its reads version-guarded in `Facts`.
- `web.go` drops the roster path (R-6d12).
- The file-backed roster code deleted.

**The members row's states:**

| Store | Row |
| --- | --- |
| absent | "no store yet; start serve, then member import or member add --admin", exit unaffected |
| older than `0003` (no `users` table) | "members not recorded before store version 3", exit unaffected; the store row reports the pending migrations |
| at `0003` or above but older than `0006` | the counts as below; "no import record yet" instead of a failed query |
| older or current, no member | fails naming `member import` and `member add --admin` |
| live members but no live admin (a tombstoned admin present) | fails naming `member add --admin` |
| members with at least one admin | passes with both counts, and the import record's time when present |
| refused or busy | the store row's own report; the members row reads "not read" |

**Acceptance criteria:**
- A `member add` then `member list` round trip through the verbs against a store Task 1's path
  created. **Fixture state:** two members added in a known order, so `List`'s `created_at`-then-`id`
  order and the count both show.
- `member add` against an absent store refuses and creates nothing (asserted by `stat`).
- `member add` against a store whose leaf another uid owns refuses naming the owner, the case
  `checkRosterDirOwner` covered (the `geteuid` override `leaf.go` already provides).
- The members row renders each state above in a golden. **Fixture state:** the no-admin fixture
  holds a tombstoned admin row, so counting tombstoned rows shows.
- `grep` finds no caller of the file-backed constructor and no `rosterFileKey` reader but Task 5's
  supersession.
- The gate green.

**Named mutations:** open inspect mode in `member add` (the write fails read-only); count tombstoned
rows as members (the no-admin golden fails); drop the `member_imports` version guard (the
`0003`-to-`0005` row refuses).

**Deliverable count: 4** (serve and the verbs, the members row, the web guard, the deletion).

---

## Task 5: `member import`

Model: `sonnet`, effort `high`. Depends on Task 4. **0.8 hooks:** lands hook 5's import channel
(ids minted for every imported member); otherwise neutral.

**Files:**
- Create: `cmd/dubplate-server/import.go` (or inside `member.go`, the implementer's call by size),
  tests, roster fixtures under `testdata/`
- Modify: `cmd/dubplate-server/auth.go` (`Required` loses the key; `AddSuperseded` gains it),
  `internal/auth` for the parse-and-validate of a roster file it already owns (`parseRoster`,
  `validateRoster`), exported only as far as the verb calls it

**Outcomes:**
- `member import --roster PATH [--dry-run]`, one `Update` under the shared lock, with the import
  record, the actor from `resolveActor`, and every read it decides on (the record, the existing rows)
  inside the callback.
- The thirteen states below, the file read as R-6d11 states (mode refusal, one read hashed and
  parsed, each entry through `validateWriteCandidate`); `--dry-run` reports every refusal at once and
  writes nothing.
- `DUBPLATE_ROSTER_FILE` superseded (R-6d12).
- The import's merge with members added before it (R-6d11).

**States** (entries are named by position and field, never by address or username, R-6d11):

| Roster file and store | Report |
| --- | --- |
| the file absent, and the file unreadable (mode 0000) | refused naming the file, each |
| the file writable by group or others | refused naming the file and its mode |
| the file malformed | refused naming the file and the parse error |
| the file lists no members | refused naming the file (6d-R4) |
| an entry failing a field rule (the write path's username pattern, a blank display name, a bad address, a refused rune, an overlong field) | refused naming its position and the rule |
| two entries with one address | refused naming both positions |
| two entries with one username, differing in case | refused naming both positions |
| entries with no `streamer_username` | refused naming each position (ruling 8) |
| no admin in the file or the table | refused naming the rule |
| an entry colliding with a different existing member on address or username | refused naming the position and the field |
| an entry whose address matches a tombstoned row | refused naming the position; the import is not a reactivation channel |
| an import record present | refused naming its time and actor |
| every entry valid, matching rows merged | imported; one record; each new member a fresh id; a differing display name on a matched row reported and left as stored |

**Acceptance criteria:**
- Spec criterion 1 (6d-R12): a valid roster of three (one admin) imports once; the rows carry
  `admin` 1 on the admin entry and 0 on the other two, so both an all-admin and a none-admin mutation
  fail; the record's hash equals `sha256.Sum256` of the fixture file's bytes, its count is 3, and its
  actor is `resolveActor`'s value; a second run refuses. The imported admin's session then adds a
  member through the roster API, and an imported non-admin's session meets the admin gate's 404.
- Spec criterion 2: each refusal by its own fixture, asserting the report text and that `users`, the
  record table, and `user_version` are unchanged. **Fixture state:** a store with one existing member
  (and, for the tombstone row, one tombstoned row), so "writes nothing" is checked against a non-empty
  table. The mode-0000 fixture assumes a non-root euid, as the gate runs; it carries no skip guard.
- `--dry-run` over a file carrying three refusals from different layers (one field rule, one
  collision with an existing member, and "no admin") reports all three and changes nothing.
- `--dry-run` over a valid three-entry file exits 0, reports the three additions, and leaves `users`,
  `member_imports`, and the settings generation unchanged.
- A fault fixture: the test adds a `BEFORE INSERT ON users` trigger to its temp store that raises on
  the third insert; the import fails, `member_imports` holds no row, and `users` holds none of the
  file's entries.
- Two imports of one valid file run concurrently: exactly one succeeds and the other refuses naming
  the record.
- Spec criterion 3: a member added through the verb before the import merges when the entry matches
  and refuses naming the position and the field when the entry's admin flag differs.
- Spec criterion 4: with `DUBPLATE_ROSTER_FILE` set, one table test over the registered command tree
  runs every verb that loads the wall and asserts each fails the load with the message verbatim.
- The gate green.

**Named mutations:** write the record outside the transaction (the fault fixture leaves a record);
let `--dry-run` enter `Update` without rolling back (the valid dry run writes); drop the mode check
(the writable-file row imports); validate entries with the loader's rule alone (the username-pattern
row imports); drop the case fold on the username duplicate check. A record read outside the callback
is a race the concurrent test catches only by timing, so the diff-reviewer checks the read's
placement by name, as in Task 3.

**Deliverable count: 4** (the verb, the state table, the supersession, the merge).

**Segment B ends here.** The conductor dispatches `web-auth-security-reviewer` over Tasks 3 to 5,
folds its findings in one fix task before segment C, and checkpoints STATUS.

---

## Task 6: The auth thresholds become settings

Model: `opus`, effort `high` (the config contract). Depends on Task 5. **0.8 hooks:** lands hook
15's `sensitive` marker on three keys under CH's meaning (R-6d6); hook 18 (every key) stays rung 14's.

**Files:**
- Modify: `internal/config/settings.go`, `registry.go`, tests (the closed-set string kind; the
  `Sensitive` doc)
- Modify: `internal/settings/evaluate.go`, `accessor.go`, tests (`parse`'s `KindString` arm; the
  string read; the withholding removed)
- Modify: `cmd/dubplate-server/auth.go` (ten settings, two correlated rules, `validateMagicLinkSource`
  trimmed, `resolveAuthConfig`'s env parse replaced), `serve.go`, `store.go` (the withholding
  removed), `doctor_test.go` goldens, and `rendered_test.go`'s `newServeShell` call sites

**Outcomes:**
- The closed-set string kind, `parse`'s arm, and the accessor's string read (R-6d7).
- The ten keys registered through `AddSetting` with the spec table's apply modes, their bounds as
  data (R-6d8), `Sensitive` on both lifetimes and the cap. The correlated rules register after the
  ten settings, since `AddCorrelatedRule` refuses an unregistered key.
- Both correlated rules registered; 6c's test pair retired.
- `Sensitive` re-documented and the value shown (R-6d6).
- **The interim read.** A key cannot be both a threshold and a setting (`claim` panics,
  `registry.go:126-131`), and `resolveAuthConfig` parses thresholds from the env. So this task
  replaces that parse: the preparation function reads all ten values once, after the seed, through
  the accessor over its handle, into `authConfig`. The task is green at that seam; Task 7 then turns
  the "immediately" and "next operation" values into per-use readers and deletes `authConfig`.

**States:**

| Check | State | Report |
| --- | --- | --- |
| Registration | a closed-set string with an empty list, or a default outside it | panics naming the key |
| Validator | `generic` / `open` for the answer | accepted / refused naming the set |
| Correlated rule | retention and expiry raised together in one seed set | passes |
| Correlated rule | only the expiry raised past the stored retention | refused naming both keys |
| Correlated rule | lifetime above the cap | refused naming both keys |
| Seed | a stored `env-seed` string row, env equal | unchanged, no refusal |
| `doctor` | each re-homed key | value, source, apply mode, and the marker on three keys |
| `backup` start | an auth env value out of bounds | logs the refusal, does not fail |
| `check` | an auth env value out of bounds | exits nonzero naming the key (6c's Ruling 2, now the deploy gate for auth values) |

**Acceptance criteria:**
- Spec criterion 5, re-run over the real pair (6c-E6): the two correlated rows above.
- Spec criterion 10: each re-homed key seeds from the env (the seed table's second row, one subtest
  per key, table-driven), and the `doctor` golden shows the source and the `sensitive` marker on both
  lifetimes and the cap, with their values printed.
- `backup` with an out-of-bounds auth value leaves every settings row unchanged and exits on its own
  outcome (the spec's "a timer-run backup never fails over an auth threshold").
- The gate green.

**Named mutations:** keep `KindString` refused (the answer cannot register); keep the withholding
(the golden fails); register the retention rule with only one key (the both-keys report fails);
drop `parse`'s string arm (the equal-env string row refuses, so a second start refuses); let `check`
pass an out-of-bounds auth value (the `check` row exits 0).

**Deliverable count: 4** (the string kind, the ten settings, the correlated rules, the sensitive
meaning).

---

## Task 7: Accessor reads at every auth call site

Model: `opus`, effort `high` (live security thresholds). Depends on Task 6. **0.8 hooks:** neutral
(the per-account limiter of S20 slice 6 is its own).

**Files:**
- Modify: `internal/auth/handlers.go`, `seam.go`, `admin.go`, `nonce.go` (`Insert`, `Sweep`, and
  `RunSweep`'s sweep function gain `ctx`), and tests (reader functions in place of values)
- Modify: `cmd/dubplate-server/serve.go` (including `startNonceSweep`), `auth.go` (`authConfig` and
  `resolveAuthConfig` deleted; readers built over the serve accessor), tests, and `rendered_test.go`'s
  `newServeShell` call sites

**Outcomes:**
- Each "next operation" and "immediately" value read through an injected reader at its point of use
  (R-6d9); `internal/auth` does not import `settings`.
- The nonce store reads the retention at each insert and sweep; a failed read in the sweep logs the
  key and skips that sweep. `OpenNonceStore`'s `linkExpiryCeiling` stays the registered ceiling
  (R-6d9).
- The four rate limits read once after the seed; the limiter built from them.
- The read orderings R-6d9 fixes (the answer before `Lookup`; the lifetime and cap before
  `Nonces.Insert`).
- `authConfig` gone.

**Acceptance criteria:**
- Spec criterion 9 (6d-R9): with env A seeded and then an `admin` row B written in-package through
  the store with the generation bumped (the rung 14 writer's stand-in):
  - a link minted after the write carries B's expiry;
  - a session's next refresh uses B's lifetime;
  - the next request enforces B's cap when it is lowered: a session whose age passes B but not A is
    refused;
  - a raised cap admits only through a refresh, because `clampExpiry` bound the cookie's expiry to A
    at mint (quoted): a session refreshed after the write carries an expiry past `AnchoredAt+A`, and
    a verification between `AnchoredAt+A` and that expiry admits it;
  - the limiter uses A until a new start, then B. At that start the env is unset (or set to B), and a
    separate sub-case with env A still beside admin row B asserts the start refuses naming both
    values.

  **Fixture state:** a current store, a member, a session minted under A; A and B limits chosen so a
  named request count trips A and not B.
- Spec criterion 12's request half: a request handled after start reads the seeded value (the
  unknown-address answer seeded `generic` answers generically).
- A failing reader per apply mode fails its request through the handler's server-error path and logs
  the key, with no default: the answer reader in the link request (a member and a stranger get the
  same 500), the cap reader in verification, and the lifetime reader in `consume`, after which the
  same link still succeeds once the reader recovers (the nonce was not consumed). The retention
  reader's failure in the sweep logs the key and deletes nothing.
- The gate green.

**Named mutations:** cache the cap at construction (the lowered-cap case passes a session it must
refuse); re-read the rate limits per request (the limiter case uses B early); default on a read error;
read the lifetime after `Nonces.Insert` (the recovered link is refused as replayed); read the answer
after `Lookup` (the member and the stranger differ); pass the value in force as `linkExpiryCeiling`
(a nonce test with an unreadable record and a lowered expiry sweeps it early).

**Deliverable count: 4** (handler and seam readers, the nonce readers, the start-time limiter,
`authConfig`'s removal).

---

## Task 8: The deployment

Model: `sonnet`, effort `high`; invokes `vps-conventions`. Depends on Task 7 (independent by files of
Tasks 6 and 7, but sequenced to keep one gate lane). **0.8 hooks:** neutral.

**Files:**
- Modify: 5d's serve unit, `scripts/deploy.sh`, `scripts/check.sh` (the state-dir leg),
  `env/dubplate-server.env.template`, `docs/server-runbook.md`
- Modify: `internal/engine/store/inspect.go` and `cmd/dubplate-server/check.go` (holders named for an
  older store), tests
- Modify: `internal/engine/store/leaf.go` (`checkOwner`'s hint, R-6d15) and the four store tests that
  assert it (`leaf_test.go`, `inspect_test.go`, `restore_test.go`, `open_test.go`)

**Outcomes:**
- The unit's `ReadWritePaths` names the `store` leaf unprefixed and no longer the roster leaf; the
  state-dir leg covers it.
- `deploy.sh` pre-creates the `store` leaf (owner, mode 0700), no longer the roster leaf; runs
  store-opening verbs as the owner by R-6d15; and, per R-6d19, renders the env, installs the binary,
  runs the owner-run `check`, halts on its nonzero exit, and only then runs `systemctl try-restart`
  on the serve unit (landing 6c's Ruling 2 here only if Task 0 found 5d did not).
- `check`'s store row names the holders' pids for an older store whose lock another process holds,
  excluding its own pid (`Inspect` holds its own shared lock while it reads, and `/proc/locks` lists
  the caller, quoted), so `deploy.sh`'s printed `check` output names the holder before the restart.
  `checkOwner`'s hint names the rendered environment instead of `runuser`.
- The template drops `DUBPLATE_ROSTER_FILE`; the runbook's install order becomes: render the env,
  start `serve` (it creates the store), `member add --admin` or `member import`, `doctor`; and a 6d
  upgrade section: remove the roster row, deploy, `member import --dry-run --roster <file>`,
  `member import`, verify with `member list`, and keep the file until 5d's `.prev` copies are
  superseded (R-6d18).

**States of `check`'s store row:**

| Store | Row |
| --- | --- |
| older, no other holder | older, pending files named, no holder line (its own pid never listed) |
| older, a backup holding the lock shared | older, pending files named, "lock held by pid N", N the other process |
| older, `/proc/locks` unreadable | older, "holders unknown" |
| current, held shared | current, no holder line (a current start never waits on a shared holder) |

**Acceptance criteria:**
- Spec criterion 13: `systemd-analyze verify` (the gate's leg) clean on the unit. The state-dir leg
  is extended to fail when the unit's store path and the template's state dir disagree, when the
  unit's `ReadWritePaths` entry for the store leaf carries a `-` prefix, and when `deploy.sh`'s leaves
  array lacks `store` or still names `roster`. Each is a gate check on the files, the leg's existing
  precedent; no gate leg runs `deploy.sh`, which needs the box.
- Spec criterion 14 (6d-R10): the template carries no `DUBPLATE_ROSTER_FILE` row (the gate's
  template-keys test agrees); `shellcheck` clean. The step order R-6d19 fixes (render, install, owner
  `check` through `systemd-run --wait`, halt, `try-restart`) is not machine-checked, since shell
  carries no behavior tests (CLAUDE.md) and `deploy.sh --check` needs ssh to the box; the
  diff-reviewer reads that order by name against R-6d19, and the task's report quotes the lines.
- `check`'s rows above, each in a golden, with the shared holder a second process.
- The owner-refusal hint: the four store tests assert the new text.
- The runbook sections read by the task's own render-and-read (the page opened and read in full).
- The gate green.

**Named mutations:** grant the leaf with a `-` prefix (the extended leg fails); leave `roster` in the
leaves array (the leg fails); break the unit's store path (the leg fails); name holders only when busy
(the shared-holder golden fails); list `check`'s own pid (the no-other-holder golden fails).

**Deliverable count: 4** (unit and leg, `deploy.sh`, the store's operator strings in `check` and
`checkOwner`, template and runbook).

**Segment C ends here.** The conductor dispatches `web-auth-security-reviewer` over Tasks 6 and 7 and
folds; then the simplifier round; then the merge ritual.

---

## Task 9: The pass close

Model: `sonnet`. One fold agent drafts, commits the drafts, then folds, with one independent
`diff-reviewer` read of the fold's diff.

**Outcomes:**
- The 6d HISTORY entry: what landed per task, the gate evidence, both security reads' dispositions,
  both budgets, "What a later pass would be wrong to rediscover" (the start order, the sensitive
  meaning, why no last-admin guard, the closed-set kind, the empty-import refusal, the `systemd-run`
  form, `try-restart` and the deploy step order, the keep extraction, the `parse` string arm).
- The errata table above applied, dated, in the spec and the 6c HISTORY entry.
- Routed items: to the revoke verb (criteria 7 and 8, minor 14, under Ruling 2's yes), to the first
  demote or remove verb (the last-admin guard), to rung 14 (the network actor on `UserID`; the admin
  write path's resolving channel for a refusing `admin` row), to rung 16 (the store identity under
  Ruling 3's yes; the runbook's remaining owed lines), to lane 1's next `web/` pass (R-6d17's comment).
- STATUS rewritten present tense, the "To 6d" line removed, the no-deploy carry naming 6d as landed.

**Deliverable count: 3** (HISTORY, errata, STATUS).

---

## The exit gate

`check.sh` green on master after the merge, plus spec criteria 1 to 6 (6 without its last-admin
clause), and 9 to 14, each by the task that names it; criteria 7 and 8 routed under Ruling 2. Each
new export lists its production caller; the diff-reviewer's records carry every named mutation
reported failing.

## Contested surfaces

Every file below is shared with rung 5's closed work or another lane; no other executor edits one
while 6d runs: `cmd/dubplate-server/serve.go`, `auth.go`, `store.go`, `doctor.go`, `check.go`,
`member.go`, `web.go`, `main.go`'s registration block, and `rendered_test.go`'s `newServeShell`
call sites; `internal/auth/*`; `internal/identity`; the two `internal/ops` test fixtures
(`ops_test.go`, `opshttp/helpers_test.go`); `internal/config/settings.go` and `registry.go`;
`internal/settings`; `internal/engine/store` (`open.go`, `inspect.go`, `leaf.go` and the tests
asserting its hint, `migrations/`); `internal/backup` (the reconciler); 5d's serve unit;
`scripts/deploy.sh` (the translit plan also edits it); `scripts/check.sh`; the env template;
`docs/server-runbook.md`. The photo lane's `check.sh` integrity line is the one expected concurrent
edit; it lands before Task 8 or after the merge, never during it.

## The conductor's actions

1. At approval: record the rulings; rewrite STATUS's "5d's deploy gate" line per R-6d1; route Ruling
   2's and 3's items in STATUS.
2. Before Task 0: the live-executor check (lane registry, `git status`), the registry row, the runaway
   guard, AC power.
3. After Task 0: re-derive the ceiling from 5d's actual; stop on any blocker.
4. At each segment boundary: STATUS; check the 80% flag (9.2M).
5. After segments B and C: `web-auth-security-reviewer`, briefed as "Review focus" says; one fold task
   each.
6. After segment C: the simplifier round with `.claude/instructions/simplifier-brief.md` pasted
   verbatim, and its `diff-reviewer`; the merge ritual; seven `go-architecture-reader` dispatches, one
   per touched package.
7. Task 9, then close the session.

## Execution model

`pass-execute`, sequential, one invocation per segment (A: Tasks 0 to 2; B: 3 to 5; C: 6 to 8), with
`implementer: "dubplate-implementer"`, `reviewer: "diff-reviewer"`, `maxFix: 1`, `stopOnEscalate:
true`, the heavy-lane gate string above, and each task's `model` as named. The security reads, the
fold tasks, the simplifier round, and Task 9 run between invocations. Each boundary lands on a green
commit. No task runs in parallel: every task after Task 2 reads the one before.

## Prior art consulted

| Entry | Verdict for 6d |
| --- | --- |
| `modernc.org/sqlite` | Adopted (spec); `0006` rides it. `NOCASE` behavior quoted and relied on for ASCII usernames only. |
| `pressly/goose`, `golang-migrate` | Declined (spec); `0006` joins the hand-rolled runner. |
| Navidrome (behavior only, GPL) | Consulted: its external-auth auto-creation of an unknown user is why `Principal` stays the streamer username and never the id (spec `:855-860`). No code. |
| OpenSubsonic specification | No 6d surface; `UserID` is what 0.8's per-user rows key on (S20 hook 5). |
| `golang.org/x/time/rate` | Already adopted by rung 5; 6d only moves when its limits are read (restart mode). |
| `crypto/sha256`, `crypto/rand` (via `store.NewID`) | Adopted (stdlib) for the import record and the id. |
| `internal/atomicfile` | Not used: the import writes through the store, not a file. |
| systemd `ReadWritePaths=`, `systemd-run`, `runuser` | Quoted; `systemd-run --uid` adopted for owner-run verbs (R-6d15). |
| gonic, taglib, mewkiz/flac, Picard tables, chromaprint | No 6d surface. |

## Not verified in this revision

- Every line citation into 5d's files, which 5d-i and 5d-ii rewrite; Task 0 re-resolves them.
- 5d's serve unit's `Restart=` value is `on-failure` with `RestartSec=5` in the approved 5d plan
  (R-5d10), on which R-6d3's retry relies; Task 0 confirms it on the landed unit.
- Whether 5d's `deploy.sh` already halts on `check`'s nonzero exit, and how it restarts `serve`
  (Task 0 item 2, R-6d19).
- The trigger-based fault fixture (Task 5) was probed through Python's `sqlite3`, not through
  `modernc.org/sqlite`; both are SQLite, and Task 5's first run of the fixture is the confirmation.
- R-6d3's 5m default is reasoned from lock semantics, not measured on a populated store.
