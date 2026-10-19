# Slice 1, Task 7: the study record

Date: 2026-10-07. Staged outside the repo; lands at B-land as
`docs/research/2026-10-06-slice-1-study-record.md`, beside its ledger,
`2026-10-06-slice-1-divergence-ledger.md`. This is the file the spec, slice 2, and each slice plan cite. It holds the
index and the decisions; the evidence lives in the seven segment A records it points into.

**Pins read.** Every pin is Task 0's (`[pins]` "Pins"): Navidrome 0.64.2 (`sha256:38dc2727...ee52`, source tag
`v0.64.2` = `1011457`), gonic v0.22.0 (`sha256:516fd964...1d49`, `5ae99e8`), the OpenSubsonic specification at
`bed1688`, `navidrome/website` at `b0fe572`, Feishin v1.17.0 (`5c169ac`), Rufin v0.16.6 (`aad2f44`), Amperfy v2.1.1
(`c165a7a`), Tempus v4.28.0 (`687d97f`), Melodee v2.2.0 (`d91342e`). Task 1 added the Subsonic 1.16.1 XSD (sha256
`9baee72f...0387`, `[contract]` header). This task ran no oracle; its fact-read fold re-read two gonic handlers and one Navidrome handler at the pins. Three primary sources were
read for quotations: RFC 9110 §14.2, SQLite's `VACUUM` page, and `go doc` for `net/http` and `time` (Go 1.26 on this
workstation).

**Spec lines** are revision 3 of `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` ("the spec") at `0cb33ae`
and later master, unchanged by this lane. The re-plan is `docs/specs/2026-10-05-rungs-12-16-replan-design.md`.

**Citation aliases.** Each names a segment A record under `docs/research/`:

| Alias | File | Owner |
|---|---|---|
| `[pins]` | `2026-10-06-slice-1-evidence-pins.md` | Task 0 |
| `[contract]` | `2026-10-06-slice-1-contract.md` | Task 1 |
| `[clients]` | `2026-10-06-slice-1-client-demands.md` | Task 2 |
| `[base-url]` | `2026-10-06-slice-1-base-url.md` | Task 3 |
| `[servers]` | `2026-10-06-slice-1-server-behavior.md` | Task 4 |
| `[melodee]` | `2026-10-06-slice-1-melodee-intake.md` | Task 5 |
| `[cells]` | `2026-10-06-slice-1-prior-art-cells.md` | Task 6 |
| `[ledger]` | `2026-10-06-slice-1-divergence-ledger.md` | this task |

Section pointers follow each record's own headings (`[contract]` E8, `[servers]` §3.2, `[clients]` §2e).

## Questions this record owns, and their outcomes

Three outcome states, per the plan's planning-miss check 2: **settled** (answered, with the reason), **owner fork**
(answered pending Geoff's ruling, with a recommendation), and **routed** (handed to a named slice, with what it
receives). An item may carry a settled core and a routed remainder; the table names the deciding state.

| # | Item (spec line) | Outcome | Section |
|---|---|---|---|
| 1 | The empty-success endpoint set (`:670-672`) | Settled | 1 |
| 2 | The launch extension subset (`:672-673`) | Settled, with `transcoding` held for fork A | 2 |
| 3 | The per-endpoint oracle assignment (`:657-661`) | Settled | 3 |
| 4 | The `musicFolderId` endpoint list (`:375`) | Settled | 4 |
| 5 | The XML validation rule (`:271`, `:939-940`) | Settled | 5 |
| 6 | The `transcoding` extension (`:418-419`, `:941-942`) | **Owner fork A** | 6 |
| 7 | Range on a filling transcode (`:443`, `:943-944`) | Settled | 7 |
| 8 | The CORS policy (`:654-655`, `:945`) | Settled | 8 |
| 9 | The AAC container (`:431`, `:946`) | Settled, with a playback test owed | 9 |
| 10 | The protocol module's license and NOTICE (`:327-328`, `:947-948`) | **Owner fork B** | 10 |
| 11 | The word "collection" and the PR watch (`:951-955`) | Settled; the re-read routed to slice 2 | 11 |
| 12 | Folder browsing at launch (`:925-926`) | **Owner fork C** | 12 |
| 13 | The `.nsp` format (`:934`) | Settled (format recorded); the import routed to slice 8 | 13 |
| 14 | The API base URL (re-plan `:427`) | Routed to slice 3 (measurement summary) | 14 |
| 15 | Melodee's intake pipeline (ladder entry 1, `:725-728`) | Settled (no verdict flips) | 15 |
| 16 | The prior-art table's open cells (ladder entry 1, `:729-730`; table `:856-891`) | Routed to Task 9, with five rulings on Task 6's flags | 16 |
| 17 | New forward hooks | Settled: one new hook (22), consumed by slice 3 | 17 |
| 18 | Readings the ledger's "to settle" rows and the client evidence force | Settled, each in its ledger row | 18 |

The owner bundle (section 19) gathers forks A to C and one hardware request. Everything else here is a method or
architecture reading, settled from the evidence under `CLAUDE.md` "Process proportionality".

---

## Cross-record tensions, resolved

The segment A records disagree in nine places. Each is resolved once here, and the items below rest on the
resolutions.

**(a) Amperfy's 404, against "never a bare 404".** Amperfy converts an HTTP 404 into error 70 before it reads the
body (`[clients]` §1 table, last row; `amperfy@c165a7a:AmperfyKit/Api/Subsonic/SubsonicServerApi.swift:976-987`), and
its podcast probe reads a thrown request as "no podcasts" (`[clients]` §2c, `getPodcasts` row). So for one endpoint
and one client, a bare 404 hides a library section that an empty success leaves visible and empty. **Resolved: the
spec's rule stands** (`:670-671`), and `getPodcasts` joins the empty-success set (section 1). The reason: a 404 under
the Subsonic prefix would mean two things, "wrong address" and "feature absent". Keeping every API answer inside the
envelope leaves a 404 meaning only "wrong address", which is the signal the base-URL work needs (section 14). A
client misconfigured without Navidrome's prefix gets a redirect to its web UI rather than a clean failure
(`[base-url]` "Cross-client facts", last bullet), so the field already lacks a clear wrong-address answer, and
dubplate should not blur its own. Amperfy's cost is an
empty Podcasts entry in its library, quiet on every sync (`[clients]` §2e, `getPodcasts` row).

**(b) Neither oracle answers an unimplemented endpoint with error 0.** Navidrome answers outside the envelope with a
plain-text 404, 410, or 501; gonic answers error 70 "view not found" in the envelope (`[servers]` §1.1;
`[contract]` E8). **Resolved: error 0 inside a well-formed envelope, HTTP 200, with a message naming the endpoint**,
as the spec says (`:670-671`). Error 70 is untrue for an endpoint (it means "the requested data was not found",
`[contract]` Q5) and sends a client to the wrong repair (`[servers]` §1.3, lesson 1). Plain text outside the envelope
breaks the clients that parse the body first: Amperfy parses every answer as XML (`[clients]` §1), and Rufin reads a JSON
`subsonic-response` on every call (`[clients]` §2b). The unimplemented answer is graded against the spec, never an
oracle diff (`[ledger]` L-14).

