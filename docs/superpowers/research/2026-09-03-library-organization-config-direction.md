# Library-organization configuration: the ratified direction

Ratified by Geoff, 2026-09-03 (evening session), as the standing input
to the organizing-engine rung's spec. Grounded in three research
artifacts of the same date: the organization-config usage survey
(`2026-09-03-beets-organization-config-survey.md`, n=124, with its
tally JSON), the configuration methodology study
(`2026-09-03-config-methodology-study.md`, suggestions S1-S6), and
the templating syntax survey
(`2026-09-03-templating-syntax-survey.md`, whose Candidate B Geoff
ratified as the settled syntax, below). The rung spec pressure-tests
the remaining specifics (option vocabulary beyond the settled core,
hook signature) at its own drafting and adversarial review; the
architecture and the template syntax below are settled and are not
re-litigated there.

## The headline ruling: break with the past

**The engine does not implement, evaluate, or accept the beets
path-format DSL. Ever.** The native layout surface below is the only
one. Beets migration is served by a separate one-time translation
tool (a script is acceptable), outside the engine, validated at the
output level by the re-scoped rung 10 parity fixtures (see the
amendment note in the build spec's ladder entry 10). This supersedes
the methodology study's "the DSL is fixed by the parity bar"
constraint: parity binds rendered paths, not template semantics.

## The surface, top to bottom

Each layer is an honest escape into the one below; every layer's
expansion is visible via `layout show`; no layer hides another.

1. **Preset plus toggles** (the default; most users stop here).
   Outcome-named presets (`standard`, `by-original-year`,
   `multi-disc-folders`, the collection-routed variant) selected by
   one wall key carrying a shipped, research-backed default (an
   application of the defaults discipline, not a wall exception —
   presets carry no site-specific value). Toggles are the survey's
   18-36% band as named options: `per_disc_numbering`,
   `original_date`, `va_name`, ascii-safety.
2. **Options plus a logic-less template — syntax ratified (Geoff,
   2026-09-03): the syntax survey's Candidate B.** The template is
   pure interpolation, one sigil pair, no conditionals and no
   function calls:

   ```yaml
   album:
     path: "{albumartist}/{album}{ ({album_disambig})}/{track:02} {title}"
     multi_disc: disc-folders    # or: prefixed, flat
     disambiguation: auto        # yields the BARE value ("1969");
                                 # the template places and punctuates it
     numbering: padded
     sanitize: standard
   ```

   **The grammar, complete and frozen — five rules:**
   1. `{name}` interpolates a declared field (snake_case, single
      word, unambiguous parse).
   2. A brace body containing at least one nested `{field}` is an
      optional group: it renders in full when every nested field is
      non-empty, and as nothing otherwise. This is the entire
      empty-handling story.
   3. A brace body that is neither a declared field nor contains one
      is a config-time error (the S2 payoff: a typo is a named
      error with a caret span, a did-you-mean, and the field roster
      printed inline — the Terraform/Rust/Elm error shape the
      syntax survey specifies in its section (e)).
   4. `{{` and `}}` are literal braces. Field values are inserted
      verbatim and never re-parsed, so metadata containing braces
      or sigils is safe by construction.
   5. `/` separates path segments; each segment is sanitized
      independently under the library policy, and `/` is never
      producible from a value.

   **Modifier vocabulary, complete and frozen: one.** `:0N`
   zero-pads an integer field to width N (`{track:02}`), the
   printf/Python/.NET spelling, leaving the colon slot exactly one
   meaning forever. The refusal rule stands beside the frozen
   preset roster: a proposed modifier that a library policy or a
   declared option can serve is refused; one needing real
   computation belongs in the pre-render field hook. `{track:00}`
   (the Servarr spelling) gets a named error suggesting `{track:02}`.

   **Two absorbed clarifications:** `multi_disc: disc-folders`
   selects a second template (Lidarr's Multi-Disc Track Format
   pattern), and a user may write `path_multi_disc` explicitly; the
   `disambiguation` option produces the bare disambiguator value,
   and the template's optional group places and punctuates it.

   The survey justifies removing control flow: observed conditional
   usage collapses into known semantic situations (multi-disc
   formatting is 82% of all inline-Python use; disambiguation, year
   presence, and VA routing are the rest), and each becomes a
   declared option. Field vocabulary is seven core fields
   (title/album/albumartist/track/artist/year/disc) plus declared
   flexible attributes. The forward-looking properties, from the
   syntax survey's convergence findings: single-brace-plus-format-
   spec is the convergent modern idiom (Python, .NET, Go, Rust,
   ICU, F2's own migration); the language is designed not to grow
   (Fluent's Principle of Least Power — new behavior arrives as
   option keys, never syntax, so adopters' templates never face a
   migration event); and the strict tiny grammar is machine-
   checkable, which every consumer (the web builder, `layout
   adopt`, editors, generated configs) inherits.
3. **The pre-render field hook** for real logic (the extensibility
   fork's lifecycle-hook seam): derive additional fields before path
   rendering, with album-scope access. Logic lives in code,
   formatting in templates. Covers the survey's genuine tail
   (primary-artist heuristics, quality labels).

Library-level concepts are first class, not template tricks:
collections are named nouns routed by the general query router (the
mechanism beets' power users hand-build via query-prefixed formats);
compilations and singletons are routed cases with choices; sanitize,
art naming, and the import file-op mode (custody: move/copy/link)
are named library-wide policies.

## The mechanisms carried from the methodology study

- **S1**: layout rules are an ordered list of named rules;
  `layout explain <album>` traces which rule matched and why
  (the `git check-ignore -v` shape). Unreachable-rule and
  undeclared-attribute checks at config load.
- **S2**: templates compile at config load; an unknown field is a
  named config-time error with a did-you-mean, never a literal
  rendered into the filesystem.
- **S3**: `layout preview` at two scales — a built-in hard-case
  sample set (multi-disc, VA, non-Latin, over-long, same-named
  albums, collection-routed), and the library-scale no-move diff
  (rung 10's machinery, one renderer, two sinks).
- **S4**: presets expand to visible layer-2 text via `layout show`
  (the `caddy adapt` shape). The preset roster is frozen
  Prettier-style: a proposed preset that an existing one plus one
  edit can serve is refused.
- **S5** (doctor/gate work, rides whichever rung next touches
  doctor): effective-config report with per-value provenance and a
  differences-only mode.
- **S6** (gate work): the env manifest, reference docs, and starter
  layout are generated from `internal/config`'s tables, with a
  drift check in the gate.
- The web UI's live layout builder (Lidarr's naming screen is the
  reference) is later web work calling the same render endpoint as
  the CLI; it never reimplements rendering.

## The migration story

A separate translator tool reads a beets config and emits the
native form: a preset-plus-toggles combination where one reproduces
the layout (the survey says this is the common case), layer-2
options-plus-template otherwise, with an explicit report of anything
untranslatable (which then belongs in a field hook or is dropped
deliberately). It is validated by the re-scoped corpus path-parity
fixture: the estate's beets config rendered through pinned beets on
the harness side versus the translated native config through the
engine, diffed case-for-case over AppliedFields records, plus the
read-only no-move diff against the live library. The estate cutover
default remains zero path churn (backups and Navidrome stay
undisturbed), proven by that diff.

## The 80% roster

The survey's section (d) cut line is the built-ins evidence base for
this surface: the seven-field core, disambiguation and the
conditional-shaped semantics as options, comp/singleton/albumtype
routing, move/copy(+link) file ops, the 20-36% policy band
(per_disc_numbering, asciify, art_filename, replace-as-policy), and
the import workflow toggles. The native multi-disc formatter is the
single highest-value built-in. Deliberately out: the rare functions
and fields, sort-name path usage (0% observed), and arbitrary inline
computation (the hook's job).
