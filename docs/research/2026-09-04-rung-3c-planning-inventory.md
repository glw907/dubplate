# Rung 3, Pass 3c Planning Inventory

Research report, 2026-09-04, prepared for pass 3c's plan authoring,
against HEAD `d518e572d0bbd2409bc1e20b037ba0e3f8aa2e74`. Scope, quoted
from the rung 3 spec's pass-split table
(`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md:330`):

> **3c** Client and limiter | The WS/2 client with its retry, timeout,
> and User-Agent behavior; the Lucene query construction and the
> `inc=` sets; the shared limiter and the per-caller budget interface;
> request-construction grading against the committed request logs; the
> stub's `Extra` replay map and `Consumed()` list (B2), against the
> cassette schema half or its written fallback. | 3a for the
> likely-tags helper the query is built from; rung 2's request logs
> and stub; the `Case.Extra` schema change, or the recorded decision
> not to make it. | Main tree: writes `internal/harness/conform` and
> `internal/harness/stub`.

Three sections follow: (a) beets' client and query mechanics, (b) the
harness, stub, cassette, and request logs as landed, (c) the engine as
landed and the proposed 3c Go surface. A fourth section, (d),
reconciles all three and lists the open questions each raised.

---

## (a) Beets client and query mechanics

Every citation is `path:line` against this tree. Pinned beets lives under
`/var/home/glw907/Projects/dubplate/tools/venv/.venv/lib/python3.13/site-packages/`, abbreviated
**SP** below. Corpus figures are read off the tree as committed at HEAD `d518e57` (215 cases).

### 0. Corrections to the port map before anything is built

Three claims the plan should not inherit unchanged.

1. **`musicbrainzngs` is not installed and is not used.** There is no `musicbrainzngs` package under
   `SP/`; the only trace of the name is a stale docstring at `SP/beetsplug/musicbrainz.py:271`. The client
   is beets' own in-tree `MusicBrainzAPI` over `requests` (`SP/beetsplug/_utils/musicbrainz.py:564-783`).
   There is no `set_useragent` call anywhere; the User-Agent is set once on the session at
   `SP/beetsplug/_utils/requests.py:76`. The mechanics study already says this
   (`docs/superpowers/research/2026-09-02-beets-matcher-mechanics.md:477-481`); the task brief's framing
   does not, so the plan should state it so no task goes hunting for a wrapper.

2. **The legacy `search_limit` spelling is `musicbrainz.match.searchlimit`, not `match.searchlimit`
   and not `musicbrainz.searchlimit`.** The shim reads `self.config["match"]["searchlimit"]`
   (`SP/beetsplug/musicbrainz.py:292-294`), and `self.config` is `beets.config[self.name]`
   (`SP/beets/plugins.py:200`), so the key resolves one level under `musicbrainz:`. Verified by probe
   against the live confuse view: setting `musicbrainz.match.searchlimit: 9` resolves, a top-level
   `match.searchlimit: 9` does not resolve at all, and `musicbrainz.searchlimit` is a different key this
   shim never reads. This contradicts the rung 3 spec's tunable enumeration
   (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`, "the tunable surface, enumerated"), 3a's
   HISTORY correction #5 (`docs/HISTORY.md:597-599`), and the shipped doc comment at
   `internal/engine/matcher/tunables.go:119-124`. 3c owns the fix because 3c owns the search surface; the
   losing documents should be corrected in the same pass, on 3a's own M2 precedent.

3. **`search_query_ascii` is a real config key in the search path that the study omits.**
   `SP/beets/metadata_plugins.py:363-365` adds it (default `False`), and `_search_api` applies
   `unidecode.unidecode(query)` to the query *text* at `:412-413`. For MusicBrainz the query text is
   always the empty string (`SP/beetsplug/musicbrainz.py:751`), so the key is inert by construction. It
   still belongs in the enumeration as a knob deliberately carrying no behavior, with the reason on it.

### 1. How a request is issued

The composition, outermost first, from `SP/beetsplug/_utils/musicbrainz.py` and
`SP/beetsplug/_utils/requests.py`.

| Layer | Where | What it does |
| --- | --- | --- |
| `MusicBrainzAPI.request` | `_utils/musicbrainz.py:615-619` | `kwargs.setdefault("params", {})` then `params["fmt"] = "json"`. Every request, no exception. |
| `RequestHandler.request` | `_utils/requests.py:204-211` | Wraps the call in `handle_http_error`. |
| `TimeoutAndRetrySession.request` | `_utils/requests.py:93-103` | `kwargs.setdefault("timeout", 10)`, then `r.raise_for_status()`. |
| `LimiterMixin.send` | `SP/requests_ratelimiter/requests_ratelimiter.py:96-106` | Blocking token acquire at `per_second=self.rate_limit`, `burst=1`, `max_delay=None` so the wait is unbounded. A 429 response refills the bucket (`limit_statuses=(429,)`). |
| `RateLimitAdapter.send` | `_utils/requests.py:128-135` | Process-wide hard floor of 0.25 s between any two requests, under a lock. |
| urllib3 `Retry` | `_utils/requests.py:78-91` | `total=6`, `backoff_factor=0.5`, `status_forcelist=[500, 502, 503, 504, 429]`. |

Load-bearing details the port must get right.

- **Retries bypass both rate limiters.** `Retry` lives inside the adapter's own `urlopen` loop, so a
  retried request re-enters neither the token bucket nor the 0.25 s floor. Only the outermost attempt is
  metered.
- **Backoff schedule.** `SP/urllib3/util/retry.py:309-326`: zero for the first retry, then
  `0.5 * 2**(n-1)`, capped by `backoff_max=120`, jitter `0.0`. Concretely 0, 1, 2, 4, 8, 16 s.
  `respect_retry_after_header` is `True` (`SP/urllib3/util/retry.py:373-385`), so a `Retry-After` header
  wins over the computed backoff.
- **Exhausted retries do not raise an `HTTPError`.** `raise_on_status` defaults `True`, so urllib3 raises
  `MaxRetryError`, which requests converts to `requests.exceptions.RetryError`. That is not an
  `HTTPError`, so `handle_http_error` (`_utils/requests.py:188-202`) never maps it and neither
  `HTTPNotFoundError` nor `UnauthorizedMBError` applies.
- **Session is a process-wide singleton** (`_utils/requests.py:44-63`), registered for `atexit` close. The
  0.25 s floor is therefore shared across every caller in the process, which is the closest thing pinned
  beets has to the per-caller budget 3c is asked to build.
- **User-Agent** is `beets/2.13.1 https://beets.io/`, verified by probe against the installed session
  (`_utils/requests.py:76`). The harness's own live capture tool already uses a different, identified UA,
  `dubplate-harness/<version> ( <contact> )` (`cmd/harness/capture.go:132-134`), gated on
  `DUBPLATE_HARNESS_CONTACT` (`:26`, `:116-128`). 3c's product client needs its own third UA and its own
  contact input; it must not reuse either literal.
- **No response caching anywhere in this path**, and pagination exists only in the browse fallback.

### 2. Host, scheme, and rate config

`MusicBrainzAPI.__post_init__` (`_utils/musicbrainz.py:586-606`) adds four defaults onto
`config["musicbrainz"]`: `host: "musicbrainz.org"`, `https: False`, `ratelimit: 1`,
`ratelimit_interval: 1`. The branch matters.

- `host == "musicbrainz.org"` hard-codes `https://musicbrainz.org` and `rate_limit = 1.0`, ignoring
  `https`, `ratelimit`, and `ratelimit_interval` entirely (`:598-599`).
- Any other host honors all three: `api_host = f"http{'s' if https else ''}://{hostname}"` and
  `rate_limit = ratelimit / ratelimit_interval`, where `ratelimit` is read `.get(int)` and
  `ratelimit_interval` `.as_number()` (`:600-606`).
- `api_root` is `f"{api_host}/ws/2"` (`:608-610`).

The corpus's own derivation seam uses exactly this branch. No corpus profile writes a `musicbrainz:`
block (`corpus/profiles/{estate,locale,stock}.yaml`); the Go adapter appends one at derivation time,
`host: <stub addr>`, `https: no`, `ratelimit: 100` (`internal/harness/oracle/oracle.go:247`), and all four
keys are excluded from the config hash as plumbing (`internal/harness/oracle/derive_oracle.py:41-49`). So
`rate_limit` during derivation is 100/s and the 0.25 s adapter floor is the binding constraint, which is
where the oracle suite's 983 s wall time (`docs/HISTORY.md:325-330`) comes from.

### 3. Endpoints and `inc=` sets

Only two endpoint shapes appear in the committed corpus. Census over all 215 cases, method and
`path_query` only:

| Shape | Count |
| --- | --- |
| `GET /ws/2/release/{mbid}?fmt=json&inc=<RELEASE_INCLUDES>` | 303 |
| `GET /ws/2/release?fmt=json&limit=5&query=<lucene>` | 48 |
| Total logged requests | 351, all `GET` |

**Release lookup.** `get_release` defaults `includes=RELEASE_INCLUDES`
(`SP/beetsplug/_utils/musicbrainz.py:687-690`), the sixteen-element list at `:50-67`, joined with a
literal `"+"` at `:634-637`. On the wire the `+` is percent-encoded to `%2B` because requests encodes
params with `quote_plus`. The corpus confirms this verbatim, and `internal/harness/mbrequest/mbrequest.go:22-25`
already carries the exact same string as `MaximalReleaseInclude`, in beets' order, byte-for-byte.

**Recording lookup.** `get_recording` defaults `RECORDING_INCLUDES`, five elements
(`_utils/musicbrainz.py:69-75`, `:692-697`). **Zero committed cases issue one.** The singleton path is
entirely unexercised by the corpus.

**Search.** `search` passes no `includes` at all (`_utils/musicbrainz.py:665-685`), so a search request
carries no `inc=`. Confirmed against all 48 committed search rows.

**Browse.** `browse_recordings(release=…, limit=100, includes=BROWSE_INCLUDES, offset=i)` fires only when
a release's summed `media[].tracks` exceeds 500 (`SP/beetsplug/musicbrainz.py:602-629`, constants at
`:58-67`). Zero committed cases reach it; the corpus maximum is 470 tracks and `conform`'s guard already
watches the line with a deliberate `>=` (`docs/HISTORY.md:295-301`). 3c owns the port if a case ever
crosses.

**Parameter emission order** (not graded, but worth recording once so nobody re-derives it): lookup emits
`inc`, then `fmt`; search emits `limit`, `query`, then `fmt`; browse would emit `release`, `limit`,
`offset`, `inc`, `fmt`. Order is irrelevant at the grading boundary because both `cassette` and `stub`
canonicalize by sorting the raw pairs byte-wise (`internal/harness/cassette/cassette.go:214-228`,
`internal/harness/stub/stub.go:128-135`). The Go client may emit in any order.

**Percent-encoding is already byte-compatible.** Python's `quote_plus` and Go's
`url.Values.Encode`/`url.QueryEscape` leave the same unreserved set (`A-Za-z0-9-_.~`) and map space to
`+`. Probed both sides over the full ASCII punctuation range and they agree character for character. Go's
`net/url` is the correct encoder with no adapter.

### 4. Lucene query construction

Three stages, each in a different file.

**Stage 1, criteria.** `get_album_criteria` (`SP/beetsplug/musicbrainz.py:714-734`) builds an
insertion-ordered dict: `{"release": album}` first, then either `{"arid": VARIOUS_ARTISTS_ID}` when
`va_likely` or `{"artist": artist}`. `VARIOUS_ARTISTS_ID` is the literal
`89ad4ac3-39f7-470e-963a-56509c546377` (`:43`). Then one entry per configured extra tag, in the order the
operator listed them in `musicbrainz.extra_tags`, mapped through `FIELDS_TO_MB_KEYS` (`:47-56`) and
filtered to known tags by `extra_mb_field_by_tag` (`:697-712`). Three value rules inside that loop
(`:721-732`): `tracks` becomes `str(len(items))`; `alias` becomes the album name; everything else becomes
`str(util.plurality(i.get(tag) for i in items)[0])`, with `catalognum` additionally having every space
removed. For a singleton the criteria are the fixed `{"artist": artist, "recording": name, "alias": name}`
(`:749`).

**Stage 2, pre-filter.** `get_search_query_with_filters` returns `("", {k: v.lower().strip() for … if
truthy})` (`:736-753`). It drops empty-after-strip values and lowercases once. The query *text* half of
the tuple is always the empty string for MusicBrainz.

**Stage 3, rendering.** `MusicBrainzAPI.search` space-joins
`filter(None, starmap(format_search_term, filters.items()))` (`_utils/musicbrainz.py:679-681`).
`format_search_term` (`:650-663`) lowercases and strips a second time, returns `""` on an empty result,
substitutes `LUCENE_SPECIAL_CHAR_PAT` with a backslash prefix, then wraps as `field:(term)` when the field
is non-empty.

`LUCENE_SPECIAL_CHAR_PAT` is `([-+&|!(){}[\]^"~*?:\\/])` (`_utils/musicbrainz.py:48`). Nineteen
characters: `- + & | ! ( ) { } [ ] ^ " ~ * ? : \ /`. Notably absent, and therefore *not* escaped:
`' , . ; < > = # % @ $`. Probed directly against the pinned function; results confirm the corpus.

