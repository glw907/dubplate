# Library filtering, sorting, and browsing: prior art and a design brief

Date: 2026-09-24. Status: research input for the rung 7 web app and the rung 9 plan. Not ratified.

This report sets out how dubplate's web app should let a member filter, sort, and browse a library
of albums. It has three parts. Part 1 lists the fields the store records, or will record, with a
citation for each. Part 2 surveys how the best music software handles large personal libraries,
and what collectors praise and complain about. Part 3 is the design brief, ranked by how much each
recommendation matters to a collector's daily use. The visual identity is settled in
`specs/2026-09-24-visual-identity-design.md`. This report covers only the interaction model and
information design.

The short version: the store will hold the facets a record collector cares about, because every
album passes through a MusicBrainz match that supplies label, catalog number, country, medium
format, release type, and original year. Genre is the exception. No MusicBrainz match supplies it,
and no rung has chosen a source. The prior art agrees on a small set of rules: combine values with
OR inside a facet and AND across facets, show a live count beside every value, let a value be
excluded as well as included, put the whole state in the URL, and keep a query language beside
the controls for people who type faster than they click.

Paths below are relative to `docs/` unless they start with `corpus/`.

## Part 1: what dubplate can filter on

### How a field reaches the store

Short file names in this part (such as `rung-6-ops-seam-store.md`) are the full spec and research names given at first mention.

The store records metadata in typed columns, one per field, added by migration in the rung whose
built-in or engine code first computes the field. It has no attribute bag
(`specs/2026-09-23-rung-6-ops-seam-store.md:524-529`). Rung 6 lands only the catalog skeleton.
The `albums` and `tracks` rows carry opaque ids, the collection reference, the track's `rel_path`,
the MusicBrainz release and recording ids, the album-and-track relation, and timestamps
(`rung-6-ops-seam-store.md:487-489`). Every descriptive field a filter needs (artist, year, label,
country, and the rest) arrives as a column at rung 9, the pipeline rung that tags and admits
albums (`specs/2026-09-02-greenfield-server-build.md:797-814`). Rung 9 closes before launch, so
those columns exist at 1.0. The fields the filter can use are then the tags as applied, which the
2.0 spec names as part of the catalog row
(`specs/2026-09-07-streamer-2-0-design.md:192-193`).

The tags as applied are the 69-key writable set that beets' `apply_metadata` produces from a
MusicBrainz match and that rung 8's pure-Go writer reproduces
(`specs/2026-09-07-rung-8-tag-io.md:417-421`; the list itself is `corpus/README.md:2398-2413`).
Sixteen of those 69 keys carry no value on any corpus subject, because a MusicBrainz match does
not supply them. `genres` is one of the sixteen (`rung-8-tag-io.md:689-697`;
`research/2026-09-07-rung-8-spec-inventory.md:199-209`).

**Planning note.** A filter facet is a store column. So the rung 9 plan, which adds these columns,
must add them for every field Part 3's 1.0 facet set names, not only the fields the layout
template reads. The rung 9 plan should list the facet fields as consumers. This is a 2.0 forward
hook in spirit too, since `getAlbumList2` and smart playlists read the same columns
(`streamer-2-0-design.md:332-335`).

### Candidate filter fields

"Source" names where the value originates. "Store" names where the store records it and which
rung lands the column. Where no document records a field, the row says "not recorded".

