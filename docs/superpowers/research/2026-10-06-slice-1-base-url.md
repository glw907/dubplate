# Slice 1, Task 3: the API base URL against a sub-path

Date read: 2026-10-07 (UTC, 04:50 to 05:12). Staged outside the repo; lands at A-land as
`docs/superpowers/research/2026-10-06-slice-1-base-url.md`.

**Pins read** (from `2026-10-06-slice-1-evidence-pins.md`): Feishin v1.17.0 `5c169ac`, Rufin v0.16.6 `aad2f44`, Amperfy
v2.1.1 `c165a7a`, Tempus v4.28.0 `687d97f` (each clone's `HEAD` was checked against its pin after cloning); Navidrome
0.64.2 (`sha256:38dc2727...ee52`), gonic v0.22.0 (`sha256:516fd964...1d49`); `navidrome/website` `b0fe572`.
Clean-room rule 1 holds: every claim about a copyleft client or server is a `repo@commit:path:line` citation or an
observation, stated in this lane's own words. Nothing from a GPL repository is reproduced. The two quoted lines in
"What the servers document" are documentation text (Apache-2.0 website table; a README row quoted as Task 0 did).

**Questions this record owns.** (1) Does each roster client work against a server URL that carries a path
(`http://127.0.0.1:<port>/dubplate`), with the verdict one of works, breaks, or unmeasured. (2) How does each join its
configured URL with `/rest/<endpoint>`, with and without a trailing slash and the `.view` suffix. (3) Which second
code paths build URLs (stream, download, cover art, share). (4) Does any client tracker carry a sub-path or
reverse-proxy base URL issue. (5) What do Navidrome and gonic document for a sub-path. This task measures and
recommends nothing binding: slice 3 decides, with the build-spec erratum (re-plan `:427`).

## Verdicts

| Client (pin) | Verdict | Measured how |
|---|---|---|
| Feishin v1.17.0 | **works** | Live walk under Flatpak 1.17.0, isolated profile, against Navidrome 0.64.2 at `/dubplate`, added as the OpenSubsonic server type. Login, album browse, play, cover art all succeeded; every one of the 57 Feishin requests the oracle logged sat under `/dubplate/rest/`. A second server entry with a trailing slash also worked. Source read for the paths the walk did not exercise. |
| Rufin v0.16.6 | **works** | Live walk under Flatpak 0.16.6, isolated profile, against Navidrome at `/dubplate`, added as the OpenSubsonic source kind. Login, library sync, album view, play of both tracks, cover art all succeeded; all 30 requests sat under `/dubplate/rest/`. Source read for the trailing-slash and `/rest` handling. |
| Amperfy v2.1.1 | **works by source, with one precondition; unmeasured live** | Source only. It cannot run on Linux; the re-plan walks it at rung 16. Every API, stream, download, and cover-art URL is built from the stored server URL plus `rest/<action>.view`, so a path survives. Its login first sends one GET to the configured URL itself and rejects any status of 400 or more except 401 (below), so the sub-path root must answer under 400. |
| Tempus v4.28.0 | **works by source; unmeasured live: no emulator** | Source only. This workstation has no hardware virtualization (Task 0: no `vmx` flag, no `/dev/kvm`), so no Android emulator exists and no Tempus APK is installed. Every request, stream, download, and cover URL derives from `<server url>/rest/` plus a relative method name with no `.view` suffix. |

"Works by source" is a reading, not a proof: it states what the code does with a path, and the live gap is Amperfy's
walk at rung 16 and Tempus's once VT-x is enabled and an emulator exists.

## Fixture state and oracle

Navidrome 0.64.2 (`docker.io/deluan/navidrome@sha256:38dc2727...ee52`), started from Task 0's `run-oracle.sh` with
`ND_BASEURL=/dubplate`, `ND_ADDRESS=127.0.0.1`, port 47130, `ND_LOGLEVEL=debug`, its own data directory, the Task 0
fixture mounted read-only. State: scanned (5 tracks, 4 albums: Carrier Wave, Long Form, Signal One, Signal Two),
a throwaway admin user `probe` created by `POST /dubplate/auth/createAdmin`, every client logged in with a
token or with the salted-token form its own login produced. The second pass (headers, below) used a fresh instance of the same recipe and a logging proxy on 47133. The root-mounted control instance (port 47131) of the first pass answered
`getAlbumList2` with 200; no client was pointed at it. Containers on 47130, 47131, and 47132, a throwaway Xvfb container, and the proxy were started by this
task and removed or stopped at the end (`podman ps -a` shows none of them; nothing listens on 4713x). The oracle's debug log redacts `s` and `t` in logged URLs, so every request path below
is as the oracle logged it.

## Feishin v1.17.0

**URL join.** Feishin has a Navidrome server type (native `/api`) and an OpenSubsonic type (`/rest`). A dubplate user
picks the OpenSubsonic type, so only that path is measured.

- The add-server form strips one trailing slash from the typed URL (`feishin@5c169ac:src/renderer/utils/normalize-server-url.ts:3-7`,
  used at `src/renderer/features/servers/components/add-server-form.tsx:153,225`). The login call strips one trailing
  slash again before appending `/rest` (`src/renderer/api/subsonic/subsonic-controller.ts:333`).
- Each API call joins the stored URL, `/rest`, `/`, and the method name that already carries `.view`
  (`src/renderer/api/subsonic/subsonic-api.ts:500-501,521`). So the on-the-wire form is always `.view`.
- A second configured address, "Public URL", replaces the base only when a server entry prefers it
  (`normalize-server-url.ts:9-26`).
- Observed: a server entered as `http://127.0.0.1:47130/dubplate` and again as `http://127.0.0.1:47130/dubplate/`
  produced requests with a single slash after `dubplate` in both cases (0 of 59 logged requests contained `//`).
  One typed trailing slash is handled. Two would not be: only one is stripped (source), and that case was not run.

**Second code paths** (each built from the stored server URL, never from a response field):

| Path | Where | Observed under `/dubplate` |
|---|---|---|
| Cover art | `subsonic-controller.ts:69` (`<url>/rest/getCoverArt.view?id=...&size=...`) | yes, `GET /dubplate/rest/getCoverArt.view`, 200 and 304 |
| Stream | `subsonic-controller.ts:1977` (`<url>/rest/stream.view?id=...`) | yes, `GET /dubplate/rest/stream.view`, 206 |
| Download | `subsonic-controller.ts:980` | not driven; source shows the same join |
| Transcoded stream | `subsonic-controller.ts:265` (`getTranscodeStream.view`) | not driven (the decision came back direct play) |
| Share | `src/renderer/features/sharing/components/share-item-context-modal.tsx:180-181` joins `<url>/share/<id>`, but the share mutation exists only for the Navidrome server type (`shareItem` appears in `src/shared/api/navidrome/navidrome-types.ts` and in no Subsonic file) | not applicable to the OpenSubsonic type |

**Live walk** (2026-10-07 04:56 to 05:00; Flatpak 1.17.0; Electron 43.7.0; OpenSubsonic type; isolated profile, see
"Isolation"). Driven through Chromium's remote-debugging port with the repo-pinned Playwright 1.63.0 used read-only.

1. Add server `dubplate-sub`, URL `http://127.0.0.1:47130/dubplate`: oracle logged `GET /dubplate/rest/ping.view`,
   `getUser.view`, `getOpenSubsonicExtensions.view` (all 200), then the app landed on Home. One
   `jukeboxControl.view?action=status` got 501 and Feishin carried on.
2. Browse: Home, then the Signal One album page: `getAlbumList2.view`, `getAlbum.view`, `getArtist.view`,
   `getGenres.view`, `getMusicFolders.view`, `getPlaylists.view` (all 200). The page showed the two tracks, 16 s.
3. Cover art: `GET /dubplate/rest/getCoverArt.view?id=...&size=300`, 200 then 304; the cover rendered on the album
   page and in the player bar.
4. Play Alpha: `getTranscodeDecision.view` (POST, 200), then `GET /dubplate/rest/stream.view?id=...` 206. The audio
   element reported `currentTime` 6.6 s of 8 s, not paused, no media error. The next track pre-fetched (second 206).
   Playback reporting followed (`reportPlayback.view`, `scrobble.view`, 200).
5. Credential channels: login `ping` and `getUser` are GET with `u`, `s`, `t` in the query; the browse calls are
   POST (Navidrome declared the `formPost` extension); the stream, cover, and download URLs always carry the
   credentials in the query. No header or cookie carried credentials.
6. All 57 logged Feishin requests began `/dubplate/rest/`; none reached the root.

Request totals by endpoint are in the walk's saved log (`walk-feishin-nd-sub.log`, kept in the lane scratch).

**Tracker.** No issue about a sub-path server URL found. The nearest, `jeffvli/feishin#865` "PUBLIC_PATH + Traefik not
mixing well" (closed), is about hosting Feishin's own web app under a path, which is a different question.
Searched (GitHub issue search, per word and phrase, 2026-10-07): subpath, sub-path, subdirectory, subfolder, reverse
proxy, base url, path prefix, proxy path, ND_BASEURL. The search matches words, so a title that uses other words
could be missed; "none found" is the finding, not "none exist".

## Rufin v0.16.6

**URL join.** One function normalizes what the user typed and one builds every endpoint, and both are
path-aware by design (`Rufin@aad2f44:crates/sources/src/subsonic/client.rs:1107-1124` and `:1126-1137`):

- It trims whitespace and trailing slashes, adds `http://` if the scheme is missing, drops a trailing `/rest` from
  the path if the user typed it, and stores the path with one trailing slash.
- The endpoint builder takes the stored path, appends `/rest/<method>.view`, and strips a `.view` the caller already
  passed so it never doubles. It therefore handles an empty path, a path with a trailing slash, and a path with a
  `/rest` suffix; none of the three produces `//`.
- Observed: the walk used no trailing slash; the source reading above is the evidence for the trailing-slash and
  `/rest` cases (not run live). Unconfirmed live: what it does with two trailing slashes. The first line of the
  normalizer trims every trailing slash, so the source says it copes; a run would confirm.

**Second code paths.** Stream, download, and cover art all go through the same builder (`authenticated_url`,
`client.rs:1220`; stream `:423-431`, download `:402-416`, cover `:480`), so none has its own URL join. A separate
Navidrome-only builder joins the base path to native-API endpoints (`crates/sources/src/subsonic/navidrome.rs:400`);
it is reached only when the user picks the Navidrome source kind (`subsonic/mod.rs:386-394`, kind is
user-chosen, not detected). The stream is fetched by GStreamer's HTTP source (a separate HTTP client with its own
user agent), which receives the finished URL, so the prefix reaches it intact: observed.

