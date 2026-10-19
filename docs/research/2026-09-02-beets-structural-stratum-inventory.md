# The structural stratum: beets test-shape inventory for pass 2b

Research report, 2026-09-02. Source of truth: the `v2.13.1` tag of
`github.com/beetbox/beets`, commit
`b7952299941543d4507ac7931edb223acd684b3d` (shallow clone under
`/var/home/glw907/.claude/jobs/daef7196/tmp/beets-2.13.1`, matching the
commit the mechanics study cites and `tools/venv/requirements.txt`
pins). Files cited: `test/autotag/test_distance.py`,
`test/autotag/test_match.py`, `test/autotag/test_hooks.py`,
`beets/config_default.yaml`. Grounding for the schema claims:
`internal/harness/cassette/cassette.go` and
`internal/harness/oracle/derive_oracle.py` as shipped by pass 2a
(commit `3112877` and prior). Grounding for the corpus's own design:
`docs/plans/2026-09-02-rung-2a-cassette-substrate.md`,
`docs/research/2026-09-01-beets-conformance-corpus.md`,
`docs/research/2026-09-02-beets-matcher-mechanics.md`.

## Zero: three tiers, not one

beets' own tests operate at three different altitudes, and pass 2b's
three deliverables (structural cassette cases, component-test
derivation tooling, the tie/proximity analysis tools) map onto them
1:1 rather than all funneling into cassette cases:

- **Tier A — component tests.** Call `Distance` methods,
  `string_dist`, or `track_distance` directly with hand-built
  arguments, often under a monkeypatched config value no committed
  profile carries (a custom `distance_weights["data_source"]`, a
  second metadata-source plugin). These never touch `tag_album`, a
  candidate list, or a recommendation. They map to **(c)**: a
  derivation tool that imports the same beets functions and replays
  the same literal input tuples, in-process, independent of the
  corpus/oracle/stub pipeline.
- **Tier B — full-pipeline structural cases.** Item/AlbumInfo shapes
  that a real `tag_album` call would encounter from a synthetic,
  hand-authored MusicBrainz response — same mechanism 2a's smoke
  `testdata/cases/smoke` already proved (a fabricated
  `responses/release.json`, never a live capture). These map to
  **(b)**, the 60-120-case structural stratum proper.
  `internal/harness/oracle/testdata/cases/smoke/case.json` and its
  `responses/release.json` are the existence proof and the template.
- **Tier C — inexpressible at 2b.** Shapes that need something no
  current mechanism supplies: a second loaded metadata source, a
  third profile, or state the schema does not persist. Catalogued in
  **(e)**.

A single beets test function can straddle tiers; several below are
flagged as Tier A only, or split.

## (a) Catalogue of scenario shapes

### From `test/autotag/test_distance.py`

