# Rung 6 spec review: contract and criteria lens

Reviewer: fresh-context adversarial read, lens CONTRACT AND CRITERIA, 2026-09-23. Target:
`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md` at `12a6941` (778 lines). Cited
code read at the working tree's `105ad50`.

**Verdict: not ready to ratify.** The spec is strong on the reasons for its choices and on quoting
SQLite. Its exit gates are weaker than its prose. Four problems block ratification. First, ruling 6
contradicts itself: the roster import needs `DUBPLATE_ROSTER_FILE` set, and the same ruling makes
that key a load error. The wall also cannot see the store, so it cannot make that error depend on
whether the import has run. Second, four of the six schema invariants cannot be decided by reading
SQLite's catalog, which is the only method the spec allows. A "violating fixture" can then be built
to match any heuristic, so those checks prove nothing. Third, the 6d exit gate lets the ten re-homed
auth thresholds seed store rows that no code reads. Every listed criterion passes while auth still
reads the env. Fourth, the job-recovery rule marks a live job interrupted whenever that job holds no
exclusive lock, and it breaks the first time a later rung adds a non-exclusive job kind. Beyond
these, about a dozen criteria can pass vacuously. Examples: a snapshot that is merely "readable", a
backup that is only ever "unavailable", an orphan insert that fails for being read-only rather than
for a foreign-key violation, and a `check` that "never creates" a store it never opens. Several
deliverables have no criterion at all: the authorizer, the archive path, library-tree restore, the
backup exit status, `doctor`'s new rows, and job retention. The restore verb has no stated
interface, and the backup has no stated layout on the remote. Counts: 5 blockers, 17 majors,
9 minors. Two findings, 11 and 16, carry owner forks, three questions between them. Finding 1's
choice of mechanism is technical.

---

## Blockers

### 1. Ruling 6's load error makes the import it depends on impossible, and the wall cannot know the import ran

- **Spec lines:** 91-101 (ruling 6), 508-513, 644 (roster import states), 677 (6d gate).
- **Defect:** The import "runs ... on the first open that finds the roster file set and the users
  table empty" (508-509). "After the import `DUBPLATE_ROSTER_FILE` is a load error" (92-93). The
  load happens in `config.Load`, before the store opens. `internal/config` "must never import
  `store`" (180-181). An unconditional supersession would fail the first boot, so the import could
  never run. A conditional error needs store state that the wall cannot read. The spec also leaves
  these open:
  - Which process removes the key between the importing boot and the next one. Unless someone
    edits the env first, the second `systemctl restart` fails production.
  - What happens when `deploy.sh` renders a template that still carries the key.
  - How a malformed, missing, or unreadable roster file is reported at import. That is the
    "many-bad" state.
  - Whether a member with an empty `streamer_username` is imported. See finding 11.
- **Evidence:** `cmd/dubplate-server/auth.go:46` lists `DUBPLATE_ROSTER_FILE` in the magic-link
  source's required set today. `internal/config/config.go:236-270` shows that `rejectNamespace`
  decides from the env and the registry alone. The 644 row "file set and users populated: the
  ruling 6 load error" names a report that the layer that raises load errors cannot produce.
- **Proposed fold:** Pick one mechanism and restate ruling 6 so a yes/no settles the build.
  - **(a) Import verb.** A one-shot, explicit, offline `member import-roster` verb. 6d removes the
    key from the required set and adds it to the superseded names, and the runbook orders the steps.
  - **(b) Store-side refusal.** The store-open path, not the load, refuses after import. The error
    wording names the key. The source layer drops the key from `Required` and moves it to
    `Optional`.

  Add a states row for the roster file at import: absent, empty, malformed, a duplicate address,
  and a member missing a streamer username, each with its report. Add a 6d criterion: after the
  import, the second boot behaves exactly as the chosen mechanism says, with the key still in the
  env.

