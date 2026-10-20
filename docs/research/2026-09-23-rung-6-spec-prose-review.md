# Rung 6 spec prose review

Target: `docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`, revision 2.1 at `9155c8b` (1,259 lines). Register: developer docs, Google Developer Documentation Style Guide (`~/.claude/docs/voice/technical-doc-go.md`), read as an agent-facing and owner-facing spec. No corpus manifest was named, so the grade rests on the register alone. `tellgrader --register docs` found no deterministic findings: 0 tells per 1,000 words, 16 tricolons, cadence CV 1.40. It returned no `measures` object because the profile did not resolve, so the measurement table below is my own count.

## Verdict

The spec's quotes are sound. I checked these against their sources and each matches verbatim:
- sqlite.org: howtocorrupt sections 1.2 and 2.5; wal.html sections 1 and 5; the pragma `synchronous` and `foreign_keys` text; both lang_vacuum passages and its empty-target rule; the lang_transaction section 3 list, `SQLITE_INTERRUPT` included.
- rclone's manual: both env-naming sentences and the `--max-delete` text.
- The man pages: flock(2) and systemd.service(5) on `Type=simple`.
- modernc v1.59.0: the DSN validation at `sqlite.go:305-391`, SQLite 3.53.4, the FTS5, SNAPSHOT and `LIKE_DOESNT_MATCH_BLOBS` build flags, and the backup API at `conn.go:1208`.
- ncruces: wasm2go, with no wazero in its requirements.
- About 25 repo citations at HEAD, plus sampled CH, BS, S20 and R5S lines.

The tier 1 defects are four kinds:
- Omissions from passages the spec quotes, and those omissions change the design: `--backup-dir` needs server-side move or copy, rclone only reads options for the immediate backend, and a remote name has a restricted character set.
- One wrong claim about what causes a failure.
- One wrong reason for declining golang-migrate.
- A ceiling calculation that counts plan prep twice.

Tier 2 is where two fold rounds left seams. Store creation contradicts "only serve migrates", the seed rule's ordering contradicts two consumers that run before it, and ruling 7 removes the rollback path. The Owner rulings are mostly decidable, but rulings 1 and 2 carry implementation detail and bundled questions. The prose reads as a plausible human author's Google-register spec. Its tells are density (a 302-word ruling, sentences up to 89 words, semicolon chains), review-process narrative inside normative text, and undefined R- and U- labels.

## Measurements

The scanner's measures were unavailable because the profile did not resolve. These are my own counts over prose outside tables.

| Measure | Value |
| --- | --- |
| sentences | 635 per tellgrader (whole file); 646 by my count outside tables |
| average sentence length | 17.4 words (prose only); 22.3 by tellgrader's whole-file ratio |
| longest sentence | 89 words, line 313 (the three test seams) |
| short-sentence share (8 words or fewer, my count) | 28% |
| paragraphs and list items | 244 |
| disproportionate paragraphs | ruling 1 (line 52, 302 words), the opening brief (line 10, 252 words), ruling 2 (line 73, 205 words), ruling 5 (line 106, 198 words), the guards paragraph (line 646, 188 words) |

## Findings

### Tier 1: invented, unsourced, or misquoted specifics

**1. Lines 347-349, the read pool.**
- **Defect:** The sentence gives the wrong cause for the failure: "A write there would fail at once, because the busy handler does not run on a deferred read-to-write upgrade."
- **Evidence:** With `_query_only=1`, SQLite refuses the write with a read-only error before it takes any lock. 6a criterion 5 (line 982) expects exactly "a read-only error". The busy-handler behavior explains why the pool needs `query_only`; it does not explain why the write fails.
- **Rewrite:** "The **read pool** opens with `_query_only=1`, so a write there fails at once with a read-only error, and no caller writes outside the write method. Without it, a write on a read connection would upgrade a deferred transaction. SQLite skips the busy handler on that upgrade (verified; see "Behaviors a task relies on"), so the write would return `SQLITE_BUSY` under contention instead of waiting."

**2. Lines 640-642 and ruling 10 (lines 156-162), `--backup-dir`.**
- **Defect:** The spec paraphrases half of rclone's rule and drops the half that limits universality.
- **Evidence:** The manual says: "The remote in use must support server-side move or copy and you must use the same remote as the destination of the sync. The backup directory must not overlap the destination directory without it being excluded by a filter rule." The spec keeps the same-remote and no-overlap parts. Ruling 10's universality argument names `local`, `sftp`, and `webdav`, but the layout works only on backends with server-side move or copy.
- **Rewrite:** Quote the whole sentence at line 641. Then add: "`doctor` reports a remote whose backend lacks server-side move and copy, and 6b-i's plan quotes rclone's per-backend feature table." If that is not done this revision, add a "Not verified" line.

