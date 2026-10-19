# Matcher word-boundary plan review: contract and criteria

Lens: contract and criteria. The question is whether every promise is testable, whether each
acceptance criterion names its fixture and why it would fail, and whether any criterion or named
mutation can pass vacuously. Reviewed: `docs/superpowers/plans/2026-09-25-matcher-word-boundary.md`,
revision 1 (uncommitted), against the research brief
`docs/superpowers/research/2026-09-24-matcher-unicode-word-boundary.md`. Code pointers were read at
the working tree: `tunables.go:334-403`, `strdist.go:25-193`, `compat.go:27-53`,
`strdist_test.go:103-290`, `conform/case.go:40-45,150-172,520-682`, `request.go:28-30`,
`analysis/ties.go:95-135`, and the corpus README's ledger section. The scratch generator
`/tmp/claude-1000/ubound/gen.py` was also read.

Counts: 0 blockers, 3 majors, 8 minors.

## Sound areas

- **R-wb5 is accurate.** `compatFor` starts from `DefaultCompat()`, and the whole-corpus compat
  sweeps run `FullCompat()` (`case_test.go:61`), which means `BeetsSDPatterns`. `gradeCaseDefault`
  re-scores committed candidates and reports a moved verdict without failing (`case.go:672-680`).
  The plan claims no more than the harness delivers.
- **The kept-character proof is real.** Under mutation 3 (the loop passes `""`), `"ember wharf
  ©ft. northern"` becomes `"ember wharf "`. After unidecode and the non-alphanumeric strip, that
  string differs from `"ember wharf ©"` (`...c`), so the `©` row in `TestStringDist` fails. Without
  the mutation, the kept `©` makes the two sides identical. Mutation 3 cannot pass vacuously.
- **The golden test's state table is complete.** It covers absent, malformed, header drift, empty
  or short, zero fired rows, many diverging rows, and the ASCII-literal control. The control is
  the right non-vacuity check, since it proves the golden can tell the translation from the
  defect.
- **Weights and order stay pinned after `TestBeetsSDPatterns_SixEntriesInBeetsOrder` is
  removed.** The translation-map assertion compares whole `SDPattern` values against the
  written-out literal table, which carries weights. The map-less rows must also carry an empty
  `Replacement`, so the privileged-field row for `Replacement` does have a refusing test.
- **Ruling 2's Go-only count pin (9,039) catches a truncated range file.** Missing ranges inflate
  the Go-only count, so the pin needs no separate range-count assertion.

## Findings, ranked by consequence

### M1 (major): Ruling 2's class test proves a constant, not the shipped row

Location: plan `:64-84` (Ruling 2), `:528-529` (Task 2 outcome), `:550-552` (states).

**Defect.** Ruling 2 promises that "the gate proves the rewrite's character class on every code
point." Task 2 says the class test "compares [the range file] with `[^\pL\pN_]`." As written, the
test would compile its own copy of the class. It would not probe the shipped featuring row. An
edit to the shipped row's class, such as `[^\pL\p{Nd}_]`, `[^\pL_]`, or a dropped `_`, would
leave this test green. Only the golden's drawn characters (`²`, `Ⅻ`, `_`) would catch the edit.
That is exactly the limit Ruling 2 says it removes. Ruling 2 also names no mutation, so nothing
shows the test can fail against production.

**Fold.** On Ruling 2's yes, drive the test through the shipped compiled row and its template,
with no separate class literal:

1. For each code point `r` outside the surrogates, substitute on `"a" + string(r) + "ft. x"`.
2. Classify `r` as a Go word character when the output is unchanged. It is a non-word character
   when the output is `"a" + string(r)`, or `"a"` for `(` and `[`, which the bracket branch
   removes.
3. Compare that classification with the range file.

The run is about 1.1M small substitutions, so report its wall time. Add a named mutation: change
the shipped class to `[^\pL\p{Nd}_]`. The Python-only count then goes nonzero and names `²`.

### M2 (major): nothing proves the curated rows reached the golden

Location: plan `:342-348` (R-wb4 curated set), `:379-383` (review focus 1 and 2), `:463-476`
(Task 1 acceptance), `:545-549` (Task 2 states).

