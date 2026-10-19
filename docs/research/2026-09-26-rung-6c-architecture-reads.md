# Rung 6c architecture reads (2026-09-26)

Six go-architecture-reader grades (Opus, no plan in context) over the committed tree at e90998a, with conductor rulings inline. Correctness items are marked C/L/S and routed to one small correctness task; everything else routes to the architecture-cleanup list.

## internal/engine/store — sound with nits
Density 0.391 vs database/sql 0.424 (conn.go 0.67).
Findings (ranked):
1. open.go:219, :85 OpenServe/ServeOptions have no production caller; serve.go never opens the store, so MigrationWait/Keep/Reconcile, settle (open.go:308), finishRestore (open.go:362), waitLock (open.go:378) run only under tests. OpenWrite (backup.go:211) refuses a fresh box with "start serve first" (open.go:656) though serve never creates a store: on a fresh box only restore can create one. CONDUCTOR RULING: already routed to 6d ("ServeOptions.Reconcile wiring", "migration-wait key"); 6d must also make serve create/migrate the store (newServeShell calls store.OpenServe before serve.Run), which the "start serve first" refusal already promises. Pairs with the settings apply ruling.
2. LATENT CORRECTNESS (to the correctness fix task): lease.go:196, :225 TakeKindLease copies h.lock unchecked, DupFiles dereferences l.store.file; RestoreSession.OpenRead (restore.go:506) and settle's handle (open.go:342,:355) build a Handle with nil lock -> panic if a kind lease is taken then duplicated. Fix: refuse in TakeKindLease when h.lock is nil, or DupFiles returns only the lease's descriptor; document which opens carry a lock (conn.go:56-58).
3. inspect.go:159, :367-370 Facts.SettingsGeneration read only by tests; extra query per Inspect.
4. lease.go:38,:41 LeaseFree/LeaseAbsent unused outside; TryLease(instance) (held bool, err).
5. Two poll loops waitLock (open.go:378) / acquireBounded (lock.go:64); acquireSharedNonblocking (lock.go:44) alias.
6. Two rollback helpers write.go:129 / migrate.go:361.
7. Four open+LOCK_NB sequences (lock.go:85-97; lease.go:65-72,120-123,185-195); busyError/leaseBusyError one function.
8. SQLite primary-code mask 5x (conn.go:227, write.go:229,243, open.go:540, migrate.go:382); constants split across 4 files; classifyErr "%w: %v" drops *sqlite.Error (conn.go:229,231) vs write.go:230 "%w: %s: %w".
9. takeSnapshot (snapshot.go:50) / backupCopy (:312) shared core; checkpoint scan duplicated (:90-93 vs :349).
10. "newer than this binary" refusal 5x (restore.go:694,858; open.go:666; migrate.go:116; snapshot.go:170).
11. openHandle callers set h.version/h.leaf by hand (open.go:175-181,293-298,353-354; restore.go:506-511).
12. restore.go:180-190 restates readIntent (intent.go:172); free-space arithmetic inspect.go:286-289 vs snapshot.go:193-198; sidecar suffix lists swap.go:20,139, restore.go:380,430,462; FUSE warning open.go:244-246 = restore.go:133-135; inspect.go:227 re-reads /proc/locks.
13. File split: probe (open.go:463) into fsprobe.go; preCreate to snapshot.go/leaf.go; classify.go for classification in open.go; backup.go for snapshot.go:274-459; migrations.go fold into migrate.go (optional).
14. Test-only: acquireExclusive lock.go:57; takeInstanceLease instance param lease.go:63; errReadOnly/errDiskFull conn.go:40, write.go:37.
15. Comments: inspect.go:190-192 test seam in public doc; leaf.go:58-61, fsprobe.go:26-28 test/workstation provenance; fsprobe.go:10, lock.go:136, :103 process dates; migrate.go:39-47 vs 372-375 twice; leaf.go:118-120 rebuttal; migrate.go:82-83; conn.go:190-193.
Best: write.go:98-121 Tx first-error; open.go:212-218+248-269 flock conversion; conn.go:71-75+147-184 DSN settings.

