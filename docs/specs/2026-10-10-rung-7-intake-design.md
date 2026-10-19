# Rung 7: intake and the uploader

Status: revision 2, 2026-10-10. Revision 1 (`9fc7c53`) was drafted from the planning brief
(`research/2026-10-10-rung-7-planning-brief.md`) and reviewed through `spec-plan-review` under four lenses (contract,
mechanics, integrity, consistency; `research/2026-10-10-rung-7-intake-review-*.md`, `01a1b99`). Revision 2 folds those
reviews and Geoff's rulings of 2026-10-10, and a second fold takes the verification read
(`research/2026-10-10-rung-7-intake-fold-verification.md`); both folds are recorded in
`research/2026-10-10-rung-7-intake-fold.md`. It is agent input: when `intake`'s turn comes, a one-page design goes to Geoff, and this spec carries the detail behind it.

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
planning brief. Paths sit under `docs/` unless they start at the repo root. Code citations were read on
master at `80f91ca` and re-checked at `01a1b99`; tusd at v2.10.1 from the module cache.

## Owner rulings (Geoff, 2026-10-10)

1. **The unmatched route stays, named "Unlisted music".** UD called it "Private Recordings" (UD:251-258, PCF section
   6), and "private" now names a collection's visibility (SS:1004-1005). Geoff chose "Unlisted music" because what a
   member sends "might not all be in album format". Uploader copy says "music" or "recording" wherever "album" would
   be wrong for what a member sends. The button that takes the route reads along the lines of "Add as unlisted music".
2. **The album page, the library grid and list, filtering, and buy links move to a web pass after rung 9.** None is
   in rung 7. The uploader still captures the store link from the source question, because upload is the only moment
   it exists (PS:757-764).
3. **No default per-member quota, and a hard free-space floor always.** An admin may set a member's cap through the
   override row; nothing caps a member by default.
4. **One uploader sitting, at the end of rung 7.** It is Geoff's design review of the uploader (BS:20-25), at the
   end of 7c, on the workstation (RP:143-144).
5. **The default collection is named "Library".**
6. **Loose tracks: the kind now, matching later.** At upload a member chooses "album" or "separate tracks". An album
   keeps the album as its verdict unit. Separate tracks give each track its own verdict, and each passing track lands
   as unlisted music with no MusicBrainz match ("Separate tracks", below). Matching loose tracks against MusicBrainz
   recordings waits until after 0.8; the matcher and its conformance corpus stay album-only.
7. **A rejected submission is final; the member re-sends it whole.** PCF's "Multi-file path retry re-sends only the
   failing files" (`passb-copy-final.md:124-126`) is owed a redraft.
8. **A FLAC file carrying ID3 data gets its own reason.** A new reason, `id3-tags`, tells the member the files carry
   ID3 tags and how to strip them. It fires for an ID3v2 prefix and a trailing ID3v1 record alike (the gate, step 1
   and step 4).

## Conductor rulings (2026-10-10)

1. **Staging lives on the music volume.** A read-only check of the box showed `/var/lib/dubplate` on `/dev/sda1`
   (mounted at `/`) and `/srv/music/library` on `/dev/sdb` (mounted at `/srv/music`). Rung 9's admit is a rename out
   of staging, and CH:472-475 requires "The tus staging directory, the library root, and every collection root must
   share a device". `serve` refuses to start when staging and the library root do not share a device and a mount.
2. **Hook 12 moves to rung 9.** Rung 7 lays out no library file; the sanitizer's first caller is rung 9's layout
   write. The backup guard's per-collection and per-extension counting and the `--max-delete` sizing (B item 19) also
   belong to rung 9, the first writer of catalog tracks.
3. **Rung 7 writes `internal/auth`, `internal/serve`, and `internal/ops` only after 14a merges.** This amends 14A's
   contested-surfaces rows for `admin.go` and `internal/serve` (14A:524-525). "Sequencing with rung 14a" names the seam.
4. **The saved styled-404 patch no longer applies.** `research/patches/2026-09-26-styled-404-fallback.patch` fails
   `git apply --check` (B section 4). Rung 7 re-derives the styled 404 from the tree after 14a's merge.
5. **A check that cannot run keeps the upload.** After the bytes arrive, `could-not-check` holds the submission in
   `check-unfinished` with its bytes and reservation, matching PCF section 3's ratified "Your album is already uploaded
   and waiting; no need to send it again" (`passb-copy-final.md:108-112`). It is never a pass (Melodee L4).
6. **W8.2 gains an exception for format names in reject reasons.** A member must learn a format's name to convert a
   file, the same reason the push-back screen may say MusicBrainz. W owes the erratum.
7. **Platform files: Linux plus one stated refusal.** The OS-specific code ships as one `_linux.go` file and one
   `!linux` file that refuses with a stated message, which hook 23 permits. `internal/engine/store` does not build for
   darwin or windows today, so the cross-compile check covers only packages that import no `store`. Porting the store
   is 0.9 work.
8. **No chunk-size setting.** No 0.8 or 0.9 deployment puts a body-limited proxy in front of `serve`.
9. **The accept-time re-route moves to 14b**, the first verb that can retire or narrow a collection.

## The split

BS bands rung 7 M (BS:798). B's routed items roughly triple the ladder entry (B section 1), so this spec re-bands rung
7 as L and splits it into three passes, as 14A did for rung 14 (14A:29-31). 7a's segments are reordered around 14a.

**7a, the intake core.** Segments A and B touch no 14a territory and can run beside 14a's lane. Segment C writes
`internal/auth`, `internal/serve`, and `internal/ops`, so it starts only after 14a merges, with a rebase onto master
and a full gate.

- Segment A: the pass-opening sweep and the three new dependencies; `internal/reject`; migration `0006`; staging
  configuration (the `intake` and `pipeline` class edits) and the startup device check; the default and unlisted
  collections, root allocation, the no-nesting check, and the Go path validator.
- Segment B: the gate as a pure package (policy, caps, the ZIP walk, decode, the audio key, the file hash, positions,
  the separate-tracks verdicts); `tagio`'s block bound; the quota and floor arithmetic as store functions.
- Segment C: the member gate; the tus mount and its wrapper; receipt; the submission operations under `/api/v1`; the
  intake lease, the startup pass, the checking worker, and the sweep; route inventory rows; the `doctor` rows.

**7b, matching and the uploader.** The user-present probe; the unlisted route's form; the uploader page with resume;
the journey and push-back views; the signed-in landing.

**7c, visual authoring and the sitting.** The photo loader and its uses; the styled 404; 5c's carried items; VI's
open items that need no catalog; the coherence read; Geoff's sitting.

B's 7a order put the member gate and the tus mount before the gate policy; under conductor ruling 3 that idles the
lane until 14a merges, so every 14a-dependent task moves to the last segment.

## Staging

**Hooks:** lands hook 23 for the startup device check and the free-space read. The unit grant, `deploy.sh`, and the
env template are the VPS deployment layer, which hook 23 names neutral. Neutral on hook 24.

### Location

A new owned wall row, `DUBPLATE_STAGING_DIR`, names the staging directory. It is required, has no default, and is
declared in rung 7's own config file (CLAUDE.md, "Facts agents get wrong"). The estate's env template carries it as a
literal, `DUBPLATE_STAGING_DIR=/srv/music/staging`, on the `DUBPLATE_LIBRARY_ROOT` row's precedent.

The name is already reserved: `internal/config/keys.go:62-67` lists it in rung 9's `pipeline` class, and `claim`
panics on a cross-family claim (`internal/config/registry.go:125-139`). The commit that registers the wall row
therefore makes two class edits. It removes `DUBPLATE_STAGING_DIR` from `pipeline`, because rung 7's staging directory
is the one rung 9 meant: rung 9 admits by renaming out of it. Whether `DUBPLATE_INBOX_DIR` and
`DUBPLATE_QUARANTINE_DIR` stay reserved is rung 9's call. It also retires the `intake` class (`keys.go:50-57`):
`DUBPLATE_UPLOADS_DIR` and `DUBPLATE_TRIGGER_DIR` move to superseded rows naming `DUBPLATE_STAGING_DIR`, on rung 5's
precedent of retiring a class in the commit that declares its replacement (PA, the `keys.go` edits row).

Staging must not lie inside the library root, nor the library root inside staging. `internal/backup` syncs the whole
library tree (`internal/backup/doc.go:4-6`), so staged partials inside it would be backed up and skew the guard's
count, and a fresh restore refuses a library root that holds files (`doc.go:26-28`). Config load checks the nesting
lexically: `filepath.Clean` on both, then `filepath.Rel` with an `IsLocal` test in each direction. It never calls
`filepath.EvalSymlinks`, which fails on a path that does not exist, so a fresh box loads config and `doctor` can
report the absence. The symlink-resolved comparison runs in the startup check, after both paths are proven to exist.

### Layout

```
<staging>/tus/            tusd's filestore: <id> and <id>.info per upload; received files stay here
<staging>/submissions/    one directory per submission, named by its hex id; a ZIP's members extract to <id>/x/
```

A member's file name never reaches the filesystem. Each tus upload is named by its submission-file id, and each
extracted ZIP member by its own file id. Member paths are kept as bytes in the store (below), which settles
R6:562-563's non-UTF-8 question for staging.

Rung 7's own staging writes, renames, and removals go through an `os.Root` opened on the staging directory. `go doc
os.Root` reads "If any component of a file name passed to a method of Root references a location outside the root,
the method returns an error", and "symbolic links may not reference a location outside the root". tusd's filestore
writes `tus/<id>` and `tus/<id>.info` with `os.OpenFile` on joined paths (`pkg/filestore/filestore.go:110`,
`:323-324`). Those writes are safe because the wrapper admits only well-formed hex ids that the create callback
minted.

### The unit grant

The serve unit runs under `ProtectSystem=strict` with `ReadWritePaths=/var/lib/dubplate/auth
/var/lib/dubplate/store` (`systemd/dubplate-server-serve.service:43`). Rung 7 adds one entry: the music volume's
directory, `/srv/music` on this box, which holds both staging and the library root. `rename(2)` fails with `EXDEV`
across mount points, "even if the same filesystem is mounted on both" (`man 2 rename`), and two `ReadWritePaths=`
entries can become two mount points in the unit's namespace. One entry over the common parent keeps rung 9's
staging-to-library rename inside one mount. Rung 7 writes nothing in the library; it grants the parent now so the
startup check proves rung 9's rename from rung 7 onward.

The state-dir agreement leg (`scripts/check-state-dir.sh`) refuses every `ReadWritePaths=` entry that is not a leaf
directly under the state directory (`:56-61`), so it gains one second grant class. The serve unit's music-volume
entry is admitted only when it is the common parent of the template's `DUBPLATE_STAGING_DIR` and
`DUBPLATE_LIBRARY_ROOT`; any other non-leaf grant still fails. `deploy.sh` pre-creates staging and the library root
with the service user's ownership, each on its own checked line outside the leaves loop, so a stranger's fresh
install can start `serve`. The leg, `deploy.sh`, and `scripts/check.sh` are edited only after the gate scope pass
(`plans/2026-10-10-gate-scope-pass.md`) merges, since another lane is editing `check.sh` now. CLAUDE.md's Gate
paragraph and its state-directory fact line owe amendments ("Owed records").

### The startup check

Before it opens a listener, `serve` checks staging and the library root. The check is a pure function over two stat
results, tested table-driven over every row below, and a thin reader makes the syscalls. Mount ids come from
`statx(2)` with `STATX_MNT_ID`, "the mount ID of the mount containing the file" (`man 2 statx`, "since Linux 5.8"),
which `golang.org/x/sys/unix.Statx_t` carries as `Mnt_id`. `statx` "may fail to return fields that were requested",
so the reader tests `stx_mask` for `STATX_MNT_ID`. Equal `st_dev` values do not prove a rename works, because a bind
mount keeps its `st_dev`, so the mount-id clause strengthens conductor ruling 1. A check that cannot run refuses the
start (L4: never a pass, `research/2026-10-06-slice-1-melodee-intake.md:350-353`).

| State | Report |
|---|---|
| staging absent | start refused: "staging directory <path> does not exist"; the runbook names `deploy.sh` |
| staging not a directory, or not writable | start refused, naming the path and the failing operation |
| library root absent | start refused, naming the path |
| the two nest after `filepath.EvalSymlinks` | start refused, naming both resolved paths |
| `st_dev` differs | start refused, naming both paths and both devices |
| mount id differs | start refused, saying the two must sit in one mount |
| `statx` fails, or `stx_mask` lacks `STATX_MNT_ID` | start refused, naming the call and its error |
| all match | start continues |

`doctor` gains a row reporting the same facts for staging, the library root, and every collection root that exists
(CH:472-473). A collection root not yet created reports "not created", because rung 9 creates them. The message names
a mount, never `ReadWritePaths=`; the runbook line for the refusal names the unit grant.

### Platforms (hook 23)

Hook 23 asks that an OS-specific check sit "in its own build-constrained file with a counterpart for the other
platforms or a stated refusal there" (SS, hook 23). Rung 7's two OS-specific pieces, the device check and the
free-space read, sit in one leaf package with two files:

- `_linux.go`: `st_dev` and the `statx` mount id, and free space as `unix.Statfs`'s `Bavail * Frsize`, as
  `internal/engine/store/snapshot.go:198` reads it (`f_bavail` counts `f_frsize` units).
- `_other.go` with `//go:build !linux`: both functions refuse with "staging checks are built for Linux only", so
  `serve` refuses to start there.

