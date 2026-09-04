# Adjudication batch 1 (pass 2c close)

Prepared 2026-09-03 for Geoff's ruling. Every row below is read
straight off the committed record and the committed
`corpus/analysis/adjudications.json` / `corpus/analysis/oracle_bugs.json`
entries; nothing here was re-derived to produce this document.

**Nothing in this batch has been self-adjudicated.** Every row's
`disposition` in `adjudications.json` sits at `pending-adjudication`,
`adjudicator: "agent"`. An agent-authored row can never carry any
other disposition (`internal/harness/analysis`'s gate-resident checker
enforces this and ships a fixture proving it). Every expected verdict
below—whether a case's own suspected defect is real, whether its
recommended verdict should move, whether a hypothesis should read
`oracle-wrong`, `oracle-correct`, or `no-correct-answer`—is Geoff's
to set, not this document's or the pass's.

**This batch does not block the pass close.** The pass's own gate
(`bash scripts/check.sh`) accepts a `pending-adjudication` row
indefinitely; only the **launch gate**, a different gate at a later
time, refuses an unadjudicated divergence. Pass 2c closes with all 19
rows below still pending, by design.

**One hypothesis is already ruled and is not repeated below**:
`bug-threshold-boundary` (`#5458`, threshold-boundary strictness),
adjudicated `oracle-correct` by Geoff in-session on 2026-09-03 (the
probe case `rec-strong-medium-boundary`, pass 2b, and `proximity.json`
already showed expected behavior at pinned beets 2.13.1; no
`adjudications.json` row exists for it, since it carries no
divergence to adjudicate).

Nineteen rows across five hypotheses, grouped so one ruling can cover
a class.

---

## Hypothesis 1: the featuring-pattern (`"ft"`) false positive

`bug-ft-pattern`, citation `742ebdf69` (`SD_PATTERNS`' featuring
discount added with no word boundary, Sept 2010), `7b59604c5` (partial
fix, beets 2.13.0), `19d091440` (symmetric word-boundary fix, still
unreleased at pinned 2.13.1). `reproduced: "yes"` (agent-set, a
machine-observable fact about whether the probe cases moved the
verdict).

**Suspected defect.** `SD_PATTERNS`' featuring discount,
`[\[\(]?(featuring|feat|ft)[\. :].+`, carries no word boundary before
its trigger, so it fires on the letters "ft" inside an ordinary
English word ("Draft", "Left", "Gift", "Soft", "Craft", "Theft")
followed by a space, period, or colon, and deletes everything from
that point to the end of the string as though it were a featuring
credit. The discarded difference is charged at the pattern's 0.1
weight instead of in full.

**Release truth (all six).** Every item set describes a genuinely
different release from its candidate; release truth is known by
construction, since every release is invented for this probe.

| case | derived (record) | defect-free (counterfactual) | suspected wrong | citation |
| --- | --- | --- | --- | --- |
| `adv-ft-draft-colon` | distance `0.010185`, `album: 0.010185` (sole nonzero), **strong, apply** | distance `0.101852`, past `strong_rec_thresh` (0.04), **medium** | "Draft: Ember Wharf" vs "Draft: Northern Lines" collapse to "dra" (colon trigger, both sides); `string_dist` reads 0.061111 against a true 0.611111 | `742ebdf69`/`7b59604c5`/`19d091440` |
| `adv-ft-left-colon` | distance `0.011765`, `album: 0.011765` (sole nonzero), **strong, apply** | distance `0.117647`, **medium** | "Left: A Distant Shore" vs "Left: The Harbor Road" collapse to "le" (colon trigger, both sides, widest gap in the family); `string_dist` reads 0.070588 against a true 0.705882 | same |
| `adv-ft-gift-space` | distance `0.010784`, `album: 0.010784` (sole nonzero), **strong, apply** | distance `0.107843`, **medium** | "Gift Wrapped Nights" vs "Gift Horse Blues" collapse to "gi" (plain-space trigger, no punctuation needed); `string_dist` reads 0.064706 against a true 0.647059 | same |
| `adv-ft-soft-onesided` | distance `0.030095`, `album: 0.030095` (sole nonzero), **strong, apply** | distance `0.086667`, **medium** | one-sided firing: item "Harbor Lights Soft. Variation" carries the trigger, candidate "Harbor Lights" does not; the item's whole trailing edition string is deleted; `string_dist` reads 0.180571 against a true 0.520000, the family's largest surviving penalty | same |
| `adv-ft-craft-titles` | distance `0.009171`, `tracks: 0.009171` (sole nonzero), **strong, apply** | distance `0.091708`, **medium** | per-track path: all three items' "Craft: ..." titles discount against three candidate "Craft: ..." tracks (0.050000/0.056250/0.058824 against true 0.500000/0.562500/0.588235); album matches exactly on both sides | same |
| `adv-ft-theft-artist` | distance `0.009091`, `artist: 0.009091` (sole nonzero), **strong, apply** | distance `0.090909`, **medium** | album-level `artist` component: items credited "Theft Marine" against candidate "Theft Signal" collapse to "the"; `string_dist` reads 0.054545 against a true 0.545455 | same |

