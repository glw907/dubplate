# Matcher featuring pattern: the Unicode word-boundary gap

The matcher's copy of beets' featuring pattern scores titles differently
from beets whenever a non-ASCII letter sits directly before `ft`, `feat`,
or `featuring`. Go's `\b` is ASCII-only and Python's is Unicode-aware, so
Go sees a word boundary inside "Geschäft" and strips the rest of the
title as a featuring credit. The measured gap is a factor of ten on the
`string_dist` figure (0.052381 in Go, 0.523810 in beets) in both matcher
modes. A two-branch rewrite of the one pattern, applied with a kept
capture group, matched beets on all 50,000 strings of a randomized
differential run. The pinned `dlclark/regexp2` engine, fed the literal
pattern unchanged, still diverged on 198 of them. This brief recommends
the rewrite, carried as a documented translation next to the unchanged
beets literal, and asks one ruling.

## Evidence

### The documented semantics

`go doc regexp/syntax` (Go 1.27.1):

```
\b             at ASCII word boundary (\w on one side and \W, \A, or \z on the other)
\w             word characters (== [0-9A-Za-z_])
```

Python's `re` documentation defines `\b` as the boundary between a `\w`
and a `\W` character. For `str` patterns, `\w` "matches Unicode word
characters; this includes all Unicode alphanumeric characters (as
defined by `str.isalnum()`), as well as the underscore". The `re.ASCII`
flag restricts it to `[a-zA-Z0-9_]`, and beets does not pass that flag.
RE2 also has no lookbehind, so Go cannot write "not preceded by a word
character" as a zero-width assertion.

### Measured figures

Go figures come from `StringDist` in the repo's matcher, run through a
`go test -overlay` scratch file so nothing in the tree changed. Beets
figures come from `beets.autotag.distance.string_dist` in the pinned
venv (`tools/venv/.venv`, beets 2.14.1). Every figure is the same in
the default table (`BeetsSDPatterns`) and the compat table
(`SDPatternsNoBracketDiscount`), because both carry the same featuring
row.

| str1 | str2 | Go | beets |
| --- | --- | --- | --- |
| `Geschäft: Ember Wharf` | `Geschäft: Northern Lines` | 0.052381 | 0.523810 |
| `Rüft: Ember Wharf` | `Rüft: Northern Lines` | 0.064706 | 0.647059 |
| `Ωft Ember Wharf` | `Ωft Northern Lines` | 0.068750 | 0.687500 |
| `Draft: Ember Wharf` | `Draft: Northern Lines` | 0.611111 | 0.611111 |
| `Ärft. Ember` | `Ärft. Northern` | 0.500000 | 0.500000 |
| `2ft Ember Wharf` | `2ft Northern Lines` | 0.687500 | 0.687500 |
| `Café ft. Northern` | `Café` | 0.071429 | 0.071429 |
| `Ember Wharf ft. Northern` | `Ember Wharf` | 0.050000 | 0.050000 |

The gap needs a non-ASCII letter or number immediately before the
trigger. `Ärft.` agrees because its boundary falls before `r`, not
before `ft`.

### Differential run

A Python script drew 50,000 random strings from a token set that mixes
the triggers with ASCII and non-ASCII letters (`ä`, `ü`, `ß`, `Ω`, the
titlecase `ǅ`), non-ASCII numbers (`²`, `Ⅻ`), a combining acute
(U+0301), `_`, a no-break space, brackets, and separators. It recorded
`re.sub(pattern, "", s)` for each. The pattern fired on 1,218 strings.
Go then substituted each string three ways.

| engine and pattern | strings that differ from beets |
| --- | --- |
| Go `regexp`, beets literal | 455 |
| `dlclark/regexp2` v1.12.0, beets literal | 198 |
| Go `regexp`, two-branch rewrite, replacement `${1}` | 0 |

The scratch files are under `/tmp/claude-1000/ubound/`.