### 2. Schema invariants 1, 4, 5, and 6 cannot be decided from the catalog, so the violating fixtures prove only a heuristic

- **Spec lines:** 263-278, 674 (6a gate "each invariant fails against its violating fixture"),
  676 (6c gate).
- **Defect:** The test must "inspect the result through SQLite's own catalog (`pragma_table_info`,
  `pragma_foreign_key_list`), not by parsing SQL."
  - **Invariant 1** ("a column that stores a filesystem path") depends on what a column means, not
    on catalog facts. The test can check only that a `path`-suffixed column is BLOB. A path column
    named `location` passes.
  - **Invariant 4** ("a column holding a source identifier is namespaced by its source") also
    depends on meaning. It cannot be read from a name or a type without a convention the spec does
    not define.
  - **Invariant 5** (attribute bag) has no stated shape predicate.
  - **Invariant 6** ("an opaque random value, never a rowid or a sequence") is a property of
    values, and the catalog does not show it. The test also cannot know which columns "a wire type
    may carry", because wire types live in `internal/ops`, which `store` does not import.

  Each violating fixture will be written to trip whatever heuristic the implementer chooses, so the
  "prove each rule can fail" demand is met without enforcing the rule. Invariant 3 is decidable
  only once the timestamp and kind column names are fixed, and the spec does not fix them.
- **Evidence:** Lines 266-278, quoted above. Line 674 lists "each invariant fails against its
  violating fixture" as the only proof.
- **Proposed fold:** Restate each invariant as a decidable naming-and-type rule that the test
  enforces. Examples:
  - Every column that carries a filesystem path ends in `_path` and is BLOB, and no BLOB or TEXT
    column outside an allowlist is named like a location.
  - Source-identifier columns match `^(mb|discogs|…)_[a-z]+_id$`, and a bare `release_id` or
    `recording_id` column fails.
  - Opaque ids are `id BLOB NOT NULL` columns of fixed length (for example 16 bytes), with no
    `INTEGER PRIMARY KEY` exposed under a public name.
  - An attribute bag is any table with one foreign-key column plus a `name`-like TEXT column plus
    a `value`-like column.
  - `_events` tables carry `occurred_at` and `kind`, both NOT NULL.

  Where a rule stays semantic, move it out of the test into a reviewer checklist and say so. For
  invariant 6, add a store-level test that ids minted by the store's id function are 16 random
  bytes and never equal a rowid.

### 3. The 6d gate passes when the re-homed thresholds seed rows nothing reads