Two grading traps this produces, both visible in the committed record.

- Every hyphen in an `arid` value is backslash-escaped, so the VA id ships as
  `arid:(89ad4ac3\-39f7\-470e\-963a\-56509c546377)`, on the wire
  `arid%3A%2889ad4ac3%5C-39f7%5C-470e%5C-963a%5C-56509c546377%29`. A port that skipped escaping because
  the id "is just hex" would miss the stub on all five committed `arid` rows.
- The escaping runs before the field wrapper, so the wrapping `:` and `()` are never escaped, while a `:`
  or `(` inside a term is. Committed evidence:
  `release%3A%28frank+zappa%5C%3A+cucamonga%29` and
  `release%3A%28symphony+no.+9%2C+%5C%22choral%5C%22%29`.

**Two Python-versus-Go divergences in the normalize step, both real, both unreachable today.**

- `str.lower()` is full Unicode lowercasing with special casing; Go's `strings.ToLower` is per-rune simple
  case mapping. Probed: `"İstanbul"` (U+0130) lowercases to `i` + U+0307 in Python and to a single `i` in
  Go. `ẞ`, `Ω`, and `ǅ` agree in both.
- `str.strip()` treats the C0 separators `\x1c`-`\x1f` as whitespace; Go's `strings.TrimSpace` does not.
  Probed: `"A\x1c\x85 "` strips to `"A"` in Python and to `"A\x1c"` in Go.

Neither fires on the committed corpus, which is ASCII-or-clean throughout. Both belong beside 3b's
guard 11 (`docs/HISTORY.md:279-284`) as members of one family: Python regex and string semantics against
Go RE2 and simple case mapping. A third member of the same family is `_extract_id`'s MBID pattern
`(\w{8}(?:-\w{4}){3}-\w{12})` (`SP/beets/util/id_extractors.py:20`, applied at
`SP/beets/metadata_plugins.py:272-278`), where Python's `\w` is Unicode-aware and Go's is ASCII-only.

**Corpus coverage of query construction.** All 48 committed search rows exercise exactly two filter
shapes: `release` + `artist` (43 rows, 43 distinct queries) and `release` + `arid` (5 rows, 4 distinct
queries). **No committed case sets `extra_tags`**, because no profile writes a `musicbrainz:` block at
all, so the whole `FIELDS_TO_MB_KEYS` branch, the `tracks` and `alias` special cases, and the
`catalognum` space-strip are corpus-unreachable. They join `match.required`, `match.ignored`, and
`preferred.original_year` on the M8 list and must be graded by hand-authored table-driven tests in the
package that owns them, with the mutation that proved each can fire named in the report.

One trap inside that unreachable branch, worth a test the day it is written: `str(most_common)` on a beets
`year` field with no value yields the string `"0"`, which survives the truthiness filter at
`SP/beetsplug/musicbrainz.py:752` and ships as `date:(0)`.

### 5. `va_likely` and the likely-tags inbound

`va_likely` is computed in `tag_album` at `SP/beets/autotag/match.py:451-455` as the disjunction of three
predicates: no artist consensus, `search_artist.lower() in VA_ARTISTS`, or any item flagged `comp`.
`VA_ARTISTS` is `("", "various artists", "various", "va", "unknown")`
(`SP/beets/autotag/distance.py:29`).

`search_artist` is `likelies["artist"]` (`match.py:445-447`), which `get_most_common_tags` may have
overwritten with `likelies["albumartist"]` when the album-artist value is unanimous and non-empty
(`SP/beets/util/__init__.py:842-844`). `consensus["artist"]`, by contrast, is computed from the raw
per-item `artist` values and is never updated by that override. The Go side already reproduces both halves
in `MostCommonTags` (`internal/engine/matcher/likelies.go:47-83`, override at `:78-80`).

All three triggers are corpus-reachable, one case each:

| Case | Trigger |
| --- | --- |
| `real-noise-zappa-cucamonga` | nine distinct item artists, so `consensus["artist"]` is false |
| `nonva-va-artists-misflag` | artist `"Unknown"`, a `VA_ARTISTS` member |
| `va-track-mismatch-all` | `comp: true` on every item, and albumartist `"Various Artists"` |

**Two concrete inbound gaps 3c must close before it can build the query.**

1. **`ItemFields` has no `Comp` field.** `internal/engine/matcher/info.go:124-143` carries eighteen fields
   and `comp` is not among them, and `ProjectItemFields`'s field-name constants at `:147-166` do not
   include it either. The corpus does carry the data: 15 cases set `comp` on every item (item counts 3, 3,
   3, 3, 12, 50, 50, 50, 50, 70, 70, 100, 100, 244, 244). So the field must be added to `ItemFields` and to
   the decoder, which is `internal/engine/matcher` territory rather than `internal/engine/musicbrainz`.
   The plan should name it as a first task and say plainly that 3c writes one matcher file.
2. **`vaArtists` is unexported.** `internal/engine/matcher/track_artist.go:21` declares
   `var vaArtists = []string{…}` in package `matcher`. `internal/engine/musicbrainz` cannot read it. Export
   it (one list, one home) rather than duplicating it; a second copy is exactly the `releaseLookupPrefix`
   triplication 3b already flagged as worth revisiting (`docs/HISTORY.md:352-357`).

### 6. `search_limit`, and where the number comes from

`MetadataSourcePlugin.__init__` adds `search_limit: 5` and
`data_source_mismatch_penalty: DEFAULT_DATA_SOURCE_MISMATCH_PENALTY`
(`SP/beets/metadata_plugins.py:197-204`, class constant `0.5` at `:180`). `_search_api` reads
`self.config["search_limit"].get(int)` and packs it into `SearchParams`
(`SP/beets/metadata_plugins.py:415-416`, the tuple at `:335-345`). `get_search_response` forwards it as
`limit=params.limit` (`SP/beetsplug/musicbrainz.py:755-763`).

There is no `musicbrainz:` block in `SP/beets/config_default.yaml` at all; every MusicBrainz key comes
from a `config.add` call in Python. All 48 committed search rows carry `limit=5`, so the corpus pins the
default and exercises no override.

3a's carry on `dataSourcePenalty` is confirmed and unchanged: `get_penalty`
(`SP/beets/metadata_plugins.py:160-169`) falls back to the *class literal*
`MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY`, while the Go port falls back to the
configured `Tunables.DataSourceMismatchPenalty`. They differ only when `source_weight` or
`data_source_mismatch_penalty` is set, and no profile sets either. The resolution property that reads
`source_weight` first is at `SP/beets/metadata_plugins.py:190-195`.

### 7. The request set one import issues, and in what order

`tag_album` (`SP/beets/autotag/match.py:383-468`), single-source case.

1. `get_most_common_tags(items)` (`:407`). No request.
2. If `search_ids` was passed explicitly, only `albums_for_ids(search_ids)` runs and the text search is
   skipped entirely (`:417-420`).
3. Otherwise `match_by_id(likelies["mb_albumid"], consensus["mb_albumid"])` (`:425-428`,
   `match_by_id` at `:252-266`). With an id and consensus this issues one release lookup.
   `album_for_id` (`SP/beetsplug/musicbrainz.py:765-800`) runs `_extract_id`, then `get_release`, then, if
   `original_release["status"] == "Pseudo-Release"` and a backward `transl-tracklisting` release relation
   exists, a **second** release lookup of the linked actual release (`:785-798`).
4. `_recommendation` over the id-derived candidates (`:430`). If they exist, `import.timid` is off, and
   the recommendation is `strong`, `tag_album` **returns here and issues no search** (`:432-442`).
5. Otherwise one search request, then one release lookup per returned id in search-response order
   (`SP/beets/metadata_plugins.py:442-446`, `:250-259`). The generator is fully materialized inside a
   worker thread (`:82-86`), so ordering is preserved.
6. Sort and recommend (`:466-467`). No further requests.

`tag_item` mirrors this for singletons with a track-ID-first fast path
(`SP/beets/autotag/match.py:471-526`), reaching `/ws/2/recording`. Corpus-unreachable.

**What the corpus actually exercises.** Every one of the 215 cases has all its items carrying the same
single non-empty `mb_albumid`, verified across the whole tree. So `likelies["mb_albumid"]` is always
non-empty and `consensus["mb_albumid"]` is always true, and step 3 always runs. Per-case shape:

| Searches | Lookups | Cases | Reading |
| --- | --- | --- | --- |
| 0 | 1 | 164 | id match reached `strong`, early return at `match.py:436-442` |
| 0 | 2 | 3 | same, plus the pseudo-release second fetch |
| 1 | 1 | 26 | id match below `strong`, search fell through |
| 1 | 2 | 4 | as above |
| 1 | 5 | 9 | as above, search returned five ids |
| 1 | 6 | 9 | as above, id candidate distinct from all five search results |

Three named coverage gaps follow, and the plan should declare them rather than let a later pass find
them: the **explicit `search_ids`** branch is unreached, the **no-id and no-consensus** branch of
`match_by_id` is unreached, and the **whole singleton path** including `RECORDING_INCLUDES` is unreached.

