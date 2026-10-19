# Group-scale principles spec: consistency review

Target: `docs/superpowers/specs/2026-09-25-group-scale-principles-design.md` at `1d6bc6a`
(revision 1, pre-review). Lens: does the draft contradict a ratified document, and does each
citation say what the draft claims? Reviewed against the build spec, the streamer spec (0.8),
configuration homes, rung 5, rung 6, the visual identity spec, the organizer spec, the rulings
ledger, ROADMAP, IDEAS, CLAUDE.md, LICENSE, and the approved 5c-iii and 6b-ii plans.

Findings are ranked by consequence. Line numbers on the draft side are the file at `1d6bc6a`;
"GS" below means the draft.

**Counts: 1 blocker, 3 major, 13 minor.**

## What checks out

The draft's three named conflicts are real and correctly located. Streamer "Shares" is at
`streamer-2-0-design.md:289-295`, slice 6 is named "Identity, app passwords, and shares" at
`:442`, share management sits in "Web surface" at `:395`, and `download`'s album, artist, and
playlist zips are at `:384-386`. Configuration homes declines sharing outside the roster and
reserves share rows "if sharing ever ships" (`configuration-homes-design.md:561-566`), and its
projector hook says "shares re-checked at fetch time" (`:706`). The device-limits item is
correctly called an addition: the streamer's identity section (`:269-287`) rate-limits failed
authentication and sets no cap.

These citations also hold: rung 5's fork 1, the roster's add paths, the admin flag the web never
grants, and fork 4's kind answer (`rung-5-identity-auth.md:178-199`, `:235-251`). The Navidrome
import invites through the roster's normal add path (`streamer:350-352`). `getOpenSubsonicExtensions`
is the one unauthenticated API carve-out (`streamer:380`). `added_by` is set only by intake from
the session (`streamer:323-325`, hook 17 at `:515-517`). Album notes reach clients through
`getAlbumInfo2` (`:325-327`). Scrobbling is on "The 2.1 list" (`:562-563`). The organizer spec
declines a plugin ABI (`organizer:338-341`), and its scraper refusal rests on maintenance
grounds (`:366-370`). Restore is offline and CLI-only, and backup is CLI or HTTP behind rung 14
(`rung-6-ops-seam-store.md:936-937`), consistent with the 6b-ii plan (`:38-50`). LICENSE is MIT,
matching `streamer:183-189`. The "group" and "member" terms are used consistently. "Family"
appears only as an example of a group, as the group-not-family ruling asks.

## Blocker

### B1. The configuration homes checkout verb is a member-initiated bulk export, and the draft does not list it

