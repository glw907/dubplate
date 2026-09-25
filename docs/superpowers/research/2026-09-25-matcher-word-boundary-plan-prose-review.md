# Prose review: matcher word-boundary plan (revision 2, `a3773d1`)

Reviewer: `prose-voice-reviewer` (read-only; the conductor saved its report). Register: Google
developer docs. tellgrader `--register docs`: 8,111 words, 430 sentences, one hard-slop hit
(`underscore`, `:215`), cadence CV 1.25; the docs-register profile did not resolve, so no
measures. Every code pointer resolved at `1306ae3`. The verification read's four minors are
present as it describes (`:917-919`, `:437-438`/`:587-588`, `:108-109`, `:231-232`) and are not
repeated here.

## Facts and contradictions, ranked by consequence

1. **"The default table" names two tables** (R-wb5 `:462-466`, Review focus 4 `:498`).
   `DefaultTunables` selects `BeetsSDPatterns` (`tunables.go:324`), but `ModeDefault` runs
   `SDPatternsNoBracketDiscount` (`case.go:535`) and `ModeCompat` runs `BeetsSDPatterns`
   (`compat.go:48-50`). Rewrite `:462-466`: "`ModeDefault` re-scores under
   `SDPatternsNoBracketDiscount` and reports a moved verdict as a `VerdictMoves` entry that
   fails nothing (`case.go:672-680`). The case therefore pins the featuring row only as
   `BeetsSDPatterns` carries it. The three unit rows and the translation-map test run both
   tables, and they pin `SDPatternsNoBracketDiscount`'s row." Rewrite `:498`:
   "`SDPatternsNoBracketDiscount` drifting from `BeetsSDPatterns` on the shared row."
2. **One row, two numbers.** `:634` says "Row 3" (one-based); `:655` says "row 2's `Pattern`"
   (zero-based). Rewrite `:655`: "which compares `StringDistPatterns[2].Pattern`, the featuring
   row, with the literal and would red after the translation".
3. **Package-doc gloss wrong; Compile quote cut.** `:192`: the quoted `go doc regexp` sentence
   is about the `All` iterator variants, not the `Find` family. Rewrite: "The sentence names
   the `All` iterator variants. `ReplaceAllString` is not one of them, and its own doc does not
   restate the rule." `:178`: end the quote with "...implementations use ..." to mark the cut.
4. **Ceiling basis misstates its source** (`:256-257`). 6a's close carried two architecture
   reads (about 0.2M) plus the tidy fix, the simplifier round, Task 8, the merge ritual, and a
   merge re-review. Rewrite: "and a close side of about 0.92M: two architecture reads (about
   0.2M) plus the tidy fix, the simplifier round, Task 8, the merge ritual, and a merge
   re-review. This pass's tasks are smaller than 6a's, and its close has no tidy fix or
   re-review."
5. **"Not verified" token list drops U+088F** (`:1001-1004`). Rewrite: "(seed 907, 36 tokens,
   among them `©`, `™`, `€`, `\n`, `)`, `]`, `é`, U+203F, and U+088F, a code point first
   assigned in Unicode 16.0), and found zero differences for the rewrite outside the 452
   U+088F rows."
6. **Size figure from the wrong encoding** (R-wb4 `:414-416`). Rewrite: "the mechanics lens
   measured the default ASCII-escaped encoding at 348,821 bytes for a comparable draw, and the
   verification measured 242,193 bytes for this draw one row per line, under the 300 KB cap."
   Fold with the verification's finding 2.
7. **Chain description includes Task 0** (`:46-49`, `:960-962`). Rewrite `:960`: "Tasks 1 to 4
   each run `dubplate-implementer` ... Task 0 is a read-only `sonnet` dispatch whose report the
   conductor reads directly."
8. **Contested-surfaces row incomplete** (`:927`). Rewrite: "Tasks 1 and 2; Task 3, one
   case-count comment".
9. **Mutation count off by one** (`:263`). Rewrite: "seven to nine mutation proofs".
10. **Category label** (`:80`). `LC` covers `Lt` too; rewrite as "`Lu` and `Ll` 106".

## Blocker

- `:215` (tellgrader slop-hard "underscore", inside a paraphrase of the Python `re` docs).
  Rewrite: `\w` "matches Unicode word characters; this includes all Unicode alphanumeric
  characters (as defined by `str.isalnum()`)", plus `_`, and beets passes no `re.ASCII` flag.

## Warning

- `:640-648`, a 70-word run-on: split the three comment edits into sub-bullets
  (`strdist_test.go:135-142`, `strdist_test.go:222-225`, `derive_defect_free.py:34-36`).
- `:666-672`, the 80-word class-test sentence. Rewrite: "On Ruling 2's yes, a class test reads
  the range file. It classifies every code point `r` outside the surrogates through the
  shipped row itself, never a copy of its class. For each `r`, it substitutes on
  `"a" + string(r) + "ft. x"`. An unchanged output marks `r` a Go word character, and the
  output `"a" + string(r)` marks it a non-word character. The bracket branch removes `(` and
  `[`, so their expected output is `"a"`. The test compares the result with the range file."
- `:649-673`, the parity-tests bullet (264 words, four tests): one sub-bullet per test.
- `:71-93`, `:95-116`, the ruling paragraphs: start a new paragraph at "What each answer
  builds"; in Ruling 3, give yes and no their own sentences.

## Suggestion

- `:276`: cut "In one line:"; "A 2.12M projection sits under a 2.2M flag, which is 80% of a
  **2.75M** ceiling."
- `:1`: sentence-case title, "Matcher word-boundary pass: implementation plan".
- `:217`: write `re.match(r"\w", "́")` as `re.match(r"\w", "́")`, matching R-wb4 `:424`.
- `:599-600`: "Drop the seed: a second run writes different bytes. If the generator cannot run
  unseeded, the report marks this mutation reviewer-read."
- `:3`: the status line moves past "awaiting verification".
- `:808`: cut the contrast frame "not merely extended by a sentence".

Verdict: the plan reads as a meticulous engineer's, written for implementers. The real risks
are a few long sentences, the ambiguous "default table", and a handful of specifics that drift
from their cited sources.