**3. Ruling 5 (lines 109-112), remote name.**
- **Defect:** The spec quotes the manual's env-naming passage but leaves out its constraint on remote names.
- **Evidence:** The same passage says: "the remote's name must be convertible into a valid environment variable name, so it can only contain letters, digits, or the _ (underscore) character."
- **Rewrite:** Add after the second quote: "So a remote name holds only letters, digits, and underscores, and `doctor` refuses a `DUBPLATE_BACKUP_REMOTE` whose name does not."

**4. Line 878, golang-migrate verdict.**
- **Defect:** "Its SQLite driver ... pulls cgo `mattn/go-sqlite3`" is wrong at the package level.
- **Evidence:** In v4.20.1, `database/sqlite/sqlite.go` imports only `_ "modernc.org/sqlite"`. mattn is imported only by `database/sqlite3`. The module's `go.mod` requires both, but graph pruning keeps mattn out of a consumer's build. The goose row applies that same pruning argument. The `modernc v1.18.1` pin is accurate.
- **Rewrite:** "**Decline.** Its modernc driver is tested against modernc v1.18.1, 41 minor versions behind the one this rung adopts, and it keeps its own version table."

**5. Lines 1143-1147, ceiling rate.**
- **Defect:** The rate of "about 0.55M per task including fix rounds" is taken from STATUS's "Pass spend about 5.6M of 7.5M" (STATUS:24) divided by ten tasks. The table then adds plan prep and Task 0 on top of that rate.
- **Evidence:** HISTORY's pass spends include plan preparation and Task 0 (HISTORY:1735, 2220, 2627), so the same prep is probably counted twice. The prep sample (0.83, 0.87, 0.87) also leaves out 1b at 1.07M (HISTORY:2961) and 4b at 1.06M (HISTORY:3136). 1b is the spec's own named analog for a config-contract pass like 6c.
- **Rewrite:** Derive the per-task rate from 5b-ii's chain spend alone and show the division. Alternatively, state that 0.55 includes prep and drop the separate 0.85 term. Either way, cite the prep range as "0.83M to 1.07M across 5a, 8a, 8b, 1b, and 4b; 1.07M at 1b, the config-contract analog".

**6. Lines 395-397, flock(2) quote.**
- **Defect:** The quote ends with a period mid-sentence.
- **Evidence:** The man page continues: "may be granted, with the result that the conversion either blocks, or fails if LOCK_NB was specified."
- **Rewrite:** End the quote with "...may be granted, with the result that the conversion either blocks, or fails if LOCK_NB was specified."

**7. Lines 411-412, `TimeoutStartSec=`.**
- **Defect:** "bounds only a daemon that must 'signal start-up completion'" says more than the man page does.
- **Evidence:** The entry also covers `Type=oneshot`, which it names as an exception, and it bounds `ExecStartPre=` commands. The word "only" is the spec's own.
- **Rewrite:** "`TimeoutStartSec=` fails a service that 'does not signal start-up completion within the configured time' (systemd.service(5)). Under `Type=simple` the unit counts as started at fork, so no start timeout bounds `serve`'s wait inside its main process."

**8. Header promise (lines 40-41) against unquoted mechanism claims.**
- **Defect:** The header says every mechanism is quoted or listed as unverified. These claims are neither:
  - Lines 290-291: SQLite creates files at 0644 and gives `-wal` and `-shm` the main file's mode.
  - Line 353 and table row 923: the VACUUM INTO refusals. The table says "quoted above", but they are not quoted anywhere.
  - Lines 462-463: "foreign key mismatch".
  - Line 712: `data_version` "cannot serve".
  - Line 816: Navidrome "creates an account for any unknown proxy user".
  - Row 921: the busy-handler quote names no source.
- **Evidence:** Sources exist for several:
  - lang_vacuum: "A VACUUM will fail if there is an open transaction on the database connection."
  - pragma.html: "It is only meaningful to compare the 'PRAGMA data_version' values returned by the same database connection at two different points in time." The spec's claim holds, but it is unquoted.
  - c3ref/busy_handler.html: its worked scenario is the rollback journal, so the WAL claim rests on verification, not on the quote.
