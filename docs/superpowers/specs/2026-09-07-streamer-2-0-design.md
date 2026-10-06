# dubplate-server 2.0: the built-in streamer

**Re-keyed to the 0.8 release (Geoff, 2026-09-25 ruling on version numbering).** Everywhere
this document says "2.0" it means 0.8. The title and body are left as written.

Status: **Ratified 2026-09-07.** Revision 3, 2026-10-05 (the built-in streamer is the group's first launch, so every passage that assumed an incumbent, a shadow, or a flip is removed or rewritten; listed under "Revision 3, 2026-10-05", with the rulings recorded under "Rulings of revision 3"). Revision 2.5, 2026-09-25 (errata from the ratified group-scale principles spec, the streamer prior-art survey, and the give-back decision; listed under "Revision 2.5, 2026-09-25"). Revision 2.4, 2026-09-24 (Geoff's ruling: five hosting and setup features join the 2.0 target, recorded under "Hosting and setup at launch", "Hosting and setup", the widened Navidrome import, forward hooks 18 and 19, and slice 8). Revision 2.3, 2026-09-24 (Geoff's ruling: four group features join the 2.0 target, recorded under "Group features at launch", "Group features", forward hooks 16 and 17, slice 7, and three open items). Revision 2.2, 2026-09-12 (forward hooks 1 and 15 and the web surface amended from the configuration homes spec, `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`). Revision 2.1, 2026-09-07 (the three owner rulings recorded in "Rulings on the review round"). Revision 2, 2026-09-07. Every decision below was taken with
Geoff in the 2026-09-07 brainstorm and none is reopened here. ROADMAP entry: `streamer-2-0`. It is rung 12 of the ladder: its eight slices are rung 12's
passes, and the group's first launch (0.8.0) runs on it.

This document governs the initiative's purview, its architecture, and the forward hooks the current build owes it. Where
it names a current rung (5, 6, 7, 8, 9, 12, 16), it asks that rung's plan to weigh a hook, never to build a 2.0 feature
early. One sentence in revision 1 trimmed rung 12 while the document claimed to amend nothing; that conflict is open
ruling 1 rather than a silent amendment.

## Revision 3, 2026-10-05

Revision 3 follows the re-plan in `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md` (revision 2.1).
On 2026-10-05 Geoff retired the incumbent stack, so the group listens through dubplate's own streamer from the first
day and nothing remains to shadow, flip to, or fall back on. This revision edits in place, and the earlier revision
sections stay as written. Each item names the reason and the section that carries the change.

1. The shadow period, the flip, and the Navidrome retention window are removed. The reason: no incumbent serves the
   group, so there is nothing to compare against live or to flip from (rulings 1 and 4). "Cutover and operations".
2. The `shareRole` doctor row and the escape-hatch text are removed, because no Navidrome serves as an escape hatch.
   "Cutover and operations".
3. The arm64 walk "before the flip" is removed. The arm64 build stays. "Hosting and setup".
4. Slice 9 (shadow period, flip, and close-out) is deleted, and slices 1 to 8 are stated as rung 12's passes. The
   reason: there is nothing to shadow or flip. "The ladder".
5. "Relation to the current ladder" and the header's "sequenced after the greenfield ladder's launch" are rewritten,
   and the rejected alternatives are marked superseded. The reason is ruling 1: first launch is the streamer, so no
   ladder finishes with Navidrome as the reference driver.
6. "Cutover and operations" is rewritten to the streamer's configuration properties: always on, no enable flag, no
   integration class, no driver, no cloudflared, tunables as registered thresholds with defaults, secrets on the auth
   secrets' pattern, and a per-slice delta to the operator-supplied set. The reason is ruling 7: a stranger's install
   stays easy.
7. Slice 8 gains four import criteria, the unit's `CAP_NET_BIND_SERVICE` in TLS mode, the certificate leaf, and Owner
   ruling 3's walk of the install path. The import's rationale is rewritten to the first-launch import (ruling 2).
   "Hosting and setup" and "The ladder".
8. Slice 6 records the bearer bypass's two contract sentences and that the bypass is path-keyed. "The ladder".
9. Hook 6 and hook 15's capability clause are retired in place, and capability reporting is dropped unless a slice
   names a consumer. Hook 19 is re-homed to slice 8, and hook 5's field is `username` (rulings 3 and 4). Hook numbers
   never change, because other documents cite them. "Forward hooks the current build owes".
10. The open items on the shadow period's length and the Navidrome retention window are marked moot, and the open item
    on whether the study rung runs beside the current ladder is marked answered: slice 1 runs now as a document lane.
    The open item on the client roster's membership points at the roster audit record. No other open item changes.
11. Two passages outside the named sections that assumed a cutover are rewritten to the first-launch import (ruling
    2): "Hosting and setup" and the prior-art table's import cell.
12. The pinned Navidrome and gonic containers stay as workstation-only behavior oracles, under the GPL behavior-only
    rule ("Evidence").
13. Slice 1 gains a study question on Melodee's intake pipeline ("The ladder").

## Rulings of revision 3

Rulings 1 to 7 are Geoff's, 2026-10-05, from the re-plan spec. Owner rulings 1 to 3 were ruled yes the same day.

1. First launch is 0.8.0, on the built-in streamer. No 0.x pre-release exists, and 0.9 and 1.0 keep their meaning.
2. `dubplate import navidrome` stays in 0.8. It serves a stranger switching from Navidrome, and the group's first
   launch runs it once against the Navidrome database in the retirement archive, so members keep their stars,
   playlists, and play counts.
3. The roster's streamer username becomes `username` everywhere: column, JSON key, flag, Go identifiers, wire codes,
   and copy.
4. dubplate integrates with no external streamer. The driver seam, hook 6, and the Navidrome and Jellyfin targets
   retire.
5. Rung 12 is the built-in streamer, and its first two slices start now: slice 1 as a document lane, slice 2 as a code
   lane in its own worktree.
6. The architecture stays clean, tidy, and beautiful through every change.
7. dubplate stays extremely easy to install and configure for a new operator, and every change is weighed against it.

Owner rulings, all yes:

1. The shadow program is deleted, superseding the build spec's ruling 1 adopter clause.
2. The forward-auth front retires along with the external streamers.
3. Slice 8 closes on a walk of the install path on a fresh throwaway host, with blocking defects.

## Revision 2.5, 2026-09-25

Revision 2.5 lands errata from three sources. Each item names its source, and the sections below carry the change.

- P is `docs/superpowers/specs/2026-09-25-group-scale-principles-design.md` (ratified at `1aa40f4`, rulings 1 to 12
  decided), through the owed-errata list in `docs/superpowers/research/2026-09-25-group-scale-principles-fold.md` as
  revised by its "Second fold" and "Bounded shares" sections. Conflict numbers below are that spec's "Conflicts with
  existing specs".
- S is `docs/superpowers/research/2026-09-25-streamer-prior-art-survey.md`, "Findings that change the spec".
- G is the give-back decision, ROADMAP `give-back` (Geoff, 2026-09-25).

1. Shares keep their four endpoints under six bounds: one item, a lossy stream, clamped expiry, a listener pause, an
   unindexed token, and server-side clamping (P conflict 1; rulings 8 to 12). "Shares" and hook 15.
2. `download` accepts a song or an album; every other id kind answers error 50 (P conflict 2, row A5). "API surface"
   and "Digital audio players".
3. Slice 6 gains the credential cap, the admin credential list, and a one-step admin revoke. Slice 5 gains the
   served-volume meter, the admin notice by email, and the volume view (P conflict 3; rulings 4 and 5). "Identity and
   app passwords", "Served volume", "Web surface", and the ladder.
4. A new forward hook puts the credential and the bytes on every serve event (P conflict 4). Hook 20.
5. Hook 17 and "Added by" follow the principles spec's channel table, and the source answer joins as hook 21 (P
   conflict 5). "Group features" and the hook list.
