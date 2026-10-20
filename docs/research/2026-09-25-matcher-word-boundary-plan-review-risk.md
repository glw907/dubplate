# Matcher word-boundary plan review: domain risk

Lens: domain risk, one of three. Subject:
`docs/plans/2026-09-25-matcher-word-boundary.md`, revision 1
(uncommitted). Read against master at `1306ae3`, the pinned venv (beets
2.14.1, CPython 3.13.15, `unicodedata` 15.1.0), and Go 1.27.1. Findings are
correctness gaps ranked by consequence. A fold may refuse any of them with a
one-line reason.

**Counts: 0 blockers, 1 major, 9 minors, 1 owner fork.**

## What was checked and found sound

These need no fold. They are recorded so the fold does not re-check them.

- **Verdict drift on existing cases: none.** A scan applied beets' featuring
  literal with and without `re.ASCII` to every lowercased JSON string under
  `corpus/` (6,697,851 strings: cases, responses, and every analysis ledger)
  and under `internal/` and `cmd/` (9,886 strings). It found zero strings
  where the ASCII and Unicode boundaries disagree. No committed record,
  tier-0 component row, or defect_free figure can move. Task 2's claim "No
  existing expected value changes" holds. A grep of `*_test.go` for a
  non-ASCII character before a trigger also finds nothing.
- **Leftmost parity of the two-branch rewrite.** Each beets match at
  position q maps to a Go match at q (bracket branch) or q-1 (boundary
  branch), and back. When a bracket precedes a trigger, both engines prefer
  the bracket at the same start. `x((ft. y)` gives `x(` in both. `.+` stops
  at a newline in both, so a consumed boundary character never belongs to a
  prior match.
- **The ledger.** `ties.go:118` requires a row for every case, so a one-row
  `ties.json` diff is the right expectation. The predicted distance, about
  0.087, is far from every threshold, so no `proximity.json` row is owed.
  `adjudications.json` is an exception log and needs no row for a case that
  agrees with beets.
- **Lane ordering.** `git diff master...lane-2-6b-i` touches nothing under
  `internal/engine/matcher`, `internal/harness`, `corpus`, `scripts/check.sh`,
  or `Makefile`. The 6b-ii plan walls off `internal/harness` and `corpus/`
  (its `:251-259`). Running this pass in the main tree after 5c-ii closes,
  beside a 6b-ii worktree, is consistent with CLAUDE.md and the unattended
  window plan.
- **Gate runtime and determinism.** 10,000 regexp replacements take
  milliseconds. A 1.1M-code-point class sweep takes well under a second
  without `-race`, and `make check` runs plain `go test ./...` (Makefile:18).
  gitleaks' generic rule needs a key-like field name on the same line
  (`check.sh:408-411`), which drawn title tokens do not carry.
- **Citations.** Every `tunables.go`, `strdist.go`, `strdist_test.go`,
  `tunables_test.go`, `compat.go`, `case.go`, `request.go`, `ties.go`, and
  `corpus/README.md` pointer spot-checked lands within a line or two of the
  named content.
- **Ratified documents.** No contradiction with the greenfield rulings, the
  2.0 spec (its `:239-240` rerank on `matcher.StringDist` is as quoted), or
  `corpus/manifest.json`, which lists no cases.

## Major

### M1. The lowercase step leaves a non-ASCII featuring divergence the golden cannot see

- **Location:** plan `:183-185` (Goal), `:327-348` (R-wb4), `:530-537`
  (Task 2 unit rows); `internal/engine/matcher/strdist_test.go:222-225`.
