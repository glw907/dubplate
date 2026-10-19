# beets matcher mechanics at v2.13.1

Research report, 2026-09-02, produced by a research agent from a
source-level read of the pinned beets release. Source of truth: the
`v2.13.1` tag of github.com/beetbox/beets, commit
`b7952299941543d4507ac7931edb223acd684b3d`, `__version__ = "2.13.1"`
(beets/__init__.py:13). All file:line citations below are against that
checkout. Estate-config claims are grounded in
`config/beets-config.yaml` in this repo. Every specific stated here
was verified in the pinned source; nothing is completed from memory.

Two version-2.13 facts up front, because they invalidate older mental
models of beets:

- beets no longer uses `musicbrainzngs`. The MusicBrainz client is
  in-tree (`beetsplug/_utils/musicbrainz.py`) on top of `requests`,
  speaking WS/2 JSON.
- The legacy `album_distance` / `track_distance` plugin pipeline
  hooks are gone. No `plugins.send` call exists anywhere in the
  distance computation (verified by grep over beets/autotag/ and
  beets/plugins.py). Plugins influence matching only as metadata
  sources and through the info-mutation events listed under Q1.

Dependency facts: `jellyfish` provides `levenshtein_distance`
(beets/autotag/distance.py:8) and is a required dependency
(pyproject.toml:29), so there is no optional-speedup fork like the
old python-Levenshtein situation. Track assignment uses `lap.lapjv`
plus numpy (beets/autotag/match.py:12-13; pyproject.toml:30-33).
`unidecode` is required (pyproject.toml:40, floor `>=1.3.6`).

## Q1. Purity

Verdict: a beets match verdict at 2.13.1 is a deterministic function
of (items including their order, MusicBrainz responses, beets config,
the current calendar year, and the exact versions of unidecode,
jellyfish, and lap) — with one live ordering hazard that the estate's
single-source config happens to neutralize. There is no randomness
anywhere in the autotag path.

Hazards found, each with liveness:

1. **Wall clock: `datetime.date.today().year` inside the year
   penalty.** Both year-distance branches divide by a span ending at
   the current year (beets/autotag/distance.py:477 and 487-489). Live
   under stock config and under the estate config whenever the items
   carry year tags and the candidate has an `original_year` (or
   `preferred.original_year` is on). The harness must pin or record
   the evaluation year per case; a Go matcher should take the
   reference date as an explicit input, and the differential job must
   run beets under a pinned clock (or record the year the expected
   verdict was computed under and re-derive at year boundaries).

2. **Thread fan-out across metadata sources.**
   `metadata_plugins.candidates` (and the three sibling entry points)
   run every metadata source plugin in a `ThreadPoolExecutor` and
   yield results in `as_completed` order
   (beets/metadata_plugins.py:88-110). With two or more sources the
   interleaving of sources in the candidate stream is
   timing-dependent. Within one plugin, results are materialized with
   `list(...)` (metadata_plugins.py:82-86), so per-source order is
   preserved. Under the estate config exactly one metadata source is
   loaded (`plugins: musicbrainz fetchart embedart replaygain`;
   fetchart/embedart/replaygain define no `data_source` and are not
   metadata sources per the filter at beets/metadata_plugins.py:51),
   so there is a single future and the hazard is dead. The
   conformance corpus should stay single-source; a multi-source case
   would need per-source ordering recorded.

3. **Item order is matcher input.** `get_most_common_tags` picks
   "likely" album fields via `plurality`
   (beets/util/__init__.py:799-846), which is `Counter.most_common(1)`
   — ties break by first-seen insertion order, i.e. by the order of
   `items`. The assignment cost matrix is also built in item order
   (match.py:64). Fixtures must pin item order per case; it is part
   of the input, not noise.

4. **String-distance library versions.** `unidecode` transliteration
   tables change across releases and sit directly in
   `_string_dist_basic` (distance.py:55-56); the corpus manifest must
   record the exact unidecode version beets ran with. `jellyfish`
   computes plain integer Levenshtein — algorithmically stable across
   its C and Rust implementations — and `lap` is deterministic for a
   fixed version. Pin all three in the differential environment.

