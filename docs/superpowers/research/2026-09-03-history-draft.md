<!--
DRAFT for the conductor's HISTORY.md batch commit (pass 2c Task 20).
Not committed as docs/HISTORY.md directly per the task's own
instruction; the conductor folds this in. Insert as a new top entry,
newest-first, above the "Rung 2, pass 2b" entry.
-->

## Rung 2, pass 2c (Real-catalog and adversarial strata): 147 new cases, six oracle-bug hypotheses, first adjudication batch (2026-09-03)

Twenty plan tasks executed across a single sequential pass-execute
chain against
`docs/superpowers/plans/2026-09-03-rung-2c-catalog-adversarial.md`
(revision 2). Commits `3bbdfbd` → `[TASK20_SHA]`.

What landed: the adjudication and known-oracle-bug ledger formats
(Task 1); the `harness seed-items` projection and `--facts` tools
(Task 2); the real-catalog selection dossier, 96 primary rows plus 12
reserve rows across seven axes (Task 3); twelve capture batches (Tasks
4-15) landing 94 real-catalog primary cases against real MusicBrainz
document shapes (two two-release sub-axis rows lost to the mechanism's
own escape hatch, real captures over MB-noise, non-Latin script,
classical/compilation, genre/era spread, reissue/remaster, and
VA/box-set releases, one `ignore_video_tracks` finding); the
adversarial stratum (Tasks 16-18), synthetic cases reproducing two
known string-heuristic beets defects (the `"ft"` word-boundary false
positive and the parenthetical/remix near-duplicate discount, ten
cases, both moving strong/apply to medium/skip) plus six
per-pair track-assignment rejection-gap cases (the album verdict never
moves even while `assignment`/`applied_fields.title` are provably
wrong) plus the alias-mis-selection hypothesis probed against two of
Task 4's own real-nonlatin cases, plus Task 18's `locale` profile and
its A/B regression-check case pair for `#6358`; the stock-config
stratum (Tasks 19a/19b), 36 siblings (28 mandatory + 8 supplementary)
proving zero verdict-level divergence between `estate` and `stock`
across every real-catalog axis, with the divergence landing only in
`applied_fields` (`year` for reissues, `tracktotal` for box sets).
Corpus total: **215 cases** (7 smoke + 61 structural + 94 real-catalog
primary + 17 adversarial + 36 stock-config), read off the tree
directly. Determinism: two independent uncached (`-count=1`) full
`make derive` runs against the whole `internal/harness/...` and
`cmd/harness/...` tree, 215 and 215 cases actually re-derived of 215
committed (`TestCorpusReDerivesToItsCommittedRecords`'s own per-case
subtests, the count that proves a case was actually re-derived rather
than skipped for want of a committed record), both runs exiting 0 with
zero failures across every package, both byte-identical to the
committed records. Wall time: run 1 1729.58s (28m50s) total (corpus
derive 927.91s, ties derive 786.04s, component derive 1.22s); run 2
1582.53s (26m23s) total (corpus derive 812.66s, ties derive 754.48s,
component derive 1.27s) — this is rung 3's budget line, exactly as
2b's was 2c's own (2b's own pair ran ~2m13s each on 68 cases; 2c's
real-catalog and box-set growth moved that budget line by roughly
13x). The heaviest single cases in either derive phase are the
box-set axis's largest releases: ties derive's own re-solve-per-pair
algorithm (`internal/harness/oracle/derive_ties.py`) pays for
`real-boxset-rachmaninov-complete`'s 470 tracks and its stock twin
individually, each taking on the order of a minute alone, the same
box sets that dominate the byte-growth figure below. Corpus byte
growth (m7): the committed corpus held 4.0 MB across 114 response
files before this pass; it now holds **81.3 MB across 351 response
files**, well past the plan's projected 10-15 MB
landing point. The overshoot is a response-*size* story, not a
response-*count* one (351 files is close to what the plan's own count
math expected): a handful of the VA/box-set axis's largest releases
serialize their entire tracklist inline in one WS/2 document
(`real-boxset-rachmaninov-complete`, 470 tracks, 9.1 MB alone;
`real-boxset-scriabin-complete`, 321 tracks, 3.1 MB;
`real-boxset-mahler-complete`, 133 tracks, 1.7 MB), and Task 19a's
mandatory-sibling `cp -r` rule doubles every one of them for its stock
twin.

