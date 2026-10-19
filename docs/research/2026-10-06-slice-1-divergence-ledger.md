# Slice 1, Task 7: the pre-seeded divergence ledger

Date: 2026-10-07. Staged outside the repo; lands at B-land as
`docs/superpowers/research/2026-10-06-slice-1-divergence-ledger.md`. Its decisions are argued in the study record,
`2026-10-06-slice-1-study-record.md` (cited as `[record]`), which also defines the citation aliases used below
(`[pins]`, `[contract]`, `[clients]`, `[base-url]`, `[servers]`, `[melodee]`, `[cells]`).

**Pins.** Every oracle behavior below was observed or read at Task 0's pins (`[pins]` "Pins"): Navidrome 0.64.2
(`sha256:38dc2727...ee52`, source `1011457`) and gonic v0.22.0 (`sha256:516fd964...1d49`, source `5ae99e8`), the
OpenSubsonic specification at `bed1688`. A version bump re-runs this ledger (spec `:278-280`).

**What this ledger owns.** One row per known divergence between an oracle and dubplate's intended behavior, merged from
Task 1's candidates C-01 to C-20 (`[contract]` "Candidate ledger rows") and Task 4's seeds D4-01 to D4-43
(`[servers]` §12), plus the rows the client and base-URL records added. Every source id appears in the evidence
column of the row it became, and the disposition map at the end accounts for all 63. Slice 2 transcribes this table
into the ledger's machine form (plan settled item 6). "Test owed" names the harness test or the slice that proves the
row (plan settled item 4).

## Columns and classes

The columns are the plan's, in its order: id, endpoint or area, divergence class, oracle, the oracle's behavior,
dubplate's intended behavior, evidence pointer, test owed. "Both" in the oracle column means the row states each
oracle's behavior. Classes:

| Class | Meaning | How the harness grades it |
|---|---|---|
| `declared: folder view` | The spec's first declared class (`:285-286`) | gonic grades the shape; Navidrome's diff is suppressed as a class |
| `declared: apiKey path` | The spec's second declared class (`:286-287`) | The specification and the roster |
| `by-design` | dubplate differs from an oracle that stays within the contract, for a spec or product reason | The spec's rule; the oracle's answer is suppressed for that row |
| `contract-deviation` | dubplate knowingly departs from the specification's text | A known-deviation entry in the contract tier that asserts the deviation still holds, so it cannot rot (`[contract]` Q3, "How Ampache does it") |
| `oracle-wrong` | The oracle contradicts the contract or a settled spec rule | The contract; the oracle's answer is a known deviation |
| `oracle-lenient` | The oracle accepts input the contract requires or rejects | The contract; dubplate is the stricter side |
| `oracles-disagree` | The two oracles differ | The oracle the row names, or the contract when it names neither |
| `feature-absent` | One oracle lacks the feature | The other oracle or the contract |
| `no-oracle` | Neither oracle exercises the behavior | A golden list, the contract, or the roster |
| `unspecified-shape` | No authority defines the form | dubplate's own rule; the oracle diff covers the fields an oracle emits |

## The ledger

