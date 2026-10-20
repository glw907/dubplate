# Rung 3, Pass 3d Planning Inventory

Research report, 2026-09-05, prepared for pass 3d's plan authoring,
against HEAD `c7a85fde96d83d6410de165d89ba4517f9c5e541`. Scope, quoted
from the rung 3 spec's pass-split table
(`docs/specs/2026-09-03-rung-3-go-matcher.md:331`):

> **3d** Verdicts and the gate | The assignment solve with tie
> mirroring; extras and their sort orders; admission filters, insertion
> order, stable sort, the ID-match early return and the search fallback
> flow; recommendation, `max_rec` downgrade, verdict-to-action; the
> compat switches; the whole-corpus conformance runner and its tier 1
> and tier 2 gates including the pending-divergence path;
> `TieEntry.Adjudicator` and its sentinel check (M4); `CheckOracleBugs`'s
> probe-case rule (B3); the matcher's public entry point. | 3a, 3b, 3c.
> | Main tree: writes `internal/harness/conform` and
> `internal/harness/analysis`.

Three sections follow: (a) beets' verdict mechanics and the assignment
solver, (b) the harness ledgers and tiers as landed, (c) the engine as
landed and the proposed 3d Go surface. A fourth section, (d), reconciles
all three and lists the open questions each raised.

---

## (a) Beets verdict mechanics and the assignment solver

Method: every claim below is read off pinned beets 2.13.1 in
`tools/venv/.venv/lib/python3.13/site-packages/`, off the committed
tree, or off an offline probe run with `tools/venv/.venv/bin/python`. No
network, no `make derive`, no response body read into context (probe
scripts read them; I did not). The probes live in
`/tmp/claude-1000/-var-home-glw907-Projects-dubplate/84407168-352a-456b-9b1c-a2686b80f7f3/scratchpad/probe_*.py`
and are reproducible; the plan should re-run rather than trust them.

Everything the port map
(`docs/research/2026-09-02-beets-matcher-mechanics.md`)
says about assignment, extras, admission, recommendation, and flow is
confirmed against source. The extensions below are what it does not
say.

### 1. `assign_items` and the cost matrix

`beets/autotag/match.py:53-81`.

- Rows are items in the caller's order, columns are `info.tracks` in
  the source's order: `costs = [[float(track_distance(i, t)) for t in
  tracks] for i in items]` (`match.py:64`).
- **Which distance.** `track_distance(item, track_info,
  incl_artist=False)` is the default (`beets/autotag/distance.py:383-385`).
  `float()` is `Distance.__float__`, which returns `self.distance`, which
  divides `raw_distance` by `max_distance`, or `0.0` when `max_distance`
  is zero (`distance.py:152-160, 218-219`). Each cell is therefore
  **normalized against its own denominator**: which track components
  fired for that pair sets that pair's `max_distance`, so cells are not
  on a common scale. The Go port must call
  `matcher.DistanceForTrack(tun, item, info, false)` and take
  `.Distance()`, never a raw sum.
- **The solve.** `lap.lapjv(np.array(costs), extend_cost=True)`
  (`match.py:66`). The third return value is indexed by **track** and
  holds an item index or `-1` (`match.py:69-71`).
- **The mapping's iteration order is by track, not by item.**
  `mapping = {items[iidx]: t for iidx, t in
  zip(assigned_item_idxs, tracks) if iidx != -1}` (`match.py:72-76`)
  zips over the track-indexed result, so the dict's insertion order is
  ascending track index. `AlbumMatch.mapping = dict(item_info_pairs)`
  preserves it (`match.py:378-380`), `item_info_pairs` reads it back
  (`match.py:157-160`), and `distance()` walks it to build both
  `dist.tracks` and the `tracks` penalty list (`distance.py:527-530`).

  **Verified against the corpus, not inferred.** For `should-ask` and
  `should-reject`, the committed `Expected.Candidates[0].Tracks` array
  equals the per-track distances taken in **track order** and differs
  from the same values taken in item order (probe `probe_ties.py`, both
  cases). This matters twice: tier 1 grades `Candidate.Tracks`
  elementwise, and `Distance.raw_distance` sums the `tracks` penalty
  list left to right (`distance.py:171-176`), so a Go port feeding
  `[]AlbumPair` in item order both reports the wrong vector and can sum
  to a different double. `matcher.addAlbumTracks`
  (`internal/engine/matcher/album_tracks.go:25-31`) iterates the
  caller's `pairs` slice verbatim, so **3d's assignment must hand
  `DistanceForAlbum` its pairs sorted by track index.**

- **Rectangular matrices.** `assign_items` never returns both a
  non-empty `extra_items` and a non-empty `extra_tracks`: the leftover
  set is one-sided. Confirmed empirically. `lapjv` matched exactly
  `min(rows, cols)` pairs on every one of 4000 random rectangular
  probes (`probe_model.py`, `probe_extend3.py`) and on all 214 corpus
  candidates. `corpus/README.md:132` states the same and says a
  two-sided case is not an expressible shape.

- **Corpus reach.** 214 candidate-bearing cases, 294 candidates. 87 of
  the 294 candidates are rectangular (57 carry `missing_tracks`, 30
  carry `unmatched_tracks`); 32 cases have at least one rectangular
  candidate; 21 cases are rectangular on their best candidate
  (`probe_census.py`, `probe_all.py`). The extension path is heavily
  exercised, not incidental.

### 2. `extra_items` and `extra_tracks`, and their sort orders

`match.py:77-81`.

```
extra_items  = list(set(items)  - mapping.keys()); extra_items.sort(key=lambda i: (i.disc, i.track, i.title))
extra_tracks = list(set(tracks) - set(mapping.values())); extra_tracks.sort(key=lambda t: (t.index, t.title))
```

Three facts the port map does not carry.

1. **Both pre-sort orders are hash-set orders over identity-hashed
   objects.** `Item` inherits `object.__hash__`/`object.__eq__`
   (verified at runtime: `Item.__hash__ is object.__hash__` is `True`),
   and `AttrDict.__hash__` returns `id(self)` (`beets/autotag/hooks.py:131-132`).
   So set iteration order derives from **addresses**, which
   `PYTHONHASHSEED=0` does not control. `derive_oracle.py:282-287`
   justifies the seed by "two extras with equal sort keys order by
   hash"; the seed stabilizes `str` hashing, not object identity. The
   committed extras are reproducible only because the sort keys are
   total.
2. **`extra_tracks`' key is provably total.** `album_info` renumbers
   every track `1..n` across the whole release
   (`beetsplug/musicbrainz.py:645-646`), so `t.index` is unique and
   `(index, title)` can never tie. Go can sort on index alone and be
   exactly right, but should sort on the pair to stay a visible port.
3. **`extra_items`' key is total only as a fact of today's corpus.** No
   committed case's best-candidate `extra_items` contains two items
   sharing `(disc, track, title)` (`probe_census.py`: zero cases). A
   future case that did would make beets' own record non-reproducible
   run to run. 3d should ship a harness guard asserting the absence, so
   the corpus lane learns about it at landing rather than at a flaky
   gate.

Go note: `sort.Slice` is not stable and `slices.SortFunc` is not
stable; use `slices.SortStableFunc` with the three-key comparator so
the port stays faithful if a tie ever appears.

### 3. Admission filters, insertion order, stable sort, dedup

`_add_candidate` (`match.py:331-380`), in this exact order:

| Step | Rule | Source |
| --- | --- | --- |
| 1 | Drop zero-track candidates: `if not info.tracks: return` | `match.py:345-347` |
| 2 | Dedup: `if info.album_id and info.identifier in results: return`. `identifier` is `(data_source, id)`; `AlbumInfo.id` is `album_id` | `match.py:350-352`; `hooks.py:165-168, 291-293` |
| 3 | `match.required`: for each tag, `if getattr(info, req_tag) is None: return` | `match.py:354-359` |
| 4 | **Solve the assignment** | `match.py:361-364` |
| 5 | Score `distance(items, info, item_info_pairs)` | `match.py:366-367` |
| 6 | `match.ignored`: `penalties = [key for key, _ in dist]`, drop if any ignored tag is in it | `match.py:369-375` |
| 7 | Insert at `results[info.identifier]` | `match.py:378-380` |

Four things worth the plan's attention.

- **Order is load-bearing.** A candidate rejected by `ignored` has
  already paid for a solve, and a candidate rejected by `required` has
  not. Only step 7 affects the record, but a port that reorders 3 and 4
  changes nothing observable and a port that reorders 2 and 1 does (a
  zero-track duplicate).
- **`ignored` reads the NONZERO view.** `for key, _ in dist` is
  `Distance.__iter__`, which is `items()`, which drops zero-weighted
  penalties (`distance.py:190-205, 240-241`). This is the m5 hazard's
  **second site**, alongside `max_rec`. `Candidate.Penalties` in the
  record is the `_penalties` (added-key) view
  (`internal/harness/oracle/derive_oracle.py:176`), so a port that
  wires `ignored` to `AddedKeys`/`PerKey` instead of
  `NonzeroKeys` (`internal/engine/matcher/penalty.go:124, 133, 166`) is
  wrong and tier 1 will not catch it, because `match.ignored` is
  corpus-unreachable.
- **`required` has no faithful Go semantics as the types stand.**
  `getattr(info, req_tag) is None` distinguishes "field absent/None"
  from "field empty". Verified at runtime: `getattr(AlbumInfo(...),
  "label")` is `None`; `getattr(..., "nosuchfield")` **raises
  `AttributeError`**, which propagates out of `tag_album` and crashes
  the import. Beets has no guard. Meanwhile `matcher.AlbumInfo`
  (`internal/engine/matcher/info.go:10-68`) carries 14 plain-typed
  fields and no nil, which is exactly 3a's R1 ruling ("an absent source
  field becomes the plain-string zero value", 3b guard 2). So Go cannot
  express `is None` at all, and cannot name the ~25 beets `AlbumInfo`
  fields it does not model. 3d must rule this explicitly: which field
  names `required` accepts, what "None" means per field kind (`""`?
  `0`? `false`?), and what an unknown name does (error at config load
  beats beets' crash). Open question (see (d), item 3).
- **Insertion order and the stable sort.** `candidates` is an
  insertion-ordered dict (`match.py:414`); ID-derived candidates go in
  first (`match.py:425-428`), then search results in the search
  response's order (`metadata_plugins.py:442-446` maps ids through
  `albums_for_ids`, `:250-259`, preserving order; `filter(None, ...)` at
  `:108` drops failed lookups without reordering). `_sort_candidates`
  is `sorted(..., key=lambda m: m.distance)` (`match.py:326-328`),
  Python's stable Timsort, comparing exact doubles through
  `Distance.__lt__`/`__eq__` (`distance.py:210-216`).

  Corpus reach: **exactly two cases carry a distance tie among
  candidates**, `should-ask` (three candidates at
  `0.23868312757201643`) and `should-reject` (five at
  `0.3891203703703704`), and in both the tie group's recorded order
  equals the order the case logged its release lookups
  (`probe_order.py`). So the stable sort **is** graded, thinly. What is
  not graded: the ID-candidate-before-search-candidate ordering at
  equal distance, and multi-source interleaving (single source by
  design, `corpus/README.md:2430`).

  Go note: `slices.SortStableFunc` with `cmp.Compare` on the float.
  `sort.Slice` here would be a silent defect that passes 213 of 215
  cases.

### 4. The ID-match early return, `match_by_id`, and the search fallback

`tag_album` (`match.py:383-468`).

```
likelies, consensus = get_most_common_tags(items)          # util/__init__.py:810-846
cur_artist, cur_album = likelies["artist"], likelies["album"]
candidates = {}                                             # insertion-ordered
if search_ids:                                              # explicit-ID branch, no early return
    for info in albums_for_ids(search_ids): _add_candidate(...)
else:
    for info in match_by_id(likelies["mb_albumid"], consensus["mb_albumid"]): _add_candidate(...)
    rec = _recommendation(list(candidates.values()))         # NOT sorted
    if candidates and not config["import"]["timid"] and rec == strong:
        return cur_artist, cur_album, Proposal(list(candidates.values()), rec)   # NOT sorted
    if not (search_artist and search_name): search_artist, search_name = cur_artist, cur_album
    va_likely = (not consensus["artist"]) or (search_artist.lower() in VA_ARTISTS) or any(item.comp for item in items)
    for c in metadata_plugins.candidates(items, search_artist, search_name, va_likely): _add_candidate(...)
