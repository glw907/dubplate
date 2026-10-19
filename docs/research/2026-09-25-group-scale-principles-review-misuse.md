# Group-scale principles spec: misuse and friction review

Target: `docs/specs/2026-09-25-group-scale-principles-design.md` at commit
`1d6bc6a` (revision 1, pre-review). Lens: misuse and friction, in place of data integrity.
Each finding is judged against the owner's framing (Geoff, 2026-09-25): dubplate should be easy
to use for its intended purpose and hard to use for nefarious ones, and helping people buy
music is a bonus. "Stop piracy" is not the bar.

Sources read beside the spec: the streamer spec (`2026-09-07-streamer-2-0-design.md`, the
streaming, identity, shares, group features, and API surface passages), the rung 5 spec
(`2026-09-08-rung-5-identity-auth.md`, fork 4 and offboarding), and the configuration homes
spec (`2026-09-11-configuration-homes-design.md`, the mirror verb).

**Counts:** 0 blocker, 5 major, 9 minor. Four findings are OWNER FORKs (M1, M4, m7, m9).

The spec's strongest mechanisms are sound and cheap: no fetching, no public links, no
federation, and invite-only membership. Those four decide whether stock dubplate can be a
distribution hub, and each holds against every persona below. Most findings concern two
weaker things. Some mechanisms add friction for honest users and none for abusers. Some public
copy promises more than the product does.

## Pass 1: red team

### The operator who wants a pirate hub

The cheapest path on stock dubplate is short. The operator raises the member cap in one
sensitive setting. The re-auth and the notification reach the operator, because the operator is
the admin. They add members with a shell loop over the `member` verb, and they load the library
by a mirror or by placing files at migration. Nothing in the spec raises that cost much, and it
shouldn't try to. The spec already says as much (lines 51-63).

What does raise the cost is structural. With no public links and no anonymous stream, every
listener needs a roster row, a magic-link login, and an app password. With no federation, one
hub can't seed another. With no fetching, the operator has to supply every file. Those
mechanisms turn a hub into manual labour. They cost honest users nothing. **Keep all four as
written.** No finding here.

### The member who wants to leak the library

The cheapest path is a short script. It uses one app password, walks `getAlbumList2`, and calls
`stream` with no format for every track. The streamer spec's direct path serves the library
file itself through range-aware serving, with no per-member limit (streamer spec lines 247-249).
The copy is bit-identical to a download, and it never touches the album-zip limit. See M1 and M2.

The second cheapest path is to hand one app password to many people. See M3.

### The outsider

The unauthenticated surface is small and serves no audio, art, or catalog data (lines 85-91).
Dropping shares removes the only path that handed audio to anyone holding a link. Roster
enumeration through the kind unknown-address answer is a ruled trade-off (rung 5 fork 4) with a
rate-limit mitigation, so this review doesn't reopen it. An outsider's realistic path is a
member's leaked app password, which M3 covers. No further finding.

## Findings

### M1 (major, OWNER FORK). The album download limit stops honest users and no one else

**Location:** spec lines 126-129, 214-215, and 269-273 (ruling 3). Also the streamer spec, lines
247-249 (the direct stream path).

**Defect:** The limit counts album zips only, and ruling 3 says single-track downloads "fall
under the volume view instead". `stream` with no format serves the same original FLAC bytes
without any cap. A member who wants to copy the library never needs an album zip. So the limit
has no effect on the leak path. It lands only on an honest member who loads albums for a trip
through a client that happens to use zips. Most OpenSubsonic clients fill their offline caches
track by track, so honest users rarely hit it either. When one does, the Subsonic error surface
gives them error code 0 and a message that many clients show poorly or not at all. This is the
spec's one mechanism that fits the "security theater" test: it costs honest users something and
abusers nothing.

The mechanism that does work against the leak path is the volume view, because bytes served
count every path. As written, though, the view is passive. It catches a bulk copy only if the
admin happens to look during the window.

**Proposed fold:** Present this as an OWNER FORK in place of ruling 3's download half.

