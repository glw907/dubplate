# Rung 8 Spec Inventory

Research report, 2026-09-07, prepared for the rung 8 spec's authoring,
against HEAD `5543d248818b0c2b55ff169ce3fb7ff10d5ab3d3`. A code lane
(pass 4b) was committing in the same tree while this was written; it
owns `scripts/deploy.sh`, `systemd/`, `env/`, the `Makefile`,
`cmd/dubplate-server/main.go`, and comment lines under
`internal/engine/shadow`. Nothing below reads those files for anything
load-bearing, and no gate was run.

Scope, quoted from the build spec's ladder entry
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:694-701`):

> **Tag I/O** (M). Pure-Go FLAC tag writing for the server scope,
> proven against the corpus's expected applied tag sets from rung 2,
> plus the wazero-taglib spike whose recorded decision the post-launch
> local organizer will consume. Writing tags is the pipeline's primary
> side effect, so this rung precedes the pipeline rather than folding
> into it. The rung's plan opens with the write-side close read the
> mechanics study scoped out (the apply-metadata field mapping and the
> mediafile layer).

Two of that entry's clauses are already superseded by owner rulings
taken on 2026-09-07 and are treated as such throughout. The
wazero-taglib spike is retired, and F1 (the 2.0 spec's ruling at
`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md:445-449`) is
the recorded decision the post-launch organizer consumes. Rung 8 is
FLAC only, behind a per-format seam shaped so MP3 and MP4 writers land
as their own later passes. The build spec gets a dated erratum at the
rung's close.

One sibling document was written in parallel and is a standing input
rather than a gap this one fills:
`docs/superpowers/research/2026-09-07-rung-8-container-needs.md`
answers the container-ordering question owner ruling 1 hands to
research. Its recommendations are carried into the candidate rulings
below rather than restated here.

Sections 1 and 2 are the write-side close read the ladder entry asks
for. Sections 3 through 8 are the as-landed state of everything the
spec touches. Section 9 is budget evidence and a pass-split proposal.
Everything is measured against the tree at that HEAD and against the
pinned venv at `tools/venv/.venv`, never against memory. Where a claim
came from running something, the command and its output are named.

---

## 1. The write side of beets, read in the pinned venv

The venv is `tools/venv/.venv/lib/python3.13/site-packages`, holding
`beets` 2.13.1, `mediafile` 0.17.0, and `mutagen` 1.48.1. Every line
citation in this section and the next is relative to that directory.

### 1.1 `apply_metadata` in 2.13.1 is a method, not a module function

The mechanics study's write-side scope-out named "the apply-metadata
field mapping". In 2.13.1 there is no module-level
`beets.autotag.apply_metadata`. There are two methods, one per match
kind, on the dataclasses in `beets/autotag/match.py`:

- `AlbumMatch.apply_metadata` (`beets/autotag/match.py:186-197`) walks
  `self.merged_pairs` and calls `item.update(data)` per item, clearing
  the item first when `from_scratch` resolves true.
- `TrackMatch.apply_metadata` (`:229-239`) does the same for one item
  against `self.info.item_data`.
- `AlbumMatch.apply_album_metadata` (`:199-201`) updates the `Album`
  object from `self.info.item_data`. It touches no file.

`from_scratch` resolves from the `from_scratch` argument when given and
from `config["import"]["from_scratch"]` otherwise (`:102-110`).
`Item.clear()` sets every field in `_media_tag_fields` to `None`
(`beets/library/models.py:887-890`), which matters downstream: a
`None` reaching `MediaFile.update` deletes the tag rather than writing
an empty one.

### 1.2 The value a track actually receives

`AlbumMatch.merged_pairs` (`match.py:178-184`) pairs each item with
`TrackInfo.merge_with_album(album_info)`. That method
(`beets/autotag/hooks.py:498-521`) is the whole album-to-track fallback
rule, and it composes three dictionaries in this order:

```
album_info.item_data | {"tracktotal": len(album_info.tracks)} | track.item_data
```

Before composing, it fills each falsy track key from the album's
`raw_data`, with `artist_credit` explicitly excluded from that
inheritance (`:509-515`). So the track's own value wins, then the
literal track count, then the album's.

`Info.raw_data` (`:174-182`) is a copy of the info mapping with the
artist-credit rewrite applied when `config["artist_credit"]` is set:
`artist` becomes `artist_credit or artist`, and `artists` becomes
`artists_credit or artists`. `AlbumInfo.raw_data` (`:299-306`) adds one
rule: when `year` is truthy, `month` and `day` are forced to their own
values or `0`. `TrackInfo.raw_data` (`:420-439`) sets
`mb_releasetrackid` from `release_track_id or track_id`, sets `track`
to `self.index`, and sets `medium_index` to the medium index under
`per_disc_numbering` and to `self.index` otherwise, adding
`tracktotal = medium_total` when `per_disc_numbering` is on and a
medium total exists.

`Info.item_data` (`:184-202`) is `raw_data` filtered and renamed. It
drops `IGNORED_FIELDS`, drops `None` and `[]` values unless the key is
in `nullable_fields` (which comes from
`config["overwrite_null"][<type>]`, `:145-148`), and then applies
`MEDIA_FIELD_MAP`, which is where the album-to-item name change
happens:

- `AlbumInfo.MEDIA_FIELD_MAP` (`:274-288`) maps `album_id` to
  `mb_albumid`, `artist` to `albumartist`, `artists` to `albumartists`,
  `artist_id` to `mb_albumartistid`, `artists_ids` to
  `mb_albumartistids`, `artist_credit` to `albumartist_credit`,
  `artists_credit` to `albumartists_credit`, `artist_sort` to
  `albumartist_sort`, `artists_sort` to `albumartists_sort`, `mediums`
  to `disctotal`, `releasegroup_id` to `mb_releasegroupid`, and `va` to
  `comp`. `IGNORED_FIELDS` is `{"data_url", "tracks"}` (`:272`).
- `TrackInfo.MEDIA_FIELD_MAP` (`:401-409`) maps `artist_id` to
  `mb_artistid`, `artists_ids` to `mb_artistids`, `medium` to `disc`,
  `release_track_id` to `mb_releasetrackid`, `track_id` to
  `mb_trackid`, and `medium_index` to `track`. `IGNORED_FIELDS` is
  `{"data_url", "index", "medium_total"}` (`:396-400`).
- `AlbumInfo.item_data` (`:309-320`) applies the `original_date`
  override: when the config flag is set and `original_year` is present,
  `year` becomes `original_year` and `month` and `day` follow the
  original values or `0`.

Both of the config keys that move applied fields without moving a
verdict, `original_date` and `per_disc_numbering`, are kept verbatim in
`corpus/profiles/estate.yaml` for exactly this reason
(`corpus/README.md`'s profile table).

### 1.3 `Item.write` and the two field sets

`Item.write` (`beets/library/models.py:934-989`) does five things in
order: it resolves `id3v23` from `beets.config["id3v23"]` when the
argument is `None` (`:961-962`), it filters `dict(self)` down to
`self._media_fields` (`:966-968`), it sends the `write` plugin event,
it opens `MediaFile(syspath(path), id3v23=id3v23)`, and it calls
`mediafile.update(item_tags)` then `mediafile.save()` (`:975-985`).
`Item.write` returns `None`, so a caller learns nothing about what
landed. That is why the corpus's ledger reads the file back
independently.

Two class-level sets bound the write:

- `Item._media_fields = set(MediaFile.readable_fields()) & _field_names`
  (`models.py:731`). Measured on the pinned venv: 78 names.
- `Item._media_tag_fields = set(MediaFile.fields()) & _field_names`
  (`:737`). This is the set `Item.clear()` nulls.

`MediaFile.update` (`mediafile/__init__.py:365-378`) iterates
`sorted_fields()` and, per field present in the dict, either
`delattr` for `None` or `setattr` for anything else. So the writable
set is the intersection of `Item._media_fields` with
`MediaFile.sorted_fields()`, and the remainder never reaches a file.

Computed against the pinned venv, the same way
`derive_file_tags.py:56-89` computes it:

```
$ BEETSDIR=$(mktemp -d) tools/venv/.venv/bin/python -c "..."
69 9
```

69 writable, 9 read-only, matching `corpus/README.md`'s handoff list
key for key. `MediaFile.fields()` yields 82, `readable_fields()` yields
91 (`fields()` plus the nine audio properties, `__init__.py:333-348`),
and `Item._media_fields` is 78. The nine read-only names are
`bitdepth`, `bitrate`, `bitrate_mode`, `channels`, `encoder_info`,
`encoder_settings`, `format`, `length`, and `samplerate`.

### 1.4 Which of the 69 come from `apply_metadata`, and which ride along

Every one of the 69 is written by `Item.write`, but only some of them
carry a value `apply_metadata` supplied. The rest are the item's own
pre-existing values, written back unchanged.

Measured over all 215 committed cases, the union of
`Expected.AppliedFields` keys is 99 names. All 69 writable keys appear
in that union, so the corpus exercises the full writable surface at the
in-memory altitude. The 30 keys outside the writable set split into the
nine read-only audio properties, bookkeeping (`id`, `path`, `mtime`,
`added`), provider ids beets keeps as flexible attributes
(`data_source`, `discogs_albumid`, `discogs_artistid`,
`discogs_labelid`), the `*_ids` credit lists (`artists_ids`,
`arrangers_ids`, `composers_ids`, `lyricists_ids`, `remixers_ids`),
and info-level names that never became item fields (`album_id`,
`release_group_title`, `releasegroupdisambig`, `style`, `track_alt`,
`trackdisambig`, `work`, `work_disambig`).

At the file altitude the picture is narrower, and this is the finding
that shapes the rung's test plan. Across the twelve subjects in
`corpus/analysis/file_tags.json`, 52 of the 69 keys carry a non-empty
value on at least one row. The other 17 are empty or absent on all
twelve:

`acoustid_fingerprint`, `acoustid_id`, `bpm`, `comments`, `encoder`,
`genres`, `grouping`, `initial_key`, `lyrics`, `r128_album_gain`,
`r128_track_gain`, `remixers`, `rg_album_gain`, `rg_album_peak`,
`rg_track_gain`, `rg_track_peak`, `subtitle`.

Those 17 are exactly the keys a MusicBrainz match does not supply:
fingerprints, gain analysis, and user-authored text. They are also the
keys with the most interesting serialization rules, since they include
all four ReplayGain floats, both R128 Q-numbers, and two of the
multi-valued list fields. A conformance runner that grades only against
the ledger proves nothing about them.

---

## 2. The mediafile 0.17.0 FLAC layer, key by key

This is the spec's core. Every row is read from the pinned
`mediafile/__init__.py`; the mechanism rows come from
`mediafile/fields.py` and `mediafile/storage/`.

### 2.1 The mechanism, before the table

A `MediaField` holds an ordered tuple of `StorageStyle` instances and
filters them per file by class name: a style applies when
`mutagen_file.__class__.__name__` appears in the style's `formats`
list (`fields.py:40-46`). The base `StorageStyle.formats` list
(`storage/base.py:29-40`) contains `FLAC` plus the Ogg family and the
APE family, so a bare `StorageStyle` is the FLAC style. Every
`MP3StorageStyle`, `MP4StorageStyle`, `ASFStorageStyle`, and
`APEv2*StorageStyle` in a declaration is invisible to a FLAC file, and
the table below lists only the styles that apply.

Reading takes the first style whose value is truthy
(`fields.py:48-54`). Writing is different and is the rule a writer has
to get right: `MediaField.__set__` (`:56-61`) writes **every**
applicable style that is not `read_only`. So `albumartist` writes three
Vorbis keys, `disctotal` writes three, `tracktotal` writes three,
`track` writes two, `disc` writes two, `comments` writes two, `label`
writes two, `albumstatus` writes two, and `encoder` writes two.
`read_only=True` blocks `set` but not `delete` (`base.py:60-65`,
`132-135`), so deleting `barcode` removes `UPC`, `EAN/UPN`, `EAN`, and
`UPN` as well as `BARCODE`.

Serialization is `StorageStyle.serialize` (`base.py:109-130`):

- A `float` with `as_type is str` becomes `"{:.<float_places>f}"`.
  `float_places` defaults to 2 and is 6 on both ReplayGain peak fields.
- A `bool` becomes `"1"` or `"0"`, never `True`/`False`.
- `bytes` decode as UTF-8 with errors ignored; anything else goes
  through `str()`.
- A non-empty `suffix` is appended after formatting. The only FLAC
  suffix in the whole table is `" dB"` on the two ReplayGain gain
  fields.

Reading back runs `deserialize` (`:86-97`), which strips the suffix
when present, then `safe_cast` to the field's `out_type`
(`utils/type_conversion.py:4-60`). `safe_cast` to `int` takes the
leading `[+-]?[0-9]+` of the trimmed string and yields 0 when there is
none; to `float` it takes the leading numeric prefix and yields 0.0;
to `bool` it is `bool(int(val))` with a `ValueError` falling to
`False`. `safe_cast(t, None)` returns `None` for every type,
so an absent key reads back as `None` rather than as a typed zero.

Multi-valued fields go through `ListStorageStyle`
(`base.py:138-200`). `fetch` returns the whole mutagen value list for
the key, `get_list` deserializes each element, and `store` assigns the
serialized list to the key. In a Vorbis comment block that means the
same field name repeated once per value, which is exactly how mutagen
serializes a list-valued key. `ListMediaField.__get__`
(`fields.py:89-94`) returns the first style's list that is non-empty,
and `__set__` (`:96-99`) writes every non-read-only style.

`single_field()` (`fields.py:101-106`) builds a plain `MediaField` over
the same styles, which is how `albumtype`, `catalognum`, `language`,
`mb_artistid`, and `mb_albumartistid` reach list-backed keys. Setting
one of them goes through `ListStorageStyle.set` (`base.py:180-187`),
which stores a one-element list.

Dates are `DateField` (`fields.py:109-210`). `_set_date_tuple`
(`:183-201`) formats `YYYY`, appends `-MM` when the month is truthy,
appends `-DD` only when both the month and the day are truthy, joins
with `-`, writes that through the main styles, and then also writes the
bare year through the year-fallback style when one exists. A `None`
year deletes the field entirely, including the year fallback
(`:146-149`, `:188-190`). `_get_date_tuple` (`:151-181`) strips
everything from a `T` or a space onward, splits on `-` or `/`,
truncates or pads to three components, reaches for the year style only
when the first component is falsy, and converts each component with a
`try: int(...)` that yields `None` on failure. The consequence for a
writer: `year=2001, month=0, day=0` produces `DATE=2001` and
`YEAR=2001`, and reads back as year 2001 with month and day `None`,
not 0.

`QNumberField` (`fields.py:272-293`) is the R128 pair. The setter
writes `round(value * 2**fraction_bits)` as an integer string, and both
R128 fields use `fraction_bits=8` (`__init__.py:870-884`). Measured:
`r128_track_gain = -1.5` writes `R128_TRACK_GAIN=-384`.

`MediaFile.update` (`__init__.py:365-378`) iterates
`sorted_fields()` (`:322-330`), which is lexicographic on the field
name except that `DateItemField` names are sorted as `date0`, `date1`,
`date2` through `_field_sort_name` (`:306-319`). Three pairs of
writable fields share a Vorbis key, and the write order decides the
outcome. In each pair the plural sorts last and therefore wins:
`albumtype` then `albumtypes` over `RELEASETYPE` and
`MUSICBRAINZ_ALBUMTYPE`; `mb_artistid` then `mb_artistids` over
`MUSICBRAINZ_ARTISTID`; `mb_albumartistid` then `mb_albumartistids`
over `MUSICBRAINZ_ALBUMARTISTID`. Confirmed by writing
`albumtype="album"` and `albumtypes=["album","live"]` in one `update`
call and reading two `RELEASETYPE` comments back.

### 2.2 The 69-key table

Keys marked read-only are read by mediafile and never written, but they
are deleted when the field is deleted.

| beets field | Vorbis key(s), in mediafile declaration order | Write shape | mediafile line(s) |
| --- | --- | --- | --- |
| `acoustid_fingerprint` | `ACOUSTID_FINGERPRINT` | one comment | `:817` |
| `acoustid_id` | `ACOUSTID_ID` | one comment | `:823` |
| `album` | `ALBUM` | one comment | `:410` |
| `albumartist` | `ALBUM ARTIST`, then `ALBUM_ARTIST`, then `ALBUMARTIST` | one comment under each of the 3 writable keys | `:542`, `:543`, `:544` |
| `albumartist_credit` | `ALBUMARTIST_CREDIT` | one comment | `:732` |
| `albumartist_sort` | `ALBUMARTISTSORT` | one comment | `:588` |
| `albumartists` | `ALBUMARTISTS`, then `ALBUM_ARTISTS`, then read-only `ALBUM ARTISTS` | one comment per value, repeated under each of the 2 writable keys | `:554`, `:555`, `:556` |
| `albumartists_credit` | `ALBUMARTISTS_CREDIT` | one comment per value, repeated under the writable key | `:738` |
| `albumartists_sort` | `ALBUMARTISTS_SORT` | one comment per value, repeated under the writable key | `:744` |
| `albumdisambig` | `MUSICBRAINZ_ALBUMCOMMENT` | one comment | `:680` |
| `albumstatus` | `RELEASESTATUS`, then `MUSICBRAINZ_ALBUMSTATUS` | one comment under each of the 2 writable keys | `:665`, `:666` |
| `albumtype` | `RELEASETYPE`, then `MUSICBRAINZ_ALBUMTYPE` | one-element list into each of the 2 writable keys, via `albumtypes.single_field()` | `:562`, `:563` |
| `albumtypes` | `RELEASETYPE`, then `MUSICBRAINZ_ALBUMTYPE` | one comment per value, repeated under each of the 2 writable keys | `:562`, `:563` |
| `arrangers` | `ARRANGER` | one comment per value, repeated under the writable key | `:444` |
| `artist` | `ARTIST` | one comment | `:398` |
| `artist_credit` | `ARTIST_CREDIT` | one comment | `:714` |
| `artist_sort` | `ARTISTSORT` | one comment | `:581` |
| `artists` | `ARTISTS` | one comment per value, repeated under the writable key | `:404` |
| `artists_credit` | `ARTISTS_CREDIT` | one comment per value, repeated under the writable key | `:720` |
| `artists_sort` | `ARTISTS_SORT` | one comment per value, repeated under the writable key | `:726` |
| `asin` | `ASIN` | one comment | `:594` |
| `barcode` | `BARCODE`, then read-only `UPC`, then read-only `EAN/UPN`, then read-only `EAN`, then read-only `UPN` | one comment | `:616`, `:617`, `:618`, `:619`, `:620` |
| `bpm` | `BPM` | one comment | `:528` |
| `catalognum` | `CATALOGNUMBER`, then read-only `CATALOGID`, then read-only `DISCOGS_CATALOG` | one-element list into the writable key, via `catalognums.single_field()` | `:604`, `:605`, `:606` |
| `comments` | `DESCRIPTION`, then `COMMENT` | one comment under each of the 2 writable keys | `:514`, `:515` |
| `comp` | `COMPILATION` | one comment | `:535` |
| `composer_sort` | `COMPOSERSORT` | one comment | `:438` |
| `composers` | `COMPOSER` | one comment per value, repeated under the writable key | `:431` |
| `country` | `RELEASECOUNTRY` | one comment | `:659` |
| `day` | `DATE`, with `YEAR` as the year fallback | component 2 of `YYYY[-MM[-DD]]`; setting any component rewrites `DATE`, and `year` also writes `YEAR` | `:695`; `DateField` `:685-692`, `DATE` `:688`, `YEAR` `:690` |
| `disc` | `DISC`, then `DISCNUMBER` | one comment under each of the 2 writable keys | `:480`, `:481` |
| `disctitle` | `DISCSUBTITLE` | one comment | `:632` |
| `disctotal` | `DISCTOTAL`, then `DISCC`, then `TOTALDISCS` | one comment under each of the 3 writable keys | `:488`, `:489`, `:490` |
| `encoder` | `ENCODEDBY`, then `ENCODER` | one comment under each of the 2 writable keys | `:638`, `:639` |
| `genres` | `GENRE` | one comment per value, repeated under the writable key | `:416` |
| `grouping` | `GROUPING` | one comment | `:452` |
| `initial_key` | `INITIALKEY` | one comment | `:888` |
| `isrc` | `ISRC` | one comment | `:626` |
| `label` | `LABEL`, then `PUBLISHER` | one comment under each of the 2 writable keys | `:574`, `:575` |
| `language` | `LANGUAGE` | one-element list into the writable key, via `languages.single_field()` | `:651` |
| `lyricists` | `LYRICIST` | one comment per value, repeated under the writable key | `:424` |
| `lyrics` | `LYRICS` | one comment | `:504` |
| `mb_albumartistid` | `MUSICBRAINZ_ALBUMARTISTID` | one-element list into the writable key, via `mb_albumartistids.single_field()` | `:801` |
| `mb_albumartistids` | `MUSICBRAINZ_ALBUMARTISTID` | one comment per value, repeated under the writable key | `:801` |
| `mb_albumid` | `MUSICBRAINZ_ALBUMID` | one comment | `:782` |
| `mb_artistid` | `MUSICBRAINZ_ARTISTID` | one-element list into the writable key, via `mb_artistids.single_field()` | `:788` |
| `mb_artistids` | `MUSICBRAINZ_ARTISTID` | one comment per value, repeated under the writable key | `:788` |
| `mb_releasegroupid` | `MUSICBRAINZ_RELEASEGROUPID` | one comment | `:809` |
| `mb_releasetrackid` | `MUSICBRAINZ_RELEASETRACKID` | one comment | `:770` |
| `mb_trackid` | `MUSICBRAINZ_TRACKID` | one comment | `:764` |
| `mb_workid` | `MUSICBRAINZ_WORKID` | one comment | `:776` |
| `media` | `MEDIA` | one comment | `:673` |
| `month` | `DATE`, with `YEAR` as the year fallback | component 1 of `YYYY[-MM[-DD]]`; setting any component rewrites `DATE`, and `year` also writes `YEAR` | `:694`; `DateField` `:685-692`, `DATE` `:688`, `YEAR` `:690` |
| `original_day` | `ORIGINALDATE`, no year fallback | component 2 of `YYYY[-MM[-DD]]` | `:708`; `DateField` `:698-705`, `ORIGINALDATE` `:702` |
| `original_month` | `ORIGINALDATE`, no year fallback | component 1 of `YYYY[-MM[-DD]]` | `:707`; `DateField` `:698-705`, `ORIGINALDATE` `:702` |
| `original_year` | `ORIGINALDATE`, no year fallback | component 0 of `YYYY[-MM[-DD]]` | `:706`; `DateField` `:698-705`, `ORIGINALDATE` `:702` |
| `r128_album_gain` | `R128_ALBUM_GAIN` | one comment | `:881` |
| `r128_track_gain` | `R128_TRACK_GAIN` | one comment | `:874` |
| `remixers` | `REMIXER` | one comment per value, repeated under the writable key | `:750` |
| `rg_album_gain` | `REPLAYGAIN_ALBUM_GAIN` | one comment; suffix `" dB"`; `%.2f` | `:846` |
| `rg_album_peak` | `REPLAYGAIN_ALBUM_PEAK` | one comment; `%.6f` | `:864` |
| `rg_track_gain` | `REPLAYGAIN_TRACK_GAIN` | one comment; suffix `" dB"`; `%.2f` | `:836` |
| `rg_track_peak` | `REPLAYGAIN_TRACK_PEAK` | one comment; `%.6f` | `:856` |
| `script` | `SCRIPT` | one comment | `:645` |
| `subtitle` | `SUBTITLE` | one comment | `:457` |
| `title` | `TITLE` | one comment | `:392` |
| `track` | `TRACK`, then `TRACKNUMBER` | one comment under each of the 2 writable keys | `:463`, `:464` |
| `tracktotal` | `TRACKTOTAL`, then `TRACKC`, then `TOTALTRACKS` | one comment under each of the 3 writable keys | `:471`, `:472`, `:473` |
| `year` | `DATE`, with `YEAR` as the year fallback | component 0 of `YYYY[-MM[-DD]]`; setting any component rewrites `DATE`, and `year` also writes `YEAR` | `:693`; `DateField` `:685-692`, `DATE` `:688`, `YEAR` `:690` |

### 2.3 Empty is not absent, and this is graded

`MediaFile.update` deletes on `None` and sets on everything else, so an
empty string writes a comment with an empty value and reads back as
`""`, while a `None` removes the key and reads back as `None`. The
ledger records both states. On row 0 of `corpus/analysis/file_tags.json`
ten keys are `null` (`arrangers`, `genres`, `initial_key`, both R128
fields, `remixers`, and all four ReplayGain fields) and ten are the
empty string (`acoustid_fingerprint`, `acoustid_id`, `albumdisambig`,
`asin`, `comments`, `disctitle`, `encoder`, `grouping`, `lyrics`,
`subtitle`). Across the twelve rows the value types are 421 strings,
152 nulls, 123 lists, 120 ints, and 12 booleans.

A Go writer therefore has three distinct behaviors to implement, not
two: write a value, write an empty comment, and remove the key. Getting
the last two backwards produces a ledger diff on twenty keys per row
while every non-empty value matches.

The measured round trip, run against a scratch copy of the committed
template rather than anything in the repo, produced this comment set
from a fifteen-field `MediaFile.update`:

```
ACOUSTID_FINGERPRINT=      DISCTOTAL=2
ALBUM=Al                   MUSICBRAINZ_ALBUMTYPE=album
ALBUMARTISTS=X             MUSICBRAINZ_ALBUMTYPE=live
ALBUMARTISTS=Y             ORIGINALDATE=1999-03-04
ALBUM_ARTISTS=X            R128_TRACK_GAIN=-384
ALBUM_ARTISTS=Y            RELEASETYPE=album
ARTIST=A                   RELEASETYPE=live
BPM=120                    REPLAYGAIN_TRACK_GAIN=-7.50 dB
COMPILATION=1              REPLAYGAIN_TRACK_PEAK=0.987654
DATE=2001                  TITLE=T
DISC=1                     TOTALDISCS=2
DISCC=2                    TOTALTRACKS=12
DISCNUMBER=1               TRACK=3
                           TRACKC=12
                           TRACKNUMBER=3
                           TRACKTOTAL=12
                           YEAR=2001