6. `getAlbum` and `getAlbumInfo2` fill `musicBrainzId`, and `notes` carries the source line (P conflict 6). "API
   surface" and "Group features".
7. While Navidrome is the escape hatch, a `doctor` row reads its effective sharing value (P row A13). "Cutover and
   operations". Superseded by revision 3: no Navidrome serves, so the row is gone.
8. XML is validated either in a legacy form against the official XSD or against a derived XSD; the choice is open (S
   finding 1). "Evidence".
9. The OpenSubsonic `transcoding` extension against the profile order is open (S finding 2). "Streaming and
   transcoding".
10. A Range request on a filling transcode is open: a declared divergence or a redesign (S finding 4). "Streaming and
    transcoding".
11. The API answers CORS for browser clients, preflight included (S finding 5). "API surface".
12. An AAC profile streams only as fragmented MP4 or ADTS (S finding 6). "Streaming and transcoding".
13. The prefix lookup is required for queries shorter than three characters (S finding 9). "Search".
14. The word "collection" may clash with OpenSubsonic PR #237 (S finding 8). "Open items".
15. Stale prior-art pins are corrected (S finding 7). "Evidence" and "Prior art".
16. The OpenSubsonic protocol layer is built as a separable package tree with no dubplate imports, ships hand-written
    wire types validated by the pinned Apache-2.0 OpenAPI file, and publishes as its own module after 0.8 (G). "The
    protocol layer is a separable module".
17. `dubplate import navidrome` preflights the member cap and writes nothing past it (P row A1). "Hosting and setup".

S finding 3 confirms the identity reasoning in "Identity: app passwords minted from the roster account"; no change.

Items that need a design decision the sources don't make are marked "open, owed to the study rung" where they sit, and
are gathered under "Open items".

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

### Group features at launch

Ruled by Geoff, 2026-09-24. Four features join the 2.0 target because they give a group a reason to prefer dubplate
over Navidrome. Navidrome reads a folder it did not build and knows nothing of who built it. dubplate owns the pipeline
and the roster, so it can do both.

1. Report a problem. A member flags an admitted track or album from the listening side, and the report joins the
   decision queue beside intake items.
2. Added by and album notes. Each admitted album records the member who added it, and that member may write a short
   note on why.
3. Collaborative playlists. A playlist's owner names members who may edit it.
4. Smart playlists. A rule-based playlist evaluated at read time. Navidrome ships these as `.nsp` files, so a switcher
   expects them.

Not targeted for 2.0: the library health report, requests, the weekly new-arrivals email, and group
listening views. Each builds on data 2.0 already records, so none needs a hook now. They are catalogued in
`IDEAS.md`.

### Hosting and setup at launch

Ruled by Geoff, 2026-09-24. Five features join the 2.0 target because they make dubplate easier to host than
Navidrome, which starts from a music folder but leaves the operator the reverse proxy, the certificates, the upgrade,
and the move off another server. Each is a modest build on seams the ladder already has, and pulling them forward from
the packaging initiative means 2.0 ships them to the first stranger rather than after.

1. HTTPS without Cloudflare. The server obtains and renews its own certificate.
2. A general Navidrome import. Anyone switching brings their members, stars, playlists, and play history.
3. arm64 builds, for Raspberry Pi and NAS hosts.
4. `dubplate init`. A setup command writes the configuration, mints the secrets, and proves mail works.
5. `dubplate upgrade`. An upgrade verifies the release, migrates under the pre-migration snapshot, and rolls back on
   a failed health check.

### Relation to the current ladder

Revision 3 (ruling 1). The built-in streamer is rung 12 and the group's first launch, 0.8.0. The ladder reaches launch
with dubplate's own streamer and no external one behind a driver seam, so there is no incumbent to shadow, flip to, or
fall back on. The streamer's eight slices are rung 12's passes, and slices 3 onward start after rung 9 lands the
catalog.

Rejected, and superseded by ruling 1: finishing the ladder to launch with Navidrome as the reference driver and
arriving at the streamer as one more driver (the group no longer runs Navidrome, so the driver would be code built to
be deleted), folding 2.0 into the ladder before launch (launch is the streamer, so no second launch event
exists), and replacing rung 12 with the built-in streamer while keeping Navidrome as a fallback (the fallback retired
with the incumbent).

### Evidence: a contract tier, two oracles, and a client roster

Evidence runs in three tiers. The contract tier validates every JSON response against the OpenSubsonic OpenAPI schema,
as Ampache's `tests/Module/Api/SubsonicSpecConformanceTest.php` already does in CI; a response can match an oracle and
still violate the schema, and only this tier catches that. The prior-art survey proved the JSON half in pure Go: the
bundled OpenAPI file loads in `getkin/kin-openapi` and rejects a malformed `getAlbum` response. The harness pins that
bundled build by checksum, as Ampache does.

Revision 2.5 (S finding 1): revision 2 also validated every response against the official Subsonic 1.16.1 XSD, and
that can't hold. The OpenAPI file covers JSON only, and the
official Subsonic 1.16.1 XSD rejects the `openSubsonic`, `type`, and `serverVersion` attributes that every
`subsonic-response` carries (see "API surface"). The harness validates XML one of two ways: a legacy form that omits
the OpenSubsonic attributes, validated against the official XSD, or every response against a derived XSD that admits
them. Either way the choice is recorded as a declared divergence. Which one is open, owed to the study rung.