- **Rewrite:** Add the source and quote at each site. Otherwise, add a "Not verified" line for the file modes, FK mismatch, and Navidrome's auto-create, and change row 923 to "the transaction refusal quoted from lang_vacuum.html; the `query_only` refusal verified".

**9. Lines 368-370, the WAL truncation quote.**
- **Defect:** The quote leaves out a clause that bears directly on the spec's own setting.
- **Evidence:** wal.html section 6 says: "The checkpoint does not normally truncate the WAL file (unless the journal_size_limit pragma is set)." This spec sets `journal_size_limit`.
- **Rewrite:** "A checkpoint truncates the WAL only down to `journal_size_limit` (sqlite.org/wal.html section 6), so every snapshot ends with `wal_checkpoint(TRUNCATE)` to return it to zero bytes."

### Tier 2: cross-section contradictions

**10. Store creation against "only serve migrates".**
- **Location:** Lines 12, 336, and 394 say only `serve` migrates, and line 336 says serve mode is "the only one that migrates or creates a store". Against that, line 330 and the table at line 940 say the write mode "creates an absent store ... in the `member` verbs", and ruling 1 runs `member` verbs under the shared lock.
- **Defect:** Creating a store applies the migration set, so as written a `member` verb migrates under a shared lock. That is the fold record's convergent defect 2 ("Migration under a shared lock from any verb").
- **Rewrite:** Pick one design:
  - Option (a): "The write mode never creates a store. A `member` verb against an absent store refuses and names 'start serve first'."
  - Option (b): "A `member` verb that finds no store takes the lock exclusive, applies the embedded set to an empty file, and releases. With no data there is nothing to snapshot, so this is creation, not migration." Then change line 336 to "the only mode that migrates an existing store".

**11. Line 249 against the 6b deliverables.**
- **Location:** Line 249 says "No production open of the store before 6d except `doctor` and `check`'s inspect open." The deliverables table ships `backup` (6b-i, line 206) and `restore` (6b-ii, line 208), and both open the store.
- **Rewrite:** "Nothing creates or migrates the store in production before 6d. `doctor` and `check` open it read-only. `backup` refuses an absent store, and `restore` can create one only from a snapshot, which cannot exist before 6d."

**12. Lines 649-650 against line 1255.**
- **Location:** Lines 649-650 state that with `--max-delete 0` "the remote is never archived while the catalog knows no track". Line 1255 lists as unverified "whether rclone counts a file moved into `--backup-dir` against `--max-delete`".
- **Defect:** The guard's claim depends on the unverified behavior.
- **Rewrite:** "...rclone runs with `--max-delete 0`. Whether rclone counts a move into `--backup-dir` against that limit is unverified; 6b-i criterion 13 settles it against the real binary. If it does not count, the job instead aborts a zero-track run whose remote `library/` is non-empty."

**13. Ruling 7 against the rollback path (lines 445-447, 6b-ii criterion 10).**
- **Defect:** Ruling 7 deletes `pre-migrate-v<N>.db` after the first successful backup that postdates it. Every remote snapshot after the upgrade is newer than the old binary, so step 2 refuses it. The one-step rollback therefore survives only until the next backup, which also cuts against ruling 7's own reason (catching a semantically wrong migration).
- **Rewrite:** For the rollback paragraph: "While the local pre-migration snapshot survives ruling 7, the way back is that binary's `restore --from-file`. After a backup has deleted it, the way back is `restore --older --snapshot <stamp>` of a pre-upgrade remote snapshot, while ruling 10 still keeps one." Add the same sentence to ruling 7's "What it changes" line.

**14. Seed-rule ordering.**
- **Location:** Line 725 says the seed rule runs "before any consumer reads a setting". Line 838 orders serve's start as open and migration, then recovery, then the seed.
- **Defect:** Ruling 7 prunes local snapshots "at the next migration" using `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`, and lines 589-591 prune job records "at every write-mode open" using `DUBPLATE_JOB_RETENTION`. Both consumers read a store-homed key before the seed runs. `internal/engine/store` also cannot import `settings` (line 262).
- **Rewrite:** Either move both prunes after the seed step in line 839's order, or add: "Pruning at open reads the stored row or the registry default. A changed env value reaches it at the next start."