```

`barcode=None` in the same call left no `BARCODE` key at all. The
`ALBUM ARTISTS` key never appeared, because its style is read-only.

### 2.4 The PICTURE block

`images` is an `ImageListField` (`__init__.py:758`,
`fields.py:296-316`) and `art` is a `CoverArtField` (`__init__.py:755`,
`fields.py:234-269`). Neither is in the writable 69, because neither is
a beets `Item` field: measured on the venv, `images` and `art` are both
in `MediaFile.fields()` and neither is in `Item._media_fields`. So the
corpus grades no picture behavior at all, and the launch ruling that
the FLAC writer writes and strips the front cover needs its own test
evidence.

For FLAC the applicable style is `FlacImageStorageStyle`
(`storage/flac.py:8-41`), which is a `ListStorageStyle` over
`mutagen_file.pictures`. `store` calls `clear_pictures()` and then
`add_picture()` per image (`:24-28`), `delete` is `clear_pictures()`
alone (`:39-41`), and `serialize` builds a `mutagen.flac.Picture` with
`data`, `type` from the image's type index, `mime` from
`image.mime_type`, and `desc` or the empty string (`:30-37`).
`VorbisImageStorageStyle` (`storage/vorbis.py:11-58`) is the
base64-in-a-comment variant for the Ogg family and does not apply to
FLAC, though its `store` is where the legacy `COVERART` and
`COVERARTMIME` keys get stripped.

Three details a Go implementation has to reproduce:

- The MIME type is sniffed from the bytes, not declared by the caller.
  `Image.mime_type` (`utils/image.py:49-52`) calls
  `image_mime_type`, which is `filetype.guess_mime(data)` (`:12-14`).
  The `filetype` package is a pinned transitive dependency.
- The picture type is the ID3v2 APIC enumeration, `ImageType` in
  `constants.py:21-44`: `other` 0 through `publisher_logo` 20, with
  `front` at 3. `Image.type_index` (`utils/image.py:54-60`) returns 0
  when the type is unset, and an out-of-range integer index degrades to
  `other` with a debug log (`:41-47`).
- Front-cover selection when several pictures exist is
  `CoverArtField.guess_cover_image` (`fields.py:253-260`): the single
  candidate when there is one, otherwise the first with type `front`,
  otherwise the first.

The width, height, colour depth, and indexed-colour fields of the
PICTURE block are whatever `mutagen.flac.Picture` defaults them to,
since mediafile sets none of them. A Go writer that fills them from a
decoded image is writing more than the reference did, which is a
deviation to state rather than to smuggle in.

---

## 3. `derive_file_tags.py`, the fixture, and what a Go test must reproduce

### 3.1 The derivation, step by step

`internal/harness/oracle/derive_file_tags.py` is 205 lines and is the
corpus's second, independent beets entry point. It starts no stub,
issues no request, renders no profile, and loads no plugin
(`:1-34`). Its Go driver is `internal/harness/oracle/file_tags.go`,
which carries the `oracle` build tag and is never called from
`oracle.Derive` (`file_tags.go:1-8`).

`main` (`:183-201`) creates a `TemporaryDirectory` and sets `BEETSDIR`
to it before anything imports beets, which is what puts the run on
beets' shipped defaults rather than the workstation's
`~/.config/beets`. The reason is named in the code: `Item.write()`
resolves `beets.config["id3v23"]` before it opens the file, so a
stray local setting could change or break a write (`:190-195`). It then
redirects `sys.stdout` to stderr so that beets' own chatter cannot
corrupt the JSON document, which goes to the real stdout (`:198-201`).

`field_lists` (`:56-89`) computes the writable and read-only lists off
the venv rather than hardcoding them, and asserts the `BEETSDIR`
isolation held at the moment `beets.library.Item` was imported
(`:76-81`). `assert_json_primitive` (`:92-109`) refuses any
`AppliedFields` value that is not a JSON primitive or a list of them,
which is the guard against `derive_oracle.py`'s `str(value)` fallback
silently stringifying a future field type.

`derive_row` (`:128-162`) is the write path: `build_items([fields])[0]`
reconstructs a beets `Item` from the frozen record, the committed
template is copied to a per-subject scratch path named
`<case_id>-<item_index>.flac`, `item.write(path)` writes it, and then a
**fresh** `MediaFile(path)` parse reads back
`{field: getattr(mf, field) for field in writable}`. The comment at
`:151-155` states why the read-back is a new parse rather than the
object `write()` already held: it mirrors `Item.read()`'s own
independent parse (`models.py:917`).

`template_block` (`:112-125`) reads the nine read-only properties once,
from an unwritten copy of the template, because they are a property of
the container and identical for every subject.

`derive` (`:165-180`) emits the document: `shim_version`,
`beets_version`, `writable_fields`, `readonly_fields`, `template`, and
`rows`.

### 3.2 The twelve subjects and the ledger as committed

The subject list is fixed in Go, not in the shim:
`analysis.FileTagSubjects` (`internal/harness/analysis/file_tags.go:39-52`)
names five smoke or structural cases and seven real-catalog cases, all
at item index 0. The twelfth, `real-decoy-mozart-requiem`, replaced a
decoy case that never resolved to a committed case; the substitution is
a recorded conductor ruling, not an implementer's choice (`:28-38`).

`corpus/analysis/file_tags.json` is 37,782 bytes,
`shim_version` `"1"`, `beets_version` `"2.13.1"`, twelve rows of 69
values each. Its `template` block is `bitdepth` 16, `bitrate` 0,
`bitrate_mode` `""`, `channels` 1, `encoder_info` `""`,
`encoder_settings` `""`, `format` `"FLAC"`, `length` 0, `samplerate`
44100. Rung 2's own checks over it are schema, coverage, and
key-membership only, because nothing at rung 2 could compare against a
Go writer (`analysis/file_tags.go:62-63`, `:90-95`).

### 3.3 The template fixture

`internal/harness/oracle/testdata/template.flac` is 42 bytes. Read with
`xxd`, it is exactly the `fLaC` signature, one metadata block header of
`80 00 00 22` (last-block flag set, type 0, 34 bytes), and a 34-byte
STREAMINFO payload whose only non-zero content is the eight-byte pack
at offset 10, `0a c4 40 f0 00 00 00 00`: sample rate 44100, one
channel, 16 bits per sample, zero total samples. Minimum and maximum
block size, minimum and maximum frame size, and the sixteen-byte MD5
are all zero. There are no audio frames, no PADDING block, and no
VORBIS_COMMENT block.

Its provenance is `internal/harness/oracle/testdata/gen_template_flac.py`,
which is run by hand and by nothing else, and which builds the
STREAMINFO through mutagen's own `StreamInfo.write`. Its docstring
records why 42 bytes suffices: mediafile dispatches a file as FLAC by
checking `type(mgfile).__name__ == "FLAC"`
(`mediafile/__init__.py:175`), which needs only the signature and a
well-formed STREAMINFO.

### 3.4 What a Go conformance test has to reproduce

To grade against `corpus/analysis/file_tags.json` the runner has to do
what the shim did, in the same order:

1. Copy the committed 42-byte template to a scratch path per subject.
   Never write into `corpus/` or `internal/harness/`; both are
   read-only to this rung.
2. Take the subject's frozen `Expected.AppliedFields` from the
   committed case, filtered to the ledger's own `writable_fields`
   header. Reading that header rather than a second hardcoded list is
   the same rule the shim follows.
3. Write those values with the three-state rule: a value writes, an
   empty string writes an empty comment, a null deletes the key.
4. Read the container back through an independent parse and project
   each of the 69 keys back to the ledger's JSON shape, including the
   `None`-versus-`""` distinction, the int and float casts, the list
   shape, and the date-component decomposition.
5. Compare against the row. The nine read-only keys are never compared
   per subject; the `template` block is the only place they appear, and
   it describes the fixture rather than any case.

The read-back projection is the part that is easy to underestimate. It
is a full re-implementation of `MediaField.__get__` for the 69 fields,
including first-truthy-style precedence, `safe_cast`, `_get_date_tuple`,
and the `QNumberField` division. The shadow's provisional reader
(section 4) already implements that projection for nineteen fields, and
it is the natural nucleus.

---

## 4. The shadow's provisional FLAC reader, as landed

`internal/engine/shadow/flac.go` is 313 lines, unexported, and named
provisional in its own doc comment (`:62-70`). The rung 4 spec records
the decision that put it there and its cost
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1002-1033`): option
(a), a minimal read-only reader inside `shadow`, chosen over depending
on rung 8 because waiting on lane T spends calendar on the one rung
whose calendar is the launch date.