Revision 3: the pinned Navidrome and gonic containers stay as workstation-only behavior oracles in the harness, never
on the box, under the GPL behavior-only rule in "License".

The oracle tier serves the same fixture
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

### The protocol layer is a separable module

Revision 2.5 (Geoff, 2026-09-25, ROADMAP `give-back`). The OpenSubsonic protocol layer fills a gap in the Go ecosystem.
The prior-art survey found no conformance suite anywhere and no maintained, permissively licensed Go Subsonic client.
The layer is therefore authored to ship as its own permissively licensed module. It holds the wire types, a
request-parsing handler over a backend interface, a client, and a conformance checker. The harness rung's client and
contract-tier validator are that client and that checker. This revision makes the harness rung's client the module's
client, and its contract-tier validator the module's conformance checker; the checker then carries `getkin/kin-openapi`
as a module dependency, where the survey adopted it for tests only, and the harness rung's plan weighs that.
dubplate's streamer is one implementation of the backend interface, and the catalog projection, the roster, the
meter, and every other dubplate concern sit behind it.

- **Built under `internal/` now.** The layer lives in its own package tree under `internal/` for the whole of 0.8, so
  no outside caller depends on it before a release proves it.
- **A gate-enforced import boundary.** `go list -deps` over the layer's packages shows no `github.com/glw907/dubplate/`
  path. The slice that first creates the package tree adds this check to the gate as a leg that fails, never as a
  warning.
- **Extracted after 0.8.** The layer moves to its own module after 0.8 ships, and no earlier.
- **A clean-room record.** Every behavior the layer implements cites the OpenSubsonic specification, its Apache-2.0
  documentation, or an observation recorded by the harness, never GPL source. The study rung's record is the
  citation source, and "License" below still governs.

Whether the wire types are generated from the OpenAPI file or written by hand is no longer open. Conductor reading
(2026-09-25, reversible), recorded in ROADMAP's give-back entry: the wire types are hand-written, and the pinned
Apache-2.0 OpenAPI file validates conformance in the harness. The evidence: the survey tried `oapi-codegen` v2, usable
in the harness but unfit as production types since it has pointer fields, no `xml` tags, and a `json.RawMessage` union
per response, and tried `ogen`, which failed on the full spec; this spec's prior-art table already hand-rolls the XML
and JSON dual encoding.

The second is the module's license. ROADMAP requires a permissive one. Whether it is MIT or Apache-2.0, and what
notice it carries from the Apache-2.0 OpenAPI file it validates against, is open.

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

Revision 2.5 (S finding 9): the prefix lookup is required for queries shorter than three characters. A trigram
`MATCH` needs at least three characters, and the survey's two-character query returned no rows on `modernc.org/sqlite`
v1.59.0, so a query shorter than three characters is answered by the prefix lookup alone.

### Streaming and transcoding

The stream endpoint has two paths. The direct path serves the library file through the standard library's range-aware
file serving, so seeking and resume are the client's ordinary byte-range requests and the server holds no per-stream
state. The transcode path runs when the request names a format or bitrate cap or the device's profile sets one. Profiles
resolve in order from the request, the device, and the estate default.

Revision 2.5 (S finding 2): OpenSubsonic added a `transcoding` extension in December 2025 (`getTranscodeDecision` and
`getTranscodeStream`). Navidrome implemented it on 2026-03-09, and Feishin calls it. A third-party issue report says
Symfonium does too; the survey didn't verify that. Under it the client
posts a capability document (`ClientInfo`), and the profile order above has no place for one. Whether 0.8 answers the
extension, and if it does, how a posted capability document interacts with the device profile, is open, owed to the
study rung. The answer is recorded in the launch extension subset.

Each profile is a named ffmpeg argument template held in configuration. A format selects a whole template by name from
the configured set and never substitutes a string. A bitrate is an integer validated against an allowlist, and a seek
offset is a validated numeric. Templates are argv slices with typed placeholders, never shell strings; no template may
carry `-i` or an output path, which the server supplies; configuration load rejects a template naming anything but the
declared placeholders. Any shipped MP3 profile writes gapless headers. ReplayGain is applied by the profile when the
profile asks for it, and never to the library file.

Revision 2.5 (S finding 6): stream-through constrains the container. MP4 with `-movflags +faststart` works only on a
finished file, which is why Koel encodes a whole AAC track before its first byte. MP3 and Ogg/Opus stream through as
they encode. A shipped AAC profile therefore writes fragmented MP4 or ADTS, and never a `faststart` MP4. Which of the
two is open, owed to the study rung, measured against what the client roster plays.

On a cache miss, ffmpeg starts under the request's context with a per-track timeout, and its output streams into the
cache entry while readers attach, so a second listener shares the encode. A Range request on a filling entry is served
from that entry once the requested offset is written.

Revision 2.5 (S finding 4): this sentence diverges from the oracle and can't rest on the standard library. Navidrome
uses `http.ServeContent` only for a complete entry, answers a filling one with `Accept-Ranges: none`, and sends an
estimated `Content-Length` only when the client passes `estimateContentLength=true`. `http.ServeContent` can't serve
a filling entry, because it "uses a seek to the end of the content to determine its size", and a range answer on a
filling entry needs a `Content-Range` with an unknown total. Whether the streaming slice keeps this behavior as a
declared divergence class, or matches the oracle and lets seeking during a transcode rest on `timeOffset` (the
`transcodeOffset` extension), is open, owed to the study rung.

The fill-while-read primitive is the streaming slice's choice: `djherbis/stream` v1.5.1 (MIT, 2026-08-29), adopted or
ported, decided by a race test on cancel, disconnect, and failed exit. That library's `Seek` to the end blocks until
the stream closes unless `SetSeekEnd` is called. A non-zero ffmpeg exit or a cancelled context evicts the entry
rather than committing it, attached readers see the error, and the request returns 500, so a truncated encode never
serves as a complete track. On a hit the entry serves like a static file, ranges included. A disconnect kills ffmpeg
only when no other reader is attached. One knob bounds concurrent ffmpeg processes (default two); requests past it queue
in arrival order under a bounded wait, and a request past the cap returns error code 0 with a message rather than
holding the connection. The cache enforces its size cap, drops every variant of a track the catalog removes or retags,
and reports size, cap headroom, and hit rate as doctor rows. Art variants share the cache and the cap.

### Served volume

Revision 2.5 (P conflict 3, rulings 1, 4, and 5; rows A6 and A8). The principles spec's "Downloads and offline listening"
section is the design. A member's own devices holding the whole library is intended use (ruling 1), so nothing here
refuses a request. Slice 5 builds the meter and the notice (P conflict 3). This revision also places the volume view
in slice 5, over the rung 14 admin API.