**(c) The OpenSubsonic repository has no NOTICE file.** The spec's prior-art cell says "never vendor it, because the
NOTICE clause would attach" (`:866`). Task 6 found no `NOTICE` file anywhere in the repository and a `LICENSE` with
the unfilled appendix (`[cells]` flag 1). Apache-2.0 §4(d) binds only a work that includes a NOTICE file, so no
NOTICE text attaches today; §4(a) to §4(c) still bind any redistribution (`[cells]` flag 1). **Resolved: the cell's
reason is corrected** (Task 9 rewrites it as "never vendor it: §4(a) to §4(c) bind redistribution, and a NOTICE file
added upstream would bind §4(d)"), and the fact feeds fork B.

**(d) PR #237's "collection" is a per-user list with a `public` flag.** Dubplate's collection is a library
partition mapped to a music folder (`[contract]` Q4, "What #237's definition means"). **Resolved in section 11:**
dubplate never declares the `collections` extension, and its API-facing prose says "music folder" for the wire concept.

**(e) Tempus never sends `.view`; Feishin, Rufin, and Amperfy always do.** Tempus's routes are bare
(`[clients]` §1 table, first row; `[base-url]` "No `.view` suffix"); the other three add the suffix on every method
(Feishin on all but two). Both oracles answer both forms (`[pins]` "The three instances"; `[base-url]` Tempus
section). **Resolved: every method answers at `<base>/rest/<name>` and `<base>/rest/<name>.view`**, whatever base
slice 3 picks. This is a requirement handed to slice 3, not a fork (`[ledger]` A-07).

**(f) No hardware virtualization, so Tempus and Amperfy are source-only.** The workstation shows no `vmx` flag and no
`/dev/kvm`, so no Android emulator runs (`[pins]` "Not done"); Amperfy cannot run on Linux at all (`[base-url]`
Verdicts). **Resolved: every Tempus and Amperfy claim in this study is a source reading, marked as one, with its
live proof owed.** Tempus's proofs move to slice 2's roster walk once Geoff enables VT-x in the firmware, which is the
one hardware request in section 19. Amperfy's stay at rung 16's attended gate (re-plan `:140-141`).

**(g) Task 1's XML reading said "revisit if a roster client parses XML", and Task 2 found one.** Amperfy sends no `f`
and parses every response as XML (`[clients]` §1, "Response format asked"; `[contract]` Q3, "Recommended reading").
**Resolved in section 5:** the legacy-form rule stands, and Amperfy's XML demands become roster-tier tests, because
they fall on core Subsonic attributes and value formats rather than on the OpenSubsonic additions.

**(h) Task 2 predicted Feishin's covers send `*/*`; Task 3 observed both forms.** Task 2 read Feishin's `fetch` path
and observed headless Chromium (`[clients]` §7). Task 3's logging proxy saw Feishin's rendered `<img>` requests send
`image/avif,image/webp,...` and its `fetch` requests send `*/*`, three each (`[base-url]` "Headers, preflights,
ranges, cookies"). **Resolved: Task 3's observation governs.** Feishin's `<img>` loads trigger the spec's
WebP-when-accepted rule (`:388`), so slice 4 receives a live trigger and a `Vary: Accept` requirement (`[ledger]` L-39).

**(i) Two Melodee quality texts.** Task 5 proposed a cell centered on the intake path (`[melodee]` "The prior-art
table's Melodee quality cell"); Task 6 graded the whole tree, including an auth carve-out and vacuous test assertions
(`[cells]` "Melodee's quality cell"). **Resolved in section 16** with one merged cell text.

---

## 1. The empty-success endpoint set

**Answer.** At 0.8 these endpoints answer `status: ok` with an empty payload of their documented shape:

| Endpoint | Why it is in the set | Evidence |
|---|---|---|
| `getInternetRadioStations` | Amperfy lists Radios in its default library and pops up an error on every visit when answered with error 0; dubplate holds no stations, so an empty list is true | `[clients]` §2c, §2e |
| `getPodcasts`, `getNewestPodcasts` | Amperfy probes podcast support through `getPodcasts`; error 0 arrives as HTTP 200, reads as "supported", then raises popups on later syncs. An empty success is supported-but-empty and quiet. Tempus has a podcast screen that is soft either way | `[clients]` §2c, §2d, §2e; tension (a) |
| `getArtistInfo`, `getArtistInfo2` | Integration-dependent: both oracles answer an empty success with no external service configured, and dubplate has none at 0.8 (Last.fm and artist images are on the 2.1 list, `:895-897`) | `[servers]` §1.2 |
| `getTopSongs`, `getSimilarSongs`, `getSimilarSongs2` | Integration-dependent, as above. Rufin and Amperfy show an error on a user's radio or top-songs action under error 0; an empty list is true | `[servers]` §1.2; `[clients]` §2e |
| `getIndexes` | **Only if fork C rules no folder browsing.** Amperfy's default Directories entry calls it after a non-empty `getMusicFolders` (collections are music folders), and error 0 would pop up per visit | `[clients]` §2c, §4; section 12 |

Two launch-list endpoints behave the same way by design and sit outside the set because the spec already names them:
`getLyrics` is the launch's lyrics stub (`:658-659`), answering an empty `lyrics` element, and `getLicense` answers its
constant `valid: true`, as both oracles do (`[servers]` §1.2).

**The rule behind the set, settled.** An empty success is allowed only on a read, and only where "nothing here" is
true of dubplate's data. A write never answers an empty success, because that would claim a change the server did
not make. Every other endpoint outside the launch surface answers error 0 (tension (b)).

**What was weighed and left out.**

- `getAlbumInfo2` (and `getAlbumInfo`, which answers the same body) is implemented, not stubbed: the spec fills
  `musicBrainzId` there (`:660-661`). Its notes and image fields stay empty, as both oracles leave them with no
  integration (`[servers]` §1.2).
- `jukeboxControl`: Feishin probes it on connect and carries on after any failure (`[clients]` §2a; observed against
  Navidrome's 501 in `[base-url]` Feishin walk, step 1). Error 0 suffices.
- `getPlayQueue` and `getScanStatus` are on the launch list and answer real data. Their empty states are ledger rows:
  `getPlayQueue` with nothing saved answers the `playQueue` element (Feishin dereferences it, `[clients]` §2a,
  `[ledger]` L-48), and `getScanStatus` carries a `lastScan` marker for Rufin (`[ledger]` L-20).
- `startScan` is not unimplemented: the catalog is the index and no scanner exists (`:180-186`), so the request is
  already satisfied. It answers the current scan status (`[ledger]` L-21).
- `setRating` stays error 0, since ratings are 2.1 (`:208`). One client quirk is recorded rather than fixed: Feishin's
  response interceptor skips its error branch for code 0 (`[clients]` §2a), so a Feishin rating looks saved until a
  reload. The alternatives were worse: an empty success would claim the write, and a nonzero code such as 50 would be
  untrue. `[ledger]` notes it under L-14.
- `getNowPlaying`, `getBookmarks`, `getLicense`'s callers, `getShares` for Tempus: no roster client needs them to
  answer for normal use (`[clients]` §2e, last two rows). `getNowPlaying` is implemented anyway under album lists
  (section 3), because `scrobble` with `submission=false` records the state only it reads (`:564-565`).

**Test owed.** Slice 2's harness: one row per member of the set asserting an `ok` envelope that validates against
the OpenAPI schema, and one probe per other non-launch path asserting error 0 (`[ledger]` L-14). The popup behavior is
proven at rung 16 for Amperfy (tension (f)).

## 2. The launch extension subset

**Answer.** Measured against the ten extensions at `bed1688` (`[pins]` "Pins"; `[contract]` Q1):

| Extension | 0.8 | Reason | Evidence |
|---|---|---|---|
| `formPost` | **Declare** | Both oracles declare it; Feishin moves every API call into a form body when it is declared, which takes credentials out of URLs and access logs; PR #278 tells clients to prefer it. The cost is small: `go doc net/http.Request.ParseForm` says that for POST it "reads the request body, parses it as a form and puts the results into both r.PostForm and r.Form" | `[contract]` Q1, Q6; `[clients]` §3; `[base-url]` Feishin walk step 5 |
| `transcodeOffset` | **Declare** | Both oracles declare it. Tempus uses `timeOffset` to seek in a transcoded stream when it is declared, and section 7's Range reading makes `timeOffset` the seek path during a transcode | `[contract]` Q1; `[clients]` §3; `[servers]` §3.3 |
| `apiKeyAuthentication` | **Declare when slice 6 lands it** | The spec already plans a hash-only "modern-only" credential for clients that support it (`:486-488`); Rufin offers API-key login as a member-chosen mode and verifies with `tokenInfo`. No oracle covers it, which is the spec's declared class (`:286-287`) | `[contract]` Q1; `[clients]` §3; `[ledger]` L-10 |
| `transcoding` | **Fork A** | Scope; section 6 | `[servers]` §4 |
| `songLyrics` | Not at 0.8 | Lyrics beyond the stub are 2.1 (`:895-897`). Declaring it would advertise a capability served only as empty lists. Without it, Rufin and Tempus fall back to `getLyrics`, the launch stub, and Feishin and Amperfy skip lyrics | `[clients]` §3 |
| `indexBasedQueue` | Not at 0.8 | Only Feishin checks it, and Feishin falls back to the id-based queue the spec specifies (`:567-569`) | `[clients]` §3 |
| `playbackReport` | Not at 0.8 | Only Feishin checks it. It adds a second now-playing channel beside `scrobble`, whose rule the spec settles (`:564-565`) | `[clients]` §3 |
| `topSongsByArtistId` | Not at 0.8 | Only Feishin checks it, and `getTopSongs` answers an empty success at 0.8 (section 1) | `[clients]` §3 |
| `sonicSimilarity`, `getPodcastEpisode` | Not at 0.8 | No roster client checks either, and no oracle serves either | `[contract]` Q1; `[clients]` §3 |

The extensions in open PRs are not declared: `collections` (section 11), `publicImageUrl` (section 11), `musicFolder`,
`extendedLibrary`, and the rest (`[contract]` Q1, "Extensions the open PRs would add"). The declared list must equal
what is implemented at each slice's close, so it grows as slices land (`[ledger]` L-54).

## 3. The per-endpoint oracle assignment

The spec's tiers are contract, oracle, and roster (`:258-289`). Every endpoint takes the contract tier; the table
names the oracle that grades it beyond that, or "contract only" when neither oracle is a fair grader. "Both" means
Navidrome primary and gonic secondary, per the spec's default (`:276-280`).

| Family (spec `:657-661`) | Endpoint | Oracle | Why, and the roster demand it carries |
|---|---|---|---|
| System | `ping` | Both | Envelope only. Roster: Amperfy needs `version` on the root; Feishin needs `openSubsonic` and `serverVersion`, or it never asks for extensions (`[clients]` §1) |
| | `getLicense` | Both | Constant (`[servers]` §1.2) |
| | `getOpenSubsonicExtensions` | Navidrome for the carve-out; contract only for the list | gonic requires credentials, against the spec (`[servers]` §5; `[ledger]` L-12); the list is dubplate's subset (section 2) |
| | `getScanStatus`, `startScan` | Contract only | dubplate has no scanner, and both oracles' answers are scanner state. Roster: Rufin's freshness marker (`[ledger]` L-20, L-21) |
| Browsing, ID3 | `getArtists`, `getArtist`, `getAlbum`, `getSong`, `getGenres`, `getAlbumInfo2`, `getAlbumInfo` | Both | Roster: Amperfy pages its album sync from `albumCount` on each artist; Feishin throws on a song without `contentType` (`[clients]` §1) |
| | `getMusicFolders` | gonic (spec `:278`), shape only | The id value differs by design (`[ledger]` L-18) |
| Browsing, folder | `getIndexes`, `getMusicDirectory` | gonic if fork C is yes; contract only if no | `[servers]` §8; section 12 |
| Info family | `getArtistInfo`, `getArtistInfo2`, `getTopSongs`, `getSimilarSongs`, `getSimilarSongs2` | Both | Both answer the empty success with no integration (`[servers]` §1.2) |
| Album lists | `getAlbumList2`, `getAlbumList`, `getRandomSongs`, `getSongsByGenre`, `getStarred2`, `getStarred` | Both | Feishin calls `getStarred`, not `getStarred2`, for its favorites (`[clients]` §2a), so both must carry data. Size edges are a ledger row (`[ledger]` L-27) |
| | `getNowPlaying` | Navidrome | gonic does not implement it (`[servers]` §1.1); Navidrome's idle answer lacks the required `entry` (`[ledger]` L-03) |
| Search | `search3`, `search2` | Both, as a set plus top-N (`:287-289`); a golden list for queries under three characters | `[servers]` §6.2; `[ledger]` L-24 |
| Media | `stream` (direct) | Both | Agree on ranges (`[servers]` §3.1) |
| | `stream` (transcode) | Navidrome | gonic blocks on a filling entry and ends cleanly after a failed encode (`[ledger]` L-28, L-31) |
| | `download` | Contract only | Navidrome zips an artist, gonic answers error 0; the one-album rule and zip layout are dubplate's (`:662-668`; `[ledger]` L-17) |
| | `getCoverArt` | Both for the success shape; contract and slice 4's own tests for WebP and the placeholder | Neither negotiates WebP (`[servers]` §11; `[ledger]` L-39) |
| | `getLyrics` | Contract only | The stub |
| Annotation | `star`, `unstar` | Both, gonic's missing-id leniency excepted | `[ledger]` L-15 |
| | `scrobble` | Navidrome | gonic counts a non-submission and reads one id (`[servers]` §7.2; `[ledger]` L-41, L-42) |
| Playlists | `getPlaylists`, `getPlaylist`, `createPlaylist`, `updatePlaylist`, `deletePlaylist` | Both; smart playlists Navidrome only | gonic has no smart playlists (`[ledger]` L-50) |
| Play queue | `getPlayQueue`, `savePlayQueue` | Navidrome for the empty answer; contract and the spec's edges for the rest | Both oracles are lenient on the spec's edges (`[servers]` §7.1; `[ledger]` L-45 to L-49) |
| Shares | `getShares`, `createShare`, `updateShare`, `deleteShare` | Navidrome for shape; the spec only for the six bounds | gonic has no shares (`[servers]` §1.1); no oracle bounds a share (`[ledger]` L-53) |
| User | `getUser` | Navidrome | gonic returns the caller's record for any username (`[ledger]` L-22) |
| Extension endpoints | `tokenInfo` | Contract and the roster (Rufin) | The declared apiKey class (`:286-287`) |
| | `getTranscodeDecision`, `getTranscodeStream` | Navidrome, if fork A is yes | gonic lacks the extension |
| Cross-cutting | The unimplemented answer and the empty-success set | Contract only | Tension (b) |
| | CORS | Contract and the roster (Feishin) | Section 8 |
| | Encoding (format selection, nulls, singletons, escaping, XML shape) | Contract (JSON schema and the legacy XSD) plus both oracles for shape | `[contract]` Q2; section 5 |

**Test owed.** Slice 2 transcribes this table into the harness's per-endpoint oracle map, and its area list adds the
three cross-cutting classes Navidrome's suite tests (multi-user isolation, per-endpoint folder filtering, permission
refusals) and gonic's named album shapes (compilation, multi-disc, collaboration credit, empty album, Unicode)
(`[servers]` §10.3).

## 4. The `musicFolderId` endpoint list

**Answer: ten endpoints.** `getIndexes`, `getArtists`, `getAlbumList`, `getAlbumList2`, `getRandomSongs`,
`getSongsByGenre`, `getStarred`, `getStarred2`, `search2`, `search3`. This is the specification's own list, and both
oracles honor exactly these and ignore the parameter elsewhere (`[servers]` §2, [D], [S], and [O] at fixtures F1 and
F2). Three readings settle with it:

- **An unknown id answers error 70**, as Navidrome does, and an id for a collection the member cannot see gets the
  same answer and text, so a private collection's existence never leaks. gonic's empty success hides a stale id after a
  reconfiguration (`[servers]` §2, "Lesson"; `[ledger]` L-19). Private collections ship at first launch (re-plan
  "Rung 14").
- **The id is a stable integer per collection, never reused** (`[contract]` E7; `[ledger]` L-18; hook 22 in
  section 17). gonic's index-into-the-path-list id changes meaning when paths are reordered (`[servers]` §2).
- **The parameter is load-bearing for one client.** Rufin's catalog sync runs `search3` once per music folder and links
  each track to the folder it came from, so `search3` must filter exactly (`[clients]` §4). Feishin and Tempus pass the
  member's chosen folder on every list and search (`[clients]` §4).

PR #162 would add `getGenres`, `getSimilarSongs2`, `getTopSongs`, `getArtist`, `getShares`, and `getBookmarks`; it has
been idle since 2026-03-22 (`[contract]` Q4). A merge reopens this list at slice 3, not before.

**Routed to slice 3:** a repeated `musicFolderId` (Navidrome reads a list in source; the specification documents one
value, `[servers]` §2, "Not observed") and the two-collection fixture that proves the filter (`[servers]` "What could
not be done").

## 5. The XML validation rule

**Answer: the legacy form, with Task 1's scoped projection, and the XML shape of every OpenSubsonic addition declared
as an unvalidated class** (`[contract]` Q3, "Recommended reading"). Settled, with Task 1's three conditions:

1. The projection is scoped per element type, keyed on the pinned OpenAPI. A global projection strips `status` and
   `version` from the envelope and fails every body for the wrong reason (`[contract]` Q3, "The two candidate rules").
2. A name found in neither the XSD nor the OpenAPI type stays in place and fails, so a misspelled attribute is caught.
3. The XML shape of the additions follows one rule, E4's convention (scalars as attributes, lists of records as repeated
   elements with no wrapper, lists of strings as repeated text elements), and the oracle diff covers the fields an
   oracle emits (`[contract]` E4; `[ledger]` L-02, L-04).

**Reasons.** The legacy form proves conformance to the one authoritative XML schema. A derived XSD would prove only
self-consistency, because no authority defines an XML form for the additions and the oracles disagree on it; a
derived XSD generated by the encoder's own rule is circular (`[contract]` Q3). The derived form also needs a license
answer for a modified XSD, and the XSD carries no license text (`[contract]` Q3, "The XSD's license"; `[cells]`
OpenAPI row). The harness therefore fetches the unmodified XSD into a pinned cache by sha256 and never redistributes
it (routed to slice 2, `[contract]` "Routed").

**Tension (g), resolved.** Amperfy parses XML, so the XML encoder is load-bearing at launch. Its demands fall on core
attributes the legacy validation covers (`isDir` as a literal boolean, `version` on the root, `albumCount` on each
artist) and on value formats no XSD pins down: `created` and `starred` must be ISO 8601 with fractional seconds or
Amperfy drops the date (`[clients]` §1). That last one is a trap for a Go encoder. `go doc time.Layout` says a
decimal point "followed by one or more nines represents a fractional second, printed to the given number of decimal
places, with trailing zeros removed", and `time.RFC3339Nano` uses nines, so a whole-second timestamp loses its
fraction. Slice 3's encoder uses a fixed-width layout (zeros, not nines) and a roster-tier test proves Amperfy keeps the
date. The XML oracle diff runs at the same weight as the JSON one.

## 6. The `transcoding` extension: owner fork A

**The question.** Whether 0.8 answers `getTranscodeDecision` and `getTranscodeStream` (`:412-419`, `:941-942`).

**The evidence.** Only Feishin calls the extension in the roster (`[clients]` §3). When the server declares it,
Feishin resolves every stream through a JSON POST decision, unless the member has turned transcoding on explicitly or
uses the MPV player, and it must never meet error 0 there: error 0 arrives as HTTP 200, Feishin skips its fallback,
dereferences an undefined decision, and the track does not play (`[clients]` §2a, `getTranscodeDecision` row).
Without the extension, Feishin builds a plain `stream` URL with whatever format and bitrate its own transcode setting
appends (off by default) (`[clients]` §3, "What the extension list does"). Against Navidrome, which declares it, the
decision for the FLAC fixture came back direct play and the track played (`[base-url]` Feishin walk, step 4). At
FLAC-only launch, Feishin's own probe list names FLAC as direct-playable (`[clients]` §3, last paragraph). The
implementation cost is two endpoints, a closed-list validator, a signed stateless token bound to the file's update
time, a decision function, and the device-ceiling rule (`[servers]` §4.4).

**If yes.**

- Slice 5 grows by those five pieces and by a token-signing secret. The secret follows the auth secrets' pattern and
  is minted by slice 8's `init` (`:710-712`), so the operator delta stays zero.
- Feishin's every stream goes through a POST, so the CORS preflight path becomes live for its web build
  (section 8), and a defect anywhere in the decision path makes tracks unplayable in Feishin rather than merely
  untranscoded.
- How a posted `ClientInfo` meets the device profile takes Navidrome's ordering, which this record adopts as the
  conditional reading: the client's document negotiates, and the device profile is a ceiling on bitrate and a
  preferred format only if the client can play it (`[servers]` §4.3, §4.4 lesson 2; `[ledger]` L-36). One decision path
  serves the extension and the legacy parameters (`[servers]` §4.4 lesson 3). Every error on the decision half is an
  envelope and every error on the stream half is a plain HTTP status, as Navidrome does (`[servers]` §4.4 lesson 4).
- AAC through the extension follows section 9's ADTS ruling; Navidrome withholds AAC from the extension for a client
  population dubplate lacks (`[servers]` §4.3; `[ledger]` L-35).

**If no.**

- Feishin plays through plain `stream` URLs, as it does against gonic, and its manual transcode setting reaches the
  server through the legacy `format` and `maxBitRate` parameters that slice 5 builds anyway.
- The `ClientInfo` interaction question is moot until the extension is taken, and the posted-document path has no
  test surface at 0.8.
- Symfonium, outside the roster, is reported to call the extension (`:415-416`, unverified by the survey), and it keeps
  working through its client-side transcoding controls (`[clients]` §8).

**Recommendation: no at 0.8; revisit at 1.0 with the format ladder.** At FLAC-only launch, the extension buys nothing
the roster uses. Feishin direct-plays FLAC, and every other roster client ignores the extension. What it costs is a
new failure mode in which Feishin cannot play a track at all. The format ladder (`CLAUDE.md`, "Versions") is when
server-side negotiation earns its keep, since that is when sources arrive that a browser engine may not play.

## 7. Range on a filling transcode

**Answer: match Navidrome.** On a filling cache entry, `stream` answers 200 from byte 0 with `Accept-Ranges: none` and
no `Content-Length`, and ignores any `Range` header. On a complete entry it serves ranges as a static file. Seeking
during a transcode rides `timeOffset` under the declared `transcodeOffset` extension (section 2), with the offset in
the cache key. This retires the spec's sentence at `:434-435` ("A Range request on a filling entry is served from
that entry once the requested offset is written") and the revision 2.1 conductor ruling 6's "serves ranges from the
filling entry" clause (`:138`); the eviction and queue halves of that ruling stand. Task 9 amends `:434-443`.

**Reasons.**

1. No oracle serves a range from a filling entry. Navidrome answers 200 with the whole entry; gonic blocks until the
   encode ends (`[servers]` §3.2, observed at the filling state with the entry's file size quoted).
2. The roster has been tested against Navidrome's shape. Amperfy is the client this matters for: its default request
   is `format=mp3`, so every default Amperfy play is a transcode (`[clients]` §5), and its player decides seekability
   from status 206 or an `Accept-Ranges` value other than `none` (`[clients]` §5, Amperfy bullet). A range answer on a
   filling entry would need a `Content-Range` with an unknown total, and how Amperfy's library treats an unknown total is
   unconfirmed (it reads the length from that total, `[clients]` §5). Navidrome's shape is the one Amperfy is known to
   play today.
3. HTTP permits it. RFC 9110 §14.2: "A server MAY ignore the Range header field."
4. The standard library's file serving cannot serve a filling entry (`:439-441`), so the divergent design needed a
   hand-rolled range writer. The matched design needs none.

**Two adjacent readings, settled in the ledger.** `estimateContentLength` is not honored at 0.8: no roster client
sends it, and both oracles' estimates run about 2.4% high, which a strict client reads as truncation (`[clients]` §5,
end; `[servers]` §3.3; `[ledger]` L-29). And the spec's "the request returns 500" for a failed encode (`:447-449`)
holds only before the first byte. After headers are written, the server aborts the connection so the client never sees
a clean end, as Navidrome does and gonic does not (`[servers]` §3.4; `[ledger]` L-31). `go doc net/http.ErrAbortHandler`
names the mechanism: "a sentinel panic value to abort a handler".

**Routed to slice 5:** Amperfy opens a `Range: bytes=0-1` probe on each track, reads the headers, and cancels before it
sends the real request (`[clients]` §5, Amperfy probe bullet). Under the spec's rule that a disconnect kills ffmpeg
when no other reader is attached (`:449-450`), the probe's cancel can kill the encode it just started, so every
Amperfy track would pay two encode starts. Slice 5's race test covers this pattern and weighs a short grace before
killing an encode that has no readers (`[ledger]` L-32).

## 8. The CORS policy

**Answer.** Settled, for every response under the Subsonic sub-prefix, success, error, and binary alike:

- `Access-Control-Allow-Origin: *` on every response, with or without an `Origin` header. The value is constant, so
  no `Vary: Origin` is needed.
- `Access-Control-Allow-Credentials` is never sent.
- A preflight answers with no body and no data, needs no credentials, and allows the methods `GET` and `POST` and the
  request header `Content-Type`.
- Nothing outside the Subsonic sub-prefix (the web app, the admin API, the share page) sends CORS headers.

**Reasons.** Only Feishin runs in a browser engine, and none of its requests is credentialed: axios sets no
`withCredentials`, `fetch` uses the default mode, and the audio element is `anonymous` (`[clients]` §6). Its cover
`fetch` and its audio element both need `Access-Control-Allow-Origin` on the response (`[clients]` §6 table).
Feishin's origin cannot be known ahead (a hosted build, an Electron `file://` page, or `localhost`), and the walk saw
`Origin: file://` (`[base-url]` header table), so a named-origin list cannot work and a wildcard can. Credentials ride
in the query or the form body for every client (`[clients]` §6, "Credential channels"), so the browser's
no-credentials rule for a wildcard costs nothing. Both oracles run the same policy in effect: any origin, no
credentials (`[servers]` §5). Stamping every response follows gonic and drops Navidrome's per-request reflection,
because a constant answer is simpler to test and to cache (`[servers]` §5.2 lesson 2). The fixed header list follows
gonic too, trimmed to the one header a roster request sends, `Content-Type` on the transcode decision POST
(`[clients]` §6). Keeping the web app outside the policy means its cookie-authenticated routes never answer a
cross-origin script.

**Observed:** neither walked client sent a preflight, Feishin's `getTranscodeDecision` POST included (`[base-url]`
header table). The preflight answer therefore guards the hosted web build, not the desktop build. **Test owed (slice
3):** the preflight rows, the no-credentials assertion, the no-data assertion (`[ledger]` A-03, L-38), and whether a
CORS-mode media element needs `Access-Control-Expose-Headers` for `Content-Range`, which is unconfirmed (`[clients]`
"Question outcomes", row 5).

## 9. The AAC container

**Answer: ADTS, with `Content-Type: audio/aac`.** ADTS is documented for all three native players: Media3's default
extractors list it for Tempus, GStreamer's `aacparse` for Rufin, and `AudioStreaming`'s README and type table for
Amperfy, which maps `audio/aac` to ADTS by name. Fragmented MP4 is documented only for Tempus. Feishin's browser engine
is the fourth player, and its probe does not distinguish the two (`[clients]` §5, "The AAC container item").
Navidrome's shipped AAC configuration also writes ADTS (`[servers]` §4.3).

**Test owed:** play an ADTS stream with a growing length in Feishin and Rufin at slice 5's roster walk, in Tempus once
the emulator exists, and in Amperfy at rung 16 (`[clients]` "Question outcomes", row 4; tension (f)).

## 10. The protocol module's license and NOTICE: owner fork B

**The question.** MIT or Apache-2.0 for the separable OpenSubsonic module, and what notice it carries from the
Apache-2.0 OpenAPI file it validates against (`:327-328`, `:947-948`).

**The open fact, answered.** The OpenSubsonic repository carries no NOTICE file, so §4(d) attaches nothing today;
§4(a) to §4(c) (the license text and notices on modified files) bind any redistribution of the file (`[cells]` flag 1;
tension (c)). The module's conformance checker is the harness's contract-tier validator, which loads the pinned
OpenAPI bundle (`:304-307`).

**If MIT, and the checker loads the OpenAPI file as caller input.** The checker takes the bundled OpenAPI file from its
caller and verifies a pinned checksum, the way the harness treats the XSD (section 5). The module redistributes no
Apache-2.0 file, so no Apache obligation attaches to it. dubplate and the module share one license, which keeps
Geoff's 2026-09-07 reasons intact: the patent grant and the NOTICE clause "buy nothing here and add friction for a
stranger forking the estate" (`:345-347`). The module's users fetch the OpenAPI file themselves, which the module's
README must say. If the module ever embeds the file, it ships under its own `LICENSE` in a subdirectory with any
modification marked, which an MIT module may do.

**If Apache-2.0.** The module matches the upstream specification's license and carries a patent grant. Every fork of
the module then carries the §4 obligations, plus a NOTICE file if the module adds one. The upstream fit is smaller than
it looks: MIT code can already be contributed into an Apache-2.0 project, so Apache-2.0 buys no upstreaming path MIT
lacks.

**Recommendation: MIT, with the checker loading the OpenAPI file as caller input, never vendored.** Geoff's 2026-09-07
reasons carry over unchanged (`:345-347`), and keeping the file out of the module keeps the module single-licensed. The
watch item: re-check for an upstream NOTICE at every specification pin bump, which slice 2's pin re-read already
performs (`[contract]` "Routed").

## 11. The word "collection" and the PR watch

**State on 2026-10-07** (`[contract]` Q4): none of #162, #237, #258, #261 has merged, `main` is still `bed1688`, and 21
PRs are open. #237 ("Add collections") was last active on 2026-09-20; #258 ("extended library browsing") on 2026-09-25;
#162 on 2026-03-22; #261 on 2026-07-20. Newer PRs that bear on dubplate: #278 (authentication rewrite, recommends API
keys and `formPost`), #255 (`playlistFolders`), #256 (`artistRoleFilter`), #202 (a `scrobble` batch with an invalid id
answers error 70 and records nothing).

