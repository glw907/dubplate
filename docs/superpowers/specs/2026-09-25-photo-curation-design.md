# Photo curation design

Status: draft for Geoff's read, 2026-09-25. Initiative: ROADMAP `photo-curation`.

This pass builds dubplate's photo pool: about 300 openly licensed period photographs, show
flyers, and zine covers. Each is stored under `photos/` with a record that says what it shows,
where and when it was taken, who made it, and what license lets dubplate ship it. The records use
a closed vocabulary of scenes, subjects, shots, formats, and places. With that vocabulary, the
pool's balance can be counted rather than judged by eye, and the future website can browse the
same records. Geoff picks every image from contact sheets. Photographers and archives whose work
is not openly licensed get a letter that Geoff sends himself. The pass writes data and letters
only: the loader, the photo column, the CREDITS generator, and the rotation belong to the rung
that first displays a photo.

The visual identity spec (`2026-09-24-visual-identity-design.md`) set the curation and licensing
rules this pass follows. Where this spec differs from it, this spec is the later ruling, and the
differences are listed at the end.

## Pool size and balance

The pool targets about 300 picked images, all embedded in the binary (Geoff, 2026-09-25: "I don't
think an additional ~45MB is going to turn anybody off"). No cap applies. A round that turns
up strong images keeps going.

Scene families carry these shares of the picked pool (Geoff, 2026-09-25):

| Family | Share |
|---|---|
| Reggae | 30% |
| Punk | 20% |
| Hip-hop | 20% |
| Everything else | 30% |

An image with no scene, such as a turntable close-up, counts toward "everything else". No single
region may hold more than half the pool. The shares steer each search round and show as running
counts on the contact sheet. Geoff's picks can override them.

Coverage has a floor as well as shares. Each of the three named families carries at least two
picked images in each of these subjects: sound systems, clubs and venues, recording studios,
record stores, crowds and dancers, and a landmark. Posed portraits stay a small share of shots,
which is how the identity spec's "scenes beat posed portraits" rule is checked.

The deep-cuts rules stand. Each headline artist gets at most one image, and each landmark gets
at most one unless Geoff picks a second on purpose.

## Categories

Every record carries the facets below. Each facet except artists and landmarks draws from a
closed list in `photos/vocab.json`. A term outside the list fails validation, so a new term joins
by editing the list, never by an agent writing it into a record.

Scene
: Optional. One primary scene and any number of secondary ones. Each scene belongs to one
  family, and the family is derived from `vocab.json`, never stored. The reggae family holds ska,
  rocksteady, reggae, dub, dancehall, 2 Tone, and lovers rock. The punk family holds protopunk,
  punk, hardcore, and post-punk. The hip-hop family holds every era, including go-go, graffiti,
  and b-boy scenes. Everything else holds jazz, blues, mod, garage, rockabilly, and the
  international scenes such as highlife, Afrobeat, and cumbia.

Subject
: One or more of: performers; crowds and dancers; sound systems; record stores; clubs and venues;
  recording studios; homes and label headquarters; instruments; record players and hi-fi; radio
  stations; pressing plants and cutting lathes; streets and neighborhoods.

Shot
: One of: live, candid, posed, interior, exterior, still life.

Format
: One of: photograph, flyer, zine cover.

Place
: A city and an ISO 3166-1 country code. The region is derived from the country through
  `vocab.json`: Caribbean, UK and Ireland, North America, Africa, Latin America, Europe, Middle
  East and North Africa, Asia and Pacific. A named place of significance, such as Dischord House
  or King Tubby's studio, adds a landmark with its name and its Wikidata ID where one exists.

Year
: The year the image was made. The decade is derived. A place counts in any year, so a recent
  photo of a landmark's front qualifies.

Artists
: The performers shown, each with a MusicBrainz artist ID where one exists. The ID enforces the
  one-image-per-headliner rule by lookup, and it lets a later rung link a photo to albums the
  group owns.

Surfaces
: One or more of: login, error, empty, gallery. An error-page photo is a closed club, found as
  clubs and venues with an exterior shot, or a club's final night.

## Storage

Each picked image is two files in `photos/`, named by a slug that stays stable for the life of
the project and later serves as the website's path for the photo:

```
photos/
  vocab.json                 closed lists and the derivation maps
  schema.json                JSON Schema for one record; enums come from vocab.json
  dischord-house-1984.webp   the image
  dischord-house-1984.json   its record
```

One file per record keeps parallel search agents from colliding on a shared manifest. A consumer
concatenates the records into whatever form it needs.

A record carries the facets above and these provenance fields:

| Field | Content |
|---|---|
| `title` | A short caption title. |
| `photographer` | The maker, as the source credits them. For a flyer or zine cover, the artist. |
| `credit` | The attribution string the licensor asks for, verbatim, when the source names one. |
| `source_url` | The page the license was verified on. |
| `license` | An SPDX identifier with its version, such as `CC-BY-SA-3.0-NL`, or `public-domain`. |
| `license_url` | The license's legal code. |
| `public_domain` | For public domain only: the basis, the jurisdiction, and the US status. |
| `verified` | The license text quoted from the source page, and the date it was checked. |
| `source_sha256` | The hash of the file as downloaded. |
| `sha256` | The hash of the stored file. |
| `status` | `picked`, `rejected`, or `pack`. |

Images are stored as WebP at a 1600px long edge, about 150KB each, which puts 300 images near
45MB. Resizing and re-encoding is a technical modification under CC 4.0, not an adaptation, so it
does not widen a share-alike duty. The silver-print treatment stays a render-time CSS filter, as
the identity spec rules. The stored file is never tinted.

Only picked images are committed. Candidates wait in untracked scratch storage until Geoff picks
them, because git keeps every blob it has ever seen. A passed candidate keeps its record with
status `rejected` and no image, so later rounds do not find it again. A `pack` record marks a
candidate for the optional photo pack (NC or dubplate-only permission) and ships nothing.

Validation is `schema.json`, run with `ajv` on the workstation after each round. The pass adds no
Go: a new package would join the `./...` gate that other lanes run, and nothing reads the records
yet. The consuming rung adds the loader and a gate leg that checks every record against the
schema, every image against its `sha256`, and every slug against both files.

## Licensing

The identity spec's licensing rules apply unchanged. The shipped classes are public domain with
its jurisdiction and US status recorded, CC0, CC BY, CC BY-SA, and the Free Art License. Every
license is read and quoted from its source page, never taken from an aggregator's field, which
rules out trusting a Flickr or search-index license field alone. An "own work" upload that
credits someone else is rejected. A flyer or zine cover is artwork, so its license must come from
its artist, not only from the photographer who scanned or shot it.

## Process

The pass runs in rounds. Each round is a small set of search agents on Sonnet, one per family or
gap, briefed with the vocabulary and the current counts. An agent finds candidates, verifies each
license on its source page, downloads the file to scratch storage, and writes the record with
status unset. The first round re-verifies the candidates in
`docs/superpowers/research/2026-09-24-identity-photos/`. Later rounds target the gaps the counts
show, and each runs a landmark sweep: the studios, clubs, label houses, and shops that mattered to
each family, searched by name. Hip-hop starts close to empty, because the earlier research barely
covered it.

A fresh Opus reader then spot-checks a sample of the round's license quotes against the live
source pages. A round with a wrong quote is re-verified whole before Geoff sees it.

Geoff picks from a private contact-sheet artifact. It shows each candidate's thumbnail, caption,
and license, with running counts per family, region, decade, subject, and shot. Picks and passes
save to the artifact's database. The conductor reads them back, sets each record's status, moves
picked images into `photos/`, and commits with path-scoped commits. Rounds continue until the
picked pool meets the shares and the coverage floor.

## Outreach

The first letters go to Beth Lesser, the DC Punk Archive at DC Public Library, Charlie Phillips,
and Janette Beckman. Beckman moves up from the research's tenth rank because she shot early New
York hip-hop, and the hip-hop family has the thinnest open record. Each letter names specific
images and asks for CC BY 4.0 on them. If the answer is no, the fallback is the optional photo
pack, where NC or a dubplate-only permission is acceptable.

The conductor drafts each letter in Geoff's voice: a person building a free music server for a
group of friends, to be published under the MIT license. Drafts live in
`docs/superpowers/research/2026-09-24-identity-photos/outreach/`, one file per recipient. Geoff
reads and sends every letter from his own address, and the conductor never sends one. A reply
that grants a license is verified like any other source and joins the pool as a record.

## Out of scope

The loader, the photo column, error-page and empty-library photos in the web app, CREDITS
generation, rotation order, and the optional pack's packaging all belong to the consuming rung.
This pass also touches nothing under `web/`.

## Changes to the identity spec

This spec changes four things in the identity spec's Photography section:

- The pool target rises from 100 or more to about 300.
- Hip-hop becomes a named family with a 20% share. It was one of the outward spreads.
- Stored images are WebP at 1600px, with a recorded source hash. The identity spec said the
  shipped file stays unmodified. It remains unmodified in content, only resized and re-encoded.
- Subjects, shots, formats, and landmarks join the scenes as recorded categories.

The identity spec gets a one-line pointer to this spec when this spec is approved.