**What ruling this class covers.** Whether the featuring-pattern's
missing word boundary is a real defect that should move all six cases'
recommended verdict from strong/apply to medium (review), across the
album, per-track-title, and album-artist components alike.

---

## Hypothesis 2: the parenthetical and remix near-duplicate discount

`bug-remix-parenthetical`, citation `#5868`. `reproduced: "yes"`.

**Suspected defect.** `SD_PATTERNS` discounts a parenthetical
(`\(.*?\)`) and a bracketed span (`\[.*?\]`) at weight 0.3. That is
the right call when the parenthetical is noise, and the wrong one
when it is the only thing distinguishing a remix, rework, alternate
version, or original mix—exactly the shape a real rip's tag
carries. The discount erases the differentiator and the matcher
confidently retags one release as the other.

| case | derived (record) | defect-free (counterfactual) | suspected wrong | citation |
| --- | --- | --- | --- | --- |
| `adv-remix-both-sides` | distance `0.024000`, `album: 0.024000` (sole nonzero), **strong, apply** | distance `0.080000`, **medium** | `#5868`'s literal shape: "Glass Harbor (Nocturne Rework)" vs "Glass Harbor (Original Mix)" collapse to "glass harbor"; `string_dist` reads 0.144000 against a true 0.480000; `applied_fields` show the remix rip retagged as the original mix | `#5868` |
| `adv-remix-onesided` | distance `0.021429`, `album: 0.021429` (sole nonzero), **strong, apply** | distance `0.071429`, **medium** | ordinary remix-vs-original shape: "Salt Meridian (Kite Remix)" vs "Salt Meridian", pattern firing on the item side alone; `string_dist` reads 0.128571 against a true 0.428571 | `#5868` |
| `adv-remix-bracket` | distance `0.027778`, `album: 0.027778` (sole nonzero), **strong, apply** | distance `0.092593`, **medium** | the square-bracket twin: "Copper Ledger [Nocturne Version]" vs "Copper Ledger"; `string_dist` reads 0.166667 against a true 0.555556; square brackets are a conventional version marker in real rip metadata | `#5868` |
| `adv-remix-track-titles` | distance `0.024561`, `tracks: 0.024561` (sole nonzero), **strong, apply** | distance `0.081871`, **medium** | per-track twin, closest to a real remix rip: every item title "\<original\> (Kite Remix)" against the candidate's own originals (0.150000/0.142105/0.150000 against true 0.500000/0.473684/0.500000); album matches on both sides | `#5868` |

**What ruling this class covers.** Whether the 0.3-weight
parenthetical/bracket discount is a real defect that should move all
four cases' recommended verdict from strong/apply to medium (review),
across the album and per-track-title components and both bracket
styles.

---

## Hypothesis 3: per-pair track-assignment rejection gaps

`bug-assignment-rejection-gap`, citation `#2981` (open since 2018,
same track count with divergent titles yields a confidently wrong
per-track assignment), `#6078` (2025, tracks detected in wrong
order). `reproduced: "yes"`.

**Suspected defect.** `assign_items` (match.py:53-81) builds a full
item-to-track cost matrix from `track_distance` and hands it to
`lap.lapjv`, which always returns a complete minimum-cost mapping: it
never rejects a pair, however costly, and never leaves an item and a
track both unmatched once the counts agree. Every probe case gives
its item set the release's own track count with an identical title on
every item (proven, not merely asserted by construction, to
contribute zero discriminating signal—`string_dist` scores it
identically for every candidate track), leaving track length as the
only signal; each case tunes a pair of item lengths to sit within
`track_length_grace` (10s) of the *wrong* track and outside it of
their own true one. **The sharpest finding in this class: the
album-level verdict never signals anything is wrong at all**, even
while individual per-track titles are provably swapped against
release truth known by construction.

