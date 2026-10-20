# Adversarial review: the design-language standard

Fresh-context review (2026-09-07) of
`specs/2026-09-07-design-language-standard.md` revision 1, in two
independent lenses. Lens A checked stack fidelity: every DaisyUI
variable name and plugin mechanism against the published
`daisyui@5.7.17` package, every SvelteKit claim against
`svelte.dev/docs/kit`, and every source citation against the repo at
HEAD `c98a916`. Lens B checked spec fidelity and scope: every clause
against its cited line in the build spec, the 2026-09-07 erratum, the
2.0 streamer spec, the theme-collection brief, `CLAUDE.md`, the
parallel spec's Phase 0c paragraph, and the salvage tree. Folded as
revision 2 the same day.

The two reviews are merged here. Findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1, M1 through M14, and m1 through m8. Every finding
was re-verified by the fold editor before folding, and every numeric
claim was re-measured, never carried. The DaisyUI measurements were
re-taken from the published package, not from either review's report,
because neither could reach the package the standard cites. Four
reviewer claims did not survive that check in the form they were
written, and each is corrected below with the measurement. None of the
four retires its finding, and no finding was rejected outright.

Both lenses agree the standard is well grounded. Its 29-name variable
set is exactly right, measured independently below. Its account of
DaisyUI's plugin merge behavior, its reading of the ratified theming
ruling, its handling of the 2026-09-07 erratum, its three-way weighing
of the tier 2 mechanism against universality, and its salvage adoption
record all hold line by line. The defects concentrate in three places:
the CSP, which was adopted verbatim from a pre-pivot stack it no longer
fits; the arithmetic of the variable contract, where one clause says 29
and three others say 32; and the gates, several of which are written as
prose a reader can follow, not as a check a script can run.

---

## Corrections to the reviews' own evidence

**1. Lens A's M3 evidence generalizes one selector to all 35 blocks.**
It reports that `themes.css` "line 1 emits every built-in as
`:root,:root:has(input.theme-controller[value=light]:checked),[data-theme=light]`".
Re-measured against the published `daisyui@5.7.17` `themes.css`: the
file carries 35 selector groups, the bare `:root` branch appears in
exactly one of them (the `light` block, which is the compiled default),
and `theme-controller` appears 35 times. So every block carries the
`:has()` branch, and only the default carries a bare `:root`. Both
halves of the finding survive, and each rests on a different one of
those two measurements: M3's specificity inversion needs the `:has()`
branch in all 35, and the standard's own R3.5 reasoning needs the bare
`:root` in the default alone.

**2. Lens A reads the plugin's merge branch correctly and its
destructure count short.** `theme/index.js` at 5.7.17 destructures five
keys, `name`, `default`, `prefersdark`, `"color-scheme"`, and `root`,
capturing the remainder with `...customThemeTokens`. Lens A reports
five and the standard reports four, so the standard is what needs
correcting (m1).

**3. Lens B files the unreachable-package citation as a major.** Folded
as a minor. The fold editor re-measured the substance from the
published package: `themes.css` contains 28 distinct custom properties,
each appearing exactly 35 times, plus `color-scheme` 35 times, which is
29 declarations across 35 identical blocks, and the names match the
standard's two lists character for character. The contract is right.
What is unreachable is the citation, so the defect is reproducibility
not correctness. The fix is a citable URL and a stated version, and no
re-derivation is owed.

**4. Lens A's caching finding overstates the shared-cache exposure.**
It argues that "the estate serves through cloudflared, and a
static-looking `index.html` is a natural cache target". Cloudflare does
not cache `text/html` under its default cache policy, and a tunnel adds
no cache of its own, so the cross-member repaint needs a shared cache
the estate does not have today. The finding stands on the browser cache
instead: two members signing in on one browser, and any future
intermediary, are enough, and the salvage precedent already sets
`no-cache` on `index.html` (`_salvage/internal/serve/serve.go:283-293`)
for the same class of reason. Folded at minor (m6).

---

## BLOCKER

### B1. R5.6's CSP blocks SvelteKit's own inline bootstrap script, so rung 5's login page loads and never boots

**Lenses.** Both, independently, and both as their blocker.

**Standard lines.** R5.6 (`:322-332`) freezes the served CSP at
`default-src 'self'; frame-ancestors 'none'; base-uri 'none'; object-src 'none'; form-action 'self'`
and calls any addition to it "a standard violation, not a configuration
choice". G6 (`:365-366`) asserts that exact string. R4.2 (`:238-240`)
reasons from the same header: "An inline bootstrap script is not an
option: the CSP of R5.6 carries no `'unsafe-inline'`."

