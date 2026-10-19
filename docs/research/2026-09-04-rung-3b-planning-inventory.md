# Rung 3, Pass 3b Planning Inventory

Research report, 2026-09-04, prepared for pass 3b's plan authoring,
against HEAD `f40a6694122c7077202caaff5a3598b2eca5afa8`. Scope, quoted
from the rung 3 spec's pass-split table
(`docs/specs/2026-09-03-rung-3-go-matcher.md:329`):

> **3b** Translation | The translation layer and the pseudo-release
> merge; the derived `AlbumInfo` ledger, its Python entry point, its
> gate-resident checker, and its oracle-tagged re-derive test (ruling
> E); the `defect_free` block, its derivation, and its checker (ruling
> A / M6). | 3a's info types; the venv; 2d closed. | Main tree: writes
> `corpus/` and `internal/harness/oracle`.

Three sections follow: (a) beets' translation-layer mechanics, (b) the
harness and corpus as landed, (c) the engine as landed and the
proposed 3b Go surface. A fourth section, (d), reconciles all three
and lists the open questions each raised.

---

## (a) Beets mechanics of the translation layer

Throughout, `$SP` abbreviates
`/var/home/glw907/Projects/dubplate/tools/venv/.venv/lib/python3.13/site-packages`.
Every line number below was read off that tree at pinned beets 2.13.1,
not off the port map. Where a claim was not obvious from reading, it
was checked with an offline probe against the committed corpus; the
probes and their outputs are named in section 12.

### 1. Scope and the one entry point

`MusicBrainzPlugin.album_info` (`$SP/beetsplug/musicbrainz.py:631-695`)
is the whole album-side translation layer. It is reached only through
`album_for_id` (`$SP/beetsplug/musicbrainz.py:765-800`), and **search
candidates reach it too**: `SearchApiMetadataSourcePlugin.candidates`
maps each search hit's id through `albums_for_ids`
(`$SP/beets/metadata_plugins.py:442-446`), which is a generator over
`album_for_id` (`$SP/beets/metadata_plugins.py:250-259`). So every
`AlbumInfo` the matcher ever scores has been through `album_info`
**and** through the pseudo-release check, one full release lookup per
candidate. The pseudo merge sits on the universal path, not as a
special case reserved for direct ID matches. This matters for 3b's
grading design: there is exactly one translation function to pin, and
the merge sits inside it rather than beside it.

The matcher itself never touches raw WS data. `beets/autotag/distance.py`
reads only `AlbumInfo`/`TrackInfo` attributes.

### 2. Key normalization

`MusicBrainzAPI._normalize_data`, a `singledispatchmethod` with three
registrations: the scalar no-op
(`$SP/beetsplug/_utils/musicbrainz.py:724-733`), the list handler
(`:735-744`), and the dict handler (`:746-775`). It runs inside
`get_json` (`:621-623`), so every response the plugin sees is already
normalized.

The dict handler's exact rule, `:758-775`:

- Iterate `list(data.items())` in source order.
- For the key `relations`: sort the array by
  `item.get("target-type")`, `groupby` on the same key, and write each
  group to `output_data[f"{target_type.replace('-','_')}_relations"]`,
  dropping the `target-type` field from each entry, then recursing
  into the group.
- For every other key: `output_data[k.replace("-", "_")] =
  cls._normalize_data(v)`.

Four properties a port has to reproduce, and one it can get wrong by
analogy.

- **The hyphen rewrite is indiscriminate and recursive.** It rewrites
  data keys, not only schema keys. Verified:
  `{"text-representation": {"a-b": 1}, "attribute-ids": {"guest-vocal":
  "x"}}` becomes `{"text_representation": {"a_b": 1}, "attribute_ids":
  {"guest_vocal": "x"}}`. A relation attribute literally named with a
  hyphen is renamed.
- **Grouping order is `sorted` then `groupby`**, so relation-group keys
  appear in ascending target-type order and entries stay in response
  order within a group. Verified: interleaved `url`/`artist`/`url`/`artist`
  relations yield keys `['artist_relations', 'url_relations']` with
  intra-group order preserved.
- **A `relations` key destructively collides with a pre-existing
  `<type>_relations` key.** Whichever appears later in the source
  dict wins, in both directions. Verified both orders.
- **A relation with no `target-type` raises** rather than being
  skipped: `AttributeError: 'NoneType' object has no attribute
  'replace'` at `:769`. With two or more mixed None/str target types,
  `sorted` at `:763` raises `TypeError` first.
- The read hazard: the inner comprehension at `:766` rebinds `k, v`
  inside the outer `for k, v in list(data.items())` loop. Python 3
  comprehension scoping keeps the outer names intact. A Go port
  written by transliteration is fine; a Python-to-Go port that hoists
  the inner variables is not.

### 3. `album_info`, step by step

`$SP/beetsplug/musicbrainz.py:631-695`.

| Step | Line | Input | Output | Config read |
|---|---|---|---|---|
| Browse fallback | `:635` (`_ensure_complete_recordings`, `:602-629`) | `release["media"][*]["tracks"]` | mutates `track["recording"]` in place | none |
| Medium filter | `:638-640` | `release["media"][*]["format"]` | `valid_media` | `match.ignored_media` |
| Track extraction | `:641-643` | `valid_media` | `track_infos` | see section 4 |
| Index assignment | `:645-646` | `track_infos` | `track_info.index` 1-based across **all** valid media | none |
| Release title | `:648` | `release["title"]`, `release["aliases"]` | `album` | `import.languages`, `import.ignored_alias_types` |
| Artist credits | `:650` | `release["artist_credit"]` | 8 fields, section 5 | as above, plus `musicbrainz.aliases_as_credits` |
| Media derivation | `:654-658` | `{t.media for t in track_infos}` | `media` | none |
| Mediums | `:659` | `len(release["media"])` | `mediums` | none |
| Scalars | `:660-668` | `barcode`, `text_representation.script`/`.language`, `asin`, `status`, `disambiguation` | same-named fields | `musicbrainz.genres`, `musicbrainz.genres_tag` (via `_parse_genres`, `:509-525`) |
| Release group | `:669` (`_parse_release_group`, `:472-494`) | `release["release_group"]` | `albumtype`, `albumtypes`, `releasegroup_id`, `release_group_title`, `releasegroupdisambig`, `original_year/month/day` | alias keys |
| Labels | `:670` (`_parse_label_infos`, `:497-507`) | `release["label_info"][0]` | `label`, `catalognum` | none |
| External ids | `:671` (`_parse_external_ids`, `:527-553`) | `release["url_relations"]` | `<source>_album_id` | `musicbrainz.external_ids.*` |
| VA override | `:673-681` | `artist_id == VARIOUS_ARTISTS_ID` | `va`, and six artist fields overwritten | `va_name` |
| Release event | `:684-689` | `release["release_events"]`, `country`, `date` | `country`, `year`, `month`, `day` | `match.preferred.countries` |
| Plugin extras | `:691-693` | `mb_album_extract` senders | arbitrary `info.update` | loaded plugin set |

Three details that are easy to lose.

**`mediums` counts every medium, including the ones `ignored_media`
dropped.** `:659` reads `len(release["media"])`, the unfiltered list,
while `:638-640` filters for track extraction. Verified on
`corpus/cases/media-baseline`: `ignored_media=[]` gives `(mediums=2,
tracks=6, media='CD')`; `ignored_media=['CD']` gives `(mediums=2,
tracks=0, media='Media')`.

**`media` can become the literal string `"Media"`.** `:654-658` reads
`medias.pop() if len(medias) == 1 else "Media"`. Zero surviving tracks
makes the set empty (length 0), and mixed per-track media makes it
larger than 1, so both land on the four-character sentinel. Ten of the
303 committed release lookups produce `media == "Media"`, and five of
those releases have zero tracks. This is truthy, so the media penalty
at `distance.py:455` **fires** on it and scores `"Media"` as if it
were a media name, against `preferred.media` regexes or against
`likelies["media"]` equality.

**`ignored_media`, `ignore_data_tracks`, `ignore_video_tracks`,
`aliases_as_credits` and `genres_field` are `@cached_property` on the
plugin instance** (`:246-268`). They are read once per instance per
process, not per call. The derivation harness must construct a fresh
plugin after any config change; a probe that mutates config between
two calls on one instance silently reads the stale value (this bit the
first version of my own probe). `_preferred_alias` and
`_preferred_release_event` read config fresh on every call, so the two
families behave differently under config mutation.

### 4. Per-medium track extraction

`get_tracks_from_medium`, `$SP/beetsplug/musicbrainz.py:555-600`.

Assembly order, `:556-563`: `pregap` first (unconditionally, no config
gate), then `medium["tracks"]`, then `medium["data_tracks"]` **only
when `ignore_data_tracks` is false**. All three use `.get`, so a medium
missing any of them contributes nothing rather than raising.

Filter, `:571-580`, applied to the assembled list:

- Drop any track whose `recording["title"] == "[data track]"`. This
  runs **regardless of `ignore_data_tracks`**, so a data track that
  leaked into the `tracks` array is dropped by title even when data
  tracks are admitted.
- Drop any track with `recording["video"]` true when
  `ignore_video_tracks` is set.

Verified on a hand-built medium carrying a pregap, one audio track, one
video track, one data track, and one `[data track]`-titled data track:

| `ignore_data_tracks` | `ignore_video_tracks` | titles yielded |
|---|---|---|
| true | true | `['Pregap', 'A']` |
| false | true | `['Pregap', 'A', 'D']` |
| true | false | `['Pregap', 'A', 'V']` |
| false | false | `['Pregap', 'A', 'V', 'D']` |

Track-level overrides, `:582-590`, applied to a shallow `copy()` of the
recording:

- `recording["length"] = track["length"] or recording["length"]`. A
  track length of `0` or `None` falls back to the recording length.
  Verified: `track.length=0` and `track.length=None` both yield
  `length=200.0` from a 200000 ms recording, while `track.length=111000`
  yields `111.0`.
- `recording["artist_credit"] = track["artist_credit"] or
  recording["artist_credit"]`. An empty list falls back.
- `recording["title"] = track["title"]` **only if** `track["title"]`
  is truthy **and** `_preferred_alias(recording["aliases"])` returns
  nothing.

That last line is the sharpest interaction in the layer, and it means
`_preferred_alias` runs **twice per track**: once at `:589` on
`recording["aliases"]`, and again inside `track_info` at `:443` via
`_key_with_preferred_alias`. Verified precedence, with a Japanese
primary alias on the recording:

| `import.languages` | recording aliases | `track["title"]` | resulting `TrackInfo.title` |
|---|---|---|---|
| `['ja']` | ja alias present | `'TrackTitle'` | `'AliasTitle'` |
| `[]` | ja alias present | `'TrackTitle'` | `'TrackTitle'` |
| `['ja']` | none | `'TrackTitle'` | `'TrackTitle'` |
| `['ja']` | none | `''` | `'RecordingTitle'` |

The two alias reads are also asymmetrically defensive. Line `:589`
indexes `recording["aliases"]` directly, while `_key_with_preferred_alias`
uses `obj.get("aliases", [])` (`:161`). A recording with no `aliases`
key raises `KeyError: 'aliases'` inside `get_tracks_from_medium`,
verified. `RELEASE_INCLUDES` always requests aliases, so this is
unreachable through a well-formed capture and reachable through a
hand-authored fixture.

Per-medium fields written at `:565-570` and `:593-598`: `medium` (from
`medium["position"]`), `medium_total` (`track_count`), `disctitle`
(`title`), `media` (`format`, which is `str | None`), `medium_index`
(`int(track["position"])`, an explicit cast), `release_track_id`
(`track["id"]`), `track_alt` (`track["number"]`).

**One of those is a direct requirement on 3b.** `AlbumInfo.media` is
derived from the per-track `media` values (`:654-658`), yet
`matcher.TrackInfo` as landed in 3a carries eight fields and no `Media`
(`internal/engine/matcher/info.go:73-105`). The translation layer
therefore has to hold per-track media somewhere the matcher's
`TrackInfo` does not currently offer.

### 5. Artist credits and `_flatten_artist_credit`

The port map names `_flatten_artist_credit`. At the pinned version the
function is `MusicBrainzPlugin._parse_artist_credits`
(`$SP/beetsplug/musicbrainz.py:301-357`); no symbol named
`_flatten_artist_credit` exists in the tree.

Per credit element, `:327-346`:

- `artists_ids.append(el["artist"]["id"])`.
- `alias = _preferred_alias(el["artist"].get("aliases", []))`.
- `artist_object = alias or el["artist"]`, so the alias supplies both
  `name` and `sort_name` when present.
- `credit_artist_object = alias if (alias and self.aliases_as_credits)
  else el`, so the credit string uses the **credited** name from the
  credit element unless `aliases_as_credits` is on.
- For each of the three names, `parts.extend([name, joinphrase])` and
  `multi.append(name)`.

Two consequences a port must carry.

**The joinphrase is appended after the last artist too.**
`"".join(artist_parts)` at `:349` includes the final element's
joinphrase. Verified: joins `[' & ', '']` give `'A & B'`, a single
credit with join `' feat. '` gives `'A feat. '` (trailing space and
all), and joins `[' & ', ' & ']` give `'A & B & '`. MusicBrainz
normally leaves the last joinphrase empty, so a malformed credit leaks
a trailing separator straight into the `artist` string that the
weight-3.0 album artist penalty reads.

**`artist_id = artists_ids[0]` at `:350` raises `IndexError` on an
empty credit list**, verified. So `artist` is never `None` and, in
practice, never `""` either, because the `""` case needs an empty
credit list that raises first.

Alias effect on the four artist strings, verified with one credit
carrying a `ja` primary alias:

| `import.languages` | `aliases_as_credits` | `artist` | `artist_sort` | `artist_credit` |
|---|---|---|---|---|
| `[]` | false | `RealName` | `RealName, S` | `CreditName` |
| `['ja']` | false | `AliasName` | `AliasSort` | `CreditName` |
| `['ja']` | true | `AliasName` | `AliasSort` | `AliasName` |
| `[]` | true | `RealName` | `RealName, S` | `CreditName` |