- **GS:** `:43-45` ("whole-library copies are backups that stay with the person who runs the
  server"), `:126` ("No bulk export over the API"), `:231-232` ("Three passages in ratified specs
  conflict").
- **Ratified side:** `configuration-homes-design.md:498-502` ("**Check out** takes a read-only
  offline copy of a server collection ... A member asks for one directly"), `:633` (check out as
  an attach verb), `:728-730` (Supersessions: the checkout "is scheduled with the TUI"), and
  `ROADMAP.md:62-63` ("the read-only checkout [is] in").
- **Defect:** Any member may check out any server collection they can see, and the group library
  is an admin-owned server collection (`configuration-homes-design.md:206`). One checkout copies
  the whole group library to a member's laptop. That falsifies the draft's headline "isn't" line
  and its "no bulk export" rule. The draft's conflict list claims completeness at three items and
  omits this one, so approving the draft as written ratifies two specs that disagree on its core
  claim. A related path follows from the same verb. A member holding connection profiles for two
  servers (`:155-162`) can check out from one and contribute to the other through intake, which
  relays a library between servers through a laptop. GS `:96-99` says only that attach and mirror
  "don't connect two servers".
- **Note on the mirror:** the laptop-to-server mirror (`:618-623`) moves music toward the server
  and does not conflict with "no bulk export". It conflicts elsewhere (M2).
- **OWNER FORK: what checkout may copy.**
  - (a) A member checks out only collections they own. The hand-over and seize demotions
    (`:500-502`, `:508-509`) already produce exactly that case, so they keep working unchanged.
  - (b) A member may check out others' collections only with admin approval per collection, and
    the volume view counts the transfer.
  - (c) Keep checkout as ratified, and rewrite the principle to "bulk copies to members are
    admin-visible and rate-limited".
  - **Recommendation: (a).** It keeps every verb the attach flow needs. It also matches the
    draft's line that whole-library copies stay with the operator. Option (c) makes the "isn't"
    list untrue as website copy.
- **Fold:** add the checkout as conflict 4, with the chosen option. Add one sentence under
  "Servers don't federate" naming the laptop relay and stating that it is bounded by the checkout
  rule and by intake.
- **Erratum needed:** `configuration-homes-design.md:498-502` and `:633`, the organizer spec's
  2026-09-12 erratum that names the checkout as the server-to-local mirror, and
  `ROADMAP.md:62-63`.

## Major

### M1. The launch estate streams through Navidrome, where several of the draft's present-tense claims do not hold

- **GS:** `:38-45` and `:85-91` (no public links, per-album downloads, and login for all audio
  through rung 5), the table at `:205` ("Login required for all audio ... Rung 5 auth | exists"),
  and `:101-110` (device caps).
- **Ratified side:**
  - `ROADMAP.md:28-29`: the ladder's launch is a 0.x pre-release the group uses.
  - `ROADMAP.md:97`: "Navidrome remains the escape hatch through 0.8's first months."
  - Build spec `:849-870`: rung 12's Navidrome reference driver.
  - Build spec `:479-481`: serve strips identity headers on every path, "the `/rest/*` and
    `/share/*` bypasses included".
  - Build spec `:560-563`: "`/rest/*` is recorded as ... an anonymous surface reachable through
    serve whose authentication is the streamer's own."
  - Rulings ledger `:88`, `:109`.
- **Defect:** From launch until the flip at 0.8's slice 9, and afterward whenever the escape
  hatch is used, audio is served by Navidrome. The draft's mechanisms are all scoped to rungs and
  0.8, and they are silent about this period.
  - Navidrome's `download` returns artist and playlist zips. Downloads are on by default in
    Navidrome.
  - `/share/*` public links exist whenever Navidrome sharing is enabled. Sharing is off by default
    in Navidrome. Verify both defaults against the pinned 0.63.2 at rung 12.
  - Navidrome has one Subsonic password per user and no device cap.
  - Rung 5 does not gate `/rest/*`.

  The table's "exists" at `:205` is therefore wrong for every release before 0.8. The "isn't"
  list, written for use as website copy, does not hold for the release the group first runs.
- **Fold:** add a short "Before 0.8" subsection.
  - Rung 12's driver pins Navidrome's sharing off in its provisioning choreography and adds a
    `doctor` row that proves it.
  - The build spec's `/share/*` bypass is recorded as inert while sharing is off.
  - The subsection states the download and device residuals until 0.8.
  - Date the "is and isn't" copy to 0.8, or add a line for the 0.x estate.
- **OWNER FORK: Navidrome downloads during 0.x.**
  - (a) Turn Navidrome downloads off. Some clients' offline sync uses `download`, so this may
    break offline listening.
  - (b) Leave downloads on and state the residual until 0.8.
  - **Recommendation: (b).** The 0.x group is the founding group. Breaking offline sync costs
    more than the residual.
- **Erratum needed:** build spec `:481` (the `/share/*` bypass is conditional on the driver's
  sharing setting). Route a rung 12 plan item for the pin and the `doctor` row.

### M2. The draft leaves out two admission channels, and `added_by` and the source answer cannot cover them as written

- **GS:** `:114-120` ("Music enters dubplate only when ... The entry paths are the uploader ...
  the laptop mirror ... and the one-time migration"; "the documented API admits music only
  through intake"), `:139-151` (source question and `added_by`), the table at `:219`
  ("`added_by` on every admitted album").
- **Ratified side:**
  - Timer-path imports: organizer `:545-547`, `:617`, and `:651-653` ("the admin-only side door
    for timer-path material"), and build spec `:808-809` and `:879-880`.
  - The mirror is not intake: configuration homes `:621-623` ("a mirror consumes disk with no
    intake verdict per album") and `:653-654` (accepted risk: "Mirrors bypass per-album intake
    verdicts"). The push is still re-validated (`:578-585`).
  - Copy bypasses intake: `:466`.
  - Hook 17 says only intake sets `added_by`: `streamer:323-324`, `:515-517`.
  - Lifecycle hooks run configured external commands at pipeline events: organizer `:343-348`,
    build spec `:844-849`.
- **Defect:** The entry-path list omits the timer path, which is a ratified launch-bar intake
  channel. That path is also where an operator would point a download client's output folder.
  The claim that the API admits music "only through intake" contradicts the mirror, which is a
  documented, versioned API that bypasses per-album intake verdicts. Hook 17 lets only intake
  set `added_by`, so mirrored albums and timer-path albums would carry null under the rule as
  ratified. Copy's new album row has no stated `added_by` either. The table's "every admitted
  album" is therefore not true. CLAUDE.md's planning rule applies here: each privileged field
  names every channel that may set it.
- **Fold:** replace the entry-path sentence with a per-channel list: uploader, TUI contribute
  (through intake), timer path (admin-only), mirror push, copy, the rung 16 migration, and
  restore (the operator's own data). For each channel, name who sets `added_by` and the source.
  Suggested settings:
  - Mirror: the device credential's member, set by the server.
  - Timer path: the admin who placed the files, or an "admin import" source.
  - Copy: carry the source row's value.

  Reword `:119-120` to say that every channel re-validates at the server and that none fetches.
  Add one sentence that the project ships no lifecycle-hook recipe that fetches music, since
  hooks are the sanctioned external-command seam.
- **Erratum needed:** streamer hook 17 (`:515-517`) and "Added by" (`:323-324`), to name the
  mirror channel's setter. If the owner wants timer-path albums attributed, the timer-path
  definition in organizer `:651` needs one too.

### M3. Ruling 5 rests on a false premise: the web app has album pages at rung 7

- **GS:** `:279-284` ("the first release that has an album page in the web app, which is the
  first-party web player after 0.8"), and `:33-34` (the "is" list says album pages link to
  places to buy, in the present tense).
- **Ratified side:** `2026-09-24-visual-identity-design.md:338-356`. Rung 7's scheduled design
  review includes "The album page: the large sleeve and record, the tracklist, the catalog line,
  and a small inscription ... naming who added the album", plus the library grid, the list view,
  and filtering (`:344-350`).
- **Defect:** The web app renders album pages from rung 7, which is before launch and far before
  0.8. The recommendation's reasoning (in 0.8 "the only album text dubplate controls there is the
  `notes` field") ignores that page. The first consumer under the hook-by-first-consumer rule may
  be rung 7 or rung 9. This decides where the MusicBrainz URL relationships are captured.
  Admission at rung 9 is the natural place, since that is where the release lookup already runs.
- **OWNER FORK: when buy links ship (ruling 5), re-posed with the correct premise.**
  - (a) Rung 7's album page renders them. Rung 9 captures the relationships at admission (a new
    forward hook).
  - (b) A later release adds them to that page.
  - **Recommendation: (a), if a rung 7 plan task can absorb it without growing the pass.**
    Otherwise take (b) and route the capture hook to whichever release renders the links.
- **Fold:** rewrite ruling 5 and the table row at `:223`. Change the "is" list at `:33-34` to
  the future tense, or drop the buy-link and wishlist claims until they are scheduled.

## Minor

1. **The shares erratum misses four reserved-share passages.** GS `:241-245` retires the
   reserved share rows and the "re-checked at fetch time" line in configuration homes only. The
   same reservation also appears in these places:
   - streamer hook 15 (`:507`, "with share rows reserved beside visibility");
   - configuration homes `:675` (its rung 6 routing list);
   - rung 6 spec `:485` ("Share rows are reserved in this spec only");
   - `createShare` in the projector's test list (configuration homes `:705`).

   Streamer Revision 2 item 12 (`:42`) and the API surface's launch list ("shares", `:384`) also
   name shares. Fold: list every passage in conflict 1.
2. **The download erratum also touches the streamer's digital audio players section.**
   `streamer:406-409` says the server's contribution to device sync "is `download`, specified
   above". The playlist zip carries an `.m3u8` "matching Navidrome" (`:385-386`). Narrowing
   downloads changes that sentence and the 2.1 DAP export premise (`:566-567`). Fold: name
   `:406-409` in conflict 2 and state that DAP loading goes per album or through the local
   binary's export profile.
3. **One device cap spans two credential tables.** GS `:108-110` counts the local client in the
   same cap as app passwords. The configuration homes device credential lives in its own table
   (hook 15, `streamer:504`) and is issued per connection profile (`configuration-homes:570-571`),
   so one laptop with two profiles counts twice. Configuration homes also already gives members a
   list of their own device credentials with last-seen address, expiry, and one-step revoke
   (`:170-172`). Fold: say that the cap counts app passwords plus device credentials, one each
   per profile, and cite `:170-172` as the member-side list the admin list extends.
4. **The draft settles the scope of 0.9, which ROADMAP leaves open.** `ROADMAP.md:113-116` says
   the 0.9 paired local client's "Scope is unspecified; nothing else about it is settled yet."
   GS `:108`, `:115-116`, `:212`, and `:218` attach the device credential, the mirror, and the
   mirror source answer to 0.9. The table marks `:212` "planned", but the draft defines
   "planned" as scheduled by a ratified spec (`:198-199`), and configuration homes schedules
   these with "the TUI, unscheduled on the ladder" (`:709-710`). Fold: write "the local client
   (the TUI), unscheduled", and mark the status "designed, unscheduled".
5. **The "2.1 list" has no mapping to the new version numbers.** The streamer header re-keys
   2.0 to 0.8 (`:3-4`), but nothing re-keys "2.1" now that 0.9 is the local client and 1.0 is
   polish plus the format ladder (`ROADMAP.md:113-122`). GS `:172-174`, `:225`, and ruling 5
   inherit the ambiguity. Fold: say "after 0.8, on the streamer spec's post-0.8 list" and route
   the naming to the streamer spec's next revision (STATUS already carries its errata queue).
6. **The roster description is stale.** GS `:67-70` says "the roster is a file-backed store".
   Rung 6 erratum 5 (`configuration-homes:61-63`) moves the roster into the store at 6d, and the
   file becomes an import source. Fold: "the roster (a file at rung 5, the store from 6d)".
7. **Two citations point at the wrong place.**
   - GS `:78` cites "The server's own knobs" for the sensitive-write rules. The re-authentication
     and notification text is under "Two surfaces, one API" (`configuration-homes:142-148`).
   - GS `:156` cites forward hook 4 for stream starts being an event kind. Hook 4 (`:469-470`)
     requires only an event-kind column. The stream-start rule is `streamer:24` and `:308-310`.
8. **The volume view needs fields that play history lacks.** GS `:153-157` asks for per-device
   counts and bytes served from "download events of the same shape". A play-history row carries
   `client`, not a device or app-password id, and no bytes (`streamer:297-301`). Play counts,
   "most played", and smart playlists are views over the same table (`:101-102`, `:336-339`).
   Fold: name the added device id and bytes fields, and state that those views filter by event
   kind.
9. **"One consistent layout" contradicts per-collection templates.** GS `:31-32` says every
   album is filed "in one consistent layout". Each collection carries its own layout template
   (`configuration-homes:188-190`, `:250-251`). Fold: "a consistent layout per collection".
10. **Mirror approval happens once per member, not once per collection.** GS `:143-144` records
    the source "when the admin approves its first mirror". The admin approves only a member's
    first mirror, after which "the member mirrors freely within quota"
    (`configuration-homes:621-623`), so a member's later collections have no approval event.
    Fold: the source answer is asked when each mirrored collection is created.
11. **The wishlist belongs in IDEAS, not the spec's table.** GS `:167-170` and `:224` merge an
    unscheduled wishlist with the Requests idea. IDEAS' header rule (`IDEAS.md:3-10`) keeps
    uncommitted features there until an owner ruling targets them. Fold: route the merged entry
    into `IDEAS.md` under Requests, and keep one pointer line in the spec.
12. **Plain terms.** GS `:38` says dubplate "isn't a way to download music", and `:43-44` says
    members "download one album at a time". The same word carries two meanings in copy meant for
    the website. "A shared shelf" and "a distribution node" (`:122-124`) are metaphors that the
    plain-terms rule steers away from. Fold: "has no way to get music from the internet", and
    drop the shelf image.
13. **The member cap and the unauthenticated surface are underspecified at their edges.**
    - GS `:74-82` applies the cap to "the roster's add path". It does not say whether the cap
      binds rung 6d's `member import` or the `member add --admin` bootstrap
      (`rung-5-identity-auth:120-124`). The bootstrap must never be refused.
    - The unauthenticated-surface list (GS `:89-91`) omits `dubplate init`'s first-run web page
      on an unconfigured server (`streamer:357-360`).

    Fold: one sentence for each.

## Where ratified documents need an erratum

| Document | Passage | Driven by |
|---|---|---|
| `2026-09-07-streamer-2-0-design.md` | Shares `:289-295`; slice 6 name `:442`; web surface `:395-396`; API surface `:384-386`; Revision 2 item 12 `:42`; hook 15 share rows `:507`; DAP section `:406-409`; identity section (device cap and admin list); hook 17 and "Added by" `:323-324`, `:515-517` | Draft's conflicts 1 to 3, B1, M2, minors 1 and 2 |
| `2026-09-11-configuration-homes-design.md` | Check out `:498-502`, `:633`; share rows `:564-566`, `:675`; projector list `:705-706` | B1, conflict 1, minor 1 |
| `2026-09-23-rung-6-ops-seam-store.md` | Reserved share rows `:485` | minor 1 |
| `2026-09-02-greenfield-server-build.md` | `/share/*` bypass `:481` (conditional on the driver's sharing setting) | M1 |
| `2026-09-01-dubplate-organizer-design.md` | The 2026-09-12 checkout erratum; timer-path attribution `:651`, if ruled | B1, M2 |
| `ROADMAP.md` | Checkout line `:62-63` | B1 |
| `IDEAS.md` | Requests entry, merged with the wishlist | minor 11 |

## Owner forks raised by this review

1. Checkout scope (B1). Recommendation: a member checks out only collections they own.
2. Navidrome downloads during 0.x (M1). Recommendation: leave them on, pin sharing off, and
   state the residual.
3. When buy links ship, re-posed (M3). Recommendation: rung 7's album page if a task can absorb
   it without growing the pass, otherwise the first later release that renders them.