- **The meter.** Every `stream` and `download` response counts against the member and the credential that
  authenticated it, in bytes sent and distinct tracks over the trailing 24 hours. The meter is store-backed, so a
  restart doesn't reset its window. Bytes are the bytes written per response, including a 206 response and an aborted
  stream. A credential's distinct tracks are those with any byte served to it in the window. A member's are the union
  over the member's credentials. A request with a nonzero Range offset adds bytes and no track. The credential id
  comes from authentication, never from the client's `c` parameter. A checkout under a device credential counts like
  any other served bytes. The meter keeps its own store-backed counts per P's acceptance row A6; it does not read
  the event table hook 20 adds, so slice 5 carries no dependency on slice 7.
- **The notice.** When a member's distinct tracks in a rolling 24 hours pass half the library's tracks or 1,000 tracks,
  whichever is larger (ruling 4, a setting), the admin gets one email per crossing through the rung 5 SMTP notify path.
  It names the member and the credentials that crossed, words a whole-library sync as expected, and asks the admin to
  confirm the credentials are the member's.
- **The volume view.** An admin view over the rung 14 admin API names members and credentials on bytes and distinct
  tracks, shows stream counts only as group totals, and shows no time-of-day breakdown (ruling 5). It names no track or
  album. Each member's own page carries ruling 5's sentence on what the admin sees and what play history records.

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

Revision 2.5 (P conflict 3, ruling 4, row A7). Slice 6 adds three things, as the principles spec's
"Every credential belongs to a member" specifies.

- **A credential cap per member**, default 10, a setting the admin can change per member. The count covers app
  passwords of both kinds (recoverable and hash-only) and the configuration homes spec's device credentials, and not
  web sessions. At the cap the device page refuses a new credential, lists the member's credentials by last use with
  any unused for 90 days first, and offers a one-step revoke on each. A member already over a lowered cap keeps every
  credential and can't mint another. The refusal is a web page and never reaches a Subsonic client.
- **An admin credential list** across all members: name, owner, kind, creation time, last use, and the number of
  distinct client addresses in the last seven days. The address comes from `serve.ClientIP`, so the count shows only in
  trusted-proxy mode, and a `doctor` row names the gap in the default mode.
- **A one-step admin revoke** of one credential, beside the roster-level epoch. The revoked credential's next request
  answers error 40.

### Shares

Shares ship at launch: `createShare`, `updateShare`, `deleteShare`, and `getShares`. A share resolves only the entity
ids it names. The token carries at least 128 bits from `crypto/rand` and is compared in constant time. The public path
is rate-limited per address, and a share is revocable alone.

Revision 2.5 (P conflict 1, rulings 8 to 12, row A4). A share is bounded so a link offers a pirate nothing a free
stream doesn't. The principles spec's "Bounded shares" section is the design and carries the evidence; this section
states what slice 6 builds. Revision 2's scope ("one album or one playlist"), its required expiry, and its
direct-path audio are superseded.

- **One item.** A share names one song or one album. A folder-view directory that holds exactly one album counts as
  that album, as it does for `download`. A playlist of up to 20 tracks may be shared, and the share snapshots its
  tracks at creation, so a later edit doesn't grow it (ruling 11). More than one id, an artist, a music folder, a
  directory above the album level, or a longer playlist answers error 50 with a message that names the rule. The web
  app offers only the allowed kinds.
- **A lossy stream, never a file.** A share serves 128 kbps Opus VBR (ruling 12), under the transcode budget. A
  browser that can't play Opus gets a codec it plays at an equivalent rate, and the slice confirms that fallback at
  build. A share has no download route, never serves the original format, and has no M3U list.
- **Expiry clamped, not refused.** A share expires 7 days after creation by default and never later than 30 days
  after creation (ruling 8). An absent, null, or zero `expires`, or a time already past, becomes the default, and a
  later time becomes the maximum. `updateShare` clamps the same way, measured from creation. Every stream and art
  request checks expiry and deletion, not only the page.
- **A pause when a link draws a crowd.** A share pauses at 50 distinct listeners over its life (ruling 9). A listener
  counts at the first audio byte, keyed by a random per-browser cookie the share page sets, and a stream request with
  no cookie counts as new. At the pause the member and the admin each get one notice through the volume notice's path.
  The member may resume once, which doubles the threshold, and the second pause is final. Share bytes count in the
  sharing member's meter under the share's id where a credential id would sit.
- **Unindexed.** The share page and every share response send `X-Robots-Tag: noindex` and `Referrer-Policy:
  no-referrer`. `robots.txt` never names the share path. No listing names a share except the member's own and the
  admin's view.
- **One share service, owner from authentication.** The web app and `createShare` call one service that enforces
  every bound. The owner is the authenticated member, and any owner, user, or username field in the request is
  ignored.

The share page names the member, as "Shared by Maya", and the share response carries the name (ruling 10). It shows
the album's buy links under the principles spec's "Money points toward buying". The share page's route is the one
public route that serves audio or art. No public route serves catalog data.

### Per-user state

Four tables keyed by surrogate user id and entity id: starred (one row per user, entity, timestamp), playlists (a header
with a monotonic version plus ordered entries), the play queue (one row per user with ordered ids, position, and
client), and play history (append-only events: user, song, client, timestamp, event kind, and how reported). No row
references a path. Revision 2.5 (P conflict 4, hook 20): every serve event also carries the authenticating credential
id and the bytes sent, and a download is an event kind. The views over play history (play counts, most played, smart
playlists) filter by event kind, so a download or a stream start never counts as a play.

Playlists are mutated by `updatePlaylist`'s deltas applied inside one transaction against the stored order, never
against a client-supplied list, so two clients removing one track each do not lose a write. The header version detects a
stale write and names the conflict behavior. `createPlaylist` with a song list is the one whole-list write.

A play is recorded when `scrobble` arrives with `submission=true`, matching the oracle. `submission=false` is
now-playing only. A stream start is a separate event kind and never a play, because most clients buffer a whole track
ahead of playback and a byte threshold would measure prefetch. The "how reported" column records provenance so a later
pass can change the rule without losing history. The play queue's edges are specified: a call with no `id` clears the
saved queue, `current` is required unless `id` is empty, an absent `position` means zero, and the response carries
`changed` and `changedBy`.

### Group features

