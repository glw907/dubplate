# Photo curation design

Status: revision 2.3. Revision 2.2 was folded from four lens reviews, a fold verification, and a
prose review (fold record: `docs/superpowers/research/2026-09-25-photo-curation-fold.md`) and
approved by Geoff 2026-09-25. Revision 2.3 adds his two rulings from the round-one plan review
(fold record: `docs/superpowers/research/2026-09-25-photo-curation-plan-fold.md`): people and
minors are flagged for his decision, and passed images are no longer pruned from the sorting
page's file store. Initiative: ROADMAP `photo-curation`.

This pass builds dubplate's photo pool: about 300 openly licensed photographs of music scenes,
show flyers, and zine covers. Each is stored under `photos/` with a record that says what it
shows, where and when it was taken, who made it, and what license lets dubplate ship it. The
records use a closed vocabulary of scenes, subjects, shots, formats, places, and licenses. With
that vocabulary, the pool's balance can be counted rather than judged by eye, and the future
website can browse the same records. Search agents propose and score candidates, and Geoff picks
every image on a private sorting page. Photographers and archives whose work is not openly
licensed get a letter that Geoff sends himself. The pass writes data, one small Node check script,
and letters. The loader, the photo column, the CREDITS generator, and the rotation belong to rung
7, the first rung that displays a photo.

The visual identity spec (`2026-09-24-visual-identity-design.md`) set the curation and licensing
rules this pass follows. Where this spec differs from it, this spec is the later ruling, and the
differences are listed at the end.

## Pool size and balance

The pool targets about 300 picked images, all embedded in the binary (Geoff, 2026-09-25: "I don't
think an additional ~45MB is going to turn anybody off"). The 45MB was an early estimate. The
measured encode puts 300 images near 48MB (Storage). No cap applies. A round that turns up strong
images keeps going.

Scene families carry these shares of the picked pool (Geoff, 2026-09-25), each within plus or
minus 3 percentage points:

| Family | Share |
|---|---|
| Reggae | 30% |
| Punk | 20% |
| Hip-hop | 20% |
| Everything else | 30% |

A record's family comes from its primary scene. An image with no scene, such as a turntable
close-up, counts toward "everything else". The region cap holds any one region to at most half of
the picked records that carry a place. The posed ceiling holds posed shots to at most 15% of
picked photographs (flyers and zine covers are not counted). The shares steer each search round
and show as running counts on the sorting page.

The floor grid sets a minimum alongside the shares. It crosses the three named families (reggae,
punk, and hip-hop) with five primary subjects (sound systems, clubs and venues, recording studios,
record stores, and crowds and dancers). Each of the 15 cells needs at least two picked images
whose first-listed subject is that cell's subject. Each named family also needs at least two
picked images with a landmark. The surface floors ask for at least three picked images on the
error surface and three on the empty surface.

The deep-cuts rules stand, keyed so a script can check them. Each named artist gets at most one
image, keyed by MusicBrainz ID, or by a normalized name (case folded, a leading "the" dropped)
when the artist has none. Each landmark gets at most one image, keyed by Wikidata ID, or by a
normalized name when it has none.

Geoff can waive any of these rules for one specific case. A waiver is one entry in
`photos/meta/waivers.json` with three fields:

- `rule`: one of `share`, `region`, `posed`, `floor`, `landmark-floor`, `surface`, `artist`, or
  `landmark`.
- `key`: the case it covers. That is the family for `share` and `landmark-floor`, `all` for
  `posed`, the region for `region`, `<family>/<subject>` for `floor`, the surface for `surface`,
  and the dedupe key for `artist` and `landmark`.
- `reason`: Geoff's words, or for a keep on the sorting page, the round and the date of the keep.

A floor cell earns a waiver proposal after two targeted rounds come back empty for it, and
proposals go to Geoff bundled into one checkpoint question.

**Done** is a checkpoint where the pool is near its target, the check script's balance report
shows no breach without a waiver, and Geoff declares the pool complete.

## Categories

Every record carries the facets below. Each facet except artists and landmarks draws from a
closed list in `photos/meta/vocab.json`. A term outside the list fails the check script, so a new
term joins by editing the list, never by an agent writing it into a record. A vocabulary edit
that removes or moves a term rewrites the affected records in the same commit.

Scene
: Optional. One primary scene and any number of secondary ones. Each scene belongs to one
  family, and the family is derived from `vocab.json`, never stored. The reggae family holds ska,
  rocksteady, reggae, dub, dancehall, 2 Tone, and lovers rock. The punk family holds protopunk,
  punk, hardcore, and post-punk. The hip-hop family holds every era, including go-go, graffiti,
  and b-boy scenes. Everything else holds jazz, blues, mod, garage, rockabilly, outsider artists,
  and the international scenes such as highlife, Afrobeat, and cumbia.

Subject
: One or more of: performers; crowds and dancers; sound systems; record stores; clubs and venues;
  recording studios; homes and label headquarters; instruments; record players and hi-fi; radio
  stations; pressing plants and cutting lathes; streets and neighborhoods. The first-listed
  subject is the primary one, and only it counts toward the floor grid.

Shot
: One of: live, candid, posed, interior, exterior, still life.

Format
: One of: photograph, flyer, zine cover.

Place
: Optional. A city and the place's current ISO 3166-1 country code (`DE` for East Berlin, `RS`
  for Belgrade), with an optional free-text historical country for the caption. The region is
  derived from the current code through `vocab.json`: Caribbean, UK and Ireland, North America,
  Africa, Latin America, Europe, Middle East and North Africa, Asia and Pacific. A named place of
  significance, such as Dischord House or King Tubby's studio, adds a landmark with its name and
  its Wikidata ID where one exists. A place counts in any year, so a recent photo of a landmark's
  front qualifies.