candidates_sorted = _sort_candidates(candidates.values())
return cur_artist, cur_album, Proposal(candidates_sorted, _recommendation(candidates_sorted))
```

- **`_recommendation` is called on an UNSORTED list at `match.py:430`,
  and the early return ships that unsorted list at `:441`.** Harmless
  in practice because `match_by_id` passes a one-element id list
  (`match.py:264`) so at most one candidate exists, but a port that
  sorts before the early return has changed the code and should say so.
- **"Consensus" is unanimity of the plurality value.**
  `consensus[field] = freq == len(values)` where `freq` is
  `Counter.most_common(1)[0][1]` (`util/__init__.py:837-840`,
  `plurality` at `:799-807`). `match_by_id` returns `()` when the id is
  falsy or consensus is `False` (`match.py:252-266`), logging which.
- **`_extract_id` sits between the id and the request.** `album_for_id`
  runs `extract_release_id("musicbrainz", id)` against
  `re.compile(r"(\w{8}(?:-\w{4}){3}-\w{12})")`
  (`beetsplug/musicbrainz.py:771-773`; `beets/util/id_extractors.py:20`)
  and returns `None` on no match, dropping the candidate. Python's
  `\w` is Unicode-aware; Go's RE2 `\w` is ASCII-only. 3c already
  carries this as one of its three regime divergences and ships
  `releaseIDPattern` at `internal/engine/musicbrainz/client.go:65`.

**Corpus coverage of this flow, measured (`probe_flow.py`, 215
cases):**

| Path | Cases | Note |
| --- | --- | --- |
| ID-match early return taken | **167** | all 167 no-search cases; all `strong`/`apply` |
| Fell through to text search | **48** | none of the 48 ends `strong` |
| `match_by_id` no-id branch | **0** | every item of every case carries an `mb_albumid` |
| `match_by_id` no-consensus branch | **0** | every case's items are unanimous |
| explicit `search_ids` branch | **0** | no corpus case sets it |
| `import.timid` gate | **0** | `timid: no` in every profile |
| `strong` reached after a search | **0** | no-search ⟺ `strong`, exactly |

So four of the flow's branches are corpus-unreachable and are 3d's new
declared gaps. Note also that `import.search_ids`
(`beets/config_default.yaml:50`) is a real config key that is **not**
in `matcher.Tunables` and not in the spec's tunable enumeration; it
belongs on the public entry point as a parameter, beside `evalDate`.

### 5. `_recommendation`, its thresholds, and the `max_rec` downgrade

`match.py:269-323`. Every config read, in order:

| Read | Default | Site |
| --- | --- | --- |
| `match.strong_rec_thresh` | `0.04` | `match.py:285`, `config_default.yaml:167` |
| `match.medium_rec_thresh` | `0.25` | `match.py:288`, `:168` |
| `match.rec_gap_thresh` | `0.25` | `match.py:296`, `:169` |
| `match.max_rec` | `{missing_tracks: medium, unmatched_tracks: medium}` | `match.py:310-320`, `:170-172` |

The ladder:

```
no results                                  -> none
d  <  strong_rec_thresh                     -> strong      # STRICT <
d  <= medium_rec_thresh                     -> medium      # <=
len(results) == 1                           -> low
results[1].distance - d >= rec_gap_thresh   -> low
otherwise                                   -> none, RETURNED IMMEDIATELY (skips the downgrade)
```

- The `<` / `<=` asymmetry is the proximity flag's whole reason (spec
  hazard 3), and `rec-strong-medium-boundary` is the corpus's single
  flagged proximity row (`corpus/analysis/proximity.json`: 1 flagged, 1
  skipped, 213 clear).
- `results[1].distance - d` resolves to `Distance.__sub__` (via
  `Distance.__rsub__`, `distance.py:221-225`): plain `float - float`.
  No special arithmetic.
- **The downgrade key set is the NONZERO view.** `keys =
  set(min_dist.keys())` and `keys()` is built from `items()`
  (`distance.py:246-247`), which drops zero-weighted penalties. A key
  added at `0.0` never downgrades. This is the spec's m5 hazard, first
  site.
- **The per-track half of the key set is dead in this corpus.** `for
  track_dist in min_dist.tracks.values(): keys.update(list(track_dist.keys()))`
  (`match.py:307-309`) only matters if `max_rec` names a track-level
  key, and it never does under the shipped default. Verified: a model
  that computes the downgrade from **album-level nonzero keys only**
  reproduces all 214 recorded recommendations with zero mismatches
  (`probe_rec.py`). So deleting that loop leaves the corpus green;
  3d needs a hand-authored fixture with a `max_rec` entry naming a
  track key (`track_length`, say) to prove it fires.
- `min(rec, max_rec)` (`match.py:321`) relies on `Recommendation`'s
  `IntEnum` ordering `none=0 < low=1 < medium=2 < strong=3`
  (`match.py:39-47`), and `as_choice` raises on any value outside those
  four (`match.py:313-320`). `Tunables.MaxRec` is `map[string]string`
  (`internal/engine/matcher/tunables.go:67`), so 3d owns the validation
  beets gets from confuse. Iteration over the Python `set` is
  unordered, but `min` is commutative and associative over the enum, so
  no ordering hazard exists here.

**Branch census, computed from the committed records
(`probe_rec.py`):**

| Pre-downgrade branch | Cases |
| --- | --- |
| `strong` | 173 |
| `medium` | 33 |
| `low` (single candidate) | 3 (`assign-large-garbled`, `real-nonlatin-hamasaki-ayumi`, `rec-single-candidate-low`) |
| `low` (gap) | 1 (`rec-gap-low`) |
| `none` (immediate return) | 4 (`rec-none-small-gap`, `rich-album-id-recovers-clean`, `rich-album-id-recovers-country`, `should-reject`) |
| `none` (no candidates) | 1 (`admit-zero-track-dropped`) |

All six branches are reached. **The `max_rec` downgrade fires on
exactly 6 cases, all `strong` → `medium`**: `assign-extra-item`,
`complete-combined-short-scrambled`, `complete-max-rec-boundary`,
`complete-short-one`, `real-noise-pink-floyd-de1994`,
`rec-max-rec-missing-tracks-downgrade`. 173 − 6 = the 167 recorded
`strong`.

**All 215 cases echo the shipped defaults**, `[0.04, 0.25, 0.25]` and
`{missing_tracks: medium, unmatched_tracks: medium}`, no variation
(verified by `jq` over `corpus/cases/*/case.json`). Tier 1's M1
cross-check therefore proves the transcription, never the behavior
under a different value; every non-default threshold path is a
declared gap covered only by package unit tests.

### 6. Verdict to action

`_summary_judgment`
(`beets/ui/commands/import_/session.py:328-360`), executed rather than
restated by the shim (`internal/harness/oracle/derive_oracle.py:193-206`):

```
if import.quiet:
    if rec == strong: return APPLY
    action = import.quiet_fallback  in {skip -> SKIP, asis -> ASIS}
elif import.timid: return None                      # ask
elif rec == none: action = import.none_rec_action in {skip -> SKIP, asis -> ASIS, ask -> None}
else: return None                                    # ask
```

`Action` values are `"APPLY"`, `"SKIP"`, `"ASIS"`
(`beets/importer/actions.py:8-19`); the shim lowercases them and maps
`None` to `"ask"` (`derive_oracle.py:206`).

The corpus's dispositions encode exactly two of them: **`apply` 167,
`skip` 48, nothing else** (`probe_flow.py`).
`derive_oracle.py:204` forces `config["import"]["quiet"] = True` on
every derivation, so:

- `quiet=False` is corpus-unreachable; the entire `timid` /
  `none_rec_action` / "ask" half of the function is ungraded;
- `quiet_fallback: asis` is corpus-unreachable (estate and locale set
  `skip`; stock takes beets' default `skip`, `config_default.yaml:33`);
- `none_rec_action` is unreachable twice over: unreachable under
  quiet, and its default `ask` never fires.

`cassette.Expected.Action`'s comment already says "executed via beets'
own quiet-mode mapping"
(`internal/harness/cassette/cassette.go:78`). 3d ports the whole
function, grades two of its five outcomes against the corpus, and
covers the rest with table-driven unit tests over
`Tunables.Quiet`/`QuietFallback`/`NoneRecAction`/`Timid`
(`tunables.go:168, 186, 191, 196`), each naming its firing mutation.

### 7. Every compat switch beets exposes that changes a verdict

The port map's Q4 list is correct and complete against source, with two
corrections and one addition worth stating.

**Under `match:`** (`config_default.yaml:166-206`): `strong_rec_thresh`,
`medium_rec_thresh`, `rec_gap_thresh`, `max_rec`, `distance_weights`
(**twenty** keys at `:173-193`, settled at 3a; the study's Q4 "19" is
the superseded document), `preferred.countries` (two regimes: the
`add_priority` at `distance.py:496-499` and release-event selection at
`beetsplug/musicbrainz.py:165-177`), `preferred.media`,
`preferred.original_year`, `ignored`, `required`, `ignored_media`,
`track_length_grace`, `track_length_max`, `ignore_data_tracks`,
`ignore_video_tracks`.

**Outside `match:`**: `musicbrainz.search_limit` (default 5,
`metadata_plugins.py:199-204`, read at `:415`),
`musicbrainz.data_source_mismatch_penalty` / legacy `source_weight`
(default `0.5`, `metadata_plugins.py:180, 190-195`),
`musicbrainz.extra_tags`, `import.timid`, `import.languages`,
`import.ignored_alias_types`, the `plugins` list (which also flips
`add_data_source`'s multi-source condition at `distance.py:357-361`,
via `find_metadata_source_plugins` at `metadata_plugins.py:46-51`), and
the verdict-to-action trio `import.quiet` / `quiet_fallback` /
`none_rec_action`. Add `import.search_ids`
(`config_default.yaml:50`), which selects the explicit-ID branch and
is on no list today.

`matcher.Tunables` (`internal/engine/matcher/tunables.go:42-...`)
already carries every one of these except `search_ids`, so 3d adds no
field for the switch surface itself.

**Which pattern table the string-distance compat switch selects.** 3b
built three tables and left the selection to 3d
(`tunables.go:348, 365, 388`):

| Table | Contents | Ledger class |
| --- | --- | --- |
| `BeetsSDPatterns` | beets' shipped six (`distance.py:35-42`) | full compat |
| `SDPatternsWordBoundary` | the six with `\b` prefixed to the featuring pattern | `bug-ft-pattern`, six `adv-ft-*` rows |
| `SDPatternsNoBracketDiscount` | the six less the two `0.3` bracket rows | `bug-remix-parenthetical`, four `adv-remix-*` rows |

**This is a design blocker, not a detail.** `Tunables.StringDistPatterns`
is one slot, and the two string-heuristic hypotheses are *both* now
ruled `oracle-wrong` (see §9), so ruling A's "switch defaults off"
would require both fixes at once, and no committed table carries both.
3d must either build a fourth table (`\b` featuring **and** no bracket
rows) or represent the compat surface as independent per-class edits
composed into a table at run time. The `\b` regime divergence rides
with it: Python `re`'s `\b` is Unicode-aware, Go RE2's is ASCII-only,
benign today because all six `adv-ft-*` pairs are ASCII (3b guard 11).

### 8. The two `inclArtist` call shapes

Beets calls `track_distance` with two different objectives inside one
candidate's scoring, and the two disagree exactly on VA releases:

- **Assignment objective:** `incl_artist=False` for every cell
  (`match.py:64`, the parameter default at `distance.py:384`).
- **Final per-track objective:** `incl_artist=album_info.va`
  (`distance.py:529`), and it is these values that land in
  `dist.tracks`, in the `tracks` penalty list, and in the record's
  `Candidate.Tracks`.
- (Singletons pass `incl_artist=True`, `match.py:494, 519`; out of
  rung 3's scope.)

3a already made this a parameter rather than a config read
(`internal/engine/matcher/track.go:45`, 3a guard 7), and
`addAlbumTracks` already passes `info.VA` (`album_tracks.go:27`). **What
3d owes is the first shape**: the assignment solve must build its
matrix with `inclArtist=false` unconditionally, including on VA
releases. `derive_ties.py:43-55` states the same rule in its own
docstring, and the corpus's tied cases are both non-VA (`should-ask`,
`should-reject`), so **no corpus case discriminates the two objectives
inside the assignment**; a Go port that passed `info.VA` into the cost
matrix would pass every committed case. That is a hand-fixture
obligation: a VA release whose track artists move the optimum.

### 9. What the ledgers say today (the spec is stale here)

Read off the tree, and it has moved since the spec was written
(`corpus/analysis/oracle_bugs.json`, `adjudications.json`):

| Hypothesis | Disposition | Probes | All probes resolve to an adjudication row? |
| --- | --- | --- | --- |
| `bug-ft-pattern` | **oracle-wrong** | 6 | yes, all 6 `oracle-wrong` with `defect_free` |
| `bug-remix-parenthetical` | **oracle-wrong** | 4 | yes, all 4 `oracle-wrong` with `defect_free` |
| `bug-nonlatin-alias-selection` | **oracle-wrong** | 4 | yes; 2 rows are `oracle-correct`, 2 `oracle-wrong` |
| `bug-assignment-rejection-gap` | **oracle-wrong** | 6 | yes, all 6 `oracle-wrong`, **no `defect_free`** |
| `bug-threshold-boundary` | oracle-correct | 1 | n/a |
| `bug-locale-alias-ignored` | oracle-correct | 2 | n/a |

`adjudications.json`: 21 rows, 18 `oracle-wrong`, 3 `oracle-correct`,
**zero pending**. `ties.json`: 215 rows, 2 `pinned`, 0 `excluded`, 1
`skipped`. `proximity.json`: 1 flagged, 1 skipped.

Three consequences the plan must absorb.

1. **The spec's B3 defect is already closed by the corpus.** It names
   three probe cases with no adjudication row; today every probe of
   every `oracle-wrong` hypothesis resolves by `subject.case_id` (per
   this section's own verification). The `CheckOracleBugs` rule
   (`internal/harness/analysis/oraclebug.go:55`, signature `(caseIDs
   []string, adjudicationIDs map[string]bool, ...)`) is still owed, and
   it will pass on landing, so its assertion-discipline proof **must**
   be a hand-authored `testdata` fixture, exactly as the spec requires.
   The plan should say the rule fires on nothing in the real corpus.
2. **`bug-assignment-rejection-gap` is ruled `oracle-wrong` and has no
   gradeable expectation.** Its six corrected verdicts are prose
   (`adj-adv-assign-swap-session-tape`: "identity assignment
   {0:0,1:1,2:2} ... would-be distance 0.052222, medium, skip"), and
   `CheckDefectFree`'s predicate `StringHeuristicHypotheses`
   deliberately excludes the class
   (`internal/harness/analysis/adjudication.go:86, 218`). Producing
   those verdicts means **inventing a pair-rejection mechanism beets
   does not have**, new behavior, which "improvements never ride the
   port" forbids. Read literally, the spec's grading table fails 3d on
   six cases. This is the largest open question (see (d), item 1).
3. **Ruling A's two-run shape is now mandatory, not optional.** With
   both string-heuristic classes at `oracle-wrong`, a default-mode run
   scores every case against a non-beets pattern table and diverges
   from every committed distance. Tier 1 (bit-identity) must run
   **full compat**; the corrected verdicts are graded in a second,
   default-mode run restricted to the ledgered probe cases. The plan
   should write that as two named runs rather than leaving it implied.

**M4's mechanism, as the tree stands.** `TieEntry` has no
`Adjudicator` field (`internal/harness/analysis/ties.go:56-66`), and
`CheckTieLedger` accepts `excluded` and `pinned` interchangeably
(`ties.go:83, 130-136`). Adding the field touches the oracle-tagged
derivation too: `writeTieLedger` carries only `Disposition` and
`Reason` forward (`internal/harness/oracle/ties_derive_test.go:133,
139`), and the comparison masks the same two fields (`:91`). Both are
`//go:build oracle`, so the gate can neither run them nor prove the
carry-forward. 3d should extract the merge into an untagged helper with
its own test, or the field silently vanishes on the next `make derive`.

### 10. The solver: `lap.lapjv` 0.5.13

**It is a C++ extension, not Python.** `lap/__init__.py:16-22`
re-exports from `lap/_lapjv.cpython-313-x86_64-linux-gnu.so` (946 KB).
The venv ships no solver source. `nm -D` on the object gives the dense
path's own symbols: `lapjv_internal`, `_ccrrt_dense` (column reduction
plus reduction transfer), `_carr_dense` (augmenting row reduction),
`_ca_dense` / `find_path_dense` / `_find_dense` / `_scan_dense`
(Dijkstra augmentation), and `strings` names `lapjv.cpp` and
`_lapjv.cpp`. That is the classic Jonker–Volgenant three-phase dense
solver. `lap.LARGE = 1000000`, `FP_1/FP_2/FP_DYNAMIC = 1/2/3` (dense
path uses none of the fp modes; those are `lapmod`'s).
`corpus/manifest.json` pins `lap 0.5.13` and `numpy 2.5.2`.

**Pivot order and tie-breaking: an algorithmic artifact, not a
canonical rule.** On an all-zeros `n×n` matrix (every permutation
optimal) `lapjv` returns row→col `[1,0]`, `[2,0,1]`, `[3,0,1,2]`,
`[4,0,2,1,3]`, `[11,0,9,8,7,6,5,4,3,2,1,10]`,
`[13,0,11,10,9,8,7,6,5,4,3,2,1,12]` for n = 2,3,4,5,12,14
(`probe_lapjv.py`). It never returns the identity. No simple rule
("first index wins", "lexicographically smallest optimum") predicts
it; mirroring means transliterating the three C++ routines.

**Determinism: yes, given the same matrix.** 200 repeats on a 6×6 zero
matrix returned one assignment (`probe_lapjv.py`). The algorithm is
pure double arithmetic with no randomness. Cross-build determinism is
not proven and is a real port hazard: the C++ may be compiled with FMA
contraction or extended precision, which a Go transliteration cannot
reproduce. The corpus's `lap` pin is what makes the question
answerable at all.

**`extend_cost=True`, measured rather than read.** Beets always passes
it (`match.py:66`). Two findings:

- **Square inputs: the extension is behaviorally transparent.**
  `extend_cost=True` and `extend_cost=False` returned identical
  assignments on 5000+ tie-rich square trials (n 2–15, binary and
  zero-cost matrices; `probe_extend.py`, `probe_extend2.py`). So for a
  square matrix a Go port may solve the raw `n×n` problem. Both pinned
  corpus cases are square (12×12, 14×14), which materially lowers
  ruling F's risk.
- **Rectangular inputs: the model is `n = max(rows, cols)` padded with
  ZERO.** I tested six candidate extension models against `lapjv` on
  2634 random rectangular matrices (`probe_model2.py`):

  | Model | Agreement |
  | --- | --- |
  | `n = max(rows, cols)`, fill `0` | **2634 / 2634 (100 %)** |
  | `n = max(rows, cols)`, fill `max+1` | 92.0 % |
  | `n = rows + cols`, fill `max+1`, no zero block | 81.3 % |
  | `n = rows + cols`, fill `max+1`, zero bottom-right block | 70.3 % |
  | `n = rows + cols`, fill `0`, zero block | 6.0 % |

  The classic `(rows+cols, max+1)` recipe (the one
  `derive_ties.py:58-66`'s docstring assumes) **disagrees with this
  `lap` build on 30 % of rectangular ties.** All models agree on the
  optimal *cost*; they differ only on which tied optimum is chosen,
  which is exactly what tier 1 grades. The `max(rows, cols)`-zero-pad
  model was then confirmed against the real corpus: it matched
  `lapjv` on **all 214** candidate-bearing cases across all three
  profiles, 21 of them rectangular (`probe_all.py`).

  Caveat the plan should carry: this is behavioral equivalence
  established on one build, not source. The Go port should implement
  the zero-pad model **and** keep an oracle-tagged differential test
  that re-establishes it against the pinned `lap`, so a `lap` bump
  surfaces as a red rather than as silent divergence.

**Whole-corpus validation of the port's target.** `probe_all.py`
rebuilt every case's best-candidate cost matrix offline (release
bodies read from disk, normalized through
`MusicBrainzAPI._normalize_data`,
`beetsplug/_utils/musicbrainz.py:746-775`, then `album_info`,
`beetsplug/musicbrainz.py:631-695`, with the pseudo-release merge where
the status calls for it) and compared `lapjv`'s assignment against the
committed record: **214 compared, zero mismatches** (estate 177, stock
36, locale 1; one case carries no candidate). The corpus's assignments
are reproducible offline without the stub, which is a cheap regression
instrument the plan can reuse.

**One benign coincidence worth a line.** `derive_ties.py:35` sets
`FORBIDDEN_COST = 1e6`, which is exactly `lap.LARGE`. 900
forbid-and-re-solve probes all matched brute force, and sentinel values
of `1e6`, `1e9`, `1e12` gave identical costs (`probe_large.py`), so the
ledger is sound. It is still a coincidence that would be better written
as a value the derivation derives from the matrix.

### 11. The two pinned tied cases, probed offline

Both are reachable offline: the committed request logs name the
release lookups, and the response bodies plus the plugin's own parser
reconstruct the exact `TrackInfo` list beets scored. The full 12×12 and
14×14 matrices print from `probe_ties.py <case>`; the decisive cells
are below.

**`should-ask`** (`corpus/cases/should-ask/`, profile `estate`, 12
items × 12 tracks, `va = False`). Optimum
`1.7962962962962963`, matching `ties.json`'s `optimal_cost` exactly.

The tie is a single free transposition. Rows 1 and 6 agree bit-for-bit
on the two contested columns:

| | t1 | t6 |
| --- | --- | --- |
| **i1** | `0.90740740740740744` | `0.88888888888888895` |
| **i6** | `0.90740740740740744` | `0.88888888888888895` |

`lapjv` chose `{0:0, 1:6, 2:2, 3:3, 4:4, 5:5, 6:1, 7:7, 8:8, 9:9,
10:10, 11:11}`, **the swap.** The one equally optimal alternative is
the identity `{1:1, 6:6}` (`probe_orbit.py`). The committed
`Expected.Assignment` is the swap. **A Go solver that returns the
obvious identity fails this row**, which is the sharpest available
statement of the mirroring requirement.

**`should-reject`** (14 items × 14 tracks, `va = False`). Optimum
`5.282407407407407`, matching the ledger. `lapjv` chose `{0:0, 1:1,
2:2, 3:3, 4:4, 5:7, 6:6, 7:9, 8:8, 9:5, 10:10, 11:11, 12:12, 13:13}`,
identity everywhere except a 3-cycle on items 5, 7, 9. Forbidding each
chosen pair in turn yields five equal-cost alternatives, all
permutations within items `{1,3,5,7,9,11}` over tracks
`{1,3,5,7,9,11}` (`probe_orbit.py`):

```
forbid (1,1)   -> {3:1, 1:3}
forbid (3,3)   -> {11:1, 7:3, 1:9, 3:11}
forbid (9,5)   -> {7:1, 9:3, 11:5, 3:9, 1:11}
forbid (7,9)   -> {7:3, 3:9}
forbid (11,11) -> {11:5, 9:11}
```

No two item rows are bit-identical in this case; the ties come from
equal sums across permutations, which is a strictly harder shape for a
mirroring test than `should-ask`.

**A finding neither ledger records: 17 more tied candidate
assignments, none of them dispositioned.** `ties.json` asks the tie
question of `Candidates[0]` only (`internal/harness/oracle/ties.go:54`;
`corpus/README.md:2110-2112`), so 80 of the 294 committed candidate
assignments were never measured. Sweeping every candidate of every
multi-candidate case with the ledger's own forbid-and-re-solve test
(`probe_multi.py`) found **17 tied candidate assignments carrying no
tie disposition**, across six cases beyond the two pinned ones:

`real-noise-dylan-bootleg` (candidates 1, 2, 3, 4),
`real-noise-pink-floyd-de1994` (2, 4),
`real-nonlatin-girls-generation` (4),
`real-nonlatin-utada-first-love` (4),
`real-nonlatin-utada-first-love-locale` (4), plus the non-best
candidates of `should-ask` (1, 2, 3) and `should-reject` (1–5).

These are not harmless. Tier 1 grades every candidate's `Tracks`
vector elementwise. Taking one alternative optimum per case
(`probe_altopt.py`), five of the seventeen produce a **different**
graded vector, and two of those also change the candidate's own
`tracks` penalty sum:

| Candidate | Alternative changes `Tracks` vector | Changes `tracks` sum |
| --- | --- | --- |
| `real-nonlatin-utada-first-love` cand 4 | yes | **yes** |
| `real-nonlatin-utada-first-love-locale` cand 4 | yes | **yes** |
| `real-noise-dylan-bootleg` cand 3 | yes | no |
| `real-noise-dylan-bootleg` cand 4 | yes | no |
| `real-noise-pink-floyd-de1994` cand 4 | yes | no |
| the other twelve | no (on the alternative found) | no |

So a Go solver that mirrors `lapjv` on the two pinned cases and
diverges on `real-nonlatin-utada-first-love`'s fifth candidate fails
tier 1 with **no ledger row explaining why**. Ladder entry 2's "every
tied case is excluded or pinned" is satisfied in letter and not in
substance. Open question (see (d), item 2).

**Mirroring test design a Go port needs.** Four layers, cheapest
first.

1. **Unit, hand-authored, gate-resident.** Fix the two pinned matrices
   as `testdata` (they are 12×12 and 14×14 of `%.17g` doubles, ~4 KB)
   and assert the Go solver returns `lapjv`'s exact assignment. Firing
   proof: a solver returning the identity on `should-ask` fails. This
   is the only layer that runs in `scripts/check.sh` and it is the one
   that catches the regression that matters.
2. **Property, gate-resident.** Assert optimality independently of tie
   choice: the returned assignment's cost equals a brute-force optimum
   for small `n`, and `min(rows, cols)` pairs are always matched. This
   catches a wrong solver; it cannot catch a wrong tie-break.
3. **Differential, oracle-tagged.** A `//go:build oracle` test that
   generates tie-rich matrices (small integer costs, `n` 2–15, plus
   rectangular shapes), runs both the Go solver and pinned `lap`
   through a Python driver in `derive_ties.py`'s shape, and requires
   identical `x` and `y`. This is the only instrument that can
   establish the extension model and the pivot order, and it
   re-establishes them after a `lap` bump. It cannot gate, so its
   result belongs in the pass report as a measured number.
4. **Whole-corpus, gate-resident.** Tier 1's own assignment comparison
   over all 214 cases and 294 candidates, which is where the 17
   unledgered ties actually bite.

The escalation trigger ruling F names, two failed attempts, should be
evaluated against layer 1 alone. Layers 3 and 4 are the confidence, not
the gate.

### 12. Float ordering and Python-versus-Go divergence points, collected

| # | Site | Hazard |
| --- | --- | --- |
| 1 | `distance.py:171-176` vs `internal/engine/matcher/penalty.go` | `raw_distance`/`max_distance` sum `_penalties` in insertion order; float addition is not associative. 3d's contribution is that the `tracks` list must be fed in **track order** (§1), and `missing_tracks`/`unmatched_tracks` one `1.0` per excess unit in loop order (`distance.py:533-538`). |
| 2 | `match.py:64` | Cost cells are `float(Distance)`, normalized per pair. Any Go shortcut that sums raw penalties instead is a different matrix and therefore possibly a different tie choice. |
| 3 | `lapjv`'s C++ | Tie-break depends on double comparisons inside reduction and augmentation. FMA contraction or x87 extended precision in either language changes them. Go's spec forbids silent fusing without explicit conversion, which helps, but the C++ side is not under our control. |
| 4 | `probe_model2.py` | The rectangular extension changes the tied optimum. `max(rows, cols)` + zero fill is the measured model; `(rows+cols)` + `max+1` is not. |
| 5 | `match.py:77-81` | Extras' pre-sort order is address-derived and outside `PYTHONHASHSEED`'s reach. Total sort keys are what makes the record reproducible; `extra_items`' totality is empirical. |
| 6 | `match.py:285, 288` | Strict `<` for strong, `<=` for medium. Identical in Go given identical doubles; the proximity flag exists because 1e-12 tolerance is inadmissible here. |
| 7 | `match.py:296` | `results[1].distance - min_dist` is a plain double subtraction compared `>=` against `rec_gap_thresh`. Bit-identity of both distances is a precondition of the comparison agreeing. |
| 8 | `id_extractors.py:20` | `\w` Unicode-aware in Python `re`, ASCII-only in Go RE2. Carried from 3c. |
| 9 | `tunables.go:365` | `\b` in `SDPatternsWordBoundary`, same regime split. Carried from 3b. |
| 10 | `match.py:354-359` | `is None` has no Go analogue on plain-typed fields (§3). |
| 11 | `derive_oracle.py:73-103` | The record's floats are `%.17g`, which round-trips a double exactly. Go's `strconv.ParseFloat` on those strings is exact, so decoding is not a divergence source. |

### 13. New declared corpus-coverage gaps this section identifies for 3d

Rung 3 carries ten declared gaps into 3d (3a three, 3b six, 3c one).
These are additions, to be recorded on top of that running total
rather than re-counted.

1. `match_by_id`'s no-id branch (0 of 215 cases).
2. `match_by_id`'s no-consensus branch (0 of 215).
3. The explicit `search_ids` branch (0 of 215).
4. The `import.timid` gate on the early return (0 of 215).
5. `strong` reached after a text search (0 of 215; no-search ⟺
   strong, exactly).
6. `_recommendation`'s per-track key contribution to the `max_rec`
   downgrade set: the loop runs but never changes an outcome, proven
   by a model that omits it reproducing all 214 recommendations.
7. `import.quiet = False`, and with it the whole `timid` /
   `none_rec_action` / "ask" half of `_summary_judgment` (0 of 215; the
   shim forces quiet).
8. `quiet_fallback: asis`, and the `ASIS` action entirely (0 of 215).
9. Every non-default value of the four recommendation knobs (all 215
   cases echo `[0.04, 0.25, 0.25]` and the default `max_rec`).
10. The assignment objective's `inclArtist=false` on a VA release: no
    tied or assignment-sensitive VA case exists, so passing `info.VA`
    into the cost matrix would pass every committed case.
11. `match.ignored` reading the nonzero view rather than the added-key
    view (unreachable, inherited from the M8 trio but now with a named
    second site).

Each needs a hand-authored fixture in its owning package's `testdata`,
with the mutation that proves it fires, per the spec's assertion
discipline. (See (d), item on gap-count reconciliation: sections (a),
(b), and (c) disagree on how many of these count as genuinely new.)

### 14. What 3d inherits, concretely

- `matcher.Tunables` already carries every switch except
  `search_ids`: `MaxRec` (`tunables.go:67`), `Ignored` (`:85`),
  `Required` (`:92`), `Timid` (`:168`), `Quiet` (`:186`),
  `QuietFallback` (`:191`), `NoneRecAction` (`:196`), and the three
  tables (`:348, 365, 388`).
- `matcher.DistanceForAlbum(items, info, pairs, tun, evalDate)`
  (`score.go:60`) takes the pairing as input and computes no
  assignment (`album_tracks.go:5-14` says so explicitly). `evalDate`
  is a parameter, never a clock read, and
  `TestEngineNeverReadsTheClock` enforces it across `internal/engine`.
- `musicbrainz.AlbumSearchFor(items, tun)` (`query.go:79`) derives its
  own likelies, consensus, and `vaLikely` (`query.go:108`, unexported;
  3d exports it the day the no-consensus branch needs it, and per §4
  that branch is corpus-unreachable, so the export may never be
  owed).
- `musicbrainz.Client.AlbumForID`, `Release`, `SearchReleases`
  (`client.go:296, 215, 255`), the stub, and the zero-miss rule are
  3c's, unchanged. R6's blind side stands: nothing reports which
  logged `Request` entries a run consumed, so 3d needs the hit set to
  prove its request *set* (as against 3c's request *construction*).
- `analysis.CheckOracleBugs` (`oraclebug.go:55`) and
  `analysis.TieEntry` (`ties.go:56-66`) are the two ledger surfaces 3d
  edits.

---

## (b) The harness ledgers and tiers as landed

Every figure below is read off the tree at `c7a85fd` with `jq` or
`ls`, never from a document's paraphrase. No `responses/*.json` body
was read.

### 1. `internal/harness/conform` as landed: four tiers' worth of instrument, two of them not yet built

The package is five files, 1,708 lines with tests, born at 3a and
extended at 3b and 3c.

| File | Lines | What it is | Pass |
| --- | --- | --- | --- |
| `doc.go` | 34 | The charter and the harness-side-of-the-wall argument (`doc.go:6-15`), plus 3c's amendment that driving a stub is not writing a harness package (`doc.go:24-33`) | 3a, 3c |
| `component.go` | 911 | Tier 0: the typed `components.json` decoder and `ReplayTable`/`ReplayRow` | 3a |
| `defaults.go` | 76 | `CheckCaseWeights` (`:21`), `CheckCaseThresholds` (`:43`), `CheckUnidecodeVersion` (`:67`) | 3a |
| `translate.go` | 443 | The translation tier: `CheckTranslatedAlbum` (`:128`), `gradeLedger` (`:329`), `mergeProfile` (`:193`), `checkProfileKeyDrift` (`:268`), `checkNoReleaseReachesBrowseThreshold` (`:396`) | 3b |
| `request.go` | 238 | The request tier: `GradeRequests` (`:72`), `gradeCaseRequests` (`:114`), the three floors (`:26-30`) | 3c |
| `arch_test.go` | 206 | `checkNoHarnessImports` (`:30`), `checkNoTimeNow` (`:75`), both with `testdata` subtree exclusions (`:37-39`, `:82-84`) | 3a, 3b |

**Tier 0.** `ReplayRow` (`component.go:101`) dispatches on
`row.Method` across eleven methods and, inside method `add`, on the
argument shape, for twelve shapes total (`component.go:96-99`).
`componentMethods` (`component.go:24-35`) is a deliberate second copy
of `analysis.componentMethods` (`analysis/component.go:23`),
maintained separately because the analysis copy sits behind the
territory wall; the two disagreeing is a gate failure by design
(`component.go:16-23`). `ReplaySummary.CountByMethod` excludes any
row that failed, which is what makes "replayed count equals table
row count" catch a silently skipped row
(`component.go:60-73, 75-93`). A near miss (matching only within the
1e-12 fallback, never bit-identical) is a reported finding rather
than a failure (`component.go:48-56`).

**The translation tier.** `gradeLedger` rebuilds all 50
`album_info.json` rows through `musicbrainz.TranslateAlbum` and
returns a compared count so a silently skipped row fails
(`translate.go:320-329`). The twenty-one-row `fieldNullability`
policy sits at `translate.go:70-96` with its three regimes;
`TestNullabilityPolicyCoversEveryField` (`translate_test.go:20`)
asserts the row count against `reflect` so a new vocabulary field
cannot slip past.

**The request tier.** `GradeRequests` opens one `stub.Server` per
case, issues one `client.Release` per distinct logged mbid and one
`client.SearchReleases` per case that logs a search, joins every
`srv.Misses()` entry into the case's error (`request.go:185-187`),
and collects `srv.Consumed()` (`request.go:188`). Floors:
`RequestCaseFloor = 215`, `RequestLookupFloor = 303`,
`RequestSearchFloor = 48` (`request.go:26-30`), deliberately floors
rather than exact counts because lane E1 adds cases.

**The floors and the mutation-proof convention.** Floors live in one
place and are read from there, never re-typed:
`analysis.ComponentRowFloor = 72` (`analysis/component.go:15`) is
reused by `conform`'s own below-floor proof
(`component_test.go:273`), `analysis.AlbumInfoLedgerRowFloor = 50`
(`analysis/album_info.go:20`), `analysis.FileTagLedgerRowFloor = 12`
(`analysis/file_tags.go:17`), and the three request floors above. The
convention every checker follows: a real-corpus test that grades
cleanly, plus a hand-authored fixture in the package's own `testdata`
whose mutation proves the check fires. `conform/testdata` carries
eight such fixtures (`unknown_method.json`, `ulp_mismatch.json`,
`method_missing.json`, `add_both_key_and_weights.json`,
`add_neither_key_nor_weights.json`, `missing_regex_key.json`,
`wrong_type_track_info.json`, plus the `arch/`, `request/miss-fixture/`
and `translate/skipped-row-ledger.json` subtrees).

**What a tier 1 and a tier 2 would slot into.** Nothing in `conform`
today grades a verdict, an assignment, a candidate list, or a
distance against a `cassette.Expected`. `request.go` is the only file
that drives a stub, and it grades request construction only, by its
own doc comment (`request.go:64-71`). Tier 1 and tier 2 are new files
in this package with no partial implementation to extend. What they
can reuse: `mergeProfile` (unexported, `translate.go:193`),
`stub.New`/`Misses`/`Consumed`, `cassette.LoadCase`, the repo-root
walk (three private copies today: `conform/component_test.go:25`,
`harness/corpus_test.go:24`, and oracle's), and the
near-miss/1e-12 comparison shape at `component.go:828`.

### 2. `internal/harness/analysis`: every ledger type, checker and firing fixture

| Type | File:line | Checker | Firing fixture |
| --- | --- | --- | --- |
| `Proximity` / `ProximityEntry` / `ProximityLedger` | `proximity.go:36, 52, 44` | `Flag` (`:65`), `CheckProximityLedger` (`:111`) | `testdata/records/{on-strong-threshold,on-gap-threshold,clear,zero-candidates,no-thresholds}.json` |
| `TieEntry` / `TieLedger` | `ties.go:56, 39` | `CheckTieLedger` (`:83`) → `checkTieEntry` (`:114`) → `checkSkippedTieEntry` (`:148`) | fourteen inline table rows, `ties_test.go:8-140` |
| `ComponentTable` / `ComponentRow` | `component.go:46, 56` | `CheckComponentTable` (`:74`) | `component_test.go:26` |
| `Claims` | `claims.go:42` | `CheckClaims` (`:86`) | `testdata/claims/mutated-{assignment,penalty,album-id,applied-field}`, `mutated-penalty-absent`, `unknown-key` |
| `AdjudicationEntry` / `AdjudicationSubject` / `DefectFree` / `AdjudicationLedger` | `adjudication.go:46, 35, 71, 104` | `CheckAdjudications` (`:121`) → `checkAdjudicationEntry` (`:141`); `CheckDefectFree` (`:218`) | nineteen inline rows (`adjudication_test.go:10`), five `defect_free`-block rows (`:115`), and the `testdata/defect_free/{oracle_bugs,adjudications}.json` pair (`:221`) |
| `OracleBugEntry` / `OracleBugLedger` | `oraclebug.go:21, 37` | `CheckOracleBugs` (`:55`) → `checkOracleBugEntry` (`:75`) | sixteen inline rows, `oraclebug_test.go:10-110` |
| `AlbumInfoLedger` / `AlbumInfoLedgerRow` / `LedgerAlbumInfo` / `LedgerTrackInfo` | `album_info.go:86, 96, 109, 128` | `CheckAlbumInfoLedger` (`:160`) | `testdata/album_info/{good,null-and-empty}` |
| `FileTagLedger` / `FileTagLedgerRow` | `file_tags.go:64, 76` | `CheckFileTagLedger` (`:100`) | `testdata/file_tags/good` |

**The `defect_free` block.** `AdjudicationEntry.DefectFree
*DefectFree` (`adjudication.go:66`), a pointer because absence is
meaningful. `DefectFree` carries `Distance`, `Recommendation`,
`Action`, `PatternTable` (`adjudication.go:71-76`). Two checkers
guard it. `checkAdjudicationEntry` enforces the within-row rules
(`adjudication.go:188-204`): a block only on an `oracle-wrong` row,
non-empty recommendation, action and pattern table, and a distance
inside `[0, 1]`. `CheckDefectFree` (`adjudication.go:218`) enforces
the cross-ledger rules: every probe case of a `StringHeuristicHypotheses`
hypothesis at `oracle-wrong` resolves to a row carrying a block, the
block's pattern table is one of `LegalPatternTables`
(`adjudication.go:96`), and the recommendation and action appear
verbatim in the row's `corrected_verdict`. It also forbids the
converse direction (`adjudication.go:265-272`).

**The `AgentAdjudicator` sentinel.** One constant
(`adjudication.go:30`), two enforcement sites, both spelled
identically: `adjudication.go:181` for adjudication rows and
`oraclebug.go:125` for oracle-bug rows. Neither touches `ties.json`.

**`CheckOracleBugs` and its `adjudicationIDs` map.** The signature is
`CheckOracleBugs(caseIDs []string, adjudicationIDs map[string]bool, l
*OracleBugLedger) error` (`oraclebug.go:55`). The map is keyed by
adjudication **row id**, and its only use is resolving
`AdjudicationRefs` (`oraclebug.go:89-93`). There is no subject index
anywhere in the exported surface. The closest thing that exists is
`CheckDefectFree`'s own local, unexported `rows map[caseID]AdjudicationEntry`
(`adjudication.go:224-229`), which is exactly the index B3 asks for,
built inside one function and thrown away.

**The subject index.** `AdjudicationSubject` (`adjudication.go:35`)
is `{case_id} | {album, release_mbid}`, exactly one shape populated,
enforced at `adjudication.go:144-157`. All 21 committed rows carry a
`case_id` subject; zero carry the shadow shape.

### 3. `internal/harness/stub`, `cassette`, `oracle`'s tie driver, `corpus_test.go`

**Stub.** `Server` holds two maps, `responses` and `extra
map[key]extraEntry` (`stub.go:66-67`), and two slices, `misses` and
`consumed []cassette.Request` (`stub.go:70-71`). `ServeHTTP` consults
the request map first, the extra map second, and records a `Miss`
plus a 404 otherwise (`stub.go:122-137`). `Misses()` (`stub.go:158`)
and `Consumed()` (`stub.go:168`) both return clones. The
derivation-inertness rule is on the package doc comment
(`stub.go:6-14`).

**The blind side, precisely.** There is no accessor for which
entries of the request map were *hit*, and none for which were left
unused. A client that issues fewer requests than the case's log
records leaves zero misses and passes. 3c covered this inside its
own sweep with a per-case issued-count assertion against
`countLookupRequests`/`countSearchRequests` (`request.go:163-183`),
which works only because 3c decides which requests to issue by
reading the log. 3d's runner decides the request set from the
matcher's own flow, so that assertion has no counterpart, and the
hit set is the thing that replaces it. 3c deliberately kept
`Unused()` out (HISTORY 3c, "wrong to rediscover" 13c).

**Cassette.** `Case` (`cassette.go:41-56`) carries `ID`, `Input`,
`Profile`, `EvalDate`, `Requests []Request`, `Extra []Request` and
`Expected *Expected`. `Expected` (`cassette.go:76-88`) is the whole
tier 1 and tier 2 expectation surface: `Recommendation`, `Action`,
`EchoedDate`, `ConfigHash`, `Weights`, `Thresholds`, `Candidates
[]Candidate`, `Assignment map[int]int`, `ExtraItems []int`,
`ExtraTracks []int`, `AppliedFields`. `Candidate` (`cassette.go:103-109`)
carries `Source`, `AlbumID`, `Distance`, `Penalties
map[string]float64`, `Tracks []float64`. `Thresholds`
(`cassette.go:95-100`) carries the three floats plus `MaxRec
map[string]string`. There is no disposition field on a `Case`: every
disposition lives in a ledger and joins by case id.

**The profile decoder.** `Case.Profile` is a bare name. The decoder
is `conform.mergeProfile` (`translate.go:193-204`), unexported, a
name-to-overlay table over three names, failing on any fourth.
`locale` sets `Languages` and `IgnoredAliasTypes`; `stock` and
`estate` are no-ops. `checkProfileKeyDrift` (`translate.go:268`)
scans the three committed YAML files line by line against a ten-key
allowlist (`translate.go:213-224`) so a profile growing an
unmodelled key reds the gate.

**The tie driver.** `DeriveTies` (`oracle/ties.go:43`, `//go:build
oracle`) starts a stub, renders the profile, and hands
`derive_ties.py` the best candidate. It returns a `TieSkipped` row
without starting the stub when the record carries no candidate
(`ties.go:47-53`). `derive_ties.py`'s `measure` emits six fields and
**never emits `disposition` or `reason`** (`derive_ties.py:118-126`);
those are hand-written and carried forward by `writeTieLedger`
(`ties_derive_test.go:139-141`) and blanked before the struct
comparison in the diff test (`ties_derive_test.go:91-92`).

**`corpus_test.go`.** Eleven gate-resident tests over the whole
corpus, including `TestTieLedgerCoversEveryCorpusCase` (`:276`),
`TestAdjudicationsLedgerIsWellFormed` (`:366`),
`TestOracleBugsLedgerIsWellFormed` (`:386`, the sole production
caller of `CheckOracleBugs`, building `adjIDs` at `:397-400`), and
`TestDefectFreeBlocksAreWellFormed` (`:419`).
`wantManifestSchemaVersion = "4"` sits at `:163`.

### 4. The corpus figures, off the tree

**Case count: 215** (`ls corpus/cases | wc -l`). Profiles: 178
`estate`, 36 `stock`, 1 `locale`.

**Ledger row counts.** `ties.json` 215, `proximity.json` 2,
`adjudications.json` 21, `oracle_bugs.json` 6, `components.json` 72,
`album_info.json` 50, `file_tags.json` 12.

**Dispositions.**

| Ledger | Distribution |
| --- | --- |
| `ties.json` | 212 unique (no disposition), **2 pinned**, **0 excluded**, 1 skipped |
| `proximity.json` | 1 flagged (`rec-strong-medium-boundary`, boundary `strong_rec_thresh`), 1 skipped (`admit-zero-track-dropped`); the other 213 are clear and therefore unlisted |
| `adjudications.json` | 18 `oracle-wrong`, 3 `oracle-correct`, **0 pending**, 0 `no-correct-answer`; all 21 adjudicated by `Geoff` |
| `oracle_bugs.json` | 4 `oracle-wrong`, 2 `oracle-correct`, **0 pending**; all six adjudicated by `Geoff` |

**Tie rows in detail.** No `adjudicator` field exists in `ties.json`
at all (`grep -c adjudicator` returns 0). The two pinned rows are
`should-ask` (optimal 1.7962962962962963, difference 0) and
`should-reject` (optimal 5.282407407407407, difference 0), each
carrying a long written reason naming the corrupted items and
asserting that both optima produce the same album distance and the
same recommendation. The skipped row is `admit-zero-track-dropped`.

**Adjudication rows by subject and status.** All 21 subjects are
corpus case ids. Ten `oracle-wrong` rows carry a `defect_free` block
(six `adv-ft-*` under `SDPatternsWordBoundary`, four `adv-remix-*`
under `SDPatternsNoBracketDiscount`, every one at `medium`/`skip`).
Six `adv-assign-*` rows are `oracle-wrong` with prose-only corrected
verdicts naming an identity assignment and a would-be distance. Two
alias rows (`adj-adv-alias-girls-generation-romanized`,
`adj-adv-alias-vissi-native-greek`) are `oracle-wrong` with
corrected verdicts naming applied `artist`/`albumartist` strings.
Three rows are `oracle-correct` with an empty `corrected_verdict`.

**Probe cases with no adjudication row, resolved by
`subject.case_id`.** 23 probe-case mentions across the six
hypotheses, 23 distinct cases, no case probed by two hypotheses.
**Two** have no adjudication row today, and both sit under
hypotheses ruled `oracle-correct`:

- `rec-strong-medium-boundary` under `bug-threshold-boundary`
  (`oracle-correct`)
- `real-nonlatin-utada-first-love` under `bug-locale-alias-ignored`
  (`oracle-correct`)

The spec's revision-2 figure of three
(`real-nonlatin-alla-pugacheva-vol1`, `real-nonlatin-x-japan-art-of-life`,
`real-nonlatin-utada-first-love`) is superseded: the first two now
carry `oracle-correct` rows written in the 2026-09-04 batch. **B3's
rule therefore fires on nothing in the committed corpus today**, and
its firing fixture must be hand-authored.

**Oracle-bug hypotheses by status.**

| Hypothesis | Disposition | Reproduced | Probes | Refs |
| --- | --- | --- | --- | --- |
| `bug-ft-pattern` | oracle-wrong | yes | 6 | 6 |
| `bug-remix-parenthetical` | oracle-wrong | yes | 4 | 4 |
| `bug-nonlatin-alias-selection` | oracle-wrong | yes | 4 | 4 |
| `bug-assignment-rejection-gap` | oracle-wrong | yes | 6 | 6 |
| `bug-threshold-boundary` | oracle-correct | no | 1 | 0 |
| `bug-locale-alias-ignored` | oracle-correct | no | 2 | 1 |

Two of `bug-nonlatin-alias-selection`'s four refs point at rows whose
own disposition is `oracle-correct`, which is the precedence rule's
second case exactly (spec :302).

**The `defect_free` block's ten rows.** Distances
0.10185185185185186, 0.11764705882352941, 0.10784313725490197,
0.08666666666666667, 0.09170751633986929, 0.0909090909090909 (the six
`adv-ft-*`, `SDPatternsWordBoundary`) and 0.08,
0.07142857142857142, 0.0925925925925926, 0.08187134502923976 (the
four `adv-remix-*`, `SDPatternsNoBracketDiscount`). All ten read
`medium` / `skip`.

**Tier 1 population: all 215 cases.** Every one carries an
`Expected` record, enforced at `corpus_test.go:92-94`. Within that:
214 cases carry at least one candidate and 1 carries none
(`admit-zero-track-dropped`); 294 candidates total (193 cases with
1, 4 with 2, 9 with 5, 8 with 6, so 21 cases exercise candidate
ordering and the gap threshold); 214 assignments graded exactly (212
unique plus 2 pinned) and 1 skipped, with **zero cases excluded from
assignment grading**; 213 distances compared at 1e-12 and 1 required
bit-identical; 351 requests (303 lookups, 48 searches) resolved with
zero misses; 7 cases carry non-empty `extra_items` (max 33), 14
carry non-empty `extra_tracks` (max 4), no case carries both, and 23
cases have a non-identity assignment.

**Tier 2 population: all 215 verdicts, 18 of them against a
non-standard expectation.** 197 cases grade against their own
committed `recommendation`/`action`. The 18 are the probe cases of
the four `oracle-wrong` hypotheses less the two whose own rows read
`oracle-correct`: 10 with a machine-readable `defect_free` block, 6
`adv-assign-*` with prose-only corrected verdicts, 2 alias rows whose
correction lives entirely in applied fields rung 3 does not compute.
Zero pending rows and zero `no-correct-answer` rows exist, so
neither of those two tier-2 paths has a live corpus instance.

**Verdict distribution, for sizing the verdict layer.** 167
strong/apply, 39 medium/skip, 4 low/skip, 5 none/skip. Every `none`
and every `low` case is `estate`.

**The three committed profiles.** `stock.yaml` sets `plugins:
musicbrainz` and nothing else (5 lines). `estate.yaml` adds two
substituted box paths, `import.{quiet_fallback: skip,
none_rec_action: skip, timid: no}`, and `original_date: yes` /
`per_disc_numbering: yes`. `locale.yaml` is estate plus
`import.languages: [ja]` and `import.ignored_alias_types: []`. None
of the three sets any `match:` or `musicbrainz:` key, which is why
all 215 cases echo the same twenty weights and the same thresholds.

### 5. Surfaces no committed profile can reach (M8), as measured

The spec names three. All three hold, and the tree adds four more
the plan should carry.

1. **`match.required`**: absent from all three profiles;
   `Tunables.Required` (`tunables.go:87-92`) has no consuming
   component in the tree yet.
2. **`match.ignored`**: same; `Tunables.Ignored` (`tunables.go:79-85`).
3. **`preferred.original_year`**: absent from all three; already
   graded by 3a's hand-authored `album_year_preferred_original.json`.
4. **`max_rec`'s downgrade against a zero-valued key.** All 215 cases
   echo the same `max_rec` (`{missing_tracks, unmatched_tracks}`).
   Across all 294 committed candidates, **no candidate carries either
   key at 0.0** and 87 carry one nonzero. The `_penalties`-versus-`keys()`
   distinction the spec's hazard 4 names (spec :208) is
   therefore invisible to the corpus for the two keys that actually
   drive a downgrade, even though 272 of 294 candidates carry *some*
   key at 0.0. A port that downgrades against the wrong view passes
   tier 1 and tier 2 both.
5. **`import.none_rec_action` and `import.timid`.**
   `derive_oracle.py`'s `quiet_action` forces
   `config["import"]["quiet"] = True` before calling beets' own
   `_summary_judgment` (`derive_oracle.py:204-206`), and that function
   returns on the quiet branch before ever reading `timid` or
   `none_rec_action` (`beets/ui/commands/import_/session.py:337-354`).
   So every committed `action` is the quiet-mode action:
   strong→apply, everything else→`quiet_fallback`, which is `skip` in
   all three profiles. This is documented at `corpus/README.md:1394-1399`
   and appears in no pass HISTORY. A faithful Go
   `_summary_judgment` reading `Tunables.Quiet` at its beets default
   of `false` would return "ask" for all 48 non-strong cases and
   diverge from every one of them.
6. **`quiet_fallback: asis`**: no profile sets it, so one of the two
   quiet-fallback branches is unreachable.
7. **`profile_hashes` covers one profile of three.**
   `corpus/manifest.json`'s `profile_hashes` carries `locale` alone.
   Tier 1 sources its non-echoed configuration from the profile, and
   an edit to `stock.yaml` or `estate.yaml` has no hash to fail
   against. `checkProfileKeyDrift`'s allowlist catches a new key and
   misses a changed value.

Rung 3's running total is ten declared gaps (3a's three, 3b's six,
3c's one). Items 4 through 7 above are 3d's own candidates on top of
that total, per the spec's instruction not to re-count. (See (d) for
the reconciliation of this count against sections (a) and (c).)

### 6. M4: `TieEntry.Adjudicator`, every reader and writer

**What the field touches.** Six sites, in two packages plus the
corpus file.

| Site | Change | Why |
| --- | --- | --- |
| `analysis/ties.go:56-66` | Add `Adjudicator string \`json:"adjudicator,omitempty"\`` with the asymmetry recorded in its own doc comment | The spec requires the reason to live on the field (spec :133) |
| `analysis/ties.go:114-142` (`checkTieEntry`) | Require `Adjudicator` whenever `Disposition` is set; require a non-`AgentAdjudicator` value when `Disposition == TieExcluded` only | The sentinel rule from `checkAdjudicationEntry:181` |
| `analysis/ties.go:148-158` (`checkSkippedTieEntry`) | Forbid `Adjudicator` on a skipped row, beside the existing `AlbumID` and `Disposition` prohibitions | A skipped row carries no disposition, so it can carry no adjudicator |
| `analysis/ties_test.go:8` | Two new table rows minimum: an `excluded` row carrying `agent` (must fail), an `excluded` row carrying a human name (must pass), plus a `pinned` row carrying `agent` (must pass, which is the asymmetry's own proof) | The convention every checker follows |
| `oracle/ties_derive_test.go:91` | `want.Disposition, want.Reason = got.Disposition, got.Reason` must gain `want.Adjudicator = got.Adjudicator`, or the struct comparison at `:92` reds `make derive` on all 215 rows | `derive_ties.py` emits no adjudicator, so a re-measured row's field is always empty |
| `oracle/ties_derive_test.go:139-141` (`writeTieLedger`) | Carry `Adjudicator` forward from the prior ledger beside `Disposition` and `Reason` | Otherwise `-update-ledgers` silently erases the field |
| `corpus/analysis/ties.json` | Hand-add `"adjudicator"` to the two pinned rows | With zero excluded rows, `agent` on both is the only value that keeps the gate green, which is exactly the outcome the asymmetry was chosen to produce |

**The pinned asymmetry, measured.** Zero excluded rows exist, so the
sentinel half of the check has **no corpus instance and cannot fire
against the committed tree**; its proof is a hand-authored table row
only. Both pinned rows were agent-authored, which is the fact the
spec's adaptation rests on (spec :133), and the tree confirms it:
neither row carries an adjudicator name of any kind today.

**Nothing else reads `TieEntry`.** The full reader/writer set is
`analysis/ties.go`, `analysis/ties_test.go`, `oracle/ties.go`,
`oracle/ties_derive_test.go`, and `harness/corpus_test.go:283-289`.
`conform` does not import the tie ledger at all today, which is the
import 3d's tier 1 adds.

### 7. B3: `CheckOracleBugs`'s probe-case rule and the signature change

**The rule.** Every probe case of a hypothesis at `oracle-wrong` must
resolve to an adjudication row by `subject.case_id`; a probe with no
row is a gate failure naming the missing row.

**The signature change.**
`CheckOracleBugs(caseIDs []string, adjudicationIDs map[string]bool, l
*OracleBugLedger)` (`oraclebug.go:55`) has to gain the subject index.
The `adjudicationIDs` map is keyed by row id and is used only at
`oraclebug.go:90`; the new rule needs a `map[caseID]AdjudicationEntry`
or the ledger itself. Passing `*AdjudicationLedger` in place of the
map collapses both needs into one parameter and lets the checker
build both indexes itself, at the cost of one more argument shape
change.

**Readers and writers of the signature.** Exactly two call sites:
`harness/corpus_test.go:406` (which builds `adjIDs` by hand at
`:397-400`, and would stop needing to) and
`analysis/oraclebug_test.go:97`. The sixteen-row table at
`oraclebug_test.go:70-90` passes `adjudicationIDs` per row and would
need its literal updated in every row.

**The precedent already in the package.** `CheckDefectFree`
(`adjudication.go:218-263`) already implements the
probe-walk-and-resolve-by-subject shape, restricted to
`StringHeuristicHypotheses`. B3's rule is the same walk over *every*
hypothesis at `oracle-wrong`. That raises a placement question the
plan should settle rather than discover: the same rule can live in
`CheckOracleBugs` (where the spec puts it) or as a widening of
`CheckDefectFree`'s first loop, and the two differ in which ledger's
error prefix a failure carries.

**The firing fixture.**
`testdata/defect_free/{oracle_bugs,adjudications}.json` is the right
shape and cannot serve directly: all four of its probe cases
resolve, including `assign-fixture-d` under the `oracle-wrong`
`bug-assignment-rejection-gap`. A new fixture pair, or an inline
table row in the sibling style of `TestCheckOracleBugs`, is owed.

**What the rule catches today: nothing.** Both unresolved probe
cases sit under `oracle-correct` hypotheses. The rule's whole value
is prospective, against the next hypothesis Geoff rules
`oracle-wrong`.

### 8. What a whole-corpus runner needs from the harness that does not exist yet

1. **A stub hit set.** No accessor reports which logged `Request`
   entries a run consumed. 3c's issued-count assertion does not
   transfer, because 3d decides the request set from the matcher's
   flow. Without it, a matcher that issues fewer requests than the
   log records passes tier 1's zero-miss rule silently (3c HISTORY
   item 4).
2. **An exported adjudication subject index.** `CheckDefectFree`'s
   local `rows` map (`adjudication.go:224-229`) is the only one,
   unexported and function-scoped. Both B3's rule and tier 2's
   expectation lookup need it.
3. **A per-case grading policy resolver.** Four ledgers answer four
   separate questions about one case, each with its own loader and
   its own key. Nothing joins `ties.json`'s disposition,
   `proximity.json`'s flag, `adjudications.json`'s subject row and
   `oracle_bugs.json`'s hypothesis into "how is this case graded".
   Tier 1 and tier 2 both need that join, and the spec's disposition
   table (spec :285-299) is its specification.
4. **A writer for a pending adjudication row.** `adjudications.json`
   has a loader and no writer anywhere in the module.
   `CheckAdjudications` requires `Suspected`, `ReleaseTruth`,
   `Reason`, `Adjudicator` and `Date` all non-empty on a pending row
   (`adjudication.go:158-172`), so a runner-authored row needs five
   populated fields plus a `subject.case_id` that resolves. The plan
   has to say whether the runner writes the row or reports what a
   human must write.
5. **The `TieEntry.Adjudicator` field itself** (section 6), which
   the excluded-count rule at spec :292 depends on.
6. **A shared repo-root helper.** Three private copies exist
   (`conform/component_test.go:25`, `harness/corpus_test.go:24`, and
   `oracle`). A fourth is the cheapest option and the least
   defensible.
7. **A verdict vocabulary in the engine.** `matcher` has no
   recommendation type, no `max_rec` downgrade, no verdict-to-action,
   and no public entry point. `DistanceForAlbum` (`score.go:60`)
   takes a `pairs []AlbumPair` the caller must already have solved,
   so the assignment solve is the missing input to everything
   downstream.
8. **A tier-1 configuration path that uses `mergeProfile`.** The M1
   cross-check as landed compares the echo against
   `matcher.DefaultTunables()` directly (`defaults_test.go:43,
   55-61`), never against the merged result the spec's M1 fold
   specifies (spec :279). The two are equal today because no profile
   sets a `match:` key, so moving the comparison onto `mergeProfile`
   is behavior-preserving and closes the drift before a fourth
   profile makes it real.
9. **Wall-time headroom.** Today `go test ./...` runs 15.08s, with
   `internal/harness` at 11.26s and `internal/harness/conform` at
   11.23s. Inside conform, the request tier is 3.34s over 215
   stub-per-case replays and the browse-fallback scan is 2.2 to
   3.4s. Tier 1 adds a translate, a score and an assignment solve per
   candidate across 294 candidates on top of the same stub-per-case
   shape. The spec's escape hatch, promoting the whole-corpus run to
   its own `scripts/check.sh` leg (spec :267), is the thing to size
   against rather than assume away.

### 9. Two documentation drifts worth folding at 3d's close

- `corpus/README.md:2091-2094` says "No case in the smoke corpus is
  flagged today, so the ledger is empty". `proximity.json` carries
  two rows: `rec-strong-medium-boundary` flagged on
  `strong_rec_thresh` and `admit-zero-track-dropped` skipped.
- The rung 3 spec's B3 paragraph (spec :300) names three unresolved
  probe cases under still-pending hypotheses. As of the 2026-09-04
  batch there are two, both under `oracle-correct` hypotheses, and no
  hypothesis is pending. The spec anticipates this ("The count is
  read off the tree at execution time"), so this is a plan-header
  restatement rather than a spec edit.

---

## (c) The engine as landed and the 3d Go surface

### 1. What 3d inherits, with citations

**`internal/engine/matcher` (21 components, the accumulator, the
vocabulary).**

| Surface 3d consumes | Site | What 3d does with it |
| --- | --- | --- |
| `Tunables` (29 fields) | `tunables.go:42-225` | Reads; adds no field except the compat seam below |
| `Required`, `Ignored` | `tunables.go:80-92` | The admission filters 3a deferred; both corpus-unreachable |
| `IgnoredMedia` | `tunables.go:94-96` | Read by the translation layer already; 3d reads none of it |
| `IgnoreDataTracks`/`IgnoreVideoTracks` | `tunables.go:98-104` | Translation-side, landed at 3b |
| `StrongRecThresh`/`MediumRecThresh`/`RecGapThresh` | `tunables.go:43-59` | `_recommendation`'s three thresholds, with the strict `<` / `<=` asymmetry stated on the fields |
| `MaxRec map[string]string` | `tunables.go:61-67` | Level names, so 3d needs a name-to-level parser |
| `Quiet`/`QuietFallback`/`NoneRecAction`/`Timid` | `tunables.go:165-196` | The verdict-to-action trio plus timid |
| `Preferred.Media`/`Preferred.Countries` | `tunables.go:229-235` | The compile panics at `album_media.go:41` and `album_country.go:36` |
| `StringDistPatterns` plus the three tables | `tunables.go:217-224`, `:348-393` | The compat seam: a table is selected by assignment, never by editing `StringDist` |
| `Distance` two views | `penalty.go:124-181` | `AddedKeys`/`PerKey` grade; `NonzeroItems`/`NonzeroKeys` downgrade (port hazard 4) |
| `DistanceForAlbum` | `score.go:60` | Takes `pairs []AlbumPair` and `evalDate`; 3d supplies both |
| `DistanceForTrack(tun, item, info, inclArtist)` | `track.go:45` | The solve passes `false`; `addAlbumTracks` already passes `info.VA` at `album_tracks.go:26` |
| `AlbumPair{Item ItemFields; Info *TrackInfo}` | `album_tracks.go:11-14` | The pairing type the extras components count against |
| `MostCommonTags` returning `(Likelies, consensus)` | `likelies.go:47` | 3d needs both halves: `Likelies.MBAlbumID` and `consensus["mb_albumid"]` drive `match_by_id` |
| `IsVariousArtistsName` | `track_artist.go:31` | Exported at 3c for `vaLikely` |
| `ProjectItemFields` | `info.go:189` | The corpus-case decoder tier 1 reuses |

**`internal/engine/musicbrainz` (the seam).** `NewClient(ClientConfig)`
at `client.go:124` requires a `UserAgent` and a `Budget`; `Release`
at `:215`, `SearchReleases` at `:255` (ids in response order),
`AlbumForID` at `:296` (fetch, `TranslateAlbum`, conditional pseudo
second fetch, `MergePseudoRelease`). `ErrNotFound` (`:72`) and
`ErrUnavailable` (`:84`) are the two sentinels 3d branches on.
`AlbumSearchFor(items, tun)` at `query.go:79` returns
`AlbumSearch{Query, Limit}` and computes `vaLikely` internally at
`:108`. The limiter ships `NewLimiter` (`limiter.go:110`),
`Limiter.Budget(name, class, limit)` (`:139`), `Budget.Wait` (`:372`)
and `Unmetered()` (`:354`), with `DefaultInterval = time.Second`
(`:62`) and the three classes at `:22-26`.

**Corpus census, measured off the tree today (215 cases).** 167
cases exercise the ID-match early return (no search logged,
recommendation `strong`); 48 exercise the search fallback. The
correspondence is exact in both directions, verified case by case.
All 215 carry a full `mb_albumid` consensus. Candidate counts: 1 case
with 0 candidates, 193 with 1, 4 with 2, 9 with 5, 8 with 6.
Recommendation branches reached, computed against each record's own
thresholds: strong-distance 173, medium-distance 33, single-candidate
low 3, gap low 1, the no-conclusion early return 4, no-candidates 1.
All six exits of `_recommendation` (`match.py:279-323`) are
corpus-reachable. **The `max_rec` downgrade fires on 21 cases**, both
default keys reached (`missing_tracks` and `unmatched_tracks`); 6 of
those sit on the strong-distance branch and downgrade to medium,
which is the branch a port that reads the wrong penalty view gets
wrong silently. (Note for the plan: section (a)'s independent branch
census counts exactly 6 total downgrade firings, all strong→medium,
rather than 21 with 6 as a subset; see (d) for the flagged
discrepancy between these two figures.)

### 2. The assignment solve

**What beets does.** `assign_items` (`match.py:53-81`) builds
`costs[i][t] = float(track_distance(i, t))` with `incl_artist`
defaulting to `False` (`distance.py:383-385`), calls
`lap.lapjv(np.array(costs), extend_cost=True)` and takes the third
return, indexed by track column, valued as an item index or `-1`.
The mapping dict is built by zipping that array with `tracks`, so
**the pair list is in track-index order**, which fixes the `tracks`
penalty summation order and the order of `Candidate.Tracks` in every
committed record. Verified against `complete-combined-over-scrambled`,
whose item-to-track map `{0:3, 1:0, 2:1}` sorts to pairs
`[(1,0),(2,1),(0,3)]`.

**Two decisions, and I recommend separating them.**

The first is *an exact optimum*. Probed against the pinned solver,
rectangular calls are genuinely optimal over `min(items, tracks)`
pairs (`[[0,0.9],[0.9,0],[0.5,0.5]]` returns cost 0.0 dropping the
third row; `[[10,0],[0,10],[5,5]]` returns cost 0.0). Padding to
square with any constant preserves that, because every complete
assignment consumes exactly `|rows-cols|` dummy cells at equal cost.
So a plain Hungarian or a plain JV over a zero-padded square matrix
returns a correct optimum. `ties.json` says 212 of 215 rows have a
unique optimum, 1 is skipped, and 2 are tied. **A correct exact
solver with any tie rule reproduces 212 of the 214 measurable rows.**

The second is *lapjv's choice among equal optima*, which is 2 rows.

**Recommendation: build the exact solver first, as its own task, and
treat lapjv mirroring as a second, bounded task under ruling F.** The
evidence that the tie rule is not derivable from a documented rule is
direct. Probing `lap` 0.5.13 (`corpus/manifest.json:6`, a compiled
extension with no source in the venv,
`tools/venv/.venv/lib/python3.13/site-packages/lap/_lapjv.cpython-313-x86_64-linux-gnu.so`),
an all-equal `n x n` matrix returns `n=2 [1,0]`, `n=3 [2,0,1]`, `n=4
[3,0,1,2]`, `n=5 [4,0,2,1,3]`, `n=6 [5,0,3,2,1,4]`. The shift-by-one
rule that fits `n<=4` breaks at `n=5`. Rectangular all-equal cases
are worse: `5x2` assigns items 1 and 3, not 0 and 1. A Hungarian
solver with a "lowest column index wins" tie rule returns the
identity on both pinned cases and fails both rows, because both
committed mappings are non-identity: `should-ask` swaps items 1 and 6
across 12 tracks, `should-reject` runs a three-cycle `5->7, 7->9,
9->5` across 14. So the honest choice is a faithful port of lapjv's
pivot order (column reduction, reduction transfer, augmenting row
reduction, then augmentation) or the escalation.

Encouraging evidence for the port attempt: a synthetic 12x12 matrix
with a uniform tied block at rows and columns 1 and 6 returns exactly
`[0,6,2,3,4,5,1,7,8,9,10,11]`, the committed `should-ask` mapping.
That is one data point, not a proof, and it is what makes two
attempts a fair budget rather than a formality.

**Go surface, `internal/engine/matcher/assign.go`:**

```go
// Assignment is one solved item-to-track matching.
type Assignment struct {
    Pairs       []Pair // in track-index order (match.py:72-76)
    ExtraItems  []int  // item indexes, sorted by (disc, track, title)
    ExtraTracks []int  // track indexes, sorted by (index, title)
}

// Pair names one item index and the track index assigned to it.
type Pair struct{ Item, Track int }

// AssignItems is assign_items (match.py:53-81).
func AssignItems(tun Tunables, items []ItemFields, tracks []TrackInfo) Assignment
```

Indexes rather than values, because Python keys the mapping on object
identity (`id(item)`, `derive_oracle.py:184-189`) and Go
`ItemFields` is a comparable struct value: two identical items would
collapse into one map key. The index form is also what
`cassette.Expected.Assignment` grades (`cassette.go:84`). The solver
itself is unexported (`solveLAP(costs [][]float64) []int`), takes no
`Tunables`, and gets its own table-driven test file plus the two
pinned matrices as fixtures.

**Ruling F's escalation shape, written so the plan carries the
trigger by name.** Two failed mirroring attempts stop the task. The
finding artifact carries: both cost matrices at full float64
precision, both solvers' optima and their costs, the committed
mapping, the Go mapping, and the demonstration from `ties.json` that
both optima produce the same album distance and the same
recommendation (`corpus/analysis/ties.json`, the `should-ask` and
`should-reject` reasons say so explicitly). **The disposition I
would recommend to Geoff, and the one the finding should offer
rather than assume, is re-dispositioning both rows from `pinned` to
`excluded`.** The tie ledger's own semantics already price that:
excluded means the assignment and the album distance go ungraded
while the verdict is still graded (spec, the dispositions table).
Both cases carry empty `extra_items` and `extra_tracks`, so the
entire graded loss is two `assignment` fields. The cost is that the
ties-excluded count moves from 0 to 2, and that count "may not grow
without an owner ruling", which is precisely the ruling being asked
for. `disposition` is owner-only and the pass never writes it.

### 3. Extras and their sort orders

`extra_items` is `list(set(items) - mapping.keys())` sorted by
`(i.disc, i.track, i.title)`; `extra_tracks` is `list(set(tracks) -
set(mapping.values()))` sorted by `(t.index, t.title)`
(`match.py:77-80`). The pre-sort order comes from a Python `set` over
identity-hashed objects, so the committed order is deterministic only
while the sort key is total.

Measured off the tree: 3 cases carry more than one extra item and 8
more than one extra track; the largest is
`real-nonlatin-hamasaki-ayumi` at 33 extra items. (These are the
counts above the single-extra threshold; section (b) separately
counts 7 cases with any non-empty `extra_items` and 14 with any
non-empty `extra_tracks`, so the two figures describe different cuts
of the same population and are not in tension.) **No case has two
extras tying on their full sort key.** I recommend 3d ship a guard
test that asserts this over every committed case, so the day lane E1
lands a case with a duplicate key the gate reds rather than baking a
hash-order expectation into a record. Go's `slices.SortFunc` with
`cmp.Or(cmp.Compare(...), ...)` reproduces the key order exactly; Go
string comparison is byte-wise over UTF-8 and Python's is by code
point, and the two agree for all valid UTF-8, so no divergence is
owed there.

### 4. Admission filters, insertion order, stable sort

`_add_candidate` (`match.py:331-380`) applies five gates in a fixed
order that 3d must not reorder, because a later gate never runs on a
candidate an earlier one dropped and the distance is computed
between gates 3 and 5:

1. zero tracks drop (`:345-347`)
2. duplicate drop, `info.album_id and info.identifier in results`,
   identity being `(data_source, album_id)` (`hooks.py:164-167`,
   `match.py:350-352`)
3. `match.required`, `getattr(info, req_tag) is None` (`:355-359`)
4. `assign_items` then `distance` (`:362-367`)
5. `match.ignored` against `[key for key, _ in dist]`, which iterates
   `Distance.items()` and so is the **nonzero** view
   (`distance.py:190-205, 240-241`; `match.py:370-375`)

Gate 3 is the subtle one for the port. Python tests `is None`, and
3a's `AlbumInfo` fields are plain Go strings and ints whose absent
form is the zero value, not nil. A Go `Required` filter therefore
tests the zero value, which is a divergence on a source that can
produce an empty-but-present field. 3b's guard 2 already settled that
no such source exists at rung 3, and the same reasoning covers this;
the filter's doc comment should say so and name the second-source
condition that would reopen it.

**Go surface, `internal/engine/matcher/candidate.go`:**

```go
// A Candidate is one admitted release with its solved pairing and its
// accumulated distance, beets' AlbumMatch (match.py:142-201).
type Candidate struct {
    Info       *AlbumInfo
    Dist       *Distance
    Assignment Assignment
}

// Candidates accumulates admitted candidates in insertion order under
// beets' (data_source, album_id) identity key, reproducing tag_album's
// own results dict (match.py:414).
type Candidates struct{ /* order []identity; byID map[identity]Candidate */ }