| id | Endpoint or area | Divergence class | Oracle | The oracle's behavior | dubplate's intended behavior | Evidence pointer | Test owed |
|---|---|---|---|---|---|---|---|
| L-01 | Format selection, `f` and JSONP | contract-deviation | Both | Absent, unknown, or mis-cased `f` answers XML on both. `f=jsonp`: Navidrome validates the callback and answers error 0 for a bad one (two bad forms observed, four inferred from the content type); gonic reflects any callback verbatim as script | XML when `f` is absent or unrecognized, never an error. JSONP is declined: `f=jsonp` gets the unrecognized-`f` answer and no callback is ever reflected. Every API response carries `X-Content-Type-Options: nosniff` | `[contract]` E2, E3; `[clients]` §1 ("Response format asked": no roster client sends `jsonp`; Amperfy sends no `f`); `[record]` §18 item 2 (was C-03, C-04, C-05) | Harness: the five `f` forms plus `f=jsonp` with a hostile callback, asserting no callback text in the body |
| L-02 | XML form of OpenSubsonic additions | unspecified-shape | Both | Both put scalars in attributes and lists in repeated elements. gonic writes booleans (`isCompilation`) and `artistImageUrl` as elements; Navidrome omits empty additions from XML | E4's convention: scalars as attributes, lists of records as repeated elements with no wrapper, lists of strings as repeated text elements. No schema validates it; the oracle diff covers emitted fields | `[contract]` E4, Q3; `[record]` §5 (was C-06) | Harness: XML oracle diff on emitted additions; rung 16: Amperfy parses XML |
| L-03 | Empty containers, `null`, idle `nowPlaying` | oracle-wrong | Both | gonic emits JSON `null` for `replayGain`, `structuredLyrics`, and `channel`, and mixes empty containers. Navidrome's idle `nowPlaying` is `{}`, without the required `entry` | An empty container for a list with no items, never `null`; `nowPlaying` carries `entry: []` | `[contract]` E6 (was C-08) | Contract tier (the OpenAPI file declares nothing nullable) |
| L-04 | The emit-empty rule | unspecified-shape | Both | Navidrome emits empty-valued additions in JSON and omits them from XML; gonic's `releaseTypes` is `[""]` | JSON emits every supported addition with its empty value (the specification's rule); XML omits an empty scalar, and an empty list is no elements; never `[""]` | `[contract]` E4, E6 (was C-09) | Harness: one empty-valued addition per type, both formats |
| L-05 | Version negotiation | contract-deviation | Both | Neither answers error 20 or 30 for any `v`. Navidrome answers error 10 when `v` is missing; gonic does not require it | `v` is required (error 10 when missing); any value is accepted; error 20 and 30 are never sent; the envelope reports `1.16.1` | `[contract]` E9; `[clients]` §1 (roster `v` values 1.11.0 to 1.16.1) (was C-13) | Harness: missing, low, equal, and high `v` |
| L-06 | Error answer on `stream`, `download`, `getCoverArt` | contract-deviation | Both | The envelope with HTTP 200, as `application/xml` or `application/json`, never the `text/xml` the specification names | Same as both oracles: the envelope in the requested format, HTTP 200 | `[contract]` E8 (was C-12) | Harness: a nonexistent id on each binary endpoint; rung 16: Amperfy's under-2,000-byte artwork parse (`[clients]` "Routed", slice 4) |
| L-07 | Credentials echoed in response URLs | oracle-wrong | gonic | `getArtistInfo2` image URLs carry the requester's `u`, `t`, `s`, `v`, `c`, `f` | Never rebuild a URL from the request's credentials | `[contract]` E11 (was C-15) | Harness: scan every captured body for `t=`, `s=`, `p=`, `apiKey=` |
| L-08 | Error message text | oracles-disagree | Both | Each oracle words errors its own way, and neither matches the specification's text exactly | Compare the code and the envelope shape, never the message | `[contract]` Q5 (was C-16) | Harness: the comparison rule itself |
| L-09 | Wrong-credential wording | oracle-wrong | gonic | Says "invalid username" or "invalid password", which tells a caller which half was wrong; Navidrome gives one message | Error 40 with one message for both cases | `[contract]` Q5; `[servers]` §5.1, §5.2 lesson 3 | Slice 6 |
| L-10 | `apiKey` authentication and `tokenInfo` | declared: apiKey path | Both | Neither declares `apiKeyAuthentication`; `apiKey` answers error 10 on both | Declared when slice 6 lands the hash-only credential (spec `:486-488`); `tokenInfo`; errors 42, 43, 44 per the extension page | `[contract]` Q1, Q6; `[clients]` §3 (Rufin's API-key mode); `[servers]` §5.1 (was C-02, and the apiKey half of D4-21) | Slice 6, graded against the specification and Rufin |
| L-11 | Plain and hex password (`u`+`p`, `p=enc:`) | by-design | Both | Both accept both forms | Refused with error 42 (spec `:491`) | `[contract]` Q6; `[servers]` §5.1; `[record]` §18 item 3 (the password half of D4-21) | Slice 6; roster: each client logs in with the token form by default |
| L-12 | `getOpenSubsonicExtensions` without credentials | oracle-wrong | gonic | Requires `u` and `c` (error 10); Navidrome answers with no parameters at all | Answers with no parameters at all; the only data-bearing unauthenticated route (spec `:648`) | `[contract]` Q6; `[servers]` §5 (was C-19, D4-03) | Harness: an unauthenticated call, Navidrome as the oracle |
| L-13 | Conflicting credentials | oracle-wrong | Both | `u`+`p`+`t`: Navidrome answers `ok`, gonic error 10. `apiKey` with `u`: Navidrome 40, gonic 10. Neither sends error 43 | Error 43 for conflicting mechanisms, per the specification's error list; slice 6 decides whether a credential in both the query and a form body is a conflict | `[contract]` Q5, Q6; `[record]` §18 item 3 | Slice 6 |
| L-14 | An endpoint dubplate does not implement | by-design | Both | Navidrome: plain-text HTTP 404, 410, or 501 outside the envelope. gonic: HTTP 200, error 70 "view not found" | Error 0 in a well-formed envelope, HTTP 200, naming the endpoint (spec `:670-671`), except the empty-success set (`[record]` §1). Recorded quirk: Feishin treats error 0 as success on status-checked writes such as `setRating` (`[clients]` §2a) | `[contract]` E8; `[servers]` §1.1; `[record]` tensions (a), (b) (was C-11, D4-01, D4-02) | Harness: one probe per non-launch path, plus one per empty-success endpoint |
| L-15 | A missing required parameter | oracle-lenient | Both | Navidrome: `getAlbum` without `id` is error 70; `getSongsByGenre` without `genre`, `getLyrics` without a title, and `search3` without `query` answer `ok`; `createPlaylist` with no parameters is error 0 "Internal Server Error". gonic: `star` and `unstar` without `id` answer `ok`; `createPlaylist` with no parameters creates an empty-named playlist | Error 10, every endpoint, every required parameter | `[contract]` Q5; `[servers]` §1.2, §6.1, §11 (was C-18, D4-23, D4-42, D4-43) | Harness: a missing-parameter sweep graded by the contract tier |
| L-16 | An id the server did not issue | oracles-disagree | Both | Unparseable id: Navidrome error 70, gonic error 10. Well-formed id with no record: both error 70 on the reads observed; gonic's writes are the exception (`scrobble` error 0, `savePlayQueue` a silent store, L-43, L-47) | Error 70 for any id the server did not issue; opaque ids have no malformed form (spec `:379`) | `[contract]` Q5; `[record]` §18 item 5 | Harness |
| L-17 | `download` of an artist, playlist, or folder id | by-design | Both | Navidrome returns a zip of the artist; gonic answers error 0 | Error 50 naming the song-or-album rule (spec `:662-668`) | `[contract]` Q5 (was C-17) | Harness: one id of each refused kind |
| L-18 | The music folder id | oracles-disagree | Both | Navidrome: the library's 1-based database id. gonic: the 0-based index into its path list, which changes meaning when paths are reordered | A stable integer per collection, never reused (hook 22) | `[contract]` E7; `[servers]` §2; `[record]` §17 (was C-10) | Slice 3: the id survives a restart, `VACUUM`, and a delete-then-create |
| L-19 | Unknown or invisible `musicFolderId` | oracles-disagree | Both | Navidrome: error 70 on all ten endpoints. gonic: `ok` with an empty result | Error 70, with the same answer for a collection the member cannot see | `[servers]` §2; `[record]` §4 (was D4-06, D4-07) | Harness: an unknown-id row per endpoint; slice 3: a private-collection row |
| L-20 | `getScanStatus` fields | by-design | Navidrome | Emits scanner state and vendor fields (`lastScan`, `scanType`, `folderCount`, `elapsedTime`) outside both contracts | `scanning: false`, `count` of admitted tracks, and `lastScan` set to the catalog's last change, on the known-deviation list | `[contract]` Q3 ("Residuals"); `[clients]` §2b (Rufin's freshness marker); `[record]` §18 item 7 | Slice 3; roster: Rufin skips a full refresh |
| L-21 | `startScan` | by-design | Both | Starts a library scan | No scan exists, since the catalog is the index (spec `:180-186`); answers the current scan status | `[clients]` §2a, §2b, §2d; `[record]` §1 | Slice 3 |
| L-22 | `getUser` for another username | oracle-wrong | gonic | Returns the caller's own record for any username; Navidrome answers error 50, also for an unknown name | A non-admin asking for any other username gets error 50 whether or not it exists; an admin asking for an unknown one gets error 70 | `[contract]` Q5 | Slice 6 |
| L-23 | Id shape | by-design | gonic | Typed prefixes with a database number (`ar-1`, `al-3`, `tr-1`) | Opaque server-issued tokens (spec `:379`), as Navidrome's are | `[servers]` §8, §11 (was D4-38) | Harness: an id-shape assertion |
| L-24 | `search3`, queries of one or two characters | no-oracle | Both | Navidrome: nothing for one character, word-prefix matches for two. gonic: substring matches, a flood for one character | The prefix lookup for queries under three characters (spec `:403-405`) | `[servers]` §6.2 (was D4-24) | Harness: a golden list, never an oracle diff |
| L-25 | `search3`, `query=*` | oracles-disagree | Both | Navidrome strips a trailing `*` and answers everything; gonic matches nothing | Navidrome's: a trailing `*` is stripped, so a bare `*` is the empty query | `[servers]` §6.1 (was D4-25) | Harness |
| L-26 | `search3`, the two-quote empty query | oracle-wrong | gonic | Returns three extra album rows with empty titles (cause unconfirmed) | Real rows only, in a deterministic natural order | `[servers]` §6.1 (gonic half of D4-22) | Harness: a whole-library page walk |
| L-27 | `getAlbumList2` `size` edges | oracle-wrong | Both | Navidrome returns every album for `size=0` or `-1`; gonic returns none for 0 and leaks a SQL syntax error for `-1` | Above 500 clamps to 500; 0 returns an empty list; negative or non-numeric answers error 0 naming the parameter; never an unbounded list, never a database error | `[servers]` §6.3 (was D4-26) | Harness: an edge-size table |
| L-28 | A second request on a filling transcode | oracle-wrong | gonic | Blocks with no headers until the first encode ends (65.8 s observed), then serves the complete file | Shares the encode from byte 0, as Navidrome does (spec `:433-434`) | `[servers]` §3.2 (was D4-09) | Slice 5: a second-listener test |
| L-29 | `estimateContentLength` | contract-deviation | Both | Navidrome's estimate ignores `timeOffset`; both run about 2.4% high | Not honored at 0.8: no roster client sends it, and an overshooting length reads as truncation | `[servers]` §3.3; `[clients]` §5; `[record]` §7 (was D4-10, D4-11) | Slice 5 |
| L-30 | `timeOffset` caching | oracles-disagree | Both | Navidrome caches each offset as its own entry and serves ranges once it completes; gonic re-encodes every request and ignores Range | Navidrome's: the offset is part of the cache key | `[servers]` §3.3 (was D4-12) | Slice 5 |
| L-31 | A failed encode | oracles-disagree | Both | gonic ends the stream cleanly after ffmpeg is killed, indistinguishable from a short track. Navidrome aborts the connection and evicts the entry | Evict and fail. Before the first byte the request returns 500 (spec `:447-449`); after it, the connection is aborted, never ended cleanly | `[servers]` §3.4; `[record]` §7 (was D4-13, D4-14) | Slice 5: the failed-exit race test |
| L-32 | A client disconnect | oracles-disagree | Both | Navidrome keeps encoding to completion by default; gonic stops ffmpeg within a second | Kill ffmpeg only when no reader is attached (spec `:449-450`); slice 5 weighs a short grace for Amperfy's probe-then-cancel pattern | `[servers]` §3.4; `[clients]` §5; `[record]` §7 (was D4-15) | Slice 5: the disconnect race test, plus an Amperfy-shaped probe test |
| L-33 | The concurrent-encode cap | oracles-disagree | Navidrome | No queue: HTTP 429, `Retry-After: 5`, and an error-0 envelope; gonic has no cap | A bounded queue in arrival order, then error 0 with HTTP 200 (spec `:450-452`) | `[servers]` §3.5 (was D4-16) | Slice 5 |
| L-34 | The `transcoding` extension | feature-absent | gonic | Not declared; both endpoints answer error 70. Navidrome declares it | Pending fork A; recommended not declared at 0.8 | `[servers]` §4; `[record]` §6 (was D4-17) | Harness: an extension-gated test set |
| L-35 | AAC in a transcode decision | oracle-lenient | Navidrome | Deletes every AAC transcoding profile a client posts | Only if fork A is yes: AAC is offered as ADTS (`[record]` §9) | `[servers]` §4.3 (was D4-18) | Slice 5, if fork A is yes |
| L-36 | The device profile against a request | oracles-disagree | Navidrome | A pinned player's bitrate ceiling outranks the request and a posted `ClientInfo`; a forced format applies only if the client can play it; the legacy parameters go through the same decision | The device profile is a bitrate ceiling and a default format; a request may lower it, never raise it; one decision path for the legacy parameters and the extension (amends spec `:411-412`) | `[servers]` §4.3, §4.4; `[clients]` §5; `[record]` §18 item 1 (was D4-19) | Slice 5 |
| L-37 | Cookies on API responses | by-design | Navidrome | Sets an `nd-player-<hex>` cookie on every `/rest` response (`Path` follows the base path) | Sets no cookie on any Subsonic response | `[contract]` E12; `[base-url]` "Navidrome's `Set-Cookie` under a base path" | Harness: no `Set-Cookie` under the Subsonic prefix |
| L-38 | CORS | oracles-disagree | Both | Both allow any origin and never send credentials. gonic stamps fixed headers on every response; Navidrome answers only when `Origin` is present and reflects requested headers | `Access-Control-Allow-Origin: *` on every Subsonic response; never `Access-Control-Allow-Credentials`; a bodyless preflight allowing `GET`, `POST`, and `Content-Type`; no CORS headers outside the Subsonic prefix | `[servers]` §5; `[clients]` §6; `[base-url]` header table; `[record]` §8 (was D4-20) | Slice 3: preflight rows; roster: Feishin's web build |
| L-39 | `getCoverArt` with `Accept: image/webp` | no-oracle | Both | Neither negotiates; both return JPEG for every `Accept` | WebP when `Accept` names it (spec `:388`), which Feishin's `<img>` loads do; a format-varying answer sends `Vary: Accept` | `[servers]` §11; `[base-url]` header table; `[record]` tension (h) (was D4-39) | Slice 4: art rows |
| L-40 | `getCoverArt` with no `id` | oracle-lenient | Navidrome | HTTP 200 and a 69,228-byte WebP placeholder | Error 10. A real id with no cover still gets the generated placeholder (spec `:390`) | `[servers]` §11 (was D4-40) | Slice 4 |
| L-41 | `scrobble` with `submission=false` | oracle-wrong | gonic | Counts a play; Navidrome records now-playing only | Now-playing only (spec `:564-565`) | `[servers]` §7.2 (was D4-31) | Slice 7: event-kind rows |
| L-42 | `scrobble` with several ids | oracle-wrong | gonic | Reads only the first id | Every id is recorded. A `time` count other than zero or the id count answers error 0 and records nothing, as Navidrome does | `[servers]` §7.2 (was D4-32) | Slice 7 |
| L-43 | `scrobble` with an unknown id | oracles-disagree | Both | `bogus` (unknown to Navidrome, unparseable for gonic's typed ids): Navidrome answers `ok` and records nothing for it; gonic answers error 10. Well-formed unknown id (`tr-999`): gonic answers error 0 "error finding track" ([S], not run) | Error 70, and nothing in the batch is recorded (one transaction, hook 9; OpenSubsonic PR #202's direction) | `[servers]` §7.2; `[contract]` Q4 (#202) (was D4-33) | Slice 7; harness: gonic's well-formed-id case |
| L-44 | `scrobble` with an unparseable `submission` | oracles-disagree | Both | Navidrome reads it as false; gonic counts it | Error 0 naming the parameter, nothing recorded, because play history is append-only | `[servers]` §7.2; `[record]` §18 item 6 (was D4-34) | Slice 7 |
| L-45 | `savePlayQueue` with no `id` | oracles-disagree | gonic | Error 10, and the old queue stays; Navidrome clears it | Clears the saved queue (spec `:567-568`) | `[servers]` §7.1 (was D4-27) | Slice 7: queue rows |
| L-46 | `savePlayQueue` with `current` absent or outside the list | oracle-lenient | Both | Navidrome defaults `current` to the first entry; gonic stores `-1` or the stray id | `current` absent while `id` is non-empty: error 10 (spec `:568`). `current` outside the list: error 0 naming it. The queue is unchanged either way | `[servers]` §7.1 (was D4-28) | Slice 7 |
| L-47 | `savePlayQueue` with an unknown `id` | oracles-disagree | Both | `bogus` (unknown to Navidrome, unparseable for gonic's typed ids): Navidrome drops it silently; gonic answers error 10. Well-formed unknown id (`tr-999`): gonic stores it silently ([S], not run) | Error 70 naming the id; the queue is unchanged, so the client learns its view is stale | `[servers]` §7.1 (was D4-29) | Slice 7; harness: gonic's well-formed-id case |
| L-48 | `getPlayQueue` with nothing saved | oracles-disagree | gonic | No `playQueue` element at all; Navidrome answers the element with the zero time | Navidrome's shape: Feishin dereferences `playQueue` without a guard, and the specification's `PlayQueue` requires its fields | `[servers]` §7.1; `[clients]` §2a (was D4-30) | Slice 7 |
| L-49 | `savePlayQueue` with a negative or non-numeric `position` | oracle-lenient | Both | Both store `-5`; both treat `abc` as absent | An absent `position` means zero (spec `:568-569`); a negative or non-numeric one answers error 0 naming it, and the queue is unchanged | `[servers]` §7.1 | Slice 7 |
| L-50 | Smart playlists and `.nsp` | feature-absent | gonic | No smart playlists. Navidrome imports `.nsp` files and serves them read-only | dubplate's own rule form, read-only to clients (spec `:597-602`); `.nsp` is read only by the switcher import | `[servers]` §9; `[record]` §13 (was D4-41) | Slice 7; slice 8 for the import |
| L-51 | The folder view against Navidrome | declared: folder view | Navidrome | A synthesized Artist/Album/Track tree and synthetic `path` values from tags | Pending fork C. If yes, a tree derived from the pipeline's paths (spec `:375-379`); if no, an empty `getIndexes` and error 70 from `getMusicDirectory` | `[servers]` §8; `[record]` §12 (was D4-35) | Harness: folder rows graded by gonic, only if fork C is yes |
| L-52 | The folder view against gonic | oracles-disagree | gonic | The real tree to any depth, an album named after its directory, typed ids, and a root-level file not scanned (cause unconfirmed) | If fork C is yes, gonic grades the tree's shape only, never its ids, album naming, or scanner choices | `[servers]` §8 (was D4-36, D4-37) | Harness, only if fork C is yes |
| L-53 | Shares | by-design | Both | Navidrome implements the four share endpoints with no bounds; gonic has none | The six bounds (spec `:512-548`), graded against the spec only | `[servers]` §1.1, §10.1 | Slice 6 |
| L-54 | The declared extension list | by-design | Both | Navidrome declares 7 of 10, gonic 3 of 10 (gonic only with credentials) | The launch subset: `formPost`, `transcodeOffset`, and `apiKeyAuthentication` from slice 6, plus `transcoding` only if fork A is yes; the list equals what is implemented at each slice's close | `[contract]` Q1; `[servers]` §4; `[record]` §2 (was C-01, D4-04) | Harness: the declared list equals the subset, names and versions |

**Rows: 54.**

## Agreement rows (no divergence; test still owed)

These source rows describe behavior where dubplate matches the oracles. They leave the divergence table, and their
tests stay owed.

| id | Area | Behavior both oracles and dubplate share | Evidence pointer | Test owed |
|---|---|---|---|---|
| A-01 | Single-element lists | Always a JSON array, never collapsed to an object | `[contract]` E5 (was C-07) | Harness: a one-element list per path |
| A-02 | Escaping | Literal UTF-8; the Go encoders' HTML-safe JSON escapes; no XML declaration; compare parsed values | `[contract]` E10 (was C-14) | Harness |
| A-03 | Preflight | No body, no data, no credentials needed | `[contract]` Q6; `[servers]` §5 (was C-20) | Slice 3: CORS rows |
| A-04 | `musicFolderId` coverage | Honored on the ten endpoints, ignored elsewhere | `[servers]` §2; `[record]` §4 (was D4-05) | Harness: a per-endpoint filter table on a two-collection fixture |
| A-05 | `search3` empty query | Returns everything, paged without gaps; dubplate adds a deterministic order | `[servers]` §6.1, §6.3 (Navidrome half of D4-22) | Harness: a whole-library page walk |
| A-06 | Range on a transcode | Navidrome's shape adopted: a filling entry answers 200 from byte 0 with `Accept-Ranges: none`; a complete entry serves ranges; spec `:434-435` is retired | `[servers]` §3.2; `[record]` §7 (was D4-08) | Slice 5; harness: the filling-entry row |
| A-07 | The `.view` suffix | Every method answers with and without it | `[pins]` "The three instances"; `[base-url]` Tempus section; `[record]` tension (e) | Slice 3 |
| A-08 | Server-emitted links under a mount prefix | Share URLs carry the prefix, as Navidrome's do | `[base-url]` "What the servers document and do" | Slice 6 |
| A-09 | Direct-path ranges | 200 with `Accept-Ranges: bytes`; 206 for a range; 416 with `bytes */<size>` when unsatisfiable | `[servers]` §3.1 | Harness |
| A-10 | A stream is never a play | No play count from `stream` | `[servers]` §7.2 | Slice 7 |
| A-11 | Form-body credentials | Accepted in an `application/x-www-form-urlencoded` POST | `[contract]` Q6; `[servers]` §5.1 | Harness |
| A-12 | The integration-dependent family | An empty success with no integration configured | `[servers]` §1.2; `[record]` §1 | Harness: the empty-success rows |

## Disposition map

Every seed id, and the row it became.

| Source ids | Row |
|---|---|
| C-01, D4-04 | L-54 |
| C-02 | L-10 |
| C-03, C-04, C-05 | L-01 |
| C-06 | L-02 |
| C-07 | A-01 |
| C-08 | L-03 |
| C-09 | L-04 |
| C-10 | L-18 |
| C-11, D4-01, D4-02 | L-14 |
| C-12 | L-06 |
| C-13 | L-05 |
| C-14 | A-02 |
| C-15 | L-07 |
| C-16 | L-08 |
| C-17 | L-17 |
| C-18, D4-23, D4-42, D4-43 | L-15 |
| C-19, D4-03 | L-12 |
| C-20 | A-03 |
| D4-05 | A-04 |
| D4-06, D4-07 | L-19 |
| D4-08 | A-06 |
| D4-09 | L-28 |
| D4-10, D4-11 | L-29 |
| D4-12 | L-30 |
| D4-13, D4-14 | L-31 |
| D4-15 | L-32 |
| D4-16 | L-33 |
| D4-17 | L-34 |
| D4-18 | L-35 |
| D4-19 | L-36 |
| D4-20 | L-38 |
| D4-21 | L-10 (the apiKey half), L-11 (the password half) |
| D4-22 | A-05 (Navidrome), L-26 (gonic's extra rows) |
| D4-24 | L-24 |
| D4-25 | L-25 |
| D4-26 | L-27 |
| D4-27 | L-45 |
| D4-28 | L-46 |
| D4-29 | L-47 |
| D4-30 | L-48 |
| D4-31 | L-41 |
| D4-32 | L-42 |
| D4-33 | L-43 |
| D4-34 | L-44 |
| D4-35 | L-51 |
| D4-36, D4-37 | L-52 |
| D4-38 | L-23 |
| D4-39 | L-39 |
| D4-40 | L-40 |
| D4-41 | L-50 |
| New, from the records | L-09, L-11 (in part), L-13, L-16, L-20, L-21, L-22, L-37, L-49, L-53; A-07 to A-12 |

All 20 C ids and all 43 D4 ids are accounted for.

## Routed

Every row's test owed names its slice. Slice 2 transcribes the table and both class lists; the rows pending forks A
and C (L-34, L-35, L-51, L-52, and L-54's `transcoding` clause) resolve when Geoff rules (`[record]` §19).
