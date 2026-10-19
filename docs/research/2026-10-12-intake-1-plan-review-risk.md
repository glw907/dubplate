# intake-1 plan review: data integrity and failure risk

Reviewer: fresh context, risk lens. Read `docs/plans/2026-10-12-intake-1.md` at `3783060` against the design,
the spec (revision 2), the shipped code at HEAD, and the module cache (tusd v2.10.1, mewkiz/flac v1.0.14,
mewkiz/pkg errutil). No finding would change the approved design. Every fold is plan text, a fixture, or a mutation,
and finding 5 adds one column to `0006` before its hash pin.

Counts: 0 blocker, 5 major, 6 minor (correctness), 5 minor (optional, two of them over-ceremony).

## Findings, ranked by consequence

### 1. Major (correctness): cleanup's scope, read as written, deletes bytes a live state still needs

- **Location.** Plan `:780-781` (T11: "cleanup after each commit per 'Staged bytes, per state'"). The plan inherits this
  from the spec (`:588-592`: "Reconcile and cleanup test a submission against this set. For loose files, the staged bytes
  are its received files in `tus/`"; `:815-816`: "It removes every byte outside each submission's staged set").
- **Defect.** Under that general rule, two in-flight byte sets sit outside the staged set:
  - An `uploading` submission's *unreceived* partial tus files. A paused or retrying upload holds no `memorylocker`
    lock between PATCHes, so a sweep every 10 minutes deletes it. The client's next HEAD gets a 404, and it re-creates
    and re-sends from zero. That breaks resume, which is `intake-2`'s AC 12.
  - A `checking` ZIP's `x/` during extraction and decode. The staged set for `checking` is the archive alone, and member
    rows are written only in the verdict `Update`. So "a file in `x/` with no row" matches every member mid-run. A sweep
    that removes `x/` under the worker makes a member write fail, which T7 maps to `unpack-failed` (finding 3). The
    result is `rejected`, and the archive is cleaned up with it: a good album is refused and its only copy deleted.
- The spec's enumerated list (`:816-819`: terminal bytes, an orphan `submissions/` directory, the archive of a ZIP past
  its verdict, the failing files and extras of an accepted separate-tracks submission) is safe. The general sentence is
  the trap, and the plan points the executor at it.
- **Fold.** T11's outcome replaces the pointer with a closed list: cleanup removes exactly the spec's four kinds plus
  the sweep's orphan `tus/` file (no row, older than the resume window), and nothing else. A non-terminal submission's
  tus files, received or not, are never touched, and a `checking` submission's `x/` is never touched. Add two fixtures
  to AC 27/28, each with a mutation that applies the general rule:
  - a sweep over an `uploading` submission holding a half-written, unlocked tus file, which is left byte-identical;
  - a sweep during a `checking` ZIP's extraction (an injected pause), which ends at the uninterrupted verdict.

### 2. Major (correctness): a graceful stop mid-check is an unnamed state, and the plan names no shutdown order

