# Rung 7 spec review, lens B: mechanics and feasibility

Target: `docs/superpowers/specs/2026-10-10-rung-7-intake-design.md` at `9fc7c53`. Every claim below was checked
against the tree at `9fc7c53`, the module cache (`github.com/tus/tusd/v2@v2.10.1`, `github.com/mewkiz/flac@v1.0.14`,
`golang.org/x/sys@v0.49.0`, `modernc.org/sqlite@v1.60.1`), the `tus-js-client@4.3.1` npm tarball, `go doc` on Go
1.27.2, or a probe under the session scratchpad (removed after the run). Correctness findings come first, ranked by
consequence; over-ceremony follows. The settled rulings (unmatched route kept and renamed, no default quota plus the
floor, one sitting, staging on the music volume with a same-device check, auth/serve writes after 14a) are not
reopened.

Counts: 0 blocker, 6 major, 10 minor, 2 over-ceremony. One low-stakes OWNER FORK (m3).

## Verified as stated

- **tusd, configured closed** (probe against a handler built exactly as spec:281-307 describes, composer by hand with
  `UseCore`, `UseTerminater`, `UseLocker`). With `Cors.Disable`, no response carried an `Access-Control-*` header,
  including an `OPTIONS` with a foreign `Origin`. GET on an upload answered 405 (`handler.go:59`). `Upload-Defer-Length`
  answered 501 `ERR_NOT_IMPLEMENTED` (`unrouted_handler.go:1454-1455`). An IETF-draft request with
  `EnableExperimentalProtocol` false answered 412 (`:1495-1497`). With an absolute `BasePath`, `Location` was
  `https://music.example.org/app/tus/<id>` for a request carrying `Host: evil.example` and `X-Forwarded-Host`
  (`:1312-1315`). `FileInfoChanges{ID, MetaData: MetaData{}}` produced a `.info` with `"MetaData":{}` and the minted
  id, after a creation carrying a forged `uploader` key. `DirModePerm` and `FileModePerm` gave `-rw-r-----`.
- **The method override is real.** A POST to `/tus/<id>` carrying `X-HTTP-Method-Override: DELETE` terminated the
  upload with 204 (`unrouted_handler.go:190-192`), so the wrapper's refusal (spec:328-330) is load-bearing.
- **No pre-PATCH/HEAD/DELETE authorization hook exists.** `Config` has only the three callbacks the spec names
  (`config.go:65-87`); no `PreAccess`-style hook appears anywhere in v2.10.1.
- **Read-deadline extension and the `Unwrap` rule** (spec:340-347). A server with `ReadTimeout` and `WriteTimeout`
  of 1s and `NetworkTimeout` 2s took a PATCH whose body arrived over 4s: the bare writer and a wrapper with `Unwrap`
  both answered 204 with offset 5000; a wrapper without `Unwrap` lost the connection at 1.2s. `statusCapture` is the
  only `ResponseWriter` wrapper in `internal/` and has `Unwrap` (`internal/serve/serve.go:283`). AC 9 is feasible.
- **`rename(2)` and the mount id** (spec:127-168). In a user namespace, two bind mounts (one per directory, as two
  `ReadWritePaths=` entries make) gave equal `st_dev` (47), different `Mnt_id` (969, 970), and `rename` failed
  "invalid cross-device link". One bind over the common parent gave one mount id and the rename succeeded. `Statx`
  with `STATX_MNT_ID` set the mask bit. `statx`, `statfs`, `renameat`, and `renameat2` are in `@file-system`, inside
  the unit's `@system-service` filter.
- **`archive/zip`**: the `zipinsecurepath=0` condition, the read bound at `UncompressedSize64`
  (`reader.go:301-303`, `:307-309`), and `NonUTF8` read as quoted.
- **SQLite** on the repo's driver (SQLite 3.53.4): `ALTER TABLE ... ADD COLUMN intake_route TEXT CHECK (...)` on a
  STRICT table succeeds and enforces the CHECK; `ADD COLUMN ... UNIQUE` is refused ("Cannot add a UNIQUE column"),
  so the separate unique index is the right shape; under that index a second `default` is refused and NULL rows
  coexist.
- **Platform types**: darwin `Stat_t.Dev` is `int32` and `Statfs_t.Bsize` `uint32`; `Statx` is absent outside
  Linux; `windows.GetDiskFreeSpaceEx` and `windows.GetVolumeInformationByHandle` exist and compile. The spec gives
  `_unix.go` an explicit `//go:build unix && !linux` line and claims no implicit constraint, which is correct.