- **Defect:** The Goal says that after this pass, `StringDist` scores a
  non-ASCII letter before a trigger "exactly as pinned beets does." That is
  false for U+0130 (İ). Python's `str.lower()` maps it to `i` plus U+0307, a
  combining mark that `\w` rejects, so beets sees a boundary before `ft`.
  `strings.ToLower` maps it to a bare `i`, so neither the literal nor the
  rewrite fires. Measured today: beets `string_dist("İft: Ember Wharf",
  "İft: Northern Lines")` is `0.06875`. A scratch Go run shows the rewrite
  leaves `ift: ember wharf` untouched, so Go scores near 0.6875. An
  all-caps Turkish title is the realistic carrier. A sweep of every code point
  finds U+0130 is the only one where Python's `lower()` changes whether the
  last character is a word character. The golden runs the compiled row on
  raw inputs and skips the pipeline's lowercase, so it cannot see this. The
  existing case-mapping sweep comment says of U+0130 that "the strip removes
  the dot, so the two agree." That holds for `sdStrip` only. The featuring
  row runs before the strip and sees the dot. The divergence exists today
  too, so this pass does not create it. But the pass claims to close the
  class, and a later shadow run or pass would have to rediscover it.
- **Proposed fold:** Task 2 adds one `StringDist` row that writes the U+0130
  divergence out as a known divergence, the way `knownDivergences` does in
  the case-mapping sweep. It pins beets' `0.06875` and Go's measured figure,
  so a fix or a second instance reds the gate. Narrow the Goal to "the
  featuring row matches beets on every input; one case-mapping divergence
  upstream of it is declared." Correct the sweep comment at `:222-225` to say
  the agreement holds for the strip, not for the pattern loop. Add this to
  Task 4's list of what a later pass would be wrong to rediscover. Whether to
  fix it instead is the owner fork below.

## Minor

### m1. ruff sees only tracked files, so Task 1's ruff criterion is vacuous before commit

- **Location:** plan `:253-256`, `:465-466`; `scripts/check.sh:388`.
- **Defect:** The ruff leg runs over `git ls-files '*.py'`. The new generator
  is untracked until committed, so a pre-commit gate never lints it, and a
  lint failure first appears on the next task's gate.
- **Proposed fold:** Task 1 runs `ruff check` and `ruff format --check` on
  the file directly, or stages it before the gate, and pastes the output.

### m2. `-run Ledger` runs five oracle tests and can rewrite two ledgers outside the territory

- **Location:** plan `:621-623`, `:794-795`, `:245-251`.
- **Defect:** The pattern matches `TestTieLedgerMatchesTheCorpus`, both
  `TestAlbumInfoLedger*` tests, and both `TestFileTagsLedger*` tests. With
  `-update-ledgers`, the matching tests rewrite `album_info.json` and
  `file_tags.json` as well. Neither file is on the plan's territory list.
  This also answers the plan's open "Not verified" item on scope, and the
  extra tests add wall time.
- **Proposed fold:** Run `-run TestTieLedgerMatchesTheCorpus -update-ledgers`,
  or keep the README command and require `album_info.json` and
  `file_tags.json` to come back byte-identical, stopping otherwise.

### m3. The family 4 README text becomes false, and one appended sentence does not fix it

- **Location:** plan `:624-628`; `corpus/README.md:232-257`, `:1572`,
  `:1917`.
- **Defect:** Family 4 is defined as pairs drawn from beets' own
  `test_distance.py` that should read "equivalent, or ... merely closer."
  The new case is neither. It is a no-fire case whose strings must stay far
  apart. "5 cases" and "Four of the five pairs" also go stale. The totals
  sum `7 + 61 + 94 + 17 + 36 = 215` needs its structural term moved to 62,
  and `:1572` credits all 61 structural cases to pass 2b.
- **Proposed fold:** Task 3 rewrites the family's defining sentences to
  admit a port-hazard probe, and moves the structural term in both totals.

### m4. Task 3's mutation 1 predicts the wrong failure

- **Location:** plan `:660-661`.
- **Defect:** Under the literal, Go scores the case strong (about 0.0087)
  and takes the strong-match early return, so it never issues the recorded
  search. The first `ModeCompat` failure is likely the request set or the
  recommendation, not a failure "naming `album`." An implementer who reads
  the plan literally may report the mutation as not behaving as predicted.
- **Proposed fold:** Accept any `ModeCompat` failure on the case, and report
  which rule fired.

### m5. The test that pins the default table to the beets literal is not named