**Settled.**

1. **dubplate never declares the `collections` extension.** #237's collection is a per-user ordered list of items with
   a `public` flag (`[contract]` Q4). Playlists already cover a member's lists, and a dubplate surface offering a
   "public collection" reads as the refused feature (`CLAUDE.md`, "Refused features": "public collections").
2. **API-facing prose says "music folder" for the wire concept and "collection" only for the store concept**, so a
   merged #237 changes no wire text (`[contract]` Q4, consequence 2). Slice 3's API documentation carries the rule.
3. **dubplate never declares `publicImageUrl` (#261).** It converts a cover-art request into a link anyone can fetch,
   which collides with the one-unauthenticated-endpoint rule (`:648`) and the refused public library (`[contract]` Q4).
4. **#202's direction is adopted for `scrobble`** (`[ledger]` L-43); **#278 is slice 6's input** for the credential
   modes.
5. **A rename of dubplate's partition to "library"** becomes Geoff's taste question only if #237 merges before launch
   and a roster client shows both concepts under one label (`[contract]` Q4, consequence 3). Today it is not a fork.

**Routed to slice 2:** the re-read of all four PRs and #278 before the harness pins a specification build (slice 2's plan
Task 1 carries it; `[contract]` "Routed").

## 12. Folder browsing at launch: owner fork C