- **mewkiz/flac** on generated fixtures: for 16-, 20-, and 24-bit files, `Frame.Hash` over every frame equals the
  STREAMINFO MD5, the decoded sample count equals `NSamples`, and a tee'd SHA-256 sees exactly the file size. A
  truncated file errors `unexpected EOF`.
- **Dependencies**: `go list -deps` over handler, filestore, and memorylocker names only `golang.org/x/exp/slog`
  beyond tusd, and no `filelocker` or `lockfile`; mewkiz/flac brings `icza/bitio`, `mewkiz/pkg`, `mewpkg/term`.
- **tus-js-client 4.3.1 defaults** (`lib/upload.js:11-47`): `chunkSize: Number.POSITIVE_INFINITY`,
  `retryDelays: [0, 1000, 3000, 5000]`, `parallelUploads: 1`, `removeFingerprintOnSuccess: false`,
  `overridePatchMethod: false`.

## Correctness findings

### M1 (major): the duplicate refusal runs at the wrong transition, so two identical albums both reach `accepted`

Location: spec:530-536; AC 22 (spec:1026-1028).

The refusal compares the new album key against `accepted` submissions, and it runs when the gate passes. On the main
route a passing gate moves to `matching`, not `accepted`. Album A passes the gate and enters `matching`; identical
album B passes the gate while A is still matching, finds no `accepted` row, and enters `matching` too; both probes
return `ActionApply` and both move to `accepted`. The single-writer argument holds only for the transaction that runs
the check, and that transaction is not the one that creates the `accepted` row. AC 22's race clause fails as
designed. `route_unmatched` changes the target collection, so a check made at gate time also reads the wrong
collection for that path.

Fold: run the refusal inside the transaction that moves a submission to `accepted` (from the probe, from the
unmatched form, and from `route_unmatched`), against the target collection as re-validated there. Keeping an early
check at gate pass for faster feedback is optional. AC 22's race fixture should race the two `accepted` transitions.

### M2 (major): `DUBPLATE_STAGING_DIR` is already reserved for rung 9, so registering it panics

Location: spec:96-102; AC 1 (spec:968-970).

`internal/config/keys.go:62-66` puts `DUBPLATE_STAGING_DIR` in the `pipeline` reserved class (rung 9), beside
`DUBPLATE_INBOX_DIR` and `DUBPLATE_QUARANTINE_DIR`. `AddWallRow` calls `claim`, which panics when a key "was already
claimed by a family other than family" (`internal/config/registry.go:125-139`): "DUBPLATE_STAGING_DIR: already
declared in reserved classes, cannot also declare in owned wall". The spec retires the `intake` class and never
mentions `pipeline`.

Fold: the commit that registers the wall row also removes `DUBPLATE_STAGING_DIR` from the `pipeline` class, and the
spec records that rung 7's staging directory is the one rung 9's reservation meant (rung 9's admit renames out of
it). Whether `INBOX_DIR` and `QUARANTINE_DIR` stay reserved is rung 9's call.

### M3 (major): no lock makes `serve` the only process running intake

Location: spec:318-321 (the memorylocker argument); spec:453-454 ("The worker takes no store lease"); spec:707-710
(startup recovery before the listener).