- (a) **Recommended.** Drop the hard album limit. Keep `download` narrowed to a track or an
  album, since that narrowing (conflict 2) is cheap and sound. Make the volume view the
  mechanism. Add one admin notice by email, which rung 5's SMTP mail already supports, when a
  member's or device's bytes served in a rolling 24 hours pass a threshold. The threshold is a
  setting whose default sits well above a heavy honest day, for example a large share of the
  library's size or a fixed number of gigabytes. Nothing is refused. The admin gets a heads-up
  and decides. This gives the leak path real friction, because a whole-library copy becomes
  visible within a day, and it adds no friction for honest users.
- (b) Keep a refusing limit, but count bytes across `stream` and `download`. This closes the
  bypass. It also breaks honest offline caching and a long listening day, which fails the
  owner's "easy for its intended purpose" test.
- (c) Keep the spec as written. It is cheap to build and does nothing.

### M2 (major). The public page promises more than the product does

**Location:** spec lines 33-34, 43-45, and 126 ("No bulk export over the API").

**Defect:** Line 22 says the "is / isn't" section is written for the website nearly as it
stands. Three of its claims won't be true when a stranger reads them:

- Lines 43-45 say members "download one album at a time" and that whole-library copies stay
  with the operator. Any member can copy the whole library one track at a time over `stream`
  (M1). The heading on line 126 makes the same claim in the spec's own voice.
- Lines 33-34 say "Album pages link to places to buy the music, and the group keeps a list of
  albums someone should buy." Buy links wait until the first-party web player after 0.8
  (ruling 5), and the wishlist is unscheduled (line 170). At 1.0 and 0.8 neither exists.

A public page that overclaims loses credibility once a reader checks it. It also invites the
argument that a technical control was promised and not delivered.

**Proposed fold:** Rewrite line 43 in terms of what the product declines to offer: "A way to
hand a library to strangers. dubplate has no whole-library or artist-wide download, no public
link, and no way to add anyone outside the group. Whole-library backups stay with the person who
runs the server." Retitle line 126 "No bulk download endpoint" and add one sentence saying that
the volume view is how unusual copying shows up. Mark the buy-link and wishlist bullet as
planned, or hold it off the website page until ruling 5's release ships.

### M3 (major). The device cap counts credentials, not the people using them

**Location:** spec lines 101-110, 210-211, and ruling 3 (lines 269-273).

**Defect:** A member can mint one app password and give it to many people. Nothing limits how
many clients use one credential at once. The admin device list shows one "last-seen address" per
device, and that value hides the sharing. For an outsider, this is the cheapest way into the
library, and the device cap doesn't touch it. The data needed to see it already exists. The
streamer spec records "first use from a new address" as an event (streamer spec lines 284-285).

**Proposed fold:** In the admin device list, replace the single last-seen address with "distinct
addresses in the last 7 days" and "most concurrent streams in the last 7 days" per device. The
bytes-served notice from M1 also applies per device. Don't refuse on either number. Phones
change addresses often, so a number is a prompt for the admin to ask, never an automatic
verdict. This costs one query over events the design already writes.

### M4 (major, OWNER FORK). A named volume view reads as surveillance inside a family

**Location:** spec lines 153-158 and ruling 4 (lines 274-278).

**Defect:** The recommendation shows stream counts, album downloads, and bytes per member and per
device over a window. For a parent admin and a teenage member, stream counts per device over time
come close to a listening log, even without titles: "your phone streamed 300 tracks between
midnight and 3 a.m." The spec itself rejects watermarking because it "tells the members … that
the server distrusts them" (lines 178-182). A standing per-person dashboard sends the same
message more directly. The view has one job, which is spotting unusual copying. Only bytes served
answers that. Stream counts and times add no signal for it.

**Proposed fold:** OWNER FORK, replacing ruling 4's options.

- (a) **Recommended.** Name members and devices only on bytes served, with no time-of-day
  breakdown. Show stream counts as group totals only. Tell each member on their own page in one
  plain sentence: "The admin can see how much data each member's devices download, not what you
  play." The M1 threshold notice names the member or device that crossed it.
