# Rung 6 spec fold record

This record folds the four adversarial reviews of the rung 6 spec draft into revision 2. The
target is `specs/2026-09-23-rung-6-ops-seam-store.md`: revision 1 at `12a6941` (778 lines),
revision 2 in the working tree (1,095 lines, 1.41 times revision 1). The reviews are committed at
`e97d761` and raised 113 findings. Each finding below has one disposition. **Folded** names where
revision 2 carries it. **Owner fork** names the numbered ruling it became. **Refused** gives the
reason. Where reviews duplicate, the finding is folded once and every ID is listed. The long
arguments revision 2 only states are in "Arguments moved out of the spec."

Finding IDs use the review's own numbers with a prefix: **CT** is the contract and criteria review,
**IN** the data integrity review, **CS** the consistency review, and **ME** the mechanics review.

## Counts

| Disposition | Findings |
| --- | --- |
| Folded | 96 |
| Owner fork | 17 |
| Refused outright | 0 |
| Total | 113 |

No finding was refused whole. Nine folded findings carry a refused part, listed in "Partial
refusals."

## The convergent defects

The conductor named ten defects that several reviews reached independently. Each was fixed at its
root rather than patched.

| Defect | IDs | Structural fix in revision 2 |
| --- | --- | --- |
| 1. The roster move was unbuildable | CT1, IN1, CS1, ME2 | Ruling 1: an explicit `member import` verb with a store-recorded import row, and `DUBPLATE_ROSTER_FILE` moved to `AddSuperseded` and out of `Required` in one commit |
| 2. Migration under a shared lock from any verb | IN8, ME1 | Migrations: only `serve` migrates, under the exclusive lock; every write transaction re-reads `user_version` |
| 3. Recovery killed live non-exclusive jobs | CT4, IN13, ME6 | The job model: per-process leases taken before insert; the four-row recovery table |
| 4. The epoch re-key lost revocations | CT26, IN14, ME11 | Identity in the store: no bulk move; a permanent dual-key read that takes the later epoch; forward-only writes under a directory lock |
| 5. Restore's three-file rename was not crash-atomic | IN3, ME21, CT10 | The restore verb: a truncating checkpoint to one file, an intent record, the lock file outside `db/` |
| 6. Durability and connection classes | IN10, ME20, IN11, IN12, ME4, ME5 | Connections: `synchronous=FULL`, `journal_size_limit`, three connection classes, a maintenance connection that is closed after use |
| 7. Env-seed ran in every process | CT18, IN18, ME7, IN17 | The seed rule: only `serve` applies it, in one transaction; other verbs evaluate it read-only |
| 8. Byte paths were not enforced | ME3 | Invariant 1: every table is `STRICT` |
| 9. Invariants the catalog cannot check | CT2, ME17, CS14, CS6 | Schema invariants: each rule restated as a catalog fact; the semantic residue moved to a reviewer checklist |
| 10. Criteria that passed vacuously | CT3, CT5-CT10, CT12-CT14, CT21, CT25, CT27, CT31 | The pass split: numbered criteria per pass, each naming its fixture and failure reason; the builds table maps every deliverable to its criteria |

## Owner rulings: how revision 2's list was built

Revision 1's seven rulings survive, renumbered by build impact, and seven forks from the reviews
join them. No ruling became moot. The draft's ruling 6 was restated because every review showed it
unbuildable; its successor, ruling 1, is still a fork, because it changes the upgrade an adopter
runs.

| Revision 2 | Source | Review IDs |
| --- | --- | --- |
| 1. Roster through an import verb | draft ruling 6, restated | CT1, IN1, CS1, ME2 |
| 2. CH tables at 6c, hook columns with their computing rungs | draft ruling 5, reframed | CS2 |
| 3. Threshold homes | draft ruling 7; busy timeout added | CS8 (folded citation and erratum) |
| 4. Backup contents, with the streamer database via rung 12 | draft ruling 1; the streamer database merged in | CS4 |
| 5. One flag, one remote, `RCLONE_CONFIG_*` credentials | draft ruling 2; the credential channel merged in | CS3, CT15 (credential half) |
| 6. Restore is offline | draft ruling 3 | none challenged it |
| 7. Pre-migration snapshots per from-version | draft ruling 4, amended | IN9 |
| 8. The import refuses a missing `streamer_username` | new | CT11, IN2 |
| 9. A re-added member reactivates the same id | new | IN15 |
| 10. dubplate prunes snapshots and the archive | new | CT16 (pruning half) |
| 11. Only the latest pair restores together | new | IN5 |
| 12. Rotate secrets rather than back up the auth leaf | new | CT16 (auth-leaf half), IN7 |
| 13. A removed seeded key reverts to the default | new | IN19, CS19 |
| 14. `doctor` warns on an unencrypted remote | new | ME24 |

Two findings named a fork that revision 2 made technical. CT1 said the mechanism was technical;
ruling 1 keeps it as a fork because ME2 asked for one and the answer changes the adopter's upgrade.
IN7 offered the auth leaf or secret rotation; both change what an adopter does after a disaster, so
ruling 12 asks it.

## Contract and criteria review (CT)

