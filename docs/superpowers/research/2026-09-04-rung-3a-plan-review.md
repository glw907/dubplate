# Adversarial review: the rung 3 pass 3a plan

Fresh-context review (2026-09-04) of plans/2026-09-04-rung-3a-scoring-core.md
revision 1, arithmetic verified contract-by-contract against pinned beets
2.13.1 in tools/venv/.venv. Folded as revision 2 the same day; the
22-item "Verified sound" list below is the do-not-re-litigate record
for the pass, alongside the spec review of the same date.

---

I've completed the review. Verified every arithmetic claim against pinned beets in `tools/venv/.venv`, every tree claim against `9a9b0e0`, and re-derived the corpus figures directly.

---

# Adversarial review: `docs/superpowers/plans/2026-09-04-rung-3a-scoring-core.md` (rev 1)

## BLOCKERS

**B1. Task 6's decoder enumeration mis-keys the seven `aggregate/*` rows: their `method` is `add`, not `aggregate`. A per-method decoder cannot decode the committed table.**

`plan:948-975` says "The decoder is typed, **per method**", lists "the `aggregate/*` and `generic_penalty_keys/*` rows" as their own bucket, and requires "A row whose method the decoder does not know is a **failure naming the row id**, never a skip." But `internal/harness/oracle/derive_component.py:307-320` emits `yield row(row_id, "add", …)` for every aggregate row. The committed method counts are:

```
add 8 (= add/basic + 7 aggregate/*)   add_data_source 4   add_equality 3   add_expr 2
add_number 4   add_priority 4   add_ratio 4   add_string 3
generic_penalty_keys 1   string_dist 23   track_distance 16   = 72
```

Method `add` therefore carries two incompatible shapes: `{"calls":[[1.0]],"key":"add"} → {"penalties":[…]}` and `{"calls":[[key,dist],…],"weights":{…}} → {distance, raw_distance, max_distance, per_key, items}`. An opus implementer building exactly what the plan says gets a shape mismatch on 7 of 72 rows.

*Fix:* state the dispatch key explicitly — dispatch on `method`, and inside `add` on the presence of `args.key` (adder shape) versus `args.weights` (aggregate shape). Say "eleven methods, twelve shapes", and keep the unknown-shape failure.

**B2. The `weights` field is an OVERRIDE layered on beets' twenty shipped defaults, not a complete weight table. The plan never says so, and `aggregate/normalized` cannot pass without it.**

`derive_component.py:169-190` does `Distance._weights.update(overrides)` and restores afterward. Two committed rows prove it decisively:

- `aggregate/normalized` supplies `{album:4, data_source:2, medium:2}` but adds a **`media`** penalty. Its `max_distance` of 6 = 1×4 + 2×1 requires `media`'s weight to be beets' shipped default 1.0.
- `track_distance/data_source/mismatch-high-weight` supplies `{data_source:5}`; its `distance` 0.3125 = 2.5/8 requires `track_title` = 3.0 from the defaults.

A decoder that uses the row's map verbatim produces 3/4 = 0.75 for the first and fails with no named cause.

*Fix:* add to Task 6: "the row's `weights` map overrides `Tunables.DistanceWeights` key by key; absent keys keep their beets default; `aggregate/all-zero-weights` introduces `zero_weighted`, a key beets does not ship, which exists only for that row."

**B3. Task 6 must model `metadata_plugins.get_penalty` and the two-plugin stub, and the plan does not say how — sixteen rows are unreplayable without it.**