**Live walk** (2026-10-07 05:04 to 05:06; Flatpak 0.16.6; OpenSubsonic source kind). The walk ran in a private
Xvfb display (a throwaway container), with screenshots read at each step; the session never touched the desktop.

1. Add Source, OpenSubsonic, name `dubplate-sub`, address `http://127.0.0.1:47130/dubplate`: oracle logged
   `GET /dubplate/rest/ping.view`, `getUser.view`, `getScanStatus.view`, `getMusicFolders.view`.
2. Library sync, all GET and all 200: `getAlbumList2.view` (5 calls), `search3.view` with an empty query and
   `songCount=500`/`musicFolderId=1` paging by offset (3 calls), `getGenres.view`, `getStarred2.view`,
   `getPlaylists.view`. Home and Albums showed all four albums.
3. Album view: opened from Rufin's local store; the oracle logged no request for it. The page listed the one track.
4. Cover art: `GET /dubplate/rest/getCoverArt.view?id=...&size=256`, 6 requests, all 200; covers rendered.
5. Play Signal One: `getOpenSubsonicExtensions.view`, `getLyricsBySongId.view` (200), then
   `GET /dubplate/rest/stream.view?id=...&format=raw` (200) for both tracks, with `scrobble.view?submission=false`
   at start and `submission=true` at end. The player bar showed 0:04 of 0:08 mid-track.
