# dubplate-server 2.0: the built-in streamer

**Re-keyed to the 0.8 release (Geoff, 2026-09-25 ruling on version numbering).** Everywhere
this document says "2.0" it means 0.8. The title and body are left as written.

Status: **Ratified 2026-09-07.** Revision 3.1, 2026-10-07 (the study rung's answers: slice 1 of rung 12 closed, so each open item it owned is answered, answered pending ruling, or routed, the prior-art cells are filled, and hook 22 is added; listed under "Revision 3.1, 2026-10-07"). Revision 3, 2026-10-05 (the built-in streamer is the group's first launch, so every passage that assumed an incumbent, a shadow, or a flip is removed or rewritten; listed under "Revision 3, 2026-10-05", with the rulings recorded under "Rulings of revision 3"). Revision 2.5, 2026-09-25 (errata from the ratified group-scale principles spec, the streamer prior-art survey, and the give-back decision; listed under "Revision 2.5, 2026-09-25"). Revision 2.4, 2026-09-24 (Geoff's ruling: five hosting and setup features join the 2.0 target, recorded under "Hosting and setup at launch", "Hosting and setup", the widened Navidrome import, forward hooks 18 and 19, and slice 8). Revision 2.3, 2026-09-24 (Geoff's ruling: four group features join the 2.0 target, recorded under "Group features at launch", "Group features", forward hooks 16 and 17, slice 7, and three open items). Revision 2.2, 2026-09-12 (forward hooks 1 and 15 and the web surface amended from the configuration homes spec, `docs/superpowers/specs/2026-09-11-configuration-homes-design.md`). Revision 2.1, 2026-09-07 (the three owner rulings recorded in "Rulings on the review round"). Revision 2, 2026-09-07. Every decision below was taken with
Geoff in the 2026-09-07 brainstorm and none is reopened here. ROADMAP entry: `streamer-2-0`. It is rung 12 of the ladder: its eight slices are rung 12's
passes, and the group's first launch (0.8.0) runs on it.

This document governs the initiative's purview, its architecture, and the forward hooks the current build owes it. Where
it names a current rung (5, 6, 7, 8, 9, 12, 16), it asks that rung's plan to weigh a hook, never to build a 2.0 feature
early. One sentence in revision 1 trimmed rung 12 while the document claimed to amend nothing; that conflict is open
ruling 1 rather than a silent amendment.

## Revision note, 2026-10-10: slice 1's forks A, B, C ruled

Slice 2b closed the three owner forks that slice 1 raised, each ruled by Geoff on 2026-10-07 with the study's
recommendation (study record section 19):

- **A. No `transcoding` extension at 0.8.** `getTranscodeDecision` and `getTranscodeStream` are not served at 0.8 and
  answer error 0.
- **B. The protocol module is MIT, and its conformance checker loads the OpenAPI file as caller input and never
  vendors it.** The same rule holds for the Subsonic XSD, which lives in the user cache, verified by sha256, and is
  never committed.
- **C. No folder browsing at launch.** gonic loses its folder-view role and keeps `getMusicFolders`. The seed
  session's `getIndexes` and `getMusicDirectory` rows are reference observations that no dubplate grade compares with
  an oracle. The oracle-tier paragraph under "Evidence" is amended to match.

## Revision 3.1, 2026-10-07

Revision 3.1 writes the study rung's answers into the spec. Slice 1 of rung 12 ran as a document lane and closed with a
study record, `docs/superpowers/research/2026-10-06-slice-1-study-record.md` ("the study record"), and a pre-seeded
divergence ledger, `docs/superpowers/research/2026-10-06-slice-1-divergence-ledger.md` ("the ledger"). Seven segment
records sit beside them, each named `docs/superpowers/research/2026-10-06-slice-1-<name>.md`: `evidence-pins`,
`contract`, `client-demands`, `base-url`, `server-behavior`, `melodee-intake`, and `prior-art-cells`. Section numbers
below are the study record's and row ids are the ledger's. This revision edits in place, and the earlier revision
sections stay as written. Three items are owner forks. The spec marks each "answered pending ruling", states the
record's recommendation, and does not act on it until Geoff rules.

1. Answered, with a pointer to the study record: the XML validation rule (section 5), Range on a filling transcode
   (section 7), the CORS policy (section 8), the AAC container (section 9), the empty-success endpoint set and the
   launch extension subset (sections 1 and 2), the `musicFolderId` endpoint list (section 4), the `.nsp` format (section
   13), the word "collection" and the PR watch (section 11), and Melodee's intake pipeline (section 15). "Evidence",
   "Streaming and transcoding", "Catalog projection", "API surface", and "Open items".
2. Answered pending ruling, each with the record's recommendation: fork A, the `transcoding` extension (no at 0.8,
   revisited at 1.0 with the format ladder; section 6); fork B, the protocol module's license and NOTICE (MIT, with the
   checker loading the OpenAPI file as caller input; section 10); fork C, folder browsing at launch (no at 0.8,
   reconsidered at 1.0 if a member asks; section 12). The body keeps its present text under each fork until Geoff rules.
   "Streaming and transcoding", "The protocol layer is a separable module", "Catalog projection", and "Open items".
3. Range on a filling transcode matches Navidrome: 200 from byte 0 with `Accept-Ranges: none`, seeking on `timeOffset`,
   and the offset in the cache key. This retires the sentence in "Streaming and transcoding" that served a range from
   the filling entry once the requested offset was written, and the revision 2 item 6 clause "serves ranges from the
   filling entry". A failed encode after the first byte aborts the connection (section 7; ledger L-28 to L-32).
   "Streaming and transcoding".
4. The device profile is a ceiling on bitrate and a default format, which amends the profile resolution order in
   "Streaming and transcoding" (section 18, item 1; L-36). "Streaming and transcoding".
