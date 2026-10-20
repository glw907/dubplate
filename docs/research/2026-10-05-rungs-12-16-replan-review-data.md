# Rungs 12 and 16 re-plan: data integrity and failure-risk review

Target: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md` at `88c535e`. Lens: data integrity
and failure risk (crash, partial write, upgrade, restore, concurrent process). Read-only; the box was not
contacted. Box facts come from `docs/STATUS.md` and `docs/HISTORY.md` (5d-ii).

Findings are correctness gaps ranked by consequence, then over-ceremony ranked by cost. Polish is omitted.

Counts: 2 high, 2 medium, 4 low; 2 over-ceremony; 1 owner fork.

## Correctness gaps

### D1 (high): the Navidrome archive is a single, unverified copy that must survive until rung 16

**Where:** spec:24-27 (ruling 2), spec:75-79 (rung 16), spec:82-83 (disposal at rung 17).

**Defect.** Ruling 2 promises members keep "their stars, playlists, and play counts", and the only source is
`/root/musicbox-retired-2026-10-05.tar.gz` on the box (HISTORY 5d-ii: "archived at
`/root/musicbox-retired-2026-10-05.tar.gz` (0600)"). The import runs at rung 16, after rungs 6d, 7, 9, 10, 11,
13, 14, and rung 12's eight slices. Over that interval, the archive faces three risks:

- **One copy.** A box loss, a reprovision, or an `rm` destroys the data for good. Rung 16 is a first install
  that "walks [the install path] as a stranger would, on the box" (spec:48-49). A clean reprovision is a
  natural way to walk that path, and it wipes `/root`.
- **No integrity record.** No sha256 or content listing is recorded, and nothing records whether Navidrome
  was stopped before the tar ran. The streamer spec's import reads "a stopped Navidrome's SQLite database"
  (streamer spec:543). A database archived while Navidrome ran may lack its `-wal` file, and the last
  writes would then be missing with no error.
- **Unconfirmed fallback.** The old stack once snapshotted `navidrome.db` to R2
  (`_salvage/internal/backup/backup.go:94`, `bucket music-library`). Nothing in the repo confirms that a
  snapshot still exists, how old it is, or that the credentials survived the retirement of 42 age-store
  names.

**Fold.** Add a document-pass item, or a conductor chore, that runs now and does not wait for rung 16:

1. Over read-only ssh, list the tar and confirm that `navidrome.db` is present, and its `-wal` and `-shm`
   files if any.
2. Record the archive's sha256.
3. Extract a copy off the box and run `PRAGMA integrity_check` on it.
4. Keep an encrypted copy off the box.
5. Record the copy's location and the sha256 in STATUS.

Rung 16's import then runs against a verified copy, never the archive in place. Rung 17's disposal decision
covers every copy.

**OWNER FORK: where the second copy lives.**

- (a) Age-encrypted on the workstation, next to the secrets store. This is cheap and needs no new
  infrastructure.
- (b) In dubplate's own rclone crypt remote. That remote does not exist until rung 16 provisions backup.
- (c) Both.

Recommendation: (a) now, and (c) once rung 16 provisions the remote. The file holds members' addresses and
listening history, so the copy must be encrypted.

### D2 (high): the roster-row migration the spec requires has no path that satisfies its own constraint

**Where:** spec:152-154.

**Defect.** The spec says the deploy migrates the old-key admin row "through an existing verb or the
bootstrap path, never by hand". No such path exists:

- **The loader refuses the old key.** After the rename it rejects the whole file. `parseRoster` decodes
  with `json.Unmarshal(data, &file, json.RejectUnknownMembers(true))` (`internal/auth/roster.go:340`), and
  `go doc encoding/json/v2 RejectUnknownMembers` reads "unknown members should be rejected when
  unmarshaling a JSON object". A `streamer_username` member therefore fails as `ErrRosterUnparseable`.
- **No verb can repair it.** Every write first re-reads the file through the same loader. `writeLocked`
  calls `readRoster(path)` (`roster.go:448`), so `member add` fails as well. Only `add` and `list` exist
  (`cmd/dubplate-server/member.go:50-51`), and neither removes or edits a member.
- **The bootstrap needs the file absent.** `CreateRoster` "refuses with ErrRosterExists when path already
  exists" (`roster.go:269-270`), and `addMember` reaches it only on `ErrRosterMissing`
  (`member.go:118-119`).

The only working sequence moves the file aside and then bootstraps, and moving the file is the by-hand step
the spec forbids. A plan written against spec:154 either stalls or improvises.

**What is safe.** "No compatibility shim" is safe:

- The failure is loud. The roster loader refuses the file, and doctor's `auth store` row reports it.
- No member can be locked out. The serve unit is installed and disabled (STATUS: "the serve unit installed
  and disabled"), the group is not live, and the file holds one admin row.
- No silent loss is possible. An empty username can't result, because the decode fails before validation
  runs.

**Fold.** Replace spec:152-154 with this path:

1. The re-bootstrap belongs to the first deploy that needs a readable roster (6d's `member import`, or
   rung 16's first install). It does not belong to the removal pass (see O1).
2. That deploy moves `roster.json` aside to `roster.json.pre-username` (mode 0600, music-owned) with one
   command the plan names.
3. It then runs `member add <address> --admin --name ... --username ...`, which takes the bootstrap path.
4. Doctor's `auth store` row reads ok.

The sequence is a named and recorded step, so it is not the hand edit that `CLAUDE.md`'s safety rail
forbids ("Never edit the box by hand and call it done").

### D3 (medium): rung 16's stated rollback has no backup behind it, and the spec orders no backup before the irreversible steps

**Where:** spec:75-79 ("Its rollback is the runbook's restore from backup"), and the carried-items table at
spec:193-211.

**Defect.** Nothing exists to restore:

- The box runs with backup off. `env/dubplate-server.env.template` carries no `DUBPLATE_BACKUP_ENABLED`
  row, and the backup verb exits 69 "when `DUBPLATE_BACKUP_ENABLED` is off"
  (`docs/server-runbook.md`, Backup, item 5).
- No deploy rollback set exists either. STATUS records "No rollback set remains: the `.prev` copies rolled
  back to the retired stack".

Rung 16 also runs two irreversible writes. The multi-disc re-layout moves library files on the only
dubplate-managed copy of `/srv/music/library`. The import writes every member's state. The
carried-items table has no row for provisioning the backup remote, its crypt wrapper, or its secrets.

**Is a runbook restore adequate for a first launch?** Yes, once it has something to restore. No incumbent
holds the group's listening, so a failed launch costs availability, never data, provided a verified
backup predates each irreversible step.

**Fold.** Add a carried-items row: "Backup provisioned and proven before the re-layout | Rung 16". Then make
rung 16's order explicit:

1. Provision the crypt remote and its secrets, and enable backup.
2. Run one backup, and prove a restore into a scratch root.
3. Take the deploy `.prev` set (runbook, Deploy).
4. Run the re-layout.
5. Take a store snapshot directly before the import. The store keeps local snapshots
   (`DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`, `cmd/dubplate-server/store.go:72`).
6. Run the import.

### D4 (medium): the rung 16 import's member mapping and rerun behavior are unspecified, and a table row calls the mapping moot

**Where:** spec:209 ("Navidrome username matching is case-insensitive | Moot"), and spec:24-27.

**Defect.** Ruling 2 runs the import against a roster that already holds Geoff's admin row, and rung 16 may
hold members added through the web page. The streamer spec says the import "invites each Navidrome user as
a member through the roster's normal add path" (streamer spec:544). That path refuses a duplicate address or
username (`errDuplicateStreamerUsername`, `roster.go:109`). Two behaviors are left unspecified:

- **Mapping.** Nothing says how a Navidrome user maps onto an existing member, or what happens on a
  collision: abort, skip, or attach the state to the existing member.
- **Reruns.** Nothing says whether a rerun after a crash is safe. Play counts are "views over" play history
  (streamer spec:149). If the import synthesizes history events, a rerun after a partial run doubles every
  count, and playlists duplicate.

Matching Navidrome user names to dubplate usernames is exactly the question the table row calls moot. The
rename gives dubplate a username, and the import must match Navidrome's names against it.

**Fold.** Change the row to "Navidrome-to-member username mapping | Slice 8 (the import)". Add four
acceptance criteria to slice 8 in streamer spec revision 3:

1. A dry-run report precedes any write.
2. An existing member is matched by username under the column's `COLLATE NOCASE`, or through an operator
   mapping list.
3. All per-user writes land in one store transaction, or are keyed so that a rerun is a no-op.
4. The import reads an extracted copy of the archive.

### D5 (low): the in-place `0003_catalog.sql` edit trips a pin the spec does not name; no data-loss path exists

**Where:** spec:145, spec:148-150.

**Finding.** No old-schema store can be silently carried forward:

- **Any old store fails loudly.** Every store records each migration's sha256 in its ledger, and the open
  refuses a mismatch: "ledger version %d records %s (sha256 %s), this binary has %s (sha256 %s)"
  (`internal/engine/store/migrate.go:237-238`). A store from the box, a workstation, or a restored
  snapshot refuses to open. It never runs against the renamed column.
- **No backup holds an old store.** Box backup was never enabled (D3), so no remote snapshot exists.

The gap is in the gate. `TestSchemaMigrationHashPins` pins 0003's hash
(`internal/engine/store/schema_test.go:295-299`), and its stated purpose is to stop exactly this edit: "an
edit to a merged file ... then fails this table". The spec does not say that the pin moves on purpose.

**Fold.** Add to spec:148-150 that the pin moves in the same commit, and that the commit message cites the
precondition evidence. Use doctor's `store` row reading `absent` (runbook: "`store` reads `absent`,
`current (v<N>)`...") as the proof, in place of an ad hoc directory listing.

### D6 (low): the rename's file list misses contract-bearing files

**Where:** spec:141-146.

**Defect.** The list omits several files and wire contracts:

- `config/roster.example.json:7,13`, the shape a stranger copies, which `TestRosterExampleFile` loads
  (`internal/auth/roster_test.go:440-445`).
- The rendered gate, `web/scripts/rendered/run.ts:3603-3630,4151-4152`.
- `internal/backup`'s and `internal/engine/store`'s test inserts.
- The roster API's wire contract: the form field `streamer_username` and the outcome codes
  `duplicate_/invalid_/missing_streamer_username` (`internal/auth/admin.go:26,37-42`).

The gate catches every omission, so the consequence is a fix round. No defect would ship.

**Fold.** Replace the list with one acceptance criterion: a grep for `streamer_username`, `StreamerUsername`,
`streamer-username`, and `streamerUsername` outside `_salvage/` and the dated docs returns nothing. Name the
wire codes in that criterion.

### D7 (low): deleting the bearer bypass deletes the only written statement of its security contract

**Where:** spec:104-105.

**Finding.** The removals cause no regression on the live serve unit:

- Production builds the strip as `serve.NewStrip("", auth.SessionCookieName, proxyConfig,
  serve.FrontExemption{})` (`cmd/dubplate-server/serve.go:324`).
- `addBearerBypass` has one caller, a test (`internal/serve/origin_test.go:127`).
- `SetFallthrough` has no production caller.

Production already uses the empty header and the empty exemption, so the strip set and the origin check
keep the same behavior.

**Defect.** The contract lives only in `internal/serve/origin.go:37-46`: "every handler behind a bypassed
pattern ... must not read the session cookie to authenticate a request. A bypassed handler that fell back to
the cookie would turn the exemption into exactly the forgery this check refuses." Slice 6 rebuilds the
bypass, and this contract is the part it must not rediscover. Spec:105 also says the rebuild is "keyed on
`netip.Addr` where it matches peers", but the origin bypass is keyed on the path (`addBearerBypass(pattern
string)`, an `http.ServeMux` pattern). The peer match belonged to `FrontExemption`.

**Fold.** Move the contract sentence into streamer spec revision 3's slice 6 (or HISTORY's "wrong to
rediscover" list), and correct spec:105 so the `/rest` exemption is path-keyed.

### D8 (low): the shadow program's on-disk state is in the archive, and no disposal names it

**Where:** spec:118-125, spec:82-83.

**Finding.** Nothing live is orphaned:

- **Box.** The shadow store under the state dir (`shadow.StoreDir`, `internal/engine/shadow/store.go:46`),
  the observe directory, the oracle state, and the review directory lived on the box. The retirement
  archived them with "the old data directories" (HISTORY 5d-ii). STATUS lists the box's remaining leaves as
  `auth`, `roster`, and `unidecode`.
- **Repo.** `corpus/analysis/adjudications.json` stays read by conformance
  (`internal/harness/conform/case.go`).
- **R2.** No shadow state was found.

**Fold.** Extend rung 17's disposal sentence (spec:82-83) to say that the archive's shadow directories go
with it. `Makefile:5`'s comment cites "shadow observe" and belongs on spec:124's comment-rewrite list.

## Over-ceremony

### O1: the removal pass's deploy, which exists only to migrate the roster row

**Where:** spec:153.

**Cost.** One box write session, with its quiet-box check, its `--check` read, and an `--apply`. That work
serves a roster file nothing reads until a later deploy: serve is disabled, and the group is not live.

**Fold.** The removal pass does not deploy. The next deploy that reads the roster carries the D2 sequence.

### O2: a roster schema-version bump for the key rename (pre-empted)

**Where:** `internal/auth/roster.go:30` (`rosterSchemaVersion = 1`).

**Cost.** A version gate and a migration branch, written for one file on a box that the D2 sequence already
handles. No released v1 file exists anywhere else, so the bump adds code that protects nothing.

**Fold.** If a plan reviewer raises the bump, refuse it, keep version 1, and record the reason in the fold
record.