What the gate caught: nothing new at Task 20 itself (README-only and
analysis-doc changes, no code); across the pass, the recurring catch
was a case's own `claims.json` failing to notice a degraded record
(the claims-checker's whole design point) and, once, a stratification
undercount in Task 19b's own README prose (fixed same-pass, commit
`9915907`).

What a later pass would be wrong to rediscover:

- **File-level derivation is 2d's Task 13, unchanged in scope**:
  `internal/harness/oracle/derive_file_tags.py` and its Go driver,
  reading back independently through `mediafile.MediaFile(path)` after
  `item.write()`, never trusting the write call's own return value,
  over 12 subject cases (5 from the committed corpus, 7 flagged by
  this pass's dossier).
- **The audio-fixture route is ruled** (M4): vendor one small static
  template FLAC blob under `internal/harness/oracle/testdata/`,
  committed, with its provenance and byte size recorded in the README.
  `make derive` gains no binary dependency; `ffmpeg`/`flac` are not
  adopted as prerequisites.
- **`derive_file_tags.py` is a second, independent entry point** (M5),
  shaped like `derive_component.py`: it reads a committed record's
  `AppliedFields`, builds beets items from those fields, allocates its
  own scratch directory, writes and reads back there, and is never
  called from `oracle.Derive`. The invariant: no item in the corpus
  derivation path ever gains a non-empty `path`, proved by `make
  derive` re-deriving all committed records byte-identically after 2d's
  change.
- **`file_tags.json` needs an oracle-tagged re-derive-and-diff test**
  (M3), `TestFileTagsLedgerMatchesTheCorpus`, matching
  `TestTieLedgerMatchesTheCorpus`/`TestComponentTableReDerivesToItsCommittedTable`'s
  precedent, plus the `-update-ledgers` regeneration path documented in
  `corpus/README.md`. A gate-resident structural check alone (case ids
  resolve, pointers resolve) is not enough; nothing else would catch a
  `mediafile` version bump moving the tag shapes.
- **`Reconcile` lives in `internal/harness/oracle/reconcile.go`, not
  `oracle.go`, and its `versionProbe` tuple names five components
  today** (m2: beets, unidecode, jellyfish, lap, numpy), not four. 2d's
  Task 13 must edit `reconcile.go` and `versionProbe` explicitly when
  it adds `mediafile` to `Manifest`, or a reconcile silently compares
  against a missing key.
- **The version-bump differential and quarterly re-capture chore docs
  are unwritten**, deferred with Task 13. The differential doc's
  wall-time section must use the M8 inter-request-floor formula
  (`1.04s + 0.25s × (requests − 1)`) at the corpus's real size once 2d
  closes, not a flat per-case figure.
- **The whole-corpus README rewrite happens again at 2d's close**,
  once, to add the file-tags section; this pass's rewrite is not
  wasted work, since 2d's version extends it rather than replacing it.
- **The response-status extension was deliberately not taken this
  pass**: the stub has no third outcome and the schema carries no
  response status, because all four string/assignment/alias
  known-oracle-bug classes this pass probed are distance or selection
  defects, not transport defects — none needs a 404 or 503 to
  reproduce. It carries forward as a named non-adoption, available to
  whichever later rung first has a case that genuinely needs it.
- **The `profile_hashes` gap (m1)**, distinct from the 2d file-tags
  items above: `corpus/manifest.json` gained a `locale` entry under
  `profile_hashes` at Task 18, but nothing in the gate or the
  derivation path computes or checks it yet for *any* profile. A
  profile edit does not fail loudly the way a beets version pin
  already does; this is scoped follow-on work, not something the
  table's presence implies is built.
- **Tier C exclusions and profile gaps 2b carried that 2c did not
  close**: multi-source ordering and most of the `data_source`
  parameter space stay blocked by the single-source ruling (unchanged
  from 2b); `match.required`/`match.ignored` candidate-dropping stays
  unprobed (same shape, still open); `preferred.original_year: yes`
  stays unreachable under all three profiles now committed (`estate`,
  `stock`, and Task 18's own `locale`, which only toggles
  `import.languages`/`import.ignored_alias_types`) — a fourth profile
  or an extension to `locale.yaml` is still owed if a later rung wants
  this branch probed.
- **The alias-duplicate incident (Task 17) and its standing rule**: an
  earlier pass of Task 17 captured four byte-identical `adv-alias-*`
  cases duplicating releases the corpus had already committed under
  Task 4's `real-nonlatin-*` rows, because case selection never
  checked the committed tree first. The conductor's escalation ruling
  folded the four duplicates back into their existing twins (commits
  `e805540`, `c685894`); nothing was lost, but the rework was pure
  waste. **The rule a later pass must inherit: check `corpus/cases/`
  before any capture**, real or synthetic, that might already exist
  under a different id.
- **The `_preferred_alias` unreachability finding and Task 18's
  resolution of it**: Task 17's alias-mis-selection probe found
  `_preferred_alias` (`beetsplug/musicbrainz.py:128-155`) structurally
  unreachable under both `estate` and `stock`, since both leave
  `import.languages` empty, so its generator yields nothing regardless
  of what a release document carries. This was left as a scoping
  finding, not a reproduction, and deferred to Task 18, which added the
  `locale` profile (`stock.yaml` plus `import.languages: [ja]` and
  `import.ignored_alias_types: []`) and an A/B case pair
  (`real-nonlatin-utada-first-love` / `-locale`) to probe it directly.
  The result: the mechanism *is* reachable and *does* fire once
  `import.languages` names a matching locale, moving exactly the
  sort-name fields it owns (`artist_sort`, `albumartist_sort`,
  `artists_sort`, `albumartists_sort`) with no other change in verdict
  or scoring — evidence against `#6358`'s report, pending Geoff's own
  ruling (`bug-locale-alias-ignored`, adjudication batch 1).

Carried:

- **The pending adjudication batch is the launch gate's open item.**
  Nineteen rows across five hypotheses sit at
  `disposition: "pending-adjudication"` in
  `corpus/analysis/adjudications.json`
  (`docs/superpowers/research/2026-09-03-adjudication-batch-1.md` is
  the artifact prepared for Geoff's ruling). This pass's own gate
  accepts them indefinitely; it is the **launch gate**, a different
  gate at a later time, that refuses an unadjudicated divergence.
  `bug-threshold-boundary` (`#5458`) is the corpus's one hypothesis
  already ruled (`oracle-correct`, Geoff, 2026-09-03) and is not part
  of the pending batch.
- `corpus/analysis/divergences.md` stays empty at this pass's close,
  reviewed and carried forward: no case across the real-catalog,
  adversarial, or stock-config strata contradicted a specific beets
  test assertion it was modeled on (findings of that shape instead
  became adjudication-batch rows or plain README prose).
- The two-release sub-axis closed at 4-of-6 primary rows delivered (2
  discarded via the mechanism's own escape hatch, the sole reserve
  spent by Task 4); a later two-release row anywhere in the corpus
  should budget for the same WS/2 search-ranking risk with no reserve
  behind it.
- Every item from 2b's own carry-forward that 2c did not separately
  close above (Tier C, `match.required`/`match.ignored`,
  `preferred.original_year`) stays open into 2d and beyond.

Budgets: [conductor: fill from the pass's own token ledger, continuing
STATUS's "≈6.5M/8M through Task 17" line with Tasks 18-20's spend and
the pass's final total against the 12M ceiling; this implementer
dispatch has no visibility into other tasks' spend].