**Evidence.** The header is adopted verbatim from
`_salvage/internal/serve/serve.go:125`, character for character,
confirmed. That file served the salvage stack, which is standalone Vite
with a hand-written document whose only script is external
(`_salvage/web/index.html`, last body line
`<script type="module" src="/src/main.ts"></script>`), so `'self'`
permitted every script the page carried. R1.6 (`:57-64`) declines that
stack and adopts SvelteKit in static mode, per CLAUDE.md:133-136 and
the erratum at BS:16-18, and CLAUDE.md:142-143 makes rung 5's login
page the first page built under it. SvelteKit generates inline scripts
in the documents it emits. Its own configuration reference is explicit
that this is what `kit.csp` exists for: "SvelteKit will augment the
specified directives with nonces or hashes (depending on `mode`) for
any inline styles and scripts it generates". With no `script-src` and
no `'unsafe-inline'`, `script-src` falls back to `default-src 'self'`,
which does not permit an inline script.

**Consequence, built as written.** Rung 5 ships a login page that
returns 200, paints the fallback document, and never hydrates. The
failure is silent in the served bytes and visible only in the browser
console, so a gate that asserts the header (G6) passes while the page
is dead.

**The mechanism the fix has to respect.** Two constraints narrow the
choice. Nonces are forbidden on prerendered pages ("Using nonces with
prerendered pages is insecure and therefore forbidden"), and
`adapter-static` prerenders every page, so `kit.csp.mode: 'hash'` is
the only legal mode here. The Go-served header and SvelteKit's own
`<meta http-equiv>` policy combine restrictively, so a resource must
satisfy both, and the Go header cannot be stricter than the meta
policy for scripts. R4.2's stamping does not disturb this: the stamp
lands on the document element's attribute and leaves every inline
script's bytes unchanged, so a hash computed over the embedded document
stays valid for the served one.

**Fix.** Replace R5.6's frozen string with a directive set.
Keep `default-src 'self'` and the four hardening directives, add a
`script-src 'self'` that carries the hashes of exactly the inline
scripts the embedded document contains, and state hash mode as the
SvelteKit half of the contract. Rewrite G6 to assert the directive set
and the hash match against the built document, never a literal.
R4.2's stamping decision is right on its own merits and stays; its CSP
justification goes.

---

## MAJORS

### M1. The same header blocks inline style attributes, which DaisyUI's radial-progress and Svelte's own transitions both need

**Lens.** A.

**Evidence.** R5.6's header carries no `style-src`, so `style-src` and
`style-src-attr` both fall back to `default-src 'self'`, which does not
permit an inline style attribute. DaisyUI 5 documents radial-progress's
basic usage as
`<div class="radial-progress" style="--value:70;" aria-valuenow="70" role="progressbar">70%</div>`
and states that "Radial progress needs `--value` CSS variable to work".
Svelte 5's `style:` directive compiles to the same attribute.
SvelteKit's configuration reference adds a second, wider case: "most
Svelte transitions work by creating an inline `<style>` element", so a
build-time hash cannot cover them at all. The salvage tree never
exercised either case, which is why the adopted header never met them:
`grep -rn 'style=' _salvage/web/src --include=*.svelte` and the same
grep for `style:` both return nothing.

**Consequence.** Rung 7's uploader reaches for a progress indicator,
rung 5 or rung 7 reaches for a transition, and either one dies against
a header the standard forbids anyone to change.

**Fix.** Rule it in section 5, before rung 7 discovers it. A ban on
inline styles would have to ban Svelte transitions as a class, which is
a large price for a directive whose risk is CSS-shaped rather than
script-shaped, so the allowance is the right
branch: `style-src 'self' 'unsafe-inline'`, with the reason recorded
and with the contrast to `script-src`, which stays hash-pinned, stated
in the same clause.

### M2. R2.1's "exactly these 29 names, no more and no fewer" contradicts R2.7, R3.1, and G2, which all require 32

**Lenses.** Both.

**Evidence.** R2.1 (`:68-70`) declares the complete set as exactly 29
names for a composed theme, 21 palette and 8 structural. R2.7
(`:145-151`) then adds `--font-sans`, `--font-serif`, and `--font-mono`
to the structural axis. R3.1 (`:163-168`) says the match axis carries
"the eight structural variables of R2.1 plus the three font variables
of R2.7" and that "their union is the complete contract", which is 32.
G2 (`:348-350`) encodes 32 as the gate: "each match block exactly the 8
structural names plus R2.7's 3 font names".

**Consequence.** A theme authored to R2.1's literal wording ships no
font pairing and fails G2. A theme authored to G2 violates R2.1's "no
more". The clause labelled "the complete variable set" is the one that
gets read first.

**Fix.** Make R2.1 the whole contract at 32 names in three groups, 21
palette, 8 structural, and 3 font, with a note that only the first 29
come from DaisyUI's own block. R2.7 then explains the font stack's
shape and adds no names, and R3.1's union sentence becomes true as
written.

### M3. The `theme-controller` selector outranks document order, so tier 2 branding can lose to tier 1 with G7 still green

**Lens.** A.

**Evidence.** R4.1 (`:226-229`) establishes precedence by document
order and states that "No clause uses a specificity trick to reorder
them, so the rule a reviewer checks is a link order in one file".
Document order only decides between declarations of equal specificity.
Measured against the published `daisyui@5.7.17`, every one of the 35
built-in blocks is emitted at
`:root:has(input.theme-controller[value=NAME]:checked),[data-theme=NAME]`,
and the plugin emits a custom theme at the same shape
(`theme/index.js`, the `selector` assignment). The `[data-theme=NAME]`
branch is specificity (0,1,0), equal to the branding stylesheet's
`:root` (`:246-248`), so order decides and R4.1 holds. The `:has()`
branch is (0,4,1), because `:has()` takes the specificity of its most
specific argument and
`input.theme-controller[value=NAME]:checked` is (0,3,1). The
`theme-controller` class is DaisyUI 5's own idiomatic theme picker, and
R4.2 (`:231-236`) implies rung 7 builds a picker.

**Consequence.** If rung 7 reaches for `theme-controller`, tier 2
branding silently loses to tier 1 at every link order, and G7
(`:368-370`) asserts link order alone, so it passes while the cascade
is inverted.

**Fix.** Forbid the `theme-controller` input pattern in a clause. The
pick is a stamped `data-theme` attribute under R4.2, so the input is
redundant already. Promote G7 from a link-order assertion to a
computed-value assertion in G10's shape: set a branding variable,
render, read the computed value, and fail when the theme block wins.

### M4. R4.5's "no tier renders server-side HTML" contradicts R4.2 and G7, which both require the Go binary to rewrite the served document

**Lenses.** Both, from different halves. Lens A named the contradiction
and lens B named the unstated consequence for rung 5's serve shell.

**Evidence.** R4.5 (`:274-277`) states "No tier renders server-side
HTML" as a normative clause. R4.2 (`:233-236`) requires the Go binary
to stamp a `data-theme` attribute onto the `index.html` it serves. G7
(`:368-370`) requires the served document to carry two `<link>`
elements the SvelteKit build did not write. Both are server-side HTML
generation over an embedded artifact that is otherwise immutable, and
the salvage precedent serves that artifact verbatim through a static
handler (`_salvage/internal/serve/serve.go:287-297`).

**Consequence.** A reviewer applying R4.5 literally rejects R4.2's
mechanism. Rung 5 owns the serve shell and needs to know that the
document is rendered and not served byte for byte.

**Fix.** Reword R4.5 to what it means, which is that no SvelteKit
server file exists and no page is server-rendered by SvelteKit
(CLAUDE.md:133-139). Then state R4.2's and G7's mechanism once and
explicitly: the Go layer renders the embedded document through a small
documented rewrite, with a named placeholder token in `app.html` as the
testable form.

### M5. G1 cannot be run as specified, and its token list misses the most common hardcoded colors

**Lenses.** Both, independently.

**Evidence.** R2.4 (`:116-122`) defines the check as "a grep over
`web/src`" that fails on "A color literal: `#[0-9a-fA-F]{3,8}` in a
color position". A grep has no notion of a color position, so the
qualifier makes the clause unimplementable as written, and an
unqualified hex grep hits fragment hrefs, element ids, and Tailwind
arbitrary values alike. The numbered-utility pattern at `:122`
enumerates the numbered Tailwind families alone, so `bg-white`,
`text-black`, `bg-transparent`, `border-white`, `text-current`,
`lab()`, `lch()`, `hwb()`, `color()`, and every named CSS color pass
the gate while naming a palette value, which R2.3 (`:106-110`) forbids
in its first sentence. The exempt path is never given: the standard
says "the theme directory" four times and names it nowhere.

**Consequence.** R2.4 also states that "An exemption anywhere else is a
gate failure, not a comment", so the false positives an unqualified hex
grep produces have no relief valve.

**Fix.** Make it a script with an explicit file-extension scope, an
explicit token list, and a named theme directory path. Drop "in a color
position" or define it mechanically as a declaration value or a `class`
or `style` attribute's text. Add `white`, `black`, `transparent`, and
`current` as bare utility suffixes, and add `lab(`, `lch(`, `hwb(`, and
`color(` to the function list.

### M6. G9's exception mechanism is prose, so the contrast gate cannot be run deterministically

**Lens.** A.

**Evidence.** R2.6 (`:140-141`) says "A palette that cannot meet the
floor for one variable records the exception and the reason in its own
file". G9 (`:376-378`) "fails on a pair below the floor that carries no
recorded exception". A free-form record in a CSS comment is not
parseable, so the gate either cannot see the exception and reds a
ratified palette, or it is written loosely enough to swallow anything.
The standard quotes CLAUDE.md's own gate rule at `:126-127`: a check
either gates or does not exist. The color space is also unstated, and
every value in the salvage themes is `oklch()` while WCAG contrast is
defined on sRGB relative luminance.

**Fix.** Give the exception a fixed machine-readable form the gate
parses: a single-line marker directly above the declaration carrying a
fixed prefix, the variable name, the measured ratio, and the reason.
State that a marker with no matching failure is itself a gate failure,
so stale exemptions cannot accumulate. State the color space the ratio
is computed in.

### M7. R2.6's floor covers four state colors where its own mechanism covers eight, declines nothing about the salvage carve-out, and says nothing about tiers 2 and 3

**Lenses.** Both, on adjacent halves of one clause.

**Evidence.** R2.6 (`:134-139`) requires `--color-error`,
`--color-success`, `--color-warning`, and `--color-info` to clear 4.5:1
as text on `--color-base-100`, "because DaisyUI's `-soft` modifiers use
the state color itself as the text color rather than routing through
`-content`". The cited evidence records that mechanism generically:
`_salvage/web/src/app.css:32-35` reads "DaisyUI 5's `-soft` modifier
uses AS the text color itself (`.alert-soft{color:var(--alert-color)}`,
`.badge-soft` the same)". The `-soft` modifiers and `text-primary`
apply to primary, secondary, accent, and neutral identically, and a
`text-primary` link on the base ground is the most common colored-text
case a login page has. Separately, the same salvage file records a
deliberate decision the standard neither adopts nor declines
(`app.css:26-30`): "primary's white content clears WCAG's >=3:1
UI-component/large-text tier, matching the ratified mockup's exact
sample ... a stricter 4.5:1 blue would drift dark enough to leave the
mockup's range." And R2.6 is scoped "within a composed theme", while
G9 reads "each composed theme's declared values", so a tier 2 branding
override can break the floor with nothing checking it.

**Fix.** Extend the second condition to every non-content color the
framework can paint as text on the base ground, which adds primary,
secondary, accent, and neutral. Where a lower tier is deliberate, name
WCAG's 3:1 UI-component tier explicitly the way the salvage file does.
Add one sentence saying tiers 2 and 3 sit outside the floor and the
risk is the operator's, as R4.4 already does for the CSS file.

### M8. G5 passes vacuously on a fresh checkout, which is the tree it exists to protect

**Lens.** A.

**Evidence.** G5 (`:360-363`) is a Go test that "resolves every
`@font-face` src in the built bundle against the embedded filesystem
and fails on a path that does not exist there", and the standard calls
it "the check that catches a font that builds locally and 404s on the
box". The adopted embedding pattern (`:446`) is
`_salvage/web/embed.go`, whose package comment records that "a bare,
empty dist/ fails //go:embed at compile time" and that a separately
committed `dist/placeholder.txt` is what keeps a never-built checkout
compiling. Measured: `git ls-files _salvage/web/dist` returns
`_salvage/web/dist/placeholder.txt` alone. On such a checkout the
embedded filesystem holds one text file, the bundle carries no
`@font-face`, the iteration is empty, and the test is green.
CLAUDE.md's Go rule forbids exactly this, under "no vacuous
assertions, prove tests can fail".

**Fix.** Make G5 assert a nonzero face count before it iterates and
fail with a named message when the embedded tree is placeholder-only.
State in the section 6 preamble which gates require a built bundle,
which covers G3's "and in the built bundle" clause and G4 as well.

### M9. Section 7 cites S20:441-444 for a ruling that spec does not contain

**Lenses.** Both.

**Evidence.** The standard (`:406-408`) states "An external streamer
keeps its own look under its own theme config, which is 2.0's business
(S20:441-444)."
`docs/specs/2026-09-07-streamer-2-0-design.md:441-444` is
ruling 1 of the review round, and it rules the opposite half: no theme
matches Navidrome's or Jellyfin's any longer, the build spec carries a
dated erratum, and ladder entry 12 is trimmed. A case-insensitive grep
for `theme` over that spec returns hits at lines 441, 442, and 443 and
nowhere else, so it carries no ruling about a streamer's own theme
config, and none is owed to it. The attribution also runs backwards:
an external streamer's theming is that streamer's business, and 2.0 is
the built-in that replaces it.

**Fix.** Keep the claim, which is sound, and attribute it to this
standard. Cite S20:441-444 for what it says, cite BS:12-15 for the
rung 12 trim and S20:370-373 for the capabilities a driver may decline,
and drop "which is 2.0's business".

### M10. The forward-hook statement calls hook 5 neutral while R4.2 introduces exactly the per-user row that hook covers

**Lenses.** Both.

**Evidence.** The standard (`:417`) states "This standard touches
exactly one hook from S20:354 onward" and names hook 6 alone, then
(`:425-428`) declares hooks 1 through 5 and 7 through 14 neutral
because "none reaches the presentation layer". R4.2 (`:234-242`)
requires the Go binary to serve the theme "from the signed-in member's
stored pick" once "a session identity and the rung 6 store exist",
which is a per-user row. S20:367-369 is hook 5: "Rung 6 mints a durable
surrogate user id that per-user rows key on, and rung 5 resolves a
roster identity to it. The roster's `streamer_username` becomes
required, unique, and stable rather than optional, and it is the
renameable handle the surrogate id protects against." S20:347 makes
per-user state its own 2.0 slice.

**Consequence.** A theme pick keyed on a renameable roster handle is
the retrofit hook 5 exists to prevent, and telling a rung 6 planner
this standard is neutral to hook 5 invites exactly that key.
CLAUDE.md's build-for-2.0 rule requires the weighing to be stated.

**Fix.** Add hook 5 to the statement and land it in R4.2: the stored
theme pick is a per-user row keyed on rung 6's durable surrogate user
id, never on a roster username or an address. The neutral sentence
becomes hooks 1 through 4 and 7 through 14.

### M11. The 44px touch-target floor is a requirement that exists only in the adoption table

**Lenses.** Both.

**Evidence.** The adoption record (`:444`) adopts
`--size-field: 0.275rem` "as a structural-axis requirement" and rules
that "the 44px floor belongs to the `dubplate` match block rather than
to each palette". No clause in section 2 or section 3 states a size
floor, and no gate checks one. G2 (`:348-350`) asserts that
`--size-field` is present and says nothing about its value. The
standard opens (`:10-11`) by promising that "Each clause is a
requirement or an explicit non-requirement", so a rung 7 author reading
the numbered clauses never sees this one. The bare value is also
DaisyUI-version-dependent, which is the reason to state the floor
rather than the number: `_salvage/web/src/app.css:58-62` records that
"DaisyUI's 0.25rem default puts a plain .btn at 40px" and that
"0.275rem lands .btn (10x) at exactly 44px; .btn-sm (8x) still needs
its own min-h-11 override".

**Fix.** Promote it to a numbered clause on the structural axis, with
the `.btn-sm` caveat, and extend G2 to assert the `dubplate` match
block's `--size-field` meets it. Leave the table row as provenance.

### M12. R3.2 keeps the match key on a rationale the erratum retired, and section 8 of the same document contradicts it

**Lens.** B.

**Evidence.** R3.2 (`:170-176`) states that the axis stays in config
"because 2.0's built-in streamer arrives as one more driver (BS:15-16)
and removing the key now makes its return a config-schema change
later". BS:15-16 reads "The seam stays universal because 2.0's
built-in streamer arrives as one more driver", which is about the
driver seam and says nothing about a match value returning. BS:7-11
retires streamer-matched themes as a class. Section 8 of this same
document then argues the opposite (`:420-423`): "An in-process streamer
has no theme to match, and 2.0's built-in streamer serves dubplate's
own surfaces, which R1.1 already binds."

**Consequence.** The clause is marked provisional with a conductor
ruling owed on whether a one-valued axis earns its key, and the
conductor is asked to rule on a premise the erratum voided.

**Fix.** Restate the reason or drop the key. If the key stays, justify
it on a ground that survives the erratum, which is config-schema
stability for a future adopter-authored structural block, and reconcile
the sentence with section 8.

### M13. G4 rejects the license files R5.2 requires beside each face

**Lens.** B.

**Evidence.** R5.2 (`:292-294`) requires that "The tree carries each
face's license text verbatim beside the file, plus a manifest naming
the family, the version, the upstream source URL, and the license URL".
G4 (`:356-358`) requires that "Every file in the font directory ends in
`.woff2`". Put the license text and the manifest where R5.2 says, and
the gate reds on them.

**Fix.** Scope G4 to font files: every font file in the directory is a
woff2 and no other font format is present. Name where the license texts
and the manifest live, so the gate reads them instead of rejecting
them.

### M14. The composition model and the single `data-theme` value are never reconciled, and rung 5 has to stamp something

**Lens.** B.

**Evidence.** R3.4 (`:188-196`) leaves the mechanism open: "The layers
compose at runtime as separate variable sets ... Whichever mechanism
rung 7 chooses, that ordering is the requirement." G10 (`:380-381`)
tests a computed structural variable "under a built-in palette", which
reads as two blocks applying at once. R4.2 (`:231-236`) fixes the
selector: "The pick is a `data-theme` attribute on the document
element", stamped by the Go binary. A single `data-theme` value selects
a single DaisyUI block. Nothing states how two axis picks become one
attribute value, whether a second attribute exists, or whether composed
themes are a precomputed cross-product.

**Consequence.** Rung 5 builds the stamping and cannot build it against
an unstated value. The measured shape of the plugin narrows the answer
further than the standard admits: `theme/index.js` emits every block,
built-in and custom alike, at `[data-theme="<name>"]`, and R1.4 forbids
a theme from introducing a selector of its own, so a match layer
authored as data cannot land on a second attribute without breaking
R1.4.

**Fix.** State the mapping. `data-theme` carries one composed identity
spelled `<match>-<palette>`, which is what rung 5 stamps and what G10
reads against. Leave the mechanism that produces the composed block to
rung 7 under R3.4's precedence rule, and record that R1.4 plus the
plugin's one-block-per-value shape points at build-time composition
from two authored sources.

---

## MINORS

**m1. R2.2's destructure count is four and the plugin's is five, and
`color-scheme` is both counted and excluded.** R2.2 (`:99-104`) says
the plugin "destructures exactly those four keys" before treating the
rest as tokens, and that metadata "is excluded from R2.1's count".
Measured, `daisyui@5.7.17` `theme/index.js` destructures `name`,
`default`, `prefersdark`, `"color-scheme"`, and `root`, then spreads
the remainder. R2.1's palette list (`:75`) counts `color-scheme` as one
of the 21, and TB:125 also files it under "block metadata". State the
five keys, and say that `color-scheme` is destructured as metadata and
re-emitted as a CSS property, which is why it counts among the 21
declarations while not being a custom property. Lens A.

**m2. R2.1's provenance cannot be re-verified from this repo, and it
was read at a version the tree does not pin.** The standard (`:25-28`)
cites "an installed DaisyUI 5.7.17 package
(`node_modules/daisyui/themes.css`)". No `node_modules` directory
exists anywhere in the repo. The only DaisyUI pin in the tree is
`_salvage/web/package.json:21`, daisyui 5.7.25, which the standard's own
adoption row declines at `:448`. Cite a fetchable URL beside the path,
state the version the count was taken at, and say that the tree's own
pin differs. The substance was independently re-measured and holds; the
citation is what is unusable. Both lenses, folded at the lower of the
two severities.

**m3. R1.1 calls the first-party web player 2.0's and omits the 2.0
surface that exists today.** R1.1 (`:35`) binds "2.0's first-party web
player (S20:433)". S20:433 sits under the heading "## The 2.1 list" at
S20:430, and section 8 of the standard gets this right at `:428` ("The
2.1 first-party web player"). The 2.0 surface this standard does bind
goes uncited: S20:303-306 places the device page and share management
"in the same web app as rung 5's login". Say 2.1 in R1.1 and add the
S20:303-306 citation. Both lenses.

**m4. R1.6's claim that every mechanism works unchanged for a
Go-templated page is false for R4.2's client half.** R1.6 (`:61-64`)
asserts that every mechanism in sections 4 and 5 "also works unchanged
for a Go-templated page, because none of them runs JavaScript before
first paint". R4.2 (`:236-237`) requires the client to change the
attribute after hydration, and `:241-242` recommends `localStorage` for
the pre-rung-6 pick. Both need JavaScript, and hydration does not exist
on a Go-templated page. Narrow the claim to what is true: nothing in
sections 4 and 5 needs JavaScript before first paint, so a Go-templated
fallback renders correctly, and the picker interaction is the one part
that would need a different implementation. Both lenses.

**m5. Three gates assert less or more than their clauses ask.** G2
(`:348-350`) checks font-variable names alone, so R2.7's requirement
(`:148-149`) that each stack end in a CSS generic family is unenforced;
extend G2 to assert the tail. R5.4's prohibition on trimming glyphs
(`:306-312`) has no gate at all, and section 6's silence is left to be
inferred; say in section 5 that it is a review rule. G6 (`:365-366`)
asserts the exact CSP string and then separately asserts no `font-src`
is present, which the exact-string assertion already proved. Lens A
recommended dropping G6's second assertion; the fold keeps it and drops
the frozen string instead, because B1 replaces the string with a
directive set that no longer implies the absence. Lens A.

**m6. R4.2 makes the served document per-member without saying anything
about caching it.** R4.2 (`:233-236`) has the Go binary stamp the
member's stored theme into the served `index.html`. Nothing in section
4 states a cache header for that document. The salvage precedent
already answers it for the same class of reason
(`_salvage/internal/serve/serve.go:283-293`: `no-cache` on index.html,
`private, max-age=31536000, immutable` on hashed assets). Add one
sentence to R4.2 and the header assertion to G7, which already reads
the served document. Lens A, with the shared-cache half corrected
above.

**m7. Citation drift in eight places.** Each cited range is off by one
to three lines against the source read today. R2.1's "the same list
appears at TB:118-125" (`:97`) is TB:120-126, since TB:118 is the
dachinat port bullet. R1.4's TB:200-204 (`:50`) is TB:202-206, since
TB:200 is the section heading and the no-plugin-ABI clause runs to 206.
R2.8's BS:682-684 (`:154`) is BS:683-684, since BS:682 is the Jellyfin
typography line. R3.4's BS:689-691 (`:195`) omits the built-ins
sentence, which runs BS:691-693. R4.4's TB:151-156 (`:270`) is
TB:151-157, since the "never means not possible" clause closes on 157.
R2.6's `app.css:32-52` (`:142`) is `app.css:32-47`, since 48-52 are the
`@import` and `@plugin` lines. R1.5's CM:133-144 (`:55`) is CM:133-143,
since CM:144 begins the systemd bullet. Section 8's S20:370-374
(`:423`) is S20:370-373, since S20:374 begins hook 7. Re-key all eight.
The standard's method is that each clause names a re-verifiable line,
so drift costs more here than in ordinary prose. Lens B, extended by
the fold editor.

**m8. R3.3 routes the theme keys through rung 1b's wall-row work they
do not need.** R3.3 (`:178-186`) registers both keys "through rung 1b's
registry so no lane edits `internal/config/keys.go` (PA:54-55)".
PA:54 scopes rung 1b to extending `Register()` to wall rows and
reserved classes, which is what `keys.go` holds (`ownedWall`,
`keys.go:19-25`). The same clause correctly says both keys carry
defaults, which makes them thresholds under the existing `Register()`
(`internal/config/thresholds.go:40-42`), a path that never touches
`keys.go` today. The only thing genuinely owed is the string kind:
`thresholds.go:15-18` declares `KindInt` and `KindDuration` alone. Say
that, and keep the provisional ruling on whose rung mints the kind.
Lens B.

---

## Rejected

None. Every finding from both lenses survived re-verification in
substance. Four carried evidence that did not survive as written, and
all four are corrected above with the measurement, and none was dropped.
One (m2) was folded at a lower severity than its lens filed it, for the
reason recorded in correction 3.

---

## Verified sound

Every item below was checked by at least one lens and re-checked by the
fold editor wherever a fold depended on it. Items 1 through 8 were
re-measured by the fold editor against the published `daisyui@5.7.17`
package, since neither lens could reach the package the standard cites
and the standard's whole variable contract rests on it. This is the
do-not-re-litigate record for the standard.

1. R2.1's 29-name set, measured against the published
   `daisyui@5.7.17` `themes.css`. The file contains exactly 28 distinct
   custom properties, each appearing exactly 35 times, plus
   `color-scheme` 35 times: 29 declarations across 35 identical blocks.
   The names match the standard's two lists character for character,
   including `--noise`, `--depth`, `--border`, `--size-selector`, and
   `--size-field`. "Every block declares the same 29 names" is true.
2. The 21/8 split. The palette list is `color-scheme`, four base names,
   and eight colors with their paired `-content`, which is 21. The
   structural list is three radius names, two size names, `--border`,
   `--depth`, and `--noise`, which is 8. The two sets are disjoint, as
   R3.1 claims.
3. R3.5's account of DaisyUI's merge behavior. `theme/index.js`
   branches on `if (allThemes[name])` and merges built-in properties
   under custom tokens only on a name match, exactly as the standard
   describes.
4. R3.5's cascade consequence. Measured, the bare `:root` branch
   appears in one block of the 35, the compiled default, so a short
   block under a novel name emits fewer declarations and the cascade
   serves the default's values for the rest. That is the "close enough
   to pass review and wrong" repaint the clause warns about.
5. R4.1's document-order rule for the `[data-theme]` branch.
   `[data-theme=NAME]` is (0,1,0) and the branding stylesheet's `:root`
   is (0,1,0), so link order decides between them and no specificity
   trick is needed. The rule survives M3's fix; M3 removes the one
   selector that outranks it.
6. A default custom theme cannot outrank tier 2 through its root
   branch. The plugin wraps that branch as `:where(${root})`, which is
   zero specificity, so the branding stylesheet's `:root` wins on
   specificity alone.
7. R2.7's font mechanism at the plugin level. `theme/index.js` captures
   every non-destructured key with a spread and emits it into the block
   unfiltered, so `--font-sans`, `--font-serif`, and `--font-mono` pass
   through. They are Tailwind 4's own font namespace, and Tailwind's
   `font-sans` utility resolves `var(--font-sans)` at use time, so a
   theme-block override takes effect. The mechanism depends on R4.2
   stamping `data-theme` on the document element rather than on
   `<body>`, and R4.2 does specify the document element.
8. R5.1 and G3's CSP reasoning about fonts. `font-src` falls back to
   `default-src`, so a single-origin default already forbids an
   external font origin at runtime. "Embedded fonts need no relaxation"
   is correct, and it is the one part of R5.6 that survives B1 intact.
9. R5.3's variable-font clause. CSS Fonts 4 accepts a range for the
   `font-weight` descriptor, so requiring a variable font to state its
   weight range is correct and checkable from the `@font-face` text
   alone.
10. R1.2's account of what rung 7 keeps. It matches PA:56 clause for
    clause: the roster, the visual authoring of every surface including
    rung 5's login pages, and the attended before/after. Section 7
    repeats the deferral and adds no visual decision.
11. R1.3 and R1.4 against their sources. BS:645-647 does make the
    theme-variable contract the documented customization API with class
    names and DOM excluded, and BS:650-652 does rule out user-supplied
    components and template overrides. TB:202-206 states the same
    constraint independently.
12. R1.5's universality test. CM:7-8 is quoted accurately, and the
    no-rebuild, no-Node, no-network test is applied consistently in
    R4.3's mechanism table. The only estate value in the document,
    `match dubplate, palette gruvbox` at `:161-162`, is quoted from the
    erratum as context, and the roster is left to rung 7.
13. R1.6's handling of the login-page stack conflict. The conflict is
    real (BS:600-601 Go-templated with no JavaScript, CM:142-143
    SvelteKit static, erratum BS:16-18 adopted before rung 5's plan),
    and the standard flags it provisional rather than deciding it
    silently.
14. R2.3's semantic-token rule. BS:645 and `_salvage/web/src/app.css:8-9`
    are quoted verbatim, and extending the ban to Tailwind's numbered
    color utilities is a sound reading of "no hardcoded colors". Only
    the gate's token list is short (M5).
15. R2.5 and R3.5's fail-loud argument. The hazard is documented at
    TB:222-226 ("a partial block silently inherits ... a plausible
    source of an inconsistent pretty close but not quite repaint") and
    the concrete cost at `app.css:13-15` ("paints a rejected album the
    same gray as an accepted one"). Both quotations check out.
16. R2.8's motion clause. BS:683-684 and TB:294-295 both put motion
    global and `prefers-reduced-motion`-aware rather than per-theme.
17. Section 3's preamble against the erratum. Streamer matching
    retired, the axes standing, the palette axis and theming-is-data
    unchanged, estate config `match dubplate, palette gruvbox`, all at
    BS:7-11.
18. R3.1's axis assignment. BS:684-691 and TB:299-305 both describe
    palette blocks setting only color variables and match blocks only
    structural ones. Assigning `--border` and `--noise` to the
    structural axis is unstated at BS:685-686 and follows from
    BS:688-689's rule and BS:677-679's full-block list.
19. R3.3's config claim about the registry's parsers.
    `internal/config/thresholds.go:15-18` is exactly
    `KindInt ThresholdKind = iota` and `KindDuration`, so a string kind
    is genuinely owed. Only the keys.go routing is loose (m8).
20. Section 4's tier list. The three tiers and their order match
    BS:647-650 exactly.
21. R4.3's mechanism weighing. The build-time option would need Node at
    the adopter's site, which CLAUDE.md:141-142 forbids, and the
    fetch-endpoint option would repaint after first paint. The adopted
    generated stylesheet is the ratified wording at BS:648-650
    verbatim, and routing branding text and images to the config
    endpoint instead is a sound split.
22. R4.4's tier 3. TB:151-157 confirms tier 3 is the designed escape
    hatch so that "not on the default roster" never means "not
    possible", and placing it outside R1.3's contract with the risk on
    the operator is consistent with BS:645-647.
23. Section 6's gate-rejoin flag. BS:637-638 and `scripts/check.sh:5`
    both put the web leg's rejoin at rung 7, and the standard raises
    the rung 5 consequence as a provisional ruling rather than assuming
    it.
24. G5's shape, setting M8's vacuity aside. Resolving every built-bundle
    `@font-face` src against the embedded filesystem is the right check
    for a hashed bundle, and it catches the failure it names.
25. R5.2's evidence discipline. S20:395-397 does require a license cell
    citing the LICENSE file by URL "because revision 1's strings were
    filled from memory and two were wrong", so the manifest requirement
    is grounded.
26. Section 7's rung 12 boundary against the erratum. BS:12-15 trims
    ladder entry 12 to the driver seam, SSO, roster sync, and the
    proxy-hygiene ingress, and drops the Navidrome theming note. The
    conclusion that the seam declares no theming capability follows, and
    documenting the post-login visual boundary rather than smoothing it
    is the right reading. Only the S20 citation attached to it is wrong
    (M9).
27. The hook 6 statement. S20:370-373 is verified, and declaring
    theming out of the driver seam entirely is a faithful landing of
    "capabilities a driver may decline to declare". PA:56's "It also
    fixes the same problem for rung 12's driver pairing" is answered.
28. The salvage adoption record, every checkable row: `app.css:8-9`,
    `app.css:13-15`, `app.css:32-47`, `app.css:58-63`,
    `serve.go:125`, `web/embed.go` with its committed
    `dist/placeholder.txt` and the stated reason, and the package pins
    (daisyui 5.7.25, tailwindcss 4.3.3 at `_salvage/web/package.json:21,25`).
29. The adoption record's two negative rows.
    `grep -rn 'font-face\|woff'` over `_salvage/web/src` and
    `_salvage/web/index.html` returns nothing and `_salvage/web/public`
    holds only `placeholder.txt`, so "Nothing to adopt" for the font
    mechanism is accurate and section 5 is new work in full.
    `_salvage/web/index.html` and `package.json` confirm the standalone
    Vite plus Svelte 5 shape the record declines.
30. The `//go:embed all:dist` plus committed placeholder pattern.
    `_salvage/web/embed.go` carries the reason in its own package
    comment, and both `_salvage/web/dist/placeholder.txt` and
    `_salvage/web/public/placeholder.txt` exist as the comment
    describes. Adopting the pattern is right; its side effect on G5 is
    M8.
31. R3.2's honesty about the one-valued match axis. Marking it
    provisional and routing the question to the conductor is the right
    handling. Only its stated reason is void (M12).
32. Every other source citation sampled resolves to the line or range
    given: BS:600, BS:643, BS:645, BS:647-650, BS:650-652, BS:679-681,
    BS:684-691; PA:56 quoted accurately; TB:288-292, TB:297-308,
    TB:305-308; CM:7, CM:133-139, CM:142; `scripts/check.sh:5`;
    `internal/config/thresholds.go`; `_salvage/internal/serve/serve.go:125`;
    `_salvage/web/embed.go`.
33. Prose. `tellgrader` in the docs register reports zero tells over
    3,336 words, cadence CV 0.78, one tricolon, no em dashes, no
    contrast frames, no connector openers, and no scaffold headers. The
    register holds through the document, and the clause-plus-source
    shape is the right form for a normative standard.
