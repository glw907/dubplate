# Slice 1, Task 1: the contract, the encoding quirks, and the XML rule

Date read: 2026-10-07 (UTC, 04:50 to 05:20). Staged outside the repo; lands at A-land as
`docs/superpowers/research/2026-10-06-slice-1-contract.md`.

**Pins read** (all from `2026-10-06-slice-1-evidence-pins.md`, Task 0):

- OpenSubsonic specification: `opensubsonic/open-subsonic-api@bed1688a7b0afe9f528a3dc1400a46ab7961b280`
  (Apache-2.0), cited below as `osapi:<path>:<line>`.
- Navidrome 0.64.2, `docker.io/deluan/navidrome@sha256:38dc2727bf...ee52`, cited as "nd".
- gonic v0.22.0, `docker.io/sentriz/gonic@sha256:516fd96456...1d49`, cited as "go".
- Navidrome documentation: `navidrome/website@b0fe5725` (Apache-2.0), cited as `nd-docs:<path>:<line>`.
- Subsonic 1.16.1 XSD, fetched 2026-10-07 from `https://subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd`
  (the `http://` URL answers 301 to this one): 32775 bytes, sha256
  `9baee72fa24087190ad22be4f7674d628acc7a3e0fb65160d74115b166df0387`, server `last-modified` 2021-09-08. Cited as
  `xsd:<line>`. This is a new pin, owed by this task; Task 0 did not pin it.
- Ampache `develop` at `3820b0658033c9591ad67cc3a367db3e55523e67` (2026-10-06), AGPL-3.0, read for method only.
- `xmllint` from libxml2 2.15.4 (`xmllint: using libxml version 21504`), the workstation Homebrew build.

**Questions this record owns, and where each ended.**

| # | Question | Outcome |
|---|---|---|
| Q1 | The extension list: endpoints, versions, what a server does to declare each | Answered ("Q1") |
| Q2 | The encoding quirks: attribute-heavy XML, the wrapper, singleton versus array, JSONP, and any others found | Answered, with test-owed rows ("Q2") |
| Q3 | The XML validation rule: legacy form against the official XSD, or a derived XSD | Answered with a recommended reading for Task 7 ("Q3") |
| Q4 | State of OpenSubsonic PRs #162, #237, #258, #261 and newer PRs near folders, collections, library browsing; what #237 means for dubplate's word | Answered as of 2026-10-07; a re-read is owed before the harness pins a build ("Q4") |
| Q5 | The error-code list and the text of errors 0, 50, and 70 | Answered ("Q5") |
| Q6 | Every channel a client uses to send credentials, and the unauthenticated extensions carve-out (plan check 3) | Answered ("Q6") |

## Method and fixture states

