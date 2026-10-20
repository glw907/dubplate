## Matcher word-boundary translation (the featuring row's ASCII RE2 rewrite): closed 2026-09-25, lane 1

Executed against `docs/plans/2026-09-25-matcher-word-boundary.md`
(ratified 2026-09-25, three rulings all yes), in the main tree, from
`a12dff9` (Task 0: no drift, pre-fix figures confirmed).

### What landed

Task 1 (`362f73d`) added the featuring-row golden generator
(`internal/engine/matcher/testdata/gen_featuring_golden.py`) and its two
fixtures: the curated-and-drawn golden (33 curated rows, a 10,000-string
draw yielding 7,955 distinct strings, 251 of which beets' pattern fires
on, per `random.seed(907)`) and the 749-range Python `\w` class golden
(Ruling 2, yes). 3/3 named mutations fired. Gate about 238s.

Task 2 (`6751725`, fix round `841cafe` after comment-only review findings)
translated the beets literal to non-capturing-trigger ASCII RE2
(`(^|[^\pL\pN_])(?:featuring|feat|ft)[\. :].+` on the un-bracketed
branch, R-wb3) and added `SDPattern.Replacement` (R-wb1) plus the
translation map living only in `strdist_test.go` (R-wb2). The golden
tests were changed in the fix round to read the shipped table row
through `beetsFeaturingRowIndex` rather than a copy of it. Figures
pinned under both tables: Geschäft 0.5238095238095238, Rüft
0.6470588235294118, Ωft 0.6875. The kept-character row (the bracket
branch keeping `©`/`™`/`€` before the trigger) reads
0.047619047619047616. The ASCII-literal control count is pinned at 98
(the plan cited about 92 as a drawn-only proxy; the shipped constant is
`wantASCIILiteralDivergences = 98`, `strdist_test.go:437`). The class
sweep (`wantGoOnlyWordCodePoints = 9039`, `strdist_test.go:452`) finds 0
Python-only word code points and exactly 9,039 Go-only ones. 8/8 named
mutations fired; mutation 9 is N/A. `TestEngineNeverReadsTheClock`
caught a `time.Now` call left in a test under `internal/engine`, which
the fix round removed.

**İft (U+0130) declared, not fixed (Ruling 3, yes).** Go scores
`StringDist("İft: Ember Wharf", "İft: Northern Lines")` near 0.6875;
beets scores it 0.06875. Go's `strings.ToLower` maps U+0130 to a bare
`i`; Python's `str.lower()` maps it to `i` plus a combining mark `\w`
rejects, so beets sees a word boundary before `ft` and Go does not. A
sweep of every code point outside the surrogates found U+0130 the only
one where Python's lowercase changes whether the last character is a
word character. Pinned as a known-divergence `StringDist` row in Task
2; a fix reds the gate.

Task 3 (`04562e6`, plus README fixes `a23a3d5`) added the
`strdist-feat-unicode-boundary` corpus case (`"Geschäft: Ember Wharf"`
vs. `"Geschäft: Northern Lines"`), graded medium/skip at
0.08730158730158731, absent from `VerdictMoves` as R-wb5 predicted.
Floors moved: case 216 (`RequestCaseFloor`, `case.go:40`), lookup 304,
search 55 (48 to 55 is catch-up: six `adv-ft-*` cases re-frozen at
`367564c` plus this case), candidate/distance/track-vector 295,
assignment 215 (candidate-bearing count, unchanged), verdict 216
(`case.go:41-45`, `request.go:28-30`). `corpus/analysis/ties.json`
grew by one row; the oracle ledger run
(`go test -tags oracle ./internal/harness/oracle -run
TestTieLedgerMatchesTheCorpus -update-ledgers`) took 474s, since tie
derivation re-measures all 216 cases. 2/2 named mutations fired. The
conductor refused a reviewer note flagging the case's invented mbid
(`ubfa0001-...`) as non-hex: beets matches `\w` for that field, so a
re-freeze under a hex-only mbid costs more than the risk it removes.

