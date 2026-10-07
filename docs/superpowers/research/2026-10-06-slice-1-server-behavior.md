# Slice 1, Task 4: where the specification is silent, what Navidrome and gonic do

Date read: 2026-10-07 (UTC, 04:50 to 05:20). Staged outside the repo; lands at A-land as
`docs/superpowers/research/2026-10-06-slice-1-server-behavior.md`. This is the spec's "written record"
(`2026-09-07-streamer-2-0-design.md:724-726`) in its server half, and the citation source for the protocol layer's
clean-room clause (`:316-318`).

**Pins read** (from `2026-10-06-slice-1-evidence-pins.md`, none changed):

- Navidrome `0.64.2`, image `docker.io/deluan/navidrome@sha256:38dc2727bfcfd5ede290f8ada114fc90368146f265ae4701ddddbcbe2a44ee52`;
  source read at tag `v0.64.2` = `navidrome@101145742f4762164202cb4858c9126258bdc463` (shallow clone, this task).
- gonic `v0.22.0`, image `docker.io/sentriz/gonic@sha256:516fd9645614ba3a596d86174216c3e944808b9ec970c581678713be4c8b1d49`;
  source read at `gonic@5ae99e8af550e4d5cbd45a3b430419d32cb54aed`.
- `navidrome/website@b0fe572513288225562fc4a71d06e01333f03602` (Apache-2.0, quotable).
- `opensubsonic/open-subsonic-api@bed1688a7b0afe9f528a3dc1400a46ab7961b280` (Apache-2.0, quotable): 87 paths in
  `openapi/openapi.json`, which is the endpoint universe used in section 1.

