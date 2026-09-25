# Matcher word-boundary pass: implementation plan

Status: **reviewed; awaiting owner approval (Rulings 1-3).** Revision 3,
2026-09-25. Three lens reviews ran (contract, mechanics, risk:
`docs/superpowers/research/2026-09-25-matcher-word-boundary-plan-review-*.md`),
and fold 1 applied them. A verification read
(`2026-09-25-matcher-word-boundary-plan-fold-verification.md`) and a prose
review (`2026-09-25-matcher-word-boundary-plan-prose-review.md`) followed,
and fold 2 applied both. The fold record
(`docs/superpowers/research/2026-09-25-matcher-word-boundary-plan-fold.md`)
gives one disposition per finding for both folds. Three rulings for Geoff
are open ("Rulings awaiting owner"), and the plan executes under either
answer to each.

Source: the research brief
`docs/superpowers/research/2026-09-24-matcher-unicode-word-boundary.md`,
which Geoff ruled yes on 2026-09-24 (`docs/STATUS.md`, "Next action", item
3). The ruling builds five things: a documented Go translation of beets'
featuring row, a translation-map test, a parity golden with its ruff-gated
Python generator, the `strdist-feat-unicode-boundary` corpus case, and three
unit rows. This plan builds exactly that set, plus the two proofs Ruling 2
asks about and the one declared divergence (or fix) Ruling 3 settles.

**Code pointers are resolved at `1306ae3` on master.** A lane 1 executor was
committing `web/`, `cmd/`, and `internal/serve` while this plan was drafted.
`git log f0eec0d..1306ae3 -- internal/engine/matcher internal/harness corpus`
is empty, so every pointer below reads the tree the executor left alone.
Task 0 re-resolves each pointer against the real HEAD.

**What this pass is.** The matcher's featuring row scores a title the way
beets does when a non-ASCII letter or number sits directly before `ft`,
`feat`, or `featuring`. Go's `\b` is ASCII-only and Python's is Unicode-aware,
so today Go sees a word boundary inside "Geschäft" and strips the rest of the
title as a credit. The fix is one rewritten regular expression with a kept
capture group, carried beside the unchanged beets literal and proved against
a committed beets-derived golden.

**What this pass is not.** It changes no other `SD_PATTERNS` row, and it
leaves the `CompilePatternOption` configured-pattern class documented only, as
the brief recommends: operators write that text, so no translation applies.
It touches no `internal/engine/musicbrainz` pattern and no `album_media.go`
`\d`. It adds no module dependency, runs nothing Python in the gate, moves no
oracle pin, and edits no `corpus/manifest.json`. It writes nothing under
`web/`, `cmd/`, `internal/serve`, `internal/engine/store`, `internal/ops`,
`internal/backup`, `systemd/`, `scripts/`, `env/`, `config/`, `Makefile`,
`tools/venv/`, or `_salvage/`. **No deploy of any kind.** A task that believes
it needs any of these stops and reports.

