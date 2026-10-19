# Adversarial review: the rung 3 pass 3b plan

Fresh-context review (2026-09-04) of
`plans/2026-09-04-rung-3b-translation.md` revision 1, in two independent
lenses: beets fidelity, verified contract by contract against pinned
beets 2.13.1 in `tools/venv/.venv`, and harness, process, and plan
integrity, verified against the tree at `f40a669`. Folded as revision 2
the same day; the merged 26-item "Verified sound" list below is the
do-not-re-litigate record for the pass, alongside the rung 3 spec review
and 3a's own review record.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1 through B5, M-a through M-l, m1 through m20. Every
finding was re-verified by the fold editor against the tree or the venv
before folding. Four claims did not survive that check and are rejected
below with the evidence. Four items are conductor-owed rather than folded
silently, and each is applied provisionally in the plan with an inline
marker.

---

## Rejected, with the evidence

**1. "Every corpus release event carries ISO codes" (review 2's B1, its
count only).** The finding is right and is folded as B2; its figure is
overstated in both reviews and is restated from a fresh measurement.
Counted over the 303 release-lookup response files: 298 top-level
`release-events` objects, 290 of them carrying a non-null `area` with
`iso-3166-1-codes`, across 208 case directories. Counted recursively over
all 351 responses, so that search-hit releases are included: 413 events,
405 coded. So eight top-level events do carry a null area, and review 2's
"405 of 405" is a recursive count read as a top-level one. Review 1's
"290 coded over 208 case directories" is exact.