**Report a problem.** A report names one admitted track or album, a kind (wrong tags, bad audio, wrong art, wrong
album grouping, other), an optional note, the reporting member's surrogate id, and a timestamp. It is an event-shaped
row under the rung 6 table rule. It surfaces as a decision-queue item (hook 16) that an admin resolves or dismisses, and
the reporter sees the outcome on the web app. OpenSubsonic has no report verb, so the web app carries the report page;
whether clients get a second path is an open item.

**Added by.** The album row carries `added_by`, the adding member's surrogate id (hook 17). Revision 2.5 (P conflict
5, row A10) replaces the rule that only intake sets it with the channel table in the principles spec's "dubplate
never fetches music" section. The server sets `added_by` from the authenticated principal of intake (the uploader
session, or a TUI contribute under a device credential) or of a mirror push (the credential's member). Copy between
collections carries it from the source row, and hand-over and seize preserve it, never resetting it to the admin.
The timer path, operator bulk intake, and rung 16's migration store null, shown as "unknown", and restore keeps the
backed-up value. No client field, API parameter, or admin form writes it. Each album also carries the member's source
answer (hook 21), which only an admin edits after admission. **Album notes** are a 2.0 table, one note per album,
written by the album's `added_by` member or an admin. The note reaches every client through `getAlbumInfo2`'s `notes`
field, which Navidrome fills from Last.fm; dubplate fills it with the member's name and note, and under the note a
source line such as "from Ana's vinyl" (P conflict 6).

**Collaborative playlists.** The playlist header already carries an owner. A collaborator table keys on playlist id
and surrogate user id. `updatePlaylist` accepts a write from the owner or a collaborator, and the stored-order delta
semantics above already keep two editors from losing each other's writes. Deleting a playlist and managing its
collaborators stay owner-only. OpenSubsonic has no collaborator verb, so the web app manages the list. A collaborator's
client sees the playlist through `getPlaylists` as it sees any public playlist.

**Smart playlists.** A smart playlist stores a rule set and no entries. `getPlaylist` evaluates it at read time for the
calling member, since the rules can read per-user state. The rules range over catalog fields (genre, year, added date,
added by, collection) and the caller's per-user state (starred, play count, last played, both derived from play
history). Clients see a read-only playlist; `updatePlaylist` on one is refused. The web app authors rules. Importing
Navidrome's `.nsp` files is an open item.

### Hosting and setup

**HTTPS.** The listener is a configured choice: plain HTTP behind a tunnel or proxy the operator runs, or TLS
terminated in the binary with certificates from Let's Encrypt through `golang.org/x/crypto/acme/autocert` (BSD-3). TLS
mode needs one value, the public hostname, plus ports 80 and 443. It keeps certificates under the state directory, so
backup covers them. A direct listener has no proxy, so the real-client-IP rule reads the connection address and trusts
no forwarded header (hook 19). Hosts behind carrier-grade NAT still need a tunnel. The setup guide names Cloudflare
Tunnel and Tailscale Funnel as equal options, and dubplate depends on neither.

**Navidrome import.** `dubplate import navidrome` reads a stopped Navidrome's SQLite database and its music folder
mapping. It invites each Navidrome user as a member through the roster's normal add path, which sends a magic link;
passwords are never carried over.

Revision 2.5 (P, "A member cap bounds the roster", row A1): it counts the members it would add before it writes
anything, and if the total passes the member cap it stops, names the setting, and writes nothing. The operator raises
the cap or passes an explicit list of members to bring, so no skipped member's stars or playlists are orphaned.

It imports each user's stars, playlists, play counts, and last-played dates, matching tracks by the alternate keys
(release MBID, then audio key) and never by path. Every track it cannot match is reported
by name, and nothing is silently merged. Revision 3 (ruling 2): the import serves a stranger switching from Navidrome, and the group's first launch runs it once
against the Navidrome database in the retirement archive, so no harness-only export exists. Reading Navidrome's data carries no license question; its code is still never copied.

**arm64.** The release builds `linux/arm64` beside `linux/amd64`. modernc SQLite is pure Go, so the build needs no
cross toolchain. The gate builds both, and the client roster's walk runs once against an arm64 host.

**`dubplate init`.** An interactive command, and a first-run web page for a server started with an empty
configuration, that asks for the site name, the admin email, the library and state directories, SMTP settings, and the
listener mode. It sends a test email before it writes anything, mints every secret the chosen integrations need, and
writes the configuration file atomically. Its questions come from the config registry, each key's description and
sensitive mark (hook 18), so a key added by a later rung reaches `init` without a second list to maintain.

**`dubplate upgrade`.** It downloads a release, verifies its checksum against the signed release manifest, and stops
the server. The store takes its pre-migration snapshot (rung 6) before the new binary migrates it. It starts the new
binary and waits for `doctor` to pass. On a failed check it restores the snapshot and the previous binary together,
because the store refuses an older binary after a migration. It never runs unattended in 2.0; an automatic schedule
would be a later ruling.

### API surface

One handler set on the standard library mux under the owned prefix, answering XML and JSON from the same typed response
structs. The Subsonic path is a sub-prefix of the one prefix the build spec's rung 1 rule owns, and a client's
configured base URL points at that sub-prefix, so the one-owned-prefix rule holds without amendment. The ingress strip
set still applies to it.

Every `subsonic-response` carries `openSubsonic`, `serverVersion`, and `type` beside `status` and `version`.
`getOpenSubsonicExtensions` is reachable unauthenticated, the surface's one carve-out.

Revision 2.5 (S finding 5): browser clients need CORS. Feishin's web build and Airsonic Refix run in a browser, and the
`transcoding` extension's JSON POST triggers a preflight. The API surface answers CORS requests, preflight for an
`application/json` body included. A preflight answer carries CORS headers and no data, so
`getOpenSubsonicExtensions` stays the only endpoint that returns data without credentials (principles row A3). The
policy itself (which origins, which headers, and whether credentials are ever allowed) is open, owed to the study
rung, measured against the browser clients in the roster.

Launch implements the endpoints the client roster exercises: system (`ping`, `getLicense`, `getOpenSubsonicExtensions`,
`getScanStatus`), browsing in both ID3 and folder views, album lists, search, media (stream, download, cover art, lyrics
stubbed), annotation (star, unstar, scrobble), playlists, play queue, shares, and read-only user. `getAlbum` and
`getAlbumInfo2` fill `musicBrainzId` with the admitted release id (revision 2.5, P conflict 6, row A15).

Revision 2.5 (P conflict 2, row A5): no single request returns more than an album. `download` accepts a song id or an
album id and returns the song or an album zip in the pipeline's own layout with sanitized names. A folder-view
directory that holds exactly one album counts as that album. Every other id kind (an artist, a playlist, a directory
above the album level, a music folder) answers error 50, which the OpenSubsonic error list defines as "User is not
authorized for the given operation.", and its message names the song-or-album rule. Error 70 would be untrue for an
id that exists, and error 50 is what Navidrome answers when downloads are off. Revision 2's artist and playlist
zips, and the playlist zip's `.m3u8`, are superseded.

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

