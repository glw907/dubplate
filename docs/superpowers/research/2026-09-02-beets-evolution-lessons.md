# beets evolution lessons for the greenfield server

Research report, 2026-09-02 (research agent; full-history clone of
beetbox/beets at master a750105a6, tags 1.0b1 through v2.13.1, plus
docs/changelog.rst, the GitHub issue tracker via `gh`, the beets blog,
and targeted web search). Commissioned by the greenfield dubplate-server
initiative to study how beets grew over eighteen years and what its
growing pains imply for a from-scratch reimplementation. The sibling
snapshot study of the matcher's mechanics at the pinned tag is out of
scope here; this report covers the arc, the recurring pain, and the
mapping onto the ratified design. Claims are cited to commits, changelog
versions, or issues; anything unverifiable was left out.

## 1. The tagging machinery's arc

beets began in May 2008 (first commits c1ed60af9, ee7bb4b9e). The
matcher's ancestry runs in five phases.

**2009-2010: origins.** The MusicBrainz query layer and
`likely_metadata` land in April 2009 (e2d6af6ac, 155c9f730). Track
assignment is an assignment-problem solve from the start: the "matching
track orderer" using the munkres Hungarian-algorithm library arrives
November 2009 (a6bb3d088). The "smart" string distance, the heuristic
pattern table that discounts "the", "(EP)", parentheticals, and
featuring suffixes, lands September 2010 (742ebdf69). Its original
featuring pattern, `[\[\(]?(featuring|feat|ft)[\. :]`, carried no word
boundaries; that detail matters below.

**2011: the NGS rewrite and the source seam.** 1.0b11 (December 2011)
transitions the tagger to MusicBrainz's Next Generation Schema via
python-musicbrainzngs, with the bindings vendored into beets at first
(changelog 1.0b11). The same season, an internal refactor "abstract[s]
source from matching logic" and unifies two styles of metadata search
(ee78391f4, 064a0432f, de2ee7e44, Oct 2011), creating `hooks.py`. This
is the seam that later allowed Discogs and Beatport candidates, and
fifteen years later the plugin extraction.

**2013: the distance model takes its current shape.** Two releases fix
the shape the matcher still has at 2.13.1:

- 1.1b1 (January 2013) makes the recommendation thresholds
  configurable (9e6b1db02): `strong_rec_thresh: 0.04`,
  `medium_rec_thresh: 0.25`, `rec_gap_thresh: 0.25`.
- 1.2.0 (June 2013) is the big rework, motivated explicitly by the new
  Discogs/Beatport sources needing disambiguation "between similar
  releases in the same MusicBrainz release group" (changelog 1.2.0):
  the `Distance` object with per-field weighted penalties (e6ac8e164),
  many more fields (disctotal, label, catalognum, country,
  albumdisambig; a5151bb32), `preferred` media/country/original-year,
  `ignored`, per-penalty `max_rec` overrides, and the weights table
  moved "to an undocumented configuration setting" (feb1d8631,
  deliberately undocumented at birth).