**A scope seam the pass split does not name.** Tier 1's request-*set* expectation is a function of the
ID-match early return, and the ID-match early return, `_recommendation`, and `_add_candidate` are all 3d
scope (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`, the pass-split table). 3c can grade
request *construction* (given inputs, is the path plus query byte-identical to the committed row) but
cannot grade the request *set* without 3d's flow. The plan should say which of the two 3c actually
delivers and hand the other forward explicitly.

**The request log is a set, not a sequence.** `TestEveryLoggedRequestResolvesToExactlyOneResponse`
(`internal/harness/corpus_test.go:105-118`) forbids two rows sharing a canonical `(method, path_query)`
key, and `stub.New` keys its map on exactly that pair (`internal/harness/stub/stub.go:55-62`). The log
therefore cannot express "issued twice", which matters because beets legitimately re-issues the same
lookup when the id candidate also appears in the search results; `_add_candidate` drops the duplicate
afterward at `SP/beets/autotag/match.py:349-352`. A grading rule that counted issuances would be wrong.

### 8. What is caught and what propagates

| Condition | Behavior | Citation |
| --- | --- | --- |
| 404 on release lookup | `HTTPNotFoundError` caught, `album_for_id` returns `None`, candidate silently dropped | `SP/beetsplug/musicbrainz.py:777-781` |
| 404 on recording lookup | `with suppress(HTTPNotFoundError)`, returns `None` | `SP/beetsplug/musicbrainz.py:810-813` |
| 404 on search | propagates out of `search`, then caught by `_search_api`'s bare `except Exception`, returns `()` | `SP/beets/metadata_plugins.py:419-427` |
| 401 anywhere | `UnauthorizedMBError` with the "check your musicbrainz.user and musicbrainz.pass" message | `SP/beetsplug/_utils/musicbrainz.py:551-561`, `:578-581` |
| Other non-2xx | `raise_for_status` raises, `handle_http_error` re-raises unchanged | `SP/beetsplug/_utils/requests.py:101`, `:196-202` |
| 429/500/502/503/504 | six urllib3 retries, then `requests.exceptions.RetryError` (**not** an `HTTPError`) | `SP/beetsplug/_utils/requests.py:78-89` |
| Anything escaping a plugin call | `maybe_handle_plugin_error` logs and swallows it, dropping that plugin's whole candidate list for that call | `SP/beets/metadata_plugins.py:62-74`, `:105-108` |
| `raise_on_error: yes` | both catch-alls re-raise instead | `SP/beets/config_default.yaml:14` (default `no`), `metadata_plugins.py:65`, `:422` |

The consequence for the conformance gate is worth stating in the plan. A stub 404 during a lookup drops a
candidate silently, and a stub 404 during a search yields an empty result list silently. Neither raises.
The zero-miss rule is the only thing that turns either into a visible failure, which is exactly why tier 1
fails on a non-empty miss list.

One latent hazard: `original_release["status"]` at `SP/beetsplug/musicbrainz.py:785` is a direct index,
not a `.get`, so a release body with no `status` key raises `KeyError` inside `album_for_id` and is
swallowed by `maybe_handle_plugin_error`. The Go port should decide deliberately whether an absent status
is an error or a plain non-pseudo release, and say which in a doc comment.

### 9. Metadata-source-set dependencies

Two behaviors depend on which sources are loaded.

- **Candidate iteration order.** With one source, `_yield_from_plugins`'s `as_completed` loop yields one
  future and order is exactly the search response's order (`SP/beets/metadata_plugins.py:91-108`). With
  two or more, the interleaving is thread-timing nondeterministic. `stock.yaml`, `estate.yaml`, and
  `locale.yaml` all pin `plugins: musicbrainz`, so the corpus is single-source by construction and the
  port may assume it. `Tunables.DataSourcePenalties` already models the loaded set
  (`internal/engine/matcher/tunables.go:138-149`).
- **`get_penalty`'s scan** over `find_metadata_source_plugins()` (`SP/beets/metadata_plugins.py:160-169`),
  which 3a already ported behind `metadataSourceSet` (`docs/HISTORY.md:604-611`).

### 10. The `Extra` replay channel and the cassette schema half

State of the tree, so the first task is not re-derived.

- `Case.Extra` is `[]string` of response filenames (`internal/harness/cassette/cassette.go:47`).
- `validate` rejects an `Extra` entry naming a file a `Request` already consumes, and proves each is
  readable (`cassette.go:169-176`).
- `stub.New` loads only `c.Requests` into its replay map (`stub.go:55-62`); anything else records a `Miss`
  and answers 404 (`stub.go:81-91`). There is no hit tracking of any kind, only `Misses()` (`:99-105`).
- **Zero committed cases carry an `Extra` entry**, verified over all 215. The migration is empty, exactly
  as the spec's B2 fold predicts.

Consumers a shape change touches, all four of them:

1. `cassette.go:47` (the field), `:169-176` (validation), and the package doc at `:7-8`.
2. `cassette_test.go:181, 198, 218, 238, 255` (five sites; `:238` is the already-consumed negative test).
3. **`internal/harness/corpus_test.go:127-151`**, `TestEveryResponseFileIsConsumed`, which iterates
   `c.Extra` as filenames at `:133-135`. This file is in package `harness` and is **not** listed in 3c's
   tree column, which names only `internal/harness/{conform,stub}`. The plan needs a written owner for it.
4. `wantManifestSchemaVersion`, currently at `corpus_test.go:160`, in the same file.

On the version bump: `corpus/README.md:1959-1979` records the precedent in two halves. The 2b bump moved
what `Expected` carries and forced a full re-freeze; the 2d bump added only a manifest pin and forced
none. An `Extra` reshape is a `Case` field change, so by that precedent it is a bump from `"3"` to `"4"`,
but it forces no re-freeze because no committed case carries an `Extra` and no `Expected` field moves.
Bumping touches `corpus/manifest.json`, which the spec's contested-surface table places with lane E1, so
the plan needs the ruling written before the task dispatches rather than discovered at a merge.

For the stub half, the natural shape is a second map keyed the same way, plus `Consumed() []Miss`
alongside `Misses()`. Note that adding hit tracking to the *request* map at the same time is nearly free
and would let tier 1 grade the other direction (every logged request was actually issued). The spec
deliberately grades only zero-misses, because a strong ID match legally orphans search captures into
`Extra` and full consumption would then be wrong. Worth an explicit decision rather than a silent one.

### 11. The `time.Now` wall, and why the limiter runs into it

**This is the closest thing to a blocker in 3c's scope.**
`internal/harness/conform/arch_test.go:67-120` walks every `.go` file under `internal/engine` and fails on
any `time.Now` call. 3b deliberately widened that walk from `internal/engine/matcher` to all of
`internal/engine` (`:72-74`, R21; `docs/HISTORY.md` 3b entry, Task 8). `internal/engine/musicbrainz/limiter.go`
is a rate limiter and needs a monotonic clock.

The rule's stated reason is that the evaluation date is an explicit parameter and never a clock read
(`arch_test.go:67-70`), which is a different concern from a limiter's pacing clock. Three routes, each
with a real cost, and the plan should pick one rather than let an implementer improvise:

- Inject the clock as a package-var or struct field, following the repo's existing
  `doFetch`/`loadProfile`/`caseDir` override convention. Keeps the walk unchanged and gives the limiter a
  deterministic test seam for free. This looks like the right answer.
- Widen the walk's exemption with a documented reason. Weakens a checker that currently catches exactly
  the defect it was built for.
- Use `golang.org/x/time/rate`. The dependency is not in `go.mod` (which today carries only cobra) and
  there is no `vendor/`, so adding it needs a network fetch and a `tidy-check` pass. A hand-rolled limiter
  is small enough that the dependency is hard to justify.

Note that `time.Sleep`, `time.After`, and `time.Ticker` are all invisible to the walk, which inspects only
`time.Now` selector calls (`arch_test.go:99-113`). A port that paced with `Ticker` would pass the gate
while still reading a clock, which is worth saying out loud so nobody treats a green walk as proof.

### 12. Constant duplication, and the direction the arch rule actually allows

`releaseLookupPrefix` already exists in three copies, deliberately unshared
(`internal/harness/oracle/album_info.go:28`, `internal/harness/analysis/album_info.go:25`,
`internal/harness/conform/translate.go:20`), and 3b's carry says a fourth is the trigger to revisit
(`docs/HISTORY.md:352-357`). 3c's client is that fourth.

The architectural rule is one-directional: `internal/engine` may never import `internal/harness`
(`arch_test.go:16-21`), and the harness importing the engine is fine and already happens throughout
`conform`. So the clean resolution is available and is worth naming in the plan. The engine's
`musicbrainz` package can own the canonical `/ws/2` path constants and the `RELEASE_INCLUDES` list, and
`internal/harness/mbrequest` can import them, collapsing `MaximalReleaseInclude` (`mbrequest.go:22-25`)
into a single definition rather than a fourth copy. Verify that `mbrequest.MaximalReleaseInclude` is
byte-identical to beets' joined `RELEASE_INCLUDES` before collapsing anything; it is today, checked
element by element against `SP/beetsplug/_utils/musicbrainz.py:50-67`. `mbrequest` is imported by
`cmd/harness/capture.go:17`, so a move touches the capture tool too.

### 13. Budget interface, and the one contended resource

The spec places the limiter in `internal/engine/musicbrainz` and requires the per-caller budget interface
with priority classes, a user-present probe outranking a batch import, re-verified at rung 9 against real
callers. Rung 3's own conformance runs are offline and spend nothing.

There are already two live spenders of the same real budget in this tree, and neither goes through a
limiter. `cmd/harness/capture.go:29-31` defines `requestInterval = time.Second` and paces with a bare
`sleep(requestInterval)` (`:319-322`), with `liveFetchTimeout = 30 * time.Second` (`:41`), no retry at
all, and its own UA. Rung 4's shadow is the third spender and is named as spending deferred background
budget through this limiter. Whether 3c's limiter is meant to subsume the capture tool's pacer, or to
coexist beside it, is a scope question the pass split does not answer. Coexistence looks right for this
pass (the capture tool is `cmd/`, outside 3c's declared territory), but the plan should say so rather than
leave it.

---

## (b) Harness, stub, cassette, and request logs as landed

Everything below is read off the tree at `d518e57`. Every figure is measured, not quoted from a document.

### 1. `internal/harness/cassette` as landed

One production file, `internal/harness/cassette/cassette.go` (253 lines), plus `cassette_test.go` (330) and one fixture directory `internal/harness/cassette/testdata/corrupted-case/`.

**The types 3c touches.**

- `Case` at `internal/harness/cassette/cassette.go:41-49`. Six fields: `ID`, `Input`, `Profile`, `EvalDate`, `Requests []Request`, `Extra []string` (`json:"extra,omitempty"`, `:47`), `Expected *Expected`.
- `Request` at `:61-66`, three fields only: `Method string`, `PathQuery string`, `Response string`. **No headers, no status, no body, no timing.** The request log records what was asked for and which file answers it, nothing about the transport.
- `Manifest` at `:104-115`, nine fields including `SchemaVersion` (`:106`) and `ProfileHashes`.

**`LoadCase` and its validation rules** (`:117-136` for the entry point, `:141-179` for `validate`):

1. `ID`, `Profile`, `EvalDate` must each be non-empty (`:143-150`).
2. Per request, in index order (`:152-167`): `Method` non-empty (`:154-156`); `PathQuery` must already equal `canonicalizePathQuery(PathQuery)`, with the error naming both the stored and the canonical form (`:157-159`); `Response` non-empty (`:160-162`); the response file must be readable under `dir` (`:163-165`). Each response name is entered into a `consumed` set (`:166`).
3. **The Extra-versus-Request exclusivity check** (`:169-176`): an `Extra` entry naming a file already in `consumed` fails with `"extra[%d]: %s already consumed by a request"` (`:171`), and each `Extra` entry must itself name a readable file (`:173-175`). Its firing proof is `TestLoadCaseRejectsExtraNamingConsumedResponse` at `cassette_test.go:236-248`, which asserts the error text contains `"already consumed"`.

`checkResponseReadable` (`:181-189`) proves existence and readability with `os.ReadFile` and discards the bytes. No response body ever enters memory through the cassette package.

**The canonical path_query rule** is `canonicalizePathQuery` at `:214-228`, twelve lines: cut on the first `?`; if there is no query or the query is empty, return the path unchanged; otherwise split the raw query on `&`, `slices.Sort` the resulting `key=value` **pair strings**, and rejoin. Nothing is decoded and nothing is re-encoded. The path, trailing slash included, is untouched, so `/release/abc` and `/release/abc/` stay distinct keys (proven at `cassette_test.go:40-45`). The three wire facts the rule exists to protect are documented at `cassette.go:15-24` and again at `corpus/README.md:1907-1943`: `fmt=json` is forced on every request, `inc=` values join with a logical `+` that is `%2B` on the wire, and search queries carry Lucene backslash escapes.

Note the sort is on the whole pair, not on the key. With today's parameter sets (`fmt`+`inc`, and `fmt`+`limit`+`query`) pair order and key order coincide, so a Go client using `url.Values.Encode()` (which sorts by key) produces the same string. That coincidence is not guaranteed by the rule and should not be relied on silently.

**`Write`** (`:191-212`) marshals the whole `Case` through `json.MarshalIndent` with a trailing newline. It is the only writer of `case.json` in the repo, reached from `harness freeze` (`cmd/harness/freeze.go:77-80`) and therefore from every `make derive` re-freeze.

**`LoadManifest`** (`:230-253`) checks shape only: non-empty `SchemaVersion` and `Beets`. It does not compare the version against anything.

### 2. The schema constant

`wantManifestSchemaVersion = "3"` lives **exactly once**, at `internal/harness/corpus_test.go:160`, read by exactly one test, `TestManifestParsesAndMatchesSchemaVersion` at `:162-174`. The value it is compared against is `corpus/manifest.json:2`. A repo-wide grep for the identifier finds it in that one file and in prose only. The spec's "add no second copy" instruction is accurate against the tree.

`corpus/manifest.json` as committed is thirteen lines: `schema_version "3"`, `beets 2.13.1`, `unidecode 1.4.0`, `jellyfish 1.2.1`, `lap 0.5.13`, `numpy 2.5.2`, `mediafile 0.17.0`, `shim_version "1"`, and a `profile_hashes` map carrying **only** `locale` (`:10-12`), which is the still-open m1 gap.

### 3. `internal/harness/stub` as landed

One production file, `internal/harness/stub/stub.go` (135 lines), plus `stub_test.go` (152). **No `testdata` directory**; every stub test builds its case in `t.TempDir()` (`stub_test.go:17-39`).

- `key` at `stub.go:21-26` is `{method, pathQuery}`. That pair is the whole request identity.
- `Miss` at `:35-38` is `{Method, PathQuery}` with json tags, and it doubles as the on-disk `misses.json` element shape shared with `cmd/harness` (`:28-34`).
- `Server` at `:40-49` holds the listener, the `http.Server`, `responses map[key][]byte`, and a mutex-guarded `misses []Miss`.
- **`New`** at `:51-77`. The replay map is built at `:55-62` and iterates **`c.Requests` only**. Every body is read eagerly into memory with `os.ReadFile` before the listener opens. `Extra` is not read, not mapped, and not mentioned anywhere in the package.
- **`ServeHTTP`** at `:79-91`. It canonicalizes `r.URL.EscapedPath()` plus `r.URL.RawQuery`, looks up `{r.Method, pathQuery}`, and on a miss calls `recordMiss` and writes `404` with an empty body (`:84-88`). On a hit it sets `Content-Type: application/json` and writes the bytes (`:89-90`). **Request headers are never read.** A `User-Agent`, an `Accept`, or an auth header is invisible to the stub in both directions.
- `recordMiss` at `:93-97`, `Misses()` at `:99-105` returning `slices.Clone` in receipt order, `Addr()` at `:107-111`, `Close()` at `:113-116`.
- `canonicalize` at `:118-135` duplicates cassette's rule deliberately, with the reason on the function (`:126-127`): cassette ships frozen and keeps its rule unexported. It uses `EscapedPath()` rather than `Path` on purpose, documented at `:121-125`.

**There is no third outcome.** The stub can answer 200 or 404 and nothing else, and `cassette.Request` carries no status field to express one. That is the "response-status extension", declined by name in 2c and again in 2d (`docs/HISTORY.md:820-823`, `:981-987`) and still open. It bears directly on 3c: **the client's retry behavior cannot be exercised through the stub**, because the stub can never produce the 503 or 429 a retry reacts to.

**Existing `stub` consumers**, all of which a signature change would break: `internal/harness/oracle/oracle.go:163`, `internal/harness/oracle/ties.go:65`, `internal/harness/oracle/album_info.go:57`, and `cmd/harness/capture.go` (`stub.Miss` only, `:231-245`). `internal/harness/conform` does **not** import `stub` today; 3c is its first non-oracle consumer.

**One asymmetry worth naming.** The stub records unmatched requests and records nothing about matched ones. There is no hit set, no `Consumed()` for `Requests`, and no unused-request report. A logged `Request` that the Go client never issues produces no miss and no signal at all, so tier 1's zero-miss rule is a one-directional check. It catches a client asking for too much and is blind to a client asking for too little.

### 4. `internal/harness/mbrequest` as landed

41 lines, `internal/harness/mbrequest/mbrequest.go`.

- `MaximalReleaseInclude` at `:17-24` is the `+`-joined sixteen-element include list. It matches pinned beets' `RELEASE_INCLUDES` element for element and in order (`tools/venv/.venv/lib/python3.13/site-packages/beetsplug/_utils/musicbrainz.py:50-67`).
- `CanonicalizeQuery(path, params)` at `:26-41` escapes each key and value with `url.QueryEscape`, joins with `=`, sorts the pairs, and joins with `&`. The doc comment at `:27-33` records why `url.QueryEscape` and not a literal `+`: Python `requests` percent-escapes a literal `+` in a value to `%2B`, and Task 5's live capture discovered the mismatch the hard way.

This package canonicalizes the query only. It does not build a search query, has no Lucene escaping, and knows nothing about `limit` or `query` parameters. A 3c Lucene builder has no existing Go home to extend.

### 5. `internal/harness/conform` as 3a and 3b left it

Seven files, 2,639 lines including tests. Package doc at `internal/harness/conform/doc.go:1-23` states the harness-to-engine direction and that the package reads `corpus/analysis/components.json`, `corpus/manifest.json`, and `corpus/cases` read-only and writes nothing.

**Tier 0** is `component.go` (910 lines) and `component_test.go` (433). `componentMethods` at `component.go:24-36` is conform's own eleven-method list, maintained separately from `analysis`'s unexported copy on purpose (`:16-23`). `ReplayOutcome` at `:38-58` carries `RowID`, `Method`, `Err`, `NearMiss`, `Detail`. `ReplaySummary` at `:60-73` carries `CountByMethod`, `NearMisses`, `TotalRows`. `ReplayTable` at `:79-94` excludes a failed row from the tally so that "replayed count equals row count" catches a silent skip. `ReplayRow` dispatches on `method`, and inside `add` on an argument shape, for eleven methods and twelve shapes (`:96-100`).

**The defaults checks** are `defaults.go` (76 lines): `CheckCaseWeights` (`:21-39`), `CheckCaseThresholds` (`:43-58`), `CheckUnidecodeVersion` (`:67-72`).

**The translation tier** is `translate.go` (443 lines). Three things in it are 3c's direct business:

- `releaseLookupPrefix = "/ws/2/release/"` at `:20`, the third copy of the constant (the others at `internal/harness/analysis/album_info.go:25` and `internal/harness/oracle/album_info.go:28`). The 3b close records the triplication as deliberate and worth revisiting if a fourth appears.
- `findReleaseRequest` at `:288-302` matches a lookup by exact path or by `prefix + "?"`, deliberately ignoring the query. `loadRelease` at `:304-327` reads the named response file directly and notes at `:306-308` that "the stub is 3c's business".
- `checkNoReleaseReachesBrowseThreshold` at `:396-443` walks every case's release-lookup responses, sums `len(medium.tracks)`, and fails at `>= browseMaxTracks` (`:29`, value 500). It uses `>=` where beets uses `>`, so it fires one track early on purpose.

**The architecture checks** are `arch_test.go` (206 lines) and they carry a hard constraint for 3c's limiter.

- `checkNoHarnessImports` (`:30-64`) walks `internal/engine` in `ImportsOnly` mode and skips any `testdata` subtree (`:36-38`).
- `checkNoTimeNow` (`:75-118`) walks the same tree and flags **any call whose selector resolves to `time.Now`** (`:108-110`). It flags nothing else. `time.Since`, `time.NewTicker`, `time.After`, `time.Sleep`, and `context.WithTimeout` all pass, and `internal/engine/matcher` already imports `time` in nineteen production files for the evaluation-date type.
- Entry points `TestEngineNeverImportsHarness` (`:141-146`) and `TestEngineNeverReadsTheClock` (`:152-157`) both root at `internal/engine`, so **`internal/engine/musicbrainz/limiter.go` is inside the walk**. A limiter written the obvious way calls `time.Now` and reds the gate for every lane. The three firing proofs are at `:163-206`.

`conform/testdata` holds seven files today, five JSON decoder fixtures plus three `arch/` violation trees, and one translation fixture at `testdata/translate/skipped-row-ledger.json`.

### 6. `analysis` and `cmd/harness`: who reads the request log

**`internal/harness/analysis`** touches the request log in exactly one place: `releaseLookups` at `album_info.go:248-264`, which cuts `releaseLookupPrefix` off each `PathQuery`, takes everything before the `?`, and returns the mbids in log order. Its doc comment records the discriminator that makes the prefix sufficient: a search path is `/ws/2/release?` and never carries the trailing slash (`:249-251`). Nothing in `analysis` reads `Extra`, a method, or a response file.

**`cmd/harness/capture.go`** (350 lines) is the capture tool.

- Identity: `contactEnvVar = "DUBPLATE_HARNESS_CONTACT"` (`:26`), enforced by `requireContact` (`:120-128`), which refuses to run unidentified. `userAgent` at `:130-134` builds `dubplate-harness/<version> ( <contact> )`. This is the harness's own User-Agent and has nothing to do with beets' or with 3c's client.
- Wire discipline: `requestInterval = time.Second` (`:31`) via `pacer` (`:312-323`), which never delays the first request.
- `liveFetch` (`:325-350`) is the only production HTTP client in the repo outside the stub. It sets `User-Agent` as a request header (`:333`), uses `http.Client{Timeout: 30 * time.Second}` (`:41`, `:335`), and accepts only 200 (`:346-348`). Pinned beets uses 10s instead (`beetsplug/_utils/requests.py:99`) and a `User-Agent` of `beets/{__version__} https://beets.io/` set on the session (`:76`).
- `releaseRequest` (`:275-283`) is the one request builder, calling `mbrequest.CanonicalizeQuery` with `fmt=json` and the maximal include. **The capture tool builds no search request at all.**
- `runCapture` (`:142-179`) seeds a case with one lookup, plus a second when `pseudoReleaseFollowUp` (`:285-310`) finds a backward `transl-tracklisting` relation on a `Pseudo-Release`. `runCaptureMisses` (`:186-225`) appends one `Request` per `misses.json` entry and deletes the miss file.

**How a search capture becomes an orphan, in code rather than in prose.** It does not, today. `cmd/harness` never sets `Case.Extra`, and neither does anything else. The mechanism the README describes at `corpus/README.md:2770-2785` is a **maintainer's manual bookkeeping act**: when a more complete capture retires beets' search fallback, the derivation issues no request, the stub records no miss, `misses.json` is absent, and `harness freeze` succeeds. The now-unfired response file stays referenced by its original `Request` entry, `TestEveryResponseFileIsConsumed` still passes because a `Request` names it, and nothing anywhere notices. Someone would have to move that filename from `requests` into `extra` by hand. That is why the corpus carries zero `Extra` entries, and it is a stronger statement than "the channel is dormant": **the channel has no producer.**

`cmd/harness/seeditems.go:79-83` and `:116-120` read `c.Requests[0].Response` positionally, which is a second reason a `Requests` reordering would be disruptive. Neither reads `Extra`.

**The shim never sees the request log.** `writeShimInput` at `internal/harness/oracle/oracle.go:280-289` marshals `{CaseID, EvalDate, Items}` and nothing else. No Python file in the repo mentions `extra`, `requests`, or `path_query`. A `Case.Extra` schema change touches zero Python.

### 7. Corpus figures, measured off the tree

| Figure | Value |
| --- | --- |
| Case directories under `corpus/cases/` | **215** (215 `case.json`) |
| Cases carrying an `Extra` entry | **0** |
| `case.json` files containing the key `"extra"` at all | **0** (top-level keys are exactly `eval_date`, `expected`, `id`, `input`, `profile`, `requests`, all 215 each) |
| Total logged requests | **351** |
| Total response files | **351** (so zero orphans on disk, consistent with the two counts above) |
| Methods | **GET 351**, no other method anywhere |
| Release lookups (`/ws/2/release/<mbid>`) | **303** |
| Searches (`/ws/2/release?`) | **48** |
| Distinct release mbids looked up | 255 |
| Distinct search `path_query` values | 46 (two are shared by a case and its `-stock` twin) |
| Cases with at least one search | **48** |
| Cases with no search (ID-match only) | **167** |

Per-case request-shape distribution:

| Shape | Cases |
| --- | --- |
| 0 search, 1 lookup | 164 |
| 1 search, 1 lookup | 26 |
| 1 search, 6 lookups | 9 |
| 1 search, 5 lookups | 9 |
| 1 search, 2 lookups | 4 |
| 0 search, 2 lookups | 3 (`non-latin-pseudo-release`, `real-nonlatin-zemfira`, `real-nonlatin-zemfira-stock`, the pseudo-release double fetch) |

No case logs more than one search. No case logs the same release mbid twice, which is what makes `album_info.json`'s deduplicated row keying safe.

**Parameter sets actually present.** Only two, and each is uniform:

- Lookups: `fmt` and `inc`, stored in that order, 303 of 303.
- Searches: `fmt`, `limit`, `query`, stored in that order, 48 of 48.

**The `inc=` sets.** There is exactly **one** distinct `inc=` value in the whole corpus, on all 303 lookups:

```
inc=artists%2Bmedia%2Brecordings%2Brelease-groups%2Blabels%2Bartist-credits%2B
aliases%2Brecording-level-rels%2Bwork-rels%2Bwork-level-rels%2Bartist-rels%2B
isrcs%2Burl-rels%2Brelease-rels%2Bgenres%2Btags
```

Sixteen includes, fifteen `%2B` joins, byte-identical to `mbrequest.MaximalReleaseInclude` (`mbrequest.go:22-24`) and to `RELEASE_INCLUDES` (`beetsplug/_utils/musicbrainz.py:50-67`). `fmt=json` on all 351. `limit=5` on all 48 searches, one distinct value.

**Search query construction, as the log records it.** Every search is `release:(<album>)` plus one artist clause, joined by a space (encoded `+`): `artist:(<artist>)` on 43 and `arid:(<mbid>)` on 5. All terms are lowercased. The escaping matches `LUCENE_SPECIAL_CHAR_PAT = re.compile(r'([-+&|!(){}[\]^"~*?:\\/])')` at `beetsplug/_utils/musicbrainz.py:48`, applied by `format_search_term` at `:650-663` before URL encoding.

The full percent-escape census across all 351 stored `path_query` values, which is the exact byte inventory a 3c builder must reproduce:

| Escape | Count | Origin |
| --- | --- | --- |
| `%2B` | 4545 | the `+` include joins, 303 × 15 exactly |
| `%3A` | 97 | 96 field-prefix colons (48 × 2) plus one Lucene-escaped `\:` in `frank zappa\: cucamonga` |
| `%28` / `%29` | 96 each | the `field:(...)` parentheses, 48 × 2 |
| `%5C` | 24 | Lucene backslashes: 20 from five `arid` uuids at four dashes each, 2 from `\"choral\"`, 1 from `bob dylan \- the bootleg`, 1 from the zappa colon |
| `%27` | 2 | apostrophes in `girls' generation` and `founder's creek tapes` |
| `%22` | 2 | the escaped quotes in `symphony no. 9, \"choral\"` |
| `%2C` | 1 | the comma in that same title |
| multibyte UTF-8 | 3 sequences | Japanese `浜崎あゆみ` and Korean `소녀시대` artist terms |

Literal characters surviving unescaped inside query strings: `-` (3051, mbids and hyphenated include names), `+` (145, spaces inside search terms), and `.` (1, the `no.` in the Beethoven title).

**Headers and User-Agent in the log: none.** `cassette.Request` has three fields (`cassette.go:62-66`), the stub reads only method and URL (`stub.go:82-83`), and no artifact anywhere in `corpus/` records a header. A Go client's `User-Agent` is **ungradeable through the stub** and can only be proven by a unit test on the request the client builds.

**Corpus weight, for the tier-1 runner's cost model.** 104 MB under `corpus/cases/`, of which 77.6 MB is response bodies across 351 files, mean 226 KB. `stub.New` loads a case's bodies eagerly and entirely into memory (`stub.go:55-62`). The largest single body is 8,894 KB (`real-boxset-rachmaninov-complete/responses/001.json` and its `-stock` twin), so those two stubs allocate ~8.7 MB each at start. A full tier-1 sweep that stubs each case once reads all 77.6 MB from disk per run at `-count=1`.

### 8. What a `Case.Extra` schema bump would touch

Complete, from a repo-wide grep excluding `_salvage/`, `tools/venv/`, and `corpus/cases/`.

**Writers of `Extra`: none in production code.** The only writes are `cassette_test.go:67` (in `baseCase()`) and the five `c.Extra = nil` / `c.Extra = []string{...}` lines at `:181`, `:198`, `:218`, `:238`, `:255`. No `cmd/`, no `internal/harness/oracle`, no Python.

**Readers, all of them:**

| Site | What it does |
| --- | --- |
| `internal/harness/cassette/cassette.go:47` | the field declaration, `[]string` |
| `internal/harness/cassette/cassette.go:7-9` | package doc, "`[Request.Response]` and `[Case.Extra]` name those files" |
| `internal/harness/cassette/cassette.go:120-121` | `LoadCase`'s doc comment stating the exclusivity rule |
| `internal/harness/cassette/cassette.go:169-176` | `validate`'s Extra loop, both rules |
| `internal/harness/cassette/cassette.go:197-212` | `Write`, which serializes the field |
| `internal/harness/corpus_test.go:123-135`, `:148` | `TestEveryResponseFileIsConsumed`, the only consumer outside `cassette` |
| `internal/harness/cassette/cassette_test.go` (7 sites above) | the fixture and the exclusivity proof |
| `corpus/README.md:41-46` | the Layout section's Extra paragraph |
| `corpus/README.md:2760` | the freeze section's "`requests`, and `extra` bytes match the first's exactly" |
| `corpus/README.md:2770-2785` | the recorded-fixed-point section, the orphaning narrative |

**Derive paths touched: the write path only, and inertly.** `harness freeze` calls `cassette.LoadCase` then `Case.Write` (`cmd/harness/freeze.go:60-80`), so every `make derive` re-freeze round-trips the field. Because `Extra` is `omitempty` and every committed case's value is nil, a change of the element type from `string` to a `Request`-shaped struct **emits the same zero bytes**, and byte-identity across all 215 records holds by construction. The capture tool never writes the field, so it needs no change to keep working, though a `Request`-shaped `Extra` is what would finally let it (or a freeze-time step) become the channel's producer.

**Schema constant: one move, `internal/harness/corpus_test.go:160`, `"3"` to `"4"`, plus `corpus/manifest.json:2`.** No second copy exists.

**Not touched by the bump:** every Python file (the shim input is `{CaseID, EvalDate, Items}` only), `internal/harness/analysis` (reads `Requests`, never `Extra`), `internal/harness/synth`, `cmd/harness/seeditems.go`, and `internal/harness/oracle` (three `stub.New` calls, none of which reads `Extra`).

### 9. The schema-bump precedent

Two bumps, and they set opposite precedents. Both are recorded at `corpus/README.md:1959-1980`.

**`"1"` to `"2"`, pass 2b Task 1.** `cassette.Expected` gained `weights` and `thresholds`, closing a gap where `oracle.Derive`'s `json.Unmarshal` silently dropped both. Because the change moved what `Expected` itself carries, **it forced a re-derivation of every committed record** through the then-new `harness freeze` verb, and `wantManifestSchemaVersion` moved with it (`README.md:1959-1966`).

**`"2"` to `"3"`, pass 2d Task 1, the closer precedent.** `cassette.Manifest` gained `MediaFile` pinning `mediafile==0.17.0`, and `Reconcile`'s `pinned` tuple and `versionProbe` string each gained a sixth component. **This bump forced no re-freeze** (`README.md:1968-1980`). The stated rule, worth quoting because 3c's case sits under it: "a schema bump forces a re-freeze only when it moves what `Expected` itself carries, never when it only adds a manifest pin nothing in `Expected` echoes." The evidence 2d produced was `make derive` re-deriving all 215 records byte-identically with the new pin in place; `docs/HISTORY.md:843` records it as "the `schema_version` `2` -> `3` bump forced no re-freeze".

3c's `Case.Extra` change is a third shape neither precedent exactly covers. It moves what `Case` carries rather than what `Expected` or `Manifest` carries. It nonetheless lands on 2d's side of the rule for a mechanical reason rather than a categorical one: the field is `omitempty` and universally empty, so no committed byte moves. The honest framing for the plan is that the re-freeze is unnecessary because the migration set is empty, not because `Case` is exempt. A cheap confirmation is available without running `make derive`: re-marshal each committed `Case` through the new type in a test and diff against the file on disk.

### 10. Findings 3c should carry into its plan

1. **The `time.Now` ban covers the limiter.** `checkNoTimeNow` roots at `internal/engine` (`arch_test.go:152-157`) and flags `time.Now` syntactically (`:108-110`). A limiter in `internal/engine/musicbrainz/limiter.go` calling `time.Now` reds the gate for every lane. `time.Since`, tickers, and timers pass. This wants a decision at plan time rather than at the first red.
2. **The stub cannot grade retries.** Two outcomes only, 200 and 404, and no schema field for a status. Retry behavior needs unit tests against an injected transport, and the response-status extension stays declined unless 3c chooses to take it.
3. **The stub cannot grade headers.** `User-Agent` conformance is a unit-test claim, not a stub claim.
4. **The zero-miss rule is one-directional.** Nothing today reports which logged `Request` entries a run actually consumed, so a Go client issuing *fewer* requests than the log records produces a green tier 1. If 3c is adding `Consumed()` for `Extra`, adding the symmetric report for `Requests` is nearly free and closes the blind side.
5. **`Case.Extra` has no producer.** Zero code paths write it. A `Request`-shaped `Extra` is a prerequisite for the capture tool or freeze ever populating it; without a producer, the channel stays hand-authored forever. Worth stating in the plan so a later pass does not assume the loop fills it.
6. **The Lucene builder has no Go home to extend.** `mbrequest` canonicalizes a query and nothing more. The escape set, the lowercasing, the `field:(...)` wrapping, and the `arid` versus `artist` selection all arrive fresh in 3c.
7. **The corpus grades the query builder narrowly.** One `inc=` value, one `limit=` value, one `fmt=` value, and 46 distinct search queries covering seven of the fourteen Lucene special characters (`-`, `"`, `:` observed escaped; `+ & | ! { } [ ] ^ ~ * ? /` and `\` never appear in a committed term). The unexercised eight need hand-authored unit fixtures on the proximity precedent.
8. **`releaseLookupPrefix` reaches its fourth copy if 3c adds one.** The 3b close flagged three copies as deliberate and named a fourth as the trigger to revisit. A 3c client or grader that needs the prefix hits that trigger.
9. **The merged-track aliasing carry is live for any 3c caller.** `MergePseudoRelease` copies shallowly and `merged.Tracks` aliases the pseudo's backing array (`docs/HISTORY.md:265-272`).

---

## (c) The engine as landed and the 3c Go surface

### 1. The engine as landed at HEAD (`d518e57`)

#### 1.1 `internal/engine/musicbrainz`, eight files, zero network

3b birthed the package and cut every seam 3c needs. Nothing in it issues a request, and the package doc says so (`internal/engine/musicbrainz/doc.go:1-5`: "the client, the Lucene query construction, and the rate limiter are 3c's territory").

| File | Exported surface 3c consumes |
| --- | --- |
| `release.go` | `Release` and its sub-types, `(*Release).Validate`, `ErrMalformedRelease` |
| `translate.go` | `TranslateAlbum(rel *Release, tun matcher.Tunables) (*matcher.AlbumInfo, error)` |
| `pseudo.go` | `PseudoReleaseTarget(rel *Release) (string, bool)`, `MergePseudoRelease(pseudo, actual *matcher.AlbumInfo) *matcher.AlbumInfo` |
| `relations.go` | `GroupRelations` |
| `tracks.go`, `alias.go`, `credit.go`, `events.go` | unexported only |

Three facts bind the client. `ErrMalformedRelease` is declared "the one sentinel every boundary failure in this package wraps" (`internal/engine/musicbrainz/release.go:10-13`), so the client's decode failures wrap it rather than inventing a peer. `PseudoReleaseTarget`'s doc names 3c outright: "The FETCH of that release is 3c's client work; 3b exports the predicate so 3c calls it" (`internal/engine/musicbrainz/pseudo.go:17-19`). And the browse fallback stays unported (`internal/engine/musicbrainz/doc.go:7-16`), guarded from the harness side by `checkNoReleaseReachesBrowseThreshold` at `>=` 500 with the corpus maximum at 470 (`internal/harness/conform/translate.go:22-29`, `internal/harness/conform/translate.go:388-443`).

`wireInt` (`internal/engine/musicbrainz/release.go:150-175`) is corpus-only. HISTORY guard 16 records that "position is the only two-shape field" and that "3c's live client meets real captures only, where the string form does not occur" (`docs/HISTORY.md:322-323`). The client adds no second custom unmarshaller.

#### 1.2 `internal/engine/matcher`, the client-relevant fields

Every citation below is the field's own line in `internal/engine/matcher/tunables.go`.

| Field | Line | Default | 3c relevance |
| --- | --- | --- | --- |
| `SearchLimit int` | `:119-124` | 5 | The `limit=` parameter on every search. Operator-facing, unregistered. |
| `ExtraTags []string` | `:151-154` | empty | Selects extra Lucene criteria. Zero committed cases exercise it. |
| `DataSourceMismatchPenalty float64` | `:126-136` | 0.5 | Carries the divergence in section 6. |
| `DataSourcePenalties map[string]float64` | `:138-149` | `{"MusicBrainz": 0.5}` | The other half of that divergence. |
| `Languages`, `IgnoredAliasTypes` | `:161-172` | empty | Read by the translation, not the query. |
| `Preferred.Countries` | `:220-222` | empty | Two matching regimes across two packages (HISTORY guard 9, `docs/HISTORY.md:273-277`). |
| `Timid bool` | `:157-159` | false | Suppresses the ID-match early return (`beets/autotag/match.py:433`). 3d's, named here so 3c's query builder does not absorb it. |

The likely-tags helper is `MostCommonTags(items []ItemFields) (Likelies, map[string]bool)` (`internal/engine/matcher/likelies.go:47-83`), returning the thirteen plurality fields plus a consensus map, with the albumartist override at `:78-80`. Ties break by first-seen insertion order through `plurality` (`internal/engine/matcher/likelies.go:95-122`), mirroring `beets/util/__init__.py:799-807`. `ItemFields` is eighteen fields (`internal/engine/matcher/info.go:124-143`) and `ProjectItemFields` decodes a corpus case's item bag into it (`:179-200`).

The data-source components are two two-line files that both delegate the fire condition to the accumulator: `internal/engine/matcher/track_data_source.go:13-15` and `internal/engine/matcher/album_data_source.go:10-12`. The condition and the fallback live at `internal/engine/matcher/penalty.go:338-346` and `:404-409`.

#### 1.3 The harness surfaces 3c writes

`stub.Server` loads only `c.Requests` into one `map[key][]byte` and answers everything else with a recorded 404 (`internal/harness/stub/stub.go:54-77`, `:81-97`). `Misses()` returns the unmatched list in receipt order (`:99-105`). `canonicalize` duplicates cassette's rule deliberately, because cassette ships frozen (`:118-135`).

`cassette.Case.Extra` is `[]string`, a bare filename list (`internal/harness/cassette/cassette.go:47`), and `validate` rejects an `Extra` entry naming a file a `Request` already consumes (`:169-176`). Verified against the tree: **zero of 215 committed cases carry an `extra` key**, so the migration is empty, exactly as the spec assumed.

`wantManifestSchemaVersion = "3"` lives once, at `internal/harness/corpus_test.go:160`, read once at `:171`. The 2b and 2d precedent is to change that constant's value alongside `corpus/manifest.json` rather than relocate the symbol (`docs/superpowers/plans/2026-09-03-rung-2d-file-level.md:400-401`).

`internal/harness/conform` already imports `internal/engine/musicbrainz` (`internal/harness/conform/translate.go:12`), so the new engine files are reachable with no new import direction. `gradeLedger`'s compared-row discipline (`internal/harness/conform/translate.go:318-380`) is the house pattern the request grader should copy.

#### 1.4 `internal/config`, and rung 1b's status

**Rung 1b has not landed at HEAD.** `ownedWall` is still a slice literal of five rows with no registry (`internal/config/keys.go:19-25`), which is the exact shape 3a's guard 2 predicted (`docs/HISTORY.md:566-572`). The four-knob registration stays deferred and `keys.go` stays untouched.

An operator-facing value reaches code two ways today. A wall row is read once at `Load` and lands as a field on `Config` (`internal/config/config.go:53-60`). A behavior tunable registers into an append-only registry through `Register` (`internal/config/thresholds.go:19-35`) and resolves through `ParseInt` (`:43-60`). One gap worth writing into the plan: `Threshold.Default` is a string and `ParseInt` is the only parser, so the three float rec thresholds have no resolver even once 1b lands. Whoever does the registration needs a float variant.

Rung 3's four knobs are `StrongRecThresh`, `MediumRecThresh`, `RecGapThresh`, and `SearchLimit`. Only the fourth is 3c's, and it is registered nowhere. 3c registers nothing.

---

### 2. `query.go`

#### 2.1 What the query is built from

The beets flow is `tag_album` → `metadata_plugins.candidates` → `get_search_query_with_filters` → `get_album_criteria` → `MusicBrainzAPI.search`. Four inputs, all derivable from what 3a landed.

- `search_artist`, `search_name` default to `likelies["artist"]` and `likelies["album"]` (`beets/autotag/match.py:407-409`, `:446-448`). Both are `matcher.Likelies` fields.
- `va_likely` is `(not consensus["artist"]) or (search_artist.lower() in VA_ARTISTS) or any(item.comp for item in items)` (`beets/autotag/match.py:451-455`).
- `len(items)` feeds the `tracks` extra tag (`beetsplug/musicbrainz.py:726`).
- The six other extra tags read a plurality over the same helper (`beetsplug/musicbrainz.py:731`, `beets/util/__init__.py:799-807`).

That last point is load-bearing and pleasant: `FIELDS_TO_MB_KEYS` (`beetsplug/musicbrainz.py:47-56`) maps `barcode`, `catalognum`, `country`, `label`, `media`, `year` to MB fields, and **all six are already `matcher.Likelies` fields** (`internal/engine/matcher/likelies.go:11-25`). The query builder needs no exported plurality helper and no second projection. It takes `Likelies`, the item count, and the two search terms.

Two things `matcher` does not supply yet.

- **`comp` is not in `ItemFields`.** 214 of 215 committed cases carry a `comp` key on their items, and `ProjectItemFields` silently drops it. `ItemFields`' own doc says "A field outside this set is not read by the distance path; adding one is a finding, not a convenience" (`internal/engine/matcher/info.go:112-113`). `comp` is read by the search path rather than the distance path, so the honest resolution is to widen the field set by one and amend that sentence to name the single search-path exception, with the reason on the field. A parallel second decoder over the same item bag would be worse.
- **`VA_ARTISTS` is unexported.** `vaArtists` sits at `internal/engine/matcher/track_artist.go:18-21`. Export `matcher.IsVariousArtistsName(s string) bool` and have `addTrackArtist` (`:33`) call it, so there is one copy with two consumers.

#### 2.2 Proposed surface

```go
// AlbumSearch is one album search request, built and not yet issued.
type AlbumSearch struct {
    Query string // the Lucene query, unescaped by the caller's standards and escaped by ours
    Limit int
}

// SearchTerms are tag_album's own derivation of what to search for.
func SearchTerms(likelies matcher.Likelies, consensus map[string]bool, items []matcher.ItemFields) (artist, album string, vaLikely bool)

// BuildAlbumSearch mirrors get_album_criteria and get_search_query_with_filters.
func BuildAlbumSearch(likelies matcher.Likelies, itemCount int, artist, album string, vaLikely bool, tun matcher.Tunables) AlbumSearch
```

Criteria order is the query's term order and is fixed: `release`, then `arid` when `vaLikely` else `artist`, then each configured extra tag in the operator's own `extra_tags` order filtered by `FIELDS_TO_MB_KEYS` membership (`beetsplug/musicbrainz.py:698-711`, `:717-733`). Go's `map` gives no order, so the extra-tag walk iterates `tun.ExtraTags` and looks up the MB key, never the reverse.

Five behaviors the port must reproduce and one it must not invent.

1. **Double lowering and stripping.** `get_search_query_with_filters` lowercases and strips each value and drops the empties (`beetsplug/musicbrainz.py:750-753`); `format_search_term` lowercases and strips again (`beetsplug/_utils/musicbrainz.py:653-654`). The second pass is idempotent under Python. Reproduce both anyway, because a later beets change to one and not the other is exactly what the port must be able to see.
2. **Lucene escaping.** `LUCENE_SPECIAL_CHAR_PAT = re.compile(r'([-+&|!(){}[\]^"~*?:\\/])')` (`beetsplug/_utils/musicbrainz.py:48`), each match prefixed with one backslash (`:659`). The wrapping `field:(term)` parentheses are added after escaping (`:661`), so they are never escaped themselves. The corpus proves both halves: `arid%3A%2889ad4ac3%5C-39f7...%29` shows an escaped hyphen inside unescaped parens, and `release%3A%28symphony+no.+9%2C+%5C%22choral%5C%22%29` shows escaped quotes.
3. **Empty values drop.** A criterion whose value lowercases and strips to empty is omitted (`beetsplug/musicbrainz.py:752`), and `format_search_term` returns "" for the same input (`:655-656`), which `filter(None, ...)` then removes (`beetsplug/_utils/musicbrainz.py:679-681`).
4. **Terms join with a single space** (`beetsplug/_utils/musicbrainz.py:679`). No literal `AND`.
5. **`catalognum` has its spaces removed** before it becomes a criterion (`beetsplug/musicbrainz.py:729-730`), and `year` becomes `str(most_common)`, so an unset year yields the string `"0"`, which survives the empty-value filter. Reproduce that, do not "fix" it.
6. **`search_query_ascii` is a structural no-op here and must not join `Tunables`.** `_search_api` unidecodes `query` (`beets/metadata_plugins.py:411-413`), the MusicBrainz plugin returns `query = ""` (`beetsplug/musicbrainz.py:750`), and `get_search_response` reads `params.filters` and `params.limit` and never `params.query` (`:755-762`). Record this as a checked non-membership so a later reader does not add the knob for completeness.

#### 2.3 The `inc=` sets

`RELEASE_INCLUDES` is sixteen values in fixed order (`beetsplug/_utils/musicbrainz.py:50-67`), joined with `+` and passed as `inc` (`:635`). `RECORDING_INCLUDES` (`:69-76`) belongs to `get_recording`, which nothing on the album path calls. `BROWSE_INCLUDES` (`beetsplug/musicbrainz.py:58-65`) belongs to the unported fallback.

Ship exactly one, `releaseIncludes`, as a `[]string` in beets' order, joined at request time. **A search request carries no `inc` at all**: `search` passes only `query` and `limit` into `_get_resource` (`beetsplug/_utils/musicbrainz.py:679-685`), and the corpus agrees, with `inc` on 303 of 351 committed requests and never on the 48 searches.

#### 2.4 Encoding, verified rather than assumed

`url.Values.Encode()` reproduces the committed canonical form byte for byte. Probe run offline against the three committed shapes: the maximal-`inc` lookup, the `"choral"` search with an escaped quote and a comma, and the VA `arid` search with escaped hyphens. All three matched the tree's own `path_query` strings exactly. `Encode` sorts by key, which coincides with cassette's byte-wise pair sort for these parameter names, and its `+`-for-space and `%2B`-for-plus behavior matches Python `requests`. Record the coincidence: cassette sorts whole `key=value` pairs (`internal/harness/cassette/cassette.go:220-228`) while `Encode` sorts keys, so a future parameter set with a shared key prefix could diverge. `mbrequest.CanonicalizeQuery` (`internal/harness/mbrequest/mbrequest.go:34-41`) is the harness-side rule, and the client cannot import it (section 7).

#### 2.5 One divergence the query path inherits and the distance path does not

3a's case-mapping sweep found 56 code points where Go's simple-case `strings.ToLower` and Python's full-case `str.lower` disagree, of which 54 are rescued by the distance pipeline's post-transliteration `[a-z0-9]` strip (`internal/engine/matcher/testdata/case-mapping-sweep.tsv`, `docs/HISTORY.md:451-462`). **The query path has no such strip**, so all 56 reach the wire. The corpus exercises the non-ASCII search path at least once (`artist%3A%28%E6%B5%9C%E5%B4%8E%E3%81%82%E3%82%86%E3%81%BF%29`), so this is reachable rather than theoretical. Give the query builder its own lowercasing helper with the sweep cited, and let the plan decide whether to mirror Python's full case mapping via `golang.org/x/text/cases` or record the 56 as a declared divergence.

A second, narrower one, verified offline: Python's `str.strip()` treats U+001C through U+001F as whitespace and Go's `unicode.IsSpace` does not. Four code points, no corpus instance, worth one named check rather than a paragraph.

---

### 3. `client.go`

#### 3.1 Shape

```go
const DefaultBaseURL = "https://musicbrainz.org"

type ClientConfig struct {
    BaseURL   string        // empty means DefaultBaseURL
    UserAgent string        // required; the client refuses to construct without one
    Budget    *Budget       // required; Unmetered() for offline runs
    Timeout   time.Duration // zero means 30s
    HTTP      *http.Client  // zero value means one built from Timeout
}

func NewClient(cfg ClientConfig) (*Client, error)

func (c *Client) Release(ctx context.Context, mbid string) (*Release, error)
func (c *Client) SearchReleases(ctx context.Context, s AlbumSearch) ([]string, error)
func (c *Client) AlbumForID(ctx context.Context, mbid string, tun matcher.Tunables) (*matcher.AlbumInfo, error)
```

`AlbumForID` is the one method that composes: fetch, `Validate`, `TranslateAlbum`, and, when `PseudoReleaseTarget` reports a target, the second fetch and `MergePseudoRelease` (`beetsplug/musicbrainz.py:765-798`). `SearchReleases` returns ids rather than releases, matching `candidates` at `beets/metadata_plugins.py:442-446`, and leaves the per-id fan-out and candidate ordering to 3d.

Base-URL injection is `ClientConfig.BaseURL`, and zero-means-default is the idiom Go's own `http.Client` uses. `conform` sets it to `"http://" + srv.Addr()`; the composition root leaves it empty. The client appends `/ws/2` itself, mirroring `api_root` (`beetsplug/_utils/musicbrainz.py:608-610`).

`UserAgent` is a required string and the engine never builds it. The universality rule puts the site-specific half in `internal/config`, and `matcher` already refuses that import for the same reason (`internal/engine/matcher/tunables.go:34-35`). So `cmd/dubplate-server` composes `dubplate-server/<version> ( <contact> )` from a `-ldflags` version var and a config value, following the harness precedent verbatim (`cmd/harness/capture.go:130-132`). The contact should be `Config.AdminEmail`, which is already a required wall row parsed as RFC 5322 (`internal/config/config.go:53-60`), so 3c mints nothing, adds no wall row, and touches no file rung 1b owns. A dedicated override is a post-1b follow-on if anyone wants one.

#### 3.2 Retry, timeout, pacing

beets' numbers, all from `beetsplug/_utils/requests.py`: `User-Agent` set at `:76`; `Retry(total=6, backoff_factor=0.5, status_forcelist=[500, 502, 503, 504, 429])` at `:78-88`; a 0.25s inter-request floor in `RateLimitAdapter` at `:89` and `:106-135`; a 10s default timeout at `:99`. On top of that the session is rate-limited at 1.0 requests per second for `musicbrainz.org`, hardcoded regardless of config (`beetsplug/_utils/musicbrainz.py:597-599`, `:612-613`), and `raise_for_status()` turns a 404 into `HTTPNotFoundError`, which `album_for_id` catches and converts to "no candidate" (`beetsplug/musicbrainz.py:777-780`).

The port takes the retry policy and departs on the timeout, with the reason stated. Seven attempts (one plus six retries), backoff `500ms * 2^(n-1)` capped, retrying on the same five statuses, honoring a `Retry-After` header when the response carries one. Timeout defaults to 30s rather than 10s, matching the capture tool's own stated reason that maximal-`inc` release lookups are large (`cmd/harness/capture.go:38-41`).

Three decisions neither beets nor the spec states, which the plan should settle explicitly.

1. **Every retry re-acquires a limiter token.** A retry is a real request against a shared wire budget. beets gets this free because its limiter sits in the adapter beneath the retry loop; a Go client with the limiter above the loop would silently burst.
2. **Every wait is context-cancellable.** Backoff uses `time.NewTimer` and a select on `ctx.Done()`, never `time.Sleep`. This also keeps the no-clock invariant intact (section 4.4).
3. **The client never follows the browse fallback.** `doc.go:7-16` refuses it and `conform` guards the assumption. 3c owns the port only if a corpus case crosses 500 tracks, which HISTORY guard 14 already assigns (`docs/HISTORY.md:313-317`). Nothing to build.

#### 3.3 Sentinels

Two new, plus the one that exists.

- `ErrNotFound`, for a 404 on a lookup. Callers must distinguish a deleted release from a real failure, because beets' own flow does (`beetsplug/musicbrainz.py:777-780`).
- `ErrUnavailable`, for retries exhausted on any transient status or transport error. Rung 4's shadow branches on it to defer rather than record a divergence.
- `ErrMalformedRelease`, existing, wrapped by decode failures and by `Validate` (`internal/engine/musicbrainz/release.go:10-13`).

A rate-limit-specific sentinel is deliberately not shipped. go-conventions makes `%w` an API commitment justified by a caller that branches on it, and no caller branches on 429 separately from 503 today. Record the omission rather than leaving it silent.

One existing sentinel needs a decision the plan should carry: HISTORY notes as a deliberate `code-simplifier` leftover that `wireInt`'s two decode errors do not wrap `ErrMalformedRelease`, while that sentinel's own comment calls it the one every boundary failure wraps (`docs/HISTORY.md:353-357`). 3c is the pass whose client makes the boundary real, so it is the natural place to close it.

#### 3.4 The `MergePseudoRelease` carry

HISTORY guard 8 warns that the merge is a shallow copy, so `merged.Tracks` aliases the pseudo's backing array and a 3c caller mutating a merged track writes through (`docs/HISTORY.md:265-272`, `internal/engine/musicbrainz/pseudo.go:49-58`).

`AlbumForID` closes it by construction. Both `AlbumInfo` values are built inside the call, only the merged one escapes, and the pseudo is dropped. Record that reasoning at the call site and add the one-line `slices.Clone` inside `MergePseudoRelease` the day a caller retains both. The hazard remains live for 3d and rung 9, so the carry does not disappear from the ledger; it changes owner.

---

### 4. `limiter.go`

#### 4.1 What the consumers actually require

Four written requirements, two of which pull against each other.

- A user-present probe outranks a batch import, because the uploader journey holds a contributor on the page through those lookups (`docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md:529-537`).
- Shadow lookups spend deferred background budget through the same limiter and never inline with the probe (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:262-264`, organizer spec `:571-575`).
- Interactive work must not starve timer imports (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:693-694`).
- Rung 9 re-verifies the policy against its real callers (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:565-567`, `:694-696`).

Read together, the first and third forbid both pure fairness and pure precedence. Three classes suffice: user-present, batch, background.

#### 4.2 Comparing the three candidate algorithms

| Algorithm | Satisfies precedence | Satisfies the anti-starvation floor | Cost |
| --- | --- | --- | --- |
| Token bucket per class | Only by over-provisioning | Yes, by construction | Wire capacity is global, so per-class rates either oversubscribe the 1/s budget or strand it: a rate reserved for an idle probe is wire time nobody spends. |
| Weighted fair queue | No | Yes | A probe queued behind a batch's virtual time waits its weighted turn, which is exactly the latency the uploader journey cannot absorb. Also the most state and the most test surface. |
| Strict priority semaphore | Yes | No | A steady arrival of probes starves the batch class indefinitely, which rung 9 forbids in as many words. |

#### 4.3 Recommendation: strict priority with a reserved floor

Build the third with one addition. One global token source paces the wire; a dispatcher hands each token to the highest-priority non-empty queue, except that every Nth token is reserved for the lowest non-empty class. That satisfies both written constraints with one integer of extra state, and the floor is the single knob rung 9 re-verifies rather than a whole policy.

Concretely, and small: an ordered `Class` enum of three values, one `*time.Ticker` paced at the wire rate, one dispatcher goroutine, three queues of waiter channels, a reservation counter. `Wait(ctx, class) error` returns nil on a token, `ctx.Err()` on cancellation.

Per-caller budgets are a concrete type, not a Go interface. The spec's phrase "the per-caller budget interface" names a seam; go-conventions is explicit that an interface waits for a second implementation. Ship:

```go
func New(perSecond int) *Limiter
func (l *Limiter) Budget(name string, class Class, cap int) *Budget
func (b *Budget) Wait(ctx context.Context) error   // ErrBudgetExhausted when cap is spent
func (b *Budget) Spent() int
func Unmetered() *Budget                           // never waits, never exhausts
func (l *Limiter) Close() error
```

The named budget is what makes spend attributable and caps a runaway caller rather than merely slowing it. Rung 4 takes a background-class budget with a per-run cap, which is precisely what "deferred background budget" means operationally. Rung 9 takes a batch budget per import. Rung 7 takes a user-present budget with a small cap and a short context deadline.

#### 4.4 How rung 3's offline runs bypass it, and why they must

`conform` passes `Unmetered()`. The bypass is required rather than convenient: the corpus holds 351 committed requests, and pacing at 1 request per second would add roughly six minutes to a gate whose `scripts/check.sh` runs in 58 to 62 seconds today (`docs/HISTORY.md:196-198`). `Unmetered()` is a named constructor with its own test rather than a nil-means-unlimited check, which would be the defensive-nil pattern go-conventions rejects.

#### 4.5 The clock invariant, which the limiter is the first file to threaten

`checkNoTimeNow` bans `time.Now` anywhere under `internal/engine` (`internal/harness/conform/arch_test.go:95-119`, run over the real tree at `:152-158`, widened to `musicbrainz` in 3b). It matches only a selector call named `Now` on the `time` package, so `time.NewTicker`, `time.NewTimer`, and `time.Since` all pass as the check is written.

Recommend keeping the invariant literal and unamended by giving the limiter no clock read at all. Pacing comes from a ticker the limiter owns; a token is "the next tick", never "the current time minus the last request". Tests replace the ticker channel with a hand-driven `chan time.Time`, which is more deterministic than any injected clock and needs no seam. That is both the faithful reading of the invariant's own stated reason, which is about the evaluation date on the distance path, and the better design.

The fallback, if the plan finds a case the ticker shape cannot express, is a named single-file exemption in `checkNoTimeNow` plus a fixture proving the exemption cannot silently widen. Flag it as a conductor ruling rather than an implementer choice, because it amends a landed architectural check.

---

### 5. How `conform` grades request construction

#### 5.1 What 3c can grade without 3d

The full tier-1 flow needs the ID-match early return and the candidate loop, which are 3d's. Two halves are entirely 3c's.

**The lookup half, 303 committed requests.** For each case, start `stub.New`, call `client.Release(ctx, mbid)` for every distinct release MBID the case's log names, and require `len(srv.Misses()) == 0`. This proves the `inc=` set, the path, `fmt=json`, and the encoding all match the bytes pinned beets actually sent.

**The search half, 48 committed requests.** For each case carrying a search request, derive the search terms through `SearchTerms`, build the query through `BuildAlbumSearch`, issue it through the stub, and require zero misses. Grading through the stub rather than by string comparison is the stronger form: the assertion is that the request resolves, which is the same assertion tier 1 will make.

Both halves report a compared count and assert it against the tree's own totals, following `gradeLedger`'s discipline (`internal/harness/conform/translate.go:318-324`): a resolver that silently skipped rows fails rather than passing quietly. The 167 cases with no search row are reported as such and assert nothing, because whether a case should have searched is the ID-match question 3d owns.

Census read off the tree for the plan's arithmetic: 351 committed requests over 215 cases, 303 lookups and 48 searches, with Lucene fields limited to `release` (48), `artist` (43), and `arid` (5). No committed case exercises any extra tag, so the extra-tag branch of `get_album_criteria` is a **declared corpus-coverage gap**, graded by hand-authored table tests in `musicbrainz`'s own `testdata`. That would be rung 3's tenth declared gap, against the nine HISTORY records at `docs/HISTORY.md:128-131`.

#### 5.2 Mutations that prove the checkers fire

Four, each naming the miss list it produces. Drop `genres` from `releaseIncludes` and every lookup misses. Remove the Lucene escape of `-` and the five `arid` searches miss. Substitute `strings.ToUpper` for the lowercasing pass and all 48 searches miss. Omit `limit` and all 48 miss. Each is a one-line edit with an unambiguous failure.

#### 5.3 The `Extra` channel

**Cassette half, first task.** `Case.Extra` becomes `[]Request`, carrying method, canonical `path_query`, and response file. `validate` applies the same canonical-form and readability checks it already applies to `Requests` (`internal/harness/cassette/cassette.go:152-176`), and the "already consumed" rule extends from a filename collision to a `(method, path_query)` collision. The migration is empty, verified: zero of 215 cases carry an `extra` key. `wantManifestSchemaVersion` moves `"3"` to `"4"` in place at `internal/harness/corpus_test.go:160` and `corpus/manifest.json` bumps with it, following the 2b and 2d precedent and adding no second copy.

Rejected alternative worth writing down: exporting the constant from `cassette` as a package API. That gives a test-only value a permanent exported surface with one consumer, against go-conventions, and the in-place bump is what the two prior schema changes already established.

Two coordination items for the conductor. The constant and the derivation path are lane E1's per the spec's contested-surface table, so the bump needs E1's acknowledgement. And the staleness pre-task should confirm the bump forces no re-freeze: the 2d review found `schema_version` appears only in the manifest and that one constant, and `Extra` is `omitempty` with every case empty, so no `case.json` bytes move.

**Stub half.** `stub.New` builds a second map from `c.Extra`. `ServeHTTP` consults the request map first, then the extra map, and records a consumption rather than a miss when the second answers (`internal/harness/stub/stub.go:81-91`). Add `Consumed() []cassette.Request` in service order, distinct from `Misses()`. Reuse `cassette.Request` rather than minting a third near-identical struct; `stub` already imports `cassette`.

Assertion-discipline proof is a hand-authored case in `stub`'s own `testdata`, never a corpus case, per the proximity precedent. Three rows: a request the `Requests` map answers and `Consumed()` stays empty; a request only `Extra` answers, reported in `Consumed()` and absent from `Misses()`; a request neither answers, in `Misses()` and absent from `Consumed()`.

**Fallback if E1 declines the schema change**, written into the plan rather than implied: the zero-miss rule stands as written, `Extra` stays unreachable through the stub, and the U8 recall-extension surface is recorded as ungraded at rung 3 with the reason.

---

### 6. Closing the `dataSourcePenalty` divergence

The Go fallback is `d.tun.DataSourceMismatchPenalty` (`internal/engine/matcher/penalty.go:404-409`). beets' `get_penalty` falls back to the class literal `MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY`, which is `0.5` (`beets/metadata_plugins.py:160-169`, `:180`). The configured value is a different thing: it is the *plugin's own* penalty, resolved `source_weight` first (`:191-195`) and wired into the plugin's config at init (`:201-202`), which is what populates the per-source map rather than the fallback.

Close it by making the two roles explicit. The fallback becomes a package constant with the citation on it, and `Tunables.DataSourceMismatchPenalty` keeps its job of populating `DataSourcePenalties["MusicBrainz"]`. That is faithful to beets and it makes the field's meaning legible, which the current shape does not.

The mutation that proves the fix fires, hand-authorable and currently silent: set `DataSourceMismatchPenalty` and `DataSourcePenalties["MusicBrainz"]` to `0.7` and call `AddDataSource("Discogs", "Bandcamp")`. beets yields `0.5`; the code as landed yields `0.7`. The branch is unreachable under the greenfield server as built, because the only candidate source is the literal `"MusicBrainz"` (`internal/engine/musicbrainz/translate.go:13`) and the map names it, which is why nothing has caught this.

Scope note for the plan header. This edits `internal/engine/matcher/penalty.go`, outside 3c's stated tree column, and it is in scope because it is 3a's own carry addressed to 3c (`docs/HISTORY.md:333-336`). Three lines plus one table row. Name it as a task so it does not read as accretion. Handing it to 3d instead is defensible and should be a conscious choice rather than a drift.

---

### 7. The `releaseLookupPrefix` triplication

**3c should not unify it, and should say why in writing.**

The count as read off the tree is worse than triple. Three named constants, at `internal/harness/oracle/album_info.go:28`, `internal/harness/analysis/album_info.go:25`, and `internal/harness/conform/translate.go:20`, plus two inline literals at `cmd/harness/capture.go:279` and `internal/harness/synth/synth.go:41`. The engine's client makes a sixth. The same shape holds for the `inc=` set, where `mbrequest.MaximalReleaseInclude` (`internal/harness/mbrequest/mbrequest.go:22-24`) will have an engine-side twin.

Two arguments, pointing the same way.

**The invariant forbids the obvious fix.** `internal/engine` imports nothing under `internal/harness`, checked over the real tree (`internal/harness/conform/arch_test.go:141-146`). `mbrequest` is a harness package, so the client cannot import it. The only legal shared home would be the engine owning the constant and `mbrequest` importing it.

**That legal direction would vacuum the grading.** `mbrequest` is what the capture tool and `synth` use to *produce* a committed request (`cmd/harness/capture.go:279-282`, `internal/harness/synth/synth.go:41-44`). The committed request log is the expectation, and the engine's request construction is the system under test. Sharing one constant between them turns section 5's whole checker into a self-derived expectation, which go-conventions names by that phrase as an anti-pattern. The two copies must be independent, and the conformance run is the mechanism that proves they agree.

So 3c's deliverable here is a written rationale on both sides, not a refactor. HISTORY's own trigger, "worth revisiting if a fourth appears" (`docs/HISTORY.md:357-359`), is met by the engine's copy, and the answer is that the fourth is deliberate and load-bearing. Writing that down is what stops a later `code-simplifier` pass from unifying them and silently emptying the check.

The one cleanup that is genuinely available, and is not 3c's, is folding the two inline literals in `capture.go` and `synth.go` into `mbrequest`'s own constant. File it as a note for the lane that owns `mbrequest`.

---

### 8. Findings the plan should carry as named items

1. `comp` is on 214 of 215 cases' items and absent from `ItemFields`; `va_likely` needs it (`beets/autotag/match.py:454`, `internal/engine/matcher/info.go:124-143`).
2. `matcher.vaArtists` is unexported and the query builder is its second consumer (`internal/engine/matcher/track_artist.go:18-21`).
3. The limiter is the first file under `internal/engine` with a legitimate reason to read a clock; the ticker shape avoids amending the invariant (`internal/harness/conform/arch_test.go:95-119`).
4. The query path inherits all 56 case-mapping divergences rather than the two the distance path keeps (`internal/engine/matcher/testdata/case-mapping-sweep.tsv`).
5. Rung 1b has not landed; `ownedWall` is still a slice literal (`internal/config/keys.go:19-25`), and the registry will need a float parser the current `ParseInt` does not supply (`internal/config/thresholds.go:43-60`).
6. `search_query_ascii` is a structural no-op for MusicBrainz and must be recorded as deliberately outside `Tunables` (`beets/metadata_plugins.py:411-413`, `beetsplug/musicbrainz.py:750`, `:755-762`).
7. Zero committed cases set any extra tag, so `get_album_criteria`'s extra-tag branch is a tenth declared corpus-coverage gap.
8. `wireInt`'s decode errors do not wrap `ErrMalformedRelease`; 3c is the pass whose client makes that boundary real (`docs/HISTORY.md:353-355`).
9. `url.Values.Encode()` sorts keys while cassette sorts whole pairs; identical for today's parameter names, and worth one recorded sentence.
10. Every retry must re-acquire a limiter token, a decision neither beets' structure nor the spec states.

---

## (d) Cross-section consistency and open questions

### Request-shape derivation (a) versus measurement (b)

Every request shape section (a) derives off pinned beets maps onto section (b)'s measured census with
no gap.

- (a)'s two endpoint shapes, release lookup and release search, match (b)'s two shapes exactly: 303
  lookups and 48 searches, 351 total, all `GET` (a) §3, (b) §7.
- (a)'s claim that the recording lookup and browse fallback are unreached is confirmed by (b)'s census:
  zero recording requests, and the corpus maximum of 470 tracks sits under the 500-track browse threshold
  (a) §3, (b) §7 and §5 (`checkNoReleaseReachesBrowseThreshold`).
- (a)'s `RELEASE_INCLUDES` sixteen-element, `+`-joined list matches (b)'s single observed `inc=` value
  byte for byte, 303 of 303 occurrences (a) §3, (b) §7.
- (a)'s Lucene escape set (nineteen characters) maps onto (b)'s measured percent-escape census: `%2B`
  (4545, the includes), `%3A`/`%28`/`%29` (the field wrapper), `%5C` (24, the escaped hyphens and quotes),
  `%27`/`%22`/`%2C` (literal punctuation the pattern does not touch, confirming the "notably absent" list
  in (a) §4) (a) §4, (b) §7.
- (a)'s claim that only two filter shapes appear, `release+artist` and `release+arid`, matches (b)'s
  43-and-5 split exactly (a) §4, (b) §7.
- (a)'s claim that no committed case sets `extra_tags` is unflagged in (b), which never contradicts it;
  (b) independently confirms via the two-parameter-set table that lookups carry only `fmt`+`inc` and
  searches only `fmt`+`limit`+`query` (a) §4, (b) §7.
- (a)'s per-case request-count table (164/3/26/4/9/9) matches (b)'s per-case distribution table exactly,
  row for row (a) §7, (b) §7.
- (a)'s claim that headers, including User-Agent, never appear on the wire is unmeasurable by definition
  from a stub-side artifact; (b) confirms independently that `cassette.Request` carries no header field
  and the stub never reads one (a) §1, (b) §3 and §7.

No request shape in (a) is left unmapped to a (b) measurement, and no (b) measurement lacks a mechanics
explanation in (a).

### Go surface (c) against the beets functions it ports

| (c) proposal | Beets function it ports | (a) citation |
| --- | --- | --- |
| `SearchTerms` | `search_artist`/`search_name` derivation, `va_likely` | `tag_album` (`match.py:407-409, 445-455`) |
| `BuildAlbumSearch` | `get_album_criteria`, `get_search_query_with_filters`, `format_search_term` | `musicbrainz.py:714-734, 736-753`; `_utils/musicbrainz.py:650-663, 679-681` |
| `releaseIncludes` | `RELEASE_INCLUDES` | `_utils/musicbrainz.py:50-67` |
| `Client.Release` | `MusicBrainzAPI.get_release`, `MusicBrainzPlugin.album_for_id`'s lookup half | `_utils/musicbrainz.py:687-690`; `musicbrainz.py:765-800` |
| `Client.SearchReleases` | `MusicBrainzAPI.search`, `SearchApiMetadataSourcePlugin.candidates`/`albums_for_ids` | `_utils/musicbrainz.py:665-685`; `metadata_plugins.py:442-446` |
| `Client.AlbumForID` | `album_for_id`'s full body including the pseudo-release second fetch | `musicbrainz.py:765-800` |
| retry/backoff/timeout policy | `Retry`, `RateLimitAdapter`, `TimeoutAndRetrySession` | `_utils/requests.py:78-135` |
| `ErrNotFound`/`ErrUnavailable` | `HTTPNotFoundError` handling, exhausted-retry `RetryError` | `musicbrainz.py:777-781`; `_utils/requests.py:78-89` |
| `Limiter`/`Budget` | the process-wide `LimiterMixin` + `RateLimitAdapter` composition | `_utils/musicbrainz.py:564-619`; `requests_ratelimiter/requests_ratelimiter.py:96-106` |
| `matcher.IsVariousArtistsName` | `VA_ARTISTS` membership half of `va_likely` | `autotag/distance.py:29`; `autotag/match.py:451-455` |
| `ItemFields.Comp` | the `comp` flag half of `va_likely` | `autotag/match.py:454` |
| `dataSourcePenalty` fallback constant | `MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY` | `metadata_plugins.py:160-169, 180` |

Every (c) proposal traces to a named beets function in (a). No (c) surface is invented without an (a)
counterpart, and no (a) mechanics section describes behavior 3c must build that (c) leaves unassigned,
except the three named coverage gaps (extra tags, singleton/`RECORDING_INCLUDES`, `search_ids`), which
(c) §5.1 and §8 both name as declared gaps rather than silent omissions.

### Corpus figures (c) depends on, restated

- 351 total committed requests, 303 lookups, 48 searches, over 215 cases ((b) §7; relied on by (c) §5.1
  for the lookup-half and search-half grading counts).
- One distinct `inc=` value across all 303 lookups, byte-identical to `mbrequest.MaximalReleaseInclude`
  ((b) §7; relied on by (c) §2.3 and §7 for the includes-collapse argument).
- 46 distinct search queries, filter shapes `release+artist` (43) and `release+arid` (5), zero exercising
  `extra_tags` ((b) §7; relied on by (c) §2.1, §2.2, and §5.1 for the declared-gap claim).
- Zero committed cases carry a `Comp`-bearing decode today because the field does not exist yet, but 15
  cases carry `comp` in their raw item JSON ((a) §5; relied on by (c) §2.1 and §8 item 1).
- Zero committed cases carry an `Extra` entry ((b) §1, §7, §9; relied on by (c) §5.3 for the empty-migration
  argument).
- Corpus maximum of 470 tracks against the 500-track browse threshold ((b) §5, quoting 3b's guard; relied
  on by (a) §3 and (c) §1.1 for the browse-fallback non-port decision).
- 104 MB / 77.6 MB response-body corpus weight, largest single body 8,894 KB ((b) §7; relied on by (c) §4.4
  for the `Unmetered()` cost argument alongside the 983 s oracle-suite figure from (a) §2).

### Deduplicated open questions

Fourteen distinct questions after merging near-duplicates across the three raw lists (the raw (a), (b),
and (c) lists totaled twenty-one entries; seven were the same question restated by a different section
and are merged below with all sources named).

1. **Does 3c deliver request-construction grading only, or the request-set expectation too?** The set is
   a function of the ID-match early return, `_recommendation`, and `_add_candidate`, all 3d scope by the
   pass-split table. 3c can grade "given these inputs, this exact `path_query`" today; it cannot grade
   "these three requests and no others" without 3d's flow. — raised in (a) §7 and (c) §5.1.
2. **Does the `time.Now` ban under `internal/engine` extend to the limiter's clock read, or was it scoped
   only to the evaluation-date rule it was written for?** Three routes: inject a clock seam (recommended),
   pace on a ticker with no clock read at all (also recommended, and arguably the same answer stated two
   ways), or narrow the walk's root back to `internal/engine/matcher` with a stated reason (undoes 3b's
   R21 widening; needs a conductor ruling). A green walk today is not proof the limiter is clock-free,
   since `time.Sleep`/`time.Ticker` pass unexamined. — raised in (a) §11, (b) §10 item 1, (c) §4.5.