| Shape | Function | Tier | What it exercises |
|---|---|---|---|
| `Distance.add`/`add_equality`/`add_expr`/`add_number`/`add_priority`/`add_ratio`/`add_string` primitive behavior | `TestDistance.test_add` (l.31-34), `test_add_methods` (l.36-89) | A | Each adder's exact penalty-list output for a table of literal inputs — the ground truth for every downstream weighted penalty. |
| Raw/max/normalized distance arithmetic, per-key indexing | `test_distance` (l.91-101) | A | `raw_distance`, `max_distance`, `distance = raw/max`, `dist["key"]` — the normalizer math the mechanics study's Q2 documents. |
| Distance comparison operators, `len`, `list()`/`items()` | `test_operators` (l.103-115) | A | `Distance.__eq__/__lt__/__len__/__iter__` against a fixed penalty set. |
| Penalty display sort (by weighted distance desc, key asc on tie) | `test_penalties_sort` (l.117-126) | A | `items()`'s stable sort — a display/adjudication concern, not verdict-moving. |
| Zero-penalty keys excluded from `generic_penalty_keys` | `test_generic_penalty_keys_excludes_zero_penalties` (l.128-132) | A | The exact rule `max_rec`'s downgrade pass depends on (match.py:304-323): a key added at 0.0 never triggers a downgrade. |
| `Distance.update` merges two penalty dicts | `test_update` (l.134-149) | A | Internal composition only; no matcher-facing shape. |
| Track title/artist basic distance, VA-artist tolerance | `TestTrackDistance.test_track_distance` (l.152-170), 4 params: identical, tolerate-va, different-artist, different-title | **B** | `track_distance(item, info, incl_artist=True)`: VA_ARTISTS tolerance (distance.py:404-410) and title/artist string distance together. |
| Clean album, distance 0 | `TestAlbumDistance.test_identical_albums` (l.211-212) | B | Baseline; already the smoke `clean-match` shape's synthetic analogue. |
| Candidate one track short of the items (one item left unmatched) | `test_incomplete_album` (l.214-217) | B | `unmatched_tracks` penalty (0.6 weight: an item beyond the matched pairs, distance.py:537-538), asserts `0 < distance < 0.2`. |
| Candidate carries one track beyond the items (one candidate track left unmatched) | `test_overly_complete_album` (l.219-224) | B | `missing_tracks` penalty (0.9 weight: a candidate track beyond the matched pairs, distance.py:533-534) — the complementary half of the completeness pair. |
| Album-artist penalty present/absent by VA flag | `test_albumartist`, params `va=[True, False]` (l.226-231) | B | `artist` component skipped entirely when `album_info.va` (distance.py:447-448) — direct test of the VA short-circuit. |
| VA album, all track artists null on both sides | `test_comp_no_track_artists` (l.233-240) | B | Confirms a VA release with genuinely missing per-track credits still scores 0, not a spurious penalty. |
| VA album, one track artist mismatched | `test_comp_track_artists_do_not_match` (l.242-246) | B | The per-track `track_artist` component still fires inside a VA album even though the album-level `artist` component is skipped. |
| Track titles swapped, pairs supplied positionally (not via `assign_items`) | `test_tracks_out_of_order` (l.248-252) | **A only** — see note below | Tests `distance()`'s behavior under a manually mis-paired `item_info_pairs` list; `tag_album`'s real flow always derives pairs from `assign_items`'s optimal solve first, so this exact positional-mismatch shape is unreachable through the oracle's full pipeline. Redundant with the assignment family (below) at Tier B; keep it in the component tool only. |
| Two-medium split, correct per-medium indexing | `test_two_medium_release` (l.254-259) | B | `mediums`/per-disc structure; distance stays 0 when medium_index is set correctly on both discs. |
| String-distance equivalences (11 pairs: casing, punctuation, `(EP)` strip, `"The X"`/`"X, The"` and `A`/`An` article rotation, empty strings, bracket strip, `&`→`and`, diacritics `"\xe9\xe1\xf1"` vs `"ean"`) | `TestStringDistance.test_matching_distance` (l.263-280) | A, ~4-5 ported to B | `string_dist == 0.0` for each pair — the lower-then-unidecode-then-lower pipeline, `SD_END_WORDS` rotation, `SD_REPLACE`. The diacritic pair is the cheapest available proxy for the corpus report's "non-Latin scripts" shape family without a live capture. |
| Distinct strings are nonzero | `test_different_distance` (l.282-283) | A | Sanity floor. |
| Relative weighting of the `SD_PATTERNS` heuristics (`(ep|single)`, `feat/ft`, parens, brackets) | `test_relative_weights` (l.285-295), 4 params | A, 2-3 ported to B | Proves each pattern's stripped-string comparison only helps (never increases) distance — `case_delta = max(0, base_dist - case_dist)`, distance.py:97-117. |
| Pattern-application does not crash on adversarial/empty input | `test_solo_pattern` (l.297-301) | A | `"The "`, `"(EP)"` vs itself, `", An"` — edge inputs to the regex passes, not a matcher-behavior assertion. |
| Data-source penalty: match/mismatch/unknown/no-info/no-original under 1 vs 2 loaded sources; penalty and weight magnitude scaling | `TestDataSourceDistance.test_distance` (l.341-364), 12 params | **A only**, 1 shape salvageable to B | Needs a monkeypatched second `MetadataSourcePlugin` and a per-test `distance_weights["data_source"]`/`data_source_mismatch_penalty` override — neither is expressible under a fixed named profile (see (e)). The one Tier-B-compatible shape it implies but does not literally parametrize: an item whose `data_source` field is pre-set to something other than `"MusicBrainz"` under a single loaded source. Per `add_data_source` (distance.py:357-361), the penalty condition is `before != after and (before is set OR multiple sources loaded)` — "before is set" alone is sufficient, so this fires even single-source. This is a dubplate-derived case (not literally in beets' own table) but still executed through real pinned beets, per the plan's "every expectation is derived by executing pinned beets code paths" rule. |

### From `test/autotag/test_match.py`

| Shape | Function | Tier | What it exercises |
|---|---|---|---|
| Item order corrected by content; unmatched leading/middle/trailing track becomes `extra_tracks`; unmatched item becomes `extra_items` | `TestAssignment.test_assign_tracks` (l.26-55), 5 params | B | `assign_items` (match.py:53-81) directly — the core of the assignment family. Titles-only, 2-3 item albums. |
| A 12-track album with every title *and* length substantially altered still assigns correctly by minimizing total cost | `test_order_works_when_track_names_are_entirely_wrong` (l.57-101) | B | A scaled, harder version of the row above: proves `lap.lapjv`'s optimum survives large per-pair distances as long as the *relative* ordering of costs still favors the true pairing. Good candidate for one deliberately large (10-14 track) structural case. |
| Two metadata sources both yield the shared (data_source, album_id) pair; candidate proposal exposes both sources | `TestTagMultipleDataSources` (l.104-188), 4 tests (`search_album_ids`, `search_album_current_id`, `search_track_ids`, `search_track_current_id`) | **C** | Requires two loaded metadata-source plugins (monkeypatched `find_metadata_source_plugins`); the corpus stays single-source per ruling 2/the mechanics study's Q1.2 and Q1.5. Not expressible at any file-free tier under the current profile design. |

### From `test/autotag/test_hooks.py`

This file tests `AlbumMatch.apply_metadata`/`TrackMatch.apply_metadata`
field-application rules (`per_disc_numbering`, `artist_credit`,
`original_date`, `from_scratch`, `overwrite_null`, list/singular field
reconciliation) — **application-time behavior, not distance or
recommendation**. It is the closest existing beets coverage of what
`Expected.AppliedFields` records, and is genuinely useful as a
cross-check list for 2b's `AppliedFields` assertions (e.g.
`test_per_disc_numbering`, l.242-248, is a direct behavioral spec for
the multi-disc structural cases' applied `track`/`tracktotal` fields),
but it does not itself define new *matcher* scenario shapes — its
inputs are `AlbumMatch`/`TrackMatch` objects built directly, not
`tag_album` candidate proposals. Treat it as an assertion-writing
reference for family 2/7's cases below, not a source of additional
case count.

### Verdict/recommendation logic (not covered by any dedicated test
file; derived from `beets/autotag/match.py` and confirmed against
`beets/config_default.yaml`)

