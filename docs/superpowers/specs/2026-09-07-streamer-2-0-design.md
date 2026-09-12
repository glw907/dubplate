# dubplate-server 2.0: the built-in streamer

Status: **Ratified 2026-09-07.** Revision 2.2, 2026-09-12 (forward hooks 1 and 15 and the web surface amended from the configuration homes spec, `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`). Revision 2.1, 2026-09-07 (the three owner rulings recorded in "Rulings on the review round"). Revision 2, 2026-09-07. Every decision below was taken with
Geoff in the 2026-09-07 brainstorm and none is reopened here. ROADMAP entry: `streamer-2-0`, sequenced after the
greenfield ladder's launch (rung 17).

This document governs the initiative's purview, its architecture, and the forward hooks the current build owes it. Where
it names a current rung (5, 6, 7, 8, 9, 12, 16), it asks that rung's plan to weigh a hook, never to build a 2.0 feature
early. One sentence in revision 1 trimmed rung 12 while the document claimed to amend nothing; that conflict is open
ruling 1 rather than a silent amendment.

## Revision 2, 2026-09-07

Revision 2 folds four reviews: the adversarial review (`research/2026-09-07-streamer-2-0-review-adversarial.md`, B1 to
B5, M1 to M15, minors 1 to 14), the comparison review (`-review-comparison.md`, findings 1 to 10), the prior-art review
(`-review-prior-art.md`, F1 to F10), and the DAP review (`-review-daps.md`, findings 1 to 10). The conductor ruled the
twelve items below; each names the finding it answers.

1. Entity ids are opaque server-issued surrogates. The release MBID and the content hash are indexed alternate keys, so
   a retag re-keys nothing and no per-user row dangles (B5, M10).
2. A play is recorded on `scrobble` with `submission=true`, matching the oracle. A stream start is a separate event kind
   and never a play (comparison finding 7, M4).
3. Playlists use `updatePlaylist`'s delta semantics against the stored order, with a version column per playlist for
   concurrent clients (M5).
4. Unimplemented endpoints return error code 0 inside a well-formed `subsonic-response`, except the endpoints the study
   rung lists as requiring an empty success (B4, minors 1, 10, 11, 13).
5. App passwords stay encrypted at rest for the legacy token path, gain a hashed modern-only mode, and are revoked by a
   per-user valid-after epoch (M1, M2, M3, M6, M7, M9).
6. The transcode miss path serves ranges from the filling entry, evicts a failed encode, and caps the queue (M8).
7. Forward hooks are re-keyed to their real rungs and extended with the rung 6, 7, 8, and 9 requirements the reviews
   found (B2, B3, F7, DAP findings 2 to 8, M12).
8. Evidence gains a contract tier above the oracle, gonic as a second oracle for the folder view, a pinned oracle
   version, and two declared divergence classes (comparison findings 1, 2, 3, 9, M15).
9. Prior art is re-scored: two license strings corrected, three entries declined, one adopted under a condition, six
   added (F1 to F9, comparison finding 4, DAP finding 9).
10. Digital audio players get a section stating the verdict, which is that device sync is a local-binary job (DAP
    findings 1, 5, 7, 10).
11. The ladder carries magnitude bands; ceilings are set per plan (M13).
12. The device page and share management are web surfaces in one app; the framework question is open ruling 3 (M11).

Three findings went to Geoff rather than to a conductor ruling, ruled in "Rulings on the review round": the build-spec conflict (B1),
rung 8's tag I/O license path (F1), and the web framework (M11).

## Purpose

dubplate-server 2.0 subsumes Navidrome. After it, the estate is one binary a stranger can set up: intake, tagging, the
catalog, and the streaming server that the group's apps connect to. The primary driver is that single integrated system
(Geoff, 2026-09-07). Owning the listening experience and the single source of truth follow from it rather than
motivating it.

Navidrome's one missing feature is intake and tagging, which dubplate already owns. 2.0 adds the half Navidrome does
well, streaming to OpenSubsonic clients, on top of the catalog the pipeline already maintains.

## Decisions

Each decision names its alternatives and the reason the alternative lost, so a later reader does not reopen it without
new evidence.

### Client surface: OpenSubsonic API first

2.0 launches with the OpenSubsonic API and no first-party player. The API is what makes phones, cars, and desktop apps
work, and it is a specified contract the harness can test the way rung 3 tested the matcher. A web player follows on the
2.1 list, built on the same API, so the browser is one more client rather than a privileged one. Rejected: web player
first (the group listens in apps), both at launch (largest scope with no proportional gain), the Jellyfin API instead
of or beside OpenSubsonic (its Subsonic bridges JellySonic and jellysub are archived and last targeted Jellyfin 10.8,
and one strong beta client does not outweigh the maintained Subsonic client field; comparison finding 6, evidence dated
2026-09-07).