The spec rests `memorylocker` and the lease-free worker on "`store.OpenWrite` takes the store's single-writer lock".
`OpenWrite` holds the store lock shared (`internal/engine/store/open.go:118-119`), and so does `OpenServe` (`:203`).
A second `dubplate-server serve` on the box (an operator's manual start, a unit restart overlapping a slow stop) opens
the store, passes the device check, re-queues every `checking` and `matching` submission, and starts workers before
its listener bind fails. Two workers then gate one album; the second's `O_EXCL` extraction fails and rejects it
with `could-not-check`, and two sweeps race on the same rows and files.

Fold: `serve` takes `Handle.TakeKindLease("intake")` (an exclusive, non-waiting flock "which admits one job of a kind
at a time across every process", `internal/engine/store/lease.go:166-200`) before startup recovery and before
mounting `tus`, and refuses to start when it answers `ErrLeaseBusy`. Then the memorylocker argument and the
lease-free worker hold as written. Ruling 10's per-album lease stays rung 9's.

### M4 (major): AC 32 cannot pass, because the store package does not cross-compile

Location: spec:200-203 ("the store is portable SQLite"); spec:191-196; hook 23 row (spec:823); AC 32
(spec:1056-1060).

`GOOS=windows go build ./internal/engine/store/` fails on `syscall.Statfs` (`fsprobe.go:43`), `syscall.Stat_t`
(`leaf.go:106`), and `syscall.Flock` (`lease.go:74`). `GOOS=darwin` fails on `Statfs_t.Type` and `Frsize` and an
`int32` `Stat_t.Dev` (`fsprobe.go:56`, `inspect.go:305`, `lock.go:130`). `internal/auth` and `internal/backup` fail
through it. `tagio`, `musicbrainz`, `matcher`, `tag`, `config`, `serve`, and `notify` all build for both. Every rung 7
package that holds the quota arithmetic, the submission writes, or the tus callbacks imports `store`, so "rung 7's
other new packages" cannot build for `windows` or `darwin`.

Fold: correct the "portable SQLite" sentence; scope AC 32 to the packages that import no `store` (the pure gate core,
`internal/reject`, and the leaf package that holds the platform files); and add the store's port (flock, `Stat_t`,
`Statfs`) to the 0.9 hook 23 owed list. The plan should keep the gate core free of `store` imports so the scoped
check means something.

### M5 (major): two memory bombs that the caps do not bound

Location: spec:497-499 and the defaults row (spec:559); spec:475-480; AC 16.

**FLAC metadata.** Step 2 says the tag and picture caps "bound what `tagio.Parse` reads into memory". `tagio.Parse`
allocates and reads every block payload before returning (`internal/engine/tagio/container.go:199-202`:
`payload := make([]byte, length)`; `f.ReadAt(payload, bodyOffset)`), padding blocks included, with no count limit.
A size check made from the parsed `Container` runs after the memory is spent. The 16 MiB picture default is FLAC's
own 24-bit block maximum, so it caps nothing, and neither cap bounds the number of blocks: a 4 GiB file of 16 MiB
PICTURE or PADDING blocks makes `Parse` allocate about 4 GiB.

**ZIP central directory.** `zip.NewReader` materializes every central-directory entry before the walk can count
members. A probe with a 35 MB ZIP of 400,000 empty entries cost 77 MiB of heap in `NewReader` alone. A ZIP at the
4 GiB album cap holds tens of millions of entries. `MAX_FILES` is applied to "the member list", after `NewReader` has
built it.

Fold: before `tagio.Parse`, walk the metadata chain by its 4-byte headers alone (seek past each payload), refusing
a block over its cap and a total metadata size over a small multiple of the picture cap; and pick a picture cap below
the format maximum. Before `zip.NewReader`, read the end-of-central-directory record (and its zip64 form) and refuse
an entry count over `MAX_FILES` with `too-many-files`. AC 16 gains a many-entries fixture, and AC 18 gains a
many-blocks fixture.

### M6 (major): startup recovery rejects a ZIP that was mid-extraction

Location: spec:484 (members opened with `O_EXCL`); spec:707-710; AC 26 (spec:1039-1042).

A deploy restarts `serve`. A ZIP submission in `checking` whose extraction was half done is re-queued. Re-running
extraction meets its own earlier members, `O_EXCL` fails, and the album is rejected with `could-not-check` (or the
`submission_files` rows for the members collide). AC 26's fixture names `checking` but not the half-extracted state,
so it passes vacuously.

Fold: recovery of a `checking` submission first removes every file and row extraction produced (anything not a
received upload) and re-runs the gate from the received files; AC 26's fixture includes a ZIP stopped partway through
extraction.

### m1 (minor): AC 7's "refuses a concatenation POST" fails, because `DisableConcatenation` is inert here

Location: spec:289-294; AC 7 (spec:987-989).

tusd reads `Upload-Concat` only when a concater is composed (`unrouted_handler.go:298-306`), and the spec composes
none. A `Upload-Concat: partial` POST with `Upload-Length` answered 201 and created an ordinary upload; a `final;...`
POST answered 400 for the missing `Upload-Length`. No hole results: the create callback still binds the upload to a
declared file row. AC 7 as written fails its test once.

Fold: the wrapper answers 400 to any request carrying `Upload-Concat`, or AC 7 states that the header is ignored and
the creation is ordinary. Drop the claim that `DisableConcatenation` refuses anything; setting it costs nothing.

### m2 (minor): the >24-bit rationale is wrong, and an 18-bit FLAC is rejected as `corrupt`

Location: spec:500-503, spec:513-517; AC 18 (spec:1014-1019).

mewkiz/flac's frame parser maps bits-per-sample code `111` to an error ("reserved sample size bit pattern (111)",
`frame/frame.go:358-361`), and leaves code `000` ("get from StreamInfo") as 0 (`:341-342`). A flac 1.5.0 32-bit
file fails `ParseNext` with that error, so it does not "hash to a wrong key without an error". An 18-bit file (code
`000`, which `flac -t` passes) fails with "non-zero padding" and would be rejected `corrupt`. The `log.Printf` branch
the spec cites is at `:178-179`, not `:175-177`, and `ParseNext` cannot reach it with a nonzero depth.