Year
: The year the image was made, with a precision of `exact`, `circa`, or `decade`. The decade is
  derived.

Artists
: The performers shown. Search agents record names only. The conductor adds MusicBrainz artist
  IDs in one serialized step at no more than one request per second, with a User-Agent naming
  dubplate and a contact address (see Licensing for where the contact comes from), as
  MusicBrainz requires. Parallel agents on one IP would otherwise draw 503 responses. The step
  runs at round close, before the integrity check.

Surfaces
: One or more of: login, error, empty, gallery. An error-page photo shows a closed club, such as
  a club's final night or its shuttered front.

The film-edge caption the identity spec describes reads its top edge from `title` (the subject
slot), the artists' names (the band slot, empty when none), and `year`. Its bottom edge reads
from `photographer` and the license's display name from the license list.

## Storage

Each picked image is two files in `photos/`, named by a slug that stays stable for the life of
the project and later serves as the website's path for the photo:

```
photos/
  meta/vocab.json            closed lists, the derivation maps, and the license list
  meta/waivers.json          Geoff's recorded waivers
  meta/rejected.json         dedupe keys of passed and dropped candidates
  meta/pack.json             records of candidates for the optional photo pack
  dischord-house-1984.webp   the image
  dischord-house-1984.json   its record
  .candidates/               gitignored candidate home, never committed
```

The conductor mints each slug once, at round close (step 3), checks it is unused, and never
renames it. A factual correction changes the record, never the slug. A slug matches
`^[a-z0-9]+(-[a-z0-9]+)*$`, which also keeps it legal in a `//go:embed` pattern.

A picked record carries the facets above and these provenance fields:

| Field | Content |
|---|---|
| `title` | A short caption title, dubplate's own. |
| `source_title` | The title the source gives the work, verbatim, when it gives one. CC 2.0 and 3.0 require it in the attribution. |
| `photographer` | The maker of the image, as the source credits them. |
| `credit` | The attribution string the licensor asks for, verbatim, when the source names one. |
| `notice` | Any copyright notice embedded in the source file's metadata, verbatim. |
| `source_id` | A canonical source key: `commons:<pageid>`, `flickr:<photo id>`, or `archive:<name>:<inventory number>`. |
| `origin_id` | The upstream copy's key for a transfer, such as the Flickr ID a Commons FlickreviewR template records. |
| `source_url` | The page the license was verified on. |
| `file_url` | The exact URL the file was downloaded from. |
| `archive_url` | A Wayback Machine snapshot of the source page, taken at verification, when one succeeds. For a Flickr transfer, the FlickreviewR template's `archive` link serves. |
| `retrieved_at` | The full timestamp of the download. |
| `license` | An ID from the closed license list in `vocab.json`. The list uses the SPDX ID where one exists (`CC-BY-2.0`, `CC-BY-SA-4.0`, `CC0-1.0`, `LAL-1.3` for the Free Art License) and a `LicenseRef-` ID otherwise (`LicenseRef-CC-BY-SA-3.0-NL`, `LicenseRef-public-domain`). The list maps each ID to its display name and its legal code URL. |
| `public_domain` | Required for `LicenseRef-public-domain`, forbidden otherwise: a basis from a closed list in `vocab.json`, the jurisdiction, the US status, which must be public domain in the US, and the fact the basis rests on (a first-publication year or the author's death year). A public-domain work always takes `LicenseRef-public-domain`; a mark such as `CC-PDM-1.0` goes into `verified` as evidence, never into `license`. |
| `verified` | The license evidence quoted verbatim from the raw source (see Licensing), the date, and who verified it. |
| `artwork` | Required when the format is flyer or zine cover: the artwork's artist, license, credit, `source_url`, and `verified`. Its license must be one of the shipped classes, and a public-domain artwork carries its own `public_domain` object under the same rule as the top level. The top-level fields then describe the photograph or scan of it. A faithful flat scan that carries no separate right records that basis in place of a second license. |
| `source_sha256` | The hash of the file as downloaded. |
| `sha256` | The hash of the stored file, written by the step that encodes it. |
| `width`, `height` | The stored file's pixel dimensions. |
| `size_exception` | Required when the stored long edge is 1000 to 1199px, forbidden otherwise: one line saying why the image is unique or valuable enough to fall below the 1200px floor. |
| `status` | `picked`; `withdrawn` for an image removed after commit (see Licensing); or `pack` for a record in `pack.json`. A candidate record has no status. |
| `withdrawn` | Required when the status is `withdrawn`, forbidden otherwise: the reason and the date. |

Candidate records also carry three fields of their own:

- `screen`: the search agent's 1-to-5 score and one-line note. It is stripped when a record is
  committed, so no judgment of a named photographer's work enters the repository.
- `pack`: `true` when the license is NC or grants dubplate-only permission. Round close moves
  such a record to `pack.json` with the status `pack` and drops the field.
- `flags`: any people flags from the closed list in `vocab.json` (Process). It is stripped when a
  record is committed, like `screen`.

Images are stored as WebP, resized down to a 1600px long edge and never upscaled, encoded with
ImageMagick 7.1.2 and cwebp 1.6.0 as `magick <source> -auto-orient -resize '1600x1600>' <tmp>.png`
then `cwebp -q 80 -m 6 -metadata icc`. `-auto-orient` matters because neither `-resize` nor cwebp
applies the EXIF orientation, so a rotated source would otherwise be stored sideways. That setting
measured a 164KB mean on 24 real candidates, which puts 300 images near 48MB. Keeping only the
ICC profile preserves tone and drops EXIF GPS, and the record's `notice` field carries any
embedded copyright notice into CREDITS. The silver-print treatment stays a render-time CSS
filter, as the identity spec rules. The stored file is never tinted.

The sorting page's thumbnails use the same encode with `-resize '600x600>'`. The agent that
downloads a candidate writes its thumbnail beside it in `photos/.candidates/`. Thumbnails are
never committed.

The size floor is a 1200px long edge (Geoff, 2026-09-25: "We can lower for genuinely unique or
valuble photos. We can go as low as 1000px, but the photo needs to justify that."). A source of
1000 to 1199px is eligible only with a `size_exception` line in its record. The sorting page shows
that line beside the image, and Geoff's keep is the approval. A source under 1000px is never
eligible, and no image is upscaled. Agents also look for an archive's own larger scan.

Only picked images are committed. Candidates wait in `photos/.candidates/`, one subdirectory per
search agent, named by source ID in filename form (`commons-12345.jpg` and `.json`). The
directory is gitignored so it outlives sessions without entering git history. An agent never
overwrites a file it did not write.

A passed or dropped candidate leaves its `source_id`, `origin_id`, `source_sha256`, and a reason
code in `photos/meta/rejected.json`, so later rounds do not find it again. The reason codes are a
closed list in `vocab.json`. The starting list is `passed`, `blurred`, `too-small`, `license`, and
`duplicate`, and the rung plan may extend it. A candidate whose `pack` field is set goes to
`photos/meta/pack.json` as a full record without an image, and it ships nothing.

`tools/photos/check.mjs` is the one check, a plain Node script with no dependencies, which
reads `vocab.json` directly. It has two modes.

- **Integrity** must pass before every photo commit. It checks every closed list, the slug
  pattern, the required fields for each status and format, that a picked record's license is one
  of the shipped classes, the `public_domain` and `artwork` rules, that every picked record has
  its `.webp` with a matching `sha256`, that no `.webp` lacks a picked record, that no withdrawn
  record keeps an image, that no JSON file fails to parse, and that no `source_id`, `origin_id`,
  or hash repeats across picked, rejected, and pack records. It fails a picked record whose long
  edge is under 1000px, or under 1200px without `size_exception`, and it fails a second image of
  one artist or landmark without a waiver.
- **Balance** reports each share against its band, the region cap, the posed ceiling, the floor
  grid, and the surface floors, with waivers applied. It exits nonzero on any unwaived breach,
  but a breach blocks only the done checkpoint, because the pool fills over many commits.

The pass proves once that the script can fail, against fixtures in `tools/photos/testdata/`: a
duplicate artist, an NC-licensed pick, an NC-licensed flyer artwork under a CC BY scan, and an
empty floor cell. The sorting page's count bar is seeded from the balance mode's output, so both
read the same numbers.

The pass adds no Go. A new package would join the `./...` gate that other lanes run, and nothing
reads the records yet. Rung 7 adds the loader and ports both modes' checks into its Go gate leg.
When that leg carries the same checks, `check.mjs` and its `scripts/check.sh` line retire.

Rung 7 also inherits these embed notes. Per `go doc embed`, "The patterns are interpreted
relative to the package directory containing the source file." The embedding package sits inside
`photos/`, the same shape as `web/embed.go`, so its pattern is `*.webp *.json`. That pattern
matches only the image and record files at the top of `photos/` and leaves out `meta/` and
`.candidates/`. The roughly
48MB belongs in a leaf package so it stays out of other packages' test binaries.

## Licensing

The identity spec's shipped classes apply, with one tightening: a public-domain work must also be
public domain in the US. The classes are public domain with its jurisdiction and US status
recorded, CC0, CC BY, CC BY-SA, and the Free Art License. An "own work" upload that credits
someone else is rejected. A flyer or zine cover is artwork, so its license must come from its
artist, recorded in the `artwork` field, not only from the photographer who scanned or shot it.

License evidence is fetched raw with `curl` under a User-Agent naming dubplate and a contact
address, never through a summarizing fetch tool, which returns a small model's paraphrase in
place of a quote. Wikimedia may block a default script User-Agent (a probe with `python-requests`
drew 403), so the descriptive one matters. The contact in every User-Agent this pass sends, to
Commons, Flickr, MusicBrainz, or an archive, comes from the `DUBPLATE_HARNESS_CONTACT`
environment variable, the same one `cmd/harness/capture.go` reads. No agent puts any other
personal address in a request.

On Commons, `verified` quotes the license template and any review template from the page's
`action=raw` wikitext. On Flickr, it quotes the license link from the page HTML. An aggregator's
license field alone is never enough. For a Flickr-origin file, the license that counts is the one
in force when it was granted, shown by the FlickreviewR record or an archive capture, because a
CC grant is irrevocable. A later change on the live page is logged, not treated as a wrong quote.

When the same photograph turns up at several sources, the license comes from the upstream-most
copy that names its author, and copies whose licenses conflict reject the image.

Resizing and re-encoding modifies the file. CC 4.0 allows it outright: its section 2(a)(4)
authorizes technical modifications and says that making them "never produces Adapted Material".
CC 2.0 and 3.0 grant only modifications "technically necessary" to use another format, and they
carry no such safe harbor. For those versions the pass relies on Creative Commons' guidance that a
mechanical format change is unlikely to create a new work. Under either reading, share-alike binds
only the image itself, never the MIT code beside it. Every CREDITS entry therefore says the image
was resized and converted to WebP, which meets CC 4.0's duty to indicate a modification on any
reading, and entries for 2.0 and 3.0 works carry `source_title`.

A committed image becomes `withdrawn` when its license is later disputed, when its source is
deleted for a copyright violation, or when Geoff changes his mind about it. The record stays with
its `withdrawn` reason and date, the `.webp` is removed, and the integrity check refuses a
withdrawn record that keeps its image. The blob stays in git history. That is acceptable because
dubplate distributes images through the binary, and no release built after the withdrawal carries
the image. Moving images out of git to guard this case would add storage infrastructure for about
300 files. Before each release, the conductor rechecks every committed source URL.

## Process

The pass runs as a document lane registered in STATUS's lane registry, in the main tree that
lane 1 shares. Before round one, the conductor adds `/photos/.candidates/` to `.gitignore` and to
the `.gitleaks.toml` allowlist, since the gate's gitleaks leg scans the whole worktree. The
pass's only `scripts/` edit is one queued line in `scripts/check.sh` that runs the integrity
mode. It lands in the batch after lane 1 finishes, because lane 1 owns `scripts/check.sh` and
`Makefile` today. Until then, the conductor runs the integrity mode before every photo commit.

The pass runs in rounds. Each round is a small set of search agents on Sonnet, one per family or
gap, briefed with the vocabulary, the current counts, and the rejected ledger. An agent finds
candidates, checks each against existing `source_id` and `origin_id` keys, fetches its license
evidence raw, downloads the file to its candidate subdirectory, and writes the candidate record.
It sets `pack` itself for an NC or permission-only license. The first round re-verifies the
candidates in `docs/superpowers/research/2026-09-24-identity-photos/`. The flyer search starts
with DC hardcore flyers, as the identity spec asks. Later rounds target the gaps the counts show,
and each runs a landmark sweep: the studios, clubs, label houses, and shops that mattered to each
family, searched by name. Hip-hop starts close to empty, because the earlier research barely
covered it.

Each search agent looks at every candidate it keeps. It drops an image that is out of focus or
below the size floor (Storage), and it writes `size_exception` for a 1000 to 1199px source it
keeps. It lists each drop, with its keys and reason code, in a `drops.json` in its candidate
subdirectory, which the conductor merges into `rejected.json` at round close. It gives each
survivor a 1-to-5 gut score and a one-line note in `screen`. The score only sets the order Geoff
sees candidates in, so it decides nothing. The crop to the login page's tall column is Geoff's
judgment on the sorting page.

A fresh Opus reader then checks a sample of the round's license evidence through the same raw
channel, against `archive_url` where one exists. The sample is every public-domain and every
flyer or zine cover record, plus 20% of the rest, and at least five. Each result, a slug or
source ID with match or mismatch and the date, goes into the round's log under
`docs/superpowers/research/photo-curation-rounds/`. A round with a wrong quote is re-verified
whole before Geoff sees it.

Geoff sorts on a private claude.ai artifact: a web page only he can open, with its own file store
and a small shared database. The page shows a grid of the 600px thumbnails in score order, each
captioned with its subject, place, year, and license, and `Space` opens the full-size file. `Y`
keeps a candidate, `N` passes it, and the arrow keys move between candidates. Clicks work too, so
the page is usable on a phone. Filter chips narrow the grid by family, subject, and region, and a
bar shows running counts against the shares. Pack candidates are not shown.

Two kinds of candidate carry a mark. A candidate's `size_exception` line shows beside its image.
A candidate that would be a second image of a picked artist or landmark is flagged, and a keep on
it is Geoff's waiver, which the conductor writes to `waivers.json` at round close.

People are flagged, never excluded (Geoff, 2026-09-25: "Flag, you decide"). A search agent flags a
candidate whose Commons page carries a personality-rights template, or which shows an apparent
minor in focus, in its `flags` field. The sorting page shows the flag, and Geoff's keep or pass
decides.

The conductor uploads candidates to the page's file store from `photos/.candidates/` in batches of
25 files, the per-call limit, since a file outside the working directory takes a call of its own.
Geoff approves each round's uploads in one burst, and passed candidates are never deleted from the
file store, so a round asks for no delete approvals (Geoff, 2026-09-25: "Batch, no pruning").

Each choice saves as a database document keyed by `source_id`, never by slug, so a sitting can
stop and resume on any device.

A round closes in this order:

1. The conductor marks the round closed in the database, and the page stops accepting changes to
   it. A later change of mind reopens the round explicitly; for an image already committed, it
   means `withdrawn`.
2. The conductor reads the closed round's choices and applies each keep only to a candidate file
   whose hash equals its record's `source_sha256`. A missing file or a mismatch goes back to
   Geoff unpicked, never onto a substitute.
3. For each keep, the conductor mints the slug, encodes the WebP, writes `sha256`, `width`, and
   `height`, strips `screen`, and moves the pair into `photos/`. It then runs the serialized
   MusicBrainz ID step over the keeps' artists and writes any waiver a flagged keep carries. Each
   candidate Geoff passed, and each agent drop, goes into `photos/meta/rejected.json`. Each pack
   candidate goes into `photos/meta/pack.json` with `screen` stripped; its file is not kept, and
   packaging re-fetches it and checks it against `source_sha256`.
4. The integrity check passes, and the round log records the counts.
5. The conductor stages new paths with `git add -N -- <paths>` (intent-to-add) only, then
   commits with `git commit -- <paths>`. Intent-to-add records no content, so another lane's
   plain `git commit` cannot sweep a photo in, and `git commit -- <paths>` commits only the named
   paths with full content, leaving anything lane 1 has staged untouched. A withdrawal commits
   the same way: the path list names the record and the deleted `.webp`, and `git commit --`
   records the deletion. The conductor checks `git show --stat HEAD` against the path list,
   retries on `.git/index.lock`, and never deletes the lock.

Rounds continue until the done checkpoint.

## Outreach

The first letters go to Beth Lesser, the British Culture Archive's People's Archive, the DC Punk
Archive at DC Public Library, Charlie Phillips, and Janette Beckman. The identity spec names
Lesser and the British Culture Archive as the strongest reggae leads, and the archive's own terms
("No usage or reproduction of any kind without prior permission") make a letter the only route to
its mod, skinhead, and 2 Tone images. Beckman moves up from the research's tenth rank because she
shot early New York hip-hop, and the hip-hop family has the thinnest open record. Each letter
names specific images and asks for CC BY 4.0 on them. If the answer is no, the fallback is the
optional photo pack, where NC or a dubplate-only permission is acceptable.

The conductor drafts each letter in Geoff's voice: a person building a free music server for a
group of friends, free and open-source under the MIT license. Letters never call the use
non-commercial, as the research template does, because a grant on those terms can only reach the
pack. Drafts live in `docs/superpowers/research/2026-09-24-identity-photos/outreach/`, one file
per recipient. Geoff reads and sends every letter from his own address, and the conductor never
sends one.

Each letter asks the grantor to post the license where it can be checked, such as a Flickr
license field or a line on their own site. That page becomes `source_url`, and the grant is
verified like any other source. A grant that arrives only by email is committed as a grant record
under `outreach/grants/`: the grant text, the named images, the license, the grantor's name, the
date, and the email's `Message-ID`, with no address or other private detail. The record's
`source_url` points at that file.

## Out of scope

The loader, the photo column, error-page and empty-library photos in the web app, CREDITS
generation, rotation order, and the optional pack's packaging all belong to rung 7 or later. This
pass touches nothing under `web/` or `cmd/`, and nothing in `Makefile`. Its only `scripts/` edit
is the queued `check.sh` line (Process).

## Changes to the identity spec

This spec changes these things in the identity spec:

- The pool target rises from 100 or more to about 300.
- Hip-hop becomes a named family with a 20% share, and every era of it counts. The identity spec
  listed "early hip-hop and go-go" among the outward spreads, and its ruling 7 named reggae and
  punk as the core; the core is now reggae, punk, and hip-hop.
- The core families widen. The identity spec's core was reggae, dub, dancehall, and ska, plus
  punk and protopunk. The reggae family now adds rocksteady, 2 Tone, and lovers rock, and the punk
  family adds hardcore and post-punk.
- A public-domain work must be public domain in the US. The identity spec asked only that its US
  status be checked.
- Stored images are resized and re-encoded to WebP, with a recorded source hash, and CREDITS says
  so. The identity spec said the shipped file stays unmodified.
- The manifest is one record file per image, not one manifest. The identity spec's free-text
  `subject` becomes `title`, and `subject` is now a closed facet. Shots, formats, and landmarks
  join the scenes as recorded categories. The per-entry license URL moves into the license list
  in `vocab.json`, keyed by license ID.
- New counted rules: the share bands, the region cap, the floor grid, the posed ceiling, the
  surface floors, one image per landmark, and one image per named artist, not only per headline
  artist.

The identity spec gets a one-line pointer to this spec when this spec is approved.

## Decisions recorded

1. **Size floor (Geoff, 2026-09-25):** "We can lower for genuinely unique or valuble photos. We
   can go as low as 1000px, but the photo needs to justify that." The floor is a 1200px long
   edge, with 1000 to 1199px eligible through `size_exception` and Geoff's keep (Storage). A
   size query over 113 research candidates found these long edges:

   | Long edge | Candidates |
   |---|---|
   | 1600px or more | 62 |
   | 1200-1599px | 15 |
   | 1000-1199px | 6 |
   | under 1000px | 30 |

   The floor keeps 77 of the 113 outright and 6 more by exception. The other 30 need an
   archive's larger scan, among them the Bundesarchiv image at 568×608 and the Tropenmuseum image
   at 700×536.

This spec carries no open rulings.