The VA override at `:673-681` then replaces `artist`, `artist_sort`,
`artists`, `artists_sort`, `artist_credit` and `artists_credit` with
`config["va_name"]` whenever `artist_id == VARIOUS_ARTISTS_ID` (`:43`).
Verified both branches and both `va_name` values. **`va_name` is not
verdict-moving.** The album artist penalty is skipped entirely when
`album_info.va` is true (`distance.py:447-448`), and the track artist
penalty reads `track_info.artist`, which the override never touches.
It is ledger-moving, so ruling E's derived `AlbumInfo` ledger records
it even though `Tunables` need not.

### 6. `_preferred_alias`

`$SP/beetsplug/musicbrainz.py:128-155`.

Config reads: `import.ignored_alias_types` at `:136-138`, lowercased
into a set; `import.languages` at `:141`, read only when the caller
passes no explicit `languages`. Both ship empty
(`$SP/beets/config_default.yaml:32` and `:58`), which is what makes the
whole path structurally unreachable under stock and estate profiles,
exactly as ruling D states.

Match rule, `:143-155`: an outer loop over configured locales in order,
an inner loop over aliases in response order, accepting the first
alias where `al["locale"] == locale` **and** `al["primary"]` is
truthy **and** `(al["type"] or "").lower()` is not in the ignored set.
`next(matches, None)` returns the first hit or `None`.

Verified behaviors:

| `languages` | `ignored_alias_types` | selected alias |
|---|---|---|
| `[]` | `[]` | `None` |
| `['ja']` | `[]` | first ja primary |
| `['ja']` | `['artist name']` | falls through to the **next ja primary alias**, not to `None` |
| `['ja']` | `['ARTIST NAME']` | same (the ignored set is lowercased) |
| `['JA']` | `[]` | matches only a `JA` locale, **case-sensitive** |
| `['en','ja']` | `[]` | the en alias (locale order wins over response order) |
| `['ja','en']` | `[]` | the ja alias |
| `['ja']` | `['artist name','search hint']` | `None` |

`al["locale"]`, `al["primary"]` and `al["type"]` are direct indexes, so
a malformed alias raises `KeyError`. `al["primary"]` is `bool | None`
and `None` is falsy
(`$SP/beetsplug/_utils/musicbrainz.py:138-155`). The explicit
`languages` parameter overrides config (verified), but no call site in
the plugin ever passes it, so that branch is dead code in the port's
scope.

`_key_with_preferred_alias` (`:158-162`) wraps it for the three title
fields: release title (`:648`), release-group title (`:487-489`),
recording title (`:443`).

### 7. Release events, country, and date

`_preferred_release_event`, `$SP/beetsplug/musicbrainz.py:165-177`.
Reads `match.preferred.countries` at `:170`. Outer loop over configured
countries, inner over `release.get("release_events", [])`, returning
`(country, event["date"])` for the first configured country with any
event whose `area["iso_3166_1_codes"]` contains it. Otherwise
`(release.get("country"), release.get("date"))`.

Two things a port gets wrong by assuming symmetry with the country
penalty.

**The membership test is a literal, case-sensitive `in` against the
ISO code list.** It compiles no regex. Verified against events for
`GB` and `US`: `['GB']` returns `('GB', '1999-01-02')`; `['gb']`
returns the fallback `('XW', '1980-05-06')`; `['US','GB']` returns
`('US', '2001-03-04')` because configured order wins. The **same
config key** is compiled as a case-insensitive regex list in the
country penalty at `distance.py:496-499`. One
`Tunables.Preferred.Countries` field therefore carries two different
matching semantics in two different packages.

**The returned country is the configured string, not the area's own
code.** With `['GB']` the value written to `info.country` is the
config's `'GB'`, which happens to coincide here and would not if the
config carried a different spelling of a matching code.

`event["area"]` may be `None` and is skipped by the walrus at `:174`;
verified, falls through to the release-level defaults. A release with
neither `release_events` nor `country`/`date` yields `(None, None)`;
verified.

Date parsing, `_get_date` at `:180-200`, splits on `-` and maps each
part through an int parse that returns `None` on `ValueError`.
Verified: `''` gives `(None,None,None)`; `'1999'` gives
`(1999,None,None)`; `'1999-05'` gives `(1999,5,None)`; `'1999-XX-06'`
gives `(1999,None,6)`, a `None` month with a real day; `'199x'` gives
all `None`; `'1999-05-06-07'` gives `(1999,5,6)`, extra parts discarded
silently.

The fallback at `:685-689` tests **truthiness** of `release_date`, not
`is not None`, so an empty-string date falls back to `(original_year,
original_month, original_day)` just as a missing one does. Verified on
a zero-media release: `country=None, year=1999, original_year=1999`,
both years from the release group's `first_release_date`.

### 8. The pseudo-release merge

Trigger, `$SP/beetsplug/musicbrainz.py:785-798`:
`original_release["status"] == "Pseudo-Release"` (exact string) plus
the first `release_relations` entry with `type == "transl-tracklisting"`
**and** `direction == "backward"`. That relation key exists only
because `_normalize_data` regrouped it out of `relations`, so the
merge depends on section 2 having run.

`_merge_pseudo_and_actual_album`, `:203-240`. `merged = pseudo.copy()`,
which is `AttrDict.copy` and therefore a **`deepcopy`**
(`$SP/beets/autotag/hooks.py:93-95`), tracks included. Then
`merged.update(from_actual)` with exactly fifteen keys, `:221-238`:

`media`, `mediums`, `country`, `catalognum`, `year`, `month`, `day`,
`original_year`, `original_month`, `original_day`, `label`, `barcode`,
`asin`, `style`, `genres`.

Each is read with `actual[k]`, a direct index that would `KeyError` on
absence. All fifteen are always present on an MB-built `AlbumInfo`
(confirmed against the full key list in section 9), so the port can
treat them as total.

Of the album fields the distance path reads, the merge copies `media`,
`mediums`, `country`, `year`, `original_year`, `label`, `catalognum`.
It **keeps the pseudo's** `artist`, `album`, `va`, `albumdisambig`,
`album_id`, `data_source` and `tracks`. That is the translated title
plus the real release's physical metadata, and it matches the
docstring's stated intent of not overwriting status or ids.

Note also that `dict.update` does not route through `Info.__setitem__`
(`hooks.py:151-158`), so the legacy-to-list-field rewriting is
bypassed. None of the fifteen is a legacy list field, so nothing
changes today.

Corpus evidence, all offline, over the four committed pseudo cases
(three distinct pseudo releases, every linked actual release captured
in the same case):

| case | fields the merge actually moved |
|---|---|
| `non-latin-pseudo-release` | `asin, barcode, catalognum, country, label, media` |
| `real-nonlatin-girls-generation` | `barcode, catalognum, country, label, media` |
| `real-nonlatin-zemfira` and `-stock` | `catalognum, country, label` |

`merged.album`, `merged.album_id` and `merged.albumstatus` are the
pseudo's in all four; track counts are equal on both sides in all four.
**`year`, `month`, `day`, `original_*`, `mediums`, `style` and
`genres` were already equal on every committed case**, so the merge's
year path has no corpus signal at all.

### 9. What can be `None`, and what each consumer does with it

This is the section that decides 3a's carried question.

An MB-built `AlbumInfo` carries forty-one keys and a `TrackInfo`
thirty-nine (dumped from a committed case, key names only). The
fields the album distance path reads are `artist`, `album`, `va`,
`media`, `mediums`, `year`, `original_year`, `country`, `label`,
`catalognum`, `albumdisambig`, `album_id`, `data_source`, `tracks`
(`$SP/beets/autotag/distance.py:429-542`). The track path reads
`title`, `artist`, `length`, `index`, `medium_index`, `medium`,
`track_id`, `data_source` (`distance.py:383-426`).

Sweep over **all 215 committed cases, 303 release lookups, 6,919
tracks**, built offline through `album_info` with zero exceptions:

| field | can be `None` | corpus count | consumer | guard |
|---|---|---|---|---|
| `album` | no (`:648`, always a `str`) | 0 | `add_string`, **unguarded** | `distance.py:451` |
| `artist` | no (`:349`, `IndexError` before `""`) | 0 | `add_string` | guarded only by `not va`, `distance.py:447-448` |
| `va` | no | 0 | branch | `distance.py:447` |
| `album_id` | no (`release["id"]`) | 0 | `add_equality` | `distance.py:521-524` |
| `data_source` | no (`"MusicBrainz"`) | 0 | `add_data_source` | `distance.py:540` |
| `mediums` | no (`len(...)`, may be 0) | 0 | `add_number` | truthy, `distance.py:468` |
| `media` | **yes** | 8 | `add_priority` / `add_equality` | truthy, `distance.py:455` |
| `year` | **yes** | 10 | `add_ratio` / `add` | truthy, `distance.py:472,480` |
| `original_year` | **yes** | 11 | `or 1889`, `elif` | truthy, `distance.py:475,484` |
| `country` | **yes** | 20 | `add_priority` / `add_string` | truthy, `distance.py:498,501` |
| `label` | **yes** | 102 | `add_string` | truthy, `distance.py:505` |
| `catalognum` | **yes** | 120 | `add_string` | truthy, `distance.py:509` |
| `albumdisambig` | **yes** (`release["disambiguation"] or None`, `:668`) | 254 | `add_string` | truthy, `distance.py:515` |
| track `length` | **yes** (`:450-452`) | 83 | `add_ratio` | truthy, `distance.py:397` |
| track `artist` | no, but can be `""` | 3 empty | `add_string` | truthy, `distance.py:405-409` |
| track `title`, `index`, `medium_index`, `medium`, `track_id`, `data_source` | no | 0 | various | mixed |

**The finding.** Every `AlbumInfo` or `TrackInfo` field that the
distance path can receive as `None` from the MusicBrainz translation
layer is consumed behind a truthiness guard, where `None`, `""` and
`0` are indistinguishable. The two call sites that pass a value to
`string_dist` without a truthiness guard are the album title
(`distance.py:451`) and, inside the `not va` branch, the album artist
(`:448`), and the MusicBrainz plugin cannot produce `None` for either.
So the divergence 3a recorded, Go scoring `StringDist(&"", &"") ==
0.0` where beets scores `string_dist("", None) == 1.0`
(`distance.py:69-72`, weight 3.0), is **unreachable through 3b's
translation layer**. It becomes reachable only through a second
metadata source or a hand-constructed `AlbumInfo`, neither of which
rung 3 has.

The corollary for the item side is already settled:
`internal/engine/matcher/info.go:115-123` records that `data_source`
is the one item field whose true beets default is `None`, and that
both consumers treat `None` and `""` identically.

Non-distance fields that are `None` on a committed case, for the
ruling-E ledger's benefit rather than the matcher's: album `asin`,
`discogs_albumid`, `discogs_artistid`, `discogs_labelid`, `genres`,
`releasegroupdisambig`, `style`; track `album`, `arrangers`, `bpm`,
`genres`, `initial_key`, `remixers`, `work_disambig`. `genres` is
`None` because `musicbrainz.genres` ships false (`:509-510`); `style`
is never written by the plugin at all and survives from
`AlbumInfo.__init__`'s default (`hooks.py:349`).

Sites where a `None` or a missing key raises rather than propagating,
all worth a named decision in the Go port: `artists_ids[0]` (`:350`,
`IndexError`), `recording["aliases"]` (`:589`, `KeyError`), `actual[k]`
in the merge (`:220`, `KeyError`), the fifteen direct indexes in
`album_info`'s constructor call (`:662-671`), and `target_type.replace`
on an untyped relation (`_utils/musicbrainz.py:769`, `AttributeError`).

### 10. Every config key the translation layer reads

| key | shipped default | citation | verdict-moving |
|---|---|---|---|
| `match.ignored_media` | `[]` | `config_default.yaml:200`, read `:255-256` | yes (drops tracks; exact, case-sensitive membership, verified `'cd'` does not match `'CD'`) |
| `match.ignore_data_tracks` | `yes` | `config_default.yaml:201`, read `:259-260` | yes |
| `match.ignore_video_tracks` | `yes` | `config_default.yaml:202`, read `:263-264` | yes |
| `match.preferred.countries` | `[]` | `config_default.yaml:195`, read `:170` | yes (sets `country` and `date`) |
| `import.languages` | `[]` | `config_default.yaml:32`, read `:141` | yes once non-empty (section 12) |
| `import.ignored_alias_types` | `[]` | `config_default.yaml:58`, read `:136-138` | yes, conditionally on `languages` |
| `musicbrainz.aliases_as_credits` | `false` | `:287`, read `:267-268` | no (moves `artist_credit` only) |
| `va_name` | `"Various Artists"` | `config_default.yaml:106`, read `:675` | no (section 5) |
| `musicbrainz.genres` | `false` | `:277` | no |
| `musicbrainz.genres_tag` | `"genre"` | `:278`, read `:246-252` | no |
| `musicbrainz.external_ids.*` | all `false` | `:279-285`, read `:537` | no |
| `musicbrainz.host` / `https` / `ratelimit` / `ratelimit_interval` | `musicbrainz.org` / `no` / `1` / `1` | `_utils/musicbrainz.py:588-606` | transport only |

`match.ignored_media` and the two track-kind booleans are already in
the spec's enumerated tunable surface. `va_name` and
`musicbrainz.aliases_as_credits` are not, and both are read inside the
translation layer.

### 11. Lowercasing, stripping, and normalization

- `_preferred_alias` lowercases the **ignored alias types** (`:137`)
  and nothing else. Locale comparison is exact and case-sensitive
  (verified `['JA']` vs `['ja']`).
- `_parse_release_group` lowercases `primary_type` and every
  `secondary_type` (`:476`, `:484`).
- `_parse_label_infos` compares the label name against the exact
  literal `"[no label]"` (`:504`), case-sensitive, no strip.
- `_preferred_release_event` does no case folding at all.
- Search-query construction lowercases and strips: `format_search_term`
  at `_utils/musicbrainz.py:656` and the filter comprehension at
  `beetsplug/musicbrainz.py:752`. That is 3c's territory, not 3b's.
- `get_album_criteria` strips spaces from `catalognum` only (`:729-730`).
- **No Unicode normalization anywhere in the translation layer.** No
  `unicodedata.normalize` call exists in either file. This matches the
  assertion 3a was asked to verify for the distance path and extends
  it to translation: the port must not helpfully normalize.
- `_normalize_data` performs no case change, only the hyphen-to-underscore
  rewrite.

### 12. Behavior that depends on the loaded metadata source set