**2. `CompilePatternOption` is cited at the wrong line (review 2's m11).**
Rejected. `func CompilePatternOption` is declared at
`internal/engine/matcher/penalty.go:368`; `:367` is the last line of its
doc comment. The plan's `:368-378` points at the function.

**3. The country truthiness guard is `distance.py:499-502`, not
`:498,501` (review 2's m11).** Rejected. `:499` and `:502` are the
`add_priority` and `add_string` calls; `:498` (`if album_info.country and
options`) and `:501` (`elif likelies["country"] and album_info.country`)
are the guards, which is what the plan's nullability table column names.
The plan is right.

**4. The `internal/engine/{matcher,musicbrainz}` contested-surface row is
a launch blocker (plan's own Task 0 item 4, and review 2's M-d as
written).** Downgraded rather than rejected outright, on review 1's m10.
The phase-1 lane table already awards lane E2 the territory
`internal/engine/matcher`, `musicbrainz`, `shadow`
(`docs/specs/2026-09-03-parallel-build-approach.md:80`), so
the missing entry is a restatement in the contested-surface table rather
than an unowned surface, and 3b is the only live code lane. The row is
still owed; it stops no launch. See M-j.

---

## BLOCKERS

### B1. Task 2's derivation shape cannot produce three of the ten figures, its input envelope is never defined, and `DefectFreeSubject` is a type nothing declares

Both lenses, independently (review 1's B1, review 2's B3).

R19 (`plan:1081-1086`) says `derive_defect_free.py` takes
`derive_component.py`'s shape, "no stub, no response, no rendered
profile", and "computes `string_dist` figures and album distances from
values it is handed, under a monkeypatched `SD_PATTERNS`, and touches no
corpus case". `plan:1619-1624` then requires that, run with `SD_PATTERNS`
untouched, the script reproduce each case's committed record distance.

Three defects sit in that pair of sentences.

**The envelope is never stated, and its type is undeclared.**
`DefectFreeSubject` appears exactly once, in `DeriveDefectFree`'s
signature (`plan:1570`), and is defined nowhere.

**An album distance cannot be computed from scalars.** `distance(items,
album_info, item_info_pairs)`
(`$SP/beets/autotag/distance.py:429-432`) needs the candidate's whole
track list, and `item_info_pairs` needs `assign_items`
(`$SP/beets/autotag/match.py:53`). The candidate side exists nowhere but
the response body: `corpus/cases/adv-ft-craft-titles/case.json`'s
`expected.candidates[0]` carries `source`, `album_id`, `distance`, a
per-key `penalties` map and three per-track distances, and no candidate
title, artist, length or track list at all.

**The arithmetic is not uniform across the ten.** Read off the committed
records, the penalty the substitution moves is `album` on seven rows,
`artist` on `adv-ft-theft-artist`, and `tracks` on `adv-ft-craft-titles`
and `adv-remix-track-titles`. The two `tracks` rows route three per-track
`track_title` distances through `dist.add("tracks", ...)` per matched
pair, and a changed `SD_PATTERNS` changes every `track_title` distance
and therefore the cost matrix `assign_items` solves. A script handed "the
two strings and the committed raw and max" reproduces seven of the ten
and silently cannot reproduce the other three.

**The achievable shape works, measured by the fold editor.** Loading the
case's own release-lookup response, running
`MusicBrainzAPI._normalize_data`, `MusicBrainzPlugin.album_info`,
`match.assign_items` and `distance.distance` under an empty scratch
`BEETSDIR` with `pin_date(eval_date)`, and building items from
`input.items`, reproduces the committed distance bit-identically and
produces the defect-free figure the plan's table names:

```
adv-ft-draft-colon      base 0.010185185185185186 == committed   free 0.10185185185185186
adv-ft-craft-titles     base 0.00917075163398693  == committed   free 0.09170751633986929
adv-remix-both-sides    base 0.023999999999999997 == committed   free 0.08
adv-remix-track-titles  base 0.02456140350877193  == committed   free 0.08187134502923976
```

One implementation mechanic the fold editor hit and the task would
otherwise rediscover: `beets.autotag.distance` resolves to a **function**
in the `beets.autotag` package namespace, so the monkeypatch target is
the module object, reached through
`importlib.import_module("beets.autotag.distance")` or
`sys.modules["beets.autotag.distance"]`. A naive `from beets.autotag
import distance as D; D.SD_PATTERNS` raises `AttributeError`.

*Fold:* R19's second paragraph is rewritten, `DefectFreeSubject` is
declared, the four steps are named, the assignment is re-solved rather
than read from `expected.assignment`, and the monkeypatch target is
stated. Review 1's measurement that the optimal mapping is unchanged on
all ten cases travels as an expectation the task verifies, not as a
premise.

### B2. R15's third declared gap rests on a false measurement, and the plan contradicts itself two paragraphs later

Both lenses (review 2's B1, review 1's M-a).

`plan:992-997` states that "measured mechanically over all 351 responses,
**not one release event anywhere in the corpus carries a non-null `area`
with `iso_3166_1_codes`**", and concludes the preferred-country branch is
"unreachable even under a hypothetical fourth profile". The measurement
is false in both framings: 290 coded events over 208 case directories in
the 303 release-lookup responses, 405 coded events recursively over all
351. `corpus/cases/clean-match`'s single event reads `('1984-04-24',
['GB'])`.

The gap itself is real, and its only true cause is the first half of the
same sentence: no committed profile sets a `match:` key
(`corpus/profiles/estate.yaml`, `stock.yaml`, `locale.yaml`, all read
directly), so every case takes the fallback at
`$SP/beetsplug/musicbrainz.py:176`. The plan then contradicts its own
resolution at `plan:1002-1004`, which presupposes the reachability the
sentence denies.

The false half is repeated at `plan:2338-2339` (Task 6's behavior
contract) and `plan:2725-2726` (Task 9's declared gap 3, which lands in
`docs/HISTORY.md` as a rediscover-guard). A false do-not-rediscover
record is the expensive kind.

*Fold:* all three sites restated. The branch is fully reachable
data-side and unreached only because no committed profile sets
`preferred.countries`; a fourth profile setting
`match.preferred.countries: [GB]` would reach it on 208 case directories
immediately. One concrete case is named so the future corpus lane has a
target: `various-artists` carries two events dated `2014-07-28` (`XE`)
and `2014-08-29` (`DE`), so a preferred country of `DE` would move both
the country and the date on that case.

### B3. `Track.Position int` fails to decode 297 committed track positions, and three of the eighteen ledger subjects break on the first run

Review 1's B2, verified.

`plan:1846-1854` declares `Position int \`json:"position"\``. beets casts
explicitly, `int(track["position"])`
(`$SP/beetsplug/musicbrainz.py:592`), and the client's own TypedDict
claims `position: int`. The corpus disagrees with the TypedDict:
`internal/harness/synth/release.go:128` declares `Position string` and
`:200-203` writes `strconv.Itoa(tr.Position)`, so every synthetic case's
response carries a JSON string.

Counted mechanically over `corpus/cases/*/responses/*.json`: **297 string
track positions across 76 case directories**, including three ledger
subjects, `media-baseline` (6), `media-uneven-split` (6) and `va-exact`
(3). Probed in Go:

```
json.Unmarshal(`{"position":"3"}`, &struct{ Position int }{}) ->
  json: cannot unmarshal string into Go struct field .position of type int
```

Task 8 decodes each committed release-lookup response into
`musicbrainz.Release` (`plan:2578-2585`), so the translation tier fails
on three of its eighteen subjects immediately, and 3c and 3d inherit the
break across 76 cases.

*Fold editor's own census, so the fix is scoped once rather than
rediscovered per field.* Every field the plan's `Release` vocabulary
declares was type-censused across all 399 release documents in the
corpus. `track.position` is the **only** mismatch: `int` 6,763 and `str`
297. Everything else decodes: `medium.position` is `int` 576 and `null`
130 (a JSON null into an `int` is a no-op and leaves 0, matching beets'
falsy `None`), `medium.title` `str` 576 / `null` 130, `medium.format`
`str` 693 / `null` 13, `release.date` `str` 379 / `null` 20,
`release.status` `str` 394 / `null` 5, `track.length` and
`recording.length` `int` with nulls, `track.title` and `recording.title`
always `str`, `labelinfo.catalog-number` `str` 311 / `null` 32,
`labelinfo.label` `dict` 340 / `null` 3.

### B4. Task 8's profile-merge decoder cannot be built as specified: no YAML dependency exists, the plan forbids adding one, and the "only two keys differ" premise is false

Both lenses (review 2's B2, review 1's M-b).

`plan:2591-2598` requires a decoder over the three committed profiles
where "**Only two keys differ across the three profiles today**" and "a
profile key the decoder does not know is a **failure naming the key**".
Three tree facts break it.

- `go.mod` carries cobra plus two indirects and nothing else, and no Go
  file in the tree imports a YAML package. The repo's only profile reads
  are `loadProfile` (`internal/harness/oracle/oracle.go:57-62`, raw
  bytes) and `countMetadataSources` (`:254-268`, line-prefix string
  matching). `plan:334-337` says "Module dependencies: none added ... a
  task that wants one stops and reports".
- `cassette.Case.Profile` (`internal/harness/cassette/cassette.go:44`) is
  a profile **name**, not content, so nothing hands the decoder parsed
  keys.
- The profiles do not differ by two keys. `stock.yaml` is one line
  (`plugins: musicbrainz`). `estate.yaml` sets `directory`, `library`,
  `import.quiet_fallback`, `import.none_rec_action`, `import.timid`,
  `original_date`, `per_disc_numbering`, `plugins`. `locale.yaml` is
  estate plus `import.languages` and `import.ignored_alias_types`. Five
  of those keys have no `Tunables` home by design, two of them
  (`original_date`, `per_disc_numbering`) because 3a's M2 resolution
  deliberately keeps them out. An unknown-key-fails decoder reds on all
  three profiles, which is all 45 rows, on day one.

*Fold, taking review 2's mechanism with review 1's honesty rule
attached.* The "decoder" becomes a profile-**name**-to-`Tunables`-overlay
table in `internal/harness/conform`: three names, one overlay
(`locale` sets `Languages: ["ja"]` and `IgnoredAliasTypes: []`), and an
unrecognized profile **name** is the failure. No YAML is parsed and no
dependency is added. The unknown-key rule is not dropped, because
dropping it would let a future `match:` key land in a profile without
reaching the translation; it moves instead to a drift guard that scans
each committed profile's own text for top-level and `import:` key names
against a written allowlist, in `countMetadataSources`' line-prefix
shape, failing and naming any key the allowlist does not carry. The five
keys outside `Tunables` are named in the allowlist with their reason.

### B5. R3's third route to the `"Media"` sentinel is wrong, and Task 5 mandates a fixture asserting the wrong value

Review 1's B3, verified.

`plan:713-722` says the sentinel's three routes are "zero surviving
tracks, mixed per-track formats, and a single-medium release whose
`format` is null", and `plan:2231-2234` turns that into a required
fixture.

beets is `medias.pop() if len(medias := {t.media for t in track_infos})
== 1 else "Media"` (`$SP/beetsplug/musicbrainz.py:654-658`). A single
null-format medium gives the set `{None}`, whose length is 1, so `pop()`
returns `None`. Verified against pinned beets: `[None]` gives `None`;
`['CD']` gives `'CD'`; `['CD','Vinyl']` and `[None,'CD']` give
`'Media'`; `[]` gives `'Media'`.

The plan contradicts itself. Task 8's nullability table (`plan:2613`)
gets it right, saying `medium["format"]` is `str | None` "so a
single-medium null format really does yield None". R3 and Task 5 get it
wrong, and Task 5 is the half an implementer builds from.

*Fold:* R3 restated to two routes to the sentinel, zero surviving tracks
and two or more distinct formats, with a `None` counting as one of the
distinct values so `[None, "CD"]` is a sentinel route worth its own
fixture. Task 5 gains a third fixture asserting the **non**-sentinel
outcome: a single null-format medium yields `AlbumInfo.Media == ""`, the
R1 zero-value mapping of beets' `None`, which is falsy and so declines
the media penalty exactly as `if album_info.media` does
(`distance.py:455`).

---

## MAJOR

### M-a. The committed `defect_free.distance`'s precision is never stated, and two of Task 2's own mutations require the answer

Review 2's M-j, verified.

`plan:1631-1642` presents nine-decimal figures as the value, while the
committed records carry full float64:
`corpus/cases/adv-remix-both-sides/case.json` reads
`0.023999999999999997` against the table's `0.024000000`, and
`adv-ft-theft-artist` reads `0.00909090909090909` against `0.009090909`.
`plan:1674-1675`'s mutation 7 perturbs a committed
`defect_free.distance` "by one ULP", which only works if the stored value
is the derivation's own float. `plan:1619-1624`'s "reproduce each case's
committed record distance to the last digit" conflates the two
comparisons.

*Fold:* the block stores the derivation's own unrounded float64; the
plan's table is the acceptance cross-check against the owner's prose at
the precision the prose states; and the both-ways check compares
bit-identically against the committed `case.json` distance.

### M-b. `CheckAlbumInfoLedger`'s check 6 is self-referential and can pass vacuously

Both lenses (review 1's M-c, review 2's M-a).

`plan:1386-1391` derives the pseudo-case set "from the rows themselves (a
row carrying `Merged`), not hardcoded", then asserts that no case outside
that set carries one. That is true by construction. If the derivation
emits no `Merged` block at all, the derived set is empty, the check
passes, and the merge is silently ungraded. Mutation 7 (`plan:1458`) adds
`clean-match` to the derived set and the check still passes. This is the
silent-success shape the spec's assertion discipline forbids (spec
`:316-321`), in the one checker whose job is to catch a derivation that
recorded the wrong thing.