No 0.8 deployment runs off Linux, and 0.9's home install is a Raspberry Pi. The store itself does not build for darwin
or windows (`syscall.Flock` at `internal/engine/store/lease.go:74`, `syscall.Statfs` at `fsprobe.go:43`), so a
non-Linux counterpart has no caller until 0.9 ports the store. The rest of rung 7's intake code is portable by
construction: `os.Root`, `archive/zip`, `crypto/*`, and `path/filepath` are stdlib, and tusd's filestore and
`memorylocker` use only `os`. A cross-compile check (acceptance criterion 34) builds for `darwin` and `windows` the
packages that import no `store`: `internal/reject`, the gate core, and the platform leaf. The plan keeps the gate core
free of `store` imports so that check means something.

## The submission model

**Hooks:** lands hook 1 (the audio key and file hash on submission files), hook 17 (the uploading member), hook 21
(the source answer), and hook 4 (the journey as an event table). Rung 9 copies each onto its album and track rows at
admission. Neutral on hooks 23 and 24.

### Submissions and verdict units

A submission is one album or one set of separate tracks, chosen by the member at upload (owner ruling 6). A member's
upload is several submissions: the page groups them, and the server judges each on its own. The page treats one ZIP,
one set of loose files picked together, and one dropped folder with its subfolders as one submission, and defaults
each to "album". The source question is asked once per upload, and each submission in the review list may override it
(PS:520-523).

- **Album.** One failing track rejects the submission, and no partial album is ever accepted (BS:812-814).
- **Separate tracks.** Each track has its own verdict. When at least one passes, the submission is accepted with its
  passing tracks, into the unlisted collection with no MusicBrainz match; when none passes, it is rejected. The
  precedent is beets' singleton import (`beet import -s`), whose tracks take its non-album path (the `singleton:`
  path, default `Non-Album/$artist/$title`). Rung 9 files them under its own non-album layout path.

### Migration `0006`

14a adds no migration (14A:429-434), and 14b follows rung 7, so rung 7 takes `0006`. It carries R-6d4's keep limit
and adds its tables to the older-store helper's map (`cmd/dubplate-server/olderstore_test.go:28-33`). The tables
follow the store invariants: STRICT, BLOB paths (invariant 1), `at` and `kind` NOT NULL on any table with a `kind`
column or an `_events` name (invariant 3, `internal/engine/store/invariant_test.go:329-346`), and a single BLOB `id`
primary key (invariant 6, `:397-416`). No new non-event column is named `kind`.