6. Credential channel: query only (`u`, `s`, `t` on every request, the logger redacts `s`/`t`).
7. All 30 logged Rufin requests began `/dubplate/rest/`. The oracle's root `/rest/...` form was never requested.

**Tracker.** No issue about a sub-path server URL found; searches as for Feishin (the matches were all unrelated:
subfolder exclusion, UI refactors).

## Amperfy v2.1.1 (source only)

**URL join.** The login screen trims whitespace and requires an `http` or `https` prefix, then stores the string as
typed (`amperfy@c165a7a:Amperfy/Screens/ViewController/LoginVC.swift:410-428`). The API URL is built by parsing
that string and appending the path components `rest` and `<action>.view` with Foundation's path-append
(`AmperfyKit/Api/Subsonic/SubsonicServerApi.swift:194-209`). So the path survives and `.view` is always present. What
the append does when the typed string already ends in `/` is Foundation behavior this lane did not run or quote;
unconfirmed: it would be confirmed by one iOS or macOS run, or Apple's `URL.appendPathComponent` documentation.
There is no normalizer that strips the path or a trailing slash.

**Second code paths.** Stream (`:336,341,363`), download for the raw format (`:348`), and cover art (`:382-389`)
each call the same authenticated-URL builder (`:211-241`) that every API call uses; none constructs its own join.
The cache key function (`:268-297`) rewrites the host and port, never the path, so a prefixed path keeps its prefix in
the cache key.