The expected set is measurable off the tree rather than hardcoded:
exactly four case directories carry a backward `transl-tracklisting`
relation, `non-latin-pseudo-release`, `real-nonlatin-girls-generation`,
`real-nonlatin-zemfira`, `real-nonlatin-zemfira-stock` (fold editor's own
count, agreeing with both reviews and with R17).

*Fold:* `analysis` gains a named `AlbumInfoPseudoSubjects` list beside
`AlbumInfoSubjects`, carrying the same why-strings and the measurement in
its doc comment. Check 6 becomes: a row carries `Merged` only if its case
is in that list, and each listed case has exactly one row carrying
`Merged`. Mutation 7 then fires, and a second mutation, removing
`non-latin-pseudo-release`'s `Merged` block, proves the other direction.

### M-c. Task 8's nullability field-count assertion is off by one and fails as written

Both lenses (review 2's M-b, review 1's m1; taken at the higher
severity because it fails as written).

`plan:2620-2625` asserts the table's row count "against
`reflect.TypeOf(matcher.AlbumInfo{}).NumField()` plus `TrackInfo`'s".
`internal/engine/matcher/info.go` declares 14 `AlbumInfo` fields and 8
`TrackInfo` fields (counted off the tree), so the assertion compares 21
against 22 and fails. The table is right at 21 rows: `AlbumInfo.Tracks`
is compared as a list rather than through the policy.

*Fold:* assert `NumField(AlbumInfo) - 1 + NumField(TrackInfo)`, with the
`- 1` named in a comment as `Tracks`, so a fifteenth vocabulary field
still fails.

### M-d. `parseArtistCredits`' stated guarantee is not one `Validate` provides, and the Go port panics on it

Review 1's M-d, verified.

`plan:2038-2039` says "Callers pass a credit list `Validate` has
accepted, so the empty-list case that raises `IndexError` at
`musicbrainz.py:350` cannot reach here." `Validate` checks an empty
`artist-credit` array at **release** level only (`plan:1921-1928`).

beets calls `_parse_artist_credits` at three sites: the release credit
(`$SP/beetsplug/musicbrainz.py:650`), the recording credit inside
`track_info` (`:457`), and the track-level override that swaps a track
credit in for a recording credit (`:586-588`). `artists_ids[0]` at `:350`
raises on any of them, and Task 5 routes `TrackInfo.Artist` through the
same function, so a release whose track and recording credits are both
empty panics with an index-out-of-range inside
`internal/engine/musicbrainz` rather than returning `ErrMalformedRelease`.

The corpus carries no such case, so this ships as a latent panic that
3c's live client is the first thing to meet.

*Fold:* `Validate` walks every surviving medium's tracks and requires a
non-empty resolved credit list, the track's own when non-empty and the
recording's otherwise, naming the medium position and track position on
failure. The doc comment's guarantee is then true. A hand-authored
fixture and a named mutation ship with it, and the pass report records
that the corpus carries no instance.

### M-e. `PseudoReleaseTarget` can nil-deref, and R10's table omits `rel["release"]["id"]`

