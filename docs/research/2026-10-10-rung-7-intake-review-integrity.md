# Rung 7 intake spec, revision 1: data integrity and failure review

Lens: data integrity and failure risk. Target: `docs/specs/2026-10-10-rung-7-intake-design.md` at
`9fc7c53` (cited below as `spec:<line>`). Code read on master at `9fc7c53`; tusd read from the module cache at
v2.10.1; `tus-js-client` read from the published 4.3.1 tarball (`npm pack tus-js-client@4.3.1`, `lib/upload.js`).

Findings are correctness gaps ranked by consequence. Settled rulings are not reopened. No finding is an owner fork.

Counts: 2 blockers, 6 majors, 4 minors. No over-ceremony finding in this lens costs enough to be worth a fold.

## What holds

- **The quota race at creation.** `Handle.Update` "runs fn inside one IMMEDIATE transaction on the write pool"
  (`internal/engine/store/write.go:44-45`), so a check-and-insert inside one `Update` serializes against every other
  writer, in any process. AC13 is reachable as specified, provided the `statfs` read and the reservation sum sit inside
  the same callback.
- **The album as the verdict unit.** The first failing file rejects the submission (`spec:491`, `:547`), AC20 proves it,
  and rung 7 admits nothing, so no half-admitted album can come from rung 7.
- **Ruling 10's lease is not preempted.** The worker "takes no store lease" (`spec:453-454`), no member action leaves
  `accepted` (`spec:630-632`, `:688-689`), and neither recovery nor the sweep touches `accepted` (`spec:707-716`).
  Rung 9 inherits `accepted` untouched. The `album_key` column is a natural key for rung 9's per-album lease, which is
  rung 9's to choose.
- **Migration `0006` on the box.** `collections` holds no row before rung 7 (`spec:569`), so the added column and its
  unique index meet no existing data. Each migration is "applied once in its own transaction"
  (`internal/engine/store/doc.go:47-48`), so a failed upgrade leaves the store at `0005`.
- **A backup taken mid-upload.** Staging sits outside the library root (`spec:104-108`), and the backup syncs only the
  library tree and a store copy (`internal/backup/doc.go:4-7`), so partial files never reach the remote or the guard's
  count. The risk is at restore, below (M3).

## Blockers

### B1. A finished upload whose finish step did not commit is reported complete and never finishes

Location: `spec:359-361` (`PreFinishResponseCallback`), `:376-378` (fingerprint options), `:929-930` (re-create 409).

tusd calls the finish callback only inside the request that writes the last byte. `finishUploadIfComplete` runs
`emitFinishEvents` when `info.Offset == info.Size`, and a callback error is returned to the client
(`pkg/handler/unrouted_handler.go:993-1009`, `:1015-1019`). HEAD calls no callback, and the filestore derives the offset
from the data file alone: `info.Offset = stat.Size()` (`pkg/filestore/filestore.go:166`). On resume,
`tus-js-client` treats a full offset as done: `if (offset === length) { this._emitProgress(length, length);
this._emitSuccess(res); return }` (`lib/upload.js:724-727`).

Failure sequence one (stuck submission). The last PATCH of a file writes every byte. Then `serve` crashes, or the
callback's `Update` fails (busy timeout, `SQLITE_FULL`), before the file is marked received. The client retries, HEAD
answers `Upload-Offset` equal to the length, and the page reports the file sent. The row stays unreceived, the
submission stays `uploading`, and the sweep expires it after `DUBPLATE_INTAKE_RESUME_WINDOW` (7 days) and deletes the
bytes (`spec:712-715`). The member saw every file delivered and loses the album. A transient store error is enough; no
crash is needed.

Failure sequence two (false failure). The callback commits and renames the data out of `tus/`, and the 204 is lost on a
residential link, the common case tus exists for. The client HEADs. The `.info` is gone, so tusd answers 404
(`filestore.go:130-137`), or the wrapper answers 404 because the submission has moved to `checking` (`spec:332-335`).
On a 4xx the client drops its fingerprint and creates a new upload (`lib/upload.js:679-697`). `PreUploadCreateCallback`
refuses a started file with 409 (`spec:930`), and 409 is retryable (`lib/upload.js:1080`), so the client retries through
`retryDelays` and then fails. The page shows a received file as failed, and a member who re-uploads makes a duplicate
in flight (see B2).