**A precondition found.** Before it tries any API, the login runs a reachability check: a plain GET (via
`URLSession.downloadTask`) to the stored server URL itself, no `/rest`, no credentials; any status of 400 or more
fails the login except exactly 401 (`AmperfyKit/Api/BackendProxy.swift:288,315-352`). Redirects are followed by
that session by default, so the final status counts. A server mounted at `/dubplate` must therefore answer
`GET /dubplate` (and the user's typed form, with or without the slash) with a non-error. This is a source reading;
no live run exists. Root-mounted servers meet the same check at `/`.

**Tracker.** No sub-path issue found. `BLeeEZ/amperfy#210` (closed) is an XML parse failure behind an
authentication proxy, unrelated to paths; no other issue matched.

## Tempus v4.28.0 (source only; unmeasured live)

**URL join.** The base is the configured server address plus `/rest/`, with the one doubled slash that a typed
trailing slash would create replaced by a single one (`tempus@687d97f:app/src/main/java/com/eddyizm/tempus/subsonic/Subsonic.java:140-143`).
The HTTP library the app uses for API calls requires its base URL path to end in a slash and resolves each
method against it, so a path in the base survives only when no method string starts with a slash. All 60-odd
service declarations under `subsonic/api/` use relative names (for example `ping`, `stream`, `getAlbumList2`),
none with a leading slash (checked by grepping every `@GET` in that directory). An unusable address falls back to
a placeholder (`subsonic/RetrofitClient.kt:28,47-60`). The address check on the login form accepts any valid
http(s) URL with a host (`ui/login/LoginServerFragment.kt:370-377`).

**No `.view` suffix.** Tempus never sends the `.view` form: the call names above have no suffix, and a search of
`app/src/main` for `.view` finds no URL use. Its wire form is `<server>/rest/ping`. Feishin, Rufin, and Amperfy send
`.view`. A server therefore has to answer both forms under the same prefix. (Navidrome answered both on the
sub-path instance: `ping.view` in the walks, the extensionless form in `curl`; Task 0 recorded both.)

**Second code paths.** Stream (`util/MusicUtil.java:47-48,153-154`), download (`:122-123`), and cover art
(`glide/CustomGlideRequest.java:90-91`) are concatenated by hand from the same `getUrl()` plus a bare method name
(`stream`, `download`, `getCoverArt`), so the prefix reaches them through the shared join. The cover-art downloader for
radio stations tests for the substring `/rest/getCoverArt` rather than anchoring at the root
(`util/RadioCoverArtDownloader.java:57`), which a prefix does not defeat. A share link is the `url` attribute the
server returns in the share response (`ui/fragment/bottomsheetdialog/ShareBottomSheetDialog.java:79`,
`SongBottomSheetDialog.java:340`), so the server must emit prefixed share URLs itself.

**Tracker.** No sub-path issue found. `eddyizm/tempus#593` (closed, 504 behind a tailnet) and `#1113` (open,
"server unreachable" behind a reverse proxy) concern connectivity, not paths.

**Why unmeasured.** No emulator, so no APK. `no emulator manifest entry is owed`: nothing was installed, and the
Android emulator install the plan carries is blocked on firmware virtualization (Task 0, "Not done"). To unblock:
enable VT-x in the firmware, install the SDK and a system image under `~/Android/`, sideload Tempus v4.28.0, and
record the install in its tier manifest that session.

## What the servers document and do about a sub-path

- **Navidrome** documents one option, from `navidrome/website@b0fe572:content/en/docs/usage/configuration/options.md:143`
  (Apache-2.0): "BaseUrl | `ND_BASEURL` | Base URL to configure Navidrome behind a proxy (examples: `/music`,
  `https://music.example.com`) | _Empty_". The Linux install page shows it as an example environment line
  (`content/en/docs/installation/linux.md:221`, `ND_BASEURL="/navidrome"`). The sharing page says shared links take the
  form `http://yourserver.com/share/XXXXXXXXXX` and to allow traffic to `/share` behind a proxy
  (`content/en/docs/usage/features/sharing.md:52`); it does not say how the prefix enters that link.
