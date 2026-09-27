# Rung 6, Pass 6d (Identity, the Handoff, and `serve`'s Store) Implementation Plan

Status: **revision 1, 2026-09-26; drafted, not reviewed, no ruling answered.** This draft goes through
`spec-plan-review` (contract, mechanics, and risk lenses) before Geoff's approval gate. It is drafted
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
  `Valid`. `internal/ops/ops.go`: `NetworkActor` ("member:<principal>").
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
- **No `web/` edit** (R-6d17), and `cmd/dubplate-server/rendered_test.go` is read, never edited.
- **Commits**: each task commits its own files, Conventional Commits, the repo's attribution footer.
- **Implementers never dispatch agents**; the conductor owns `code-simplifier` and every reviewer.

## Privileged fields and their channels

| Field | Every channel that may set it | Never |
| --- | --- | --- |
| `users.id` (the surrogate id) | the store's `NewID`, in the write path (`AddMember`) and in `member import` | a request body, the roster file, a CLI flag |
| `users.admin` | `member add --admin` (the CLI grant, 5b-ii's rule); `member import`, from the file's `admin` field, once | the roster API and the add-member page (cookie channel), which refuse it as today |
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
shared lock. The spec's "then recovery" (spec `:872-874`) reads as `OpenServe`'s own staging cleanup
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
`settings.EvaluateFacts` and passes the local keep's value in force (the stored row, or the registry
default when the store holds none or predates the settings table) as `ServeOptions.Keep`, and
`cfg.IntegrationEnabled` of the backup class as `ServeOptions.Backup`. A stored keep row outside its
bounds refuses the start naming the key and the bound, before any migration. An env value that
differs reaches the keep after this start's seed, which is the spec's stated order (spec `:411-414`).

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
must be empty for it. The accessor gains a typed string read, whose caller is the handlers' answer
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

### R-6d10: the users-backed roster