- **Spec lines:** 443-452, 478-499, 677 (6d gate "each re-homed key seeds from the env and `doctor`
  shows its source").
- **Defect:** None of the 6d criteria shows that a consumer reads the accessor. Consider an
  implementation that writes the ten seed rows and leaves every auth consumer calling the wall's
  threshold value. It passes "seeds from the env", "doctor shows its source", and "the correlated
  retention rule refuses a write". The spec also leaves open whether `config.Load` still exposes a
  store-homed key's env value, which is the bypass channel.
- **Evidence:** Line 445: keys "read through the settings accessor, never from a long-lived
  struct". No gate item checks it. `cmd/dubplate-server/auth.go:99-200` registers these keys as
  wall thresholds today.
- **Proposed fold:** Add a 6d criterion for each apply-mode class. With the env holding value A
  and the store row holding value B, written in-package as an `admin` row, the consumer behaves
  per B:
  - a link minted after the write carries B's expiry;
  - a session's next refresh uses B's lifetime;
  - `VerifySession` enforces B's absolute cap at the next request;
  - a limiter built at boot uses B.

  State that `config.Config` no longer returns a store-homed key's value, or returns it only as
  the seed, so the compiler forbids the bypass.

### 4. Job recovery marks a live non-exclusive job interrupted, and queued records are never recovered

- **Spec lines:** 363-368, 640 (job recovery states), 675 (6b gate "a job whose process died reads
  as interrupted").
- **Defect:** "The runner marks interrupted every running record whose exclusive lock no live
  process holds." Only exclusive kinds hold a lock (365-366). A non-exclusive running job therefore
  always has its lock "free". Suppose rung 7 or 9 adds such a job and a timer-run `backup` opens
  the store while serve runs that job. The backup process marks serve's live job interrupted. "A
  terminal state never changes" (352), so serve's job can no longer record its real outcome. The
  states table has two rows (lock held, lock free) and no row for:
  - a non-exclusive running job;
  - a `queued` record whose process died;
  - a job that finishes after another process marked it interrupted.

  Rung 6 has only one kind, so every 6b criterion passes, and the defect ships to rungs 7 and 9 as
  contract.
- **Evidence:** Lines 363-367, quoted. Line 640 lists only two states.
- **Proposed fold:** Make liveness per process, not per kind. Every process that runs jobs holds
  an advisory lock on a per-process (or per-job) file whose name the record stores. Recovery then
  marks interrupted any running or queued record whose owner lock is free. Add these to the states
  table: queued with owner dead, running non-exclusive with owner alive, and finish-after-interrupt
  (the late writer's update is refused and logged). Add a 6b criterion that a record owned by a
  live second process survives another process's open. The fixture is a second store handle
  holding its owner lock.

### 5. Store open during `check` and `doctor`: migration on open contradicts "read-only", and "never creates" is vacuous at 6a

- **Spec lines:** 204-206, 453-455 ("`check` ... evaluates the rule read-only"), 639 (store open
  states), 674 (6a gate "`check` never creates one").
- **Defect:** The store "migrate[s] forward on open" (8-9, 243-252) and writes a pre-migration
  snapshot. The spec never says whether `check` and `doctor` open through that path. If they do,
  running `check` against an older store after an upgrade rewrites the database and the snapshot.
  That contradicts "read-only", and it also contradicts ruling 4's premise that only a newer
  binary's start migrates. If they do not, the "older: migrated after the snapshot" row does not
  apply to them, and their report for an older or a newer store is unspecified. The 6a criterion
  "`check` never creates one" is vacuous as written: at 6a, `check` has no code that opens the
  store (`cmd/dubplate-server/check.go` has no store reference), and 6a's territory lists no edit to
  `check.go` or `doctor.go` (719-721).
- **Evidence:** `grep -n store cmd/dubplate-server/check.go` returns nothing. The contested-surface
  row for `doctor.go` (721) names only 6b and 6c.
- **Proposed fold:** Split the open into two operations. One is a read-only inspect: `mode=ro`,
  it never migrates and never creates, and it reports absent, older (naming the pending
  migrations), newer, or current. The other is the writing open. State which verbs use which. Move
  the "check never creates" criterion to the pass that first makes `check` open the store (6c, when
  the seed rule is evaluated). Rewrite it as "`check` against an absent, an older, and a newer store
  leaves the state directory byte-identical and reports each state by name". Add `check.go` to the
  contested-surface rows for that pass.

---

## Majors

### 6. The 6a proofs for foreign keys, writer contention, and network filesystems can pass for the wrong reason

- **Spec lines:** 210-213, 593-601, 674.
- **Defect:** These three proofs each have a way to pass without testing what they name:
  - **Foreign keys.** "An orphan insert through a second pooled connection fails." The write pool
    holds one connection (212), so the only "second pooled connection" is the read pool's. If
    read-pool connections are read-only (the spec never says; see finding 20), the insert fails
    with `SQLITE_READONLY`, and the test passes without testing foreign keys.
  - **Writer contention.** "A second writer waits for the first and succeeds." Within one store
    handle, Go's pool queues the second caller on the single connection, and SQLite's
    `busy_timeout` is never exercised. The risk model's threat is two processes (593-595).
  - **Network filesystem.** "The store refuses to open on a network filesystem." A unit test cannot
    mount NFS. The seam that fakes the filesystem type, and its fixture, are unnamed.
- **Evidence:** Lines 212 and 674.
- **Proposed fold:** Fix each proof's fixture and assertion:
  - **Foreign keys:** assert a foreign-key constraint error by its extended code
    (`SQLITE_CONSTRAINT_FOREIGNKEY`), and force a fresh write connection by closing the pooled
    connection first (`SetConnMaxLifetime` or a `Conn` round trip). Otherwise, run the insert
    through a second `*store.Store` opened on the same file.
  - **Contention:** use two independent `Open` calls on one file. The first holds an `IMMEDIATE`
    transaction for about half the busy timeout, and the second commits after it. Add the negative
    case: holding the transaction past the timeout yields a typed busy error.
  - **Network filesystem:** name a `statfs` probe the test can substitute, list the magic numbers
    refused (NFS, CIFS/SMB, FUSE?), and add a report for "unknown" filesystem types.

### 7. "A readable pre-migration snapshot" is satisfied by a snapshot of the post-migration store

- **Spec lines:** 69-74, 251-252, 674.
- **Defect:** A snapshot taken after migrating is also readable. The criterion also leaves out
  what happens when writing the snapshot fails (disk full). Ruling 4 then needs "refuse to
  migrate", but the spec never says it. "Non-empty store" (251) is also undefined:
  `user_version > 0`, any user table, or any row?
- **Proposed fold:** Replace the criterion with this one. After migrating a populated store from
  version N to N+1, the snapshot's `user_version` is N, and its rows equal the fixture's. A
  snapshot write that fails leaves the store at N and returns an error naming the snapshot path.
  Define non-empty as `user_version > 0`.

### 8. The populated-fixture migration test is vacuous at 6a, and no rule makes later rungs extend it

- **Spec lines:** 263-266, 591, 660-661.
- **Defect:** "Applies every migration ... to a populated fixture at each earlier version." At 6a
  there is one migration, and its only earlier version is 0 (empty). So the populated-fixture leg
  runs on nothing. The spec also puts no duty on a later rung's migration to ship a populated
  fixture of its own prior version. Without that duty, the "every earlier version" claim decays
  into "the versions someone remembered".
- **Proposed fold:** Make it a contract rule in the migrations section. Every migration file ships
  with a fixture generator (or SQL) that populates the schema as of that version. The invariant
  test fails when a migration has no fixture, and it fails on that missing fixture itself. State
  that 6a's proof is "migration 1 over an empty store" and that 6c's is the first populated one.

### 9. The backup path has no end-to-end criterion, no named rclone fixture, and an unstated gate dependency

- **Spec lines:** 44-51 (ruling 1), 408-423, 641, 675.
- **Defect:** The 6b gate proves that backup refuses (unavailable, conflict, empty library, missing
  library). It never proves that a backup succeeds: that the snapshot and the tree reach the
  remote, and that a superseded file lands under the dated archive path. "Before rclone runs" is
  observable only through a stand-in rclone that records its invocations. A stand-in that always
  exits 0 is the silent-success fake CLAUDE.md forbids, and the spec does not name it. The spec
  also does not say whether the gate needs a real `rclone` binary. CLAUDE.md's gate rule ("never
  guard one with `command -v || skip`") makes that a real decision. Line 641 lists "remote
  failure: failed with rclone's exit status", and no criterion covers it.
- **Proposed fold:** Name the fixture. Either a real pinned `rclone` against a `local` backend in
  `t.TempDir()` (then add rclone to the gate prerequisites and fail when it is missing), or a
  recording stand-in whose exit code each case sets. Then add these criteria:
  - a successful run leaves the snapshot object and every library file on the remote;
  - a file deleted from the library appears under the archive path, not deleted;
  - a stand-in exit 7 makes the job `failed` with the code carried in the record.

### 10. The restore verb has no stated interface, the tree restore has no flow, and the fixtures miss the absent `-wal`

- **Spec lines:** 62-67, 425-433, 642, 675.
- **Defect:** The spec never states `restore`'s arguments:
  - how a snapshot is named (a date, an object path, a local file);
  - "fetches or reads" (which one, or both);
  - the flag for "names it explicitly".

  "Restoring the library tree" appears in one sentence (432-433), with no steps (an rclone copy
  from where, including the archive?) and no criterion. After a clean shutdown SQLite removes
  `-wal`, so the common restore meets an absent `-wal` and `-shm`. The "moves the database, `-wal`,
  and `-shm` aside together" criterion needs a fixture in which they exist, plus a stated behavior
  when they do not. Where the pre-restore set goes, and how long it is kept, are also unstated.
- **Proposed fold:** Add a restore contract subsection covering the command line, the snapshot
  naming (tied to the remote layout; see finding 15), and the tree-restore steps. Add criteria:
  - `restore` with a planted non-empty `-wal` and `-shm` moves all three into one named pre-restore
    directory;
  - with none present, it moves the database alone and succeeds;
  - a tree restore into an empty root yields the backed-up files;
  - into a non-empty root without the flag, it refuses and names the flag.

### 11. `streamer_username` becomes required at 6d, with no source for existing members

- **Spec lines:** 284-286, 517.
- **Defect:** "`streamer_username` becomes required, unique, and stable". The roster makes it
  optional today (`internal/auth/roster.go:109`, `json:"streamer_username,omitempty"`). The import
  has no rule for a member without one. The possible rules are to refuse the import (the box fails
  to boot), to derive one, or to leave it null until rung 12. Each changes what an adopter must do.
- **OWNER FORK:** whether an adopter must supply streamer usernames before upgrading, or the server
  derives them.
- **Proposed fold:** Add the state to the roster-import row, with its report, and put the choice in
  ruling 6's text.

### 12. The authorizer is a deliverable with no criterion

- **Spec lines:** 392-396, 655, 675.
- **Defect:** The projection "requires an authorizer with no default, which it consults with each
  operation's declared privilege before dispatch". No 6b criterion covers any of it. A projection
  that ignores the authorizer passes every listed item. "Every catalog entry declares a privilege"
  passes with the zero value, and the privilege vocabulary is never defined.
- **Proposed fold:** Add these 6b criteria:
  - a denying authorizer returns the forbidden code over HTTP, and the operation's side effect did
    not happen (no job record written);
  - constructing the projection with a nil authorizer fails at construction;
  - every catalog entry's privilege is a named, non-zero value from a defined set (state the set,
    for example `member` and `admin`).

### 13. The availability criterion passes for a backup that is always unavailable

- **Spec lines:** 381-383, 675.
- **Defect:** "With the flag off the catalog reports backup unavailable with its reason." An
  implementation that always reports unavailable passes. The spec also never says what starting an
  unavailable operation returns.
- **Proposed fold:** Test both directions. Flag off gives unavailable with the exact reason, and a
  start returns the unavailable code, with no job record written. Flag on gives available.

### 14. "Every error code survives the round trip" needs a producing case for each code

- **Spec lines:** 370-372, 399-401.
- **Defect:** At rung 6 no operation naturally produces `internal`. `invalid request` needs a
  malformed body. A suite that iterates over the codes it happens to observe cannot fail for a code
  it never meets.
- **Proposed fold:** State that the suite holds one case per sentinel, and name each case's trigger:
  - unavailable: flag off;
  - not found: an unknown job id;
  - conflict: a second start;
  - invalid: a malformed body;
  - canceled: canceling a running job;
  - internal: an injected store fault through a test-only constructor option, never a catalog
    entry.

  The suite also fails if a sentinel has no case.

### 15. Ruling 2 names no keys, and neither the remote layout nor the rclone credential channel is specified

- **Spec lines:** 53-60, 675 ("names its replacement").
- **Defect:** "One required remote spec" has no key name, so "names its replacement" (the
  supersession message) has nothing definite to name. The remote layout under the one `remote:path`
  is also undefined. It needs paths for snapshot objects and their date format, the library prefix,
  and the archive prefix, and restore needs to resolve them. Nothing says how rclone's own
  configuration reaches the box. It could be a file (whose path, under which unit's
  `ProtectHome`?) or rclone's `RCLONE_CONFIG_<REMOTE>_*` env form. CLAUDE.md requires every secret
  to originate in the age store and be rendered by `deploy.sh`. A credential "in rclone's own
  configuration" has no route there. It falls between 6b (which deploys nothing) and rung 16 (which
  schedules).
- **Evidence:** `internal/config/keys.go:60-71`. `scripts/deploy.sh:468-470` still creates the
  legacy staging directory.
- **Proposed fold:** Name the keys, for example `DUBPLATE_BACKUP_ENABLED` and
  `DUBPLATE_BACKUP_REMOTE`, and write each retired name's supersession message verbatim. Add a
  "remote layout" table. State which rung provisions rclone's configuration from the age store, and
  add that rung to the non-scope list with its owner.

### 16. The backup's retention and its coverage of state outside the store go unstated

- **Spec lines:** 44-51, 157-159.
- **Defect:** Snapshots are "dated objects" with no pruning. The archive is left to "the remote's
  own lifecycle policy". A `local`, `sftp`, or `webdav` remote has no such policy, which fails BS's
  universality bar for any rclone backend. The backup also carries only the store and the tree. The
  auth leaf, which holds the epoch (revocation) and nonce records, stays in the state directory
  (R18, line 158) and is not backed up. A restored box therefore forgets every revocation, and a
  session revoked before the disaster becomes valid again under the same session secret.
- **OWNER FORK:** whether dubplate prunes snapshots and the archive itself (with a store-homed
  retention), and whether the auth leaf is part of a backup.
- **Proposed fold:** Put both into ruling 1's question. If the auth leaf is excluded, the restore
  runbook states that the session secret must be rotated after a restore, and the spec says so.

### 17. The env-seed table has no state for a changed default or an out-of-bounds stored value

- **Spec lines:** 456-471.
- **Defect:** "No row | absent → the row is written at its default, source `default`." A later
  release that changes the researched default leaves every existing box on the old value, because
  "`default` | absent → nothing changes". The table also has no row for:
  - a stored value that the current binary's bounds now reject (bounds tightened across releases);
  - a row for a key that is no longer registered (renamed or retired);
  - a store-homed key whose env value is unparseable. Does the wall still parse it, or does the
    seed?
- **Proposed fold:** Either do not persist `default` rows (read-through to the registry default),
  or add the rule that a `default` row follows the binary's default on boot. Add three rows:
  - stored value out of bounds: the load fails naming the key and the bound;
  - orphan row: ignored and reported by `doctor`;
  - unparseable env value: a wall error, as today.

### 18. The spec does not say which processes run the seed rule or whether seeding writes audit rows

- **Spec lines:** 453-455, 468-471, 676 ("a settings write writes one audit row").
- **Defect:** The seed rule "runs once at boot" without naming the process. Serve, the `backup`
  verb (whose rung 16 unit may carry a different env, as CH notes for the pings unit), and the
  `member` verbs all open the store. Two processes with different env values would re-seed back
  and forth. An `admin`-and-different row would also fail a timer-run backup over an auth
  threshold it never uses. Whether a seed or a default write counts as a "settings write" that
  audits, and what actor and channel it records, is undefined, so the audit criterion is ambiguous.
- **Proposed fold:** State that only `serve` applies the seed rule, and that other writing verbs
  read settings and never seed. Add a criterion that `backup` with a divergent env leaves the rows
  unchanged. State that seeds audit with actor `system` and channel `env-seed`, or that they do not
  audit, and make the criterion match.

### 19. Job-record retention falls between 6b and 6c

- **Spec lines:** 367-368, 473-475, 173, 675-676.
- **Defect:** 6b creates the job model, and terminal records are "retained for a store-homed
  threshold's duration". The settings home arrives at 6c, and `ops/local` imports `settings`
  (173), a package that does not exist when 6b runs. Neither exit gate proves pruning happens. The
  spec also does not say whether pruning keeps the last successful backup, which `doctor`'s "age of
  the last successful backup job" (387-388) depends on.
- **Proposed fold:** Assign pruning to 6c with a criterion: a terminal record older than the
  retention is pruned at the next job start or open, and the newest successful record of each kind
  is never pruned. Remove `settings` from `ops/local`'s imports at 6b.

### 20. The read pool's write capability, the accessor's cross-process invalidation, and the privileged-channel table are incomplete

- **Spec lines:** 210-213, 231-235, 444-445, 647-656.
- **Defect:** "The store exposes one way to write", but the read pool is not stated to be read-only
  (`mode=ro` or `query_only`). Any caller can write through it outside `IMMEDIATE`, which is an
  unnamed channel for every privileged field. The accessor's "version that invalidates cached
  values after a write" is not stated as in-process or cross-process (a CLI write while serve
  caches). The privileged-field table leaves out the settings row's `default` and `env-seed`
  sources, a job's terminal state (which recovery in another process can set), and the users
  table's surrogate id (which must never come from a request or the roster file).