- **Location:** plan `:515-525`; `strdist_test.go:130-132`.
- **Defect:** `TestBeetsSDPatternsMatchBeetsLiterals` ends by comparing
  `DefaultTunables().StringDistPatterns[2].Pattern` with the beets literal.
  After the translation, that assertion reds. The plan rewrites the test's
  table assertions but does not name this line, so an implementer may
  "fix" it by weakening it.
- **Proposed fold:** Name it. The line should route through the translation
  map like the rest.

### m6. Stale 215 counts in Go comments beyond the floor block

- **Location:** plan `:629-631`; `internal/harness/conform/case.go:29`
  ("AssignmentFloor is the 214 candidate-bearing cases"), `case_test.go:286`,
  `defaults.go:17`, `internal/engine/tag/tag_test.go:244`.
- **Defect:** The plan moves only the constants and the measured date. The
  prose counts go stale, and some are claims about what the corpus holds.
- **Proposed fold:** Task 3 sweeps `git grep -nw 215` and `214` in Go
  comments and moves each one that counts cases.

### m7. Two regexp behaviors the rewrite relies on are not quoted

- **Location:** plan `:123-177` (planning-miss item 1).
- **Defect:** The rewrite relies on two behaviors the plan does not quote.
  First, `[^\pL\pN_]` matches `\n`, which the multi-line row depends on.
  Second, `ReplaceAllString` replaces successive non-overlapping matches,
  which the two-trigger row depends on. The plan quotes neither from
  `go doc`. The planning-miss rule asks for each relied-on behavior to be
  quoted.
- **Proposed fold:** Quote the `regexp` package doc on non-overlapping
  matches. Name the golden's newline row as the proof for the negated class,
  since `regexp/syntax` does not state it. Quote `strings.ToLower` next to
  M1.

### m8. CLAUDE.md lists the tracked Python files, and this pass adds one

- **Location:** plan `:252-256`, `:727-741`; CLAUDE.md "Gate" and "Stack
  rules: Python."
- **Defect:** Both CLAUDE.md passages enumerate every tracked `*.py` file and
  every place Python may live. The new
  `internal/engine/matcher/testdata/gen_featuring_golden.py` makes both lists
  incomplete.
- **Proposed fold:** Add a conductor action at the close to add the file to
  both lists. The conductor owns CLAUDE.md.

### m9. Ruling 2's count test reds on a routine Go bump with no stated procedure

- **Location:** plan `:64-84`, `:552`.
- **Defect:** The dependency-upgrade policy takes Go minors by default. The
  first Go release on Unicode 18 reds the gate at the 9,039 count, and the
  message names only the count and `unicode.Version`. Without a stated
  procedure, the easy fix is to bump the constant. That defeats the re-read
  the test exists to force.
- **Proposed fold:** On Ruling 2's yes, the failure message states the
  procedure. Confirm the Python-only count is still zero. Check whether any
  newly assigned code point can sit before a trigger in a real title. Then
  update the count, recording it in HISTORY.

A tenth item is recorded without a fold: `shadow.configHash`
(`internal/engine/shadow/match.go:55-75`) hashes the Tunables JSON, so the
new field and the new pattern text move `matcher_config_hash` at the next
deploy. That move is intended. Task 4's HISTORY draft can mention it so a
shadow report showing two hashes is not a surprise.

## Owner fork

**Declare the U+0130 divergence (M1), or fix it in this pass?**

- **Declare (recommended).** One pinned known-divergence row, the corrected
  comment, and a HISTORY line. This stays inside the 2026-09-24 ruling's
  five deliverables plus one row. The carrier is rare: an İ directly before
  `ft`, `feat`, or `featuring`.
- **Fix now.** Before `strings.ToLower`, replace U+0130 with `i` plus
  U+0307, which is Python's full lowercase mapping for the one code point
  where Go differs. That is a second named departure, and it also changes
  what `sdStrip` sees (the strip removes the dot, so the strip still
  agrees). It is about one line, but it adds a deliverable to Task 2 and
  widens the ruling's scope.
- **File for later.** Declare now and add an IDEAS or ROADMAP entry for the
  fix.

Recommendation: declare now. It keeps the pass inside what Geoff ruled, and
the gate makes the known divergence visible.