- `data_source` is `cls.__name__.replace("Plugin","")`
  (`$SP/beets/metadata_plugins.py:182-188`), so `MusicBrainzPlugin`
  yields the literal `"MusicBrainz"`, verified.
- `get_penalty` (`:160-169`) scans `find_metadata_source_plugins()` for
  a plugin whose `data_source` matches, falling back to the class
  constant `0.5` (`:180`). Both `get_penalty` and
  `find_metadata_source_plugins` are `@cache`d for the process lifetime
  (`:46`, `:159`).
- `add_data_source` fires only when `before != after` **and**
  (`before` is truthy **or** more than one source plugin is loaded)
  (`distance.py:357-361`). Under the corpus's single-plugin profiles
  the second disjunct is false, so the penalty depends entirely on the
  item's own `data_source`.
- `plugins.send("mb_album_extract")` (`:691-693`) and
  `send("mb_track_extract")` (`:465-467`) let any loaded plugin
  overwrite arbitrary fields. `mb_album_extract` runs **after** the
  release-event assignment, so a plugin can overwrite `country` and
  `year`. Under `plugins: musicbrainz` both send lists are empty.
- 3a's carried note stands and is refined: `dataSourcePenalty` falling
  back to `Tunables.DataSourceMismatchPenalty` rather than the class
  literal `0.5` diverges only when `musicbrainz.source_weight` or
  `musicbrainz.data_source_mismatch_penalty` is set, and no committed
  profile sets either.

### 13. Corpus evidence and the coverage picture

All figures produced offline, no network, no `make derive`, no
response body read into context. Probe scripts are in the session
scratchpad at
`/tmp/claude-1000/-var-home-glw907-Projects-dubplate/3215671c-d45a-4c80-9069-7df94d37964e/scratchpad/`:
`probe_translate.py`, `sweep_none.py`, `alias_ab.py`, `pseudo_sweep.py`,
`probe_mechanics.py`, `probe_media.py`, `probe_norm.py`,
`probe_credits.py`.

- 215 committed cases, 303 release-lookup responses, 6,919 tracks.
  Every one builds through `album_info` cleanly.
- Profile split: `estate` 178, `stock` 36, `locale` 1
  (`corpus/profiles/locale.yaml` is the only profile setting
  `import.languages`, to `['ja']`, and its single case is
  `real-nonlatin-utada-first-love-locale`).
- **The alias path is the highest-leverage part of the layer and is
  graded by one case.** A/B over all 303 releases with `languages=[]`
  against `languages=['ja']`: the album `artist` changes on **90**
  releases, `artist_sort` on 97, `release_group_title` on 2, and track
  `artist` on **3,272** tracks across 110 releases. `album` and track
  `title` change on none. Under the corpus as committed the path is
  inert on 302 of 303 releases, so tier 1 sees almost none of it.
- Four pseudo-release cases, three distinct pseudo releases, every
  linked actual captured. The merge moves only `media`, `country`,
  `label`, `catalognum`, `barcode`, `asin` on committed data; `year`
  and `mediums` never move.
- Ten releases produce `media == "Media"`, five of them with zero
  tracks.
- **The browse fallback has 30 tracks of headroom.** `BROWSE_MAXTRACKS`
  is 500 (`:67`) and the largest committed release is
  `real-boxset-rachmaninov-complete-stock` at 470 tracks. A slightly
  larger box set added to the corpus would trip
  `_ensure_complete_recordings` (`:602-629`), which issues paginated
  browse requests no committed cassette carries and which the stub
  would answer with recorded 404 misses.

### 14. Where this section corrects or extends the port map

The port map
(`docs/research/2026-09-02-beets-matcher-mechanics.md`,
Q5) is accurate on every line number checked. Three things to fix or
add.

1. **Basename ambiguity.** Q5 cites two different files as bare
   `musicbrainz.py`. Citations at `:50-67` (RELEASE_INCLUDES), `:551-561`,
   `:586-613`, `:634-637`, `:665-685`, `:687-697` and `:724-775` are
   `_utils/musicbrainz.py`; citations at `:58-67` (BROWSE_*), `:203-240`,
   `:433-469`, `:555-600`, `:602-629`, `:631-695`, `:714-734`, `:749`
   and `:777-781` are `beetsplug/musicbrainz.py`. The pair `:50-67` and
   `:58-67` resolve in different files, which is the one that will
   actually mislead a reader.
2. **`_flatten_artist_credit` does not exist** at the pinned version.
   The function is `MusicBrainzPlugin._parse_artist_credits`
   (`:301-357`).
3. **Two mechanics Q5 does not mention** and that move verdicts:
   `mediums` counting unfiltered media while track extraction filters
   them, and the literal `"Media"` sentinel for zero-track or
   mixed-media releases.

**Corpus and box safety.** No network call, no `make derive`, no write
to `corpus/`, no `responses/*.json` body read into context. Reads of
the venv and short Python probes only.

---

## (b) Harness and corpus as landed

Read off the tree at `HEAD f40a669`, main tree. One warm uncommitted
change exists and it is the conductor's own: `docs/STATUS.md` carries
an unstaged lane-registry edit recording this plan-prep workflow.
Nothing else is dirty. `corpus/staging/` holds five untracked
abandoned staging directories; that path is gitignored, so none is a
committed fixture.

### 1. Package map, and who owns what 3b touches

| Package | Files | Build tag | 3b's relation |
| --- | --- | --- | --- |
| `internal/harness/oracle` | 13 Go, 4 tracked Python | every Go file carries `//go:build oracle` except `reconcile.go` | 3b adds the fourth Python entry point and its Go driver here |
| `internal/harness/analysis` | 16 Go | none, gate-resident | 3b adds the `defect_free` block and its checker here |
| `internal/harness/cassette` | 2 Go | none | 3b reads only; the `Extra` schema half is 3c's |
| `internal/harness/stub` | 2 Go | none | 3b reads only |
| `internal/harness/conform` | 8 Go | none | born in 3a |
| `internal/harness/synth`, `mbrequest` | 7 + 2 Go | none | untouched by 3b |
| `internal/harness/corpus_test.go` | 1 | none | the gate-resident whole-corpus driver that wires every `analysis` checker to the real corpus (`internal/harness/corpus_test.go:1-7`) |

### 2. `internal/harness/oracle`: the entry-point shape 3b must copy

Four beets entry points exist. `derive_oracle.py` is the full one
(stub, rendered profile, `tag_album`). Three are narrow.

**`derive_component.py`** (460 lines). No stub, no `tag_album`, no
response, no rendered profile (`internal/harness/oracle/derive_component.py:1-16`).
It sets an empty scratch `BEETSDIR` so the run sits on beets' shipped
defaults (`derive_component.py:443-447`), redirects beets' own stdout
to stderr, and writes one `stable_json` document
(`derive_component.py:448-451`). Its Go driver is one call and nothing
else: `DeriveComponentTable` at `internal/harness/oracle/component.go:15`,
delegating to `runPython`.

**`derive_file_tags.py`** (205 lines) is the newest and the closest
template for 3b's `AlbumInfo` ledger. It takes one argument, a JSON
array of subjects (`derive_file_tags.py:183-201`); imports `build_items`
and `stable_json` from `derive_oracle` rather than re-implementing them
(`derive_file_tags.py:45`); asserts `BEETSDIR` isolation at the exact
moment beets is first imported and exits loudly if the runner's own
config leaked in (`derive_file_tags.py:74-81`); computes its writable
and read-only field sets off the live venv rather than hardcoding them
(`derive_file_tags.py:56-89`); and refuses any value that is not
already a JSON primitive so a `str(value)` fallback cannot silently
ship (`derive_file_tags.py:92-109`). Its Go driver, `DeriveFileTags` at
`internal/harness/oracle/file_tags.go:34`, marshals the subject list
into a scratch `input.json` and hands the path to `runPython`.

**`derive_ties.py`** (146 lines) is the one narrow entry point that
does need the stub and a rendered profile, because its raw material is
a real candidate's `TrackInfo` list. `DeriveTies` at
`internal/harness/oracle/ties.go:43` starts `stub.New`, renders the
profile against the stub's ephemeral address, writes a `tieInput`
envelope (`ties.go:27-33`), and decodes the result straight into
`analysis.TieEntry`. It returns a `TieSkipped` row without ever
starting the stub when the record carries no candidate
(`ties.go:48-54`).

**`runPython`** (`internal/harness/oracle/ties.go:120`) is the shared
launcher for the three narrow scripts. It sets `PYTHONHASHSEED=0` and
returns stdout verbatim. `derive_oracle.py` and `derive_ties.py` both
self-check that seed and exit if it is unset (`derive_oracle.py:286-287`,
`derive_ties.py:136-137`); `derive_component.py` and
`derive_file_tags.py` do not, because neither touches a set
difference.

**`Options`** (`internal/harness/oracle/oracle.go:29-33`) carries
`Venv`, `Shim`, and `Dir`. `venvOptions`
(`internal/harness/oracle/derive_test.go:22`) resolves
`tools/venv/.venv`, honoring `DUBPLATE_ORACLE_VENV`, and fails the test
with a pointer to `tools/venv/README.md` when no interpreter is there.

#### The oracle-tagged re-derive test, as shaped

Every derived ledger has the same three-part test.
`TestComponentTableReDerivesToItsCommittedTable`
(`internal/harness/oracle/component_derive_test.go:22`) and
`TestFileTagsLedgerMatchesTheCorpus`
(`internal/harness/oracle/file_tags_derive_test.go:65`) are the two to
copy:

1. `venvOptions(t)`, then re-derive.
2. Run the gate-resident checker over the *fresh* derivation first,
   failing with "the freshly derived ledger does not satisfy its own
   schema" before any diff runs (`file_tags_derive_test.go:74-76`).
3. Honor `-update-ledgers` by rewriting the committed file and
   returning (`file_tags_derive_test.go:78-88`).
4. Otherwise diff. **The comparison is decoded values, never bytes.**
   Both tests say why in the same comment: the committed file is
   indented for review while the shim prints one line, and both sides
   decode to the same doubles either way
   (`component_derive_test.go:58-63`, `file_tags_derive_test.go:176-181`).
   `decodeAny` (`component_derive_test.go:66`) is the shared helper.

The byte-identity invariant lives one level down and is separate.
`TestShimStdoutIsByteIdentical`
(`internal/harness/oracle/derive_test.go:119`) runs `deriveRaw` twice
against one fixture case and compares raw stdout with `bytes.Equal`,
then asserts the output is non-empty so the comparison cannot pass
vacuously. `corpus_derive_test.go:21-37` states plainly why the
whole-corpus test cannot do that: a case directory stores only the
decoded record in `case.json`, never the shim's original bytes, so
there is nothing at the raw-stdout level to diff a re-derivation
against.

The teeth for the decoded-value diff are a separate test:
`TestFileTagsLedgerReDerivationCatchesAPerturbation`
(`file_tags_derive_test.go:119`) perturbs one `values["title"]` in a
copy of the committed bytes and asserts the diff fires. **3b's ledger
owes the same perturbation test.**

`updateLedgers` is one package-level flag
(`internal/harness/oracle/ties_derive_test.go:26`), shared by all
three ledgers. The regeneration commands are recorded at
`corpus/README.md:2214-2218`.

`Reconcile` (`internal/harness/oracle/reconcile.go:36`) probes the live
venv for six components. The venv reconciles exactly against
`corpus/manifest.json` today: beets 2.13.1, unidecode 1.4.0, jellyfish
1.2.1, lap 0.5.13, numpy 2.5.2, mediafile 0.17.0.

### 3. `internal/harness/analysis`, in full

`readJSON` (`internal/harness/analysis/ledger.go:14`) is the one
loader every ledger goes through, and it sets
`DisallowUnknownFields` (`ledger.go:20`). An invented key is the
failure the package exists to catch. **A `defect_free` block added to
`AdjudicationEntry` in Go and not simultaneously in the committed
JSON, or the reverse, reds the gate immediately.**

**`AdjudicationEntry`** (`adjudication.go:44-55`) carries `ID`,
`Subject`, `Suspected`, `Citation`, `ReleaseTruth`, `Disposition`,
`CorrectedVerdict` (`adjudication.go:51`, `omitempty`), `Reason`,
`Adjudicator`, `Date`. `CorrectedVerdict` is a bare `string`, and every
committed row uses free prose, so the ten string-heuristic figures
live inside sentences on both `suspected` and `corrected_verdict`
today.

**`AdjudicationSubject`** (`adjudication.go:33-37`) is a three-field
union: `case_id`, or `album` plus `release_mbid`, never both and never
neither, enforced at `adjudication.go:103-116`.

**`checkAdjudicationEntry`** (`adjudication.go:100`) reports every
failure rather than the first. It requires `suspected`,
`release_truth`, `reason`, `adjudicator`, `date`; validates the
disposition against the four constants (`adjudication.go:15-20`);
requires `corrected_verdict` whenever the disposition is
`oracle-wrong` (`adjudication.go:143-145`); and enforces the sentinel
rule at `adjudication.go:140-142`.

**`AgentAdjudicator`** (`adjudication.go:28`) is the literal `"agent"`.
The doc comment states the reason plainly: Go has no other way to
recognize "a human name". The identical rule is duplicated for oracle
bugs at `oraclebug.go:125-127`.

**`OracleBugEntry`** (`oraclebug.go:21-33`) is one row per hypothesis,
not per case. It carries `ProbeCaseIDs`, `Reproduced` (agent-settable,
`oraclebug.go:8-15`), `ReproducedReason`, and `AdjudicationRefs`.
`CheckOracleBugs` (`oraclebug.go:55`) takes `adjudicationIDs
map[string]bool`, so today it can resolve a *row id* and cannot
resolve a *subject case id*. That signature is exactly what the spec's
B3 fold says pass 3d must change.

**`LoadComponentTable`** (`component.go:64`) and `CheckComponentTable`
(`component.go:74`) enforce `ComponentRowFloor = 72` (`component.go:15`),
row-id uniqueness, non-empty `args` and `output`, and a bidirectional
method-coverage check against the unexported `componentMethods`
(`component.go:23-35`). `ComponentRow.Args` and `.Output` stay
`json.RawMessage` by design (`component.go:52-61`).