- **Proposed fold:** Open read connections with `query_only=1` in the DSN, and prove it: a write
  through the read pool fails with a read-only error. Define the version as a store row read on
  each access, and add a cross-handle invalidation criterion. Add the three rows to the table.

### 21. The CLI `backup` exit status and the SIGINT path have no mapping and no criterion

- **Spec lines:** 359-361.
- **Defect:** "Exits with a status derived from the terminal state." Rung 16's timer and the
  failure ping depend on this contract, but the mapping is unstated for succeeded, failed,
  canceled, and unavailable, and for a conflict when another run is active. No criterion covers the
  path from SIGINT to cancel to the canceled state.
- **Proposed fold:** Add a table of exit codes and a 6b criterion: `backup` receiving SIGINT
  mid-sync ends the job `canceled`, the stand-in rclone was killed (no orphan process after
  `WaitDelay`), and the process exits with the stated code.

### 22. Cancel semantics are unstated for a queued job and for a job running in another process

- **Spec lines:** 353-357, 359, 401.
- **Defect:** A job runs in the process that started it. A cancel sent to serve for a job the CLI
  started, which is the rung 16 scenario once the projection is mounted, cannot cancel a context in
  another process. The spec says only that cancellation never relies on a context crossing a
  process. Canceling a `queued` job is also unspecified. The HTTP criterion "polled and canceled
  over HTTP" does not state the expected terminal state.