5. **No locale sensitivity found.** The path uses `str.lower()`
   (locale-independent Unicode default casing) and ASCII regexes;
   there is no `locale` module usage in beets/autotag/ or the MB
   plugin. Unicode normalization (NFC/NFD) is *not* performed
   anywhere in the distance path: byte-different but
   canonically-equivalent tag strings hit unidecode as-is. Frozen
   fixtures preserve exact strings, so this is stable, but a Go port
   must not "helpfully" normalize.

6. **Dict/set iteration.** All verdict-bearing containers are
   insertion-ordered dicts written in fixed code order: the
   candidates map keyed by `(data_source, id)` (match.py:414,
   486), `Distance._penalties` (distance.py:130, filled in the fixed
   order of `distance()` / `track_distance()`), and `dist.tracks`.
   The `max_rec` downgrade iterates a `set` of penalty keys
   (match.py:306-321) but folds with `min`, which commutes. The
   only genuine set-order leak: `extra_items` / `extra_tracks` come
   from set differences and are then sorted by `(disc, track,
   title)` / `(index, title)` (match.py:77-80); two extras with
   identical sort keys keep arbitrary relative order (PYTHONHASHSEED
   / address dependent). Extras never contribute to the distance
   (only their counts do, distance.py:532-538), so this cannot move
   a verdict, only display/application order of leftovers.

7. **Plugin hook points that can touch matching.** Complete list at
   this version: (a) metadata source plugins supply candidates;
   (b) `albuminfo_received` / `trackinfo_received` fire once per
   yielded info object (metadata_plugins.py:113-134 via
   `notify_info_yielded`, beets/plugins.py:513-533) and handlers
   receive the mutable info object; (c) `mb_album_extract` /
   `mb_track_extract` let plugins merge extra fields into MB-built
   infos (beetsplug/musicbrainz.py:465-467, 691-693);
   (d) `import_task_before_choice` can decide the action outright
   (beets/ui/commands/import_/session.py:50-61);
   (e) `album_matched` fires after an AlbumMatch is constructed
   (match.py:153-155), post-verdict. None of the estate's non-MB
   plugins (fetchart, embedart, replaygain) registers any of these
   or defines `data_source` (verified by grep); under the estate
   config the hook surface is inert.

8. **Config read-once caching.** `Distance._weights` is a cached
   class property (distance.py:133-140), the track-length knobs are
   `@cache`d (distance.py:364-373), and `get_penalty` is `@cache`d
   (metadata_plugins.py:159-169). Config is effectively frozen at
   first use per process; irrelevant across processes but worth
   knowing when driving beets in-process.

What the harness must pin or record per case: item order, the
evaluation year, the full effective config (or a named profile), the
single-source constraint, and the {unidecode, jellyfish, lap, beets}
version quadruple in the corpus manifest.

## Q2. The distance and penalty pipeline

This section is the port map. Everything below is
beets/autotag/distance.py unless cited otherwise.

### String distance

`_string_dist_basic(str1, str2)` (lines 47-61):

1. `unidecode()` each string, coerce to str (`as_string`,
   beets/util/__init__.py:786-796).
2. Lowercase, then strip every char not in `[a-z0-9]` (lines 57-58).
3. Both empty after stripping: 0.0.
4. Else `levenshtein_distance(a, b) / max(len(a), len(b))` (line 61).

`string_dist(str1, str2)` (lines 64-119):

1. Both None: 0.0. Exactly one None: 1.0 (lines 69-72).
2. Lowercase both (74-75). Note the order: this lowercase happens
   *before* unidecode (which runs inside `_string_dist_basic`); a
   port must replicate lower-then-unidecode-then-lower exactly.
3. Article rotation: for each of `SD_END_WORDS = ["the", "a", "an"]`
   (line 33), a trailing `", <word>"` moves to the front:
   `"something, the"` becomes `"the something"` (lines 80-84).
4. Replacement: `&` -> `and` via regex (`SD_REPLACE`, line 44,
   applied 87-89).
