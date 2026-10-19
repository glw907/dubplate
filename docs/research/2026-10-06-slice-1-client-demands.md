# Slice 1, Task 2: the client roster's demands

Date read: 2026-10-07. Staged outside the repo; lands at A-land as
`docs/research/2026-10-06-slice-1-client-demands.md`.

## Pins read

From `2026-10-06-slice-1-evidence-pins.md` (Task 0), each client read at its release commit. A new clone of each repository, checked out at the pin, was read as text only; nothing from a clone was built or run.

| Client | Pin | Abbreviation in citations |
|---|---|---|
| Feishin v1.17.0 | `jeffvli/feishin@5c169aca8d550032c9aa6e8e2b9f624d03d42fec` | F |
| Rufin v0.16.6 | `screwys/Rufin@aad2f44e5171817c96d65ff3f75522ef7aab19c4` | R |
| Amperfy v2.1.1 | `BLeeEZ/amperfy@c165a7a8df7bbf4f593dd75b2e34719d74c7851f` | A |
| Tempus v4.28.0 | `eddyizm/tempus@687d97fdde38f231f2168e8fd47755a9b9092684` | T |

Player libraries read at their own pins (each cited where used): `dimitris-c/AudioStreaming@a729ee2` (v1.4.4, MIT, the version `A`'s `Package.resolved` pins), `androidx/media@1.9.2` (Apache-2.0, the version `T`'s `gradle/libs.versions.toml` pins), `axios/axios@v1.17.0` (MIT, `F`'s `package.json:102` range `^1.17.0`; the tag was read, the installed build was not). The OpenSubsonic specification is `opensubsonic/open-subsonic-api@bed1688`.

Citation roots, so each cell reads as `repo@commit:path:line`:

- F root: `src/renderer/api/subsonic/` unless a full path is given. `api.ts` is `subsonic-api.ts`, `ctrl` is `subsonic-controller.ts`.
- R root: `crates/sources/src/subsonic/`. `client` is `client.rs`, `refresh` is `refresh.rs`, `mod` is `mod.rs`.
- A root: `AmperfyKit/Api/Subsonic/`. `ServerApi` is `SubsonicServerApi.swift`, `Syncer` is `SubsonicLibrarySyncer.swift`.
- T root: `app/src/main/java/com/eddyizm/tempus/`. `subsonic/api/*` are the Retrofit service files, `repository/*` the data layer.

## Questions this record owns

1. Every endpoint each client calls, and for each whether the client fails, degrades, or carries on when the server answers error 0. The empty-success candidates follow from it (spec `:670-672`).
2. Every OpenSubsonic extension each client checks, and what it does with and without it (the `transcoding` extension above all, spec `:412-419`).
3. Whether each client browses by ID3, folder, or both, and whether it passes `musicFolderId`, and on which endpoints (spec `:925-926`).
4. Which stream formats and containers each player accepts, from the player library (AAC container item, spec `:427-431`, `:946`).
5. Whether each client runs in a browser context that needs CORS (spec `:650-655`), and how each sends credentials (planning-miss check 3).
6. Whether any client sends `Accept: image/webp` on `getCoverArt` (survey "Not verified" `:267-268`).
7. Secondary inputs: the Symfonium notes (spec `:885`) and Melodee's compatibility matrix.

Outcome states are in "Question outcomes" at the end.

## Evidence channels and method

Every claim below is from one of three channels, and the table cells name which: pinned source reading (every row unless marked), published documentation (the spec, the Symfonium wiki, library READMEs), and one local observation (Chromium 154.0.8037.57 headless, Fedora build, against a throwaway `python3` HTTP server I wrote, bound to `127.0.0.1:47121`, stopped afterward; it observes browser behavior, not a roster client). No Navidrome or gonic instance was started, and no container was started or left. No desktop client was launched (Task 3 owns that). No Android emulator exists on the workstation, so Tempus and the Android player are source-only. Amperfy cannot run on Linux, so it is source-only. Clean-room rule 1 holds: all four clients are GPL-3.0 (Task 0), no client code appears here, and every behavior is stated in dubplate's own words.

"Error 0" means a well-formed `subsonic-response` with `status` "failed" and `error.code` 0, answered with HTTP 200. Whether a client "fails, degrades, or carries on" is read from what its code does with that body. Where the answer depends on a runtime fact the source cannot settle, the cell says "unconfirmed".

## 1. Request shape every client shares (not owned, but the later tables lean on it)

Findings that bear on the response layer and on slice 3's path rule. Each is a measured demand, not a recommendation.

| Demand | Feishin | Rufin | Amperfy | Tempus |
|---|---|---|---|---|
| Path suffix | `.view` on every route but two (`getSimilarSongs`, `getSimilarSongs2`, bare); the contract list is `F:api.ts:21-385`, bare paths at `:211`, `:219` | always `.view`; `endpoint()` strips any `.view` and re-adds it, `R client:1126-1138` | always `.view`, `A ServerApi:205-206` | never `.view`; every Retrofit route is bare, `T subsonic/api/system/SystemService.java:12-19` and the other services |
| Response format asked | `f=json`, `F:api.ts:555-561` | `f=json`, `R client:1013` | none sent, so XML; every response is parsed with `XMLParser`, `A ServerApi:308-311` | `f=json`, `T subsonic/Subsonic.java:161` |
| Protocol `v` sent | `1.13.0`, `F:api.ts:557` | `1.16.1`, `R client:1141` | `1.13.0` (token) or `1.11.0` (legacy), `A ServerApi:103-104` | `1.15.0`, `T subsonic/Subsonic.java:22` |
| Client name `c` | `Feishin` | `Rufin`, `R client:1140` | `Amperfy`, `A ServerApi:229` | `preferences.getClientName()`, `T subsonic/Subsonic.java:160` |
| Base URL path | base URL plus `/rest`; stream and cover URLs are built two ways, `F ctrl:1976` uses `server.url` raw while the API client at `F:api.ts:500-501` uses `getServerUrl`; Task 3 measures it | base path is kept: `normalize_base_url` strips a trailing `/rest` and keeps a prefix, `R client:1107-1124` | `activeBackendServerUrl` plus `rest/<action>.view`, `A ServerApi:201-206` | `serverUrl + "/rest/"`, `//rest` collapsed, `T subsonic/Subsonic.java:140-143` |
| Each client's HTTP 404 handling | any HTTP error goes through axios error branch, `F:api.ts:586-611` | any non-ok body is an error; HTTP status handling in `remote_http` (not read further) | HTTP 404 is converted to Subsonic error 70 ("requested data not found") before the body is read, `A ServerApi:976-987` | Retrofit `response.isSuccessful()` false, repositories fall to the empty branch (see 2) |

Observations that matter beyond the table:

- Amperfy is the only roster client that parses XML. Its XML demands (all from the parser delegates at `A`): the element for a song may be named `song`, `entry`, `child`, or `episode`, and a song is skipped unless `isDir` reads as a literal Swift `Bool` ("true" or "false"; `A SsSongParserDelegate.swift:45`). `created` and a song's `starred` must be ISO 8601 with fractional seconds or the date is dropped (`A SsSongParserDelegate.swift:134-137`, `A SsPlayableParserDelegate.swift:90-97`). `starred` on any item makes it a favorite by presence alone (`A SsArtistParserDelegate.swift:68`, `SsAlbumParserDelegate.swift:71`). `getArtists` must carry `albumCount` per artist because the initial album sync counts pages from it (`A Syncer:112-121`; see 2). The ping must carry `version` on `subsonic-response` (`A SsPingParserDelegate.swift:44-47`), and `ping` is valid when no `error` element appears, not when `status` reads "ok" (`:62-63`).
- Feishin's song normalizer dereferences `contentType` with no guard (`F` `src/shared/api/subsonic/subsonic-normalize.ts:207`), so a song entry without `contentType` throws inside the query. Its zod schemas also list `parent`, `path`, `size`, `suffix`, `type`, `isDir`, `isVideo`, `created`, and `discNumber` as required on a song (`F` `src/shared/api/subsonic/subsonic-types.ts:179-217`), but these are types only: runtime response validation is behind a build flag (`F` `src/renderer/api/response-validation.ts:6`, `:63-65`). Rufin needs only `id` on an album, song, artist, or playlist and defaults the rest (`R` `item.rs:387,445,547,572`). Tempus reads whatever Gson finds and null-checks the containers.
- Feishin calls `ping` and `getUser` before anything else on each connect (`F` `src/renderer/hooks/use-server-authenticated.ts:77-98`) and reads `ping.openSubsonic` and `ping.serverVersion`: if either is absent it never asks for the extension list (`F ctrl:1443-1445`). So a server that omits either hides every extension from Feishin.

## 2. Q1: endpoints called, and what error 0 does

Reading rule. "Hard" means the client dereferences the data object without a guard, so an error-0 body throws or breaks the view. "Soft" means a guarded read, so error 0 reads as empty. A row says what the user sees.

### 2a. Feishin v1.17.0

How error 0 reaches the controller. The response interceptor skips its whole error branch when `error.code` is 0, so no toast shows and `response.status` stays 200 (`F:api.ts:393-419`, branch at `:398`). Any other error code toasts the server message (when logged in) and rewrites `status` to the code, which the controller treats as failure (`:407-417`). So for Feishin, error 0 is a silent 200 whose body lacks the data object, and the controller's guard decides the rest. The rows below read each controller function against that.

| Endpoint | Where it is called (cite) | What error 0 does |
|---|---|---|
| `getUser` | connect and login, `F ctrl:316-373` (authenticate), `:2137-2155`; connect gate `use-server-authenticated.ts:77-88` | Hard (`resp.body.user.adminRole`, `:369`). Fails: login and every reconnect fail. Must carry data. |
| `ping` | `F ctrl:1435-1445` | A failed status (non-error-0) throws (`:1437`); with 200 the body decides extension discovery. Must carry `openSubsonic` and `serverVersion`. |
| `getOpenSubsonicExtensions` | `F ctrl:1447-1458`; sent with credentials (auth params added to every request with a server, `F:api.ts:499-511`) | Failure is caught by the caller and logged (`use-server-authenticated.ts:119-127`): carries on, no features. |
| `jukeboxControl` (status probe) | connect, `F ctrl:1483-1498` | Probe only: a body with `error` logs and skips the jukebox feature (`:1488-1493`). Carries on. |
| `getAlbumList2` | album lists and home, `F ctrl:718-745`, `:852-880` | Hard (`res.body.albumList2.album?.map`, `:736`; `.album.length`, `:877`). Fails. Must carry data. |
| `getAlbum` | `F ctrl:577-591` | Hard (`ssNormalize.album(res.body.album)`, `:590`). Fails the album page. |
| `getArtist` | `F ctrl:483-502` | Hard (`res.body.artist`, `:496`). Fails the artist page. |
| `getArtists` | artist lists, `F ctrl:533-575`, `:909-947` | Soft (`res.body.artists?.index || []`, `:546`, `:922`). Reads as empty. |
| `getAlbumInfo2` | only `navidrome-controller.ts:407`, not the Subsonic controller | Never called on a server that is not typed Navidrome. |
| `getArtistInfo2` | `F ctrl:504-531` | Soft: failure returns null, and `artistInfo?.biography` is guarded (`:518-531`). Quiet. |
| `getTopSongs` | `F ctrl:2087-2135` | Soft (`topSongs?.song || []`, `:2107`). Empty. Uses `id` when `topSongsByArtistId` is declared, else `artist` (`:2092-2096`). |
| `getSimilarSongs`, `getSimilarSongs2` | radio, `F ctrl:887-907`, `:954-985`, `:1502-1527` | Soft (`?.song`, `:901`, `:968`, `:1516`): empty. |
| `getSong` | `F ctrl:1528-1541` | Hard (`ssNormalize.song(res.body.song)`, `:1541`). |
| `search3` | album, artist, and song lists and search: `F ctrl:592-660`, `:1543-1570`, `:1696-1727`, `:2379-2410`; the Tracks list sends `query` as an empty string (`:1702-1712`) | Soft (`searchResult3?.…`, `:614`, `:1567`, `:1721`, `:2400-2406`): empty. Must return data for the Tracks page to show anything. |
| `getStarred` | favorites, `F ctrl:660-690`, `:808-820`, `:1599-1630`, `:1828-1840` | Soft (`starred?.…`): empty. (Feishin calls `getStarred`, not `getStarred2`; contract `F:api.ts:241-248`.) |
| `getGenres` | `F ctrl:1138-1178` | Soft (`genres?.genre || []`, `:1147`). |
| `getRandomSongs` | `F ctrl:1362-1388`; shuffle, auto-DJ | Soft (`randomSongs?.song || []`, `:1379`). |
| `getSongsByGenre` | `F ctrl:1576-1600` | Soft (`:1589`). |
| `getIndexes` | folder root, `F ctrl:1037-1067` | Soft within the folder page (`indexes?.index?.flatMap`, `:1054`); a non-200 throws with the body. |
| `getMusicDirectory` | folder page, `F ctrl:1088-1135` | Hard (`ssNormalize.folder(directoryRes.body.directory)`, `:1098`). The Folders page fails. |
| `getMusicFolders` | `F ctrl:1193-1209`; sidebar and the folder picker, `F src/renderer/features/shared/api/shared-api.ts:9-18` | Hard (`musicFolders.musicFolder.map`, `:1203`) but the callers only fill a menu, so the picker stays empty. Degrades. |
| `getPlaylists` | `F ctrl:1227-1296` | Soft (`playlists?.playlist || []`, `:1236`, `:1285`). |
| `getPlaylist` | `F ctrl:1212-1226`, `:1302-1322` | Hard (`res.body.playlist`, `:1225`; `.entry?.map`, `:1314`). |
| `createPlaylist`, `updatePlaylist`, `deletePlaylist` | `F ctrl:302-315`, `:414-428`, `:468-481`, `:2186-2279`, `:2411-2426`, `:2472` | A non-200 throws "Failed to …"; error 0 is 200 so a failed write looks like success. `createPlaylist` is hard (`res.body.playlist.id`, `:426`). |
| `star`, `unstar`, `setRating` | `F ctrl:374-396`, `:430-451`, `:2427-2441` | Status-checked only: error 0 reads as success. |
| `scrobble` | `F ctrl:2314-2336` (submission, and start) | Status-checked only; silent. |
| `getPlayQueue` / `savePlayQueue` | `F ctrl:1324-1361`, `:2281-2312`; the by-index forms only when `indexBasedQueue` is declared | `getPlayQueue` is hard (`res.body.playQueue`, `:1350`); a throw on queue restore degrades. |
| `getScanStatus`, `startScan` | `F ctrl:1416-1431`, `:2175-2185`; poll every 2 s while watching (`src/renderer/features/shared/hooks/use-scan-status.ts:11`) | `getScanStatus` is hard (`:1426-1429`); the poll just fails. |
| `getLyricsBySongId` | `F ctrl:2047-2085`; only when `songLyrics` v1 or v2 is declared (`:2051-2062`) | Not called without the extension. |
| `getInternetRadioStations` and the radio writes | `F ctrl:1180-1192`, `:397-411`, `:453-465`, `:2454-2470`; Radio page | Soft (`:1189`): empty. |
| `getTranscodeDecision`, `getTranscodeStream` | only when `transcoding` is declared, `F ctrl:1990-2043` | Hard when `transcoding` is declared. A non-200 decision falls back to the plain stream URL (`:2013-2018`), but error 0 arrives as HTTP 200, so that fallback is skipped. `td` is then undefined (`:2020`), `!td?.canDirectPlay` reads as "needs transcoding" (`:2021`), and the log line reads `td.transcodeReason` with no optional chaining (`:2028`), which throws a TypeError before `appendTranscodeParams` (`:2030-2032`) runs. `getStreamUrl` throws and the track does not play. |
| `stream`, `getCoverArt` | URL builders, `F ctrl:1973-1976`, `:48-75` | Not API calls: see 4, 5, and 6. |

Endpoints in `F`'s contract that the Subsonic controller calls only conditionally: `getAlbumInfo2` (Navidrome controller only) and `reportPlayback` (only under `playbackReport`, `F ctrl:2338-2371`). There is no `download`, `getShares`, `createShare`, or `getLicense` in the contract (`F:api.ts:21-385`).

### 2b. Rufin v0.16.6

How error 0 reaches the caller. Every request goes through `subsonic_json`, which turns any envelope whose `status` is not "ok" into an error (`R client:1043-1068`). Codes 40 to 44 are an auth failure that prompts a re-login; every other code, 0 included, is a server error (`:1060-1067`). Rufin sends `f=json` and reads a JSON `subsonic-response`.

| Endpoint | Where called | What error 0 does |
|---|---|---|
| `ping` | login, `R mod:486-490` | Fails the login (`subsonic_json` error propagates). Must succeed with `status` "ok". |
| `getUser` | login and metadata editing, `R mod:491-499`, `:578-583` | Tolerated: the result is `.ok()` (`mod:499`), so a failure only leaves the user name as typed and disables editing (admin flag false). The code states the intent in its own test name (`R client:1340`). |
| `tokenInfo` | only in the API-key login mode, `R mod:517-530` | Fails that login. Not called in the password mode. |
| `getMusicFolders` | start of every catalog refresh, `R refresh:191-197`, `:358-371` | Failure is logged, the refresh continues unscoped (`:191-197`, `incomplete_collection` at `:16-23`). Carries on. |
| `getAlbumList2` (`alphabeticalByName`, 500 per page, until a short page) | catalog albums, `R refresh:258-295` | Failure marks the refresh incomplete (`:216-218`) and keeps the old catalog; on a first sync the library stays empty. Breaks normal use. Must carry data. |
| `search3` with an empty `query` and `songCount=500`, once per music folder when `getMusicFolders` answered a list (`musicFolderId` set on each pass), once unscoped otherwise | catalog tracks, `R refresh:297-356`; the folder id is added at `:323-325` | Same: incomplete refresh, empty track list. Must carry data, and must honor an empty query, paging by `songOffset`. |
| `search3` with an empty `query` and `artistCount=500` | catalog artists, `R refresh:430-472` | Same, for artists. |
| `getGenres` | `R refresh:373-389`, called from the catalog at `:235-242` | Optional: failure keeps the cached genres (`source.rs:382-388`). |
| `getStarred2` | `R refresh:53-74`; user collections `:95-97` | Optional: failure keeps cached favorites. |
| `getPlaylists`, `getPlaylist` | `R refresh:391-428`, `client:365-385` | Optional as a set: a failed list keeps the cached playlists (`refresh:100-106`). `getPlaylist` on one playlist returns NotFound when the playlist object is absent (`client:372`). |
| `getAlbumList2` for home (`frequent`, `newest`, `recent`, `byYear` with `fromYear` and `toYear=0`, 24 each) | `R refresh:139-174`, `:474-494` | Optional: a failure keeps the cached section. |
| `getScanStatus` | freshness check before each refresh, `R refresh:175-182`, `:560-568`; poll after `startScan` for admin metadata edits, `:496-535` | Failure or a missing `lastScan` reads as "no freshness marker", which forces a full refresh every time (`R source.rs:921`, `:963`). Degrades. A usable marker needs `scanning` false and a non-empty `lastScan` (`:560-568`). |
| `getAlbum`, `getArtist`, `getSong` | on-demand refresh of one item, `R client:27-142`, `:298-333`, `:356-361` | A failure returns the error, except NotFound drops the item (`:52`). |
| `getIndexes`, `getMusicDirectory` | Folders page, `R client:180-227` | Fails the folder page for that load. |
| `getRandomSongs` (by genre), `getSimilarSongs` | radio from a seed, `R client:144-178`, `:1248-1262` | Fails that radio. |
| `star`, `unstar`, `setRating` | `R client:442-467` | Error 0 returns an error to the caller. |
| `scrobble` | `submission=false` at play start; `submission=true` with `time` in milliseconds at a qualified play, `R client:758-793` | Error 0 returns an error to the reporter. |
| `createPlaylist` (also used to replace a playlist's tracks with `playlistId` plus repeated `songId`, the reorder path), `updatePlaylist` (`name`, `public`, `songIdToAdd`, `songIndexToRemove`), `deletePlaylist` | `R client:486-593`, `:1264-1276` | Error 0 returns an error to the caller. |
| `getOpenSubsonicExtensions` | each time lyrics are read, `R client:596-610` | A failed fetch is `unwrap_or_default` (`:600`), so no extension: carries on. |
| `getLyricsBySongId`, `getLyrics` (fallback by artist and title) | `R client:611-656` | Fallback runs when `songLyrics` is not declared. |
| `stream`, `getCoverArt` | URL and byte reads, `R client:389-438`, `:469-482` | Not API envelopes: see 4, 5, and 6. |

Not called by the OpenSubsonic path: `getArtists`, `getArtistInfo`, `getAlbumInfo`, `getTopSongs`, `getInternetRadioStations`, `getPlayQueue`, `savePlayQueue`, `getNowPlaying`, `getShares`, `getLicense`, `download`, and any of the ten extensions' endpoints. A repository-wide search (`crates/`, tests excluded) found no `getTranscode`, `savePlayQueueByIndex`, `reportPlayback`, or `getPlayQueue` text.

### 2c. Amperfy v2.1.1

How error 0 reaches the caller. Amperfy parses every response as XML and records an `error` element's `code` (`A SsXmlParser.swift:44-48`). Two parse modes exist. The initial library sync parses with errors suppressed (`isThrowingErrorsAllowed: false`, for example `A Syncer:61-65`, `:90-94`, `:136-140`, `:207-211`), so an error-0 body reads as an empty list and the sync carries on silently. Every other sync (home, playlists, radios, folders, detail pages) parses the second pass strictly (`A Syncer:1377-1405`, default `true`), throws a `ResponseError`, and the caller reports it with a popup by default (`A AmperfyKit/Api/EventLogger.swift:103-140`; the popup is skipped only when `supressAlerts` is set, `:61`).

| Endpoint | Where called | What error 0 does |
|---|---|---|
| `ping` | login and server-version discovery, `A ServerApi:296-328`, `:391-425` | Valid only if no `error` element (`SsPingParserDelegate.swift:62-63`) and `version` is present and parses; else login fails. |
| `getGenres`, `getArtists` | initial sync, `A Syncer:54`, `:86` | Suppressed: reads as empty, silent. An empty `getArtists` yields a zero album count. |
| `getAlbumList2` `alphabeticalByName`, 500 per request | initial sync; the request count is `ceil(sum of artist albumCount / 500)`, with `max(1, …)`, and the loop requests indices `0…pollCount` inclusive, `A Syncer:112-161` | Suppressed: empty library. Needs `albumCount` on each artist (`A SsArtistParserDelegate.swift:56-58`) and honors `offset` and `size`. |
| `getPlaylists` | initial sync `A Syncer:197`; home section, `:949-960` | Initial: suppressed. Home: strict, popup. |
| `getPodcasts` (and `getNewestPodcasts`) | podcast support probe and sync, `A ServerApi:427-443`, `A Syncer:214-245` | The support probe returns true whenever the request does not throw (`ServerApi:436-441`), so an HTTP 200 error-0 body reads as "podcasts supported"; the sync that follows parses with errors suppressed, so a quiet empty Podcasts section results. A bare HTTP 404 becomes error 70 (`ServerApi:976-987`), the probe's request throws, and the server is read as "no podcasts" (`:436-442`). The library display default lists Podcasts (`A AmperfyKit/Storage/LibraryDisplaySettings.swift:170-183`). |
| `getAlbumList2` `newest`, `recent` | default home sections "Recently Played Albums" and "Newest Albums" (`A AmperfyKit/Storage/HomeSection.swift:37-42`), refreshed on every home load, `A Amperfy/Screens/ViewController/HomeManager.swift:200-230` | Strict: popup "Newest Albums Sync" or "Recent Albums Sync" on error 0. Quiet if the server answers an empty success. |
| `getInternetRadioStations` | Radios library entry (in the default library, `LibraryDisplaySettings.swift:170-183`), `RadiosVC.swift:142`, `:219`; home section only if the member enabled it | Strict: popup "Radios Sync" on error 0. |
| `getMusicFolders` | Directories entry (default library), refreshed on each visit, `A Amperfy/Screens/ViewController/MusicFoldersVC.swift:96-104` | The id pass is suppressed but the parse is strict (`A Syncer:788-809`): popup "Music Folders Sync". An empty list shows "No Directories" (`MusicFoldersVC.swift:73-79`). |
| `getIndexes` (with `musicFolderId`), `getMusicDirectory` | Directories drill-down, `A ServerApi:849-866`, `A Syncer:813-875` | Strict: popup. |
| `getAlbum`, `getArtist`, `getSong`, `getPlaylist` | detail pages and playlist open, `A Syncer:262-340`, `:335-460`, `:435`, `:963-1000` | Strict: popup. |
| `getStarred2` | favorites, `A Syncer:706-753` | Strict: popup. |
| `search3` | the user types a query; three calls with `artistCount`, `albumCount`, `songCount` of 40, each `query` non-empty (`A ServerApi:679-719`) | Strict: popup. Amperfy never sends an empty `query`. |
| `getRandomSongs`, `getSimilarSongs2` | user-initiated, `A ServerApi:657-666`, `:945-955` | Strict: popup. |
| `getLyricsBySongId` | only if `songLyrics` is declared, `A ServerApi:561-571`, `A Syncer:464-519` | Not called otherwise. |
| `scrobble` (now-playing as `submission=false`, and a finished play with `time` in ms), `setRating` (song, album, artist ids), `star`/`unstar` (`id`, `albumId`, `artistId`) | `A ServerApi:868-943`, `A Syncer:1123-1218` | Strict: popup on error 0. |
| `createPlaylist` (name only), `updatePlaylist` (`playlistId`, `name`, `songIdToAdd`, `songIndexToRemove`), `deletePlaylist` | `A ServerApi:739-796`, `A Syncer:1014-1085`, `:1315-1370` | Strict; creation also reads the new id back by listing (`:1351-1369`). |
| `download` | cached-file download when the cache format is "raw" (the settings default is mp3, so by default not used): `A ServerApi:345-352` | n/a under defaults. |
| `stream`, `getCoverArt` | `A ServerApi:356-389` | Not API envelopes: see 4, 5, and 6. |

Called but never answered meaningfully by a non-podcast server: `getPodcasts`, `getNewestPodcasts`, `deletePodcastEpisode`. `getAlbumInfo`, `getAlbumInfo2`, `getArtistInfo`, `getTopSongs`, `getPlayQueue`, `getShares`, and `getNowPlaying` have no caller in `SubsonicServerApi.swift` (its `forAction:` list, `:582-946`).

### 2d. Tempus v4.28.0

How error 0 reaches the caller. Tempus parses JSON with Gson (lenient) into `ApiResponse` (`T subsonic/RetrofitClient.kt:22-25`, `subsonic/base/ApiResponse.kt:9-14`), and its repositories test only whether the data container is non-null; none of the read paths below tests `status`. An error-0 body has no data container, so every read repository falls to its empty branch with no message. The exceptions are `ping` (reads `status` and shows the error text), `star` and `unstar` (read `status` and surface a refusal), and `startScan` (shows `error.message`). No client request carries `.view`.

| Endpoint | Where called | What error 0 does |
|---|---|---|
| `ping` | `onStart` and `onResume`, `T ui/activity/MainActivity.java:586-645`, `repository/SystemRepository.java:26-67` | Login: `status` failed shows "code - message" (`:39-41`). Later pings: a non-null body counts as reachable; `openSubsonic` is read from it (`MainActivity.java:617`). |
| `getOpenSubsonicExtensions` | once on launch if logged in, `MainActivity.java:163`, `:698-706`, `SystemRepository.java:102-123` | A body with no list stores nothing (`:111-113`): carries on, no extensions. |
| `getAlbumList2` | home and library lists, `T repository/AlbumRepository.java:27-56` | Soft: logs and posts an empty list. |
| `getArtists`, `getArtist`, `getArtistInfo2`, `getTopSongs`, `getStarred2` | `T repository/ArtistRepository.java:40-50`, `:137-145`, `:169-190`, `:234-262`, `:372-386` | Soft each (null-checked containers). |
| `getAlbum`, `getAlbumInfo2` | `T repository/AlbumRepository.java:107-134`, `:188-203` | Soft: empty track list; no notes. |
| `getSong` | `T repository/SongRepository.java` | Soft. |
| `search3` | `T repository/SearchingRepository.java:43-165`; sends `songCount`, `albumCount`, `artistCount` of 20 or 5 or a pager limit, plus `musicFolderId` | Soft: empty results. |
| `getGenres` | `T repository/GenreRepository.java:24-30` | Soft. |
| `getMusicFolders`, `getIndexes`, `getMusicDirectory` | `T repository/DirectoryRepository.java:25-80` | Soft: the library "folders" sector stays hidden when the list is empty or missing (`T ui/fragment/LibraryFragment.java:188-195`). |
| `getPlaylists`, `getPlaylist`, `createPlaylist`, `updatePlaylist`, `deletePlaylist` | `T repository/PlaylistRepository.java:85-` | `getPlaylists` soft (`:94-`). The cache sweep that deletes playlists missing from the list runs only on a non-null list. |
| `star`, `unstar` | `T repository/FavoriteRepository.java:55-100` | Reads `status`: failed means "refused", which strikes the local decision and reports a refusal (`:36-39`, `:60-67`). Error 0 therefore reads as a refusal. |
| `setRating`, `scrobble` (`submission=false` at start, `true` at completion or skip, `T service/BaseMediaService.kt:400`, `:531-592`) | `T repository/AlbumRepository.java:90-105`, `T service/MediaManager.java:569-572` | Ignored (empty callbacks): carries on. |
| `getPlayQueue`, `savePlayQueue` | `T repository/QueueRepository.java:63-100`, `T viewmodel/PlayerBottomSheetViewModel.java:284-296` | Soft. |
| `getRandomSongs`, `getSongsByGenre`, `getSimilarSongs`, `getSimilarSongs2` | `T subsonic/api/albumsonglist/AlbumSongListClient.java:30-45` and the radio and mix code | Soft. |
| `getLyricsBySongId` (only with `songLyrics`), `getLyrics` (fallback) | `T viewmodel/PlayerBottomSheetViewModel.java:179-181` | Soft. |
| `getShares`, `createShare`, `updateShare`, `deleteShare` | `T repository/SharingRepository.java:23-90`; `createShare` sends `id`, `description`, `expires` | `getShares` soft; `createShare` reads `shares.shares.get(0)` only behind null checks (`:51-52`). |
| `getInternetRadioStations` and the radio writes; `getPodcasts`, `getNewestPodcasts`, `refreshPodcasts`, `createPodcastChannel`, `deletePodcastChannel`, `deletePodcastEpisode`, `downloadPodcastEpisode` | `T repository/RadioRepository.java`, `PodcastRepository.java` | Soft (null-checked). |
| `startScan` | `T repository/ScanRepository.java:15-50` | Shows `error.message` (`:21-22`). |
| `stream`, `download`, `getCoverArt` | URL builders: `T util/MusicUtil.java:42-80`, `:114-146`, `T glide/CustomGlideRequest.java:84-111` | Not API envelopes: see 4, 5, and 6. |

Defined in `T`'s Retrofit services but not called from any repository (read from the method-name search across non-API code): `getLicense`, `getVideos`, `getVideoInfo`, `getAlbumList`, `getNowPlaying`, `search2`, `getStarred`, `getArtistInfo`, `getAlbumInfo`, `getScanStatus`, `getBookmarks`-style calls, and `getPodcastEpisode`.

### 2e. The empty-success candidate list this evidence produces

Spec `:670-672` asks for "the set of endpoints clients require to answer with an empty success". Read against the launch list at `:660-669` (system, browsing in both views, album lists, search, media, annotation, playlists, play queue, shares, read-only user), the endpoints roster clients call that sit outside it are the candidates. Each row gives the clients that would otherwise misbehave and what each does with an error 0. This is an input for Task 7, not a ruling.

| Endpoint | Why a candidate | Behavior with error 0 | Behavior with an empty success |
|---|---|---|---|
| `getInternetRadioStations` | Amperfy lists Radios in its default library (`LibraryDisplaySettings.swift:170-183`); Feishin has a Radio page; Tempus a radio screen | Amperfy: popup on each visit. Feishin and Tempus: quiet empty. | All four quiet. |
| `getPodcasts` | Amperfy lists Podcasts by default and probes support through it; Tempus has a podcast screen | Amperfy reads HTTP 200 as "supported", empty and silent in the initial sync, popup on later syncs. | Amperfy: supported, empty, quiet. Tempus: empty. |
| `getArtistInfo2`, `getAlbumInfo2` | Feishin and Tempus call them on detail pages | Feishin: quiet (soft, null). Tempus: quiet. | Same. These two do not need a candidate row; error 0 is already quiet. |
| `getTopSongs`, `getSimilarSongs`, `getSimilarSongs2` | Feishin and Tempus call them on artist and radio screens; Rufin and Amperfy only on a user action | Feishin: soft empty. Tempus: soft empty. Rufin: an error shown to the user on the action. Amperfy: popup on the action. | Rufin and Amperfy would see an empty list instead of an error. |
| `getMusicFolders`, `getIndexes`, `getMusicDirectory` | All four have a folder view (see 4); this is the folder-browsing question | Amperfy: popup on every Directories visit. Feishin and Rufin: the folder page fails. Tempus: sector stays hidden. | Amperfy: "No Directories". Feishin: empty picker. Rufin: unscoped catalog. Tempus: hidden sector. |
| `getScanStatus` | Rufin's freshness marker, Feishin's poll | Rufin: full refresh every time. Feishin: poll error, quiet. | Rufin needs `lastScan` non-empty and `scanning` false to cache (`R refresh:560-568`); an empty success without it behaves like error 0. |
| `jukeboxControl` | Feishin probes it on connect | Feishin: logs and carries on (`F ctrl:1483-1498`). | Same. Not a candidate. |
| `getPlayQueue`, `savePlayQueue` | Feishin restores and saves a queue; Tempus does too | Feishin: restore throws (hard deref `:1350`) and degrades. Tempus soft. | Feishin would need an empty `playQueue` object to avoid the throw. |
| `getLicense` | no roster client calls it | n/a | n/a |
| `getNowPlaying`, `getShares` (Tempus only), bookmarks | no client needs them for normal use | n/a | n/a |

Short reading for Task 7. The endpoints where an error 0 breaks or noisily interrupts a roster client are `getInternetRadioStations`, `getPodcasts`, and the folder trio on Amperfy (popup per visit), and `getPlayQueue` on Feishin. Where the server answers real data (launch list), the rest of the roster is quiet. Hard dereferences (a client throws on a missing container) fall on endpoints in the launch list anyway: `getUser`, `getAlbumList2`, `getAlbum`, `getArtist`, `getSong`, `getMusicDirectory`, `getPlaylist`, `getScanStatus`, `getPlayQueue`, `createPlaylist` (Feishin), `getTranscodeDecision` (Feishin, relevant only when `transcoding` is declared) and `ping` (all).

## 3. Q2: extensions checked

| Extension (spec name at `bed1688`) | Feishin | Rufin | Amperfy | Tempus |
|---|---|---|---|---|
| `transcoding` (endpoints `getTranscodeDecision`, `getTranscodeStream`) | Checked: `transcoding` sets `osTranscodeDecision` (`F ctrl:1460-1462`). With it, every stream URL is first resolved by a POST decision when transcoding is not explicitly on and the player is not MPV (`:1990-2043`); the body carries its direct-play profile list and a transcoding profile list (`:1996-2005`). Without it, the stream URL is built directly (`:1975-1976`, `:2044`). | Not checked. | Not checked. | Not checked. |
| `songLyrics` | Checked: v1 and v2 (`F ctrl:1464-1466`, `:2047-2085`); v2 adds `enhanced=true` | Checked on each lyrics read: v1 or higher uses `getLyricsBySongId`, v2 adds `enhanced=true` (`R client:596-620`); without it falls back to `getLyrics` | Checked by name (`A ServerApi:67-69`, `:455-559`) | Checked (`T util/OpenSubsonicExtensionsUtil.java:38-40`) |
| `formPost` | Checked: with it every request becomes a POST with `application/x-www-form-urlencoded` (`F:api.ts:543-553`) | Not checked. | Not checked. | Helper exists (`OpenSubsonicExtensionsUtil.java:34-36`); no caller found. |
| `indexBasedQueue` | Checked: switches to `getPlayQueueByIndex` and `savePlayQueueByIndex` (`F ctrl:1472-1474`, `:1325-1343`, `:2282-2296`) | Not checked. | Not checked. | Not checked. |
| `playbackReport` | Checked: sends `reportPlayback` states (`F ctrl:1476-1478`, `:2338-2371`) | Not checked. | Not checked. | Not checked. |
| `topSongsByArtistId` | Checked: `getTopSongs` by `id` instead of `artist` name (`F ctrl:1480-1482`, `:2092-2096`) | Not checked. | Not checked. | Not checked. |
| `transcodeOffset` | Not checked (`timeOffset` has no Subsonic-path use; the only text hit is lyrics timing). | Not checked. | Not checked. | Checked: a transcoded stream (`format` present and not "raw") plus this extension selects a custom media source that seeks by new requests with `timeOffset` (`T util/DynamicMediaSourceFactory.kt:59-66`, `T util/OpenSubsonicExtensionsUtil.java:30-32`, URL at `T util/MusicUtil.java:72-73`). |
| `apiKeyAuthentication` | Not checked. | Offered as a login mode the member picks, not discovered: sends `apiKey`, no `u`, and verifies with `tokenInfo` (`R mod:517-530`, `R client:1006-1013`); the spec says `u` with an API key is error 43 (`apiKeyAuth.md`) | Not checked. | Not checked. |
| `sonicSimilarity`, `getPodcastEpisode` | Not checked. | Not checked. | Not checked. | Not checked. |

What the extension list does for each client without it. Feishin: no transcoding decision, so a plain `stream` URL with whatever `format` and `bitrate` its manual transcode setting (default off) appends (`F` `src/renderer/store/settings.store.ts:2130-2132`, `F ctrl:234-245`, `:1978-1982`); no lyrics, queue-by-index, or playback report; GET requests. Tempus without `transcodeOffset`: the default media source with metadata duration; a seek in a transcoded stream is then an ordinary range request (`DynamicMediaSourceFactory.kt:64-65`). Rufin and Amperfy: only lyrics change.

The extension-discovery channel. Every client sends credentials with `getOpenSubsonicExtensions`: Feishin through the shared auth params (`F:api.ts:499-511`), Rufin through `get_json` (`R client:1232-1241`), Amperfy through `createAuthApiUrlComponent` (`A ServerApi:445-453`), Tempus through `getParams()` (`T subsonic/api/system/SystemService.java:18-19`). No roster client calls it unauthenticated, so the unauthenticated carve-out (spec `:640-673`) is not exercised by this roster.

The `transcoding` extension, the question above all: in the roster only Feishin calls it, and its profile lists have a shape worth recording for the AAC item (see 4). Its transcoding profiles name flac, opus in ogg, and mp3 (and only mp3 on Safari), with protocol `http` (`F` `src/renderer/features/player/components/audio-players.tsx:73-79`), and its direct-play list is whatever the browser's `canPlayType` accepts from a probe list that includes `audio/mp4; codecs="mp4a.40.2"`, `audio/aac`, `audio/x-m4a`, ogg opus, webm opus, ogg vorbis, flac, wav, and alac (`:53-71`, `:96-112`). Feishin sets `maxAudioBitrate` and `maxTranscodingAudioBitrate` to 0 and an empty `codecProfiles` (`F ctrl:1996-2005`). The decision response it reads is `canDirectPlay`, `transcodeReason`, and `transcodeParams` (`:2020-2033`); the stream URL it then builds is `getTranscodeStream` with `mediaId`, `mediaType`, `offset=0`, and `transcodeParams` (`:247-266`, `:2035-2042`).

## 4. Q3: ID3, folder, `musicFolderId`

All four roster clients offer a folder view next to the ID3 one, and all four use the same three calls for it: `getMusicFolders`, `getIndexes` with `musicFolderId`, and `getMusicDirectory`.

| | Feishin | Rufin | Amperfy | Tempus |
|---|---|---|---|---|
| ID3 browse | `getArtists`, `getArtist`, `getAlbum`, `getAlbumList2`, `search3` | catalog from `getAlbumList2` and `search3`; `getArtist` and `getAlbum` on demand | `getArtists`, `getArtist`, `getAlbum`, `getAlbumList2`, `search3` | `getArtists`, `getArtist`, `getAlbum`, `getAlbumList2`, `search3` |
| Folder browse | "Folders" sidebar item enabled by default (`F` `src/renderer/store/settings.store.ts:1217-1221`) and a route (`F` `src/renderer/router/app-router.tsx:296-298`). `getIndexes` for the root (the page uses id "0" as its root token), `getMusicDirectory` below (`F ctrl:1037-1135`). | "Folders" route and sidebar item (`R` `crates/gtk-widgets/src/route.rs:109`); `getIndexes` for the root and `getMusicDirectory` below (`R client:180-227`). | "Directories" is in the default library list (`A AmperfyKit/Storage/LibraryDisplaySettings.swift:170-183`): `getMusicFolders` then `getIndexes` (with the folder id) then `getMusicDirectory` (`A Syncer:788-875`). | A "folders" sector in the Library tab, shown when `getMusicFolders` returns a non-empty list (`T ui/fragment/LibraryFragment.java:188-195`); `getIndexes` with the folder id, `getMusicDirectory` (`T repository/DirectoryRepository.java:44-80`). |
| Passes `musicFolderId` | Only when the member selects a folder (a list-filter value or a server-level choice, `F` `src/renderer/features/shared/hooks/use-music-folder-id-filter.ts:10-20`, `F` `src/renderer/api/utils-music-folder.ts:3-23`); the value is the first selected id. Endpoints: `getArtists` (`F ctrl:538`, `:914`), `search3` (`:602`, `:758`, `:1554`, `:1708`, `:1749`, `:1923`, `:1952`, `:2388`), `getStarred` (`:662`, `:810`, `:1601`, `:1830`), `getAlbumList2` (`:722`, `:856`), `getIndexes` (`:1045`), `getRandomSongs` (`:1369`), `getSongsByGenre` (`:1580`, `:1782`, `:1807`). | By default on every catalog pass: once per folder returned by `getMusicFolders`, on `search3` for tracks only (`R refresh:304-325`); unscoped when the folder list is empty or the call failed. Also on `getIndexes` when a music folder is chosen (`R client:208-211`). Not on `getAlbumList2`, artist `search3`, or `getGenres`. | Only on `getIndexes` (`A ServerApi:849-853`). Never on `search3`, `getAlbumList2`, `getArtists`, or `getRandomSongs`. | The member's active folder, default none, on `getArtists`, `getAlbumList2`, `getRandomSongs` (both forms), `getSongsByGenre`, and `search3` (`T subsonic/api/browsing/BrowsingClient.java:45`, `T subsonic/api/albumsonglist/AlbumSongListClient.java:30-45`, `T subsonic/api/searching/SearchingClient.java:30`); the stored value is null for "all", which Retrofit drops (`T util/Preferences.kt:160-170`). `getIndexes` takes the folder id explicitly. |

Reading for the folder-browsing item (`:925-926`). Not shipping folder browsing means answering `getMusicFolders` empty (or error 0), `getIndexes`, and `getMusicDirectory` without data. The effect per client with an empty `getMusicFolders` list: Tempus hides the sector; Feishin hides its folder picker (`F` `src/renderer/features/sidebar/components/server-selector-items.tsx:225`) but its Folders sidebar item stays and the page then fails or is empty (hard `getMusicDirectory` deref, `F ctrl:1098`; the sidebar item is on by default); Rufin's catalog runs unscoped (the whole library, `R refresh:304-309`) but its Folders route stays; Amperfy shows a "No Directories" screen it lists by default. None of this breaks library use. With folders served, Rufin's per-folder `search3` makes `musicFolderId` on `search3` load-bearing: each pass must return only that folder's songs, since Rufin then links each track to the folder it came from (`R refresh:337-350`). One more demand: Feishin and Tempus both call `search3` with `musicFolderId` on every search once a folder is selected, so the id must be honored there to keep a chosen folder consistent.

## 5. Q4: stream formats, containers, request shape

What each client asks for by default, from its URL builder and its settings defaults.

| | Feishin | Rufin | Amperfy | Tempus |
|---|---|---|---|---|
| Default `stream` request | `stream.view?id&v&c&u&t&s` (credentials appended from the stored credential string); no `format`, no `maxBitRate` while transcode is off (default off) (`F ctrl:1973-1976`, `F` `settings.store.ts:2130-2132`) | `format=raw` when no cap; `format=mp3&maxBitRate=<kbps>` when capped (`R client:389-400`, `:418-438`) | `format=mp3` (the default format preference) and no `maxBitRate` (default no limit): the default request is already a transcode (`A AmperfyKit/Storage/SettingEnumerations.swift:123-128`, `:95-110`; URL at `A ServerApi:356-380`) | `maxBitRate=0&format=raw` on Wi-Fi and mobile (defaults `raw`, bitrate "0"), plus `id` last; both omitted only when the member picks "server prioritized" (`T util/MusicUtil.java:42-80`, `T util/Preferences.kt:495-511`, `:724-726`) |
| Format values the member can choose | free text for `format`, a number for bitrate (`F` `src/renderer/features/settings/components/playback/transcode-settings.tsx:66-71`) | mp3 for any cap, no choice (and opus for downloads when the server type is typed Navidrome, `R client:407-413`) | mp3, raw, or server decides (omit `format`) (`SettingEnumerations.swift:123-128`) | raw, opus, aac, mp3, flac (`T app/src/main/res/values/arrays.xml:153-195`) |
| Transcoded download | n/a | `format=raw`, or mp3 | cache format default mp3: `stream&format=mp3`, not `download` (`A ServerApi:330-354`, `SettingEnumerations.swift:345-350`) | `download` by default; transcoded download via `stream` with the stored format and bitrate (`T util/MusicUtil.java:114-179`) |
| Seeking | the browser's media element issues byte ranges | GStreamer `playbin3` with accurate seeks (`R` `crates/playback-gstreamer/src/pipeline.rs:640-668`) | See below: a `Range: bytes=0-1` probe, then ranged requests | ExoPlayer ranges; with `transcodeOffset` and a transcoded stream, a new request with `timeOffset` (`T util/MusicUtil.java:72-73`) |

The player library, read for what containers it takes.

- Feishin plays through the browser's media element (react-player, `crossOrigin="anonymous"`, `F` `src/renderer/features/player/audio-player/engine/web-player-engine.tsx:327-330`; the default player type is web, `settings.store.ts:2133`). Accepted containers are whatever the engine's `canPlayType` returns for its probe list; the probe covers MP3, AAC in MP4, ADTS AAC (`audio/aac`), M4A, Opus and Vorbis in Ogg and WebM, FLAC, WAV, and ALAC in MP4 (`F` `audio-players.tsx:53-71`). The probe says which containers Feishin will declare as direct-playable to a `transcoding` server, not which an engine decodes: that depends on the engine (Chromium in the Electron build). Fragmented MP4 versus ADTS is not distinguished at all by Feishin's probe; both would be named `aac` in container `mp4` or `aac`.
- Rufin plays through GStreamer `playbin3` (`R` `crates/playback-gstreamer/src/pipeline.rs:757-778`) and relies on typefind; the repository does not list containers. GStreamer's own `aacparse` documents ADTS (and ADIF, LOAS, raw) as accepted and producible streams (`https://gstreamer.freedesktop.org/documentation/audioparsers/aacparse.html`, read 2026-10-07: `stream-format` of `raw`, `adts`, `adif`, `loas`). The `qtdemux` page states only that it "supports both push and pull-based scheduling" and says nothing about fragmented MP4 (`https://gstreamer.freedesktop.org/documentation/isomp4/qtdemux.html`, read 2026-10-07). So ADTS is documented for Rufin and fragmented MP4 is unconfirmed.
- Amperfy plays through `AudioStreaming` v1.4.4 (`A AmperfyKit/Player/BackendAudioPlayer.swift:22-24`, `:640-662`; `Package.resolved` revision `a729ee2`). Its README (`dimitris-c/AudioStreaming@a729ee2:README.md:7-11`) lists "AIFF, AIFC, WAVE, CAF, NeXT, ADTS, MPEG Audio Layer 3, AAC audio formats", "M4A (optimized and non-optimized) from v1.2.0", and Ogg Vorbis with no seeking. It chooses the container from the response `Content-Type`, mapped through a fixed table: `audio/mpeg` and `audio/mp3` to MP3, `audio/aac` and `audio/aacp` to ADTS, `audio/mp4`, `audio/x-m4a`, `audio/m4a` to MP4 or M4A, `audio/flac`, `audio/ogg`, and others (`dimitris-c/AudioStreaming@a729ee2:AudioStreaming/Streaming/Helpers/AudioFileType.swift:13-42`, parse at `AudioStreaming/Streaming/Parsers/HTTPHeaderParser.swift:50-51`); an unknown type falls back to the URL's path extension (`RemoteAudioSource.swift:45-50`), and a stream URL ends in `stream.view`, which names no type. Further demands from the same library: the header parser returns nothing unless the response has more than two headers (`HTTPHeaderParser.swift:47`); length comes from `Content-Length` on 200 or from the total in `Content-Range` on 206 (`:55-69`); a stream is seekable only on status 206 or when `Accept-Ranges` is not "none" (`:82`, `RemoteAudioSource.swift:340-346`); an MP4 is read as `[ftyp][moov][mdat]` and, if the media data comes before the movie header, the library refuses unless the server answers ranges (error `nonOptimizedMp4AndServerCannotSeek`, `…/Mp4/RemoteMp4Restructure.swift:58-92`). Fragmented MP4 is not named anywhere in the library (a search for "moof", "fragment", and "fmp4" over its sources and README found nothing), so for Amperfy ADTS is documented and fragmented MP4 is unconfirmed.
  - A per-track probe. Before it opens each track, `AudioStreaming` sends a request with `Range: bytes=0-1`, `Accept: */*`, `Icy-MetaData: 1`, and `Accept-Encoding: identity`, reads the headers, and cancels it (`RemoteAudioSource.swift:223-236`, `:389-403`); the real request follows with the same headers and a `Range` only when seeking (`:368-387`). On a transcode this means one request opens a cache entry and is dropped before a second one opens the stream.
  - Amperfy also marks `audio/x-ms-wma` and the unknown-type string as not playable and rewrites `audio/x-flac` to `audio/flac` and `audio/m4a` to `audio/mp4` when it hands a type to the player (`A AmperfyKit/Storage/CacheFileManager.swift:35-47`); that is about the type Amperfy parsed from `contentType`, not from the stream.
- Tempus plays through Media3 ExoPlayer 1.9.2 with `DefaultExtractorsFactory` (`T gradle/libs.versions.toml:21`, `T util/DynamicMediaSourceFactory.kt:56-57`). That factory's own documentation lists MP4 and M4A, fMP4, Matroska and WebM, Ogg Vorbis and FLAC, MP3, AAC (`AdtsExtractor`), MPEG TS and PS, FLV, WAV, AC3, AC4, AMR, and FLAC (`androidx/media@1.9.2:libraries/extractor/src/main/java/androidx/media3/extractor/DefaultExtractorsFactory.java:66-90`, Apache-2.0). Both ADTS and fragmented MP4 are therefore documented for Tempus.

The AAC container item (`:427-431`, `:946`), what the roster evidence says. Of four players, Tempus documents both ADTS and fragmented MP4 (Media3), Rufin documents ADTS and leaves fragmented MP4 unconfirmed (GStreamer), Amperfy documents ADTS and is silent on fragmented MP4 (AudioStreaming), and Feishin's engine is the browser's (no container distinction, no AAC in its own transcoding profile list). ADTS is documented for all three native players; fragmented MP4 for one. So the evidence favors ADTS, with the content type `audio/aac` (the exact string Amperfy maps to ADTS). That is a reading for Task 7, with one gap named: no roster client has been observed playing either form, so the choice carries a test owed (play an ADTS stream and a fragmented-MP4 stream, with a growing length, in each reachable client; Amperfy and Tempus cannot be driven on this workstation).

A separate default that matters more for the streaming slice than the AAC choice: Amperfy's out-of-the-box request is `format=mp3`, and Tempus's and Rufin's is `format=raw`. A server that answers `format=mp3` with a transcode and `format=raw` with the file serves all four defaults. `format=raw` meaning "disable transcoding" is stated in the spec (`opensubsonic/open-subsonic-api@bed1688:content/en/docs/Endpoints/stream.md`, parameter `format`: "Starting with 1.9.0 you can use the special value \"raw\" to disable transcoding").

`estimateContentLength` is sent by no roster client (a search over all four trees finds none), so the oracle's estimated-length rule (spec `:430-445`) is not exercised by this roster. `timeOffset` is sent only by Tempus (see 3).

## 6. Q5: CORS and the credential channels

### Browser context

Only Feishin runs in a browser engine. Its renderer is Chromium in the Electron build (and a browser tab in the web build). Rufin (Rust, `reqwest`), Amperfy (Alamofire and `URLSession`), and Tempus (OkHttp and Glide) make native HTTP calls with no CORS.

Feishin's browser-side requests and what each needs of the server:

| Request | How it is made | CORS consequence |
|---|---|---|
| API GETs | axios on the XHR adapter; `Content-Type` is stripped on a request without a body (`axios/axios@v1.17.0:lib/adapters/xhr.js:155-156`), and Feishin's only base header is `Content-Type: application/json` (`F:api.ts:613-615`) | No preflight. The response needs `Access-Control-Allow-Origin`. |
| `formPost` POSTs | `application/x-www-form-urlencoded` (`F:api.ts:543-553`) | A simple request, no preflight, only when the server declares `formPost`. |
| `getTranscodeDecision` | POST with `application/json` (`F:api.ts:524-542`) | Preflight: `OPTIONS` with `Access-Control-Request-Method: POST` and `Access-Control-Request-Headers: content-type`, answered with CORS headers and no data. Only when the server declares `transcoding`. |
| Cover art | `fetch(url)` with no headers and no `credentials` option, then a blob (`F` `src/shared/components/image/use-native-image.ts:93-115`) | CORS-mode fetch: `getCoverArt` responses need `Access-Control-Allow-Origin`, or the image fails. |
| Audio | an `<audio>` element created with `crossOrigin="anonymous"` (`F` `web-player-engine.tsx:327-330`, `:353-356`) | The `stream` response needs `Access-Control-Allow-Origin`; with the Web Audio graph on by default (`settings.store.ts:2134`, `createMediaElementSource` at `web-player.tsx:562`, `:583`) a response without it would be silenced or blocked. |

Credentials: none of the browser requests is credentialed. axios sets no `withCredentials`, `fetch` uses the default (same-origin) credentials mode, and the audio element is `anonymous`. Feishin therefore never needs `Access-Control-Allow-Credentials`, and no cookie crosses. The Electron build turns the browser's checks off only if the member flips "Ignore CORS" on the add-server form (default off, `F` `src/renderer/features/servers/components/ignore-cors-ssl-switches.tsx:14-31`; the switch only renders under Electron, `:33-35`), which sets `webSecurity: false` (`F src/main/index.ts:675`); so CORS applies to the desktop build by default too, not only to a hosted web build. The allowed origin is the origin of Feishin's own page, which dubplate cannot know ahead of time (a hosted instance, `app://`-style Electron origin, or `localhost`). That is the open policy question (`:650-655`); what this evidence adds is that it can be a wildcard without credentials, since no roster request is credentialed.

Headers a browser client reads from a response: none are required by the source (Feishin reads `res.headers` only as a passthrough, `F:api.ts:583`). `Content-Range` and `Accept-Ranges` are read by the media element itself and need no `Access-Control-Expose-Headers` for the element's own use; whether a CORS-mode media request needs the server to expose them is unconfirmed here (it needs a browser run against a server; Task 3 and the harness are the places to prove it).

### Credential channels, by client

| | Query string | Header | Body |
|---|---|---|---|
| Feishin | `u` plus `s` and `t`, or `p` (legacy), on every API call and every stream and cover URL (`F:api.ts:499-511`, `:555-561`; stream and cover URLs append the stored credential string, `F ctrl:1976`, `:67-74`). `formPost`: moves them into the form body | None | With `formPost`: the same keys in a urlencoded body (`F:api.ts:543-553`). The transcode decision POST keeps credentials in the query string and the capability document in the JSON body (`:527-542`). |
| Rufin | `u`, `s`, `t` or `u`, `p` or `apiKey`, plus `v`, `c`, `f=json` (`R client:1001-1016`); URLs are redacted when logged (`:1144-1161`) | None | None |
| Amperfy | `u`, `t` and `s` (token) or `u` and `p` (legacy), `v`, `c` (`A ServerApi:227-238`); no `f` | An optional set of custom headers from the account (`A ServerApi:254-262`), sent on API calls, on playback (`A BackendAudioPlayer.swift:581`, `:651-654`), and on downloads (`A AmperfyKit/Download/DownloadManagerSessionExtension.swift:30-36`) | None |
| Tempus | `u` plus `p` or `s` and `t`, plus `v`, `c` (and `f=json` on API calls only), built by string concatenation for stream, download, and cover URLs (`T util/MusicUtil.java:42-61`, `:114-136`, `T glide/CustomGlideRequest.java:84-104`) and by `@QueryMap` on API calls (`T subsonic/Subsonic.java:145-163`) | None | None |

The channel set, for the privileged-input question: the query string is the only channel every client uses; no roster client uses an `Authorization` header or a cookie. A custom-header channel exists in Amperfy only, and it is meant for a reverse proxy rather than for the Subsonic credentials. A `u` and `apiKey` conflict is error 43 in the spec; Rufin's API-key mode sends `apiKey` and no `u`, as required.

The carve-out. As recorded in 3, every roster client sends credentials with `getOpenSubsonicExtensions`, and Tempus's `ping` carries them too. None of the four needs `getOpenSubsonicExtensions` to answer without credentials.

## 7. Q6: `Accept: image/webp` on `getCoverArt`

| Client | What the source sets | Result |
|---|---|---|
| Feishin | Loads covers with `fetch(url, { credentials, headers })` where both are undefined for a Subsonic server (`F` `src/shared/components/image/use-native-image.ts:93-103`, request built at `F ctrl:48-75` with no headers). The only `webp` text in the app is a file-extension map and a file-picker filter. | Observed behavior of the engine, not of Feishin: a headless Chromium 154 `fetch()` to my local server sent `Accept: */*`, `Sec-Fetch-Mode: cors`; an `<img>` tag on the same page sent `Accept: image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8` and `Sec-Fetch-Mode: no-cors`. Feishin uses `fetch`, so it should send `*/*`, not `image/webp`. Electron 43's Chromium was not the version observed: unconfirmed for the shipped build, to be captured by Task 3's live walk. |
| Rufin | `image_bytes` builds a plain `GET` through `reqwest` and sets no `Accept` (`R client:469-482`, `subsonic_bytes` at `:1074-1084`); the only `Accept` headers anywhere in `crates/sources` belong to the Jellyfin and Emby path (`jellyfin_emby/client.rs:669`, which does list `image/webp`), the Plex lyrics path, a discovery request, and Navidrome's event stream (`R navidrome/events.rs:17`). | No `image/webp` on the OpenSubsonic path in source. The `reqwest` default `Accept` is unconfirmed; Task 3's live walk can capture it. |
| Amperfy | Artwork goes through `URLSession` downloads that set only the custom headers (`A AmperfyKit/Download/DownloadManagerSessionExtension.swift:28-36`); the library `AudioStreaming` is for audio only. | No `image/webp` in source. The `URLSession` default `Accept` for a download task is unconfirmed and cannot be captured here (no iOS device). Route to rung 16's attended gate. |
| Tempus | Glide loads the cover URL through a custom `StringLoader` (`T glide/CustomGlideModule.java:25-28`, `IPv6StringLoader`); the app sets no `Accept` header (a search for `"Accept"` and `webp` in `T` found only a `.webp` extension check at `T util/MusicUtil.java:575`). | No `image/webp` in source. Glide's default connection `Accept` is unconfirmed and cannot be captured here (no emulator). Route to rung 16's attended gate. |

So no roster client is evidenced to send `Accept: image/webp` on `getCoverArt`. Feishin, the one browser-engine client, uses a path that Chromium answers with `*/*`; the other three set nothing. The WebP-when-accepted rule (spec art slice, survey `:267-268`) would not be triggered by Feishin by default; whether the other three do depends on three platform defaults this lane could not observe.

## 8. Q7: secondary inputs

Symfonium (`https://docs.symfonium.app/wiki/providers/supported-features-by-media-providers/` and `https://docs.symfonium.app/wiki/providers/subsonic-opensubsonic-media-provider-configuration/`, read 2026-10-07; proprietary, so only short quotes). The features page says OpenSubsonic support "varies by server and the extensions it implements" and that an OpenSubsonic connection "does not guarantee every field in the table will be supplied"; transcoding "depends on the server and its configuration"; it names nothing per extension. The configuration page is more useful: a "Compatibility mode" is "required for servers lacking search3 endpoint support with empty queries" (so Symfonium's catalog sync reads the library with an empty-query `search3`, as Rufin and Feishin do); a "Do not use download endpoint" option covers servers without `download`; a "Transcode to MP3 instead of OPUS" option exists for servers without Opus; and "Ignore server transcoding settings" makes Symfonium "request the raw audio files from the server when there is no need for local transcoding". A forum thread indexed by the search tool says Symfonium controls transcoding client side on non-OpenSubsonic servers (`https://support.symfonium.app/t/subsonic-opensubsonic-media-provider-configuration/160` was returned by the search; its text was not read). Symfonium is not on the roster, so none of this changes a measured demand; it corroborates the empty-query `search3` and `download` demands.

Melodee's compatibility matrix (`melodee-project/melodee@d91342e` (v2.2.0), `docs/pages/opensubsonic-matrix.md`, MIT, "Last updated: January 13, 2026", "Applies to: Melodee 1.8.0+", so it predates the pinned release by two minor versions). It is an endpoint-by-endpoint support table with a "Test Coverage" column (schema validation or manual) and a client-notes section. It names Symfonium, Feishin, DSub and Ultrasonic, and Supersonic, each marked "Works" on login, browse, stream, playlists, search, favorites, and scrobble, "Verified By: Community" at version 1.8.0 on 2026-01-12 (`:220-287`). It lists no Rufin, Amperfy, or Tempus entry, so it gives no evidence for three of the four roster clients and only a generic pass for Feishin with no version. It says of itself that a "Supported" mark does not guarantee "error handling matches specification" (`:36-46`). It lists the extensions `apiKeyAuthentication`, `formPost`, `songLyrics`, `transcodeOffset`, and a Melodee-specific one (`:176-186`); the spec names `apiKeyAuth.md`'s extension `apiKeyAuthentication`, so the names agree. Streaming and transcoding are marked "Implemented, not yet verified with automated tests" (`:196-197`). Read as secondary evidence, it adds nothing the source readings above do not.

## 9. Clean-room and rule checks

- Rule 1 (GPL cited, never reproduced). Every client and server behavior above cites `repo@commit:path:line` and is stated in own words. Identifier names that appear (endpoint names, parameter names, field names, extension names) are the OpenSubsonic wire vocabulary the specification itself defines. No fenced block in this record holds copyleft text. The player libraries are MIT (`AudioStreaming`) and Apache-2.0 (`androidx/media`); the short quotes from their README and class documentation are within rule 2.
- Rule 2 (permissive and documentation sources quoted briefly with citation): the OpenSubsonic specification, the two player-library documents, GStreamer documentation pages, and Symfonium's public pages are quoted as short phrases.
- Rule 3 (observation beats reading): one local observation was made (the browser `Accept` header), with the browser version and server fixture state given in 7. All other items are source readings, stated as such.
- Rule 4 (every claim carries its source; unconfirmed marked): the "unconfirmed" markers in 5, 6, and 7 each name what would confirm them.
- Planning-miss check 3 (privileged input channels): named in 6 for all four clients, and the extension-discovery carve-out is named in 3 and 6.
- Planning-miss check 4 (fixture state): the one observation (7) was a static one-pixel PNG served over a loopback HTTP server with the page loaded once in a clean profile; the request logged three fetches (document, `<img>`, `fetch()`) plus a favicon.
- Prior art consulted: Symfonium notes (adopted as a study input, spec `:885`): read, short and thin, see 8. Feishin's type schemas (study for required fields, survey `:165`): read, see 1, and they are Navidrome-shaped types the runtime does not enforce. The GPL Go Subsonic clients stay declined (survey `:172`) and none was used as a probe. Melodee's matrix: read as a secondary input, see 8; its code was not reused. Neutral on hooks: nothing here is built, and no hook is landed or created.

## Question outcomes

| # | Question | State |
|---|---|---|
| 1 | Endpoints called and what error 0 does | Answered from source for all four clients (2a to 2d), with the candidate list in 2e. Runtime effects of the Amperfy popups and Tempus's empty branches are source-read; a live proof against dubplate's own server is **test owed** for Amperfy and Tempus (slice 2's harness can cover the JSON shape; the app popups need rung 16's attended gate). |
| 2 | Extensions checked | Answered (3). |
| 3 | ID3, folder, `musicFolderId` | Answered (4). |
| 4 | Stream formats and containers | **Partly answered.** Defaults and request shapes are answered (5); containers: ADTS documented for Tempus, Rufin, and Amperfy, fragmented MP4 documented only for Tempus. Which of the two dubplate ships is Task 7's reading; **test owed**: play both forms in Feishin and Rufin (Task 3, desktop) and in Amperfy and Tempus (rung 16, no iOS or Android here). |
| 5 | CORS and credential channels | Answered (6). One sub-point is **test owed**: whether a CORS-mode media request needs `Access-Control-Expose-Headers` for `Content-Range` (browser run against a server). |
| 6 | `Accept: image/webp` | Answered for Feishin by engine behavior (not the shipped Electron build: capture owed to Task 3) and by absence in source for the others. **Routed** to Task 3 (Feishin and Rufin captures) and to rung 16 (Amperfy, Tempus). |
| 7 | Secondary inputs | Answered (8). |

## Routed

- To Task 7 (synthesis): the empty-success candidate list in 2e; the ID3 and folder evidence for the folder-browsing item (4); the ADTS-over-fragmented-MP4 reading with its gap (5); the extension subset input: only Feishin calls `transcoding`, and only Tempus calls `transcodeOffset`, with `songLyrics` checked by all four and the other extensions by Feishin alone (3); the CORS policy input: wildcard origin is enough, no credentials ever cross (6); Amperfy's default `format=mp3` against Tempus's and Rufin's `format=raw` (5).
- To Task 3 (live walks): capture the `Accept` header of Feishin's and Rufin's `getCoverArt` requests; capture whether Feishin's web player sends a preflight on `stream`, and with which `Range` headers; capture Rufin's `ping`, `getUser`, and catalog request order against Navidrome at `/dubplate`; confirm Feishin's `stream.view` URL uses the same base as its API calls at a sub-path (`F ctrl:1976` uses `server.url` raw while `F:api.ts:500-501` uses `getServerUrl`).
- To slice 3 (browse, `musicFolderId`): `search3` with an empty `query` is the Tracks, catalog, and artists enumeration for Feishin and Rufin (`F ctrl:1702-1712`, `R refresh:314-326`, `:441-452`) and is Symfonium's compatibility switch; `getArtists` must carry `albumCount` for Amperfy's paging; `getAlbumList2` must honor `offset` and `size`; `search3` must honor `musicFolderId` once folders are served; both `/rest/x` and `/rest/x.view` must answer (Tempus is bare, the others `.view`); XML must be the default response form because Amperfy sends no `f`.
- To slice 4 (art): no roster evidence of `Accept: image/webp`; Feishin loads covers by `fetch` and needs CORS headers on `getCoverArt` in its browser contexts; Amperfy treats a file under 2,000 bytes as a possible error body and parses it (`A SubsonicArtworkDownloadDelegate.swift:78-90`, size limit `:29`), so a tiny valid image is handled only if it is not XML.
- To slice 5 (streaming, transcode): the AAC container evidence (5); `format=raw` and `maxBitRate=0` are Tempus's default stream parameters and must read as "no transcode"; `format=mp3` with no cap is Amperfy's default; Amperfy opens a `Range: bytes=0-1` probe per track before the real request, and needs `Content-Type`, `Content-Length` (or `Content-Range`), and more than two response headers; its seekability needs status 206; `transcodeOffset` is used by Tempus; `getTranscodeStream` is used by Feishin only when `transcoding` is declared, and a server declaring `transcoding` must never answer `getTranscodeDecision` with error 0 (Feishin's unguarded `td.transcodeReason` read, `F ctrl:2028`, makes the track unplayable); an `Accept-Ranges` value other than "none" is treated as seekable by Amperfy.
- To slice 6 (identity): API-key login exists in Rufin as a member-chosen mode (`apiKey`, `tokenInfo`) and is not discovered from the extension list; the three other clients use the salted-token or legacy form only; no roster client uses a header or cookie for credentials.
- To rung 16 (attended gate): Amperfy and Tempus cannot be driven on this workstation. Their popup and empty-branch behavior, the cover-art `Accept`, and playback of both AAC forms wait for the attended roster walk. The Android emulator is blocked by the missing hardware virtualization recorded in Task 0.

## Could not do

- No live walk of any client (Task 3 owns them), so no request captured from a roster client itself. The one observation was of a browser engine.
- No Android or iOS execution, so Tempus and Amperfy are read from source only.
- The `https://support.symfonium.app/t/subsonic-opensubsonic-media-provider-configuration/160` thread was returned by a search but not read.
- Whether Feishin's Electron 43 and the `reqwest`, `URLSession`, and Glide defaults send `Accept: image/webp` is unconfirmed for the shipped builds.
- A GStreamer-side statement on fragmented MP4 was not found in the pages read (`qtdemux` is silent on it).