No `test/` file exercises `_recommendation` (match.py:269-323) or
`_summary_judgment` (session.py:328-360) directly as unit tests
(confirmed: `grep -rln "recommendation\|max_rec\|rec_gap_thresh" test/`
returns nothing under `test/autotag/`). This is a real gap in beets'
own suite, and it is exactly what pass 2a's `should-ask`/`should-reject`
smoke cases and the threshold-proximity flag (item d) exist to cover
from first principles rather than by porting a beets test. 2b should
still budget cases for it — see family 15 in (b).

## (b) Shape families → cassette-case budget

Every family below is Tier B: a synthetic `ItemFields` input plus a
hand-authored minimal MusicBrainz response, run through the real
`oracle.Derive` exactly as `testdata/cases/smoke` already does. Counts
are per-family targets, not per-parametrize-row copies — several
`test_distance.py` params collapse into one constructed album that
exercises two or three at once (e.g. one multi-disc album can carry
both a `mediums` shape and a `track_index` per-disc-numbering shape).

| # | Family | Source shapes | Cases |
|---|---|---|---|
| 1 | Assignment/ordering (scrambled order, unmatched leading/middle/trailing track, unmatched item, large-scale garbled-but-recoverable) | test_match.py `test_assign_tracks` (5 params), `test_order_works_when_track_names_are_entirely_wrong` | 8 |
| 2 | Album completeness (candidate 1/2 tracks short of items → `unmatched_tracks`; candidate 1/2 tracks beyond items → `missing_tracks`; combined; one crossing the `max_rec` downgrade boundary) | test_distance.py `test_incomplete_album`, `test_overly_complete_album`; match.py:304-323 | 7 |
| 3 | VA / artist handling (VA exact, VA mismatched track artist, VA missing track artists both sides, non-VA wrong artist, non-VA correct artist, single-artist album misflagged VA-likely via `VA_ARTISTS` string) | test_distance.py `test_albumartist`, `test_comp_no_track_artists`, `test_comp_track_artists_do_not_match`; distance.py `VA_ARTISTS` (l.29) | 7 |
| 4 | String-distance edges ported to full pipeline (article rotation, `(EP)` strip, `feat.` strip, `&`/`and`, one diacritic-transliteration pair) | test_distance.py `test_matching_distance`, `test_relative_weights` (curated subset) | 5 |
| 5 | Single-source `data_source` penalty (item pre-tagged with a foreign `data_source`) | test_distance.py `TestDataSourceDistance` (derived shape, see (a)) | 2 |
| 6 | Media/mediums/disc structure (disctotal mismatch, media-format mismatch, uneven disc split, correct two-medium baseline synthetic analogue of the smoke `multi-disc` case) | test_distance.py `test_two_medium_release`; distance.py:455-469 | 5 |
| 7 | Year-penalty branches beyond the smoke case's one shape (candidate with `original_year` absent → flat 1.0 branch; item year equalling the release year, not the original year) | distance.py:472-493 (Q2); mechanics study Q1.1 | 4 |
| 8 | Country component (code mismatch, present one side only) | distance.py:496-502 | 2 |
| 9 | Label/catalognum/albumdisambig components (present-both, one-sided) | distance.py:505-518 | 3 |
| 10 | `album_id` component and the wrong-ID-then-search-recovers path | distance.py:521-524; match.py:432-442 | 2 |
| 11 | `track_length` grace/max boundary (within grace, beyond grace under max, beyond max, `track_info.length` absent) | distance.py:364-373, 396-399 | 4 |
| 12 | `track_index` per-disc-vs-absolute tolerance (absolute matches, per-disc matches, mismatch, index absent) | match.py:376-380, 412-414 | 3 |
| 13 | `track_id` component (match, mismatch) | distance.py:416-418 | 2 |
| 14 | Candidate admission filters expressible without profile config (duplicate `(source, album_id)` in one search response, a zero-track candidate release) | match.py `_add_candidate` (Q2, "Candidate admission filters") | 2 |
| 15 | Recommendation/threshold boundary (single-candidate `low`, an explicit `max_rec` downgrade demonstration distinct from family 2's) | match.py:269-323, config_default.yaml:167-169 | 5 |
| | **Subtotal** | | **61** |

61 lands at the low end of the plan's 60-120 range with headroom to
double several families (string-distance edges and the boundary family
are the cheapest to extend — each additional case is one more small
synthetic release) if the pass wants more margin, or to compress by
merging adjacent families' cases (e.g. folding family 8-10's small
string-component families into a couple of "richly-tagged candidate"
cases that each carry country+label+catalognum+albumdisambig
simultaneously) if 61-90 proves sufficient once the implementer sees
how much a hand-authored release costs to write. `test_tracks_out_of_order`
and `TestTagMultipleDataSources` are excluded from this budget entirely
(Tier A / Tier C, per (a)).

