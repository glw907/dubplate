# Rung 6 spec review: consistency against the ratified documents

Target: `docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md` at `12a6941` (revision 1,
draft). Lens: every claim and citation checked against the ratified record, and every obligation
the ratified record hands to rung 6 checked against the draft. Reviewer: fresh context, Opus 5.5,
2026-09-23.

## Verdict

The draft is unusually well cited, and most of what it says holds up. Every item that the
configuration homes spec (CH) routes to "the rung 6 store" appears in the schema list or is
explicitly deferred with a reason. The 2026-09-23 errata on root relocation and per-collection
parity are carried correctly. The code and modernc citations resolve to the lines the draft
claims, apart from four that drift. Three problems stop it from going to the owner sitting as
written. First, ruling 6 (the roster file) cannot be implemented as stated. It makes
`DUBPLATE_ROSTER_FILE` a load error. That key is a required key of the magic-link identity
source today. The import depends on the key being set. The error depends on store state that
`internal/config` may not read. And the ruling reinterprets CH's wall list with no erratum line.
Second, ruling 5's reasoning is inconsistent. It lands CH's tables at rung 6 because "a table is
schema." Then it defers hook columns 1, 2, 3, and 11, which the 2.0 spec also assigns to "rung 6
schema." The hook table calls those hooks "Neutral," and neither the 2.0 spec nor this draft
records an erratum. Third, two owner-facing gaps in backup are unflagged. Ruling 2 moves backup
credentials off the wall, which CH says holds "every secret." Ruling 1's backup omits the
streamer's database, which is the only home of members' stars and playlists at 1.0. Beyond those
three, a last-admin guard is dropped at the roster move, and hook 4's invariant is narrower than
the hook itself. Minor findings cover several silent supersessions of ratified passages that need
erratum lines, and citation drift.

Counts: 1 blocker, 5 major, 18 minor. Four findings carry an OWNER FORK.

## Findings

### 1. BLOCKER. Ruling 6 cannot be implemented as written, and it supersedes CH and R5S silently

- **Draft:** lines 91-101 (ruling 6), 497-499 ("the handoff needs no secret minted and no
  template row removed"), 644 (the roster-import row), 677 (6d scope).
- **Ratified:**
  - `cmd/dubplate-server/auth.go:45-46`: the magic-link source's `Required` list opens with
    `"DUBPLATE_ROSTER_FILE"`.
  - CH:52-54: "The wall holds the owned, required, no-default keys (... the identity source and
    its correlated set, and the roster file)".
  - R5S:1552-1555: "`DUBPLATE_ROSTER_FILE` stays the only key the roster needs".
  - Draft line 181 (its own rule): "`internal/config` must never import `store`".
- **Contradiction:**
  - Ruling 6 says that after the import the key "is a load error." A registry supersession is
    unconditional. That rules out setting the key at all, so a fresh install has no roster file
    to import from, and the magic-link source's required set still demands the key.
  - Line 644 makes the error conditional instead ("file set and users populated"). A condition
    on store state cannot live in `config.Load`, because `config` may not import the store.
  - Ruling 6 also says "a template still carrying the key fails loudly," while lines 497-499 say
    no template row is removed. The estate's `env/dubplate-server.env.template:72` names the key.
  - The draft "resolves" CH's wall sentence "as the bootstrap source it was on the day CH was
    written." That is a reinterpretation of a ratified passage with no erratum line. R5S:1552
    needs the same treatment.
  - R5S:198-200 has `deploy.sh` pre-create the roster leaf, and PA:218 records that leaf. The
    draft does not say what happens to either.