Fold: step 3 admits STREAMINFO depths 16, 20, and 24 only (the depths whose frame codes mewkiz decodes and that pass
the CD floor) and answers `unsupported-flac` for every other depth. Correct the rationale and the citation.

### m3 (minor, OWNER FORK): ID3-wrapped FLAC files get misleading reasons

Location: spec:493-496, spec:509; the states table (spec:946).

Two real-world states are unnamed. An ID3v2-prefixed FLAC: mewkiz skips the prefix (`flac.go:163-173`) and
`flac -t` passes it, but `tagio.Sniff` compares the first four bytes to `fLaC` (`internal/engine/tagio/seam.go:83-104`)
and answers `ErrUnsupportedContainer`, so step 1 says `bad-extension` for a `.flac` file. A trailing ID3v1 tag: every
frame decodes and the audio key matches, then `ParseNext` errors "invalid sync-code", so the verdict is `corrupt`.
`flac -t` also fails this file ("LOST_SYNC after processing 88200 samples"), so rejecting matches the salvage gate.

Options: (a) reject both under `corrupt`, mapping "the extension says FLAC and the sniff fails" to `corrupt`, with no
new copy; (b) reject both under a new reason whose next step says to remove the ID3 tag; (c) accept them by ignoring
a 128-byte trailing `TAG` record and a skipped ID3v2 prefix, which leaves the ID3 bytes in the library for rung 9's
tag writes to meet. Recommendation: (a). It is the cheapest and it keeps the library clean; the frequency is low.
Either way, the states table names both.

### m4 (minor): the `unix && !linux` file does not build on four unix targets

Location: spec:183-187.

A probe file with that constraint, calling `unix.Statfs` and reading `Bavail` and `Bsize`, built for darwin, ios,
freebsd, dragonfly, and aix, and failed for openbsd (fields `F_bavail`, `F_bsize`), netbsd (no `Statfs`), solaris and
illumos (no `Statfs_t`). plan9, js, and wasip1 have no file at all. Hook 23 asks for "a counterpart for the other
platforms or a stated refusal there", and a build failure is neither. See O1 for the fold.

### m5 (minor): `filepath.IsLocal` passes paths the store refuses

Location: spec:461-463, spec:477-479; AC 16.

`IsLocal` returned true for `a/../b.flac`, `a//b.flac`, `./a.flac`, `dir/`, `a\x00b`, and `..\..\x.flac`. The
`rel_path` CHECK (`0003_catalog.sql:15-22`) refuses the first four. A manifest path or ZIP member that passes `IsLocal`
and fails the CHECK becomes an insert error, which surfaces as `could-not-check` or `internal` rather than
`unsafe-path`. The zip reader's own insecure-path rule also treats backslash names as unsafe.

Fold: the manifest and the member walk run the Go path validator of AC 6, which subsumes `IsLocal`, and map its
refusal to `unsafe-path`. State whether a backslash name is refused or kept as one component.

### m6 (minor): the absent-path states at config load and on a fresh box

Location: spec:104-108; spec:160-168; AC 1.

Config load compares the paths "after `filepath.EvalSymlinks`", which errors on a path that does not exist. Config
load runs for every verb, so an absent staging directory would fail `doctor`, the verb meant to report it, and would
preempt the startup table's "staging directory absent" row. Separately, the startup check refuses an absent library
root, and `deploy.sh` pre-creates only the state leaves and the unidecode directory (`scripts/deploy.sh:264-299`), not
the library root, so a stranger's fresh install cannot start `serve`.