- **gonic** documents `GONIC_PROXY_PREFIX` / `-proxy-prefix`: "**optional** url path prefix to use if behind reverse
  proxy. eg `/gonic` (see example configs below)" (`sentriz/gonic@5ae99e8:README.md:63`, GPL-3.0 documentation).
  The README rows below that line carry no proxy example this lane could find (a grep for proxy-related words
  returns only the option row).

**What the two servers do, observed** (fixture: scanned; Navidrome 0.64.2 and gonic v0.22.0):

- Navidrome with `ND_BASEURL=/dubplate` serves its API under the prefix. A request for the unprefixed form is
  answered with a redirect to the prefixed web UI (Task 0: `/rest/ping.view` returned 302). Every URL it handed
  back in this walk carried the prefix: `getShares` returned `http://127.0.0.1:47130/dubplate/share/<id>` for a
  share created by `createShare` (host from the request, prefix from the option).
- gonic with `GONIC_PROXY_PREFIX=/dubplate`, on port 47132 in this task: `/rest/ping.view` and `/rest/ping` at the
  root returned 200; `/dubplate/rest/ping.view`, `/dubplate/rest/ping`, `/dubplate`, `/dubplate/`, and `/` each
  returned 303 to `/dubplate/admin/home`. So the option leaves the API at the root and only changes where the web UI
  redirects; it suits a proxy that strips the prefix, and a client given a path-bearing URL gets a redirect. gonic was
  therefore not walked with a client. The field's two servers split: one serves under a prefix natively, one
  expects the proxy to strip it.

## Headers, preflights, ranges, cookies (second pass, through a logging proxy)

Requested by the conductor after Task 2: what each client sends on the paths that decide CORS, `Accept`, and `Range`
behavior. A second walk of each drivable client (same isolated method, fresh profile, fresh Navidrome 0.64.2 at
`/dubplate` on port 47130 with a new data directory, same fixture, scanned, logged in) went through a header-logging
reverse proxy written for this task (a 40-line Python forwarder on port 47133, kept in the lane scratch). The clients
were pointed at `http://127.0.0.1:47133/dubplate`; the proxy logged method, path, selected request headers, status, and
selected response headers, and forwarded unchanged. Navidrome's own debug log does not record headers, so the proxy is
the only header channel. Both walks completed login, browse (Feishin), library sync (Rufin), play, and cover art.

