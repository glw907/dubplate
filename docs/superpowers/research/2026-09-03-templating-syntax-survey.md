# Modern templating syntax survey for the layout language

Dispatched 2026-09-03 (Geoff, evening session) to ground the native
layout language's concrete syntax in modern prior art, after the
break-with-the-past ruling killed the beets DSL. Companion to
`2026-09-03-library-organization-config-direction.md`, whose syntax
section this survey resolves. Covers the *arr naming tokens, Picard,
yt-dlp, Python/.NET format specs, the Liquid and Mustache lineages,
ICU MessageFormat and Fluent, F2, PowerRename, and beets as the
error-UX baseline; three complete syntax candidates rendered over the
same three hard cases; recommendation: Candidate B.

---

## (a) Comparative catalogue**Lidarr / Sonarr / Radarr (the *arr family).** Tokens are `{Human Readable Name}` inside a single path string, with `/` making folders. Numeric padding uses a colon slot (`{track:00}`, `{medium:00}`, `{season:00}`). Empty handling is an optional group: literal text placed *inside* the braces alongside the token disappears with it. The wiki states it plainly: "Any token that resolves to an empty string Lidarr omits, including any surrounding literal text that's inside the same `{}`. This is how conditional formatting works." Multi-disc is not a conditional at all; Lidarr ships a second named template, **Multi-Disc Track Format**, selected by the release. Case, word separator, and colon replacement are global dropdowns (Space/Period/Underscore/Dash; Default/Upper/Lower; Delete/Dash/Space Dash/Space Dash Space/Smart/Custom), not inline modifiers. What it got right: the optional group, the second template instead of an `if`, global sanitization policy, and a live preview. What users hit: (1) the colon slot is overloaded four ways, meaning zero-pad (`{season:00}`), truncate (`{Album Title:150}`), language filter (`{MediaInfo Full:EN+DE}`), and named-argument select (`{Custom Format:AMZN}`), disambiguated only by which token precedes it; (2) transforms are baked into token *names*, producing `{Series Title}`, `{Series CleanTitle}`, `{Series TitleThe}`, `{Series CleanTitleThe}`, `{Series TitleYear}`, `{Series CleanTitleYear}`, `{Series TitleTheYear}`, `{Series CleanTitleTheYear}`, `{Series TitleWithoutYear}` and three more, twelve tokens for one field; (3) because token names contain spaces, the parser hunts for a known token inside the brace body, so a misspelling silently degrades to literal text and an optional group spanning two tokens has to be written as `{[Mediainfo AudioCodec}{ Mediainfo AudioChannels]}`, which is in the official TRaSH recommended scheme, brackets split across brace groups and the token spelled `Mediainfo` where the table says `MediaInfo`; (4) whitespace placement inside versus outside the braces is a live footgun, and the official Lidarr wiki ships both spellings on the same page, Davo's `{Album Title} {(Album Disambiguation)}` which leaves a trailing space when empty, and JasonVelocity's `{ (Album Disambiguation:150)}` which does not. Escaping a literal brace is by doubling, documented only in a forum answer: "You need to use double braces, `{{tvdb-{TvdbId}}}` will do it."

**MusicBrainz Picard.** Variables are `%name%`, functions are `$name(a,b)`, and the whole thing is a real language with `$if`, `$if2`, `$and`, `$gt`, `$num`, `$left`, `$set`. The shipped default naming script is the cautionary artifact:

```
$if2(%albumartist%,%artist%)/
$if(%albumartist%,%album%/,)
$if($gt(%totaldiscs%,1),$if($gt(%totaldiscs%,9),$num(%discnumber%,2),%discnumber%)-,)
$if($and(%albumartist%,%tracknumber%),$num(%tracknumber%,2) ,)
$if(%_multiartist%,%artist% - ,)
%title%
```

That is exactly the semantics dubplate's layer 1 covers (artist folder, album folder, disc prefix when multi-disc, padded track, track-artist prefix on VA), expressed as five nested conditionals. Parentheses in literal text need backslash escaping. The docs' own reassurance is the tell: "just don't get intimidated by all the '$', '%' and parentheses." Picard is the opposite pole and it validates the exile of conditionals into options.

**yt-dlp.** The documented field grammar is `%(name[.keys][addition][>strf][,alternate][&replacement][|default])[flags][width][.precision][length]type`. Nine punctuation operators plus a mandatory trailing type character. Why it hurts: the trailing `s` is required and invisible in intent, so `%(title)` is the single most common beginner error; `%` collides with Windows batch variables, requiring `%%(title)s` in `.bat` files, which is documented as a FAQ; `>` for strftime nests one cryptic language inside another; and `|default` fills missing values with the `--output-na-placeholder`, which in a path means real directories named `NA`. Every operator is a sigil and none is guessable. yt-dlp is the case for keeping the operator count near zero.