5. AAC streams as ADTS with `Content-Type: audio/aac`, and a playback test is owed (section 9). "Streaming and
   transcoding".
6. The CORS policy: `Access-Control-Allow-Origin: *` on every Subsonic response, no credentials, a bodyless preflight,
   and nothing outside the Subsonic sub-prefix (section 8; L-38). "API surface".
7. JSONP is declined, every API response carries `X-Content-Type-Options: nosniff`, no Subsonic response sets a cookie,
   and every method answers with and without the `.view` suffix (section 18, item 2; section 14; L-01, L-37, A-07). "API
   surface".
8. The empty-success set, the unimplemented-endpoint answer, and the launch extension subset are written into "API
   surface" (sections 1 and 2). `getScanStatus` carries `lastScan` and `startScan` answers the current scan status
   (section 18, item 7; L-20, L-21). "API surface".
9. API-facing prose says "music folder" for the wire concept and "collection" for the store concept, and dubplate never
   declares the `collections` or `publicImageUrl` extensions (section 11). "API surface".
10. The XML rule is the legacy form: every response is validated against the unmodified XSD after a scoped projection,
    the XML shape of the OpenSubsonic additions is a declared unvalidated class, and the harness fetches the XSD into a
    pinned cache by sha256 without redistributing it (section 5). "Evidence".
11. The prior-art table's open cells take the cell record's license URLs, each fetched with status 200 on 2026-10-07,
    and its dated figures. Melodee's quality cell takes the merged text (section 16). Five verdict wordings change, each
    with its reason: the OpenSubsonic row's NOTICE reason is corrected (the repository carries no NOTICE file), the
    Melodee row gains a provenance condition and a bursty maintenance reading, the Picard row gains a citation rule, the
    `modernc.org/sqlite` row gains the `libc` pairing condition, and the XSD row records the XML rule. "Prior art".
12. Hook 22, a stable never-reused integer wire id per collection, joins the forward-hooks list after hook 21, with
    slice 3 as its first consumer (section 17; L-18). "Forward hooks the current build owes".
13. Ladder entries 1 to 8 name what the study routes to each slice (study record, "Routed"). "The ladder".
14. "Open items" marks each item the study owned. The credential-before-slice-6 bullet stays open and is routed to the
    slice 5 plan, because the study does not answer it.

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
are gathered under "Open items". Revision 3.1 marks each one answered, answered pending ruling, or routed, at its site.

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
   Revision 3.1 retires the ranges clause: a filling entry ignores Range ("Streaming and transcoding"); the eviction and
   queue halves stand.
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
them. Either way the choice is recorded as a declared divergence. Answered (revision 3.1; study record, section 5): the
legacy form. Every response is validated against the unmodified official XSD after a projection that removes the
OpenSubsonic additions, scoped per element type and keyed on the pinned OpenAPI file, because a global strip would
remove `status` and `version` from the envelope and fail every body for the wrong reason. A name found in neither the
XSD nor the OpenAPI type stays in place and fails, so a misspelled attribute is caught. The XML shape of the additions
follows one rule (scalars as attributes, lists of records as repeated elements with no wrapper, lists of strings as
repeated text elements) and is a declared unvalidated class, because no authority defines it and an XSD derived from the
encoder's own rule would prove only self-consistency. The oracle diff covers the fields an oracle emits. The harness
fetches the XSD into a pinned cache by sha256 and never redistributes it, since the file carries no license text.
Amperfy parses XML, so its demands on core attributes and on ISO 8601 timestamps with a fixed-width fraction become
roster-tier tests.

Revision 3: the pinned Navidrome and gonic containers stay as workstation-only behavior oracles in the harness, never
on the box, under the GPL behavior-only rule in "License".