## (c) The component-test derivation tool

`test_distance.py`'s `TestDistance`, `TestStringDistance`, and
`TestDataSourceDistance` classes (Tier A above) pin down, precisely:

- Every `Distance.add_*` method's exact output for a literal input
  table (l.36-89) — the primitive penalty-list builders every
  higher-level component call goes through.
- The raw/max/normalized arithmetic and per-key indexing (l.91-101,
  133-140 for `_weights`), including that a key added at `0.0` still
  inflates `max_distance`'s denominator (l.128-132) — the mechanics
  study's central Q2 finding, here as an executable proof rather than
  a source read.
- `string_dist`'s full pipeline in isolation: lowercase, unidecode,
  article rotation, `&`→`and`, and each `SD_PATTERNS` heuristic's
  strip-then-recompute-then-adopt-if-better logic (l.263-301),
  independent of any album or track context.
- The `data_source` penalty's full parameter space — match/mismatch,
  single vs. multiple sources, penalty and weight magnitude scaling
  (l.304-364) — which no fixed profile can express (see (e)) but
  which the component tool reaches trivially, because it sets
  `config["match"]["distance_weights"]` and monkeypatches
  `find_metadata_source_plugins` in-process exactly as beets' own test
  does, never touching a rendered YAML profile at all.

**How a derivation tool replays these through the shim's per-key
penalties:** the tool is a second, narrower Python entry point beside
`derive_oracle.py` — call it `derive_component.py` — living in the
same `internal/harness/oracle` package (same U7 boundary: a beets
bump edits only this package) and gated by the same `oracle` build
tag. Unlike `derive_oracle.py`, it never starts a stub server, never
calls `tag_album`, and takes no MusicBrainz response at all. It reads
a small table of literal argument tuples (mirroring
`test_distance.py`'s `@pytest.mark.parametrize` tables almost
verbatim), calls the real `Distance` method or `string_dist` or
`track_distance` directly for each row, and emits the resulting
penalty list, weighted per-key distance, or boolean exactly as the
beets test asserts it. Output is a small JSON table (row id, inputs,
observed output), checked in beside a Go test that table-drives an
assertion — this is not a `cassette.Case`/`Expected` record at all,
because there is no `Case.Requests`/response body/`Profile` for a
call that never touches the stub or a rendered config; forcing it
through the cassette schema would fabricate a fake request log for a
derivation that issues none. The tool's own small schema (row id,
method name, args, expected) is a 2b deliverable independent of
`cassette.go`.