| Question | Feishin v1.17.0 (44 logged requests, all under `/dubplate/rest/`) | Rufin v0.16.6 (25 logged requests, all under `/dubplate/rest/`) |
|---|---|---|
| `Accept` on `getCoverArt` | Both forms; Task 2 predicted `*/*` only: the rendered `<img>` requests send `image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8` (3 requests); the programmatic fetch sends `*/*` (3 requests). | `*/*` on all 5 (a native HTTP client, no `image/webp` offer). |
| CORS preflight | None. 0 `OPTIONS` requests and 0 `Access-Control-Request-*` headers across the whole walk, including the JSON POST `getTranscodeDecision`. Only 2 cover requests carried an `Origin` (value `file://`); the other 42 carried none. The server answered those two with `Access-Control-Allow-Origin: *` and no credentials header. | None; native client, no `Origin` on any request. |
| `Range` on stream | `Range: bytes=0-` on both stream requests, answered `206` with `Accept-Ranges: bytes` and a full-length `Content-Range`. A programmatic seek to 5 s produced no new request (the whole 8 s file was buffered). | None sent. The GStreamer HTTP source requested the stream with no `Range` and no `Accept`; the server answered `200` with `Accept-Ranges: bytes`. Seeking was not driven. |
| Stream URL against the API base | Same. The stream URL is the stored server URL plus `/rest/stream.view` and the API calls go to the stored server URL plus `/rest`, so with `http://127.0.0.1:47133/dubplate` stored both landed at `/dubplate/rest/...` (all 44 requests, stream included, under that prefix). They can diverge in source only: API and cover calls pick the optional "Public URL" when a server entry prefers it, while stream, download, and transcoded-stream calls read the raw stored URL (`subsonic-controller.ts:265,980,1977` against `normalize-server-url.ts:9-26`). With no Public URL set, as here, the two agree. | Single builder for all calls (see above); observed under the prefix. |
| Cookie from the server | Never sent: 0 of 44 requests carried a `Cookie` header. | Sent on 1 of 25: the GStreamer stream request (its HTTP stack keeps a cookie jar); the API client sent none. |

**Navidrome's `Set-Cookie` under a base path.** Observed on every `/rest` response in this walk: Navidrome sets a
cookie named `nd-player-<hex of the username>` with `Path=/dubplate`, `Max-Age=31536000`, `HttpOnly`,
`SameSite=Strict`. So the cookie's path follows the base path (Task 1's `Path=/` was a root-mounted instance). Under a
sub-path this is the narrower, correct scope and matters to none of the four clients here: Feishin never sent it,
and the one request where Rufin's stream source did send it was a request under the same prefix. For dubplate's
server the lesson is one of scope only: if it ever sets a cookie on `/rest`, give it the mount prefix as its path.
Whether Navidrome's cookie does anything for a client (it appears to key a player identity for the debug log's
"Found matching player" line) was not measured.

## Cross-client facts for slice 3

Facts only; slice 3 decides (re-plan `:427`).

- Three clients (Feishin, Rufin, Amperfy) send `.view` on every method; Tempus sends none. A path mount has to
  answer both on every method.
- Feishin and Rufin carry the prefix correctly in the live walks. Amperfy and Tempus carry it by source reading.
  No client was observed to drop a path or double a slash with a single path segment and zero or one trailing slash.
- No preflight was seen from either walked client, and no walked request needed a CORS header to succeed; Feishin's two
  `Origin: file://` cover requests got `Access-Control-Allow-Origin: *` from Navidrome. Credentials travel in the query string for every stream, cover, and download URL in all four clients, so the
  prefix does not change the credential channels. Feishin's POST form calls (when the server declares `formPost`)
  carry them in a body; the other three never POST credentials.
- Amperfy needs `GET <typed base URL>` to return under 400 (or 401) before it logs in.
- Share and other links a client displays come from the server's own response (Tempus) or are built only for the
  Navidrome server type (Feishin), so the server must emit prefixed URLs when mounted under one, as Navidrome does.
- Navidrome's answer to an unprefixed request is a redirect to its web UI, not an error: a client misconfigured
  without the prefix sees 302 and not a clean failure.

## Isolation and the profile hashes

Each walk ran only under an isolated configuration home, verified before login.

- **Feishin** (`org.jeffvli.feishin`; both passes): `flatpak run --command=sh` with a granted filesystem path, `HOME` and `XDG_*`
  exported inside the sandbox command line (the `--env` flag does not override Flatpak's own `XDG_*` values; exporting
  inside the shell does), `--no-talk-name=org.freedesktop.secrets` and `--no-talk-name=org.kde.kwalletd5`, and
  `--password-store=basic` on the command line (the app honors it: `feishin@5c169ac:src/main/index.ts:357-360`; its
  default store is the host keyring). Before login, the profile's `config/feishin/` (Electron userData, `config.json`,
  IndexedDB, cookies) existed under the lane's isolated directory, and `~/.var/app/org.jeffvli.feishin` held no file.
  The host GNOME keyring files (`~/.local/share/keyrings/login.keyring`, `user.keystore`) hashed identically before
  and after.