// Add scores info against items and admits it, reporting whether it was
// admitted. A drop is not an error: beets logs and returns.
func (c *Candidates) Add(tun Tunables, items []ItemFields, info *AlbumInfo, evalDate time.Time) bool

// Sorted returns the candidates by ascending distance, stably, so an
// exact distance tie keeps insertion order (match.py:326-328).
func (c *Candidates) Sorted() []Candidate
```

`Candidates` is an accumulator rather than a builder, and it earns
its state: the ID stage and the search stage add into the same set,
so the duplicate rule spans both. It stays in `matcher` because `Add`
does no I/O.

**The stable sort is corpus-graded, and the mutation is cheap.**
Exactly two committed cases carry an exact candidate-distance tie,
`should-ask` (0.23868312757201643 at positions 1 and 2) and
`should-reject` (0.3891203703703704). Swapping
`slices.SortStableFunc` for `slices.SortFunc` is a legitimate
mutation that fires on those two, and it is the proof the assertion
discipline wants. Insertion order on the search path comes from
`SearchReleases`' response order, which 3c already grades as C14.

### 5. Recommendation and the max_rec downgrade

```go
// Recommendation is beets' Recommendation (match.py:39-47). The values
// are ordered so min() reproduces the max_rec downgrade.
type Recommendation int

