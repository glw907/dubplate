# IDEAS

> Future consideration: features and requests nobody has committed to. Committed work lives
> in `ROADMAP.md` (initiatives) and `BACKLOG.md` (work items); the 0.8 spec keeps its own
> 2.1 list. An idea leaves this file when an owner ruling targets it. Record the target and
> date in the entry and move it to Promoted, so the idea's origin survives. Read on demand,
> never at session start.
>
> Each entry says what the feature is, why a group would want it, what it builds on, and
> where it came from. Member requests go under Requests with the asker's name and date.

## Ideas

### Library health report
A report of what the library is missing or has twice: albums short of their MusicBrainz
tracklist, the same album uploaded by two members, missing art, and mixed audio quality.
A first version reports; a later one proposes fixes into the decision queue. Builds on
the matcher, the catalog's audio keys (0.8 hook 1), and the art blob (hook 3). Source:
the 2026-09-24 differentiator review.

### Add or fix a MusicBrainz entry
When MusicBrainz lacks an album or has it wrong, a member can add or fix the entry from
the web app or the local TUI, then re-match the upload. The known route is release editor
seeding (MusicBrainz's "Release Editor Seeding" developer page). Picard and Harmony both
use it. dubplate would pre-fill the editor from the upload's tags and tracklist, and the
member submits under their own MusicBrainz account, since MusicBrainz has no general edit
API. It gives the unmatched-album route (ruled for 0.8 on 2026-10-10) a way back to a
matched album. It builds on the matcher, intake's held uploads, and the 0.9 local client.
Source: Geoff, 2026-10-10, wanted after 0.8 with no version target yet.

### Requests
A member asks whether anyone has an album. When another member uploads a match, intake
closes the request and tells the asker. A single-user streamer has no reason to build it,
which is why it suits dubplate. Builds on intake's album matching and the notify path.
Source: the 2026-09-24 differentiator review.

The group wishlist folds into this entry: an album someone should buy. A member can claim
it ("I'm buying this"), and the entry shows the album's buy links under the group-scale
principles spec's rules ("Money points toward buying"). Requests and the wishlist live in
the web app only. They stay out of the documented client
API and emit no lifecycle-hook event, because a machine-readable want list is the input a
fetcher consumes. Source: the group-scale principles spec, 2026-09-25.

### Our record store
An optional group setting that names a local record store the group likes. Its link shows
beside the MusicBrainz buy links on album pages, so a member who wants the album on vinyl
or CD can buy it close to home. Builds on the buy links (the group-scale principles spec,
"Money points toward buying") and the store-backed settings table. Source: the group-scale
principles spec, 2026-09-25.

### New-arrivals email
A weekly email of what the group added, grouped by member, with each album's note (0.8
group features). It brings people back to the library without any client feature. Builds
on the SMTP notify path from rung 5 and `added_by` (0.8 hook 17). Source: the 2026-09-24
differentiator review.

### Group listening views
"What the group has on repeat" and "a member's favorites," derived from play history.
Each member needs a switch that keeps their listening out of the group views. Builds on
play history as events (0.8 hook 4). Source: the 2026-09-24 differentiator review.

### Unauthenticated local mail relay
Let an operator point sign-in mail at a relay on `localhost:25` with no credentials and
opportunistic TLS, for a host that already runs an MTA. Builds on `internal/notify`'s
transport, which today requires authenticated submission over TLS. Whether this is a
supported mail class is a product question. Source: rung 5d's universality review
(UNI-5d M2), 2026-10-05.

### Mail send test before go-live
A `doctor --mail` or `notify test` verb that sends one message, so an operator proves mail
works before the first member signs in; doctor never dials SMTP today. Source: rung 5d's
universality review (UNI-5d M2), 2026-10-05.

### External streamer integration
A driver seam that wires dubplate to a separate music streamer, with Navidrome as the
reference driver and Jellyfin on demand, so a group keeps the streamer it already runs.
Target systems were trimmed to Navidrome and Jellyfin by Geoff on 2026-09-03: living
Subsonic use is mostly Navidrome itself, Airsonic-Advanced is stale with no theming
surface, and headless servers' users live in clients. A group would want it to adopt
dubplate's library tools without changing how it listens. It builds on the verify-identity
seam and the lifecycle hooks. Retired because dubplate's own OpenSubsonic streamer is the
first launch (rung 12): one binary serves the web app and the music apps, so a second
service has nothing left to add. The 0.8 spec's hook 6 retired with it. Source: ruling 4
of the 2026-10-05 re-plan, Geoff; the entry moved from ROADMAP's folded organizer-spec
bullets.

### A forward-auth front
Cloudflare Access, or any identity-aware proxy, as an optional front that signs members
in before dubplate sees them, with the verify-identity seam reading the identity the front
asserts. A group already behind such a front would want one sign-in. Music apps
authenticate with app passwords that a front cannot gate, so a front would guard only the
web app. The promise also lived in three places (the seam's doc, the `DUBPLATE_ACCESS_*`
messages, and the runbook's topology section), which Owner ruling 2 removed. Builds on
`auth.Seam`. Source: Owner ruling 2 of the 2026-10-05 re-plan, Geoff; the entry moved from
ROADMAP's folded organizer-spec bullets.

### Beets-migration confidence check
A tool a switcher runs against their own live beets library, to see how dubplate's matcher
and writer would treat it before committing to the move. A group leaving beets would want
the proof before the import. It builds on the matcher and the conformance corpus. The
shadow program was the first form of it, deleted by the removal pass under Owner ruling 1
because few strangers run a live beets import timer and 0.8's switcher path is
`dubplate import navidrome`. Source: the `pre-greenfield` tag and commit `680afed`, the
last commit with `internal/engine/shadow` and its verb, plus Owner ruling 1 of the
2026-10-05 re-plan, Geoff.

## Requests

None yet.

## Promoted

- Report a problem, `added_by` with album notes, collaborative playlists, and smart
  playlists: targeted for 0.8 by Geoff, 2026-09-24 (streamer 0.8 spec revision 2.3;
  re-keyed from "2.0", Geoff, 2026-09-25).
- HTTPS without Cloudflare, a general Navidrome import, arm64 builds, `dubplate init`, and
  `dubplate upgrade`: targeted for 0.8 by Geoff, 2026-09-24 (streamer 0.8 spec revision 2.4;
  re-keyed from "2.0", Geoff, 2026-09-25).