- **Proposed fold:** restate ruling 6 as a two-step supersession.
  1. 6d removes `DUBPLATE_ROSTER_FILE` from `magicLinkSource.Required` and adds an optional
     one-shot import source, read only when the users table is empty. Name the check's home: the
     store-open path in `cmd/dubplate-server`, never `Load`.
  2. The key becomes an unconditional supersession, naming the `member` verbs, one release
     later or at rung 16's convergence. Name which one.

  Add erratum lines to CH:52-54 and R5S:1552-1555. Add the template edit, and the roster leaf's
  fate (retired, or kept as the import's archive), to the 6d row and to the contested-surface
  row at draft line 725. Correct lines 497-499.

### 2. MAJOR. Ruling 5 draws its line inconsistently, and the hook table mislabels S20's rung-6 schema hooks as "Neutral"

- **Draft:** lines 76-89 (ruling 5), 146-149, 536-550 (hook rows 1, 2, 3, 11).
- **Ratified:**
  - S20:363: "1. Content hash on the catalog row, rung 6 schema, computed at rung 7's gate".
  - S20:370: "2. Folded sort key on the catalog row, rung 6 schema".
  - S20:372: "3. Art blob recorded at admit, rung 6 schema and rung 9 admission".
  - S20:392: "11. ReplayGain and R128 gain on the catalog row, rung 6 schema".
  - CLAUDE.md, the 2026-09-12 refinement: "a 2.0 forward hook is named in the rung plan and
    built by the rung that first consumes it, and nothing is exported ahead of a caller."
- **Contradiction:**
  - Ruling 5 lands CH's tables because "a table is schema, not exported surface." The hook
    columns are also schema and not exported surface, yet the ruling defers them.
  - The criterion actually applied is "values a later rung computes." That criterion also fits
    `device_credentials` (first writer rung 14 or the TUI), `quota_counters` (rung 7),
    `move_intents` (rung 9 or 14), and `member_preferences`, all of which land at 6c.
  - Calling hooks 1, 2, 3, and 11 "Neutral" misstates S20. S20 assigns their schema to rung 6. A
    deferral is not neutrality, and S20 carries no erratum moving them.
- **Proposed fold:**
  - Restate the rule the recommendation actually uses. One option: "a table or column lands with
    its first writer; CH's routed tables are the owner's explicit exception, because CH:637
    routes them here by name". Alternatively, land hook columns 1, 2, 3, and 11 as nullable
    columns at 6c on the same logic as the tables.
  - Relabel the four hook rows "Deferred to rung N's migration (S20 erratum)".
  - Add a dated erratum to S20 hooks 1, 2, 3, and 11 if the owner rules for deferral.
  - This is an OWNER FORK already inside ruling 5, but the sitting needs the corrected framing.

### 3. MAJOR, OWNER FORK. Ruling 2 moves backup credentials off the wall and out of the secrets flow

- **Draft:** lines 53-60.
- **Ratified:**
  - CH:52-57: "The wall holds ... every secret, and every integration class whole ... backup
    targets and schedule".
  - CLAUDE.md, Facts: "`env/dubplate.env.template` is a secrets MANIFEST rendered by deploy.sh
    from the age store".
  - CLAUDE.md, Safety rails: "Never edit the box by hand and call it done".
  - BS:218-220: "backup targets (the rclone remote pair; the server starts without one and
    `doctor` says so plainly)".
- **Gap:**
  - "Credentials in rclone's own configuration and never in a `DUBPLATE_*` key" puts a secret in
    a file that neither the wall, the age-store render, nor `doctor`'s required-when-enabled
    check covers.
  - Nothing names who writes `rclone.conf` on the box, or where. A stranger would hand-edit a
    file outside `deploy.sh`, which cuts against the product goal of stranger setup.
  - rclone reads remotes from `RCLONE_CONFIG_<REMOTE>_*` environment variables. That path would
    keep the credential in the rendered env and the age store without a `DUBPLATE_*` name, but
    the draft does not weigh it.
- **Proposed fold:** name the credential's home explicitly, with the options for the owner:
  1. `RCLONE_CONFIG_*` names rendered by `deploy.sh` from the age store, declared through the
     class so `doctor` can report presence.
  2. An rclone config file path as a wall key, with the file rendered by `deploy.sh`.
  3. The draft's position, with a CH erratum narrowing "every secret."

  State how `doctor` proves the remote is usable when the flag is on.

### 4. MAJOR, OWNER FORK. Ruling 1's backup omits the streamer's state, which at 1.0 is the only home of stars and playlists

- **Draft:** lines 44-51. The draft cites `_salvage/internal/backup/backup.go:98-104` as syncing
  "both."
- **Ratified:**
  - `_salvage/internal/backup/backup.go:91-96` snapshots `library.db` and `navidrome.db` before
    the two syncs, so the salvage backup covers three things, not two.
  - CH:461-462: Navidrome "purges them with its own per-user stars and playlist entries, which
    at 1.0 have no other home."
  - OS:77-80 puts backup and restore in the server's purview.
- **Gap:** after a restore under the Navidrome driver, every member's stars, playlists, and play
  history are gone. The draft neither backs up the streamer's database nor states the loss.
  Whether a backup covers driver-owned state is a product decision about what an adopter can
  recover.
