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

## Requests

None yet.

## Promoted

- Report a problem, `added_by` with album notes, collaborative playlists, and smart
  playlists: targeted for 0.8 by Geoff, 2026-09-24 (streamer 0.8 spec revision 2.3;
  re-keyed from "2.0", Geoff, 2026-09-25).
- HTTPS without Cloudflare, a general Navidrome import, arm64 builds, `dubplate init`, and
  `dubplate upgrade`: targeted for 0.8 by Geoff, 2026-09-24 (streamer 0.8 spec revision 2.4;
  re-keyed from "2.0", Geoff, 2026-09-25).