const (
    RecNone Recommendation = iota
    RecLow
    RecMedium
    RecStrong
)

func (r Recommendation) String() string // "none" | "low" | "medium" | "strong"

// Recommend is _recommendation (match.py:269-323). cands is Sorted's
// output; an empty slice returns RecNone.
func Recommend(tun Tunables, cands []Candidate) Recommendation
```

Three things a port gets wrong and this design pins:

- The no-conclusion branch **returns before the downgrade**
  (`match.py:300-302`). Four committed cases take it. A port that
  falls through to the downgrade loop still returns `none`, so the
  defect is invisible on the corpus; it needs a hand-authored fixture
  where a `max_rec` key would otherwise lower a value that is already
  `none`, or simply an assertion that the function returns at that
  point.
- The downgrade key set is `set(min_dist.keys())` **unioned with
  every track sub-distance's keys** (`match.py:306-309`).
  `Distance.keys()` is the nonzero view (`distance.py:246-247`), so a
  key added at 0.0 never downgrades. Under shipped `max_rec` the
  track union is inert, because neither `missing_tracks` nor
  `unmatched_tracks` is ever a track-level key; the union is
  therefore a declared gap closed by a hand-authored `Tunables` whose
  `MaxRec` names `track_length`.
- `MaxRec`'s values are level names and beets resolves them through
  `as_choice` (`match.py:313-320`), which raises on an unknown name.
  Go needs `parseRecommendation(string) (Recommendation, error)`, and
  the error belongs at config-load time beside the `Preferred` regex
  validation below rather than mid-scoring.

### 6. The verdict type and its action mapping

```go
// Action is the decision beets takes unprompted (importer.Action).
type Action string