**15. Ruling 5 against ruling 14.**
- **Location:** Ruling 5 says credentials live "in rclone's own `RCLONE_CONFIG_<NAME>_*` environment form", and doctor checks only `<NAME>_TYPE` (line 112). Ruling 14 recommends a `crypt` remote.
- **Evidence:** rclone's manual says: "you can only set the options of the immediate backend, so RCLONE_CONFIG_MYS3CRYPT_ACCESS_KEY_ID has no effect, if myS3Crypt is a crypt remote based on an S3 remote." For a crypt remote, the credentials sit under the wrapped remote's name, and doctor's check misses a missing wrapped remote.
- **Rewrite:** Ruling 5: "...credentials in rclone's own `RCLONE_CONFIG_<NAME>_*` form, for the named remote and any remote it wraps." Line 112: "`doctor` checks that the named remote's `_TYPE` is set and, for a `crypt` remote, that the remote it wraps is defined too."

**16. Line 47 against ruling 2.**
- **Defect:** Line 47 says all fourteen questions "change what an adopter sees, configures, or is asked for". Ruling 2 (line 85) says "What it changes: nothing an adopter sees."
- **Rewrite:** "Fourteen yes-or-no questions, ordered by how much build each answer changes. Thirteen change what an adopter sees, configures, or is asked for; ruling 2 changes only when schema lands."

**17. The [CH] index (lines 1244-1245).**
- **Defect:** The index names only "job retention as a store-homed key". Line 752, under a [CH] heading, re-homes five constants, and rulings 7 and 10 say "re-homed at 6c" without the mark.
- **Rewrite:** "...and the five 6b-i constants re-homed as store-homed keys."

**18. Hook table row 4 (line 853).**
- **Defect:** The row says "`settings_audit` (6c) is its first table". The 6a job-records table has a `kind` column (lines 566 and 591), so invariant 3 (line 464) binds it first.
- **Rewrite:** "**The rule lands at 6a** as invariant 3, binding the job-records table first; `settings_audit` (6c) is its first event table."

**19. Line 1195 against the 6c heading (line 1078) and contested row 1172.**
- **Defect:** Line 1195 runs 6a-6c as lane 2 unconditionally. The 6c heading and the contested row make 6c conditional.
- **Rewrite:** "Passes 6a and 6b run as lane 2 beside 5c and 5d. 6c joins lane 2 only if 5c and 5d leave `internal/config` untouched; otherwise it waits for 5d."

**20. Restore interface (lines 671-675).**
- **Defect:** The flag list leaves out `--discard-current`, which step 3 (line 684) and the states table (line 944) use.
- **Rewrite:** Add "`--discard-current` moves an unopenable current store aside" to the flag list.

**21. Ruling 9's index.**
- **Defect:** The address index lands with `users` at 6a (lines 488-489), but its proof is 6c criterion 3 (line 1083). No 6a criterion exercises it.
- **Rewrite:** Move the criterion to 6a, or state that the index lands at 6c.

**22. Import table (lines 796-803).**
- **Defect:** There is no row for two entries sharing one `streamer_username`. The handle is unique (line 490), so a duplicate within the file becomes a constraint failure at commit, not a named refusal, and `--dry-run` cannot "report every refusal at once".
- **Rewrite:** Add the row "two entries with one handle | refused, naming both entries". Change line 805 to "nine states, eight refusals" and 6d criterion 2 to "eight".

**23. Required handle against the rung 5 write path.**
- **Defect:** 6d makes `streamer_username` required (line 490), but rung 5's member record makes it optional (`internal/auth/roster.go:109`, `omitempty`). The spec never says 6d makes the page and the verb require it, so a page add after 6d fails at the constraint.
- **Rewrite:** In "Identity in the store", add: "From 6d, the add-member page and verb require `streamer_username` and refuse a duplicate by name."

### Tier 3: owner-facing clarity

**24. Ruling 1 (lines 52-71, 302 words).**
- **Defect:** Half the ruling is implementation the owner cannot rule on: `IMMEDIATE`, the shared lock, the `Required` list line, `AddSuperseded`. Its "What it changes" line is also inaccurate. It says "A member added through the page ... in that window merges", but during the outage no roster admin can log in to reach the page.
- **Rewrite:** Keep the question, the recommendation, one reason, and "What it changes". Move the mechanics to "Identity in the store". Change the last sentence to: "A member added through the CLI verb in that window merges rather than blocking the import."