## (d) What the tie detector and the proximity flag need

Confirmed against `internal/harness/cassette/cassette.go` (l.68-88)
and `internal/harness/oracle/derive_oracle.py` (l.168-190, 225-276) as
shipped:

**The Expected record does carry**, and both tools can lean on
directly: per-component weighted penalties per candidate
(`Candidate.Penalties`, matching `Distance.items()`'s "nonzero,
weighted, per-key" semantics — `derive_oracle.py`'s
`candidate_record` deliberately reports `_penalties`'s raw key set
instead, l.172-176, so a 0.0-valued-but-present key stays visible for
denominator reasoning); per-candidate overall `Distance`; the
candidates in beets' own final (distance-sorted, stable-tie) order;
the best candidate's item→track `Assignment`; `ExtraItems`/
`ExtraTracks` in beets' own sort order.

**The proximity flag** needs, beyond the record: the live
`strong_rec_thresh` (0.04), `medium_rec_thresh` (0.25), and
`rec_gap_thresh` (0.25) constants themselves (config_default.yaml:
167-169) — and here is a genuine gap. The shim already computes and
echoes these (`derive_oracle.py`'s `record["thresholds"]`, l.254-259,
alongside `record["weights"]`), but `oracle.Derive`'s
`json.Unmarshal(raw, &exp)` (oracle.go:114) decodes only into
`cassette.Expected`, whose fields do not include a `Thresholds` or
`Weights` member — Go's decoder silently drops both, and neither
`estate.yaml` nor `stock.yaml` overrides `match:` (confirmed: neither
profile file contains a `match:` block; both therefore run at
config_default.yaml's own values, verified above). The flag therefore
either (i) hardcodes the three defaults with a comment tying its
correctness to "neither profile touches `match:`, checked by
`ProfileHashes`" — fragile, since a future profile edit could silently
invalidate the assumption without the flag noticing, or (ii) has 2b
add two small `omitempty` fields to `cassette.Expected`
(`Thresholds map[string]float64`, `Weights map[string]float64`) fed
from the JSON the shim is *already* emitting, at zero shim-side cost.
**(ii) is the recommended fix**: it is additive (old records still
decode; the corpus's negative-fixture and round-trip tests are
unaffected), and it removes a cross-file assumption the corpus has no
mechanism to catch if it ever became false. With those fields present,
the flag is a pure Go computation: best `d = Candidates[0].Distance`,
gap `= Candidates[1].Distance - d` when `len(Candidates) >= 2`, flagged
whenever `d` is within ~1e-6 of `strong_rec_thresh` or
`medium_rec_thresh`, or the gap is within ~1e-6 of `rec_gap_thresh` —
the epsilon and the strict-`<`-vs-`<=` asymmetry (match.py:285 vs. 288)
are both already documented in the mechanics study's "epsilon regime"
section (finding 3) and need no new research.

**The tie detector** needs the item×track *assignment cost matrix* —
`track_distance(item, track, incl_artist=False)` for every (item,
track) pair, the exact objective `assign_items` minimizes (match.py:64)
— to determine whether the optimal assignment is unique. This is
**not** in the Expected record and cannot be recovered from it:
`Candidate.Tracks` is described precisely as "per-track distances,
assignment order" and `derive_oracle.py` (l.183) confirms it is
`dist.tracks[track]` for the matched pairs only, computed with
`incl_artist=album_info.va` — the *final scoring* objective
(distance.py:529), which the mechanics study's threat 2 already
establishes as a different number from the assignment objective on VA
albums. Neither objective's *off-diagonal* costs (the values for
item-track pairs `assign_items` did NOT choose) appear anywhere in the
committed schema, for any candidate, best or otherwise. Recommend
against a schema change here: unlike the threshold gap, the raw
inputs needed to reconstruct the full cost matrix — the case's
`Input.Items` and the committed `responses/*.json` the candidate's
`TrackInfo` list comes from — are already fully present in every
corpus case directory. The tie detector should be a Tier-A-style
side tool (living beside `derive_component.py`, sharing its "no stub,
no full `tag_album` call" shape) that re-parses a case's committed
response into `TrackInfo` objects, calls `track_distance` for every
(item, track) pair itself, and either perturbs a near-tied entry and
re-solves with `lap.lapjv` or checks for duplicate row/column minima,
flagging non-unique cases for adjudication rather than encoding tie
state into `Expected`. This keeps `cassette.Expected` exactly as pass
2a shipped it for the assignment-mapping purpose the plan already
claims it serves, while being honest that the *tie-detection* purpose
needs a tool, not a bigger record.

## (e) Shapes the file-free tier cannot express

1. **Multi-source ordering and `data_source`-penalty magnitude
   variants** (`TestTagMultipleDataSources`, most of
   `TestDataSourceDistance`'s params). Blocked by the single-metadata-
   source ruling itself, not by file-freeness; stays out of scope for
   both 2b and 2c per the mechanics study's Q1.2/Q1.5 and the plan's
   "the corpus should stay single-source" line. The one
   single-source-compatible sub-shape is already folded into family 5
   above.
2. **`preferred.original_year: yes`'s ratio branch**
   (distance.py:472-478, the `original = original_year or 1889`
   path). Both shipped profiles leave `preferred.original_year` at its
   default `no` (config_default.yaml:197; neither `estate.yaml` nor
   `stock.yaml` sets a `match:` block at all), so this branch is
   currently unreachable by *any* corpus case regardless of stratum.
   Not a 2b blocker on its own budget, but worth flagging to whoever
   owns 2c's "stock-config stratum per ruling 3": if that stratum's
   point is to probe config surfaces the estate profile doesn't touch,
   this specific key is the clearest concrete candidate, and it may
   need a third named profile (`preferred-original-year.yaml` or
   similar) rather than fitting inside stock's bare-defaults shape.
3. **`match.required`/`match.ignored`-driven candidate dropping**
   (match.py `_add_candidate`, the two filters beyond duplicate-ID and
   zero-track). Both keys default to `[]` in both profiles; exercising
   the drop rules needs a profile that sets them, which neither
   current profile does. Same shape as item 2 — a 2c/profile-variant
   question, not a 2b blocker, since family 14 already covers the
   two filters that need no config at all.
4. **The assignment cost matrix**, per (d) above — not a "cannot
   express" case in the sense of needing a config change, but a
   genuine schema gap: the tie detector's raw material is not
   persisted anywhere in `cassette.Expected` today and the
   recommended fix is a side tool, not a corpus case.
5. **File-level applied tags** (what `item.write()` puts in a FLAC via
   mediafile). Already named explicitly as pass 2c's deliverable and
   rung 8's stated precondition in the 2a plan (l.209-211); nothing in
   this inventory changes that boundary — every family above stays at
   the field-level `AppliedFields` tier `Expected` already supports.
6. **Further synthetic pseudo-release edge cases**
   (`TestPseudoRelease`'s `PSEUDO_WITHOUT_LINKS`/`PSEUDO_INVALID_LINK`
   shapes, which test the MB *plugin's* raw-JSON-to-`AlbumInfo`
   parsing — specifically which release supplies `album.country` after
   a merge — rather than the *distance/recommendation* math this
   inventory otherwise targets). 2a's live `non-latin-pseudo-release`
   smoke case already proves the capture-derive loop's double-fetch
   end-to-end against a real qualifying release. Minting additional
   synthetic pseudo-release variants would test MB-plugin parsing
   correctness, which is closer to 2c's real-catalog stratum (parsing
   edge cases against genuinely captured data) than to 2b's
   distance-math structural stratum; recommend 2b skip this family
   rather than duplicate scope 2c is better positioned to own.
7. **Wall-clock/library-version hazards** (mechanics study Q1.1, Q1.4).
   Not corpus-case shapes at all — already fully handled by
   `EvalDate`/`EchoedDate` pinning and `corpus/manifest.json`'s
   version quadruple, both shipped in 2a. Listed here only to confirm
   this inventory found no new gap in that handling.

## Sources

- `github.com/beetbox/beets`, tag `v2.13.1`, commit
  `b7952299941543d4507ac7931edb223acd684b3d`. Files cited:
  `test/autotag/test_distance.py`, `test/autotag/test_match.py`,
  `test/autotag/test_hooks.py`, `beets/autotag/distance.py`,
  `beets/autotag/match.py`, `beets/config_default.yaml`,
  `beets/ui/commands/import_/session.py`.
- This repo: `docs/plans/2026-09-02-rung-2a-cassette-substrate.md`,
  `docs/research/2026-09-01-beets-conformance-corpus.md`,
  `docs/research/2026-09-02-beets-matcher-mechanics.md`,
  `internal/harness/cassette/cassette.go`,
  `internal/harness/oracle/oracle.go`,
  `internal/harness/oracle/derive_oracle.py`,
  `internal/harness/oracle/testdata/cases/smoke/`, `corpus/README.md`,
  `corpus/profiles/estate.yaml`, `corpus/profiles/stock.yaml`.
