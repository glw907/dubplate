# Group-scale principles spec: mechanics and feasibility review

Target: `docs/specs/2026-09-25-group-scale-principles-design.md` at commit
`1d6bc6a` (revision 1, pre-review). Lens: does every mechanism behave as the spec states?
Reviewer: Opus 5.5, 2026-09-25. Findings are ranked by consequence. Each carries a severity, a
location in the target spec, the evidence, and a proposed fold. Product choices are marked
OWNER FORK and are not ruled here.

Counts: 1 blocker, 4 major, 12 minor, 2 owner forks.

Most of the spec holds up. The citations to the streamer, configuration homes, rung 5, rung 6,
and organizer specs say what the spec claims, with the exceptions below. The share erratum is
safe for clients, because clients already cope with Navidrome's sharing-off mode. App password
per device plus a device cap fits both OpenSubsonic auth models. The defects are concentrated
in one place: how bulk copying is metered, and what the spec tells readers about it.

## Evidence gathered

- **OpenSubsonic spec** (`opensubsonic/open-subsonic-api`, HEAD `bed1688`, 2026-08-07).
  - `content/en/docs/Responses/error.md` lists the error codes: 0 generic, 10, 20, 30, 40 to
    44 (auth), 50 "User is not authorized for the given operation", 60, and 70. No code means
    "rate limited", and the spec never mentions HTTP 429.
  - `Endpoints/download.md`: "Downloads a given media file ... Returns binary data on
    success, or an XML document on error (in which case the HTTP content type will start
    with 'text/xml')." Album, artist, and playlist zips are a Navidrome behavior. The spec
    does not define them.
  - `Responses/user.md` defines `downloadRole` and `shareRole` as required booleans.
  - `Extensions/apiKeyAuth.md`: "the server may provide a page where the user can configure
    one or more API keys"; "Servers ... must provide some mechanism for viewing active API
    key(s) and allow for revoking"; "these API keys do not expire". A request with `apiKey`
    must not carry `u`.
  - `api-reference.md`: "token = md5(password + salt)", with `u` naming the user.
- **Navidrome** (`navidrome/navidrome` master, fetched 2026-09-25).
  - `server/subsonic/api.go:207-217`: when `EnableSharing` is false, the four share endpoints
    are registered through `h501`. That handler returns HTTP 501 with a plain-text body, not a
    `subsonic-response`.
  - `server/subsonic/users.go`: `ShareRole: conf.Server.EnableSharing` and
    `DownloadRole: conf.Server.EnableDownloads`.
  - `server/subsonic/stream.go:64-66`: disabled downloads answer
    `newError(responses.ErrorAuthorizationFail, "downloads are disabled")`, which is code 50.
  - `stream.go:119-127`: album, artist, and playlist ids each return a zip.
  - `conf/configuration.go:994,1035` default both settings to true today. Sharing shipped
    disabled by default in 0.49.0, so clients have faced a sharing-off server for years.