| ID | Disposition | Where |
| --- | --- | --- |
| CT1 | Owner fork | Ruling 1. The import states and the 6d criteria 1-3 and 11 fold the technical half. With IN1, CS1, ME2. |
| CT2 | Folded | Schema invariants 1-7 restated as catalog facts, plus the reviewer-checklist paragraph. The allowlist of location-like names is a partial refusal. With ME17. |
| CT3 | Folded | The seed handoff's required-behavior column; 6d criterion 9 (env A, admin row B, per apply class); the rule that a store-homed key has no wall getter. With ME13. |
| CT4 | Folded | The job model: leases, the recovery table, the late-update refusal; 6b-i criteria 3-5. With IN13, ME6. |
| CT5 | Folded | "Two opens"; the check criterion moved to 6a criterion 13 (inspect byte-identical) and 6c criterion 15; `check.go` in the contested rows. With IN24, ME18. |
| CT6 | Folded | 6a criterion 3 (two handles, positive and negative), criterion 5 (the extended foreign-key code, and read-only proven separately), criterion 14 (the `statfs` seam and the magic-number set). With ME27. |
| CT7 | Folded | 6a criterion 9; "non-empty" defined as `user_version > 0` in Migrations. |
| CT8 | Folded | Migrations: every migration ships a populated fixture, and the test fails on a missing one; 6a proves migration 1 over an empty store, and 6c the first populated one. |
| CT9 | Folded | The backup job: a real `rclone` with a `local` backend, no skip, and a recording stand-in with a set exit code; 6b-i criteria 10-13. |
| CT10 | Folded | The restore verb: the interface, the flow, and the tree restore; 6b-ii criteria 5-9. The absent `-wal` case is moot, because the checkpoint leaves one file. |
| CT11 | Owner fork | Ruling 8. The import state and 6d criteria 2 and 11. With IN2. |
| CT12 | Folded | The HTTP projection: a nil authorizer fails, a denial has no side effect; privileges come from a closed set; 6b-i criterion 1, 6b-ii criterion 2. |
| CT13 | Folded | Operation availability, both directions; 6b-i criterion 8. |
| CT14 | Folded | The contract suite: one case per sentinel, with named triggers. `internal` is triggered by a store closed under the implementation rather than a test-only constructor option, so no test seam is added. 6b-ii criterion 1. |
| CT15 | Folded | Ruling 5 names `DUBPLATE_BACKUP_ENABLED` and `DUBPLATE_BACKUP_REMOTE`; the remote layout paragraph; supersession messages checked by 6b-i criterion 15. The credential channel is ruling 5 (with CS3). |
| CT16 | Owner fork | Rulings 10 and 12. |
| CT17 | Folded | The seed table: `default` stores no value; rows for out-of-bounds, orphan, and unparseable values; 6c criteria 7-8 and 13. With ME8, and the technical halves of IN19 and CS19. |
| CT18 | Folded | The seed rule: `serve` only, audit actor `system` and channel `env-seed`; 6c criteria 9-10. With IN18, ME7. |
| CT19 | Folded | The job model: a fixed constant at 6b-i, re-homed at 6c; the newest success per kind survives; `settings` leaves `ops/local`'s 6b imports; 6c criterion 14. With ME33. |
| CT20 | Folded | The read pool uses `_query_only=1` (6a criterion 5); the settings generation integer (6c criterion 11); privileged-field rows for `env-seed`, terminal state, and surrogate id. |
| CT21 | Folded | The job model: distinct exit codes, numbers fixed in the plan against existing verbs; 6b-i criteria 16-17. |
| CT22 | Folded | The job model: a `cancel_requested` column, and a queued cancel goes straight to canceled; 6b-i criterion 6, 6b-ii criteria 3-4. With ME6. |
| CT23 | Folded | The seed handoff: the template carries none of the auth thresholds (`env/dubplate-server.env.template:110-117`). |
| CT24 | Folded | The contested rows give 6a the `main.go`, `doctor.go`, and `check.go` entries. The per-pass Files column is a partial refusal. |
| CT25 | Folded | 6b-i criterion 9: the snapshot's content before and after, not only its integrity. |
| CT26 | Folded | Moot by design. The dual-key read has no copy and no re-key run; 6d criteria 7-8 state the outcomes. With IN14, ME11. |
| CT27 | Folded | Golden criteria: 6a criterion 13, 6b-i criterion 18, 6c criterion 13, 6d criterion 10. |
| CT28 | Folded | "Does not build": the first rung that writes a collection creates the default one. |
| CT29 | Folded | The schema: share rows are reserved in the spec only; `visibility` has a CHECK; `writer` has its S20 domain; 6c criterion 4. |
| CT30 | Folded | The hook table, row 4: "the rule lands"; `settings_audit` is its first table. |
| CT31 | Folded | 6a criterion 12 (zero-byte and non-database files); `backup` refuses an absent store (6b-i criterion 14). With ME26. |

## Data integrity review (IN)

| ID | Disposition | Where |
| --- | --- | --- |
| IN1 | Owner fork | Ruling 1 (with CT1, CS1, ME2). The import record, exclusive-lock ownership, and the refusals of an existing record or populated users fold in. The one-release `doctor` warning is a partial refusal. |
| IN2 | Owner fork | Ruling 8 (with CT11). `check` evaluates the import before the deploy (6d criterion 11). |
| IN3 | Folded | The restore verb, steps 3-5, and the layout's lock outside `db/`; 6b-ii criteria 6-7. With ME21. |
| IN4 | Folded | The intent record persists until the library half and reconciliation finish; writing opens refuse on it; 6b-ii criteria 7 and 10. |
| IN5 | Owner fork | Ruling 11. The technical half folds: the snapshot is taken after the sync, the object metadata carries the run id and completion time, and the reconciliation report runs. With ME9. |
| IN6 | Folded | The store-aware count guard and `--max-delete`; ruling 10 sets the archive retention floor; 6b-i criterion 12. The expected-device check is a partial refusal. With ME10. |
| IN7 | Owner fork | Ruling 12. |
| IN8 | Folded | Migrations: `serve` only, exclusive lock, `user_version` read inside the transaction; 6a criteria 7 and 12. With ME1. |
| IN9 | Owner fork | Ruling 7. Folded technical half: from-version names, the staging and rename write, refusal on failure, a free-space check, and the rollback-cost sentence. The `--no-migrate` flag is a partial refusal. |
| IN10 | Folded | `synchronous=FULL` with its reason; 6a criterion 4 reads it back. With ME20. |
| IN11 | Folded | Free-space checks before every full copy, `journal_size_limit`, the truncating checkpoint, the named disk-full error, and `doctor`'s rows. The ballast file is a partial refusal. With CS17. |
| IN12 | Folded | The maintenance connection is a separate `*sql.DB`, closed after use; 6a criterion 10. With ME4. |
| IN13 | Folded | As CT4. |
| IN14 | Folded | Identity in the store, "The epoch"; 6d criteria 7-8. With CT26, ME11. |
| IN15 | Owner fork | Ruling 9. The index semantics fold into the `users` row. |
| IN16 | Folded | The schema: track `rel_path` is relative to the collection root; the `move_intents` contract; the rung 16 relocation note on rename tracking. |
| IN17 | Folded | The seed rule: one transaction, correlated rules checked against the post-seed set; 6c criterion 6. |
| IN18 | Folded | As CT18. |
| IN19 | Owner fork | Ruling 13. The changed-default half folds into the seed table (with CS19). |
| IN20 | Folded | Identity in the store: the rollback runbook line. The `member import --merge` verb is a partial refusal. |
| IN21 | Folded | The write function's contract: rollback on any statement error, reads inside the callback. |
| IN22 | Folded | The backup job states that `WriteFile` does not fsync the directory and `WriteStream` does (`atomicfile.go:82-89`), and that renames of SQLite-written files fsync by hand; staging leftovers are cleaned at open. With ME23. |
| IN23 | Folded | Migrations: the applied-migration ledger of name and SHA-256, and the refusal on mismatch; 6a criterion 8. |
| IN24 | Folded | `check` and `doctor` take the lock shared and non-blocking and open `mode=ro`, and the spec states that they may create `-shm`. With CT5, ME28. |
| IN25 | Folded | Invariant 7: no `ON DELETE CASCADE` to `users` or `collections`. |

## Consistency review (CS)