**The question.** Whether 0.8 ships the folder view (`getIndexes` and `getMusicDirectory`) (`:375-377`, `:925-926`).

**The evidence.** All four roster clients offer a folder view beside the ID3 one, and all four use it through the same
three calls; each browses ID3 first (`[clients]` §4). None needs the folder view for library use: with no folders
served, Tempus hides its sector, Amperfy shows an empty Directories screen, and Feishin and Rufin show an empty Folders
page (`[clients]` §4, "Reading for the folder-browsing item"). Collections are music folders, so `getMusicFolders`
answers real data either way (section 4). Navidrome itself serves only a synthesized Artist/Album/Track tree built
from tags and documents that it has no plan to support browse-by-folder (`[servers]` §8, quoting
`navidrome/website@b0fe572`). gonic serves the real tree with typed ids (`[servers]` §8).

**If yes.** Slice 3 builds the folder projection the spec describes: directory rows minted from the pipeline's paths,
`getIndexes` and `getMusicDirectory` with parent chains, a second id space, gonic as the grader, and the folder-view
divergence class live in the harness (`:375-379`, `:285-286`; `[ledger]` L-51, L-52). The `download` and share rules
that treat a one-album directory as its album (`:663-664`, `:523-524`) stay live. One observation narrows the cost.
Wherever a collection's layout template equals the tag layout, the real tree and Navidrome's synthesized tree coincide
(`[servers]` §8), so a cheaper form exists: a tree served from the artist and album ids. That form would change the
spec's "folder ids are minted rows" (`:379`) and make Navidrome the grader, and slice 3 would weigh it.