Every oracle claim below names its instance and state. Two sessions ran on 2026-10-07, each from the Task 0 run
recipe (`work/run-oracle.sh`, whose text is in the evidence pins record) with an empty data directory under `work/task-1/data/` (a workstation-scratch path that did not land in the repo), bound to `127.0.0.1`, ports
47110 (Navidrome) and 47111 (gonic). Both data directories and both containers were removed after each session
(`podman ps -a` and `ss -ltn` showed nothing of this lane's afterward).

- **Library state, both servers:** the Task 0 fixture (4 albums, 5 tracks, one 45 minute track), mounted read-only,
  scanned (`getScanStatus` reported `scanning: false`, `count: 5`) before the first request. No stream was started.
- **Login state:** token authentication (`u`, `t`, `s`) with `v=1.16.1` and `c=probe` unless a row says otherwise.
  Navidrome user `probe` (created through `/auth/createAdmin`, admin), gonic user `admin` (the shipped default,
  admin). One extra non-admin Navidrome user, `plain`, existed for the error 50 rows.
- **Other state:** one playlist named with markup and non-ASCII characters existed on each server for the escaping
  rows (created by `createPlaylist`). Library idle: no scan, no play history, no stars, no shares, no bookmarks, no
  play queue. A row whose result depends on that emptiness says "idle".
- **Session 2** (after a stop and recreate, same recipe, fresh data) observed the binary-endpoint error content
  types and `getUser`. Everything else came from session 1.
- **Request shape:** `GET <base>/rest/<endpoint>?u=...&t=md5(password+salt)&s=...&v=1.16.1&c=probe&f=<fmt>` through
  `curl`. A `.view` suffix is not re-observed here: Task 0 recorded 200 with the suffix on both servers.

Captured bodies lived in `work/task-1/obs/` (JSON and XML per endpoint per server, errors, escaping), a workstation-scratch location that did not land in the repo. Clean-room rule
1 holds: nothing from a copyleft repository appears as text below. Spec and documentation lines are quoted briefly
from Apache-2.0 sources, and the two servers are described from their observed behavior.

---

## Q1. The extension list

At `bed1688` the `content/en/docs/Extensions/` directory holds ten extension pages (plus `template.md` and
`_index.md`). The name a server returns is not always the file name: `apiKeyAuth.md` declares
`apiKeyAuthentication` (`osapi:content/en/docs/Extensions/apiKeyAuth.md:12`) and `TranscodeOffset.md` declares
`transcodeOffset` (`TranscodeOffset.md:12`). The versions page lists only the template and `sonicSimilarity`
(`osapi:content/en/docs/opensubsonic-versions.md:15`), so it is stale and the per-extension pages are the source.

**Declaring an extension.** A server lists `{name, versions[]}` objects in the `openSubsonicExtensions` array of
`getOpenSubsonicExtensions` (`osapi:content/en/docs/Responses/OpenSubsonicExtension.md:24-27`). The endpoint is
"Unlike all other APIs ... **must** be publicly accessible"
(`osapi:content/en/docs/Endpoints/getopensubsonicextensions.md:20`), and implementing it is one of three required
changes for any OpenSubsonic server (`osapi:content/en/docs/opensubsonic-changes.md:31-35`). The other two are the
envelope fields and returning error 41 when token authentication is unsupported while advertising a version above
1.13.0.

| Extension name (as returned) | Page | Versions | What it adds (endpoints, parameters) | What a server must do to declare it | nd 0.64.2 | go v0.22.0 |
|---|---|---|---|---|---|---|
| `apiKeyAuthentication` | `apiKeyAuth.md` | 1 | The `apiKey` query parameter; `tokenInfo` (`Endpoints/tokeninfo.md:15`); errors 42, 43, 44; `helpUrl` on `error` | Accept `apiKey=<key>` and answer error 43 if `u` or another mechanism accompanies it (`:34,38`); provide a way to view and revoke keys (`:28`); keys do not expire (`:29`); keep a key under 2048 URL-encoded characters (`:26`). Dropping token auth means answering 41, dropping another mechanism means 42 (`:40-42`). | No. `apiKey=xyz` answered error 10 "missing parameter: 'u'" | No. Same probe answered error 10 |
| `formPost` | `formPost.md` | 1 | `application/x-www-form-urlencoded` POST on every endpoint (`:16`); the OpenAPI file carries a POST operation for 86 of its 87 paths (85 as `get`+`post`, `getTranscodeDecision` as `post` only, `getTranscodeStream` as `get` only) | Accept the arguments, authentication included, in the POST body | Yes | Yes |
| `getPodcastEpisode` | `getPodcastEpisode.md` | 1 | `getPodcastEpisode` | Serve one episode's metadata | No | No |
| `indexBasedQueue` | `indexBasedQueue.md` | 1 | `savePlayQueueByIndex`, `getPlayQueueByIndex` (`:14`) | Save and read the play queue by index, so duplicates are allowed (`:18-19`) | Yes | No |
| `playbackReport` | `playbackReport.md` | 1 | `reportPlayback` (`mediaId`, `mediaType`, `positionMs`, `state`, `playbackRate`, `ignoreScrobble`), timeline fields on `getNowPlaying` (`:14,18`) | With `ignoreScrobble=true`, update now-playing state only and cause no play-count or scrobble effect (`:20`) | Yes | No |
| `songLyrics` | `songLyrics.md` | 1 and 2 | `getLyricsBySongId`; v2 adds the `enhanced` parameter, `kind`, `agents`, `cueLine`, `cue` (`:28-33`) | v1: structured line-level lyrics by song id (`:16-20`). v2 fields appear only with `enhanced=true`, and a v2 server advertises `[1, 2]` (`:39-41`) | Yes, `[1, 2]` | Yes, `[1]` |
| `sonicSimilarity` | `sonicSimilarity.md` | 1 | `findSonicPath`, `getSonicSimilarTracks`, both returning `sonicMatch` entries (`:16-21`) | Serve both endpoints | No | No |
| `topSongsByArtistId` | `topSongsByArtistId.md` | 1 | The `id` parameter on `getTopSongs`, which "takes precedence over `artist`" (`:14,20`) | Honor `id` | Yes | No |
| `transcodeOffset` | `TranscodeOffset.md` | 1 | The `timeOffset` parameter of `stream` for music (`:14,20`) | Honor `timeOffset` for music, not only video (`osapi:content/en/docs/Endpoints/stream.md:41`) | Yes | Yes |
| `transcoding` | `transcoding.md` | 1 | `getTranscodeDecision` (POST, JSON `ClientInfo` body; `Endpoints/getTranscodeDecision.md:15,26`) and `getTranscodeStream` (`:18-21`) | Serve both; the client must replay the returned `transcodeParams` unchanged (`Endpoints/getTranscodeStream.md:14`) | Yes | No |

Navidrome 0.64.2 answered `getOpenSubsonicExtensions` with seven of the ten: `transcodeOffset`, `formPost`,
`songLyrics` `[1,2]`, `indexBasedQueue`, `transcoding`, `playbackReport`, `topSongsByArtistId` (authenticated and
unauthenticated bodies were byte-identical, 440 bytes). gonic v0.22.0 answered three: `transcodeOffset`, `formPost`,
`songLyrics` `[1]`. gonic answers the unauthenticated request with error 10 (see Q6), so its list is only readable
with credentials.

**Extensions the open PRs would add** (Q4): `musicFolder` (#162), `collections` (#237), `extendedLibrary` (#258),
`publicImageUrl` (#261), `artistRoleFilter` (#256), `playlistFolders` (#255), `ratingID3` (#184),
`search3Playlists` (#221). None is merged.

**For Task 7's launch subset, three facts the table does not show.** (1) `apiKeyAuthentication` has no oracle
coverage at all (neither oracle declares it), matching the spec's declared divergence class
(`streamer-2-0-design.md:285-287`), and PR #278 (Q4) is rewriting the surrounding authentication text now. (2)
`formPost` is the one extension both oracles declare and the one that moves credentials out of URLs; PR #278
recommends clients prefer it. (3) `transcoding` and `transcodeOffset` interact with the profile order the spec
describes at `:407-445`; the observation of both oracles under load belongs to Task 4.

---

## Q2. The encoding quirks

Each item gives what the specification says, what the two servers did, and the consequence for a hand-written Go
encoder. "Test owed" rows feed Task 7's ledger (section "Candidate ledger rows").

### E1. The wrapper and the envelope

Every JSON body is `{"subsonic-response": {...}}`. The envelope carries `status`, `version`, `type`,
`serverVersion`, `openSubsonic`, all required (`osapi:content/en/docs/Responses/Subsonic-response.md:33-40`). In XML
the root element is `subsonic-response` in the default namespace `http://subsonic.org/restapi`, and the envelope
fields are attributes (`xsd:1-9,11`; observed on both servers). The `version` field is the server's protocol
version. nd answered `1.16.1` and go answered `1.15.0` whatever the client sent. The spec asks OpenSubsonic servers to
support at least 1.14.0 (`osapi:content/en/docs/subsonic-versions.md:39`).

### E2. Format selection

The spec's default is XML: parameter `f`, "Supported values are 'xml', 'json' (since 1.4.0) and 'jsonp' (since
1.6.0)", default `xml` (`osapi:content/en/docs/api-reference.md:25`). The OpenAPI file says "Must be 'xml' or 'json'.
For the purposes of this document, only `json` is allowed." (`osapi:openapi/openapi.json:426`), and the documentation
set shows only JSON because "most clients now only rely on JSON" (`osapi:content/en/docs/_index.md:13`).

Observed on nd and go (ping, idle): `f` absent, `f=xml`, `f=bogus`, and `f=JSON` (uppercase) each returned XML, so
the match is case-sensitive and anything unrecognized falls back to XML. `f=json` returned `application/json`. XML
returned `application/xml`. Neither sets a charset, although the spec says the XML is UTF-8
(`api-reference.md:31`). **Consequence:** the default answers XML, so the XML path is live for any client that omits
`f`, and an unknown `f` must not be an error.

### E3. JSONP

The spec mentions JSONP once, in the `f` row, with a `callback` parameter (`api-reference.md:25`). The OpenAPI file
omits it. The two servers differ sharply (ping, idle, `f=jsonp`):

- **nd** validates the callback. `foo`, `foo.bar`, and `foo_1$x` were wrapped as `foo({...})` with no trailing
  semicolon and `application/javascript`. A missing callback and `a.b;evil()` answered error 0 "invalid callback parameter"
  as `application/json`; an empty callback, `<script>`, `9a`, and `a-b` also answered `application/json` (those bodies were not
  read, so error 0 is inferred for them). `f=json&callback=cb`
  ignored the callback.
- **go** validates nothing. A missing callback defaulted to `cb`. `a.b;evil()` and `<script>` were reflected
  verbatim as the function prefix, with a trailing semicolon, `application/javascript`, and no
  `X-Content-Type-Options`. An empty callback produced a bare `(` prefix.

**Consequence:** a Go server that answers `f=jsonp` must validate the identifier, because the callback is reflected
into script text on a response that also carries the member's data. Whether to answer JSONP at all is Task 7's call,
with Task 2's evidence on whether any roster client sends it. Test owed (candidate rows C-04, C-05).

### E4. Attribute-heavy XML and its shape rules

The 1.16.1 XSD puts nearly every scalar in an attribute (every `xs:attribute` in `xsd:80-640`) and uses child elements
only for nested records and lists. Two types carry text content: `Genre` is `mixed="true"` (`xsd:117`) and so is
`Lyrics` (`xsd:403`), so the JSON `value` field is the element's character data in XML (`osapi:content/en/docs/Responses/genre.md:13-18`
shows the JSON form).

Both servers extend that shape the same way for the OpenSubsonic fields, observed on `getAlbum` of a one-song album
(idle):

- scalars as attributes (nd: `sortName`, `displayArtist`, `mediaType`, `channelCount`, `samplingRate`, `bitDepth`;
  go: `displayArtist`, `displayAlbumArtist`, `musicBrainzId`);
- a JSON array of objects as repeated child elements named by the JSON key, with no wrapper (`artists`,
  `albumArtists`);
- a JSON array of scalars as repeated child elements with text (nd `getOpenSubsonicExtensions`:
  `<openSubsonicExtensions name="songLyrics"><versions>1</versions><versions>2</versions></...>`);
- the root element carries a JSON object's name (`album`, `albumList2`, `openSubsonicExtensions`).

They differ at the edges. go emitted `isCompilation` and `releaseTypes` as child elements
(`<isCompilation>false</isCompilation>`), where the XSD convention would make a boolean an attribute. nd omitted
`isCompilation` and every empty list from the XML while the JSON carried them (E6). go also wrote
`artistImageUrl` as a child element of `artistInfo2`. The XSD declares that name only as an attribute on the artist
types (`xsd:105`, `xsd:140`), and neither the XSD nor the OpenAPI file declares it on `artistInfo2`. Neither XML shape for the OpenSubsonic additions is specified anywhere in the pinned documentation: the
Responses pages and the OpenAPI file describe JSON (`E2`), apart from one schema that also declares `application/xml`
for `transcodeDecision` (`osapi:openapi/schemas/TranscodeDecisionResponse.json:18-30`). **Consequence:** the XML form of
every OpenSubsonic addition is a dubplate decision. The convention both oracles mostly share (scalars as
attributes, lists as repeated elements, text for string lists) is the defensible default. Test owed (C-06).

PR #258's `extendedLibrary` is the one proposal that declares its endpoints "intentionally JSON-only and POST-only"
and unsupported in "form-encoded POST, XML" (Q4). It is the first precedent for an endpoint with no XML form.

### E5. Singleton versus array in JSON

Both servers always emit a JSON array for a list that holds one element. Observed paths with a single element on
the idle fixture: `musicFolders.musicFolder` (one folder), `indexes.index[].artist`, `artists.index[].artist`,
`album.song` (one song), `directory.child` (one child), `randomSongs.song` (size 1), `albumList2.album` (size 1),
`song.artists` and `song.albumArtists` (one artist each), and on go `album.releaseTypes` and `user.folder`. No
single-element list collapsed to an object on either server. The OpenAPI schemas declare every list property as
`type: array` (`osapi:openapi/schemas/AlbumList2.json`, `MusicFolders.json`, `Indexes.json`, and the rest), so an
object in that position violates the schema.

The pinned documentation does not mention a collapse rule, so the claim that the original Subsonic server collapsed
singletons is unconfirmed here. What would confirm it: running the original Subsonic server (out of scope) or a
roster client that special-cases it (Task 2). **Consequence:** the wire types use slices for every list, and the
encoder never switches on length. Test owed (C-07): a one-element list at each path above.

### E6. Empty lists, empty values, and null

The spec has two statements. For every OpenSubsonic-added field: "All OpenSubsonic added fields are **optionals**. But
if a server support a field it **must** return it with an empty / default value when not present in it's database so
that clients knows what the server supports." (`osapi:content/en/docs/Responses/Child.md:254`, repeated at
`AlbumID3.md:157`, `AlbumID3WithSongs.md:271`, `ArtistID3.md:65`, `ArtistWithAlbumsID3.md:105`, `playlist.md:62`,
`playlistWithSongs.md:123`). And for `ReplayGain`: "If the data is not present the field must be omitted in the
answer. (But the replayGain field on Child must always be present)" (`Responses/ReplayGain.md:36`).

The OpenAPI schemas make every container's list property optional (`required: []` for `AlbumList2`, `Genres`,
`MusicFolders`, `SearchResult3`, `Starred2`, `Playlists`, `LyricsList`, `Podcasts`, `TopSongs`,
`InternetRadioStations`, `Shares`, `Songs`), require `NowPlaying.entry`, and declare no property `nullable` (no file under
`openapi/` contains the word). Observed, idle:

| Case | nd 0.64.2 JSON | go v0.22.0 JSON |
|---|---|---|
| Container with no items (`getGenres`, `getPlaylists`, `getStarred2`, `searchResult3` no match, `getTopSongs`, `getAlbumList2` past the end) | the container, empty: `{"genres":{}}` | mixed: `{"genres":{"genre":[]}}`, `{"playlists":{"playlist":[]}}`, `{"albumList2":{"album":[]}}`, but `{"starred2":{}}` and `{"topSongs":{}}` |
| An added list inside a record (`genres`, `moods`, `isrc`, `artists`) | `[]`, present | `[]`, present for `genres` and `isrc`; `releaseTypes` is `[""]` |
| `replayGain` on a song | `{}` | `null` |
| `lyricsList` for a song with no lyrics | `{}` | `{"structuredLyrics":null}` |
| `podcasts` | not implemented (bare 501, E8) | `{"channel":null}` |
| `nowPlaying` while idle | `{}` (the schema requires `entry`) | answered error 70 "view not found" |
| `getPlayQueue` with no saved queue | `{"playQueue":{"username":"probe","changed":"0001-01-01T00:00:00Z","changedBy":""}}` | no payload element at all, `status: ok` |

Three findings follow. go emits JSON `null` for three fields (`replayGain`, `structuredLyrics`, `channel`) and the OpenAPI schema forbids it. nd's `nowPlaying: {}` on
an idle server contradicts the schema's required `entry`. nd emits empty values for added JSON fields as the spec
asks, then omits the same values from XML (E4), so the spec's emit-empty rule cannot be seen in XML at all.
**Consequence:** dubplate emits empty-valued additions in JSON per the spec, uses an empty container for a list with
no items, never emits `null`, and declares XML omission of empties as a divergence. The oracle diff must treat the
nulls and the idle `nowPlaying` as oracle defects, not targets. Test owed (C-08, C-09).

### E7. Identifier types

Identifiers are strings everywhere except `musicFolder.id`, which is an integer (`xsd:81`, `xs:int`;
`osapi:openapi/schemas/MusicFolder.json`, `id: integer`). Observed: nd `getMusicFolders` returned `{"id":1,"name":"Music
Library"}` and go returned `{"id":0,"name":"music"}`. Navidrome's documentation warns that its string identifiers
break clients that "insist in converting IDs to integers" (`nd-docs:content/en/docs/developers/subsonic-api.md:27`).
**Consequence:** dubplate's collection, which maps to a music folder (spec `:951-955`), needs an integer wire id in
`getMusicFolders` and in every `musicFolderId` parameter, beside the string ids of songs and albums. Routed to
slice 3 (catalog projection). Test owed (C-10).

