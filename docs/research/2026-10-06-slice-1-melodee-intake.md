# Slice 1, Task 5: Melodee's intake pipeline

Date read: 2026-10-07. Staged outside the repo; lands at A-land as
`docs/superpowers/research/2026-10-06-slice-1-melodee-intake.md`.

**Pin read (Task 0 record).** Melodee release `v2.2.0` (2026-07-11) = commit
`d91342e5d07632b7a1707b62acd595c080806f64`, `prerelease: false`, MIT
(https://github.com/melodee-project/melodee/blob/main/LICENSE; the file in the clone at the pin opens "MIT License",
"Copyright (c) 2024 Steven Hildreth"). Cited below as `M:` for `melodee@d91342e:src/Melodee.Common/` unless a path
starts elsewhere. The clone is untrusted data; no code in it was built or run.

**Channels.** Every claim names one: *source* (read at the pin), *docs* (Melodee's own `docs/pages/*` and README at the
pin), or *observation*. Melodee has no oracle in this study, so the only observation is one directory-mtime probe run
on this workstation (finding F13). A behavior stated from source alone was read, never run, and the record says so
where a runtime premise remains.

**Licence posture.** Melodee is MIT, so short quotations appear with their citation. The record states ideas. Code
worth reusing is named by path for the rung that would reuse it, and none is copied here. Melodee is C# and dubplate is
Go, so "reuse" means an algorithm, a table, or a schedule, never a transplant.

## Questions this record owns

1. What does Melodee do at each stage of its pipeline (inbound, staging, validation, admission into the library, the
   database insert)? *Answered, source.*
2. How does that compare with dubplate's intake-as-the-one-door model? *Answered.*
3. What lesson does each stage teach dubplate, positive or negative, and which rung does it bear on (7 intake, 9
   admission, 13 decision queue)? *Answered; 15 lessons, each with its rung.*
4. What goes in the prior-art table's Melodee quality cell (spec `:870`), against the house bar? *Answered; proposed
   text below.*
5. Where do Melodee's own docs and its source disagree? *Answered for four points; one runtime premise unconfirmed
   (F9, Quartz's `TriggerJob`).*
6. Which Melodee pieces are worth reusing by path? *Answered, three ideas, two declined.*

## The one-paragraph answer

Melodee's pipeline is four filesystem stages run by timers. A drop folder (inbound) is scanned every ten minutes, each
album directory is parsed and validated into a sidecar `melodee.json` and copied to a staging folder, a second timer
moves every album whose status is `Ok` from staging to a storage folder, and a third timer (daily by default) reads the
sidecars in storage and inserts rows into PostgreSQL. State lives in files and directory names; the database learns
about an album last. The verdict is tag completeness, never match confidence, and a human reaches an album only after
it has been staged and failed. dubplate's door is the reverse on every axis that matters: a contributor is present, the
album is checked and matched before anything lands, the verdict is all-or-nothing, and the catalog row is written by
the admission itself. Melodee's best ideas are small and sit around the edges (the unstable-file check, the bounded
revalidation backoff, reasons as a closed set). Its worst defects sit in the commit step (copy-then-delete without
verification, a merge that discards, errors swallowed) and are exactly what dubplate's custody bar exists to forbid.

## dubplate's model, as the study compares against it

Facts below are the repo's own specs, cited so the comparison table has a fixed other side.

- **One door.** "Intake stays the one door" (`streamer-2-0-design.md:184`); contributions enter through the one gate,
  and `restore` is the admin bulk path outside that claim (`organizer-design.md:212-213`).
- **Album is the verdict unit.** One failing track rejects the batch before anything lands (`uploader-design.md:40`;
  `CLAUDE.md`, "Facts agents get wrong"; build spec rung 7, `greenfield-server-build.md:798-820`).
- **The contributor is present.** The uploader holds the contributor on the page through a synchronous tag check; "the
  uploader accepts an album only when beets can properly tag it", otherwise push-back with the fix in the contributor's
  hands (`uploader-design.md:232-250`). Journey states: uploading, checking, matching, importing, in-library,
  pushed-back, queued, stalled (`:260-261`).
- **Reject wording is a contract** (`internal/reject`; `uploader-design.md:78-81`).
- **The timer path is an admin-only side door**, and the review queue is the admin backstop for it
  (`organizer-design.md:651-654`); the custody policy decides who resolves an uncertain match (`:295-330`).
- **The format allowlist is config, bounded by tag I/O** (rung 7, `greenfield-server-build.md:803-810`); FLAC-only at
  launch.
- **Hooks that touch intake.** Hook 1 (content hash and decoded-audio key computed at the gate), 3 (art blob at admit),
  12 (portable layout), 16 (a decision-queue item that names an admitted entity), 17 and 21 (the uploading member and
  the source answer on the submission, carried to the album row) (`streamer-2-0-design.md:767-855`).

## What Melodee does, stage by stage

### Stage 0: the drop and the upload