**25. Ruling 2.**
- **Defect:** It bundles three decisions into one yes/no, and the "no" branch covers only the hook columns. "As the owner's explicit exception" assumes a ruling that line 7 says is not yet given.
- **Rewrite:** "**2. Do S20's hook columns 1, 2, 3, and 11 land with the rungs that compute them, rather than as nullable columns at 6c?**" Move the 6a skeleton and 6c table timing to "The schema rung 6 lands" as a technical decision, and write "because CH routes them to rung 6 by name".

**26. Ruling 5's "What it changes" line.**
- **Defect:** It leaves out that rclone credential rows reach the template only at rung 16 (lines 236-237, erratum 14). Until then an adopter cannot configure backup through `deploy.sh`.
- **Rewrite:** Add: "The template rows for rclone's names arrive at rung 16; until then the runbook sets them by hand."

**27. Ruling 7's question.**
- **Defect:** The question carries three conditions, uses the jargon "per from-version", and ends with the vague "newest few".
- **Rewrite:** "**7. Before each upgrade, does the store keep a copy of itself, one per version, until a backup has safely captured the upgraded store?** With backup off, it keeps the newest `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` copies."

**28. Undefined labels.**
- **Location:** R4, R18, and R35 (lines 833-834, 1224); U3 and U4 (headings at lines 523 and 534; "U4's attribute bag" at line 470); "the #604 class" (line 468); "the reviews' scenarios" (line 835).
- **Rewrite:** Define each on first use, for example "R35 (rung 5's revocation notification)" and "U3 (the build spec's byte-path rule)". Also cut "the reviews' scenarios" (see finding 33).

**29. Ruling 13 (line 181).**
- **Defect:** "The alternative fails the load" is wrong. The seed rule runs in `serve` after `config.Load`, so the alternative fails `serve`'s start.
- **Rewrite:** "The alternative refuses to start `serve` until the operator restores the key or accepts the default."

**30. 6b-ii criterion 13 (line 1076).**
- **Defect:** "A restored file is owned by the leaf's owner at 0600" can be read as covering the library files from step 6. Library files at 0600 would lock out the streamer.
- **Rewrite:** "The swapped-in database file is owned by the leaf's owner at 0600."

### Tier 4: style and cadence

**31. Sentences that chain clauses.**
- **Location:** Line 313 (89 words, three seams packed around a colon and semicolons), lines 10-19 (73 words; the paragraph is 252 words), lines 303-305 (a 57-word semicolon chain of lock modes), and line 1241 (67 words).
- **Rewrite for line 313:** "Rung 6 declares three test seams. Each exists because no unit test can produce the real condition:" followed by three bullets, one per seam. Turn lines 303-305 into a two-column table of verb and lock mode.

**32. Bold lead phrases as headings.**
- **Location:** Lines 792, 811, 825, and 838 in "Identity in the store".
- **Rewrite:** Make them H3 headings ("The import", "The resolution point", "The epoch", "`serve` and the store").

**33. Review narrative inside normative text.**
- **Location:** Lines 654-656 ("so the fresh review reverses that argument"), 835-836 ("This closes the reviews' scenarios"), 48-49 ("restated where the reviews showed them unbuildable"), and the draft-history reasons in rulings 1 and 7 (lines 64-67, 135-136).
- **Rewrite:** State the rule and move the history to the fold record. For example, line 656 becomes: "A wiped mount under surviving collection directories passes a top-level check, so the count guard replaces it."

**34. "Not X" contrast frames.**
- **Location:** Lines 762-763 ("a **required behavior** 6d builds, not a description of today's code"), 481, and 850.
- **Rewrite for line 762:** "So 6d builds the apply-mode column as a required behavior; today's code applies every auth threshold at restart."

## Length that does no work

Moving these sections to the fold record (`research/2026-09-23-rung-6-spec-fold-record.md`) or to the 6a plan would save about 100 to 120 lines without loss:
- "Behaviors a task relies on" (lines 915-933). It is a verification ledger, and the inline quotes already carry it.
- "Every state a check can meet" (lines 935-947). Most rows point back to tables the spec already has. Keep the inspect, write, and serve rows, or move them to the 6a plan.
- The risk model table (lines 892-906). It restates guards named in their own sections. Keep a pointer for the 6b-ii reviewer brief.
- The ceiling derivation prose (lines 1143-1147). Keep the band, the table, and the total.
- The systemd rationale (lines 409-417). Keep one sentence and the citation.
- The status provenance (lines 3-8). Two sentences are enough.
- "What depends on the configuration homes spec" (lines 1239-1246). It duplicates the [CH] marks; keep one or the other.
- The draft-history clauses in rulings 1, 2, and 7.