- **Tempus** (`eddyizm/tempus`, `4d2b061`, 2026-09-19).
  - `util/MappingUtil.java:282-286` and `util/MusicUtil.java:114-160`: offline downloads fetch
    one track at a time, through `download?id=<song>` or through `stream` ("transcoded
    download"). Tempus never fetches an album zip.
  - `repository/SharingRepository.java:18-40`: `getShares` failures are swallowed and the
    list stays empty.
  - `util/Preferences.kt:800-802`: share UI is gated on a local preference that defaults to
    false.
- **DSub** (`daneren2005/Subsonic`, `449a6dc`).
  - `service/RESTMusicService.java:818-866`: the offline download (`getDownloadInputStream`)
    calls `stream`, not `download`. When the response content type is `text/xml` or
    `text/html`, DSub parses it as an error and the task fails.
  - `service/DownloadFile.java:505-512`: a 404 marks the download failed past `MAX_FAILURES`,
    with no retry.
- **Feishin** (`jeffvli/feishin`, `6c5173f`, 2026-09-24).
  - `features/context-menu/menus/album-context-menu.tsx:40` and `artist-context-menu.tsx:38`
    both call `DownloadAction`, which opens `rest/download.view?id=<albumId|artistId>`
    (`api/subsonic/subsonic-controller.ts:976-980`). Feishin is the one client in the roster
    that requests artist zips.
  - Feishin's share action runs only through the Navidrome-native controller
    (`api/navidrome/navidrome-api.ts:245`), not the Subsonic share endpoints.
- **Symfonium** (closed source; its docs are the evidence).
  - [Provider configuration](https://docs.symfonium.app/wiki/providers/subsonic-opensubsonic-media-provider-configuration/)
    says: "Symfonium uses the download endpoint, which is usually more stable. However, some
    servers may have this option disabled. If you encounter this issue, please check this
    option" (the per-provider "Do not use download endpoint" toggle).
  - The same page describes "Automatic media offline cache": "Enabling this option will
    cache all media files for offline usage after sync".
  - Failed cache downloads show "Waiting to retry"
    ([forum 8073](https://support.symfonium.app/t/issue-with-downloading-to-offline-cache/8073)).
    The retry cadence is unpublished.
  - Subsonic share links arrived as "Share file" in 15.x
    ([forum 1377](https://support.symfonium.app/t/create-share-links-on-subsonic-servers-that-can-make-them/1377)).
  - substreamer and play:Sub are closed source and publish nothing on endpoint use. Their
    behavior is unverified and belongs to the study rung's roster walk.
- **MusicBrainz**
  - [The release-URL relationship types](https://musicbrainz.org/relationships/release-url)
    include "purchase for download", "purchase for mail-order", "download for free", "free
    streaming" ("such as Spotify"), and "streaming" ("for a subscription fee, such as Tidal").
  - Sample: 45 release MBIDs drawn at random from `corpus/`. 28 resolved; 16 were not found
    (synthetic corpus ids) and 1 was busy. Each resolved release was fetched with
    `inc=url-rels+artist-credits+release-groups+release-group-level-rels`, and each credited
    artist with `inc=url-rels`.
  - Results: 3 of 28 releases (11%) carry any buy or stream link, and 2 of 28 carry a
    purchase link. Release groups carry none (their URL types are allmusic, discogs, wikidata,
    review, and similar). Every one of the 27 credited artists carries "free streaming" or
    "purchase for download".
  - Four release groups whose own release had no links (Master of Puppets, Jolene, At Folsom
    Prison, The Velvet Underground) were browsed with `release?release-group=<id>&inc=url-rels`.
    Each has 3 to 6 sibling releases carrying buy or stream links.
  - Caveat: the corpus is weighted toward well-known albums, so obscure releases will do
    worse.
  - The matcher's lookup already requests `url-rels` (`internal/engine/musicbrainz/query.go:28`).

## Blocker

### B1. The anti-bulk mechanism meters a path clients do not use, and the "isn't" copy states a false fact

Location: lines 43-45, 126-129, and 269-273 (ruling 3's "The limit counts album zips only").

**Defect.** The spec claims three things:

- "Members stream one track at a time and download one album at a time, and whole-library
  copies are backups that stay with the person who runs the server."
- The per-member album download limit is the "no bulk export" mechanism.
- The limit counts album zips only.

The evidence shows otherwise:

- No offline-sync client in the roster fetches album zips.
  - Tempus downloads per track, through `download` or `stream`.
  - DSub's offline download calls `stream`.
  - Symfonium caches per track through `download` by default, and it ships a user toggle
    that moves to `stream` when a server refuses `download`.
  - Only Feishin's desktop context menu requests an album or artist zip.
- One Symfonium toggle, "Automatic media offline cache", copies every media file in the
  library to the member's phone after each sync. Every track goes through the unmetered
  per-track path.

As written, the 20-per-day zip limit binds almost nothing. A whole-library copy to a member
device is one checkbox away in the roster's leading client. The "isn't" bullet was drafted
for reuse on the website, and it would publish a mechanical claim that the stock product and
stock clients falsify. The ratified checkout verb (M1) falsifies it a second time.

The opposite fix also fails. Applying a hard limit to per-track `download` would push
Symfonium users to the "Do not use download endpoint" toggle and leave queued downloads at
"Waiting to retry". That penalizes ordinary offline listening, which is the product's
intended use, and it still meters nothing.

**Proposed fold.**

1. Rewrite the "isn't" bullet and the "No bulk export" paragraph to state what holds:
   - dubplate has no endpoint that returns an artist's catalog, a playlist, or the library as
     one archive.
   - Every byte a member's devices fetch is counted against that member and that device.
   - An unusual volume is visible to the admin.
   - Drop "one album at a time".
2. Replace the album-zip counter with a per-member and per-device meter over everything
   served (`stream`, `download`, and zip), measured in distinct tracks and bytes over a
   rolling window. A threshold notifies the admin (the configuration homes spec's
   notification channel), and the member's page shows the member's own count.
3. If the owner keeps a hard stop (fork F1), it refuses with a well-formed error code 50 in
   the requested format (`text/xml` or JSON, per `download.md`), with a message and a
   `helpUrl`. A bare HTTP 429 is out, because DSub treats only an XML or HTML content type
   as an error, and Symfonium's retry on a refusal is unmeasured. The study rung's roster
   walk must record each client's behavior on code 50 for `stream` and for `download`
   before any hard stop ships.
4. Keep the artist and playlist zip removal. Feishin's artist download then gets an error
   body in a browser tab, which degrades gracefully.

**OWNER FORK F1: is a member's full-library offline cache on their own devices intended use?**

- (a) Yes. Meter it and surface it, and refuse nothing.
- (b) Yes, up to a per-member served-volume ceiling, then a code 50 refusal until the window
  rolls.
- (c) No. Refuse library-scale offline caching.

Recommendation: (a). For a family, a phone holding the shared library is the ordinary use,
and Symfonium's auto-cache is how the roster's leading client provides it. A ceiling (b)
fires on exactly that member. Option (c) cannot be enforced, because a per-track fetch cannot
be told apart from playback. The owner's framing is "easy for its intended purpose, hard for
nefarious ones", and the meter plus admin notice already makes the nefarious case visible.

## Major

### M1. The ratified checkout verb is a member-side whole-collection copy, and the conflicts list omits it

Location: lines 43-45, 126, and 231-258 (the "Conflicts" list and "No spec proposes ... any
way to fetch music").

**Defect.** The configuration homes spec's **Check out** verb "takes a read-only offline copy
of a server collection" (`2026-09-11-configuration-homes-design.md:498-502`, and again at
633). A member "asks for one directly". ROADMAP line 63 schedules the "read-only checkout" with
the TUI, and the organizer spec carries an erratum admitting it (config homes :728-729).

That verb is the member-held bulk export this spec says doesn't exist: "whole-library copies
are backups that stay with the person who runs the server". The streamer spec's "Digital
audio players" section (:404-409) also names `download` as the server's contribution to
device sync, and it puts "a DAP export profile in the local binary" on the 2.1 list. Neither
passage is listed as a conflict. The spec's statement that only three passages conflict is
therefore wrong.

**Proposed fold.** Add a fourth conflict entry that names the checkout verb and the DAP
export. Route both to the same meter as B1: a checkout is served bytes under the member's
device credential, counted and shown like any other. The spec must also say whether checkout
survives (fork F2).

**OWNER FORK F2: does the member checkout verb survive these principles?**

- (a) It survives, metered under B1's meter, with its first use per member announced to the
  admin.
- (b) It is admin-approved per member, like the first mirror.
- (c) It is dropped, and a member's offline copy comes only from a client's cache.

Recommendation: (a). A checkout is the TUI's version of Symfonium's offline cache (B1), and
the configuration homes spec built it for a member's own laptop.

### M2. `added_by` misses the one bulk entry path, the laptop mirror

Location: lines 139-151 and table row 219 ("`added_by` on every admitted album").

**Defect.** The spec relies on the streamer spec's rule that intake sets `added_by` "and is
the only channel that does" (`streamer-2-0-design.md:323-325`, hook 17). The configuration
homes spec's mirror is a separate transport that "bypass[es] per-album intake verdicts"
(accepted risks, :653-654). Mirrored albums therefore reach the album row outside intake, so
under the cited rule their `added_by` is either never set or set by a channel the rule
forbids.

The provenance record is blank exactly where one act adds the most albums. The same gap
applies to a seize and a hand-over, both of which transfer rows without intake.

**Proposed fold.** State that the mirror push sets `added_by` server-side from the member bound
to the device credential. The configuration homes spec already binds that credential to "the
surrogate user id and a device id" (:571-576) and never reads a client-asserted field. Amend
hook 17's "only intake sets it" to "only the authenticated principal of intake or of a mirror
push sets it" in the streamer erratum. Say what a seize or hand-over does to `added_by`: it
preserves the value and never resets it to the admin.

### M3. Per-device volume and bytes served are not derivable from the event shape the spec reuses

Location: lines 153-158 and table rows 221-222.

**Defect.** The spec says "Stream starts are already an event kind in play history (forward
hook 4). This spec adds download events of the same shape." That shape is "user, song,
client, timestamp, event kind, and how reported" (`streamer-2-0-design.md:299-302`).

- `client` is the Subsonic `c` parameter. The client asserts it, and it collides across
  devices: two phones running Symfonium send one value.
- The row has no device or app-password id, so the per-device view promised on line 154 and
  in ruling 4 cannot be computed.
- The row records a start, not bytes, so "bytes served" (line 155) has no source.
- A start is undefined under Range requests. DSub resumes with `Range: bytes=<offset>-`
  (`RESTMusicService.java:847-850`), and clients retry, so a naive count per request
  inflates both streams and downloads.

**Proposed fold.** State the event shape this spec needs:

- the authenticating credential's id (the app password row or the device credential), set
  from auth and never from `c`;
- bytes sent, recorded at response completion or as a per-credential counter;
- a dedupe rule for a start (for example, one start per credential, track, and window, with
  a Range continuation counted as bytes only).

Play history is not yet built (rung 6 lands only hook 4's invariant, per
`2026-09-23-rung-6-ops-seam-store.md:448`), so the fold costs a column now against a
migration later.

### M4. MusicBrainz release relationships leave about nine album pages in ten without links

Location: lines 33-34 and 162-165, and table row 223.

**Defect.** The spec reads "relationships from a release to the places that sell or stream
it" for "each admitted release". The relationship types exist, but the sample shows they
rarely sit on the release a member owns:

- 3 of 28 sampled releases (11%) carry any buy or stream link.
- 0 release groups carry one.
- All 27 credited artists do.
- Each sampled release group has 3 to 6 sibling releases with links.

Store links attach to the digital-media release. A member's CD or vinyl rip matches a
physical release, which usually carries none. As specified, the feature renders an empty
section on most album pages. The capture itself is free, because the matcher's lookup already
includes `url-rels` (`internal/engine/musicbrainz/query.go:28`).

**Proposed fold.** Specify a fallback order:

1. the admitted release's links;
2. links on sibling releases in the release group, fetched through one browse call
   (`release?release-group=<id>&inc=url-rels`) under the shared MusicBrainz rate limiter;
3. the artist's "purchase for download", "purchase for mail-order", and Bandcamp links,
   labeled as the artist's rather than this album's.

Record that links are captured at admission and can go stale, and that the refresh cadence
belongs to the pass that builds the feature under ruling 5. The recommended order (album
links first, artist links as a labeled fallback) stays a product choice for that pass.

## Minor

**m1. Answer share endpoints with code 50 and report `shareRole=false`, not code 0 (lines
241-243).** Code 50 ("not authorized for the given operation") describes a deliberate
refusal. `getUser`'s required `shareRole` flag is the one signal a client can read before it
calls. Navidrome sets `shareRole` from `EnableSharing`, and its sharing-off mode answers with
a bare HTTP 501, which Tempus and DSub survive (Tempus swallows the failure). A well-formed
code 50 is strictly gentler. Keep the study rung's empty-success escape.

**m2. The share erratum's inventory is incomplete (lines 234-245).** Four passages also
retire:

- the streamer spec's hook 15 ("with share rows reserved beside visibility", :507);
- the API surface's launch list ("... play queue, shares, and read-only user", :384);
- revision 2's summary item 12 ("share management", :42);
- the `createShare` entry in configuration homes' `visibleCollections` test list (:705).

**m3. The roster is no longer file-backed (lines 67-70).** Rung 6d moves it into the store
through `member import`, and `DUBPLATE_ROSTER_FILE` retires
(`2026-09-23-rung-6-ops-seam-store.md` ruling 1; the configuration homes erratum at :61-64).
Cite the store-backed roster. The cap's add path then needs to state that it covers the CLI
verb, the add-member page, `member import`, and `dubplate import navidrome`, or it must name
any exemption. The founding roster's 6d import predates the rung 14 cap.

**m4. The Navidrome import at the cap must also skip the skipped members' data (lines
80-82).** Stars, playlists, and play history of a member who was never added have no owner.
Say they are skipped and reported together, and say which members are admitted first (the
Navidrome admin, then creation order), so the skip set is deterministic.

**m5. The device cap counts credentials, not devices (lines 101-110).** One app password can
be pasted into several devices, and the configuration homes spec mints "a device credential
per connection profile, so two profiles on one laptop are two writers" (:569-571). Call it a
per-member credential cap. The admin list's last-seen address is how shared credentials show
up. On the mechanics, the cap fits both auth models:

- Under token and salt, the server identifies the device by which of the member's secrets
  satisfies `md5(secret + salt)`. That costs one AES-GCM decrypt and one MD5 per credential
  per request (or a cache of plaintext in memory), so the cap also bounds per-request
  verification work.
- Under `apiKey` (no `u`), the key maps straight to its row.

Nothing in either model blocks the design.

**m6. The mirror's source answer has no approval moment after the first mirror (line
143-144).** The admin approves "each member's first mirror", and "after that the member
mirrors freely within quota" (configuration homes :621-623). A member's second mirrored
collection never passes an approval step. Ask the question when a mirror is created, per
collection, from the TUI.

**m7. The source answer must be an intake API field (lines 139-145).** The TUI's
**contribute** verb pushes through intake with no uploader page (configuration homes
:612-615). If ruling 2 makes the question required, intake enforces it and the page only
renders it. Otherwise every non-page intake client skips it.

**m8. State the refusal wire shape for the album limit and the device cap.** The API refusal
is code 50 in the requested format with a `helpUrl` (see B1). The device-cap refusal lives on
the web device page and never reaches a Subsonic client. Say both.

**m9. "Without becoming a listening log" overstates it (lines 155-157).** The view shows
counts, but play history is a per-member, per-song record that the streamer builds anyway
for recently played and most played. An admin with shell or backup access sees titles. Ruling
4's disclosure to members should say that the admin sees counts in the app, and that
per-song history exists on the server.

**m10. "Planned" for the 0.9 local client overstates it (line 108 and table row 212).**
ROADMAP (:113-116) says 0.9's "scope is unspecified; nothing else about it is settled yet".
The device credential is ratified, but its 0.9 use is not. Mark the row "new (routes to
0.9's spec)".

**m11. "The 2.1 list" is a stale name after the re-key (lines 172-173 and table row 225).**
The streamer spec's header maps 2.0 to 0.8. Say which release the list now targets, or cite
it as "the streamer spec's post-launch list".

**m12. "Never fetches" is a policy, not a mechanism, at the API edge (lines 114-121).** The
documented intake API, which is the organizer's "documented client API for anything
tool-shaped" (organizer :349), accepts uploads from any script that holds a member
credential. The organizer's lifecycle hooks run operator-configured external commands at
post-import and on-reject (:344-348). Either path can host an external fetcher without a line
of dubplate code. The spec's refusal stands as a project policy, and the text should say so
in one sentence. If the owner wants the upstream analog of B1's meter, it is a per-member
intake volume counter shown in the same admin view.

## Citations verified as stated

- Rung 5, "Fork 1" and "The roster" (:178-200, :1118-1140): an admin-only add-member page,
  and "The web never grants the admin flag to a first login". This holds except for the
  store move in m3.
- Rung 5 "Fork 4" (:235-252): the kind answer names the admin contact.
- Configuration homes, "The server's own knobs" (:143-148): a sensitive write requires fresh
  re-authentication and notifies the admin address.
- Configuration homes: sharing outside the roster is declined, with share rows reserved as a
  hook (:560-566), and "shares re-checked at fetch time" (:706).
- The streamer spec's "Shares" section (:289-295), the API surface (:384-386), slice 6
  (:442), and the web surface (:395) are at the cited lines.
- `getOpenSubsonicExtensions` is the one unauthenticated carve-out (:380).
- The app password is minted per device and revocable alone, with the valid-after epoch
  (:83-89, :280-281).
- The Navidrome import uses the roster's normal add path (:350-352).
- `getAlbumInfo2` `notes` carries the member's note (:326-327).
- The first-party web player is on the 2.1 list (:564-566), which is consistent with ruling 5.
- Rung 6: restore is offline, with no HTTP route (:83-85, :164).
- Organizer: the plugin ABI is declined (:338-341), and scrapers are refused on fragility
  grounds (:366).
- OpenSubsonic `public` on a playlist: "`true` if the playlist should be visible to all
  users" (`updateplaylist.md:21`).
