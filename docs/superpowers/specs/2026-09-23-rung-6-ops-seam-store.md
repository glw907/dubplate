# Rung 6: the ops seam and the store

Status: **revision 1, draft for the owner's ratification**, 2026-09-23. No owner ruling below is
given yet. Tree citations are read at `60dfee5`; the staleness pre-task each pass opens with
re-checks the ones that pass leans on.

Rung 6 builds dubplate-server's library store and its operation vocabulary. The store is one
SQLite file on the pure-Go modernc driver, opened in WAL mode with one writer, migrated forward
on open by a hand-rolled runner over embedded SQL, and guarded by a schema-invariant test that
binds every later rung's migrations to the rules this spec fixes: paths as library-relative
bytes, typed columns instead of an attribute bag, timestamped events, and enforced foreign keys.
The ops seam is one Go interface with JSON-serializable request and response types, a job model
with explicit cancellation, and an operation catalog that carries availability and required
privilege, so the organizer spec's parity rule has a mechanism from the first version. Backup
and restore are the first operations: a consistent snapshot of the store plus the library tree,
sent to any rclone remote behind an integration flag, and an offline restore that refuses to run
under a live server. Rung 6 also lands the schema the ratified configuration homes spec routes
to it, the settings home with its env-seed rule, and, once rung 5 closes, the surrogate user id
that moves the roster into the store. The work was drawn as band M. It is now four passes
(band L), and the section "Band and ceilings" says why.

Written against the ratified build spec (`specs/2026-09-02-greenfield-server-build.md`, BS),
the organizer spec (`specs/2026-09-01-dubplate-organizer-design.md`, OS), the configuration
homes spec (`specs/2026-09-11-configuration-homes-design.md`, revision 3, ratified by Geoff
2026-09-23 at `fd122d9`, CH), the 2.0 streamer spec (`specs/2026-09-07-streamer-2-0-design.md`,
revision 2.2, S20), the parallel-build approach (`specs/2026-09-03-parallel-build-approach.md`,
PA), the rung 5 spec (`specs/2026-09-08-rung-5-identity-auth.md`, R5S), the beets evolution
lessons (`research/2026-09-02-beets-evolution-lessons.md`, EVO), the dependency analysis
(`research/2026-09-03-rung-dependency-analysis.md`, DEP), and the reverse-edge completion
(`research/2026-09-04-reverse-edge-completion.md`, RE).

Every passage that exists because CH routes it here carries the mark **[CH]**. CH is ratified,
so those passages are binding; the mark keeps a later owner reversal of a CH route to one edit
per marked passage. The section "What depends on the configuration homes spec" indexes them.

---

## Owner rulings

Seven questions change what an adopter sees, configures, or is asked for. Each carries a
recommendation and its reason, batched for one sitting. Everything else in this spec is a
technical decision closed in its own section.

**1. What does a backup contain?** Recommendation: the store snapshot and the library tree,
together, in one run. The snapshot alone does not restore a library: the store records where
every file lives, and the files are the custody. The salvage backup already syncs both
(`_salvage/internal/backup/backup.go:98-104`), and the organizer spec's purview names backup
and restore as one server capability (OS:77-81). What it changes for an adopter: remote storage
grows with the library, and the first run uploads the whole tree. Snapshots are dated objects;
superseded library files move to a dated archive path on the remote rather than being deleted,
and pruning that archive is the remote's own lifecycle policy, which the runbook names.

**2. How is a backup target configured?** Recommendation: one integration class, a flag plus
one required remote spec in rclone's own `remote:path` form, with credentials in rclone's own
configuration and never in a `DUBPLATE_*` key. The six reserved names in the "backup targets"
class (`internal/config/keys.go:60-71`) are R2-specific, and five of them fail BS's fourth
universality test ("the design makes sense with no Cloudflare account," BS:192-194). They
become supersessions naming the replacement. `DUBPLATE_BACKUP_STAGING` also retires, because
staging moves inside the store's own leaf. What it changes for an adopter: they configure any
rclone backend the way rclone documents it, and dubplate asks for one name.