### 4.1 Function by function

- `readFLAC(path) (matcher.ItemFields, error)` (`:77-87`) reads the
  whole file with `os.ReadFile`, parses, and projects. A malformed
  container is an error and never a partly filled bag.
- `parseFLAC(data) (flacFile, error)` (`:90-126`) checks the `fLaC`
  signature, then walks metadata block headers: the last-block flag is
  `data[off]&0x80`, the type is `data[off]&0x7f`, and the length is
  three big-endian bytes. It reads STREAMINFO (type 0) and
  VORBIS_COMMENT (type 4) and skips every other type by its declared
  length. It stops at the last-block flag, so it never touches audio
  frames.
- `parseStreamInfo(payload) (streamInfo, error)` (`:135-146`) requires
  exactly 34 bytes and decodes the eight-byte pack at `[10:18]`: sample
  rate from bits 63 to 44, channels less one from 43 to 41, bits per
  sample less one from 40 to 36, and the 36-bit total sample count. The
  doc comment cites mutagen's own write order
  (`mutagen/flac.py:233-257`) and names `[18:34]` as the audio MD5.
- `(*flacFile).readVorbisComment(payload) error` (`:153-180`) decodes
  the little-endian layout: vendor length and string, comment count,
  then each comment's length and its `KEY=value` bytes. Keys fold to
  upper case and the **first** occurrence of a key wins; later ones are
  dropped, justified by mediafile reading only a comment's first value
  (`storage/base.py:77-84`).