3. **Who owns `internal/harness/corpus_test.go`?** Reshaping `Case.Extra` breaks
   `TestEveryResponseFileIsConsumed` at `:133-135` and moves `wantManifestSchemaVersion` at `:160`, both
   in a file the 3c tree column (`internal/harness/{conform,stub}`) does not name. — raised in (b) §1
   item 3 and §8.
4. **Does the `Case.Extra` reshape bump `schema_version` from `"3"` to `"4"`, and does lane E1 accept it?**
   By the 2b/2d precedent it is a `Case`-field change and therefore a bump, but forces no re-freeze
   because zero cases carry an `Extra` and no `Expected` field moves. The bump edits
   `corpus/manifest.json`, which the contested-surface table places with lane E1, so the plan needs E1's
   ruling before Task 1, and a written fallback if E1 declines. — raised in (a) §10, (b) §9 and §10 item 4,
   (c) §5.3.
5. **Is the empty-migration argument sufficient proof, or does the conductor want a `make derive` run to
   confirm byte-identity the way 2d did?** A re-marshal-and-diff test over the 215 committed files proves
   the same fact offline in seconds without touching the venv; 2d's own precedent used a full re-derive.
   The plan should name which proof it is offering. — raised in (b) §9 (open question list) and (c) §5.3.