const (
    ActionApply Action = "apply"
    ActionSkip  Action = "skip"
    ActionAsIs  Action = "asis"
    ActionAsk   Action = "ask"
)

// ActionFor is _summary_judgment (beets/ui/commands/import_/session.py:328-360),
// with the shim's "ask" standing for beets' None return.
func ActionFor(tun Tunables, rec Recommendation) (Action, error)
```

The branch order is quiet, then timid, then `rec == none`, then ask.
**The quiet check comes first, so `Timid` and `NoneRecAction` are
unreachable whenever `Quiet` is set.**

**The trap the plan must name loudly.** `derive_oracle.py:193-206`
sets `config["import"]["quiet"] = True` before calling the mapping,
because dubplate imports with `beet import -q`. `Tunables.Quiet`
defaults to `false` (`tunables.go:186`). So a tier 1 run that grades
`action` against the record while using unmodified default tunables
computes `ask` for all 215 cases. The failure is loud rather than
silent, but it costs a task if it is discovered rather than planned:
**the conformance run overlays `Quiet: true` on the merged tunables,
and the reason is the shim's own line.** `QuietFallback` at its
`skip` default is the only fallback the corpus reaches; `asis`,
`Timid`, and `NoneRecAction` are a declared corpus-coverage gap
closed by a table-driven test in `verdict.go`'s own file.

### 7. The compat switches, and the largest open question in the pass

Read off the tree today, **`corpus/analysis/oracle_bugs.json` carries
zero pending hypotheses.** Four are `oracle-wrong` and two are
`oracle-correct`, all adjudicated by Geoff. Ruling A's scheduling
rule therefore leaves every one of the four switches defaulting
**off**, and the grading table's `oracle-wrong` row requires each of
their 20 probe cases to produce the corrected verdict on its own
`adjudications.json` row.

**B3's rule fires cleanly on the tree as it stands.** Every probe
case of an `oracle-wrong` hypothesis now resolves to an adjudication
row by `subject.case_id`. The three cases spec `:300` named as
missing rows do resolve: `real-nonlatin-alla-pugacheva-vol1` and
`real-nonlatin-x-japan-art-of-life` carry `oracle-correct` rows, and
`real-nonlatin-utada-first-love` belongs to `bug-locale-alias-ignored`,
which is `oracle-correct`. So `CheckOracleBugs`' new rule ships
green and its fire-proving fixture must be hand-authored in
`analysis/testdata`, never a corpus case. Its signature changes from
`adjudicationIDs map[string]bool` to a subject index, and the
existing `adjudication_refs` check still needs the id set, so the
honest shape is one loaded ledger passed in rather than two maps.

**Only two of the four fixes are implementable and gradable as
designed.**

| Hypothesis | Probes | The "off" behavior | Gradable at rung 3? |
| --- | --- | --- | --- |
| `bug-ft-pattern` | 6 | `SDPatternsWordBoundary` | Yes, `defect_free` blocks exist on all six rows |
| `bug-remix-parenthetical` | 4 | `SDPatternsNoBracketDiscount` | Yes, all four rows |
| `bug-nonlatin-alias-selection` | 4 (2 rows `oracle-wrong`, 2 `oracle-correct`) | Read `CreditPart.Name` in place of `part.Artist.Name` at `credit.go:53` | The corrected verdict is an **applied artist string** ("artist and albumartist: 소녀시대"), and rung 3 computes no applied field. It is visible in `AlbumInfo.Artist`, so it is gradable only through 3b's translation tier, not through tier 2 |
| `bug-assignment-rejection-gap` | 6 | A per-pair rejection mechanism beets does not have | No. Ruling A says so itself: "the fix has no beets counterfactual". The corrected verdicts name distances like 0.052222 reachable only by inventing rejection |

**Two more mechanical consequences nobody has recorded.**

First, **the two string switches are not independent under a
single-table seam.** `Tunables.StringDistPatterns` holds one table,
and neither `SDPatternsWordBoundary` nor
`SDPatternsNoBracketDiscount` fixes both defects. Both switches
defaulting off needs a fourth table, the word-boundary-guarded
featuring pattern with the two 0.3-weight bracket rows removed. This
is 3b's "pattern-table selection" carry arriving with a name. I
measured whether building it invalidates the committed `defect_free`
figures: across all ten probe cases, on both the item side and the
response side, no `adv-ft-*` string carries a bracket or parenthesis
and no `adv-remix-*` string carries a featuring trigger. So the
combined table reproduces both sets of figures, and 3d should
re-derive to confirm rather than take the measurement's word.

Second, **the default mode's blast radius over the other 205 cases
is unmeasured and is not small.** By substring presence, 133
committed cases carry a bracket or parenthesis somewhere in a
compared string and 44 carry a featuring trigger that the
word-boundary guard would refuse. Those are upper bounds, because
`StringDist` only shifts when the deletion lowers the distance
(`strdist.go`'s `caseDelta == 0` continue, matching
`distance.py:110-111`), so a parenthetical present on both sides is
inert. The real count is a cheap measurement: run the corpus twice
and diff. **It should be a named deliverable of the default-mode
leg, and any case whose verdict (rather than distance) moves without
an adjudication row is a deliberate divergence needing a pending
row.**

**Go surface, `internal/engine/matcher/compat.go`:**

```go
// Compat selects, per ledgered defect class, whether the matcher
// reproduces pinned beets' defect or the adjudicated corrected
// behavior. A true field means "reproduce beets".
type Compat struct {
    FtPattern          bool
    RemixParenthetical bool
    AliasCreditName    bool
    AssignmentRejection bool
}

