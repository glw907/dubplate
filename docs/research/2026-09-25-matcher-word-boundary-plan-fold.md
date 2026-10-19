# Matcher word-boundary plan: fold 1 record

Subject: `docs/plans/2026-09-25-matcher-word-boundary.md`,
revision 1 folded to revision 2 (status "reviewed, fold 1; awaiting
verification"). Inputs: the contract, mechanics, and risk lens reviews at
`docs/research/2026-09-25-matcher-word-boundary-plan-review-{contract,mechanics,risk}.md`.
Every finding was checked against the tree at `1306ae3` or against a
reviewer's recorded probe before it was folded. The fold wrote only the plan
and this record; it committed nothing and ran no gate.

Checks the fold ran itself: `scripts/check.sh:385-387` (the ruff leg's
tracked-files comment and `git ls-files '*.py'`);
`/tmp/claude-1000/ubound/gen.py` (seed 907, `randint(1, 8)`, 23 tokens, the
second space U+00A0); `go doc regexp` (the "successive non-overlapping
matches" sentence names the `Find` family, not `ReplaceAll`);
`go doc strings.ToLower`; `strdist_test.go:125-133` and `:218-228`;
`tunables.go:318-376`; `derive_defect_free.py:30-37`; the five
`*Ledger*` tests under `internal/harness/oracle`; and
`git grep -nw -e 215 -e 214 -- '*.go'`, which also found
`internal/engine/matcher/verdict_test.go:86`, a site the risk lens did not
list.

## Counts

- 25 distinct findings after merging 5 duplicate pairs (contract 11,
  mechanics 9, risk 10).
- Folded: 23.
- Refused: 0.
- Owner fork: 1 (risk M1, as Ruling 3).
- Owed to the conductor: 1 (risk m8).
- Reviewer notes outside the counts: 4 (two folded as wording, two left
  as they stand).

## Dispositions

| ID | Disposition | Where |
| --- | --- | --- |
| contract M1 | Folded. The class test drives every code point through the shipped row on `"a" + r + "ft. x"`; it never compiles its own class copy. Adds a "neither form" state row and mutation 8 (`[^\pL\p{Nd}_]` names `²`). | Ruling 2 text; Task 2 parity-tests outcome, states table, mutation 8 |
| contract M2 | Folded, by the reviewer's second option. The golden carries a `curated` array of `[label, input, beets_output]` rows; the test writes out the label list and fails on a missing label. | R-wb4; Task 1 outcome and acceptance; Task 2 states; Review focus 3 |
| contract M3, mechanics m5 | Folded once. R-wb4 writes out the seed, the length range, and the 23 tokens in order, plus the no-post-15.1-code-point constraint (docstring). Floors: fired ≥ 150 (Task 1 stops; Task 2 fails) and ASCII-literal divergence ≥ 40 (Task 2 fails); Task 2 pins the exact counts. Task 0 note 8 now points at the written-out draw. | R-wb4; Task 0 note 8; Task 1 states and acceptance; Task 2 states; exit gate 2 |
| contract m1, mechanics M1 | Folded once. R-wb3 states group numbering by opening parenthesis, the branch-1 failure (6,351 divergences), the branch-2 no-op (0), and why branch 2 stays non-capturing. Mutation 4 names branch 1's capture and warns that branch 2's is not the mutation. | R-wb3; Task 2 mutation 4 |
| contract m2 | Folded. The grep now targets the replaced sentences (`must stay byte-identical`, `No committed case carries such a string`), and the diff-reviewer reads the new comment against R-wb2. | Task 2 acceptance |
| contract m3 | Folded. Exit gate 4 is a `git grep -nF` over `*_test.go` expecting only the translation-map line; exit gate 5 is a `go doc -all` diff between base and HEAD. | The exit gate |
| contract m4 | Folded. "Any other `ties.json` row changes: stop; never hand-trim." | Privileged-fields table; Task 3 states |
| contract m5 | Folded. The case is expected absent from `VerdictMoves`; present stops the task. R-wb5 keeps its point that the harness fails nothing. | R-wb5; Task 3 states and acceptance |
| contract m6, risk m4 | Folded once. Mutation 1 accepts any compat failure on the case and reports which rule fired, with the strong-match early-return reason. | Task 3 mutation 1 |
| contract m7 | Folded. The criterion is scoped to expected figures; the three pattern-table tests change only as the parity outcome states. | Task 2 acceptance |
| contract m8 | Folded, with one change. Header refusal: the diff-reviewer compares `sha256sum` of each committed golden with the two hashes Task 1 pastes, rather than re-running the generator, because the reviewer is read-only and the generator writes into `testdata/`. Beets-version state: mutation 3 added. Floors: old and new values pasted beside the census. | Privileged-fields table; Task 1 mutation 3; Task 3 acceptance |
| mechanics M2, risk m1 | Folded once. Task 1 runs `ruff check` and `ruff format --check` on the generator directly and pastes both outputs; Global constraints quotes `check.sh:385-387` for why the gate cannot see an untracked file. Chosen over `git add -N` because the direct run needs no index change in a shared tree and its behavior is certain. | Global constraints; Task 1 acceptance |
| mechanics m1 | Folded. R-wb4 names `from beets.autotag.distance import SD_PATTERNS` and quotes the trap. | Behaviors quoted; R-wb4; Task 1 outcome |
| mechanics m2 | Folded. `dict.fromkeys`, first-draw order, never a `set`. | R-wb4; Task 1 outcome |
| mechanics m3 | Folded. `ensure_ascii=False`, one row per line; the measured sizes are quoted. | R-wb4; Task 1 outcome |
| mechanics m4, risk m2 | Folded once. Task 3 runs `-run TestTieLedgerMatchesTheCorpus -update-ledgers` and checks `album_info.json` and `file_tags.json` unchanged; the "Not verified" item on derivation scope is removed, since `ties_derive_test.go:54-80` answers it. | Privileged-fields table; Task 3 outcome, states, acceptance; Not verified |
| mechanics m6 | Folded, by the one-comment edit. Task 2 updates `TestStringDistFeaturingNeedsAWordBoundary`'s doc comment and the table-parity sentence in `derive_defect_free.py:34-36`. The Python file is ruff-gated and outside no wall, and an edit now costs less than a carried stale claim. | Global constraints; Task 2 Files and outcome; Contested surfaces |
| mechanics m7 | Folded. Mutation 7 swaps the branches; R-wb3 and Review focus 2 state the order (827 divergences). | R-wb3; Review focus 2; Task 2 mutation 7 |
| risk M1 | Owner fork, Ruling 3. The plan executes under either answer: yes declares one pinned U+0130 row and narrows the Goal; no fixes it in `StringDist` as a second departure. Both answers correct the sweep comment at `strdist_test.go:222-225` and put the divergence in Task 4's HISTORY list. | Rulings 3; Goal; Behaviors quoted; Task 2 outcomes, states, mutation 9; Task 4 |
| risk m3 | Folded. Task 3 rewrites family 4's defining sentences and moves the structural term to 62 in the `:1917` sum and in `:1572`'s pass-2b attribution. | Task 3 README outcome |
| risk m5 | Folded. The `DefaultTunables` check at `strdist_test.go:130-132` is named and routes through the map; it is not weakened. | Task 2 parity-tests outcome |
| risk m6 | Folded, with one more site. Task 3 sweeps case-count comments; the fold's own grep added `verdict_test.go:86`. | Global constraints; Task 3 Files and outcome |
| risk m7 | Folded. The package doc's non-overlap sentence and `strings.ToLower` are quoted. The quote names the `Find` family only, so the plan names the two-trigger and newline curated rows as the proofs. | Behaviors quoted |
| risk m8 | Owed to the conductor at the close (below); also written as conductor action 5. | The conductor's actions |
| risk m9 | Folded. The Go-only count failure message states the re-read procedure. | Ruling 2 text; Task 2 states |

### Reviewer notes outside the counts

- Risk's tenth item (`shadow.configHash` moves at the next deploy): folded
  as a line in Task 4's HISTORY draft.
- Mechanics note on the per-item overrides: folded as wording (three
  `field` overrides, index 0 to 2) in Task 3.
- Mechanics note on the field name (`SDPattern.Replacement` is a template,
  `SDReplacement.Replacement` is literal): left as is. R-wb1 already
  requires the doc comment to say so; a rename is a naming choice the
  implementer and the architecture read can still take.
- Mechanics note on the untested lowercase non-ASCII `release:` query term:
  left as is. Task 3's `path_query` stop row already covers it.

## Rulings awaiting owner

The plan's "Rulings awaiting owner" section holds all three. One line each:

1. A 2.75M execution ceiling (flag 2.2M) against a 2.12M projection?
   Recommend yes. Revision 1 asked 2.5M against 1.95M; the fold's added
   proofs raised Tasks 1 to 3. No holds 2.5M and asks after Task 4,
   before the merge-side reads.
2. Carry Python's whole `\w` class as 749 ranges and sweep every code point
   through the shipped row, pinning 0 Python-only and 9,039 Go-only?
   Recommend yes. Unchanged question; the fold changed only how the test
   proves it (contract M1) and its failure message (risk m9).
3. Declare the U+0130 divergence with a pinned row and leave it unfixed?
   Recommend yes. No fixes it now as a second named departure in
   `StringDist`. Filing an IDEAS entry is not recommended.

## Ceiling and deliverables

Projection 2.12M: Task 0 0.08M, Task 1 0.27M, Task 2 0.50M, Task 3 0.50M,
Task 4 0.12M, one fix round 0.20M, the simplifier round 0.15M, two
architecture reads 0.25M, the merge 0.05M. Ruling 3's no adds about 0.03M.
Deliverables per task stay at or below four under every answer: Task 1
three (two on Ruling 2's no), Task 2 four, Task 3 four, Task 4 two. The
U+0130 row joins Task 2's unit rows, its fix joins the loop deliverable, and
each stale comment joins the deliverable whose code it describes.

## Owed to the conductor

- At the close, add `internal/engine/matcher/testdata/gen_featuring_golden.py`
  to both CLAUDE.md passages that list tracked Python ("Gate" and "Stack
  rules: Python"). The fold did not edit CLAUDE.md.
- Record the three rulings' outcomes before Task 0 (conductor action 1 now
  says three).
- Revision 2 needs its verification read before approval. The plan, this
  record, and the three lens reviews are uncommitted; the conductor commits
  them.
- Planning spend for this fold is outside the execution ceiling, as Ruling 1
  states.

## Fold 2

Revision 2 folded to revision 3 (status "reviewed; awaiting owner approval
(Rulings 1-3)"). Inputs: the verification read
(`2026-09-25-matcher-word-boundary-plan-fold-verification.md`, findings
V1-V4) and the prose review
(`2026-09-25-matcher-word-boundary-plan-prose-review.md`, fact findings
P1-P10, one blocker, four warnings, six suggestions). Each finding was
checked against the tree at `b9e281b` before it was applied. The fold wrote
only the plan and this record, committed nothing, and ran no gate. Task
scope, the ceiling, and each ruling's substance are unchanged.

Counts: 25 findings, 25 applied, 0 refused.

Checks run: `go help doc` ("-short: One-line representation for each
symbol. Cannot be combined with -all."); `go doc -short
./internal/engine/matcher`, which prints `type SDPattern struct{ ... }` with
the fields elided; `go doc matcher.SDPattern`; a reproduction of R-wb4's
draw in the pinned venv (7,955 distinct, 251 fired, 92 `re.ASCII`
divergences, 234,239 bytes for the rows joined by newlines without the
7,954 separating commas, which is consistent with the verification's
242,193); `tunables.go:324`, `compat.go:39-53`, `case.go:530-540`;
`strdist_test.go:110-133`; `go doc regexp` (the "'All' variants return an
iterator" sentence) and `go doc regexp.Compile` (the sentence continues
past "use"); the 6a "Both budgets" entry in `docs/HISTORY.md`; and the
mechanics lens's token list, size table, and category counts.

| Finding | Disposition |
| --- | --- |
| V1, exit gate 5 | Applied. The gate now compares `go doc -short` listings, quoted from `go help doc`, and diffs `go doc matcher.SDPattern` for the field. It allows the doc text that describes the field, since the type's doc says the pattern is "deleted from both inputs". |
| V2, stale draw figures | Applied. R-wb4 and Task 1 acceptance quote 7,955, 251, about 92, and 242,193 bytes, reproduced by this fold. |
| V3, "second instance" | Applied. Dropped from Ruling 3; Task 4's HISTORY list gains the one-code-point sweep result. |
| V4, Goal "every input" | Applied. The Goal excludes code points first assigned after Unicode 15.1 (Ruling 2). |
| P1, "the default table" | Applied. R-wb5 and Review focus 4 name `SDPatternsNoBracketDiscount` and `BeetsSDPatterns`. |
| P2, row 2 vs row 3 | Applied. The closing check is described as `StringDistPatterns[2].Pattern`, the featuring row. |
| P3, package-doc gloss and Compile quote | Applied. The sentence names the `All` iterator variants; the Compile quote ends with "...". |
| P4, ceiling basis | Applied. The close side lists its six parts. |
| P5, "not verified" token list | Applied. It names 36 tokens with U+088F and the 452 rows. |
| P6, size figure encoding | Applied with V2. |
| P7, chain includes Task 0 | Applied in the header note and the execution model. |
| P8, contested-surfaces row | Applied. |
| P9, mutation count | Applied: "seven to nine". |
| P10, `LC` label | Applied: "`Lu` and `Ll` 106", as the mechanics lens measured. |
| Blocker, "underscore" | Applied. The quote stops at `str.isalnum()`, followed by "plus `_`". |
| Warning, the comment-edit run-on | Applied as three sub-bullets. |
| Warning, the class-test sentence | Applied as proposed. |
| Warning, the parity-tests bullet | Applied as four sub-bullets. |
| Warning, the ruling paragraphs | Applied. Each ruling starts a paragraph at "What each answer builds", and Ruling 3's yes and no are separate sentences. |
| Suggestion, "In one line:" | Applied. |
| Suggestion, title case | Applied. |
| Suggestion, the combining mark in `re.match` | Applied as the `"́"` escape, since the literal character is invisible. |
| Suggestion, mutation 2 wording | Applied. |
| Suggestion, status line | Applied through the revision 3 status. |
| Suggestion, "not merely extended" | Applied. |

tellgrader `--register docs` on revision 3: 0 findings.