**Python format spec and .NET interpolation.** `{field:spec}`, escape by doubling: "If you need to include a brace character in the literal text, it can be escaped by doubling: `{{` and `}}`." The spec mini-language is one uniform grammar (`[[fill]align][sign][#][0][width][grouping][.precision][type]`) applied identically to every field, which is why nobody memorizes per-field rules. This is the most widely known modifier syntax alive, and `:02` is its zero-pad spelling.

**Liquid (Shopify).** `{{ object | filter: arg }}`, filters chain left to right, arguments after a colon. Readable and beloved, but the filter set is large and still growing, and the escaping story is bad: a literal `{{` requires `{{ '{{' }}`. Filters are function calls, so adopting the pipe form invites exactly the vocabulary creep the ratified direction excludes.

**Mustache and Handlebars.** Mustache is the logic-less ancestor and its own limits produced Handlebars, which added `if`/`unless`/helpers. The standard criticism, from Boronine's widely cited piece, is that logic-less templates force "a bunch of data-massaging in your code before running it through the template system." That critique is worth naming honestly, because dubplate's answer is that the data-massaging place is deliberate and named (the pre-render field hook), and the recurring cases are declared options rather than user code. Mustache's real lesson is the failure mode: users hit the wall and the ecosystem grew a helper language around it, so a logic-less layer needs its escape hatch shipped from day one, not discovered.

**Jinja2 / Ansible / Helm.** The overkill pole. The consensus in the config-management literature is that templates that become mini-programs are unmaintainable, that "huge logic blobs belong in variables, not in the template," and that mixing YAML, Jinja, and Python is itself a leading source of complexity. Nothing here argues for adopting Jinja for a path template.