## internal/config — workmanlike
Density 0.69 vs flag 0.38 (expvar 0.22). Excess is paragraph rationale on unexported code + 41 process citations.
Correctness findings (small, real; route to a small fix task, not the cleanup backlog):
C1. config.go:300 — class flag set but unparseable: `err != nil || !enabled` reports every set key as "set without FLAG" though FLAG is set (bad flag already reported at config.go:423). Form: `if enabled, err := w.integrationEnabled(class); err == nil && !enabled {`.
C2. config.go:284 — `for name := range w.env` appends errors in map order; serve's startup refusal lists them in a different order each run. Form: `for _, name := range slices.Sorted(maps.Keys(w.env))`.
Architecture findings (ranked):
1. Identity selector declared by two families (auth.go:226 AddWallRow + resolveIdentitySource config.go:344; identitySourceKey registry.go:447 a third). wall.missing map exists only to dedupe. Pending machinery vestigial (WallRow.Pending, registry.go:153 branch, ownedNow, Load skip; only test sets it). Fix: selector owned by AddIdentitySource; delete missing/Pending/status-row/auth.go AddWallRow; unexport AddWallRow.
2. config.go:392 / doctor.go:207-230 — Load's joined error per-key structure is text only; doctor splits on ": " and "\n". Fix: type KeyError{Key, Err} with Unwrap; errors.As in doctor.
3. Unused exports: keys.go:28 OwnedWall(), keys.go:109 ReservedClasses() (delete + tests keys_test.go:32,:96); AddSuperseded registry.go:537, AddReservedClass :170, Superseded() :548 test-only; IntegrationClass.Rung keys.go:186 write-only; ApplyRestart/ApplyImmediately settings.go:21,27 no production setting.
4. CorrelatedRule settings.go:96 / AddCorrelatedRule registry.go:496 test-only registrants; settings/evaluate.go:239 production loop fed by nothing. Remove until a rung registers a rule.
5. Seven linear classifiers (keys.go:36,116,167,200; registry.go:447; thresholds.go:112; settings.go:169) restate Registry.claims (registry.go:52); recorded family is call-order dependent (registry.go:252,:280). IntegrationClass.declares predicate duplicated (keys.go:202, registry.go:241).
6. claim tolerates same-family repeats so each Add* re-scans (registry.go:149,171-182,234-245,276,424-428,468,539); make claim reject any repeat, identity sources skip.
7. Kind-to-parser table 3x in package (registry.go:313-334, settings.go:120-131, thresholds.go:84-102) + settings/evaluate.go:368 + doctor.go:290; threshold ParseFloat (thresholds.go:99) accepts NaN, Setting refuses. Redundant bad-default branches thresholds.go:65-69, :133-135.
8. ThresholdKind zero value = KindInt vs ApplyMode iota+1 (thresholds.go:16-21 vs settings.go:19-28); start at iota+1, delete registry.go:302-306 paragraph and TestKindInt_IsZeroValue (thresholds_test.go:166).
9. registry.go:148 requirePrefix unreachable.
10. File split: keys.go mixes seed tables with family types/classifiers; move types beside Add methods (families.go).
11. Enabled doc keys.go:209-212 duplicates config.go:258-262.
12. Tests: registry_test.go:34-41 derives expectations from the seed tables (cannot fail); keys_test.go:144 name mismatches fixture.
13. Comments: 41 process citations in non-test files; stale: registry.go:269-271, config.go:59-60, keys.go:182 line cites; config.go:9-20 restates reserved table; long rationale config.go:106-126, 403-419, registry.go:106-131, 286-310.
Best: config.go:231-256 requireSiteName/isBidiControl; settings.go:120-163 buildValidator/boundedValidator; registry.go:26-33 panic-on-misdeclaration rationale.

## internal/settings — workmanlike
Density 0.348 vs net/http/cookiejar 0.352.
Findings (ranked):
1. HIGHEST: apply.go:29 apply/writeChanges(:65)/errNoGeneration(:17) have no production caller; serve.go never references settings; yet cmd/dubplate-server/backup.go:274 and store.go:280-284 tell operators serve seeds "at serve's next start". Accessor generation check (accessor.go:92-106) guards a writer that is only a test. Fix: export Apply and call it at serve startup after migrations, or move apply.go out until the consuming rung lands (forward-hook rule). CONDUCTOR: verify whether 6d's ServeOptions.Reconcile wiring handoff owns this; if not, it is an operator-facing promise with no code.
2. evaluate.go:29-43 Row copies store.SettingsRow field for field; SELECT+scan exists twice (evaluate.go:26,190 and store inspect.go:343-361). Use store.SettingsRow; scan once in store.
3. evaluate.go:368-379 parse restates config.buildValidator (config/settings.go:120) and accessor.go:43-57 a third time; already drifted (config refuses NaN/Inf via parseFiniteFloat, parse accepts). evaluate.go:286-287 can %w-wrap nil if they disagree. Fix: config.ErrOutOfBounds sentinel + Setting.Equal.
4. evaluate.go:356-364 lookupEnv restates config newWall (config.go:174); doctor.go:412 lookupEnviron third copy. One exported lookup in config.
5. evaluate.go:97 Outcome.setting unexported -> doctor re-finds by key (store.go:246) and re-implements withholding (settingShown store.go:323). Fix: config.Setting.Show + exported Outcome.Setting.
6. evaluate.go:86,89 EnvSet/Value, Row.SeedValue/Actor/UpdatedAt unread outside package.
7. evaluate.go:150-162 table-absent branch owned twice (check.go:90-92, store.go:223,293).
8. Row scan helpers (evaluate.go:24-27,164-206) belong with the row type.
9. Comments: evaluate.go:132-133 caller list; evaluate.go:188-189 scanRows doc untrue for apply's path (store.Rows has no Err; write.go:196-204).
10. evaluate_test.go:72-75 process citation; TestEvaluateFacts_ConvertsStoreRows (:33) vacuous on SeedValue/Actor/UpdatedAt.
11. Nit accessor.go:27 loaded duplicates a.rows != nil.
Best: accessor.go:86-96; evaluate.go:297-334 seed switch; evaluate.go:239-264 correlated rules.