**3. Is restore an offline verb?** Recommendation: yes. `dubplate-server restore` refuses to run
while any process holds the store, and it is not an operation of the ops interface, so no
network channel reaches it. Replacing a database file under an open connection is undefined
behavior in SQLite ("unlinking or renaming an open database file results in behavior that is
undefined and probably undesirable," sqlite.org/howtocorrupt.html section 2.5). What it changes
for an adopter: a restore means stopping the server first, and the refusal names the holder.

**4. Does the store keep a snapshot before it migrates?** Recommendation: yes, one, overwritten
by the next migration. A store that migrates forward on open changes the file the moment a
newer binary starts, including the case OS names, an older store restored onto a newer box
(OS:425-433). One local snapshot makes a bad migration recoverable without the remote. What it
changes for an adopter: the state directory holds a second copy of the store after any upgrade
that migrates.

**5. Which schema lands at rung 6?** Recommendation: every table CH routes to rung 6, plus the
catalog skeleton those tables reference, with Go accessors arriving with their first callers and
hook columns arriving with the rungs that compute them. **[CH]** Two of Geoff's rulings meet
here. CH routes a list of tables and columns to "the rung 6 store" (CH:637-645), and its
visibility ruling says "the column lands on the row at rung 6" (CH:195). The 2026-09-12
refinement in CLAUDE.md says a 2.0 hook is built by the rung that first consumes it and nothing
is exported ahead of a caller. The recommendation reads the two together: a table is schema,
not exported surface, so it lands at rung 6; the Go code that reads or writes it is exported
surface, so it lands with its caller. The catalog skeleton (collections, albums, tracks, users)
is forced, because CH's routed tables hold foreign keys into it. Hook columns whose values a
later rung computes (the audio key, the file hash, the folded sort key, the gains, the art hash)
arrive with that rung's migration, as the refinement says, and the schema-invariant test fixes
their shape before they arrive. What it changes: nothing an adopter sees; it decides whether
rungs 7, 9, and 14 find their tables waiting.

**6. What happens to the roster file when members move into the store?** Recommendation: at
pass 6d the store imports the roster file once, and after the import `DUBPLATE_ROSTER_FILE` is
a load error naming the `member` verbs. **[CH]** CH classifies members as created-after-boot
state that the store holds (CH:67-70), and the surrogate id that 2.0 hook 5 requires is minted
at the store's write path (R5S:1153-1158), so one transaction must own both the membership and
the id. CH's wall list also names "the roster file" (CH:52-53); this ruling resolves that
sentence as the bootstrap source it was on the day CH was written. Keeping the file as a live
source would give membership two homes, which CH's rule forbids. What it changes for an adopter:
after the upgrade, adding or listing members is the web page or the `member` verbs, never a
hand edit, and a template still carrying the key fails loudly rather than being silently
ignored.

**7. Where do the thresholds registered before the store live afterward?** Recommendation: the
ten auth thresholds move to the store with their env values as seeds; the three serve keys and
the two shadow keys stay on the wall. **[CH]** CH moves tunables to the store with an env seed
(CH:66-86) and keeps on the wall whatever the process needs before it can start or whatever a
consumer outside the server reads (CH:52-66). The serve prefix, bind addresses, and port are
read before the store opens, and CH names the bind address as a wall key. The shadow floor and
budget are read by the shadow unit, whose grant covers only its own leaf. The section "The seed
handoff" gives every key's row. What it changes for an adopter: after pass 6d, `doctor` shows
each auth threshold with its source. Once rung 14 lets an admin set one, an env value that
disagrees with the admin's fails the load by name, so a rotated value is never silently ignored.

**Not asked: rung 9's concurrency model.** The store's design does not force it. The store
serializes writers at the database, and every write transaction is short and bounded (the
section "Connections and transactions"). A per-album lease and an explicit queue both sit above
that, so the choice stays in rung 9's spec (BS:760-764). The one constraint the store places on
rung 9 is stated there: no write transaction spans a MusicBrainz lookup or a file copy.

---

## What rung 6 builds

| Deliverable | Pass | What it is |
| --- | --- | --- |
| `internal/engine/store` | 6a | Opens the database under its own state-directory leaf, applies the connection settings, holds the store lock, migrates forward on open behind a pre-migration snapshot, refuses a newer store, runs write transactions, and writes consistent snapshots. |
| The migration set | 6a, then 6c | Embedded SQL files applied in order, with the schema-invariant test that fails a later rung's migration that breaks this spec's rules. |
| The schema | 6a, then 6c | The job records at 6a; the catalog skeleton and every table and column CH routes to rung 6 at 6c **[CH]**. |
| `internal/ops` | 6b | One interface, the request and response types, the job model, the error vocabulary with stable codes, and the operation catalog. |
| `internal/ops/local` | 6b | The in-process implementation, with the job runner, cancellation, and job recovery at open. |
| `internal/ops/opshttp` | 6b | The HTTP projection, unexported until rung 14 mounts it, and the contract suite that runs every operation in-process and over HTTP. |
| `internal/backup` | 6b | The rclone transport for the snapshot and the library tree, and the restore staging. |
| The `backup` and `restore` verbs | 6b | The two verbs BS's subcommand list assigns to rung 6 (`cmd/dubplate-server/main.go:13-14`), each in its own file. |
| `internal/settings` **[CH]** | 6c | The settings accessor with version invalidation, the env-seed rule applied at boot, and the audit record for every settings write. |
| Registry metadata **[CH]** | 6c | A home and an apply mode per key, the contract change BS's erratum of 2026-09-12 records (BS:28-40), and the `sensitive` marker CH routes to rung 6 beside them (CH:638). |
| Identity in the store | 6d | The users table filled from the roster, the surrogate id at the roster's resolution point and its write path, the epoch re-key R5S:333-338 hands to rung 6, and the seed handoff of rung 5's thresholds. |
| The store leaf's deployment | 6d | The serve unit's grant and the `deploy.sh` pre-create, because 6d makes serve open the store. |

## What rung 6 does not build

- No HTTP route. The HTTP projection is not mounted in `internal/serve`, because backup is an
  admin operation and rung 14 owns the authorization model that must stand in front of it
  (BS:831-837). Rung 14 exports the projection and mounts it under `/api/v1`.
- No `ops/client`. It is post-launch (BS:668-672). The contract suite is the wire contract's
  regression net until then, and not a parity proof.
- No Go accessor for a table whose first caller is a later rung (ruling 5). Rung 7, rung 9,
  rung 13, and rung 14 add the accessors for collections, the catalog, quotas, the move-intent
  journal, decision records, preferences, overrides, and device credentials.
- No hook column whose value a later rung computes (ruling 5).
- No backup schedule. The timer is a systemd unit, and rung 16 converges the estate's units
  (BS:844-851). CH keeps the schedule on the wall for that reason (CH:58-66).
- No re-authentication gate and no sensitive-change notice. Both are rung 14's (CH:657-660),
  because rung 14 is the first surface that can write a setting. Rung 6 lands only the
  `sensitive` marker CH routes here (CH:638), declared per key and shown by `doctor`.
- No re-homing of the shadow thresholds (ruling 7) and no decision on rung 9's concurrency
  model.
- No record from rung 4's or rung 5's state leaves moves into the store, except the roster at
  6d (ruling 6). R5S's R18 keeps the nonce and epoch records in the state directory
  (R5S:533-539), and PA's state-directory row says the store supersedes neither.
- No last-admin lockout guard. It is routed to rung 5's pass 5c.

---

## Package territory

Every package rung 6 creates is new, so its lanes write no file another lane owns, except the
registration lines and the one `keys.go` deletion the contested-surface rows name.

| Package | Role | Imports it may take |
| --- | --- | --- |
| `internal/engine/store` | the database, its connections, migrations, snapshots | `atomicfile`, the driver |
| `internal/ops` | interface, wire types, job model, catalog, error codes | `identity` only |
| `internal/ops/local` | in-process implementation and job runner | `ops`, `store`, `backup`, `settings` |
| `internal/ops/opshttp` | the HTTP projection and the contract suite | `ops` |
| `internal/backup` | rclone transport, library guards, restore staging | `store`, `atomicfile` |
| `internal/settings` **[CH]** | accessor, env seed, audit | `config`, `store` |

`internal/ops` imports no server package, which is what RE cell (5,6) required
`internal/identity` to exist for (RE:116; `internal/identity/identity.go:1-8`). The engine stays
a pure library with no HTTP knowledge (BS:375). `internal/config` must never import `store` or
`settings`, because every package imports `config`; the dependency runs the other way.

---

## The store

### Open, layout, and the store lock

The store lives under one leaf of `DUBPLATE_STATE_DIR` named `store`, following the precedent
of rung 4's `shadow` leaf and rung 5's `auth` leaf (`internal/engine/shadow/store.go:39-49`,
`internal/auth/store.go:12-40`): one package-level constant, one function joining it, and the
directories created at open. The leaf holds the database file, SQLite's `-wal` and `-shm`
files, the lock file, the pre-migration snapshot, and the snapshot staging directory. The leaf
is 0700 and its files 0600, on R5S's R10 policy, because after pass 6d the store holds member
addresses.

Opening the store takes an advisory lock on the lock file. A process that serves or writes
holds it shared; `restore` takes it exclusive and refuses on contention, naming the reason.
The lock is released when the process exits, so a crash never strands it. WAL works only when
every process is on one host ("WAL does not work over a network filesystem," sqlite.org/wal.html
section 1), so `doctor` reports the state directory's filesystem type and the store refuses to
open on a network filesystem.

Which verbs may create an absent store is part of the contract. The verbs that write (`serve`,
`backup`, and after 6d the `member` verbs) create it. `check` and `doctor` never create it; they
report "absent." `restore` creates it only from a snapshot.

### Connections and transactions

2.0 hook 9 asks for one writer, WAL with a `busy_timeout`, and a seam a later projector can join
so its delta commits with its event (S20:387-389). The store meets it with two pools over one
file. The write pool holds at most one connection, and every write transaction begins
`IMMEDIATE`. The read pool serves concurrent readers.

The two choices close known failure modes. A deferred transaction that reads and then writes
cannot upgrade when another connection has written: "the write statement will fail with
SQLITE_BUSY" (sqlite.org/lang_transaction.html), and the plan quotes whether the busy handler
runs in that case rather than assuming it. Beginning `IMMEDIATE` takes the write lock first, so
the busy timeout governs the wait. In WAL mode `IMMEDIATE` still lets readers proceed ("EXCLUSIVE
and IMMEDIATE are the same in WAL mode," same page).

Every per-connection setting rides the DSN, never a one-time statement after open, because
`database/sql` opens new connections whenever it needs them. modernc v1.59.0 accepts
`_txlock=immediate` (`sqlite.go:385-391` in the module source) and applies `_pragma` values
with `busy_timeout` first and the rest in lexicographic order (`sqlite.go:425-447`); the plan
quotes both from the pinned version. The settings are WAL journal mode, foreign keys on, an
explicit `synchronous` level, and the busy timeout. `synchronous` is explicit because a
snapshot written by `VACUUM INTO` is synced to disk only "if the PRAGMA synchronous setting of
the original database is NORMAL or FULL" (sqlite.org/lang_vacuum.html).

The store exposes one way to write: a function that runs a callback inside one `IMMEDIATE`
transaction and commits or rolls back as a unit. That callback is the seam hook 9 names. A
catalog event and any projection of it commit inside the same call. The busy timeout is a
registered threshold with a researched default, homed on the wall, because the store needs it
before it can read any setting.

### Migrations

2.0 hook 7 asks that the mechanism be chosen with a schema that evolves across versions in view
(S20:383-384). The runner is hand-rolled over `embed.FS` and `PRAGMA user_version`, as S20's
prior-art row already records (S20:435); the section "Prior art" re-scores the alternatives.

The runner's contract:

- Migrations are numbered SQL files applied in order, each in its own transaction, with
  `user_version` set inside that transaction. A migration that fails leaves the store at the
  prior version, and a test proves it by failing a migration after its version write.
- A store whose `user_version` exceeds the highest migration the binary carries is refused with
  an error naming both versions. This is CH's "laptop ahead" direction (CH:165-168) on the
  store's own terms, and it covers a newer snapshot restored under an older binary.
- Before applying any migration to a non-empty store, the runner writes the pre-migration
  snapshot (ruling 4) and fsyncs it.
- A migration that rebuilds a table runs with foreign keys off, which must happen outside the
  transaction: foreign-key enforcement "may only be enabled or disabled when there is no pending
  BEGIN or SAVEPOINT" (sqlite.org/pragma.html). The runner holds one dedicated connection for
  such a migration, runs `PRAGMA foreign_key_check` before commit, and refuses to commit on any
  row it returns.
- Migration files are append-only once merged. A lane that rebases onto another lane's new
  migration renumbers its own; the contested-surface row makes that mechanical.

### Schema invariants

One test in the gate applies every migration to an empty database and to a populated fixture
at each earlier version, then inspects the result through SQLite's own catalog (`pragma_table_info`,
`pragma_foreign_key_list`), not by parsing SQL. It fails a migration set that breaks any of
these rules, and a deliberately violating fixture migration proves each rule can fail:

1. A column that stores a filesystem path is named with a `path` suffix and declared `BLOB`
   (U3, below).
2. Every foreign key the schema declares resolves to a table and column that exist.
3. A table whose name ends in `_events` carries a non-null timestamp and a non-null kind, which
   is 2.0 hook 4's rule (S20:374-375).
4. A column holding a source identifier is namespaced by its source, so a MusicBrainz id is
   never a generic `release_id` (EVO:575-578, the #604 class).
5. No table is an attribute bag, meaning a table whose rows pair a free-form name with a value
   for another table's entity (U4, below).
6. Every public identifier column that a wire type may carry is an opaque random value, never a
   rowid or a sequence (S20:163-165; CH:357-359).

### The schema rung 6 lands

Pass 6a lands one table, the job records. Pass 6c lands the rest **[CH]**. The table names are working names the 6c plan may change; the contents are the contract.

- `users`: the surrogate id 2.0 hook 5 names (S20:376-378), the normalized address as an
  indexed alternate key, the display name, the streamer principal, the admin flag, a tombstone
  time (CH:519-521), and timestamps. Pass 6c creates it empty; pass 6d fills it.
- `collections`: name, a server-allocated `root_path`, `owner` as a non-null reference to
  users, `visibility` with 1.0's single value, `writer`, the writer epoch, the mirror
  generation, the last-mirror time, the duplicate-collapse flag, the layout template, the
  tagging policy, and the admission overrides (CH:173-181, CH:637-642). Share rows are
  reserved by name beside visibility and not created, because CH files them as a schema hook
  designed together with sharing if sharing ever ships (CH:535-538).
- `albums` and `tracks`: the skeleton the routed tables reference, with opaque ids, the
  collection reference, the track's library-relative `path` under U3, the MusicBrainz release
  and recording ids under invariant 4, the album-and-track relation, and timestamps.
- `settings` and `settings_audit`: one row per store-homed key with its value, source, seed
  value, actor, and time (CH:79-87), and one audit row per write naming actor, channel, key, and
  old and new values (CH:130-132).
- `member_overrides` and `member_preferences`: per-member rows keyed to users (CH:122-125,
  CH:161).
- `device_credentials`: the credential hash, the user, the device id, creation, last-seen
  address, expiry, and revocation (CH:543-548).
- `quota_counters`: the per-member counter and the group total (CH:488-500).
- `move_intents`: the journal CH's move verb commits with each catalog change (CH:451-458).

Root allocation must support rung 16's relocation (the 2026-09-23 errata, BS:131-141 and
CH:662-664). A migrated collection gets a root allocated from its id like any other, so the
schema carries no column for a grandfathered or legacy root. If rung 12's Navidrome check fails
and Geoff rules for grandfathered roots through rung 10's plan, that ruling adds its column by
migration then. The allocation function and the no-nesting check are Go code with their first
caller at rung 7 or rung 9, and they land there (ruling 5); rung 6 fixes only that `root_path` is
unique and never client-supplied.

### Paths (U3)

The store keeps every path library-relative and byte-preserving from the first schema (BS:662-664,
EVO:356-366). A path column is `BLOB` and crosses the store boundary as `[]byte`, never as a Go
string bound as `TEXT`, so no driver conversion and no SQLite text function can touch it. Paths
compare bytewise. The store refuses a path that is absolute, that has an empty, `.`, or `..`
component, or that carries NUL. The modernc build compiles with `SQLITE_LIKE_DOESNT_MATCH_BLOBS`
(visible in the generated file's header, `lib/sqlite_linux_amd64.go`), so `LIKE` never matches a
path; rung 9's query verb searches paths another way and its plan says how.

Where non-UTF-8 filenames enter is rung 7's gate, which states whether they are admitted,
normalized, or rejected (EVO:363-366). Symlinks are canonicalized once at that boundary
(EVO:600-602). The store's rule is narrower and holds regardless: whatever bytes the gate admits,
the store returns unchanged. Rung 10's migrator relies on the same rule for the estate's existing
bytes.

### Per-track metadata (U4)

The mechanism is typed columns added by migration, one per field a built-in or the engine owns,
with no attribute bag (invariant 5). Adding a built-in is a migration and a typed accessor in the
built-in's own rung. EVO's reading of beets' own maintainers weighs the choice this way ("the
weak-typing complaint corroborates U4," EVO:489-494), and the forward-on-open runner is what makes
it cheap. The cost is stated: an adopter cannot store an arbitrary field without a release, which
is the no-plugin-ABI fork OS already settled (OS:338-352).

---

## The ops seam

### Vocabulary, job model, and cancellation

`internal/ops` defines one interface. Its request and response types are JSON-serializable and
are the source of the API's wire types (OS:230-236). Every request carries the acting identity: a
verified `identity.Identity` for a network caller, or a local operator named by the operating
system user for a CLI verb, on rung 5's audit convention for the `member` verb (R5S:1141-1143).
The implementation sets the actor from its caller, never from a request body field.

Long-running operations return a job handle (OS:241-248). A job's states are queued, running,
succeeded, failed, canceled, and interrupted. A terminal state never changes. Progress is polled,
never pushed, on the uploader's settled ruling about event-stream buffering. Cancellation is an
explicit operation, never a reliance on context cancellation reaching across a process. Canceling
a terminal job is a no-op that reports the terminal state. Canceling a running job cancels its
context, and a job that honors cancellation ends canceled; one that finishes first ends in its own
terminal state, and the response says which.

Job handles are opaque random values. A job runs in the process that started it: the CLI
`backup` verb starts a backup through `ops/local`, polls it, cancels it on SIGINT or SIGTERM
through the explicit operation, and exits with a status derived from the terminal state.

Job records persist in the store. On open, and before starting an exclusive job, the runner marks
interrupted every running record whose exclusive lock no live process holds. An exclusive kind
(backup is one) holds an advisory lock for the job's life, so a second start returns a typed
conflict naming the running job rather than a second concurrent run. A client that polls a job
after the process that ran it died receives "interrupted," never "unknown." Terminal records are
retained for a store-homed threshold's duration **[CH]**.

The error vocabulary is a small set of sentinels, each with a stable code on the wire:
unavailable (with a reason), not found, conflict, invalid request, canceled, and internal. A
caller matches with `errors.Is` in Go and by code over HTTP.

### Operation availability

The catalog lists every operation with its availability, the reason when it is unavailable, and
its required privilege. It is the ops seam's mechanism against ad-hoc parity opt-outs, and it
never shares a term with the streamer seam's driver capabilities (OS:273-277). Every asymmetry it
reports is named in a spec (OS:279-285). This spec names two:

- **Backup is config-gated.** It is unavailable, with the reason "backup integration disabled,"
  when the backup flag is off. This is the kind OS already names for the contributor resolution
  verbs.
- **Restore is not an operation.** Neither implementation offers it, so it is not an asymmetry;
  it is an offline verb (ruling 3).

`doctor` is the catalog's first production reader: its backup row reports the operation's
availability and the age of the last successful backup job.

### The HTTP projection and the contract suite

The HTTP projection maps each operation to a handler that marshals the same types a future
`ops/client` unmarshals. It stays unexported at rung 6 and requires an authorizer with no default,
which it consults with each operation's declared privilege before dispatch. Rung 14 exports it,
supplies the real authorizer, and mounts it; until then nothing mounts it, so no network channel
reaches an admin operation.

The contract suite runs each operation's cases twice: against `ops/local` directly, and against
the projection through `httptest` and an in-test client. It asserts that every catalog entry
declares a privilege, that every error code survives the round trip, and that a job started over
HTTP can be polled and canceled over HTTP. Its scope is stated as BS states it: the wire
contract's regression net, and not a two-implementation parity proof (BS:668-672).

---

## Backup and restore

A backup job writes a consistent snapshot of the store, then sends the snapshot and the library
tree to the configured rclone remote. 2.0 hook 10 is why the snapshot is never a byte copy of a
live WAL database (S20:390-391): "a backup copy then might contain some old and some new content,
and thus be corrupt" (sqlite.org/howtocorrupt.html section 1.2). The snapshot is `VACUUM INTO` a
temporary file in the store leaf's staging directory, which yields "a consistent snapshot of the
original database" (sqlite.org/lang_vacuum.html), followed by an integrity check of the result
and an atomic rename. The driver's online backup API (`conn.go:1208-1219`, reached through the
raw driver connection) is the named alternative; the plan picks one and quotes its guarantee.

The library tree goes by `rclone sync` with an archive directory, adopting the salvage guards
under fresh review: a missing library root and an empty one each abort before rclone runs,
because sync treats an empty source as the new truth and deletes every object on the remote
(`_salvage/internal/backup/backup.go:22-34`). Every rclone invocation runs under
`exec.CommandContext`. The plan quotes the `os/exec` documentation on what cancellation sends to
the process and how `WaitDelay` bounds the wait, and quotes rclone's documentation on sync, the
archive flag, and exit codes.

The restore verb is offline (ruling 3). It takes the store lock exclusive, fetches or reads the
named snapshot, checks its integrity, and refuses a snapshot whose `user_version` exceeds the
binary's. It then moves the current database together with its `-wal` and `-shm` files aside as
one pre-restore set, renames the verified snapshot into place, and fsyncs the directory. The three
files move together because "overwriting a database file with another without also deleting any
hot journal associated with the original database" is on SQLite's list of actions likely to
corrupt (sqlite.org/howtocorrupt.html section 1.4). The first open after a restore migrates the
restored store forward like any other open. Restoring the library tree refuses a non-empty
library root unless the operator names it explicitly.

Backup does not preserve hardlinks, so a restored library is larger than the box it came from
(CH:410-414). The runbook states this, and the relink pass CH describes stays optional.

---

## The settings home and the seed handoff

### The accessor and the env-seed rule **[CH]**

Every store-homed key reads through the settings accessor, never from a long-lived struct, with
a version that invalidates cached values after a write (CH:88-92). The registry gains three
pieces of metadata per key: its home (wall or store) and its apply mode (immediately, next
operation, or restart), the contract change BS's erratum records (BS:28-40), and the `sensitive`
marker CH routes to rung 6 (CH:128-129, CH:638). Rung 14 is the marker's enforcing reader;
`doctor`, which shows it, is the reader it has at rung 6. A key registered without a home
stays on the wall, so every key registered today keeps its behavior until a pass re-homes it.
`doctor` shows each store-homed key with its value, source, apply mode, and sensitive marker.

The env-seed rule runs once at boot, after the wall loads and the store opens, and before any
consumer reads a setting. `check` validates the env as it does today and, when the store exists,
evaluates the rule read-only. Its outcomes are a closed table:

| Row source | Env value | Outcome |
| --- | --- | --- |
| no row | absent | the row is written at its default, source `default` |
| no row, or `default` | present | the env value seeds the row, source `env-seed` |
| `default` | absent | nothing changes |
| `env-seed` | equal to the seed | nothing changes |
| `env-seed` | different, or absent | a present value re-seeds the row and the change is logged; an absent one leaves the row and `doctor` notes the seed's origin is gone |
| `admin` | equal | nothing changes |
| `admin` | different | the load fails, naming both values and the action that resolves it |
| `admin` | absent | the admin value stands |

Every write validates the value against the key's bounds and against every correlated rule that
names the key, and a seed that fails validation fails the load. No channel sets source `admin` at
rung 6: the write path that does is unexported until rung 14's admin API calls it, and the
conflict row is proved in-package by a test that writes an admin row directly.

Pass 6c's own tunables are the first store-homed consumers: job-record retention, and any
backup threshold the plan finds. Each ships a researched default recorded where the key is
defined (BS:236-248).

### The seed handoff

Fifteen thresholds are registered through the env today: rung 4's two shadow keys, pass 5a's
three serve keys, and rung 5's ten auth keys, six of which landed at pass 5b-ii Task 3
(`cmd/dubplate-server/auth.go:99-200`, commits `323a817` and `e0a3aa8`). The handoff runs in pass 6d, after rung 5 closes, because it edits rung 5's files.

| Key | Registered | Home after 6d | Apply mode | Rule for state already issued |
| --- | --- | --- | --- | --- |
| `DUBPLATE_MAGIC_LINK_EXPIRY` | `auth.go:99` | store | next operation | a link carries the expiry it was minted with (`MintLink`, `internal/auth/token.go:118`) |
| `DUBPLATE_SESSION_LIFETIME` | `auth.go:105` | store | next operation | applies at each session's next refresh (`RefreshSession`, `session.go:132`) |
| `DUBPLATE_SESSION_ABSOLUTE_CAP` | `auth.go:111` | store | immediately | verification reads the live cap (`VerifySession`, `session.go:206`), so a change reaches every session at its next request, in both directions |
| `DUBPLATE_NONCE_RETENTION` | `auth.go:117` | store | next operation | the next sweep uses it; the correlated rule that it stays at or above the link expiry moves from the source's `Validate` hook to every write of either key |
| the four rate limits | `auth.go:166-184` | store | restart | the limiters are built once (`NewLimiter`, `internal/auth/limit.go:62`); a live apply is rung 14's option |
| `DUBPLATE_AUTH_REQUEST_BODY_CAP` | `auth.go:190` | store | next operation | the next request reads it |
| `DUBPLATE_UNKNOWN_ADDRESS_ANSWER` | `auth.go:196` | store | next operation | the next answer reads it |
| `DUBPLATE_SERVE_PREFIX`, `DUBPLATE_SERVE_BIND_ADDRS`, `DUBPLATE_SERVE_PORT` | `serve.go:35-59` | wall | restart | unchanged; read before the store opens |
| `DUBPLATE_SHADOW_FLOOR`, `DUBPLATE_SHADOW_BUDGET` | `shadow.go:73-88` | wall | restart | unchanged; the shadow unit's grant does not cover the store leaf |

The plan re-reads each cited line at its staleness pre-task, since rung 5's passes 5c and 5d may
move them. The immediate-apply row for the absolute cap is a statement of what the code does,
which is why 6d declares both lifetimes and the absolute cap `sensitive`, so rung 14's re-authentication gate covers them from its first commit. The environment
template keeps rendering every re-homed key, which stays valid as a seed, so the handoff needs no
secret minted and no template row removed.

---

## Identity in the store

Pass 6d opens only after rung 5's pass 5d merges, because every file it edits is rung 5's
territory until then.

The roster moves into the users table (ruling 6). The import runs in one transaction on the first
open that finds the roster file set and the users table empty, and it mints each member's
surrogate id there. The resolution point R5S names, where a login address becomes an identity
(`Lookup`, `internal/auth/roster.go:286`), reads the users table and returns the surrogate id as
the identity's `Principal`, the field `identity.go` reserves for it (`internal/identity/identity.go:115-119`).
The write path mints the id for a member added by the page or the verb. `Roster` is a concrete
type today (`roster.go:117`), and the plan decides whether 6d introduces the interface R5S
describes or replaces the implementation in place. The rule 5b-ii ruled, that `AddMember` refuses
the admin flag on every channel except the CLI grant 5c adds, carries into the store unchanged.
`streamer_username` becomes required, unique, and stable (S20:376-378).

The epoch re-keys from the digest of the address to the digest of the surrogate id, the change of
one derivation R5S's R34 prepared (R5S:1687-1690). That sentence holds for the code and not for
the data: every existing record's name is a digest of the old input, so each record still moves.
The records stay files (R18). The re-key is
idempotent and crash-safe: each record is written under its new name with `atomicfile` before the
old one is removed, and the reader consults the new key first and the old one second until a
completion marker is written. No member's revocation is lost, including across a crash mid-way.

`web-auth-security-reviewer` runs inside 6d's chain.

---

## The 2.0 forward hooks

Per CLAUDE.md's build-for-2.0 rule, each pass states per task whether it lands a hook or is
neutral. A hook is built by the rung that first consumes it.

| Hook (S20) | Rung 6's position |
| --- | --- |
| 1, audio key and file hash (:363) | Neutral. Named for rung 7, whose gate computes both; its migration adds the columns under invariant 6's opaque-id rule where it applies and invariant 1's for any path. |
| 2, folded sort key (:370) | Neutral. Named for rung 9's tagging. |
| 3, art blob (:372) | Neutral. Named for rung 9's admission, which decides blob storage. |
| 4, timestamped events (:374) | **Lands at 6a** as schema invariant 3; the job records carry timestamps. |
| 5, stable identity (:376) | **Lands at 6c and 6d.** 6c creates users; 6d mints the surrogate id and re-keys the epoch. |
| 6, universal driver seam (:379) | Neutral. |
| 7, schema migrations (:383) | **Lands at 6a.** |
| 8, collection writer (:385) | **Lands at 6c** as the `writer` column with its epoch and generation **[CH]**. |
| 9, transaction model (:387) | **Lands at 6a.** One writer, WAL, busy timeout, the one write function. |
| 10, live-file backup consistency (:390) | **Lands at 6b.** |
| 11, gains (:392) | Neutral. Named for rung 9. |
| 12, portable layout | Neutral. Rung 7. |
| 13, sort tags; 14, ID3 version | Neutral. Rungs 8 and 9. |
| 15, configuration homes (:401) | **Lands at 6c** for the schema, the settings home, and the `sensitive` marker; the admin API's two mounts and the re-authentication gate are rung 14's **[CH]**. |

---

## Prior art

Scored on CLAUDE.md's four criteria. Versions and dates are from `go list -m -versions` and
`go list -m -json` in a scratch module on 2026-09-23; licenses are read from each module's
`LICENSE` file in the module cache; importer counts are pkg.go.dev's "Known importers" figures
read the same day; repository figures are from the GitLab and GitHub APIs.

| Entry | Role | License | Quality | Maintenance | Adoption | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `modernc.org/sqlite` | The driver | BSD-3-Clause | Transpiled C under a hand-written `database/sql` layer, the only part the house bar grades. Its DSN parser validates every shorthand key before applying any (`sqlite.go:290-305`), which is the fail-before-partial-change behavior the store wants. Embeds SQLite 3.53.4, built with FTS5, snapshots, and `LIKE_DOESNT_MATCH_BLOBS`. | v1.59.0, 2026-09-15; ten releases since 2026-05-10, about every two weeks. 25 open issues, the oldest from 2020-09-05. Bus factor 1: 514 of the top contributors' commits are one author's, against 34 and 25 for the next two. | 3,518 importers; goose itself requires it | **Adopt**, meeting S20's condition: a gate test proves the FTS5 trigram tokenizer and WAL with a reader held open across a writer's commit. The bus factor is the field's norm, and the fallback below is the answer to it. |
| `github.com/ncruces/go-sqlite3` | The named fallback | MIT | A Wasm build of SQLite under wazero, well-documented VFS design | v0.35.6, 2026-09-23 | 96 importers; gonic uses it | **Record as the fallback** if the trigram or WAL test fails, as S20 records it. Still v0. |
| `github.com/pressly/goose/v3` | Migrations | MIT | Readable, with an `embed.FS` provider | v3.28.0, 2026-09-02; ten releases since 2025-01-07. 133 open issues, the oldest from 2019-01-10. Top contributors 261, 172, and 128 commits. | 1,725 importers | **Decline.** Its go.mod carries nineteen direct requirements, nine of them database drivers plus the moby Docker client, for a numbered-SQL-in-a-transaction mechanism the runner does in one small file. Its Go-function migrations answer a need no hook names. |
| `github.com/golang-migrate/migrate/v4` | Migrations | MIT | Driver-per-package design | v4.20.1, 2026-09-09 | 5,219 importers | **Decline.** Thirty-seven direct requirements, and its SQLite driver targets cgo and a separate modernc package whose version pairing becomes ours to track. |
| `zombiezen.com/go/sqlite` | A non-`database/sql` API over modernc | ISC (not re-read this session) | Ergonomic, with its own migration helper | v1.4.2, 2025-05-23 | not measured | **Decline.** It leaves `database/sql`, which every Go reader knows, for a second vocabulary, and its last release is sixteen months old. |
| `github.com/mattn/go-sqlite3` | The common driver | MIT | Mature | v1.14.52 | very high | **Decline** on OS's settled no-cgo fork (OS:412-413). |
| rclone, as a subprocess | Backup transport | MIT | The estate's proven tool | current | the estate already runs it | **Adopt** as a probed subprocess, as salvage did. `librclone` in-process is declined: it would pull rclone's entire dependency tree into the binary for one verb. |
| SQLite `VACUUM INTO` and the online backup API | Snapshots | public domain | Documented guarantees quoted above | SQLite's own | universal | **Adopt one**, chosen in 6b's plan with its guarantee quoted. |
| `internal/atomicfile` | Snapshot rename, epoch records | in-repo | house bar | in-repo | rungs 4 and 5 | **Adopt.** |
| `_salvage/internal/backup` and `restore` | Guards and flow | in-repo, salvage tier | salvage | frozen | ran on the box | **Adopt the guards on merit** under fresh review; import nothing. |

---

## Risk model

The store holds the only record the greenfield server keeps of what it admitted, which CLAUDE.md
now states in those terms. Its failure classes, ordered by cost:

- **Data loss by restore.** A restore over a live store, or one that leaves the old `-wal` beside
  the new database, corrupts silently. Guarded by ruling 3's exclusive lock, the three-file move,
  and a test that proves a restore refused while a reader holds the store.
- **Data loss by backup.** An empty or missing library root syncs to an empty remote. Guarded by
  the two salvage guards, each with a test whose fixture is genuinely empty and genuinely absent.
- **A torn snapshot.** Guarded by hook 10's snapshot mechanism, proved by a test that writes
  during the snapshot and checks integrity and content.
- **A bad migration.** Guarded by per-migration transactions, the pre-migration snapshot, the
  newer-store refusal, the foreign-key check on table rebuilds, and the invariant test run
  against populated stores at every earlier version.
- **Writer contention.** Two processes write the store: serve and a timer-run `backup`. Guarded
  by one write pool per process, `IMMEDIATE` transactions, and the busy timeout, proved by a
  two-connection test in which the second writer waits rather than failing. No write transaction
  spans network or file I/O.
- **Crash consistency.** SQLite's own journal covers the database. Files outside it (the snapshot,
  the epoch records) go through `atomicfile`, which fsyncs the directory
  (`internal/atomicfile/atomicfile.go:45-55`). The move-intent journal CH defines is the pattern
  for any later operation that spans the filesystem and the store.