**2013-2024: near-total numeric stability.** The three thresholds have
the same values in `config_default.yaml` at master as at their January
2013 introduction. The weight table is unchanged in values since
mid-2013; every later edit is additive (a `source` penalty when plugin
sources arrived) or a rename (`source` to `data_source`, 60e0eef-era,
2.5.0). The `distance_weights` table was only properly documented in
2025 (issue #6081). Structural work in the gap was modest: Levenshtein
moved to the jellyfish library for speed in 1.3.15 (:bug:`1389`),
explicit sorting of candidates by distance in 2016 (41fb0d928),
`get_distance` modularized toward plugins in 2019 (8010488f3,
ed80e915f).

**2024-2026: the 2.x decomposition.** After 2.0.0 (May 2024) the
matcher's packaging churns rapidly: munkres replaced with lapjv for
track assignment (420117b59, December 2024); `mb.py` moved to
`beetsplug/` (e6e610a3e, 2.3.0); the MusicBrainz tagger made a real
plugin in 2.4.0 (changelog :bug:`2686` :bug:`4605`); `Distance` split
into `beets/autotag/distance.py` (adbd50b23, 2.4.0-era); plugin-owned
`track_distance`/`album_distance` removed and source penalties pulled
into core in 2.6.0 because the plugin-side computation was wrong
(:bug:`6066`); a `Source` abstraction refactor in mid-2026 (97904e7e1).

**Communication and regressions.** Behavior changes were communicated
through changelog prose only. There was never a verdict-level
regression corpus; the autotag tests construct objects inline and no
cassette infrastructure exists (confirmed independently by the
conformance-corpus report). Consequences, all found by user reports:

- The 1.2.0 scoring rework shipped with a qualitative narrative and no
  way for a user to predict which of their verdicts would flip.
- The 2.4.0 plugin extraction silently removed MusicBrainz as a source
  for any user with a customized `plugins:` list; the changelog warned,
  the runtime gave no signal, and users reported "no matches from MB"
  months later (#6229, December 2025, resolved as configuration).
- The extraction era shipped real matcher regressions: wrong
  data-source penalty application (#6066, fixed 2.6.0), only the first
  candidate returned when one ID exists in multiple sources (#6178),
  cross-source candidates treated as duplicates of each other (#6181;
  both fixed 2.7.0, March 2026), plus a 2.4.0 import-logfile
  regression fixed in 2.6.x and a 2.8.0 Deezer selection regression
  fixed in 2.9.0 (changelog).
- The featuring-pattern bug is the starkest quality finding: from
  September 2010, any title word ending in "ft" followed by a space,
  period, or colon ("Draft ...", "Left: ...") triggered the featuring
  discount and made "genuinely different titles/artists score as
  near-identical matches" (changelog, Unreleased section at master).
  It was partially fixed in v2.13.0 (7b59604c5) and the symmetric
  word-boundary fix (19d091440) was still unreleased at 2.13.1. A
  false-match-pressure bug lived in the scorer for sixteen years
  because nothing adversarial ever probed it.

## 2. Recurring bug classes in matching

From the changelog and tracker, ordered roughly by how structurally
they were resolved.

**Reissue and pressing ambiguity: fixed structurally once.** The 1.2.0
penalty rework (label, catalognum, country, albumdisambig, album_id
weight 5.0, `preferred` settings) is the one time beets answered a
complaint class with a redesign. Residual complaints continued at the
preference layer (#3564, 2020: country preference ignored, weaker
match picked) but the class largely held.

**Multi-disc albums: patched for fifteen years, never structural.**
The mechanism has always been a directory-name heuristic (1.0b11,
December 2011). Patches: per-disc track numbering tolerance (1.0b13,
2012), `per_disc_numbering` display handling (1.2.0), data-track and
video-track ignoring (2018, 1701aefe5 and d325bceb1), the matcher
ignoring medium number entirely, breaking hybrid SACDs and releases
with duplicate tracks (#5148, fixed 2.2.0, November 2024), and disc
markers extended to "cassette", "digital media", "vinyl" only in
2.12.0 (#6059, 2026). The heuristic itself was never replaced.

**Non-Latin scripts and transliteration: never fixed structurally.**
Following MusicBrainz pseudo-releases was filed April 2014 (#654) and
closed June 2023, shipping in 2.0.0 nine years later. Romaji
transliteration (#1916) has been open since March 2016. The new
mbpseudo plugin then generated its own 2026 bug tail (#6339, #6448,
#6652, #6441). Ten-plus years of demand, still churning. This class
deserves the adversarial weight the corpus plan already gives it.

**Various artists: patched early, quiet since.** No penalty for
"Various Artists" track artists (#251, 1.0b12-era, 2012), albumartist
used when autotagging VA albums (1.0b14, 2012). Largely settled by
those patches.

**Track assignment: stable algorithm, churning dependency.** The
Hungarian solve has been the design since 2009. The pain was the
library: multiple incompatible munkres versions (#3168, 2019), then
wholesale replacement by lapjv (420117b59, 2024). User-visible
assignment complaints exist (#4388: no way to override a wrong track
assignment) but the class is small.

**Threshold and recommendation tuning: absorbed by config, defaults
frozen.** Complaints were answered with knobs (timid mode, `max_rec`,
`quiet_fallback`, `ignored`, `required`), never with retuned defaults;
0.04/0.25/0.25 have survived thirteen years. Even knob bugs recurred:
the `ignored` setting "was sometimes, well, ignored" (#1487, 1.3.15).

**String-distance heuristics: latent and adversarially untested.** The
"ft" bug above; also "and"/"&" equivalence (1.0b-era), "the"-prefix
discounts, parenthetical discounts. Each is a hand-tuned regex with
weight; none had adversarial tests until the 2026 fix. A class this
old surfacing a false-positive generator in 2026 says the confident
band was never under systematic pressure.

## 3. Structural growing pains beyond the matcher

**Plugin API sprawl.** 82 plugins live in-tree at master; the tracker
maintains a label per plugin. The changelogs of every release since
1.2 are dominated by plugin fixes, and scraper plugins (lyrics above
all) break continuously; 2026 alone has lyrics fixes in most releases
and an unthrottled-scraper bug (#6728). The cost was carrying every
integration's maintenance forever, because in-tree plugins cannot be
abandoned individually. What would have avoided it: a process boundary
and a smaller blessed core, which is exactly the fork dubplate settled.

**Config system.** The 1.1 rewrite (January 2013) replaced ad-hoc
config with a YAML system built on a bespoke library (confit, later
renamed Confuse), which beets then had to maintain as a separate
project; it was extracted as an external dependency only in 1.5.0
(August 2021). Building a config library to serve one application cost
a second repo's upkeep for a decade.

**Import pipeline concurrency.** The pipeline is coroutine stages with
optional one-thread-per-stage parallelism (7600a8c0a, April 2011).
Threads plus SQLite produced the project's worst bug saga: "database
is locked" crashes persisted "across eleven releases over nearly two
years" because the trigger (SQLite compiled without usleep, sleeping a
full second per lock retry) "only appeared on users' machines, never
the developer's" (beets blog, "the SQLite lock timeout nightmare",
2012). The fix was a global one-writer mutex with a per-thread
transaction stack. Deadlock and nondeterministic-crash fixes recur in
the 1.0b-era changelog (lines noting "rare deadlock when finishing the
import pipeline", multithreaded SQLite crashes). Lesson: concurrency
that exists for pipelining, layered over a single-writer store, must
be designed around the store's one-writer truth from day one.

**Database schema and paths.** Flexible attributes (1.3.0, September
2013; beets.io/blog/flexattr.html) solved schema rigidity and enabled
the plugin ecosystem to persist data without migrations; the cost was
semantic surprises at the edges (unknown-field queries flipped from
match-everything to match-nothing at introduction; uppercase flexattr
names were unqueryable until an Unreleased 2026 fix; a plugin field
stored as flexattr crashed `beet update`, #5580). Paths are stored as
bytes because "on Unixes, paths are fundamentally bytes" and the
platform lies about encoding (beets.io/blog/paths.html, 2016); the
whole codebase pays `bytestring_path`/`syspath`/`displayable_path`
boilerplate forever. Library-relative path storage, requested in issue
#133, shipped in 2.10.0 (April 2026), roughly fourteen years later,
with an automatic first-run migration that then needed a follow-up fix
for hand-edited TEXT rows (#6561).

**Python 2 to 3.** Sampson wrote about the transition in 2012
(beets.io/blog/py3k.html); alpha py3 support arrived in 1.4 (November
2016) with the changelog joke that it "may replace all your music with
Limp Bizkit"; py2 was dropped only in 1.6.0 (November 2021). Nine
years from first consideration to done, five of them dual-runtime via
six, entangled with the bytes-path discipline. A single-language,
single-toolchain rewrite dodges the class entirely, though Go version
migrations are not free either.

**The MusicBrainz extraction.** "Always prefer a different metadata
source over MusicBrainz" was filed in 2017 (#2686) and closed in 2025;
the extraction landed across 2.3.0 (file move), 2.4.0 (real plugin),
and 2.6.0 (source penalties pulled back into core after plugin-side
computation proved wrong, #6066). Untangling a privileged data source
after fifteen years of entanglement took three releases and shipped at
least three matcher regressions (section 1). The design choice that
would have avoided it is the one the 2011 refactor almost made: a real
source seam with no privileged caller, held from birth.

**Maintainer capacity.** Release history tells it plainly: 1.4.9 (May
2019) to 1.5.0 (August 2021) is a two-year gap; 1.6.0 (November 2021)
announced "our first experiment with time-based releases... every 3
months" and the next release, 2.0.0, took two and a half years (May
2024). Authorship shifted from Sampson (615 commits in 2013, zero in
2025) to a new lead (Šarūnas Nejus, 1632 commits since 2024). The
project survived the handoff, and the 2.x revitalization is real
(fifteen releases in ~26 months), yet much of the 2.x work is paying
down the coupling debts above rather than adding capability.

## 4. External pressure: MusicBrainz and dependencies

**The NGS transition (2011)** forced a rewrite of the tagger's whole
data layer in one release (1.0b11), and beets vendored the
musicbrainzngs bindings to ship it. A hosted web service can and did
change schema underneath the tagger.

**musicbrainzngs churn.** beets repeatedly depended on unreleased
upstream: 1.3.15 required "a new version of the python-musicbrainzngs
library that is not yet released", 1.4.2-era temporarily bundled a fix
pending release (:bug:`506`), 1.5-era pointed at unmerged PR #266. The
library then went unmaintained for four years (discussion #4660) and
beets replaced it with an in-house lightweight client in 2.6.0
(February 2026); the replacement itself shipped a packaging bug
(missing requests_ratelimiter module, #6351). Sixteen years of the
tagger's most important dependency being someone else's release
schedule.

**Rate limiting is the stable part.** The anonymous 1 request/second
policy has held for well over a decade (musicbrainz.org rate-limiting
doc; complaint threads from 2016's #1816 to 2026's #6728 are about
clients exceeding it, never about the policy moving). HTTPS became the
default in 1.5.0 (a46c97538, 2021); 429 handling was patched in the
1.6.0 era (:bug:`4370`).

**Response surface grows.** Works fetching arrived in 1.5.0, genres
in 2020 (a79b239d4), pseudo-release/transliteration relations in
2.0.0, external IDs made configurable in 2023 (fd4cecc29). MB
responses accrete fields as the tagger learns to want them.

Implications for dubplate's corpus discipline: the frozen-fixture
design with two independent freshness axes is validated by this
history, since every class of external change beets suffered (schema
transition, data drift, client-library death) maps onto either the
re-capture axis or the version-bump axis. The stable 1 rps policy
means capture-tool throughput assumptions will hold. The growing
response surface is the one under-priced axis; see U6.

## 5. Lessons mapped to dubplate

Graded against the organizer design spec (2026-09-01) and greenfield
build spec (2026-09-02). "Avoided" names the ratified decision that
prices the pain; "unpriced" states a falsifiable implication aimed at
a named rung.

**Avoided by ratified decisions.**

- Verdict regressions found by user reports (2.4.0-2.8.0 era):
  avoided by the conformance corpus, the frozen-verdict fast gate, and
  the whole-build shadow (build spec, launch bar 2-3). This study is
  direct evidence for that ratification: beets shipped at least three
  matcher regressions in twelve months while decomposing exactly the
  seam dubplate is reimplementing.
- Plugin API sprawl and scraper rot: avoided by the settled
  extensibility fork (built-ins, lifecycle hooks, client API; scrapers
  refused). beets' 82 in-tree plugins and perpetual lyrics tail are
  the counterfactual.
- Privileged-source entanglement (#2686, fifteen years): avoided by
  the tree shape, where `engine/matcher` and `engine/musicbrainz` are
  separate packages from rung 1 and the streamer/auth/storage seams
  follow the same rule.
- Dependency-churn exposure (musicbrainzngs, munkres): avoided by the
  pure-Go no-cgo fork and the in-house MB client with cassette replay.
- MB service change and data drift: avoided by the two-axis harness
  (version-bump differential vs quarterly re-capture) and the
  registered client identity at rung 2.
- Bespoke config library: avoided by `internal/config` as application
  code with a universal wall, not a reusable library.
- SQLite-under-threads crashes: largely avoided by single-executor
  discipline today and the flagged concurrency decision at rung 9;
  see U3 for the store-side residue.

**Partially avoided.**

- Threshold ossification: dubplate reproduces beets' full tunable
  surface and adjudicates post-swap divergence under the success bar,
  which is a policy beets never had. What thirteen frozen years of
  0.04/0.25/0.25 add: expect the estate's own defaults to become
  load-bearing the day the family system launches, so any future
  default change needs the corpus rerun plus a HISTORY entry, and the
  spec already gives both mechanisms.
- Multi-disc handling: the intake product shape (album batch as the
  verdict unit, contributor uploads one album) sidesteps beets'
  directory-name heuristic, yet rung 9's layout and rung 2's corpus
  still meet medium numbering. The corpus plan's multi-disc stratum
  covers it; the #5148 hybrid-SACD case (duplicate track titles across
  media) is a specific shape worth porting.
- Non-Latin material: the corpus already carries adversarial and
  migrator weight here. beets' record (nine years to pseudo-release
  support, a decade-open transliteration request, a 2026 bug tail)
  argues this stratum should be the last one trimmed if the corpus is
  ever squeezed.

**Unpriced lessons.** Each is a concrete implication for a named rung.

- **U1, oracle-bug policy (rung 2 spec; success-bar section).** Pinned
  beets 2.13.1 contains known false-match-pressure defects: the
  featuring-pattern fix is only partial at that tag (7b59604c5 in
  2.13.0; the symmetric-boundary fix 19d091440 is post-2.13.1).
  Conforming to the oracle can therefore manufacture exactly the
  tier-one failure the bar forbids. Implication: rung 2's plan defines
  a known-oracle-bug ledger, sibling to the known-acceptable-divergence
  artifact, where a case whose pinned-beets verdict is demonstrably
  wrong gets its expected verdict set by adjudication rather than by
  the oracle, and the corpus adds SD_PATTERNS adversarial cases
  (titles containing "draft", "left", "gift" against near-miss
  candidates). Falsifiable: craft such a case, run pinned beets; if
  beets confidently matches the wrong release and the harness would
  have graded Go "conformant" for copying it, the ledger is needed.
- **U2, assignment tie-breaking (rung 3 spec; rung 2 corpus).** beets
  replaced munkres with lapjv in late 2024 (420117b59); 2.13.1
  depends on `lap>=0.5.12`. Equally-optimal assignments can differ
  between solvers, and the Go matcher must reimplement the assignment
  solve from scratch. The spec's stable-ordering clause covers
  candidate lists, and says nothing about within-album track
  assignment ties. Implication: rung 3 pins tie-breaking to reproduce
  pinned-beets lapjv output, and rung 2's structural stratum includes
  degenerate equal-cost assignment cases (duplicate titles, identical
  lengths). Falsifiable: a corpus case with a tied cost matrix where
  two optimal assignments produce different track mappings.
- **U3, store path representation (rung 6 spec).** The design prices
  the migrator's BLOB rule and says nothing about the greenfield
  store's own path encoding. beets took fourteen years to reach
  library-relative storage (#133, 2.10.0) and its auto-migration then
  broke on edge rows (#6561). Implication: rung 6's schema stores
  library-relative paths from version one and states a byte-preserving
  encoding policy; rung 7's intake spec states explicitly whether
  non-UTF8 filenames are admitted, normalized, or rejected at the
  gate, since a server that controls intake can retire the class
  beets had to engineer around, while the migrator must still accept
  the estate's existing bytes.
- **U4, built-in metadata storage (rung 6 spec, consumed by rung
  10).** Flexible attributes are what let beets' ecosystem persist
  per-track data without schema migrations; dubplate's built-ins
  (ReplayGain, fingerprints, genres) will each need per-track
  storage, and the store rung's migrations-on-open mechanism makes
  typed columns viable. The choice between typed columns per built-in
  and a flexattr-like side table is currently unmade. Implication:
  rung 6's spec names the mechanism before rung 10 builds against it,
  so adding a built-in never becomes either an unplanned migration or
  an untyped attribute bag by default.
- **U5, loud config-compat breaks (rung 1, config core).** beets
  2.4.0's extraction silently degraded configured users: a customized
  `plugins:` list lost MusicBrainz with no runtime signal, and reports
  arrived months later (#6229). The changelog was the only defense.
  Implication: the rung-1 config wall treats unknown and superseded
  keys as load errors, and any future key rename ships with explicit
  old-name detection that names the replacement, so a semantic config
  change can never fail silently. Falsifiable: feed the server a
  config using a retired key; launch must fail with the migration
  message, never start degraded.
- **U6, capture breadth (rung 2, capture tool).** MB's response
  surface accretes fields (works, genres, pseudo-release relations,
  external IDs), and beets grew to consume them release by release.
  Fixtures frozen with only the fields today's matcher reads would
  force a full re-capture the first time a matcher improvement wants
  another field. Implication: the capture tool records the maximal
  practical `inc=` response per release rather than the minimal
  matcher view, and the fixture format stores the raw response so
  fixture readers, never the fixtures, decide which fields exist.
- **U7, beets-3.0 break resilience in the differential job (rung 2,
  harness).** beets' one active plan is v3.0.0, a
  compatibility-shedding release: the legacy plugin layer is
  TODO-marked for removal (beets/plugins.py, metadata_plugins.py),
  docs promise "Legacy support will be removed in beets v3.0.0"
  (docs/dev/plugins/autotagger.rst), and the matcher's orchestration
  is still moving (the 2026 Source refactor, open matcher-internal
  refactor issues). The standing version-bump differential job drives
  pinned beets as a subprocess with an estate config; a major version
  that sheds compatibility can break the invocation contract (config
  keys, plugin enablement, CLI behavior) independently of any verdict
  change. Implication: the harness rung isolates everything that
  knows how to invoke beets (config rendering, process entry, output
  parsing) behind one adapter, so a beets major bump costs an adapter
  update and an adjudication pass, never a harness rewrite.
  Falsifiable: when the differential is pointed at a beets 3.x, only
  the adapter package should need edits.

One meta-lesson needs no rung: beets' worst defects were invisible to
its own tests for years (eleven releases of lock timeouts, sixteen
years of the "ft" discount) and surfaced only through environmental
diversity and adversarial accident. The shadow rung is dubplate's
manufactured version of that accident, and the calendar time it
accumulates is the cheapest evidence the initiative buys.

## 6. Roadmap and stated future plans

beets has no roadmap artifact. There is no ROADMAP, GOALS, or vision
file in the repo, `docs/dev/` is developer reference only, CONTRIBUTING
names no priorities, the blog's last post is a 2022 walkthrough, and
the changelog's Unreleased section at master is an ordinary dev
changelog (a `duplicate_action: upgrade` feature, bug fixes), not a
plan. What planning exists lives in four scattered artifacts of very
different freshness.

**The one live plan: v3.0.0 as a compatibility-shedding release.** The
3.0.0 milestone is the only actively updated one (June 2026), and its
contents are deprecation removals: retiring the deprecated
acousticbrainz, absubmit, and beatport plugins (PR #6778), removing
`format_raw_length` (#6018), and dropping a mediafile shim (PR #6016).
The code carries matching intent: `beets/plugins.py` marks its legacy
plugin compatibility layer "TODO: Remove in v3.0.0" in three places,
`metadata_plugins.py` plans a strict `MetadataSourcePlugin` isinstance
check for 3.0, and the plugin-developer docs state that "Legacy
support will be removed in **beets v3.0.0**"
(docs/dev/plugins/autotagger.rst). Read together, the maintainers'
own stated priority is finishing the metadata-source decoupling and
shedding the compatibility shims it required, which is a direct
self-diagnosis that the legacy plugin surface is the project's
heaviest debt. That corroborates the plugin-sprawl finding in section
3 and the settled extensibility fork, and it is what motivates U7:
dubplate's differential harness will meet this planned break.

**Stale milestones as capacity evidence.** The other three open
milestones are wishlists whose contents predate the current
maintainership: "MediaFile++" holds tag-mapping and custom-field
issues filed 2013-2017 (#350, #572, #656, #2621); "Modularise beets"
holds a single 2015 Sampson issue, "Finish disentangling dbcore from
beets globals" (#1531), untouched since 2019; "Modern MPD support" is
a 2019 batch of bpd protocol items. The `refactor` label runs the
same way: "Use pathlib everywhere" (#1409, 2015), "Replace optparse
with Click" (#1442, 2015), "Refactor ImportTask" (#704, 2014) all
remain open, alongside newer maintainer-filed items (pytest
migration #5361 and SQL-only sorting #5318 from 2024, matcher-internal
cleanup #6920 from August 2026). A decade-old open refactor list is
intent outliving capacity, the same signal as the 1.6.0 quarterly
release promise (section 3), and it argues that dubplate should trust
beets' shipped behavior, never its stated direction, when deciding
what to conform to.

**The 2018 wiki page as the last architectural self-assessment.** The
GitHub wiki's Refactoring page, last edited by Sampson in April 2018,
names the pains the founder saw: album-level field consistency under
`update`, weak typing on Item fields ("confusing None values",
`modify` failing on non-string fields via SQLite's coercions),
importer-prompt complexity, and, farthest out, "asynchronous import
decisions" to decouple indexing from user choices. Two map directly
onto dubplate decisions. The weak-typing complaint corroborates U4:
beets' own maintainers flagged the untyped-attribute path as a pain,
which weighs the U4 choice toward typed columns via migration. The
async-import-decisions idea is beets' own unbuilt sketch of the
decision queue the organizer spec ratified; the field's most
experienced maintainer wanted the feature dubplate is building, and
never got the architecture room to build it against a synchronous
prompt-driven importer.

**Governance statement.** docs/team.rst is a knowledge-owner roster
("Beets is a huge project and not everyone involved, knows
everything") that routes questions per plugin and subsystem. It is the
maintainers' own admission that the surface exceeds any one person's
comprehension, the end state the built-ins-only fork exists to avoid.

Boundary: GitHub Projects boards, if any exist, were unreadable (the
available token lacks the `read:project` scope), and no "state of the
project" post by the current maintainer was found; the maintainer's
priorities here are inferred from milestones, labels, code TODOs, and
shipped-release patterns rather than from a stated manifesto.

## Sources

Repository evidence (clone at
/var/home/glw907/.claude/jobs/daef7196/tmp/beets-history, master
a750105a6):

- Commits: c1ed60af9/ee7bb4b9e (2008 origin), e2d6af6ac/155c9f730
  (2009 matcher origin), a6bb3d088 (munkres orderer), 742ebdf69 (2010
  smart string distance), ee78391f4/064a0432f/de2ee7e44 (2011 source
  refactor), 9e6b1db02 (thresholds), feb1d8631/a5151bb32/e6ac8e164
  (1.2.0 distance rework), 41fb0d928 (candidate sort), 8010488f3
  (2019 distance modularization), 420117b59 (munkres to lapjv),
  e6e610a3e (mb.py to beetsplug), adbd50b23 (distance.py split),
  97904e7e1 (Source abstraction), 7b59604c5/19d091440 (featuring
  fixes), a46c97538/449617050 (MB HTTPS), 06b6f4f27
  (metadata_plugins.py), 32e9e58a3 (SACD fix), 1701aefe5/d325bceb1
  (data/video tracks), 7600a8c0a (pipeline parallel stages).
- docs/changelog.rst sections: 1.0b5, 1.0b11, 1.0b13, 1.1b1, 1.2.0,
  1.3.0, 1.3.15, 1.4.1, 1.5.0, 1.6.0, 2.0.0, 2.2.0, 2.3.0, 2.4.0,
  2.5.x, 2.6.0, 2.7.0, 2.10.0, Unreleased (post-2.13.1).
- beets/config_default.yaml at 9e6b1db02, feb1d8631, and master
  (threshold and weight stability).
- Authorship: git shortlog/log counts, 2013-2026.

Issues and discussions (github.com/beetbox/beets):

- #133 (relative paths, 2.10.0), #251 (VA track artist), #654
  (pseudo-releases, 2014-2023), #1389 (jellyfish), #1487 (ignored
  ignored), #1816 (rate-limited imports), #1916 (romaji, open since
  2016), #2686 (prefer non-MB source, 2017-2025), #3168 (munkres
  versions), #3564 (country preference), #4370 (429 handling), #4388
  (track-assignment override), #4605, #5148 (medium number / hybrid
  SACD), #5580 (flexattr crash), #6059 (disc markers), #6066 (source
  penalty logic), #6081 (weights documented), #6178/#6181 (candidate
  regressions, fixed 2.7.0), #6229 (silent MB loss after 2.4.0),
  #6351 (custom-client packaging), #6561 (path migration), #6728
  (scraper throttling), discussion #4660 (musicbrainzngs
  replacement).

Roadmap section (section 6):

- GitHub milestones via `gh api repos/beetbox/beets/milestones`:
  3.0.0 (updated 2026-06), MediaFile++ (#350, #572, #656, #2621),
  Modularise beets (#1531), Modern MPD support; closed milestones
  1.1 through "Python 3.10 support".
- 3.0.0 milestone contents: PR #6778, #6018, PR #6016.
- In-repo v3.0.0 intent: beets/plugins.py (three "Remove in v3.0.0"
  TODOs), beets/metadata_plugins.py, beets/ui/__init__.py,
  docs/dev/plugins/autotagger.rst ("Legacy support will be removed in
  beets v3.0.0").
- `refactor`-labeled open issues: #704, #842, #1409, #1442, #1531,
  #1715, #3246, #5318, #5361, #6920.
- github.com/beetbox/beets/wiki/Refactoring (last edited April 2018).
- docs/team.rst at master.

Web:

- beets.io/blog/sqlite-nightmare.html (lock timeout saga, 2012)
- beets.io/blog/paths.html (filename encoding, 2016)
- beets.io/blog/flexattr.html (flexible attributes, 2013)
- beets.io/blog/py3k.html (Python 3, 2012)
- musicbrainz.org/doc/MusicBrainz_API/Rate_Limiting

Dubplate documents graded against:

- docs/superpowers/specs/2026-09-01-dubplate-organizer-design.md
- docs/superpowers/specs/2026-09-02-greenfield-server-build.md
- docs/superpowers/research/2026-09-01-beets-conformance-corpus.md
  (house shape and corpus strata)