6. **Should tier 1 also grade that every logged request was consumed (a symmetric `Consumed()`/hit-tracking
   report on the `Requests` map, not only on `Extra`)?** The spec deliberately grades only zero-misses,
   because a strong ID match legally orphans search captures into `Extra`; the addition is nearly free
   given the map and mutex already exist, and it is the difference between grading request construction
   and grading only request over-reach. — raised in (a) §10 finding 4, (b) §3 and §10 item 4.
7. **Who becomes `Case.Extra`'s producer once it carries `Request` shape?** Nothing writes it today. If
   the intent is that the capture-derive loop eventually orphans a retired search automatically, that is
   `cmd/harness` work in lane E1's territory, not 3c's, and the handoff should be named now. — raised in
   (a) §10 finding 5, (b) §9 (open question list) and §10 item 5.
8. **Who adds `ItemFields.Comp`, and does it join the shared type with its doc comment amended, or does
   the query builder take a narrower private input type?** `va_likely` reads it, 15 corpus cases set it,
   and adding it means 3c writes one `internal/engine/matcher` file the pass split does not anticipate.
   Same question, smaller, for exporting `vaArtists`/`IsVariousArtistsName`. — raised in (a) §5, (c) §2.1
   and §8 items 1-2.
9. **Does 3c collapse `mbrequest.MaximalReleaseInclude` (and the `releaseLookupPrefix` triplication) into
   an engine-owned constant, or accept a fourth-plus copy of the `/ws/2` literals?** The arch rule permits
   the collapse (harness may import engine), but collapsing would let the request-log expectation and the
   client's construction share a definition, which is exactly the self-derived-expectation anti-pattern
   the conformance run exists to avoid. (c) §7 recommends declining and writing the rationale down; the
   raw question lists frame it as still open. — raised in (a) §12, (b) §10 item 8, (c) §7 and §8 item 9.