- **Source of albums.** Inbound is a library of type `Inbound`, "should be 1" (`M:Enums/LibraryType.cs:7-10`). The
  documented way in is to place files there (docs: `docs/pages/libraries.md`, "Inbound Library"; README, "Music
  Ingestion Pipeline"). A web upload exists beside it.
- **The upload page** accepts ZIP files only (`accept=".zip"`, `src/Melodee.Blazor/Components/Pages/Media/UploadAlbumDialog.razor:24`),
  caps each at 500 MB (`:204`), requires an inbound library that is unlocked (`:159-169`), and extracts straight into
  a new folder inside inbound with `overwriteFiles: true` (`:184-213`). It then reports "Complete" and closes the dialog
  (`:247-260`) after triggering the inbound job with chaining on (`:271-297`). The page is limited to the roles
  Administrator and Editor (`Library.razor:35`, `[Authorize(Roles = "Administrator,Editor")]`). No field records who
  uploaded: a grep for `UserId`, `AddedBy`, and `UploadedBy` over `Data/Models/Album.cs` and `Models/Album.cs` finds
  none.
- **Verdict at the door: none.** Success means "unzipped". Whether the album is any good is learned later, by timers,
  and only by someone who opens the staging screen.

### Stage 1: inbound to staging (`LibraryInboundProcessJob`, every 10 minutes)

Source: `M:Jobs/LibraryInboundProcessJob.cs`; seed schedule `M:Data/MelodeeDbContext.cs:1157-1160` (`0 */10 * ? * *`).

- **Skip conditions.** No path configured (`:72-76`), library locked (`:78-83`), or "does not need scanning" on a
  scheduled run (`:85-93`). A manual trigger or force flag bypasses the last check (`:151-165`).
- **"Needs scanning" is the root directory's own mtime.** `Library.NeedsScanning()` returns true when `LastScanAt <
  LastWriteTime()` and `LastWriteTime()` is `Directory.GetLastWriteTimeUtc(Path)` of the library root
  (`M:Data/Models/Library.cs:41-56`). See F13 for what that misses.
- **Each directory is one candidate.** `ProcessDirectoryAsync` renames top-level inbound folders whose name carries a
  dot, replacing "." with "_", in place in the drop zone (`M:Services/Scanning/DirectoryProcessorToStagingService.cs:357-377`),
  then walks every subdirectory modified since the last scan and processes them in parallel up to the core count
  (`:379-447`).
- **Stability check.** Before reading a directory the service snapshots every file's size and last-write time, waits 250
  ms, snapshots again, and defers the directory if anything changed or appeared (`:583-592`, `:1378-1409`). It is a
  heuristic for "the copy is finished". No fingerprint beyond size and mtime is taken.
- **Per-directory pipeline** (`:562-848`): delete any existing sidecar unless `processing.dontDeleteExistingMelodeeDataFiles`
  (`:605-617`); run directory plugins that turn CUE, SFV, M3U, NFO, or Blackbeard files into album JSON (`:641-685`,
  registered `:170-192`); run conversion plugins on every file (`:691-769`); if no album JSON exists, build albums from
  media files (`:778-796`); load each sidecar (`:798-821`); then per-album work (`ProcessAlbumsAsync`, `:850-1349`).
- **Conversion default is destructive.** `conversion.enabled` seeds `true` with bitrate 384 (`M:Data/MelodeeDbContext.cs:385-398`),
  and `MediaConvertor` converts every media file whose sniffed format is not MP3, FLAC included, to MP3 through ffmpeg
  (`M:Plugins/Conversion/Media/MediaConvertor.cs:17-95`; `M:Metadata/AudioTags/AudioTagManager.cs:111-145`). The
  output is accepted when a 16-byte header sniff says MP3 (`M:Metadata/AudioTags/FileFormatDetector.cs:5-60`;
  `MediaConvertor.cs:140-156`). The original is deleted only if `processing.doDeleteOriginal` (seeded `false`,
  `Data/MelodeeDbContext.cs:493-496`) is on, otherwise it is renamed or kept.
- **Staging copy.** Songs and images are copied (optionally moved) into `staging/<album dir>` in a batch
  (`DirectoryProcessorToStagingService.cs:1004-1022`), tags are rewritten in the staged copy when a repair changed them
  (`:1024-1082`).
- **Inline network enrichment.** The service asks an artist search engine for the album artist, takes the single
  highest-ranked hit, copies identifiers (MusicBrainz, Spotify, iTunes, and others) onto the album, and sets
  `Status = Ok` (`:1086-1153`, `OrderByDescending(x => x.Rank).FirstOrDefault()` at `:1101`). If the album has no
  image it searches for one and downloads the URL the search returned (`:1163-1241`). A run-context object caches
  search results across a run (`:1092-1097`).
- **Validation and the sidecar.** `_albumValidator.ValidateAlbum(album)` sets `Status` and `StatusReasons` (`:1248-1251`);
  the whole album object, messages included, is serialized to `melodee.json` in the staged folder (`:1253-1260`). That
  file is the album's state from here on.
- **Optional auto-repair ("magic").** When `magic.enabled` (seeded `true`, `Data/MelodeeDbContext.cs:305-308`) the
  service runs `DoMagic` on the album and re-writes the sidecar (`:1282-1296`). `DoMagic` runs only on albums that fail
  validation, applies each repair behind its own flag, and re-validates (`M:Services/Scanning/MediaEditService.cs:171-269`).
- **Result.** Counts of new albums, artists, and songs, plus a scan-history row (`LibraryInboundProcessJob.cs:105-136`).
  When the run was scheduled and anything was processed, it fires the next job (`:138-143`). The whole body sits in
  one `catch (Exception)` that only logs (`:145-148`).

### Stage 2: validation (`AlbumValidator`)

Source: `M:Plugins/Validation/AlbumValidator.cs`; enums `M:Enums/AlbumStatus.cs`, `M:Enums/AlbumNeedsAttentionReasons.cs`,
`M:Models/Validation/ValidationResultMessageSeverity.cs`.

- **Output.** A status (`Ok`, `New`, `Invalid`), a `[Flags]` set of "needs attention" reasons (22 named bits, for
  example `HasInvalidArtists`, `HasNoImages`, `SongsAreNotUniquelyNumbered`, `IsNotStudioTypeAlbum`), and a list of
  messages each tagged `Undesired` ("should be fixed") or `Critical` ("must be fixed").
- **The rule that decides.** `Ok` when no message is `Critical`, otherwise `Invalid` (`AlbumValidator.cs:129-131`).
  `Undesired` findings (unwanted text in a title, `:333-394`) never block.
- **Checks that are `Critical`** (`:35-269`, `:396-445`): no songs, no tags, invalid artist, an invalid song, invalid
  year, no title, song numbers not sequential from 1, duplicated song numbers, songs with differing album artists
  unless the album is Various Artists, song total tag not equal to song count, no cover image, artist unknown to the
  search engine, not a studio-type album, a song with zero or missing duration, fewer than the minimum songs or minutes.
- **Seeded thresholds.** Minimum songs 3 and minimum album duration 10 minutes (`Data/MelodeeDbContext.cs:690-705`),
  so a two-track single is `Invalid` by default. Both are settings, and "0 disables".
- **"Known artist" is weak.** The check passes when the artist has a search-engine unique id or `IsValid()`
  (`AlbumValidator.cs:208-229`), and `IsValid()` is true when any one of four ids is set and the name is non-empty
  (`M:Models/Extensions/ArtistExtensions.cs:74-82`). A top-ranked search hit therefore satisfies it. No distance or
  confidence enters the verdict.
- **No audio integrity check.** The only format checks read are header sniffs (`FileFormatDetector.cs:5-60`). A grep for
  `corrupt|integrity|decode` over the project's `.cs` files hits tag readers, image code, search services, and
  artist-search corruption counters in the run context (`M:Services/Scanning/DirectoryRunContext.cs:22-24`), none an
  audio-frame check. This is a negative from a grep, so it is marked: no audio decode step was found in the files read.
- **Track numbers ignore the disc.** `SongNumber()` returns the track-number tag alone
  (`M:Models/Extensions/SongExtensions.cs:257-260`), and `AreSongsUniquelyNumbered` and `AreAllSongNumbersValid` group
  and sequence on it (`AlbumValidator.cs:263-288`, `:396-445`). A two-disc album whose tracks restart at 1 fails both.
  The repair for that is `RenumberSongs`, documented in its summary as "This removes any concept of multi disc as with
  digital media it is somewhat irrelevant" (`M:Services/Scanning/MediaEditService.cs:734-736`). It rewrites each track
  number to 1..N, derives a sort order from the disc (`:737-760`, multiplier 10000, `:40`), and sets the disc total to
  1 in the tags (`album.SetTagValue(MetaTagIdentifier.DiscTotal, 1)`). `magic.doRenumberSongs` seeds `true`
  (`Data/MelodeeDbContext.cs:315-318`).
- **A check that throws does not fail the album.** The twelve checks run inside one `try` with
  `catch (Exception e) { Trace.WriteLine(e); }` (`AlbumValidator.cs:109-127`), and the verdict is then computed from
  whatever messages exist (`:129-131`). A check that throws before it adds its `Critical` message leaves the album
  `Ok`, as far as the source shows. This is a static reading; it was not run.
- **One validator instance serves many albums.** The validator keeps its message list and reasons in instance fields and
  clears them at the start of `ValidateAlbum` (`:31-33`, `:45-46`). The staging service builds one instance at
  initialization (`DirectoryProcessorToStagingService.cs:152`) and calls it from directories processed in parallel
  (`:389-447`, `:1248`). Nothing in the read file serializes those calls, so concurrent albums can overwrite each
  other's messages. Static reading; not run (F11).

### Stage 3: staging to storage (`StagingAutoMoveJob`, every 15 minutes)

Source: `M:Jobs/StagingAutoMoveJob.cs`; `M:Services/LibraryService.cs:487-655`, `:1009-1233`;
`M:Services/ServiceBase.cs:56-227`; `M:Models/Extensions/FileSystemDirectoryInfoExtensions.cs:556-598`.

- **Admission is a predicate on a sidecar.** The job finds every `melodee.json` under staging
  (`LibraryService.cs:1081`), loads each, and moves those whose `Status == Ok` (`StagingAutoMoveJob.cs:102-107`). No
  human and no confidence threshold stands between `Ok` and storage.
- **The destination is the first unlocked storage library** (`StagingAutoMoveJob.cs:83`); there is no routing.
- **Target path.** `<storage>/<artist dir>/<album dir>` computed from the album's tags (`LibraryService.cs:511-513`).
- **The move.** If the target directory already holds files, the album is *merged* (below). Otherwise
  `album.Directory.MoveToDirectory(path)` runs (`:560`), which, per its own summary, "must be copied and then deleted"
  because a cross-mount rename fails: it copies each file with overwrite, returns silently if a path is too long, then
  deletes the source tree (`FileSystemDirectoryInfoExtensions.cs:552-598`, `:575-580`, `:593`). No checksum compares
  source and copy before the delete.
- **Errors are per album and swallowed.** The loop body catches everything and logs (`LibraryService.cs:645-648`), and
  the method returns `Data = movedCount > 0` (`:651-654`). A run where every album failed returns a result with
  `Data = false` and no error.
- **The merge** (`ServiceBase.cs:56-227`) compares incoming and existing songs by track number, drops an incoming song
  when its CRC32 matches an existing one or when "the existing song is better" (larger file, or higher bit rate and bit
  depth) (`:177-200`), moves what remains, keeps the higher-resolution image when names collide (`:114-141`), and then
  deletes the incoming directory outright (`:216-222`). The loser of each comparison is discarded with no record. A
  partial album can be admitted this way.
- **Concurrency.** Up to four albums move in parallel (`LibraryService.cs:501-507`).
- **Tests.** The 17 references to the move method in `tests/Melodee.Tests.Common/Services/LibraryServiceTests.cs:934-1128`
  cover the guards (same library, invalid, locked, not a storage library) and two zero-album cases. A search of `tests/`
  for the merge helper's name and for `AlbumsMergedWithExisting` finds only `ScanStepResultTests.cs`. No test found moves
  a real album or exercises the merge. This is a grep, not a run.

### Stage 4: storage to database (`LibraryInsertJob`, daily at midnight)

Source: `M:Jobs/LibraryInsertJob.cs`; schedule `M:Data/MelodeeDbContext.cs:1168-1171` (`0 0 0 * * ?`).

- **Discovery.** For each unlocked storage library, find `melodee.json` files modified since the last scan, or all in
  force mode (`:119-260`, `:205-222`).
- **Insert.** For each album it resolves the artist row; an album whose artist cannot be found is skipped with a
  warning and stays in storage unindexed (`:439-457`). It looks for an existing album by artist, normalized title, and
  ids; an existing one is counted and skipped (`:469-478`).
- **Duplicates are renamed on disk.** A second album with the same artist and normalized title, or the same
  MusicBrainz or Spotify id, within one batch gets its directory prefixed (`:511-519`). The prefix default is
  `_duplicate_ ` in the seeded setting (`Data/MelodeeDbContext.cs:649-653`), `__duplicate_ ` in the job's fallback
  (`LibraryInsertJob.cs:166-167`), and `__duplicate_` in the docs (F10).
- **Songs.** A missing media file deletes the sidecar and sends a reprocess message that re-runs the staging service
  over that directory (`:540-556`; `M:MessageBus/EventHandlers/MelodeeAlbumReprocessEventHandler.cs:13-36`). Each
  present file is re-read in full to compute CRC32 as its stored file hash (`:558`, `:571`).
- **Commit.** Albums go in with `AddRangeAsync` and one `SaveChangesAsync`; a unique-constraint violation reloads and
  retries only the albums not already present (`:602-642`); any other exception is logged and swallowed (`:643-646`).
  The counter of inserted albums is then increased by the whole batch size whether or not the save succeeded (`:648`).
- **Revalidation of the stuck.** A weekly job re-asks the search engine for staged albums whose reasons include an
  invalid or unknown artist (`M:Jobs/StagingAlbumRevalidationJob.cs:76-95`, `:155-158`, docs: `jobs.md`,
  "StagingAlbumRevalidationJob"), gated by a persisted backoff table (below).

### The chain, as the docs describe it and as the source shows it

Docs: "Drop files in inbound, music is playable within ~20 minutes" through a chain of inbound, auto-move, insert
(`docs/pages/jobs.md`, "Job Chaining"; `libraries.md`, "Automated Job Chain"). Source: see F9.

## Findings

Each finding is a numbered statement with its channel. Lessons that follow from it are in the next section.

- **F1.** State lives in files: the sidecar carries status and reasons, a directory name carries the duplicate verdict,
  and the database learns of an album last (source: `DirectoryProcessorToStagingService.cs:1253-1260`,
  `LibraryService.cs:1081`, `LibraryInsertJob.cs:511-519`).
- **F2.** The verdict is completeness of tags and a weak artist-known test, never a match confidence (source,
  Stage 2).
- **F3.** The unit of validation is the album, as in dubplate, but the unit of admission is not: the merge can admit
  part of an album and discard the rest (source, `ServiceBase.cs:177-222`).
- **F4.** The commit step is copy-with-overwrite then delete, without verification, with per-album error swallowing
  and a success flag that means "at least one moved" (source, `FileSystemDirectoryInfoExtensions.cs:556-598`,
  `LibraryService.cs:645-654`).
- **F5.** Validation fails open on a thrown check and may share mutable state across parallel albums (source,
  Stage 2; static reading).
- **F6.** Multi-disc albums are handled by collapsing the disc structure and rewriting the file's tags, on by default
  (source, `MediaEditService.cs:734-760`; seed `Data/MelodeeDbContext.cs:315-318`).
- **F7.** Conversion to 384 kbps MP3 is on by default and its output is accepted on a 16-byte sniff (source, Stage 1).
- **F8.** Enrichment (artist search, image download) runs inside the staging pass, takes the top-ranked hit with no
  threshold, and sets `Ok` before the validator runs (source, `:1101`, `:1149`). The image download wraps a retry
  pipeline (4 attempts, 3-minute exponential delay, `M:Extensions/HttpClientExtensions.cs:109-128`) around an action
  that catches every exception and returns `false` (`:200-203`). A retry strategy that handles exceptions never sees a
  `false` return; that Polly's default handles exceptions only is recalled from its documentation and is
  *unconfirmed here*, so the point is a low-confidence quality note. The download has no size cap in the code read
  (`:130-215`), and `ISsrfValidator` is referenced only by the podcast code and the client factory extension, while
  staging gets a plain `httpClientFactory.CreateClient()` (`DirectoryProcessorToStagingService.cs:861`); whether that
  default client is guarded elsewhere was not traced.
- **F9.** The scheduled chain stops after the auto-move job. `ScheduleJobIfConfigured` builds cron triggers with no
  `ChainOnComplete` entry (`src/Melodee.Blazor/Extensions/QuartzSchedulerExtensions.cs:27-43`). The inbound job forwards
  `ChainOnComplete` only when its own data map has it (`LibraryInboundProcessJob.cs:189-196`), and only the upload
  dialog sets it (`UploadAlbumDialog.razor:280-286`). The next job is started with `scheduler.TriggerJob`, which the
  auto-move job then sees as a manual trigger (`context.Trigger is not ICronTrigger`, `StagingAutoMoveJob.cs:153-156`),
  so it chains onward only if `ChainOnComplete` is set (`:158-167`), which it is not. The insert job then waits for its
  own midnight cron. Outcome: **partly answered**. The premise that `TriggerJob` yields a non-cron trigger is
  consistent with the jobs' own manual-trigger logic and with Quartz.NET's description of `TriggerJob` as manual
  triggering (https://www.quartz-scheduler.net/documentation/quartz-4.x/tutorial/jobs-and-triggers.html), but the
  trigger class was not read in Quartz's source. Confirm by reading `QuartzScheduler.TriggerJob` in the Quartz.NET
  repository. No harness test applies (Melodee is not run here); the finding feeds a design note only, so nothing waits
  on it.
- **F10.** Documentation drift: the duplicate prefix has three spellings (Stage 4 above and `docs/pages/jobs.md`,
  "LibraryInsertJob", "Special Handling"); `JobRegistry` describes the insert job as "Processes staged albums and moves
  to storage" (`M:Constants/JobRegistry.cs:27-29`) though the job reads storage and inserts rows. Docs are not a
  reliable source for Melodee behavior; this record cites source wherever the two differ.
- **F11.** The parallel-validator and counter points are source readings. Counters incremented from the parallel loop
  (`numberOfAlbumsProcessed += ...`, `DirectoryProcessorToStagingService.cs:418-419`) are plain integers captured by the
  lambda and not atomic. Neither was run, and neither is needed to draw the lesson.
- **F12.** The staging service holds a `_stopProcessingTriggered` flag that is set when the maximum count is reached
  and never assigned `false` anywhere in the file (`:108`, `:1307`, grep over the file shows one assignment). The
  service is registered `Scoped` (`src/Melodee.Blazor/Program.cs:572`), so how far the flag leaks depends on scope
  lifetime in the job runner; unconfirmed.
- **F13.** (Observation.) Scheduled inbound runs depend on the root directory's mtime (Stage 1). On this workstation,
  Linux 7.1.10-200.fc44 on tmpfs, 2026-10-07, in a scratch directory outside the repo: creating a file inside an
  existing subfolder left the root directory's mtime unchanged (`00:57:30.30`), while creating a new top-level folder
  advanced it (`00:57:32.71`). So a track added to an existing top-level folder in inbound does not make the library
  "need scanning"; a new top-level folder, which the upload page always creates, does. Filesystem type: tmpfs, so the
  result is one filesystem's behavior, the usual POSIX one.
- **F14.** Melodee's Requests feature auto-completes a request when a matching album arrives, matching by exact
  normalized artist and title and an optional year (`M:Services/RequestAutoCompletionService.cs:28-73`), and posts a
  system comment (`:85-102`). The only caller of its hook is `AlbumAddEventHandler` (`M:MessageBus/EventHandlers/AlbumAddEventHandler.cs:375`),
  and a grep over `src/` for `AlbumAddEvent` finds one publisher: the artist-rescan handler
  (`ArtistRescanEventHandler.cs:183`). `LibraryInsertJob` does not publish it. As far as the source shows, an album
  admitted through the main pipeline does not complete requests, though the docs say it does (`docs/pages/requests.md`,
  "Automatic Completion"). Static reading; not run.
- **F15.** Reports and intake share no queue in Melodee. Staging holds albums that failed validation; Requests holds
  tickets, including "Artist Correction" and "Album Correction" categories (`docs/pages/requests.md`, "Request
  Categories"). They are two systems.

## Comparison with dubplate's door

| Axis | Melodee (v2.2.0) | dubplate's model |
|---|---|---|
| How an album arrives | Files dropped in a folder, or a ZIP unzipped into it (Stage 0) | tus upload with a session id that is the journey's spine (`uploader-design.md:42`) |
| Who is present | Nobody; timers do the work every 10 and 15 minutes, then daily (Stages 1 to 4) | The contributor, held on the page through the check (`:232-250`) |
| First verdict | After staging, by a timer (F2) | Before anything lands, per file as the stream checks run (`:60-66`) |
| Basis of the verdict | Tag completeness plus any top-ranked artist hit (F2) | Matcher confidence; zero false confident matches is the success bar (`organizer-design.md:20-31`) |
| Unit of verdict | Album | Album; one failing track rejects the batch |
| Unit of admission | Album, but merge can admit part (F3) | Never a partial album (`CLAUDE.md`) |
| Reject path | Album stays in staging as `Invalid` with reasons; an admin edits or deletes (docs: `libraries.md`) | Push-back with fix paths in the contributor's hands; timer-path material to an admin queue (`organizer-design.md:651-654`) |
| Reasons | A flag set plus messages at two severities (Stage 2) | `internal/reject` Reasons with copy and a next step (`uploader-design.md:78-81`) |
| Where state lives | Sidecar JSON and directory names, then rows (F1) | The store; catalog row written by admission |
| Commit | Copy, overwrite, delete; errors swallowed (F4) | Not yet designed beyond "atomic" bars; rung 9 owns it |
| Duplicates | Prefix on a directory name; merge by CRC32 and size (F3) | Duplicate refusal on the decoded-audio key (hook 1) |
| Integrity | Header sniff; no decode found (Stage 2) | Decoded-audio key computed at the gate (hook 1), `mewkiz/flac` adopted |
| Formats | Everything to 384 kbps MP3 by default (F7) | FLAC-only allowlist as config; no transcode at intake |
| Who may contribute | Administrator and Editor roles (Stage 0) | Any rostered member |
| Who contributed | Not recorded (Stage 0) | `added_by` and the source answer (hooks 17, 21) |
| Admission to index | A daily timer reads sidecars into rows (Stage 4, F9) | The catalog is the index; no scanner admits music (`streamer-2-0-design.md:178-186`) |
| After admission | Requests hook reachable only from a rescan (F14) | The notify path (`IDEAS.md`, Requests) |

The two models agree on one thing, that the album is what gets judged. They differ on who is present, what the verdict
means, and what the commit guarantees, and dubplate's answers to all three are the ones its custody bar needs.

## Lessons for dubplate, each with its rung

Rung numbers are the ladder's (7 intake, 9 pipeline and admission, 13 decision and review queues). "Test owed" names a
test the receiving rung plan should carry; none is slice 2's, because slice 2 builds the OpenSubsonic contract harness.

**L1 (positive, rung 9).** *Check that a source file has stopped changing before reading it.* Melodee's two-snapshot
test (Stage 1, `:1378-1409`) is the right shape for a filesystem inbox, which has no "upload finished" signal. tus
has one, so rung 7's uploader needs nothing from it. The timer-path side door the organizer spec keeps
(`organizer-design.md:651`) and rung 16's folder import read directories with no such signal, so rung 9's plan should
name a stable-file rule for them. Take it with a fingerprint (size, mtime, and a short re-check) and a named
"deferred, still changing" outcome, as Melodee records one (`:587-589`). Test owed (rung 9): a file that grows between
two reads is deferred, a quiet file is taken.

**L2 (negative, rungs 7 and 9).** *Tag completeness is a pre-check, never a custody verdict.* Melodee admits on "no
critical message" with a top-ranked artist hit counting as known (F2). dubplate's bar is zero false confident matches,
which a completeness rule cannot give. Keep admission on matcher confidence. The completeness checks are still good
cheap reject reasons for rung 7's gate: no cover, song total not equal to count, track numbers with gaps, zero
durations. Feed them into rung 7's reject-reason set (STATUS "To rung 7" carries that set), each with `internal/reject`
copy and a next step. Test owed (rung 7): each pre-check has a reject-wording case and no path admits on completeness
alone.

**L3 (negative, rungs 7 and 8).** *Disc and track are a pair, and a repair must not rewrite the contributor's tags to
make a check pass.* Melodee's validator keys on track number alone, fails multi-disc albums, then "fixes" them by
renumbering and writing `DiscTotal = 1` into the files, by default (F6). dubplate's gate keys on (disc, track), its
tag writer (rung 8) preserves disc structure, and a number that must change to pass is a reject reason, never a silent
rewrite. Melodee's own comment on the sort-order arithmetic shows the cost: it needed a multiplier to restore the order
the collapse destroyed. Test owed (rung 7): a two-disc album with restarting track numbers is accepted unchanged; a
(disc, track) duplicate is rejected.

**L4 (negative, rung 7).** *A check that cannot run is a rejection, never a pass.* Melodee's validator logs a thrown
check and computes the verdict from what it has (F5). dubplate's gate returns a named reason ("could not check this
file") for any check that errors, in `internal/reject` wording. Test owed (rung 7): a check forced to error rejects
the album; the same fixture without the fault is accepted, so the test cannot pass vacuously.

**L5 (negative, rung 9).** *A verdict is a value computed per album, never state on a shared checker.* Melodee's
validator holds results in fields and one instance serves parallel albums (F5, F11). The Go form is a pure function
from album to verdict, which the repo's `go test -race` leg can hold. This also bears on rung 9's concurrency-model
decision (a per-album lease or an explicit queue): whichever it picks, the checker carries no per-run state. Test owed
(rung 9): concurrent verdicts over distinct albums equal their sequential verdicts under `-race`.

**L6 (negative, rung 9).** *The commit is atomic, verified, and recoverable, and reports its real outcome.* Melodee's
move copies with overwrite, deletes the source, returns silently on a long path, swallows per-album errors, and reports
success when any album moved (F4). Its own summary explains the reason for copy-then-delete (a cross-mount rename
fails), which is a real constraint dubplate's plan must answer, not dodge: when staging and the collection root share a
device, rename; otherwise copy, fsync, verify the hash that hook 1 already computed, then rename into place, with the
move-intent journal (hook 15) recording the intent so a crash leaves one complete album. Each album's outcome is
reported, and a batch with any failure returns an error. `internal/atomicfile` is the existing convention to extend.
Test owed (rung 9): a crash injected between copy and delete leaves exactly one complete album and a resumable
journal row; a failed album makes the batch report failure.

**L7 (negative, rungs 7 and 9).** *Never merge into an existing album.* Melodee's merge admits some tracks, drops
others by file size, and deletes the rest of the incoming folder with no record (F3). dubplate refuses duplicates on
the audio key and treats album sameness as the ordered set of track audio keys (hook 1), and a better copy of an
existing album is a decision (rung 13 queue, or a refusal with the reason), never an automatic overwrite. Test owed
(rung 9): an incoming album sharing some audio keys with an admitted one is refused whole, and no existing file
changes.

**L8 (negative, rung 13).** *A verdict about an album lives on a record, never in a directory name.* Melodee renames a
duplicate's folder with a prefix (F1) and then every later query must know to skip that prefix
(`LibraryService.cs:1110-1117`). dubplate's decision records (rung 13) carry the verdict and its reason, and hook 16
lets a decision item name an admitted entity.

**L9 (positive, rung 7).** *Two severities and a closed reason set.* Melodee's `Undesired` versus `Critical`, with each
finding naming a reason from a closed set, is a clean split between "tell the contributor" and "stop the album"
(Stage 2). It matches the reject-wording contract's insistence on no raw-text fallthrough (`uploader-design.md:78-81`).
Rung 7's reject-reason set should state, per reason, whether it blocks. Melodee's 22 flags are a seed list to read, not
a list to copy; dubplate's set follows from its own checks.

**L10 (positive, rungs 9 and 13).** *Spend a lookup only when it is due, back off by attempt, and reset when the
content changes.* The revalidation state store keeps, per album, a fingerprint, an attempt count, and a next-attempt
time, and gives delays of 6 hours, 12 hours, 1 day, 3 days, then 7 days (`M:Services/Scanning/StagingAlbumRevalidationStateStore.cs:59-71`);
a changed fingerprint resets the count (`:263-267`) and a success deletes the row (`:277-286`). A guard,
`CanAttemptArtistRevalidation`, skips albums that cannot benefit (`StagingAlbumRevalidationJob.cs:76-95`). dubplate's
MusicBrainz limiter budgets one request per second (`organizer-design.md:540-546`), so a stalled timer-path album
should not re-spend lookups on every pass. Rung 13's `defer` verb is the natural owner: a deferred decision's re-ask
follows a documented schedule and resets when the album changes. Rung 9 owns the timer path's use of it. Take the idea
and the schedule shape; the numbers are Melodee's and rung 13's plan should state its own with a reason. Test owed
(rung 13): the next-attempt time follows the schedule, a changed album resets it, a resolved album is removed.

**L11 (negative, rung 9).** *Enrichment inside admission needs a threshold, a bound, and a place that cannot stall the
pass.* Melodee takes the top hit with no confidence, sets `Ok` before validating, and downloads whatever URL a search
returned with no size cap in the code read (F8). dubplate already has a thresholded matcher; the lesson is for any
image or metadata fetch rung 9 adds: a size cap, a timeout inside the pass's lease, a recorded "no art found" outcome,
and no retry layer that is dead because the inner function swallows the error. (Metadata fetching is a separate
question from the refused music-fetching features in `CLAUDE.md`; this record rules nothing on it.)

**L12 (negative, rung 7).** *The verdict belongs at the door.* Melodee's upload reports "Complete" after unzipping and
the contributor never learns the outcome (Stage 0), while the album waits for a timer. dubplate's journey states and
in-page verdict are the correction; this is evidence for keeping them, and for keeping the timer path admin-only, where
a delay costs an admin and not a contributor. Also, Melodee's upload has no uploader record (hook 17) and no source
answer (hook 21), which supports both hooks being at the submission.

**L13 (negative, rung 9).** *Admission and indexing are one step.* Melodee's chain runs through the scheduler, and on a
scheduled run the last hop depends on a midnight timer (F9, partly confirmed). With the catalog as the index,
dubplate's admission writes the catalog row itself, so no timer separates "in storage" from "playable". The journey's
"in your library appears when the link is real" (`uploader-design.md:68-70`) holds only if that is true; rung 9's
acceptance should state it.

**L14 (negative, rung 9).** *The "album admitted" notification hangs off the commit point.* Melodee's request
auto-completion is wired to a rescan path and not to the main insert (F14). The IDEAS "Requests" entry builds on
intake's matching and the notify path; whichever rung builds it should attach to the admission commit, in the same
transaction as the catalog event (hook 9's rule), so no admission path can skip it. This asks for no new hook:
hook 9 already requires an event and its projector delta to commit together, and an "album admitted" kind belongs in
that event table. Task 9 should confirm the event kinds in hook 4 and hook 9's wording include it.

**L15 (negative, rung 9).** *Find new material by content, not by a root directory's mtime.* The scheduled inbound gate
misses a file added to an existing top-level folder (F13). The timer path should walk and compare size and mtime per
file against a ledger, as the spec's reconcile walk already does for drift (`streamer-2-0-design.md:178-186`), and hash
only on suspicion. Test owed (rung 9): a file added inside an existing folder is found by the next scheduled scan.

## What worth reusing, by path

Melodee's C# cannot be reused as code in a Go module, so reuse means the idea or the table.

- **Take (idea): the revalidation schedule and reset rule.** `M:Services/Scanning/StagingAlbumRevalidationStateStore.cs:59-71`
  and `:247-286`. Owner: rung 13's `defer` plan (L10). Twelve lines of schedule and one reset rule.
- **Take (idea): the stable-file snapshot.** `M:Services/Scanning/DirectoryProcessorToStagingService.cs:1378-1409`.
  Owner: rung 9's timer-path plan (L1).
- **Read, do not copy: the validator's check list as a seed for rung 7's reject reasons.** `M:Plugins/Validation/AlbumValidator.cs:35-269`,
  `:396-445`, with disc-aware numbering (L2, L3).
- **Decline: the unwanted-title and featuring regexes** (`AlbumValidator.cs:19-27`, `:527-528`). dubplate's matcher
  folds featuring artists and normalizes titles already, with a golden generator
  (`internal/engine/matcher/testdata/gen_featuring_golden.py`), and a second folding path is the defect the folded sort
  key exists to prevent (`streamer-2-0-design.md`, "Declined and recorded").
- **Decline: the Quartz job chain and the sidecar-as-state design** (F1, F9). They are the architecture dubplate's
  store replaces.

## The prior-art table's Melodee quality cell (spec `:870`)

**Proposed cell text:**

> Ideas sound, code below the house bar. 4,479 test attributes across six test projects and CodeQL, gitleaks, and
> container-scan workflows, but the admission path is thinly tested (its move and merge have no test that moves an
> album). Read at v2.2.0: validation fails open on a thrown check and may share mutable state across parallel albums;
> the commit copies then deletes without verification and swallows per-album errors; a default repair rewrites
> multi-disc tags. The revalidation backoff and the stable-file check are worth taking as ideas.

**Evidence behind each clause.**

| Clause | Evidence | Channel |
|---|---|---|
| 4,479 test attributes, six projects | A grep count of `[Fact]` and `[Theory]` over `tests/` at the pin (not a run); projects `Melodee.IntegrationTests`, `Melodee.Mql.Tests`, `Melodee.Tests.Blazor`, `Melodee.Tests.Cli`, `Melodee.Tests.Common`, `Melodee.Tests.OpenSubsonic` | source |
| CodeQL, gitleaks, container scan | `.github/workflows/codeql.yml`, `gitleaks.yml`, `sca-container-scan.yml` (also `dotnet.yml`, `docker-publish.yml`, `localization.yml`) | source |
| Move and merge untested | 17 references to the move method in `LibraryServiceTests.cs:934-1128`, all guards or zero-album cases; no test names the merge helper | source (grep) |
| Fail-open validation | `AlbumValidator.cs:109-131` | source, static |
| Shared mutable validator | `AlbumValidator.cs:31-33`, `:45-46`; `DirectoryProcessorToStagingService.cs:152`, `:389-447`, `:1248` | source, static |
| Unverified copy-then-delete | `FileSystemDirectoryInfoExtensions.cs:556-598` | source |
| Swallowed per-album errors | `LibraryService.cs:645-654` | source |
| Multi-disc rewrite by default | `MediaEditService.cs:734-760`; seed `Data/MelodeeDbContext.cs:315-318` | source |

**Two authorship facts, because the table grades quality and maintenance.** The repository carries an `AGENTS.md` for
AI agents and `kilo.json` at its root. Since 2026-01-01 the commit log at the pin shows 446 commits by the maintainer
(Steven Hildreth), 52 by `copilot-swe-agent[bot]`, and 1 by another author (499 in all; `git log --since=2026-01-01
--format=%an`). Total commits at the pin: 1,474 (`git rev-list --count HEAD`); the table's "1,478 commits since
2024-10" is Task 6's cell to re-date, not this record's. Neither fact is a quality verdict. Both tell the reader that
the code was written with agents in the loop and is graded on what it does.

**Verdict for the table's Verdict column.** Unchanged: adopt as readable prior art, with the added clause that the
code is read for ideas and its commit step is a negative example. The cell's text is for Task 9 to place; this record
changes no verdict.

## What this record could not do

- Run Melodee. No oracle exists for it in this study and the plan forbids building code from a clone, so every
  runtime claim (F5, F9's trigger class, F11, F12, F14) is a reading of source and marked as one.
- Read Quartz.NET's source for `TriggerJob`'s trigger class (F9). Named above with what would confirm it.
- Confirm Polly's default retry predicate (F8). Recalled from its documentation, not read; marked unconfirmed.
- Trace whether the staging service's `HttpClient` is guarded by the SSRF validator (F8).

## Prior art consulted

Melodee (MIT), adopted as readable and reusable prior art (spec `:870`), read at `v2.2.0`. Quartz.NET's tutorial page
for one citation (F9). Nothing else adopted or declined anew, except the two declines named under "What worth
reusing". No GPL source was read in this task, and no copyleft text appears here.

**Dependency sweep and operator delta.** This record edits no dependency manifest, so the pass-opening sweep does not
apply (plan header). Operator-supplied values: delta zero.

## Routed

- **Rung 7 plan** (intake): L2, L3, L4, L9, L12. Reject-reason set candidates from Melodee's check list, disc-aware
  numbering, errors reject, per-reason blocking flag, verdict at the door. Tests owed as listed under each lesson.
- **Rung 8 plan** (tag I/O): L3. The writer preserves disc structure; no repair rewrites disc or track tags.
- **Rung 9 plan** (admission and the timer path): L1, L5, L6, L7, L11, L13, L14, L15. Atomic and verified commit with
  the journal, no merge, a stable-file rule, content-walk discovery, enrichment bounds, admission and indexing as one
  step, a concurrency model with a stateless checker. Tests owed as listed.
- **Rung 13 plan** (decision queue): L8, L10. Verdicts on records, the `defer` re-ask schedule with a reset rule.
- **Task 9 (spec revision 3.1).** (a) The Melodee quality cell text above. (b) No new forward hook. L14 asks Task 9 to
  confirm that hook 4's and hook 9's event kinds include an album-admitted event; if the wording does not, a one-line
  clarification is the fix, not a new hook. (c) Hook 1's decision to compute the audio key at the gate is supported:
  Melodee computes only CRC32, after the move, at database insert (`LibraryInsertJob.cs:558`), and has no decode step.
  (d) Hooks 17 and 21 are supported: Melodee records no uploader.
- **Slice 7 plan** (open item: a report path). Melodee's tickets for "Artist Correction" and "Album Correction"
  (`docs/pages/requests.md`) are a separate system from staging (F15). The spec's candidate, a reserved playlist
  whose adds file a report into the decision queue (hook 16), avoids a second queue; F15 is evidence for that choice.
- **IDEAS, "Requests" entry** (not a rung). Melodee is prior art for "intake closes the request when a match arrives":
  exact normalized artist and title with an optional year, a system comment on completion (F14). Its reach into the
  main pipeline is the defect to avoid (L14). Melodee's tracker accepts "Add Album" requests and nothing there fetches
  music, which stays consistent with the refused-features list; any dubplate version stays in the web app only, per
  the IDEAS entry.
- **Task 7 (synthesis).** Nothing flips a verdict. F9's one unconfirmed premise changes no design note.
- **Not routed because out of scope for a study:** Melodee's Jellyfin and OpenSubsonic layers, jukebox, party mode, and
  podcast code were not read; the survey (`2026-09-25-streamer-prior-art-survey.md:115`) already covers its streaming
  limiter and OpenSubsonic matrix.