| ID | Disposition | Where |
| --- | --- | --- |
| CS1 | Owner fork | Ruling 1 (with CT1, IN1, ME2). Errata 5 and 8 cover CH:52-54, R5S:1552-1555, R5S:198-200, and PA:218; 6d criterion 13 covers the template edit and the roster leaf; the template sentence is corrected. |
| CS2 | Owner fork | Ruling 2, reframed as "first writer, with CH's tables the owner's exception"; hook rows 1-3 and 11 relabeled Deferred; erratum 11. |
| CS3 | Owner fork | Ruling 5 names the alternatives and states how `doctor` reports the class. |
| CS4 | Owner fork | Ruling 4: rung 12 adds the streamer database. The salvage citation is corrected to `backup.go:91-104`. |
| CS5 | Folded | Identity in the store: the last-admin guard and the collision check run inside the write transaction; 6d criterion 6. |
| CS6 | Folded | Invariant 3 keys on `kind` and the `_events`, `_audit`, and `_journal` suffixes; `settings_audit` gains `at` and `kind`. The member audit record stays rung 5's log record. |
| CS7 | Folded | The citations are now R5S:194-196, R5S:287-292, PA:209, and R5S:1809-1811 (the epoch and surrogate obligations only); the open interface question became the ruled constraint. |
| CS8 | Folded | Ruling 3 cites CH:66 and CH:74-75; erratum 4. |
| CS9 | Folded | Erratum 1. Reconciling against PA:235 is a partial refusal. |
| CS10 | Folded | Erratum 2, which covers rung 5's same delta. |
| CS11 | Folded | "Does not build" sends the schedule key to rung 16 with its timer; ruling 5 states one remote; errata 3 and 6. |
| CS12 | Folded | "Does not build": the pings arrive with rung 9's monitoring seam. |
| CS13 | Folded | The job model: "polled; the interface leaves room for a local push layer." |
| CS14 | Folded | Invariant 6 is scoped to entity ids with alternate keys exempt; the citation is S20:166-167; hook row 1 drops the invariant 6 clause. |
| CS15 | Folded | The backup job: `VACUUM INTO` meets hook 10's intent, with the backup API recorded as the alternative. |
| CS16 | Folded | "Does not build": the row version is deferred to rung 14 by name. |
| CS17 | Folded | As IN11. |
| CS18 | Folded | "Does not build": no URL-valued store-homed key; the validator is where rung 14 attaches. |
| CS19 | Owner fork | Ruling 13 (with IN19). The `default` half folds into the seed table. |
| CS20 | Folded | The prior-art cell names the `_pragma` exception (`sqlite.go:305-306`); settings use the shorthand keys; 6a criterion 4 reads every one back. |
| CS21 | Folded | Prior art: one line for the list's other entries; Navidrome's schema is consulted at 6c for behavior only. |
| CS22 | Folded | The schema's `settings` row and the registry metadata paragraph; erratum 7. |
| CS23 | Folded | "Does not build": two mounts, and the authorizer takes the channel. |
| CS24 | Folded | "Does not build": rung 16 authors the runbook, and `users.admin` is authoritative with the contact row routed to rung 14. |

## Mechanics review (ME)

| ID | Disposition | Where |
| --- | --- | --- |
| ME1 | Folded | As IN8. The downgrade to shared after migrating is stated. |
| ME2 | Owner fork | Ruling 1, the verb form ME2 proposed. |
| ME3 | Folded | Invariant 1: every table `STRICT`; 6a criterion 11 and 6c criterion 2 bind a Go `string`. |
| ME4 | Folded | As IN12. |
| ME5 | Folded | Connections: three classes; the busy-handler result moved into the verified-behaviors table. |
| ME6 | Folded | As CT4 and CT22. The lease is taken before insert. |
| ME7 | Folded | As CT18; the settings generation integer replaces `data_version`. |
| ME8 | Folded | As CT17. |
| ME9 | Folded | The backup job states the pair is not atomic, and the brief no longer claims a consistent pair; the snapshot follows the sync; reconciliation. The owner half is ruling 11 (IN5). |
| ME10 | Folded | As IN6; the salvage argument's reversal is recorded in the backup job and in prior art. |
| ME11 | Folded | As IN14; the bulk move is dropped as ME11 proposed. |
| ME12 | Folded | Identity in the store: `Identity.UserID`, with `Principal` kept as the streamer handle; erratum 9. |
| ME13 | Folded | The seed handoff: a required-behavior column, and 6d replaces `authConfig`; 6d criterion 9 and the 6d task count. |
| ME14 | Folded | Registry metadata: the fourth item (a value validator) and the correlated-rule list; 6c criterion 5. |
| ME15 | Folded | Layout: the uid check; 6a criterion 14 and 6b-ii criterion 11. |
| ME16 | Folded | 6b is split into 6b-i (ops seam and backup) and 6b-ii (projection and restore). The cut differs from ME16's so each half ships with a production caller: see "Arguments moved out of the spec." |
| ME17 | Folded | As CT2; invariant 2 gains the parent-uniqueness check. |
| ME18 | Folded | 6a lands `doctor`'s store row and `check`'s inspect report as production callers; `check.go` is in the contested rows. |
| ME19 | Folded | Layout: `O_CREATE|O_EXCL` at 0600; 6a criterion 14. |
| ME20 | Folded | As IN10. |
| ME21 | Folded | As IN3. |
| ME22 | Folded | Connections: the shorthand keys (verified in `sqlite.go:307-391`), `journal_size_limit` through `_pragma`; Migrations: the version formatted into the pragma. |
| ME23 | Folded | As IN22. |
| ME24 | Owner fork | Ruling 14. The layout folds into the backup job's remote-layout paragraph. |
| ME25 | Folded | The write function: a deadline is required and a nested call fails fast; 6a criterion 6. |
| ME26 | Folded | As CT31. |
| ME27 | Folded | Layout: refuse NFS, CIFS, SMB2, and 9p, warn on FUSE; the seam's justification is named. |
| ME28 | Folded | Layout: the holder promise narrowed to "held by another process" with `/proc/locks` pids. |
| ME29 | Folded | Vocabulary: the actor is a method parameter; 6b-i criterion 2. |
| ME30 | Folded | Prior art, goose row: the measured cost and the silent-ahead behavior replace the requirements-count argument; the golang-migrate row states its modernc v1.18.1 pairing. |
| ME31 | Folded | Prior art, ncruces row corrected; erratum 12. |
| ME32 | Folded | "Band and ceilings" is re-derived from 5b-ii's rate; the double count is removed. |
| ME33 | Folded | As CT19. |

## Partial refusals

| ID | Part refused | Reason |
| --- | --- | --- |
| CT2 | A test that flags BLOB or TEXT columns named like a location outside an allowlist | A name heuristic proves only the heuristic, which is the defect CT2 itself names. The semantic residue goes to a reviewer checklist line. |
| CT24 | A Files column on every pass row | Plans carry Files per task. The contested-surface table already assigns each shared file to its pass, and a second copy would drift. |
| IN1 | A `doctor` warning for one release when the key is still set and its digest matches; `member` writes refusing until the import row exists | Under ruling 1 the key is superseded, so any set key is already a load error. The warning period and the write gate have nothing left to guard. |
| IN6 | Requiring the library root on the device `doctor` expects | No component records an expected device, and the store-aware count guard catches the wrong-mount case without one. |
| IN9 | A `restore --no-migrate` path | The older binary's own `restore --from-file` of its pre-migration snapshot already rolls back without a flag, and a newer binary must never place a store it would not open. |
| IN11 | A ballast file in the leaf | The free-space checks, the pre-migration snapshots an operator can delete, and `doctor`'s free-space row cover the un-wedge case. A ballast file is a second mechanism for the same thing. |
| IN20 | A `member import --merge` verb | It is a feature for a rollback path that the runbook covers by re-adding members through the verb; nothing asks for it. |
| CS9 | Reconciling the ceiling total against PA:235 | PA:235 sets no number. It says each phase carries a combined ceiling set at launch, so revision 2 defers to PA section 4. |
| CT14 | A test-only constructor option that injects a store fault | Replaced with a production-real trigger, a store closed under the implementation, because the architecture reader grades test-only seams. The finding's goal is folded. |

