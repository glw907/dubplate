# Rung 7: intake and the uploader

Status: revision 1, drafted 2026-10-10 from the planning brief (`research/2026-10-10-rung-7-planning-brief.md`) and
Geoff's rulings of the same day. It has had no review round yet. It goes through `spec-plan-review`, then to Geoff,
before any plan is written.

This spec covers rung 7 in three passes, 7a, 7b, and 7c. It records the staging location, the submission model, the
transfer, the gate, the quota rules, the user-present match probe, the reject set, the journey, and the uploader page.
It also records how rung 7 sequences against rung 14a, which owns `internal/auth`, `internal/serve`, and
`internal/ops` while its lane runs. Where CH, R6, PS, or the brief already settle a point, this spec cites them rather
than restating them.

Abbreviations follow the brief: BS is the build spec (`specs/2026-09-02-greenfield-server-build.md`), RP the re-plan
(`specs/2026-10-05-rungs-12-16-replan-design.md`), SS the streamer spec (`specs/2026-09-07-streamer-2-0-design.md`),
CH the configuration homes spec (`specs/2026-09-11-configuration-homes-design.md`), R6 the rung 6 spec
(`specs/2026-09-23-rung-6-ops-seam-store.md`), PA the parallel build approach
(`specs/2026-09-03-parallel-build-approach.md`), PS the group-scale principles spec
(`specs/2026-09-25-group-scale-principles-design.md`), VI the visual identity spec
(`specs/2026-09-24-visual-identity-design.md`), W the web surface standard (`specs/2026-09-12-web-surface-standard.md`),
UD the salvage-era uploader design (`specs/2026-08-31-dubplate-uploader-design.md`), OD the organizer spec
(`specs/2026-09-01-dubplate-organizer-design.md`), PCF the reject and uploader string authority
(`research/passb-copy-final.md`), 14A the rung 14a spec (`specs/2026-10-10-rung-14a-admin-api-design.md`), and B the
planning brief. Paths sit under `docs/superpowers/` unless they start at the repo root. Code citations were read on
master at `80f91ca`. No file under `internal/`, `cmd/`, `systemd/`, or `env/` changed between the brief's `e6fdf2b` and
`80f91ca`; the 2026-10-10 dependency sweep (`2af4404`) moved `go.mod` and `web/package.json`.

## Owner rulings (Geoff, 2026-10-10)

1. **The unmatched-album route stays, renamed.** UD called it "Private Recordings" (UD:251-258, PCF section 6).
   "Private" now names a collection's visibility (SS:1004-1005), so the route takes a plain-terms name. This spec
   proposes one under "Rulings for Geoff".
2. **The album page, the library grid and list, filtering, and buy links move to a web pass after rung 9.** None is
   in rung 7. The uploader still captures the store link from the source question, because upload is the only moment
   it exists (PS:757-764).
3. **No default per-member quota, and a hard free-space floor always.** An admin may set a member's cap through the
   override row; nothing caps a member by default.
4. **One uploader sitting, at the end of rung 7.** It is Geoff's design review of the uploader (BS:20-25), at the
   end of 7c, on the workstation (RP:143-144).

## Conductor rulings (2026-10-10)

1. **Staging lives on the music volume.** A read-only check of the box showed `/var/lib/dubplate` on `/dev/sda1`
   (mounted at `/`) and `/srv/music/library` on `/dev/sdb` (mounted at `/srv/music`), two devices. Rung 9's admit is
   a rename out of staging, and CH:472-475 requires "The tus staging directory, the library root, and every collection
   root must share a device". The staging location comes from `internal/config`, with no site path outside it. `serve`
   refuses to start when staging and the library root do not share a device (`st_dev`). "Staging" below adds one
   strengthening to that check, with its reason.
2. **Hook 12 moves to rung 9.** Rung 7 lays out no library file; the sanitizer's first caller is rung 9's layout
   write. SS owes an erratum. The backup guard's per-collection and per-extension counting and the `--max-delete`
   sizing (B item 19) also belong to rung 9, the first writer of catalog tracks.
3. **Rung 7 writes `internal/auth`, `internal/serve`, and `internal/ops` only after 14a merges.** 14a lands first
   and owns those packages while its lane runs. Rung 7's member gate, its styled 404, and its `/api/v1` operations
   are written after 14a's merge, on 14a's merged code. This amends 14A's contested-surfaces rows for `admin.go` and
   `internal/serve` (14A:524-525), which gave rung 7 those regions concurrently. "Sequencing with rung 14a" names the
   seam.
4. **The saved styled-404 patch no longer applies.** `research/patches/2026-09-26-styled-404-fallback.patch` fails
   `git apply --check` on `cmd/dubplate-server/web.go:169` and `internal/serve/serve.go:31`, and it targets
   `web/svelte.config.js`, which does not exist (B section 4). Rung 7 re-derives the styled 404 from the tree as it
   stands after 14a's merge. The patch is reference material only.

## The split

BS bands rung 7 M (BS:798). B's routed items roughly triple the ladder entry (B section 1), so this spec splits it
into three passes, as 14A split rung 14 and re-banded it L (14A:29-31). The 7a/7b/7c split from B section 7 stands,
with one change: 7a's segments are reordered around 14a's merge.

**7a, the intake core.** Segments A and B touch no 14a territory and can run beside 14a's lane. Segment C writes
`internal/auth`, `internal/serve`, and `internal/ops`, so it starts only after 14a merges, and it opens with a rebase
onto master and a full gate.

- Segment A: the pass-opening sweep and the three new dependencies; `internal/reject`; migration `0006`; staging
  configuration and the startup device check; the default collection, root allocation, the no-nesting check, and
  the Go path validator.
- Segment B: the gate as a pure package (policy, caps, the ZIP walk, decode, the audio key, the file hash, the
  completeness pre-checks, the duplicate refusal); the quota and floor arithmetic as store functions.
- Segment C: the member gate; the tus mount and its wrapper; the submission operations under `/api/v1`; the
  checking worker and startup recovery; route inventory rows; the `doctor` rows.

**7b, matching and the uploader.** The user-present probe; the unmatched route's form and routing; the uploader
page; the journey and push-back views; the signed-in landing.

**7c, visual authoring and the sitting.** The photo loader and its uses; the styled 404; 5c's carried items; VI's
open items that need no catalog; the coherence read; Geoff's sitting.

The reason for the reorder: B's 7a order put the member gate and the tus mount (Task 4) before the gate policy (Task
6). Under conductor ruling 3 that would idle the lane until 14a merges. Moving every 14a-dependent task to the last
segment lets segments A and B run while 14a is open.

## Staging

**Hooks:** lands hook 23 for the startup device check (build-constrained files, below). The unit grant, `deploy.sh`,
and the env template are the VPS deployment layer, which hook 23 names neutral. Neutral on hook 24.

### Location

A new owned wall row, `DUBPLATE_STAGING_DIR`, names the staging directory. It is required, has no default, and is
declared in rung 7's own config file (CLAUDE.md, "Facts agents get wrong"). The estate's env template carries it as a
literal, `DUBPLATE_STAGING_DIR=/srv/music/staging`, on the `DUBPLATE_LIBRARY_ROOT` row's precedent
(`env/dubplate-server.env.template`). The commit that registers it retires the `intake` reserved class
(`internal/config/keys.go:50-57`): `DUBPLATE_UPLOADS_DIR` and `DUBPLATE_TRIGGER_DIR` move to superseded rows naming
`DUBPLATE_STAGING_DIR` as the replacement, on rung 5's precedent of retiring a class in the commit that declares its
replacement (PA, the `internal/config/keys.go` edits row).

The staging directory must not lie inside the library root, and the library root must not lie inside it. Two facts
decide that. `internal/backup` syncs the whole library tree to the remote (`internal/backup/doc.go:4-6`), so staged
partials inside it would be backed up and would skew the guard's catalog count. A fresh restore refuses a library root
that holds files (`doc.go:26-28`). Config load refuses either nesting, comparing the two paths after
`filepath.EvalSymlinks` and `filepath.Clean`.

### Layout

```
<staging>/tus/            tusd's filestore: <id> and <id>.info per upload
<staging>/submissions/    one directory per submission, named by its hex id
```

Every file in `submissions/` is named by its submission-file id plus its extension. A member's file name never
reaches the filesystem. It is kept as bytes in the store (below), which settles R6:562-563's non-UTF-8 question for
staging: there is no member-chosen name to sanitize. Every staging write and rename goes through an `os.Root` opened
on the staging directory. `go doc os.Root` reads "If any component of a file name passed to a method of Root
references a location outside the root, the method returns an error", and "symbolic links may not reference a
location outside the root". That is R6:563's "canonicalizes symlinks once": the staging path resolves once when the
root opens, and nothing beneath it is followed out.

### The unit grant

The serve unit runs under `ProtectSystem=strict` with `ReadWritePaths=/var/lib/dubplate/auth
/var/lib/dubplate/store` (`systemd/dubplate-server-serve.service:43`). Rung 7 adds one entry: the music volume's
directory, `/srv/music` on this box, which holds both the staging directory and the library root. It grants that one
directory rather than two separate entries for staging and the library. `rename(2)` fails with `EXDEV` when "oldpath
and newpath are not on the same mounted filesystem", and its manual adds "rename() does not work across different
mount points, even if the same filesystem is mounted on both" (`man 2 rename`). Two `ReadWritePaths=` entries can
become two mount points inside the unit's namespace. One entry over their common parent keeps the staging-to-library
rename inside one mount.

The grant writes the library root's parent from rung 7, though rung 7 writes nothing in the library. Rung 9 needs the
grant to admit, and granting it now is what lets the startup check below prove rung 9's rename from rung 7 onward.

`deploy.sh` pre-creates the staging directory with the service user's ownership. The state-dir agreement leg
(`scripts/check-state-dir.sh`) gains a clause: the template's `DUBPLATE_STAGING_DIR` and `DUBPLATE_LIBRARY_ROOT` both
sit under the serve unit's music-volume grant. That leg and `scripts/check.sh` are edited only after the gate scope
pass (`plans/2026-10-10-gate-scope-pass.md`) merges, since another lane is editing `check.sh` now.

### The startup check

Before it opens a listener, `serve` stats the staging directory and the library root and refuses to start, naming
both paths, when either holds:

- their `st_dev` values differ (conductor ruling 1; `syscall.Stat_t.Dev`);
- their mount ids differ. `statx(2)` with `STATX_MNT_ID` returns "the mount ID of the mount containing the file"
  (`man 2 statx`), available "since Linux 5.8", and `golang.org/x/sys/unix.Statx_t` carries it as `Mnt_id`.

