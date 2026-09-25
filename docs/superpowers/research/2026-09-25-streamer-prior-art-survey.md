# Streamer prior-art survey (release 0.8)

Evidence date: 2026-09-25. Every license, release date, and importer count below was read on that date from the
GitHub API, the Go module proxy (`proxy.golang.org`), or `pkg.go.dev`. Four claims were proven by running code in a
scratch module: code generation from the OpenSubsonic OpenAPI file, response validation with kin-openapi, XSD validation
with xmllint, and FTS5 trigram search on `modernc.org/sqlite`. None of that code entered the repo.

## Brief

The survey leaves the spec's build shape standing. The study rung and the harness rung still come first, and vertical
slices still follow. It changes four things inside that shape. First, the contract tier works in pure Go for JSON: the
bundled OpenSubsonic OpenAPI file loads in kin-openapi and rejects a malformed `getAlbum` response. The official 1.16.1
XSD, however, rejects any XML response that carries OpenSubsonic's `type`, `serverVersion`, or `openSubsonic`
attributes, so the spec's "validate every response against the XSD" cannot hold as written. Second, OpenSubsonic added a
`transcoding` extension (`getTranscodeDecision`, `getTranscodeStream`) in December 2025. Navidrome implements it, and
Feishin and Symfonium call it. The streaming slice has to decide whether 0.8 answers it, because it moves profile choice
from the server's device table to a capability document the client posts. Third, Navidrome serves an in-flight transcode
with `Accept-Ranges: none`, while the spec serves ranges from a filling entry. That is a divergence the study rung must
declare, and `http.ServeContent` cannot serve a filling entry because it seeks to the end to learn the size. Fourth, no
OpenSubsonic conformance suite exists anywhere, and no maintained, permissively licensed Go Subsonic client exists, so the
harness builds both its validator wiring and its client. Two open cells close: FTS5 trigram with `remove_diacritics` works
on `modernc.org/sqlite` v1.59.0, and `djherbis/stream`, the fill-while-read primitive under Navidrome's cache, now has a
tagged release.

## Ranked short list

Ranked by how much each one changes or de-risks a slice.

1. **OpenSubsonic bundled OpenAPI plus `getkin/kin-openapi`**, for the harness rung's contract tier. Pin one bundled build
   by checksum, as Ampache does, and validate every JSON response in a Go test. Proven on `getAlbum`.
2. **The OpenSubsonic `transcoding` extension**, for the study rung and the streaming slice. Decide whether 0.8 ships
   `getTranscodeDecision` and `getTranscodeStream`, and record the answer in the launch extension subset.
3. **The Subsonic 1.16.1 XSD through `xmllint`**, for the harness rung. Validate only a legacy XML form, or a derived
   XSD that admits the OpenSubsonic attributes, and record the choice as a declared divergence.
4. **`djherbis/stream` v1.5.1**, for the streaming slice. Read it closely. Adopt it or port its design, decided by a race
   test at the streaming slice's plan. Record the filling-entry range behavior as a declared divergence from Navidrome.
5. **BeetstreamNext**, for the study rung. It is the only MIT server that serves an existing beets library over
   OpenSubsonic with no scanner of its own. Read it for the projection and for which clients it tests.
6. **Koel's new Subsonic layer**, for the streaming slice. MIT, 17,263 stars, Subsonic endpoints since 2026-05-27. Read
   it for the transcode cache it built and the MP4 `faststart` constraint it hit.
7. **`modernc.org/sqlite` v1.59.0 trigram search**, for the art-and-search slice. The spec's condition on this entry is
   half met. WAL under concurrent readers is still owed.
8. **Navidrome's documentation site**, for the study rung. It is Apache-2.0, unlike the server, so the study rung can
   quote its Subsonic compatibility page and its `.nsp` smart playlist format directly.
9. **`golang.org/x/image/draw` plus `gen2brain/webp` built with the `nodynamic` tag**, for the art slice. Without the tag
   the library loads a host `libwebp` when one is installed, so output would differ by host.
10. **`golang.org/x/crypto/acme/autocert`**, for the hosting slice's TLS listener. `caddyserver/certmagic` is the
    fallback if DNS challenges become a requirement.

## Findings that change the spec

These are the findings a later rung would be wrong to rediscover. Each names the spec section it touches.

1. **The XSD cannot validate OpenSubsonic XML.** "Evidence" says the contract tier validates every response against the
   OpenAPI schema and the 1.16.1 XSD. The OpenAPI file covers JSON only; its own documentation says it "Requires the use
   of `format=json`". The XSD rejects `openSubsonic="true"` and `type="..."` on `subsonic-response` (xmllint:
   "The attribute 'openSubsonic' is not allowed"). Ampache's test states the same gap: "OpenSubsonic deliberately extends
   the envelope beyond what the 1.16.1 XSD allows, and Subsonic never published a JSON schema." The harness rung needs a
   rule for XML: validate a legacy form that omits the OpenSubsonic attributes, or derive an XSD that admits them.
