# Rung 2, pass 2c planning inventory

Research report, 2026-09-03 (Sonnet planning agent), commissioned to
prepare pass 2c's plan: the real-catalog and adversarial strata, the
stock-config stratum, the #6358 alias-regression verification, the
file-level applied-tag derivation, the version-bump differential job
as a documented job, the quarterly re-capture chore doc, and the
known-oracle-bug ledger and adjudication artifact formats with the
first adjudication batch to Geoff. Scope is exactly the 2c paragraph
in `docs/plans/2026-09-02-rung-2a-cassette-substrate.md`
(lines 31-41). Findings below are organized (a) through (f) per the
dispatch; every claim cites a file path, and code claims were checked
against the shipped source, not assumed from prose.

## (a) Known-oracle-bug classes to seed the adversarial stratum

Primary sources:
`docs/research/2026-09-02-beets-evolution-lessons.md`
("U1, oracle-bug policy," section 5, and section 7's disposition
survey) and
`docs/research/2026-09-02-beets-matcher-mechanics.md`
(Q4's verdict-moving-key table, lines 444-451). Ruling 2
(`docs/research/2026-09-02-greenfield-server-rulings.md`,
item 2; restated in the build spec at
`docs/specs/2026-09-02-greenfield-server-build.md:712-717`)
is what makes these cases legal to author at all: tier one is
measured against release truth, and a case whose pinned-beets verdict
is demonstrably wrong gets its expected verdict set by adjudication
rather than by the oracle.

**1. The featuring-pattern ("ft") false-positive.** The
`SD_PATTERNS` featuring discount
(`beets/autotag/distance.py`, `string_dist`, weight table entry for
`feat`/`ft`/`featuring`) has carried no word boundary since its
September 2010 introduction (evolution-lessons commit `742ebdf69`);
any title word ending in "ft" followed by a space, period, or colon
("Draft ...", "Left: ...", "Gift ...") triggers the discount and can
make a genuinely different title/artist score near-identical to a
decoy. Partially fixed at 2.13.0 (`7b59604c5`); the symmetric
word-boundary fix (`19d091440`) is still unreleased at pinned 2.13.1
(evolution-lessons section 1, "Communication and regressions," and
section 5's U1 entry, which names this exact seed recipe: "titles
containing 'draft', 'left', 'gift' against near-miss candidates").
**Case shape:** a real or synthetic release whose title contains one
of those trigger words, paired with a genuinely different candidate
whose own title is unrelated except for sharing enough other tokens
that the featuring discount is what tips the distance under a
recommendation threshold. `corpus/README.md`'s family 4
(`strdist-*`) already proves the pipeline mechanics for the
*legitimate* `feat.`-strip case (`strdist-feat-strip`); the
adversarial version needs a title where the "ft" match is a false
positive, not a real featuring credit, and a candidate whose true
content differs enough that a human (or release-truth check) would
call the match wrong. The corresponding open-issue disposition
(evolution-lessons section 7, "String-heuristic discounts erasing
differentiators," #5868) also names remix-title near-duplicates as
adversarial-stratum material via the same `SD_PATTERNS` parenthetical
discount (weight 0.3).

**2. Non-Latin / pseudo-release / alias handling.** Ten-plus years of
open demand (evolution-lessons section 1, "Non-Latin scripts and
transliteration: never fixed structurally"; #1916 romaji open since
2016). The corpus's existing `non-latin-pseudo-release` smoke case
(`corpus/README.md`'s smoke table) proves the merge mechanism works
for one release; the adversarial weight belongs on cases designed to
find where alias/transliteration selection picks the *wrong* string,
not merely a correctly-merged one. #6358 (below) is this class's
sharpest current instance.

**3. Per-pair track-assignment rejection gaps.** #2981 (2018, open:
same track count with divergent titles yields a confidently wrong
per-track assignment) and #6078 (2025: tracks detected in wrong
order) — evolution-lessons section 7, "Forced total track
assignment." The assignment solve (`assign_items`, already exercised
structurally by pass 2b's family 1, `corpus/README.md` lines 106-124)
never rejects a pair; it always maps every input track onto some
candidate track. **Case shape:** an item set with the same track
count as the candidate, every item carrying a title unrelated to its
"correct" track (so title contributes no signal) but item lengths
close enough between adjacent tracks that the solver's optimum picks
a plausible-looking but wrong mapping. The album-level verdict can
still land `strong` while individual `AppliedFields.title` values are
wrong — the corpus's own frozen record makes the wrongness visible
per-track even though the recommendation doesn't flag it, which is
the sharpest evidence for why this class needs adversarial weight
rather than trusting the confident band alone.

**4. Threshold-boundary strictness.** #5458 (2024:
`strong_rec_thresh: 0` cannot mean exact-only because
`match.py`'s comparison is strict `<`). Small and mechanical; pass
2b's family 15 already parks cases at the boundary
(`rec-strong-medium-boundary`, `corpus/README.md` lines 592-598) and
`internal/harness/analysis`'s `proximity.json` ledger already tracks
strict-vs-inclusive comparison risk. This class does not need a new
adversarial case; it needs the existing proximity ledger carried
forward untouched (it is currently empty for the smoke corpus per
`corpus/README.md` line 725, and pass 2b's own boundary case is
already an acknowledged flag by construction).

### #6358, precisely, and how to verify it at pinned 2.13.1

**What it is**, per evolution-lessons section 7 ("Alias and
non-Latin artist selection") and section 5's U1 entry: issue #6358
(2026), "locale aliases ignored after an mbsync," suspected as a side
effect of beets dropping the `musicbrainzngs` library for an in-house
MB client in 2.6.0 (evolution-lessons section 4, "musicbrainzngs
churn"; the client replacement itself shipped a separate packaging
bug, #6351). Neither the evolution-lessons report nor any file in
this repo carries the issue body verbatim — a live `gh issue view
6358 --repo beetbox/beets` read is a cheap first step for the 2c plan
to do before Task authoring, since it will confirm whether the defect
is scoped to the `beet mbsync` command specifically or (more likely,
per the mechanism below) to the shared alias-selection code the
importer's own matching also calls.

**The mechanism it plausibly breaks:** `_preferred_alias`
(`beetsplug/musicbrainz.py:128-155` per
`docs/research/2026-09-02-beets-matcher-mechanics.md`
lines 446-451), gated by `import.languages` and
`import.ignored_alias_types` (`config_default.yaml:32,58`), rewrites
candidate titles, artist names, and release titles from the release's
`aliases` list *before any distance is computed*
(musicbrainz.py:158-162, 329-334, 443, 648). This is the same code
path `album_for_id`/the plugin's `AlbumInfo` construction uses for
both a fresh import lookup and an `mbsync` re-fetch, since both are
built from the same `inc=...+aliases+...` release response
(mechanics study Q5, lines 486-491, confirms `aliases` is already
part of `RELEASE_INCLUDES`, so every case this corpus already
captures carries the raw alias data; only the item-set profile needs
`import.languages` set to exercise the selection logic). It is
checked and confirmed verdict-moving by the mechanics study (Q4, line
446-451), which is exactly why an alias-selection defect could
silently rewrite a title or artist the year/string-distance
penalties then score against.

**How to verify within the harness (the falsifiable test the 2c plan
should run):** author a case against a real release whose
`artist-credit` carries a documented locale-specific alias in its MB
data (the `non-latin-pseudo-release` smoke case's own captured
`responses/*.json` is a candidate source to inspect first, since its
alias array is already frozen and committed); derive it twice, once
under a profile with `import.languages: []` (today's estate/stock
default — no alias preference, primary name used) and once under a
variant profile setting `import.languages` to the locale the alias
targets. If `_preferred_alias` is working, the second derivation's
`Candidates[].` title/artist strings (and `AppliedFields`) should
shift to the locale alias; if #6358's defect is real at 2.13.1, they
will not, and the divergence is directly observable in the frozen
`Expected` record without any new harness machinery — only a third,
narrowly-scoped profile variant (`corpus/profiles/` gains a
`languages`-setting sibling, or the adapter grows a
derivation-time-only config override, whichever the plan's Task 1
scopes) and one case authored against it. This is a `(b)
known-oracle-bug ledger candidate` per evolution-lessons' own
disposition (section 7: "if confirmed it is a textbook (b) ledger
entry, since conforming to it would degrade the estate's own
non-Latin material").

## (b) Real-catalog selection method

Target: ~80-150 cases (2a plan line 32; corpus report recommendation
section, `docs/research/2026-09-01-beets-conformance-corpus.md`
lines 23-28), live-capturable at 1 req/s, stressing the confident
band with genre/era/edition diversity, VA, box sets, non-Latin,
reissues.

**Selection axes** (drawn from the corpus report's own real-catalog
description plus the evolution-lessons "partially avoided" findings,
which name where beets' own record argues for stratum weight):

- **Genre/era spread**, to avoid the corpus silently over-indexing on
  whatever genre the smoke corpus (rock/pop, 1970s-2010s) already
  covers. Classical and compilation material get explicit
  weight per the uncertain-match-ux report's strongest finding
  (`docs/research/2026-09-01-uncertain-match-ux.md` line
  37): MB's classical coverage gaps are structural, not a
  tagging-quality problem, evidenced independently by beets issues
  #1665, #6244, #4882, #6598 (false negatives even with explicit
  MBIDs) — the real-catalog stratum should therefore include enough
  classical/compilation releases to observe whether pinned beets'
  *confident* band ever includes one of these structural misses,
  since a confident wrong match here is exactly what tier one exists
  to catch.
- **Edition/reissue diversity**: remasters and reissues where a
  release's own first-release-date diverges from its release date,
  which is the real-world analogue of pass 2b's synthetic family 7
  (`year-*`, `corpus/README.md` lines 322-357) and is where
  `original_date` is actually load-bearing on real MB data rather
  than a hand-specified `ReleaseSpec.original_date`.
- **Various Artists and box sets**: the smoke corpus already has one
  VA case (`various-artists`) and one multi-disc case
  (`multi-disc`); the real-catalog stratum should widen both —
  compilations with many contributing artists, and genuine
  multi-disc box sets (3+ discs), which pass 2b's synthetic family 6
  never attempted (its own cases are two-medium only,
  `corpus/README.md` line 289). A real box set also stresses U6's
  capture-breadth concern directly: a large multi-disc release's
  single `inc=`-maximal release response can be substantial, which
  is capture-tool wall time, not request count (see the arithmetic
  below).
- **Non-Latin scripts**: beyond the one pseudo-release smoke case,
  real releases in Japanese, Korean, Cyrillic, and Greek scripts,
  chosen to also serve the #6358 verification in (a) — a release
  whose artist carries a genuine locale alias is exactly what that
  verification needs, so the same captured release can serve both
  the real-catalog stratum's diversity requirement and the adversarial
  stratum's #6358 case, rather than requiring two separate captures.
- **Mistagged/partial-rip real material**: the corpus report's own
  recommendation (line 24) names "bootleg-vs-official duplicates,
  partial rips, and mistagged inputs" as real-catalog content;
  pass 2b's synthetic corpus already covers the *mechanism* for
  missing/extra tracks and wrong tags via `overrides`
  (`corpus/README.md`'s "Items are derived from the release by
  default, then perturbed," lines 950-976), so the real-catalog
  stratum's marginal value here is specifically real MB data's own
  noise (bootleg releases that exist in MB with sparse metadata,
  official reissues with near-duplicate catalog numbers) rather than
  reproducing the perturbation mechanism a third time.

**Practical method**: browse MusicBrainz's own web search and the
canonical-data / JSON-dump reference material named in the corpus
report (lines 60-66) to *select* candidate MBIDs — this research step
carries no rate-limit obligation, since it never calls the WS/2 API
directly — then capture only the chosen MBIDs through `harness
capture --mbid <release> --out corpus/cases/<id>`
(`corpus/README.md` "The capture tool," lines 859-899), which is the
only step actually bound by 1 req/s. This mirrors exactly how the
smoke corpus's seven live cases were chosen (per HISTORY.md's pass 2a
entry: real releases picked for "machinery coverage," here picked for
matcher-behavior coverage instead).

### Capture wall-time arithmetic

The only *live* precedent is the seven-case smoke corpus
(`corpus/README.md`'s smoke table); pass 2b's 61 structural cases are
synthetic (`harness synth`, no network, per `corpus/README.md`'s
"The synth tool" section) and are not evidence for capture throughput.
Response-file counts per live case (verified against the committed
tree, `ls corpus/cases/*/responses | wc -l` per case):

| case | responses (≈ requests) |
| --- | --- |
| `clean-match` | 1 |
| `multi-disc` | 1 |
| `various-artists` | 1 |
| `year-penalty` | 1 |
| `non-latin-pseudo-release` | 2 (the pseudo-release double fetch) |
| `should-ask` | 6 (1 ID lookup + 1 search + 4 pressings) |
| `should-reject` | 7 (1 ID lookup + 1 search + 5 pressings) |

19 requests across 7 cases. The two outliers (`should-ask`,
`should-reject`) are *deliberately corrupted* items that trigger
beets' search fallback and then the capture-derive loop's own
convergence discipline (`corpus/README.md`'s "A recorded fixed
point," lines 1048-1063): each pressing the search returns becomes a
new miss, consumed by a second `harness capture --case` invocation.
A real-catalog stratum optimized for the *confident band* (this
section's own design goal) should look far more like the other five
cases — a single strong ID-lookup match needs exactly one request —
with the pseudo-release/non-Latin subset needing two, and only a
deliberately edge-sitting minority (ambiguous editions, a release
whose search fallback the capture-derive loop discovers only after
the first attempt) reaching the `should-ask`/`should-reject` shape.

Weighted estimate: roughly 75% of cases at 1 request, 20% at 2
(pseudo-release or box-set follow-ups), 5% at the `should-ask` shape
(≈6 requests) ≈ **1.4 requests/case average**. At 1 req/s (first
request free, ~1s per request after):

- 80 cases × 1.4 ≈ 112 requests ≈ **~2 minutes** of pure network wait.
- 150 cases × 1.4 ≈ 210 requests ≈ **~3.5 minutes**.

**This number is not the real budget driver, and the 2c plan should
say so explicitly rather than sizing the task against it.** The 1
req/s ceiling is generous relative to per-case *selection and
authoring* cost: choosing a release that genuinely serves a
stratification axis, verifying via `harness freeze` that the
capture-derive loop converges (which can itself trigger a second or
third capture round, per the `should-ask`/`should-reject` precedent
above — each round pays its own network wait plus a full derive
cycle), writing the case's row into `corpus/README.md`'s inventory
table, and authoring its `claims.json`. Across 80-150 cases that
authoring loop, not the rate limit, is what turns the task into a
multi-hour effort; the plan should budget human/agent attention time
for authoring, and treat the ~2-4 minutes of network wait as noise
inside it.

## (c) The stock-config stratum's design (ruling 3)

**What `cassette.Case.Profile` supports as shipped**: a single
string field (`internal/harness/cassette/cassette.go:44`), required
non-empty by `LoadCase`'s validation (line 144-146), consumed once by
`oracle.Derive`/`renderProfile` (`internal/harness/oracle/oracle.go:175,
228-232`) to render exactly one profile per derivation. `Expected`
is likewise a single field on `Case` (cassette.go:48), written once
by `harness freeze` (`corpus/README.md`'s "The freeze tool," lines
1017-1046). **There is no multi-profile record inside one case
directory as shipped**: a case cannot carry both an estate-profile
and a stock-profile `Expected` under one `case.json`. The stock-config
stratum therefore needs **sibling case directories**, not a second
record field — e.g. `corpus/cases/<id>-stock/` alongside an existing
`corpus/cases/<id>/`, each self-contained (its own `case.json`,
`responses/`) per the format `corpus/README.md`'s Layout section
already documents (no cross-case reference mechanism exists in the
schema, so a sibling cannot point at another case's `responses/`
directory by reference — only by copying the response bytes, which
keeps every case dir the same self-contained shape `LoadCase`
already assumes).

**What actually diverges between estate and stock**, checked against
the shipped profile files and beets' own pinned defaults (not
assumed):

| key | `corpus/profiles/estate.yaml` | pinned beets 2.13.1 default (`tools/venv/.venv/.../beets/config_default.yaml`) | moves what |
| --- | --- | --- | --- |
| `original_date` | `yes` | `no` (line 103) | **Recommendation/distance** — album_info.year source for the year-penalty branch (family 7, `corpus/README.md` lines 322-357) |
| `per_disc_numbering` | `yes` | `no` (line 102) | **Recommendation/distance and AppliedFields** — multi-disc track distance/assignment, per the profile transform table's own claim (`corpus/README.md` line 833) |
| `none_rec_action` | `skip` | `ask` (line 34) | **Action only**, not Recommendation — quiet-mode mapping for a `none`-recommendation case (session.py, per the mechanics study Q4 line 456-459) |
| `quiet_fallback` | `skip` | `skip` (line 33) | **no divergence** — estate matches the default here |
| `timid` | `no` | `no` (line 27) | **no divergence** |
| `plugins` | `musicbrainz` | (not applicable; corpus forces single-source in both profiles per the design constraint) | n/a |
| `match:*` | absent (defaults) | defaults | **no divergence** — neither profile overrides thresholds or weights (`corpus/README.md` line 839, confirmed by the mechanics study's Q4 verdict-moving table) |

This narrows the stratum's *genuine* exposure surface to three
things: (1) real-catalog cases with a true original/release-date
divergence (reissues/remasters — already a selection axis in (b)),
(2) multi-disc real-catalog cases, and (3) any case landing on a
`none` recommendation, where Action (not Recommendation) differs —
irrelevant to tier one, which is about false *confident* matches, but
relevant if the plan wants stock-config Action parity documented too.

**Recommended stratum design**: not a blind full re-derivation of
every real-catalog case under stock. Mandatory inclusion: every
real-catalog case selected in (b) that is a reissue/remaster (axis
2) or a box-set/multi-disc release (axis 3), since those are the only
shapes where estate and stock can *actually* disagree on a verdict —
re-deriving them under stock is what proves ruling 3's "zero false
confident matches" claim is not vacuously true. Supplementary
coverage: a broader stratified sample of the remaining real-catalog
cases (ordinary confident matches, where estate and stock are
expected to agree byte-for-byte on Recommendation/distance), sized to
give ruling 3's claim breadth without doubling the whole real-catalog
stratum's authoring cost — the plan should pick a concrete count
(a range like 20-40 stock-sibling cases, roughly a quarter to a third
of the 80-150 real-catalog stratum, weighted toward the two
divergence-bearing axes) rather than leaving it open, since "the
stock-config stratum" as named in the 2a plan (line 34) implies a
bounded, deliberately-sized set, not an automatic mirror of the whole
real-catalog stratum.

**Mechanical layout**: a stock sibling copies its estate twin's
`Input` (JSON, trivial to copy) and its already-captured
`responses/*.json` verbatim, sets `profile: "stock"`, and runs
through `harness freeze`. Where stock's search-fallback behavior
diverges from estate's (possible whenever `original_date` or
`per_disc_numbering` moves the initial ID-lookup's Recommendation
across the strong/non-strong boundary — a stock derivation might need
a search request the estate capture never triggered, or vice versa),
the miss-file mechanism (`corpus/README.md`'s "The miss-file
mechanism," lines 900-915) handles it exactly as it does for any
other case: `harness capture --case <stock-sibling-dir>` fetches
whatever `misses.json` names.

## (d) File-level applied-tag derivation (rung 8's precondition)

**Why it is not an `Expected` field, restated from the source**: the
2a plan's Task 1 states this explicitly (2a plan lines 205-210):
`AppliedFields` is deliberately the *in-memory* post-apply field bag,
because the study verified the recommendation-and-apply half needs no
files, no importer, no library; "the file-level tag set (what
`item.write()` puts in a FLAC via mediafile) is explicitly NOT in
this record; it is pass 2c's deliverable and rung 8's stated
precondition." `cassette.Expected`'s schema is frozen at v2
(`corpus/README.md` line 654-661, the schema-version bump discipline)
and gains no new field for this — the same discipline `claims.json`
already established for a different reason applies here: "the record
schema is frozen, and a claim is a reading of a record rather than
part of one" (`corpus/README.md` line 681). File-level tags are a
*derived reading* of an `AppliedFields` record through a real write
path, not part of the oracle record itself.

**What the derivation needs, concretely:**

1. **Audio fixture generation for FLAC.** No pure-Go FLAC fixture
   generator exists in this repo; the salvage tier's precedent
   (admissible on merit per CLAUDE.md's salvage policy) is `ffmpeg`
   (a deterministic sine-wave source) piped through the real `flac`
   encoder, wrapped in `exec.CommandContext`:
   `_salvage/e2e/main_test.go:104-117` (`generateGoodFlac`, encoded
   once and reused via a package-level `goodFlacBytes` var) and
   `_salvage/internal/gate/gate_test.go:246-257` (`makeFlac`,
   `make8BitFlac`, parameterized by sample rate/bit depth). Both
   assume `ffmpeg` and `flac` are present on the workstation, which
   pass 2a/2b's own gate does not currently require — the 2c plan
   should state whether it adopts the same assumption (workstation-only,
   `make derive`-gated, never in `scripts/check.sh`) or vendors a
   static template FLAC byte blob instead, since a real derivation
   only needs a handful of representative fixtures generated once,
   not per-case synthesis.
2. **Copy-into-tempdir import discipline.** `oracle.Derive` already
   allocates a per-run scratch directory (`os.MkdirTemp`,
   `internal/harness/oracle/oracle.go:169-173`) that becomes the
   rendered profile's `directory`/`library` placeholders
   (`renderProfile(c.Profile, scratch, ...)`, line 175) — today
   unused for actual files, since `Input` is file-free at 2a
   (cassette.go:43's own comment). File-level derivation extends this
   existing scratch dir rather than inventing new plumbing: for each
   item that needs a file-backed derivation, copy the (once-generated)
   template FLAC bytes into a fresh path under `scratch`, point that
   item's `path` field at the copy, and let beets' own `item.write()`
   mutate the copy — never the shared template, and never a
   cross-item shared file, so tag writes from one item can never leak
   into another's read-back.
3. **mediafile read-back.** `mediafile` is already an installed venv
   dependency (`tools/venv/.venv/lib/python3.13/site-packages/mediafile`,
   confirmed present; it is beets' own tag-I/O library, pulled in
   transitively) but **is not currently pinned in
   `corpus/manifest.json`** (`Manifest` lists `Beets`, `Unidecode`,
   `Jellyfish`, `Lap`, `Numpy` only — `internal/harness/cassette/cassette.go:104-114`;
   the committed manifest confirms no `mediafile` key). Since the
   file-level tag set's exact shape (Vorbis-comment key names, N/T
   track-number formatting, date formatting) is mediafile's own
   surface, not beets', pass 2c's Task should add a `MediaFile`
   version field to `Manifest` and reconcile it in
   `oracle.(*Manifest).Reconcile` alongside the existing four. The
   read-back itself is independent verification, not a trust of
   `item.write()`'s return value: after writing, open the same file
   path fresh through `mediafile.MediaFile(path)` and read its tags
   back, since a write can silently coerce or reformat a value (a
   `year` int becoming a `date` string, a `track`/`tracktotal` pair
   becoming a Vorbis-comment `"N/T"` combined field) in ways the
   in-memory `AppliedFields` record never shows.
4. **Artifact form.** Following the `claims.json` precedent (sibling
   to the frozen record, not inside it) but scoped like the
   corpus-wide analyses (`corpus/analysis/*.json`, since this is a
   representative-subset deliverable rather than a per-case
   requirement — see below): a new `corpus/analysis/file_tags.json`
   (naming is the plan's own call, but should follow the ties/
   proximity/components ledger shape already established, one row
   per case id, gate-checked for coverage the same way) recording,
   per selected case, the read-back tag map and a pointer to which
   `AppliedFields` it was derived from. This keeps `cassette.Expected`
   untouched (schema stays v2) and reuses the exact "flag is not a
   failure, an unacknowledged flag is" discipline the other three
   analyses already carry (`corpus/README.md` line 721).

**Scope**: not every corpus case needs a file-level derivation — the
mechanism (write path, tag-name mapping, N/T formatting) is proven
once it is proven for the shapes that actually stress it. A
representative subset covering: a plain single-disc case
(`clean-match`), a multi-disc case with uneven split
(`media-uneven-split`, whose `corpus/README.md` row already documents
the exact `track`/`tracktotal` pair a file-level check would verify,
lines 319), a VA case with per-track artist credits (`va-exact`), a
non-Latin case (`non-latin-pseudo-release`), and a year/original-date
case (`year-reissue-ratio-penalty`) covers every `AppliedFields` shape
`corpus/README.md`'s families 6, 7, and the VA family already
document as interesting — on the order of 10-15 cases, not the full
corpus.

## (e) Adjudication artifact and known-oracle-bug ledger formats

**Design constraint, stated once and load-bearing for the whole
section**: this format is not single-use. The build spec's rung 4
(Shadow) names "the divergence log ... and the adjudication artifact"
as future infrastructure
(`docs/specs/2026-09-02-greenfield-server-build.md:550-551`)
that will record live shadow-run divergences once the Go matcher
exists. Pass 2c is originating the schema rung 4 will reuse for a
structurally identical problem (a human adjudicating a matcher
disagreement from a written artifact alone, without re-deriving
anything) — the 2c plan should design the row shape forward-compatible
with a shadow divergence, not narrowly for a pre-launch known-bug
case.

**What a divergence row needs, so Geoff can adjudicate from the
artifact alone**, synthesized from three sources: the corpus's own
existing disposition-ledger idiom (`ties.json`'s `disposition` +
`reason` per tied case, and `proximity.json`'s flag-plus-reason
requirement, both in `corpus/README.md`'s "The analyses" section,
lines 704-767); the uncertain-match-ux report's Option C precedent
(an owner decision queue is the only validated multi-user pattern for
this exact judgment call, `docs/research/2026-09-01-uncertain-match-ux.md`
line 15); and that report's own strongest finding about what makes a
divergence genuinely hard to adjudicate (structural MB coverage gaps
versus tagging-quality problems look identical from a bare verdict,
line 19) — which is the argument for why a row needs *evidence*, not
just a verdict pair:

- **Case/release identity**: the corpus case id (or, for a live
  shadow divergence, the album/MBID) — the join key back to the
  frozen fixture.
- **The oracle's own record**: recommendation, action, distance, and
  the full penalty breakdown for the best candidate (already exactly
  `cassette.Expected`'s shape, so this can be a direct reference
  rather than a re-transcription).
- **What is suspected wrong and why**: the bug-class citation (an
  issue number, or an evolution-lessons finding id like "U1/ft-pattern"),
  stated in prose, not just a link — matching the corpus's own
  convention of writing the *mechanism* in every family's
  `corpus/README.md` row rather than only a test-file pointer.
- **Release truth**: what the correct verdict is, with its own
  citation (the actual correct MusicBrainz release, or an explicit
  statement that the input material has no correct MB entry at all,
  per the uncertain-match-ux report's MB-style-guide finding on
  bootlegs, line 17) — this is what ruling 2 requires ("measured
  against release truth") and what makes the row adjudicable without
  re-deriving anything.
- **Disposition + reason**, the exact two-field shape `ties.json`
  already uses: adjudicated outcome (for the known-oracle-bug ledger,
  something like `oracle-wrong` with the corrected expected verdict
  attached, or `oracle-correct` if the crafted case failed to
  reproduce the suspected defect) plus the human-readable reason a
  future reader needs, never re-derived from scratch.
- **Adjudicator and date**: who ruled it and when, since (per the
  build spec's improvement-adoption discipline, lines 100-110) an
  adjudication is a recorded decision with written reasoning, not an
  implicit one.

**Known-oracle-bug ledger shape**, specializing the above for U1
specifically: one row per known-defect *hypothesis* (not one row per
corpus case, since a hypothesis may be tested by more than one case,
or by none if the crafted case never reproduces it), carrying the bug
class, the case id(s) authored to probe it, whether the case actually
reproduced the defect (a hypothesis that *doesn't* reproduce is still
a recorded row — matching the corpus's "an unacknowledged flag is a
failure" discipline, so a checked-and-cleared hypothesis is not
silently dropped), the adjudicated expected verdict where the defect
is confirmed, and the disposition/reason/adjudicator/date fields
above. **First adjudication batch to Geoff**: the plan's own Task
structure should batch every row this pass produces (the #6358
verification, the ft-pattern adversarial cases, any real-catalog
case that lands unexpectedly outside the confident band during
authoring) into one combined review point, per CLAUDE.md's "batch
mid-execution judgment calls into one combined question" — not one
ask per row.

## (f) Derive-time scaling arithmetic

Baseline, from HISTORY.md's pass 2b entry (the pass's own stated
"2c's budget line"): 68 committed cases, two independent uncached
`make derive` runs (`-count=1`, since a cached run "is a cache hit
and proves nothing" per the pass's own carried lesson), each taking
~2m13s total: **corpus derive 70.5s, ties derive 54s, component
derive 0.7s**.

**What scales with case count and what doesn't**, checked against
each tool's own description in `corpus/README.md`:

- **Corpus derive** (`oracle.Derive`, one shim subprocess per case)
  scales linearly with case count — each case is an independent
  `PYTHONHASHSEED=0` Python process launch plus a `tag_album`-equivalent
  call, and per-case Python startup/import cost is the dominant fixed
  overhead per case (`corpus/README.md`'s "Oracle process contract,"
  2a plan lines 97-106). 70.5s / 68 cases ≈ **1.04s/case**.
- **Ties derive** (`derive_ties.py`, "The analyses" section, lines
  730-767) likewise runs once per case (fetches the best candidate,
  builds and solves the cost matrix, re-solves per assigned pair) —
  also linear in case count, with per-pair re-solves adding a small,
  track-count-dependent constant per case that the corpus's own
  case sizes (2-14 tracks) keep small. 54s / 68 ≈ **0.79s/case**.
- **Component derive** (`derive_component.py`, "The analyses"
  section, lines 768-793) is **flat, not case-count-dependent**: it
  reads literal argument tuples mirroring `test_distance.py`'s fixed
  parametrize tables, independent of how many corpus cases exist.
  0.7s total regardless of corpus size.

**Projection to ~230 cases** (68 existing + roughly 80-150 real-catalog
+ 15-30 adversarial + a 20-40-case stock-config subset per (c),
landing in the low-to-mid 200s depending on the plan's exact sizing
choices):

| tool | 68-case measured | 230-case projected (linear) |
| --- | --- | --- |
| corpus derive | 70.5s | 230/68 × 70.5s ≈ **238s (~4m)** |
| ties derive | 54s | 230/68 × 54s ≈ **183s (~3m)** |
| component derive | 0.7s | **0.7s (flat)** |
| **total** | ~2m13s | **≈ 7 minutes** |

Two additions the plan should price separately rather than folding
into the linear projection: the file-level applied-tag subset from
(d) (10-15 cases, each paying an `ffmpeg`+`flac` fixture write plus a
real file write/read-back — likely 1-3s of *additional* per-case
work on top of the normal derive cost, so on the order of **+15-45s**
total, not a rerouting of the whole corpus); and the determinism
proof, which stays flat regardless of corpus size since it re-derives
only one fixed smoke-shaped case twice (2a plan, "The determinism
proof," lines 361-365) — it does not scale with the corpus at all,
so it stays a non-factor in this arithmetic even at 230 cases.

## Sources

- `docs/plans/2026-09-02-rung-2a-cassette-substrate.md`
  (the 2c scope paragraph, lines 31-41; the frozen `cassette`
  interface block, lines 135-238; the determinism proof, lines
  361-369).
- `docs/research/2026-09-01-beets-conformance-corpus.md`
  (stratum design and sizing, lines 9-32).
- `docs/specs/2026-09-02-greenfield-server-build.md`
  (the launch bar's two-tier gate, lines 51-111; ladder entry 2, lines
  489-526; ladder entry 4's divergence-log/adjudication-artifact
  mention, lines 548-556; the rulings, lines 696-722).
- `docs/research/2026-09-02-greenfield-server-rulings.md`
  (rulings 1-3 in full).
- `docs/research/2026-09-02-beets-evolution-lessons.md`
  (section 1's featuring-pattern and non-Latin history; section 5's
  U1/U6/U7/U8; section 7's disposition survey and #6358 entry).
- `docs/research/2026-09-02-beets-matcher-mechanics.md`
  (Q4's verdict-moving-key table, lines 415-474; Q5's HTTP seam and
  `inc=` includes, lines 476-519).
- `docs/research/2026-09-01-uncertain-match-ux.md` (the
  four-option spectrum and the classical/compilation structural gap
  finding).
- `docs/HISTORY.md` (pass 2a and 2b entries, including the derive-time
  budget line and the carried-forward 2c inventory note).
- `corpus/README.md` (layout, canonicalization, the analyses, the
  capture/synth/freeze tools, the config-profile transform table).
- `corpus/manifest.json`, `corpus/profiles/estate.yaml`,
  `corpus/profiles/stock.yaml`, `internal/harness/cassette/cassette.go`,
  `internal/harness/oracle/oracle.go` (schema and code read directly,
  not inferred from prose).
- `tools/venv/.venv/lib/python3.13/site-packages/beets/config_default.yaml`
  (pinned-beets default values for `original_date`,
  `per_disc_numbering`, `none_rec_action`, `quiet_fallback`, `timid`,
  checked directly against `estate.yaml`).
- `_salvage/e2e/main_test.go:99-134`, `_salvage/internal/gate/gate_test.go:240-272`
  (the FLAC-fixture-generation precedent for (d), admitted on merit
  per the salvage tier's adoption policy).