- `vorbisField(b)` (`:184-193`) splits one length-prefixed field.
- `(flacFile).fields() matcher.ItemFields` (`:208-229`) is the
  projection, one row per mediafile `MediaField`, in mediafile's own
  declaration order, with a trailing `// mediafile:NNN` citation per
  row.
- `tag(keys ...string) string` (`:235-242`) implements
  first-truthy-style precedence, and its comment records the sharp
  edge: `"0"` is a truthy Python string, so `TRACK=0` beside
  `TRACKNUMBER=5` yields 0.
- `year()` (`:250-260`) and `dateYear()` (`:263-271`) implement
  `_get_date_tuple`'s year component with the `YEAR` fallback.
- `castInt` (`:277-298`) and `castBool` (`:304-307`) reproduce
  `safe_cast` for int and bool.
- `firstBytes` (`:311-313`) bounds an error message about a file that
  may hold anything.

`streamInfo.length()` (`:41-46`) is total samples over sample rate,
returning zero rather than dividing by zero.

### 4.2 What it deliberately lacks

No audio frame decoding, no writing, no format but FLAC, and only two
of the seven block types. It carries no PICTURE handling, no
multi-valued read (the map is `map[string]string` and drops repeats),
and no `float`, `list`, or Q-number casts, because none of the nineteen
fields needs them.