- **Foreign keys silently off.** A pool connection opened without the setting enforces nothing.
  Guarded by the DSN rule and a test that inserts an orphan through a second pooled connection.
- **Settings drift.** An env value an admin has overridden could be silently ignored or silently
  win. Guarded by the env-seed table, whose admin-and-different row fails the load.
- **Revocation lost at the re-key.** Guarded by the idempotent, reader-tolerant re-key in 6d.

---

## Planning-miss prevention

Every fix round in pass 5b-ii was an ambiguity its plan left open (CLAUDE.md, "Planning-miss
prevention"). The equivalent questions for rung 6 are closed here or in the rulings.

### Behaviors a task relies on, quoted rather than recalled

| Behavior | Where it is settled |
| --- | --- |
| modernc's DSN keys, their validation, and the `_pragma` order | quoted above from v1.59.0; the plan re-quotes from the pinned version |
| `database/sql` opening new connections at will | the DSN rule; the plan quotes the package documentation |
| a deferred transaction's failed upgrade | quoted above; the plan quotes whether the busy handler runs |
| foreign keys not toggling inside a transaction | quoted above |
| `user_version` rolling back with a failed migration | proved by a test, not assumed |
| `VACUUM INTO`'s consistency and its sync condition | quoted above |
| WAL on one host only, and its read-only conditions | quoted from sqlite.org/wal.html sections 1 and 5 |
| the hot-journal pairing rule for restore | quoted above |
| `LIKE` never matching a BLOB in this build | the build flags above |
| `exec.CommandContext` cancellation and `WaitDelay` | the plan quotes `os/exec` |
| rclone sync, its archive flag, and its exit codes | the plan quotes rclone's documentation |
| advisory lock release on exit | the plan quotes the lock call's documentation |

One inherited claim is corrected rather than carried: R5S says the digest makes the epoch
re-key "a change of one derivation rather than a rename of every record" (R5S:1689-1690). That
is true of the code. Every existing record is still named by the old digest, so 6d moves each
one, as the section "Identity in the store" states.

### Every state a check can meet, with its report

| Check | States and reports |
| --- | --- |
| store open | absent: created by writing verbs, "absent" from `check` and `doctor`; zero bytes: refused as not a store; not a database: refused; older: migrated after the snapshot; newer: refused naming both versions; lock contended: writers wait, `restore` refuses; network filesystem: refused |
| job recovery | running with its lock held: left alone; running with its lock free: marked interrupted |
| backup | flag off: unavailable with a reason; library absent or empty: aborted before rclone; already running: conflict naming the job; remote failure: failed with rclone's exit status |
| restore | store held: refused naming the holder; snapshot absent, corrupt, or newer: refused; library non-empty: refused without the explicit flag |
| env seed | the eight rows of the table above |
| roster import | file set and users empty: import; file set and users populated: the ruling 6 load error; file unset and users empty: `doctor` reports no members, as rung 5's roster row does today |
| epoch re-key | new key only, old key only, both, neither: each read resolves, and a test covers all four |

### Privileged fields and every channel that may set them

| Field | Channels |
| --- | --- |
| settings source `admin` | none at rung 6; rung 14's admin API alone |
| a job's actor | the implementation, from its caller; never a request field |
| a collection's `root_path` and `writer` | the server alone; no request type carries either |
| a user's admin flag | the CLI grant alone, as 5b-ii ruled |
| backup | the CLI verb, and over HTTP only behind rung 14's authorizer, since the projection is unmounted |
| restore | the CLI verb alone, offline |

### Proofs and the fixture state each needs

Each proof names its fixture so it cannot pass vacuously. The migration test runs against a
populated store at each earlier version, never only an empty one. The WAL test holds a read
transaction open across a committed write. The snapshot test writes while the snapshot runs. The
interruption test plants a running record with its lock free. The env-seed conflict test plants an
admin-sourced row. The invariant test carries one violating fixture per rule. The foreign-key test
forces a second pooled connection. The backup guard tests use a library root that exists and is
empty, and one that does not exist. The restore refusal test holds a real reader open.

---

## The pass split

| Pass | Scope | Exit gate, as outcomes | Tree |
| --- | --- | --- | --- |
| **6a** The store foundation | The driver with its adoption tests; the `store` leaf, the lock, the two pools, the DSN settings; the write function; the migration runner with the pre-migration snapshot and the newer-store refusal; the schema-invariant test; the job records table. | `check.sh` green, plus: the trigram tokenizer answers a substring query; a reader holding a transaction open sees its snapshot while a writer commits; a second writer waits for the first and succeeds; an orphan insert through a second pooled connection fails; a failed migration leaves `user_version` and the schema unchanged; a newer store is refused naming both versions; a migration over a populated store leaves a readable pre-migration snapshot; each invariant fails against its violating fixture; the store refuses to open on a network filesystem and `check` never creates one. | Worktree (lane 2) |
| **6b** The ops seam, backup, and restore | `internal/ops`, `ops/local`, the unexported projection and the contract suite; `internal/backup`; the backup integration class declared in rung 6's own file, the reserved class removed from `keys.go`, and its names registered as supersessions; the `backup` and `restore` verbs; `doctor`'s backup row. | `check.sh` green, plus: every operation's cases pass in-process and over HTTP; every catalog entry declares a privilege; a job started over HTTP is polled and canceled over HTTP; canceling a finished job reports its state; a job whose process died reads as interrupted; a second backup start returns a conflict; with the flag off the catalog reports backup unavailable with its reason; an empty and a missing library each abort before rclone runs; a snapshot taken during writes passes its integrity check; `restore` refuses while a reader holds the store and moves the database, `-wal`, and `-shm` aside together; setting a retired R2 name fails the load and names its replacement. | Worktree (lane 2) |
| **6c** The schema and the settings home **[CH]** | Every table CH routes to rung 6 and the catalog skeleton; the registry's home, apply-mode, and sensitive metadata; `internal/settings` with the accessor, the env-seed rule, and the audit write; `doctor`'s value, source, and apply-mode rendering; rung 6's own tunables as the first store-homed keys. | `check.sh` green, plus: the invariant test passes over the full schema; an orphan collection owner is refused; every row of the env-seed table behaves as stated, each by its own test; a seed outside its bounds fails the load; a settings write writes one audit row and a refused write writes none; a key registered without a home still reads from the env; a cached value is invalidated by a write. | Worktree (lane 2), if 5c and 5d leave `internal/config` untouched; otherwise after 5d |
| **6d** Identity in the store and the handoff | The roster import and ruling 6's supersession; the surrogate id at the resolution point and the write path; the epoch re-key; the ten auth thresholds re-homed; the serve unit's grant and the `deploy.sh` pre-create for the `store` leaf, and the state-dir check extended to it. `web-auth-security-reviewer` runs inside. | `check.sh` green, plus: a roster imports once and a second open imports nothing; an identity's principal is the surrogate id; a member added through the store receives a new id and an admin flag only through the CLI grant; a session revoked before the re-key stays revoked after it, including after a crash injected mid-way; each re-homed key seeds from the env and `doctor` shows its source; the correlated retention rule refuses a write that breaks it; `systemd-analyze verify` is clean and the unit starts with the leaf writable. | After rung 5's 5d merges |

### Band and ceilings

The ladder drew rung 6 as band M, one pass (BS:656, with the band definitions at BS:563-567). Two later sources grew it: the ratified
CH routes a schema and a settings home here **[CH]**, and rung 5's spec hands rung 6 the surrogate
id, the epoch re-key, and the roster's move (R5S:1807-1811). Four passes make it band L.

The ceilings come from the recorded actuals in `docs/HISTORY.md`. Plan preparation is roughly
fixed per pass: 0.83M at 5a, 0.87M at 8a and 8b, 1.07M at 1b. The chain runs about 0.09M to 0.1M
per agent: 1.68M across 18 agents at 5a, 1.51M across 16 at 8b, 2.27M across 24 at 8a. Whole
passes closed at 2.9M against 3.25M (8b, HISTORY:2220), 3.6M against 4.5M (8a, HISTORY:2627),
3.9M against 4.25M (5a, HISTORY:1735), and 3.7M against 2.5M (1b, HISTORY:2961), and 5b-i ran
about 5.5M against 5.0M over ten tasks (HISTORY:578-590). Two lessons bind: a pass that changes the
config contract overran (1b), and a pass over ten tasks overran (5b-i).

| Pass | Tasks, estimated | Ceiling | Flag (80%) | Basis |
| --- | --- | --- | --- | --- |
| 6a | 6 | 3.5M | 2.8M | 8b's size with a margin for novel correctness |
| 6b | 8 | 4.5M | 3.6M | 8a's shape: a new contract over eight to nine tasks |
| 6c | 8 | 4.5M | 3.6M | a config-contract change, priced above 1b's 3.7M actual |
| 6d | 7 | 4.5M | 3.6M | 5a's actual plus its security review (0.18M) and fold (0.22M) |

The total is 17M. If a plan exceeds ten tasks, the conductor splits the pass rather than the
tasks, names the follow-on, and re-derives both ceilings. Each ceiling is re-derived from the
previous pass's actual before the next plan is drafted, and the lane's sub-ceiling comes from the
phase ceiling at launch (PA section 4).

### Contested surfaces

The conductor writes these rows into PA's ownership table (PA:173) before lane 2 launches, per
lane-launch checklist item 2.

| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/engine/store` (new) | rung 6, created by 6a; no other lane writes it | rungs 7, 9, 10, 13, 14, 16; 2.0's projector |
| The migration set under `internal/engine/store` (new) | rung 6 owns the runner and the invariant test; every later rung appends migrations | a rebase onto another lane's migration renumbers the rebasing lane's own files, and that rebase is non-fast-forward, so merge ritual step 5 runs |
| `internal/ops`, `internal/ops/local` (new) | rung 6, created by 6b | rungs 7, 9, 13, 14 add operations; the TUI post-launch |
| `internal/ops/opshttp` (new) | rung 6, created by 6b, unexported; rung 14 exports and mounts it | rung 14 |
| `internal/backup` (new) | rung 6, created by 6b | rung 16 schedules it |
| `internal/settings` (new) **[CH]** | rung 6, created by 6c | every rung that declares a store-homed key; rung 14's admin API |
| `internal/config/registry.go`, `thresholds.go` (amends the `internal/config` row) **[CH]** | rung 1b's contract; rung 6's 6c adds the home, apply-mode, and sensitive metadata additively | every rung; rung 5's lane does not edit these files while 6c runs, and if it must, 6c waits |
| `internal/config/keys.go` edits (amends the existing row) | rung 6's 6b removes the "backup targets" reserved class in the commit that declares its replacement in rung 6's own file, and registers the retired names through `AddSuperseded` there | every later rung, by the same rule |
| `cmd/dubplate-server/main.go`, `newProductionRegistry` and the `AddCommand` block (amends the existing row) | rung 1's skeleton; 6b adds one registration line and the two verbs, 6c one registration line if its keys need one; rung 5's 5c adds the `member` group in the same block | a textual conflict keeps both lines in landing order, and the rebase triggers merge ritual step 5 |
| `cmd/dubplate-server/doctor.go` (amends the existing row) | rung 1's skeleton; 6b adds one call to a row function defined in its own file, 6c the source and apply-mode rendering | rungs 5, 9, 13 add rows the same way |
| `go.mod`, `go.sum` (new row) | every lane adds its own requirements | a rebase conflict is resolved by `go mod tidy` in the rebasing lane, and merge ritual step 5 runs |
| The state directory layout (amends the existing row) | rung 4 owns it; rung 6's 6a adds the `store` leaf and its permissions | rung 16 converges |
| `internal/auth` roster and epoch, `cmd/dubplate-server/auth.go`, `internal/identity` (amends rung 5's rows) | rung 5; rung 6's 6d after 5d merges, with no concurrent auth-writing lane | rungs 7, 12, 14 |
| `systemd/dubplate-serve.service`, `scripts/deploy.sh`, the state-dir leg in `scripts/check.sh`, `env/dubplate-server.env.template` (amends the existing rows) | rung 6's 6d after 5d merges: the `store` leaf's grant, its pre-create, and the check | rung 16 converges |

---

## Process and review

Each pass opens with a staleness pre-task that re-reads every citation it leans on, then runs the
implementer, `diff-reviewer`, and gate chain per task, with segments of three to four tasks on
green commits. Rung 6 is not one of the three rungs that keep the full fresh-context adversarial
cycle (CLAUDE.md, the build-process section), so review scales to its risk. The data-loss paths
earn one targeted fresh-context read at 6b's close, over the migration runner, the snapshot, and
restore, by an `opus` reviewer briefed with the risk model. Pass 6d touches authentication, so
`web-auth-security-reviewer` runs inside its chain. At each merge the conductor dispatches one
`go-architecture-reader` per touched package, and every `code-simplifier` dispatch carries
`.claude/instructions/simplifier-brief.md` verbatim.

Passes 6a through 6c are disjoint from rung 5's territory except the rows above, so they can run
as lane 2 beside 5c and 5d. Pass 6d waits for 5d. Rung 6 touches the box only at 6d's
deployment, and the no-deploy carry in `docs/STATUS.md` holds until 5d lifts it.

---

## What depends on the configuration homes spec

CH is ratified. These passages exist because of its routes, and each is marked **[CH]** in the
text above:

1. Owner ruling 5, the schema scope, as it applies to CH's routed tables.
2. Owner ruling 6, the roster file, which rests on CH classifying members as store state and
   resolves CH's own wall list naming the roster file.
3. Owner ruling 7 and the seed-handoff table, which apply CH's wall-versus-store rule.
4. The tables in "The schema rung 6 lands" other than the job records, and the root-allocation
   paragraph that follows CH's 2026-09-23 erratum.
5. `internal/settings`, the registry metadata, and the env-seed table.
6. Pass 6c in full, and 6d's re-homing half.
7. The hook table's rows for hooks 8 and 15.
8. The contested-surface rows for `internal/settings` and the registry files.
9. Job-record retention as a store-homed threshold.

The foreign-key DSN rule does not depend on CH. CH requires it (CH:524-526), and hook 9 and the
schema's own references require it independently.

---

## Not verified in this draft

- Whether SQLite's busy handler runs when a deferred transaction's upgrade fails. The page quoted
  above is silent, and the `IMMEDIATE` rule makes the answer moot for the store's own writes.
- `zombiezen.com/go/sqlite`'s license and importer count, which were not re-read; its verdict does
  not rest on either.
- The exact wording of `database/sql`, `os/exec`, rclone, and the advisory-lock documentation the
  planning-miss table sends each plan to quote.
- The line citations into rung 5's files, which rung 5's passes 5c and 5d may move before 6d's
  staleness pre-task reads them.
