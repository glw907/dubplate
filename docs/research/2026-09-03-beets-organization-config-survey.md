# Beets library-organization config survey

Dispatched 2026-09-03 (Geoff, evening session) as the organization-surface
companion to the 2026-08-31 plugin-adoption survey. Method replayed from
that survey; sampling caveats identical and restated inline. Feeds the
organizing-engine rung spec (the 80% roster) alongside the config
methodology study of the same date.

---

# Beets library-organization config survey (n=124)

## (a) Method and n

Replayed the 2026-08-31 plugin-adoption survey's method (`docs/research/2026-08-31-beets-plugin-usage.md`) against the organization surface instead of plugins.

- **Search**: `gh api search/code` (GitHub code search, authenticated) with 10 query variants targeting beets-config markers (`"paths:" "default:"`, `"asciify_paths"`, `"item_fields:"`, `"directory:" "plugins:"`, `"%aunique"`, `"replace:" "paths:"`, `"art_filename"`, `"max_filename_length"`, all `filename:config.yaml`). 900 raw hits, 714 unique file hits across 393 unique repos.
- **Dedup**: one file per repo (first hit). Downloaded 393 raw files via `raw.githubusercontent.com`.
- **Beets-fingerprint filter**: kept files containing at least one hard beets marker (`fetchart`, `musicbrainz`, `lastgenre`, `chroma`, `embedart`, `beetsdir`, `$albumartist`, `asciify_paths`, `%aunique`, `ftintitle`, `mbsync`) → 133 candidates.
- **False-positive sweep**: cross-checked each candidate for `library:`/`directory:`/`plugins:`/`include:` (real beets top-level keys). Found and dropped 4 contaminants — generic YAML configs (a RAG pipeline, an AI-agent config, a settings file) that matched only on the ambiguous word "chroma" as a substring. One (`cupkax/beets-config`) verified legitimate despite lacking those keys (it uses `include:` to split plugins into a separate file).
- **Final filter**: required a genuine top-level `paths:` block (not a nested nested/unrelated `paths:` key elsewhere in the file) — this is the actual denominator for organization-surface analysis.
- **n = 124** usable configs, **120 unique repo owners** (4 owners contributed 2 configs each). Regex-based parser (no PyYAML available in the sandbox); spot-checked against raw file content throughout.

**Sampling bias, same caveat as the prior survey**: 55.6% of the sample are repos literally named `dotfiles`/`dots` — power users who publish their whole config. Percentages here describe **what organization behavior power users reach for**, not casual out-of-box usage; the *ranking* is the durable signal, not the absolute share. Cross-check against the prior survey: `%the`/"the" plugin usage here (6.5% of paths, 6.5% of plugin lists) lands close to that survey's 9.2% "the" plugin figure, and inline-Python usage here (26.6% via item_fields/album_fields) is close to that survey's 21.8% `inline` plugin figure — consistent methodology, consistent population.

## (b) Organization features ranked by share of configs

| Feature | Share (n=124) |
|---|---|
| Explicit import file-op mode (move/copy/link/hardlink/reflink set) | 89.5% |
| `import:` block present at all | 91.1% |
| Query-prefixed paths present (any non-`default` key) | 89.5% |
| `comp:` path (compilations/Various Artists) | 79.8% |
| `singleton:` path (non-album tracks) | 77.4% |
| `per_disc_numbering` set | 35.5% |
| `albumtype:X` path (e.g. soundtrack) | 33.9% |
| `incremental` import toggle set | 39.5% |
| `timid` import toggle set | 34.7% |
| `replace:` overridden | 30.6% |
| `quiet_fallback` import toggle set | 27.4% |
| `item_fields`/`album_fields` (inline Python) present | 26.6% |
| `art_filename` set | 25.0% |
| `autotag` import toggle set | 20.2% |
| `asciify_paths` set | 21.8% (12.9% explicitly `yes`) |
| `original_date: yes` | 18.5% |
| `duplicate_action` set | 16.9% |
| `max_filename_length` set | 10.5% |
| Custom flexible-attribute query prefix (e.g. `playlist:X:`) | 7.3% |
| `va_name` customized | 7.3% |
| `ext:X` art/file-extension routing under `paths:` | 4.0% |
| `genre:X` per-genre path | 2.4% |
| `artist_credit` field used in paths | 1.6% |
| `$albumartist_sort`/`$artist_sort` used in paths | 0% |

## (c) Template fields and functions ranked (the language core)

**Fields** (share of configs whose `paths:` block references `$field`):