// DefaultCompat returns the selection the ledger's own dispositions
// imply, as literals with the hypothesis id cited per field. The engine
// never reads corpus/; conform cross-checks these against
// oracle_bugs.json.
func DefaultCompat() Compat

// Apply returns tun with c's string-heuristic selection installed.
func (c Compat) Apply(tun Tunables) Tunables
```

The cross-check between `DefaultCompat` and `oracle_bugs.json`'s
dispositions is the check worth having, on M1's precedent: it fires
the day Geoff moves a disposition and nobody moves the engine.
`AliasCreditName` is consumed in
`internal/engine/musicbrainz/credit.go`, which means `Compat` has to
reach the translation layer; passing it beside `Tunables` into
`TranslateAlbum` changes a 3b signature, so the cheaper shape is a
`Compat` field on `Tunables` itself, at the cost of muddying
"tunable". I lean to the separate type and the signature change, and
flag it for the conductor.

### 8. The public entry point

**Home.** The flow interleaves I/O with scoring: fetch by ID, score,
and return without searching when the result is strong
(`match.py:425-442`). It cannot be split into "fetch everything, then
score", because the request set is exactly what tier 1's zero-miss
rule grades. So the entry point needs a package that sees both
`matcher` and `musicbrainz`, and `matcher` cannot import
`musicbrainz` because `musicbrainz` already imports `matcher`
(`query.go`, `client.go`, `translate.go`). **Recommendation: a third
small package `internal/engine/tag`, one file, holding beets'
`tag_album` and nothing else.** It keeps `matcher` pure, it stays
inside `internal/engine` so `TestEngineNeverReadsTheClock` and
`TestEngineNeverImportsHarness` (`arch_test.go:141,152`) keep
covering it, and it reads correctly at rung 9's call site. This
deviates from the rung 3 spec's tree listing, which names only
`matcher`, `musicbrainz` and `conform`; the plan states the
deviation, on the precedent of 3b's and 3c's conductor-ruled
tree-column corrections. The alternative is `TagAlbum` inside
`musicbrainz`, which costs the package's stated charter in `doc.go`.

```go
// Proposal is the matcher's answer for one album: beets' Proposal
// (match.py:247-249) with tag_album's two leading return values folded in.
type Proposal struct {
    CurArtist      string
    CurAlbum       string
    Candidates     []matcher.Candidate // ascending distance, stable
    Recommendation matcher.Recommendation
    Action         matcher.Action
}