**ICU MessageFormat and Mozilla Fluent.** These are the most directly relevant design thinking. ICU expresses variation as declared named cases: `{gender_of_host, select, female {…} male {…} other {…}}` and `{num_guests, plural, offset:1 =0 {…} =1 {…} other {…}}`. There is no general `if`. Fluent goes further and writes the rule down: select expressions and variants exist "only when the language grammar or style requires them, not for UI logic," under an explicit **Principle of Least Power** ("The less powerful the language, the more you can do with the data stored in that language"), and a syntax "designed for reading first, editing second, writing from scratch last." Fluent's escaping cost is instructive too: ICU escapes literal braces with ASCII apostrophes, so `"This '{isn''t}' obvious"` renders `This {isn't} obvious`, a rule nobody remembers correctly. Take the case-declaration idea; do not take the escape rule.

**Date formats.** The move from strftime's `%Y/%m/%d` to Unicode/CLDR and Luxon's `yyyy-MM-dd` is the clearest evidence that readable tokens beat sigils for anything a non-programmer edits. The relevant conclusion for dubplate is narrower and more useful: the survey shows path templates need a *year*, not a date. `$year` at 30.6% and `$original_year` at 11.3% are the whole demand. Ship `{year}` and `{original_year}` as plain fields and ship no date-formatting mini-language at all.

**F2 (Go batch renamer, ID3-aware).** The closest tool in shape to dubplate's layer 2. Variables are `{f}`, `{p}`, `{ext}`, `{id3.artist}`, with dot-chained transforms like `{f.up}` and indexed parents like `{3p.up}`. The significant fact is the migration direction: F2 previously used `{{var}}` and **moved to single braces**, keeping double only for backward compatibility. Its weak point is silent failure: "missing attributes will be replaced with an empty string to ensure the renaming process continues smoothly," which leaves orphan separators in filenames.

**PowerRename (PowerToys).** The 2025 "Smart Rename Template Language" proposal uses `$FOLDER`, `$NUM:2`, `$EXT`, and its own worked example, `$FOLDER_$NUM:2$EXT`, demonstrates the bare-sigil delimitation bug in the proposal text itself: nothing says whether the variable is `FOLDER` or `FOLDER_`. It also confirms live preview as an expected feature.

**Beets (the incumbent, for the error-UX baseline).** The documentation states the behavior outright: if you write `$foo` and `foo` is not a field, `$foo` appears in the resulting path, and the same is true of syntax errors like unclosed braces, so "if you ever see template syntax constructs leaking into your paths, check your template for errors." A typo becomes a directory.

## (b) Convergence findings

1. **Single braces won.** The audience's own tools (`{Album Title}`), the most familiar general syntax (Python/.NET), ICU, and the closest modern tool (F2, which migrated to it) all use `{field}`. Double braces belong to HTML templating where the delimiter competes with markup, which is not this problem.
2. **Escape by doubling.** Python documents it, .NET uses it, and Sonarr answers it the same way (`{{tvdb-{TvdbId}}}`). ICU's apostrophe rule and Liquid's `{{ '{{' }}` are the counterexamples nobody gets right.
3. **Recurring variation becomes a declared case, never a conditional.** Lidarr ships a second Multi-Disc Track Format. ICU has `plural`/`select` cases. Fluent has variants plus an explicit ban on using them for logic. This is precisely the ratified exile of multi-disc, disambiguation, and VA routing into options, and the best i18n design thinking in 2026 arrived at the same shape independently.
4. **Empty handling is an optional group, not a default value.** *arr's brace-group omission covers every empty case that actually occurs. yt-dlp's `|default` and `--output-na-placeholder` produce literal `NA` path segments, which is the wrong answer for filesystems.
5. **Whole-library string transforms are policy, not inline modifiers.** Sonarr puts word separator, case, and colon replacement in global dropdowns. Lidarr has one Replace Illegal Characters switch. Beets' two most-adopted transforms are global scalars (`replace:` at 30.6%, `asciify_paths` at 21.8%), and its inline transform functions are all under the cut line (`%the` 12.9%, `%asciify` 8.1%, `%lower` 6.5%). Sanitization has to be per-path-segment and uniform to be correct, which is a policy property, not a per-call one.
6. **Encoding transforms in token names is the anti-pattern.** `{Series CleanTitleTheWithoutYear}` is the end state. A modifier grammar with a tiny vocabulary is strictly better than a large token roster, and a policy is better than both when the transform is library-wide.
7. **The modifier slot must mean exactly one thing.** Python's format spec is one grammar applied uniformly. Sonarr's colon means four things depending on the preceding token, which is the readability failure users actually report.
8. **Live preview is table stakes.** Lidarr, Sonarr, and even the PowerRename proposal all have it. The direction doc's `layout preview` is aligned with the field.
9. **Every tool in this domain fails silently on a typo'd field**, beets by rendering `$foo` literally, F2 and Lidarr by rendering empty, yt-dlp by rendering `NA`. Not one validates at config load. The S2 ruling is the single most differentiating decision in this design.

## (c) Three concrete candidates

Shared data for all three.

- **Case A**, multi-disc classical box: `albumartist` = Ludwig van Beethoven, `album` = The Complete Symphonies, `disc` = 3 of 5, `track` = 4, `title` = `Symphony No. 5 in C minor, Op. 67: IV. Allegro`, no disambiguation, `sanitize: standard`.
- **Case B**, VA compilation with a same-named album: `albumartist` = Various Artists, `album` = Greatest Hits, `artist` = The Chemical Brothers, `track` = 7, `title` = `Block Rockin' Beats`, a second Various Artists "Greatest Hits" from 1988 exists so `disambiguation: auto` yields the value `1994`, routed by the `comp` rule.
- **Case C**, non-Latin with ascii-safety: `albumartist` = 坂本龍一, `album` = 音楽図鑑, `track` = 2, `title` = 羽の林で, `sanitize: ascii`.

Target output in all three candidates:

```
Ludwig van Beethoven/The Complete Symphonies/Disc 3/04 Symphony No. 5 in C minor, Op. 67 - IV. Allegro.flac
Compilations/Greatest Hits (1994)/07 The Chemical Brothers - Block Rockin' Beats.flac
Sakamoto Ryuichi/Ongaku Zukan/02 Hane no Hayashi de.flac
```

### Candidate A: Servarr-native tokens

```yaml
album:
  path:            "{Album Artist}/{Album Title}{ (Album Disambiguation)}/{track:00} {Track Title}"
  path_multi_disc: "{Album Artist}/{Album Title}{ (Album Disambiguation)}/Disc {medium:0}/{track:00} {Track Title}"
comp:
  path:            "Compilations/{Album Title}{ (Album Disambiguation)}/{track:00} {Track Artist} - {Track Title}"
sanitize: standard   # or: ascii
```

Renders all three cases correctly. Maximum familiarity: a Lidarr user can paste their existing format string and it mostly works. The costs are the ones Lidarr already pays. Token names contain spaces, so the parser must find the longest known token inside a brace body, which means `{Album Dismabiguation}` cannot be distinguished from an optional group containing only literal text and degrades to a literal rather than an error. That single property forfeits the S2 ruling, which is the design's most valuable piece. It also inherits the `:00` versus `:150` overload and pressures the token roster toward variant names.