The mount-id clause is the strengthening. Equal `st_dev` values do not prove that a rename works, because a bind mount
of one filesystem keeps its `st_dev`, and `rename(2)` refuses across mount points. A `statx` call that cannot return
a mount id refuses the start too, under L4's rule that a check that cannot run is never a pass
(`research/2026-10-06-slice-1-melodee-intake.md:350-353`).

States the check meets, and what each reports:

| State | Report |
|---|---|
| staging directory absent | start refused: "staging directory <path> does not exist"; the runbook names `deploy.sh` as what creates it |
| staging directory is not a directory, or not writable | start refused, naming the path and the failing operation |
| library root absent | start refused, naming the path |
| `st_dev` differs | start refused, naming both paths and both devices |
| mount id differs | start refused, naming both paths and saying the two must sit in one mount |
| `statx` fails or returns no mount id | start refused, naming the call and its error |
| both match | start continues |

`doctor` gains a row reporting the same facts for staging, the library root, and every collection root that exists
on disk (CH:472-473 names `doctor`). A collection root that does not exist yet reports "not created", because rung 9
creates them; it is not a failure. The message text names a mount, never `ReadWritePaths=`: systemd is the VPS
deployment's concern, and the runbook line for this refusal names the unit grant.

### Platforms (hook 23)

Hook 23 reads: "Code outside the deployment layer assumes no systemd, no Linux-only syscall, and no fixed path. An
OS-specific check, such as rung 7's same-device test on staging and the library, sits in its own build-constrained
file with a counterpart for the other platforms or a stated refusal there" (SS, hook 23). The device check and the
free-space read ("Quota and the free-space floor") are rung 7's two OS-specific pieces. Each sits behind one small
internal function with these files:

| File | Device check | Free space |
|---|---|---|
| `_linux.go` | `st_dev` and `statx` mount id, as above | `unix.Statfs`, `Bavail * Bsize` |
| `_unix.go` with `//go:build unix && !linux` | `st_dev` alone; `golang.org/x/sys/unix` has no `Statx` outside Linux (`GOOS=darwin go doc golang.org/x/sys/unix.Statx`: "no symbol"), and darwin's `syscall.Stat_t.Dev` is an `int32` | `unix.Statfs`; darwin's `Bsize` is a `uint32`, so the product is computed in `uint64` |
| `_windows.go` | a stated refusal: `serve` refuses to start with "the staging device check is not built for Windows yet" | `windows.GetDiskFreeSpaceEx`, the free bytes available to the caller |

The Windows device check is a refusal at 0.8 because no 0.8 deployment runs on Windows. 0.9's home-install plan owes
the counterpart: comparing volume serial numbers through `windows.GetVolumeInformationByHandle`, which
`golang.org/x/sys/windows` carries. The rest of rung 7's intake code is portable by construction: `os.Root`,
`archive/zip`, `crypto/*`, and `path/filepath` are stdlib on every target, tusd's filestore and `memorylocker` use only
`os`, and no intake code names a path outside `internal/config`. The file permissions rung 7 sets (0750 and 0640) are
ignored on Windows, which is harmless. A gate check cross-compiles the intake packages for `darwin` and `windows`
(acceptance criterion 32), so a Linux-only call outside these files fails the build where it would otherwise pass on
the workstation.

## The submission model

**Hooks:** lands hook 1 (the audio key and file hash on submission files), hook 17 (the uploading member on the
submission), hook 21 (the source answer on the submission), and hook 4 (the journey as an event table). Rung 9 copies
each onto its album and track rows at admission. Neutral on hooks 23 and 24: the store is portable SQLite, and this
section adds no platform-specific file.

### The album is the verdict unit

A submission is one album. One failing track rejects the submission, and no partial album is ever accepted
(BS:812-814). A member's batch is several submissions: the page groups them, and the server judges each on its own.
The page treats one ZIP as one album, one set of loose files picked together as one album, and one dropped folder,
its subfolders included, as one album. The source question is asked once per batch, and each album in the review
list may override it (PS:520-523).

### Migration `0006`

14a adds no migration (14A:429-434), and 14b follows rung 7, so rung 7 takes `0006`. It carries R-6d4's keep limit
and adds its tables to the older-store helper's map (`cmd/dubplate-server/olderstore_test.go:28-33`), per STATUS's
"First migration after `0005`". The tables follow the store invariants: STRICT, BLOB paths (invariant 1), `at` and
`kind` NOT NULL on any table with a `kind` column or an `_events` name (invariant 3,
`internal/engine/store/invariant_test.go:329-346`), and a single BLOB `id` primary key (invariant 6, `:397-416`).

- **`submissions`**: `id`; `uploader` (BLOB, NOT NULL, references `users`); `collection_id` (BLOB, NOT NULL,
  references `collections`); `container` (`files` or `zip`); `state` (the journey's current state, written in the same
  transaction as the event that sets it); `source` (one of PS's seven values, NOT NULL); `source_note` (optional text:
  a store name, or the note on "Other"); `store_url` (optional, `https` only); `declared_bytes` and `reserved_bytes`
  (INTEGER, NOT NULL, at least 0); the match proposal (`mb_release_id`, the serialized candidate, and the evaluation
  date) or the entered album fields for the unmatched route (artist, album, year); `album_key` (BLOB, set when the gate
  passes); `created_at` and `updated_at`. The column is named `mb_release_id`, as `albums` names it, because
  invariant 4 refuses a bare `release_id`.