- **Proposed fold:** Define cross-process cancel. Either the record carries a cancel-requested flag
  that the owning runner polls, or cancel returns `conflict` naming the owning process. Define a
  queued job's cancel as going straight to `canceled`. Make the HTTP criterion "ends `canceled`,
  and the running work stopped".

---

## Minors

### 23. Two claims about the environment template are inaccurate

- **Spec lines:** 497-499, 725.
- **Defect:** "The environment template keeps rendering every re-homed key". None of the ten auth
  thresholds appear in `env/dubplate-server.env.template`, which carries only the shadow floor and
  budget among thresholds (lines 110-117).
- **Proposed fold:** Correct the sentence. Drop the template from 6d's contested-surface row
  unless 6d edits it for ruling 6.

### 24. 6a writes files that no contested-surface row names

- **Spec lines:** 710-725.
- **Defect:** The busy timeout is "a registered threshold ... homed on the wall" (234-235), and it
  must be declared in a rung-6 file and registered in `newProductionRegistry`. The `main.go` row
  names only 6b and 6c. `doctor` reporting the filesystem type (201) is also unassigned to a pass.
- **Proposed fold:** Add 6a to the `main.go` and `doctor.go` rows. Give each pass row a Files
  column.

### 25. The torn-snapshot criterion checks integrity only, but the risk model promises content