| case | wrong assignment (derived) | correct assignment (would-be) | swapped tracks | citation |
| --- | --- | --- | --- | --- |
| `adv-assign-swap-session-tape` | `{"0":1,"1":0,"2":2}`, distance `0.016667`, `tracks: 0.016667` (sole nonzero), **strong, apply** | `{"0":0,"1":1,"2":2}` would cost distance `0.052222`, **medium** | items 0 and 1 (adjacent) | `#2981`/`#6078` |
| `adv-assign-swap-field-recording` | `{"0":0,"1":1,"2":3,"3":2}`, distance `0.016000`, **strong, apply** | identity would cost distance `0.048000`, **medium** | items 2 and 3 (adjacent) | same |
| `adv-assign-cycle-rehearsal-reel` | `{"0":1,"1":2,"2":0,"3":3,"4":4}`, distance `0.019481`, **strong, apply** | identity would cost distance `0.070390`, **medium** (widest correct-vs-wrong gap of the six) | items 0, 1, 2 (a genuine 3-cycle) | same |
| `adv-assign-double-swap-broadcast` | `{"0":1,"1":0,"2":2,"3":4,"4":3,"5":5}`, distance `0.021429`, **strong, apply** | identity would cost distance `0.032540`, still **strong** by a narrow margin | items 0/1 and items 3/4 (two adjacent pairs) | same |
| `adv-assign-endpoint-swap-demo-pass` | `{"0":3,"1":1,"2":2,"3":0}`, distance `0.026667`, **strong, apply** | identity would cost distance `0.106667`, **medium** | items 0 and 3 (the release's endpoints, proving the defect keys on length proximity, not index proximity) | same |
| `adv-assign-tight-margin-archive-disc` | `{"0":1,"1":0,"2":2}`, distance `0.016667`, **strong, apply** | identity would cost distance `0.022593`, still **strong** (narrowest margin of the six) | items 0 and 1 (adjacent) | same |

**Release truth (all six).** The correct assignment is always the
identity mapping (item *n* is track *n*'s own rip); known by
construction. `applied_fields.title` on every swapped item shows the
wrong candidate track's title applied. Every case's uniqueness is
confirmed by `ties.json` (`TieMargin` 1e-9): the wrong mapping is the
solver's actual, sole optimum, not a near-tie artifact.

**What ruling this class covers.** Whether `assign_items`/`lap.lapjv`
never rejecting a costly pair is a real defect worth flagging (since
no case here crosses a recommendation threshold—the class's own
finding is that the album-level verdict is silent regardless), and
whether the per-track `applied_fields.title` values should be treated
as wrong against release truth known by construction.

---

## Hypothesis 4: non-Latin, pseudo-release, and alias mis-selection

`bug-nonlatin-alias-selection`, citation `#1916` (romaji/locale
transliteration, no preference mechanism, open since 2016), `#6358`
(locale aliases ignored after `mbsync`, 2026). `reproduced: "yes"`,
against two of Task 4's already-committed `real-nonlatin-*` cases (two
others, `real-nonlatin-alla-pugacheva-vol1` and
`real-nonlatin-x-japan-art-of-life`, show no divergence and are not
counted toward this reproduction).

**Suspected defect.** `_parse_artist_credits`
(`beetsplug/musicbrainz.py:301-331`) always reads
`el['artist']['name']`, the MusicBrainz *artist entity's* own
canonical name, never `el['name']`, the specific *release's* own
credited display string. When the two diverge in script—a real,
observed MusicBrainz data shape, not a corpus invention—the applied
artist silently discards what the release itself documents. The two
rows below fire in opposite script directions, ruling out a simple
"beets always prefers script X" explanation.

| case | derived (record) | release's own credit (per `--facts`) | applied artist | suspected wrong | citation |
| --- | --- | --- | --- | --- | --- |
| `real-nonlatin-girls-generation` | distance `0.076471`, `artist: 0.076471` (sole nonzero), **medium, skip** | native Hangul `소녀시대` | `Girls' Generation` (Latin, matching a documented `locale=en` alias) | applied artist diverges from this release's own credited display, sourced from the artist entity instead | `#1916`/`#6358` |
| `real-nonlatin-anna-vissi-essential` | distance `0.014493`, `artist: 0.014493` (sole nonzero), **strong, apply** | Latin `Anna Vissi` | `Άννα Βίσση` (Greek, matching a documented `locale=el` alias) | same mechanism, opposite script direction; **the divergence is invisible to a quiet-mode import** (strong/apply, no review) | `#1916`/`#6358` |

**Release truth (both).** Each release's own metadata credits the
display named above (confirmed via `harness seed-items --facts`
against the case's own captured, unedited response, never by reading
`responses/*.json` directly); a listener importing that specific
release would reasonably expect the applied tag to match what the
release itself documents, not a string sourced from a different
MusicBrainz record (the artist entity).

**Distinct scoping finding, not itself a reproduction:**
`_preferred_alias` (`musicbrainz.py:128-155`) is structurally
unreachable under both shipped profiles (`estate`, `stock`), since
`import.languages` is empty in both and its generator yields nothing.
Hypothesis 5 below probes this mechanism directly.

**What ruling this class covers.** Whether the artist-entity-name-
over-release-credit substitution is a real defect, and whether it
should move `real-nonlatin-girls-generation`'s applied `artist` (the
case that stays under review) and, separately, whether
`real-nonlatin-anna-vissi-essential`'s silent strong/apply divergence
(no review surfaced at all) warrants its own severity note distinct
from the medium-band case.