**If no.** `getIndexes` answers an empty index list (section 1), and `getMusicDirectory` answers error 70 for any id,
since no directory exists. The folder-view class and gonic's grading of `getIndexes` and `getMusicDirectory` go idle,
and gonic still grades `getMusicFolders`. The one-album-directory clauses in `download` and shares are moot. Slice 3
decides what a song's `parent` field carries; omitting it keeps clients from linking to a directory that does not
exist. A member who browses Directories in Amperfy (a default library entry) sees each collection and an empty list
inside it.

**Recommendation: no at 0.8; reconsider at 1.0 if a member asks.** The pipeline chooses every path from tags
(`:375-377`), so a folder view shows no organization a member made, only a second route to what the ID3 view already
shows. It would cost a second id space and a whole harness class. Leaving it out costs one empty screen per client, with
no broken flow (`[clients]` §4).

## 13. The `.nsp` format

**Answer: recorded.** A `.nsp` file is a JSON rule document: a name and comment, a nestable `all` or `any` rule group,
`sort`, `order`, `limit` or `limitPercent`, and `refreshDelay`. Its 17 operators range over about 70 track, album,
artist, history, and MusicBrainz fields. Navidrome imports `.nsp` files during a library scan, assigns them to the
first admin, serves them as read-only playlists, and evaluates them lazily at first access (`[servers]` §9, quoting
`navidrome/website@b0fe572` and observing an import at fixture F2).