Simplifier round (`107908c`, accepted, reviewed): routed the golden
checks through `beetsFeaturingRowIndex` at all four call sites,
dropping duplicated lookups; no behavior change.

### Spend

Roughly: Task 0 0.08M, Task 1 0.14M plus review 0.06M, Task 2 about
0.27M plus reviews 0.18M, Task 3 0.20M plus review 0.07M, simplifier
0.11M plus review 0.04M, Task 4 (this close), plus two
`go-architecture-reader` reads after the fold. Lane total about 1.2M
against the 2.75M ceiling (flag 2.2M, never reached). Attended
interactions after plan approval: 0.

### Architecture reads

[`go-architecture-reader` findings for `internal/engine/matcher` and
`internal/harness/conform`, run after this fold with no plan in
context, land here.]

### What a later pass would be wrong to rediscover

- **R-wb3's non-capturing trigger and branch order are load-bearing.**
  A capture on the un-bracketed branch's trigger group would become
  group 1 under Go's opening-parenthesis numbering, so `${1}` would
  re-insert the trigger text into every bracket match and drop the kept
  character everywhere else: the mechanics lens measured 6,351
  divergences from beets over a 60,000-string draw. The bracket branch
  must come first under leftmost-first matching, or a bracket before a
  trigger survives instead of being removed with it (827 divergences
  with the branches swapped). A reviewer proposing to restore the
  trigger's own capture group, or to reorder the branches, is proposing
  a measured regression.
- **The kept character's reason.** The bracket branch keeps whichever
  non-word character preceded the trigger (via `${1}`) because
  unidecode maps some non-word symbols to ASCII letters before the
  featuring row ever runs (`©` to `(c)`, `™` to `(tm)`, `€` to `EUR`);
  dropping the kept character would strip those transliterated letters
  along with the bracket.
- **The Unicode-version gap Ruling 2 settled.** 9,039 code points are
  word characters in Go's `[\pL\pN_]` (Unicode 17.0.0) and not in
  Python's `\w` (Unicode 15.1.0, `unicodedata.unidata_version`); zero
  are the reverse. A Go toolchain bump that moves the Go-only count
  fails `TestStringDistCaseMappingSweep`'s sibling test with a message
  naming the re-read procedure: re-run the sweep, confirm the
  Python-only count is still zero, check whether any newly assigned
  code point can sit before a trigger in a real title, then update
  `wantGoOnlyWordCodePoints` and record the new figure here.
- **The U+0130 divergence is the sole such code point.** The sweep that
  found it (Ruling 3) checked every code point outside the surrogates
  for whether Python's `str.lower()` changes word-character status on
  the last character before a trigger; U+0130 is the only one. A later
  toolchain bump re-runs that sweep rather than assuming the count
  still holds at one.
- **The `-run Ledger` trap.** The corpus README's regeneration command
  (`corpus/README.md:2254`) runs `-run Ledger`, which also matches the
  album-info and file-tags ledger tests, not only the tie ledger. Under
  `-update-ledgers` those two rewrite `album_info.json` and
  `file_tags.json` (`album_info_derive_test.go:91`,
  `file_tags_derive_test.go:87`) even when neither changed. Task 3
  named the one test it needed
  (`-run TestTieLedgerMatchesTheCorpus`) rather than the README's
  broader pattern, to avoid rewriting those two files incidentally.
- **`CompilePatternOption`'s configured-pattern class stays documented
  only,** not built: the plan scoped this pass to the one shipped
  featuring row, and a configurable-pattern surface was never a
  deliverable.
- **`shadow.configHash` (`internal/engine/shadow/match.go:55-75`)
  hashes the Tunables JSON**, so the new `Replacement` field and the
  translated pattern text move `matcher_config_hash` at the next
  deploy. This is intended, not a regression: a shadow report showing
  two distinct hashes either side of the deploy is expected, not a
  finding.