| Field | Share | Field | Share |
|---|---|---|---|
| `$title` | 97.6% | `$disc` | 21.8% |
| `$album` | 95.2% | `$original_year` | 11.3% |
| `$albumartist` | 92.7% | `$multidisc` (custom, derived) | 11.3% |
| `$track` | 88.7% | `$genre` | 7.3% |
| `$artist` | 73.4% | `$label`, `$format`, `$albumpath` | 2–4% each |
| `$year` | 30.6% | everything else (30+ fields) | ≤1 config each |

Long tail past `$year`/`$disc` is almost entirely one-off derived fields from `item_fields:` (see (e)) — not core-language candidates, seam candidates instead.

**Functions**:

| Function | Share | Function | Share |
|---|---|---|---|
| `%aunique{}` | 74.2% | `%lower{}` | 6.5% |
| `%if{}` | 25.0% | `%title{}`, `%ifdef{}` | 2–3% |
| `%the{}` | 12.9% | `%substitute{}`, `%sunique{}` | 1.6% |
| `%asciify{}` | 8.1% | `%left{}`, custom plugin funcs | 0.8% |

`%aunique` is functionally load-bearing (disambiguates same-name albums) and is the single highest-priority function after the raw field substitution engine itself. `%if` is the only other function above 20%. Everything past `%the`/`%asciify`/`%lower` is long tail.

**Query-prefix mechanism** (the part of the template language that routes different releases through different path templates): `comp:` and `singleton:` are near-universal (77–80%) and effectively **not optional** — any beets-parity organizer needs a first-class "special-case path by query" mechanism, not just a single template. `albumtype:X` (34%) is the next tier. Genre/ext/custom-flex-attribute prefixes are long tail but prove the mechanism must be a **general flexible-attribute query router**, not an enum of known cases (see `playlist:AnimeSongs:` example below).

## (d) Proposed 80% cut line

**In the 80% core** (build as first-class, not configurable escape hatches):
1. Template engine with fields: `$albumartist`, `$album`, `$artist`, `$track`, `$title`, `$year`, `$disc`, plus a computed multi-disc-aware track/disc formatter as a *built-in*, not requiring inline code (see (e) — this alone would obsolete most of the 26.6% inline-Python usage).
2. `%aunique{}` (near-mandatory disambiguation) and `%if{}`.
3. Query-prefixed path routing with at minimum `comp:`/`singleton:` as built-in special cases, plus a general "route by any field or flexible attribute" mechanism (covers `albumtype:`, `genre:`, custom attrs, and the `ext:` art-routing pattern in one primitive instead of four).
4. Import file-op mode: move and copy at minimum (91% cover), with link/hardlink/reflink as cheap additions since they're the same code path family.
5. `per_disc_numbering`, `asciify_paths`, `art_filename`, basic `replace:` override capability (even a fixed default ruleset plus override, not full customizability) — all in the 20–36% band, each cheap and independently valuable.
6. Import toggles: `incremental`, `timid`, `quiet_fallback`, `duplicate_action`, `autotag` (17–40% each) — these are workflow/UX toggles, low implementation cost relative to value.

**Outside the cut, deliberately**: `%the`, `%asciify`, `%lower` as user-composable functions rather than built into every default (8–13% each — nice-to-have, not core); `max_filename_length` (10.5%, cheap but low-value, can be a fixed sane default); `va_name` customization (7.3%); rare functions (`%substitute`, `%sunique`, `%left`, `%ifdef`); rare fields past `$genre`/`$label`/`$format`. This tier covers roughly the next 10–15% and is reasonable to ship later or leave as config knobs rather than day-one features.

**Firmly in the tail** (see (e) for the seam argument): arbitrary `item_fields`/`album_fields` inline-Python computation (26.6% of configs use *some* version of this, but each does something different); one-off custom flexible-attribute query prefixes beyond the common ones; `$albumartist_sort`/`$artist_sort` in paths (0% observed, despite existing in beets — dead weight).

## (e) Long-tail patterns worth a seam, not a feature

**Inline field computation (`item_fields`/`album_fields`) — the standout, 26.6% of configs (33/124).** These are Python snippets returning a derived value usable as `$fieldname` in the path template. Categorized the 33 configs' snippets:

- **Multi-disc formatting — 27 of 33 (82% of the inline-Python users, 21.8% of the whole sample)**: overwhelmingly the same idiom, `1 if disctotal > 1 else 0` (or a zero-padded `disc.track` string). This is the single most valuable built-in dubplate can ship: a native multi-disc-aware formatter would eliminate the majority of why people reach for inline Python at all.
- **Artist-name normalization — 5 configs**: extracting a "primary" artist from a multi-artist/feat. string (`leadartist`, `first_artist`, `artist_fallback` chains, regex splits on `feat.`/`&`/`,`). Genuinely idiosyncratic per user (different split rules, different fallback order) — a good seam candidate (a "primary artist" derivation hook with a default heuristic, overridable), not a fixed feature.
- **Quality/format labels — 7 configs**: aggregating format/bitrate across an album's items (`custom_media_format`, `albumformat`, summed `bitrate`). Needs access to the whole album's item list, not just one field — this is architecturally a lifecycle hook (runs after tagging, before path render, with album-scope access), not a template function.
- **Other (2 configs)**: country-code lookup via `pycountry`, and year-prefixed title bucketing (`aybum`/`tytle`: `f"[{original_year}] {album}"`). Genuinely one-off; not worth designing for beyond "users can write a hook."

**Recommendation**: expose one lifecycle hook point ("derive additional fields before path rendering, with album-scope access") rather than trying to enumerate use cases. Ship the multi-disc formatter as a real built-in since it alone covers 82% of the observed need.

**Custom flexible-attribute query-prefixed paths — 7.3% (9/124), but proves the shape of the mechanism.** Example from `mudssky/powershellScripts`, a config with 17 distinct path rules keyed on a custom flexible attribute:
```yaml
paths:
    playlist:AnimeSongs: AnimeSongs/$title%if{$artist, - $artist,}
    playlist:GalgameSongs: GalgameSongs/$title%if{$artist, - $artist,}
    ... (15 more playlist:X variants)
```
This is why the query-router in (d) needs to be a general "any field, any value, arbitrary count of rules" primitive rather than an enum of `comp`/`singleton`/`albumtype`/`genre` — real users invent their own routing keys.

## (f) Raw tally data (compact, for later commit)

```
n=124 unique-repo beets configs with a genuine top-level paths: block
(120 unique owners; 55.6% dotfiles-named repos — power-user sampling bias, see (a))

FIELDS (share of configs, $field in paths:):
title 97.6 | album 95.2 | albumartist 92.7 | track 88.7 | artist 73.4 |
year 30.6 | disc 21.8 | original_year 11.3 | multidisc(derived) 11.3 |
genre 7.3 | disc_and_track 4.0 | albumpath 4.0 | label 2.4 | format 1.6 |
artist_credit 1.6 | atypes 1.6 | (30+ fields at 1 config each)

FUNCTIONS (share of configs, %func{} in paths:):
aunique 74.2 | if 25.0 | the 12.9 | asciify 8.1 | lower 6.5 | title 3.2 |
ifdef 2.4 | substitute 1.6 | sunique 1.6 | left 0.8

QUERY-PREFIXED PATHS (per-config presence):
comp: 79.8 | singleton: 77.4 | albumtype:X 33.9 | other custom flex-attr 7.3 |
ext:X 4.0 | genre:X 2.4

SCALARS:
replace: overridden 30.6 (rule count 1-24, median ~8-10) |
asciify_paths set 21.8 (yes 12.9) | max_filename_length set 10.5 (values: 0,150,155,156,200,255) |
art_filename set 25.0 (values: cover 14, albumart 7, folder/Folder 6, custom 4) |
per_disc_numbering set 35.5 | original_date:yes 18.5 | va_name customized 7.3 |
artist_credit field 1.6 | sort-name in paths 0.0 | "the" plugin present 6.5

IMPORT:
any file-op mode explicit 89.5 (move=yes 46.0, copy=yes 32.3, link/hardlink/reflink <5 each) |
import: block present 91.1 | incremental set 39.5 | timid set 34.7 |
quiet_fallback set 27.4 | autotag set 20.2 | duplicate_action set 16.9

INLINE FIELD COMPUTATION (item_fields/album_fields):
present 26.6 (33/124) -> multidisc formatting 82% of those (27/33, 21.8% of whole sample) |
artist-name normalization 15% of those (5/33) | quality/format labels 21% of those (7/33) |
other/one-off 6% of those (2/33)
```

Full machine-readable tally (124 records, all fields/funcs/keys/import settings per config) is at `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/889ce70b-0ec4-4183-b205-fc124b2a9239/scratchpad/beets/tally_results.json` (scratchpad, not committed — copy into `docs/superpowers/research/` if you want it as the companion artifact to the plugin survey, analogous to `2026-08-31-beets-plugin-tally.json`). Parser and query scripts are alongside it in the same directory (`tally.py`, `categorize.py`, `summarize.py`) if you want to rerun or extend the sample later.