### Candidate B: braced snake fields, nested optional groups, one format spec

```yaml
album:
  path: "{albumartist}/{album}{ ({album_disambig})}/{track:02} {title}"
  multi_disc: disc-folders
  disambiguation: auto
comp:
  path: "Compilations/{album}{ ({album_disambig})}/{track:02} {artist} - {title}"
sanitize: standard   # or: ascii
```

`layout show` expands `multi_disc: disc-folders` to the second template it selects, so nothing is hidden:

```yaml
  path_multi_disc: "{albumartist}/{album}{ ({album_disambig})}/Disc {disc}/{track:02} {title}"
```

Grammar, complete:

- `{name}` where `name` is a declared field interpolates it.
- A brace body containing at least one nested `{field}` is an **optional group**: it renders in full when every nested field is non-empty, and renders as nothing otherwise. `{ ({album_disambig})}` is the whole empty-handling story.
- A brace body that is neither a declared field nor contains a nested field is a **config-load error**. This is what buys S2.
- `{{` and `}}` are literal braces.
- `/` separates path segments. Each segment is sanitized independently under the library policy.
- Field values are inserted verbatim and never rescanned, so an album genuinely titled `{Awayland}` or `()` needs no escaping and cannot inject syntax.

Renders all three cases correctly. Case A's `Disc 3` segment comes from the option, not the template. Case B's ` (1994)` comes from the optional group collapsing on Case A and Case C. Case C's template text is byte-identical to the others; only `sanitize: ascii` differs.

### Candidate C: segment list with pipe filters

```yaml
album:
  folders:
    - "{{ albumartist }}"
    - "{{ album }}{{ album_disambig | wrap: ' (%s)' }}"
  file: "{{ track | pad: 2 }} {{ title }}"
  multi_disc: disc-folders
comp:
  folders:
    - "Compilations"
    - "{{ album }}{{ album_disambig | wrap: ' (%s)' }}"
  file: "{{ track | pad: 2 }} {{ artist }} - {{ title }}"
```

Renders the same three outputs. It buys precise per-segment error locations for free and removes any question about which parts are folders. It costs more than it buys. `wrap:` and `pad:` are function calls with arguments, which the ratified constraint excludes, and once one filter exists the Liquid trajectory toward fifty is well documented. The `{{ }}` delimiter needs a worse escape than doubling. Four YAML keys are needed where one string sufficed, the comp variant duplicates the whole structure, and the format can no longer be pasted into a forum post as a single line, which is how the *arr community demonstrably shares configurations (TRaSH Guides publishes copy-paste one-liners).

## (d) Recommendation

**Ship Candidate B.**

It sits inside the audience's existing muscle memory (single braces, `/` for folders, an optional group written as literal text inside the braces) while removing the three defects that make the *arr syntax hard to teach. Snake-case single-word field names make the parse unambiguous, which is what turns a typo into a config-time error instead of a directory. The optional group is the only empty-handling mechanism, matching what the survey and the *arr community formats show actually occurs (disambiguation absent, track artist absent on non-VA, year absent, album absent on singletons), all of which are one field plus punctuation. Nothing in the grammar is a conditional or a function call.

**Modifier vocabulary, complete and frozen: one.**

- `:0N` zero-pads an integer field to width N. `{track:02}`, `{track:03}`, `{disc:02}`.

That is the entire list. Justification for each refusal: padding as a *default* is already the `numbering: padded` option, so the inline form is an escape hatch for the rare custom layout; case, article moving ("The" at the end), ASCII transliteration, illegal-character replacement, colon replacement, and length capping are all library-wide policies, because correctness requires them applied to every segment uniformly and because the survey's adoption data puts each of them in the global-scalar column, not the per-call column. A refusal rule should be written down alongside the Prettier-style frozen preset roster: a proposed modifier that a library policy or a declared option can serve is refused, and a proposed modifier that needs real computation belongs in the pre-render field hook.

Choose `:02` over Servarr's `:00`. It is the printf/Python/.NET/Go/Rust spelling, it scales to `:03` without ambiguity, and it leaves the colon slot with exactly one meaning forever, which is the specific thing Sonarr lost. Absorb the compatibility cost in the error message: `{track:00}` should produce a named error saying dubplate uses the width spelling and suggesting `{track:02}`. The alternative (matching `:00` for audience familiarity) is defensible, and if the rung spec prefers it, the cost is a permanently ambiguous slot the moment any second numeric modifier is ever wanted.

**Escaping rule, complete.** `{{` renders `{`. `}}` renders `}`. Field values are inserted verbatim and are never re-parsed, so metadata containing braces, dollar signs, or percent signs is safe by construction. `/` is the path separator in the template and is never producible from a value, since sanitization replaces it per policy.