CONDUCTOR RULING on finding 1: deliberate per R-6c9 ("the apply is unexported until 6d"). Route to 6d's handoff as a named item: 6d must wire Apply at serve startup, and the operator strings at cmd/dubplate-server/backup.go:274 and store.go:280-284 are false until it does, so 6d must land before any deploy that exposes them (5d's deploy gate should check this). The accessor's cross-process generation check also first gets a real writer then.

## internal/ops/local — sound with nits
Density 0.290 vs database/sql 0.424 (local.go 0.388).
Correctness findings (real; route to the small correctness fix task with config C1/C2):
L1. jobs.go:139-141, :123-127 — if finish's terminal write fails (disk full, busy past writeWait), run logs and exits: record stays 'running', owner lease held, job out of r.jobs, so no recovery ever interrupts it; backup verb's awaitBackup (cmd/dubplate-server/backup.go:288) polls forever; SIGINT sets cancel_requested nothing reads; 2nd SIGINT ignored (done=nil). markRunning failure leaves 'queued' likewise. Fix: retry the terminal write on storeContext with backoff, keep j in r.jobs until it lands so Close waits; or recoverJobs treats a live record owned by own instance with no r.jobs entry as dead.
L2. local.go:120-126 vs :233-246 — Close does wg.Wait then releases lease; a concurrent StartBackup can wg.Go racing Wait at zero (WaitGroup contract violation) or insert a record after the owner lease is released (another process's recovery interrupts a running job). Type claims concurrency safety. Fix: closed bool under r.mu (net.ErrClosed form), start checks it before insert.
Architecture findings:
1. jobs.go:89,:115 insert's body param only non-nil in jobs_test.go:25; drop it.
2. local.go:78-82, :187-190 bindBackup test seam + binder type (jobs.go:31-33); use package var backupBody.
3. jobs.go:147-173 / 204-231 markRunning & finish duplicate; one transition helper.
4. local.go:145-151 !ok branch repeats ops.BuildCatalog's check (catalog.go:108-115).
5. live-state pair written 3x (recovery.go:34,44; local.go:263); prune's terminal set jobs.go:101 restates ops.JobState.Terminal; use NOT IN live fragment.
6. jobs.go:362-367 parseID double hex decode.
7. local.go:96-98 New checks Logger but not Retention or the three backup readers (local.go:39-43, 59-66) -> panic at first StartBackup.
8. File split: backupJob (jobs.go:35-55) beside backupBind; wrapCanceled (local.go:311-328) below CancelOwn.
9. Rationale twice: local.go:39-42 vs 206-211; 59-63 vs 182-186; wrapCanceled 12 lines -> ~4.
10. recovery.go:14-15 doc inaccurate (absent lease = dead owner).
11. Test citations R-6c11, R-6bii5, "spec criterion N" (helpers_test.go, local_test.go, process_test.go, recovery_test.go, jobs_test.go).
Best: jobs.go:184-199 outcomeOf; jobs.go:77-82 storeContext; recovery.go:10-13 liveOwners.

## internal/backup — sound with nits
Density 0.314 vs os/exec 0.670 (excess in backup.go 0.446, fetch.go 0.362).
Findings (ranked):
1. remote.go:21,:40 — cmd/dubplate-server/backup.go:425-485 discards ParseRemote's Remote, re-cuts the string, reimplements validName as firstInvalidRemoteNameRune with an "unreachable" fallback. Fix: Remote.Name() + typed *RemoteError{Value, NoColon, Rune}; cmd uses errors.AsType.
2. backup.go:174 derives store snapshots dir via double filepath.Dir; prune.go:169-194 walks it with store.IsPreMigrateCopy (only caller). Fix: store Handle.RemovePreMigrateCopies(before), unexport IsPreMigrateCopy.
3. prune.go:80 checkSafety, :104 pruneSnapshots, fetch.go:63 pickSnapshot — three stamp loops; one otherStamps(entries, own) helper.
4. fetch.go:115-117 and restore.go:416-418 identical failure-code wrap; withCode(err) helper in rclone.go.
5. backup.go:111, restore.go:219 duplicate LookPath+transport; newTransport constructor. (reconcile.go:67-76 classification: leave.)
6. File split: restore.go:510 source.checkLibrary belongs in fetch.go, rename checkRemoteLibrary; restore.go:430-625 library-root section -> libraryroot.go; failure/fail (rclone.go:27-41) -> backup.go.
7. backup.go:19 FinishedMessage string sentinel (caller store.go:371); low leverage.
8. restore.go:114-124, doc.go:28-32 errors phrased in CLI flag names inside a library package.
9. rclone.go:47-51,:61 nil-inherit branch test-only; RestoreConfig.Now vs time.Now() inconsistency.
10. Comments: registry-owns-policy rationale restated 4x (backup.go:64-73,:44-50; guard.go:40-44; prune.go:27-35). Test comments cite R-6bi15 (backup_test.go:11,25).
Best: rclone.go:59-84 transport.run; prune.go:80-99 checkSafety; fetch.go:128-151 syncStage.

## cmd/dubplate-server — workmanlike
Density 0.430 vs cmd/covdata 0.176 (gofmt 0.298); main.go 0.878, serve.go 0.598, auth.go 0.554. 117 process-citation comment lines (auth.go 32, backup.go 17, main.go 15, shadow.go 13, doctor.go 12).
Findings (ranked):
1. doctor.go:215 rowStatuses parses config.Load's text on ": "; auth.go:246 sourceKeyError typed only for identity path; doctor.go:394 identitySourceValues and :412 lookupEnviron (11 call sites) copy config's resolveIdentitySource/wall.get; doctor.go:428 splitKeyReason. Fix: config.KeyError + config.Lookup (pairs with config finding 2 and settings finding 4).
2. Magic-link key names declared in parallel: serve.go:26-36 consts, auth.go:46-59 literals, auth.go:268-273 literals, doctor.go:156 rosterKey 3rd; DUBPLATE_BACKUP_REMOTE literal at backup.go:37,131,424, restore.go:168.
3. store.go:218-219 restates the five failed states of storeRowStatus (store.go:156-180); lists disagree on an unrecognized state (storeRowStatus fails it, settingsDoctorRows falls through). store.go:223-237 vs check.go:63-68 table-absent answered by two implementations. [latent correctness: unknown state]
4. backup.go:455-485 backupRemoteParseError/firstInvalidRemoteNameRune rebuild backup validName; backup.go:428 re-cut. (pairs with backup finding 1)
5. store.go:106 storeBusyTimeout returns (d, parseErr, rangeErr), 3 of 4 callers treat alike (check.go:43-49, backup.go:198-204, restore.go:156-162); check.go:43-53 restates storeRow.
6. Exit codes 75/130 declared twice (backup.go:143-147, restore.go:23-26); one block in main.go.
7. SIGHUP handled two ways (restore.go:174 NotifyContext vs backup.go:289-311 raw Notify).
8. serve.go:243 newServeShell returns a test-only channel (discarded serve.go:191).
9. serve.go:349-355 authHandlersConfig 10 params for a test; serve_test.go:653 expected value computed from code under test (vacuous-ish).
10. Verb constructors take stdout/stderr instead of cmd.OutOrStdout (main.go:138-144, :160; member.go:63 `_`); matcherCommit param shadow.go:220-237.
11. File split: store.go mixes key registration (17-118), settings doctor rows (203-328), backup facts rows (330-432); row type misnamed backupFactRow (store.go:199); backup.go:384-573 is doctor's remote row + rclone probe; doctor.go:409-434 shared helpers; serve.go:26-63 auth keys + route names. Proposed: doctor.go owns all row renderers (doctorRow), verbs only in backup.go/restore.go, auth consts to auth.go, route names to web.go.
12. shadow.go:48 ShadowIntegration only exported identifier; unexport.
13. main.go:3-19 package doc lists import/ping-fail (unregistered), omits member; rung tags. Stale line cites auth.go:513, serve.go:336.
14. Heavy history passages: backup.go:89-100, serve.go:65-79 (detached), doctor.go:75-95, 206-214, 255-270, 312-329, auth.go:23-42, main.go:102-126. Keep auth.go:431-440.
15. Nits: doctor.go:276-306 repeated arms; doctor.go:343-352/396-405 slices.Concat; shadow.go:265/:365 repeated refusal; backup.go:567 var `i`.
Best: web.go:658-709 readOperatorCSS; auth.go:441-450 validatePublicURL; backup.go:207,301-311 WithoutCancel + nil-out.