- **Proposed fold:** add to ruling 1: "the streamer driver may declare state it wants in the
  backup (rung 12 lands the declaration; 6b's backup runs a declared-paths hook, empty until
  then)." Alternatively, state the 1.0 loss in the ruling and the runbook. Either way, correct
  the salvage citation to 91-104.

### 5. MAJOR. The last-admin lockout guard is not carried across the roster's move into the store

- **Draft:** lines 160 ("No last-admin lockout guard. It is routed to rung 5's pass 5c."),
  513-517, 677 (6d exit gate).
- **Ratified:**
  - CH:668-669: "Rung 5's plan: ... a last-admin lockout guard, atomic against concurrent
    self-demotion."
  - `docs/STATUS.md:18` carries it to 5c.
  - R5S:328-329: the write-time checks run "inside the same lock the write holds, so two
    simultaneous adds cannot both pass it."
- **Gap:** 5c builds the guard against the file store's mutex. 6d replaces that write path with
  a SQLite transaction. The draft carries forward the admin-flag channel rule (line 515) but not
  the guard or its atomicity, and 6d's exit gate has no test for either. A guard that was atomic
  under a mutex becomes a read-then-write race unless it is re-established inside the
  `IMMEDIATE` transaction.
- **Proposed fold:** add to "Identity in the store": "the last-admin guard and the
  streamer-principal collision check run inside the write transaction". Add to 6d's exit gate: "a
  concurrent self-demotion of the last two admins leaves one admin."

### 6. MAJOR. Hook 4 is claimed as landed, but invariant 3 is narrower than the hook, and `settings_audit` escapes it

- **Draft:** lines 271-272 (invariant 3 keys on the `_events` name suffix), 296-298
  (`settings_audit`: "actor, channel, key, and old and new values"), 541 (hook 4 "Lands at 6a").
- **Ratified:** S20:374-375: "Any event-shaped table carries a timestamp and an event kind, never
  a bare counter."
- **Gap:**
  - The suffix test passes any event-shaped table not named `*_events`: `settings_audit`,
    `move_intents`, and the job records.
  - The draft's `settings_audit` columns name neither a timestamp nor a kind.
  - The roster audit record R5S:1139-1144 requires ("the only trace a roster change leaves")
    has no stated home after 6d either.
- **Proposed fold:** give invariant 3 a declared marker rather than a name suffix. One option is
  that every table ending `_audit`, `_events`, or `_journal` must carry the pair, with the list
  fixed in the invariant test. Add `at` and `kind` to `settings_audit`. State where member-change
  audit rows live after 6d, whether a `member_audit` table or the same audit table with a kind.

### 7. MINOR. The roster move is attributed to the wrong R5S passage, and the ruled interface is reopened

