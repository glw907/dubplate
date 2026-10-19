# beets 2.14 and the disc layout: findings passed in on 2026-09-23

Another session passed these in at Geoff's request. That session read HEAD `e8bd5f5` and changed
nothing. They are information for routing and need no action in the overnight window.

## beets v2.14.0 released the "ft" pattern fix

beets v2.14.0 (2026-09-07) fixes the distance calculation that read ordinary words containing
"ft" (draft, left, gift, craft) as a "featuring" suffix. v2.14.1 (2026-09-17) followed with bug
fixes only, and none of them touches the matcher. Release notes:
https://github.com/beetbox/beets/releases/tag/v2.14.0

`internal/engine/matcher/tunables.go:377-381` cites that fix (`19d091440`) as "unreleased at
pinned 2.13.1". The oracle pin is still beets 2.13.1: STATUS lists it under "Pinned versions", and
`scripts/deploy.sh:32` sets `BEETS_VERSION`. Bumping the box's beets would move the conformance
oracle in the middle of the shadow run. That makes this a pin decision for Geoff, not a free minor
bump, and it goes through the `dependency-upgrade` skill.

## The default layout puts no disc number in file names

`config/beets-config.yaml` sets `default: $albumartist/$album%aunique{}/$track $title` with
`per_disc_numbering: yes`. The tracks of a multi-disc album therefore interleave in one folder. Two
tracks with the same number and title on different discs collide, and beets appends `.1` to the
second. Tag-based browsing in Navidrome is unaffected. Folder-mode Subsonic clients and file
browsers show the tracks in the wrong order.

Two commitments fix this template in place at the swap: the launch bar's layout-parity check, and
the organizer spec's no-move assertion (`2026-09-01-dubplate-organizer-design.md:455-458`). A change
is cheapest before rung 16. After that it takes a deliberate one-time re-layout. This is an owner
fork. The case was seen live on New Order's *Power, Corruption & Lies*, the 2-CD deluxe edition
(MB `9a33a288-8fbe-42d7-b9ab-662101c75e58`), in a workstation import running beets 2.14.1.

## Candidate corpus rows from the same local import

The import covered 15 albums. Three correct matches fell below the strong threshold only because
of title suffixes such as "(Album Version)", "(2012 Remaster)", and "(Remastered)". They fall in the
`bug-remix-parenthetical` neighborhood:

- G. Love, *Yeah, It's That Easy*, 94.3%
- The Beat, *Special Beat Service* (2012), 92.4%
- Aim, *Hinterland* (remastered), 93.5%

illuminati hotties' *POWER* (2024) got no correct candidate from text search, because compilation
noise crowded it out. Matching by the release ID `48cd5a05-d9cb-4511-aed0-118600091f6b` scored 99.8%.