### 4.3 The nineteen-field mapping

The table with a citation per row lives in the pass 4a plan
(`docs/superpowers/plans/2026-09-05-rung-4a-shadow-runner.md:2453-2478`),
and the same citations are inline in `flac.go:210-228`. Eighteen rows
map to Vorbis keys; `Length` comes from STREAMINFO; `DataSource` has no
row at all, because a grep of pinned mediafile for `data_source`
returns nothing and beets keeps it as a database flexible attribute
(`plan:2485-2492`, `flac.go:203-207`). The nineteen fields are the
whole of `matcher.ItemFields`
(`internal/engine/matcher/info.go:124-152`).

Every one of the eighteen mapped fields is inside the writable 69, so
rung 8's surface is a strict superset of what the shadow already reads.

### 4.4 What absorbing it requires

Three call sites reference the reader:

- `internal/engine/shadow/run.go:209`, `LibraryReadFn: readFLAC`.
- `internal/engine/shadow/capture.go:158`, a direct
  `readFLAC(filepath.Join(dir, entry.Name()))`.
- `internal/engine/shadow/reconcile.go:35`, the `LibraryReadFn` field's
  own doc comment, which names `readFLAC`.

Two more mention it only in prose (`capture_test.go:47`, `:203`).

The test move is the harder half. `flac_test.go` is 361 lines with
seven tests, and its fixture builder `buildFLAC` is used by three
files, not one: `capture_test.go:34` and `:173`, and `run_test.go:107`.
`buildFLAC` builds a container in memory rather than vendoring a
fixture, deliberately, so a mutation is a byte edit
(`flac_test.go:12-21`, spec `:1036-1043`). Two tests pin the format to
something outside the package: `TestReadFLACParsesTheCommittedTemplateBytes`
(`:259-271`) parses a hand-transcribed copy of the committed 42 bytes
(`:244-253`), and `TestBuildFLACEmitsTheStreaminfoLayoutByteForByte`
(`:273`) pins the builder against a hand-written expectation.

So absorbing the reader means deciding where `buildFLAC` lives. Moving
it to `tagio`'s test files leaves three shadow tests without a fixture
builder; leaving it in `shadow` leaves `tagio`'s own tests to grow a
second one. Neither is obviously right, and a duplicated container
builder is the kind of drift the two pinning tests exist to catch.

One architectural point the absorption forces. `flac.go` imports
`internal/engine/matcher` for `ItemFields`, but `internal/engine/tag`'s
doc comment says `tagio` "reads and writes FLAC tags and knows nothing
about matching" (`internal/engine/tag/doc.go:8-11`). A `tagio` that
returns `matcher.ItemFields` contradicts that. The clean shape is a
`tagio` that reads and writes a tag set in its own vocabulary, with the
projection to `matcher.ItemFields` staying on the shadow side.

Rung 4's spec also states the one inbound edge in advance: the deployed
binary carries landed-but-unwired packages from other lanes, none
reachable from the `shadow` subcommand's call graph, and repointing is
rung 8's task and never a silent adoption (`spec:1591-1597`,
`docs/HISTORY.md:442-447`).

---

## 5. The FLAC format facts a writer needs

Citations are to RFC 9639, the FLAC specification. Section numbers were
read from the published RFC rather than from memory.

**The stream layout.** Section 6 gives the order: the `fLaC` marker
(0x664C6143), then the metadata blocks, then the audio frames.

**The metadata block header.** Section 8.1: a four-byte header of one
last-metadata-block flag bit (0 means more blocks follow, 1 means this
is the final one), a seven-bit block type, and a three-byte big-endian
block size in bytes, excluding the header. Note the asymmetry that
catches implementers: the header length is big-endian, and the
VORBIS_COMMENT body's lengths are little-endian.

**Block types.** Section 8.1, Table 2: 0 STREAMINFO, 1 PADDING, 2
APPLICATION, 3 SEEKTABLE, 4 VORBIS_COMMENT, 5 CUESHEET, 6 PICTURE, 7
through 126 reserved, and 127 forbidden. A writer preserves every block
it does not understand, by its declared length, which is what the
shadow's reader already does.

**STREAMINFO.** Section 8.2: "It MUST be present as the first metadata
block in the stream", and "There MUST be no more than one streaminfo
metadata block per FLAC stream." The block carries the block and frame
size bounds, sample rate, channel count, bit depth, total sample count,
and an MD5 checksum of the unencoded audio data, whose purpose the spec
gives as letting a decoder detect an error in the audio data even when
the bitstream is valid. "A value of `0` signifies that the value is not
known", which is exactly what the corpus's 42-byte template carries.

Nothing a tag writer does may change any STREAMINFO field. The MD5 is
over unencoded audio, and a tag rewrite moves no audio sample, so the
correct behavior is to copy the STREAMINFO payload through byte for
byte and to copy the frame region byte for byte. Recomputing the MD5,
or re-encoding frames to produce one, is the failure mode that makes a
tag write lossy.

**VORBIS_COMMENT.** Section 8.6: the header is directly followed by
four bytes holding the vendor string length, little-endian; the UTF-8
vendor string, "not terminated in any way"; then a four-byte
little-endian field count; then, per field, a four-byte little-endian
length and that many bytes of `NAME=value` in UTF-8. FLAC's version
carries no framing bit, unlike the original Vorbis specification. "A
FLAC file MUST NOT contain more than one Vorbis comment metadata
block", which means an in-place rewrite replaces the one block rather
than appending a second. The vendor string is the writer's own
identity; pinned mutagen writes `Mutagen 1.48.1`, measured on a scratch
copy of the template.

Section 8.6.4 adds two rules a writer depends on. A field name "MUST
only consist of UTF-8 code points U+0020 through U+007E, excluding
U+003D", and field name evaluation "MUST be case insensitive". Nothing
forbids the same field name appearing more than once, which is what
makes the repeated-key serialization of a multi-valued field legal.

**PICTURE.** Section 8.8, Table 12, in order: picture type `u(32)`,
media type length `u(32)`, media type string, description length
`u(32)`, description, width `u(32)`, height `u(32)`, colour depth
`u(32)`, number of colours `u(32)`, picture data length `u(32)`, and
the picture data. All big-endian, unlike VORBIS_COMMENT. More than one
PICTURE block is allowed.

**PADDING.** Section 8.3: the contents are "n '0' bits (n MUST be a
multiple of 8)", and "There MAY be one or more padding metadata blocks
per FLAC stream." SEEKTABLE is capped at one per stream (section 8.5).

**Padding strategy for a rewrite.** Two shapes are available, and the
difference is what a crash costs.

- *In-place*, when the new metadata block region fits inside the old
  one plus the trailing PADDING: rewrite the header region and shrink
  the PADDING block to absorb the difference. No audio byte moves, and
  the write is a few kilobytes regardless of file size. It is not
  atomic: a torn write leaves a file whose metadata region is half old
  and half new, though the audio frames are untouched.
- *Full rewrite*, when the new region does not fit: write signature,
  new metadata blocks, and a copy of the whole frame region to a new
  file, then rename. Atomic through the usual temp-file discipline, and
  it costs one full copy of the file.

Reserving padding on the write is what makes the cheap path available
later. Pinned mutagen reserves 1024 bytes by default, measured: writing
fifteen fields into the 42-byte template produced a 1,649-byte file
whose blocks were STREAMINFO 34 bytes, VORBIS_COMMENT 575 bytes, and a
last-block PADDING of 1024 bytes.

A third fact matters for rung 9's retag verb. Because the audio region
is byte-identical across a tag rewrite, a content hash taken over the
frame region alone is stable under retagging, while a hash over the
whole file is not. Forward hook 1 puts a content hash on the catalog
row computed at rung 7's gate on the uploaded bytes, so rung 8 has to
say which bytes the hash covers or rung 9's first retag invalidates it.

---

## 6. Prior art for reading and writing FLAC tags in Go

Scored on CLAUDE.md's four criteria. Repository facts come from the
GitHub API on 2026-09-07; package facts from pkg.go.dev; code judgments
from reading the source at the named tag.