`internal/auth`'s roster keeps its type and methods and gains a constructor over a `*store.Handle`
(the spec's one-constructor substitution, R5S `:194-196`). Its methods take `ctx` first, which the
spec's "without touching a call site" did not foresee: a database read needs a context
(`go-conventions`), and every caller already holds a request or command context. This is reading
6d-R2 in the deviations table.

- `Lookup` reads the live row (`tombstoned_at IS NULL`) by normalized address; no row is
  `ErrMemberNotFound`; a store error is returned as itself, never as not-found.
- `AddMember` runs in one `Update`: it validates as `validateWriteCandidate` does today, then reads
  inside the callback: a live row on the address refuses as today's duplicate; a tombstoned row on the
  address is **reactivated under the same id** with the new fields (spec ruling 9); another row whose
  `streamer_username` matches case-insensitively refuses naming the username. A new row gets
  `store.NewID`. The admin flag rule carries over unchanged (5b-ii: the cookie channel refuses it).
  The audit record stays the log record (`logAddition`).
- **No last-admin guard is built.** The spec's "5c builds the last-admin guard, and 6d re-establishes
  it in the write transaction" (spec `:182-183`) presumes a demotion or removal path; none exists
  (`grep -rn -i 'demot\|RemoveMember' internal/auth cmd` finds only comments), so the guard and
  criterion 6's "concurrent self-demotion of the last two admins" route to the first verb that can
  remove or demote. Reading 6d-R3.
- `Member` gains the id; `Member.principal`'s address fallback stays for no row, since every `users`
  row carries a username.

### R-6d11: the import record and the import's states

Migration `0006` adds `member_imports`: a `BLOB` `id`, `at`, the file's SHA-256 as a `BLOB`, the
member count, and the OS actor, `STRICT`, no `kind` column (so invariant 3 does not apply), no path
column (the file's path is not stored), and no change to invariant 6's exempt list. "An import record
already present" means any row. The match rule is the spec's: an entry whose address, streamer
username, and admin flag equal an existing live row is that member and changes nothing; the display
name is not part of the match, and a differing one is reported by `--dry-run` and left as stored.
**A file listing no members is refused naming the file**, a tenth state the spec's table lacks: an
empty import would write the record and block every later import while adding no one. Reading 6d-R4.

### R-6d12: `DUBPLATE_ROSTER_FILE` retires through `AddSuperseded`

In the same commit as the verb (spec `:815-817`), the key leaves `magicLinkSource.Required` and
registers through `AddSuperseded` with the message "superseded: the roster lives in the store; move a
roster file in once with dubplate-server member import --roster PATH". `member import` takes the path
as a flag, never from the env. `web.go`'s `privateFiles.roster` entry goes with the key, since the
store leaf sits under the state directory, which the same check already refuses.

### R-6d13: `doctor`'s roster row becomes a members row

Read through `store.Inspect`'s facts (a member count, an admin count, and the import record's time
when present), never by opening the store for writing. Its states are Task 4's table.

### R-6d14: the network actor stays `member:<principal>`

`ops.NetworkActor` has no production caller until rung 14 mounts the projection, so 6d changes
nothing in `internal/ops`. The routed note for rung 14: once an admin can edit a streamer username,
a job's actor should key on `UserID`, because a renamed principal breaks the actor's correlation.

### R-6d15: store-opening verbs from `deploy.sh`

The spec says `deploy.sh` runs every store-opening verb "through `runuser -u <owner>`" (spec `:251-252`,
criterion 14). `runuser` sets only `HOME` and `SHELL` (quoted) and loads no environment file, while
the verbs need the rendered env. So `deploy.sh` uses the operator form 5d's runbook already uses:
`systemd-run --wait --pipe --uid=<owner> --gid=<owner> -p EnvironmentFile=<env>` with the unit's
filesystem properties, which makes the same uid change and loads the env. Reading 6d-R5.

### R-6d16: `OpenServe`'s loop and the reconciler

The reconciler: `internal/backup` exports one function whose production caller is `serve`, returning
the same report body `restore` writes (`restorer.report`), over the root the restore-intent record
names when it names one and the configured library root otherwise, matching `restore.go:633`. Since
`ServeOptions.Reconcile` receives only the handle, the store passes the record's library root to it;
the signature may change, as it has no production caller yet. The settle loop (6a's open item): each
round needs another process to win a lock conversion by completing a restore swap, which is an
operator action, and each wait is already bounded; 6d adds a warning line per re-entry naming the
reason, and no iteration bound, since a bound would refuse a legitimate operator sequence. The nil
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
for rung 16.

---

## Spec deviations and readings that take effect on approval

| # | Spec text | 6d reading | Reason |
| --- | --- | --- | --- |
| 6d-R1 | STATUS "5d's deploy gate checks that 6d's two wiring items have landed" | 6d lands before any deploy that enables or routes to the serve unit | R-6d1 |
| 6d-R2 | "The store moves behind the roster's interface without touching a call site" (`:853-855`) | The constructor is the one substitution; the methods gain `ctx` | R-6d10 |
| 6d-R3 | Criterion 6's last-admin clause; `:182-183` | Routed to the first demote or remove verb | R-6d10 |
| 6d-R4 | The import's "nine states" (`:820-831`) | Ten: an empty file refuses | R-6d11 |
| 6d-R5 | Criterion 14's "through `runuser`" | `systemd-run --uid` with the env file | R-6d15 |
| 6d-R6 | "then recovery" (`:872-874`) | `OpenServe`'s staging cleanup and restore finish; job recovery with rung 14 | R-6d2 |
| 6d-E1 | 6c's `Setting.Sensitive` ("must not print") | CH's marker; the value is shown | R-6d6 |
| 6d-R7 | Criteria 7 and 8 (Ruling 2 yes) | Routed to the revoke verb | Ruling 2 |
| 6d-R8 | STATUS "To 6d: store identity" (Ruling 3 yes) | Routed to rung 16 | Ruling 3 |

---

## Review focus

The `diff-reviewer` reads each task against its criteria. The conductor dispatches
`web-auth-security-reviewer` (Opus) after segment B (Tasks 3 to 5) and after segment C (Tasks 6 and
7), each briefed with this plan's privileged-fields table and asked for correctness gaps ranked by
consequence, not polish; a fold task may refuse a finding whose fix costs more than the risk it
removes, with a one-line reason. Checklist lines for both:

- `Identity.UserID` has one producer (the resolution point); `Principal` is never the id.
- The admin flag reaches `users` only through the CLI grant and the import.
- A tombstoned address re-adds under its old id; a live one refuses; a username collision refuses by
  name, case-insensitively, before the insert.
- The seed apply runs after `OpenServe` and before `serve.Run`, and a refusal leaves no listener.
- The absolute cap is read at every verification; the four rate limits are read once, after the seed.
- No auth reader falls back to a default on a read error.
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
   2, owed by 5d), the state-dir leg in `scripts/check.sh`, the template's `DUBPLATE_ROSTER_FILE`
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

**Outcomes:**
- `DUBPLATE_STORE_MIGRATION_WAIT`, R-6d3, with its `doctor` threshold row.
- `newServeCmd` opens the store through `OpenServe` per R-6d2 and R-6d4, closes the handle when
  `serve.Run` returns, and opens no listener on any refusal.
- The exported reconciler (R-6d16), whose one production caller is `serve`.
- The settle loop's re-entry warning (R-6d16).

**States each check meets, and its report:**

| Store state at start | Report |
| --- | --- |
| absent (no leaf) | created at the binary's version; the listener follows |
| an interrupted create (zero-byte database, no `-wal`) | completed empty |
| older: at `0004` with an `env-seed` keep row of 1 and two pre-migration copies, `0005` onward pending | migrated; `pre-migrate-v4.db` written; with backup off, one copy survives (the keep row's 1) |
| older, same, with the keep row absent | migrated; copies pruned to the registry default |
| older, with a stored keep row outside its bounds | refused naming the key and bound; the store stays at `0004`, no copy written |
| newer than the binary | refused naming both versions |
| a `reconcile-pending` intent over an older store | migrated, the report written into the leaf over the record's root, the intent removed |
| any other intent | refused naming `restore --resume` |
| the lock held exclusive by another process past the wait | exits nonzero naming the holder's pid |
| `snapshots/` holds a copy and the database is gone | refused naming the newest copy (E11, unchanged) |

**Acceptance criteria:**
- Spec criterion 11: a test drives the serve verb's start path against an absent store and against
  the older-store fixture above, and asserts the store is at the binary's version when the function
  that precedes `serve.Run` returns. **Fixture state:** the older store carries a populated row in
  `jobs` so a migration that silently recreated the store would lose it, and the assertion checks it.
- The lock-wait row, with the lock held by a second process (the re-executed test binary the store's
  criterion 14 test uses), so the named pid differs from the test's own; `MigrationWait` set to 1s.
- D10's prune proof: the copy count after migration equals the keep row's value, and a second run
  with the row removed equals the default. **Fixture state:** more copies than either value.
- The reconcile row asserts the report file's body names a planted row without a file, and that the
  intent is gone.
- The gate green.

**Named mutations:** pass a fixed `Keep` of 3 (the keep-row case fails); open the listener before
`OpenServe` (the order test fails); pass the configured root when the record names one (the
reconcile row fails); drop the handle's close (a test that runs `restore` after the verb returns
fails on `ErrLockBusy`).

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
| an in-package `admin` row, env different | refused naming both values; no row changed; no listener |
| an env value that does not parse | refused naming the key and the accepted form |

**Acceptance criteria:**
- Spec criterion 12's start half: each row by its own subtest, through the verb's start path; the
  request half lands in Task 7. **Fixture state:** a current store, created by Task 1's path, with the
  row each case names; the unchanged row asserts the generation value, never only a nil error.
- The `doctor` and `backup` strings: a test renders `doctor`'s settings row for an env value that
  differs from the stored row, runs the serve start path, and asserts the row then reads the value in
  force. This proves the promise the string makes.
- The gate green.

**Named mutations:** call the apply after `serve.Run` starts (the refusal row opens a listener);
swallow the apply's error (the refusal row starts).

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
| `AddMember` | a tombstoned row on the address | reactivated: same id, new fields, `tombstoned_at` cleared |
| `AddMember` | another row's username, differing only in ASCII case | refused naming the username |
| `AddMember` | no username | refused (`errMissingStreamerUsername`) |
| `AddMember` | admin requested on the cookie channel | refused (5b-ii) |
| `AddMember` | admin on the command-line channel | admin row written |
| `List` | an empty table | an empty list, no error |

**Acceptance criteria:**
- Spec criterion 5: a verified identity's `UserID` equals the row's id and its `Principal` the row's
  username. **Fixture state:** a member whose address differs from its username, so a `Principal`
  built from the address fails.
- Spec criterion 6, minus the last-admin clause (6d-R3): each `AddMember` row by its own subtest.
  **Fixture state:** the tombstone row is written directly by the test (no removal path exists), and
  the reactivation asserts the id byte for byte.
- The invariant test passes with `0006`'s fixture; the table refuses a `NULL` actor.
- The gate green.

**Named mutations:** mint a new id on reactivation; compare usernames case-sensitively; drop the
`tombstoned_at IS NULL` filter in `Lookup`; set `Principal` from the address.

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
- Modify: `internal/auth/roster.go` (delete the file-backed constructor and its helpers once no
  caller remains)

**Outcomes:**
- `serve` builds the roster over its handle. `member add` and `member list` open write mode, under
  the store's shared lock and its uid rule (`checkRosterDirOwner` retires). `member add` on an absent
  store refuses with `OpenWrite`'s "start serve first", which is now satisfiable.
- The members row (R-6d13).
- `web.go` drops the roster path (R-6d12).
- The file-backed roster code deleted.

**The members row's states:**

| Store | Row |
| --- | --- |
| absent | "no store yet; start serve, then member import or member add --admin", exit unaffected |
| older or current, no member | fails naming `member import` and `member add --admin` |
| members but no admin | fails naming `member add --admin` |
| members with at least one admin | passes with both counts, and the import record's time when present |
| refused or busy | the store row's own report; the members row reads "not read" |

**Acceptance criteria:**
- A `member add` then `member list` round trip through the verbs against a store Task 1's path
  created. **Fixture state:** two members, so an order or count bug shows.
- `member add` against an absent store refuses and creates nothing (asserted by `stat`).
- The members row renders each state above in a golden.
- `grep` finds no caller of the file-backed constructor and no `rosterFileKey` reader but Task 5's
  supersession.
- The gate green.

**Named mutations:** open inspect mode in `member add` (the write fails read-only); count tombstoned
rows as members (the golden fails).

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
  record, the actor from `resolveActor`.
- The ten states below; `--dry-run` reports every refusal at once and writes nothing.
- `DUBPLATE_ROSTER_FILE` superseded (R-6d12).
- The import's merge with members added before it (R-6d11).

**States:**

| Roster file and store | Report |
| --- | --- |
| the file absent, and the file unreadable (mode 0000) | refused naming the file, each |
| the file malformed | refused naming the file and the parse error |
| the file lists no members | refused naming the file (6d-R4) |
| two entries with one address | refused naming both entries |
| two entries with one username, differing in case | refused naming both entries |
| entries with no `streamer_username` | refused naming each (ruling 8) |
| no admin in the file or the table | refused naming the rule |
| an entry colliding with a different existing member on address or username | refused naming both |
| an import record present | refused naming its time and actor |
| every entry valid, matching rows merged | imported; one record; each new member a fresh id |

**Acceptance criteria:**
- Spec criterion 1: a valid roster of three (one admin) imports once, carries the admin flag, writes
  the record, and a second run refuses; the imported admin then runs `member add --admin` for a new
  member successfully.
- Spec criterion 2: each refusal by its own fixture, asserting the report text and that `users`, the
  record table, and `user_version` are unchanged. `--dry-run` over a file carrying three different
  refusals reports all three and changes nothing. **Fixture state:** a store with one existing member,
  so "writes nothing" is checked against a non-empty table.
- Spec criterion 3: a member added through the verb before the import merges when the entry matches
  and refuses naming both when the entry's admin flag differs.
- Spec criterion 4: with `DUBPLATE_ROSTER_FILE` set, `doctor`, `check`, `serve`, and `member list`
  each fail the load with the message verbatim.
- The gate green.

**Named mutations:** write the record outside the transaction; let `--dry-run` enter `Update`
without rolling back; drop the case fold on the username duplicate check.

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
- Modify: `internal/settings/evaluate.go`, `accessor.go`, tests (the string read; the withholding
  removed)
- Modify: `cmd/dubplate-server/auth.go` (ten settings, two correlated rules, `validateMagicLinkSource`
  trimmed), `store.go` (the withholding removed), `doctor_test.go` goldens

**Outcomes:**
- The closed-set string kind and the accessor's string read (R-6d7).
- The ten keys registered through `AddSetting` with the spec table's apply modes, their bounds as
  data (R-6d8), `Sensitive` on both lifetimes and the cap.
- Both correlated rules registered; 6c's test pair retired.
- `Sensitive` re-documented and the value shown (R-6d6).

**States:**

| Check | State | Report |
| --- | --- | --- |
| Registration | a closed-set string with an empty list, or a default outside it | panics naming the key |
| Validator | `generic` / `open` for the answer | accepted / refused naming the set |
| Correlated rule | retention and expiry raised together in one seed set | passes |
| Correlated rule | only the expiry raised past the stored retention | refused naming both keys |
| Correlated rule | lifetime above the cap | refused naming both keys |
| `doctor` | each re-homed key | value, source, apply mode, and the marker on three keys |
| `backup` start | an auth env value out of bounds | logs the refusal, does not fail |

**Acceptance criteria:**
- Spec criterion 5, re-run over the real pair (6c-E6): the two correlated rows above.
- Spec criterion 10: each re-homed key seeds from the env (the seed table's second row, one subtest
  per key, table-driven), and the `doctor` golden shows the source and the `sensitive` marker on both
  lifetimes and the cap, with their values printed.
- `backup` with an out-of-bounds auth value leaves every settings row unchanged and exits on its own
  outcome (the spec's "a timer-run backup never fails over an auth threshold").
- The gate green.

**Named mutations:** keep `KindString` refused (the answer cannot register); keep the withholding
(the golden fails); register the retention rule with only one key (the both-keys report fails).

**Deliverable count: 4** (the string kind, the ten settings, the correlated rules, the sensitive
meaning).

---

## Task 7: Accessor reads at every auth call site

Model: `opus`, effort `high` (live security thresholds). Depends on Task 6. **0.8 hooks:** neutral
(the per-account limiter of S20 slice 6 is its own).

**Files:**
- Modify: `internal/auth/handlers.go`, `seam.go`, `admin.go`, `nonce.go`, and tests (reader
  functions in place of values)
- Modify: `cmd/dubplate-server/serve.go`, `auth.go` (`authConfig` and `resolveAuthConfig` deleted;
  readers built over the serve accessor), tests

**Outcomes:**
- Each "next operation" and "immediately" value read through an injected reader at its point of use
  (R-6d9); `internal/auth` does not import `settings`.
- The nonce store reads the retention at each insert and sweep.
- The four rate limits read once after the seed; the limiter built from them.
- `authConfig` gone.

**Acceptance criteria:**
- Spec criterion 9: with env A seeded and then an `admin` row B written in-package through the store
  with the generation bumped (the rung 14 writer's stand-in), a link minted after the write carries B's
  expiry; a session's next refresh uses B's lifetime; the next request enforces B's absolute cap, in
  both directions (a cap lowered below a live session's age refuses it, raised above admits it); and
  the limiter uses A until a new start, then B. **Fixture state:** a current store, a member, a session
  minted under A.
- Spec criterion 12's request half: a request handled after start reads the seeded value (the
  unknown-address answer seeded `generic` answers generically).
- A reader returning an error fails the request through the handler's server-error path and logs the
  key; no default.
- The gate green.

**Named mutations:** cache the cap at construction (the lowered-cap case passes a session it must
refuse); re-read the rate limits per request (the limiter case uses B early); default on a read error.

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

**Outcomes:**
- The unit's `ReadWritePaths` names the `store` leaf unprefixed; the state-dir leg covers it.
- `deploy.sh` pre-creates the `store` leaf (owner, mode 0700), no longer the roster leaf; runs
  store-opening verbs as the owner by R-6d15; runs `check` before restarting `serve` and halts on its
  nonzero exit (landing 6c's Ruling 2 here only if Task 0 found 5d did not).
- `check`'s store row names the holders' pids for an older store whose lock another process holds, so
  `deploy.sh`'s printed `check` output names the holder before the restart.
- The template drops `DUBPLATE_ROSTER_FILE`; the runbook's install order becomes: render the env,
  start `serve` (it creates the store), `member add --admin` or `member import`, `doctor`; and a 6d
  upgrade section: remove the roster row, deploy, `member import --dry-run --roster <file>`,
  `member import`, delete the file after verifying `member list` (R-6d18).

**States of `check`'s store row:**

| Store | Row |
| --- | --- |
| older, no other holder | older, pending files named, no holder line |
| older, a backup holding the lock shared | older, pending files named, "lock held by pid N" |
| older, `/proc/locks` unreadable | older, "holders unknown" |
| current, held shared | current, no holder line (a current start never waits on a shared holder) |

**Acceptance criteria:**
- Spec criterion 13: `systemd-analyze verify` (the gate's leg) clean on the unit; the state-dir leg
  fails when the unit's store path and the template's state dir disagree (a mutation proves it).
- Spec criterion 14: the template carries no `DUBPLATE_ROSTER_FILE` row (the gate's template-keys test
  agrees); `deploy.sh --check` (dry) names the store leaf's pre-create and no roster leaf, and shows
  the owner-run `check` before the restart step; `shellcheck` clean.
- `check`'s rows above, each in a golden, with the shared holder a second process.
- The runbook sections read by the task's own render-and-read (the page opened and read in full).
- The gate green.

**Named mutations:** grant the leaf with a `-` prefix (the unit reads clean but the leg must fail on
the pre-create's absence, so the plan's leg asserts the pre-create); name holders only when busy (the
shared-holder golden fails).

**Deliverable count: 4** (unit and leg, `deploy.sh`, holders in `check`, template and runbook).

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
  form).
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
`member.go`, `web.go`, `main.go`'s registration block; `internal/auth/*`; `internal/identity`;
`internal/config/settings.go` and `registry.go`; `internal/settings`; `internal/engine/store`
(`open.go`, `inspect.go`, `migrations/`); `internal/backup` (the reconciler); 5d's serve unit;
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
- 5d's serve unit's `Restart=` value, on which R-6d3's retry relies.
- Whether 5d's `deploy.sh` already halts on `check`'s nonzero exit (Task 0 item 2).
- R-6d3's 5m default is reasoned from lock semantics, not measured on a populated store.
