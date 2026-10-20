# Matcher word-boundary plan: fold 1 verification

Subject: `docs/plans/2026-09-25-matcher-word-boundary.md`,
revision 2 at `a3773d1`, against the fold record
(`2026-09-25-matcher-word-boundary-plan-fold.md`) and the three lens reviews.
This reader had no part in the fold. Revision 1 was never committed, so each
finding was checked against revision 2 directly.

Verdict: **approvable after four minor folds; no blocker, no major.** Every
major closed where the fold record says it did. The fold introduced one
contradiction (exit gate 5) and a few stale or overbroad sentences. None of
them changes what ships.

## Checks run

- `scripts/check.sh:380-392`: lines 385-387 read "requires uv or beets.
  Tracked files only, so an unstaged experiment / does not gate; ruff.toml
  carries the rule set." then `mapfile -t py_files < <(git ls-files '*.py')`.
  The plan's quote is exact.
- Beets 2.14.1 `distance.py` in the venv: `SD_PATTERNS` opens at `:36`, the
  featuring row is `:39`, and `re.sub(pat, "", str1)` is at `:100`. All exact.
- `go doc` at go1.27.1: the `regexp/syntax` lines, `strings.ToLower`,
  `Regexp.Expand`, and `ReplaceAllString` quotes are exact. The package-doc
  sentence names the `All` iterator family (not `Find`, as the plan says). The
  plan's point stands: the sentence does not cover `ReplaceAll`. The same
  package doc states group numbering "in order of opening parenthesis", which
  backs R-wb3's new reason.
- Line pointers: `strdist.go:36-37`, `:43-44`, `:103`, `:128`, `:134`;
  `tunables.go:324`, `:365`, `:371-374`, `:375`, `:398`;
  `strdist_test.go:113`, `:131`, `:143`, `:223`, `:240`;
  `tunables_test.go:185`, `:202`; `compat.go:47-53`;
  `derive_defect_free.py:24-27`, `:34-36`; `case.go:672-680`;
  `ties.go:118`. All resolve to the content the plan describes.
- `git grep -nw -e 215 -e 214 -- '*.go'` returns exactly the five comment
  sites Task 3 names, plus the constants. `ls corpus/cases | wc -l` is 215.
- `-run Ledger` matches five tests (one tie, two album-info, two file-tags).
  `-run TestTieLedgerMatchesTheCorpus` matches one.
- `/tmp/claude-1000/ubound/gen.py` still exists. Its token list, in order,
  matches R-wb4's 23 tokens (the second space is U+00A0).
- A reproduction of R-wb4's written-out draw in the pinned venv (CPython
  3.13.15; seed 907, 10,000 strings, `dict.fromkeys`) found 7,955 distinct
  strings and 251 fired. Using `re.ASCII` as the proxy for Go's ASCII `\b`,
  92 rows diverge from the beets literal. One-row-per-line output with
  `ensure_ascii=False` writes 242,193 bytes for the drawn rows, and
  `indent=1` writes 305,837.
- `go doc -all ./internal/engine/matcher` prints `var BeetsSDPatterns` with
  its full initializer and its doc comment.
- No corpus string carries U+0130 directly before a trigger, so Ruling 3's
  fix cannot move a committed figure.

## Q1: did each major close at the cited location?

| Major | Closed? | Where in revision 2 |
| --- | --- | --- |
| contract M1 (the class test proved a constant) | Yes | Ruling 2 (`:85-90`), Task 2 parity outcome (`:666-673`), states (`:703-705`), mutation 8 (`:752-753`). The sweep drives `"a" + string(r) + "ft. x"` through the shipped row and names the bracket special case. Mutation 8 is sound: U+00B2 is the lowest `No` code point, so it is the first Python-only code point the failure names. |
| contract M2 (curated rows unproved) | Yes | R-wb4 (`:409-412`, `:443-452`), Task 1 acceptance (`:586`), Task 2 state (`:698`), Review focus 3 (`:494-497`). |
| contract M3 (draw not written out; no floor) | Yes | R-wb4 (`:421-441`) matches `gen.py`. The floors are at Task 1 (`:577`) and Task 2 (`:699`, `:701`), exit gate 2, and Task 0 note 8. Both floors hold with margin on the real draw (251 against 150, and about 92 against 40). |
| mechanics M1 (mutation 4 can survive) | Yes | R-wb3 (`:389-396`) states the group numbering and the 6,351 and 0 counts. Mutation 4 (`:743-746`) names branch 1 and rules out branch 2. |
| mechanics M2 (ruff never sees the generator) | Yes | Global constraints (`:314-323`) quotes `check.sh:385-387` exactly. Task 1 acceptance (`:581-583`) runs ruff directly and pastes the output. |
| risk M1 (the U+0130 divergence) | Yes | Ruling 3 (`:95-116`), Goal (`:234-235`), Behaviors (`:197-200`), Task 2 outcomes (`:629-632`, `:643-645`, `:681-688`), states (`:707`), mutation 9, and Task 4 (`:885-886`). |

## Q2: contradictions, build order, deliverable counts

**Build order holds.** Task 1 commits Python and data that no Go test reads,
so its gate stays green. Task 2 adds the reader and the fix together. Task 3
needs Task 2's fix to grade `ModeCompat` green. The simplifier round runs
after Task 3 and before Task 4. The execution model and conductor action 3
agree on that order.