- **Rufin** (`io.github.screwys.Rufin`): same method, run on a private Xvfb display so no input reached the desktop.
  Before login, `settings.json` and the SQLite store (`rufin-store.sqlite`, `-wal`, `-shm`) were created under the
  lane's isolated `config/` and `data/`. After login the credential file `data/keyrings/default.keyring` appeared in
  the isolated `data/` and not in the real profile. Rufin's secret service call went to the portal (a `zbus` warning
  about the portal properties was logged); the encrypted file itself sits in the isolated directory.
- Tempus and Amperfy: not run, so nothing to isolate.

**Hash method.** Before the first walk and after the last (the second, header-capturing pass included; the final hash
equals the one taken after the first pass, so the second pass changed nothing further), every file under `~/.var/app/<id>/` except `cache/` was
hashed with SHA-256 (size and mtime beside it; directories and symlinks listed).

| Client | Before | After | Changed files and cause |
|---|---|---|---|
| Rufin (Geoff's profile) | `settings.json` `0c07ad71...`, `rufin-store.sqlite` `9218c12b...` (+ `-wal` `f4306cc2...`, `-shm` `1e23b302...`), `keyrings/default.keyring` `8726e974...`, `pulse/cookie` `04e5dd6c...`, three log files `aaa19197...`, `b942b8ae...`, `813214d8...` | identical | **No change** to any config, data, keyring, or log file. One entry differs: Flatpak's own dynamic-linker cache under `.ld.so/` was regenerated on launch; its content hash is the same (`fc06204d...`, 46087 bytes) under a new file name (`7eb3feb9...` replacing `2d8f3752...`) and the `active` symlink was retargeted. Cause: `flatpak run` writes that cache; the app did not. |
| Feishin (never run before) | directory absent | directories `config/`, `data/`, `.local/state/`, `sandbox/`, and `.ld.so/` (one file, `active` symlink) created; no other file | Cause: `flatpak run` creates the per-app skeleton and the linker cache on first launch. The app's profile went to the isolated directory instead (nothing under `config/` or `data/` in the real tree). |

The Rufin profile Geoff may use was never opened by an app run: every Rufin launch used the isolated directories.
The same linker-cache regeneration happens on any launch of that Flatpak, so Geoff's next normal launch would have made
it too.

## Not done, and what would confirm

- Tempus and Amperfy live walks (no emulator, no iOS): the re-plan walks Amperfy at rung 16; Tempus waits on
  VT-x and an emulator.
- Two trailing slashes, and Amperfy's trailing-slash append, are unconfirmed (source or documentation read only).
- Feishin's download URL and transcoded-stream URL were read, not driven; Rufin's download likewise.
- gonic was not walked with a client (its option does not mount the server under a prefix).

## Routed

- **Slice 3** (the decision and the erratum): the whole "Cross-client facts" section, above all the `.view` split,
  the Amperfy base-URL reachability check, and the server-emitted share URLs.
- **Task 2** (client demands): the header table above answers its three routed items (cover `Accept` forms, Feishin
  preflight and `Range`, Feishin stream URL against the API base). Also: Feishin sends `jukeboxControl` with `action=status` on connect (501 from Navidrome
  tolerated); Rufin calls `getLyricsBySongId` with `enhanced=true`, `getScanStatus`, and `search3` with an empty
  query paged by 500; Feishin calls `getTranscodeDecision` when the server declares the `transcoding` extension.
- **Task 4** (server behavior): Navidrome's 302 to its web UI for an unprefixed request; its prefixed share URL; the
  gonic prefix behavior above.
- **Slice 2** (harness): the walk method that worked (isolated Flatpak profile via an in-sandbox `export`, a private
  Xvfb display in a throwaway container for Rufin, a CDP port for Feishin) is reusable for its roster walk.
- **Geoff, a routed item and not a ruling:** enabling VT-x unblocks the emulator for Tempus and slice 2.

## Prior art consulted

Navidrome documentation (Apache-2.0, quoted), gonic README (documentation row quoted), Navidrome and gonic as live
oracles (observation only). The four roster clients read at their pinned commits for behavior, never copied. Nothing
adopted or declined anew. Neutral on hooks.