**Settled.** dubplate's smart playlists ship at launch in dubplate's own rule form (`:597-602`), never as `.nsp`
evaluation. A switcher import is a parse and a translation into that form: rules over fields dubplate holds translate,
and every rule that does not is reported by name, so nothing is silently merged (`:619-621`). dubplate's rule fields
(genre, year, added date, added by, collection, starred, play count, last played; `:598-600`) are a subset of `.nsp`'s,
so a partial translation is the expected case.

**Routed to slice 8:** whether the retired Navidrome held any `.nsp` file or smart-playlist rows, which only the
archive can say and this lane did not open (settled item 2; `[servers]` §9), and whether Navidrome keeps a smart
playlist's rules in its database as well as the file, which decides what the import reads. It becomes an owner fork
only if slice 8 finds the group's rules need a field dubplate lacks.

## 14. The API base URL: the measurement, for slice 3

This lane measures and slice 3 decides, with a build-spec erratum (re-plan `:427`; plan settled item 8). The summary,
from `[base-url]`:

- **Every roster client carries a path in its server URL.** Feishin and Rufin worked live against Navidrome mounted at
  `/dubplate`, with all 57 and 30 logged requests under `/dubplate/rest/` and none at the root. Amperfy and Tempus
  carry the path by source reading (`[base-url]` Verdicts).
- **Both suffix forms must answer** (tension (e)).
- **Amperfy's precondition.** Before login, Amperfy sends a plain GET to the configured base URL itself, with no
  `/rest` and no credentials, and fails the login on any final status of 400 or more except 401
  (`[base-url]` Amperfy, "A precondition found"). So the sub-prefix a client is given must itself answer under 400.
- **Server-emitted links carry the prefix.** Tempus shows the share `url` the server returns, so share URLs must carry
  the mount prefix, as Navidrome's do (`[base-url]` Tempus, "Second code paths"; "What the servers document and do").