## Fix options, ranked

### 1. Documented translation of the one row (recommended)

The rewrite is:

```
[\[\(](?:featuring|feat|ft)[\. :].+|(^|[^\pL\pN_])(?:featuring|feat|ft)[\. :].+
```

The replacement template is `${1}`. The rewrite rests on two facts
about the beets pattern.

- The trailing `\b` always holds, because the class after it (`[\. :]`)
  holds no word character. `TestStringDistFeaturingNeedsAWordBoundary`
  already records this.
- The leading `\b` is automatic after a bracket, since a bracket is not
  a word character. Without a bracket, the character before the trigger
  must be a non-word character or the start of the string.

The first branch takes the bracket case and removes the bracket, as
beets does. The second branch consumes the one non-word character
before the trigger and puts it back through `${1}`. Beets keeps that
character, because its match starts at the trigger. Go's leftmost-first
alternation tries the bracket branch first at each position, which
gives the same leftmost match that Python finds.

The kept group is load-bearing. Dropping the consumed character instead
would change the distance, because `unidecode` maps some non-word
symbols to ASCII letters: `©` becomes `(c)`, `™` becomes `(tm)`, and
`€` becomes `EUR`. That would also change the strings the later
patterns see.

`[^\pL\pN_]` matches Python's `\W`. Python's `str.isalnum()` is true for
the `L*` and `N*` categories. It is false for combining marks (`Mn`,
`Mc`), which `\pL` and `\pN` also leave out. The differential run
exercised all of these. The two languages' Unicode table versions could
disagree on a newly assigned code point, which is an edge that no real
title is likely to reach.

Cost: `SDPattern` gains a per-row replacement (or a `Keep` group
index), because a blanket `${1}` would break the rows whose group 1 is
real content, such as `(ep|single)`. The `StringDist` loop reads that
field instead of `""`. The beets literal survives unchanged as the
parity record, and the table stops being byte-identical in exactly one
row.

The parity tests would express this in three parts:

- `TestBeetsSDPatternsMatchBeetsLiterals` keeps asserting the literal
  table byte for byte against beets.
- A new translation test asserts that each row's Go form equals the
  literal unless the row appears in a named translation map.
- A committed differential golden holds (input, beets output) pairs
  from a ruff-gated Python generator in `testdata/`, the same way as
  tagio's golden-FLAC generator. The gate can then prove the translated
  row against beets without running Python.

The gate never runs Python, so the golden, not a live oracle call, is
what proves the translation.

This option fits the house bar. It stays in the standard library, it is
one regular expression, and every departure from beets is named where a
reader looks for it.

### 2. A hand-written matcher for the featuring row

A small function (scan for the trigger, check `unicode.IsLetter` and
`unicode.IsNumber` on the rune before it, and require `[. :]` after it)
would also be exact. It breaks the table's uniform shape: one row
becomes code, and `Tunables.StringDistPatterns` can no longer carry it
as data. That costs clarity against option 1 and buys nothing.

### 3. Pre- or post-processing around the literal

One form masks non-ASCII letters before the pattern runs and restores
them afterward. Another re-checks each Go match against the rune before
it. Both keep the table byte-identical, but they add a hidden second
step to every row. They also have to reproduce Python's leftmost-match
choice by hand. The code is harder to follow than option 1 and the
result is no more exact.

### 4. `dlclark/regexp2`

Adoption evidence:

- License: MIT.
- Latest release: v1.12.0, already in the workstation module cache.
- Importers on pkg.go.dev: 1,021, including chroma, the tiktoken-go
  ports, and mihomo.
- Maintainer: one (Doug Clark).

The package's own README says "You'll likely be better off with the RE2
engine from the regexp package".