---

## Hypothesis 5: locale-specific alias silently ignored (`#6358`, the `_preferred_alias` mechanism itself)

`bug-locale-alias-ignored`, citation `#6358` (primary locale aliases
ignored after `mbsync`, 2026). Distinct from Hypothesis 4: that
hypothesis is the entity-name-over-release-credit substitution;
this one targets `_preferred_alias` itself, gated by
`import.languages`. `reproduced: "no"`—the evidence below argues
the mechanism works, contradicting the report.

**Suspected defect (as reported by `#6358`).** A documented,
primary, locale-specific artist alias is silently ignored after
`mbsync`, implicating `_preferred_alias`
(`beetsplug/musicbrainz.py:128-155`), gated by `import.languages` and
reachable from both `mbsync` and the importer's own candidate-
building.

**The probe.** `real-nonlatin-utada-first-love-locale` is
byte-identical to its host case, `real-nonlatin-utada-first-love`
(Task 6), in `Input.Items` and `responses/`, differing only in
`profile`: `corpus/profiles/locale.yaml` sets `import.languages: [ja]`
and `import.ignored_alias_types: []` where `estate.yaml` leaves both
empty.

| case | profile | derived (record) | applied sort fields |
| --- | --- | --- | --- |
| `real-nonlatin-utada-first-love` (host) | `estate` | distance `0.072222`, `artist: 0.072222` (sole nonzero), **medium, skip** | `artist_sort`/`albumartist_sort`: `Utada, Hikaru` (the artist entity's own default sort name) |
| `real-nonlatin-utada-first-love-locale` (probe) | `locale` | distance `0.072222`, `artist: 0.072222` (sole nonzero), **medium, skip**—identical to the host | `artist_sort`/`albumartist_sort`/`artists_sort`/`albumartists_sort`: `うただヒカル` (the sort name carried by the release's own documented `locale=ja` alias) |

`best_album_id`, `recommendation`, `action`, and the `artist` penalty
are identical across both records; `artist`/`albumartist`/
`artist_credit`/`albumartist_credit` are also unchanged in both
(`宇多田ヒカル`/`Hikaru Utada`), since Hypothesis 4's own mechanism
already sources those fields from the artist entity's canonical name
regardless of `import.languages`. The shift is isolated to the four
sort-name fields, which only move when `_preferred_alias` returns a
non-nil match.

**Release truth.** The release documents a `locale=ja` alias for its
credited artist, carrying its own hiragana sort_name `うただヒカル`
(confirmed via `harness seed-items --facts` against the case's own
captured, unedited response). A listener whose beets config sets
`import.languages` to include `"ja"` would reasonably expect that
alias to be preferred; the `locale` sibling's record shows it is,
which is the direct, isolated evidence that `_preferred_alias` fires
and its selected alias's fields propagate at pinned beets 2.13.1.

**What ruling this hypothesis covers.** Whether this A/B evidence is
sufficient to close `#6358` as `oracle-correct` for `_preferred_alias`
itself (distinct from Hypothesis 4's separate, already-reproduced
entity-name substitution, which remains open regardless of this
ruling), or whether a further probe is owed before the report is
considered addressed.

---

## Summary table

| hypothesis | bug id | citation | rows | reproduced | current disposition |
| --- | --- | --- | --- | --- | --- |
| 1. featuring-pattern false positive | `bug-ft-pattern` | `742ebdf69`/`7b59604c5`/`19d091440` | 6 | yes | pending-adjudication |
| 2. parenthetical/remix discount | `bug-remix-parenthetical` | `#5868` | 4 | yes | pending-adjudication |
| 3. assignment rejection gap | `bug-assignment-rejection-gap` | `#2981`/`#6078` | 6 | yes | pending-adjudication |
| 4. non-Latin alias mis-selection | `bug-nonlatin-alias-selection` | `#1916`/`#6358` | 2 | yes | pending-adjudication |
| 5. locale alias ignored | `bug-locale-alias-ignored` | `#6358` | 1 | no | pending-adjudication |
| (already ruled, not in this batch) | `bug-threshold-boundary` | `#5458` | 0 | no | oracle-correct (Geoff, 2026-09-03) |

**19 pending rows across 5 hypotheses.** Ruling a hypothesis
(`oracle-wrong` with a `corrected_verdict`, `oracle-correct`, or
`no-correct-answer`) covers every row grouped under it in one pass;
individual rows may still diverge in outcome within a hypothesis if
Geoff finds one case's evidence weaker than its siblings'.