- **`submission_files`**: `id`; `submission_id`; `rel_path` (BLOB, the member's path within the album, under the
  relative-path CHECKs `0003_catalog.sql:15-22` uses); `size`; `role` (`audio`, `extra`, or `archive`);
  `received_at`; `audio_key` (BLOB of 16 bytes, NULL until the gate decodes the file); `header_md5` (BLOB of 16 bytes,
  NULL when the header's value is all zero); `file_hash` (BLOB of 32 bytes); `sample_rate`, `bits_per_sample`,
  `channels`, `samples`; `disc` and `track` as read from the tags; `entered_title` (the unmatched route's form);
  `verdict` (a reject code, or NULL for a pass).
- **`submission_events`**: `id`, `submission_id`, `at`, `kind`, and optional `reason` and `detail`. `kind` is a
  journey state. It carries no SQL CHECK, on `move_intents.kind`'s precedent (`0005_routed.sql`), so rung 9 adds its
  states without rebuilding the table. Go validates `kind` against the closed vocabulary at write, and a test asserts
  that every `submissions.state` equals its latest event's `kind`.

The `collections` table gains `intake_route`, which is NULL, `default`, or `unmatched`, under a unique index, so at
most one collection holds each route. SQLite treats NULLs as distinct in a unique index; the plan quotes the SQLite
page that says so and the `ALTER TABLE ADD COLUMN` rules for a STRICT table before writing the migration.

Why a column and not a settings row: CH puts "routing rules: which collection an upload lands in" in the settings
table (CH:287). A settings value holding a collection id has no referential integrity and could be seeded from the
environment. A nullable column under a unique index names the routed collection with a foreign-key-grade link. CH
owes an erratum that narrows its routing row to rules beyond these two fixed routes.

### Journey states

Rung 7 defines only the states it emits (B D7, the 2026-09-12 refinement that "nothing is exported ahead of a
caller"):

| State | Meaning | Terminal |
|---|---|---|
| `uploading` | created; bytes are moving | no |
| `checking` | every file received; the gate is running or waiting for a worker | no |
| `matching` | the gate passed; the probe is running | no |
| `match-unavailable` | MusicBrainz did not answer; the bytes stay staged | no |
| `pushed-back` | no confident match; the member may retry, take the unmatched route, or discard | no |
| `accepted` | waiting for admission, with a stored proposal or entered album fields | no (rung 9 moves it) |
| `rejected` | the gate failed; per-file reasons recorded | yes |
| `cancelled` | the member cancelled | yes |
| `expired` | idle past its window (uploading or pushed-back) | yes |
| `stalled` | `checking` or `matching` ran past its threshold; the admin is mailed once | no |
| `rerouted` | an event only: the target collection changed at accept (CH:560-565) | n/a |

Rung 9 adds its own states. 14A's version policy makes a new value additive (14A:81-88). No state carries a queue
position, which keeps ruling 10's "`/api/v1` publishes no queue position"
(`research/2026-09-02-greenfield-server-rulings.md:63-65`). A submission waiting for a checking worker stays
`checking`. PCF section 7's "queued" belongs to rung 9's import.

## The transfer

**Hooks:** neutral, including on hook 23 (tusd, its filestore, and `memorylocker` are portable Go, and the wrapper adds
no platform-specific file). Neutral on hook 24: the base path below reads `DUBPLATE_PUBLIC_URL`, which a LAN install
sets to its LAN origin, so nothing here needs a public host.

### tusd, configured closed

Rung 7 adopts `github.com/tus/tusd/v2` v2.10.1 ("Prior art"). The handler is built with these settings, each quoted
from `pkg/handler/config.go` at v2.10.1:

- **CORS off.** `DefaultCorsConfig` sets `AllowOrigin: regexp.MustCompile(".*")` (`:143-145`). The mount sets
  `Cors: &CorsConfig{Disable: true}`; `Disable` "instructs the handler to ignore all CORS-related headers and never
  set a CORS-related header in a response" (`:114-116`). The shell's origin check stays the one CSRF defense
  (`internal/serve/origin.go:107-120`).
- **Downloads off.** `DisableDownload` refuses downloads "by not mounting the GET handler" (`:30-32`).
- **Concatenation off, and not composed.** `DisableConcatenation` refuses POSTs "that use the concatenation
  extension" (`:36-38`). The store composer is built by hand with `UseCore`, `UseTerminater`, and `UseLocker`. It
  never calls `FileStore.UseIn`, which also composes the concater, the length deferrer, and the content server
  (`pkg/filestore/filestore.go`, `UseIn`). With no length deferrer, a creation carrying `Upload-Defer-Length` answers
  `ErrNotImplemented` (`unrouted_handler.go:1454-1455`), so every upload declares its length before a byte moves,
  which the quota reservation needs.
- **The experimental IETF protocol off.** `EnableExperimentalProtocol` stays false (`:26-29`).
- **Forwarded headers ignored.** `RespectForwardedHeaders` stays false (`:61-64`).
- **An absolute base path.** `BasePath` "may specify an absolute URL containing a scheme" (`:21-24`), and with one,
  `absFileURL` returns `basePath + id` without reading `Host` (`unrouted_handler.go:1312-1315`). The base path is
  `DUBPLATE_PUBLIC_URL`, the shell's owned prefix, and `/tus/`, so a `Location` header never echoes a request's
  `Host` header.
- **`MaxSize`** is the album size cap. tusd answers a larger `Upload-Length` with `ErrMaxSizeExceeded`, HTTP 413
  (`unrouted_handler.go:45`, `:1440-1444`).
- **Termination on.** A member may cancel. `PreUploadTerminateCallback` checks ownership and moves the submission to
  `cancelled`.
- **Permissions.** `filestore.New` uses `DefaultDirPerm` 0775 and `DefaultFilePerm` 0664
  (`filestore.go:60-69`). The mount sets `DirModePerm` 0750 and `FileModePerm` 0640 on the result, matching the
  unit's `UMask=0027`.

**The logger.** `Config.Logger` is a `*golang.org/x/exp/slog.Logger` (`config.go:9`, `:59-60`), a distinct type from
`log/slog`'s: `x/exp/slog/logger.go:79` declares its own `Logger` struct. A nil logger becomes `slog.Default()` of the
`x/exp` package (`config.go:154-156`), which writes outside the shell's sink. Issue #1159, "slog: replace
golang.org/x/exp/slog with log/slog", is open since 2024-07-24 (`gh api repos/tus/tusd/issues/1159`, read 2026-10-10).
The mount passes an `x/exp/slog` logger whose handler discards, and dubplate logs from its own wrapper and callbacks
through the shell's logger. A forwarding handler between the two packages would be an adapter written for one
dependency's lag; when #1159 lands, the bump replaces the discard logger with the shell's (the `dependency-upgrade`
skill's rule that a workaround never outlives the bump that fixes it).

**The locker.** `memorylocker`'s locks "will only exist as long as this object is kept in reference and will be
erased if the program exits" (`pkg/memorylocker/memorylocker.go`). One `serve` process holds the staging directory,
because `store.OpenWrite` takes the store's single-writer lock and only `serve` writes submissions. `memorylocker` is
enough, and `filelocker` (with `github.com/tus/lockfile`) is declined.

### The wrapper

tusd offers no hook on PATCH or HEAD (`handler.go:40-80` routes them straight to `PatchFile` and `HeadFile`), so a
rung 7 wrapper sits between the shell and tusd and authorizes every request before tusd sees it:

1. It refuses `X-HTTP-Method-Override`. tusd lets a POST carrying it become any method (`unrouted_handler.go:187-189`
   at the comment "Allow overriding the HTTP method"), which would let a request pass the wrapper as one method and
   run as another.
2. It requires a verified member session (the member gate, segment C). No session answers 401.
3. For a path naming an upload, it parses the id as the hex form of a `store.NewID` id and refuses any other shape
   with 404 before tusd joins it onto a file path. It loads the file row and its submission, and answers 404 unless
   the session's `UserID` is the submission's `uploader` and the file belongs to an open submission in `uploading`.
   An unknown id, another member's id, and a closed submission give the same 404.
4. For PATCH, it checks the free-space floor by `statfs` (below) and answers the `no-space` refusal before tusd
   reads the body.
5. It cuts the shell prefix and `/tus` itself and hands tusd the rest, as 14a's mount does for `api` (14A:117-119).

tusd extends its own read deadline after every successful body read: "Update the read deadline for every successful
read operation" (`unrouted_handler.go:922-929`), through `ResponseController.SetReadDeadline`. That is what keeps a
long PATCH alive past the server's 30-second `ReadTimeout` (`internal/serve/server.go:36`), which by
`go doc net/http.Server` bounds "reading the entire request, including the body". `NewResponseController` needs the
writer to "be the original value passed to the Handler.ServeHTTP method, or have an Unwrap method"
(`go doc net/http.NewResponseController`). The shell's `statusCapture` has `Unwrap` (`internal/serve/serve.go:283`),
and the rung 7 wrapper either passes the writer through untouched or gives its own wrapper `Unwrap`. A test drives a
PATCH whose body arrives over longer than `ReadTimeout` and asserts it completes.

### The callbacks

- **`PreUploadCreateCallback`.** It reads the tus metadata keys `submission` and `file` (the two ids), checks that
  the file row exists, belongs to the session's open `uploading` submission, has not started, and declares a size
  equal to `Upload-Length`, and checks the floor. It returns `FileInfoChanges{ID: <file id hex>, MetaData:
  MetaData{}}`. `ID` "will be passed to the data store", and "all data stores in the github.com/tus/tusd package do
  respect pre-defined IDs" (`datastore.go:61-66`). An empty `MetaData` map stores no client metadata
  (`datastore.go:68-75`). The session's member reaches the callback through `HookEvent.Context`, which "retains the
  same values as the request context" (`hooks.go:9-22`; `context.go:87-89` uses `context.WithoutCancel` "to preserve
  the values").
- **`PreFinishResponseCallback`.** It marks the file received and renames the tus data file into the submission's
  directory through the staging root, then removes the `.info` file. When the last file of the submission arrives, it
  moves the submission to `checking` and hands it to the checking worker.
- **`PreUploadTerminateCallback`.** It moves the submission to `cancelled`, releases its reservation, and removes
  every staged file of the submission.

### The client

The page pins `tus-js-client` 4.3.1 exactly (CLAUDE.md's 4.3.x rule). Each option below is quoted from the library's
`docs/api.md` at `v4.3.1`:

- `chunkSize` keeps its default, `Infinity`, which "means that tus-js-client will try to upload the entire file in
  one request". The docs warn "**Do not set this value**, unless you are being forced to", and name "a tus server or
  proxy with a limit on how big request bodies may be" as a valid reason. A registered setting,
  `DUBPLATE_INTAKE_CHUNK_BYTES`, default unset, lets an adopter behind a body-limited proxy set one. The page reads it
  from the server.
- `parallelUploads` stays 1. A higher value uses "the concatenation extension", which the server does not compose.
- `removeFingerprintOnSuccess` is true. Once a file is received the server moves its data out of `tus/`, so a stored
  fingerprint for a finished upload would resume against a 404.
- `storeFingerprintForResuming` keeps its default, true, so a reloaded page can find an unfinished upload.
- `retryDelays` keeps its default, `[0, 1000, 3000, 5000]`.

The salvage driver's chunk downshift on 502, 504, and 524 (`_salvage/web/src/lib/tus.ts:12-25`) is declined. 524 is
Cloudflare's status, and the tunnel is deleted (CLAUDE.md, Cloudflare). With no default chunk size, there is nothing
to downshift.

## Quota and the free-space floor

**Hooks:** consumes hook 15's quota counters and override rows. Lands hook 23 for the free-space read, which sits in
the build-constrained files under "Platforms (hook 23)". Neutral on hook 24.

### What a reservation is

A reservation is an open submission's `reserved_bytes`. A member's use is their `member_quota.charged_bytes` plus the
`reserved_bytes` of every submission they uploaded that is not terminal and not yet admitted. A submission's
reservation is taken in the write transaction that creates it, from the manifest's declared total, so "two concurrent
uploads cannot each pass the check and together overshoot" (CH:555-557): the store has one writer, and each check and
its insert commit together. Rung 7 never writes `charged_bytes`. Rung 9 moves a submission's bytes into it in the
transaction that admits the album, and drops the reservation in the same transaction.

**Who is charged.** CH:548-551 charges a member for "everything they own". At 0.8 every upload lands in an admin-owned
collection, so a cap counted by ownership would never bind a member. Rung 7 charges the uploading member for what they
upload, so an admin's per-member cap (owner ruling 3) means what it says. CH owes an erratum: the per-member counter
counts what the member added, and the collection-owner charge waits for the first member-owned collection.

**True-up.** When the gate passes, `reserved_bytes` becomes the summed size of the kept files: extracted members for
a ZIP (the archive itself is deleted), or the received files. The true-up can lower a reservation, and can raise it
for a ZIP, whose extracted total is checked against the cap before extraction (below).

**Release.** A `rejected`, `cancelled`, or `expired` submission holds no reservation, because terminal states are not
counted.

### The limits checked at creation

In order, each refusal named by its reject code ("The reject set"):

1. `may_upload` on the member's override row is 0: `uploads-off`. NULL means allowed.
2. The member already has `upload_concurrency_cap` submissions in `uploading` (override, else the
   `DUBPLATE_INTAKE_SESSIONS_PER_MEMBER` setting): `too-many-sessions`.
3. The member's override `quota_bytes` is set and their use plus the declared total exceeds it: `over-quota`. No
   default cap exists (owner ruling 3).
4. A group total cap is set (a `DUBPLATE_INTAKE_GROUP_QUOTA_BYTES` setting, default unset) and `group_quota` plus open
   reservations plus the declared total exceeds it: `over-quota`.
5. The free-space floor below: `no-space`.

### The floor

The floor is a registered setting, `DUBPLATE_INTAKE_FREE_SPACE_FLOOR`, a byte count. On Linux, free space is
`Statfs_t.Bavail * Bsize` on the staging directory, as `internal/engine/store/fsprobe.go:43-50` reads it;
`statfs(2)` defines `f_bavail` as "Free blocks available to unprivileged user". The other platforms' reads are in the
table under "Platforms (hook 23)". The check runs before every write:

- at submission creation, requiring free space minus every open submission's unwritten reserved bytes minus the new
  declared total to stay at or above the floor;
- at each tus creation, and before each PATCH (the wrapper), requiring free space to stay at or above the floor;
- before a ZIP is extracted, requiring free space minus the declared extracted total to stay at or above the floor.

CH:554 gives the floor's reason as the store's write-ahead log. On this box the store sits on another device
(conductor ruling 1), so the floor on the music volume protects the library and rung 9's admission instead: a tag
write copies a whole file through `internal/atomicfile` before its rename. The default is research the plan records
(BS:323-331). The working proposal is 5 GiB, the salvage floor (`_salvage/internal/config/config.go:176-181`), which
holds one maximum-size album's tag-write copies with room left over.

A `statfs` that fails refuses the write with `could-not-check`, never a pass (L4).

## The gate

**Hooks:** lands hook 1 (the audio key and the file hash). Neutral on hook 23: the gate is pure Go over `os.Root`,
`archive/zip`, `mewkiz/flac`, and `tagio`, and adds no platform-specific file. Neutral on the rest.

The gate is a pure function from a staged submission to a verdict, holding no per-run state, so concurrent verdicts
over distinct albums equal their sequential verdicts (Melodee L5, which bears on rung 9's lease as well). It lives in a
new `internal/intake` package. It runs in a checking worker in `serve`, with in-process concurrency bounded by a
`DUBPLATE_INTAKE_CHECK_WORKERS` setting (default 1, so decode cannot starve request handling on a small host; the plan
measures decode throughput on the box's CPU class and records it). The worker takes no store lease, which leaves
ruling 10's per-album lease wholly to rung 9.

### At creation (the manifest)

The create call carries the album's manifest: each file's relative path and size, and the container kind. The server
checks the manifest before any byte moves:

- Every path is relative, local, and NUL-free: `filepath.IsLocal` reports "is within the subtree rooted at the
  directory in which path is evaluated" and "is not an absolute path", and it "is a purely lexical operation"
  (`go doc path/filepath.IsLocal`). A failing path: `unsafe-path`.
- Path depth at most the `DUBPLATE_INTAKE_MAX_PATH_DEPTH` setting: `path-too-deep`.
- File count at most `DUBPLATE_INTAKE_MAX_FILES`: `too-many-files`.
- Declared total at most `DUBPLATE_INTAKE_MAX_ALBUM_BYTES`: `too-large`.
- A file with an audio extension outside the allowlist: `bad-extension`, parameterized by the extension. A file with
  no audio extension is an extra: booklets, scans, and cue sheets "ride along" (PCF section 1), and their bytes count
  toward the album cap.
- At least one audio file, or for a ZIP, exactly one file with a `.zip` extension: `no-audio` otherwise.
- The source answer present and one of the seven values: `missing-source` (PS A9).

### After the last byte (the files)

**ZIP.** The walk reads the central directory without extracting a member, adopting
`_salvage/internal/intake/intake.go`'s shape. `zip.NewReader` refuses a non-local name only when "the GODEBUG
environment variable contains `zipinsecurepath=0`" (`go doc archive/zip.NewReader`), so the walk applies
`filepath.IsLocal` to every name itself. It refuses a symlink member (`FileHeader.Mode()` has `fs.ModeSymlink`), an
absolute, escaping, or NUL-bearing name, and a directory entry with content: `unsafe-path`. It sums every member's
`UncompressedSize64` and refuses a sum over the album cap before writing a byte: `too-large`. That sum bounds
decompression, because the reader stops at the declared size: in `archive/zip`'s `checksumReader.Read`, "if r.nread >
r.f.UncompressedSize64 { return 0, ErrFormat }" (`$GOROOT/src/archive/zip/reader.go:301-303`), and an early end
returns `io.ErrUnexpectedEOF` (`:307-309`). A ZIP that does not open, a member read that errors, and a CRC mismatch
answer `unpack-failed`. Members are then extracted through the staging root, each opened with `O_EXCL`, with the
manifest rules above applied to the member list. A member name is kept as bytes in `rel_path`. A name that is not
valid UTF-8, or whose `FileHeader.NonUTF8` is set, is accepted: `NonUTF8` "indicates that Name and Comment are not
encoded in UTF-8" and "many ZIP readers interpret Name and Comment as whatever the system's local character encoding
happens to be" (`go doc archive/zip.FileHeader`), so no decoding is guessed. Display replaces invalid bytes with
U+FFFD. Loose files arrive as JSON strings, which are always UTF-8.

**Per audio file**, in this order. The first failing check is the file's verdict:

1. **Container and format.** `tagio.Sniff` names FLAC, and the extension's format is in the allowlist, which is the
   collection's `admission_format_allowlist` when set, else the `DUBPLATE_INTAKE_FORMAT_ALLOWLIST` setting, default
   `flac`. The allowlist is bounded by rung 8's tag I/O: registering a format `tagio.SupportedFormats` does not report
   is a startup refusal (BS:808-811). A mismatch: `bad-extension`.
2. **Metadata size.** No Vorbis comment block over `DUBPLATE_INTAKE_MAX_TAG_BYTES` and no picture block over
   `DUBPLATE_INTAKE_MAX_PICTURE_BYTES`: `tags-too-large` and `picture-too-large`. CH:348-349 calls "a FLAC metadata
   block" a possible "bomb", and these caps bound what `tagio.Parse` reads into memory.
3. **Stream format.** From STREAMINFO: sample rate at least `DUBPLATE_INTAKE_MIN_SAMPLE_RATE` (default 44100) and bits
   per sample at least `DUBPLATE_INTAKE_MIN_BITS_PER_SAMPLE` (default 16): `below-cd-quality`. These are the salvage
   constants (`_salvage/internal/serve/upload.go:48-51`) and PCF section 1's "CD quality or better". Bits per sample
   above 24: `unsupported-flac` (see the audio key below).
4. **Decode, the audio key, and the file hash.** One pass over the file: a `crypto/sha256` hash over every byte, and a
   `mewkiz/flac` decode of every frame. `flac.New` "reads and parses the FLAC signature and the StreamInfo metadata
   block, but skips all other metadata blocks" (`flac.go:67-72`). `ParseNext` "parses the entire next frame including
   audio samples" and "returns io.EOF to signal a graceful end of FLAC stream" (`flac.go:308-310`). A frame header's
   CRC-8 and a frame's CRC-16 are checked, and a mismatch is an error ("CRC-8 checksum mismatch", `frame/frame.go:317`;
   "CRC-16 checksum mismatch", `:142`). Any decode error: `corrupt`. After the last frame, the bytes hashed must equal
   the file's size, and the decoded sample count must equal STREAMINFO's `NSamples` when that is nonzero ("A 0 value
   implies unknown", `meta/streaminfo.go:34-37`): `corrupt` otherwise. Zero samples: `empty-track`.

   The audio key is `Frame.Hash` over every decoded frame into one `crypto/md5`. `Hash` "adds the decoded audio
   samples of the frame to a running MD5 hash" and "can be used in conjunction with StreamInfo.MD5sum to verify the
   integrity of the decoded audio samples" (`frame/frame.go:148-152`). It handles 1 to 24 bits per sample and, above
   that, logs "support for %d-bit sample size not yet implemented" and writes nothing (`:175-177`). A 32-bit stream
   would therefore hash to a wrong key without an error, which is why step 3 refuses it before decode. When the
   header's `MD5sum` is nonzero, it must equal the audio key: `corrupt` otherwise. An all-zero header value "is absent,
   never a value" (CH:424-425) and is stored as NULL. The key is always the gate's own, so "recording that it did"
   (CH:423-424) is the non-NULL `audio_key`.
5. **Tags for position.** `tagio.Read` gives `disc` and `track` (Melodee L3: "dubplate's gate keys on (disc,
   track)"). An audio file with no track number: `missing-track-number`. A missing disc number reads as disc 1.

**Per album**, after every file passes:

6. **Positions.** Two files at one (disc, track): `duplicate-track-position`. Within a disc, track numbers with a gap,
   or a count that differs from a `TRACKTOTAL` the tags carry: `incomplete-album`, because "no partial album is ever
   admitted". A two-disc album whose track numbers restart on disc 2 passes unchanged (L3's owed test). No check
   rewrites a tag to pass (L3).
7. **Duplicate refusal.** The album key is SHA-256 over the track count and the audio keys in (disc, track) order,
   CH:433-435's "ordered set of a release's track audio keys with its track count". A submission whose album key
   equals that of an `accepted` submission in the same target collection is refused: `already-in-library`. The check
   and the move to the next state commit in one write transaction, so of two identical albums in flight, the second
   to commit is refused. Duplicates across collections are allowed by design (CH:567-568), so the refusal reads only
   the target collection, which the member can always see; that meets CH:651's visibility rule. Rung 9 extends the
   check to admitted tracks once `tracks` carries the audio key.

**Completeness checks that warn and never block** (Melodee L2 and L9): no picture block and no cover image among the
extras; the album total from tags differs from the count across discs. Each is shown on the page with its own copy.
None admits or rejects on its own.

**A check that cannot run** rejects the album with `could-not-check` (L4): a `statfs` error, a read error on a staged
file, a `tagio` error other than its two sentinel errors, and a store error during the verdict. A test injects a
fault into each and asserts the album is rejected, and runs the same fixture without the fault and asserts it passes,
so the test cannot pass vacuously.

On `rejected`, the gate deletes the submission's staged files and keeps its rows and events.

### Defaults

Each cap's default is researched in the plan, which records its basis where the key is defined (BS:323-331). The
working proposals, each with its basis:

| Setting | Proposal | Basis |
|---|---|---|
| `MAX_ALBUM_BYTES` | 4 GiB | a 79.8-minute 24-bit/192 kHz stereo album is about 5.5 GB raw (9.216 Mbit/s for 4,788 s) and about 3.3 GB at a typical 0.6 FLAC ratio; the salvage 2 GiB (`_salvage/internal/config/config.go:153`) refuses it. The plan measures the ratio on corpus files. |
| `MAX_FILES` | 500 | a multi-disc box set plus scans; the plan checks the largest track counts MusicBrainz lists |
| `MAX_PATH_DEPTH` | 8 | album, disc, and scan subfolders with margin |
| `MAX_TAG_BYTES`, `MAX_PICTURE_BYTES` | 1 MiB and 16 MiB | FLAC's 24-bit block length caps any block at 16 MiB; the plan reads `tagio`'s limits |
| `SESSIONS_PER_MEMBER` | 3 | the salvage value (`config.go:176-181`), enough for a batch to pipeline |
| `FREE_SPACE_FLOOR` | 5 GiB | above |
| `FORMAT_ALLOWLIST` | `flac` | BS:802-811 and the household-format research BS:329-331 asks for, which the plan records within the tag I/O bound |

## The default collection and the unmatched route

**Hooks:** consumes hooks 8 and 15 (the `writer`, `owner`, and `visibility` columns); neutral on hook 22, because
"a rung that creates collections before slice 3 is neutral" (SS:1033-1040). Neutral on hooks 23 and 24.

**The default collection.** No collection row exists (R6:185-186). The write transaction that creates the first
submission creates the default collection when none holds `intake_route = 'default'`: `writer` `server`, `visibility`
`group`, `owner` the earliest-created live admin, `layout_template` empty (the template machinery is rung 9's,
B D9), and the name this spec proposes under "Rulings for Geoff". Its `root_path` is allocated from its id by the
server (CH:322-323), and the no-nesting check runs against every existing root before the row lands (CH:323-325;
R6:181). The Go path validator that mirrors the schema's CHECKs (R6:567, "lands with rung 7's first writer") runs on
every path rung 7 writes to the store: `root_path` and every `rel_path`. No live admin at that moment refuses the
submission with `internal` and a log line; it cannot happen after bootstrap, which adds an admin.

**The unmatched route.** UD's escape for albums MusicBrainz cannot place (UD:251-258) survives under its new name
(owner ruling 1). It is a second server-written, admin-owned, group-visible collection with `intake_route =
'unmatched'`, created the same way the first time a member takes the route. Taking it changes the contract, and the
page says so before upload starts (UD:252-253): no MusicBrainz matching; a short album form (artist, album, year)
and track titles prefilled from file names, all editable. A member reaches it two ways:

- at selection, by choosing it for an album, which shows the form before upload can start; the create call carries
  the entered fields, and a passing gate moves the submission straight to `accepted`;
- from `pushed-back`, by choosing it for the staged album, which shows the same form; the bytes are not sent again.

The gate's checks are the same on both routes. The duplicate refusal reads the route's own collection. Rung 9 admits
an `accepted` submission without a proposal by writing the entered fields. Rung 13's appeal action (OD:326-336) is a
third way out of `pushed-back`, added by rung 13.

**Re-validation at write (CH:560-565).** The transaction that moves a submission to `accepted` re-reads its target
collection. If the row is gone or no longer writable, the submission moves to the route's current collection,
records a `rerouted` event, and the page says where the album will land. At 0.8 no channel retires a collection, so
this path is reachable only by a direct store write, which is how its test reaches it. The member's
default-collection preference and the per-upload override wait for 14b, the first verb that can make a second
collection to pick from (B D9; W7.19's picker hides at one collection).

## Matching: the user-present probe

**Hooks:** neutral, hooks 23 and 24 included (the probe is portable Go over the existing client).

When the gate passes on the main route, the submission moves to `matching` and the probe runs `tag.Album`
(`internal/engine/tag/tag.go:66-67`) over the submission's tracks, with `tagio`'s tag set projected onto
`matcher.ItemFields` and the length from STREAMINFO. The probe is the first production caller of
`musicbrainz.ClassUserPresent`, whose doc reads "a person is waiting on this request"
(`internal/engine/musicbrainz/limiter.go:16-21`), and the first production construction of a `musicbrainz.Client`.
`NewClient` "fails rather than defaulting on a missing UserAgent or Budget" (`go doc ./internal/engine/musicbrainz
NewClient`), so the probe's budget is a registered setting, `DUBPLATE_INTAKE_MATCH_BUDGET`, whose default the plan
derives from `tag.Album`'s request pattern: one lookup per consensus album id, then a search and its fan-out
(`go doc ./internal/engine/tag Album`). The User-Agent's contact is the one 14A:384-385 names, the MusicBrainz contact
at `internal/engine/musicbrainz/client.go:84-93`. The `TagSet` to `ItemFields` projection is new production code; the
harness builds `ItemFields` from captured fields only (`internal/harness/conform/case.go`). The plan quotes `tagio`'s
field names for each `ItemFields` field.

Strict custody is rung 7's only policy (OD:311-313; the policy enum is rung 13's). A proposal is confident when
`Proposal.Action` is `matcher.ActionApply`, the decision "beets takes unprompted" (`go doc
./internal/engine/matcher Action`). Then:

| Probe outcome | State | Stored |
|---|---|---|
| `ActionApply` | `accepted` | `mb_release_id`, the chosen candidate serialized, and the evaluation date |
| any other action, or `tag.ErrNoCandidates` | `pushed-back` | the top candidate's distance for the page's message, no proposal |
| an error matching `musicbrainz.ErrUnavailable` | `match-unavailable` | nothing |
| any other error | `match-unavailable`, logged at error | nothing |

The stored proposal lets rung 9 admit without matching again (B D6). `evalDate` is the probe's wall clock date, stored
beside the proposal, because `tag.Album` takes it "as a parameter and never a clock read".

From `pushed-back` the member may retry (back to `matching`), take the unmatched route, or discard (`cancelled`). From
`match-unavailable` the member may retry. A submission left in `pushed-back` past `DUBPLATE_INTAKE_PUSHED_BACK_HOLD`
(proposal: 7 days, long enough to add a release to MusicBrainz and come back) expires and its files are deleted.

If rung 9's import runs in a process other than `serve`, its limiter will not share the probe's one-request-per-second
wire budget. That is rung 9's to settle; the limiter is in-process (`limiter.go:93-102`).

## The reject set

**Hooks:** neutral, hooks 23 and 24 included.

`internal/reject` holds one enumerable set (B D11). Each reason carries a stable wire code, a short chip label, the
friendly sentence, the next step, whether it blocks, and, for a parameterized reason, the parameter and a canonical
wording with the parameter filled generically. The project site's reject-page generator reads that canonical wording
(`specs/2026-09-25-project-site-design.md:301-306`). A test enumerates the set and fails on any reason missing a field.

Nine reasons and their strings are adopted verbatim from the authority (BS:508-511): `bad-extension`, `symlink`,
`corrupt`, `below-cd-quality`, `too-large`, `unpack-failed`, `unsafe-path`, `no-space`, `too-many-sessions`, with
the wording and labels at `_salvage/internal/reject/reject.go:44-125` and PCF section 4. `symlink` stays in the set
for rung 9's timer path; rung 7's gate files a ZIP symlink member under `unsafe-path`, as the salvage did
(`reject.go:27-31`).

New reasons, drafted to PCF's rules (no em dashes, no "This [thing] couldn't be [verbed]" shape, no engineering
state names, W8.2's leak list) and gated by one `prose-voice-reviewer` read, as 5c-iv's deck was: `missing-source`,
`uploads-off`, `over-quota`, `too-many-files`, `path-too-deep`, `no-audio`, `tags-too-large`, `picture-too-large`,
`unsupported-flac`, `empty-track`, `missing-track-number`, `duplicate-track-position`, `incomplete-album`,
`already-in-library`, and `could-not-check`. The codes are not copy; W8.2 bans "tags" and "FLAC" from member copy
outside the push-back screen's MusicBrainz exception, so the strings for `tags-too-large` and `unsupported-flac` avoid
both words. `bad-extension` keeps its adopted wording and gains its parameter, written to ROADMAP:133-137's shape:
the message names the supported set and suggests a lossless conversion to FLAC where one exists. Either answer to
STATUS's open Fork B question leaves this wording unchanged.

The two completeness warnings carry copy in the same deck and block nothing.

Uploader sessions send no reject email; the page carries every verdict (UD:261-262).

## The journey and the API

**Hooks:** lands hook 4's shape on `submission_events`; neutral on hook 20, which concerns serve events. Neutral on
hook 23: the worker and the sweep are goroutines in `serve`, not systemd timers. Neutral on hook 24, with one
property kept: the sweep marks `stalled` whether or not a mailer is configured, and logs that no mail was sent when
none is, so a LAN install with no mail relay still records the state. At 0.8 a mailer always exists, because magic-link
sign-in needs one.

### Operations

Rung 7's JSON operations join `/api/v1` on 14a's mount, as catalog operations at member privilege, in the contract
document under 14A's additive version policy. They are written after 14a merges (conductor ruling 3), so they follow
14a's merged form: the seam pattern 14a uses for its admin operations (14A:143-150), its session middleware, its
error body with `error.detail`, and its catalog. The plan reads 14a's merged code and names the seam type. The
operations:

- `create_submission`: the manifest, the source answer, the optional note and store URL, the route (`match` or
  `unmatched`), and, for `unmatched`, the album form. Answers the submission id and each file's id, or a refusal
  carrying its reject code.
- `list_submissions`: the caller's own submissions, newest first, for "Your uploads" (PCF section 9).
- `get_submission`: one of the caller's submissions with its state, events, per-file verdicts, and warnings. The page
  polls it every 5 seconds; polling is the contract (UD:131-137).
- `retry_match`, `route_unmatched` (with the form), and `cancel_submission`: the member actions above, each valid
  only from the states named there; any other state answers `conflict`.

Every operation reads the caller from the verified session and returns only the caller's own submissions. Another
member's submission id answers `not_found`, the same answer as an unknown id. No operation takes an uploader, a
collection, a state, a reservation, or a key as a request field; a body naming one is an unknown key, which `opshttp`'s
decoder ignores (14A:171-173).

The store URL must parse as an absolute `https` URL with a host, within a length bound: `invalid_request` with detail
`store_url` otherwise. PS A14 renders only `https` links.

### The tus mount

The tus endpoint is its own subtree, `tus`, registered from a rung 7 file through `Shell.Register`. It cannot live
under `api`, because 14a's `api` subtree terminates every path under its prefix (14A:117-124) and speaks only the JSON
contract.

### The checking worker, recovery, and the sweep

The checking worker runs inside `serve`. At startup, before the listener opens, `serve` re-queues every submission in
`checking` or `matching` (their bytes are staged) and leaves `uploading` submissions to tus resumption. This is
intake's own work, not an ops job, so 14A's routing of `serve`'s job runner and job recovery to a later rung
(14A:158-165) is untouched.

A sweep runs on a registered interval. It moves an `uploading` submission idle past `DUBPLATE_INTAKE_RESUME_WINDOW`
(proposal: 7 days, a weekend away from a slow residential upload plus margin) to `expired`, and the same for
`pushed-back` past its hold. It deletes the expired submission's staged files and any `tus/` file with no open file
row, since "No cleanup is performed" by the filestore (`pkg/filestore/filestore.go:7-8`). It marks a submission
`stalled` when it has sat in `checking` or `matching` past `DUBPLATE_INTAKE_STALL_AFTER`, and mails every live admin
once per stalled submission through `internal/notify`, after the write commits and outside the write callback, as
14A:280-285 sends its notice. PCF section 8 tells the member "The admin gets notified automatically", so that sentence
is true only with this mail. The stall threshold is set from the measured decode throughput.

## The member gate, the landing, and the styled 404

**Hooks:** neutral, hooks 23 and 24 included. The member gate reads the verify-identity seam, so a second sign-in
method at 0.9 (hook 24) reaches it without rework.

These are the `internal/auth` and `internal/serve` edits, all in segment C or 7c, after 14a merges.

- **`RequireMember`** sits beside `RequireAdmin` in `internal/auth/admin.go`. It verifies through the same seam
  (`internal/auth/seam.go:107`), refreshes the cookie after the check passes as `RequireAdmin` does (`admin.go:129-131`),
  and places the verified identity on the context. A page request with no valid session answers `303 See Other` to
  the sign-in page. Unlike an admin page, a member page's existence reveals nothing the public sign-in page does not,
  so the gate sends the visitor where they can act. The tus wrapper and 14a's `api` middleware answer 401 instead,
  because a script reads them.
- **The landing.** The uploader page is the signed-in member's landing page. The consumed-link page's `signed_in`
  state gains a Continue link to it, and the header wordmark targets it (B item 14). The shell's owned prefix root,
  which no subtree can claim (`subtreeNameRefusal` refuses an empty name, `internal/serve/serve.go:107-109`), answers
  a redirect to the landing page from the same `Shell.terminated` edit as the styled 404. 14a's plan records "whether
  rung 7 and a running rung 12 slice may both write `internal/serve`" as an open decision (`plans/2026-10-10-rung-14a-admin-api.md:815-816`);
  the conductor rules it before 7c's task that edits `Shell.terminated`.
- **The styled 404.** `Shell.terminated` answers unclaimed paths with `http.NotFoundHandler` (`serve.go:248-258`).
  7c replaces that with the web error document at status 404, and routes `RequireAdmin`'s refusal through the same
  handler, so a member's 404 on an admin page and a 404 on an unregistered path stay byte-identical (B item 14;
  `admin.go:125-128` calls `http.NotFound` today). A `web-auth-security-reviewer` read covers the task. The task is
  re-derived from the tree after 14a's merge (conductor ruling 4).

## The uploader page

**Hooks:** neutral, hooks 23 and 24 included (static assets embedded in the binary).

7b builds the page in SvelteKit static mode under the visual identity spec, which outranks any styling habit
(CLAUDE.md, Stack rules). The spec requires:

- **W7.19's four surfaces.** The drop target's states (idle, dragover, invalid type, a drop outside the target, the
  `dragleave`-on-child trap) with W6.10's single-pointer equivalent; the file picker as a first-class path, the iPad
  path; a multi-file progress surface that runs for minutes on a residential uplink; and a collection picker that hides
  at one collection, which at 0.8 is always.
- **The source question** asked once per batch, required, with a per-album override in the review list (PS:510-533).
  "Bought as a download" offers the optional store name or link; "Other" offers the optional note.
- **The unmatched route's choice and form**, with its contract stated before upload starts.
- **The journey view** for each album, polling `get_submission`, and **"Your uploads"**, listing in-flight
  resumables first (PCF section 9).
- **The push-back view** with its fix paths: fix the names and retry, add the release to MusicBrainz (a link and a
  one-line how, the push-back screen's W8.2 exception), or take the unmatched route.
- **The empty-library state** pairing a photo with "No albums yet. Upload the first one." (VI:274-275), which every
  box shows until rung 9 admits an album.
- **Copy** from PCF verbatim where PCF covers the surface. Two PCF strings name Navidrome ("Open in Navidrome", "its
  own section in Navidrome"); dubplate replaces Navidrome (CLAUDE.md, Reviewers), so those strings are redrafted under
  the same register read as the new reject reasons. PCF's in-library card belongs to rung 9's admission and to the web
  pass after it.
- **VI throughout**: the two palettes, the inks, the component layer, the hard printed shadows, squared outline badges
  for the per-track chips, and the copy register. No hardcoded color (the static web gates check it).
- **Wake Lock** where supported (PCF section 2's "Keep my screen awake").

Every state is rendered and read at the rendered harness's widths in both palettes before Geoff sees anything
(`see-it-before-geoff-tests-it`). W11.5's fresh-context coherence read is the uploader's gate, at k=3 with 2-of-3
consensus (W:1298-1310). The salvage components (`_salvage/web/src/lib/components/`) are source material under fresh
review, never copies (BS:512-516).

## 7c: visual items and the sitting

**Hooks:** neutral, hooks 23 and 24 included. The photo loader's Go port reads embedded files and adds no
platform-specific file.

- The photo loader as a leaf Go package, porting `tools/photos/check.mjs`'s two modes into a gate leg, and the CREDITS
  generator with name-only lines for pack grantors (`specs/2026-09-25-photo-curation-design.md:236-247`;
  `specs/2026-09-25-project-site-design.md:170-184`).
- The photo column on the login page, error-page photos, and the empty-library photo (VI:252-275).
- The styled 404 (above).
- 5c's carried items (B item 14 and HISTORY:1445-1448, swept by the plan).
- VI's open items that need no catalog, built for a verdict at the sitting rather than as settled: the phone header,
  error alerts as a tinted panel, and motion (VI:356-372). The library list view, filtering, the album page, the
  catalog line's format, and the phone album grid move to the web pass after rung 9 (owner ruling 2).
- W11.5's coherence read and the render-and-read fold.
- **The sitting.** Geoff drives real uploads on the workstation: a FLAC album as files, the same album as a ZIP, one
  reject, one push-back, and the unmatched route, at desktop and phone widths in both palettes, and rules VI's three
  unruled items. The "before" is VI's committed mockups and the 5c pages as they stand (B question 1).

## Moved out of rung 7

| Item | Goes to | Reason |
|---|---|---|
| Hook 12, the portable layout and `cover.jpg` | rung 9 | conductor ruling 2; SS erratum owed |
| The backup guard's counting and `--max-delete` sizing | rung 9 | conductor ruling 2 |
| The template grammar, sanitizer, injectivity check, root-confined library writes | rung 9 | their first caller is the layout write (B D9) |
| The album page, library grid and list, filtering, buy-link rendering, the catalog line | the web pass after rung 9 | owner ruling 2 |
| The default-collection preference and per-upload override | 14b | the first verb that can make a second collection |
| The archive-upload carry from 4a's shadow gap | retired | the shadow is retired (RP:126-130); ZIP uploads stay in scope on their own evidence (PCF:39-41) |
| Capturing MusicBrainz buy links at admission | rung 9 | PS:762 names it a new forward hook there |

## Forward hooks (SS:945-1040)

| Hook | Verdict | Section |
|---|---|---|
| 1, audio key and file hash | lands on `submission_files`; rung 9 copies them to `tracks` | The gate |
| 4, event-shaped tables | lands: `submission_events` carries `at` and `kind` | The submission model |
| 17, the uploading member | lands: `submissions.uploader` from the session alone | The submission model |
| 21, the source answer | lands: required on `submissions` | The submission model |
| 5, 8, 9, 15 | consumed: `UserID`, `writer`, the single writer, the collection row's fields and the quota rows | throughout |
| 12, portable layout | moved to rung 9 | Moved out |
| 18, description and mark | neutral: 14a merges first, so every key rung 7 registers carries a description and a mark | Sequencing |
| 22, integer wire id | neutral: slice 3's migration backfills | The default collection |
| 20, serve events | neutral: submission events are not serve events | The journey |
| 23, no platform assumption in the core | lands: the device check and the free-space read sit in `_linux.go`, `_unix.go` (`unix && !linux`), and `_windows.go` files, Windows' device check a stated refusal until 0.9; a cross-compile check holds it; every other section is neutral and adds no platform-specific file | Staging; Quota |
| 24, an install with no public host | neutral: the tus base path reads `DUBPLATE_PUBLIC_URL`, which may be a LAN origin; the stall sweep works with no mailer; the member gate goes through the identity seam | The transfer; The journey |
| 2, 3, 11, 13, 14, 16, 19 | neutral: each names rung 8, 9, 13, or a slice | n/a |
| 6, 7, 10 | retired (6) or rung 6's (7, 10) | n/a |

## Prior art

Each adopted library is judged on license, quality against the house bar, maintenance, and adoption (CLAUDE.md,
"Build for 0.8"). Versions were read on 2026-10-10.

| Entry | Verdict | License | Quality | Maintenance | Adoption |
|---|---|---|---|---|---|
| `github.com/tus/tusd/v2` v2.10.1 (`go list -m -versions`; proxy time 2026-09-14) | adopt, configured closed | MIT (`LICENSE.txt`: "Copyright (c) 2013-2017 Transloadit Ltd and Contributors") | the protocol's reference server; documented hooks; three defaults unsafe for dubplate (CORS any origin, GET download, `x/exp/slog`), each closed above | five tags in 12 months, v2.9.0 to v2.10.1 (`gh api repos/tus/tusd/releases`); last push 2026-10-07; 85 open issues; #1159 open | 3,892 stars; 71 importers of `pkg/handler` (B section 3); tus-js-client is tested against it; the salvage ran v2.10.0 |
| `tus-js-client` 4.3.1 (`npm view`: `latest` 4.3.1, published 2025-01-16; `next` 5.0.0-pre3, 2026-09-30) | adopt, pinned exactly | MIT (`npm view tus-js-client@4.3.1 license`) | documented options; seven small runtime dependencies | last push 2026-09-30; 77 open issues; 5.0 in pre-release | 2,627 stars; the reference client |
| `github.com/mewkiz/flac` v1.0.14 (proxy time 2026-08-19, still `@latest`) | adopt | Unlicense (`LICENSE`: "released into the public domain") | idiomatic package split; frame CRCs checked; `Frame.Hash` silent above 24 bits, closed by the gate's bit-depth check | v1.0.13 2025-07-11, v1.0.14 2026-08-19; 9 open issues; last push 2026-08-19 | 363 stars; 96 importers (SS prior-art row); SS already ruled "Adopt for rung 7" (SS prior art) and rung 8 routes decode here (`specs/2026-09-07-rung-8-tag-io.md:997`) |
| `github.com/tus/lockfile` | decline | n/a | n/a | n/a | only `filelocker` needs it; one process holds staging |
| salvage `intake.go` ZIP walk | adopt its shape | in-repo | central-directory walk that "never extracts a member" | n/a | n/a |
| salvage `gate.go` (`flac -t`, `metaflac`) | decline | in-repo | runs subprocesses; the serve unit records "serve runs no subprocess" (`systemd/dubplate-server-serve.service:68-69`) | n/a | n/a |
| salvage `journey.go` | decline the structure, keep the state names | in-repo | coupled to the retired ledger | n/a | n/a |
| salvage `reject.go` | adopt the strings verbatim, restructure the set | in-repo | an `iota` set with switches; the wire `label` from PCF:20-23 | n/a | n/a |
| salvage `upload.go` | adopt the session-grouping idea only | in-repo | reads Cloudflare Access identity, retired | n/a | n/a |
| salvage `tus.ts`, `preflight.ts`, `platform.ts`, `wakeLock.svelte.ts`, 13 components | source material | in-repo | predate static mode and VI; the 524 branch declined | n/a | n/a |
| stdlib `archive/zip`, `crypto/md5`, `crypto/sha256`, `os.Root`, `path/filepath` | adopt | BSD-3-Clause | quoted above | Go team | n/a |
| `golang.org/x/sys/unix` (`Statx`, `Statfs`) and `golang.org/x/sys/windows` (`GetDiskFreeSpaceEx`) | adopt; the module is already a direct dependency at v0.49.0 | BSD-3-Clause | n/a | Go team | n/a |
| OpenSubsonic specification, Navidrome, gonic, taglib via wazero, Picard's tables, chromaprint, goose, modernc sqlite | neutral | n/a | n/a | n/a | none bears on intake beyond the store already in use |
| Melodee | consulted (`research/2026-10-06-slice-1-melodee-intake.md`) | MIT | lessons L2, L3, L4, L5, L7, L9, L12 adopted as tests above | n/a | n/a |

Adding `mewkiz/flac` brings `icza/bitio` v1.1.0 (Apache-2.0), `mewkiz/pkg`, and `mewpkg/term`. `go mod tidy` picks
`mewkiz/pkg` at a 2025-04-17 pseudo-version where a 2026-07-03 one exists (B section 3); the sweep weighs that
indirect. tusd brings only `golang.org/x/exp` beyond itself once `filelocker` is not imported (B section 3).

The pass opens with the `dependency-upgrade` sweep. The 2026-10-10 sweep already landed at `2af4404`
(`research/2026-10-10-dependency-sweep.md`), taking `x/sys` v0.49.0, `x/text` v0.43.0, `playwright` 1.64.0, and `vite`
8.3.4, so 7a's sweep is likely one line plus the three adoptions.

## Sequencing with rung 14a

14a lands first. These rows amend 14A's "Contested surfaces with the rung 7 lane" (14A:510-532) for rung 7's lane:

| Surface | Rung 7's rule |
|---|---|
| `internal/auth` (`RequireMember`, the styled 404's hook in `RequireAdmin`'s refusal) | written only after 14a merges, on 14a's merged `admin.go` |
| `internal/serve` (`Shell.terminated`, the prefix-root redirect) | after 14a merges, and after the conductor rules the rung 12 question 14a's plan left open |
| `internal/ops`, `opshttp`, the contract document | rung 7's operations join after 14a merges, as catalog rows and document paths |
| route inventory (PS A2) | 14a lands it; rung 7 adds rows for `tus`, its pages, and its operations |
| config declarations (hook 18) | 14a lands the description and mark first; rung 7's keys carry both |
| `internal/config/keys.go` | rung 7 retires the `intake` class and adds the staging wall row; 14a's wall-row edits are disjoint |
| migration `0006` and the older-store map | rung 7 |
| `cmd/dubplate-server/serve.go`, `doctor.go`, `web.go` tables | rung 7 adds from its own files and rows (14A:526-528) |
| `web/` | rung 7 owns it (PA:210), with 14a's `routes/members`, admin settings route, and API client as 14a's |
| `scripts/check.sh`, `scripts/check-state-dir.sh`, `Makefile` | edited only after the gate scope pass merges |

**The seam.** Segments 7a-A and 7a-B write only rung 7's own packages (`internal/reject`, `internal/intake`), the
store's migration and query files, rung 7's config file, and `internal/config/keys.go`'s `intake` class. Segment 7a-C
opens with 14a merged: rebase onto master, run the full gate, then write the member gate, the tus mount, and the
operations. If 14a has not merged when 7a-B closes, the lane pauses at that green boundary and records it in STATUS.

The lane registry entry is written before 7a launches (PA:173), and 14A's contested-surfaces table gains this
section's rows as an erratum.

## Planning-miss prevention

### Behaviors relied on, quoted

Quoted above: tusd's `MaxSize`, `DisableDownload`, `DisableConcatenation`, `Cors.Disable`, the default CORS origin,
the `x/exp/slog` logger and its nil default, the three callbacks, `FileInfoChanges.ID` and `MetaData`,
`HookEvent.Context`'s values, the absolute `BasePath` and `absFileURL`, the deferred-length refusal without a
deferrer, the per-read deadline extension, `UseIn`'s composition, the filestore's default permissions and its lack of
cleanup, and `memorylocker`'s lifetime; `tus-js-client`'s `chunkSize`, `parallelUploads`, `removeFingerprintOnSuccess`,
`storeFingerprintForResuming`, and `retryDelays`; `mewkiz/flac`'s `New`, `ParseNext`, the CRC-8 and CRC-16 errors,
`Frame.Hash` and its 24-bit limit, and `NSamples`' zero; `archive/zip`'s `ErrInsecurePath` condition, the read bound
at `UncompressedSize64`, and `NonUTF8`; `filepath.IsLocal`; `os.Root`; `http.Server.ReadTimeout` and
`NewResponseController`'s `Unwrap` rule; `rename(2)`'s `EXDEV` across mount points; `statx(2)`'s mount id;
`statfs(2)`'s `f_bavail`; `Statx`'s absence outside Linux and darwin's `Stat_t.Dev` and `Statfs_t.Bsize` types; `windows.GetDiskFreeSpaceEx`; `musicbrainz.NewClient`'s refusals and `ClassUserPresent`; `matcher.Action`. The plan quotes
any behavior it adds, and quotes before writing: SQLite's NULL handling in a unique index and its `ALTER TABLE ADD
COLUMN` rules for a STRICT table; `tagio.Read`'s and `tagio.Parse`'s errors and block-size handling; and each
`tagio` field it projects onto `ItemFields`.

### Privileged fields and every channel that sets them

| Field | Channels at rung 7 |
|---|---|
| `submissions.uploader` (rung 9's `added_by`) | `create_submission`, from the verified session's `UserID` alone; never a request field or tus metadata |
| `submissions.collection_id` | the server, from the route's `intake_route` collection at create; the `route_unmatched` action on the caller's own `pushed-back` submission; the accept-time re-validation |
| `collections.intake_route`, `owner`, `writer`, `visibility`, `root_path` | the lazy creation in rung 7's write path alone; no API sets them at rung 7 (14b may) |
| `submissions.state` and every event | the intake state machine alone; member actions request a transition, which the machine validates against the current state |
| `submissions.source`, `source_note`, `store_url` | `create_submission`, by the member, per album; no channel edits them at rung 7 (after admission only an admin, PS:355-356) |
| `reserved_bytes` | the server, from the validated manifest at create and the gate's true-up; each file's `Upload-Length` must equal its declared size |
| `audio_key`, `header_md5`, `file_hash`, `album_key`, stream fields, `verdict` | the gate alone |
| the match proposal | the probe alone |
| the entered album fields and `entered_title` | `create_submission` on the unmatched route and `route_unmatched`, by the uploader of that submission |
| the tus upload id | `PreUploadCreateCallback`, set to the server-minted file id |
| `member_quota.charged_bytes` | none at rung 7; rung 9's admission |
| `member_overrides` (`may_upload`, `quota_bytes`, `upload_concurrency_cap`) | read only; written by 14b's override endpoints, and before 14b by a direct store write alone |

### States each check meets, and the report each names

| Check | State | Report |
|---|---|---|
| startup device check | each row of the table under "The startup check" | as that table says |
| any rung 7 operation or tus request | no valid session | 401 (`unauthenticated` for operations) |
| member page | no valid session | 303 to sign-in |
| tus request | id malformed (wrong length, uppercase, non-hex) | 404, before tusd |
| tus request | id unknown, another member's, or a closed submission | 404, the same body |
| tus request | `X-HTTP-Method-Override` present | 400 |
| tus create | `Upload-Defer-Length` | tusd's `ErrNotImplemented` |
| tus create | `Upload-Length` differs from the declared size, or over `MaxSize` | refused; the declared-size mismatch as `invalid_request`-shaped tus error, the cap as 413 |
| tus create | file already started | 409, no second upload |
| `create_submission` | manifest empty (no files) | `no-audio` |
| `create_submission` | manifest malformed (bad JSON, negative size, duplicate path) | `invalid_request` with a detail |
| `create_submission` | many bad files (two bad extensions and an unsafe path) | refused; every failing file listed with its own reason, one alert composed by the page (PCF section 4) |
| `create_submission` | source absent, empty, or not one of seven | `missing-source` |
| `create_submission` | store URL malformed or not `https` | `invalid_request`, detail `store_url` |
| `create_submission` | override `may_upload` 0 | `uploads-off` |
| `create_submission` | sessions at cap | `too-many-sessions` |
| `create_submission` | cap set and exceeded; cap NULL | `over-quota`; no cap check |
| any write | floor crossed | `no-space` |
| any write | `statfs` fails | `could-not-check` |
| ZIP | does not open, or a member errors or fails CRC | `unpack-failed` |
| ZIP | symlink, absolute, escaping, or NUL-bearing name | `unsafe-path` |
| ZIP | declared total over cap | `too-large`, before extraction |
| ZIP | non-UTF-8 name | accepted, kept as bytes |
| ZIP | holds no audio file | `no-audio` |
| audio file | each per-file check | its reason as listed under "The gate" |
| audio file | header MD5 all zero | stored NULL; no comparison |
| audio file | header MD5 nonzero and unequal | `corrupt` |
| album | many bad files | `rejected`; each file's own verdict recorded; the first failing check per file |
| album | duplicate (disc, track); gap or total mismatch | `duplicate-track-position`; `incomplete-album` |
| album | album key present in the target collection | `already-in-library` |
| album | any check errors | `could-not-check` |
| probe | each outcome | as the probe table says |
| member action | from a state the action does not leave | `conflict` |
| `get_submission` | unknown id or another member's | `not_found` |
| sweep | idle `uploading` or held `pushed-back` past its window | `expired`, files deleted |
| sweep | `checking` or `matching` past the stall threshold | `stalled`, one mail to each live admin |
| re-validation | target collection gone or unwritable at accept | `rerouted` event, then `accepted` in the route's current collection |
| default collection | no live admin at first creation | `internal`, logged |
| startup device check | built for Windows | start refused with the stated refusal (hook 23) |

### Fixtures

Each acceptance criterion below names the fixture state it needs, so none can pass on an empty fixture.

## Acceptance criteria

1. `DUBPLATE_STAGING_DIR` is a required wall row with no default. Config load refuses it inside the library root and
   the library root inside it. The `intake` reserved class is retired, and setting `DUBPLATE_UPLOADS_DIR` or
   `DUBPLATE_TRIGGER_DIR` is a load error naming `DUBPLATE_STAGING_DIR`.
2. With staging and the library root on one filesystem in one mount, `serve` starts. With an injected `Stat` giving
   different `st_dev` values, and separately an injected `statx` giving different mount ids, and separately a `statx`
   error, `serve` refuses each start naming both paths. Fixture: a temporary directory holding both, and the injected
   functions for the three refusals.
3. The serve unit grants one music-volume entry holding both paths. The state-dir leg fails when the template's
   `DUBPLATE_STAGING_DIR` or `DUBPLATE_LIBRARY_ROOT` leaves that grant, proven by editing a fixture copy of each.
   `deploy.sh` pre-creates the staging directory.
4. Migration `0006` passes every store invariant test, carries R-6d4's keep limit, and adds its tables to the
   older-store map. A test fails when `submissions.state` differs from its latest event's `kind`, proven with a
   fixture submission whose state was written out of step.
5. The first `create_submission` on an empty catalog creates exactly one default collection, owned by the earliest
   live admin, `writer` `server`, `visibility` `group`, with a server-allocated root that passes the no-nesting check.
   Fixture: two live admins with distinct creation times and one tombstoned earlier admin. A second create reuses it.
   A fixture with a pre-existing root that would nest under the allocated one is refused by the check.
6. The Go path validator refuses an absolute path, an empty, `.`, or `..` component, and a NUL, matching the schema
   CHECKs: a table test runs each case through both the validator and a direct insert and asserts the same refusal.
7. The tus mount answers no `Access-Control-*` header on any response, answers a GET on an upload with 405, refuses a
   concatenation POST and an `Upload-Defer-Length` POST, refuses `X-HTTP-Method-Override`, and returns a `Location`
   built from `DUBPLATE_PUBLIC_URL` when the request carries a different `Host`. Fixture: a created upload for the GET.
8. A member's PATCH, HEAD, or DELETE on another member's upload answers 404 with the same body as an unknown id, and
   the other member's file is unchanged. Fixture: two members, each with an in-flight upload holding bytes.
9. A PATCH whose body arrives over longer than the server's `ReadTimeout`, with reads spaced under tusd's
   `NetworkTimeout`, completes. Fixture: a test server with a short `ReadTimeout` and a body reader paced past it.
10. The client metadata of a tus creation never reaches the `.info` file, and the tus id equals the server's file
    id. Fixture: a creation carrying extra metadata keys and a forged `uploader` key.
11. `create_submission` sets `uploader` from the session; a body naming another member's id as `uploader`, a
    collection, a state, or a reservation changes none of them (PS A10). Fixture: two members.
12. A submission without a source, with an empty source, and with an unknown value each answer `missing-source` in
    the set's wording, and a submission with each of the seven values stores it (PS A9).
13. Two concurrent creates whose declared totals each fit under a member's cap but together exceed it: exactly one
    succeeds. The same for the free-space floor, with an injected `statfs`. With the cap NULL, a create far above any
    plausible cap succeeds. Fixture: a member override with `quota_bytes` set, and one with it NULL.
14. A rejected, cancelled, or expired submission stops counting toward use; an accepted one keeps counting. The
    true-up after a ZIP's extraction sets `reserved_bytes` to the extracted total. Fixture: one submission in each
    state for one member.
15. Every reject reason in the set has a code, a label, a friendly sentence, a next step, a blocking flag, and, when
    parameterized, a canonical wording; the nine adopted reasons' strings equal the authority byte for byte. The new
    strings carry a `prose-voice-reviewer` record.
16. The ZIP walk refuses a symlink member, an absolute name, a `../` name, and a NUL name with `unsafe-path`; refuses
    a declared total over the cap before writing any file; and accepts a non-UTF-8 member name, storing its bytes
    unchanged. Fixture: one crafted ZIP per case, plus a valid one that passes.
17. A ZIP member whose actual data exceeds its declared `UncompressedSize64` fails extraction with `unpack-failed`
    and leaves no partial file. Fixture: a ZIP whose header understates one member.
18. The gate rejects a FLAC with a flipped frame byte (`corrupt`), a nonzero header MD5 that differs from the decode
    (`corrupt`), a truncated file (`corrupt`), a 22.05 kHz file and a 12-bit file (`below-cd-quality`), a 32-bit file
    (`unsupported-flac`), and a zero-sample file (`empty-track`). A clean file passes, its `audio_key` equals its
    header MD5, and its `file_hash` equals `sha256sum`'s output. A file with an all-zero header MD5 passes with
    `header_md5` NULL and `audio_key` set. Fixture: generated FLACs for each case, made by the plan's fixture
    generator.
19. A retagged copy of a passing file keeps its `audio_key` and changes its `file_hash`. Fixture: one file and its
    retagged copy.
20. One failing track rejects the album, and the passing tracks of that album are not kept. Fixture: an album of
    three good tracks and one corrupt track.
21. A two-disc album whose track numbers restart on disc 2 passes unchanged; a (disc, track) duplicate is rejected; a
    gap is rejected; a `TRACKTOTAL` mismatch is rejected; no tag is rewritten. Fixture: one album per case.
22. An album uploaded twice to the same collection: the second is refused with `already-in-library`. The same album
    on the unmatched route while the first sits in the default collection is accepted. Two identical submissions
    racing to `accepted`: exactly one is accepted. Fixture: an accepted submission, and two concurrent submissions.
23. Each check forced to error (`statfs`, a staged-file read, `tagio`, the store) rejects the album with
    `could-not-check`; the same fixture without the fault passes.
24. The probe stores a proposal and moves to `accepted` on `ActionApply`; moves to `pushed-back` on another action and
    on `ErrNoCandidates`; moves to `match-unavailable` on `ErrUnavailable`. The probe's requests carry
    `ClassUserPresent`. Fixture: a fake MusicBrainz transport with a strong match, a weak match, no candidates, and an
    outage.
25. From `pushed-back`, `route_unmatched` with a form moves the same submission to `accepted` in the unmatched
    collection without a second upload; `retry_match` returns it to `matching`; `cancel_submission` cancels it and
    deletes its files. Each action from any other state answers `conflict`. Fixture: a pushed-back submission with
    staged files.
26. After a restart with one submission in `checking` and one in `matching`, both resume and reach a verdict. The
    sweep expires an idle upload and a held push-back and deletes their files and orphaned `tus/` files, and marks a
    long-running check `stalled` with one mail to each live admin and none to a removed admin. Fixture: an injected
    clock, two live admins and one removed admin.
27. `get_submission` and `list_submissions` return only the caller's own submissions; another member's id answers
    `not_found`. Every rung 7 operation is in the contract document and the catalog, and validates under 14a's
    contract tests. Fixture: two members with one submission each.
28. A signed-out request for the uploader page answers 303 to sign-in; a signed-in member gets the page. The prefix
    root redirects to the landing page. An unregistered path and a member's request for an admin page answer the
    same styled 404, byte for byte. Fixture: one admin and one member session.
29. The route inventory carries a row with an auth class for `tus`, each rung 7 page, and each rung 7 operation.
30. The uploader page builds W7.19's four surfaces, the source question with per-album override, the unmatched
    route's form, the journey and push-back views, "Your uploads", and the empty-library state. Every state is
    rendered and read in both palettes at the harness's widths, passes the `web/` leg's static and rendered gates,
    and passes W11.5's coherence read. `svelte-reviewer` and `daisyui-a11y-reviewer` read it against VI.
31. `go.mod` requires tusd v2.10.1 and `mewkiz/flac` v1.0.14, `web/package.json` pins `tus-js-client` 4.3.1 exactly,
    and `go list -deps ./cmd/dubplate-server` names no `filelocker` or `lockfile` package.
32. `GOOS=darwin` and `GOOS=windows` builds of `internal/intake` and rung 7's other new packages succeed, and every
    Linux-only call rung 7 adds sits in a `_linux.go` file. On Windows the device check returns its stated refusal,
    proven by a unit test of the refusal function built under `windows` in the cross-compile (it is compiled, not run,
    on the workstation). The plan names where the cross-compile runs (a gate leg or the task's own check) once the gate
    scope pass has merged.

## Owed records

- SS erratum: hook 12 moves to rung 9 (conductor ruling 2).
- CH errata: the routing row narrowed by `intake_route`; the per-member counter counts what the member added; the
  startup refusal supersedes CH:474's copy fallback for staging and the library root (a collection root on another
  device stays a `doctor` condition, which rung 9 handles).
- 14A erratum: its contested-surfaces rows for `admin.go` and `internal/serve` are sequenced by this spec's
  "Sequencing with rung 14a".
- BACKLOG #8: the `tus-js-client` watch item names 5.0.0-pre3 (2026-09-30), not 5.0.0-pre2. This spec does not edit
  BACKLOG.
- B's note that ruling 10 was "not yet in the tree" is discharged: the rulings ledger records it as ruling 10
  (`research/2026-09-02-greenfield-server-rulings.md:48-65`).
- PCF's two Navidrome strings are redrafted under the register read (above).

## Ceiling and clock

The plans derive tasks, segments, ceilings, and clock estimates from B section 7. B's proposals stand as the starting
point: 7a 8.5M (flag 6.8M), 7b 6.5M (flag 5.2M), 7c 7.5M (flag 6.0M), about 22.5M for the rung, and about 16 hours of
execution excluding the sitting and owner pauses. 7a's reorder adds one rebase and one full gate at segment C's
start.

## Rulings for Geoff

1. **Should the unmatched route's collection be named "Not on MusicBrainz"?** Recommendation: yes. It says what the
   collection holds in words a member already meets on the push-back screen, where W8.2 allows "MusicBrainz", and it
   cannot be read as a visibility setting. The button that takes the route reads "Add without a MusicBrainz match".
   Alternatives: "Other music" (plain, but says nothing about why an album is there) or "Unlisted albums"
   ("unlisted" already describes share links in PS). The name is a collection row's field, so an admin can rename it
   once 14b ships the collection page.
2. **Should the default collection be named "Library"?** Recommendation: yes. A client's music folder list shows it
   beside "Not on MusicBrainz" from slice 3, and "Library" is the plain word PCF's copy already uses ("the main
   library"). An admin can rename it at 14b.