5. `base_dist = _string_dist_basic(str1, str2)` (line 95).
6. Heuristic pattern pass, in the fixed order of `SD_PATTERNS`
   (lines 35-42): `(^the , 0.1)`, `([\[\(]?(ep|single)[\]\)]?, 0.0)`,
   `([\[\(]?(featuring|feat|ft)[\. :].+, 0.1)`, `(\(.*?\), 0.3)`,
   `(\[.*?\], 0.3)`, `((, )?(pt\.|part) .+, 0.2)`. For each pattern:
   delete it from both strings; if either changed, recompute
   `case_dist`; `case_delta = max(0, base_dist - case_dist)`; if the
   delta is positive, adopt the stripped strings as the new
   baseline, set `base_dist = case_dist`, and add
   `weight * case_delta` to a running penalty (lines 97-117).
7. Return `base_dist + penalty` (line 119).

### The Distance object

State: `_penalties: dict[str, list[float]]`, each entry in [0, 1]
(enforced in `add`, lines 269-277), plus `tracks: dict[TrackInfo,
Distance]` (lines 129-131).

Weights come from `config["match"]["distance_weights"]`, read once
(lines 133-140). Defaults (beets/config_default.yaml:173-193):

    data_source 2.0, artist 3.0, album 3.0, media 1.0, mediums 1.0,
    year 1.0, country 0.5, label 0.5, catalognum 0.5,
    albumdisambig 0.5, album_id 5.0, tracks 2.0, missing_tracks 0.9,
    unmatched_tracks 0.6, track_title 3.0, track_artist 2.0,
    track_index 1.0, track_length 2.0, track_id 5.0, medium 1.0

The math (lines 152-176):

    raw_distance = sum over keys of sum(penalties[key]) * weight[key]
    max_distance = sum over keys of len(penalties[key]) * weight[key]
    distance     = raw_distance / max_distance   (0.0 if max is 0)

So the normalizer is not a fixed table: only penalties that were
*added* (even at value 0.0) contribute weight to the denominator.
Which `add_*` calls fire — and therefore the denominator — depends on
which fields exist on both sides. A per-key weighted distance
`dist[key]` is `sum(penalties[key]) * weight[key] / max_distance`
(lines 232-238). `items()` returns nonzero keys sorted by
(-weighted_dist, key) (lines 190-205); `keys()` therefore contains
only keys with nonzero weighted distance (line 246-247) — this
matters for `max_rec` and for `match.ignored`.

Component adders (lines 260-361):

- `add(key, dist)`: append one penalty in [0,1].
- `add_equality(key, value, options)`: 0.0 if any option equals
  value (compiled regexes match instead of compare), else 1.0.
- `add_expr(key, expr)`: bool to 1.0/0.0.
- `add_number(key, n1, n2)`: appends `abs(n1-n2)` separate 1.0
  penalties (or one 0.0) — one key can carry multiple units, which
  also inflates the denominator by weight per unit.
- `add_priority(key, value, options)`: penalty `i * (1/len(options))`
  for the first matching option index i, 1.0 if none match.
- `add_ratio(key, n1, n2)`: `clamp(n1, 0, n2) / n2` (0.0 if n2
  falsy).
- `add_string(key, s1, s2)`: `string_dist`.
- `add_data_source(before, after)` (lines 357-361): only when
  `before != after` and (`before` is set or more than one metadata
  source is loaded), adds `get_penalty(after)` — the target source
  plugin's `data_source_mismatch_penalty` (default 0.5, legacy name
  `source_weight`; beets/metadata_plugins.py:159-169, 180, 190-195).
  Fresh imports with a single source add nothing at all (not even a
  0.0), so the key stays out of the denominator.

### Track distance

`track_distance(item, track_info, incl_artist=False)`
(lines 383-426), components in this fixed order:

1. `track_length` (weight 2.0): only if `track_info.length` is
   truthy. `diff = abs(item.length - info.length) -
   track_length_grace` (default 10, config_default.yaml:203), then
   `add_ratio("track_length", diff, track_length_max)` (default 30,
   line 204). Negative diff clamps to 0; diff beyond max clamps
   to 1.
2. `track_title` (3.0): `add_string` always.
3. `track_artist` (2.0): only if `incl_artist` and
   `track_info.artist` and `item.artist.lower()` not in
   `VA_ARTISTS = ("", "various artists", "various", "va",
   "unknown")` (lines 29, 404-410).