2. **The `transcoding` extension is live and used.** It first appeared in the spec repo on 2025-12-01. Navidrome
   implemented it on 2026-03-09 (commit "implement OpenSubsonic Transcoding extension (#4990)"). Feishin's
   `subsonic-api.ts` calls `getTranscodeDecision`, and a Feishin fix merged 2026-09-18 falls back to direct play when
   that POST fails a CORS check. The spec's profile order (request, device, estate default) has no place for a
   client-posted `ClientInfo`. The study rung should decide the extension's fate and, if it ships, how a posted capability
   document interacts with the device profile.
3. **Navidrome advertises seven or eight extensions, not the spec's assumption.** Its `GetOpenSubsonicExtensions` returns
   `transcodeOffset`, `formPost`, `songLyrics` (versions 1 and 2), `indexBasedQueue`, `transcoding`, `playbackReport`,
   `topSongsByArtistId`, and `sonicSimilarity` when a provider exists. It still omits `apiKeyAuthentication`, so the
   spec's identity reasoning stands. `playbackReport` has an `ignoreScrobble` flag that maps onto the spec's rule that a
   stream start is never a play. `indexBasedQueue` touches the saved-queue table.
4. **Ranges on a filling transcode diverge from the oracle.** Navidrome's `core/stream/media_streamer.go` uses
   `http.ServeContent` only for a seekable (complete) stream. For anything else it sets `Accept-Ranges: none`, and it sends
   an estimated `Content-Length` only when the client passes `estimateContentLength=true`. The spec serves a Range request
   from a filling entry. That needs a `Content-Range` with an unknown total, and `http.ServeContent` cannot produce one: its
   documentation says it "uses a seek to the end of the content to determine its size". `djherbis/stream`'s `Seek` to the
   end blocks until the stream closes unless `SetSeekEnd` is called. The streaming slice owes a declared divergence class
   or a change to match the oracle. Seeking during a transcode may rest on `timeOffset` (`transcodeOffset`) instead.
5. **Browser clients need CORS.** Feishin's web build and Airsonic Refix run in a browser. Navidrome depends on
   `go-chi/cors`. The spec does not mention CORS. The transcoding extension's JSON POST triggers a preflight. The API
   surface section should state the CORS policy, including preflight for `application/json` bodies.
6. **Stream-through transcoding constrains the container.** Koel encodes AAC into MP4 with `-movflags +faststart`, which
   only works on a finished file, so Koel encodes the whole track before the first byte. Stream-through suits MP3 and
   Ogg/Opus. An MP4/AAC profile needs fragmented MP4 or ADTS output. The profile templates should say which containers
   stream-through admits.
7. **Small corrections to the prior-art table.** Ampache's test is `tests/Module/Api/SubsonicSpecConformanceTest.php`, not
   `OpenSubsonicSpecConformanceTest.php`. `modernc.org/sqlite` is at v1.59.0 (2026-09-15), which the repo already pins.
   `golang.org/x/image` is at v0.46.0 (2026-09-08). `djherbis/fscache`'s last tag is v0.10.1, dated 2020-04-15 by the module
   proxy. `gen2brain/webp` is libwebp transpiled to Go through `wasm2go`, not a wazero runtime, and it prefers a host
   shared library through `purego` unless built with `nodynamic`.
8. **A name collision is coming.** OpenSubsonic PR #237, "Add collections" (open, 51 comments, updated 2026-09-20), defines
   a collection as a user-owned ordered list of items. dubplate's collection is a music folder. If the PR lands, API docs
   and UI copy need a distinct word for one of them. PR #162 (open) would extend `musicFolderId` to more endpoints, which
   bears on the spec's "honored per endpoint" rule.
9. **Trigram search needs the prefix pass.** A two-character `MATCH` against a trigram table returned no rows in the
   scratch test. The spec's separate prefix lookup is therefore required for short queries, not an optimization.

## 1. Machine-readable API definitions