| Entry | Role | License | Quality against the house bar | Maintenance | Adoption | Verdict |
|---|---|---|---|---|---|---|
| `mewkiz/flac` and its `meta` package | FLAC decode; metadata block decode; a full-stream encoder | Unlicense | Idiomatic package split; the encoder is honest about what it is | v1.0.14, 2026-08-19; last push 2026-08-19; 9 open issues, the oldest from 2018; 13 contributors | 361 stars, 96 importers; already adopted for rung 7's integrity check | Adopt for reading and for frame decode. Decline for tag writing. |
| `go-flac/go-flac` v2 | Metadata block split and reassembly, with an in-place save | Apache-2.0 | Fails. `saveInPlace` is documented as using a temporary file and uses none | v2.0.4, 2025-08-06; 0 open issues; 3 contributors | 44 stars, 8 importers | Decline as a dependency. Read as a design reference for the padding-shift problem. |
| `go-flac/flacvorbis` v2 | VORBIS_COMMENT parse and marshal | Apache-2.0 | Mixed. Endianness correct; no delete or replace; `SCREAMING_SNAKE` exported constants | v2.0.2, 2024-08-02; last push 2024-08-02; 1 open issue from 2021-10-31; 2 contributors | 13 stars, 6 importers | Decline. |
| `go-flac/flacpicture` v2 | PICTURE block parse and marshal | Apache-2.0 | Not read in depth; declined with its base | v2.0.2, 2024-08-02; 0 open issues; 2 contributors | 8 stars, 4 importers | Decline. |
| `dhowden/tag` | Read-only tag detection and parse across MP3, MP4, OGG, FLAC | BSD-2-Clause | Not read in depth for this rung | No tagged release; pseudo-version 2024-04-17; last push 2024-07-30; 32 open issues; 17 contributors | 654 stars, 351 importers | Not needed at rung 8. Already scored for 2.0's read path by F1. |
| `bogem/id3v2` | MP3 tag writing | MIT | Not read | v2 line, last push 2023-10-28; 24 open issues; 10 contributors | 370 stars | Out of scope at rung 8 by ruling 1. Named by F1 for the later MP3 pass. |
| `jacopo-degattis/flacgo` | High-level FLAC tag writing | GPL-3.0 | Not read | last push 2025-08-18 | 0 stars | Decline on license alone. |
| `ambeloe/oggv` | Ogg Vorbis comment writing | AGPL-3.0 | Not read | last push 2024-03-05 | 1 star | Decline on license alone. |
| `go.senan.xyz/taglib` | Tag read and write through wazero | LGPL-2.1 | Recorded by F1 | Recorded by F1 | 11 importers | Already declined by ruling F1. The wazero spike is retired by ruling 2. |

### 6.1 The two findings that decide the path

**`mewkiz/flac` cannot rewrite tags in place, by design.** Its `meta`
package parses blocks and exposes `StreamInfo`, `VorbisComment`,
`Picture`, and the rest, with `New`, `Parse`, `Block.Parse`, and
`Block.Skip` as the whole API: no block encoder is exported. Encoding
exists one level up, in the package root: `encode_meta.go` holds
`encodeBlock`, `encodeStreamInfo`, `encodeApplication`, and their
siblings, and they are unexported and reached only through
`flac.NewEncoder(w, info, blocks...)` (`encode.go`), which writes the
signature and the metadata blocks and then expects the caller to write
frames through the encoder, recomputing the MD5 over unencoded samples
as it goes. Retagging through that API means decoding and re-encoding
every audio frame. That is the opposite of what section 5 requires. The
library is well kept for what it does: its 1.0.14 release notes record
parsing hardened against corrupt and malicious FLAC files, which is the
reading half rung 7 wants.

**`go-flac/go-flac` v2's in-place save is dangerous.** `saveInPlace`
carries the comment "performs a safe overwrite of the original file
using a temporary file" and then does no such thing: it truncates the
file to a larger size, shifts the whole audio region forward in
256 KiB chunks with `ReadAt`/`WriteAt`, and finally writes the new
metadata at offset 0. A crash or a full disk part-way through the shift
leaves the audio region permanently interleaved with itself, with no
temporary file and no rename to recover from. The v1-shaped code path
in the module root is the other extreme: `Frames` is a `[]byte`, and
`Save` is `ioutil.WriteFile(fn, c.Marshal(), 0644)`, so tagging a 2 GB
file allocates 2 GB and writes non-atomically. Neither passes the
house bar, and the second is the exact failure the repo's
`internal/atomicfile` exists to prevent.

Both libraries also miss the cheaper option entirely. Neither uses the
PADDING block to absorb a metadata size change, which is the mechanism
the format provides for precisely this operation (section 8.3).

### 6.2 The recommendation

Extend the repo's own reader and write the writer. The reading half
already exists and is proven against a committed fixture: nineteen
fields, `safe_cast` and date semantics reproduced, block walking that
skips unknown types by length. What is missing is a value model wide
enough for the 69 keys, list and float and Q-number casts, PICTURE, and
the write path itself. Adopting `go-flac` would replace roughly 200
lines of block walking with a dependency whose save path we would have
to route around anyway, and adopting `mewkiz/flac` for tags would mean
re-encoding audio. Neither trade is worth making.

This also matches F1's shape: pure-Go per-format readers and writers
under permissive licenses, extending the repo's own FLAC metadata
reader. `mewkiz/flac` stays adopted for rung 7's decode-integrity
check, where it supplies exactly the audio-frame decode the repo's own
reader deliberately lacks.

---

## 7. The atomic write path

`internal/atomicfile` was re-landed unchanged from `_salvage` at pass
4a Task 1 (`bc21683`, fix `a1d3d6c`), on merit under the exemplary-Go
bar (`docs/HISTORY.md:493-495`). It is 42 lines of implementation and
one exported function:

```go
func WriteFile(path string, data []byte, perm os.FileMode) error
```