// TagAlbum is tag_album (match.py:383-468): the matcher's whole public
// entry point. It issues one release lookup per item-consensus album id,
// scores what comes back, and returns without searching when that match
// is strong and tun.Timid is false. Otherwise it searches, fans out over
// the returned ids in response order, and returns every admitted
// candidate sorted by distance.
//
// Every request TagAlbum issues spends one token from src's budget, so a
// caller caps its own MusicBrainz spend by the budget it built the client
// with. evalDate is a parameter and never a clock read.
//
// It returns ErrNoCandidates when nothing was admitted, so a caller can
// tell "searched and found nothing" from a transport failure. A lookup
// that 404s drops that candidate silently, reproducing album_for_id's
// return None (beetsplug/musicbrainz.py:777-781). Every other client
// error is returned wrapped, including musicbrainz.ErrUnavailable, which
// rung 4's shadow branches on to defer rather than record a divergence.
func TagAlbum(ctx context.Context, src *musicbrainz.Client, items []matcher.ItemFields,
    tun matcher.Tunables, compat matcher.Compat, evalDate time.Time) (*Proposal, error)
```

`ctx` first. Inputs are the already-projected `[]matcher.ItemFields`
rather than raw maps, because `ProjectItemFields` (`info.go:189`) is
the boundary decoder and rung 9 will build items from the store, not
from JSON. The budget is spent through the client, not passed
separately, matching `NewClient`'s requirement that a `Budget` be
present (`client.go:130-132`); a conformance run passes
`Unmetered()`. Sentinels: `ErrNoCandidates` is the one new one, and
`musicbrainz.ErrNotFound` / `ErrUnavailable` propagate wrapped. `%w`
is justified on both by a real branching caller (rung 4 defers on
`ErrUnavailable`, rung 9 reports on `ErrNoCandidates`).

**Two branches the corpus cannot reach, and the plan should say so
up front rather than discover them.** `tag_album`'s explicit
`search_ids` branch (`:417-420`) has no rung 3 caller, and 3c already
recorded beets' `search_artist`/`search_name` override as a non-port.
And `match_by_id`'s two early branches (`:258-261`) are unreachable:
all 215 committed cases carry a full `mb_albumid` consensus, so
neither the no-id nor the no-consensus path runs. Both get
hand-authored unit tests in the `tag` package's own testdata, and
both join the declared-gap list.

### 9. Tier 1 and tier 2 in `internal/harness/conform`

The landed conventions to extend rather than reinvent: `ReplayTable`
(tier 0, `component.go:79`), `gradeLedger` (translation,
`translate.go:329`), `GradeRequests` (`request.go:72`) with
`RequestGradeSummary` reporting what was compared rather than what
was visited, `mergeProfile` (`translate.go:193`) for the
`Tunables`-defaults-merged-with-profile sourcing M1 requires, and
`CheckCaseWeights` / `CheckCaseThresholds` (`defaults.go:21,43`) for
the echo cross-check, already green over 215 cases.

```go
// GradeMode selects which of ruling A's two runs this sweep is.
type GradeMode int

const (
    // ModeCompat reproduces every ledgered defect: bit-identity against
    // every committed record, including the adversarial stratum.
    ModeCompat GradeMode = iota
    // ModeDefault takes the ledger's own dispositions: each probe case of
    // an oracle-wrong hypothesis is graded against the corrected verdict
    // on its own adjudications.json row, and no other case's verdict is
    // graded, because no other case has a corrected expectation.
    ModeDefault
)

func GradeCases(root string, mode GradeMode) (CaseGradeSummary, error)