| Candidate | License (evidence) | Last change, cadence | Completeness and currency | Go wire types? | Verdict | Serves |
|---|---|---|---|---|---|---|
| OpenSubsonic OpenAPI 3.0.0, split source under `openapi/` in [opensubsonic/open-subsonic-api](https://github.com/opensubsonic/open-subsonic-api) | Apache-2.0 ([LICENSE](https://github.com/opensubsonic/open-subsonic-api/blob/main/LICENSE)) | Last schema commit 2026-07-24 (`topSongsByArtistId`); HEAD `bed1688`, 2026-08-07. No releases or tags; the site rebuilds on every merge. | 87 paths, matching the 87 endpoint pages in the docs, including `getTranscodeDecision`, `reportPlayback`, `tokenInfo`, and `hls.m3u8`. 195 component schemas. Current to `ArtistID3.disambiguation` (2026-07-16). JSON only; XML appears only on binary and HLS responses. Its own page carries a "WIP" warning about docs-versus-schema inconsistencies. Seven docs pages still carry `// TODO` examples, all in video and legacy `search`. | Yes, with limits. `oapi-codegen` v2 generated 10,605 lines of types that pass `go vet`. The fields are pointers, there are no `xml` tags, and each `subsonic-response` is a `json.RawMessage` union. Unfit as production wire types; usable in the harness. | Adopt as a harness input. Pin a bundled build by checksum. Never vendor into shipped code. | Harness rung, contract tier |
| Bundled build at `https://opensubsonic.netlify.app/docs/openapi/openapi.json` | Same repo | Rebuilt per merge by `swagger-cli bundle` and linted by `vacuum` (`package.json`, `build:openapi`) | 453,720 bytes on 2026-09-25; no external `$ref` left. | Same as above | Adopt as the pinned artifact. Ampache pins it "to one build rather than latest" and records the audit date. | Harness rung |
| Subsonic REST API XSD 1.16.1, `http://www.subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd` | No license text in the file; Subsonic is proprietary. Ampache commits a copy as `docs/subsonic-rest-api-1.16.1.xsd.xml`. | Frozen with Subsonic 1.16.1 | 62 complex types. Rejects OpenSubsonic attributes (finding 1). | `xuri/xgen` could generate XML types; declined below. | Adopt as a harness input fetched at test setup, not committed, until the license question is settled. | Harness rung |
| OpenSubsonic docs pages (`content/en/docs/Endpoints`, `Responses`, `Extensions`) | Apache-2.0, same LICENSE | Same as the repo | The prose source of truth when schema and docs disagree. 10 extension pages. | n/a | Adopt as the study rung's reading list. Quote, never vendor. | Study rung |
| `oapi-codegen/oapi-codegen` | Apache-2.0 ([LICENSE](https://github.com/oapi-codegen/oapi-codegen/blob/main/LICENSE)) | 6 releases in 12 months, last 2026-07-17 | 317 open issues and PRs; two main maintainers | Proven: types generate cleanly | Optional: generate a typed client for the harness driver only. Decline for production types. | Harness rung |
| `ogen-go/ogen` v1.24.0 | Apache-2.0 ([LICENSE](https://github.com/ogen-go/ogen/blob/main/LICENSE)) | 15 releases in 12 months | Two main maintainers | Fails on the full spec ("complex form schema is not implemented"). With `ignore_not_implemented: ["all"]` it generates but skips the `formPost` operations. | Decline. | n/a |
| `xuri/xgen` (XSD to Go) | BSD-3-Clause ([LICENSE](https://github.com/xuri/xgen/blob/master/LICENSE)) | No tagged releases; last push 2026-09-02 | 35 open issues and PRs; one maintainer | Not run | Decline. The XSD is legacy-only, and hand-written response structs are the spec's plan. | n/a |

## 2. Servers

### Permissively licensed servers

| Server | License (evidence) | Language | Last release, cadence | Open issues and PRs | Bus factor | Adoption | Code-quality note | Verdict | Serves |
|---|---|---|---|---|---|---|---|---|---|
| [Melodee](https://github.com/melodee-project/melodee) | MIT ([LICENSE](https://github.com/melodee-project/melodee/blob/main/LICENSE)) | C# | 2026-07-11; 7 releases in 12 months | 1 | 1 (sphildreth 1,383 commits; next is Copilot at 94) | 66 stars; publishes an [endpoint matrix](https://github.com/melodee-project/melodee/blob/main/docs/pages/opensubsonic-matrix.md) dated 2026-07-11 | `StreamingLimiter.cs` is plain and readable, with optimistic increment and rollback. Its schema validator is hand-written, not driven by the upstream schema. It returns HTTP 410 or 501 for unsupported endpoints, where the spec returns error code 0. | Read closely, as the spec already says. Reuse permitted. | Study rung; streaming slice |
| [Koel](https://github.com/koel/koel) | MIT ([LICENSE](https://github.com/koel/koel/blob/master/LICENSE)) | PHP | 2026-09-25; 37 releases in 12 months | 11, oldest 2024-05-15 | 1 (phanan 3,159 commits) | 17,263 stars | Subsonic controllers since 2026-05-27, one class per endpoint. Advertises `songLyrics` and `transcodeOffset` only. `Transcoder.php` builds an argv array with no shell, encodes AAC with `faststart`, and caches by (song, bitrate) with a checksum check. | Read closely for transcode-cache lessons (finding 6). | Streaming slice |
| [BeetstreamNext](https://github.com/FlorentLM/BeetstreamNext) | MIT ([LICENSE](https://github.com/FlorentLM/BeetstreamNext/blob/main/LICENSE)) | Python | 2026-09-16; 1 release in 12 months | 0 | 1 (FlorentLM 834 commits; BinaryBrain, the original Beetstream author, 96) | 7 stars | A beets plugin that projects the beets database to OpenSubsonic. Its docs claim "essentially all" of the API and list direct play, on-the-fly transcoding, HLS, and server-side ReplayGain. Source not read beyond the README and docs. | Read closely. It is the nearest analog to "the catalog is the index" and to rung 16's beets import. | Study rung; catalog projection slice |
| [Qm-Music](https://github.com/chenqimiao/qm-music) | Apache-2.0 ([LICENSE](https://github.com/chenqimiao/qm-music/blob/main/LICENSE)) | Java | 2026-09-10; 11 releases in 12 months | 1 | 1 (chenqimiao 607 commits) | 581 stars; on the OpenSubsonic server list | `FFmpegStreamUtils.java` is a static utility class with a thread pool on an unbounded queue. Below the house bar. | Decline. | n/a |
| [smolsonic](https://github.com/tsirysndr/smolsonic) | MIT ([LICENSE](https://github.com/tsirysndr/smolsonic/blob/main/LICENSE)) | Rust | 2026-07-16; 17 releases in 12 months | 0 | 1 | 17 stars | Not read. Serves Subsonic, Jellyfin, and S3 APIs. | Decline; too young to teach. | n/a |
| [peerphonic](https://github.com/randomtoy/peerphonic) | MIT ([LICENSE](https://github.com/randomtoy/peerphonic/blob/main/LICENSE)) | Go | No releases; last push 2026-08-19 | 0 | 1 | 4 stars | `internal/streaming/ffmpeg.go` is clean Go with a bitrate allowlist. It calls `cmd.Wait` in a goroutine while the caller reads `StdoutPipe`, which the `os/exec` docs call incorrect ("It is thus incorrect to call Wait before all reads from the pipe have completed"). | Decline. Record the `StdoutPipe` lesson: give ffmpeg's stdout the cache writer directly. | Streaming slice |
| [nina](https://github.com/evilvic/nina), [tinysonic](https://github.com/tsirysndr/tinysonic), [ubersonic](https://github.com/simonvetter/ubersonic) | MIT (LICENSE links: [nina](https://github.com/evilvic/nina/blob/main/LICENSE), [tinysonic](https://github.com/tsirysndr/tinysonic/blob/main/LICENSE), [ubersonic](https://github.com/simonvetter/ubersonic/blob/master/LICENSE)) | Go | No releases; ubersonic last pushed 2017 | 0 | 1 each | 0 to 16 stars | Not read. | Decline. | n/a |
| [subfin-plugin](https://github.com/williamkray/subfin-plugin) | MIT ([LICENSE](https://github.com/williamkray/subfin-plugin/blob/main/LICENSE)) | C# | 2026-09-21; 2 releases in 12 months | 0 | 1 | 5 stars | An OpenSubsonic layer as a Jellyfin plugin. Not read. | Decline. The spec already rejected a Jellyfin bridge. | n/a |

Non-Subsonic permissive servers worth reading for one design point:

| Server | License (evidence) | Language | Last release, cadence | Open issues and PRs | Bus factor | Adoption | Design point | Verdict | Serves |
|---|---|---|---|---|---|---|---|---|---|
| [Black Candy](https://github.com/blackcandy-org/blackcandy) | MIT ([LICENSE](https://github.com/blackcandy-org/blackcandy/blob/master/LICENSE)) | Ruby | 2026-08-26; 2 releases in 12 months | 29, oldest 2019-07-04 | 1 | 4,420 stars | `transcoded_stream_controller.rb` tees ffmpeg output to the client and to a cache file keyed on the file's MD5 and bitrate. It writes the cache in place with no rename, then trusts an entry whose duration is within two seconds of the source. | Study for a negative lesson. It confirms content keying and shows why a failed encode must be evicted, not guessed at. | Streaming slice |
| [Polaris](https://github.com/agersant/polaris) | MIT ([LICENSE](https://github.com/agersant/polaris/blob/master/LICENSE)) | Rust | 2026-05-08; 2 releases in 12 months | 28, oldest 2020-04-21 | 1 | 2,732 stars | `src/app/thumbnail.rs` keys thumbnails on a `DefaultHasher` hash of path and options. That hash is not stable across Rust releases and changes when a file moves. | Study for a negative lesson. It supports keying art variants on the art hash. | Art slice |

### Copyleft servers: study for behavior only

Each line cites the LICENSE the GitHub API reports. None may be copied or transliterated.

- Navidrome, GPL-3.0 ([LICENSE](https://github.com/navidrome/navidrome/blob/master/LICENSE)). v0.64.2 on 2026-09-24, 15 releases in 12 months. The primary oracle. Its `server/subsonic/e2e` Ginkgo suite shows which behaviors its author tests.
- gonic, GPL-3.0 ([LICENSE](https://github.com/sentriz/gonic/blob/master/LICENSE)). 2026-06-08, 5 releases in 12 months. The folder-view oracle. It keeps 129 golden fixtures under `server/ctrlsubsonic/testdata` and diffs them with `josephburnett/jd`.
- Ampache, AGPL-3.0 ([LICENSE](https://github.com/ampache/ampache/blob/develop/LICENSE)). 2026-09-15, 17 releases in 12 months. Its fixture capture script and known-deviations list are the pattern for the contract tier; reimplement the idea.
- LMS, GPL-3.0 ([LICENSE](https://github.com/epoupon/lms/blob/master/LICENSE)). 2026-09-18, 11 releases in 12 months.
- Supysonic, AGPL-3.0 ([LICENSE](https://github.com/spl0k/supysonic/blob/master/LICENSE)). Last release 2024-12-15.
- Airsonic, GPL-3.0, archived ([repo](https://github.com/airsonic/airsonic)); Airsonic-Advanced, GPL-3.0 ([LICENSE](https://github.com/airsonic-advanced/airsonic-advanced/blob/master/LICENSE)), last release 2024-04-24.
- Funkwhale, AGPL-3.0 ([LICENSE](https://dev.funkwhale.audio/funkwhale/funkwhale/-/raw/develop/LICENSE)).
- Nextcloud Music, AGPL-3.0 ([LICENSE](https://github.com/nc-music/music/blob/master/LICENSE)).
- Jellyfin, GPL-2.0 ([LICENSE](https://github.com/jellyfin/jellyfin/blob/master/LICENSE)).
- mStream, GPL-3.0 ([LICENSE](https://github.com/IrosTheBeggar/mStream/blob/master/LICENSE)).
- EdgeSonic, AGPL-3.0 ([LICENSE](https://github.com/wuyilingwei/edgesonic/blob/main/LICENSE)).
- nghe (Rust), AGPL-3.0 ([LICENSE](https://github.com/vnghia/nghe/blob/main/LICENSE)); waveflow-server (Rust), AGPL-3.0 ([repo](https://github.com/InstaZDLL/waveflow-server)); crossonic-server (Go), AGPL-3.0 ([LICENSE](https://github.com/juho05/crossonic-server/blob/main/LICENSE)); AudioMuse-AI-MusicServer (Go), AGPL-3.0 ([repo](https://github.com/NeptuneHub/AudioMuse-AI-MusicServer)).

Navidrome's documentation is a separate repository under Apache-2.0
([navidrome/website](https://github.com/navidrome/website), license per the GitHub API). Its
`content/en/docs/developers/subsonic-api.md` and `content/en/docs/usage/features/smart-playlists/index.md` can be quoted
for behavior and for the `.nsp` format without GPL exposure.

## 3. Conformance suites and test drivers

No conformance suite exists. The OpenSubsonic repo has no test tooling beyond `vacuum` linting of its own schema. It has
issues turned off and discussions turned on; 19 PRs are open. A repository search for OpenSubsonic or Subsonic
conformance, compliance, or test-suite projects found none. Server authors test their own servers.

| Candidate | License (evidence) | What it is | Verdict | Serves |
|---|---|---|---|---|
| Ampache `SubsonicSpecConformanceTest.php` and `capture_subsonic_fixtures.php` | AGPL-3.0 | Captured fixtures validated against the pinned OpenAPI (JSON) and the XSD (legacy XML). A `KNOWN_DEVIATIONS` list is asserted to still fail, so it cannot rot. Its entries seed a divergence ledger: the `User` schema lacks `email`; an empty `PlayQueue` cannot satisfy its own required fields; `getTranscodeDecision` declares its schema inline and escapes `$ref` lookup. | Study for the method. Reimplement in Go. | Harness rung |
| Melodee `tests/Melodee.Tests.OpenSubsonic` and `scripts/test-opensubsonic-api.sh` | MIT | Per-area endpoint tests and a curl walk with token auth. | Read. Reuse permitted, but the Go harness will not share a runtime with it. | Harness rung |
| Navidrome `server/subsonic/e2e` | GPL-3.0 | Ginkgo end-to-end suite over a fake filesystem library, split by API area (browsing, search, stream, transcode, sharing, multi-user). | Study the area list and fixture design only. | Study rung |
| Symfonium's "Supported features by media providers" page ([docs.symfonium.app](https://docs.symfonium.app/wiki/providers/supported-features-by-media-providers/)) | Documentation; Symfonium is proprietary | Says OpenSubsonic support "varies by server and the extensions it implements". Names no per-extension detail. | Study input, as the spec says. Thinner than expected. | Study rung |
| Feishin `src/shared/api/subsonic/subsonic-types.ts` | GPL-3.0 ([repo](https://github.com/jeffvli/feishin)) | Zod schemas for what a major client parses, including `getTranscodeDecision` and `getTranscodeStream` parameters. | Study for which fields a client requires. | Study rung |
| `kin-openapi` `openapi3filter.ValidateResponse` ([getkin/kin-openapi](https://github.com/getkin/kin-openapi)) | MIT ([LICENSE](https://github.com/getkin/kin-openapi/blob/master/LICENSE)) | v0.149.0 on 2026-08-28; 16 releases in 12 months; 481 importers of `openapi3filter`; 111 open issues and PRs; maintainers fenollp and reuvenharrison. Loaded the bundled spec with no `doc.Validate` error, flagged a missing required `song` array, and flagged a numeric `id`. The error paths are precise. | Adopt as a test-only dependency. | Harness rung, contract tier |
| `pb33f/libopenapi-validator` | MIT ([LICENSE](https://github.com/pb33f/libopenapi-validator/blob/main/LICENSE)) | v0.14.0 on 2026-07-11; 32 releases in 12 months; 27 importers; one main maintainer. Not run. | Record as the fallback to kin-openapi. | Harness rung |
| `santhosh-tekuri/jsonschema` v6 | Apache-2.0 ([LICENSE](https://github.com/santhosh-tekuri/jsonschema/blob/boon/LICENSE)) | v6.0.3; 241 importers; Navidrome uses it. Validates schemas, not routed responses. | Decline in favor of kin-openapi, which routes the response to its schema. | n/a |
| `xmllint` from libxml2 | MIT (libxml2) | Installed on the workstation through Homebrew, libxml 2.15.4. Validated a minimal envelope and rejected the OpenSubsonic attributes. | Adopt as a workstation tool the harness names and fails without, like the gate's `rclone` leg. | Harness rung |
| `lestrrat-go/libxml2`, `terminalstatic/go-xsd-validate` | MIT (LICENSE links: [libxml2](https://github.com/lestrrat-go/libxml2/blob/master/LICENSE), [go-xsd-validate](https://github.com/terminalstatic/go-xsd-validate/blob/master/LICENSE)) | Both cgo. `lestrrat-go/libxml2` is archived. | Decline; cgo would enter the module. | n/a |
| `josephburnett/jd` v2 | MIT ([LICENSE](https://github.com/josephburnett/jd/blob/master/LICENSE)) | v2.5.0 on 2026-02-23; 7 importers of v2; one maintainer. gonic's golden differ. | Decline. The oracle differ compares search as a set plus ordering, which a generic JSON diff does not express. | n/a |
| Go Subsonic clients: `delucks/go-subsonic`, `supersonic-app/go-subsonic` | GPL-3.0 (LICENSE links: [delucks](https://github.com/delucks/go-subsonic/blob/master/LICENSE), [supersonic-app](https://github.com/supersonic-app/go-subsonic/blob/master/LICENSE)) | The only maintained Go clients. Supersonic pins an untagged commit from 2026-04-16. | Decline, even as a test driver: a test binary that links GPL code is still a derivative. | n/a |
| `mdlayher/gosubsonic` | MIT ([repo](https://github.com/mdlayher/gosubsonic)) | Last push 2017-11-19. | Decline. Write the harness client; token auth is `md5(password + salt)` and the rest is `net/http`. | Harness rung |

## 4. Go libraries for the streaming path

### ffmpeg orchestration and transcoding

| Candidate | License (evidence) | Last release, cadence | Importers | Quality note | Verdict |
|---|---|---|---|---|---|
| stdlib `os/exec` | BSD-3 | Go 1.27.1 | n/a | `CommandContext` with `Cmd.Cancel` and `Cmd.WaitDelay` covers a graceful stop and a hard kill. Setting `cmd.Stdout` to the cache writer avoids the `StdoutPipe` and `Wait` ordering trap. | Adopt, as the spec says. Streaming slice. |
| `u2takey/ffmpeg-go` | Apache-2.0 ([LICENSE](https://github.com/u2takey/ffmpeg-go/blob/master/LICENSE)) | v0.5.0, 2023-08-08; no release since | 347 | A fluent builder ported from Python's ffmpeg-python. Its `go.mod` directly requires `aws-sdk-go` and `gocv`. | Decline. |
| `floostack/transcoder` | MIT ([LICENSE](https://github.com/floostack/transcoder/blob/master/LICENSE)) | 2020-11-14; archived | not read | n/a | Decline. |
| `xfrr/goffmpeg` | MIT ([LICENSE](https://github.com/xfrr/goffmpeg/blob/master/LICENSE)) | v1.0.0, 2023-09-29 | 0 | n/a | Decline. |

### HTTP range serving and the stream-through cache

| Candidate | License (evidence) | Last release, cadence | Importers | Quality note | Verdict |
|---|---|---|---|---|---|
| stdlib `http.ServeContent` | BSD-3 | Go 1.27.1 | n/a | Handles Range, `If-Range`, and ETags on any `io.ReadSeeker` of known size. Cannot serve a filling entry (finding 4). | Adopt for the direct path and the cache-hit path. |
| [`djherbis/stream`](https://github.com/djherbis/stream) | MIT ([LICENSE](https://github.com/djherbis/stream/blob/master/LICENSE)) | v1.5.1, 2026-08-29; 2 releases in 12 months, both added for Navidrome's error propagation (`CancelWithErr`) | 7 | About 1,800 lines with tests. `Reader` has `ReadAt` that blocks until the offset is written, `Seek`, and `Size`. `Stream.CancelWithErr` delivers an encode failure to attached readers, which is the spec's "attached readers see the error". Bus factor 1 (djherbis 42 commits); Navidrome's maintainer lands fixes. | Read closely. Adopt or port at the streaming slice, decided by a race test on cancel, disconnect, and failed exit. It is the primitive the spec's F2 hand-roll would rebuild. |
| `djherbis/fscache` | MIT ([LICENSE](https://github.com/djherbis/fscache/blob/master/LICENSE)) | Last tag v0.10.1, 2020-04-15; Navidrome pins pseudo-version `v0.10.2-0.20260829235704` | 7 | Adds LRU "haunters" and reapers on top of `stream`. | Keep the spec's decline. The size cap and eviction stay in-repo. |
| `peterbourgon/diskv` | MIT ([LICENSE](https://github.com/peterbourgon/diskv/blob/master/LICENSE)) | v2.0.1, 2017-08-14 | 659 | Key-value on disk; no concurrent readers of a partial write. | Decline. |
| `hashicorp/golang-lru` v2 | MPL-2.0 ([LICENSE](https://github.com/hashicorp/golang-lru/blob/main/LICENSE)), noted separately as weak copyleft | v2.0.7, 2023-09-21 | 1,261 | In-memory only. | Decline. The cache index belongs in SQLite so it survives a restart. |
| `golang.org/x/sync/singleflight` | BSD-3 ([LICENSE](https://github.com/golang/sync/blob/master/LICENSE)) | v0.23.0, 2026-08-31 | 3,802 | Deduplicates concurrent callers of one key. | Optional. It covers "a second listener shares the encode" only for the start of an encode, not for readers attaching later. |

### Image resizing and encoding

| Candidate | License (evidence) | Last release, cadence | Importers | Quality note | Verdict |
|---|---|---|---|---|---|
| `golang.org/x/image/draw` | BSD-3 ([LICENSE](https://github.com/golang/image/blob/master/LICENSE)) | v0.46.0, 2026-09-08; monthly | 1,948 | Go team. `CatmullRom` and `ApproxBiLinear` scalers. | Adopt, as the spec says. Art slice. |
| [`gen2brain/webp`](https://github.com/gen2brain/webp) | MIT ([LICENSE](https://github.com/gen2brain/webp/blob/main/LICENSE)); wraps libwebp (BSD-3) | v0.6.4, 2026-07-01; 31 tags | 18 | libwebp transpiled to Go through `wasm2go`. It loads a host shared library through `purego` first unless built with `nodynamic`. Its only direct dependency is `ebitengine/purego`. Navidrome depends on it. Bus factor 1. | Adopt with the `nodynamic` build tag so output is identical on every host. |
| [`gen2brain/vpx`](https://github.com/gen2brain/vpx) | BSD-3-Clause ([LICENSE](https://github.com/gen2brain/vpx/blob/main/LICENSE)) | v0.2.1, 2026-08-18; 3 tags | not published on pkg.go.dev at survey time | A pure-Go port of libwebp, "byte-exact against it", with SIMD. Same author. | Watch. Revisit at the art slice's plan; it would drop the transpiled C. |
| [`HugoSmits86/nativewebp`](https://github.com/HugoSmits86/nativewebp) | MIT ([LICENSE](https://github.com/HugoSmits86/nativewebp/blob/main/LICENSE)) | v1.3.0, 2026-05-10 | 20 | Pure Go, but "supports only WebP lossless images (VP8L)". | Decline. Lossless WebP is the wrong format for cover thumbnails. |
| `gen2brain/avif` | MIT ([LICENSE](https://github.com/gen2brain/avif/blob/main/LICENSE)) | v0.6.0, 2026-07-05 | 24 | libavif with aom and dav1d under wazero by default. | Decline for 0.8. No evidence found that Subsonic clients request AVIF. |
| `disintegration/imaging` | MIT ([LICENSE](https://github.com/disintegration/imaging/blob/master/LICENSE)) | v1.6.2, 2019-11-16 | 2,945 | gonic's choice; unmaintained. | Decline. |
| `nfnt/resize` | ISC ([LICENSE](https://github.com/nfnt/resize/blob/master/LICENSE)) | Archived | not read | n/a | Decline. |
| `anthonynsimon/bild` | MIT ([LICENSE](https://github.com/anthonynsimon/bild/blob/master/LICENSE)) | v0.17.1, 2026-09-07; 5 releases in 12 months | 224 | Broad image-effects library. | Decline; `x/image/draw` covers resize. |
| `davidbyttow/govips`, `chai2010/webp`, `kolesa-team/go-webp` | MIT, BSD-3-Clause, MIT | Various | 116 for govips | All cgo. | Decline. |

### ReplayGain and loudness

The gain values are computed at the current ladder's rung 9, "probe-backed" per the organizer spec, and land on the
catalog row under forward hook 11. The streamer only reads them. OpenSubsonic's `Child.replayGain` carries `trackGain`,
`albumGain`, `trackPeak`, `albumPeak`, `baseGain`, and `fallbackGain`, so most clients apply gain themselves. Server-side
application is an ffmpeg `volume` filter in a profile template, which the spec already allows.

| Candidate | License (evidence) | Last release | Importers | Verdict |
|---|---|---|---|---|
| ffmpeg `ebur128` and `volume` filters | Subprocess; not linked | n/a | n/a | Adopt as the only loudness path in the streamer. |
| `exaring/ebur128` | MIT (GitHub API) | Last push 2026-02-17; 1 star | not read | Decline. Too new to judge. |
| `rsgain` | BSD-style license text ([LICENSE](https://github.com/complexlogic/rsgain/blob/master/LICENSE)); GitHub reports NOASSERTION | 2026-09-02; 2 releases in 12 months | 700 stars | Record as the rung 9 reference tool if a probe is chosen. Not a streamer concern. |

### FTS5 trigram with modernc sqlite

`modernc.org/sqlite` v1.59.0 bundles SQLite 3.53.4. In the scratch test, `CREATE VIRTUAL TABLE ... USING
fts5(name, tokenize="trigram remove_diacritics 1")` succeeded. The query `MATCH 'bjork'` found "Björk". A two-character
`MATCH` found nothing (finding 9). License: BSD-3 ([LICENSE](https://gitlab.com/cznic/sqlite/-/raw/master/LICENSE)).
3,518 importers; 169 tagged versions. The spec's rung 6 condition still owes the WAL check under concurrent readers.
`remove_diacritics` does not replace the spec's unidecode fold, which also handles ligatures and non-Latin scripts.

### Playlist formats

| Candidate | License (evidence) | Last release | Importers | Verdict |
|---|---|---|---|---|
| Hand-rolled M3U | n/a | n/a | n/a | Keep the spec's hand-roll. |
| `grafov/m3u8` | BSD-3-Clause ([LICENSE](https://github.com/grafov/m3u8/blob/master/LICENSE)) | v0.12.1, 2024-12-08; archived | 281 | Decline. It is HLS, and it is archived. |
| XSPF libraries (`sir-go/iptv-playlist`, `Cycloctane/xplay`, `arxdsilva/m3uspiff_go`) | MIT, MIT, Apache-2.0 (GitHub API) | 0 to 4 stars | not read | Decline. `encoding/xml` covers XSPF if it is ever needed; no slice needs it. |

HLS: the OpenSubsonic spec has an `hls.m3u8` endpoint and an `hls` protocol value in the transcoding extension.
Navidrome validates `hls` as a protocol but marks it "TODO: derive from decision when HLS support is added" in
`server/subsonic/transcode.go`. BeetstreamNext serves AAC HLS. HLS belongs on the 2.1 list if anywhere.

### Scrobbling clients (2.1 list; noted, not ranked)

- ListenBrainz: no maintained Go client library was found. The GitHub hits are applications (Koito, Explo, Navidrome
  plugins by kgarner7), all MIT. The ListenBrainz submit API is one authenticated JSON POST.
- Last.fm: `shkh/lastfm-go`, MIT ([LICENSE](https://github.com/shkh/lastfm-go/blob/master/LICENSE)), untagged, last
  commit 2019-12-15, 57 importers. `twoscott/gobble-fm`, MIT ([LICENSE](https://github.com/twoscott/gobble-fm/blob/master/LICENSE)),
  v1.0.9 on 2026-01-28, 1 importer.

## 5. Other findings for the build shape

| Candidate | License (evidence) | Last release, cadence | Importers | Note | Verdict | Serves |
|---|---|---|---|---|---|---|
| `golang.org/x/crypto/acme/autocert` | BSD-3 ([LICENSE](https://github.com/golang/crypto/blob/master/LICENSE)) | v0.57.0, 2026-09-08; monthly | 3,292 | Go team. TLS-ALPN and HTTP-01 challenges. Not read beyond the package list. | Adopt candidate for the TLS listener. | Hosting slice; hook 19 |
| `caddyserver/certmagic` | Apache-2.0 ([LICENSE](https://github.com/caddyserver/certmagic/blob/master/LICENSE)) | v0.25.4, 2026-06-09 | 665 | Adds DNS challenges and on-demand issuance. Bus factor 1 (mholt 328 commits). | Fallback if DNS-01 is required. | Hosting slice |
| `creativeprojects/go-selfupdate` | MIT ([LICENSE](https://github.com/creativeprojects/go-selfupdate/blob/main/LICENSE)) | v1.6.0, 2026-07-08 | 84 | Release lookup, checksum and signature checks, binary swap. Not read. | Read at the hosting slice's plan. `dubplate upgrade`'s migrate-and-rollback logic stays in-repo regardless. | Hosting slice |
| `minio/selfupdate` | Apache-2.0 ([LICENSE](https://github.com/minio/selfupdate/blob/master/LICENSE)) | v0.6.0, 2022-10-19 | 119 | No release in almost four years. | Decline. | n/a |
| `ncruces/go-sqlite3` | MIT ([LICENSE](https://github.com/ncruces/go-sqlite3/blob/main/LICENSE)) | v0.35.6, 2026-09-23; 111 tags | 96 | Unchanged from the spec's row. | Keep as the named fallback. The trigram half of its trigger no longer applies. | Rung 6 condition |
| OpenSubsonic open PRs #162 (musicFolder), #237 (collections), #258 (extended library browsing), #261 (public image URL) | Apache-2.0 | Open | n/a | None has merged. #237 collides with dubplate's word "collection" (finding 8). | The study rung tracks them and re-reads before the harness rung pins a spec build. | Study rung |

## Not verified

- Which roster clients besides Feishin and Symfonium call `getTranscodeDecision`. Symfonium's use rests on a
  DroppedNeedle issue report found by search, not on Symfonium's own documentation.
- Whether any Subsonic client sends `Accept: image/webp` on `getCoverArt`. The spec's WebP-when-accepted rule depends on
  it. The art slice should capture real client requests before building WebP negotiation.
- BeetstreamNext's id scheme and projection code. Only its README and two docs pages were read.
- The code quality of `x/crypto/acme/autocert`, `certmagic`, and `go-selfupdate`. None was read beyond package metadata.
- `smolsonic`, `nina`, `tinysonic`, `ubersonic`, and `subfin-plugin`: licenses verified, code not read.
- The Subsonic XSD's license. The file carries no license text, and subsonic.org states none next to it.
- Open-issue age for most repositories. After the search API's rate limit was reached, the counts above come from the
  repository's `open_issues_count`, which includes pull requests.
- `gen2brain/vpx`'s importer count. `pkg.go.dev` did not report one at survey time.
- The WAL behavior of `modernc.org/sqlite` under concurrent readers. It is the other half of the spec's rung 6
  condition, and this survey did not test it.