Failure sequence three (lost creation response). The POST succeeds and the 201 is lost. The client retries the POST, the
callback answers 409, and the file can never be sent. The same holds for a member on a second device or with cleared
storage, since resumption rests on `storeFingerprintForResuming` (`spec:378`). The submission holds a session slot and
its reservation until expiry.

"Has not started" (`spec:353`) is also undefined. tusd's create truncates: `os.OpenFile(path,
os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerm)` (`filestore.go:323-324`). So any rule that lets a create through while
bytes exist destroys them.

Proposed fold:

1. Make the finish step idempotent and recoverable. Rename the data file into `submissions/` first, then commit
   "received" (and, for the last file, `checking`) in one `Update`. A finish that finds the data already at its
   destination only commits.
2. In the wrapper, answer HEAD from the store for the caller's own file. A received file answers 200 with
   `Upload-Offset` and `Upload-Length` both equal to its size, so the client reports success. A `tus/` upload whose
   offset equals its length on an unreceived row runs the finish step before answering.
3. At startup, before the listener opens, run the finish step for every complete `tus/` upload on an unreceived row.
4. Define "started" as "tusd's `GetUpload` for the id succeeds", never as a store flag.
5. Have the page always pass `uploadUrl` set to the tus base path plus the file id, which the server already makes
   deterministic (`spec:351-355`). `tus-js-client` then resumes with a HEAD first (`lib/upload.js:432-437`) and creates
   only after a 404. That covers a lost 201, a second device, and cleared storage, and it makes the fingerprint store
   optional.
6. Add acceptance criteria: a fault between the last byte and the finish commit, then a HEAD, ends with the file
   received; a response dropped after the commit ends with the client reporting success; a lost 201 ends with the file
   received through `uploadUrl`.

### B2. The duplicate refusal cannot stop two identical albums racing to `accepted`

Location: `spec:530-536`, `:586`, `:592-595`, AC22 at `spec:1026-1028`.

The check reads "an `accepted` submission in the same target collection", and "the check and the move to the next state
commit in one write transaction" (`spec:532-533`). On the main route, the next state after the gate is `matching`
(`spec:603`), not `accepted`.

Failure sequence. Members A and B upload the same album. A's gate passes; no `accepted` twin exists, so A moves to
`matching`. B's gate passes; A is not `accepted` yet, so B moves to `matching`. Both probes return `ActionApply`, and
both move to `accepted` (`spec:622`). Rung 9 admits both, and the library holds the album twice. AC22's "exactly one is
accepted" contradicts the mechanism the spec gives, so the implementer must invent the rule. The same gap opens on
three other paths that enter `accepted` without the check:

- `route_unmatched` from `pushed-back` sends no bytes again and runs no gate (`spec:586`), so two pushed-back copies
  both land in the unmatched collection.
- The accept-time re-validation reroutes a submission to another collection (`spec:592-595`) without re-checking that
  collection.
- `retry_match` from `pushed-back` re-enters `matching` past the gate.

Proposed fold: run the authoritative check inside the one `Update` that enters `accepted`, on every path into it (probe
`ActionApply`, the unmatched route's gate pass, `route_unmatched`, and after a reroute against the new collection). It
reads every submission in the target collection whose state means "claims this album": `accepted` at rung 7, with rung 9
adding its admission states to that set. Keep the gate-pass check as an early refusal so a member learns sooner, reading
every post-gate non-terminal state. Rung 9 owes the same check under its admission transaction once `tracks` carries the
audio key (`spec:535-536` already routes the extension). A partial unique index on `(collection_id, album_key)` would
let the database enforce the rule, but its `WHERE` clause would need a rebuild each time rung 9 adds a state, so the Go
check is the cheaper home. Extend AC22's fixture to the race through `matching` and the race through `route_unmatched`.

## Majors

### M1. Re-running the gate after a crash rejects a good album

Location: `spec:404-406`, `:483-485`, `:707-708`, AC26 at `spec:1039`.

Members are "extracted through the staging root, each opened with `O_EXCL`" (`spec:484-485`), and "the archive itself is
deleted" (`spec:405`). The spec gives no order between that deletion, the member rows, and the verdict commit. Startup
recovery re-queues every `checking` submission (`spec:707-708`).

Failure sequence. `serve` stops mid-extraction (a deploy restart, OOM, a crash). On restart the gate runs again, its
first `O_EXCL` open meets the member it already wrote, the open fails with `EEXIST`, and the album is rejected
`could-not-check` or `unpack-failed`. If the stop came after the archive's deletion and before the verdict commit, the
re-run finds no archive and rejects it as well. A good multi-gigabyte album is rejected for a server fault, and the
member must upload it again. Member rows written per member before the verdict would also duplicate on the re-run.

Proposed fold: make the gate idempotent per submission. Extraction writes into a per-attempt directory
(`submissions/<id>/x/`) that the gate removes before it starts. Member rows are written only in the verdict `Update`.
The archive is deleted only after the verdict commits, by the gate's last step or by the cleanup in m1. Extend AC26's
fixture with a half-extracted ZIP and an archive whose extraction finished but whose verdict did not commit.

### M2. Four non-terminal states have no exit, so reservations and session slots leak

Location: `spec:254-266`, `:630-632`, `:688-689`, `:707-719`; tusd `pkg/handler/unrouted_handler.go:1195-1225`.

A reservation counts while a submission "is not terminal and not yet admitted" (`spec:392-394`), and the session cap
counts submissions in `uploading` (`spec:416-417`). Four states can hold both indefinitely:

- **`stalled`.** No action or sweep leaves it. Recovery re-queues only `checking` and `matching` (`spec:707-708`), so a
  stalled submission is never run again. If the worker is still running when the sweep marks it, the spec does not say
  whether the worker's verdict may leave `stalled`. A compare-and-set on `checking` would drop the verdict and strand
  the submission forever.
- **A hung gate.** The gate has no deadline, and `DUBPLATE_INTAKE_CHECK_WORKERS` defaults to 1 (`spec:452`). One file
  that hangs the decode holds the only worker. Every later submission then waits in `checking`, stalls, and mails the
  admins once each, and intake is down until a restart.
- **`match-unavailable`.** The only action is retry (`spec:631`). No hold expires it (the sweep names only `uploading`
  and `pushed-back`, `spec:712-714`), and a member cannot discard it.
- **`uploading` with no live tus upload.** `cancel_submission` is valid only from `pushed-back` (`spec:630`, `:688-689`).
  Cancelling in `uploading` goes through tus DELETE (`spec:362-363`), and tusd's `DelFile` first calls
  `GetUpload` and answers its error (`unrouted_handler.go:1221-1225`). A file never started or already received has no
  upload to delete. A member who closes the tab after `create_submission`, or between two files, cannot cancel.

Failure sequence for the last case: a member starts three batches and closes the tab each time. With
`SESSIONS_PER_MEMBER` at 3 (`spec:560`), they are refused `too-many-sessions` for 7 days, and the three reservations
count against their cap and the floor for that whole window.

Proposed fold:

1. Make `cancel_submission` valid from `uploading`, `pushed-back`, `match-unavailable`, and `stalled`. From
   `uploading` it also terminates any live tus upload of the submission.
2. Run the gate under a per-album deadline derived from the declared size and the measured throughput. Hitting the
   deadline rejects the album `could-not-check`. A stall is then only a slow queue, never a hang.
3. Let the worker's verdict leave `checking` or `stalled`, and let recovery re-queue `stalled` too.
4. Give `match-unavailable` the pushed-back hold, and have the sweep re-probe it each interval while MusicBrainz is
   reachable, so an outage resolves without the member.
5. State that every transition is a compare-and-set on the current state inside one `Update`, and that a worker whose
   compare-and-set fails discards its result.

### M3. A restore leaves rows whose staged bytes are gone, and an `accepted` one blocks re-upload forever

Location: `spec:707-716`; `internal/backup/doc.go:4-7`, `:20-26`; `cmd/dubplate-server/serve.go:397`.

The backup sends "a verified copy of the store to snapshots/<stamp>.db" and syncs only the library tree
(`internal/backup/doc.go:4-7`). Staging is never backed up. A restore replaces the store, and the restore reconciler
compares only tracks against library files (`internal/backup/reconcile.go:120-133`).

Failure sequence. A restore onto a fresh box, or onto a box whose staging was lost, brings back submissions in
`accepted`, `pushed-back`, `checking`, and `uploading` with no bytes behind them. The `accepted` rows never move at rung
7, hold their reservations, and fail rung 9's admission. They also match the album key of any re-upload of the same
album, so `already-in-library` refuses the member's attempt to send again the album the restore lost (`spec:532-533`).
In the other direction, a restore to an older snapshot on the same box leaves `submissions/<id>/` directories with no
row. The sweep removes only orphaned `tus/` files (`spec:714-715`), so those bytes stay on the music volume for good,
counted against the floor.

The same divergence follows a crash in the window m1 describes, and a manual cleanup of staging.

Proposed fold: an intake reconciliation at every `serve` start, before the listener opens, beside recovery.

- A non-terminal submission whose received files are missing from staging moves to `expired`, with a `reason` naming
  the missing bytes, and its reservation ends.
- A `submissions/` directory with no row, or with a terminal row, is removed.
- The run logs its counts, so a post-restore start reports what it found.

Add an acceptance criterion: a store holding an `accepted` and a `checking` submission over an empty staging directory
starts with both expired, and a re-upload of the same album is not refused.

### M4. Nothing makes `serve` the only process on the store and staging; the spec's premise misquotes the store

Location: `spec:318-321`, `:707-708`; `internal/engine/store/open.go:117-130`, `:201-219`.

The spec rests `memorylocker` and recovery on "One `serve` process holds the staging directory, because
`store.OpenWrite` takes the store's single-writer lock" (`spec:319-320`). The code does not say that. `OpenWrite` opens
the store "holding its lock shared until the handle closes" (`open.go:117-118`). `serve` does not call `OpenWrite` at
all: it calls `store.OpenServe` (`cmd/dubplate-server/serve.go:391`), which also holds "the lock shared until it closes"
(`open.go:201-203`). `serve` takes no instance or kind lease: no call to `TakeInstanceLease` or `TakeKindLease` exists
under `cmd/` or `internal/` outside the store package.

Failure sequence. While the unit runs, an operator starts a second `serve` against the same state directory, for
example to try a flag on another port. The spec opens the listener after recovery (`spec:707`), so the second process
re-queues every `checking` submission before any bind can fail. Two gates then run on one submission at once. Their
extractions collide (M1), and a good album is rejected, or two verdicts race. Had both processes taken traffic,
`memorylocker`'s per-process locks would also let two PATCHes write one upload.

Proposed fold: `serve` takes a kind lease (`intake`) through `Handle.TakeKindLease` before recovery and holds it for its
life. A kind lease "admits one job of a kind at a time across every process" (`internal/engine/store/doc.go:16-17`). A
second start refuses and names the holder. Correct `spec:319-320` to cite the lease. This also gives rung 9 a way to see
that `serve` owns staging, if its import runs in another process (`spec:634-635`).

### M5. ZIP extraction writes bytes no reservation covers, so the floor can be raced and a quota bypassed

Location: `spec:404-406`, `:431-434`, `:479-480`.

A ZIP's reservation is its declared archive size (`spec:394-395`, `:466`). Extraction is bounded only by the album cap:
the walk "sums every member's `UncompressedSize64` and refuses a sum over the album cap" (`spec:479-480`). Before
extraction the floor check subtracts only "the declared extracted total" (`spec:434`), not other submissions' unwritten
reservations, as the creation check does (`spec:431-432`). The true-up "can raise" the reservation (`spec:405-406`)
with no quota re-check named.

Failure sequence. Extras carry no content check, so a member may zip a 4 GiB file of zeros named `scan.tif` beside one
FLAC. The archive is a few megabytes; that is the reservation. Extraction writes 4 GiB. Three concurrent sessions write
12 GiB against about 12 MiB reserved. Meanwhile other members' uploads passed their creation checks on the assumption
that this submission writes nothing more. Their PATCHes then meet `no-space` mid-transfer, or the volume fills under
rung 9's tag-write copies, which are what the floor protects (`spec:436-440`). The member's cap is also exceeded by the
true-up, unchecked. Honest ZIPs of FLAC barely compress, so the honest case moves little, but the floor is an owner
ruling ("a hard free-space floor always", `spec:35-36`).

Proposed fold: before extraction, take the extracted total as the submission's new reservation in one `Update`. That
`Update` applies the same member-cap, group-cap, and floor arithmetic as creation, others' unwritten reservations
included, and refuses with `over-quota` or `no-space` as a verdict. The true-up after the archive's deletion then only
lowers the reservation. Add an acceptance criterion with a high-ratio ZIP against a member cap and against an injected
`statfs`.

### M6. Staged files are not flushed before the commits that vouch for them

Location: `spec:359-361`, `:484-485`, `:504-520`; `pkg/filestore/filestore.go:220-228`, `:311-313`.

tusd's filestore writes through `io.Copy` with no `Sync` (`filestore.go:220-228`), and its `FinishUpload` is
`return nil` (`:311-313`). The spec adds no flush before it records a file as received, nor before the verdict that
stores `audio_key` and `file_hash`. The store lives on another device (`spec:42-44`), so its commit, which SQLite makes
durable, has no ordering against the music volume's dirty pages.

Failure sequence. A gate passes, and the `accepted` commit is durable. The host then resets (a hypervisor fault or power
loss) before writeback reaches `/dev/sdb`. The last staged files come back truncated. Rung 9 admits them and copies the
gate's `audio_key` and `file_hash` onto `tracks` (`spec:201-202`). The library then holds files whose bytes differ from
their recorded identity, silently. This is unlikely, but the consequence is the worst this lens names.

Proposed fold: `fsync` each staged file and its directory before the `Update` that marks it received, and each
extracted member before the verdict `Update`, on `internal/atomicfile`'s precedent (`tmp.Sync()` at
`internal/atomicfile/atomicfile.go:29`, the directory `Sync` at `:87`). The cost is one flush per file against a
network transfer. Route one line to rung 9: admission re-verifies `file_hash` while its tag-write copy reads the file,
which costs nothing extra.

## Minors

### m1. Cleanup after a terminal transition is not crash-safe

Location: `spec:362-363`, `:547`, `:714-715`.

`rejected` deletes staged files after its commit (`spec:547`), and cancel removes them in the callback (`spec:362-363`).
A stop between the commit and the deletion leaves the files with a terminal row, and the sweep deletes files only for
`expired` submissions and orphaned `tus/` files (`spec:714-715`). The bytes stay forever, counted against the floor.
Fold: the sweep, and M3's startup pass, remove the staging directory of every terminal submission. Deletion then
becomes idempotent cleanup, never part of the transition.

### m2. A long PATCH can finish into a submission that has closed

Location: `spec:332-335`, `:359-361`, `:714-715`.

The wrapper authorizes a PATCH when it starts (`spec:332-335`). With `chunkSize` at `Infinity`, one PATCH carries a
whole file for hours. A cancel from another tab, an expiry, or a sibling file's DELETE can close the submission while
that PATCH runs. The sweep can also delete the open `tus/` file, since its row is no longer open. The finish callback
then marks a file received in a closed submission, or renames into a directory cleanup has removed. Separately,
"when the last file of the submission arrives" (`spec:360-361`) must be decided inside the same `Update` as the mark,
or two concurrent finishes both hand the submission to the worker. Fold: the finish `Update` compares and sets on
`uploading`, decides the last file inside that `Update`, and on a closed submission deletes the data and answers an
error. The worker's queue claims a submission by state. The sweep deletes a `tus/` file only when its row is terminal
or absent and the file has passed an age grace.

### m3. Nothing keeps member paths unique within a submission

Location: `spec:229-230`, `:484-486`, `:932`.

The manifest refuses a duplicate path (`spec:932`), but the ZIP rules apply "the manifest rules above"
(`spec:485-486`), and that list (`spec:461-471`) does not include duplicates. A ZIP's central directory can name one
path twice. Both members then extract, since staged files are named by id (`spec:117`), and two `submission_files` rows
carry one `rel_path`. Rung 9's layout then writes one over the other. For audio, the (disc, track) check catches it;
for extras, nothing does. Fold: a unique index on `(submission_id, rel_path)`, and a ZIP duplicate name answers
`unsafe-path`. A case-only collision ("Scan.jpg" and "scan.jpg") stays rung 9's injectivity check, which matters from
0.9's home copy on case-insensitive filesystems.

### m4. "Unwritten reserved bytes" is unnamed, and the store holds no per-file progress

Location: `spec:431-432`.

The creation check subtracts "every open submission's unwritten reserved bytes", but the store records only
`received_at` per file, never a partial offset. Reading `.info` sizes inside the write `Update` would put filesystem I/O
in the store's one write transaction. Fold: name the value as `reserved_bytes` minus the sizes of the submission's
received files, zero once the gate has passed. That double-counts in-flight partial bytes, which errs toward refusing,
and the spec should say so.

## Owner forks

None. Each fold above has one answer that follows from the album-is-the-verdict-unit rule, the hard floor, or the
store's existing mechanisms.