- **Spec lines:** 587-588, 675.
- **Defect:** The exit gate says only "passes its integrity check". Consider a byte copy of the main
  file without its WAL. It passes `integrity_check` and loses committed rows.
- **Proposed fold:** Add the content condition: every row committed before the snapshot began is
  present, and none from a transaction that had not committed.

### 26. The epoch re-key's criterion "each read resolves" states no outcome

- **Spec lines:** 523-525, 645, 677.
- **Defect:** The criterion does not say what a read resolves to when both keys exist. A
  concurrent `Bump` during the copy can be overwritten by the older copied value. The spec also
  does not say when the re-key runs, whether at serve boot before listening or with serve live.
- **Proposed fold:** "Both present resolves to the later epoch. The copy writes max(old, new).
  The re-key runs at serve start, before the listener opens." Add a criterion that a session
  neither revoked nor expired before the re-key is still valid after it.

### 27. `doctor`'s new rows have no criterion

- **Spec lines:** 387-388, 451-452, 497, 675-677.
- **Defect:** The backup row, the value, source, apply-mode, and sensitive rendering, and 6d's
  `sensitive` declarations on the two lifetimes and the cap are each a stated deliverable with no
  criterion.
- **Proposed fold:** Add golden-output criteria to 6b, 6c, and 6d, including the "never backed up"
  and pruned-history states.