**The file_tags loader is the newest ledger precedent and the closest
model for 3b.** `FileTagLedger` (`file_tags.go:64-74`) carries a shared
header (`shim_version`, `beets_version`, two field-name lists, a
shared `template` block) plus per-subject `Rows`. `FileTagSubjects`
(`file_tags.go:39`) is the single home for the twelve-subject list, and
its doc comment records the conductor's ruled substitution rather than
leaving it as an implementer's own choice. `CheckFileTagLedger`
(`file_tags.go:100`) does five things worth copying: it resolves every
row into a committed case, asserts the row set equals the named
subject set in both directions, checks every value key against the
ledger's own header rather than a second hand-maintained copy, rejects
a read-only key appearing on a row, and enforces `FileTagLedgerRowFloor
= 12` (`file_tags.go:17`).

**`TieEntry`** (`ties.go:56-70`) has no `Adjudicator` field, confirming
M4 as written. `checkTieEntry` (`ties.go:114`) accepts `excluded` and
`pinned` interchangeably given any non-empty reason (`ties.go:135-139`).

`ProximityEpsilon = 1e-6` (`proximity.go:20`); `TieMargin = 1e-9`
(`ties.go:15`). `Claims` (`claims.go:42-63`) and `CheckClaims`
(`claims.go:86`) distinguish a key added at 0.0 from a key never
added, through `PenaltyKeysZero` versus `PenaltyKeysAbsent`.

### 4. `internal/harness/cassette` and `stub`

`Case` (`cassette.go:41-49`) is `{id, input, profile, eval_date,
requests, extra?, expected?}`. `Expected` (`cassette.go:69-81`) is
eleven fields. `Candidate` (`cassette.go:96-102`) is `{source,
album_id, distance, penalties, tracks}` and carries no display strings
at all, which is exactly ruling E's premise.

`LoadCase` (`cassette.go:122`) and `validate` (`cassette.go:141`)
require a non-empty id, profile, and eval_date; reject a
non-canonical `path_query` (`cassette.go:157-159`); prove every
referenced response file is readable without decoding it
(`cassette.go:184-189`); and reject an `Extra` entry naming a file a
`Request` already consumes (`cassette.go:169-176`).

`Manifest` (`cassette.go:105-115`) has eight fields. `LoadManifest`
(`cassette.go:236`) checks shape only. **`wantManifestSchemaVersion =
"3"` lives at `internal/harness/corpus_test.go:160`, and it is the only
copy in the tree.**

`stub.New` (`internal/harness/stub/stub.go:54`) loads only
`c.Requests` into its replay map (`stub.go:55-62`); anything else
records a `Miss` and answers 404 (`stub.go:81-91`). The spec's B2
reading is correct as landed.

### 5. `internal/harness/conform` as 3a left it

`doc.go:1-23` states the package's charter and the harness-to-engine
direction. Five source files, two of them tests.

`component.go` (910 lines) is tier 0. `ReplayTable` (`component.go:80`)
returns a `ReplaySummary` (`component.go:60-72`) with `CountByMethod`,
`NearMisses`, and `TotalRows`, and a failed row is excluded from the
tally so a silently swallowed row cannot inflate it. `ReplayRow`
(`component.go:100`) dispatches on `method` across eleven methods and,
inside `add`, on an add-key-versus-weights shape probe
(`component.go:143-169`), giving twelve shapes. It recovers from a
panic into a row failure (`component.go:101-105`). `componentMethods`
(`component.go:24-35`) is a deliberate second copy of `analysis`'s
unexported list; the doc comment says two independently maintained
lists stating the same fact is the point.

Three decoder facts 3b or a later pass will lean on. A row's `weights`
map is an override layered onto `DefaultTunables().DistanceWeights`,
never a complete table (`component.go:399-407`, `tunablesWithOverrides`
at `component.go:417`). The `get_penalty` stub is wired through
`metadataSourceSet` (`component.go:562-568`). A compiled-regex option
decodes from `{"regex","ignorecase"}` into a `matcher.PatternOption`,
and an object lacking `regex` is a decode failure rather than a silent
compile of the empty pattern (`decodeOption`, `component.go:363-386`).

`compareFloats` (`component.go:817`) is the bit-identity-first
comparison. Equality passes clean; otherwise it admits 1e-12 relative
and reports a near miss rather than a silent pass.

`defaults.go` (76 lines) is the M1 cross-check. `CheckCaseWeights`
(`defaults.go:21`) compares key by key in both directions.
`CheckCaseThresholds` (`defaults.go:43`) covers the three thresholds
plus `max_rec`. `CheckUnidecodeVersion` (`defaults.go:67`) compares
`corpus/manifest.json`'s pin against `matcher.UnidecodeVersion`.

`arch_test.go` (182 lines) holds both architecture checks and both
fixtures that prove they fire. `checkNoHarnessImports` (`arch_test.go:30`)
walks `internal/engine` in `parser.ImportsOnly` mode; `checkNoTimeNow`
(`arch_test.go:72`) walks `internal/engine/matcher`. Both skip any
`testdata` subtree entirely (`arch_test.go:37-39`, `arch_test.go:79-81`),
which is the Task 6b hardening.

Tier 0 is green today: `corpus/analysis/components.json` carries
exactly 72 rows at beets 2.13.1, shim version 1, distributed `add` 8,
`add_data_source` 4, `add_equality` 3, `add_expr` 2, `add_number` 4,
`add_priority` 4, `add_ratio` 4, `add_string` 3, `generic_penalty_keys`
1, `string_dist` 23, `track_distance` 16.

### 6. Gate, Makefile, venv

`Makefile:38-39` defines `derive` as `go test -tags oracle -timeout 90m
./internal/harness/... ./cmd/harness/...`. The 90m timeout is explicit
because the oracle package alone already runs about 585s. `check` is
`vet lint test tidy-check` over `./...` (`Makefile:27`); `derive` sits
deliberately outside it (`Makefile:29-37`).

`scripts/check.sh` runs seven legs. Five hard-require their tool and
exit 1 when absent (`check.sh:19-38`); only docker (`check.sh:117`) and
systemd-analyze (`check.sh:136`) skip. The ruff leg runs `ruff check`
and `ruff format --check` over `git ls-files '*.py'` (`check.sh:85-92`),
and fails the gate outright if the list is empty (`check.sh:93-95`).
Six tracked Python files exist today, so **3b's new shim is linted the
moment it is tracked, with no gate edit.** `ruff.toml` selects `E,F,D`
with the presence family off and `D401` restored. Gitleaks scans the
tracked worktree with `--no-git` (`check.sh:107`), and
`.gitleaksignore` carries the single line-pinned finding
`internal/config/keys_test.go:generic-api-key:66`.

`tools/venv/README.md:1-8` records that the gate never needs the venv.
`tools/venv/README.md:24-36` states the pin doctrine: unidecode,
jellyfish, and lap sit directly in the distance path, and a version
change is a corpus-revalidation event.

### 7. Corpus figures, off the tree

| Figure | Value | Method |
| --- | --- | --- |
| Committed cases | **215** | `ls corpus/cases \| wc -l` |
| Tracked files under `corpus/cases` | 781 | `git ls-files` |
| Response files | 351 | `find` |
| Response bytes | 81,342,331 (81.3 MB), matching `corpus/README.md:1563` exactly | `find -printf %s` |
| `case.json` bytes | 25,266,959 (24.1 MiB) | same |
| Total logged requests | 351, all `GET` | `jq` over every `case.json` |
| Request path shapes | 303 `/ws/2/release/<mbid>`, 48 `/ws/2/release?` searches | same |
| Candidates across all records | 294 | `jq` |
| Candidates carrying a key at 0.0 | **272**, matching the spec's port-hazard 4 figure exactly | script |
| Cases carrying an `Extra` entry | **0** | `jq` |
| `misses.json` present anywhere | **none** | `find` |

**Strata** (`corpus/README.md:1555-1560`, confirmed against the tree):
7 smoke, 61 structural, 94 real-catalog, 17 adversarial (6
`adv-ft-*`, 4 `adv-remix-*`, 6 `adv-assign-*`, plus
`real-nonlatin-utada-first-love-locale`), 36 stock-config siblings.
The stock siblings are all named `real-*-stock`.

**Profiles**: 178 `estate`, 36 `stock`, 1 `locale`.

**Pseudo-releases: three case directories.** `non-latin-pseudo-release`
(smoke, Perfume *GAME*, `corpus/README.md:68`),
`real-nonlatin-zemfira` (`corpus/README.md:987`), and its stock twin
`real-nonlatin-zemfira-stock`. Each carries two release-lookup requests
and two response files, the second being the backward
`transl-tracklisting` target. Twenty-two other cases carry two or more
release lookups, and every one of those is search-fallback capture
rather than a pseudo-release follow-up.

**`preferred.countries`: zero cases, and the reason is stronger than an
empty config.** No committed profile sets any `match:` key at all
(`corpus/profiles/*.yaml`), so `preferred_countries` is empty
everywhere and `_preferred_release_event`'s loop never runs
(`beetsplug/musicbrainz.py:170-173`). Beyond that, measured
mechanically over all 351 responses, **not one release event anywhere
in the corpus carries a non-null `area` with `iso_3166_1_codes`.** The
preferred-country branch is therefore unreachable over this corpus
even under a hypothetical fourth profile that set the list. Every case
takes the fallback at `musicbrainz.py:175`.

Eight cases do carry a release with two or more `release-events`, so
the fallback-versus-events distinction is gradeable:
`real-boxset-mahler-complete` (and its stock twin),
`real-boxset-now-jukebox-classics` (and its stock twin),
`real-compilation-queen-greatest-hits-ii`,
`real-genre-metallica-master-of-puppets`, `should-reject`,
`various-artists`. In four of them the second event's date differs
from the top-level `date` the fallback actually returns, so a port
that read `release_events[0]` or merged events would diverge on those.

**Non-empty `languages`: one case.** `corpus/profiles/locale.yaml:20-22`
is the only profile setting `import.languages`, and
`real-nonlatin-utada-first-love-locale` is its only consumer.
**`_preferred_alias` (`musicbrainz.py:128-155`) returns `None`
unconditionally on the other 214 cases.** Two consequences for the
translation layer. First, `_key_with_preferred_alias`
(`musicbrainz.py:158-162`) always falls back to `obj[key]` on those
214. Second, and easier to miss, the track-title override at
`musicbrainz.py:589-590` is guarded by `not
_preferred_alias(recording["aliases"])`, so on 214 of 215 cases the
guard is unconditionally true and `recording["title"] =
track["title"]` fires whenever the track carries a title.

**Cases that would exercise aliases: 138 of 215 carry at least one
non-empty `aliases` array in their responses.** They matter through
`_parse_artist_credits` (`musicbrainz.py:301-331`), which is not gated
by `languages` at all: line 329 computes `alias =
_preferred_alias(el["artist"].get("aliases", []))` and line 330 sets
`artist_object = alias or el["artist"]`, so under empty `languages`
the applied artist is always the MusicBrainz artist entity's canonical
name rather than `el["name"]`, this release's own credited display.
That is the substitution `bug-nonlatin-alias-selection` now names as
ruled `oracle-wrong`. The four probe cases are
`real-nonlatin-alla-pugacheva-vol1`, `real-nonlatin-x-japan-art-of-life`,
`real-nonlatin-girls-generation`, `real-nonlatin-anna-vissi-essential`
(`corpus/README.md:1281-1286`).