4. `track_index` (1.0): only if `info.index` and `item.track`;
   penalty 1.0 iff `item.track not in (info.medium_index,
   info.index)` (lines 376-380, 412-414) — tolerating either
   per-disc or absolute numbering.
5. `track_id` (5.0): only if `item.mb_trackid` set; 1.0 iff it
   differs from `info.track_id` (416-418).
6. `medium` (1.0): only if `info.medium` and `item.disc`; 1.0 iff
   unequal (420-422).
7. `data_source` (2.0): as above (424).

### Album distance

`distance(items, album_info, item_info_pairs)` (lines 429-542),
`likelies` from `get_most_common_tags` (beets/util/__init__.py:
810-846; albumartist consensus overrides artist, lines 842-844):

1. `artist` (3.0): skipped entirely when `album_info.va` (447-448).
2. `album` (3.0): always (451).
3. `media` (1.0): if `album_info.media`: when
   `match.preferred.media` patterns exist, `add_priority` against
   compiled `(\d+x)?(<pat>)` case-insensitive regexes (455-462);
   else when `likelies["media"]`, `add_equality` (464-465).
4. `mediums` (1.0): if both sides present, `add_number` on
   disctotal vs `album_info.mediums` — each disc of difference is a
   full weighted unit (468-469).
5. `year` (1.0): if `preferred.original_year` and `album_info.year`:
   `original = original_year or 1889`; `add_ratio(|year -
   original|, |today.year - original|)` (472-478). Else if both
   `likelies["year"]` and `album_info.year`: 0.0 when the item year
   equals the release year or the original year; else with an
   original year, `add_ratio(|likelies - year|, today.year -
   original_year)`; else flat 1.0 (480-493).
6. `country` (0.5): with `preferred.countries`, `add_priority`
   against case-insensitive regexes; else when both present,
   `add_string` on the country codes (496-502).
7. `label` (0.5), `catalognum` (0.5), `albumdisambig` (0.5):
   `add_string` when both sides present (505-518).
8. `album_id` (5.0): when the items carry an mb_albumid,
   `add_equality` against `album_info.album_id` (521-524).
9. `tracks` (2.0): for each matched (item, track) pair, compute
   `track_distance(item, track, incl_artist=album_info.va)` and add
   that track's *normalized* distance as one penalty under
   `tracks` (527-530).
10. `missing_tracks` (0.9): one 1.0 penalty per candidate track
    beyond the matched pairs (533-534). `unmatched_tracks` (0.6):
    one 1.0 penalty per item beyond the matched pairs (537-538).
11. `data_source` (2.0): as above (540).

### Track assignment

`assign_items(items, tracks)` (match.py:53-81): a full cost matrix
`costs[i][t] = float(track_distance(item_i, track_t))` — note this
calls `track_distance` with the default `incl_artist=False`, so the
assignment objective differs from the final per-track distances,
which pass `album_info.va` (distance.py:529). The matrix is solved
with `lap.lapjv(np.array(costs), extend_cost=True)` (rectangular
matrices padded internally); the third return value maps each track
to an item index or -1. Unassigned leftovers become `extra_items`
(sorted by disc, track, title) and `extra_tracks` (sorted by index,
title) (match.py:72-81).

### From distance to verdict

`_recommendation(results)` (match.py:269-323) on the
distance-sorted candidate list:

    d = results[0].distance
    d <  strong_rec_thresh (default 0.04)  -> strong    # strict <
    d <= medium_rec_thresh (default 0.25)  -> medium
    len(results) == 1                      -> low
    results[1].distance - d >= rec_gap_thresh (0.25) -> low
    otherwise -> none (returned immediately, skips downgrades)

(Thresholds: config_default.yaml:167-169. Note the strict `<` for
strong at match.py:285 versus `<=` for medium at 288; the display
color property uses `<=` for both, distance.py:180-184 — a
display-only inconsistency.)

Downgrade pass (match.py:304-323): collect the best match's nonzero
penalty keys — album-level keys plus, for an AlbumMatch, every key
of every per-track Distance — and for each key present in
`match.max_rec` (defaults `missing_tracks: medium`,
`unmatched_tracks: medium`, config_default.yaml:170-172), clamp
`rec = min(rec, configured level)`. Because `keys()` only yields
nonzero penalties, a 0.0 penalty never triggers a downgrade.