### 28. The default collection's creation falls between rungs

- **Spec lines:** 287-292, 306-312.
- **Defect:** `collections.owner` is a non-null reference to users, and users is empty until 6d.
  CH's default collection is admin-owned. The spec does not say who creates it or when (6d after
  import? rung 7? rung 10's migrator?).
- **Proposed fold:** Name the rung and add it to the non-scope list.

### 29. The meaning of "reserved by name" for share rows is unclear

- **Spec lines:** 291-292, 288.
- **Defect:** The spec leaves undefined:
  - what reserving share rows by name produces (a comment, a reserved table name the invariant
    test refuses, or nothing);
  - `visibility`'s single value and its CHECK constraint;
  - `writer`'s domain (a server sentinel, or a reference to a device credential).
- **Proposed fold:** State each in one line, so that hook 8's "lands" claim has a checkable column
  type.

### 30. Hook 4 "lands" at 6a on zero tables

- **Spec lines:** 541.
- **Defect:** No `_events` table exists at rung 6, so invariant 3 is proven only by its violating
  fixture.
- **Proposed fold:** Say "the rule lands, and the first events table is rung N's".

### 31. Some states in the states table have no proof

- **Spec lines:** 639.
- **Defect:** These states have reports but no exit criterion:
  - a zero-byte file refused;
  - not a database refused;
  - `backup` creating an absent store. This one is questionable: a backup of a store that does not
    exist should probably refuse rather than snapshot an empty one.
- **Proposed fold:** Add one criterion per state. Reconsider `backup` as a creating verb.

---

## Top five

1. Blocker: ruling 6 makes `DUBPLATE_ROSTER_FILE` a load error, yet the import needs it set, and the
   wall cannot see the store to make that error conditional.
2. Blocker: schema invariants 1, 4, 5, and 6 cannot be decided from `pragma_table_info`, so their
   violating fixtures prove only a heuristic.
3. Blocker: the 6d gate passes while auth still reads its thresholds from the env; no criterion
   shows a consumer following the store row.
4. Blocker: job recovery marks any lock-free running record interrupted, which kills live
   non-exclusive jobs, and queued records are never recovered.
5. Blocker: `check` and `doctor` opening the store would migrate it, which breaks "read-only", and
   the 6a criterion "`check` never creates one" is vacuous because `check` does not open the store
   at 6a.