Revision 2.5 (P conflict 3): the admin pages also carry the admin credential list with its one-step revoke and the
volume view, and the device page carries the credential cap's refusal ("Identity and app passwords", "Served volume").

### Digital audio players

Device sync is a local-binary job, not a server feature. Every capability a sync needs is present on a workstation and
absent on the VPS, which reaches the group through Cloudflare. 2.0 grows no device verb and no MTP. The largest real
DAP class is a phone, which the clients' own offline caches already serve. The server's contribution is `download`,
specified above.

Revision 2.5 (P conflict 2): `download` returns at most one album, so a player loads one album at a time. From 2.1,
the local binary's DAP export profile is the other path, gated on a real device in the group.

### Cutover and operations

Revision 3 rewrote this section; the heading keeps its name because other documents cite it. No switch from an
incumbent happens: the group's first launch is the streamer (ruling 1). The box runs the dubplate binary, ffmpeg, and
the timers, with no Navidrome, no cloudflared, and no driver. Starred lists, playlists, and play history built in Navidrome enter once
through `dubplate import navidrome` (see "Hosting and setup"); differences are reported, never silently merged.

The streamer's configuration holds these properties for every slice:

- The streamer is always on. It declares no enable flag and no integration class.
- Its tunables (cache cap, concurrent encodes, the profile set) are registered thresholds with defaults.
- Any streamer secret, such as the app-password key, follows the auth secrets' pattern (the session secret and its
  `_PREVIOUS` rotation row) and is generated by slice 8's `init`, so an operator never types one.
- Each slice plan states its delta to the set of values a default install's operator supplies by hand. A nonzero delta
  needs an owner ruling (ruling 7).

Provisioning adds ffmpeg and the cache directory. The streamer's state lives under the existing state directory, so
backup covers it. Doctor gains rows for ffmpeg's version, cache size against cap, index rebuild parity, and the drift
count.

## The ladder

Magnitude bands are relative and carry no time units, matching the build spec's ladder. A rung's ceiling and checkpoint
interval are set in its own plan, never here.

1. Study (M). A written record of what Navidrome and gonic do where the spec is silent, cited to a version and a harness
   test, with the ideas and lessons their code teaches recorded in dubplate's own words. One study question: how
   Melodee's ingest pipeline (staging, validation, admission into the library) compares with dubplate's
   intake-as-the-one-door model, recorded as ideas and lessons. Melodee is MIT, so its code may be reused as well as
   read. Ships the pre-seeded divergence
   ledger, the empty-success endpoint set, the launch extension subset, the per-endpoint oracle assignment, and the
   prior-art table's open cells. Revision 2.5 adds the answers to the items "Open items" marks as owed to the study
   rung.
2. Harness (M). The fixture library, the pinned Navidrome and gonic containers, the JSON schema validator and the XML
   rule the study rung sets, the recorded session, the differ, the divergence ledger, and the client roster's scripted
   walk. The client and the validator wiring are the protocol module's client and conformance checker, built under its
   import boundary. The checker carries `getkin/kin-openapi` as a module dependency, where the survey adopted it for
   tests only; this rung's plan weighs that.
3. Catalog projection and browsing (L). Splits into passes at its own planning; it carries the projector, both browse
   views, and the id space.
4. Art and search (M).
5. Streaming and transcoding (L). Splits at its own planning; it carries the cache, the concurrency model, and range
   behavior on both paths. Revision 2.5 adds the served-volume meter, its admin notice, and the volume view.
6. Identity, app passwords, and shares (M). Revision 2.5 adds the credential cap, the admin credential list and its
   one-step revoke, and the share bounds. Revision 3 records the bearer bypass's two contract sentences: a bypassed
   handler must not read the session cookie, and the pattern registers twice because either registration alone leaves
   one form refusing. The bypass is path-keyed.
7. Per-user state and group features (L). Splits at its own planning; it carries starred, playlists (collaborative and
   smart), the queue, play history, album notes, and the report page.
8. Hosting and setup (M). The TLS listener, `dubplate import navidrome`, the arm64 build, `dubplate init`, and
   `dubplate upgrade`. The import serves a stranger switching from Navidrome, and the group's first launch runs it once
   against the Navidrome database in the retirement archive (ruling 2). Revision 3 adds these acceptance criteria:
   - The import prints a dry-run report before any write.
   - It matches a member by address, then by username under `COLLATE NOCASE`, and refuses with a report when the two
     disagree.
   - Per-user writes land in one transaction, or are keyed so that a rerun is a no-op.
   - The import reads an extracted copy of the database, never the archive itself.
   - The unit carries `CAP_NET_BIND_SERVICE` in TLS mode only.
   - The certificate directory is a leaf: the unit's `ReadWritePaths`, `deploy.sh`'s pre-create, and the state-dir
     agreement leg all name it.
   - Owner ruling 3: one short-lived VPS with no age store and no `deploy.sh` goes from the release artifact through
     `init` to a played track. Each unclear step either blocks this slice's close or goes to `IDEAS.md` by owner
     ruling.

Slices 1 to 8 are rung 12's passes. Each is a pass with its own plan, ceiling, oracle diff, and roster walk, under the
conducting loop in CLAUDE.md. Slices 1 and 2 touch no production surface: slice 1 runs now as a document lane, and
slice 2 as a code lane in its own worktree.

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
   resolves a roster identity to it. The roster's `username` (renamed from `streamer_username` by revision 3, ruling 3) becomes required, unique, and stable
   rather than optional, and it is the renameable handle the surrogate id protects against.
6. Retired in revision 3 (ruling 4: dubplate integrates with no external streamer, so no driver seam exists). Text kept
   as written: The driver seam stays universal, rung 12. Address, proxy, and provisioning choreography are per-driver capabilities a
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