Review 2's M-i, verified.

beets reads `rel["release"]["id"]` at
`$SP/beetsplug/musicbrainz.py:794`, a direct index that raises
`KeyError`. The Go type is `Release *struct{ ID string }`
(`plan:1866-1868`), a pointer that is nil for most relations: relations
carry `target-type` `url` (351), `release` (36) and `artist` (285) across
the corpus, and an artist relation carries no `release` key at all.
`Validate` checks three things and none of them is this, and R10's table
(`plan:870-876`) omits `:794`. So a release whose status is
`Pseudo-Release` and whose matching relation lacks a release object
panics inside the engine on external input, which is the exact failure
R10's single-sentinel boundary discipline exists to prevent.

No corpus instance exists: zero backward `transl-tracklisting` relations
and zero `target-type: release` relations lack a release object.

*Fold:* the site joins R10's table and `Validate`'s rule list, so
`Validate` grows from three rules to five (this and M-d), and
`PseudoReleaseTarget`'s doc comment states the precondition.

### M-f. The `defect_free` population's owning task is mis-numbered as Task 8 in four places, and Task 2 never inherits R14's acceptance criterion

Review 2's M-c, verified.

`plan:130`, `:952`, `:955` and `:973` all say "Task 8" for work Task 2
owns (`plan:1474`, `:1625-1646`). Task 8 is the translation tier and its
Files list does not include `adjudications.json`.

The consequence is not cosmetic: R14's byte-identity-of-the-other-ten-
fields criterion, which R14 itself calls "a task failure" if violated,
appears nowhere in Task 2's behavior contract or its deliverable count.
Task 2's own byte-identity bullet is R20's different invariant.

*Fold:* all four references renumbered to Task 2, and R14's criterion
restated inside Task 2 as an explicit acceptance line.

### M-g. The mechanism that writes `defect_free` into `adjudications.json` is unspecified, while R14 forbids any other field moving

Review 2's M-f, verified and strengthened.

`adjudications.json` is **not** in the derived set: `corpus/README.md:2214-2218`
lists `ties.json`, `components.json` and `file_tags.json` as the files
rewritten rather than hand-edited. So neither candidate mechanism is
safe by default, and the plan names neither.

Measured by the fold editor:

- A Python round trip, `json.load` then `json.dumps(indent=2,
  ensure_ascii=False) + "\n"`, reproduces the committed file
  **byte-identically today** (49,252 bytes both ways), because `json`
  preserves the file's own key order and the file is already at that
  spelling.
- A Go re-serialize through `AdjudicationEntry` would **not**. The file's
  key order per entry is `id, subject, suspected, citation,
  release_truth, disposition, reason, adjudicator, date,
  corrected_verdict`, while the struct
  (`internal/harness/analysis/adjudication.go:45-56`) emits
  `corrected_verdict` immediately after `disposition`. That reorders 18
  rows and trips R14's own failure condition on the first run.

*Fold:* the mechanism is named as a targeted Python writer inside the
derivation, setting `entry["defect_free"]` on exactly the ten resolved
ids so the key appends last, and writing with the measured spelling. The
task reports `git diff --stat` and a per-field diff over the ten rows as
evidence. A `-update-ledgers` re-serialize of the whole file is named as
the rejected alternative with the reason.

### M-h. R3's claim that the corpus's ten `"Media"` instances are Task 8's grading evidence is false for six of them, and the mixed-format route reaches no ledger row

Review 2's M-g, verified.

Recomputed under `DefaultTunables` (`IgnoreDataTracks: true`,
`IgnoreVideoTracks: true`, `IgnoredMedia: nil`), exactly ten release
documents land on the sentinel and five carry zero surviving tracks,
which is T4's figure exactly. They live in six cases:
`admit-zero-track-dropped` (0 tracks), `real-nonlatin-hamasaki-ayumi`
(four rows, 0 tracks), `real-noise-bowie-us1967` (CD and SHM-CD),
`real-noise-dylan-bootleg` (two rows, Hybrid SACD layers),
`real-reissue-doolittle-2008` and `real-reissue-doolittle-2008-stock`
(Hybrid SACD layers). Only `real-nonlatin-hamasaki-ayumi` is a ledger
subject, so four of the ten reach Task 8 and all four are the zero-track
route.

*Fold, provisional and conductor-owed.* `real-noise-dylan-bootleg` is
added as a nineteenth ledger subject so the mixed-format route is graded
against beets rather than only against a hand-authored fixture. Review 2
costed this at "+2 to the floor"; measured off the tree the case carries
**five** release-lookup requests, so the floor moves 45 to 50.

The alternative is to correct the sentence to "four of the ten, all on
the zero-track route" and record the mixed-format route as fixture-graded
only. It costs nothing and grades less. See `conductor_rulings_needed`.

### M-i. `credits.SortName` is a struct field nothing reads, and its stated rationale is false

Review 2's M-h, verified.

`plan:2020-2029` says `SortName` is there "because the ruling-E ledger's
own derivation records it beside the artist". It does not: R6
(`plan:766-769`) puts `artist_sort` in rung 8's territory and out of the
graded set, and `LedgerAlbumInfo` (`plan:1323-1338`) carries no sort-name
field. Tracing consumers, `TranslateAlbum`'s composition reads `c.Artist`
and the VA flag, `applyVariousArtists` reads `ArtistID`, and `SortName`
is read nowhere. Task 4's own test table asserts it, which makes it
tested dead data, the shape 3a's close had to fix (`docs/HISTORY.md`, 3a:
the `Key` field "sat unread from Task 5 through the start of Task 8").

*Fold:* `SortName` is dropped from `credits`, the A/B table loses its
`SortName` column, and `credit.go`'s doc comment states that
`artist_sort` is rung 8's territory and deliberately not produced. The
alias-versus-credit split stays tested through `Artist` and the credit
string.

### M-j. Two contested-surface rows are marked "written" that no table grants, and the engine row is a restatement rather than a blocker

Merged: review 2's M-d for the two false "written" rows, review 1's m10
for the downgrade.

