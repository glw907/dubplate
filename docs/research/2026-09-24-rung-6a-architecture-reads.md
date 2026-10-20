# Rung 6a architecture reads

Two `go-architecture-reader` grades at pass 6a's merge (master `8736528`), each read with no
plan in context. Both packages grade **workmanlike**. The store findings route to 6b-i, the
pass that first exports the store's write path and so is the cheapest place to reshape it.
The command findings route to the architecture cleanup pass, because 5c-ii is editing that
package now.

## internal/engine/store (workmanlike; density 0.360 against database/sql's 0.424)

Findings, ranked by the reader, with the idiomatic form:

1. **Test-only seam, `euid`.** `euid int` runs through `leaf.create`, `checkOwner`,
   `leaf.stat`, `openWrite`, `openServe`, and `inspect` (leaf.go:65, :100; open.go:101, :172,
   :308; inspect.go:119-124), and every production caller passes `os.Geteuid()`. The package
   already uses the stdlib seam form for `statfsFunc` (fsprobe.go:43). Use
   `var geteuid = os.Geteuid`, fold `inspect` into `Inspect`, and shrink `openServe` and
   `openWrite` before 6b-i binds to them.
2. **Test-only seam, `migrateOptions.backup`** (migrate.go:62, open.go:261). No production
   caller sets it; tests default it to true, the opposite of production. Drop it until a
   caller has a backup setting to pass, or make production the zero value (`skipPrune`).
3. **`migrate` takes `src fs.FS`** (migrate.go:90, :94; open.go:174, :261) and re-reads the
   set `openServe` already parsed. Take `[]migration` instead.
4. **No production caller** for `openWrite`, `openServe` and its helpers, `handle.update` and
   `writeTx`, `acquireExclusive`, `id`, and several sentinels whose doc comments describe
   `errors.Is` callers that cannot exist while they are unexported. 6b-i is their first
   caller; land each with it, and correct the comments.
5. **`State.String`** (inspect.go:48-69) has no caller while cmd/dubplate-server/store.go:104-127
   restates its table word for word; `storeHolders` restates `busyError`; inspect.go:155-156
   reads /proc/locks twice. Make `String` the implementation and compose from it.
6. **Duplicate rollback** (write.go:129-135, migrate.go:327-333): one free
   `rollback(tx, err, stmtErr)`.
7. **Duplicate poll loops** (lock.go:63-78, open.go:269-301): one `waitLock(ctx, path, how,
   logger)` with the caller bounding ctx; delete `acquireBounded`, `acquireShared`,
   `acquireExclusive`.
8. **Hand-copied SQLite result codes** in four files and five
   `errors.As(err, &se) && se.Code()&0xff == X` sites. Use `modernc.org/sqlite/lib`'s
   `sqlite3.SQLITE_*` constants and one `primaryCode(err) (int, bool)` helper.
9. **DSN keys repeated** across `writeDSN`, `readDSN`, `maintenanceDSN` (conn.go:106-147):
   build from one base list.
10. **Shared preamble** in `openWrite` and `inspect` (open.go:102-123, inspect.go:125-151);
    `openHandle` should take `version` and `lock`.
11. **File split**: open.go (547 lines) holds four concerns. Move classification to
    classify.go, the refusal type beside `State` in inspect.go, the leaf methods to leaf.go,
    and the copy-naming helpers and `preCreate` to snapshot.go.
12. **Comments**: process citations (fsprobe.go:10, lock.go:135, lock.go:102, id.go:9-11,
    leaf.go:28-29), a stale doc.go:1-4, and rebuttals of worries the code never raises.

Best passages: writeTx's first-error latch (write.go:99-121, :215-221); `openServe`'s
handling of flock's non-atomic conversion with re-classification under the exclusive lock
(open.go:165-225); classification and verification opening untrusted files only through the
read-only DSN (open.go:366-372, snapshot.go:98-104).

## cmd/dubplate-server (workmanlike; density 0.450 against cmd/gofmt's 0.298)

1. **Magic-link key names hand-copied in five places** (auth.go:46-59, :268-273, :365, :368;
   serve.go:24-36; doctor.go:128). Move the constant block to auth.go and build every site
   from it.
2. **`doctor` re-derives what `config.Load` returned** (doctor.go:326-336, :362-395): walk
   `loadErr`'s `*sourceKeyError` leaves instead; key every hook error through
   `sourceKeyError` so the "KEY: reason" string split in `rowStatuses` can go.
3. **stdout/stderr parameters** on every constructor (main.go:109-113, check.go:19,
   doctor.go:27, member.go:46, :63, :142, serve.go:157, shadow.go:226): use
   `cmd.OutOrStdout()` and `SetOut`, as `newVersionCmd` already does.
4. **`matcherCommit` threaded as a parameter** (shadow.go:220-226, :235, :237, :250, :354):
   read the package var, as member.go does.
5. **`newServeShell`'s done channel** exists only for its test (serve.go:200-204, :216,
   :273): give it a production reader or drop it.
6. **check restates `storeRow`** (check.go:39-54 against store.go:82-98); `storeBusyTimeout`
   returns three values where one error and a sentinel would do.
7. **`ShadowIntegration`** is the package's only export and reaches nobody (shadow.go:48).
8. **File split**: move the magic-link constants to auth.go; merge the delivery-order var
   and const blocks; merge `registerShadowIntegration` and `registerShadowThresholds`.
9. **Duplicated stanzas**: the unstamped-commit refusal (shadow.go:265, :365),
   `resolveAuthConfig`'s ten parse stanzas, `thresholdStatus`'s error arms.
10. **Comments**: a stale package doc listing verbs that do not exist (main.go:1-19), a
    floating paragraph (serve.go:50-64), about 100 process-citation lines.
11. **`storeBusyTimeoutFloor` and `Ceiling`** are vars holding constants (store.go:42-45).
12. **Test discipline**: serve_test.go:538-545 restates the code under test; the resolution
    sequence is copied into two test helpers (serve_test.go:56-71, :520-535); extract a
    production `buildServe`.

Best passages: `validatePublicURL` refusing "@" before `url.Parse` so a parse error can never
quote a password (auth.go:441-469); `checkRosterDirOwner` and `resolveActor`'s threat-model
choices (member.go:192-223); `executeWithSignalContext` making serve's drain reachable
(main.go:69-73).