### E8. HTTP status, unknown endpoints, and errors on binary endpoints

Both servers answered every API error with HTTP 200 and the envelope. The exceptions are unknown or unimplemented
endpoints. nd answered a bare `404 page not found` for an unknown name and a bare `501` with the plain text "This
endpoint is not implemented, but may be in future releases" for a known-but-unimplemented one (`getPodcasts`,
`createUser`), both `text/plain`, no envelope. go answered HTTP 200 with error 70 "view not found" in the envelope,
in the requested format. The spec design is error 0 inside a well-formed envelope "never a bare 404"
(`streamer-2-0-design.md:670-672`), which neither oracle does.

For `stream`, `getCoverArt`, and `download` the spec says the endpoint "Returns binary data on success, or an XML
document on error (in which case the HTTP content type will start with 'text/xml')"
(`osapi:content/en/docs/Endpoints/stream.md:36`, `getcoverart.md:35`; `osapi:openapi/responses/BinaryResponse.json:10`
declares `text/xml`). Observed, session 2, nonexistent id, idle: both servers answered HTTP 200 with the envelope,
`application/xml` when `f` was absent or `xml` and `application/json` for `f=json`. Neither sent `text/xml`.
**Consequence:** a client that tells an error from audio by a `text/xml` prefix would misread both oracles, and the
roster evidence (Task 2) decides what clients test. Routed to Task 2 (the empty-success set and error detection) and
recorded as ledger rows C-11 and C-12.

### E9. Version negotiation

Neither server answers error 20 or 30. nd and go accepted `v=2.0.0`, `v=0.9.0`, and `v=1.17.0` on `ping` with `ok`.
A missing `v` gave nd error 10 "missing parameter: 'v'" and go `ok`. The spec defines both codes (Q5) and documents
the compatibility rule that a server supports clients at its own major version and a minor version at or below its
own (`osapi:content/en/docs/subsonic-versions.md:36`). **Consequence:** dubplate may answer `ok` for any `v` and
still match both oracles. Whether to enforce the rule is Task 7's reading, and the roster's behavior decides it
(a client that sends 1.16.1 to a server reporting 1.16.1 never meets the rule). Ledger row C-13.

### E10. Escaping and character handling

Observed with a playlist named with a quote, ampersand, angle brackets, accented letters, a check mark, and CJK
characters (one playlist, idle). Both servers wrote the characters as literal UTF-8 in JSON and XML. JSON escaped
`<`, `>`, `&` as `\u003c`, `\u003e`, `\u0026` (the Go standard library's default HTML-safe encoding, valid JSON)
and left `"` as `\"`. XML escaped `"` as `&#34;`, and `&`, `<`, `>` as entities. Neither wrote an XML declaration. The
XSD does not require one. **Consequence:** the Go encoders give the same bytes by default, so no hand-rolled escaping
is needed, and the JSON tier should compare parsed values, not bytes. Test owed (C-14).

### E11. Credentials echoed into URLs

go's `getArtistInfo2` returned three image URLs (`smallImageUrl`, `mediumImageUrl`, `largeImageUrl`) each carrying
the requester's `u`, `t`, `s`, `v`, `c`, and `f` as query parameters, in both JSON and XML (idle, token authentication,
no external service configured). nd's `getArtistInfo2` returned `{}`. Of the 68 captured JSON and XML bodies, only go's
`getArtistInfo2` (one JSON, one XML) contained the credential parameters. **Consequence:** an endpoint that returns a URL must never
rebuild it from the request's authentication. Dubplate hands out cover-art ids, and the one URL-bearing shape
(`artistInfo`, image URLs) is a place a test must assert no `t=`, `s=`, `p=`, or `apiKey=` appears. Test owed (C-15).

### E12. Other observed behavior, noted for the slice that owns it

- **CORS.** go answered every response with `Access-Control-Allow-Origin: *`; its preflight allowed
  `POST, GET, OPTIONS, PUT, DELETE`. nd answered a preflight with `Access-Control-Allow-Origin: *`, methods `POST`,
  headers `Content-Type`, and `Vary: Origin`, but sent no CORS headers on a plain GET. Routed to the CORS owner
  (Task 2, then Task 7); not measured further here.
- **Cookies.** nd set `Set-Cookie: nd-player-<hex>=...; Path=/; Max-Age=31536000; HttpOnly; SameSite=Strict` on
  authenticated `/rest` responses. A `Path=/` cookie is relevant to the base-path measurement; routed to Task 3.
- **Authorization header.** Neither server accepted HTTP Basic (`curl -u`): both answered error 10 for the missing
  `u`. No pinned spec text mentions a header credential.
- **A long query.** A 700-parameter `songId` query (about 6 KB) returned HTTP 200 from both (GET). The `formPost`
  motivation (`api-reference.md:39`) is the URL length limit; neither server's limit was found.
- **Error 50 text.** nd's wording drops the final period of the spec's text (Q5).

---

## Q3. The XML validation rule

### What the official XSD does to an OpenSubsonic response

The XSD is strict. It has no `xs:any` and no `xs:anyAttribute` (a search of the file finds neither), and its
`Response` type declares two attributes, `status` and `version` (`xsd:11-60`). Validating real oracle output (xmllint
2.15.4) shows the spec's description (spec `:266-271`, three rejected envelope attributes) is true and incomplete:

| Stage (one XML body per endpoint, per server; 32 bodies each, including error envelopes: nd 1, go 3) | nd 0.64.2 valid | go v0.22.0 valid |
|---|---|---|
| Raw output | 0 of 32 | 0 of 32 |
| After dropping only `type`, `serverVersion`, `openSubsonic` | 20 of 32 | 18 of 32 |
| After the scoped projection below | 28 of 32 | 23 of 32 |

The 12 nd and 14 go bodies that still fail after the envelope-only strip carry OpenSubsonic additions on `song`, `album`, `artist`,
`child` (for example `displayArtist`, `sortName`, `musicBrainzId`, `mediaType`, `channelCount`, `samplingRate`,
`bitDepth`, and the child elements `artists` and `albumArtists`). The XSD also lacks the root element for several
OpenSubsonic responses: `openSubsonicExtensions`, `lyricsList`, `tokenInfo`, `playQueueByIndex`, `podcastEpisode`,
`sonicMatch`, and `transcodeDecision` are not in the `Response` choice (my count from the OpenAPI endpoint
directories against `xsd:11-56`), so a derived XSD must extend the root choice too.

What the pinned OpenAPI file adds beyond the XSD for the types a launch response carries (the script in the appendix
computes it, property names minus XSD attributes and child elements): `Child` 23 names, `AlbumID3` and
`AlbumID3WithSongs` 16 (without `song`), `ArtistID3` and `ArtistWithAlbumsID3` 4, `Playlist` and `PlaylistWithSongs`
2, the envelope 3, and `error` 1 (`helpUrl`): **49 names on six types**, plus eight nested record types
(`ItemGenre`, `ItemDate`, `DiscTitle`, `RecordLabel`, `Contributor`, `ReplayGain`, `Work`, `Movement`) that the
additions reference, plus the root elements above.