**Two clarifications the rung spec should absorb.**

First, `multi_disc: disc-folders` cannot be a string substitution; it selects a second template, exactly as Lidarr's Multi-Disc Track Format does, and layer 2 must let a user write `path_multi_disc` explicitly. Otherwise the option has to guess where in the path the disc segment goes.

Second, `disambiguation: auto` produces the disambiguator *value* (`1994`), and the template places it (`{ ({album_disambig})}`). If the option produced `" (1994)"` including punctuation, every template would render double parentheses. The direction doc's example comment (`same-named albums get " (1969)"`) reads as the rendered result, which is right, but the field's value must be bare.

## (e) The config-time error for a typo'd field

Model the shape on Terraform's diagnostic (severity line, file and position, the offending source line echoed with its line number, a plain-English statement, then a suggestion), add Rust's "available fields are" enumeration, and adopt Elm's assistive tone. Print:

```
Error: unknown field "albumartis"

  dubplate.yaml, album.path — line 12, column 22

     12 │   path: "{albumartist}/{albumartis}/{track:02} {title}"
        │                         ^^^^^^^^^^

  There is no field named "albumartis". Did you mean "albumartist"?

  Album fields:  album  albumartist  artist  disc  title  track  year
  Also declared: album_disambig  collection

  Nothing was renamed. Run `dubplate layout fields` for the full list with example values.
```

The load-bearing details, each earned from the survey above. The offending line is echoed with a caret span, because a path template is one long string and a bare "line 12" does not locate the token. The suggestion is a single nearest match by edit distance, following Rust's practice, rather than a list of candidates. The full available-field roster is printed inline, because the roster is seven names plus declared attributes and the user is one glance from the answer, which is the specific thing every tool in this domain makes you leave the terminal for. The closing line states that nothing was renamed, because the entire point of the ruling is that this class of mistake used to reach the filesystem, and beets documents that it does.

Three sibling errors deserve the same treatment: an unclosed group (`Error: unclosed "{"` with the caret on the opening brace, since beets documents this one leaking into paths too), an optional group with no field in it (`Error: "{ (1994)}" contains no field; write it as literal text, or did you mean "{ ({album_disambig})}"?`), and the Servarr habit (`Error: "{track:00}" — dubplate spells padding as a width. Did you mean "{track:02}"?`).

Sources: [Lidarr File Naming Guide](https://wiki.servarr.com/lidarr/naming-guide), [Sonarr Settings](https://wiki.servarr.com/sonarr/settings), [Sonarr forum, literal curly braces](https://forums.sonarr.tv/t/adding-a-literal-curly-brace-in-episode-naming/30527), [Picard naming script tutorial](https://picard-docs.musicbrainz.org/en/latest/tutorials/naming_script.html), [Picard defaults source](https://github.com/metabrainz/picard), [yt-dlp README output template](https://github.com/yt-dlp/yt-dlp), [yt-dlp FAQ](https://github.com/yt-dlp/yt-dlp/wiki/FAQ), [Python format string syntax](https://docs.python.org/3/library/string.html#format-string-syntax), [Liquid introduction](https://shopify.github.io/liquid/basics/introduction/), [Handlebars built-in helpers](https://handlebarsjs.com/guide/builtin-helpers.html), [Cult of Logic-less Templates](https://www.boronine.com/2012/09/07/Cult-Of-Logic-less-Templates/), [ICU MessageFormat](https://unicode-org.github.io/icu/userguide/format_parse/messages/), [Fluent Design Principles](https://github.com/projectfluent/fluent/wiki/Design-Principles), [Fluent Good Practices for Developers](https://github.com/projectfluent/fluent/wiki/Good-Practices-for-Developers), [Luxon formatting](https://github.com/moment/luxon/blob/master/docs/formatting.md), [F2 how variables work](https://f2.freshman.tech/guide/how-variables-work.html), [F2 ID3 variables](https://f2.freshman.tech/guide/id3-variables.html), [PowerRename Smart Rename Template Language](https://github.com/microsoft/powertoys/issues/40251), [beets Path Formats](https://docs.beets.io/en/latest/reference/pathformat.html), [Terraform Unsupported argument](https://support.hashicorp.com/hc/en-us/articles/46039773606547-Resolving-Unsupported-attribute-argument-Error-in-Terraform), [rustc field suggestion issue](https://github.com/rust-lang/rust/issues/81220), [TRaSH Sonarr recommended naming](https://trash-guides.info/Sonarr/Sonarr-recommended-naming-scheme/).