- **Trailing slashes.** One typed trailing slash is handled by all four; two are unconfirmed (`[base-url]` "Not done").
- **Feishin's "Public URL".** When a server entry prefers it, API and cover calls use it while stream, download, and
  transcoded-stream URLs keep the raw stored URL, so the two can diverge (`[base-url]` header table, "Stream URL
  against the API base"). With no Public URL set, they agree.
- **The servers split.** Navidrome mounts its API under `ND_BASEURL` and redirects an unprefixed request to its web UI
  with a 302. gonic's prefix option leaves the API at the root and expects a stripping proxy (`[pins]` "Base-path
  option, quoted"; `[base-url]` "What the servers document and do").
- **No CORS preflight and no credential cookie** in either walk; credentials rode the query (and Feishin's form body)
  everywhere (`[base-url]` header table; "Cross-client facts").
- **Cookie scope.** If dubplate ever sets a cookie on an API response, its path is the mount prefix, as Navidrome's is
  under a base path (`[base-url]` "Navidrome's `Set-Cookie` under a base path"). The settled reading is stricter:
  dubplate sets none (`[ledger]` L-37).

The spec's shape (the Subsonic path as a sub-prefix of the one owned prefix, `:643-646`) meets every measured client
demand. The open decision is slice 3's.

## 15. Melodee's intake pipeline

**Answer** (`[melodee]` "The one-paragraph answer"): Melodee runs four timer-driven filesystem stages (inbound,
staging, storage, database), keeps state in sidecar files and directory names, and admits on tag completeness with no
contributor present. dubplate's door reverses each of those choices. No verdict flips. Fifteen lessons are routed by
rung, and none is slice 1's to build (`[melodee]` "Lessons for dubplate"; "Routed"):

- Rung 7: L2, L3, L4, L9, L12 (completeness checks as cheap reject reasons, disc-aware numbering, a check that errors
  rejects, a per-reason blocking flag, the verdict at the door).
- Rung 8: L3 (no repair rewrites disc or track tags).
- Rung 9: L1, L5, L6, L7, L11, L13, L14, L15 (a stable-file rule for the timer path, a stateless checker, an atomic and
  verified commit with the move-intent journal, no merge into an admitted album, bounded enrichment, admission and
  indexing as one step, the admitted event at the commit, content-walk discovery).
- Rung 13: L8, L10 (verdicts on records, the `defer` re-ask schedule with a reset rule).
- Slice 7: F15 supports the spec's report-a-problem playlist over a second queue.

**L14's hook question, settled here so Task 9 needs no wording change.** Hook 4 requires an event kind and a timestamp
on any event-shaped table and names no closed list of kinds (`:783-784`); hook 9 requires a projector delta to commit
in the same transaction as its event (`:797-798`). An album-admitted event fits both as written. Rung 9's plan names the
kind. The same events give slice 3 its `lastScan` marker (`[ledger]` L-20).

## 16. The prior-art table's open cells

Task 9 fills the cells from `[cells]` "The table, cell by cell" and "Dated figures to refresh at Task 9", with every
license URL as re-checked there. No verdict changes. Task 6's five flags are ruled:

1. **OpenSubsonic row.** The Verdict's reason is corrected as in tension (c). Settled.
2. **Melodee row, Verdict.** "Adopt as readable and reusable prior art" gains a condition: any code reuse needs a
   per-file provenance read first. Two source comments say a parameter and a flow follow "Navidrome semantics" and were
   "seen in the Navidrome source code" (`[cells]` "Defects", last bullet), and the GPL clean-room rule binds anything
   derived from that. Settled.
3. **Melodee row, Maintenance.** "Active" becomes "Bursty: 659 commits in December and January, 21 from February to
   April, 79 from May to July, none since 2026-07-12" (`[cells]` flag 3). Settled.
4. **Picard row.** Cite the generated `tag_mapping.rst` and the rendered v3.0 page at `picard-docs@35f376e`, never the
   GPL-2.0-or-later generator script (`[cells]` flag 4). Settled.
5. **modernc row.** A standing condition joins the Verdict: `modernc.org/libc` moves only together with
   `modernc.org/sqlite`, at the version the sqlite module pins (`[cells]` flag 5). Routed to the next dependency sweep
   (6d's Task 0) to check, and to the conductor as a candidate gate leg, since a rule lives where it executes.

**The Melodee quality cell, merged (tension (i)).** Task 9 places this text:

> Below the house bar in structure, assertion discipline, auth, and its commit step, though broad in test volume.
> Read at v2.2.0 (`d91342e`): 4,479 test attributes across six projects, yet no test moves an album through the
> admission path, and the streaming compatibility tests accept error statuses as passes. Validation fails open on a
> thrown check; the commit copies then deletes without verification and swallows per-album errors; a default repair
> rewrites multi-disc tags; an auth carve-out keys on the client's `Host` header and a path substring (a source
> reading, not run). Worth taking as ideas: the revalidation backoff and the stable-file check.

Each clause cites its evidence: `[melodee]` "The prior-art table's Melodee quality cell" (the test count, the untested
move and merge, fail-open validation, the unverified commit, the multi-disc rewrite) and `[cells]` "Melodee's quality
cell" (the vacuous streaming assertions, the `Host`-header carve-out). The two records' commit totals differ by four
(1,474 at the tag against 1,478 on `main`), because `main` carries commits after the v2.2.0 tag (`[melodee]`
"Two authorship facts"; `[cells]` "Maintenance").

## 17. Forward hooks

**One new hook, routed to Task 9 as hook 22:**

> 22. A stable integer wire id per collection, slice 3. The OpenSubsonic music folder id is an integer
> (`xsd:81`, `openapi/schemas/MusicFolder.json`), and clients keep a chosen folder id in their settings, so each
> collection carries an integer that survives restarts and `VACUUM` and is never reused after a delete. The store's
> collection id is a BLOB, and SQLite's implicit rowid does not qualify: "The VACUUM command may change the ROWIDs of
> entries in any tables that do not have an explicit INTEGER PRIMARY KEY". Built by slice 3, its first consumer; a rung
> that creates collections before slice 3 is neutral, because slice 3's migration backfills.

Evidence: `[contract]` E7 (the integer type, both oracles' values); `[clients]` §4 (Feishin's server-level folder
choice, Tempus's stored active folder); `internal/engine/store/migrations/0003_catalog.sql:12-13` (`id BLOB NOT NULL
PRIMARY KEY`); `https://www.sqlite.org/lang_vacuum.html` (read 2026-10-07).

**Weighed and not new.** The catalog's last-change time for `getScanStatus` (hooks 4 and 9 cover it, section 15). The
transcode token bound to the file (hook 1's file hash, if fork A is ever yes). An album-admitted event kind (section
15). Every segment A record reported no hook (`[servers]` "Routed"; `[melodee]` "Routed" (b); `[contract]` "Prior art
consulted").

## 18. Readings the evidence forces

The ledger settles every row Task 4 left "to settle" and every candidate Task 1 raised; each row carries its reason.
Seven readings reach beyond one row, so they are named here.

1. **The device profile is a ceiling, not the last resort** (`[ledger]` L-36). The spec resolves profiles "in order
   from the request, the device, and the estate default" (`:411-412`). Read against the roster's defaults, that order
   leaves the device profile almost dead: Tempus and Rufin send `format=raw` by default, and Amperfy sends `format=mp3`,
   so a request names a format for three of four clients before the device profile is consulted (`[clients]` §5).
   **Settled:** the device profile is a bitrate ceiling and a default format; a request may lower the ceiling, never
   raise it, and `format=raw` yields the original only when it fits under the ceiling. This is Navidrome's ordering
   (`[servers]` §4.3), and one decision path serves the legacy parameters and, if taken, the extension. Task 9 amends
   `:411-412`; slice 5 designs the details.
2. **JSONP is declined** (`[ledger]` L-01). No roster client sends `f=jsonp` (`[clients]` §1, "Response format asked"),
   gonic shows how a reflected callback becomes script injection on a response carrying member data (`[contract]` E3),
   and section 8's CORS policy serves browser clients. `f=jsonp` gets the same answer as any unrecognized `f`, and
   every API response carries `X-Content-Type-Options: nosniff`.
3. **The plain and hex password forms are refused with error 42** (`:491`), against both oracles (`[ledger]` L-11).
   Every roster client also has a salted-token mode (`[clients]` §6). **Routed to slice 6:** prove each roster client
   logs in with the token form by default; a client that defaults to `p` turns the rule into a slice 6 owner question.
   Also slice 6's: whether a credential present in both the query and a form body is error 43, since `go doc
   net/http.Request.ParseForm` says "Request body parameters take precedence over URL query string values in r.Form"
   (`[ledger]` L-13).
4. **Errors before and after the first byte differ** (`[ledger]` L-31; section 7).
5. **Every missing required parameter answers error 10, and every id the server did not issue answers error 70**
   (`[ledger]` L-15, L-16). Opaque ids have no malformed form.
6. **Writes that fail are loud.** An unparseable `submission`, an unknown id in a `scrobble` batch or a saved queue,
   and a `current` outside the saved list each answer an error and record nothing, because play history is
   append-only and a silent guess cannot be undone (`[ledger]` L-42 to L-44, L-46, L-47).
7. **`getScanStatus` carries `lastScan`** set to the catalog's last change, so Rufin caches instead of re-walking the
   library on every launch (`[clients]` §2b, `getScanStatus` row). It is a Navidrome vendor field outside both
   contracts (`[contract]` Q3, "Residuals"), so it joins the contract tier's known-deviation list, which follows
   Ampache's self-expiring method (`[contract]` Q3, "How Ampache does it") (`[ledger]` L-20).

## 19. The owner bundle

For Task 9's one combined checkpoint question. Each fork names both answers' consequences in its section.

| Fork | Recommendation | Section |
|---|---|---|
| A. The `transcoding` extension at 0.8 | No; revisit at 1.0 with the format ladder | 6 |
| B. The protocol module's license | MIT, with the checker loading the OpenAPI file as caller input, never vendored | 10 |
| C. Folder browsing at launch | No; reconsider at 1.0 if a member asks | 12 |

**One hardware request, not a ruling:** enable VT-x in the ThinkPad's firmware at the next sitting at the machine. It
unblocks the Android emulator for Tempus's live walk in slice 2 and every later slice's roster walk (`[pins]` "Not
done"; tension (f)).

## Routed

- **Slice 2 (harness):** the ledger, transcribed mechanically into its machine form (plan settled item 6); the
  per-endpoint oracle map (section 3); the XSD fetched into a pinned cache and the scoped projection (section 5); the
  PR re-read before pinning (section 11); the known-deviation list with `lastScan` and the oracle nulls (`[contract]`
  "Routed"); the area list (`[servers]` §10.3); a two-collection fixture (section 4); Tempus's live walk once VT-x is on.
- **Slice 3 (projection and browsing):** both suffix forms; the base-URL decision and its erratum (section 14);
  error 0 for unimplemented endpoints and the empty-success set; the musicFolderId ten, error 70 for unknown or
  invisible ids, and hook 22; the CORS policy; the fixed-width timestamp layout; `contentType` on every song (Feishin
  throws without it, `[clients]` §1); `albumCount` on each artist (Amperfy); `lastScan` and `startScan`; the
  `parent` field under fork C; search edges and size validation; the encoding rules E1 to E10 (`[contract]`).
- **Slice 4 (art and search):** WebP is triggered by Feishin's `<img>` loads, and a format-varying `getCoverArt` sends
  `Vary: Accept` (tension (h); `[ledger]` L-39); a missing `id` is error 10 (`[ledger]` L-40); Amperfy parses an
  artwork body under 2,000 bytes as a possible error (`[clients]` "Routed", slice 4).
- **Slice 5 (streaming):** section 7's Range reading, `timeOffset` in the cache key, failure before and after the first
  byte, Amperfy's probe and the disconnect rule, the bounded queue, the device ceiling (section 18, item 1), the ADTS
  container and its playback test, `estimateContentLength` not honored, and fork A's conditional readings if it is
  ever yes.
- **Slice 6 (identity):** the apiKey path and `tokenInfo`; error 42's roster check; credential conflict and error 43;
  one error 40 text for a wrong user or password; `getUser` answering 50 for another member's username whether or not
  it exists; PR #278; Melodee's `Host`-header carve-out as the case for an enumerated allow-list, never a path pattern
  or a client header (`[cells]` "Routed").
- **Slice 7 (per-user state):** the queue and `scrobble` rows (`[ledger]` L-41 to L-49); `getStarred` carries data for
  Feishin; the smart-playlist row; Melodee F15 for the report path.
- **Slice 8 (import):** the `.nsp` reads (section 13).
- **Rungs 7, 8, 9, 13:** Melodee's lessons (section 15).
- **Task 9 (spec revision 3.1):** mark each item answered or answered pending ruling with a pointer here; amend
  `:411-412` (section 18, item 1), `:434-443` (section 7), `:447-449` (section 7), and `:866` (tension (c)); add
  hook 22; fill the prior-art cells and the merged Melodee cell (section 16); bundle section 19.
- **Conductor:** a candidate gate leg for the modernc `libc` pairing (section 16, flag 5).

## Not done, and unconfirmed

- No live proof exists for Tempus or Amperfy (tension (f)).
- The cost of fork C's cheaper form (a tree on ID3 ids) was not measured; slice 3 measures it if the fork is yes.
- Whether a CORS-mode media element needs exposed headers is unconfirmed (section 8).
- How Amperfy's player treats a `Content-Range` with an unknown total is unconfirmed (section 7, reason 2). The
  settled reading avoids that case, so nothing waits on it.
- Every unconfirmed point in the segment A records stays as each record states it; this record adds no new
  observation.

## Prior art consulted, and the clean-room check

The seven segment A records, which cite Navidrome, gonic, and the four GPL clients only by `repo@commit:path:line`
and observation. This record cites those records and the Apache-2.0 specification, and it quotes only primary
permissive or public sources: RFC 9110, SQLite's documentation, and Go's `go doc` output. No text, code, pseudo-code, or
identifier list from a copyleft repository appears here; endpoint, parameter, and extension names are the
OpenSubsonic wire vocabulary the specification defines. Nothing is adopted or declined anew beyond the rulings on
Task 6's flags (section 16). Neutral on hooks except hook 22, routed. Operator-supplied values: delta zero.

## Fact-read fold

Task 8's fresh-context read returned "sound with 5 fixes". Each was checked against its source before the edit, and
all five are folded. No finding was refused.

1. **Folded.** The "unknown id" probes sent `bogus`, which gonic's typed ids cannot parse (`gonic@5ae99e8:server/ctrlsubsonic/handlers_common.go:58-61`). The rows now say so: `[servers]` §7.1 and §7.2 tables, lesson 4 of §7.3, and D4-29 and D4-33. The well-formed case is added as a source reading with a test owed: `scrobble` answers error 0 "error finding track" (`:71-73`), and `savePlayQueue` stores the id without a lookup (`:260-288`). Ledger L-43 and L-47 correct the gonic column, and L-16 notes the exception for these two writes. No disposition changes.
2. **Folded.** L-47's claim that Navidrome resets `current` is deleted. The probe sent `current=a`, and Navidrome's save stores an index into the id list as sent (`navidrome@1011457:server/subsonic/bookmarks.go:112-147`), so the claim had no support.
3. **Folded.** `[servers]` §6.2 keeps the precedence order of Navidrome's search rules with one file citation, and drops the per-branch line ranges and the paraphrase of a source comment's rationale.
4. **Folded.** `[base-url]`'s header table now reads "Both forms; Task 2 predicted `*/*` only", matching `[clients]` §7.
5. **Folded.** L-01 now qualifies Navidrome's JSONP error 0: two bad callbacks observed, four inferred from the content type (`[contract]` E3).