## Errata the spec owes

Revision 2's section "Errata this spec owes" lists thirteen lines. Each maps to the silent change
the conductor or a review named.

| # | Document and lines | Change | Source |
| --- | --- | --- | --- |
| 1 | BS:656 | Band M to L, five passes | CS9; conductor |
| 2 | BS:934-935 | Rungs 5 and 6 edit deploy files in the repo; the deploy rides the next ruled one | CS10; conductor |
| 3 | BS:218-220 | One remote plus a flag, and rclone env credentials (conditional on ruling 5) | CS11 |
| 4 | CH:66, CH:74-75 | Wall homing of the serve keys, the busy timeout, and the shadow keys (conditional on ruling 3) | CS8; conductor |
| 5 | CH:52-54 | "The roster file" retires from the wall list (conditional on ruling 1) | CS1; conductor |
| 6 | CH:56 | The backup schedule key is rung 16's | CS11 |
| 7 | CH:638 | Apply mode and the sensitive flag are registry metadata (clarifying) | CS22 |
| 8 | R5S:1552-1555, R5S:198-200, PA:218 | The roster key and the roster leaf pre-create retire at 6d (conditional on ruling 1) | CS1 |
| 9 | R5S:1153-1158, `identity.go:115-119` | The surrogate id is `UserID`; `Principal` stays the handle | ME12 |
| 10 | R5S:1687-1690 | The re-key becomes a dual-key read | IN14, ME11 |
| 11 | S20:363-372, S20:392 | Hook columns 1, 2, 3, and 11 arrive with their computing rungs (conditional on ruling 2) | CS2; conductor |
| 12 | S20 prior art | ncruces is wasm2go, not wazero | ME31 |
| 13 | BS:662-664 | A track path is relative to its collection's library-relative root (clarifying) | IN16 |

## Arguments moved out of the spec

**Why only `serve` migrates, with no `migrate` verb.** A `migrate` verb would add a subcommand to
BS's list for no case `serve` does not already cover: after a deploy, the restart migrates. The
risk is the deploy window, when binary N+1 is on disk and serve N still runs. Only-serve plus the
exclusive lock means a timer-run `backup` at N+1 refuses the old store by version. Its failure
ping names the cause. It never migrates under the live serve N.