- **Location.** Plan T6 `:609` (the context checked between frames), T11 `:777-779` and `:787-790` (the worker and
  sweep on `serve`'s context). Shipped code: `serve.Run(cmd.Context(), ...)`, then a deferred `h.Close()`
  (`cmd/dubplate-server/serve.go:229-234`).
- **Defect.** Every deploy and every `systemctl stop` sends SIGTERM, which cancels `serve`'s context while a check runs.
  The plan says the gate returns on cancellation but not what the worker records:
  - Mapping `ctx.Err()` to "could not run" moves every in-flight check to `check-unfinished` on every deploy, and each
    member must press retry.
  - Mapping it through the decode path can yield a verdict.
  - Committing on the cancelled context fails, which is accidentally right but untested.

  Separately, nothing orders teardown. The worker and sweep must stop and be joined before the intake lease is
  released and before `h.Close()`, or a verdict `Update` and a cleanup run against a closing store. The close smoke uses
  `kill -9` only, so the common path, a graceful stop, is never exercised.
- **Fold.**
  - T11 states the rule: a context cancellation is neither a verdict nor "could not run". The worker commits nothing,
    the submission stays `checking`, and the next startup pass re-queues it.
  - Shutdown runs in this order: cancel, wait for the worker and sweep, release the lease, close the store.
  - Fixture: `serve`'s context cancelled mid-decode leaves `checking` with its bytes and reservation, and a second start
    reaches the verdict. Mutation: map cancellation to `check-unfinished`.
  - The close smoke adds a SIGTERM restart beside the `kill -9` one.

### 3. Major (correctness): server I/O faults can be classified as member faults, and a rejection deletes the bytes

- **Location.** Plan T7 `:649-650` ("`unpack-failed` for an open failure, a member read error, a CRC mismatch, or an
  overrun"), T6 `:610-611`, T11 AC 24 fixture `:810-811` ("a staged file made unreadable").
- **Defect.** The spec (`:631-634`) makes a read error on a staged file "could not run", which keeps the bytes. Three
  paths in the plan lose that line:
  - **Extraction write failure.** T7 names no class for a *write* failure during extraction (ENOSPC, EIO, a removed
    `x/`). "An open failure" reads as either the ZIP's or the destination's. The spec's floor check runs before
    extraction, but concurrent PATCHes or other writers can still fill the volume. An executor folding these into
    `unpack-failed` rejects a good album and cleanup deletes its archive. This is the full-disk case the brief asks
    about.
  - **Mid-stream read error.** mewkiz/flac's frame path returns the underlying reader error raw (`frame/frame.go:128`,
    `:234`, and others), but its `errutil.Err` wrapper has no `Unwrap` (`mewkiz/pkg/errutil/err.go`, no `Unwrap`
    method). An EIO mid-decode can therefore surface in a shape the gate cannot tell from damage, and become `corrupt`.
    A `zip.Reader` over a failing `ReaderAt` has the same problem against `zip.ErrFormat`.
  - **Fixture.** AC 24's "made unreadable" (a chmod) exercises the open path only, never a mid-stream fault.
- **Fold.**
  - T6 and T7 classify by the gate's own I/O, never by the decoder's error. The sha256 tee that already reads every
    byte records the first non-EOF error from the staged file, and any recorded read error is "could not run" whatever
    the decoder returned.
  - In the ZIP walk, `unpack-failed` is exactly `zip.ErrFormat`, `zip.ErrAlgorithm`, `zip.ErrChecksum`, and the
    declared-size overrun. Every error from reading the archive file or writing a member is "could not run".
  - Fixtures: a reader that fails with EIO after N bytes, during decode and during the walk, each ending
    `check-unfinished` with the bytes kept; and an extraction whose destination write fails, ending `check-unfinished`
    with the archive kept. Mutation: map the write error to `unpack-failed`.

### 4. Major (correctness): extracted members are never synced before the archive, the only other copy, is removed

- **Location.** Plan T7 `:648-650`, T8 `:683-686`, T11 cleanup. Spec `:576-579` ("extraction writes directly rather
  than through `internal/atomicfile`") against `:383-385` (receipt syncs because "the store's commit is durable on
  another device").
- **Defect.** After a ZIP's verdict commits, its members in `x/` become its staged bytes, and cleanup then removes the
  archive. The members were written without `fsync`. A power loss after the archive's removal leaves zero-length or
  truncated members (ext4 delayed allocation), while the store, on another device, durably says `matching` or
  `accepted`. Reconcile tests presence ("missing"), so a present, truncated member passes. Rung 9 would then admit a
  corrupt file, which is silent corruption. Receipt's own rationale applies here word for word.
- **Fold.**
  - T7/T11: sync each extracted member and the `x/` directory before the verdict `Update`. That costs one fsync per
    member per check and needs no atomic rename, so the spec's "writes directly" still holds.
  - Reconcile tests a staged file's size against its row's `size`, not presence alone, so a short file is `expired`
    with a reason rather than kept.
  - Fixture: a `matching` ZIP whose member is truncated over a restored store ends `expired`. This does not change the
    design; the spec is silent on sync for members.

### 5. Major (correctness): the state-equals-event assertion is ill-defined, and its fix belongs in `0006` before the pin

- **Location.** Plan `:116-119` ("each `submissions.state` equals its latest `submission_events` row's `kind`"),
  against the spec preamble `:1094-1096` ("latest **state** event's kind"). T4 `:532-534`. T11 AC 28 fixture
  `:819-821`.
- **Defect.** There are two problems.
  - **The event-only kind.** `stalled` is a `submission_events` row whose `kind` is no state (P5). AC 28's two-sweep
    fixture ends with a `checking` submission whose latest row is `stalled`, so the assertion as the plan words it fails
    by construction. The executor then either special-cases it in a test or, worse, makes `stalled` a state.
  - **No deterministic order.** "Latest" has no defined order. The store's `at` is `UnixMicro` (`ops/local/jobs.go:75`,
    `settings/apply.go:61`) and `id` is random (`NewID`), so ties fall in random order, and a wall-clock step back
    misorders events. T11's fixtures inject a clock for the sweep while create and receipt may stamp real time, so a
    sweep's `expired` event can sort *before* the `uploading` event it ends. `get_submission`'s `events` list has the
    same problem.
- **Fold.**
  - Word the shared assertion as the spec does: the latest event whose `kind` is a state.
  - Give `submission_events` a `seq INTEGER NOT NULL` with `UNIQUE (submission_id, seq)`, written as max+1 inside the
    transition's `Update`. "Latest" and the wire order then read `seq`, and the unique index also refuses a second
    writer that slipped past a compare-and-set.
  - T4 decides this now, because `0006`'s hash pin makes a later change a new migration. A lighter alternative is
    `ORDER BY at, rowid`, which works on a rowid table but is implicit.
  - Mutation: append `stalled` and assert that the shared assertion still passes.

### 6. Minor (correctness): tusd's `MaxSize` is fixed at mount from a next-operation setting

- **Location.** Plan P3 `:223` (`MAX_ALBUM_BYTES`, `ApplyNextOperation`), T9 `:722` ("`MaxSize` from the album cap").
- **Defect.** The handler config is built once at startup. If an admin raises the cap, the manifest admits a file
  between the old and new caps, and tusd then answers 413 at the tus create (`unrouted_handler.go`, `validateMaxSize`).
  The submission sits in `uploading` holding its reservation and a session slot for the 7-day resume window.
- **Fold.** Set `MaxSize` to the key's ceiling (`17179869184`). The create callback's equality of `Upload-Length` to
  the declared size, which the live cap already bounded at create, is the real bound. The alternative, marking the cap
  `ApplyRestart`, is weaker. Add a fixture: raise the cap, create, and upload a file between the two caps.

### 7. Minor (correctness): the hand-off from a move to `checking` into the worker is unnamed across three tasks

- **Location.** T8 `:680-682` (receipt moves to `checking`), T9 (the callback calls it), T10 (`retry_submission` moves
  `check-unfinished` to `checking`), T11 `:777` (the worker pool).
- **Defect.** No task says how a commit that moves a submission to `checking` reaches the worker. Each failure mode
  strands a submission in `checking` until a restart, and fires the stall mail for a check that is not running:
  - an in-memory queue that drops when full;
  - a queue that blocks inside the tus callback, which holds the upload lock;
  - a signal sent before the commit.

  AC 24's retry fixture would catch a missing wire only after a fix cycle.
- **Fold.** T11 names one mechanism:
  - The worker selects `checking` rows not in flight from the store, so the store is the queue.
  - A capacity-1 wake channel is poked after commit by receipt's callers, `retry_submission`, the startup pass, and
    each sweep tick.
  - No in-memory queue can lose work, and the sweep tick also retries a verdict that hit a store error, which the spec
    leaves until restart.

### 8. Minor (correctness): the wrapper's id must be tusd's id, byte for byte

- **Location.** Plan T9 `:723-724` ("the id parsed and owned ... then the cut path to tusd").
- **Defect.** tusd's routed handler takes the whole remaining path as the id: `path := strings.Trim(r.URL.Path, "/")`
  (`handler/handler.go:38`) and `extractIDFromPath` returns `strings.Trim(path, "/")` (`unrouted_handler.go:1703-1705`).
  The plan does not say the wrapper computes the id the same way from the same field. A wrapper that authorizes the
  first segment, or parses `RawPath` or `EscapedPath()`, while handing tusd a different `r.URL.Path` can authorize
  `<own id>` and let tusd act on `<own id>/<other>`, which is the "reaches tusd before the wrapper authorizes" hole. The
  plan also leaves `<prefix>/tus` with no trailing slash unnamed (`Shell.Register` routes it, `serve.go:64-66`).
- **Fold.**
  - T9 states that the wrapper cuts the prefix from `r.URL.Path`, sets `RawPath` to `""` as `opshttp` does
    (`mount.go:78`), and computes the id as `strings.Trim(rest, "/")` before requiring 32 lowercase hex characters.
  - `<prefix>/tus` takes the create branch or the shared 404, whichever the T9 author names.
  - Add a path-shape table to AC 8: `/<id>/`, `//<id>`, `/<own>/<other's>`, `/<id>%2F<id>`, and uppercase hex, each
    answering the shared 404 or acting on the authorized id only. Mutation: parse the first segment.

### 9. Minor (correctness): a torn `.info` is an unnamed state

- **Location.** Plan T8 `:680` (receipt syncs "the tus data file and the `tus/` directory"), T11 startup pass. tusd
  `filestore.go:303-308` (`writeInfo` calls `createFile` with `O_TRUNC` and no sync, `:323-348`) and `:130-140`
  (`GetUpload` treats only `IsNotExist` as not-found).
- **Defect.** A power loss after a 201 can leave `<id>.info` empty. `GetUpload` then returns a JSON error, a plain
  error that `sendError` turns into 500 (`unrouted_handler.go:1288-1291`). The client retries a 500 and then fails, and
  the create callback's "no upload yet" probe errors the same way. The file can never be sent, and the submission sits
  until the 7-day expiry. Receipt does not sync `.info` either, so a received file's HEAD can 500 after a power loss.
- **Fold.**
  - Receipt also syncs `<id>.info`.
  - The startup pass treats an unparseable `.info` of an *unreceived* file as absent and removes both files, so the
    client's HEAD gets a 404 and it re-creates through `uploadUrl`.
  - Add a row to the states table, and a fixture with an empty `.info`.

### 10. Minor (correctness): T0's sweep cannot see tusd or mewkiz/flac

- **Location.** Plan `:342` ("tusd v2.10.1 is still the latest release (proxy 2026-09-14)"), T0 `:441-448`, P14
  `:323-324`.
- **Defect.** The reading is four weeks old and presented as current, which is a recalled behavior rather than a
  quoted one. T0 sweeps `go.mod`, and neither module is in `go.mod` until T6 and T9, so CLAUDE.md's "newest production
  version" rule never checks them.
- **Fold.** T0 runs `go list -m -versions github.com/tus/tusd/v2 github.com/mewkiz/flac` and records the result. A newer
  minor or patch is taken in the importing task, with the spec's tusd line citations re-read against it. A new major
  is held for Geoff.

### 11. Minor (correctness): receipt runs on the request's context

- **Location.** Plan T8 `:680-682`, T9 `:728`.
- **Defect.** `PreFinishResponseCallback` receives the request's context. A client that drops right after the last
  byte cancels the receipt `Update`, so the file waits up to one sweep interval (10 minutes) before receipt. The spec
  covers recovery; the fold is free.
- **Fold.** Receipt's `Update` runs on `context.WithoutCancel` under its own timeout, following `notice.go:26-53`.

### 12. Minor (optional): `matching` has no exit in this build

- **Location.** P5 `:270-276`.
- **Defect.** A passing album waits in `matching` with no probe, no cancel, no stall, no expiry, and no startup
  re-queue, holding its reservation and its bytes indefinitely. That is harmless while nothing deploys.
- **Fold.** The PR body and STATUS state that `intake-1` must not deploy before `intake-2`. This is a sentence, not
  code.

### 13. Minor (optional): "open reservations" and "non-terminal" are both used for one set

- **Location.** T8 `:672` (the group cap: "open reservations") against `:671` and `:691` ("non-terminal").
- **Defect.** "Open" can be read as `uploading` and `checking` only, which would drop `matching` and `accepted` from
  the group cap and double-admit.
- **Fold.** Say "non-terminal" in both places. AC 15's group-cap fixture includes an `accepted` submission.

### 14. Minor (optional): one store lease cannot keep two `serve` processes off one staging directory

- **Location.** T11 `:788-789`, spec `:345-350`.
- **Defect.** `TakeKindLease("intake")` lives under one state directory's `leases/` (`lease.go:184`). Two `serve`
  processes on different state directories that share one staging directory both pass the lease. Each startup cleanup
  then removes the other's `submissions/` directories as rowless. This is a misconfiguration only, never the box's unit.
- **Fold.** Either one sentence in the runbook ("one staging directory per store"), or an exclusive flock on
  `<staging>/.intake-lease` beside the store lease. The runbook sentence is the minimum.

### 15. Minor (optional, over-ceremony): the security reviewer reads T8 to T11 twice

- **Location.** T11 `:795-801` and Close `:884-885`. The checklist globs bring `web-auth-security-reviewer` for every
  `auth-data` path.
- **Cost.** One extra Opus read over the same range, about 15 minutes and a few hundred thousand tokens.
- **Fold.** Hand the close review T11's `security-read=` sha, and scope its security read to the changes since it.

### 16. Minor (optional, over-ceremony): four mutation runs for one exact-set pin

- **Location.** T1 `:469-470`.
- **Defect.** `TestProductionRegistry_StoreHomedSensitiveSet` compares an exact sorted set, so one reverted mark proves
  that it can fail.
- **Fold.** One mutation run. This saves about 3 minutes.

## The four planning-miss gaps

- **Behaviors quoted, never recalled.** These citations match v2.10.1: tusd `error.go:25-34`, `filestore.go:73-79`,
  `unrouted_handler.go:922-933` (the plan's `:926-931` sits inside it), `:1288-1297`, and `PostFile`'s
  callback-then-`NewUpload` order (`:366-401`). One quote is stale (finding 10). The id-extraction rule the wrapper
  relies on is not quoted (finding 8).
- **Each state a check can meet names its report.** These states are unnamed:
  - a graceful cancellation mid-check (finding 2);
  - a write failure during extraction, and a mid-stream read fault (finding 3);
  - cleanup meeting an in-flight upload or extraction (finding 1);
  - a torn `.info` (finding 9);
  - the wrapper's path shapes and `<prefix>/tus` with no trailing slash (finding 8);
  - a tus create rejected 413 after a cap raise (finding 6).
- **Each privileged field names every channel.** One optional addition: the tus binary path. `FileInfoChanges.Storage`
  can redirect a write to an absolute path (`filestore.go:90-97`). The plan returns it nil, and the States table could
  say "never set" so that a later edit is a visible change.
- **Each proof names its fixture state.** Mostly met. AC 24's "staged file made unreadable" names a state that reaches
  only the open path (finding 3). The finding-1 and finding-2 fixtures are missing.

## Questions from the brief, answered

- **P11, `0006` on a store at version 5.** The decision holds. The registry default is 3 with a floor of 1
  (`cmd/dubplate-server/store.go:101-111`), so a busy pre-lock `Inspect` prunes to 3 pre-migration copies at most.
  `0006` is the first migration that can meet this, so no older pre-migration copies exist to lose.
- **Upgrade and rollback.** After `0006`, the pre-intake binary refuses the store (future version) and refuses its
  environment: a set `DUBPLATE_STAGING_DIR` is a reserved `pipeline` name to it (`config.go:308-309`). A rollback is
  therefore a pre-migration snapshot restore plus a `deploy.sh` of the old commit, which renders the old template. This
  is unchanged from 6d. The runbook line T12 writes could say so; that is optional.
- **Restore.** The spec closes this. Reconcile expires rows whose bytes are gone, and rowless `submissions/` directories
  and old orphan tus files are removed. Finding 4 adds size to the presence test.
- **Concurrency and leases.** The compare-and-set transitions, `IMMEDIATE` `Update`, limits taken over `*store.Tx`,
  and the lease before recovery all hold as planned. The one gap is the cross-state-directory case (finding 14).
- **Reservation leaks and double counts.** None in the planned states. `matching` holds a reservation by design in
  this build (finding 12). The ZIP's archive is uncounted in quota while both copies exist, which the spec chose and
  the floor covers.
- **Cross-member access.** None found beyond finding 8's precondition. Use the shared 404, the ownership check before
  any tus lookup in the create callback, and `actor.UserID()` alone, as planned.
- **Does any of P1 to P17 reopen a hole the spec closed?** No. P3's `ApplyNextOperation` on the album cap interacts
  with P7's `MaxSize` (finding 6), and P5's narrowed state set leaves `matching` without an exit (finding 12). Neither
  reopens a spec-closed hole.