- (b) Exceptions only. Show group totals by default, and name a member or device only when it
  passes the M1 threshold.
- (c) As the spec recommends: all counts, named.

Either (a) or (b) keeps the view's purpose and drops the parts a family would read as
monitoring. The download events the spec adds (line 157) can then carry bytes without a per-event
title. That matches the spec's own "counts only" promise, which the event row as written doesn't
guarantee. Hook 4's event shape includes the song id.

### M5 (major). The member cap doesn't say who counts, and offboarded members never leave

**Location:** spec lines 74-83 and 204. Also rung 5 spec lines 1101 and 1445-1447.

**Defect:** Rung 5 offboards a person by bumping their valid-after epoch. Its add-member page
"offers no removal control". So a departed member's roster row stays in place, and the spec
doesn't say whether that row counts against the cap. If it does, a real group with normal churn
fills the cap over the years and gets refused with a message that points at raising a sensitive
setting. That tells an honest operator to grow the group when the actual problem is a departed
cousin. It is also unclear whether admins count.

**Proposed fold:** State that the cap counts roster members who haven't been offboarded, admins
included. The refusal names both routes out: offboard someone who has left, or raise the setting.
This is one sentence in the spec and one predicate in the add path.

### m1 (minor). The Navidrome import skips members partway through

**Location:** spec lines 80-82.

**Defect:** An honest switcher with 25 Navidrome users gets a partial import. The skip report
arrives after the fact, and the skipped members' stars, playlists, and history are then orphaned
or dropped.

**Proposed fold:** The import counts members before it writes anything. If the count exceeds the
cap, it stops before starting and names the setting. The operator raises the cap and re-runs, or
passes an explicit list of members to bring. A preflight refusal is friendlier than a partial
success.

### m2 (minor). Device cap churn falls on the least technical member

**Location:** spec lines 104-106 and ruling 3.

**Defect:** App passwords are shown once. So a reinstall, a new phone, or a client switch each
mints a new one, and the old one stays. Shared household devices, such as a car, a speaker, or a
living-room tablet, count against whichever member set them up, usually the parent admin. After
a few years a grandparent reaches the cap and meets a refusal they can't act on.

**Proposed fold:** The refusal page lists the member's devices by last seen and offers a one-step
revoke on each, with any device unused for 90 days marked first. Ten is a fine default. The fix
is the recovery path, not the number.

### m3 (minor). One source answer per upload doesn't fit a batch upload

**Location:** spec lines 139-145 and ruling 2 (lines 265-268).

**Defect:** A cousin uploading a mixed batch, some vinyl rips and some Bandcamp downloads, gets
one question for all of them, so some answers will be wrong. The laptop mirror records one answer
for a whole collection (line 143), even though a collection mixes sources by nature. The M1 and
M4 folds don't depend on this answer, so its only value is the norm and provenance (m6). Wrong
answers weaken both.

**Proposed fold:** Ask once per upload, then let each album in the batch override it on the
review list the uploader already shows. For a mirror, record the collection's answer as a default
that a later album-level answer can replace. Keep ruling 2's recommendation of required. The
question is not theater, because the spec never claims it stops anyone (line 141). It sets a norm
at the cost of one click.

### m4 (minor). The source options leave out common honest cases

**Location:** spec lines 140-141.

**Defect:** The list is Bandcamp, Qobuz, CD rip, vinyl rip, gift, and other. It doesn't cover
other stores, such as a label's own shop or another download store. It doesn't cover free or
name-your-price downloads, or a member's own recordings. All of these land in "other", so the
answer records less than it could. "Gift" is ambiguous. A gifted CD is fine, but "a friend gave
me a copy" is exactly the line the product is drawing. Line 29 ("music someone gave you") has the
same ambiguity on the public page.

**Proposed fold:** Use these options: "Bought as a download (store: ___)", "Ripped from my CD",
"Ripped from my vinyl or cassette", "Free or name-your-price download", "My own recording",
"Gift (a CD, record, or gifted download)", and "Other (note)". On line 29, write "music given to
you as a gift, such as a CD or a gifted download".