> For agentic workers: per-task Agent dispatch (five tasks, under the
> workflow threshold, and no task is independent of its predecessor). Tasks
> 1 to 4 each run `dubplate-implementer` (`sonnet`, effort `high`), then
> `diff-reviewer` (`claude-opus-5-5`, effort `high`), then the gate inside
> the chain. Task 0 is a read-only `sonnet` dispatch whose report the
> conductor reads directly. The conductor consumes structured reports and never reads a
> diff, a test log, or a gate transcript inline. Subagents never dispatch
> agents, so `code-simplifier` is the conductor's dispatch. The conductor owns
> `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, and the lane registry.

---

## Rulings awaiting owner

Three yes-or-no questions. The 2026-09-24 ruling stands and is not reopened.

**1. Does this pass run under a 2.75M execution ceiling?** Recommendation:
yes. Reason: "Token ceiling" prices the pass at a 2.12M projection from 6a's
measured actuals, scaled down for tasks this size. Revision 1 asked for 2.5M
against a 1.95M projection; the fold added the curated-row, draw-floor,
class-through-the-shipped-row, and U+0130 proofs, which raise Tasks 1 to 3.
A 2.75M ceiling puts the 80% flag at 2.2M, just above the projection.
Planning is recorded outside the ceiling, as 6a's Ruling 1 set.

What each answer builds: yes sets the flag at 2.2M. No holds a 2.5M
ceiling (flag 2.0M), which the projection crosses during the merge-side
reads, so the conductor asks after Task 4, before those reads.

**2. Does the golden also carry Python's whole `\w` class as code-point
ranges, so the gate proves the rewrite's character class on every code point
and pins the Unicode-version gap by count?** Recommendation: yes. Reason: the
string-pair golden proves the class only on the characters its draw uses. A
measurement for this plan compared Python's `\w` (CPython 3.13.15,
`unicodedata.unidata_version` 15.1.0) with Go's `[\pL\pN_]` (Go 1.27.1,
`unicode.Version` 17.0.0) over every code point outside the surrogates. No
code point is a word character in Python and not in Go. 9,039 are word
characters in Go and not in Python: letters and numbers (`Lo` 8,827, `Nd` 90,
`Lu` and `Ll` 106, and a few more) that Unicode 16.0 and 17.0 assigned and
15.1 leaves unassigned, starting at U+088F, U+0C5C, and U+0CDC. Python's whole word class
compresses to 749 ranges. The repo already pins a Unicode-version skew this
way: `TestStringDistCaseMappingSweep` (`strdist_test.go:240-290`) declares
its two known divergences so that a third one reds the gate.

What each answer builds: yes adds a second golden file of 749 ranges and a
test that drives every code point through the shipped featuring row (never a copy of its
class) and asserts zero Python-only word characters and exactly 9,039
Go-only ones. A Go toolchain bump that moves the count fails the gate, and
the failure message states the re-read procedure. No drops that file and
that test. The gap then lives only as a comment, and the string-pair golden
proves the class on the drawn characters alone. Neither answer changes what
a real title scores: a divergence needs a code point first assigned after
Unicode 15.1 directly before the trigger.

**3. Does this pass declare the U+0130 divergence and leave it unfixed,
rather than fix it now?** Recommendation: yes (declare). Reason: the risk
lens measured one input the featuring row cannot see. Python's `str.lower()`
maps U+0130 (İ) to `i` plus U+0307, a combining mark `\w` rejects, so beets
sees a boundary before `ft`. Go's `strings.ToLower` maps it to a bare `i`,
so neither the beets literal nor the rewrite fires. Beets scores
`string_dist("İft: Ember Wharf", "İft: Northern Lines")` at `0.06875`; Go
scores it near `0.6875`. A sweep of every code point found U+0130 the only
one where Python's lowercase changes whether the last character is a word
character. The divergence predates this pass, and its carrier is rare: an
İ directly before `ft`, `feat`, or `featuring`. Declaring keeps the pass
inside the 2026-09-24 ruling's five deliverables plus one row.

What each answer builds: yes adds one known-divergence `StringDist` row to
Task 2. The row pins beets' `0.06875` and Go's measured figure, so a fix
reds the gate. Yes also narrows the Goal to name the one upstream
divergence and puts it in Task 4's "wrong to rediscover" list. No (fix now)
adds a second named departure to Task 2: `StringDist` replaces U+0130 with
`i` plus U+0307 before `strings.ToLower` on both sides. The same row then
pins beets' `0.06875` under both tables. The fix leaves `sdStrip` agreeing,
since the strip removes the dot. Filing the fix as an IDEAS entry is not
recommended: the pinned row and the HISTORY line already keep it from being
rediscovered, and IDEAS holds features and member requests.

---

## Primary inputs

Read these before Task 1, in this order.

1. **The research brief**, in full: the measured figures, the differential
   run, option 1's two facts about the beets pattern, the kept group's
   reason, and the proposed corpus case.
2. **`internal/engine/matcher` at `1306ae3`**: `tunables.go:334-342`
   (`SDPattern`), `:363-382` (`BeetsSDPatterns` and its doc comment, with
   the gap note at `:371-374`), `:384-403` (`SDPatternsNoBracketDiscount`);
   `strdist.go:28-65` (`StringDist`, whose loop at `:41-63` calls
   `ReplaceAllString(a, "")` at `:43-44`), `:134-139` (`applySDReplace`,
   whose comment explains why that call is literal), `:146-159`
   (`compileSDPatterns`), and `:181-193` (`sdRegexp`); `strdist_test.go:103-133`
   (`TestBeetsSDPatternsMatchBeetsLiterals`), `:135-162`
   (`TestStringDistFeaturingNeedsAWordBoundary`), and `:212-290`
   (`TestStringDistCaseMappingSweep` with `testdata/case-mapping-sweep.tsv`,
   the precedent for a beets-derived fixture with declared divergences);
   `tunables_test.go:183-213` (two more byte-identity tests on the tables);
   `compat.go:47-53` (`Compat.Apply`, which assigns either table).
3. **The golden-fixture precedents**: `internal/engine/tagio/testdata/gen_golden_flac.py`
   (a generator that is the fixture's provenance record, run by hand under
   the pinned venv, never by the gate) and `tools/unidecode-table/generate.py`
   (the same stance for the embedded table).
4. **The corpus authoring path**: `corpus/README.md` "Family 4:
   string-distance edges" (`:230-266`), "The featuring-pattern false
   positive" (`:1139-1187`, the `adv-ft-*` shape this case mirrors), "The
   synth tool" and "Authoring a synthetic case" (`:2680-2778`), "The freeze
   tool" (`:2780-2810`), and "Regenerating a derived ledger" (`:2248-2267`);
   the sibling case `corpus/cases/adv-ft-draft-colon/`.
5. **The conformance floors and modes**: `internal/harness/conform/case.go:40-45`
   (`CaseFloor` through `TrackVectorFloor`), `:150-172` (`GradeMode`),
   `:530-552` and `:640-682` (`gradeCaseDefault`, which logs a moved
   verdict and fails nothing); `request.go:28-30` (the request floors).

### Behaviors this plan relies on, quoted

From `go doc` at `go1.27.1 linux/amd64`.

`go doc regexp/syntax`:

```
\b             at ASCII word boundary (\w on one side and \W, \A, or \z on the other)
\w             word characters (== [0-9A-Za-z_])
\pN            Unicode character class (one-letter name)
^              at beginning of text or line (flag m=true)
s              let . match \n (default false)
Unicode character classes are those in unicode.Categories,
unicode.CategoryAliases, and unicode.Scripts.
```

`go doc unicode.Version`: `const Version = "17.0.0"`, "the Unicode edition
from which the tables are derived."

`go doc regexp.Compile`: "When matching against text, the regexp returns a
match that begins as early as possible in the input (leftmost), and among
those it chooses the one that a backtracking search would have found first.
This so-called leftmost-first matching is the same semantics that Perl,
Python, and other implementations use ..."

`go doc regexp.Regexp.ReplaceAllString`: "replacing matches of the Regexp
with the replacement string repl. Inside repl, $ signs are interpreted as in
Regexp.Expand."

`go doc regexp.Regexp.Expand`: "a variable is denoted by a substring of the
form $name or ${name} ... A purely numeric name like $1 refers to the
submatch with the corresponding index ... A reference to an out of range or
unmatched index ... is replaced with an empty slice. In the $name form, name
is taken to be as long as possible: $1x is equivalent to ${1x}, not ${1}x."

`go doc regexp` (package doc): "The 'All' variants return an iterator over
successive non-overlapping matches of the entire expression." The sentence
names the `All` iterator variants. `ReplaceAllString` is not one of them,
and its own doc does not restate the rule.
The two-trigger curated row is the proof that the replace path also takes
successive non-overlapping matches, and the newline curated row is the proof
that `[^\pL\pN_]` matches `\n`, which `regexp/syntax` does not state.

`go doc strings.ToLower`: "ToLower returns s with all Unicode letters mapped
to their lower case." The doc does not say which mapping applies to U+0130;
the risk lens measured it: Go lowers U+0130 to a bare `i`, and CPython 3.13
lowers it to `i` plus U+0307 (Ruling 3).

The beets source, read from the pinned venv
(`tools/venv/.venv/lib/python3.13/site-packages/beets/autotag/distance.py`,
beets 2.14.1): `SD_PATTERNS` opens at `:36`, the featuring row at `:39` is
`(r"[\[\(]?\b(featuring|feat|ft)\b[\. :].+", 0.1)`, and `string_dist`
substitutes each row with `re.sub(pat, "", str1)` at `:100`. The table is
read with `from beets.autotag.distance import SD_PATTERNS`. The shorter
`import beets.autotag.distance as d` binds the function named `distance`,
not the module, so `d.SD_PATTERNS` raises `AttributeError` (mechanics lens
probe; `derive_defect_free.py:24-27` documents the same trap).

Python's `re` semantics are quoted from the research brief, not re-fetched:
`\w` "matches Unicode word characters; this includes all Unicode
alphanumeric characters (as defined by `str.isalnum()`)", plus `_`, and
beets passes no `re.ASCII` flag. This plan confirmed the
behaviors it relies on in the pinned venv (CPython 3.13.15) on 2026-09-25:
`re.match(r"\w", "\u0301")` is false (a combining mark), `re.match(r"\w",
"²")` and `re.match(r"\w", "Ⅻ")` are true, and `re.match(r"\w", "‿")` is
false (connector punctuation other than `_`).

The rewrite, checked by a scratch Go program for this plan (outside the
tree, under this session's scratchpad), gives the beets output on each of
these strings: `geschäft: ember` unchanged; `x\nft. y\nft. z` to `x\n\n`;
`©ft. x` to `©`; `a_ft. b` unchanged; `é(ft. x)` to `é`; `rüft: a` and `ωft b`
unchanged. Each beets output was read from the venv the same day.

---

## Goal

After this pass, the featuring row matches pinned beets on every input it
sees outside code points first assigned after Unicode 15.1 (Ruling 2), in
both pattern tables, so `StringDist` scores a title with a
non-ASCII letter or number before a featuring trigger as beets does. One
case-mapping divergence upstream of the row, U+0130, is either declared by a
pinned test (Ruling 3's yes) or fixed (its no). The gate proves the translated row against a committed
beets-derived golden without running Python. A reader who opens
`tunables.go` finds the one departure from beets' literal named, with its
reason, where the table is.

---

## Owner touchpoints

The three rulings above, at the plan-approval gate. Nothing else is expected.
What would become one: the corpus case diverging after the fix (the gate
prints an adjudication row a human must write), a new module dependency, or
any change to what the box runs.

---

## Token ceiling: 2.75M execution (Ruling 1)

The basis is 6a's recorded actuals (`docs/HISTORY.md`, the 6a entry's "Both
budgets"): about 0.41M per implementation task, Task 0 about 0.11M, and a
close side of about 0.92M: two architecture reads (about 0.2M) plus the tidy
fix, the simplifier round, Task 8, the merge ritual, and a merge re-review.
This pass's tasks are smaller than 6a's, and its close has no tidy fix or
re-review.

1. Task 0, a staleness read at `sonnet`: **0.08M**.
2. Task 1, the generator and golden (Python and data only): **0.27M**,
   with the curated labels, the draw floors, and a third mutation.
3. Task 2, the translation and its tests: **0.50M**, above 6a's per-task
   average, since it carries four deliverables, seven to nine mutation
   proofs, the class sweep through the shipped row, and the U+0130 row.
4. Task 3, the corpus case: **0.50M**, above the average for the synth,
   freeze, and oracle ledger runs, whose wall time is long, plus the README
   family rewrite and the case-count comment sweep.
5. Task 4, the close drafts: **0.12M**.
6. One fix round: **0.20M**.
7. The simplifier round with its one review: **0.15M**.
8. The merge-side reads: two `go-architecture-reader` dispatches
   (`internal/engine/matcher`, `internal/harness/conform`) at **0.25M**, and
   the merge ritual at **0.05M**.
9. **The projection is 2.12M.**

A 2.12M projection sits under a 2.2M flag, which is 80% of a
**2.75M** ceiling. Ruling 3's no (fix now) adds about 0.03M to Task 2 and
stays under the flag.

Checkpoint: STATUS after Task 2 (the interval is two tasks here, since the
pass is short and Task 3 is the costliest) and at the close. Every task
reports its spend, its mutation outcomes, and its measured gate wall time.

Deliverable count: thirteen across Tasks 1 to 4, plus Task 0's report. Task 1
carries three (two on Ruling 2's no), Tasks 2 and 3 carry four each, and
Task 4 carries two. No task passes four under any answer to the rulings: the
U+0130 row joins Task 2's unit rows, its fix (Ruling 3's no) joins Task 2's
loop deliverable, and the stale comments join the deliverable whose code they
describe.

---

## Global constraints

- **The main tree, as lane 1, after 5c-ii closes.** CLAUDE.md: "Corpus and
  harness work never moves to a worktree." The pass launches only when the
  lane registry shows lane 1 closed; it may run beside a lane 2 worktree on
  6b-ii, whose territory (`internal/engine/store`, `internal/ops`,
  `internal/backup`, `cmd/dubplate-server`) is disjoint from this pass's.
- **Package territory, stated as a wall.** The pass edits
  `internal/engine/matcher/tunables.go`, `strdist.go`, `strdist_test.go`,
  `tunables_test.go`, and `verdict_test.go` (one case-count comment), and
  creates files under
  `internal/engine/matcher/testdata/`. It edits
  `internal/harness/conform/case.go` and `request.go` for floor constants
  and their comments only, and edits comments only in
  `internal/harness/conform/case_test.go`, `internal/harness/conform/defaults.go`,
  `internal/engine/tag/tag_test.go` (the case count), and
  `internal/harness/oracle/derive_defect_free.py` (the table-parity sentence).
  It creates `corpus/cases/strdist-feat-unicode-boundary/` and edits
  `corpus/analysis/ties.json` (regenerated, never hand-edited),
  `corpus/analysis/proximity.json` (only if the gate names a row owed), and
  `corpus/README.md`. Task 4 writes drafts only.
- **Python stays workstation-only.** The generator runs by hand under
  `tools/venv/.venv`. Neither the gate nor `make derive` invokes it. The
  gate's ruff leg lints tracked files only: `scripts/check.sh:385-387` reads
  "Tracked files only, so an unstaged experiment does not gate" and then
  `mapfile -t py_files < <(git ls-files '*.py')`. The implementer's gate runs
  before its commit, so Task 1's gate cannot see the new generator. Task 1
  therefore runs `ruff check` and `ruff format --check` on the generator
  directly (ruff reads the repo's `ruff.toml` by discovery) and pastes the
  output; from Task 2 on, the committed file rides the gate's leg. The
  implementer invokes `python-conventions` before writing it.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar. **The em
  dash is banned in Go comments.** No process citation (rung, task, ruling,
  plan line, or research path) in a shipped Go comment; the staged
  process-citation leg warns on one.
- **Assertion discipline at the repo's bar.** Every hard-coded expectation
  cites the pinned venv, the golden, or a figure this plan quotes. Every test
  names the fixture state it needs, and every golden test carries a
  non-vacuity count, so it cannot pass on an empty or truncated file.
- **Universality.** No estate hostname, path, or personal name in any new
  file. Corpus fixtures use invented names and an invented `mb_albumid`.
- **The gate is** `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`,
  green at every task end, re-issued on exit 75 until it prints `gate exit:`.
  The gate launches a browser leg for `web/`, so it runs in the default
  (heavy) lane. Evidence runs use `-count=1` and report wall time. Commit
  specific files only, footer `Co-Authored-By: Claude Sonnet 5
  <noreply@anthropic.com>` or the dispatch's footer.

---

## Privileged fields and their channels

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| `SDPattern.Replacement` on a shipped row | The two shipped tables' literals in `tunables.go`, and only on the translated featuring row. | The translation-map test (Task 2) fails when any other shipped row carries a non-empty value. A table assigned from outside the package sets its own rows; its zero value keeps beets' empty substitution. |
| `Tunables.StringDistPatterns` | `DefaultTunables` (`tunables.go:324`) and `Compat.Apply` (`compat.go:47-53`), plus tests. | Unchanged by this pass; no config key reads it. |
| The featuring golden and the class-range golden | The generator alone, run by hand under the pinned venv. | The golden test reads the header's beets version, literal, and Unicode version, and fails when any differs from the values the test writes out. The header survives a hand-edited row, so the diff-reviewer also runs `sha256sum` on each committed file and compares it with the two hashes Task 1's report pastes. |
| A case's `expected` record | `harness freeze` alone (`corpus/README.md:2782-2789`). | Hand edits are refused by the README's rule and by the diff-reviewer. |
| A `ties.json` row | The oracle-tagged regeneration (`-run TestTieLedgerMatchesTheCorpus -update-ledgers`). | The gate's tie-ledger check fails on a missing or hand-shaped row (`internal/harness/analysis/ties.go:118`). A regeneration that moves any other row stops the task; nothing hand-trims the file. |

---

## Resolutions this plan takes rather than discovers

A task landing the surface a resolution decides writes nothing about the
resolution in a shipped comment; the close maps each resolution to its
landing commit.

### R-wb1: the row carries an Expand template

`SDPattern` gains an exported `Replacement string` field, read by the
`StringDist` loop in place of the literal `""` at `strdist.go:43-44`. It is a
`Regexp.Expand` template (Go's `${1}` syntax, never Python's `\1`), and every
row whose beets replacement is the empty string leaves it zero. The brief's
alternative, a keep-group index, needs code of its own around the regexp;
the template uses the stdlib directly. The field's doc comment says it is a
Go template and why it is empty on every beets row, beside the existing
reason `applySDReplace` gives for staying literal (`strdist.go:128-133`).

### R-wb2: the literals and the translation map live in the test file

The beets literal table stays written out in `strdist_test.go`, where
`TestBeetsSDPatternsMatchBeetsLiterals` already keeps it (`:114-121`), so an
edit to the shipped table takes two edits and cannot happen quietly. The
translation map, keyed by beets literal and giving the Go form and the
template, is written out in the same file. Production carries the
translated row only, with a doc comment that quotes the beets literal and
states the brief's two facts: the trailing `\b` always holds, and the
leading one is automatic after a bracket. No production symbol exists only
for a test to read.

### R-wb3: the trigger group becomes non-capturing

The beets literal's group 1 is the trigger, `(featuring|feat|ft)`. The
rewrite, `[\[\(](?:featuring|feat|ft)[\. :].+|(^|[^\pL\pN_])(?:featuring|feat|ft)[\. :].+`,
makes the trigger non-capturing so `${1}` names the one boundary character
the second branch consumed. Groups are numbered by the order of their
opening parentheses. A capture on branch 1's trigger would become group 1,
so `${1}` would re-insert the trigger in every bracket match and drop the
kept character in every other match; the mechanics lens measured 6,351
divergences from beets over a 60,000-string draw. A capture on branch 2's
trigger opens after the boundary group and changes no output (0
divergences), so branch 2's trigger stays non-capturing only so the pattern
has exactly one group. A reviewer must not restore branch 1's capture.
Branch order matters too: under leftmost-first matching, the bracket branch
must come first, or a bracket before a trigger survives (827 divergences
with the branches swapped). The kept character is load-bearing,
since unidecode maps some non-word symbols to ASCII letters (`©` to `(c)`,
`™` to `(tm)`, `€` to `EUR`).

### R-wb4: the golden's shape

The featuring golden is one JSON file under
`internal/engine/matcher/testdata/`, with a header (beets version, CPython
version, `unicodedata.unidata_version`, the beets literal as read from
`SD_PATTERNS` through `from beets.autotag.distance import SD_PATTERNS`, the
seed, and the draw count) and two arrays. `curated` holds
`[label, input, beets_output]` rows, where a unique label names the row's
preceding-character class or shape (`ä`, `©`, `newline`, `start`, and so on).
`drawn` holds `[input, beets_output]` pairs. JSON, not TSV, because a curated
row carries a newline. It is written with `ensure_ascii=False` and one row
per line. The mechanics lens measured the default ASCII-escaped encoding at
348,821 bytes for a comparable draw, over the cap. The verification read
measured 242,193 bytes for this draw's rows one per line, under the 300 KB
cap. The generator reads the literal from beets' own
table and refuses to write when it differs from the literal the script
states, so a beets bump cannot silently regenerate against a different
pattern.

The draw is written out here, because the scratch script that first ran it
(`/tmp/claude-1000/ubound/gen.py`) lives on tmpfs. `random.seed(907)`; each
string joins `random.randint(1, 8)` tokens, each `random.choice` of these 23,
in this order: `a`, `ä`, `Ω`, `ǅ`, `²`, `Ⅻ`, U+0301, `_`, `2`, U+0020, `(`,
`[`, `:`, `.`, `ft`, `feat`, `featuring`, `e`, `r`, `-`, U+00A0, `ü`, `ß`.
The golden draws 10,000 strings (the scratch script drew 50,000) and keeps
each distinct string once, in first-draw order, through `dict.fromkeys`,
never a `set`: Python randomizes string hashes per run, so a set's order
changes between runs (mechanics lens probe). No token may be a code point
first assigned after Unicode 15.1, because the rewrite differs from beets on
every such character by construction (the mechanics lens's U+088F rows
diverged 452 of 452); the generator's docstring states this constraint. The
file stays under 300 KB. Output is deterministic: two runs under the pin
write identical bytes.

Two floors guard the draw's power. At least 150 drawn rows fire (beets'
output differs from the input), and at least 40 rows diverge between beets and the beets
literal compiled by Go. Task 1 (Python only) stops below the first floor.
Task 2's golden test fails below either floor and pins the exact measured
counts. The verification read reproduced this draw in the pinned venv:
7,955 distinct strings, 251 fired, and about 92 rows where the beets literal
under Python's `re.ASCII` flag (the proxy for Go's ASCII `\b`) diverges from
beets.

The curated rows are the brief's table inputs (`Geschäft:`, `Rüft:`, `Ωft`,
`Draft:`, `Ärft.`, `2ft`, `Café ft.`, `Ember Wharf ft.`), each at both
cases, since the row runs on lowercased text but is case-sensitive; one
string per preceding-character class the brief names (`ä`, `ß`, `Ω`, the
titlecase `ǅ`, `²`, `Ⅻ`, the combining U+0301, `_`, a no-break space, `(`,
`[`, `©`, `™`, `€`); the start of the string; two triggers in one string; and
a newline between two credits. The draw holds no `©`, `™`, `€`, newline, or
uppercase letter, so those classes reach the golden only through the
curated rows, and Task 2's test writes out the label list and fails on any
label the file lacks.

On Ruling 2's yes, a second file records Python's `\w` over every code point
outside the surrogates as sorted, non-overlapping inclusive ranges (749 at
CPython 3.13.15), with the same header.

### R-wb5: what the corpus case pins, stated exactly

`GradeCases` grades the new case end to end in `ModeCompat` (bit identity on
the record, the verdict, and the request set), which runs
`BeetsSDPatterns`. `ModeDefault` re-scores under
`SDPatternsNoBracketDiscount` and reports a moved verdict as a
`VerdictMoves` entry that fails nothing (`case.go:672-680`). The case
therefore pins the featuring row only as `BeetsSDPatterns` carries it. The
three unit rows and the translation-map test run both tables, and they pin
`SDPatternsNoBracketDiscount`'s row. The plan claims no more than that. It does state one
expectation Task 3 checks: the case's titles carry no bracket, so both
tables score them alike once both carry the translated row, and the case
should be absent from `VerdictMoves`. Present means a default-mode figure
moved, and the task stops and reports which one; the harness itself still
fails nothing.

### R-wb6: the corpus case's predicted record

Following `adv-ft-draft-colon`'s arithmetic (its `string_dist` 0.611111
records as `album: 0.101851852`, one sixth), beets' 0.523810 predicts an
`album` penalty and distance near 0.0873, medium, skip. A medium verdict
takes no strong-match early return, so freeze reports the text-search
fallback as a miss, answered by the corpus's default empty search result, as
`adv-ft-draft-colon/responses/002.json` is. These are predictions; freeze
records the real values, and the task reports any difference.

---

## Review focus

1. **The kept character.** Expected: `©ft. x` keeps `©`, and a `StringDist`
   row whose figure depends on that `©` reads beets' value. Pinned in Task 2.
2. **Leftmost choice across the two branches.** Expected: a bracket before a
   trigger is removed with it, as beets removes it. Pinned by the golden's
   curated `(` and `[` rows and the draw's bracket rows, and proved by Task
   2's branch-swap mutation.
3. **A test that passes on an empty or weak golden.** Expected: never; every
   golden test counts its rows, its fired rows, and the rows where the ASCII
   literal diverges, holds each above the R-wb4 floors, and checks every
   curated label is present. Pinned in Task 2.
4. **`SDPatternsNoBracketDiscount` drifting from `BeetsSDPatterns` on the
   shared row.**
   Expected: never; the unit rows and the translation-map test run both.
   Pinned in Task 2.
5. **A corpus record written by hand.** Expected: never; freeze writes it.
   Checked in Task 3.

---

## Task 0: The staleness pre-task

Model: `sonnet`, fresh context, before any implementer. **Its output is
blockers with evidence, not fixes.** It writes no code and edits no file.
Every check runs against the real HEAD, and the task reports that HEAD. A
pointer whose content survives at a new offset is a note naming the new
site; a pointer whose content is gone is a blocker.

**Blockers (a "no" stops the launch):**

1. Is lane 1 closed in the lane registry, and is the main tree clean of
   uncommitted changes the conductor did not author? Report `git status` and
   the registry rows.
2. Do `git log 1306ae3..HEAD -- internal/engine/matcher internal/harness
   corpus` and a diff of the same paths show no change to a surface this
   plan names? Report each commit by function.
3. Does `tools/venv/.venv` report beets 2.14.1 and CPython 3.13, and does
   `distance.py:39` still read the literal quoted above?
4. Does `ls corpus/cases | wc -l` read 215, with no case named
   `strdist-feat-unicode-boundary`?
5. Does `StringDist` still read the three brief figures before the fix
   (0.052381, 0.064706, 0.068750 for the `Geschäft`, `Rüft`, and `Ωft`
   pairs)? Run through `go test -overlay` or a scratch test outside the
   tree, never a tree edit.

**Notes:**

6. Report `go version` and `unicode.Version`; a version other than 1.27.1
   and 17.0.0 is a note that Task 2 re-measures Ruling 2's count against.
7. Report the floor constants at `case.go:40-45` and `request.go:28-30`.
8. Report whether `/tmp/claude-1000/ubound/` still exists; its absence is
   not a blocker, since R-wb4 writes out the seed, the length range, and the
   23 tokens.

**2.0 hook:** neutral. **Deliverable count: 1** (the report).

---

## Task 1: The golden generator and its fixtures

Model: `sonnet`. Depends on Task 0. **Neutral on 2.0 hooks.** This task lands
Python and data only. No Go test reads the golden until Task 2, because a
test against the untranslated row would commit red.

**Files:**

- Create: `internal/engine/matcher/testdata/gen_featuring_golden.py`
- Create: `internal/engine/matcher/testdata/featuring-golden.json`
- Create, on Ruling 2's yes: `internal/engine/matcher/testdata/python-word-ranges.json`

**Outcomes:**

- **The generator** writes both files from the pinned venv, per R-wb4. Its
  module docstring states the run command, that the gate and `make derive`
  never run it, that it is the fixture's provenance record, the literal
  check, and the Unicode 15.1 constraint on the draw's tokens. It imports
  the table as R-wb4 names, removes duplicates through `dict.fromkeys`, and
  writes JSON with `ensure_ascii=False`, one row per line. It follows PEP 257
  under the repo's `ruff.toml` (`E`, `F`, `D` with the google convention).
- **The featuring golden** carries the labeled curated rows and the seeded
  draw.
- **The class-range golden** (Ruling 2's yes) carries Python's `\w` ranges.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| The venv's beets version | not 2.14.1 | the generator exits nonzero, naming both versions, and writes nothing |
| The literal read from `SD_PATTERNS` | differs from the stated literal | exits nonzero, printing both, and writes nothing |
| An output path | its directory is absent | exits nonzero naming it |
| Two runs | same pin | byte-identical files |
| The drawn rows | fewer than 150 fire | the task stops and reports the count; it does not reseed or widen the token set on its own |

**Acceptance criteria:**

- `ruff check` and `ruff format --check`, run directly on the generator
  (Global constraints says why the gate cannot see it yet), pass; the report
  pastes both outputs.
- Running the generator twice under the pin writes byte-identical files; the
  report pastes both `sha256sum` lines.
- The report states the curated row count and label list, the distinct drawn
  row count, the fired drawn count (at least 150; the verification read
  measured 251 for this draw), and the file sizes (featuring golden
  under 300 KB).
- On Ruling 2's yes, the range count reads 749 at CPython 3.13.15.
- gitleaks passes over the new files. If it flags a drawn string, the task
  stops and reports; it adds no allowlist entry on its own.
- The gate is green.

**Named mutations this task reports:**

1. Change the stated literal in the generator by one character, then run it:
   it refuses and writes nothing.
2. Drop the seed: a second run writes different bytes. If the generator
   cannot run unseeded, the report marks this mutation reviewer-read.
3. Change the expected beets version to `2.14.0`, then run it: it refuses,
   naming both versions, and writes nothing.

**Deliverable count: 3** (the generator, the featuring golden, the
class-range golden; 2 on Ruling 2's no).

---

## Task 2: The translated row and its proofs

Model: `sonnet`; the conductor upshifts to `opus` only if the implementer
reports a golden row it cannot reconcile. Depends on Task 1. **Neutral on
2.0 hooks.** The 2.0 spec's "Search" section holds an edit-distance rerank
for 2.1 on `matcher.StringDist`; it would inherit this fix and needs nothing
from it now.

**Files:**

- Modify: `internal/engine/matcher/tunables.go`, `strdist.go`,
  `strdist_test.go`, `tunables_test.go`
- Modify, one comment only: `internal/harness/oracle/derive_defect_free.py`
  (`:34-36`)

**Outcomes:**

- **The field and the loop** (R-wb1). `SDPattern` gains `Replacement`; the
  loop passes each row's template to `ReplaceAllString`. `SDReplace`,
  `applySDReplace`, `compileSDPatterns`, and `sdRegexp` are unchanged. On
  Ruling 3's no (fix now), `StringDist` also replaces U+0130 with `i` plus
  U+0307 on both sides before `strings.ToLower` (`strdist.go:36-37`), with a
  comment naming it the second departure and its reason; `sdStrip`'s
  lowercase at `:103` is unchanged, since the strip removes the dot.
- **The translated row and the comments that describe it** (R-wb2, R-wb3).
  Row 3 of `BeetsSDPatterns` and of `SDPatternsNoBracketDiscount` carries
  the rewrite and `${1}`, from one unexported pair of constants (or one
  unexported row value) so the two tables cannot drift. The doc comment on
  `BeetsSDPatterns` replaces the "must stay byte-identical" sentence
  (`:365-366`) and the gap note (`:371-374`) with the one named departure:
  the beets literal, the two facts, and the kept character's reason. Three
  more comments that describe the untranslated row move with it:
  - `TestStringDistFeaturingNeedsAWordBoundary`'s doc comment
    (`strdist_test.go:135-142`). It names "the leading \b", which the
    shipped row no longer carries.
  - The case-mapping sweep's U+0130 sentence (`strdist_test.go:222-225`).
    It holds for the strip but not for the pattern loop, and it then names
    Ruling 3's outcome.
  - The table-parity sentence in `derive_defect_free.py:34-36`. It then says
    the Python table carries beets' literal, the Go table carries the
    featuring row translated, and why.
- **The parity tests.**
  - `TestBeetsSDPatternsMatchBeetsLiterals` keeps its written-out literal
    table and asserts every row of `BeetsSDPatterns` through the translation
    map. A row whose literal is not in the map equals it with an empty
    `Replacement`. A row whose literal is in the map carries the map's Go
    form and template. The test's closing check that `DefaultTunables`
    selects the shipped table (`strdist_test.go:130-132`) compares
    `StringDistPatterns[2].Pattern`, the featuring row, with the literal
    and would red after the translation. It routes through the same map and
    is not weakened or dropped. A second assertion does the same for
    `SDPatternsNoBracketDiscount` against the literal table less its two
    bracket rows. Every map entry is used by at least one row.
  - `TestBeetsSDPatterns_SixEntriesInBeetsOrder` (`tunables_test.go:185-197`)
    duplicates the literals test and is removed.
    `TestSDPatternsNoBracketDiscount_OnlyBracketPatternsRemoved` (`:202-213`)
    routes through the same map, so no test outside the map writes the Go
    form.
  - A golden test reads `featuring-golden.json` and runs the shipped row's
    compiled pattern and template over every curated and drawn input,
    expecting beets' output. It writes out the curated label list and the
    R-wb4 floors.
  - On Ruling 2's yes, a class test reads the range file. It classifies
    every code point `r` outside the surrogates through the shipped row
    itself, never a copy of its class. For each `r`, it substitutes on
    `"a" + string(r) + "ft. x"`. An unchanged output marks `r` a Go word
    character, and the output `"a" + string(r)` marks it a non-word
    character. The bracket branch removes `(` and `[`, so their expected
    output is `"a"`. The test compares the result with the range file. The
    report states the sweep's wall time (about 1.1M substitutions).
- **The unit rows.** `TestStringDistFeaturingNeedsAWordBoundary` gains the
  `Geschäft`, `Rüft`, and `Ωft` pairs, each run under both tables. Measured
  in the pinned venv on 2026-09-25 with `repr`: `0.5238095238095238`,
  `0.6470588235294118`, and `0.6875`. `TestStringDist` gains one row that
  proves the loop keeps the consumed character: `"Ember Wharf ©ft.
  Northern"` against `"Ember Wharf ©"`, measured at `0.047619047619047616`.
  This row passes before the change too, since Go's ASCII `\b` holds between
  `©` and `f`; it exists to fail the mutation that drops the kept group. One
  more row runs `"İft: Ember Wharf"` against `"İft: Northern Lines"` under
  both tables; beets reads `0.06875` (risk lens, pinned venv). On Ruling 3's
  yes it is a declared known divergence, in the case-mapping sweep's style:
  it writes out beets' figure and Go's figure as the task measures it, fails
  when Go's figure moves, and says in its message that a Go figure equal to
  beets' means the divergence is fixed and the declaration comes out. On
  Ruling 3's no it expects beets' `0.06875` exactly.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| The golden file | absent | the test fails naming the path; it never skips |
| The golden file | malformed JSON | fails naming the decode error |
| The golden header | beets version, literal, or Unicode version differs from the test's written-out values | fails naming the field and both values |
| The golden rows | empty, or fewer than the counts Task 1 reported | fails naming both counts |
| The curated rows | a written-out label is absent from the file | fails naming the label |
| The drawn rows | fired rows below 150 | fails naming the count: the golden is too weak to prove the row |
| The golden rows | many diverge | reports each diverging row (label or input, want, got) up to 20, then the total |
| The ASCII-literal control | fewer than 40 rows diverge, or the count differs from the pinned value | fails naming both counts: below the floor, the golden cannot tell the translation from the defect |
| The range file (Ruling 2) | absent, malformed, unsorted, or overlapping | fails naming the defect |
| The class sweep (Ruling 2) | an output is neither unchanged nor the expected non-word form | fails naming the code point and the output |
| The class sweep (Ruling 2) | a Python-only word code point exists | fails naming the first few |
| The class sweep (Ruling 2) | the Go-only count differs from 9,039 | fails naming the count and the toolchain's `unicode.Version`, and states the procedure: confirm the Python-only count is still zero, check whether any newly assigned code point can sit before a trigger in a real title, then update the count and record it in HISTORY |
| The translation map | an entry no row uses | fails naming it |
| The U+0130 row (Ruling 3's yes) | Go's figure differs from the pinned one | fails naming both, and says a figure equal to beets' means the divergence is fixed |

The ASCII-literal control is the non-vacuity check for the golden test: it
compiles the beets literal with Go's `regexp` over the same rows, counts the
rows where Go disagrees with beets, and pins that count (reported by the
task), at or above R-wb4's floor of 40.

**Acceptance criteria:**

- The golden test passes with zero diverging rows; the report states the
  curated and drawn row counts, the fired count, and the pinned
  ASCII-literal divergence count.
- The three unit rows read beets' figures under both tables, compared
  exactly, per the file's standing rule (`strdist_test.go:9-15`). The
  U+0130 row reads as Ruling 3's answer states, and the report pastes Go's
  figure.
- No existing `StringDist`, component, or corpus expected figure changes;
  `TestComponentTableReplaysAllRows` and the two whole-corpus sweeps pass
  unchanged. The three pattern-table tests change only as the parity-test
  outcome states. (The mechanics lens overlaid the rewrite and found only
  those three tests red.)
- `git grep -nF -e 'must stay byte-identical' -e 'No committed case carries
  such a string' -- internal/engine/matcher/tunables.go` finds nothing, and
  the diff-reviewer reads the new doc comment against R-wb2's three
  elements.
- The gate is green.

**Named mutations this task reports:**

1. Restore row 3 to the beets literal with an empty template: the golden test
   fails on the pinned count of rows, and the three unit rows read 0.052381,
   0.064706, and 0.068750.
2. Keep the rewrite with an empty template: the golden fails on every fired
   row with a character before the trigger.
3. Make the loop pass `""` instead of `Replacement`: the `©` row in
   `TestStringDist` fails.
4. Make branch 1's trigger capturing, `[\[\(](featuring|feat|ft)`: the
   golden fails (the trigger is re-inserted in bracket rows, and the kept
   character is dropped elsewhere). A capture on branch 2's trigger is not
   this mutation: it changes no output (R-wb3).
5. Translate only `BeetsSDPatterns`, leaving `SDPatternsNoBracketDiscount`
   on the literal: the unit rows fail under the second table.
6. Truncate the golden to its header: the row-count check fails.
7. Swap the two branches: the golden fails, at least on the curated `(` and
   `[` rows.
8. On Ruling 2's yes, change the shipped class to `[^\pL\p{Nd}_]`: the class
   sweep's Python-only count goes nonzero and names `²`.
9. On Ruling 3's no, drop the U+0130 replacement: the U+0130 row fails.

**Deliverable count: 4** (the field and loop, with Ruling 3's fix on its no;
the translated row with the comments that describe it; the parity tests with
the golden and class tests; the unit rows, the U+0130 row included).

---

## Task 3: The `strdist-feat-unicode-boundary` corpus case

Model: `sonnet`. Depends on Task 2: before the fix, `ModeCompat` grades this
case as a divergence and the gate reds. **Neutral on 2.0 hooks.**

**Files:**

- Create: `corpus/cases/strdist-feat-unicode-boundary/case.json`,
  `claims.json`, `responses/001.json`, `responses/002.json`
- Modify: `corpus/analysis/ties.json` (regenerated), `corpus/README.md`,
  `internal/harness/conform/case.go`, `internal/harness/conform/request.go`
- Modify, comments that count cases only: `internal/harness/conform/case_test.go`,
  `internal/harness/conform/defaults.go`, `internal/engine/tag/tag_test.go`,
  `internal/engine/matcher/verdict_test.go`
- Modify only if the gate names a row owed: `corpus/analysis/proximity.json`

**Outcomes:**

- **The case**, authored through the README's synthetic path (`:2747-2778`):
  a synth spec for a clean 3-track release with an invented `mb_albumid`
  (unique across `corpus/cases`), profile `estate`, and `eval_date`
  `2026-09-02`, matching family 4's siblings. The candidate album is
  `"Geschäft: Ember Wharf"`, and three `field` overrides, one per item index
  0 to 2, set each item's `album` to `"Geschäft: Northern Lines"` (README,
  "Items are derived"). Freeze under the pinned venv, answer the search miss
  with the default empty search result, re-freeze to zero misses, and move
  the staging directory into `corpus/cases/`. The spec itself stays in
  `corpus/staging/`, which is gitignored, as for every synthetic case.
- **The claims and ledger rows.** `claims.json` names `best_album_id`, the
  recommendation freeze recorded, and `penalty_keys_nonzero: ["album"]`
  with the other four keys zero, as `adv-ft-draft-colon/claims.json` does.
  `ties.json` gains the case's row through `go test -tags oracle
  ./internal/harness/oracle -run TestTieLedgerMatchesTheCorpus
  -update-ledgers`. The README's command (`:2254`) runs `-run Ledger`, which
  also matches the album-info and file-tags ledger tests; under
  `-update-ledgers` those rewrite `album_info.json` and `file_tags.json`
  (`album_info_derive_test.go:91`, `file_tags_derive_test.go:87`, mechanics
  lens), so this task names the one test. The tie derivation re-measures
  every case (`ties_derive_test.go:54-80`), so its wall time scales with all
  216.
- **The README.** Family 4's count moves from 5 to 6, and its table gains
  the row. Family 4 is defined today as pairs drawn from beets'
  `test_distance.py` that read "equivalent, or ... merely closer," and the
  new case is neither: it is a no-fire case whose strings must stay apart.
  So the family's defining sentences (`:232-257`, "Four of the five pairs"
  included) are rewritten to admit one port-hazard probe of the Go port's
  word boundary. The structural inventory
  (`:1681`) gains the row and its count; the corpus totals (`:1572`,
  `:1607`, `:1612`, `:1917`) move from 215 to 216, the `:1917` sum moves its
  structural term from 61 to 62, and `:1572`'s attribution of all 61
  structural cases to pass 2b is corrected.
- **The floors and the counts.** Each floor at `case.go:40-45` and
  `request.go:28-30` rises to the value the census measures with the case
  in, and the doc comment's measured date moves to the task's date. Every
  Go comment that counts committed or candidate-bearing cases moves with
  it: `git grep -nw -e 215 -e 214 -- '*.go'` finds today `case.go:29`,
  `case_test.go:286`, `defaults.go:17`, `tag_test.go:244`, and
  `verdict_test.go:86`, besides the constants; the task moves each comment
  that counts cases and reports any it leaves, with the reason.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Freeze | misses remain | freeze refuses to write; hand-author the empty search response and re-freeze |
| The frozen record | verdict or distance differs from R-wb6's prediction | report both; the record stands, since freeze is authoritative |
| `ModeCompat` grading | the search request's `path_query` does not match Go's | stop and report; a request-construction divergence is a finding outside this pass |
| `ModeCompat` grading | a penalty or the verdict diverges | stop and report the adjudication row the gate prints; never write it |
| The tie regeneration | changes any `ties.json` row other than the new case's | stop and report the diff; never hand-trim |
| `album_info.json`, `file_tags.json` | changed by any command this task ran | stop and report |
| `ModeDefault` grading | the case appears in `VerdictMoves` | stop and report which figure moved (R-wb5) |
| The proximity ledger | the gate names the case owed | add the row with its reason |
| The proximity ledger | nothing owed | leave the file untouched |

**Acceptance criteria:**

- The record's `album` penalty is the only nonzero penalty, and the report
  pastes its value, the distance, and the verdict.
- `TestWholeCorpusGradesCleanlyInCompatMode` and
  `TestWholeCorpusGradesVerdictsInCompatMode` pass with the case graded; the
  report pastes the case count, 216.
- The report pastes the oracle ledger run's wall time and states that
  `ties.json`'s diff is one row and that `album_info.json` and
  `file_tags.json` are unchanged.
- The report pastes the `ModeDefault` log line; the case is absent from
  `VerdictMoves`.
- The report pastes each floor constant's old and new value beside the
  census output.
- `corpus/manifest.json` is unchanged.
- The gate is green.

**Named mutations this task reports:**

1. With the case committed, restore row 3 to the beets literal (Task 2's
   mutation 1) and run the compat sweep: the case fails. Under the literal,
   Go scores the case strong (about 0.0087), takes the strong-match early
   return, and skips the recorded search, so the first failure may name the
   request set or the verdict before the `album` penalty. The report names
   which rule fired.
2. Set one floor one above the measured value: the census test fails.

**Deliverable count: 4** (the case, its claims and ledger rows, the README
entries, the floors with the case-count comments).

---

## Task 4: The pass close drafts

Model: `sonnet`. Depends on Tasks 0 to 3 and the simplifier round.

**Files (drafts only; the conductor folds and commits):**

- Create: `docs/superpowers/research/handoffs/draft-HISTORY-matcher-word-boundary.md`
- Create: `docs/superpowers/research/handoffs/draft-STATUS-matcher-word-boundary.md`

**Outcomes:**

- **The HISTORY draft** carries what landed per task with commit ids, the
  gate evidence, R-wb1 to R-wb6 with where each landed, the rulings' outcomes,
  both budgets, and "what a later pass would be wrong to rediscover." That
  section carries at minimum R-wb3's non-capturing trigger, the kept
  character's reason, the branch order under leftmost-first matching, the
  Unicode-version count Ruling 2 settled (or its comment-only record on
  no), the U+0130 case-mapping divergence as Ruling 3 settled it (declared
  and pinned, or fixed as a second departure), the sweep's finding that
  U+0130 is the only code point whose Python lowercase changes whether the
  last character is a word character (a later toolchain bump re-runs that
  sweep), the `-run Ledger` trap
  (it rewrites two more ledgers), and the `CompilePatternOption` class left
  documented only. It also notes that `shadow.configHash`
  (`internal/engine/shadow/match.go:55-75`) hashes the Tunables JSON, so
  the new field and pattern text move `matcher_config_hash` at the next
  deploy, as intended; a shadow report showing two hashes is expected.
- **The STATUS draft** removes "Next action" item 3 and records nothing
  carried from this pass unless the pass found something.

**Acceptance criteria:** every count in the drafts comes from a command the
report pastes; every commit id resolves; the STATUS draft is at most 60
lines. The gate is green as a formality.

**Deliverable count: 2.**

---

## The exit gate

`cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`
green on master after the last commit, plus:

1. The brief's three figures read beets' values under both tables (Task 2).
2. The golden test passes with a pinned ASCII-literal divergence count at or
   above 40 (Task 2).
3. The corpus case grades cleanly in `ModeCompat` (Task 3).
4. No test outside the translation map writes the featuring row's Go form:
   `git grep -nF '(^|[^\pL\pN_])(?:featuring|feat|ft)' -- '*_test.go'`
   returns exactly the translation-map line. No shipped row other than the
   featuring row carries a `Replacement`: the translation-map test asserts
   it (Task 2).
5. No new exported symbol other than `SDPattern.Replacement`, which
   `StringDist` reads. `go doc -short ./internal/engine/matcher` ("One-line
   representation for each symbol", `go help doc`) prints the same lines at
   the base commit and at HEAD. That listing elides struct fields
   (`type SDPattern struct{ ... }`), so `go doc matcher.SDPattern` at the
   two commits differs only by the `Replacement` field and the doc text
   that describes it.

---

## Contested surfaces

| Surface | Owner | This pass's position | Contention |
| --- | --- | --- | --- |
| `internal/engine/matcher` | rung 3 | Tasks 1 and 2; Task 3, one case-count comment | none; no open lane touches it |
| `internal/harness/conform` floors | rung 3/4 | Task 3, two constant blocks and case-count comments | none |
| `internal/engine/tag/tag_test.go`, `internal/harness/oracle/derive_defect_free.py` | rung 3, rung 2 | one comment each (Tasks 3 and 2) | none; no open lane touches either |
| `corpus/`, `corpus/analysis/ties.json` | rung 2 | Task 3 | none; main tree only |
| `scripts/check.sh`, `web/` | rung 5 | untouched; the gate's web leg runs | a fresh `npm ci --prefix web` may be owed after 5c-ii's close |
| `docs/` ledgers | conductor | Task 4 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: confirm 5c-ii closed and lane 1 is free; record the
   planning spend; write the lane registry row; arm the runaway guard; check
   AC power; record the three rulings' outcomes.
2. After Task 2: STATUS; re-derive the ceiling from Tasks 0 to 2's reports.
3. After Task 3: one `code-simplifier` round over Tasks 2 and 3's Go,
   pasting `.claude/instructions/simplifier-brief.md` verbatim, with one
   `diff-reviewer` over its diff. The Python generator is outside its scope.
4. After Task 4: fold the drafts, then one `go-architecture-reader` for
   `internal/engine/matcher` and one for `internal/harness/conform`, with no
   plan in context.
5. At the close: add `internal/engine/matcher/testdata/gen_featuring_golden.py`
   to CLAUDE.md's two lists of tracked Python ("Gate" and "Stack rules:
   Python"). The conductor owns CLAUDE.md.
6. **Do not deploy.**

---

## Execution model

Per-task Agent dispatch, sequential, one segment: Task 0, then Tasks 1 to 3,
then the simplifier round, then Task 4. Tasks 1 to 4 each run
`dubplate-implementer` (`sonnet`, `high`), then `diff-reviewer`
(`claude-opus-5-5`, `high`), then the gate. Task 0 is a read-only `sonnet`
dispatch whose report the conductor reads directly. One re-dispatch on `fix`; a
second `fix` is the conductor's decision. Every reviewer dispatch asks for
correctness gaps ranked by consequence, not polish.

Dependencies: Task 0 → Task 1 (the pin checks) → Task 2 (the golden it
reads) → Task 3 (the fix the case needs to grade green) → Task 4. No task is
independent: Tasks 1 and 2 share `testdata/`, and Task 3's gate needs Task 2.

**Budget:** the ceiling and flag in "Token ceiling" (Ruling 1), planning
recorded separately, STATUS after Task 2 and at the close, spend scored
against the ceiling and against zero attended interactions after approval.

---

## Prior art consulted

Scored on CLAUDE.md's four criteria where a dependency is in question. The
2.0 spec's prior-art list (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`,
"Prior art") and forward-hooks list were both read; no hook is landed, and
every task is neutral.

| Entry | Verdict for this pass |
| --- | --- |
| Go `regexp` (RE2), stdlib | **Adopt.** One rewritten expression with an `Expand` template; linear time, no new dependency. |
| `dlclark/regexp2` v1.12.0 | **Decline**, on the brief's evidence: MIT, 1,021 importers, one maintainer, and its own README recommends the stdlib. Its .NET `\w` diverged from beets on 198 of 50,000 strings, so it does not deliver beets' behavior, and it brings a backtracking engine. |
| Python `re` documentation | **Adopt as the behavior specification**, quoted through the brief and confirmed in the venv. |
| beets 2.14.1 `autotag/distance.py` | **Adopt as the oracle**, read from the pinned venv; the generator imports its `SD_PATTERNS`. |
| The case-mapping sweep (`strdist_test.go:240`) and `gen_golden_flac.py` | **Adopt as in-repo precedent** for a declared Unicode-version skew and for a generator that is a fixture's provenance record. |
| `xrash/smetrics` | Already declined in the 2.0 spec, since `strdist.go` and the pinned unidecode table cover it. No change. |
| Navidrome, gonic, taglib, `mewkiz/flac`, Picard, chromaprint, goose, `modernc.org/sqlite` | No surface in this pass. |

---

## Not verified in this revision

- Python's `re` documentation text is quoted from the research brief, not
  re-fetched. The behaviors this plan relies on were confirmed in the pinned
  venv.
- The brief's 50,000-string differential run was not re-run by the plan's
  author. The mechanics lens ran its own 60,000-string draw (seed 907, 36
  tokens, among them `©`, `™`, `€`, `\n`, `)`, `]`, `é`, U+203F, and
  U+088F, a code point first assigned in Unicode 16.0), and found zero
  differences for the rewrite outside the 452 U+088F rows.
- The corpus case's figures (distance near 0.0873, medium, skip) are an
  arithmetic prediction from `adv-ft-draft-colon`'s ratio; freeze records the
  real ones.
- Whether Go's search query for `geschäft: northern lines` canonicalizes to
  the same `path_query` as beets'. Committed non-ASCII search requests (the
  `real-nonlatin-*` cases) grade cleanly today, so the risk is low; Task 3
  names the stop.
- Whether gitleaks flags any drawn string in the golden.
- Whether CPython's seeded `random` draws the same sequence on a later 3.13
  patch release. The committed golden is the record, so this touches only
  regeneration, and Task 1's two-run check covers the pinned interpreter.