The four `add_data_source/*` and twelve `track_distance/data_source/*` rows carry `data_source_mismatch_penalty` and `multiple_data_sources`. beets resolves via `get_penalty(after)` (`metadata_plugins.py:160-169`): the derivation loads `OriginalPlugin` (class default 0.5) and `OtherPlugin` (the row's penalty), or `OtherPlugin` alone when `multiple_data_sources` is false (`derive_component.py:192-224`). Any other `after` — `"unknown"`, `null` — falls back to `DEFAULT_DATA_SOURCE_MISMATCH_PENALTY` = 0.5. `track_distance/data_source/mismatch-unknown` reads `0.5` for exactly that reason.

*Fix:* state it as a decoder contract in Task 6, wired through the "loaded metadata source set" Task 1 already puts on `Tunables`.

**B4. `tools/` is lane E1's declared territory, and Task 3 writes into it with no ownership row.**

`parallel-build-approach.md:79` gives lane E1 `corpus/`, `internal/harness`, `cmd/harness`, **`tools/`**. `plan:282` claims `tools/unidecode-table/` for 3a (lane E2). The four contested-surface rows the rung 3 spec proposes (`2026-09-03-rung-3-go-matcher.md:336-342`) cover conform, oracle, `{cassette,stub,analysis}`, and `corpus/analysis` + `corpus/cases` — none covers `tools/`. Lane-launch checklist item 2 requires a written owner for every surface a lane touches, and Task 0 item 4 (`plan:497-502`) enumerates only the four.

*Fix:* add `tools/` as a fifth row to Task 0 item 4's blocker list, or move the generator under a path 3a already owns.

## MAJOR

**M-a. The plan's stated safety property for `SD_PATTERNS` is false: tier 0 cannot catch a "fixed" pattern table.**

`plan:253-256`: "Tier 0 is graded in oracle-faithful mode by construction, and a 3a that 'fixed' `SD_PATTERNS` would fail tier 0 on its own pass." Measured against pinned beets: applying the word-boundary guard named in `adj-adv-ft-draft-colon` to the featuring pattern changes **zero of the 72 rows**. `string_dist/relative/3/near` is `("My Song feat Someone", "My Song")`, where "feat" is a standalone word and the guard is satisfied — 0.06470588235294118 under both tables.

`plan:250-252`'s explanation is also wrong: the pattern fires *correctly* there, on the word "feat". The adjudicated defect is `ft` inside "Draft" (`adjudications.json`, `adj-adv-ft-draft-colon`).

*Fix:* delete the "would fail tier 0" claim and replace it with an explicit guard in Task 4 — a test asserting `BeetsSDPatterns` is byte-identical to beets' six literals, plus a discriminating fixture. Measured today: `string_dist("Draft: Ember Wharf", "Draft: Northern Lines")` = **0.061111111111111116** unguarded, **0.6111111111111112** guarded. Carry both figures in the pass report for 3b's `defect_free` derivation.

**M-b. Task 2's firing mutations are specified against `components.json` rows the task is forbidden to read.** `plan:692-697` names "must fail `aggregate/zero-added-key`", "`aggregate/all-zero-weights`'s empty `items`", "`aggregate/sorted-penalties`"; `plan:699-701` says "This task does not read `components.json`." *Fix:* restate the four mutations against Task 2's hand-authored equivalents; move the row-name assertions into Task 6's mutation list.

**M-c. Task 3 puts a corpus-reading manifest-pin test inside `internal/engine/matcher`.** `plan:764-768`. Every other engine task is explicitly barred from reading the corpus (`plan:699, 840, 909, 1137`), and reaching `cassette.Manifest` for the decode trips the arch check Task 6 lands (`plan:1000-1004`). *Fix:* export the table's version stamp from `matcher` and land the manifest comparison in `internal/harness/conform` at Task 6.

**M-d. `ItemFields` is named but never enumerated, and its test also reads a committed case.** `plan:606-620` gives explicit field lists for `AlbumInfo` and `TrackInfo` and none for `ItemFields`. The distance path reads at least eighteen item fields: the thirteen `get_most_common_tags` fields (`beets/util/__init__.py:822-836` — artist, album, albumartist, year, disctotal, mb_albumid, label, barcode, catalognum, country, media, albumdisambig, data_source) plus `length`, `title`, `artist`, `track`, `mb_trackid`, `disc`, `data_source` from `track_distance` (`distance.py:394-424`). Separately, `plan:619` ("matches what a committed case's `Input.Items` map produces") is the same corpus-read-from-engine conflict as M-c. *Fix:* enumerate the eighteen with their beets types; make the projection assertion hand-authored and carry the corpus cross-check to Task 6.

**M-e. Contract 6 omits scalar-to-list wrapping on `AddEquality`/`AddPriority`, and the regex-option JSON encoding is unstated.** `distance.py:287-288` and `:328-329` wrap a non-list option — which is how `distance()` itself calls `add_equality("media", …, likelies["media"])` and `add_equality("album_id", …)` (`:465, :522-524`). Tier 0 exercises both forms: `add_priority/2`'s first call passes a scalar, its second a list. And a compiled-regex option is serialized as `{"regex": "<pat>", "ignorecase": <bool>}` (`derive_component.py:35-51`), which `plan:953-975` never mentions.

**M-f. Contract 7's anchoring rule will mislead an implementer into anchoring `SD_PATTERNS`.** `plan:427-433` says "The Go port anchors explicitly (`^(?:…)`), never `MatchString` on a bare pattern." That belongs only to `Distance._eq`'s options (`distance.py:265-267`). `SD_PATTERNS` are used with `re.sub` — unanchored, global find-and-delete (`distance.py:99-100`) — and anchoring them destroys the pattern loop. *Fix:* split the contract into "regex **options** anchor" and "`SD_PATTERNS` never anchor; they are `ReplaceAll` over the whole string", and restate it in Task 4.

**M-g. Task 6 requires reading `componentMethods`, which is unexported and behind the territory wall.** `internal/harness/analysis/component.go:22-35` declares `var componentMethods` lowercase; `plan:283` forbids writing `internal/harness/analysis`. The "one derived fact, one home" argument (`plan:976-978`) works for `ComponentRowFloor` and cannot work here. *Fix:* say conform owns its own decoder-method set (it must, to be typed), that `analysis.CheckComponentTable` already enforces table-side coverage, and that conform's obligation is "every method my decoder knows has ≥1 replayed row, and every row's method is known".

**M-h. The lane-T-beside-3a framing contradicts the plan's own lane-registry reading.** `plan:39-47` presents "lane T beside 3a" as the allocation and a constraint on this plan; `plan:172-174` records lane-2 BLOCKED until 0a/0b/0c land, and Task 0 item 1 expects 0b/0c unlanded. `reverse-edge-completion.md:317-320` makes all three a precondition. As committed, no lane T can run beside 3a. *Fix:* state the pairing as conditional, and say the 12M phase-1 figure is a reservation, not a live sum.

**M-i. Album component count is inconsistent.** Contract 8 (`plan:440-442`) enumerates fourteen album-order components; `plan:1036-1041` ships thirteen files and `plan:1100-1101` calls them "the thirteen album components". `missing_tracks` and `unmatched_tracks` share `album_extras.go`, which collides with the task's own "One component, one file, one test" rule (`plan:1054`). *Fix:* fourteen components in thirteen files, with `album_extras.go` named as the deliberate exception.

## MINOR

- **m1.** `plan:14-15` "the worktree early-start is **dead** by the conductor's fold ruling" overstates it; the ruling (rung-3 spec:49-50) says it "remains an explicitly costed option nobody currently needs", which `plan:271-272` gets right.
- **m2.** `plan:560-561` cites `metadata_plugins.py:191-195` for `DataSourceMismatchPenalty` 0.5; that range is the `source_weight`-first property. The default lives at `:180` and is wired at `:202`.
- **m3.** `plan:559`'s "legacy `match.searchlimit` spelling" is not beets' top-level `match:`. It is the metadata-source plugin's own `match.searchlimit` subkey; the deprecation message names it `musicbrainz.searchlimit` (`beetsplug/musicbrainz.py:286-295`).
- **m4.** `plan:238-239, 555`: `distance_weights` is twenty entries at `config_default.yaml:174-193`; `:173` is the header. Every other Task 1 citation verified exactly (`:167-169, :170-172, :194-197, :198-200, :201-202, :203-204, :27, :28, :32, :33, :34, :58`).
- **m5.** `plan:200-202`: `AlbumMatch.apply_album_metadata` (`match.py:199-201`) is a *sibling* of `apply_metadata`, not inside it, and there is a third consumer, `TrackMatch.apply_metadata` (`:239`). All post-verdict, so the M2 conclusion stands. Note in the drafted correction that the README sentence quoted at `plan:208-212` is itself imprecise — `per_disc_numbering` is read in `TrackInfo.raw_data` (`hooks.py:421-452`), not `AlbumInfo.item_data`; the plan's replacement text already fixes this and should say so.
- **m6.** Contract 10's "typed default" excludes `data_source`: it is in `Item._types`, not `Item._fields` (`library/models.py:66-70`), so an unset read returns `None` (`dbcore/db.py:452-468`). Outcomes are identical through `add_data_source` and `plurality`; the Task 1 test bullet should not use it as its example. Contract 10 should also note `get_most_common_tags` filters falsy items (`util/__init__.py:838`).
- **m7.** `plan:156-158`: nine tasks against 2d's four is ~2.25×, not "three times"; 3 × 1.6M = 4.8M, not 6M. The 6M figure is the spec's own proposal (rung-3 spec:360) and is fine; the justifying sentence is not.
- **m8.** `plan:846-847` "the two solo cases" — there are three `string_dist/solo/*` rows; the third is named a line earlier as "the `(EP)` zero-weight pattern".
- **m9.** `Tunables` is ~26 fields as enumerated, not "thirty-odd" (`plan:624`). `ExtraTags` (`plan:562`) is the one default with no citation; it lives on the musicbrainz plugin, not `config_default.yaml`.
- **m10.** `match:` carries two keys `Tunables` deliberately omits — `album_disambig_fields` / `singleton_disambig_fields` (`config_default.yaml:205-206`), display-only via `Match.disambig_fields` (`match.py:113-142`). Say so, or a Task 0/1 reader reads the "verified-complete" claim as falsified.
- **m11.** `.gitleaksignore` is on neither side of the territory wall (`plan:280-284`) yet Task 3 permits an append (`plan:774-777`). Name it as conductor-owned and conditionally permitted, repeating the append-only-below-line-66 rule.
- **m12.** `repoRoot` is test-only and unexported (`internal/harness/corpus_test.go:19-36`). Primary-input 5 says conform "needs" it; it needs the *shape* and must reimplement.
- **m13.** `Distance.tracks`' Go key type is unstated. Python keys by `TrackInfo` identity (`distance.py:131, 527-530`); Go needs `map[*TrackInfo]*Distance` or an index. Both Task 2 (`plan:648-650`) and Task 7 (`plan:1085-1086`) assume it exists.
- **m14.** `Tunables.MaxRec`'s Go type is unstated and must compare against `cassette.Thresholds.MaxRec map[string]string` (`cassette.go:88-93`) for Task 6's cross-check.
- **m15.** Conductor note, outside the plan: `docs/STATUS.md:14` is stale ("one `oracle-correct`, five pending") against `oracle_bugs.json`'s 4 `oracle-wrong` / 2 `oracle-correct`. Task 0 item 7 will surface it; STATUS is read in full every session.

---

## VERIFIED SOUND — do not re-litigate

Checked directly against pinned beets 2.13.1 and the tree at `9a9b0e0`.

1. **Contract 1, the `string_dist` pipeline.** lower → transliterate → lower is exact (`distance.py:74-75, 55-58`). `SD_END_WORDS = ["the","a","an"]`, trailing `", <word>"` rotates to the front, applied per string and per word in order (`:80-84`). `&`→`and` (`:44, 87-89`). Both `None`→0.0, one `None`→1.0 (`:69-72`). Order `base_dist + penalty` (`:119`).
2. **Contract 5, the normalizer.** `raw = Σ sum(p[k])·w[k]`, `max = Σ len(p[k])·w[k]`, `distance = raw/max` with 0.0 on zero max, per-key `sum(p[k])·w[k]/max` with the same guard, nonzero view sorted `(-weighted, key)` (`:157-176, 190-205, 232-238`).
3. **Contract 6, all eight adders.** Every semantic checks out verbatim, including `Add`'s `ValueError` outside `[0,1]`, `AddPriority`'s `i × 1/(len(options) or 1)`, `AddRatio`'s `float(max(min(n1,n2),0))/n2` with 0.0 on falsy `n2`, `AddNumber`'s per-unit 1.0 penalties, and `AddDataSource`'s three-way condition adding *nothing* on a match (`:269-361`). `add_number/3`'s six-penalty output and `add_ratio/3`'s `(5,0)→0.0` and `(10,5)→1.0` confirm it from the corpus side.
4. **Contract 3, the threshold asymmetry.** `min_dist < strong_rec_thresh` strict, `elif min_dist <= medium_rec_thresh` (`match.py:284-288`).
5. **Contract 4, the two views.** `derive_oracle.py:176` is `{key: dist[key] for key in dist._penalties}` verbatim; `keys()` derives from `items()` and yields nonzero only (`distance.py:190-247`); `_recommendation`'s downgrade reads `set(min_dist.keys())` (`match.py:306`). **272 of 294 committed candidates carry a 0.0-valued key** — recomputed off the tree.
6. **Contract 8, both component orders.** Track: `track_length, track_title, track_artist, track_index, track_id, medium, data_source` (`distance.py:394-426`). Album: `artist` (skipped on `va`), `album, media, mediums, year, country, label, catalognum, albumdisambig, album_id, tracks, missing_tracks, unmatched_tracks, data_source` (`:444-542`).
7. **Contract 10 and 11.** The thirteen `get_most_common_tags` fields in the plan's exact order; `plurality` is `Counter.most_common(1)[0]`, whose `nlargest` is stable, so ties break by first-seen insertion order = item order; the albumartist-consensus override of `likelies["artist"]` (`util/__init__.py:799-846`). `Model._get` returns `self._type(key).null` for declared-but-unset fixed fields (`dbcore/db.py:452-468`), so `""`/`0`, not nil — subject to m6.
8. **Task 5's seven track components**, every weight and every fire condition, including `track_length`'s grace/max clamping through `add_ratio`, the `VA_ARTISTS` tuple `("", "various artists", "various", "va", "unknown")`, `track_index_changed`'s per-disc/absolute tolerance, and `inclArtist` as a parameter.
9. **Task 7's album components**, every weight and branch, including the three year branches with `original_year or 1889`, the `(\d+x)?(<pat>)` case-insensitive media compile, the bare case-insensitive country compile, and `dist.add("tracks", …distance)` per matched pair with `inclArtist = album_info.va`.
10. **The M2 resolution.** `original_date` is read once, in `AlbumInfo.item_data` (`hooks.py:313`); `per_disc_numbering` in `TrackInfo.raw_data` (`:433, :437`), `library/models.py:515`, and `ui/commands/import_/display.py:153` — nowhere else. Neither appears in `distance.py` or on the distance path; `track_index_changed` reads attributes, not `raw_data`. **Both are application-time only. `Tunables` correctly carries neither.** `corpus/README.md:2428-2429` does carry the two "verdict-moving" cells the correction targets.
11. **The m2 resolution.** `config_default.yaml` distance_weights is exactly twenty entries; all 215 committed cases echo one weight set of those same twenty keys and one threshold set (`0.04 / 0.25 / 0.25`, `max_rec {missing_tracks: medium, unmatched_tracks: medium}`) — recomputed off the tree. **Q4's "19 keys" is superseded.**
12. **jellyfish's `levenshtein_distance` is plain unit-cost edit distance.** `kitten/sitting`=3, `ab/ba`=2 (no transposition), empty-string cases = length. `plan:769-770` is correct.
13. **Every `unidecode` 1.4.0 figure is exact.** 190 block files, 48,597 entries, 41,379 non-empty, 5,477 `None`, 157,450 bytes of mapped output — recomputed. The `None`-versus-empty distinction is real and material (1,741 empty-string entries). The `>0xeffff` cutoff, the `<0x80` passthrough, and the surrogate warning are all at `unidecode/__init__.py:81-115`.
14. **Task 4's pattern-loop contract is exact**, including adopt-only-on-positive-delta (`distance.py:97-117`) and `re.sub`'s all-non-overlapping deletion. The `relative/3/near` arithmetic reproduces exactly: base 11/17 = 0.6470588235294118, delta adopted at weight 0.1 → 0.06470588235294118.
15. **The harness APIs the plan calls all exist with the named signatures.** `analysis.ComponentRowFloor = 72` (exact — `components.json` carries exactly 72), `analysis.LoadComponentTable(path string) (*ComponentTable, error)`, `ComponentTable`, `ComponentRow` with `Args`/`Output` as `json.RawMessage`, `cassette.LoadCase`, `cassette.Expected` (`Weights map[string]float64`, `Thresholds`), `cassette.Thresholds`, `cassette.Candidate.Penalties map[string]float64`, `cassette.Manifest`.
16. **The M1 cross-check has a real, non-empty overlap today.** All three profiles (`estate.yaml`, `stock.yaml`, `locale.yaml`) set **no** `match:` keys; every committed record echoes beets' defaults. The Task 6 defaults cross-check will fire on any transcription error.
17. **The tree facts.** `internal/engine` does not exist; `go.mod` carries cobra plus two indirects only; `ownedWall` is still a slice literal in `internal/config/keys.go:19-25`; `.gitleaksignore` holds exactly one line, `internal/config/keys_test.go:generic-api-key:66`; 215 cases; three profiles; `corpus/manifest.json` carries six dependency pins.
18. **`scripts/check.sh` gates the new files as claimed.** `git ls-files '*.py'` (`:85`) so `tools/unidecode-table/generate.py` is ruff-gated with no `check.sh` edit; `vale-comments.sh:54-55` unions tracked and untracked `*.go` excluding `_salvage/`, so `internal/engine` prose is gated. **No territory violation here** — the plan's claim is true.
19. **The adjudication snapshot is exact as committed.** `oracle_bugs.json`: `bug-ft-pattern`, `bug-remix-parenthetical`, `bug-nonlatin-alias-selection`, `bug-assignment-rejection-gap` at `oracle-wrong`; `bug-threshold-boundary`, `bug-locale-alias-ignored` at `oracle-correct`; all `adjudicator: "Geoff"`. `adjudications.json`: 21 entries, 18/3, zero pending.
20. **The split-point reasoning is correct.** `components.json` carries no album-distance rows at all; tier 0 grades all 72 rows using only Tasks 1-6, so the Task 6/7 boundary is a genuinely clean cut. Task-count and deliverable-count discipline is followed (10 tasks, counts stated per task, split proposed rather than scope packed), matching the 2d plan's house shape.
21. **Lane conformance is otherwise correct.** Main tree matches `docs/STATUS.md`'s lane-1 entry; the plan correctly reads lane-2 as BLOCKED; Task 0 item 4 correctly identifies the four contested-surface rows as **not yet written** into `parallel-build-approach.md:156-166` — I confirmed the table carries no harness or corpus row today, so that blocker check is live and load-bearing.
22. **The spec-review "Verified sound" list is exactly twenty items**, and the plan's summary of them at `plan:20-26` is faithful.

---

## Verdict

**Needs revision — a tight, bounded one, then fold-and-proceed.** The arithmetic half of this plan is exceptionally good: I checked all twelve stated contracts, the eight adder semantics, both component orders, all three year branches, every track fire condition, all twenty weights, and every `config_default.yaml` line citation against pinned beets, and found the substance correct essentially everywhere, with only citation-level drift (m2–m5). The five unidecode measurements are exact to the digit, the corpus figures (215 cases, 20 weight keys, 272 zero-key candidates, 72 rows, 23/16 row counts, the 4+2 and 18/3 dispositions) all recompute, and the M2 and m2 resolutions are both correctly decided against the venv. The failure is concentrated in one place and one direction: **Task 6's decoder contract was written from a reading of the row *ids* rather than the row *methods* and the derivation shim**, which produces four blockers (aggregate rows keyed `add`; `weights` as an override rather than a table; the `get_penalty` stub model; the regex-option encoding) that between them make Task 6 unexecutable as written and would burn an opus implementer's budget rediscovering `derive_component.py`. Beside that sit one false safety claim (M-a — the word-boundary guard is provably invisible to tier 0, so the plan's stated defence against a well-meaning `SD_PATTERNS` "fix" does not exist), one uncovered territory row (`tools/` is lane E1's, B4), and two engine-reads-the-corpus leaks the plan's own discipline forbids everywhere else (M-c, M-d). None of these touch the ratified rulings or the do-not-re-litigate record; all are fixable inside the plan document without re-opening the spec. Fold B1–B4 and M-a through M-i into a revision 2, take the minors as editorial, and dispatch Task 0 unchanged — its checklist already catches the two blockers that live outside this document.