**Residuals after the scoped projection** are all vendor extras or defects, outside both contracts: nd put
`lastModified` on `artists`, `songCount` on `directory`, and `lastScan`, `scanType`, `folderCount`, `elapsedTime` on
`scanStatus`, and `lyricsList` has no XSD element. go put `title` and `album` attributes on ID3 `album`, emitted
`artistImageUrl` as an element, wrote `albumCount` on non-ID3 `artist`, omitted the required `lastModified` on
`indexes` (its value was zero), and wrote `folder` on `user` as an attribute where `xsd:614` declares a child element.
For dubplate's own output the expected residual is zero, so any residual is a bug.

### The two candidate rules

**Legacy form** (the spec's first option). Validate the response against the unmodified XSD after removing every
OpenSubsonic addition.

- *What it needs:* the unmodified XSD, pinned by sha256 and treated as a workstation prerequisite beside the
  Unidecode wheel (the gate opens no network connection, `CLAUDE.md` "Gate"), or vendored. A projection step that
  drops the additions. Ampache's legacy mode (below) is a server switch, but dubplate has no legacy mode and should
  not grow one: a test-only response path would test something other than production.
- *The projection must be scoped by element type.* A projection that drops names globally breaks the core: the first
  prototype dropped every name any OpenAPI schema adds anywhere, and that removed `status` and `version` from the
  envelope (both are additions on other types: `PodcastEpisode.status`, `AlbumID3.version`) and failed every body it parsed
  (66) for the wrong reason. The scoped form (per element name: only the names the OpenAPI adds to that type) passed the
  counts in the table. The appendix script is that form.
- *What it proves:* the Subsonic-core part of the output (required attributes present, types right, structure right,
  list and nesting shapes) conforms to the one authoritative XML schema. The OpenSubsonic additions' XML shape is
  covered by nothing in this tier. It is covered only by the oracle-diff tier for the fields an oracle carries, and
  the JSON tier proves their values.
- *What it costs:* one projection script keyed on the pinned OpenAPI, no hand-maintained file, and an `xmllint`
  prerequisite. A name that is in neither the XSD nor the OpenAPI type is left in place and fails validation, so a
  misspelled optional attribute on a core type is caught, which a drop-everything-unknown projection would miss.

**Derived XSD** (the spec's second option). A copy of the XSD extended to admit the OpenSubsonic attributes and
elements.

- *What it needs:* the 49 names above declared with types, eight nested types declared, the root choice extended by
  at least seven elements, and a re-derivation at each OpenAPI bump (the specification adds fields often: five
  merged field or extension PRs between 2026-04 and 2026-07: #219, #228, #232, #257, #259).
- *What it proves:* that dubplate's XML encoder follows one shape rule across all types. The shape it checks is the one
  dubplate chose (E4), because no external authority defines an XML form for the additions, and the two oracles
  disagree on it (go's `isCompilation` element against nd's omission). A derived XSD therefore proves
  self-consistency, not conformance. If the derivation is generated from the OpenAPI by a rule ("scalar becomes an
  attribute, list becomes repeated elements"), the check is circular: the rule is the thing under test.
- *What it costs:* hand authoring or a generator, plus a published derivative of a file whose license is unknown
  (below).
- *What it buys that the legacy form lacks:* detection of a dubplate XML addition whose shape is inconsistent with
  its siblings. That value is real but small, because the additions have JSON as their primary form and the roster
  reads JSON (Task 2's evidence decides).

### How Ampache does it (method only, AGPL-3.0)

Studied at `ampache/ampache@3820b0658033`, described in my own words. Ampache keeps two committed corpora of captured
responses, one per mode, with 70 files each in both JSON and XML (`tests/Fixtures/Api/subsonic` and
`tests/Fixtures/Api/opensubsonic`). The server has a per-user "legacy" preference, and the capture script records one
corpus with it on and the other with it off. The conformance test validates the legacy-mode XML corpus against the
official 1.16.1 XSD (the XSD sits in the repository's `docs/` directory) and the OpenSubsonic-mode JSON corpus against the
OpenAPI file. It leaves the other two combinations unvalidated and states the reason in its header: no authoritative
machine-readable schema covers OpenSubsonic XML or any Subsonic JSON. A list of known deviations is asserted to still
fail, so an entry cannot rot. The idea transfers directly: validate only what an authority covers, record the rest as a
declared gap, and make the gap list self-expiring. The mechanism (a server legacy mode) does not transfer.

### The XSD's license

The file at `subsonic.org` carries no license header, and the page that links it states no license. The same file is
vendored (a 1.15.0 copy) in the Airsonic-Advanced repository (GPL-3.0,
`subsonic-rest-api/src/main/resources/subsonic-rest-api.xsd`) and in Ampache's AGPL repository. I found no separate license
for the file in either. Unconfirmed: whether dubplate may vendor
or modify it. What would confirm it: a statement from the Subsonic project, or a decision that the harness fetches the
unmodified file into a pinned cache (sha256 above) and never redistributes it. That decision favors the legacy form,
which needs no modified copy.

### Recommended reading for Task 7

**Adopt the legacy form, with the scoped projection derived from the pinned OpenAPI, and declare the OpenSubsonic
additions' XML shape as an unvalidated class.** Reasons: it proves exactly what an authority covers, it adds no
hand-maintained file, it needs no legal answer about a derivative, and its failure mode is a loud residual. State three
conditions in the decision: the projection is scoped per element type (the global form fails); a name found in neither
the XSD nor the OpenAPI type fails the test; and the XML shape of every addition is recorded once as dubplate's rule
(E4's convention) with the oracle diff covering the fields an oracle emits. Revisit if a roster client turns out to
parse XML. If Task 7 prefers the derived XSD, the cost is the 49 names, eight nested types, and the root-choice
extension above, and the legal question must be closed first.

---

## Q4. OpenSubsonic pull requests

Read through the GitHub API on 2026-10-07. The repository's `main` is still `bed1688` (2026-08-07), 21 PRs are open
(the survey counted 19 on 2026-09-25), and none of the four has merged. The newest merges are `topSongsByArtistId`
(#259, 2026-07-24) and `ArtistID3.disambiguation` (#257, 2026-07-16), both inside the pin.

| PR | Title | State on 2026-10-07 | Last activity | What it changes | Bearing on dubplate |
|---|---|---|---|---|---|
| #162 | Music Folder extension to extend `musicFolderId` support to more endpoints | Open, draft, mergeable state "blocked", 3 commits, 7 files | 2026-03-22 | New extension `musicFolder` v1. Requires `musicFolderId` on `getGenres`, `getSimilarSongs2`, `getTopSongs`, `getArtist`, `getShares`, `getBookmarks` (`Extensions/musicFolder.md`). Open questions in the body: playlists spanning folders, and whether the `getGenres` count language says "should" or "must" | Bears on spec "honored per endpoint" (`:374-376`). The thread splits on `getShares` and `getPlaylists`; a server with no folder support does not want the cost. No activity since March |
| #237 | Add collections | Open, ready, blocked, 15 commits, 33 files, +1816 | 2026-09-20 (author added item-uniqueness text) | New extension `collections` v1: `getCollections`, `getCollection`, `createCollection`, `updateCollection`, `deleteCollection`. A collection is "an ordered list of items - songs, artist, albums, playlists, etc." with unique items | The word clash and the `public` flag, below. Active; maintainers are asking design questions (duplicates, ordering) |
| #258 | Add extended library browsing extension | Open, draft, mergeable state "dirty", 1 commit, 40 files, +2661 | 2026-09-25 | New extension `extendedLibrary` v1: `getLibrarySongs`, `getLibraryAlbums`, `getLibraryArtists`, `getLibraryFacetValues`, JSON-only and POST-only with recursive filters, sorts, and cursor paging (`Extensions/extendedLibrary.md`) | Navidrome's and LMS's maintainers commented on 2026-09-21; one suggested opening a v2 discussion instead. Out of scope for 0.8 browse, but it is the first endpoint family with no XML and no GET form (E4) |
| #261 | Add public image URL extension | Open, ready, blocked, 1 commit, 8 files | 2026-07-20 | New extension `publicImageUrl`: `getPublicImageUrl`, converting a `getCoverArt` request to a link anyone can fetch | Collides with the refused feature "a link that opens a collection", and with the rule that no endpoint but one answers without credentials. Dubplate would not declare it. Maintainers questioned it as unsuitable for proxied deployments (2026-07-19, 2026-07-20) |

**Newer or adjacent PRs found** (open on 2026-10-07 unless noted):

- **#278** "Revisions to authentication section" (2026-10-02, ready): rewrites `api-reference.md` to list four auth
  methods, recommend API keys ("start here" for a new server), tell clients to prefer `formPost`, define `tokenInfo`
  for key holders, and have servers answer 41 or 42 for an unimplemented method. A member asked that nothing say
  "prefer plain password". Bears on the slice 5 and 6 credential decisions; not merged.
- **#255** `playlistFolders` (2026-09-27, 24 files): nested per-user playlist folders. The word "folder" again, on
  playlists.
- **#256** `artistRoleFilter` (2026-07-16): lets a client ask `getArtists` for all artists rather than album artists
  only. Its table says gonic and Navidrome return album artists only. The fixture has no track-only artist, so the oracles
  could not confirm or refute that here.
- **#202** scrobble clarification (2026-09-21): an invalid id in a `scrobble` batch answers error 70 and records no
  scrobble for the valid ids. Bears on slice 7's annotation endpoints.
- **#262** `songLyrics` v2 `line.end` (2026-08-26), **#184** `ratingID3` (2026-01-28), **#221** `search3Playlists`
  (2026-03-19), **#144** play-count split (2025-05-11): none touches folders, collections, or library browsing.
- **#230** `infiniteLibrary` (closed unmerged 2026-04-09, a library-browsing proposal for very large catalogs):
  closed by its author after he concluded that the target client dumps the whole catalog, which makes a paging patch moot.

### What #237's definition of "collection" means for dubplate's word

OpenSubsonic's collection is a per-user list: fields `id`, `name`, `owner`, `public`, `itemCount`, `created`,
`changed`, `coverArt`, `readonly` (`osapi` PR #237, `openapi/schemas/Collection.json`), whose items are songs, albums,
artists, playlists, genres, radio stations, or podcast items. Dubplate's collection is a library partition: a store
row whose root the server allocates from its id, mapped on the wire to a music folder (spec `:951-955`). If #237
merges, a client such as Symfonium would show two unrelated things under "collections": the member's lists and, by
whatever label the client uses for music folders, dubplate's partitions.

Three consequences for Task 7 (a naming and extension decision, not decided here):

1. Dubplate should not declare the `collections` extension at 0.8. Its `public` flag collides with the refused
   "public collections" (`CLAUDE.md`, "Refused features"), and playlists already cover the member's lists.
2. Dubplate's API-facing prose should say "music folder" when it names the OpenSubsonic parameter and "collection"
   only for the store concept, so a merged #237 changes no wire text.
3. If the merge happens and a client UI shows both, the rename option is "library" for dubplate's partition. That is a
   product taste call and goes to Geoff only if #237 merges before launch.

The re-read before the harness pins a build is owed to slice 2 (its plan's Task 1 already includes it).

---

## Q5. Error codes and text

The specification lists twelve codes. Text as stated at `osapi:content/en/docs/api-reference.md:156-167` and
`osapi:content/en/docs/Responses/error.md:35-48` (the two tables are identical):

| Code | Description |
|---|---|
| 0 | A generic error. |
| 10 | Required parameter is missing. |
| 20 | Incompatible Subsonic REST protocol version. Client must upgrade. |
| 30 | Incompatible Subsonic REST protocol version. Server must upgrade. |
| 40 | Wrong username or password. |
| 41 | Token authentication not supported for LDAP users. |
| 42 | Provided authentication mechanism not supported. |
| 43 | Multiple conflicting authentication mechanisms provided. |
| 44 | Invalid API key. |
| 50 | User is not authorized for the given operation. |
| 60 | The trial period for the Subsonic server is over. Please upgrade to Subsonic Premium. Visit subsonic.org for details. |
| 70 | The requested data was not found. |

Codes 42, 43, 44 and the `helpUrl` field are the OpenSubsonic additions (`Responses/error.md:31`,
`api-reference.md:170-175`). `message` is optional (`error.md:30`; the XSD declares it `use="optional"`, `xsd:637`),
and `code` is required (`xsd:636`). Code 41's text names LDAP, but "it actually implies that Token authentication is not
supported for _any_ reason" (`api-reference.md:172`).

**Errors 0, 50, and 70, as the specification states them.** Error 0 is "A generic error." Error 50 is "User is not
authorized for the given operation." Error 70 is "The requested data was not found." Dubplate's design uses error 50
for a `download` of any id kind but a song or an album, quoting that text (spec `:662-668`); error 50 is the right
code, because error 70 would be untrue for an id that exists.

**What the oracles did** (idle, token authentication):

| Situation | nd 0.64.2 | go v0.22.0 |
|---|---|---|
| Wrong password | 40 "Wrong username or password" | 40 "invalid password" |
| Unknown user | 40, same text | 40 `invalid username "nobody"` |
| Missing `u` | 10 "missing parameter: 'u'" | 10 `please provide a "u" parameter` |
| Missing `c` | 10 "missing parameter: 'c'" | 10 `please provide a "c" parameter` |
| Missing `id` on `getAlbum` | 70 "Album not found" | 10 ``please provide an `id` parameter`` |
| Well-formed id, no such record | 70 ("Album not found", "Song not found", "Artwork not found", "data not found" for `stream`) | 70 ("couldn't find an album with that id", and likewise for a track and an artist) |
| An id the server cannot parse (`nope`) | 70 | 10 (same message as a missing id) |
| Non-admin calling an admin endpoint (`startScan`, `getUsers`) | 50 "User is not authorized for the given operation" | not observed (go has no non-admin probe user) |
| `getUser` for an unknown username | 50, same text | returned the caller's own record |
| `download` of an artist id (idle) | HTTP 200 and a zip of the artist | 0 "error looking up id ar-1: not found" |
| Unknown endpoint | bare 404 text | 70 "view not found" |

Both servers use their own wording, so error text is not contract and the spec marks `message` optional; the code is.
nd's error 50 text omits the spec's final period, and nd's error 40 text omits its period too. **Consequences:** the
harness compares codes and the envelope shape, never message strings (a candidate divergence class, C-16). Navidrome
returns a zip for an artist `download`, which dubplate's design deliberately refuses with error 50 (C-17). The
precedence of error 10 against error 70 for a missing `id` differs between the oracles, so a missing required
parameter answers 10 (the spec's text) and is graded against the contract tier, not an oracle (C-18). Errors 20 and 30
were never observed (E9), and 41 to 44 were never observed either (no oracle supports API keys, Q1).

---

## Q6. Credential channels and the unauthenticated carve-out

**Channels the specification defines.** Credentials travel in the query string: `u`+`p` (clear or `enc:` hex), `u`+`t`+`s`
(token), or `apiKey` alone (`api-reference.md:16-29`; the OpenAPI security schemes at `openapi.json:380-436` all have
`in: query`). With `formPost` the same parameters travel in an `application/x-www-form-urlencoded` body
(`api-reference.md:33-41`). No header channel appears in the pinned pages. PR #278 adds only prose about these.

**Observed on nd and go** (ping, idle):

- Query `u`+`t`+`s`: accepted by both. Query `u`+`p` clear: accepted. `u`+`p=enc:<hex>`: accepted.
- Body: all parameters, auth included, in a form-encoded POST: accepted by both; authentication in the query with `f` in
  the body: accepted; a JSON POST body with the parameters in the query: accepted and ignored (both).
- Header: HTTP Basic: rejected by both with error 10.
- `apiKey`: unsupported by both (error 10).
- Conflicting parameters (`u`+`p`+`t`): nd accepted it as `ok` (it did not answer 43), go answered error 10. `apiKey`
  with `u`: nd answered 40, go answered 10. `t` without `s`: both answered 40. Neither implements the spec's error 43.
- Server-built URLs carry credentials on go (E11).

**The carve-out.** The spec requires `getOpenSubsonicExtensions` to answer with no credentials
(`getopensubsonicextensions.md:20`; `opensubsonic-changes.md:34`). Observed unauthenticated (no `u`, `t`, `s`, `v`,
`c`): nd answered `ok` with the extension list, in XML by default and in JSON for `f=json`, also when the credentials
were wrong, and also by form-encoded POST. go answered error 10 `please provide a "u" parameter`, and with valid
credentials but no `c` it answered error 10 `please provide a "c" parameter`, so go is outside the spec on this
endpoint. **Consequence:** dubplate matches nd, and the carve-out is the only route in the API that returns data
without credentials. Tests owed: the carve-out answers with no parameters at all (C-19), and a preflight answers with no data
(spec `:650-656`, C-20), so no other route leaks. The oracle assignment for the carve-out is nd only.

---

## Candidate ledger rows (input to Task 7)

Columns follow Task 7's ledger: id, area, class, oracle, oracle behavior, dubplate's intended behavior, evidence,
test owed. Oracle behavior is stated for both servers. "Class" is `new` unless it matches one of the spec's two
declared classes (`:285-287`: the folder view; the apiKey path).

| ID | Area | Class | Oracle behavior | Dubplate's intended behavior | Evidence | Test owed |
|---|---|---|---|---|---|---|
| C-01 | Extension list | new | nd declares 7 of 10; go declares 3 (authenticated only) | The launch subset Task 7 sets | Q1 | Harness: declared list equals the subset, name and versions |
| C-02 | `apiKeyAuthentication` | apiKey class | neither declares it | Per slice 6; graded against the spec and the roster | Q1, Q6 | Slice 6 |
| C-03 | Default format | new | both answer XML when `f` is absent, unknown, or mis-cased | The same; unknown `f` is never an error | E2 | Harness: five `f` forms |
| C-04 | JSONP callback | new | nd validates and answers error 0 for a bad callback; go reflects any callback | Validate the identifier; bad or missing answers error 0 as JSON (if JSONP is answered at all) | E3 | Harness: callback table |
| C-05 | JSONP content type | new | both `application/javascript`, nd no trailing `;`, go with `;` | `application/javascript`, with `X-Content-Type-Options: nosniff` | E3 | Harness |
| C-06 | XML shape of additions | new | both: scalars as attributes, lists as repeated elements; go: booleans as elements, nd omits empties | Dubplate's rule (E4's convention); declared unvalidated class | E4, Q3 | Oracle diff for emitted fields; no schema |
| C-07 | Single-element lists | new | both emit arrays | Slices everywhere, no length switch | E5 | Harness: one-element list per path |
| C-08 | Empty containers and null | new | nd empty object; go mixed and three `null` fields | Empty container, never `null` | E6 | Contract tier (OpenAPI rejects null) |
| C-09 | Emit-empty rule | new | nd emits empty added fields in JSON, omits in XML | Emit empty-valued additions in JSON; XML omits | E6 | Harness |
| C-10 | `musicFolder.id` integer | new | nd 1, go 0 | Integer wire id for a collection | E7 | Slice 3 test |
| C-11 | Unknown endpoint | new | nd bare 404 or 501 text; go error 70 envelope | Error 0 inside a well-formed envelope, per spec | E8 | Harness |
| C-12 | Binary-endpoint error | new | both envelope, `application/xml` or `application/json`, never `text/xml` | Envelope in the requested format; content type per the roster's needs (Task 2) | E8 | Roster |
| C-13 | Version negotiation | new | neither answers 20 or 30; nd requires `v`, go does not | Answer for any `v`; require `v` | E9 | Harness |
| C-14 | Escaping | new | both UTF-8 literal, HTML-safe JSON escapes, no XML declaration | Go standard encoders; compare parsed values | E10 | Harness |
| C-15 | Credentials in URLs | new | go `getArtistInfo2` image URLs carry `u`, `t`, `s` | Never echo request credentials into a response URL | E11 | Harness: scan every body for `t=`, `s=`, `p=`, `apiKey=` |
| C-16 | Error text | new | oracle wording differs from spec and from each other | Compare code and shape, never message | Q5 | Harness |
| C-17 | `download` of an artist id | new | nd zip, go error 0 | Error 50 (spec `:662-668`) | Q5 | Harness |
| C-18 | Missing required parameter | new | nd 70 for `getAlbum`, go 10 | Error 10; contract tier grades it | Q5 | Contract tier |
| C-19 | Extensions carve-out | new | nd open; go error 10 | Open, with no parameters at all | Q6 | Harness; nd only |
| C-20 | Preflight carries no data | new | nd and go answer a preflight with headers only | Same; no body, no credentials | Q6, E12 | Slice 3 CORS test |

---

## Routed

- **Task 2** (roster): whether any client sends `f=jsonp`, omits `f`, relies on a `text/xml` content type to detect a
  binary error, parses XML, or collapses singleton lists (E2, E3, E5, E8); the CORS observations (E12); the empty-success
  endpoint set against nd's bare 501 and go's error 70 (E8).
- **Task 3** (base URL): nd's `Set-Cookie ... Path=/` on `/rest` responses (E12); neither server accepted a header
  credential, so a proxy that injects one would not authenticate here.
- **Task 4** (silences): `transcoding` and `transcodeOffset` behavior under load (Q1); the oracle `download` zip.
- **Task 7** (synthesis): the XML rule, with the recommended reading above (Q3); the launch extension subset (Q1); the
  JSONP decision (E3); the word "collection" and the `collections` extension (Q4); the ledger rows above.
- **Slice 2** (harness): pin the XSD by sha256 and fetch it into a prerequisite cache; build the scoped projection from
  the appendix script's logic; re-read the OpenSubsonic PRs before pinning a build (Q4); the OpenAPI nulls and the
  required `NowPlaying.entry` mean the contract tier will flag both oracles, so oracle-flagged findings need the
  known-deviations treatment the Ampache method describes.
- **Slice 3** (response layer, catalog projection): encoding rules E1 to E10; the integer music-folder id (E7).
- **Slice 5 and 6** (identity, app passwords): PR #278 and the apiKey path; no oracle covers it (Q1, Q6).
- **Slice 7** (annotation): PR #202's scrobble batch rule (Q4).
- **Geoff:** nothing to rule. The `collections` word question reaches Geoff only if PR #237 merges before launch.

## Not done, and unconfirmed

- The XSD's license is unconfirmed (Q3). The conclusion above does not depend on it.
- The claim that the original Subsonic server collapsed single-element JSON lists is unconfirmed (E5).
- No oracle response was validated against the OpenAPI file in this task. The null and `entry` findings in E6 come from
  reading the schema files and the captured bodies. Slice 2's contract tier runs the real validation.
- The derived-XSD option was costed from the OpenAPI delta and not built. The scoped projection was prototyped and
  measured.
- XML bodies were sampled at 32 endpoints per server on a small, idle fixture. The counts in Q3 show the classes of
  failure, not the full coverage the harness will measure.
- Neither server exposed `apiKey`, `getPodcastEpisode`, or `sonicSimilarity`, so no oracle behavior exists for them.
- The Navidrome source and the gonic wiki's compatibility page were not read: every server claim here is from
  observation or from Apache-2.0 documentation.

## Prior art consulted

- OpenSubsonic specification, OpenAPI file, and the 1.16.1 XSD (adopted as the contract and harness inputs, spec
  `:866-867`): read at the pinned commit and the pinned XSD hash.
- Navidrome and gonic (adopted as oracles, spec `:868-869`): observed through containers; `navidrome/website`
  documentation quoted once (Apache-2.0).
- Ampache (study the method only, survey `:137`, `:161`): method described in Q3, no text copied.
- Neutral on hooks: this task builds nothing, and no new forward hook arose.

---

## Appendix: the scoped projection (this lane's own script, verbatim)

`legacy-projection.py <openapi-dir> <xsd-file> <in-dir> <out-dir> <scoped|envelope>` produced the counts in Q3.
Run it with `python3 -I` and validate with `xmllint --noout --schema <xsd-file> <file>`. The script reads the pinned OpenAPI
and XSD; it contains no copyleft text.

```python
#!/usr/bin/env python3
"""Legacy-form projection prototype (slice 1, Task 1; this lane's own script).

Usage: legacy-projection.py <openapi-dir> <xsd-file> <in-dir> <out-dir> <scoped|envelope>

Reads each XML response in <in-dir>, drops the OpenSubsonic additions, and writes the result to <out-dir>.
"scoped" drops, per element name, only the names the pinned OpenAPI schemas add to that type beyond the XSD.
"envelope" drops only the three envelope attributes. Validate the output with xmllint --schema <xsd-file>.
"""
import json
import os
import sys
import xml.etree.ElementTree as ET

XS = "{http://www.w3.org/2001/XMLSchema}"
NSU = "{http://subsonic.org/restapi}"
ET.register_namespace("", "http://subsonic.org/restapi")

OPENAPI, XSD, SRC, DST, MODE = sys.argv[1:6]

# OpenAPI schema file -> XSD complexType name, for the types a launch response carries.
MAPPING = {"Child": "Child", "AlbumID3": "AlbumID3", "AlbumID3WithSongs": "AlbumWithSongsID3",
           "ArtistID3": "ArtistID3", "ArtistWithAlbumsID3": "ArtistWithAlbumsID3",
           "Playlist": "Playlist", "PlaylistWithSongs": "PlaylistWithSongs"}


def load(path):
    with open(path) as f:
        return json.load(f)


def properties(node, base):
    """Property names of an OpenAPI schema node, following $ref and allOf."""
    names = set()
    if "$ref" in node:
        p = os.path.normpath(os.path.join(base, node["$ref"]))
        return properties(load(p), os.path.dirname(p))
    for sub in node.get("allOf", []):
        names |= properties(sub, base)
    names |= set(node.get("properties", {}))
    return names


def xsd_names(root, type_name):
    """Attribute and child-element names an XSD complexType declares, following extension."""
    ct = next(c for c in root.findall(XS + "complexType") if c.get("name") == type_name)
    names = {a.get("name") for a in ct.findall(XS + "attribute")}
    names |= {e.get("name") for e in ct.iter(XS + "element")}
    ext = ct.find(".//" + XS + "extension")
    if ext is not None:
        names |= xsd_names(root, ext.get("base").split(":")[-1])
    return names


def additions(xsd_root, oas, xt):
    base = os.path.join(OPENAPI, "schemas")
    return properties(load(os.path.join(base, oas + ".json")), base) - xsd_names(xsd_root, xt)


xroot = ET.parse(XSD).getroot()
d = {oas: additions(xroot, oas, xt) for oas, xt in MAPPING.items()}
child, artist = d["Child"], d["ArtistID3"] | d["ArtistWithAlbumsID3"]
album = (d["AlbumID3"] | d["AlbumID3WithSongs"]) - {"song"}
playlist = d["Playlist"] | d["PlaylistWithSongs"]
SCOPE = {"subsonic-response": {"type", "serverVersion", "openSubsonic"}, "error": {"helpUrl"},
         "album": album, "song": child, "child": child, "entry": child, "artist": artist, "playlist": playlist}


def project(el):
    if el.text is not None and not el.text.strip():
        el.text = None
    if el.tail is not None and not el.tail.strip():
        el.tail = None
    drop = SCOPE.get(el.tag.replace(NSU, ""), set())
    for a in [a for a in el.attrib if a in drop]:
        del el.attrib[a]
    for c in list(el):
        if c.tag.replace(NSU, "") in drop:
            el.remove(c)
        else:
            project(c)


os.makedirs(DST, exist_ok=True)
for name in sorted(os.listdir(SRC)):
    try:
        tree = ET.parse(os.path.join(SRC, name))
    except ET.ParseError:
        continue
    if MODE == "scoped":
        project(tree.getroot())
    else:
        for a in SCOPE["subsonic-response"]:
            tree.getroot().attrib.pop(a, None)
    tree.write(os.path.join(DST, name), encoding="utf-8")
print({k: len(v) for k, v in SCOPE.items()})
```