type CaseGradeSummary struct {
    Cases          int
    Candidates     int      // candidate comparisons, not candidates visited
    Distances      int
    Assignments    int
    TiesExcluded   int      // reported every run; may not grow without an owner ruling
    Pending        []string // case ids whose verdict a pending row leaves ungraded
    NearMisses     []NearMiss
    VerdictMoves   []string // ModeDefault only: non-probe cases whose verdict moved
}
```

Per-disposition grading follows the spec's table verbatim.
`proximity` flagged demands bit-identity with no epsilon and no
pending escape; `proximity` skipped demands the Go run also produce
no candidate, which the one zero-candidate case
(`admit-zero-track-dropped`) exercises through `_add_candidate`'s
zero-tracks gate. `ties` excluded skips the assignment and album
distance and still grades the verdict, and the excluded count is a
reported field so its growth is visible; today it is zero.
`adjudications` `no-correct-answer` skips the verdict and keeps the
distance, and no committed row carries that disposition, so its
checker needs a hand-authored fixture.

**The pending-divergence path.** A tier 2 verdict disagreement whose
resolution is an owner question is recorded as a row in
`adjudications.json` with `disposition: pending-adjudication` and
`adjudicator: agent`, which both ledger checkers already accept
(`adjudication.go:181`). While it stands the case's verdict is not
graded, `check.sh` stays green, and the launch gate refuses it. The
escape is tier 2 only. Tier 0, tier 1, and every `proximity`-flagged
case admit no pending state, because a mechanical disagreement is a
Go defect until an owner says otherwise.

**The stub's blind side (R6).** `stub.Server.ServeHTTP` records a
`Miss` and a `Consumed` entry but never records a hit from the
request map (`stub/stub.go:122-137`), so a client issuing fewer
requests than the log records leaves zero misses. 3c covered it
inside 3c with a per-case issued-count assertion, which 3d's flow
invalidates: 3d decides the request set rather than reading it from
the log, so the issued count is no longer knowable in advance. **The
close is a `Unused() []cassette.Request` accessor on the stub, the
complement of the hit set, and a tier 1 rule that the union of
`Unused()` and `Misses()` is empty on every case.** That is the
assertion 3c's carry actually wants, and it reads better at the call
site than a hit set. The name was already contemplated at 3c and
deliberately deferred.

**Floors and mutations.** Reuse the floor discipline `request.go:27-31`
established, floors rather than exact counts because lane E1 adds
cases. Proposed floors, measured off the tree today: 215 cases, 214
assignments graded, 1 zero-candidate case, 21 max_rec downgrades
(see the flag on this figure, (d)), 2 stable-sort ties, 2 pinned tie
rows, 0 excluded tie rows, 10 probe cases with a `defect_free` block,
20 probe cases of `oracle-wrong` hypotheses. Named mutations that
must fire: `SortStableFunc` to `SortFunc` (fires on 2 cases), the
nonzero view swapped for the added-key view in the downgrade (fires
on the strong-branch downgrades), the no-conclusion early return
deleted, the ID-match early return's `rec == strong` guard relaxed
(fires as 167 unexpected searches, all misses), the pair list built
in item order rather than track order (fires on every
scrambled-assignment case), and `AssignItems` returning the identity
(fires on `should-ask` and `should-reject` plus every scrambled
case).

**Wall time.** 3c's request sweep runs 3.34s over 215 cases at
`-count=1` with one stub per case. Tier 1 adds scoring and one to six
candidate translations per case on top of the same stub lifecycle,
so a low-single-digit multiple is the expectation, and two modes
doubles it. `go test ./...` is 15.08s today and `check.sh` 1m25.9s
with gitleaks dominating. If the doubled sweep pushes the module gate
past tolerance, the spec's answer is that it becomes its own leg of
`check.sh` at full scope, never an optional target.

### 10. How the runner reports a red when the owner is unavailable

The runner reports and the pass records. It never adjudicates.

- A tier 0 or tier 1 disagreement is a Go defect. It fails the gate,
  names the case, the field, both values, and where relevant the
  component. There is no pending escape and no owner involvement.
- A tier 2 verdict disagreement writes a pending row in
  `adjudications.json` with `adjudicator: agent`, which is the only
  disposition an agent may set (`adjudication.go:24-30, 181`). The
  gate stays green, the count is reported, and the launch gate
  refuses the row. The pass report names the case and the suspicion,
  because "the pending count may not grow without the growth being
  named".
- The conductor, not Geoff, rules whether a given divergence is lane
  E1's red to clear (revert or land with a pending row) or lane E2's
  defect to fix. That is a judgment call rather than a diff read and
  costs no attended moment.
- A `ties` excluded count that grew, or a `DefaultCompat` field that
  no longer matches its hypothesis's disposition, is a hard red
  rather than a pending row. Both encode an owner decision, and an
  agent silently absorbing one is exactly what M4's sentinel exists
  to prevent.

The one case with no clean answer is a `ModeDefault` run in which a
non-probe case's verdict moves. It is a deliberate divergence by
construction, so it needs a row, and writing 20 or 40 of them
unattended is not obviously right. My recommendation is that
`ModeDefault` reports `VerdictMoves` as a count and a case list, the
pass records it, and the disposition batch goes to Geoff as one
adjudication batch, which is a touchpoint the process already
expects.

### 11. Does 3d take a fresh-context review of the runner alone (spec `:362`)?

**Yes, and scoped to the runner alone.** Three reasons, and none is
generic caution.

The instrument argument is the spec's own: every later claim about
the matcher rests on this runner, and a runner that grades fewer
fields than it claims produces a green gate that means nothing.
Rung 15's re-run criterion inherits it verbatim, and rung 4's shadow
inherits the same grading vocabulary.

The evidence argument is that this failure mode has already occurred
twice in rung 3 and both times the in-chain `diff-reviewer` caught it
only because a plan bullet named the proof. 3a's
`aggregate/sorted-penalties-tied` fixture could not discriminate a
`(distance, key)` sort from no sort at all, and the specified
mutation did not fire. 3b's Task 5 mutation was provably
unobservable. Both were found by mutation testing rather than by
reading the code, which is what a fresh-context reviewer with a
graded-field table can repeat.

The cost argument is that the scope is small and checkable. The
review reads `GradeCases` against `cassette.Expected`'s ten fields
(`cassette.go:76-88`) field by field, confirms every disposition row
of the spec's table has a checker and a fixture, and re-runs the
named mutations. That is one dispatch against a bounded surface, not
the full adversarial cycle rungs 5, 12 and 16 carry. The
`code-simplifier` and pass-end `diff-reviewer` dispatches stay as
they are.

I would not extend it to the engine half. `assign.go` and
`verdict.go` are graded by 214 committed assignments and six
reachable recommendation branches, which is stronger evidence than a
reviewer produces.

### 12. Closing or carrying each named "To 3d" item

| Carry | Disposition in 3d |
| --- | --- |
| `Preferred.Media` / `Preferred.Countries` config-load validation | **Close.** The panics sit at `album_media.go:41` and `album_country.go:36`. 3d adds a `Tunables.Validate() error` that compiles both pattern lists and parses `MaxRec`, `QuietFallback` and `NoneRecAction` through the same name tables `ActionFor` and `Recommend` use, since 3d is the first pass to own all four name parsers. Wrapped error, no panic, one sentinel |
| Compat switch's pattern-table selection | **Close, with a fourth table.** Both string hypotheses are `oracle-wrong`, so both default off, and the single-table seam cannot express both. Build the combined table, confirm by re-derivation that it reproduces all ten committed `defect_free` figures |
| The `\b` regime divergence, travelling with the query path's lowering and `_extract_id`'s `\w` | **Carry, unchanged.** Go RE2's `\b` is ASCII-only and Python's is Unicode-aware. All six `adv-ft-*` pairs are ASCII, so nothing moves. The new combined table inherits the same guard comment. Record the three together as one regime family, as 3c did |
| Tie mirroring needing both `inclArtist` call shapes | **Close.** `AssignItems` passes `false` (`distance.py:383-385`); `addAlbumTracks` already passes `info.VA` (`album_tracks.go:26`). Both shapes are exercised on the same case, and a mutation that makes the solve pass `info.VA` fires on the VA cases |
| The stub's blind side (R6) | **Close** with `Unused()` and the union-empty rule above |
| R1 seam: 3d owns the ID-match early return, `_recommendation`, `_add_candidate`, the explicit `search_ids` branch and `match_by_id`'s no-id/no-consensus branch | **Close for the first three, declare the last two as gaps.** 167 cases grade the early return, 48 the search fallback, all six `_recommendation` exits and both `max_rec` keys are reached. `search_ids` and both `match_by_id` early branches are corpus-unreachable and get hand-authored tests |
| `vaLikely` exported the day `match_by_id`'s no-consensus branch needs it | **Close as "stays unexported."** That branch is corpus-unreachable, and the search path calls `vaLikely` inside `AlbumSearchFor` (`query.go:79`), so no caller outside `musicbrainz` ever needs it. Record the reason so a later pass does not export it for symmetry |
| `Case.Extra` has no producer (R7) | **Carry, unchanged.** A producer is `cmd/harness` work in lane E1's territory. Zero of 215 cases carry an `Extra` entry, so tier 1's zero-miss rule reads against the union of both maps and every consumption is reported rather than failed, exactly as 3c built it |
| `MergePseudoRelease` aliasing (`pseudo.go:49-58`) | **Carry to rung 9, with the trigger named.** `merged.Tracks` aliases the pseudo's backing array. `AlbumForID` closes it by construction because only the merged value escapes, and 3d's `TagAlbum` inherits that same call site unchanged. `slices.Clone` is owed the day a caller retains both values, which is rung 9's pipeline, not 3d |
| Interim `AdminEmail` User-Agent contact (R11) | **Carry.** It belongs beside the four-knob registration, post rung 1b. `TagAlbum` takes a built client and never composes a User-Agent |
| Four-knob wall registration | **Carry unless rung 1b has landed at 3d's staleness pre-task.** `internal/config/keys.go` is the file no lane edits until then, and `Threshold.Default` is a string with `ParseInt` the only parser, so the three float thresholds need a float variant the registry does not yet have |
| The retry path, ungraded by the corpus forever | **Carry.** The stub answers 200 or 404 only, `cassette.Request` carries no status field, and the response-status extension stays declined |
| Ten declared corpus-coverage gaps | **Read the running total, add 3d's own, never re-count.** 3d's additions as I count them: `match.required`, `match.ignored` (both inherited from 3a's three and now consumed for the first time, so they change from deferred to closed-by-unit-test rather than becoming new gaps), `search_ids`, `match_by_id`'s no-id branch, `match_by_id`'s no-consensus branch, the `max_rec` track-key union, `Timid`, `NoneRecAction`, and `QuietFallback: asis`. Seven new, taking rung 3 to seventeen. (Sections (a) and (b) count this differently; see (d).) |
| The rung 15 re-run criterion | **Carry forward verbatim**, as every rung 3 pass close has |

---

## (d) Cross-section consistency and open questions

### Beets rules (a) to Go surfaces (c)

Every mechanic section (a) derives maps to a surface in section (c),
or is named as an explicit non-port/carry, with one exception flagged
below.

| (a)'s rule | (c)'s surface | Status |
| --- | --- | --- |
| Cost matrix, `inclArtist=false`, track-order pairs (§1) | `assign.go`'s `AssignItems`/`Assignment` (§2) | Mapped |
| `extra_items`/`extra_tracks` sort orders (§2) | `Assignment.ExtraItems`/`ExtraTracks` (§2), the totality guard (§3) | Mapped |
| Admission filter order and `ignored`'s nonzero view (§3) | `candidate.go`'s `Candidates.Add` (§4) | Mapped |
| `tag_album` flow, `match_by_id`, early return (§4) | `tag` package's `TagAlbum` (§8) | Mapped |
| `_recommendation`, thresholds, `max_rec` downgrade (§5) | `verdict.go`'s `Recommend` (§5) | Mapped |
| `_summary_judgment` (§6) | `verdict.go`'s `ActionFor` (§6) | Mapped |
| Compat switches, pattern tables (§7) | `compat.go`'s `Compat`/`DefaultCompat` (§7) | Mapped |
| Two `inclArtist` call shapes (§8) | Closed at 3a/3c per (c)'s closing table (§12) | Mapped, already closed |
| Ledger dispositions (§9) | `GradeMode`, `CaseGradeSummary`, the pending-divergence path (§9, §10) | Mapped |
| `lapjv` solver mechanics, extension model, pivot order (§10) | `assign.go`'s unexported solver, ruling F escalation (§2) | Mapped |
| The two pinned tied cases and the 17 unledgered ties (§11) | Ruling F's mirroring-test design (§2, four layers); the 17 unledgered ties surface only as an open question, not a designed Go check | **Partial.** (c) does not propose a mechanism inside the runner for the 17 unledgered ties beyond citing the risk; it is carried entirely as an open question (item 2 below) rather than as a stated declared-gap or fixture obligation the way other corpus-unreachable surfaces are handled elsewhere in (c) |
| Float-ordering hazards #1-11 (§12) | Distributed: #1-2 in `assign.go`/track-order pairing (§2), #3-4 in the solver and its differential test (§2), #5 in the extras guard (§3), #6-7 in `verdict.go` (§5), #8-9 carried from 3c/3b unchanged, #10 in `candidate.go`'s `required` discussion (§4), #11 needs no Go action | Mapped |
| New declared corpus-coverage gaps (§13) | (c)'s closing-table gap count (§12) | **Disagreement in count**, not coverage; see item 4 below |

### Ledger figures (c) depends on, checked against (b)'s tree reading

- Case count and profile split (215; 178/36/1): (c) §1 and (b) §4
  agree.
- Candidate counts (1 case 0 candidates, 193 with 1, 4 with 2, 9 with
  5, 8 with 6, 294 total): (c) §1 and (b) §4 agree exactly.
- Assignment grading population (214 graded, 212 unique + 2 pinned, 1
  skipped, 0 excluded): (c) §2 and (b) §4 agree.
- `ties.json`/`proximity.json`/`adjudications.json`/`oracle_bugs.json`
  dispositions (2 pinned/0 excluded/1 skipped; 1 flagged/1 skipped;
  18 oracle-wrong/3 oracle-correct/0 pending; 4 oracle-wrong/2
  oracle-correct/0 pending): (c) §7 and (b) §4 agree exactly.
- Compat-switch gradability (two of four hypotheses deliverable): (c)
  §7's table and (a) §9's three-consequence list agree.
- **`max_rec` downgrade case count: disagreement.** (a) §5's branch
  census states the downgrade fires on **exactly 6 cases**, all
  strong→medium, and the arithmetic 173 − 6 = 167 depends on that
  figure. (c) §1 states the downgrade "fires on 21 cases, both
  default keys reached ... 6 of those sit on the strong-distance
  branch and downgrade to medium," implying 21 total firings of which
  6 are one subtype. The 21 figure is very likely (b) §4's "21 cases
  exercise candidate ordering and the gap threshold" (a fact about
  multi-candidate cases, not about the downgrade) read into the wrong
  sentence. The plan should re-measure this directly against the tree
  before sizing the verdict layer's floor, since (c) §9's own proposed
  floor list repeats "21 max_rec downgrades" while (c) §9's own
  mutation-proof list separately anchors to "the strong-branch
  downgrades" without repeating the number. Flagged rather than
  resolved here.
- Extras figures: (b) §4's "7 cases carry non-empty `extra_items`
  (max 33), 14 carry non-empty `extra_tracks` (max 4)" and (c) §3's
  "3 cases carry more than one extra item and 8 more than one extra
  track" are not in tension. They describe different thresholds
  (non-empty versus more-than-one) over the same population and are
  restated together in (c) §3's parenthetical.

### Corpus figures, restated once

215 cases (178 estate, 36 stock, 1 locale). Ledger rows: `ties.json`
215 (212 unique, 2 pinned, 0 excluded, 1 skipped), `proximity.json` 2
(1 flagged, 1 skipped, 213 clear), `adjudications.json` 21 (18
oracle-wrong, 3 oracle-correct, 0 pending), `oracle_bugs.json` 6 (4
oracle-wrong, 2 oracle-correct, 0 pending), `components.json` 72,
`album_info.json` 50, `file_tags.json` 12. 294 candidates across 214
candidate-bearing cases (1 case carries none); 87 rectangular
candidates, 21 cases rectangular on their best candidate. Verdict
distribution: 167 strong/apply, 39 medium/skip, 4 low/skip, 5
none/skip. Two candidate-distance ties among cases (`should-ask`,
`should-reject`), both also the two pinned tie-ledger rows, with 17
further tied candidate assignments outside those two carrying no
disposition at all. Seven cases carry a non-empty `extra_items` (max
33), fourteen a non-empty `extra_tracks` (max 4), none both. The
`max_rec` downgrade's case count is disputed between sections; see
above.

### Combined open questions

Deduplicated across (a)'s O1-O9, (b)'s nine unlabeled questions, and
(c)'s six unlabeled questions. Each entry names every section that
raised it.

1. **`bug-assignment-rejection-gap` (and `bug-nonlatin-alias-selection`)
   have no gradeable rung-3 expectation once ruled `oracle-wrong`.**
   Does the plan carry a compat switch for the class that stays on
   with a written improvements-register deferral, ask Geoff to
   re-disposition the hypothesis, invent a `defect_free`-shaped
   expectation the class cannot actually express, or ship both
   switches but grade only the two string classes while recording the
   other two as ungraded with reasons? Raised in (a) as O1, (b)'s
   first question, and (c)'s first question, the last of which adds
   the three-option framing and a recommendation.
2. **The tie ledger measures `Candidates[0]` only, leaving 17 tied
   candidate assignments elsewhere with no disposition**, five of
   which change tier-1's graded `Tracks` vector on an alternative
   optimum and two of which also change the `tracks` penalty sum.
   Does 3d extend the tie detector to every candidate (a
   corpus/derivation change needing an owner disposition batch), grade
   `Tracks` only for the best candidate (weakening tier 1), or record
   the 17 as a named risk and rely on the Go solver mirroring `lapjv`
   exactly? Raised in (a) as O2.
3. **`match.required`'s `is None` test has no Go analogue** on
   `matcher.AlbumInfo`'s plain-typed, no-nil fields. Which field names
   does the filter accept, what does "None" mean per field kind, and
   is an unknown name a config-load error or a runtime failure? Raised
   in (a) as O3.
4. **Both string-heuristic switches defaulting off needs a table
   neither existing table expresses.** Does 3d build a fourth composed
   table, or represent the compat surface as independent per-class
   edits composed into a table at run time? Raised in (a) as O4.
5. **Where does `Compat` live: a field on `Tunables`, or a separate
   type threaded as its own parameter into `TranslateAlbum`?** The
   `AliasCreditName` switch's fix site is in
   `internal/engine/musicbrainz/credit.go`, so `Compat` must reach the
   translation layer either way. Related to item 4 but a distinct
   design question. Raised in (c)'s third question.
6. **Ruling A's mechanism needs two named runs** (full compat for
   bit-identity, a restricted default-mode run for corrected
   verdicts), not one. Should the plan state this explicitly as two
   runs with their own case sets rather than leaving it implied?
   Raised in (a) as O5.
7. **In a `ModeDefault` run, what happens when a non-probe case's
   verdict moves?** Up to 133 cases carry a bracket or parenthesis and
   up to 44 a featuring trigger the word-boundary guard would refuse,
   though the real moved count (after `StringDist`'s
   only-if-it-lowers-the-distance rule) is expected to be small and is
   unmeasured. Is reporting `VerdictMoves` as a count and case list,
   with the batch going to Geoff as one adjudication batch, the right
   shape? Related to item 6. Raised in (c)'s fourth question.
8. **The rectangular extension model (`max(rows, cols)` zero-padded)
   is established behaviorally, not from source, and disagrees with
   `derive_ties.py`'s own docstring 30% of the time on synthetic
   matrices.** Should 3d implement the measured model plus an
   oracle-tagged differential test that re-establishes it against the
   pinned `lap`, and correct the docstring (a comment-only edit to an
   oracle-tagged file that cannot itself be gate-verified)? Raised in
   (a) as O6.
9. **Two of the spec's own premises have moved since it was written**:
   B3's "three probe cases with no adjudication row" is now zero
   (both resolve to `oracle-correct` rows), and ruling A's "five
   hypotheses pending" is now zero pending (four oracle-wrong, two
   oracle-correct). Should the plan confirm both against the tree at
   dispatch, per the staleness pre-task, rather than against the
   spec's own text? Raised in (a) as O7.
10. **`TieEntry.Adjudicator`'s mechanism**: does the oracle-tagged
    derivation (`writeTieLedger`, the diff test's field mask) carry the
    new field forward the way it carries `Disposition` and `Reason`,
    and do the two committed pinned rows take `agent` by hand-edit
    since the gate cannot run the oracle-tagged carry-forward proof?
    Raised in (a) as O8 and (b)'s fourth question.
11. **Does 3d take a fresh-context review of the conformance runner
    alone**, scoped to a graded-field table checked against
    `cassette.Expected`? (c) §11 already argues yes with three
    reasons; the open-question lists in (a) and (b) still carry it as
    a decision for the plan to state rather than assume. Raised in (a)
    as O9 and (b)'s eighth question.
12. **How does verdict-to-action grade given that every committed
    `action` was derived with `import.quiet` forced true**, making
    `import.timid` and `import.none_rec_action` structurally
    unreachable? Does the conformance run grade with `Quiet` forced,
    model the shim's forcing explicitly, or take a `Tunables` default
    that differs from beets' shipped value (which ruling D forbids)?
    Raised in (b)'s second question.
13. **Where does the `Quiet: true` overlay live syntactically**: inside
    `mergeProfile` beside the profile-name table, or as an explicit
    argument to `GradeCases` so the divergence between the shim's
    config and the shipped default stays visible at the call site?
    Related to item 12, a narrower syntactic question. Raised in (c)'s
    fifth question.
14. **The `max_rec` downgrade's `keys()`-versus-`_penalties` distinction
    is corpus-blind**: no committed candidate carries `missing_tracks`
    or `unmatched_tracks` at 0.0. Is the proof a hand-authored fixture
    in `matcher`'s own testdata, a declared corpus-coverage gap, or a
    corpus-lane ask for a case that would reach it? Raised in (b)'s
    third question.
15. **Does B3's probe-case rule live inside `CheckOracleBugs`** (as the
    spec places it) **or as a widening of `CheckDefectFree`'s existing
    probe walk**, and does `CheckOracleBugs` take `*AdjudicationLedger`
    in place of `adjudicationIDs map[string]bool` so it can build both
    indexes itself? Raised in (b)'s fifth question.
16. **Who authors a pending-divergence row**: does the runner write it
    into `adjudications.json` directly (which needs a writer the
    module does not have, plus five populated prose fields), or does
    it report the row a human must write and fail nothing? Raised in
    (b)'s sixth question.
17. **Does the stub gain the `Unused()` hit-set accessor**, and is a
    logged-but-never-issued request a tier-1 hard failure or a
    reported finding? (c) §9-10 already recommends closing this with
    `Unused()` plus a union-empty rule, but the open-question framing
    in (b) treats the fail-versus-report choice as still unsettled;
    the plan should state which it is taking rather than let the
    recommendation stand as an assumption. Raised in (b)'s seventh
    question.
18. **Does `profile_hashes` gain entries for `stock` and `estate`**,
    given it pins `locale` alone today and `checkProfileKeyDrift`
    catches a new key but not a changed value in the other two files?
    Raised in (b)'s ninth question.
19. **Where does the public entry point live**: `internal/engine/tag`
    (recommended in (c) §8, deviating from the spec's named tree), or
    inside `internal/engine/musicbrainz` at the cost of that package's
    stated charter? Raised in (c)'s second question.
20. **If ruling F's two mirroring attempts fail, is re-dispositioning
    `should-ask` and `should-reject` from `pinned` to `excluded`**
    (moving the ties-excluded count from 0 to 2) an acceptable price to
    Geoff, given both cases' own `ties.json` reasons already record
    that both optima produce the same album distance and
    recommendation? Raised in (c)'s sixth question.

### The declared-gap count: an unreconciled figure

Three different counts of "how many new corpus-coverage gaps does 3d
add" appear across the three sections, and the plan should settle
which one it adopts rather than average them.

- **(a) §13 lists eleven** new items, including `match.ignored`
  reading the nonzero view and the assignment objective's
  `inclArtist=false` on a VA release, both of which (c)'s closing
  table explicitly says are **not** new gaps because they are closed
  by construction or reclassified from "deferred" to
  "closed-by-unit-test."
- **(b) §5 lists seven items total, of which the first three restate
  3a's existing three gaps and only items 4 through 7 are new**: the
  `max_rec` zero-valued-key blind spot, `import.timid`/`none_rec_action`
  unreachability, `quiet_fallback: asis`, and `profile_hashes`
  covering one profile of three. That is four new gaps by (b)'s own
  count.
- **(c)'s closing table states seven new gaps, taking the running
  total from ten to seventeen**, an internally consistent count that
  explicitly excludes `match.required`/`match.ignored` as
  reclassifications rather than new gaps, but which also does not
  include (b)'s `profile_hashes` item or (a)'s "every non-default
  value of the four recommendation knobs" item.

None of the three lists is a strict subset or superset of another; the
plan authoring 3d needs one canonical list, built by reconciling all
three against the tree at plan-drafting time, rather than adopting any
one section's count as given.