- **Draft:** lines 136, 682-683 (cites R5S:1807-1811 for "the roster's move"), 514-515 ("the plan
  decides whether 6d introduces the interface R5S describes or replaces the implementation in
  place").
- **Ratified:**
  - R5S:1809-1811 names two obligations only, the epoch re-key and the surrogate id.
  - The roster's move is Geoff's fork-1 ruling at R5S:195-196 ("Rung 6 later moves the store
    behind the same interface without touching a call site"), in R1 at R5S:289-292, and in
    PA:209.
  - R5S:287: "The interface is the whole ruling."
- **Proposed fold:** cite R5S:194-196, R5S:287-292, and PA:209. Replace the open question with
  the ruled constraint: "no call site changes; `internal/auth/roster.go:113-116` already frames
  the substitution as one constructor."

### 8. MINOR. The wall keeps tunables that CH homes in the store, and the draft cites the wrong CH lines for the exception

- **Draft:** lines 105-110 (ruling 7), 234-235 (busy timeout "homed on the wall"), 492-493.
- **Ratified:**
  - CH:66: "The store holds every tunable that ships with a default".
  - The draft's quoted principle ("what the process needs before it can start") is CH:74-75,
    not CH:52-66.
- **Gap:** the serve keys and the busy timeout are defensible under CH:74-75. The shadow keys
  rest on the shadow unit's current `ReadWritePaths` grant, which is a deployment choice rather
  than a consumer that "cannot see the store" (CH:62-63). The draft adds a third wall category
  without a CH erratum.
- **Proposed fold:** cite CH:74-75. Record one CH erratum line: "keys read before the store
  opens, or by a unit not granted the store leaf, stay on the wall". Alternatively, give the
  shadow keys the plainer reason that the shadow retires at launch.

### 9. MINOR. The draft re-bands a ratified ladder entry without an erratum line

- **Draft:** lines 19-20, 681-683.
- **Ratified:** BS:656: "6. **Ops seam and store** (M)". BS:955 says L rungs "split at their own
  planning," but it does not say that planning re-bands them.
- **Proposed fold:** add a dated erratum line to the build spec's ladder entry 6 (M to L, with
  reasons), or have the conductor record it in STATUS as a ladder note. The 17M total should
  also be reconciled against the spine phase's combined ceiling (PA:235).

### 10. MINOR. Rung 6 touches the box, which the build spec says it never does

- **Draft:** lines 137, 742-743 ("Rung 6 touches the box only at 6d's deployment").
- **Ratified:** BS:934-935: "Rungs 1 through 3, 5 through 9, and 11 through 15 touch the box not
  at all".
- **Proposed fold:** either say 6d edits the unit, `deploy.sh`, and the check in the repo while
  the deploy rides the next ruled deploy, or add an erratum line to "Box safety during the
  build." Rung 5 appears to have the same unrecorded delta, and one erratum could cover both.

### 11. MINOR. The backup integration class drops the schedule and the rclone pair without saying so

- **Draft:** lines 53-60, 150-151.
- **Ratified:** CH:56 lists the class as "backup targets and schedule." BS:218-219 describes it as
  "the rclone remote pair."
- **Gap:** ruling 2 declares a flag and one remote. The schedule key that CH puts in the class
  is declared by no rung. Rung 16's ladder entry (BS:844-851) does not name the backup timer, so
  the draft's routing to rung 16 lands nowhere.
- **Proposed fold:** either declare the schedule key in 6b's class, with the timer to follow at
  rung 16, or state the deferral and add it to rung 16's carried list. Say that one remote
  replaces BS's "pair," and why.

### 12. MINOR. The backup monitoring ping is unassigned

- **Draft:** section "Backup and restore" (lines 406-436) is silent on it.
- **Ratified:** `internal/config/keys.go:93-100` reserves `DUBPLATE_PING_BACKUP_URL` in the
  "monitoring" class for rung 9. The salvage backup pinged on success
  (`_salvage/internal/backup/backup.go:106`). CLAUDE.md notes that failure pings are POST.
- **Proposed fold:** add one line: "backup's success and failure pings arrive with rung 9's
  monitoring seam; 6b's job exposes the terminal state the seam will read."

### 13. MINOR. "Progress is polled, never pushed" forecloses what the organizer spec permits

- **Draft:** line 352.
- **Ratified:** OS:243-245: "Progress is polled ... a local implementation may layer push on the
  same abstraction."
- **Proposed fold:** "Progress is polled; the interface leaves room for a local push layer
  (OS:244-245)."

### 14. MINOR. Invariant 6 conflicts with S20's alternate keys, and its citation drifts

- **Draft:** lines 277-278, 538.
- **Ratified:** S20:166-167: "Entity ids are opaque surrogates ... The release MBID and the
  content hash are indexed alternate keys, never the id". The draft cites S20:163-165, which is
  the projector-determinism sentence.
- **Gap:** a wire type does carry MBIDs and may carry the audio key, and neither is random.
  "Every public identifier column that a wire type may carry" would fail both. Hook row 1 then
  applies invariant 6 to the audio key "where it applies." The catalog-only test (line 264) also
  cannot tell which columns a wire type carries.
- **Proposed fold:** scope invariant 6 to entity ids (primary keys and `*_id` columns
  referencing them), with alternate keys named as exempt. Correct the citation to S20:166-167.
  Drop the invariant 6 clause from hook row 1.

### 15. MINOR. The snapshot mechanism is a third option that S20 hook 10 does not name

- **Draft:** lines 408-415, 571.
- **Ratified:** S20:390-391: "the backup job uses the SQLite backup API or checkpoints before it
  copies."
- **Proposed fold:** `VACUUM INTO` meets the hook's intent. Say so in one sentence so hook 10's
  "Lands at 6b" does not read as a deviation.

### 16. MINOR. The collection row lacks the version column CH's `if_version` check needs

- **Draft:** lines 287-292.
- **Ratified:** CH:256-257: "Shape edits carry an `if_version` check, because a template write
  triggers a re-layout of files".
- **Proposed fold:** add a row version to `collections`, or defer it to rung 14 by name under
  ruling 5.

### 17. MINOR. The free-space floor is not applied to the store's own large writes

- **Draft:** lines 251-252 (pre-migration snapshot), 411-414 (`VACUUM INTO` staging).
- **Ratified:** CH:494-496: "A hard free-space floor is checked before every write, because a full
  disk wedges the store's write-ahead log and takes the server down."
- **Gap:** both snapshots write a whole copy of the database into the state directory.
- **Proposed fold:** add a free-space check before each snapshot, with a report row in the
  state table at line 639 or 641.

### 18. MINOR. The settings write path omits CH's URL validation rule

- **Draft:** lines 468-471.
- **Ratified:** CH:132-133: "Any URL-valued setting is validated against loopback, link-local, and
  private ranges before it is stored."
- **Proposed fold:** state that no URL-valued key is store-homed at rung 6, and that the write
  path's validator hook is where rung 14's check (CH:659) attaches.

### 19. MINOR, OWNER FORK. Two env-seed rows take behavior decisions that CH leaves open

- **Draft:** lines 459-463.
- **Ratified:**
  - CH:79-86 defines only three present-in-env cases.
  - CH:85-86: "an operator who rotates a value in the age store never has it silently ignored."
- **Gap:**
  - "`env-seed` | absent": an operator who deletes a key to return to the default keeps the old
    seeded value, with only a `doctor` note. That is close to the silent outcome CH forbids.
  - "no row | absent → written at its default": a row pinned at an old default ignores a
    researched default improved in a later release.
- **Proposed fold:** put both rows to the owner with a recommendation. One option: an absent env
  value reverts an `env-seed` row to `default`, logged. Another: `default` rows store no value
  and read the registry's current default.

### 20. MINOR. The draft overstates modernc's validation guarantee for the settings that matter

- **Draft:** lines 222-227, 564 (prior-art "Quality" cell).
- **Ratified (source):** `modernc.org/sqlite@v1.59.0/sqlite.go:305-306`: "_pragma values are the
  one exception: they are executed verbatim and cannot be checked here, so a malformed _pragma can
  still fail partway."
- **Gap:** the draft puts journal mode, foreign keys, `synchronous`, and the busy timeout in
  `_pragma`, the one path the cited validation does not cover.
- **Proposed fold:** qualify the prior-art cell. Add a proof to 6a's exit gate that each pragma
  reads back as set on a fresh pooled connection.

### 21. MINOR. The prior-art list does not account for every entry CLAUDE.md names

- **Draft:** lines 555-573.
- **Ratified:** CLAUDE.md (Build for 2.0): "a rung plan names which entries it consulted and
  adopts or declines each with a reason" over the list that includes OpenSubsonic, Navidrome and
  gonic, taglib, mewkiz/flac, Picard, and chromaprint. ROADMAP:93 names modernc and goose for
  rung 6, and the draft covers both.
- **Gap:** the other list entries go unmentioned. Navidrome's own SQLite schema and migration
  practice is relevant, as behavior-only prior art under GPL, for users, libraries, and
  per-user tables.
- **Proposed fold:** add one line: the remaining entries do not apply to rung 6. Add Navidrome's
  schema as a behavior-only reference for `users` and `collections`, or decline it with a reason.

### 22. MINOR. Apply mode and the sensitive flag move from the table to the registry without a note

- **Draft:** lines 134-135, 445-448.
- **Ratified:**
  - CH:638: "the settings table with source, apply mode, the sensitive flag, and the audit
    table".
  - S20:401 says the same.
  - BS:39-40 puts home and apply mode on the registry.
- **Proposed fold:** one sentence stating that apply mode and the sensitive flag are
  per-key registry metadata. Say that the table carries only per-row state, which reads CH:638
  together with the BS erratum.

### 23. MINOR. The HTTP projection is described with one mount, and CH requires two

- **Draft:** lines 141-143 ("mounts it under `/api/v1`"), 392-396.
- **Ratified:** CH:107-110: "The admin API is served on two mounts, one cookie-authenticated ...
  and one bearer-authenticated".
- **Proposed fold:** "rung 14 mounts it twice (CH:107-110); the authorizer seam takes the mount's
  channel so one projection serves both."

### 24. MINOR. Two ownership questions are left unassigned: the runbook, and the admin email beside the admin flag

- **Draft:** line 436 ("The runbook states this"), and ruling 6.
- **Ratified:**
  - CH:661-662 routes "state the restore-size expectation in the runbook" to rung 16.
  - `internal/config/keys.go:24` keeps `DUBPLATE_ADMIN_EMAIL` as an owned wall row.
  - BS:833 names "AdminEmail" in rung 14's authorization model.
  - R5S:1548-1551 derives admin contact from the roster's admin entry.
- **Proposed fold:** name rung 16 as the runbook's author, or name 6b if a restore runbook ships
  with the verb. Add one line to ruling 6 stating which source is authoritative for "admin"
  after the move, the wall email or `users.admin`, and route any reconciliation to rung 14.

## Citation spot-check

More than eighty citations were checked. All but the four listed after this list resolve to
what the draft claims:

- CH:52-53, CH:58-66, CH:66-86, CH:67-70, CH:79-87, CH:88-92, CH:122-125, CH:128-132,
  CH:161, CH:165-168, CH:173-181, CH:195, CH:357-359, CH:410-414, CH:451-458, CH:488-500,
  CH:519-521, CH:524-526, CH:535-538, CH:543-548, CH:637-645, CH:657-660, CH:662-664.
- BS:28-40, BS:131-141, BS:192-194, BS:236-248, BS:375, BS:563-567, BS:656, BS:662-664,
  BS:668-672, BS:760-764, BS:831-837, BS:844-851.
- OS:77-81, OS:230-236, OS:241-248, OS:273-277, OS:279-285, OS:338-352, OS:412-413,
  OS:425-433.
- S20:374-378, S20:383-391, S20:435.
- R5S:333-338, R5S:533-539, R5S:1141-1143, R5S:1153-1158, R5S:1687-1690.
- EVO:356-366, EVO:489-494, EVO:575-578, EVO:600-602. RE:116. PA:173, PA:197.
- `keys.go:60-71`, `main.go:13-14`, `identity.go:1-8` and `115-119`, `roster.go:117` and
  `286`, `token.go:118`, `session.go:132` and `206`, `limit.go:62`, `auth/store.go:12-40`,
  `shadow/store.go:39-49`, `atomicfile.go:45-55`, `auth.go:99-196` (every row of the seed
  table), `serve.go:35-59`, `shadow.go:73-88`, `backup.go:22-34`.
- modernc `sqlite.go:385-391` and `425-447`, `conn.go:1208-1219`, and the
  `LIKE_DOESNT_MATCH_BLOBS` build flag.
- HISTORY:578-590, 1735, 2220, 2627, 2961. Commits `323a817`, `e0a3aa8`, and `fd122d9`.
- The goose v3.28.0 figures (nineteen direct requirements, nine of them drivers) match the
  module cache.

Four citations drift:

- CH:52-66 for "what the process needs before it can start" is CH:74-75 (finding 8).
- S20:163-165 is S20:166-167 (finding 14).
- R5S:1807-1811 does not contain the roster move (finding 7).
- `sqlite.go:290-305` omits the `_pragma` exception at 305-306 (finding 20).

A fifth, `backup.go:98-104`, resolves, but the draft characterizes it incompletely (finding 4).

S20:435 says goose "direct-requires ten database drivers," where the draft measures nine at
v3.28.0. The draft's figure is the current one, and nothing needs folding.

## Obligations checked and found carried

These obligations need no action.

- Every CH:637-644 rung 6 item is carried: the settings and audit tables, overrides and
  preferences, and the collection-row columns. So are device credentials, the two quota
  counters, the move-intent journal, foreign keys in the DSN, and the share rows.
- CH's two 2026-09-23 errata are carried: root allocation with no legacy column, and the rung 12
  conditional routed through rung 10.
- CH:410-414 restore size is carried. CH:657-660 re-authentication is deferred to rung 14 with a
  reason.
- BS ladder entry 6 is carried in full: the interface, the job model, cancellation, the
  availability mechanism, modernc, forward migrations, U3, U4, backup behind the flag, and the
  contract suite at its honest scope. The non-UTF-8 policy defers to rung 7, as EVO:362-366
  directs.
- R5S R4 (the epoch re-key), R3 (the surrogate id at the one resolution point), R17
  (`internal/ops` imports only `identity`), and R18 (the records stay files) are all carried.
- The PA state-directory row ("supersedes neither") is carried.
- The proposed contested-surface rows do not conflict with PA:177-221. The `keys.go` row mirrors
  PA:212's retirement form. The `config/registry.go` row follows PA:202's ownership.