- `plan:360` claims `internal/harness/corpus_test.go` is "covered by the
  `{cassette,stub,analysis}` row's consumers". That row
  (`parallel-build-approach.md:163`) covers
  `internal/harness/{cassette,stub,analysis}`; `corpus_test.go` sits in
  the parent `internal/harness` package, outside all three, and a
  Consumers column is not an ownership grant. 3b writes this file twice.
- `plan:356`: that row's Owner column reads "lane E1 as shipped; rung 3
  pass **3c** makes the ruled `Case.Extra` schema-and-stub change". It
  carves out one rung-3 exception and it is 3c's. 3b's new
  `analysis/album_info.go` and its `AdjudicationEntry` change are not
  covered.
- The engine row is genuinely missing from the contested-surface table
  and is genuinely owed, but the lane table at `:80` already awards lane
  E2 `internal/engine/matcher`, `musicbrainz`, `shadow`, and 3b is the
  only live code lane.

*Fold:* the conductor-owed request widens to three grants, and Task 0
item 4's expected answer names three missing rows rather than one, none
of them a launch stop. See `conductor_rulings_needed`.

### M-k. Task 0 item 3's expected answer contradicts `docs/STATUS.md`, and its blocker predicate names a pid the registry does not carry

Review 2's M-e, verified.

`plan:1138-1145` expects "**lane-1 idle**" and calls "a lane-1 entry
still reading PREP with a live pid" a blocker. `docs/STATUS.md:29-33`
reads "lane-1 (code, main tree): **PREP for 3b**. The 3b plan-prep
workflow ... **is running**, run `wf_5b24de07-ef0`", and it records no
pid at all, where `parallel-build-approach.md:194` requires the entry to
carry "lane name, worktree path, pid". So the pre-task cannot evaluate
the predicate, and the honest answer, "PREP, no pid recorded", is turned
into an unresolvable blocker by the plan's own rule.

*Fold:* the expected answer is restated as the tree gives it, and
flipping lane-1 to a 3b-execution entry with a pid is named as a
conductor edit owed before Task 1 dispatches rather than a finding for
the pre-task to litigate.

### M-l. Three of R7's subject rationales are contradicted by the tree, and no subject grades release-event selection at all

Review 1's m2, verified, promoted to major because `plan:1273-1291`
writes these strings verbatim into committed Go doc data.

Measured off the tree:

- `clean-match` is described as having "no release events" and carries
  one, `('1984-04-24', ['GB'])`, equal to the top-level date.
- `real-genre-metallica-master-of-puppets` is described as carrying "two
  or more `release-events` whose dates differ from the top-level `date`"
  and carries one event, `('1999-07-13', ['US'])`, equal to the top-level
  date.
- `real-compilation-queen-greatest-hits-ii` is described the same way and
  carries two events, both `1991-10-28`, both equal to the top-level
  date.

More generally, with `preferred.countries` empty every case takes
`_preferred_release_event`'s fallback
(`$SP/beetsplug/musicbrainz.py:176`), so no subject grades event
selection. `various-artists` is the one subject whose events differ from
each other (`2014-07-28` XE and `2014-08-29` DE), which is what makes it
the concrete target B2's restated gap names.

*Fold:* the three why-strings re-derived off the tree, and the plan says
once that release-event selection is graded by Task 6's hand-authored
fixtures rather than by any subject.

---

## MINOR