**Channels.** Every claim names its channel. **[D]** documentation, **[S]** source at the pinned commit (cited
`repo@commit:path:line`, stated in dubplate's own words), **[O]** observation of a local `podman` container on
`127.0.0.1`. Source citations name `navidrome@v0.64.2` as `N:` and `gonic@v0.22.0` as `G:`; paths are relative to each
repository root.

**Fixture states** (planning-miss check 4). Each observation names one.

- **F1.** The Task 0 fixture, mounted read-only: 5 tracks, 4 albums, 2 artists, one library, one 45-minute
  `Endurance` track. Scanned (`getScanStatus` reported `scanning: false`, `count: 5`), logged in with token
  authentication (`md5(password + salt)`). Throwaway credentials, deleted with the data directories.
- **F2.** A second scratch fixture, built with `ffmpeg` from synthetic 3-second tones, to show real paths against tags:
  `Loose Track.flac` at the root, `Misc/Mixtape/` holding two tracks of one album with two different track artists,
  and `Deep/One/Two/Three/` holding one track. Scanned, logged in. For two observations Navidrome ran a second
  library (`/music/Misc`, created through its admin API) nested inside the first; that state is named where used.
- **Transcode cache states.** *Cold* (no entry for the request's key), *filling* (the first request's encode still
  running, a few seconds in), *complete* (the encode finished). The `Endurance` track took about 99 s to encode on
  Navidrome and about 70 s on gonic while another lane loaded the machine, so every filling observation sat inside the
  window. The cache directory's file size was read at each probe and is quoted where it proves the state.

Instances: Navidrome on `127.0.0.1:47140` (F1) and `:47142` (F2), gonic on `:47141` (F1) and `:47143` (F2), each with its
own data directory under this task's work directory, started from Task 0's run recipe. All four containers were stopped
and removed; `ss` showed no 4714x listener afterward. The box was never contacted. The retirement archive was never
opened. Transcodes ran one at a time.

**Clean-room rule 1 holds.** No GPL text, code, pseudo-code, or identifier list appears here. Navidrome's and gonic's
endpoint registrations, header values, and decision rules are described in prose and cited by line. Quoted text comes
only from `navidrome/website` and the OpenSubsonic documentation (Apache-2.0). Wire values observed from the running
oracles (status codes, header names, response shapes) are observations, not source. The one place an oracle's
configuration data would match its source (Navidrome's built-in transcoding command templates) is described in words and
not reproduced.

## Questions owned, and their outcome states

| # | Question (plan Task 4) | State |
|---|---|---|
| 1 | Endpoints neither oracle implements, and the empty-success set (server side) | Answered (section 1). The client-side filter that makes the final set is Task 2 and Task 7. |
| 2 | Which endpoints honor `musicFolderId` | Answered (section 2); multi-folder semantics observed on Navidrome, test owed for gonic's second folder |
| 3 | Range on a transcode: complete versus filling, `Accept-Ranges`, `estimateContentLength`, `timeOffset` | Answered (section 3) |
| 4 | How Navidrome implements `transcoding`, and how a posted `ClientInfo` meets its configuration | Answered (section 4) |
| 5 | CORS on a simple request and a preflight; whether credentials cross | Answered (section 5) |
| 6 | `search3` empty query, short query, paging over the whole library | Answered (section 6) |
| 7 | Play queue edges and `scrobble` `submission` against `:563-569` | Answered (section 7) |
| 8 | The folder view: Navidrome's synthesized tree against gonic's real paths | Answered (section 8); one gonic scan result is partly answered |
| 9 | The `.nsp` format | Answered (section 9); whether the group holds any `.nsp` file is routed to slice 8 |
| 10 | The area list of Navidrome's e2e suite and gonic's golden fixtures | Answered (section 10) |

The "ideas and lessons" half is each section's **Lesson** paragraph, written as dubplate design notes, and each section
ends with the slice it feeds.

---

## 1. Endpoints neither oracle implements, and the empty-success set

**Method.** One authenticated request per path in the specification's `openapi.json` (87 paths), with no parameters,
plus two unknown names (`fooBar`, `getPodcastz`), on both oracles at F1, then valid-parameter requests for every path
whose no-parameter answer hid its behavior. JSON requested (`f=json`). [O]

### 1.1 How each answers an endpoint it does not implement

| Oracle | What an absent or unimplemented endpoint receives | Source |
|---|---|---|
| Navidrome 0.64.2 | A **plain-text HTTP error outside the `subsonic-response` envelope**, in four classes. Four user-management paths (`createUser`, `updateUser`, `deleteUser`, `changePassword`) and eight others (the seven podcast endpoints, `jukeboxControl` while the jukebox is off) answer **HTTP 501**, "not implemented, but may be in future releases". Six paths (`search`, `getChatMessages`, `addChatMessage`, `getVideos`, `getVideoInfo`, `getCaptions`) answer **HTTP 410**, "will not be implemented", and so does `hls`. A name never registered (`fooBar`, `getPodcastEpisode`, `tokenInfo`, `getPodcastz`) gets the router's own **HTTP 404**. The two sonic-similarity paths answer a bare 404 with an empty body while the plugin is absent. | [S] `N:server/subsonic/api.go:228-236` (the 501 and 410 lists), `:270-290` (the two helpers); `:216-226` (sharing and jukebox flip to 501 when their feature is disabled). Sharing was enabled here, so those four answered normally. |
| gonic v0.22.0 | **HTTP 200 and a well-formed envelope** with `status: failed`, **error 70, "view not found"**, for every absent name, on every format. | [S] `G:server/ctrlsubsonic/ctrl.go:183` (one catch-all route answers an unknown path) |

The same plain-text answers came back with `f=xml`, so the envelope absence is not a JSON artifact. [O]

Counting the 87 specification paths at F1: Navidrome answers 62 normally, plus `getTranscodeDecision` (POST only; GET
returned 405) and `getTranscodeStream` (a request without `transcodeParams` returned 400) as plain-text statuses when
called wrongly, plus the two plugin-gated sonic paths. gonic answers 60 normally and 27 with error 70. The ten paths
Navidrome answers and gonic does not: `createShare`, `deleteShare`, `updateShare`, `getShares`, `getNowPlaying`,
`getPlayQueueByIndex`, `savePlayQueueByIndex`, `reportPlayback`, `getAlbumInfo`, `getUsers`. The eight gonic answers
where Navidrome returns 501 are the seven podcast endpoints and `jukeboxControl` (gonic answers error 0, "jukebox not
enabled"). [O]

The specification's own rule is neither: "Every other endpoint returns error code 0 inside a well-formed
`subsonic-response` with a message naming the endpoint, never a bare 404" (`2026-09-07-streamer-2-0-design.md:670`).
Navidrome sends no envelope at all, and gonic sends the envelope with error 70 (`70` is the specification's "data not
found" code, which an absent endpoint is not).

### 1.2 Which endpoints answer an empty success

Two different things hide under "empty success", and the oracles agree on both. All of the following are [O] at F1, with
no external integration configured (no Last.fm, no Spotify, no plugin).

**An implemented endpoint with nothing to return answers `ok` and an empty body**, the normal case, not a stub:
`getStarred`, `getStarred2`, `getGenres` (the fixture has no genre tags), `getBookmarks`, `getInternetRadioStations`,
`getPlaylists`, `getSongsByGenre` on Navidrome, and `getNowPlaying`, `getShares` on Navidrome.

**An endpoint that depends on an external service answers `ok` with an empty body when the service is absent**, never an
error. Navidrome: `getArtistInfo`, `getArtistInfo2`, `getTopSongs`, `getSimilarSongs` (each an empty element),
`getLyricsBySongId` (an empty list), `getLyrics` (an empty `value`). `getSimilarSongs2` returned the artist's other
local songs, so Navidrome falls back to a local answer there. `getAlbumInfo` and `getAlbumInfo2` returned image URLs
built from the local cover and no text. gonic: `getArtistInfo`, `getArtistInfo2` (biography and URLs empty),
`getAlbumInfo2` (empty notes), `getSimilarSongs`, `getSimilarSongs2`, `getTopSongs` (empty elements),
`getLyricsBySongId` (a null list), `getLyrics` (the artist and title echoed, no text). Navidrome's documentation lists
the same dependence: the info, top-songs, and similar-songs endpoints "Require external integrations"
(`navidrome/website@b0fe572:content/en/docs/developers/subsonic-api.md`, the "Notes" column). [D]

**Two real constants.** `getLicense` answers `valid: true` on both. Navidrome's documentation says "Always valid ;)" [D].
`ping` and `getOpenSubsonicExtensions` carry no state.

**A missing required parameter is not an empty success.** Both answer `ok: false` error 10 for the usual endpoints. The
exceptions are in the ledger: Navidrome answers `getSongsByGenre` with no `genre` as an empty `ok` and `getLyrics` with
no title as an empty `ok`; gonic answers `star` and `unstar` with no `id` as `ok` (it changed nothing), and
`getAlbumList2 type=byYear` with no years as `ok` with albums. Navidrome also answers `savePlayQueue` and
`savePlayQueueByIndex` with no parameters as `ok` (a clear, section 7).

### 1.3 Lesson for dubplate

1. Neither oracle is a model for the unimplemented answer. Both give a client a failure it can read (a status or an
   error code), and the spec's choice, a well-formed envelope with error 0, is the safest of the three for clients
   that parse the body before looking at the status. gonic's error 70 reads as "that id was not found", which sends a
   client to the wrong repair.
2. The empty-success set the oracles suggest is the integration-dependent family, not a list of stubs: artist and album
   info, top and similar songs, and lyrics. Whether a roster client hard-fails when those answer an error is Task 2's
   question, and the set Task 7 settles is the intersection of the two.
3. A parameter that is missing must say so (error 10), not succeed. gonic's `star` with no `id` succeeding is the
   behavior to avoid.
4. Declare the family of "off by configuration" endpoints in one place, as Navidrome does with its two helper lists
   and its feature switches: a route table with an explicit unimplemented list is easier to test than a catch-all.

**Feeds** slice 3 (response layer, the unimplemented answer) and Task 7 (the empty-success set).

---

## 2. Which endpoints honor `musicFolderId` (`:374-376`)

**The specification defines the parameter on ten browse and list endpoints** and on `createUser` and `updateUser`
(access control). The ten: `getIndexes`, `getArtists`, `getAlbumList`, `getAlbumList2`, `getRandomSongs`,
`getSongsByGenre`, `getStarred`, `getStarred2`, `search2`, `search3`
(`open-subsonic-api@bed1688:content/en/docs/Endpoints/{getindexes,getartists,getalbumlist,getalbumlist2,getrandomsongs,getsongsbygenre,getstarred,getstarred2,search2,search3}.md`,
the `musicFolderId` row of each parameter table). [D]

**Both oracles honor exactly those ten.** Navidrome reads the parameter in `getIndexes` and `getArtists`
(`N:server/subsonic/browsing.go:100-118`), the album lists and starred lists (`N:server/subsonic/album_lists.go:66,124`),
`getRandomSongs` and `getSongsByGenre` (`:239,:266`), and `search2` and `search3`
(`N:server/subsonic/searching.go:103,140`), through one helper that returns every library the user may see when the
parameter is absent (`N:server/subsonic/helpers.go:541-571`). gonic reads it in `getIndexes`, `getAlbumList`, `search2`,
`getStarred` (`G:server/ctrlsubsonic/handlers_by_folder.go:30,173,214,310`), `getRandomSongs`
(`G:server/ctrlsubsonic/handlers_common.go:340`), and `getArtists`, `getAlbumList2`, `search3`, `getSongsByGenre`,
`getStarred2` (`G:server/ctrlsubsonic/handlers_by_tags.go:32,192,234,467,496`). [S]

**Observed at F1 and F2** with a valid id and an unknown id, on every endpoint above plus six that do not take it
(`getGenres`, `getNowPlaying`, `getMusicDirectory`, `getTopSongs`, `getScanStatus`, `getPlaylists`): the six ignored the
parameter on both oracles, and the ten honored it. [O]

| State | Navidrome | gonic |
|---|---|---|
| Parameter absent | all libraries | all folders |
| Valid id, one folder | same rows as absent | same rows as absent |
| **Unknown id** | **error 70** on every one of the ten. The text is "Library 999 not found or not accessible" on eight, and "Library not found or empty" on `getIndexes` and `getArtists`. | **`ok` with an empty result**, on all ten |
| id semantics | the library's database id (1-based, one row each) | the **index** into the configured music paths (0-based), so an out-of-range index selects nothing |

[O] and [S] (`G:server/ctrlsubsonic/handlers_common.go:782-791`, which turns an out-of-range index into a path that matches
nothing).

With two libraries (F2, a second library nested in the first, so a file inside it belongs to both): `getAlbumList2`,
`getArtists`, `search3`, and `getRandomSongs` returned only that library's rows for each id, and **the nested file
appeared twice under no filter** (6 song rows for 4 files). Navidrome therefore reproduces the spec's
accepted "the same release in two collections is two albums" (`:369-371`). dubplate's no-nested-roots rule
(`CLAUDE.md`, "Facts agents get wrong") means a file can belong to one collection only, so the duplication cannot arise
in dubplate. [O]

**Not observed:** gonic with two music paths; multi-valued `musicFolderId` (Navidrome reads a list in source,
`N:server/subsonic/helpers.go:547`; the specification documents one value for the ten endpoints). A harness fixture with
two collections owes both.

**Lesson.** The honest set is the ten, and it matches the specification, so the spec's "honored per endpoint" is the
specification's own list. Two design notes: an unknown id is a client error (Navidrome's error 70 teaches the client
something; gonic's empty success hides a stale id after a reconfiguration), and gonic's index-as-id means an id changes
meaning when the path list is reordered, which dubplate's surrogate collection ids avoid.
OpenSubsonic PR #162 would extend the parameter to more endpoints; that watch is Task 1's.

**Feeds** slice 3 (browse views), slice 5 (none), Task 7 (the `musicFolderId` list at `:375`).

---

## 3. Range, `Accept-Ranges`, `estimateContentLength`, and `timeOffset` on a transcode

All requests are `stream` for `Endurance` (45 minutes, FLAC, 151,621,505 bytes) with `format=mp3&maxBitRate=128`
unless stated, at F1, with the transcode cache state named per row. Navidrome's complete entry was 43,200,847 bytes and
gonic's 43,200,846 (one byte apart). [O]

### 3.1 The direct path (no transcode)

Both oracles serve the file through the standard library's range-aware file serving (`N:core/stream/media_streamer.go:159-163`
for a seekable stream, `G:server/ctrlsubsonic/handlers_raw.go:301-309` for `format=raw` and for a bitrate cap at or above the source
bitrate). Observed on a short track (F1): no Range gives 200 with `Accept-Ranges: bytes` and the length; `bytes=0-99`
gives 206 with `Content-Range: bytes 0-99/132362`; `bytes=1000-` and the suffix form `bytes=-50` give 206;
`bytes=9999999-` gives 416 with `Content-Range: bytes */132362`; HEAD gives 200 with the length. Navidrome adds
`X-Content-Duration` and `X-Content-Type-Options`; gonic adds neither. Navidrome ignores `timeOffset` on `format=raw`
(206 over the whole file, total 151,621,505). [O]

### 3.2 A transcode, by cache state

| State and request | Navidrome 0.64.2 | gonic v0.22.0 |
|---|---|---|
| **Cold**, first GET, no Range | 200, chunked, **`Accept-Ranges: none`**, no `Content-Length`, `Content-Type: audio/mpeg`, `X-Content-Duration: 2700`. Headers in 0.17 s, then the body streams as ffmpeg writes. | 200, chunked, **no `Accept-Ranges`**, no `Content-Length`. Headers in 0.2 s. |
| **Filling** (4 s in; the entry file held 1.35 MB on Navidrome and 2.1 MB on gonic), second GET, no Range | 200 and the entry **from byte 0**, `Accept-Ranges: none`, chunked | The request **waited 65.8 s** (until the first encode ended), then answered 200 with `Accept-Ranges: bytes` and `Content-Length: 43200846`, serving the **complete** file |
| Filling, GET `Range: bytes=1000000-` | **200**, the whole entry from byte 0. The Range header was ignored. | Waited 65.8 s, then **206**, `Content-Range: bytes 1000000-43200845/43200846` |
| Filling, GET `Range: bytes=0-99` | 200, whole entry, `Accept-Ranges: none` | Waited, then 206, 100 bytes |
| Filling, GET `Range: bytes=0-` | 200, whole entry | (as above) |
| Filling, HEAD | 200, `Accept-Ranges: none`, no length | Waited, then 200 with `Accept-Ranges: bytes` and the final length |
| Filling, GET `estimateContentLength=true` | 200, **`Content-Length: 44236800`** (the actual complete size is 43,200,847) | Waited, then the real complete length |
| **Complete**, GET `Range: bytes=1000000-1000099` | 206, `Content-Range: bytes 1000000-1000099/43200847`, `Accept-Ranges: bytes`, `Last-Modified` present | 206, same shape |
| Complete, GET `Range: bytes=0-` | 206, full length | 206, full length |
| Complete, GET no Range, HEAD, `estimateContentLength=true` | 200, `Content-Length: 43200847` in all three (the estimate is replaced by the true size) | 200, `Content-Length: 43200846` in all three |

State proofs: on Navidrome the cache file grew 1.35 MB, 3.1 MB, 4.9 MB, ... at 4-second polls, and a HEAD poll returned
`Accept-Ranges: none` until the file reached 43,200,847 bytes at +99.4 s, when the same HEAD returned `bytes` and the
length. On gonic the entry file read 2,136,356 B at +4 s and 4,665,009 B at +8 s, one `ffmpeg` process was running at both
instants (`podman top`), and the five queued requests all returned headers at +69.9 s, the moment the first request
finished. [O]

**Why, from source.** Navidrome's cache is an `fscache` over `djherbis/stream` (`N:utils/cache/file_caches.go:151-214`;
`go.mod` pins `djherbis/fscache` and `djherbis/stream v1.5.1`, the version the spec names). A second request for the same
key attaches as a reader to the entry being written and never starts a second encode. The cache hands the response a
seekable view only when the entry's final size is known (`:197-205`); otherwise it hands a plain reader. A plain reader
makes the response non-seekable, and a non-seekable response sets `Accept-Ranges: none`, writes the estimated length only
on request, and copies the stream to the client (`N:core/stream/media_streamer.go:159-198`). The entry key is the media
id, its update time, the bitrate, sample rate, bit depth, channels, format, and **the start offset**
(`:59-61`), so every offset is its own entry. gonic's caching transcoder holds a **per-key mutex for the whole encode**,
writing the encode to the client and to a cache file at once (`G:transcode/transcoder_caching.go:30-71`), and the
route first asks for a finished cache path, which takes the same mutex (`:73-112`; `G:server/ctrlsubsonic/handlers_raw.go:333-343`).
A concurrent request therefore blocks until the first encode ends, then finds a non-empty file and serves it with the
standard library's file serving. [S]

So **neither oracle serves a range from a filling entry**, which is the sentence the spec's "Streaming and transcoding"
writes at `:434-435` ("A Range request on a filling entry is served from that entry once the requested offset is
written"). The spec's own note at `:436-443` already says the oracle diverges. The observations add: Navidrome's answer to
a Range on a filling entry is not an error and not a 206, it is a **200 with the whole entry**, which a client doing a
seek mid-encode reads as "the server restarted my stream at zero"; gonic's answer is a long wait with no headers.

### 3.3 `timeOffset` (the `transcodeOffset` extension)

Both oracles declare `transcodeOffset` v1 and honor `timeOffset` on a transcode, ignore it on `format=raw`. Observed at F1
with `timeOffset=2640` (60 s left): the body was 960,638 B on Navidrome and 960,637 B on gonic, which is 60 s at 128 kbit/s,
so the offset is honored. [O]

| Observation (cold, then complete) | Navidrome | gonic |
|---|---|---|
| Offset entry cached? | **Yes.** The first request had `Accept-Ranges: none`; the second, 0.0 s, had `Accept-Ranges: bytes` and the length; `Range: bytes=0-99` gave 206 and `bytes=5000-` gave 206 | **No.** Every request re-ran ffmpeg (7 to 9 s each under load), none had `Accept-Ranges`, and a Range request answered **200 with the whole body** |
| `estimateContentLength=true`, cold, `timeOffset=2650` | `Content-Length: 44236800`: the estimate **ignores the offset**, and the body was 800,977 B | `Content-Length: 820000`, the offset subtracted (body 800,976 B, an overshoot of 2.4%) |
| Past the end (`timeOffset=9999`) | 200, a 168-byte body | not run |

Both estimates run about 2.4% high against the real encode (Navidrome 44,236,800 against 43,200,847; gonic 984,000 against
960,637). The causes differ. Navidrome's formula multiplies the bitrate in kilobits by 1024 where the encoder's kilobit is 1000,
which is an accident of units, and it has no offset term (`N:core/stream/media_streamer.go:149-151`). gonic's overshoot is
deliberate: its estimate adds 2.5% headroom to bytes per second and subtracts the offset (`G:transcode/transcode.go:82-86`;
`G:server/ctrlsubsonic/handlers_raw.go:348-353`). A client that trusts either estimate sees a body shorter than the declared
length; on gonic Python's HTTP client raised an incomplete-read error, which is what a strict client does. [O] and [S]

### 3.4 When the encode fails or the client leaves

Cold state, a distinct key (`maxBitRate=96`, then `112`), the ffmpeg process killed by `kill -9` inside the container
3 s in, or the client closing after 2 KB. [O]

| Event | Navidrome | gonic |
|---|---|---|
| ffmpeg killed mid-encode | The client received 1.2 MB, then the connection was **aborted** (the client raised an incomplete-read error). The cache entry vanished, and the next request started a fresh encode that grew to 647 KB in 1.5 s. | The client received 1.85 MB and saw a **clean end of stream**, indistinguishable from a complete short track. The partial cache file was removed, and the next request started fresh. |
| Client disconnects after 2 KB | Navidrome's **default continues the encode to completion** and stores the finished entry (the first filling observation: the 2 KB reader left at +0.2 s and the entry finished at +99 s). Its option `Transcoding.EnableCancellation` stops the encode instead. | **ffmpeg stops within a second** and no partial entry persists |

[S] `N:core/stream/media_streamer.go:243-260` (the request context drives ffmpeg only when the option or a concurrency
cap is on; otherwise a detached context does), `N:core/stream/media_streamer.go:187-188` (a failure after bytes were sent
drops the connection rather than ending cleanly), `N:utils/cache/file_caches.go:246-275` (a failed copy cancels the entry so
readers fail with the cause); gonic: `G:server/ctrlsubsonic/handlers_raw.go:355-357` (an ffmpeg-killed or cancelled error is
swallowed and the response flushed), `G:transcode/transcoder_caching.go:65-67` (the cache file is removed on any error).
[D] `navidrome/website@b0fe572:content/en/docs/usage/configuration/options.md:279`: "Disabled by default so that an
in-progress transcode runs to completion even after the client disconnects: the finished file is stored in the
transcoding cache and served instantly the next time it is requested."

### 3.5 Concurrency

Navidrome's limiter never queues: it either reserves a slot or rejects at once, with **HTTP 429**, `Retry-After: 5`, and an
error-0 envelope reading "too many concurrent transcodes, please retry shortly" (`N:core/stream/limiter.go:11-60`;
`N:server/subsonic/api.go:318-323`). The caps are `Transcoding.MaxConcurrent` and `Transcoding.MaxConcurrentPerUser`, both
off (zero) by default, and turning either on also ties ffmpeg's lifetime to the request
(`navidrome/website@b0fe572:.../options.md:280-281`). gonic has no cap [S: a search of the source found none]. The cap was
not exercised in an observation; the harness owes it. [S] and [D]

### 3.6 Lesson for dubplate

1. The spec's fill-while-read promise (a Range on a filling entry served once the offset is written) has no oracle. Both
   oracles reach seekability only on a **complete** entry. A dubplate that serves ranges from a filling entry is a
   declared divergence (the spec's `:436-443` frames this); a dubplate that answers `Accept-Ranges: none` on a filling
   entry and a clean 200-from-zero, as Navidrome does, matches the oracle but loses a client's seek until `timeOffset`
   carries it. The cheaper honest version is Navidrome's: `timeOffset` makes the offset a new cache key, which both
   oracles do, and Range only on complete entries. That is Task 7's decision.
2. gonic's blocking wait is the failure to avoid (a client holds a connection with no bytes for the length of the encode),
   and its **clean end after a killed ffmpeg is the other**: a truncated encode must never read as a finished track. The spec's
   rule (evict, the request returns 500, attached readers see the error) is what Navidrome approaches (it aborts the
   connection and evicts) and gonic does not.
3. An offset belongs in the cache key, and the estimate must subtract it; Navidrome's estimate ignores the offset, and both
   estimates overshoot (gonic by design). If dubplate honors `estimateContentLength`, compute it from the offset and the encoder's real
   bitrate, and decide whether to overshoot or to undershoot, since a body shorter than its declared length reads as a truncated stream.
4. Disconnect handling is a knob in Navidrome with the cache as the reason to keep encoding. The spec's rule (kill only
   when no other reader is attached) is a third answer, the safest of the three for CPU and the friendliest for a second reader.
5. Concurrency: Navidrome rejects with 429 and a retry hint outside the envelope's usual 200; the spec's bounded queue
   then error 0 reads better to clients, and the `Retry-After` header is worth copying.

**Feeds** slice 5 (the streaming slice's Range, offset, estimate, failure, and concurrency rules), and Task 7 (Range on a
filling transcode, `:943-944`). `djherbis/stream` appears here only because Navidrome's filling behavior rests on it
(through `fscache`); its adoption stays slice 5's race test (`:445-447`).

---

## 4. Navidrome's `transcoding` extension, and how a posted `ClientInfo` meets its configuration

The extension appears in Navidrome's `getOpenSubsonicExtensions` as `transcoding` v1; gonic does not declare it and
answers `getTranscodeDecision` and `getTranscodeStream` with error 70. [O] The seven extensions Navidrome declares at
F1 are `transcodeOffset`, `formPost`, `songLyrics` (v1 and v2), `indexBasedQueue`, `transcoding`, `playbackReport`, and
`topSongsByArtistId`. gonic's three are `transcodeOffset`, `formPost`, and `songLyrics` v1. [O]

### 4.1 The two endpoints

**`getTranscodeDecision`** is **POST only** (a GET returned a plain-text 405). It takes `mediaId` and `mediaType` (`song`
only; other values answer error 0, "not yet supported") as query or form parameters and a JSON `ClientInfo` body of at most
1 MiB: client name and platform, a global `maxAudioBitrate` and `maxTranscodingAudioBitrate` (bits per second), direct-play
profiles (containers, codecs, protocols, channel cap), transcoding profiles (container, codec, protocol, channel cap), and
codec profiles with limitations. The protocol, codec profile type, limitation name, and comparison values are validated
against short closed lists; a bad one answers error 0. A bad body answers error 0, "invalid JSON request body". An unknown
`mediaId` answers error 70. [O] [S] (`N:server/subsonic/transcode.go:20-221`, `:238-356`)

The response (F1, an 8-second FLAC) carries `canDirectPlay`, `canTranscode`, a list of reasons ("container 'flac' not
supported by profile [mp3/mp3]", "audio bitrate not supported"), an `errorReason` ("no compatible playback profile found")
when neither is possible, a `transcodeParams` token, a `sourceStream` (container, codec, channels, bitrate in bits per
second, sample rate, bit depth) and, when transcoding, a `transcodeStream` in the same shape. [O]

**`getTranscodeStream`** takes `mediaId`, `mediaType`, `transcodeParams`, and an optional `offset`, and reports **every
error as a plain HTTP status**, never an envelope: 400 for a missing parameter or a media type other than `song`, 404 for an
unknown media id, **410** for a token that is tampered, for another track, or stale. [O] [S] (`N:server/subsonic/transcode.go:374-440`)
The streamed body behaves like any `stream` response: cold and filling `Accept-Ranges: none`, complete `bytes` and 206 on a
Range request (observed on the 8-second track, 128,890 B encoded); a direct-play token streams the original with ranges. [O]

### 4.2 The token

`transcodeParams` is a **signed, stateless token** (a JWT) carrying the decision: the media id, whether to direct-play,
and, if transcoding, the target format, bitrate in kilobits, channels, sample rate, and bit depth, plus the **source file's
update time** and a **48-hour expiry**. The server keeps no decision state. A token whose source time no longer matches
the file's is stale and answers 410, so a retag or re-rip invalidates every outstanding decision
(`N:core/stream/token.go:15-141`). [S] Decoded from an observed token: claims `mid`, `ua`, `exp`, `iss`, and for a
transcode `f`, `b`, `ch`, `sr`. [O]

### 4.3 How a posted `ClientInfo` meets the server's own configuration

Navidrome keeps a **database table of transcoding configurations** (a name, a target format, a command template with
placeholders for the offset, source, and bitrate, and a default bitrate; four ship: MP3, Opus, AAC, and FLAC. The default
AAC template writes ADTS) and a **player** per (client name, user agent, user) that can pin one configuration and a maximum bitrate.
The admin API exposes both (`/api/transcoding`, `/api/player`, session-authenticated). The shipped configurations came from
`GET /api/transcoding`; the command strings are not reproduced here. [O] [S] (`N:core/stream/decider.go:326-344`
resolves a format to its command by database row first, built-in default second.)

The decision engine reads the player. Observed at F1 with a player pinned to MP3 at 96 kbit/s [O]:

1. A client posts a direct-play profile for FLAC only. **The player's 96 kbit/s ceiling outranks the client's profile**:
   direct play is refused (the FLAC's 132 kbit/s is over the cap), the client offered no MP3 transcoding profile to
   fall back on, and the answer is `canTranscode: false`, "no compatible playback profile found".
2. A client posts only an Opus transcoding profile. The server **keeps the client's format and applies the player's
   ceiling**: it transcodes to Opus at 96 kbit/s, not MP3.
3. A client posts `maxAudioBitrate` 64000 and an MP3 profile. The lower of the two wins: 64 kbit/s MP3.
4. A plain `stream` call with no parameters from the same player returned MP3 at 96 kbit/s (96,708 B for 8 seconds). So
   a pinned player transcodes even a client that never negotiates.

Source: when the player forces a format, the server uses it only if the client's profiles can play it; otherwise it falls
back to ordinary negotiation but still applies the player's bitrate ceiling, and a player's own maximum bitrate beats the forced
configuration's default (`N:server/subsonic/transcode.go:283-310`; `N:core/stream/decider.go:157-179` builds the synthetic
profile). [S]

**The legacy `stream` and `download` calls go through the same engine**: a request's `format` and `maxBitRate`, and a
player's cap, are turned into a synthetic `ClientInfo` named "legacy" and resolved by the same decision function, so there
is one decision path rather than two (`N:core/stream/legacy_client.go:14-80`; `N:server/subsonic/stream.go:23-45`). [S]

**Navidrome declines AAC in this extension today.** The decision function deletes every client-offered AAC transcoding
profile, with a source comment that cites a public discussion of AAC reliability with streaming-speaker clients and marks the filter
temporary (`N:server/subsonic/transcode.go:273-281`). An AAC-only client therefore gets "no compatible playback profile found"
(observed). Navidrome's shipped AAC configuration still exists for the legacy `stream` path and writes ADTS. [O] [S]

### 4.4 Lesson for dubplate

1. A decision-then-stream split with a **signed, stateless token bound to the file's update time** is an elegant fit for
   dubplate's retag model: a token minted before a retag expires without a revocation list. The 48-hour life and the
   per-decision tuple are Navidrome's choices.
2. The spec's profile order "request, device, estate default" (`:410-412`) has no slot for a posted capability document. Navidrome's
   answer is the posted document **negotiates, and the server's per-device setting caps and may force**: a device
   configuration is a ceiling on bitrate, a preferred format only if the client can play it, and a catch-all for a client
   that posts nothing. That ordering is a candidate for Task 7 to adopt as dubplate's reading of how `ClientInfo` and the
   device profile interact.
3. One decision path for the legacy parameters and the extension is the maintainable shape; two paths drift.
4. Every error on the streaming half is a plain HTTP status and every error on the decision half is an envelope. If dubplate
   answers the extension, the same split is the natural reading of the extension's contract.
5. AAC through the extension is the part Navidrome has not made reliable, and the reason given is a client population
   dubplate does not have. The evidence for dubplate's AAC container choice (fragmented MP4 or ADTS) is the roster's, which
   is Task 2's; this record's contribution is that Navidrome's shipped configuration writes ADTS and withholds AAC from
   the extension.

**Whether 0.8 answers `transcoding` is scope and is Geoff's** (the plan names it an owner fork); this section gives
Task 7 what the implementation costs: two endpoints, a closed-list validator, a token, a decision function, and the
device-ceiling rule.

**Feeds** slice 5 (the profile order, the token), Task 7 (the `transcoding` owner fork at `:941-942`, the AAC container).

---

## 5. CORS and credential channels

Request shapes at F1, with the cross-origin `Origin: https://app.example`. [O]

| Case | Navidrome | gonic |
|---|---|---|
| Simple GET `ping` with `Origin` | `Access-Control-Allow-Origin: *`; `Access-Control-Expose-Headers: X-Content-Duration, X-Total-Count, X-Nd-Authorization`; `Vary: Origin` | `Access-Control-Allow-Origin: *`; `Access-Control-Allow-Methods: POST, GET, OPTIONS, PUT, DELETE`; `Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization` |
| Simple GET with **no** `Origin` | no CORS headers | the **same three headers, always** |
| Preflight `OPTIONS` (GET with `X-Foo`) | 200, empty body; `Allow-Origin: *`, `Allow-Methods: GET`, `Allow-Headers: X-Foo` (the requested headers **reflected**), three `Vary` lines | 200, empty body; the fixed three headers |
| Preflight for a JSON POST to `getTranscodeDecision`, `Content-Type` and `Authorization` requested, **no credentials** | 200, empty; `Allow-Methods: POST`, `Allow-Headers: Content-Type, Authorization` | 200, empty; the fixed three headers (gonic has no such endpoint; the preflight does not look at the path) |
| Preflight on a protected GET path, no credentials | 200, empty body, no data | 200, empty body, no data |
| **`Access-Control-Allow-Credentials`** | **never sent** | **never sent** |

Both policies are the same in effect: **any origin, no credentialed requests**. The browser will therefore refuse to
expose a response to a script that fetched with cookies or with `credentials: 'include'`, which both servers can afford
because they authenticate by request parameter, not by cookie. [S] Navidrome: a CORS middleware on the whole server with any
origin, any request header, six methods, credentials off, and three exposed headers (`N:server/middlewares.go:90-104`,
installed at `N:server/server.go:173`). gonic: a wrapper that stamps every response with the fixed values and answers
`OPTIONS` itself (`G:handlerutil/handlerutil.go:38-55`, installed at `G:cmd/gonic/gonic.go:292`).

**Preflights never carry data**, and neither oracle's preflight needed credentials. **The unauthenticated carve-out
differs.** Navidrome answers `getOpenSubsonicExtensions` with no credentials (HTTP 200, `ok`, the extension list), and
every other endpoint with error 10 for the first missing parameter (`u`, then `v`, then `c`) (`N:server/subsonic/api.go:97-101`;
the public route sits before the middleware that requires parameters and authentication). **gonic requires credentials for
`getOpenSubsonicExtensions` too** (error 10, "please provide a "u" parameter"; `G:server/ctrlsubsonic/ctrl.go:111` puts it in the
same chain as every route). [O] [S] The spec's carve-out (`:648`) matches Navidrome, not gonic.

### 5.1 Every channel that carries credentials (planning-miss check 3)

Observed on both oracles with a `ping`, F1. [O]

| Channel | Navidrome | gonic |
|---|---|---|
| Query `u`, `t`, `s` (token and salt) | works | works |
| Query `u`, `p` plain | works | works |
| Query `u`, `p=enc:<hex>` | works | works |
| POST form body (`application/x-www-form-urlencoded`) | works, including on a path whose query is empty (the `formPost` extension; the router registers a form-to-query step first, `N:server/subsonic/api.go:94`) | works (both declare `formPost`) |
| `Authorization: Basic` header | **not accepted** (error 10, missing `u`) | not accepted (error 10) |
| `Authorization: Bearer` header | not accepted | not accepted |
| `apiKey` parameter | not accepted (neither declares `apiKeyAuth`) | not accepted |

A wrong password answers error 40 on both (Navidrome: "Wrong username or password"; gonic: "invalid password", and
"invalid username" for an unknown name, which tells a client which half was wrong). Navidrome requires `v` and `c`; gonic
requires `c` only. [O]

### 5.2 Lesson for dubplate

1. A policy of `*` with no credentials suits an API whose credentials ride in the request. dubplate's own channels, once
   app passwords exist, are query parameters and the POST form body (the `formPost` extension), and the `apiKey` path
   (`apiKeyAuth`, which neither oracle declares) is the one with no oracle at all, as the spec already says. If an
   `Authorization` header ever carries an API key, the preflight must list `Authorization` in the allowed headers (the
   oracles differ on how: Navidrome reflects, gonic fixes the list).
2. Navidrome's reflected `Allow-Headers` and its `Vary` lines are the more careful form; gonic's static stamp on every
   response, even without an `Origin`, is simpler and cacheable but sends headers a same-origin request never needs.
3. The two servers' error text for an unknown user (gonic) against a single "wrong username or password" (Navidrome) is a
   small enumeration risk worth not copying.
4. Preflight carries no data and needs no credentials on both; the spec's rule that `getOpenSubsonicExtensions` stays the one
   data-bearing unauthenticated endpoint holds on Navidrome only.

The spec's open item (`:945`: which origins, which headers, whether credentials are ever allowed) is measured by the roster's
browser use, which is Task 2's; this record gives the two oracles' behavior for Task 7 to measure it against.

**Feeds** slice 3 (the response layer's CORS), Task 7 (the CORS policy).

---

## 6. `search3`: the empty query, short queries, and paging over the whole library

F1 (5 tracks, 4 albums, 2 artists), no `musicFolderId`. [O] unless marked.

### 6.1 The empty query

Both return **everything**: all 2 artists, 4 albums, 5 songs, within the 20-per-kind default. `query=""` (the two-quote
literal, which Subsonic clients send) and `query=` (empty) are equivalent on Navidrome; **on gonic the two-quote
literal also returned three extra album rows with empty titles** (7 albums for 4 real ones; its albums table apparently holds the
directory rows of its folder tree, cause unconfirmed).
Navidrome's `*` also returned everything (a trailing `*` is stripped); gonic's `*` matched nothing. [O]

**A query parameter that is entirely absent** differs: Navidrome treats it as the two-quote literal and answers
everything (`N:server/subsonic/searching.go:35`); gonic answers error 10 (`G:server/ctrlsubsonic/handlers_by_tags.go:218-222`).

Navidrome's empty query returns rows in a stable natural order (a consistent order across offsets: `Gamma, Delta` then
`Endurance, Beta` then `Alpha`), gonic's in scan order (`Alpha, Beta, Gamma, Delta, Endurance`). Both paged
consistently. [O] [S] (`N:persistence/sql_search.go:54-65`, the empty-query branch sorts by a natural column and pages by
row id before joining.)

### 6.2 Short queries

| Query | Navidrome (FTS5, word-prefix) | gonic (substring) |
|---|---|---|
| 1 character (`a`, `b`, `e`) | **empty** for every one | `a`: all 2 artists, 3 of 4 albums, all 5 songs; `b`: 1 artist and `Beta`; `e`: 2 artists, 2 albums, 3 songs |
| 2 characters (`al`, `au`, `bo`, `en`) | word-prefix matches: `al` found `Alpha` only (not `Borealis`, which holds `al` mid-word); `au` found the artist, 2 albums, 3 songs | substring: `al` found `Borealis Probe`, 2 albums, `Alpha` |
| 3+ (`alp`, `Alp`, `alpha`) | `Alpha`; case-folded | `Alpha` |
| Mid-word (`pha`) | **empty** (not a prefix) | `Alpha` |
| Multi-word (`signal one`) | album `Signal One` and its songs `Alpha`, `Beta` (each word matched as a prefix) | album `Signal One` only |
| `sig` | albums `Signal One`, `Signal Two` and the three songs (a song matches on its album's words) | the two albums only |
| Space and `a ` | empty | everything holding `a` (the query is split on whitespace and joined with wildcards) |

[S] Navidrome: a query is trimmed, a trailing `*` stripped, accents folded and lower-cased (`N:server/subsonic/searching.go:67`), and
the search applies these rules in this order of precedence (`N:persistence/sql_search.go`): an empty or two-quote query returns all, a UUID matches
MusicBrainz id fields, **a query shorter than two bytes returns nothing**, and otherwise a strategy by backend setting: CJK text goes to a substring match, everything
else to an FTS5 word-prefix query, with a "legacy" backend option. gonic matches `LIKE '%q%'` on the name and its
unidecoded form, a UUID matches the MusicBrainz id column, and `""` means all
(`G:server/ctrlsubsonic/handlers_by_tags.go:224-249`).

**Neither is the spec's design.** The spec uses trigram FTS with a prefix lookup unioned ahead for queries under three
characters (`:396-405`). Navidrome refuses one-character queries and word-prefix-matches the rest; gonic substring-matches
everything. The spec's prefix lookup for short queries therefore has Navidrome's two-character behavior as its nearest
oracle, and **no oracle returns anything for a one-character query that matches Navidrome's refusal and gonic's flood**: the
harness's search comparison is "a set plus a top-N ordering" (`:287-289`), and for short queries the right assertion is
membership on a golden list written by the harness, not an oracle diff.

### 6.3 Paging over the whole library

`songCount=2` with `songOffset` 0, 2, 4, 6 returned 2, 2, 1, 0 songs on both oracles, and the pages concatenated to all five
songs with no repeat or gap. `albumCount=3` with `albumOffset` 0 and 3 returned 3 then 1. With a count of 0 for every kind
the response held no songs, albums, or artists on both. [O]

Navidrome applies no upper clamp to `search3` counts (`N:server/subsonic/searching.go:36-41`: defaults of 20, no cap).
`getAlbumList2` clamps `size` at 500 (`N:server/subsonic/album_lists.go:73`), yet **`size=0` and `size=-1` return every album on Navidrome** and
**`size=-1` on gonic returns error 0 with a leaked SQL syntax error** (`size=0` returns none). [O] [S]

### 6.4 Lesson for dubplate

1. `search3` with an empty query returning everything (the spec at `:399`) is the oracles' behavior and what offline sync depends
   on; paging works on both when the offsets are applied to a stable order, so dubplate's projection must give the empty
   query a deterministic order (a natural column), not insertion order of an unordered table.
2. A query under three characters has no faithful oracle: Navidrome's two-character prefix is the closest and its
   one-character refusal is a defensible edge (no flood); gonic's substring flood is the negative case.
3. Clamp and validate counts: gonic leaks a SQL error text on a negative size, and Navidrome treats 0 or negative as unlimited
   on albums and as "none" on search. The harness owes one table of edge sizes.
4. An absent `query` is a missing-parameter error in the specification's tables and on gonic; Navidrome's answer-all is a
   leniency to decline.

**Feeds** slice 3 (search), the harness (the search divergence), Task 7 (the ledger rows).

---

## 7. The play queue and `scrobble`, against `:563-569`

F1, on both oracles, with `id` lists drawn from the fixture's songs, `savePlayQueue` then `getPlayQueue`. [O]

### 7.1 `savePlayQueue` and `getPlayQueue`

The specification (`open-subsonic-api@bed1688:content/en/docs/Endpoints/saveplayqueue.md:23-33`): `id` is optional, "Send a
call without any parameters to clear the currently saved queue"; "`current` is required **unless** `id` is empty"; "If
`position` is empty, servers should treat the position as 0." [D]

| Call | Navidrome | gonic |
|---|---|---|
| `id=[a,b] current=b position=1234` | saved; `getPlayQueue`: entries `[a,b]`, `current: b`, `position: 1234`, `username`, `changed` (a timestamp), `changedBy` (the client name `c`) | same fields; `changedBy` the client name |
| `id=[a,b,c]`, **no `current`** | `ok`; `current` becomes the **first entry** | `ok`; `current` stored as the string **"-1"** |
| `id=[a,b] current=c` (**c not in the list**) | `ok`; `current` becomes the first entry | `ok`; `current: c`, stored as given |
| `id=[a,b]`, **no `position`** | `ok`; the response **omits `position`** (zero omitted) | `ok`; omits `position` |
| `id=[a,'bogus']` (an **unparseable** id for gonic, whose ids are typed) | `ok`; the unknown id is **dropped silently** (entries `[a]`) | error 10, "please provide some `id` parameters", queue unchanged |
| `id=[a,tr-999]` (a **well-formed** id with no record) | not run | not run; [S] gonic stores a well-formed track id without looking it up (`G:server/ctrlsubsonic/handlers_common.go:260-288`), so the queue would hold it silently. Test owed (harness) |
| **No `id` at all** | `ok`; the queue is **cleared**; `getPlayQueue` returns the envelope with `username`, `changed` at the zero time, `changedBy` empty, and no entries | **error 10** "please provide some `id` parameters"; the old queue stays |
| No `id`, `current=a` | `ok`; clears | error 10 |
| `position=-5` | `ok`; stored and returned as -5 | `ok`; stored as -5 |
| `position=abc` | `ok`; treated as absent | `ok`; absent |
| `getPlayQueue` when none was ever saved | `ok`; a `playQueue` element with `username` and the zero time | `ok`; **no `playQueue` element at all** |

[S] Navidrome's save defaults the current index to the first entry and finds `current` only by matching the id list
(`N:server/subsonic/bookmarks.go:112-150`); gonic's handlers are `G:server/ctrlsubsonic/handlers_common.go:197,260`. The ledger records
that Navidrome also answers a bare `savePlayQueueByIndex` (the `indexBasedQueue` extension) and gonic does not.

**Against the spec's statements (`:563-569`).** The spec's four edges are "a call with no `id` clears the saved queue" (Navidrome
yes, gonic no), "`current` is required unless `id` is empty" (the specification says so, and **neither oracle enforces it**: Navidrome
defaults, gonic stores `-1` or the stray id), "an absent `position` means zero" (both), and "the response carries `changed`
and `changedBy`" (both). So the second edge, if dubplate enforces it, is a **stricter behavior than both oracles** and must be a
declared divergence class or an error 10.

### 7.2 `scrobble` and `submission`

The specification (`open-subsonic-api@bed1688:content/en/docs/Endpoints/scrobble.md:16-26`): `submission` defaults to true;
updating the play count happens on a submission; a `getNowPlaying` entry is made for the file; `time` and multiple `id`
values are allowed. The `stream` page adds that a server "**must not** count access to this endpoint as a play"
(`stream.md:39`). [D]

Observed against `getSong`'s `playCount` (and, on Navidrome, `getNowPlaying`), baseline zero: [O]

| Action | Navidrome | gonic |
|---|---|---|
| `stream` of the track (raw) | no count | no count |
| `scrobble submission=false` | **now-playing only**; the song appears in `getNowPlaying`, count unchanged | **the play count rose to 1** (`getNowPlaying` does not exist on gonic) |
| `scrobble submission=true` | count 1, a last-played time | count 2 |
| `scrobble` with no `submission` | counts (default true) | counts |
| `scrobble` with `time` | counts; the last-played time takes the given time on a first play (a past `time` did not move the latest time back) | counts |
| `id=[a,b]` with two `time` values | **both counted** (b got a 2017 last-played time) | **only a counted**; b's count stayed unset |
| `id=[a,b]` and **one** `time` | error 0, "Wrong number of timestamps" (source) | not run |
| unparseable `id` (`bogus`) | `ok`, nothing recorded (the error is logged) | **error 10** with the text "please provide a `id` parameter" (gonic's typed ids fail to parse, `G:server/ctrlsubsonic/handlers_common.go:58-61`) |
| well-formed unknown `id` (`tr-999`) | not run | not run; [S] the lookup fails and the answer is error 0, "error finding track" (`G:server/ctrlsubsonic/handlers_common.go:71-73`). Test owed (harness) |
| `submission=bogus` | **not counted** (parsed as false) | counted |
| two immediate scrobbles | each counted; no de-duplication | each counted |

[S] Navidrome: `submission` defaults true, a submission records each id with its time, a non-submission records now-playing for
the first id only and logs a warning if more were sent, and an error is logged and answered `ok`
(`N:server/subsonic/media_annotation.go:164-212`). gonic: **the local play-count update runs on every call; `submission` is passed only
to external scrobblers**, and only the first `id` is read (`G:server/ctrlsubsonic/handlers_common.go:54-125`).

**The spec's rule** ("a play is recorded when `scrobble` arrives with `submission=true`, matching the oracle; `submission=false` is
now-playing only", `:564-565`) matches **Navidrome and the specification, not gonic**; gonic would be a wrong oracle for this edge.
The spec's "a stream start is a separate event kind and never a play" (`:565-566`) is the same as both oracles' stream rule.

### 7.3 Lesson for dubplate

1. The queue edges: keep the spec's four statements and test that the oracle's leniencies are declined: `current` absent or
   absent from the list should be an error 10 (the specification's rule), not a silent first-entry default. Clear on no `id`
   (Navidrome, the spec). Drop nothing silently: an unknown id in a saved queue should error or be kept and flagged, not
   vanish.
2. The empty-queue response: Navidrome's `playQueue` element with a zero time is what a client expects; gonic's missing element is
   the form a strict schema (the specification's `PlayQueue` required fields) would reject.
3. `scrobble`: follow Navidrome and the specification. gonic's choice to update stats before looking at `submission` is the
   negative lesson, and its first-id-only reading means a batch from an offline client loses plays.
4. An unknown id should be error 70 (data not found), not Navidrome's silent `ok`. gonic's error 10 was observed for an unparseable id
   (`bogus`); for a well-formed unknown id its source gives error 0 on `scrobble` and a silent store on `savePlayQueue`, neither
   observed.

**Feeds** slice 7 (events, scrobble, play history), slice 3 (the play queue), Task 7 (the ledger rows).

---

## 8. The folder view

F2 (real paths differ from tags), then F1. [O]

| | Navidrome | gonic |
|---|---|---|
| Source of the tree | **Synthesized from tags.** `getIndexes` lists artists; an artist's `getMusicDirectory` lists its albums as directories; an album's lists songs. The tree is Artist, Album, Track. | **The real directory tree.** `getIndexes` lists the top-level directories as "artists"; every level below is a real directory, to any depth (a four-deep path `Deep/One/Two/Three` appeared as four nested directories). |
| A song's `path` field | **Synthetic**: `Zed Artist/Different Album/01 - Real Title.flac` for a file really at `Misc/Mixtape/01 - Real Title.flac`. Artist and album come from tags, the title from the tag, the track number from the tag. | The real relative path (`Misc/Mixtape/01 - Real Title.flac`) |
| Directory names | the album and artist tags | the directory's real name (`Mixtape`, `Three`); the song's `album` reads the directory name, not the tag |
| Two tracks of one album with different track artists | one album directory under the **album artist** | one directory holding both, each song keeping its own artist |
| A loose file at the library root | appears (artist and album from tags) | **not scanned** (3 tracks scanned of 4; the artist never appears in `getArtists`). The cause is unconfirmed; confirming it needs gonic's scanner read at `v0.22.0` or its scan log. |
| Ids | opaque strings, no prefix | **typed prefixes**: `ar-`, `al-` (for album and for every directory), `tr-` |
| `getIndexes` ignored articles | `The El La Los Las Le Les Os As O A` | empty |

[S] Navidrome's own documentation says it plainly: "There are currently no plans to support browse-by-folder. Endpoints for this
functionality (Ex: `getIndexes`, `getMusicDirectory`) returns a simulated directory tree, using the format:
`/Artist/Album/01 - Song.mp3`" (`navidrome/website@b0fe572:content/en/docs/developers/subsonic-api.md`, the second bullet under
"Supported Subsonic API endpoints"). [D] gonic's folder handlers are `G:server/ctrlsubsonic/handlers_by_folder.go:23`. The scanner
keys albums by their parent directory (`G:scanner/scanner.go:300-320`).

**The spec's design (`:376`, `:276-280`)**: a folder view "derived from the paths the pipeline chose, served from a projection
table rather than a filesystem walk", with gonic authoritative for `getIndexes`, `getMusicDirectory`, and `getMusicFolders` and the
folder view a declared divergence class against Navidrome. The observations confirm the class and sharpen it:

- Against **Navidrome**, dubplate's folder view diverges in tree shape, `path` values, and root-level files (declared class).
- Against **gonic**, the *shape* is the right oracle (a real path tree, real directory names, arbitrary depth), but gonic's
  **typed id prefixes** and its **"album = directory name"** reading are gonic specifics dubplate's opaque-id rule (`:379`) and tag-led
  album model decline. gonic's skipping a root-level file is a scanner choice, not a folder-view contract.
- dubplate's folder view is derived from the paths the pipeline chose, so wherever the pipeline's layout equals the tag layout
  (Artist, Album, track), dubplate's real tree and Navidrome's synthetic tree coincide, and gonic's real-tree shape matches too.
  The pipeline's actual layout is not read here; the harness's folder rows should state which case each fixture is.

**Lesson.** The folder view is small when the layout is the tag layout. The cost the spec worries about (doubling the browse
surface and id space) is real and is Navidrome's reason for refusing it; the evidence adds one more cost: **every directory is an
id**, which gonic spends on typed prefixes and dubplate would spend on minted rows. Whether folder browsing ships at launch is
scope and is Geoff's; this section gives the shape and the oracle assignment for Task 7.

**Feeds** slice 3 (browse views), Task 7 (the folder-browsing fork at `:925-926`, the oracle assignment).

---

## 9. The `.nsp` smart-playlist format

[D] `navidrome/website@b0fe572:content/en/docs/usage/features/smart-playlists/index.md` (Apache-2.0, quoted).

**Format.** "They are created using JSON objects stored in files with a `.nsp` extension." A file holds, per the page's examples
and tables: `name`, `comment`, a rule group under `all` or `any` (nestable), `sort` (a field name or a comma list with `+` or `-` per
field, plus `random`), `order`, `limit` or `limitPercent` (1 to 100), and `refreshDelay` (a duration with `d` and `w` units).
Rules are single-key objects naming an operator over a field, with these operators: `is`, `isNot`, `gt`, `lt`, `contains`,
`notContains`, `startsWith`, `endsWith`, `inTheRange`, `before`, `after`, `inTheLast`, `notInTheLast`, `inPlaylist`, `notInPlaylist`,
`isMissing`, `isPresent`. The page's field table lists about 70 track, album, artist, history, and MusicBrainz fields, with album and
artist fields prefixed `album` and `artist`, plus `library_id` and any custom tag. Dates are `"YYYY-MM-DD"`; booleans are bare.

One example, quoted from the page (Example 2, "80's Top Songs"):

```json
{
  "all": [
    { "any": [{ "is": { "loved": true } }, { "gt": { "rating": 3 } }] },
    { "inTheRange": { "year": [1981, 1990] } }
  ],
  "sort": "year",
  "order": "desc",
  "limit": 25
}
```

**Import and lifecycle.** "Smart Playlists are imported the same way as regular (`.m3u`) playlists, during the library scan. Place
your `.nsp` files in any folder in your library or the path specified by the `PlaylistsPath` configuration." Ownership:
"By default, Smart Playlists are owned by the first admin user." The track list "is read-only and cannot be edited directly"; a playlist is
re-evaluated on access with a minimum spacing (`SmartPlaylistRefreshDelay`, default `5s`), and a per-file `refreshDelay` overrides it.
Rules that reference another playlist match nothing, with a logged warning, when that playlist is not visible to the owner. Feishin is
named as an editor (the page shows its `.nsp` editor). [D]

**Observed** (Navidrome, F2, a `Quick.nsp` placed in the library root, then `startScan`): the file imported as a playlist
owned by the admin, `readonly: true`; the playlist **list showed `songCount: 0` before first access and `getPlaylist` showed 3
after**, with a `validUntil` timestamp (the refresh deadline); the rule `contains title` with `sort title`, `order
desc`, `limit 3` returned three songs in descending title order (one song twice, because of the nested second library in this
fixture state). [O] So evaluation is **lazy, at first access**, and the playlist is a real read-only playlist from the client's
view.

**What this answers (`:934`, "whether smart playlists import Navidrome's `.nsp` files for switchers").** The format is a small JSON
rule language over fields dubplate's catalog also holds (title, album, year, dates, rating, loved, play count, last played,
MusicBrainz ids). A switcher import is **mechanically a parse and a translation** to dubplate's own rule form, never a runtime
dependence. Whether the group's retired Navidrome held any `.nsp` file is a fact only the archive carries, and this lane does not open
the archive: **routed to slice 8** (the import), as one read-only check of the archive's playlist folder.

**Feeds** slice 8 (the import), a later smart-playlist slice, Task 7 (the `.nsp` item).

---

## 10. What the two test suites chose to test

### 10.1 Navidrome's `server/subsonic/e2e` suite

The suite runs the full HTTP stack against a real SQLite database and a small fake in-memory library (several artists, genres, and years) with two users, one admin and one regular, and a golden database snapshot for isolation (`N:server/subsonic/e2e/doc.go:1-40`).
Areas by file, with the number of examples (`It(` blocks) in each, from `N:server/subsonic/e2e/`:

| Area | Examples | What it covers |
|---|---|---|
| playlists | 52 | CRUD, permissions, smart-playlist protection, a boolean-string normalization regression |
| transcode (the extension) | 46 | decision errors, direct play, transcode, response shape, player bitrate caps, format-aware default bitrate, forced formats |
| browsing | 36 | music folders, indexes, artists, directories, album, song, genres, artist and album info, top and similar songs |
| album lists | 25 | album list types, starred, highest, `getAlbumList2`, now playing, random, songs by genre |
| media annotation | 20 | star and unstar, rating, scrobble, `reportPlayback` |
| searching | 20 | `search2`, `search3`, MBID search, CJK search, the legacy backend |
| media retrieval | 19 | stream, download, cover art, avatar, lyrics |
| stream (legacy) | 18 | raw and direct, explicit format, downsampling by bitrate only, `timeOffset`, creation failure, empty output |
| sharing | 18 | CRUD, cross-user isolation, downloadable default |
| multi-library | 17 | folders, per-endpoint library filtering, cross-library playlists and shares, access control |
| sonic similarity | 12 | with and without the plugin |
| artwork | 11 | art serving |
| radio | 11 | internet radio CRUD |
| bookmarks and play queue | 10 | both |
| multi-user | 7 | admin-only restrictions, regular-user browsing, `getUser` and `getUsers` authorization |
| system | 6 | `ping`, `getLicense`, `getOpenSubsonicExtensions` |
| users | 4 | user endpoints |
| scan | 3 | scan endpoints |
| lyrics | (structure only) | v1 line-level, v2 enhanced, legacy artist and title |

About 5,500 lines across the suite. The shape is one area per Subsonic family with **cross-cutting suites for the three failure
classes**: multi-user isolation, multi-library filtering, and permissions. [S] (counts from `grep` of `It(` per file; the examples'
text is not reproduced.)

### 10.2 gonic's golden fixtures

`G:server/ctrlsubsonic/testdata/` holds **129 files**: 127 golden responses and two audio files (`5s.flac`, `10s.flac`). By
area: browse and list ("get ..." names) **71**, internet radio **27**, search **12** (`search2` and `search3`, including empty,
UUID, and per-folder), `getSongsByGenre` 4, `setRating` 3, `getIndexes` 3, playlists 6 (create, update, delete, list, get), star and
unstar 1. The goldens' names carry the fixture state (`..._music_folder_0`, `..._starred_admin`, `..._empty_zero_tracks`,
`..._va_compilation`, `..._multi_disc_contributors`, `..._unicode`), so the suite tests each browse endpoint against named album shapes
(compilation, multi-disc, collaboration credit, zero tracks, split credits, Unicode names). Separate Go test files hold the stream
(`handlers_raw_test.go`), mutation, playlist, bookmark, and radio tests. The two oracles agree that **browse, list, and search carry
the most tests** and diverge in what they test beside them: Navidrome tests permissions, isolation, and transcode decisions; gonic
tests album shapes and radio. [S]

### 10.3 Lesson

The harness's own area list should include the three cross-cutting classes Navidrome tests (multi-user isolation, per-endpoint folder
filtering, permission refusals) and the album-shape fixtures gonic names (compilation, multi-disc, collaboration credit, empty album,
Unicode). dubplate's group-scale model puts permissions and isolation where Navidrome does; its album-as-verdict-unit model makes the
album-shape fixtures the cheapest high-value tests.

**Feeds** slice 2 (the harness's area list), Task 7.

---

## 11. Side observations (not owned, recorded because the probe showed them)

- **Cover art.** At F1 neither oracle negotiated WebP: `getCoverArt` with `Accept: image/webp`, `image/jpeg`, and none all returned
  `image/jpeg` for a JPEG cover (Navidrome 17,582 B, the original; gonic 21,141 B, re-encoded). `size=64` returned a smaller JPEG on both.
  An unknown id answered an error envelope on both. **A request with no `id` returned a 69,228-byte `image/webp` placeholder with
  HTTP 200 on Navidrome** and an error envelope on gonic. This is evidence for the spec's WebP-when-accepted rule having no oracle at
  all and for Task 2's `Accept: image/webp` question (survey "Not verified" `:267-268`). Feeds slice 4.
- **Playlists and `createPlaylist`.** gonic created an empty-named playlist from a `createPlaylist` with no parameters, and Navidrome
  answered error 0, "Internal Server Error: required parameter name is missing" (a missing parameter reported as a server error).
- **Ids.** Navidrome ids are opaque 22-character strings (and a cover-art id embeds a content hash); gonic ids are typed prefixes with
  a database number (`ar-1`, `al-3`, `tr-1`). The spec's opaque-surrogate rule (`:379`) matches Navidrome. Navidrome's
  documentation notes ids are strings "normally MD5 hashes or UUIDs" and that some clients insist on converting ids to integers.
- **Reported protocol version.** Navidrome reports 1.16.1, gonic 1.15.0 (Task 0 found the same).

---

## 12. Divergence-ledger seed rows (for Task 7; pre-seeded under settled item 6)

Columns are the plan's: id, endpoint or area, divergence class, oracle, the oracle's behavior, dubplate's intended behavior, evidence
pointer (a section here), and the test owed. **"Intended" is the spec's statement where one exists (cited) and "to settle (Task 7)"
otherwise; Task 7 owns every disposition.** Classes: *declared* is a class the spec already declares (`:285-287`); the rest are
proposed names: *absent-by-design* (dubplate answers where the oracle does not), *oracle-lenient* (the oracle accepts what the
contract forbids), *oracles-disagree*, *oracle-wrong* (the oracle contradicts the contract), *feature-absent* (one oracle lacks the
feature).

| id | Endpoint or area | Class | Oracle | Oracle behavior | Dubplate intended | Evidence | Test owed |
|---|---|---|---|---|---|---|---|
| D4-01 | unimplemented endpoint | absent-by-design | Navidrome | plain-text HTTP 404, 410, or 501 outside the envelope | error 0 in an envelope naming the endpoint (spec `:670`) | s1.1 | harness: one probe per non-launch path |
| D4-02 | unimplemented endpoint | oracle-wrong | gonic | HTTP 200, envelope error 70 "view not found" | error 0 (spec `:670`) | s1.1 | same probe |
| D4-03 | `getOpenSubsonicExtensions` auth | oracles-disagree | gonic | requires credentials | unauthenticated (spec `:648`) | s5 | harness: unauthenticated call |
| D4-04 | extension set | feature-absent | gonic | declares 3 of 10 | launch subset: to settle (Task 7) | s4 | harness: extension list assertion |
| D4-05 | `musicFolderId` set | oracles agree | both | honored on the ten browse endpoints, ignored elsewhere | the ten (spec's endpoint list) | s2 | harness: per-endpoint filter table |
| D4-06 | unknown `musicFolderId` | oracles-disagree | Navidrome | error 70 (two message texts) | error 70: to settle (Task 7) | s2 | harness: unknown-id row per endpoint |
| D4-07 | unknown `musicFolderId` | oracle-lenient | gonic | `ok` with an empty result | error 70: to settle (Task 7) | s2 | same |
| D4-08 | transcode, Range on a filling entry | absent-by-design | Navidrome | 200, whole entry, `Accept-Ranges: none` | to settle (Task 7); spec `:434-443` is the divergence | s3.2 | slice 5: race test; harness: filling-entry row |
| D4-09 | transcode, concurrent request on a filling entry | oracle-wrong | gonic | blocks until the first encode ends, then a full-file range answer | share the encode (spec `:433-434`) | s3.2 | slice 5: second-listener test |
| D4-10 | transcode, `estimateContentLength` | oracle-wrong | Navidrome | ignores `timeOffset`; runs about 2.4% high | to settle (Task 7) | s3.3 | slice 5 |
| D4-11 | transcode, `estimateContentLength` | oracle-lenient | gonic | subtracts the offset; runs about 2.4% high | to settle (Task 7) | s3.3 | slice 5 |
| D4-12 | transcode, `timeOffset` caching | oracles-disagree | gonic | not cached; each request re-encodes; no Range | offset in the cache key (Navidrome shape) | s3.3 | slice 5 |
| D4-13 | transcode, failed encode | oracle-wrong | gonic | clean end of stream after a killed ffmpeg | evict, error, truncated encode never serves (spec `:447-449`) | s3.4 | slice 5: failed-exit test |
| D4-14 | transcode, failed encode | oracles-disagree | Navidrome | aborts the connection; evicts the entry | evict, request returns 500 (spec `:447-449`) | s3.4 | slice 5 |
| D4-15 | transcode, client disconnect | oracles-disagree | Navidrome | continues to completion by default | kill only if no other reader (spec `:449-450`) | s3.4 | slice 5 |
| D4-16 | transcode, concurrency cap | oracles-disagree | Navidrome | no queue; HTTP 429, `Retry-After: 5`, error-0 envelope | bounded queue, then error 0 (spec `:450-452`) | s3.5 | slice 5 |
| D4-17 | `transcoding` extension | feature-absent | gonic | error 70, none | owner fork (Task 7) | s4 | harness: extension-gated test set |
| D4-18 | `getTranscodeDecision`, AAC | oracle-lenient | Navidrome | AAC profiles removed from every posted `ClientInfo` | follows the AAC-container ruling | s4.3 | slice 5 |
| D4-19 | `getTranscodeDecision`, device ceiling | to settle | Navidrome | a pinned player's bitrate ceiling applies to a posted `ClientInfo`; a forced format applies only if the client can play it | to settle (Task 7), candidate adopted | s4.3 | slice 5 |
| D4-20 | CORS | oracles-disagree | both | `*`, no credentials; gonic stamps every response, Navidrome only with `Origin` and reflects headers | to settle (Task 7) | s5 | slice 3: preflight rows |
| D4-21 | credential channels | declared | both | query token, query `p` and `enc:`, form body; no header, no `apiKey` | the apiKey path is the declared class (spec `:285-287`) | s5.1 | slice 6: apiKey rows |
| D4-22 | `search3`, empty query | oracles agree | both | all rows, paged | all, paged, deterministic order (spec `:399`) | s6.1 | harness: whole-library page walk |
| D4-23 | `search3`, absent `query` | oracle-lenient | Navidrome | answers as an empty query | error 10 (contract's required parameter) | s6.1 | harness |
| D4-24 | `search3`, one-character query | oracles-disagree | Navidrome, gonic | Navidrome none; gonic substring flood | prefix lookup for under three characters (spec `:403-405`) | s6.2 | harness: golden list, not oracle diff |
| D4-25 | `search3`, `*` | oracles-disagree | gonic | matches nothing | to settle (Task 7) | s6.1 | harness |
| D4-26 | `getAlbumList2`, `size` of 0 or negative | oracle-wrong | both | Navidrome returns all (clamp is 500 above); gonic returns none for 0 and a SQL error for a negative | clamp and validate | s6.3 | harness: edge-size table |
| D4-27 | `savePlayQueue`, no `id` | oracles-disagree | gonic | error 10, queue kept | clears the saved queue (spec `:567-568`) | s7.1 | slice 7: queue rows |
| D4-28 | `savePlayQueue`, `current` absent or absent from the list | oracle-lenient | both | Navidrome defaults to the first entry; gonic stores `-1` or the stray id | `current` required unless `id` is empty (spec `:568`) | s7.1 | slice 7 |
| D4-29 | `savePlayQueue`, unparseable `id` (`bogus`) | oracles-disagree | Navidrome | drops silently (gonic: error 10; a well-formed unknown id is stored silently, [S], test owed) | to settle (Task 7) | s7.1 | slice 7 |
| D4-30 | `getPlayQueue`, none saved | oracles-disagree | gonic | no `playQueue` element | element with the zero time (Navidrome shape): to settle | s7.1 | slice 7 |
| D4-31 | `scrobble`, `submission=false` | oracle-wrong | gonic | counts a play | now-playing only (spec `:564-565`) | s7.2 | slice 7: event-kind rows |
| D4-32 | `scrobble`, multiple ids | oracle-wrong | gonic | first id only | every id; times must match ids (Navidrome shape): to settle | s7.2 | slice 7 |
| D4-33 | `scrobble`, unparseable id (`bogus`) | oracles-disagree | both | Navidrome `ok` silently; gonic error 10 (a well-formed unknown id gives error 0, [S], test owed) | error 70: to settle | s7.2 | slice 7 |
| D4-34 | `scrobble`, `submission=bogus` | oracles-disagree | both | Navidrome not counted; gonic counted | to settle (Task 7) | s7.2 | slice 7 |
| D4-35 | folder view, tree shape and `path` | declared | Navidrome | synthesized Artist/Album/Track tree and path | folder view derived from pipeline paths (spec `:376`) | s8 | harness: folder rows graded by gonic |
| D4-36 | folder view, tree shape | oracles-disagree | gonic | real tree to any depth; album = directory name | gonic grades shape, not ids (spec `:277-278`) | s8 | harness |
| D4-37 | folder view, root-level file | oracle-wrong | gonic | not scanned (cause unconfirmed) | not graded | s8 | harness: if a root file ever ships |
| D4-38 | ids | oracles-disagree | gonic | typed prefixes | opaque surrogates (spec `:379`) | s8, s11 | harness: id-shape assertion |
| D4-39 | `getCoverArt`, `Accept: image/webp` | feature-absent | both | never negotiated | WebP when accepted (spec art rule) | s11 | slice 4: art rows |
| D4-40 | `getCoverArt`, no `id` | oracle-lenient | Navidrome | 200 with a placeholder image | error 10 | s11 | slice 4 |
| D4-41 | `.nsp` | feature-absent | gonic | no smart playlists | none at launch; import is slice 8's | s9 | slice 8 |
| D4-42 | missing required parameter | oracle-lenient | gonic | `star` and `unstar` with no `id` answer `ok` | error 10 | s1.2 | harness: missing-parameter sweep |
| D4-43 | missing required parameter | oracle-lenient | Navidrome | `getSongsByGenre` and `getLyrics` answer `ok` empty | error 10 | s1.2 | harness |

**Rows: 43.**

---

## Routed

- **Slice 3:** the unimplemented answer (D4-01, D4-02), error 10 for missing parameters, the `musicFolderId` ten and its unknown-id
  answer, the folder-view shape and path rule, the CORS policy and the credential channels, search edge behavior.
- **Slice 4:** Navidrome and gonic never negotiate WebP; Navidrome's no-`id` placeholder image (D4-39, D4-40). Capturing real client
  `Accept` headers remains Task 2's.
- **Slice 5:** every row of section 3 and section 4: the Range and filling rules, the offset in the cache key, the estimate, the failed
  exit and disconnect rules, the concurrency answer, the decision-and-token shape, the device-ceiling ordering, and AAC withheld from the
  extension. `djherbis/stream` adoption stays slice 5's race test.
- **Slice 7:** the play queue and `scrobble` rows (D4-27 to D4-34).
- **Slice 8:** whether the retired Navidrome's archive holds any `.nsp` file (a read-only check of its playlist folder, not done here);
  Navidrome's schema is unobserved by this lane.
- **Slice 2 (harness):** the area list in section 10.3, the search comparison note (golden lists for short queries), a fixture with two
  collections for `musicFolderId`, and an edge-size table. The throwaway probe scripts (`work/task-4/*.py`, `*.sh`) are workstation-scratch staging only and did
  not land in the repo.
- **Task 2 (client demands):** which of the section 1.2 empty-success endpoints a roster client fails on, and whether any roster client
  posts a `ClientInfo`.
- **Task 7:** every ledger row's disposition; the empty-success set (the integration-dependent family is the server-side input); the
  `transcoding`, folder-browsing, and CORS items.
- **New hook candidates:** none. Section 4's token bound to the file's update time is a design idea for slice 5, not a hook.

## What could not be done, or is unconfirmed

- gonic with two music paths, and Navidrome with a multi-valued `musicFolderId`: not observed (a two-collection harness fixture owes
  both).
- The concurrency cap (`Transcoding.MaxConcurrent`) was not exercised; only its source and documentation were read.
- gonic's skipping of a root-level file: the cause is unconfirmed.
- Navidrome's `scrobble` with one `time` for several ids was read in source (error 0) and not run.
- gonic's `getTranscodeDecision` path does not exist, so no interaction with its own per-client transcode preference was studied; gonic's
  transcoding preference store was not read.
- The probe credentials, data directories, and the login token were throwaway and deleted; the scratch fixtures stay under the work directory.

## Prior art consulted

Navidrome and gonic (adopted as oracles, read for ideas and lessons, never transliterated); `navidrome/website` (Apache-2.0, quoted:
the smart-playlist page, the Subsonic-compatibility page, and the options table); the OpenSubsonic specification (the adopted contract,
quoted for `saveplayqueue`, `scrobble`, `stream`, `TranscodeOffset`, and the `musicFolderId` rows). `djherbis/stream` is named only
because Navidrome's filling-entry behavior rests on it (through `fscache`), and its adoption stays slice 5's race test (`:445-447`).
Nothing adopted or declined anew. Neutral on hooks.