| Field | Source | Store | 1.0? | Citation |
|---|---|---|---|---|
| Album artist, with credit and multi-artist list (`albumartist`, `albumartists`) | MusicBrainz release artist credit, applied as tags | Typed column, rung 9 | Yes | `corpus/README.md:2399-2400`; `rung-6-ops-seam-store.md:524-529` |
| Album artist sort name (`albumartist_sort`) | MusicBrainz sort name, written to `ALBUMARTISTSORT` at admit (hook 13) | Typed column, rung 9 | Yes | `streamer-2-0-design.md:489-490`; `corpus/README.md:2399` |
| Folded sort key (accent-folded, uses the sort tags) | Derived at tagging from the matcher's unidecode table | Column, rung 9 (hook 2, erratum 11) | Yes | `streamer-2-0-design.md:462-463`, `:494-499`, `:235-238` |
| Album title (`album`), disambiguation (`albumdisambig`) | MusicBrainz release | Typed column, rung 9 | Yes | `corpus/README.md:2398`, `:2401` |
| Track artist, title, composers, arrangers, lyricists | MusicBrainz recording and work credits | Typed columns, rung 9 | Yes (search scope, not facets) | `corpus/README.md:2402-2407` |
| Release year and date (`year`, `month`, `day`) | MusicBrainz release event | Typed column, rung 9 | Yes | `corpus/README.md:2405`, `:2409`, `:2413` |
| Original year (`original_year`, month, day) | MusicBrainz release group's first release date | Typed column, rung 9 | Yes | `corpus/README.md:2410`; the reissue case at `corpus/README.md:712` shows 2011 release, 1991 original |
| Label (`label`) | MusicBrainz release label | Typed column, rung 9 | Yes | `corpus/README.md:2406`; matcher weight at `research/2026-09-02-beets-matcher-mechanics.md:279` |
| Catalog number (`catalognum`) | MusicBrainz release label info | Typed column, rung 9 | Yes (search, not a facet) | `corpus/README.md:2404`; empty-value rule at `rung-8-tag-io.md:446-453` |
| Release country (`country`) | MusicBrainz release event | Typed column, rung 9 | Yes | `corpus/README.md:2405`; `specs/2026-09-03-rung-3-go-matcher.md:222` |
| Release type (`albumtype`, `albumtypes`: album, EP, single, compilation, live, soundtrack, ...) | MusicBrainz release group primary and secondary types | Typed column, rung 9 | Yes | `corpus/README.md:2401`; write collision noted at `rung-8-tag-io.md:514-515` |
| Release status (`albumstatus`: official, promotion, bootleg) | MusicBrainz release | Typed column, rung 9 | Yes | `corpus/README.md:2401` |
| Compilation flag (`comp`) | MusicBrainz (various-artists credit) | Typed column, rung 9 | Yes | `corpus/README.md:2404`; `streamer-2-0-design.md:209-210` |
| Medium format (`media`: CD, 12" Vinyl, Digital Media, Cassette, ...) | MusicBrainz medium format | Typed column, rung 9 | Yes | `corpus/README.md:2409`; matcher preference `rung-3-go-matcher.md:229` |
| Disc count (`disctotal`) | MusicBrainz release | Typed column, rung 9 | Yes | `corpus/README.md:2405`; `streamer-2-0-design.md:211` |
| Language, script (`language`, `script`) | MusicBrainz release text representation | Typed column, rung 9 | Yes (low value) | `corpus/README.md:2406`, `:2412` |
| Barcode, ASIN | MusicBrainz release | Typed column, rung 9 | Yes (search only) | `corpus/README.md:2403` |
| MusicBrainz release id, release group id, artist ids | MusicBrainz | Release id at rung 6 (`mb_<entity>_id`); the rest rung 9 | Yes | `rung-6-ops-seam-store.md:449-451`, `:487-489`; `corpus/README.md:2407-2408` |
| Collection | The store: `albums` references `collections` | Rung 6 skeleton | Yes | `rung-6-ops-seam-store.md:479-489` |
| Date added | The store: album row timestamps | Rung 6 skeleton | Yes | `rung-6-ops-seam-store.md:489` |
| Added by (member) | The upload record: the authenticated uploader on the submission (rung 7), carried to the album row as `added_by` at rung 9 admission; intake is the only writer | Column, rung 9 (hook 17) | Yes; migrated albums are null, shown as "unknown" | `streamer-2-0-design.md:320-322`, `:512-514` |
| Genre (`genres`) | Not supplied by a MusicBrainz match. The built-ins survey recommends shipping a genre built-in; the roster is decided at rung 11 under a "scrapers refused" policy | Not recorded until rung 11 decides a source | Conditional | `rung-8-tag-io.md:689-697`; `research/2026-08-31-beets-plugin-usage.md:33`; `specs/2026-09-01-dubplate-organizer-design.md:354-367`; `greenfield-server-build.md:843-848` |
| Style (Discogs-sense subgenre) | Not recorded. beets keeps `style` only as a Discogs info-level name that never becomes an item field | Not recorded | No | `research/2026-09-07-rung-8-spec-inventory.md:189-191` |
| Audio format | The gate's allowlist; the estate is FLAC-only | No column needed at 1.0 | Facet is meaningless at 1.0 | `greenfield-server-build.md:722-730`; `specs/2026-09-07-rung-8-tag-io.md:5` |
| Bit depth, sample rate | Read-only audio properties of the file (`bitdepth`, `samplerate`) | Not recorded. They are outside the 69 writable keys and no spec names a column | No, until a rung adds the column | `research/2026-09-07-rung-8-spec-inventory.md:167-172` |
| ReplayGain | Probe-backed built-in | Column, rung 9 (hook 11) | Not a filter | `streamer-2-0-design.md:484-485` |
| Starred, play count, last played | Per-member play history, 2.0 | 2.0 tables (hook 4) | No (2.0) | `streamer-2-0-design.md:98-100`, `:333-335`, `:466-467` |
| Album note | 2.0 table, written by the `added_by` member or an admin | 2.0 | No (2.0); searchable later | `streamer-2-0-design.md:322-324` |
| Ratings | 2.1 | Not recorded | No | `streamer-2-0-design.md:561-562` |
| Visibility (group or private) | Collection row | Rung 6 | Private is refused until the driver supports per-user folders, so every visible album is group-visible at 1.0 | `rung-6-ops-seam-store.md:479-482`; `configuration-homes-design.md:196-198` |

### Facts that constrain the design

1. **1.0 is FLAC-only.** Format is config, but the estate admits FLAC alone and rung 8 writes
   FLAC alone (`greenfield-server-build.md:722-730`; `rung-8-tag-io.md:5`). A format facet has
   one value at 1.0 and would be noise. It joins when a second format does (the formats 1.1
   initiative).
2. **Collections are store rows.** A collection is a row with a name, owner, visibility, and
   writer, and every album references one (`configuration-homes-design.md:185-206`;
   `rung-6-ops-seam-store.md:479-489`). Collection is therefore a cheap, exact facet. An album in
   two collections is two albums by design (`streamer-2-0-design.md:207-209`), so a collection
   facet with multi-select can show the same release twice.
3. **Each album records who added it.** `added_by` is set only by intake, from the session
   (`streamer-2-0-design.md:320-322`). "Added by" is a trustworthy facet, and it is the one facet
   no rival music server can offer, because none owns intake. Migrated albums carry null, so the
   facet needs an "Unknown" value.
4. **Two pressings are two albums.** The album is a release MBID within a collection
   (`streamer-2-0-design.md:205-210`). A collector who owns the 1991 and 2011 *Nevermind* sees two
   sleeves. The release group id is recorded, so a "group pressings" toggle is possible.
5. **Not every album has a MusicBrainz match.** An album with no release MBID is identified by a
   folded tuple (`streamer-2-0-design.md:205-207`). Every MusicBrainz-sourced facet needs a
   visible "Unknown" bucket rather than silently dropping those albums.
6. **Sort order is a member preference.** The configuration homes spec classifies "sort order the
   server applies to a listing" as a per-member preference (`configuration-homes-design.md:127-129`),
   stored in `member_preferences` (`rung-6-ops-seam-store.md:495-496`). Saved views have a natural
   home in the same table.
7. **Search already has an engine.** FTS5 with the trigram tokenizer over accent-folded text is
   adopted, gated by a rung 6 test (`rung-6-ops-seam-store.md:902`;
   `streamer-2-0-design.md:233-240`). The web filter's free-text box and `search3` should share
   that index so the web app and every Subsonic client agree on what a query finds.
8. **Typed columns, never a bag.** Adding a filterable field later is a migration plus a typed
   accessor in the rung that owns it (`rung-6-ops-seam-store.md:526-529`). The design should name
   its fields up front so rung 9 adds them once.

## Part 2: prior art

Each entry records what works, what collectors complain about, and one interaction to adopt or
avoid. A claim marked "unverified" came from a search summary rather than a page read directly.

### Roon: Focus and bookmarks

What works:

- Focus filters the library by any number of criteria at once, such as "1980s Pop" or "24-bit Jazz
  recordings" ([Roon: music organization](https://roon.app/en/music/organization);
  [Feature spotlight: Focus](https://blog.roonlabs.com/roon-feature-spotlight-focus/)).
- A criterion can be inverted to exclude matches instead of including them
  ([Feature spotlight: Focus](https://blog.roonlabs.com/roon-feature-spotlight-focus/)).
- A Focus can be saved as a bookmark, and the bookmark stays live: albums added later that match
  the criteria appear in it ([Tags/bookmarks question](https://community.roonlabs.com/t/tags-bookmarks-question/107464);
  [Focus, bookmarks, tags and playlists](https://community.roonlabs.com/t/email-focus-bookmarks-tags-and-playlists/104522)).
- The same Focus works in the mobile app, Roon ARC
  ([Focus in ARC](https://community.roonlabs.com/t/focus-feature-in-roon-arc-is-live/318737)).

What collectors complain about:

- Roon 1.8 changed how two values in one category combine, from OR to AND. A "Rock" plus "Jazz"
  selection stopped meaning "either". One member's saved "Smooth jazz" bookmark dropped from over
  300 albums to one, and users still ask for an AND/OR choice
  ([Make Focus an OR operation](https://community.roonlabs.com/t/make-focus-an-or-operation-instead-of-and/140860)).
- Favorites can be included but not excluded
  ([Focus: need an option for NOT favorites](https://community.roonlabs.com/t/focus-need-an-option-for-not-favorites/241344)).
- Roon removed the setting to ignore leading articles when sorting, so article handling is fixed
  ([Artist and album names with "The"](https://community.roonlabs.com/t/artists-album-names-with-the/320738);
  [Roon FAQ on compilation sorting](https://help.roonlabs.com/portal/en/kb/articles/faq-my-compilations-and-soundtracks-aren-t-sorted-by-artist-name-in-the-album-browser-can-i-change-that)).

**Adopt:** the live saved view. **Avoid:** changing how filters combine between releases. Once a
member has saved views, the combining rule is a contract.

### Discogs: collection and marketplace facets

What works:

- The collection view filters by folder, artist, format, label, year, genre, style, country, and
  custom fields, and the filters stack
  ([Discogs forum: collection filters](https://www.discogs.com/forum/thread/735729)).
- Marketplace search shows facets with counts for country, format, genre, style, format
  description, condition, and year. The combining rule is not stated in any help page this survey
  read (unverified).
- Genre and style form a governed two-level taxonomy, so facet values stay consistent across the
  catalog ([Database guidelines 9: genres and styles](https://support.discogs.com/hc/en-us/articles/360005055213-Database-Guidelines-9-Genres-Styles)).
- Custom fields and folders let collectors add their own layer, such as shelf location
  ([How to categorize your collection](https://www.discogs.com/forum/thread/1026837)).

What collectors complain about:

- Years of threads ask for better sorting and combined genre-and-format filtering in the personal
  collection, which lags the marketplace's own facets
  ([Sort by genre in collection](https://www.discogs.com/forum/thread/372692);
  [Sort collection by genre and format](https://www.discogs.com/forum/thread/410031);
  [Search by genre/style](https://www.discogs.com/forum/thread/220899)).

**Adopt:** label, country, format, and year as first-class facets, because those are the fields
Discogs collectors search by. **Avoid:** giving the personal library weaker filtering than the
public catalog. In dubplate the library is the only surface, so it must carry the full set.

### foobar2000: Facets, Album List, and query syntax

What works:

- The Facets component shows several linked columns side by side, each narrowing the next, with
  art, statistics, and search ([foobar2000 components: Facets](https://www.foobar2000.org/components/view/foo_facets);
  [Hydrogenaudio wiki: Facets](https://wiki.hydrogenaudio.org/index.php?title=Foobar2000%3AComponents%2FFacets_%28foo_facets%29)).
- Each column is defined by a title-formatting pattern, so a facet can be built on any tag
  ([Title formatting reference](https://wiki.hydrogenaudio.org/index.php?title=Foobar2000%3ATitle_Formatting_Reference)).
- A separate query syntax powers ad hoc searches and autoplaylists (same reference).
- Sorting without "The" is a pattern the user writes, not a fixed rule
  ([Head-Fi: sorting without "The"](https://www.head-fi.org/threads/sorting-without-the-the-prefix-in-front-of-artists-in-the-foobar-playlist.495919/)).

What collectors complain about: useful facets need hand-written patterns. This is an inference
from the configuration model; the survey found no dedicated complaint thread.

**Adopt:** linked narrowing, where each facet's values reflect the others' selections. **Avoid:**
making the member define facets. dubplate's schema is typed and known, so facets can ship ready.

### beets: the query language

What works ([beets: queries](https://beets.readthedocs.io/en/stable/reference/query.html)):

- `field:value` matches a substring in one field, as in `artist:dream` or `year:2012`.
- Two dots give a range: `year:1990..1999`, `..7`, or `4..`.
- A doubled colon gives a regular expression: `artist::Ann(a|ie)`.
- A leading `-` or `^` negates any term: `^love`, `-genre:soundtrack`.
- A field name with `+` or `-` sorts, and several chain: `genre+ year+`.

What collectors complain about: the language exists only on the command line, with no visual
counterpart. This is inferred from beets' design, not from a complaint thread. dubplate's own
beets research found related sharp edges in untyped fields: an unknown-field query flipped from
"match everything" to "match nothing" when flexible attributes arrived, and uppercase attribute
names were unqueryable for years (`research/2026-09-02-beets-evolution-lessons.md:192-198`).

**Adopt:** `field:value`, ranges, quoting, and negation as dubplate's query syntax, over typed
fields only. **Avoid:** regular expressions at 1.0. The search index is trigram FTS5
(`specs/2026-09-07-streamer-2-0-design.md:235-238`), which cannot use an index for a regex.

### Plex and Plexamp

What works:

- Library filters combine, and an advanced filter builds a condition from several fields
  ([Plex: using the library view](https://support.plex.tv/articles/200392126-using-the-library-view/)).
- Smart playlists and advanced filters nest "match all" and "match any" groups, so AND and OR are
  both explicit ([Kometa: Plex smart filter](https://kometa.wiki/en/latest/files/builders/plex/smart-filter/);
  [Ryan Cragun: Plex smart music playlists](https://ryancragun.com/2019/08/26/plex-how-to-create-smart-auto-updating-music-playlists/)).
- Sonic analysis extracts about 50 audio features per track and powers mood and decade radio and
  "Sonic Adventure", a path between two tracks
  ([Plex: sonic analysis](https://support.plex.tv/articles/sonic-analysis-music/);
  [TechCrunch on sonic features](https://techcrunch.com/2021/08/12/plexs-new-feature-matches-your-sonically-similar-music-to-make-playlists)).

What collectors complain about:

- Album style and mood come from Plex's metadata agent and vary with match quality; track mood is
  often thin ([Kometa: Plex smart filter](https://kometa.wiki/en/latest/files/builders/plex/smart-filter/)).
- Several discovery features sit behind the Plex Pass subscription
  ([Sonos community: appraisal of Plexamp](https://en.community.sonos.com/controllers-and-music-services-229131/appraisal-of-plex-and-sonos-particularly-plexamp-6908205)).

**Adopt:** explicit "match all" and "match any" in the smart playlist editor, which 2.0 owes
(`specs/2026-09-07-streamer-2-0-design.md:332-336`). **Avoid:** facets fed by a scraped source
whose coverage varies by album. A half-filled facet misleads, because an album missing from
"Style: Dub" may simply lack the tag.

### Navidrome

What works:

- The web UI filters albums, songs, and artists by genre, and filtering by several genres at once
  was added with multi-valued genre support
  ([PR #3677: filter on multiple genres](https://github.com/navidrome/navidrome/pull/3677);
  [PR #1251: multiple genres](https://github.com/navidrome/navidrome/pull/1251)).
- `IgnoredArticles` strips a configured list of leading words for sorting; its default is
  `The El La Los Las Le Les Os As O A`
  ([discussion #5439](https://github.com/navidrome/navidrome/discussions/5439);
  [conf/configuration.go](https://github.com/deluan/navidrome/blob/master/conf/configuration.go)).
- `PreferSortTags` makes the file's `ARTISTSORT` and `ALBUMARTISTSORT` win over the article list
  (discussion #5439).
- Smart playlists exist as `.nsp` rule files
  ([Navidrome: smart playlists](https://www.navidrome.org/docs/usage/features/smart-playlists/)).

What collectors complain about:

- The band As It Is sorted under "I", because the article list ran before the sort tags were
  consulted ([discussion #5439](https://github.com/navidrome/navidrome/discussions/5439)).
- Users ask for a genre browsing page, not only a filter dropdown
  ([issue #1155](https://github.com/navidrome/navidrome/issues/1155)).
- Long album and playlist names truncate
  ([issue #5068](https://github.com/navidrome/navidrome/issues/5068)).

**Adopt:** sort tags win over any heuristic. **Avoid:** an article heuristic that runs before an
explicit sort name. dubplate writes MusicBrainz sort names at admit
(`specs/2026-09-07-streamer-2-0-design.md:489-490`), so the heuristic is needed only for albums
with no match.

### Jellyfin

What works: genre filtering on the library grid works reliably where other filters do not
([jellyfin-web #7620](https://github.com/jellyfin/jellyfin-web/issues/7620)).

What collectors complain about:

- In 10.11, applying a filter on a large library could hang until the request timed out. The
  thread traces it to the new data layer loading whole datasets into memory before filtering
  ([jellyfin #14285](https://github.com/jellyfin/jellyfin/issues/14285)).
- The tags filter takes up to 20 seconds to fill on NAS hardware
  ([jellyfin #14380](https://github.com/jellyfin/jellyfin/issues/14380)); disabling filters restores
  fast loading on Android TV ([jellyfin-androidtv #3073](https://github.com/jellyfin/jellyfin-androidtv/issues/3073)).
- A stale saved filter can lock a user out of a library with no way to clear it from the UI
  ([jellyfin-web #6283](https://github.com/jellyfin/jellyfin-web/issues/6283)).
- Most filters fail on movie collections, so coverage differs by view
  ([jellyfin-web #7620](https://github.com/jellyfin/jellyfin-web/issues/7620)).

**Avoid:** filtering in application memory, and any saved state the UI cannot clear. Both are
the failures a 10,000-album library exposes first.

### MusicBee, and the iTunes column browser

What works:

- MusicBee's column browser shows up to six linked columns, each bound to any tag
  ([MusicBee wiki: library navigation](https://musicbee.fandom.com/wiki/Library_Navigation_Elements)).
- Its filters combine rules under an explicit ANY or ALL choice
  ([MusicBee wiki: filters](https://musicbee.fandom.com/wiki/Filters)).
- A genre category tag rolls many raw genres into a few broad ones
  ([MusicBee wiki: tagging](https://musicbee.fandom.com/wiki/Tagging)).

What collectors complain about: the iTunes column browser filtered by three linked columns at
once, and its loss in Apple's Music app is the most repeated complaint in this survey. One user
wrote that "users could filter by three things simultaneously, but now they can only filter by
one thing" ([Apple Community: Music column browser](https://discussions.apple.com/thread/250722825?page=2);
[Where did the column browser go](https://discussions.apple.com/thread/250720666);
[MacRumors: Music missing features](https://forums.macrumors.com/threads/music-missing-features-bugs-issues-troubleshooting.2189298/);
[Apple developer forums](https://developer.apple.com/forums/thread/117308)).

**Adopt:** linked narrowing across several facets at once. **Avoid:** replacing multi-facet
filtering with a single search box. Apple did, and collectors still ask for it back.

### Bandcamp collection

What works: the collection page lists every purchase and wishlist item with instant streaming and
a public view ([essay on Bandcamp collections](https://medium.com/@asip/bandcamp-ideas-for-the-future-of-collections-and-the-relationship-between-artists-and-labels-99e30a147d8d)).

What collectors complain about:

- Sorting is capped at about 1,000 items, and Bandcamp itself hosts a page acknowledging the
  limit ([Bandcamp: I want my collection sorted](https://bandcamp.com/iwantmycollectionsorted)).
- A third-party tool exists only to re-sort a collection
  ([bandcamp-collection-sorter](https://github.com/darnfish/bandcamp-collection-sorter)).

**Avoid:** any ceiling on how many albums a sort or filter covers.

### Rate Your Music

What works (the charts FAQ refused direct fetch, so these rest on search summaries; unverified):

- Custom charts filter by year range, genre, and country together, and include or exclude
  release types such as live albums, archival releases, and soundtracks
  ([RYM charts FAQ](https://rateyourmusic.com/wiki/RYM:Charts+FAQ)).
- Releases carry primary and secondary genres, voted by the community, plus descriptors such as
  "melancholic" or "hazy" ([RYM wiki: add a genre](https://rateyourmusic.com/wiki/Music:Add%20a%20genre)).

What collectors complain about: the survey found no thread about the filter UI itself.

**Adopt:** release type as an include-and-exclude facet. "Hide live albums and compilations" is
the single most common narrowing a collector makes, and RYM puts it one click away.

### Apple Music

Apple Music is the cautionary case. Separate threads across several OS versions report the same
missing column browser ([thread 6646291](https://discussions.apple.com/thread/6646291);
[thread 250740428](https://discussions.apple.com/thread/250740428);
[thread 250726761](https://discussions.apple.com/thread/250726761);
[thread 250871885](https://discussions.apple.com/thread/250871885)). Where a remnant exists, it
can show results that do not match the filter
([thread 255484635](https://discussions.apple.com/thread/255484635)).

**Avoid:** trading multi-facet filtering for a simpler surface.

### General faceted-search practice

- Nielsen Norman Group recommends facets that follow the user's mental model rather than the
  schema, and dynamic facets that offer only values that would return results
  ([NN/g: ecommerce search and faceted search](https://www.nngroup.com/reports/ecommerce-ux-search-including-faceted-search/)).
- Baymard recommends a live count beside each value, AND across filter types and OR within one,
  a clear link per facet and for all facets, and on phones a full-screen filter panel with a
  button that states the result count ([Baymard: filter UI best practices](https://baymard.com/blog/ecommerce-filter-ui)).
- GOV.UK announces only the result count through an `aria-live="polite"` region, never the whole
  result list, so a screen reader is not flooded on each change
  ([GOV.UK: improving accessibility on search](https://technology.blog.gov.uk/2014/08/14/improving-accessibility-on-gov-uk-search/);
  [MDN: aria-live](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Reference/Attributes/aria-live)).
- SQLite's trigram tokenizer documents that "substrings consisting of fewer than 3 unicode
  characters do not match any rows when used with a full-text query"
  ([SQLite: FTS5 trigram tokenizer](https://www.sqlite.org/fts5.html#the_trigram_tokenizer)).
  A search for "U2" or "X" needs another path, which the 2.0 spec's separate prefix pass provides
  (`specs/2026-09-07-streamer-2-0-design.md:237-238`).

### The three strongest examples

1. **Roon Focus.** Multi-facet filtering with include and exclude, saved as live bookmarks, is the
   closest existing thing to what a collector wants. Its OR-to-AND regression shows the combining
   rule is a contract.
2. **The iTunes column browser, and MusicBee's heir to it.** Linked narrowing across several
   facets is the pattern collectors mourn most when it is taken away.
3. **beets queries.** `field:value`, ranges, and negation give a compact syntax that maps directly
   onto facet state and onto a URL.

## Part 3: design brief

### Ranked recommendations

Ranked by effect on a collector's daily use. "Needs" names store fields or tables that are
specified but not yet built (planned), or that no document records (missing).

1. **Every view has a URL, and every value on an album page is a link.** The URL carries the full
   filter, sort, and layout state, so a view can be bookmarked, reloaded, and pasted to another
   member. On the album page, the artist, label, catalog line, year, country, medium, and "added
   by" are each a link to the grid filtered by that value. Discogs trained collectors to browse
   this way, and it costs little. Needs: the rung 9 columns (planned).
2. **Search that finds what collectors type.** One box searches album title, album artist, track
   artists, track titles, label, and catalog number, accent-folded, over the FTS5 trigram index
   that `search3` also uses, so the web app and every Subsonic client agree. Catalog numbers match
   with spaces and hyphens ignored, so "BLP 1577" finds "BLP-1577". Terms under three characters
   take the prefix path. Needs: the rung 9 columns and the folded key (planned).
3. **Sort by artist uses MusicBrainz sort names.** "Bowie, David" and "Beatles, The" come from the
   `ALBUMARTISTSORT` tag written at admit, so personal names and articles are handled without a
   heuristic. The article list (Navidrome's default is a fair start) applies only to an album with
   no sort name. Secondary sort within an artist is original year, then title. Needs: hook 13 and
   the folded sort key (planned).
4. **OR within a facet, AND across facets, and exclude per value.** Picking "Label: Blue Note" and
   "Label: Prestige" shows both labels. Adding "Country: US" narrows both. Any value can be
   excluded instead, shown as a struck chip ("not Compilation"). The rule is stated in the UI's
   help text and never changes. Roon's regression shows why. Needs: nothing new.
5. **The 1.0 facet set.** Release type, year (original year by default), label, country, medium,
   collection, and added by. Each shows an "Unknown" value for albums without the field. Genre
   joins when rung 11 chooses a source. See the facet table below. Needs: the rung 9 columns
   (planned); `added_by` (planned, hook 17).
6. **Live counts beside every value.** Counts are disjunctive: a value's count assumes the rest of
   its own facet's selection is cleared, so "Label: Prestige (38)" still shows while Blue Note is
   selected. Values with a zero count are hidden, unless selected. Long facets (label, country)
   show the top ten by count and a "Show all" that opens a searchable list. Needs: nothing new.
7. **Saved views.** A member names the current filter and sort, such as "Blue Note, 1955 to 1967,
   no reissues", and it stays live as albums arrive. Saved views are the member's own by default.
   Sharing works by link at 1.0; a "Share with the group" switch that lists a view for everyone
   comes later. A saved view that names a removed collection or member opens with that chip
   marked and removable, never as a locked or empty page (Jellyfin's lesson). Needs: the
   `member_preferences` table (planned, 6c) or a small table of its own; group sharing needs a
   visibility column (missing).
8. **A query syntax that stays in sync with the controls.** A member can type
   `label:"Blue Note" year:1955..1967 -type:compilation country:US` into the search box. Each
   recognized `field:value` term becomes a chip; bare words stay free-text search. Clicking a
   facet writes the same term into the box. The query string is the canonical form: the URL holds
   it, and a saved view stores it. An unknown field or a bad range shows an inline error and is
   never silently dropped (beets' flipped-default lesson). No regular expressions at 1.0. Needs:
   nothing new.
9. **Fast at 10,000 albums with no ceiling.** Filtering, counting, and sorting run as SQL in the
   store, never in Go memory or in the browser. The budget is a p95 under 100 ms for a result page
   plus all facet counts over 10,000 albums and about 120,000 tracks, on the estate's box. Every
   facet column is indexed. The grid pages or virtualizes, and nothing caps how far a member can
   scroll or how many albums a sort covers. Needs: indexes on the rung 9 columns (planned).
10. **Phone, keyboard, and screen reader parity.** On a phone, a sticky "Filters" button with the
    active count opens a full-screen panel whose close button reads "Show 42 albums". On a
    desktop, `/` focuses search, Escape clears the box, and the grid moves by arrow keys. Each
    facet is a `fieldset` with a `legend`; only the result count is announced, through a polite
    live region. Needs: nothing new.

Also worth doing, below the top ten:

- **An empty result state that helps.** "No albums match these filters." with each active chip
  removable in place, a "Clear all" button, and, when one removal would bring albums back, a line
  such as "Remove *Country: JP* to see 12 albums." The counts make this cheap. An empty library
  keeps its own sentence, "No albums yet."
- **A "one per release" switch.** Two pressings are two albums by design
  (`specs/2026-09-07-streamer-2-0-design.md:209`). A switch groups albums by release group id and
  shows one sleeve with a pressing count. Needs: `mb_releasegroupid` as a column (planned with the
  rung 9 tag columns).
- **A year choice.** The year facet and year sort use the original year by default, with a
  switch to release year. A 2011 reissue of a 1991 album belongs in the 1990s for browsing and in
  2011 for "what pressing is this". Needs: both year columns (planned).

### The facet set

| Facet | Field | Source | Control | At |
|---|---|---|---|---|
| Release type | `albumtypes` | MusicBrainz release group types | Multi-select with exclude; values: Album, EP, Single, Compilation, Live, Soundtrack, Remix, and the other secondary types | 1.0 |
| Year | `original_year`, switchable to `year` | MusicBrainz | Range (two year fields) plus decade shortcuts; "Unknown" as a checkbox | 1.0 |
| Label | `label` | MusicBrainz release label | Multi-select with exclude, searchable list | 1.0 |
| Country | `country` | MusicBrainz release event | Multi-select with exclude, shown as country names | 1.0 |
| Medium | `media` | MusicBrainz medium format | Multi-select with exclude (CD, Vinyl, Digital Media, Cassette) | 1.0 |
| Collection | collection reference | The store | Multi-select; hidden when the group has one collection | 1.0 |
| Added by | `added_by` | The upload record, set only by intake | Multi-select; "Unknown" for migrated albums | 1.0 |
| Status | `albumstatus` | MusicBrainz | Collapsed under "More"; multi-select with exclude (Official, Promotion, Bootleg) | 1.0 |
| Genre | `genres` | Not recorded; rung 11 decides a source | Multi-select with exclude | When rung 11 lands one |
| Format | file format | The gate | Multi-select | When a second format ships (formats 1.1) |
| Bit depth and sample rate | `bitdepth`, `samplerate` | File properties | Multi-select ("24-bit", "96 kHz") | When a rung adds the columns |
| Starred, played, not played | play history, stars | 2.0 per-member state | Checkboxes | 2.0 |
| Rating | ratings | 2.1 | Minimum rating | 2.1 |

Catalog number, barcode, and disambiguation are search terms, not facets: their values are nearly
unique, so a facet list of them is a list of albums.

### Combining rules

- Within a facet, selected values combine with OR. Across facets, they combine with AND. An
  excluded value removes matching albums whatever else is selected.
- A multi-valued field (release types, and genres later) matches when any of the album's values
  matches. "Type: Live" and "not Compilation" together keep a live album and drop a live
  compilation.
- Free text ANDs with the facets.
- Explicit "match all" and "match any" groups belong to the 2.0 smart playlist editor, not to the
  everyday filter. The everyday filter keeps one fixed rule so a saved view never changes meaning.

### Free-text search and the query syntax

Search scope: album title, album artist and credit, track artists, track titles, label, catalog
number, and barcode. Album notes join at 2.0. Composers join as a field term (`composer:`), not the
default scope, because classical credits would crowd pop results.

Query terms at 1.0:

| Form | Meaning | Example |
|---|---|---|
| `word` | Free-text search | `coltrane` |
| `field:value` | Field contains the value | `label:impulse` |
| `field:"two words"` | Quoted value | `label:"Blue Note"` |
| `field:a..b`, `field:a..`, `field:..b` | Numeric range | `year:1965..1972` |
| `-term` | Exclude | `-type:compilation` |
| `sort:field` or `sort:-field` | Sort, descending with `-` | `sort:-added` |

Field names are plain words: `artist`, `album`, `label`, `catalog`, `country`, `type`, `media`,
`year`, `released` (release year), `added`, `by` (added by), `collection`, `status`. A help link
beside the box lists them. The box and the chips are two views of one state: the parser turns
the string into chips, the chips serialize back to the string, and the round trip is a unit test.

### Sorting

| Sort | Primary key | Then |
|---|---|---|
| Artist (default) | Album artist sort name | Original year, then album title |
| Album title | Folded title, articles kept | Artist sort name |
| Year | Original year (release year when switched) | Artist sort name, then title |
| Date added (newest first) | Album row creation time | Artist sort name |
| Label | Label, then catalog number in natural order | Original year |
| Shuffle | Random, stable within one page load | (none) |

Compilations sort under their MusicBrainz album artist, which is "Various Artists" for most. The
label sort is shelf order for a record collector: catalog numbers compare numerically within a
prefix, so "BLP 1577" comes before "BLP 4003". The member's last sort is a preference
(`specs/2026-09-11-configuration-homes-design.md:127-129`).

### Saved views

Call them "saved views" in the UI. A member saves the current state with a name; the view lists
in the sidebar (a drawer on phones) and opens its URL. A saved view stores the query string, so
it survives a schema change that adds a facet. At 1.0 a view belongs to the member who made it.
Group sharing, where a view appears for every member, fits the group's character and can follow
once a view carries a visibility. Until then, sharing a link does the same job without new state.

### URL state

Use one readable query parameter for the filter and one each for sort and layout:
`/library?q=label%3A%22Blue+Note%22+year%3A1955..1967&sort=artist&view=grid`. A facet click
pushes a history entry, so Back undoes it. Typing replaces the current entry, so Back does not
step through every keystroke. The app is static and client-rendered, so the page reads the URL on
load and on `popstate` and asks the API for results.

### Performance

- One API call returns a page of albums, the total, and the counts for every facet. Counts are a
  `GROUP BY` per facet over the filtered set, excluding that facet's own selection.
- Index every facet column and the sort keys. At 10,000 albums these queries are small for
  SQLite, but the Jellyfin reports show what happens when filtering leaves the database.
- Add a gate test that seeds 10,000 albums and asserts the budget, so a later change cannot slow
  it silently.

### Phone pattern

A sticky bar holds the search box and a "Filters" button showing the number of active filters.
The button opens a full-screen panel (a DaisyUI modal at full height), with facets as collapsible
sections, a "Clear all" link, and a primary button that reads "Show 42 albums" and updates live.
Active filters show as removable chips above the results, in a row that wraps rather than
scrolling sideways. Sort sits in the same panel.

### Keyboard support

- `/` focuses the search box; Escape clears it, then leaves it.
- Tab reaches the facets in order; Space toggles a checkbox; each value's exclude button is its
  own tab stop with `aria-pressed`.
- The grid is a list of links; arrow keys move between sleeves, and Enter opens the album.
- No shortcut fires while focus is in a text field.

### Accessibility

- Each facet is a `fieldset` with a `legend`, and each value is a native checkbox whose label
  includes its count, as in "Blue Note, 42 albums".
- Include and exclude are two controls, a checkbox and a toggle button, never a three-state
  checkbox, whose third state screen readers announce poorly.
- A polite live region announces only "42 albums", after the result arrives, and never the list.
  Focus stays where the member was.
- Chips are buttons labeled "Remove Label: Blue Note".
- The empty state is a heading plus text, reachable in reading order after the filters.

### What the rung 9 plan needs from this brief

The 1.0 facets and sorts need these columns on the album row, each indexed: album artist and its
sort name, the folded sort key, album title, original year, release year, label, catalog number,
country, release types, status, medium, release group id, and `added_by`. Track rows need artist
and title for search. Every one is in the 69-key writable set or a named hook, so none is new
data. What is new is that the filter is their consumer, and the rung 9 plan should name it.