- **`submissions`**: `id`; `uploader` (BLOB, NOT NULL, references `users`); `collection_id` (BLOB, NOT NULL,
  references `collections`); `verdict_unit` (`album` or `track`); `container` (`files` or `zip`); `state` (written in
  the same transaction as the event that sets it); `source` (one of PS's seven values, NOT NULL); `source_note`;
  `store_url` (`https` only); `declared_bytes` and `reserved_bytes` (INTEGER, NOT NULL, at least 0); the match
  proposal (`mb_release_id`, the serialized candidate, and the evaluation date) or the entered album fields for the
  unlisted route (artist, album, year); `album_key` (BLOB, set for an album when the gate passes); `created_at` and
  `updated_at`. The column is `mb_release_id`, as `albums` names it, because invariant 4 refuses a bare `release_id`.
- **`submission_files`**: `id`; `submission_id`; `rel_path` (BLOB, under the relative-path CHECKs
  `0003_catalog.sql:15-22` uses, unique per submission); `size`; `role` (`audio`, `extra`, or `archive`, derived by
  the server from the extension); `received_at`; `audio_key` (BLOB of 16 bytes, NULL until decoded); `header_md5`
  (BLOB of 16 bytes, NULL when the header's value is all zero); `file_hash` (BLOB of 32 bytes); `sample_rate`,
  `bits_per_sample`, `channels`, `samples`; `disc` and `track` as read from the tags; `entered_artist` and
  `entered_title` (the unlisted route's form); `verdict` (a reject code, or NULL for a pass).
- **`submission_events`**: `id`, `submission_id`, `at`, `kind`, and optional `reason` and `detail`. `kind` is a
  journey state or one of the event-only kinds below. It carries no SQL CHECK, on `move_intents.kind`'s precedent
  (`0005_routed.sql`), so rung 9 adds its states without rebuilding the table. Go validates `kind` against the closed
  vocabulary at write.

The `collections` table gains `intake_route`, which is NULL, `default`, or `unlisted`, under a unique index, so at most
one collection holds each route. On the repo's SQLite (3.53.4), `ALTER TABLE ... ADD COLUMN` with a CHECK succeeds on
a STRICT table while `ADD COLUMN ... UNIQUE` is refused, so the separate index is the shape, and NULL rows coexist
under it (verified by the mechanics review). The plan quotes the SQLite pages before writing the migration.

Why a column and not a settings row: CH puts "routing rules: which collection an upload lands in" in the settings
table (CH:289). A settings value holding a collection id has no referential integrity and could be seeded from the
environment. CH owes an erratum that narrows its routing row to rules beyond these two fixed routes.

### Journey states

Rung 7 defines only the states it emits (B D7, the 2026-09-12 refinement that "nothing is exported ahead of a
caller"):

| State | Meaning | Terminal |
|---|---|---|
| `uploading` | created; bytes are moving | no |
| `checking` | every file received; the gate is running or waiting for a worker | no |
| `check-unfinished` | the gate could not run; the bytes stay staged (conductor ruling 5) | no |
| `matching` | an album passed the gate; the probe is running | no |
| `match-unavailable` | MusicBrainz did not answer; the bytes stay staged | no |
| `pushed-back` | no confident match | no |
| `accepted` | waiting for admission, with a stored proposal or entered fields | no (rung 9 moves it) |
| `rejected` | the gate failed, or a duplicate was refused; per-file reasons recorded | yes |
| `cancelled` | the member cancelled | yes |
| `expired` | idle or held past its window, or its staged bytes were found missing | yes |

One event kind is not a state: `stalled`, recorded when a submission has sat in `checking` or `matching` past
`DUBPLATE_INTAKE_STALL_AFTER`. The submission keeps its state, so recovery and the worker need no extra transition,
and "mail the admins once" reads as "no earlier `stalled` event", which survives a restart.

Transitions, each a compare-and-set on the current state inside one `store.Handle.Update`; a writer whose
compare-and-set fails discards its result:

| From | To | By |
|---|---|---|
| `uploading` | `checking` | receipt of the last file |
| `checking` | `matching`, `accepted`, `rejected`, `check-unfinished` | the worker's verdict |
| `matching` | `accepted`, `pushed-back`, `match-unavailable`, `rejected` (a duplicate) | the probe |
| `check-unfinished` | `checking` | `retry_submission` |
| `match-unavailable`, `pushed-back` | `matching` | `retry_submission` |
| `pushed-back` | `accepted` in the unlisted collection, or `rejected` (a duplicate) | `route_unlisted` |
| `uploading`, `check-unfinished`, `match-unavailable`, `pushed-back` | `cancelled` | `cancel_submission` |
| `uploading` | `expired` | the sweep, idle past `DUBPLATE_INTAKE_RESUME_WINDOW` |
| `check-unfinished`, `match-unavailable`, `pushed-back` | `expired` | the sweep, past `DUBPLATE_INTAKE_HOLD` |
| any non-terminal | `expired` | the startup pass, when its staged bytes are missing |

Rung 9 adds its own states; 14A's version policy makes a new value additive (14A:81-88). No state carries a queue
position, which keeps ruling 10's "`/api/v1` publishes no queue position"
(`research/2026-09-02-greenfield-server-rulings.md:63-65`). PCF section 7's "queued" belongs to rung 9's import.

## The transfer

**Hooks:** neutral, including on hook 23 (tusd, its filestore, and `memorylocker` are portable Go). Neutral on hook
24: the base path reads `DUBPLATE_PUBLIC_URL`, which a LAN install sets to its LAN origin.

### tusd, configured closed

Rung 7 adopts `github.com/tus/tusd/v2` v2.10.1 ("Prior art"). The handler's settings, quoted from
`pkg/handler/config.go` and verified against a probe handler by the mechanics review:

- **CORS off.** `DefaultCorsConfig` sets `AllowOrigin: regexp.MustCompile(".*")` (`:143-145`). The mount sets
  `Cors: &CorsConfig{Disable: true}`, which ignores and never sets a CORS header (`:114-116`). The shell's origin
  check stays the one CSRF defense (`internal/serve/origin.go:107-120`).
- **Downloads and termination off.** `DisableDownload` and `DisableTermination` unmount GET and DELETE (`:30-35`), so
  both answer 405 (`handler.go:59-66`). A member cancels through `cancel_submission`, never a tus DELETE, so no tus
  request can remove a received file.
- **No concatenation.** The composer is built by hand with `UseCore` and `UseLocker`, never `FileStore.UseIn`, which
  also composes the concater, the length deferrer, and the content server. With no concater, tusd ignores
  `Upload-Concat` (`unrouted_handler.go:298-306`), so the wrapper refuses that header with 400. With no length
  deferrer, `Upload-Defer-Length` answers 501 (`:1454-1455`), so every upload declares its length first.
- **The experimental IETF protocol off**, and **forwarded headers ignored** (`:26-29`, `:61-64`).
- **An absolute base path.** With a scheme in `BasePath`, `absFileURL` returns `basePath + id` without reading `Host`
  (`unrouted_handler.go:1312-1315`). The base path is `DUBPLATE_PUBLIC_URL`, the shell's prefix, and `/tus/`.
- **`MaxSize`** is the album size cap: a larger `Upload-Length` answers 413 (`unrouted_handler.go:1440-1444`).
- **Permissions.** The mount sets `DirModePerm` 0750 and `FileModePerm` 0640 on the filestore (defaults 0775 and
  0664, `filestore.go:60-69`), matching the unit's `UMask=0027`.

**The logger.** `Config.Logger` is a `*golang.org/x/exp/slog.Logger` (`config.go:59-60`), a distinct type from
`log/slog`'s, and a nil logger becomes the `x/exp` package's default, outside the shell's sink (`:154-156`). Issue
#1159, "slog: replace golang.org/x/exp/slog with log/slog", is open (read 2026-10-10). The mount passes an `x/exp/slog`
JSON handler at warn level writing to the stream the shell's logger writes, so tusd's `BodyReadError`,
`NetworkControlError` (the only trace of a lost `Unwrap`), and internal errors stay visible. That is a handler of the
dependency's own package, not an adapter; when #1159 lands, the bump passes the shell's logger.

**One process owns intake.** `store.OpenServe` holds the store lock shared (`internal/engine/store/open.go:201-203`),
so it does not make `serve` the only process on staging. `serve` takes `Handle.TakeKindLease("intake")`, "an exclusive
flock(2) ... which admits one job of a kind at a time across every process" (`lease.go:166-200`), before the startup
pass and before mounting `tus`, and holds it for its life. `ErrLeaseBusy` refuses the start, naming the holder.
`internal/ops/local` takes its job leases the same way (`local.go:227`). Under that lease, `memorylocker`, whose locks
"will be erased if the program exits" (`pkg/memorylocker/memorylocker.go`), is enough, and `filelocker` (with
`github.com/tus/lockfile`) is declined. The lease also tells rung 9 that `serve` owns staging.

### The wrapper

tusd offers no hook on PATCH or HEAD (`handler.go:40-80`), so a rung 7 wrapper authorizes every request before tusd
sees it:

1. It refuses `X-HTTP-Method-Override` with 400. tusd lets a POST carrying it become any method
   (`unrouted_handler.go:187-192`); the mechanics probe terminated an upload this way.
2. It refuses `Upload-Concat` with 400.
3. It requires a verified member session (the member gate, segment C). No session answers 401.
4. For a path naming an upload, it parses the id as the hex form of a `store.NewID` id and answers 404 to any other
   shape before tusd joins it onto a path. It answers 404 unless the session's `UserID` is the submission's
   `uploader`. HEAD is admitted for a file of any non-terminal submission, PATCH only while the submission is
   `uploading`. An unknown id, another member's id, and a disallowed state give the same 404. A creation names no id
   in its path, so the create callback checks its file.
5. For POST and PATCH, it checks the free-space floor before tusd reads the body.
6. It cuts the shell prefix and `/tus` and hands tusd the rest, as 14a's mount does for `api` (14A:117-119).

tusd extends its read deadline after every successful body read (`unrouted_handler.go:922-929`) through
`ResponseController.SetReadDeadline`, which keeps a long PATCH alive past the server's 30-second `ReadTimeout`
(`internal/serve/server.go:36`). That needs the writer to "be the original value passed to the Handler.ServeHTTP
method, or have an Unwrap method" (`go doc net/http.NewResponseController`). The shell's `statusCapture` has `Unwrap`
(`internal/serve/serve.go:283`), and any wrapper rung 7 adds has it too.

### The callbacks and receipt

Received bytes stay where tusd wrote them, in `tus/<id>`, until the submission ends. tusd derives an upload's offset
from the data file (`info.Offset = stat.Size()`, `filestore.go:166`), so a HEAD on a received file answers its full
length, and tus-js-client then reports success (`lib/upload.js:724-727`). Nothing moves a received file at receipt,
so a lost response or a crash cannot strand one.

- **`PreUploadCreateCallback`.** It reads the tus metadata keys `submission` and `file`, and checks that the file row
  belongs to the session's `uploading` submission, is unreceived, has no tus upload yet (`FileStore.GetUpload` fails
  with not found, `filestore.go:130`), and declares a size equal to `Upload-Length`. It checks the floor. It returns
  `FileInfoChanges{ID: <file id hex>, MetaData: MetaData{}}`: `ID` is respected by "all data stores in the
  github.com/tus/tusd package" and an empty map stores no client metadata (`datastore.go:61-75`). The member reaches
  the callback through `HookEvent.Context`, which "retains the same values as the request context" (`hooks.go:9-22`).
- **`PreFinishResponseCallback`** runs receipt for its file.
- **Receipt** is one idempotent function. It syncs `tus/<id>` and the `tus/` directory, because the filestore writes
  with no `Sync` (`filestore.go:220-228`, `:311-313`) and the store's commit is durable on another device; the
  precedent is `internal/atomicfile` (`atomicfile.go:29`, `:87`). Then one `Update` marks the file received with a
  compare-and-set on `uploading`, decides inside the same transaction whether every file is received, and if so moves
  the submission to `checking`. When the compare-and-set fails because the submission has closed, receipt changes
  nothing and the callback answers 403; cleanup removes the bytes. The startup pass and every sweep also run receipt
  for each unreceived file of an `uploading` submission whose tus upload is complete (offset equal to size).

Two creates racing for one file id (two tabs on one submission) can let the second truncate the first's file, because
the filestore creates with `O_TRUNC`. The gate's decode refuses the damaged file as `corrupt`, so the state is named
and needs no lock.

### Refusals on the tus path

Callback refusals are `handler.Error` values built with `handler.NewError`; a plain Go error reaches the client as 500
with its text (`sendError`, `unrouted_handler.go:1288-1291`). tus-js-client's default `onShouldRetry` retries 409,
423, and any status outside 4xx (`lib/upload.js:1076-1081`), so a definitive refusal uses 400 or 403, and only a
transient one is retried.

| Refusal | Status | Code in the body |
|---|---|---|
| no session | 401 | `unauthenticated` |
| malformed, unknown, another member's, or disallowed id | 404 | one body for all |
| method override, `Upload-Concat` | 400 | `invalid_request` |
| create: size differs from the declared size, or the file already has an upload | 400 | `invalid_request` |
| create: over `MaxSize` | 413 | tusd's |
| create or PATCH: the floor | 403 | `no-space` |
| create or PATCH: `statfs` fails | 503 | `could-not-check` (retried, then shown) |
| receipt into a closed submission | 403 | `conflict` |
| receipt: a store error | 503 | `could-not-check` (retried; the retry's HEAD answers the full offset, and the sweep finishes receipt) |

### The client

The page pins `tus-js-client` 4.3.1 exactly (CLAUDE.md's 4.3.x rule). Each option is quoted from `docs/api.md` at
`v4.3.1`, and the defaults were read from `lib/upload.js:11-47`:

- `endpoint` is the tus base path, and `uploadUrl` is the base path plus the file id that `create_submission`
  returned. The client then HEADs first and creates only after a 404 (`lib/upload.js:432-437`), which covers a lost
  201, a second device, and a reloaded page with no stored state. Without `endpoint`, a 404 ends the upload with "new
  upload cannot be created without an endpoint" (`lib/upload.js:685-693`).
- `storeFingerprintForResuming` is false: the server's file id is the resume key.
- `chunkSize` keeps its default, `Infinity`; the docs say "**Do not set this value**, unless you are being forced to".
- `parallelUploads` stays 1; a higher value uses the concatenation extension, which the server does not compose.
- `retryDelays` keeps its default, `[0, 1000, 3000, 5000]`.

**Resume after a reload.** "Your uploads" offers "Resume" on an `uploading` submission. The page asks the member to
pick the files again, binds each to `get_submission`'s file list by relative path and size, and uploads with that
file's `uploadUrl`, so no second reservation or session slot is taken. A picked file that matches no row is named,
with a cancel action offered.

The salvage driver's chunk downshift on 502, 504, and 524 (`_salvage/web/src/lib/tus.ts:12-25`) is declined: 524 is
Cloudflare's, the tunnel is deleted, and with no chunk size there is nothing to downshift.

## Quota and the free-space floor

**Hooks:** consumes hook 15's quota counters and override rows. Lands hook 23 for the free-space read. Neutral on hook
24.

### What a reservation is

A reservation is an open submission's `reserved_bytes`. A member's use is their `member_quota.charged_bytes` plus the
`reserved_bytes` of every non-terminal submission they uploaded. A reservation is taken in the `Update` that creates
the submission, from the manifest's declared total, so "two concurrent uploads cannot each pass the check and together
overshoot" (CH:555-557). The property is structural: the limit functions take the `*store.Tx` that `Update` hands its
callback, which "runs fn inside one IMMEDIATE transaction on the write pool" (`internal/engine/store/write.go:44-47`),
so no check can run outside the write transaction. Rung 7 never writes `charged_bytes`; rung 9 moves a submission's
bytes into it and drops the reservation in the transaction that admits it.

**Who is charged.** CH:548-551 charges a member for "everything they own". At 0.8 every upload lands in an admin-owned
collection, so a cap counted by ownership would never bind. Rung 7 charges the uploading member for what they upload,
so an admin's per-member cap means what it says. CH owes an erratum. CH's per-collection level (CH:548-549) is neutral
at rung 7: `collections` has no quota column, and 14b is its first setter.

**A ZIP's extracted total.** A ZIP's reservation starts as its archive size. Before extraction, one `Update` replaces
it with the sum of the members' `UncompressedSize64`, re-running limits 3 to 5 against that total with other
submissions' unwritten bytes counted. A failing limit rejects the submission with `over-quota` or `no-space`. The
true-up after the verdict only lowers a reservation: to the kept files' total, or for separate tracks to the passing
tracks'.

**Unwritten bytes.** An `uploading` submission's unwritten bytes are its reservation minus its received files' sizes;
a `checking` ZIP's are its extracted total until the verdict; every other state's are zero. This counts partly written
files in full, which errs toward refusing.

### The limits checked at creation

The create call's checks run in this order. Malformed input answers `invalid_request` and ends the check. Then every
per-file manifest reason and every submission-level manifest reason is collected, and if any is found the call
answers them all. Only a clean manifest reaches the limits, and the first failing limit ends the check:

1. `may_upload` on the member's override row is 0: `uploads-off`. NULL means allowed.
2. The member already has `upload_concurrency_cap` submissions in `uploading` (override, else
   `DUBPLATE_INTAKE_SESSIONS_PER_MEMBER`): `too-many-sessions`. The page holds a batch's extra creates until a slot
   frees, so a member never sees this for their own upload.
3. The member's override `quota_bytes` is set and their use plus the declared total exceeds it: `over-quota`.
4. `DUBPLATE_INTAKE_GROUP_QUOTA_BYTES` is set (default unset) and `group_quota` plus open reservations plus the
   declared total exceeds it: `over-quota`.
5. The floor: `no-space`.

### The floor

The floor is a registered setting, `DUBPLATE_INTAKE_FREE_SPACE_FLOOR`, a byte count, read against free space on
staging ("Platforms"). The check runs before every write: at creation, requiring free space minus every open
submission's unwritten bytes minus the new declared total to stay at or above the floor; at each tus creation and
before each PATCH, requiring free space at or above the floor; and before a ZIP's extraction, as above.

CH:554 gives the floor's reason as the store's write-ahead log. On this box the store sits on another device, so the
floor on the music volume protects the library and rung 9's admission: a tag write copies a whole file through
`internal/atomicfile` before its rename. The working default is 5 GiB, the salvage floor
(`_salvage/internal/config/config.go:176-181`), which holds one maximum-size album's tag-write copies; the plan records
its basis (BS:323-331). A `statfs` that fails refuses the write with `could-not-check`, never a pass (L4).

## The gate

**Hooks:** lands hook 1. Neutral on hook 23: the gate is pure Go over `os.Root`, `archive/zip`, `mewkiz/flac`, and
`tagio`. Neutral on the rest.

The gate is a pure function from a staged submission to a verdict, holding no per-run state, so concurrent verdicts
over distinct submissions equal their sequential verdicts (Melodee L5). It lives in a new `internal/intake` package
that imports no `store`. It runs in a checking worker in `serve`, with concurrency bounded by
`DUBPLATE_INTAKE_CHECK_WORKERS` (default 1, so decode cannot starve request handling; the plan measures decode
throughput on the box's CPU class). It takes a context and checks it between frames, so shutdown frees the worker. The
worker takes no store lease beyond intake's, which leaves ruling 10's per-album lease to rung 9.

### At creation (the manifest)

The create call carries the manifest: each file's relative path and size, the container kind, and the verdict unit.

- Every path passes the Go path validator that mirrors the schema's CHECKs (R6:567): relative, no empty, `.`, or `..`
  component, no NUL, no backslash, and equal to its own `path.Clean` form. `filepath.IsLocal` alone admits `a/./b`,
  `a//b`, and a trailing slash, which the `rel_path` CHECK refuses. A failing path: `unsafe-path`. A duplicate path:
  `invalid_request`.
- Path depth at most `DUBPLATE_INTAKE_MAX_PATH_DEPTH`: `path-too-deep`. File count at most
  `DUBPLATE_INTAKE_MAX_FILES`: `too-many-files`. Declared total at most `DUBPLATE_INTAKE_MAX_ALBUM_BYTES`:
  `too-large`.
- A file with an audio extension outside the allowlist: `bad-extension`, parameterized by the extension. A file with
  no audio extension is an extra: booklets, scans, and cue sheets "ride along" (PCF section 1), and their bytes count
  toward the cap. A separate-tracks submission keeps no extras. The page leaves them out of its manifest, so a
  separate-tracks manifest naming an extra answers `invalid_request` with detail `extra`, a client fault.
- At least one audio file, or for a ZIP exactly one file with a `.zip` extension: `no-audio` otherwise.
- The source answer present and one of the seven values: `missing-source` (PS A9).

### After the last byte (the files)

**ZIP.** The gate first removes `submissions/<id>/x/`, so a re-run after a crash starts clean. Before `zip.NewReader`
materializes the central directory, the gate counts its headers in constant memory and refuses the `MAX_FILES + 1`th
with `too-many-files`. The mechanics probe measured 77 MiB of heap in `NewReader` for a 35 MB ZIP of 400,000 empty
entries. The end record's entry count bounds nothing: `Reader.init` reads headers "until we encounter a bad one" and
compares the count only modulo 65536 (`if uint16(len(r.File)) != uint16(end.directoryRecords)`,
`$GOROOT/src/archive/zip/reader.go:139-161`), and the verification probe read 196,609 entries from a ZIP whose zip64
end record claimed 1. The count therefore uses `Reader.init`'s own stop rule. It seeks to the directory offset as
`readDirectoryEnd` computes it, base offset included (`reader.go:544-610`; unexported, so the plan ports that offset
arithmetic under its BSD-3-Clause license). It reads each 46-byte header, skips the name, extra, and comment lengths
the header carries, and stops at the first record whose signature is not `0x02014b50` or that ends early, the
`ErrFormat` and `io.ErrUnexpectedEOF` that break `Reader.init`'s loop (`readDirectoryHeader`, `reader.go:356-380`;
`directoryHeaderLen` 46, `struct.go:44`). The walk then reads the central directory without extracting, adopting
`_salvage/internal/intake/intake.go`'s shape. `zip.NewReader` refuses a non-local name only under `GODEBUG`
`zipinsecurepath=0` (`go doc archive/zip.NewReader`), so the walk runs the path validator on every name itself. A
symlink member (`FileHeader.Mode()` has `fs.ModeSymlink`), a name the validator refuses, a directory entry with
content, and two members with one name all answer `unsafe-path`. The walk sums every member's `UncompressedSize64` and
refuses a sum over the cap before writing a byte (`too-large`); the reader stops at the declared size ("if r.nread >
r.f.UncompressedSize64 { return 0, ErrFormat }", `$GOROOT/src/archive/zip/reader.go:301-303`), so that sum bounds
decompression. Then comes the reservation `Update` under "Quota". A ZIP that does not open, a member read that errors,
and a CRC mismatch answer `unpack-failed`.

Members extract into `submissions/<id>/x/` through the staging root. They are scratch until the verdict: nothing
reads one before it commits, a re-run removes the directory first, and `rejected` and `check-unfinished` remove it,
so extraction writes directly rather than through `internal/atomicfile`. Member rows are written only in the verdict
`Update`. A member name is kept as bytes in `rel_path`. A name that is not valid UTF-8, or whose `FileHeader.NonUTF8`
is set, is accepted with no decoding guessed (`go doc archive/zip.FileHeader`); display replaces invalid bytes with
U+FFFD.

**Staged bytes, per state.** Reconcile and cleanup test a submission against this set. For loose files, the staged
bytes are its received files in `tus/`. For a ZIP, they are the archive in `tus/` through `uploading`, `checking`, and
`check-unfinished`, so a retry re-extracts from it. From the verdict that leaves `checking` for `matching` or
`accepted`, they are the kept members in `x/`, and only after that verdict commits is the archive removed. For an
accepted separate-tracks submission, they are its passing files.

**Per audio file**, in this order. The first failing check is the file's verdict:

1. **Container and format.** A file whose first three bytes are `ID3`, the ID3v2 identifier, gets `id3-tags` (owner
   ruling 8). mewkiz/flac tests the same three bytes before it skips the prefix (`id3Signature`, `flac.go:136`,
   `:163-173`), as mutagen 1.48.1 does (`header_data.startswith(b"ID3")`, `mutagen/id3/_file.py:377`). The reference
   decoder accepts such a file with a warning ("ID3v2 tag found. This is non-standard and strongly discouraged",
   `flac -t`, run by the verification probe), and so does mewkiz, so only this check catches it; `tagio.Sniff` would
   misreport it as the wrong format. Otherwise `tagio.Sniff` names FLAC, and the extension's format is in the
   effective allowlist: the collection's `admission_format_allowlist` intersected with the
   `DUBPLATE_INTAKE_FORMAT_ALLOWLIST` bound (CH:280-282, "the bounds an admin sets that no collection may exceed"),
   within `tagio.SupportedFormats`; a format outside `tagio`'s set is a startup refusal (BS:808-811). The default is
   `flac`. A mismatch: `bad-extension`. `tagio.ErrMalformedContainer` maps to `corrupt` and
   `tagio.ErrUnsupportedContainer` to `bad-extension`.
2. **Metadata size.** No Vorbis comment block over `DUBPLATE_INTAKE_MAX_TAG_BYTES`, no picture block over
   `DUBPLATE_INTAKE_MAX_PICTURE_BYTES`, and no chain totalling more than four times the picture cap:
   `tags-too-large` and `picture-too-large`. `tagio.Parse` allocates each block's payload as it walks the chain
   (`internal/engine/tagio/container.go:199-202`), so a size check made on its result runs after the memory is spent.
   The bound goes into `tagio`'s own walk, the one place that reads block headers, and is checked from each 4-byte
   header before the allocation.
3. **Stream format.** From STREAMINFO: a sample rate below `DUBPLATE_INTAKE_MIN_SAMPLE_RATE` (default 44100) or bits
   per sample below `DUBPLATE_INTAKE_MIN_BITS_PER_SAMPLE` (default 16): `below-cd-quality`, the salvage constants
   (`_salvage/internal/serve/upload.go:48-51`) and PCF section 1's "CD quality or better". Any depth other than 16, 20,
   or 24: `unsupported-flac`. mewkiz/flac's frame parser refuses code `111` ("reserved sample size bit pattern",
   `frame/frame.go:358-361`) and reads code `000` as "get from StreamInfo", which fails on an 18-bit stream with
   "non-zero padding"; `Frame.Hash` handles 1 to 24 bits only (`:178-179`). The three admitted depths are the ones
   mewkiz decodes by frame code and that pass the CD floor.
4. **Decode, the audio key, and the file hash.** One pass over the file: a `crypto/sha256` hash over every byte, and a
   `mewkiz/flac` decode of every frame. `flac.New` "reads and parses the FLAC signature and the StreamInfo metadata
   block, but skips all other metadata blocks" (`flac.go:67-72`). `ParseNext` "parses the entire next frame including
   audio samples" and "returns io.EOF to signal a graceful end of FLAC stream" (`flac.go:308-310`). Frame CRC-8 and
   CRC-16 mismatches are errors (`frame/frame.go:317`, `:142`). Any decode error, or bytes hashed unequal to the file's
   size: `corrupt`. Zero samples: `empty-track`.

   A decode error whose file ends in a 128-byte record beginning `TAG`, the ID3v1 identifier, gets `id3-tags` instead
   of `corrupt`. mutagen 1.48.1 finds ID3v1 the same way: it seeks 128 bytes from the end (`fileobj.seek(-128 -
   extra_read, 2)`) and looks for `b"TAG"` (`mutagen/id3/_id3v1.py`, `find_id3v1`). Such a file fails both decoders:
   mewkiz with "frame.Frame.parseHeader: invalid sync-code", and `flac -t` with `LOST_SYNC` (the verification probe).
   The record is tested only after a decode error, so a valid FLAC whose last bytes happen to spell `TAG` still passes.

   The audio key is `Frame.Hash` over every decoded frame into one `crypto/md5`; `Hash` "adds the decoded audio
   samples of the frame to a running MD5 hash" for use with `StreamInfo.MD5sum` (`frame/frame.go:148-152`). When the
   header's `MD5sum` is nonzero and differs from the audio key, or STREAMINFO's nonzero `NSamples` differs from the
   decoded count ("A 0 value implies unknown", `meta/streaminfo.go:34-37`): `source-damaged`. The file arrived intact
   but was written wrong at its source, so re-sending it cannot help; `corrupt`'s adopted copy ("Uploading it again
   usually fixes this") stays true for transit damage alone. An all-zero header value "is absent, never a value"
   (CH:424-425) and is stored as NULL.
5. **Tags.** `tagio.Read` gives `disc` and `track` (Melodee L3: "dubplate's gate keys on (disc, track)"). A present
   but non-numeric value reads as 0, because `readInt` projects through `safeCastInt`, which returns 0 when there is no
   numeric prefix (`internal/engine/tagio/read.go:296-303`, `encode.go:380-390`). For an album, a track number absent
   or at or below 0 (a vinyl rip's `A1` reads as 0): `missing-track-number`. A disc absent or at or below 0 reads as
   disc 1. For separate tracks, tags only prefill the form.

**Per album**, after every file passes:

6. **Positions.** Two files at one (disc, track): `duplicate-track-position`. Tracks must run 1 to n on each disc and
   discs 1 to m; any break, a count that differs from a `TRACKTOTAL` or `DISCTOTAL` the tags carry, or tracks that
   disagree on a total: `incomplete-album`, because "no partial album is ever admitted". A two-disc album whose
   numbers restart on disc 2 passes. No check rewrites a tag to pass (L3).
7. **The album key.** SHA-256 over the track count and the audio keys in (disc, track) order, CH:433-435's "ordered
   set of a release's track audio keys with its track count". The duplicate refusal ("Duplicates") reads it.

**Completeness checks that warn and never block** (Melodee L2 and L9), for albums: no picture block and no cover image
among the extras; the album total from tags differs from the count across discs. Each is shown with its own copy.

**Separate tracks.** Checks 1 to 4 run per file; 5 to 7 and the warnings do not. A file's failing check is its own
verdict. With at least one passing file, the submission moves to `accepted` in the unlisted collection; with none, to
`rejected`. No probe runs. A ZIP sent as separate tracks keeps only its audio members: each extra member gets no row,
is removed with `x/`, and is listed in a non-blocking warning that separate tracks keep audio files only.

**A check that cannot run** (L4): a read error on a staged file, a `tagio` error other than its two sentinels, or a
`statfs` error at extraction moves the submission to `check-unfinished`, keeping its staged bytes (a ZIP's archive)
and its reservation (conductor ruling 5). A store error during the verdict cannot commit any state; the worker logs
it, and the submission stays `checking` for the startup pass or the stall event.

On `rejected`, cleanup removes the submission's staged files; its rows and events stay.

### Duplicates

The authoritative refusal runs inside the `Update` that moves a submission to `accepted`, on every path into it: the
probe's `ActionApply`, an unlisted-route gate pass, separate tracks, and `route_unlisted`. It reads the target
collection, so CH:651's visibility rule holds, and duplicates across collections stay allowed (CH:567-568).

- An album whose `album_key` equals an `accepted` submission's in the target collection moves to `rejected` with
  `already-in-library`. Of two identical albums racing through `matching`, the second to commit is refused.
- A separate track whose `audio_key` equals that of a passing file of an `accepted` submission in the target
  collection gets the verdict `already-in-library`; the others go on. When every passing track is refused this way,
  the submission ends `rejected`.

No early check runs at gate pass, so a member is never refused for a twin that may still be cancelled. Rung 9 extends
the refusal to admitted tracks once `tracks` carries the audio key, and adds its admission states to the set it reads.

### Defaults

Each cap's default is researched in the plan, which records its basis where the key is defined (BS:323-331):

| Setting | Proposal | Basis |
|---|---|---|
| `MAX_ALBUM_BYTES` | 4 GiB | a 79.8-minute 24-bit/192 kHz stereo album is about 5.5 GB raw and about 3.3 GB at a typical 0.6 FLAC ratio; the salvage 2 GiB refuses it. The plan measures the ratio on corpus files. |
| `MAX_FILES` | 500 | a multi-disc box set plus scans; the plan checks the largest track counts MusicBrainz lists |
| `MAX_PATH_DEPTH` | 8 | album, disc, and scan subfolders with margin |
| `MAX_TAG_BYTES`, `MAX_PICTURE_BYTES` | 1 MiB and 8 MiB | below FLAC's 16 MiB block maximum, so the picture cap bounds something; the plan reads `tagio`'s limits |
| `SESSIONS_PER_MEMBER` | 3 | the salvage value (`config.go:176-181`) |
| `FREE_SPACE_FLOOR` | 5 GiB | above |
| `FORMAT_ALLOWLIST` | `flac` | BS:802-811, within the tag I/O bound |
| `RESUME_WINDOW` | 7 days | a weekend away from a slow residential upload, plus margin |
| `HOLD` | 7 days | long enough to add a release to MusicBrainz and come back |

## The default collection and the unlisted route

**Hooks:** consumes hooks 8 and 15 (the `writer`, `owner`, and `visibility` columns); neutral on hook 22, because "a
rung that creates collections before slice 3 is neutral" (SS:1033-1040). Neutral on hooks 23 and 24.

**The default collection.** No collection row exists (R6:185-186). The `Update` that creates the first main-route
submission creates the default collection, named "Library", when none holds `intake_route = 'default'`: `writer`
`server`, `visibility` `group`, `owner` the earliest-created live admin, and `layout_template` empty (the template
machinery is rung 9's, B D9). Its `root_path` is allocated from its id (CH:322-323), and the no-nesting check runs
against every existing root before the row lands (CH:323-325; R6:181). The path validator runs on every path rung 7
writes to the store: `root_path` and every `rel_path`. No live admin at that moment refuses the submission with
`internal` and a log line; it cannot happen after bootstrap.

**The unlisted route.** UD's escape for music MusicBrainz cannot place (UD:251-258) survives as "Unlisted music"
(owner ruling 1): a second server-written, admin-owned, group-visible collection with `intake_route = 'unlisted'`,
created the same way the first time a submission takes the route. Taking it changes the contract, and the page says so
before upload starts (UD:252-253): no MusicBrainz matching; for an album, a short form (artist, album, year) and track
titles prefilled from file names; for separate tracks, an artist and title per track prefilled from tags, then file
names; all editable. A submission reaches it three ways:

- at selection, by choosing it for an album, which shows the form before upload; a passing gate then moves the
  submission straight to `accepted`;
- as separate tracks, which always take this route;
- from `pushed-back`, by `route_unlisted` with the form; the bytes are not sent again.

The gate's checks are the same on every route. Rung 9 admits an `accepted` submission without a proposal by writing
the entered fields. Rung 13's appeal action (OD:326-336) is a later way out of `pushed-back`.

**The collection picker.** It lists only collections a member may pick as a main-route target. The unlisted
collection is reached through the route choice alone and never appears in the picker, so W7.19's picker, which hides
"when one collection is configured", stays hidden at 0.8. W8.2 lets the route's copy name the collection, since a
deployment with the unlisted collection has more than one. The member's default-collection preference and the
per-upload override wait for 14b, the first verb that can make a second pickable collection (B D9).

CH:560-565's re-validation at write ("re-validates at write") is built by 14b (conductor ruling 9). At rung 7 no
channel can retire or narrow a collection, so the accept transaction reads the target only for the duplicate refusal.

## Matching: the user-present probe

**Hooks:** neutral, hooks 23 and 24 included.

When an album passes the gate on the main route, the submission moves to `matching` and the probe runs `tag.Album`
(`internal/engine/tag/tag.go:66-67`) over its tracks, with `tagio`'s tag set projected onto `matcher.ItemFields` and the
length from STREAMINFO. The probe is the first production caller of `musicbrainz.ClassUserPresent`, "a person is
waiting on this request" (`internal/engine/musicbrainz/limiter.go:16-21`), and the first production
`musicbrainz.Client`.
`serve` builds one `Limiter`. A `Budget` spends "at most limit requests" over its life (`limiter.go:122-128`), and a
`Client` holds one (`client.go:91-95`), so each probe run takes a fresh `ClassUserPresent` budget and a fresh `Client`.
The per-run limit is `DUBPLATE_INTAKE_MATCH_BUDGET`, derived by the plan from `tag.Album`'s request pattern. The
User-Agent's contact is the one 14A:384-385 names. The `TagSet` to `ItemFields` projection is new production code; it
reads `tagio`'s field names, which follow Picard's tag-mapping tables as rung 8 adopted them
(`specs/2026-09-07-rung-8-tag-io.md:1157`), and the plan quotes each field it projects.

Strict custody is rung 7's only policy (OD:311-313). A proposal is confident when `Proposal.Action` is
`matcher.ActionApply`, the decision "beets takes unprompted" (`go doc ./internal/engine/matcher Action`):

| Probe outcome | State | Stored |
|---|---|---|
| `ActionApply` | `accepted`, unless the duplicate refusal rejects it | `mb_release_id`, the candidate, the evaluation date |
| any other action, or `tag.ErrNoCandidates` | `pushed-back` | the top candidate's distance for the page's message |
| an error matching `musicbrainz.ErrUnavailable` | `match-unavailable` | nothing |
| any other error | `match-unavailable`, logged at error | nothing |

The stored proposal lets rung 9 admit without matching again (B D6). `evalDate` is the probe's wall-clock date,
because `tag.Album` takes it "as a parameter and never a clock read".

From `pushed-back` the member may fix names (the page cancels the submission and returns to selection with the source
answer kept, since the staged bytes carry the old tags), try again (`retry_submission`, for after a release is added to
MusicBrainz), take the unlisted route, or cancel. From `match-unavailable` the member may try again or cancel. The
sweep expires both past the hold.

If rung 9's import runs in a process other than `serve`, its limiter will not share the probe's wire budget
(`limiter.go:93-102`); that is rung 9's to settle.

## The reject set

**Hooks:** neutral, hooks 23 and 24 included.

`internal/reject` holds one enumerable set (B D11). Each reason carries a stable wire code, a short chip label, the
friendly sentence, the next step, whether it blocks, and, for a parameterized reason, the parameter and a canonical
wording with the parameter filled generically, which the project site's reject-page generator reads
(`specs/2026-09-25-project-site-design.md:301-306`).

Nine reasons are adopted verbatim from PCF section 4, the single authority (BS:508-511): `bad-extension`, `symlink`,
`corrupt`, `below-cd-quality`, `too-large`, `unpack-failed`, `unsafe-path`, `no-space`, `too-many-sessions`. A test
holds their strings as literals transcribed from PCF, so the contract cannot drift silently. `symlink` stays for rung
9's timer path; rung 7 files a ZIP symlink member under `unsafe-path`, as the salvage did. Two adopted strings name
FLAC, which W8.2's leak list names; conductor ruling 6 adds the format-name exception for reject reasons.
`bad-extension` renders PCF's strings byte for byte under the FLAC-only default. Its parameterized canonical wording,
which names the configured set and suggests a lossless conversion to FLAC where one exists (`ROADMAP.md:155-159`), is
new copy.

New reasons, drafted to PCF's rules (no em dashes, no "This [thing] couldn't be [verbed]" shape, no engineering state
names, W8.2's leak list) and gated by one `prose-voice-reviewer` read: `missing-source`, `uploads-off`, `over-quota`,
`too-many-files`, `path-too-deep`, `no-audio`, `tags-too-large`, `picture-too-large`, `unsupported-flac`,
`source-damaged`, `empty-track`, `missing-track-number`, `duplicate-track-position`, `incomplete-album`,
`already-in-library`, `could-not-check`, and `id3-tags`. The codes are not copy; "tags" stays a leak in member copy,
except in `id3-tags`, whose string must tell the member the files carry ID3 tags and how to strip them (owner ruling
8). The two completeness warnings and the separate-tracks extras warning carry copy in the same deck. Uploader
sessions send no reject email; the page carries every verdict (UD:261-262).

## The journey and the API

**Hooks:** lands hook 4's shape on `submission_events`; neutral on hook 20. Neutral on hook 23: the worker and the
sweep are goroutines in `serve`. Neutral on hook 24: the stall event is recorded whether or not a mailer is
configured, and the sweep logs that no mail was sent when none is.

### Operations

Rung 7's JSON operations join `/api/v1` on 14a's mount, as catalog operations at member privilege, in the contract
document under 14A's additive version policy. They follow 14a's merged form: its seam pattern (14A:143-150), session
middleware, error body, and catalog. The plan reads 14a's merged code and names the seam type.

- `create_submission`: the manifest, the verdict unit, the source answer, the optional note and store URL, the route
  (`match` or `unlisted`), and for `unlisted` the form. Answers the submission id and each file's id. A refusal uses
  14a's error envelope with an additive `rejects` field: submission-level codes, plus a list of (path, code) pairs.
- `list_submissions`: the caller's own submissions, newest first, for "Your uploads" (PCF section 9).
- `get_submission`: one of the caller's submissions with its state, events, per-file verdicts, and warnings. The page
  polls it every 5 seconds (UD:131-137).
- `retry_submission`, `route_unlisted` (with the form), and `cancel_submission`: valid only from the states in the
  transition table; any other state answers `conflict`.

Every operation reads the caller from the verified session and returns only the caller's own submissions; another
member's id answers `not_found`, as an unknown id does. No operation takes an uploader, a collection, a state, a
reservation, a role, or a key as a request field; `opshttp`'s decoder ignores unknown keys (14A:171-173). The store URL
must parse as an absolute `https` URL with a host, within a length bound: `invalid_request` with detail `store_url`
otherwise (PS A14).

### The tus mount

The tus endpoint is its own subtree, `tus`, registered from a rung 7 file through `Shell.Register`. It cannot live
under `api`, because 14a's `api` subtree terminates every path under its prefix (14A:117-124).

### The startup pass, the worker, and the sweep

At startup, after taking the intake lease and before the listener opens, `serve` runs one pass:

1. **Reconcile.** Staging is never backed up (`internal/backup/doc.go:4-7`), so a restore can bring back rows whose
   bytes are gone. A non-terminal submission any of whose staged bytes ("Staged bytes, per state") are missing moves
   to `expired` with a `reason` naming them, which ends its reservation and frees its album key. A ZIP past its
   verdict is tested against its members, never its archive. The restore reconciler (`internal/backup/reconcile.go`)
   is the precedent.
2. **Clean up.** It removes every byte outside each submission's staged set: all of a terminal submission's bytes, a
   `submissions/` directory with no row, the archive of a ZIP past its verdict (never in `check-unfinished`), and the
   failing files and extras of an accepted separate-tracks submission. Cleanup takes
   each tus upload's `memorylocker` lock and leaves a file whose lock is held for the next sweep, so deletion is
   idempotent and never part of a transition.
3. **Receive** every complete, unreceived tus upload of an `uploading` submission.
4. **Re-queue** every `checking` and `matching` submission.

The pass logs its counts. This is intake's own work, not an ops job, so 14A's routing of `serve`'s job runner to a
later rung (14A:158-165) is untouched.

A sweep runs on a registered interval. It repeats steps 2 and 3, removes a `tus/` file with no row once it is older
than the resume window, and applies the sweep transitions in the table. An `uploading` submission's idle clock is the
later of its `updated_at` and the newest modification time of its tus data files, since a long PATCH writes no store
row. It records a `stalled` event for a submission in `checking` or `matching` past the stall threshold with no
earlier `stalled` event, and mails every live admin once through `internal/notify`, after the write commits and
outside the write callback, as 14A:280-285 sends its notice. PCF section 8 tells the member "The admin gets notified
automatically", and this mail makes it true. The threshold is set from the measured decode throughput.

## The member gate, the landing, and the styled 404

**Hooks:** neutral, hooks 23 and 24 included. The member gate reads the verify-identity seam, so a second sign-in
method at 0.9 reaches it without rework.

These are the `internal/auth` and `internal/serve` edits, in segment C or 7c, after 14a merges.

- **`RequireMember`** sits beside `RequireAdmin` in `internal/auth/admin.go`. It verifies through the same seam
  (`internal/auth/seam.go:107`), refreshes the cookie after the check passes as `RequireAdmin` does
  (`admin.go:129-131`), and places the verified identity on the context. A page request with no valid session answers
  `303 See Other` to the sign-in page: a member page's existence reveals nothing the public sign-in page does not. The
  tus wrapper and 14a's `api` middleware answer 401, because a script reads them.
- **The landing.** The uploader page is the signed-in member's landing page. The consumed-link page's `signed_in`
  state gains a Continue link to it, and the header wordmark targets it (B item 14). The shell's prefix root, which no
  subtree can claim (`internal/serve/serve.go:107-109`), redirects to it from the same `Shell.terminated` edit as the
  styled 404. 14a's plan leaves open "whether rung 7 and a running rung 12 slice may both write `internal/serve`"
  (`plans/2026-10-10-rung-14a-admin-api.md:815-816`); the conductor rules it before 7c's task that edits
  `Shell.terminated`.
- **The styled 404.** `Shell.terminated` answers unclaimed paths with `http.NotFoundHandler` (`serve.go:248-258`). 7c
  replaces that with the web error document at status 404 and routes `RequireAdmin`'s refusal through the same
  handler, so a member's 404 on an admin page and a 404 on an unregistered path stay byte-identical (B item 14;
  `admin.go:125-128`). A `web-auth-security-reviewer` read covers the task.

## The uploader page

**Hooks:** neutral, hooks 23 and 24 included (static assets embedded in the binary).

7b builds the page in SvelteKit static mode under the visual identity spec, which outranks any styling habit
(CLAUDE.md, Stack rules). The spec requires:

- **W7.19's four surfaces.** The drop target's states (idle, dragover, invalid type, a drop outside the target, the
  `dragleave`-on-child trap) with W6.10's single-pointer equivalent; the file picker as a first-class path, the iPad
  path; a multi-file progress surface that runs for minutes on a residential uplink; and the collection picker, hidden
  at 0.8 as above.
- **The kind choice**, "album" or "separate tracks", per submission, defaulting to album.
- **The source question** asked once per upload, required, with a per-submission override (PS:510-533). "Bought as a
  download" offers the optional store name or link; "Other" offers the optional note.
- **The unlisted route's choice and form**, with its contract stated before upload starts.
- **The journey view** per submission, polling `get_submission`, including `check-unfinished`'s PCF section 3 copy;
  **"Your uploads"**, listing in-flight resumables first with "Resume" (PCF section 9); and a batch's held creates
  when it exceeds the session cap.
- **The push-back view** with its fix paths: fix the names (back to selection), add the release to MusicBrainz (a link
  and a one-line how, W8.2's push-back exception), try again, or take the unlisted route.
- **The empty-library state** pairing a photo with "No albums yet. Upload the first one." (VI:274-275), which every
  box shows until rung 9 admits an album.
- **Copy** from PCF verbatim where PCF covers the surface, except the strings "Owed records" lists for redraft, which
  go through the same register read as the new reject reasons. "Album" gives way to "music" or "recording" wherever a
  member may be sending something else (owner ruling 1). PCF's in-library card belongs to rung 9 and the web pass.
- **VI throughout**: the two palettes, the inks, the component layer, the hard printed shadows, squared outline badges
  for the per-track chips, and the copy register. No hardcoded color.
- **Wake Lock** where supported (PCF section 2's "Keep my screen awake").

Every state is rendered and read at the rendered harness's widths in both palettes before Geoff sees anything
(`see-it-before-geoff-tests-it`). W11.5's fresh-context coherence read is the uploader's gate, at k=3 with 2-of-3
consensus (W:1298-1310). The salvage components (`_salvage/web/src/lib/components/`) are source material under fresh
review, never copies (BS:512-516).

## 7c: visual items and the sitting

**Hooks:** neutral, hooks 23 and 24 included.

- The photo loader as a leaf Go package, porting `tools/photos/check.mjs`'s two modes into a gate leg, and the CREDITS
  generator with name-only lines for pack grantors (`specs/2026-09-25-photo-curation-design.md:236-247`;
  `specs/2026-09-25-project-site-design.md:170-184`).
- The photo column on the login page, error-page photos, and the empty-library photo (VI:252-275).
- The styled 404 (above).
- 5c's carried items (B item 14 and BACKLOG #57, swept by the plan).
- VI's open items that need no catalog, built for a verdict at the sitting: the phone header, error alerts as a tinted
  panel, and motion (VI:356-372). The catalog line's format (VI:243-244), the list view, filtering, the album page, and
  the phone album grid move to the web pass after rung 9 (owner ruling 2); VI owes an erratum.
- W11.5's coherence read and the render-and-read fold.
- **The sitting.** Geoff drives real uploads on the workstation: a FLAC album as files, the same album as a ZIP,
  separate tracks, one reject, one push-back, and the unlisted route, at desktop and phone widths in both palettes,
  and rules VI's two remaining unruled items, the phone header and error alerts. The "before" is VI's committed
  mockups and the 5c pages as they stand (B question 1).

## Moved out of rung 7

| Item | Goes to | Reason |
|---|---|---|
| Hook 12, the portable layout and `cover.jpg` | rung 9 | conductor ruling 2; SS erratum owed |
| The backup guard's counting and `--max-delete` sizing | rung 9 | conductor ruling 2 |
| The template grammar, sanitizer, injectivity check, root-confined library writes | rung 9 | their first caller is the layout write (B D9) |
| The non-album layout path for separate tracks | rung 9 | owner ruling 6; beets' `singleton:` path is the precedent |
| Re-verifying `file_hash` as admission's tag-write copy reads the file | rung 9 | it costs nothing extra there and catches a staged file changed after the verdict |
| Whether `INBOX_DIR` and `QUARANTINE_DIR` stay reserved | rung 9 | the `pipeline` class edit above |
| The album page, library grid and list, filtering, buy-link rendering, the catalog line | the web pass after rung 9 | owner ruling 2 |
| The default-collection preference, per-upload override, re-validation at write, and the per-collection quota | 14b | the first verbs that make a second pickable collection or retire one, and the first quota setter |
| Matching loose tracks against MusicBrainz recordings | after 0.8 | owner ruling 6; the matcher and corpus stay album-only |
| The archive-upload carry from 4a's shadow gap | retired | the shadow is retired (RP:126-130); ZIP uploads stay in scope (PCF:39-41) |
| Capturing MusicBrainz buy links at admission | rung 9 | PS:762; an SS hook is owed |
| A non-Linux device check and free-space read, and the store's port | 0.9 | conductor ruling 7 |

## Forward hooks (SS:945-1040)

| Hook | Verdict | Section |
|---|---|---|
| 1, audio key and file hash | lands on `submission_files`; rung 9 copies them to `tracks` | The gate |
| 4, event-shaped tables | lands: `submission_events` carries `at` and `kind` | The submission model |
| 17, the uploading member | lands: `submissions.uploader` from the session alone | The submission model |
| 21, the source answer | lands: required on `submissions` | The submission model |
| 5, 8, 9, 15 | consumed: `UserID`, `writer`, the single writer, the collection row's fields and the quota rows | throughout |
| 12, portable layout | moved to rung 9 | Moved out |
| 18, description and mark | neutral: 14a merges first, so every key rung 7 registers carries both | Sequencing |
| 22, integer wire id | neutral: slice 3's migration backfills | The default collection |
| 20, serve events | neutral: submission events are not serve events | The journey |
| 23, no platform assumption in the core | lands: the device check and the free-space read sit in `_linux.go` and a `!linux` stated refusal; a scoped cross-compile holds the rest; the store's port is 0.9's | Staging |
| 24, an install with no public host | neutral: the tus base path reads `DUBPLATE_PUBLIC_URL`; the stall event needs no mailer; the member gate uses the identity seam | The transfer; The journey |
| 2, 3, 11, 13, 14, 16, 19 | neutral: each names rung 8, 9, 13, or a slice | n/a |
| 6, 7, 10 | retired (6) or rung 6's (7, 10) | n/a |

## Prior art

Each adopted library is judged on license, quality against the house bar, maintenance, and adoption (CLAUDE.md,
"Build for 0.8"). Versions were read on 2026-10-10.

| Entry | Verdict | License | Quality | Maintenance | Adoption |
|---|---|---|---|---|---|
| `github.com/tus/tusd/v2` v2.10.1 (proxy time 2026-09-14) | adopt, configured closed | MIT | the protocol's reference server; documented hooks; unsafe defaults (CORS any origin, GET download, DELETE, `x/exp/slog`) each closed above | five tags in 12 months; last push 2026-10-07; 85 open issues; #1159 open | 3,892 stars; 71 importers of `pkg/handler`; the salvage ran v2.10.0 |
| `tus-js-client` 4.3.1 (`latest`; `next` 5.0.0-pre3, 2026-09-30) | adopt, pinned exactly | MIT | documented options, `uploadUrl` among them; seven small runtime dependencies | last push 2026-09-30; 77 open issues; 5.0 in pre-release | 2,627 stars; the reference client |
| `github.com/mewkiz/flac` v1.0.14 (2026-08-19, `@latest`) | adopt | Unlicense | idiomatic package split; frame CRCs checked; depths outside 16, 20, and 24 closed by step 3 | v1.0.13 2025-07-11, v1.0.14 2026-08-19; 9 open issues | 363 stars; 96 importers; SS ruled "Adopt for rung 7" |
| `github.com/tus/lockfile` | decline | n/a | n/a | n/a | only `filelocker` needs it; the intake lease makes one process |
| salvage `intake.go` ZIP walk | adopt its shape | in-repo | a central-directory walk that "never extracts a member" | n/a | n/a |
| salvage `gate.go` (`flac -t`, `metaflac`) | decline | in-repo | runs subprocesses; "serve runs no subprocess" (`systemd/dubplate-server-serve.service:68-69`) | n/a | n/a |
| salvage `journey.go` | decline the structure, keep the state names | in-repo | coupled to the retired ledger | n/a | n/a |
| salvage `reject.go` | adopt the strings through PCF, restructure the set | in-repo | an `iota` set with switches | n/a | n/a |
| salvage `upload.go` | adopt the session-grouping idea only | in-repo | reads Cloudflare Access identity, retired | n/a | n/a |
| salvage `tus.ts`, `preflight.ts`, `platform.ts`, `wakeLock.svelte.ts`, 13 components | source material | in-repo | predate static mode and VI; the 524 branch declined | n/a | n/a |
| stdlib `archive/zip`, `crypto/md5`, `crypto/sha256`, `os.Root`, `path/filepath` | adopt | BSD-3-Clause | quoted above | Go team | n/a |
| `golang.org/x/sys/unix` (`Statx`, `Statfs`) | adopt; already a direct dependency at v0.49.0 | BSD-3-Clause | n/a | Go team | n/a |
| beets singleton import (`beet import -s`, the `singleton:` path) | consulted for separate tracks; behavior only | MIT | n/a | n/a | n/a |
| Picard's tag-mapping tables | consulted through `tagio`, which follows them (rung 8) | CC0 1.0 | n/a | n/a | n/a |
| chromaprint | decline for 0.8: SS places fingerprinting at 2.1 as a subprocess (SS:1083), and `serve` runs no subprocess | n/a | n/a | n/a | n/a |
| OpenSubsonic specification, Navidrome, gonic, taglib via wazero, goose, modernc sqlite | neutral | n/a | n/a | n/a | none bears on intake beyond the store in use |
| Melodee | consulted (`research/2026-10-06-slice-1-melodee-intake.md`) | MIT | lessons L2, L3, L4, L5, L7, L9, L12 adopted as tests above | n/a | n/a |

Adding `mewkiz/flac` brings `icza/bitio` v1.1.0 (Apache-2.0), `mewkiz/pkg`, and `mewpkg/term`; `go mod tidy` picks
`mewkiz/pkg` at a 2025-04-17 pseudo-version where a 2026-07-03 one exists, which the sweep weighs. tusd brings only
`golang.org/x/exp` once `filelocker` is not imported. The pass opens with the `dependency-upgrade` sweep; the
2026-10-10 sweep landed at `2af4404`, so 7a's is likely one line plus the three adoptions.

## Sequencing with rung 14a

14a lands first. These rows amend 14A's "Contested surfaces with the rung 7 lane" (14A:510-532):

| Surface | Rung 7's rule |
|---|---|
| `internal/auth` (`RequireMember`, the styled 404's hook in `RequireAdmin`'s refusal) | written only after 14a merges, on 14a's merged `admin.go` |
| `internal/serve` (`Shell.terminated`, the prefix-root redirect) | after 14a merges, and after the conductor rules the rung 12 question 14a's plan left open |
| `internal/ops`, `opshttp`, the contract document | rung 7's operations join after 14a merges, as catalog rows and document paths |
| route inventory (PS A2) | 14a lands it; rung 7 adds rows for `tus`, its pages, and its operations |
| config declarations (hook 18) | 14a lands the description and mark first; rung 7's keys carry both |
| `internal/config/keys.go` | rung 7 retires the `intake` class, removes `DUBPLATE_STAGING_DIR` from `pipeline`, and adds the wall row; 14a's wall-row edits are disjoint |
| `internal/engine/tagio` | rung 7 adds the block bound to the chain walk |
| migration `0006` and the older-store map | rung 7 |
| `cmd/dubplate-server/serve.go`, `doctor.go`, `web.go` tables | rung 7 adds from its own files and rows (14A:526-528) |
| `web/` | rung 7 owns it (PA:210), with 14a's `routes/members`, admin settings route, and API client as 14a's |
| `scripts/check.sh`, `scripts/check-state-dir.sh`, `scripts/deploy.sh`, `Makefile` | edited only after the gate scope pass merges |

**The seam.** Segments 7a-A and 7a-B write only rung 7's own packages (`internal/reject`, `internal/intake`, the
platform leaf), `tagio`'s walk, the store's migration and query files, rung 7's config file, and `keys.go`'s two
classes. Segment 7a-C opens with 14a merged: rebase onto master, run the full gate, then write the member gate, the
tus mount, and the operations. If 14a has not merged when 7a-B closes, the lane pauses at that green boundary and
records it in STATUS. The lane registry entry is written before 7a launches (PA:173).

## Planning-miss prevention

### Behaviors relied on, quoted

Quoted above: tusd's `MaxSize`, `DisableDownload`, `DisableTermination`, `Cors.Disable`, the default CORS origin, the
inert `Upload-Concat` without a concater, the `x/exp/slog` logger and its nil default, the two callbacks,
`FileInfoChanges.ID` and `MetaData`, `HookEvent.Context`'s values, the absolute `BasePath`, the deferred-length
refusal, the per-read deadline extension, `UseIn`'s composition, the filestore's permissions, offset, `O_TRUNC`,
missing `Sync`, and lack of cleanup, `GetUpload`, `sendError`, and `memorylocker`'s lifetime; `tus-js-client`'s
`uploadUrl`, `chunkSize`, `parallelUploads`, `storeFingerprintForResuming`, `retryDelays`, its success on a full
offset, its retry rule, and its `endpoint` requirement; `mewkiz/flac`'s `New`, `ParseNext`, the CRC errors, the
bit-depth codes, `Frame.Hash`, `NSamples`' zero, and its ID3v2 test; mutagen's ID3v2 and ID3v1 tests; `archive/zip`'s
insecure-path condition, the read bound, `NonUTF8`, `Reader.init`'s header loop and 16-bit count check, and
`readDirectoryHeader`; `filepath.IsLocal`; `os.Root`; `http.Server.ReadTimeout` and `NewResponseController`'s `Unwrap`
rule; `rename(2)` across mount points; `statx(2)`'s mount id and `stx_mask`; `f_bavail` in `f_frsize` units;
`store.Handle.Update`, `OpenServe`'s shared lock, and `TakeKindLease`; `tagio`'s `safeCastInt` and its per-block
allocation; the config registry's cross-family panic; `musicbrainz.NewClient`'s refusals, `Budget`'s lifetime limit,
and `ClassUserPresent`; `matcher.Action`. The plan quotes before writing: the SQLite pages on unique-index NULLs and
`ADD COLUMN` on a STRICT table; `tagio.Read`'s and `tagio.Parse`'s errors; each `tagio` field it projects;
`handler.NewError`; and `readDirectoryEnd`'s offset arithmetic, which the header count ports.

### Privileged fields and every channel that sets them

| Field | Channels at rung 7 |
|---|---|
| `submissions.uploader` (rung 9's `added_by`) | `create_submission`, from the verified session's `UserID` alone; never a request field or tus metadata |
| `submissions.collection_id` | the server, from the route's `intake_route` collection at create; `route_unlisted` on the caller's own `pushed-back` submission |
| `collections.intake_route`, `owner`, `writer`, `visibility`, `root_path`, `name` | the lazy creation in rung 7's write path alone; no API sets them at rung 7 (14b may) |
| `submissions.state` and every event | the state machine alone; member actions request a transition, which a compare-and-set validates |
| `submissions.source`, `source_note`, `store_url`, `verdict_unit`, `container` | `create_submission`, by the member, per submission; no channel edits them at rung 7 |
| `declared_bytes`, `submission_files.size` | the manifest at create; each file's `Upload-Length` must equal its declared size |
| `submission_files.role` | the server, from the extension; never a request field |
| `reserved_bytes` | the server, at create, at a ZIP's extraction, and at the true-up |
| `audio_key`, `header_md5`, `file_hash`, `album_key`, stream fields, `verdict`, extracted member rows | the gate alone |
| the match proposal | the probe alone |
| the entered album fields, `entered_artist`, `entered_title` | `create_submission` on the unlisted route and `route_unlisted`, by the uploader of that submission |
| the tus upload id | `PreUploadCreateCallback`, set to the server-minted file id |
| `received_at` | receipt alone |
| `member_quota.charged_bytes` | none at rung 7; rung 9's admission |
| `member_overrides` (`may_upload`, `quota_bytes`, `upload_concurrency_cap`) | read only; written by 14b, and before 14b by a direct store write alone |

### States each check meets, and the report each names

| Check | State | Report |
|---|---|---|
| config load | staging or library absent; the two nest lexically | loads; refused, naming both |
| startup device check | each row of its table | as that table says |
| startup | the intake lease is held | start refused, naming the holder |
| any rung 7 operation or tus request | no valid session | 401 |
| member page | no valid session | 303 to sign-in |
| tus request | each refusal | as "Refusals on the tus path" says |
| tus create | two creates for one file at once | the second truncates; the gate refuses the file as `corrupt` |
| receipt | last byte written, receipt not committed (crash or store error) | received at the next startup pass or sweep |
| receipt | response lost after receipt | the client's HEAD answers the full offset; success |
| receipt | two files finish at once | exactly one move to `checking` |
| receipt | the submission closed during a PATCH | no change; 403; cleanup removes the bytes |
| receipt | a store error | 503, retried; the sweep finishes receipt |
| `create_submission` | manifest empty | `no-audio` |
| `create_submission` | malformed (bad JSON, negative size, duplicate path) | `invalid_request` with a detail |
| `create_submission` | many bad files | every failing path listed with its code in `rejects` |
| `create_submission` | source absent, empty, or unknown | `missing-source` |
| `create_submission` | store URL malformed or not `https` | `invalid_request`, detail `store_url` |
| `create_submission` | each limit | its code, first failing limit only |
| `create_submission` | cap NULL | no cap check |
| any write | floor crossed; `statfs` fails | `no-space`; `could-not-check` |
| ZIP | more than `MAX_FILES` central-directory headers, whatever the end record claims | `too-many-files`, before `NewReader` |
| ZIP | does not open, member errors, or CRC fails | `unpack-failed` |
| ZIP | symlink, refused name, directory with content, or duplicate name | `unsafe-path` |
| ZIP | declared total over cap; extracted total over a limit | `too-large` before extraction; `over-quota` or `no-space` |
| ZIP | non-UTF-8 name | accepted, kept as bytes |
| ZIP | holds no audio file | `no-audio` |
| ZIP | re-run after a crash mid-extraction or before the verdict | `x/` removed, the same verdict as an uninterrupted run |
| ZIP | `check-unfinished`; past its verdict | the archive kept, `x/` removed; the members kept, the archive removed |
| audio file | each per-file check | its reason as listed under "The gate" |
| audio file | header MD5 all zero | stored NULL; no comparison |
| audio file | a block over its cap | refused before the block is allocated |
| audio file | first bytes `ID3`; a decode error with a trailing `TAG` record | `id3-tags`; `id3-tags` |
| album | many bad files | `rejected`; each file's first failing check recorded |
| album | each position fault | `duplicate-track-position` or `incomplete-album` |
| album | track tag non-numeric or at or below 0 | `missing-track-number` |
| separate tracks | some files fail; all fail; every passing file a duplicate | `accepted` with the passing files; `rejected`; `rejected` |
| separate tracks | an extra in the manifest; an extra in a ZIP | `invalid_request`, detail `extra`; dropped, with the warning |
| accept | duplicate album; duplicate separate track | `rejected` with `already-in-library`; that file's verdict |
| gate | a check cannot run | `check-unfinished`, bytes kept |
| gate | a store error during the verdict | stays `checking`; the startup pass or stall event |
| probe | each outcome | as the probe table says |
| member action | from a state the action does not leave | `conflict` |
| `get_submission` | unknown id or another member's | `not_found` |
| sweep | each window passed | `expired`, cleanup follows |
| sweep | `checking` or `matching` past the threshold | one `stalled` event, one mail to each live admin, never repeated |
| startup pass | staged bytes missing for a non-terminal row | `expired` with a reason |
| startup pass | a `submissions/` directory with no row, or a terminal row's bytes | removed |
| default collection | no live admin at first creation | `internal`, logged |
| non-Linux build | the device check or free-space read | start refused with the stated message |

## Acceptance criteria

Each criterion names the fixture state it needs, so none can pass on an empty fixture. Every test that moves a
submission between states (11, 19, and 22 to 29) ends by asserting that each `submissions.state` equals its latest
state event's `kind`; criterion 4 proves that assertion can fail.

1. `DUBPLATE_STAGING_DIR` is a required wall row with no default, and registering it does not panic, because the
   `pipeline` class no longer lists it. Config load refuses staging inside the library root and the reverse, compared
   lexically, and loads when staging is absent. Setting `DUBPLATE_UPLOADS_DIR` or `DUBPLATE_TRIGGER_DIR` is a load
   error naming `DUBPLATE_STAGING_DIR`. Fixture: one config per case.
2. The startup check's pure function returns each row of its table, symlink-resolved nesting and a mask lacking
   `STATX_MNT_ID` included, from a table of stat results. The thin reader passes once against a temporary directory
   holding both paths.
3. The serve unit grants one music-volume entry holding both paths. A one-time mutation proof, recorded in the task
   report, shows the state-dir leg failing when either template path leaves that grant and when a second non-leaf
   grant is added. `deploy.sh` pre-creates staging and the library root.
4. Migration `0006` passes every store invariant test, carries R-6d4's keep limit, adds its tables to the older-store
   map, and refuses a second `rel_path` in one submission and a second collection per route. A fixture submission
   whose state was written out of step fails the state-equals-event assertion.
5. The first main-route `create_submission` on an empty catalog creates one "Library" collection owned by the earliest
   live admin, `writer` `server`, `visibility` `group`; a second create reuses it. Fixture: two live admins with
   distinct creation times and one tombstoned earlier admin. The first unlisted submission creates one "Unlisted
   music" collection the same way. The no-nesting check, run on two literal roots, refuses the nested pair.
6. The path validator refuses an absolute path, an empty, `.`, or `..` component, a NUL, a backslash, and a path
   unequal to its clean form; a table test runs each through the validator and a direct insert and asserts the same
   refusal, and a manifest path and a ZIP name of each kind answer `unsafe-path`.
7. The tus mount answers no `Access-Control-*` header, 405 to GET and DELETE on an upload, 400 to `Upload-Concat` and
   to `X-HTTP-Method-Override`, 501 to `Upload-Defer-Length`, and a `Location` built from `DUBPLATE_PUBLIC_URL` when the
   request carries another `Host`. Fixture: a created upload.
8. With no session, a tus request answers 401. A member's PATCH or HEAD on another member's upload answers 404 with the
   unknown-id body, and the other's file is unchanged. Fixture: two members, each with an in-flight upload holding
   bytes.
9. A PATCH whose body arrives over longer than `ReadTimeout`, with reads spaced under `NetworkTimeout`, completes when
   the tus subtree is mounted through `Shell.Register` and the shell's full middleware, using the production server
   constructor with only `ReadTimeout` lowered. The test fails when the wrapper's `Unwrap` is removed.
10. A creation's client metadata never reaches `.info`, and the tus id equals the file id. Fixture: a creation carrying
    extra keys and a forged `uploader`. A creation declaring one byte more than the manifest, and a second creation
    for a file whose upload exists, each answer 400 before `.info` is written. No definitive callback refusal answers
    500.
11. Receipt: a fault injected between the last byte and the receipt commit, then a restart, ends with the file
    received and the submission in `checking`; a response dropped after receipt ends with the client reporting
    success; a lost 201 ends with the file sent through `uploadUrl`; two files finishing at once move the submission to
    `checking` once; a PATCH finishing after a cancel changes nothing and its bytes are removed. Fixture: a two-file
    submission.
12. Resume: a submission with one file half sent, after a page reload and a re-pick, resumes that file from its offset
    on the same submission, with use and session count unchanged; a re-picked file matching no row is named. Fixture:
    the half-sent submission.
13. `create_submission` sets `uploader` from the session; a body naming another member as `uploader`, a collection, a
    state, a reservation, or a role changes none of them (PS A10). Fixture: two members.
14. A submission without a source, with an empty source, and with an unknown value each answer `missing-source`, and
    each of the seven values is stored (PS A9).
15. Two concurrent creates that each fit a member's cap and together exceed it: exactly one succeeds; the same for the
    group cap and for the floor with an injected `statfs`. The limit functions take `*store.Tx`, so the property is
    compiled in; one plain concurrent test per property backs it. A member with a NULL cap creates several submissions
    whose sum exceeds the other fixture member's `quota_bytes`, all under the album cap and the floor, and every one
    succeeds. Fixture: one override with `quota_bytes` set, one NULL.
16. A rejected, cancelled, or expired submission stops counting toward use; an accepted one keeps counting. A
    deflated ZIP whose archive fits a member's cap and whose extracted total does not is rejected `over-quota` before
    extraction, and against an injected `statfs` is rejected `no-space`; a passing deflated ZIP's reservation trues up
    to its extracted total. Fixture: one submission per state for one member, and the deflated ZIPs.
17. Every reason in the set has a code, a label, a friendly sentence, a next step, a blocking flag, and, when
    parameterized, a canonical wording. The nine adopted reasons equal PCF section 4's text, transcribed as literals.
    The new strings carry a `prose-voice-reviewer` record. A table-driven test keyed by reject code produces every
    blocking code from a create, tus-path, or gate fixture, and fails when the set holds a code its table lacks;
    `symlink` is exempt, with its reason written in the table.
18. The ZIP walk refuses a symlink member, an absolute name, a `../` name, a NUL name, a backslash name, a directory
    entry with content, and a duplicate name with `unsafe-path`; refuses a declared total over the cap before writing
    a file; refuses more than `MAX_FILES` central-directory headers before `zip.NewReader` runs, including when the
    end record understates the count; and stores a non-UTF-8 name's bytes unchanged. Fixture: one crafted ZIP per
    case, a many-entries ZIP, a many-entries ZIP whose zip64 end record claims one entry, and a valid one that
    passes.
19. A ZIP member whose data exceeds its declared `UncompressedSize64` fails with `unpack-failed` and leaves no file.
    A ZIP stopped partway through extraction, and one whose extraction finished but whose verdict did not commit, each
    reach the uninterrupted run's verdict after a restart. Fixture: a header-understating ZIP and a passing ZIP with
    injected stops.
20. The gate rejects a flipped frame byte and a truncated file (`corrupt`); a nonzero header MD5 that differs from the
    decode and a sample count unequal to a nonzero `NSamples` (`source-damaged`); a 22.05 kHz file and a 12-bit file
    (`below-cd-quality`); an 18-bit and a 32-bit file (`unsupported-flac`); a zero-sample file (`empty-track`); and a
    tag block, a picture block, and a chain over their caps, each refused before the block is allocated. A file with
    an ID3v2 prefix and a file with a trailing ID3v1 record each answer `id3-tags`, and a clean file whose last 128
    bytes begin `TAG` by chance passes. A clean file
    passes, its `audio_key` equals its header MD5 and its `file_hash` equals `sha256sum`'s output; an all-zero header
    MD5 passes with `header_md5` NULL; a retagged copy keeps its `audio_key` and changes its `file_hash`. Fixture:
    generated FLACs for each case, made by the plan's fixture generator.
21. One failing track rejects an album, and its passing tracks are not kept. A two-disc album restarting on disc 2
    passes; a duplicate position, a gap, numbering that starts at 2, a disc gap, a `TRACKTOTAL` or `DISCTOTAL`
    mismatch, and disagreeing totals are rejected; an `A1`, `A2`, `B1` album answers `missing-track-number`; disc 0
    reads as disc 1; no tag is rewritten. Fixture: one album per case.
22. Separate tracks: three files, one corrupt, end `accepted` in "Unlisted music" with two passing files, the corrupt
    file's verdict recorded and its bytes removed; three failing files end `rejected`; no track number is required;
    the fake MusicBrainz transport records no request. Two passing files that both duplicate accepted files end
    `rejected`. A manifest naming an extra answers `invalid_request`; a ZIP of two tracks and a scan ends `accepted`
    with the two tracks and the extras warning. Fixture: the two three-file submissions, an accepted twin pair, and
    the ZIP.
23. An album uploaded twice to one collection: the second is refused `already-in-library`. The same album on the
    unlisted route while the first sits in "Library" is accepted. Two identical albums racing through `matching`, and
    two racing through `route_unlisted`: exactly one is accepted in each race. A separate track whose audio key equals
    an accepted file's in "Unlisted music" gets `already-in-library`. Fixture: an accepted submission and each racing
    pair.
24. A staged-file read fault, a `tagio` non-sentinel error, and a `statfs` fault at extraction each move a passing
    two-track album to `check-unfinished` with its bytes kept; `retry_submission` after the fault clears reaches the
    album's verdict; the album without a fault passes. The same album as a ZIP, faulted during decode, holds its
    archive and no `x/` in `check-unfinished`, and its retry re-extracts and passes. A `statfs` fault at create answers
    `could-not-check`. Fixture: the two-track album as loose files and as a ZIP.
25. The probe stores a proposal and moves to `accepted` on `ActionApply`, to `pushed-back` on another action and on
    `ErrNoCandidates`, and to `match-unavailable` on `ErrUnavailable`, with `ClassUserPresent` on its requests. More
    probes than one budget's limit all reach the transport. Fixture: a fake transport with a strong match, a weak
    match, no candidates, and an outage.
26. A table of every journey state against `retry_submission`, `route_unlisted`, and `cancel_submission` asserts the
    transition table's move or `conflict` in each cell, and that `route_unlisted` lands the submission in "Unlisted
    music" without a second upload. Fixture: one submission per state with staged files. The one race a cancel can
    meet, a receipt finishing after it, is AC 11's.
27. A second `serve` on one state directory refuses at the intake lease before any recovery runs. After a restart with
    one submission in `checking` and one in `matching`, both reach a verdict. A store holding an `accepted` and a
    `checking` submission over empty staging starts with both `expired`, and a re-upload of that album is not refused.
    A `submissions/` directory with no row and a rejected submission's leftover bytes are removed. An `accepted` ZIP
    whose archive is gone and whose members are present, and a `check-unfinished` ZIP holding only its archive, both
    survive the restart unchanged. Fixture: the restored store, the orphaned files, and the two ZIPs.
28. The sweep expires an idle upload and each held state past its window, and leaves alone an upload whose tus data
    file was written within the window though its `updated_at` is older. Two sweeps over a long-running check record
    one `stalled` event and send one mail to each of two live admins and none to a removed admin. Fixture: an injected
    clock, two live admins, and one removed admin.
29. `get_submission` and `list_submissions` return only the caller's own submissions; another member's id answers
    `not_found`. A refused create carries every failing path in `rejects`. Every rung 7 operation is in the contract
    document and the catalog and validates under 14a's contract tests. Fixture: two members with one submission each.
30. A signed-out request for the uploader page answers 303 to sign-in; a signed-in member gets the page. The prefix
    root redirects to the landing page. An unregistered path and a member's request for an admin page answer the same
    styled 404, byte for byte. Fixture: one admin and one member session.
31. The route inventory carries a row with an auth class for `tus`, each rung 7 page, and each rung 7 operation, and
    its test fails on a fixture inventory missing the `tus` row.
32. The uploader page builds W7.19's four surfaces, the kind choice, the source question with per-submission override,
    the unlisted route's form for albums and tracks, the journey views with `check-unfinished`, the push-back view,
    "Your uploads" with resume, a batch held at the session cap, and the empty-library state. Every state is rendered
    and read in both palettes at the harness's widths, passes the `web/` leg's gates, and passes W11.5's coherence
    read. `svelte-reviewer` and `daisyui-a11y-reviewer` read it against VI.
33. `go.mod` requires tusd v2.10.1 and `mewkiz/flac` v1.0.14, `web/package.json` pins `tus-js-client` 4.3.1 exactly,
    and `go list -deps ./cmd/dubplate-server` names no `filelocker` or `lockfile` package.
34. `GOOS=darwin` and `GOOS=windows` builds of `internal/reject`, the gate core, and the platform leaf succeed, and
    every Linux-only call rung 7 adds sits in a `_linux.go` file. The plan names where the cross-compile runs once the
    gate scope pass has merged.

## Owed records

Errata to ratified documents, none edited by this spec. These are living-reference edits the `intake` passes make as tasks, under the adapter's Errata rule:

- SS: hook 12 moves to rung 9 (conductor ruling 2); a new forward hook for rung 9's MusicBrainz buy-link capture
  (PS:762).
- CH: the routing row (CH:289) narrowed by `intake_route`; the per-member counter counts what the member added; the
  startup refusal supersedes CH:474's copy fallback for staging and the library root.
- 14A: moot (frozen; lean ruling 8). Its contested-surfaces rows for `admin.go` and `internal/serve` are sequenced by "Sequencing with rung 14a"; its
  `keys.go` row gains the `pipeline` edit; the subtree-adoption clauses (14A:36-38, :522, :634) are moot.
- W: W8.2 gains the format-name exception for reject reasons (conductor ruling 6), and "ID3 tags" in `id3-tags`
  (owner ruling 8).
- VI: the catalog line's format (VI:243-244) and the list view, filtering, album page, and phone layouts in "Open for
  rung 7" (VI:350-372) move to the web pass after rung 9; the sitting rules two items.
- BS: moot (frozen; lean ruling 8). Rung 7 re-bands from M (BS:798) to L.
- PCF: the retry line (`passb-copy-final.md:124-126`) redrafted for owner ruling 7; the three Navidrome strings (`:105`,
  `:143-145`, `:160-161`); every "Private Recordings" string in sections 5 and 6, renamed "Unlisted music" with
  "music" or "recording" where "album" would be wrong; the section 1 hint "One album at a time. Albums go in whole."
  (`:45`), which a multi-submission upload contradicts. The redrafts go through the register read.
- PCF section 4, the copy contract: the `id3-tags` string (owner ruling 8, Geoff 2026-10-10), which names ID3 tags
  and how to strip them, and the separate-tracks extras warning, added after the register read with the other new
  reasons.
- CLAUDE.md: "The album is the verdict unit" gains the separate-tracks kind (owner ruling 6); the Gate paragraph's
  state-dir leg admits the serve unit's music-volume grant; the fact line on `/var/lib/dubplate/` gains that staging
  lives on the music volume by design.

Other records:

- BACKLOG #8: the `tus-js-client` watch item names 5.0.0-pre3 (2026-09-30), not 5.0.0-pre2. The fix is T3's (row I67).
- B's note that ruling 10 was "not yet in the tree" is discharged
  (`research/2026-09-02-greenfield-server-rulings.md:48-65`).

## Ceiling and clock

The plans derive tasks, segments, ceilings, and clock estimates from B section 7. Revision 2 adds receipt and the
startup pass, the intake lease, the held state, the ZIP and FLAC bounds, separate tracks, and resume, and drops two
platform implementations, the chunk setting, and the accept-time re-route. The working figures: 7a 9.0M (flag 7.2M),
7b 6.8M (flag 5.4M), 7c 7.5M (flag 6.0M), about 23.3M for the rung (revision 1: 22.5M), and about 17 hours of
execution excluding the sitting and owner pauses (revision 1: about 16).

## Rulings for Geoff

None open. The ID3 fork of revision 2 was ruled by Geoff on 2026-10-10 (owner ruling 8).