- **m1.** `docs/STATUS.md:66` records 3a's spend as "≈2.26M/6M";
  `docs/HISTORY.md`'s 3a Budgets line records ≈2.84M and itemizes it to
  2,836K. The plan's whole ceiling basis (`plan:175-187`) takes HISTORY's
  figure, and Task 0's checklist never reads STATUS's spend line.
  Conductor-owed ledger correction, not a plan edit. (Review 2's m1.)
- **m2.** R16's correction 4 cites `corpus/README.md:2469-2470`; the "a
  no-op on `stock.yaml` and `locale.yaml`" sentence begins at **`:2467`**.
  The other three citations verify exactly, as does correction 4's
  substance. Task 0 item 11 checks the cited lines and would report the
  correction already landed. (Review 2's m2.)
- **m3.** `AlbumInfoSubject.Why` (`plan:1266-1269`) is read by none of
  `CheckAlbumInfoLedger`'s six checks, where its cited precedent
  `FileTagSubject` (`internal/harness/analysis/file_tags.go:23-26`)
  carries `CaseID` and `ItemIndex` only and puts the rationale in
  `FileTagSubjects`' own doc comment. Folded as an adaptation rather than
  the literal fix: `Why` stays and gains a reader, a check that every
  subject carries a non-empty one, so a subject added without a rationale
  fails. (Review 2's m4.)
- **m4.** `plan:311-315` claims a mismatch "or the reverse" reds the gate.
  `readJSON`'s `DisallowUnknownFields`
  (`internal/harness/analysis/ledger.go:20`) rejects a JSON key with no
  Go field only; a Go field absent from the JSON stays zero and the gate
  stays green. The conclusion survives, the stated mechanism is
  one-directional. (Review 2's m5.)
- **m5.** `plan:131-132` lists seven fields Task 2 must leave
  byte-identical, omitting `id`, `subject` and `citation`.
  `AdjudicationEntry` has ten existing fields, which is what R14's "other
  ten fields" correctly says. (Review 2's m6.)
- **m6.** `plan:1394-1396` has the driver hand the script "the list of
  `(release_mbid, is_search)` lookups". Nothing consumes `is_search`, and
  R9 rejected a dead parameter for the same reason. All 45 lookups are
  mbid paths; the search endpoint is `/ws/2/release?`, which the plan's
  own jq filter correctly excludes. (Review 2's m7.)
- **m7.** Task 8's mutation 7 expects the lowered guard to name
  `real-boxset-rachmaninov-complete-stock` at 470. Recomputed: **both**
  `real-boxset-rachmaninov-complete` and its stock twin sum to 470. Name
  both. (Review 2's m8.)
- **m8.** R21 widens the walk's root but not its words.
  `checkNoTimeNow`'s error message hardcodes "which
  **internal/engine/matcher** must never do"
  (`internal/harness/conform/arch_test.go:107`) and its doc comment is
  matcher-specific. Both need the widening the test name gets. (Review
  1's m9 and review 2's m9.)
- **m9.** Task 8's deliverable count is **5**, past the house "roughly
  four" bar, and the plan does not say so where CLAUDE.md requires it to.
  Every other task is at 3 or 4. (Review 2's m10.)
- **m10.** Citation drift, verified: `runPython` is declared at
  `internal/harness/oracle/ties.go:120`, not `:129` (`:129` is the
  `PYTHONHASHSEED=0` line, cited correctly elsewhere);
  `TestMatcherNeverReadsTheClock` starts at `arch_test.go:147`, not
  `:146`; the track-artist truthiness guard spans `distance.py:404-410`,
  not `:405-409`. The other two claims in review 2's m11 are rejected
  above. (Review 2's m11, partly.)
- **m11.** `DefectFree.PatternTable`'s two legal values are the Go
  identifier names `"SDPatternsWordBoundary"` and
  `"SDPatternsNoBracketDiscount"`, asserted as string literals inside
  `internal/harness/analysis`, which must not import `internal/engine`.
  Nothing checks they still resolve, so a rename in `matcher` orphans the
  ledger silently. A `conform`-side test asserting both names resolve to
  a non-empty `[]matcher.SDPattern`, in `CheckUnidecodeVersion`'s shape
  (`conform/defaults.go:67-72`), is the cheap guard. (Review 2's m12.)
- **m12.** `checkAdjudicationEntry`'s new rule 1 forbids a block on a
  non-`OracleWrong` row, and `CheckDefectFree` requires one on the ten
  string-heuristic cases. Nothing forbids a block on the six
  `adv-assign-*` rows, which are `oracle-wrong` and which R12 says "owe
  no block". A fabricated block there would pass both checks. (Review 2's
  m14.)
- **m13.** `Relation.Release *struct{ ID string }` is an anonymous struct
  inside an exported type: invisible to `go doc` as a named shape and
  awkward for the caller M-e now requires to test it. A named
  `RelationTarget` costs two lines. (Review 2's m15.)
- **m14.** The spec's own 3b tree column reads "Main tree: writes
  `corpus/` and `internal/harness/oracle`" (spec `:329`). This pass also
  writes `internal/harness/{analysis,conform,corpus_test.go}` and births
  `internal/engine/musicbrainz`. The plan justifies each and drafts four
  `corpus/README.md` corrections but no spec correction. Conductor-owed.
  (Review 2's m16.)
- **m15.** R10's table claims to enumerate "the five sites where beets
  raises" and is not exhaustive. Also raising: `release["media"]` and
  `m["format"]` in the `valid_media` filter
  (`$SP/beetsplug/musicbrainz.py:638-639`), and `medium_data`'s four
  direct indexes on `position`, `track_count`, `title` and `format`
  (`:565-570`), plus `recording["length"]` and
  `recording["disambiguation"]` in `track_info`. All are inert in Go,
  where an absent key decodes to a zero value, and every committed medium
  carries all four keys. An absent `media` array is the one worth naming:
  beets raises, while Go produces a zero-track candidate carrying the
  `"Media"` sentinel, harmless only because `_add_candidate` discards
  zero-track candidates (`$SP/beets/autotag/match.py:344-347`), which is
  3d's code. (Review 2's m13 and review 1's m8, merged.)
- **m16.** Task 2's derivation never pins the calendar date, while
  `distance()` divides by `datetime.date.today().year` in both year
  branches (`distance.py:477`, `:488`). All ten cases take `dist.add("year",
  0.0)`, so no date is read today, and `derive_oracle.pin_date`
  (`internal/harness/oracle/derive_oracle.py:52-69`) exists for exactly
  this reason. Call it anyway and say why. (Review 1's m3.)
- **m17.** The empty scratch `BEETSDIR` also puts Task 2's run on beets'
  shipped `import.none_rec_action: ask`
  (`$SP/beets/config_default.yaml:34`) where all ten cases derive under
  `estate`, which sets `skip` (`corpus/profiles/estate.yaml:18`).
  Harmless for these ten, because `quiet_action` sets `quiet: True` and a
  medium recommendation resolves through `quiet_fallback`, `skip` in both
  (`derive_oracle.py:193-206`). State the equivalence rather than leaving
  it to be rediscovered. (Review 1's m4.)
- **m18.** `SDPatternsNoBracketDiscount`'s doc comment repeats the corpus
  README's claim that dropping the two bracket rows "is arithmetically
  the same as charging a parenthetical or a bracketed suffix at full
  weight". The two agree only when no later pattern fires, because a
  full-weight discount still shifts `str1`, `str2` and `base_dist` for
  the rest of the loop while a removed pattern does not
  (`$SP/beets/autotag/distance.py:97-117`). Verified equal on all four
  remix pairs, so the ten figures stand. Qualify the sentence. (Review
  1's m5.)
- **m19.** `SDPatternsWordBoundary` uses `\b`, which is Unicode-aware in
  Python's `re` on `str` and ASCII-only in Go's RE2, so on non-ASCII
  input the derived figure and 3d's compat-switch result can differ. The
  six `adv-ft-*` pairs are ASCII, so nothing moves today. Carry it to 3d
  beside the `SD_PATTERNS`-as-a-value seam (3a HISTORY guard 3). (Review
  1's m6.)
- **m20.** Task 6's step 8 says `Label` and `CatalogNum` come from
  `rel.LabelInfo[0]` "with the `"[no label]"` literal comparison". beets
  reads `catalognum = label_info["catalog_number"]` unconditionally and
  gates only `label` on the label object being truthy and its name
  differing from `"[no label]"` (`$SP/beetsplug/musicbrainz.py:497-506`).
  The corpus carries 32 null catalog numbers and 3 null labels. Split the
  sentence and add a fixture row. (Review 1's m7.)

---

## VERIFIED SOUND, the do-not-re-litigate record

The two reviews' lists merged and deduplicated. Everything here was
checked directly against pinned beets 2.13.1 and the tree at `f40a669`,
by at least one reviewer and, where the two lists overlapped, by both.

1. **T1, the single album-side entry point.** `MusicBrainzPlugin.album_info`
   at `$SP/beetsplug/musicbrainz.py:631`, reached only through
   `album_for_id` at `:765`, which is what `albums_for_ids` maps search
   hits through. The pseudo check sits inside `album_for_id` at
   `:785-798`.
2. **T2, key normalization.** `_normalize_data`'s dict handler
   (`$SP/beetsplug/_utils/musicbrainz.py:746-774`) regroups `relations`
   by `target-type` with `sorted` then `groupby`, strips `target-type`
   from each entry, recurses, and rewrites every other key's hyphens to
   underscores. Ascending group-key order, stable within group,
   later-wins collision, and the `AttributeError`/`TypeError` on an
   untyped relation are all exact. The struct-tag discharge of the hyphen
   half is sound.
3. **T3, `mediums` against extraction.** `mediums=len(release["media"])`
   at `:659`, unfiltered, while `valid_media` filters at `:638-640`.
   Re-measured on `corpus/cases/media-baseline`: `[]` gives `(2, 6,
   'CD')`, `['CD']` gives `(2, 0, 'Media')`, `['cd']` gives `(2, 6,
   'CD')`. Case-sensitive exact membership confirmed.
4. **T4, the sentinel's reach.** Ten of the 303 committed release lookups
   produce `media == "Media"`, five of them with zero tracks. Both
   reviews and the fold editor agree on the figures. Only the "which
   cases they live in" inference is wrong (M-h).
5. **T5, assembly order and both filters.** `get_tracks_from_medium` at
   `:555-600`: pregap unconditionally, tracks, data tracks only when
   `ignore_data_tracks` is false, all through `.get`; then the `"[data
   track]"` title drop regardless of `ignore_data_tracks`, then the video
   drop. The four-row truth table reproduces exactly.
6. **T6, the three track-level overrides.** `:582-590`, on a shallow
   copy, with truthiness fallbacks on `length` and `artist_credit` and
   the title replaced only when `track["title"]` is truthy and
   `_preferred_alias(recording["aliases"])` returns nothing. The double
   `_preferred_alias` read (`:589` and `:443`) is real. Per-medium fields
   and the 1-based `index` across all valid media at `:645-646` are
   exact.
7. **T7, `_preferred_alias`.** `:128-155`. All eight measured rows
   reproduce, including the ignored-type fall-through to the **next**
   matching alias rather than to nil, the lowercased ignored set against
   a case-sensitive locale compare, and configured-locale order beating
   response order. `primary` is `bool | None`, so `*bool` with a falsy
   null is right.
8. **T8, artist credits.** `_parse_artist_credits` at `:301-357`
   reproduced verbatim: `[' & ', '']` gives `'A & B'`, a single `' feat.
   '` gives `'A feat. '` with the trailing space, `[' & ', ' & ']` gives
   `'A & B & '`, and an empty list raises `IndexError`. Task 4's A/B
   table reproduces exactly. No `_flatten_artist_credit` symbol exists at
   the pinned version, as the plan says.
9. **T9, the VA override.** `:673-681` replaces the six artist fields
   with `config["va_name"]`; the album artist penalty is skipped under
   `album_info.va` (`distance.py:447-448`) and the track artist penalty
   reads `track_info.artist`, untouched. Ledger-moving and not
   verdict-moving, as stated. R8's exclusion of `aliases_as_credits` is
   correct.
10. **T10, release-event selection.** `_preferred_release_event` at
    `:165-177` reproduced exactly, including that a null `area` is
    skipped by the walrus and that the returned country is the
    **configured** string. R4's two-regimes hazard is real:
    `internal/engine/matcher/album_country.go:31-40` routes the same list
    through `CompilePatternOption`, which anchors and folds case
    (`penalty.go:368-378`).
11. **T11, `_get_date`.** `:180-200`. All six rows reproduce, including
    `'1999-XX-06'` giving `(1999, None, 6)` and `'1999-05-06-07'`
    discarding the fourth part. The `:685-689` fallback tests
    truthiness.
12. **T12, the pseudo merge.** `_merge_pseudo_and_actual_album` at
    `:203-240`: `pseudo.copy()` is `AttrDict.copy`, which is `deepcopy`
    (`$SP/beets/autotag/hooks.py:93-95`), then the fifteen named keys
    from the actual. Seven of the fifteen are in `matcher.AlbumInfo`. The
    four committed cases' deltas reproduce exactly, and `year`, `month`,
    `day`, `original_*`, `mediums`, `style` and `genres` are equal on
    every committed case, so the merge's year path genuinely has no
    corpus signal.
13. **T13, no Unicode normalization.** No `unicodedata.normalize` in
    either `beetsplug/musicbrainz.py` or `beetsplug/_utils/musicbrainz.py`.
    The three case-folding sites named are the only ones.
14. **T14 and T15, config reads and their defaults.** Every citation
    exact: `ignored_media` `[]`, `ignore_data_tracks` `yes`,
    `ignore_video_tracks` `yes`, `preferred.countries` `[]`, `languages`
    `[]`, `ignored_alias_types` `[]`, `va_name` `"Various Artists"`,
    `aliases_as_credits` `False`. The five `@cached_property` reads are
    at `:246-268`, and a fresh plugin per config change is required.
15. **T16, the browse fallback.** `_ensure_complete_recordings` at
    `:602-629`, `BROWSE_MAXTRACKS = 500` at `:67`, `BROWSE_CHUNKSIZE =
    100` at `:66`, and the sum counts `medium["tracks"]` only, across all
    media, unfiltered. The largest committed release-lookup response is
    470 tracks, so Task 8's mutation 7 fires (see m7 for the case count).
16. **R1's resolution and its evidence.** Exactly three `add_string`
    sites on the distance path are unguarded on the info side: `album` at
    `distance.py:451`, `artist` at `:448` guarded only by `not va`, and
    `track_title` at `:402`. Every other component guards both sides. All
    three are `str` by construction through the MusicBrainz path. The
    largest committed `album` penalty is 0.169 over 294 candidates. The
    plain-string vocabulary is the right call.
17. **The ten `defect_free` figures and both pattern tables.** All ten
    derived and defect-free figures match `adjudications.json`'s prose.
    `SDPatternsWordBoundary` reproduces every `adv-ft-*` figure and
    `SDPatternsNoBracketDiscount` every `adv-remix-*` figure, four of
    them re-derived independently by the fold editor. R18's choice of a
    four-entry table with the two 0.3-weight bracket rows dropped is
    correct, and `corpus/README.md:1107-1108` and `:1186-1189` say what
    the plan quotes.
18. **R12's predicate and R14's scope.** `oracle_bugs.json` resolves
    `bug-ft-pattern` to six `adv-ft-*` probes and `bug-remix-parenthetical`
    to four `adv-remix-*` probes, both at `oracle-wrong`, with the other
    four hypotheses at 2 `oracle-wrong` and 2 `oracle-correct` and none
    pending. `adjudications.json` carries 21 rows at 18/3/0. All ten
    `corrected_verdict` rows read "medium, skip", so `CheckDefectFree`'s
    categorical overlap is non-empty and can fire.
19. **R6 and R7's arithmetic.** The eighteen named subjects carry
    **exactly 45** `/ws/2/release/` requests, counted off the tree, with
    the per-subject counts matching the plan's table row for row. The
    corpus is 215 cases at 178 `estate` / 36 `stock` / **1 `locale`**,
    and that one case is `real-nonlatin-utada-first-love-locale`. Only
    the three why-strings in M-l are wrong.
20. **The alias-path leverage figures.** Under `import.languages: ['ja']`
    the album `artist` changes on 90 of 303 releases and track `artist`
    on 3,272 of 6,919 tracks. The 6,919 figure is confirmed: 7,060 wire
    tracks less the 141 video recordings `ignore_video_tracks: yes`
    drops. Task 4's upshift to `opus` is justified by these numbers.
21. **R15's first two gaps, and R17.** Zero media anywhere carry a
    `pregap` key or a non-empty `data-tracks` array, so both branches are
    corpus-unreachable. Exactly four case directories carry a backward
    `transl-tracklisting` relation. Only the third gap is wrong (B2).
22. **The harness APIs and tree facts the plan calls.** All present with
    the named shapes: `analysis.LoadAdjudications`, `CheckAdjudications`,
    `AdjudicationEntry` (ten fields), `AgentAdjudicator` and the sentinel
    rule at `adjudication.go:140-142`; `FileTagSubjects` at
    `file_tags.go:39-52` and `CheckFileTagLedger` at `:100`;
    `ledger.go:20`'s `DisallowUnknownFields`; `corpus_test.go:270` and
    `:313`; `conform/defaults.go:21,43,67`; `conform/doc.go:1-23`;
    `arch_test.go:138`, `:147`, and the `testdata` skip at `:37`/`:79`;
    `oracle/ties.go`'s `stub.New` / `renderProfile` / `runPython` chain;
    `derive_component.py:168-190`'s weights swap-and-restore and
    `:443-451`'s empty scratch `BEETSDIR`; `derive_test.go:119`;
    `component_derive_test.go:66`; `file_tags_derive_test.go:119`.
    `internal/engine` holds `matcher` and nothing else;
    `internal/config/keys.go:19-25` still carries `ownedWall` as a slice
    literal.
23. **The four `corpus/README.md` corrections in R16 are all real.**
    `:1096` still says every adversarial case sits at
    `pending-adjudication`; `:1341` still calls `bug-locale-alias-ignored`
    pending; `:2317-2318` still says every hypothesis but
    `bug-threshold-boundary` is pending; and the placeholder sentence
    still claims `locale.yaml` carries neither, where
    `corpus/profiles/locale.yaml:13-14` carries both. Only correction 4's
    line citation is off (m2).
24. **The gate coverage claims hold.** `scripts/check.sh:85` is `mapfile
    -t py_files < <(git ls-files '*.py')` with `ruff check` at `:87` and
    `ruff format --check` at `:90`, so two new tracked `*.py` files are
    linted with no gate edit. `make derive` is `go test -tags oracle
    -timeout 90m ./internal/harness/... ./cmd/harness/...`
    (`Makefile:38-39`), so the byte-identity invariant is genuinely
    exercised by the four existing re-derive tests on every run.
25. **The two architectural invariants are as claimed.**
    `TestEngineNeverImportsHarness` (`arch_test.go:138-144`) already
    roots at `internal/engine`, so `musicbrainz` inherits it free.
    `TestMatcherNeverReadsTheClock` roots at `internal/engine/matcher`,
    so R21's widening is real work, and the existing fixture is flat, so
    the nested fixture is genuinely owed. `internal/harness/analysis`
    imports only `cassette` today.
26. **The plan's structural discipline.** Nine tasks plus a pre-task,
    deliverable counts stated per task, the split point named at the Task
    2 / Task 3 boundary with the recommendation against taking it, two
    named upshifts with reasons, the fresh-context-review refusal
    recorded rather than defaulted, every stop condition routed to the
    conductor rather than to Geoff, and no placeholder anywhere
    (`grep -nE "TODO|TBD|FIXME|XXX"` returns nothing). The main-tree
    placement, the `make derive` scoping to Tasks 1 and 2, the
    byte-identity invariant, and the no-network rule all match the spec
    and the standing policy.

---

## Verdict

Five blockers, twelve majors, twenty minors, four rejected claims, four
conductor-owed items. The plan's beets research is unusually strong:
sixteen translation contracts reproduce against pinned beets essentially
line for line, including the three easiest to get plausibly wrong (the
double `_preferred_alias` read in the title override, the trailing
joinphrase, and the configured-string-not-area-code return), and every
corpus figure that could be recomputed does reproduce off the tree.

The failures cluster in two shapes. Two measurements were asserted rather
than run and were destined for `docs/HISTORY.md` as rediscover-guards
(B2, M-h), which is the expensive place for a false record. Three
instruments were specified without their mechanism: the `defect_free`
derivation's input envelope and arithmetic (B1), the profile decoder's
parser (B4), and the ledger write-back's writer (M-g). Beside those sit
one vacuous checker (M-b), one assertion that fails as written (M-c), two
latent panics the boundary discipline exists to prevent (M-d, M-e), one
decoder type that breaks 76 case directories (B3), and a self-contradiction
about the `"Media"` sentinel (B5). None of it touches the ratified
rulings or either standing do-not-re-litigate record, and all of it is
fixable inside the plan document.