> Erratum, 2026-09-23 (rung 6 spec, erratum 11, ruling 2 yes): hooks 1, 2, 3, and 11 above each
> read "rung 6 schema" for where their column lands. That column lands with the rung that
> computes it, not as a nullable column added at rung 6c: the audio key, file hash, and sort
> key migrate in with rungs 7 and 9, and the gain fields migrate in with rung 9. CLAUDE.md's
> 2026-09-12 refinement ("a hook is built by its first consumer") governs; nothing an adopter
> sees changes.
15. Configuration homes, rung 6 and rung 14. The store carries a settings table (source, apply mode, sensitive flag,
    audit table), per-member overrides and preferences tables, a device-credential table, two quota counters, a
    move-intent journal, and on the collection row `owner` (non-null surrogate id), `visibility`, a server-allocated
    disjoint `root_path`, a writer epoch, a mirror generation, a last-mirror time, and a duplicate-collapse flag beside
    `writer`, with share rows reserved beside visibility and foreign keys enabled in the DSN. The rung 14 admin API
    exposes settings and collections on two mounts (cookie and bearer) so the web app and the TUI share one surface.
    Retired in revision 3: hook 6's capability report, which also declared a driver's shape constraints (subtree collections, art filenames read,
    tags read) and per-user folder access, is retired with hook 6, and private stays unsettable until slice 3's
    projector enforces visibility on every endpoint. Capability reporting is dropped unless a slice names a consumer. Design:
    `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`. Amended 2026-09-25 (P conflict 1): the share
    rows stay reserved, with no table or column before slice 6 builds them, and they carry the fields the share bounds
    need: the item (or a playlist share's track snapshot), creation and expiry times, the listener count and threshold,
    the paused and resumed state, and the listener keys.
16. A decision-queue item that names an admitted entity, rung 13. The queue's item model admits an item referencing an
    admitted track or album, filed by a member, beside the intake items it holds today, so "Report a problem" joins the
    queue without a second queue or a reshaped item table. Added 2026-09-24 (group features).
17. The uploading member on the submission, rung 7, carried to the album row at rung 9's admission as `added_by`. The
    submission knows its authenticated uploader only at intake; recovering it later means guessing. Added 2026-09-24
    (group features). Amended 2026-09-25 (P conflict 5): "only intake sets it" is superseded by the principles spec's
    channel table. The server sets `added_by` from the authenticated principal of intake or of a mirror push; copy,
    hand-over, and seize preserve it; no client field writes it ("Group features", "Added by").
18. A description and a sensitive mark on every registered config key, rung 14. The settings surface is the first
    consumer of both, and `dubplate init` reads the same declarations, so setup questions are never a hand-kept list.
    The wall rows and integration-class keys carry neither today. Added 2026-09-24 (hosting and setup).
19. The listener as a configured choice, slice 8 (re-homed from rung 12's driver seam by revision 3). The proxy-hygiene ingress keeps plain HTTP and a direct TLS listener
    both possible, and the real-client-IP rule has a direct mode that trusts no forwarded header. Added 2026-09-24
    (hosting and setup).
20. The credential and the bytes on every serve event, 0.8 slice 7. Every serve event carries the authenticating
    credential id and the bytes sent, and a download is an event kind. The event table's first builder is 0.8's
    per-user state, under the "built by its first consumer" rule, so no current rung builds it; a current rung plan
    that shapes an event-shaped table (hook 4) weighs this hook. Added 2026-09-25 (P conflict 4).
21. The source answer on the submission, rung 7, carried to the album row at rung 9's admission beside `added_by`.
    The uploader asks where the music came from once per upload, with a per-album override, and the answer is an
    intake API field so every intake client carries it. Intake refuses a submission without a known value, in the
    reject-wording contract (ruling 6, required). A mirror records the member's answer when each mirrored collection is
    created, as a default. After admission only an admin edits it. Added 2026-09-25 (P conflict 5, "The source
    question").

## Prior art

Every row needs a license cell citing the repository's LICENSE file by URL, because revision 1's strings were filled
from memory and two were wrong. Cells the reviews established are filled below; the rest are the study rung's
deliverable, scored on CLAUDE.md's four criteria with evidence per cell. Maintenance keeps its criterion but loses its
bus-factor weight: bus factor is 1 across this entire field, Navidrome, gonic, LMS, Koel, Black Candy, and Melodee
alike, so that signal discriminates nothing and adoption evidence does the real work.

| Entry | Role | License (URL) | Quality | Maintenance | Adoption | Verdict |
|---|---|---|---|---|---|---|
| OpenSubsonic specification and extension list | The contract | Apache-2.0 ([LICENSE](https://raw.githubusercontent.com/opensubsonic/open-subsonic-api/main/LICENSE)) | n/a | Live, 10 extensions | The de facto contract | Adopt. Read it; never vendor it, because the NOTICE clause would attach. |
| OpenSubsonic OpenAPI schema and Subsonic 1.16.1 XSD | The contract tier | OpenAPI: Apache-2.0 ([LICENSE](https://github.com/opensubsonic/open-subsonic-api/blob/main/LICENSE)); XSD: no license text, open | n/a | Schema self-declares work ongoing | Ampache validates JSON against the OpenAPI file and legacy XML against the XSD in CI | Adopt as harness inputs, the OpenAPI file as a bundled build pinned by checksum. The OpenAPI file covers JSON only, and the XSD rejects the OpenSubsonic envelope attributes, so the XML rule is open ("Evidence"; revision 2.5, S findings 1 and 7). |
| Navidrome | Primary oracle; ideas and lessons | GPL-3.0, URL from study rung | n/a | Very active, near-monthly | 23,396 stars; every client names it | Adopt as oracle, pinned by digest. Read for ideas; never transliterate. |
| gonic | Folder-view oracle; ideas and lessons | GPL-3.0, URL from study rung | n/a | Active | 2,538 stars; the second server clients name | Adopt as the oracle for `getIndexes`, `getMusicDirectory`, `getMusicFolders`. Its author also writes taglib, so both share one bus factor. |
| Melodee | Readable reference implementation of the ingest-pipeline-plus-OpenSubsonic shape | MIT, URL from study rung | study rung | Active, 1,478 commits since 2024-10 | 65 stars; publishes its own compatibility matrix | Adopt as readable and reusable prior art. Its Postgres and Blazor stack rules it out as an oracle. |
| `go.senan.xyz/taglib` | Tag read and write | LGPL-2.1, whole module ([LICENSE](https://github.com/sentriz/go-taglib/blob/master/LICENSE)) | Mostly good; wazero cache hardcoded to `os.TempDir()`, host module panics with no recover | v0.14.0, 2026-07-24; depends on wazero at an untagged pseudo-version | 11 importers; Navidrome runs its own fork | Open ruling 2. The "MIT wrapper" claim was false and the MPL-1.1 escape does not exist. |
| `tetratelabs/wazero` | Runs the WebAssembly | Apache-2.0, URL from study rung | High | v1.12.0, 2026-05-29 | 714 importers | Adopt only as taglib's transitive dependency, never directly. |
| `modernc.org/sqlite` | cgo-free SQLite | BSD-3, URL from study rung | Transpiled C; the house bar applies only to the `database/sql` layer | v1.59.0, 2026-09-15, near-weekly; the repo pins it | 3,518 importers | Adopt on a condition: rung 6 proves FTS5's trigram tokenizer and WAL under concurrent readers by a test in the gate. The prior-art survey's scratch test proved trigram with `remove_diacritics` on v1.59.0, outside the gate; the gate test and the WAL half are still owed (revision 2.5, S finding 7). |
| `ncruces/go-sqlite3` | The named fallback | MIT, URL from study rung | Well-documented VFS design; a wasm2go translation of SQLite, not a hand-written driver (erratum 2026-09-23, rung 6 spec, erratum 12) | Active | 96 importers; gonic uses it | Record as the fallback if the trigram or WAL check fails. |
| `mewkiz/flac` | Pure-Go FLAC decode; the intake integrity check | Unlicense, URL from study rung | Idiomatic package split | v1.0.14, 2026-08-19 | 96 importers | Adopt for rung 7. It supplies audio-frame decode, which the repo's own FLAC reader deliberately lacks. |
| Picard tag-mapping tables | Which tag names each container uses | CC0 1.0, URL from study rung | n/a | Versioned with Picard | The canonical mapping | Adopt, cited to a pinned doc version. |
| `golang.org/x/image`, `gen2brain/webp` | Art resize and encode | BSD-3 and MIT, URLs from study rung | Exemplar; webp is libwebp transpiled to Go through `wasm2go`, not a wazero runtime, and it loads a host shared library through `purego` first unless built with the `nodynamic` tag | x/image v0.46.0, 2026-09-08; webp v0.6.4 | 1,948 and 18 importers | Adopt, with webp built under `nodynamic` so output is identical on every host. The resize filter is a stated choice with a measurement, because the cache pays the encode once (revision 2.5, S finding 7). |
| `golang.org/x/time/rate` | Per-account and per-address auth rate limits | BSD-3, URL from study rung | Exemplar | Go team | 14,348 importers | Adopt. The repo's MusicBrainz limiter solves a different problem. |
| `crypto/aes` and `cipher.NewGCM` | App-password encryption at rest | stdlib | n/a | n/a | n/a | Adopt the stdlib; no KDF library is needed. |
| Schema migrations | Rung 6's mechanism | stdlib | n/a | n/a | n/a | Hand-roll with `embed.FS` and `PRAGMA user_version`. `pressly/goose` is declined: it direct-requires ten database drivers for a Go-function migration mechanism no hook asks for, and its decline rests chiefly on its silent-ahead `Up`, which applies every migration newer than the tracked version with no confirmation step (erratum 2026-09-23, rung 6 spec, erratum 12). |
| Stream-through transcode cache | 2.0 rung 5 | stdlib; `djherbis/stream` MIT if adopted | n/a | n/a | n/a | Hand-roll the size cap and eviction. `djherbis/fscache` is declined: 7 importers, its last tag is v0.10.1 (2020-04-15), and Navidrome pins an untagged pseudo-version. The fill-while-read primitive under it, `djherbis/stream` (MIT), now has a tagged v1.5.1 (2026-08-29); the streaming slice adopts or ports it, decided by a race test (revision 2.5, S finding 7). |
| `internal/atomicfile` | Streaming atomic writes for blobs and cache entries | in-repo | n/a | n/a | n/a | Extend it with a streaming write rather than inventing a second temp-file convention. |
| M3U read and write; Subsonic XML and JSON dual encoding | First-launch import and the response layer | stdlib | n/a | n/a | n/a | Hand-roll both. The study rung enumerates the encoding quirks (attribute-heavy XML, the wrapper, singleton-versus-array in JSON, JSONP) so the harness tests them. |
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

- The client roster's exact membership. Geoff ruled it on 2026-10-05: Feishin and Rufin on the Linux desktop, Amperfy on
  iOS and iPadOS, and Tempus on Android with Navic as the alternate. Open source is favored, and a closed client joins
  only if a member uses it daily. The record is
  `docs/superpowers/research/2026-10-05-client-roster-audit.md`.
- Whether folder browsing ships at launch at all. It doubles the browse surface and the id space, modern clients use
  ID3, and only gonic can grade it. The study rung brings the evidence.
- The app-password key's exact name in the age store. The identity slice mints it.
- Moot (revision 3): the shadow period's length and the Navidrome retention window. No shadow or retention exists.
- Whether clients get a report path beside the web app page. The candidate is a reserved per-member playlist named
  "Report a problem": adding a track files a report, which works in every client with no new verb. The slice 7 plan
  decides.
- Whether an admin's fix for a reported album has an edit path by 2.0, or a report resolves by re-upload through
  intake. Rung 13's plan records which one exists.
- Whether smart playlists import Navidrome's `.nsp` files for switchers. The study rung records the format.
- Answered (revision 3): whether the study rung runs beside the current ladder. Slice 1 runs now as a document lane.

Revision 2.5 adds these, each open, owed to the study rung. The sources state the facts and don't make the decision.

- The XML validation rule: a legacy XML form against the official XSD, or a derived XSD that admits the OpenSubsonic
  attributes ("Evidence"; S finding 1).
- The `transcoding` extension: whether 0.8 answers `getTranscodeDecision` and `getTranscodeStream`, and how a posted
  `ClientInfo` interacts with the device profile ("Streaming and transcoding"; S finding 2).
- A Range request on a filling transcode: a declared divergence class, or `Accept-Ranges: none` as the oracle does,
  with seeking on `timeOffset` ("Streaming and transcoding"; S finding 4).
- The CORS policy: allowed origins and headers, and whether credentials are ever allowed ("API surface"; S finding 5).
- The AAC container for stream-through: fragmented MP4 or ADTS ("Streaming and transcoding"; S finding 6).
- The protocol module's license, MIT or Apache-2.0, and what NOTICE obligation it carries from the Apache-2.0 OpenAPI
  file it validates against ("The protocol layer is a separable module"; G).
- Which credential a stream authenticates with before slice 6 builds app passwords, owed to the slice 5 plan
  ("Served volume"; P conflict 3).
- The word "collection". OpenSubsonic PR #237, "Add collections" (open, 51 comments, updated 2026-09-20), defines a
  collection as a user-owned ordered list of items, while dubplate's collection maps to a music folder. If the PR
  lands, API docs and UI copy need a distinct word for one of them. PR #162 (open) would extend `musicFolderId` to
  more endpoints, which bears on "honored per endpoint" under "Catalog projection". The study rung tracks both, with
  PRs #258 and #261, and re-reads them before the harness rung pins a spec build (S finding 8).