It keeps the table byte-identical, but it does not deliver beets'
behavior. Its `IsWordChar` follows .NET (`L`, `Mn`, `Nd`, `Pc`, ZWJ,
ZWNJ). Python's `\w` differs from that on combining marks, on `No` and
`Nl` numbers such as `²`, and on connector punctuation other than `_`.
That difference produced the 198 diverging strings. The package would
also bring a backtracking engine that needs `MatchTimeout` discipline.
Decline.

### 5. Accept and document

This is the status quo: `tunables.go:371-375` already records the gap.
It costs nothing today. The shadow run will meet the gap as an
unexplained divergence whenever a member uploads a German, Nordic, or
Greek title with `ft` after a letter. Every such divergence then needs
an adjudication.

## Other hits in `internal/engine`

| location | construct | same class? |
| --- | --- | --- |
| `matcher/tunables.go:378,401` | featuring row, both tables | yes, the subject of this brief |
| `musicbrainz/client.go:65` | `releaseIDPattern`, `\w{8}(?:-\w{4}){3}-\w{12}` | yes, already documented at :61-64; no committed `mb_albumid` is non-ASCII |
| `matcher/album_media.go:44` | `(?i)(\d+x)?(…)`; Python `\d` also matches non-ASCII digits | yes, low risk: media values are MusicBrainz format names |
| `matcher/album_media.go:44`, `album_country.go:39`, `harness/conform/component.go:399` via `CompilePatternOption` | configured pattern text compiled as RE2 | yes, latent: any `\b`, `\w`, `\d`, or `\s` an operator writes means ASCII in Go and Unicode in beets |
| other `SD_PATTERNS` rows | no `\b`, `\w`, `\d`, or `\s` | no |
| `tagio/encode.go:372-373` | explicit `[0-9]` | no |
| `shadow/oraclereport.go:222` | `\d` over generated file names | no, not a beets-parity surface |

## Proposed corpus case

`strdist-feat-unicode-boundary`, in family F4 (`strdist-*`). It follows
the shape of `strdist-feat-strip` and the steps in `corpus/README.md`
under "Authoring a synthetic case".

- **Baseline:** a clean 3-track release.
- **Perturbation:** the candidate album is `"Geschäft: Ember Wharf"`.
  Every item's `album` is overridden to `"Geschäft: Northern Lines"`.
- **Response hand-edits:** none, except the empty-search response that
  step 3 of the authoring steps names, if freeze reports that miss.
- **Expected beets result:** the `album` key's `string_dist` is 0.523810
  in beets and 0.052381 in Go.
- **Predicted verdict:** the album weight (3.0) times 0.5238 puts
  beets' distance near medium, and Go's figure puts Go at strong. The
  verdict is therefore predicted to flip between the two, and freeze
  records the real value.
- **Claims:** `penalty_keys_nonzero: ["album"]`.

The case pins both the default and compat modes, because the two
tables share the row. `TestStringDistFeaturingNeedsAWordBoundary`
should gain the `Geschäft`, `Rüft`, and `Ωft` rows as unit-level
companions.

## Recommendation

Take option 1. It keeps the change in the standard library, it has an
exact differential record, and it confines the departure from beets to
one named row. The literal table stays the parity record. A translation
map and a committed beets-derived golden prove the one translated row.
Land it with the corpus case above, in the rung that next touches the
matcher. Leave the `CompilePatternOption` configured-pattern class
documented only, since operators supply that text.

## Ruling question for Geoff

Should the matcher carry a documented Go translation of beets' featuring
pattern so it scores non-ASCII titles the way beets does?

- **Yes:** `SDPattern` gains a replacement field. Row 3 of both tables
  becomes the two-branch rewrite with `${1}`. The literal table, a
  translation-map test, a ruff-gated differential golden, the
  `strdist-feat-unicode-boundary` case, and three new unit rows land
  with it.
- **No:** the tables stay byte-identical. The `tunables.go` note stays
  as the record. Every non-ASCII `…ft:` title the shadow run meets is
  adjudicated as a known Go-side divergence.
