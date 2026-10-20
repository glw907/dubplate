# Matcher word-boundary plan review: mechanics and feasibility

Lens: mechanics and feasibility only. Subject:
`docs/plans/2026-09-25-matcher-word-boundary.md`, revision 1
(uncommitted). Tree read at `1306ae3`. Probes live in this session's scratch
directory (`/tmp/claude-1000/-var-home-glw907-Projects-dubplate/bc83700f-945c-4755-8eac-4c458ffb90b6/scratchpad/wb/`),
run with Go 1.27.1 (`unicode.Version` 17.0.0) and the pinned venv read-only
(CPython 3.13.15, `unidata_version` 15.1.0, beets 2.14.1). No tree file was
edited; the Go probes ran through `go test -overlay`.

**Verdict: sound, with no blockers.** The core mechanism works as the plan
states. Two majors would each likely cost a fix round: a named mutation that
cannot fire, and a ruff acceptance check that the gate's own ordering makes
vacuous. Seven minors follow.

Counts: 0 blocker, 2 major, 7 minor.

## What was verified and holds

- **The `go doc` quotes are exact.** Checked: `regexp.Compile` (leftmost-first),
  `ReplaceAllString`, `Expand` ("A reference to an out of range or unmatched
  index ... is replaced with an empty slice"), and `regexp/syntax`'s `\b` and
  `\w` lines.
- **The `${1}` template with a non-capturing trigger behaves as stated.** The
  bracket branch leaves group 1 unmatched, so it expands to empty. The `^`
  alternative expands to empty. The class alternative puts back the one
  character it consumed. `"$ft. x"` gives `"$"`, so a `$` in the input is copied
  from the source and never read as template syntax.
- **The rewrite has zero differences from beets over my own 60,000-string
  draw.** The draw uses seed 907 and 36 tokens: the brief's set plus `©`, `™`,
  `€`, `\n`, `)`, `]`, `é`, and U+203F. Rows that contain U+088F were counted
  separately. It is a Go-only word character (Unicode 16), and the rewrite
  diverges on all 452 of those rows by construction (see m5). By contrast, the
  beets literal compiled by Go diverges on 3,188 rows. So the ASCII-literal
  control in Task 2 will be nonzero.
- **Ruling 2's figures reproduce exactly.** Python's `\w` compresses to 749
  ranges over 138,558 code points, and no range spans the surrogate gap.
  `re.match(r"\w")` agrees with `isalnum() or "_"` on every code point. Against
  Go's `[\pL\pN_]`, 0 code points are word characters in Python only and 9,039
  are word characters in Go only. The Go-only set breaks down as Lo 8,827,
  Nd 90, Lu+Ll 106, Lm 13, and Nl 3. The first three are U+088F, U+0C5C, and
  U+0CDC. A Go test that sweeps every code point through the class takes
  0.26 s.
- **The beets source matches the plan's pointers.** `SD_PATTERNS = [` is at
  `distance.py:36`, the featuring row is at `:39`, and
  `re.sub(pat, "", str1)` is at `:100`. The venv reports beets 2.14.1.
- **Every unit-row figure matches beets bit for bit under both tables.** I
  overlaid the rewrite on `tunables.go` and `strdist.go`, with a
  `Replacement` field and one shared row value. `Geschäft` reads
  0.5238095238095238, `Rüft` 0.6470588235294118, `Ωft` 0.6875, and the `©`
  row 0.047619047619047616. Each equals beets' `repr` from the venv.
- **Mutation 3 fires.** When the loop passes `""`, the `©` row reads
  0.12943722943722943 under both tables.
- **"No existing expected value changes" holds.** Under the same overlay, I
  ran `internal/engine/matcher`, `internal/harness/...` (with
  `conform`'s two whole-corpus sweeps and the default-mode sweep, 178 s),
  `internal/engine/shadow`, `internal/engine/tag`, and
  `internal/engine/musicbrainz`. Only three tests fail, and each is one the
  plan rewrites or removes: `TestBeetsSDPatternsMatchBeetsLiterals`,
  `TestBeetsSDPatterns_SixEntriesInBeetsOrder`, and
  `TestSDPatternsNoBracketDiscount_OnlyBracketPatternsRemoved`. The four
  `defect_free` figures still reproduce.
- **R-wb5's table mapping is correct.** `compatFor` gives `FullCompat` under
  `ModeCompat`, and `Compat.Apply` then assigns `BeetsSDPatterns`
  (`compat.go:47-53`). `ModeDefault` runs `SDPatternsNoBracketDiscount`, and
  `gradeCaseDefault`'s default branch records a moved verdict without failing
  (`case.go:672-680`).
- **R-wb6's arithmetic agrees with the sibling case.** `adv-ft-draft-colon`
  records `album` 0.10185185185185186, which is 0.6111/6, and reads medium,
  skip. By the same arithmetic, 0.5238095/6 gives about 0.0873.
- **Task 3's mechanics check out.**
  - `ties.json` carries a row for every case, which `ties.go:118` enforces.
  - The floors are minimums (`request_test.go:85`, `case_test.go:249-263`),
    so Task 3's mutation 2 fires.
  - `corpus/manifest.json` holds pins only.
  - 215 cases exist today.
  - A committed search `path_query` already percent-encodes non-ASCII text
    (the `real-nonlatin-*` cases).

## Majors

### M1. Mutation 4 can survive, and R-wb3 gives the wrong reason

- **Severity:** major.
- **Location:** plan R-wb3 (`:317-325`) and Task 2 mutation 4 (`:582`).
- **Defect:** R-wb3 says a restored trigger capture makes `${1}` re-insert the
  trigger. That holds only for branch 1. Groups are numbered by the order of
  their opening parentheses. A capture on branch 2's trigger opens after the
  boundary group, so group 1 is still the boundary character and the output
  does not change. Mutation 4 says only "restore the capturing trigger group."
  An implementer who restores it on branch 2, which sits next to the kept
  group, gets a mutation that cannot fire. The implementer's definition says
  to report such a mutation as a finding, so a fix round follows.
- **Evidence:** a differential over the 60,000-string draw, counted as
  divergences from beets:

  | Variant | Divergences |
  | --- | --- |
  | Capture restored on branch 2 only | **0** |
  | Capture restored on branch 1 only | 6,351 |
  | Capture restored on both branches | 6,351 |
  | Plan rewrite | 0 |

  With a capture on branch 1, `"\n\n(ft:Ⅻ"` should give `"\n\n"` but gives
  `"\n\nft"`.
- **Fold:**
  - Rewrite R-wb3's reason. A capturing group in branch 1 becomes group 1, so
    `${1}` then re-inserts the trigger in bracket matches and drops the kept
    character in every other match.
  - State that branch 2's trigger stays non-capturing so that the pattern has
    exactly one group. It does not affect the output.
  - Reword mutation 4 as "make branch 1's trigger capturing: the golden
    fails."

### M2. Task 1's ruff check never runs on the new generator

- **Severity:** major.
- **Location:** plan Task 1 acceptance (`:465-466`) and Global constraints
  (`:252-256`).
- **Defect:** The gate lints only files that `git ls-files '*.py'` lists.
  `dubplate-implementer` runs the gate at step 5 and stages and commits at
  step 6. So when Task 1's gate runs, the new generator is untracked and ruff
  skips it. "ruff check and ruff format --check pass (through the gate)" is
  satisfied without checking the file. The first real lint happens at Task 2's
  gate. That turns Task 2 red on a file outside Task 2's Files list.
- **Evidence:**
  - `scripts/check.sh:381-392`: "Tracked files only, so an unstaged experiment
    does not gate", followed by `mapfile -t py_files < <(git ls-files '*.py')`.
  - `.claude/agents/dubplate-implementer.md:72-82`: the gate runs at step 5,
    and the commit comes at step 6.
- **Fold:** Task 1's acceptance runs `ruff check <generator>` and
  `ruff format --check <generator>` directly and pastes the output. Staging
  the file before the gate would also work, but the direct run is simpler.

## Minors

### m1. The obvious import of beets' pattern table fails

- **Location:** R-wb4 (`:331-336`).
- **Defect:** The generator is told to read "`beets.autotag.distance.SD_PATTERNS`".
  But `import beets.autotag.distance as d` binds the function
  `beets.autotag.distance`, not the module, so `d.SD_PATTERNS` raises
  `AttributeError`.
- **Evidence:** In the venv, `import beets.autotag.distance as d` gives
  `<class 'function'>`, and the attribute is missing. Two forms work:
  `from beets.autotag.distance import SD_PATTERNS` and
  `importlib.import_module("beets.autotag.distance").SD_PATTERNS`.
  `derive_defect_free.py:24-27` documents the same trap.
- **Fold:** Name the working import in R-wb4.

### m2. Removing duplicates with a Python set breaks the two-run check

- **Location:** R-wb4 ("keeps each distinct string once") and Task 1's
  byte-identical acceptance.
- **Defect:** Python randomizes string hashes on each run, so a `set` iterates
  in a different order every time. The two-run check would catch this, but
  only after the output is already wrong.
- **Evidence:** Two identical runs hash the joined set output to
  `fe507b13…` and `d48af4ed…`. The sorted output hashes to `1452b611…` both
  times.
- **Fold:** Specify `dict.fromkeys(draw)` (which keeps the first-draw order)
  or `sorted`.

### m3. The 300 KB cap holds only for one JSON encoding

- **Location:** R-wb4 (`:339`) and Task 1 acceptance (`:472`).
- **Defect:** The size cap depends on the encoding, and the plan names none.
  Python's default `json.dumps` escapes non-ASCII characters
  (`ensure_ascii=True`), which pushes the file over the cap.
- **Evidence:** A 10,000-string draw over the brief's tokens gives 7,898
  distinct strings, 331 of which fire. The file sizes for each encoding:

  | Encoding | Size |
  | --- | --- |
  | Compact, `ensure_ascii=False` | 236,532 B |
  | `indent=1`, `ensure_ascii=False` | 331,314 B |
  | Compact, `ensure_ascii=True` (the default) | 348,821 B |

- **Fold:** Specify `ensure_ascii=False` and one row per line (about 245 KB),
  or raise the cap.

### m4. `-run Ledger` rewrites two more ledgers

- **Location:** Task 3 (`:621-623`) and "Not verified" (`:794-795`).
- **Defect:** The pattern `Ledger` also matches
  `TestAlbumInfoLedgerMatchesTheCorpus` and
  `TestFileTagsLedgerMatchesTheCorpus`. Under `-update-ledgers`, both rewrite
  `album_info.json` and `file_tags.json`. The rewrite should change nothing,
  but neither file is in Task 3's Files list or its checks.
- **Evidence:** `grep '^func Test' internal/harness/oracle/*_test.go | grep
  Ledger` lists five tests. `album_info_derive_test.go:91` and
  `file_tags_derive_test.go:87` both write when `*updateLedgers` is set.
- **Fold:** Run `-run TestTieLedgerMatchesTheCorpus -update-ledgers`, or add an
  acceptance line that both other files are unchanged. The plan's open
  question also resolves: the tie derivation re-measures every case
  (`ties_derive_test.go:54-80`), so its wall time scales with all 216 cases.

### m5. The plan never writes out the draw's token set

- **Location:** R-wb4 ("the brief's token set") and Task 0 note 8 (`:425-426`).
- **Defect:** Task 0 note 8 says R-wb4 restates the token set, but R-wb4 does
  not. The exact 23-token list exists only in
  `/tmp/claude-1000/ubound/gen.py`, which is volatile. The brief's prose names
  most of the tokens but not every one (for example, `e`, `r`, `2`, and `-`).
  The list also carries a constraint the plan never states: no token may be a
  code point assigned after Unicode 15.1, or the rewrite differs from beets by
  construction.
- **Evidence:** My draw included U+088F, and the rewrite diverged on all 452
  rows that contain it.
- **Fold:** Write the 23 tokens into R-wb4:
  - `a`, `ä`, `Ω`, `ǅ`, `²`, `Ⅻ`, U+0301, `_`, `2`, and a space
  - `(`, `[`, `:`, `.`, `ft`, `feat`, `featuring`, `e`, and `r`
  - `-`, a second space, `ü`, and `ß`

  Add the Unicode-15.1 constraint to the generator's docstring.

### m6. Two comments go stale when the row is translated

- **Location:** "What this pass is not" (`:30-39`) and Task 2's Files list.
- **Defect:** Two comments describe the untranslated row.
  - `internal/harness/oracle/derive_defect_free.py:34-36` says: "The table is
    the same literals `matcher.SDPatternsNoBracketDiscount` carries, so a
    divergence ... is visible in review". After this pass, that is false for
    row 3, although deliberately so, and no test enforces it.
  - `TestStringDistFeaturingNeedsAWordBoundary`'s doc comment
    (`strdist_test.go:135-142`) describes "the leading \b". The shipped row
    will no longer contain one.
- **Fold:** Let Task 2 update the test's doc comment. Either allow a
  one-comment edit to `derive_defect_free.py`, which is ruff-gated Python
  outside the territory wall, or record it as a carry in the HISTORY draft.

### m7. No mutation proves the branch order

- **Location:** Review focus 2 (`:381-383`) and Task 2's mutations.
- **Defect:** Review focus 2 depends on branch order under leftmost-first
  matching. No named mutation proves that the golden detects a swapped order.
- **Evidence:** With the branches swapped, 827 of the 60,000 draws diverge.
  For example, `"\n(feat ft:["` should give `"\n"` but gives `"\n("`. The
  curated `(` and `[` rows would catch this.
- **Fold:** Add mutation 7, "swap the two branches: the golden fails."

## Notes, not findings

- **The field name repeats one with different behavior.**
  `SDReplacement.Replacement` is literal (`applySDReplace` uses
  `ReplaceAllLiteralString`). `SDPattern.Replacement` would be an `Expand`
  template. R-wb1's doc comment covers the difference, and a name such as
  `Template` would remove it. This is a naming choice, not a defect.
- **The case needs one override per item.** The corpus case's "an `overrides`
  entry sets every item's `album`" takes three `field` overrides, one for each
  index from 0 to 2 (`corpus/README.md`, "Items are derived"). This is
  wording only.
- **One query step is untested.** R-wb6's search query will carry
  `geschäft\: northern lines`. No committed case yet lowercases a non-ASCII
  letter in a `release:` query term. Task 3's stop row covers the risk, and I
  found nothing that points to a mismatch.

No owner fork arises under this lens.