**Video tracks: exactly two cases, measured mechanically.**
`real-nonlatin-hamasaki-ayumi` carries 138 recordings marked `video:
true` across its responses, and `real-nonlatin-girls-generation`
carries 3. The hamasaki case is load-bearing for 3b and is documented
at `corpus/README.md:814-853`: beets' shipped `ignore_video_tracks:
yes` drops every video track at `musicbrainz.py:577-578`, which
empties `AlbumInfo.tracks` for a video-medium release, and
`_add_candidate`'s zero-track filter then discards the candidate
before any distance runs. The case's own captured source can never
appear as a candidate under either shipped profile, and it lands
`low`/`skip` on an unrelated release.

**Data tracks: zero cases.** No response anywhere carries a non-empty
`data-tracks` array. The `ignore_data_tracks` branch
(`musicbrainz.py:562-563`) and the `"[data track]"` title filter
(`musicbrainz.py:573`) both have no corpus fixture.

**Pregap: zero cases.** No response anywhere carries a non-null
`pregap`. The pregap-inclusion branch (`musicbrainz.py:557-558`) has no
corpus fixture either.

Pregap, data tracks, and the preferred-country branch are therefore a
fourth, fifth, and sixth corpus-coverage gap of exactly the shape the
spec's M8 fold already names for `match.required`, `match.ignored`,
and `preferred.original_year`. **3b should declare them the same way
3a declared its three**, with hand-authored `testdata` fixtures in the
owning package and a named mutation per checker, rather than discover
them at execution time.

### 8. `corpus/analysis/adjudications.json`: every row and disposition

Twenty-one rows. **Eighteen `oracle-wrong`, three `oracle-correct`,
zero pending.** Every row's `adjudicator` is `Geoff` and every `date`
is `2026-09-04`. This matches 3a's HISTORY entry and `docs/STATUS.md`
exactly.

| Row id | Subject case | Disposition | `corrected_verdict` |
| --- | --- | --- | --- |
| `adj-adv-ft-draft-colon` | `adv-ft-draft-colon` | oracle-wrong | medium, skip; defect-free distance 0.101852 |
| `adj-adv-ft-left-colon` | `adv-ft-left-colon` | oracle-wrong | medium, skip; defect-free distance 0.117647 |
| `adj-adv-ft-gift-space` | `adv-ft-gift-space` | oracle-wrong | medium, skip; defect-free distance 0.107843 |
| `adj-adv-ft-soft-onesided` | `adv-ft-soft-onesided` | oracle-wrong | medium, skip; defect-free distance 0.086667 |
| `adj-adv-ft-craft-titles` | `adv-ft-craft-titles` | oracle-wrong | medium, skip; defect-free distance 0.091708 |
| `adj-adv-ft-theft-artist` | `adv-ft-theft-artist` | oracle-wrong | medium, skip; defect-free distance 0.090909 |
| `adj-adv-remix-both-sides` | `adv-remix-both-sides` | oracle-wrong | medium, skip; defect-free distance 0.080000 |
| `adj-adv-remix-onesided` | `adv-remix-onesided` | oracle-wrong | medium, skip; defect-free distance 0.071429 |
| `adj-adv-remix-bracket` | `adv-remix-bracket` | oracle-wrong | medium, skip; defect-free distance 0.092593 |
| `adj-adv-remix-track-titles` | `adv-remix-track-titles` | oracle-wrong | medium, skip; defect-free distance 0.081871 |
| `adj-adv-assign-swap-session-tape` | `adv-assign-swap-session-tape` | oracle-wrong | identity `{0:0,1:1,2:2}`; would-be distance 0.052222, medium, skip |
| `adj-adv-assign-swap-field-recording` | `adv-assign-swap-field-recording` | oracle-wrong | identity `{0:0,1:1,2:2,3:3}`; would-be 0.048000, medium, skip |
| `adj-adv-assign-cycle-rehearsal-reel` | `adv-assign-cycle-rehearsal-reel` | oracle-wrong | identity 5-item; would-be 0.070390, medium, skip |
| `adj-adv-assign-double-swap-broadcast` | `adv-assign-double-swap-broadcast` | oracle-wrong | identity 6-item; would-be 0.032540, still strong, apply; the correction is the assignment and applied titles, not the recommendation |
| `adj-adv-assign-endpoint-swap-demo-pass` | `adv-assign-endpoint-swap-demo-pass` | oracle-wrong | identity 4-item; would-be 0.106667, medium, skip |
| `adj-adv-assign-tight-margin-archive-disc` | `adv-assign-tight-margin-archive-disc` | oracle-wrong | identity `{0:0,1:1,2:2}`; would-be 0.022593, still strong, apply |
| `adj-adv-alias-girls-generation-romanized` | `real-nonlatin-girls-generation` | oracle-wrong | artist and albumartist become 소녀시대, the release's own credited display |
| `adj-adv-alias-vissi-native-greek` | `real-nonlatin-anna-vissi-essential` | oracle-wrong | artist and albumartist become Anna Vissi, the release's own credited display |
| `adj-real-nonlatin-alla-pugacheva-vol1` | `real-nonlatin-alla-pugacheva-vol1` | oracle-correct | (none) |
| `adj-real-nonlatin-x-japan-art-of-life` | `real-nonlatin-x-japan-art-of-life` | oracle-correct | (none) |
| `adj-real-nonlatin-utada-first-love-locale` | `real-nonlatin-utada-first-love-locale` | oracle-correct | (none) |

**`corpus/analysis/oracle_bugs.json`: six hypotheses, four
`oracle-wrong`, two `oracle-correct`, zero pending.**

| Hypothesis | Disposition | Probe cases | `adjudication_refs` |
| --- | --- | --- | --- |
| `bug-ft-pattern` | oracle-wrong (Geoff, 2026-09-04) | 6 `adv-ft-*` | 6 |
| `bug-remix-parenthetical` | oracle-wrong | 4 `adv-remix-*` | 4 |
| `bug-nonlatin-alias-selection` | oracle-wrong | 4 `real-nonlatin-*` | 4 |
| `bug-assignment-rejection-gap` | oracle-wrong | 6 `adv-assign-*` | 6 |
| `bug-threshold-boundary` | oracle-correct (Geoff, 2026-09-03) | `rec-strong-medium-boundary` | null |
| `bug-locale-alias-ignored` | oracle-correct | `real-nonlatin-utada-first-love`, `real-nonlatin-utada-first-love-locale` | 1 |

**The spec's B3 gap is closed on the tree, and 3b should record that.**
Revision 2 read three probe cases under still-pending hypotheses as
having no adjudication row. As read now, every probe case of every
`oracle-wrong` hypothesis resolves to an adjudication row by
`subject.case_id`, all twenty of them. The one probe case still
lacking a row, `real-nonlatin-utada-first-love`, sits under
`bug-locale-alias-ignored`, which is ruled `oracle-correct`, so the
rule 3d proposes would not fire on it. 3d's `CheckOracleBugs` change
can therefore land with zero backfill and zero new owner touchpoint.

The precedence rule the spec states is live on real data:
`bug-nonlatin-alias-selection` is `oracle-wrong`, yet two of its four
probe cases carry their own `oracle-correct` rows, because for those
two releases the entity name and the release credit coincide.

### 9. The ten string-heuristic cases, with the figures their `suspected` prose carries verbatim

Every one of the ten `suspected` fields carries three figures in
prose: the discounted `string_dist` reading, the undiscounted "true"
reading, and the derived and defect-free record distances. Each also
names the `SD_PATTERNS` edit that produced the counterfactual. The
corpus README's own framing is at `corpus/README.md:1099-1112`, and
the two case tables are at `corpus/README.md:1164-1169` and
`1186-1189`.

**The six `adv-ft-*` rows.** All six name the same edit, phrased in the
`suspected` field as "a word-boundary-guarded pattern" against
`[\[\(]?(featuring|feat|ft)[\. :].+`, and all six carry the same
citation string, "742ebdf69 (SD_PATTERNS' featuring discount, no word
boundary, Sept 2010); 7b59604c5 (partial fix, beets 2.13.0); 19d091440
(symmetric word-boundary fix, still unreleased at pinned 2.13.1)".

| Case | `string_dist` reads | against a true | derived distance | defect-free distance |
| --- | --- | --- | --- | --- |
| `adv-ft-draft-colon` | 0.061111 | 0.611111 | 0.010185185, strong, apply | 0.101851852, medium, skip |
| `adv-ft-left-colon` | 0.070588 | 0.705882 | 0.011764706 | 0.117647059 |
| `adv-ft-gift-space` | 0.064706 | 0.647059 | 0.010784314 | 0.107843137 |
| `adv-ft-soft-onesided` | 0.180571 | 0.520000 | 0.030095238 | 0.086666667 |
| `adv-ft-craft-titles` | 0.050000 / 0.056250 / 0.058824 | 0.500000 / 0.562500 / 0.588235 | 0.009170752 | 0.091707516 |
| `adv-ft-theft-artist` | 0.054545 | 0.545455 | 0.009090909 | 0.090909091 |

Each of the six also names its sole nonzero penalty key: `album` for
the first four, `tracks` for craft-titles, `artist` for theft-artist.
Every one closes with the identical clause, "past strong_rec_thresh
(0.04) and inside the medium band, so the verdict moves from
strong/apply to medium/skip. The discount is what tips it."

**The four `adv-remix-*` rows.** All four cite `#5868`. Two distinct
edits are named across them, and the `suspected` prose distinguishes
them precisely.

| Case | Edit named in `suspected` | `string_dist` reads | against a true | derived | defect-free |
| --- | --- | --- | --- | --- | --- |
| `adv-remix-both-sides` | a full-weight parenthetical, against `\(.*?\)` at 0.3 | 0.144000 | 0.480000 | 0.024000000 | 0.080000000 |
| `adv-remix-onesided` | a full-weight parenthetical | 0.128571 | 0.428571 | 0.021428571 | 0.071428571 |
| `adv-remix-bracket` | a full-weight bracket, against `\[.*?\]` at 0.3 | 0.166667 | 0.555556 | 0.027777778 | 0.092592593 |
| `adv-remix-track-titles` | discounted at 0.3 per track, the `track_title` twin | 0.150000 / 0.142105 / 0.150000 | 0.500000 / 0.473684 / 0.500000 | 0.024561404 | 0.081871345 |

`corpus/README.md:1105-1108` states the mechanical form of the remix
edit: "the two 0.3-weight bracket patterns removed, which is
arithmetically the same as charging a parenthetical at full weight".
The README also records that the computation was checked both ways:
run with `SD_PATTERNS` untouched it reproduces every committed
record's penalty to the last digit (`corpus/README.md:1109-1112`).

**One table exists, and 3b needs a second.** 3a landed the seam at
`internal/engine/matcher/tunables.go:189-196`
(`Tunables.StringDistPatterns`), with `BeetsSDPatterns` at
`tunables.go:316-323` and `SDPatternsWordBoundary` at `tunables.go:333-340`.
The word-boundary variant covers the six `adv-ft-*` figures by
assignment. **No table covering the remix edit exists**, so the four
`adv-remix-*` figures need a third variant that drops or full-weights
`\(.*?\)` and `\[.*?\]`. `SDReplace` sits at `tunables.go:302-304`.

3a's own warning travels with all of this and belongs in 3b's plan
header: the word-boundary guard changes **zero of the 72 tier-0 rows**,
so tier 0 is blind to a "fixed" table, and Task 4's byte-identity test
against beets' six shipped literals is the only check that catches a
silent improvement to the default.

### 10. The exact shape of a committed case directory

Read with `ls` and `jq keys` only; no response body was read.

**Directory contents are uniform across all 215 cases**, exactly three
entries each: `case.json`, `claims.json`, `responses/`. No case
carries a `misses.json`, and no case carries anything else.

```
corpus/cases/<id>/
  case.json          # cassette.Case
  claims.json        # analysis.Claims
  responses/
    001.json         # verbatim wire body, zero-padded three-digit, sequential
    002.json
```

`case.json` top-level keys, in the order `cassette.Case`'s tags emit
them: `id`, `input`, `profile`, `eval_date`, `requests`, `expected`.
`extra` is absent from all 215 because it is `omitempty` and empty
everywhere.

- `input` has exactly one key, `items`, a list of flat `ItemFields`
  objects. `clean-match`'s item 0 keys are `album`, `albumartist`,
  `artist`, `length`, `mb_albumid`, `mb_trackid`, `title`, `track`,
  `year`.
- `requests[]` keys are `method`, `path_query`, `response`. All 351
  are `GET`. `response` is a directory-relative path such as
  `responses/001.json`, never a bare filename. A representative
  lookup key: `/ws/2/release/<mbid>?fmt=json&inc=artists%2Bmedia%2Brecordings%2Brelease-groups%2Blabels%2Bartist-credits%2Baliases%2Brecording-level-rels%2Bwork-rels%2Bwork-level-rels%2Bartist-rels%2Bisrcs%2Burl-rels%2Brelease-rels%2Bgenres%2Btags`.
  A representative search key:
  `/ws/2/release?fmt=json&limit=5&query=release%3A%28ok+computer%29+artist%3A%28radiohead%29`.
  Note both facts the canonicalization rule protects, `%2B` for the
  logical `+` joining `inc=` values and byte-wise sorted pairs
  (`corpus/README.md:1905-1942`).
- `expected` keys: `recommendation`, `action`, `echoed_date`,
  `config_hash`, `weights`, `thresholds`, `candidates`, `assignment`,
  `extra_items`, `extra_tracks`, `applied_fields`.
- `expected.candidates[]` keys: `source`, `album_id`, `distance`,
  `penalties`, `tracks`.
- `claims.json` keys are per family; `clean-match` carries
  `best_album_id`, `extra_items_len`, `extra_tracks_len`,
  `penalty_keys_zero`, `recommendation`.

**Response bodies are raw wire JSON, not beets-normalized.** `jq keys`
on `clean-match/responses/001.json` returns hyphenated keys
throughout: `artist-credit`, `cover-art-archive`, `label-info`,
`packaging-id`, `release-events`, `release-group`, `status-id`,
`text-representation`, plus a generic `relations` array. `media[0]`
keys are `format`, `format-id`, `id`, `position`, `title`,
`track-count`, `track-offset`, `tracks`, with no `pregap` and no
`data-tracks` key at all. **The Go translation layer therefore has to
perform the normalization itself**, matching `_normalize_data` at
`beetsplug/_utils/musicbrainz.py:746-775`, which regroups a generic
`relations` array by `target-type` into `<target_type>_relations` keys
with `target-type` stripped from each entry, and converts every other
key's hyphens to underscores recursively. The regroup uses
`sorted(v, key=get_target_type)` before `groupby`, and Python's sort is
stable, so within a target type the original array order survives.

**Committed ledger figures.** `proximity.json` carries two entries, one
`skipped` (`admit-zero-track-dropped`) and one `flagged` on
`strong_rec_thresh` (`rec-strong-medium-boundary`, distance
0.0400002). `ties.json` carries 215 rows: 212 measured unique, 2
`pinned` (`should-ask`, `should-reject`), 1 `skipped`. `file_tags.json`
carries 12 rows, 69 writable fields, 9 read-only fields.
`divergences.md` is deliberately empty of rows and says so
(`corpus/analysis/divergences.md:12-25`).

**The twenty echoed weight keys**, identical on all 215 cases: `album`
3, `album_id` 5, `albumdisambig` 0.5, `artist` 3, `catalognum` 0.5,
`country` 0.5, `data_source` 2, `label` 0.5, `media` 1, `medium` 1,
`mediums` 1, `missing_tracks` 0.9, `track_artist` 2, `track_id` 5,
`track_index` 1, `track_length` 2, `track_title` 3, `tracks` 2,
`unmatched_tracks` 0.6, `year` 1.

### 11. Staleness and defects found while reading

Three documentation defects and one empirical answer to a carried
question. All four are cheap for 3b to fold.

1. **`corpus/README.md` is stale on adjudication state, in three
   places.** `corpus/README.md:1096` says every adversarial case
   carries an `adjudications.json` row "at `pending-adjudication`";
   `corpus/README.md:1341` calls `bug-locale-alias-ignored`
   "pending-adjudication, owner-only"; `corpus/README.md:2317-2318`
   says "Every hypothesis but `bug-threshold-boundary` sits at
   `disposition: pending-adjudication`". All three were true at 2c's
   close and none is true now. The ledgers are the source of truth and
   they were ruled on 2026-09-04. Since 3b is the pass that rewrites
   the adjudication schema, correcting these three cells belongs in
   it, matching how 3a folded its own two `corpus/README.md`
   corrections.

2. **`corpus/README.md:2469-2471` misstates `locale.yaml`.** It says
   placeholder substitution is "a no-op on `stock.yaml` and
   `locale.yaml`, which carry neither". `corpus/profiles/locale.yaml:13-14`
   does carry both `{{DUBPLATE_LIBRARY_DIR}}` and
   `{{DUBPLATE_LIBRARY_DB}}`, and has to, because it is rebased on
   `estate.yaml`. Only `stock.yaml` carries neither. Harmless at
   runtime, wrong in the record a reviewer checks a profile change
   against.

3. **The remix defect-free table does not exist.** Covered in section
   9. This is a deliverable 3b needs before it can derive four of the
   ten `defect_free` figures, not a note.

4. **3a's carried absent-source-field divergence is latent, not
   live.** 3a asked what a beets `album_info.album` of `None` should
   become, since `addAlbumAlbum` passes `&likelies.Album`/`&info.Album`
   and scores 0.0 where beets scores `string_dist("", None) = 1.0`.
   Measured across all 294 committed candidates, the largest `album`
   penalty is 0.169, nowhere near the 3/61 a raw 1.0 would produce, so
   no committed case exercises the divergence today. The question is a
   design decision for the `AlbumInfo` vocabulary, and it can be
   settled on its merits without a gate failure forcing the answer. It
   does become live the moment the `AlbumInfo` ledger grades display
   strings, since a beets `TrackInfo.title` of `None` and a `""` are
   different values that 3a's plain-string types cannot tell apart.

Two further mechanical facts worth carrying into the plan header.
`internal/harness/analysis`'s loader sets `DisallowUnknownFields`, so
the Go type and the committed JSON must move in the same commit. And
`internal/config/keys_test.go:66` is pinned by line number in
`.gitleaksignore`, so any insertion above that line reds the gate for
every lane at once; 3b writes no `internal/config`, but the rider
applies to the tree it shares.

---

## (c) The engine as landed and the 3b Go surface

### 1. The engine as landed

`internal/engine/matcher` is one package, twenty-nine non-test files,
no subpackages. `internal/engine` holds `matcher` and nothing else:
there is no `musicbrainz` directory yet, so 3b creates it.

Exported surface, read off `go doc -all ./internal/engine/matcher`:
types `AlbumInfo`, `TrackInfo`, `ItemFields`, `Likelies`, `Tunables`,
`Preferred`, `AlbumPair`, `Distance`, `TrackDistance`, `PenaltyItem`,
`PatternOption`, `SDPattern`, `SDReplacement`; funcs `DefaultTunables`,
`ProjectItemFields`, `MostCommonTags`, `NewDistance`,
`DistanceForAlbum`, `DistanceForTrack`, `StringDist`, `Transliterate`,
`Levenshtein`, `CompilePatternOption`; the `Distance` methods `Add`,
`AddDataSource`, `AddEquality`, `AddExpr`, `AddNumber`, `AddPriority`,
`AddRatio`, `AddString`, `AddedKeys`, `PerKey`, `NonzeroItems`,
`NonzeroKeys`, `GenericPenaltyKeys`, `Penalties`, `RawDistance`,
`MaxDistance`, `Distance`, `Update`; vars `BeetsSDPatterns`,
`SDPatternsWordBoundary`, `SDEndWords`, `SDReplace`; const
`UnidecodeVersion`. Every component function (`addAlbum*`,
`addTrack*`) and both composition slices (`albumComponents`,
`trackComponents`) are unexported. 3b therefore builds `AlbumInfo`
values from outside the package and can reach nothing else.

#### `info.go` — the vocabulary 3b must populate

`AlbumInfo` (`/var/home/glw907/Projects/dubplate/internal/engine/matcher/info.go:10-67`),
fourteen fields, every one a plain Go value with no pointer and no
presence flag:

| Go field | Type | beets field (cited in the file) |
| --- | --- | --- |
| `Artist` | `string` | `album_info.artist`, Info base class |
| `Album` | `string` | `album_info.album`, Info base class |
| `VA` | `bool` | `album_info.va`, hooks.py:381 |
| `Media` | `string` | `album_info.media`, Info base class |
| `Mediums` | `int` | `album_info.mediums`, hooks.py:371 |
| `Year` | `int` | `album_info.year`, hooks.py:382 |
| `OriginalYear` | `int` | `album_info.original_year`, hooks.py:375 |
| `Country` | `string` | `album_info.country`, hooks.py:364 |
| `Label` | `string` | `album_info.label`, hooks.py:369 |
| `CatalogNum` | `string` | `album_info.catalognum`, hooks.py:363 |
| `AlbumDisambig` | `string` | `album_info.albumdisambig`, hooks.py:357 |
| `AlbumID` | `string` | `album_info.album_id`, hooks.py:356 |
| `DataSource` | `string` | `album_info.data_source`, Info base class |
| `Tracks` | `[]TrackInfo` | `album_info.tracks`, hooks.py:355 |

`TrackInfo` (`info.go:73-105`), eight fields: `Title string`, `Artist
string`, `Length float64`, `Index int`, `MediumIndex int`, `Medium
int`, `TrackID string`, `DataSource string`.

`ItemFields` (`info.go:124-143`) is the user-item side, eighteen
fields, built by `ProjectItemFields(raw map[string]any) ItemFields`
(`info.go:179-200`) with `stringField`/`intField`/`floatField` helpers
that return the type's zero for an absent or mistyped key. This is the
JSON-to-typed-default precedent 3b's ledger decoder should echo, and
`info.go:114-123` already records the one beets field (`data_source`)
whose real default is `None` rather than `""`, with the argument for
why `""` models it faithfully.

#### `tunables.go` — what the translation layer reads

`Tunables` (`/var/home/glw907/Projects/dubplate/internal/engine/matcher/tunables.go:42-197`)
with `DefaultTunables()` at `:219-276`. The translation-relevant
fields, each with the citation it carries:

- `Preferred.Countries []string` — `tunables.go:203`,
  `config_default.yaml:194-197`, default empty. Read by release-event
  selection **and** by the country distance component.
- `Preferred.Media []string` — `tunables.go:207`. Distance only.
- `Preferred.OriginalYear bool` — `tunables.go:214`. Distance only.
- `IgnoredMedia []string` — `tunables.go:96`, `config_default.yaml:200`,
  default empty. Translation only (medium filter).
- `IgnoreDataTracks bool` — `tunables.go:100`, `config_default.yaml:201`,
  default true.
- `IgnoreVideoTracks bool` — `tunables.go:104`, `config_default.yaml:202`,
  default true.
- `Languages []string` — `tunables.go:167`, `config_default.yaml:32`,
  default empty, with the doc comment recording that the empty default
  makes `_preferred_alias` structurally unreachable at launch (Ruling
  D).
- `IgnoredAliasTypes []string` — `tunables.go:172`, `config_default.yaml:58`,
  default empty.
- `SearchLimit int` — `tunables.go:124`, `metadata_plugins.py:201,415`,
  default 5. 3c's, not 3b's.
- `ExtraTags []string` — `tunables.go:154`, `beetsplug/musicbrainz.py:286`,
  default empty. 3c's query construction.
- `DataSourcePenalties map[string]float64` /
  `DataSourceMismatchPenalty float64` — `tunables.go:149`/`:136`. The
  translation supplies the `data_source` string those score against.

Two fields the translation needs and `Tunables` does **not** carry
today: `va_name` (`config_default.yaml:106`, `"Various Artists"`),
which `album_info` writes into `info.artist` on a VA release
(`beetsplug/musicbrainz.py:673-681`), and `musicbrainz.aliases_as_credits`
(default `False`, `beetsplug/musicbrainz.py:288`), which selects which
object supplies the credit name in `_parse_artist_credits`
(`musicbrainz.py:311-313`). Neither moves a distance — the artist
component is skipped when `VA` (`album_artist.go:13-15`), and
`artist_credit` is applied-field only — but both move
`AlbumInfo.Artist`, which the ruling-E ledger grades field by field.
3b must add at least `VAName` to `Tunables` or state in the checker
why `Artist` is not compared on a VA case. Neither is echoed by any
committed case, so the M1 defaults cross-check
(`conform.CheckCaseWeights`/`CheckCaseThresholds`,
`/var/home/glw907/Projects/dubplate/internal/harness/conform/defaults.go:21,41`)
cannot cover them; coverage is a unit test.

#### `likelies.go` — the item side of every album component

`Likelies` (`/var/home/glw907/Projects/dubplate/internal/engine/matcher/likelies.go:11-25`),
thirteen plurality fields: `Artist`, `Album`, `AlbumArtist`, `Year`,
`DiscTotal`, `MBAlbumID`, `Label`, `Barcode`, `CatalogNum`, `Country`,
`Media`, `AlbumDisambig`, `DataSource`. `MostCommonTags(items
[]ItemFields) (Likelies, map[string]bool)` at `:47-83`, with
`plurality` breaking ties by first-seen index (`:95-122`) and the
albumartist-consensus override at `:78-80`. Its doc comment
(`likelies.go:3-10`) is the precedent for the input-type question
below: it chose a typed struct over `map[string]any` because every
field already has a known Go type, and throwing that away buys
nothing.

#### The consumers of translated fields

- `album_year.go:24-53` reads `info.Year`, `info.OriginalYear`,
  `tun.Preferred.OriginalYear`, and `evalDate time.Time`. Three
  branches in order; branch 1 is the declared corpus-coverage gap.
- `album_country.go:20-29` reads `info.Country` and
  `tun.Preferred.Countries`, compiling each country as an anchored
  case-insensitive `PatternOption` via `compileCountryPatterns`
  (`:31-40`), which **panics** on a malformed pattern.
- `album_media.go:21-34` reads `info.Media` and `tun.Preferred.Media`,
  compiling each as `(\d+x)?(<pat>)` via `compileMediaPatterns`
  (`:36-46`), also panicking.
- `album_mediums.go:10-14` reads `info.Mediums`; `album_label.go`,
  `album_catalognum.go`, `album_disambig.go`, `album_id.go`,
  `album_data_source.go`, `album_extras.go`, `album_tracks.go` read
  the rest.

**A port hazard the tree does not yet name.** `Preferred.Countries`
feeds two different matching rules. `addAlbumCountry` compiles each
entry as an anchored regex (`album_country.go:34` →
`CompilePatternOption(pattern, true)`, `penalty.go:368-378`), while
`_preferred_release_event` does plain list membership, `country in
area["iso_3166_1_codes"]`
(`tools/venv/.venv/lib/python3.13/site-packages/beetsplug/musicbrainz.py:174`).
One config list, regex in the scorer, exact string in the translator.
A 3b implementer reusing `compileCountryPatterns` for release-event
selection would be silently wrong on any pattern-shaped entry.

#### The `*string` nil-as-None convention, and the pointer ruling

`StringDist(tun Tunables, str1, str2 *string) float64` (`strdist.go:28-65`):
both nil returns 0.0, one nil returns 1.0, `strdist.go:29-34`.
`AddString(key string, str1, str2 *string)` (`strdist.go:69-71`) is
one call into it. The doc comment at `strdist.go:13-17` states the
convention and that the `info.go` vocabulary types never produce a
nil, so a nil comes from a caller modelling an unset source field.

Every in-tree caller passes a typed default's address, never a literal
nil: `album_album.go:9` (`&likelies.Album`, `&info.Album`),
`album_artist.go:16`, `album_country.go:27`, `album_label.go:12`,
`album_catalognum.go:12`, `album_disambig.go:12`, `track_title.go:16`,
`track_artist.go:36`. That is 3a's pointer ruling, and
`track_title.go:12-14` records it inline. The one other place
`*string` appears is the tier-0 decoder's `callAddString`
(`/var/home/glw907/Projects/dubplate/internal/harness/conform/component.go:293-306`),
which unmarshals each argument into `*string` so a JSON `null` becomes
nil — the exact mechanism 3b's ledger decoder should reuse.

#### `internal/config` — shape, and what registry exists

Three files, 425 non-test lines. `Threshold{Key, Default, Doc}` plus
`Register`/`Thresholds`/`ParseInt`
(`/var/home/glw907/Projects/dubplate/internal/config/thresholds.go:13-59`)
is the tunable registry; `registeredThresholds` is empty by design
(`:19-22`, "Rung 1 registers none: the registry is the deliverable").
`ownedWall` is a slice literal of five `WallRow`s
(`/var/home/glw907/Projects/dubplate/internal/config/keys.go:19-25`),
`reservedClasses` at `:57`, `supersededKeys` at `:164`, and
`Load(environ []string)` at `config.go:159`. Everything is
string-valued env with an int parser. There is no typed variant and no
validation hook, which is why 3a's second carried item (regex
validation) is still open: `internal/config` today has nowhere to put
a compiled-pattern check. **3b must not touch `keys.go`** (the
four-knob registration is blocked on rung 1b; `docs/STATUS.md` lane
registry, and the `.gitleaksignore` line-number pin at
`keys_test.go:66`).

### 2. The proposed 3b Go surface

New package `internal/engine/musicbrainz`, two files in 3b
(`translate.go`, `pseudo.go`); `query.go`, `client.go`, `limiter.go`
are 3c's. Note that `TestEngineNeverImportsHarness`
(`/var/home/glw907/Projects/dubplate/internal/harness/conform/arch_test.go:138-144`)
already walks all of `internal/engine`, so the new package inherits
the invariant free. `TestMatcherNeverReadsTheClock` (`:146-152`) walks
only `internal/engine/matcher`; 3b should widen it to
`internal/engine`, because release-event date selection is exactly
where a `time.Now` would creep in.

#### 2.1 Input type: typed structs with hyphenated tags, plus one explicit regrouping pass

**Proposal: typed structs, not `map[string]any`.**

```go
// Release is the subset of a normalized WS/2 release the translation reads.
type Release struct {
    ID            string        `json:"id"`
    Title         string        `json:"title"`
    Status        string        `json:"status"`
    Disambiguation string       `json:"disambiguation"`
    Barcode       *string       `json:"barcode"`
    Country       *string       `json:"country"`
    Date          string        `json:"date"`
    Aliases       []Alias       `json:"aliases"`
    ArtistCredit  []ArtistCredit `json:"artist-credit"`
    ReleaseEvents []ReleaseEvent `json:"release-events"`
    LabelInfo     []LabelInfo   `json:"label-info"`
    Media         []Medium      `json:"media"`
    ReleaseGroup  ReleaseGroup  `json:"release-group"`
    Relations     []Relation    `json:"relations"`
}
```

Reasoning. The hyphen-to-underscore half of beets' normalizer
(`tools/venv/.venv/lib/python3.13/site-packages/beetsplug/_utils/musicbrainz.py:774`)
exists so a Python `dict` can be read with underscore keys. A Go
struct tag names the wire key directly, so that half is discharged by
the tags and needs no runtime pass. This is `Likelies`' own argument
(`likelies.go:3-10`) applied one layer out, and it follows
go-conventions' concrete-types bias.

The `relations` regrouping is genuinely structural, not a rename: one
array is split by `target-type` into `artist_relations`,
`work_relations`, `url_relations`, `release_relations` with the
discriminator dropped (`_utils/musicbrainz.py:759-772`). Its consumers
are `release.get("release_relations", [])` in the pseudo path
(`musicbrainz.py:788`), `_parse_external_ids(release.get("url_relations",
[]))` (`:671`), and the recording-level `work_relations`/`artist_relations`
(`:458-460`). So land it as one named exported function with its own
test:

```go
// GroupRelations splits rels by target type, mirroring the MusicBrainz
// client's own regrouping. The sort is stable so within-type order
// survives, which Python's sorted() gives for free.
func GroupRelations(rels []Relation) map[string][]Relation
```

Implement with `slices.SortStableFunc` on `TargetType` then a group
walk. Record in the doc comment that the produced key order is
unobservable (every beets consumer uses `.get`), so callers must not
depend on it, and that the hyphen half of the normalizer is carried by
the struct tags. Both halves of the spec's first translation bullet
then have a named home and a test.

Rejected alternative: decode to `map[string]any`, run a faithful
recursive normalizer, then re-decode. It reproduces a normalizer for
hundreds of keys nothing reads, loses every type, and makes the
decoder untestable except through goldens, which go-conventions calls
last-resort.

#### 2.2 Function signatures

```go
// TranslateAlbum builds the matcher's candidate vocabulary from one
// normalized WS/2 release, mirroring MusicBrainzPlugin.album_info.
func TranslateAlbum(rel *Release, tun matcher.Tunables) (*matcher.AlbumInfo, error)

// MergePseudoRelease folds an actual release's own fields into a
// pseudo-release candidate, mirroring _merge_pseudo_and_actual_album:
// the pseudo supplies identity and titles, the actual supplies the
// physical and date fields.
func MergePseudoRelease(pseudo, actual *matcher.AlbumInfo) *matcher.AlbumInfo
```

Reasoning on each choice.

- **`tun` by value, last, not stored.** `Tunables` is a plain value
  struct already passed this way by `DistanceForAlbum` (`score.go:60`)
  and `DistanceForTrack` (`track.go:45`). No config read, no
  package-level state, no options struct.
- **No `ctx`, no error-returning fetcher parameter.** 3b translates
  two already-decoded responses; fetching is 3c's. But see the open
  question on `_ensure_complete_recordings`.
- **`error`, not a panic.** The matcher's two existing pattern
  compilers panic (`album_media.go:41`, `album_country.go:36`) because
  a pattern table is code. A WS response is external input, so
  translation validates at the boundary and returns a wrapped error,
  per go-conventions ("validate at boundaries"). Sentinels only if a
  caller branches.
- **Merge returns a new value rather than mutating.**
  `_merge_pseudo_and_actual_album` starts from `pseudo.copy()`
  (`musicbrainz.py:218`), and the fifteen-key `from_actual` list at
  `:220-237` is the whole contract. Of those, seven are in
  `matcher.AlbumInfo`: `media`, `mediums`, `country`, `catalognum`,
  `year`, `original_year`, `label`. The other eight (`month`, `day`,
  `original_month`, `original_day`, `barcode`, `asin`, `style`,
  `genres`) are outside the vocabulary. `Album`, `Artist`, `VA`,
  `AlbumDisambig`, `AlbumID`, `DataSource`, and `Tracks` stay the
  pseudo's. Write that seven-field list in the doc comment; it is the
  merge's entire verdict-moving effect.
- **Pseudo detection stays out of the merge.** `status ==
  "Pseudo-Release"` and the `transl-tracklisting` + `direction ==
  "backward"` relation walk (`musicbrainz.py:782-797`) select *which*
  second release to fetch, which is 3c's client work. 3b exports the
  predicate so 3c calls it, and merges over two `AlbumInfo` values as
  the spec scopes it.

**Evaluation date.** It does not reach translation at all. `album_info`
never reads a clock; only `distance()` does, and 3a already made it an
explicit `evalDate time.Time` parameter (`score.go:60`, contract 9).
Adding it to `TranslateAlbum` would invite a clock read into a new
package. The widened `checkNoTimeNow` walk is the enforcement.

#### 2.3 The absent-source-field question (3a's carried album-None divergence)

**Recommendation: keep plain-string fields, translate `None` to the
zero value, and land a gate-resident proof that the divergence is
unreachable through the only source dubplate ships.**

The evidence, derived from pinned beets rather than from any document.
Only **three** `add_string` call sites on the whole distance path are
unguarded by a truthiness test on the info side, so only three can
ever see a `None`:

1. `dist.add_string("album", likelies["album"], album_info.album)` —
   `tools/venv/.venv/lib/python3.13/site-packages/beets/autotag/distance.py:451`,
   unguarded.
2. `dist.add_string("artist", likelies["artist"], album_info.artist)`
   — `distance.py:448`, guarded only by `not album_info.va`.
3. `dist.add_string("track_title", item.title, track_info.title)` —
   `distance.py:402`, unguarded.

Every other album component is guarded on both sides
(`distance.py:504-518` label, catalognum, albumdisambig; `:499-501`
country), and for those `None` and `""` are indistinguishable.
`album_info.media`, `.mediums`, `.year`, `.original_year`, `.country`
are all truthiness-guarded (`distance.py:454, 467, 471-481, 495-499`),
so `None`/`""`/`0` collapse. `add_equality("album_id", ...)` is guarded
by `if likelies["mb_albumid"]` (`distance.py:519`), so a `None`
candidate ID scores 1.0 either way. `add_data_source` already handles
the `None`-as-`""` case explicitly in the Go port (`penalty.go:338-347`)
and matches beets' `before != after and (before or len(plugins) > 1)`
(`distance.py:357-359`) under one loaded plugin.

Probe run under `tools/venv/.venv/bin/python`, confirming the
arithmetic: `string_dist("", None) = 1.0`, `string_dist("", "") = 0.0`,
`string_dist(None, None) = 0.0`; `AlbumInfo(album_id="x",
tracks=[]).album` is `None` (Info sets it explicitly,
`beets/autotag/hooks.py:212,226`) and `TrackInfo(track_id="y").title`
is `None` (`hooks.py:463,491`).

Now the reachability half. Through the MusicBrainz translation path
all three are always `str`:

- `album` = `_key_with_preferred_alias(release, key="title")` →
  `alias["name"] if alias else obj["title"]`
  (`beetsplug/musicbrainz.py:154-157`). `Release.title` and
  `Alias.name` are both non-optional `str` in the client's own typing
  (`_utils/musicbrainz.py:523`, `:140`).
- `artist` = `"".join(artist_parts)` (`musicbrainz.py:349`), a `str`
  by construction; and on a VA release it is overwritten with
  `va_name` (`:676`).
- `track_info.title` = `_key_with_preferred_alias(recording,
  key="title")` (`musicbrainz.py:443`), same shape; the track-level
  override only ever *replaces* it with a truthy `track["title"]`
  (`:585-586`).

So 3a's divergence is real arithmetic on an input the shipped source
cannot produce. Modelling it costs the whole vocabulary: `*string` on
`AlbumInfo.Album`, `.Artist`, and `TrackInfo.Title` would ripple into
all eight `AddString` call sites, every component test, and 3d's
assignment work, to represent a state nothing reaches.

The proof must be a check, not an assumption. Two pieces, both cheap:

- **The ledger decoder types every string field as `*string`.** A JSON
  `null` decodes to nil, exactly as `callAddString` already does
  (`conform/component.go:298-305`).
- **The checker carries a per-field nullability policy.** For `album`,
  `artist`, and each track's `title`, a nil is a gate failure naming
  the case and the field, because it is the only shape in which a real
  divergence could arrive. For every other field a nil compares equal
  to the Go zero value, because beets' own guard makes them
  indistinguishable, and the field's entry in the policy table says so
  in one line.

That second bullet is not hypothetical. `album_info.media` is
`medias.pop() if len(medias) == 1 else "Media"` over `{t.media for t
in track_infos}` (`musicbrainz.py:653-658`), and `t.media` is
`medium["format"]`, which the client types as `str | None`
(`_utils/musicbrainz.py:468`). A single-medium release with a null
format yields `AlbumInfo.media = None`, which the ledger will show as
`null` and the Go translation as `""`. The nullability policy is what
turns that from a false failure into a recorded equivalence.

Assertion discipline: a hand-authored ledger fixture in
`internal/harness/conform/testdata` with one row carrying `"album":
null` proves the strict half fires, and one carrying `"media": null`
proves the lenient half does not, per the spec's rule that these
fixtures are never corpus cases.

The residual, recorded rather than absorbed: the proof is source-scoped.
A second metadata source (or a beets bump that lets an alias carry a
null name) reopens it, and the strict-nil check is the thing that
would catch it.

Rejected alternatives. Pointer fields on `AlbumInfo`: correct but pays
the full vocabulary cost for an unreachable state, and turns eight
simple call sites into nil-juggling. A presence bitset: a Go idiom
nobody in this repo uses, invisible to `go doc`, and it would need its
own decoder. A sentinel string: it can collide with real data and
every reader must know it.

#### 2.4 The ledger's JSON to Go mapping

Follow 2d's `file_tags.json` template exactly
(`/var/home/glw907/Projects/dubplate/internal/harness/analysis/file_tags.go`):
a versioned ledger struct, a fixed subject list with a row floor, a
`Load*` reader, gate-resident schema and coverage checks, and an
`//go:build oracle` re-derive test. The three-package split 2d used is
the one to copy: schema and checkers in `internal/harness/analysis`,
the Python entry point and its Go driver in `internal/harness/oracle`,
the field-by-field comparison against the engine in
`internal/harness/conform` (which is where the harness-to-engine
direction is legal, per `conform/defaults.go:61-67`).

```go
// internal/harness/analysis
const AlbumInfoLedgerRowFloor = /* the subject count, read off the tree */

type AlbumInfoLedger struct {
    ShimVersion  string               `json:"shim_version"`
    BeetsVersion string               `json:"beets_version"`
    Rows         []AlbumInfoLedgerRow `json:"rows"`
}

type AlbumInfoLedgerRow struct {
    CaseID    string           `json:"case_id"`
    Candidate int              `json:"candidate"` // index into Expected.Candidates
    Album     LedgerAlbumInfo  `json:"album"`
}

// LedgerAlbumInfo mirrors matcher.AlbumInfo field for field, with every
// string a *string so a beets None survives the round trip.
type LedgerAlbumInfo struct {
    Artist        *string           `json:"artist"`
    Album         *string           `json:"album"`
    VA            bool              `json:"va"`
    Media         *string           `json:"media"`
    Mediums       *int              `json:"mediums"`
    Year          *int              `json:"year"`
    OriginalYear  *int              `json:"original_year"`
    Country       *string           `json:"country"`
    Label         *string           `json:"label"`
    CatalogNum    *string           `json:"catalognum"`
    AlbumDisambig *string           `json:"albumdisambig"`
    AlbumID       *string           `json:"album_id"`
    DataSource    *string           `json:"data_source"`
    Tracks        []LedgerTrackInfo `json:"tracks"`
}
```

`LedgerTrackInfo` mirrors the eight `TrackInfo` fields the same way,
with `Length *float64`.

Why field-for-field rather than `map[string]any`. The comparison the
checker performs is per field with a per-field nullability rule; a map
would push the field names into string literals at every comparison
site and lose the compile-time guarantee that the ledger and the
vocabulary stay the same shape. Adding a fifteenth field to `AlbumInfo`
without extending `LedgerAlbumInfo` should not compile the comparison
silently — a struct gives that, a map does not. The pointer types are
not defensive nil-checking (which go-conventions forbids for internal
code); they are the JSON boundary case the same document carves out
for `[]T{}`, and the reason belongs in the type's doc comment.

The comparison, in `conform`:

```go
// CheckTranslatedAlbum compares one ledger row against the AlbumInfo
// the Go translation built, field by field, and reports every
// disagreement rather than the first.
func CheckTranslatedAlbum(caseID string, want analysis.LedgerAlbumInfo, got *matcher.AlbumInfo) error
```

`errors.Join` over per-field errors, matching `CheckCaseWeights`'
shape (`conform/defaults.go:21-37`). The nullability policy lives as a
small table in this file with one line of reason per field, so a
reader sees why `media` tolerates null and `album` does not without
reading beets.

The Python entry point is `internal/harness/oracle/derive_album_info.py`,
driven by a `DeriveAlbumInfo` in an `//go:build oracle` file beside
`component.go`
(`/var/home/glw907/Projects/dubplate/internal/harness/oracle/component.go:1-16`).
Unlike `derive_component.py` it needs the stub, the rendered profile,
and the response set, so it is `derive_oracle.py`'s shape, not
`derive_component.py`'s: reuse `runShim`, `renderProfile`,
`writeShimInput` (`oracle.go:232,280,294`), keep `PYTHONHASHSEED=0`,
keep `stable_json`, and pin a `SHIM_VERSION`. It dumps `AlbumInfo` per
candidate rather than calling `tag_album`'s verdict path, so the
natural entry point is `MusicBrainzPlugin.album_info` over each
replayed release response.

#### 2.5 The `defect_free` block

Additive on the existing schema
(`/var/home/glw907/Projects/dubplate/internal/harness/analysis/adjudication.go:44-55`):

```go
// DefectFree is the distance a case scores once the named defect is
// removed, derived through pinned beets with one SD_PATTERNS edit. It
// is the expectation an oracle-wrong case is graded against once its
// hypothesis's compat switch defaults off (ruling A / M6).
type DefectFree struct {
    Distance       float64 `json:"distance"`
    Recommendation string  `json:"recommendation"`
    Action         string  `json:"action"`
    PatternTable   string  `json:"pattern_table"` // the named table, e.g. "SDPatternsWordBoundary"
}
```

on `AdjudicationEntry` as `DefectFree *DefectFree
\`json:"defect_free,omitempty"\``. Pointer, because absence is
meaningful: `omitempty` on a struct value does not omit. `PatternTable`
names the table rather than restating the regex, because
`Tunables.StringDistPatterns` and `SDPatternsWordBoundary` already
exist as the swap seam (`tunables.go:196`, `:333-340`, and guard 3 of
the HISTORY entry). The checker extends `checkAdjudicationEntry`
(`adjudication.go:100-146`) with the rule that an `oracle-wrong` row
on a string-heuristic case carries one, and its fixture is a
hand-authored ledger in the package's own `testdata` with the block
removed.

The pre-computed figures to seed it are in the HISTORY entry's guard
3: `string_dist("Draft: Ember Wharf", "Draft: Northern Lines")` reads
`0.061111111111111116` under `BeetsSDPatterns` and
`0.6111111111111112` under `SDPatternsWordBoundary`; the
`adj-adv-ft-draft-colon` prose figures are `0.010185185` and
`0.101851852`.

### 3. Conformance to go-conventions

The proposal was held to `~/.claude/skills/go-conventions/SKILL.md`.
Checks that bind: no interfaces (both new funcs take concrete types);
no functional options (`Tunables` is a plain config struct, passed by
value as `DistanceForAlbum` already does); errors lowercase,
context-first, `%w` only where a caller would branch;
`slices.SortStableFunc` with `cmp.Compare` for the relations grouping,
never `sort.SliceStable`; `for range n` in the extras loops;
`errors.Join` for multi-field checkers, matching `CheckCaseWeights`; no
`init()`; no `context.Context` in a struct; `any` not `interface{}`.
Two places need a written reason in the doc comment because they look
like anti-patterns and are not: the `*string` ledger fields (the JSON
null/empty distinction the skill's `[]T{}` rule carves out) and
`GroupRelations` returning a map whose key order is unobservable.

---

## (d) Cross-section consistency and open questions

### Field-name reconciliation: (a)'s beets fields against (c)'s Go vocabulary

Every beets `AlbumInfo`/`TrackInfo` field section (a) names as
distance-read or translation-relevant is checked here against the Go
fields section (c) names in `matcher.AlbumInfo` / `matcher.TrackInfo`
(`info.go:10-105`).

| beets field | Go field | Status |
| --- | --- | --- |
| `album_info.artist` | `AlbumInfo.Artist` | mapped |
| `album_info.album` | `AlbumInfo.Album` | mapped |
| `album_info.va` | `AlbumInfo.VA` | mapped |
| `album_info.media` | `AlbumInfo.Media` | mapped |
| `album_info.mediums` | `AlbumInfo.Mediums` | mapped |
| `album_info.year` | `AlbumInfo.Year` | mapped |
| `album_info.original_year` | `AlbumInfo.OriginalYear` | mapped |
| `album_info.country` | `AlbumInfo.Country` | mapped |
| `album_info.label` | `AlbumInfo.Label` | mapped |
| `album_info.catalognum` | `AlbumInfo.CatalogNum` | mapped |
| `album_info.albumdisambig` | `AlbumInfo.AlbumDisambig` | mapped |
| `album_info.album_id` | `AlbumInfo.AlbumID` | mapped |
| `album_info.data_source` | `AlbumInfo.DataSource` | mapped |
| `album_info.tracks` | `AlbumInfo.Tracks` | mapped |
| `track_info.title` | `TrackInfo.Title` | mapped |
| `track_info.artist` | `TrackInfo.Artist` | mapped |
| `track_info.length` | `TrackInfo.Length` | mapped |
| `track_info.index` | `TrackInfo.Index` | mapped |
| `track_info.medium_index` | `TrackInfo.MediumIndex` | mapped |
| `track_info.medium` | `TrackInfo.Medium` | mapped |
| `track_info.track_id` | `TrackInfo.TrackID` | mapped |
| `track_info.data_source` | `TrackInfo.DataSource` | mapped |
| `track_info.media` (per-track, feeds `AlbumInfo.media`'s derivation, section (a).4) | **none** | **missing** — flagged as open question 1 below; (c) proposes keeping it out of `TrackInfo` and computing `AlbumInfo.Media` inside `translate.go` from an internal per-track record |
| `va_name` (config, writes `info.artist` on VA) | **none in `Tunables`** | **missing** — (c) §2.1's `Tunables` reading and open question 6 both name this; (c) recommends adding `VAName` |
| `musicbrainz.aliases_as_credits` (config, selects credit source in `_parse_artist_credits`) | **none in `Tunables`** | **missing** — same open question |
| `artist_sort`, `artist_credit`, `artists`, `artists_sort`, `artists_credit`, `artists_ids`, `release_group_title`, `script`, `language`, `disctitle`, `medium_total`, `track_alt`, `release_track_id`, external-id fields | **none, by rung 3's stated scope** | out of scope for the distance-facing `AlbumInfo`; these are applied-field territory (rung 8) and surface only in open question 2 of the second review round, on the ruling-E ledger's own field scope, not in `matcher.AlbumInfo` |

No beets field that (a) marks distance-read is unmapped in (c)'s
`matcher.AlbumInfo`/`TrackInfo`. The three genuine gaps are per-track
`media`, `va_name`, and `aliases_as_credits`, all already carried as
named open questions.

### Corpus figures from (b) that (c)'s proposal depends on

(c)'s design leans on several of (b)'s measured figures. Restated
here so the plan can cite one place:

- **The `defect_free` remix table has no corpus-derived precedent
  today** (b) §9's "one table exists, and 3b needs a second" is the
  basis for (c) §2.5's `PatternTable` field needing a third
  `Tunables.StringDistPatterns` variant beyond `BeetsSDPatterns` and
  `SDPatternsWordBoundary`.
- **The alias path changes 90/303 album `artist` values and 3,272/6,919
  track `artist` values under `languages=['ja']`, but only one
  committed case actually runs under a locale profile** (b) §7 and
  §13 — this is why (c) §2.1's `VAName`/`AliasesAsCredits` additions
  and the ledger's field scope (open question 2, second round) both
  need a ruling on how the alias path gets graded, not just how it is
  typed.
- **Ten releases produce `media == "Media"`, five with zero tracks**
  (a) §3, (b) §13 — this is why (c) §2.1's per-track media gap (open
  question 1) is not a corner case but a measured 10/303 occurrence,
  and why (c)'s recommendation to keep `Media` out of `TrackInfo` and
  compute it separately in `translate.go` must still reproduce the
  sentinel exactly.
- **Four pseudo-release cases exist, and the merge moves only
  `media`, `country`, `label`, `catalognum`, `barcode`, `asin` on
  committed data; `year` and `mediums` never move** (a) §8, (b) §7 —
  this is the entire evidence base for (c) §2.2's `MergePseudoRelease`
  seven-field contract; the merge's `year` path is unverified by any
  committed case, a gap (c) does not separately flag but that a
  reviewer should weigh against the confidence of that function's
  contract.
- **Twenty-one adjudication rows, eighteen `oracle-wrong`, three
  `oracle-correct`, zero pending; the B3 gap is closed** (b) §8 — this
  is what lets (c) §2.5 treat the `defect_free` block's ten seed
  figures as settled data rather than as pending adjudication.
- **Zero committed cases exercise pregap, data tracks, or the
  preferred-country release-event branch** (a) §13, (b) §7 — this
  grounds (b)'s recommendation (open question 5, second round) that
  3b declare three more M8-style corpus-coverage gaps, and it is also
  why (c)'s `TranslateAlbum` can be written and tested against
  hand-authored fixtures for these branches without corpus support.
- **Response bodies are raw, hyphenated wire JSON with no
  normalization performed by the corpus** (b) §10 — this is the direct
  premise of (c) §2.1's decision that the Go translation layer must
  perform hyphen and relations normalization itself, via struct tags
  plus `GroupRelations`.

### Deduplicated open questions

Some questions recur near-verbatim across the three source documents'
own open-question lists (embedded in (a), (b), and (c) as authored)
and across the two review rounds supplied. Each entry below names
every section that raised it.

1. **The absent-source-field ruling (3a's carried item):** keep plain
   strings with a source-scoped unreachability proof, switch
   `AlbumInfo.Album`/`.Artist` to `*string`, or something else.
   Raised by: (a) §9/§14, (c) §2.3, first-round open questions.

2. **`matcher.TrackInfo` has no per-track `Media` field, but
   `AlbumInfo.Media` is derived from the set of per-track media
   values, including the "Media" sentinel.** Does 3b add a `Media`
   field to `TrackInfo`, or keep an internal per-track record in
   `translate.go` and compute `AlbumInfo.Media` before building the
   matcher's `TrackInfo` values?
   Raised by: (a) §4, (c) §1 ("A port hazard" / vocabulary gap),
   first-round open questions.

3. **The literal `"Media"` sentinel and the all-ignored-media
   release:** is deriving and grading this in 3b's scope, and does a
   hand-authored fixture prove it fires rather than relying on the ten
   corpus cases that happen to exercise it?
   Raised by: (a) §3/§13, first-round open questions.

4. **`preferred.countries` carries two different matching semantics**
   (exact case-sensitive membership in release-event selection versus
   a case-insensitive compiled regex in the country penalty). Does
   `Tunables` expose one field consumed two ways, or two derived
   views? Related: does the translate side need any config
   validation, given it compiles no regex itself?
   Raised by: (a) §7, (c) §1 ("A port hazard"), first-round open
   questions, second-round open question 5 (validation ownership,
   3b vs. 3d).

5. **How is the alias path graded, given it is the layer's
   highest-leverage surface but only one committed case derives under
   a locale profile?** Options include deriving the ledger under both
   a case's own profile and a locale variant, hand-authored
   table-driven tests, or accepting the one-case coverage as a
   declared gap.
   Raised by: (a) §13, (b) §7, first-round open questions,
   second-round open question 4.

6. **What does the ruling-E derived `AlbumInfo` ledger record, and how
   is it keyed?** Field scope (all 41/39 beets keys, or the 14/8
   distance-read fields, or those plus the alias-hypothesis display
   strings), and keying (case-id alone is not unique across the four
   pseudo-release cases with two release lookups each; should the
   ledger key on `(case_id, mbid)` and record pre-merge pseudo and
   actual values alongside the merged one)?
   Raised by: (a) §5/§8, first-round open questions, second-round
   open question 2.

7. **`va_name` and `musicbrainz.aliases_as_credits` are read inside
   translation, absent from the spec's enumerated tunable surface, and
   ledger-moving but not verdict-moving.** Do they join `Tunables` with
   that rationale recorded, or does translation carry its own small
   config struct?
   Raised by: (a) §10, (c) §2.1, first-round open questions,
   second-round open question 1.

8. **Is the browse fallback (`_ensure_complete_recordings`,
   `BROWSE_MAXTRACKS=500`, largest committed release at 470 tracks)
   ported, or declared out of scope with a named refusal and a guard
   against a future release crossing the line?**
   Raised by: (a) §13, first-round open questions, second-round open
   question 2.

9. **How does the Go translation layer behave at the five sites where
   beets raises rather than propagating a zero value** (`artists_ids[0]`
   IndexError, `recording["aliases"]` KeyError, `actual[k]` KeyError in
   the merge, the direct indexes in `album_info`'s constructor call,
   `target_type.replace` AttributeError on an untyped relation)?
   Raised by: (a) §9, first-round open questions.

10. **Does the `defect_free` block's derivation entry point reuse
    `derive_component.py`'s shim shape and `derive_oracle.py`'s
    `stable_json`, riding the same `SD_PATTERNS`-as-a-value seam 3a
    landed, given tier 0 is blind to a "fixed" table?**
    Raised by: (a) §9 (guard reference), (c) §2.5, first-round open
    questions.

11. **Where does the derived `AlbumInfo` ledger's checker live,
    `internal/harness/analysis` or `internal/harness/conform`?** The
    spec's pass-split row names only `corpus/` and
    `internal/harness/oracle` as 3b's tree, `file_tags` is the closest
    precedent and lives in `analysis`, and the contested-surface table
    awards `internal/harness/{cassette,stub,analysis}` to lane E1
    through 2d's close (now passed).
    Raised by: second-round open question 1.

12. **What is the graded projection of the derived `AlbumInfo`
    ledger** — 3a's distance-only vocabulary, or the larger set of
    beets fields the alias hypothesis's evidence lives in (`artist_sort`,
    `artist_credit`, etc.), given rung 3 is stated to grade no applied
    field?
    Raised by: second-round open question 2 (a variant of question 6
    above, focused on the scope boundary against rung 8).

13. **Does the `defect_free` block carry structured figures alongside
    or instead of `CorrectedVerdict`'s existing prose, and does the
    checker cross-check the two, or are the prose figures retired?**
    Raised by: second-round open question 3.

14. **Do the six `adv-assign-*` rows need a `defect_free` block, and
    if the M6 population is fixed at the ten string-heuristic cases,
    what predicate distinguishes a "string-heuristic case" from a
    hardcoded id list?**
    Raised by: second-round open question 4.

15. **Should 3b declare pregap, data tracks, and the preferred-country
    release-event branch as M8-style corpus-coverage gaps, given all
    three are measurably unreachable from the committed corpus?**
    Raised by: (a) §13, (b) §7, second-round open question 5 (a
    variant of question 3 above, generalized to three named gaps).

16. **Who corrects `corpus/README.md`'s three stale adjudication cells
    and the `locale.yaml` placeholder error — 3b itself, or a separate
    document lane, given both lanes would otherwise write the same
    file?**
    Raised by: (b) §11, second-round open question 6.

17. **Does `Tunables` gain `VAName` and `AliasesAsCredits`, given
    neither moves a distance but both move `AlbumInfo.Artist`, which
    the ruling-E ledger grades, and no committed case echoes either
    for the M1 cross-check to cover?**
    Raised by: third-round open question 1 (a restatement of question
    7 above with the M1-coverage detail added).

18. **How does 3b scope `_ensure_complete_recordings`, given
    `TranslateAlbum(rel, tun)` cannot itself issue paginated browse
    requests — omit and record the gap, take an injected fetcher left
    nil in 3b, or defer the whole entry point to 3c?**
    Raised by: third-round open question 2 (a restatement of question
    8 above with three concrete signature options).

19. **What are the ledger's subjects and its row floor** — every
    candidate of every case, or a named subject list biased toward the
    alias, pseudo-release, and locale strata, given the gate's
    dominant cost is already the corpus scan?
    Raised by: third-round open question 3 (a restatement of question
    6 above, focused on the row-count budget rather than field scope
    or keying).

20. **Does the alias-selection path get graded at all under the
    shipped defaults, given `import.languages` ships empty and only
    `locale.yaml` sets it?**
    Raised by: third-round open question 4 (a restatement of question
    5 above).

21. **Is 3a's second carried item (config validation for
    `Preferred.Media`/`Preferred.Countries` regexes, which panic
    today) 3b's or 3d's, given `internal/config` has no validation
    hook and `keys.go` is frozen until rung 1b?**
    Raised by: third-round open question 5 (a restatement of question
    4's validation-ownership half above).