**Defect.** The draw's token set (read from `gen.py`) is `a ä Ω ǅ ² Ⅻ U+0301 _ 2`, space, NBSP,
`( [ : . - e r ü ß`, and the three triggers. It holds no `©`, `™`, `€`, newline, or uppercase
letter. Several classes therefore exist in the golden only through the curated rows: the kept
symbols that unidecode maps to letters (R-wb3's stated reason), the newline between two credits
(the only row that tests `.` and `^` against Python), and both cases of each brief input.

No criterion checks that the curated rows are present. A generator that emits only the draw, or
that drops the newline row in serialization, passes every Task 1 and Task 2 criterion. Review
focus 1 and 2 say "pinned by the golden's bracket rows", but that holds only for the draw's
brackets.

**Fold.** Write the curated inputs out in the golden test as a list, keyed by preceding-character
class, and assert that each is present in the golden with beets' output. Alternatively, mark
curated rows in the file (`"curated": true`) and pin their count. Add a state row: "a curated
input is absent from the golden: fails naming it."

### M3 (major): the draw's power has no floor, and the plan does not restate the token set

Location: plan `:337-339` (R-wb4), `:425-426` (Task 0 note 8), `:469-472` (Task 1 acceptance),
`:555-558` (the ASCII-literal control).

**Defect.** Task 0 note 8 says that losing `/tmp/claude-1000/ubound/` is not a blocker "since R-wb4
restates the token set." R-wb4 does not restate it. It says "the brief's token set," and the
brief lists only some of the tokens. It omits `a e r 2 - :`, the two space kinds, and the length
range of 1 to 8 tokens. `/tmp` is tmpfs, so a reboot before Task 1 leaves the implementer to
invent the draw.

The acceptance criteria also give the draw no pass-fail floor. The fired count is "expected in the
low hundreds," which is advisory, and the ASCII-literal control pins "that count (reported by the
task)," requiring only that it be nonzero. A weak draw with 3 fired rows and 1 ASCII-literal
divergence satisfies every criterion. The golden's coverage is set by whatever was drawn, not by
what the plan intended.

**Fold.**

- Write the token list, the length range (`randint(1, 8)`), and the seed (907) into R-wb4
  verbatim.
- Set floors in Task 1's acceptance and in Task 2's golden-test states. At 10,000 draws the
  brief's rate predicts about 240 fired rows and about 90 ASCII-literal divergences before
  dedupe, so floors of fired ≥ 150 and ASCII-literal divergences ≥ 40 would work. The test pins
  the exact values measured and fails below either floor.
- Keep dedupe order deterministic by insertion order, never by set iteration. Task 1's two-run
  check already catches a violation.

### m1 (minor): mutation 4 is ambiguous and one reading is a no-op

Location: plan `:582` (Task 2 mutation 4), `:319-323` (R-wb3).

"Restore the capturing trigger group" does not say which branch. A capture restored in the second
branch only leaves `(^|[^\pL\pN_])` as group 1. `${1}` is then still correct, so the mutation
survives. R-wb3's warning holds only when the first branch's trigger captures. **Fold:** name the
mutation "make the first branch's trigger capturing, `[\[\(](featuring|feat|ft)`." Its expected
failure is the golden's bracket rows (the trigger is re-inserted) and every second-branch row
(group 1 is now unmatched, so the kept character is dropped).

### m2 (minor): the grep for the gap note can pass vacuously

Location: plan `:569-570`.

A reworded gap note without "Rüft" passes the grep. The new doc comment may also legitimately
quote `Rüft`, which would fail the criterion for no reason. **Fold:** grep for the sentence being
replaced instead (`must stay byte-identical` and `No committed case carries such a string`). Let
the diff-reviewer read the new comment against R-wb2's three required elements.

### m3 (minor): two exit-gate items have no named check

Location: plan `:707-711`.

Items 4 and 5 are assertions with no command behind them, and the exported-unused leg only warns.
**Fold:** make item 4 a `git grep -nF` over `*_test.go` for the rewrite's text, expecting exactly
the translation-map hit. Make item 5 a `go doc -all ./internal/engine/matcher` diff between the
base commit and HEAD, expecting one added line (the field).

### m4 (minor): `ties.json` has no state row for "more than one row changed"

Location: plan `:621-623`, `:651-652`, `:794-795`.

The plan states that the `ties.json` diff is one row, but it also lists "whether `-update-ledgers`
re-derives every case" as unverified. The README says the re-derive "is expected to be a no-op
unless the venv's pins moved." If other rows move, the task has no stop rule and might hand-trim
the file, breaking the privileged-field row for ties. **Fold:** add the state "the regeneration
changes any row other than the new case's: stop and report the diff; never hand-trim."

### m5 (minor): Task 3 leaves the default-table result unpinned where it could be expected

Location: plan `:653-654`, R-wb5 `:356-363`.

The new title carries no bracket, so `SDPatternsNoBracketDiscount` and `BeetsSDPatterns` score it
identically once both carry the translated row. `gradeCaseDefault` re-scores the committed
candidate under the default table, so the case is expected not to appear in `VerdictMoves`. "May
not; reports either way" gives up a cheap corpus-level check of the default table. **Fold:** state
the expectation as "absent from `VerdictMoves`; present means the default table's featuring row
diverges, so stop and report." Keep R-wb5's point that the harness itself fails nothing.

### m6 (minor): Task 3 mutation 1 predicts the wrong failure message

Location: plan `:660-661`.

With the literal restored, Go scores the case strong (about 0.0087). That takes the strong-match
early return and skips the search the case logged. Compat grading can then report the verdict or
the request set before it reports the `album` penalty. **Fold:** "the case fails; report the
first failure message (album penalty, verdict, or request set)."

### m7 (minor): the "no existing expected value changes" criterion contradicts Task 2's own edits

Location: plan `:566-568` against `:515-525`.

Task 2 rewrites `TestBeetsSDPatternsMatchBeetsLiterals`, removes
`TestBeetsSDPatterns_SixEntriesInBeetsOrder`, and reroutes
`TestSDPatternsNoBracketDiscount_OnlyBracketPatternsRemoved`, so "every existing test passes
unchanged" cannot hold as written. **Fold:** scope the criterion to figures: "no existing
`StringDist`, component, or corpus expected figure changes; the three pattern-table tests change
only as the parity-test outcome states."

### m8 (minor): some privileged-channel refusals and Task 1 states lack an enforcer or a proof

Location: plan `:282` (the golden's privileged-field row), `:456-461` (Task 1 states), `:629-631`
(floors).

- **The golden header.** The header check does not refuse a hand-edited row, because the header
  survives the edit. **Fold:** restate the refusal as "the diff-reviewer re-runs the generator and
  compares `sha256sum` with the report." Task 1 already pastes both hashes.
- **Task 1 states.** The states for beets-version mismatch and an absent output directory have no
  named mutation. The consequence is low, since the generator is hand-run Python. **Fold:** add a
  mutation that changes the expected beets version, or accept the gap in one line.
- **The floors.** The floors are growth-tolerant, so the census passes whether or not they rise.
  Mutation 2 proves the new values are exact. **Fold:** have the acceptance criteria paste each
  constant's old and new values beside the census output.

## Owner forks

None. Every finding here resolves inside the approved scope without a product or priority choice.