### m5 (minor). The spec leaves out the cheapest artist link in the 0.8 client era

**Location:** spec lines 162-165 and ruling 5 (lines 279-284).

**Defect:** Ruling 5 puts buy links off until a first-party album page exists, so during 0.8 the
"money points back at artists" principle has nothing in front of a listener. `getAlbumInfo2`
already has a `musicBrainzId` field alongside `notes`, and several clients render it as a link
out. The MusicBrainz release page lists the release's purchase and streaming links. dubplate
already knows every admitted album's release id from tagging.

**Proposed fold:** State that 0.8 fills `musicBrainzId` in `getAlbumInfo2` and `getAlbum`, if the
streamer spec doesn't already. The ruling 5 recommendation stands for dubplate's own links. This
costs nothing and gives members a path to a buy page in the client era.

### m6 (minor). Provenance is recorded but never shown

**Location:** spec lines 142-143 and 147-151.

**Defect:** The source answer lands on the album row, and no member ever sees it. Showing it is
the cheapest pro-artist and pro-physical-media signal in the design. "Added by Ana, from her
vinyl" reads as pride in owning the record, and it makes buying the norm without anyone
preaching.

**Proposed fold:** Show the source beside `added_by` wherever added-by appears. In 0.8, add one
short line under the member's note in `getAlbumInfo2`'s `notes`, such as "From Ana's vinyl".
That's one line, so it doesn't crowd the note, which is the concern ruling 5 raises. When the
source is a download store, let the uploader take an optional store link. That link becomes the
album's first buy link at rung 7, with no MusicBrainz lookup. It is also the link most likely to
pay the artist directly, since it's where a member actually bought the music.

### m7 (minor, OWNER FORK). What order buy links appear in

**Location:** spec lines 162-165.

**Defect:** MusicBrainz URL relationships mix artist-direct stores, such as Bandcamp and the
artist's official site, with large retailers, mail-order for physical copies, and free streaming
services. Shown in arbitrary order, a streaming service can end up first, which does little for
the artist.

**Proposed fold:** OWNER FORK. (a) **Recommended.** List the artist's own store and Bandcamp
first, then physical copies by mail order, then other download stores. Show streaming services
separately under "Listen elsewhere", or leave them out. (b) Show MusicBrainz's order unchanged.
(c) Show only the artist's own store and Bandcamp.

### m8 (minor). The wishlist has no purchase step

**Location:** spec lines 167-170.

**Defect:** The wishlist as written closes when someone uploads a match. Nothing in between helps
the group buy the album. Two members may both buy it, or neither does because each assumes the
other will.

**Proposed fold:** When the wishlist is scheduled, each entry shows the album's buy links (from
m5, m6, or m7) and has a "I'm buying this" claim that names the member. The entry then closes on
upload, as the spec says. Record this as a note on the unscheduled feature, not new scope.

### m9 (minor, OWNER FORK). A local record store link

**Location:** spec lines 160-174.

**Defect:** MusicBrainz carries no local stores. Many groups have a shop they like, and linking to
it is the most direct physical-media nudge available.

**Proposed fold:** OWNER FORK. (a) **Recommended.** File it in `IDEAS.md` as an optional group
setting, "our record store" (a name and a link), shown beside buy links once they exist. (b) Add
it to this spec now. (c) Decline it.

## Checked and not raised

- **The member cap's re-auth and admin notice** reach the operator themself on a one-admin
  server. That isn't a defect, because the spec calls the cap soft and a deliberate act, and it
  claims no deterrence (lines 78-80).
- **Required source question.** It adds friction for abusers too (they click "other"), but the
  spec states it is a norm, not a control, and it costs one click. This review agrees with
  ruling 2, subject to m3 and m4.
- **Member cap default of 20.** A family of 12 and a friend group of 12 both fit easily. The
  number is fine once M5 settles who counts.
- **Device cap default of 10.** A teenager with three devices fits easily. The number is fine
  once m2 settles recovery.
- **Declined watermarking and ethical license.** The reasoning holds under the owner's framing.
- **Roster enumeration** through the kind answer is ruled (rung 5 fork 4).