`WriteFile` (`internal/atomicfile/atomicfile.go:15-42`) creates a temp
file in `path`'s own directory with `os.CreateTemp(dir,
filepath.Base(path)+".tmp-*")`, defers its removal, writes, `Sync`s,
`Close`s, `Chmod`s to `perm`, and `Rename`s over the target. Its tests
(`atomicfile_test.go`) cover creation, overwrite, permissions, empty
content, no leftover temp file, the temp file's directory, a missing
directory, and that a failed write leaves the old content intact.

Two properties matter for rung 8.

**It takes `[]byte`.** A tag rewrite of a multi-gigabyte FLAC through
`WriteFile` reads and buffers the whole file. That is the same defect
`go-flac`'s v1 `Save` has, and shipping it would be worse, because the
estate's library is exactly where multi-gigabyte files live. The 2.0
spec already anticipates this and states the intended fix in its
prior-art table: extend `internal/atomicfile` with a streaming write
rather than inventing a second temp-file convention
(`streamer-2-0-design.md`, the `internal/atomicfile` row). Rung 8 is
the first rung that needs it.

**It does not fsync the containing directory after the rename.** The
data is durable and the rename is atomic with respect to readers, but
the directory entry itself is not forced. For the shadow's small JSON
records that is a defensible omission. For a library file it is worth
naming explicitly rather than inheriting.

**The two write shapes rung 8 needs.**

- *Full rewrite through a streaming atomic writer*: open a temp file in
  the library directory, write signature and new metadata blocks, copy
  the frame region with `io.Copy`, sync, close, rename. Cost is one
  full file copy and one file's worth of free space. Safe under a
  crash.
- *In-place PADDING rewrite*: when the new metadata region fits in the
  old region plus its trailing PADDING, write the new blocks and the
  shrunken PADDING over the head of the existing file. Cost is a few
  kilobytes. Not crash-atomic, and it cannot use `atomicfile` at all.

The honest default is the full rewrite, with the in-place path as a
later optimization gated on evidence that the rewrite cost matters. A
writer that reserves a PADDING block on every write keeps the
optimization available without taking it now.

**What rung 9 will ask for.** Three call shapes, from the ladder entry
and the corpus's own model of the pipeline:

1. *Write on admit*, once per track, immediately after the match is
   accepted and before the file moves to its library path. Writing
   before the move means the temp file and the target share a
   directory, and the move is a separate step.
2. *Retag against a named MusicBrainz release*, a launch-bar repair
   verb. This rewrites tags on a file already in the library, which is
   the case where the in-place path pays and where the content-hash
   question from section 5 lands.
3. *Move*, which is a rename and touches no tag, but which shares the
   same-filesystem assumption the atomic rename depends on.

All three want one function that takes an open file or a path and a tag
set, so the rung 8 seam is a writer over one file, never a pipeline
step.

---

## 8. Config, territory, and the gate

### 8.1 Config keys: none

`internal/config/keys.go` carries the whole owned wall, the reserved
classes, and the superseded names. The reserved classes
(`keys.go:57-140`) are auth and identity (rung 5), serve (5), backup
targets (6), intake (7), pipeline (9), monitoring (9), streamer (12),
and decision and review queues (13). There is no rung 8 class, and
nothing in the wall names tagging.

That is correct and should stay correct. Tag I/O has no operator-facing
setting: the field mapping is the beets contract, and the one policy
question near it, whether the pipeline embeds cover art by default, is
rung 9's knob by owner ruling 3. The rung 8 spec should say "this rung
declares no configuration key" in those words, so the absence is a
decision rather than an oversight.

### 8.2 Territory

From `docs/superpowers/research/2026-09-04-reverse-edge-completion.md`,
cell (8,3) at `:161-190`:

| Surface | Owner | Consumers |
|---|---|---|
| Rung 8's applied-field conformance instrument | rung 8, under `internal/engine/tagio` with its own `testdata`, read-only against `corpus/analysis/file_tags.json` and `internal/harness/oracle/testdata/template.flac` | rung 15's re-run. Rung 8 pins no schema version of its own and never calls `Reconcile`. |
| Additional derived `file_tags.json` subjects | main tree, venv-bound, `make derive`; not a worktree lane's to take | if rung 8 needs them, they are scheduled ahead of rung 3 taking the primary tree, or as a named task in that lane |

The constraint sentence is quoted verbatim in the same cell: a lane
runs beside rung 3 only while its plan states that its conformance
instrument is read-only against `corpus/` and `internal/harness`, and a
rung 8 task that would derive a new corpus record moves to the primary
tree. Rung 3 has since closed, so the contention that motivated the
rule is gone, but the ownership rows stand: the instrument lives under
`internal/engine/tagio`, and any new derived subject is `make derive`
work under the venv in the main tree.

Cell (8,4) at `:174-190` is argued for NONE and stopped at SOFT for two
reasons that are both now discharged or scheduled: rung 4's deploy is
merge point M1, and the adjudication artifact's home. The cell's
standing note is the one rung 8 still carries: a landed but unwired
`tagio` ships as unreachable code inside the deployed shadow binary,
which is harmless and worth one sentence rather than a discovery at
deploy time.

One gap the same cell names and nothing has closed: the record never
says where rung 8's applied-field disagreements go when the resolution
is an owner question. Rung 3 built a tier 2 pending-row escape for
exactly that situation, and nothing grants rung 8 one.

### 8.3 The gate

`bash scripts/check.sh` is the gate and the code lane owns it during
this window. Its legs are `make check` over the module, shellcheck over
`scripts/`, ruff `check` and `format --check` over every tracked
`*.py`, gitleaks over the tracked worktree, vale over Go comment prose,
`docker compose config`, and a pattern-filtered `systemd-analyze
verify`.

Rung 8 touches one leg and adds nothing. `internal/engine/tagio` is Go
under `./...`, so `make check` covers it. Two shapes would extend the
gate and should be declined unless the spec argues for them: a
conformance run that shells out to the venv (the gate never requires
uv or beets, by design), and a new binary fixture that gitleaks and the
repo's data-directory policy would have to account for. The instrument
reads `corpus/analysis/file_tags.json` and the committed 42-byte
template, both already tracked, so neither is needed.

### 8.4 The 2.0 forward hooks that touch `tagio`

Named in `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`
from `:359` onward:

- **Hook 13** (`:388-389`), sort tags written at admit, rungs 8 and 9.
  `ALBUMARTISTSORT` and `ARTISTSORT` with their ID3 equivalents are
  written once or never; retrofitting means re-tagging the library.
  Rung 8 lands this hook directly: `albumartist_sort`, `artist_sort`,
  `albumartists_sort`, `artists_sort`, and `composer_sort` are all in
  the writable 69, and the table in section 2 gives their keys.
- **Hook 14** (`:390-391`), an ID3 version ruling at rung 8. Ruled by
  Geoff on 2026-09-07: ID3v2.4. Rung 8 records it and writes no ID3.
- **Hook 2** (`:361-362`), the folded sort key on the catalog row,
  rung 6 schema, computed at rung 9's tagging, and it "prefers the sort
  tags of hook 13". Rung 8 is the rung that makes those tags exist.
- **Hook 3** (`:363-364`), the art blob recorded at admit, with
  embedded pictures extracted once and never at serve time. Rung 8 owns
  the PICTURE read and write that hook depends on.
- **Hook 11** (`:383-384`), ReplayGain and R128 gain and peak on the
  catalog row, rung 6 schema, computed at rung 9. Six of the writable
  69 are those values, and section 2 gives their exact serialization,
  including the `" dB"` suffix and the Q8 shift.
- **Hook 1** (`:359-360`), the content hash, and **hook 12**
  (`:385-387`), the portable layout whose "unsanitized title stays in
  the tag". Both constrain what a tag write may do rather than what it
  writes.

Per CLAUDE.md's build-for-2.0 rule, the rung 8 plan states per task
whether it lands a hook or is neutral. On this reading the writer core
lands hooks 13 and 14, the PICTURE work lands hook 3, and the
conformance runner is neutral.

---

## 9. Budget evidence and a pass-split proposal

### 9.1 What comparable rungs cost

Figures from `docs/HISTORY.md`, per pass, subagent tokens.

| Pass | Band | Spend | Ceiling | Notes |
|---|---|---|---|---|
| 2d (file-level applied tags) | part of rung 2 | ≈1.6M | 2.5M | `:2506-2510`. Three tasks plus `code-simplifier`. The pass that produced this rung's conformance target. |
| 3a (scoring core, tier 0) | rung 3, L | ≈2.84M | 6M | `:2330-2335`. Three batches, simplify, close draft, prose review, fold. |
| 3b (translation layer) | rung 3, L | ≈4.82M | 6M | `:2016-2020`. Execution alone closed at 2.7M; prep was the rest. |
| 3c (client, limiter) | rung 3, L | ≈4.9M, summed from the entry's lines | 6M | `:1639-1646`. Prep 2.55M, chain 1.74M, the remainder simplify, review, close. |
| 3d (verdicts, gate) | rung 3, L | ≈5.15M | 6M, re-derived to an expected 4.8M and not raised | `:1229-1237`. Prep 2.51M, three runs, fix round, runner review. |
| 4a-i (shadow runner, Tasks 1 to 9) | rung 4, M | ≈4.99M | 5M for all of 4a | `:746-755`. Prep 2.3M, chain 2.14M. Execution against the 3M sub-ceiling was 2.69M. |
| 4a-ii (Tasks 10 to 15) | rung 4, M | ≈2.7M | 3M | `:468-481`. Chain 1.32M across 16 agents, staleness and extraction 0.26M, fix and review rounds 0.49M, simplify 0.18M, close 0.16M, conductor 0.2M. |

Two patterns hold across all seven. Prep, meaning plan authoring plus
the staleness check, runs 0.25M to 2.55M and is the single largest line
in the L-band passes. Execution through the chain runs 1.3M to 2.2M for
a nine-to-fifteen-task pass. Simplify plus its review is consistently
about 0.2M, and the close draft plus prose review plus fold is 0.15M to
0.5M.

Rung 4 is the closest comparable: same M band, same shape of pure-Go
library work with an external contract to conform to, and it split into
two passes at 5M and 3M for fifteen tasks total.

### 9.2 Why rung 8 splits into three

Rung 8 carries four deliverable clusters that do not share a natural
review boundary: the container writer, the 69-key value model, PICTURE,
and the conformance runner with its absorption of the shadow's reader.
That is more than the roughly four-deliverable line CLAUDE.md's
pass-sizing rule sets for a single task, several times over, and the
clusters have a real dependency order. Three passes, each with its own
ceiling and its own gate:

**Pass 8a, the writer core.** The block walker extended to write, the
full metadata block model including preserve-unknown-by-length, the
streaming atomic write path (section 7's first shape, and the
`internal/atomicfile` streaming extension the 2.0 spec already
sanctions), the PADDING reservation, and the 69-key value model with
every serialization rule from section 2: the write-every-writable-style
rule, list repetition, the date decomposition, the Q8 shift, the float
formatting and the `" dB"` suffix, the bool encoding, and the
three-state empty-versus-absent rule. Tests are Go-side and
hand-authored, in the manner of `flac_test.go`'s builder. Lands hooks
13 and 14. **Ceiling 3M.**

**Pass 8b, PICTURE and the conformance runner.** The PICTURE block read
and write, front-cover selection and the strip path per owner ruling 3,
MIME sniffing from the bytes, and then the conformance instrument under
`internal/engine/tagio` with its own `testdata`, read-only against
`corpus/analysis/file_tags.json` and the committed template, grading
all twelve subjects over the 69 keys. This is the pass where a
divergence against the ledger surfaces, so it carries the adjudication
question section 8.2 flags. Lands hook 3. **Ceiling 3M.**

**Pass 8c, absorb and repoint.** The shadow's provisional reader moves
into `tagio`, the three call sites repoint, the fixture builder finds
one home, and `internal/engine/shadow` keeps the projection to
`matcher.ItemFields`. Mechanical, reviewable by the in-chain
`diff-reviewer`, and small enough that it could ride 8b if 8b comes in
light. **Ceiling 1M.**

Total 7M across three passes, against 8M for rung 4's two. That is the
right direction: rung 8 has a narrower external surface than the shadow
did, no systemd or deploy work, and no box contact.

The split point that matters is between 8a and 8b. A writer that has
never been graded against the ledger is the risk, and putting the
grading in its own pass with its own fresh-context review is what
catches a value model that is subtly wrong on twenty keys at once.

---

## Surprises

Things the tree or the venv says that the handoff, the ladder entry, or
a carried note does not.

1. **Seventeen of the 69 keys are never exercised by the ledger.**
   `corpus/README.md` presents the twelve subjects over the 69-key
   writable set as the conformance target without qualification. Across
   all twelve rows, `acoustid_fingerprint`, `acoustid_id`, `bpm`,
   `comments`, `encoder`, `genres`, `grouping`, `initial_key`,
   `lyrics`, both R128 fields, `remixers`, all four ReplayGain fields,
   and `subtitle` are empty or absent. Those are the keys with the
   float formatting, the suffix, and the Q8 shift. Grading only against
   the ledger leaves the trickiest quarter of the surface unproven.

2. **The ledger grades three distinct write behaviors, not two.** An
   empty string writes an empty comment and reads back as `""`; a null
   deletes the key and reads back as `None`. Row 0 carries ten of each.
   Nothing in the handoff mentions this, and a writer that treats
   "empty" and "absent" alike fails twenty keys per row.

3. **The PICTURE block is outside the conformance target entirely.**
   `images` and `art` are `MediaFile` fields but not beets `Item`
   fields, measured on the venv, so they are excluded from
   `Item._media_fields` and therefore from the 69. Owner ruling 3 puts
   PICTURE write and strip at launch, so that work needs its own
   evidence rather than a corpus row.

4. **Three writable fields collide on a Vorbis key, and sort order
   decides the winner.** `albumtype` and `albumtypes` both write
   `RELEASETYPE` and `MUSICBRAINZ_ALBUMTYPE`; `mb_artistid` and
   `mb_artistids` both write `MUSICBRAINZ_ARTISTID`;
   `mb_albumartistid` and `mb_albumartistids` both write
   `MUSICBRAINZ_ALBUMARTISTID`. `MediaFile.update` iterates
   `sorted_fields()`, the plural sorts last, and the plural wins. A Go
   writer that iterates a map gets a non-deterministic result on three
   keys.

5. **`apply_metadata` is not a function in 2.13.1.** The ladder entry
   and the 2d handoff both say "the apply-metadata field mapping",
   which reads as a module-level function to look up. It is two
   dataclass methods plus `merge_with_album` plus the `raw_data` and
   `item_data` properties, and the actual mapping lives in the two
   `MEDIA_FIELD_MAP` class variables.

6. **The nineteen-field mapping is not where the carry says it is.**
   `docs/HISTORY.md:442-447` states that the mapping with its per-row
   citations "travels in 4a-i's entry". The 4a-i entry (`:504`) names
   it in one clause and carries no table. The table is in the pass 4a
   plan (`plans/2026-09-05-rung-4a-shadow-runner.md:2453-2478`) and
   inline in `internal/engine/shadow/flac.go:210-228`.

7. **The shadow's fixture builder is shared by three test files.**
   `buildFLAC` lives in `flac_test.go` and is used by `capture_test.go`
   and `run_test.go` as well. Absorbing the reader is not a
   file-for-file move.

8. **A `tagio` returning `matcher.ItemFields` contradicts the tree's
   own stated split.** `internal/engine/tag/doc.go:8-11` says `tagio`
   "knows nothing about matching". `flac.go` imports
   `internal/engine/matcher`.

9. **`internal/atomicfile` has no streaming write.** The only exported
   function takes `[]byte`. The 2.0 spec's prior-art table already
   commits to extending it rather than inventing a second convention,
   which makes that extension rung 8's, not a later rung's.

10. **`go-flac` v2's `saveInPlace` does not do what its own doc comment
    says.** It claims a temporary file and uses none, shifting the
    whole audio region in place. This is worth recording because the
    library is the obvious candidate a future reader would reach for.

11. **CLAUDE.md's gate description undercounts the Python files.** It
    says "six today"; `git ls-files '*.py'` returns eight. The gate
    itself globs tracked files and is unaffected. Worth a one-line
    correction at some close, not this rung's.

12. **No rung 8 reserved config class exists, and that is right.**
    `internal/config/keys.go:57-140` has classes for rungs 5, 6, 7, 9,
    12, and 13 and none for 8. Stating the absence deliberately is
    better than leaving a reader to wonder.

## Candidate rulings

Each is a reading the rung 8 spec has to settle. Every one carries a
recommendation. Only the two marked as product forks change what an
adopter sees or configures; the rest are conductor rulings under the
standing grant.

1. **What grades the seventeen ledger-silent keys.** Recommend
   Go-side unit tests derived from the mediafile source, one per
   serialization rule, plus a named deviation note, rather than
   deriving new corpus subjects. Deriving new subjects is main-tree
   `make derive` work under the venv and would re-open a frozen
   corpus record for keys a MusicBrainz match never supplies. Not a
   product fork.

2. **Where an applied-field disagreement goes when the answer is an
   owner question.** The reverse-edge document states that rung 3 has a
   tier 2 pending-row escape and rung 8 has none
   (`2026-09-04-reverse-edge-completion.md:192`). Recommend that
   rung 8 write a divergence record beside its conformance run and
   route it into `corpus/analysis/adjudications.json` through the same
   schema-addition path rung 4 uses, never by hand-editing a derived
   record. Not a product fork.

3. **Whether `tagio` depends on `internal/engine/matcher`.** Recommend
   no. `tagio` exposes its own tag set type; `internal/engine/shadow`
   keeps the projection to `matcher.ItemFields`. This preserves
   `internal/engine/tag/doc.go`'s stated split and keeps the writer
   usable by rung 9 without dragging the matcher in. Not a product
   fork.

4. **Where `buildFLAC` lives after the absorption.** Recommend it moves
   to `tagio` as an exported test helper in a small `tagiotest` package
   or, if that reads as ceremony, that `shadow`'s three uses switch to
   `tagio`-written fixtures built by the real writer. The second option
   is stronger: once a writer exists, a hand-rolled container builder
   in a test is a second implementation of the thing under test. Not a
   product fork.

5. **In-place PADDING rewrite at launch, or full rewrite only.**
   Recommend full rewrite only at launch, with a PADDING block reserved
   on every write so the in-place path stays available, and the
   in-place path deferred until a measurement says the copy cost
   matters. Not a product fork.

6. **The PADDING reservation size.** Recommend 8 KiB rather than
   mutagen's 1 KiB. The 69-key set with MusicBrainz ids and credit
   lists runs well past a kilobyte on a real album (the measured
   fifteen-field write already produced a 575-byte VORBIS_COMMENT
   block), and a reservation too small to absorb a retag makes the
   optimization dead on arrival. Not a product fork.

7. **Whether the writer preserves unknown metadata blocks.**
   Recommend yes, unconditionally: APPLICATION, SEEKTABLE, CUESHEET,
   and reserved types are copied through by declared length. This is
   the shadow reader's existing behavior and the format's expectation.
   Not a product fork.

8. **What the content hash covers.** Forward hook 1 puts a content hash
   on the catalog row computed at rung 7's gate; a tag rewrite changes
   the file's bytes but not the frame region's. Recommend that rung 8
   state the frame region as the hashable extent and expose the
   offsets, so rung 6's schema and rung 9's retag verb agree. **Product
   fork**, because it decides whether an adopter's drift check survives
   a retag.

9. **Whether the writer refuses a container it did not fully
   understand.** Recommend refusing: a file with more than one
   VORBIS_COMMENT block, a STREAMINFO that is not first, or a declared
   length running past the end is an error and never a best-effort
   write. Section 5's uniqueness rules make each of these a spec
   violation rather than an edge case. Not a product fork.

10. **How a non-UTF-8 or control-character tag value is handled.**
    RFC 9639 section 8.6.4 restricts a field name to U+0020 through
    U+007E excluding U+003D, and requires case-insensitive evaluation.
    Recommend that the writer
    reject an invalid field name and replace an invalid UTF-8 sequence
    the way beets' own `safe_cast` does, decoding with errors ignored,
    so the Go writer and the oracle agree. Not a product fork.

11. **Whether `internal/atomicfile` grows the streaming write in this
    rung.** Recommend yes, in pass 8a, as an added exported function
    beside `WriteFile` rather than a new package. The 2.0 spec already
    named this as the intended shape. Not a product fork.

12. **Whether the streaming write fsyncs the containing directory.**
    Recommend yes for the new streaming path, and leaving `WriteFile`
    alone. A library file that survives a power cut but whose directory
    entry does not is a worse failure than the extra syscall. Not a
    product fork.

13. **Whether rung 8 declares a config key.** Recommend none, stated in
    those words. The cover-art embed policy is rung 9's knob per owner
    ruling 3, and the accepted-format allowlist is rung 7's, bounded by
    what rung 8 supports. Not a product fork.

14. **How the per-format seam is shaped.** Owner ruling 1 requires MP3
    and MP4 writers to land as their own later passes, and the sibling
    container-needs research adds that the seam dispatches by container
    to a tag-standard implementation, because five tag standards cover
    fourteen containers, and that rung 7's allowlist reads the
    supported set from the seam rather than restating it as a config
    default. Recommend adopting both, implemented as a registry keyed
    by container with one entry at rung 8, not an exported interface;
    per the repo's Go conventions the interface arrives when the second
    implementation does. Not a product fork.

15. **Whether to adopt the container-needs recommendation.** That
    research landed during this inventory's own writing, as
    `research/2026-09-07-rung-8-container-needs.md`, so the question is
    no longer whether to run it. It recommends launching FLAC-only,
    then MP3, then MP4 covering AAC and ALAC, and it proposes erratum
    wording superseding F1's "at launch" clause. Recommend the spec
    carries that recommendation to Geoff unchanged, since it is his
    ruling to take, and that the three launch-scope items it names (the
    container-keyed seam, rung 7's allowlist reading the seam, and
    per-format reject wording) are treated as binding on rung 8's own
    surface. **Product fork**, because the answer decides which formats
    an adopter may upload at 1.0.