10. **Does 3c's limiter subsume `cmd/harness/capture.go`'s bare `sleep(requestInterval)` pacer, or
    coexist beside it?** The capture tool is a live spender of the same MusicBrainz budget, sits in
    `cmd/` outside 3c's declared territory, and has no retry and a 30 s timeout against beets' 10 s.
    Coexistence looks right for this pass but the plan should say so. — raised in (a) §13.
11. **Which document loses on the `search_limit` legacy spelling, and when?** The truth is
    `musicbrainz.match.searchlimit` (probe-confirmed against `SP/beetsplug/musicbrainz.py:292-294` and
    `SP/beets/plugins.py:200`). The rung 3 spec's enumeration, `docs/HISTORY.md:597-599` (3a's
    correction #5), and `internal/engine/matcher/tunables.go:119-124` all say something else. — raised in
    (a) §0 item 2.
12. **How are the corpus-unreachable client surfaces graded, and where do the fixtures live?** The list
    is longer than M8's three: the whole `extra_tags` criteria branch (including `tracks`, `alias`, and
    the catalognum space-strip), `RECORDING_INCLUDES` and the singleton path, the browse fallback, the
    explicit `search_ids` branch, and the no-id/no-consensus branch of `match_by_id`. Each needs a
    hand-authored fixture in its owning package's own `testdata` plus a named mutation, on the proximity
    precedent, and each needs recording as a declared coverage gap. — raised in (a) §4 and §7, (c) §5.1
    and §8 item 7.
13. **Does 3c reproduce beets' retry semantics exactly (the 0/1/2/4/8/16s schedule, `Retry-After`
    honored, the bypass-both-limiters-on-retry behavior, the `RetryError`-not-`HTTPError` shape), or state
    a deliberate divergence?** Rung 3's runs are offline, so nothing in the corpus can grade any of it
    either way; this is a design call with no conformance evidence on either side. — raised in (a) §1 and
    §7, (c) §3.2 and §3.3.
14. **Should the query builder mirror Python's full case mapping through `golang.org/x/text/cases`
    (accepting one new dependency), or record the 56 divergent code points as a declared divergence?**
    The corpus exercises the non-ASCII search path at least once, so this is reachable rather than
    theoretical, unlike the distance path's already-rescued 54 of 56. — raised in (a) §4 and (c) §2.5.