**Deliverables** are 3 (2 on Ruling 2's no), 4, 4, and 2, which sums to the
thirteen stated. No task passes four under any answer to the rulings.

**Ceiling arithmetic** is correct. The line items sum to 2.12M. 80% of 2.75M
is 2.2M. Under Ruling 1's no, spend is about 1.82M before the merge-side
reads, and the 0.25M of reads crosses 2.0M, as the ruling says.

**One contradiction was introduced. See finding 1.**

## Q3: new mechanisms stated from memory

Every new mechanism the fold added is quoted or backed by a cited probe.
There are two exceptions:

- **Exit gate 5's `go doc -all` diff** is stated from memory, and it is
  wrong (finding 1).
- **"ruff reads the repo's `ruff.toml` by discovery"** is unquoted. It
  matches ruff's documented hierarchical config lookup, and the file sits at
  the repo root with the `E`, `F`, `D` (google) set the plan names. The
  consequence is low. If discovery failed, the D rules would not run on the
  generator, and Task 2's gate would catch the file one task late. No fold
  is proposed.

The fold's draw figures ("the mechanics lens measured 331 of 7,898") are
cited to a probe. That probe did not use the draw R-wb4 now writes out
(finding 2).

## Q4: can the owner rule each ruling, and does the plan execute under either answer?

- **Ruling 1: yes, on both counts.** Each line item and its basis in 6a's
  actuals is shown. Under the no, the plan says exactly where the flag trips
  and what the conductor does then.
- **Ruling 2: yes, on both counts.** The measurement is quoted with its
  versions and counts, and the mechanics lens reproduced it. Under the no,
  the plan drops one file, one test, and mutation 8, and Task 1 falls to two
  deliverables.
- **Ruling 3: yes, on both counts.** Each option is costed, and the
  case-mapping sweep's comment shows the precedent. Under the no, the fix
  makes Go's lowercased input the same text as Python's before the featuring
  row runs. The strip removes U+0307. So `0.06875` should reproduce exactly;
  this reader reasoned that and did not probe it. If it fails to reproduce,
  the state table fails the row, and the task stops rather than tolerating
  it. The sweep test calls `sdStrip(strings.ToLower(...))` directly, not
  `StringDist`, so the fix leaves it alone. One sentence overclaims
  (finding 3), but the owner can still rule on the evidence given.

## Findings, ranked by consequence

### 1. minor: exit gate 5 cannot pass as written

- **Location:** the exit gate, item 5 (`:917-919`). It conflicts with Task
  2's outcome (`:633-639`).
- **Defect:** The gate says `go doc -all` at base and HEAD "differ by the
  field and its doc comment alone." `go doc -all` prints each exported var's
  full initializer and its doc comment. Task 2 requires changing both for
  `BeetsSDPatterns`: it replaces the "must stay byte-identical" sentence and
  the gap note, and it rewrites row 3. It also rewrites row 3 in
  `SDPatternsNoBracketDiscount`'s initializer. So the check fails on every
  correct execution. The cost is a spurious failure at the close, and a
  conductor ruling or a fix round to clear it. Shipped code is not affected.
- **Proposed fold:** Restate the check's intent, which is no new exported
  identifier. One way: `go doc -short ./internal/engine/matcher` at base and
  HEAD list the same symbols, and `go doc matcher.SDPattern` differs by the
  `Replacement` field and its comment alone.

### 2. minor: the plan's expected draw figures do not match the draw it writes out

- **Location:** R-wb4 (`:437-438`) and Task 1 acceptance (`:587-588`),
  "the mechanics lens measured 331 of 7,898 distinct strings".
- **Defect:** R-wb4's own draw gives 7,955 distinct strings and 251 fired.
  The lens's 331 of 7,898 came from a different token set. Both floors still
  hold, so no gate reds. But an implementer or diff-reviewer who compares the
  report with "331" may spend a round on a figure that is not a defect.
- **Proposed fold:** Quote the reproduced figures instead: 7,955 distinct,
  251 fired, and about 92 ASCII-literal divergences (Python `re.ASCII`
  proxy). Name them as this verification's probe.

### 3. minor: Ruling 3's yes claims more than one row can pin

- **Location:** Ruling 3 (`:107-109`), "so a fix or a second instance reds
  the gate".
- **Defect:** The row pins U+0130 only. A second code point whose lowercase
  crosses the word-character line would need a Go case-table change, and
  that row would not see it. Neither would the strip sweep, which tests
  `sdStrip`, not the pattern loop.
- **Proposed fold:** Drop "or a second instance". Put the one-code-point
  sweep result in Task 4's HISTORY line, so a later toolchain bump knows to
  re-run the sweep.

### 4. minor: the Goal says "every input", but Ruling 2 records inputs where the row cannot match

- **Location:** Goal (`:231-233`).
- **Defect:** "Matches pinned beets on every input it sees" is false for the
  9,039 code points that Go treats as word characters and Python does not.
  Ruling 2 and R-wb4 both state this, and mechanics measured 452 of 452
  divergent rows.
- **Proposed fold:** Add "outside code points first assigned after Unicode
  15.1 (Ruling 2)".

## Counts

0 blockers, 0 majors, 4 minors. One is a fold-introduced contradiction
(finding 1). Two are stale or overbroad sentences the fold added (findings 2
and 3). One is an overclaim this pass left in place (finding 4).