### Index: the catalog is the index

The store rung 6 builds, filled from the beets database at rung 16, is the only catalog. The streaming index is a
projection of it. No scanner admits music. A reconcile walk runs on a timer and reports drift (a file missing, a hash
that no longer matches) as doctor rows and an admin notice; it never admits or removes anything. The walk compares size
and mtime first and hashes only on suspicion, under a configured schedule and an I/O bound, because re-hashing the
library on a timer is the most expensive thing 2.0 could do (minor 3). Intake stays the one door. Consequence, accepted:
"point 2.0 at a folder and play" is not a feature; a pre-existing library enters through the pipeline, as rung 16's
migration already does for the box. Rejected: a scanner beside the catalog (two truths, the thing 2.0 exists to remove),
scanner only with the catalog as overlay (Navidrome's architecture with its friction).

### Identity: app passwords minted from the roster account

The roster stays the only account store. A user signs in by magic link (rung 5) and mints a named app password per
device from a device page, shown once and revocable alone. One stored secret serves both the legacy Subsonic
salted-token scheme and the OpenSubsonic API-key scheme. Rejected: one Subsonic password per user (a lost phone rotates
every device), API keys only (client support is uneven, and Navidrome implements neither `apiKeyAuthentication` nor
error codes 41 to 44, so every client in the roster uses the legacy path today).

### Transcoding: ffmpeg under a content-keyed cache

ffmpeg runs as a managed subprocess under `exec.CommandContext`. A content-addressed cache keyed on (track hash,
profile) sits between the transcoder and the client, filled on first play, never pre-warmed at launch, evicted
least-recently-played under a configured size cap. Geoff chose the cache on the numbers: a full cache of one profile
costs about a tenth of the library's disk, and the pipeline's content hash makes invalidation a delete on retag rather
than a heuristic. Rejected: no cache (repeat plays and seeks pay the encode every time), direct streaming only (lossless
on cellular is unusable).

### Per-user state at launch

Starred, playlists, play history, and the saved queue ship at launch. Play history is an append-only event table from
day one; play counts, recently played, and most played are views over it. Ratings and external scrobbling wait for 2.1.
Rejected: full Navidrome parity at launch (scrobbling brings credentials and rate limits into the launch), playlists
only (starred and play counts are cheap and missed on day one).

### Relation to the current ladder

The ladder finishes to launch with Navidrome as the reference driver. The driver seam stays, so 2.0's streamer arrives
as one more driver and the cutover is a configuration flip per estate. Rung 12's scope and rung 7's theming axis are
open ruling 1. Rejected: folding 2.0 into the ladder before launch (launch slides behind the largest subsystem on a
system that is not live), replacing rung 12 with the built-in streamer (loses the escape hatch).

### Evidence: a contract tier, two oracles, and a client roster

Evidence runs in three tiers. The contract tier validates every response against the OpenSubsonic OpenAPI schema and the
official Subsonic 1.16.1 XSD, as Ampache's `OpenSubsonicSpecConformanceTest.php` already does in CI; a response can
match an oracle and still violate the schema, and only this tier catches that. The oracle tier serves the same fixture
library from Navidrome and from gonic and diffs dubplate's responses against them, with gonic authoritative for
`getIndexes`, `getMusicDirectory`, and `getMusicFolders` because Navidrome synthesizes its folder tree from ID3 rather
than from real paths. The oracle image digest is pinned, the ledger records the oracle version per adjudication, and a
version bump is a deliberate pass that re-runs the ledger. The roster tier walks each slice in the group's apps.
Rejected: the roster alone (a client hides most of the API), the spec alone (silent where clients depend on de facto
behavior), a Navidrome-only oracle (it bakes in one server's choices, including two surfaces where it cannot grade
dubplate at all).

Two divergence classes are declared at the harness rung and adjudicated once as classes. The folder view diverges by
construction against Navidrome, which is why gonic grades it. The apiKey path has no oracle coverage at all and is
graded against the specification and the client roster. Search results compare as a set plus a top-N ordering assertion,
never field by field, because two FTS5 implementations with different tokenizers agree on membership and differ on order
constantly.

### Build shape: evidence first, then vertical slices

The study rung and the harness rung open the initiative and touch no production surface. The streamer then lands in
slices that each reach a real client end to end. Rejected: slices first with evidence retrofitted (every early slice is
re-verified once the harness exists, the plan-staleness defect class this repo has already paid for), a port of
Navidrome's architecture (license risk and the scanner model).

### Local mirroring: one writer per collection

The server stands on its own. Its intake, catalog, and streamer are the complete product, and the launch estate has no
local binary anywhere. When a person runs the future local dubplate binary (the organizer spec's post-launch two-binary
split), the local catalog is authoritative for the collections it owns and mirrors them to the server: a push of catalog
rows plus the content-addressed files the server lacks, keyed on hashes and idempotent. Collections the server owns are
written only by the server's intake. Every collection has exactly one writer, so no merge rule exists. Rejected:
bidirectional sync (a conflict model, the class of problem that sinks sync products), server as writer with the local
binary as a cache (inverts the local workflow Geoff described). Collection ownership is therefore a catalog attribute,
forward hook 8.

### License

dubplate is MIT (Geoff, 2026-09-07). Navidrome and gonic are GPLv3. The study rung reads them for ideas and for lessons,
positive and negative, and records what it learns in its own document (Geoff, 2026-09-07). Implementers build from that
record and from the OpenSubsonic contract, and never copy or transliterate GPL code. Melodee is MIT, so its code may be
reused as well as read. Permissive dependencies are ours to adopt. Apache-2.0 was considered and declined for dubplate's
own license: its patent grant and NOTICE clause buy nothing here and add friction for a stranger forking the estate.

## Architecture

### Catalog projection

The catalog has one row per admitted track: content hash, path, release and recording MBIDs, the tags as applied, gain
values, and the collection. A projector derives the OpenSubsonic entities (artist, album, song, genre, music folder)
into tables of their own inside the same SQLite file. The projector is a pure function of the catalog: a full rebuild
produces byte-identical tables under a determinism test, and every catalog event (admit, retag, move, remove) applies
its delta in the transaction that commits the event.

Entity ids are opaque surrogates the server mints once per entity. The release MBID and the content hash are indexed
alternate keys, never the id, so a retag updates an alternate key on an existing row and every starred row, playlist
entry, queue entry, and history event survives it untouched. The residual case is a retag that moves tracks between
albums, merging or splitting an entity; the projector records the merge and re-points per-user rows at the surviving id,
and the harness adjudicates a split as a divergence class. A catalog removal cascades to the per-user rows that named
the removed entity.

An album is a release MBID within a collection. An album with no release MBID is identified by the tuple (collection,
folded album artist, folded album title, content-hash set), which the pipeline already computes, because intake treats
the album as the verdict unit. A re-rip of the same release into the same collection is the same album with replaced
tracks, not a second album. The same release in two collections is two albums by design, and a client browsing all music
folders sees it twice under different ids; that is accepted and recorded in the ledger. Two pressings are two albums. A
compilation is whatever MusicBrainz says, with the album artist from the release credit. Multi-disc releases are one
album with disc numbers.

Collections map to Subsonic music folders, which yields multi-library as a consequence of the mapping. `musicFolderId`
is honored per endpoint rather than as a blanket filter, and the study rung lists the endpoints that take it. Clients
that browse by directory get a folder view derived from the paths the pipeline chose, served from a projection table
rather than a filesystem walk.

Every id on the wire is an opaque server-issued token and never an encoded path, and folder ids are minted rows in the
projection. `stream`, `download`, and `getCoverArt` resolve a path only by id lookup in the catalog, and the resolved
path is verified under the configured library root before any read. No client-supplied path reaches the filesystem.

### Art

Art has one origin, the pipeline. The admitted cover image is stored as a content-addressed blob with its hash on the
catalog row, and a plain `cover.jpg` is written beside the album in the layout, because several players read only
external images. Embedded pictures are extracted once at admit; the serve path never instantiates a tag reader per
request. The cover-art endpoint resizes on demand with `golang.org/x/image`, encodes WebP when the client accepts it and
JPEG otherwise, and writes each variant into the same content-keyed cache the transcoder uses under (art hash, size,
format). Variants are immutable, served with a one-year cache lifetime and an ETag. A missing cover returns a generated
placeholder, never a 404. `getCoverArt` accepts the dedicated coverArt id and, for legacy clients, an entity id under
the same lookup rule. Artist images are on the 2.1 list.

### Search

The search endpoint answers from an FTS5 index the projector maintains beside the entity tables, using the trigram
tokenizer over accent-folded text. Folding reuses the matcher's unidecode table through the folded sort key stored on
the catalog row. Ranking runs a prefix pass as a separate indexed lookup unioned ahead of the FTS5 rank, then the
matched artist's own albums. Results carry the same ids as browsing. `search3` accepts an empty query and returns
everything, which offline sync depends on, and paging survives a whole-library result. An edit-distance rerank is held
for 2.1 if trigram recall proves thin, on `matcher.StringDist` so the system folds text one way.

### Streaming and transcoding

The stream endpoint has two paths. The direct path serves the library file through the standard library's range-aware
file serving, so seeking and resume are the client's ordinary byte-range requests and the server holds no per-stream
state. The transcode path runs when the request names a format or bitrate cap or the device's profile sets one. Profiles
resolve in order from the request, the device, and the estate default.

Each profile is a named ffmpeg argument template held in configuration. A format selects a whole template by name from
the configured set and never substitutes a string. A bitrate is an integer validated against an allowlist, and a seek
offset is a validated numeric. Templates are argv slices with typed placeholders, never shell strings; no template may
carry `-i` or an output path, which the server supplies; configuration load rejects a template naming anything but the
declared placeholders. Any shipped MP3 profile writes gapless headers. ReplayGain is applied by the profile when the
profile asks for it, and never to the library file.

On a cache miss, ffmpeg starts under the request's context with a per-track timeout, and its output streams into the
cache entry while readers attach, so a second listener shares the encode. A Range request on a filling entry is served
from that entry once the requested offset is written. A non-zero ffmpeg exit or a cancelled context evicts the entry
rather than committing it, attached readers see the error, and the request returns 500, so a truncated encode never
serves as a complete track. On a hit the entry serves like a static file, ranges included. A disconnect kills ffmpeg
only when no other reader is attached. One knob bounds concurrent ffmpeg processes (default two); requests past it queue
in arrival order under a bounded wait, and a request past the cap returns error code 0 with a message rather than
holding the connection. The cache enforces its size cap, drops every variant of a track the catalog removes or retags,
and reports size, cap headroom, and hit rate as doctor rows. Art variants share the cache and the cap.

### Identity and app passwords

App passwords are long random strings the server generates. The stored secret is encrypted at rest with AES-GCM from the
standard library, under a per-row nonce and a domain-separation label, because the legacy salted-token scheme needs the
plaintext to verify and because Navidrome does not implement `apiKey`, which makes the legacy path the one every client
uses. Encryption protects an R2 snapshot and a stolen disk, not a box compromise, because the key sits in the box's
environment file. The key is minted through the estate's age store per CLAUDE.md's secrets rules, rows carry a key
version column, and rotation re-encrypts rows under the new version with a `_PREVIOUS` overlap. A device may instead be
minted modern-only, storing a hash rather than a recoverable secret, when the client supports the `apiKeyAuthentication`
extension; the roster measures how many devices take that path.

Every app password carries the user's valid-after epoch, so a roster-level revoke invalidates all of a person's devices
at once and offboarding needs no runbook step. Rung 5's identity layer exposes the revocation seam a credential store
subscribes to. The deprecated plaintext and hex password parameters are refused with error 42. Failed authentications
are rate-limited per account and per address under the build spec's settled client-address rule, using
`golang.org/x/time/rate` token buckets. Every mint, first use from a new address, and revoke is an event. The Subsonic
username is a roster handle: short, ASCII, unique, stable, distinct from the login email, and resolved to the durable
surrogate user id the per-user tables key on. Roles map from the roster (admin or member); the API's finer flags derive
from the role.

### Shares

Shares ship at launch: `createShare`, `updateShare`, `deleteShare`, and `getShares`. A share is scoped to one album or
one playlist and resolves only the entity ids it names. An expiry is required, enforced server-side. The token carries
at least 128 bits from `crypto/rand` and is compared in constant time. The public path is rate-limited per address, and
a share is revocable alone. A share cannot drive an unbounded transcode: it serves the direct path, or a configured
per-share encode budget.

### Per-user state

Four tables keyed by surrogate user id and entity id: starred (one row per user, entity, timestamp), playlists (a header
with a monotonic version plus ordered entries), the play queue (one row per user with ordered ids, position, and
client), and play history (append-only events: user, song, client, timestamp, event kind, and how reported). No row
references a path.

Playlists are mutated by `updatePlaylist`'s deltas applied inside one transaction against the stored order, never
against a client-supplied list, so two clients removing one track each do not lose a write. The header version detects a
stale write and names the conflict behavior. `createPlaylist` with a song list is the one whole-list write.

A play is recorded when `scrobble` arrives with `submission=true`, matching the oracle. `submission=false` is
now-playing only. A stream start is a separate event kind and never a play, because most clients buffer a whole track
ahead of playback and a byte threshold would measure prefetch. The "how reported" column records provenance so a later
pass can change the rule without losing history. The play queue's edges are specified: a call with no `id` clears the
saved queue, `current` is required unless `id` is empty, an absent `position` means zero, and the response carries
`changed` and `changedBy`.

### API surface

One handler set on the standard library mux under the owned prefix, answering XML and JSON from the same typed response
structs. The Subsonic path is a sub-prefix of the one prefix the build spec's rung 1 rule owns, and a client's
configured base URL points at that sub-prefix, so the one-owned-prefix rule holds without amendment. The ingress strip
set still applies to it.

Every `subsonic-response` carries `openSubsonic`, `serverVersion`, and `type` beside `status` and `version`.
`getOpenSubsonicExtensions` is reachable unauthenticated, the surface's one carve-out.

Launch implements the endpoints the client roster exercises: system (`ping`, `getLicense`, `getOpenSubsonicExtensions`,
`getScanStatus`), browsing in both ID3 and folder views, album lists, search, media (stream, download, cover art, lyrics
stubbed), annotation (star, unstar, scrobble), playlists, play queue, shares, and read-only user. `download` returns
album, artist, or playlist zips in the pipeline's own layout with sanitized names, and the playlist zip carries an
`.m3u8` with relative entries, matching Navidrome.

Every other endpoint returns error code 0 inside a well-formed `subsonic-response` with a message naming the endpoint,
never a bare 404. The exception is the set of endpoints clients require to answer with an empty success; the study rung
produces that set, measured against the client roster. The launch extension subset is also a study-rung deliverable,
measured against the ten extensions the spec repo carries today. Every endpoint is one function with one test table.

### Web surface

The device page and share management are pages in the same web app as rung 5's login. They build to static assets
embedded through `go:embed` and served by the Go binary, with Node at build time only and nothing Node on the box. The
flow is magic link, then session, then the device page under the owned prefix. The framework question, standalone Svelte
5 as the repo rule stands or SvelteKit in static mode, is open ruling 3 and is not decided here.

Revision 2.2 (Geoff, 2026-09-12): the web app also carries an admin settings surface (settings, per-member overrides,
collection management) over the rung 14 admin API, the same calls the TUI's admin section makes. The configuration homes
spec names the homes; this spec owns the page.

### Digital audio players

Device sync is a local-binary job, not a server feature. Every capability a sync needs is present on a workstation and
absent on the VPS, which reaches the group through Cloudflare. 2.0 grows no device verb and no MTP. The largest real
DAP class is a phone, which the clients' own offline caches already serve. The server's contribution is `download`,
specified above. A DAP export profile in the local binary is on the 2.1 list, gated on a real device in the group.

### Cutover and operations

The streamer is one more driver behind rung 12's seam, selected by one configuration value. Its first weeks run as a
shadow: Navidrome stays the group's server, 2.0 serves the same library on a second hostname, the harness runs on a
timer against the live box, and the roster walks it. The flip is a configuration change and a deploy in one owner
sitting. Navidrome's container stays defined but stopped for a named period and is removed at a dated close-out. Starred
lists and playlists built in Navidrome are imported once through an export the harness reads; differences are reported,
never silently merged.

After the flip the box runs the dubplate binary, ffmpeg, cloudflared, and the timers. Provisioning adds ffmpeg and the
cache directory. The streamer's state lives under the existing state directory, so backup covers it. Doctor gains rows
for ffmpeg's version, cache size against cap, index rebuild parity, and the drift count. Configuration adds an
integration class in rung 4's shape: enabled, cache cap, concurrent encodes, the profile set, and the app-password key,
each required when enabled.

## The ladder

Magnitude bands are relative and carry no time units, matching the build spec's ladder. A rung's ceiling and checkpoint
interval are set in its own plan, never here.

1. Study (M). A written record of what Navidrome and gonic do where the spec is silent, cited to a version and a harness
   test, with the ideas and lessons their code teaches recorded in dubplate's own words. Ships the pre-seeded divergence
   ledger, the empty-success endpoint set, the launch extension subset, the per-endpoint oracle assignment, and the
   prior-art table's open cells.
2. Harness (M). The fixture library, the pinned Navidrome and gonic containers, the schema and XSD validators, the
   recorded session, the differ, the divergence ledger, and the client roster's scripted walk.
3. Catalog projection and browsing (L). Splits into passes at its own planning; it carries the projector, both browse
   views, and the id space.
4. Art and search (M).
5. Streaming and transcoding (L). Splits at its own planning; it carries the cache, the concurrency model, and range
   behavior on both paths.
6. Identity, app passwords, and shares (M).
7. Per-user state (M).
8. Shadow period, flip, and close-out (M). Retires Navidrome and records the post-mortem.

Each slice is a pass with its own plan, ceiling, oracle diff, and roster walk, under the conducting loop in CLAUDE.md.
Rungs 1 and 2 touch no production surface and may run as a document-and-corpus lane beside the current ladder's later
rungs.

## Forward hooks the current build owes

Each hook names the rung whose plan weighs it. That plan states per task whether it lands a hook or is neutral
(CLAUDE.md, "Build for 2.0").

1. Content hash on the catalog row, rung 6 schema, computed at rung 7's gate on the uploaded bytes. The cache key, the
   alternate key, and the drift check all read it. Amended 2026-09-12 (configuration homes spec): the row carries two
   keys. The audio key is the decoded-audio MD5, computed by decode at the gate and never read from the header, and it
   identifies the same audio across collections, drives the overlap report and duplicate refusal, and survives a
   retag. The file hash is computed streaming over the file bytes at the same gate; it changes on retag and drives the
   drift check, the transcode-cache key, and backup change detection. Album sameness is the ordered set of track audio
   keys with the track count, never a per-track match. The formats rung decides the audio key per later format.
2. Folded sort key on the catalog row, rung 6 schema, computed at rung 9's tagging. The projector and the search index
   agree on folding without re-deriving it. The key prefers the sort tags of hook 13.
3. Art blob recorded at admit, rung 6 schema and rung 9 admission. The admitted cover is stored content-addressed with
   its hash on the row, and embedded pictures are extracted once, never at serve time.
4. Play history as timestamped events, rung 6 schema. Any event-shaped table carries a timestamp and an event kind,
   never a bare counter.
5. A stable identity, rung 5 and rung 6. Rung 6 mints a durable surrogate user id that per-user rows key on, and rung 5
   resolves a roster identity to it. The roster's `streamer_username` becomes required, unique, and stable rather than
   optional, and it is the renameable handle the surrogate id protects against.
6. The driver seam stays universal, rung 12. Address, proxy, and provisioning choreography are per-driver capabilities a
   driver may decline to declare, because an in-process streamer has no address and needs no proxy. Scan trigger and
   scan status must be satisfiable by a driver that owns the catalog and never scans. Capability reporting is the
   mechanism that makes both true, and rung 12 builds no more of the Navidrome driver than SSO and roster sync.
7. Schema migrations, rung 6. 2.0 guarantees the schema evolves across versions, so the migration mechanism is chosen
   with that in view.
8. Collection ownership, rung 6. Each collection records its writer (this server, or a named local binary), so mirroring
   has one writer per collection from the first schema.
9. A transaction model across catalog events, rung 6. One writer, WAL with a `busy_timeout`, and an ops seam a later
   projector can join, so a projector delta commits in the same transaction as its event. Retrofitting this is a seam
   rewrite.
10. Live-file backup consistency, rung 6. A byte-for-byte copy of a live WAL database yields a torn snapshot, so the
    backup job uses the SQLite backup API or checkpoints before it copies. This is worth fixing regardless of 2.0.
11. ReplayGain and R128 gain on the catalog row, rung 6 schema, computed at rung 9. Track and album gain and peak, plus
    the Opus R128 fields. Discovered at 2.0 instead, it means re-reading every file in the library.
12. A portable layout, rung 7. Filenames avoid the FAT and exFAT forbidden set, hold every path component under 255
    bytes, and carry no trailing dot or space and no reserved DOS name. The unsanitized title stays in the tag. The same
    rung writes a plain `cover.jpg` beside the album in addition to the blob.
13. Sort tags written at admit, rung 8 and rung 9. `ALBUMARTISTSORT` and `ARTISTSORT` with their ID3 equivalents are
    written once or never; retrofitting means re-tagging the library.
14. An ID3 version ruling, rung 8. Record whether the library writes ID3v2.3 or v2.4, so the question is not answered by
    reading files later.
15. Configuration homes, rung 6 and rung 14. The store carries a settings table (source, apply mode, sensitive flag,
    audit table), per-member overrides and preferences tables, a device-credential table, two quota counters, a
    move-intent journal, and on the collection row `owner` (non-null surrogate id), `visibility`, a server-allocated
    disjoint `root_path`, a writer epoch, a mirror generation, a last-mirror time, and a duplicate-collapse flag beside
    `writer`, with share rows reserved beside visibility and foreign keys enabled in the DSN. The rung 14 admin API
    exposes settings and collections on two mounts (cookie and bearer) so the web app and the TUI share one surface.
    Hook 6's capability report also declares a driver's shape constraints (subtree collections, art filenames read,
    tags read) and per-user folder access; private is unsettable without it. Design:
    `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`.

## Prior art

Every row needs a license cell citing the repository's LICENSE file by URL, because revision 1's strings were filled
from memory and two were wrong. Cells the reviews established are filled below; the rest are the study rung's
deliverable, scored on CLAUDE.md's four criteria with evidence per cell. Maintenance keeps its criterion but loses its
bus-factor weight: bus factor is 1 across this entire field, Navidrome, gonic, LMS, Koel, Black Candy, and Melodee
alike, so that signal discriminates nothing and adoption evidence does the real work.

| Entry | Role | License (URL) | Quality | Maintenance | Adoption | Verdict |
|---|---|---|---|---|---|---|
| OpenSubsonic specification and extension list | The contract | Apache-2.0 ([LICENSE](https://raw.githubusercontent.com/opensubsonic/open-subsonic-api/main/LICENSE)) | n/a | Live, 10 extensions | The de facto contract | Adopt. Read it; never vendor it, because the NOTICE clause would attach. |
| OpenSubsonic OpenAPI schema and Subsonic 1.16.1 XSD | The contract tier | study rung | n/a | Schema self-declares work ongoing | Ampache validates against both in CI | Adopt as harness inputs. |
| Navidrome | Primary oracle; ideas and lessons | GPL-3.0, URL from study rung | n/a | Very active, near-monthly | 23,396 stars; every client names it | Adopt as oracle, pinned by digest. Read for ideas; never transliterate. |
| gonic | Folder-view oracle; ideas and lessons | GPL-3.0, URL from study rung | n/a | Active | 2,538 stars; the second server clients name | Adopt as the oracle for `getIndexes`, `getMusicDirectory`, `getMusicFolders`. Its author also writes taglib, so both share one bus factor. |
| Melodee | Readable reference implementation of the ingest-pipeline-plus-OpenSubsonic shape | MIT, URL from study rung | study rung | Active, 1,478 commits since 2024-10 | 65 stars; publishes its own compatibility matrix | Adopt as readable and reusable prior art. Its Postgres and Blazor stack rules it out as an oracle. |
| `go.senan.xyz/taglib` | Tag read and write | LGPL-2.1, whole module ([LICENSE](https://github.com/sentriz/go-taglib/blob/master/LICENSE)) | Mostly good; wazero cache hardcoded to `os.TempDir()`, host module panics with no recover | v0.14.0, 2026-07-24; depends on wazero at an untagged pseudo-version | 11 importers; Navidrome runs its own fork | Open ruling 2. The "MIT wrapper" claim was false and the MPL-1.1 escape does not exist. |
| `tetratelabs/wazero` | Runs the WebAssembly | Apache-2.0, URL from study rung | High | v1.12.0, 2026-05-29 | 714 importers | Adopt only as taglib's transitive dependency, never directly. |
| `modernc.org/sqlite` | cgo-free SQLite | BSD-3, URL from study rung | Transpiled C; the house bar applies only to the `database/sql` layer | v1.58.0, 2026-09-01, near-weekly | 3,518 importers | Adopt on a condition: rung 6 proves FTS5's trigram tokenizer and WAL under concurrent readers by a test in the gate. |
| `ncruces/go-sqlite3` | The named fallback | MIT, URL from study rung | Well-documented VFS design | Active | 96 importers; gonic uses it | Record as the fallback if the trigram or WAL check fails. |
| `mewkiz/flac` | Pure-Go FLAC decode; the intake integrity check | Unlicense, URL from study rung | Idiomatic package split | v1.0.14, 2026-08-19 | 96 importers | Adopt for rung 7. It supplies audio-frame decode, which the repo's own FLAC reader deliberately lacks. |
| Picard tag-mapping tables | Which tag names each container uses | CC0 1.0, URL from study rung | n/a | Versioned with Picard | The canonical mapping | Adopt, cited to a pinned doc version. |
| `golang.org/x/image`, `gen2brain/webp` | Art resize and encode | BSD-3 and MIT, URLs from study rung | Exemplar; webp is generated but the seam is honest | x/image v0.45.0; webp v0.6.4 | 1,948 and 18 importers | Adopt. The resize filter is a stated choice with a measurement, because the cache pays the encode once. |
| `golang.org/x/time/rate` | Per-account and per-address auth rate limits | BSD-3, URL from study rung | Exemplar | Go team | 14,348 importers | Adopt. The repo's MusicBrainz limiter solves a different problem. |
| `crypto/aes` and `cipher.NewGCM` | App-password encryption at rest | stdlib | n/a | n/a | n/a | Adopt the stdlib; no KDF library is needed. |
| Schema migrations | Rung 6's mechanism | stdlib | n/a | n/a | n/a | Hand-roll with `embed.FS` and `PRAGMA user_version`. `pressly/goose` is declined: it direct-requires ten database drivers for a Go-function migration mechanism no hook asks for. |
| Stream-through transcode cache | 2.0 rung 5 | stdlib | n/a | n/a | n/a | Hand-roll. `djherbis/fscache` is declined: 7 importers, no tag since 2021, and the partial-write semantics the design needs exist only in an untagged commit landed by its one serious user. |
| `internal/atomicfile` | Streaming atomic writes for blobs and cache entries | in-repo | n/a | n/a | n/a | Extend it with a streaming write rather than inventing a second temp-file convention. |
| M3U read and write; Subsonic XML and JSON dual encoding | Cutover import and the response layer | stdlib | n/a | n/a | n/a | Hand-roll both. The study rung enumerates the encoding quirks (attribute-heavy XML, the wrapper, singleton-versus-array in JSON, JSONP) so the harness tests them. |
| Chromaprint via `fpcalc` | Acoustic fingerprinting for untagged uploads | MIT source, but the distributed library is LGPL-2.1+ | n/a | Stable upstream | The AcoustID standard | Adopt for 2.1 as a subprocess only. Never link the library, and pin the FFT backend so a GPL build never lands. |
| Symfonium compatibility notes | Where servers diverge in practice | Documentation | n/a | Live wiki and forum | The best public record of client-side divergence | Adopt as a study input. |
| beets `convert`; MusicBee device sync | The export-profile shape for the local binary | Documentation | n/a | Current | Behavior references | Adopt as prior art for the future local binary, never for the server. |

Declined and recorded: `xrash/smetrics`, because `internal/engine/matcher/strdist.go` and the pinned unidecode table
already cover it, and a second folding path is the defect the folded sort key exists to prevent. Navidrome's remaining
dependencies (chi, jwtauth, viper, notify, cron, extism, mpvipc, squirrel, prometheus, pocketbase/dbx, lestrrat-go/jwx,
santhosh-tekuri/jsonschema) are architecture 2.0 chose against or already covers.

## The 2.1 list

Scrobbling (ListenBrainz, Last.fm); ratings; artist images; lyrics beyond the launch stub; transcode pre-warming per
profile; the first-party web player; metrics; an edit-distance search rerank built on `matcher.StringDist`; acoustic
fingerprinting for untagged uploads; a DAP export profile in the local binary (path template, sanitizer, transcode rule,
art writer, playlist writer), gated on a real device in the group; jukebox mode, if ever.

## Rulings on the review round

The three items the reviews routed to the owner were ruled by Geoff on 2026-09-07, all as recommended.

1. The build-spec conflict (B1). Ruled: no theme matches Navidrome's or Jellyfin's any longer. The build spec carries a
   dated erratum (its header): ladder entry 7's structural theme axis targets dubplate's own design only, the
   streamer-matched themes are not authored, and the estate config is `match dubplate, palette gruvbox`; ladder entry
   12 is trimmed to the driver seam, SSO, roster sync, and the proxy-hygiene ingress. Landed at commit `8c54fff`.
2. Rung 8's tag I/O license path (F1). Ruled: option (b), pure-Go per-format readers and writers under permissive
   licenses, extending the repo's own FLAC metadata reader, scoring `dhowden/tag` for reading and `bogem/id3v2` for MP3
   writing, with an MP4 atom writer still to be scored; FLAC, MP3, and MP4/AAC at launch, an unusual container rejected
   at intake rather than mistagged, and each further container its own small pass. `go.senan.xyz/taglib` (LGPL-2.1) is
   not adopted.
3. The web framework (M11). Ruled: SvelteKit in static mode (adapter-static, client-side rendering only, server-only
   files forbidden and gate-checked), adopted before rung 5's plan is authored, with clean, idiomatic SvelteKit as the
   bar as in the cairn family. CLAUDE.md's web rule is updated at commit `8c54fff`.

## Open items

These do not block the study rung.

- The client roster's exact membership. The live candidates are Symfonium, Feishin, Supersonic, Substreamer, and Tempus;
  which the group installs decides the roster, answered at the harness rung. Symfonium is proprietary and paid, so the
  group must own a copy.
- Whether folder browsing ships at launch at all. It doubles the browse surface and the id space, modern clients use
  ID3, and only gonic can grade it. The study rung brings the evidence.
- The app-password key's exact name in the age store. The identity slice mints it.
- The shadow period's length and the Navidrome retention window. Owner rulings at the cutover slice.
- Whether the study rung runs beside the current ladder or after launch. A document lane is permitted by the standing
  grant; the conductor decides on the lane registry's state.