The oracle tier serves the same fixture
library from Navidrome and from gonic and diffs dubplate's responses against them, with gonic authoritative for
`getMusicFolders` alone (Geoff, 2026-10-07, ruling C: no folder browsing at launch, so `getIndexes` and
`getMusicDirectory` are graded against the contract and never against an oracle; Navidrome synthesizes its folder tree
from ID3 rather than from real paths, and gonic's tree is reference observation only). The oracle image digest is pinned, the ledger records the oracle version per adjudication, and a
version bump is a deliberate pass that re-runs the ledger. The roster tier walks each slice in the group's apps.
Rejected: the roster alone (a client hides most of the API), the spec alone (silent where clients depend on de facto
behavior), a Navidrome-only oracle (it bakes in one server's choices, including two surfaces where it cannot grade
dubplate at all).

Two divergence classes are declared at the harness rung and adjudicated once as classes. The folder view diverges by
construction against Navidrome, which is why gonic grades it. The apiKey path has no oracle coverage at all and is
graded against the specification and the client roster. Search results compare as a set plus a top-N ordering assertion,
never field by field, because two FTS5 implementations with different tokenizers agree on membership and differ on order
constantly.

Revision 3.1 (study record, sections 1 to 3; ledger): the per-endpoint oracle assignment is answered. Every endpoint
takes the contract tier, and the study record's table names the oracle that grades it beyond that, Navidrome primary and
gonic secondary by default, or "contract only" where neither oracle is a fair grader. The pre-seeded divergence ledger
holds 54 divergence rows and 12 agreement rows, each with a class, the oracles' behavior, dubplate's intended behavior,
an evidence pointer, and the test or slice that proves it. Slice 2 transcribes the table and the ledger into the
harness.

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
  documentation, or an observation recorded by the harness, never GPL source. The study rung's record
  (`docs/superpowers/research/2026-10-06-slice-1-study-record.md`, with the segment records it indexes) is the citation
  source, and "License" below still governs.

Whether the wire types are generated from the OpenAPI file or written by hand is no longer open. Conductor reading
(2026-09-25, reversible), recorded in ROADMAP's give-back entry: the wire types are hand-written, and the pinned
Apache-2.0 OpenAPI file validates conformance in the harness. The evidence: the survey tried `oapi-codegen` v2, usable
in the harness but unfit as production types since it has pointer fields, no `xml` tags, and a `json.RawMessage` union
per response, and tried `ogen`, which failed on the full spec; this spec's prior-art table already hand-rolls the XML
and JSON dual encoding.

The second is the module's license. ROADMAP requires a permissive one. Whether it is MIT or Apache-2.0, and what notice
it carries from the Apache-2.0 OpenAPI file it validates against, is answered pending ruling (revision 3.1; owner fork
B; study record, section 10). The OpenSubsonic repository carries no NOTICE file, so Apache-2.0 section 4(d) attaches no
NOTICE text today, and sections 4(a) to 4(c) bind any redistribution of the OpenAPI file. The record recommends MIT,
with the conformance checker loading the OpenAPI file as caller input and verifying a pinned checksum, never vendoring
it. The module then redistributes no Apache-2.0 file and shares dubplate's license, which keeps Geoff's 2026-09-07
reasons under "License" intact. Under that option, the module's README tells its users to fetch the file themselves. The
watch item is a re-check of the specification repository for a NOTICE file at every pin bump. Geoff has not ruled.

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
is honored per endpoint rather than as a blanket filter, and the study record lists the endpoints that take it. Answered
(revision 3.1; study record, section 4): ten endpoints, `getIndexes`, `getArtists`, `getAlbumList`, `getAlbumList2`,
`getRandomSongs`, `getSongsByGenre`, `getStarred`, `getStarred2`, `search2`, and `search3`, which is the specification's
own list. Both oracles honor exactly these and ignore the parameter elsewhere. An unknown id answers error 70, and an id
for a collection the member cannot see gets the same answer and text, so a private collection's existence never leaks.
The id on the wire is a stable integer per collection, never reused (hook 22). `search3` filters exactly, because Rufin
runs it once per music folder and links each track to its folder. Clients that browse by directory get a folder view
derived from the paths the pipeline chose, served from a projection table rather than a filesystem walk.
Revision 3.1: whether the folder view ships at launch is answered pending ruling (owner fork C; study record, section
12). The record recommends no at 0.8, reconsidered at 1.0 if a member asks. Under no, `getIndexes` answers an empty
index list, `getMusicDirectory` answers error 70 for any id, gonic still grades `getMusicFolders`, and slice 3 decides
what a song's `parent` field carries.

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
state. The transcode path runs when the request names a format or a bitrate cap, or when the device's profile sets one.
The device profile is a ceiling on bitrate and a default format: a request may lower the ceiling and never raise it,
`format=raw` yields the original only when it fits under the ceiling, and the estate default applies where neither the
request nor the device names a profile. Revision 3.1 (study record, section 18, item 1; ledger L-36) replaces the
earlier order of request, then device, then estate default. That order left the device profile almost dead, because
Tempus and Rufin send `format=raw` and Amperfy sends `format=mp3` by default, so three of four roster clients name a
format before the device is consulted. The ceiling is Navidrome's ordering, and one decision path serves the legacy
parameters. Slice 5 designs the details.

Revision 2.5 (S finding 2): OpenSubsonic added a `transcoding` extension in December 2025 (`getTranscodeDecision` and
`getTranscodeStream`). Navidrome implemented it on 2026-03-09, and Feishin calls it. A third-party issue report says
Symfonium does too; the survey didn't verify that. Under it the client
posts a capability document (`ClientInfo`), and the profile order above has no place for one. Whether 0.8 answers the
extension, and if it does, how a posted capability document interacts with the device profile, is answered pending
ruling (revision 3.1; owner fork A; study record, section 6). Feishin is the only roster client that calls the
extension. With it declared, Feishin resolves each stream through a JSON POST decision by default and cannot play a
track when that decision answers error 0. At FLAC-only launch it direct-plays FLAC, and every other roster client
ignores the extension. The record recommends no at 0.8, revisited at 1.0 with the format ladder. Under no, `transcoding`
stays out of the launch extension subset, and Feishin plays through plain `stream` URLs with the legacy `format` and
`maxBitRate` parameters. If Geoff rules yes, slice 5 grows by two endpoints, a closed-list validator, a signed stateless
token bound to the file's update time, a decision function, and a token-signing secret that `init` mints, so the
operator delta stays zero. The posted `ClientInfo` then negotiates beneath the device ceiling above, and AAC through the
extension follows the ADTS ruling below.

Each profile is a named ffmpeg argument template held in configuration. A format selects a whole template by name from
the configured set and never substitutes a string. A bitrate is an integer validated against an allowlist, and a seek
offset is a validated numeric. Templates are argv slices with typed placeholders, never shell strings; no template may
carry `-i` or an output path, which the server supplies; configuration load rejects a template naming anything but the
declared placeholders. Any shipped MP3 profile writes gapless headers. ReplayGain is applied by the profile when the
profile asks for it, and never to the library file.

Revision 2.5 (S finding 6): stream-through constrains the container. MP4 with `-movflags +faststart` works only on a
finished file, which is why Koel encodes a whole AAC track before its first byte. MP3 and Ogg/Opus stream through as
they encode. A shipped AAC profile therefore writes fragmented MP4 or ADTS, and never a `faststart` MP4. Answered
(revision 3.1; study record, section 9): ADTS, served as `audio/aac`. ADTS is documented for all three native roster
players (Media3's default extractors for Tempus, GStreamer's `aacparse` for Rufin, and the `AudioStreaming` type table
for Amperfy), while fragmented MP4 is documented only for Tempus, and Navidrome's shipped AAC configuration also writes
ADTS. A playback test of an ADTS stream with a growing length is owed: Feishin and Rufin at slice 5's roster walk,
Tempus once the emulator exists, and Amperfy at rung 16.

On a cache miss, ffmpeg starts under the request's context with a per-track timeout, and its output streams into the
cache entry while readers attach, so a second listener shares the encode. A Range request on a filling entry is ignored:
`stream` answers 200 from byte 0 with `Accept-Ranges: none` and no `Content-Length`, and a complete entry serves ranges
as a static file. Seeking during a transcode rides `timeOffset` under the `transcodeOffset` extension, with the offset
in the cache key, and `estimateContentLength` is not honored. Revision 3.1 replaces the earlier sentence, which served a
range from the filling entry once the requested offset was written.

Revision 2.5 (S finding 4): this sentence diverges from the oracle and can't rest on the standard library. Navidrome
uses `http.ServeContent` only for a complete entry, answers a filling one with `Accept-Ranges: none`, and sends an
estimated `Content-Length` only when the client passes `estimateContentLength=true`. `http.ServeContent` can't serve
a filling entry, because it "uses a seek to the end of the content to determine its size", and a range answer on a
filling entry needs a `Content-Range` with an unknown total. Whether the streaming slice keeps this behavior as a
declared divergence class, or matches the oracle and lets seeking during a transcode rest on `timeOffset` (the
`transcodeOffset` extension), is answered (revision 3.1; study record, section 7; ledger L-28 to L-30): the slice
matches the oracle. The reasons are four. First, no oracle serves a range from a filling entry, since Navidrome answers
200 with the whole entry and gonic blocks until the encode ends. Second, Amperfy's default request is `format=mp3`, so
every default Amperfy play is a transcode, and its player decides seekability from status 206 or an `Accept-Ranges`
value other than `none`, which Navidrome's shape satisfies. Third, RFC 9110 section 14.2 says "A server MAY ignore the
Range header field." Fourth, the standard library cannot serve a filling entry, so a divergent design needed a
hand-rolled range writer, and the matched design needs none.

The fill-while-read primitive is the streaming slice's choice: `djherbis/stream` v1.5.1 (MIT, 2026-08-29), adopted or
ported, decided by a race test on cancel, disconnect, and failed exit. That library's `Seek` to the end blocks until
the stream closes unless `SetSeekEnd` is called. A non-zero ffmpeg exit or a cancelled context evicts the entry
rather than committing it, attached readers see the error, and the request returns 500 while no byte has been sent. Once
headers are written the server aborts the connection, as Navidrome does, so a truncated encode never serves as a
complete track and never ends cleanly (revision 3.1; study record, section 7; ledger L-31; `net/http.ErrAbortHandler` is
"a sentinel panic value to abort a handler"). On a hit the entry serves like a static file, ranges included. A
disconnect kills ffmpeg only when no other reader is attached; slice 5 weighs a short grace first, because Amperfy opens
a `Range: bytes=0-1` probe on each track and cancels it (ledger L-32). One knob bounds concurrent ffmpeg processes
(default two); requests past it queue in arrival order under a bounded wait, and a request past the cap returns error
code 0 with a message rather than holding the connection. The cache enforces its size cap, drops every variant of a
track the catalog removes or retags, and reports size, cap headroom, and hit rate as doctor rows. Art variants share the
cache and the cap.

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

Revision 3.1 (study record, sections 14 and 18; ledger L-01, L-20, L-21, L-37, A-07): every method answers at
`<base>/rest/<name>` and at `<base>/rest/<name>.view`, because Tempus never sends the suffix and Feishin, Rufin, and
Amperfy always do. JSONP is declined: `f=jsonp` gets the answer any unrecognized `f` gets, which is XML, and no callback
is reflected, since no roster client sends it and a reflected callback is script injection on a response that carries
member data. Every API response carries `X-Content-Type-Options: nosniff`, and no Subsonic response sets a cookie.
`getScanStatus` carries `lastScan`, set to the catalog's last change, and `startScan` answers the current scan status,
since the catalog is the index. API-facing prose says "music folder" for the wire concept and "collection" for the store
concept, and dubplate never declares the `collections` extension (OpenSubsonic PR #237) or `publicImageUrl` (PR #261).
The base-URL decision is slice 3's. The record measured every roster client carrying a path in its server URL, found
that the sub-prefix shape above meets every measured demand, and noted that Amperfy sends a plain GET to the configured
base URL before login and fails on any final status of 400 or more except 401, so the sub-prefix itself must answer
below 400.

Revision 2.5 (S finding 5): browser clients need CORS. Feishin's web build and Airsonic Refix run in a browser, and the
`transcoding` extension's JSON POST triggers a preflight. The API surface answers CORS requests, preflight for an
`application/json` body included. A preflight answer carries CORS headers and no data, so
`getOpenSubsonicExtensions` stays the only endpoint that returns data without credentials (principles row A3). The
policy itself is answered in revision 3.1 (study record, section 8; ledger L-38). Every response under the Subsonic
sub-prefix, success, error, and binary alike, carries `Access-Control-Allow-Origin: *`, with or without an `Origin`
header, and `Access-Control-Allow-Credentials` is never sent. A preflight answers with no body and no data, needs no
credentials, and allows `GET`, `POST`, and the request header `Content-Type`. Nothing outside the Subsonic sub-prefix
sends CORS headers. Only Feishin runs in a browser engine, none of its requests is credentialed, and its origin cannot
be known ahead (a hosted build, an Electron `file://` page, or `localhost`), so a named-origin list cannot work and a
wildcard can. Slice 3 owes the preflight rows, the no-credentials and no-data assertions, and a check of whether a
CORS-mode media element needs `Access-Control-Expose-Headers` for `Content-Range`.

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
never a bare 404. The exception is the set of endpoints clients require to answer with an empty success. Answered
(revision 3.1; study record, section 1): `getInternetRadioStations`, `getPodcasts`, `getNewestPodcasts`,
`getArtistInfo`, `getArtistInfo2`, `getTopSongs`, `getSimilarSongs`, and `getSimilarSongs2` answer `status: ok` with an
empty payload of their documented shape, and so does `getIndexes` if folder browsing is ruled out (fork C). An empty
success is allowed only on a read, and only where "nothing here" is true of dubplate's data; a write never answers one.
The unimplemented answer is graded against the specification, never an oracle diff. The launch extension subset is
answered too (study record, section 2), measured against the ten extensions at OpenSubsonic specification commit
`bed1688`: `formPost` and `transcodeOffset`, `apiKeyAuthentication` when slice 6 lands it, and `transcoding` only if
fork A is ruled yes. `songLyrics`, `indexBasedQueue`, `playbackReport`, `topSongsByArtistId`, `sonicSimilarity`, and
`getPodcastEpisode` are not declared at 0.8, and the declared list equals what is implemented at each slice's close.
Every endpoint is one function with one test table.

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
   rung. Revision 3.1: done; the record is `docs/superpowers/research/2026-10-06-slice-1-study-record.md`, and every
   item it owned is answered, answered pending ruling, or routed.
2. Harness (M). The fixture library, the pinned Navidrome and gonic containers, the JSON schema validator and the XML
   rule the study rung sets, the recorded session, the differ, the divergence ledger, and the client roster's scripted
   walk. The client and the validator wiring are the protocol module's client and conformance checker, built under its
   import boundary. The checker carries `getkin/kin-openapi` as a module dependency, where the survey adopted it for
   tests only; this rung's plan weighs that. Revision 3.1 routes here, among others, the ledger's transcription, the
   per-endpoint oracle map, the XSD's pinned cache and the scoped projection, the known-deviation list, a two-collection
   fixture, and a re-read of OpenSubsonic PRs #162, #237, #258, #261, and #278 before it pins a specification build
   (study record, "Routed" lists the rest).
3. Catalog projection and browsing (L). Splits into passes at its own planning; it carries the projector, both browse
   views, and the id space. Revision 3.1 routes here, among others, the integer music folder id (hook 22), both
   method-name forms, the CORS policy, the base-URL decision with its build-spec erratum, and the fixed-width timestamp
   layout (the full list is the study record's "Routed").
4. Art and search (M). Revision 3.1 routes here the WebP trigger from Feishin's `<img>` loads, with `Vary: Accept` on a
   format-varying `getCoverArt` (ledger L-39), error 10 for a missing `id` (L-40), and Amperfy reading an artwork body
   under 2,000 bytes as a possible error (study record, "Routed").
5. Streaming and transcoding (L). Splits at its own planning; it carries the cache, the concurrency model, and range
   behavior on both paths. Revision 2.5 adds the served-volume meter, its admin notice, and the volume view. Revision
   3.1 routes here, among others, the Range reading, `timeOffset` in the cache key, the device ceiling, the ADTS
   container and its playback test, and failure before and after the first byte (the full list is the study record's
   "Routed").
6. Identity, app passwords, and shares (M). Revision 2.5 adds the credential cap, the admin credential list and its
   one-step revoke, and the share bounds. Revision 3 records the bearer bypass's two contract sentences: a bypassed
   handler must not read the session cookie, and the pattern registers twice because either registration alone leaves
   one form refusing. The bypass is path-keyed. Revision 3.1 routes here, among others, the `apiKey` path and
   `tokenInfo`, error 42's roster check, error 43 for a credential in both the query and a form body, and the case for
   an enumerated allow-list over any path pattern or client header (the full list is the study record's "Routed").
7. Per-user state and group features (L). Splits at its own planning; it carries starred, playlists (collaborative and
   smart), the queue, play history, album notes, and the report page. Revision 3.1 routes here, among others, the queue
   and `scrobble` rows (ledger L-41 to L-49) and the smart-playlist row (the full list is the study record's "Routed").
8. Hosting and setup (M). The TLS listener, `dubplate import navidrome`, the arm64 build, `dubplate init`, and
   `dubplate upgrade`. The import serves a stranger switching from Navidrome, and the group's first launch runs it once
   against the Navidrome database in the retirement archive (ruling 2). Revision 3.1 routes the `.nsp` reads here:
   whether the retired Navidrome held any `.nsp` file or smart-playlist rows, and whether it keeps a smart playlist's
   rules in its database as well as the file (study record, section 13). Revision 3 adds these acceptance criteria:
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
22. A stable integer wire id per collection, slice 3. The OpenSubsonic music folder id is an integer in both the XSD and
    the OpenAPI schema, and clients keep a chosen folder id in their settings, so each collection carries an integer
    that survives restarts and `VACUUM` and is never reused after a delete. The store's collection id is a BLOB, and
    SQLite's implicit rowid does not qualify: "The VACUUM command may change the ROWIDs of entries in any tables that do
    not have an explicit INTEGER PRIMARY KEY" (`https://www.sqlite.org/lang_vacuum.html`, read 2026-10-07). Built by
    slice 3, its first consumer; a rung that creates collections before slice 3 is neutral, because slice 3's migration
    backfills. Added 2026-10-07 (revision 3.1; study record, section 17; ledger L-18).
23. No platform assumption in the core, every rung. 0.9's home install runs on Linux (arm64 and amd64), in a
    multi-arch container image, on macOS, and on Windows (Geoff, 2026-10-10). Code outside the deployment layer
    assumes no systemd, no Linux-only syscall, and no fixed path. An OS-specific check, such as rung 7's same-device
    test on staging and the library, sits in its own build-constrained file with a counterpart for the other platforms
    or a stated refusal there. `systemd/`, `cloud-init/`, and `deploy.sh` stay the VPS deployment and are neutral.
    A rung plan states per task whether it adds a platform-specific file. Added 2026-10-10 (Geoff's home install
    ruling; ROADMAP, `local-client-0-9`).
24. An install that needs no public host, slice 8 and rung 16. `dubplate init` and the listener (hook 19) accept a
    LAN-only setup with no public hostname, no ACME, and no outbound mail. The verify-identity seam (rung 5) admits a
    second sign-in method without rework, because a LAN install may have no mail relay for magic links. 0.8 builds
    neither the LAN sign-in nor the container image; it only keeps both possible. Added 2026-10-10 (Geoff: the
    standalone home install ships at 0.9, and 0.8 is built ready for it).

## Prior art

Every row needs a license cell citing the repository's LICENSE file by URL, because revision 1's strings were filled
from memory and two were wrong. Cells the reviews established were filled first; revision 3.1 fills the rest from the
study's cell record (`docs/superpowers/research/2026-10-06-slice-1-prior-art-cells.md`), scored on CLAUDE.md's four
criteria with evidence per cell. Every license URL below was fetched with status 200 on 2026-10-07, and every figure is
dated 2026-10-07 unless the cell says otherwise. Maintenance keeps its criterion but loses its bus-factor weight: bus
factor is 1 across this entire field, Navidrome, gonic, LMS, Koel, Black Candy, and Melodee alike, so that signal
discriminates nothing and adoption evidence does the real work.

| Entry | Role | License (URL) | Quality | Maintenance | Adoption | Verdict |
|---|---|---|---|---|---|---|
| OpenSubsonic specification and extension list | The contract | Apache-2.0 ([LICENSE](https://raw.githubusercontent.com/opensubsonic/open-subsonic-api/main/LICENSE)) | n/a | Live: 10 extensions at `bed1688`; 34 commits in 12 months, uneven; newest 2026-08-07 | The de facto contract; 251 stars; 21 open PRs | Adopt. Read it; never vendor it: sections 4(a) to 4(c) of the license bind any redistribution, and a NOTICE file added upstream would bind section 4(d). The repository carries no NOTICE file today (revision 3.1; study record, tension (c)). |
| OpenSubsonic OpenAPI schema and Subsonic 1.16.1 XSD | The contract tier | OpenAPI: Apache-2.0 ([LICENSE](https://github.com/opensubsonic/open-subsonic-api/blob/main/LICENSE)); XSD: no license text, open ([file](https://subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd), re-confirmed 2026-10-07) | n/a | Schema self-declares work ongoing | Ampache validates JSON against the OpenAPI file and legacy XML against the XSD in CI | Adopt as harness inputs, the OpenAPI file as a bundled build pinned by checksum. The OpenAPI file covers JSON only, and the XSD rejects the OpenSubsonic envelope attributes, so XML is validated in the legacy form after a scoped projection, fetched into a pinned cache and never redistributed ("Evidence"; revision 3.1, study record, section 5). |
| Navidrome | Primary oracle; ideas and lessons | GPL-3.0 ([LICENSE](https://github.com/navidrome/navidrome/blob/master/LICENSE)) | n/a | Very active, near-monthly: 15 releases in 12 months; v0.64.2 on 2026-09-24 | 24,007 stars; every client names it | Adopt as oracle, pinned by digest. Read for ideas; never transliterate. |
| gonic | Folder-view oracle; ideas and lessons | GPL-3.0 ([LICENSE](https://github.com/sentriz/gonic/blob/master/LICENSE)) | n/a | Active: 5 releases in 12 months; v0.22.0 on 2026-06-08, with `master` moving weekly; the author holds 84% of commits | 2,563 stars; the second server clients name | Adopt as the oracle for `getIndexes`, `getMusicDirectory`, `getMusicFolders`. Its author also writes taglib, so both share one bus factor. |
| Melodee | Readable reference implementation of the ingest-pipeline-plus-OpenSubsonic shape | MIT ([LICENSE](https://github.com/melodee-project/melodee/blob/main/LICENSE)) | Below the house bar in structure, assertion discipline, auth, and its commit step, though broad in test volume. Read at v2.2.0 (`d91342e`): 4,479 test attributes across six projects, yet no test moves an album through the admission path, and the streaming compatibility tests accept error statuses as passes. Validation fails open on a thrown check; the commit copies then deletes without verification and swallows per-album errors; a default repair rewrites multi-disc tags; an auth carve-out keys on the client's `Host` header and a path substring (a source reading, not run). Worth taking as ideas: the revalidation backoff and the stable-file check. | Bursty: 659 commits in December and January, 21 from February to April, 79 from May to July, none since 2026-07-12; 1,478 commits since 2024-10 | 66 stars; publishes [its own compatibility matrix](https://melodee.org/opensubsonic-matrix/) | Adopt as readable prior art for ideas. Any code reuse needs a per-file provenance read first, because two source comments say a parameter and a flow follow "Navidrome semantics" and were "seen in the Navidrome source code", and the GPL clean-room rule binds anything derived from that (revision 3.1; study record, section 16). Its Postgres and Blazor stack rules it out as an oracle. |
| `go.senan.xyz/taglib` | Tag read and write | LGPL-2.1, whole module ([LICENSE](https://github.com/sentriz/go-taglib/blob/master/LICENSE)) | Mostly good; wazero cache hardcoded to `os.TempDir()`, host module panics with no recover | v0.14.0, 2026-07-24, still the newest on 2026-10-07; depends on wazero at an untagged pseudo-version | 11 importers; Navidrome runs its own fork | Open ruling 2. The "MIT wrapper" claim was false and the MPL-1.1 escape does not exist. |
| `tetratelabs/wazero` | Runs the WebAssembly | Apache-2.0 ([LICENSE](https://github.com/wazero/wazero/blob/main/LICENSE); the repository moved to the `wazero` organization and the old URL redirects, while the module path is still `github.com/tetratelabs/wazero`) | High, confirmed at v1.12.0: Wasm Core 1.0 and 2.0 compliance claimed, and `golang.org/x/sys` its only requirement | v1.12.0, 2026-05-29, still the newest on 2026-10-07 | 714 importers | Adopt only as taglib's transitive dependency, never directly. |
| `modernc.org/sqlite` | cgo-free SQLite | BSD-3-Clause ([LICENSE](https://gitlab.com/cznic/sqlite/-/raw/master/LICENSE)) | Transpiled C; the house bar applies only to the `database/sql` layer | v1.60.1, 2026-09-29, near-weekly; the repo pins it | 6,070 importers on 2026-10-07 | Adopt on a condition: rung 6 proves FTS5's trigram tokenizer and WAL under concurrent readers by a test in the gate. The prior-art survey's scratch test proved trigram with `remove_diacritics` on v1.59.0, outside the gate; the gate test and the WAL half are still owed (revision 2.5, S finding 7). A standing condition joins it (revision 3.1; study record, section 16): `modernc.org/libc` moves only together with `modernc.org/sqlite`, at the version the sqlite module pins, which is the maintainers' own rule (their issue #177). |
| `ncruces/go-sqlite3` | The named fallback | MIT ([LICENSE](https://github.com/ncruces/go-sqlite3/blob/main/LICENSE); its embedded SQLite build is a separate module under [MIT-0](https://github.com/ncruces/go-sqlite3-wasm/blob/main/LICENSE)) | Well-documented VFS design; a wasm2go translation of SQLite, not a hand-written driver (erratum 2026-09-23, rung 6 spec, erratum 12) | Active: v0.35.6, 2026-09-23; 23 releases in 12 months | 96 importers; gonic's default driver | Record as the fallback if the trigram or WAL check fails. |
| `mewkiz/flac` | Pure-Go FLAC decode; the intake integrity check | Unlicense ([LICENSE](https://github.com/mewkiz/flac/blob/master/LICENSE)) | Idiomatic package split; each frame's CRC-16 is checked and a mismatch is an error | v1.0.14, 2026-08-19, still the newest on 2026-10-07 | 96 importers | Adopt for rung 7. It supplies audio-frame decode, which the repo's own FLAC reader deliberately lacks. |
| Picard tag-mapping tables | Which tag names each container uses | CC0 1.0 ([LICENSE](https://github.com/metabrainz/picard-docs/blob/main/LICENSE) in `metabrainz/picard-docs`; the code repository is GPL-2.0 and is not the source) | n/a | Versioned with Picard; Picard 3.0 shipped 2026-10-04 | The canonical mapping | Adopt. Cite the generated `appendices/tag_mapping.rst` and its [rendered v3.0 page](https://picard-docs.musicbrainz.org/en/v3.0/appendices/tag_mapping.html) at `picard-docs@35f376e`, never the GPL-2.0-or-later generator script (revision 3.1; study record, section 16). |
| `golang.org/x/image`, `gen2brain/webp` | Art resize and encode | BSD-3-Clause ([x/image LICENSE](https://github.com/golang/image/blob/master/LICENSE)) and MIT ([webp LICENSE](https://github.com/gen2brain/webp/blob/main/LICENSE), plus [libwebp's](https://github.com/gen2brain/webp/blob/main/lib/LICENSE.libwebp)) | Exemplar; webp is libwebp transpiled to Go through `wasm2go`, not a wazero runtime, and it loads a host shared library through `purego` first unless built with the `nodynamic` tag | x/image v0.46.0, 2026-09-08; webp v0.6.4, 2026-07-01 | 1,948 and 32 importers on 2026-10-07 | Adopt, with webp built under `nodynamic` so output is identical on every host. The resize filter is a stated choice with a measurement, because the cache pays the encode once (revision 2.5, S finding 7). |
| `golang.org/x/time/rate` | Per-account and per-address auth rate limits | BSD-3-Clause ([LICENSE](https://github.com/golang/time/blob/master/LICENSE)) | Exemplar | Go team; v0.16.0, 2026-08-19 | 16,473 importers on 2026-10-07 | Adopt. The repo's MusicBrainz limiter solves a different problem. |
| `crypto/aes` and `cipher.NewGCM` | App-password encryption at rest | stdlib | n/a | n/a | n/a | Adopt the stdlib; no KDF library is needed. |
| Schema migrations | Rung 6's mechanism | stdlib | n/a | n/a | n/a | Hand-roll with `embed.FS` and `PRAGMA user_version`. `pressly/goose` is declined: it direct-requires ten database drivers for a Go-function migration mechanism no hook asks for, and its decline rests chiefly on its silent-ahead `Up`, which applies every migration newer than the tracked version with no confirmation step (erratum 2026-09-23, rung 6 spec, erratum 12). |
| Stream-through transcode cache | 2.0 rung 5 | stdlib; `djherbis/stream` MIT if adopted ([LICENSE](https://github.com/djherbis/stream/blob/master/LICENSE)) | n/a | n/a | n/a | Hand-roll the size cap and eviction. `djherbis/fscache` is declined: 7 importers, its last tag is v0.10.1 (2020-04-15), and Navidrome pins an untagged pseudo-version. The fill-while-read primitive under it, `djherbis/stream` (MIT), has a tagged v1.5.1 (2026-08-29, still the newest on 2026-10-07; 7 importers); the streaming slice adopts or ports it, decided by a race test (revision 2.5, S finding 7). |
| `internal/atomicfile` | Streaming atomic writes for blobs and cache entries | in-repo | n/a | n/a | n/a | Extend it with a streaming write rather than inventing a second temp-file convention. |
| M3U read and write; Subsonic XML and JSON dual encoding | First-launch import and the response layer | stdlib | n/a | n/a | n/a | Hand-roll both. The study rung enumerated the encoding quirks as rules E1 to E12 in its contract record (`docs/superpowers/research/2026-10-06-slice-1-contract.md`: attribute-heavy XML, the wrapper, singleton-versus-array in JSON, JSONP, and the rest) so the harness tests them. JSONP is declined (revision 3.1; ledger L-01). |
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

These did not block the study rung. Revision 3.1 marks each item the study owned.

- The client roster's exact membership. Geoff ruled it on 2026-10-05: Feishin and Rufin on the Linux desktop, Amperfy on
  iOS and iPadOS, and Tempus on Android with Navic as the alternate. Open source is favored, and a closed client joins
  only if a member uses it daily. The record is
  `docs/superpowers/research/2026-10-05-client-roster-audit.md`.
- Whether folder browsing ships at launch at all. It doubles the browse surface and the id space, modern clients use
  ID3, and only gonic can grade it. Answered pending ruling (revision 3.1; owner fork C; study record, section 12). All
  four roster clients offer a folder view beside the ID3 view and browse ID3 first, and none needs the folder view for
  library use: with no folders served, Tempus hides its sector, Amperfy shows an empty Directories screen, and Feishin
  and Rufin show an empty Folders page. The record recommends no at 0.8, reconsidered at 1.0 if a member asks, because
  the pipeline chooses every path from tags and a folder view would show no organization a member made. Under no,
  `getIndexes` answers an empty index list, `getMusicDirectory` answers error 70, the one-album-directory clauses in
  `download` and shares are moot, and gonic still grades `getMusicFolders`. Under yes, slice 3 builds the folder
  projection and gonic grades it.
- The app-password key's exact name in the age store. The identity slice mints it.
- Moot (revision 3): the shadow period's length and the Navidrome retention window. No shadow or retention exists.
- Whether clients get a report path beside the web app page. The candidate is a reserved per-member playlist named
  "Report a problem": adding a track files a report, which works in every client with no new verb. The slice 7 plan
  decides.
- Whether an admin's fix for a reported album has an edit path by 2.0, or a report resolves by re-upload through
  intake. Rung 13's plan records which one exists.
- Whether smart playlists import Navidrome's `.nsp` files for switchers. Answered (revision 3.1; study record, section
  13): the format is recorded. A `.nsp` file is a JSON rule document with a name, a nestable `all` or `any` rule group,
  sort and limit fields, and a refresh delay; its 17 operators range over about 70 fields, and Navidrome imports the
  files during a library scan and serves them as read-only playlists. dubplate's smart playlists ship in dubplate's own
  rule form and never evaluate `.nsp`. A switcher import parses a file and translates the rules over fields dubplate
  holds, and reports every other rule by name. Slice 8 owns the import.
- Answered (revision 3): whether the study rung runs beside the current ladder. Slice 1 runs now as a document lane.

Revision 2.5 added these, each owed to the study rung. Revision 3.1 marks each below; the credential bullet is routed
and stays open. The sources state the facts and don't make the decision.

- The XML validation rule: a legacy XML form against the official XSD, or a derived XSD that admits the OpenSubsonic
  attributes ("Evidence"; S finding 1). Answered (revision 3.1; study record, section 5): the legacy form, validated
  against the unmodified XSD after a scoped projection.
- The `transcoding` extension: whether 0.8 answers `getTranscodeDecision` and `getTranscodeStream`, and how a posted
  `ClientInfo` interacts with the device profile ("Streaming and transcoding"; S finding 2). Answered pending ruling
  (revision 3.1; owner fork A; study record, section 6): the recommendation is no at 0.8, revisited at 1.0 with the
  format ladder.
- A Range request on a filling transcode: a declared divergence class, or `Accept-Ranges: none` as the oracle does,
  with seeking on `timeOffset` ("Streaming and transcoding"; S finding 4). Answered (revision 3.1; study record, section
  7): `Accept-Ranges: none` as the oracle does, with seeking on `timeOffset`.
- The CORS policy: allowed origins and headers, and whether credentials are ever allowed ("API surface"; S finding 5).
  Answered (revision 3.1; study record, section 8): any origin, no credentials, a bodyless preflight, and nothing
  outside the Subsonic sub-prefix.
- The AAC container for stream-through: fragmented MP4 or ADTS ("Streaming and transcoding"; S finding 6). Answered
  (revision 3.1; study record, section 9): ADTS, with a playback test owed.
- The protocol module's license, MIT or Apache-2.0, and what NOTICE obligation it carries from the Apache-2.0 OpenAPI
  file it validates against ("The protocol layer is a separable module"; G). Answered pending ruling (revision 3.1;
  owner fork B; study record, section 10): the recommendation is MIT, with the checker loading the OpenAPI file as
  caller input and never vendoring it. The repository carries no NOTICE file today.
- Which credential a stream authenticates with before slice 6 builds app passwords, owed to the slice 5 plan
  ("Served volume"; P conflict 3). Routed, not answered (revision 3.1): the study does not answer it, and it stays open
  for the slice 5 plan.
- The word "collection". OpenSubsonic PR #237, "Add collections" (open, 51 comments, updated 2026-09-20), defines a
  collection as a user-owned ordered list of items, while dubplate's collection maps to a music folder. If the PR
  lands, API docs and UI copy need a distinct word for one of them. PR #162 (open) would extend `musicFolderId` to
  more endpoints, which bears on "honored per endpoint" under "Catalog projection". The study rung tracked both, with
  PRs #258 and #261 (S finding 8). Answered (revision 3.1; study record, section 11): on 2026-10-07 none of #162, #237,
  #258, or #261 had merged, and the specification's `main` was still `bed1688`. dubplate never declares the
  `collections` extension, because #237's collection is a per-user ordered list with a `public` flag and a surface
  offering a "public collection" reads as a refused feature. API-facing prose says "music folder" for the wire concept
  and "collection" only for the store concept, so a merged #237 changes no wire text. dubplate never declares
  `publicImageUrl` (#261). The direction of PR #202 is adopted for `scrobble`, and PR #278 is slice 6's input. Renaming
  dubplate's partition to "library" becomes a taste question only if #237 merges before launch and a roster client shows
  both concepts under one label. Slice 2 re-reads all four PRs and #278 before it pins a specification build.
