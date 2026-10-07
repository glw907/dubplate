# Rung 6, Pass 6d (Identity, the Handoff, and `serve`'s Store) Implementation Plan

Status: **revision 2.3, 2026-10-06; awaits Geoff's approval.**

History:

- Revision 1 at `5afe46b`.
- Revision 2 folds the three lens reviews (contract, mechanics, and risk) per
  `docs/superpowers/research/2026-09-26-rung-6d-plan-fold.md`.
- Revision 2.1 folds the verification read (`2026-09-26-rung-6d-plan-fold-verification.md`) per that
  record's "Second fold" section.
- Revision 2.2 (`5ed193f`) folds the prose review (`2026-09-26-rung-6d-plan-prose-review.md`) per the
  fold record's "Prose fold" section.
- Revision 2.3 rebases the plan onto master at `0cb33ae`, after rung 5's close, the navidrome-removal
  pass, and the rungs 12 to 16 re-plan. The fold record is
  `docs/superpowers/research/2026-10-06-rung-6d-plan-rebase.md`. Revision 2.2's rulings were never
  answered, and its approval gate is superseded by this revision's.

**Citation basis.** Every code, spec, and plan line below is cited at `0cb33ae` on master unless the
citation names another commit. Task 0 re-resolves only the pointers into files that change after
`0cb33ae`.

**Pass class:** `auth-data` for Tasks 0 to 8, `docs` for Task 9 (the removal pass's precedent).
**Ceiling (Ruling 1):** 10.5M execution, flag 8.4M, under the recommended answers to Rulings 3 and 4.
Planning spend is recorded against rung 6, outside this ceiling. The ceiling is re-derived below,
because the rebase changes scope: Task 0 gains the dependency sweep, Task 8 loses its `deploy.sh`
restart and `check` work to rung 16, and Task 5 shrinks under Ruling 4's yes.
**Checkpoint interval:** each segment boundary (at most three tasks apart), and the close.
**Segments:** A (Tasks 0, 1, and 2), B (Tasks 3, 4, and 5, then the security read of B), C (Tasks 6,
7, and 8, then the security read of C, the simplifier round, and the merge), then Task 9, the close.
**Tasks:** ten (Task 0, eight build tasks, and the close), at the spec's ten-task line (spec
`:1166-1169`) and not past it. A "no" on Ruling 3 adds a task and splits the pass (Ruling 1). A "no"
on Ruling 4 keeps ten tasks and raises the ceiling.

Sources, in order of authority:

- The ratified rung 6 spec (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`, revision
  2.2 with errata E1 to E11 and the 2026-10-06 `0003` erratum, "the spec"): owner ruling 1
  (`:52-59`), the 6d pass section and criteria 1 to 14 (`:1105-1138`), "The settings home" and "The
  seed handoff" (`:737-819`), "Identity in the store" (`:821-893`), "Three open modes" and
  "Migrations" (`:287-451`), "Package territory" (`:194-225`), "The 2.0 forward hooks" (`:896-912`),
  "Prior art" (`:914-936`), "Planning-miss prevention" (`:938-958`), "Band and ceilings"
  (`:1140-1169`), "Contested surfaces" (`:1171-1180`), and "Not verified" (`:1273-1284`).
- The rungs 12 to 16 re-plan spec (`docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`,
  revision 2.1, approved, "the re-plan"): rulings 1 to 7 (`:36-70`), the four things revision 2.3
  carries (`:202-209`), "The username rename" and its roster step (`:293-331`), and "Where the carried
  items land" (`:410-445`).
- The configuration homes spec (`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`,
  CH), above all the `sensitive` paragraph (CH `:167-173`) and its 2026-10-05 erratum (CH `:12-29`),
  which touches no 6d surface.
- What earlier passes handed 6d: `docs/STATUS.md`'s "To 6d" line; `docs/HISTORY.md`'s 6c entry ("The
  6c to 6d handoff" items 1 to 6, "To 5d and 6d", errata 6c-E5, 6c-E6, 6c-E10); 6b-ii-b's "To 6d"
  (`ServeOptions.Reconcile`, the store identity, the network actor); 6b-i's "To 6d" (the store
  identity, R9 option b); 6a's open items (serve's settle loop, the logger guard; the migration wait
  registering at 6d, E1); 5c-iv's "To 6d Task 4" (`rosterResetServer`'s baseline); 5c-ii's
  re-validation item (every existing row's username, routed to 6d by the re-plan `:432`).
- The 6c architecture reads (`docs/superpowers/research/2026-09-26-rung-6c-architecture-reads.md`):
  `internal/engine/store` finding 1 (`OpenServe` has no production caller; `OpenWrite`'s "start serve
  first" cannot be satisfied on a fresh box) and `internal/settings` finding 1 (the apply is unexported
  and uncalled while `cmd/dubplate-server/backup.go:274` and `store.go:286-307` promise serve applies
  settings at its next start). The removal pass's reads repeat the store finding (HISTORY, "Navidrome
  removal", reviewer dispositions). 6d wires both.
- The 0.8 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, revision 3,
  "S20"): "Forward hooks the current build owes" (`:767-854`), "Prior art" (`:856-891`), and
  "Identity and app passwords" (`:478-511`).
- `docs/server-runbook.md` at `0cb33ae`: "Install order" (`:14-214`), "Re-create a roster written
  before the username rename" (`:216-250`), "The rollback copies" (`:800-850`).
- The repo `CLAUDE.md`, the 6c plan (`2026-09-25-rung-6c-routed-schema-settings-home.md`) for shape.
- Labels this plan borrows: B1, M1, and M2 are revision 2's verification read's blocker and majors
  (`2026-09-26-rung-6d-plan-fold-verification.md`); D10 is the 6c plan's deviation row D10 (the local
  keep's prune proof moved to 6d); R5S is the rung 5 spec (`2026-09-08-rung-5-identity-auth.md`). "Ruling
  N" alone means this plan's ruling; another document's ruling is named with its document.

**What this pass is.** It makes `serve` open the store through `store.OpenServe`, creating or
migrating it, finishing a reconcile-pending restore, and then applying the seed rule before the
listener opens (spec criteria 11 and 12, the two wiring items the 6c reads route here). It moves the
roster into the store: `internal/auth`'s roster reads and writes the `users` table, the resolution
point returns `Identity.UserID`, and `DUBPLATE_ROSTER_FILE` retires (criteria 4 to 6). It re-homes rung
5's ten auth thresholds as store-homed settings with accessor reads at every call site (criteria 9 and
10). It lands the deployment side the box needs: the store leaf's grant, `deploy.sh`'s leaf list, the
state-dir leg, the template, and the runbook's install order and 6d upgrade (criteria 13 and 14's
template and leaf halves).

**What this pass is not.** It deploys nothing. It builds no HTTP route, no admin write path (rung 14),
no member removal, demotion, or revoke verb, no web page change, and no `ops/local` runner inside
`serve`. It never enables or restarts the serve unit and adds no `check` step to `deploy.sh`: both are
rung 16's (the re-plan `:414-417`). It builds no dual-key epoch (R-6d20). Under Rulings 3 and 4's
recommended answers it builds neither the store identity nor `member import`.

---

## Rulings

Three questions, ordered by how much build each answer changes. Every other decision is a Resolution
below, taken on its architectural merit and reported. Revision 2.2's Ruling 2 (the dual-key epoch)
is now resolution R-6d20, because the repo's standing rule decides it; its number is not reused.

**1. Does 6d run as one pass under a 10.5M execution ceiling, flag 8.4M, in three segments?**
Recommendation: yes. Reason: the ceiling is derived from 6c's actual rate, the nearest analog, and
ten tasks sit at the spec's split line without crossing it ("Token ceiling"). What yes builds: one
pass, three segments, one close. What no builds: the conductor re-cuts at the B/C seam into 6d-i
(Tasks 0 to 5 and a close) and 6d-ii (Tasks 6 to 8 and a close), about 0.5M more for the second Task
0 and close. A "no" on Ruling 3 takes the task count past ten, and that split happens regardless,
with each half's ceiling re-derived from the "Token ceiling" arithmetic.

**3. Does the store identity route to rung 16, where backups first run on a timer, rather than land
in 6d?** The store identity is 6b-i's R9 option b: a random id minted at create, a remote marker, and
backup and restore refusing another store's remote or snapshot. The re-plan already lists it under
rung 16 "if 6d's Ruling 3 is yes" (`:422`).

Recommendation: yes. Reasons:

- Nothing binds it to create, because a migration can mint the id for an existing store at any later
  version.
- Its exposure, two servers pointed at one remote, starts when backup runs unattended. Rung 16's
  timer begins that, and backup has never been enabled on the box (HISTORY, "Navidrome removal").
- Its consumers are in `internal/backup`, which 6d otherwise touches only for one exported
  reconciler.

What yes builds: nothing in 6d. STATUS's rung 16 line keeps the item with R9's text.

What no builds: a Task 8b (a singleton identity row minted by a migration, the backup-run marker
check refusing a foreign remote, and restore refusing a foreign snapshot), about 0.9M, and the pass
splits under Ruling 1.

**4. Does `member import` drop from 6d, so `DUBPLATE_ROSTER_FILE` retires straight to `member add
--admin`?** This reverses the import half of the rung 6 spec's owner ruling 1 (`:52-59`), which built
the verb as the 6d upgrade's path. The key's retirement, ruling 1's other half, stands either way.

Recommendation: yes. Reasons:

- No roster file the verb could read exists. Nothing was released, so no operator outside the box
  holds one (the re-plan `:305`).
- The box's one file carries the pre-rename `streamer_username` key, and the loader refuses an
  unknown member (`roster.go:340`, quoted). The re-plan's roster step already replaces that file with
  one `member add --admin` (`:323-331`).
- An operator switching from Navidrome brings members through `dubplate import navidrome` at slice
  8 (the re-plan's ruling 2), never through a dubplate roster file.
- The repo builds a surface when its first consumer arrives. This verb has no consumer, and it would
  add a migration, a once-only record table, and fourteen states nobody meets.

What yes builds: Task 3 adds no migration; Task 5 retires the key with a message naming `member add
--admin`; Task 4 deletes the file-backed roster code whole. Criteria 1 to 3 are not built (6d-R4), and
criterion 4 names `member add` (6d-R13). Ceiling 10.5M, flag 8.4M.

What no builds: the appendix, "What Ruling 4's no builds," in place of Task 5's outcomes, migration
`0006` in Task 3, and the import record's guard in Task 4's members row. Ceiling 11.0M, flag 8.8M;
still ten tasks.

---

## Goal

After 6d, a fresh box's `serve` creates the store on its first start; an upgrade's `serve` migrates it
behind a pre-migration copy; the seed rule runs before any request; every member lives in `users`
with a surrogate id the resolution point returns; no verb reads a roster file; the ten auth thresholds
are settings read at their point of use; and the deploy files grant and pre-create the store leaf.
Every operator string that says "serve applies it at its next start" is true.

## Owner touchpoints

The plan-approval gate and Rulings 1, 3, and 4, and nothing else. No attended sitting: nothing
deploys in 6d. The box today runs binary `d1e93e9` with the serve unit installed and disabled, the
`auth`, `roster`, and `unidecode` leaves, no store leaf, and one admin in a roster file under the
pre-rename key (STATUS; the re-plan `:323`). The 6d upgrade rides the next ruled deploy after 6d merges, in the words of
the runbook section Task 8 writes:

1. Take the rollback copies (runbook, "The rollback copies"), then run `deploy.sh --apply`. The
   rendered env carries no roster row, the deploy pre-creates the `store` leaf, and it starts nothing.
2. Run one supervised start and stop of the serve unit. The start creates the store. `doctor`'s
   members row then fails naming `member add --admin`, as expected here.
3. Move `roster.json` aside, as the re-plan's roster step 1 does.
4. Run `member add <address> --admin --name ... --username ...` through `systemd-run` with the store
   leaf granted, then `member list` (one admin) and `doctor` (exit 0).

Apart from the supervised start in step 2, nothing in the upgrade starts `serve`, and nothing enables
it. Enabling the unit is rung 16's.

## Token ceiling: 10.5M execution (Ruling 1)

**Basis: 6c's actual** (`docs/HISTORY.md`, 6c "Both budgets", `:1210-1220`): about 4.5M before the
merge for Task 0 (about 0.1M), five build tasks, four fix rounds, and the simplifier round with its
review; so (4.5 − 0.1) / 5 ≈ **0.88M per build task**, fix rounds and the simplifier share included.
The merge ritual and six architecture reads cost about 0.75M, about **0.11M per read**. The removal
pass (HISTORY, "Navidrome removal", Budgets: about 3.7M for seven mechanical build tasks, two folds,
ten reads, and the close) is the evidence for a lighter rate on a task with no new behavior.

1. Task 0 (`opus`, fresh context): the dependency sweep (0.15M, the removal pass's own Task 0 with one
   bump) and the staleness read, now scoped to files changed since `0cb33ae` (0.10M): **0.25M**.
2. Tasks 1, 2, 3, 4, 6, and 7 at 0.88M: **5.28M**.
3. Tasks 5 and 8 at half rate, 0.44M each: **0.88M**. Under Ruling 4's yes, Task 5 retires a key and
   deletes test environments' copies with no new behavior. Task 8 lost the `deploy.sh` restart and
   `check` rework and the `check` holder line to rung 16, leaving a unit grant, a leaf list, a gate leg
   extension, one refusal string, and the runbook.
4. Two `web-auth-security-reviewer` reads (after segments B and C) and their fold rounds, the spec's
   own allowance (spec `:1160`): **0.70M**.
5. The close: Task 9's drafts (0.15M), the merge ritual (0.10M), seven architecture reads
   (`internal/engine/store`, `internal/settings`, `internal/config`, `internal/backup`,
   `internal/auth`, `internal/identity`, `cmd/dubplate-server`) at 0.11M (0.77M): **1.02M**.
6. **Projection 8.13M.** 8.13 / 0.8 = 10.16, rounded up to the next 0.5M: **10.5M**, flag at 80%,
   **8.4M**.

Variants. A split includes about 0.5M for its second Task 0 and close (Ruling 1).

- Ruling 4's no (Task 5 at the full rate, the migration inside Task 3's rate): 8.13 + 0.44 = 8.57M;
  8.57 / 0.8 = 10.71, ceiling **11.0M**, flag **8.8M**; ten tasks, no split.
- Ruling 3's no: 8.13 + 0.88 + 0.5 = 9.51M projected; 9.51 / 0.8 = 11.89, ceiling **12.0M**, flag
  **9.6M**, split.
- Both no: 8.57 + 0.88 + 0.5 = 9.95M; 9.95 / 0.8 = 12.44, ceiling **12.5M**, flag **10.0M**, split.

Revision 2.2's ceiling was 11.5M on a projection of 8.91M. The difference is Task 0's sweep (+0.10M),
Task 8's routed work (−0.44M), and Task 5 under Ruling 4's yes (−0.44M). The spec's own row was 6.6M
for nine tasks (spec `:1160`), priced at 0.46M per task (spec `:1147-1149`); 6c's actual rate is the
better evidence, and the spec requires re-deriving from the previous pass's actual (spec
`:1166-1168`).

Deliverable count: twenty-eight across Tasks 1 to 8 (Tasks 1, 4, 6, 7, and 8 carry four; Tasks 3
and 5 carry three; Task 2 carries two), plus Task 0's two and Task 9's three. No task passes four.

---

## Primary inputs

### Code pointers (at `0cb33ae`)

- `cmd/dubplate-server/serve.go`: `rosterFileKey` (`:29`); `checkDirWritable` (`:150-155`) and
  `checkRosterDirWritable` (`:157-166`); `newServeCmd` (`:222-260`: `config.Load`,
  `resolveServeConfig`, `resolveAuthConfig` at `:247`, `newServeShell`, then `serve.Run`, the only
  listener); `newServeShell` (`:263-386`, signature `(ctx, environ, prefix, bindAddrs, cfg, authCfg,
  logger) (*serve.Shell, error)` since the removal pass's F1): `auth.Open` (`:316`),
  `auth.OpenNonceStore` with `magicLinkExpiryCeiling` (`:320`), `auth.OpenRoster` and
  `checkRosterDirWritable` (`:326-332`), `auth.NewLimiter` (`:334-339`), `auth.NewMagicLinkSeam`
  (`:346-349`), `privateFiles.roster` (`:375`), the nonce sweep (`:384`); `authHandlersConfig`
  (`:426-449`); `sweepNonces` (`:485-489`, revision 2.2's `startNonceSweep`). Nothing here references
  `store` or `settings`.
- `cmd/dubplate-server/auth.go`: `magicLinkSource` (`:43-62`, `Required` lists `DUBPLATE_ROSTER_FILE`
  first at `:47`); the bound constants (`:64-171`); the ten thresholds (`:107-221`, the answer at
  `:210-215` with `Kind: config.KindString`); `registerAuthConfig` (`:225-238`);
  `validateMagicLinkSource` (`:265-338`, the environ-based bounds and the two relational checks at
  `:303-322`); `authConfig` (`:499-513`); `resolveAuthConfig` (`:518-563`).
- `cmd/dubplate-server/store.go`: `storeBusyTimeoutThreshold` (`:23`), `jobRetentionSetting` (`:55`),
  `storeLocalSnapshotKeepSetting` (`:71`), `registerStoreConfig` (`:84`), `storeBusyTimeout` (`:106`),
  `storeRow` (`:134-150`, returning the `store.Report` it read), `storeVouches` (`:191-197`, absent,
  older, and current vouch), `settingsDoctorRows` (`:226-252`), `settingsOutcomeStatus` and
  `settingsChangeNote` (`:270-296`), `settingsNoTableStatus` (`:302-310`), `settingShown` (`:333-338`,
  the sensitive-withholding branch). The "serve's next start" strings are at `:290`, `:292`, `:294`,
  and `:307`. STATUS cites them as `:280-284`; the content sits at `:286-307`.
- `cmd/dubplate-server/backup.go`: `logSettingsEvaluation` (`:262-277`, the "serve applies it at its
  next start" line at `:274`); `runBackup`'s `store.OpenWrite` (`:211`) and `settings.NewAccessor`
  (`:230`).
- `cmd/dubplate-server/check.go`: the store row and the settings rows (`:23-81`); the settings
  evaluation runs only when `r.Facts != nil` (`:62-68`), so an absent store evaluates nothing
  (R-6d21). `printSettingsRows` (`:83-102`).
- `cmd/dubplate-server/member.go`: `memberVerbWriteWait` (`:26`, used at `:95`),
  `errForeignRosterOwner` (`:39`), `newMemberCmd` (`:46-54`, registering `add` and `list`),
  `newMemberAddCmd` (`:63-111`, `--name`, `--username`, `--admin`), `addMember` (`:119-140`,
  `auth.OpenRoster`/`auth.CreateRoster`), `newMemberListCmd` (`:142-171`, `List()` at `:162`),
  `checkRosterDirOwner` (`:192-210`), `resolveActor` (`:212-223`).
- `cmd/dubplate-server/doctor.go`: the roster row (`:87-114`, printing only on failure),
  `selectedRosterPath` (`:212-224`), `rosterRowStatus` (`:226-247`, `List()` at `:236`),
  `rosterStatus` (`:340`), the exit's flag chain (`:198-201`). `doctor_test.go`'s
  `TestDoctorCmd_LineCountFollowsRegistry` (`:272`, conductor ruling R10) counts every unconditional
  row as a named constant.
- `cmd/dubplate-server/web.go`: `privateFiles` (`:111-117`) and its roster check (`:691-693`).
  `cmd/dubplate-server/doctorprobe.go`: a second `privateFiles` literal reading `rosterFileKey`
  (`:103-106`).
- `cmd/dubplate-server/main.go`: `newProductionRegistry` (`:99-108`) and `newRootCmd`'s
  `AddCommand` block (`:126-135`). Under Ruling 4's yes, 6d adds no command and edits neither block;
  under its no, `member import` registers in `member.go`'s `newMemberCmd`, not here.
- `internal/engine/store/open.go`: `ServeOptions` (`:85-116`: `BusyTimeout`, `MigrationWait`, `Keep`,
  `Backup`, `Logger`, and `Reconcile func(ctx, *Handle) ([]byte, error)` at `:108-115`); `OpenWrite`
  (`:118-185`, "start serve first" at `:122` and `noStore` at `:655-657`); `OpenServe` (`:200-270`,
  the nil-logger refusal at `:220-222`, the loop at `:248-269`); `openShared` (`:272-300`); `settle`
  (`:302-356`); `finishRestore` (`:358-372`); `waitLock` (`:374-410`, bounded per wait by
  `MigrationWait`, logging holders); `statLeaf`'s owner check (`:440`).
- `internal/engine/store/inspect.go`: `Report.Holders` (`:104-106`); `jobsVersion` and
  `settingsVersion` (`:122-127`); `Facts` (`:132-161`, no member counts); `Inspect` returns early with
  no facts on an intent, on `StateBusy`, and below `jobsVersion` (`:205-242`); `readFacts`
  (`:252-290`). `write.go`: `errNoDeadline` (`:23-25`), `Handle.Update` (`:44-64`, the deadline check
  at `:68-70`), `Row.Scan` (`:178-180`). `conn.go`: `Handle.QueryRow` (`:107-112`). `id.go`: `NewID`
  (`:8-11`). `lock.go`: `tryLock` (`:85-93`). `leaf.go`: `leafDirName = "store"` (`:15`),
  `leafDirPerm = 0o700` (`:19`), the `geteuid` seam (`:57-63`), `create` (`:64-77`), `checkOwner`
  (`:101-114`, its `runuser` hint at `:113`, asserted by `leaf_test.go:129` and `:179`,
  `inspect_test.go:182`, `open_test.go:633`, and `restore_test.go:128-134`). `migrate.go`: a current
  store returns before any snapshot or prune (`:121-123`). Migrations `0001` to `0005`; `users` is
  `0003_catalog.sql:1-10`.
- `internal/settings`: `apply` (`apply.go:19-63`, unexported) and `writeChanges` (`:65`);
  `Evaluate` (`evaluate.go:117`), `EvaluateFacts` (`:128-148`, which returns `evaluateTableAbsent`
  for a store below the settings table), `evaluateTableAbsent` (`:150-163`, which "never reads
  environ"), `outcome` (`:268-337`), `parse` (`:368-378`); `Accessor` (`accessor.go:22-90`: `Int`,
  `Duration`, `Float`; no string read).
- `internal/config/settings.go`: `Setting` (`:55-75`; `Kind` refuses `KindString`; `Sensitive` at
  `:69-70` documented as "a value a report must not print"), `CorrelatedRule` (`:96`),
  `buildValidator` (`:118-131`), `boundedValidator` (`:136-150`). `registry.go`: `claim`
  (`:125-130`), `AddSetting` (`:431`), `AddCorrelatedRule` (`:461`), `AddSuperseded` (`:495-509`, a
  `Registry` method since the removal pass's simplifier). `config.go`: `rejectNamespace` (`:283-315`),
  which treats a set but empty key as unset (`:288-290`).
- `internal/backup`: `reconcile` (`reconcile.go:55`) and `report.render` (`:151`), unexported;
  `restorer.report` (`restore.go:629-660`: the root is the record's `LibraryRoot`, else the configured
  root, `:633`; the head is the stamp, a `source:` line from `sourceLine` at `:674`, and the
  displaced line from `displacedLine` at `:683`). `store.RestoreIntent` (`internal/engine/store/
  restore.go:580-593`).
- `internal/auth`: `roster.go`: the error sentinels (`:105-160`; `errDuplicatePrincipal` at `:118`,
  `errMissingUsername` at `:141`, `errNoAdmin` at `:151`, `errActorNotAdmin` at `:156`);
  `errRosterWritableByOthers` (`:237-241`); `OpenRoster` (`:258`), `CreateRoster` (`:277`),
  `readRoster` (`:300-334`, the mode refusal at `:318-320`); `parseRoster` (`:338-358`,
  `json.RejectUnknownMembers(true)` at `:340`); the file write and lock (`:360-506`);
  `validateRoster` (`:530-564`); `memberRecord.validate` (`:566-596`); `validateWriteCandidate`
  (`:598-622`); `Lookup` (`:703`), `List` (`:727`), `AddMember` (`:737-771`), `checkAddition`
  (`:773-790`), `addTo` (`:792-814`), `Member.principal` (`:815-822`, falling back to the address
  when the username is empty), `principalCollision` (`:824-842`). `seam.go`: `NewMagicLinkSeam`
  (`:81`), `VerifyIdentity`'s `Lookup` (`:112`) and `principal()` (`:120`). `handlers.go`:
  `HandlersConfig`'s `LinkExpiry`, `SessionLifetime`, `SessionAbsoluteCap` (`:93-98`), `BodyCap`
  (`:109-111`), `GenericAnswer` (`:113-118`); `Lookup` at `:209` and `:365`, `List` at `:307`.
  `admin.go`: `outcomeDuplicatePrincipal` (`:38`), the gate's `Valid()` checks (`:92`, `:121`),
  `rosterAPI.writeWait` (`:165-169`, the deadline at `:255`), `List` at `:206`, the outcome mapping
  (`:296-315`, `errDuplicatePrincipal` at `:308`). `nonce.go`: `linkExpiryCeiling`'s doc (`:69-76`),
  `OpenNonceStore` (`:93`), `RunSweep` (`:123`), `Insert` (`:149`), `Sweep` (`:200`). `limit.go`:
  `Limits` (`:37`), `NewLimiter` (`:67`). `session.go`: `RefreshSession` (`:138`), `clampExpiry`
  (`:144-147`), `VerifySession` (`:213-219`). `epoch.go`: `Bump` (`:103`).
- `internal/identity/identity.go`: `Identity{Address, Principal, DisplayName, Source, Channel}`
  (`:184-201`; `Principal`'s doc at `:190-193` still promises the surrogate-id substitution), `Valid`
  (`:203-210`, which already refuses `ChannelCommandLine`). `internal/ops/ops.go`: `NetworkActor`
  ("member:<principal>", `:47-52`), which panics unless `id.Valid()`. Two test fixtures build a
  cookie identity with no id: `validIdentity()` (`internal/ops/ops_test.go:38-46`) and `ann`
  (`internal/ops/opshttp/helpers_test.go:21-28`).
- `cmd/dubplate-server/rendered_test.go` never calls `newServeShell` itself: its two builders
  (`startRenderedServer` `:430`, `startFixtureServer` `:451`) call `buildServeShellWithLogger`, which
  lives in `serve_test.go` (`:67-81`) and calls `resolveAuthConfig` and `newServeShell` directly
  (revision 2.2's `buildServeShellWithContext` is gone). `serve_test.go` also calls `newServeShell`
  at `:596-610` and `resolveAuthConfig` at `:603` and `:690-700`. The rendered harness is built on the
  roster file: `renderedEnviron` (`:378-400`) writes `roster.json` (two members, neither with a
  `username`) and sets `DUBPLATE_ROSTER_FILE`; `checkRateBudget` (`:494-500`) and `mintLinkFor`
  (`:541-545`) call `resolveAuthConfig`; `rosterResetServer` (`:618-670`) resets the capture matrix by
  rewriting that file, with "the file as the first POST finds it" as its baseline, and answers each
  POST `{"members": n}`, the contract `web/scripts/rendered/run.ts` reads (`captureRosterResetURL` at
  `:202`, `resetCaptureRoster` at `:4778-4783`, called at `:5302`); it is constructed at `:861`; and
  the `members/roster-file` subtest (`:924-933`) proves the add-member write by reading the file.
  The usernames `run.ts` types on a real write are `newmember1` and the capture matrix's
  `captureUsername(kind)` values; `newmember1x`, `newmember2`, and `inherited` are typed on refused or
  fulfilled submits (`run.ts:3562-3626`, `:4203`, `:5040`, `:5075`).
- `cmd/dubplate-server/templatekeys_test.go`: `renderedDoctor` (`:291-326`, which sets
  `DUBPLATE_ROSTER_FILE` at `:310`) and `TestRenderedTemplateDoctor` (`:328-410`), whose cases at
  `:347-386` assert the roster row (exit 0 with a fixture roster; "missing"; "no admin"; and "deleting
  the roster row" failing the magic-link source).
- Test environments that set `DUBPLATE_ROSTER_FILE`: `auth_test.go`, `backup_test.go`,
  `check_test.go`, `doctor_test.go`, `member_test.go`, `rendered_test.go`, `serve_test.go`, and
  `templatekeys_test.go` (`grep -ln DUBPLATE_ROSTER_FILE cmd/dubplate-server/*_test.go`).
- Deploy files: `systemd/dubplate-server-serve.service` (`Type=exec`, `User=music`,
  `EnvironmentFile=/etc/dubplate/server.env`, `Restart=on-failure`, `RestartSec=5`, and
  `ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/roster` with its derivation comment);
  `scripts/deploy.sh` (header `:27-36`, "It never enables, starts, or restarts a unit"; `UNIT_FILES`
  `:42-44`; `DUBPLATE_LIB_DIR` `:264-266`; the leaf table comment and `DUBPLATE_SERVER_LEAVES=(auth
  roster)` with its loop `:268-285`; the unidecode leaf's comment naming "the auth store and the
  roster" `:287-298`; the reload's comment, "Enabling or restarting the serve unit is the operator's
  own step" `:422-426`); `env/dubplate-server.env.template` (the header's "nine required keys"
  `:63-64` and the `DUBPLATE_ROSTER_FILE` row with its comment `:66-70`); `scripts/check.sh`'s
  state-dir leg (`:657-737`), which already refuses a prefixed entry and a unit leaf missing from the
  array, but not an array leaf no unit grants.
- `docs/server-runbook.md`: install steps 5, 6, and 8 name the `roster` leaf and grant (`:73-168`);
  step 11 starts the unit after `doctor` (`:197-205`); the re-create section (`:216-250`); the
  `roster` doctor row (`:265`); the staged swap's "roster beside it" (`:773`).
- `web/src/routes/members/+page.svelte`: the username input is `required` with the write path's
  pattern (`:304-322`); the roster line's `member.username !== ''` conditionals and their "a blank
  username leaves no separator" comment (`:252-258`); the `duplicate_principal` field mapping
  (`:53`). `web/src/lib/copy/deck.ts:150`: the `duplicate_principal` copy. 6d changes no `web/` file
  (R-6d17).

### Behaviors this plan relies on, quoted

| Behavior | Quote | Where 6d relies on it |
| --- | --- | --- |
| `database/sql` `Row.Scan` (`go doc database/sql Row.Scan`, Go 1.27.1) | "If more than one row matches the query, Scan uses the first row and discards the rest. If no row matches the query, Scan returns ErrNoRows." | `Lookup` on an absent or tombstoned address |
| `store.Row.Scan` (`write.go:178-180`) | "Scan copies the row's columns into dest, as [sql.Row.Scan] does, and records any error but [sql.ErrNoRows], which it returns." | A collision read inside `Update` that finds nothing does not poison the transaction |
| `store.Handle.Update` (`write.go:44-64`) | "commits only when fn returns nil and no statement fn issued through tx failed"; "ctx must carry a deadline, which bounds the wait for the write pool's one connection inside this process. The wait on another process's write lock is bounded by the busy timeout alone"; "The first statement reads user_version, and Update refuses without calling fn when it differs from the binary's version." | Every write: add, the seed apply |
| `errNoDeadline` (`write.go:23-25`) | "errNoDeadline is returned by a write whose context carries no deadline, before it waits on anything." `var errNoDeadline = errors.New("store write needs a context deadline")` | R-6d5's named write wait |
| `store.NewID` (`id.go:8-11`) | "NewID returns 16 random bytes for a new entity's primary key ... It reports no error because [rand.Read] never returns one" | The surrogate id's one mint |
| SQLite `NOCASE` (sqlite.org/datatype3.html) | "the 26 upper case characters of ASCII are folded to their lower case equivalents before the comparison is performed. Note that only ASCII characters are case folded." | `users.username ... UNIQUE COLLATE NOCASE` agrees with a case-insensitive Go comparison only because the write path's pattern admits ASCII alone; Task 3 keeps the pattern check before the insert |
| `flock(2)`, `LOCK_SH` (`man 2 flock`) | "More than one process may hold a shared lock for a given file at a given time." | A current-version `serve` start never waits on a running backup; only a migrating start (exclusive) does |
| `flock(2)`, one process, two descriptors (`man 2 flock`) | "If a process uses open(2) (or similar) to obtain more than one file descriptor for the same file, these file descriptors are treated independently by flock()." `tryLock` opens a fresh descriptor per attempt (`lock.go:85-90`) | Task 4: the rendered harness's reset server may take its own shared write-mode handle beside the serve shell's in one test process |
| `flock(2)` conversion (`man 2 flock`; spec `:375-376`) | "Converting a lock (shared to exclusive, or vice versa) is not guaranteed to be atomic: the existing lock is first removed, and then a new lock is established." | Already honored by `OpenServe` (`open.go:211-217`); 6d adds no conversion |
| `systemd.exec(5)`, `ReadWritePaths=` | "Paths in ReadWritePaths=, ReadOnlyPaths=, InaccessiblePaths=, ExecPaths= and NoExecPaths= may be prefixed with "-", in which case they will be ignored when they do not exist." | The store leaf is granted unprefixed, so `deploy.sh` must pre-create it (CLAUDE.md's standing fact) |
| `runuser(1)` | "runuser defaults to not changing the current directory and to setting only the environment variables HOME and SHELL (plus USER and LOGNAME if the target user is not root)." | `runuser` loads no environment file, so `checkOwner`'s hint routes an operator to a command that fails the wall load (R-6d15) |
| `encoding/json/v2` `RejectUnknownMembers` (`go doc encoding/json/v2 RejectUnknownMembers`) | "RejectUnknownMembers specifies that unknown members should be rejected when unmarshaling a JSON object." | Ruling 4: the box's file, under the pre-rename key, cannot be read by `parseRoster` |
| `config.Setting` (`settings.go:55-75`) | "A string has no bounds, so KindString is refused at registration." `buildValidator` panics on any other kind (`:129`); `boundedValidator` panics when a floor or ceiling does not parse (`:139`, `:143`). | `DUBPLATE_UNKNOWN_ADDRESS_ANSWER` and `DUBPLATE_SESSION_LIFETIME` cannot register as settings unchanged (R-6d7, R-6d8) |
| `Registry.AddSuperseded` (`registry.go:495-509`) | registers "name as a pre-pivot key with no meaning in the greenfield wall, reason being the message rejectNamespace reports for it"; `requirePrefix` demands `DUBPLATE_` (`:112-116`) | Retiring `DUBPLATE_ROSTER_FILE` (R-6d12) |
| `wall.rejectNamespace` (`config.go:288-290`) | `continue // set but empty: treated as unset, nothing to flag` | Task 5's empty-value state |
| `Registry.claim` (`registry.go:125-130`) | panics "if key was already claimed by a family other than family" | Task 6: a key cannot be a threshold and a setting at once |
| CH, `sensitive` (CH `:167-173`) | "Some settings are dangerous to change live even though they are not secrets ... The registry marks these `sensitive`. A sensitive write requires fresh re-authentication ... A settings key never holds a credential." | R-6d6 |
| `settings` `outcome` (`evaluate.go:268-274`) | "A bound applies to the value in force after the seed, so an out-of-bounds environment-seeded row that the environment re-seeds or removes is healed rather than refused." `Outcome.Value` is "the value in force once the seed applies" (`:87-89`). | R-6d4: the keep read cannot use `Outcome.Value` for an in-bounds row |
| `settings` `parse` (`evaluate.go:368-378`) | `default: return nil, fmt.Errorf("setting kind %v has no parser", kind)`; `outcome` parses every stored row | R-6d7: a closed-set string needs a `parse` arm |
| `settings` `evaluateTableAbsent` (`evaluate.go:150-155`) | "It never reads environ: a store with pending migrations reads no setting from the environment ahead of its own seed, so the report only ever says the store holds no settings yet." | R-6d21 |
| `store.Inspect` (`inspect.go:205-247`) | returns early with `Facts` nil on a restore intent and on `StateBusy`; reads facts for any store at `jobsVersion` (2) or above, and settings only at `settingsVersion` (4) or above; a failed facts read becomes the report's refusal (`reportFrom`) | R-6d4's no-facts states; R-6d13's version guards |
| `Handle.QueryRow` (`conn.go:107-112`) | "A read on the read pool sees the last committed state but holds no lock against a write that follows it, so it serves reports; a read that a write decides on runs inside that write's [Handle.Update] callback instead." | `Lookup` and `List` read the pool; every read `AddMember` decides on runs inside `Update` |
| `users` schema (`0003_catalog.sql:1-10`) | `address TEXT NOT NULL UNIQUE`, `username TEXT NOT NULL UNIQUE COLLATE NOCASE`, `admin INTEGER NOT NULL DEFAULT 0 CHECK (admin IN (0, 1))`, `tombstoned_at INTEGER`, `created_at INTEGER NOT NULL` | Both uniqueness constraints reach tombstoned rows (R-6d10) |
| `validateWriteCandidate` (`roster.go:598-603`) | "validates candidate the way validate does, then adds the write path's own rules: the display name is not blank, username is required, and its shape is the narrower pattern usernamePattern describes" | R-6d10, R-6d22 |
| `AddMember` (`roster.go:752-757`) | "On any other channel the actor must also hold the admin flag in the disk copy re-read under the lock ... Whatever the channel, the roster written must name at least one admin." (`errActorNotAdmin`, `errNoAdmin`) | R-6d10 carries both into the `Update` callback |
| `principalCollision` (`roster.go:824-829`) | returns "errDuplicateUsername when both principals are usernames, errDuplicatePrincipal when either is an address" | R-6d10: with every row carrying a username, `errDuplicatePrincipal` has no path |
| `clampExpiry`, `VerifySession` (`session.go:144-147`, `:215-218`) | `clampExpiry` "sets s.ExpiresAt to the earlier of s.IssuedAt plus lifetime and s.AnchoredAt plus absoluteCap"; "The order is fixed: the MAC first, then the absolute cap, then the expiry." | Task 7: a raised cap can admit a session only after a refresh under it |
| `NonceStore.linkExpiryCeiling` (`nonce.go:69-76`) | "mtimeFallback answers with max(retention, linkExpiryCeiling) rather than retention alone, so a record this store can neither read nor parse cannot be swept, and its nonce replayed, before every token that could have produced it has already expired." | Task 7: the argument stays the registered ceiling |
| `OpenServe` (`open.go:200-203`) | "It refuses opts without a Logger before it touches any file." | 6a's logger item closes with no code |
| `context.WithTimeout` (`go doc context WithTimeout`) | "WithTimeout returns WithDeadline(parent, time.Now().Add(timeout))." | R-6d5's write wait gives `Update` its deadline |

Task 0 re-quotes an in-repo row only when its file changed after `0cb33ae`, and a library row only
when the sweep moved its module. A changed quote is a note naming the new text; a removed behavior is
a blocker.

---

## Global constraints

- **Where it runs.** The main tree (lane 1). The conductor writes the lane registry row before Task 0.
  Slice 2 may run as a third code lane in its own worktree (the re-plan's ruling 5); it touches no 6d
  file but `go.mod` and `go.sum`, which the merge ritual reconciles. No other code lane may edit a
  file in "Contested surfaces" while 6d runs.
- **Go** follows the `go-conventions` skill: `ctx` first, wrapped errors with sentinels matched by
  `errors.Is`, table-driven tests under the assertion-discipline rules (no silent-success fakes, no
  vacuous assertion; each task names its mutations and reports each one failing).
- **Deploy files** follow the `vps-conventions` skill (Task 8).
- **Config keys** are declared through the registry in the rung's own file
  (`cmd/dubplate-server/store.go` for store keys, `auth.go` for auth keys), never in
  `internal/config/keys.go`.
- **Nothing exported ahead of a caller.** Each new export names its production caller in the task
  that adds it; the diff-reviewer checks it. `internal/auth` imports `store`, never `settings`
  (spec `:209-212`); `cmd/dubplate-server` injects each setting as a reader function.
- **The architecture bar** (the re-plan's ruling 6): a deletion reshapes what remains so it reads as
  if the removed part never existed; a task never trims a deletion to the compiler's complaints. A
  constant, sentinel, or helper whose last reader a task removes goes in that task.
- **Ease of setup** (the re-plan's ruling 7): no task adds a setup step, a required key, or a
  site-specific default. 6d removes one required key (`DUBPLATE_ROSTER_FILE`, a template literal),
  so the required set drops from 14 to 13 and the operator-typed set (8) is unchanged against the
  removal pass's baseline; the new wall threshold carries a default.
- **The gate**: `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`, the heavy
  lane (the web leg launches Chromium). On exit 75, re-issue the same command until it prints
  `gate exit:`. Never commit red.
- **No deploy, no box write.** Box reads are read-only ssh, and only Task 0 reads (item 6).
- **No `web/` edit** (R-6d17). `cmd/dubplate-server/rendered_test.go` changes only as the store move
  and the settings move require (Tasks 1, 4, 5, and 6 name each change). Its reset server keeps its
  HTTP contract (a POST answers `{"members": n}`), so `web/scripts/rendered/run.ts` needs no change.
- **Commits**: each task commits its own files, Conventional Commits, the repo's attribution footer.
- **Implementers never dispatch agents**; the conductor owns `code-simplifier` and every reviewer.

## Privileged fields and their channels

| Field | Every channel that may set it | Never |
| --- | --- | --- |
| `users.id` (the surrogate id) | the store's `NewID`, in the write path (`AddMember`) alone | a request body, a file, a CLI flag |
| `users.admin` | `member add --admin` (the CLI grant, 5b-ii's rule); a reactivation writes it from the validated candidate, so a cookie-channel reactivation writes 0 | the roster API and the add-member page (cookie channel), which refuse it as today; a stored flag surviving a reactivation |
| `users.username` | the add path on every channel (page, API, CLI) | no edit path exists; a reactivation (R-6d10) takes the new value as the add would |
| `users.tombstoned_at` | no channel sets it at 6d (no removal path); the write path's reactivation clears it | any request |
| any `users` row | `AddMember`'s one `Update` (R-6d22) | any other `INSERT INTO users` in production code |
| `Identity.UserID` | the resolution point (`Lookup`) alone, from `users.id` | a session field, a cookie, a request |
| `Identity.Principal` | the resolution point, from `users.username` | the session's address |
| settings source `env-seed`, and `default` by removal | `serve`'s seed apply at start alone (now wired) | `check`, `doctor`, `backup`, the member verbs (read-only evaluation) |
| settings source `admin` | none at 6d (rung 14); tests write one in-package | any 6d code path |
| `ServeOptions.Keep` | the registry value resolved through `store.Inspect` and `settings.EvaluateFacts` before `OpenServe` | the env directly |
| `ServeOptions.MigrationWait` | `DUBPLATE_STORE_MIGRATION_WAIT` (wall) | a setting |

Under Ruling 4's no, the appendix adds `member import` as a second channel for `users.id`,
`users.admin`, and `users.username`, and the `member_imports` row.

---

## Resolutions this plan takes rather than discovers

### R-6d1: placement

Rung 5 closed on 2026-10-05 (HISTORY 5d-ii), so 6d's spec gate (spec `:823-824`, `:1105`) is met.
6d executes in the main tree after approval. Revision 2.2's question of which deploy 6d must precede
is settled by the ladder: the box's serve unit stays disabled, and enabling it is rung 16's, which
follows 6d. STATUS's "To 6d" line still asks for the two wiring items "before any deploy exposing the
strings." The box's template carries no store-homed key, so no deployed `doctor` prints a seed note
today, and 6d lands both items before rung 16's first deploy that enables `serve`. Task 9 removes the
line.

### R-6d2: `serve`'s start order

`newServeCmd` runs: the wall load; the serve and store thresholds; the local-keep read (R-6d4);
`store.OpenServe` (create, complete, or migrate; finish a reconcile-pending restore); the seed apply
(R-6d5); an accessor over the handle; the auth construction, where every restart-mode setting is read
once (R-6d9); then `serve.Run`. The handle lives until `Run` returns and is closed then, releasing the
shared lock. Everything before `serve.Run` stays inside the one preparation function the verb calls
(today `newServeShell`, which gains the store open and the seed apply and hands the verb what it
needs to close the handle), and `serve.Run` stays the only call in the verb that opens a listener.
The start-order proofs therefore test that function: its return is the moment the listener could
first open, and a refusal returns an error with no shell, which leaves `serve.Run` nothing to run. No
test-only listen seam is added. The spec's "then recovery" (spec `:891-893`) reads as `OpenServe`'s
own staging cleanup and restore finish; job-record recovery is `ops/local`'s, and `serve` runs no job
until rung 14 mounts the first job route, so `serve` builds no runner at 6d. A store with an empty
`users` table starts: every login then answers as an unknown address, which is the spec's stated
outage window (owner ruling 1, `:56-57`).

### R-6d3: the migration wait

`DUBPLATE_STORE_MIGRATION_WAIT` registers as a wall threshold in `cmd/dubplate-server/store.go`
beside the busy timeout (spec `:813-814`, E1). Default **5m**, floor 1s, ceiling 24h. Reason: a
current-version start waits only on an exclusive holder (a restore), since shared locks coexist
(quoted); a migrating start waits on shared holders, a backup's sync above all, which on a first run
lasts hours. Five minutes outlasts every member verb and every snapshot copy a group-scale store
takes, and is far shorter than a first sync. On expiry `serve` exits nonzero naming the holders, and
the unit's `Restart=on-failure` with `RestartSec=5` (the landed unit) retries, inside its
`StartLimitBurst=30` per 300s. The bounds are range checks in the registering file, on
`storeBusyTimeout`'s precedent.

### R-6d4: the local keep reaches `OpenServe` through the facts

Per 6c-E10 and handoff item 1: before `OpenServe`, `serve` runs `store.Inspect` and
`settings.EvaluateFacts`, and passes `cfg.IntegrationEnabled` of the backup class as
`ServeOptions.Backup`. `ServeOptions.Keep` is extracted by one rule, on `settingsOutcomeStatus`'s
precedent in `store.go`, and never from `Outcome.Value` for an in-bounds row, because that is the
post-seed value (quoted):

- **Facts present and a keep row stored, in bounds** (`storeLocalSnapshotKeepSetting`'s validator
  passes): `Keep` is the row's value, even when the env differs. An env value that differs reaches
  the keep after this start's seed, the spec's stated order (spec `:414-416`).
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

R-6d21 changes what `EvaluateFacts` reports for a store below `0004`; this rule reads only a stored
row, so the change does not reach it.

### R-6d5: the seed apply is exported and called once

`internal/settings` exports the apply (its production caller is `serve`), keeping its one-transaction
contract. A refusal (an `admin` row disagreeing with the env, a stored value in force outside bounds,
an unparseable env value, a correlated rule failing) ends `serve`'s start nonzero with the joined
refusal, and the listener never opens. Each change logs one line after the commit, as `apply` does
today. `Update` refuses a context without a deadline (`errNoDeadline`, quoted), and `apply` sets
none, so the preparation function wraps the apply's context in `context.WithTimeout` under a named
write wait of 30s, on `memberVerbWriteWait`'s precedent (`member.go:26`, `:95`); the deadline bounds
only the in-process wait for the write connection, and the busy timeout bounds another process's
lock (`Update`'s doc, quoted). With it wired, `backup.go:274`'s line and `store.go`'s strings at
`:290-307` become true; Task 2 re-reads each and changes a word only if one is still false. Handoff
item 6 (a restored `admin` row may refuse the seed) needs no code: the refusal names the conflict,
and its resolving channel is rung 14's.

### R-6d6: `sensitive` means CH's marker, and the value is shown

CH defines `sensitive` as "dangerous to change live even though they are not secrets," gated by
re-authentication, and rules that "a settings key never holds a credential." 6c documented
`Setting.Sensitive` as "a value a report must not print" and withholds it in `doctor`
(`settingShown`, `store.go:333-338`) and in `Outcome.Report`. 6d marks both lifetimes and the absolute
cap sensitive (spec `:814-815`); under 6c's meaning, `doctor` would hide an operator's session
lifetime for no reason. So Task 6 re-documents the field to CH's meaning and removes the withholding,
and `doctor` shows the value and the marker. This is erratum 6d-E1 to 6c's R-6c3.

### R-6d7: a closed-set string setting

`DUBPLATE_UNKNOWN_ADDRESS_ANSWER` (`kind` or `generic`) is store-homed under the spec's table, but
`Setting` refuses `KindString` because a string has no floor or ceiling. Task 6 admits `KindString`
only with a non-empty list of allowed values, which is its bound: registration panics on an empty list
or a default outside it, the validator refuses any other value naming the set, and `Floor`/`Ceiling`
must be empty for it. `internal/settings`' `parse` gains a `KindString` arm returning the string
itself, so `equal` is exact byte equality; without it, `outcome` refuses every stored string row
(quoted), and the second `serve` start after a string seed refuses. The accessor gains a typed string
read, whose caller is the handlers' answer reader. Leaving the key on the wall was weighed and
declined: it is a tunable with a default, which CH homes in the store (CH `:106`).

### R-6d8: bounds and both correlated rules move into the registry

Every auth setting's bounds become `Floor`/`Ceiling` data, from the constants `auth.go` holds today
(`:64-171`). `DUBPLATE_SESSION_LIFETIME` has no ceiling today (`auth.go:304` passes 0); it takes the
absolute cap's ceiling (two years), since the lifetime can never usefully exceed the cap and the
correlated rule enforces that anyway. Both relational checks (`validateMagicLinkSource`'s
retention-at-least-expiry and lifetime-at-most-cap, `auth.go:303-322`) become registry correlated
rules, each error naming both keys, and 6c's test-only rule pair is replaced by the real pair; spec
criterion 5 is re-run over it (6c-E6, handoff item 2). `validateMagicLinkSource` keeps only the source
keys it owns (secrets, public URL, SMTP). Consequence, as the spec rules: a bad auth value no longer
fails every verb's load. `doctor` and `check` report it in every store state (R-6d21), `backup` logs
it, and `serve` refuses at the seed.

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
(the spec's one-constructor substitution, R5S `:195-197`). Its methods take `ctx` first, which the
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
  3. **The username:** any other row, live or tombstoned, whose `username` matches
     case-insensitively refuses with `errDuplicateUsername`, because the `UNIQUE COLLATE NOCASE`
     constraint reaches tombstoned rows (quoted) and a raw constraint error would answer 500. The
     reactivated row's own old username is not a collision.
  4. **At least one admin** (today's `errNoAdmin`): the table after the write must hold a live
     admin, so a non-admin `member add` into an empty table refuses naming `member add --admin`.

  A new row gets `store.NewID`. The admin grant rule carries over unchanged (5b-ii: the cookie
  channel refuses it). The audit record stays the log record (`logAddition`). The deadline `Update`
  requires is already set by both callers (`admin.go:255`, `member.go:95`). A refusal no longer
  names a file path, since there is none.
- **Every row carries a username** (`NOT NULL`, quoted), so `Member.principal`'s address fallback
  (`roster.go:815-822`) has no case. Task 4 deletes it with the file code, and with it
  `errDuplicatePrincipal` and `outcomeDuplicatePrincipal` (`admin.go:38`, `:308`), which only that
  fallback reached (quoted). The roster API then never answers `duplicate_principal`. The page's
  mapping and copy for it stay untouched (R-6d17), and slice 6's principal-uniqueness item (BACKLOG
  #16, a username equal to another member's address, the re-plan `:432`) decides whether to reuse
  the code.
- **No last-admin guard is built.** The spec's "5c builds the last-admin guard, and 6d re-establishes
  it in the write transaction" (spec `:183-184`) presumes a demotion or removal path; none exists
  (`grep -rn -i 'demot\|RemoveMember' internal/auth cmd --include=*.go`, outside tests, finds only
  two comments, `admin.go:299` and `roster.go:754`), so the guard and criterion 6's "concurrent
  self-demotion of the last two admins" route to the first verb that can remove or demote. Reading
  6d-R3.
- `List` returns live rows only, ordered by `created_at` and then `id`, since the file's order is
  gone.
- `Member` gains the id.

### R-6d11: the import record (Ruling 4's no only)

Under Ruling 4's yes, 6d builds no import and no migration. Under its no, the appendix's R-6d11 is
this resolution.

### R-6d12: `DUBPLATE_ROSTER_FILE` retires through `AddSuperseded`

In one commit (Task 5), the key leaves `magicLinkSource.Required` and registers through
`AddSuperseded` from `registerAuthConfig`, its production caller. Under Ruling 4's yes the message is
"superseded: members live in the store; add the first admin with dubplate-server member add ADDRESS
--admin" (the task may reword it; it must name `member add --admin` and no file). Under its no, the
message names `member import --roster PATH`, as revision 2.2 had it. The same commit drops the
template's row and the header's count of required keys, and every test environment's copy of the key,
since from that commit a rendered env carrying the row fails every verb's wall load.
`TestRenderedTemplateDoctor` follows in that commit. `web.go`'s and `doctorprobe.go`'s
`privateFiles.roster` go in Task 4, since no file remains to guard and the store leaf sits under the
state directory, which the same check already refuses.

### R-6d13: `doctor`'s roster row becomes a members row

Read through `store.Inspect`'s facts (a live member count and a live admin count), never by opening
the store for writing. The row prints immediately after the store row, from the report the store
row already read (`storeRow`, `store.go:134-150`), so `doctor` inspects once. It prints
unconditionally, as the store row does, so `TestDoctorCmd_LineCountFollowsRegistry` gains a named
constant for it beside `storeRowLines` (R10's arithmetic). The roster row it replaces printed only on
failure; the members row prints its counts when it passes, which the runbook's doctor-rows section
then states. Each read is guarded by a version constant beside `settingsVersion`, on
`SettingsAbsent`'s precedent: the counts need `users` (version 3). An unguarded query would turn
every older store into `StateRefused` (quoted), failing the store row. Its states are Task 4's table.

### R-6d14: `Principal` stays the username, and the network actor stays `member:<principal>`

The re-plan asks revision 2.3 to re-justify or change `Principal` (`:208-209`), since the proxy header
and Navidrome's auto-creation, the spec's reasons (`:858-866`), are gone. `Principal` stays the
username, on two present reasons. The 0.8 streamer authenticates a music app by its Subsonic username,
which S20 defines as "a roster handle: short, ASCII, unique, stable, distinct from the login email,
and resolved to the durable surrogate user id the per-user tables key on" (S20 `:493-495`); the
re-plan's ruling 3 names it "the name they type into a music app." And `Principal` is the readable
name in log lines and in the network actor, where an opaque 16-byte id would read as noise. The
durable key is `UserID`, which per-user rows key on (hook 5). Reading 6d-R14.

`ops.NetworkActor` has no production caller until rung 14 mounts the projection, so 6d changes
nothing in `internal/ops` beyond Task 4's two test fixtures. The routed note for rung 14 and slice 6:
once any channel can edit a username, a job's actor should key on `UserID`, because a renamed
principal breaks the actor's correlation.

### R-6d15: the operator form for store-opening verbs, and the owner refusal

At 6d, `deploy.sh` runs no store-opening verb: it never runs `check`, and its only `systemd-run` is
the table install. Criterion 14's "runs store-opening verbs through `runuser`" therefore has no
subject until rung 16 adds the `check` halt (the re-plan `:417`); reading 6d-R5 routes it there. The
operator form is the runbook's: `systemd-run --wait --pipe --collect -p User=music -p
EnvironmentFile=/etc/dubplate/server.env` with the serve unit's properties and its grant (install step
8). The store's own owner refusal (`leaf.go:113`) names `runuser -u <name> -- dubplate-server
<command>`, which loads no environment file (quoted) and so sends an operator to a command that then
fails the wall load. 6d makes the member verbs store openers, so Task 8 rewords the hint to name the
owner and say to run the verb as that user with the server's environment file loaded, naming no
`runuser`; the four store tests that assert the hint follow.

### R-6d16: `OpenServe`'s loop and the reconciler

The reconciler: `internal/backup` exports one function whose production caller is `serve`, over the
root the restore-intent record names when it names one and the configured library root otherwise,
matching `restore.go:633`. The report file carries the same body `restore` writes and the same head:
`restorer.report` prefixes the stamp, a `source:` line built from the intent's remote, snapshot, or
file, and the displaced line (`restore.go:629-660`). So `ServeOptions.Reconcile` receives the handle,
the stamp `finishRestore` already holds, and the `store.RestoreIntent` (library root and source); the
signature may change, as it has no production caller yet. The settle loop (6a's open item): each
round needs another process to win a lock conversion by completing a restore swap, which is an
operator action, and each wait is already bounded. The loop re-enters once after every successful
settle by design (`open.go:248-269`), so 6d warns only on a second or later settle within one
`OpenServe` call, naming the reason; a plain create or migration logs no such line. No iteration bound
is added, since a bound would refuse a legitimate operator sequence. The nil logger guard already
exists (`open.go:220-222`, quoted). Both 6a items close.

### R-6d17: no `web/` change

The add-member page already requires a username with the write path's pattern
(`web/src/routes/members/+page.svelte:304-322`), and the roster API's JSON shape does not change. Three
`web/` items become stale and route to lane 1's next `web/` pass as one note, since `web/` is not
6d's territory:

- the roster line's `member.username !== ''` conditionals and their "a blank username leaves no
  separator" comment (`+page.svelte:252-258`), dead once every member carries a username;
- the `duplicate_principal` mapping (`+page.svelte:53`) and its copy (`web/src/lib/copy/deck.ts:150`),
  unreachable once R-6d10 deletes the address fallback, unless slice 6 reuses the code;
- `web/scripts/rendered/run.ts`'s comments naming "the fixture roster" and its file (`:198-200`,
  `:4136`, `:4774`), once Task 4 moves the fixture into the store.

### R-6d18: the runbook

`docs/server-runbook.md` exists (5d-ii Task 7) and the re-plan added its roster re-create step
(`a4a2586`). After 6d, a runbook that tells an operator to `member add` before the store exists is
false, since `member add` refuses an absent store naming "start serve first" (`open.go:655-657`). So
Task 8:

- **Install order.** The step that starts the unit moves before the first admin, and `doctor` runs
  after both. Steps 5, 6, and 8 name the `auth` and `store` leaves and the grant
  `ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/store`, and no step names
  `DUBPLATE_ROSTER_FILE`. The order adds no step (the re-plan's ruling 7). While the unit runs before
  the first admin exists, every sign-in answers as an unknown address; the window lasts one `member
  add`.
- **The 6d upgrade section** replaces "Re-create a roster written before the username rename," whose
  only reader is the box. Its steps are this plan's "Owner touchpoints," in that order, each verb
  through the install step 8 form with the store leaf granted. It states that `member add --admin`
  into an empty `users` table is the bootstrap, and that every other member is re-added afterward
  (the box has one).
- **Rollback past 6d.** The runbook's rollback copies restore the pre-6d binary, env, and units. That
  binary reads a roster file, which the upgrade moved aside, so the section says to re-create it with
  the pre-6d runbook's own section (this runbook at `0cb33ae`, `:216-250`). The store leaf stays on
  disk, unread by the pre-6d binary.
- The doctor-rows section's `roster` line (`:265`) becomes the members row's states, and the staged
  swap's "the roster beside it" (`:773`) becomes "the store beside it."

The remaining owed runbook lines (6b-ii-b's, the backup schedule, enabling the unit) stay rung 16's.

### R-6d19: `deploy.sh` grants the store leaf and still starts nothing

`deploy.sh` never enables, starts, or restarts a unit (header `:27-36`, reload comment `:422-426`), and
6d keeps that. Its only 6d change is the leaf list: `store` in, `roster` out, with the comments that
name the leaves. 6c's Ruling 2 (`deploy.sh` halts on `check`'s nonzero exit before restarting
`serve`) and enabling the unit are rung 16's (the re-plan `:414`, `:417`), and so is the `check`
holder line criterion 14 asked for ("names the holder before restarting `serve`"), since its only
reader is that halt (reading 6d-R10). Revision 2.2's `try-restart` order and its gate presence check
guarded a restart `deploy.sh` does not perform, so neither is built. The roster leaf stays on the box
after the deploy; the upgrade section moves its file aside and leaves the directory for a rollback.

### R-6d20: the dual-key epoch routes to the revoke verb (revision 2.2's Ruling 2)

The dual-key epoch is the revocation check that reads both epoch records, one keyed by the session's
address and one by its resolved `UserID`, and takes the later (spec `:875-888`). `EpochStore.Bump`
has no production caller (`grep -rn '\.Bump(' --include=*.go internal cmd`, outside tests, returns
nothing at `0cb33ae`), so no epoch record has ever been written on any box, and 6d's tests would prove
a revocation path nothing in production reaches. The repo's standing rule (CLAUDE.md, Geoff
2026-09-12) builds a hook with its first consumer and exports nothing ahead of a caller, which decides
the question, so this plan takes it as a resolution. Spec ruling 9's "same id on re-add" means the
address-to-id resolution never changes in production, and S20's slice 6 now joins the revoke verb as
an epoch consumer (app passwords carry the valid-after epoch, S20 `:489-491`; the re-plan `:433`); the
first of the two to write an epoch weighs whether it needs the address key at all. 6d leaves
`epoch.go` untouched. Criteria 7 and 8, the rung 5 review's minor finding 14, and the spec's erratum
row 10 ("bumps to both keys") route to the revoke verb's plan by name, with slice 6 named beside it.
Reading 6d-R7.

### R-6d21: `check` and `doctor` evaluate the seed when no settings table exists

Today an absent store gives `check` no facts, so it evaluates nothing (`check.go:62-68`), and a store
below `0004` evaluates through `evaluateTableAbsent`, which "never reads environ" (quoted). That was
harmless while every 6c setting's env value was also free of any wall check. Task 6 moves the auth
thresholds' bounds out of `config.Load` (R-6d8), so on a fresh install an out-of-bounds or unparseable
auth value would pass `check` and `doctor` and then refuse `serve`'s first start. That is a regression
from today, where `Load` refuses it.

For a store with no settings table, the post-seed value of every key is the env value or the default,
since the table `serve` will seed into is empty. So `check` and `doctor` evaluate the seed rule
against an empty row set for an absent store and for a store below `0004`, which reports the same
refusals `serve` would raise. An in-bounds env value still reports "seeds at serve's next start." A
store `doctor` cannot vouch for stays "not checked." This reverses 6c's "never reads environ" for that
case, erratum 6d-E2 to 6c's evaluation design; it changes no stored row and no `serve` behavior.

### R-6d22: existing usernames meet the write path's rules by construction

The re-plan routes "re-validating existing rows' usernames" to 6d (`:432`; 5c-ii's item, because the
loader admitted any ASCII-printable username while the write path admits the narrower pattern). Under
Ruling 4's yes, every `users` row comes from `AddMember`, which runs `validateWriteCandidate`
(quoted) before its one `Update`; no other production statement inserts into `users`; and the box's
one old row is re-added through `member add` in the upgrade. So no row can exist that skipped the
write path's rules, and no re-validation pass over stored rows is built. Task 4's grep proves the
single insert site. Under Ruling 4's no, the appendix's import runs each entry through
`validateWriteCandidate`, which carries the same guarantee.

### R-6d23: `rosterResetServer`'s baseline is a named set

5c-iv routed the reset server's baseline to 6d Task 4 with "must make the snapshot explicit" (HISTORY
5c-iv, "What a later pass would be wrong to rediscover"): today the baseline is "the file as the first
POST finds it," valid only by script order. On the store, the baseline is the set of addresses the
harness names: the two seeded fixture members (`renderedAdminAddress`, `renderedMemberAddress`) and
the add-member row's address (`renderedNewMemberAddress`). Each POST deletes, in one `Update`, every
`users` row whose address is outside that set, and answers `{"members": n}` with the number of
baseline rows present. The answer no longer depends on when the first POST arrives, and the rows the
capture matrix adds (`capture-<kind>@example.test`) never survive a reset.

---

## Spec deviations and readings that take effect on approval

| # | Spec text | 6d reading | Reason |
| --- | --- | --- | --- |
| 6d-R1 | Revision 2.2's "6d lands before any deploy that enables or routes to the serve unit" | Retired: rung 5 has closed, and enabling the unit is rung 16's | R-6d1 |
| 6d-R2 | "The store moves behind the roster's interface without touching a call site" (`:858-859`) | The constructor is the one substitution; the methods gain `ctx` | R-6d10 |
| 6d-R3 | Criterion 6's last-admin clause; `:183-184` | Routed to the first demote or remove verb | R-6d10 |
| 6d-R4 | Owner ruling 1's import, the import's states (`:826-855`), and criteria 1 to 3 | Not built (Ruling 4's yes); under its no, the appendix's fourteen states | Ruling 4 |
| 6d-R5 | Criterion 14's "runs store-opening verbs through `runuser`"; spec `:252-253` | Routed to rung 16, whose `check` halt is `deploy.sh`'s first store-opening verb; the operator form is `systemd-run` with the env file | R-6d15 |
| 6d-R6 | "then recovery" (`:891-893`) | `OpenServe`'s staging cleanup and restore finish; job recovery with rung 14 | R-6d2 |
| 6d-E1 | 6c's `Setting.Sensitive` ("must not print") | CH's marker; the value is shown | R-6d6 |
| 6d-E2 | 6c's `evaluateTableAbsent` ("never reads environ") | With no settings table, `check` and `doctor` evaluate the seed against an empty row set | R-6d21 |
| 6d-R7 | Criteria 7 and 8 | Routed to the revoke verb, with slice 6 named | R-6d20 |
| 6d-R8 | STATUS "To 6d: store identity" (Ruling 3 yes) | Routed to rung 16 | Ruling 3 |
| 6d-R9 | Criterion 9's "with the env holding A and an in-package admin row holding B ... a limiter built at the next start uses B" | At the next start the env is unset or set to B; env A beside admin row B refuses the start (the seed rule, spec `:779`), and that refusal is its own sub-case | Task 7 |
| 6d-R10 | Criterion 14's "names the holder before restarting `serve` onto an older store whose lock is held" | Routed to rung 16 with 6c's Ruling 2: `deploy.sh` restarts nothing at 6d | R-6d19 |
| 6d-R11 | "Restoring a pre-6d snapshot under a 6d binary leaves the users table empty, and the report says to run `member import`" (`:732-733`) | Not built: no pre-6d store or snapshot exists. The box has no store leaf and never enabled backup (the removal pass's read, 2026-10-06), and any store from before the `0003` edit fails at its ledger check (spec erratum, `:428-444`) | Fold record |
| 6d-R12 | Criterion 1's "an imported admin can grant admin through the CLI" | Ruling 4's no only (appendix) | Appendix |
| 6d-R13 | Criterion 4's "naming `member import`" | Under Ruling 4's yes the message names `member add --admin` | R-6d12 |
| 6d-R14 | Criterion 5's "streamer handle"; criterion 6's "`streamer_username`"; `:860-866`'s reason for `Principal` | The field is `username` (the re-plan's ruling 3); `Principal` stays the username for S20's reason | R-6d14 |

---

## Review focus

The `diff-reviewer` reads each task against its criteria. The conductor dispatches
`web-auth-security-reviewer` (Opus) after segment B (Tasks 3 to 5) and after segment C (Tasks 6 and
7), each briefed with this plan's privileged-fields table and asked for correctness gaps ranked by
consequence, not polish; a fold task may refuse a finding whose fix costs more than the risk it
removes, with a one-line reason. Checklist lines for both:

- `Identity.UserID` has one producer (the resolution point); `Principal` is never the id.
- The admin flag reaches `users` only through the CLI grant; a cookie-channel reactivation writes 0.
- `AddMember` re-reads the cookie actor's live admin flag and enforces at-least-one-admin inside the
  `Update` callback, never through the read pool.
- A tombstoned address re-adds under its old id; a live one refuses; a username collision refuses by
  name, case-insensitively, before the insert.
- No production statement but `AddMember`'s inserts into `users`.
- The seed apply runs after `OpenServe` and before `serve.Run`, and a refusal leaves no listener.
- The absolute cap is read at every verification; the four rate limits are read once, after the seed.
- No auth reader falls back to a default on a read error; the answer is read before `Lookup`, and the
  lifetime and cap before `Nonces.Insert`.
- `check` and `doctor` refuse an auth env value `serve` would refuse, in every store state they can
  vouch for (R-6d21).
- `internal/auth` imports `store` and never `settings`.
- The generation-first read and the settings-read-before-the-write rule (6c HISTORY) are untouched.

---

## Task 0: The dependency sweep and the staleness pre-task

Model: `opus`, fresh context. Depends on approval. **0.8 hooks:** neutral.

**Part 1, the sweep.** Run the `dependency-upgrade` skill over `go.mod`, `web/package.json`,
`tools/venv/requirements.txt`, and the pinned binaries (rclone, golangci-lint, Playwright), starting
from the "after" column of `docs/superpowers/research/2026-10-05-dependency-sweep.md`. Every minor and
patch lands in its own commit with the full gate green; a major is held with its unblock condition and
does not block 6d. A sweep that moves something writes its record beside the 2026-10-05 one; a sweep
with nothing to move is one line in this task's report. The survey's take-now decisions follow the
skill. If slice 2 has merged `getkin/kin-openapi` by then, the sweep covers it like any other module.

**Part 2, the staleness read.** It writes no code and edits no file. It grades a moved pointer by
content: content that survives under a new name or offset is a note naming the new site, and content
that is gone is a blocker.

**Blockers: a "no" stops the launch.**

1. **Is the tree quiet?** Report `git log -1` on master, the lane registry, and `git status` (no warm
   change under any file in "Contested surfaces"). Name any lane merged since `0cb33ae`.
2. **Does every pointer into a changed file survive?** Run `git diff --name-only 0cb33ae..HEAD` and
   intersect it with the files "Code pointers" names. Re-resolve each pointer into a changed file; a
   pointer into an unchanged file stands without a re-read. Report each changed signature, above all
   `newServeShell`'s and `authHandlersConfig`'s.
3. **Is `Bump` still without a production caller, and does no demote or remove path exist?** Report
   both greps (R-6d20's and R-6d10's premises).
4. **Is the highest migration `0005`?** Report `ls internal/engine/store/migrations/`. Under Ruling
   4's no, Task 3 numbers from it.
5. **Did the sweep move a module a quoted behavior comes from?** Re-quote that row from `go doc` or
   the module source.
6. **Is the box as this plan assumes?** Read-only ssh: no `/var/lib/dubplate/store`; the serve unit
   disabled and inactive; `/var/lib/dubplate/roster/roster.json` present; doctor's `store` row
   `absent`. A box that already holds a store is a blocker for Task 8's upgrade section, not for Tasks
   1 to 7.

**Notes, not blockers:**

7. Re-quote every in-repo row of "Behaviors this plan relies on" whose file changed.
8. Report the conductor's recorded outcome of Rulings 1, 3, and 4.

**Deliverable count: 2** (the sweep and its record or line, the staleness report).

---

## Task 1: `serve` opens the store

Model: `sonnet`, effort `high`. Depends on Task 0. **0.8 hooks:** consumes hooks 7 and 9 (the
migration runner and the one writer); lands none.

**Files:**
- Modify: `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/store.go` (the migration-wait
  threshold and the keep read), and their tests
- Modify: `internal/engine/store/open.go` (the reconciler's arguments, the re-entry log line) and
  tests
- Modify: `internal/backup/reconcile.go` or `restore.go` (the exported reconciler) and tests
- Modify: `cmd/dubplate-server/serve_test.go`'s shell builders (`buildServeShellWithLogger`
  `:67-81`, which `rendered_test.go` reaches, and the direct `newServeShell` call at `:596-610`);
  `rendered_test.go` itself does not change here. Every shell-building test's
  `DUBPLATE_STATE_DIR` is a `t.TempDir()` the test's uid owns, which `leaf.create`'s owner check
  needs; a helper that shares a state directory between shells is reported.

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
| the state directory owned by another uid | refused naming the owner (`leaf.create`), nothing created |

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
prune rows fail); take `Keep` from `Outcome.Value` for an in-bounds row (the keep-row-1 case, with
the env unset, prunes to the default instead); take the lock exclusive unconditionally (the
current-store shared row blocks); pass the configured root when the record names one (the reconcile
row fails); drop the handle's close (a test that runs `restore` after the verb returns fails on
`ErrLockBusy`).

**Deliverable count: 4** (the threshold, the open, the reconciler, the re-entry line).

---

## Task 2: `serve` applies the seed before listening

Model: `sonnet`, effort `high`. Depends on Task 1. **0.8 hooks:** completes hook 15's seed half
(6c built it; `serve` now applies it).

**Files:**
- Modify: `internal/settings/apply.go`, `apply_test.go` (the export)
- Modify: `cmd/dubplate-server/serve.go`, `serve_test.go`; re-read `backup.go:262-277` and
  `store.go:286-310`, editing only a string still false

**Outcomes:**
- The apply exported, called once after `OpenServe` and before `serve.Run` (R-6d5), under the named
  30s write wait R-6d5 fixes (`Update` refuses a context with no deadline).
- A refusal ends the start nonzero with the joined refusal, before any listener.

**States:**

| Env and stored rows at start | Report |
| --- | --- |
| `DUBPLATE_JOB_RETENTION=720h`, no row | one row, source `env-seed`, one audit row (actor `system`, channel `env-seed`), generation +1 |
| env equal to the stored `env-seed` row | nothing written, generation unchanged |
| env absent, an `env-seed` row present | the row removed, audited |
| an in-package `admin` row, env different | refused naming both values and the resolving action `outcome` already prints ("unset KEY, or set it to V"); no row changed; no shell |
| an in-package `admin` row outside the binary's bounds, env unset | refused naming the key and the bound (spec `:781`, under 6c-E8's value-in-force reading) |
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
shell); drop the write wait's deadline (every writing row fails with "store write needs a context
deadline").

**Deliverable count: 2** (the export and call, the refusal path).

**Segment A ends here.** Checkpoint: STATUS (ledger, spend, next task).

---

## Task 3: Members in the store, and `Identity.UserID`

Model: `opus`, effort `high` (the identity surface). Depends on Task 2. **0.8 hooks:** **lands hook
5**: the surrogate id minted at every add, the resolution point returning it as `UserID`, and the
username required and unique. Hook 4 and hook 20 weighed: 6d shapes no event table; neutral.

**Files:**
- Modify: `internal/auth/roster.go`, `roster_test.go`, `seam.go`, `seam_test.go`
- Modify: `internal/identity/identity.go`, `identity_test.go` (the `UserID` field, and
  `Principal`'s doc comment at `:190-193`, which still promises the surrogate-id substitution R-6d14
  rules out; `Valid` changes in Task 4)
- Modify (the `ctx` threading only, a mechanical change): the six production callers of `Lookup` and
  `List` outside this task's own files, `internal/auth/handlers.go` (`:209`, `:307`, `:365`),
  `internal/auth/admin.go` (`:206`), `cmd/dubplate-server/member.go` (`:162`), and
  `cmd/dubplate-server/doctor.go` (`:236`, in the roster row Task 4 replaces, which takes its
  caller's context), and the tests that call either method
- Modify: `internal/auth/admin_test.go` (the roster API's duplicate-username case, below)
- Under Ruling 4's no only: the appendix's migration `0006` and its fixture

**Outcomes:**
- The users-backed roster (R-6d10): constructor over a handle; `Lookup`, `List`, `AddMember` with
  `ctx`, threaded through every caller in one commit, so the task is green on its own; the
  file-backed constructor and its file I/O deleted once Task 4 moves the callers (this task may keep
  them only until then, and the diff-reviewer checks Task 4 deletes them).
- `Identity.UserID`: comparable, its zero value meaning unresolved, set only by the seam from
  `Lookup`. `Principal` is the username (R-6d14). **`Valid` does not change here**: `serve` stays on
  the file-backed roster until Task 4, whose rows carry no id, so a zero-id refusal now would fail
  every admin-gated test over `serve` (`RequireAdmin` refuses on `!v.Identity.Valid()`, `admin.go:92`,
  `:121`) between the two tasks. Task 4 makes `Valid` refuse a zero id once `serve` reads `users`.
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
| `AddMember` | another live row's username, differing only in ASCII case | refused with `errDuplicateUsername` |
| `AddMember` | a new address asking for a tombstoned row's username, differing in ASCII case | refused with `errDuplicateUsername`, never a constraint error |
| `AddMember` | no username | refused (`errMissingUsername`) |
| `AddMember` | admin requested on the cookie channel | refused (5b-ii) |
| `AddMember` | admin on the command-line channel | admin row written |
| `AddMember` | a cookie actor with no live row, a non-admin live row, or only a tombstoned admin row | refused with `errActorNotAdmin`; nothing written |
| `AddMember` | a command-line non-admin add into a table with no live admin | refused with `errNoAdmin`, naming `member add --admin` |
| `List` | an empty table | an empty list, no error |
| `List` | two live rows and one tombstoned row | the two live rows, in `created_at`, then `id`, order |

**Acceptance criteria:**
- Spec criterion 5 (6d-R14): a verified identity's `UserID` equals the row's id and its `Principal`
  the row's username. **Fixture state:** a member whose address differs from its username, so a
  `Principal` built from the address fails.
- Spec criterion 6, minus the last-admin clause (6d-R3): each `AddMember` row by its own subtest.
  **Fixture state:** tombstoned rows are written directly by the test (no removal path exists); the
  reactivation asserts the id byte for byte and reuses its own old username, so a fixture that picks
  a new username cannot pass it vacuously.
- The page's half of criterion 6 after the swap: one test through the roster API adds a new address
  asking for an existing username in a different case, and asserts the `duplicate_username` outcome,
  never a 500.
- The gate green.

**Named mutations:** mint a new id on reactivation; keep the stored admin flag on reactivation;
compare usernames case-sensitively; limit the username collision read to live rows (the tombstoned
username row answers a constraint error); count a tombstoned admin as the actor's authority; drop the
at-least-one-admin check; drop the `tombstoned_at IS NULL` filter in `Lookup`; set `Principal` from
the address. Reading the actor's flag outside the `Update` (through the read pool) is a race no
deterministic test catches; the diff-reviewer checks every read `AddMember` decides on sits inside the
callback, by name.

**Deliverable count: 3** (the roster over the store, `UserID`, the write rules). The `ctx` threading
rides the roster deliverable; `Valid`'s zero-id refusal is Task 4's. Under Ruling 4's no, the
migration is a fourth.

---

## Task 4: The roster's callers move to the store

Model: `sonnet`, effort `high`. Depends on Task 3. **0.8 hooks:** consumes hook 5; neutral.

**Files:**
- Modify: `cmd/dubplate-server/serve.go` (the roster over the serve handle; `checkRosterDirWritable`
  and the roster read at `:326-332` go), `member.go` (`add` and `list` over `store.OpenWrite`;
  `checkRosterDirOwner` and `errForeignRosterOwner` go), `doctor.go` (the members row replacing the
  roster row, `selectedRosterPath`, `rosterRowStatus`, and `rosterStatus`), `web.go` and
  `doctorprobe.go` (`privateFiles.roster`, and `rosterFileKey` once its last reader is gone), and
  their tests, `doctor_test.go`'s line-count constant among them
- Modify: `internal/engine/store/inspect.go` (the live member and admin counts in `Facts`, guarded
  at version 3), tests
- Modify: `internal/auth/roster.go` (delete the file-backed constructor, `CreateRoster`, the file
  read, parse, write, and lock, `validateRoster`, `principalCollision`, `Member.principal`'s address
  fallback, `errDuplicatePrincipal`, and every sentinel and type only that code used; the field checks
  `validateWriteCandidate` needs stay, without the `memberRecord` detour), `admin.go`
  (`outcomeDuplicatePrincipal` and its mapping arm), and tests. Under Ruling 4's no, the appendix
  keeps `parseRoster`, `validateRoster`, and `memberRecord` for the import.
- Modify: `internal/identity/identity.go`, `identity_test.go` (`Valid` refuses a zero `UserID`); the
  two test fixtures that then need an id, `internal/ops/ops_test.go` (`validIdentity`) and
  `internal/ops/opshttp/helpers_test.go` (`ann`)
- Modify: `cmd/dubplate-server/rendered_test.go`, the harness's member fixture (B1): the two fixture
  members seeded into the store with usernames after the preparation function creates it and before
  the server starts; `rosterResetServer` resetting the `users` table to R-6d23's named baseline,
  through its own write-mode handle over the fixture state directory (the two-descriptor `flock` row,
  quoted) or the shell's, keeping its HTTP contract so `run.ts` needs no change; and the
  `members/roster-file` subtest becoming a `users` read of the added address. The fixture still sets
  `DUBPLATE_ROSTER_FILE` and still writes the file here, since the key stays required until Task 5.
  **Constraint:** the fixture usernames match none of the usernames `run.ts` types
  (`newmember1`, `newmember1x`, `newmember2`, `inherited`, and every `captureUsername` value),
  compared case-insensitively, or a rendered row's expected outcome changes.
- Modify: `cmd/dubplate-server/templatekeys_test.go` (`TestRenderedTemplateDoctor`): its assertions on
  the `roster` row become the members row's, under the table below (the rendered environment has no
  store, so the row reads "no store yet" with the exit unaffected)

**Outcomes:**
- `serve` builds the roster over its handle. `member add` and `member list` open write mode, under
  the store's shared lock and its owner rule (`leaf.create` and `checkOwner`, quoted in the code
  pointers). `member add` on an absent store refuses with `OpenWrite`'s "start serve first", which is
  now satisfiable.
- The members row (R-6d13), its reads version-guarded in `Facts`.
- `web.go` and `doctorprobe.go` drop the roster path (R-6d12).
- The file-backed roster code deleted whole under Ruling 4's yes, and with it the last roster whose
  rows carry no id, so `Valid` now refuses a zero `UserID`.
- The rendered harness runs on the store (B1), with R-6d23's baseline, and its capture matrix and
  members subtest green with no `web/` change.

**The member verbs' states:**

| Store | `member add` and `member list` |
| --- | --- |
| absent | refused with `OpenWrite`'s "start serve first"; nothing created (asserted by `stat`) |
| older or newer than the binary | refused with `OpenWrite`'s version refusal |
| a restore-intent record | refused naming the record |
| the leaf owned by another uid | refused naming the owner (`checkOwner`), the case `checkRosterDirOwner` covered |
| current | `add` writes one row through `AddMember`; `list` prints live rows in `List`'s order |

**The members row's states:**

| Store | Row |
| --- | --- |
| absent | "no store yet; start serve, then member add --admin", exit unaffected |
| older than `0003` (no `users` table) | "members not recorded before store version 3", exit unaffected; the store row reports the pending migrations |
| at `0003` or above, no live member | fails naming `member add --admin` |
| live members but no live admin (a tombstoned admin present) | fails naming `member add --admin` |
| live members with at least one live admin | passes with both counts |
| refused, busy, newer, a restore intent, or not a store | the store row's own report; the members row reads "not read" |

**Acceptance criteria:**
- A `member add` then `member list` round trip through the verbs against a store Task 1's path
  created. **Fixture state:** two members added in a known order, so `List`'s `created_at`-then-`id`
  order and the count both show.
- Each member-verb state above by its own subtest; the foreign-owner row uses the `geteuid` override
  `leaf.go:57-63` provides.
- The members row renders each state above in a golden. **Fixture state:** the no-admin fixture
  holds a tombstoned admin row, so counting tombstoned rows shows.
- `TestDoctorCmd_LineCountFollowsRegistry` counts the members row as one unconditional line.
- `grep` finds no caller of the file-backed constructor and no `rosterFileKey` reader; `grep -rn
  'INSERT INTO users' --include=*.go internal cmd`, outside tests, finds only `AddMember`'s statement
  (R-6d22).
- `Valid`: a cookie-channel identity with a zero `UserID` is false, by its own subtest; an identity
  the seam resolves over `users` is true.
- The full rendered leg passes (it runs in the gate): the capture matrix's reset answers the named
  baseline's count on every POST, and the members subtest finds the added address in `users`.
  **Fixture state:** the add-member row has added `renderedNewMemberAddress` before the first reset,
  and a capture cell has added a `capture-<kind>@example.test` member. A reset that counts only the
  two seeded rows answers 2 where the correct reset answers 3; a reset that deletes nothing answers
  the right count but leaves the rendered list longer than that count.
- `TestRenderedTemplateDoctor` passes with its members-row assertions.
- The gate green.

**Named mutations:** open inspect mode in `member add` (the write fails read-only); count tombstoned
rows as members (the no-admin golden fails); drop the version guard (the `0002` fixture refuses);
make `Valid` ignore `UserID` (the zero-id subtest passes an unresolved identity); make the reset
server delete nothing (a capture cell's added member survives into the next cell, so the rendered list
no longer matches the count `resetCaptureRoster` returns, `run.ts:4778-4783`); take the baseline from
the table at the first POST (a test that posts before the add-member write answers 2 where the named
set answers 3 after it).

**Deliverable count: 4** (serve and the verbs, the members row, the web guard, the deletion with
`Valid`'s refusal). This is the heaviest task: the rendered harness rework, `Valid`'s refusal, and
the template test's members-row assertions land here (B1, M1, M2).

---

## Task 5: The roster file retires

Model: `sonnet`, effort `high`. Depends on Task 4. **0.8 hooks:** neutral (under Ruling 4's no, the
appendix lands hook 5's import channel).

**Files:**
- Modify: `cmd/dubplate-server/auth.go` (`Required` loses the key; `registerAuthConfig` calls
  `AddSuperseded`)
- Modify: `env/dubplate-server.env.template` (the `DUBPLATE_ROSTER_FILE` row and its comment
  `:66-70`, and the header's "nine required keys" `:63-64`) and `cmd/dubplate-server/
  templatekeys_test.go` (`renderedDoctor`'s roster parameter and env row; the "deleting the roster row"
  case retires, and the test instead asserts a rendered env carrying the row fails the load with the
  supersession message)
- Modify: every test environment that sets the key, which from this commit fails the wall load:
  `renderedEnviron` in `rendered_test.go` (the file write goes too), `auth_test.go`,
  `backup_test.go`, `check_test.go`, `doctor_test.go`, `member_test.go`, and `serve_test.go`. This
  task's own criterion-4 test is the one place that keeps the key.
- Under Ruling 4's no, the appendix's files replace this list's purpose.

**Outcomes:**
- `DUBPLATE_ROSTER_FILE` superseded (R-6d12), with the template's row and every test environment's
  copy dropped in the same commit.
- The required key count drops from 14 to 13 (the re-plan's ruling 7); the task's report records it.

**States:**

| `DUBPLATE_ROSTER_FILE` in the env | Report |
| --- | --- |
| absent | every verb loads |
| set to a path, an absent path, or a malformed value | every wall-loading verb fails the load with the supersession message |
| set to the empty string | loads; `rejectNamespace` treats it as unset (quoted) |

**Acceptance criteria:**
- Spec criterion 4 (6d-R13): with the key set, one table test over the registered command tree runs
  every verb that loads the wall and asserts each fails the load with the message verbatim.
  **Fixture state:** the key carries a real path to an existing file, so a check that only refuses a
  missing file cannot pass.
- The empty-value row by its own subtest.
- Spec criterion 14's template half (M2): the template carries no `DUBPLATE_ROSTER_FILE` row,
  `TestTemplateKeysClassifiedAgainstProductionRegistry` agrees, and `TestRenderedTemplateDoctor`
  passes against the template as this commit leaves it.
- `grep -rn DUBPLATE_ROSTER_FILE --include=*.go cmd internal`, outside tests, finds only the
  supersession registration.
- The gate green.

**Named mutations:** re-add the template's row (`TestRenderedTemplateDoctor` fails naming the
supersession message); leave the key in `Required` beside the supersession (the registry's `claim`
panics at registration); register the supersession with a message naming a file (the criterion-4
test's message comparison fails).

**Deliverable count: 3** (the supersession, the template and its test, the test environments).

**Segment B ends here.** The conductor dispatches `web-auth-security-reviewer` over Tasks 3 to 5,
folds its findings in one fix task before segment C, and checkpoints STATUS.

---

## Task 6: The auth thresholds become settings

Model: `opus`, effort `high` (the config contract). Depends on Task 5. **0.8 hooks:** lands hook
15's `sensitive` marker on three keys under CH's meaning (R-6d6); hook 18 (a description and a
sensitive mark on every key) stays rung 14's.

**Files:**
- Modify: `internal/config/settings.go`, `registry.go`, tests (the closed-set string kind; the
  `Sensitive` doc)
- Modify: `internal/settings/evaluate.go`, `accessor.go`, tests (`parse`'s `KindString` arm; the
  string read; the withholding removed; R-6d21's evaluation with no settings table)
- Modify: `cmd/dubplate-server/auth.go` (ten settings, two correlated rules, `validateMagicLinkSource`
  trimmed, `resolveAuthConfig`'s env parse replaced), `serve.go`, `store.go` (the withholding removed;
  the no-table doctor rows), `check.go` (R-6d21's absent-store evaluation), `doctor_test.go` and
  `check_test.go` goldens, `serve_test.go`'s callers of `resolveAuthConfig` (`:77`, `:603`, `:696`),
  and `rendered_test.go`'s `checkRateBudget` (`:494-500`) and `mintLinkFor` (`:541-545`). In this task
  every test caller stops calling `resolveAuthConfig` and reads the values the fixture's seed puts in
  force another way (the task's report names which)

**Outcomes:**
- The closed-set string kind, `parse`'s arm, and the accessor's string read (R-6d7).
- The ten keys registered through `AddSetting` with the spec table's apply modes, their bounds as
  data (R-6d8), `Sensitive` on both lifetimes and the cap. The correlated rules register after the
  ten settings, since `AddCorrelatedRule` refuses an unregistered key.
- Both correlated rules registered; 6c's test pair retired.
- `Sensitive` re-documented and the value shown (R-6d6).
- `check` and `doctor` evaluate the seed against an empty row set when no settings table exists
  (R-6d21).
- **The interim read.** A key cannot be both a threshold and a setting (`claim` panics, quoted), and
  `resolveAuthConfig` parses thresholds from the env. So this task replaces that parse: the
  preparation function reads all ten values once, after the seed, through the accessor over its
  handle, into `authConfig`. The task is green at that seam; Task 7 then turns the "immediately" and
  "next operation" values into per-use readers and deletes `authConfig`.

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
| `check` | a current store, an auth env value out of bounds | exits nonzero naming the key (6c's Ruling 2, 6c-E3) |
| `check` and `doctor` | an absent store, an auth env value out of bounds | fails naming the key and the bound (R-6d21) |
| `check` and `doctor` | a store below `0004`, an auth env value that does not parse | fails naming the key and the accepted form |
| `check` and `doctor` | an absent store, the env setting the expiry above the default retention | fails naming both keys |
| `check` and `doctor` | an absent store, an in-bounds auth env value | passes; the row says the value seeds at serve's next start |
| `doctor` | a busy or refused store, any env value | "not checked", as today |

**Acceptance criteria:**
- Spec criterion 5, re-run over the real pair (6c-E6): the correlated rows above.
- Spec criterion 10: each re-homed key seeds from the env (the seed table's second row, one subtest
  per key, table-driven), and the `doctor` golden shows the source and the `sensitive` marker on both
  lifetimes and the cap, with their values printed.
- Each R-6d21 row by its own subtest. **Fixture state:** the absent-store rows run with no store
  leaf at all, so `check` has no facts; the below-`0004` row runs against a store at `0003`, so its
  facts mark the settings table absent. A fixture at `0004` or above would reach the existing
  evaluation and prove nothing about R-6d21.
- `backup` with an out-of-bounds auth value leaves every settings row unchanged and exits on its own
  outcome (the spec's "a timer-run backup never fails over an auth threshold").
- The gate green.

**Named mutations:** keep `KindString` refused (the answer cannot register); keep the withholding
(the golden fails); register the retention rule with only one key (the both-keys report fails); drop
`parse`'s string arm (the equal-env string row refuses, so a second start refuses); let `check` pass
an out-of-bounds auth value (the `check` row exits 0); keep `check`'s evaluation behind `r.Facts !=
nil` (the absent-store row exits 0).

**Deliverable count: 4** (the string kind, the ten settings, the correlated rules with R-6d21's
evaluation, the sensitive meaning).

---

## Task 7: Accessor reads at every auth call site

Model: `opus`, effort `high` (live security thresholds). Depends on Task 6. **0.8 hooks:** neutral
(the per-account limiter of S20 slice 6 is its own).

**Files:**
- Modify: `internal/auth/handlers.go`, `seam.go`, `admin.go`, `nonce.go` (`Insert`, `Sweep`, and
  `RunSweep`'s sweep function gain `ctx`), and tests (reader functions in place of values)
- Modify: `cmd/dubplate-server/serve.go` (including `sweepNonces`), `auth.go` (`authConfig` and
  `resolveAuthConfig` deleted; readers built over the serve accessor), and tests (`auth_test.go`'s
  `resolveAuthConfig` table, `:830-890`, retires); the task confirms that no caller of
  `resolveAuthConfig` remains

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

## Task 8: The deployment files and the runbook

Model: `sonnet`, effort `high`; invokes `vps-conventions`. Depends on Task 7 (independent by files of
Tasks 6 and 7, but sequenced to keep one gate lane). **0.8 hooks:** neutral. Slice 8 later adds a
certificate leaf to the same unit and leg; the leg's reverse check below accepts any leaf a unit
grants and the array lists, so slice 8 needs no new leg logic.

**Files:**
- Modify: `systemd/dubplate-server-serve.service` (`ReadWritePaths=` and its derivation comment)
- Modify: `scripts/deploy.sh` (`DUBPLATE_SERVER_LEAVES`, its leaf table comment, the header's leaf
  list `:29-30`, the unidecode comment `:287-298`)
- Modify: `scripts/check.sh` (the state-dir leg's two new checks)
- Modify: `docs/server-runbook.md` (R-6d18)
- Modify: `internal/engine/store/leaf.go` (`checkOwner`'s hint, R-6d15) and the four store tests that
  assert it (`leaf_test.go`, `inspect_test.go`, `open_test.go`, `restore_test.go`)

**Outcomes:**
- The unit grants `/var/lib/dubplate/auth /var/lib/dubplate/store`, both unprefixed, and no longer the
  roster leaf; its comment states the two leaves' reasons.
- `deploy.sh` pre-creates `auth` and `store` (owner `music`, mode 0700, which `leafDirPerm` matches),
  no longer `roster`, and still enables, starts, and restarts nothing (R-6d19).
- The state-dir leg also fails when `DUBPLATE_SERVER_LEAVES` names a leaf no `dubplate-server` unit
  grants, and when the serve unit grants no `store` leaf.
- `checkOwner`'s hint names the owner and the environment file, never `runuser`.
- The runbook per R-6d18: the install order, the 6d upgrade section, the rollback note, and the two
  string fixes.

**The state-dir leg's states:**

| Files | Report |
| --- | --- |
| unit grants `auth store`, array `(auth store)` | passes |
| array still names `roster`, no unit grants it | fails naming the leaf (new reverse check) |
| serve unit grants no `store` leaf | fails naming the unit (new check) |
| the store entry prefixed with `-` | fails naming the entry (the existing prefix check) |
| the unit grants `store`, the array lacks it | fails naming the leaf (the existing check) |
| the unit's store path misspelled (`/var/lib/dubplate/stor`) | fails twice: the existing array check and the new store check |

**Acceptance criteria:**
- Spec criterion 13: `systemd-analyze verify` (the gate's leg) clean on the unit. Each failing leg
  state above is shown by applying its one edit in the working tree, running the state-dir leg (a
  full `check.sh` run if the leg cannot run alone), quoting the leg's refusal in the report, and
  reverting; the committed files pass.
- Spec criterion 14's leaf half: the array carries `store` and not `roster`; `shellcheck` clean.
  Criterion 14's `runuser` and holder clauses are rung 16's (6d-R5, 6d-R10).
- The owner-refusal hint: the four store tests assert the new text, and no store test or source
  string names `runuser`.
- The runbook read by the task's own render-and-read (the page opened and read in full). The read
  checks that the install order starts the unit before the first admin and runs `doctor` after both,
  that the upgrade section runs the deploy, the supervised start and stop, the move aside, and the
  `member add`, `member list`, and `doctor` steps in that order with no step enabling `serve`, and
  that no line names `DUBPLATE_ROSTER_FILE` or grants the roster leaf outside the rollback note.
- The gate green.

**Named mutations:** grant the leaf with a `-` prefix (the leg fails); leave `roster` in the leaves
array (the reverse check fails); drop `store` from the unit's grant (the store check fails); keep
`runuser` in the hint (the four tests fail).

**Deliverable count: 4** (the unit and the leg, `deploy.sh`, the owner hint, the runbook).

**Segment C ends here.** The conductor dispatches `web-auth-security-reviewer` over Tasks 6 and 7 and
folds; then the simplifier round; then the merge ritual.

---

## Task 9: The pass close

Model: `sonnet`. One fold agent drafts, commits the drafts, then folds, with one independent
`diff-reviewer` read of the fold's diff.

**Outcomes:**
- The 6d HISTORY entry: what landed per task, the gate evidence, both security reads' dispositions,
  both budgets, "What a later pass would be wrong to rediscover" (the start order, the sensitive
  meaning, why no last-admin guard, the closed-set kind, the no-table evaluation, the keep extraction,
  the `parse` string arm, the named reset baseline, `Principal`'s reason, and why `deploy.sh` still
  restarts nothing).
- The errata table above applied, dated, in the spec and the 6c HISTORY entry; under Ruling 4's yes,
  an erratum at the spec's owner ruling 1 and its import section.
- Routed items: to the revoke verb and slice 6 (criteria 7 and 8, minor 14, R-6d20); to the first
  demote or remove verb (the last-admin guard); to rung 14 (the network actor on `UserID`; the admin
  write path's resolving channel for a refusing `admin` row); to rung 16 (the store identity under
  Ruling 3's yes; criterion 14's `runuser` and holder clauses with 6c's Ruling 2 and the unit's
  enabling; the remaining runbook lines); to slice 6 (the `duplicate_principal` code and BACKLOG #16);
  to lane 1's next `web/` pass (R-6d17's note).
- `CLAUDE.md`'s gate paragraph describes the state-dir leg as Task 8 leaves it.
- STATUS rewritten present tense, the "To 6d" line removed.

**Deliverable count: 3** (HISTORY, errata, STATUS with `CLAUDE.md`'s line).

---

## The exit gate

`check.sh` green on master after the merge, plus spec criteria 4 (6d-R13), 5 and 6 (6 without its
last-admin clause), 9 to 13, and 14's template and leaf halves, each by the task that names it.
Criteria 1 to 3 are not built under Ruling 4's yes (6d-R4), criteria 7 and 8 are routed (R-6d20), and
criterion 14's `runuser` and holder clauses are rung 16's. Each new export lists its production
caller; the diff-reviewer's records carry every named mutation reported failing.

## Contested surfaces

Every file below is shared with rung 5's closed work or another lane; no other executor edits one
while 6d runs: `cmd/dubplate-server/serve.go`, `auth.go`, `store.go`, `doctor.go`, `doctorprobe.go`,
`check.go`, `member.go`, `web.go`, `serve_test.go`'s shell builders, `templatekeys_test.go`, every
`_test.go` environment setting `DUBPLATE_ROSTER_FILE`, and `rendered_test.go`'s member fixture, reset
server, `checkRateBudget`, and `mintLinkFor`; `internal/auth/*` (rung 12's since rung 5's close, idle
until slice 3); `internal/identity`; the two `internal/ops` test fixtures (`ops_test.go`,
`opshttp/helpers_test.go`); `internal/config/settings.go` and `registry.go`; `internal/settings`;
`internal/engine/store` (`open.go`, `inspect.go`, `leaf.go` and the tests asserting its hint);
`internal/backup` (the reconciler); `systemd/dubplate-server-serve.service`; `scripts/deploy.sh`;
`scripts/check.sh`; the env template; `docs/server-runbook.md`; and `go.mod`, `go.sum`, and
`web/package*.json` during Task 0's sweep. Slice 2's `go.mod` addition is the one expected concurrent
edit; it lands before Task 0 or after 6d's merge, and the merge ritual reconciles it either way.

## The conductor's actions

1. At approval: record Rulings 1, 3, and 4 in this plan and STATUS; record R-6d20's routing in STATUS
   (the revoke verb line names slice 6 beside it); under Ruling 4's no, fold the appendix into Tasks
   3 to 5 before Task 0.
2. Before Task 0: the live-executor check (lane registry, `git status`), the registry row, the runaway
   guard, AC power.
3. After Task 0: stop on any blocker. Re-derive the ceiling only if the sweep or a merged lane moved a
   task's scope.
4. At each segment boundary: STATUS; check the 80% flag (8.4M under the recommended answers).
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
commit. No task runs in parallel. Every task from Task 3 to Task 7 reads the one before, and Task 8
runs last to keep one gate lane.

## Prior art consulted

| Entry | Verdict for 6d |
| --- | --- |
| `modernc.org/sqlite` | Adopted (spec); 6d adds no migration under Ruling 4's yes. `NOCASE` behavior quoted and relied on for ASCII usernames only. |
| `pressly/goose`, `golang-migrate` | Declined (spec); under Ruling 4's no, `0006` joins the hand-rolled runner. |
| OpenSubsonic specification | `Principal` is the Subsonic username a music app sends (S20 `:493-495`), and `UserID` is what 0.8's per-user rows key on (hook 5). No 6d wire surface. |
| Navidrome, gonic (behavior only, GPL) | No 6d surface: dubplate integrates with no external streamer (the re-plan's ruling 4). Revision 2.2's Navidrome reason for `Principal` is gone; R-6d14 replaces it. |
| `golang.org/x/time/rate` | Already adopted by rung 5; 6d only moves when its limits are read (restart mode). |
| `crypto/rand` (via `store.NewID`), `encoding/json/v2` | Adopted (stdlib); the id's mint, and the loader whose unknown-member refusal Ruling 4 cites. |
| `internal/atomicfile` | Not used: every member write goes through the store. |
| systemd `ReadWritePaths=`, `systemd-run`, `runuser` | Quoted; the runbook's `systemd-run` form stays the operator form (R-6d15). |
| taglib, mewkiz/flac, Picard tables, chromaprint, `kin-openapi` | No 6d surface. |

## Not verified in this revision

- R-6d3's 5m default is reasoned from lock semantics, not measured on a populated store.
- Task 4's reset server taking its own write-mode handle beside the serve shell's in one process is
  quoted from `flock(2)` and `tryLock`'s fresh descriptor, not probed; Task 4's first rendered run is
  the confirmation, and the shell's own handle is the fallback.
- The usernames `run.ts` types were read by grep at `0cb33ae`; a `run.ts` change before Task 4 can
  add one, and Task 0's changed-file intersection catches it.
- The box's state is from STATUS and the removal pass's 2026-10-06 read-only check; Task 0 item 6
  re-reads it.
- R-6d21's reversal of `evaluateTableAbsent` is reasoned from the seed rule's empty-table case and
  not yet tested against 6c's existing table-absent tests, which Task 6 updates.
- Under Ruling 4's no only: the trigger-based fault fixture was probed through Python's `sqlite3`,
  not through `modernc.org/sqlite`; the appendix's Task 5 run is the confirmation.

---

## Appendix: what Ruling 4's no builds

Under Ruling 4's no, these blocks are folded into the plan at approval, each where it names. They are
revision 2.2's import design (`5ed193f`), rebased: `streamer_username` reads `username`, citations sit
at `0cb33ae`, and one state is added for the pre-rename key.

**Task 3 gains** `internal/engine/store/migrations/0006_member_imports.sql` (numbered from Task 0 item
4) and its populated fixture under `testdata/fixtures/`, passing the invariant test, with the table
refusing a `NULL` actor; its fourth deliverable.

**Task 4 keeps** `parseRoster`, `validateRoster`, and `memberRecord` for the verb, and the members
row reads the import record's time when present, guarded at the new migration's version: at `0003` to
`0005` the row says "no import record yet" instead of a failed query; the named mutation "drop the
`member_imports` version guard" fails the `0003`-to-`0005` row.

**R-6d11, the import record and the import's states.** Migration `0006` adds `member_imports`: a
`BLOB` `id`, `at`, the file's SHA-256 as a `BLOB`, the member count, and the OS actor, `STRICT`, no
`kind` column (so invariant 3 does not apply), no path column (the file's path is not stored), and no
change to invariant 6's exempt list. "An import record already present" means any row. The match rule
is the spec's: an entry whose address, username, and admin flag equal an existing live row is that
member and changes nothing; the display name is not part of the match, and a differing one is reported
by both `--dry-run` and the real import and left as stored. The check for an existing record runs
inside the `Update` callback, so two concurrent imports serialize on the write lock and exactly one
succeeds.

The import reads the file the way `readRoster` does (`roster.go:300-334`): it opens and stats it,
refuses a mode with `0o022` set naming the file and the mode (an admin-flag channel must not trust a
file another local account can write), and hashes the same bytes it parses. Each entry then meets the
write path's rules (`validateWriteCandidate`, quoted), which is also R-6d22's guarantee. Every report
names an entry by its position and the field, never its address or username, under `validateRoster`'s
rule ("A record's own address or principal never appears in a returned error, only its position,"
`roster.go:535-536`), so "naming both entries" means both positions.

Five states join the spec's nine. **A file listing no members is refused naming the file**: an empty
import would write the record and block every later import while adding no one. **A file writable by
group or others is refused.** **An entry failing a field rule is refused naming its position and the
rule.** **An entry whose address matches a tombstoned row is refused naming its position**: the import
is not a reactivation channel. **A file carrying the pre-rename `streamer_username` key is refused
naming the file and the unknown member** (quoted, `json.RejectUnknownMembers`), with the report
pointing at the runbook's upgrade section; this is the box's own file.

**Privileged fields gain:** `users.id`, `users.admin` (from the file's `admin` field, once, from a
file no other account can write), and `users.username` each gain `member import` as a channel, and the
`member_imports` row is set by `member import` alone. R-6d22's "any `users` row" row names the import's
`Update` as the second insert site.

**Task 5 becomes `member import`.** Files: `cmd/dubplate-server/import.go` (or inside `member.go`,
the implementer's call by size), registered in `newMemberCmd` (`member.go:46-54`), tests, roster
fixtures under `testdata/`; `internal/auth`'s parse-and-validate exported only as far as the verb
calls it; plus Task 5's own supersession files above, with the message naming `member import --roster
PATH`. Outcomes: `member import --roster PATH [--dry-run]`, one `Update` under the shared lock and
under `memberVerbWriteWait`'s deadline, with the import record, the actor from `resolveActor`, and
every read it decides on inside the callback; `--dry-run` reports every refusal at once and writes
nothing; the merge with members added before it. States (entries named by position and field):

| Roster file and store | Report |
| --- | --- |
| the file absent, and the file unreadable (mode 0000) | refused naming the file, each |
| the file writable by group or others | refused naming the file and its mode |
| the file malformed | refused naming the file and the parse error |
| the file carries the pre-rename `streamer_username` key | refused naming the file and the unknown member |
| the file lists no members | refused naming the file |
| an entry failing a field rule (the username pattern, a blank display name, a bad address, a refused rune, an overlong field) | refused naming its position and the rule |
| two entries with one address | refused naming both positions |
| two entries with one username, differing in case | refused naming both positions |
| entries with no `username` | refused naming each position (spec ruling 8) |
| no admin in the file or the table | refused naming the rule |
| an entry colliding with a different existing member on address or username | refused naming the position and the field |
| an entry whose address matches a tombstoned row | refused naming the position |
| an import record present | refused naming its time and actor |
| every entry valid, matching rows merged | imported; one record; each new member a fresh id; a differing display name on a matched row reported and left as stored |

Acceptance criteria: spec criterion 1 (reading 6d-R12): a valid roster of three (one admin) imports
once; the rows carry `admin` 1 on the admin entry and 0 on the other two; the record's hash equals
`sha256.Sum256` of the fixture's bytes ("Sum256 returns the SHA256 checksum of the data," `go doc`),
its count is 3, and its actor is `resolveActor`'s value; a second run refuses; the imported admin's
session then adds a member through the roster API, and an imported non-admin's session meets the
admin gate's 404. Spec criterion 2: each refusal by its own fixture, asserting the report text and
that `users`, the record table, and `user_version` are unchanged, against a store with one existing
member (and one tombstoned row for the tombstone state); the mode-0000 fixture assumes a non-root
euid, as the gate runs, with no skip guard. `--dry-run` over a file carrying three refusals from
different layers reports all three and changes nothing; over a valid file it exits 0 and changes
nothing, the settings generation included. A fault fixture adds a `BEFORE INSERT ON users` trigger
raising on the third insert (revision 2.2's probe: after `ROLLBACK`, both tables held 0 rows): the
import fails, `member_imports` holds no row, `users` none of the file's entries, and the diff-reviewer
checks by name that the record's insert sits in the same callback as the users inserts. Two concurrent
imports: exactly one succeeds. Spec criterion 3: a member added before the import merges when the
entry matches and refuses naming the position and the field when the admin flag differs. Spec
criterion 4 names `member import`. Named mutations: write the record in its own `Update` before the
inserts; let `--dry-run` enter `Update` without rolling back; drop the mode check; validate entries
with the loader's rule alone; drop the case fold on the username duplicate check. Deliverable count
4 (the verb, the state table, the supersession, the merge).

**Ceiling** 11.0M, flag 8.8M ("Token ceiling," variants). **Review focus gains** "the import refuses a
file writable by group or others" and "a migration column's meaning: no path column, no source id
under another name (spec `:481-482`)." **Deviations:** 6d-R4 reads "fourteen states"; 6d-R12 applies;
6d-R13 does not.
