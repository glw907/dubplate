# 6b-i architecture reads (go-architecture-reader, master 6f2c3e0, 2026-09-25)

Routing (conductor): the internal/ops findings and backup's rclone-launch and copy-classification duplication route to 6b-ii, which edits both packages and is ops.Seam's first consumer. The ops/local test seams and the remaining nits route to the architecture cleanup pass.

## internal/ops — workmanlike (density 0.59 vs io/fs 0.57)
- ops.go:140-165 Seam has no consumer (implementor-side interface); the projection should declare the interface it calls.
- catalog.go + ops.go:131-138: catalog chain ends at Runner.Catalog with no non-test caller; KindBackup excepted.
- errors.go: Code consts and Error.Code/Reason unread outside ops; ErrCanceled/CodeCanceled/NewCanceledError zero references (6b-ii Ruling 2 decides its fate); NewInvalidRequestError single in-package caller.
- errors.go:13-82 sentinel list written four times + runtime panic kept in sync by a test; collapse to a self-coding sentinel type once Code has a reader.
- catalog.go:84-101 Catalog/BuildCatalog/Entries duplicate MarshalText's zero-Privilege guard; error branch unreachable in prod.
- ops.go:52-60 ParseHandle restates store's validInstance; its one caller re-decodes the hex.
- nits: Privilege name table x3; KindBackup belongs in ops.go; errors_test wantLiterals self-check; TestImportsStandardLibraryOnly carries a phase expiry ("lands with 6b-ii").
- comments: errors.go:21-23 doc-links nonexistent [ErrForbidden]/[ErrInternal]; errors.go:8-10 process citation; the actor rationale stated four times; test comments cite spec criteria and R-6bi5.
Best: Actor's opaque wire + LocalActor; Privilege iota+1 with refusing MarshalText; Error's fs.PathError shape.

## internal/ops/local — sound with nits (density 0.243)
- local.go:136-153, :88 Runner.Catalog + `var _ ops.Seam` have no consumer (ops.Seam unconsumed); 6b-ii's projection is the first caller, so land/keep them there.
- local.go:80-123 open/New split and bindBackup field are test-only seams; use a package var `bindBackup = (*Runner).backupJob` swapped in tests, or delete open.
- jobs.go:83,109,64 insert's body arg always nil in prod; pass body to run(j, b); drop job.body.
- jobs.go:141-167 vs :198-225 markRunning/finish one body; extract transition(j, query, args...).
- jobs.go:95 hand-listed terminal states vs ops.JobState.Terminal; use state NOT IN ('queued','running').
- nits: errJobNotFound helper (jobs.go:287-327); binder/backupJob belongs in local.go; TestJobStatus builds its expectation with handleOf, the function under test (jobs_test.go:141,146).
Best: jobs.go:178-193 outcomeOf; jobs.go:73-78 storeContext; recovery.go:373-378 owner-only recovery by invariant.

## internal/backup — sound with nits (density 0.265)
- remote.go:11-50 exported Remote with no usable surface; cmd/dubplate-server/backup.go:325-367 restates validName's alphabet (firstInvalidRemoteNameRune + unreachable fallback). Fix: one validator, error carries the rune, Remote.Name().
- prune.go:199-220 isLocalCopyName re-implements store.isCopy / parsePreMigrateName and already disagrees (pre-migrate-v01.db; pre-restore stamp rule); backup.go:132 rebuilds the store's leaf layout. Fix: store method RemoveCopiesBefore on isCopy; unexport PreMigratePrefix/Suffix, PreRestorePrefix.
- prune.go:248-294 vs rclone.go:51-72: three rclone launch paths; prune.list and runRclone lack own process group, SIGINT+WaitDelay cancel, and inherited lease/lock (doc.go:17-20 promises all three). Fix: transport.run takes a stdout writer; prune holds transport; delete runRclone. (Correctness-adjacent: a purge outliving its caller holds no lock.)
- rclone.go:25-39 failure/fail belong in backup.go; every rclone launch in rclone.go.
- rclone.go:101-130 signalNames hand table vs unix.SignalName (nit).
- backup.go:56-62 Run doc restates FinishedMessage contract (nit).
Best: rclone.go:74-99 classify; guard.go:74-89 catalog-derived --max-delete; prune.go:79-105 checkSafety.