**Why the lease is per process.** A per-kind lock tells a dead job from a live one only for
exclusive kinds. A per-job lock file would multiply files and locks with no gain, because one
process's death ends all of its jobs together. The lease is taken before insert, so no opener can
see a running record whose owner has not yet locked (ME6's ordering race).

**Why the epoch gets no bulk move.** Record names are one-way digests, so a bulk re-key cannot
reach an address removed before 6d. A copy also races a bump from another process, because the
epoch mutex is in-process only. A permanent read of both keys costs one extra file read per
verification. It survives rollback to a 5d binary, a restored pre-6d snapshot, and re-adding a
member. It deletes the completion marker, the crash-safety machinery, and the four-state proof.
ME11 proposed it, and IN14 reached the same read.

**Why the snapshot follows the tree sync.** Neither order makes the pair atomic without pausing
library mutations for the length of a sync, which on the first run is hours. With the snapshot
second, every file the catalog named when the sync began is on the remote or in its archive, so
the only gap is a change committed during the sync. Restore's reconciliation report names that
gap. With the snapshot first, a move during the sync archives the path the catalog still names.
Rung 9's lease may choose to pause mutations during a sync; the spec records that option as a
constraint on rung 9 rather than ruling it.

**Why three connection classes.** The mechanics review measured each edge. A write reaching a
pool without `_txlock` fails at once, because the busy handler does not run on a deferred upgrade.
`_query_only=1` prevents that, but `VACUUM INTO` then fails on the read pool. Running the snapshot
on the one write connection stalls every in-process writer for the whole snapshot. A separate,
closed-after-use `*sql.DB` serves both `VACUUM INTO` and the foreign-keys-off rebuild. Because it
never re-enters a pool, the pragma leak ME4 measured cannot happen.

**Why `STRICT` over a `typeof` CHECK.** Both refuse a TEXT bind into a BLOB path column (ME3's
run). `STRICT` is one catalog bit per table (`pragma_table_list.strict`), checkable without
parsing SQL. A CHECK is visible only in the schema text, which the invariant test may not parse.

**Why the 6b cut differs from ME16's.** ME16 proposed splitting the seam from backup, with a fake
long-running operation at 6b-i. Exported ops code whose only callers are tests breaks CLAUDE.md's
"nothing exported ahead of a caller," and the exported-unused leg will gate on it. Putting backup
and its verb in 6b-i gives the seam its production caller. The projection stays unexported until
rung 14, so it carries no such cost in 6b-ii beside restore.

**Why `Principal` keeps the streamer handle.** Rung 12 injects `Principal` into the proxy header
(R5S:1243-1245), and Navidrome creates an account for any unknown proxy user. An opaque id there
would create a fresh, unreadable account per member and orphan that member's history. The
`identity.go` comment that anticipated the substitution is corrected by erratum 9.

**How the ceilings were derived.** 5b-ii is the nearest analog: an auth-touching pass with
security review inside the chain and several fix rounds. Its segments A through C spent about 4M
across Tasks 1-8, about 0.5M per task. The draft priced tasks at 8b's chain rate, 0.09M to 0.1M
per agent, with no fix rounds. It also added 5a's security review and fold to 5a's actual, which
already contained them (ME32). Plan preparation stays at the measured 0.83M to 0.87M. The task
counts come from counting one deliverable per task at 5a, 8a, and 8b granularity:

| Pass | Tasks, one per deliverable |
| --- | --- |
| 6a (7) | driver adoption tests; open, leaf, lock, uid, and filesystem; connections and the write function; the runner, ledger, and snapshot; the maintenance connection and rebuild; the invariant test with fixtures; the job table plus the `doctor` and `check` callers |
| 6b-i (8) | ops types, codes, privileges, and the catalog; leases, recovery, and cancel; the `ops/local` runner; the snapshot job; the rclone transport, guards, and layout; the class and supersessions; the `backup` verb with SIGINT and exit codes; `doctor`'s backup row |
| 6b-ii (5) | the projection and authorizer; the contract suite; the restore swap and intent; library restore and reconciliation; the `restore` verb interface |
| 6c (8) | the skeleton and users migrations; the CH routed tables; the registry metadata; the accessor and generation; the seed rule; audit writes; `doctor` and `check` rendering; retention re-home and pruning |
| 6d (7) | the import verb and supersession; the users-backed roster with the guard in a transaction; `Identity.UserID`; the dual-key epoch; re-homing and the `authConfig` rewire; per-class proofs and `sensitive`; deployment files |

No pass exceeds ten tasks. The ceilings total 27.5M against the draft's 17M.

## Unresolved

- The exit-code numbers for the `backup` verb are left to 6b-i's plan, to match the existing
  verbs' convention. The spec fixes the classes and their distinctness, not the integers.
- The bound on `serve`'s wait for the exclusive migration lock depends on systemd's default start
  timeout. The serve unit sets no `TimeoutStartSec`, and this fold did not quote systemd's default.
- The collection `visibility` value and the `writer` domain's exact spelling are CH's and S20's.
  The spec refers to them rather than restating them, because this fold did not re-read those lines.
- `internal/auth/seam.go` was staged but uncommitted in lane 1 during this fold, so rung 5
  citations may move. The 6d staleness pre-task re-reads them.

---

## Revision 2.1

A fresh verification read of revision 2 (`research/2026-09-23-rung-6-spec-fold-verification.md`,
committed at `7c1b46b`) found 1 blocker, 12 majors, and 15 minors. It confirmed that the first fold
was real and that every original blocker was fixed at its root. This section gives each of its 28
findings one disposition. The findings are cited as **V1** through **V28**. Revision 2.1 is 1,259
lines, 1.62 times revision 1. It grows past the earlier 1.4 times target because the verification's
fixes add criteria, quoted mechanisms, and two tables: the import states and the lock procedure.
Where this section contradicts an earlier section of this record, this section governs. That
covers the IN1 partial refusal, "How the ceilings were derived," and the "Unresolved" list.

### Counts

| Disposition | Findings |
| --- | --- |
| Folded | 28 |
| Owner fork | 0 |
| Refused | 0 |

No verification finding became a new owner fork. V2, V3, V5, V9, and V19 changed the evidence or the
wording of rulings 1, 3, 7, 8, 9, and 10. They did not add a ruling.

### Dispositions

| ID | Disposition | Where in revision 2.1 |
| --- | --- | --- |
| V1 | Folded | The catalog skeleton (users, collections, albums, tracks) moves to 6a, so 6b-i's guard and 6b-ii's reconciliation read tables that exist. CH's routed tables stay at 6c. Ruling 2's reason is corrected: the skeleton's first readers are the guard and the reconciliation, and CH's routed tables reference only users and collections. The guard fraction is a 6b-i constant that 6c re-homes. With zero catalog tracks, rclone runs with `--max-delete 0` (quoted from rclone's manual). 6b-i criteria 12-13 and 6a criterion 11 cover it. The task counts are re-derived. |
| V2 | Folded | Identity in the store, "The epoch": every bump after 6d writes both the address key and the `UserID` key. Ruling 9's "no" branch now states why it is safe. 6d criterion 7 adds the re-added-under-a-new-id case. |
| V3 | Folded | The pre-deploy `check` is gone. `member import --dry-run --roster PATH` reports every refusal at once and writes nothing. Because an import refusal leaves `serve` running, the check no longer has to precede the deploy. Ruling 8 and 6d criterion 2 are rewritten against it. |
| V4 | Folded | The privileged-field table's admin-flag row adds "the import verb, from the roster file's `admin` field, once." Ruling 1 and the import table refuse a file and table with no admin. 6d criteria 1 and 6 cover it. |
| V5 | Folded | Ruling 1 now states the real sequence: the deploy, then `--dry-run`, then the import, with the login outage from the restart until the import commits. The import holds the lock shared, like every `member` verb, so `serve` keeps running. The dead end is closed by merging: a matching existing row is the same member, and a differing collision refuses naming both. 6d criterion 3 covers it. This replaces the IN1 partial refusal, whose reason the verification showed unsound. |
| V6 | Folded | "Migrations" quotes `flock(2)`'s conversion paragraph verbatim. It replaces the downgrade with a four-step procedure in which `serve` never holds a connection across a conversion and re-reads the version after each one. 6a criterion 13 grants a second process the exclusive lock during the conversion. |
| V7 | Folded | Every citation of `systemd/dubplate-serve.service` is gone. The contested row names 5d's greenfield unit and records that the incumbent file is frozen (R5S:1708-1713). The timeout sentence is replaced: systemd.service(5) is quoted on `Type=simple` and `TimeoutStartSec=`. The real exposure, a backup holding the lock for hours, is stated. The bound is the new wall threshold `DUBPLATE_STORE_MIGRATION_WAIT`, and `deploy.sh` names the holder before restarting. 6d criteria 11 and 14 cover it. |
| V8 | Folded | `serve.go` wiring is a named 6d deliverable and task, with its own contested row. No pass before 6d makes `serve` open the store, so a 5d deploy from a master carrying 6a through 6c is safe ("does not build," "Process and review"). "Package territory" states that exported surface follows its first caller: the open at 6a (`doctor`, `check`), the write method at 6b-i (`ops/local`), the serve mode and seed rule at 6d (`serve`). |
| V9 | Folded | Rulings 7 and 10 name keys and the rules for local deletion, for the backup-off cap, and for remote pruning. The undefined "detection window" clause is removed. 6b-i gains a pruning task and criteria 15-16. The five constants are re-homed at 6c (criterion 14). |
| V10 | Folded | The inspect mode quotes wal.html section 5 verbatim. Its invariant is now "the database file's bytes and mode never change; `-wal` and `-shm` may appear, owned by the leaf owner." 6a criterion 14 and 6c criterion 15 are rewritten to match. |
| V11 | Folded | Three test seams are declared, each with its justification. The runner's source becomes an `fs.FS` parameter, which is dependency injection. Restore gets a step hook run in a re-executed test binary, so each crash is a real process death. The false "one seam" claim is gone. 6a embeds three migrations, which reconciles criterion 9, and criteria 8 and 10 name their injected migrations. |
| V12 | Folded | Restore step 3 opens the current store on a raw connection that skips the version check and accepts a newer store, which is the rollback path. Step 7 reconciles only when the versions match; otherwise it leaves a `reconcile-pending` intent that serve mode completes. 6b-ii criteria 10-11 cover it. The IN9 partial refusal now holds, because step 3 accepts a newer current store. |
| V13 | Folded | Layout: every writing open checks the intent first, and staging cleanup never removes a file an intent names. 6b-ii criterion 8 covers the crash between steps 4 and 5. |
| V14 | Folded | The maintenance connection's DSN is stated, and the lang_vacuum sync quote names it. 6a criterion 4 includes that connection. |
| V15 | Folded | Recovery removes a dead owner's lease file. The kind lock is `leases/kind-<kind>`. 6b-i criterion 4 covers it. |
| V16 | Folded | Invariant 6 applies to entity tables. It names its exempt list (`settings`, the ledger, the generation row, the group quota total). The inert alternate-key exemption is reworded into the hook table. |
| V17 | Folded | The `visibility` CHECK admits both of CH's values, group and private, and rung 14's API refuses private (CH:205-207, re-read). 6c criterion 4 covers it. |
| V18 | Folded | Erratum 10 widens to R4, R18, R35, R5S:1687-1690, and R5S:1809-1811, and states R35's notification key. Erratum 12 adds goose's rationale. Erratum 14 is new and gives rungs 12 and 16 their carried items. |
| V19 | Folded | Ruling 3 is narrowed to the shadow pair. The rest is stated as CH's rule applied. |
| V20 | Folded | Rulings 10 through 14 each carry a "What it changes" line. Ruling 12's rotation clause is rewritten as the verification proposed. |
| V21 | Folded | 6b-i criteria 17, 19, and 21 cover the retention constant, the supersession messages (now verbatim in the criterion), and `doctor`'s WAL, free-space, and snapshot rows. The job records table is proved through criteria 3-6 and 17. |
| V22 | Folded | Layout: an absent leaf reports absent with no uid check, and `deploy.sh` runs store-opening verbs through `runuser`. The obligation is in the 5d unit's contested row and 6d criterion 14. |
| V23 | Folded | "Package territory" states 6d's one new edge, `internal/auth` to `store`. Auth thresholds reach `internal/auth` as accessor functions injected by `cmd`. |
| V24 | Folded | Ruling 5 derives `<NAME>` from the remote name before the colon, uppercased. It quotes rclone's manual on uppercase naming and the required `_TYPE` variable. `doctor` checks `RCLONE_CONFIG_<NAME>_TYPE` (6b-i criterion 21), and ruling 14 reads the crypt type from it. |
| V25 | Folded | The backup job: 6b-i's close routes the `rclone` prerequisite to CLAUDE.md's Gate section. |
| V26 | Folded | The rate is 0.55M per task, from STATUS's "about 5.6M of 7.5M" after 5b-ii Task 10. The pass table is re-priced. |
| V27 | Folded | 6d's re-homing is split into the bounds move and the `authConfig` rewire. `serve.go` wiring (V8) adds a third task, so 6d has 9 tasks. |
| V28 | Folded | The import states are a table of eight states, seven refusals and one success. 6d criterion 2 proves each refusal. |

### Re-derived passes

| Pass | Tasks, one per deliverable | Ceiling | Flag |
| --- | --- | --- | --- |
| 6a (8) | driver adoption tests; open modes, leaf, lock, uid, filesystem; connections and the write method; the runner, ledger, snapshot, and serve-mode procedure; the maintenance connection and rebuild; the invariant test with fixtures; the job-records and skeleton migrations; the `doctor` and `check` callers | 6.1M | 4.9M |
| 6b-i (9) | ops types, codes, privileges, and the catalog; leases, recovery, and cancel; the `ops/local` runner; the snapshot job; the rclone transport, guards, and layout; local deletion and remote pruning; the class, constants, and supersessions; the `backup` verb with SIGINT and exit codes; `doctor`'s backup rows | 6.6M | 5.3M |
| 6b-ii (5) | the projection and authorizer; the contract suite; the restore swap, intent, and step hook; library restore, reconciliation, and `reconcile-pending`; the `restore` verb interface and cross-version cases | 4.7M | 3.8M |
| 6c (7) | the CH routed tables; the registry metadata; the accessor and generation; the seed rule; audit writes; `doctor` and `check` rendering; re-homing the five constants | 6.1M | 4.9M |
| 6d (9) | the import verb with merge and `--dry-run`; the users-backed roster with the guard in a transaction; `Identity.UserID`; the dual-key epoch with bumps to both keys; the bounds move into the registry; the `authConfig` rewire; `serve` opening the store; per-class proofs and `sensitive`; deployment files | 7.3M | 5.8M |

The total is 30.8M, up from revision 2's 27.5M. The difference comes from the 0.55M rate, 6d's
split task, `serve` wiring, and 6b-i's pruning task. Each rate is 0.85M for plan preparation, plus
0.2M for the staleness pre-task, plus tasks at 0.55M each, plus 0.6M for the close. 6b-ii adds 0.3M
for its targeted review, 6c adds 10% for a config-contract change, and 6d adds 0.7M for security
review inside the chain. No pass exceeds ten tasks.

### Mechanisms quoted this round

Revision 2.1 quotes each mechanism it touches from the local man pages or from sqlite.org. The
quotes are `flock(2)` on conversion, systemd.service(5) on `Type=simple` and `TimeoutStartSec=`,
sqlite.org/wal.html section 5 on read-only opens, and rclone's manual on `--max-delete` and on
naming `RCLONE_CONFIG_<NAME>_*`.

### Unresolved after revision 2.1

- rclone's documentation does not say whether a file moved into `--backup-dir` counts against
  `--max-delete`. 6b-i criterion 13 settles it against the real binary.
- 5d's serve unit has not landed, so its file name and `Type=` are unknown. 6d's staleness pre-task
  reads both, and the migration-wait rule holds whatever `Type=` 5d chooses.
- The `backup` verb's exit-code numbers are still left to 6b-i's plan.
- The length is 1,259 lines against the earlier target of about 1.4 times revision 1. Cutting more
  would drop criteria or quotes that the verification asked for.

---

## Revision 2.2

The prose review of revision 2.1 (`research/2026-09-23-rung-6-spec-prose-review.md`, committed at
`7597052`) raised 34 findings, cited here as **P1** through **P34**. It confirmed that every quote it
checked matches its source. Its defects were omissions from quoted passages that change the design,
one wrong failure cause, one wrong golang-migrate reason, a double-counted prep term in the
ceilings, and fourteen seams between sections. Revision 2.2 is 1,171 lines, down from 1,259 and 1.51
times revision 1. It sits above the 1,100-line target: the prose review's move list and the tighter
rulings removed about 150 lines, but P10 through P23 and P31 added criteria, quotes, and two tables
(restore flags, verb locks). Where this section contradicts an earlier one, this section governs.

### Counts

| Disposition | Findings |
| --- | --- |
| Folded | 34 |
| Owner fork | 0 |
| Refused | 0 |

### Dispositions

| ID | Disposition | Where in revision 2.2 |
| --- | --- | --- |
| P1 | Folded | Connections, the read-pool bullet: the write fails with a read-only error; the busy-handler fact explains why the pool needs `query_only`, with the mechanics review's measurement cited. |
| P2 | Folded | The backup job quotes rclone's full `--backup-dir` sentence, server-side move or copy included (confirmed against the local `rclone(1)`). `doctor` reports a backend without it (6b-i criterion 21), and "Not verified" lists the per-backend feature table. Ruling 10 no longer names `sftp` or `webdav`. |
| P3 | Folded | The backup job's credentials paragraph quotes the character-set sentence (confirmed against `rclone(1)`). `doctor` refuses an invalid remote name (6b-i criterion 21). |
| P4 | Folded | Prior art, golang-migrate row: the reason is the modernc v1.18.1 pairing and its own version table. mattn is no longer named. |
| P5 | Folded | Band and ceilings: the chain rate is derived as (5.6 − 0.85 − 0.2) / 10 ≈ 0.46M. The prep range 0.83M-1.07M across 5a, 8a, 8b, 1b, and 4b is cited, with 1.07M used for 6c. The arithmetic is below. |
| P6 | Folded | Migrations: the `flock(2)` quote now runs to the end of its sentence. |
| P7 | Folded | Migrations: "only" is gone; the text quotes `Type=simple` and states that no start timeout bounds the wait. |
| P8 | Folded | File modes, the `query_only` refusal, the busy-handler behavior, "foreign key mismatch", and the TEXT-in-BLOB storage now cite the mechanics review's measurements. The `VACUUM` transaction refusal is quoted from lang_vacuum.html. `data_version` is quoted from pragma.html. Navidrome's auto-create is quoted from its external-authentication documentation, the quote the mechanics review recorded. The header promise now reads "quoted, backed by a recorded measurement, or listed." |
| P9 | Folded | Connections: the wal.html section 6 quote now includes the `journal_size_limit` clause. |
| P10 | Folded, option (a) | Three open modes: the write mode never creates a store, and an absent store refuses with "start serve first". Only serve mode creates or migrates. 6a criterion 13 and 6d criterion 11 cover it. Option (b) was not strictly safer, because it adds a second exclusive-lock path. |
| P11 | Folded | "Does not build", the last bullet: the production-open statement is rewritten as the review proposed. |
| P12 | Folded | The backup guards: `--max-delete 0` is conditional on the unverified count, with the fallback abort. 6b-i criterion 13 covers both. |
| P13 | Folded | Migrations, the rollback paragraph, and ruling 7's "What it changes": the rollback uses the local copy while it survives, then `restore --older` of a pre-upgrade remote snapshot. |
| P14 | Folded | Job pruning moves to each job start in `ops/local`, after `serve`'s seed. The migration runner's local-snapshot cap takes its value from the caller, as the stored row or the registry default. The seed rule states this ordering, and "Package territory" says the store never reads a setting. |
| P15 | Folded | Credentials paragraph: the immediate-backend quote. `doctor` checks that a `crypt` remote's wrapped remote is defined (6b-i criterion 21). Ruling 5 names both remotes in its body. |
| P16 | Folded | The rulings intro separates ruling 2 from the thirteen adopter-facing rulings. |
| P17 | Folded | The five constants carry **[CH]** where 6c re-homes them. The CH index section is removed per the length list; the marks carry the index. |
| P18 | Folded | The hook list binds invariant 3 to the job-records table first. |
| P19 | Folded | Process and review: 6c joins lane 2 only if 5c and 5d leave `internal/config` untouched. |
| P20 | Folded | The restore verb's flag table includes `--discard-current`; 6b-ii criterion 9 covers it. |
| P21 | Folded | The index criterion moves to 6a (criterion 12), with its table. |
| P22 | Folded | The import table adds "two entries with one handle", making nine states and eight refusals; 6d criterion 2 now says eight. |
| P23 | Folded | "The resolution point": from 6d the page and verb require `streamer_username` and refuse a duplicate; 6d criterion 6 covers it. |
| P24 | Folded | Ruling 1 is reduced to a question, recommendation, reason, and what it changes. The mechanics move to "The import". The window sentence now names the CLI verb. |
| P25 | Folded | Ruling 2 asks only about the hook columns. The skeleton and routed-table timing move to "The schema rung 6 lands" as a technical decision. "The owner's explicit exception" is gone. |
| P26 | Folded | Ruling 5 says rclone's template rows arrive at rung 16 and the runbook sets them by hand until then. |
| P27 | Folded | Ruling 7 uses the review's plain question. |
| P28 | Folded | U3, U4, R18, R35, and #604 are defined at first use; "the reviews' scenarios" is cut. R4 appears only in the errata, as a citation label. |
| P29 | Folded | Ruling 13's alternative "refuses to start `serve`". |
| P30 | Folded | 6b-ii criterion 13 names "the swapped-in database file". |
| P31 | Folded | The test seams are a bulleted list, the verb locks a table, and the opening brief is split into three paragraphs. |
| P32 | Folded | "Identity in the store" uses H3 headings. |
| P33 | Folded | The review narrative is gone from the rulings intro, rulings 1 and 7, the guards paragraph, and the epoch paragraph. |
| P34 | Folded | The seed handoff's contrast frame is rewritten as the review proposed. The invariant 6 alternate-key clause and the hook row's clause are reworded. |

### Sections moved out of the spec

Per the prose review's "Length that does no work" list, these sections now live here.

**Status provenance.** Revision 1 at `12a6941`; four reviews at `e97d761` (113 findings); revision
2 at `80ab7ff`; the verification at `7c1b46b` (28 findings); revision 2.1 at `9155c8b`; the prose
review at `7597052` (34 findings); revision 2.2 in the working tree.

**What depends on the configuration homes spec.** Each **[CH]** mark exists because of CH's routes.
The marks cover rulings 1 and 13, the routed tables and root allocation, the settings home, the
registry metadata, and the seed table, pass 6c and 6d's re-homing half, hooks 8 and 15, the
contested rows for `internal/settings` and the registry files, and the five 6b-i constants re-homed
as store-homed keys. The foreign-key DSN rule does not depend on CH: CH requires it (CH:524-526),
and hook 9 and the schema's references require it independently.

**Behaviors a task relies on (the verification ledger).**

| Behavior | Source |
| --- | --- |
| modernc's shorthand keys, validation, and the `_pragma` exception | `sqlite.go:305-391` at v1.59.0 |
| the busy handler on a deferred upgrade | measured by the mechanics review (`SQLITE_BUSY` in 0s); c3ref/busy_handler.html's worked case is the rollback journal |
| foreign keys not toggling inside a transaction | sqlite.org/pragma.html |
| `user_version` and DDL rolling back; no bound pragma parameter | measured by the mechanics review |
| `VACUUM INTO`'s consistency, sync condition, empty-target rule, and transaction refusal | sqlite.org/lang_vacuum.html |
| `VACUUM INTO` refused on `query_only` | measured by the mechanics review |
| `synchronous=NORMAL` losing a commit | sqlite.org/pragma.html |
| `data_version` per connection | sqlite.org/pragma.html |
| WAL on one host; read-only opens; truncation and `journal_size_limit` | sqlite.org/wal.html sections 1, 5, and 6 |
| the hot-journal rule; replacing an open file | howtocorrupt.html sections 1.2, 1.4, and 2.5 |
| statement-level undo | lang_transaction.html section 3 |
| TEXT stored in a non-`STRICT` `BLOB` column; FK to a non-unique parent | measured by the mechanics review |
| file modes 0644 under umask 0022; `-wal` and `-shm` inheriting the mode | measured by the mechanics review |
| `flock` conversion | `flock(2)` |
| `Type=simple`, `TimeoutStartSec=` | systemd.service(5) |
| `--backup-dir`, `--max-delete`, `RCLONE_CONFIG_<NAME>_*` naming, the character set, the immediate backend | `rclone(1)` v1.74.3 |
| Navidrome's auto-create of an unknown proxy user | navidrome.org external-authentication documentation |

**Every state a check can meet.** Each row points to the table or flow that carries it in the spec.

| Check | Where the states are |
| --- | --- |
| inspect, write, and serve modes | "Three open modes" and the serve-mode procedure in "Migrations" |
| job recovery | the recovery table in "The job model" |
| backup | the guard list and remote-layout paragraph in "The backup job" |
| restore | the flag table and eight-step flow in "The restore verb" |
| seed rule | the eleven-row table in "The seed rule" |
| roster import | the nine-row table in "The import" |
| epoch read | "The epoch" (address key only, `UserID` key only, both with either later, neither) |

**The risk model** (the 6b-ii targeted reviewer's brief).

| Failure | Guarded by |
| --- | --- |
| data loss by restore | the exclusive lock, the one-file swap after a truncating checkpoint, the intent record checked before staging cleanup, a crash injected at each step |
| data loss by backup | the missing-root guard, the store-aware count guard, `--max-delete` (0 or the fallback abort with an empty catalog), ruling 10's retention floor |
| a migration under a live older process, or a restore in a lock conversion | serve-only creation and migration under the exclusive lock, no connection held across a conversion, the version check in every write transaction |
| a bad migration | per-migration transactions, the ledger, per-version snapshots, the newer-store refusal, the foreign-key check, populated fixtures |
| foreign keys silently off; lost durability | the DSN, the maintenance connection that never returns to a pool, the readback proof; `synchronous=FULL` |
| disk full | free-space checks, `journal_size_limit`, the truncating checkpoint, the named error, local snapshot deletion, `doctor`'s rows |
| a live job marked interrupted; settings flap | per-process leases taken before insert; serve-only seeding in one transaction |
| revocation lost at 6d; a root-owned store file | the permanent dual-key read, bumps to both keys, forward-only writes; the uid check |

**Ceiling arithmetic.** The chain rate is 0.46M per task. STATUS records about 5.6M spent after 5b-ii
Task 10, and HISTORY's pass spends include plan preparation and Task 0, so (5.6 − 0.85 − 0.2) / 10 ≈
0.46M. Revision 2.1's 0.55M rate plus a separate prep term counted prep twice. The prep range is
0.83M-1.07M (5a, 8a, 8b, 1b, 4b at HISTORY:1735, 2220, 2627, 2961, 3136); 0.95M is used, and 1.07M
for 6c.

| Pass | Arithmetic | Ceiling | Flag |
| --- | --- | --- | --- |
| 6a (8) | 0.95 + 0.2 + 8 × 0.46 + 0.6 = 5.43 | 5.5M | 4.4M |
| 6b-i (9) | 0.95 + 0.2 + 9 × 0.46 + 0.6 = 5.89 | 5.9M | 4.7M |
| 6b-ii (5) | 0.95 + 0.2 + 5 × 0.46 + 0.6 + 0.3 review = 4.35 | 4.4M | 3.5M |
| 6c (7) | (1.07 + 0.2 + 7 × 0.46 + 0.6) × 1.1 = 5.60 | 5.6M | 4.5M |
| 6d (9) | 0.95 + 0.2 + 9 × 0.46 + 0.6 + 0.7 security = 6.59 | 6.6M | 5.3M |

The total is 28.0M. The task lists are unchanged from revision 2.1, except that the `users` index
criterion moved from 6c to 6a with its table; both passes keep their task counts.

**The systemd rationale.** Under `Type=simple` a unit counts as started at fork, so no start timeout
bounds `serve`'s wait inside its main process. `TimeoutStartSec=` fails a service that "does not
signal start-up completion within the configured time" (systemd.service(5)); it also bounds
`ExecStartPre=` commands and is disabled by default for `Type=oneshot`. The bound is therefore
`DUBPLATE_STORE_MIGRATION_WAIT` inside `serve`. 5d's unit type is re-read at 6d's staleness
pre-task.

**Contested surfaces.** The conductor writes these rows into PA's ownership table (PA:173) before lane 2 launches.


| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/engine/store` and its migration set (new) | rung 6's 6a owns the runner, ledger, and invariant test; every later rung appends migrations with fixtures | a rebase onto another lane's migration renumbers the rebasing lane's files, and merge ritual step 5 runs |
| `internal/ops`, `internal/ops/local`, `internal/backup` (new) | rung 6's 6b-i | rungs 7, 9, 13, 14 add operations; rung 12 adds the streamer database to the backup set; rung 16 schedules backup |
| `internal/ops/opshttp` (new) | rung 6's 6b-ii, unexported | rung 14 exports and mounts it |
| `internal/settings` (new) **[CH]** | rung 6's 6c | every rung that declares a store-homed key; rung 14 |
| `internal/config/registry.go`, `thresholds.go` **[CH]** | rung 1b's contract; 6c adds the four metadata items and the correlated-rule list additively | rung 5's lane does not edit these files while 6c runs; if it must, 6c waits |
| `internal/config/keys.go` | 6b-i removes the "backup targets" reserved class in the commit that declares its replacement, registering the retired names through `AddSuperseded`; 6d supersedes `DUBPLATE_ROSTER_FILE` | every later rung, by the same rule |
| `cmd/dubplate-server/main.go` (`newProductionRegistry`, `AddCommand`) | 6a registers the busy timeout and the migration wait; 6b-i the class, its constants, and `backup`; 6b-ii `restore`; 6d `member import`; rung 5's 5c adds the `member` group in the same block | a textual conflict keeps both lines in landing order, and merge ritual step 5 runs |
| `cmd/dubplate-server/doctor.go`, `check.go` | 6a the store row and the inspect report; 6b-i the backup rows; 6c the settings rendering and read-only seed evaluation | rungs 5, 9, 13 add rows the same way |
| `cmd/dubplate-server/serve.go` | rung 5 through 5d; 6d then adds the store open, recovery, and the seed rule before the listener | rungs 7, 12, 14 |
| `go.mod`, `go.sum` | every lane adds its own requirements | a rebase conflict is resolved by `go mod tidy`, and merge ritual step 5 runs |
| The state directory layout | rung 4; 6a adds the `store` leaf | rung 16 converges |
| `internal/auth` roster and epoch, `cmd/dubplate-server/auth.go`, `internal/identity` | rung 5; 6d after 5d merges, with no concurrent auth-writing lane | rungs 7, 12, 14 |
| 5d's greenfield serve unit (a distinct file from the frozen `systemd/dubplate-serve.service`, R5S:1708-1713), `scripts/deploy.sh`, the state-dir leg in `scripts/check.sh`, `env/dubplate-server.env.template` | 6d after 5d merges; `deploy.sh`'s store-opening verbs run through `runuser -u <owner>` | rung 16 converges |

### Unresolved after revision 2.2

- rclone's documentation does not say whether a move into `--backup-dir` counts against
  `--max-delete`. 6b-i criterion 13 settles it, and the spec carries the fallback abort for the
  other answer.
- The spec does not name which rclone backends support server-side move or copy. `doctor` reports
  it per remote, and the 6b-i plan quotes rclone's feature table.
- 5d's serve unit name and `Type=`, and the `backup` exit-code numbers, are unchanged from revision
  2.1.
- The length is 1,171 lines against a target of about 1,100.