Fold: the load-time nesting check is lexical (`Clean` plus a separator-aware prefix test); the resolved comparison
runs in the startup check and in `doctor`, after existence. `deploy.sh` pre-creates the library root beside staging.

### m7 (minor): callback races and crash ordering are unnamed

Location: spec:351-363; the states table (spec:925-930).

Three states need a named report:

- **Concurrent creates for one file.** The create callback checks "has not started" but does not say it records the
  start in the same transaction. The filestore creates the data file with `O_TRUNC` (`filestore.go:324`), so two
  concurrent POSTs for one file id both pass and the second truncates the first.
- **A finish after the submission left `uploading`.** A terminate callback deletes "every staged file of the
  submission", including sibling `tus/` files whose PATCHes hold their own locks. Those PATCHes complete on unlinked
  inodes and then run `PreFinishResponseCallback` for a cancelled submission. The callback must re-check the state
  and discard.
- **Crash between the store commit, the rename, and the `.info` removal.** If the data file has moved but the row
  is not marked received, the client's HEAD gets 404, its re-create gets the "already started" refusal, and the
  submission sits in `uploading` until it expires. Name the order and what startup does with each partial state.

### m8 (minor): the error statuses tusd and tus-js-client actually produce

Location: spec:930; spec:351-363.

`tus-js-client`'s `defaultOnShouldRetry` retries any 409 (`lib/upload.js:1076-1081`), so answering "file already
started" with 409 spends all four retries before failing. Use a non-retried 4xx such as 400 or 403. A callback that
returns a plain Go error reaches the client as 500 with `err.Error()` as the body (`sendError`,
`unrouted_handler.go:1288-1291`), leaking internal text. Every refusal from a callback must be a `handler.Error`
built with `handler.NewError`.

### m9 (minor): the probe's budget is a lifetime cap per `Budget`

Location: spec:603-611.

`Limiter.Budget(class, limit)` returns a budget "spending at most limit requests" (`limiter.go:122-128`), and
`reserve` refuses once `spent+1 > limit` (`:369-374`). A `Client` holds one budget (`client.go:91-95`). A single
`Client` built at startup with `DUBPLATE_INTAKE_MATCH_BUDGET` would stop matching after that many requests, and
every later probe would land in `match-unavailable`. Fold: one `Limiter` per `serve` process, and a fresh `Budget`
(`ClassUserPresent`) and `Client` per probe run.

### m10 (minor): the discard logger hides tusd's own failures, and the `os.Root` claim is too wide

Location: spec:309-316; spec:119-123.

With the discard logger, tusd's `BodyReadError`, its `NetworkControlError` warning (the only trace of a missing
`Unwrap`, which the probe showed as a silent dropped connection), and its `InternalServerError` message go nowhere.
An `x/exp/slog` text or JSON handler at warn level writing to the shell's own output stream is not an adapter and
keeps them; failing that, the wrapper logs every 5xx it passes through. Separately, "every staging write and rename
goes through an `os.Root`" is false for tusd: the filestore writes `tus/<id>` and `.info` with `os.OpenFile` on joined
paths (`filestore.go:110`, `:324`), and `tagio` opens by path. Reword: rung 7's own writes go through the root, and
tusd's are safe because the wrapper admits only well-formed hex ids that the callback minted.

## Over-ceremony

### O1: darwin and Windows implementations with no 0.8 caller

Location: spec:175-196; AC 32. Cost: two extra platform files, their tests, a cross-compile leg's wiring, and a
Windows refusal unit test that is compiled and never run.

Hook 23 accepts "a stated refusal there". The repo rule is that nothing is built ahead of its caller, and no 0.8
deployment runs off Linux; M4 shows the store cannot run there anyway. Fold: one `_linux.go` file and one
`//go:build !linux` file that returns the stated refusal for both the device check and the free-space read. 0.9's
home-install plan writes the darwin and Windows bodies with its caller. This also closes m4. A
`GOOS=windows go vet` of the gate core alone is cheap and can stay.

### O2: `DUBPLATE_INTAKE_CHUNK_BYTES` has no caller

Location: spec:370-374. Cost: a registered setting, its description and mark, a server-to-page read path, and a test.

The setting exists for "an adopter behind a body-limited proxy". The tunnel is deleted and no 0.8 or 0.9 deployment
puts a body-limited proxy in front of `serve`. Fold: drop it, and file it as a forward hook for the first deployment
shape that has such a proxy.