Verdict to action (`_summary_judgment`,
beets/ui/commands/import_/session.py:328-360): with
`import.quiet` (the estate's `beet import -q`): strong applies
`task.candidates[0]` (session.py:64-68, 121-125); anything else
takes `quiet_fallback` (estate: `skip`). Without quiet: `timid`
always asks; `none` takes `none_rec_action` (skip/asis/ask); medium
and low ask. Plugins may preempt via `import_task_before_choice`
(session.py:50-61).

### Candidate admission filters

`_add_candidate` (match.py:331-380), applied before a candidate ever
gets a distance rank: zero-track albums are dropped; duplicate
`(data_source, album_id)` identifiers are dropped; a candidate
missing any `match.required` field is dropped; after scoring, a
candidate with any nonzero penalty named in `match.ignored` is
dropped.

### tag_album flow

match.py:383-468. With explicit `search_ids`, only ID lookups run.
Otherwise: (1) if the items agree on an `mb_albumid`, look it up
(`match_by_id`, 252-266); if that yields candidates and
`import.timid` is off and the recommendation is strong, return
immediately with only those candidates (432-442); (2) else fall
through to text search with the likely artist/album, computing
`va_likely` = no artist consensus, or artist in `VA_ARTISTS`, or any
item flagged comp (451-456); (3) every source's candidates run
through `_add_candidate`; (4) stable sort by distance, recommend
(464-468). `tag_item` mirrors this for singletons with a
track-ID-first fast path (471-526).

## Q3. Candidate ordering

Origin and order, single-source (MusicBrainz) case:

1. The MB plugin builds one search query (Lucene, fields ANDed,
   terms lowercased/stripped/escaped;
   beetsplug/_utils/musicbrainz.py:650-685) and requests up to
   `search_limit` results (default 5;
   beets/metadata_plugins.py:199-204, 415).
2. `SearchApiMetadataSourcePlugin.candidates` maps the returned IDs
   through `albums_for_ids` with a generator, preserving the search
   response's order exactly (metadata_plugins.py:442-446, 250-259);
   failed lookups are filtered out without reordering.
3. `tag_album` inserts them into an insertion-ordered dict keyed by
   `(data_source, album_id)` (match.py:414, 459-462); ID-derived
   candidates (from the items' mb_albumid) were inserted before any
   search results (425-428).
4. `_sort_candidates` is a plain `sorted(...)` keyed on distance
   (match.py:326-328) — Python's Timsort, stable.

So the claim "beets preserves MusicBrainz candidate iteration order
and sorts stably" is CONFIRMED for a single metadata source: ties in
distance resolve to MB search-result order, with ID-match candidates
ahead of search candidates at equal distance. No unstable sort
exists in the path. The claim is REFUTED in general: with two or
more metadata sources the pre-sort stream interleaves sources in
`as_completed` (thread-timing) order (metadata_plugins.py:105-108),
so equal-distance ties across sources are nondeterministic. Distance
ties compare exact float equality (`Distance.__eq__`/`__lt__`,
distance.py:210-215), so "tie" means bit-identical doubles —
uncommon but real (e.g. two pressings differing only in fields
absent from the items).

## Q4. The tunable surface, verified complete

The believed surface (strong_rec_thresh, medium_rec_thresh,
rec_gap_thresh, max_rec, distance_weights, preferred, ignored,
required, ignored_media) is real but incomplete. The full
verdict-moving surface on unchanged matcher input and unchanged
frozen responses:

Under `match:` (defaults config_default.yaml:166-206):

- `strong_rec_thresh` 0.04, `medium_rec_thresh` 0.25,
  `rec_gap_thresh` 0.25 (match.py:285-297).
- `max_rec` {missing_tracks: medium, unmatched_tracks: medium}
  (match.py:310-321).
- `distance_weights` (19 keys, defaults listed in Q2).
- `preferred.countries` [] — moves verdicts TWICE: as the country
  add_priority (distance.py:496-499) and, missed by the believed
  list, by selecting which release event supplies the candidate's
  country AND date (beetsplug/musicbrainz.py:165-177), which feeds
  the year penalty.
- `preferred.media` [] (distance.py:457-462),
  `preferred.original_year` no (distance.py:472-478).
- `ignored` [] (match.py:369-375), `required` [] (match.py:354-359).
- `ignored_media` [] — filters whole media out of the candidate's
  track list (beetsplug/musicbrainz.py:638-639).
- **Missing from the believed list:** `track_length_grace` 10 and
  `track_length_max` 30 (distance.py:364-373, 396-399) — directly
  scale the track_length penalty.
- **Missing:** `ignore_data_tracks` yes, `ignore_video_tracks` yes
  (config_default.yaml:201-202; beetsplug/musicbrainz.py:555-580) —
  change the candidate track list and hence
  missing/unmatched-track counts.

Outside `match:`:

- **`musicbrainz.search_limit`** (default 5, set in
  beets/metadata_plugins.py:199-204; the legacy
  `match.searchlimit` spelling is honored with a deprecation,
  beetsplug/musicbrainz.py:290-299). Changes which candidates exist,
  which moves the single-candidate and rec_gap branches.
- **`musicbrainz.data_source_mismatch_penalty`** (legacy
  `source_weight`), default 0.5
  (beets/metadata_plugins.py:180-195) — the data-source penalty
  magnitude. The believed list had no data-source weight entry;
  both this and the `data_source: 2.0` weight belong on it. Note
  the penalty only fires when the item already carries a
  `data_source` differing from the candidate's, or when multiple
  sources are loaded (distance.py:357-361); under the estate's
  fresh-import, single-source setup it never fires.
- **`import.timid`** — gates the strong-ID-match early return
  (match.py:432-442, 499-501), changing which candidates compete.
- **`import.languages`** [] and **`import.ignored_alias_types`** []
  (config_default.yaml:32, 58) — drive `_preferred_alias`
  (beetsplug/musicbrainz.py:128-155), which rewrites candidate
  titles, artist names, and release titles
  (musicbrainz.py:158-162, 329-334, 443, 648) before any distance
  is computed.
- **`musicbrainz.extra_tags`** [] — adds search criteria
  (musicbrainz.py:697-734), changing the candidate set.
- **The `plugins` list itself** — which metadata sources load; also
  flips the multi-source condition inside `add_data_source`.
- **Verdict-to-action layer:** `import.quiet`, `quiet_fallback`,
  `none_rec_action` (session.py:328-360) — part of the surface if
  "verdict" means the final action, as it does for dubplate's
  apply/skip semantics.

Checked and found NOT verdict-moving: `va_name` (applied only after
`info.va` is decided from the MB Various-Artists artist ID,
musicbrainz.py:673-681, and the artist penalty is skipped for VA
albums); `artist_credit`, `original_date`, `per_disc_numbering`,
`overwrite_null` (application-time only, hooks.py raw_data/item_data
paths); `musicbrainz.genres`, `external_ids`, `aliases_as_credits`
(fields the distance never reads — aliases_as_credits only alters
credit strings, musicbrainz.py:331-333).

The estate config sets none of the `match:` or `musicbrainz:` keys,
so the whole surface runs at the defaults above; its verdict-moving
lines are `plugins: musicbrainz ...` (single source) and the
quiet-mode trio (`quiet_fallback: skip`, `none_rec_action: skip`,
`timid: no`).

## Q5. The HTTP seam

Client: in-tree `MusicBrainzAPI`
(beetsplug/_utils/musicbrainz.py:564-775) over a shared `requests`
session — not musicbrainzngs. Every request forces `fmt=json`
(musicbrainz.py:615-619), so the wire format is WS/2 JSON, never
XML.

Endpoints and inc= combinations actually used:

- Release lookup: `GET {root}/ws/2/release/{mbid}` with
  `inc=artists+media+recordings+release-groups+labels+
  artist-credits+aliases+recording-level-rels+work-rels+
  work-level-rels+artist-rels+isrcs+url-rels+release-rels+
  genres+tags` (RELEASE_INCLUDES, musicbrainz.py:50-67; joined with
  `+` at 634-637; get_release 687-690).
- Recording lookup (singletons): `GET /ws/2/recording/{mbid}` with
  `inc=artists+aliases+isrcs+work-level-rels+artist-rels`
  (RECORDING_INCLUDES 69-75; get_recording 692-697).
- Search: `GET /ws/2/release` or `/ws/2/recording` with
  `query=<lucene>&limit=<search_limit>&fmt=json` (search 665-685).
  Query construction: terms lowercased and stripped, Lucene special
  characters backslash-escaped (LUCENE_SPECIAL_CHAR_PAT, line 48),
  each filter rendered `field:(term)`, space-joined (AND semantics).
  Album criteria: `release` plus either `arid=<VA id>` (when VA is
  likely) or `artist`, plus any extra_tags criteria
  (beetsplug/musicbrainz.py:714-734). Track criteria: `artist`,
  `recording`, `alias` (line 749).
- Browse (only for releases whose media exceed 500 tracks total):
  `GET /ws/2/recording?release={id}&limit=100&offset=N` with
  `inc=artist-credits+aliases+work-rels+artist-rels+recording-rels+
  release-rels` (BROWSE_INCLUDES/CHUNKSIZE/MAXTRACKS 58-67;
  `_ensure_complete_recordings` 602-629), paginated by offset.

Response shape by the time the matcher sees it: raw WS JSON is
normalized by `_normalize_data` (musicbrainz.py:724-775) — every key
hyphen becomes underscore, recursively, and each generic
`relations` array is regrouped into `{target_type}_relations` keys
(sorted by target-type, stable within type). The MB plugin then
builds `AlbumInfo`/`TrackInfo` (`album_info` 631-695, `track_info`
433-469, per-medium track extraction 555-600 including pregap
inclusion, data/video-track filtering, and track-level overrides of
recording length/credit/title). The matcher (`match.py`,
`distance.py`) touches only AlbumInfo/TrackInfo, never raw WS data.

Where a stub sits: the host is configuration. `musicbrainz.host`
defaults to `musicbrainz.org` (forced https, hard 1 req/s); ANY
other host honors `musicbrainz.https` (default false) and
`ratelimit`/`ratelimit_interval` (musicbrainz.py:586-613). So a
fixture-replay server at `http://127.0.0.1:PORT` needs only
`musicbrainz: {host: "127.0.0.1:PORT", https: no, ratelimit: 100}`
in the beets-under-test config, serving `/ws/2/...` with JSON. This
is the cleanest possible seam — no monkeypatching needed.

Behavior the capture tool and differential job must reproduce or
account for (`beetsplug/_utils/requests.py`):

- Retries: urllib3 `Retry(total=6, backoff_factor=0.5)` on
  429/500/502/503/504 (requests.py:78-91). A stub must not emit
  those statuses unless testing retry behavior.
- Timeout: 10 s default per request (requests.py:99).
- Rate limiting is doubled up: a global `RateLimitAdapter` floor of
  0.25 s between any two requests (requests.py:89, 106-135) plus
  `requests_ratelimiter.LimiterMixin` at the configured per-second
  rate (musicbrainz.py:78-79, 612-613). Against a local stub with a
  high configured ratelimit the 0.25 s floor still applies — budget
  for it in differential-run wall time.
- Session is a process-wide singleton (SingletonMeta,
  requests.py:44-63); User-Agent is `beets/2.13.1 https://beets.io/`
  (requests.py:76) — the capture tool should present its own
  compliant UA against the real WS.
- Error mapping: 404 on release lookup returns None and the
  candidate is silently skipped (musicbrainz.py:777-781); 401 raises
  a credentials error (musicbrainz.py:551-561). Non-2xx otherwise
  raises.
- No response caching exists anywhere in this path.
- Pagination exists only in the >500-track browse fallback.
- **Pseudo-release follow-up fetch:** a release with status
  `Pseudo-Release` and a backward `transl-tracklisting` release
  relation triggers a SECOND `get_release` of the linked actual
  release, and the two AlbumInfos merge
  (musicbrainz.py:785-798, `_merge_pseudo_and_actual_album`
  203-240). Fixtures for such cases must record both responses; a
  capture tool recording only the requested MBID will break replay.

## Threats to the frozen-corpus design and the epsilon regime

1. **The wall-clock year (Q1.1)** is the only genuine impurity in
   the arithmetic. Verdicts drift at calendar-year boundaries for
   any case exercising the year penalty with an original_year
   present. Record the evaluation year in every fixture and inject
   it (libfaketime for the beets subprocess; an explicit reference
   date parameter in Go).

2. **Assignment ties are implementation-defined.** `lapjv` picks one
   optimal assignment; when several assignments share the minimal
   total cost, a different-but-equally-optimal choice by a Go
   solver yields the same assignment cost but NOT necessarily the
   same final distance, because the final per-track distances are
   recomputed with `incl_artist=va` while the assignment cost
   ignores the track artist (match.py:64 vs distance.py:529). The
   two objectives diverge exactly on VA albums. The corpus should
   detect cases with non-unique optimal assignments (perturb and
   re-solve, or check for duplicate row minima) and either exclude
   them or pin the expected mapping, and the Go port should mirror
   lapjv's tie behavior where feasible.

3. **Float regime: bit-equality is attainable, but do not lean on
   epsilon at thresholds.** Every accumulation runs in a fixed,
   code-determined order (penalty keys in call order, per-key lists
   in append order, raw/max sums over an insertion-ordered dict),
   and all operations are IEEE-754 double +, ×, ÷ with no
   transcendentals, so a Go port that replicates operation order
   reproduces bit-identical distances on any IEEE platform. The
   pragmatic acceptance criterion: compare distances at ~1e-12
   relative, but treat the *verdict*, not the distance, as the
   conformance unit — and flag any case whose best distance falls
   within ~1e-6 of 0.04, 0.25, or whose candidate gap falls within
   ~1e-6 of 0.25, because the strong boundary is strict `<` while
   medium is `<=` (match.py:285-288) and an epsilon cannot paper
   over a comparison-direction disagreement. Weights like 0.9 and
   thresholds like 0.04 are not exactly representable in binary;
   the Go side must parse them to the same doubles Python does
   (both use round-to-nearest decimal conversion, so identical
   literals yield identical doubles).

4. **Library-version pins are corpus metadata.** unidecode (
   transliteration tables feed the edit distance), jellyfish, lap,
   and beets itself must be recorded in the corpus manifest and
   frozen in the differential environment; a beets bump is a
   corpus-revalidation event (already the standing plan).

5. **Multi-source ordering (Q1.2)** would break tie reproducibility;
   the corpus stays single-source, matching the estate.

6. **Item order and the group_albums pre-step** are inputs: fixtures
   pin the item list order explicitly, and the corpus tests the
   matcher from the (items, responses) boundary, below the
   importer's grouping logic.

7. **Pseudo-release double-fetch (Q5)** requires the capture tool to
   follow and record the second release request, or replay 404s will
   silently change candidates.

## Estate plugin participation

The estate enables `musicbrainz fetchart embedart replaygain`
(config/beets-config.yaml:47). Only musicbrainz is a metadata
source; fetchart, embedart, and replaygain define no `data_source`
attribute and register none of the matching-adjacent events
(`albuminfo_received`, `trackinfo_received`, `mb_album_extract`,
`mb_track_extract`, `import_task_before_choice`, `album_matched`) —
verified by grep over their sources. None participates in matching
or distances. The estate's quiet-mode import maps the
recommendation to actions as: strong -> apply best candidate,
everything else -> skip (session.py:337-342 with
`quiet_fallback: skip`); `none_rec_action` is unreachable under
`-q`.

## Sources

- github.com/beetbox/beets, tag `v2.13.1`, commit
  `b7952299941543d4507ac7931edb223acd684b3d` (shallow clone,
  studied 2026-09-02). Files cited: `beets/autotag/distance.py`,
  `beets/autotag/match.py`, `beets/autotag/hooks.py`,
  `beets/metadata_plugins.py`, `beets/plugins.py`,
  `beets/util/__init__.py`, `beets/config_default.yaml`,
  `beets/ui/commands/import_/session.py`,
  `beetsplug/musicbrainz.py`, `beetsplug/_utils/musicbrainz.py`,
  `beetsplug/_utils/requests.py`, `pyproject.toml`.
- `config/beets-config.yaml` (this repo) for every estate-config
  claim.
- No web sources were needed; the pinned source answered every
  question.
